---
name: error-handling
description: >
  Node.js/Express error handling patterns — custom error classes, error middleware,
  async error wrapping, structured error responses, and production logging.
  Triggers: "error handling express", "error middleware", "custom error class",
  "async error", "error response format", "express error handler".
  NOT for: frontend error boundaries (use React patterns), validation errors (use zod-schemas).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Error Handling Patterns

## Custom Error Classes

```typescript
// src/lib/errors.ts
export class AppError extends Error {
  constructor(
    public statusCode: number,
    public code: string,
    message: string,
    public details?: Record<string, unknown>,
  ) {
    super(message);
    this.name = 'AppError';
    Error.captureStackTrace(this, this.constructor);
  }
}

export class NotFoundError extends AppError {
  constructor(resource = 'Resource') {
    super(404, 'NOT_FOUND', `${resource} not found`);
  }
}

export class UnauthorizedError extends AppError {
  constructor(message = 'Unauthorized') {
    super(401, 'UNAUTHORIZED', message);
  }
}

export class ForbiddenError extends AppError {
  constructor(message = 'Forbidden') {
    super(403, 'FORBIDDEN', message);
  }
}

export class ConflictError extends AppError {
  constructor(message = 'Resource already exists') {
    super(409, 'CONFLICT', message);
  }
}

export class ValidationError extends AppError {
  constructor(details: Record<string, string[]>) {
    super(400, 'VALIDATION_ERROR', 'Invalid input', details);
  }
}

export class RateLimitError extends AppError {
  constructor(retryAfter?: number) {
    super(429, 'RATE_LIMIT_EXCEEDED', 'Too many requests', { retryAfter });
  }
}

export class ServiceUnavailableError extends AppError {
  constructor(service: string) {
    super(503, 'SERVICE_UNAVAILABLE', `${service} is temporarily unavailable`);
  }
}
```

## Error Handler Middleware

```typescript
// src/middleware/error-handler.ts
import { Request, Response, NextFunction } from 'express';
import { AppError } from '../lib/errors';
import { logger } from '../lib/logger';
import { ZodError } from 'zod';
import { Prisma } from '@prisma/client';

export function errorHandler(
  err: Error,
  req: Request,
  res: Response,
  _next: NextFunction,
) {
  // Already sent response (e.g., streaming)
  if (res.headersSent) {
    return _next(err);
  }

  // Custom application errors
  if (err instanceof AppError) {
    logger.warn({
      err,
      method: req.method,
      url: req.originalUrl,
      statusCode: err.statusCode,
    });

    return res.status(err.statusCode).json({
      success: false,
      error: {
        code: err.code,
        message: err.message,
        ...(err.details && { details: err.details }),
      },
    });
  }

  // Zod validation errors
  if (err instanceof ZodError) {
    const details: Record<string, string[]> = {};
    for (const issue of err.issues) {
      const path = issue.path.join('.') || '_root';
      if (!details[path]) details[path] = [];
      details[path].push(issue.message);
    }

    return res.status(400).json({
      success: false,
      error: {
        code: 'VALIDATION_ERROR',
        message: 'Invalid input',
        details,
      },
    });
  }

  // Prisma errors
  if (err instanceof Prisma.PrismaClientKnownRequestError) {
    return handlePrismaError(err, res);
  }

  // JWT errors
  if (err.name === 'JsonWebTokenError') {
    return res.status(401).json({
      success: false,
      error: { code: 'INVALID_TOKEN', message: 'Invalid token' },
    });
  }

  if (err.name === 'TokenExpiredError') {
    return res.status(401).json({
      success: false,
      error: { code: 'TOKEN_EXPIRED', message: 'Token has expired' },
    });
  }

  // Multer file upload errors
  if (err.name === 'MulterError') {
    const multerMessages: Record<string, string> = {
      LIMIT_FILE_SIZE: 'File too large',
      LIMIT_FILE_COUNT: 'Too many files',
      LIMIT_UNEXPECTED_FILE: 'Unexpected file field',
    };

    return res.status(400).json({
      success: false,
      error: {
        code: 'FILE_UPLOAD_ERROR',
        message: multerMessages[(err as any).code] || 'File upload error',
      },
    });
  }

  // Syntax errors (malformed JSON body)
  if (err instanceof SyntaxError && 'body' in err) {
    return res.status(400).json({
      success: false,
      error: { code: 'INVALID_JSON', message: 'Malformed JSON in request body' },
    });
  }

  // Unknown errors — log full stack, return generic message
  logger.error({
    err,
    method: req.method,
    url: req.originalUrl,
    stack: err.stack,
  });

  res.status(500).json({
    success: false,
    error: {
      code: 'INTERNAL_ERROR',
      message:
        process.env.NODE_ENV === 'production'
          ? 'An unexpected error occurred'
          : err.message,
    },
  });
}

function handlePrismaError(
  err: Prisma.PrismaClientKnownRequestError,
  res: Response,
) {
  switch (err.code) {
    case 'P2002': {
      const field = (err.meta?.target as string[])?.join(', ') || 'field';
      return res.status(409).json({
        success: false,
        error: {
          code: 'DUPLICATE_ENTRY',
          message: `A record with this ${field} already exists`,
        },
      });
    }
    case 'P2025':
      return res.status(404).json({
        success: false,
        error: { code: 'NOT_FOUND', message: 'Record not found' },
      });
    case 'P2003':
      return res.status(400).json({
        success: false,
        error: { code: 'FOREIGN_KEY_ERROR', message: 'Referenced record does not exist' },
      });
    default:
      return res.status(500).json({
        success: false,
        error: { code: 'DATABASE_ERROR', message: 'A database error occurred' },
      });
  }
}
```

## Async Error Wrapper

```typescript
// src/lib/async-handler.ts
import { Request, Response, NextFunction, RequestHandler } from 'express';

// Wraps async route handlers to forward errors to Express error middleware
export function asyncHandler(
  fn: (req: Request, res: Response, next: NextFunction) => Promise<any>,
): RequestHandler {
  return (req, res, next) => {
    Promise.resolve(fn(req, res, next)).catch(next);
  };
}

// Usage — no more try/catch in every handler:
// router.get('/:id', asyncHandler(async (req, res) => {
//   const post = await postService.findById(req.params.id);
//   if (!post) throw new NotFoundError('Post');
//   res.json({ success: true, data: post });
// }));
```

## Structured Logger

```typescript
// src/lib/logger.ts
import pino from 'pino';

export const logger = pino({
  level: process.env.LOG_LEVEL || 'info',
  transport:
    process.env.NODE_ENV === 'development'
      ? { target: 'pino-pretty', options: { colorize: true } }
      : undefined,
  redact: {
    paths: ['req.headers.authorization', 'req.body.password', 'req.body.refreshToken'],
    censor: '[REDACTED]',
  },
  serializers: {
    err: pino.stdSerializers.err,
    req: (req) => ({
      method: req.method,
      url: req.url,
      remoteAddress: req.remoteAddress,
    }),
    res: (res) => ({
      statusCode: res.statusCode,
    }),
  },
});

// Request-scoped logger with request ID
export function createRequestLogger(requestId: string) {
  return logger.child({ requestId });
}
```

## Error Response Format

All error responses follow this shape:

```typescript
interface ErrorResponse {
  success: false;
  error: {
    code: string;           // Machine-readable error code (UPPER_SNAKE_CASE)
    message: string;        // Human-readable message
    details?: unknown;      // Additional context (validation errors, retry info)
  };
}

// Examples:
// 400 Validation
{ success: false, error: { code: "VALIDATION_ERROR", message: "Invalid input", details: { email: ["Invalid email format"] } } }

// 401 Unauthorized
{ success: false, error: { code: "UNAUTHORIZED", message: "Invalid credentials" } }

// 404 Not Found
{ success: false, error: { code: "NOT_FOUND", message: "Post not found" } }

// 409 Conflict
{ success: false, error: { code: "DUPLICATE_ENTRY", message: "A record with this email already exists" } }

// 429 Rate Limit
{ success: false, error: { code: "RATE_LIMIT_EXCEEDED", message: "Too many requests", details: { retryAfter: 60 } } }

// 500 Internal
{ success: false, error: { code: "INTERNAL_ERROR", message: "An unexpected error occurred" } }
```

## Not Found Handler (404 for unknown routes)

```typescript
// Add BEFORE error handler, AFTER all routes
app.use((_req, res) => {
  res.status(404).json({
    success: false,
    error: { code: 'ROUTE_NOT_FOUND', message: 'The requested endpoint does not exist' },
  });
});
```

## Process-Level Error Handling

```typescript
// src/server.ts — add after server.listen()

// Unhandled promise rejections
process.on('unhandledRejection', (reason: unknown) => {
  logger.fatal({ err: reason }, 'Unhandled promise rejection');
  // Let the process crash — restart via process manager
  process.exit(1);
});

// Uncaught exceptions
process.on('uncaughtException', (err: Error) => {
  logger.fatal({ err }, 'Uncaught exception');
  // Attempt graceful shutdown
  server.close(() => process.exit(1));
  // Force kill after 5s
  setTimeout(() => process.exit(1), 5000);
});
```

## Gotchas

1. **Error handler must have 4 parameters.** Express identifies error middleware by the `(err, req, res, next)` signature. If you omit `next`, Express won't recognize it as an error handler. Use `_next` if unused.

2. **Error handler must be registered last.** `app.use(errorHandler)` goes after ALL route registrations. If it's before a route, errors from that route won't reach it.

3. **`express-async-errors` as an alternative.** Instead of wrapping every handler with `asyncHandler()`, you can `import 'express-async-errors'` at the top of your app. It monkey-patches Express to catch async errors automatically.

4. **Never expose stack traces in production.** Check `NODE_ENV` before including `err.stack` or `err.message` in responses. Unknown errors should return generic messages.

5. **`res.headersSent` check is critical.** If you've already started sending a response (e.g., SSE streaming) and an error occurs, calling `res.json()` will throw. Check `res.headersSent` first and call `_next(err)` to let Express handle it.

6. **Prisma error codes are strings, not numbers.** `P2002` is a unique constraint violation, `P2025` is record not found, `P2003` is foreign key constraint. Always handle these in your error middleware.

7. **Log redaction is essential.** Use `pino`'s `redact` option to automatically censor passwords, tokens, and API keys from log output. This prevents credentials from appearing in log aggregation services.
