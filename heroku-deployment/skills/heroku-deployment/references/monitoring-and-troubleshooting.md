# Monitoring and Troubleshooting

Production monitoring, log analysis, performance tracking, and common issue resolution.

## Logging

### Live Log Streaming

```bash
# Stream all logs
heroku logs --tail

# Last N lines
heroku logs -n 200

# Only app output (your code's console.log, errors)
heroku logs --tail --source app

# Only Heroku platform messages (routing, dyno state changes)
heroku logs --tail --source heroku

# Filter by dyno type
heroku logs --tail --dyno web

# Filter by specific dyno
heroku logs --tail --dyno web.1
```

### Log Format

Heroku logs follow this format:
```
2026-02-20T18:00:00.000000+00:00 source[dyno]: message
```

- `source` — `app` (your code), `heroku` (platform), `api` (config changes)
- `dyno` — `web.1`, `worker.1`, `api`, `run.1234`

### Key Log Patterns

**Successful request:**
```
heroku[router]: at=info method=GET path="/" status=200 duration=45ms
```

**Error codes** (from Heroku platform):
```
heroku[router]: at=error code=H10 desc="App crashed"
heroku[router]: at=error code=H14 desc="No web dynos running"
heroku[web.1]: Error R14 (Memory quota exceeded)
```

### Persistent Logging

Heroku's built-in log buffer holds ~1500 lines. For persistent logs:

```bash
# Add Papertrail
heroku addons:create papertrail:choklad
heroku addons:open papertrail
```

Or drain logs to an external service:

```bash
heroku drains:add https://logs.example.com/heroku
heroku drains
```

## Process Monitoring

```bash
# Check running dynos and their state
heroku ps

# Detailed dyno info
heroku ps:type

# Restart all dynos
heroku ps:restart

# Restart a specific dyno
heroku ps:restart web.1
```

### Dyno States

| State | Meaning |
|-------|---------|
| `up` | Running normally |
| `starting` | Booting up |
| `crashed` | Process exited with error |
| `idle` | Free dyno sleeping (inactive >30 min) |
| `down` | Scaled to 0 |

## Release Tracking

```bash
# View release history
heroku releases

# View details of a specific release
heroku releases:info v42

# Compare what changed between releases
heroku releases:info v42
heroku releases:info v41
```

Each release corresponds to a deploy, config change, or rollback.

## App Information

```bash
# Full app info (region, stack, git URL, web URL)
heroku info

# App metrics (if using paid dynos)
heroku ps:type

# Check app's buildpacks
heroku buildpacks
```

## Performance Monitoring

### Heroku Metrics (Paid Dynos)

Paid dynos include basic metrics in the Heroku dashboard:
- Response time (p50, p95, p99)
- Throughput (requests/second)
- Memory usage
- CPU load

Access via `heroku open` → More → Metrics.

### Application-Level Monitoring

For deeper monitoring, add an APM addon:

```bash
# New Relic
heroku addons:create newrelic:wayne

# Scout APM
heroku addons:create scout:chair
```

## Rollbacks

```bash
# Roll back to previous release (instant, no rebuild)
heroku rollback

# Roll back to a specific version
heroku rollback v42

# Check what version you rolled back to
heroku releases
```

Rollbacks redeploy a previous slug. They don't revert config changes — only code.

## Troubleshooting

### First Steps — Always

1. **Check logs**: `heroku logs --tail` — find the actual error
2. **Check processes**: `heroku ps` — are dynos running?
3. **Check releases**: `heroku releases` — was there a recent bad deploy?
4. **Rollback if urgent**: `heroku rollback` — restore previous version instantly

### Build Failures

#### Missing devDependencies

**Symptom:** Build script fails — TypeScript, Vite, or other build tools not found.

**Cause:** Heroku defaults to `NODE_ENV=production` during install, which skips devDependencies.

**Fix:**
```bash
heroku config:set NPM_CONFIG_PRODUCTION=false
```

Or move build tools to `dependencies` instead of `devDependencies`.

#### Wrong Node.js Version

**Symptom:** Syntax errors or API not found during build.

**Cause:** Heroku's default Node.js version doesn't match your local version.

**Fix:** Set the engine in `package.json`:
```json
{
  "engines": {
    "node": ">=20.0.0"
  }
}
```

#### Out of Memory During Build

**Symptom:** `ENOMEM` or `JavaScript heap out of memory` during build.

**Fix:**
```bash
heroku config:set NODE_OPTIONS="--max-old-space-size=2048"
```

#### heroku-postbuild Script Error

**Symptom:** Non-zero exit code from `heroku-postbuild`.

**Cause:** One of the chained commands failed.

**Fix:** Break the chain into individual commands to identify which step fails:
```bash
heroku run bash
# Then run each command separately:
npm run db:generate
npm run build
npx prisma db push
```

### App Crashes

#### H10 — App Crashed

**Symptom:** App returns 503, logs show `H10`.

**Causes and fixes:**

1. **Missing Procfile** — Ensure `Procfile` (capital P, no extension) exists in the project root

2. **Wrong start command** — Verify Procfile command works locally. Check that the file path in the start command exists after build.

3. **Port binding error** — Your app must bind to `process.env.PORT`:
   ```javascript
   const PORT = process.env.PORT || 3000;
   app.listen(PORT);
   ```

4. **Uncaught exception on startup** — Check logs: `heroku logs --tail --source app`. Common causes: missing env var, failed database connection, import error.

5. **Missing dependency** — Check if the package is in `dependencies` (not just `devDependencies`).

#### H14 — No Web Dynos Running

**Symptom:** App returns 503, logs show `H14`.

**Cause:** No web dynos are scaled up.

**Fix:**
```bash
heroku ps:scale web=1
```

#### H20 — App Boot Timeout

**Symptom:** App fails to bind to port within 60 seconds.

**Fix:**
- Defer slow initialization (database connections, cache warming) to after the HTTP server starts listening
- Ensure `app.listen(PORT)` runs early, not after async operations

### Performance Issues

#### R14 — Memory Quota Exceeded

**Symptom:** App becomes slow or restarts, logs show `R14`.

**Fix:**
- Profile memory usage locally
- Check for memory leaks (unbounded caches, event listener leaks)
- Upgrade dyno: `heroku ps:type standard-1x`

#### R10 — Boot Timeout

**Symptom:** App restarts repeatedly, logs show `R10`.

**Fix:**
- Defer heavy initialization
- Use lazy loading for expensive imports
- Pre-build assets so startup doesn't compile anything

### Slug Size Issues

**Symptom:** `Slug size exceeds the limit` during deploy.

**Cause:** Built slug exceeds 500MB.

**Fix:** Create `.slugignore` in the project root:
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

### Database Issues

| Issue | Cause | Fix |
|-------|-------|-----|
| `DATABASE_URL` not set | Addon not provisioned | `heroku addons:create heroku-postgresql:essential-0` |
| `PrismaClientInitializationError` | Client not generated | Add `prisma generate` to `heroku-postbuild` |
| `too many connections` | Exceeding plan limit | Add `?connection_limit=5` to DATABASE_URL |
| `relation does not exist` | Tables not created | Run `heroku run npx prisma db push` |
| Connection refused | Using localhost URL | Verify `heroku config:get DATABASE_URL` is a Heroku URL |

### Domain and SSL

```bash
# Add custom domain
heroku domains:add www.example.com

# View DNS target (point your CNAME here)
heroku domains

# Enable automatic SSL (paid dynos)
heroku certs:auto:enable

# Check certificate status
heroku certs:auto
```

## Emergency Recovery

```bash
# Instant rollback (no rebuild)
heroku rollback

# Restart all dynos (clears stuck processes)
heroku ps:restart

# Maintenance mode (shows maintenance page while you fix)
heroku maintenance:on

# Run diagnostics in a one-off dyno
heroku run bash

# After fixing, turn off maintenance
heroku maintenance:off
```
