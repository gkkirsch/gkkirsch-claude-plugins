---
name: docker-kubernetes-suite
description: >
  Docker & Kubernetes Mastery Suite — complete development toolkit for containerizing, deploying, and securing
  production-grade applications. Dockerfile optimization with multi-stage builds and BuildKit, Kubernetes
  architecture with deployments, autoscaling, and networking, Helm chart engineering with templating and testing,
  and container security with image scanning, Pod Security Standards, and supply chain protection.
  Triggers: "dockerfile", "docker build", "docker compose", "multi-stage build", "buildkit",
  "docker image", "container", "containerize", ".dockerignore", "docker optimize",
  "kubernetes", "k8s", "kubectl", "deployment", "pod", "service", "ingress",
  "hpa", "horizontal pod autoscaler", "keda", "scaling kubernetes",
  "rbac", "network policy", "namespace", "statefulset", "daemonset",
  "helm", "helm chart", "helm template", "helm values", "helmfile",
  "helm install", "helm upgrade", "chart development", "helm test",
  "container security", "image scanning", "trivy", "grype", "docker scout",
  "pod security", "seccomp", "kyverno", "opa gatekeeper",
  "sealed secrets", "external secrets", "vault kubernetes",
  "cosign", "sigstore", "image signing", "sbom", "supply chain",
  "argocd", "gitops", "tekton", "kustomize", "ci cd containers".
  Dispatches the appropriate specialist agent: docker-expert, kubernetes-architect,
  helm-engineer, or container-security-expert.
  NOT for: Application code without containers, cloud provider console setup,
  bare-metal server administration, or non-container orchestration.
version: 1.0.0
argument-hint: "<docker|k8s|helm|security> [target]"
user-invocable: true
allowed-tools: Read, Grep, Glob, Bash
model: sonnet
---

# Docker & Kubernetes Mastery Suite

Production-grade Docker and Kubernetes development agents for Claude Code. Four specialist agents that handle containerization, cluster architecture, Helm chart engineering, and container security — the complete container lifecycle.

## Available Agents

### Docker Expert (`docker-expert`)
Builds and optimizes Docker images for any language or framework. Multi-stage builds, BuildKit cache mounts and secret mounts, Docker Compose orchestration, .dockerignore optimization, layer caching strategies, non-root users, distroless images, health checks, and development workflows with watch mode.

**Invoke**: Dispatch via Task tool with `subagent_type: "docker-expert"`.

**Example prompts**:
- "Containerize my Node.js app for production"
- "Optimize my Dockerfile — image is 1.2GB"
- "Set up Docker Compose with PostgreSQL and Redis"
- "Add multi-platform build support for ARM"

### Kubernetes Architect (`kubernetes-architect`)
Designs and deploys production-grade Kubernetes workloads. Rolling updates, blue-green, and canary deployments with Argo Rollouts. HPA with custom metrics, KEDA event-driven scaling, VPA. Services, Gateway API, Ingress, network policies. RBAC, ServiceAccounts, Pod Security Standards. Resource management, PDB, topology spread, priority classes.

**Invoke**: Dispatch via Task tool with `subagent_type: "kubernetes-architect"`.

**Example prompts**:
- "Deploy my app to Kubernetes with autoscaling"
- "Set up canary deployments with automatic rollback"
- "Configure RBAC for multi-team namespace isolation"
- "Add network policies for zero-trust networking"

### Helm Engineer (`helm-engineer`)
Creates and manages Helm charts with production-grade templating. Chart development with proper _helpers.tpl, values.yaml design with JSON Schema validation, environment overrides, hooks for migrations and initialization, dependency management with subcharts and library charts, OCI registry publishing, chart testing with ct/kubeconform/polaris, and release management with Helmfile.

**Invoke**: Dispatch via Task tool with `subagent_type: "helm-engineer"`.

**Example prompts**:
- "Create a Helm chart for my microservice"
- "Design values.yaml for staging and production"
- "Add database migration hooks to my chart"
- "Set up chart testing in CI with ct and kubeconform"

### Container Security Expert (`container-security-expert`)
Secures containerized applications across build, deploy, and runtime. Vulnerability scanning with Trivy/Grype/Scout, SBOM generation, distroless hardening. Pod Security Standards enforcement, Kyverno/OPA policies. Zero-trust network policies, service mesh mTLS. External Secrets Operator, Sealed Secrets, Vault integration, workload identity. Image signing with cosign/Sigstore, SLSA provenance. Falco runtime monitoring, seccomp profiles, CIS benchmarks.

**Invoke**: Dispatch via Task tool with `subagent_type: "container-security-expert"`.

**Example prompts**:
- "Scan my images and fix critical vulnerabilities"
- "Set up Pod Security Standards for all namespaces"
- "Implement secrets management with External Secrets Operator"
- "Add image signing and verification to our pipeline"

## Quick Start: /docker-k8s

Use the `/docker-k8s` command for guided Docker and Kubernetes development:

```
/docker-k8s                          # Auto-detect and suggest improvements
/docker-k8s docker                   # Dockerfile optimization
/docker-k8s docker --compose         # Docker Compose setup
/docker-k8s docker --optimize        # Image size and build time
/docker-k8s k8s                      # Kubernetes deployment
/docker-k8s k8s --scale              # Autoscaling setup
/docker-k8s k8s --network            # Services, ingress, network policies
/docker-k8s helm                     # Helm chart development
/docker-k8s helm --values            # Multi-environment values
/docker-k8s helm --test              # Chart testing and CI
/docker-k8s security                 # Security audit
/docker-k8s security --scan          # Vulnerability scanning
/docker-k8s security --policies      # Pod Security Standards
/docker-k8s --review                 # Full infrastructure review
```

The `/docker-k8s` command auto-detects your configuration, discovers your setup, and routes to the right agent.

## Agent Selection Guide

| Need | Agent | Trigger |
|------|-------|---------|
| Dockerfile creation | docker-expert | "Containerize my app" |
| Multi-stage builds | docker-expert | "Optimize my Docker image" |
| Docker Compose | docker-expert | "Set up dev environment" |
| BuildKit features | docker-expert | "Use cache mounts" |
| K8s deployment | kubernetes-architect | "Deploy to Kubernetes" |
| Autoscaling | kubernetes-architect | "Set up HPA" |
| Ingress/networking | kubernetes-architect | "Configure ingress" |
| RBAC/security | kubernetes-architect | "Set up RBAC" |
| Helm chart creation | helm-engineer | "Create Helm chart" |
| Values engineering | helm-engineer | "Multi-env values" |
| Chart testing | helm-engineer | "Test Helm chart" |
| Release management | helm-engineer | "Helm upgrade strategy" |
| Image scanning | container-security-expert | "Scan for vulnerabilities" |
| Pod security | container-security-expert | "Harden pod security" |
| Secrets management | container-security-expert | "Set up External Secrets" |
| Supply chain | container-security-expert | "Sign images with cosign" |
| Network policies | container-security-expert | "Zero-trust networking" |

## Reference Materials

This skill includes comprehensive reference documents in `references/`:

- **kubernetes-patterns.md** — Sidecar, init containers, ambassador, adapter, jobs/CronJobs, CRDs, operators, service mesh, lifecycle, multi-tenancy
- **docker-best-practices.md** — Image optimization, .dockerignore, health checks, logging, multi-platform builds, security, compose patterns, entrypoints
- **ci-cd-containers.md** — GitHub Actions, GitLab CI, ArgoCD GitOps, Tekton pipelines, Kustomize, registry patterns, deployment strategies, rollback

Agents automatically consult these references when working. You can also read them directly for quick answers.

## How It Works

1. You describe what you need (e.g., "deploy my Node.js app to Kubernetes with autoscaling")
2. The SKILL.md routes to the appropriate agent
3. The agent reads your code, discovers your stack and existing configuration
4. Solutions are designed and implemented following production best practices
5. The agent provides results and next steps

All generated artifacts follow industry best practices:
- **Security**: Non-root, read-only FS, least privilege, secrets management, image signing
- **Performance**: Multi-stage builds, layer caching, resource limits, autoscaling
- **Reliability**: Health checks, graceful shutdown, PDB, rolling updates, circuit breakers
- **Observability**: Prometheus metrics, structured logging, distributed tracing
- **GitOps**: Declarative configs, Helm charts, ArgoCD, sealed secrets, signed images
