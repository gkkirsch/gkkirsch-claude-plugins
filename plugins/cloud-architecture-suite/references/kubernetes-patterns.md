# Kubernetes Patterns Reference

Comprehensive reference for Kubernetes deployment patterns, operators, CRDs, networking, storage, security, and operational patterns.

---

## Core Workload Patterns

### Pattern 1: Sidecar

A helper container that extends the main container's functionality.

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: app-with-sidecar
spec:
  containers:
    # Main application container
    - name: app
      image: myapp:1.0
      ports:
        - containerPort: 8080
      volumeMounts:
        - name: logs
          mountPath: /var/log/app

    # Sidecar: log collector
    - name: log-collector
      image: fluentbit:latest
      volumeMounts:
        - name: logs
          mountPath: /var/log/app
          readOnly: true
        - name: fluentbit-config
          mountPath: /fluent-bit/etc/

  volumes:
    - name: logs
      emptyDir: {}
    - name: fluentbit-config
      configMap:
        name: fluentbit-config
```

**Common Sidecar Use Cases:**
- Log collection (Fluentbit, Filebeat)
- Service mesh proxy (Envoy via Istio/Linkerd)
- Authentication proxy (OAuth2 Proxy)
- TLS termination
- Monitoring agent (Datadog, New Relic)
- Configuration reloading

### Pattern 2: Ambassador

A proxy container that simplifies access to external services.

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: app-with-ambassador
spec:
  containers:
    - name: app
      image: myapp:1.0
      env:
        - name: DB_HOST
          value: "localhost"
        - name: DB_PORT
          value: "5432"

    # Ambassador: database proxy with connection pooling
    - name: cloud-sql-proxy
      image: gcr.io/cloud-sql-connectors/cloud-sql-proxy:2.8.0
      args:
        - "--structured-logs"
        - "--port=5432"
        - "project:region:instance"
      securityContext:
        runAsNonRoot: true
      resources:
        requests:
          memory: 64Mi
          cpu: 50m
```

### Pattern 3: Adapter

Transforms the main container's output into a standard format.

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: app-with-adapter
spec:
  containers:
    - name: app
      image: legacy-app:1.0
      ports:
        - containerPort: 8080

    # Adapter: converts app's custom metrics to Prometheus format
    - name: metrics-adapter
      image: prometheus-adapter:1.0
      ports:
        - containerPort: 9090
          name: metrics
      env:
        - name: APP_METRICS_URL
          value: "http://localhost:8080/internal/stats"
```

### Pattern 4: Init Container

Run setup tasks before the main containers start.

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: app-with-init
spec:
  initContainers:
    # Wait for database to be ready
    - name: wait-for-db
      image: busybox:1.36
      command: ['sh', '-c', 'until nc -z db-service 5432; do echo waiting for db; sleep 2; done']

    # Run database migrations
    - name: migrate
      image: myapp:1.0
      command: ['node', 'migrate.js']
      env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: db-credentials
              key: url

    # Download configuration from remote source
    - name: fetch-config
      image: curlimages/curl:8.5.0
      command: ['sh', '-c', 'curl -o /config/app.yaml https://config-server/app.yaml']
      volumeMounts:
        - name: config
          mountPath: /config

  containers:
    - name: app
      image: myapp:1.0
      volumeMounts:
        - name: config
          mountPath: /config
          readOnly: true

  volumes:
    - name: config
      emptyDir: {}
```

---

## Deployment Strategies

### Strategy 1: Rolling Update

```yaml
spec:
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 25%   # Max pods that can be unavailable
      maxSurge: 25%          # Max pods above desired count

# Timeline:
# t0: [v1] [v1] [v1] [v1]         (4 pods, desired=4)
# t1: [v1] [v1] [v1] [v1] [v2]   (surge: +1 new)
# t2: [v1] [v1] [v1] [v2] [v2]   (terminate 1 old, start 1 new)
# t3: [v1] [v1] [v2] [v2] [v2]
# t4: [v1] [v2] [v2] [v2] [v2]
# t5: [v2] [v2] [v2] [v2]         (complete)
```

### Strategy 2: Recreate

```yaml
spec:
  strategy:
    type: Recreate

# Timeline:
# t0: [v1] [v1] [v1] [v1]   (all running)
# t1: [  ] [  ] [  ] [  ]   (all terminated — DOWNTIME)
# t2: [v2] [v2] [v2] [v2]   (all new started)
```

Use for: Stateful apps that can't run two versions simultaneously, dev environments.

### Strategy 3: Blue-Green (via Service selector)

```yaml
# Blue deployment (current)
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app-blue
spec:
  replicas: 4
  selector:
    matchLabels:
      app: myapp
      version: blue
  template:
    metadata:
      labels:
        app: myapp
        version: blue
    spec:
      containers:
        - name: app
          image: myapp:1.0

---
# Green deployment (new)
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app-green
spec:
  replicas: 4
  selector:
    matchLabels:
      app: myapp
      version: green
  template:
    metadata:
      labels:
        app: myapp
        version: green
    spec:
      containers:
        - name: app
          image: myapp:2.0

---
# Service — switch selector to route traffic
apiVersion: v1
kind: Service
metadata:
  name: myapp
spec:
  selector:
    app: myapp
    version: blue    # Change to "green" to switch
  ports:
    - port: 80
      targetPort: 8080
```

### Strategy 4: Canary (via Ingress weight)

```yaml
# Stable ingress (95% traffic)
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: myapp-stable
  annotations:
    nginx.ingress.kubernetes.io/canary: "false"
spec:
  rules:
    - host: myapp.example.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: myapp-stable
                port:
                  number: 80

---
# Canary ingress (5% traffic)
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: myapp-canary
  annotations:
    nginx.ingress.kubernetes.io/canary: "true"
    nginx.ingress.kubernetes.io/canary-weight: "5"
spec:
  rules:
    - host: myapp.example.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: myapp-canary
                port:
                  number: 80
```

---

## Scaling Patterns

### Vertical Pod Autoscaler (VPA)

```yaml
apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: api-vpa
spec:
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: api
  updatePolicy:
    updateMode: "Auto"  # "Off" = recommendations only, "Auto" = auto-restart
  resourcePolicy:
    containerPolicies:
      - containerName: api
        minAllowed:
          cpu: 100m
          memory: 128Mi
        maxAllowed:
          cpu: 4
          memory: 8Gi
        controlledResources: ["cpu", "memory"]
```

### KEDA (Kubernetes Event-Driven Autoscaling)

```yaml
apiVersion: keda.sh/v1alpha1
kind: ScaledObject
metadata:
  name: worker-scaler
spec:
  scaleTargetRef:
    name: worker
  minReplicaCount: 0    # Scale to zero!
  maxReplicaCount: 50
  pollingInterval: 15
  cooldownPeriod: 60
  triggers:
    # Scale based on SQS queue depth
    - type: aws-sqs-queue
      metadata:
        queueURL: https://sqs.us-east-1.amazonaws.com/123456789012/work-queue
        queueLength: "5"
        awsRegion: us-east-1
      authenticationRef:
        name: keda-aws-credentials

    # Scale based on Prometheus metric
    - type: prometheus
      metadata:
        serverAddress: http://prometheus.monitoring:9090
        metricName: http_requests_per_second
        query: sum(rate(http_requests_total{service="worker"}[2m]))
        threshold: "100"
```

### Cluster Autoscaler vs Karpenter

| Feature | Cluster Autoscaler | Karpenter |
|---------|-------------------|-----------|
| Scheduling | Uses node groups | Direct provisioning |
| Speed | Minutes | Seconds |
| Flexibility | Fixed instance types per group | Dynamic instance selection |
| Spot | Per node group | Automatic diversification |
| Consolidation | No | Yes (bin-packing) |
| Disruption | Manual | Automated drift, expiry |

**Karpenter NodePool:**

```yaml
apiVersion: karpenter.sh/v1beta1
kind: NodePool
metadata:
  name: default
spec:
  template:
    metadata:
      labels:
        managed-by: karpenter
    spec:
      requirements:
        - key: kubernetes.io/arch
          operator: In
          values: ["amd64", "arm64"]
        - key: karpenter.sh/capacity-type
          operator: In
          values: ["spot", "on-demand"]
        - key: karpenter.k8s.aws/instance-category
          operator: In
          values: ["m", "c", "r"]
        - key: karpenter.k8s.aws/instance-generation
          operator: Gt
          values: ["5"]
        - key: karpenter.k8s.aws/instance-size
          operator: In
          values: ["large", "xlarge", "2xlarge"]
      nodeClassRef:
        name: default

  limits:
    cpu: 1000       # Max 1000 vCPUs across all nodes
    memory: 4000Gi  # Max 4000 Gi memory

  disruption:
    consolidationPolicy: WhenUnderutilized
    expireAfter: 720h  # Replace nodes after 30 days

---
apiVersion: karpenter.k8s.aws/v1beta1
kind: EC2NodeClass
metadata:
  name: default
spec:
  amiFamily: AL2023
  subnetSelectorTerms:
    - tags:
        karpenter.sh/discovery: production
  securityGroupSelectorTerms:
    - tags:
        karpenter.sh/discovery: production
  blockDeviceMappings:
    - deviceName: /dev/xvda
      ebs:
        volumeSize: 50Gi
        volumeType: gp3
        encrypted: true
```

---

## Storage Patterns

### StatefulSet with PVC

```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: postgres
spec:
  serviceName: postgres
  replicas: 3
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
          image: postgres:16
          ports:
            - containerPort: 5432
          env:
            - name: POSTGRES_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: postgres-credentials
                  key: password
            - name: PGDATA
              value: /var/lib/postgresql/data/pgdata
          volumeMounts:
            - name: data
              mountPath: /var/lib/postgresql/data
          resources:
            requests:
              cpu: 500m
              memory: 1Gi
            limits:
              cpu: "2"
              memory: 4Gi

  volumeClaimTemplates:
    - metadata:
        name: data
      spec:
        accessModes: ["ReadWriteOnce"]
        storageClassName: gp3-encrypted
        resources:
          requests:
            storage: 100Gi
```

**StorageClass:**

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: gp3-encrypted
provisioner: ebs.csi.aws.com
parameters:
  type: gp3
  iops: "3000"
  throughput: "125"
  encrypted: "true"
  kmsKeyId: arn:aws:kms:us-east-1:123456789012:key/mrk-xxx
reclaimPolicy: Retain
allowVolumeExpansion: true
volumeBindingMode: WaitForFirstConsumer
```

---

## Networking Patterns

### Ingress with TLS

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: myapp
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/rate-limit: "100"
    nginx.ingress.kubernetes.io/rate-limit-window: "1m"
    nginx.ingress.kubernetes.io/proxy-body-size: "10m"
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/configuration-snippet: |
      more_set_headers "X-Frame-Options: DENY";
      more_set_headers "X-Content-Type-Options: nosniff";
      more_set_headers "X-XSS-Protection: 1; mode=block";
spec:
  ingressClassName: nginx
  tls:
    - hosts:
        - myapp.example.com
        - api.example.com
      secretName: myapp-tls
  rules:
    - host: myapp.example.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: frontend
                port:
                  number: 80
    - host: api.example.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: api
                port:
                  number: 80
```

### Gateway API (Kubernetes-native, successor to Ingress)

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: production
spec:
  gatewayClassName: nginx
  listeners:
    - name: https
      protocol: HTTPS
      port: 443
      tls:
        mode: Terminate
        certificateRefs:
          - name: wildcard-tls

---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: api-routes
spec:
  parentRefs:
    - name: production
  hostnames:
    - "api.example.com"
  rules:
    - matches:
        - path:
            type: PathPrefix
            value: /v1
      backendRefs:
        - name: api-v1
          port: 80
          weight: 95
        - name: api-v2
          port: 80
          weight: 5  # Canary
    - matches:
        - path:
            type: PathPrefix
            value: /v2
      backendRefs:
        - name: api-v2
          port: 80
```

### Service Types

```yaml
# ClusterIP — internal only (default)
apiVersion: v1
kind: Service
metadata:
  name: api
spec:
  type: ClusterIP
  selector:
    app: api
  ports:
    - port: 80
      targetPort: 8080

---
# NodePort — expose on every node (testing/dev)
apiVersion: v1
kind: Service
metadata:
  name: api-nodeport
spec:
  type: NodePort
  selector:
    app: api
  ports:
    - port: 80
      targetPort: 8080
      nodePort: 30080    # 30000-32767

---
# LoadBalancer — cloud provider LB
apiVersion: v1
kind: Service
metadata:
  name: api-lb
  annotations:
    service.beta.kubernetes.io/aws-load-balancer-type: "nlb"
    service.beta.kubernetes.io/aws-load-balancer-scheme: "internet-facing"
spec:
  type: LoadBalancer
  selector:
    app: api
  ports:
    - port: 443
      targetPort: 8080

---
# Headless — DNS-based service discovery (StatefulSets)
apiVersion: v1
kind: Service
metadata:
  name: postgres
spec:
  clusterIP: None
  selector:
    app: postgres
  ports:
    - port: 5432
# DNS: postgres-0.postgres.namespace.svc.cluster.local
```

---

## Security Patterns

### RBAC (Role-Based Access Control)

```yaml
# Namespace-scoped Role
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: developer
  namespace: staging
rules:
  - apiGroups: [""]
    resources: ["pods", "pods/log", "pods/exec", "services", "configmaps"]
    verbs: ["get", "list", "watch"]
  - apiGroups: ["apps"]
    resources: ["deployments", "replicasets"]
    verbs: ["get", "list", "watch"]
  - apiGroups: [""]
    resources: ["pods/portforward"]
    verbs: ["create"]

---
# Bind role to user
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: developer-binding
  namespace: staging
subjects:
  - kind: Group
    name: developers
    apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: Role
  name: developer
  apiGroup: rbac.authorization.k8s.io

---
# Cluster-scoped for read-only across all namespaces
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: cluster-reader
rules:
  - apiGroups: [""]
    resources: ["namespaces", "nodes", "persistentvolumes"]
    verbs: ["get", "list", "watch"]
  - apiGroups: ["apps"]
    resources: ["deployments", "statefulsets", "daemonsets"]
    verbs: ["get", "list", "watch"]
```

### External Secrets Operator

```yaml
apiVersion: external-secrets.io/v1beta1
kind: SecretStore
metadata:
  name: aws-secrets
spec:
  provider:
    aws:
      service: SecretsManager
      region: us-east-1
      auth:
        jwt:
          serviceAccountRef:
            name: external-secrets

---
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: db-credentials
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: aws-secrets
    kind: SecretStore
  target:
    name: db-credentials
    creationPolicy: Owner
  data:
    - secretKey: url
      remoteRef:
        key: production/database
        property: connection_string
    - secretKey: password
      remoteRef:
        key: production/database
        property: password
```

---

## Observability Patterns

### Prometheus Stack (kube-prometheus-stack)

```yaml
# Helm values for kube-prometheus-stack
prometheus:
  prometheusSpec:
    retention: 30d
    retentionSize: 50GB
    storageSpec:
      volumeClaimTemplate:
        spec:
          storageClassName: gp3-encrypted
          resources:
            requests:
              storage: 100Gi
    resources:
      requests:
        cpu: 500m
        memory: 2Gi
      limits:
        cpu: "2"
        memory: 8Gi

grafana:
  adminPassword: "${GRAFANA_ADMIN_PASSWORD}"
  persistence:
    enabled: true
    size: 10Gi
  dashboardProviders:
    dashboardproviders.yaml:
      apiVersion: 1
      providers:
        - name: default
          orgId: 1
          folder: ''
          type: file
          disableDeletion: false
          editable: true
          options:
            path: /var/lib/grafana/dashboards/default

alertmanager:
  config:
    route:
      receiver: slack
      group_by: ['alertname', 'namespace']
      group_wait: 30s
      group_interval: 5m
      repeat_interval: 4h
      routes:
        - receiver: pagerduty
          match:
            severity: critical
        - receiver: slack
          match:
            severity: warning
    receivers:
      - name: slack
        slack_configs:
          - api_url: '${SLACK_WEBHOOK_URL}'
            channel: '#alerts'
            send_resolved: true
      - name: pagerduty
        pagerduty_configs:
          - routing_key: '${PAGERDUTY_KEY}'
```

### PrometheusRule

```yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: app-alerts
  labels:
    release: prometheus
spec:
  groups:
    - name: app.rules
      rules:
        - alert: HighErrorRate
          expr: |
            sum(rate(http_requests_total{status=~"5.."}[5m])) by (service)
            /
            sum(rate(http_requests_total[5m])) by (service)
            > 0.05
          for: 5m
          labels:
            severity: critical
          annotations:
            summary: "High error rate on {{ $labels.service }}"
            description: "{{ $labels.service }} has {{ $value | humanizePercentage }} 5xx error rate"

        - alert: HighLatency
          expr: |
            histogram_quantile(0.99,
              sum(rate(http_request_duration_seconds_bucket[5m])) by (le, service)
            ) > 2
          for: 10m
          labels:
            severity: warning
          annotations:
            summary: "High p99 latency on {{ $labels.service }}"
            description: "p99 latency is {{ $value }}s"

        - alert: PodCrashLooping
          expr: |
            rate(kube_pod_container_status_restarts_total[15m]) * 60 * 15 > 3
          for: 5m
          labels:
            severity: critical
          annotations:
            summary: "Pod {{ $labels.namespace }}/{{ $labels.pod }} crash looping"

        - alert: PodNotReady
          expr: |
            kube_pod_status_ready{condition="true"} == 0
          for: 15m
          labels:
            severity: warning
          annotations:
            summary: "Pod {{ $labels.namespace }}/{{ $labels.pod }} not ready for 15m"
```

---

## Custom Resource Definitions (CRDs)

### Creating a CRD

```yaml
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: databases.myapp.example.com
spec:
  group: myapp.example.com
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
                  enum: ["postgres", "mysql"]
                version:
                  type: string
                storage:
                  type: string
                  pattern: "^[0-9]+Gi$"
                replicas:
                  type: integer
                  minimum: 1
                  maximum: 5
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
      additionalPrinterColumns:
        - name: Engine
          type: string
          jsonPath: .spec.engine
        - name: Version
          type: string
          jsonPath: .spec.version
        - name: Status
          type: string
          jsonPath: .status.phase
        - name: Endpoint
          type: string
          jsonPath: .status.endpoint
        - name: Age
          type: date
          jsonPath: .metadata.creationTimestamp
  scope: Namespaced
  names:
    plural: databases
    singular: database
    kind: Database
    shortNames:
      - db

---
# Using the CRD
apiVersion: myapp.example.com/v1
kind: Database
metadata:
  name: users-db
spec:
  engine: postgres
  version: "16"
  storage: 100Gi
  replicas: 3
  backup:
    enabled: true
    schedule: "0 */6 * * *"
```

---

## Operational Patterns

### Resource Quotas

```yaml
apiVersion: v1
kind: ResourceQuota
metadata:
  name: team-quota
  namespace: team-alpha
spec:
  hard:
    requests.cpu: "20"
    requests.memory: 40Gi
    limits.cpu: "40"
    limits.memory: 80Gi
    pods: "50"
    services: "20"
    persistentvolumeclaims: "20"
    requests.storage: 500Gi
```

### LimitRange

```yaml
apiVersion: v1
kind: LimitRange
metadata:
  name: defaults
  namespace: team-alpha
spec:
  limits:
    - type: Container
      default:
        cpu: 500m
        memory: 512Mi
      defaultRequest:
        cpu: 100m
        memory: 128Mi
      max:
        cpu: "4"
        memory: 8Gi
      min:
        cpu: 50m
        memory: 64Mi
    - type: Pod
      max:
        cpu: "8"
        memory: 16Gi
    - type: PersistentVolumeClaim
      max:
        storage: 100Gi
      min:
        storage: 1Gi
```

### Priority Classes

```yaml
apiVersion: scheduling.k8s.io/v1
kind: PriorityClass
metadata:
  name: critical
value: 1000000
globalDefault: false
description: "Critical system workloads"
preemptionPolicy: PreemptLowerPriority

---
apiVersion: scheduling.k8s.io/v1
kind: PriorityClass
metadata:
  name: high
value: 100000
globalDefault: false
description: "Important production workloads"

---
apiVersion: scheduling.k8s.io/v1
kind: PriorityClass
metadata:
  name: default
value: 0
globalDefault: true
description: "Default priority for all workloads"

---
apiVersion: scheduling.k8s.io/v1
kind: PriorityClass
metadata:
  name: batch
value: -100
globalDefault: false
description: "Low priority batch jobs — can be preempted"
preemptionPolicy: Never
```

---

## Troubleshooting Quick Reference

### Common kubectl Commands

```bash
# Pod debugging
kubectl get pods -A -o wide                     # All pods, all namespaces
kubectl describe pod <name>                     # Detailed pod info
kubectl logs <pod> -c <container> --previous    # Previous container logs
kubectl logs <pod> -f --since=5m               # Stream recent logs
kubectl exec -it <pod> -- sh                    # Shell into pod
kubectl port-forward <pod> 8080:8080           # Local port forward

# Resource usage
kubectl top pods --sort-by=memory              # Memory usage
kubectl top nodes                               # Node usage

# Events
kubectl get events --sort-by='.lastTimestamp'  # Recent events
kubectl get events --field-selector type=Warning  # Warnings only

# Network debugging
kubectl run debug --image=nicolaka/netshoot --rm -it -- bash
# Inside: curl, dig, nslookup, tcpdump, iperf, etc.

# Node debugging
kubectl debug node/<node-name> -it --image=ubuntu
```

### Common Issues and Fixes

```
CrashLoopBackOff:
  → Check logs: kubectl logs <pod> --previous
  → Common causes: wrong command, missing env var, failed health check

ImagePullBackOff:
  → Check image name/tag: kubectl describe pod <pod>
  → Check registry auth: kubectl get secret <pull-secret> -o yaml
  → Check network: can nodes reach the registry?

Pending:
  → Check events: kubectl describe pod <pod>
  → Insufficient resources: check node capacity vs requests
  → No matching nodes: check nodeSelector, tolerations, affinity
  → PVC not bound: check StorageClass and provisioner

OOMKilled:
  → Increase memory limits
  → Check for memory leaks
  → Tune JVM heap size / Node.js max-old-space-size

Evicted:
  → Node under disk pressure: check node conditions
  → Pods exceeding ephemeral storage limits
  → Fix: add resource limits, clean up logs/tmp files
```
