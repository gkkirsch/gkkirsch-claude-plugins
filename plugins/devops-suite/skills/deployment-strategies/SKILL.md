---
name: deployment-strategies
description: >
  Deployment strategies — blue-green, canary, rolling updates, feature flags,
  rollback procedures, database migrations in zero-downtime deploys,
  and deployment automation with GitHub Actions.
  Triggers: "deployment strategy", "blue green", "canary deploy", "rolling update",
  "zero downtime", "rollback", "feature flag", "deploy pipeline".
  NOT for: Container/Kubernetes configuration (use container-orchestration).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Deployment Strategies

## Rolling Update (Kubernetes)

```yaml
# k8s/deployment.yaml — rolling update is the default
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-server
spec:
  replicas: 5
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1        # 1 extra pod during update
      maxUnavailable: 0   # Never reduce below desired count
  template:
    spec:
      containers:
        - name: api
          image: registry.example.com/api:2.1.0
          readinessProbe:
            httpGet:
              path: /health/ready
              port: 8080
            initialDelaySeconds: 5
            periodSeconds: 5
          lifecycle:
            preStop:
              exec:
                command: ["/bin/sh", "-c", "sleep 10"]
                # Give load balancer time to drain connections
      terminationGracePeriodSeconds: 30
```

```bash
# Deploy with rolling update
kubectl set image deployment/api-server api=registry.example.com/api:2.1.0

# Monitor rollout
kubectl rollout status deployment/api-server

# Rollback if something goes wrong
kubectl rollout undo deployment/api-server

# Rollback to specific revision
kubectl rollout history deployment/api-server
kubectl rollout undo deployment/api-server --to-revision=3
```

## Blue-Green Deployment

```yaml
# k8s/blue-green/blue-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-blue
  labels:
    app: api
    version: blue
spec:
  replicas: 3
  selector:
    matchLabels:
      app: api
      version: blue
  template:
    metadata:
      labels:
        app: api
        version: blue
    spec:
      containers:
        - name: api
          image: registry.example.com/api:2.0.0

---
# k8s/blue-green/green-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-green
  labels:
    app: api
    version: green
spec:
  replicas: 3
  selector:
    matchLabels:
      app: api
      version: green
  template:
    metadata:
      labels:
        app: api
        version: green
    spec:
      containers:
        - name: api
          image: registry.example.com/api:2.1.0

---
# k8s/blue-green/service.yaml — switch traffic by changing selector
apiVersion: v1
kind: Service
metadata:
  name: api
spec:
  selector:
    app: api
    version: blue  # Change to "green" to switch
  ports:
    - port: 80
      targetPort: 8080
```

```bash
#!/bin/bash
# scripts/blue-green-switch.sh

CURRENT=$(kubectl get svc api -o jsonpath='{.spec.selector.version}')
TARGET=$([[ "$CURRENT" == "blue" ]] && echo "green" || echo "blue")

echo "Current: $CURRENT -> Switching to: $TARGET"

# Deploy new version to inactive environment
kubectl set image deployment/api-$TARGET api=registry.example.com/api:$NEW_VERSION

# Wait for rollout
kubectl rollout status deployment/api-$TARGET --timeout=300s

# Run smoke tests against inactive service
kubectl run smoke-test --rm -it --image=curlimages/curl -- \
  curl -sf http://api-$TARGET:80/health/ready

# Switch traffic
kubectl patch svc api -p "{\"spec\":{\"selector\":{\"version\":\"$TARGET\"}}}"

echo "Traffic switched to $TARGET"
echo "Previous version ($CURRENT) still running — scale down after verification"
```

## Canary Deployment (Nginx Ingress)

```yaml
# k8s/canary/stable-ingress.yaml
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
# k8s/canary/canary-ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: api-canary
  annotations:
    nginx.ingress.kubernetes.io/canary: "true"
    nginx.ingress.kubernetes.io/canary-weight: "10"  # 10% of traffic
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

```bash
#!/bin/bash
# scripts/canary-promote.sh — progressive canary rollout

WEIGHTS=(10 25 50 75 100)

for weight in "${WEIGHTS[@]}"; do
  echo "Setting canary weight to ${weight}%"
  kubectl annotate ingress api-canary \
    nginx.ingress.kubernetes.io/canary-weight="$weight" \
    --overwrite

  echo "Waiting 5 minutes for metrics..."
  sleep 300

  # Check error rate
  ERROR_RATE=$(curl -s "http://prometheus:9090/api/v1/query?query=rate(http_requests_total{status=~'5..', version='canary'}[5m])/rate(http_requests_total{version='canary'}[5m])" \
    | jq '.data.result[0].value[1] // "0"' -r)

  if (( $(echo "$ERROR_RATE > 0.01" | bc -l) )); then
    echo "ERROR: Canary error rate ${ERROR_RATE} exceeds threshold. Rolling back."
    kubectl annotate ingress api-canary \
      nginx.ingress.kubernetes.io/canary-weight="0" --overwrite
    exit 1
  fi

  echo "Canary healthy at ${weight}%. Error rate: ${ERROR_RATE}"
done

echo "Canary promotion complete. Updating stable deployment."
kubectl set image deployment/api-stable api=registry.example.com/api:$NEW_VERSION
kubectl rollout status deployment/api-stable
kubectl annotate ingress api-canary nginx.ingress.kubernetes.io/canary-weight="0" --overwrite
```

## Feature Flags

```typescript
// src/features/feature-flags.ts

interface FeatureFlag {
  name: string;
  enabled: boolean;
  rolloutPercentage?: number; // 0-100
  allowedUsers?: string[];
  allowedGroups?: string[];
}

class FeatureFlagService {
  private flags: Map<string, FeatureFlag> = new Map();

  constructor(private source: "env" | "db" | "launchdarkly") {}

  async isEnabled(
    flagName: string,
    context?: { userId?: string; groups?: string[] }
  ): Promise<boolean> {
    const flag = await this.getFlag(flagName);
    if (!flag) return false;
    if (!flag.enabled) return false;

    // Check user allowlist
    if (flag.allowedUsers?.length && context?.userId) {
      if (flag.allowedUsers.includes(context.userId)) return true;
    }

    // Check group allowlist
    if (flag.allowedGroups?.length && context?.groups) {
      if (context.groups.some((g) => flag.allowedGroups!.includes(g))) return true;
    }

    // Percentage rollout — deterministic by userId
    if (flag.rolloutPercentage !== undefined && context?.userId) {
      const hash = this.hashUserId(context.userId);
      return hash % 100 < flag.rolloutPercentage;
    }

    return flag.enabled;
  }

  private hashUserId(userId: string): number {
    let hash = 0;
    for (let i = 0; i < userId.length; i++) {
      hash = (hash * 31 + userId.charCodeAt(i)) | 0;
    }
    return Math.abs(hash);
  }

  private async getFlag(name: string): Promise<FeatureFlag | null> {
    // Implementation depends on source (env, database, LaunchDarkly)
    return this.flags.get(name) ?? null;
  }
}

// Usage in routes
const features = new FeatureFlagService("db");

app.get("/api/search", async (req, res) => {
  const useNewSearch = await features.isEnabled("new-search-engine", {
    userId: req.user.id,
    groups: req.user.groups,
  });

  if (useNewSearch) {
    return newSearchHandler(req, res);
  }
  return legacySearchHandler(req, res);
});

// Environment-based flags (simplest approach)
// .env: FEATURE_NEW_SEARCH=true
function envFlag(name: string): boolean {
  return process.env[`FEATURE_${name.toUpperCase()}`] === "true";
}
```

## Zero-Downtime Database Migrations

```typescript
// migrations/safe-migration-patterns.ts

// SAFE: Add column (nullable, no default computed from data)
export async function up(knex: Knex) {
  await knex.schema.alterTable("users", (table) => {
    table.string("display_name").nullable(); // nullable = no lock
  });
}

// SAFE: Add column with server-side default
export async function up(knex: Knex) {
  await knex.schema.alterTable("orders", (table) => {
    table.string("status").defaultTo("pending"); // constant default = fast
  });
}

// SAFE: Create index concurrently (PostgreSQL)
export async function up(knex: Knex) {
  // CREATE INDEX CONCURRENTLY doesn't lock the table
  await knex.raw(
    "CREATE INDEX CONCURRENTLY idx_orders_status ON orders (status)"
  );
}

// DANGEROUS: Rename column — breaks old code still running
// SAFE alternative: expand-contract pattern
export async function phase1_expand(knex: Knex) {
  // Phase 1: Add new column, backfill, update app to write both
  await knex.schema.alterTable("users", (table) => {
    table.string("full_name").nullable();
  });
  // Backfill
  await knex.raw("UPDATE users SET full_name = name WHERE full_name IS NULL");
}

export async function phase2_migrate_reads(knex: Knex) {
  // Phase 2: App reads from full_name (deploy app first)
  // This migration is a no-op — it's a checkpoint
}

export async function phase3_contract(knex: Knex) {
  // Phase 3: Drop old column (only after all instances use full_name)
  await knex.schema.alterTable("users", (table) => {
    table.dropColumn("name");
  });
}

// DANGEROUS: NOT NULL on existing column — locks table during check
// SAFE alternative: add constraint with NOT VALID
export async function up(knex: Knex) {
  // Add constraint without validating existing rows (instant)
  await knex.raw(`
    ALTER TABLE orders
    ADD CONSTRAINT orders_status_not_null
    CHECK (status IS NOT NULL) NOT VALID
  `);
  // Validate in background (doesn't lock writes)
  await knex.raw(`
    ALTER TABLE orders
    VALIDATE CONSTRAINT orders_status_not_null
  `);
}
```

## GitHub Actions Deployment Pipeline

```yaml
# .github/workflows/deploy.yml
name: Deploy

on:
  push:
    branches: [main]

concurrency:
  group: deploy-production
  cancel-in-progress: false  # Never cancel a running deploy

jobs:
  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:16
        env:
          POSTGRES_DB: test
          POSTGRES_PASSWORD: test
        ports: ['5432:5432']
        options: --health-cmd pg_isready --health-interval 10s
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with: { node-version: 20, cache: 'npm' }
      - run: npm ci
      - run: npm test
        env:
          DATABASE_URL: postgres://postgres:test@localhost:5432/test

  build:
    needs: test
    runs-on: ubuntu-latest
    outputs:
      image-tag: ${{ steps.meta.outputs.tags }}
    steps:
      - uses: actions/checkout@v4

      - name: Docker meta
        id: meta
        uses: docker/metadata-action@v5
        with:
          images: registry.example.com/api
          tags: |
            type=sha,prefix=
            type=ref,event=branch

      - uses: docker/setup-buildx-action@v3

      - uses: docker/build-push-action@v6
        with:
          context: .
          push: true
          tags: ${{ steps.meta.outputs.tags }}
          cache-from: type=gha
          cache-to: type=gha,mode=max

  deploy-staging:
    needs: build
    runs-on: ubuntu-latest
    environment: staging
    steps:
      - uses: actions/checkout@v4

      - name: Deploy to staging
        run: |
          helm upgrade --install api-server ./charts/api-server \
            --namespace staging \
            --set image.tag=${{ github.sha }} \
            --values ./charts/api-server/values-staging.yaml \
            --wait --timeout 5m

      - name: Run smoke tests
        run: |
          npm run test:smoke -- --base-url=https://staging-api.example.com

  deploy-production:
    needs: deploy-staging
    runs-on: ubuntu-latest
    environment:
      name: production
      url: https://api.example.com
    steps:
      - uses: actions/checkout@v4

      - name: Deploy to production
        run: |
          helm upgrade --install api-server ./charts/api-server \
            --namespace production \
            --set image.tag=${{ github.sha }} \
            --values ./charts/api-server/values-production.yaml \
            --wait --timeout 10m

      - name: Verify deployment
        run: |
          for i in {1..10}; do
            STATUS=$(curl -sf https://api.example.com/health/ready | jq -r '.status')
            if [ "$STATUS" = "healthy" ]; then
              echo "Deployment verified healthy"
              exit 0
            fi
            sleep 10
          done
          echo "Deployment health check failed"
          exit 1

      - name: Rollback on failure
        if: failure()
        run: |
          helm rollback api-server --namespace production
          echo "::error::Production deploy failed. Rolled back to previous release."
```

## Rollback Procedures

```bash
#!/bin/bash
# scripts/rollback.sh — production rollback runbook

set -euo pipefail

NAMESPACE="${1:-production}"
DEPLOYMENT="${2:-api-server}"

echo "=== ROLLBACK PROCEDURE ==="
echo "Namespace: $NAMESPACE"
echo "Deployment: $DEPLOYMENT"

# Step 1: Check current status
echo ""
echo "--- Current State ---"
kubectl get deployment $DEPLOYMENT -n $NAMESPACE \
  -o jsonpath='{.spec.template.spec.containers[0].image}'
echo ""
kubectl rollout status deployment/$DEPLOYMENT -n $NAMESPACE --timeout=10s || true

# Step 2: Show revision history
echo ""
echo "--- Revision History ---"
kubectl rollout history deployment/$DEPLOYMENT -n $NAMESPACE

# Step 3: Rollback
echo ""
read -p "Rollback to previous revision? (y/N) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
  kubectl rollout undo deployment/$DEPLOYMENT -n $NAMESPACE
  kubectl rollout status deployment/$DEPLOYMENT -n $NAMESPACE --timeout=300s

  echo ""
  echo "--- Post-Rollback Status ---"
  kubectl get pods -n $NAMESPACE -l app=$DEPLOYMENT
  NEW_IMAGE=$(kubectl get deployment $DEPLOYMENT -n $NAMESPACE \
    -o jsonpath='{.spec.template.spec.containers[0].image}')
  echo "Current image: $NEW_IMAGE"
fi
```

## Deployment Checklist

```yaml
# .github/DEPLOYMENT_CHECKLIST.md
## Pre-Deploy
- [ ] All tests passing on main
- [ ] Database migrations are backward-compatible
- [ ] Feature flags configured for new features
- [ ] Monitoring dashboards reviewed (baseline metrics noted)
- [ ] Rollback plan documented and tested

## During Deploy
- [ ] Deploy to staging first
- [ ] Run smoke tests against staging
- [ ] Verify staging metrics (error rate, latency, throughput)
- [ ] Deploy to production (canary or rolling)
- [ ] Monitor error rate during rollout
- [ ] Verify health endpoints responding

## Post-Deploy
- [ ] Check application logs for new errors
- [ ] Verify key user flows manually
- [ ] Check Sentry/error tracking for new issues
- [ ] Confirm metrics within normal ranges
- [ ] Notify team of successful deploy
- [ ] Update deployment log
```

## Gotchas

1. **Database migrations run before new code** — In a rolling deploy, migration runs first, then old pods still handle traffic with the new schema. If your migration renames or removes a column, old pods crash. Always use the expand-contract pattern: add new → migrate data → deploy code using new → remove old.

2. **Health check grace period too short** — Java/Spring apps take 30-60 seconds to start. A `startupProbe` with only 10 seconds of `initialDelaySeconds` kills pods before they're ready, causing a restart loop. Set `startupProbe.failureThreshold * periodSeconds` to exceed your worst-case startup time.

3. **`preStop` hook and connection draining** — Without a `preStop` sleep, Kubernetes removes the pod from the Service endpoints and sends SIGTERM simultaneously. In-flight requests get dropped because the load balancer hasn't updated yet. Add `preStop: exec: command: ["sleep", "10"]` to give the load balancer time to drain.

4. **Helm `--wait` doesn't catch CrashLoopBackOff** — `helm upgrade --wait` succeeds as long as pods reach Ready state at least once, even if they immediately crash after. Add a post-deploy health check that verifies the service responds correctly for 30+ seconds, not just once.

5. **Feature flag cleanup debt** — Old feature flags accumulate and make code unreadable. Every flag should have an expiry date. After a feature is fully rolled out (100% for 2+ weeks), remove the flag and the old code path. Track flag lifecycle in your issue tracker.

6. **Canary metrics poisoned by caching** — If your CDN or reverse proxy caches responses, canary traffic metrics are misleading — cached responses don't hit the canary. Bypass the cache for canary traffic with a custom header (`X-Canary: true`) or use cache-busting query parameters during the canary window.
