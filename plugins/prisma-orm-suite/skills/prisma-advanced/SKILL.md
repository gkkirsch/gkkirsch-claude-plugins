---
name: prisma-advanced
description: >
  Advanced Prisma patterns — middleware, extensions, client extensions API,
  multi-tenancy, testing, connection pooling, edge deployment, and performance.
  Triggers: "prisma middleware", "prisma extension", "prisma testing", "prisma multi-tenant",
  "prisma edge", "prisma accelerate", "prisma performance", "prisma pool".
  NOT for: basic schema (use prisma-schema), basic queries (use prisma-queries).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Advanced Prisma Patterns

## Prisma Client Extensions (v4.16+)

```typescript
// Extend Prisma Client with custom methods
const prisma = new PrismaClient().$extends({
  // Add methods to specific models
  model: {
    user: {
      async findByEmail(email: string) {
        return prisma.user.findUnique({ where: { email } });
      },
      async signUp(data: { email: string; name: string; password: string }) {
        const hashedPassword = await hash(data.password, 10);
        return prisma.user.create({
          data: { ...data, password: hashedPassword },
        });
      },
    },
    post: {
      async publish(id: string) {
        return prisma.post.update({
          where: { id },
          data: { status: 'PUBLISHED', publishedAt: new Date() },
        });
      },
    },
  },

  // Add computed fields to query results
  result: {
    user: {
      fullName: {
        needs: { firstName: true, lastName: true },
        compute(user) {
          return `${user.firstName} ${user.lastName}`;
        },
      },
    },
    post: {
      excerpt: {
        needs: { content: true },
        compute(post) {
          return post.content.slice(0, 200) + (post.content.length > 200 ? '...' : '');
        },
      },
    },
  },

  // Intercept and modify queries
  query: {
    $allModels: {
      async findMany({ args, query }) {
        // Auto-filter soft-deleted records
        args.where = { ...args.where, deletedAt: null };
        return query(args);
      },
    },
    user: {
      async delete({ args, query }) {
        // Soft delete instead of hard delete
        return prisma.user.update({
          where: args.where,
          data: { deletedAt: new Date() },
        });
      },
    },
  },

  // Add client-level methods
  client: {
    async $resetDatabase() {
      if (process.env.NODE_ENV === 'production') {
        throw new Error('Cannot reset production database');
      }
      const tables = await prisma.$queryRaw<{ tablename: string }[]>`
        SELECT tablename FROM pg_tables WHERE schemaname = 'public'
      `;
      for (const { tablename } of tables) {
        if (tablename !== '_prisma_migrations') {
          await prisma.$executeRawUnsafe(`TRUNCATE TABLE "${tablename}" CASCADE`);
        }
      }
    },
  },
});

// Usage
const user = await prisma.user.findByEmail('alice@example.com');
const post = await prisma.post.publish(postId);
await prisma.$resetDatabase(); // Client-level method
```

## Middleware (Legacy — Use Extensions for New Code)

```typescript
// Logging middleware
prisma.$use(async (params, next) => {
  const before = Date.now();
  const result = await next(params);
  const after = Date.now();

  console.log(`${params.model}.${params.action} took ${after - before}ms`);
  return result;
});

// Soft delete middleware
prisma.$use(async (params, next) => {
  // Auto-filter deleted records on reads
  if (['findFirst', 'findMany', 'count'].includes(params.action)) {
    if (!params.args.where) params.args.where = {};
    if (params.args.where.deletedAt === undefined) {
      params.args.where.deletedAt = null;
    }
  }

  // Convert delete to soft delete
  if (params.action === 'delete') {
    params.action = 'update';
    params.args.data = { deletedAt: new Date() };
  }

  return next(params);
});

// Audit log middleware
prisma.$use(async (params, next) => {
  const result = await next(params);

  if (['create', 'update', 'delete'].includes(params.action)) {
    await prisma.auditLog.create({
      data: {
        model: params.model!,
        action: params.action,
        recordId: result?.id,
        data: JSON.stringify(params.args.data),
        timestamp: new Date(),
      },
    });
  }

  return result;
});
```

## Multi-Tenancy

```prisma
// Schema: Row-Level Security approach
model Organization {
  id   String @id @default(cuid())
  name String
  // All tenant-scoped models reference this
}

model Post {
  id             String       @id @default(cuid())
  title          String
  organizationId String
  organization   Organization @relation(fields: [organizationId], references: [id])

  @@index([organizationId])
  @@index([organizationId, createdAt(sort: Desc)])
}
```

```typescript
// Tenant-scoped Prisma Client via extension
function createTenantClient(organizationId: string) {
  return prisma.$extends({
    query: {
      $allModels: {
        async findMany({ args, query }) {
          args.where = { ...args.where, organizationId };
          return query(args);
        },
        async findFirst({ args, query }) {
          args.where = { ...args.where, organizationId };
          return query(args);
        },
        async create({ args, query }) {
          args.data = { ...args.data, organizationId };
          return query(args);
        },
        async update({ args, query }) {
          args.where = { ...args.where, organizationId };
          return query(args);
        },
        async delete({ args, query }) {
          args.where = { ...args.where, organizationId };
          return query(args);
        },
        async count({ args, query }) {
          args.where = { ...args.where, organizationId };
          return query(args);
        },
      },
    },
  });
}

// Usage in API route
app.use(async (req, res, next) => {
  const orgId = req.headers['x-org-id'] as string;
  req.db = createTenantClient(orgId);
  next();
});

app.get('/posts', async (req, res) => {
  // Automatically scoped to tenant
  const posts = await req.db.post.findMany();
  res.json(posts);
});
```

## Testing with Prisma

```typescript
// test/helpers/prisma.ts
import { PrismaClient } from '@prisma/client';

const prisma = new PrismaClient({
  datasourceUrl: process.env.DATABASE_URL_TEST,
});

// Reset database between tests
async function resetDatabase() {
  const tables = await prisma.$queryRaw<{ tablename: string }[]>`
    SELECT tablename FROM pg_tables
    WHERE schemaname = 'public' AND tablename != '_prisma_migrations'
  `;

  await prisma.$transaction(
    tables.map(({ tablename }) =>
      prisma.$executeRawUnsafe(`TRUNCATE TABLE "${tablename}" CASCADE`)
    )
  );
}

// Test data factories
function createUserData(overrides = {}) {
  return {
    email: `test-${Date.now()}@example.com`,
    name: 'Test User',
    password: 'hashed-password',
    role: 'USER' as const,
    ...overrides,
  };
}

async function createUser(overrides = {}) {
  return prisma.user.create({
    data: createUserData(overrides),
  });
}

async function createPostWithAuthor(authorOverrides = {}, postOverrides = {}) {
  const author = await createUser(authorOverrides);
  const post = await prisma.post.create({
    data: {
      title: 'Test Post',
      content: 'Test content',
      authorId: author.id,
      status: 'PUBLISHED',
      ...postOverrides,
    },
  });
  return { author, post };
}

export { prisma, resetDatabase, createUser, createPostWithAuthor };
```

```typescript
// tests/user.test.ts
import { describe, it, expect, beforeEach } from 'vitest';
import { prisma, resetDatabase, createUser } from './helpers/prisma';

describe('User', () => {
  beforeEach(async () => {
    await resetDatabase();
  });

  it('creates a user with default role', async () => {
    const user = await createUser();
    expect(user.role).toBe('USER');
  });

  it('enforces unique email', async () => {
    await createUser({ email: 'dupe@example.com' });

    await expect(
      createUser({ email: 'dupe@example.com' })
    ).rejects.toThrow('Unique constraint');
  });

  it('cascades delete to posts', async () => {
    const user = await createUser();
    await prisma.post.create({
      data: { title: 'Test', content: 'Content', authorId: user.id },
    });

    await prisma.user.delete({ where: { id: user.id } });

    const posts = await prisma.post.findMany({
      where: { authorId: user.id },
    });
    expect(posts).toHaveLength(0);
  });
});
```

## Connection Pooling

```typescript
// PgBouncer configuration (production)
// DATABASE_URL="postgresql://user:pass@pgbouncer:6432/myapp?pgbouncer=true"

const prisma = new PrismaClient({
  datasources: {
    db: {
      url: process.env.DATABASE_URL,
    },
  },
});

// Connection pool settings via URL params:
// ?connection_limit=5     — max connections per Prisma instance
// ?pool_timeout=10        — seconds to wait for a connection
// ?pgbouncer=true         — required when using PgBouncer
// ?connect_timeout=10     — seconds to wait for initial connection
```

```prisma
// For PgBouncer: add directUrl for migrations
datasource db {
  provider  = "postgresql"
  url       = env("DATABASE_URL")       // PgBouncer URL (for queries)
  directUrl = env("DIRECT_DATABASE_URL") // Direct URL (for migrations)
}
```

## Edge & Serverless Deployment

```typescript
// Prisma Accelerate (connection pooling + caching for edge)
// prisma/schema.prisma
generator client {
  provider = "prisma-client-js"
}

datasource db {
  provider  = "postgresql"
  url       = env("DATABASE_URL")  // Accelerate connection string
  directUrl = env("DIRECT_URL")    // Direct URL for migrations
}

// Usage with caching
const posts = await prisma.post.findMany({
  cacheStrategy: {
    ttl: 60,      // Cache for 60 seconds
    swr: 120,     // Serve stale for 120s while revalidating
  },
  where: { status: 'PUBLISHED' },
  orderBy: { createdAt: 'desc' },
});
```

## Performance Optimization

```typescript
// 1. Select only needed fields
const users = await prisma.user.findMany({
  select: { id: true, name: true, email: true },
  // NOT: include everything by default
});

// 2. Batch operations in transactions
const results = await prisma.$transaction(
  ids.map(id =>
    prisma.post.update({
      where: { id },
      data: { status: 'ARCHIVED' },
    })
  )
);

// 3. Use createMany for bulk inserts
await prisma.event.createMany({
  data: events, // Array of objects
  skipDuplicates: true,
});

// 4. Avoid N+1 with include/select
// BAD: N+1 problem
const posts = await prisma.post.findMany();
for (const post of posts) {
  const author = await prisma.user.findUnique({ where: { id: post.authorId } });
  // ^ This runs N additional queries
}

// GOOD: Single query with include
const posts = await prisma.post.findMany({
  include: { author: { select: { id: true, name: true } } },
});

// 5. Use raw SQL for complex aggregations
const stats = await prisma.$queryRaw`
  SELECT
    DATE_TRUNC('day', created_at) as day,
    COUNT(*) as count,
    SUM(amount) as total
  FROM orders
  WHERE created_at >= ${startDate}
  GROUP BY 1
  ORDER BY 1 DESC
`;
```

## Logging & Debugging

```typescript
const prisma = new PrismaClient({
  log: [
    { level: 'query', emit: 'event' },
    { level: 'error', emit: 'stdout' },
    { level: 'warn', emit: 'stdout' },
  ],
});

// Log slow queries
prisma.$on('query', (e) => {
  if (e.duration > 100) { // > 100ms
    console.warn(`Slow query (${e.duration}ms): ${e.query}`);
    console.warn(`Params: ${e.params}`);
  }
});
```

## Gotchas

1. **Prisma Client Extensions replace middleware for new projects.** Extensions are type-safe, composable, and don't mutate the original client. Middleware is legacy but still supported.

2. **PgBouncer requires `?pgbouncer=true` in the URL.** Without it, Prisma's prepared statements fail because PgBouncer in transaction mode doesn't support them. Also use `directUrl` for migrations.

3. **`$transaction` has a default 5-second timeout.** Long-running transactions silently fail. Set `timeout` explicitly: `$transaction(async (tx) => { ... }, { timeout: 30000 })`.

4. **Extension `query` hooks don't apply to `$queryRaw`.** If you rely on extensions for soft-delete filtering, raw queries bypass them. Add filtering manually to raw queries.

5. **Each `$extends()` creates a NEW client instance.** Don't call `$extends()` on every request — create the extended client once at startup (or per tenant if needed).

6. **Prisma Accelerate requires a paid plan for production.** The free tier has rate limits. For edge deployment without Accelerate, consider Neon's serverless driver or Supabase's edge functions.
