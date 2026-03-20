---
name: k8s-operations
description: >
  Kubernetes operations and deployment patterns for production clusters.
  Use when deploying applications, managing rollouts, configuring autoscaling,
  setting up monitoring, or troubleshooting Kubernetes issues.
  Triggers: "kubernetes", "k8s", "kubectl", "deployment", "pod", "service",
  "ingress", "helm", "kustomize", "HPA", "GitOps", "ArgoCD".
  NOT for: Docker-only workflows, cloud-specific services without K8s, basic container builds.
version: 1.0.0
allowed-tools: Read, Grep, Glob, Bash
---

# Kubernetes Operations

## Production Deployment Manifest

```yaml
# deployment.yaml — production-ready template
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-server
  namespace: production
  labels:
    app: api-server
    version: v1.2.3
spec:
  replicas: 3
  revisionHistoryLimit: 5
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 1
      maxSurge: 1
  selector:
    matchLabels:
      app: api-server
  template:
    metadata:
      labels:
        app: api-server
        version: v1.2.3
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "9090"
    spec:
      serviceAccountName: api-server
      terminationGracePeriodSeconds: 30
      securityContext:
        runAsNonRoot: true
        runAsUser: 1000
        fsGroup: 1000
      containers:
        - name: api
          image: registry.example.com/api-server:v1.2.3  # Never use :latest
          imagePullPolicy: IfNotPresent
          ports:
            - containerPort: 8080
              name: http
            - containerPort: 9090
              name: metrics
          env:
            - name: NODE_ENV
              value: "production"
            - name: DATABASE_URL
              valueFrom:
                secretKeyRef:
                  name: api-secrets
                  key: database-url
            - name: POD_NAME
              valueFrom:
                fieldRef:
                  fieldPath: metadata.name
          resources:
            requests:
              cpu: 100m
              memory: 128Mi
            limits:
              cpu: 500m
              memory: 512Mi
          readinessProbe:
            httpGet:
              path: /healthz/ready
              port: http
            initialDelaySeconds: 5
            periodSeconds: 10
            failureThreshold: 3
          livenessProbe:
            httpGet:
              path: /healthz/live
              port: http
            initialDelaySeconds: 15
            periodSeconds: 20
            failureThreshold: 3
          startupProbe:
            httpGet:
              path: /healthz/live
              port: http
            failureThreshold: 30
            periodSeconds: 2
          lifecycle:
            preStop:
              exec:
                command: ["/bin/sh", "-c", "sleep 5"]  # Drain connections
          securityContext:
            readOnlyRootFilesystem: true
            allowPrivilegeEscalation: false
            capabilities:
              drop: ["ALL"]
          volumeMounts:
            - name: tmp
              mountPath: /tmp
      volumes:
        - name: tmp
          emptyDir: {}
      affinity:
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
            - weight: 100
              podAffinityTerm:
                labelSelector:
                  matchExpressions:
                    - key: app
                      operator: In
                      values: ["api-server"]
                topologyKey: kubernetes.io/hostname
      topologySpreadConstraints:
        - maxSkew: 1
          topologyKey: topology.kubernetes.io/zone
          whenUnsatisfiable: DoNotSchedule
          labelSelector:
            matchLabels:
              app: api-server
```

## Service & Ingress

```yaml
# service.yaml
apiVersion: v1
kind: Service
metadata:
  name: api-server
  namespace: production
spec:
  type: ClusterIP
  selector:
    app: api-server
  ports:
    - name: http
      port: 80
      targetPort: http
      protocol: TCP
---
# ingress.yaml (nginx ingress controller)
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: api-server
  namespace: production
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/rate-limit: "100"
    nginx.ingress.kubernetes.io/rate-limit-window: "1m"
    nginx.ingress.kubernetes.io/proxy-body-size: "10m"
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
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
                name: api-server
                port:
                  name: http
```

## Autoscaling

```yaml
# HPA — Horizontal Pod Autoscaler
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: api-server
  namespace: production
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: api-server
  minReplicas: 3
  maxReplicas: 20
  behavior:
    scaleUp:
      stabilizationWindowSeconds: 60
      policies:
        - type: Percent
          value: 50
          periodSeconds: 60
    scaleDown:
      stabilizationWindowSeconds: 300  # 5 min cooldown
      policies:
        - type: Percent
          value: 25
          periodSeconds: 60
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
          averageValue: "100"
---
# PodDisruptionBudget — protect availability during updates
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: api-server
  namespace: production
spec:
  minAvailable: 2
  selector:
    matchLabels:
      app: api-server
---
# VPA — Vertical Pod Autoscaler (right-size resources)
apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: api-server
  namespace: production
spec:
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: api-server
  updatePolicy:
    updateMode: "Off"  # "Off" = recommendations only, "Auto" = auto-resize
  resourcePolicy:
    containerPolicies:
      - containerName: api
        minAllowed:
          cpu: 50m
          memory: 64Mi
        maxAllowed:
          cpu: 2000m
          memory: 4Gi
```

## KEDA (Event-Driven Autoscaling)

```yaml
# ScaledObject for queue-based worker
apiVersion: keda.sh/v1alpha1
kind: ScaledObject
metadata:
  name: order-processor
  namespace: production
spec:
  scaleTargetRef:
    name: order-processor
  minReplicaCount: 1
  maxReplicaCount: 50
  pollingInterval: 15
  cooldownPeriod: 300
  triggers:
    - type: rabbitmq
      metadata:
        host: amqp://rabbitmq.production:5672
        queueName: orders
        queueLength: "10"
    - type: cron
      metadata:
        timezone: America/New_York
        start: "0 8 * * *"
        end: "0 20 * * *"
        desiredReplicas: "5"
```

## ConfigMaps & Secrets

```yaml
# ConfigMap — non-sensitive configuration
apiVersion: v1
kind: ConfigMap
metadata:
  name: api-config
  namespace: production
data:
  APP_LOG_LEVEL: "info"
  APP_CACHE_TTL: "300"
---
# ExternalSecret — pull from AWS Secrets Manager / Vault
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: api-secrets
  namespace: production
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: aws-secrets-manager
    kind: ClusterSecretStore
  target:
    name: api-secrets
    creationPolicy: Owner
  data:
    - secretKey: database-url
      remoteRef:
        key: production/api/database-url
    - secretKey: api-key
      remoteRef:
        key: production/api/external-api-key
```

## Troubleshooting Commands

```bash
# Pod diagnostics
kubectl get pods -n production -o wide
kubectl describe pod <name> -n production
kubectl logs <pod> -n production --previous
kubectl logs <pod> -n production -f
kubectl logs <pod> -c <container>

# Resource usage
kubectl top pods -n production
kubectl top nodes

# Network debugging
kubectl exec -it <pod> -- curl localhost:8080/healthz
kubectl exec -it <pod> -- nslookup api-server.production.svc.cluster.local
kubectl get endpoints api-server -n production
kubectl port-forward svc/api-server 8080:80 -n production

# Events
kubectl get events -n production --sort-by=.metadata.creationTimestamp | tail -20

# Rolling restart and rollback
kubectl rollout restart deployment/api-server -n production
kubectl rollout status deployment/api-server -n production
kubectl rollout undo deployment/api-server -n production
kubectl rollout history deployment/api-server -n production

# Debug container (ephemeral)
kubectl debug <pod> -it --image=busybox --target=api -n production
```

## Kustomize

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
namespace: production
patches:
  - target:
      kind: Deployment
      name: api-server
    patch: |
      - op: replace
        path: /spec/replicas
        value: 5
images:
  - name: registry.example.com/api-server
    newTag: v1.2.3
```

```bash
kubectl apply -k overlays/production/
```

## GitOps with ArgoCD

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: api-server
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://github.com/org/k8s-manifests.git
    targetRevision: main
    path: overlays/production
  destination:
    server: https://kubernetes.default.svc
    namespace: production
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
      - CreateNamespace=true
    retry:
      limit: 5
      backoff:
        duration: 5s
        factor: 2
        maxDuration: 3m
```

## Helm

```bash
helm install api-server ./chart -f values-production.yaml -n production
helm upgrade api-server ./chart -f values-production.yaml -n production
helm rollback api-server 1 -n production
helm history api-server -n production
```

## Gotchas

1. **No resource limits = noisy neighbor problem** — a single pod without limits can consume all node CPU/memory. Always set both requests AND limits. Start with VPA in "Off" mode for recommendations.

2. **livenessProbe checking database** — if your DB goes down, liveness fails, all pods restart simultaneously, thundering herd on DB recovery. Use readinessProbe for dependency checks, livenessProbe only for "is the process alive."

3. **Missing preStop hook** — pods receive SIGTERM and immediately stop serving. But the Endpoints update takes a few seconds. Add `sleep 5` in preStop so in-flight requests complete before the pod is removed from the service.

4. **Using `latest` tag** — can't rollback, can't audit, can't reproduce. Use immutable tags (git SHA or semver). Set `imagePullPolicy: IfNotPresent` with versioned tags.

5. **HPA + VPA conflict** — both try to adjust resources. Use VPA in "Off" or "Initial" mode with HPA. Never use VPA "Auto" mode on the same dimension HPA scales.

6. **PDB too restrictive during cluster upgrades** — `maxUnavailable: 0` means nodes can never drain. Always allow at least 1 unavailable, or use `minAvailable: N-1`.
