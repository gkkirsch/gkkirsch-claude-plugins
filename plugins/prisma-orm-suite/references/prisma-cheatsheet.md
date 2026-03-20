# Prisma ORM Cheatsheet

## CLI Commands

```bash
npx prisma init                          # Initialize Prisma
npx prisma generate                      # Generate client from schema
npx prisma migrate dev --name <name>     # Create + apply migration
npx prisma migrate dev --create-only     # Create without applying
npx prisma migrate deploy               # Apply in production
npx prisma migrate reset                # Drop + recreate + seed
npx prisma migrate status               # Check migration status
npx prisma db push                      # Sync schema (no migration file)
npx prisma db pull                      # Introspect existing DB
npx prisma db seed                      # Run seed script
npx prisma studio                       # Open data browser
npx prisma format                       # Format schema file
npx prisma validate                     # Validate schema
```

## Schema Quick Reference

```prisma
model User {
  id        String   @id @default(cuid())
  email     String   @unique
  name      String?                          // nullable
  role      Role     @default(USER)
  data      Json?                            // JSONB
  bio       String   @db.Text               // TEXT column
  tags      String[]                         // array (Postgres)
  createdAt DateTime @default(now())
  updatedAt DateTime @updatedAt

  posts     Post[]                           // one-to-many
  profile   Profile?                         // one-to-one

  @@index([email])
  @@index([role, createdAt(sort: Desc)])
  @@map("users")                             // table name
}
```

## Relation Patterns

```prisma
// One-to-one
profile   Profile?
userId    String   @unique
user      User     @relation(fields: [userId], references: [id])

// One-to-many
posts     Post[]
userId    String
user      User     @relation(fields: [userId], references: [id])

// Many-to-many (implicit)
tags      Tag[]    // on both models

// Many-to-many (explicit)
postTags  PostTag[]
@@id([postId, tagId])

// Self-referential
managerId String?
manager   User?  @relation("Mgmt", fields: [managerId], references: [id])
reports   User[] @relation("Mgmt")
```

## Query Patterns

```typescript
// Singleton client
const prisma = globalForPrisma.prisma ?? new PrismaClient();
if (process.env.NODE_ENV !== 'production') globalForPrisma.prisma = prisma;

// CRUD
prisma.user.create({ data: { ... } })
prisma.user.createMany({ data: [...], skipDuplicates: true })
prisma.user.findUnique({ where: { id } })
prisma.user.findUniqueOrThrow({ where: { id } })
prisma.user.findFirst({ where: { ... }, orderBy: { ... } })
prisma.user.findMany({ where, orderBy, take, skip, select, include })
prisma.user.update({ where: { id }, data: { ... } })
prisma.user.updateMany({ where: { ... }, data: { ... } })
prisma.user.upsert({ where, update, create })
prisma.user.delete({ where: { id } })
prisma.user.deleteMany({ where: { ... } })
prisma.user.count({ where: { ... } })
prisma.user.aggregate({ _sum, _avg, _min, _max, _count })
prisma.user.groupBy({ by: [...], _count: true })
```

## Filter Operators

```typescript
{ equals: value }           // = (default, can omit)
{ not: value }              // !=
{ in: [a, b, c] }          // IN
{ notIn: [a, b] }          // NOT IN
{ gt: n }                  // >
{ gte: n }                 // >=
{ lt: n }                  // <
{ lte: n }                 // <=
{ contains: str }          // LIKE '%str%'
{ startsWith: str }        // LIKE 'str%'
{ endsWith: str }          // LIKE '%str'
{ mode: 'insensitive' }   // case insensitive (add to string filter)
null                       // IS NULL
{ not: null }              // IS NOT NULL
AND: [...]                 // AND conditions
OR: [...]                  // OR conditions
NOT: { ... }               // NOT condition
```

## Relation Filters

```typescript
{ author: { role: 'ADMIN' } }                  // related field match
{ comments: { some: { ... } } }                // has at least one
{ comments: { none: { ... } } }                // has zero
{ comments: { every: { ... } } }               // all must match
{ _count: { select: { posts: true } } }        // count relations
```

## Atomic Operations

```typescript
{ increment: 1 }    // field + 1
{ decrement: 1 }    // field - 1
{ multiply: 2 }     // field * 2
{ divide: 2 }       // field / 2
```

## Pagination

```typescript
// Offset-based
{ skip: (page - 1) * size, take: size }

// Cursor-based
{ take: size + 1, cursor: { id: lastId }, skip: 1 }
// hasMore = results.length > size
```

## Transactions

```typescript
// Interactive (auto-rollback)
await prisma.$transaction(async (tx) => {
  await tx.model.action({ ... });
  await tx.model.action({ ... });
}, { timeout: 30000 });

// Batch (all-or-nothing)
await prisma.$transaction([
  prisma.model.action({ ... }),
  prisma.model.action({ ... }),
]);
```

## Extensions

```typescript
prisma.$extends({
  model: { user: { async customMethod() { ... } } },
  result: { user: { computed: { needs: {}, compute() {} } } },
  query: { $allModels: { async findMany({ args, query }) { ... } } },
  client: { async $customMethod() { ... } },
});
```

## Common Patterns

| Pattern | Implementation |
|---------|---------------|
| Soft delete | `deletedAt DateTime?` + middleware/extension filter |
| Audit trail | `createdAt @default(now())` + `updatedAt @updatedAt` |
| Slug | `slug String @unique` + generate before create |
| Multi-tenant | `orgId String` on every model + extension scope |
| Optimistic lock | `version Int @default(0)` + check before update |
| Full-text search | `@@index([field], type: Gin)` + `$queryRaw` |

## Environment

```bash
DATABASE_URL="postgresql://user:pass@host:5432/db?schema=public"
# PgBouncer: append &pgbouncer=true
# Prisma Accelerate: use accelerate URL
DIRECT_DATABASE_URL="postgresql://user:pass@host:5432/db"
# For migrations when using PgBouncer/Accelerate
```
