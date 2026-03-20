---
name: postgres-operations
description: >
  PostgreSQL operations — backup and restore, replication setup,
  point-in-time recovery, monitoring queries, user management,
  schema migrations, and disaster recovery procedures.
  Triggers: "postgres backup", "postgres restore", "postgres replication",
  "postgres pitr", "postgres monitoring", "postgres user", "postgres migration",
  "postgres disaster recovery", "pg_dump", "pg_basebackup".
  NOT for: Query optimization or performance tuning (use postgres-performance).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# PostgreSQL Operations

## Backup and Restore

```bash
# Logical backup — pg_dump
# Full database backup (custom format, compressed)
pg_dump -h localhost -U postgres -d myapp \
  --format=custom --compress=9 \
  --file=myapp_$(date +%Y%m%d_%H%M%S).dump

# Schema-only backup
pg_dump -h localhost -U postgres -d myapp \
  --schema-only --file=schema.sql

# Data-only backup
pg_dump -h localhost -U postgres -d myapp \
  --data-only --file=data.sql

# Single table backup
pg_dump -h localhost -U postgres -d myapp \
  --table=users --format=custom \
  --file=users_backup.dump

# Parallel dump (faster for large databases)
pg_dump -h localhost -U postgres -d myapp \
  --format=directory --jobs=4 \
  --file=backup_dir/

# All databases
pg_dumpall -h localhost -U postgres > all_databases.sql

# Restore from custom format
pg_restore -h localhost -U postgres -d myapp \
  --clean --if-exists \
  --jobs=4 \
  myapp_20240115_120000.dump

# Restore specific table
pg_restore -h localhost -U postgres -d myapp \
  --table=users --data-only \
  myapp_backup.dump

# Restore from SQL file
psql -h localhost -U postgres -d myapp < backup.sql
```

```bash
# Physical backup — pg_basebackup
# Full binary backup (for PITR)
pg_basebackup -h localhost -U replicator \
  --pgdata=/var/lib/postgresql/backup \
  --format=tar --gzip \
  --wal-method=stream \
  --checkpoint=fast \
  --progress --verbose

# Automated backup script
#!/bin/bash
set -euo pipefail

BACKUP_DIR="/backups/postgres"
RETENTION_DAYS=30
DB_HOST="db.example.com"
DB_NAME="myapp"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="${BACKUP_DIR}/${DB_NAME}_${TIMESTAMP}.dump"

# Create backup
pg_dump -h "$DB_HOST" -U postgres -d "$DB_NAME" \
  --format=custom --compress=9 \
  --file="$BACKUP_FILE"

# Verify backup
pg_restore --list "$BACKUP_FILE" > /dev/null 2>&1
if [ $? -eq 0 ]; then
  echo "Backup verified: $BACKUP_FILE ($(du -h "$BACKUP_FILE" | cut -f1))"
else
  echo "ERROR: Backup verification failed!" >&2
  exit 1
fi

# Upload to S3
aws s3 cp "$BACKUP_FILE" "s3://backups-bucket/postgres/${DB_NAME}/"

# Cleanup old local backups
find "$BACKUP_DIR" -name "*.dump" -mtime +$RETENTION_DAYS -delete

echo "Backup complete: $BACKUP_FILE"
```

## Streaming Replication

```bash
# On PRIMARY — configure replication
# postgresql.conf
wal_level = replica
max_wal_senders = 5
wal_keep_size = 1GB
hot_standby = on

# pg_hba.conf — allow replication connections
# host    replication    replicator    10.0.0.0/8    scram-sha-256
```

```sql
-- On PRIMARY: create replication user
CREATE ROLE replicator WITH REPLICATION LOGIN PASSWORD 'secure_password';
```

```bash
# On REPLICA: initial setup
pg_basebackup -h primary-host -U replicator \
  --pgdata=/var/lib/postgresql/16/main \
  --wal-method=stream \
  --checkpoint=fast \
  --progress

# Create standby signal file
touch /var/lib/postgresql/16/main/standby.signal

# postgresql.conf on replica
primary_conninfo = 'host=primary-host port=5432 user=replicator password=secure_password'
hot_standby = on
```

```sql
-- Monitor replication status (on PRIMARY)
SELECT
  client_addr,
  state,
  sent_lsn,
  write_lsn,
  flush_lsn,
  replay_lsn,
  pg_wal_lsn_diff(sent_lsn, replay_lsn) as replay_lag_bytes,
  write_lag,
  flush_lag,
  replay_lag
FROM pg_stat_replication;

-- Monitor replication status (on REPLICA)
SELECT
  pg_is_in_recovery() as is_replica,
  pg_last_wal_receive_lsn() as received_lsn,
  pg_last_wal_replay_lsn() as replayed_lsn,
  pg_last_xact_replay_timestamp() as last_replayed_at,
  NOW() - pg_last_xact_replay_timestamp() as replay_delay;
```

## Point-in-Time Recovery (PITR)

```bash
# Prerequisites: continuous WAL archiving
# postgresql.conf
archive_mode = on
archive_command = 'aws s3 cp %p s3://wal-archive/myapp/%f'
# OR: archive_command = 'cp %p /archive/%f'

# Restore to a specific point in time
# 1. Stop PostgreSQL
systemctl stop postgresql

# 2. Move existing data directory
mv /var/lib/postgresql/16/main /var/lib/postgresql/16/main.old

# 3. Restore base backup
pg_basebackup -h backup-host -U replicator \
  --pgdata=/var/lib/postgresql/16/main

# 4. Create recovery configuration
cat > /var/lib/postgresql/16/main/postgresql.auto.conf << 'EOF'
restore_command = 'aws s3 cp s3://wal-archive/myapp/%f %p'
recovery_target_time = '2024-06-15 14:30:00 UTC'
recovery_target_action = 'promote'
EOF

# 5. Create recovery signal
touch /var/lib/postgresql/16/main/recovery.signal

# 6. Start PostgreSQL — it will replay WAL up to the target time
systemctl start postgresql

# 7. Verify recovery
psql -c "SELECT pg_is_in_recovery();"
# Should return false after recovery completes
```

## Monitoring Queries

```sql
-- Active queries (find long-running or blocked queries)
SELECT
  pid,
  usename,
  state,
  wait_event_type,
  wait_event,
  query_start,
  NOW() - query_start as duration,
  LEFT(query, 100) as query_preview
FROM pg_stat_activity
WHERE state != 'idle'
  AND pid != pg_backend_pid()
ORDER BY query_start;

-- Kill a specific query
SELECT pg_cancel_backend(12345);    -- Graceful cancel (SIGINT)
SELECT pg_terminate_backend(12345); -- Force kill (SIGTERM)

-- Lock monitoring (find blocked queries)
SELECT
  blocked.pid as blocked_pid,
  blocked.usename as blocked_user,
  LEFT(blocked.query, 60) as blocked_query,
  blocking.pid as blocking_pid,
  blocking.usename as blocking_user,
  LEFT(blocking.query, 60) as blocking_query,
  NOW() - blocked.query_start as wait_duration
FROM pg_stat_activity blocked
JOIN pg_locks bl ON bl.pid = blocked.pid
JOIN pg_locks l ON l.locktype = bl.locktype
  AND l.database IS NOT DISTINCT FROM bl.database
  AND l.relation IS NOT DISTINCT FROM bl.relation
  AND l.page IS NOT DISTINCT FROM bl.page
  AND l.tuple IS NOT DISTINCT FROM bl.tuple
  AND l.pid != bl.pid
JOIN pg_stat_activity blocking ON l.pid = blocking.pid
WHERE NOT bl.granted;

-- Connection usage
SELECT
  count(*) as total,
  count(*) FILTER (WHERE state = 'active') as active,
  count(*) FILTER (WHERE state = 'idle') as idle,
  count(*) FILTER (WHERE state = 'idle in transaction') as idle_in_txn,
  count(*) FILTER (WHERE wait_event_type = 'Lock') as waiting,
  current_setting('max_connections') as max_allowed
FROM pg_stat_activity;

-- Database size
SELECT
  datname,
  pg_size_pretty(pg_database_size(datname)) as size
FROM pg_database
ORDER BY pg_database_size(datname) DESC;

-- Table sizes with indexes
SELECT
  schemaname || '.' || tablename as table,
  pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as total,
  pg_size_pretty(pg_relation_size(schemaname||'.'||tablename)) as data,
  pg_size_pretty(pg_indexes_size((schemaname||'.'||tablename)::regclass)) as indexes,
  pg_size_pretty(
    pg_total_relation_size(schemaname||'.'||tablename) -
    pg_relation_size(schemaname||'.'||tablename) -
    pg_indexes_size((schemaname||'.'||tablename)::regclass)
  ) as toast
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC
LIMIT 20;

-- Slow queries from pg_stat_statements
SELECT
  queryid,
  calls,
  ROUND(total_exec_time::numeric / 1000, 2) as total_seconds,
  ROUND(mean_exec_time::numeric, 2) as avg_ms,
  ROUND(max_exec_time::numeric, 2) as max_ms,
  rows,
  LEFT(query, 80) as query_preview
FROM pg_stat_statements
ORDER BY total_exec_time DESC
LIMIT 20;
```

## User Management

```sql
-- Create application user with minimal privileges
CREATE ROLE app_user WITH LOGIN PASSWORD 'secure_password';
GRANT CONNECT ON DATABASE myapp TO app_user;
GRANT USAGE ON SCHEMA public TO app_user;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO app_user;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO app_user;

-- Apply to future tables too
ALTER DEFAULT PRIVILEGES IN SCHEMA public
  GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO app_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public
  GRANT USAGE, SELECT ON SEQUENCES TO app_user;

-- Read-only user for analytics/reporting
CREATE ROLE readonly_user WITH LOGIN PASSWORD 'readonly_pass';
GRANT CONNECT ON DATABASE myapp TO readonly_user;
GRANT USAGE ON SCHEMA public TO readonly_user;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO readonly_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public
  GRANT SELECT ON TABLES TO readonly_user;

-- Migration user (needs DDL privileges)
CREATE ROLE migration_user WITH LOGIN PASSWORD 'migration_pass';
GRANT ALL PRIVILEGES ON DATABASE myapp TO migration_user;
GRANT ALL PRIVILEGES ON SCHEMA public TO migration_user;

-- Revoke dangerous permissions
REVOKE CREATE ON SCHEMA public FROM PUBLIC;
REVOKE ALL ON DATABASE myapp FROM PUBLIC;

-- Password rotation
ALTER ROLE app_user WITH PASSWORD 'new_secure_password';

-- List all roles and permissions
SELECT
  r.rolname,
  r.rolsuper,
  r.rolcreaterole,
  r.rolcreatedb,
  r.rolcanlogin,
  r.rolreplication,
  ARRAY(SELECT b.rolname FROM pg_auth_members m
    JOIN pg_roles b ON m.roleid = b.oid
    WHERE m.member = r.oid) as member_of
FROM pg_roles r
WHERE r.rolname NOT LIKE 'pg_%'
ORDER BY r.rolname;
```

## Schema Migrations

```sql
-- Migration table pattern
CREATE TABLE IF NOT EXISTS schema_migrations (
  version     TEXT PRIMARY KEY,
  applied_at  TIMESTAMPTZ DEFAULT NOW(),
  description TEXT
);

-- Safe migration patterns

-- Add column (always nullable first, never NOT NULL with default on large table)
ALTER TABLE users ADD COLUMN phone TEXT;

-- Add NOT NULL constraint safely (PostgreSQL 12+)
-- Step 1: Add constraint as NOT VALID (instant, no table scan)
ALTER TABLE users ADD CONSTRAINT users_phone_not_null
  CHECK (phone IS NOT NULL) NOT VALID;
-- Step 2: Backfill data
UPDATE users SET phone = '' WHERE phone IS NULL;
-- Step 3: Validate (scans table but doesn't lock writes)
ALTER TABLE users VALIDATE CONSTRAINT users_phone_not_null;

-- Rename column safely (expand-contract)
-- Step 1: Add new column
ALTER TABLE users ADD COLUMN display_name TEXT;
-- Step 2: Backfill
UPDATE users SET display_name = name;
-- Step 3: Deploy app code that reads/writes both columns
-- Step 4: Drop old column (after all app instances updated)
ALTER TABLE users DROP COLUMN name;

-- Add index without locking
CREATE INDEX CONCURRENTLY idx_users_phone ON users (phone);

-- Drop index safely
DROP INDEX CONCURRENTLY IF EXISTS idx_old_index;
```

## Disaster Recovery Checklist

```bash
#!/bin/bash
# scripts/dr-runbook.sh

echo "=== PostgreSQL Disaster Recovery Runbook ==="

echo ""
echo "1. ASSESS THE SITUATION"
echo "   - Is the database responding? psql -c 'SELECT 1;'"
echo "   - Is it a data loss or availability issue?"
echo "   - When did the problem start?"

echo ""
echo "2. IF DATABASE IS DOWN"
echo "   a. Check PostgreSQL logs: journalctl -u postgresql --since '1 hour ago'"
echo "   b. Check disk space: df -h /var/lib/postgresql"
echo "   c. Check connections: ss -tlnp | grep 5432"
echo "   d. Try restart: systemctl restart postgresql"
echo "   e. If won't start: check pg_log for FATAL/PANIC entries"

echo ""
echo "3. IF DATA CORRUPTION"
echo "   a. Stop the server immediately"
echo "   b. DO NOT run VACUUM or any writes"
echo "   c. Copy the data directory as-is for forensics"
echo "   d. Restore from latest backup + WAL replay"
echo "   e. See PITR steps above"

echo ""
echo "4. IF ACCIDENTAL DATA DELETION"
echo "   a. Identify the exact time of deletion"
echo "   b. Check if table has soft-delete (deleted_at column)"
echo "   c. If yes: UPDATE table SET deleted_at = NULL WHERE ..."
echo "   d. If no: PITR to just before the deletion"
echo "   e. Selective restore: dump specific table from recovered DB"

echo ""
echo "5. PROMOTE REPLICA (if primary is unrecoverable)"
echo "   a. On replica: SELECT pg_promote();"
echo "   b. Update connection strings in application config"
echo "   c. Verify: SELECT pg_is_in_recovery(); -- should be false"
echo "   d. Set up new replica from promoted primary"

echo ""
echo "6. POST-RECOVERY"
echo "   a. Verify data integrity: run application health checks"
echo "   b. Compare row counts against last known good state"
echo "   c. Restart replication to all replicas"
echo "   d. Take a fresh backup immediately"
echo "   e. Write incident report: timeline, root cause, prevention"
```

## Gotchas

1. **`pg_dump` doesn't backup roles or tablespaces** — `pg_dump` only backs up a single database's schema and data. Global objects (roles, users, tablespaces) require `pg_dumpall --globals-only`. A restore from `pg_dump` alone fails if the owning roles don't exist on the target server.

2. **`DROP DATABASE` can't be undone** — There's no recycle bin. `DROP DATABASE myapp` immediately removes all data files. Always verify you're connected to the right server before running DDL. Use `SET search_path` and `\conninfo` to confirm your context.

3. **Long transactions block autovacuum** — A transaction open for hours prevents autovacuum from cleaning any rows visible to that transaction. Dead tuples accumulate, table bloats, performance degrades. Monitor `idle in transaction` connections and set `idle_in_transaction_session_timeout` to kill stale transactions.

4. **`pg_basebackup` requires WAL retention** — If WAL segments are recycled before the backup finishes, the backup is invalid. Set `wal_keep_size` large enough to cover the backup duration, or use replication slots (but monitor slot lag — unconsumed slots prevent WAL cleanup and fill the disk).

5. **Replica promotion is one-way** — `SELECT pg_promote()` permanently turns a replica into a standalone primary. You can't demote it back. After promotion, you need to rebuild replicas from scratch using `pg_basebackup` from the new primary. Plan your promotion carefully.

6. **`ALTER TABLE ... ADD COLUMN ... DEFAULT` rewrites the entire table on PostgreSQL < 11** — On PostgreSQL 10 and earlier, adding a column with a default value rewrites every row (exclusive lock for hours on large tables). PostgreSQL 11+ stores the default in the catalog without rewriting. Always check your PostgreSQL version before running schema changes.
