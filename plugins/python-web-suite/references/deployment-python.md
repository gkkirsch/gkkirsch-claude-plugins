# Python Web Deployment Reference

Quick-reference guide for deploying Python web applications (Django, Flask, FastAPI) with Gunicorn, Uvicorn, Docker, nginx, and cloud platforms. Consult this when deploying or optimizing Python web services.

---

## Table of Contents

1. [Gunicorn Configuration](#gunicorn-configuration)
2. [Uvicorn Configuration](#uvicorn-configuration)
3. [Docker Configuration](#docker-configuration)
4. [Nginx Configuration](#nginx-configuration)
5. [Systemd Service Files](#systemd-service-files)
6. [Heroku Deployment](#heroku-deployment)
7. [AWS Deployment](#aws-deployment)
8. [Health Checks & Monitoring](#health-checks--monitoring)
9. [CI/CD Pipeline](#cicd-pipeline)

---

## Gunicorn Configuration

### gunicorn.conf.py — Full Production Config

```python
# gunicorn.conf.py
import multiprocessing
import os

# --- Binding ---
# Use a Unix socket for nginx reverse proxy (faster than TCP)
bind = "unix:/run/gunicorn/gunicorn.sock"
# Or bind to TCP for Docker/cloud environments:
# bind = "0.0.0.0:8000"

# --- Worker Processes ---
# Rule of thumb: (2 * CPU cores) + 1 for CPU-bound apps
# For I/O-bound apps (most Django/Flask), you can go higher
workers = (multiprocessing.cpu_count() * 2) + 1

# Worker class options:
#   "sync"    — default, good for CPU-bound, no long-running requests
#   "gthread" — threaded sync worker, good for mixed I/O
#   "gevent"  — async green threads, good for many concurrent connections
#   "uvicorn.workers.UvicornWorker" — for ASGI apps (FastAPI, Django async)
worker_class = "sync"

# Number of threads per worker (only for gthread worker class)
threads = 2

# Max simultaneous connections per worker (for gevent/eventlet)
worker_connections = 1000

# --- Timeouts ---
# Kill and restart a worker that hasn't responded in this many seconds
timeout = 30

# Timeout for graceful worker restart (SIGTERM → SIGKILL delay)
graceful_timeout = 30

# Keep-alive connections: seconds to wait for next request on keep-alive
keepalive = 5

# --- Request Limits ---
# Max requests per worker before automatic restart (prevents memory leaks)
max_requests = 1000

# Randomize max_requests per worker to avoid thundering herd on restarts
max_requests_jitter = 100

# Max size of HTTP request line (bytes) — protects against malformed requests
limit_request_line = 4096

# Max number of HTTP headers
limit_request_fields = 100

# Max size of each HTTP header field
limit_request_field_size = 8190

# --- Process Naming ---
proc_name = "myapp_gunicorn"

# --- Directories ---
# Chdir to app directory before loading
chdir = "/app"

# Path to store PID file
pidfile = "/run/gunicorn/gunicorn.pid"

# Worker temp directory — use /dev/shm for RAM-backed temp files
worker_tmp_dir = "/dev/shm"

# --- Logging ---
# Log to stdout/stderr for Docker/systemd (use "-" for stdout)
accesslog = "-"
errorlog = "-"
loglevel = "info"

# Custom access log format (includes response time in microseconds)
access_log_format = (
    '%(h)s %(l)s %(u)s %(t)s "%(r)s" %(s)s %(b)s '
    '"%(f)s" "%(a)s" %(D)s'
)

# Disable access log for health check paths (reduce noise)
# Use a custom log filter instead — see post_request hook below

# --- SSL (if terminating at Gunicorn, not recommended — use nginx) ---
# keyfile = "/etc/ssl/private/server.key"
# certfile = "/etc/ssl/certs/server.crt"

# --- Server Hooks ---
def on_starting(server):
    """Called just before the master process is initialized."""
    pass

def on_reload(server):
    """Called before reloading workers via SIGHUP."""
    pass

def when_ready(server):
    """Called just after the server is started."""
    server.log.info("Gunicorn server ready. Workers: %s", server.cfg.workers)

def pre_fork(server, worker):
    """Called just before a worker is forked."""
    pass

def post_fork(server, worker):
    """Called just after a worker has been forked.
    Use this to reinitialize connections that should not be shared
    between parent and child (e.g., database connection pools).
    """
    # Close any inherited DB connections — Django will reopen them
    from django.db import connections
    for conn in connections.all():
        conn.close()

def pre_exec(server):
    """Called just before a new master process is forked on reload."""
    server.log.info("Forking new master")

def worker_int(worker):
    """Called when a worker receives SIGINT or SIGQUIT."""
    worker.log.info("Worker interrupted (PID %s)", worker.pid)

def worker_abort(worker):
    """Called when a worker receives SIGABRT (e.g., timeout)."""
    worker.log.error("Worker aborted (PID %s) — likely timeout", worker.pid)

def post_request(worker, req, environ, resp):
    """Called after a worker processes a request."""
    # Example: skip logging health check endpoints
    if environ.get("PATH_INFO") in ("/health/", "/readyz/"):
        worker.log.loglevel = 100  # suppress log for this request
```

### Graceful Restart Commands

```bash
# Reload workers gracefully (zero-downtime) — sends SIGHUP to master
kill -HUP $(cat /run/gunicorn/gunicorn.pid)

# Or with systemd
sudo systemctl reload gunicorn

# Force restart (kills workers immediately)
sudo systemctl restart gunicorn

# Increase worker count at runtime (USR1 = reopen logs, USR2 = upgrade)
kill -TTIN $(cat /run/gunicorn/gunicorn.pid)   # add one worker
kill -TTOU $(cat /run/gunicorn/gunicorn.pid)   # remove one worker
```

### Worker Count Calculation Reference

```
# CPU-bound (image processing, crypto, etc.)
workers = CPU_COUNT + 1

# I/O-bound (typical web apps with DB queries)
workers = (CPU_COUNT * 2) + 1

# High-concurrency I/O (async or gevent)
workers = CPU_COUNT
worker_connections = 1000  # per worker → total = CPU_COUNT * 1000
```

---

## Uvicorn Configuration

### Running Uvicorn with Gunicorn as Process Manager (Recommended for Production)

```bash
# FastAPI app at app/main.py exposing `app`
gunicorn app.main:app \
  --workers 4 \
  --worker-class uvicorn.workers.UvicornWorker \
  --bind 0.0.0.0:8000 \
  --timeout 60 \
  --keepalive 5 \
  --max-requests 1000 \
  --max-requests-jitter 100 \
  --log-level info \
  --access-logfile - \
  --error-logfile -
```

### gunicorn.conf.py for ASGI Apps

```python
# gunicorn.conf.py — ASGI (FastAPI / Starlette / Django async)
import multiprocessing

bind = "0.0.0.0:8000"
workers = multiprocessing.cpu_count() * 2  # ASGI is async — fewer workers needed
worker_class = "uvicorn.workers.UvicornWorker"
worker_connections = 1000
timeout = 120          # longer for async long-polling endpoints
keepalive = 5
max_requests = 1000
max_requests_jitter = 100
accesslog = "-"
errorlog = "-"
loglevel = "info"
```

### Programmatic uvicorn.run() Config (Development / Simple Deployments)

```python
# run.py — programmatic Uvicorn startup
import uvicorn

if __name__ == "__main__":
    uvicorn.run(
        "app.main:app",         # import string (required for reload)
        host="0.0.0.0",
        port=8000,
        workers=1,              # use Gunicorn for multi-worker in production
        loop="uvloop",          # faster event loop (install uvloop separately)
        http="httptools",       # faster HTTP parser (install httptools separately)
        log_level="info",
        access_log=True,
        use_colors=False,       # disable ANSI colors in production logs
        # SSL:
        # ssl_keyfile="/etc/ssl/private/key.pem",
        # ssl_certfile="/etc/ssl/certs/cert.pem",
        # Development only:
        reload=False,
        reload_dirs=["app"],
    )
```

### SSL/TLS Setup with Uvicorn Directly

```bash
# Standalone Uvicorn with SSL (not recommended for production — use nginx)
uvicorn app.main:app \
  --host 0.0.0.0 \
  --port 443 \
  --ssl-keyfile /etc/letsencrypt/live/example.com/privkey.pem \
  --ssl-certfile /etc/letsencrypt/live/example.com/fullchain.pem \
  --ssl-ca-certs /etc/letsencrypt/live/example.com/chain.pem
```

### Hot Reload for Development

```bash
# Development with auto-reload on file changes
uvicorn app.main:app \
  --host 127.0.0.1 \
  --port 8000 \
  --reload \
  --reload-dir app \
  --log-level debug

# With watchfiles (faster file watcher, install watchfiles package)
uvicorn app.main:app --reload --reload-delay 0.1
```

### Access Log Format for Uvicorn

```python
# Custom log config for structured JSON logging
import logging
import uvicorn

LOG_CONFIG = {
    "version": 1,
    "disable_existing_loggers": False,
    "formatters": {
        "json": {
            "()": "pythonjsonlogger.jsonlogger.JsonFormatter",
            "format": "%(asctime)s %(name)s %(levelname)s %(message)s",
        }
    },
    "handlers": {
        "default": {
            "class": "logging.StreamHandler",
            "formatter": "json",
            "stream": "ext://sys.stdout",
        }
    },
    "loggers": {
        "uvicorn": {"handlers": ["default"], "level": "INFO"},
        "uvicorn.access": {"handlers": ["default"], "level": "INFO"},
        "uvicorn.error": {"handlers": ["default"], "level": "INFO"},
    },
}

uvicorn.run("app.main:app", log_config=LOG_CONFIG)
```

---

## Docker Configuration

### Multi-Stage Dockerfile for Django

```dockerfile
# Dockerfile — Django production image
# Stage 1: Build dependencies
FROM python:3.12-slim AS builder

WORKDIR /build

# Install build dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    libpq-dev \
    && rm -rf /var/lib/apt/lists/*

# Copy and install Python dependencies into a prefix directory
COPY requirements.txt .
RUN pip install --upgrade pip && \
    pip install --prefix=/install --no-cache-dir -r requirements.txt

# Stage 2: Production image
FROM python:3.12-slim AS production

WORKDIR /app

# Install runtime dependencies only
RUN apt-get update && apt-get install -y --no-install-recommends \
    libpq5 \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Copy installed packages from builder
COPY --from=builder /install /usr/local

# Create non-root user
RUN groupadd --gid 1001 appgroup && \
    useradd --uid 1001 --gid appgroup --shell /bin/bash --create-home appuser

# Copy application source
COPY --chown=appuser:appgroup . .

# Create directories for runtime files
RUN mkdir -p /app/staticfiles /app/mediafiles /run/gunicorn && \
    chown -R appuser:appgroup /app /run/gunicorn

USER appuser

# Collect static files at build time
RUN python manage.py collectstatic --noinput

EXPOSE 8000

# Health check — hits the /health/ endpoint
HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
    CMD curl -f http://localhost:8000/health/ || exit 1

CMD ["gunicorn", "myproject.wsgi:application", \
     "--config", "gunicorn.conf.py"]
```

### Multi-Stage Dockerfile for FastAPI

```dockerfile
# Dockerfile — FastAPI production image
FROM python:3.12-slim AS builder

WORKDIR /build

RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    && rm -rf /var/lib/apt/lists/*

COPY requirements.txt .
RUN pip install --upgrade pip && \
    pip install --prefix=/install --no-cache-dir -r requirements.txt

FROM python:3.12-slim AS production

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends \
    curl \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /install /usr/local

RUN groupadd --gid 1001 appgroup && \
    useradd --uid 1001 --gid appgroup --shell /bin/bash --create-home appuser

COPY --chown=appuser:appgroup . .

RUN mkdir -p /run/gunicorn && chown appuser:appgroup /run/gunicorn

USER appuser

EXPOSE 8000

HEALTHCHECK --interval=30s --timeout=10s --start-period=20s --retries=3 \
    CMD curl -f http://localhost:8000/health || exit 1

CMD ["gunicorn", "app.main:app", \
     "--worker-class", "uvicorn.workers.UvicornWorker", \
     "--config", "gunicorn.conf.py"]
```

### Docker Compose — Full Stack (Django + Postgres + Redis + Celery)

```yaml
# docker-compose.yml
version: "3.9"

services:
  db:
    image: postgres:16-alpine
    restart: unless-stopped
    environment:
      POSTGRES_DB: ${POSTGRES_DB:-myapp}
      POSTGRES_USER: ${POSTGRES_USER:-myapp}
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${POSTGRES_USER:-myapp}"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    restart: unless-stopped
    command: redis-server --appendonly yes --requirepass ${REDIS_PASSWORD}
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "-a", "${REDIS_PASSWORD}", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

  web:
    build:
      context: .
      target: production
    restart: unless-stopped
    environment:
      DATABASE_URL: postgres://${POSTGRES_USER:-myapp}:${POSTGRES_PASSWORD}@db:5432/${POSTGRES_DB:-myapp}
      REDIS_URL: redis://:${REDIS_PASSWORD}@redis:6379/0
      DJANGO_SETTINGS_MODULE: myproject.settings.production
      SECRET_KEY: ${SECRET_KEY}
      ALLOWED_HOSTS: ${ALLOWED_HOSTS}
    volumes:
      - staticfiles:/app/staticfiles
      - mediafiles:/app/mediafiles
    depends_on:
      db:
        condition: service_healthy
      redis:
        condition: service_healthy
    ports:
      - "8000:8000"   # expose for development; use nginx in production

  celery_worker:
    build:
      context: .
      target: production
    restart: unless-stopped
    command: celery -A myproject worker --loglevel=info --concurrency=4
    environment:
      DATABASE_URL: postgres://${POSTGRES_USER:-myapp}:${POSTGRES_PASSWORD}@db:5432/${POSTGRES_DB:-myapp}
      REDIS_URL: redis://:${REDIS_PASSWORD}@redis:6379/0
      DJANGO_SETTINGS_MODULE: myproject.settings.production
      SECRET_KEY: ${SECRET_KEY}
    depends_on:
      db:
        condition: service_healthy
      redis:
        condition: service_healthy

  celery_beat:
    build:
      context: .
      target: production
    restart: unless-stopped
    command: celery -A myproject beat --loglevel=info --scheduler django_celery_beat.schedulers:DatabaseScheduler
    environment:
      DATABASE_URL: postgres://${POSTGRES_USER:-myapp}:${POSTGRES_PASSWORD}@db:5432/${POSTGRES_DB:-myapp}
      REDIS_URL: redis://:${REDIS_PASSWORD}@redis:6379/0
      DJANGO_SETTINGS_MODULE: myproject.settings.production
      SECRET_KEY: ${SECRET_KEY}
    depends_on:
      db:
        condition: service_healthy
      redis:
        condition: service_healthy

  nginx:
    image: nginx:1.25-alpine
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf:ro
      - ./nginx/conf.d:/etc/nginx/conf.d:ro
      - staticfiles:/srv/static:ro
      - mediafiles:/srv/media:ro
      - /etc/letsencrypt:/etc/letsencrypt:ro
    depends_on:
      - web

volumes:
  postgres_data:
  redis_data:
  staticfiles:
  mediafiles:
```

### .dockerignore

```
# .dockerignore
.git
.gitignore
.github
.env
.env.*
!.env.example

# Python
__pycache__
*.pyc
*.pyo
*.pyd
.Python
*.egg-info
dist/
build/
.eggs/

# Virtual environments
.venv
venv
env

# Testing
.pytest_cache
.coverage
htmlcov/
.tox/

# Editors
.vscode
.idea
*.swp
*.swo

# OS
.DS_Store
Thumbs.db

# Docs
docs/
*.md
!README.md

# Node (if frontend assets exist)
node_modules/
npm-debug.log

# Logs
*.log
logs/

# Media and collected static (generated at runtime)
mediafiles/
staticfiles/
```

---

## Nginx Configuration

### Reverse Proxy with SSL Termination

```nginx
# /etc/nginx/conf.d/myapp.conf

# Rate limiting zone — 10 requests/second per IP, burst of 20
limit_req_zone $binary_remote_addr zone=api_limit:10m rate=10r/s;

# Upstream definition — Gunicorn via Unix socket
upstream gunicorn {
    server unix:/run/gunicorn/gunicorn.sock fail_timeout=0;
    # For TCP:
    # server 127.0.0.1:8000 fail_timeout=0;
    # Load balancing (multiple app servers):
    # server app1:8000 weight=3;
    # server app2:8000 weight=3;
    # server app3:8000 backup;
    keepalive 32;   # keep connections open to upstream
}

# Redirect HTTP → HTTPS
server {
    listen 80;
    listen [::]:80;
    server_name example.com www.example.com;
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name example.com www.example.com;

    # --- SSL/TLS ---
    ssl_certificate /etc/letsencrypt/live/example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/example.com/privkey.pem;
    ssl_trusted_certificate /etc/letsencrypt/live/example.com/chain.pem;

    # Modern TLS configuration
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers off;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 1d;
    ssl_session_tickets off;

    # OCSP stapling
    ssl_stapling on;
    ssl_stapling_verify on;
    resolver 1.1.1.1 1.0.0.1 valid=300s;
    resolver_timeout 5s;

    # --- Security Headers ---
    add_header Strict-Transport-Security "max-age=63072000; includeSubDomains; preload" always;
    add_header X-Frame-Options DENY always;
    add_header X-Content-Type-Options nosniff always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;
    add_header Permissions-Policy "geolocation=(), microphone=(), camera=()" always;

    # --- Gzip Compression ---
    gzip on;
    gzip_vary on;
    gzip_proxied any;
    gzip_comp_level 6;
    gzip_min_length 1024;
    gzip_types
        text/plain
        text/css
        text/javascript
        application/javascript
        application/json
        application/xml
        image/svg+xml
        font/woff2;

    # --- Client Settings ---
    client_max_body_size 20M;         # max upload size
    client_body_timeout 60s;
    client_header_timeout 60s;
    send_timeout 60s;

    # --- Static Files (Django) ---
    location /static/ {
        alias /srv/static/;
        expires 1y;
        add_header Cache-Control "public, immutable";
        access_log off;
    }

    # --- Media Files (Django) ---
    location /media/ {
        alias /srv/media/;
        expires 30d;
        add_header Cache-Control "public";
        access_log off;
    }

    # --- Rate Limit API endpoints ---
    location /api/ {
        limit_req zone=api_limit burst=20 nodelay;
        limit_req_status 429;
        include proxy_params;
        proxy_pass http://gunicorn;
    }

    # --- WebSocket Support ---
    location /ws/ {
        proxy_pass http://gunicorn;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_read_timeout 86400s;  # keep WebSocket connections open
        proxy_send_timeout 86400s;
    }

    # --- Main Application ---
    location / {
        include proxy_params;
        proxy_pass http://gunicorn;
    }
}
```

### /etc/nginx/proxy_params

```nginx
# /etc/nginx/proxy_params
proxy_http_version 1.1;
proxy_set_header Host $http_host;
proxy_set_header X-Real-IP $remote_addr;
proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
proxy_set_header X-Forwarded-Proto $scheme;
proxy_set_header X-Forwarded-Host $server_name;
proxy_redirect off;

# Timeouts
proxy_connect_timeout 10s;
proxy_send_timeout 60s;
proxy_read_timeout 60s;

# Buffering — disable for streaming responses
proxy_buffering on;
proxy_buffer_size 8k;
proxy_buffers 8 8k;
```

---

## Systemd Service Files

### Gunicorn Socket + Service (Socket Activation)

```ini
# /etc/systemd/system/gunicorn.socket
[Unit]
Description=Gunicorn WSGI HTTP Socket
PartOf=gunicorn.service

[Socket]
ListenStream=/run/gunicorn/gunicorn.sock
SocketUser=www-data
SocketMode=0660

[Install]
WantedBy=sockets.target
```

```ini
# /etc/systemd/system/gunicorn.service
[Unit]
Description=Gunicorn WSGI HTTP Server
After=network.target
Requires=gunicorn.socket

[Service]
Type=notify
User=appuser
Group=appgroup
RuntimeDirectory=gunicorn
WorkingDirectory=/app
ExecStart=/app/.venv/bin/gunicorn myproject.wsgi:application \
          --config /app/gunicorn.conf.py
ExecReload=/bin/kill -s HUP $MAINPID
KillMode=mixed
TimeoutStopSec=5
PrivateTmp=true
Restart=on-failure
RestartSec=5s

# Environment
EnvironmentFile=/etc/myapp/environment
Environment="DJANGO_SETTINGS_MODULE=myproject.settings.production"

# Resource limits
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
```

### Uvicorn Systemd Service

```ini
# /etc/systemd/system/uvicorn.service
[Unit]
Description=Uvicorn ASGI HTTP Server
After=network.target

[Service]
Type=exec
User=appuser
Group=appgroup
WorkingDirectory=/app
ExecStart=/app/.venv/bin/gunicorn app.main:app \
          --worker-class uvicorn.workers.UvicornWorker \
          --workers 4 \
          --bind unix:/run/uvicorn/uvicorn.sock \
          --timeout 120 \
          --log-level info \
          --access-logfile - \
          --error-logfile -
ExecReload=/bin/kill -s HUP $MAINPID
KillMode=mixed
TimeoutStopSec=5
PrivateTmp=true
RuntimeDirectory=uvicorn
Restart=on-failure
RestartSec=5s

EnvironmentFile=/etc/myapp/environment
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
```

### Celery Worker Systemd Service

```ini
# /etc/systemd/system/celery.service
[Unit]
Description=Celery Worker
After=network.target redis.service

[Service]
Type=forking
User=appuser
Group=appgroup
WorkingDirectory=/app
PIDFile=/run/celery/worker.pid
EnvironmentFile=/etc/myapp/environment

ExecStart=/bin/bash -c '/app/.venv/bin/celery multi start worker \
    -A myproject \
    --pidfile=/run/celery/%n.pid \
    --logfile=/var/log/celery/%n%I.log \
    --loglevel=INFO \
    --concurrency=4 \
    --queues=default,priority'

ExecStop=/bin/bash -c '/app/.venv/bin/celery multi stopwait worker \
    --pidfile=/run/celery/%n.pid'

ExecReload=/bin/bash -c '/app/.venv/bin/celery multi restart worker \
    -A myproject \
    --pidfile=/run/celery/%n.pid \
    --logfile=/var/log/celery/%n%I.log \
    --loglevel=INFO'

RuntimeDirectory=celery
Restart=on-failure
RestartSec=10s

[Install]
WantedBy=multi-user.target
```

### Celery Beat Systemd Service

```ini
# /etc/systemd/system/celerybeat.service
[Unit]
Description=Celery Beat Scheduler
After=network.target redis.service celery.service

[Service]
Type=simple
User=appuser
Group=appgroup
WorkingDirectory=/app
EnvironmentFile=/etc/myapp/environment

ExecStart=/app/.venv/bin/celery -A myproject beat \
    --loglevel=INFO \
    --scheduler django_celery_beat.schedulers:DatabaseScheduler \
    --pidfile=/run/celery/beat.pid

RuntimeDirectory=celery
Restart=on-failure
RestartSec=10s

[Install]
WantedBy=multi-user.target
```

### Systemd Management Commands

```bash
# Enable and start all services
sudo systemctl enable --now gunicorn.socket gunicorn celery celerybeat

# Check status
sudo systemctl status gunicorn
sudo journalctl -u gunicorn -f   # follow logs

# Reload without downtime
sudo systemctl reload gunicorn

# Reload systemd after editing unit files
sudo systemctl daemon-reload
```

---

## Heroku Deployment

### Procfile

```procfile
# Procfile — Heroku process declarations

# Django (WSGI)
web: gunicorn myproject.wsgi --workers 3 --timeout 30 --log-file -

# FastAPI (ASGI via Gunicorn + Uvicorn)
# web: gunicorn app.main:app --workers 2 --worker-class uvicorn.workers.UvicornWorker --bind 0.0.0.0:$PORT

# Flask
# web: gunicorn app:app --workers 3 --timeout 30 --bind 0.0.0.0:$PORT

# Celery worker
worker: celery -A myproject worker --loglevel=info --concurrency=2

# Celery beat (only run ONE dyno of this)
beat: celery -A myproject beat --loglevel=info --scheduler django_celery_beat.schedulers:DatabaseScheduler

# Release phase: run migrations before new web dynos start
release: python manage.py migrate --noinput
```

### runtime.txt

```
python-3.12.3
```

### app.json (Heroku App Manifest)

```json
{
  "name": "My Django App",
  "description": "Production Django application",
  "keywords": ["python", "django"],
  "env": {
    "SECRET_KEY": {
      "description": "Django secret key",
      "generator": "secret"
    },
    "DJANGO_SETTINGS_MODULE": {
      "description": "Django settings module",
      "value": "myproject.settings.production"
    },
    "ALLOWED_HOSTS": {
      "description": "Comma-separated list of allowed hosts"
    }
  },
  "formation": {
    "web": { "quantity": 1, "size": "standard-1x" },
    "worker": { "quantity": 1, "size": "standard-1x" }
  },
  "addons": [
    { "plan": "heroku-postgresql:essential-0" },
    { "plan": "heroku-redis:mini" }
  ],
  "buildpacks": [
    { "url": "heroku/python" }
  ]
}
```

### Heroku CLI Setup Commands

```bash
# Create app and add addons
heroku create myapp-production
heroku addons:create heroku-postgresql:essential-0
heroku addons:create heroku-redis:mini

# Set config vars
heroku config:set DJANGO_SETTINGS_MODULE=myproject.settings.production
heroku config:set SECRET_KEY="$(python -c 'from django.core.management.utils import get_random_secret_key; print(get_random_secret_key())')"
heroku config:set ALLOWED_HOSTS="myapp-production.herokuapp.com"

# Deploy
git push heroku main

# Run one-off commands
heroku run python manage.py createsuperuser
heroku run python manage.py shell

# Scale dynos
heroku ps:scale web=2 worker=1

# View logs
heroku logs --tail --dyno web
```

### Django Settings for Heroku (using dj-database-url)

```python
# settings/production.py
import dj_database_url
import os

# Database from DATABASE_URL env var (Heroku sets this automatically)
DATABASES = {
    "default": dj_database_url.config(
        conn_max_age=600,
        conn_health_checks=True,
        ssl_require=True,
    )
}

# Django-storages for S3 (Heroku has ephemeral filesystem)
DEFAULT_FILE_STORAGE = "storages.backends.s3boto3.S3Boto3Storage"
STATICFILES_STORAGE = "storages.backends.s3boto3.S3StaticStorage"
AWS_STORAGE_BUCKET_NAME = os.environ["AWS_STORAGE_BUCKET_NAME"]
AWS_S3_REGION_NAME = os.environ.get("AWS_S3_REGION_NAME", "us-east-1")
AWS_S3_FILE_OVERWRITE = False

# Whitenoise for static files (alternative to S3 for static only)
# STATICFILES_STORAGE = "whitenoise.storage.CompressedManifestStaticFilesStorage"

SECURE_SSL_REDIRECT = True
SECURE_PROXY_SSL_HEADER = ("HTTP_X_FORWARDED_PROTO", "https")
```

---

## AWS Deployment

### ECS Fargate Task Definition

```json
{
  "family": "myapp-web",
  "networkMode": "awsvpc",
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "512",
  "memory": "1024",
  "executionRoleArn": "arn:aws:iam::ACCOUNT_ID:role/ecsTaskExecutionRole",
  "taskRoleArn": "arn:aws:iam::ACCOUNT_ID:role/ecsTaskRole",
  "containerDefinitions": [
    {
      "name": "web",
      "image": "ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com/myapp:latest",
      "portMappings": [
        { "containerPort": 8000, "protocol": "tcp" }
      ],
      "environment": [
        { "name": "DJANGO_SETTINGS_MODULE", "value": "myproject.settings.production" }
      ],
      "secrets": [
        { "name": "SECRET_KEY", "valueFrom": "arn:aws:ssm:us-east-1:ACCOUNT_ID:parameter/myapp/SECRET_KEY" },
        { "name": "DATABASE_URL", "valueFrom": "arn:aws:ssm:us-east-1:ACCOUNT_ID:parameter/myapp/DATABASE_URL" },
        { "name": "REDIS_URL", "valueFrom": "arn:aws:ssm:us-east-1:ACCOUNT_ID:parameter/myapp/REDIS_URL" }
      ],
      "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "/ecs/myapp",
          "awslogs-region": "us-east-1",
          "awslogs-stream-prefix": "web"
        }
      },
      "healthCheck": {
        "command": ["CMD-SHELL", "curl -f http://localhost:8000/health/ || exit 1"],
        "interval": 30,
        "timeout": 10,
        "retries": 3,
        "startPeriod": 60
      },
      "essential": true
    }
  ]
}
```

### EC2 with Application Load Balancer — Terraform Snippet

```hcl
# main.tf — ALB + Target Group for Django on EC2

resource "aws_lb" "app" {
  name               = "myapp-alb"
  internal           = false
  load_balancer_type = "application"
  security_groups    = [aws_security_group.alb.id]
  subnets            = var.public_subnet_ids

  enable_deletion_protection = true

  access_logs {
    bucket  = aws_s3_bucket.alb_logs.bucket
    prefix  = "myapp-alb"
    enabled = true
  }
}

resource "aws_lb_target_group" "app" {
  name        = "myapp-tg"
  port        = 8000
  protocol    = "HTTP"
  vpc_id      = var.vpc_id
  target_type = "ip"   # Use "instance" for EC2, "ip" for Fargate

  health_check {
    enabled             = true
    healthy_threshold   = 2
    unhealthy_threshold = 3
    timeout             = 5
    interval            = 30
    path                = "/health/"
    matcher             = "200"
  }
}

resource "aws_lb_listener" "https" {
  load_balancer_arn = aws_lb.app.arn
  port              = 443
  protocol          = "HTTPS"
  ssl_policy        = "ELBSecurityPolicy-TLS13-1-2-2021-06"
  certificate_arn   = var.acm_certificate_arn

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.app.arn
  }
}
```

### RDS PostgreSQL Connection (Django)

```python
# settings/production.py — RDS PostgreSQL

import boto3

def get_ssm_parameter(name):
    client = boto3.client("ssm", region_name="us-east-1")
    response = client.get_parameter(Name=name, WithDecryption=True)
    return response["Parameter"]["Value"]

DATABASES = {
    "default": {
        "ENGINE": "django.db.backends.postgresql",
        "NAME": os.environ.get("DB_NAME", get_ssm_parameter("/myapp/DB_NAME")),
        "USER": os.environ.get("DB_USER", get_ssm_parameter("/myapp/DB_USER")),
        "PASSWORD": os.environ.get("DB_PASSWORD", get_ssm_parameter("/myapp/DB_PASSWORD")),
        "HOST": os.environ.get("DB_HOST", get_ssm_parameter("/myapp/DB_HOST")),
        "PORT": "5432",
        "CONN_MAX_AGE": 60,      # persistent connections
        "CONN_HEALTH_CHECKS": True,
        "OPTIONS": {
            "sslmode": "require",
            "connect_timeout": 5,
        },
    }
}
```

### S3 for Static and Media Files (django-storages)

```python
# settings/production.py — S3 storage

AWS_STORAGE_BUCKET_NAME = os.environ["AWS_STORAGE_BUCKET_NAME"]
AWS_S3_REGION_NAME = "us-east-1"
AWS_S3_CUSTOM_DOMAIN = f"{AWS_STORAGE_BUCKET_NAME}.s3.amazonaws.com"
AWS_S3_FILE_OVERWRITE = False
AWS_DEFAULT_ACL = None  # use bucket policy, not per-object ACLs
AWS_S3_OBJECT_PARAMETERS = {
    "CacheControl": "max-age=86400",
}

# Use separate buckets for static and media
STORAGES = {
    "default": {
        "BACKEND": "storages.backends.s3boto3.S3Boto3Storage",
        "OPTIONS": {
            "bucket_name": os.environ["MEDIA_BUCKET_NAME"],
            "location": "media",
        },
    },
    "staticfiles": {
        "BACKEND": "storages.backends.s3boto3.S3StaticStorage",
        "OPTIONS": {
            "bucket_name": os.environ["STATIC_BUCKET_NAME"],
            "location": "static",
        },
    },
}
```

### CloudWatch Structured Logging

```python
# settings/production.py — structured logging to CloudWatch

LOGGING = {
    "version": 1,
    "disable_existing_loggers": False,
    "formatters": {
        "json": {
            "()": "pythonjsonlogger.jsonlogger.JsonFormatter",
            "format": "%(asctime)s %(name)s %(levelname)s %(message)s %(pathname)s %(lineno)d",
        }
    },
    "handlers": {
        "console": {
            "class": "logging.StreamHandler",
            "formatter": "json",
        }
    },
    "root": {
        "handlers": ["console"],
        "level": "INFO",
    },
    "loggers": {
        "django": {"handlers": ["console"], "level": "INFO", "propagate": False},
        "django.request": {"handlers": ["console"], "level": "WARNING", "propagate": False},
        "django.security": {"handlers": ["console"], "level": "ERROR", "propagate": False},
    },
}
```

### SSM Parameter Store — Fetching Secrets at Startup

```bash
# Fetch all parameters under a path prefix at container startup
# entrypoint.sh
#!/bin/bash
set -e

# Export SSM parameters as environment variables
if [ -n "$SSM_PARAMETER_PATH" ]; then
    eval $(aws ssm get-parameters-by-path \
        --path "$SSM_PARAMETER_PATH" \
        --with-decryption \
        --query "Parameters[*].[Name,Value]" \
        --output text \
        | awk '{split($1,a,"/"); printf "export %s=%s\n", a[length(a)], $2}')
fi

exec "$@"
```

---

## Health Checks & Monitoring

### Django Health Check Endpoint

```python
# myproject/health/views.py
import time
from django.db import connections
from django.db.utils import OperationalError
from django.core.cache import cache
from django.http import JsonResponse


def health_check(request):
    """Lightweight liveness probe — just returns 200 if the process is alive."""
    return JsonResponse({"status": "ok"})


def readiness_check(request):
    """Readiness probe — checks all dependencies before accepting traffic."""
    checks = {}
    status_code = 200

    # Database check
    try:
        start = time.monotonic()
        conn = connections["default"]
        conn.ensure_connection()
        conn.cursor().execute("SELECT 1")
        checks["database"] = {
            "status": "ok",
            "latency_ms": round((time.monotonic() - start) * 1000, 2),
        }
    except OperationalError as e:
        checks["database"] = {"status": "error", "detail": str(e)}
        status_code = 503

    # Redis/cache check
    try:
        start = time.monotonic()
        cache.set("health_check", "ok", timeout=5)
        val = cache.get("health_check")
        assert val == "ok", "Cache read/write mismatch"
        checks["cache"] = {
            "status": "ok",
            "latency_ms": round((time.monotonic() - start) * 1000, 2),
        }
    except Exception as e:
        checks["cache"] = {"status": "error", "detail": str(e)}
        status_code = 503

    return JsonResponse(
        {"status": "ok" if status_code == 200 else "degraded", "checks": checks},
        status=status_code,
    )
```

```python
# myproject/urls.py
from django.urls import path
from myproject.health.views import health_check, readiness_check

urlpatterns = [
    path("health/", health_check),
    path("readyz/", readiness_check),
    # ... other urls
]
```

### FastAPI Health Check with Dependency Checks

```python
# app/routers/health.py
import asyncio
import time
from fastapi import APIRouter, Response
from sqlalchemy import text
from app.database import async_session_maker
from app.redis import get_redis_client

router = APIRouter(tags=["health"])


@router.get("/health", include_in_schema=False)
async def liveness():
    """Liveness probe — process is alive and can serve requests."""
    return {"status": "ok"}


@router.get("/readyz", include_in_schema=False)
async def readiness(response: Response):
    """Readiness probe — all dependencies are available."""
    checks = {}
    failed = False

    # Database check
    try:
        start = time.monotonic()
        async with async_session_maker() as session:
            await session.execute(text("SELECT 1"))
        checks["database"] = {
            "status": "ok",
            "latency_ms": round((time.monotonic() - start) * 1000, 2),
        }
    except Exception as e:
        checks["database"] = {"status": "error", "detail": str(e)}
        failed = True

    # Redis check
    try:
        start = time.monotonic()
        redis = await get_redis_client()
        await redis.ping()
        checks["redis"] = {
            "status": "ok",
            "latency_ms": round((time.monotonic() - start) * 1000, 2),
        }
    except Exception as e:
        checks["redis"] = {"status": "error", "detail": str(e)}
        failed = True

    if failed:
        response.status_code = 503

    return {"status": "degraded" if failed else "ok", "checks": checks}
```

### Flask Health Check

```python
# app/health.py
import time
from flask import Blueprint, jsonify, current_app
from sqlalchemy import text

health_bp = Blueprint("health", __name__)


@health_bp.route("/health")
def liveness():
    return jsonify({"status": "ok"})


@health_bp.route("/readyz")
def readiness():
    checks = {}
    status_code = 200

    # Database check
    try:
        start = time.monotonic()
        db = current_app.extensions["sqlalchemy"]
        db.session.execute(text("SELECT 1"))
        checks["database"] = {
            "status": "ok",
            "latency_ms": round((time.monotonic() - start) * 1000, 2),
        }
    except Exception as e:
        checks["database"] = {"status": "error", "detail": str(e)}
        status_code = 503

    return jsonify({"status": "ok" if status_code == 200 else "degraded", "checks": checks}), status_code
```

### Prometheus Metrics Export

```python
# app/metrics.py — Prometheus metrics for FastAPI
from prometheus_client import Counter, Histogram, Gauge, generate_latest, CONTENT_TYPE_LATEST
from fastapi import Request, Response
import time

REQUEST_COUNT = Counter(
    "http_requests_total",
    "Total HTTP requests",
    ["method", "endpoint", "status_code"],
)
REQUEST_LATENCY = Histogram(
    "http_request_duration_seconds",
    "HTTP request latency",
    ["method", "endpoint"],
    buckets=[0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0],
)
ACTIVE_REQUESTS = Gauge("http_requests_active", "Active HTTP requests")


async def metrics_middleware(request: Request, call_next):
    ACTIVE_REQUESTS.inc()
    start = time.monotonic()
    response = await call_next(request)
    duration = time.monotonic() - start

    REQUEST_COUNT.labels(
        method=request.method,
        endpoint=request.url.path,
        status_code=response.status_code,
    ).inc()
    REQUEST_LATENCY.labels(
        method=request.method,
        endpoint=request.url.path,
    ).observe(duration)
    ACTIVE_REQUESTS.dec()

    return response


# In main.py:
# app.middleware("http")(metrics_middleware)
# @app.get("/metrics", include_in_schema=False)
# async def metrics():
#     return Response(generate_latest(), media_type=CONTENT_TYPE_LATEST)
```

### Sentry Integration

```python
# settings/production.py — Sentry error tracking

import sentry_sdk
from sentry_sdk.integrations.django import DjangoIntegration
from sentry_sdk.integrations.celery import CeleryIntegration
from sentry_sdk.integrations.redis import RedisIntegration
from sentry_sdk.integrations.logging import LoggingIntegration
import logging

sentry_sdk.init(
    dsn=os.environ.get("SENTRY_DSN"),
    integrations=[
        DjangoIntegration(
            transaction_style="url",
            middleware_spans=True,
            signals_spans=False,
        ),
        CeleryIntegration(monitor_beat_tasks=True),
        RedisIntegration(),
        LoggingIntegration(
            level=logging.INFO,       # capture INFO and above as breadcrumbs
            event_level=logging.ERROR,  # send ERROR and above as events
        ),
    ],
    traces_sample_rate=0.1,    # 10% of transactions for performance monitoring
    profiles_sample_rate=0.01, # 1% for profiling
    environment=os.environ.get("ENVIRONMENT", "production"),
    release=os.environ.get("GIT_SHA"),
    send_default_pii=False,    # do not send user IPs or emails by default
)
```

### Structlog Configuration

```python
# myproject/logging_config.py
import logging
import structlog

def configure_structlog():
    shared_processors = [
        structlog.contextvars.merge_contextvars,
        structlog.processors.add_log_level,
        structlog.processors.TimeStamper(fmt="iso"),
        structlog.stdlib.add_logger_name,
    ]

    if os.environ.get("LOG_FORMAT") == "json":
        renderer = structlog.processors.JSONRenderer()
    else:
        renderer = structlog.dev.ConsoleRenderer(colors=True)

    structlog.configure(
        processors=shared_processors + [
            structlog.stdlib.ProcessorFormatter.wrap_for_formatter,
        ],
        wrapper_class=structlog.make_filtering_bound_logger(logging.INFO),
        logger_factory=structlog.stdlib.LoggerFactory(),
        cache_logger_on_first_use=True,
    )

# Usage:
# import structlog
# log = structlog.get_logger(__name__)
# log.info("user_login", user_id=42, ip="1.2.3.4")
```

---

## CI/CD Pipeline

### GitHub Actions Workflow

```yaml
# .github/workflows/deploy.yml
name: Test, Build, and Deploy

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

env:
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}

jobs:
  test:
    name: Test & Lint
    runs-on: ubuntu-latest

    services:
      postgres:
        image: postgres:16-alpine
        env:
          POSTGRES_DB: testdb
          POSTGRES_USER: testuser
          POSTGRES_PASSWORD: testpass
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 5432:5432

      redis:
        image: redis:7-alpine
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 6379:6379

    steps:
      - uses: actions/checkout@v4

      - name: Set up Python
        uses: actions/setup-python@v5
        with:
          python-version: "3.12"
          cache: "pip"

      - name: Install dependencies
        run: |
          pip install --upgrade pip
          pip install -r requirements.txt -r requirements-dev.txt

      - name: Lint with Ruff
        run: ruff check .

      - name: Type check with mypy
        run: mypy . --ignore-missing-imports

      - name: Run tests
        env:
          DATABASE_URL: postgres://testuser:testpass@localhost:5432/testdb
          REDIS_URL: redis://localhost:6379/0
          DJANGO_SETTINGS_MODULE: myproject.settings.test
          SECRET_KEY: test-secret-key-not-for-production
        run: |
          python manage.py migrate --noinput
          pytest --cov=. --cov-report=xml --cov-report=term-missing -q

      - name: Upload coverage
        uses: codecov/codecov-action@v4
        with:
          token: ${{ secrets.CODECOV_TOKEN }}
          file: coverage.xml

  build:
    name: Build & Push Docker Image
    runs-on: ubuntu-latest
    needs: test
    if: github.ref == 'refs/heads/main'
    permissions:
      contents: read
      packages: write

    steps:
      - uses: actions/checkout@v4

      - name: Log in to Container Registry
        uses: docker/login-action@v3
        with:
          registry: ${{ env.REGISTRY }}
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Extract metadata
        id: meta
        uses: docker/metadata-action@v5
        with:
          images: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}
          tags: |
            type=sha,prefix=sha-
            type=ref,event=branch
            type=raw,value=latest,enable=${{ github.ref == 'refs/heads/main' }}

      - name: Build and push
        uses: docker/build-push-action@v5
        with:
          context: .
          target: production
          push: true
          tags: ${{ steps.meta.outputs.tags }}
          labels: ${{ steps.meta.outputs.labels }}
          cache-from: type=gha
          cache-to: type=gha,mode=max
          build-args: |
            GIT_SHA=${{ github.sha }}

  deploy:
    name: Deploy to Production
    runs-on: ubuntu-latest
    needs: build
    if: github.ref == 'refs/heads/main'
    environment: production

    steps:
      - name: Deploy to ECS
        uses: aws-actions/amazon-ecs-deploy-task-definition@v1
        with:
          task-definition: ecs-task-definition.json
          service: myapp-web
          cluster: myapp-cluster
          wait-for-service-stability: true
        env:
          AWS_ACCESS_KEY_ID: ${{ secrets.AWS_ACCESS_KEY_ID }}
          AWS_SECRET_ACCESS_KEY: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
          AWS_DEFAULT_REGION: us-east-1

      - name: Notify Sentry of deployment
        run: |
          curl -X POST "https://sentry.io/api/0/organizations/${{ secrets.SENTRY_ORG }}/releases/" \
            -H "Authorization: Bearer ${{ secrets.SENTRY_AUTH_TOKEN }}" \
            -H "Content-Type: application/json" \
            -d '{
              "version": "${{ github.sha }}",
              "projects": ["myapp"],
              "refs": [{"repository": "${{ github.repository }}", "commit": "${{ github.sha }}"}]
            }'
```

### requirements-dev.txt

```
# requirements-dev.txt
pytest
pytest-django
pytest-asyncio
pytest-cov
pytest-xdist           # parallel test execution
factory-boy            # test fixtures
faker
httpx                  # async test client for FastAPI
ruff                   # fast linter
mypy
django-stubs           # type stubs for Django
types-redis
pre-commit
```

### pre-commit Configuration

```yaml
# .pre-commit-config.yaml
repos:
  - repo: https://github.com/astral-sh/ruff-pre-commit
    rev: v0.4.4
    hooks:
      - id: ruff
        args: [--fix]
      - id: ruff-format

  - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v4.6.0
    hooks:
      - id: trailing-whitespace
      - id: end-of-file-fixer
      - id: check-yaml
      - id: check-json
      - id: check-merge-conflict
      - id: detect-private-key

  - repo: https://github.com/pre-commit/mirrors-mypy
    rev: v1.10.0
    hooks:
      - id: mypy
        additional_dependencies: [django-stubs, types-redis]
```

---

## Quick Reference — Common Commands

```bash
# Start development server
python manage.py runserver                          # Django
flask run --debug                                   # Flask
uvicorn app.main:app --reload                       # FastAPI

# Database migrations
python manage.py makemigrations && python manage.py migrate    # Django
alembic revision --autogenerate -m "description"              # SQLAlchemy/FastAPI
alembic upgrade head

# Collect static files (Django)
python manage.py collectstatic --noinput

# Gunicorn production start
gunicorn myproject.wsgi:application --config gunicorn.conf.py

# Gunicorn + Uvicorn for ASGI
gunicorn app.main:app --worker-class uvicorn.workers.UvicornWorker --workers 4

# Docker build and run
docker build --target production -t myapp:latest .
docker run --env-file .env -p 8000:8000 myapp:latest

# Docker Compose
docker compose up -d
docker compose logs -f web
docker compose exec web python manage.py shell
docker compose down -v   # remove volumes

# Celery
celery -A myproject worker --loglevel=info
celery -A myproject beat --loglevel=info
celery -A myproject flower   # monitoring UI at :5555

# nginx
sudo nginx -t                  # test config
sudo systemctl reload nginx    # reload without downtime

# Let's Encrypt
sudo certbot --nginx -d example.com -d www.example.com
sudo certbot renew --dry-run   # test auto-renewal
```
