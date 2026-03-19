# PostgreSQL Query Optimization Reference

Comprehensive reference for understanding, diagnosing, and fixing query performance
issues in PostgreSQL.

---

## 1. Reading EXPLAIN ANALYZE Output

### Basic Usage

```sql
EXPLAIN (ANALYZE, BUFFERS, TIMING, FORMAT TEXT) SELECT ...;
```

Always use `ANALYZE` to get actual execution data. Without it, you only see estimates.
Add `BUFFERS` for I/O information and `TIMING` for per-node time measurements.

### Output Fields Explained

#### Cost Fields

```
Seq Scan on users  (cost=0.00..1523.00 rows=50000 width=244)
                    ^^^^^  ^^^^^^^     ^^^^^       ^^^^^^^^^
                    |      |           |           |
                    |      |           |           Average row width in bytes
                    |      |           Estimated number of rows returned
                    |      Total cost to retrieve all rows
                    Startup cost (before first row can be returned)
```

- **cost (startup..total)**: Measured in arbitrary "cost units" (sequential page reads
  by default). Startup cost is the work done before the first row is produced (e.g.,
  sorting). Total cost is for retrieving all rows.
- **rows**: Planner's estimate of rows returned by this node. Compare with `actual rows`
  to detect estimation errors.
- **width**: Estimated average size (bytes) of each output row.

#### Actual Execution Fields (with ANALYZE)

```
Seq Scan on users  (cost=0.00..1523.00 rows=50000 width=244)
                   (actual time=0.012..25.314 rows=49873 loops=1)
                    ^^^^^^^^^^^^^^^^^^^       ^^^^^      ^^^^^^^
                    |                         |          |
                    |                         |          Number of times this node executed
                    |                         Actual rows returned per loop
                    Actual time (ms): startup..total per loop
```

- **actual time (startup..total)**: Wall-clock milliseconds. Startup is time to first
  row, total is time for all rows. These are *per loop* values.
- **rows**: Actual rows returned per loop iteration.
- **loops**: How many times this node was executed. For inner sides of nested loops,
  this is often > 1. Multiply `actual time * loops` for the true total time.

#### Buffer Fields (with BUFFERS)

```
Buffers: shared hit=128 read=45 dirtied=3 written=1
         ^^^^^^^^^^^^^^ ^^^^^^^  ^^^^^^^^^  ^^^^^^^^^
         |              |        |          |
         |              |        |          Pages written back to disk
         |              |        Pages modified in shared buffers
         |              Pages read from OS (may still come from OS cache)
         Pages found in PostgreSQL shared buffer cache
```

- **shared hit**: Blocks found in PostgreSQL's shared_buffers (fast).
- **shared read**: Blocks that had to be read from the OS (may hit OS page cache).
- **shared dirtied**: Blocks modified in memory.
- **shared written**: Blocks written to disk during this query.
- **temp read/written**: Temp file I/O (indicates work_mem overflow — bad for performance).

#### I/O Timing (with TIMING and track_io_timing = on)

```
I/O Timings: shared read=12.435 write=0.321
```

Shows actual I/O wait time. If `shared read` count is high but I/O time is low, data
came from the OS page cache. If I/O time is high, you have true disk I/O bottlenecks.

#### Planning vs Execution Time

```
Planning Time: 0.254 ms
Execution Time: 25.891 ms
```

- **Planning Time**: Time to parse SQL, generate plans, and select the best plan.
  If high, you may have too many partitions or very complex queries.
- **Execution Time**: Actual execution. Includes row retrieval, sorting, etc.
  Does NOT include result transfer to the client.

### Format Options

```sql
-- Human-readable (default)
EXPLAIN (ANALYZE, FORMAT TEXT) SELECT ...;

-- JSON (best for programmatic parsing)
EXPLAIN (ANALYZE, FORMAT JSON) SELECT ...;

-- YAML
EXPLAIN (ANALYZE, FORMAT YAML) SELECT ...;

-- XML
EXPLAIN (ANALYZE, FORMAT XML) SELECT ...;
```

---

## 2. Node Types Reference

### Scan Nodes

| Node Type             | Description                                          | When Chosen                                    |
|-----------------------|------------------------------------------------------|------------------------------------------------|
| Seq Scan              | Full sequential table scan                           | No usable index or large fraction of table     |
| Index Scan            | Traverse B-tree, then fetch heap tuple               | Selective predicate on indexed column           |
| Index Only Scan       | Reads only from the index (no heap fetch)            | All needed columns in the index, VM up-to-date |
| Bitmap Index Scan     | Builds a bitmap of matching TIDs                     | Moderate selectivity, or combining indexes     |
| Bitmap Heap Scan      | Fetches heap pages using the bitmap                  | Follows a Bitmap Index Scan                    |
| TID Scan              | Fetches by physical tuple ID                         | WHERE ctid = '(page, offset)'                  |
| TID Range Scan        | Scans a range of physical tuple IDs                  | WHERE ctid >= ... AND ctid < ...               |
| Subquery Scan         | Wraps a subquery result                              | FROM (subquery) patterns                       |
| Function Scan         | Scans rows returned by a set-returning function      | FROM generate_series(), unnest(), etc.         |
| Values Scan           | Scans inline VALUES lists                            | VALUES (...), (...) in FROM clause             |
| CTE Scan              | Scans a materialized CTE                             | WITH cte AS (...) SELECT ... FROM cte          |
| Work Table Scan       | Scans recursive CTE working table                    | WITH RECURSIVE ...                             |
| Custom Scan           | Extension-provided scan (e.g., Citus, TimescaleDB)   | Extension-specific                             |

### Join Nodes

| Node Type        | Description                                          | When Chosen                                     |
|------------------|------------------------------------------------------|-------------------------------------------------|
| Nested Loop      | For each outer row, scan inner side                  | Small outer set, indexed inner lookup            |
| Hash Join        | Build hash table from inner, probe with outer        | Medium-to-large unsorted datasets                |
| Merge Join       | Merge two pre-sorted inputs                          | Both sides pre-sorted or cheaply sortable        |

### Aggregate / Group Nodes

| Node Type        | Description                                          | When Chosen                                     |
|------------------|------------------------------------------------------|-------------------------------------------------|
| Aggregate        | Computes aggregate (SUM, COUNT, etc.) over all rows  | No GROUP BY                                      |
| HashAggregate    | Groups rows using a hash table                       | Moderate number of groups                        |
| GroupAggregate   | Groups pre-sorted rows                               | Input already sorted on group key                |
| Mixed Aggregate  | Combination of hash and group aggregation            | GROUPING SETS with mixed strategies              |

### Sort / Unique / Limit Nodes

| Node Type        | Description                                          | When Chosen                                     |
|------------------|------------------------------------------------------|-------------------------------------------------|
| Sort             | Sorts rows (quicksort or external merge)             | ORDER BY, Merge Join input, GroupAggregate       |
| Incremental Sort | Sorts partially-sorted data (PG 13+)                | Input pre-sorted on leading keys                 |
| Unique           | Removes consecutive duplicates from sorted input     | DISTINCT on sorted data                          |
| Limit            | Returns only the first N rows                        | LIMIT clause                                     |

### Miscellaneous Nodes

| Node Type        | Description                                          | When Chosen                                     |
|------------------|------------------------------------------------------|-------------------------------------------------|
| Materialize      | Caches inner result for re-scanning                  | Nested Loop needs to rescan inner                |
| Memoize          | Hash-caches parameterized inner results (PG 14+)    | Nested Loop with repeated parameter values       |
| Append           | Concatenates results from multiple sub-plans         | UNION ALL, partitioned table scans               |
| Merge Append     | Ordered merge of pre-sorted sub-plans                | ORDER BY on partitioned table                    |
| Gather           | Collects rows from parallel workers                  | Parallel query                                   |
| Gather Merge     | Merge-sorts rows from parallel workers               | Parallel query with ORDER BY                     |
| Result           | Evaluates a constant expression                      | SELECT 1, WHERE false                            |
| ProjectSet       | Evaluates set-returning functions in target list     | SELECT generate_series(1,10)                     |
| LockRows         | Acquires row locks                                   | SELECT ... FOR UPDATE                            |

---

## 3. Join Algorithm Selection

### Nested Loop Join

```
Nested Loop  (cost=0.43..1250.20 rows=100 ...)
  ->  Index Scan on orders  (cost=0.43..850.00 rows=100 ...)
  ->  Index Scan on customers  (cost=0.29..3.99 rows=1 ...)
```

**When chosen**: Small outer set, inner side has a fast lookup (usually an index scan).
Good when outer side returns few rows, and you can look up each inner match cheaply.

**Influence**: Ensure indexes exist on the join column of the inner table. Reduce the
outer set with selective WHERE clauses.

### Hash Join

```
Hash Join  (cost=1250.00..3500.00 rows=50000 ...)
  Hash Cond: (o.customer_id = c.id)
  ->  Seq Scan on orders o  (cost=0.00..1500.00 rows=100000 ...)
  ->  Hash  (cost=1000.00..1000.00 rows=50000 ...)
        ->  Seq Scan on customers c  (...)
```

**When chosen**: Medium-to-large datasets on both sides, no pre-sorted order. The inner
table must fit in work_mem (otherwise it batches to disk).

**Influence**: Increase `work_mem` to fit the hash table. The smaller table should be
the hash (inner) side.

### Merge Join

```
Merge Join  (cost=500.00..2000.00 rows=50000 ...)
  Merge Cond: (o.customer_id = c.id)
  ->  Index Scan on orders_customer_id_idx  (...)
  ->  Index Scan on customers_pkey  (...)
```

**When chosen**: Both sides are already sorted on the join key (e.g., via index scans),
or when the sort cost is acceptable. Excellent for large equi-joins when data is ordered.

**Influence**: Create indexes matching the join key order. Good for one-to-many joins
where both sides are large and sorted.

### Disabling Join Types for Debugging

```sql
-- Temporarily disable to test alternatives (never do this in production permanently)
SET enable_nestloop = off;
SET enable_hashjoin = off;
SET enable_mergejoin = off;
```

---

## 4. Common Query Anti-Patterns

### Anti-Pattern 1: SELECT * Instead of Specific Columns

```sql
-- BAD: fetches all columns, prevents index-only scans
SELECT * FROM orders WHERE customer_id = 42;

-- GOOD: fetch only what you need, enables index-only scan if covered
SELECT id, order_date, total FROM orders WHERE customer_id = 42;
```

**Why it's bad**: Wastes I/O and memory transferring unused columns. Prevents the
planner from using index-only scans. If schema changes, queries may break silently.

### Anti-Pattern 2: OR Conditions Preventing Index Use

```sql
-- BAD: OR across different columns often leads to Seq Scan
SELECT * FROM orders
WHERE customer_id = 42 OR status = 'pending';

-- GOOD: split into UNION ALL for separate index scans
SELECT * FROM orders WHERE customer_id = 42
UNION ALL
SELECT * FROM orders WHERE status = 'pending'
  AND customer_id != 42;  -- avoid duplicates
```

**Why it's bad**: The planner often cannot use separate indexes with OR. It may fall
back to a sequential scan. UNION ALL allows each branch to use its own index.

### Anti-Pattern 3: NOT IN with NULLs

```sql
-- BAD: NOT IN returns no rows if subquery contains any NULL
SELECT * FROM customers
WHERE id NOT IN (SELECT customer_id FROM banned_customers);

-- GOOD: NOT EXISTS handles NULLs correctly
SELECT * FROM customers c
WHERE NOT EXISTS (
  SELECT 1 FROM banned_customers b WHERE b.customer_id = c.id
);
```

**Why it's bad**: If `banned_customers.customer_id` has any NULL value, `NOT IN`
returns zero rows — a subtle and dangerous bug. `NOT EXISTS` handles NULLs correctly
and is often faster (anti-join).

### Anti-Pattern 4: Functions on Indexed Columns

```sql
-- BAD: function on column prevents B-tree index use
SELECT * FROM users WHERE lower(email) = 'user@example.com';

-- GOOD option 1: create an expression index
CREATE INDEX idx_users_email_lower ON users (lower(email));
SELECT * FROM users WHERE lower(email) = 'user@example.com';

-- GOOD option 2: use citext type or collation
ALTER TABLE users ALTER COLUMN email TYPE citext;
SELECT * FROM users WHERE email = 'user@example.com';
```

**Why it's bad**: B-tree indexes store the raw column value. Applying a function to the
column means the index cannot be used. Expression indexes solve this.

### Anti-Pattern 5: Correlated Subqueries

```sql
-- BAD: executes subquery once per outer row (N+1 at the SQL level)
SELECT o.id, o.total,
  (SELECT c.name FROM customers c WHERE c.id = o.customer_id) AS customer_name
FROM orders o;

-- GOOD: use a JOIN
SELECT o.id, o.total, c.name AS customer_name
FROM orders o
JOIN customers c ON c.id = o.customer_id;

-- GOOD: use LATERAL for set-returning subqueries
SELECT o.id, o.total, latest.created_at
FROM orders o
JOIN LATERAL (
  SELECT created_at FROM order_events e
  WHERE e.order_id = o.id
  ORDER BY created_at DESC
  LIMIT 1
) latest ON true;
```

**Why it's bad**: Correlated subqueries execute once per outer row. While PostgreSQL may
optimize some into joins internally, complex ones will not be optimized. Explicit JOINs
or LATERAL give the planner more optimization options.

### Anti-Pattern 6: LIKE '%term%' Without Trigram Index

```sql
-- BAD: leading wildcard forces full sequential scan
SELECT * FROM products WHERE name LIKE '%widget%';

-- GOOD: create a GIN trigram index
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX idx_products_name_trgm ON products USING gin (name gin_trgm_ops);
SELECT * FROM products WHERE name LIKE '%widget%';
-- also supports: ILIKE, ~, ~*
```

**Why it's bad**: B-tree indexes can only handle prefix patterns (`LIKE 'term%'`). For
infix or suffix patterns, you need a trigram (pg_trgm) GIN index which indexes all
3-character subsequences.

### Anti-Pattern 7: OFFSET for Pagination

```sql
-- BAD: OFFSET still scans and discards rows — O(N) for page N
SELECT * FROM orders ORDER BY created_at DESC LIMIT 20 OFFSET 10000;

-- GOOD: keyset (cursor-based) pagination — O(1) for any page
SELECT * FROM orders
WHERE created_at < '2025-06-15T10:30:00Z'  -- last value from previous page
ORDER BY created_at DESC
LIMIT 20;
```

**Why it's bad**: With OFFSET 10000, PostgreSQL must generate 10,020 rows and discard
the first 10,000. Performance degrades linearly with page depth. Keyset pagination
uses an indexed WHERE clause for constant-time access to any "page."

### Anti-Pattern 8: COUNT(*) on Large Tables

```sql
-- BAD: full table scan to count rows
SELECT COUNT(*) FROM events;  -- on a 500M row table, this takes minutes

-- GOOD: use estimated count from pg_class for approximate counts
SELECT reltuples::bigint AS estimated_count
FROM pg_class
WHERE relname = 'events';

-- GOOD: use a count tracking table for exact counts (maintained by triggers)
-- GOOD: use HyperLogLog (hll extension) for approximate distinct counts
```

**Why it's bad**: `COUNT(*)` on a large table requires scanning every row due to MVCC
(each transaction may see different rows). The planner statistics in `pg_class.reltuples`
give a good estimate without a scan, updated by ANALYZE.

### Anti-Pattern 9: Unnecessary DISTINCT

```sql
-- BAD: DISTINCT hides a faulty join that produces duplicates
SELECT DISTINCT c.id, c.name, c.email
FROM customers c
JOIN orders o ON o.customer_id = c.id
JOIN order_items oi ON oi.order_id = o.id;

-- GOOD: fix the join — if you want customers, don't join to detail tables needlessly
SELECT c.id, c.name, c.email
FROM customers c
WHERE EXISTS (
  SELECT 1 FROM orders o
  JOIN order_items oi ON oi.order_id = o.id
  WHERE o.customer_id = c.id
);
```

**Why it's bad**: DISTINCT is expensive (sort or hash all output rows). If you need it,
your query is likely producing unintended duplicates due to a many-to-many join.
Fix the query logic instead.

### Anti-Pattern 10: Implicit Casting Preventing Index Use

```sql
-- BAD: comparing varchar column to integer causes implicit cast
-- The index on account_number (varchar) cannot be used
SELECT * FROM accounts WHERE account_number = 12345;

-- GOOD: use matching type
SELECT * FROM accounts WHERE account_number = '12345';
```

**Why it's bad**: When types don't match, PostgreSQL casts the column value for each row
instead of casting the literal once. This prevents index use because the index stores
the original type. Always match types in predicates.

### Anti-Pattern 11: Excessive CTE Materialization

```sql
-- BAD (pre-PG 12): CTEs are always materialized, acting as optimization fences
WITH active_users AS (
  SELECT * FROM users WHERE active = true
)
SELECT * FROM active_users WHERE email = 'user@example.com';
-- The email filter cannot be pushed into the CTE scan

-- GOOD (PG 12+): use NOT MATERIALIZED to allow predicate pushdown
WITH active_users AS NOT MATERIALIZED (
  SELECT * FROM users WHERE active = true
)
SELECT * FROM active_users WHERE email = 'user@example.com';
-- Now the planner can push the email filter through

-- GOOD (any version): inline the subquery
SELECT * FROM users WHERE active = true AND email = 'user@example.com';
```

**Why it's bad**: In PostgreSQL < 12, CTEs are always materialized — the entire result
is computed and stored in memory/disk before the outer query can filter it. This
prevents the planner from pushing predicates down. In PG 12+, single-reference CTEs
are automatically inlined unless you force materialization.

### Anti-Pattern 12: N+1 Queries

```python
# BAD: application issues one query per row
customers = db.query("SELECT id FROM customers WHERE region = 'US'")
for c in customers:
    orders = db.query(f"SELECT * FROM orders WHERE customer_id = {c.id}")
    # ... process orders

# GOOD: batch with IN or JOIN
customer_ids = [c.id for c in customers]
orders = db.query("""
    SELECT o.* FROM orders o
    WHERE o.customer_id = ANY(%s)
""", [customer_ids])

# GOOD: single query with JOIN
results = db.query("""
    SELECT c.id, c.name, o.id AS order_id, o.total
    FROM customers c
    JOIN orders o ON o.customer_id = c.id
    WHERE c.region = 'US'
""")
```

**Why it's bad**: Each query incurs network round-trip latency, parse/plan overhead,
and prevents the database from optimizing the access pattern. Batching with IN/ANY
or using JOINs lets PostgreSQL optimize the entire operation.

---

## 5. CTE Materialization Control

### Pre-PostgreSQL 12

All CTEs are materialized. They act as "optimization fences" — the planner cannot
push predicates into or pull predicates out of a CTE.

```sql
-- This always materializes the full result of expensive_cte
WITH expensive_cte AS (
  SELECT * FROM large_table WHERE category = 'A'
)
SELECT * FROM expensive_cte WHERE id = 42;
-- The id = 42 filter is applied AFTER the CTE produces all rows
```

### PostgreSQL 12+: Automatic Inlining

CTEs referenced only once are automatically inlined (treated like subqueries) unless
they have side effects (INSERT/UPDATE/DELETE with RETURNING) or are recursive.

```sql
-- PG 12+: this CTE is automatically inlined (equivalent to a subquery)
WITH recent_orders AS (
  SELECT * FROM orders WHERE created_at > now() - interval '7 days'
)
SELECT * FROM recent_orders WHERE total > 100;
-- Planner combines predicates: created_at > ... AND total > 100
```

### Explicit Control

```sql
-- Force materialization (useful when CTE is referenced multiple times
-- and you want to compute it once)
WITH popular_products AS MATERIALIZED (
  SELECT product_id, COUNT(*) AS order_count
  FROM order_items
  GROUP BY product_id
  HAVING COUNT(*) > 100
)
SELECT * FROM popular_products p1
JOIN popular_products p2 ON p1.product_id != p2.product_id;

-- Force inlining (useful when the CTE is large but heavily filtered)
WITH all_events AS NOT MATERIALIZED (
  SELECT * FROM events  -- millions of rows
)
SELECT * FROM all_events WHERE user_id = 42 AND event_type = 'click';
-- Planner can push user_id and event_type predicates down
```

### When to Use MATERIALIZED

- The CTE is referenced multiple times and is expensive to recompute.
- You want a consistent snapshot (the CTE result won't change between references).
- The CTE result is small and fits easily in work_mem.

### When to Use NOT MATERIALIZED

- The CTE produces a large result but the outer query filters aggressively.
- You want the planner to push predicates down for better index use.
- Performance is worse with materialization due to large intermediate result.

---

## 6. Parallel Query Tuning

### Key Settings

```sql
-- Maximum parallel workers per query (default: 2)
SET max_parallel_workers_per_gather = 4;

-- Total parallel workers across all queries (default: 8)
SET max_parallel_workers = 8;

-- Minimum table size to consider parallel scan (default: 8MB)
SET min_parallel_table_scan_size = '8MB';

-- Minimum index size to consider parallel scan (default: 512kB)
SET min_parallel_index_scan_size = '512kB';

-- Cost of launching a parallel worker (default: 1000)
SET parallel_setup_cost = 1000;

-- Cost of passing a tuple from worker to leader (default: 0.1)
SET parallel_tuple_cost = 0.1;
```

### When Parallel Query Helps

- Large sequential scans on big tables (millions of rows).
- Aggregations over many rows (COUNT, SUM, AVG on large tables).
- Hash joins and merge joins on large datasets.
- Parallel index scans (PG 12+) for bitmap heap scans.
- Parallel CREATE INDEX (PG 11+).

### When Parallel Query Hurts

- OLTP workloads with many concurrent short queries (context-switching overhead).
- Queries that return quickly already (< 10 ms) — worker startup cost dominates.
- Very small tables where the overhead exceeds the benefit.
- Queries with functions marked `PARALLEL UNSAFE`.

### Parallel Safety Levels for Functions

```sql
-- Mark a function as safe for parallel execution
CREATE OR REPLACE FUNCTION my_func(x int) RETURNS int
LANGUAGE sql PARALLEL SAFE
AS $$ SELECT x * 2; $$;

-- PARALLEL RESTRICTED: can run in parallel worker but not pushed down
-- PARALLEL UNSAFE: forces serial execution (default for PL/pgSQL)
```

### Verifying Parallel Execution

```sql
EXPLAIN (ANALYZE) SELECT COUNT(*) FROM large_table;

-- Look for:
-- Gather (actual workers launched: 4)
--   Workers Planned: 4
--   Workers Launched: 4
--   ->  Partial Aggregate
--        ->  Parallel Seq Scan on large_table
```

---

## 7. JIT Compilation

### Overview

JIT (Just-In-Time compilation) compiles query expressions and tuple deforming into
native machine code using LLVM at runtime. Available since PostgreSQL 11.

### Key Settings

```sql
-- Enable/disable JIT (default: on in PG 12+)
SET jit = on;

-- Cost thresholds for JIT activation
SET jit_above_cost = 100000;           -- enable JIT if query cost > this
SET jit_inline_above_cost = 500000;    -- inline functions if cost > this
SET jit_optimize_above_cost = 500000;  -- apply LLVM optimizations if cost > this
```

### When JIT Helps

- Long-running analytical queries with complex expressions.
- Queries that process millions of rows with WHERE filters, aggregations.
- Queries with many column evaluations (wide tables, computed columns).
- CPU-bound queries where expression evaluation dominates.

### When to Disable JIT

```sql
-- Disable for short OLTP queries where compile time > savings
SET jit = off;
```

- Short queries (< 100 ms) — JIT compilation overhead may exceed savings.
- Queries that are I/O-bound rather than CPU-bound.
- When you see high `JIT: Generation Time` in EXPLAIN output relative to execution.
- Prepared statements that are executed many times (JIT happens per execution).

### Reading JIT in EXPLAIN Output

```sql
EXPLAIN (ANALYZE) SELECT sum(price * quantity) FROM order_items WHERE category_id = 5;

-- JIT section at the bottom:
-- JIT:
--   Functions: 6
--   Options: Inlining true, Optimization true, Expressions true, Deforming true
--   Timing: Generation 1.234 ms, Inlining 5.678 ms, Optimization 12.345 ms,
--           Emission 8.901 ms, Total 28.158 ms
```

If `Total` JIT time is a large fraction of execution time, consider disabling JIT for
that query.

---

## 8. Query Planner Statistics

### ANALYZE: Collecting Statistics

```sql
-- Analyze a specific table
ANALYZE orders;

-- Analyze specific columns
ANALYZE orders (customer_id, status);

-- Analyze all tables in the database
ANALYZE;

-- Verbose output showing progress
ANALYZE VERBOSE orders;
```

ANALYZE collects statistics about column value distribution (most common values,
histogram, distinct count, NULL fraction, correlation). The planner uses these to
estimate selectivity and choose optimal plans.

### Autovacuum and Auto-Analyze

```sql
-- Check last analyze time
SELECT schemaname, relname, last_analyze, last_autoanalyze, n_mod_since_analyze
FROM pg_stat_user_tables
ORDER BY n_mod_since_analyze DESC;
```

Auto-analyze runs when `n_mod_since_analyze` exceeds
`autovacuum_analyze_threshold + autovacuum_analyze_scale_factor * reltuples`.
Default is 50 + 10% of the table. For large tables, you may want to lower the
scale factor:

```sql
ALTER TABLE orders SET (autovacuum_analyze_scale_factor = 0.02);
```

### default_statistics_target

```sql
-- Global setting (default: 100, range: 1-10000)
SET default_statistics_target = 200;

-- Per-column override for columns with skewed distributions
ALTER TABLE orders ALTER COLUMN status SET STATISTICS 500;

-- After changing statistics target, re-analyze
ANALYZE orders;
```

Higher values mean more histogram buckets and more common values tracked, which
improves selectivity estimates for skewed data at the cost of more planning time.

### Extended Statistics (CREATE STATISTICS)

Standard statistics are collected per-column independently. Extended statistics capture
multi-column correlations, which help when WHERE clauses filter on multiple correlated
columns.

```sql
-- Functional dependencies: tells planner that city depends on zip_code
CREATE STATISTICS stats_zip_city (dependencies)
ON zip_code, city FROM addresses;

-- Distinct count: tells planner the combined distinct count of columns
CREATE STATISTICS stats_product_combos (ndistinct)
ON category_id, brand_id FROM products;

-- MCV (Most Common Values): multi-column value frequencies (PG 12+)
CREATE STATISTICS stats_status_priority (mcv)
ON status, priority FROM tickets;

-- Combine multiple types
CREATE STATISTICS stats_orders_multi (dependencies, ndistinct, mcv)
ON customer_id, product_id, status FROM orders;

-- After creating, re-analyze
ANALYZE orders;

-- View extended statistics
SELECT * FROM pg_statistic_ext;
SELECT * FROM pg_statistic_ext_data;
```

### Diagnosing Bad Estimates

```sql
-- Compare estimated vs actual rows in EXPLAIN ANALYZE output
-- Look for nodes where rows estimate is far off from actual rows:
--   estimated: 1 row, actual: 50,000 rows  --> needs better stats
--   estimated: 100,000 rows, actual: 1 row  --> over-estimation

-- Check column statistics
SELECT attname, n_distinct, most_common_vals, most_common_freqs, correlation
FROM pg_stats
WHERE tablename = 'orders' AND attname = 'status';
```

When estimated and actual row counts differ by more than 10x, consider:

1. Running `ANALYZE` on the table.
2. Increasing `STATISTICS` target for the column.
3. Creating extended statistics if the estimation error involves multiple columns.
4. Checking for type mismatches or implicit casts that confuse the planner.

### Forcing Plan Choices (Last Resort)

```sql
-- These are debug tools, not production solutions
SET enable_seqscan = off;      -- force index usage
SET enable_nestloop = off;     -- force hash/merge joins
SET random_page_cost = 1.0;    -- tell planner random I/O = sequential I/O (SSD)
SET seq_page_cost = 1.0;       -- baseline sequential page cost

-- For SSDs, these settings are commonly adjusted:
SET random_page_cost = 1.1;    -- nearly equal to sequential on SSD
SET effective_cache_size = '24GB';  -- tell planner how much data is cached
```

`pg_hint_plan` extension allows per-query hints without changing GUC settings:

```sql
-- Using pg_hint_plan extension
/*+ IndexScan(orders orders_customer_id_idx) */
SELECT * FROM orders WHERE customer_id = 42;

/*+ HashJoin(orders customers) */
SELECT * FROM orders o JOIN customers c ON o.customer_id = c.id;
```
