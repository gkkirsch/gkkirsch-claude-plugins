---
name: express-api
description: >
  Express.js API patterns — router setup, middleware composition, request validation,
  response helpers, file uploads, streaming, and production server configuration.
  Triggers: "express api", "express router", "express middleware", "express route",
  "REST api express", "express server setup", "express request handling".
  NOT for: architecture decisions (use api-architect agent), security (use security-engineer agent).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Express API Patterns

## Server Setup

```typescript
// src/server.ts
import app from './app';
import { env } from './config/env';
import { logger } from './lib/logger';

const server = app.listen(env.PORT, () => {
  logger.info(`Server running on port ${env.PORT}`);
});

// Graceful shutdown
const shutdown = (signal: string) => {
  logger.info(`${signal} received. Starting graceful shutdown...`);
  server.close(() => {
    logger.info('HTTP server closed');
    process.exit(0);
  });

  // Force shutdown after 10s
  setTimeout(() => {
    logger.error('Forced shutdown after timeout');
    process.exit(1);
  }, 10_000);
};

process.on('SIGTERM', () => shutdown('SIGTERM'));
process.on('SIGINT', () => shutdown('SIGINT'));
```

```typescript
// src/app.ts
import express from 'express';
import helmet from 'helmet';
import cors from 'cors';
import { router } from './routes';
import { errorHandler } from './middleware/error-handler';
import { requestLogger } from './middleware/request-logger';
import { rateLimiter } from './middleware/rate-limit';

const app = express();

// Middleware (order matters)
app.use(helmet());
app.use(cors({ origin: process.env.CORS_ORIGIN, credentials: true }));
app.use(express.json({ limit: '1mb' }));
app.use(express.urlencoded({ extended: true }));
app.use(requestLogger());
app.use(rateLimiter());

// Routes
app.use('/api', router);

// Health check (before auth, outside /api)
app.get('/health', (_req, res) => {
  res.json({ status: 'ok', timestamp: new Date().toISOString() });
});

// Error handling (must be last)
app.use(errorHandler);

export default app;
```

## Environment Validation

```typescript
// src/config/env.ts
import { z } from 'zod';

const envSchema = z.object({
  NODE_ENV: z.enum(['development', 'production', 'test']).default('development'),
  PORT: z.coerce.number().default(3000),
  DATABASE_URL: z.string().url(),
  JWT_SECRET: z.string().min(32),
  JWT_EXPIRES_IN: z.string().default('15m'),
  CORS_ORIGIN: z.string().default('http://localhost:5173'),
  LOG_LEVEL: z.enum(['fatal', 'error', 'warn', 'info', 'debug', 'trace']).default('info'),
});

export const env = envSchema.parse(process.env);
export type Env = z.infer<typeof envSchema>;
```

## Router Organization

```typescript
// src/routes/index.ts
import { Router } from 'express';
import { authRoutes } from './auth.routes';
import { userRoutes } from './users.routes';
import { postRoutes } from './posts.routes';
import { authenticate } from '../middleware/auth';

export const router = Router();

// Public routes
router.use('/auth', authRoutes);

// Protected routes
router.use('/users', authenticate, userRoutes);
router.use('/posts', authenticate, postRoutes);
```

```typescript
// src/routes/posts.routes.ts
import { Router } from 'express';
import { PostController } from '../controllers/posts.controller';
import { validate } from '../middleware/validate';
import { createPostSchema, updatePostSchema, listPostsSchema } from '../schemas/posts.schema';

const router = Router();
const controller = new PostController();

router.get('/', validate(listPostsSchema, 'query'), controller.list);
router.get('/:id', controller.getById);
router.post('/', validate(createPostSchema), controller.create);
router.put('/:id', validate(updatePostSchema), controller.update);
router.delete('/:id', controller.delete);

export { router as postRoutes };
```

## Controller Pattern

```typescript
// src/controllers/posts.controller.ts
import { Request, Response, NextFunction } from 'express';
import { PostService } from '../services/posts.service';
import { NotFoundError } from '../lib/errors';

export class PostController {
  private service = new PostService();

  list = async (req: Request, res: Response, next: NextFunction) => {
    try {
      const { page = 1, limit = 20, sort = 'createdAt' } = req.query;
      const result = await this.service.list({
        page: Number(page),
        limit: Number(limit),
        sort: String(sort),
        userId: req.user!.id,
      });

      res.json({
        success: true,
        data: result.items,
        meta: {
          page: result.page,
          limit: result.limit,
          total: result.total,
          totalPages: result.totalPages,
        },
      });
    } catch (err) {
      next(err);
    }
  };

  getById = async (req: Request, res: Response, next: NextFunction) => {
    try {
      const post = await this.service.findById(req.params.id);
      if (!post) throw new NotFoundError('Post');

      res.json({ success: true, data: post });
    } catch (err) {
      next(err);
    }
  };

  create = async (req: Request, res: Response, next: NextFunction) => {
    try {
      const post = await this.service.create({
        ...req.body,
        userId: req.user!.id,
      });

      res.status(201).json({ success: true, data: post });
    } catch (err) {
      next(err);
    }
  };

  update = async (req: Request, res: Response, next: NextFunction) => {
    try {
      const post = await this.service.update(req.params.id, req.body, req.user!.id);
      res.json({ success: true, data: post });
    } catch (err) {
      next(err);
    }
  };

  delete = async (req: Request, res: Response, next: NextFunction) => {
    try {
      await this.service.delete(req.params.id, req.user!.id);
      res.status(204).send();
    } catch (err) {
      next(err);
    }
  };
}
```

## Validation Middleware

```typescript
// src/middleware/validate.ts
import { Request, Response, NextFunction } from 'express';
import { ZodSchema, ZodError } from 'zod';

type ValidationTarget = 'body' | 'query' | 'params';

export function validate(schema: ZodSchema, target: ValidationTarget = 'body') {
  return (req: Request, res: Response, next: NextFunction) => {
    try {
      const parsed = schema.parse(req[target]);
      req[target] = parsed; // Replace with parsed (coerced) values
      next();
    } catch (err) {
      if (err instanceof ZodError) {
        const details: Record<string, string[]> = {};
        for (const issue of err.issues) {
          const path = issue.path.join('.');
          if (!details[path]) details[path] = [];
          details[path].push(issue.message);
        }

        res.status(400).json({
          success: false,
          error: {
            code: 'VALIDATION_ERROR',
            message: 'Invalid input',
            details,
          },
        });
        return;
      }
      next(err);
    }
  };
}
```

## Request Schemas

```typescript
// src/schemas/posts.schema.ts
import { z } from 'zod';

export const createPostSchema = z.object({
  title: z.string().min(1, 'Title is required').max(200),
  content: z.string().min(1, 'Content is required'),
  published: z.boolean().default(false),
  tags: z.array(z.string()).max(10).default([]),
});

export const updatePostSchema = createPostSchema.partial();

export const listPostsSchema = z.object({
  page: z.coerce.number().int().positive().default(1),
  limit: z.coerce.number().int().min(1).max(100).default(20),
  sort: z.enum(['createdAt', 'updatedAt', 'title']).default('createdAt'),
  order: z.enum(['asc', 'desc']).default('desc'),
  search: z.string().optional(),
});

// Common schemas
export const paginationSchema = z.object({
  page: z.coerce.number().int().positive().default(1),
  limit: z.coerce.number().int().min(1).max(100).default(20),
});

export const idParamSchema = z.object({
  id: z.string().uuid('Invalid ID format'),
});

export type CreatePostInput = z.infer<typeof createPostSchema>;
export type UpdatePostInput = z.infer<typeof updatePostSchema>;
export type ListPostsQuery = z.infer<typeof listPostsSchema>;
```

## Rate Limiting

```typescript
// src/middleware/rate-limit.ts
import rateLimit from 'express-rate-limit';

// General API rate limit
export const rateLimiter = () =>
  rateLimit({
    windowMs: 15 * 60 * 1000, // 15 minutes
    max: 100,
    standardHeaders: true,
    legacyHeaders: false,
    message: {
      success: false,
      error: {
        code: 'RATE_LIMIT_EXCEEDED',
        message: 'Too many requests, please try again later',
      },
    },
  });

// Strict rate limit for auth endpoints
export const authLimiter = rateLimit({
  windowMs: 15 * 60 * 1000,
  max: 10,
  standardHeaders: true,
  legacyHeaders: false,
  message: {
    success: false,
    error: {
      code: 'RATE_LIMIT_EXCEEDED',
      message: 'Too many authentication attempts',
    },
  },
});

// Per-user rate limit (requires authentication)
export const userLimiter = rateLimit({
  windowMs: 60 * 1000, // 1 minute
  max: 30,
  keyGenerator: (req) => req.user?.id || req.ip || 'unknown',
  standardHeaders: true,
});
```

## File Upload

```typescript
// src/middleware/upload.ts
import multer from 'multer';
import path from 'path';
import crypto from 'crypto';

const ALLOWED_TYPES = ['image/jpeg', 'image/png', 'image/webp', 'application/pdf'];
const MAX_FILE_SIZE = 10 * 1024 * 1024; // 10MB

const storage = multer.diskStorage({
  destination: './uploads',
  filename: (_req, file, cb) => {
    const ext = path.extname(file.originalname);
    const name = crypto.randomUUID();
    cb(null, `${name}${ext}`);
  },
});

export const upload = multer({
  storage,
  limits: { fileSize: MAX_FILE_SIZE },
  fileFilter: (_req, file, cb) => {
    if (ALLOWED_TYPES.includes(file.mimetype)) {
      cb(null, true);
    } else {
      cb(new Error(`File type ${file.mimetype} not allowed`));
    }
  },
});

// Usage in routes:
// router.post('/avatar', upload.single('avatar'), controller.uploadAvatar);
// router.post('/gallery', upload.array('images', 10), controller.uploadGallery);
```

## Response Helpers

```typescript
// src/lib/response.ts
import { Response } from 'express';

export function success<T>(res: Response, data: T, status = 200) {
  return res.status(status).json({ success: true, data });
}

export function created<T>(res: Response, data: T) {
  return success(res, data, 201);
}

export function noContent(res: Response) {
  return res.status(204).send();
}

export function paginated<T>(
  res: Response,
  items: T[],
  meta: { page: number; limit: number; total: number },
) {
  return res.json({
    success: true,
    data: items,
    meta: {
      ...meta,
      totalPages: Math.ceil(meta.total / meta.limit),
    },
  });
}
```

## Request Logger

```typescript
// src/middleware/request-logger.ts
import { Request, Response, NextFunction } from 'express';
import { logger } from '../lib/logger';

export function requestLogger() {
  return (req: Request, res: Response, next: NextFunction) => {
    const start = Date.now();

    res.on('finish', () => {
      const duration = Date.now() - start;
      const logData = {
        method: req.method,
        url: req.originalUrl,
        status: res.statusCode,
        duration: `${duration}ms`,
        ip: req.ip,
        userAgent: req.get('user-agent'),
      };

      if (res.statusCode >= 400) {
        logger.warn(logData, 'Request failed');
      } else {
        logger.info(logData, 'Request completed');
      }
    });

    next();
  };
}
```

## SSE Streaming

```typescript
// Server-Sent Events for real-time updates
import { Request, Response } from 'express';

export function sseHandler(req: Request, res: Response) {
  res.writeHead(200, {
    'Content-Type': 'text/event-stream',
    'Cache-Control': 'no-cache',
    Connection: 'keep-alive',
  });

  // Send initial connection event
  res.write('event: connected\ndata: {"status":"connected"}\n\n');

  // Keep-alive ping every 30s
  const keepAlive = setInterval(() => {
    res.write(': ping\n\n');
  }, 30_000);

  // Send events
  const sendEvent = (event: string, data: unknown) => {
    res.write(`event: ${event}\ndata: ${JSON.stringify(data)}\n\n`);
  };

  // Subscribe to events (example with EventEmitter)
  const onUpdate = (data: unknown) => sendEvent('update', data);
  eventBus.on('post:updated', onUpdate);

  // Cleanup on disconnect
  req.on('close', () => {
    clearInterval(keepAlive);
    eventBus.off('post:updated', onUpdate);
  });
}
```

## Express Type Extensions

```typescript
// src/types/express.d.ts
import { User } from '@prisma/client';

declare global {
  namespace Express {
    interface Request {
      user?: {
        id: string;
        email: string;
        role: 'user' | 'admin';
      };
      requestId?: string;
    }
  }
}
```

## Gotchas

1. **Middleware order matters.** `express.json()` must come before route handlers. Error handler must be the LAST `app.use()` call. Auth middleware goes on protected route groups, not globally.

2. **`async` route handlers need error forwarding.** Express doesn't catch async errors automatically. Either wrap in try/catch with `next(err)`, use express-async-errors package, or use a wrapper: `const asyncHandler = (fn) => (req, res, next) => fn(req, res, next).catch(next)`.

3. **`req.body` is `undefined` without body parser.** Always add `express.json()` middleware. Set `limit` to prevent large payload attacks.

4. **CORS preflight needs `OPTIONS` handling.** The `cors()` middleware handles this, but if you have custom middleware before it that returns early, preflight requests may fail.

5. **`res.json()` after `res.json()` silently fails.** Only the first response is sent. Use `return res.json(...)` to prevent accidentally calling it twice.

6. **`req.params.id` is always a string.** Even if the URL looks numeric. Parse it: `parseInt(req.params.id, 10)` or validate with Zod `z.coerce.number()`.
