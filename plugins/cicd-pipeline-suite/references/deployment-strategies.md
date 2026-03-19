# Deployment Strategies Reference

Comprehensive reference for zero-downtime deployment patterns, rollback procedures, progressive delivery, and production safety nets.

---

## Strategy Comparison

| Strategy | Downtime | Rollback Speed | Infrastructure Cost | Risk | Best For |
|----------|----------|----------------|---------------------|------|----------|
| Recreate | Yes | Slow | 1x | High | Dev/staging |
| Rolling | No | Medium | 1x + surge | Medium | Most workloads |
| Blue-Green | No | Instant | 2x | Low | Critical services |
| Canary | No | Fast | 1x + small % | Very Low | High-traffic services |
| A/B Testing | No | Fast | 1x + small % | Very Low | Feature experiments |
| Shadow | No | N/A | 2x | None | Migration validation |

---

## Recreate Deployment

Terminate all old pods, then start new ones. Simple but causes downtime.

```yaml
# Kubernetes
apiVersion: apps/v1
kind: Deployment
spec:
  strategy:
    type: Recreate
```

**When to use**: Development environments, batch jobs, or when the application can't run two versions simultaneously (e.g., database schema lock).

---

## Rolling Update

Replace instances gradually. Kubernetes default strategy.

### Kubernetes Configuration

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api
spec:
  replicas: 6
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 2           # Extra pods during update (absolute or %)
      maxUnavailable: 1     # Pods that can be unavailable during update
  minReadySeconds: 15       # Wait after pod is ready
  revisionHistoryLimit: 10  # ReplicaSets to keep for rollback
  progressDeadlineSeconds: 600  # Fail if no progress in 10 min
  template:
    spec:
      terminationGracePeriodSeconds: 30
      containers:
        - name: api
          image: api:2.0.0
          ports:
            - containerPort: 8080

          # Readiness probe — pod only receives traffic when ready
          readinessProbe:
            httpGet:
              path: /readyz
              port: 8080
            initialDelaySeconds: 5
            periodSeconds: 5
            successThreshold: 1
            failureThreshold: 3

          # Liveness probe — restart pod if unhealthy
          livenessProbe:
            httpGet:
              path: /livez
              port: 8080
            initialDelaySeconds: 15
            periodSeconds: 10
            failureThreshold: 3

          # Startup probe — delay liveness until app is started
          startupProbe:
            httpGet:
              path: /healthz
              port: 8080
            failureThreshold: 30
            periodSeconds: 5

          # Graceful shutdown
          lifecycle:
            preStop:
              exec:
                command: ["/bin/sh", "-c", "sleep 5"]
```

### Rolling Update Math

With 6 replicas, maxSurge=2, maxUnavailable=1:
- Maximum pods during update: 6 + 2 = 8
- Minimum available pods: 6 - 1 = 5
- Process: Create 2 new pods → wait until ready → terminate 1 old pod → repeat

### Rollback

```bash
# View history
kubectl rollout history deployment/api

# Rollback to previous
kubectl rollout undo deployment/api

# Rollback to specific revision
kubectl rollout undo deployment/api --to-revision=5

# Check status
kubectl rollout status deployment/api
```

---

## Blue-Green Deployment

Maintain two identical environments. Deploy to inactive, switch traffic.

### Implementation with Kubernetes Services

```yaml
# Deployment v1 (Blue — currently serving)
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-blue
  labels:
    app: api
    slot: blue
spec:
  replicas: 3
  selector:
    matchLabels:
      app: api
      slot: blue
  template:
    metadata:
      labels:
        app: api
        slot: blue
    spec:
      containers:
        - name: api
          image: api:1.0.0
          ports:
            - containerPort: 8080

---
# Deployment v2 (Green — new version, not yet serving)
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-green
  labels:
    app: api
    slot: green
spec:
  replicas: 3
  selector:
    matchLabels:
      app: api
      slot: green
  template:
    metadata:
      labels:
        app: api
        slot: green
    spec:
      containers:
        - name: api
          image: api:2.0.0
          ports:
            - containerPort: 8080

---
# Production service — selector determines which slot gets traffic
apiVersion: v1
kind: Service
metadata:
  name: api
spec:
  selector:
    app: api
    slot: blue    # ← Change to "green" to switch traffic
  ports:
    - port: 80
      targetPort: 8080

---
# Test service — always points to inactive slot for pre-deployment testing
apiVersion: v1
kind: Service
metadata:
  name: api-test
spec:
  selector:
    app: api
    slot: green   # ← Always points to the NEW version
  ports:
    - port: 80
      targetPort: 8080
```

### Blue-Green Switch Script

```bash
#!/bin/bash
set -euo pipefail

SERVICE="api"
NAMESPACE="${NAMESPACE:-default}"

# Determine current and target slots
CURRENT_SLOT=$(kubectl get service "$SERVICE" -n "$NAMESPACE" \
  -o jsonpath='{.spec.selector.slot}')

if [ "$CURRENT_SLOT" = "blue" ]; then
  TARGET_SLOT="green"
else
  TARGET_SLOT="blue"
fi

echo "Current: $CURRENT_SLOT → Target: $TARGET_SLOT"

# 1. Verify target deployment is healthy
echo "Checking target deployment health..."
READY=$(kubectl get deployment "${SERVICE}-${TARGET_SLOT}" -n "$NAMESPACE" \
  -o jsonpath='{.status.readyReplicas}' 2>/dev/null || echo "0")
DESIRED=$(kubectl get deployment "${SERVICE}-${TARGET_SLOT}" -n "$NAMESPACE" \
  -o jsonpath='{.spec.replicas}')

if [ "$READY" != "$DESIRED" ]; then
  echo "ERROR: ${SERVICE}-${TARGET_SLOT} not ready ($READY/$DESIRED replicas)"
  exit 1
fi

# 2. Run smoke tests against test service
echo "Running smoke tests..."
TEST_IP=$(kubectl get service "${SERVICE}-test" -n "$NAMESPACE" \
  -o jsonpath='{.spec.clusterIP}')
kubectl run smoke-test --rm -i --restart=Never --image=curlimages/curl -- \
  curl -sf "http://${TEST_IP}/healthz" || {
    echo "ERROR: Smoke test failed"
    exit 1
  }

# 3. Switch traffic
echo "Switching traffic to $TARGET_SLOT..."
kubectl patch service "$SERVICE" -n "$NAMESPACE" \
  -p "{\"spec\":{\"selector\":{\"slot\":\"${TARGET_SLOT}\"}}}"

echo "Traffic switched to $TARGET_SLOT"

# 4. Verify
sleep 5
kubectl get endpoints "$SERVICE" -n "$NAMESPACE"
echo "Blue-green switch complete"
```

### AWS ALB Blue-Green with CodeDeploy

```yaml
# appspec.yml
version: 0.0
Resources:
  - TargetService:
      Type: AWS::ECS::Service
      Properties:
        TaskDefinition: <TASK_DEFINITION>
        LoadBalancerInfo:
          ContainerName: api
          ContainerPort: 8080
        PlatformVersion: LATEST

Hooks:
  - BeforeInstall: "arn:aws:lambda:us-east-1:123456789:function:before-install"
  - AfterInstall: "arn:aws:lambda:us-east-1:123456789:function:after-install"
  - AfterAllowTestTraffic: "arn:aws:lambda:us-east-1:123456789:function:smoke-test"
  - BeforeAllowTraffic: "arn:aws:lambda:us-east-1:123456789:function:integration-test"
  - AfterAllowTraffic: "arn:aws:lambda:us-east-1:123456789:function:verify-production"
```

---

## Canary Deployment

Send a small percentage of traffic to the new version, monitor, then gradually increase.

### Progressive Traffic Shifting

```
Step 1:  5% canary → Monitor 5 min
Step 2: 10% canary → Monitor 5 min
Step 3: 25% canary → Monitor 10 min
Step 4: 50% canary → Monitor 10 min
Step 5: 75% canary → Monitor 5 min
Step 6: 100% → Promote
```

### Argo Rollouts Canary

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Rollout
metadata:
  name: api
spec:
  replicas: 10
  revisionHistoryLimit: 5
  selector:
    matchLabels:
      app: api
  strategy:
    canary:
      canaryService: api-canary
      stableService: api-stable
      trafficRouting:
        nginx:
          stableIngress: api-ingress
          annotationPrefix: nginx.ingress.kubernetes.io
      steps:
        - setWeight: 5
        - pause: { duration: 5m }
        - analysis:
            templates:
              - templateName: success-rate
              - templateName: latency
        - setWeight: 25
        - pause: { duration: 10m }
        - analysis:
            templates:
              - templateName: success-rate
              - templateName: latency
        - setWeight: 50
        - pause: { duration: 10m }
        - analysis:
            templates:
              - templateName: success-rate
        - setWeight: 100
      rollbackWindow:
        revisions: 3
      maxSurge: 20%
      maxUnavailable: 0

---
apiVersion: argoproj.io/v1alpha1
kind: AnalysisTemplate
metadata:
  name: success-rate
spec:
  metrics:
    - name: success-rate
      interval: 60s
      count: 5
      successCondition: result[0] >= 0.995
      failureLimit: 2
      provider:
        prometheus:
          address: http://prometheus:9090
          query: |
            sum(rate(http_requests_total{app="api",status=~"2.."}[2m]))
            /
            sum(rate(http_requests_total{app="api"}[2m]))

---
apiVersion: argoproj.io/v1alpha1
kind: AnalysisTemplate
metadata:
  name: latency
spec:
  metrics:
    - name: p99-latency
      interval: 60s
      count: 5
      successCondition: result[0] < 0.5
      failureLimit: 2
      provider:
        prometheus:
          address: http://prometheus:9090
          query: |
            histogram_quantile(0.99,
              sum(rate(http_request_duration_seconds_bucket{app="api"}[2m])) by (le)
            )
```

### Canary with Nginx Ingress (No Service Mesh)

```yaml
# Stable ingress (receives most traffic)
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: api-stable
spec:
  ingressClassName: nginx
  rules:
    - host: api.example.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: api-stable
                port:
                  number: 80

---
# Canary ingress (receives canary percentage)
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: api-canary
  annotations:
    nginx.ingress.kubernetes.io/canary: "true"
    nginx.ingress.kubernetes.io/canary-weight: "10"
    # Or route specific users:
    # nginx.ingress.kubernetes.io/canary-by-header: "X-Canary"
    # nginx.ingress.kubernetes.io/canary-by-header-value: "true"
    # nginx.ingress.kubernetes.io/canary-by-cookie: "canary"
spec:
  ingressClassName: nginx
  rules:
    - host: api.example.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: api-canary
                port:
                  number: 80
```

### Canary with Istio

```yaml
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: api
spec:
  hosts:
    - api.example.com
  http:
    # Route internal testers to canary
    - match:
        - headers:
            x-canary:
              exact: "true"
      route:
        - destination:
            host: api
            subset: canary

    # Percentage-based split
    - route:
        - destination:
            host: api
            subset: stable
          weight: 90
        - destination:
            host: api
            subset: canary
          weight: 10

---
apiVersion: networking.istio.io/v1beta1
kind: DestinationRule
metadata:
  name: api
spec:
  host: api
  subsets:
    - name: stable
      labels:
        version: v1
    - name: canary
      labels:
        version: v2
```

### Canary Metrics to Monitor

Before promoting a canary, verify these metrics are within acceptable range compared to stable:

```
1. Error rate: canary error rate <= stable error rate + 0.5%
2. Latency p50: canary p50 <= stable p50 * 1.1 (10% tolerance)
3. Latency p99: canary p99 <= stable p99 * 1.2 (20% tolerance)
4. Saturation: CPU/memory not significantly higher than stable
5. Business metrics: conversion rate, cart abandonment, etc.
```

---

## Shadow/Dark Deployment

Send a copy of production traffic to the new version without affecting users. The shadow version's responses are discarded.

### When to Use

- Major refactors where you want to compare behavior
- Database migration validation
- Performance benchmarking under real load
- New service validation before cutover

### Implementation with Istio

```yaml
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: api
spec:
  hosts:
    - api.example.com
  http:
    - route:
        - destination:
            host: api-v1      # Real traffic goes here
      mirror:
        host: api-v2          # Copy of traffic goes here
      mirrorPercentage:
        value: 100.0           # Mirror 100% of traffic
```

### Shadow Comparison Script

```bash
#!/bin/bash
# Compare responses between production and shadow

echo "Comparing production vs shadow responses..."

# Sample responses from both versions
for i in $(seq 1 100); do
  PROD=$(curl -s https://api.example.com/users/1)
  SHADOW=$(curl -s https://api-shadow.example.com/users/1)

  if [ "$PROD" != "$SHADOW" ]; then
    echo "DIFF at request $i:"
    diff <(echo "$PROD" | jq .) <(echo "$SHADOW" | jq .)
  fi
done

echo "Comparison complete"
```

---

## Feature Flags for Deployment

Decouple deployment from release. Ship code behind flags, enable gradually.

### Feature Flag Deployment Flow

```
1. Deploy code with feature behind flag (flag OFF)
2. Enable flag for internal users → test in production
3. Enable flag for 1% → monitor
4. Ramp to 10% → 25% → 50% → 100%
5. Remove flag and old code path (cleanup!)
```

### Server-Side Feature Flags

```typescript
// Simple percentage-based flag
function isFeatureEnabled(flagName: string, userId: string): boolean {
  const flag = getFlag(flagName);
  if (!flag.enabled) return false;

  // Consistent hashing — same user always gets same result
  const hash = createHash('md5')
    .update(`${flagName}:${userId}`)
    .digest();
  const percentage = hash.readUInt32BE(0) % 100;

  return percentage < flag.rolloutPercentage;
}

// Usage
app.get('/api/checkout', (req, res) => {
  if (isFeatureEnabled('new-checkout', req.user.id)) {
    return newCheckoutFlow(req, res);
  }
  return legacyCheckoutFlow(req, res);
});
```

### Kill Switch Pattern

```typescript
// Emergency kill switch — instantly disable a feature
const KILL_SWITCHES = {
  'payment-processing': true,   // Feature is enabled
  'email-notifications': true,
  'new-search': false,          // KILLED — disabled for everyone
};

function checkKillSwitch(feature: string): boolean {
  return KILL_SWITCHES[feature] !== false;
}

// In your deployment pipeline, flip the kill switch:
// 1. Detect incident
// 2. Set kill switch to false
// 3. Feature is instantly disabled without deployment
```

---

## Zero-Downtime Database Changes

### The Expand-Contract Pattern

Never make breaking schema changes in a single deployment. Always expand first, then contract.

### Adding a Column

```sql
-- Deploy 1: Add column (nullable, with default)
ALTER TABLE users ADD COLUMN email_verified BOOLEAN DEFAULT false;

-- Deploy 2: Backfill existing data
UPDATE users SET email_verified = true
WHERE email IS NOT NULL AND created_at < '2024-01-01';

-- Deploy 3: Start using the column in application code

-- Deploy 4 (optional): Add NOT NULL constraint
ALTER TABLE users ALTER COLUMN email_verified SET NOT NULL;
```

### Renaming a Column

```sql
-- Deploy 1: Add new column
ALTER TABLE orders ADD COLUMN total_amount DECIMAL(10,2);

-- Deploy 2: Start dual-writing (app writes to both columns)
-- Application code: order.total_amount = order.amount;

-- Deploy 3: Backfill
UPDATE orders SET total_amount = amount WHERE total_amount IS NULL;

-- Deploy 4: Switch reads to new column
-- Application code: reads from total_amount

-- Deploy 5: Stop writing to old column

-- Deploy 6: Drop old column (after verification period)
ALTER TABLE orders DROP COLUMN amount;
```

### Adding an Index (Large Tables)

```sql
-- WRONG: Locks the table
CREATE INDEX idx_orders_status ON orders(status);

-- RIGHT: Non-blocking (PostgreSQL)
CREATE INDEX CONCURRENTLY idx_orders_status ON orders(status);

-- Check progress
SELECT phase, blocks_total, blocks_done,
       round(100.0 * blocks_done / nullif(blocks_total, 0), 1) as pct
FROM pg_stat_progress_create_index;
```

### Changing a Column Type

```sql
-- WRONG: Locks table and rewrites
ALTER TABLE orders ALTER COLUMN status TYPE varchar(50);

-- RIGHT: Expand-Contract
-- Deploy 1: Add new column
ALTER TABLE orders ADD COLUMN status_v2 varchar(50);

-- Deploy 2: Dual-write + backfill
UPDATE orders SET status_v2 = status::varchar(50);

-- Deploy 3: Switch reads
-- Deploy 4: Stop writing old column
-- Deploy 5: Drop old column, rename new
ALTER TABLE orders DROP COLUMN status;
ALTER TABLE orders RENAME COLUMN status_v2 TO status;
```

### Migration Rollback Strategy

```typescript
// Always write reversible migrations
// migrations/001_add_user_status.ts

export async function up(db: Database) {
  await db.query(`ALTER TABLE users ADD COLUMN status TEXT DEFAULT 'active'`);
}

export async function down(db: Database) {
  await db.query(`ALTER TABLE users DROP COLUMN IF EXISTS status`);
}
```

---

## Rollback Procedures

### Rollback Decision Tree

```
Is the issue affecting users?
├── No → Monitor, investigate, no rush
└── Yes
    ├── Can you fix forward quickly? (< 15 min)
    │   ├── Yes → Hotfix, deploy, monitor
    │   └── No → Rollback
    └── Is there a feature flag?
        ├── Yes → Toggle flag off (instant)
        └── No → Full rollback
```

### Kubernetes Rollback

```bash
# Instant rollback to previous revision
kubectl rollout undo deployment/api

# Rollback to specific revision
kubectl rollout history deployment/api
kubectl rollout undo deployment/api --to-revision=5

# Verify
kubectl rollout status deployment/api
kubectl get pods -l app=api
```

### Helm Rollback

```bash
# List release history
helm history api -n production

# Rollback to previous release
helm rollback api -n production

# Rollback to specific revision
helm rollback api 5 -n production

# Rollback with timeout
helm rollback api 5 -n production --timeout 5m
```

### ArgoCD Rollback

```bash
# Via CLI
argocd app history myapp
argocd app rollback myapp 3

# Via GitOps (preferred)
cd k8s-manifests
git revert HEAD
git push

# ArgoCD auto-syncs the reverted state
```

### Docker/ECS Rollback

```bash
# List task definition revisions
aws ecs list-task-definitions --family api --sort DESC --max-items 5

# Update service to previous revision
aws ecs update-service \
  --cluster production \
  --service api \
  --task-definition api:42  # Previous working revision

# Verify
aws ecs describe-services --cluster production --services api \
  --query 'services[0].deployments'
```

---

## Deployment Safeguards

### Pre-Deployment Checklist

```yaml
# GitHub Actions pre-deploy checks
- name: Pre-deployment validation
  run: |
    # 1. All tests pass
    npm test

    # 2. No critical vulnerabilities
    npm audit --audit-level=critical

    # 3. Build succeeds
    npm run build

    # 4. Smoke test the build
    npm run test:smoke

    # 5. Database migrations are safe
    npx prisma migrate diff --from-url "$DATABASE_URL" --to-migrations ./prisma/migrations

    # 6. No secrets in code
    npx secretlint "**/*"
```

### Post-Deployment Verification

```yaml
- name: Post-deployment verification
  run: |
    # Wait for deployment to stabilize
    sleep 30

    # Health check
    HTTP_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$DEPLOY_URL/healthz")
    if [ "$HTTP_STATUS" != "200" ]; then
      echo "Health check failed: $HTTP_STATUS"
      # Trigger rollback
      kubectl rollout undo deployment/api
      exit 1
    fi

    # Smoke tests
    npm run test:smoke -- --url "$DEPLOY_URL"

    # Check error rate
    ERROR_RATE=$(curl -s "$PROMETHEUS_URL/api/v1/query?query=..." | jq -r '.data.result[0].value[1]')
    if (( $(echo "$ERROR_RATE > 0.05" | bc -l) )); then
      echo "Error rate too high: $ERROR_RATE"
      kubectl rollout undo deployment/api
      exit 1
    fi

    echo "Deployment verified successfully"
```

### Deployment Windows

```yaml
# Only deploy during business hours (safer for monitoring)
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - name: Check deployment window
        run: |
          HOUR=$(TZ=America/New_York date +%H)
          DOW=$(date +%u)  # 1=Monday, 7=Sunday

          # No deploys on weekends
          if [ "$DOW" -gt 5 ]; then
            echo "No deployments on weekends"
            exit 1
          fi

          # No deploys outside 9 AM - 4 PM ET
          if [ "$HOUR" -lt 9 ] || [ "$HOUR" -gt 16 ]; then
            echo "Outside deployment window (9 AM - 4 PM ET)"
            exit 1
          fi

          # No deploys on Fridays after 2 PM
          if [ "$DOW" -eq 5 ] && [ "$HOUR" -gt 14 ]; then
            echo "No Friday afternoon deployments"
            exit 1
          fi
```

---

## GitOps Deployment

### ArgoCD Application

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: api
  namespace: argocd
  finalizers:
    - resources-finalizer.argocd.argoproj.io
spec:
  project: default
  source:
    repoURL: https://github.com/org/k8s-manifests.git
    targetRevision: main
    path: apps/api/overlays/production
    kustomize:
      images:
        - api=ghcr.io/org/api
  destination:
    server: https://kubernetes.default.svc
    namespace: production
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
      allowEmpty: false
    syncOptions:
      - CreateNamespace=true
      - PrunePropagationPolicy=foreground
      - PruneLast=true
      - ServerSideApply=true
    retry:
      limit: 5
      backoff:
        duration: 5s
        factor: 2
        maxDuration: 3m
```

### GitOps Image Update Automation

```yaml
# In CI pipeline: update manifests repo with new image tag
name: Update Manifests

on:
  workflow_run:
    workflows: [Build]
    types: [completed]
    branches: [main]

jobs:
  update:
    if: github.event.workflow_run.conclusion == 'success'
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          repository: org/k8s-manifests
          token: ${{ secrets.GITOPS_PAT }}

      - name: Update image tag
        run: |
          cd apps/api/overlays/production
          kustomize edit set image "api=ghcr.io/org/api:${{ github.sha }}"

      - name: Commit and push
        run: |
          git config user.name "github-actions[bot]"
          git config user.email "github-actions[bot]@users.noreply.github.com"
          git add .
          git commit -m "chore(api): update image to ${{ github.sha }}"
          git push
```

---

## Multi-Environment Pipeline

### Environment Promotion Flow

```
dev → staging → production
 │       │          │
 │       │          └── Manual approval + deployment window
 │       └── Auto-deploy on main merge + smoke tests
 └── Auto-deploy on every commit + unit tests
```

### GitHub Actions Multi-Environment

```yaml
name: Deploy Pipeline

on:
  push:
    branches: [main]

jobs:
  build:
    runs-on: ubuntu-latest
    outputs:
      image-tag: ${{ steps.build.outputs.tag }}
    steps:
      - uses: actions/checkout@v4
      - id: build
        run: |
          TAG="${{ github.sha }}"
          docker build -t "api:$TAG" .
          docker push "ghcr.io/org/api:$TAG"
          echo "tag=$TAG" >> "$GITHUB_OUTPUT"

  deploy-staging:
    needs: build
    runs-on: ubuntu-latest
    environment:
      name: staging
      url: https://staging.example.com
    steps:
      - run: deploy --env staging --tag ${{ needs.build.outputs.image-tag }}
      - run: npm run test:smoke -- --url https://staging.example.com

  deploy-production:
    needs: deploy-staging
    runs-on: ubuntu-latest
    environment:
      name: production   # Requires manual approval
      url: https://example.com
    steps:
      - run: deploy --env production --tag ${{ needs.build.outputs.image-tag }}
      - run: npm run test:smoke -- --url https://example.com
```

---

## Incident Response During Deployment

### Runbook Template

```markdown
## Deployment Incident Runbook

### Detection
- Alert fires: [Alert Name]
- Dashboard: [Link to Grafana dashboard]
- Error rate threshold: > 1% for 5 minutes

### Immediate Actions
1. Check deployment status: `kubectl rollout status deployment/api`
2. Check pod health: `kubectl get pods -l app=api`
3. Check recent logs: `kubectl logs -l app=api --tail=100 --since=5m`

### Rollback Decision
- If error rate > 5%: Immediate rollback
- If error rate 1-5%: Investigate for 5 minutes, then rollback
- If p99 latency > 2x normal: Rollback

### Rollback Steps
1. `kubectl rollout undo deployment/api`
2. Verify: `kubectl rollout status deployment/api`
3. Verify: Check error rate returns to normal

### Communication
1. Post in #incidents: "Rolling back API deployment due to [reason]"
2. Update status page if user-facing
3. Page on-call if rollback fails

### Post-Incident
1. Root cause analysis within 24 hours
2. Update this runbook if needed
3. Add missing monitoring/alerts
```
