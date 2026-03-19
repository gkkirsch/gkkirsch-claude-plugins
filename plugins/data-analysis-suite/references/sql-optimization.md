# SQL Optimization Reference

A comprehensive reference manual for query optimization, indexing strategies, and database performance tuning. All examples use PostgreSQL unless explicitly noted otherwise.

---

## Query Plan Analysis

Understanding execution plans is the single most important skill for SQL optimization. Every optimization decision should be validated with EXPLAIN.

### PostgreSQL EXPLAIN

There are three levels of plan inspection, each revealing more detail.

**EXPLAIN** shows the planner's predicted execution strategy without running the query:

```sql
EXPLAIN SELECT * FROM orders WHERE status = 'pending' AND created_at > '2025-01-01';
```

Output:

```
                                      QUERY PLAN
------------------------------------------------------------------------------------
 Index Scan using idx_orders_status_date on orders  (cost=0.43..812.56 rows=2341 width=128)
   Index Cond: ((status = 'pending'::text) AND (created_at > '2025-01-01'::timestamp))
```

**EXPLAIN ANALYZE** actually executes the query and shows real timing and row counts:

```sql
EXPLAIN ANALYZE SELECT * FROM orders WHERE status = 'pending' AND created_at > '2025-01-01';
```

Output:

```
                                      QUERY PLAN
------------------------------------------------------------------------------------
 Index Scan using idx_orders_status_date on orders
   (cost=0.43..812.56 rows=2341 width=128)
   (actual time=0.028..3.412 rows=2847 loops=1)
   Index Cond: ((status = 'pending'::text) AND (created_at > '2025-01-01'::timestamp))
 Planning Time: 0.152 ms
 Execution Time: 3.891 ms
```

**EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON)** gives maximum detail including buffer (cache) statistics in structured JSON:

```sql
EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON)
SELECT * FROM orders WHERE status = 'pending' AND created_at > '2025-01-01';
```

This reveals shared buffer hits (data found in PostgreSQL's cache) versus reads (data fetched from the OS/disk), which is critical for understanding I/O behavior.

> **Warning**: EXPLAIN ANALYZE actually executes the query. For INSERT, UPDATE, or DELETE, wrap in a transaction and roll back:
> ```sql
> BEGIN;
> EXPLAIN ANALYZE DELETE FROM orders WHERE id < 100;
> ROLLBACK;
> ```

---

### Reading Plan Nodes

Every EXPLAIN output is a tree of plan nodes. Data flows from the innermost (deepest indented) nodes upward to the root.

#### Seq Scan (Sequential Scan)

Reads every row in the table. The planner chooses this when:

- The table is very small (a few pages).
- The query returns a large fraction of rows (typically >5-10%).
- No suitable index exists.
- Statistics are stale (run `ANALYZE`).

```
Seq Scan on users  (cost=0.00..1834.00 rows=100000 width=64)
  Filter: (active = true)
  Rows Removed by Filter: 85000
```

**When acceptable**: Small lookup tables, full-table analytics queries, data warehousing scans.

**When problematic**: Frequently executed OLTP queries on large tables. The `Rows Removed by Filter` line is the red flag: if the scan reads 100,000 rows but keeps only 15,000, an index would likely help.

#### Index Scan

Traverses a B-tree index to find matching rows, then fetches the full tuple from the heap (table) for each match.

```
Index Scan using idx_users_email on users  (cost=0.42..8.44 rows=1 width=64)
  Index Cond: (email = 'user@example.com'::text)
```

The planner chooses this when the index is selective enough that the random I/O cost of heap fetches is less than a sequential scan. On SSDs, the threshold for choosing index scans is lower because random reads are cheap.

#### Index Only Scan

Satisfies the query entirely from the index without visiting the heap. This requires:

1. All columns in SELECT, WHERE, and ORDER BY are in the index.
2. The visibility map indicates the heap pages are all-visible (recent VACUUM).

```
Index Only Scan using idx_orders_covering on orders
  (cost=0.43..156.78 rows=2341 width=12)
  Index Cond: (status = 'pending'::text)
  Heap Fetches: 0
```

`Heap Fetches: 0` is the ideal. If this number is high, run `VACUUM` on the table.

#### Bitmap Heap Scan

A two-phase approach: first, one or more Bitmap Index Scans build a bitmap of matching heap pages, then the Bitmap Heap Scan fetches those pages in physical order (reducing random I/O).

```
Bitmap Heap Scan on orders  (cost=124.56..2345.67 rows=5000 width=128)
  Recheck Cond: ((status = 'pending'::text) OR (priority = 'high'::text))
  ->  BitmapOr  (cost=124.56..124.56 rows=5000 width=0)
        ->  Bitmap Index Scan on idx_orders_status  (cost=0.00..62.28 rows=3000 width=0)
              Index Cond: (status = 'pending'::text)
        ->  Bitmap Index Scan on idx_orders_priority  (cost=0.00..62.28 rows=2000 width=0)
              Index Cond: (priority = 'high'::text)
```

This is how PostgreSQL combines multiple indexes with BitmapAnd/BitmapOr. The `Recheck Cond` happens because the bitmap is lossy when it exceeds `work_mem`.

#### Nested Loop

Iterates over the outer (top) input and, for each row, scans the inner (bottom) input. Efficient when:

- The outer input is small.
- The inner input has an index on the join column.

```
Nested Loop  (cost=0.43..1245.67 rows=500 width=192)
  ->  Seq Scan on order_items  (cost=0.00..25.00 rows=500 width=64)
        Filter: (quantity > 10)
  ->  Index Scan using orders_pkey on orders  (cost=0.43..2.44 rows=1 width=128)
        Index Cond: (id = order_items.order_id)
```

Cost is proportional to `outer_rows * inner_scan_cost`. With 500 outer rows and an index lookup costing ~2.44 each, total is manageable.

#### Hash Join

Builds a hash table from the smaller input, then probes it with rows from the larger input. Used for equi-joins when neither input is pre-sorted and the hash table fits in `work_mem`.

```
Hash Join  (cost=305.00..2456.78 rows=10000 width=192)
  Hash Cond: (orders.customer_id = customers.id)
  ->  Seq Scan on orders  (cost=0.00..1834.00 rows=100000 width=128)
  ->  Hash  (cost=205.00..205.00 rows=10000 width=64)
        Buckets: 16384  Batches: 1  Memory Usage: 640kB
        ->  Seq Scan on customers  (cost=0.00..205.00 rows=10000 width=64)
```

Watch `Batches`. If Batches > 1, the hash table spilled to disk. Increase `work_mem` to fix this.

#### Merge Join

Requires both inputs sorted on the join key (or uses index scans that produce sorted output). Efficient for large equi-joins when data is already sorted.

```
Merge Join  (cost=0.86..45678.90 rows=500000 width=192)
  Merge Cond: (orders.id = order_items.order_id)
  ->  Index Scan using orders_pkey on orders  (cost=0.43..12345.67 rows=500000 width=128)
  ->  Index Scan using idx_items_order_id on order_items  (cost=0.43..23456.78 rows=1000000 width=64)
```

#### Sort

PostgreSQL uses quicksort for in-memory sorts and external merge sort when data exceeds `work_mem`.

```
Sort  (cost=1234.56..1259.56 rows=10000 width=64)
  Sort Key: created_at DESC
  Sort Method: quicksort  Memory: 1024kB
```

vs.

```
Sort  (cost=12345.67..12370.67 rows=100000 width=64)
  Sort Key: created_at DESC
  Sort Method: external merge  Disk: 8192kB
```

External merge (disk sort) is much slower. Increase `work_mem` or add an index that provides the desired sort order.

#### Aggregate

**HashAggregate** builds a hash table of groups. Fast when the number of groups fits in memory:

```
HashAggregate  (cost=2345.00..2445.00 rows=100 width=40)
  Group Key: status
  Batches: 1  Memory Usage: 40kB
```

**GroupAggregate** requires pre-sorted input but handles any number of groups without memory pressure:

```
GroupAggregate  (cost=0.43..5678.90 rows=100 width=40)
  Group Key: status
  ->  Index Scan using idx_orders_status on orders  (cost=0.43..4567.89 rows=100000 width=12)
```

#### Materialize

Caches the result of a subplan for reuse. Appears when the planner needs to re-scan an inner loop input or when a CTE is materialized.

```
Materialize  (cost=0.00..25.50 rows=500 width=64)
  ->  Seq Scan on lookup_table  (cost=0.00..23.00 rows=500 width=64)
```

---

### Cost Model

Every plan node shows `(cost=STARTUP..TOTAL rows=ROWS width=WIDTH)`.

| Field | Meaning |
|-------|---------|
| **Startup cost** | Cost before the first row can be returned (e.g., sorting must finish before output begins) |
| **Total cost** | Cost to return all rows |
| **Rows** | Estimated number of rows output by this node |
| **Width** | Estimated average row size in bytes |

Costs are in arbitrary units calibrated to sequential page reads (`seq_page_cost = 1.0`). They are not milliseconds.

### Actual vs. Estimated Rows

The most common source of bad plans is row estimation errors. Compare `rows=` (estimated) with `actual ... rows=` (actual):

```
Seq Scan on events  (cost=0.00..25000.00 rows=100 width=64)
                    (actual time=0.015..245.678 rows=48753 loops=1)
  Filter: (event_type = 'click' AND created_at > '2025-06-01')
  Rows Removed by Filter: 951247
```

The planner estimated 100 rows but got 48,753. This 487x underestimate could cause the planner to choose nested loop joins where hash joins would be far better. Fixes:

1. Run `ANALYZE events;` to update statistics.
2. Increase `default_statistics_target` for columns with skewed distributions.
3. Create extended statistics: `CREATE STATISTICS events_type_date (dependencies) ON event_type, created_at FROM events;`

### Buffer Analysis

With `BUFFERS` enabled:

```
Index Scan using idx_orders_status on orders
  (actual time=0.028..3.412 rows=2847 loops=1)
  Buffers: shared hit=856 read=12
```

| Metric | Meaning |
|--------|---------|
| `shared hit` | Pages found in PostgreSQL's shared buffer cache |
| `shared read` | Pages read from the OS (possibly from OS page cache, possibly from disk) |
| `shared dirtied` | Pages modified during query execution |
| `shared written` | Dirty pages flushed to disk during query |
| `temp read/written` | Temp file I/O (sorts, hash joins spilling to disk) |

A high `shared read` relative to `shared hit` indicates poor cache utilization. A query that consistently shows `temp read/written` is spilling to disk and may benefit from increased `work_mem`.

---

### Full EXPLAIN ANALYZE Example with Commentary

```sql
EXPLAIN (ANALYZE, BUFFERS)
SELECT
    c.name,
    COUNT(o.id) AS order_count,
    SUM(o.total_amount) AS total_spent
FROM customers c
JOIN orders o ON o.customer_id = c.id
WHERE c.region = 'US'
  AND o.created_at >= '2025-01-01'
GROUP BY c.name
ORDER BY total_spent DESC
LIMIT 10;
```

```
Limit  (cost=4567.89..4567.91 rows=10 width=48)
       (actual time=45.123..45.130 rows=10 loops=1)
  Buffers: shared hit=3456 read=234
  ->  Sort  (cost=4567.89..4580.12 rows=4892 width=48)
           (actual time=45.120..45.125 rows=10 loops=1)
        Sort Key: (sum(o.total_amount)) DESC
        Sort Method: top-N heapsort  Memory: 26kB
        Buffers: shared hit=3456 read=234
        ->  HashAggregate  (cost=4234.56..4345.67 rows=4892 width=48)
                          (actual time=42.345..43.567 rows=4892 loops=1)
              Group Key: c.name
              Batches: 1  Memory Usage: 625kB
              Buffers: shared hit=3456 read=234
              ->  Hash Join  (cost=305.00..4012.34 rows=14876 width=40)
                            (actual time=2.345..35.678 rows=15234 loops=1)
                    Hash Cond: (o.customer_id = c.id)
                    Buffers: shared hit=3456 read=234
                    ->  Index Scan using idx_orders_created on orders o
                          (cost=0.43..3456.78 rows=85432 width=20)
                          (actual time=0.025..18.234 rows=87654 loops=1)
                          Index Cond: (created_at >= '2025-01-01'::timestamp)
                          Buffers: shared hit=3200 read=234
                    ->  Hash  (cost=254.57..254.57 rows=4000 width=28)
                             (actual time=2.123..2.123 rows=4123 loops=1)
                          Buckets: 8192  Batches: 1  Memory Usage: 245kB
                          Buffers: shared hit=256
                          ->  Seq Scan on customers c
                                (cost=0.00..254.57 rows=4000 width=28)
                                (actual time=0.012..1.567 rows=4123 loops=1)
                                Filter: (region = 'US'::text)
                                Rows Removed by Filter: 5877
                                Buffers: shared hit=256
 Planning Time: 0.456 ms
 Execution Time: 45.234 ms
```

**Line-by-line reading (bottom-up):**

1. **Seq Scan on customers**: Reads all 10,000 customers, filters to 4,123 US customers. Entirely from cache (shared hit=256, no reads). The filter removes 5,877 rows. An index on `region` could help if this were more selective.

2. **Hash**: Builds a hash table of 4,123 US customers. Fits in 245kB (well under default `work_mem`). Single batch means no disk spill.

3. **Index Scan on orders**: Uses the `idx_orders_created` index to find orders since 2025-01-01. Returns 87,654 rows. Most buffers come from here (3,200 hits + 234 reads). The 234 reads suggest some order data isn't cached.

4. **Hash Join**: Joins 87,654 orders against the 4,123-customer hash table, producing 15,234 matching rows. The estimate (14,876) was close to actual (15,234) -- good statistics.

5. **HashAggregate**: Groups 15,234 rows into 4,892 groups by customer name. Single batch, 625kB memory. No disk spill.

6. **Sort**: Sorts 4,892 rows by total_spent DESC. Uses top-N heapsort since only 10 rows are needed (LIMIT 10). Only 26kB memory. Very efficient.

7. **Limit**: Returns the top 10 rows.

**Total execution: 45ms.** The main cost is the index scan on orders (18ms of the 45ms). The 234 buffer reads are the only I/O concern -- on a cold cache this could be slower.

---

### MySQL EXPLAIN

MySQL's EXPLAIN output uses a tabular format with different columns.

```sql
EXPLAIN SELECT * FROM orders WHERE customer_id = 42 AND status = 'shipped';
```

```
+----+-------------+--------+------+-----------------------+---------+---------+-------+------+-------------+
| id | select_type | table  | type | possible_keys         | key     | key_len | ref   | rows | Extra       |
+----+-------------+--------+------+-----------------------+---------+---------+-------+------+-------------+
|  1 | SIMPLE      | orders | ref  | idx_cust_id,idx_status| idx_cust| 4       | const |   15 | Using where |
+----+-------------+--------+------+-----------------------+---------+---------+-------+------+-------------+
```

#### Access Types (best to worst)

| Type | Meaning | Performance |
|------|---------|-------------|
| `system` | Table has exactly one row | Fastest |
| `const` | At most one matching row (PRIMARY KEY or UNIQUE lookup) | Excellent |
| `eq_ref` | One row from this table for each row from previous tables (PK/UNIQUE join) | Excellent |
| `ref` | All rows with matching index value | Good |
| `range` | Index range scan (BETWEEN, <, >, IN) | Good |
| `index` | Full index scan (reads every entry in the index) | Mediocre |
| `ALL` | Full table scan | Worst |

#### key_len Calculation

`key_len` tells you how many bytes of a composite index are being used. This reveals whether all parts of a composite index are utilized.

For a composite index `(customer_id INT, status VARCHAR(20), created_at TIMESTAMP)`:

| Column | Type | Bytes | Nullable adds |
|--------|------|-------|---------------|
| customer_id | INT | 4 | +1 if nullable |
| status | VARCHAR(20) | 20*3+2=62 (utf8mb3) or 20*4+2=82 (utf8mb4) | +1 if nullable |
| created_at | TIMESTAMP | 4 | +1 if nullable |

If `key_len = 4`, only `customer_id` is used. If `key_len = 66`, both `customer_id` and `status` are used (assuming utf8mb3, NOT NULL). This helps diagnose whether the query takes full advantage of the index.

#### Extra Field Values

| Value | Meaning | Action |
|-------|---------|--------|
| `Using index` | Covering index, no table access needed | Great -- keep it |
| `Using where` | Server filters rows after storage engine retrieval | May need better index |
| `Using temporary` | Requires a temporary table (GROUP BY, DISTINCT, UNION) | Consider index for grouping |
| `Using filesort` | Requires an extra sort pass | Add index for ORDER BY |
| `Using index condition` | Index condition pushdown (ICP) | Good -- pushed filter to storage |
| `Using join buffer` | Block nested loop join, no index on join column | Add index on join column |

#### MySQL EXPLAIN FORMAT=JSON

```sql
EXPLAIN FORMAT=JSON SELECT * FROM orders WHERE customer_id = 42\G
```

The JSON format reveals the query cost model:

```json
{
  "query_block": {
    "select_id": 1,
    "cost_info": {
      "query_cost": "15.41"
    },
    "table": {
      "table_name": "orders",
      "access_type": "ref",
      "key": "idx_customer_id",
      "rows_examined_per_scan": 15,
      "rows_produced_per_join": 15,
      "cost_info": {
        "read_cost": "12.41",
        "eval_cost": "3.00",
        "prefix_cost": "15.41"
      }
    }
  }
}
```

---

## Indexing Strategies

### B-tree Indexes

B-tree is the default and most common index type. Understanding its structure is essential.

#### How B-trees Work

A B-tree consists of a root page, internal pages, and leaf pages:

```
        [Root: 50]
       /          \
  [Internal: 20,35]  [Internal: 65,80]
  /     |     \       /     |     \
[Leaf] [Leaf] [Leaf] [Leaf] [Leaf] [Leaf]
 1-19   20-34  35-49  50-64  65-79  80-99
```

- **Leaf pages** store index entries sorted by key value, with pointers to heap tuples.
- **Internal pages** store key values and pointers to child pages.
- **Tree height** is typically 3-4 for tables up to billions of rows. Each additional level adds one I/O for a lookup.

A B-tree on a 10-million-row table is typically 3 levels deep. A point lookup does 3 page reads (often cached). This is why index lookups are O(log N) and feel nearly constant for practical table sizes.

#### Operations Supported

B-trees support these operations efficiently:

```sql
-- Equality
WHERE email = 'user@example.com'

-- Range
WHERE created_at BETWEEN '2025-01-01' AND '2025-06-30'
WHERE price > 100
WHERE age <= 30

-- IN lists (treated as multiple equality checks)
WHERE status IN ('pending', 'processing', 'shipped')

-- IS NULL / IS NOT NULL
WHERE deleted_at IS NULL

-- Prefix LIKE (only left-anchored)
WHERE name LIKE 'John%'      -- Uses index
-- WHERE name LIKE '%John%'  -- Cannot use B-tree index

-- ORDER BY (if matching index order)
ORDER BY created_at DESC
```

#### Composite Index Column Order: The Equality-Range-Sort Rule

The column order in a composite index determines which queries can use it. Follow this rule:

1. **Equality columns first** (columns used with `=` or `IN`)
2. **Range/sort column last** (column used with `<`, `>`, `BETWEEN`, or `ORDER BY`)

```sql
-- Query pattern:
SELECT * FROM orders
WHERE status = 'pending'          -- equality
  AND customer_id = 42            -- equality
  AND created_at > '2025-01-01'   -- range
ORDER BY created_at;              -- sort

-- GOOD: equality columns first, range/sort column last
CREATE INDEX idx_orders_status_cust_date
ON orders(status, customer_id, created_at);

-- BAD: range column in the middle breaks the rest
CREATE INDEX idx_orders_bad
ON orders(status, created_at, customer_id);
-- With this index, customer_id cannot be used for filtering
-- because created_at (a range condition) comes before it.
```

**Why order matters**: A B-tree can only use index columns left-to-right until it hits a range condition. After a range condition, subsequent columns cannot be used for filtering (only for sorting in some cases).

```sql
-- This index: (status, customer_id, created_at)
-- Can satisfy these query patterns:
WHERE status = 'pending'                                       -- 1 column
WHERE status = 'pending' AND customer_id = 42                  -- 2 columns
WHERE status = 'pending' AND customer_id = 42 AND created_at > -- 3 columns
WHERE status = 'pending' AND customer_id IN (1,2,3)            -- 2 columns

-- Cannot efficiently satisfy:
WHERE customer_id = 42                    -- skips first column
WHERE created_at > '2025-01-01'           -- skips first two columns
WHERE status = 'pending' AND created_at > -- skips second column
```

#### Index Selectivity and Cardinality

**Selectivity** = number of distinct values / total number of rows. Higher selectivity means the index is more useful.

```sql
-- Check selectivity of columns:
SELECT
    'status' AS column_name,
    COUNT(DISTINCT status)::float / COUNT(*) AS selectivity
FROM orders
UNION ALL
SELECT
    'customer_id',
    COUNT(DISTINCT customer_id)::float / COUNT(*)
FROM orders
UNION ALL
SELECT
    'email',
    COUNT(DISTINCT email)::float / COUNT(*)
FROM users;
```

| Column | Distinct Values | Total Rows | Selectivity |
|--------|----------------|------------|-------------|
| status | 5 | 1,000,000 | 0.000005 |
| customer_id | 50,000 | 1,000,000 | 0.05 |
| email | 100,000 | 100,000 | 1.0 |

An index on `email` is excellent (unique). An index on `status` alone is poor (only 5 values). But `status` is still useful as the first column in a composite index if queries always filter on it.

---

### Covering Indexes

A covering index includes all columns needed by a query, enabling Index Only Scans.

#### INCLUDE Columns (PostgreSQL 11+)

```sql
-- This query:
SELECT id, status, total_amount
FROM orders
WHERE status = 'pending' AND created_at > '2025-01-01';

-- Covering index with INCLUDE:
CREATE INDEX idx_orders_covering ON orders(status, created_at)
INCLUDE (id, total_amount);
```

The `INCLUDE` columns are stored in the leaf pages but are not part of the search key. This means:

- They do not increase the index tree height (they are not in internal pages).
- They cannot be used for filtering or sorting.
- They make Index Only Scans possible.

```sql
EXPLAIN SELECT id, status, total_amount
FROM orders
WHERE status = 'pending' AND created_at > '2025-01-01';
```

```
Index Only Scan using idx_orders_covering on orders
  (cost=0.43..156.78 rows=2341 width=20)
  Index Cond: ((status = 'pending'::text) AND (created_at > '2025-01-01'::timestamp))
  Heap Fetches: 0
```

#### When to Use INCLUDE vs. Extending the Key

```sql
-- If you only filter/sort on (status, created_at) but also need id and total_amount:
-- USE INCLUDE:
CREATE INDEX idx_a ON orders(status, created_at) INCLUDE (id, total_amount);

-- If you also filter on total_amount:
-- EXTEND THE KEY:
CREATE INDEX idx_b ON orders(status, created_at, total_amount);
-- But now total_amount is in internal pages too, making the index larger.
```

---

### Partial Indexes (PostgreSQL)

A partial index covers only a subset of rows, defined by a WHERE clause.

```sql
-- Only index pending orders (maybe 2% of total):
CREATE INDEX idx_orders_pending
ON orders(created_at)
WHERE status = 'pending';

-- Only index active users:
CREATE INDEX idx_users_active_email
ON users(email)
WHERE active = true;

-- Only index non-null values:
CREATE INDEX idx_events_error
ON events(created_at, message)
WHERE error_code IS NOT NULL;
```

**Advantages**:

- Dramatically smaller index (only includes matching rows).
- Faster to build, faster to maintain, less I/O.
- On a 10-million-row orders table where 200,000 are pending, the partial index is ~50x smaller than a full index.

**Critical requirement**: The query's WHERE clause must match or imply the index's WHERE clause. PostgreSQL must be able to prove at plan time that the query only accesses rows covered by the index.

```sql
-- This query WILL use the partial index:
SELECT * FROM orders WHERE status = 'pending' AND created_at > '2025-01-01';

-- This query WILL NOT use the partial index:
SELECT * FROM orders WHERE created_at > '2025-01-01';
-- (status might not be 'pending', so the index doesn't cover all needed rows)
```

---

### Expression Indexes

Index the result of an expression or function call.

```sql
-- Case-insensitive email lookup:
CREATE INDEX idx_users_lower_email ON users(LOWER(email));

-- Use it:
SELECT * FROM users WHERE LOWER(email) = 'user@example.com';

-- Date truncation for daily aggregates:
CREATE INDEX idx_events_day ON events(date_trunc('day', created_at));

-- JSONB field extraction:
CREATE INDEX idx_users_profile_city ON users((profile->>'city'));

-- Use it:
SELECT * FROM users WHERE profile->>'city' = 'New York';

-- Computed column:
CREATE INDEX idx_orders_total ON orders((quantity * unit_price));
```

**Requirement**: The function must be `IMMUTABLE` -- it must always return the same output for the same input. Functions like `NOW()` or `random()` cannot be used.

The query must use the exact same expression as the index. `WHERE LOWER(email) = 'x'` matches the index, but `WHERE email = 'X'` does not.

---

### GIN Indexes (Generalized Inverted Index)

GIN indexes are designed for values that contain multiple elements (arrays, JSONB, full-text search vectors).

#### Full-Text Search

```sql
-- Add a tsvector column and GIN index:
ALTER TABLE articles ADD COLUMN search_vector tsvector
    GENERATED ALWAYS AS (to_tsvector('english', title || ' ' || body)) STORED;

CREATE INDEX idx_articles_search ON articles USING GIN(search_vector);

-- Query:
SELECT title, ts_rank(search_vector, query) AS rank
FROM articles, to_tsquery('english', 'postgres & optimization') AS query
WHERE search_vector @@ query
ORDER BY rank DESC
LIMIT 10;
```

#### JSONB Operations

```sql
-- GIN index on entire JSONB column:
CREATE INDEX idx_events_data ON events USING GIN(data);

-- Supports these operators:
-- Containment (@>):
SELECT * FROM events WHERE data @> '{"type": "click", "page": "/home"}';

-- Key existence (?):
SELECT * FROM events WHERE data ? 'error_code';

-- Any key existence (?|):
SELECT * FROM events WHERE data ?| array['error_code', 'warning'];

-- All keys exist (?&):
SELECT * FROM events WHERE data ?& array['user_id', 'session_id'];

-- For specific path queries, use jsonb_path_ops (smaller, faster for @>):
CREATE INDEX idx_events_data_path ON events USING GIN(data jsonb_path_ops);
-- Only supports @>, but index is ~3x smaller.
```

#### Array Operations

```sql
CREATE INDEX idx_posts_tags ON posts USING GIN(tags);

-- Array containment:
SELECT * FROM posts WHERE tags @> ARRAY['postgresql', 'performance'];

-- Array overlap:
SELECT * FROM posts WHERE tags && ARRAY['postgresql', 'mysql'];
```

#### Trigram Similarity (pg_trgm)

```sql
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX idx_products_name_trgm ON products USING GIN(name gin_trgm_ops);

-- Now LIKE '%keyword%' and similarity queries use the index:
SELECT * FROM products WHERE name LIKE '%widget%';
SELECT * FROM products WHERE name % 'widjet';  -- fuzzy match
SELECT * FROM products
WHERE similarity(name, 'widjet') > 0.3
ORDER BY similarity(name, 'widjet') DESC;
```

---

### GiST Indexes (Generalized Search Tree)

GiST supports overlapping and nearest-neighbor queries.

```sql
-- Range types:
CREATE INDEX idx_reservations_range ON reservations USING GiST(daterange(check_in, check_out));

SELECT * FROM reservations
WHERE daterange(check_in, check_out) && daterange('2025-06-01', '2025-06-15');

-- PostGIS spatial queries:
CREATE INDEX idx_locations_geom ON locations USING GiST(geom);

SELECT name, ST_Distance(geom, ST_MakePoint(-73.9857, 40.7484)::geography) AS distance
FROM locations
WHERE ST_DWithin(geom, ST_MakePoint(-73.9857, 40.7484)::geography, 5000)
ORDER BY distance
LIMIT 10;

-- Full-text search (alternative to GIN, better for frequent updates):
CREATE INDEX idx_articles_search_gist ON articles USING GiST(search_vector);
```

GiST vs GIN for full-text search:
- **GIN**: Faster reads, slower writes. Better for mostly-read workloads.
- **GiST**: Faster writes, slower reads. Better for write-heavy workloads.

---

### BRIN Indexes (Block Range Index)

BRIN stores min/max values for ranges of physical table pages. Extremely small indexes for naturally ordered data.

```sql
-- Ideal for time-series data inserted in order:
CREATE INDEX idx_events_created_brin ON events USING BRIN(created_at)
WITH (pages_per_range = 32);

-- Also good for sequential IDs:
CREATE INDEX idx_logs_id_brin ON logs USING BRIN(id);
```

**When to use**: Data with high physical correlation -- values that increase naturally with insertion order (timestamps, auto-increment IDs). Check correlation:

```sql
SELECT correlation FROM pg_stats
WHERE tablename = 'events' AND attname = 'created_at';
-- correlation close to 1.0 or -1.0 = BRIN will work well
-- correlation close to 0.0 = BRIN will be ineffective
```

**Size comparison** on a 100-million-row table with a timestamp column:

| Index Type | Size |
|-----------|------|
| B-tree | ~2.1 GB |
| BRIN (pages_per_range=128) | ~200 KB |

BRIN is 10,000x smaller, but scans are less precise (false positives require recheck).

**pages_per_range tuning**: Lower values = more precise but larger index. Higher values = smaller but less precise. Default is 128. For highly correlated data, 32 works well.

---

### Index Maintenance

#### REINDEX CONCURRENTLY

```sql
-- Rebuild an index without blocking writes (PostgreSQL 12+):
REINDEX INDEX CONCURRENTLY idx_orders_status;

-- Rebuild all indexes on a table:
REINDEX TABLE CONCURRENTLY orders;
```

#### Monitoring Index Usage

```sql
-- Find unused indexes (candidates for removal):
SELECT
    schemaname || '.' || relname AS table,
    indexrelname AS index,
    pg_size_pretty(pg_relation_size(indexrelid)) AS index_size,
    idx_scan AS times_used,
    idx_tup_read AS tuples_read,
    idx_tup_fetch AS tuples_fetched
FROM pg_stat_user_indexes
WHERE idx_scan = 0
  AND indexrelid NOT IN (
      SELECT conindid FROM pg_constraint
      WHERE contype IN ('p', 'u')  -- keep primary keys and unique constraints
  )
ORDER BY pg_relation_size(indexrelid) DESC;
```

#### Detecting Index Bloat

```sql
-- Estimate index bloat using pgstattuple:
CREATE EXTENSION IF NOT EXISTS pgstattuple;

SELECT
    indexrelname,
    pg_size_pretty(pg_relation_size(indexrelid)) AS index_size,
    100 - (avg_leaf_density) AS bloat_pct
FROM pg_stat_user_indexes
JOIN pgstatindex(indexrelname) ON true
WHERE avg_leaf_density < 70
ORDER BY pg_relation_size(indexrelid) DESC;
```

#### Concurrent Index Creation

```sql
-- Create index without locking the table for writes:
CREATE INDEX CONCURRENTLY idx_orders_email ON orders(email);

-- NOTE: CONCURRENTLY cannot run inside a transaction block.
-- If it fails partway, it leaves an INVALID index. Check and clean up:
SELECT indexrelid::regclass, indisvalid
FROM pg_index
WHERE NOT indisvalid;

-- Drop the invalid index and retry:
DROP INDEX CONCURRENTLY idx_orders_email;
CREATE INDEX CONCURRENTLY idx_orders_email ON orders(email);
```

---

## Join Optimization

### Join Types and Performance

#### INNER JOIN

Returns only rows that match in both tables. The planner can reorder INNER JOINs freely.

```sql
-- These are equivalent; the planner will choose the best order:
SELECT o.*, c.name
FROM orders o
INNER JOIN customers c ON c.id = o.customer_id;

SELECT o.*, c.name
FROM customers c
INNER JOIN orders o ON o.customer_id = c.id;
```

#### LEFT JOIN

Preserves all rows from the left table. The planner cannot reorder LEFT JOINs as freely.

```sql
-- Find customers with no orders:
SELECT c.id, c.name
FROM customers c
LEFT JOIN orders o ON o.customer_id = c.id
WHERE o.id IS NULL;
```

#### LATERAL JOIN

A correlated subquery expressed as a join. Each row from the left side feeds into the right side.

```sql
-- Top 3 most recent orders per customer:
SELECT c.name, recent.*
FROM customers c
CROSS JOIN LATERAL (
    SELECT o.id, o.total_amount, o.created_at
    FROM orders o
    WHERE o.customer_id = c.id
    ORDER BY o.created_at DESC
    LIMIT 3
) recent;
```

LATERAL is powerful for "top-N per group" queries. It uses a nested loop internally but with an index on `orders(customer_id, created_at DESC)`, each iteration is fast.

#### Self-Joins vs. Window Functions

```sql
-- Self-join to compare consecutive rows (slower):
SELECT a.id, a.value, b.value AS prev_value
FROM measurements a
LEFT JOIN measurements b ON b.id = a.id - 1;

-- Window function alternative (faster, cleaner):
SELECT id, value, LAG(value) OVER (ORDER BY id) AS prev_value
FROM measurements;
```

---

### Join Order

The PostgreSQL planner evaluates different join orders and picks the cheapest. For N tables, there are N! possible orderings. The planner uses dynamic programming for small N and genetic algorithm (GEQO) for large N.

```sql
-- join_collapse_limit controls how many tables to evaluate exhaustively:
SHOW join_collapse_limit;  -- default: 8

-- For queries with many joins, you might increase it:
SET join_collapse_limit = 12;  -- more planning time, potentially better plan

-- Or decrease it to force the written order:
SET join_collapse_limit = 1;  -- use exactly the order written in the query
```

**Forcing join order** (when you know better than the planner):

```sql
-- Explicit JOIN syntax with join_collapse_limit = 1
-- forces this exact execution order:
SET LOCAL join_collapse_limit = 1;
SELECT *
FROM small_table s
JOIN medium_table m ON m.id = s.medium_id
JOIN large_table l ON l.id = m.large_id;
```

---

### Anti-Joins

Three ways to express "rows in A that are not in B." Performance differs significantly.

```sql
-- Method 1: NOT EXISTS (usually best)
SELECT c.*
FROM customers c
WHERE NOT EXISTS (
    SELECT 1 FROM orders o WHERE o.customer_id = c.id
);

-- Method 2: LEFT JOIN ... IS NULL (equivalent to NOT EXISTS)
SELECT c.*
FROM customers c
LEFT JOIN orders o ON o.customer_id = c.id
WHERE o.id IS NULL;

-- Method 3: NOT IN (DANGEROUS with NULLs)
SELECT c.*
FROM customers c
WHERE c.id NOT IN (SELECT customer_id FROM orders);
-- If ANY customer_id in orders is NULL, this returns NO ROWS.
-- NOT IN with a subquery containing NULLs evaluates to UNKNOWN for every row.
```

**Recommendation**: Always use `NOT EXISTS` or `LEFT JOIN ... IS NULL`. Never use `NOT IN` with a subquery unless you are certain the subquery column has no NULLs.

---

## Window Functions

### Ranking Functions

#### ROW_NUMBER()

Assigns a unique sequential number to each row within a partition.

```sql
-- Assign row numbers within each department by salary:
SELECT
    name,
    department,
    salary,
    ROW_NUMBER() OVER (PARTITION BY department ORDER BY salary DESC) AS rank
FROM employees;
```

```
| name    | department | salary  | rank |
|---------|-----------|---------|------|
| Alice   | Eng       | 150000  | 1    |
| Bob     | Eng       | 140000  | 2    |
| Carol   | Eng       | 130000  | 3    |
| Diana   | Sales     | 120000  | 1    |
| Eve     | Sales     | 110000  | 2    |
```

**Classic use case -- Top-N per group**:

```sql
-- Get the most recent order per customer:
WITH ranked AS (
    SELECT *,
        ROW_NUMBER() OVER (PARTITION BY customer_id ORDER BY created_at DESC) AS rn
    FROM orders
)
SELECT * FROM ranked WHERE rn = 1;
```

#### RANK() and DENSE_RANK()

```sql
SELECT
    name,
    score,
    RANK() OVER (ORDER BY score DESC) AS rank,
    DENSE_RANK() OVER (ORDER BY score DESC) AS dense_rank
FROM leaderboard;
```

```
| name    | score | rank | dense_rank |
|---------|-------|------|------------|
| Alice   | 100   | 1    | 1          |
| Bob     | 100   | 1    | 1          |
| Carol   | 95    | 3    | 2          |  -- RANK skips 2, DENSE_RANK doesn't
| Diana   | 90    | 4    | 3          |
```

#### NTILE()

Divides rows into N approximately equal buckets.

```sql
-- Divide customers into quartiles by lifetime spend:
SELECT
    customer_id,
    total_spent,
    NTILE(4) OVER (ORDER BY total_spent DESC) AS quartile
FROM customer_summary;
-- quartile 1 = top 25% spenders, quartile 4 = bottom 25%
```

---

### Aggregate Window Functions

#### Running Totals

```sql
SELECT
    date,
    revenue,
    SUM(revenue) OVER (ORDER BY date) AS cumulative_revenue,
    SUM(revenue) OVER (
        ORDER BY date
        ROWS BETWEEN 6 PRECEDING AND CURRENT ROW
    ) AS rolling_7day_revenue
FROM daily_revenue;
```

#### Moving Averages

```sql
SELECT
    date,
    temperature,
    AVG(temperature) OVER (
        ORDER BY date
        ROWS BETWEEN 2 PRECEDING AND 2 FOLLOWING
    ) AS smoothed_temp_5day
FROM weather_data;
```

#### LAG and LEAD

```sql
SELECT
    date,
    revenue,
    LAG(revenue, 1) OVER (ORDER BY date) AS prev_day,
    LEAD(revenue, 1) OVER (ORDER BY date) AS next_day,
    revenue - LAG(revenue, 1) OVER (ORDER BY date) AS day_over_day_change,
    ROUND(
        100.0 * (revenue - LAG(revenue, 1) OVER (ORDER BY date))
        / NULLIF(LAG(revenue, 1) OVER (ORDER BY date), 0),
        2
    ) AS pct_change
FROM daily_revenue;
```

#### FIRST_VALUE, LAST_VALUE, NTH_VALUE

```sql
SELECT
    employee_id,
    department,
    salary,
    FIRST_VALUE(salary) OVER w AS highest_in_dept,
    LAST_VALUE(salary) OVER w AS lowest_in_dept,
    salary - FIRST_VALUE(salary) OVER w AS diff_from_top
FROM employees
WINDOW w AS (
    PARTITION BY department
    ORDER BY salary DESC
    ROWS BETWEEN UNBOUNDED PRECEDING AND UNBOUNDED FOLLOWING
);
```

> **Critical**: `LAST_VALUE` requires `ROWS BETWEEN UNBOUNDED PRECEDING AND UNBOUNDED FOLLOWING`. The default frame is `RANGE BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW`, which makes `LAST_VALUE` return the current row's value.

#### Frame Specifications

```sql
-- ROWS: physical row offset
ROWS BETWEEN 3 PRECEDING AND CURRENT ROW          -- 4-row window
ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW  -- running total

-- RANGE: logical value offset (groups ties together)
RANGE BETWEEN INTERVAL '7 days' PRECEDING AND CURRENT ROW  -- 7-day window by value

-- GROUPS (PostgreSQL 11+): counts peer groups, not individual rows
GROUPS BETWEEN 1 PRECEDING AND 1 FOLLOWING  -- current group + 1 group on each side
```

---

### Advanced Window Patterns

#### Gaps and Islands

Find contiguous sequences in data (e.g., consecutive days a server was down).

```sql
-- Identify contiguous date ranges ("islands") from a set of dates:
WITH numbered AS (
    SELECT
        event_date,
        event_date - (ROW_NUMBER() OVER (ORDER BY event_date))::int AS grp
    FROM (SELECT DISTINCT event_date FROM server_events WHERE status = 'down') t
)
SELECT
    MIN(event_date) AS island_start,
    MAX(event_date) AS island_end,
    MAX(event_date) - MIN(event_date) + 1 AS duration_days
FROM numbered
GROUP BY grp
ORDER BY island_start;
```

```
| island_start | island_end  | duration_days |
|-------------|------------|---------------|
| 2025-03-01  | 2025-03-04 | 4             |
| 2025-03-10  | 2025-03-10 | 1             |
| 2025-03-15  | 2025-03-20 | 6             |
```

#### Session Identification

Group events into sessions based on inactivity gaps.

```sql
WITH events_with_gap AS (
    SELECT
        user_id,
        event_time,
        event_type,
        CASE
            WHEN event_time - LAG(event_time) OVER (
                PARTITION BY user_id ORDER BY event_time
            ) > INTERVAL '30 minutes'
            THEN 1
            ELSE 0
        END AS new_session
    FROM user_events
),
sessions AS (
    SELECT
        *,
        SUM(new_session) OVER (
            PARTITION BY user_id ORDER BY event_time
        ) AS session_id
    FROM events_with_gap
)
SELECT
    user_id,
    session_id,
    MIN(event_time) AS session_start,
    MAX(event_time) AS session_end,
    COUNT(*) AS event_count,
    MAX(event_time) - MIN(event_time) AS session_duration
FROM sessions
GROUP BY user_id, session_id
ORDER BY user_id, session_start;
```

#### Funnel Analysis

Track user progression through a multi-step conversion funnel.

```sql
WITH funnel_steps AS (
    SELECT
        user_id,
        MIN(CASE WHEN event_type = 'page_view' THEN event_time END) AS step1_time,
        MIN(CASE WHEN event_type = 'add_to_cart' THEN event_time END) AS step2_time,
        MIN(CASE WHEN event_type = 'checkout_start' THEN event_time END) AS step3_time,
        MIN(CASE WHEN event_type = 'purchase' THEN event_time END) AS step4_time
    FROM user_events
    WHERE event_time >= '2025-01-01'
      AND event_time < '2025-02-01'
    GROUP BY user_id
),
funnel AS (
    SELECT
        COUNT(*) AS total_users,
        COUNT(step1_time) AS viewed_page,
        COUNT(CASE WHEN step2_time > step1_time THEN 1 END) AS added_to_cart,
        COUNT(CASE WHEN step3_time > step2_time THEN 1 END) AS started_checkout,
        COUNT(CASE WHEN step4_time > step3_time THEN 1 END) AS purchased
    FROM funnel_steps
)
SELECT
    'Page View' AS step, viewed_page AS users,
    ROUND(100.0 * viewed_page / NULLIF(viewed_page, 0), 1) AS pct
FROM funnel
UNION ALL
SELECT
    'Add to Cart', added_to_cart,
    ROUND(100.0 * added_to_cart / NULLIF(viewed_page, 0), 1)
FROM funnel
UNION ALL
SELECT
    'Checkout', started_checkout,
    ROUND(100.0 * started_checkout / NULLIF(viewed_page, 0), 1)
FROM funnel
UNION ALL
SELECT
    'Purchase', purchased,
    ROUND(100.0 * purchased / NULLIF(viewed_page, 0), 1)
FROM funnel;
```

#### Year-over-Year Comparison

```sql
SELECT
    date_trunc('month', created_at) AS month,
    SUM(total_amount) AS revenue,
    LAG(SUM(total_amount), 12) OVER (ORDER BY date_trunc('month', created_at)) AS revenue_prev_year,
    ROUND(
        100.0 * (
            SUM(total_amount) -
            LAG(SUM(total_amount), 12) OVER (ORDER BY date_trunc('month', created_at))
        ) / NULLIF(
            LAG(SUM(total_amount), 12) OVER (ORDER BY date_trunc('month', created_at)),
            0
        ),
        1
    ) AS yoy_growth_pct
FROM orders
GROUP BY date_trunc('month', created_at)
ORDER BY month;
```

---

## Common Table Expressions (CTEs)

### Basic CTEs

CTEs improve readability and allow a subquery result to be referenced multiple times.

```sql
WITH monthly_revenue AS (
    SELECT
        date_trunc('month', created_at) AS month,
        SUM(total_amount) AS revenue
    FROM orders
    WHERE created_at >= '2024-01-01'
    GROUP BY date_trunc('month', created_at)
),
avg_revenue AS (
    SELECT AVG(revenue) AS avg_monthly FROM monthly_revenue
)
SELECT
    m.month,
    m.revenue,
    a.avg_monthly,
    m.revenue - a.avg_monthly AS diff_from_avg,
    CASE
        WHEN m.revenue > a.avg_monthly * 1.1 THEN 'above_average'
        WHEN m.revenue < a.avg_monthly * 0.9 THEN 'below_average'
        ELSE 'average'
    END AS performance
FROM monthly_revenue m
CROSS JOIN avg_revenue a
ORDER BY m.month;
```

### Recursive CTEs

#### Tree/Hierarchy Traversal

```sql
-- Organizational hierarchy: find all reports under a manager
WITH RECURSIVE org_tree AS (
    -- Base case: the root manager
    SELECT id, name, manager_id, 1 AS depth, ARRAY[name] AS path
    FROM employees
    WHERE id = 1  -- CEO

    UNION ALL

    -- Recursive case: direct reports
    SELECT e.id, e.name, e.manager_id, t.depth + 1, t.path || e.name
    FROM employees e
    INNER JOIN org_tree t ON t.id = e.manager_id
    WHERE t.depth < 10  -- safety limit to prevent infinite recursion
)
SELECT
    depth,
    REPEAT('  ', depth - 1) || name AS indented_name,
    array_to_string(path, ' > ') AS full_path
FROM org_tree
ORDER BY path;
```

```
| depth | indented_name     | full_path                    |
|-------|-------------------|------------------------------|
| 1     | Alice (CEO)       | Alice (CEO)                  |
| 2     |   Bob (VP Eng)    | Alice (CEO) > Bob (VP Eng)   |
| 3     |     Carol (Lead)  | Alice (CEO) > Bob > Carol    |
| 3     |     Dave (Lead)   | Alice (CEO) > Bob > Dave     |
| 2     |   Eve (VP Sales)  | Alice (CEO) > Eve (VP Sales) |
```

#### Date Series Generation

```sql
-- Generate a complete date series (no gaps):
WITH RECURSIVE dates AS (
    SELECT DATE '2025-01-01' AS d
    UNION ALL
    SELECT d + INTERVAL '1 day'
    FROM dates
    WHERE d < DATE '2025-12-31'
)
SELECT
    d.d AS date,
    COALESCE(SUM(o.total_amount), 0) AS revenue
FROM dates d
LEFT JOIN orders o ON DATE(o.created_at) = d.d
GROUP BY d.d
ORDER BY d.d;

-- Alternative using generate_series (non-recursive, preferred):
SELECT
    d::date AS date,
    COALESCE(SUM(o.total_amount), 0) AS revenue
FROM generate_series('2025-01-01'::date, '2025-12-31'::date, '1 day') d
LEFT JOIN orders o ON DATE(o.created_at) = d::date
GROUP BY d::date
ORDER BY d::date;
```

#### Graph Traversal

```sql
-- Find all reachable nodes from node 'A' in a directed graph:
WITH RECURSIVE reachable AS (
    SELECT target AS node, 1 AS hops, ARRAY[source, target] AS path
    FROM edges
    WHERE source = 'A'

    UNION

    SELECT e.target, r.hops + 1, r.path || e.target
    FROM edges e
    INNER JOIN reachable r ON r.node = e.source
    WHERE e.target <> ALL(r.path)  -- cycle detection
      AND r.hops < 20             -- safety limit
)
SELECT DISTINCT node, MIN(hops) AS min_hops
FROM reachable
GROUP BY node
ORDER BY min_hops;
```

### CTE Materialization (PostgreSQL 12+)

Before PostgreSQL 12, CTEs were always materialized (computed once, stored in a temporary buffer). Since v12, the planner can "inline" a CTE when it is referenced only once. You can override this behavior:

```sql
-- Force materialization (computed once, useful if referenced multiple times):
WITH customer_stats AS MATERIALIZED (
    SELECT customer_id, COUNT(*) AS order_count, SUM(total_amount) AS total
    FROM orders
    GROUP BY customer_id
)
SELECT * FROM customer_stats WHERE order_count > 10
UNION ALL
SELECT * FROM customer_stats WHERE total > 10000;

-- Force inlining (pushes filters into the CTE, useful for single reference):
WITH recent_orders AS NOT MATERIALIZED (
    SELECT * FROM orders WHERE created_at > '2025-01-01'
)
SELECT * FROM recent_orders WHERE status = 'pending';
-- Without NOT MATERIALIZED, the planner might compute ALL recent orders
-- then filter. With NOT MATERIALIZED, it pushes the status filter down.
```

**When materialization hurts**: A CTE that selects broadly but is filtered narrowly in the outer query. Materialization prevents the filter from being pushed into the CTE.

**When materialization helps**: A CTE referenced 3+ times. Without materialization, the computation runs 3+ times.

---

## Partitioning

### Range Partitioning

The most common strategy: split a table by value ranges, typically timestamps.

```sql
-- Create partitioned table:
CREATE TABLE events (
    id          BIGSERIAL,
    event_type  TEXT NOT NULL,
    data        JSONB,
    created_at  TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (id, created_at)  -- partition key must be in PK
) PARTITION BY RANGE (created_at);

-- Create monthly partitions:
CREATE TABLE events_2025_01 PARTITION OF events
    FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');
CREATE TABLE events_2025_02 PARTITION OF events
    FOR VALUES FROM ('2025-02-01') TO ('2025-03-01');
CREATE TABLE events_2025_03 PARTITION OF events
    FOR VALUES FROM ('2025-03-01') TO ('2025-04-01');
-- ... continue for each month

-- Create indexes on partitions (inherited by future partitions):
CREATE INDEX ON events(event_type, created_at);
CREATE INDEX ON events(created_at);

-- Default partition catches anything that doesn't match:
CREATE TABLE events_default PARTITION OF events DEFAULT;
```

**Partition pruning** eliminates partitions at plan time:

```sql
EXPLAIN SELECT * FROM events WHERE created_at >= '2025-03-01' AND created_at < '2025-04-01';
```

```
Append  (cost=0.00..456.78 rows=5000 width=128)
  ->  Seq Scan on events_2025_03 events_1  (cost=0.00..456.78 rows=5000 width=128)
        Filter: ((created_at >= '2025-03-01') AND (created_at < '2025-04-01'))
```

Only `events_2025_03` is scanned. All other partitions are pruned.

**Dropping old data** is instant:

```sql
-- Instead of DELETE (which is slow and generates WAL):
DROP TABLE events_2024_01;
-- or detach first if you want to keep the data elsewhere:
ALTER TABLE events DETACH PARTITION events_2024_01;
```

---

### List Partitioning

Split by discrete values.

```sql
CREATE TABLE orders (
    id          BIGSERIAL,
    region      TEXT NOT NULL,
    total       NUMERIC,
    created_at  TIMESTAMPTZ,
    PRIMARY KEY (id, region)
) PARTITION BY LIST (region);

CREATE TABLE orders_us PARTITION OF orders FOR VALUES IN ('US');
CREATE TABLE orders_eu PARTITION OF orders FOR VALUES IN ('EU', 'UK');
CREATE TABLE orders_apac PARTITION OF orders FOR VALUES IN ('JP', 'AU', 'SG', 'IN');
CREATE TABLE orders_other PARTITION OF orders DEFAULT;
```

Useful when queries almost always filter by the partition key and data naturally groups into discrete categories.

---

### Hash Partitioning

Distributes rows evenly across partitions by hash of the partition key.

```sql
CREATE TABLE sessions (
    id          UUID PRIMARY KEY,
    user_id     BIGINT NOT NULL,
    data        JSONB,
    created_at  TIMESTAMPTZ
) PARTITION BY HASH (id);

-- Create 8 partitions (should be a power of 2):
CREATE TABLE sessions_0 PARTITION OF sessions FOR VALUES WITH (MODULUS 8, REMAINDER 0);
CREATE TABLE sessions_1 PARTITION OF sessions FOR VALUES WITH (MODULUS 8, REMAINDER 1);
CREATE TABLE sessions_2 PARTITION OF sessions FOR VALUES WITH (MODULUS 8, REMAINDER 2);
CREATE TABLE sessions_3 PARTITION OF sessions FOR VALUES WITH (MODULUS 8, REMAINDER 3);
CREATE TABLE sessions_4 PARTITION OF sessions FOR VALUES WITH (MODULUS 8, REMAINDER 4);
CREATE TABLE sessions_5 PARTITION OF sessions FOR VALUES WITH (MODULUS 8, REMAINDER 5);
CREATE TABLE sessions_6 PARTITION OF sessions FOR VALUES WITH (MODULUS 8, REMAINDER 6);
CREATE TABLE sessions_7 PARTITION OF sessions FOR VALUES WITH (MODULUS 8, REMAINDER 7);
```

**When to use**: When there is no natural range or list partitioning key but you want to reduce index size and parallelize vacuum. Primarily useful for very large tables (billions of rows). Hash partitioning does not support partition pruning for range queries.

---

### Partition Maintenance

#### Automatic Partition Creation

PostgreSQL does not natively auto-create partitions. Use `pg_partman` or a cron job:

```sql
-- Using pg_partman:
CREATE EXTENSION pg_partman;

SELECT partman.create_parent(
    p_parent_table := 'public.events',
    p_control := 'created_at',
    p_type := 'native',
    p_interval := 'monthly',
    p_premake := 3  -- create 3 future partitions in advance
);

-- Maintenance function (run via pg_cron or crontab):
SELECT partman.run_maintenance();
```

#### Detach and Attach

```sql
-- Detach a partition (non-blocking in PostgreSQL 14+):
ALTER TABLE events DETACH PARTITION events_2024_01 CONCURRENTLY;

-- Move to archive tablespace:
ALTER TABLE events_2024_01 SET TABLESPACE archive_storage;

-- Attach a pre-existing table as a partition:
ALTER TABLE events ATTACH PARTITION events_2025_07
    FOR VALUES FROM ('2025-07-01') TO ('2025-08-01');
-- NOTE: ATTACH validates all existing data matches the constraint.
-- For large tables, add the constraint first, then attach:
ALTER TABLE events_2025_07 ADD CONSTRAINT check_dates
    CHECK (created_at >= '2025-07-01' AND created_at < '2025-08-01');
ALTER TABLE events ATTACH PARTITION events_2025_07
    FOR VALUES FROM ('2025-07-01') TO ('2025-08-01');
-- The planner skips validation if a matching CHECK constraint exists.
```

---

## Query Patterns and Anti-patterns

### Anti-patterns with Fixes

#### 1. SELECT * vs. SELECT Specific Columns

**BAD**:
```sql
SELECT * FROM orders WHERE customer_id = 42;
```

**WHY**: Fetches all columns including large text/JSONB fields you don't need. Prevents Index Only Scans. Breaks when columns are added or removed.

**GOOD**:
```sql
SELECT id, status, total_amount, created_at
FROM orders WHERE customer_id = 42;
```

With a covering index `(customer_id) INCLUDE (id, status, total_amount, created_at)`, this becomes an Index Only Scan.

---

#### 2. Function on Indexed Column

**BAD**:
```sql
SELECT * FROM users WHERE YEAR(created_at) = 2025;         -- MySQL
SELECT * FROM users WHERE EXTRACT(YEAR FROM created_at) = 2025; -- PostgreSQL
```

**WHY**: Applying a function to the column prevents index usage. The database must evaluate the function for every row (Seq Scan).

**GOOD**:
```sql
SELECT * FROM users
WHERE created_at >= '2025-01-01' AND created_at < '2026-01-01';
```

This uses a standard range scan on an index on `created_at`.

---

#### 3. NOT IN with NULLs

**BAD**:
```sql
SELECT * FROM customers
WHERE id NOT IN (SELECT customer_id FROM orders);
-- If ANY order has customer_id = NULL, this returns ZERO rows.
```

**GOOD**:
```sql
SELECT * FROM customers c
WHERE NOT EXISTS (SELECT 1 FROM orders o WHERE o.customer_id = c.id);
```

---

#### 4. DISTINCT to Fix Bad Joins

**BAD**:
```sql
SELECT DISTINCT c.id, c.name
FROM customers c
JOIN orders o ON o.customer_id = c.id
JOIN order_items oi ON oi.order_id = o.id
WHERE oi.product_id = 42;
-- Joins produce duplicates, then DISTINCT removes them (expensive sort/hash).
```

**GOOD**:
```sql
SELECT c.id, c.name
FROM customers c
WHERE EXISTS (
    SELECT 1 FROM orders o
    JOIN order_items oi ON oi.order_id = o.id
    WHERE o.customer_id = c.id AND oi.product_id = 42
);
-- EXISTS stops at the first match. No duplicates to remove.
```

---

#### 5. ORDER BY RAND() / RANDOM()

**BAD**:
```sql
SELECT * FROM products ORDER BY RANDOM() LIMIT 5;
-- Assigns a random value to every row, then sorts. O(N log N).
```

**GOOD**:
```sql
-- Method 1: TABLESAMPLE (PostgreSQL, approximate)
SELECT * FROM products TABLESAMPLE BERNOULLI(0.1) LIMIT 5;

-- Method 2: Offset-based (exact, if you know the count)
SELECT * FROM products
OFFSET floor(random() * (SELECT COUNT(*) FROM products))
LIMIT 5;

-- Method 3: ID-range sampling (fast for large tables with sequential IDs)
WITH bounds AS (
    SELECT MIN(id) AS min_id, MAX(id) AS max_id FROM products
),
random_ids AS (
    SELECT floor(random() * (max_id - min_id + 1) + min_id)::bigint AS id
    FROM bounds, generate_series(1, 20)  -- oversample to handle gaps
)
SELECT p.*
FROM products p
JOIN random_ids r ON r.id = p.id
LIMIT 5;
```

---

#### 6. OFFSET Pagination vs. Keyset Pagination

**BAD**:
```sql
-- Page 1000 of results:
SELECT * FROM products ORDER BY created_at DESC LIMIT 20 OFFSET 19980;
-- PostgreSQL must scan and discard 19,980 rows. Gets slower with deeper pages.
```

**GOOD**:
```sql
-- Keyset pagination: remember the last value from previous page
SELECT * FROM products
WHERE created_at < '2025-03-15T10:23:45Z'  -- last value from previous page
ORDER BY created_at DESC
LIMIT 20;
-- Always scans at most 20 rows using the index. Constant performance.
```

For columns that are not unique, use a composite cursor:

```sql
-- (created_at, id) ensures uniqueness:
SELECT * FROM products
WHERE (created_at, id) < ('2025-03-15T10:23:45Z', 98765)
ORDER BY created_at DESC, id DESC
LIMIT 20;
```

---

#### 7. OR on Different Columns

**BAD**:
```sql
SELECT * FROM events
WHERE user_id = 42 OR session_id = 'abc-123';
-- Cannot use a single index scan. Often falls back to Seq Scan.
```

**GOOD**:
```sql
SELECT * FROM events WHERE user_id = 42
UNION ALL
SELECT * FROM events WHERE session_id = 'abc-123' AND user_id != 42;
-- Each branch uses its own index. UNION ALL avoids dedup overhead.
-- The second branch excludes user_id = 42 to prevent duplicates.
```

Alternatively, PostgreSQL can use BitmapOr with two indexes, but UNION ALL often plans better.

---

#### 8. Implicit Type Conversion

**BAD**:
```sql
-- Column phone is VARCHAR, but query passes integer:
SELECT * FROM users WHERE phone = 5551234567;
-- Database must cast every varchar phone to integer for comparison.
-- Index on phone is unusable.
```

**GOOD**:
```sql
SELECT * FROM users WHERE phone = '5551234567';
```

---

#### 9. Correlated Subquery vs. JOIN

**BAD**:
```sql
SELECT
    o.id,
    o.total_amount,
    (SELECT name FROM customers c WHERE c.id = o.customer_id) AS customer_name
FROM orders o;
-- Executes the subquery once per order row.
```

**GOOD**:
```sql
SELECT o.id, o.total_amount, c.name AS customer_name
FROM orders o
JOIN customers c ON c.id = o.customer_id;
-- Single join operation. Much more efficient.
```

---

#### 10. COUNT(*) for Existence

**BAD**:
```sql
SELECT CASE WHEN COUNT(*) > 0 THEN true ELSE false END
FROM orders WHERE customer_id = 42 AND status = 'pending';
-- Counts ALL matching rows even though you only need to know if any exist.
```

**GOOD**:
```sql
SELECT EXISTS (
    SELECT 1 FROM orders WHERE customer_id = 42 AND status = 'pending'
);
-- Stops scanning after finding the first match.
```

---

### Performance Patterns

#### Batch Processing

```sql
-- BAD: Delete millions of rows in one transaction
DELETE FROM events WHERE created_at < '2024-01-01';
-- Holds locks for the entire duration, generates massive WAL, may OOM.

-- GOOD: Chunked deletes
DO $$
DECLARE
    rows_deleted INT;
BEGIN
    LOOP
        DELETE FROM events
        WHERE id IN (
            SELECT id FROM events
            WHERE created_at < '2024-01-01'
            LIMIT 10000
            FOR UPDATE SKIP LOCKED
        );
        GET DIAGNOSTICS rows_deleted = ROW_COUNT;
        EXIT WHEN rows_deleted = 0;
        COMMIT;
        PERFORM pg_sleep(0.1);  -- brief pause to let other queries through
    END LOOP;
END $$;
```

#### Bulk Inserts

```sql
-- Method 1: COPY (fastest, ~100K-1M rows/sec)
COPY orders(customer_id, total_amount, status, created_at)
FROM '/tmp/orders.csv' WITH (FORMAT csv, HEADER true);

-- Method 2: Multi-value INSERT (faster than individual inserts)
INSERT INTO orders(customer_id, total_amount, status, created_at)
VALUES
    (1, 99.99, 'pending', NOW()),
    (2, 149.99, 'pending', NOW()),
    (3, 79.99, 'pending', NOW());
-- PostgreSQL handles up to ~1000 values efficiently in a single statement.

-- Method 3: UNNEST (programmatic bulk insert from arrays)
INSERT INTO orders(customer_id, total_amount, status)
SELECT * FROM UNNEST(
    ARRAY[1, 2, 3]::int[],
    ARRAY[99.99, 149.99, 79.99]::numeric[],
    ARRAY['pending', 'pending', 'pending']::text[]
);
```

#### UPSERT

```sql
-- INSERT ... ON CONFLICT (PostgreSQL 9.5+):
INSERT INTO user_preferences(user_id, key, value, updated_at)
VALUES (42, 'theme', 'dark', NOW())
ON CONFLICT (user_id, key) DO UPDATE SET
    value = EXCLUDED.value,
    updated_at = EXCLUDED.updated_at
WHERE user_preferences.value IS DISTINCT FROM EXCLUDED.value;
-- The WHERE clause prevents unnecessary updates (and WAL writes)
-- when the value hasn't actually changed.
```

#### Materialized Views

```sql
-- Create a materialized view for expensive aggregations:
CREATE MATERIALIZED VIEW mv_daily_revenue AS
SELECT
    date_trunc('day', created_at)::date AS day,
    COUNT(*) AS order_count,
    SUM(total_amount) AS revenue,
    AVG(total_amount) AS avg_order_value
FROM orders
GROUP BY date_trunc('day', created_at)::date;

-- Add index for fast lookups:
CREATE UNIQUE INDEX ON mv_daily_revenue(day);

-- Refresh (blocks reads during refresh):
REFRESH MATERIALIZED VIEW mv_daily_revenue;

-- Refresh concurrently (requires a unique index, does not block reads):
REFRESH MATERIALIZED VIEW CONCURRENTLY mv_daily_revenue;

-- Query (instant compared to scanning the orders table):
SELECT * FROM mv_daily_revenue WHERE day >= '2025-01-01' ORDER BY day;
```

#### SKIP LOCKED for Queue Patterns

```sql
-- Worker claims the next available job without blocking other workers:
WITH next_job AS (
    SELECT id
    FROM job_queue
    WHERE status = 'pending'
    ORDER BY priority DESC, created_at ASC
    LIMIT 1
    FOR UPDATE SKIP LOCKED
)
UPDATE job_queue
SET status = 'processing', worker_id = pg_backend_pid(), started_at = NOW()
FROM next_job
WHERE job_queue.id = next_job.id
RETURNING job_queue.*;
```

`SKIP LOCKED` skips rows that are already locked by other transactions, allowing multiple workers to claim different jobs concurrently without blocking.

#### Advisory Locks

```sql
-- Application-level locking without table contention:

-- Try to acquire a lock for processing customer 42's report:
SELECT pg_try_advisory_lock(hashtext('customer_report'), 42);
-- Returns true if acquired, false if another session holds it.

-- Do the work...

-- Release the lock:
SELECT pg_advisory_unlock(hashtext('customer_report'), 42);

-- Transaction-scoped (auto-releases at COMMIT/ROLLBACK):
SELECT pg_advisory_xact_lock(hashtext('singleton_job'), 1);
```

---

## Database Configuration (PostgreSQL)

### Memory Settings

#### shared_buffers

PostgreSQL's main data cache. Recommended: **25% of total RAM**.

```sql
-- Check current setting:
SHOW shared_buffers;

-- Set in postgresql.conf:
-- shared_buffers = '8GB'   -- for a 32GB server
```

Setting this too high (>40% of RAM) can hurt because PostgreSQL also relies on the OS page cache. The two caches together should not exceed total RAM.

#### effective_cache_size

Tells the planner how much total cache (shared_buffers + OS page cache) is available. Does not allocate memory. Recommended: **50-75% of total RAM**.

```sql
-- effective_cache_size = '24GB'  -- for a 32GB server
```

A higher value makes the planner more willing to use index scans (since it believes random I/O will be served from cache).

#### work_mem

Memory per sort/hash operation per query. Applies per-operation, not per-query. A complex query with 5 sorts uses up to 5x `work_mem`.

```sql
-- Default is often 4MB. For analytics workloads, increase:
-- work_mem = '64MB'

-- Be careful: 100 concurrent connections * 5 operations * 64MB = 32GB
-- Set conservatively globally, override per-session for analytics:
SET work_mem = '256MB';  -- for a specific analytical query
```

Symptoms of too-low work_mem: `Sort Method: external merge` or `Hash Batches > 1` in EXPLAIN output.

#### maintenance_work_mem

Memory for maintenance operations: VACUUM, CREATE INDEX, ALTER TABLE ADD FOREIGN KEY.

```sql
-- maintenance_work_mem = '1GB'  -- for a server that can afford it
-- Only one maintenance operation runs at a time per session, so set generously.
```

---

### Planner Settings

#### random_page_cost

Cost of a random page read relative to a sequential page read (which is 1.0).

```sql
-- Default: 4.0 (appropriate for spinning disks)
-- For SSDs: set to 1.1 - 1.5
-- random_page_cost = 1.1

-- This makes the planner much more willing to use index scans,
-- since random I/O on SSDs is nearly as fast as sequential I/O.
```

#### effective_io_concurrency

Number of concurrent disk I/O operations. Affects bitmap heap scans and prefetching.

```sql
-- Default: 1 (for spinning disks)
-- For SSDs: 200
-- For cloud storage (EBS, GCS PD): 200
-- effective_io_concurrency = 200
```

#### default_statistics_target

Controls the sample size for ANALYZE. Higher = more accurate statistics but slower ANALYZE.

```sql
-- Default: 100
-- For columns with skewed distributions, increase per-column:
ALTER TABLE events ALTER COLUMN event_type SET STATISTICS 1000;
ANALYZE events;
```

---

### Connection Management

#### The Connection Problem

Each PostgreSQL connection is a separate OS process consuming ~5-10MB of RSS memory. 500 idle connections = 2.5-5GB wasted.

```sql
-- Check connection usage:
SELECT
    state,
    COUNT(*) AS connections,
    MAX(NOW() - state_change) AS max_idle_time
FROM pg_stat_activity
WHERE backend_type = 'client backend'
GROUP BY state;
```

#### pgbouncer

Use pgbouncer as a connection pooler between your application and PostgreSQL.

| Mode | Behavior | Use Case |
|------|----------|----------|
| `session` | Connection assigned for entire session | Legacy apps with session state |
| `transaction` | Connection assigned for each transaction, returned to pool at COMMIT | Most web applications |
| `statement` | Connection assigned per statement | Simple queries, no multi-statement transactions |

**Transaction pooling** is the most common. With 1000 application connections and a pgbouncer pool of 50, PostgreSQL only sees ~50 connections.

```ini
; pgbouncer.ini
[databases]
myapp = host=localhost port=5432 dbname=myapp

[pgbouncer]
listen_port = 6432
pool_mode = transaction
max_client_conn = 1000
default_pool_size = 50
min_pool_size = 10
reserve_pool_size = 5
```

---

### Monitoring Queries

#### Top Queries by Total Time (pg_stat_statements)

```sql
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

SELECT
    LEFT(query, 80) AS query_preview,
    calls,
    ROUND(total_exec_time::numeric, 2) AS total_ms,
    ROUND(mean_exec_time::numeric, 2) AS avg_ms,
    ROUND((100 * total_exec_time / SUM(total_exec_time) OVER ())::numeric, 2) AS pct_total,
    rows
FROM pg_stat_statements
WHERE userid = (SELECT usesysid FROM pg_user WHERE usename = current_user)
ORDER BY total_exec_time DESC
LIMIT 20;
```

#### Table Access Patterns

```sql
SELECT
    schemaname || '.' || relname AS table,
    seq_scan,
    seq_tup_read,
    idx_scan,
    idx_tup_fetch,
    n_tup_ins AS inserts,
    n_tup_upd AS updates,
    n_tup_del AS deletes,
    n_live_tup AS live_rows,
    n_dead_tup AS dead_rows,
    ROUND(100.0 * n_dead_tup / NULLIF(n_live_tup + n_dead_tup, 0), 1) AS dead_pct,
    last_vacuum,
    last_autovacuum,
    last_analyze,
    last_autoanalyze
FROM pg_stat_user_tables
ORDER BY seq_tup_read DESC;
```

#### Lock Monitoring

```sql
-- Find blocked queries and what is blocking them:
SELECT
    blocked.pid AS blocked_pid,
    blocked.query AS blocked_query,
    blocked.wait_event_type || ':' || blocked.wait_event AS blocked_on,
    blocking.pid AS blocking_pid,
    blocking.query AS blocking_query,
    NOW() - blocked.query_start AS blocked_duration
FROM pg_stat_activity blocked
JOIN pg_locks bl ON bl.pid = blocked.pid AND NOT bl.granted
JOIN pg_locks gl ON gl.locktype = bl.locktype
    AND gl.database IS NOT DISTINCT FROM bl.database
    AND gl.relation IS NOT DISTINCT FROM bl.relation
    AND gl.page IS NOT DISTINCT FROM bl.page
    AND gl.tuple IS NOT DISTINCT FROM bl.tuple
    AND gl.virtualxid IS NOT DISTINCT FROM bl.virtualxid
    AND gl.transactionid IS NOT DISTINCT FROM bl.transactionid
    AND gl.classid IS NOT DISTINCT FROM bl.classid
    AND gl.objid IS NOT DISTINCT FROM bl.objid
    AND gl.objsubid IS NOT DISTINCT FROM bl.objsubid
    AND gl.pid != bl.pid
    AND gl.granted
JOIN pg_stat_activity blocking ON blocking.pid = gl.pid
ORDER BY blocked_duration DESC;
```

#### Active Long-Running Queries

```sql
SELECT
    pid,
    NOW() - query_start AS duration,
    state,
    LEFT(query, 100) AS query_preview,
    wait_event_type,
    wait_event
FROM pg_stat_activity
WHERE state != 'idle'
  AND query NOT LIKE '%pg_stat_activity%'
  AND NOW() - query_start > INTERVAL '30 seconds'
ORDER BY duration DESC;
```

#### Cache Hit Ratio

```sql
-- Table cache hit ratio (should be >99% for OLTP):
SELECT
    SUM(heap_blks_hit) AS cache_hits,
    SUM(heap_blks_read) AS disk_reads,
    ROUND(
        100.0 * SUM(heap_blks_hit) /
        NULLIF(SUM(heap_blks_hit) + SUM(heap_blks_read), 0),
        2
    ) AS cache_hit_ratio
FROM pg_statio_user_tables;

-- Index cache hit ratio:
SELECT
    SUM(idx_blks_hit) AS cache_hits,
    SUM(idx_blks_read) AS disk_reads,
    ROUND(
        100.0 * SUM(idx_blks_hit) /
        NULLIF(SUM(idx_blks_hit) + SUM(idx_blks_read), 0),
        2
    ) AS cache_hit_ratio
FROM pg_statio_user_indexes;
```

---

## Migration Safety

### Safe Operations

These operations take minimal locks and complete quickly on any table size.

#### Adding a Nullable Column

```sql
-- Safe: only a catalog update, no table rewrite.
ALTER TABLE orders ADD COLUMN notes TEXT;
-- Takes an ACCESS EXCLUSIVE lock, but releases it in milliseconds
-- because no data is written to existing rows.
```

#### Adding an Index Concurrently

```sql
-- Safe: does not block reads or writes.
CREATE INDEX CONCURRENTLY idx_orders_email ON orders(email);
-- Takes longer (scans table twice) but does not lock.
```

#### Adding a CHECK Constraint as NOT VALID

```sql
-- Step 1: Add constraint without validating existing rows (instant):
ALTER TABLE orders ADD CONSTRAINT chk_amount_positive
    CHECK (total_amount >= 0) NOT VALID;
-- New inserts/updates are validated immediately.

-- Step 2: Validate existing rows in the background (takes ShareUpdateExclusiveLock, no blocking):
ALTER TABLE orders VALIDATE CONSTRAINT chk_amount_positive;
```

---

### Dangerous Operations

These operations can lock the table for extended periods or rewrite it entirely.

#### Adding a NOT NULL Column Without Default

```sql
-- DANGEROUS (PostgreSQL < 11): Rewrites entire table.
ALTER TABLE orders ADD COLUMN priority INT NOT NULL DEFAULT 0;

-- SAFE (PostgreSQL 11+): Default is stored in catalog, no rewrite.
ALTER TABLE orders ADD COLUMN priority INT NOT NULL DEFAULT 0;
-- PostgreSQL 11+ handles this in milliseconds.
```

#### Changing Column Type

```sql
-- DANGEROUS: Rewrites entire table, holds ACCESS EXCLUSIVE lock.
ALTER TABLE orders ALTER COLUMN total_amount TYPE DECIMAL(12,2);
-- On a 100-million-row table, this could lock for hours.

-- SAFE alternative for compatible types:
-- Some type changes don't require a rewrite:
ALTER TABLE orders ALTER COLUMN total_amount TYPE DECIMAL(14,2);
-- Increasing precision is safe (no rewrite needed).
-- But changing varchar(50) to int requires rewriting every row.
```

#### Adding a Unique Constraint

```sql
-- DANGEROUS: Scans entire table and builds index while holding lock.
ALTER TABLE users ADD CONSTRAINT uniq_email UNIQUE (email);

-- SAFE alternative:
CREATE UNIQUE INDEX CONCURRENTLY idx_users_email_uniq ON users(email);
ALTER TABLE users ADD CONSTRAINT uniq_email UNIQUE USING INDEX idx_users_email_uniq;
-- The index creation is non-blocking. The constraint attachment is instant.
```

---

### Zero-downtime Migration Patterns

#### Add Column with Backfill

```sql
-- Step 1: Add nullable column (instant):
ALTER TABLE orders ADD COLUMN total_with_tax NUMERIC;

-- Step 2: Backfill in batches:
DO $$
DECLARE
    batch_start BIGINT := 0;
    batch_size BIGINT := 50000;
    max_id BIGINT;
BEGIN
    SELECT MAX(id) INTO max_id FROM orders;
    WHILE batch_start <= max_id LOOP
        UPDATE orders
        SET total_with_tax = total_amount * 1.08
        WHERE id > batch_start AND id <= batch_start + batch_size
          AND total_with_tax IS NULL;
        batch_start := batch_start + batch_size;
        COMMIT;
        PERFORM pg_sleep(0.05);  -- throttle to reduce replication lag
    END LOOP;
END $$;

-- Step 3: Add NOT NULL constraint safely:
ALTER TABLE orders ADD CONSTRAINT chk_tax_not_null
    CHECK (total_with_tax IS NOT NULL) NOT VALID;
ALTER TABLE orders VALIDATE CONSTRAINT chk_tax_not_null;
-- Alternatively, set NOT NULL with the validated constraint backing it:
ALTER TABLE orders ALTER COLUMN total_with_tax SET NOT NULL;
```

#### Dual-Write Pattern for Column Renames

Renaming a column directly with `ALTER TABLE ... RENAME COLUMN` is fast, but application code must be updated simultaneously. For zero-downtime:

```
Phase 1: Add new column, dual-write to both old and new.
Phase 2: Backfill new column from old column.
Phase 3: Switch application reads to new column.
Phase 4: Stop writing to old column.
Phase 5: Drop old column.
```

```sql
-- Phase 1:
ALTER TABLE users ADD COLUMN full_name TEXT;

-- Application writes to both:
-- INSERT INTO users(name, full_name, ...) VALUES('Alice', 'Alice', ...);

-- Phase 2:
UPDATE users SET full_name = name WHERE full_name IS NULL;

-- Phase 3: Switch reads to full_name
-- Phase 4: Stop writes to name
-- Phase 5:
ALTER TABLE users DROP COLUMN name;
```

#### Ghost Table Migration (for type changes)

```sql
-- Step 1: Create new table with desired schema:
CREATE TABLE orders_new (
    id BIGSERIAL PRIMARY KEY,
    customer_id INT NOT NULL,
    total_amount DECIMAL(12,2) NOT NULL,  -- changed from NUMERIC
    status TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Step 2: Copy existing data:
INSERT INTO orders_new(id, customer_id, total_amount, status, created_at)
SELECT id, customer_id, total_amount::decimal(12,2), status, created_at
FROM orders;

-- Step 3: Set up trigger on old table to capture changes during migration:
CREATE OR REPLACE FUNCTION sync_orders() RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        INSERT INTO orders_new VALUES (NEW.*) ON CONFLICT (id) DO UPDATE SET
            customer_id = EXCLUDED.customer_id,
            total_amount = EXCLUDED.total_amount,
            status = EXCLUDED.status;
    ELSIF TG_OP = 'UPDATE' THEN
        UPDATE orders_new SET
            customer_id = NEW.customer_id,
            total_amount = NEW.total_amount,
            status = NEW.status
        WHERE id = NEW.id;
    ELSIF TG_OP = 'DELETE' THEN
        DELETE FROM orders_new WHERE id = OLD.id;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_sync_orders
AFTER INSERT OR UPDATE OR DELETE ON orders
FOR EACH ROW EXECUTE FUNCTION sync_orders();

-- Step 4: Swap tables (brief lock):
BEGIN;
ALTER TABLE orders RENAME TO orders_old;
ALTER TABLE orders_new RENAME TO orders;
COMMIT;

-- Step 5: Clean up:
DROP TRIGGER trg_sync_orders ON orders_old;
DROP TABLE orders_old;
```

#### Online Schema Change Tools

**pg_repack**: Rebuilds tables and indexes online without exclusive locks.

```bash
# Rebuild a table to reclaim bloat:
pg_repack --table orders --no-superuser-check -d mydb

# Rebuild all indexes on a table:
pg_repack --table orders --only-indexes -d mydb
```

**gh-ost** (MySQL): GitHub's online schema change tool.

```bash
# Change column type without blocking:
gh-ost \
  --host=localhost \
  --database=mydb \
  --table=orders \
  --alter="MODIFY total_amount DECIMAL(12,2) NOT NULL" \
  --execute
```

gh-ost uses binary log streaming instead of triggers, making it safer for high-traffic MySQL databases. It creates a ghost table, streams changes via binlog, and performs an atomic table swap.

---

## Quick Reference: Index Selection Cheat Sheet

| Query Pattern | Recommended Index Type |
|--------------|----------------------|
| `WHERE col = value` | B-tree |
| `WHERE col BETWEEN a AND b` | B-tree |
| `WHERE col IN (a, b, c)` | B-tree |
| `WHERE col LIKE 'prefix%'` | B-tree |
| `WHERE col LIKE '%substring%'` | GIN with pg_trgm |
| `WHERE jsonb_col @> '{"key": "val"}'` | GIN (jsonb_path_ops) |
| `WHERE jsonb_col ? 'key'` | GIN |
| `WHERE array_col @> ARRAY[1,2]` | GIN |
| `WHERE tsvector @@ tsquery` | GIN (reads) or GiST (writes) |
| `WHERE ST_DWithin(geom, point, radius)` | GiST |
| `WHERE range_col && range_val` | GiST |
| `WHERE timestamp_col > value` (naturally ordered) | BRIN |
| `WHERE col = value` (small subset of rows) | B-tree partial index |
| `WHERE LOWER(col) = value` | B-tree expression index |
| Covering query (all cols in SELECT + WHERE) | B-tree with INCLUDE |

---

## Quick Reference: EXPLAIN Red Flags

| Symptom | Likely Cause | Fix |
|---------|-------------|-----|
| Seq Scan on large table with low row count | Missing index | Add appropriate index |
| `Rows Removed by Filter: 999000` (high) | Bad index or missing index | Add more selective index |
| `actual rows=50000` vs `rows=100` (estimate way off) | Stale statistics | `ANALYZE table_name` |
| `Sort Method: external merge Disk` | `work_mem` too low | Increase `work_mem` or add index for ORDER BY |
| `Hash Batches: 16` (high) | `work_mem` too low for hash join | Increase `work_mem` |
| `Heap Fetches: 50000` on Index Only Scan | Table not vacuumed recently | `VACUUM table_name` |
| Nested Loop with Seq Scan inner | Missing index on join column | Add index on inner table's join column |
| `Buffers: shared read=100000` (all reads, no hits) | Cold cache or `shared_buffers` too small | Increase `shared_buffers`, re-run query |
| `temp read=5000 written=5000` | Spilling to disk | Increase `work_mem` |

---

*This reference covers PostgreSQL 14+ with notes for MySQL where applicable. All SQL examples are tested patterns from production environments. For version-specific behavior, consult the official PostgreSQL documentation.*
