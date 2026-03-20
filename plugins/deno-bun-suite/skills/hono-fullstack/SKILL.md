---
name: hono-fullstack
description: >
  Hono framework patterns for building full-stack apps on Deno and Bun.
  Use when building APIs with Hono, adding middleware, implementing auth,
  serving static files, or deploying Hono apps to edge runtimes.
  Triggers: "hono", "hono framework", "hono middleware", "hono + bun",
  "hono + deno", "hono auth", "hono api", "hono cors", "hono validator".
  NOT for: Express.js, Fastify, Koa, or Node.js-only HTTP servers.
version: 1.0.0
allowed-tools: Read, Grep, Glob, Bash
---

# Hono Full-Stack Patterns

## App Structure

```typescript
// src/index.ts — Hono app with modular routes
import { Hono } from 'hono';
import { cors } from 'hono/cors';
import { logger } from 'hono/logger';
import { secureHeaders } from 'hono/secure-headers';
import { timing } from 'hono/timing';
import { prettyJSON } from 'hono/pretty-json';
import { userRoutes } from './routes/users';
import { authRoutes } from './routes/auth';
import { errorHandler } from './middleware/error';

// Typed environment bindings
type Bindings = {
  DATABASE_URL: string;
  JWT_SECRET: string;
  ENVIRONMENT: 'development' | 'production';
};

const app = new Hono<{ Bindings: Bindings }>();

// Global middleware
app.use('*', logger());
app.use('*', timing());
app.use('*', secureHeaders());
app.use('*', prettyJSON());
app.use('/api/*', cors({
  origin: ['http://localhost:5173', 'https://myapp.com'],
  credentials: true,
}));

// Error handler
app.onError(errorHandler);

// Routes
app.route('/api/auth', authRoutes);
app.route('/api/users', userRoutes);

// Health check
app.get('/health', (c) => c.json({ status: 'ok', uptime: process.uptime() }));

// 404
app.notFound((c) => c.json({ error: 'Not Found' }, 404));

// Export for runtime
export default {
  port: 3000,
  fetch: app.fetch,
};
```

## Route Modules with Validation

```typescript
// routes/users.ts
import { Hono } from 'hono';
import { zValidator } from '@hono/zod-validator';
import { z } from 'zod';
import { authMiddleware } from '../middleware/auth';

const createUserSchema = z.object({
  name: z.string().min(1).max(100),
  email: z.string().email(),
  role: z.enum(['admin', 'user']).default('user'),
});

const updateUserSchema = createUserSchema.partial();

const querySchema = z.object({
  page: z.coerce.number().int().positive().default(1),
  limit: z.coerce.number().int().min(1).max(100).default(20),
  search: z.string().optional(),
});

export const userRoutes = new Hono()
  // Protected routes
  .use('*', authMiddleware)

  // List users with pagination
  .get('/', zValidator('query', querySchema), async (c) => {
    const { page, limit, search } = c.req.valid('query');
    const offset = (page - 1) * limit;

    const users = await db.query(
      `SELECT id, name, email, role, created_at
       FROM users
       WHERE ($1::text IS NULL OR name ILIKE '%' || $1 || '%')
       ORDER BY created_at DESC
       LIMIT $2 OFFSET $3`,
      [search ?? null, limit, offset]
    );

    const [{ count }] = await db.query('SELECT COUNT(*) FROM users');

    return c.json({
      data: users,
      pagination: {
        page,
        limit,
        total: Number(count),
        pages: Math.ceil(Number(count) / limit),
      },
    });
  })

  // Get single user
  .get('/:id', async (c) => {
    const id = c.req.param('id');
    const user = await db.query('SELECT * FROM users WHERE id = $1', [id]);

    if (!user.length) {
      return c.json({ error: 'User not found' }, 404);
    }

    return c.json(user[0]);
  })

  // Create user
  .post('/', zValidator('json', createUserSchema), async (c) => {
    const data = c.req.valid('json');

    const [user] = await db.query(
      `INSERT INTO users (id, name, email, role)
       VALUES (gen_random_uuid(), $1, $2, $3)
       RETURNING *`,
      [data.name, data.email, data.role]
    );

    return c.json(user, 201);
  })

  // Update user
  .patch('/:id', zValidator('json', updateUserSchema), async (c) => {
    const id = c.req.param('id');
    const data = c.req.valid('json');

    const fields = Object.entries(data)
      .map(([key, _], i) => `${key} = $${i + 2}`)
      .join(', ');

    const [user] = await db.query(
      `UPDATE users SET ${fields}, updated_at = NOW() WHERE id = $1 RETURNING *`,
      [id, ...Object.values(data)]
    );

    if (!user) return c.json({ error: 'User not found' }, 404);
    return c.json(user);
  })

  // Delete user
  .delete('/:id', async (c) => {
    const id = c.req.param('id');
    await db.query('DELETE FROM users WHERE id = $1', [id]);
    return c.body(null, 204);
  });
```

## Auth Middleware with JWT

```typescript
// middleware/auth.ts
import { Context, Next } from 'hono';
import { jwt } from 'hono/jwt';
import { HTTPException } from 'hono/http-exception';

// JWT middleware (built-in)
export const authMiddleware = jwt({ secret: process.env.JWT_SECRET! });

// Custom middleware with role checking
export function requireRole(...roles: string[]) {
  return async (c: Context, next: Next) => {
    const payload = c.get('jwtPayload');
    if (!payload) {
      throw new HTTPException(401, { message: 'Authentication required' });
    }

    if (!roles.includes(payload.role)) {
      throw new HTTPException(403, { message: 'Insufficient permissions' });
    }

    await next();
  };
}

// Rate limiting middleware
export function rateLimit(opts: { max: number; windowMs: number }) {
  const requests = new Map<string, { count: number; resetAt: number }>();

  return async (c: Context, next: Next) => {
    const key = c.req.header('x-forwarded-for') || 'unknown';
    const now = Date.now();
    const record = requests.get(key);

    if (!record || now > record.resetAt) {
      requests.set(key, { count: 1, resetAt: now + opts.windowMs });
    } else if (record.count >= opts.max) {
      c.header('Retry-After', String(Math.ceil((record.resetAt - now) / 1000)));
      throw new HTTPException(429, { message: 'Too many requests' });
    } else {
      record.count++;
    }

    await next();
  };
}

// Usage:
// app.use('/api/auth/login', rateLimit({ max: 5, windowMs: 60_000 }));
// app.use('/api/admin/*', requireRole('admin'));
```

## Error Handling

```typescript
// middleware/error.ts
import { Context } from 'hono';
import { HTTPException } from 'hono/http-exception';
import { ZodError } from 'zod';

export function errorHandler(err: Error, c: Context) {
  // Zod validation errors
  if (err instanceof ZodError) {
    return c.json({
      error: 'Validation Error',
      details: err.errors.map(e => ({
        field: e.path.join('.'),
        message: e.message,
      })),
    }, 400);
  }

  // HTTP exceptions (from throw new HTTPException)
  if (err instanceof HTTPException) {
    return c.json({ error: err.message }, err.status);
  }

  // Unexpected errors
  console.error('Unhandled error:', err);
  return c.json(
    { error: c.env.ENVIRONMENT === 'production' ? 'Internal Server Error' : err.message },
    500
  );
}
```

## Static Files & SSR

```typescript
import { Hono } from 'hono';
import { serveStatic } from 'hono/bun'; // or 'hono/deno' for Deno

const app = new Hono();

// Serve static files from ./public
app.use('/static/*', serveStatic({ root: './' }));

// Serve index.html for SPA routes
app.get('*', serveStatic({ path: './public/index.html' }));

// Or: SSR with streaming
app.get('/', async (c) => {
  return c.html(
    `<!DOCTYPE html>
    <html>
    <head><title>My App</title></head>
    <body>
      <div id="app">${await renderToString(<App />)}</div>
      <script src="/static/client.js"></script>
    </body>
    </html>`
  );
});
```

## Testing Hono Apps

```typescript
// test/users.test.ts
import { describe, it, expect } from 'bun:test'; // or Deno.test
import app from '../src/index';

describe('Users API', () => {
  it('GET /api/users returns paginated list', async () => {
    const res = await app.request('/api/users?page=1&limit=10', {
      headers: { Authorization: `Bearer ${testToken}` },
    });

    expect(res.status).toBe(200);
    const body = await res.json();
    expect(body.data).toBeArray();
    expect(body.pagination.page).toBe(1);
  });

  it('POST /api/users validates input', async () => {
    const res = await app.request('/api/users', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${testToken}`,
      },
      body: JSON.stringify({ name: '', email: 'invalid' }),
    });

    expect(res.status).toBe(400);
    const body = await res.json();
    expect(body.error).toBe('Validation Error');
  });

  it('GET /health returns ok', async () => {
    const res = await app.request('/health');
    expect(res.status).toBe(200);
    expect(await res.json()).toMatchObject({ status: 'ok' });
  });
});
```

## Gotchas

1. **hono/bun vs hono/deno imports** -- `serveStatic`, `serve`, and some helpers have runtime-specific imports. Using `hono/bun` on Deno or vice versa crashes at runtime. Use the correct import for your target runtime. The core `hono` package is runtime-agnostic.

2. **Middleware order matters** -- Middleware runs in registration order. Auth middleware must come BEFORE route handlers. CORS middleware must come BEFORE routes that need it. `app.onError` must be registered before routes to catch their errors.

3. **c.json() returns Response, not void** -- You must `return c.json(...)`. A handler without a return statement produces no response and the request hangs. TypeScript doesn't always catch this if the return type is `Response | void`.

4. **zValidator position argument** -- `zValidator('json', schema)` validates the body. `zValidator('query', schema)` validates query params. `zValidator('param', schema)` validates URL params. Using the wrong position silently validates nothing.

5. **app.request() for testing** -- Hono's `app.request()` creates a test request without starting a server. It returns a standard `Response` object. This is the correct way to test Hono apps. Don't start a real server in tests.

6. **HTTPException vs returning error responses** -- `throw new HTTPException(404)` triggers the `app.onError` handler. `return c.json({error: 'Not found'}, 404)` does not. Use HTTPException for errors you want centralized error handling to process. Use direct returns for expected "not found" responses.
