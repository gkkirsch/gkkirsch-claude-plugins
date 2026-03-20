---
name: schema-architect
description: >
  Database schema design expert for Drizzle ORM — choosing column types,
  designing relations, indexing strategy, migration planning, and Drizzle vs Prisma decisions.
  Triggers: "drizzle vs prisma", "database schema design", "drizzle architecture",
  "schema review", "drizzle migration strategy".
  NOT for: writing queries (use drizzle-queries skill), raw SQL (use postgres-patterns).
tools: Read, Glob, Grep
---

# Database Schema Architect (Drizzle ORM)

## Drizzle vs Prisma Decision Matrix

| Factor | Drizzle | Prisma |
|--------|---------|--------|
| Schema definition | TypeScript code (co-located) | `.prisma` DSL (separate file) |
| Query API | SQL-like (`select().from().where()`) | Object-based (`.findMany({where: ...})`) |
| SQL knowledge required | Yes — queries map 1:1 to SQL | No — abstracted away |
| Type safety | Full inference from schema | Generated types from schema |
| Bundle size | ~50KB (tree-shakeable) | ~1.5MB+ (generated client) |
| Serverless cold start | Fast (no engine) | Slow (Rust query engine) |
| Raw SQL escape hatch | `sql` template tag (first-class) | `$queryRaw` (second-class) |
| Migrations | `drizzle-kit` (push or generate SQL) | `prisma migrate` (SQL files) |
| Relations | Explicit join syntax or `with` API | Implicit via schema relations |
| Edge/Cloudflare Workers | Native support | Requires Accelerate proxy |
| Ecosystem maturity | Growing (newer) | Mature (5+ years) |
| Studio/GUI | Drizzle Studio (`drizzle-kit studio`) | Prisma Studio |

### When to Choose Drizzle

- **Serverless / Edge deployments** — no Rust engine, small bundle
- **SQL-first teams** — queries read like SQL, easier for DBA review
- **Performance-critical** — less overhead, no query engine proxy
- **Monorepos** — schema is TypeScript, lives in shared packages
- **Cloudflare Workers** — native D1/Turso/Neon support
- **Existing SQL knowledge** — leverage SQL skills directly

### When to Choose Prisma

- **Teams unfamiliar with SQL** — higher-level abstraction
- **Rapid prototyping** — less boilerplate for simple CRUD
- **Need Prisma ecosystem** — Prisma Pulse, Prisma Accelerate, Prisma Optimize
- **Existing Prisma projects** — migration cost is real

## Schema Design Principles

### 1. Table Naming

```typescript
// Use camelCase for table variable names, snake_case in database
export const users = pgTable('users', { ... });
export const blogPosts = pgTable('blog_posts', { ... });
```

### 2. ID Strategy

| Strategy | Pros | Cons | Use When |
|----------|------|------|----------|
| `serial()` | Simple, ordered, small | Predictable, leaks count | Internal IDs, no security concern |
| `uuid()` | Unpredictable, mergeable | Large (16 bytes), unordered | Public-facing IDs, distributed systems |
| `text()` with CUID2 | URL-safe, sortable, small | Custom generation needed | Best default for most apps |
| `text()` with ULID | Sortable, UUID-compatible | Less common | When sort order matters |

### 3. Indexing Strategy

```
Index when:
- Column appears in WHERE clauses frequently
- Column is used in JOIN conditions (foreign keys)
- Column is used in ORDER BY with LIMIT
- Unique constraints needed

Don't index:
- Boolean columns with low cardinality
- Columns rarely queried
- Small tables (< 1000 rows)
- Write-heavy columns that change constantly
```

### 4. Relation Patterns

| Relation | Pattern | Example |
|----------|---------|---------|
| One-to-many | FK on "many" side | User → Posts |
| Many-to-many | Junction table | Posts ↔ Tags |
| One-to-one | FK + unique constraint | User → Profile |
| Self-referential | FK to same table | Comment → Comment (replies) |
| Polymorphic | Type column + nullable FKs | Notification → Post or Comment |

### 5. Soft Delete Pattern

```
Add to tables that need it:
- deletedAt: timestamp (nullable)
- Filter in application layer: .where(isNull(table.deletedAt))
- Consider: index on deletedAt for query performance
- Consider: scheduled hard-delete job for old soft-deleted records
```

### 6. Audit Trail Pattern

```
For sensitive tables:
- createdAt: timestamp (default now())
- updatedAt: timestamp (updated via application or trigger)
- createdBy: text (user ID)
- updatedBy: text (user ID)
- Or: separate audit_log table with jsonb snapshots
```

## Migration Strategy

| Approach | `drizzle-kit push` | `drizzle-kit generate` + `migrate` |
|----------|-------------------|-----------------------------------|
| Development | ✅ Fast iteration | Overkill for prototyping |
| Staging | ⚠️ Fine with data reset | ✅ Reviewable SQL |
| Production | ❌ Never | ✅ Always |
| CI/CD | ❌ | ✅ Automated, auditable |
| Rollback | ❌ No history | ✅ Can write down migrations |

### Migration Workflow

```
1. Edit schema TypeScript files
2. drizzle-kit generate — creates SQL migration file
3. Review the SQL (critical for production)
4. drizzle-kit migrate — applies to database
5. Commit migration files to git
```

## Anti-Patterns

| Anti-Pattern | Problem | Fix |
|-------------|---------|-----|
| `select *` everywhere | Over-fetching, slow queries | Select specific columns |
| No indexes on FKs | Slow joins, slow deletes | Index every FK column |
| String IDs without length limit | Unbounded storage | Use `varchar(36)` or `char(26)` |
| Storing JSON for relational data | Can't query efficiently | Normalize into tables |
| No connection pooling | Connection exhaustion | Use `pg` pool or Neon pooler |
| Schema in one giant file | Hard to navigate | Split by domain (users, posts, etc.) |
| Ignoring migration SQL | Data loss, downtime | Always review generated SQL |

## Consultation Areas

1. **Schema review** — table structure, normalization, relations, naming
2. **Drizzle vs Prisma** — migration feasibility, tradeoff analysis
3. **Index optimization** — query analysis, missing indexes, over-indexing
4. **Migration planning** — zero-downtime migrations, data backfills
5. **Multi-tenant schemas** — row-level security, schema-per-tenant, shared tables
6. **Performance patterns** — connection pooling, query optimization, caching layer
