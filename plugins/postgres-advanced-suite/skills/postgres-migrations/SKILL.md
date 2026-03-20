---
name: postgres-migrations
description: >
  PostgreSQL migration strategies — Prisma Migrate, raw SQL migrations with dbmate,
  data backfills on large tables, zero-downtime patterns, rollback strategies, and CI/CD.
  Triggers: "database migration", "prisma migrate", "schema change", "add column",
  "alter table", "data migration", "dbmate", "rollback migration".
  NOT for: query performance (use postgres-performance), index strategy (use postgres-indexing).
version: 1.0.0
allowed-tools: Read, Grep, Glob, Bash, Edit, Write
---

# PostgreSQL Migrations

## Migration Tool Comparison

| Tool | Language | Style | Rollback | Best For |
|------|----------|-------|----------|----------|
| **Prisma Migrate** | TypeScript | Schema-first (declarative) | Manual | Full-stack TypeScript apps |
| **dbmate** | Any | SQL files (imperative) | Built-in | Polyglot teams, raw SQL control |
| **golang-migrate** | Go/Any | SQL files | Built-in | Go projects |
| **Flyway** | Java/Any | SQL + Java | Paid feature | Enterprise Java |
| **Knex** | JavaScript | JS/TS builder | Built-in | Express/Node apps |
| **TypeORM** | TypeScript | Auto-generated | Built-in | TypeORM projects |
| **Drizzle Kit** | TypeScript | Schema-first | Manual | Drizzle ORM projects |

## Prisma Migrate

### Setup

```bash
npm install prisma @prisma/client
npx prisma init
```

```prisma
// prisma/schema.prisma
datasource db {
  provider = "postgresql"
  url      = env("DATABASE_URL")
}

generator client {
  provider = "prisma-client-js"
}

model User {
  id        String   @id @default(cuid())
  email     String   @unique
  name      String?
  posts     Post[]
  createdAt DateTime @default(now()) @map("created_at")
  updatedAt DateTime @updatedAt @map("updated_at")

  @@map("users")
}

model Post {
  id        String   @id @default(cuid())
  title     String
  content   String?
  published Boolean  @default(false)
  author    User     @relation(fields: [authorId], references: [id])
  authorId  String   @map("author_id")
  createdAt DateTime @default(now()) @map("created_at")

  @@index([authorId])
  @@map("posts")
}
```

### Workflow

```bash
# Create migration from schema changes
npx prisma migrate dev --name add_posts_table

# Apply migrations in production
npx prisma migrate deploy

# Reset database (WARNING: drops all data)
npx prisma migrate reset

# Check migration status
npx prisma migrate status

# Generate client without migrating
npx prisma generate
```

### Migration Files

```
prisma/migrations/
  20250101120000_init/
    migration.sql
  20250102130000_add_posts_table/
    migration.sql
```

Each `migration.sql` is raw SQL that Prisma generated:

```sql
-- CreateTable
CREATE TABLE "posts" (
    "id" TEXT NOT NULL,
    "title" TEXT NOT NULL,
    "content" TEXT,
    "published" BOOLEAN NOT NULL DEFAULT false,
    "author_id" TEXT NOT NULL,
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "posts_pkey" PRIMARY KEY ("id")
);

-- CreateIndex
CREATE INDEX "posts_author_id_idx" ON "posts"("author_id");

-- AddForeignKey
ALTER TABLE "posts" ADD CONSTRAINT "posts_author_id_fkey"
  FOREIGN KEY ("author_id") REFERENCES "users"("id") ON DELETE RESTRICT ON UPDATE CASCADE;
```

### Editing Migrations Before Applying

You can edit the SQL before running `migrate dev`:

```bash
# Create migration without applying
npx prisma migrate dev --name add_status_column --create-only

# Edit the generated SQL file
# e.g., add a default value, backfill, or use CONCURRENTLY

# Then apply
npx prisma migrate dev
```

### Custom SQL in Prisma Migrations

When Prisma can't express what you need:

```sql
-- prisma/migrations/20250301_add_search/migration.sql

-- Prisma generated this:
ALTER TABLE "posts" ADD COLUMN "search_vector" tsvector;

-- You add this manually:
CREATE INDEX CONCURRENTLY idx_posts_search ON posts USING gin (search_vector);

UPDATE posts SET search_vector =
  setweight(to_tsvector('english', coalesce(title, '')), 'A') ||
  setweight(to_tsvector('english', coalesce(content, '')), 'B');

CREATE OR REPLACE FUNCTION posts_search_trigger() RETURNS trigger AS $$
BEGIN
  NEW.search_vector :=
    setweight(to_tsvector('english', coalesce(NEW.title, '')), 'A') ||
    setweight(to_tsvector('english', coalesce(NEW.content, '')), 'B');
  RETURN NEW;
END
$$ LANGUAGE plpgsql;

CREATE TRIGGER posts_search_update
  BEFORE INSERT OR UPDATE ON posts
  FOR EACH ROW EXECUTE FUNCTION posts_search_trigger();
```

## dbmate (Raw SQL Migrations)

### Setup

```bash
# Install
brew install dbmate
# Or: npm install -g dbmate

# Configure
export DATABASE_URL="postgres://user:pass@localhost:5432/mydb?sslmode=disable"
```

### Workflow

```bash
# Create migration
dbmate new add_users_table

# Apply pending migrations
dbmate up

# Rollback last migration
dbmate rollback

# Check status
dbmate status

# Dump schema
dbmate dump
```

### Migration Files

```sql
-- db/migrations/20250301120000_add_users_table.sql

-- migrate:up
CREATE TABLE users (
  id SERIAL PRIMARY KEY,
  email TEXT NOT NULL UNIQUE,
  name TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_users_email ON users (email);

-- migrate:down
DROP TABLE IF EXISTS users;
```

### Multi-Statement Migration

```sql
-- migrate:up
BEGIN;

ALTER TABLE users ADD COLUMN role TEXT NOT NULL DEFAULT 'user';
ALTER TABLE users ADD COLUMN last_login TIMESTAMPTZ;

CREATE INDEX CONCURRENTLY idx_users_role ON users (role);
-- Note: CONCURRENTLY can't run inside a transaction block
-- Move it outside BEGIN/COMMIT or run separately

COMMIT;

-- migrate:down
BEGIN;

DROP INDEX IF EXISTS idx_users_role;
ALTER TABLE users DROP COLUMN IF EXISTS last_login;
ALTER TABLE users DROP COLUMN IF EXISTS role;

COMMIT;
```

## Zero-Downtime Migration Patterns

### Safe Operations (No Lock Issues)

| Operation | Lock Level | Duration | Safe? |
|-----------|-----------|----------|-------|
| `CREATE TABLE` | None on existing | Instant | Yes |
| `CREATE INDEX CONCURRENTLY` | ShareUpdateExclusive | Slow but non-blocking | Yes |
| `ADD COLUMN` (nullable, no default) | AccessExclusive | Instant (PG 11+) | Yes |
| `ADD COLUMN ... DEFAULT x` | AccessExclusive | Instant (PG 11+) | Yes |
| `DROP COLUMN` | AccessExclusive | Instant | Yes* |
| `RENAME COLUMN` | AccessExclusive | Instant | Risky** |

*Safe from lock perspective, but app must stop reading the column first.
**Instant but breaks all queries referencing the old name.

### Dangerous Operations

| Operation | Problem | Safe Alternative |
|-----------|---------|-----------------|
| `ADD COLUMN NOT NULL` (no default) | Rewrites table (PG < 11) | Add nullable + backfill + constraint |
| `ALTER COLUMN TYPE` | Rewrites table, full lock | New column + backfill + swap |
| `ADD CONSTRAINT ... CHECK` | Full table scan with lock | `NOT VALID` + `VALIDATE` separately |
| `CREATE INDEX` (not CONCURRENTLY) | Blocks all writes | Always use `CONCURRENTLY` |
| `ALTER TABLE ... SET NOT NULL` | Full table scan | Use CHECK constraint instead |

### Pattern: Add NOT NULL Column Safely

```sql
-- Step 1: Add nullable column with default (instant in PG 11+)
ALTER TABLE users ADD COLUMN status TEXT DEFAULT 'active';

-- Step 2: Backfill in batches (non-blocking)
UPDATE users SET status = 'active'
WHERE id IN (
  SELECT id FROM users WHERE status IS NULL LIMIT 10000
);
-- Repeat until all rows are updated

-- Step 3: Add NOT NULL constraint without full scan
ALTER TABLE users ADD CONSTRAINT users_status_not_null
  CHECK (status IS NOT NULL) NOT VALID;

-- Step 4: Validate constraint (scans but doesn't block writes)
ALTER TABLE users VALIDATE CONSTRAINT users_status_not_null;
```

### Pattern: Change Column Type Safely

```sql
-- Step 1: Add new column
ALTER TABLE orders ADD COLUMN amount_v2 BIGINT;

-- Step 2: Dual-write (in application code)
-- UPDATE both amount and amount_v2 in all writes

-- Step 3: Backfill existing data
UPDATE orders SET amount_v2 = amount::bigint
WHERE id IN (SELECT id FROM orders WHERE amount_v2 IS NULL LIMIT 10000);

-- Step 4: Verify all rows are backfilled
SELECT count(*) FROM orders WHERE amount_v2 IS NULL;

-- Step 5: Swap columns (in a single migration)
ALTER TABLE orders RENAME COLUMN amount TO amount_old;
ALTER TABLE orders RENAME COLUMN amount_v2 TO amount;

-- Step 6: Drop old column (after verifying app works)
ALTER TABLE orders DROP COLUMN amount_old;
```

### Pattern: Rename Column Safely

```sql
-- Never rename directly — breaks all queries instantly

-- Step 1: Add new column
ALTER TABLE users ADD COLUMN full_name TEXT;

-- Step 2: Backfill
UPDATE users SET full_name = name WHERE full_name IS NULL;

-- Step 3: Application reads from BOTH columns
-- SELECT COALESCE(full_name, name) AS name FROM users

-- Step 4: Application writes to BOTH columns
-- UPDATE users SET name = ?, full_name = ? WHERE id = ?

-- Step 5: Stop reading old column in application

-- Step 6: Drop old column
ALTER TABLE users DROP COLUMN name;
```

## Data Backfills

### Batch Update Pattern

```sql
-- Don't do this (locks entire table):
UPDATE users SET status = 'active' WHERE status IS NULL;

-- Do this (batch updates):
DO $$
DECLARE
  batch_size INT := 10000;
  affected INT;
BEGIN
  LOOP
    UPDATE users SET status = 'active'
    WHERE id IN (
      SELECT id FROM users
      WHERE status IS NULL
      LIMIT batch_size
      FOR UPDATE SKIP LOCKED
    );

    GET DIAGNOSTICS affected = ROW_COUNT;
    RAISE NOTICE 'Updated % rows', affected;

    EXIT WHEN affected = 0;

    PERFORM pg_sleep(0.1);  -- Brief pause to reduce load
    COMMIT;
  END LOOP;
END $$;
```

### Backfill with Progress Tracking

```typescript
async function backfillInBatches(
  db: Pool,
  batchSize: number = 5000
): Promise<void> {
  let totalUpdated = 0;

  while (true) {
    const result = await db.query(`
      WITH batch AS (
        SELECT id FROM users
        WHERE status IS NULL
        LIMIT $1
        FOR UPDATE SKIP LOCKED
      )
      UPDATE users SET status = 'active'
      WHERE id IN (SELECT id FROM batch)
    `, [batchSize]);

    totalUpdated += result.rowCount ?? 0;

    if ((result.rowCount ?? 0) === 0) break;

    console.log(`Backfilled ${totalUpdated} rows so far...`);

    // Brief pause to reduce database load
    await new Promise(resolve => setTimeout(resolve, 100));
  }

  console.log(`Backfill complete. Total: ${totalUpdated} rows.`);
}
```

## CI/CD Integration

### GitHub Actions

```yaml
# .github/workflows/migrate.yml
name: Database Migration

on:
  push:
    branches: [main]
    paths:
      - 'prisma/migrations/**'
      - 'db/migrations/**'

jobs:
  migrate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Install dependencies
        run: npm ci

      # Prisma
      - name: Run migrations
        run: npx prisma migrate deploy
        env:
          DATABASE_URL: ${{ secrets.DATABASE_URL }}

      # Or dbmate
      - name: Run migrations
        run: |
          curl -fsSL -o /usr/local/bin/dbmate https://github.com/amacneil/dbmate/releases/latest/download/dbmate-linux-amd64
          chmod +x /usr/local/bin/dbmate
          dbmate up
        env:
          DATABASE_URL: ${{ secrets.DATABASE_URL }}
```

### Pre-Deploy Migration Check

```bash
#!/bin/bash
# scripts/check-migrations.sh

# Check for pending migrations before deploying
STATUS=$(npx prisma migrate status 2>&1)

if echo "$STATUS" | grep -q "have not yet been applied"; then
  echo "ERROR: Pending migrations detected!"
  echo "$STATUS"
  exit 1
fi

echo "All migrations are applied."
```

## Rollback Strategies

### With dbmate

```bash
# Built-in rollback of last migration
dbmate rollback
```

### With Prisma (Manual)

Prisma doesn't have built-in rollback. Create a reverse migration:

```bash
# Create a new "undo" migration
npx prisma migrate dev --name undo_last_change --create-only

# Write the reverse SQL manually in the generated file
# Then apply
npx prisma migrate dev
```

### Emergency Rollback Script

```bash
#!/bin/bash
# scripts/emergency-rollback.sh

set -euo pipefail

MIGRATION_TABLE="_prisma_migrations"  # or "schema_migrations" for dbmate

echo "Last 5 migrations:"
psql "$DATABASE_URL" -c "
  SELECT migration_name, finished_at
  FROM $MIGRATION_TABLE
  ORDER BY finished_at DESC
  LIMIT 5;
"

read -p "Enter migration name to roll back to: " TARGET

echo "Rolling back to: $TARGET"
echo "This will mark migrations after $TARGET as rolled back."
echo "You must manually reverse the SQL changes."
read -p "Continue? (y/N): " CONFIRM

if [ "$CONFIRM" = "y" ]; then
  psql "$DATABASE_URL" -c "
    DELETE FROM $MIGRATION_TABLE
    WHERE finished_at > (
      SELECT finished_at FROM $MIGRATION_TABLE
      WHERE migration_name = '$TARGET'
    );
  "
  echo "Migration records cleaned. Apply reverse SQL now."
fi
```

## Gotchas

1. **Always use `CONCURRENTLY` for indexes** — `CREATE INDEX` without it blocks all writes for the entire build time. On a large table, this can be minutes.

2. **Test migrations against production-size data** — a migration that's instant on 1,000 rows might take 30 minutes on 10 million. Always test with realistic data volumes.

3. **Lock timeout** — set `SET lock_timeout = '5s';` before DDL statements to fail fast instead of waiting indefinitely for a lock.

4. **Transaction boundaries** — `CREATE INDEX CONCURRENTLY` and `ALTER TABLE ... VALIDATE CONSTRAINT` cannot run inside a transaction. Some migration tools wrap everything in a transaction by default.

5. **Prisma shadow database** — `prisma migrate dev` creates a temporary "shadow database" to diff against. Ensure your database user has CREATE DATABASE privileges, or configure `shadowDatabaseUrl`.

6. **Backfills are separate from schema changes** — never put a slow backfill in the same migration as a DDL statement. The DDL holds a lock while the backfill runs.

7. **Foreign keys on large tables** — `ADD CONSTRAINT ... FOREIGN KEY` validates all existing rows (full table scan with lock). Use `NOT VALID` and validate separately.

8. **Sequence ownership** — when renaming tables or moving columns, check that sequences (for SERIAL/BIGSERIAL) are still owned by the correct column. Orphaned sequences leak.
