# Database and Addons

Heroku addons extend your app with managed services. This reference covers Postgres (the most common), Prisma integration, and other popular addons.

## Heroku Postgres

### Provisioning

```bash
# Add Postgres (essential-0 is the cheapest paid plan, ~$5/mo)
heroku addons:create heroku-postgresql:essential-0

# Check database info
heroku pg:info

# View the connection string
heroku config:get DATABASE_URL
```

Heroku automatically sets `DATABASE_URL` in your app's environment. No manual configuration needed.

### Plan Options

| Plan | Price | Rows | Connections | Use Case |
|------|-------|------|-------------|----------|
| essential-0 | ~$5/mo | 10K | 20 | Development, small projects |
| essential-1 | ~$9/mo | 10M | 40 | Small production apps |
| essential-2 | ~$15/mo | 10M | 80 | Growing apps |
| standard-0 | ~$50/mo | Unlimited | 120 | Production |

Start with `essential-0` and upgrade as needed:

```bash
heroku addons:upgrade heroku-postgresql:essential-1
```

### Production Database Access

```bash
# Interactive psql console
heroku pg:psql

# Run a SQL query directly
heroku pg:psql -c "SELECT count(*) FROM users"

# Detailed database status
heroku pg:info

# Active queries / connection stats
heroku pg:ps

# Table sizes
heroku pg:psql -c "SELECT relname, pg_size_pretty(pg_total_relation_size(relid)) FROM pg_catalog.pg_statio_user_tables ORDER BY pg_total_relation_size(relid) DESC LIMIT 10"
```

### Backups

```bash
# Create a manual backup
heroku pg:backups:capture

# List backups
heroku pg:backups

# Download latest backup
heroku pg:backups:download

# Restore from a backup
heroku pg:backups:restore b001 DATABASE_URL

# Schedule automatic backups (daily at 2am UTC)
heroku pg:backups:schedule DATABASE_URL --at '02:00 UTC'
```

**Always capture a backup before deploying schema changes.**

### Connection Management

Heroku Postgres has connection limits per plan. Set a limit via URL parameter:

```bash
heroku config:set DATABASE_URL="postgresql://...?connection_limit=10&pool_timeout=30"
```

## Prisma Integration

### Schema Configuration

In your `prisma/schema.prisma`:

```prisma
datasource db {
  provider = "postgresql"
  url      = env("DATABASE_URL")
}

generator client {
  provider = "prisma-client-js"
}

model User {
  id    Int     @id @default(autoincrement())
  email String  @unique
  name  String?
}
```

The `env("DATABASE_URL")` reads the Heroku-provided connection string automatically.

### Local Development

Create a `.env` file (never commit this):

```
DATABASE_URL="postgresql://user:password@localhost:5432/mydb"
```

### Build Pipeline

Add `heroku-postbuild` to run Prisma operations during deploy:

```json
{
  "scripts": {
    "db:generate": "prisma generate",
    "heroku-postbuild": "npm run db:generate && npm run build && npx prisma db push",
    "build": "tsc",
    "start:prod": "node dist/server.js"
  }
}
```

**Build order:**
1. `prisma generate` — generates the Prisma Client from your schema
2. `build` — compiles your app (TypeScript, etc.)
3. `prisma db push` — syncs the schema to the database

### Monorepo Prisma Pattern

When Prisma schema lives in a separate package (e.g., `packages/db/prisma/schema.prisma`):

```json
{
  "scripts": {
    "db:generate": "npm run generate --workspace=@myapp/db",
    "heroku-postbuild": "npm run db:generate && npm run build && npx prisma db push --schema=packages/db/prisma/schema.prisma"
  }
}
```

**Note the `--schema` flag** — Prisma needs the explicit path when the schema isn't at the default location.

### Real-World Example: Supercharge Platform

```json
{
  "scripts": {
    "heroku-postbuild": "npm run db:generate && npm run build && npx prisma db push --schema=packages/db/prisma/schema.prisma",
    "db:generate": "npm run generate --workspace=@plugin-viewer/db"
  }
}
```

### db push vs migrate deploy

**`prisma db push`** — Pushes schema changes directly. No migration history. Good for prototyping and simple deployments.

```bash
npx prisma db push
```

**`prisma migrate deploy`** — Applies pending migrations from `prisma/migrations/`. Maintains migration history. Better for teams and production with breaking changes.

```bash
npx prisma migrate deploy
```

**Recommendation:** Use `db push` for simple apps and solo projects. Use `migrate deploy` (in a release phase) when you need migration history or have breaking schema changes.

```
# Using migrate deploy in Procfile release phase
release: npx prisma migrate deploy
web: node dist/server.js
```

## Other Popular Addons

### Heroku Redis

```bash
# Add Redis
heroku addons:create heroku-redis:mini

# View connection info
heroku config:get REDIS_URL

# Connect to Redis CLI
heroku redis:cli
```

Use `REDIS_URL` in your app for caching, sessions, or job queues.

### Logging (Papertrail)

```bash
# Add Papertrail for log aggregation
heroku addons:create papertrail:choklad

# Open Papertrail dashboard
heroku addons:open papertrail
```

Useful for searching historical logs beyond Heroku's 1500-line log buffer.

### Scheduler

```bash
# Add the scheduler
heroku addons:create scheduler:standard

# Open scheduler dashboard to configure jobs
heroku addons:open scheduler
```

Run recurring tasks (e.g., cleanup scripts, reports) on a schedule.

### SendGrid (Email)

```bash
# Add SendGrid for transactional email
heroku addons:create sendgrid:starter

# View API key
heroku config:get SENDGRID_API_KEY
```

### Managing Addons

```bash
# List all addons on your app
heroku addons

# Get addon info
heroku addons:info heroku-postgresql

# Remove an addon (WARNING: destroys data)
heroku addons:destroy heroku-postgresql
```

## Common Database Issues

| Issue | Cause | Fix |
|-------|-------|-----|
| `DATABASE_URL` not set | Addon not provisioned | `heroku addons:create heroku-postgresql:essential-0` |
| `PrismaClientInitializationError` | Client not generated | Add `prisma generate` to `heroku-postbuild` |
| `too many connections` | Exceeding plan limit | Add `?connection_limit=5` to DATABASE_URL |
| Schema out of sync | Forgot db push/migrate | Add `prisma db push` to `heroku-postbuild` |
| `relation does not exist` | Tables not created | Run `heroku run npx prisma db push` |
| SSL connection error | Missing SSL config | Heroku handles SSL automatically via DATABASE_URL |
