---
name: k8s-security
description: >
  Kubernetes security hardening and access control patterns.
  Use when configuring RBAC, network policies, pod security standards,
  secrets management, or security auditing in Kubernetes clusters.
  Triggers: "k8s security", "RBAC", "network policy", "pod security",
  "kubernetes secrets", "service account", "security context".
  NOT for: application-level security (XSS, SQL injection), cloud IAM without K8s.
version: 1.0.0
allowed-tools: Read, Grep, Glob, Bash
---

# Kubernetes Security

## RBAC (Role-Based Access Control)

```yaml
# ServiceAccount — identity for pods
apiVersion: v1
kind: ServiceAccount
metadata:
  name: api-server
  namespace: production
  annotations:
    # AWS IRSA
    eks.amazonaws.com/role-arn: arn:aws:iam::123456789:role/api-server
    # GCP Workload Identity
    iam.gke.io/gcp-service-account: api@project.iam.gserviceaccount.com
automountServiceAccountToken: false  # Don't mount unless needed
---
# Role — namespace-scoped permissions
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: api-server-role
  namespace: production
rules:
  - apiGroups: [""]
    resources: ["configmaps"]
    verbs: ["get", "list", "watch"]
  - apiGroups: [""]
    resources: ["secrets"]
    verbs: ["get"]
    resourceNames: ["api-secrets"]  # Restrict to specific secret
  - apiGroups: [""]
    resources: ["pods"]
    verbs: ["get", "list"]
---
# RoleBinding
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: api-server-binding
  namespace: production
subjects:
  - kind: ServiceAccount
    name: api-server
    namespace: production
roleRef:
  kind: Role
  name: api-server-role
  apiGroup: rbac.authorization.k8s.io
---
# ClusterRole — cluster-wide permissions (use sparingly)
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: metrics-reader
rules:
  - apiGroups: ["metrics.k8s.io"]
    resources: ["pods", "nodes"]
    verbs: ["get", "list"]
  - nonResourceURLs: ["/healthz", "/readyz"]
    verbs: ["get"]
```

## RBAC Best Practices

```yaml
# Developer Role — read-only access
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: developer
rules:
  - apiGroups: ["", "apps", "batch"]
    resources: ["pods", "deployments", "services", "jobs", "configmaps"]
    verbs: ["get", "list", "watch"]
  - apiGroups: [""]
    resources: ["pods/log"]
    verbs: ["get"]
  - apiGroups: [""]
    resources: ["pods/exec"]
    verbs: ["create"]
  # NEVER give developers access to secrets in production
---
# CI/CD Role — deploy but not delete
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: deployer
  namespace: production
rules:
  - apiGroups: ["apps"]
    resources: ["deployments"]
    verbs: ["get", "list", "patch", "update"]
  - apiGroups: [""]
    resources: ["services"]
    verbs: ["get", "list"]
```

```bash
# RBAC debugging
kubectl auth can-i create pods -n production --as system:serviceaccount:production:api-server
kubectl auth can-i --list --as system:serviceaccount:production:api-server -n production
kubectl get rolebindings,clusterrolebindings -A -o wide | grep api-server
```

## Network Policies

```yaml
# Default deny all — start with zero trust
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: default-deny-all
  namespace: production
spec:
  podSelector: {}
  policyTypes:
    - Ingress
    - Egress
---
# Allow API server to receive traffic from ingress and internal services
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: api-server-ingress
  namespace: production
spec:
  podSelector:
    matchLabels:
      app: api-server
  policyTypes:
    - Ingress
  ingress:
    - from:
        - namespaceSelector:
            matchLabels:
              kubernetes.io/metadata.name: ingress-nginx
        - podSelector:
            matchLabels:
              app: frontend
      ports:
        - protocol: TCP
          port: 8080
---
# Allow API server to reach database and DNS
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: api-server-egress
  namespace: production
spec:
  podSelector:
    matchLabels:
      app: api-server
  policyTypes:
    - Egress
  egress:
    # Database
    - to:
        - podSelector:
            matchLabels:
              app: postgres
      ports:
        - protocol: TCP
          port: 5432
    # DNS — always allow
    - to: []
      ports:
        - protocol: UDP
          port: 53
        - protocol: TCP
          port: 53
    # External HTTPS
    - to:
        - ipBlock:
            cidr: 0.0.0.0/0
            except:
              - 10.0.0.0/8
              - 172.16.0.0/12
              - 192.168.0.0/16
      ports:
        - protocol: TCP
          port: 443
```

## Pod Security

```yaml
# Pod Security Standards (PSS) — namespace-level enforcement
apiVersion: v1
kind: Namespace
metadata:
  name: production
  labels:
    pod-security.kubernetes.io/enforce: restricted
    pod-security.kubernetes.io/audit: restricted
    pod-security.kubernetes.io/warn: restricted
---
# Restricted pod — passes PSS "restricted" level
apiVersion: v1
kind: Pod
metadata:
  name: secure-pod
  namespace: production
spec:
  securityContext:
    runAsNonRoot: true
    runAsUser: 1000
    runAsGroup: 1000
    fsGroup: 1000
    seccompProfile:
      type: RuntimeDefault
  containers:
    - name: app
      image: registry.example.com/app:v1.0.0
      securityContext:
        readOnlyRootFilesystem: true
        allowPrivilegeEscalation: false
        capabilities:
          drop: ["ALL"]
      volumeMounts:
        - name: tmp
          mountPath: /tmp
        - name: cache
          mountPath: /var/cache
  volumes:
    - name: tmp
      emptyDir:
        sizeLimit: 100Mi
    - name: cache
      emptyDir:
        sizeLimit: 500Mi
```

## Secrets Management

```yaml
# External Secrets Operator — pull from external vaults
apiVersion: external-secrets.io/v1beta1
kind: ClusterSecretStore
metadata:
  name: aws-secrets-manager
spec:
  provider:
    aws:
      service: SecretsManager
      region: us-east-1
      auth:
        jwt:
          serviceAccountRef:
            name: external-secrets
            namespace: external-secrets
---
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
    template:
      engineVersion: v2
      data:
        DATABASE_URL: "postgresql://{{ .user }}:{{ .password }}@{{ .host }}:5432/{{ .dbname }}"
  data:
    - secretKey: user
      remoteRef:
        key: production/db
        property: username
    - secretKey: password
      remoteRef:
        key: production/db
        property: password
    - secretKey: host
      remoteRef:
        key: production/db
        property: host
    - secretKey: dbname
      remoteRef:
        key: production/db
        property: dbname
```

## Security Scanning

```bash
# Scan images for vulnerabilities
trivy image registry.example.com/api-server:v1.2.3
trivy image --severity HIGH,CRITICAL registry.example.com/api-server:v1.2.3

# Scan running cluster
trivy k8s --report summary cluster
trivy k8s --namespace production --report all

# Scan manifests before applying
trivy config ./k8s/
kubeaudit all -f deployment.yaml

# Check RBAC permissions
kubectl auth can-i --list -n production

# CIS Kubernetes Benchmark
kube-bench run
polaris audit --audit-path ./k8s/
```

## Admission Control

```yaml
# Kyverno policy — enforce image registry
apiVersion: kyverno.io/v1
kind: ClusterPolicy
metadata:
  name: restrict-image-registries
spec:
  validationFailureAction: Enforce
  rules:
    - name: validate-registry
      match:
        any:
          - resources:
              kinds: ["Pod"]
      validate:
        message: "Images must be from approved registries"
        pattern:
          spec:
            containers:
              - image: "registry.example.com/*"
---
# Kyverno policy — require resource limits
apiVersion: kyverno.io/v1
kind: ClusterPolicy
metadata:
  name: require-resource-limits
spec:
  validationFailureAction: Enforce
  rules:
    - name: check-limits
      match:
        any:
          - resources:
              kinds: ["Pod"]
      validate:
        message: "CPU and memory limits are required"
        pattern:
          spec:
            containers:
              - resources:
                  limits:
                    memory: "?*"
                    cpu: "?*"
```

## Runtime Security

```yaml
# Falco rules — runtime threat detection
- rule: Terminal shell in container
  desc: Detect shell spawned in a container
  condition: >
    spawned_process and container and
    proc.name in (bash, sh, zsh) and
    not proc.pname in (cron, supervisord)
  output: >
    Shell spawned in container
    (user=%user.name container=%container.name
     shell=%proc.name parent=%proc.pname)
  priority: WARNING
  tags: [container, shell]

- rule: Sensitive file access
  desc: Detect reads of sensitive files
  condition: >
    open_read and container and
    fd.name in (/etc/shadow, /etc/passwd, /proc/self/environ)
  output: >
    Sensitive file accessed
    (file=%fd.name container=%container.name)
  priority: CRITICAL
  tags: [container, filesystem]
```

## Security Audit Checklist

```
Cluster Level:
  [ ] RBAC enabled
  [ ] API server audit logging enabled
  [ ] etcd encryption at rest enabled
  [ ] Node OS hardened (CIS benchmarks)
  [ ] Network plugin supports NetworkPolicy (Calico/Cilium)
  [ ] Admission controllers active (PodSecurity, OPA/Kyverno)

Namespace Level:
  [ ] Pod Security Standards enforced (restricted level)
  [ ] Default deny NetworkPolicy applied
  [ ] ResourceQuota set
  [ ] LimitRange set
  [ ] ServiceAccount tokens not auto-mounted

Workload Level:
  [ ] Images from trusted registry only
  [ ] No :latest tags
  [ ] runAsNonRoot: true
  [ ] readOnlyRootFilesystem: true
  [ ] allowPrivilegeEscalation: false
  [ ] capabilities: drop ALL
  [ ] Secrets via External Secrets
  [ ] Resource limits set
  [ ] Health probes configured

CI/CD Level:
  [ ] Image scanning in pipeline
  [ ] Manifest scanning
  [ ] Signed images (cosign/Notary)
  [ ] GitOps deployment (no kubectl from laptops)
  [ ] Least-privilege CI/CD service account
```

## Gotchas

1. **Default ServiceAccount has too many permissions** — every pod gets the default SA unless you specify one. Create dedicated ServiceAccounts per workload. Set `automountServiceAccountToken: false` unless the pod needs API access.

2. **NetworkPolicy requires a CNI that supports it** — the default kubenet doesn't enforce NetworkPolicies. You need Calico, Cilium, or Weave. Check your CNI before relying on NetworkPolicies.

3. **Secrets are base64-encoded, not encrypted** — anyone with `kubectl get secret` access can decode them. Use External Secrets Operator or SealedSecrets. Enable etcd encryption at rest. Never commit plain Secrets to git.

4. **PSS "restricted" breaks many Helm charts** — most community Helm charts run as root by default. Test in "audit" mode first, fix security contexts, then switch to "enforce."

5. **RBAC wildcards are dangerous** — `resources: ["*"]` or `verbs: ["*"]` grants access to everything, including future resource types. Always enumerate specific resources and verbs.

6. **Image pull from public registry in production** — if Docker Hub goes down, your deploys fail. Mirror images to a private registry. Use `imagePullPolicy: IfNotPresent` with versioned tags.
