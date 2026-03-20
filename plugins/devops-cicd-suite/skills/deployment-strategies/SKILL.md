---
name: deployment-strategies
description: >
  Deployment strategies — blue-green, canary, rolling updates, feature flags,
  zero-downtime deployments, database migrations during deploys, and rollback
  procedures.
  Triggers: "deployment strategy", "blue green deploy", "canary deploy",
  "rolling update", "zero downtime", "feature flags", "deploy rollback",
  "database migration deploy", "deployment pipeline".
  NOT for: GitHub Actions syntax (use github-actions), monitoring (use monitoring-observability).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Deployment Strategies

## Blue-Green Deployment

Two identical environments. Switch traffic instantly between them.

```
┌─────────┐     ┌──────────────┐
│  Users   │────>│ Load Balancer │
└─────────┘     └──────┬───────┘
                       │
              ┌────────┴────────┐
              │                 │
        ┌─────▼─────┐   ┌──────▼────┐
        │   Blue    │   │   Green   │
        │ (current) │   │  (next)   │
        │  v1.2.3   │   │  v1.3.0   │
        └───────────┘   └───────────┘
```

```bash
# Heroku blue-green with pipelines
heroku pipelines:promote --app myapp-staging --to myapp-production

# Manual blue-green with Nginx
# 1. Deploy to green
ssh green-server "cd /app && git pull && npm ci && npm run build && pm2 restart all"

# 2. Health check green
curl -f https://green.internal/health || exit 1

# 3. Switch traffic
ssh lb-server "sed -i 's/blue-server/green-server/' /etc/nginx/conf.d/upstream.conf && nginx -s reload"

# 4. Verify production
curl -f https://myapp.com/health || {
  echo "Rollback!"
  ssh lb-server "sed -i 's/green-server/blue-server/' /etc/nginx/conf.d/upstream.conf && nginx -s reload"
  exit 1
}
```

```yaml
# Docker Compose blue-green
# docker-compose.yml
services:
  app-blue:
    image: myapp:current
    ports: ["3001:3000"]
    profiles: ["blue"]

  app-green:
    image: myapp:next
    ports: ["3002:3000"]
    profiles: ["green"]

  nginx:
    image: nginx:alpine
    ports: ["80:80"]
    volumes:
      - ./nginx.conf:/etc/nginx/conf.d/default.conf
    depends_on:
      - app-blue

# Switch: update nginx.conf upstream, reload
```

## Canary Deployment

Route a small percentage of traffic to the new version. Gradually increase.

```
Step 1: 5% → canary    (monitor errors, latency)
Step 2: 25% → canary   (monitor for 10 min)
Step 3: 50% → canary   (monitor for 30 min)
Step 4: 100% → canary  (full rollout)
```

```nginx
# Nginx canary with split_clients
upstream canary {
    server new-version:3000;
}
upstream stable {
    server current-version:3000;
}

split_clients "${remote_addr}" $upstream_variant {
    5%   canary;
    *    stable;
}

server {
    location / {
        proxy_pass http://$upstream_variant;
        # Add header so you can identify which version served the request
        add_header X-Canary $upstream_variant;
    }
}
```

```typescript
// Application-level canary with feature flags
function shouldUseCanary(userId: string): boolean {
  // Hash-based consistent routing
  const hash = createHash("sha256").update(userId).digest();
  const value = hash.readUInt32BE(0) % 100;
  return value < CANARY_PERCENTAGE; // e.g., 5
}

app.use((req, res, next) => {
  req.isCanary = shouldUseCanary(req.user?.id || req.ip);
  res.setHeader("X-Canary", req.isCanary ? "true" : "false");
  next();
});
```

## Rolling Update

Replace instances one at a time. Always some instances running.

```yaml
# Kubernetes rolling update
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
spec:
  replicas: 4
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1        # Max extra pods during update
      maxUnavailable: 1   # Max pods down during update
  template:
    spec:
      containers:
        - name: myapp
          image: myapp:v1.3.0
          readinessProbe:
            httpGet:
              path: /health
              port: 3000
            initialDelaySeconds: 5
            periodSeconds: 10
          livenessProbe:
            httpGet:
              path: /health
              port: 3000
            initialDelaySeconds: 15
            periodSeconds: 20
```

```bash
# PM2 rolling restart (Node.js)
pm2 reload ecosystem.config.js  # Graceful reload, zero downtime
pm2 restart ecosystem.config.js # Hard restart, brief downtime
```

## Zero-Downtime Database Migrations

```
Rule: Every migration must be backward-compatible with the PREVIOUS code version.

Deploy sequence:
1. Run migration (adds new column/table, doesn't remove old)
2. Deploy new code (reads from new, writes to both old and new)
3. Run data migration (backfill new column from old)
4. Deploy cleanup code (reads/writes only new)
5. Run cleanup migration (remove old column)
```

```sql
-- Step 1: Add new column (non-blocking)
ALTER TABLE users ADD COLUMN full_name TEXT;

-- Step 3: Backfill (do in batches for large tables)
UPDATE users SET full_name = first_name || ' ' || last_name
WHERE full_name IS NULL
LIMIT 1000;  -- Run in loop until no rows updated

-- Step 5: Remove old columns (after code no longer uses them)
ALTER TABLE users DROP COLUMN first_name;
ALTER TABLE users DROP COLUMN last_name;
```

```typescript
// Step 2: Dual-write code
async function updateUser(id: string, firstName: string, lastName: string) {
  await db.query(
    `UPDATE users SET
      first_name = $2,
      last_name = $3,
      full_name = $2 || ' ' || $3
    WHERE id = $1`,
    [id, firstName, lastName]
  );
}

// Step 4: Cleanup code (reads only new column)
async function getUser(id: string) {
  const user = await db.query("SELECT id, full_name FROM users WHERE id = $1", [id]);
  return user;
}
```

## Feature Flags

```typescript
// Simple feature flag system
interface FeatureFlags {
  [key: string]: {
    enabled: boolean;
    percentage?: number;  // Gradual rollout
    allowlist?: string[]; // Specific user IDs
  };
}

const flags: FeatureFlags = {
  new_checkout: { enabled: true, percentage: 10 },
  dark_mode: { enabled: true },
  ai_search: { enabled: false, allowlist: ["user-123", "user-456"] },
};

function isFeatureEnabled(flag: string, userId?: string): boolean {
  const f = flags[flag];
  if (!f || !f.enabled) return false;

  // Allowlist check
  if (f.allowlist && userId && f.allowlist.includes(userId)) return true;

  // Percentage rollout
  if (f.percentage !== undefined && userId) {
    const hash = createHash("md5").update(`${flag}:${userId}`).digest();
    return (hash.readUInt32BE(0) % 100) < f.percentage;
  }

  return f.enabled;
}

// Usage
app.get("/checkout", (req, res) => {
  if (isFeatureEnabled("new_checkout", req.user?.id)) {
    return renderNewCheckout(req, res);
  }
  return renderOldCheckout(req, res);
});

// Environment-based feature flags (simplest)
const FEATURES = {
  NEW_DASHBOARD: process.env.FEATURE_NEW_DASHBOARD === "true",
  AI_SEARCH: process.env.FEATURE_AI_SEARCH === "true",
};
```

## Rollback Procedures

```bash
# Heroku instant rollback
heroku rollback v42 --app myapp          # Rollback to specific release
heroku releases --app myapp              # List releases to find the right one

# Git-based rollback
git revert HEAD --no-edit && git push    # Revert last commit (creates new commit)

# Docker rollback
docker service update --image myapp:v1.2.3 myapp  # Previous version
docker compose up -d --force-recreate              # Recreate with old image

# Kubernetes rollback
kubectl rollout undo deployment/myapp              # Previous version
kubectl rollout undo deployment/myapp --to-revision=3  # Specific revision
kubectl rollout history deployment/myapp           # View history

# Database rollback (if needed)
npx prisma migrate reset   # DANGEROUS: drops and recreates
dbmate down                # Rollback last migration
```

## Health Checks

```typescript
// Express health check endpoint
app.get("/health", async (req, res) => {
  const checks = {
    status: "ok",
    timestamp: new Date().toISOString(),
    version: process.env.APP_VERSION || "unknown",
    uptime: process.uptime(),
    checks: {} as Record<string, { status: string; latency?: number }>,
  };

  // Database check
  try {
    const start = Date.now();
    await db.query("SELECT 1");
    checks.checks.database = { status: "ok", latency: Date.now() - start };
  } catch {
    checks.checks.database = { status: "error" };
    checks.status = "degraded";
  }

  // Redis check
  try {
    const start = Date.now();
    await redis.ping();
    checks.checks.redis = { status: "ok", latency: Date.now() - start };
  } catch {
    checks.checks.redis = { status: "error" };
    checks.status = "degraded";
  }

  const statusCode = checks.status === "ok" ? 200 : 503;
  res.status(statusCode).json(checks);
});

// Separate liveness vs readiness
app.get("/health/live", (req, res) => {
  // Am I running?
  res.status(200).json({ status: "alive" });
});

app.get("/health/ready", async (req, res) => {
  // Can I serve traffic?
  const dbOk = await checkDatabase();
  res.status(dbOk ? 200 : 503).json({ ready: dbOk });
});
```

## Gotchas

1. **Blue-green requires backward-compatible database schemas.** Both blue and green must work with the same database. Never deploy schema changes that break the old version. Use the expand-contract pattern (add, migrate, remove in separate deploys).

2. **Canary routing must be consistent per user.** If user A gets canary on request 1 but stable on request 2, they'll see inconsistent behavior. Use hash-based routing on user ID, not random selection.

3. **Feature flags are tech debt.** Every flag adds a code path. Clean up flags after full rollout. Set a TTL/review date when creating each flag. If a flag has been 100% for 30+ days, remove it.

4. **`pm2 reload` vs `pm2 restart`.** `reload` does zero-downtime rolling restart (waits for old processes to finish requests). `restart` kills all processes immediately. Always use `reload` in production.

5. **Health check endpoints shouldn't require auth.** Load balancers and orchestrators need to hit `/health` without tokens. Exclude health endpoints from auth middleware.

6. **Rolling updates with database migrations are tricky.** During a rolling update, OLD and NEW code versions run simultaneously. The database schema must work with both versions. Always make migrations additive (add columns, add tables) and clean up in a follow-up deploy.
