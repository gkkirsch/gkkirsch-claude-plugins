# Indexing Strategies Reference

Comprehensive cross-database indexing reference. Covers B-tree internals, hash indexes, multi-column
index design, covering indexes, partial indexes, and database-specific index types across PostgreSQL,
MySQL, MongoDB, Redis, and SQLite.

## B-tree Index Fundamentals

### How B-trees Work

A B-tree (balanced tree) is the default and most common index type across all SQL databases.

```
Structure:
                    [40 | 80]                     ← Root node
                   /    |    \
          [10|20|30]  [50|60|70]  [90|100|110]    ← Internal nodes
          / | | \    / | | \     / |  |  \
        [pages with actual data rows/pointers]     ← Leaf nodes

Properties:
- Balanced: all leaf nodes are at the same depth
- Sorted: values are ordered left to right
- Linked: leaf nodes are doubly-linked for range scans
- Logarithmic: O(log n) for point lookups
- Sequential: O(n) for range scans on sorted data
```

### B-tree Supported Operations

| Operation | Supported | Example |
|-----------|-----------|---------|
| Equality (=) | Yes | `WHERE id = 42` |
| Range (<, >, <=, >=, BETWEEN) | Yes | `WHERE price > 100` |
| IN list | Yes | `WHERE status IN ('a', 'b')` |
| IS NULL / IS NOT NULL | Yes | `WHERE deleted_at IS NULL` |
| LIKE prefix | Yes | `WHERE name LIKE 'Joh%'` |
| LIKE suffix | No | `WHERE name LIKE '%son'` — full scan |
| ORDER BY | Yes | `ORDER BY created_at DESC` |
| MIN/MAX | Yes | `SELECT MIN(price)` |
| Not equal (!=, <>) | Partial | Scans most of the index |

### Multi-Column B-tree Index Rules

The order of columns in a multi-column index is critical. The index can only be used
from left to right, in order.

```sql
CREATE INDEX idx_orders ON orders(customer_id, status, created_at);
```

**This index supports (leftmost prefix rule):**
```sql
WHERE customer_id = 42                                    -- ✅ First column
WHERE customer_id = 42 AND status = 'shipped'             -- ✅ First two columns
WHERE customer_id = 42 AND status = 'shipped'
  AND created_at > '2024-01-01'                           -- ✅ All three columns
WHERE customer_id = 42 AND created_at > '2024-01-01'      -- ⚠️ Uses customer_id only
                                                          --    (skips status)
```

**This index does NOT support:**
```sql
WHERE status = 'shipped'                                  -- ❌ Not leftmost column
WHERE created_at > '2024-01-01'                           -- ❌ Not leftmost column
WHERE status = 'shipped' AND created_at > '2024-01-01'    -- ❌ Not leftmost column
```

### The Equality-Sort-Range (ESR) Rule

When designing multi-column indexes, order columns by:
1. **E**quality columns first (WHERE col = value)
2. **S**ort columns next (ORDER BY col)
3. **R**ange columns last (WHERE col > value, col BETWEEN, col IN)

```sql
-- Query to optimize:
SELECT * FROM orders
WHERE status = 'shipped'        -- Equality
  AND total > 100               -- Range
ORDER BY created_at DESC        -- Sort
LIMIT 20;

-- OPTIMAL index (ESR order):
CREATE INDEX idx_orders_esr ON orders(status, created_at DESC, total);
-- status: equality → narrows to subtree
-- created_at: sort → already ordered, no filesort needed
-- total: range → checked last

-- SUBOPTIMAL index (wrong order):
CREATE INDEX idx_orders_bad ON orders(status, total, created_at);
-- status: equality → good
-- total: range → scans range, THEN must sort by created_at
-- Result: requires filesort
```

### Why Range Stops Index Traversal

After a range condition, subsequent columns in the index cannot be used for
further narrowing. The database must scan within the range.

```sql
CREATE INDEX idx ON orders(a, b, c);

WHERE a = 1 AND b = 2 AND c = 3   -- Uses all 3 columns (all equality)
WHERE a = 1 AND b > 2 AND c = 3   -- Uses a (equality), b (range), CANNOT use c
                                   -- Must scan all b > 2 entries, filtering c in memory
WHERE a = 1 AND b IN (2,3) AND c = 4  -- PostgreSQL treats IN as multiple equalities
                                       -- Can use all 3 columns (index skip scan)
```

## Covering Indexes (Index-Only Scans)

A covering index includes all columns needed by a query, eliminating the need
to fetch the actual table row (heap tuple).

### PostgreSQL INCLUDE Clause

```sql
-- Include non-searchable columns in the index
CREATE INDEX idx_orders_covering ON orders(customer_id, status)
    INCLUDE (total, created_at, order_number);

-- This query uses index-only scan (no heap fetch):
SELECT customer_id, status, total, created_at, order_number
FROM orders
WHERE customer_id = 42 AND status = 'shipped';

-- INCLUDE columns:
-- ✅ Returned in SELECT without heap access
-- ❌ NOT searchable (can't use in WHERE on these columns via this index)
-- ❌ NOT sortable (can't use in ORDER BY via this index)
-- ✅ Lower maintenance cost than adding them as index key columns
```

### MySQL Covering Index

```sql
-- MySQL doesn't have INCLUDE — add columns as key columns
CREATE INDEX idx_orders_covering ON orders(customer_id, status, total, created_at);

-- Query uses covering index (Extra: Using index):
SELECT customer_id, status, total, created_at
FROM orders
WHERE customer_id = 42 AND status = 'shipped';

-- MySQL 8.0.13+: InnoDB supports "invisible" columns in secondary indexes
-- that include the primary key automatically. This means:
-- If PK is `id`, an index on (customer_id) implicitly includes `id`
```

### When to Use Covering Indexes

| Scenario | Use Covering Index? | Why |
|----------|-------------------|-----|
| Hot query path (runs 1000s/sec) | Yes | Eliminates heap I/O |
| Dashboard query returning few columns | Yes | Major speedup |
| Full row fetch (SELECT *) | No | Can't cover all columns efficiently |
| Write-heavy table | Careful | Wider indexes slow writes |
| Infrequently run query | No | Index maintenance cost not worth it |

## Partial Indexes (Conditional Indexes)

Index only a subset of rows. Smaller, faster, and cheaper to maintain.

### PostgreSQL Partial Indexes

```sql
-- Index only active orders (most queries filter to these)
CREATE INDEX idx_orders_active ON orders(customer_id, created_at)
    WHERE status NOT IN ('cancelled', 'refunded', 'archived');

-- Queue pattern: index only unprocessed items
CREATE INDEX idx_jobs_pending ON jobs(priority DESC, created_at)
    WHERE status = 'pending' AND scheduled_for <= NOW();

-- Soft delete: index only non-deleted
CREATE INDEX idx_users_active_email ON users(email)
    WHERE deleted_at IS NULL;

-- Unique constraint on subset (unique active email)
CREATE UNIQUE INDEX idx_users_unique_active_email ON users(email)
    WHERE deleted_at IS NULL;

-- Index on rare values only
CREATE INDEX idx_orders_flagged ON orders(id, created_at)
    WHERE is_flagged = true;
-- If only 0.1% of orders are flagged, this index is tiny
```

### Benefits of Partial Indexes

```
1. Size: Much smaller than full index (only matching rows)
2. Maintenance: Only updated when matching rows change
3. Scan speed: Fewer index entries to scan
4. Unique constraints: Can enforce uniqueness on a subset

Limitations:
1. Query must include the WHERE condition to use the index
2. PostgreSQL only (MySQL has no partial indexes)
3. Must be mindful that the condition matches your query
```

## Unique Indexes

```sql
-- PostgreSQL
CREATE UNIQUE INDEX idx_users_email ON users(email);
-- Equivalent to: ALTER TABLE users ADD CONSTRAINT unique_email UNIQUE (email);
-- But CREATE INDEX CONCURRENTLY is available; ADD CONSTRAINT is not

-- Composite unique index
CREATE UNIQUE INDEX idx_enrollment_unique ON enrollments(student_id, course_id);

-- Unique with NULL handling
-- Standard SQL: multiple NULLs are allowed in UNIQUE columns
-- PostgreSQL: follows standard (NULLs are considered distinct)
CREATE UNIQUE INDEX idx_users_ssn ON users(ssn);
-- Multiple rows with ssn = NULL are allowed

-- PostgreSQL 15+: NULLS NOT DISTINCT
CREATE UNIQUE INDEX idx_users_ssn_strict ON users(ssn) NULLS NOT DISTINCT;
-- Only one NULL allowed

-- MySQL: NULLs are considered distinct in UNIQUE indexes
-- Multiple NULL values allowed (like PostgreSQL default)
CREATE UNIQUE INDEX idx_users_ssn ON users(ssn);
```

## Index Cardinality and Selectivity

### Cardinality

Number of distinct values in a column relative to total rows.

```sql
-- PostgreSQL: Check cardinality
SELECT
    attname AS column_name,
    n_distinct,
    CASE
        WHEN n_distinct > 0 THEN n_distinct
        WHEN n_distinct < 0 THEN ROUND(ABS(n_distinct) * reltuples)
        ELSE 0
    END AS estimated_distinct_values,
    reltuples AS total_rows
FROM pg_stats
JOIN pg_class ON pg_class.relname = pg_stats.tablename
WHERE tablename = 'orders'
ORDER BY n_distinct DESC;

-- n_distinct > 0: actual count of distinct values
-- n_distinct < 0: fraction of rows that are distinct (e.g., -0.5 = 50% unique)
```

### Selectivity

How much an index condition narrows down the result set.

```sql
-- High selectivity (good for indexing):
-- WHERE email = 'user@example.com'  → Returns 1 row out of millions
-- WHERE id = 42                     → Returns 1 row

-- Low selectivity (poor for indexing):
-- WHERE status = 'active'           → Returns 80% of rows (sequential scan faster)
-- WHERE is_deleted = false          → Returns 95% of rows
-- WHERE gender = 'M'               → Returns ~50% of rows

-- Rule of thumb:
-- < 5-15% selectivity: B-tree index is useful
-- > 15-20% selectivity: Sequential scan may be faster
-- Exception: Covering indexes are always useful regardless of selectivity
```

### When to Index Low-Cardinality Columns

```sql
-- Low cardinality alone doesn't mean "don't index"

-- GOOD: Low cardinality + high selectivity for common query value
-- If 95% of orders are 'completed' but you always query 'pending' (5%)
CREATE INDEX idx_orders_pending ON orders(created_at) WHERE status = 'pending';

-- GOOD: Low cardinality as part of composite index
CREATE INDEX idx_orders_status_customer ON orders(status, customer_id);
-- The combo of status + customer_id has high cardinality

-- BAD: Boolean column as sole index
CREATE INDEX idx_users_active ON users(is_active);
-- Almost never useful unless very few true or very few false

-- GOOD: Boolean in partial index
CREATE INDEX idx_users_active ON users(email, name) WHERE is_active = true;
```

## PostgreSQL-Specific Indexes

### GIN (Generalized Inverted Index)

```
Purpose: Multi-valued data (arrays, JSONB, full-text)
Internals: Maps each element/key to a list of row IDs (posting list)
            Like an inverted index in search engines

Best for:
- JSONB containment (@>, ?, ?|, ?&)
- Array operations (@>, &&, =)
- Full-text search (@@)
- Trigram similarity (%, ILIKE)

Not good for:
- Range queries on scalar values
- Ordering
- Prefix searches on scalar values
```

```sql
-- JSONB indexing variants
CREATE INDEX idx_gin_default ON products USING gin (attributes);
-- Supports: @>, ?, ?|, ?&
-- Size: larger

CREATE INDEX idx_gin_pathops ON products USING gin (attributes jsonb_path_ops);
-- Supports: @> only (containment)
-- Size: 2-3x smaller
-- Speed: faster for @>

-- Array indexing
CREATE INDEX idx_gin_tags ON posts USING gin (tags);
-- Supports: @>, &&, = (contains, overlaps, equals)

-- When to use GIN vs B-tree for arrays:
-- B-tree on array: supports = (exact match), <, > (sorted comparison)
-- GIN on array: supports @> (contains), && (overlaps) — what you usually want

-- Pending list optimization
ALTER INDEX idx_gin_default SET (fastupdate = on);  -- Default on
ALTER INDEX idx_gin_default SET (gin_pending_list_limit = 4096);  -- KB
-- fastupdate: batch index updates for faster writes, slower reads until VACUUM
```

### GiST (Generalized Search Tree)

```
Purpose: Spatial data, ranges, geometric types, nearest-neighbor
Internals: Balanced tree of bounding boxes/ranges that can overlap

Best for:
- PostGIS spatial queries (ST_DWithin, ST_Contains)
- Range type queries (&&, @>, <@)
- Exclusion constraints (EXCLUDE USING gist)
- Nearest-neighbor (ORDER BY point <-> reference)
- Full-text search (alternative to GIN, sometimes faster for updates)

Not good for:
- Exact equality on scalar values (use B-tree)
- Simple range queries on scalars (use B-tree)
```

```sql
-- Range exclusion constraint (no overlapping reservations)
CREATE TABLE reservations (
    id SERIAL PRIMARY KEY,
    room_id INT NOT NULL,
    during TSTZRANGE NOT NULL,
    EXCLUDE USING gist (room_id WITH =, during WITH &&)
);

-- Nearest-neighbor search (K-NN)
CREATE INDEX idx_locations ON locations USING gist (coordinates);
SELECT *, coordinates <-> point '(40.7128, -74.0060)' AS dist
FROM locations
ORDER BY coordinates <-> point '(40.7128, -74.0060)'
LIMIT 5;
-- GiST processes this efficiently: doesn't scan all rows
```

### BRIN (Block Range Index)

```
Purpose: Very large tables where data is physically ordered
Internals: Stores min/max value per block range (128 pages = ~1MB default)
Size: ~1000x smaller than B-tree

Best for:
- Time-series data (timestamp column, append-only)
- Auto-increment ID columns
- Any column with high physical correlation

Not good for:
- Randomly ordered data (low correlation)
- Point lookups (B-tree is much faster)
- Small tables (overhead not worth it)

Key metric: correlation
- Run: SELECT correlation FROM pg_stats WHERE tablename = 'x' AND attname = 'y'
- Values near 1.0 or -1.0: BRIN is effective
- Values near 0.0: BRIN is useless
```

```sql
-- Optimal BRIN usage: append-only time-series
CREATE TABLE sensor_readings (
    id BIGSERIAL PRIMARY KEY,
    sensor_id INT NOT NULL,
    value FLOAT NOT NULL,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- BRIN index with tuned pages_per_range
CREATE INDEX idx_readings_time ON sensor_readings USING brin (recorded_at)
    WITH (pages_per_range = 32);
-- Smaller pages_per_range = more granular (slightly larger index, better filtering)
-- Larger pages_per_range = less granular (smaller index, less filtering)
-- Default 128: good for most cases

-- Multi-column BRIN
CREATE INDEX idx_readings_multi ON sensor_readings USING brin (recorded_at, sensor_id);

-- Force BRIN summarization (after bulk inserts)
SELECT brin_summarize_new_values('idx_readings_time');
```

## MySQL InnoDB Indexes

### Clustered Index (Primary Key)

```
InnoDB stores data in B-tree order of the primary key.
The primary key IS the table. There is no separate heap.

Implications:
1. Range scans on PK are sequential I/O (very fast)
2. Secondary indexes store the PK value (not a row pointer)
3. Large PKs (UUIDs) make ALL secondary indexes larger
4. Random inserts (UUID v4) cause page splits and fragmentation
5. Sequential PKs (auto-increment, UUID v7) insert at the end (fast)
```

```sql
-- Good PK for InnoDB: auto-increment (sequential inserts, small)
CREATE TABLE orders (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    ...
) ENGINE=InnoDB;

-- UUID v4 PK is BAD for InnoDB (random inserts, 16 bytes in every secondary index)
-- If you must use UUID, use UUID v7 (time-ordered) or ORDERED_UUID()
-- Or use auto-increment PK + UUID unique column for external reference
```

### Secondary Indexes in InnoDB

```sql
-- Secondary index stores: (index_columns, primary_key)
CREATE INDEX idx_orders_customer ON orders(customer_id);
-- Actually stores: (customer_id, id) where id is the PK

-- This means:
-- 1. Every secondary index lookup requires a second lookup to the clustered index
-- 2. Larger PKs = larger secondary indexes
-- 3. Covering index can avoid the second lookup

-- Covering index in MySQL (includes PK implicitly)
CREATE INDEX idx_orders_customer_status ON orders(customer_id, status);
-- Query: SELECT id, customer_id, status FROM orders WHERE customer_id = 42
-- This is a covering index because id (PK) is implicitly included
-- EXPLAIN will show: Extra: Using index
```

### MySQL Index Types

```sql
-- B-tree (default)
CREATE INDEX idx_name ON users(name);

-- Full-text index
CREATE FULLTEXT INDEX idx_search ON articles(title, body);
SELECT * FROM articles WHERE MATCH(title, body) AGAINST('database optimization' IN NATURAL LANGUAGE MODE);
SELECT * FROM articles WHERE MATCH(title, body) AGAINST('+database +optimization' IN BOOLEAN MODE);

-- Spatial index
CREATE SPATIAL INDEX idx_location ON places(point);

-- Prefix index (for long string columns)
CREATE INDEX idx_email_prefix ON users(email(20));
-- Indexes only first 20 characters
-- Saves space but can't cover full-length comparisons
-- Useful for TEXT/BLOB columns that can't be fully indexed

-- Descending index (MySQL 8.0+)
CREATE INDEX idx_orders_recent ON orders(created_at DESC);

-- Invisible index (MySQL 8.0+)
ALTER TABLE orders ALTER INDEX idx_old_index INVISIBLE;
-- Index is maintained but not used by optimizer
-- Useful for testing: "what if I dropped this index?"
ALTER TABLE orders ALTER INDEX idx_old_index VISIBLE;  -- Re-enable
```

### MySQL Index Hints

```sql
-- Force MySQL to use a specific index
SELECT * FROM orders FORCE INDEX (idx_orders_customer)
WHERE customer_id = 42 AND status = 'shipped';

-- Suggest an index (optimizer may still ignore)
SELECT * FROM orders USE INDEX (idx_orders_customer)
WHERE customer_id = 42;

-- Prevent use of an index
SELECT * FROM orders IGNORE INDEX (idx_orders_status)
WHERE status = 'shipped';

-- Generally: avoid hints. Fix the root cause instead.
-- If optimizer makes wrong choice, usually: ANALYZE TABLE or check statistics
```

## MongoDB Indexes

### Basic MongoDB Indexes

```javascript
// Single field index
db.users.createIndex({ email: 1 });  // 1 = ascending, -1 = descending

// Compound index (ESR rule applies)
db.orders.createIndex({ customerId: 1, status: 1, createdAt: -1 });

// Unique index
db.users.createIndex({ email: 1 }, { unique: true });

// Sparse index (only documents with the field)
db.users.createIndex({ phone: 1 }, { sparse: true });
// Equivalent to PostgreSQL partial index WHERE phone IS NOT NULL

// TTL index (auto-delete documents after time period)
db.sessions.createIndex({ createdAt: 1 }, { expireAfterSeconds: 86400 });
// Documents deleted 24 hours after createdAt

// Background index creation
db.orders.createIndex({ total: 1 }, { background: true });
// MongoDB 4.2+: all index builds are automatically background
```

### MongoDB Compound Index Strategy

```javascript
// The compound index follows the same leftmost-prefix rule as SQL B-trees

db.orders.createIndex({ customerId: 1, status: 1, createdAt: -1 });

// Supports:
db.orders.find({ customerId: 42 })                                    // ✅
db.orders.find({ customerId: 42, status: "shipped" })                 // ✅
db.orders.find({ customerId: 42, status: "shipped" })
    .sort({ createdAt: -1 })                                          // ✅
db.orders.find({ customerId: 42 }).sort({ status: 1, createdAt: -1 }) // ✅

// Does NOT support:
db.orders.find({ status: "shipped" })                                 // ❌ (not leftmost)
db.orders.find({ createdAt: { $gt: new Date("2024-01-01") } })        // ❌ (not leftmost)
db.orders.find({ customerId: 42 }).sort({ createdAt: 1 })             // ⚠️ Sort reversed
```

### MongoDB Multikey Indexes (Arrays)

```javascript
// Automatically created when indexing a field that contains arrays
db.products.createIndex({ tags: 1 });

// Queries:
db.products.find({ tags: "electronics" });       // ✅ Single value in array
db.products.find({ tags: { $in: ["electronics", "sale"] } }); // ✅
db.products.find({ tags: { $all: ["electronics", "sale"] } }); // ✅

// Limitation: compound index can have at most ONE array field
db.orders.createIndex({ items: 1, tags: 1 });
// ERROR if both items and tags are arrays (at insert time)
```

### MongoDB Text Indexes

```javascript
// Text index for full-text search
db.articles.createIndex({ title: "text", body: "text" });

// Weighted text index
db.articles.createIndex(
  { title: "text", body: "text", tags: "text" },
  { weights: { title: 10, tags: 5, body: 1 } }
);

// Search
db.articles.find({
  $text: {
    $search: "database optimization",
    $language: "english",
    $caseSensitive: false
  }
}).sort({ score: { $meta: "textScore" } });

// Only ONE text index per collection
// For multiple search configurations, use Atlas Search (Lucene-based)
```

### MongoDB Wildcard Indexes

```javascript
// Index all fields in a subdocument (flexible schema)
db.products.createIndex({ "attributes.$**": 1 });

// Supports queries on any attribute:
db.products.find({ "attributes.color": "red" });
db.products.find({ "attributes.weight": { $gt: 100 } });

// Wildcard index on entire document
db.products.createIndex({ "$**": 1 });
// Indexes ALL fields (large index, use carefully)

// Wildcard with projection (index specific paths)
db.products.createIndex(
  { "$**": 1 },
  { wildcardProjection: { "attributes": 1, "metadata": 1 } }
);
```

### MongoDB Index Intersection

```javascript
// MongoDB can combine multiple indexes for a single query (index intersection)
db.orders.createIndex({ customerId: 1 });
db.orders.createIndex({ status: 1 });

// This query MAY use both indexes:
db.orders.find({ customerId: 42, status: "shipped" });
// MongoDB intersects the two index results

// However: a compound index is almost always better than intersection
db.orders.createIndex({ customerId: 1, status: 1 });
// This single compound index is faster and more predictable
```

## Redis as an Index

Redis data structures can serve as indexes for external data.

### Sorted Set as Index

```
# Leaderboard / ranking index
ZADD leaderboard 1500 "player:42"
ZADD leaderboard 2300 "player:17"
ZADD leaderboard 1800 "player:99"

# Get top 10 players
ZREVRANGE leaderboard 0 9 WITHSCORES

# Get rank of a player
ZREVRANK leaderboard "player:42"

# Range query
ZRANGEBYSCORE leaderboard 1000 2000 WITHSCORES
```

### Set as Tag Index

```
# Tag-based lookup
SADD tag:electronics "product:42" "product:99" "product:17"
SADD tag:sale "product:42" "product:55"

# Products with both tags (intersection)
SINTER tag:electronics tag:sale

# Products with either tag (union)
SUNION tag:electronics tag:sale

# Products with electronics but not sale
SDIFF tag:electronics tag:sale
```

### Hash as Secondary Index

```
# Index by email (email → user_id)
HSET idx:users:email "user@example.com" "user:42"

# Lookup
HGET idx:users:email "user@example.com"
# Returns "user:42", then: HGETALL user:42
```

## SQLite Indexes

### SQLite Index Characteristics

```sql
-- SQLite uses B-tree for all indexes
-- The rowid (or INTEGER PRIMARY KEY) is the implicit clustered index

-- SQLite automatically creates indexes for:
-- PRIMARY KEY (creates unique index)
-- UNIQUE constraints

-- Create index
CREATE INDEX idx_orders_customer ON orders(customer_id);

-- Unique index
CREATE UNIQUE INDEX idx_users_email ON users(email);

-- Partial index (SQLite 3.8+)
CREATE INDEX idx_orders_pending ON orders(created_at)
    WHERE status = 'pending';

-- Expression index (SQLite 3.9+)
CREATE INDEX idx_users_lower_email ON users(LOWER(email));

-- Check if index is used
EXPLAIN QUERY PLAN SELECT * FROM orders WHERE customer_id = 42;
-- SEARCH orders USING INDEX idx_orders_customer (customer_id=?)

-- List indexes
SELECT name, sql FROM sqlite_master WHERE type = 'index';
```

### SQLite Automatic Indexes

```sql
-- SQLite may create temporary automatic indexes for queries
-- This appears in EXPLAIN QUERY PLAN as:
-- SEARCH ... USING AUTOMATIC COVERING INDEX (...)

-- If you see automatic indexes frequently, create permanent ones
-- Automatic indexes are recreated for each query (expensive)

-- Disable automatic indexing (not recommended for production):
PRAGMA automatic_index = OFF;
```

## Index Maintenance

### PostgreSQL Index Maintenance

```sql
-- Rebuild an index (exclusive lock on table)
REINDEX INDEX idx_orders_customer;

-- Rebuild concurrently (no lock, PostgreSQL 12+)
REINDEX INDEX CONCURRENTLY idx_orders_customer;

-- Rebuild all indexes on a table
REINDEX TABLE orders;
REINDEX TABLE CONCURRENTLY orders;

-- Check index size
SELECT
    indexname,
    pg_size_pretty(pg_relation_size(indexname::regclass)) AS size,
    idx_scan AS usage_count
FROM pg_indexes
JOIN pg_stat_user_indexes ON indexrelname = indexname
WHERE tablename = 'orders'
ORDER BY pg_relation_size(indexname::regclass) DESC;

-- Check index health (bloat)
CREATE EXTENSION IF NOT EXISTS pgstattuple;
SELECT * FROM pgstatindex('idx_orders_customer');
-- Check: avg_leaf_density (should be >90%), leaf_fragmentation
```

### MySQL Index Maintenance

```sql
-- Analyze table (update index statistics)
ANALYZE TABLE orders;

-- Optimize table (defragment, rebuild indexes)
OPTIMIZE TABLE orders;
-- Warning: locks table during operation

-- Check index cardinality
SHOW INDEX FROM orders;

-- Check for duplicate/redundant indexes
-- MySQL 8.0+ sys schema:
SELECT * FROM sys.schema_redundant_indexes;
SELECT * FROM sys.schema_unused_indexes;
```

### MongoDB Index Maintenance

```javascript
// List indexes
db.orders.getIndexes()

// Drop an index
db.orders.dropIndex("idx_orders_customer_1")
db.orders.dropIndex({ customerId: 1 })  // By key pattern

// Compact a collection (rebuild indexes)
db.runCommand({ compact: "orders" })

// Index stats
db.orders.aggregate([{ $indexStats: {} }])
// Shows: accesses.ops (usage count), accesses.since (since when)

// Find unused indexes
db.orders.aggregate([
  { $indexStats: {} },
  { $match: { "accesses.ops": 0 } }
])
```

## Anti-Patterns

### Over-Indexing

```
Problem: Too many indexes on a table
Symptoms:
- Slow INSERT/UPDATE/DELETE
- High disk usage
- Each write updates ALL indexes

Fix:
1. Remove unused indexes
2. Consolidate overlapping indexes
3. Use covering indexes instead of multiple single-column indexes
4. Consider partial indexes for filtered queries
```

### Under-Indexing

```
Problem: Missing indexes for common query patterns
Symptoms:
- Sequential scans on large tables
- Slow queries despite low data volume
- High disk I/O

Fix:
1. Check EXPLAIN plans for sequential scans
2. Look at pg_stat_user_tables.seq_scan vs idx_scan
3. Add indexes for foreign keys (not automatic in PostgreSQL!)
4. Add indexes for WHERE, JOIN, and ORDER BY columns
```

### Index Column Order Mistakes

```sql
-- BAD: Wrong column order for the query
CREATE INDEX idx_orders ON orders(created_at, customer_id);
-- For: WHERE customer_id = 42 ORDER BY created_at
-- The index can't be used because customer_id is not first

-- GOOD: Match column order to query pattern
CREATE INDEX idx_orders ON orders(customer_id, created_at);
-- Equality column first, sort column second
```

### Indexing Low-Selectivity Columns Alone

```sql
-- BAD: Boolean column as primary index column
CREATE INDEX idx_users_active ON users(is_active);
-- If 95% of users are active, this index is almost useless for WHERE is_active = true
-- Sequential scan is faster for the majority case

-- GOOD: Partial index on the minority case
CREATE INDEX idx_users_inactive ON users(id, name) WHERE is_active = false;
-- Small index, covers the rare query case
```

### Redundant Indexes

```sql
-- These indexes are redundant:
CREATE INDEX idx_a ON orders(customer_id);
CREATE INDEX idx_b ON orders(customer_id, status);
CREATE INDEX idx_c ON orders(customer_id, status, created_at);

-- idx_a is redundant: idx_b and idx_c both start with customer_id
-- idx_b MIGHT be redundant if queries always include all three columns
-- But idx_b is useful if queries only filter on customer_id + status without created_at

-- Rule: A shorter index is redundant if a longer index has the same leading columns
-- Exception: shorter covering indexes may still be useful for index-only scans
```

## Decision Framework

### Should I Add an Index?

```
1. Is the query slow? → EXPLAIN ANALYZE first
2. Is it a sequential scan on a large table? → Likely needs index
3. What's the selectivity? → <15% → B-tree useful
4. Is it a frequently run query? → Worth the write overhead
5. Can I use a partial index? → Smaller, faster, less write overhead
6. Can I combine with an existing index? → Modify instead of adding
7. What's the write volume? → High writes = fewer indexes
```

### Index Type Selection

| Query Pattern | PostgreSQL | MySQL | MongoDB | SQLite |
|---------------|-----------|-------|---------|--------|
| Equality (=) | B-tree | B-tree | Single/Compound | B-tree |
| Range (<, >, BETWEEN) | B-tree | B-tree | Single/Compound | B-tree |
| Sorting (ORDER BY) | B-tree | B-tree | Compound | B-tree |
| Pattern (LIKE 'x%') | B-tree | B-tree | Text/Regex | B-tree |
| Full-text search | GIN + tsvector | FULLTEXT | Text | FTS5 |
| JSONB containment | GIN | Generated column + B-tree | Wildcard | json_extract + B-tree |
| Array contains | GIN | N/A | Multikey | N/A |
| Spatial (distance) | GiST + PostGIS | SPATIAL | 2dsphere | R-tree |
| Time-series (ordered) | BRIN | B-tree (clustered PK) | Compound (time) | B-tree |
| Fuzzy/similarity | GIN + pg_trgm | FULLTEXT | Atlas Search | FTS5 |
| Vector/embedding | HNSW/IVFFlat (pgvector) | N/A | Atlas Vector Search | N/A |
