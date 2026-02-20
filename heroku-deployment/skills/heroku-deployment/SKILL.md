---
name: heroku-deployment
description: >
  Use when deploying a Node.js application to Heroku. Covers the full deployment lifecycle:
  app creation, Procfile setup, environment configuration, database provisioning, deployment
  via git push, monitoring, and rollbacks. Use when: (1) user says "deploy to heroku",
  "heroku deploy", "push to heroku", "heroku setup", "create heroku app", (2) user is
  building a Node.js app and needs hosting, (3) user needs to troubleshoot a Heroku
  deployment, (4) user needs to set up Postgres on Heroku with Prisma.
version: 1.0.0
---

# Heroku Deployment

Deploy Node.js applications to Heroku. This skill covers the complete lifecycle from app creation through production monitoring.

For deep-dive references:
- Procfile patterns and deployment workflows → `references/setup-and-deployment.md`
- Postgres, Prisma, and other addons → `references/database-and-addons.md`
- Monitoring, logs, and troubleshooting → `references/monitoring-and-troubleshooting.md`

## Prerequisites

- **Heroku CLI**: `brew tap heroku/brew && brew install heroku` (or `npm install -g heroku`)
- **Git**: initialized in the project directory
- **Node.js**: 18+ with a `package.json`
- **Heroku account**: `heroku login` to authenticate

Verify setup:

```bash
heroku --version
heroku auth:whoami
```

## 1. Create the App

```bash
# Create with auto-generated name
heroku create

# Create with a specific name
heroku create my-app-name

# Create in a specific region
heroku create my-app-name --region eu
```

This adds a `heroku` git remote automatically. Verify:

```bash
git remote -v
# heroku  https://git.heroku.com/my-app-name.git (fetch)
# heroku  https://git.heroku.com/my-app-name.git (push)
```

If you need to add the remote manually (e.g., existing app):

```bash
heroku git:remote -a my-app-name
```

## 2. Set Up the Procfile

Create a `Procfile` in the project root (no extension):

```
web: npm run start:prod
```

The `web` process type is required — it receives HTTP traffic and binds to `$PORT`.

**Your start command must:**
- Run the production server (not dev/watch mode)
- Bind to `process.env.PORT` (Heroku assigns the port dynamically)

```javascript
// In your server code
const PORT = process.env.PORT || 3000;
app.listen(PORT, () => console.log(`Listening on ${PORT}`));
```

For more Procfile patterns (monorepos, build steps, workers), see `references/setup-and-deployment.md`.

## 3. Configure Environment Variables

```bash
# Set a single variable
heroku config:set NODE_ENV=production

# Set multiple at once
heroku config:set NODE_ENV=production SECRET_KEY=abc123 API_URL=https://api.example.com

# View all config
heroku config

# View a single variable
heroku config:get DATABASE_URL

# Remove a variable
heroku config:unset DEBUG
```

**Common variables to set:**
- `NODE_ENV=production` — enables production optimizations
- `NPM_CONFIG_PRODUCTION=false` — install devDependencies (needed if build tools are in devDeps)
- API keys, secrets, and external service URLs

**Never commit secrets to git.** Use `heroku config:set` for all sensitive values.

### Node.js Version

Heroku uses the `engines` field in `package.json`:

```json
{
  "engines": {
    "node": ">=20.0.0"
  }
}
```

## 4. Build Configuration

Heroku runs `npm install` automatically, then looks for a `heroku-postbuild` script:

```json
{
  "scripts": {
    "heroku-postbuild": "npm run build",
    "build": "tsc && vite build",
    "start:prod": "node dist/server.js"
  }
}
```

**Build order on Heroku:**
1. `npm install` (installs dependencies)
2. `heroku-postbuild` script (if present — use this for build steps)
3. `Procfile` command starts the app

If you don't have a `heroku-postbuild` script, Heroku runs the `build` script automatically.

### Real-World Example: Supercharge Platform

The [supercharge-claude-code](https://superchargeclaudecode.com) platform uses a monorepo with Prisma:

```json
{
  "scripts": {
    "heroku-postbuild": "npm run db:generate && npm run build && npx prisma db push --schema=packages/db/prisma/schema.prisma",
    "build": "npm run build --workspace=@plugin-viewer/db && npm run build --workspace=@plugin-viewer/api && npm run build --workspace=@plugin-viewer/web",
    "start:prod": "node apps/api/dist/server.js"
  },
  "engines": {
    "node": ">=20.0.0"
  }
}
```

Procfile:
```
web: npm run start:prod
```

This pattern: `prisma generate → build all workspaces → prisma db push` handles the full monorepo + database workflow in a single deploy.

## 5. Database Setup (Postgres)

If your app needs a database:

```bash
# Add Postgres addon (essential-0 is the cheapest paid plan)
heroku addons:create heroku-postgresql:essential-0

# Check database status
heroku pg:info
```

This automatically sets `DATABASE_URL` in your config. No manual configuration needed.

For Prisma setup, migrations, backups, and production access, see `references/database-and-addons.md`.

## 6. Deploy

```bash
# Deploy main branch
git push heroku main

# Deploy a different branch to heroku's main
git push heroku my-branch:main
```

Watch the build output — Heroku streams it to your terminal.

### First Deploy Checklist

Before your first `git push heroku main`:

- [ ] `Procfile` exists in project root
- [ ] Start command binds to `process.env.PORT`
- [ ] `engines.node` set in `package.json`
- [ ] `heroku-postbuild` or `build` script defined (if app needs building)
- [ ] Environment variables configured via `heroku config:set`
- [ ] Database addon provisioned (if needed)

## 7. Verify and Monitor

```bash
# Open the app in browser
heroku open

# View logs (live stream)
heroku logs --tail

# View recent logs
heroku logs -n 200

# Check running processes
heroku ps

# Check release history
heroku releases

# Show app info
heroku info
```

### Log Filtering

```bash
# Only app logs (your code's output)
heroku logs --tail --source app

# Only Heroku router logs (request info)
heroku logs --tail --source heroku

# Filter by dyno type
heroku logs --tail --dyno web
```

## 8. Scaling

```bash
# Check current dynos
heroku ps

# Scale web dynos
heroku ps:scale web=1

# Scale to zero (stop the app)
heroku ps:scale web=0

# Add a worker process
heroku ps:scale worker=1
```

## 9. Rollbacks

If a deploy breaks something:

```bash
# View recent releases
heroku releases

# Roll back to previous release
heroku rollback

# Roll back to a specific release
heroku rollback v42
```

Rollbacks are instant — Heroku re-deploys the previous slug without rebuilding.

## 10. Maintenance

```bash
# Enable maintenance mode (shows maintenance page)
heroku maintenance:on

# Disable maintenance mode
heroku maintenance:off

# Restart all dynos
heroku ps:restart

# Run a one-off command
heroku run bash
heroku run node scripts/seed.js
```

## 11. Pipelines (Staging → Production)

For apps that need a staging environment before production, Heroku Pipelines connect multiple apps into a promotion workflow.

```bash
# Create a pipeline
heroku pipelines:create my-pipeline -a my-app-production --stage production

# Add a staging app
heroku pipelines:add my-pipeline -a my-app-staging --stage staging

# View pipeline
heroku pipelines:info my-pipeline

# Promote staging to production (no rebuild — moves the slug)
heroku pipelines:promote -a my-app-staging
```

**Pipeline flow:**
1. Push code to staging: `git push heroku main` (with staging app's remote)
2. Test on staging
3. Promote to production: `heroku pipelines:promote` (instant, no rebuild)

### Review Apps

Pipelines can automatically create temporary apps for pull requests:

```bash
# Enable review apps (configure in Heroku Dashboard → Pipeline → Enable Review Apps)
# Requires GitHub integration
```

Review apps spin up for each PR and are destroyed when the PR is closed.

**When to use pipelines:** Production apps that need testing before deploy, teams with multiple developers, apps requiring approval workflows.

**For most solo projects and simple deployments, direct `git push heroku main` is sufficient.** Pipelines add value when you need staging environments or team review workflows.

## Quick Reference

| Task | Command |
|------|---------|
| Create app | `heroku create my-app` |
| Add git remote | `heroku git:remote -a my-app` |
| Set env var | `heroku config:set KEY=value` |
| View env vars | `heroku config` |
| Deploy | `git push heroku main` |
| View logs | `heroku logs --tail` |
| Check status | `heroku ps` |
| Release history | `heroku releases` |
| Rollback | `heroku rollback` |
| Add Postgres | `heroku addons:create heroku-postgresql:essential-0` |
| DB console | `heroku pg:psql` |
| Run command | `heroku run <command>` |
| Open app | `heroku open` |
| Restart | `heroku ps:restart` |
| Scale | `heroku ps:scale web=1` |
| Maintenance on | `heroku maintenance:on` |

## Troubleshooting

For common issues (H10 crashes, build failures, database errors), see `references/monitoring-and-troubleshooting.md`.

Quick checks when something goes wrong:

```bash
# Check logs first — always
heroku logs --tail

# Check if dynos are running
heroku ps

# Check recent releases for bad deploy
heroku releases

# Rollback if latest deploy is broken
heroku rollback
```
