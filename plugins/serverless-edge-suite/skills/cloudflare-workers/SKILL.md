---
name: cloudflare-workers
description: >
  Build and deploy Cloudflare Workers — fetch handlers, KV storage, D1 database,
  R2 object storage, Durable Objects, and Hono framework integration.
  Triggers: "Cloudflare Workers", "CF Workers", "Wrangler", "Cloudflare KV",
  "D1 database", "R2 storage".
  NOT for: AWS Lambda, Vercel functions, traditional Node.js servers.
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Cloudflare Workers

## Quick Start

```bash
# Create new project
npm create cloudflare@latest my-worker

# Or manually
mkdir my-worker && cd my-worker
npm init -y
npm install -D wrangler typescript @cloudflare/workers-types
```

### wrangler.toml

```toml
name = "my-worker"
main = "src/index.ts"
compatibility_date = "2024-12-01"

# KV Namespace binding
[[kv_namespaces]]
binding = "MY_KV"
id = "abc123"

# D1 Database binding
[[d1_databases]]
binding = "DB"
database_name = "my-database"
database_id = "def456"

# R2 Bucket binding
[[r2_buckets]]
binding = "MY_BUCKET"
bucket_name = "my-bucket"

# Environment variables
[vars]
API_URL = "https://api.example.com"

# Secrets (set via wrangler secret put)
# SECRET_KEY is accessed via env.SECRET_KEY
```

## Basic Worker

```typescript
// src/index.ts
export interface Env {
  MY_KV: KVNamespace;
  DB: D1Database;
  MY_BUCKET: R2Bucket;
  API_URL: string;
  SECRET_KEY: string;
}

export default {
  async fetch(request: Request, env: Env, ctx: ExecutionContext): Promise<Response> {
    const url = new URL(request.url);

    switch (url.pathname) {
      case '/':
        return new Response('Hello from the edge!');

      case '/api/data':
        return handleData(request, env);

      default:
        return new Response('Not Found', { status: 404 });
    }
  },
} satisfies ExportedHandler<Env>;
```

## Hono Framework (Recommended)

Hono is a lightweight web framework built for edge runtimes:

```bash
npm install hono
```

```typescript
// src/index.ts
import { Hono } from 'hono';
import { cors } from 'hono/cors';
import { jwt } from 'hono/jwt';
import { logger } from 'hono/logger';
import { zValidator } from '@hono/zod-validator';
import { z } from 'zod';

interface Env {
  Bindings: {
    DB: D1Database;
    MY_KV: KVNamespace;
    JWT_SECRET: string;
  };
}

const app = new Hono<Env>();

// Middleware
app.use('*', logger());
app.use('/api/*', cors());
app.use('/api/*', jwt({ secret: 'your-secret' })); // Or use env binding

// Routes
app.get('/', (c) => c.json({ message: 'Hello from Cloudflare Workers!' }));

app.get('/api/users', async (c) => {
  const { results } = await c.env.DB.prepare(
    'SELECT * FROM users ORDER BY created_at DESC LIMIT 20'
  ).all();

  return c.json({ data: results });
});

const CreateUserSchema = z.object({
  name: z.string().min(1),
  email: z.string().email(),
});

app.post('/api/users',
  zValidator('json', CreateUserSchema),
  async (c) => {
    const { name, email } = c.req.valid('json');

    const result = await c.env.DB.prepare(
      'INSERT INTO users (name, email) VALUES (?, ?) RETURNING *'
    ).bind(name, email).first();

    return c.json(result, 201);
  }
);

app.get('/api/users/:id', async (c) => {
  const id = c.req.param('id');
  const user = await c.env.DB.prepare(
    'SELECT * FROM users WHERE id = ?'
  ).bind(id).first();

  if (!user) return c.json({ error: 'Not found' }, 404);
  return c.json(user);
});

export default app;
```

## KV Storage

Cloudflare KV is a global key-value store. Eventually consistent, optimized for reads.

```typescript
// Write
await env.MY_KV.put('user:123', JSON.stringify({ name: 'Alice' }), {
  expirationTtl: 3600,  // TTL in seconds
  metadata: { type: 'user' },
});

// Read
const value = await env.MY_KV.get('user:123', { type: 'json' });
// Returns parsed JSON or null

// Read with metadata
const { value, metadata } = await env.MY_KV.getWithMetadata('user:123', {
  type: 'json',
});

// Delete
await env.MY_KV.delete('user:123');

// List keys
const keys = await env.MY_KV.list({ prefix: 'user:', limit: 100 });
for (const key of keys.keys) {
  console.log(key.name, key.metadata);
}
```

### KV Best Practices

- **Read-heavy workloads only.** KV is optimized for reads. Writes propagate globally in ~60 seconds.
- **Use JSON type.** `get('key', { type: 'json' })` auto-parses. No need for manual `JSON.parse`.
- **Cache control with TTL.** Set `expirationTtl` for auto-expiring data (sessions, rate limits).
- **Metadata for filtering.** Store small metadata alongside values for list operations.

## D1 Database (SQLite at the Edge)

```bash
# Create database
npx wrangler d1 create my-database

# Run migrations
npx wrangler d1 migrations create my-database init
npx wrangler d1 migrations apply my-database
```

```sql
-- migrations/0001_init.sql
CREATE TABLE users (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  email TEXT NOT NULL UNIQUE,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

```typescript
// Query patterns
// Single row
const user = await env.DB.prepare(
  'SELECT * FROM users WHERE id = ?'
).bind(userId).first();

// Multiple rows
const { results } = await env.DB.prepare(
  'SELECT * FROM users WHERE role = ? LIMIT ?'
).bind('admin', 20).all();

// Insert
const { meta } = await env.DB.prepare(
  'INSERT INTO users (name, email) VALUES (?, ?)'
).bind('Alice', 'alice@example.com').run();
console.log(meta.last_row_id); // New row ID

// Batch (transaction-like)
const batch = await env.DB.batch([
  env.DB.prepare('INSERT INTO users (name, email) VALUES (?, ?)').bind('Bob', 'bob@example.com'),
  env.DB.prepare('INSERT INTO users (name, email) VALUES (?, ?)').bind('Carol', 'carol@example.com'),
]);
```

## R2 Object Storage

```typescript
// Upload
await env.MY_BUCKET.put('images/photo.jpg', imageBuffer, {
  httpMetadata: {
    contentType: 'image/jpeg',
    cacheControl: 'public, max-age=86400',
  },
  customMetadata: {
    uploadedBy: 'user-123',
  },
});

// Download
const object = await env.MY_BUCKET.get('images/photo.jpg');
if (!object) return new Response('Not Found', { status: 404 });

return new Response(object.body, {
  headers: {
    'Content-Type': object.httpMetadata?.contentType ?? 'application/octet-stream',
    'ETag': object.httpEtag,
  },
});

// Delete
await env.MY_BUCKET.delete('images/photo.jpg');

// List
const listed = await env.MY_BUCKET.list({ prefix: 'images/', limit: 100 });
for (const obj of listed.objects) {
  console.log(obj.key, obj.size, obj.uploaded);
}
```

## Scheduled Workers (Cron)

```toml
# wrangler.toml
[triggers]
crons = ["0 */6 * * *", "0 0 * * MON"]  # Every 6 hours + every Monday midnight
```

```typescript
export default {
  async fetch(request, env) { /* ... */ },

  async scheduled(event: ScheduledEvent, env: Env, ctx: ExecutionContext) {
    switch (event.cron) {
      case '0 */6 * * *':
        ctx.waitUntil(cleanupExpiredSessions(env));
        break;
      case '0 0 * * MON':
        ctx.waitUntil(sendWeeklyReport(env));
        break;
    }
  },
} satisfies ExportedHandler<Env>;
```

## Deployment

```bash
# Development
npx wrangler dev              # Local dev server with hot reload

# Deploy
npx wrangler deploy           # Deploy to production

# Deploy to staging
npx wrangler deploy --env staging

# Secrets
npx wrangler secret put SECRET_KEY    # Prompted for value
npx wrangler secret list
```

## Gotchas

- **No Node.js APIs.** No `fs`, `path`, `child_process`, `net`. Workers use Web APIs (fetch, crypto, streams, URL, TextEncoder, etc.).
- **128 MB memory limit.** You can't load large datasets into memory. Stream or paginate.
- **CPU time limits.** Free: 10ms CPU per request. Paid: 30s. This is CPU time, not wall clock time — `await fetch()` doesn't count.
- **`ctx.waitUntil()` for background work.** If you need to do work after responding (logging, analytics), use `ctx.waitUntil(promise)`. The response is sent immediately, and the Worker stays alive until the promise resolves.
- **D1 is SQLite.** No PostgreSQL features (arrays, JSONB, CTEs work but not all). Good for simple schemas.
- **KV is eventually consistent.** Writes take up to 60 seconds to propagate globally. Don't use KV for data that needs immediate consistency.
- **R2 has no CDN by default.** To serve R2 objects fast, put Cloudflare CDN or a custom domain in front of it.
- **Wrangler dev uses Miniflare locally.** Some behaviors differ from production. Always test with `wrangler deploy` to a staging environment before production.
