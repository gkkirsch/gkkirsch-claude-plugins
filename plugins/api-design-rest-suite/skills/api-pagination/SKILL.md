---
name: api-pagination
description: >
  Implement API pagination — offset-based, cursor-based, keyset pagination,
  and page-based patterns with proper response metadata.
  Triggers: "API pagination", "paginate results", "cursor pagination", "list endpoint".
  NOT for: frontend infinite scroll (UI-side), database query optimization.
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# API Pagination Patterns

## Strategy Comparison

| Strategy | Best For | Pros | Cons |
|----------|---------|------|------|
| **Offset/Limit** | Admin dashboards, simple lists | Jump to any page, easy total count | Slow on large datasets, inconsistent with inserts |
| **Cursor-based** | Feeds, infinite scroll, large datasets | Consistent, performant at scale | Can't jump to page N, complex implementation |
| **Keyset** | Time-series, chronological data | Very fast, uses indexes | Must sort by indexed column |
| **Page-based** | Simple UIs with page numbers | Familiar UX, easy to implement | Same issues as offset |

**Default choice**: Offset for admin/internal tools. Cursor for public APIs and large datasets.

## Offset-Based Pagination

The simplest approach. Client sends `page` and `limit` (or `offset` and `limit`).

### Implementation

```typescript
import { z } from 'zod';
import { Request, Response } from 'express';
import { Pool } from 'pg';

const PaginationSchema = z.object({
  page: z.coerce.number().int().min(1).default(1),
  limit: z.coerce.number().int().min(1).max(100).default(20),
  sort: z.enum(['createdAt', 'name', 'email']).default('createdAt'),
  order: z.enum(['asc', 'desc']).default('desc'),
});

async function listUsers(req: Request, res: Response) {
  const { page, limit, sort, order } = PaginationSchema.parse(req.query);
  const offset = (page - 1) * limit;

  // Two queries: data + count (run in parallel)
  const [dataResult, countResult] = await Promise.all([
    pool.query(
      `SELECT id, name, email, created_at
       FROM users
       WHERE deleted_at IS NULL
       ORDER BY ${sort} ${order}
       LIMIT $1 OFFSET $2`,
      [limit, offset]
    ),
    pool.query(
      `SELECT COUNT(*) as total FROM users WHERE deleted_at IS NULL`
    ),
  ]);

  const total = parseInt(countResult.rows[0].total);
  const totalPages = Math.ceil(total / limit);

  res.json({
    data: dataResult.rows,
    meta: {
      page,
      limit,
      total,
      totalPages,
      hasNextPage: page < totalPages,
      hasPrevPage: page > 1,
    },
  });
}
```

### Response Format

```json
{
  "data": [
    { "id": "1", "name": "Alice", "email": "alice@example.com" },
    { "id": "2", "name": "Bob", "email": "bob@example.com" }
  ],
  "meta": {
    "page": 1,
    "limit": 20,
    "total": 142,
    "totalPages": 8,
    "hasNextPage": true,
    "hasPrevPage": false
  }
}
```

### When Offset Breaks Down

```
Page 1: [A, B, C, D, E]  ← User reads page 1
Insert: [NEW, A, B, C, D, E]  ← New item inserted
Page 2: [E, F, G, H, I]  ← User loads page 2 — sees E again!
```

Offset pagination is inconsistent when data changes between requests. For feeds or frequently-updated data, use cursor-based.

## Cursor-Based Pagination

Uses an opaque cursor (encoded ID or timestamp) to mark position. No duplicate items, consistent ordering.

### Implementation

```typescript
import { z } from 'zod';

const CursorSchema = z.object({
  cursor: z.string().optional(),
  limit: z.coerce.number().int().min(1).max(100).default(20),
});

// Encode/decode cursors (base64 of the sort value)
function encodeCursor(value: string | Date): string {
  return Buffer.from(String(value)).toString('base64url');
}

function decodeCursor(cursor: string): string {
  return Buffer.from(cursor, 'base64url').toString();
}

async function listUsers(req: Request, res: Response) {
  const { cursor, limit } = CursorSchema.parse(req.query);

  let query = `
    SELECT id, name, email, created_at
    FROM users
    WHERE deleted_at IS NULL
  `;
  const params: any[] = [limit + 1]; // Fetch one extra to detect hasNextPage

  if (cursor) {
    const decodedCursor = decodeCursor(cursor);
    query += ` AND created_at < $2`;
    params.push(decodedCursor);
  }

  query += ` ORDER BY created_at DESC LIMIT $1`;

  const result = await pool.query(query, params);

  const hasNextPage = result.rows.length > limit;
  const items = hasNextPage ? result.rows.slice(0, limit) : result.rows;

  const nextCursor = hasNextPage
    ? encodeCursor(items[items.length - 1].created_at)
    : null;

  res.json({
    data: items,
    meta: {
      hasNextPage,
      nextCursor,
      limit,
    },
  });
}
```

### Response Format

```json
{
  "data": [
    { "id": "10", "name": "Alice", "createdAt": "2026-03-19T10:00:00Z" },
    { "id": "9", "name": "Bob", "createdAt": "2026-03-18T15:30:00Z" }
  ],
  "meta": {
    "hasNextPage": true,
    "nextCursor": "MjAyNi0wMy0xOFQxNTozMDowMFo",
    "limit": 20
  }
}
```

Client fetches next page: `GET /users?cursor=MjAyNi0wMy0xOFQxNTozMDowMFo&limit=20`

### Compound Cursor (Multiple Sort Fields)

When sorting by non-unique fields, include a tiebreaker:

```typescript
// Sort by name + id (tiebreaker)
function encodeCursor(name: string, id: string): string {
  return Buffer.from(JSON.stringify({ name, id })).toString('base64url');
}

function decodeCursor(cursor: string): { name: string; id: string } {
  return JSON.parse(Buffer.from(cursor, 'base64url').toString());
}

// Query with compound WHERE
if (cursor) {
  const { name, id } = decodeCursor(cursor);
  query += ` AND (name, id) > ($2, $3)`;
  params.push(name, id);
}
query += ` ORDER BY name ASC, id ASC LIMIT $1`;
```

## Keyset Pagination

Like cursor-based but using visible sort values instead of opaque cursors. Best for time-series data.

```typescript
async function listEvents(req: Request, res: Response) {
  const { after, before, limit = 50 } = req.query;

  let query = `
    SELECT id, type, data, created_at
    FROM events
    WHERE 1=1
  `;
  const params: any[] = [];

  if (after) {
    params.push(after);
    query += ` AND created_at > $${params.length}`;
  }
  if (before) {
    params.push(before);
    query += ` AND created_at < $${params.length}`;
  }

  params.push(Number(limit) + 1);
  query += ` ORDER BY created_at DESC LIMIT $${params.length}`;

  const result = await pool.query(query, params);
  const hasMore = result.rows.length > Number(limit);
  const items = hasMore ? result.rows.slice(0, Number(limit)) : result.rows;

  res.json({
    data: items,
    meta: {
      hasMore,
      oldest: items[items.length - 1]?.created_at,
      newest: items[0]?.created_at,
    },
  });
}

// Client: GET /events?after=2026-03-18T00:00:00Z&limit=50
```

## Prisma Pagination

### Offset

```typescript
const [users, total] = await Promise.all([
  prisma.user.findMany({
    skip: (page - 1) * limit,
    take: limit,
    orderBy: { createdAt: 'desc' },
    where: { deletedAt: null },
  }),
  prisma.user.count({ where: { deletedAt: null } }),
]);
```

### Cursor

```typescript
const users = await prisma.user.findMany({
  take: limit + 1,
  cursor: cursor ? { id: cursor } : undefined,
  skip: cursor ? 1 : 0,  // Skip the cursor item itself
  orderBy: { createdAt: 'desc' },
});

const hasNextPage = users.length > limit;
const items = hasNextPage ? users.slice(0, limit) : users;
const nextCursor = hasNextPage ? items[items.length - 1].id : null;
```

## Link Headers (HATEOAS)

Include navigation links in response headers:

```typescript
function setPaginationLinks(
  res: Response,
  baseUrl: string,
  page: number,
  totalPages: number,
  limit: number
) {
  const links: string[] = [];

  if (page > 1) {
    links.push(`<${baseUrl}?page=1&limit=${limit}>; rel="first"`);
    links.push(`<${baseUrl}?page=${page - 1}&limit=${limit}>; rel="prev"`);
  }
  if (page < totalPages) {
    links.push(`<${baseUrl}?page=${page + 1}&limit=${limit}>; rel="next"`);
    links.push(`<${baseUrl}?page=${totalPages}&limit=${limit}>; rel="last"`);
  }

  if (links.length > 0) {
    res.set('Link', links.join(', '));
  }
}
```

## Gotchas

- **Always cap `limit`.** Never let clients request `limit=999999`. Set a hard max (50-100 is standard).
- **COUNT(*) is slow on large tables.** For offset pagination, consider caching the total count or returning an estimate. Cursor pagination avoids this entirely.
- **OFFSET doesn't use indexes.** `OFFSET 10000` still scans 10,000 rows before returning results. Keyset/cursor pagination uses index seeks.
- **Sort order must be deterministic.** If sorting by `name`, two users with the same name have undefined order. Always add a tiebreaker (`name, id`).
- **Cursor must be opaque.** Don't expose raw database IDs or timestamps as cursors. Base64 encode them. This lets you change the underlying implementation without breaking clients.
- **Don't paginate small collections.** If your API never returns more than 50 items, skip pagination complexity and return all items.
- **Include `hasNextPage` / `hasPrevPage`.** Don't make clients calculate this from `total` and `page`. Explicit booleans are clearer.
- **`limit + 1` trick.** Fetch one extra row to detect if more data exists, then slice it off. This avoids a separate COUNT query for cursor pagination.
