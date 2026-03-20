---
name: postgres-indexing
description: >
  PostgreSQL indexing strategies — B-tree, GIN, GiST, hash, BRIN, partial indexes,
  covering indexes, expression indexes, and multi-column index design.
  Triggers: "postgres index", "create index", "index strategy", "composite index",
  "gin index", "full text search index", "partial index", "covering index".
  NOT for: general query tuning (use postgres-performance), migrations (use postgres-migrations).
version: 1.0.0
allowed-tools: Read, Grep, Glob, Bash, Edit, Write
---

# PostgreSQL Indexing

## Index Type Decision Matrix

| Index Type | Best For | Size | Read Speed | Write Overhead |
|-----------|----------|------|------------|----------------|
| **B-tree** | Equality, range, sorting, LIKE 'prefix%' | Medium | Fast | Medium |
| **Hash** | Equality only (=) | Small | Fastest for = | Low |
| **GIN** | Arrays, JSONB, full-text search, trgm | Large | Fast for containment | High |
| **GiST** | Geometry, ranges, full-text (ranking) | Medium | Good | Medium |
| **BRIN** | Naturally ordered data (timestamps, IDs) | Tiny | Good for ranges | Very low |
| **SP-GiST** | Non-balanced trees, phone numbers, IP addresses | Medium | Specialized | Medium |

### When to Use Each

```
Equality (=)?
├── Only equality, high cardinality → Hash
├── Equality + range/sort → B-tree (default, best general purpose)
└── Equality on JSONB/array → GIN

Range queries (<, >, BETWEEN)?
├── Data naturally ordered (timestamps, sequential IDs) → BRIN (tiny!)
└── General range queries → B-tree

Full-text search?
├── Simple search, no ranking → GIN
└── Need ranking (ts_rank) → GiST (or GIN + separate ranking)

JSONB?
├── Key existence, containment (@>, ?) → GIN
└── Specific key equality → B-tree on expression

Array?
├── Contains (@>), overlap (&&) → GIN
└── Element equality → GIN

Geometry/spatial?
└── Always GiST (or SP-GiST)
```

## B-tree Index Patterns

### Single Column

```sql
-- Basic equality/range
CREATE INDEX idx_users_email ON users (email);

-- Descending (for ORDER BY ... DESC)
CREATE INDEX idx_posts_created ON posts (created_at DESC);

-- Unique index (also enforces constraint)
CREATE UNIQUE INDEX idx_users_email_unique ON users (email);
```

### Multi-Column (Composite)

```sql
-- Order matters! Left-to-right prefix matching
CREATE INDEX idx_orders_status_date ON orders (status, created_at DESC);

-- This index serves:
-- WHERE status = 'active'                          ✓ (first column)
-- WHERE status = 'active' AND created_at > '...'   ✓ (both columns)
-- WHERE status = 'active' ORDER BY created_at DESC ✓ (both columns)
-- ORDER BY status, created_at DESC                 ✓ (both columns)
-- WHERE created_at > '...'                         ✗ (can't skip first column!)
```

### Column Order Rules

1. **Equality columns first** — `WHERE status = 'active' AND type = 'post'`
2. **Range column last** — `AND created_at > '2025-01-01'`
3. **Consider selectivity** — put the most selective column first

```sql
-- For: WHERE tenant_id = ? AND status = ? AND created_at > ?
CREATE INDEX idx_items_tenant_status_date
  ON items (tenant_id, status, created_at DESC);

-- tenant_id first (equality, very selective)
-- status second (equality, few distinct values)
-- created_at last (range)
```

### Covering Index (INCLUDE)

```sql
-- Avoids heap fetch — "Index Only Scan"
CREATE INDEX idx_users_email_covering
  ON users (email)
  INCLUDE (name, avatar_url);

-- This query uses Index Only Scan — no table access needed:
-- SELECT name, avatar_url FROM users WHERE email = 'user@example.com'
```

Use INCLUDE for columns that are SELECTed but not filtered/sorted.

### Partial Index

```sql
-- Only index active users (90% smaller if most users are inactive)
CREATE INDEX idx_users_active_email
  ON users (email)
  WHERE active = true;

-- Only index non-null values
CREATE INDEX idx_users_phone
  ON users (phone)
  WHERE phone IS NOT NULL;

-- Only index recent data
CREATE INDEX idx_orders_pending
  ON orders (created_at DESC)
  WHERE status = 'pending';
```

### Expression Index

```sql
-- Index on lower(email) for case-insensitive lookups
CREATE INDEX idx_users_email_lower ON users (lower(email));
-- Query must match: WHERE lower(email) = lower('User@Example.com')

-- Index on JSONB field
CREATE INDEX idx_profiles_city ON profiles ((data->>'city'));
-- Query: WHERE data->>'city' = 'New York'

-- Index on computed value
CREATE INDEX idx_orders_year ON orders (EXTRACT(YEAR FROM created_at));
-- Query: WHERE EXTRACT(YEAR FROM created_at) = 2025
```

## GIN Index Patterns

### Full-Text Search

```sql
-- Create a tsvector column (or use generated column)
ALTER TABLE posts ADD COLUMN search_vector tsvector
  GENERATED ALWAYS AS (
    setweight(to_tsvector('english', coalesce(title, '')), 'A') ||
    setweight(to_tsvector('english', coalesce(body, '')), 'B')
  ) STORED;

-- Index it
CREATE INDEX idx_posts_search ON posts USING gin (search_vector);

-- Query
SELECT * FROM posts
WHERE search_vector @@ plainto_tsquery('english', 'database optimization')
ORDER BY ts_rank(search_vector, plainto_tsquery('english', 'database optimization')) DESC;
```

### JSONB

```sql
-- General containment (@>, ?, ?|, ?&)
CREATE INDEX idx_metadata ON events USING gin (metadata);

-- Query: WHERE metadata @> '{"type": "click"}'
-- Query: WHERE metadata ? 'user_id'

-- Specific path (smaller, faster for known paths)
CREATE INDEX idx_metadata_type ON events USING gin ((metadata->'type'));
```

### Array

```sql
CREATE INDEX idx_posts_tags ON posts USING gin (tags);

-- Query: WHERE tags @> ARRAY['javascript']     (contains)
-- Query: WHERE tags && ARRAY['react', 'vue']    (overlaps / any of)
```

### Trigram (pg_trgm) for LIKE/ILIKE

```sql
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX idx_users_name_trgm ON users USING gin (name gin_trgm_ops);

-- Now these work with the index:
-- WHERE name LIKE '%smith%'    (substring search!)
-- WHERE name ILIKE '%Smith%'   (case-insensitive!)
-- WHERE name % 'smth'          (similarity/fuzzy)
```

## BRIN Index

```sql
-- Tiny index for naturally ordered data
CREATE INDEX idx_events_timestamp ON events USING brin (created_at);

-- Size comparison on 100M rows:
-- B-tree: ~2GB
-- BRIN:   ~100KB (!)

-- Great for: time-series, logs, events, sequential IDs
-- Bad for: data inserted out of order
```

## Index Maintenance

### Find Unused Indexes

```sql
SELECT
  schemaname,
  relname AS table,
  indexrelname AS index,
  idx_scan AS times_used,
  pg_size_pretty(pg_relation_size(indexrelid)) AS size
FROM pg_stat_user_indexes
WHERE idx_scan = 0
  AND indexrelname NOT LIKE '%_pkey'  -- Keep primary keys
  AND indexrelname NOT LIKE '%_unique%'  -- Keep unique constraints
ORDER BY pg_relation_size(indexrelid) DESC;
```

### Find Missing Indexes

```sql
-- Tables with high sequential scan ratio
SELECT
  relname AS table,
  seq_scan,
  seq_tup_read,
  idx_scan,
  idx_tup_fetch,
  CASE WHEN seq_scan + idx_scan > 0
    THEN round(100.0 * idx_scan / (seq_scan + idx_scan), 1)
    ELSE 0
  END AS idx_scan_pct,
  n_live_tup AS rows
FROM pg_stat_user_tables
WHERE n_live_tup > 10000
  AND seq_scan > idx_scan
ORDER BY seq_tup_read DESC
LIMIT 20;
```

### Index Size

```sql
SELECT
  tablename,
  indexname,
  pg_size_pretty(pg_relation_size(indexname::regclass)) AS size
FROM pg_indexes
WHERE schemaname = 'public'
ORDER BY pg_relation_size(indexname::regclass) DESC;
```

### Rebuild Bloated Indexes

```sql
-- Non-blocking index rebuild
REINDEX INDEX CONCURRENTLY idx_users_email;

-- Or drop and recreate
DROP INDEX CONCURRENTLY idx_users_email;
CREATE INDEX CONCURRENTLY idx_users_email ON users (email);
```

## Multi-Column Index Strategy

### The Query Worksheet

For each important query, note:
1. **WHERE equality columns** (`status = ?`, `tenant_id = ?`)
2. **WHERE range columns** (`created_at > ?`, `price BETWEEN ? AND ?`)
3. **ORDER BY columns** and direction
4. **SELECT columns** (for covering index INCLUDE)

```
Query: SELECT id, title, created_at FROM posts
       WHERE tenant_id = ? AND status = 'published'
       AND created_at > ? ORDER BY created_at DESC LIMIT 20

Equality: tenant_id, status
Range: created_at
Order: created_at DESC
Select: id, title, created_at

Index: CREATE INDEX idx_posts_tenant_status_date
       ON posts (tenant_id, status, created_at DESC)
       INCLUDE (title);
```

## Common Gotchas

1. **Functions prevent index use** — `WHERE YEAR(created_at) = 2025` won't use an index on `created_at`. Use `WHERE created_at >= '2025-01-01' AND created_at < '2026-01-01'`.

2. **Type mismatches** — `WHERE id = '123'` when id is integer causes a cast, bypassing the index.

3. **OR conditions** — `WHERE a = 1 OR b = 2` can't use a composite index on `(a, b)`. Use separate indexes and rely on Bitmap Or.

4. **NOT IN / != / IS NOT NULL** — these are low-selectivity and usually trigger seq scans. Restructure queries to use positive conditions.

5. **Too many indexes** — each index slows writes. Aim for the minimum set that covers your queries. Audit unused indexes monthly.

6. **LIKE '%prefix'** — trailing wildcard works with B-tree (`LIKE 'prefix%'`). Leading wildcard needs trigram GIN index.

7. **Always use CONCURRENTLY** — `CREATE INDEX` without CONCURRENTLY blocks writes for the entire build duration. On a large table, this can be minutes or hours.

8. **Multi-column order matters** — `(a, b)` serves `WHERE a = ?` but NOT `WHERE b = ?`. The index is a sorted list — you can't skip to the middle.
