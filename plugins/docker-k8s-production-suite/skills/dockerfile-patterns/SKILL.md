---
name: dockerfile-patterns
description: >
  Production Dockerfile patterns — multi-stage builds, layer caching, security hardening,
  and language-specific optimized Dockerfiles for Node.js, Python, Go, and Rust.
  Triggers: "Dockerfile", "docker build", "multi-stage build", "container image",
  "optimize docker image", "reduce image size".
  NOT for: Docker Compose (use docker-compose skill), Kubernetes (use kubernetes-deployments).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Dockerfile Patterns

## Multi-Stage Build Template

```dockerfile
# Stage 1: Build
FROM node:22-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci --omit=dev
COPY . .
RUN npm run build

# Stage 2: Production
FROM node:22-alpine AS production
WORKDIR /app
RUN addgroup -g 1001 appgroup && adduser -u 1001 -G appgroup -s /bin/sh -D appuser
COPY --from=builder --chown=appuser:appgroup /app/dist ./dist
COPY --from=builder --chown=appuser:appgroup /app/node_modules ./node_modules
COPY --from=builder --chown=appuser:appgroup /app/package.json ./
USER appuser
EXPOSE 3000
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
  CMD wget -qO- http://localhost:3000/health || exit 1
CMD ["node", "dist/index.js"]
```

## Node.js (TypeScript + Express/Fastify)

```dockerfile
# syntax=docker/dockerfile:1

FROM node:22-alpine AS base
WORKDIR /app
RUN corepack enable

# Dependencies stage — cached unless package files change
FROM base AS deps
COPY package.json pnpm-lock.yaml ./
RUN --mount=type=cache,target=/root/.local/share/pnpm/store \
    pnpm install --frozen-lockfile

# Build stage
FROM base AS builder
COPY --from=deps /app/node_modules ./node_modules
COPY . .
RUN pnpm build
RUN pnpm prune --prod

# Production stage
FROM base AS production
ENV NODE_ENV=production

RUN addgroup -g 1001 -S nodejs && \
    adduser -S nodejs -u 1001 -G nodejs

COPY --from=builder --chown=nodejs:nodejs /app/dist ./dist
COPY --from=builder --chown=nodejs:nodejs /app/node_modules ./node_modules
COPY --from=builder --chown=nodejs:nodejs /app/package.json ./

USER nodejs
EXPOSE 3000

HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
  CMD wget -qO- http://localhost:3000/health || exit 1

CMD ["node", "dist/index.js"]
```

## Next.js (Standalone Output)

```dockerfile
FROM node:22-alpine AS base

# Dependencies
FROM base AS deps
WORKDIR /app
COPY package.json pnpm-lock.yaml ./
RUN corepack enable && pnpm install --frozen-lockfile

# Build
FROM base AS builder
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .

ENV NEXT_TELEMETRY_DISABLED=1
RUN corepack enable && pnpm build

# Production
FROM base AS runner
WORKDIR /app
ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1

RUN addgroup -g 1001 -S nodejs && \
    adduser -S nextjs -u 1001 -G nodejs

# Copy standalone output
COPY --from=builder /app/public ./public
COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/.next/static ./.next/static

USER nextjs
EXPOSE 3000
ENV PORT=3000
ENV HOSTNAME="0.0.0.0"

CMD ["node", "server.js"]
```

Requires `output: 'standalone'` in `next.config.js`.

## Python (FastAPI / Django)

```dockerfile
FROM python:3.12-slim AS builder
WORKDIR /app

# Install build dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    && rm -rf /var/lib/apt/lists/*

# Create virtualenv and install dependencies
RUN python -m venv /opt/venv
ENV PATH="/opt/venv/bin:$PATH"

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Production stage
FROM python:3.12-slim AS production
WORKDIR /app

# Copy virtualenv from builder
COPY --from=builder /opt/venv /opt/venv
ENV PATH="/opt/venv/bin:$PATH"

# Create non-root user
RUN groupadd -g 1001 appgroup && \
    useradd -u 1001 -g appgroup -s /bin/bash appuser

COPY --chown=appuser:appgroup . .

USER appuser
EXPOSE 8000

HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
  CMD python -c "import urllib.request; urllib.request.urlopen('http://localhost:8000/health')" || exit 1

CMD ["gunicorn", "app.main:app", "-w", "4", "-k", "uvicorn.workers.UvicornWorker", "-b", "0.0.0.0:8000"]
```

## Go (Minimal Binary)

```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app

# Dependencies (cached)
COPY go.mod go.sum ./
RUN go mod download

# Build static binary
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/server ./cmd/server

# Minimal production image
FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/server /server

EXPOSE 8080
ENTRYPOINT ["/server"]
```

**Result**: ~10-15MB final image (just the binary + CA certs).

For debugging, use `gcr.io/distroless/static-debian12` instead of `scratch` — adds basic
debugging capability without a full OS.

## Rust

```dockerfile
FROM rust:1.78-slim AS builder
WORKDIR /app

# Cache dependencies
COPY Cargo.toml Cargo.lock ./
RUN mkdir src && echo "fn main() {}" > src/main.rs
RUN cargo build --release
RUN rm -rf src

# Build actual application
COPY . .
RUN touch src/main.rs  # Force rebuild of main
RUN cargo build --release

# Minimal production image
FROM gcr.io/distroless/cc-debian12
COPY --from=builder /app/target/release/myapp /myapp
EXPOSE 8080
ENTRYPOINT ["/myapp"]
```

## Layer Caching Best Practices

### Order of Operations (Most Stable First)

```dockerfile
# 1. Base image + system deps (rarely changes)
FROM node:22-alpine
RUN apk add --no-cache dumb-init

# 2. Dependency manifests (changes when deps change)
COPY package.json pnpm-lock.yaml ./

# 3. Install dependencies (cached unless manifests change)
RUN pnpm install --frozen-lockfile

# 4. Application code (changes most often — LAST)
COPY . .

# 5. Build step
RUN pnpm build
```

### BuildKit Cache Mounts

```dockerfile
# Cache npm/pnpm store across builds
RUN --mount=type=cache,target=/root/.npm \
    npm ci

# Cache apt packages
RUN --mount=type=cache,target=/var/cache/apt \
    --mount=type=cache,target=/var/lib/apt \
    apt-get update && apt-get install -y build-essential

# Cache Go modules
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Cache pip packages
RUN --mount=type=cache,target=/root/.cache/pip \
    pip install -r requirements.txt
```

## .dockerignore

```
# Always ignore these
.git
.gitignore
node_modules
.env
.env.*
*.md
LICENSE
docker-compose*.yml
Dockerfile*
.dockerignore

# Build artifacts
dist
build
.next
__pycache__
*.pyc
target

# IDE
.vscode
.idea
*.swp
*.swo

# OS
.DS_Store
Thumbs.db

# Test
coverage
.nyc_output
*.test.*
__tests__
```

## Security Hardening

```dockerfile
# 1. Use specific version tags (never :latest)
FROM node:22.2.0-alpine3.19

# 2. Run as non-root user
RUN addgroup -g 1001 -S app && adduser -S app -u 1001 -G app
USER app

# 3. Read-only filesystem (set at runtime)
# docker run --read-only --tmpfs /tmp myimage

# 4. No new privileges
# docker run --security-opt no-new-privileges myimage

# 5. Drop all capabilities
# docker run --cap-drop ALL myimage

# 6. Scan for vulnerabilities
# docker scout cves myimage
# trivy image myimage

# 7. Don't store secrets in image
# Use build secrets instead:
RUN --mount=type=secret,id=npmrc,target=/root/.npmrc \
    npm ci
```

## HEALTHCHECK Patterns

```dockerfile
# HTTP health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
  CMD wget -qO- http://localhost:3000/health || exit 1

# TCP port check (no HTTP endpoint)
HEALTHCHECK --interval=30s --timeout=3s --retries=3 \
  CMD nc -z localhost 5432 || exit 1

# Custom script
COPY healthcheck.sh /healthcheck.sh
HEALTHCHECK --interval=30s --timeout=5s --retries=3 \
  CMD /healthcheck.sh
```

## Gotchas

1. **Layer caching is order-dependent** — put `COPY package*.json` BEFORE `COPY .` to cache dependencies. If you `COPY .` first, every code change invalidates the npm install cache.

2. **`npm install` vs `npm ci`** — always use `npm ci` in Docker. It's faster, deterministic, and respects the lockfile exactly.

3. **Alpine uses musl, not glibc** — some native Node.js modules (sharp, bcrypt) need musl-compatible builds. If you hit segfaults, switch to `-slim` (Debian-based).

4. **Don't run as root** — always add a non-root user and `USER` directive. Running as root in production is a security risk.

5. **Each RUN creates a layer** — combine related commands with `&&` to reduce layers. But don't combine unrelated commands (breaks caching).

6. **Secrets in build args persist in image history** — use `--mount=type=secret` for build-time secrets, not `ARG`.

7. **`.dockerignore` is essential** — without it, `COPY .` includes `node_modules`, `.git`, `.env`, and other files that bloat the image and leak secrets.

8. **ENTRYPOINT vs CMD** — `ENTRYPOINT` sets the executable (hard to override). `CMD` sets defaults (easily overridden). Use `ENTRYPOINT` for the main binary, `CMD` for default arguments.
