---
name: rest-api-patterns
description: >
  REST API implementation patterns — Express router setup, CRUD endpoints, pagination,
  filtering, sorting, error handling middleware, and response formatting.
  Triggers: "REST API", "Express routes", "CRUD endpoints", "API pagination",
  "API filtering", "API error handling", "Express middleware".
  NOT for: GraphQL (use graphql-patterns), API versioning strategy (use api-versioning).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# REST API Patterns

## Express Router Setup

```typescript
import { Router, Request, Response, NextFunction } from 'express';

const router = Router();

// Async handler wrapper (catches async errors)
function asyncHandler(fn: (req: Request, res: Response, next: NextFunction) => Promise<any>) {
  return (req: Request, res: Response, next: NextFunction) => {
    fn(req, res, next).catch(next);
  };
}

// CRUD routes
router.get('/', asyncHandler(listUsers));
router.get('/:id', asyncHandler(getUser));
router.post('/', validate(createUserSchema), asyncHandler(createUser));
router.put('/:id', validate(updateUserSchema), asyncHandler(updateUser));
router.patch('/:id', validate(patchUserSchema), asyncHandler(patchUser));
router.delete('/:id', asyncHandler(deleteUser));

export default router;

// Mount in app
app.use('/api/users', authenticate, router);
```

## CRUD Endpoint Implementations

```typescript
// LIST with pagination, filtering, sorting
async function listUsers(req: Request, res: Response) {
  const {
    page = '1',
    limit = '20',
    sort = 'createdAt',
    order = 'desc',
    search,
    role,
  } = req.query as Record<string, string>;

  const pageNum = Math.max(1, parseInt(page));
  const limitNum = Math.min(100, Math.max(1, parseInt(limit)));
  const offset = (pageNum - 1) * limitNum;

  const where: any = {};
  if (search) {
    where.OR = [
      { name: { contains: search, mode: 'insensitive' } },
      { email: { contains: search, mode: 'insensitive' } },
    ];
  }
  if (role) where.role = role;

  const [users, total] = await Promise.all([
    db.user.findMany({
      where,
      orderBy: { [sort]: order },
      skip: offset,
      take: limitNum,
      select: { id: true, name: true, email: true, role: true, createdAt: true },
    }),
    db.user.count({ where }),
  ]);

  res.json({
    data: users,
    pagination: {
      page: pageNum,
      limit: limitNum,
      total,
      totalPages: Math.ceil(total / limitNum),
      hasMore: pageNum * limitNum < total,
    },
  });
}

// GET single resource
async function getUser(req: Request, res: Response) {
  const user = await db.user.findUnique({
    where: { id: req.params.id },
    select: { id: true, name: true, email: true, role: true, createdAt: true },
  });

  if (!user) {
    return res.status(404).json({
      error: { code: 'NOT_FOUND', message: 'User not found' },
    });
  }

  res.json({ data: user });
}

// CREATE
async function createUser(req: Request, res: Response) {
  const { email, name, role } = req.body;

  const existing = await db.user.findUnique({ where: { email } });
  if (existing) {
    return res.status(409).json({
      error: { code: 'CONFLICT', message: 'Email already registered' },
    });
  }

  const user = await db.user.create({
    data: { email, name, role: role || 'user' },
    select: { id: true, name: true, email: true, role: true, createdAt: true },
  });

  res.status(201)
    .location(`/api/users/${user.id}`)
    .json({ data: user });
}

// UPDATE (full replace)
async function updateUser(req: Request, res: Response) {
  const user = await db.user.findUnique({ where: { id: req.params.id } });
  if (!user) {
    return res.status(404).json({
      error: { code: 'NOT_FOUND', message: 'User not found' },
    });
  }

  const updated = await db.user.update({
    where: { id: req.params.id },
    data: req.body,
    select: { id: true, name: true, email: true, role: true, createdAt: true },
  });

  res.json({ data: updated });
}

// DELETE
async function deleteUser(req: Request, res: Response) {
  const user = await db.user.findUnique({ where: { id: req.params.id } });
  if (!user) {
    return res.status(404).json({
      error: { code: 'NOT_FOUND', message: 'User not found' },
    });
  }

  await db.user.delete({ where: { id: req.params.id } });
  res.status(204).send();
}
```

## Cursor-Based Pagination

```typescript
async function listPosts(req: Request, res: Response) {
  const { cursor, limit = '20' } = req.query as Record<string, string>;
  const take = Math.min(100, parseInt(limit));

  const posts = await db.post.findMany({
    take: take + 1, // Fetch one extra to determine hasMore
    ...(cursor && {
      cursor: { id: cursor },
      skip: 1, // Skip the cursor item itself
    }),
    orderBy: { createdAt: 'desc' },
    include: { author: { select: { id: true, name: true } } },
  });

  const hasMore = posts.length > take;
  const data = hasMore ? posts.slice(0, take) : posts;
  const nextCursor = hasMore ? data[data.length - 1].id : null;

  res.json({
    data,
    pagination: {
      nextCursor,
      hasMore,
    },
  });
}
```

## Error Handling Middleware

```typescript
// Custom error classes
class AppError extends Error {
  constructor(
    public statusCode: number,
    public code: string,
    message: string,
    public details?: any
  ) {
    super(message);
    this.name = 'AppError';
  }
}

class NotFoundError extends AppError {
  constructor(resource: string) {
    super(404, 'NOT_FOUND', `${resource} not found`);
  }
}

class ValidationError extends AppError {
  constructor(details: any) {
    super(400, 'VALIDATION_ERROR', 'Validation failed', details);
  }
}

class ConflictError extends AppError {
  constructor(message: string) {
    super(409, 'CONFLICT', message);
  }
}

// Global error handler
function errorHandler(err: Error, req: Request, res: Response, next: NextFunction) {
  // Zod validation errors
  if (err.name === 'ZodError') {
    return res.status(400).json({
      error: {
        code: 'VALIDATION_ERROR',
        message: 'Validation failed',
        details: (err as any).flatten().fieldErrors,
      },
    });
  }

  // Custom app errors
  if (err instanceof AppError) {
    return res.status(err.statusCode).json({
      error: {
        code: err.code,
        message: err.message,
        ...(err.details && { details: err.details }),
      },
    });
  }

  // Database unique constraint violations (Prisma)
  if ((err as any).code === 'P2002') {
    const field = (err as any).meta?.target?.[0];
    return res.status(409).json({
      error: {
        code: 'CONFLICT',
        message: `${field || 'Resource'} already exists`,
      },
    });
  }

  // Unexpected errors
  console.error('Unhandled error:', err);
  res.status(500).json({
    error: {
      code: 'INTERNAL_ERROR',
      message: process.env.NODE_ENV === 'production'
        ? 'An unexpected error occurred'
        : err.message,
    },
  });
}

app.use(errorHandler);
```

## Request Validation Middleware

```typescript
import { z, ZodSchema } from 'zod';

function validate(schema: ZodSchema) {
  return (req: Request, res: Response, next: NextFunction) => {
    const result = schema.safeParse(req.body);
    if (!result.success) {
      return res.status(400).json({
        error: {
          code: 'VALIDATION_ERROR',
          message: 'Validation failed',
          details: result.error.flatten().fieldErrors,
        },
      });
    }
    req.body = result.data;
    next();
  };
}

// Validate query params
function validateQuery(schema: ZodSchema) {
  return (req: Request, res: Response, next: NextFunction) => {
    const result = schema.safeParse(req.query);
    if (!result.success) {
      return res.status(400).json({
        error: {
          code: 'VALIDATION_ERROR',
          message: 'Invalid query parameters',
          details: result.error.flatten().fieldErrors,
        },
      });
    }
    req.query = result.data;
    next();
  };
}

// Schemas
const createUserSchema = z.object({
  email: z.string().email(),
  name: z.string().min(1).max(100),
  role: z.enum(['user', 'admin']).optional(),
});

const listQuerySchema = z.object({
  page: z.coerce.number().int().positive().optional().default(1),
  limit: z.coerce.number().int().min(1).max(100).optional().default(20),
  sort: z.enum(['createdAt', 'name', 'email']).optional().default('createdAt'),
  order: z.enum(['asc', 'desc']).optional().default('desc'),
  search: z.string().max(100).optional(),
});
```

## Response Envelope

```typescript
// Consistent response format
interface ApiResponse<T> {
  data: T;
  pagination?: {
    page: number;
    limit: number;
    total: number;
    totalPages: number;
    hasMore: boolean;
  };
}

interface ApiError {
  error: {
    code: string;
    message: string;
    details?: Record<string, string[]>;
  };
}

// Helper functions
function success<T>(res: Response, data: T, status = 200) {
  return res.status(status).json({ data });
}

function created<T>(res: Response, data: T, location: string) {
  return res.status(201).location(location).json({ data });
}

function paginated<T>(res: Response, data: T[], pagination: any) {
  return res.json({ data, pagination });
}

function noContent(res: Response) {
  return res.status(204).send();
}
```

## Nested Resource Routes

```typescript
// /api/users/:userId/posts
const postRouter = Router({ mergeParams: true }); // mergeParams to access :userId

postRouter.get('/', asyncHandler(async (req, res) => {
  const posts = await db.post.findMany({
    where: { authorId: req.params.userId },
    orderBy: { createdAt: 'desc' },
  });
  res.json({ data: posts });
}));

postRouter.post('/', validate(createPostSchema), asyncHandler(async (req, res) => {
  const post = await db.post.create({
    data: { ...req.body, authorId: req.params.userId },
  });
  res.status(201).json({ data: post });
}));

router.use('/:userId/posts', postRouter);

// For deeply nested, prefer flat routes with query params:
// GET /api/comments?postId=123&userId=456
// Instead of: GET /api/users/456/posts/123/comments
```

## Gotchas

1. **Express doesn't catch async errors by default.** Without `asyncHandler` (or `express-async-errors`), unhandled promise rejections crash the server. Always wrap async route handlers or use `express-async-errors` which patches Express to handle them.

2. **PUT should replace the entire resource.** Don't use PUT for partial updates — that's what PATCH is for. If the client sends `PUT { name: "Bob" }`, the email should be nulled/removed. Most apps actually want PATCH, not PUT.

3. **`req.params` values are always strings.** `req.params.id` is `"123"`, not `123`. Always parse/validate before using in queries. Zod's `z.coerce.number()` handles this for query params.

4. **Don't return database models directly.** They may contain password hashes, internal IDs, or metadata. Use `select` (Prisma) or create a serializer to control exactly what's returned.

5. **Pagination `total` count can be slow.** `COUNT(*)` on large tables is expensive, especially with complex WHERE clauses. For cursor pagination, consider returning just `hasMore` without `total`. Or cache the count.

6. **`mergeParams: true` is required for nested routers.** Without it, `req.params.userId` is undefined in child routers. This is the #1 gotcha with Express nested routes.
