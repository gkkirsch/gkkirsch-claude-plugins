# PostgreSQL Index Strategies Reference

Comprehensive reference for understanding, creating, and maintaining indexes in
PostgreSQL for optimal query performance.

---

## 1. Index Type Overview

### Comparison Matrix

| Feature                  | B-tree          | GIN               | GiST               | BRIN              | Hash           |
|--------------------------|-----------------|--------------------|---------------------|-------------------|----------------|
| **Default type**         | Yes             | No                 | No                  | No                | No             |
| **Equality (=)**         | Yes             | Yes                | Yes                 | Yes               | Yes            |
| **Range (<, >, BETWEEN)**| Yes             | No                 | Yes (for ranges)    | Yes               | No             |
| **Pattern (LIKE 'a%')**  | Yes (C locale)  | Yes (pg_trgm)      | No                  | No                | No             |
| **Pattern (LIKE '%a%')** | No              | Yes (pg_trgm)      | No                  | No                | No             |
| **Full-text search**     | No              | Yes (tsvector)     | Yes (tsvector)      | No                | No             |
| **Array containment**    | No              | Yes (@>, &&)       | No                  | No                | No             |
| **JSONB containment**    | No              | Yes (@>, ?, ?&)    | No                  | No                | No             |
| **Geometric/spatial**    | No              | No                 | Yes (PostGIS)       | No                | No             |
| **Range types**          | Yes (equality)  | No                 | Yes (overlap, adj.) | Yes               | No             |
| **Nearest neighbor**     | No              | No                 | Yes (ORDER BY <->)  | No                | No             |
| **Multi-column**         | Yes (up to 32)  | Yes (up to 32)     | Yes (up to 32)      | Yes (up to 32)    | No             |
| **Index-only scan**      | Yes             | Limited (PG 13+)   | No                  | No                | No             |
| **Unique constraint**    | Yes             | No                 | No                  | No                | No             |
| **Exclusion constraint** | No              | No                 | Yes                 | No                | No             |
| **Covering (INCLUDE)**   | Yes (PG 11+)   | No                 | Yes (PG 12+)        | No                | No             |
| **Size**                 | Medium          | Large              | Medium-Large        | Very Small        | Medium         |
| **Build speed**          | Fast            | Slow               | Slow                | Very Fast         | Fast           |
| **Write overhead**       | Low             | Medium-High        | Medium              | Very Low          | Low            |
| **Parallel build**       | Yes (PG 11+)   | Yes (PG 14+)       | No                  | Yes (PG 14+)      | No             |
| **WAL-logged**           | Yes             | Yes                | Yes                 | Yes               | Yes (PG 10+)   |

### Quick Selection Guide

| Use Case                             | Recommended Index Type           |
|--------------------------------------|----------------------------------|
| Primary key / unique constraint      | B-tree (automatic)               |
| Equality + range queries             | B-tree                           |
| Full-text search                     | GIN on tsvector                  |
| JSONB field queries                  | GIN with jsonb_path_ops          |
| Array containment                    | GIN                              |
| Trigram / fuzzy text search          | GIN with pg_trgm                 |
| Geometric / PostGIS spatial          | GiST                             |
| Range overlap / adjacency            | GiST                             |
| Nearest-neighbor / KNN              | GiST                             |
| Time-series / append-only data       | BRIN                             |
| Equality-only (no range needed)      | Hash (or B-tree)                 |
| Exclusion constraints                | GiST                             |

---

## 2. B-tree Indexes

### Internal Structure

B-tree indexes organize data in a balanced tree where leaf nodes contain index entries
sorted by key value. Internal nodes contain routing entries that guide traversal to the
correct leaf page. Each leaf node links to adjacent leaves, enabling efficient range
scans.

Structure: Root -> Internal Nodes -> Leaf Nodes (doubly linked list)

### Supported Operators

B-tree indexes support: `<`, `<=`, `=`, `>=`, `>`, `BETWEEN`, `IN`, `IS NULL`,
`IS NOT NULL`, and pattern matching with `LIKE 'prefix%'` (with C locale or
`text_pattern_ops`).

### Creating B-tree Indexes

```sql
-- Single column
CREATE INDEX idx_orders_customer_id ON orders (customer_id);

-- Multi-column (leftmost prefix rule applies)
CREATE INDEX idx_orders_customer_date ON orders (customer_id, created_at DESC);
-- This index supports:
--   WHERE customer_id = 42
--   WHERE customer_id = 42 AND created_at > '2025-01-01'
--   WHERE customer_id = 42 ORDER BY created_at DESC
-- But NOT efficiently:
--   WHERE created_at > '2025-01-01'  (without customer_id filter)

-- Multi-column with mixed sort order
CREATE INDEX idx_orders_status_date ON orders (status ASC, created_at DESC NULLS LAST);

-- Unique index
CREATE UNIQUE INDEX idx_users_email ON users (email);

-- Concurrent creation (doesn't lock the table for writes)
CREATE INDEX CONCURRENTLY idx_orders_total ON orders (total);
```

### Covering Indexes (INCLUDE)

```sql
-- Include columns that are only needed for retrieval, not filtering/sorting
CREATE INDEX idx_orders_customer_covering
ON orders (customer_id)
INCLUDE (order_date, total, status);

-- Now this query uses an index-only scan (no heap fetch):
SELECT order_date, total, status FROM orders WHERE customer_id = 42;
```

The INCLUDE columns are stored in leaf nodes but not in internal nodes, so they
don't bloat the tree structure. They cannot be used for filtering or sorting — only
for satisfying the query's SELECT list.

### Index-Only Scans

For an index-only scan to work:
1. All columns in SELECT, WHERE, and ORDER BY must be in the index (key + INCLUDE).
2. The visibility map must be up-to-date (run VACUUM regularly).
3. Check with `EXPLAIN` — look for "Index Only Scan" and "Heap Fetches: 0".

```sql
-- Check visibility map status
SELECT relname, n_tup_mod, n_tup_hot_upd,
       pg_stat_get_live_tuples(oid) AS live_tuples
FROM pg_class WHERE relname = 'orders';

-- If Heap Fetches is high, VACUUM the table
VACUUM orders;
```

### Text Pattern Matching

```sql
-- For LIKE 'prefix%' queries with non-C locale, use operator classes:
CREATE INDEX idx_users_name_pattern ON users (name text_pattern_ops);
-- Supports: WHERE name LIKE 'John%'
-- Does NOT support: WHERE name = 'John' (use a regular index for that)

-- For varchar:
CREATE INDEX idx_users_name_varchar ON users (name varchar_pattern_ops);
```

---

## 3. GIN Indexes (Generalized Inverted Index)

### How GIN Works

GIN is an inverted index structure: it maps each element value to a list of row
locations containing that value. Think of it like a book's back-of-book index —
each keyword points to all pages where it appears.

Structure: Entry Tree -> Posting Tree/List (for each entry, a sorted list of TIDs)

### Supported Types and Operator Classes

```sql
-- Array operations (@>, &&, <@)
CREATE INDEX idx_tags ON articles USING gin (tags);
-- Supports: WHERE tags @> ARRAY['postgresql', 'database']
-- Supports: WHERE tags && ARRAY['postgresql', 'mysql']  (overlap)

-- JSONB containment
CREATE INDEX idx_data ON events USING gin (data);
-- Supports: WHERE data @> '{"type": "click"}'
-- Supports: WHERE data ? 'user_id'
-- Supports: WHERE data ?& array['type', 'user_id']

-- JSONB with path_ops (smaller, faster for @> only)
CREATE INDEX idx_data_path ON events USING gin (data jsonb_path_ops);
-- Supports: WHERE data @> '{"type": "click"}'
-- Does NOT support: ? or ?| or ?& operators

-- Full-text search (tsvector)
CREATE INDEX idx_content_fts ON articles USING gin (to_tsvector('english', content));
-- Supports: WHERE to_tsvector('english', content) @@ to_tsquery('postgresql & index')

-- Trigram similarity (pg_trgm)
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX idx_name_trgm ON users USING gin (name gin_trgm_ops);
-- Supports: WHERE name LIKE '%john%'
-- Supports: WHERE name ILIKE '%john%'
-- Supports: WHERE name % 'john'  (similarity)
-- Supports: WHERE name ~* 'j.hn'  (regex)
```

### GIN Tuning: fastupdate and Pending Lists

```sql
-- GIN maintains a "pending list" for fast inserts (fastupdate=on by default)
-- Pending entries are merged into the main index during VACUUM or when the
-- pending list exceeds gin_pending_list_limit

-- Disable fastupdate for read-heavy workloads (slower writes, faster reads)
CREATE INDEX idx_tags ON articles USING gin (tags) WITH (fastupdate = off);

-- Adjust pending list limit (default: 64kB)
ALTER INDEX idx_tags SET (gin_pending_list_limit = 256);

-- Force merge of pending list
VACUUM articles;
```

### GIN vs GiST for Full-Text Search

| Aspect          | GIN                              | GiST                             |
|-----------------|----------------------------------|----------------------------------|
| Build speed     | Slower (3x)                      | Faster                           |
| Search speed    | Faster (3x for exact matches)    | Slower                           |
| Index size      | Larger (2-3x)                    | Smaller                          |
| Update speed    | Slower (with fastupdate: okay)   | Faster                           |
| Best for        | Read-heavy, exact matching       | Write-heavy, ranking queries     |

**Recommendation**: Use GIN for most full-text search. Use GiST only if write
performance is critical and searches are infrequent.

---

## 4. GiST Indexes (Generalized Search Tree)

### How GiST Works

GiST is a balanced tree where each internal node contains a "bounding" predicate
that covers all entries in its subtree. This makes it ideal for data types where
containment and overlap are meaningful — geometric shapes, ranges, network addresses.

### Geometric and PostGIS Use

```sql
-- Geometric types (built-in)
CREATE INDEX idx_locations ON places USING gist (location);
-- Supports: WHERE location <@ box '((0,0),(100,100))'

-- PostGIS spatial index
CREATE INDEX idx_geom ON spatial_data USING gist (geom);
-- Supports: WHERE ST_DWithin(geom, ST_MakePoint(-73.9857, 40.7484)::geography, 1000)
-- Supports: WHERE ST_Intersects(geom, other_geom)
-- Supports: ORDER BY geom <-> ST_MakePoint(-73.9857, 40.7484)::geometry  (KNN)
```

### Range Type Indexes

```sql
-- Range overlap and containment
CREATE INDEX idx_reservation_during ON reservations USING gist (during);
-- Supports: WHERE during && '[2025-06-01, 2025-06-30]'::daterange  (overlap)
-- Supports: WHERE during @> '2025-06-15'::date  (contains point)
-- Supports: WHERE during <@ '[2025-01-01, 2025-12-31]'::daterange  (contained by)
```

### Nearest-Neighbor (KNN) Searches

```sql
-- Find 10 nearest points to a given location
SELECT id, name, location <-> point '(40.7128, -74.0060)' AS distance
FROM places
ORDER BY location <-> point '(40.7128, -74.0060)'
LIMIT 10;
-- GiST index on location enables efficient KNN via index scan

-- PostGIS KNN
SELECT id, name, geom <-> ST_MakePoint(-73.9857, 40.7484)::geometry AS dist
FROM places
ORDER BY geom <-> ST_MakePoint(-73.9857, 40.7484)::geometry
LIMIT 10;
```

### Exclusion Constraints

```sql
-- Prevent overlapping reservations for the same room
CREATE TABLE reservations (
  id serial PRIMARY KEY,
  room_id int NOT NULL,
  during tsrange NOT NULL,
  EXCLUDE USING gist (room_id WITH =, during WITH &&)
);

-- With btree_gist extension for mixing equality and range operators
CREATE EXTENSION IF NOT EXISTS btree_gist;
-- Now room_id (integer =) and during (range &&) work together in the GiST index

INSERT INTO reservations (room_id, during) VALUES
  (1, '[2025-06-01 14:00, 2025-06-01 16:00)');
INSERT INTO reservations (room_id, during) VALUES
  (1, '[2025-06-01 15:00, 2025-06-01 17:00)');
-- ERROR: conflicting key value violates exclusion constraint
```

### Full-Text Search with GiST

```sql
CREATE INDEX idx_content_gist ON articles USING gist (
  to_tsvector('english', content)
);
-- Smaller than GIN but slower for exact lookups
-- Better for ranking and write-heavy workloads
```

---

## 5. BRIN Indexes (Block Range Index)

### How BRIN Works

BRIN divides a table into consecutive block ranges (default 128 pages) and stores
the minimum and maximum values for each range. When a query filters on the indexed
column, BRIN can skip entire block ranges that cannot contain matching rows.

### Ideal Use Cases

BRIN is perfect for:
- **Time-series data** where rows are inserted in chronological order.
- **Append-only tables** where the physical order matches the logical order.
- **Very large tables** where B-tree index size would be prohibitive.

The key requirement is **high physical correlation** between the column value and
the row's physical position in the table.

```sql
-- Check correlation (1.0 = perfectly correlated, ideal for BRIN)
SELECT attname, correlation
FROM pg_stats
WHERE tablename = 'events' AND attname = 'created_at';
-- correlation should be close to 1.0 or -1.0
```

### Creating BRIN Indexes

```sql
-- Default pages_per_range = 128
CREATE INDEX idx_events_created_at ON events USING brin (created_at);

-- Smaller ranges = more precise but larger index
CREATE INDEX idx_events_created_at ON events USING brin (created_at)
  WITH (pages_per_range = 32);

-- Larger ranges = smaller index but less precise filtering
CREATE INDEX idx_events_created_at ON events USING brin (created_at)
  WITH (pages_per_range = 256);

-- Multi-column BRIN
CREATE INDEX idx_events_multi ON events USING brin (created_at, device_type);

-- Enable autosummarize (summarizes new ranges automatically, PG 10+)
CREATE INDEX idx_events_created_at ON events USING brin (created_at)
  WITH (autosummarize = on);
```

### Size Comparison

```sql
-- Compare index sizes (BRIN is typically 100-1000x smaller than B-tree)
SELECT pg_size_pretty(pg_relation_size('idx_events_btree')) AS btree_size,
       pg_size_pretty(pg_relation_size('idx_events_brin'))  AS brin_size;
-- Example: B-tree = 2.1 GB, BRIN = 48 kB
```

### Tuning pages_per_range

| pages_per_range | Index Size | Precision | Best For                    |
|-----------------|------------|-----------|-----------------------------|
| 16              | Larger     | High      | Frequently queried ranges   |
| 128 (default)   | Small      | Medium    | General time-series         |
| 512             | Very small | Low       | Very large rarely-queried   |

### Manual Summarization

```sql
-- Summarize unsummarized ranges
SELECT brin_summarize_new_values('idx_events_created_at');

-- Desummarize and resummarize a specific range (after bulk updates)
SELECT brin_desummarize_range('idx_events_created_at', 0);
SELECT brin_summarize_range('idx_events_created_at', 0);
```

---

## 6. Hash Indexes

### Overview

Hash indexes use a hash table structure. They support only equality comparisons (`=`)
and cannot be used for range queries, sorting, or multi-column indexes.

### History and Current Status

- **Pre-PG 10**: Not WAL-logged, not crash-safe, not recommended.
- **PG 10+**: WAL-logged and crash-safe. Usable in production.
- **PG 11+**: Improved performance, space utilization.

### Creating Hash Indexes

```sql
CREATE INDEX idx_sessions_token ON sessions USING hash (session_token);
-- Only supports: WHERE session_token = 'abc123'
```

### When to Use Hash Over B-tree

Hash indexes are smaller than B-tree for long keys (e.g., UUIDs, text tokens) and
can be faster for pure equality lookups. However, the difference is usually marginal.

```sql
-- Compare sizes
CREATE INDEX idx_sessions_btree ON sessions (session_token);
CREATE INDEX idx_sessions_hash ON sessions USING hash (session_token);

SELECT pg_size_pretty(pg_relation_size('idx_sessions_btree')) AS btree_size,
       pg_size_pretty(pg_relation_size('idx_sessions_hash'))  AS hash_size;
```

**Recommendation**: In most cases, B-tree is preferred because it supports range
queries, sorting, and index-only scans. Use hash only when you are certain you will
never need anything beyond equality checks and want to save space on long keys.

---

## 7. Partial Indexes

Partial indexes include only rows matching a WHERE condition. This reduces index size
and maintenance cost while targeting the queries that matter.

### Use Cases and Examples

```sql
-- Index only active users (if 95% of queries filter on active = true)
CREATE INDEX idx_users_active ON users (email)
WHERE active = true;
-- Supports: SELECT * FROM users WHERE active = true AND email = 'user@example.com'
-- Does NOT support: SELECT * FROM users WHERE active = false AND email = '...'

-- Index only pending orders
CREATE INDEX idx_orders_pending ON orders (created_at)
WHERE status = 'pending';
-- Much smaller than indexing all orders

-- Index only non-null values
CREATE INDEX idx_users_phone ON users (phone)
WHERE phone IS NOT NULL;

-- Index only recent data (use with caution — needs periodic recreation)
CREATE INDEX idx_events_recent ON events (user_id, event_type)
WHERE created_at > '2025-01-01';

-- Unique partial index (unique email among active users only)
CREATE UNIQUE INDEX idx_users_unique_active_email ON users (email)
WHERE active = true;
-- Allows multiple inactive users with the same email
```

### Query Matching Rules

For PostgreSQL to use a partial index, the query's WHERE clause must logically imply
the index's WHERE condition. The planner performs simple constant folding but cannot
reason about arbitrary expressions.

```sql
-- This WILL use the partial index (WHERE active = true)
SELECT * FROM users WHERE active = true AND email = 'test@example.com';

-- This WILL NOT use the partial index (planner can't prove active = true)
SELECT * FROM users WHERE email = 'test@example.com';

-- This WILL NOT use the partial index (different condition)
SELECT * FROM users WHERE active = false AND email = 'test@example.com';
```

---

## 8. Expression Indexes

Expression indexes index the result of an expression or function applied to columns,
rather than the raw column values.

### Examples

```sql
-- Case-insensitive email lookup
CREATE INDEX idx_users_email_lower ON users (lower(email));
SELECT * FROM users WHERE lower(email) = 'user@example.com';

-- Date part extraction
CREATE INDEX idx_orders_year_month ON orders (
  date_part('year', created_at),
  date_part('month', created_at)
);
SELECT * FROM orders
WHERE date_part('year', created_at) = 2025
  AND date_part('month', created_at) = 6;

-- JSONB field extraction
CREATE INDEX idx_events_user_id ON events ((data->>'user_id'));
SELECT * FROM events WHERE data->>'user_id' = '42';

-- Computed value
CREATE INDEX idx_orders_total_with_tax ON orders ((total * 1.08));
SELECT * FROM orders WHERE total * 1.08 > 100;

-- Immutable function (must be IMMUTABLE for expression indexes)
CREATE OR REPLACE FUNCTION normalize_phone(text) RETURNS text
LANGUAGE sql IMMUTABLE AS $$
  SELECT regexp_replace($1, '[^0-9]', '', 'g');
$$;
CREATE INDEX idx_users_phone_normalized ON users (normalize_phone(phone));
SELECT * FROM users WHERE normalize_phone(phone) = '5551234567';
```

### Important Rules

- The function used in the expression must be marked `IMMUTABLE`.
- The query must use the exact same expression for the index to be used.
- Expression indexes are maintained automatically on INSERT/UPDATE.

---

## 9. Covering Indexes (INCLUDE Clause)

The INCLUDE clause adds columns to the leaf nodes of a B-tree index without making
them part of the index key. This enables index-only scans without bloating the tree.

### Syntax and Examples

```sql
-- Covering index: key columns for searching, included columns for retrieval
CREATE INDEX idx_orders_customer_covering ON orders (customer_id)
INCLUDE (order_date, total, status);

-- This query now uses an index-only scan:
EXPLAIN SELECT order_date, total, status
FROM orders WHERE customer_id = 42;
-- -> Index Only Scan using idx_orders_customer_covering

-- Without INCLUDE, you'd need all columns in the key:
CREATE INDEX idx_orders_all_cols ON orders (customer_id, order_date, total, status);
-- This works but bloats internal nodes and wastes space
```

### Key vs INCLUDE Columns

| Aspect            | Key columns                     | INCLUDE columns              |
|-------------------|---------------------------------|------------------------------|
| Used for search   | Yes                             | No                           |
| Used for sorting  | Yes                             | No                           |
| Stored in         | All tree levels                 | Leaf nodes only              |
| Increases tree depth | Yes (wider keys = deeper tree) | No                          |
| Unique constraint | Enforced                        | Not enforced                 |

### GiST Covering Indexes (PG 12+)

```sql
-- GiST also supports INCLUDE for covering index-like behavior
CREATE INDEX idx_places_location ON places USING gist (location)
INCLUDE (name, category);

-- Enables returning name and category without heap fetch for spatial queries
SELECT name, category FROM places
WHERE location <@ box '((0,0),(100,100))';
```

---

## 10. Index Maintenance

### Detecting Index Bloat

Index bloat occurs when dead tuples are not reclaimed, leaving empty space in index
pages. This increases index size and degrades scan performance.

```sql
-- Estimate B-tree bloat using pgstattuple extension
CREATE EXTENSION IF NOT EXISTS pgstattuple;

SELECT * FROM pgstatindex('idx_orders_customer_id');
-- Key metrics:
--   tree_level: depth of B-tree (should be low, typically 2-4)
--   internal_pages, leaf_pages: page counts
--   empty_pages: pages with no entries (bloat indicator)
--   deleted_pages: pages marked for reuse
--   avg_leaf_density: percentage of leaf page utilization (< 50% = bloated)
--   leaf_fragmentation: percentage of out-of-order leaf pages

-- Quick bloat estimation query (no extension needed)
SELECT
  schemaname || '.' || indexrelname AS index_name,
  pg_size_pretty(pg_relation_size(indexrelid)) AS index_size,
  idx_scan AS times_used,
  idx_tup_read AS tuples_read,
  idx_tup_fetch AS tuples_fetched
FROM pg_stat_user_indexes
ORDER BY pg_relation_size(indexrelid) DESC;
```

### Finding Unused Indexes

```sql
-- Indexes that have never been scanned (candidates for removal)
SELECT
  schemaname || '.' || relname AS table_name,
  indexrelname AS index_name,
  pg_size_pretty(pg_relation_size(indexrelid)) AS index_size,
  idx_scan AS times_scanned
FROM pg_stat_user_indexes
WHERE idx_scan = 0
  AND indexrelname NOT LIKE '%_pkey'  -- keep primary keys
  AND indexrelname NOT LIKE '%_unique%'  -- keep unique constraints
ORDER BY pg_relation_size(indexrelid) DESC;
```

**Warning**: Reset `pg_stat_user_indexes` with `pg_stat_reset()`. Make sure the stats
have been accumulating long enough to cover all query patterns (at least one full
business cycle — weekly/monthly reporting, etc.).

### Finding Duplicate Indexes

```sql
-- Find indexes with identical column definitions
SELECT
  a.indrelid::regclass AS table_name,
  a.indexrelid::regclass AS index_a,
  b.indexrelid::regclass AS index_b,
  pg_size_pretty(pg_relation_size(a.indexrelid)) AS size_a,
  pg_size_pretty(pg_relation_size(b.indexrelid)) AS size_b
FROM pg_index a
JOIN pg_index b ON a.indrelid = b.indrelid
  AND a.indexrelid < b.indexrelid
  AND a.indkey = b.indkey
  AND a.indclass = b.indclass;
```

### REINDEX CONCURRENTLY

```sql
-- Rebuild an index without locking the table (PG 12+)
REINDEX INDEX CONCURRENTLY idx_orders_customer_id;

-- Rebuild all indexes on a table
REINDEX TABLE CONCURRENTLY orders;

-- Rebuild all indexes in a schema
REINDEX SCHEMA CONCURRENTLY public;
```

`REINDEX CONCURRENTLY` creates a new index alongside the old one, then swaps them
atomically. It takes longer but doesn't block reads or writes.

### Monitoring Index Health

```sql
-- Index usage statistics
SELECT
  schemaname,
  relname AS table_name,
  indexrelname AS index_name,
  idx_scan AS scans,
  idx_tup_read AS tuples_read,
  idx_tup_fetch AS tuples_fetched,
  pg_size_pretty(pg_relation_size(indexrelid)) AS index_size,
  pg_size_pretty(pg_total_relation_size(relid)) AS table_total_size
FROM pg_stat_user_indexes
ORDER BY idx_scan DESC;

-- Cache hit ratio for indexes
SELECT
  sum(idx_blks_hit) AS cache_hits,
  sum(idx_blks_read) AS disk_reads,
  round(100.0 * sum(idx_blks_hit) / nullif(sum(idx_blks_hit) + sum(idx_blks_read), 0), 2)
    AS hit_ratio_pct
FROM pg_statio_user_indexes;
```

---

## 11. Index Selection Checklist

Follow this step-by-step process when choosing an index for a query.

### Step 1: Identify the Problem Query

```sql
-- Find slow queries via pg_stat_statements
SELECT query,
  calls,
  mean_exec_time AS avg_ms,
  total_exec_time AS total_ms,
  rows
FROM pg_stat_statements
ORDER BY mean_exec_time DESC
LIMIT 20;
```

### Step 2: Analyze Current Execution Plan

```sql
EXPLAIN (ANALYZE, BUFFERS, TIMING) <your_query>;
```

Look for:
- Sequential scans on large tables (potential for index).
- High buffer read counts (I/O-heavy operations).
- Large row estimate mismatches (need ANALYZE or better statistics).
- Sort nodes with large memory/disk usage (potential for index-ordered scan).

### Step 3: Identify the WHERE Clause Patterns

| Pattern                          | Index Strategy                          |
|----------------------------------|-----------------------------------------|
| `column = value`                 | B-tree on column                        |
| `column > value` / range         | B-tree on column                        |
| `column = v1 AND column2 = v2`  | B-tree on (column, column2)             |
| `column = v1 OR column = v2`     | B-tree on column (bitmap combine)       |
| `column IN (v1, v2, ...)`        | B-tree on column                        |
| `array_col @> ARRAY[val]`        | GIN on array_col                        |
| `jsonb_col @> '{...}'`           | GIN on jsonb_col                        |
| `text_col LIKE '%term%'`         | GIN with pg_trgm on text_col            |
| `tsvector @@ tsquery`            | GIN on tsvector column                  |
| `range_col && range_val`         | GiST on range_col                       |
| `geom <-> point` (KNN)           | GiST on geom                            |
| `column = value` (time-ordered)  | BRIN if high correlation                |

### Step 4: Consider Column Ordering for Multi-Column B-tree

1. **Equality columns first**: Columns used with `=` go leftmost.
2. **Range column last**: The column used with `>`, `<`, `BETWEEN` goes rightmost.
3. **Sort columns**: Match the ORDER BY to avoid a separate sort step.

```sql
-- Query pattern: WHERE status = 'active' AND created_at > '2025-01-01' ORDER BY created_at
-- Optimal index:
CREATE INDEX idx_orders_status_created ON orders (status, created_at);
```

### Step 5: Consider Partial and Covering

- If only a fraction of rows match the common filter, add a WHERE clause (partial).
- If the query SELECTs additional columns, add them via INCLUDE (covering).

```sql
CREATE INDEX idx_orders_active_covering ON orders (customer_id)
INCLUDE (total, order_date)
WHERE status = 'active';
```

### Step 6: Create and Validate

```sql
-- Create concurrently to avoid locking
CREATE INDEX CONCURRENTLY idx_name ON table (columns);

-- Verify the planner uses the new index
EXPLAIN (ANALYZE, BUFFERS) <your_query>;

-- Confirm improvement
-- Compare: actual time, buffer hits/reads, plan type
```

### Step 7: Monitor Over Time

```sql
-- After a few days/weeks, check if the index is being used
SELECT indexrelname, idx_scan, idx_tup_read
FROM pg_stat_user_indexes
WHERE indexrelname = 'idx_name';

-- If idx_scan = 0 after a full business cycle, consider dropping it
DROP INDEX CONCURRENTLY idx_name;
```
