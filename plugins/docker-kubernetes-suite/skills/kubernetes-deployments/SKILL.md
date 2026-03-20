---
name: kubernetes-deployments
description: >
  Kubernetes deployment patterns — Deployments, Services, Ingress,
  HPA autoscaling, ConfigMaps, Secrets, health probes, and Helm charts.
  Triggers: "kubernetes deployment", "k8s deploy", "kubectl", "helm chart",
  "kubernetes service", "ingress", "hpa autoscaling", "kubernetes secrets",
  "kubernetes health check", "pod", "kubernetes namespace".
  NOT for: Dockerfile optimization (use dockerfile-optimization), Docker Compose.
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Kubernetes Deployments

## Deployment

```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
  namespace: production
  labels:
    app: myapp
    version: v1
spec:
  replicas: 3
  selector:
    matchLabels:
      app: myapp
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1          # max pods above desired during update
      maxUnavailable: 0    # zero-downtime: never remove before new is ready
  template:
    metadata:
      labels:
        app: myapp
        version: v1
    spec:
      serviceAccountName: myapp
      terminationGracePeriodSeconds: 30
      containers:
        - name: myapp
          image: myregistry/myapp:1.2.3  # always pin version, never :latest
          ports:
            - containerPort: 3000
              protocol: TCP
          env:
            - name: NODE_ENV
              value: "production"
            - name: DATABASE_URL
              valueFrom:
                secretKeyRef:
                  name: myapp-secrets
                  key: database-url
            - name: REDIS_URL
              valueFrom:
                configMapKeyRef:
                  name: myapp-config
                  key: redis-url
          resources:
            requests:
              cpu: "100m"       # 0.1 CPU cores
              memory: "128Mi"
            limits:
              cpu: "500m"
              memory: "512Mi"
          livenessProbe:
            httpGet:
              path: /health/live
              port: 3000
            initialDelaySeconds: 10
            periodSeconds: 15
            timeoutSeconds: 3
            failureThreshold: 3
          readinessProbe:
            httpGet:
              path: /health/ready
              port: 3000
            initialDelaySeconds: 5
            periodSeconds: 5
            timeoutSeconds: 2
            failureThreshold: 3
          startupProbe:
            httpGet:
              path: /health/live
              port: 3000
            initialDelaySeconds: 0
            periodSeconds: 5
            failureThreshold: 30  # 30 x 5s = 150s max startup time
          lifecycle:
            preStop:
              exec:
                command: ["/bin/sh", "-c", "sleep 5"]  # drain connections
      topologySpreadConstraints:
        - maxSkew: 1
          topologyKey: kubernetes.io/hostname
          whenUnsatisfiable: DoNotSchedule
          labelSelector:
            matchLabels:
              app: myapp
```

## Service

```yaml
# service.yaml
apiVersion: v1
kind: Service
metadata:
  name: myapp
  namespace: production
spec:
  type: ClusterIP
  selector:
    app: myapp
  ports:
    - port: 80
      targetPort: 3000
      protocol: TCP

---
# For external access without ingress
apiVersion: v1
kind: Service
metadata:
  name: myapp-external
spec:
  type: LoadBalancer
  selector:
    app: myapp
  ports:
    - port: 443
      targetPort: 3000
```

## Ingress

```yaml
# ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: myapp
  namespace: production
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/rate-limit: "100"
    nginx.ingress.kubernetes.io/rate-limit-window: "1m"
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/proxy-body-size: "10m"
spec:
  ingressClassName: nginx
  tls:
    - hosts:
        - api.example.com
      secretName: api-tls
  rules:
    - host: api.example.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: myapp
                port:
                  number: 80
          - path: /api/v2
            pathType: Prefix
            backend:
              service:
                name: myapp-v2
                port:
                  number: 80
```

## Horizontal Pod Autoscaler (HPA)

```yaml
# hpa.yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: myapp
  namespace: production
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: myapp
  minReplicas: 3
  maxReplicas: 20
  behavior:
    scaleUp:
      stabilizationWindowSeconds: 60    # wait 60s before scaling up again
      policies:
        - type: Pods
          value: 4
          periodSeconds: 60             # add max 4 pods per minute
    scaleDown:
      stabilizationWindowSeconds: 300   # wait 5 min before scaling down
      policies:
        - type: Percent
          value: 25
          periodSeconds: 60             # remove max 25% per minute
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 70
    - type: Resource
      resource:
        name: memory
        target:
          type: Utilization
          averageUtilization: 80
    - type: Pods
      pods:
        metric:
          name: http_requests_per_second
        target:
          type: AverageValue
          averageValue: "1000"
```

## ConfigMap & Secrets

```yaml
# configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: myapp-config
  namespace: production
data:
  redis-url: "redis://redis:6379"
  log-level: "info"
  feature-flags: |
    {
      "new-dashboard": true,
      "beta-api": false
    }

---
# secret.yaml (values are base64-encoded)
apiVersion: v1
kind: Secret
metadata:
  name: myapp-secrets
  namespace: production
type: Opaque
data:
  database-url: cG9zdGdyZXNxbDovL3VzZXI6cGFzc0Bob3N0OjU0MzIvZGI=
  jwt-secret: bXktc3VwZXItc2VjcmV0LWtleQ==
```

```bash
# Create secret from CLI
kubectl create secret generic myapp-secrets \
  --from-literal=database-url='postgresql://user:pass@host:5432/db' \
  --from-literal=jwt-secret='my-super-secret-key' \
  -n production

# Create from file
kubectl create secret generic tls-certs \
  --from-file=cert.pem \
  --from-file=key.pem \
  -n production
```

## Namespace & RBAC

```yaml
# namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: production
  labels:
    env: production

---
# service-account.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: myapp
  namespace: production

---
# role.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: myapp-role
  namespace: production
rules:
  - apiGroups: [""]
    resources: ["configmaps", "secrets"]
    verbs: ["get", "list"]

---
# rolebinding.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: myapp-binding
  namespace: production
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: myapp-role
subjects:
  - kind: ServiceAccount
    name: myapp
    namespace: production
```

## Helm Chart

```bash
# Create a new chart
helm create myapp

# Install
helm install myapp ./charts/myapp -n production -f values-prod.yaml

# Upgrade
helm upgrade myapp ./charts/myapp -n production -f values-prod.yaml

# Rollback
helm rollback myapp 1 -n production

# Template (dry run)
helm template myapp ./charts/myapp -f values-prod.yaml
```

```yaml
# charts/myapp/values.yaml
replicaCount: 3

image:
  repository: myregistry/myapp
  tag: "1.2.3"
  pullPolicy: IfNotPresent

service:
  type: ClusterIP
  port: 80
  targetPort: 3000

ingress:
  enabled: true
  className: nginx
  hosts:
    - host: api.example.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: api-tls
      hosts:
        - api.example.com

resources:
  requests:
    cpu: 100m
    memory: 128Mi
  limits:
    cpu: 500m
    memory: 512Mi

autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 20
  targetCPUUtilizationPercentage: 70

env:
  NODE_ENV: production

secrets:
  DATABASE_URL: ""
  JWT_SECRET: ""
```

## kubectl Quick Reference

```bash
# Pods
kubectl get pods -n production
kubectl describe pod myapp-abc123 -n production
kubectl logs myapp-abc123 -n production --tail=100 -f
kubectl exec -it myapp-abc123 -n production -- /bin/sh

# Deployments
kubectl get deployments -n production
kubectl rollout status deployment/myapp -n production
kubectl rollout undo deployment/myapp -n production
kubectl scale deployment myapp --replicas=5 -n production
kubectl set image deployment/myapp myapp=myregistry/myapp:1.2.4 -n production

# Debug
kubectl get events -n production --sort-by='.lastTimestamp'
kubectl top pods -n production
kubectl top nodes
kubectl describe node <node-name>

# Port forward (local debugging)
kubectl port-forward svc/myapp 3000:80 -n production
kubectl port-forward pod/myapp-abc123 3000:3000 -n production

# Apply manifests
kubectl apply -f k8s/ -n production
kubectl apply -f k8s/deployment.yaml -n production
kubectl diff -f k8s/ -n production  # preview changes
```

## Health Endpoint Implementation

```typescript
import express from "express";

const app = express();

// Liveness: is the process alive and not deadlocked?
app.get("/health/live", (req, res) => {
  res.status(200).json({ status: "alive" });
});

// Readiness: can the process handle requests?
app.get("/health/ready", async (req, res) => {
  try {
    // Check database connection
    await db.$queryRaw`SELECT 1`;
    // Check Redis connection
    await redis.ping();

    res.status(200).json({
      status: "ready",
      checks: {
        database: "connected",
        redis: "connected",
      },
    });
  } catch (error) {
    res.status(503).json({
      status: "not ready",
      error: error instanceof Error ? error.message : "Unknown error",
    });
  }
});

// Startup: has the app finished initializing?
let isStarted = false;
app.get("/health/startup", (req, res) => {
  if (isStarted) {
    res.status(200).json({ status: "started" });
  } else {
    res.status(503).json({ status: "starting" });
  }
});

// Mark as started after initialization
async function start() {
  await db.$connect();
  await runMigrations();
  await loadConfig();
  isStarted = true;
  app.listen(3000);
}
```

## Gotchas

1. **Never use `:latest` tag in production.** Image tags are mutable — `:latest` can point to different images on different nodes. Always pin to a specific version or SHA digest. `imagePullPolicy: Always` doesn't help if nodes have a cached stale `:latest`.

2. **Set resource requests AND limits.** Without requests, the scheduler doesn't know where to place pods. Without limits, a misbehaving pod consumes all node resources. Memory limit exceeded = OOMKilled. CPU limit exceeded = throttled.

3. **Liveness probes should be lightweight.** A liveness probe that queries the database means a database outage restarts ALL pods (making things worse). Liveness = "is the process alive?" Readiness = "can it handle traffic?"

4. **`preStop` hook with `sleep` prevents dropped connections.** When a pod terminates, the Service endpoint is removed and in-flight requests complete. But there's a race condition — endpoints update is async. A short sleep in `preStop` gives the system time to stop routing new traffic.

5. **HPA `stabilizationWindowSeconds` prevents thrashing.** Without stabilization, the autoscaler rapidly scales up and down ("flapping"). Set scaleDown stabilization to 5+ minutes. ScaleUp can be more aggressive (30-60s).

6. **Secrets are base64-encoded, not encrypted.** Anyone with RBAC access to secrets can read them. Use external secret managers (AWS Secrets Manager, Vault) with the External Secrets Operator for real security.
