---
name: db-design
description: >
  Database design command — analyzes your codebase, detects database setup, and routes to the appropriate
  specialist agent for schema design, query optimization, migration planning, or performance tuning.
  Triggers: "/db-design", "design database", "optimize queries", "plan migration", "tune database".
user-invocable: true
argument-hint: "<schema|query|migrate|perf> [table-or-query] [--db postgres|mysql|sqlite|mongo|redis]"
allowed-tools: Read, Write, Edit, Bash, Glob, Grep
model: sonnet
---

# /db-design Command

One-command database design, optimization, and migration planning. Analyzes your codebase, detects your
database stack, and routes to the right specialist agent.

## Usage

```
/db-design                          # Auto-detect and suggest improvements
/db-design schema                   # Design or review database schema
/db-design schema users orders      # Design schema for specific tables
/db-design query                    # Find and optimize slow queries
/db-design query "SELECT * FROM..." # Optimize a specific query
/db-design migrate                  # Plan database migration
/db-design migrate --rename col     # Plan specific migration type
/db-design perf                     # Analyze and tune database performance
/db-design perf --connections       # Focus on connection pooling
```

## Subcommands

| Subcommand | Agent | Description |
|------------|-------|-------------|
| `schema` | schema-architect | Design tables, relationships, indexes, constraints |
| `query` | query-optimizer | Analyze EXPLAIN plans, optimize slow queries, add indexes |
| `migrate` | migration-planner | Plan zero-downtime migrations with rollback |
| `perf` | performance-tuner | Tune configuration, connection pools, vacuum, monitoring |

## Procedure

### Step 1: Detect Database Stack

Read project configuration files to identify the database setup:

1. **Read** `package.json`, `requirements.txt`, `go.mod`, `Cargo.toml`, `Gemfile`, `pom.xml`, `composer.json` — detect language and framework
2. **Glob** for ORM/database files:
   - Prisma: `**/prisma/schema.prisma`
   - Drizzle: `**/drizzle.config.*`, `**/drizzle/**/*.ts`
   - Knex: `**/knexfile.*`, `**/migrations/**`
   - Sequelize: `**/config/config.json`, `**/models/**`
   - TypeORM: `**/ormconfig.*`, `**/typeorm.config.*`
   - Django: `**/models.py`, `**/settings.py`
   - Rails: `**/db/schema.rb`, `**/db/migrate/**`
   - Alembic: `**/alembic.ini`, `**/alembic/versions/**`
   - Mongoose: `**/models/**`
3. **Grep** for database driver/connection:
   - `DATABASE_URL`, `MONGODB_URI`, `REDIS_URL`
   - `pg`, `postgres`, `mysql`, `sqlite`, `mongodb`, `redis`
4. **Check** for existing migrations, seeds, schema files

Report findings:
```
Detected:
- Database: PostgreSQL 16
- ORM: Prisma 5.x
- Tables: 24
- Migrations: 18
- Indexes: 42
- Estimated data: ~500K rows largest table
```

### Step 2: Route to Agent

Based on the subcommand, dispatch the appropriate agent:

#### `schema` — Schema Architect

```
Task tool:
  subagent_type: "schema-architect"
  mode: "bypassPermissions"
  prompt: |
    Design/review the database schema for this project.
    Database: [detected database]
    ORM: [detected ORM]
    Existing tables: [discovered tables]
    Target tables: [specific tables if provided, or all]
    Generate:
    - Entity-relationship model
    - Table definitions with proper types, constraints, and keys
    - Index recommendations
    - Migration files in the project's ORM format
```

#### `query` — Query Optimizer

```
Task tool:
  subagent_type: "query-optimizer"
  mode: "bypassPermissions"
  prompt: |
    Analyze and optimize database queries in this project.
    Database: [detected database]
    ORM: [detected ORM]
    Target query: [specific query if provided, or find slow queries]
    Provide:
    - EXPLAIN ANALYZE interpretation
    - Query rewrites for better performance
    - Index recommendations
    - Before/after comparison
```

#### `migrate` — Migration Planner

```
Task tool:
  subagent_type: "migration-planner"
  mode: "bypassPermissions"
  prompt: |
    Plan a database migration for this project.
    Database: [detected database]
    ORM/Migration tool: [detected tool]
    Change requested: [specific change or general review]
    Provide:
    - Migration files in the project's format
    - Zero-downtime strategy
    - Rollback plan
    - Data migration steps
    - Testing checklist
```

#### `perf` — Performance Tuner

```
Task tool:
  subagent_type: "performance-tuner"
  mode: "bypassPermissions"
  prompt: |
    Analyze and optimize database performance for this project.
    Database: [detected database]
    Focus area: [connections, memory, queries, vacuum, or general]
    Provide:
    - Current configuration assessment
    - Bottleneck identification
    - Specific tuning recommendations
    - Connection pooling configuration
    - Monitoring setup
```

### Step 3: Results Summary

After the agent completes, present a summary of:

**For schema:**
```
Schema Design Results:
- Tables designed: 8
- Relationships: 12 (5 one-to-many, 2 many-to-many, 5 one-to-one)
- Indexes recommended: 15
- Migration file: prisma/migrations/20240115_design/migration.sql
- Key decisions: UUID PKs, soft deletes, tenant isolation via RLS
```

**For query optimization:**
```
Query Optimization Results:
- Queries analyzed: 6
- Queries improved: 4
- Indexes added: 3
- Estimated speedup: 12x for dashboard query, 5x for search
- Files modified: src/queries/dashboard.ts, prisma/schema.prisma
```

**For migration:**
```
Migration Plan:
- Changes: Rename column, add index, split table
- Strategy: Expand-contract (3 deployments)
- Estimated duration: ~2 minutes for backfill
- Risk: LOW (all changes backward compatible)
- Files created: migrations/001_expand.sql, migrations/002_contract.sql
```

**For performance:**
```
Performance Tuning Results:
- Bottleneck: Connection exhaustion (100 max, 95 active)
- Changes: PgBouncer config, shared_buffers increase, new indexes
- Expected improvement: 3x connection capacity, 40% query time reduction
- Files created: pgbouncer.ini, postgresql.conf.recommended
```

## Error Recovery

| Error | Cause | Fix |
|-------|-------|-----|
| No database detected | Non-standard setup | Specify --db flag manually |
| Can't connect to database | No DATABASE_URL | Set environment variable |
| Unknown ORM | Custom database layer | Use raw SQL mode |
| Migration tool not found | Missing dependency | Install the ORM CLI tool |

## Notes

- Always review generated migrations before applying to production
- Query optimization recommendations should be tested with production-like data
- Performance tuning changes may require database restart
- Schema designs follow the project's existing conventions when detected
