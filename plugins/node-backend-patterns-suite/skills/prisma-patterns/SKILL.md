---
name: prisma-patterns
description: >
  Prisma ORM patterns — schema design, CRUD operations, relations, transactions,
  migrations, query optimization, and production configuration.
  Triggers: "prisma", "prisma schema", "prisma query", "prisma migration",
  "prisma relation", "prisma transaction", "database orm", "prisma client".
  NOT for: raw SQL optimization (use postgres-patterns), schema architecture (use api-architect agent).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Prisma Patterns

## Schema Design

```prisma
// prisma/schema.prisma
generator client {
  provider = "prisma-client-js"
}

datasource db {
  provider = "postgresql"
  url      = env("DATABASE_URL")
}

model User {
  id        String   @id @default(uuid())
  email     String   @unique
  password  String?
  name      String
  role      Role     @default(USER)
  avatar    String?
  googleId  String?  @unique

  posts     Post[]
  comments  Comment[]
  apiKeys   ApiKey[]
  refreshTokens RefreshToken[]

  createdAt DateTime @default(now())
  updatedAt DateTime @updatedAt

  @@index([email])
  @@map("users")
}

model Post {
  id        String   @id @default(uuid())
  title     String
  slug      String   @unique
  content   String
  excerpt   String?
  published Boolean  @default(false)

  author    User     @relation(fields: [authorId], references: [id], onDelete: Cascade)
  authorId  String

  tags      Tag[]
  comments  Comment[]
  category  Category? @relation(fields: [categoryId], references: [id])
  categoryId String?

  publishedAt DateTime?
  createdAt   DateTime  @default(now())
  updatedAt   DateTime  @updatedAt

  @@index([authorId])
  @@index([slug])
  @@index([published, createdAt])
  @@map("posts")
}

model Tag {
  id    String @id @default(uuid())
  name  String @unique
  posts Post[]

  @@map("tags")
}

model Comment {
  id      String @id @default(uuid())
  content String

  author   User   @relation(fields: [authorId], references: [id], onDelete: Cascade)
  authorId String

  post   Post   @relation(fields: [postId], references: [id], onDelete: Cascade)
  postId String

  parent   Comment?  @relation("CommentReplies", fields: [parentId], references: [id])
  parentId String?
  replies  Comment[] @relation("CommentReplies")

  createdAt DateTime @default(now())
  updatedAt DateTime @updatedAt

  @@index([postId])
  @@index([authorId])
  @@map("comments")
}

model Category {
  id          String  @id @default(uuid())
  name        String  @unique
  slug        String  @unique
  description String?
  posts       Post[]

  @@map("categories")
}

model RefreshToken {
  id        String   @id @default(uuid())
  token     String   @unique
  user      User     @relation(fields: [userId], references: [id], onDelete: Cascade)
  userId    String
  expiresAt DateTime
  createdAt DateTime @default(now())

  @@index([token])
  @@index([userId])
  @@map("refresh_tokens")
}

model ApiKey {
  id         String    @id @default(uuid())
  keyHash    String    @unique
  name       String
  scope      String    @default("read")
  user       User      @relation(fields: [userId], references: [id], onDelete: Cascade)
  userId     String
  lastUsedAt DateTime?
  revokedAt  DateTime?
  createdAt  DateTime  @default(now())

  @@index([keyHash])
  @@index([userId])
  @@map("api_keys")
}

enum Role {
  USER
  ADMIN
}
```

## Client Setup

```typescript
// src/lib/prisma.ts
import { PrismaClient } from '@prisma/client';

// Prevent multiple instances in development (hot reload)
const globalForPrisma = globalThis as unknown as { prisma: PrismaClient };

export const prisma =
  globalForPrisma.prisma ||
  new PrismaClient({
    log:
      process.env.NODE_ENV === 'development'
        ? ['query', 'info', 'warn', 'error']
        : ['error'],
  });

if (process.env.NODE_ENV !== 'production') {
  globalForPrisma.prisma = prisma;
}
```

## CRUD Service Pattern

```typescript
// src/services/posts.service.ts
import { prisma } from '../lib/prisma';
import { Prisma } from '@prisma/client';
import { NotFoundError } from '../lib/errors';

interface ListOptions {
  page: number;
  limit: number;
  sort: string;
  order?: 'asc' | 'desc';
  search?: string;
  userId?: string;
  published?: boolean;
}

export class PostService {
  async list(options: ListOptions) {
    const { page, limit, sort, order = 'desc', search, userId, published } = options;
    const skip = (page - 1) * limit;

    const where: Prisma.PostWhereInput = {
      ...(userId && { authorId: userId }),
      ...(published !== undefined && { published }),
      ...(search && {
        OR: [
          { title: { contains: search, mode: 'insensitive' } },
          { content: { contains: search, mode: 'insensitive' } },
        ],
      }),
    };

    const [items, total] = await prisma.$transaction([
      prisma.post.findMany({
        where,
        skip,
        take: limit,
        orderBy: { [sort]: order },
        include: {
          author: { select: { id: true, name: true, avatar: true } },
          tags: { select: { id: true, name: true } },
          _count: { select: { comments: true } },
        },
      }),
      prisma.post.count({ where }),
    ]);

    return {
      items,
      page,
      limit,
      total,
      totalPages: Math.ceil(total / limit),
    };
  }

  async findById(id: string) {
    return prisma.post.findUnique({
      where: { id },
      include: {
        author: { select: { id: true, name: true, avatar: true } },
        tags: true,
        comments: {
          where: { parentId: null },
          include: {
            author: { select: { id: true, name: true, avatar: true } },
            replies: {
              include: {
                author: { select: { id: true, name: true, avatar: true } },
              },
            },
          },
          orderBy: { createdAt: 'desc' },
        },
      },
    });
  }

  async findBySlug(slug: string) {
    return prisma.post.findUnique({
      where: { slug },
      include: {
        author: { select: { id: true, name: true, avatar: true } },
        tags: true,
      },
    });
  }

  async create(data: {
    title: string;
    content: string;
    published?: boolean;
    tags?: string[];
    authorId: string;
  }) {
    const slug = this.generateSlug(data.title);

    return prisma.post.create({
      data: {
        title: data.title,
        slug,
        content: data.content,
        published: data.published ?? false,
        publishedAt: data.published ? new Date() : null,
        author: { connect: { id: data.authorId } },
        tags: data.tags?.length
          ? {
              connectOrCreate: data.tags.map((name) => ({
                where: { name },
                create: { name },
              })),
            }
          : undefined,
      },
      include: {
        author: { select: { id: true, name: true } },
        tags: true,
      },
    });
  }

  async update(id: string, data: Partial<{
    title: string;
    content: string;
    published: boolean;
    tags: string[];
  }>, userId: string) {
    const post = await prisma.post.findUnique({ where: { id } });
    if (!post) throw new NotFoundError('Post');
    if (post.authorId !== userId) throw new Error('Not authorized');

    return prisma.post.update({
      where: { id },
      data: {
        ...(data.title && { title: data.title, slug: this.generateSlug(data.title) }),
        ...(data.content !== undefined && { content: data.content }),
        ...(data.published !== undefined && {
          published: data.published,
          publishedAt: data.published && !post.publishedAt ? new Date() : post.publishedAt,
        }),
        ...(data.tags && {
          tags: {
            set: [], // Disconnect all existing
            connectOrCreate: data.tags.map((name) => ({
              where: { name },
              create: { name },
            })),
          },
        }),
      },
      include: { tags: true },
    });
  }

  async delete(id: string, userId: string) {
    const post = await prisma.post.findUnique({ where: { id } });
    if (!post) throw new NotFoundError('Post');
    if (post.authorId !== userId) throw new Error('Not authorized');

    await prisma.post.delete({ where: { id } });
  }

  private generateSlug(title: string): string {
    return title
      .toLowerCase()
      .replace(/[^\w\s-]/g, '')
      .replace(/\s+/g, '-')
      .replace(/-+/g, '-')
      .slice(0, 100);
  }
}
```

## Transactions

```typescript
// Simple transaction — all or nothing
const [post, notification] = await prisma.$transaction([
  prisma.post.create({ data: postData }),
  prisma.notification.create({ data: notifData }),
]);

// Interactive transaction — with business logic
const result = await prisma.$transaction(async (tx) => {
  const account = await tx.account.findUnique({
    where: { id: fromAccountId },
  });

  if (!account || account.balance < amount) {
    throw new Error('Insufficient funds');
  }

  const [from, to] = await Promise.all([
    tx.account.update({
      where: { id: fromAccountId },
      data: { balance: { decrement: amount } },
    }),
    tx.account.update({
      where: { id: toAccountId },
      data: { balance: { increment: amount } },
    }),
  ]);

  return { from, to };
}, {
  maxWait: 5000,    // Max time to acquire a transaction slot
  timeout: 10000,   // Max transaction duration
  isolationLevel: Prisma.TransactionIsolationLevel.Serializable,
});
```

## Migrations

```bash
# Create migration from schema changes
npx prisma migrate dev --name add-posts-table

# Apply migrations in production
npx prisma migrate deploy

# Reset database (drops all data)
npx prisma migrate reset

# Generate client after schema change (no migration)
npx prisma generate

# Seed database
npx prisma db seed
```

```typescript
// prisma/seed.ts
import { PrismaClient } from '@prisma/client';
import { hashPassword } from '../src/lib/password';

const prisma = new PrismaClient();

async function main() {
  const adminPassword = await hashPassword('admin123');

  const admin = await prisma.user.upsert({
    where: { email: 'admin@example.com' },
    update: {},
    create: {
      email: 'admin@example.com',
      name: 'Admin',
      password: adminPassword,
      role: 'ADMIN',
    },
  });

  const categories = ['Technology', 'Design', 'Business'];
  for (const name of categories) {
    await prisma.category.upsert({
      where: { name },
      update: {},
      create: {
        name,
        slug: name.toLowerCase(),
      },
    });
  }

  console.log('Seeded:', { admin: admin.email, categories });
}

main()
  .catch(console.error)
  .finally(() => prisma.$disconnect());
```

```json
// package.json
{
  "prisma": {
    "seed": "tsx prisma/seed.ts"
  }
}
```

## Query Optimization

```typescript
// Select only needed fields
const users = await prisma.user.findMany({
  select: {
    id: true,
    name: true,
    email: true,
    // password NOT selected
  },
});

// Cursor-based pagination (better for large datasets)
const posts = await prisma.post.findMany({
  take: 20,
  skip: 1,                          // Skip the cursor itself
  cursor: { id: lastPostId },
  orderBy: { createdAt: 'desc' },
});

// Batch operations
await prisma.post.updateMany({
  where: { authorId: userId, published: false },
  data: { published: true, publishedAt: new Date() },
});

await prisma.post.deleteMany({
  where: { authorId: userId, createdAt: { lt: oneYearAgo } },
});

// Raw queries for complex operations
const stats = await prisma.$queryRaw<{ month: string; count: bigint }[]>`
  SELECT
    TO_CHAR(created_at, 'YYYY-MM') as month,
    COUNT(*) as count
  FROM posts
  WHERE author_id = ${userId}
  GROUP BY month
  ORDER BY month DESC
  LIMIT 12
`;
```

## Middleware (Prisma-Level)

```typescript
// Soft delete middleware
prisma.$use(async (params, next) => {
  if (params.model === 'Post') {
    if (params.action === 'delete') {
      params.action = 'update';
      params.args.data = { deletedAt: new Date() };
    }
    if (params.action === 'deleteMany') {
      params.action = 'updateMany';
      if (params.args.data) {
        params.args.data.deletedAt = new Date();
      } else {
        params.args.data = { deletedAt: new Date() };
      }
    }
    // Filter out soft-deleted records from reads
    if (params.action === 'findMany' || params.action === 'findFirst') {
      if (!params.args) params.args = {};
      if (!params.args.where) params.args.where = {};
      params.args.where.deletedAt = null;
    }
  }
  return next(params);
});

// Query timing middleware
prisma.$use(async (params, next) => {
  const start = Date.now();
  const result = await next(params);
  const duration = Date.now() - start;
  if (duration > 1000) {
    logger.warn({ model: params.model, action: params.action, duration }, 'Slow query');
  }
  return result;
});
```

## Testing with Prisma

```typescript
// test/helpers/prisma.ts
import { PrismaClient } from '@prisma/client';

const prisma = new PrismaClient();

export async function cleanDatabase() {
  const tables = await prisma.$queryRaw<{ tablename: string }[]>`
    SELECT tablename FROM pg_tables WHERE schemaname = 'public'
  `;

  for (const { tablename } of tables) {
    if (tablename === '_prisma_migrations') continue;
    await prisma.$executeRawUnsafe(`TRUNCATE TABLE "${tablename}" CASCADE`);
  }
}

// Use in tests
beforeEach(async () => {
  await cleanDatabase();
});

afterAll(async () => {
  await prisma.$disconnect();
});
```

## Gotchas

1. **Global PrismaClient in development.** Hot module reload creates new PrismaClient instances. Use the `globalThis` pattern to reuse the client across reloads. Without it, you'll exhaust database connections.

2. **N+1 queries with relations.** Use `include` or `select` with nested relations to fetch everything in one query. Don't loop over results and query each relation individually.

3. **`BigInt` from raw queries.** PostgreSQL `COUNT(*)` returns `bigint`. JSON.stringify can't serialize BigInt. Convert with `Number()` or use `count` helper instead of raw queries for simple counts.

4. **Migration vs `db push`.** Use `prisma migrate dev` for production schemas (creates SQL migration files). Use `prisma db push` only for prototyping (no migration history, can lose data).

5. **Unique constraint on connect-or-create.** `connectOrCreate` needs a `where` with a unique field. Non-unique fields won't work and throw confusing errors.

6. **`@@index` placement matters.** Add indexes for fields used in `where`, `orderBy`, and `join` conditions. Don't index everything — each index slows writes. Start with: foreign keys, unique lookups, common filters.

7. **`onDelete: Cascade` is opt-in.** By default, Prisma uses `Restrict` (prevents deletion if related records exist). Use `Cascade` explicitly for parent-child relationships where deleting the parent should delete children.
