---
name: docker-compose
description: >
  Docker Compose for local development and production — service definitions,
  networking, volumes, environment variables, health checks, and multi-environment setups.
  Triggers: "docker compose", "docker-compose", "compose file", "local development docker",
  "multi-container", "compose services".
  NOT for: Dockerfiles (use dockerfile-patterns), Kubernetes (use kubernetes-deployments).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Docker Compose

## Full-Stack Development Setup

```yaml
# docker-compose.yml
services:
  app:
    build:
      context: .
      dockerfile: Dockerfile
      target: development  # Use dev stage of multi-stage build
    ports:
      - "3000:3000"
    volumes:
      - .:/app
      - /app/node_modules  # Anonymous volume — don't mount host node_modules
    environment:
      - NODE_ENV=development
      - DATABASE_URL=postgresql://postgres:postgres@db:5432/myapp
      - REDIS_URL=redis://redis:6379
    depends_on:
      db:
        condition: service_healthy
      redis:
        condition: service_healthy
    command: npm run dev

  db:
    image: postgres:16-alpine
    ports:
      - "5432:5432"
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: myapp
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./init.sql:/docker-entrypoint-initdb.d/init.sql
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 5s
      timeout: 3s
      retries: 5

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 3s
      retries: 5
    command: redis-server --appendonly yes

  worker:
    build:
      context: .
      target: development
    volumes:
      - .:/app
      - /app/node_modules
    environment:
      - NODE_ENV=development
      - DATABASE_URL=postgresql://postgres:postgres@db:5432/myapp
      - REDIS_URL=redis://redis:6379
    depends_on:
      db:
        condition: service_healthy
      redis:
        condition: service_healthy
    command: npm run worker

volumes:
  postgres_data:
  redis_data:
```

## Environment-Specific Overrides

### Base + Override Pattern

```yaml
# docker-compose.yml (base — shared config)
services:
  app:
    build:
      context: .
    environment:
      - DATABASE_URL=postgresql://postgres:postgres@db:5432/myapp

  db:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: myapp

volumes:
  postgres_data:
```

```yaml
# docker-compose.override.yml (auto-loaded for development)
services:
  app:
    build:
      target: development
    ports:
      - "3000:3000"
    volumes:
      - .:/app
      - /app/node_modules
    command: npm run dev
    environment:
      - NODE_ENV=development

  db:
    ports:
      - "5432:5432"
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
```

```yaml
# docker-compose.prod.yml (production override)
services:
  app:
    build:
      target: production
    restart: always
    deploy:
      replicas: 2
      resources:
        limits:
          memory: 512M
          cpus: "0.5"
    environment:
      - NODE_ENV=production

  db:
    restart: always
    environment:
      POSTGRES_USER: ${DB_USER}
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
```

```bash
# Run with override
docker compose up                                    # Uses base + override
docker compose -f docker-compose.yml -f docker-compose.prod.yml up  # Uses base + prod
```

## .env File

```bash
# .env (loaded automatically by docker compose)
COMPOSE_PROJECT_NAME=myapp
POSTGRES_USER=postgres
POSTGRES_PASSWORD=localdevpassword
REDIS_PASSWORD=localdevpassword
APP_PORT=3000
```

```yaml
# Reference in docker-compose.yml
services:
  app:
    ports:
      - "${APP_PORT:-3000}:3000"  # Default value if not set
  db:
    environment:
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
```

## Networking

### Default Network (Services Communicate by Name)

```yaml
services:
  app:
    # Reaches db at hostname "db", redis at "redis"
    environment:
      - DATABASE_URL=postgresql://postgres:postgres@db:5432/myapp
      - REDIS_URL=redis://redis:6379

  db:
    # Accessible from app as "db:5432"
    image: postgres:16-alpine
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
      - frontend  # Public-facing

networks:
  frontend:
  backend:
    internal: true  # No external access
```

## Volume Patterns

```yaml
services:
  app:
    volumes:
      # Bind mount — source code (hot reload)
      - .:/app

      # Anonymous volume — preserve container's node_modules
      - /app/node_modules

      # Named volume — persistent data
      - uploads:/app/uploads

  db:
    volumes:
      # Named volume — database storage
      - postgres_data:/var/lib/postgresql/data

      # Bind mount — init scripts
      - ./sql/init.sql:/docker-entrypoint-initdb.d/init.sql:ro

volumes:
  postgres_data:
    driver: local
  uploads:
```

## Health Checks

```yaml
services:
  app:
    healthcheck:
      test: ["CMD", "wget", "-qO-", "http://localhost:3000/health"]
      interval: 30s
      timeout: 5s
      retries: 3
      start_period: 10s

  db:
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 5s
      timeout: 3s
      retries: 5

  redis:
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 3s
      retries: 5

  elasticsearch:
    healthcheck:
      test: ["CMD-SHELL", "curl -fs http://localhost:9200/_cluster/health || exit 1"]
      interval: 10s
      timeout: 5s
      retries: 10
      start_period: 30s
```

## Common Service Recipes

### Nginx Reverse Proxy

```yaml
  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf:ro
      - ./nginx/certs:/etc/nginx/certs:ro
    depends_on:
      app:
        condition: service_healthy
```

### Mailhog (Email Testing)

```yaml
  mailhog:
    image: mailhog/mailhog
    ports:
      - "1025:1025"  # SMTP
      - "8025:8025"  # Web UI
```

### MinIO (S3-Compatible Storage)

```yaml
  minio:
    image: minio/minio
    ports:
      - "9000:9000"
      - "9001:9001"
    environment:
      MINIO_ROOT_USER: minioadmin
      MINIO_ROOT_PASSWORD: minioadmin
    volumes:
      - minio_data:/data
    command: server /data --console-address ":9001"
    healthcheck:
      test: ["CMD", "mc", "ready", "local"]
      interval: 5s
      timeout: 3s
      retries: 5
```

### RabbitMQ

```yaml
  rabbitmq:
    image: rabbitmq:3-management-alpine
    ports:
      - "5672:5672"   # AMQP
      - "15672:15672" # Management UI
    environment:
      RABBITMQ_DEFAULT_USER: guest
      RABBITMQ_DEFAULT_PASS: guest
    volumes:
      - rabbitmq_data:/var/lib/rabbitmq
    healthcheck:
      test: ["CMD", "rabbitmq-diagnostics", "-q", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5
```

## Useful Commands

```bash
# Start all services
docker compose up -d

# Start specific service
docker compose up -d app db

# View logs
docker compose logs -f app
docker compose logs --tail=50 app

# Rebuild images
docker compose build
docker compose build --no-cache app

# Stop and remove
docker compose down
docker compose down -v  # Also remove volumes (DANGER: deletes data)

# Run one-off command
docker compose run --rm app npm test
docker compose run --rm app npx prisma migrate dev

# Execute in running container
docker compose exec app sh
docker compose exec db psql -U postgres myapp

# Scale service
docker compose up -d --scale worker=3

# Check status
docker compose ps
docker compose top
```

## Gotchas

1. **`depends_on` doesn't wait for readiness** — without `condition: service_healthy`, Compose only waits for the container to START, not for the service inside to be ready. Always use health checks with `condition: service_healthy`.

2. **Bind mounts override container files** — if you mount `.:/app`, the container's `/app/node_modules` is hidden by the host's (empty) directory. Use an anonymous volume: `- /app/node_modules`.

3. **`.env` file is for Compose variables, not container env** — `.env` sets variables for the `docker-compose.yml` file interpolation. To set container environment, use `environment:` or `env_file:`.

4. **Port conflicts** — if port 5432 is already in use on your host (local PostgreSQL), either stop the local service or remap: `"5433:5432"`.

5. **Volume data persists across `down`** — `docker compose down` does NOT remove volumes. Data in named volumes survives. Use `down -v` to clean everything (careful with databases).

6. **File permissions with bind mounts** — files created inside the container may be owned by root on the host. Use `user: "${UID}:${GID}"` in the service definition to match host user.

7. **Build cache invalidation** — changing any file before `COPY .` in the Dockerfile invalidates the cache from that point. Structure your Dockerfile to copy dependency files first.

8. **Compose V2 vs V1** — V2 is `docker compose` (space). V1 was `docker-compose` (hyphen). V1 is deprecated. If scripts use `docker-compose`, update to `docker compose`.
