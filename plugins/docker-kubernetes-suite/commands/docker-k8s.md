---
name: docker-k8s
description: >
  Quick Docker & Kubernetes development command — analyzes your project and helps containerize, deploy, secure,
  or manage your applications. Routes to the appropriate specialist agent based on your request.
  Triggers: "/docker-k8s", "dockerfile", "docker compose", "kubernetes deploy", "helm chart",
  "container security", "k8s manifest", "docker build", "kubernetes scaling", "ingress setup".
user-invocable: true
argument-hint: "<docker|k8s|helm|security> [target] [--review] [--audit]"
allowed-tools: Read, Write, Edit, Bash, Glob, Grep
model: sonnet
---

# /docker-k8s Command

One-command Docker and Kubernetes development. Analyzes your project, identifies the stack and existing container configuration, and routes to the appropriate specialist agent for Docker optimization, Kubernetes architecture, Helm chart engineering, or container security.

## Usage

```
/docker-k8s                             # Auto-detect and suggest improvements
/docker-k8s docker                      # Dockerfile optimization
/docker-k8s docker --compose            # Docker Compose setup
/docker-k8s docker --optimize           # Image size and build time optimization
/docker-k8s k8s                         # Kubernetes deployment manifests
/docker-k8s k8s --scale                 # Autoscaling and resource management
/docker-k8s k8s --network               # Services, ingress, network policies
/docker-k8s helm                        # Helm chart development
/docker-k8s helm --values               # Values engineering for multi-env
/docker-k8s helm --test                 # Chart testing and validation
/docker-k8s security                    # Container security audit
/docker-k8s security --scan             # Image vulnerability scanning
/docker-k8s security --policies         # Pod Security Standards and admission
/docker-k8s --review                    # Full infrastructure review
```

## Subcommands

### `docker` — Docker Image & Compose

Builds and optimizes Docker images with multi-stage builds, BuildKit, and Docker Compose orchestration.

**What it does:**
1. Scans for existing Dockerfiles, .dockerignore, and compose files
2. Identifies the application language, framework, and package manager
3. Analyzes image layers for optimization opportunities
4. Suggests or generates improvements

**Flags:**
- `--compose` — Docker Compose service orchestration and dev environment
- `--optimize` — Image size reduction and build time optimization
- `--multi-stage` — Multi-stage build patterns for the detected language
- `--review` — Review existing Dockerfiles for best practices

**Routes to:** `docker-expert` agent

### `k8s` — Kubernetes Architecture

Designs and deploys Kubernetes workloads with production-grade patterns.

**What it does:**
1. Analyzes existing Kubernetes manifests or Helm charts
2. Reviews deployment strategies, scaling, and resource management
3. Checks networking, ingress, and service configuration
4. Suggests or generates improvements

**Flags:**
- `--scale` — HPA, VPA, KEDA autoscaling setup
- `--network` — Services, ingress, Gateway API, network policies
- `--rbac` — RBAC, service accounts, and namespace isolation
- `--review` — Review existing manifests for best practices

**Routes to:** `kubernetes-architect` agent

### `helm` — Helm Chart Engineering

Creates and manages Helm charts with proper templating, values, and testing.

**What it does:**
1. Scans for existing Helm charts or Kustomize overlays
2. Reviews chart structure, templates, and values engineering
3. Checks dependency management and hook configuration
4. Suggests or generates improvements

**Flags:**
- `--values` — Values.yaml design for multi-environment deployments
- `--test` — Chart testing, linting, and CI validation
- `--deps` — Dependency management and library charts
- `--release` — Release management, upgrades, and rollbacks

**Routes to:** `helm-engineer` agent

### `security` — Container Security

Secures containerized applications across build, deploy, and runtime.

**What it does:**
1. Scans images for vulnerabilities (Trivy, Grype)
2. Reviews Kubernetes security contexts and RBAC
3. Checks secrets management and network policies
4. Provides security hardening recommendations

**Flags:**
- `--scan` — Image vulnerability scanning and SBOM generation
- `--policies` — Pod Security Standards and Kyverno/OPA policies
- `--secrets` — Secrets management setup (ESO, Vault, Sealed Secrets)
- `--supply-chain` — Image signing, provenance, admission policies
- `--audit` — Full security audit with CIS benchmarks

**Routes to:** `container-security-expert` agent

## Auto-Detection

When no subcommand is specified, `/docker-k8s` auto-detects your setup:

1. **Security issue detected** (root user, secrets in image, no network policies) → Routes to `container-security-expert`
2. **Helm chart detected** (Chart.yaml) → Routes to `helm-engineer`
3. **Kubernetes manifests detected** (k8s/, deploy/) → Routes to `kubernetes-architect`
4. **Dockerfile detected** → Routes to `docker-expert`
5. **No containers** → Routes to `docker-expert` for initial containerization

## Agent Selection Guide

| Need | Agent | Command |
|------|-------|---------|
| Dockerfile creation | docker-expert | `/docker-k8s docker` |
| Multi-stage builds | docker-expert | `/docker-k8s docker --multi-stage` |
| Docker Compose | docker-expert | `/docker-k8s docker --compose` |
| Image optimization | docker-expert | `/docker-k8s docker --optimize` |
| Kubernetes deployment | kubernetes-architect | `/docker-k8s k8s` |
| Autoscaling | kubernetes-architect | `/docker-k8s k8s --scale` |
| Ingress/networking | kubernetes-architect | `/docker-k8s k8s --network` |
| RBAC/namespaces | kubernetes-architect | `/docker-k8s k8s --rbac` |
| Helm chart creation | helm-engineer | `/docker-k8s helm` |
| Values engineering | helm-engineer | `/docker-k8s helm --values` |
| Chart testing | helm-engineer | `/docker-k8s helm --test` |
| Vulnerability scan | container-security-expert | `/docker-k8s security --scan` |
| Security policies | container-security-expert | `/docker-k8s security --policies` |
| Secrets management | container-security-expert | `/docker-k8s security --secrets` |
| Supply chain | container-security-expert | `/docker-k8s security --supply-chain` |
| Full review | All agents | `/docker-k8s --review` |

## Reference Materials

This suite includes comprehensive reference documents in `references/`:

- **kubernetes-patterns.md** — Sidecar, init containers, jobs, CronJobs, CRDs, operators, service mesh patterns
- **docker-best-practices.md** — Image optimization, .dockerignore, health checks, logging, multi-platform builds
- **ci-cd-containers.md** — GitHub Actions, GitLab CI, ArgoCD, Tekton pipelines for container workflows

Agents automatically consult these references when working. You can also read them directly for quick answers.

## How It Works

1. You describe what you need (e.g., "containerize my Node.js app for production")
2. The command analyzes your project structure and existing container configuration
3. It routes to the appropriate specialist agent
4. The agent reads your code, understands your patterns, and generates solutions
5. Configurations are written following production best practices

All generated configurations follow these principles:
- **Security**: Non-root, read-only filesystem, least privilege, secrets management
- **Performance**: Multi-stage builds, layer caching, resource limits, autoscaling
- **Reliability**: Health checks, graceful shutdown, PDB, rolling updates
- **Observability**: Prometheus metrics, structured logging, distributed tracing
- **GitOps**: Declarative configs, Helm charts, sealed secrets, image signing
