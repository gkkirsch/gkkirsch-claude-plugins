---
name: postgres-performance
description: >
  PostgreSQL performance tuning — query optimization, EXPLAIN ANALYZE,
  indexing strategies, connection pooling, configuration tuning,
  partitioning, vacuum management, and performance monitoring.
  Triggers: "postgres performance", "postgres slow query", "explain analyze",
  "postgres index", "postgres tuning", "pgbouncer", "postgres vacuum",
  "postgres partition", "query optimization".
  NOT for: Backup, replication, or operational tasks (use postgres-operations).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# PostgreSQL Performance

## EXPLAIN ANALYZE

```sql
-- Always use ANALYZE for actual execution times (not estimates)
EXPLAIN (ANALYZE, BUFFERS, FORMAT TEXT)
SELECT u.id, u.email, COUNT(o.id) as order_count
FROM users u
LEFT JOIN orders o ON o.user_id = u.id
WHERE u.created_at > '2024-01-01'
GROUP BY u.id, u.email
HAVING COUNT(o.id) > 5
ORDER BY order_count DESC
LIMIT 20;

-- Reading the output:
-- Seq Scan on users  (cost=0.00..1234.00 rows=500 width=40)
--                     (actual time=0.05..12.34 rows=487 loops=1)
--   Buffers: shared hit=100 read=50
--
-- Key metrics:
-- actual time: startup..total in milliseconds
-- rows: actual rows returned (compare with estimated "rows=500")
-- Buffers shared hit: pages from cache (good)
-- Buffers shared read: pages from disk (slow)
-- loops: how many times this node executed

-- Compare estimated vs actual rows — big differences = stale statistics
ANALYZE users;  -- Update statistics for the table
```

## Indexing Strategies

```sql
-- B-tree (default) — equality and range queries
CREATE INDEX idx_users_email ON users (email);
CREATE INDEX idx_orders_created ON orders (created_at DESC);

-- Composite index — column order matters
-- Good for: WHERE status = 'active' AND created_at > '2024-01-01'
-- Also covers: WHERE status = 'active' (prefix)
-- Does NOT cover: WHERE created_at > '2024-01-01' alone
CREATE INDEX idx_orders_status_created
  ON orders (status, created_at DESC);

-- Partial index — smaller, faster, for common queries
CREATE INDEX idx_orders_pending
  ON orders (created_at DESC)
  WHERE status = 'pending';
-- Only indexes pending orders — much smaller than full index

-- Covering index (INCLUDE) — avoids table lookups
CREATE INDEX idx_orders_user_covering
  ON orders (user_id)
  INCLUDE (status, total, created_at);
-- The query can be answered entirely from the index

-- GIN index — for JSONB, arrays, full-text search
CREATE INDEX idx_users_metadata ON users USING GIN (metadata);
CREATE INDEX idx_products_tags ON products USING GIN (tags);
CREATE INDEX idx_posts_search ON posts USING GIN (
  to_tsvector('english', title || ' ' || body)
);

-- BRIN index — for naturally ordered data (timestamps, sequences)
CREATE INDEX idx_events_created ON events USING BRIN (created_at);
-- Tiny index (much smaller than B-tree), great for append-only tables

-- Expression index
CREATE INDEX idx_users_lower_email ON users (LOWER(email));
-- Matches: WHERE LOWER(email) = 'user@example.com'

-- Concurrent index creation (doesn't lock the table)
CREATE INDEX CONCURRENTLY idx_orders_total ON orders (total);

-- Find unused indexes
SELECT
  schemaname, tablename, indexname,
  idx_scan as times_used,
  pg_size_pretty(pg_relation_size(indexrelid)) as index_size
FROM pg_stat_user_indexes
WHERE idx_scan = 0
  AND indexrelid NOT IN (
    SELECT conindid FROM pg_constraint WHERE contype IN ('p', 'u')
  )
ORDER BY pg_relation_size(indexrelid) DESC;

-- Find missing indexes (sequential scans on large tables)
SELECT
  schemaname, relname as table_name,
  seq_scan, seq_tup_read,
  idx_scan, idx_tup_fetch,
  pg_size_pretty(pg_relation_size(relid)) as table_size
FROM pg_stat_user_tables
WHERE seq_scan > 100
  AND pg_relation_size(relid) > 10000000  -- >10MB
ORDER BY seq_tup_read DESC
LIMIT 20;
```

## Query Optimization Patterns

```sql
-- ANTI-PATTERN: SELECT * in application queries
-- Fix: select only needed columns
SELECT id, email, name FROM users WHERE id = $1;

-- ANTI-PATTERN: N+1 queries
-- Bad: SELECT * FROM orders WHERE user_id = $1 (in a loop)
-- Fix: batch with ANY or JOIN
SELECT * FROM orders WHERE user_id = ANY($1::uuid[]);

-- ANTI-PATTERN: OFFSET for pagination (scans and discards rows)
-- Bad:
SELECT * FROM orders ORDER BY created_at DESC OFFSET 10000 LIMIT 20;
-- Fix: cursor-based pagination
SELECT * FROM orders
WHERE created_at < $1  -- cursor from previous page
ORDER BY created_at DESC
LIMIT 20;

-- ANTI-PATTERN: OR on different columns (prevents index use)
-- Bad:
SELECT * FROM users WHERE email = $1 OR phone = $2;
-- Fix: UNION ALL
SELECT * FROM users WHERE email = $1
UNION ALL
SELECT * FROM users WHERE phone = $2 AND email != $1;

-- ANTI-PATTERN: function on indexed column
-- Bad: WHERE LOWER(email) = 'user@test.com' (doesn't use idx_users_email)
-- Fix: expression index or store lowercase
CREATE INDEX idx_users_email_lower ON users (LOWER(email));

-- EXISTS vs IN for subqueries
-- EXISTS stops at first match (faster for large subquery results)
SELECT * FROM users u
WHERE EXISTS (
  SELECT 1 FROM orders o WHERE o.user_id = u.id AND o.total > 100
);

-- CTE materialization control (PostgreSQL 12+)
WITH recent_orders AS NOT MATERIALIZED (
  SELECT * FROM orders WHERE created_at > NOW() - INTERVAL '7 days'
)
SELECT u.email, COUNT(ro.id)
FROM users u
JOIN recent_orders ro ON ro.user_id = u.id
GROUP BY u.email;
-- NOT MATERIALIZED lets the planner inline the CTE and optimize together
```

## Connection Pooling (PgBouncer)

```ini
; /etc/pgbouncer/pgbouncer.ini
[databases]
myapp = host=db-host.example.com port=5432 dbname=myapp

[pgbouncer]
listen_addr = 0.0.0.0
listen_port = 6432
auth_type = md5
auth_file = /etc/pgbouncer/userlist.txt

; Pool mode:
; session    = connection held for entire client session (safest, least efficient)
; transaction = connection returned after each transaction (recommended)
; statement  = connection returned after each statement (most aggressive)
pool_mode = transaction

; Pool sizing
default_pool_size = 20        ; connections per user/db pair
min_pool_size = 5             ; minimum idle connections
reserve_pool_size = 5         ; extra connections for burst
reserve_pool_timeout = 3      ; seconds before using reserve pool
max_client_conn = 200         ; total client connections accepted
max_db_connections = 50       ; max connections to actual database

; Timeouts
server_idle_timeout = 300     ; close idle server connections after 5 min
client_idle_timeout = 0       ; never close idle client connections
query_timeout = 30            ; kill queries running longer than 30s
client_login_timeout = 60
server_connect_timeout = 15

; Logging
log_connections = 1
log_disconnections = 1
log_pooler_errors = 1
stats_period = 60
```

```sql
-- Monitor PgBouncer
SHOW POOLS;
SHOW STATS;
SHOW CLIENTS;
SHOW SERVERS;

-- Check wait times (high = pool exhaustion)
SELECT database, user, cl_active, cl_waiting, sv_active, sv_idle,
       maxwait, maxwait_us
FROM pgbouncer.pools;
```

## Configuration Tuning

```sql
-- Key parameters to tune (for a server with 16GB RAM, 4 cores)

-- Memory
ALTER SYSTEM SET shared_buffers = '4GB';          -- 25% of RAM
ALTER SYSTEM SET effective_cache_size = '12GB';    -- 75% of RAM
ALTER SYSTEM SET work_mem = '64MB';                -- Per-sort/hash operation
ALTER SYSTEM SET maintenance_work_mem = '1GB';     -- VACUUM, CREATE INDEX

-- Write-ahead log
ALTER SYSTEM SET wal_buffers = '64MB';
ALTER SYSTEM SET min_wal_size = '1GB';
ALTER SYSTEM SET max_wal_size = '4GB';
ALTER SYSTEM SET checkpoint_completion_target = 0.9;

-- Parallelism
ALTER SYSTEM SET max_parallel_workers_per_gather = 2;
ALTER SYSTEM SET max_parallel_workers = 4;
ALTER SYSTEM SET max_worker_processes = 8;

-- Query planner
ALTER SYSTEM SET random_page_cost = 1.1;           -- SSD storage (default 4.0 for HDD)
ALTER SYSTEM SET effective_io_concurrency = 200;    -- SSD
ALTER SYSTEM SET default_statistics_target = 200;   -- More accurate plans (default 100)

-- Connections
ALTER SYSTEM SET max_connections = 200;             -- Use with PgBouncer

-- Apply changes
SELECT pg_reload_conf();
```

## Table Partitioning

```sql
-- Range partitioning by date (most common)
CREATE TABLE events (
  id          BIGINT GENERATED ALWAYS AS IDENTITY,
  event_type  TEXT NOT NULL,
  payload     JSONB,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
) PARTITION BY RANGE (created_at);

-- Create monthly partitions
CREATE TABLE events_2024_01 PARTITION OF events
  FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');
CREATE TABLE events_2024_02 PARTITION OF events
  FOR VALUES FROM ('2024-02-01') TO ('2024-03-01');

-- Default partition catches anything without a match
CREATE TABLE events_default PARTITION OF events DEFAULT;

-- Indexes on partitioned tables apply to each partition
CREATE INDEX idx_events_type ON events (event_type);
CREATE INDEX idx_events_created ON events (created_at DESC);

-- Automated partition creation (run monthly via cron)
DO $$
DECLARE
  start_date DATE := DATE_TRUNC('month', NOW() + INTERVAL '1 month');
  end_date DATE := start_date + INTERVAL '1 month';
  partition_name TEXT := 'events_' || TO_CHAR(start_date, 'YYYY_MM');
BEGIN
  EXECUTE FORMAT(
    'CREATE TABLE IF NOT EXISTS %I PARTITION OF events FOR VALUES FROM (%L) TO (%L)',
    partition_name, start_date, end_date
  );
END $$;

-- Drop old partitions (faster than DELETE)
DROP TABLE events_2023_01;
```

## VACUUM and Maintenance

```sql
-- Check tables needing vacuum
SELECT
  schemaname, relname,
  n_dead_tup,
  n_live_tup,
  ROUND(n_dead_tup::numeric / GREATEST(n_live_tup, 1) * 100, 1) as dead_pct,
  last_vacuum,
  last_autovacuum,
  last_analyze
FROM pg_stat_user_tables
WHERE n_dead_tup > 1000
ORDER BY n_dead_tup DESC;

-- Tune autovacuum per table (heavy-write tables need aggressive settings)
ALTER TABLE orders SET (
  autovacuum_vacuum_scale_factor = 0.01,    -- vacuum at 1% dead tuples (default 20%)
  autovacuum_analyze_scale_factor = 0.005,  -- analyze at 0.5% changed
  autovacuum_vacuum_cost_delay = 2          -- less delay = faster vacuum
);

-- Monitor vacuum progress
SELECT
  relid::regclass as table_name,
  phase,
  heap_blks_total,
  heap_blks_scanned,
  heap_blks_vacuumed,
  ROUND(heap_blks_vacuumed::numeric / GREATEST(heap_blks_total, 1) * 100, 1) as pct_complete
FROM pg_stat_progress_vacuum;

-- Table bloat estimation
SELECT
  schemaname, tablename,
  pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as total_size,
  pg_size_pretty(pg_relation_size(schemaname||'.'||tablename)) as table_size,
  pg_size_pretty(pg_indexes_size(schemaname||'.'||tablename::regclass)) as index_size
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC
LIMIT 20;
```

## Gotchas

1. **`work_mem` is per-operation, not per-query** — A query with 5 sort/hash operations uses 5x `work_mem`. Setting it to 256MB with 100 concurrent connections = potential 128GB memory usage. Keep it conservative (16-64MB) globally and increase per-session for batch jobs: `SET work_mem = '256MB';`.

2. **OFFSET pagination gets slower as you go deeper** — `OFFSET 100000 LIMIT 20` scans and discards 100,000 rows before returning 20. Use cursor-based pagination with `WHERE created_at < $last_seen ORDER BY created_at DESC LIMIT 20`. The query time is constant regardless of page depth.

3. **Partial indexes need matching WHERE clauses** — `CREATE INDEX idx_pending ON orders (created_at) WHERE status = 'pending'` only helps queries that include `WHERE status = 'pending'` in their predicate. The planner won't use it for `WHERE status = 'active'` or queries without a status filter.

4. **VACUUM FULL locks the table** — `VACUUM FULL` rewrites the entire table and holds an exclusive lock for the duration. For a 100GB table, this can take hours. Use `pg_repack` extension instead for online table repacking without exclusive locks.

5. **PgBouncer transaction mode breaks session features** — In transaction pooling mode, `SET`, `LISTEN/NOTIFY`, prepared statements, advisory locks, and temp tables don't work because each transaction may use a different server connection. Test your application thoroughly before switching from session to transaction mode.

6. **Autovacuum defaults are too conservative for high-write tables** — The default `autovacuum_vacuum_scale_factor = 0.2` means a table with 10M rows needs 2M dead tuples before autovacuum runs. Tables with heavy UPDATE/DELETE patterns accumulate bloat and slow down. Set per-table thresholds: `ALTER TABLE hot_table SET (autovacuum_vacuum_scale_factor = 0.01)`.
