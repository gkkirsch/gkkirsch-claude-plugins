# Express Patterns Reference

Quick-reference guide for production Express and Fastify patterns.

---

## Middleware Chain Order

The correct order for middleware matters. Follow this sequence:

```typescript
// 1. Trust proxy (if behind reverse proxy/load balancer)
app.set('trust proxy', 1);

// 2. Request ID (before logging)
app.use(requestId());

// 3. Security headers
app.use(helmet());

// 4. CORS
app.use(cors(corsConfig));

// 5. Compression
app.use(compression());

// 6. Request logging (after request ID, before routes)
app.use(pinoHttp({ logger }));

// 7. Body parsing (with size limits)
app.use(express.json({ limit: '10kb' }));
app.use(express.urlencoded({ extended: true, limit: '10kb' }));

// 8. Cookie parsing (if using cookies)
app.use(cookieParser(process.env.COOKIE_SECRET));

// 9. Rate limiting (global)
app.use(rateLimiter({ max: 100, window: '1m' }));

// 10. Health/readiness checks (before auth)
app.get('/health', healthCheck);
app.get('/ready', readinessCheck);

// 11. Static files (if serving any)
app.use('/static', express.static('public', { maxAge: '1d' }));

// 12. API routes (with auth, validation, business logic)
app.use('/api', apiRouter);

// 13. 404 handler (after all routes)
app.use(notFoundHandler);

// 14. Error handler (must be last, must have 4 params)
app.use(errorHandler);
```

---

## Async Error Handling

### Express 5 (Built-in)

```typescript
// Express 5 automatically catches async errors
app.get('/users/:id', async (req, res) => {
  const user = await db.user.findUnique({ where: { id: req.params.id } });
  if (!user) throw new NotFoundError('User not found');
  res.json({ data: user });
});
// If findUnique throws, Express 5 passes it to error handler automatically
```

### Express 4 (Needs Wrapper)

```typescript
// Option 1: asyncHandler wrapper
const asyncHandler = (fn) => (req, res, next) =>
  Promise.resolve(fn(req, res, next)).catch(next);

app.get('/users/:id', asyncHandler(async (req, res) => {
  const user = await db.user.findUnique({ where: { id: req.params.id } });
  res.json(user);
}));

// Option 2: express-async-errors package
import 'express-async-errors';
// Patches Express to handle async errors automatically
```

---

## Error Handler Pattern

```typescript
// 4-parameter signature is REQUIRED for Express error handlers
function errorHandler(err: Error, req: Request, res: Response, _next: NextFunction) {
  // Operational errors (expected)
  if (err instanceof AppError && err.isOperational) {
    return res.status(err.statusCode).json(err.toJSON());
  }

  // Specific library errors
  if (err.name === 'SyntaxError' && 'body' in err) {
    return res.status(400).json({
      error: { code: 'INVALID_JSON', message: 'Invalid JSON in request body' },
    });
  }

  if (err.name === 'PayloadTooLargeError') {
    return res.status(413).json({
      error: { code: 'PAYLOAD_TOO_LARGE', message: 'Request body too large' },
    });
  }

  // Unexpected errors (programmer bugs)
  logger.error({ err, path: req.path, method: req.method }, 'Unhandled error');

  res.status(500).json({
    error: {
      code: 'INTERNAL_ERROR',
      message: 'An unexpected error occurred',
    },
  });
}
```

---

## Request Validation Patterns

### Zod Schema Validation

```typescript
// Define schemas
const CreatePostSchema = z.object({
  body: z.object({
    title: z.string().min(1).max(200).trim(),
    content: z.string().min(1),
    categoryId: z.string().uuid().optional(),
    tags: z.array(z.string().uuid()).max(10).default([]),
    status: z.enum(['draft', 'published']).default('draft'),
  }),
  params: z.object({}).optional(),
  query: z.object({}).optional(),
});

// Validation middleware
function validate<T extends z.ZodSchema>(schema: T) {
  return async (req: Request, _res: Response, next: NextFunction) => {
    const result = schema.safeParse({
      body: req.body,
      params: req.params,
      query: req.query,
    });

    if (!result.success) {
      throw new ValidationError('Validation failed', {
        errors: result.error.issues.map(i => ({
          path: i.path.join('.'),
          message: i.message,
        })),
      });
    }

    req.body = result.data.body;
    req.params = result.data.params ?? req.params;
    next();
  };
}

// Usage
router.post('/posts', validate(CreatePostSchema), controller.create);
```

---

## Response Patterns

### Standard Response Format

```typescript
// Success responses
// Single resource:
res.json({ data: user });

// Collection with pagination:
res.json({
  data: users,
  meta: {
    total: 150,
    page: 1,
    limit: 20,
    totalPages: 8,
  },
});

// Created:
res.status(201).json({ data: newUser });

// No content:
res.status(204).end();

// Error responses
res.status(400).json({
  error: {
    code: 'VALIDATION_ERROR',
    message: 'Validation failed',
    details: {
      errors: [
        { path: 'body.email', message: 'Invalid email format' },
      ],
    },
  },
});
```

### Content Negotiation

```typescript
app.get('/users/:id', async (req, res) => {
  const user = await getUser(req.params.id);

  res.format({
    'application/json': () => {
      res.json({ data: user });
    },
    'text/csv': () => {
      res.type('text/csv');
      res.send(`id,name,email\n${user.id},${user.name},${user.email}`);
    },
    default: () => {
      res.status(406).json({
        error: { code: 'NOT_ACCEPTABLE', message: 'Content type not supported' },
      });
    },
  });
});
```

---

## Router Organization

### Group by Domain

```
src/routes/
├── index.ts         # Route registry
├── auth/
│   ├── index.ts     # Auth router
│   ├── schemas.ts   # Validation schemas
│   └── handlers.ts  # Route handlers
├── users/
│   ├── index.ts
│   ├── schemas.ts
│   └── handlers.ts
└── posts/
    ├── index.ts
    ├── schemas.ts
    └── handlers.ts
```

### Auto-Loading Routes

```typescript
// src/routes/index.ts
import { Router } from 'express';
import { readdirSync } from 'node:fs';
import { join, dirname } from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = dirname(fileURLToPath(import.meta.url));

export async function loadRoutes(): Promise<Router> {
  const router = Router();
  const entries = readdirSync(__dirname, { withFileTypes: true });

  for (const entry of entries) {
    if (entry.isDirectory()) {
      const routeModule = await import(join(__dirname, entry.name, 'index.js'));
      const prefix = `/${entry.name}`;
      router.use(prefix, routeModule.router ?? routeModule.default);
    }
  }

  return router;
}
```

---

## Graceful Shutdown

```typescript
async function shutdown(signal: string) {
  logger.info({ signal }, 'Shutdown initiated');

  // 1. Stop accepting new connections
  server.close();

  // 2. Set health check to unhealthy (LB stops sending traffic)
  isShuttingDown = true;

  // 3. Wait for in-flight requests to complete
  await new Promise<void>((resolve) => {
    const checkConnections = setInterval(() => {
      server.getConnections((err, count) => {
        if (err || count === 0) {
          clearInterval(checkConnections);
          resolve();
        }
      });
    }, 1000);
  });

  // 4. Close external connections
  await Promise.allSettled([
    db.$disconnect(),
    redis.quit(),
    messageQueue.close(),
  ]);

  // 5. Exit
  process.exit(0);
}

// Health check respects shutdown state
app.get('/health', (req, res) => {
  if (isShuttingDown) {
    return res.status(503).json({ status: 'shutting_down' });
  }
  res.json({ status: 'healthy' });
});
```

---

## File Upload Handling

### Multer Configuration

```typescript
import multer from 'multer';
import path from 'node:path';
import { randomUUID } from 'node:crypto';

const ALLOWED_MIME_TYPES = new Set([
  'image/jpeg', 'image/png', 'image/webp', 'image/gif',
  'application/pdf',
  'text/csv',
]);

const MAX_FILE_SIZE = 10 * 1024 * 1024; // 10MB

const storage = multer.diskStorage({
  destination: (req, file, cb) => {
    cb(null, './uploads/temp');
  },
  filename: (req, file, cb) => {
    const ext = path.extname(file.originalname).toLowerCase();
    cb(null, `${randomUUID()}${ext}`);
  },
});

export const upload = multer({
  storage,
  limits: {
    fileSize: MAX_FILE_SIZE,
    files: 5,
    fields: 10,
  },
  fileFilter: (req, file, cb) => {
    if (!ALLOWED_MIME_TYPES.has(file.mimetype)) {
      cb(new BadRequestError(`File type ${file.mimetype} not allowed`));
      return;
    }
    cb(null, true);
  },
});
```

---

## Logging Best Practices

### Pino Configuration

```typescript
import pino from 'pino';

const logger = pino({
  level: process.env.LOG_LEVEL ?? 'info',
  // Development: pretty print
  ...(process.env.NODE_ENV === 'development' && {
    transport: { target: 'pino-pretty' },
  }),
  // Redact sensitive data
  redact: {
    paths: [
      'req.headers.authorization',
      'req.headers.cookie',
      '*.password',
      '*.token',
      '*.secret',
    ],
    censor: '[REDACTED]',
  },
});

// Create request-scoped logger
app.use((req, res, next) => {
  req.log = logger.child({
    requestId: req.id,
    method: req.method,
    path: req.path,
  });
  next();
});
```

### Log Levels Guide

```
fatal — Process is about to crash. Unrecoverable error.
error — Operation failed. Needs investigation.
warn  — Unexpected but recoverable. Degraded behavior.
info  — Significant events. Request/response logging. Startup.
debug — Detailed operational info. Query logging.
trace — Very detailed debugging. Loop iterations.
```

### What to Log (and What NOT)

```typescript
// DO log:
logger.info({ userId, action: 'login' }, 'User authenticated');
logger.info({ orderId, amount }, 'Order created');
logger.warn({ userId, attempts }, 'Rate limit approaching');
logger.error({ err, requestId }, 'Database query failed');

// DON'T log:
// - Passwords, tokens, API keys, credit card numbers
// - Full request bodies (may contain PII)
// - Health check requests (too noisy)
// - Individual loop iterations (too verbose)
```

---

## Express 5 Migration Checklist

### Breaking Changes from Express 4

```typescript
// 1. req.query is now a getter that returns URLSearchParams-style values
// Express 4: req.query.name (direct property access)
// Express 5: May need to handle differently based on query parser

// 2. Async errors are caught automatically
// Express 4: Need express-async-errors or asyncHandler wrapper
// Express 5: Just throw from async handlers

// 3. app.del() removed
// Express 4: app.del('/resource/:id', handler)
// Express 5: app.delete('/resource/:id', handler)

// 4. Path parameter regex changes (path-to-regexp v8)
// Express 4: app.get('/user/:id(\\d+)', handler)  // Regex in parens
// Express 5: Different syntax — check path-to-regexp v8 docs

// 5. res.send(status) removed
// Express 4: res.send(200)
// Express 5: res.sendStatus(200) or res.status(200).send('')

// 6. res.json(null) and res.json(undefined) handled differently
// Always send explicit values: res.json({ data: null })
```

---

## Security Quick Reference

### Essential Middleware Stack

```typescript
// Minimum security for any Express API:
app.use(helmet());                          // Security headers
app.use(cors({ origin: allowedOrigins }));  // Restrict origins
app.use(express.json({ limit: '10kb' }));   // Body size limit
app.use(rateLimiter({ max: 100, window: '1m' })); // Rate limiting

// On auth routes specifically:
authRouter.use(rateLimiter({ max: 5, window: '15m' }));

// Never trust client input:
// - Always validate with Zod/Joi
// - Always parameterize database queries
// - Always sanitize HTML output
// - Always verify resource ownership
```

### Cookie Security

```typescript
res.cookie('session', token, {
  httpOnly: true,       // Not accessible via JavaScript
  secure: true,         // Only sent over HTTPS
  sameSite: 'strict',   // No cross-site requests
  maxAge: 3600000,      // 1 hour
  path: '/',            // Cookie scope
  signed: true,         // Signed with cookie secret
});
```

---

## Testing Patterns

### Supertest Setup

```typescript
import { describe, it, expect, beforeAll, afterAll, beforeEach } from 'vitest';
import request from 'supertest';
import { createApp } from '../src/app.js';

const app = createApp();

describe('POST /api/users', () => {
  it('creates a user with valid data', async () => {
    const res = await request(app)
      .post('/api/users')
      .set('Authorization', `Bearer ${token}`)
      .send({ email: 'test@example.com', name: 'Test' })
      .expect('Content-Type', /json/)
      .expect(201);

    expect(res.body.data).toMatchObject({
      email: 'test@example.com',
      name: 'Test',
    });
    expect(res.body.data.id).toBeDefined();
  });

  it('returns 422 for invalid email', async () => {
    const res = await request(app)
      .post('/api/users')
      .set('Authorization', `Bearer ${token}`)
      .send({ email: 'not-an-email', name: 'Test' })
      .expect(422);

    expect(res.body.error.code).toBe('VALIDATION_ERROR');
  });

  it('returns 401 without auth', async () => {
    await request(app)
      .post('/api/users')
      .send({ email: 'test@example.com', name: 'Test' })
      .expect(401);
  });
});
```

---

## Performance Quick Wins

```typescript
// 1. Enable compression (already in middleware stack)
app.use(compression());

// 2. Use streaming for large responses
res.setHeader('Content-Type', 'application/json');
const stream = db.createReadStream();
stream.pipe(res);

// 3. Set proper cache headers
res.set('Cache-Control', 'public, max-age=3600');
res.set('ETag', etag);

// 4. Use keepAliveTimeout > LB idle timeout
server.keepAliveTimeout = 65000;
server.headersTimeout = 66000;

// 5. Avoid synchronous operations
// BAD: fs.readFileSync, crypto.randomBytes (sync), JSON.parse on huge strings
// GOOD: await fs.promises.readFile, crypto.randomBytes with callback
```

---

## Deployment Checklist

```
[ ] NODE_ENV=production
[ ] Environment variables validated at startup
[ ] Security headers enabled (Helmet)
[ ] CORS restricted to production origins
[ ] Rate limiting configured
[ ] Body size limits set
[ ] Health check endpoint working
[ ] Graceful shutdown handling SIGTERM
[ ] Structured logging (not console.log)
[ ] Error messages don't leak internals
[ ] Database connection pooling configured
[ ] Keep-alive timeout set correctly
[ ] Compression enabled
[ ] Process manager (PM2) or container orchestration
[ ] Monitoring/metrics endpoint
```
