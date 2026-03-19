# CI/CD for Containers Reference

Quick-reference guide for container CI/CD pipelines. Agents consult this automatically — you can also read it directly for quick answers.

---

## GitHub Actions

### Build, Scan, Push

```yaml
name: Build and Push
on:
  push:
    branches: [main]
    tags: ["v*"]
  pull_request:
    branches: [main]

env:
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}

permissions:
  contents: read
  packages: write
  security-events: write
  id-token: write       # For cosign keyless signing

jobs:
  build:
    runs-on: ubuntu-latest
    outputs:
      digest: ${{ steps.build.outputs.digest }}
      tags: ${{ steps.meta.outputs.tags }}
    steps:
      - uses: actions/checkout@v4

      - uses: docker/setup-buildx-action@v3

      - uses: docker/login-action@v3
        if: github.event_name != 'pull_request'
        with:
          registry: ${{ env.REGISTRY }}
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - uses: docker/metadata-action@v5
        id: meta
        with:
          images: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}
          tags: |
            type=ref,event=branch
            type=ref,event=pr
            type=semver,pattern={{version}}
            type=semver,pattern={{major}}.{{minor}}
            type=sha,prefix=

      - uses: docker/build-push-action@v6
        id: build
        with:
          context: .
          push: ${{ github.event_name != 'pull_request' }}
          tags: ${{ steps.meta.outputs.tags }}
          labels: ${{ steps.meta.outputs.labels }}
          cache-from: type=gha
          cache-to: type=gha,mode=max
          platforms: linux/amd64,linux/arm64
          provenance: true
          sbom: true

  scan:
    needs: build
    runs-on: ubuntu-latest
    if: github.event_name != 'pull_request'
    steps:
      - uses: aquasecurity/trivy-action@master
        with:
          image-ref: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}@${{ needs.build.outputs.digest }}
          format: sarif
          output: trivy-results.sarif
          severity: CRITICAL,HIGH

      - uses: github/codeql-action/upload-sarif@v3
        with:
          sarif_file: trivy-results.sarif

  sign:
    needs: build
    runs-on: ubuntu-latest
    if: github.event_name != 'pull_request'
    steps:
      - uses: sigstore/cosign-installer@v3

      - run: |
          cosign sign --yes \
            ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}@${{ needs.build.outputs.digest }}
```

### Matrix Build — Multiple Services

```yaml
jobs:
  build:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        service: [api, worker, frontend]
        include:
          - service: api
            dockerfile: services/api/Dockerfile
            context: .
          - service: worker
            dockerfile: services/worker/Dockerfile
            context: .
          - service: frontend
            dockerfile: services/frontend/Dockerfile
            context: services/frontend
    steps:
      - uses: actions/checkout@v4
      - uses: docker/setup-buildx-action@v3
      - uses: docker/build-push-action@v6
        with:
          context: ${{ matrix.context }}
          file: ${{ matrix.dockerfile }}
          push: true
          tags: ghcr.io/${{ github.repository }}/${{ matrix.service }}:${{ github.sha }}
          cache-from: type=gha,scope=${{ matrix.service }}
          cache-to: type=gha,scope=${{ matrix.service }},mode=max
```

### Helm Deploy Action

```yaml
  deploy:
    needs: [build, scan]
    runs-on: ubuntu-latest
    environment: production
    steps:
      - uses: actions/checkout@v4

      - uses: azure/setup-kubectl@v4
      - uses: azure/setup-helm@v4

      - uses: aws-actions/configure-aws-credentials@v4
        with:
          role-to-assume: ${{ secrets.AWS_ROLE_ARN }}
          aws-region: us-east-1

      - run: aws eks update-kubeconfig --name production --region us-east-1

      - name: Deploy with Helm
        run: |
          helm upgrade --install myapp ./charts/myapp \
            --namespace production \
            --set image.tag=${{ github.sha }} \
            -f charts/myapp/values.production.yaml \
            --atomic \
            --timeout 10m \
            --wait
```

---

## GitLab CI

### Complete Pipeline

```yaml
stages:
  - build
  - test
  - scan
  - deploy

variables:
  IMAGE: $CI_REGISTRY_IMAGE:$CI_COMMIT_SHORT_SHA
  IMAGE_LATEST: $CI_REGISTRY_IMAGE:latest

# Build and push
build:
  stage: build
  image: docker:27
  services:
    - docker:27-dind
  variables:
    DOCKER_BUILDKIT: "1"
  before_script:
    - docker login -u $CI_REGISTRY_USER -p $CI_REGISTRY_PASSWORD $CI_REGISTRY
  script:
    - docker pull $IMAGE_LATEST || true
    - docker build
      --cache-from $IMAGE_LATEST
      --build-arg BUILDKIT_INLINE_CACHE=1
      -t $IMAGE
      -t $IMAGE_LATEST
      .
    - docker push $IMAGE
    - docker push $IMAGE_LATEST

# Run tests in container
test:
  stage: test
  image: $IMAGE
  services:
    - postgres:17-alpine
    - redis:7-alpine
  variables:
    DATABASE_URL: "postgres://postgres:test@postgres:5432/testdb"
    REDIS_URL: "redis://redis:6379"
    POSTGRES_DB: testdb
    POSTGRES_PASSWORD: test
  script:
    - npm test

# Vulnerability scan
scan:
  stage: scan
  image:
    name: aquasec/trivy:latest
    entrypoint: [""]
  script:
    - trivy image --exit-code 0 --severity HIGH $IMAGE
    - trivy image --exit-code 1 --severity CRITICAL $IMAGE
  allow_failure: false

# Deploy to staging (auto)
deploy-staging:
  stage: deploy
  image: alpine/helm:3.16
  environment:
    name: staging
    url: https://staging.example.com
  before_script:
    - aws eks update-kubeconfig --name staging
  script:
    - helm upgrade --install myapp ./charts/myapp
      --namespace staging
      --set image.tag=$CI_COMMIT_SHORT_SHA
      -f charts/myapp/values.staging.yaml
      --atomic --wait
  only:
    - main

# Deploy to production (manual)
deploy-production:
  stage: deploy
  image: alpine/helm:3.16
  environment:
    name: production
    url: https://example.com
  before_script:
    - aws eks update-kubeconfig --name production
  script:
    - helm upgrade --install myapp ./charts/myapp
      --namespace production
      --set image.tag=$CI_COMMIT_SHORT_SHA
      -f charts/myapp/values.production.yaml
      --atomic --wait --timeout 10m
  only:
    - tags
  when: manual
```

---

## ArgoCD — GitOps

### Application Definition

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: myapp
  namespace: argocd
  finalizers:
    - resources-finalizer.argocd.argoproj.io
spec:
  project: default
  source:
    repoURL: https://github.com/org/k8s-manifests.git
    targetRevision: HEAD
    path: apps/myapp/overlays/production
  destination:
    server: https://kubernetes.default.svc
    namespace: production
  syncPolicy:
    automated:
      prune: true              # Delete resources removed from Git
      selfHeal: true           # Revert manual changes
      allowEmpty: false
    syncOptions:
      - CreateNamespace=true
      - PruneLast=true
      - ApplyOutOfSyncOnly=true
    retry:
      limit: 5
      backoff:
        duration: 5s
        factor: 2
        maxDuration: 3m
```

### Helm-Based Application

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: myapp
  namespace: argocd
spec:
  source:
    repoURL: https://github.com/org/helm-charts.git
    targetRevision: HEAD
    path: charts/myapp
    helm:
      releaseName: myapp
      valueFiles:
        - values.yaml
        - values.production.yaml
      parameters:
        - name: image.tag
          value: "1.2.3"
  destination:
    server: https://kubernetes.default.svc
    namespace: production
```

### ApplicationSet — Multi-Environment

```yaml
apiVersion: argoproj.io/v1alpha1
kind: ApplicationSet
metadata:
  name: myapp
  namespace: argocd
spec:
  generators:
    - list:
        elements:
          - cluster: staging
            url: https://staging-cluster.example.com
            values: staging
          - cluster: production
            url: https://production-cluster.example.com
            values: production
  template:
    metadata:
      name: "myapp-{{cluster}}"
    spec:
      project: default
      source:
        repoURL: https://github.com/org/k8s-manifests.git
        targetRevision: HEAD
        path: "apps/myapp/overlays/{{values}}"
      destination:
        server: "{{url}}"
        namespace: "myapp"
      syncPolicy:
        automated:
          prune: true
          selfHeal: true
```

### Image Updater

```yaml
# ArgoCD Image Updater — auto-update image tags
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: myapp
  annotations:
    argocd-image-updater.argoproj.io/image-list: >-
      myapp=ghcr.io/org/myapp
    argocd-image-updater.argoproj.io/myapp.update-strategy: semver
    argocd-image-updater.argoproj.io/myapp.allow-tags: "regexp:^[0-9]+\\.[0-9]+\\.[0-9]+$"
    argocd-image-updater.argoproj.io/write-back-method: git
```

---

## Tekton Pipelines

### Build and Deploy Pipeline

```yaml
apiVersion: tekton.dev/v1
kind: Pipeline
metadata:
  name: build-deploy
spec:
  params:
    - name: repo-url
      type: string
    - name: image
      type: string
    - name: tag
      type: string
  workspaces:
    - name: source
    - name: docker-credentials

  tasks:
    - name: fetch-source
      taskRef:
        name: git-clone
      workspaces:
        - name: output
          workspace: source
      params:
        - name: url
          value: $(params.repo-url)

    - name: build-image
      runAfter: [fetch-source]
      taskRef:
        name: kaniko
      workspaces:
        - name: source
          workspace: source
        - name: dockerconfig
          workspace: docker-credentials
      params:
        - name: IMAGE
          value: "$(params.image):$(params.tag)"

    - name: scan-image
      runAfter: [build-image]
      taskRef:
        name: trivy-scanner
      params:
        - name: IMAGE
          value: "$(params.image):$(params.tag)"
        - name: SEVERITY
          value: "CRITICAL,HIGH"

    - name: deploy
      runAfter: [scan-image]
      taskRef:
        name: helm-upgrade
      params:
        - name: release-name
          value: myapp
        - name: chart
          value: ./charts/myapp
        - name: set
          value: "image.tag=$(params.tag)"
---
# Trigger on push
apiVersion: triggers.tekton.dev/v1beta1
kind: EventListener
metadata:
  name: github-push
spec:
  triggers:
    - name: push-trigger
      bindings:
        - ref: github-push-binding
      template:
        ref: build-deploy-template
```

---

## Kustomize for CD

### Directory Structure

```
k8s/
├── base/
│   ├── kustomization.yaml
│   ├── deployment.yaml
│   ├── service.yaml
│   └── ingress.yaml
└── overlays/
    ├── staging/
    │   ├── kustomization.yaml
    │   ├── replica-count.yaml
    │   └── ingress-patch.yaml
    └── production/
        ├── kustomization.yaml
        ├── replica-count.yaml
        ├── hpa.yaml
        └── ingress-patch.yaml
```

### Base Kustomization

```yaml
# k8s/base/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
  - deployment.yaml
  - service.yaml
  - ingress.yaml

commonLabels:
  app.kubernetes.io/name: myapp

images:
  - name: myapp
    newName: ghcr.io/org/myapp
    newTag: latest
```

### Production Overlay

```yaml
# k8s/overlays/production/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namespace: production

resources:
  - ../../base
  - hpa.yaml

patches:
  - path: replica-count.yaml
  - path: ingress-patch.yaml

images:
  - name: myapp
    newName: ghcr.io/org/myapp
    newTag: "1.2.3"        # Override in CI

configMapGenerator:
  - name: app-config
    literals:
      - NODE_ENV=production
      - LOG_LEVEL=info

secretGenerator:
  - name: app-secrets
    envs:
      - secrets.env        # NOT committed to Git
```

```yaml
# replica-count.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
spec:
  replicas: 5
```

**Deploy with Kustomize:**
```bash
# Preview
kubectl kustomize k8s/overlays/production

# Apply
kubectl apply -k k8s/overlays/production

# In CI — update image tag
cd k8s/overlays/production
kustomize edit set image myapp=ghcr.io/org/myapp:${COMMIT_SHA}
kubectl apply -k .
```

---

## Container Registry Patterns

### Registry Authentication

```bash
# GitHub Container Registry (GHCR)
echo $GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin

# AWS ECR
aws ecr get-login-password --region us-east-1 | \
  docker login --username AWS --password-stdin 123456789.dkr.ecr.us-east-1.amazonaws.com

# Google Artifact Registry
gcloud auth configure-docker us-central1-docker.pkg.dev

# Azure Container Registry
az acr login --name myregistry

# Docker Hub
echo $DOCKER_TOKEN | docker login -u USERNAME --password-stdin
```

### Kubernetes Image Pull Secrets

```bash
# Create pull secret
kubectl create secret docker-registry ghcr-secret \
  --docker-server=ghcr.io \
  --docker-username=USERNAME \
  --docker-password=$GITHUB_TOKEN \
  -n production

# Use in pod
spec:
  imagePullSecrets:
    - name: ghcr-secret

# Or attach to ServiceAccount (applies to all pods)
kubectl patch serviceaccount default \
  -n production \
  -p '{"imagePullSecrets": [{"name": "ghcr-secret"}]}'
```

---

## Deployment Strategies in CI/CD

### Progressive Delivery Flow

```
1. PR merged to main
2. CI builds image, scans, signs
3. Image pushed to registry with SHA tag
4. CD tool detects new image (ArgoCD Image Updater, Flux)
5. Canary deployment: 5% traffic → monitor metrics
6. Gradual increase: 20% → 50% → 80% → 100%
7. Automatic rollback if error rate exceeds threshold
8. Full promotion after all steps pass
```

### Rollback Strategy

```bash
# Kubernetes native rollback
kubectl rollout undo deployment/myapp -n production

# Helm rollback
helm rollback myapp 0 -n production           # Previous version
helm rollback myapp 3 -n production           # Specific revision

# ArgoCD rollback
argocd app rollback myapp

# GitOps rollback — revert the Git commit
git revert HEAD && git push
```

---

## Secrets in CI/CD

### GitHub Actions Secrets

```yaml
# Use repository or environment secrets
env:
  DATABASE_URL: ${{ secrets.DATABASE_URL }}

# Create K8s secret from CI secret
- run: |
    kubectl create secret generic api-secrets \
      --from-literal=database-url="${{ secrets.DATABASE_URL }}" \
      --from-literal=jwt-secret="${{ secrets.JWT_SECRET }}" \
      --dry-run=client -o yaml | kubectl apply -f -
```

### Build-Time Secrets

```yaml
# Mount secrets during build — never in image layers
- uses: docker/build-push-action@v6
  with:
    context: .
    push: true
    tags: ${{ steps.meta.outputs.tags }}
    secrets: |
      npm_token=${{ secrets.NPM_TOKEN }}
      github_token=${{ secrets.GITHUB_TOKEN }}
```

```dockerfile
# In Dockerfile
RUN --mount=type=secret,id=npm_token \
    NPM_TOKEN=$(cat /run/secrets/npm_token) npm ci
```

---

## Monitoring Deployments

### Deployment Notifications

```yaml
# GitHub Actions — Slack notification
- name: Notify Slack
  if: always()
  uses: slackapi/slack-github-action@v2
  with:
    webhook: ${{ secrets.SLACK_WEBHOOK }}
    webhook-type: incoming-webhook
    payload: |
      {
        "text": "${{ job.status == 'success' && '✅' || '❌' }} Deploy ${{ github.sha }} to production: ${{ job.status }}"
      }
```

### Smoke Tests After Deploy

```yaml
deploy:
  steps:
    - name: Deploy
      run: helm upgrade --install myapp ./charts/myapp --atomic --wait

    - name: Smoke test
      run: |
        # Wait for rollout
        kubectl rollout status deployment/myapp -n production --timeout=5m

        # Test health endpoint
        POD=$(kubectl get pod -n production -l app=myapp -o jsonpath='{.items[0].metadata.name}')
        kubectl exec -n production $POD -- wget -qO- http://localhost:3000/health

        # Test via ingress
        for i in $(seq 1 5); do
          STATUS=$(curl -s -o /dev/null -w "%{http_code}" https://api.example.com/health)
          if [ "$STATUS" = "200" ]; then
            echo "Health check passed"
            exit 0
          fi
          sleep 5
        done
        echo "Health check failed"
        exit 1
```
