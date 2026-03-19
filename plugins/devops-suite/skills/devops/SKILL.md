---
name: devops-suite
description: >
  DevOps & Deployment Suite — AI-powered toolkit for containerization, CI/CD pipelines,
  cloud deployment, and monitoring. Creates production-ready Dockerfiles, GitHub Actions
  workflows, deploys to any cloud platform, and sets up observability.
  Triggers: "docker", "dockerfile", "containerize", "ci/cd", "pipeline", "github actions",
  "deploy", "deployment", "heroku", "vercel", "cloudflare", "aws", "monitoring", "logging",
  "health check", "devops", "infrastructure", "cloud", "docker-compose".
  Dispatches the appropriate specialist agent: dockerfile-builder, ci-cd-architect,
  cloud-deployer, or monitoring-setup.
  NOT for: Kubernetes cluster management, Terraform/Pulumi IaC, database administration,
  or network security configuration.
version: 1.0.0
argument-hint: "<docker|ci-cd|deploy|monitor> [target]"
user-invocable: true
allowed-tools: Read, Grep, Glob, Bash
model: sonnet
---

# DevOps & Deployment Suite

Production-grade DevOps agents for Claude Code. Four specialist agents that handle containerization, CI/CD pipelines, cloud deployment, and monitoring — the infrastructure work that slows down every project.

## Available Agents

### Dockerfile Builder (`dockerfile-builder`)
Creates optimized, production-ready Docker configurations. Multi-stage builds, security hardening, Docker Compose for multi-service apps, and platform-specific optimizations.

**Invoke**: Dispatch via Task tool with `subagent_type: "dockerfile-builder"`.

**Example prompts**:
- "Create a Dockerfile for my Node.js Express app with a React frontend"
- "Set up Docker Compose with my API, PostgreSQL, Redis, and a background worker"
- "Optimize my existing Dockerfile — it's 2GB and takes 10 minutes to build"

### CI/CD Architect (`ci-cd-architect`)
Designs complete CI/CD pipelines for any platform. GitHub Actions, GitLab CI, matrix testing, deployment strategies, and caching optimization.

**Invoke**: Dispatch via Task tool with `subagent_type: "ci-cd-architect"`.

**Example prompts**:
- "Create a GitHub Actions workflow that tests, builds, and deploys on merge to main"
- "Set up matrix testing across Node 18, 20, 22 on Ubuntu and macOS"
- "Add a canary deployment stage to my existing pipeline"

### Cloud Deployer (`cloud-deployer`)
Deploys applications to any cloud platform. Generates platform-specific configs, handles environment setup, domain configuration, and database provisioning.

**Invoke**: Dispatch via Task tool with `subagent_type: "cloud-deployer"`.

**Example prompts**:
- "Deploy my app to Heroku with PostgreSQL and Redis add-ons"
- "Set up Vercel deployment with serverless API routes"
- "Deploy to AWS ECS with Fargate — I need auto-scaling"

### Monitoring Setup (`monitoring-setup`)
Sets up observability for your application. Structured logging, error tracking, health checks, uptime monitoring, alerting, and dashboards.

**Invoke**: Dispatch via Task tool with `subagent_type: "monitoring-setup"`.

**Example prompts**:
- "Add Sentry error tracking and structured logging to my Express API"
- "Set up health check endpoints with dependency checks"
- "Configure Datadog monitoring with custom dashboards and Slack alerts"

## Quick Start: /deploy

Use the `/deploy` command for a guided, end-to-end deployment workflow:

```
/deploy              # Auto-detect and deploy
/deploy heroku       # Deploy to Heroku
/deploy aws          # Deploy to AWS
```

The `/deploy` command chains all four agents: detect → containerize → pipeline → deploy → monitor.

## Agent Selection Guide

| Need | Agent | Command |
|------|-------|---------|
| Containerize an app | dockerfile-builder | "Create a Dockerfile" |
| Docker Compose setup | dockerfile-builder | "Set up Docker Compose" |
| CI/CD pipeline | ci-cd-architect | "Create GitHub Actions workflow" |
| Deploy to cloud | cloud-deployer | "Deploy to [platform]" |
| Set up logging | monitoring-setup | "Add structured logging" |
| Error tracking | monitoring-setup | "Set up Sentry" |
| Health checks | monitoring-setup | "Add health check endpoints" |
| Full deployment | /deploy command | "/deploy [platform]" |

## Reference Materials

This skill includes comprehensive reference documents in `references/`:

- **docker-patterns.md** — Multi-stage build templates, security hardening, Compose patterns, optimization techniques
- **ci-cd-templates.md** — Copy-paste ready pipeline configs for GitHub Actions and GitLab CI
- **cloud-deployment-guides.md** — Step-by-step deployment guides for Heroku, Vercel, Cloudflare, AWS, GCP, Railway, Fly.io

Agents automatically consult these references when working. You can also read them directly for quick answers.

## How It Works

1. You describe what you need (e.g., "containerize my app")
2. The SKILL.md routes to the appropriate agent
3. The agent reads your project structure, detects the stack, and generates production-ready configs
4. Configs are written directly to your project
5. The agent provides next steps and deployment commands

All generated configs follow industry best practices:
- Security: Non-root containers, minimal base images, secret management
- Performance: Layer caching, multi-stage builds, efficient CI caching
- Reliability: Health checks, graceful shutdown, rolling deployments
- Observability: Structured logging, error tracking, monitoring
