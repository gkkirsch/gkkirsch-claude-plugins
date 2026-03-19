---
name: deploy
description: >
  Quick deployment command — analyzes your project, recommends the best deployment target,
  generates all necessary configs (Dockerfile, CI/CD pipeline, platform config), and
  walks you through deployment step by step.
  Triggers: "/deploy", "deploy my app", "deploy to production", "set up deployment".
user-invocable: true
argument-hint: "[platform] [--skip-docker] [--skip-ci]"
allowed-tools: Read, Write, Edit, Bash, Glob, Grep
model: sonnet
---

# /deploy Command

One-command deployment setup. Analyzes your project, generates Docker configs, CI/CD pipelines, and platform-specific deployment configs.

## Usage

```
/deploy                    # Auto-detect best platform
/deploy heroku             # Target specific platform
/deploy vercel --skip-ci   # Skip CI/CD pipeline generation
/deploy aws --skip-docker  # Skip Docker config (e.g., for Lambda)
```

## Supported Platforms

| Platform | Best For |
|----------|----------|
| Heroku | Quick deploys, small-medium apps |
| Vercel | Frontend, Next.js, serverless |
| Cloudflare | Edge apps, Workers, static sites |
| AWS ECS | Production containers at scale |
| AWS Lambda | Event-driven, serverless functions |
| GCP Cloud Run | Containerized serverless |
| Railway | Developer-friendly PaaS |
| Fly.io | Global edge deployment |
| Render | Simple container hosting |

## Procedure

### Step 1: Detect Project

Read the project root to understand the stack:

1. **Read** `package.json`, `requirements.txt`, `Cargo.toml`, `go.mod`, `Gemfile`, `pom.xml` — determine language/framework
2. **Read** existing deployment files: `Dockerfile`, `docker-compose.yml`, `.github/workflows/`, `Procfile`, `vercel.json`, `fly.toml`, `railway.json`
3. **Glob** for entry points: `src/index.*`, `app.*`, `main.*`, `server.*`
4. **Detect framework**: Express, Next.js, FastAPI, Django, Rails, Spring Boot, etc.
5. **Check for databases**: Look for ORM configs, migration folders, connection strings in env files

Report findings:

```
Detected:
- Language: TypeScript (Node.js 20)
- Framework: Express + React (Vite)
- Database: PostgreSQL (Drizzle ORM)
- Existing configs: None
```

### Step 2: Recommend Platform

Based on detection, recommend the best platform. Use this decision tree:

- **Static site / SPA only** → Vercel or Cloudflare Pages
- **Next.js** → Vercel
- **Serverless functions** → AWS Lambda or Cloudflare Workers
- **Simple API + frontend** → Heroku or Railway
- **Containers needed** → AWS ECS or GCP Cloud Run
- **Global edge** → Fly.io or Cloudflare Workers
- **Budget-conscious** → Railway or Render

Present the recommendation with reasoning. Ask the user to confirm or pick a different platform.

### Step 3: Generate Docker Config (unless --skip-docker)

Dispatch the `dockerfile-builder` agent:

```
Task tool:
  subagent_type: "dockerfile-builder"
  mode: "bypassPermissions"
  prompt: |
    Create a production-ready Dockerfile for this project.
    Stack: [detected stack]
    Entry point: [detected entry point]
    Build command: [from package.json scripts]
    Also create .dockerignore.
    Write files to the project root.
```

### Step 4: Generate CI/CD Pipeline (unless --skip-ci)

Dispatch the `ci-cd-architect` agent:

```
Task tool:
  subagent_type: "ci-cd-architect"
  mode: "bypassPermissions"
  prompt: |
    Create a CI/CD pipeline for this project.
    Stack: [detected stack]
    Target platform: [chosen platform]
    Include: lint, test, build, deploy stages.
    Write to .github/workflows/deploy.yml
```

### Step 5: Generate Platform Config

Based on the chosen platform, generate the platform-specific config:

**Heroku**: `Procfile`, `app.json`, `heroku.yml`
**Vercel**: `vercel.json`
**Cloudflare**: `wrangler.toml`
**AWS ECS**: Task definition JSON, service config
**Railway**: `railway.json`, `nixpacks.toml`
**Fly.io**: `fly.toml`
**Render**: `render.yaml`

Write the config files to the project root.

### Step 6: Environment Variables

1. **Grep** for `process.env.`, `os.environ`, `env::var` to find all env vars used
2. List them out with descriptions
3. Generate a `.env.example` with all variables (no values)
4. Provide platform-specific instructions for setting env vars:
   - Heroku: `heroku config:set KEY=value`
   - Vercel: `vercel env add KEY`
   - AWS: Parameter Store or Secrets Manager commands

### Step 7: Deployment Checklist

Print a deployment checklist:

```markdown
## Pre-Deployment Checklist

- [ ] Environment variables configured on [platform]
- [ ] Database provisioned and connection string set
- [ ] Domain configured (if applicable)
- [ ] SSL certificate active
- [ ] CI/CD pipeline tested (push to branch, verify it runs)
- [ ] Health check endpoint responds at /health
- [ ] Error tracking configured (Sentry recommended)
- [ ] Logging configured for production

## Deploy Command

[platform-specific deploy command]

## Post-Deployment

- [ ] Verify app loads at production URL
- [ ] Run smoke tests
- [ ] Check logs for errors
- [ ] Monitor for 15 minutes
```

### Step 8: Execute Deployment (if confirmed)

If the user confirms, execute the deployment:

- **Heroku**: `git push heroku main` or `heroku container:push web`
- **Vercel**: `vercel --prod`
- **Cloudflare**: `wrangler deploy`
- **Railway**: `railway up`
- **Fly.io**: `fly deploy`

Watch for errors and help troubleshoot.

## Error Recovery

Common deployment errors and fixes:

| Error | Cause | Fix |
|-------|-------|-----|
| Port binding | App not using PORT env var | Add `const port = process.env.PORT \|\| 3000` |
| Build failure | Missing dependencies | Check devDependencies vs dependencies |
| Health check fail | No /health endpoint | Add health check route |
| OOM | Memory limit too low | Increase dyno/container memory |
| Cold start timeout | Slow initialization | Optimize imports, lazy load |

## Notes

- Never commit secrets or .env files
- Always use environment variables for configuration
- Test locally with `docker-compose up` before deploying
- Set up monitoring before going live
