# Drizzle ORM Cheat Sheet

## Setup

```bash
npm install drizzle-orm postgres
npm install -D drizzle-kit
```

```typescript
// drizzle.config.ts
import { defineConfig } from 'drizzle-kit';
export default defineConfig({
  schema: './src/db/schema/*',
  out: './drizzle',
  dialect: 'postgresql',
  dbCredentials: { url: process.env.DATABASE_URL! },
});
```

```typescript
// src/db/index.ts
import { drizzle } from 'drizzle-orm/postgres-js';
import postgres from 'postgres';
import * as schema from './schema';

const connection = postgres(process.env.DATABASE_URL!);
export const db = drizzle(connection, { schema });
```

## Schema Quick Reference

```typescript
import { pgTable, text, varchar, integer, boolean, timestamp, pgEnum, index } from 'drizzle-orm/pg-core';
import { relations } from 'drizzle-orm';

export const roleEnum = pgEnum('role', ['user', 'admin']);

export const users = pgTable('users', {
  id: text('id').primaryKey().$defaultFn(() => crypto.randomUUID()),
  email: varchar('email', { length: 255 }).notNull().unique(),
  name: varchar('name', { length: 100 }).notNull(),
  role: roleEnum('role').notNull().default('user'),
  createdAt: timestamp('created_at', { withTimezone: true }).notNull().defaultNow(),
  updatedAt: timestamp('updated_at', { withTimezone: true }).notNull().defaultNow().$onUpdate(() => new Date()),
}, (t) => [index('users_email_idx').on(t.email)]);

export const usersRelations = relations(users, ({ many }) => ({
  posts: many(posts),
}));
```

## Query Quick Reference

```typescript
import { eq, ne, gt, lt, gte, lte, and, or, not, like, ilike, isNull, isNotNull, inArray, between, desc, asc, count, sum, avg, sql } from 'drizzle-orm';

// SELECT
await db.select().from(users);
await db.select({ id: users.id, name: users.name }).from(users);

// WHERE
await db.select().from(users).where(eq(users.id, '123'));
await db.select().from(posts).where(and(eq(posts.published, true), isNull(posts.deletedAt)));

// ORDER + LIMIT
await db.select().from(posts).orderBy(desc(posts.createdAt)).limit(20).offset(40);

// INSERT
const [user] = await db.insert(users).values({ email: 'a@b.com', name: 'A' }).returning();

// UPSERT
await db.insert(users).values(data).onConflictDoUpdate({ target: users.email, set: { name: data.name } });

// UPDATE
await db.update(users).set({ name: 'New' }).where(eq(users.id, '123'));
await db.update(posts).set({ views: sql`${posts.views} + 1` }).where(eq(posts.id, id));

// DELETE
await db.delete(posts).where(eq(posts.id, '123'));

// JOIN
await db.select({ post: posts, author: users }).from(posts)
  .innerJoin(users, eq(posts.authorId, users.id));

// AGGREGATION
const [{ total }] = await db.select({ total: count() }).from(posts);

// RELATIONAL QUERY (requires schema in drizzle())
const post = await db.query.posts.findFirst({
  where: eq(posts.id, id),
  with: { author: true, comments: { with: { author: true } } },
});

// RAW SQL
const result = await db.execute(sql`SELECT * FROM users WHERE id = ${id}`);
```

## Migration Commands

```bash
npx drizzle-kit push       # Dev: apply schema directly
npx drizzle-kit generate   # Prod: create SQL migration file
npx drizzle-kit migrate    # Prod: apply pending migrations
npx drizzle-kit studio     # Open database GUI
npx drizzle-kit check      # Verify schema matches DB
```

## Type Inference

```typescript
import { InferSelectModel, InferInsertModel } from 'drizzle-orm';

type User = InferSelectModel<typeof users>;       // What you GET
type NewUser = InferInsertModel<typeof users>;     // What you INSERT
type UserUpdate = Partial<NewUser>;                // What you UPDATE
```

## Transaction

```typescript
await db.transaction(async (tx) => {
  const [order] = await tx.insert(orders).values(data).returning();
  await tx.insert(orderItems).values(items.map(i => ({ ...i, orderId: order.id })));
  if (somethingWrong) tx.rollback();
  return order;
});
```

## Drizzle vs Prisma at a Glance

| | Drizzle | Prisma |
|---|---------|--------|
| Schema | TypeScript | `.prisma` DSL |
| Queries | SQL-like | Object API |
| Bundle | ~50KB | ~1.5MB |
| Serverless | Native | Needs proxy |
| Edge/Workers | Yes | Via Accelerate |
| SQL knowledge | Required | Optional |
| Raw SQL | First-class `sql` tag | Second-class `$queryRaw` |

## Common Patterns

| Pattern | Code |
|---------|------|
| Soft delete | `deletedAt: timestamp().`, filter: `.where(isNull(t.deletedAt))` |
| Slug | `slug: varchar(250).unique()`, generate in app |
| Audit | `createdAt: timestamp().defaultNow()`, `updatedAt: timestamp().$onUpdate(() => new Date())` |
| Cursor pagination | `.where(lt(t.createdAt, cursor)).limit(20)` |
| Full-text search | `sql\`to_tsvector('english', ${t.title}) @@ to_tsquery('english', ${query})\`` |
| JSON column | `metadata: jsonb().$type<MyType>()` |
| Array column | `tags: text().array()` |
| Increment | `.set({ count: sql\`${t.count} + 1\` })` |

## Package.json Scripts

```json
{
  "scripts": {
    "db:generate": "drizzle-kit generate",
    "db:migrate": "drizzle-kit migrate",
    "db:push": "drizzle-kit push",
    "db:studio": "drizzle-kit studio",
    "db:seed": "tsx src/db/seed.ts"
  }
}
```
