# PostgreSQL Quick Reference

## Data Type Selection

| Need | Use | Not |
|------|-----|-----|
| Auto-increment ID | `BIGSERIAL` or `GENERATED ALWAYS AS IDENTITY` | `SERIAL` (legacy) |
| UUID primary key | `UUID DEFAULT gen_random_uuid()` | `TEXT` with app-generated UUID |
| Short text (name, email) | `TEXT` | `VARCHAR(n)` — no performance benefit in PG |
| Long text (content, body) | `TEXT` | `TEXT` is fine for any length |
| Boolean | `BOOLEAN` | `INTEGER` or `SMALLINT` |
| Money | `NUMERIC(12,2)` or `BIGINT` (cents) | `MONEY` (locale-dependent), `FLOAT` (precision loss) |
| Timestamp | `TIMESTAMPTZ` (with timezone) | `TIMESTAMP` (without timezone) |
| Date only | `DATE` | `TEXT` with date string |
| JSON document | `JSONB` | `JSON` (no indexing, no operators) |
| Array | `TEXT[]`, `INTEGER[]` | Separate join table (unless you need containment queries) |
| IP address | `INET` | `TEXT` |
| Enum | `TEXT` with CHECK constraint | `ENUM` type (hard to modify) |

## Index Type Quick Reference

```
Need equality (=)?
  └── B-tree (default, best general purpose)

Need range (<, >, BETWEEN)?
  ├── Data naturally ordered → BRIN (tiny index!)
  └── General → B-tree

Need full-text search?
  ├── Speed → GIN on tsvector
  └── Ranking → GiST on tsvector

Need JSONB queries?
  ├── Containment (@>, ?) → GIN
  └── Specific key lookup → B-tree expression index

Need array queries?
  └── GIN

Need LIKE '%substring%'?
  └── GIN with pg_trgm extension

Need geometry/spatial?
  └── GiST with PostGIS
```

## Essential psql Commands

```sql
-- Connection
\conninfo            -- Show current connection info
\c dbname            -- Switch database
\l                   -- List all databases

-- Schema exploration
\dt                  -- List tables
\dt+                 -- List tables with size
\d tablename         -- Describe table (columns, indexes, constraints)
\di                  -- List indexes
\df                  -- List functions
\dn                  -- List schemas
\du                  -- List roles/users

-- Query helpers
\x                   -- Toggle expanded display
\timing              -- Toggle query timing
\e                   -- Edit last query in $EDITOR
\i filename.sql      -- Execute SQL file
\copy                -- Client-side COPY (no superuser needed)

-- Output
\o filename          -- Send output to file
\o                   -- Reset to stdout
```

## Performance Diagnostic Queries

### Top 10 Slowest Queries (pg_stat_statements)
```sql
SELECT
  calls,
  round(total_exec_time::numeric / calls, 2) AS avg_ms,
  round(total_exec_time::numeric, 2) AS total_ms,
  rows,
  query
FROM pg_stat_statements
ORDER BY total_exec_time DESC
LIMIT 10;
```

### Table Sizes
```sql
SELECT
  relname AS table,
  pg_size_pretty(pg_total_relation_size(relid)) AS total,
  pg_size_pretty(pg_relation_size(relid)) AS data,
  pg_size_pretty(pg_indexes_size(relid)) AS indexes
FROM pg_catalog.pg_statio_user_tables
ORDER BY pg_total_relation_size(relid) DESC;
```

### Cache Hit Ratio (Should be > 99%)
```sql
SELECT
  sum(heap_blks_hit) AS hits,
  sum(heap_blks_read) AS reads,
  round(
    sum(heap_blks_hit)::numeric /
    GREATEST(sum(heap_blks_hit) + sum(heap_blks_read), 1) * 100, 2
  ) AS hit_ratio
FROM pg_statio_user_tables;
```

### Index Usage (Find Unused Indexes)
```sql
SELECT
  relname AS table,
  indexrelname AS index,
  idx_scan AS scans,
  pg_size_pretty(pg_relation_size(indexrelid)) AS size
FROM pg_stat_user_indexes
WHERE idx_scan = 0
  AND indexrelname NOT LIKE '%_pkey'
ORDER BY pg_relation_size(indexrelid) DESC;
```

### Active Queries & Locks
```sql
-- Running queries
SELECT pid, now() - query_start AS duration, state, query
FROM pg_stat_activity
WHERE state != 'idle'
  AND pid != pg_backend_pid()
ORDER BY duration DESC;

-- Blocking locks
SELECT
  blocked.pid AS blocked_pid,
  blocked.query AS blocked_query,
  blocking.pid AS blocking_pid,
  blocking.query AS blocking_query
FROM pg_stat_activity blocked
JOIN pg_locks bl ON bl.pid = blocked.pid AND NOT bl.granted
JOIN pg_locks kl ON kl.locktype = bl.locktype
  AND kl.relation IS NOT DISTINCT FROM bl.relation
  AND kl.pid != bl.pid AND kl.granted
JOIN pg_stat_activity blocking ON blocking.pid = kl.pid;
```

### Dead Tuples (Need VACUUM?)
```sql
SELECT
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

## Configuration Quick Tune

### By Server RAM

| RAM | shared_buffers | effective_cache_size | work_mem | maintenance_work_mem |
|-----|---------------|---------------------|----------|---------------------|
| 1 GB | 256 MB | 768 MB | 4 MB | 64 MB |
| 4 GB | 1 GB | 3 GB | 16 MB | 256 MB |
| 8 GB | 2 GB | 6 GB | 32 MB | 512 MB |
| 16 GB | 4 GB | 12 GB | 64 MB | 1 GB |
| 32 GB | 8 GB | 24 GB | 128 MB | 2 GB |

### Must-Change Defaults (SSD)
```ini
random_page_cost = 1.1          # Default 4.0 assumes spinning disk!
effective_io_concurrency = 200  # Default 1 assumes spinning disk!
```

## Migration Safety Checklist

```
Before running a migration in production:

[ ] Tested on staging with production-sized data
[ ] CREATE INDEX uses CONCURRENTLY (no table lock)
[ ] ALTER TABLE ADD COLUMN has no DEFAULT on large tables (PG 11+ is safe)
[ ] No ALTER TABLE ... ALTER TYPE on large tables (rewrites entire table)
[ ] No DROP COLUMN in hot path (marks as dropped, doesn't free space)
[ ] Backup taken (pg_dump or WAL archiving verified)
[ ] Maintenance window scheduled if needed
[ ] Rollback migration written and tested
[ ] Application handles both old and new schema during deploy
```

## Common Gotchas

1. **`VARCHAR(n)` vs `TEXT`** — In PostgreSQL, there is NO performance difference. `TEXT` is preferred. `VARCHAR(n)` just adds a length check.

2. **`TIMESTAMP` vs `TIMESTAMPTZ`** — Always use `TIMESTAMPTZ`. Without timezone, PostgreSQL stores the literal value with no timezone context. You will have bugs.

3. **`SERIAL` vs `IDENTITY`** — `SERIAL` is legacy. Use `GENERATED ALWAYS AS IDENTITY` for auto-increment. It's SQL standard and prevents manual override mistakes.

4. **`IN (SELECT ...)` vs `EXISTS`** — For correlated subqueries, `EXISTS` is almost always faster. The planner can optimize it differently.

5. **`COUNT(*)` is slow** — On large tables, `COUNT(*)` does a full table scan. Use `pg_class.reltuples` for estimates, or maintain a counter table.

6. **Forgetting `CONCURRENTLY`** — `CREATE INDEX` locks the table for writes for the entire build. On a 100M row table, that's minutes of downtime. Always `CREATE INDEX CONCURRENTLY`.

7. **Too many connections** — Each PostgreSQL connection is a process (~10 MB). 100 connections = 1 GB of RAM just for processes. Use PgBouncer.

8. **Not analyzing after bulk operations** — After a large INSERT, UPDATE, or DELETE, run `ANALYZE tablename` to update planner statistics. Stale stats = bad query plans.
