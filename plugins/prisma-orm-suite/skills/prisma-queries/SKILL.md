---
name: prisma-queries
description: >
  Prisma Client queries — CRUD operations, filtering, sorting, pagination,
  relation loading, aggregation, transactions, and raw SQL.
  Triggers: "prisma query", "prisma findMany", "prisma create", "prisma update",
  "prisma where", "prisma include", "prisma select", "prisma filter".
  NOT for: schema definition (use prisma-schema), migrations (use prisma-migrations).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Prisma Client Queries

## Client Setup

```typescript
// lib/prisma.ts — singleton pattern (critical for serverless/Next.js)
import { PrismaClient } from '@prisma/client';

const globalForPrisma = globalThis as unknown as { prisma: PrismaClient };

export const prisma =
  globalForPrisma.prisma ??
  new PrismaClient({
    log: process.env.NODE_ENV === 'development'
      ? ['query', 'error', 'warn']
      : ['error'],
  });

if (process.env.NODE_ENV !== 'production') globalForPrisma.prisma = prisma;
```

## CREATE

```typescript
// Create one
const user = await prisma.user.create({
  data: {
    email: 'alice@example.com',
    name: 'Alice',
    role: 'USER',
  },
});

// Create with nested relation
const post = await prisma.post.create({
  data: {
    title: 'My First Post',
    content: 'Hello world!',
    author: {
      connect: { id: userId },        // Link to existing user
    },
    tags: {
      connectOrCreate: [               // Link or create tags
        {
          where: { name: 'typescript' },
          create: { name: 'typescript' },
        },
      ],
    },
  },
  include: { author: true, tags: true },
});

// Create many (returns count only)
const { count } = await prisma.user.createMany({
  data: [
    { email: 'bob@example.com', name: 'Bob' },
    { email: 'carol@example.com', name: 'Carol' },
  ],
  skipDuplicates: true, // Skip rows with duplicate unique fields
});

// Create and return with specific fields
const created = await prisma.user.create({
  data: { email: 'dave@example.com', name: 'Dave' },
  select: { id: true, email: true },
});
```

## READ

```typescript
// Find unique (by unique field or composite unique)
const user = await prisma.user.findUnique({
  where: { email: 'alice@example.com' },
});

// Find unique or throw
const user = await prisma.user.findUniqueOrThrow({
  where: { id: userId },
});

// Find first matching
const post = await prisma.post.findFirst({
  where: { status: 'PUBLISHED' },
  orderBy: { createdAt: 'desc' },
});

// Find many with filtering
const posts = await prisma.post.findMany({
  where: {
    AND: [
      { status: 'PUBLISHED' },
      { authorId: userId },
    ],
  },
  orderBy: { createdAt: 'desc' },
  take: 20,
  skip: 0,
});

// Select specific fields (reduces data transfer)
const users = await prisma.user.findMany({
  select: {
    id: true,
    email: true,
    name: true,
    _count: { select: { posts: true } }, // Count relations
  },
});

// Include relations (eager loading)
const post = await prisma.post.findUnique({
  where: { id: postId },
  include: {
    author: {
      select: { id: true, name: true, avatar: true },
    },
    comments: {
      where: { deletedAt: null },
      orderBy: { createdAt: 'asc' },
      take: 50,
      include: {
        author: { select: { id: true, name: true } },
      },
    },
    _count: { select: { likes: true } },
  },
});
```

## FILTERING

```typescript
// String filters
where: {
  email: { contains: '@example.com' },     // LIKE '%@example.com%'
  email: { startsWith: 'admin' },           // LIKE 'admin%'
  email: { endsWith: '.com' },              // LIKE '%.com'
  name: { contains: 'alice', mode: 'insensitive' }, // Case-insensitive
  email: { not: 'admin@example.com' },       // != value
}

// Number filters
where: {
  age: { gt: 18 },           // >
  age: { gte: 18 },          // >=
  age: { lt: 65 },           // <
  age: { lte: 65 },          // <=
  age: { in: [18, 21, 25] }, // IN (...)
  age: { notIn: [0, -1] },   // NOT IN (...)
}

// Date filters
where: {
  createdAt: { gte: new Date('2024-01-01') },
  createdAt: { lte: new Date() },
}

// Null checks
where: {
  deletedAt: null,      // IS NULL
  deletedAt: { not: null }, // IS NOT NULL
}

// Logical operators
where: {
  AND: [{ status: 'PUBLISHED' }, { authorId: userId }],
  OR: [{ status: 'PUBLISHED' }, { authorId: currentUserId }],
  NOT: { status: 'DELETED' },
}

// Relation filters
where: {
  author: { role: 'ADMIN' },                    // Filter by related model
  comments: { some: { authorId: userId } },      // Has at least one matching
  comments: { none: { status: 'SPAM' } },        // Has zero matching
  comments: { every: { approved: true } },        // All must match
  tags: { some: { name: { in: ['typescript', 'react'] } } },
}

// JSON field filters (PostgreSQL)
where: {
  metadata: { path: ['country'], equals: 'US' },
  metadata: { path: ['tags'], array_contains: ['featured'] },
}
```

## UPDATE

```typescript
// Update one
const user = await prisma.user.update({
  where: { id: userId },
  data: {
    name: 'Alice Updated',
    role: 'ADMIN',
  },
});

// Update with nested relation
const post = await prisma.post.update({
  where: { id: postId },
  data: {
    title: 'Updated Title',
    tags: {
      set: [],                              // Remove all tags
      connect: [{ id: tag1Id }],            // Add tags
      disconnect: [{ id: tag2Id }],          // Remove specific tags
    },
  },
});

// Atomic number operations
const post = await prisma.post.update({
  where: { id: postId },
  data: {
    viewCount: { increment: 1 },   // viewCount + 1
    likes: { decrement: 1 },        // likes - 1
    score: { multiply: 2 },         // score * 2
  },
});

// Update many (returns count)
const { count } = await prisma.post.updateMany({
  where: { authorId: userId, status: 'DRAFT' },
  data: { status: 'ARCHIVED' },
});

// Upsert (create or update)
const user = await prisma.user.upsert({
  where: { email: 'alice@example.com' },
  update: { name: 'Alice Updated', lastLoginAt: new Date() },
  create: { email: 'alice@example.com', name: 'Alice', role: 'USER' },
});
```

## DELETE

```typescript
// Delete one
const deleted = await prisma.user.delete({
  where: { id: userId },
});

// Delete many
const { count } = await prisma.post.deleteMany({
  where: { status: 'ARCHIVED', createdAt: { lt: oneYearAgo } },
});

// Soft delete pattern (middleware approach)
// In schema: deletedAt DateTime?

// Middleware to auto-filter
prisma.$use(async (params, next) => {
  if (params.action === 'findMany' || params.action === 'findFirst') {
    if (!params.args.where) params.args.where = {};
    params.args.where.deletedAt = null;
  }
  if (params.action === 'delete') {
    params.action = 'update';
    params.args.data = { deletedAt: new Date() };
  }
  if (params.action === 'deleteMany') {
    params.action = 'updateMany';
    if (!params.args.data) params.args.data = {};
    params.args.data.deletedAt = new Date();
  }
  return next(params);
});
```

## PAGINATION

```typescript
// Offset-based (simple, good for < 100k rows)
async function getPosts(page: number, pageSize: number) {
  const [posts, total] = await Promise.all([
    prisma.post.findMany({
      skip: (page - 1) * pageSize,
      take: pageSize,
      orderBy: { createdAt: 'desc' },
    }),
    prisma.post.count(),
  ]);

  return {
    data: posts,
    pagination: {
      page,
      pageSize,
      total,
      totalPages: Math.ceil(total / pageSize),
    },
  };
}

// Cursor-based (performant for large datasets)
async function getPostsCursor(cursor?: string, take: number = 20) {
  const posts = await prisma.post.findMany({
    take: take + 1, // Fetch one extra to detect if more exist
    ...(cursor && {
      skip: 1,
      cursor: { id: cursor },
    }),
    orderBy: { createdAt: 'desc' },
  });

  const hasMore = posts.length > take;
  const data = hasMore ? posts.slice(0, take) : posts;

  return {
    data,
    nextCursor: hasMore ? data[data.length - 1].id : undefined,
  };
}
```

## AGGREGATION

```typescript
// Count
const count = await prisma.post.count({
  where: { status: 'PUBLISHED' },
});

// Aggregate
const stats = await prisma.order.aggregate({
  _sum: { amount: true },
  _avg: { amount: true },
  _min: { amount: true },
  _max: { amount: true },
  _count: true,
  where: { status: 'COMPLETED' },
});

// Group by
const postsByStatus = await prisma.post.groupBy({
  by: ['status'],
  _count: true,
  _avg: { viewCount: true },
  orderBy: { _count: { status: 'desc' } },
});
// Returns: [{ status: 'PUBLISHED', _count: 42, _avg: { viewCount: 150 } }, ...]

// Group by with having
const activeAuthors = await prisma.post.groupBy({
  by: ['authorId'],
  _count: { id: true },
  having: { id: { _count: { gte: 5 } } }, // Authors with 5+ posts
});
```

## TRANSACTIONS

```typescript
// Sequential transaction (auto-rollback on error)
const result = await prisma.$transaction(async (tx) => {
  const sender = await tx.account.update({
    where: { id: senderId },
    data: { balance: { decrement: amount } },
  });

  if (sender.balance < 0) {
    throw new Error('Insufficient funds'); // Rolls back everything
  }

  const recipient = await tx.account.update({
    where: { id: recipientId },
    data: { balance: { increment: amount } },
  });

  const transfer = await tx.transfer.create({
    data: { senderId, recipientId, amount },
  });

  return transfer;
});

// Batch transaction (all operations or none)
const [post, notification] = await prisma.$transaction([
  prisma.post.create({ data: { title: 'New Post', authorId: userId } }),
  prisma.notification.create({ data: { userId: adminId, message: 'New post' } }),
]);

// Transaction with isolation level
const result = await prisma.$transaction(
  async (tx) => { /* ... */ },
  {
    maxWait: 5000,          // Max time to wait for transaction slot
    timeout: 10000,         // Max transaction duration
    isolationLevel: 'Serializable',
  }
);
```

## RAW SQL

```typescript
// Raw query (returns typed result with $queryRaw)
const users = await prisma.$queryRaw<User[]>`
  SELECT * FROM users
  WHERE email LIKE ${`%${domain}`}
  ORDER BY created_at DESC
  LIMIT ${limit}
`;

// Raw execute (for INSERT, UPDATE, DELETE)
const count = await prisma.$executeRaw`
  UPDATE posts SET view_count = view_count + 1
  WHERE id = ${postId}
`;

// Raw query with Prisma.sql helper (dynamic queries)
import { Prisma } from '@prisma/client';

const orderBy = Prisma.sql`ORDER BY ${Prisma.raw(sortColumn)} ${Prisma.raw(sortDir)}`;
const users = await prisma.$queryRaw`
  SELECT id, email, name FROM users
  WHERE role = ${role}
  ${orderBy}
`;
```

## Gotchas

1. **`select` and `include` are mutually exclusive.** You can't use both in the same query. Use `select` with nested `select` for relations: `select: { id: true, author: { select: { name: true } } }`.

2. **`findUnique` requires a unique field.** You can't use `findUnique` with `where: { status: 'PUBLISHED' }` — use `findFirst` instead. `findUnique` only works with `@id` or `@unique` fields.

3. **`updateMany` and `deleteMany` don't trigger relation cascades.** Only `update` and `delete` on single records honor `onDelete: Cascade`. For bulk operations, handle cascades manually.

4. **Prisma Client is generated code.** After every schema change, run `npx prisma generate` to regenerate the client. TypeScript won't reflect schema changes until you do.

5. **The singleton pattern is essential.** Without it, each hot reload in development creates a new PrismaClient instance, eventually exhausting database connections. Always use the global singleton pattern.

6. **`@updatedAt` doesn't update on `updateMany`.** The `@updatedAt` attribute only auto-updates on single-record `update()` calls. For `updateMany`, set `updatedAt: new Date()` manually.
