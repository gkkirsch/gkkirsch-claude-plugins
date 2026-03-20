---
name: drizzle-advanced
description: >
  Advanced Drizzle ORM patterns — transactions, prepared statements, dynamic queries,
  SQL templates, connection pooling, multi-tenant, and performance optimization.
  Triggers: "drizzle transaction", "drizzle prepared statement", "drizzle dynamic query",
  "drizzle sql template", "drizzle performance", "drizzle connection pool", "drizzle multi-tenant".
  NOT for: basic CRUD (use drizzle-queries), schema design (use drizzle-schema).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Advanced Drizzle Patterns

## Transactions

```typescript
// Basic transaction
const result = await db.transaction(async (tx) => {
  const [user] = await tx.insert(users).values({
    email: 'new@example.com',
    name: 'New User',
  }).returning();

  const [profile] = await tx.insert(profiles).values({
    userId: user.id,
    bio: 'Hello!',
  }).returning();

  return { user, profile };
});

// Transaction with rollback on error
const result = await db.transaction(async (tx) => {
  const [from] = await tx
    .update(accounts)
    .set({ balance: sql`${accounts.balance} - ${amount}` })
    .where(eq(accounts.id, fromId))
    .returning();

  if (from.balance < 0) {
    tx.rollback(); // Throws, rolls back entire transaction
  }

  await tx
    .update(accounts)
    .set({ balance: sql`${accounts.balance} + ${amount}` })
    .where(eq(accounts.id, toId));

  return from;
});

// Nested transactions (savepoints)
await db.transaction(async (tx) => {
  await tx.insert(orders).values(orderData);

  try {
    await tx.transaction(async (nestedTx) => {
      await nestedTx.insert(payments).values(paymentData);
      // If this fails, only the nested transaction rolls back
    });
  } catch (err) {
    // Payment failed, but order is still saved
    await tx.update(orders)
      .set({ status: 'payment_failed' })
      .where(eq(orders.id, orderData.id));
  }
});

// Transaction config
await db.transaction(async (tx) => {
  // ...
}, {
  isolationLevel: 'serializable',   // 'read committed' | 'repeatable read' | 'serializable'
  accessMode: 'read write',         // 'read only' | 'read write'
  deferrable: true,                 // only for serializable + read only
});
```

## Prepared Statements

```typescript
// Prepare a reusable query
const getUserById = db
  .select()
  .from(users)
  .where(eq(users.id, sql.placeholder('id')))
  .prepare('get_user_by_id');

// Execute with parameters
const user = await getUserById.execute({ id: '123' });

// Prepared query with multiple placeholders
const searchPosts = db
  .select()
  .from(posts)
  .where(
    and(
      eq(posts.published, sql.placeholder('published')),
      ilike(posts.title, sql.placeholder('search')),
    ),
  )
  .orderBy(desc(posts.createdAt))
  .limit(sql.placeholder('limit'))
  .offset(sql.placeholder('offset'))
  .prepare('search_posts');

const results = await searchPosts.execute({
  published: true,
  search: '%typescript%',
  limit: 20,
  offset: 0,
});

// Prepared insert
const createUser = db
  .insert(users)
  .values({
    email: sql.placeholder('email'),
    name: sql.placeholder('name'),
    role: sql.placeholder('role'),
  })
  .returning()
  .prepare('create_user');

const [newUser] = await createUser.execute({
  email: 'new@example.com',
  name: 'New User',
  role: 'user',
});
```

## Dynamic Query Building

```typescript
// Build queries conditionally
function buildPostsQuery(filters: {
  search?: string;
  authorId?: string;
  published?: boolean;
  categoryId?: string;
  tags?: string[];
  page?: number;
  limit?: number;
  sort?: 'newest' | 'oldest' | 'popular';
}) {
  let query = db
    .select({
      post: posts,
      authorName: users.name,
      commentCount: count(comments.id),
    })
    .from(posts)
    .innerJoin(users, eq(posts.authorId, users.id))
    .leftJoin(comments, eq(posts.id, comments.postId))
    .$dynamic();

  // Apply filters
  const conditions: SQL[] = [isNull(posts.deletedAt)];

  if (filters.search) {
    conditions.push(
      or(
        ilike(posts.title, `%${filters.search}%`),
        ilike(posts.content, `%${filters.search}%`),
      )!,
    );
  }

  if (filters.authorId) {
    conditions.push(eq(posts.authorId, filters.authorId));
  }

  if (filters.published !== undefined) {
    conditions.push(eq(posts.published, filters.published));
  }

  if (filters.categoryId) {
    conditions.push(eq(posts.categoryId, filters.categoryId));
  }

  query = query.where(and(...conditions));

  // Group by
  query = query.groupBy(posts.id, users.name);

  // Sorting
  switch (filters.sort) {
    case 'oldest':
      query = query.orderBy(asc(posts.createdAt));
      break;
    case 'popular':
      query = query.orderBy(desc(count(comments.id)));
      break;
    default:
      query = query.orderBy(desc(posts.createdAt));
  }

  // Pagination
  const limit = filters.limit ?? 20;
  const page = filters.page ?? 1;
  query = query.limit(limit).offset((page - 1) * limit);

  return query;
}

// Usage
const results = await buildPostsQuery({
  search: 'drizzle',
  published: true,
  sort: 'newest',
  page: 2,
  limit: 10,
});
```

## SQL Template Tag

```typescript
import { sql } from 'drizzle-orm';

// Type-safe raw SQL with automatic parameterization
const userId = '123';
const result = await db.execute(sql`
  SELECT u.name, COUNT(p.id) as post_count
  FROM ${users} u
  LEFT JOIN ${posts} p ON u.id = p.author_id
  WHERE u.id = ${userId}
  GROUP BY u.id, u.name
`);

// SQL fragments in queries
const fullName = sql<string>`${users.firstName} || ' ' || ${users.lastName}`;
await db.select({
  id: users.id,
  fullName,
}).from(users);

// Conditional SQL
const orderClause = sortDir === 'asc'
  ? sql`ORDER BY ${posts.createdAt} ASC`
  : sql`ORDER BY ${posts.createdAt} DESC`;

// JSON aggregation (PostgreSQL)
const postsWithTags = await db.execute(sql`
  SELECT
    p.*,
    COALESCE(
      json_agg(json_build_object('id', t.id, 'name', t.name))
      FILTER (WHERE t.id IS NOT NULL),
      '[]'
    ) as tags
  FROM ${posts} p
  LEFT JOIN ${postsToTags} pt ON p.id = pt.post_id
  LEFT JOIN ${tags} t ON pt.tag_id = t.id
  WHERE p.published = true
  GROUP BY p.id
  ORDER BY p.created_at DESC
`);

// Type the result of raw queries
const stats = await db.execute<{
  month: string;
  count: number;
}>(sql`
  SELECT
    TO_CHAR(created_at, 'YYYY-MM') as month,
    COUNT(*)::integer as count
  FROM ${posts}
  GROUP BY month
  ORDER BY month DESC
`);
```

## Connection Pooling

```typescript
// postgres.js with pooling
import postgres from 'postgres';

const connection = postgres(process.env.DATABASE_URL!, {
  max: 10,                    // Max pool size
  idle_timeout: 20,           // Seconds before idle connection is closed
  connect_timeout: 10,        // Seconds to wait for connection
  max_lifetime: 60 * 30,     // Max connection lifetime (30 min)
  prepare: true,              // Use prepared statements
});

// For serverless (new connection per request)
import { neon } from '@neondatabase/serverless';
const sql = neon(process.env.DATABASE_URL!);
export const db = drizzle(sql, { schema });

// For serverless with connection pooling (Neon)
import { Pool, neonConfig } from '@neondatabase/serverless';
import ws from 'ws';
neonConfig.webSocketConstructor = ws;

const pool = new Pool({ connectionString: process.env.DATABASE_URL });
export const db = drizzle(pool, { schema });

// Graceful shutdown
process.on('SIGTERM', async () => {
  await connection.end();
  process.exit(0);
});
```

## Multi-Tenant Patterns

### Row-Level Isolation

```typescript
// Every query scoped to tenant
export function createTenantDb(tenantId: string) {
  return {
    posts: {
      list: () =>
        db.select().from(posts)
          .where(and(
            eq(posts.tenantId, tenantId),
            isNull(posts.deletedAt),
          )),

      create: (data: NewPost) =>
        db.insert(posts).values({ ...data, tenantId }).returning(),

      findById: (id: string) =>
        db.select().from(posts)
          .where(and(
            eq(posts.id, id),
            eq(posts.tenantId, tenantId),
          )),
    },
  };
}

// Usage in middleware
app.use((req, res, next) => {
  const tenantId = req.headers['x-tenant-id'] as string;
  req.tenantDb = createTenantDb(tenantId);
  next();
});
```

### PostgreSQL Row-Level Security

```sql
-- Enable RLS on table
ALTER TABLE posts ENABLE ROW LEVEL SECURITY;

-- Policy: users can only see their tenant's posts
CREATE POLICY tenant_isolation ON posts
  USING (tenant_id = current_setting('app.tenant_id'));

-- Set tenant in session
SET app.tenant_id = 'tenant-123';
```

```typescript
// Set tenant context before queries
async function withTenant<T>(tenantId: string, fn: () => Promise<T>): Promise<T> {
  await db.execute(sql`SET LOCAL app.tenant_id = ${tenantId}`);
  return fn();
}
```

## Batch Operations

```typescript
// Batch insert
const BATCH_SIZE = 1000;
const allRecords = [...]; // 10,000 records

for (let i = 0; i < allRecords.length; i += BATCH_SIZE) {
  const batch = allRecords.slice(i, i + BATCH_SIZE);
  await db.insert(records).values(batch);
}

// Batch update with CASE
await db.execute(sql`
  UPDATE ${posts}
  SET status = CASE
    WHEN created_at < ${thirtyDaysAgo} THEN 'archived'
    WHEN published = false THEN 'draft'
    ELSE status
  END
  WHERE author_id = ${userId}
`);

// Bulk upsert
await db.insert(tags)
  .values(tagNames.map(name => ({ name })))
  .onConflictDoNothing({ target: tags.name });
```

## Query Logging and Debugging

```typescript
// Enable query logging
const db = drizzle(connection, {
  schema,
  logger: true, // Logs all queries to console
});

// Custom logger
const db = drizzle(connection, {
  schema,
  logger: {
    logQuery(query: string, params: unknown[]) {
      console.log({ query, params });
      // Or send to your logging service
    },
  },
});

// Per-query timing
async function timedQuery<T>(name: string, fn: () => Promise<T>): Promise<T> {
  const start = Date.now();
  const result = await fn();
  const duration = Date.now() - start;
  if (duration > 500) {
    console.warn(`Slow query "${name}": ${duration}ms`);
  }
  return result;
}

const user = await timedQuery('getUserById', () =>
  db.query.users.findFirst({ where: eq(users.id, id) })
);
```

## Testing Patterns

```typescript
// Test helper: clean database between tests
export async function cleanDb() {
  await db.execute(sql`
    DO $$ DECLARE
      r RECORD;
    BEGIN
      FOR r IN (SELECT tablename FROM pg_tables WHERE schemaname = 'public') LOOP
        IF r.tablename != '__drizzle_migrations' THEN
          EXECUTE 'TRUNCATE TABLE ' || quote_ident(r.tablename) || ' CASCADE';
        END IF;
      END LOOP;
    END $$;
  `);
}

// Test helper: factory functions
export function createTestUser(overrides?: Partial<NewUser>) {
  return db.insert(users).values({
    email: `test-${Date.now()}@example.com`,
    name: 'Test User',
    role: 'user',
    ...overrides,
  }).returning();
}

// Integration test example
describe('PostService', () => {
  beforeEach(() => cleanDb());

  it('creates a post with tags', async () => {
    const [user] = await createTestUser();
    const service = new PostService();

    const post = await service.create({
      title: 'Test Post',
      content: 'Content',
      authorId: user.id,
      tags: ['typescript', 'drizzle'],
    });

    expect(post.title).toBe('Test Post');

    const tags = await db.select().from(postsToTags)
      .where(eq(postsToTags.postId, post.id));
    expect(tags).toHaveLength(2);
  });
});
```

## Gotchas

1. **`$dynamic()` is required for conditional query building.** Without it, TypeScript won't let you reassign the query variable. Call `.$dynamic()` after the initial `from()`.

2. **Prepared statements are cached per connection.** In connection pools, the same prepared statement name may conflict across connections. Drizzle handles this internally, but be aware when using raw `postgres.js` prepared statements alongside Drizzle.

3. **`tx.rollback()` throws an error.** It's implemented as a thrown exception that Drizzle catches. Don't wrap it in try/catch inside the transaction callback — let it propagate.

4. **Connection pool exhaustion in serverless.** Each Lambda/Worker invocation creates connections. Use Neon's pooler (`-pooler` suffix in URL) or PgBouncer. Set `max: 1` for serverless.

5. **`sql` template tag auto-parameterizes.** Values interpolated with `${value}` become query parameters (safe from injection). Table/column references like `${posts}` become identifiers. Never use string concatenation for user input.

6. **Multi-tenant RLS needs `SET LOCAL`.** Use `SET LOCAL` (not `SET`) so the tenant context is scoped to the current transaction, not the entire connection.

7. **Batch inserts have a PostgreSQL parameter limit.** PostgreSQL supports max ~65,535 parameters per query. If each row has 10 columns, max batch size is ~6,500 rows. Split larger batches.
