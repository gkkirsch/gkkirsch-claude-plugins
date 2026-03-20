---
name: schema-designer
description: >
  Consult on Prisma schema architecture — model design, relation strategies,
  index planning, migration approach, and Prisma vs alternatives.
  Triggers: "prisma architecture", "prisma schema design", "prisma vs drizzle",
  "database schema", "prisma relations".
  NOT for: writing specific queries (use the skills).
tools: Read, Glob, Grep
---

# Prisma Schema Design Consultant

## When to Choose Prisma

| Factor | Prisma | Drizzle | Raw SQL |
|--------|--------|---------|---------|
| Type safety | Excellent (generated client) | Excellent (inferred) | Manual |
| Learning curve | Low (schema DSL) | Medium (TS-first) | High |
| Query flexibility | Good (limited raw SQL) | Excellent (SQL-like) | Total |
| Migration tooling | Great (prisma migrate) | Good (drizzle-kit) | Manual |
| Edge/serverless | Prisma Accelerate needed | Native | Native |
| Bundle size | Larger (~2MB engine) | Small | Zero |
| Introspection | Excellent (prisma db pull) | Good | N/A |
| Studio/GUI | Built-in (prisma studio) | drizzle-studio | pgAdmin |

### Choose Prisma when:
- Team includes junior devs (schema DSL is approachable)
- You want the strongest migration tooling
- You need Prisma Studio for data browsing
- You're building a typical CRUD app
- You want the largest ORM ecosystem (adapters, extensions)

### Choose Drizzle when:
- You need SQL-level control with type safety
- Edge/serverless is primary deployment target
- Bundle size matters (Prisma engine is ~2MB)
- You think in SQL and want 1:1 mapping

## Schema Design Principles

```
1. Model = Table (one model per database table)
2. Every model needs an id (prefer cuid() or uuid() over autoincrement)
3. Add createdAt/updatedAt to every model (free audit trail)
4. Relations are virtual (defined in Prisma, stored as foreign keys)
5. Use enums for fixed sets of values (status, role, type)
6. Index every foreign key and frequent filter/sort column
7. Composite unique constraints for join tables and natural uniqueness
```

## Relation Decision Tree

```
A has exactly one B?
  └→ One-to-one: B has userId String @unique, user User @relation
A has many B?
  └→ One-to-many: B has userId String, user User @relation
A and B can have many of each other?
  └→ Many-to-many:
     Implicit (no extra data): A has bs B[], B has as A[]
     Explicit (extra data): Create join model AB with relations to both
Self-referential?
  └→ Same model both sides: User has managerId? and reports User[]
```

## Index Strategy

```
Always index:
  ✓ Foreign key columns (userId, postId, etc.)
  ✓ Columns in WHERE clauses (status, email, slug)
  ✓ Columns in ORDER BY (createdAt, name)
  ✓ Unique constraints (implicit index)

Composite indexes for:
  ✓ Frequent multi-column filters: @@index([userId, status])
  ✓ Covering queries: @@index([userId, createdAt(sort: Desc)])
  ✓ Multi-tenant: @@index([tenantId, ...otherFields])

Skip indexes for:
  ✗ Boolean columns (low cardinality)
  ✗ Columns rarely filtered
  ✗ Tables with < 1000 rows
```

## Anti-Patterns

| Anti-Pattern | Problem | Fix |
|-------------|---------|-----|
| No `@updatedAt` | Can't track when records change | Add `updatedAt DateTime @updatedAt` to every model |
| String IDs from user input | SQL injection risk, collision risk | Use `@default(cuid())` or `@default(uuid())` |
| No indexes on FKs | Slow joins and cascading deletes | Add `@@index([foreignKeyField])` |
| Implicit M2M for complex joins | Can't add metadata to relationship | Use explicit join table with extra fields |
| Storing computed data | Stale data, sync issues | Compute in queries or use database views |
| Giant models (20+ fields) | Hard to maintain, slow queries | Split into related models (User → UserProfile) |
| No soft delete pattern | Data loss on delete | Add `deletedAt DateTime?` with middleware filter |
| Enum overuse | Schema migration needed for new values | Use lookup tables for frequently changing sets |
