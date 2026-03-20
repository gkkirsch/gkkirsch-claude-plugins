---
name: drizzle-queries
description: >
  Drizzle ORM query patterns — select, insert, update, delete, joins,
  aggregations, subqueries, the relational query API, and type-safe filters.
  Triggers: "drizzle query", "drizzle select", "drizzle insert", "drizzle update",
  "drizzle delete", "drizzle join", "drizzle where", "drizzle find", "drizzle with".
  NOT for: schema definition (use drizzle-schema), migrations (use drizzle-migrations).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Drizzle Query Patterns

## Select Queries

```typescript
import { db } from './db';
import { users, posts } from './db/schema';
import { eq, ne, gt, gte, lt, lte, like, ilike, and, or, not, isNull, isNotNull, inArray, notInArray, between, sql, desc, asc, count, sum, avg } from 'drizzle-orm';

// Select all columns
const allUsers = await db.select().from(users);

// Select specific columns
const userEmails = await db
  .select({ id: users.id, email: users.email })
  .from(users);

// Aliased columns
const result = await db
  .select({
    userName: users.name,
    postTitle: posts.title,
  })
  .from(users)
  .innerJoin(posts, eq(users.id, posts.authorId));
```

## Where Conditions

```typescript
// Equality
await db.select().from(users).where(eq(users.email, 'test@example.com'));

// Not equal
await db.select().from(users).where(ne(users.role, 'admin'));

// Comparison
await db.select().from(posts).where(gt(posts.createdAt, oneWeekAgo));
await db.select().from(posts).where(gte(posts.views, 100));
await db.select().from(posts).where(lt(posts.createdAt, cutoffDate));

// Like / ILike (case-insensitive)
await db.select().from(users).where(like(users.name, '%john%'));
await db.select().from(users).where(ilike(users.email, '%@gmail.com'));

// NULL checks
await db.select().from(posts).where(isNull(posts.deletedAt));
await db.select().from(posts).where(isNotNull(posts.publishedAt));

// IN / NOT IN
await db.select().from(users).where(inArray(users.role, ['admin', 'moderator']));
await db.select().from(posts).where(notInArray(posts.status, ['draft', 'archived']));

// BETWEEN
await db.select().from(posts).where(between(posts.createdAt, startDate, endDate));

// AND / OR / NOT
await db.select().from(posts).where(
  and(
    eq(posts.published, true),
    eq(posts.authorId, userId),
  ),
);

await db.select().from(users).where(
  or(
    eq(users.role, 'admin'),
    eq(users.role, 'moderator'),
  ),
);

await db.select().from(posts).where(
  not(eq(posts.status, 'archived')),
);

// Complex conditions
await db.select().from(posts).where(
  and(
    eq(posts.published, true),
    or(
      ilike(posts.title, `%${search}%`),
      ilike(posts.content, `%${search}%`),
    ),
    isNull(posts.deletedAt),
  ),
);
```

## Ordering and Pagination

```typescript
// Order by
await db.select().from(posts)
  .orderBy(desc(posts.createdAt));

// Multiple order columns
await db.select().from(posts)
  .orderBy(desc(posts.published), asc(posts.title));

// Offset pagination
await db.select().from(posts)
  .where(eq(posts.published, true))
  .orderBy(desc(posts.createdAt))
  .limit(20)
  .offset(40); // Page 3

// Cursor pagination (better for large datasets)
await db.select().from(posts)
  .where(
    and(
      eq(posts.published, true),
      lt(posts.createdAt, cursorDate),
    ),
  )
  .orderBy(desc(posts.createdAt))
  .limit(20);
```

## Insert

```typescript
// Single insert
const [newUser] = await db.insert(users).values({
  email: 'john@example.com',
  name: 'John',
  role: 'user',
}).returning();

// Multi-row insert
await db.insert(posts).values([
  { title: 'Post 1', content: 'Content 1', authorId: userId },
  { title: 'Post 2', content: 'Content 2', authorId: userId },
]);

// Insert with returning
const [created] = await db.insert(users).values({
  email: 'new@example.com',
  name: 'New User',
}).returning({
  id: users.id,
  email: users.email,
});

// Upsert (insert or update on conflict)
await db.insert(users).values({
  email: 'john@example.com',
  name: 'John Updated',
}).onConflictDoUpdate({
  target: users.email,
  set: { name: sql`excluded.name`, updatedAt: new Date() },
});

// Insert ignore (skip on conflict)
await db.insert(tags).values({ name: 'typescript' })
  .onConflictDoNothing({ target: tags.name });
```

## Update

```typescript
// Update with where
await db.update(users)
  .set({ name: 'New Name', updatedAt: new Date() })
  .where(eq(users.id, userId));

// Update with returning
const [updated] = await db.update(posts)
  .set({ published: true, publishedAt: new Date() })
  .where(eq(posts.id, postId))
  .returning();

// Increment / decrement
await db.update(posts)
  .set({ views: sql`${posts.views} + 1` })
  .where(eq(posts.id, postId));

// Conditional update
await db.update(posts)
  .set({
    status: 'archived',
    archivedAt: new Date(),
  })
  .where(
    and(
      eq(posts.published, false),
      lt(posts.createdAt, thirtyDaysAgo),
    ),
  );
```

## Delete

```typescript
// Delete with where
await db.delete(posts).where(eq(posts.id, postId));

// Delete with returning
const [deleted] = await db.delete(posts)
  .where(eq(posts.id, postId))
  .returning({ id: posts.id, title: posts.title });

// Bulk delete
await db.delete(sessions)
  .where(lt(sessions.expiresAt, new Date()));

// Soft delete pattern
await db.update(posts)
  .set({ deletedAt: new Date() })
  .where(eq(posts.id, postId));
```

## Joins

```typescript
// Inner join
const postsWithAuthors = await db
  .select({
    post: posts,
    author: { id: users.id, name: users.name },
  })
  .from(posts)
  .innerJoin(users, eq(posts.authorId, users.id))
  .where(eq(posts.published, true));

// Left join
const usersWithPosts = await db
  .select({
    user: users,
    postCount: count(posts.id),
  })
  .from(users)
  .leftJoin(posts, eq(users.id, posts.authorId))
  .groupBy(users.id);

// Multiple joins
const postDetails = await db
  .select({
    title: posts.title,
    authorName: users.name,
    categoryName: categories.name,
    commentCount: count(comments.id),
  })
  .from(posts)
  .innerJoin(users, eq(posts.authorId, users.id))
  .leftJoin(categories, eq(posts.categoryId, categories.id))
  .leftJoin(comments, eq(posts.id, comments.postId))
  .where(eq(posts.published, true))
  .groupBy(posts.id, users.name, categories.name)
  .orderBy(desc(posts.createdAt));

// Many-to-many through junction table
const postsWithTags = await db
  .select({
    post: posts,
    tagName: tags.name,
  })
  .from(posts)
  .innerJoin(postsToTags, eq(posts.id, postsToTags.postId))
  .innerJoin(tags, eq(postsToTags.tagId, tags.id));
```

## Relational Query API (with)

```typescript
// findFirst with relations (requires relations defined in schema)
const post = await db.query.posts.findFirst({
  where: eq(posts.id, postId),
  with: {
    author: true,
    comments: {
      with: { author: true },
      orderBy: [desc(comments.createdAt)],
      limit: 10,
    },
    tags: {
      with: { tag: true },
    },
  },
});

// findMany with filtering
const publishedPosts = await db.query.posts.findMany({
  where: and(
    eq(posts.published, true),
    isNull(posts.deletedAt),
  ),
  with: {
    author: {
      columns: { id: true, name: true, avatar: true },
    },
  },
  orderBy: [desc(posts.createdAt)],
  limit: 20,
  offset: 0,
});

// Select specific columns in relational queries
const users = await db.query.users.findMany({
  columns: {
    id: true,
    name: true,
    email: true,
    // password NOT included
  },
  with: {
    posts: {
      columns: { id: true, title: true },
      where: eq(posts.published, true),
    },
  },
});
```

## Aggregations

```typescript
// Count
const [{ total }] = await db
  .select({ total: count() })
  .from(posts)
  .where(eq(posts.published, true));

// Count with group by
const postsByAuthor = await db
  .select({
    authorId: posts.authorId,
    authorName: users.name,
    postCount: count(posts.id),
  })
  .from(posts)
  .innerJoin(users, eq(posts.authorId, users.id))
  .groupBy(posts.authorId, users.name)
  .orderBy(desc(count(posts.id)));

// Sum, Average
const [stats] = await db
  .select({
    totalViews: sum(posts.views),
    avgViews: avg(posts.views),
    postCount: count(),
  })
  .from(posts)
  .where(eq(posts.authorId, userId));

// Having clause
const activeAuthors = await db
  .select({
    authorId: posts.authorId,
    postCount: count(posts.id),
  })
  .from(posts)
  .groupBy(posts.authorId)
  .having(gt(count(posts.id), 5));
```

## Subqueries

```typescript
// Subquery in WHERE
const subquery = db
  .select({ authorId: posts.authorId })
  .from(posts)
  .where(eq(posts.published, true))
  .groupBy(posts.authorId)
  .having(gt(count(), 5));

const prolificAuthors = await db
  .select()
  .from(users)
  .where(inArray(users.id, subquery));

// Subquery as column
const sq = db.$with('sq').as(
  db.select({
    authorId: posts.authorId,
    postCount: count().as('post_count'),
  })
  .from(posts)
  .groupBy(posts.authorId)
);

const result = await db
  .with(sq)
  .select({
    name: users.name,
    postCount: sq.postCount,
  })
  .from(users)
  .leftJoin(sq, eq(users.id, sq.authorId));
```

## Raw SQL

```typescript
// sql template tag
const result = await db.execute(sql`
  SELECT u.name, COUNT(p.id) as post_count
  FROM users u
  LEFT JOIN posts p ON u.id = p.author_id
  WHERE p.published = true
  GROUP BY u.id, u.name
  ORDER BY post_count DESC
  LIMIT ${limit}
`);

// sql in where clauses
await db.select().from(posts).where(
  sql`${posts.title} ILIKE ${'%' + search + '%'}`,
);

// sql for computed columns
await db.select({
  ...posts,
  daysSincePublished: sql<number>`
    EXTRACT(DAY FROM NOW() - ${posts.publishedAt})
  `.as('days_since_published'),
}).from(posts);
```

## CRUD Service Pattern

```typescript
// src/services/posts.service.ts
export class PostService {
  async list(opts: { page: number; limit: number; search?: string; userId?: string }) {
    const { page, limit, search, userId } = opts;
    const conditions = [isNull(posts.deletedAt), eq(posts.published, true)];

    if (userId) conditions.push(eq(posts.authorId, userId));
    if (search) conditions.push(ilike(posts.title, `%${search}%`));

    const [items, [{ total }]] = await Promise.all([
      db.query.posts.findMany({
        where: and(...conditions),
        with: {
          author: { columns: { id: true, name: true, avatar: true } },
          tags: { with: { tag: true } },
        },
        orderBy: [desc(posts.createdAt)],
        limit,
        offset: (page - 1) * limit,
      }),
      db.select({ total: count() }).from(posts).where(and(...conditions)),
    ]);

    return { items, page, limit, total, totalPages: Math.ceil(total / limit) };
  }

  async findById(id: string) {
    return db.query.posts.findFirst({
      where: and(eq(posts.id, id), isNull(posts.deletedAt)),
      with: {
        author: { columns: { id: true, name: true, avatar: true } },
        comments: {
          where: isNull(comments.parentId),
          with: { author: true, replies: { with: { author: true } } },
          orderBy: [desc(comments.createdAt)],
        },
        tags: { with: { tag: true } },
      },
    });
  }

  async create(data: { title: string; content: string; authorId: string; tags?: string[] }) {
    return db.transaction(async (tx) => {
      const [post] = await tx.insert(posts).values({
        title: data.title,
        slug: this.slugify(data.title),
        content: data.content,
        authorId: data.authorId,
      }).returning();

      if (data.tags?.length) {
        for (const tagName of data.tags) {
          const [tag] = await tx.insert(tags).values({ name: tagName })
            .onConflictDoNothing({ target: tags.name })
            .returning();

          const existingTag = tag || await tx.query.tags.findFirst({
            where: eq(tags.name, tagName),
          });

          if (existingTag) {
            await tx.insert(postsToTags).values({
              postId: post.id,
              tagId: existingTag.id,
            });
          }
        }
      }

      return post;
    });
  }

  private slugify(title: string): string {
    return title.toLowerCase().replace(/[^\w\s-]/g, '').replace(/\s+/g, '-').slice(0, 100);
  }
}
```

## Gotchas

1. **`select()` without `from()` is a type error.** Unlike Prisma's `findMany()`, Drizzle requires explicit `select().from(table)`. The SQL-like API means SQL-like structure.

2. **Relational queries need `schema` passed to `drizzle()`.** The `db.query.*` API only works if you passed `{ schema }` when creating the client. Without it, only the SQL-like API (`select/insert/update/delete`) works.

3. **`count()` returns a string.** PostgreSQL returns `bigint` for count. Drizzle preserves this as a string. Parse with `Number()`: `Number(result.total)`.

4. **`.returning()` is PostgreSQL/SQLite only.** MySQL doesn't support RETURNING. Use `insertId` from the result instead.

5. **Dynamic conditions need array spreading.** Build conditions as an array and spread into `and()`: `and(...conditions)`. An empty conditions array returns `undefined` (matches all).

6. **The `with` API doesn't generate JOINs.** It executes separate queries and merges in JS. For performance-critical code, use explicit JOINs instead.

7. **`eq(column, null)` doesn't work.** Use `isNull(column)` instead. SQL `= NULL` is always false — you must use `IS NULL`.
