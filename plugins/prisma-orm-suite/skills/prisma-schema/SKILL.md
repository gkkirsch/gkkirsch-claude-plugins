---
name: prisma-schema
description: >
  Prisma schema definition — models, fields, relations, enums, indexes,
  unique constraints, default values, and database mapping.
  Triggers: "prisma schema", "prisma model", "prisma relation", "prisma enum",
  "prisma index", "prisma unique", "schema.prisma".
  NOT for: queries (use prisma-queries), migrations (use prisma-migrations).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Prisma Schema Design

## Setup

```bash
npm install prisma @prisma/client
npx prisma init --datasource-provider postgresql
```

```prisma
// prisma/schema.prisma
generator client {
  provider = "prisma-client-js"
}

datasource db {
  provider = "postgresql"
  url      = env("DATABASE_URL")
}
```

## Complete Model Example

```prisma
model User {
  id        String   @id @default(cuid())
  email     String   @unique
  name      String?
  password  String
  role      Role     @default(USER)
  avatar    String?
  bio       String?  @db.Text

  // Timestamps
  createdAt DateTime @default(now())
  updatedAt DateTime @updatedAt

  // Relations
  posts     Post[]
  comments  Comment[]
  profile   Profile?
  sessions  Session[]

  // Indexes
  @@index([email])
  @@index([role])
  @@map("users") // table name in database
}

enum Role {
  USER
  ADMIN
  MODERATOR
}
```

## Field Types

```prisma
// Scalar types
String    → VARCHAR     // @db.Text for TEXT, @db.VarChar(255) for limited
Int       → INTEGER     // @db.SmallInt, @db.BigInt for size variants
Float     → DOUBLE      // @db.Real for single precision
Decimal   → DECIMAL     // @db.Decimal(10, 2) for money
Boolean   → BOOLEAN
DateTime  → TIMESTAMP   // @db.Date for date only, @db.Time for time only
Json      → JSONB       // PostgreSQL JSONB
Bytes     → BYTEA       // Binary data
BigInt    → BIGINT      // For large numbers

// Modifiers
String?            // Optional (nullable)
String[]           // Array (PostgreSQL only)
String @unique     // Unique constraint
String @default("value")  // Default value

// ID strategies
@id @default(cuid())       // Collision-resistant unique ID (recommended)
@id @default(uuid())       // UUID v4
@id @default(autoincrement()) // Auto-incrementing integer
@id @default(dbgenerated("gen_random_uuid()")) // Database-generated

// Default values
@default(now())            // Current timestamp
@default(true)             // Boolean default
@default(0)                // Numeric default
@default("")               // Empty string
@updatedAt                 // Auto-update on save
@default(dbgenerated("gen_random_uuid()")) // DB function
```

## Relations

```prisma
// ONE-TO-ONE
model User {
  id      String   @id @default(cuid())
  profile Profile?
}

model Profile {
  id     String @id @default(cuid())
  bio    String @db.Text
  userId String @unique  // @unique makes it 1-to-1
  user   User   @relation(fields: [userId], references: [id], onDelete: Cascade)

  @@index([userId])
}

// ONE-TO-MANY
model User {
  id    String @id @default(cuid())
  posts Post[]
}

model Post {
  id       String @id @default(cuid())
  title    String
  content  String @db.Text
  userId   String
  user     User   @relation(fields: [userId], references: [id], onDelete: Cascade)

  @@index([userId])
}

// MANY-TO-MANY (implicit — Prisma manages join table)
model Post {
  id   String @id @default(cuid())
  tags Tag[]
}

model Tag {
  id    String @id @default(cuid())
  name  String @unique
  posts Post[]
}

// MANY-TO-MANY (explicit — with extra data on relationship)
model Post {
  id       String       @id @default(cuid())
  tags     PostTag[]
}

model Tag {
  id    String    @id @default(cuid())
  name  String    @unique
  posts PostTag[]
}

model PostTag {
  postId    String
  tagId     String
  addedAt   DateTime @default(now())
  addedById String?

  post Post @relation(fields: [postId], references: [id], onDelete: Cascade)
  tag  Tag  @relation(fields: [tagId], references: [id], onDelete: Cascade)

  @@id([postId, tagId]) // Composite primary key
  @@index([tagId])
}

// SELF-REFERENTIAL
model User {
  id        String @id @default(cuid())
  managerId String?
  manager   User?  @relation("Management", fields: [managerId], references: [id])
  reports   User[] @relation("Management")

  @@index([managerId])
}

// MULTIPLE RELATIONS TO SAME MODEL
model User {
  id             String    @id @default(cuid())
  writtenPosts   Post[]    @relation("author")
  editedPosts    Post[]    @relation("editor")
}

model Post {
  id       String @id @default(cuid())
  authorId String
  editorId String?
  author   User   @relation("author", fields: [authorId], references: [id])
  editor   User?  @relation("editor", fields: [editorId], references: [id])

  @@index([authorId])
  @@index([editorId])
}
```

## Referential Actions

```prisma
@relation(fields: [userId], references: [id], onDelete: Cascade)
//                                              onDelete options:
// Cascade    — delete child when parent deleted
// SetNull    — set FK to null (field must be optional)
// Restrict   — prevent parent deletion if children exist
// NoAction   — like Restrict, but checked at end of transaction
// SetDefault — set FK to default value

@relation(fields: [userId], references: [id], onUpdate: Cascade)
// Same options for onUpdate (when parent ID changes)
```

## Indexes and Constraints

```prisma
model Post {
  id        String   @id @default(cuid())
  slug      String
  authorId  String
  status    Status
  createdAt DateTime @default(now())

  // Single column index
  @@index([authorId])

  // Composite index (order matters for query optimization)
  @@index([authorId, status])

  // Index with sort order
  @@index([createdAt(sort: Desc)])

  // Composite unique constraint
  @@unique([authorId, slug])

  // Full-text search index (PostgreSQL)
  @@index([title], type: Gin)

  // Table name mapping
  @@map("posts")
}
```

## Enums

```prisma
enum Status {
  DRAFT
  PUBLISHED
  ARCHIVED
}

enum Role {
  USER
  ADMIN
  MODERATOR
  SUPER_ADMIN
}

// Usage in models
model Post {
  status Status @default(DRAFT)
}
```

## Complete SaaS Schema Example

```prisma
model Organization {
  id        String   @id @default(cuid())
  name      String
  slug      String   @unique
  plan      Plan     @default(FREE)
  members   Member[]
  projects  Project[]
  createdAt DateTime @default(now())
  updatedAt DateTime @updatedAt

  @@map("organizations")
}

model Member {
  id             String       @id @default(cuid())
  role           MemberRole   @default(MEMBER)
  userId         String
  organizationId String
  user           User         @relation(fields: [userId], references: [id], onDelete: Cascade)
  organization   Organization @relation(fields: [organizationId], references: [id], onDelete: Cascade)
  joinedAt       DateTime     @default(now())

  @@unique([userId, organizationId])
  @@index([organizationId])
  @@map("members")
}

enum Plan {
  FREE
  PRO
  ENTERPRISE
}

enum MemberRole {
  OWNER
  ADMIN
  MEMBER
  VIEWER
}
```

## Gotchas

1. **`@unique` on relation fields makes it one-to-one.** Without `@unique`, `userId String` creates a one-to-many. With `@unique`, it's one-to-one. This is a common source of bugs.

2. **Always add `@@index` on foreign key columns.** Prisma doesn't auto-create indexes on FK columns. Without them, JOIN queries and cascade deletes are slow on large tables.

3. **`@updatedAt` only triggers on Prisma operations.** Direct SQL updates bypass it. If you use raw queries or database triggers, manage timestamps manually.

4. **Implicit M2M creates a hidden join table.** Named `_PostToTag` with columns `A` and `B`. You can't add extra fields (like `addedAt`) to implicit M2M. Use explicit join models if you need metadata.

5. **`@default(cuid())` vs `@default(uuid())`.** CUIDs are shorter (25 chars vs 36), URL-safe, and sortable by creation time. UUIDs are standard but longer. Both are collision-resistant.

6. **Enum changes require migration.** Adding a new enum value needs `prisma migrate dev`. For frequently changing sets (categories, tags), use a lookup table instead of an enum.

7. **`Json` fields lose type safety.** Prisma treats JSON as `any`. Use Zod for runtime validation, or consider structured columns instead of JSON for important data.
