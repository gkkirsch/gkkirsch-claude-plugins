# /db-admin Command

Expert PostgreSQL database administration — performance tuning, schema architecture, replication, and backup/recovery with production-proven patterns.

## Usage

```
/db-admin [subcommand] [options]
```

## Subcommands

### `tune`

```
/db-admin tune
```

Activates the **postgres-dba** agent to help you:
- Analyze slow queries with EXPLAIN ANALYZE and read execution plans
- Choose optimal index types (B-tree, GIN, GiST, BRIN, hash) for your workload
- Configure autovacuum for high-throughput tables
- Tune `shared_buffers`, `work_mem`, `effective_cache_size`, and checkpoint settings
- Set up `pg_stat_statements` and query performance monitoring
- Design table partitioning strategies (range, list, hash)
- Diagnose lock contention and connection saturation
- Optimize bulk operations (COPY, batch inserts, parallel queries)

### `schema`

```
/db-admin schema
```

Activates the **schema-architect** agent to help you:
- Design normalized schemas (1NF through 5NF) with proper constraints
- Plan zero-downtime migrations using expand-contract patterns
- Create and manage triggers, generated columns, and domain types
- Configure PostgreSQL extensions (pg_trgm, PostGIS, uuid-ossp, hstore, ltree)
- Design multi-tenant schemas (shared tables vs separate schemas)
- Implement row versioning and soft-delete patterns
- Build JSONB hybrid schemas combining relational and document models
- Generate migration files for popular frameworks (Prisma, Drizzle, Knex, Alembic, golang-migrate)

### `replicate`

```
/db-admin replicate
```

Activates the **replication-expert** agent to help you:
- Set up streaming replication with synchronous and asynchronous modes
- Configure logical replication for selective table sync
- Deploy pgBouncer or PgCat for connection pooling
- Design read-replica routing for horizontal read scaling
- Plan failover and switchover procedures
- Monitor replication lag and slot status
- Configure cascading replication topologies
- Set up change data capture (CDC) with logical decoding

### `backup`

```
/db-admin backup
```

Activates the **backup-recovery-expert** agent to help you:
- Configure `pg_dump` and `pg_dumpall` for logical backups
- Set up `pg_basebackup` for physical base backups
- Implement Point-in-Time Recovery (PITR) with WAL archiving
- Deploy pgBackRest for enterprise backup management
- Design backup retention and rotation policies
- Test and verify backup integrity
- Plan disaster recovery runbooks with RTO/RPO targets
- Automate backup scheduling with cron or systemd timers

### `full`

```
/db-admin full
```

Activates all agents for a comprehensive database administration review covering performance, schema, replication, and backup/recovery.

### `audit`

```
/db-admin audit
```

Runs a full PostgreSQL health check:
- Configuration review (`postgresql.conf`, `pg_hba.conf`)
- Performance baseline with `pg_stat_statements`
- Index usage analysis and bloat detection
- Vacuum and autovacuum health
- Replication status and lag
- Backup verification
- Security posture (roles, RLS, SSL, password policies)

## Examples

```
# Optimize a slow query
/db-admin tune

# Design a multi-tenant schema
/db-admin schema

# Set up streaming replication with pgBouncer
/db-admin replicate

# Configure pgBackRest with PITR
/db-admin backup

# Full database health audit
/db-admin audit
```

## Reference Files

The suite includes detailed reference documents:
- **query-optimization.md** — EXPLAIN ANALYZE output reading, join strategies, query planner internals, common anti-patterns with fixes
- **index-strategies.md** — B-tree, GIN, GiST, BRIN, hash index deep-dive with selection matrices and real-world examples
- **security-hardening.md** — Row-level security, SSL/TLS, role management, pg_hba.conf patterns, audit logging

## Supported Environments

All agents provide production-ready configurations and scripts for:
- **PostgreSQL 14, 15, 16, 17** (with version-specific features noted)
- **Linux** (Ubuntu/Debian, RHEL/Rocky, Amazon Linux)
- **Docker** and **Kubernetes** (Helm charts, operators)
- **Cloud** (AWS RDS/Aurora, GCP Cloud SQL, Azure Database for PostgreSQL)
- **Languages** — SQL, PL/pgSQL, Node.js (node-postgres, Prisma), Python (psycopg, SQLAlchemy, Alembic), Go (pgx, sqlc)

## What This Suite Covers

```
PostgreSQL Administration
├── Performance Tuning
│   ├── EXPLAIN ANALYZE & Query Plans
│   ├── Index Selection & Management
│   ├── Autovacuum & VACUUM
│   ├── Memory & Checkpoint Tuning
│   ├── Partitioning Strategies
│   ├── Connection Pooling
│   └── Monitoring & Alerting
├── Schema Architecture
│   ├── Normalization & Constraints
│   ├── Zero-Downtime Migrations
│   ├── Triggers & Functions
│   ├── Extensions & Custom Types
│   ├── Multi-Tenant Design
│   └── JSONB Hybrid Schemas
├── Replication & High Availability
│   ├── Streaming Replication
│   ├── Logical Replication & CDC
│   ├── pgBouncer & PgCat
│   ├── Failover & Switchover
│   └── Read-Replica Routing
└── Backup & Recovery
    ├── pg_dump & pg_dumpall
    ├── pg_basebackup
    ├── PITR & WAL Archiving
    ├── pgBackRest
    ├── Disaster Recovery Planning
    └── Backup Verification
```
