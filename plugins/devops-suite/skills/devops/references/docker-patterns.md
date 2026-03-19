# Docker Patterns Reference

Advanced Docker patterns, multi-stage build templates, security hardening, Docker Compose configurations, and optimization techniques.

## Multi-Stage Build Templates

### Node.js (Express / Fastify / NestJS)

```dockerfile
# ============================================
# Node.js Multi-Stage Production Build
# ============================================

# Stage 1: Install production dependencies only
FROM node:20-bookworm-slim AS deps
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci --omit=dev && npm cache clean --force

# Stage 2: Build the application
FROM node:20-bookworm AS build
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY tsconfig.json ./
COPY src ./src
RUN npm run build

# Stage 3: Production image
FROM node:20-bookworm-slim AS production
WORKDIR /app

# Security: create non-root user
RUN groupadd -r appuser && useradd -r -g appuser -d /app -s /sbin/nologin appuser

# Copy only what's needed
COPY --from=deps --chown=appuser:appuser /app/node_modules ./node_modules
COPY --from=build --chown=appuser:appuser /app/dist ./dist
COPY --chown=appuser:appuser package.json ./

# Security: switch to non-root
USER appuser

ENV NODE_ENV=production
EXPOSE 3000

HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
  CMD node -e "const http=require('http');const req=http.get('http://localhost:3000/health',(res)=>{process.exit(res.statusCode===200?0:1)});req.on('error',()=>process.exit(1));req.end()"

CMD ["node", "dist/index.js"]
```

### Node.js with pnpm (Monorepo / Turborepo)

```dockerfile
FROM node:20-bookworm-slim AS base
RUN corepack enable && corepack prepare pnpm@9.0.0 --activate
WORKDIR /app

# Stage 1: Install all dependencies (for building)
FROM base AS deps
COPY pnpm-lock.yaml pnpm-workspace.yaml package.json ./
COPY apps/api/package.json ./apps/api/
COPY apps/web/package.json ./apps/web/
COPY packages/db/package.json ./packages/db/
COPY packages/shared/package.json ./packages/shared/
RUN pnpm install --frozen-lockfile

# Stage 2: Build
FROM base AS build
COPY --from=deps /app/node_modules ./node_modules
COPY --from=deps /app/apps/api/node_modules ./apps/api/node_modules
COPY --from=deps /app/apps/web/node_modules ./apps/web/node_modules
COPY --from=deps /app/packages/db/node_modules ./packages/db/node_modules
COPY --from=deps /app/packages/shared/node_modules ./packages/shared/node_modules
COPY . .
RUN pnpm turbo build --filter=api

# Stage 3: Production deps only
FROM base AS prod-deps
COPY pnpm-lock.yaml pnpm-workspace.yaml package.json ./
COPY apps/api/package.json ./apps/api/
COPY packages/db/package.json ./packages/db/
COPY packages/shared/package.json ./packages/shared/
RUN pnpm install --frozen-lockfile --prod

# Stage 4: Runtime
FROM node:20-bookworm-slim
WORKDIR /app
RUN groupadd -r app && useradd -r -g app -d /app app
COPY --from=prod-deps --chown=app:app /app/node_modules ./node_modules
COPY --from=prod-deps --chown=app:app /app/apps/api/node_modules ./apps/api/node_modules
COPY --from=prod-deps --chown=app:app /app/packages/db/node_modules ./packages/db/node_modules
COPY --from=build --chown=app:app /app/apps/api/dist ./apps/api/dist
COPY --from=build --chown=app:app /app/packages/db/dist ./packages/db/dist
COPY --from=build --chown=app:app /app/packages/shared/dist ./packages/shared/dist
USER app
ENV NODE_ENV=production
EXPOSE 3001
CMD ["node", "apps/api/dist/index.js"]
```

### Python (FastAPI / Django)

```dockerfile
# ============================================
# Python Multi-Stage Production Build
# ============================================

FROM python:3.12-slim-bookworm AS base
ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1 \
    PIP_NO_CACHE_DIR=1 \
    PIP_DISABLE_PIP_VERSION_CHECK=1

# Stage 1: Build dependencies (some need gcc for native extensions)
FROM python:3.12-bookworm AS build
WORKDIR /app
COPY requirements.txt ./
RUN pip install --no-cache-dir --prefix=/install -r requirements.txt

# Stage 2: Production
FROM base AS production
WORKDIR /app

# Copy installed packages from build stage
COPY --from=build /install /usr/local

# Security: non-root user
RUN groupadd -r app && useradd -r -g app -d /app -s /sbin/nologin app

COPY --chown=app:app . .
USER app

EXPOSE 8000

HEALTHCHECK --interval=30s --timeout=3s --retries=3 \
  CMD python -c "import urllib.request; urllib.request.urlopen('http://localhost:8000/health')"

# FastAPI with gunicorn + uvicorn workers
CMD ["gunicorn", "app.main:app", \
     "--worker-class", "uvicorn.workers.UvicornWorker", \
     "--bind", "0.0.0.0:8000", \
     "--workers", "4", \
     "--timeout", "120", \
     "--graceful-timeout", "30", \
     "--access-logfile", "-", \
     "--error-logfile", "-"]
```

### Python with Poetry

```dockerfile
FROM python:3.12-slim-bookworm AS base
ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1

FROM base AS deps
RUN pip install poetry==1.8.0
WORKDIR /app
COPY pyproject.toml poetry.lock ./
RUN poetry config virtualenvs.create false \
    && poetry install --no-interaction --no-ansi --only main

FROM base AS production
WORKDIR /app
RUN groupadd -r app && useradd -r -g app -d /app app
COPY --from=deps /usr/local/lib/python3.12/site-packages /usr/local/lib/python3.12/site-packages
COPY --from=deps /usr/local/bin /usr/local/bin
COPY --chown=app:app . .
USER app
EXPOSE 8000
CMD ["gunicorn", "app.main:app", "-w", "4", "-k", "uvicorn.workers.UvicornWorker", "-b", "0.0.0.0:8000"]
```

### Go

```dockerfile
# ============================================
# Go Multi-Stage Production Build
# ============================================

FROM golang:1.22-bookworm AS build
WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.version=$(git describe --tags --always 2>/dev/null || echo 'unknown')" \
    -o /app/server \
    ./cmd/server

# Production: distroless for minimal attack surface
FROM gcr.io/distroless/static-debian12
COPY --from=build /app/server /server
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/server"]
```

### Go with CGO (SQLite, etc.)

```dockerfile
FROM golang:1.22-bookworm AS build
WORKDIR /app
RUN apt-get update && apt-get install -y --no-install-recommends gcc libc6-dev && rm -rf /var/lib/apt/lists/*
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-w -s" -o /app/server ./cmd/server

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && rm -rf /var/lib/apt/lists/*
RUN groupadd -r app && useradd -r -g app app
COPY --from=build /app/server /usr/local/bin/server
USER app
EXPOSE 8080
CMD ["server"]
```

### Rust

```dockerfile
# ============================================
# Rust Multi-Stage Production Build
# ============================================

FROM rust:1.77-bookworm AS build
WORKDIR /app

# Cache dependency compilation with dummy project
COPY Cargo.toml Cargo.lock ./
RUN mkdir src && echo "fn main() {}" > src/main.rs
RUN cargo build --release
RUN rm -rf src

# Build actual application
COPY src ./src
RUN touch src/main.rs && cargo build --release

# Production: minimal image
FROM gcr.io/distroless/cc-debian12
COPY --from=build /app/target/release/myapp /
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/myapp"]
```

### Java (Spring Boot with Maven)

```dockerfile
# ============================================
# Java Multi-Stage Production Build
# ============================================

FROM eclipse-temurin:21-jdk AS build
WORKDIR /app

# Cache Maven dependencies
COPY pom.xml mvnw ./
COPY .mvn .mvn
RUN chmod +x mvnw && ./mvnw dependency:resolve dependency:resolve-plugins

# Build
COPY src ./src
RUN ./mvnw package -DskipTests -Dspring-boot.build-image.skip=true

# Extract Spring Boot layers for better caching
FROM eclipse-temurin:21-jdk AS extract
WORKDIR /app
COPY --from=build /app/target/*.jar app.jar
RUN java -Djarmode=layertools -jar app.jar extract

# Production
FROM eclipse-temurin:21-jre-alpine
WORKDIR /app
RUN addgroup -S app && adduser -S app -G app

COPY --from=extract --chown=app:app /app/dependencies/ ./
COPY --from=extract --chown=app:app /app/spring-boot-loader/ ./
COPY --from=extract --chown=app:app /app/snapshot-dependencies/ ./
COPY --from=extract --chown=app:app /app/application/ ./

USER app
EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --retries=3 \
  CMD wget -qO- http://localhost:8080/actuator/health || exit 1

ENTRYPOINT ["java", \
  "-XX:MaxRAMPercentage=75.0", \
  "-XX:+UseG1GC", \
  "-XX:+UseContainerSupport", \
  "org.springframework.boot.loader.launch.JarLauncher"]
```

### Java with Gradle

```dockerfile
FROM eclipse-temurin:21-jdk AS build
WORKDIR /app
COPY build.gradle.kts settings.gradle.kts gradlew ./
COPY gradle ./gradle
RUN chmod +x gradlew && ./gradlew dependencies --no-daemon
COPY src ./src
RUN ./gradlew bootJar --no-daemon -x test

FROM eclipse-temurin:21-jre-alpine
WORKDIR /app
RUN addgroup -S app && adduser -S app -G app
COPY --from=build --chown=app:app /app/build/libs/*.jar app.jar
USER app
EXPOSE 8080
ENTRYPOINT ["java", "-XX:MaxRAMPercentage=75.0", "-jar", "app.jar"]
```

### Ruby (Rails)

```dockerfile
FROM ruby:3.3-bookworm AS build
WORKDIR /app
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential libpq-dev nodejs npm && \
    rm -rf /var/lib/apt/lists/*

COPY Gemfile Gemfile.lock ./
RUN bundle config set --local deployment true && \
    bundle config set --local without 'development test' && \
    bundle install

COPY . .
RUN RAILS_ENV=production SECRET_KEY_BASE=placeholder bundle exec rake assets:precompile

FROM ruby:3.3-slim-bookworm
WORKDIR /app
RUN apt-get update && apt-get install -y --no-install-recommends libpq5 && rm -rf /var/lib/apt/lists/*
RUN groupadd -r app && useradd -r -g app -d /app app
COPY --from=build --chown=app:app /app /app
USER app
ENV RAILS_ENV=production RAILS_SERVE_STATIC_FILES=true RAILS_LOG_TO_STDOUT=true
EXPOSE 3000
CMD ["bundle", "exec", "puma", "-C", "config/puma.rb"]
```

## Docker Compose Patterns

### Development Environment (Full Stack)

```yaml
# docker-compose.yml — base configuration
services:
  app:
    build:
      context: .
      dockerfile: Dockerfile
      target: production
    ports:
      - "${PORT:-3000}:3000"
    environment:
      - DATABASE_URL=postgresql://postgres:postgres@db:5432/myapp
      - REDIS_URL=redis://redis:6379
      - NODE_ENV=production
    depends_on:
      db:
        condition: service_healthy
      redis:
        condition: service_healthy
    restart: unless-stopped
    deploy:
      resources:
        limits:
          memory: 512M
          cpus: "1.0"
    healthcheck:
      test: ["CMD", "node", "-e", "require('http').get('http://localhost:3000/health',(r)=>process.exit(r.statusCode===200?0:1))"]
      interval: 30s
      timeout: 5s
      retries: 3
      start_period: 15s

  db:
    image: postgres:16-alpine
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./scripts/init-db.sql:/docker-entrypoint-initdb.d/init.sql
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: myapp
    ports:
      - "5432:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres -d myapp"]
      interval: 10s
      timeout: 5s
      retries: 5
    restart: unless-stopped
    deploy:
      resources:
        limits:
          memory: 256M

  redis:
    image: redis:7-alpine
    volumes:
      - redis_data:/data
    ports:
      - "6379:6379"
    command: >
      redis-server
      --appendonly yes
      --maxmemory 128mb
      --maxmemory-policy allkeys-lru
      --save 60 1000
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 3s
      retries: 3
    restart: unless-stopped

  worker:
    build:
      context: .
      dockerfile: Dockerfile
      target: production
    command: ["node", "dist/worker.js"]
    environment:
      - DATABASE_URL=postgresql://postgres:postgres@db:5432/myapp
      - REDIS_URL=redis://redis:6379
      - NODE_ENV=production
    depends_on:
      db:
        condition: service_healthy
      redis:
        condition: service_healthy
    restart: unless-stopped
    deploy:
      resources:
        limits:
          memory: 256M

volumes:
  postgres_data:
  redis_data:
```

### Development Override

```yaml
# docker-compose.dev.yml — development overrides
services:
  app:
    build:
      target: deps
    command: npx tsx watch src/index.ts
    volumes:
      - .:/app
      - /app/node_modules
    environment:
      - NODE_ENV=development
      - LOG_LEVEL=debug
    ports:
      - "3000:3000"
      - "9229:9229"

  db:
    ports:
      - "5432:5432"

  redis:
    ports:
      - "6379:6379"

  # Development-only services
  mailhog:
    image: mailhog/mailhog
    ports:
      - "1025:1025"
      - "8025:8025"

  adminer:
    image: adminer
    ports:
      - "8080:8080"
    depends_on:
      - db
```

Usage: `docker compose -f docker-compose.yml -f docker-compose.dev.yml up`

### Testing Environment

```yaml
# docker-compose.test.yml
services:
  test:
    build:
      context: .
      target: build
    command: npm test -- --coverage --forceExit
    environment:
      - DATABASE_URL=postgresql://postgres:postgres@db:5432/test
      - REDIS_URL=redis://redis:6379
      - NODE_ENV=test
    depends_on:
      db:
        condition: service_healthy
      redis:
        condition: service_healthy

  db:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: test
    tmpfs:
      - /var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 5s
      timeout: 3s
      retries: 5

  redis:
    image: redis:7-alpine
    tmpfs:
      - /data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 3s
      retries: 3
```

Usage: `docker compose -f docker-compose.test.yml run --rm test`

## Security Hardening

### Non-Root User Patterns

```dockerfile
# Debian/Ubuntu-based images
RUN groupadd -r appuser --gid=1001 && \
    useradd -r -g appuser --uid=1001 -d /app -s /sbin/nologin appuser
COPY --chown=appuser:appuser . .
USER appuser

# Alpine-based images
RUN addgroup -S -g 1001 appuser && \
    adduser -S -u 1001 -G appuser -h /app -s /sbin/nologin appuser

# Distroless images (user already exists)
USER nonroot:nonroot
```

### Read-Only Filesystem

```yaml
# docker-compose.yml
services:
  app:
    read_only: true
    tmpfs:
      - /tmp
      - /var/run
    volumes:
      - app_logs:/app/logs  # Only writable directory
```

### Security Scanning

```bash
# Trivy — comprehensive vulnerability scanner
docker run --rm -v /var/run/docker.sock:/var/run/docker.sock \
  aquasec/trivy image myapp:latest

# Docker Scout (built into Docker Desktop)
docker scout cves myapp:latest
docker scout recommendations myapp:latest

# Snyk Container
snyk container test myapp:latest

# Hadolint — Dockerfile linter
docker run --rm -i hadolint/hadolint < Dockerfile
```

### Secret Management in Docker

```dockerfile
# NEVER do this:
# COPY .env .
# ENV API_KEY=sk-1234567890

# BuildKit secrets (build-time, not baked into image):
RUN --mount=type=secret,id=npm_token \
    NPM_TOKEN=$(cat /run/secrets/npm_token) npm ci

# Runtime secrets via environment:
# docker run -e API_KEY=sk-xxx myapp
# docker compose with .env file (gitignored)
```

### Minimal Base Images Comparison

| Base Image | Size | Shell | Package Manager | Best For |
|-----------|------|-------|-----------------|----------|
| `alpine:3.19` | ~7MB | ash | apk | Minimal containers |
| `debian:bookworm-slim` | ~80MB | bash | apt | General purpose |
| `gcr.io/distroless/static` | ~2MB | none | none | Go, Rust static binaries |
| `gcr.io/distroless/cc` | ~22MB | none | none | Rust with C deps |
| `gcr.io/distroless/base` | ~20MB | none | none | C/C++ apps |
| `scratch` | 0MB | none | none | Fully static binaries only |

## Image Size Optimization

### Layer Optimization

```dockerfile
# Bad: multiple RUN commands = multiple layers
RUN apt-get update
RUN apt-get install -y curl
RUN apt-get clean

# Good: single RUN command = one layer, clean in same layer
RUN apt-get update && \
    apt-get install -y --no-install-recommends curl && \
    rm -rf /var/lib/apt/lists/*
```

### .dockerignore for Minimal Build Context

```
# Version control
.git
.gitignore
.gitattributes

# Dependencies (rebuilt in Docker)
node_modules
vendor
__pycache__
*.pyc
.venv

# IDE
.vscode
.idea
*.swp
*.swo

# Environment (secrets!)
.env
.env.*
!.env.example

# Docker files (prevent recursive builds)
Dockerfile
Dockerfile.*
docker-compose*.yml
.dockerignore

# Documentation
*.md
!README.md
LICENSE
CHANGELOG*
docs/

# Tests
__tests__
*.test.*
*.spec.*
coverage
.nyc_output
jest.config.*
vitest.config.*

# CI/CD
.github
.gitlab-ci.yml
.circleci
Jenkinsfile

# OS files
.DS_Store
Thumbs.db

# Build output (rebuilt in Docker)
dist
build
out
.next
```

### Multi-Platform Builds

```bash
# Build for multiple architectures
docker buildx create --name multiplatform --use
docker buildx build --platform linux/amd64,linux/arm64 -t myapp:latest --push .
```

## BuildKit Features

### Build Arguments and Secrets

```dockerfile
# Build arguments (visible in image history)
ARG NODE_VERSION=20
FROM node:${NODE_VERSION}-slim

# BuildKit secrets (NOT visible in image history)
RUN --mount=type=secret,id=npmrc,target=/root/.npmrc npm ci

# BuildKit cache mounts (persist across builds)
RUN --mount=type=cache,target=/root/.npm npm ci
RUN --mount=type=cache,target=/root/.cache/pip pip install -r requirements.txt
RUN --mount=type=cache,target=/root/.cargo/registry cargo build --release
```

### Cache Mounts for Package Managers

```dockerfile
# npm
RUN --mount=type=cache,target=/root/.npm \
    npm ci

# pip
RUN --mount=type=cache,target=/root/.cache/pip \
    pip install -r requirements.txt

# Go modules
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Cargo
RUN --mount=type=cache,target=/usr/local/cargo/registry \
    --mount=type=cache,target=/app/target \
    cargo build --release && \
    cp target/release/myapp /usr/local/bin/

# Maven
RUN --mount=type=cache,target=/root/.m2 \
    ./mvnw package -DskipTests

# apt
RUN --mount=type=cache,target=/var/cache/apt \
    --mount=type=cache,target=/var/lib/apt \
    apt-get update && apt-get install -y curl
```

## Common Pitfalls and Solutions

### Pitfall: node_modules in Build Context

**Problem**: Docker sends GB of node_modules to the daemon.
```
Sending build context to Docker daemon  1.2GB
```

**Solution**: Add `node_modules` to `.dockerignore`. Dependencies are installed inside the container.

### Pitfall: Layer Cache Invalidation

**Problem**: Changing one source file invalidates the dependency install cache.
```dockerfile
# Bad
COPY . .
RUN npm ci  # Runs every time ANY file changes
```

**Solution**: Copy dependency files first.
```dockerfile
COPY package.json package-lock.json ./
RUN npm ci  # Only runs when package files change
COPY . .
```

### Pitfall: Zombie Processes

**Problem**: Node.js as PID 1 doesn't handle signals properly.

**Solution**: Use `tini` as init system.
```dockerfile
RUN apt-get update && apt-get install -y --no-install-recommends tini && rm -rf /var/lib/apt/lists/*
ENTRYPOINT ["tini", "--"]
CMD ["node", "dist/index.js"]
```

Or use Docker's built-in init:
```bash
docker run --init myapp
```

### Pitfall: DNS Resolution in Alpine

**Problem**: Alpine uses musl libc which has different DNS resolution behavior. Some Node.js and Python apps fail to resolve hostnames.

**Solution**: Use `-slim` (Debian-based) images instead of Alpine for Node.js and Python. If Alpine is required, install `libc6-compat`:
```dockerfile
FROM node:20-alpine
RUN apk add --no-cache libc6-compat
```

### Pitfall: Build Failures on ARM Macs

**Problem**: Building x86 images on M1/M2 Macs is slow (qemu emulation).

**Solution**: Use multi-platform builds or build natively:
```dockerfile
# Explicitly set platform for consistency
FROM --platform=linux/amd64 node:20-slim
```

Or build for your architecture and use multi-platform in CI:
```bash
# Local development (native speed)
docker build -t myapp .

# CI (multi-platform)
docker buildx build --platform linux/amd64,linux/arm64 -t myapp --push .
```

### Pitfall: Timezone Issues

**Solution**: Set timezone in the container:
```dockerfile
ENV TZ=UTC
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone
```

### Pitfall: Large Final Images

**Diagnosis**: Check what's taking space:
```bash
docker history myapp:latest
docker run --rm myapp:latest du -sh /*
```

**Common culprits**:
- Dev dependencies in production (`npm ci --omit=dev`)
- Build tools in final stage (use multi-stage)
- Package manager caches (`npm cache clean --force`, `rm -rf /var/lib/apt/lists/*`)
- Unnecessary locales (add `--no-install-recommends`)

## Network Configuration

### Service Discovery in Docker Compose

Services communicate by service name:
```typescript
// In app container, connect to db container:
const DATABASE_URL = 'postgresql://postgres:postgres@db:5432/myapp';
//                                                   ^^
//                                          service name, not localhost
```

### Custom Networks

```yaml
services:
  app:
    networks:
      - frontend
      - backend

  db:
    networks:
      - backend  # Only accessible from backend network

  nginx:
    networks:
      - frontend

networks:
  frontend:
  backend:
    internal: true  # No external access
```

### Exposing Ports

```yaml
services:
  app:
    ports:
      - "3000:3000"           # Accessible from host
    expose:
      - "9229"                # Only accessible from other containers
  db:
    # No ports section = only accessible from other containers on same network
```

## Volume Patterns

### Named Volumes (persistent data)

```yaml
volumes:
  postgres_data:
    driver: local
  redis_data:
    driver: local
  uploads:
    driver: local
    driver_opts:
      type: none
      o: bind
      device: ./data/uploads
```

### Bind Mounts (development)

```yaml
services:
  app:
    volumes:
      - .:/app                # Source code (hot reload)
      - /app/node_modules     # Prevent overwriting container's node_modules
      - /app/.next            # Prevent overwriting build cache
```

### tmpfs (test databases, caches)

```yaml
services:
  test-db:
    tmpfs:
      - /var/lib/postgresql/data  # In-memory, fast, ephemeral
```
