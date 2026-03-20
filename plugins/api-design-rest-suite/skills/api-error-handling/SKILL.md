---
name: api-error-handling
description: >
  Design and implement structured API error handling — error response format,
  error classes, validation errors, global error middleware, and error codes.
  Triggers: "API error handling", "error responses", "validation errors", "error middleware".
  NOT for: frontend error boundaries, logging/monitoring setup.
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# API Error Handling

## Standard Error Response Format

Every error response should follow the same structure:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Request validation failed",
    "details": [
      { "field": "email", "message": "Must be a valid email address" },
      { "field": "name", "message": "Required" }
    ],
    "requestId": "req_abc123"
  }
}
```

**Rules:**
- `code` — machine-readable string (UPPER_SNAKE_CASE). Clients switch on this.
- `message` — human-readable description. For developers, not end users.
- `details` — optional array of sub-errors (validation, bulk operations).
- `requestId` — for support/debugging correlation.
- **Never expose stack traces or internal details in production.**

## Error Classes

```typescript
// src/lib/errors.ts

export class AppError extends Error {
  constructor(
    public readonly statusCode: number,
    public readonly code: string,
    message: string,
    public readonly details?: Array<{ field?: string; message: string }>,
  ) {
    super(message);
    this.name = 'AppError';
  }
}

// 400 — Bad Request
export class BadRequestError extends AppError {
  constructor(message = 'Bad request', details?: AppError['details']) {
    super(400, 'BAD_REQUEST', message, details);
  }
}

// 401 — Unauthorized
export class UnauthorizedError extends AppError {
  constructor(message = 'Authentication required') {
    super(401, 'UNAUTHORIZED', message);
  }
}

// 403 — Forbidden
export class ForbiddenError extends AppError {
  constructor(message = 'Insufficient permissions') {
    super(403, 'FORBIDDEN', message);
  }
}

// 404 — Not Found
export class NotFoundError extends AppError {
  constructor(resource = 'Resource', id?: string) {
    super(404, 'NOT_FOUND', id ? `${resource} '${id}' not found` : `${resource} not found`);
  }
}

// 409 — Conflict
export class ConflictError extends AppError {
  constructor(message = 'Resource already exists') {
    super(409, 'CONFLICT', message);
  }
}

// 422 — Unprocessable Entity
export class ValidationError extends AppError {
  constructor(details: Array<{ field: string; message: string }>) {
    super(422, 'VALIDATION_ERROR', 'Validation failed', details);
  }
}

// 429 — Too Many Requests
export class RateLimitError extends AppError {
  constructor(retryAfter: number) {
    super(429, 'RATE_LIMIT_EXCEEDED', `Rate limit exceeded. Retry after ${retryAfter} seconds`);
  }
}
```

### Using Error Classes in Routes

```typescript
// Clean, expressive error throwing
async function getUser(req: Request, res: Response) {
  const user = await db.user.findUnique({ where: { id: req.params.id } });
  if (!user) throw new NotFoundError('User', req.params.id);

  if (user.organizationId !== req.user.organizationId) {
    throw new ForbiddenError('You can only access users in your organization');
  }

  res.json(user);
}

async function createUser(req: Request, res: Response) {
  const existing = await db.user.findUnique({ where: { email: req.body.email } });
  if (existing) throw new ConflictError('A user with this email already exists');

  const user = await db.user.create({ data: req.body });
  res.status(201).json(user);
}
```

## Zod Validation Middleware

```typescript
import { z, ZodError, ZodSchema } from 'zod';
import { Request, Response, NextFunction } from 'express';
import { ValidationError } from './errors.js';

// Validate request body, query, or params
export function validate(schema: {
  body?: ZodSchema;
  query?: ZodSchema;
  params?: ZodSchema;
}) {
  return (req: Request, res: Response, next: NextFunction) => {
    try {
      if (schema.body) req.body = schema.body.parse(req.body);
      if (schema.query) req.query = schema.query.parse(req.query) as any;
      if (schema.params) req.params = schema.params.parse(req.params) as any;
      next();
    } catch (err) {
      if (err instanceof ZodError) {
        throw new ValidationError(
          err.errors.map((e) => ({
            field: e.path.join('.'),
            message: e.message,
          }))
        );
      }
      throw err;
    }
  };
}

// Usage
const CreateUserSchema = z.object({
  email: z.string().email('Must be a valid email'),
  name: z.string().min(1, 'Required').max(100, 'Must be 100 characters or less'),
  role: z.enum(['admin', 'user']).optional(),
});

router.post('/users',
  validate({ body: CreateUserSchema }),
  createUser,
);
```

## Global Error Middleware

```typescript
// src/middleware/error-handler.ts
import { Request, Response, NextFunction } from 'express';
import { AppError } from '../lib/errors.js';

export function errorHandler(
  err: Error,
  req: Request,
  res: Response,
  next: NextFunction
) {
  // Request ID for correlation
  const requestId = req.headers['x-request-id'] as string ?? crypto.randomUUID();

  // Known application errors
  if (err instanceof AppError) {
    return res.status(err.statusCode).json({
      error: {
        code: err.code,
        message: err.message,
        details: err.details,
        requestId,
      },
    });
  }

  // Prisma errors
  if (err.constructor.name === 'PrismaClientKnownRequestError') {
    const prismaErr = err as any;
    if (prismaErr.code === 'P2002') {
      // Unique constraint violation
      const field = prismaErr.meta?.target?.[0] ?? 'field';
      return res.status(409).json({
        error: {
          code: 'CONFLICT',
          message: `A record with this ${field} already exists`,
          requestId,
        },
      });
    }
    if (prismaErr.code === 'P2025') {
      // Record not found
      return res.status(404).json({
        error: {
          code: 'NOT_FOUND',
          message: 'Resource not found',
          requestId,
        },
      });
    }
  }

  // JSON parse errors
  if (err.type === 'entity.parse.failed') {
    return res.status(400).json({
      error: {
        code: 'INVALID_JSON',
        message: 'Request body contains invalid JSON',
        requestId,
      },
    });
  }

  // Unexpected errors — don't leak internals
  console.error(`[${requestId}] Unhandled error:`, err);

  res.status(500).json({
    error: {
      code: 'INTERNAL_ERROR',
      message: process.env.NODE_ENV === 'production'
        ? 'An unexpected error occurred'
        : err.message,
      requestId,
    },
  });
}

// Mount AFTER all routes
app.use(errorHandler);
```

## Async Error Wrapper

Express doesn't catch async errors by default. Wrap async handlers:

```typescript
// Option 1: Wrapper function
function asyncHandler(fn: (req: Request, res: Response, next: NextFunction) => Promise<any>) {
  return (req: Request, res: Response, next: NextFunction) => {
    fn(req, res, next).catch(next);
  };
}

// Usage
router.get('/users/:id', asyncHandler(async (req, res) => {
  const user = await getUser(req.params.id);
  if (!user) throw new NotFoundError('User', req.params.id);
  res.json(user);
}));

// Option 2: express-async-errors (auto-patches Express)
import 'express-async-errors';
// Now async errors are caught automatically — no wrapper needed
```

**Recommended**: Use `express-async-errors`. It's a one-line import that patches Express. No wrappers needed.

```bash
npm install express-async-errors
```

```typescript
// Import BEFORE express
import 'express-async-errors';
import express from 'express';
```

## Error Code Registry

Maintain a list of all error codes for documentation:

```typescript
// src/lib/error-codes.ts

export const ERROR_CODES = {
  // Authentication
  UNAUTHORIZED: { status: 401, message: 'Authentication required' },
  INVALID_TOKEN: { status: 401, message: 'Invalid or expired token' },
  TOKEN_EXPIRED: { status: 401, message: 'Token has expired' },

  // Authorization
  FORBIDDEN: { status: 403, message: 'Insufficient permissions' },
  OWNERSHIP_REQUIRED: { status: 403, message: 'You do not own this resource' },

  // Validation
  VALIDATION_ERROR: { status: 422, message: 'Validation failed' },
  BAD_REQUEST: { status: 400, message: 'Invalid request' },
  INVALID_JSON: { status: 400, message: 'Invalid JSON in request body' },

  // Resources
  NOT_FOUND: { status: 404, message: 'Resource not found' },
  CONFLICT: { status: 409, message: 'Resource conflict' },
  ALREADY_EXISTS: { status: 409, message: 'Resource already exists' },

  // Rate limiting
  RATE_LIMIT_EXCEEDED: { status: 429, message: 'Too many requests' },
  DAILY_LIMIT_EXCEEDED: { status: 429, message: 'Daily limit exceeded' },

  // Server
  INTERNAL_ERROR: { status: 500, message: 'Internal server error' },
  SERVICE_UNAVAILABLE: { status: 503, message: 'Service temporarily unavailable' },
} as const;
```

## Not-Found Handler

Catch unmatched routes before the error middleware:

```typescript
// Mount AFTER all routes, BEFORE error handler
app.use((req, res) => {
  res.status(404).json({
    error: {
      code: 'ENDPOINT_NOT_FOUND',
      message: `${req.method} ${req.path} does not exist`,
      requestId: req.headers['x-request-id'] ?? crypto.randomUUID(),
    },
  });
});

app.use(errorHandler);
```

## Error Logging Best Practices

```typescript
// In the error handler, log appropriately
if (err instanceof AppError && err.statusCode < 500) {
  // 4xx — client errors. Log at info level (not the server's fault).
  console.info(`[${requestId}] ${err.code}: ${err.message}`, {
    statusCode: err.statusCode,
    path: req.path,
    method: req.method,
    userId: req.user?.id,
  });
} else {
  // 5xx — server errors. Log at error level with full stack.
  console.error(`[${requestId}] Unhandled error:`, {
    error: err.message,
    stack: err.stack,
    path: req.path,
    method: req.method,
    userId: req.user?.id,
    body: req.body, // Be careful with PII
  });
}
```

## Gotchas

- **Always return JSON for API errors.** Never let Express return its default HTML error page. The global error handler catches everything.
- **Express needs 4 parameters** in error middleware: `(err, req, res, next)`. If you forget `next`, Express won't recognize it as an error handler.
- **`express-async-errors` must be imported before Express.** It patches Express prototype methods. Import order matters.
- **Don't throw strings.** Always throw `Error` instances (or subclasses). `throw "bad"` gives you no stack trace.
- **Validate early, fail fast.** Use Zod validation middleware before the route handler. Don't validate inside the handler.
- **Prisma errors are not AppErrors.** Map them in the global handler (P2002 → 409, P2025 → 404). Don't let raw Prisma errors reach clients.
- **`requestId` enables debugging.** Log it server-side, return it to the client. When a user reports "error req_abc123", you can find it instantly.
- **4xx errors are not bugs.** Don't alert on them. Only alert on 5xx errors. A spike in 4xx might indicate a broken client, not a server issue.
