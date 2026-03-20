---
name: sql-optimization
description: >
  SQL query optimization — indexing strategies, query analysis with EXPLAIN,
  N+1 prevention, join optimization, partitioning, and common performance pitfalls.
  Triggers: "sql optimization", "slow query", "index strategy", "explain analyze",
  "query performance", "n+1 query", "database performance".
  NOT for: NoSQL databases or schema design (use schema-patterns).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# SQL Query Optimization

## EXPLAIN ANALYZE

```sql
-- Always start here — understand the query plan before optimizing
EXPLAIN (ANALYZE, BUFFERS, FORMAT TEXT)
SELECT u.name, COUNT(o.id) AS order_count
FROM users u
LEFT JOIN orders o ON o.user_id = u.id
WHERE u.created_at > '2026-01-01'
GROUP BY u.id
ORDER BY order_count DESC
LIMIT 20;

-- Key metrics to read:
-- "Seq Scan"     → full table scan, usually bad on large tables
-- "Index Scan"   → using an index, good
-- "Bitmap Scan"  → index + heap, good for medium selectivity
-- "Hash Join"    → building hash table, check memory
-- "Sort"         → external sort if work_mem too small
-- "Rows Removed by Filter" → high number = index needed
-- "actual time"  → first row..last row in milliseconds
-- "Planning Time" vs "Execution Time"
```

### Reading the Plan

| Node Type | What It Means | Action |
|-----------|--------------|--------|
| Seq Scan | Full table scan | Add index on filter/join columns |
| Index Scan | B-tree lookup | Good — verify selectivity |
| Index Only Scan | Covered by index | Best case — no heap access |
| Bitmap Index Scan | Multiple index conditions | Good for OR/range queries |
| Nested Loop | Row-by-row join | Fine for small outer, bad for large |
| Hash Join | Hash table join | Good for equality, check memory |
| Merge Join | Sorted merge | Good when both inputs are sorted |
| Sort | External sort | Increase work_mem or add index with ORDER BY |
| Materialize | Cache subquery result | Check if subquery is too expensive |

## Indexing Strategies

```sql
-- B-tree (default) — equality and range queries
CREATE INDEX idx_users_email ON users (email);
CREATE INDEX idx_orders_created ON orders (created_at);

-- Composite index — column ORDER matters
-- Follows the "leftmost prefix" rule
CREATE INDEX idx_orders_user_status ON orders (user_id, status);
-- Supports: WHERE user_id = ?
-- Supports: WHERE user_id = ? AND status = ?
-- Does NOT support: WHERE status = ? (skips first column)

-- Partial index — index only rows that matter
CREATE INDEX idx_orders_pending ON orders (created_at)
  WHERE status = 'pending';
-- Smaller index, faster queries on common filters

-- Covering index (INCLUDE) — avoid heap lookup
CREATE INDEX idx_orders_user_covering ON orders (user_id)
  INCLUDE (total, status);
-- Index Only Scan when SELECT only needs user_id, total, status

-- Expression index — index computed values
CREATE INDEX idx_users_lower_email ON users (LOWER(email));
-- Now WHERE LOWER(email) = 'user@test.com' uses the index

-- GIN index — for JSONB, arrays, full-text search
CREATE INDEX idx_products_tags ON products USING GIN (tags);
-- Supports: WHERE tags @> '["electronics"]'

CREATE INDEX idx_posts_search ON posts USING GIN (to_tsvector('english', title || ' ' || body));
-- Full-text search with ts_query

-- BRIN index — for naturally ordered data (timestamps, serial IDs)
CREATE INDEX idx_events_created ON events USING BRIN (created_at);
-- Tiny index, great for append-only tables
```

### Index Selection Rules

| Query Pattern | Best Index Type |
|--------------|----------------|
| `WHERE col = value` | B-tree |
| `WHERE col > value` | B-tree |
| `WHERE col IN (...)` | B-tree |
| `WHERE col1 = ? AND col2 = ?` | Composite B-tree (col1, col2) |
| `WHERE col LIKE 'prefix%'` | B-tree |
| `WHERE col LIKE '%suffix'` | GIN trigram (`pg_trgm`) |
| `WHERE jsonb_col @> '{}'` | GIN |
| `WHERE array_col @> ARRAY[1]` | GIN |
| `WHERE col = ? ORDER BY other` | Composite B-tree (col, other) |
| Timestamp range on append-only | BRIN |
| Full-text search | GIN `tsvector` |

## Common Query Optimizations

### N+1 Problem

```typescript
// BAD: N+1 — 1 query for users + N queries for orders
const users = await db.query("SELECT * FROM users LIMIT 100");
for (const user of users) {
  user.orders = await db.query("SELECT * FROM orders WHERE user_id = $1", [user.id]);
  // 101 queries total!
}

// GOOD: JOIN or subquery — 1 query
const usersWithOrders = await db.query(`
  SELECT u.*, json_agg(o.*) AS orders
  FROM users u
  LEFT JOIN orders o ON o.user_id = u.id
  GROUP BY u.id
  LIMIT 100
`);

// GOOD with Prisma: use include (generates JOIN)
const users = await prisma.user.findMany({
  take: 100,
  include: { orders: true },
});
```

### Pagination

```sql
-- BAD: OFFSET pagination — gets slower as offset grows
SELECT * FROM products ORDER BY id LIMIT 20 OFFSET 10000;
-- Scans and discards 10,000 rows every time

-- GOOD: Cursor-based pagination — constant performance
SELECT * FROM products
WHERE id > $1  -- last seen ID
ORDER BY id
LIMIT 20;

-- For non-unique sort columns, use composite cursor
SELECT * FROM products
WHERE (created_at, id) > ($1, $2)
ORDER BY created_at, id
LIMIT 20;
```

### Batch Operations

```sql
-- BAD: Insert one at a time in a loop
INSERT INTO events (type, data) VALUES ('click', '{}');
INSERT INTO events (type, data) VALUES ('view', '{}');
-- N round-trips to database

-- GOOD: Batch insert
INSERT INTO events (type, data) VALUES
  ('click', '{}'),
  ('view', '{}'),
  ('scroll', '{}');
-- 1 round-trip

-- GOOD: Upsert batch (PostgreSQL)
INSERT INTO products (sku, name, price)
VALUES ('A1', 'Widget', 9.99), ('A2', 'Gadget', 19.99)
ON CONFLICT (sku) DO UPDATE SET
  name = EXCLUDED.name,
  price = EXCLUDED.price;
```

### Avoiding Sequential Scans

```sql
-- BAD: Function on indexed column prevents index use
SELECT * FROM users WHERE YEAR(created_at) = 2026;
-- Forces sequential scan even with index on created_at

-- GOOD: Use range comparison
SELECT * FROM users
WHERE created_at >= '2026-01-01' AND created_at < '2027-01-01';

-- BAD: Implicit type cast
SELECT * FROM users WHERE id = '123';  -- id is integer, '123' is text
-- May not use index due to type mismatch

-- GOOD: Match types
SELECT * FROM users WHERE id = 123;

-- BAD: OR on different columns
SELECT * FROM products WHERE name = 'Widget' OR category_id = 5;
-- Often results in sequential scan

-- GOOD: UNION ALL for different columns
SELECT * FROM products WHERE name = 'Widget'
UNION ALL
SELECT * FROM products WHERE category_id = 5 AND name != 'Widget';
```

## Partitioning

```sql
-- Range partitioning for time-series data
CREATE TABLE events (
  id BIGSERIAL,
  event_type TEXT NOT NULL,
  payload JSONB,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
) PARTITION BY RANGE (created_at);

-- Create monthly partitions
CREATE TABLE events_2026_01 PARTITION OF events
  FOR VALUES FROM ('2026-01-01') TO ('2026-02-01');
CREATE TABLE events_2026_02 PARTITION OF events
  FOR VALUES FROM ('2026-02-01') TO ('2026-03-01');

-- Auto-create partitions (pg_partman extension)
SELECT create_parent('public.events', 'created_at', 'native', 'monthly');

-- Queries automatically prune irrelevant partitions
SELECT * FROM events WHERE created_at > '2026-03-01';
-- Only scans events_2026_03, skips all earlier partitions
```

## Query Monitoring

```sql
-- Find slowest queries (requires pg_stat_statements extension)
SELECT
  calls,
  round(total_exec_time::numeric, 2) AS total_ms,
  round(mean_exec_time::numeric, 2) AS avg_ms,
  round(max_exec_time::numeric, 2) AS max_ms,
  rows,
  query
FROM pg_stat_statements
ORDER BY mean_exec_time DESC
LIMIT 20;

-- Find unused indexes
SELECT
  schemaname, relname AS table, indexrelname AS index,
  idx_scan AS scans, pg_size_pretty(pg_relation_size(indexrelid)) AS size
FROM pg_stat_user_indexes
WHERE idx_scan = 0
ORDER BY pg_relation_size(indexrelid) DESC;

-- Find missing indexes (sequential scans on large tables)
SELECT
  relname AS table,
  seq_scan, seq_tup_read,
  idx_scan, idx_tup_fetch,
  n_live_tup AS row_count
FROM pg_stat_user_tables
WHERE seq_scan > 100 AND n_live_tup > 10000
ORDER BY seq_tup_read DESC;
```

## Gotchas

1. **Composite index column order matters** — `INDEX (user_id, status)` supports `WHERE user_id = ?` but NOT `WHERE status = ?` alone. The leftmost prefix rule means the first column must be in the WHERE clause to use the index.

2. **Too many indexes slow writes** — Every INSERT/UPDATE/DELETE must update every index on that table. A table with 10 indexes has 10x the write overhead. Drop unused indexes (check `pg_stat_user_indexes`).

3. **OFFSET pagination is O(n)** — `OFFSET 10000` scans and discards 10,000 rows. Use cursor-based pagination (WHERE id > last_id) for constant-time page loads on large datasets.

4. **COUNT(*) is expensive in PostgreSQL** — `SELECT COUNT(*) FROM big_table` does a full sequential scan. For approximate counts, use `SELECT reltuples FROM pg_class WHERE relname = 'big_table'`.

5. **NULL values are not indexed by default** — `WHERE col IS NULL` won't use a B-tree index. Create a partial index: `CREATE INDEX idx ON table (col) WHERE col IS NULL` if you query for NULLs frequently.

6. **VACUUM is critical** — Dead rows from UPDATE/DELETE bloat the table. Autovacuum handles this, but tune `autovacuum_vacuum_scale_factor` for large tables. Without vacuuming, indexes bloat and queries slow down.
