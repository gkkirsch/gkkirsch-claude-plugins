---
name: schema-patterns
description: >
  Database schema design patterns — normalization, relationships, migrations,
  soft deletes, audit trails, multi-tenancy, and common modeling patterns.
  Triggers: "schema design", "database modeling", "migrations", "soft delete",
  "audit trail", "multi-tenant", "polymorphic", "data modeling".
  NOT for: Query optimization (use sql-optimization).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Database Schema Patterns

## Core Table Template

```sql
-- Every table should have these baseline columns
CREATE TABLE resources (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  -- domain columns here
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Auto-update updated_at
CREATE OR REPLACE FUNCTION update_timestamp()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER set_updated_at
  BEFORE UPDATE ON resources
  FOR EACH ROW EXECUTE FUNCTION update_timestamp();
```

### ID Strategy

| Strategy | Pros | Cons | Use When |
|----------|------|------|----------|
| UUID v4 | No collisions, no sequence bottleneck | 128 bits, random = bad index locality | Distributed systems, APIs |
| UUID v7 | Time-ordered + random, great index performance | 128 bits | New projects (best default) |
| BIGSERIAL | Small, fast, sortable | Exposes count, sequence bottleneck | Internal-only, single DB |
| ULID | Time-ordered, string-sortable | Custom type | When you need string IDs with ordering |
| nanoid | Short, URL-safe | No time ordering | Public-facing short IDs |

## Relationships

### One-to-Many

```sql
CREATE TABLE teams (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
  email TEXT NOT NULL UNIQUE,
  name TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_team ON users (team_id);
```

### Many-to-Many

```sql
CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT NOT NULL
);

CREATE TABLE roles (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT NOT NULL UNIQUE
);

-- Join table with extra metadata
CREATE TABLE user_roles (
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
  granted_by UUID REFERENCES users(id),
  granted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (user_id, role_id)
);

CREATE INDEX idx_user_roles_role ON user_roles (role_id);
```

### Self-Referencing (Tree/Hierarchy)

```sql
-- Adjacency list (simple, recursive queries needed)
CREATE TABLE categories (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  parent_id UUID REFERENCES categories(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  depth INT NOT NULL DEFAULT 0
);

CREATE INDEX idx_categories_parent ON categories (parent_id);

-- Recursive query to get full tree
WITH RECURSIVE tree AS (
  SELECT id, name, parent_id, 0 AS depth
  FROM categories WHERE parent_id IS NULL
  UNION ALL
  SELECT c.id, c.name, c.parent_id, t.depth + 1
  FROM categories c JOIN tree t ON c.parent_id = t.id
)
SELECT * FROM tree ORDER BY depth, name;

-- Materialized path (fast reads, complex writes)
CREATE TABLE categories (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT NOT NULL,
  path TEXT NOT NULL,  -- e.g., '/electronics/phones/smartphones'
  depth INT NOT NULL DEFAULT 0
);

-- Find all descendants
SELECT * FROM categories WHERE path LIKE '/electronics/phones/%';
```

## Soft Deletes

```sql
CREATE TABLE posts (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  title TEXT NOT NULL,
  content TEXT,
  deleted_at TIMESTAMPTZ,  -- NULL = active, timestamp = deleted
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Partial index for active records (most queries)
CREATE INDEX idx_posts_active ON posts (created_at)
  WHERE deleted_at IS NULL;

-- Views for convenience
CREATE VIEW active_posts AS
  SELECT * FROM posts WHERE deleted_at IS NULL;

-- Soft delete
UPDATE posts SET deleted_at = NOW() WHERE id = $1;

-- Restore
UPDATE posts SET deleted_at = NULL WHERE id = $1;

-- Hard delete (cleanup job)
DELETE FROM posts WHERE deleted_at < NOW() - INTERVAL '90 days';
```

```typescript
// Prisma middleware for automatic soft delete filtering
prisma.$use(async (params, next) => {
  if (params.model === "Post") {
    if (params.action === "delete") {
      params.action = "update";
      params.args.data = { deletedAt: new Date() };
    }
    if (params.action === "findMany" || params.action === "findFirst") {
      params.args.where = { ...params.args.where, deletedAt: null };
    }
  }
  return next(params);
});
```

## Audit Trail

```sql
CREATE TABLE audit_log (
  id BIGSERIAL PRIMARY KEY,
  table_name TEXT NOT NULL,
  record_id UUID NOT NULL,
  action TEXT NOT NULL CHECK (action IN ('INSERT', 'UPDATE', 'DELETE')),
  old_data JSONB,
  new_data JSONB,
  changed_by UUID REFERENCES users(id),
  changed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  ip_address INET
);

CREATE INDEX idx_audit_record ON audit_log (table_name, record_id);
CREATE INDEX idx_audit_time ON audit_log (changed_at);

-- Generic audit trigger
CREATE OR REPLACE FUNCTION audit_trigger()
RETURNS TRIGGER AS $$
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

-- Apply to any table
CREATE TRIGGER audit_users
  AFTER INSERT OR UPDATE OR DELETE ON users
  FOR EACH ROW EXECUTE FUNCTION audit_trigger();
```

## Multi-Tenancy

### Column-Based (Shared Schema)

```sql
-- Add tenant_id to every table
CREATE TABLE projects (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL REFERENCES tenants(id),
  name TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Composite index ensures tenant isolation with performance
CREATE INDEX idx_projects_tenant ON projects (tenant_id, created_at);

-- Row Level Security (RLS) — automatic tenant filtering
ALTER TABLE projects ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON projects
  USING (tenant_id = current_setting('app.tenant_id')::UUID);

-- Set tenant context per request
SET app.tenant_id = 'tenant-uuid-here';
-- Now all queries automatically filter by tenant
SELECT * FROM projects; -- Only returns current tenant's projects
```

```typescript
// Express middleware to set tenant context
async function setTenantContext(req: Request, res: Response, next: NextFunction) {
  const tenantId = req.user?.tenantId;
  if (!tenantId) return res.status(403).json({ error: "No tenant" });

  await db.query("SET app.tenant_id = $1", [tenantId]);
  next();
}
```

### Schema-Based (Isolated Schema)

```sql
-- One schema per tenant
CREATE SCHEMA tenant_abc123;

CREATE TABLE tenant_abc123.projects (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT NOT NULL
);

-- Set search_path per request
SET search_path = tenant_abc123, public;
```

| Approach | Isolation | Complexity | Scale |
|----------|-----------|-----------|-------|
| Column + RLS | Low | Low | High (millions of tenants) |
| Schema per tenant | Medium | Medium | Medium (thousands) |
| Database per tenant | High | High | Low (hundreds) |

## Polymorphic Associations

```sql
-- Option A: Separate tables (preferred — maintains FK integrity)
CREATE TABLE comment_targets (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  target_type TEXT NOT NULL CHECK (target_type IN ('post', 'video', 'photo'))
);

CREATE TABLE posts (
  id UUID PRIMARY KEY REFERENCES comment_targets(id),
  title TEXT NOT NULL
);

CREATE TABLE comments (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  target_id UUID NOT NULL REFERENCES comment_targets(id) ON DELETE CASCADE,
  body TEXT NOT NULL,
  author_id UUID NOT NULL REFERENCES users(id)
);

-- Option B: Discriminated column (simpler, no FK possible)
CREATE TABLE comments (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  commentable_type TEXT NOT NULL,  -- 'post', 'video', 'photo'
  commentable_id UUID NOT NULL,    -- No FK constraint possible
  body TEXT NOT NULL,
  author_id UUID NOT NULL REFERENCES users(id)
);

CREATE INDEX idx_comments_target ON comments (commentable_type, commentable_id);
```

## Migrations

```typescript
// Knex migration example
export async function up(knex: Knex) {
  await knex.schema.createTable("products", (table) => {
    table.uuid("id").primary().defaultTo(knex.fn.uuid());
    table.string("name", 255).notNullable();
    table.decimal("price", 10, 2).notNullable();
    table.text("description");
    table.uuid("category_id").references("id").inTable("categories").onDelete("SET NULL");
    table.timestamps(true, true); // created_at, updated_at
    table.index(["category_id"]);
  });
}

export async function down(knex: Knex) {
  await knex.schema.dropTable("products");
}
```

### Safe Migration Practices

| Do | Don't |
|----|-------|
| Add columns as nullable first, backfill, then add NOT NULL | Add NOT NULL column without default (locks table) |
| Create index CONCURRENTLY | Create index without CONCURRENTLY (locks writes) |
| Deploy migration before code that uses it | Deploy code that requires new column before migration |
| Add new column, deploy new code, drop old column | Rename column (breaks running code) |
| Use transactions for multi-step DDL | Run DDL without transaction (partial failure) |

```sql
-- Safe: Add index without locking writes
CREATE INDEX CONCURRENTLY idx_orders_email ON orders (email);

-- Safe: Add nullable column (instant, no rewrite)
ALTER TABLE users ADD COLUMN phone TEXT;

-- Dangerous: Add NOT NULL without default (rewrites entire table)
ALTER TABLE users ADD COLUMN status TEXT NOT NULL; -- LOCKS TABLE

-- Safe: Add with default (instant in PostgreSQL 11+)
ALTER TABLE users ADD COLUMN status TEXT NOT NULL DEFAULT 'active';
```

## JSONB Patterns

```sql
-- Structured JSONB for flexible attributes
CREATE TABLE products (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT NOT NULL,
  attributes JSONB NOT NULL DEFAULT '{}',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- GIN index for JSONB queries
CREATE INDEX idx_products_attrs ON products USING GIN (attributes);

-- Query patterns
SELECT * FROM products WHERE attributes @> '{"color": "red"}';
SELECT * FROM products WHERE attributes->>'brand' = 'Acme';
SELECT * FROM products WHERE (attributes->>'weight')::numeric > 5.0;

-- JSONB validation with CHECK constraint
ALTER TABLE products ADD CONSTRAINT valid_attributes CHECK (
  jsonb_typeof(attributes) = 'object'
  AND attributes ? 'color'  -- must have color key
);
```

## Gotchas

1. **UUID v4 has bad index performance** — Random UUIDs fragment B-tree indexes because inserts are scattered. Use UUID v7 (time-ordered) or ULID for better insert performance and index locality on large tables.

2. **ON DELETE CASCADE can be dangerous** — Deleting a team cascades to all users, which cascades to all their posts, comments, etc. Use ON DELETE SET NULL or ON DELETE RESTRICT for important relationships and handle cleanup explicitly.

3. **Adding NOT NULL to existing column locks the table** — `ALTER TABLE ADD COLUMN col TEXT NOT NULL` without a default rewrites every row. In PostgreSQL 11+, `DEFAULT 'value'` makes it instant. Always add nullable first on large tables.

4. **Don't use ENUM for evolving values** — Adding a new enum value requires `ALTER TYPE`, which can be problematic. Use a TEXT column with a CHECK constraint instead: `CHECK (status IN ('active', 'paused', 'deleted'))`.

5. **Soft deletes break unique constraints** — If email is UNIQUE and you soft-delete a user, they can't re-register with the same email. Use a partial unique index: `CREATE UNIQUE INDEX ON users (email) WHERE deleted_at IS NULL`.

6. **Multi-tenant RLS needs SET per connection** — Row Level Security only works if `app.tenant_id` is set on each database connection. Miss it in one middleware path and you leak data between tenants.
