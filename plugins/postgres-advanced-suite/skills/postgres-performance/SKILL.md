---
name: postgres-performance
description: >
  PostgreSQL performance tuning — EXPLAIN ANALYZE, configuration tuning, connection
  pooling, vacuum strategies, monitoring, and diagnosing slow queries.
  Triggers: "postgres performance", "slow query", "explain analyze", "postgresql tuning",
  "connection pooling", "pgbouncer", "vacuum", "postgres monitoring".
  NOT for: index-specific optimization (use postgres-indexing), migrations (use postgres-migrations).
version: 1.0.0
allowed-tools: Read, Grep, Glob, Bash, Edit, Write
---

# PostgreSQL Performance

## EXPLAIN ANALYZE

### The Command

```sql
-- Always use ANALYZE + BUFFERS for real performance data
EXPLAIN (ANALYZE, BUFFERS, FORMAT TEXT)
SELECT u.name, COUNT(p.id) as post_count
FROM users u
JOIN posts p ON p.author_id = u.id
WHERE u.created_at > '2025-01-01'
GROUP BY u.id
ORDER BY post_count DESC
LIMIT 10;
```

### Reading the Output

```
Limit  (cost=1234.56..1234.58 rows=10 width=40) (actual time=45.123..45.130 rows=10 loops=1)
  ->  Sort  (cost=1234.56..1240.12 rows=2224 width=40) (actual time=45.121..45.125 rows=10 loops=1)
        Sort Key: (count(p.id)) DESC
        Sort Method: top-N heapsort  Memory: 26kB
        ->  HashAggregate  (cost=1180.00..1202.24 rows=2224 width=40) (actual time=44.500..44.800 rows=2224 loops=1)
              Group Key: u.id
              Batches: 1  Memory Usage: 369kB
              ->  Hash Join  (cost=100.00..1100.00 rows=16000 width=36) (actual time=5.200..40.100 rows=16000 loops=1)
                    Hash Cond: (p.author_id = u.id)
                    ->  Seq Scan on posts p  (cost=0.00..800.00 rows=50000 width=8) (actual time=0.010..15.000 rows=50000 loops=1)
                          Buffers: shared hit=300
                    ->  Hash  (cost=80.00..80.00 rows=2224 width=36) (actual time=5.100..5.100 rows=2224 loops=1)
                          Buckets: 4096  Batches: 1  Memory Usage: 160kB
                          ->  Seq Scan on users u  (cost=0.00..80.00 rows=2224 width=36) (actual time=0.010..3.500 rows=2224 loops=1)
                                Filter: (created_at > '2025-01-01'::date)
                                Rows Removed by Filter: 776
                                Buffers: shared hit=40
Planning Time: 0.250 ms
Execution Time: 45.200 ms
```

### What to Look For

| Signal | Problem | Fix |
|--------|---------|-----|
| `Seq Scan` on 100K+ rows | No index | Create index on filter/join columns |
| `Rows Removed by Filter: 99%+` | Very unselective scan | Better index or query restructure |
| `actual rows` >> `rows` (estimated) | Stale statistics | `ANALYZE tablename;` |
| `Sort Method: external merge` | Sort spills to disk | `SET work_mem = '256MB';` for this query |
| `Buffers: shared read` >> `hit` | Cold cache or table too large | Check shared_buffers, consider partitioning |
| `Hash Batches: 8` | Hash spills to disk | Increase `work_mem` |
| `Nested Loop` with `loops=10000` | O(n*m) join | Add index on join key, or force hash join |

## Configuration Tuning

### Essential Parameters

```ini
# postgresql.conf — Start here for a dedicated server with 16GB RAM

# === MEMORY ===
shared_buffers = 4GB              # 25% of RAM (caches table/index pages)
effective_cache_size = 12GB       # 75% of RAM (tells planner about OS cache)
work_mem = 64MB                   # Per-sort/hash operation (careful: multiplied by parallel workers)
maintenance_work_mem = 1GB        # For VACUUM, CREATE INDEX, ALTER TABLE

# === WAL / WRITE PERFORMANCE ===
wal_buffers = 64MB                # WAL write buffer (auto-tuned if set to -1)
checkpoint_completion_target = 0.9 # Spread checkpoint writes
min_wal_size = 1GB
max_wal_size = 4GB

# === QUERY PLANNER ===
random_page_cost = 1.1            # SSD (default 4.0 is for spinning disk!)
effective_io_concurrency = 200    # SSD (default 1 is for spinning disk)
default_statistics_target = 100   # Accuracy of planner estimates (increase for complex queries)

# === CONNECTIONS ===
max_connections = 100             # Keep low! Use connection pooling
```

### Quick Tuning by RAM

| RAM | shared_buffers | effective_cache_size | work_mem | maintenance_work_mem |
|-----|---------------|---------------------|----------|---------------------|
| 1GB | 256MB | 768MB | 4MB | 64MB |
| 4GB | 1GB | 3GB | 16MB | 256MB |
| 8GB | 2GB | 6GB | 32MB | 512MB |
| 16GB | 4GB | 12GB | 64MB | 1GB |
| 32GB | 8GB | 24GB | 128MB | 2GB |
| 64GB | 16GB | 48GB | 256MB | 4GB |

### Per-Query Memory Tuning

```sql
-- Temporarily increase work_mem for a single complex query
SET LOCAL work_mem = '256MB';
SELECT ... complex aggregation ...;
-- Resets at end of transaction
```

## Connection Pooling

### Why You Need It

PostgreSQL forks a new process per connection (~10MB each). 100 connections = 1GB just for processes. Most apps need 5-20 actual connections, not 100.

### PgBouncer Setup

```ini
# pgbouncer.ini
[databases]
myapp = host=localhost port=5432 dbname=myapp

[pgbouncer]
listen_addr = 0.0.0.0
listen_port = 6432
auth_type = md5
auth_file = /etc/pgbouncer/userlist.txt

# Pool settings
pool_mode = transaction        # Best for most apps
default_pool_size = 20         # Connections per database/user pair
max_client_conn = 200          # Total client connections accepted
min_pool_size = 5              # Keep this many connections open
reserve_pool_size = 5          # Extra connections for burst
reserve_pool_timeout = 3       # Seconds before using reserve
```

### Pool Modes

| Mode | Behavior | Best For |
|------|----------|----------|
| `session` | Connection held until client disconnects | Legacy apps, prepared statements |
| `transaction` | Connection returned after each transaction | Most web apps (recommended) |
| `statement` | Connection returned after each statement | Simple queries, autocommit |

### Application-Level Pooling (Node.js)

```typescript
// If using Prisma (built-in pooling)
const prisma = new PrismaClient({
  datasources: {
    db: {
      url: process.env.DATABASE_URL, // ?connection_limit=10&pool_timeout=30
    },
  },
});

// If using pg directly
import { Pool } from 'pg';

const pool = new Pool({
  connectionString: process.env.DATABASE_URL,
  max: 10,                    // Max connections in pool
  idleTimeoutMillis: 30000,   // Close idle connections after 30s
  connectionTimeoutMillis: 5000, // Fail if can't connect in 5s
});
```

## VACUUM Strategy

### What VACUUM Does

1. Reclaims space from deleted/updated rows (dead tuples)
2. Updates visibility map (for Index Only Scans)
3. Updates free space map (for INSERT performance)
4. Prevents transaction ID wraparound

### Autovacuum Tuning

```ini
# For a high-write workload
autovacuum_vacuum_scale_factor = 0.05    # VACUUM at 5% dead tuples (default 20%)
autovacuum_analyze_scale_factor = 0.02   # ANALYZE at 2% changed rows (default 10%)
autovacuum_vacuum_cost_delay = 2ms       # Less sleep between vacuum I/O (default 20ms)
autovacuum_max_workers = 4               # Parallel autovacuum workers (default 3)
```

### Per-Table Settings

```sql
-- Hot table with millions of writes/day
ALTER TABLE events SET (
  autovacuum_vacuum_scale_factor = 0.01,      -- VACUUM at 1% dead
  autovacuum_analyze_scale_factor = 0.005,    -- ANALYZE at 0.5% changed
  autovacuum_vacuum_cost_delay = 0            -- No delay (aggressive)
);
```

### Manual VACUUM

```sql
-- Regular vacuum (reclaims space for reuse, non-blocking)
VACUUM (VERBOSE) tablename;

-- Analyze (update statistics only)
ANALYZE tablename;

-- Both
VACUUM ANALYZE tablename;

-- Full vacuum (reclaims space to OS, BLOCKS ALL ACCESS)
-- Only use during maintenance windows
VACUUM FULL tablename;
```

## Monitoring Queries

### Slow Query Identification

```sql
-- Enable slow query logging
-- In postgresql.conf:
-- log_min_duration_statement = 100   -- Log queries > 100ms

-- Find slow queries from pg_stat_statements
SELECT
  calls,
  round(total_exec_time::numeric / calls, 2) AS avg_ms,
  round(total_exec_time::numeric, 2) AS total_ms,
  query
FROM pg_stat_statements
ORDER BY total_exec_time DESC
LIMIT 20;
```

### Table Bloat

```sql
-- Check for table bloat (dead tuples)
SELECT
  schemaname,
  relname,
  n_live_tup,
  n_dead_tup,
  round(n_dead_tup::numeric / GREATEST(n_live_tup, 1) * 100, 1) AS dead_pct,
  last_vacuum,
  last_autovacuum
FROM pg_stat_user_tables
WHERE n_dead_tup > 1000
ORDER BY n_dead_tup DESC;
```

### Cache Hit Ratio

```sql
-- Should be > 99% for production
SELECT
  sum(heap_blks_read) AS heap_read,
  sum(heap_blks_hit) AS heap_hit,
  round(sum(heap_blks_hit)::numeric / GREATEST(sum(heap_blks_hit) + sum(heap_blks_read), 1) * 100, 2) AS cache_hit_ratio
FROM pg_statio_user_tables;
```

### Active Connections

```sql
SELECT
  state,
  count(*),
  max(now() - state_change) AS max_duration
FROM pg_stat_activity
WHERE datname = current_database()
GROUP BY state;
```

### Lock Monitoring

```sql
-- Find blocking queries
SELECT
  blocked.pid AS blocked_pid,
  blocked.query AS blocked_query,
  blocking.pid AS blocking_pid,
  blocking.query AS blocking_query,
  now() - blocked.query_start AS blocked_duration
FROM pg_stat_activity blocked
JOIN pg_locks bl ON bl.pid = blocked.pid
JOIN pg_locks kl ON kl.locktype = bl.locktype
  AND kl.database IS NOT DISTINCT FROM bl.database
  AND kl.relation IS NOT DISTINCT FROM bl.relation
  AND kl.page IS NOT DISTINCT FROM bl.page
  AND kl.tuple IS NOT DISTINCT FROM bl.tuple
  AND kl.transactionid IS NOT DISTINCT FROM bl.transactionid
  AND kl.classid IS NOT DISTINCT FROM bl.classid
  AND kl.objid IS NOT DISTINCT FROM bl.objid
  AND kl.objsubid IS NOT DISTINCT FROM bl.objsubid
  AND kl.pid != bl.pid
JOIN pg_stat_activity blocking ON blocking.pid = kl.pid
WHERE NOT bl.granted;
```

## Common Gotchas

1. **`random_page_cost = 4.0` is the default** — change to `1.1` for SSD. The default assumes spinning disks and discourages index usage.

2. **`work_mem` is per-sort, not per-query** — a query with 5 sorts uses 5x work_mem. With parallel workers, multiply again. Don't set too high globally.

3. **Prepared statements break PgBouncer transaction mode** — use session mode if your app uses prepared statements, or configure PgBouncer's `max_prepared_statements`.

4. **VACUUM FULL locks the table** — never run in production without a maintenance window. Regular VACUUM is non-blocking and usually sufficient.

5. **Statistics drift** — run `ANALYZE` after bulk inserts/deletes. The planner uses stale statistics → bad query plans.

6. **Too many connections** — PostgreSQL performs worse with > 200 active connections. Use connection pooling. The sweet spot is `2 * CPU cores + effective_spindle_count`.

7. **Transaction ID wraparound** — if autovacuum falls behind, PostgreSQL will eventually refuse writes to prevent data corruption. Monitor `age(datfrozenxid)`.

8. **Indexes slow down writes** — every INSERT/UPDATE/DELETE must update all indexes. A table with 10 indexes is 10x more write work. Only keep indexes you actually use.
