---
name: database-patterns
description: >
  Database patterns for Node.js APIs — Prisma ORM, transactions, repository pattern,
  migrations, seeding, query optimization, and connection management.
  Triggers: "prisma", "database pattern", "repository pattern", "transaction",
  "migration", "database query", "prisma schema", "database seed".
  NOT for: raw SQL (use postgres-security), MongoDB (different ORM patterns).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Database Patterns

## Prisma Setup

```typescript
// src/lib/prisma.ts
import { PrismaClient } from '@prisma/client';

// Singleton pattern — prevent multiple instances in development (hot reload)
const globalForPrisma = globalThis as unknown as { prisma: PrismaClient };

export const prisma =
  globalForPrisma.prisma ||
  new PrismaClient({
    log:
      process.env.NODE_ENV === 'development'
        ? ['query', 'warn', 'error']
        : ['error'],
  });

if (process.env.NODE_ENV !== 'production') {
  globalForPrisma.prisma = prisma;
}
```

## Common Schema Patterns

```prisma
// prisma/schema.prisma

generator client {
  provider        = "prisma-client-js"
  previewFeatures = ["fullTextSearch"]
}

datasource db {
  provider = "postgresql"
  url      = env("DATABASE_URL")
}

// Base fields pattern (copy to each model)
model User {
  id        String   @id @default(uuid())
  email     String   @unique
  name      String
  role      Role     @default(USER)
  avatar    String?
  createdAt DateTime @default(now()) @map("created_at")
  updatedAt DateTime @updatedAt @map("updated_at")

  posts     Post[]
  comments  Comment[]

  @@map("users")
}

model Post {
  id          String     @id @default(uuid())
  title       String
  slug        String     @unique
  content     String
  excerpt     String?
  published   Boolean    @default(false)
  publishedAt DateTime?  @map("published_at")
  createdAt   DateTime   @default(now()) @map("created_at")
  updatedAt   DateTime   @updatedAt @map("updated_at")

  author      User       @relation(fields: [authorId], references: [id])
  authorId    String     @map("author_id")
  category    Category?  @relation(fields: [categoryId], references: [id])
  categoryId  String?    @map("category_id")
  tags        Tag[]
  comments    Comment[]

  @@index([authorId])
  @@index([categoryId])
  @@index([publishedAt])
  @@index([slug])
  @@map("posts")
}

model Category {
  id    String @id @default(uuid())
  name  String @unique
  slug  String @unique
  posts Post[]

  @@map("categories")
}

model Tag {
  id    String @id @default(uuid())
  name  String @unique
  posts Post[]

  @@map("tags")
}

model Comment {
  id        String   @id @default(uuid())
  content   String
  createdAt DateTime @default(now()) @map("created_at")

  author    User     @relation(fields: [authorId], references: [id])
  authorId  String   @map("author_id")
  post      Post     @relation(fields: [postId], references: [id], onDelete: Cascade)
  postId    String   @map("post_id")

  // Self-referencing for replies
  parent    Comment? @relation("CommentReplies", fields: [parentId], references: [id])
  parentId  String?  @map("parent_id")
  replies   Comment[] @relation("CommentReplies")

  @@index([postId])
  @@index([authorId])
  @@map("comments")
}

enum Role {
  USER
  ADMIN
  MODERATOR
}
```

## Service Layer (CRUD)

```typescript
// src/services/posts.service.ts
import { Prisma } from '@prisma/client';
import { prisma } from '../lib/prisma';
import { NotFoundError, ForbiddenError } from '../lib/errors';

interface ListOptions {
  page: number;
  limit: number;
  sort: string;
  order: 'asc' | 'desc';
  search?: string;
  authorId?: string;
  published?: boolean;
}

export class PostService {
  async list(options: ListOptions) {
    const { page, limit, sort, order, search, authorId, published } = options;
    const skip = (page - 1) * limit;

    const where: Prisma.PostWhereInput = {
      ...(authorId && { authorId }),
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
        select: {
          id: true,
          title: true,
          slug: true,
          excerpt: true,
          published: true,
          publishedAt: true,
          createdAt: true,
          author: { select: { id: true, name: true, avatar: true } },
          category: { select: { id: true, name: true } },
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
    const post = await prisma.post.findUnique({
      where: { id },
      include: {
        author: { select: { id: true, name: true, avatar: true } },
        category: true,
        tags: true,
        comments: {
          where: { parentId: null },
          orderBy: { createdAt: 'desc' },
          include: {
            author: { select: { id: true, name: true, avatar: true } },
            replies: {
              include: {
                author: { select: { id: true, name: true, avatar: true } },
              },
            },
          },
        },
      },
    });

    if (!post) throw new NotFoundError('Post');
    return post;
  }

  async findBySlug(slug: string) {
    const post = await prisma.post.findUnique({
      where: { slug },
      include: {
        author: { select: { id: true, name: true, avatar: true } },
        tags: true,
      },
    });

    if (!post) throw new NotFoundError('Post');
    return post;
  }

  async create(data: {
    title: string;
    content: string;
    authorId: string;
    categoryId?: string;
    tags?: string[];
    published?: boolean;
  }) {
    const slug = this.generateSlug(data.title);

    return prisma.post.create({
      data: {
        title: data.title,
        slug,
        content: data.content,
        excerpt: data.content.slice(0, 200),
        published: data.published ?? false,
        publishedAt: data.published ? new Date() : null,
        authorId: data.authorId,
        categoryId: data.categoryId,
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

  async update(
    id: string,
    data: Partial<{
      title: string;
      content: string;
      categoryId: string;
      tags: string[];
      published: boolean;
    }>,
    userId: string,
  ) {
    const post = await this.findById(id);
    if (post.author.id !== userId) {
      throw new ForbiddenError('You can only edit your own posts');
    }

    const updateData: Prisma.PostUpdateInput = {
      ...(data.title && { title: data.title, slug: this.generateSlug(data.title) }),
      ...(data.content && {
        content: data.content,
        excerpt: data.content.slice(0, 200),
      }),
      ...(data.categoryId && { category: { connect: { id: data.categoryId } } }),
      ...(data.published !== undefined && {
        published: data.published,
        publishedAt: data.published ? new Date() : null,
      }),
    };

    // Handle tags (disconnect all, reconnect)
    if (data.tags) {
      updateData.tags = {
        set: [], // Disconnect all
        connectOrCreate: data.tags.map((name) => ({
          where: { name },
          create: { name },
        })),
      };
    }

    return prisma.post.update({
      where: { id },
      data: updateData,
      include: { tags: true },
    });
  }

  async delete(id: string, userId: string) {
    const post = await this.findById(id);
    if (post.author.id !== userId) {
      throw new ForbiddenError('You can only delete your own posts');
    }

    await prisma.post.delete({ where: { id } });
  }

  private generateSlug(title: string): string {
    return title
      .toLowerCase()
      .replace(/[^a-z0-9]+/g, '-')
      .replace(/^-|-$/g, '');
  }
}
```

## Transactions

```typescript
// Simple transaction (all-or-nothing)
const [post, notification] = await prisma.$transaction([
  prisma.post.create({ data: postData }),
  prisma.notification.create({ data: notifData }),
]);

// Interactive transaction (conditional logic)
const result = await prisma.$transaction(async (tx) => {
  const user = await tx.user.findUnique({ where: { id: userId } });
  if (!user) throw new NotFoundError('User');

  if (user.credits < amount) {
    throw new BadRequestError('Insufficient credits');
  }

  const [updatedUser, purchase] = await Promise.all([
    tx.user.update({
      where: { id: userId },
      data: { credits: { decrement: amount } },
    }),
    tx.purchase.create({
      data: { userId, productId, amount },
    }),
  ]);

  return { updatedUser, purchase };
}, {
  maxWait: 5000,    // Max time to wait for a connection
  timeout: 10000,   // Max time for the transaction
  isolationLevel: Prisma.TransactionIsolationLevel.Serializable,
});
```

## Soft Delete Pattern

```prisma
model Post {
  // ... other fields
  deletedAt DateTime? @map("deleted_at")
}
```

```typescript
// Middleware approach — filter deleted records automatically
const prisma = new PrismaClient().$extends({
  query: {
    post: {
      async findMany({ args, query }) {
        args.where = { ...args.where, deletedAt: null };
        return query(args);
      },
      async findFirst({ args, query }) {
        args.where = { ...args.where, deletedAt: null };
        return query(args);
      },
    },
  },
});

// Soft delete
await prisma.post.update({
  where: { id },
  data: { deletedAt: new Date() },
});

// Hard delete (bypass middleware)
await prisma.$queryRaw`DELETE FROM posts WHERE id = ${id}`;

// Include deleted records
await prisma.post.findMany({
  where: { deletedAt: { not: null } }, // Explicitly include deleted
});
```

## Pagination Helpers

```typescript
// src/lib/pagination.ts
interface PaginationParams {
  page: number;
  limit: number;
}

interface PaginatedResult<T> {
  items: T[];
  page: number;
  limit: number;
  total: number;
  totalPages: number;
  hasNext: boolean;
  hasPrev: boolean;
}

export function paginate<T>(
  items: T[],
  total: number,
  params: PaginationParams,
): PaginatedResult<T> {
  const totalPages = Math.ceil(total / params.limit);

  return {
    items,
    page: params.page,
    limit: params.limit,
    total,
    totalPages,
    hasNext: params.page < totalPages,
    hasPrev: params.page > 1,
  };
}

// Cursor-based pagination (better for infinite scroll)
export async function cursorPaginate<T extends { id: string }>(
  model: any,
  params: { cursor?: string; limit: number; where?: any; orderBy?: any },
) {
  const { cursor, limit, where, orderBy } = params;

  const items = await model.findMany({
    where,
    take: limit + 1, // Take one extra to check for next page
    ...(cursor && { cursor: { id: cursor }, skip: 1 }),
    orderBy: orderBy || { createdAt: 'desc' },
  });

  const hasMore = items.length > limit;
  if (hasMore) items.pop(); // Remove the extra item

  return {
    items,
    nextCursor: hasMore ? items[items.length - 1].id : null,
    hasMore,
  };
}
```

## Seeding

```typescript
// prisma/seed.ts
import { PrismaClient } from '@prisma/client';
import bcrypt from 'bcrypt';

const prisma = new PrismaClient();

async function main() {
  // Create admin user
  const admin = await prisma.user.upsert({
    where: { email: 'admin@example.com' },
    update: {},
    create: {
      email: 'admin@example.com',
      name: 'Admin',
      password: await bcrypt.hash('admin123', 12),
      role: 'ADMIN',
    },
  });

  // Create categories
  const categories = await Promise.all(
    ['Technology', 'Design', 'Business'].map((name) =>
      prisma.category.upsert({
        where: { name },
        update: {},
        create: { name, slug: name.toLowerCase() },
      }),
    ),
  );

  // Create sample posts
  for (let i = 1; i <= 10; i++) {
    await prisma.post.upsert({
      where: { slug: `sample-post-${i}` },
      update: {},
      create: {
        title: `Sample Post ${i}`,
        slug: `sample-post-${i}`,
        content: `Content for sample post ${i}`,
        excerpt: `Excerpt for sample post ${i}`,
        published: i <= 5,
        publishedAt: i <= 5 ? new Date() : null,
        authorId: admin.id,
        categoryId: categories[i % categories.length].id,
      },
    });
  }

  console.log('Seed data created');
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

## Migration Commands

```bash
# Create migration from schema changes
npx prisma migrate dev --name add_posts_table

# Apply migrations in production
npx prisma migrate deploy

# Reset database (dev only — drops and recreates)
npx prisma migrate reset

# Generate client after schema changes
npx prisma generate

# View database in browser
npx prisma studio

# Push schema without migration (prototyping)
npx prisma db push

# Pull schema from existing database
npx prisma db pull
```

## Query Optimization

```typescript
// Select only needed fields (reduces data transfer)
const users = await prisma.user.findMany({
  select: { id: true, name: true, email: true },
});

// Include related data in one query (avoid N+1)
const posts = await prisma.post.findMany({
  include: {
    author: { select: { id: true, name: true } },
    _count: { select: { comments: true } },
  },
});

// Batch operations
await prisma.post.updateMany({
  where: { authorId: userId, published: false },
  data: { published: true, publishedAt: new Date() },
});

// Raw query for complex aggregations
const stats = await prisma.$queryRaw<{ category: string; count: bigint }[]>`
  SELECT c.name as category, COUNT(p.id) as count
  FROM categories c
  LEFT JOIN posts p ON p.category_id = c.id AND p.published = true
  GROUP BY c.name
  ORDER BY count DESC
`;
```

## Connection Management

```typescript
// Graceful disconnect on shutdown
process.on('SIGTERM', async () => {
  await prisma.$disconnect();
  process.exit(0);
});

// Connection pool settings (via DATABASE_URL)
// postgresql://user:pass@host:5432/db?connection_limit=10&pool_timeout=20

// Health check
app.get('/health', async (_req, res) => {
  try {
    await prisma.$queryRaw`SELECT 1`;
    res.json({ status: 'ok', db: 'connected' });
  } catch {
    res.status(503).json({ status: 'error', db: 'disconnected' });
  }
});
```

## Gotchas

1. **Prisma singleton in dev.** Hot module reload creates multiple PrismaClient instances, exhausting database connections. Always use the global singleton pattern shown above.

2. **`findUnique` requires unique fields.** You can't use `findUnique` with non-unique fields. Use `findFirst` instead, but beware it returns the first match, not a guaranteed single result.

3. **`updateMany` and `deleteMany` don't return records.** They return a count `{ count: number }`. If you need the records, use `findMany` first or individual `update`/`delete`.

4. **Implicit many-to-many (Tag↔Post) needs `set: []` to disconnect all.** `update({ tags: { connect: [...] } })` ADDS to existing. To REPLACE: `tags: { set: [], connectOrCreate: [...] }`.

5. **BigInt from raw queries.** PostgreSQL COUNT returns BigInt. JavaScript can't serialize BigInt to JSON. Convert: `Number(result.count)` or use `JSON.stringify` with a replacer.

6. **Transaction timeouts.** Default interactive transaction timeout is 5 seconds. For long operations, increase: `prisma.$transaction(async (tx) => {...}, { timeout: 30000 })`.
