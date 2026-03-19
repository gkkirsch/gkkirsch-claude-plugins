# Kubernetes Patterns Reference

Quick-reference guide for common Kubernetes patterns. Agents consult this automatically — you can also read it directly for quick answers.

---

## Sidecar Pattern

A helper container that runs alongside the main application container in the same pod.

**Use cases:** Log shipping, metrics collection, TLS proxy, config reload, vault agent.

```yaml
spec:
  containers:
    - name: app
      image: myapp:1.0
      ports:
        - containerPort: 3000
    - name: log-shipper
      image: fluentbit:3.2
      volumeMounts:
        - name: logs
          mountPath: /var/log/app
    - name: envoy-proxy
      image: envoyproxy/envoy:v1.32
      ports:
        - containerPort: 8080
  volumes:
    - name: logs
      emptyDir: {}
```

**Kubernetes 1.28+ native sidecar containers:**
```yaml
spec:
  initContainers:
    - name: log-shipper
      image: fluentbit:3.2
      restartPolicy: Always          # Makes it a sidecar
      volumeMounts:
        - name: logs
          mountPath: /var/log/app
  containers:
    - name: app
      image: myapp:1.0
```

Native sidecars:
- Start before main containers
- Keep running (restartPolicy: Always)
- Shut down after main containers exit
- Don't block pod readiness

---

## Init Container Pattern

Containers that run to completion before app containers start. Used for setup tasks.

**Use cases:** Database migration, config generation, waiting for dependencies, downloading assets.

```yaml
spec:
  initContainers:
    # Wait for database to be ready
    - name: wait-for-db
      image: busybox:1.37
      command: ['sh', '-c', 'until nc -z db 5432; do echo waiting for db; sleep 2; done']

    # Run database migrations
    - name: migrate
      image: myapp:1.0
      command: ['npx', 'prisma', 'migrate', 'deploy']
      env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: db-secrets
              key: url

    # Download ML model
    - name: download-model
      image: alpine:3.21
      command: ['wget', '-O', '/models/v2.bin', 'https://models.example.com/v2.bin']
      volumeMounts:
        - name: models
          mountPath: /models

  containers:
    - name: app
      image: myapp:1.0
      volumeMounts:
        - name: models
          mountPath: /models
          readOnly: true

  volumes:
    - name: models
      emptyDir: {}
```

---

## Ambassador Pattern

A proxy container that handles communication on behalf of the app. The app talks to localhost, the ambassador proxies to external services.

```yaml
spec:
  containers:
    - name: app
      image: myapp:1.0
      env:
        - name: REDIS_URL
          value: "redis://localhost:6379"
    # Redis ambassador — handles auth and TLS to managed Redis
    - name: redis-proxy
      image: haproxy:3.0-alpine
      ports:
        - containerPort: 6379
      volumeMounts:
        - name: haproxy-config
          mountPath: /usr/local/etc/haproxy/haproxy.cfg
          subPath: haproxy.cfg
  volumes:
    - name: haproxy-config
      configMap:
        name: redis-proxy-config
```

---

## Adapter Pattern

A container that transforms the output of the main container into a standardized format.

```yaml
spec:
  containers:
    - name: app
      image: legacy-app:1.0
      # Writes custom-format logs to shared volume
      volumeMounts:
        - name: logs
          mountPath: /var/log/app
    # Adapter: converts custom logs to JSON for log aggregator
    - name: log-adapter
      image: log-transformer:1.0
      volumeMounts:
        - name: logs
          mountPath: /var/log/app
          readOnly: true
  volumes:
    - name: logs
      emptyDir: {}
```

---

## Jobs and CronJobs

### One-Off Job

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: data-import
spec:
  backoffLimit: 3
  activeDeadlineSeconds: 3600
  ttlSecondsAfterFinished: 86400      # Clean up after 24h
  template:
    spec:
      restartPolicy: OnFailure
      containers:
        - name: import
          image: myapp:1.0
          command: ["node", "scripts/import.js"]
          resources:
            requests:
              cpu: 500m
              memory: 512Mi
            limits:
              cpu: "2"
              memory: 2Gi
```

### Parallel Job

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: batch-process
spec:
  completions: 10               # Total tasks
  parallelism: 3                # Run 3 at a time
  completionMode: Indexed       # Each pod gets JOB_COMPLETION_INDEX
  backoffLimit: 5
  template:
    spec:
      restartPolicy: OnFailure
      containers:
        - name: worker
          image: batch-worker:1.0
          env:
            - name: BATCH_INDEX
              valueFrom:
                fieldRef:
                  fieldPath: metadata.annotations['batch.kubernetes.io/job-completion-index']
```

### CronJob Patterns

```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: cleanup
spec:
  schedule: "0 3 * * *"           # 3 AM daily
  timeZone: "America/New_York"
  concurrencyPolicy: Forbid       # Don't run if previous still running
  startingDeadlineSeconds: 600    # Skip if >10 min late
  successfulJobsHistoryLimit: 3
  failedJobsHistoryLimit: 5
  jobTemplate:
    spec:
      backoffLimit: 2
      activeDeadlineSeconds: 1800
      template:
        spec:
          restartPolicy: OnFailure
          containers:
            - name: cleanup
              image: myapp:1.0
              command: ["node", "scripts/cleanup.js"]
```

**Concurrency policies:**
- `Allow` — Multiple jobs can run simultaneously (default)
- `Forbid` — Skip new run if previous still running
- `Replace` — Kill previous job, start new one

---

## Custom Resource Definitions (CRDs)

### Defining a CRD

```yaml
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: databases.example.com
spec:
  group: example.com
  versions:
    - name: v1
      served: true
      storage: true
      schema:
        openAPIV3Schema:
          type: object
          properties:
            spec:
              type: object
              required: ["engine", "version", "storage"]
              properties:
                engine:
                  type: string
                  enum: ["postgres", "mysql", "redis"]
                version:
                  type: string
                storage:
                  type: string
                  pattern: "^[0-9]+(Gi|Ti)$"
                replicas:
                  type: integer
                  minimum: 1
                  maximum: 7
                  default: 1
                backup:
                  type: object
                  properties:
                    enabled:
                      type: boolean
                      default: true
                    schedule:
                      type: string
                      default: "0 2 * * *"
            status:
              type: object
              properties:
                phase:
                  type: string
                endpoint:
                  type: string
                readyReplicas:
                  type: integer
      subresources:
        status: {}
        scale:
          specReplicasPath: .spec.replicas
          statusReplicasPath: .status.readyReplicas
      additionalPrinterColumns:
        - name: Engine
          type: string
          jsonPath: .spec.engine
        - name: Version
          type: string
          jsonPath: .spec.version
        - name: Phase
          type: string
          jsonPath: .status.phase
        - name: Age
          type: date
          jsonPath: .metadata.creationTimestamp
  scope: Namespaced
  names:
    plural: databases
    singular: database
    kind: Database
    shortNames: ["db"]
    categories: ["all"]
```

### Using a CR

```yaml
apiVersion: example.com/v1
kind: Database
metadata:
  name: myapp-db
  namespace: production
spec:
  engine: postgres
  version: "17"
  storage: 50Gi
  replicas: 3
  backup:
    enabled: true
    schedule: "0 2 * * *"
```

---

## Operator Pattern

Operators extend Kubernetes by watching custom resources and reconciling desired state.

### Common Operators to Know

| Operator | Purpose | CRDs |
|----------|---------|------|
| cert-manager | TLS certificates | Certificate, Issuer, ClusterIssuer |
| external-secrets | Secrets from external stores | ExternalSecret, SecretStore |
| prometheus-operator | Monitoring | ServiceMonitor, PodMonitor, PrometheusRule |
| keda | Event-driven autoscaling | ScaledObject, ScaledJob, TriggerAuthentication |
| argo-rollouts | Progressive delivery | Rollout, AnalysisTemplate |
| crossplane | Cloud infrastructure | various cloud resource CRDs |
| strimzi | Kafka on K8s | Kafka, KafkaTopic, KafkaUser |
| zalando/postgres-operator | PostgreSQL | postgresql |

### cert-manager Example

```yaml
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: admin@example.com
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
      - http01:
          ingress:
            ingressClassName: nginx
---
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: api-tls
  namespace: production
spec:
  secretName: api-tls
  issuerRef:
    name: letsencrypt-prod
    kind: ClusterIssuer
  dnsNames:
    - api.example.com
    - "*.api.example.com"
  duration: 2160h    # 90 days
  renewBefore: 360h  # 15 days before expiry
```

### Prometheus ServiceMonitor

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: api
  namespace: production
  labels:
    release: prometheus
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: api
  endpoints:
    - port: http
      path: /metrics
      interval: 30s
      scrapeTimeout: 10s
  namespaceSelector:
    matchNames:
      - production
---
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: api-alerts
  namespace: production
  labels:
    release: prometheus
spec:
  groups:
    - name: api.rules
      rules:
        - alert: HighErrorRate
          expr: |
            sum(rate(http_requests_total{service="api",status=~"5.."}[5m]))
            /
            sum(rate(http_requests_total{service="api"}[5m]))
            > 0.05
          for: 5m
          labels:
            severity: critical
          annotations:
            summary: "High error rate on API"
            description: "Error rate is {{ $value | humanizePercentage }} over 5 minutes"

        - alert: HighLatency
          expr: |
            histogram_quantile(0.99,
              sum(rate(http_request_duration_seconds_bucket{service="api"}[5m])) by (le)
            ) > 2
          for: 5m
          labels:
            severity: warning
          annotations:
            summary: "High p99 latency on API"
```

---

## Service Mesh Patterns

### Istio Traffic Management

```yaml
# Virtual Service — traffic splitting and routing
apiVersion: networking.istio.io/v1
kind: VirtualService
metadata:
  name: api
spec:
  hosts:
    - api.production.svc.cluster.local
  http:
    # Header-based routing (canary)
    - match:
        - headers:
            x-canary:
              exact: "true"
      route:
        - destination:
            host: api.production.svc.cluster.local
            subset: canary
    # Weight-based splitting
    - route:
        - destination:
            host: api.production.svc.cluster.local
            subset: stable
          weight: 95
        - destination:
            host: api.production.svc.cluster.local
            subset: canary
          weight: 5
      timeout: 30s
      retries:
        attempts: 3
        perTryTimeout: 10s
        retryOn: gateway-error,connect-failure,refused-stream
---
apiVersion: networking.istio.io/v1
kind: DestinationRule
metadata:
  name: api
spec:
  host: api.production.svc.cluster.local
  trafficPolicy:
    connectionPool:
      tcp:
        maxConnections: 100
      http:
        h2UpgradePolicy: DEFAULT
        http1MaxPendingRequests: 100
        http2MaxRequests: 1000
    outlierDetection:
      consecutive5xxErrors: 3
      interval: 10s
      baseEjectionTime: 30s
      maxEjectionPercent: 50
  subsets:
    - name: stable
      labels:
        version: stable
    - name: canary
      labels:
        version: canary
```

### Circuit Breaker Pattern

```yaml
apiVersion: networking.istio.io/v1
kind: DestinationRule
metadata:
  name: external-api
spec:
  host: external-api.example.com
  trafficPolicy:
    connectionPool:
      tcp:
        maxConnections: 50
      http:
        http1MaxPendingRequests: 50
        http2MaxRequests: 100
        maxRequestsPerConnection: 10
    outlierDetection:
      consecutive5xxErrors: 5       # Trip after 5 consecutive 5xx
      interval: 10s                 # Check every 10s
      baseEjectionTime: 30s         # Remove for 30s
      maxEjectionPercent: 100       # Can eject all endpoints
```

---

## ConfigMap and Secret Patterns

### Config Reload Without Restart

```yaml
# Mount ConfigMap as volume — K8s updates it automatically (~1 min)
spec:
  containers:
    - name: app
      volumeMounts:
        - name: config
          mountPath: /app/config
          readOnly: true
  volumes:
    - name: config
      configMap:
        name: app-config

# Combine with inotify-based reload or periodic check in app:
# fs.watch('/app/config', () => reloadConfig())
```

### Projected Volumes

```yaml
# Combine multiple sources into single volume
spec:
  containers:
    - name: app
      volumeMounts:
        - name: all-config
          mountPath: /app/config
          readOnly: true
  volumes:
    - name: all-config
      projected:
        sources:
          - configMap:
              name: app-config
              items:
                - key: config.yaml
                  path: config.yaml
          - secret:
              name: app-secrets
              items:
                - key: tls.crt
                  path: tls/cert.pem
                - key: tls.key
                  path: tls/key.pem
          - serviceAccountToken:
              path: token
              expirationSeconds: 3600
              audience: api
```

---

## Pod Lifecycle

### Graceful Shutdown Sequence

```
1. Pod marked for deletion
2. Pod removed from Service endpoints (no new traffic)
3. preStop hook runs (sleep 10 — allows in-flight LB connections to drain)
4. SIGTERM sent to PID 1
5. App handles SIGTERM, stops accepting connections, finishes in-flight requests
6. terminationGracePeriodSeconds countdown (default 30s)
7. SIGKILL if still running
8. Pod deleted
```

```yaml
spec:
  terminationGracePeriodSeconds: 60    # Total time allowed
  containers:
    - name: app
      lifecycle:
        preStop:
          exec:
            command: ["sh", "-c", "sleep 10"]  # Wait for LB to update
```

### Startup and Shutdown Probe Strategy

```yaml
# Startup probe: generous for slow-starting apps
startupProbe:
  httpGet:
    path: /health
    port: http
  failureThreshold: 30      # 30 * 2s = 60s max startup time
  periodSeconds: 2

# Liveness probe: detect deadlocks (restart pod)
livenessProbe:
  httpGet:
    path: /health
    port: http
  periodSeconds: 10
  timeoutSeconds: 3
  failureThreshold: 3        # 3 failures = restart

# Readiness probe: control traffic routing
readinessProbe:
  httpGet:
    path: /ready             # Different from /health!
    port: http
  periodSeconds: 5
  timeoutSeconds: 2
  failureThreshold: 2        # 2 failures = remove from service
```

**Key distinction:**
- `/health` — Am I alive? (can I serve any request?)
- `/ready` — Am I ready? (have dependencies connected, caches warmed?)

---

## Multi-Tenancy Patterns

### Namespace-per-Tenant

```yaml
# Create tenant namespace with quotas and policies
apiVersion: v1
kind: Namespace
metadata:
  name: tenant-acme
  labels:
    tenant: acme
    pod-security.kubernetes.io/enforce: restricted
---
apiVersion: v1
kind: ResourceQuota
metadata:
  name: tenant-quota
  namespace: tenant-acme
spec:
  hard:
    requests.cpu: "4"
    requests.memory: 8Gi
    limits.cpu: "8"
    limits.memory: 16Gi
    pods: "20"
---
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: tenant-isolation
  namespace: tenant-acme
spec:
  podSelector: {}
  policyTypes: [Ingress, Egress]
  ingress:
    - from:
        - podSelector: {}    # Only within same namespace
  egress:
    - to:
        - podSelector: {}
    - to: []
      ports:
        - port: 53
          protocol: UDP
```

---

## Label and Annotation Conventions

### Recommended Labels (Kubernetes Standard)

```yaml
metadata:
  labels:
    # Standard Kubernetes labels
    app.kubernetes.io/name: api              # Application name
    app.kubernetes.io/instance: api-prod     # Unique instance
    app.kubernetes.io/version: "1.2.3"       # App version
    app.kubernetes.io/component: backend     # Component role
    app.kubernetes.io/part-of: myapp         # Higher-level app
    app.kubernetes.io/managed-by: helm       # Management tool

    # Custom labels
    team: platform
    environment: production
    cost-center: engineering
```

### Common Annotations

```yaml
metadata:
  annotations:
    # Prometheus scraping
    prometheus.io/scrape: "true"
    prometheus.io/port: "9090"
    prometheus.io/path: "/metrics"

    # Config checksums (trigger rollout on config change)
    checksum/config: "sha256abc..."

    # Deployment info
    deploy.example.com/deployed-by: "ci/cd"
    deploy.example.com/git-sha: "abc1234"
    deploy.example.com/pipeline-url: "https://..."
```
