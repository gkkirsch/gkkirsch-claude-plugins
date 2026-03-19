---
name: database-design-suite
description: >
  Database Design & Optimization Suite — AI-powered toolkit for schema design, query optimization,
  migration planning, and performance tuning. Covers PostgreSQL, MySQL, MongoDB, Redis, and SQLite.
  Triggers: "database design", "schema design", "design tables", "data model", "entity relationship",
  "normalize", "denormalize", "query optimization", "optimize query", "slow query", "explain analyze",
  "index strategy", "add index", "query plan", "migration plan", "zero downtime migration",
  "schema migration", "database migration", "migrate database", "rename column", "change column type",
  "performance tuning", "tune database", "connection pooling", "pgbouncer", "vacuum", "database slow",
  "database performance", "db design", "db optimize", "db migrate", "db perf", "db tune".
  Dispatches the appropriate specialist agent: schema-architect, query-optimizer, migration-planner,
  or performance-tuner.
  NOT for: Application-level caching logic, ORM configuration, API design, frontend data fetching,
  or non-database storage systems.
version: 1.0.0
argument-hint: "<schema|query|migrate|perf> [target]"
user-invocable: true
allowed-tools: Read, Grep, Glob, Bash
model: sonnet
---

# Database Design & Optimization Suite

Production-grade database design and optimization agents for Claude Code. Four specialist agents that
handle schema architecture, query optimization, migration planning, and performance tuning — the
database engineering work that every backend project needs.

## Available Agents

### Schema Architect (`schema-architect`)
Designs normalized and denormalized database schemas. Entity-relationship modeling, table design with
proper keys and constraints, polymorphic associations, multi-tenant patterns, temporal data, audit
trails, hierarchical data structures, JSON/JSONB columns, and schema versioning.

**Invoke**: Dispatch via Task tool with `subagent_type: "schema-architect"`.

**Example prompts**:
- "Design a database schema for my e-commerce app"
- "Review my Prisma schema for normalization issues"
- "Design a multi-tenant schema with row-level security"
- "Model a hierarchical category tree in PostgreSQL"

### Query Optimizer (`query-optimizer`)
Analyzes and optimizes SQL queries. EXPLAIN/EXPLAIN ANALYZE interpretation, index selection,
query plan reading, join optimization, subquery refactoring, window functions, CTE performance,
partitioning, materialized views, and bulk operation strategies.

**Invoke**: Dispatch via Task tool with `subagent_type: "query-optimizer"`.

**Example prompts**:
- "This dashboard query takes 8 seconds — optimize it"
- "Help me understand this EXPLAIN ANALYZE output"
- "What indexes should I add for my most common queries?"
- "Rewrite this subquery-heavy report for better performance"

### Migration Planner (`migration-planner`)
Plans safe, zero-downtime database migrations. Backward-compatible schema changes, expand-contract
pattern, online DDL, data migration strategies, rollback plans, and works with Prisma, Knex,
Sequelize, Drizzle, Alembic, Flyway, Liquibase, Django, and Rails migration tools.

**Invoke**: Dispatch via Task tool with `subagent_type: "migration-planner"`.

**Example prompts**:
- "Plan a zero-downtime migration to rename the users.name column"
- "Generate Prisma migrations to split the address fields into a new table"
- "Plan a migration from MySQL to PostgreSQL"
- "How do I safely change a column type from integer to decimal?"

### Performance Tuner (`performance-tuner`)
Optimizes database performance at the infrastructure level. Connection pooling (PgBouncer, ProxySQL),
memory configuration, vacuum/analyze tuning, table bloat detection, lock contention analysis,
slow query logging, replication, and capacity planning.

**Invoke**: Dispatch via Task tool with `subagent_type: "performance-tuner"`.

**Example prompts**:
- "My PostgreSQL database is running slow — diagnose and fix"
- "Set up PgBouncer for connection pooling"
- "Configure autovacuum for my high-write tables"
- "Plan read replica setup for my PostgreSQL database"

## Quick Start: /db-design

Use the `/db-design` command for guided database work:

```
/db-design                          # Auto-detect and suggest improvements
/db-design schema                   # Design or review database schema
/db-design schema users orders      # Design schema for specific tables
/db-design query                    # Find and optimize slow queries
/db-design query "SELECT ..."       # Optimize specific query
/db-design migrate                  # Plan database migration
/db-design perf                     # Performance tuning
```

## Agent Selection Guide

| Need | Agent | Prompt |
|------|-------|--------|
| Design new tables | schema-architect | "Design a schema for..." |
| Review existing schema | schema-architect | "Review my schema for issues" |
| Model relationships | schema-architect | "How should I model X?" |
| Multi-tenant design | schema-architect | "Design tenant isolation" |
| Fix slow query | query-optimizer | "Optimize this query" |
| Read EXPLAIN output | query-optimizer | "Explain this query plan" |
| Add indexes | query-optimizer | "What indexes do I need?" |
| Rewrite complex SQL | query-optimizer | "Refactor this query" |
| Rename/move column | migration-planner | "Rename column safely" |
| Change column type | migration-planner | "Change type from X to Y" |
| Split/merge tables | migration-planner | "Split addresses into own table" |
| Database migration | migration-planner | "Migrate from MySQL to PG" |
| Slow database | performance-tuner | "Diagnose slow database" |
| Connection issues | performance-tuner | "Set up connection pooling" |
| High disk usage | performance-tuner | "Detect and fix table bloat" |
| Lock problems | performance-tuner | "Analyze lock contention" |

## Reference Materials

This skill includes comprehensive reference documents in `references/`:

- **postgresql-deep-dive.md** — MVCC internals, WAL mechanics, TOAST, advanced indexes (GIN, GiST, BRIN), full-text search, advisory locks, LISTEN/NOTIFY, FDW, RLS, pgvector, PostGIS, logical replication
- **indexing-strategies.md** — B-tree internals, ESR rule, covering indexes, partial indexes, GIN/GiST/BRIN details, MongoDB indexes, Redis data structures as indexes, MySQL InnoDB clustered indexes, cardinality and selectivity analysis
- **data-modeling-patterns.md** — Document vs relational decisions, event sourcing, CQRS, time-series data, graph relationships, polymorphic data, multi-tenant isolation, Redis caching patterns, MongoDB schema patterns (bucket, computed, schema versioning, subset, extended reference)

Agents automatically consult these references when working. You can also read them directly.

## How It Works

1. You describe what you need (e.g., "design a schema for my app")
2. The SKILL.md routes to the appropriate agent
3. The agent reads your code, detects your database stack, and understands your data model
4. Schema designs, optimized queries, migration plans, or tuning recommendations are generated
5. The agent provides results and next steps

All generated artifacts follow database engineering best practices:
- Schema: Proper normalization, constraints, keys, and indexes
- Queries: EXPLAIN-verified optimizations with before/after comparison
- Migrations: Zero-downtime with tested rollback plans
- Performance: Measurable improvements with monitoring setup
