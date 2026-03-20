---
name: prisma-migrations
description: >
  Prisma migrations — prisma migrate dev, deploy, reset, seeding,
  baseline migrations, zero-downtime strategies, and production workflows.
  Triggers: "prisma migrate", "prisma seed", "prisma db push", "prisma deploy",
  "prisma baseline", "prisma migration", "database migration".
  NOT for: schema design (use prisma-schema), queries (use prisma-queries).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Prisma Migrations

## Migration Commands

```bash
# Development: create and apply migration
npx prisma migrate dev --name add_user_avatar

# Development: apply without creating (schema matches)
npx prisma migrate dev

# Production: apply pending migrations
npx prisma migrate deploy

# Reset database: drop → recreate → apply all → seed
npx prisma migrate reset

# Check migration status
npx prisma migrate status

# Quick sync without migration files (prototyping only)
npx prisma db push

# Pull existing database into schema
npx prisma db pull

# Generate client without migrating
npx prisma generate

# Open database browser
npx prisma studio
```

## Development Workflow

```bash
# 1. Edit schema.prisma
# 2. Create migration
npx prisma migrate dev --name descriptive_name
#    → Creates prisma/migrations/YYYYMMDDHHMMSS_descriptive_name/migration.sql
#    → Applies migration to dev database
#    → Regenerates Prisma Client

# 3. If you need to edit the migration SQL before applying:
npx prisma migrate dev --create-only --name descriptive_name
#    → Creates migration file but does NOT apply
# Edit the migration.sql manually
npx prisma migrate dev
#    → Applies the edited migration
```

## Migration Naming Conventions

```bash
# Model operations
npx prisma migrate dev --name create_users_table
npx prisma migrate dev --name add_posts_table
npx prisma migrate dev --name add_comments_and_likes

# Column operations
npx prisma migrate dev --name add_avatar_to_users
npx prisma migrate dev --name add_status_to_posts
npx prisma migrate dev --name remove_legacy_fields

# Index and constraint operations
npx prisma migrate dev --name add_email_index_to_users
npx prisma migrate dev --name add_unique_constraint_slug

# Data migrations
npx prisma migrate dev --name backfill_user_roles
npx prisma migrate dev --name migrate_status_enum_values
```

## Seeding

```typescript
// prisma/seed.ts
import { PrismaClient } from '@prisma/client';
import { hash } from 'bcrypt';

const prisma = new PrismaClient();

async function main() {
  // Clean existing data (order matters for FK constraints)
  await prisma.comment.deleteMany();
  await prisma.post.deleteMany();
  await prisma.user.deleteMany();

  // Create admin user
  const admin = await prisma.user.create({
    data: {
      email: 'admin@example.com',
      name: 'Admin',
      password: await hash('admin123', 10),
      role: 'ADMIN',
    },
  });

  // Create test users with posts
  const users = await Promise.all(
    Array.from({ length: 10 }, (_, i) =>
      prisma.user.create({
        data: {
          email: `user${i}@example.com`,
          name: `User ${i}`,
          password: await hash('password', 10),
          role: 'USER',
          posts: {
            create: Array.from({ length: 3 }, (_, j) => ({
              title: `Post ${j + 1} by User ${i}`,
              content: `Content for post ${j + 1}`,
              status: j === 0 ? 'DRAFT' : 'PUBLISHED',
            })),
          },
        },
      })
    )
  );

  console.log(`Seeded ${users.length} users and ${users.length * 3} posts`);
}

main()
  .catch((e) => {
    console.error(e);
    process.exit(1);
  })
  .finally(async () => {
    await prisma.$disconnect();
  });
```

```json
// package.json
{
  "prisma": {
    "seed": "tsx prisma/seed.ts"
  }
}
```

```bash
# Run seed manually
npx prisma db seed

# Seed runs automatically on:
npx prisma migrate reset  # Always seeds after reset
npx prisma migrate dev    # Seeds if database was reset
```

## Production Deployment

```bash
# Dockerfile / CI pipeline
# 1. Generate Prisma Client (at build time)
npx prisma generate

# 2. Apply migrations (at deploy time, before app starts)
npx prisma migrate deploy

# In Docker:
# Dockerfile
FROM node:20-slim
WORKDIR /app
COPY package*.json prisma/ ./
RUN npm ci
RUN npx prisma generate
COPY . .
RUN npm run build

# Start script:
CMD ["sh", "-c", "npx prisma migrate deploy && node dist/server.js"]
```

```yaml
# GitHub Actions
- name: Apply migrations
  run: npx prisma migrate deploy
  env:
    DATABASE_URL: ${{ secrets.DATABASE_URL }}
```

## Baseline Migration (Existing Database)

```bash
# When you have an existing database and want to start using Prisma Migrate:

# 1. Introspect the existing database
npx prisma db pull
# → Updates schema.prisma to match current database

# 2. Create the initial migration without applying
npx prisma migrate dev --name init --create-only
# → Creates migration.sql matching current schema

# 3. Mark this migration as already applied
npx prisma migrate resolve --applied 20240101000000_init
# → Tells Prisma this migration is already in the database

# 4. Future migrations work normally
npx prisma migrate dev --name add_new_feature
```

## Zero-Downtime Migration Patterns

```sql
-- SAFE: Add nullable column (no table lock)
ALTER TABLE "users" ADD COLUMN "avatar" TEXT;

-- SAFE: Add column with default (Postgres 11+ is instant)
ALTER TABLE "users" ADD COLUMN "role" TEXT NOT NULL DEFAULT 'USER';

-- SAFE: Create index concurrently
CREATE INDEX CONCURRENTLY "idx_users_email" ON "users" ("email");

-- DANGEROUS: Rename column (breaks running code)
-- Instead: add new column → backfill → update code → drop old column
-- Step 1 migration:
ALTER TABLE "users" ADD COLUMN "full_name" TEXT;
UPDATE "users" SET "full_name" = "name";
-- Step 2 (after code deploys):
ALTER TABLE "users" DROP COLUMN "name";

-- DANGEROUS: Change column type
-- Instead: add new column → backfill → swap → drop
-- Step 1:
ALTER TABLE "orders" ADD COLUMN "amount_cents" INTEGER;
UPDATE "orders" SET "amount_cents" = CAST("amount" * 100 AS INTEGER);
-- Step 2 (after code update):
ALTER TABLE "orders" DROP COLUMN "amount";
ALTER TABLE "orders" RENAME COLUMN "amount_cents" TO "amount";
```

```bash
# Edit generated migration for concurrent index creation:
npx prisma migrate dev --create-only --name add_email_index
# Then edit the migration.sql:
# Change: CREATE INDEX "idx_users_email" ON "users" ("email");
# To:     CREATE INDEX CONCURRENTLY "idx_users_email" ON "users" ("email");
# Then apply:
npx prisma migrate dev
```

## Multi-Environment Setup

```bash
# .env (development)
DATABASE_URL="postgresql://user:pass@localhost:5432/myapp_dev"

# .env.test
DATABASE_URL="postgresql://user:pass@localhost:5432/myapp_test"

# .env.production
DATABASE_URL="postgresql://user:pass@prod-host:5432/myapp"

# Run migrations against specific environment
dotenv -e .env.test -- npx prisma migrate deploy
```

## Troubleshooting

```bash
# Migration drift: schema doesn't match migrations
npx prisma migrate diff \
  --from-migrations ./prisma/migrations \
  --to-schema-datamodel ./prisma/schema.prisma \
  --script
# Shows SQL needed to sync

# Reset migration history (development only!)
npx prisma migrate reset
# Drops database → recreates → applies all migrations → seeds

# Mark a failed migration as rolled back
npx prisma migrate resolve --rolled-back 20240101000000_failed_migration

# Delete migration file and re-create
rm -rf prisma/migrations/20240101000000_broken_migration
npx prisma migrate dev --name fixed_migration
```

## Gotchas

1. **`prisma db push` doesn't create migration files.** It's for prototyping only. In production, always use `prisma migrate dev` (development) and `prisma migrate deploy` (production).

2. **Never edit applied migration files.** Once a migration has been applied to any environment, editing it causes drift. Create a new migration instead.

3. **`prisma migrate reset` drops the entire database.** Including all data. Never run this in production. It's a development tool only.

4. **`prisma migrate deploy` never creates migration files.** It only applies existing migrations from `prisma/migrations/`. Use it in CI/CD and production.

5. **Concurrent index creation requires `--create-only`.** Prisma wraps migrations in transactions by default, but `CREATE INDEX CONCURRENTLY` can't run in a transaction. Edit the migration SQL to remove the transaction wrapper.

6. **Seeding is tied to `migrate reset`.** The seed function runs automatically after `prisma migrate reset` but NOT after `prisma migrate deploy`. In production, handle initial data separately.
