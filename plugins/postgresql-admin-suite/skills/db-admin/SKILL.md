---
name: db-admin
description: >
  Expert PostgreSQL database administration suite. Activate when the user needs
  help with PostgreSQL performance tuning, EXPLAIN ANALYZE interpretation, index
  selection, vacuum optimization, schema design, zero-downtime migrations,
  streaming or logical replication, connection pooling, backup and recovery,
  PITR, pgBackRest, or general PostgreSQL DBA tasks.
version: 1.0.0
---

# PostgreSQL & Database Administration Suite

## Metadata
- Name: db-admin
- Description: Production-grade PostgreSQL DBA — performance, schema, replication, and backup/recovery
- Version: 1.0.0

## Trigger

Activate when the user asks about:
- PostgreSQL performance tuning or slow query analysis
- EXPLAIN ANALYZE output interpretation
- Index selection (B-tree, GIN, GiST, BRIN, hash)
- Autovacuum configuration or VACUUM operations
- Table partitioning (range, list, hash)
- Connection pooling (pgBouncer, PgCat)
- `shared_buffers`, `work_mem`, `effective_cache_size`, or other postgresql.conf settings
- Schema design, normalization, or database constraints
- Database migrations (zero-downtime, expand-contract)
- PostgreSQL extensions (pg_trgm, PostGIS, uuid-ossp, hstore, ltree)
- Triggers, functions, or PL/pgSQL
- Multi-tenant database architecture
- Streaming replication or logical replication
- Read replicas, failover, or high availability
- Change data capture (CDC) or logical decoding
- pg_dump, pg_basebackup, or pgBackRest
- Point-in-Time Recovery (PITR) or WAL archiving
- Backup scheduling, retention, or verification
- Disaster recovery planning (RTO/RPO)
- PostgreSQL security, row-level security, or role management
- pg_hba.conf configuration or SSL/TLS setup
- pg_stat_statements, pg_stat_activity, or monitoring

## Agents

### postgres-dba
**When to use:** The user needs to optimize query performance, tune PostgreSQL configuration, analyze execution plans, choose indexes, configure autovacuum, set up partitioning, or diagnose performance issues.

**Capabilities:**
- Read and interpret EXPLAIN ANALYZE output (costs, rows, buffers, timing)
- Recommend index types based on query patterns and data characteristics
- Tune postgresql.conf for OLTP, OLAP, and mixed workloads
- Configure autovacuum per-table for high-write scenarios
- Design partitioning strategies with partition pruning verification
- Set up pg_stat_statements and build monitoring dashboards
- Diagnose lock contention, connection saturation, and bloat
- Optimize bulk operations with COPY, batch inserts, and parallel queries

### schema-architect
**When to use:** The user needs to design database schemas, plan migrations, work with triggers and functions, configure extensions, or build multi-tenant architectures.

**Capabilities:**
- Design normalized schemas through 5NF with proper constraints
- Plan zero-downtime migrations using expand-contract patterns
- Create triggers, generated columns, domain types, and custom functions
- Configure and use PostgreSQL extensions
- Design multi-tenant schemas (shared tables, separate schemas, row-level isolation)
- Build JSONB hybrid schemas combining relational and document patterns
- Generate migration files for Prisma, Drizzle, Knex, Alembic, golang-migrate

### replication-expert
**When to use:** The user needs to set up replication, configure connection pooling, plan failover strategies, or implement change data capture.

**Capabilities:**
- Configure streaming replication (sync, async, quorum)
- Set up logical replication with publication/subscription
- Deploy and tune pgBouncer (transaction, session, statement pooling)
- Deploy PgCat for advanced connection pooling with sharding
- Plan and execute failover/switchover procedures
- Monitor replication lag and manage replication slots
- Design cascading replication topologies
- Implement CDC with logical decoding output plugins

### backup-recovery-expert
**When to use:** The user needs to set up backups, plan disaster recovery, implement PITR, or configure pgBackRest.

**Capabilities:**
- Configure pg_dump/pg_dumpall with optimal flags and scheduling
- Set up pg_basebackup for physical backups
- Implement PITR with continuous WAL archiving
- Deploy pgBackRest with stanza configuration, retention, and S3/GCS storage
- Design backup strategies with RTO/RPO targets
- Create and test disaster recovery runbooks
- Verify backup integrity with automated restore testing
- Set up backup monitoring and alerting

## References

### query-optimization
Deep-dive into EXPLAIN ANALYZE output reading, PostgreSQL query planner internals, join algorithm selection (nested loop, hash join, merge join), common query anti-patterns with fixes, CTE materialization control, and parallel query tuning.

### index-strategies
Comprehensive index type guide — B-tree internals and use cases, GIN for full-text and JSONB, GiST for geometric and range types, BRIN for time-series and append-only, hash for equality-only lookups. Includes partial indexes, expression indexes, covering indexes, and index-only scans.

### security-hardening
PostgreSQL security hardening — row-level security policies, pg_hba.conf patterns, SSL/TLS configuration, role hierarchy design, password policies with SCRAM-SHA-256, audit logging with pgAudit, and GRANT/REVOKE best practices.

## Workflow

1. **Assess** — Gather PostgreSQL version, current configuration, workload characteristics, and pain points
2. **Diagnose** — Run diagnostic queries (pg_stat_statements, pg_stat_activity, pg_stat_user_tables) to identify bottlenecks
3. **Recommend** — Provide specific, actionable recommendations with expected impact
4. **Implement** — Generate production-ready SQL, configuration changes, and migration scripts
5. **Verify** — Provide verification queries and monitoring setup to confirm improvements
6. **Document** — Generate runbooks and operational documentation for the changes

## Languages

All agents provide production-ready code and configurations for:
- **SQL** and **PL/pgSQL** — native PostgreSQL
- **Node.js** — node-postgres (pg), Prisma, Drizzle ORM
- **Python** — psycopg2/psycopg3, SQLAlchemy, Alembic
- **Go** — pgx, sqlc, golang-migrate

## Quality Standards

- Every recommendation includes version compatibility notes (PG 14-17)
- All configuration changes include rollback procedures
- Performance recommendations include before/after verification queries
- Security configurations follow CIS PostgreSQL Benchmark guidelines
- Backup procedures include automated restore verification
- Replication setups include monitoring and alerting configuration
- Migration scripts are tested for zero-downtime compatibility
