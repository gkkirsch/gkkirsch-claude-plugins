# Cloud Deployment Guides

Step-by-step deployment guides for every major cloud platform. Each guide takes you from zero to production with real commands and configurations.

## Heroku — Complete Setup

### Prerequisites

```bash
# Install Heroku CLI
brew tap heroku/brew && brew install heroku  # macOS
# or: curl https://cli-assets.heroku.com/install.sh | sh  # Linux

# Login
heroku login
```

### Step 1: Create the App

```bash
# Create app (Heroku assigns a random name)
heroku create

# Or with a specific name
heroku create my-app-name

# Or in a specific region
heroku create my-app-name --region eu
```

### Step 2: Configure Buildpack

```bash
# Auto-detected for most languages, but you can set explicitly:
heroku buildpacks:set heroku/nodejs
heroku buildpacks:set heroku/python

# Multiple buildpacks (e.g., Node.js frontend + Python backend)
heroku buildpacks:add --index 1 heroku/nodejs
heroku buildpacks:add --index 2 heroku/python
```

### Step 3: Add Database

```bash
# PostgreSQL (most common)
heroku addons:create heroku-postgresql:essential-0   # $5/mo, 10K rows
heroku addons:create heroku-postgresql:essential-1   # $15/mo, 10M rows
heroku addons:create heroku-postgresql:standard-0    # $50/mo, 64GB

# Redis
heroku addons:create heroku-redis:mini               # Free, 25MB
heroku addons:create heroku-redis:premium-0           # $15/mo, 50MB

# Check database URL (auto-set as DATABASE_URL)
heroku config:get DATABASE_URL
```

### Step 4: Set Environment Variables

```bash
heroku config:set NODE_ENV=production
heroku config:set JWT_SECRET=$(openssl rand -hex 32)
heroku config:set SENTRY_DSN=https://xxx@sentry.io/xxx
heroku config:set API_KEY=your-api-key

# View all config
heroku config

# Remove a variable
heroku config:unset OLD_VARIABLE
```

### Step 5: Create Required Files

**Procfile** (tells Heroku what to run):
```
web: node dist/index.js
worker: node dist/worker.js
release: npx drizzle-kit push
```

**app.json** (for Heroku Button and Review Apps):
```json
{
  "name": "My App",
  "description": "A great application",
  "formation": {
    "web": { "quantity": 1, "size": "basic" }
  },
  "addons": [
    { "plan": "heroku-postgresql:essential-0" }
  ],
  "env": {
    "NODE_ENV": { "value": "production" },
    "JWT_SECRET": { "generator": "secret" }
  },
  "buildpacks": [
    { "url": "heroku/nodejs" }
  ]
}
```

### Step 6: Deploy

```bash
# Git push deploy
git push heroku main

# If your branch isn't main:
git push heroku my-branch:main

# Or use container deploy
heroku container:push web
heroku container:release web
```

### Step 7: Post-Deploy

```bash
# Run database migrations
heroku run npx drizzle-kit push

# Open the app
heroku open

# View logs (live)
heroku logs --tail

# Check app status
heroku ps

# Scale dynos
heroku ps:scale web=2 worker=1
```

### Step 8: Pipelines (Staging → Production)

```bash
# Create pipeline
heroku pipelines:create my-pipeline

# Add apps to pipeline
heroku pipelines:add my-pipeline --app my-app-staging --stage staging
heroku pipelines:add my-pipeline --app my-app-prod --stage production

# Promote staging to production
heroku pipelines:promote --app my-app-staging
```

### Step 9: Custom Domain

```bash
heroku domains:add example.com
heroku domains:add www.example.com

# Get DNS target
heroku domains

# Add CNAME record pointing to the DNS target in your domain registrar
```

### Heroku Quick Reference

| Command | Purpose |
|---------|---------|
| `heroku logs --tail` | Live logs |
| `heroku ps` | Process status |
| `heroku run bash` | SSH into dyno |
| `heroku pg:info` | Database info |
| `heroku pg:psql` | PostgreSQL shell |
| `heroku restart` | Restart all dynos |
| `heroku releases` | Deployment history |
| `heroku rollback v42` | Rollback to version |
| `heroku maintenance:on` | Enable maintenance mode |

---

## Vercel — Static + Serverless + Edge

### Prerequisites

```bash
npm install -g vercel
vercel login
```

### Static Site / SPA Deployment

```bash
# From project root with build output in dist/
vercel

# Follow prompts:
# - Link to existing project? No
# - Project name? my-app
# - Framework? Other
# - Build command? npm run build
# - Output directory? dist
# - Override settings? No

# Production deploy
vercel --prod
```

### vercel.json Configuration

```json
{
  "buildCommand": "npm run build",
  "outputDirectory": "dist",
  "installCommand": "npm ci",
  "rewrites": [
    { "source": "/(.*)", "destination": "/index.html" }
  ],
  "headers": [
    {
      "source": "/assets/(.*)",
      "headers": [
        {
          "key": "Cache-Control",
          "value": "public, max-age=31536000, immutable"
        }
      ]
    },
    {
      "source": "/(.*)",
      "headers": [
        { "key": "X-Frame-Options", "value": "DENY" },
        { "key": "X-Content-Type-Options", "value": "nosniff" },
        { "key": "Referrer-Policy", "value": "strict-origin-when-cross-origin" }
      ]
    }
  ]
}
```

### Serverless API Routes

Create files in `api/` directory:

```typescript
// api/hello.ts
import type { VercelRequest, VercelResponse } from '@vercel/node';

export default function handler(req: VercelRequest, res: VercelResponse) {
  const { name = 'World' } = req.query;
  res.status(200).json({ message: `Hello ${name}!` });
}
```

```typescript
// api/users/[id].ts — Dynamic routes
import type { VercelRequest, VercelResponse } from '@vercel/node';

export default async function handler(req: VercelRequest, res: VercelResponse) {
  const { id } = req.query;

  if (req.method === 'GET') {
    // fetch user by id
    res.json({ id, name: 'John' });
  } else if (req.method === 'PUT') {
    // update user
    res.json({ id, ...req.body });
  } else {
    res.setHeader('Allow', 'GET, PUT');
    res.status(405).end();
  }
}
```

### Edge Functions

```typescript
// api/edge.ts
export const config = { runtime: 'edge' };

export default async function handler(request: Request) {
  const { searchParams } = new URL(request.url);
  const name = searchParams.get('name') || 'World';

  return new Response(
    JSON.stringify({ message: `Hello ${name} from the edge!` }),
    {
      status: 200,
      headers: { 'Content-Type': 'application/json' },
    }
  );
}
```

### Environment Variables

```bash
# Add env vars (prompts for value)
vercel env add DATABASE_URL production
vercel env add JWT_SECRET production preview

# Pull env vars to local .env
vercel env pull .env.local

# List env vars
vercel env ls
```

### Custom Domain

```bash
vercel domains add example.com
vercel domains add www.example.com

# Verify DNS
vercel domains inspect example.com
```

### Vercel Quick Reference

| Command | Purpose |
|---------|---------|
| `vercel` | Preview deploy |
| `vercel --prod` | Production deploy |
| `vercel ls` | List deployments |
| `vercel inspect <url>` | Deployment details |
| `vercel logs <url>` | View function logs |
| `vercel env ls` | List env vars |
| `vercel domains ls` | List domains |
| `vercel rollback` | Rollback to previous |

---

## Cloudflare — Pages + Workers + D1

### Prerequisites

```bash
npm install -g wrangler
wrangler login
```

### Cloudflare Pages (Static Sites)

```bash
# Deploy from build output
wrangler pages deploy dist --project-name my-site

# Or connect to Git (auto-deploy on push)
wrangler pages project create my-site
# Then connect via Cloudflare Dashboard → Pages → Create → Connect to Git
```

### Cloudflare Workers

**Create a new Worker project**:
```bash
npm create cloudflare@latest my-api -- --template hello-world
cd my-api
```

**wrangler.toml**:
```toml
name = "my-api"
main = "src/index.ts"
compatibility_date = "2024-01-01"
compatibility_flags = ["nodejs_compat"]

# Environment variables
[vars]
ENVIRONMENT = "production"
API_VERSION = "v1"

# D1 Database
[[d1_databases]]
binding = "DB"
database_name = "my-database"
database_id = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"

# KV Namespace (key-value store)
[[kv_namespaces]]
binding = "CACHE"
id = "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"

# R2 Bucket (object storage)
[[r2_buckets]]
binding = "STORAGE"
bucket_name = "my-files"

# Durable Objects
[[durable_objects.bindings]]
name = "COUNTER"
class_name = "Counter"

# Cron triggers
[triggers]
crons = ["0 * * * *"]  # Every hour
```

**Worker with Hono**:
```typescript
// src/index.ts
import { Hono } from 'hono';
import { cors } from 'hono/cors';

type Env = {
  DB: D1Database;
  CACHE: KVNamespace;
  STORAGE: R2Bucket;
  ENVIRONMENT: string;
};

const app = new Hono<{ Bindings: Env }>();

app.use('*', cors());

app.get('/health', (c) => c.json({ status: 'ok', env: c.env.ENVIRONMENT }));

// D1 Database queries
app.get('/api/users', async (c) => {
  const { results } = await c.env.DB.prepare('SELECT * FROM users LIMIT 100').all();
  return c.json(results);
});

app.post('/api/users', async (c) => {
  const { name, email } = await c.req.json();
  const result = await c.env.DB
    .prepare('INSERT INTO users (name, email) VALUES (?, ?) RETURNING *')
    .bind(name, email)
    .first();
  return c.json(result, 201);
});

// KV Cache
app.get('/api/config/:key', async (c) => {
  const key = c.req.param('key');
  const cached = await c.env.CACHE.get(key);
  if (cached) return c.json(JSON.parse(cached));
  return c.json({ error: 'Not found' }, 404);
});

// R2 Storage
app.post('/api/upload', async (c) => {
  const formData = await c.req.formData();
  const file = formData.get('file') as File;
  const key = `uploads/${Date.now()}-${file.name}`;
  await c.env.STORAGE.put(key, file.stream(), {
    httpMetadata: { contentType: file.type },
  });
  return c.json({ key, url: `/api/files/${key}` });
});

export default app;
```

### D1 Database Setup

```bash
# Create database
wrangler d1 create my-database
# Note the database_id from output, add to wrangler.toml

# Create schema
cat > schema.sql << 'EOF'
CREATE TABLE IF NOT EXISTS users (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  email TEXT UNIQUE NOT NULL,
  created_at TEXT DEFAULT (datetime('now'))
);
CREATE INDEX idx_users_email ON users(email);
EOF

# Apply schema
wrangler d1 execute my-database --file=schema.sql

# Run queries
wrangler d1 execute my-database --command="SELECT * FROM users"

# Deploy
wrangler deploy
```

### Custom Domains for Workers

```bash
# Add custom domain (must be on Cloudflare DNS)
# Via Dashboard: Workers & Pages → your worker → Settings → Triggers → Custom Domains
# Or via wrangler.toml:
```

```toml
routes = [
  { pattern = "api.example.com/*", zone_name = "example.com" }
]
```

### Cloudflare Quick Reference

| Command | Purpose |
|---------|---------|
| `wrangler deploy` | Deploy Worker |
| `wrangler dev` | Local development |
| `wrangler tail` | Live logs |
| `wrangler d1 execute DB --command "SQL"` | Run SQL query |
| `wrangler kv key put KEY VALUE` | Set KV value |
| `wrangler r2 object put BUCKET/key file` | Upload to R2 |
| `wrangler pages deploy dist` | Deploy Pages |
| `wrangler secret put NAME` | Set secret |

---

## AWS ECS (Fargate) — Container Deployment

### Prerequisites

```bash
# Install AWS CLI
brew install awscli  # macOS
# or: pip install awscli

# Configure credentials
aws configure
# Enter: Access Key ID, Secret Access Key, Region, Output format
```

### Step 1: Create ECR Repository

```bash
# Create container registry
aws ecr create-repository \
  --repository-name my-app \
  --image-scanning-configuration scanOnPush=true

# Login to ECR
aws ecr get-login-password --region us-east-1 | \
  docker login --username AWS --password-stdin \
  ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com
```

### Step 2: Build and Push Image

```bash
# Build
docker build -t my-app .

# Tag
docker tag my-app:latest ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com/my-app:latest

# Push
docker push ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com/my-app:latest
```

### Step 3: Create ECS Cluster

```bash
aws ecs create-cluster --cluster-name my-cluster
```

### Step 4: Create Task Definition

Save as `task-definition.json`:
```json
{
  "family": "my-app",
  "networkMode": "awsvpc",
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "256",
  "memory": "512",
  "executionRoleArn": "arn:aws:iam::ACCOUNT_ID:role/ecsTaskExecutionRole",
  "containerDefinitions": [
    {
      "name": "app",
      "image": "ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com/my-app:latest",
      "portMappings": [
        {
          "containerPort": 3000,
          "protocol": "tcp"
        }
      ],
      "environment": [
        { "name": "NODE_ENV", "value": "production" },
        { "name": "PORT", "value": "3000" }
      ],
      "secrets": [
        {
          "name": "DATABASE_URL",
          "valueFrom": "arn:aws:ssm:us-east-1:ACCOUNT_ID:parameter/my-app/database-url"
        },
        {
          "name": "JWT_SECRET",
          "valueFrom": "arn:aws:ssm:us-east-1:ACCOUNT_ID:parameter/my-app/jwt-secret"
        }
      ],
      "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "/ecs/my-app",
          "awslogs-region": "us-east-1",
          "awslogs-stream-prefix": "ecs",
          "awslogs-create-group": "true"
        }
      },
      "healthCheck": {
        "command": ["CMD-SHELL", "curl -f http://localhost:3000/health || exit 1"],
        "interval": 30,
        "timeout": 5,
        "retries": 3,
        "startPeriod": 60
      },
      "essential": true
    }
  ]
}
```

```bash
# Register task definition
aws ecs register-task-definition --cli-input-json file://task-definition.json
```

### Step 5: Store Secrets

```bash
# Store in AWS Systems Manager Parameter Store
aws ssm put-parameter \
  --name "/my-app/database-url" \
  --type SecureString \
  --value "postgresql://user:pass@host:5432/db"

aws ssm put-parameter \
  --name "/my-app/jwt-secret" \
  --type SecureString \
  --value "$(openssl rand -hex 32)"
```

### Step 6: Create Service

```bash
# First, you need a VPC, subnets, and security group
# Create ALB (Application Load Balancer) via AWS Console or CloudFormation

# Create ECS service
aws ecs create-service \
  --cluster my-cluster \
  --service-name my-app-service \
  --task-definition my-app \
  --desired-count 2 \
  --launch-type FARGATE \
  --network-configuration "awsvpcConfiguration={subnets=[subnet-xxx,subnet-yyy],securityGroups=[sg-xxx],assignPublicIp=ENABLED}" \
  --load-balancers "targetGroupArn=arn:aws:elasticloadbalancing:...,containerName=app,containerPort=3000"
```

### Step 7: Update Deployment

```bash
# Build, tag, and push new image
docker build -t my-app .
docker tag my-app:latest ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com/my-app:$(git rev-parse --short HEAD)
docker push ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com/my-app:$(git rev-parse --short HEAD)

# Force new deployment (pulls latest image)
aws ecs update-service \
  --cluster my-cluster \
  --service my-app-service \
  --force-new-deployment
```

### AWS ECS Quick Reference

| Command | Purpose |
|---------|---------|
| `aws ecs list-services --cluster X` | List services |
| `aws ecs describe-services --cluster X --services Y` | Service details |
| `aws ecs list-tasks --cluster X --service-name Y` | List running tasks |
| `aws ecs execute-command --cluster X --task Z --command bash` | SSH into container |
| `aws logs tail /ecs/my-app --follow` | Live logs |
| `aws ecs update-service --cluster X --service Y --desired-count 3` | Scale |

---

## GCP Cloud Run — Containerized Serverless

### Prerequisites

```bash
# Install gcloud CLI
brew install google-cloud-sdk  # macOS

# Login and configure
gcloud auth login
gcloud config set project MY_PROJECT_ID
gcloud config set run/region us-central1
```

### Deploy from Source (Easiest)

```bash
# Cloud Run builds and deploys from source code
gcloud run deploy my-app \
  --source . \
  --region us-central1 \
  --allow-unauthenticated \
  --set-env-vars NODE_ENV=production,PORT=3000 \
  --memory 512Mi \
  --cpu 1 \
  --min-instances 0 \
  --max-instances 10 \
  --port 3000
```

### Deploy from Container Image

```bash
# Build and push to Google Container Registry
gcloud builds submit --tag gcr.io/MY_PROJECT/my-app

# Deploy
gcloud run deploy my-app \
  --image gcr.io/MY_PROJECT/my-app:latest \
  --region us-central1 \
  --allow-unauthenticated
```

### Configure Secrets

```bash
# Create secret
echo -n "postgresql://user:pass@host/db" | \
  gcloud secrets create database-url --data-file=-

# Grant Cloud Run access
gcloud secrets add-iam-policy-binding database-url \
  --member=serviceAccount:MY_PROJECT_NUMBER-compute@developer.gserviceaccount.com \
  --role=roles/secretmanager.secretAccessor

# Use in Cloud Run
gcloud run deploy my-app \
  --set-secrets DATABASE_URL=database-url:latest
```

### Custom Domain

```bash
gcloud run domain-mappings create \
  --service my-app \
  --domain api.example.com \
  --region us-central1
```

### Cloud SQL (Managed PostgreSQL)

```bash
# Create instance
gcloud sql instances create my-db \
  --database-version=POSTGRES_16 \
  --tier=db-f1-micro \
  --region=us-central1

# Create database
gcloud sql databases create myapp --instance=my-db

# Set password
gcloud sql users set-password postgres \
  --instance=my-db \
  --password=secure-password

# Connect Cloud Run to Cloud SQL
gcloud run deploy my-app \
  --add-cloudsql-instances MY_PROJECT:us-central1:my-db \
  --set-env-vars INSTANCE_CONNECTION_NAME=MY_PROJECT:us-central1:my-db
```

### Cloud Run Quick Reference

| Command | Purpose |
|---------|---------|
| `gcloud run services list` | List services |
| `gcloud run services describe my-app` | Service details |
| `gcloud run services logs read my-app` | View logs |
| `gcloud run deploy --source .` | Deploy from source |
| `gcloud run services update my-app --memory 1Gi` | Update config |
| `gcloud run revisions list --service my-app` | List revisions |
| `gcloud run services update-traffic my-app --to-latest` | Route traffic |

---

## Railway — Developer-Friendly PaaS

### Prerequisites

```bash
npm install -g @railway/cli
railway login
```

### Quick Deploy

```bash
# Initialize (creates project)
railway init

# Link to existing project
railway link

# Deploy
railway up

# Open app
railway open
```

### Add Services

```bash
# Add PostgreSQL
railway add --plugin postgresql

# Add Redis
railway add --plugin redis

# View connection details
railway variables
```

### Environment Variables

```bash
railway variables set NODE_ENV=production
railway variables set JWT_SECRET=$(openssl rand -hex 32)

# View all variables
railway variables
```

### Custom Domain

Configure in Railway Dashboard → Settings → Domains → Add Custom Domain

### railway.json

```json
{
  "$schema": "https://railway.app/railway.schema.json",
  "build": {
    "builder": "NIXPACKS",
    "buildCommand": "npm ci && npm run build"
  },
  "deploy": {
    "startCommand": "node dist/index.js",
    "healthcheckPath": "/health",
    "healthcheckTimeout": 30,
    "restartPolicyType": "ON_FAILURE",
    "restartPolicyMaxRetries": 5
  }
}
```

---

## Fly.io — Global Edge Deployment

### Prerequisites

```bash
brew install flyctl  # macOS
fly auth login
```

### Quick Deploy

```bash
# Launch (creates app + fly.toml)
fly launch

# Deploy
fly deploy
```

### fly.toml

```toml
app = "my-app"
primary_region = "iad"

[build]
  dockerfile = "Dockerfile"

[env]
  NODE_ENV = "production"
  PORT = "3000"

[http_service]
  internal_port = 3000
  force_https = true
  auto_stop_machines = true
  auto_start_machines = true
  min_machines_running = 1

[[http_service.checks]]
  grace_period = "10s"
  interval = "30s"
  method = "GET"
  path = "/health"
  protocol = "http"
  timeout = "5s"

[[vm]]
  memory = "512mb"
  cpu_kind = "shared"
  cpus = 1
```

### Add PostgreSQL

```bash
fly postgres create --name my-app-db
fly postgres attach my-app-db
```

### Secrets

```bash
fly secrets set JWT_SECRET=$(openssl rand -hex 32)
fly secrets set SENTRY_DSN=https://xxx@sentry.io/xxx
fly secrets list
```

### Multi-Region

```bash
# Scale to multiple regions
fly scale count 2 --region iad,lhr,nrt

# Check status
fly status
```

### Fly.io Quick Reference

| Command | Purpose |
|---------|---------|
| `fly deploy` | Deploy |
| `fly status` | App status |
| `fly logs` | Live logs |
| `fly ssh console` | SSH into machine |
| `fly scale count N` | Scale machines |
| `fly scale vm shared-cpu-2x` | Change VM size |
| `fly secrets set KEY=val` | Set secret |
| `fly postgres connect` | PostgreSQL shell |
| `fly proxy 5432 -a my-app-db` | Tunnel to DB |

---

## Quick-Start Checklists

### Any Platform Checklist

- [ ] Application builds successfully locally
- [ ] Health check endpoint exists at `/health`
- [ ] App reads PORT from environment variable
- [ ] All secrets are in environment variables (not files)
- [ ] Database migrations run on deploy
- [ ] `.env` is in `.gitignore`
- [ ] `.env.example` exists with all variable names
- [ ] Error tracking configured (Sentry)
- [ ] Logging outputs to stdout (not files)
- [ ] Graceful shutdown handles SIGTERM

### Platform Cost Comparison (as of 2024)

| Platform | Free Tier | Starter | Production |
|----------|-----------|---------|------------|
| Heroku | None | $5/mo (Eco dyno) | $25/mo (Basic) |
| Vercel | 100GB BW | $20/mo/member | $20/mo/member |
| Cloudflare Workers | 100K req/day | $5/mo (10M req) | $5/mo base |
| AWS ECS | None | ~$30/mo (Fargate) | Varies |
| GCP Cloud Run | 2M req/mo | Pay per use | Pay per use |
| Railway | $5 free credit | $5/mo + usage | $20/mo + usage |
| Fly.io | 3 shared VMs | $1.94/mo per VM | Varies |
| Render | Static sites free | $7/mo (Instance) | $25/mo+ |

### Platform Selection Decision Tree

```
Is it a static site or SPA?
├── Yes → Vercel or Cloudflare Pages (free, fast, global CDN)
└── No
    ├── Is it Next.js?
    │   └── Yes → Vercel (first-party support)
    └── No
        ├── Do you need containers?
        │   ├── Yes, at scale → AWS ECS or GCP Cloud Run
        │   └── Yes, simple → Fly.io or Railway
        └── No
            ├── Is it serverless / event-driven?
            │   └── Yes → Cloudflare Workers or AWS Lambda
            └── No
                ├── Budget under $25/mo?
                │   └── Yes → Railway or Render
                └── No → Heroku (simplest) or AWS (most flexible)
```
