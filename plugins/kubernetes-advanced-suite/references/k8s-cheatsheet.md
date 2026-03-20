# Kubernetes Cheatsheet

## Essential Commands

```bash
# Pod diagnostics
kubectl get pods -n <ns> -o wide
kubectl describe pod <pod> -n <ns>
kubectl logs <pod> -n <ns> --previous
kubectl logs <pod> -c <container> -f
kubectl exec -it <pod> -- /bin/sh
kubectl debug <pod> -it --image=busybox

# Resource usage
kubectl top pods -n <ns>
kubectl top nodes

# Deployments
kubectl rollout restart deployment/<name> -n <ns>
kubectl rollout status deployment/<name> -n <ns>
kubectl rollout undo deployment/<name> -n <ns>
kubectl rollout history deployment/<name> -n <ns>
kubectl scale deployment/<name> --replicas=5 -n <ns>

# Networking
kubectl get endpoints <svc> -n <ns>
kubectl port-forward svc/<name> 8080:80 -n <ns>
kubectl get events -n <ns> --sort-by=.metadata.creationTimestamp

# RBAC
kubectl auth can-i create pods -n <ns> --as <user>
kubectl auth can-i --list -n <ns>
```

## Troubleshooting Decision Tree

```
Pod not starting?
  ImagePullBackOff    -> Wrong image/tag or registry auth
  CrashLoopBackOff    -> Check: kubectl logs <pod> --previous
  Pending             -> Insufficient resources or selector mismatch
  ContainerCreating   -> Volume mount or secret issue

Pod running but no traffic?
  readinessProbe failing?  -> Fix health endpoint
  Service selector match?  -> Compare labels
  Endpoints empty?         -> kubectl get endpoints <svc>
  NetworkPolicy blocking?  -> Check ingress rules

High latency?
  CPU throttling?     -> Check limits vs usage
  HPA not scaling?    -> Check metrics-server
  DNS slow?           -> Test from inside pod
```

## Resource Guidelines

```
Web API:     requests: 100m/128Mi   limits: 500m/512Mi
Worker:      requests: 250m/256Mi   limits: 1000m/1Gi
Database:    requests: 500m/1Gi     limits: 2000m/4Gi
Cache:       requests: 100m/256Mi   limits: 500m/1Gi

Rule: requests = avg usage, limits = peak * 1.5
```

## Deployment Template (Minimal Production)

```yaml
apiVersion: apps/v1
kind: Deployment
spec:
  replicas: 3
  strategy:
    rollingUpdate: { maxUnavailable: 1, maxSurge: 1 }
  template:
    spec:
      serviceAccountName: <name>
      securityContext:
        runAsNonRoot: true
        runAsUser: 1000
      containers:
        - image: registry/app:v1.2.3  # Never :latest
          resources:
            requests: { cpu: 100m, memory: 128Mi }
            limits: { cpu: 500m, memory: 512Mi }
          readinessProbe:
            httpGet: { path: /healthz, port: 8080 }
          livenessProbe:
            httpGet: { path: /healthz, port: 8080 }
          lifecycle:
            preStop:
              exec: { command: ["/bin/sh", "-c", "sleep 5"] }
          securityContext:
            readOnlyRootFilesystem: true
            allowPrivilegeEscalation: false
            capabilities: { drop: ["ALL"] }
```

## HPA Quick Reference

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
spec:
  minReplicas: 3
  maxReplicas: 20
  behavior:
    scaleDown:
      stabilizationWindowSeconds: 300
  metrics:
    - type: Resource
      resource:
        name: cpu
        target: { type: Utilization, averageUtilization: 70 }
```

## Network Policy Patterns

```yaml
# Default deny all
spec:
  podSelector: {}
  policyTypes: [Ingress, Egress]

# Allow specific ingress
spec:
  podSelector: { matchLabels: { app: api } }
  ingress:
    - from:
        - podSelector: { matchLabels: { app: frontend } }
      ports:
        - { protocol: TCP, port: 8080 }

# Always allow DNS egress
egress:
  - to: []
    ports:
      - { protocol: UDP, port: 53 }
      - { protocol: TCP, port: 53 }
```

## RBAC Patterns

```yaml
# Read-only role
rules:
  - apiGroups: ["", "apps"]
    resources: ["pods", "deployments", "services"]
    verbs: ["get", "list", "watch"]

# Deployer role (no delete)
rules:
  - apiGroups: ["apps"]
    resources: ["deployments"]
    verbs: ["get", "list", "patch", "update"]

# Restrict to named resources
rules:
  - apiGroups: [""]
    resources: ["secrets"]
    resourceNames: ["api-secrets"]
    verbs: ["get"]
```

## Security Checklist

```
Pod Level:
  [x] runAsNonRoot: true
  [x] readOnlyRootFilesystem: true
  [x] allowPrivilegeEscalation: false
  [x] capabilities: drop ALL
  [x] No :latest tags
  [x] Resource limits set
  [x] automountServiceAccountToken: false

Namespace Level:
  [x] PSS enforce: restricted
  [x] Default deny NetworkPolicy
  [x] ResourceQuota set
  [x] LimitRange set

Cluster Level:
  [x] RBAC enabled
  [x] etcd encryption at rest
  [x] Audit logging enabled
  [x] CNI supports NetworkPolicy
```

## Probe Configuration

```yaml
# Readiness — controls traffic routing
readinessProbe:
  httpGet: { path: /healthz/ready, port: 8080 }
  initialDelaySeconds: 5
  periodSeconds: 10
  failureThreshold: 3

# Liveness — controls pod restart (NEVER check DB)
livenessProbe:
  httpGet: { path: /healthz/live, port: 8080 }
  initialDelaySeconds: 15
  periodSeconds: 20
  failureThreshold: 3

# Startup — for slow-starting apps
startupProbe:
  httpGet: { path: /healthz/live, port: 8080 }
  failureThreshold: 30
  periodSeconds: 2
```

## Autoscaling Decision Table

| Scaler | Signal | Use For |
|--------|--------|---------|
| HPA (CPU) | CPU % | Compute-bound workloads |
| HPA (custom) | RPS, queue depth | Web services, workers |
| VPA | Historical usage | Right-sizing resources |
| KEDA | External events | Event-driven, batch |
| Cluster Autoscaler | Node utilization | Node pool scaling |
| Karpenter | Pod scheduling | Fast node provisioning |

## Tools Reference

| Tool | Purpose |
|------|---------|
| kustomize | Template-free manifest overlays |
| helm | Package manager, templated charts |
| ArgoCD | GitOps continuous delivery |
| Flux | GitOps (alternative to ArgoCD) |
| cert-manager | Automated TLS certificates |
| external-secrets | Pull secrets from Vault/AWS/GCP |
| sealed-secrets | Encrypt secrets for git |
| kyverno | Policy engine (simpler than OPA) |
| trivy | Image + cluster vulnerability scanner |
| falco | Runtime threat detection |
| kube-bench | CIS benchmark compliance |

## Common Anti-Patterns

1. **No resource limits** — noisy neighbors consume all node resources
2. **:latest tag** — can't rollback, can't reproduce
3. **DB in livenessProbe** — DB down = all pods restart = thundering herd
4. **Single replica** — zero availability during deploys
5. **kubectl apply from laptop** — no audit trail, use GitOps
6. **Privileged containers** — container escape = cluster compromise
7. **Default ServiceAccount** — too many permissions by default
8. **Wildcards in RBAC** — `"*"` grants access to future resources too
9. **Plain K8s Secrets in git** — base64 != encryption
10. **Missing PDB** — cluster upgrade drains all pods at once
