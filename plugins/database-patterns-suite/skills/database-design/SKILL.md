---
name: database-design
description: >
  Database design patterns — schema modeling, relationships, migrations,
  connection pooling, transactions, and PostgreSQL-specific features like
  enums, triggers, views, and row-level security.
  Triggers: "database design", "schema design", "database migration", "connection pooling",
  "database transaction", "row level security", "postgres enum", "database trigger",
  "materialized view", "database relationships".
  NOT for: SQL query writing (use sql-patterns), Redis (use redis-caching).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Database Design Patterns

## Schema Modeling

### One-to-Many

```sql
-- Parent table
CREATE TABLE organizations (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT NOT NULL,
  slug TEXT NOT NULL UNIQUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Child table
CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  email TEXT NOT NULL,
  name TEXT NOT NULL,
  role TEXT NOT NULL DEFAULT 'member' CHECK (role IN ('owner', 'admin', 'member')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(organization_id, email)
);

CREATE INDEX idx_users_org ON users(organization_id);
```

### Many-to-Many

```sql
-- Join table with extra columns
CREATE TABLE user_projects (
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  role TEXT NOT NULL DEFAULT 'viewer' CHECK (role IN ('owner', 'editor', 'viewer')),
  invited_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (user_id, project_id)
);

-- Indexes for both lookup directions
CREATE INDEX idx_user_projects_project ON user_projects(project_id);
-- user_id is already indexed as part of the primary key
```

### Self-Referencing (Tree/Hierarchy)

```sql
-- Adjacency list (simplest)
CREATE TABLE categories (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  parent_id UUID REFERENCES categories(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  sort_order INT NOT NULL DEFAULT 0
);

CREATE INDEX idx_categories_parent ON categories(parent_id);

-- Materialized path (fast reads, slow writes)
CREATE TABLE categories_mp (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT NOT NULL,
  path TEXT NOT NULL,  -- e.g., '/electronics/phones/iphone'
  depth INT NOT NULL DEFAULT 0
);

CREATE INDEX idx_categories_path ON categories_mp(path text_pattern_ops);
-- Query all children: WHERE path LIKE '/electronics/phones/%'

-- Closure table (fast reads and writes, more storage)
CREATE TABLE category_closure (
  ancestor_id UUID NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
  descendant_id UUID NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
  depth INT NOT NULL,
  PRIMARY KEY (ancestor_id, descendant_id)
);
```

### Polymorphic Associations

```sql
-- Option 1: Shared table with type column (simple, less strict)
CREATE TABLE comments (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  commentable_type TEXT NOT NULL CHECK (commentable_type IN ('post', 'video', 'product')),
  commentable_id UUID NOT NULL,
  body TEXT NOT NULL,
  author_id UUID NOT NULL REFERENCES users(id),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_comments_target ON comments(commentable_type, commentable_id);

-- Option 2: Separate foreign keys (strict, nullable columns)
CREATE TABLE comments_strict (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  post_id UUID REFERENCES posts(id) ON DELETE CASCADE,
  video_id UUID REFERENCES videos(id) ON DELETE CASCADE,
  product_id UUID REFERENCES products(id) ON DELETE CASCADE,
  body TEXT NOT NULL,
  author_id UUID NOT NULL REFERENCES users(id),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CHECK (
    (post_id IS NOT NULL)::int +
    (video_id IS NOT NULL)::int +
    (product_id IS NOT NULL)::int = 1
  )
);
```

## PostgreSQL Enums

```sql
-- Create enum type
CREATE TYPE order_status AS ENUM (
  'pending', 'confirmed', 'processing',
  'shipped', 'delivered', 'cancelled', 'refunded'
);

-- Use in table
CREATE TABLE orders (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  status order_status NOT NULL DEFAULT 'pending',
  total NUMERIC(10,2) NOT NULL
);

-- Add new value (only append, cannot remove or reorder)
ALTER TYPE order_status ADD VALUE 'on_hold' AFTER 'confirmed';

-- Alternative: CHECK constraint (more flexible, easier to modify)
CREATE TABLE orders_check (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  status TEXT NOT NULL DEFAULT 'pending'
    CHECK (status IN ('pending', 'confirmed', 'processing',
                      'shipped', 'delivered', 'cancelled', 'refunded')),
  total NUMERIC(10,2) NOT NULL
);
```

## Triggers

```sql
-- Auto-update updated_at timestamp
CREATE FUNCTION update_modified_column() RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply to any table
CREATE TRIGGER set_updated_at
  BEFORE UPDATE ON users
  FOR EACH ROW
  EXECUTE FUNCTION update_modified_column();

-- Audit log trigger
CREATE TABLE audit_log (
  id BIGSERIAL PRIMARY KEY,
  table_name TEXT NOT NULL,
  record_id UUID NOT NULL,
  action TEXT NOT NULL CHECK (action IN ('INSERT', 'UPDATE', 'DELETE')),
  old_data JSONB,
  new_data JSONB,
  changed_by UUID,
  changed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE FUNCTION audit_trigger() RETURNS TRIGGER AS $$
BEGIN
  INSERT INTO audit_log (table_name, record_id, action, old_data, new_data, changed_by)
  VALUES (
    TG_TABLE_NAME,
    COALESCE(NEW.id, OLD.id),
    TG_OP,
    CASE WHEN TG_OP != 'INSERT' THEN to_jsonb(OLD) END,
    CASE WHEN TG_OP != 'DELETE' THEN to_jsonb(NEW) END,
    current_setting('app.current_user_id', true)::UUID
  );
  RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER users_audit
  AFTER INSERT OR UPDATE OR DELETE ON users
  FOR EACH ROW
  EXECUTE FUNCTION audit_trigger();
```

## Views and Materialized Views

```sql
-- Regular view (always fresh, computed on query)
CREATE VIEW user_stats AS
SELECT
  u.id,
  u.name,
  u.email,
  COUNT(DISTINCT o.id) AS order_count,
  COALESCE(SUM(o.total), 0) AS total_spent,
  MAX(o.created_at) AS last_order_at
FROM users u
LEFT JOIN orders o ON o.user_id = u.id
GROUP BY u.id;

-- Materialized view (cached, must refresh manually)
CREATE MATERIALIZED VIEW monthly_revenue AS
SELECT
  DATE_TRUNC('month', created_at) AS month,
  COUNT(*) AS order_count,
  SUM(total) AS revenue,
  AVG(total) AS avg_order_value
FROM orders
WHERE status = 'delivered'
GROUP BY 1
ORDER BY 1;

-- Create index on materialized view
CREATE UNIQUE INDEX idx_monthly_revenue ON monthly_revenue(month);

-- Refresh (blocks reads during refresh)
REFRESH MATERIALIZED VIEW monthly_revenue;

-- Concurrent refresh (no read blocking, requires unique index)
REFRESH MATERIALIZED VIEW CONCURRENTLY monthly_revenue;
```

## Row-Level Security (RLS)

```sql
-- Enable RLS on table
ALTER TABLE documents ENABLE ROW LEVEL SECURITY;

-- Policy: users can only see their own documents
CREATE POLICY documents_select ON documents
  FOR SELECT
  USING (owner_id = current_setting('app.current_user_id')::UUID);

-- Policy: users can only update their own documents
CREATE POLICY documents_update ON documents
  FOR UPDATE
  USING (owner_id = current_setting('app.current_user_id')::UUID)
  WITH CHECK (owner_id = current_setting('app.current_user_id')::UUID);

-- Policy: org members can see org documents
CREATE POLICY documents_org ON documents
  FOR SELECT
  USING (
    org_id IN (
      SELECT organization_id FROM users
      WHERE id = current_setting('app.current_user_id')::UUID
    )
  );

-- Set the user context in your app
-- (call this at the start of each request/transaction)
SET LOCAL app.current_user_id = 'user-uuid-here';

-- Bypass RLS for admin/service roles
ALTER TABLE documents FORCE ROW LEVEL SECURITY;
-- Table owner bypasses by default. To force even owner:
-- Grant specific roles: GRANT ALL ON documents TO app_user;
```

## Migrations

```bash
# Prisma migrations
npx prisma migrate dev --name add_users_table
npx prisma migrate deploy   # Production
npx prisma migrate reset     # Dev only — drops and recreates

# Drizzle migrations
npx drizzle-kit generate
npx drizzle-kit push         # Dev — direct push
npx drizzle-kit migrate      # Production — apply migrations

# Raw SQL migrations (using dbmate)
npm install -g dbmate
dbmate new add_users_table   # Creates timestamped .sql file
dbmate up                     # Apply pending migrations
dbmate down                   # Rollback last migration
```

### Safe Migration Patterns

```sql
-- Adding a column (safe, non-blocking)
ALTER TABLE users ADD COLUMN bio TEXT;

-- Adding NOT NULL column (safe approach)
-- Step 1: Add nullable
ALTER TABLE users ADD COLUMN role TEXT;
-- Step 2: Backfill
UPDATE users SET role = 'member' WHERE role IS NULL;
-- Step 3: Add constraint
ALTER TABLE users ALTER COLUMN role SET NOT NULL;
ALTER TABLE users ALTER COLUMN role SET DEFAULT 'member';

-- Creating index concurrently (non-blocking)
CREATE INDEX CONCURRENTLY idx_users_email ON users(email);

-- Renaming a column (use a view for backward compatibility)
ALTER TABLE users RENAME COLUMN name TO full_name;
CREATE VIEW users_compat AS
  SELECT *, full_name AS name FROM users;

-- Dropping a column (safe approach)
-- Step 1: Stop reading from the column in code
-- Step 2: Deploy
-- Step 3: Drop the column
ALTER TABLE users DROP COLUMN IF EXISTS legacy_field;
```

## Connection Pooling

```typescript
// pg-pool (built into 'pg' package)
import { Pool } from "pg";

const pool = new Pool({
  host: process.env.DB_HOST,
  port: parseInt(process.env.DB_PORT || "5432"),
  database: process.env.DB_NAME,
  user: process.env.DB_USER,
  password: process.env.DB_PASSWORD,

  // Pool settings
  min: 2,                    // Minimum idle connections
  max: 20,                   // Maximum total connections
  idleTimeoutMillis: 30000,  // Close idle connections after 30s
  connectionTimeoutMillis: 5000, // Fail if can't connect in 5s

  // SSL for production
  ssl: process.env.NODE_ENV === "production"
    ? { rejectUnauthorized: false }
    : false,
});

// Query helper
async function query<T>(sql: string, params?: any[]): Promise<T[]> {
  const client = await pool.connect();
  try {
    const result = await client.query(sql, params);
    return result.rows as T[];
  } finally {
    client.release(); // Always release back to pool
  }
}

// Graceful shutdown
process.on("SIGTERM", async () => {
  await pool.end();
  process.exit(0);
});
```

```
# External pooler: PgBouncer config
# pgbouncer.ini
[databases]
myapp = host=localhost port=5432 dbname=myapp

[pgbouncer]
listen_port = 6432
listen_addr = 0.0.0.0
pool_mode = transaction       # Best for web apps
max_client_conn = 1000        # Max client connections
default_pool_size = 20        # Connections per database
min_pool_size = 5
reserve_pool_size = 5
server_idle_timeout = 300
```

## Transactions

```typescript
// Basic transaction
async function transferFunds(fromId: string, toId: string, amount: number) {
  const client = await pool.connect();
  try {
    await client.query("BEGIN");

    // Debit
    const debit = await client.query(
      "UPDATE accounts SET balance = balance - $1 WHERE id = $2 AND balance >= $1 RETURNING balance",
      [amount, fromId]
    );
    if (debit.rowCount === 0) {
      throw new Error("Insufficient funds");
    }

    // Credit
    await client.query(
      "UPDATE accounts SET balance = balance + $1 WHERE id = $2",
      [amount, toId]
    );

    // Record transfer
    await client.query(
      "INSERT INTO transfers (from_id, to_id, amount) VALUES ($1, $2, $3)",
      [fromId, toId, amount]
    );

    await client.query("COMMIT");
  } catch (err) {
    await client.query("ROLLBACK");
    throw err;
  } finally {
    client.release();
  }
}

// Transaction with savepoints
async function complexOperation(client: PoolClient) {
  await client.query("BEGIN");
  try {
    await client.query("INSERT INTO orders ...");

    // Savepoint for optional step
    await client.query("SAVEPOINT optional_step");
    try {
      await client.query("INSERT INTO notifications ...");
    } catch {
      // Notification failure is non-critical
      await client.query("ROLLBACK TO optional_step");
    }

    await client.query("COMMIT");
  } catch (err) {
    await client.query("ROLLBACK");
    throw err;
  }
}

// Advisory locks (application-level locking)
async function withAdvisoryLock<T>(
  lockId: number,
  fn: () => Promise<T>
): Promise<T> {
  const client = await pool.connect();
  try {
    // Acquire lock (blocks until available)
    await client.query("SELECT pg_advisory_lock($1)", [lockId]);
    const result = await fn();
    return result;
  } finally {
    await client.query("SELECT pg_advisory_unlock($1)", [lockId]);
    client.release();
  }
}
```

## Soft Deletes

```sql
-- Soft delete column
ALTER TABLE users ADD COLUMN deleted_at TIMESTAMPTZ;

-- Partial index for active records (most queries only want active)
CREATE INDEX idx_users_active ON users(email) WHERE deleted_at IS NULL;

-- View for convenience
CREATE VIEW active_users AS
SELECT * FROM users WHERE deleted_at IS NULL;

-- Soft delete function
CREATE FUNCTION soft_delete() RETURNS TRIGGER AS $$
BEGIN
  NEW.deleted_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;
```

## Gotchas

1. **UUID primary keys are larger than serial.** 16 bytes vs 4-8 bytes. This affects index size, join performance, and memory usage. For high-volume tables (billions of rows), consider `bigserial`. For most web apps, UUIDs are fine.

2. **`ON DELETE CASCADE` can be dangerous.** Deleting a parent cascades to ALL children. For critical data, use `ON DELETE RESTRICT` and handle deletion explicitly in application code.

3. **Materialized views must be refreshed manually.** There's no auto-refresh in PostgreSQL. Set up a cron job or trigger. `REFRESH MATERIALIZED VIEW CONCURRENTLY` requires a unique index but doesn't block reads.

4. **Connection pool sizing: don't over-provision.** A good rule: `max_connections = (CPU cores * 2) + effective_spindle_count`. For a 4-core server, 10-20 connections is usually optimal. More connections = more context switching = worse performance.

5. **Migrations that lock tables.** `ALTER TABLE ... ADD COLUMN ... DEFAULT` acquired an exclusive lock in PostgreSQL <11. In 11+, it's safe. `CREATE INDEX` without `CONCURRENTLY` locks writes. Always use `CREATE INDEX CONCURRENTLY` in production.

6. **RLS policies are per-role, not per-user.** If you use a single database role for your app (common), you must pass user identity via `SET LOCAL`. RLS doesn't work with connection poolers that share connections across users unless you set the context per transaction.
