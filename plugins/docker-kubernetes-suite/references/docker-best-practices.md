# Docker Best Practices Reference

Quick-reference guide for Docker best practices. Agents consult this automatically — you can also read it directly for quick answers.

---

## Image Optimization

### Size Reduction Strategies

**1. Use minimal base images:**

| Language | Full | Slim | Alpine | Distroless |
|----------|------|------|--------|------------|
| Node.js 22 | ~1.1GB | ~240MB | ~180MB | ~130MB |
| Python 3.13 | ~1.0GB | ~150MB | ~60MB | ~50MB |
| Go 1.23 | ~800MB | N/A | ~250MB | ~2MB (scratch) |
| Java 21 | ~500MB | N/A | ~200MB | ~220MB |

**2. Multi-stage builds — don't ship build tools:**
```dockerfile
FROM node:22 AS build
RUN npm ci && npm run build

FROM node:22-slim AS runtime
COPY --from=build /app/dist ./dist
COPY --from=build /app/node_modules ./node_modules
```

**3. Remove unnecessary files in same layer:**
```dockerfile
RUN apt-get update \
    && apt-get install -y --no-install-recommends build-essential \
    && npm ci \
    && apt-get purge -y build-essential \
    && apt-get autoremove -y \
    && rm -rf /var/lib/apt/lists/*
```

**4. Use --omit=dev for Node.js:**
```dockerfile
RUN npm ci --omit=dev --ignore-scripts
```

**5. Use pip --no-cache-dir for Python:**
```dockerfile
RUN pip install --no-cache-dir -r requirements.txt
```

**6. Strip Go binaries:**
```dockerfile
RUN go build -ldflags="-w -s" -o /app
# -w: omit DWARF debug info
# -s: omit symbol table
```

### Build Speed Optimization

**1. Layer ordering — least changing first:**
```dockerfile
# System deps (rarely change)
RUN apt-get update && apt-get install -y curl

# Package manager files (change when deps change)
COPY package.json package-lock.json ./
RUN npm ci

# Source code (changes most often)
COPY . .
RUN npm run build
```

**2. BuildKit cache mounts:**
```dockerfile
RUN --mount=type=cache,target=/root/.npm npm ci
RUN --mount=type=cache,target=/root/.cache/pip pip install -r requirements.txt
RUN --mount=type=cache,target=/go/pkg/mod go mod download
```

**3. Parallel multi-stage builds:**
```dockerfile
# These stages run in parallel automatically with BuildKit
FROM node:22-slim AS build-frontend
COPY frontend/ ./
RUN npm ci && npm run build

FROM golang:1.23 AS build-backend
COPY backend/ ./
RUN go build -o /server

FROM scratch AS runtime
COPY --from=build-backend /server /server
COPY --from=build-frontend /app/dist /static
```

**4. Remote caching:**
```bash
docker buildx build \
  --cache-from type=registry,ref=ghcr.io/org/myapp:cache \
  --cache-to type=registry,ref=ghcr.io/org/myapp:cache,mode=max \
  -t myapp .
```

---

## .dockerignore Best Practices

```dockerignore
# MUST ignore — security and performance
.git
.env
.env.*
!.env.example
*.pem
*.key
node_modules
__pycache__
.venv
vendor

# SHOULD ignore — not needed in build context
Dockerfile*
docker-compose*.yml
compose*.yaml
.dockerignore
.github
.gitlab-ci.yml
README.md
CHANGELOG.md
LICENSE
docs/
*.md

# Build artifacts — rebuilt in container
dist
build
out
target
coverage
.nyc_output

# IDE and OS
.vscode
.idea
*.swp
.DS_Store
Thumbs.db
```

**Check context size:**
```bash
# See what's being sent
docker build --progress=plain . 2>&1 | head -5
# "transferring context: 2.3MB" — good
# "transferring context: 500MB" — .dockerignore needs work
```

---

## Health Checks

### Dockerfile HEALTHCHECK

```dockerfile
# HTTP endpoint
HEALTHCHECK --interval=30s --timeout=5s --start-period=30s --retries=3 \
  CMD curl -f http://localhost:3000/health || exit 1

# TCP check (no curl needed)
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
  CMD node -e "require('net').createConnection(3000, 'localhost').on('error', () => process.exit(1))"

# Node.js fetch (no extra tools)
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
  CMD node -e "fetch('http://localhost:3000/health').then(r => {if(!r.ok) throw 1})"

# Alpine (wget, not curl)
HEALTHCHECK --interval=30s --timeout=3s CMD wget -qO- http://localhost:8080/health || exit 1

# PostgreSQL
HEALTHCHECK --interval=10s --timeout=5s --retries=5 CMD pg_isready -U postgres || exit 1

# Redis
HEALTHCHECK --interval=10s --timeout=3s --retries=5 CMD redis-cli ping | grep -q PONG
```

### Docker Compose Health Checks

```yaml
services:
  api:
    healthcheck:
      test: ["CMD-SHELL", "curl -f http://localhost:3000/health || exit 1"]
      interval: 30s
      timeout: 5s
      retries: 3
      start_period: 30s
    depends_on:
      db:
        condition: service_healthy
```

**Parameters:**
- `interval`: Time between checks (default: 30s)
- `timeout`: Max time for a check (default: 30s)
- `start_period`: Grace period during startup (default: 0s)
- `retries`: Consecutive failures before unhealthy (default: 3)

---

## Logging Best Practices

### Log to stdout/stderr

```dockerfile
# Application logs should go to stdout/stderr — Docker captures them
# Don't write log files inside containers

# Nginx
RUN ln -sf /dev/stdout /var/log/nginx/access.log \
    && ln -sf /dev/stderr /var/log/nginx/error.log

# Apache
RUN ln -sf /proc/self/fd/1 /var/log/apache2/access.log \
    && ln -sf /proc/self/fd/2 /var/log/apache2/error.log
```

### Logging Drivers

```yaml
# Compose — configure logging
services:
  api:
    logging:
      driver: json-file          # Default, good for most cases
      options:
        max-size: "10m"          # Rotate at 10MB
        max-file: "3"            # Keep 3 files
        compress: "true"
        tag: "{{.Name}}/{{.ID}}"

# Other drivers:
# - local (default in newer Docker, compressed, faster)
# - fluentd (send to Fluentd)
# - gelf (send to Graylog)
# - awslogs (CloudWatch)
# - gcplogs (Cloud Logging)
```

### Structured Logging

```javascript
// Always use structured (JSON) logging in containers
// Makes parsing by log aggregators trivial

// Good — structured
console.log(JSON.stringify({
  level: 'info',
  msg: 'Request processed',
  method: 'GET',
  path: '/api/users',
  status: 200,
  duration_ms: 45,
  request_id: 'abc-123'
}));

// Bad — unstructured
console.log('GET /api/users 200 45ms');
```

---

## Multi-Platform Builds

```bash
# Create a multi-platform builder
docker buildx create --name multiplatform --driver docker-container --use

# Build for multiple platforms
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t ghcr.io/org/myapp:1.0 \
  --push .

# Inspect manifest
docker buildx imagetools inspect ghcr.io/org/myapp:1.0
```

### Cross-Compilation in Dockerfile

```dockerfile
# Use BUILDPLATFORM for build stage, TARGETPLATFORM for runtime
FROM --platform=$BUILDPLATFORM golang:1.23 AS build
ARG TARGETOS TARGETARCH

COPY . .
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /server

FROM --platform=$TARGETPLATFORM gcr.io/distroless/static:nonroot
COPY --from=build /server /server
ENTRYPOINT ["/server"]
```

---

## Security Quick Reference

### Non-Root User

```dockerfile
# Debian/Ubuntu
RUN groupadd -r app && useradd -r -g app -d /app -s /sbin/nologin app
USER app

# Alpine
RUN addgroup -S app && adduser -S -G app -h /app app
USER app

# Node.js (built-in)
USER node

# Distroless
FROM gcr.io/distroless/nodejs22-debian12:nonroot

# Scratch (numeric UID)
USER 65534:65534
```

### Read-Only Filesystem

```yaml
# Docker run
docker run --read-only --tmpfs /tmp --tmpfs /var/run myapp

# Compose
services:
  api:
    read_only: true
    tmpfs:
      - /tmp
      - /var/run
```

### Drop Capabilities

```yaml
services:
  api:
    cap_drop:
      - ALL
    cap_add:
      - NET_BIND_SERVICE    # Only if needed
    security_opt:
      - no-new-privileges:true
```

### Secret Mounts

```dockerfile
# Build-time secrets — never in image layers
RUN --mount=type=secret,id=npm_token \
    NPM_TOKEN=$(cat /run/secrets/npm_token) npm ci

# SSH forwarding
RUN --mount=type=ssh git clone git@github.com:org/private.git
```

---

## Compose Development Patterns

### Watch Mode (Hot Reload)

```yaml
services:
  api:
    develop:
      watch:
        # Sync source files — instant hot-reload
        - action: sync
          path: ./src
          target: /app/src

        # Rebuild on dependency changes
        - action: rebuild
          path: ./package.json

        # Sync + restart on config changes
        - action: sync+restart
          path: ./config
          target: /app/config

# Start with: docker compose up --watch
```

### Development vs Production

```yaml
# compose.yaml — base (production)
services:
  api:
    image: ghcr.io/org/myapp:latest
    restart: unless-stopped

# compose.override.yaml — auto-applied in dev
services:
  api:
    build: .
    restart: "no"
    volumes:
      - ./src:/app/src
    ports:
      - "9229:9229"  # Debug port
    environment:
      NODE_ENV: development

# Dev: docker compose up (auto-includes override)
# Prod: docker compose -f compose.yaml up (explicit, no override)
```

### Database Initialization

```yaml
services:
  db:
    image: postgres:17-alpine
    volumes:
      # Init scripts run on first start only (alphabetical order)
      - ./db/01-schema.sql:/docker-entrypoint-initdb.d/01-schema.sql:ro
      - ./db/02-seed.sql:/docker-entrypoint-initdb.d/02-seed.sql:ro
      # Persistent data
      - postgres_data:/var/lib/postgresql/data
```

---

## Entrypoint Patterns

### Exec Form vs Shell Form

```dockerfile
# GOOD — exec form, PID 1, receives SIGTERM
CMD ["node", "server.js"]
ENTRYPOINT ["python", "-m", "uvicorn"]

# BAD — shell form, /bin/sh is PID 1
CMD node server.js
ENTRYPOINT python -m uvicorn
```

### Entrypoint Script

```dockerfile
COPY <<-"EOF" /entrypoint.sh
#!/bin/sh
set -euo pipefail

# Wait for dependencies
echo "Checking database connection..."
until pg_isready -h "$DB_HOST" -p 5432 -U "$DB_USER" -q; do
  echo "Database not ready, waiting..."
  sleep 2
done

# Run migrations if needed
if [ "${RUN_MIGRATIONS:-false}" = "true" ]; then
  echo "Running migrations..."
  npx prisma migrate deploy
fi

# Use exec to replace shell — app gets PID 1
echo "Starting application..."
exec "$@"
EOF

RUN chmod +x /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
CMD ["node", "dist/server.js"]
```

### ENTRYPOINT + CMD Pattern

```dockerfile
# ENTRYPOINT = always runs (the executable)
# CMD = default arguments (can be overridden)

ENTRYPOINT ["python", "-m"]
CMD ["uvicorn", "app:app", "--host", "0.0.0.0"]

# docker run myapp                         → python -m uvicorn app:app --host 0.0.0.0
# docker run myapp pytest                  → python -m pytest
# docker run myapp flask db upgrade        → python -m flask db upgrade
```

---

## Docker Network Patterns

### Service Discovery

```yaml
# Services resolve each other by name
services:
  api:
    environment:
      DATABASE_URL: postgres://postgres:pass@db:5432/myapp     # "db" resolves
      REDIS_URL: redis://cache:6379                            # "cache" resolves
  db:
    image: postgres:17
  cache:
    image: redis:7
```

### Network Isolation

```yaml
services:
  frontend:
    networks: [public, internal]
  api:
    networks: [internal, database]
  db:
    networks: [database]       # Only api can reach db

networks:
  public:
  internal:
    internal: true             # No external access
  database:
    internal: true
```

---

## Volume Patterns

### Named Volumes

```yaml
services:
  db:
    volumes:
      - postgres_data:/var/lib/postgresql/data     # Named volume

volumes:
  postgres_data:
    driver: local
```

### Backup and Restore

```bash
# Backup
docker run --rm -v postgres_data:/data -v $(pwd):/backup \
  alpine tar czf /backup/db_backup.tar.gz -C /data .

# Restore
docker run --rm -v postgres_data:/data -v $(pwd):/backup \
  alpine sh -c "cd /data && tar xzf /backup/db_backup.tar.gz"
```

---

## Docker Cleanup

```bash
# Remove all stopped containers
docker container prune

# Remove unused images
docker image prune -a

# Remove unused volumes (CAREFUL: deletes data!)
docker volume prune

# Remove everything unused
docker system prune -a --volumes

# Check disk usage
docker system df
docker system df -v

# Remove specific old images
docker images --filter "dangling=true" -q | xargs docker rmi
docker images --filter "before=myapp:1.0" -q | xargs docker rmi
```

---

## Debugging Quick Reference

```bash
# Build
docker build --no-cache --progress=plain .           # Full output, no cache
docker build --target build -t debug .               # Stop at stage
docker history myapp:latest                          # Show layers

# Runtime
docker exec -it container /bin/sh                    # Shell into container
docker logs -f --tail=100 container                  # Follow logs
docker logs --since 5m container                     # Recent logs
docker stats container                               # Resource usage
docker inspect container                             # Full metadata
docker diff container                                # Filesystem changes
docker top container                                 # Process list

# Network
docker network inspect bridge                        # Network details
docker exec container cat /etc/resolv.conf           # DNS config
docker exec container wget -qO- http://service:3000  # Test connectivity

# Compose
docker compose config                                # Validate and render
docker compose logs -f api                           # Follow service logs
docker compose exec api /bin/sh                      # Shell into service
docker compose top                                   # All processes
docker compose ps -a                                 # All containers including stopped
```
