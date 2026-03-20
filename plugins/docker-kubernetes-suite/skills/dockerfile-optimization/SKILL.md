---
name: dockerfile-optimization
description: >
  Dockerfile best practices — multi-stage builds, layer caching,
  BuildKit features, security hardening, and image size optimization.
  Triggers: "dockerfile", "docker build", "multi-stage build", "docker optimize",
  "docker image size", "docker security", "buildkit", ".dockerignore",
  "docker layer cache", "distroless", "docker best practices".
  NOT for: Kubernetes deployments (use kubernetes-deployments), Docker Compose services.
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Dockerfile Optimization

## Multi-Stage Build (Node.js)

```dockerfile
# syntax=docker/dockerfile:1

# Stage 1: Dependencies
FROM node:20-alpine AS deps
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci --production=false

# Stage 2: Build
FROM node:20-alpine AS build
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .
RUN npm run build
# Prune dev dependencies
RUN npm prune --production

# Stage 3: Production
FROM node:20-alpine AS production
WORKDIR /app

# Security: run as non-root
RUN addgroup -g 1001 -S appgroup && \
    adduser -S appuser -u 1001 -G appgroup

# Only copy what's needed
COPY --from=build --chown=appuser:appgroup /app/dist ./dist
COPY --from=build --chown=appuser:appgroup /app/node_modules ./node_modules
COPY --from=build --chown=appuser:appgroup /app/package.json ./

# Security headers
ENV NODE_ENV=production
EXPOSE 3000

USER appuser
CMD ["node", "dist/server.js"]
```

## Multi-Stage Build (TypeScript + Prisma)

```dockerfile
FROM node:20-alpine AS deps
WORKDIR /app
COPY package.json package-lock.json ./
COPY prisma ./prisma/
RUN npm ci
RUN npx prisma generate

FROM node:20-alpine AS build
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .
RUN npm run build
RUN npm prune --production
# Re-generate Prisma client for production node_modules
RUN npx prisma generate

FROM node:20-alpine AS production
WORKDIR /app
RUN addgroup -g 1001 -S app && adduser -S app -u 1001 -G app

COPY --from=build --chown=app:app /app/dist ./dist
COPY --from=build --chown=app:app /app/node_modules ./node_modules
COPY --from=build --chown=app:app /app/package.json ./
COPY --from=build --chown=app:app /app/prisma ./prisma

ENV NODE_ENV=production
EXPOSE 3000
USER app

# Run migrations then start
CMD ["sh", "-c", "npx prisma migrate deploy && node dist/server.js"]
```

## Layer Caching Strategy

```dockerfile
# BAD: Any code change invalidates npm install
COPY . .
RUN npm ci

# GOOD: Dependencies cached unless package.json changes
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
```

### Cache Ordering (Most → Least Stable)

```dockerfile
# 1. System packages (rarely change)
FROM node:20-alpine
RUN apk add --no-cache curl

# 2. Dependencies (change occasionally)
COPY package.json package-lock.json ./
RUN npm ci

# 3. Prisma schema (changes sometimes)
COPY prisma ./prisma/
RUN npx prisma generate

# 4. Source code (changes frequently)
COPY . .
RUN npm run build
```

## BuildKit Features

```dockerfile
# syntax=docker/dockerfile:1

# Cache mounts (persistent across builds)
FROM node:20-alpine AS build
WORKDIR /app
COPY package.json package-lock.json ./
RUN --mount=type=cache,target=/root/.npm \
    npm ci

# Secret mounts (not stored in image layers)
RUN --mount=type=secret,id=npmrc,target=/root/.npmrc \
    npm ci

# SSH mounts (for private repos)
RUN --mount=type=ssh \
    git clone git@github.com:private/repo.git

# Bind mounts (avoid COPY for build-only files)
RUN --mount=type=bind,source=package.json,target=package.json \
    --mount=type=bind,source=package-lock.json,target=package-lock.json \
    --mount=type=cache,target=/root/.npm \
    npm ci
```

### Build with BuildKit

```bash
# Enable BuildKit
DOCKER_BUILDKIT=1 docker build -t myapp .

# Or set globally
export DOCKER_BUILDKIT=1

# With secrets
docker build --secret id=npmrc,src=.npmrc -t myapp .

# With SSH
docker build --ssh default -t myapp .

# Multi-platform build
docker buildx build --platform linux/amd64,linux/arm64 -t myapp:latest --push .
```

## .dockerignore

```
# Version control
.git
.gitignore

# Dependencies (installed in container)
node_modules

# Build output (built in container)
dist
build
.next

# Development files
.env
.env.local
.env.*.local
*.md
LICENSE
docker-compose*.yml
Dockerfile*

# IDE
.vscode
.idea
*.swp

# Testing
coverage
.nyc_output
__tests__
*.test.ts
*.spec.ts

# OS
.DS_Store
Thumbs.db
```

## Security Hardening

```dockerfile
# 1. Use specific version tags (never :latest)
FROM node:20.11.1-alpine3.19

# 2. Run as non-root user
RUN addgroup -g 1001 -S app && \
    adduser -S -u 1001 -G app -h /app -s /sbin/nologin app
USER app

# 3. Drop all capabilities
# (in docker run: --cap-drop=ALL)

# 4. Read-only filesystem
# (in docker run: --read-only --tmpfs /tmp)

# 5. No new privileges
# (in docker run: --security-opt=no-new-privileges:true)

# 6. Scan for vulnerabilities
# docker scout cves myapp:latest
# trivy image myapp:latest

# 7. Use distroless for minimal attack surface
FROM gcr.io/distroless/nodejs20-debian12 AS production
COPY --from=build /app/dist /app/dist
COPY --from=build /app/node_modules /app/node_modules
WORKDIR /app
CMD ["dist/server.js"]
```

## Image Size Optimization

| Technique | Savings |
|-----------|---------|
| Alpine base (`node:20-alpine`) | ~900MB → ~180MB |
| Multi-stage (no dev deps) | ~400MB → ~150MB |
| Distroless | ~180MB → ~120MB |
| `.dockerignore` | Prevents bloated context |
| `npm ci --production` | No devDependencies |
| `npm prune --production` | Remove dev deps after build |

```bash
# Check image size
docker images myapp

# Analyze layers
docker history myapp:latest

# Detailed analysis with dive
dive myapp:latest
```

## Health Checks

```dockerfile
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
  CMD curl -f http://localhost:3000/health || exit 1

# Or without curl (smaller image)
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
  CMD node -e "require('http').get('http://localhost:3000/health', (r) => { process.exit(r.statusCode === 200 ? 0 : 1) })"
```

## Docker Compose for Development

```yaml
# docker-compose.yml
services:
  app:
    build:
      context: .
      dockerfile: Dockerfile
      target: deps  # stop at deps stage for dev
    volumes:
      - .:/app            # mount source for hot reload
      - /app/node_modules # preserve container's node_modules
    ports:
      - "3000:3000"
    environment:
      - NODE_ENV=development
      - DATABASE_URL=postgresql://postgres:postgres@db:5432/myapp
    depends_on:
      db:
        condition: service_healthy
    command: npm run dev

  db:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: myapp
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
    ports:
      - "5432:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 5s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

volumes:
  pgdata:
```

## Common Patterns

### Graceful Shutdown

```dockerfile
# Use tini as PID 1 (handles signals correctly)
RUN apk add --no-cache tini
ENTRYPOINT ["/sbin/tini", "--"]
CMD ["node", "dist/server.js"]
```

```typescript
// In your Node.js app
process.on("SIGTERM", async () => {
  console.log("SIGTERM received, shutting down gracefully");
  server.close(() => {
    db.$disconnect();
    process.exit(0);
  });
  setTimeout(() => process.exit(1), 10000); // force after 10s
});
```

### Environment-Specific Builds

```dockerfile
FROM node:20-alpine AS base
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci

FROM base AS development
COPY . .
CMD ["npm", "run", "dev"]

FROM base AS production
COPY . .
RUN npm run build && npm prune --production
CMD ["node", "dist/server.js"]
```

```bash
# Build for specific target
docker build --target development -t myapp:dev .
docker build --target production -t myapp:prod .
```

## Gotchas

1. **`COPY . .` before `npm ci` busts the cache on every code change.** Always copy `package.json` and lock file first, install dependencies, THEN copy source code. This is the single most impactful optimization.

2. **Alpine uses musl libc, not glibc.** Some npm packages with native bindings (sharp, bcrypt, canvas) need different build steps on Alpine. If builds fail, either install build deps (`apk add python3 make g++`) or use the slim Debian variant instead.

3. **`npm install` is not deterministic, `npm ci` is.** Always use `npm ci` in Dockerfiles. It installs exactly what's in `package-lock.json` and is faster because it deletes `node_modules` first.

4. **COPY --chown requires BuildKit.** Without `DOCKER_BUILDKIT=1`, the `--chown` flag on COPY silently fails. Always enable BuildKit for modern Dockerfile features.

5. **Node.js doesn't handle SIGTERM as PID 1.** Inside a container, your app is PID 1 and doesn't receive signals correctly without a proper init system. Use `tini` or `dumb-init`, or `docker run --init`.

6. **`docker build` sends the entire context directory.** Without `.dockerignore`, your `node_modules`, `.git`, and other large directories are sent to the Docker daemon on every build. A proper `.dockerignore` can reduce build context from GBs to KBs.
