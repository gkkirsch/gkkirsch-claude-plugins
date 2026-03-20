---
name: drizzle-migrations
description: >
  Drizzle Kit migrations — drizzle.config.ts setup, generate, push, migrate,
  custom migrations, seeding, and zero-downtime migration patterns.
  Triggers: "drizzle migration", "drizzle-kit", "drizzle push", "drizzle generate",
  "drizzle config", "drizzle seed", "database migration drizzle".
  NOT for: schema definition (use drizzle-schema), queries (use drizzle-queries).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Drizzle Migrations

## Configuration

```typescript
// drizzle.config.ts
import { defineConfig } from 'drizzle-kit';

export default defineConfig({
  // Schema source (file or glob)
  schema: './src/db/schema/*',

  // Migration output directory
  out: './drizzle',

  // Database dialect
  dialect: 'postgresql',  // or 'mysql', 'sqlite', 'turso'

  // Database connection
  dbCredentials: {
    url: process.env.DATABASE_URL!,
  },

  // Optional settings
  verbose: true,            // Show SQL in console
  strict: true,             // Fail on warnings
});
```

### Multi-Schema Config

```typescript
// drizzle.config.ts — multiple schema files
export default defineConfig({
  schema: [
    './src/db/schema/users.ts',
    './src/db/schema/posts.ts',
    './src/db/schema/comments.ts',
  ],
  out: './drizzle',
  dialect: 'postgresql',
  dbCredentials: { url: process.env.DATABASE_URL! },
});
```

### Per-Database Configs

```typescript
// drizzle.config.ts — named configs for multi-DB setups
export default defineConfig({
  schema: './src/db/schema/*',
  out: './drizzle/main',
  dialect: 'postgresql',
  dbCredentials: { url: process.env.MAIN_DB_URL! },
});

// drizzle-analytics.config.ts
export default defineConfig({
  schema: './src/db/analytics-schema/*',
  out: './drizzle/analytics',
  dialect: 'postgresql',
  dbCredentials: { url: process.env.ANALYTICS_DB_URL! },
});

// Usage: npx drizzle-kit generate --config=drizzle-analytics.config.ts
```

## Commands

```bash
# Development: push schema directly (no migration files)
npx drizzle-kit push

# Generate SQL migration from schema diff
npx drizzle-kit generate

# Apply pending migrations
npx drizzle-kit migrate

# Open Drizzle Studio (database GUI)
npx drizzle-kit studio

# Check schema diff without applying
npx drizzle-kit check

# Drop a migration file
npx drizzle-kit drop
```

## Development Workflow

### Rapid Prototyping (push)

```bash
# 1. Edit schema TypeScript files
# 2. Push directly to dev database
npx drizzle-kit push

# This applies schema changes without creating migration files.
# Fast for development, but no migration history.
# NEVER use in production.
```

### Production Workflow (generate + migrate)

```bash
# 1. Edit schema TypeScript files

# 2. Generate migration SQL
npx drizzle-kit generate
# Creates: drizzle/0001_migration_name.sql

# 3. Review the generated SQL (critical!)
cat drizzle/0001_migration_name.sql

# 4. Apply migration
npx drizzle-kit migrate

# 5. Commit migration files to git
git add drizzle/ && git commit -m "Add posts table migration"
```

## Migration Files

```
drizzle/
├── 0000_init.sql                    # First migration
├── 0001_add_posts_table.sql         # Second migration
├── 0002_add_comments.sql            # Third migration
├── meta/
│   ├── 0000_snapshot.json           # Schema snapshot after each migration
│   ├── 0001_snapshot.json
│   ├── 0002_snapshot.json
│   └── _journal.json                # Migration journal (tracks applied)
```

### Example Generated Migration

```sql
-- drizzle/0001_add_posts_table.sql
CREATE TABLE IF NOT EXISTS "posts" (
  "id" text PRIMARY KEY NOT NULL,
  "title" varchar(200) NOT NULL,
  "slug" varchar(250) NOT NULL,
  "content" text NOT NULL,
  "published" boolean DEFAULT false NOT NULL,
  "author_id" text NOT NULL,
  "created_at" timestamp with time zone DEFAULT now() NOT NULL,
  "updated_at" timestamp with time zone DEFAULT now() NOT NULL,
  CONSTRAINT "posts_slug_unique" UNIQUE("slug")
);

CREATE INDEX IF NOT EXISTS "posts_author_id_idx" ON "posts" USING btree ("author_id");
CREATE INDEX IF NOT EXISTS "posts_slug_idx" ON "posts" USING btree ("slug");

ALTER TABLE "posts"
  ADD CONSTRAINT "posts_author_id_users_id_fk"
  FOREIGN KEY ("author_id") REFERENCES "public"."users"("id")
  ON DELETE cascade ON UPDATE no action;
```

## Custom Migrations

```sql
-- drizzle/custom/0003_backfill_slugs.sql
-- Manual migration for data backfill

-- Add slug column (already done by generated migration)
-- Now backfill existing rows:
UPDATE posts
SET slug = LOWER(REGEXP_REPLACE(title, '[^\w\s-]', '', 'g'))
WHERE slug IS NULL;

-- Add NOT NULL constraint after backfill
ALTER TABLE posts ALTER COLUMN slug SET NOT NULL;
```

```typescript
// Apply custom migrations programmatically
import { migrate } from 'drizzle-orm/postgres-js/migrator';

await migrate(db, {
  migrationsFolder: './drizzle',
  migrationsTable: 'drizzle_migrations',
});
```

## Seeding

```typescript
// src/db/seed.ts
import { db } from './index';
import { users, posts, categories, tags } from './schema';

async function seed() {
  console.log('Seeding database...');

  // Clear tables (order matters for FK constraints)
  await db.delete(posts);
  await db.delete(categories);
  await db.delete(tags);
  await db.delete(users);

  // Seed users
  const [admin] = await db.insert(users).values({
    email: 'admin@example.com',
    name: 'Admin User',
    password: '$2b$12$...', // pre-hashed password
    role: 'admin',
  }).returning();

  const [user1] = await db.insert(users).values({
    email: 'user@example.com',
    name: 'Regular User',
    password: '$2b$12$...',
    role: 'user',
  }).returning();

  // Seed categories
  const categoryData = ['Technology', 'Design', 'Business'].map(name => ({
    name,
    slug: name.toLowerCase(),
  }));
  const insertedCategories = await db.insert(categories).values(categoryData).returning();

  // Seed posts
  await db.insert(posts).values([
    {
      title: 'Getting Started with Drizzle',
      slug: 'getting-started-with-drizzle',
      content: 'Drizzle ORM is a TypeScript ORM...',
      published: true,
      publishedAt: new Date(),
      authorId: admin.id,
      categoryId: insertedCategories[0].id,
    },
    {
      title: 'Advanced Queries',
      slug: 'advanced-queries',
      content: 'Learn about joins, subqueries...',
      published: false,
      authorId: user1.id,
    },
  ]);

  console.log('Seed complete');
  process.exit(0);
}

seed().catch((err) => {
  console.error('Seed failed:', err);
  process.exit(1);
});
```

```json
// package.json
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

## Programmatic Migration (CI/CD)

```typescript
// src/db/migrate.ts — run at app startup or in CI
import { drizzle } from 'drizzle-orm/postgres-js';
import { migrate } from 'drizzle-orm/postgres-js/migrator';
import postgres from 'postgres';

const connection = postgres(process.env.DATABASE_URL!, { max: 1 });
const db = drizzle(connection);

async function runMigrations() {
  console.log('Running migrations...');
  await migrate(db, { migrationsFolder: './drizzle' });
  console.log('Migrations complete');
  await connection.end();
}

runMigrations().catch((err) => {
  console.error('Migration failed:', err);
  process.exit(1);
});
```

```yaml
# GitHub Actions example
- name: Run Migrations
  env:
    DATABASE_URL: ${{ secrets.DATABASE_URL }}
  run: npx tsx src/db/migrate.ts
```

## Zero-Downtime Migration Patterns

### Adding a Column

```sql
-- Safe: add nullable column (no lock, no rewrite)
ALTER TABLE posts ADD COLUMN views integer DEFAULT 0;

-- Then backfill in batches (not in migration, in a script):
UPDATE posts SET views = 0 WHERE views IS NULL AND id IN (
  SELECT id FROM posts WHERE views IS NULL LIMIT 1000
);

-- Then add NOT NULL constraint:
ALTER TABLE posts ALTER COLUMN views SET NOT NULL;
```

### Renaming a Column

```sql
-- Step 1: Add new column
ALTER TABLE posts ADD COLUMN "title_new" varchar(200);

-- Step 2: Backfill (application writes to both columns)
UPDATE posts SET title_new = title WHERE title_new IS NULL;

-- Step 3: Switch reads to new column (deploy app change)
-- Step 4: Drop old column
ALTER TABLE posts DROP COLUMN "title";
ALTER TABLE posts RENAME COLUMN "title_new" TO "title";
```

### Adding an Index

```sql
-- Use CONCURRENTLY to avoid locking the table
CREATE INDEX CONCURRENTLY IF NOT EXISTS "posts_published_idx"
  ON "posts" ("published", "created_at");

-- Note: drizzle-kit doesn't generate CONCURRENTLY.
-- For production, create a custom migration with CONCURRENTLY.
```

## Gotchas

1. **`push` vs `generate`+`migrate`.** `push` is for development only — it has no migration history and can destroy data. Always use `generate` + `migrate` for staging and production.

2. **Review generated SQL before applying.** Drizzle Kit may generate destructive operations (DROP COLUMN, DROP TABLE) if you rename or remove fields. Always `cat` the migration file before `migrate`.

3. **Migration order matters.** Files are applied in alphabetical order (0000, 0001, ...). Never rename migration files. If you need to reorder, create a new migration.

4. **`drizzle-kit generate` needs database access.** It connects to your database to compare current state with schema. Make sure DATABASE_URL is set.

5. **Custom migrations aren't tracked by `generate`.** If you write manual SQL migrations, Drizzle Kit's `generate` won't know about them. Use `check` to verify your schema matches the database.

6. **Concurrent index creation can't be in a transaction.** `CREATE INDEX CONCURRENTLY` can't run inside a transaction block. Custom migrations using this must set appropriate flags.

7. **Seed scripts should be idempotent.** Use upserts or check-then-insert patterns so seeds can be run multiple times safely.
