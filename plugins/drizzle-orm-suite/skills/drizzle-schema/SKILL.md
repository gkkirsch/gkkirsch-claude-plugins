---
name: drizzle-schema
description: >
  Drizzle ORM schema definition — tables, columns, types, relations, indexes,
  enums, composite types, and multi-file schema organization.
  Triggers: "drizzle schema", "drizzle table", "drizzle column", "drizzle relation",
  "drizzle index", "drizzle enum", "drizzle pgTable", "drizzle define table".
  NOT for: queries (use drizzle-queries), migrations (use drizzle-migrations).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Drizzle Schema Definition

## Setup

```bash
npm install drizzle-orm postgres    # PostgreSQL
npm install drizzle-orm mysql2      # MySQL
npm install drizzle-orm better-sqlite3  # SQLite
npm install -D drizzle-kit          # CLI tools
```

## PostgreSQL Schema

```typescript
// src/db/schema/users.ts
import {
  pgTable,
  text,
  varchar,
  timestamp,
  boolean,
  integer,
  serial,
  uuid,
  pgEnum,
  index,
  uniqueIndex,
} from 'drizzle-orm/pg-core';
import { relations } from 'drizzle-orm';

// Enum
export const roleEnum = pgEnum('role', ['user', 'admin', 'moderator']);

// Users table
export const users = pgTable('users', {
  id: text('id').primaryKey().$defaultFn(() => crypto.randomUUID()),
  email: varchar('email', { length: 255 }).notNull().unique(),
  name: varchar('name', { length: 100 }).notNull(),
  password: text('password'),
  role: roleEnum('role').notNull().default('user'),
  avatar: text('avatar'),
  googleId: text('google_id').unique(),
  emailVerified: boolean('email_verified').notNull().default(false),
  createdAt: timestamp('created_at', { withTimezone: true }).notNull().defaultNow(),
  updatedAt: timestamp('updated_at', { withTimezone: true }).notNull().defaultNow().$onUpdate(() => new Date()),
}, (table) => [
  index('users_email_idx').on(table.email),
  index('users_google_id_idx').on(table.googleId),
]);

// User relations
export const usersRelations = relations(users, ({ many }) => ({
  posts: many(posts),
  comments: many(comments),
}));
```

```typescript
// src/db/schema/posts.ts
import {
  pgTable,
  text,
  varchar,
  timestamp,
  boolean,
  integer,
  index,
  uniqueIndex,
} from 'drizzle-orm/pg-core';
import { relations } from 'drizzle-orm';
import { users } from './users';

export const posts = pgTable('posts', {
  id: text('id').primaryKey().$defaultFn(() => crypto.randomUUID()),
  title: varchar('title', { length: 200 }).notNull(),
  slug: varchar('slug', { length: 250 }).notNull().unique(),
  content: text('content').notNull(),
  excerpt: text('excerpt'),
  published: boolean('published').notNull().default(false),
  authorId: text('author_id').notNull().references(() => users.id, { onDelete: 'cascade' }),
  categoryId: text('category_id').references(() => categories.id, { onDelete: 'set null' }),
  publishedAt: timestamp('published_at', { withTimezone: true }),
  createdAt: timestamp('created_at', { withTimezone: true }).notNull().defaultNow(),
  updatedAt: timestamp('updated_at', { withTimezone: true }).notNull().defaultNow().$onUpdate(() => new Date()),
}, (table) => [
  index('posts_author_id_idx').on(table.authorId),
  index('posts_slug_idx').on(table.slug),
  index('posts_published_idx').on(table.published, table.createdAt),
]);

export const postsRelations = relations(posts, ({ one, many }) => ({
  author: one(users, { fields: [posts.authorId], references: [users.id] }),
  category: one(categories, { fields: [posts.categoryId], references: [categories.id] }),
  comments: many(comments),
  tags: many(postsToTags),
}));
```

## Column Types Reference (PostgreSQL)

```typescript
import {
  // Text
  text,           // text — unlimited length
  varchar,        // varchar(n) — limited length
  char,           // char(n) — fixed length

  // Numbers
  integer,        // integer (4 bytes)
  smallint,       // smallint (2 bytes)
  bigint,         // bigint (8 bytes) — returns string in JS
  serial,         // auto-increment integer
  bigserial,      // auto-increment bigint
  real,           // float4
  doublePrecision,// float8
  numeric,        // numeric(precision, scale)

  // Boolean
  boolean,        // boolean

  // Date/Time
  timestamp,      // timestamp (with or without timezone)
  date,           // date only
  time,           // time only
  interval,       // time interval

  // JSON
  json,           // json (stored as text)
  jsonb,          // jsonb (binary, indexable, preferred)

  // Special
  uuid,           // uuid type
  pgEnum,         // custom enum
  cidr,           // IP network
  inet,           // IP address
  macaddr,        // MAC address
} from 'drizzle-orm/pg-core';

// Common column patterns
const columns = {
  // Primary key options
  id: text('id').primaryKey().$defaultFn(() => crypto.randomUUID()),  // UUID
  id: serial('id').primaryKey(),                                       // Auto-increment
  id: uuid('id').primaryKey().defaultRandom(),                        // PG uuid_generate_v4()

  // Timestamps
  createdAt: timestamp('created_at', { withTimezone: true }).notNull().defaultNow(),
  updatedAt: timestamp('updated_at', { withTimezone: true }).notNull().defaultNow().$onUpdate(() => new Date()),
  deletedAt: timestamp('deleted_at', { withTimezone: true }),

  // JSON column with type
  metadata: jsonb('metadata').$type<{ theme: string; language: string }>(),
  settings: jsonb('settings').$type<Record<string, unknown>>().default({}),

  // Array column (PostgreSQL only)
  tags: text('tags').array(),

  // Not null with default
  status: varchar('status', { length: 20 }).notNull().default('active'),
  count: integer('count').notNull().default(0),
};
```

## Many-to-Many Relations

```typescript
// src/db/schema/tags.ts
export const tags = pgTable('tags', {
  id: text('id').primaryKey().$defaultFn(() => crypto.randomUUID()),
  name: varchar('name', { length: 50 }).notNull().unique(),
});

// Junction table
export const postsToTags = pgTable('posts_to_tags', {
  postId: text('post_id').notNull().references(() => posts.id, { onDelete: 'cascade' }),
  tagId: text('tag_id').notNull().references(() => tags.id, { onDelete: 'cascade' }),
}, (table) => [
  // Composite primary key
  { primaryKey: { columns: [table.postId, table.tagId] } },
  index('posts_to_tags_post_idx').on(table.postId),
  index('posts_to_tags_tag_idx').on(table.tagId),
]);

export const postsToTagsRelations = relations(postsToTags, ({ one }) => ({
  post: one(posts, { fields: [postsToTags.postId], references: [posts.id] }),
  tag: one(tags, { fields: [postsToTags.tagId], references: [tags.id] }),
}));

export const tagsRelations = relations(tags, ({ many }) => ({
  posts: many(postsToTags),
}));
```

## Self-Referential Relations

```typescript
// Comments with replies
export const comments = pgTable('comments', {
  id: text('id').primaryKey().$defaultFn(() => crypto.randomUUID()),
  content: text('content').notNull(),
  authorId: text('author_id').notNull().references(() => users.id, { onDelete: 'cascade' }),
  postId: text('post_id').notNull().references(() => posts.id, { onDelete: 'cascade' }),
  parentId: text('parent_id').references((): any => comments.id, { onDelete: 'cascade' }),
  createdAt: timestamp('created_at', { withTimezone: true }).notNull().defaultNow(),
}, (table) => [
  index('comments_post_idx').on(table.postId),
  index('comments_author_idx').on(table.authorId),
  index('comments_parent_idx').on(table.parentId),
]);

export const commentsRelations = relations(comments, ({ one, many }) => ({
  author: one(users, { fields: [comments.authorId], references: [users.id] }),
  post: one(posts, { fields: [comments.postId], references: [posts.id] }),
  parent: one(comments, {
    fields: [comments.parentId],
    references: [comments.id],
    relationName: 'replies',
  }),
  replies: many(comments, { relationName: 'replies' }),
}));
```

## Multi-File Schema Organization

```typescript
// src/db/schema/index.ts — re-export everything
export * from './users';
export * from './posts';
export * from './tags';
export * from './comments';
export * from './categories';
```

```typescript
// drizzle.config.ts
import { defineConfig } from 'drizzle-kit';

export default defineConfig({
  schema: './src/db/schema/*',     // glob for multi-file
  out: './drizzle',                // migration output
  dialect: 'postgresql',
  dbCredentials: {
    url: process.env.DATABASE_URL!,
  },
});
```

## Database Client Setup

```typescript
// src/db/index.ts
import { drizzle } from 'drizzle-orm/postgres-js';
import postgres from 'postgres';
import * as schema from './schema';

const connection = postgres(process.env.DATABASE_URL!, {
  max: 10,                    // connection pool size
  idle_timeout: 20,           // close idle connections after 20s
  connect_timeout: 10,        // connection timeout
});

export const db = drizzle(connection, { schema });

// For serverless (Neon)
import { drizzle } from 'drizzle-orm/neon-http';
import { neon } from '@neondatabase/serverless';

const sql = neon(process.env.DATABASE_URL!);
export const db = drizzle(sql, { schema });

// For Cloudflare D1
import { drizzle } from 'drizzle-orm/d1';

export const db = drizzle(env.DB, { schema });
```

## Type Inference

```typescript
// Infer types from schema (instead of writing interfaces manually)
import { InferSelectModel, InferInsertModel } from 'drizzle-orm';

// Select type (what you GET from queries)
export type User = InferSelectModel<typeof users>;
// { id: string; email: string; name: string; role: 'user' | 'admin'; ... }

// Insert type (what you PASS to insert)
export type NewUser = InferInsertModel<typeof users>;
// { id?: string; email: string; name: string; role?: 'user' | 'admin'; ... }

// Partial update type
export type UserUpdate = Partial<InferInsertModel<typeof users>>;
```

## Gotchas

1. **Relations are separate from foreign keys.** `references()` creates a DB-level foreign key. `relations()` defines the ORM relation for the `with` query API. You need BOTH — FK for data integrity, relation for query convenience.

2. **`$defaultFn` runs in JavaScript, not SQL.** `$defaultFn(() => crypto.randomUUID())` generates the UUID in your app, not in PostgreSQL. For DB-level defaults, use `.default(sql`gen_random_uuid()`)`.

3. **`bigint` returns strings.** PostgreSQL `bigint` exceeds JavaScript's Number.MAX_SAFE_INTEGER. Drizzle returns it as a string. Parse with `BigInt()` if you need arithmetic.

4. **Schema file must be importable.** Drizzle Kit imports your schema files at build time. Don't use runtime-only values (env vars, async imports) in table definitions.

5. **Index names must be unique per database.** Even across tables. Use descriptive names: `users_email_idx`, not just `email_idx`.

6. **`$onUpdate` is app-level only.** It runs in JavaScript on each update query. For database-level auto-update timestamps, create a PostgreSQL trigger instead.

7. **Circular references in self-referential tables.** TypeScript can't infer the type when a column references its own table. Use `(): any =>` to break the cycle: `references((): any => comments.id)`.
