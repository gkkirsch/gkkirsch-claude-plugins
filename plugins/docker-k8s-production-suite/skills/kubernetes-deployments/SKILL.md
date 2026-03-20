---
name: kubernetes-deployments
description: >
  Kubernetes deployment patterns — Deployments, Services, Ingress, ConfigMaps,
  Secrets, HPA, PVC, and production-ready manifest templates.
  Triggers: "kubernetes deployment", "k8s manifest", "kubectl apply",
  "kubernetes service", "ingress", "HPA", "kubernetes scaling".
  NOT for: Docker/Dockerfiles (use dockerfile-patterns), debugging (use kubernetes-debugging).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Kubernetes Deployments

## Complete Application Manifest

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
      maxSurge: 1
      maxUnavailable: 0
  template:
    metadata:
      labels:
        app: myapp
        version: v1
    spec:
      serviceAccountName: myapp
      securityContext:
        runAsNonRoot: true
        runAsUser: 1001
        fsGroup: 1001
      containers:
        - name: myapp
          image: myregistry/myapp:1.2.3
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
              cpu: 100m
              memory: 128Mi
            limits:
              cpu: 500m
              memory: 512Mi
          readinessProbe:
            httpGet:
              path: /health
              port: 3000
            initialDelaySeconds: 5
            periodSeconds: 10
            failureThreshold: 3
          livenessProbe:
            httpGet:
              path: /health
              port: 3000
            initialDelaySeconds: 15
            periodSeconds: 20
            failureThreshold: 3
          startupProbe:
            httpGet:
              path: /health
              port: 3000
            failureThreshold: 30
            periodSeconds: 2
      terminationGracePeriodSeconds: 30
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
```

### Service Types

| Type | Access | Use Case |
|------|--------|----------|
| `ClusterIP` | Internal only | Default. Backend services, databases |
| `NodePort` | External via node IP:port | Simple testing, on-prem |
| `LoadBalancer` | External via cloud LB | Cloud-native public services |
| `ExternalName` | DNS alias | Bridging to external services |

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
```

### Multiple Services on One Ingress

```yaml
spec:
  rules:
    - host: example.com
      http:
        paths:
          - path: /api
            pathType: Prefix
            backend:
              service:
                name: api-service
                port:
                  number: 80
          - path: /
            pathType: Prefix
            backend:
              service:
                name: frontend-service
                port:
                  number: 80
```

## ConfigMap and Secrets

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
  max-connections: "100"
  app-config.json: |
    {
      "features": {
        "newCheckout": true,
        "darkMode": false
      }
    }

---
# secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: myapp-secrets
  namespace: production
type: Opaque
stringData:
  database-url: "postgresql://user:pass@db:5432/myapp"
  jwt-secret: "your-secret-key"
  stripe-key: "sk_live_..."
```

### Mount ConfigMap as File

```yaml
containers:
  - name: myapp
    volumeMounts:
      - name: config
        mountPath: /app/config
        readOnly: true
volumes:
  - name: config
    configMap:
      name: myapp-config
      items:
        - key: app-config.json
          path: config.json
```

### External Secrets (Recommended for Production)

```yaml
# Using External Secrets Operator with AWS Secrets Manager
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: myapp-secrets
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: aws-secrets-manager
    kind: ClusterSecretStore
  target:
    name: myapp-secrets
  data:
    - secretKey: database-url
      remoteRef:
        key: production/myapp
        property: database_url
    - secretKey: jwt-secret
      remoteRef:
        key: production/myapp
        property: jwt_secret
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
  minReplicas: 2
  maxReplicas: 10
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
  behavior:
    scaleUp:
      stabilizationWindowSeconds: 60
      policies:
        - type: Pods
          value: 2
          periodSeconds: 60
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
        - type: Percent
          value: 10
          periodSeconds: 60
```

## Persistent Volume Claims

```yaml
# pvc.yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: postgres-data
  namespace: production
spec:
  accessModes:
    - ReadWriteOnce
  storageClassName: gp3  # AWS EBS gp3
  resources:
    requests:
      storage: 20Gi
```

### StatefulSet for Databases

```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: postgres
  namespace: production
spec:
  serviceName: postgres
  replicas: 1
  selector:
    matchLabels:
      app: postgres
  template:
    metadata:
      labels:
        app: postgres
    spec:
      containers:
        - name: postgres
          image: postgres:16-alpine
          ports:
            - containerPort: 5432
          env:
            - name: POSTGRES_DB
              value: myapp
            - name: POSTGRES_USER
              valueFrom:
                secretKeyRef:
                  name: postgres-secrets
                  key: username
            - name: POSTGRES_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: postgres-secrets
                  key: password
            - name: PGDATA
              value: /var/lib/postgresql/data/pgdata
          volumeMounts:
            - name: postgres-data
              mountPath: /var/lib/postgresql/data
          resources:
            requests:
              cpu: 250m
              memory: 512Mi
            limits:
              cpu: 1000m
              memory: 2Gi
  volumeClaimTemplates:
    - metadata:
        name: postgres-data
      spec:
        accessModes: ["ReadWriteOnce"]
        storageClassName: gp3
        resources:
          requests:
            storage: 20Gi
```

## CronJob

```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: cleanup-expired
  namespace: production
spec:
  schedule: "0 2 * * *"  # 2 AM daily
  concurrencyPolicy: Forbid
  successfulJobsHistoryLimit: 3
  failedJobsHistoryLimit: 3
  jobTemplate:
    spec:
      activeDeadlineSeconds: 300
      backoffLimit: 2
      template:
        spec:
          restartPolicy: Never
          containers:
            - name: cleanup
              image: myregistry/myapp:1.2.3
              command: ["node", "scripts/cleanup.js"]
              env:
                - name: DATABASE_URL
                  valueFrom:
                    secretKeyRef:
                      name: myapp-secrets
                      key: database-url
              resources:
                requests:
                  cpu: 100m
                  memory: 128Mi
                limits:
                  cpu: 500m
                  memory: 256Mi
```

## Namespace + RBAC

```yaml
# namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: production
  labels:
    environment: production

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
  name: myapp-rolebinding
  namespace: production
subjects:
  - kind: ServiceAccount
    name: myapp
    namespace: production
roleRef:
  kind: Role
  name: myapp-role
  apiGroup: rbac.authorization.k8s.io
```

## Network Policy

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: myapp-network-policy
  namespace: production
spec:
  podSelector:
    matchLabels:
      app: myapp
  policyTypes:
    - Ingress
    - Egress
  ingress:
    - from:
        - namespaceSelector:
            matchLabels:
              name: ingress-nginx
      ports:
        - port: 3000
  egress:
    - to:
        - podSelector:
            matchLabels:
              app: postgres
      ports:
        - port: 5432
    - to:
        - podSelector:
            matchLabels:
              app: redis
      ports:
        - port: 6379
    - to:  # Allow DNS
        - namespaceSelector: {}
      ports:
        - port: 53
          protocol: UDP
        - port: 53
          protocol: TCP
```

## Kustomize (Multi-Environment)

```
k8s/
  base/
    kustomization.yaml
    deployment.yaml
    service.yaml
    ingress.yaml
  overlays/
    dev/
      kustomization.yaml
      replicas-patch.yaml
    staging/
      kustomization.yaml
    production/
      kustomization.yaml
      hpa.yaml
```

```yaml
# base/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - deployment.yaml
  - service.yaml
  - ingress.yaml

# overlays/production/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - ../../base
  - hpa.yaml
namespace: production
patches:
  - path: replicas-patch.yaml
images:
  - name: myregistry/myapp
    newTag: 1.2.3
```

```bash
# Apply
kubectl apply -k k8s/overlays/production
```

## Gotchas

1. **Probes are critical** — without a readinessProbe, Kubernetes sends traffic to unready pods. Without a livenessProbe, stuck processes keep running. Without a startupProbe, slow-starting apps get killed.

2. **Resource requests affect scheduling** — if you request 2Gi memory but your node only has 1Gi free, the pod stays Pending. Set requests to typical usage, not peak.

3. **Rolling updates need `maxUnavailable: 0`** — the default allows killing pods before new ones are ready. Set `maxUnavailable: 0, maxSurge: 1` for zero-downtime deploys.

4. **Secrets are base64-encoded, not encrypted** — anyone with RBAC access to Secrets can decode them. Use External Secrets Operator or Sealed Secrets for real security.

5. **ConfigMap updates don't restart pods** — if you update a ConfigMap, existing pods keep the old values. You must restart the deployment or use a checksum annotation trick.

6. **PVCs are not deleted with `kubectl delete deployment`** — persistent data survives pod and deployment deletion. PVCs must be deleted explicitly.

7. **Namespace isolation requires NetworkPolicy** — namespaces don't isolate network traffic by default. Any pod can reach any other pod across namespaces unless you create NetworkPolicies.

8. **`imagePullPolicy: Always`** is the default for `:latest` tag — use specific version tags and `imagePullPolicy: IfNotPresent` for faster deployments.
