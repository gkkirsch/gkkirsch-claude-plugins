# DevOps & CI/CD Cheatsheet

## GitHub Actions Workflow Syntax

```yaml
name: CI/CD
on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

permissions:
  contents: read

concurrency:
  group: ${{ github.workflow }}-${{ github.ref }}
  cancel-in-progress: true

jobs:
  job-name:
    runs-on: ubuntu-22.04
    timeout-minutes: 15
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: 20
          cache: "npm"
      - run: npm ci
      - run: npm test
```

## GitHub Actions Contexts

| Context | Example | Use |
|---------|---------|-----|
| `github.sha` | `abc123` | Current commit SHA |
| `github.ref` | `refs/heads/main` | Branch or tag ref |
| `github.ref_name` | `main` | Branch or tag name |
| `github.event_name` | `push` | Trigger event |
| `github.actor` | `username` | User who triggered |
| `github.repository` | `owner/repo` | Full repo name |
| `github.run_number` | `42` | Sequential run number |
| `github.workspace` | `/home/runner/work/repo` | Checkout directory |
| `runner.os` | `Linux` | Runner OS |
| `secrets.GITHUB_TOKEN` | `ghp_...` | Auto-generated token |

## Conditional Steps

```yaml
# Only on push to main
- if: github.event_name == 'push' && github.ref == 'refs/heads/main'

# Only on pull request
- if: github.event_name == 'pull_request'

# Only on success/failure
- if: success()
- if: failure()
- if: always()  # Run even if previous steps failed
- if: cancelled()

# Check for file changes
- uses: dorny/paths-filter@v3
  id: changes
  with:
    filters: |
      src: 'src/**'
      docs: 'docs/**'

- if: steps.changes.outputs.src == 'true'
  run: npm test
```

## Dockerfile Best Practices

```dockerfile
# Multi-stage build
FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM node:20-alpine AS runner
WORKDIR /app
RUN addgroup --system --gid 1001 nodejs && \
    adduser --system --uid 1001 appuser
COPY --from=builder --chown=appuser:nodejs /app/dist ./dist
COPY --from=builder --chown=appuser:nodejs /app/node_modules ./node_modules
COPY --from=builder --chown=appuser:nodejs /app/package.json ./
USER appuser
EXPOSE 3000
HEALTHCHECK --interval=30s --timeout=3s CMD wget -qO- http://localhost:3000/health || exit 1
CMD ["node", "dist/server.js"]
```

## Docker Commands

| Command | Use |
|---------|-----|
| `docker build -t myapp .` | Build image |
| `docker build --no-cache -t myapp .` | Build without cache |
| `docker run -p 3000:3000 myapp` | Run container |
| `docker run -d --name myapp -p 3000:3000 myapp` | Run detached |
| `docker compose up -d` | Start services |
| `docker compose down` | Stop services |
| `docker compose logs -f app` | Follow service logs |
| `docker compose exec app sh` | Shell into service |
| `docker ps` | List running containers |
| `docker images` | List images |
| `docker system prune -a` | Clean up everything |
| `docker stats` | Resource usage |

## Docker Compose

```yaml
services:
  app:
    build: .
    ports: ["3000:3000"]
    environment:
      DATABASE_URL: postgresql://postgres:postgres@db:5432/myapp
      REDIS_URL: redis://redis:6379
    depends_on:
      db:
        condition: service_healthy
      redis:
        condition: service_started
    restart: unless-stopped

  db:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: myapp
    volumes:
      - pgdata:/var/lib/postgresql/data
    ports: ["5432:5432"]
    healthcheck:
      test: pg_isready -U postgres
      interval: 5s
      timeout: 3s
      retries: 5

  redis:
    image: redis:7-alpine
    ports: ["6379:6379"]
    volumes:
      - redisdata:/data

volumes:
  pgdata:
  redisdata:
```

## Deployment Platform Commands

### Heroku

| Command | Use |
|---------|-----|
| `heroku create myapp` | Create app |
| `git push heroku main` | Deploy |
| `heroku logs --tail` | Stream logs |
| `heroku run bash` | Interactive shell |
| `heroku config:set KEY=value` | Set env var |
| `heroku config` | List env vars |
| `heroku ps` | List dynos |
| `heroku ps:scale web=2` | Scale dynos |
| `heroku releases` | List releases |
| `heroku rollback v42` | Rollback to version |
| `heroku pg:psql` | Database console |
| `heroku pg:info` | Database info |
| `heroku maintenance:on` | Enable maintenance mode |

### Cloudflare

| Command | Use |
|---------|-----|
| `npx wrangler pages deploy dist` | Deploy Pages |
| `npx wrangler deploy` | Deploy Worker |
| `npx wrangler dev` | Local development |
| `npx wrangler tail` | Stream logs |
| `npx wrangler secret put NAME` | Set secret |
| `npx wrangler pages project list` | List projects |

### PM2 (Process Manager)

| Command | Use |
|---------|-----|
| `pm2 start app.js` | Start process |
| `pm2 start ecosystem.config.js` | Start from config |
| `pm2 reload all` | Zero-downtime restart |
| `pm2 restart all` | Hard restart |
| `pm2 stop all` | Stop all processes |
| `pm2 logs` | View logs |
| `pm2 monit` | Real-time monitor |
| `pm2 save` | Save current process list |
| `pm2 startup` | Generate startup script |

## Nginx Configuration

```nginx
# Reverse proxy for Node.js
upstream app {
    server 127.0.0.1:3000;
    keepalive 64;
}

server {
    listen 80;
    server_name myapp.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name myapp.com;

    ssl_certificate /etc/letsencrypt/live/myapp.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/myapp.com/privkey.pem;

    # Security headers
    add_header X-Frame-Options DENY;
    add_header X-Content-Type-Options nosniff;
    add_header X-XSS-Protection "1; mode=block";
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains";

    # Gzip
    gzip on;
    gzip_types text/plain text/css application/json application/javascript;
    gzip_min_length 1000;

    # Static files
    location /static/ {
        alias /app/public/;
        expires 30d;
        add_header Cache-Control "public, immutable";
    }

    # Proxy to app
    location / {
        proxy_pass http://app;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
    }
}
```

## Monitoring Stack Quick Start

```yaml
# docker-compose.monitoring.yml
services:
  prometheus:
    image: prom/prometheus:latest
    ports: ["9090:9090"]
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml

  grafana:
    image: grafana/grafana:latest
    ports: ["3001:3000"]
    environment:
      GF_SECURITY_ADMIN_PASSWORD: admin
    volumes:
      - grafana_data:/var/lib/grafana

volumes:
  grafana_data:
```

```yaml
# prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: "app"
    static_configs:
      - targets: ["host.docker.internal:3000"]
    metrics_path: /metrics
```

## SSL/TLS Setup (Let's Encrypt)

```bash
# Certbot (standalone)
sudo certbot certonly --standalone -d myapp.com -d www.myapp.com

# Certbot (with nginx)
sudo certbot --nginx -d myapp.com -d www.myapp.com

# Auto-renewal
sudo certbot renew --dry-run  # Test
# Cron: 0 0 * * * certbot renew --quiet
```

## Environment Variable Management

```bash
# .env file format
DATABASE_URL=postgresql://user:pass@localhost:5432/myapp
REDIS_URL=redis://localhost:6379
API_KEY=sk-abc123
NODE_ENV=production
PORT=3000

# dotenv loading
npm install dotenv
# In code: import 'dotenv/config'

# Never commit .env files
echo ".env*" >> .gitignore
echo "!.env.example" >> .gitignore
```

```bash
# .env.example (commit this, no real values)
DATABASE_URL=postgresql://user:password@localhost:5432/dbname
REDIS_URL=redis://localhost:6379
API_KEY=your-api-key-here
NODE_ENV=development
PORT=3000
```

## Key SRE Formulas

| Metric | Formula |
|--------|---------|
| Availability | `uptime / total_time` |
| Error budget | `1 - SLO` (e.g., 99.9% SLO = 0.1% error budget) |
| Error budget remaining | `error_budget - (errors / total_requests)` |
| MTTR | `total_downtime / number_of_incidents` |
| MTTF | `total_uptime / number_of_failures` |
| MTBF | `MTTF + MTTR` |
| Change failure rate | `failed_deploys / total_deploys` |
| Deploy frequency | `deploys / time_period` |

| Availability | Monthly Downtime |
|-------------|-----------------|
| 99% (two nines) | 7.3 hours |
| 99.9% (three nines) | 43.8 minutes |
| 99.95% | 21.9 minutes |
| 99.99% (four nines) | 4.3 minutes |
| 99.999% (five nines) | 26.3 seconds |
