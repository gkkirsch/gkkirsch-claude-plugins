# Setup and Deployment

Deep-dive reference on Procfile patterns, build configuration, and deployment workflows for different app types.

## Procfile Format

The `Procfile` tells Heroku how to start your app. It lives in the project root with no file extension.

```
<process-type>: <command>
```

The `web` process type is special — it receives HTTP traffic and must bind to `$PORT`.

## Procfile Patterns

### Simple Node.js App

```
web: node server.js
```

Or using an npm script:

```
web: npm start
```

For production builds with a separate start script:

```
web: npm run start:prod
```

**When to use:** Single-file server, Express app, any app where the entry point runs directly.

### With a Build Step

Most apps need a build step (TypeScript compilation, bundling, etc.). Heroku runs `heroku-postbuild` or `build` scripts from `package.json` before starting the Procfile command.

**Procfile:**
```
web: node dist/server.js
```

**package.json:**
```json
{
  "scripts": {
    "build": "tsc",
    "heroku-postbuild": "npm run build",
    "start:prod": "node dist/server.js"
  }
}
```

**When to use:** TypeScript apps, apps with a compile/bundle step.

### heroku-postbuild vs build

- `heroku-postbuild` — Heroku-specific. Runs after `npm install`. Use this when your build needs Heroku-specific steps (e.g., Prisma generate, database push).
- `build` — Standard npm lifecycle. Heroku runs this automatically if `heroku-postbuild` doesn't exist.

**Rule of thumb:** Use `heroku-postbuild` when you need to do more than just build (database migrations, code generation). Use `build` for simple compile-only steps.

### Monorepo (npm Workspaces)

For monorepos with multiple packages/apps:

**Procfile:**
```
web: npm run start:prod
```

**package.json (root):**
```json
{
  "workspaces": ["apps/*", "packages/*"],
  "scripts": {
    "heroku-postbuild": "npm run db:generate && npm run build && npx prisma db push --schema=packages/db/prisma/schema.prisma",
    "build": "npm run build --workspace=@myapp/db && npm run build --workspace=@myapp/api && npm run build --workspace=@myapp/web",
    "start:prod": "node apps/api/dist/server.js"
  }
}
```

**Key details:**
- Build workspaces in dependency order (shared packages first, then apps)
- The Procfile starts only the web-serving process
- `heroku-postbuild` handles the full build pipeline including database operations

### Real-World Example: Supercharge Platform

The supercharge-claude-code platform runs this exact pattern:

```json
{
  "workspaces": ["apps/*", "packages/*"],
  "scripts": {
    "heroku-postbuild": "npm run db:generate && npm run build && npx prisma db push --schema=packages/db/prisma/schema.prisma",
    "build": "npm run build --workspace=@plugin-viewer/db && npm run build --workspace=@plugin-viewer/api && npm run build --workspace=@plugin-viewer/web",
    "start:prod": "node apps/api/dist/server.js"
  },
  "engines": { "node": ">=20.0.0" }
}
```

Procfile:
```
web: npm run start:prod
```

Build order: `prisma generate` → build db package → build api → build web → `prisma db push`.

### Multiple Process Types

For apps that need background workers alongside the web process:

```
web: node server.js
worker: node worker.js
```

Scale each independently:

```bash
heroku ps:scale web=1 worker=1
```

**When to use:** Background job processing, queue workers, scheduled tasks.

### Release Phase

Run one-off commands on every deploy (before the new version starts receiving traffic):

```
web: node server.js
release: npx prisma migrate deploy
```

The `release` process runs once per deploy. If it exits non-zero, the deploy is cancelled.

**When to use:** Database migrations, cache warming, notification webhooks.

**Caution:** Release commands have a time limit. For long-running operations, use `heroku run` instead.

### Static Sites (with a Server)

For serving a built frontend with a simple static server:

```
web: npx serve dist -l $PORT
```

Or with Express:

**Procfile:**
```
web: node server.js
```

**server.js:**
```javascript
const express = require('express');
const path = require('path');
const app = express();

app.use(express.static(path.join(__dirname, 'dist')));
app.get('*', (req, res) => res.sendFile(path.join(__dirname, 'dist', 'index.html')));

app.listen(process.env.PORT || 3000);
```

## Deployment Workflow

### Standard Deploy

```bash
# Commit your changes
git add .
git commit -m "deploy: feature update"

# Push to Heroku
git push heroku main
```

Heroku streams build output to your terminal. Watch for errors.

### Deploy a Non-Main Branch

```bash
# Deploy a feature branch to Heroku's main
git push heroku my-feature-branch:main
```

### Force Deploy (Overwrite History)

```bash
# Force push (use carefully — overwrites Heroku's git history)
git push heroku main --force
```

### Deploy from a Subdirectory

If your Heroku app lives in a monorepo subdirectory:

```bash
# Using git subtree
git subtree push --prefix apps/api heroku main
```

Or set the project root via buildpack:

```bash
heroku config:set PROJECT_PATH=apps/api
heroku buildpacks:set https://github.com/timanovsky/subdir-heroku-buildpack
heroku buildpacks:add heroku/nodejs
```

## Build Configuration

### Build Order on Heroku

1. `npm install` (installs dependencies from `package.json`)
2. `heroku-postbuild` script (if present)
3. If no `heroku-postbuild`, runs `build` script (if present)
4. Prunes devDependencies (unless `NPM_CONFIG_PRODUCTION=false`)
5. Caches `node_modules` for future builds
6. Creates a slug (compressed snapshot of your app)
7. Deploys the slug to dynos

### Environment Variables for Builds

```bash
# Keep devDependencies during install (needed for build tools)
heroku config:set NPM_CONFIG_PRODUCTION=false

# Set Node.js memory limit for builds
heroku config:set NODE_OPTIONS="--max-old-space-size=2048"

# Use specific npm version
heroku config:set NPM_CONFIG_ENGINE_STRICT=true
```

### Slug Size Optimization

Heroku slugs have a 500MB soft limit. If you're close:

Create `.slugignore` in the project root:
```
*.md
.git
.env
test/
tests/
docs/
.github/
*.test.ts
*.spec.ts
node_modules/.cache
```

### Build Cache

Heroku caches `node_modules` between deploys. To clear:

```bash
heroku config:set NODE_MODULES_CACHE=false
git push heroku main
heroku config:unset NODE_MODULES_CACHE
```

## Procfile Common Mistakes

| Mistake | Fix |
|---------|-----|
| Named `procfile` or `Procfile.txt` | Must be exactly `Procfile` (capital P, no extension) |
| Using `npm run dev` | Use production start command (no watch mode, no HMR) |
| Hardcoded port | Always use `process.env.PORT` |
| Missing `web:` process | `web` is required for HTTP traffic |
| Build output not in `start` path | Ensure `start:prod` points to the actual build output directory |
