---
name: trpc-server
description: >
  tRPC server setup — router definition, procedures, middleware, context, input validation,
  error handling, and Next.js App Router integration.
  Triggers: "trpc router", "trpc procedure", "trpc server", "trpc middleware",
  "trpc context", "trpc setup", "trpc next.js server", "trpc init".
  NOT for: client-side usage (use trpc-client), advanced patterns (use trpc-patterns).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# tRPC Server

## Installation

```bash
npm install @trpc/server @trpc/client @trpc/react-query @trpc/next @tanstack/react-query zod superjson
```

## Base Setup

```typescript
// src/server/trpc.ts
import { initTRPC, TRPCError } from '@trpc/server';
import superjson from 'superjson';
import { ZodError } from 'zod';
import { db } from '@/lib/db';
import { getServerSession } from '@/lib/auth';

// 1. Context — available in every procedure
export const createTRPCContext = async (opts: { headers: Headers }) => {
  const session = await getServerSession();
  return {
    db,
    session,
    headers: opts.headers,
  };
};

type Context = Awaited<ReturnType<typeof createTRPCContext>>;

// 2. Initialize tRPC
const t = initTRPC.context<Context>().create({
  transformer: superjson, // Handles Date, Map, Set, BigInt serialization
  errorFormatter({ shape, error }) {
    return {
      ...shape,
      data: {
        ...shape.data,
        zodError:
          error.cause instanceof ZodError ? error.cause.flatten() : null,
      },
    };
  },
});

// 3. Export reusable pieces
export const createCallerFactory = t.createCallerFactory;
export const createTRPCRouter = t.router;

// 4. Public procedure — no auth required
export const publicProcedure = t.procedure;

// 5. Protected procedure — requires authentication
export const protectedProcedure = t.procedure.use(({ ctx, next }) => {
  if (!ctx.session?.user) {
    throw new TRPCError({ code: 'UNAUTHORIZED' });
  }
  return next({
    ctx: {
      session: { ...ctx.session, user: ctx.session.user },
    },
  });
});

// 6. Admin procedure — requires admin role
export const adminProcedure = protectedProcedure.use(({ ctx, next }) => {
  if (ctx.session.user.role !== 'admin') {
    throw new TRPCError({ code: 'FORBIDDEN', message: 'Admin access required' });
  }
  return next({ ctx });
});
```

## Router Definition

```typescript
// src/server/routers/post.ts
import { z } from 'zod';
import { createTRPCRouter, publicProcedure, protectedProcedure } from '../trpc';
import { TRPCError } from '@trpc/server';

export const postRouter = createTRPCRouter({
  // Query — read data
  list: publicProcedure
    .input(z.object({
      page: z.number().min(1).default(1),
      limit: z.number().min(1).max(100).default(20),
      search: z.string().optional(),
    }))
    .query(async ({ ctx, input }) => {
      const { page, limit, search } = input;
      const offset = (page - 1) * limit;

      const where = {
        published: true,
        ...(search && {
          title: { contains: search, mode: 'insensitive' as const },
        }),
      };

      const [posts, total] = await ctx.db.$transaction([
        ctx.db.post.findMany({
          where,
          skip: offset,
          take: limit,
          orderBy: { createdAt: 'desc' },
          include: { author: { select: { name: true, image: true } } },
        }),
        ctx.db.post.count({ where }),
      ]);

      return {
        items: posts,
        meta: { total, page, limit, totalPages: Math.ceil(total / limit) },
      };
    }),

  // Query by ID
  byId: publicProcedure
    .input(z.object({ id: z.string() }))
    .query(async ({ ctx, input }) => {
      const post = await ctx.db.post.findUnique({
        where: { id: input.id },
        include: {
          author: { select: { name: true, image: true } },
          comments: {
            orderBy: { createdAt: 'desc' },
            include: { author: { select: { name: true } } },
          },
        },
      });

      if (!post) {
        throw new TRPCError({ code: 'NOT_FOUND', message: 'Post not found' });
      }

      return post;
    }),

  // Mutation — create
  create: protectedProcedure
    .input(z.object({
      title: z.string().min(1).max(200),
      content: z.string().min(1),
      published: z.boolean().default(false),
    }))
    .mutation(async ({ ctx, input }) => {
      const slug = input.title
        .toLowerCase()
        .replace(/[^\w\s-]/g, '')
        .replace(/\s+/g, '-');

      return ctx.db.post.create({
        data: {
          ...input,
          slug,
          authorId: ctx.session.user.id,
        },
      });
    }),

  // Mutation — update
  update: protectedProcedure
    .input(z.object({
      id: z.string(),
      title: z.string().min(1).max(200).optional(),
      content: z.string().min(1).optional(),
      published: z.boolean().optional(),
    }))
    .mutation(async ({ ctx, input }) => {
      const { id, ...data } = input;

      const post = await ctx.db.post.findUnique({ where: { id } });
      if (!post) throw new TRPCError({ code: 'NOT_FOUND' });
      if (post.authorId !== ctx.session.user.id) {
        throw new TRPCError({ code: 'FORBIDDEN' });
      }

      return ctx.db.post.update({ where: { id }, data });
    }),

  // Mutation — delete
  delete: protectedProcedure
    .input(z.object({ id: z.string() }))
    .mutation(async ({ ctx, input }) => {
      const post = await ctx.db.post.findUnique({ where: { id: input.id } });
      if (!post) throw new TRPCError({ code: 'NOT_FOUND' });
      if (post.authorId !== ctx.session.user.id) {
        throw new TRPCError({ code: 'FORBIDDEN' });
      }

      await ctx.db.post.delete({ where: { id: input.id } });
      return { success: true };
    }),
});
```

## Root Router

```typescript
// src/server/routers/_app.ts
import { createTRPCRouter } from '../trpc';
import { postRouter } from './post';
import { userRouter } from './user';
import { authRouter } from './auth';

export const appRouter = createTRPCRouter({
  post: postRouter,
  user: userRouter,
  auth: authRouter,
});

export type AppRouter = typeof appRouter;
```

## Next.js App Router Integration

```typescript
// src/app/api/trpc/[trpc]/route.ts
import { fetchRequestHandler } from '@trpc/server/adapters/fetch';
import { appRouter } from '@/server/routers/_app';
import { createTRPCContext } from '@/server/trpc';

const handler = (req: Request) =>
  fetchRequestHandler({
    endpoint: '/api/trpc',
    req,
    router: appRouter,
    createContext: () => createTRPCContext({ headers: req.headers }),
    onError:
      process.env.NODE_ENV === 'development'
        ? ({ path, error }) => {
            console.error(`tRPC error on ${path ?? '<no-path>'}:`, error);
          }
        : undefined,
  });

export { handler as GET, handler as POST };
```

## Middleware

```typescript
// Logging middleware
const loggerMiddleware = t.middleware(async ({ path, type, next }) => {
  const start = Date.now();
  const result = await next();
  const duration = Date.now() - start;

  if (result.ok) {
    console.log(`${type} ${path} — ${duration}ms`);
  } else {
    console.error(`${type} ${path} — ${duration}ms — ERROR`);
  }

  return result;
});

// Rate limiting middleware
const rateLimitMiddleware = t.middleware(async ({ ctx, next }) => {
  const ip = ctx.headers.get('x-forwarded-for') ?? 'unknown';
  const key = `rate-limit:${ip}`;

  // Check rate limit (implement with Redis or in-memory)
  const requests = await checkRateLimit(key, { max: 100, window: 60 });
  if (requests > 100) {
    throw new TRPCError({
      code: 'TOO_MANY_REQUESTS',
      message: 'Rate limit exceeded',
    });
  }

  return next();
});

// Apply to procedures
export const publicProcedure = t.procedure.use(loggerMiddleware);
export const rateLimitedProcedure = t.procedure
  .use(loggerMiddleware)
  .use(rateLimitMiddleware);
```

## Error Handling

```typescript
import { TRPCError } from '@trpc/server';

// Throw tRPC errors with appropriate codes
throw new TRPCError({
  code: 'NOT_FOUND',
  message: 'Post not found',
  cause: originalError, // Optional — original error for debugging
});

// Custom error formatter (in trpc.ts init)
errorFormatter({ shape, error }) {
  return {
    ...shape,
    data: {
      ...shape.data,
      zodError: error.cause instanceof ZodError ? error.cause.flatten() : null,
      // Zod errors get flattened for easy client-side field mapping
    },
  };
},
```

## Standalone Express Adapter

```typescript
// server.ts — if not using Next.js
import express from 'express';
import cors from 'cors';
import { createExpressMiddleware } from '@trpc/server/adapters/express';
import { appRouter } from './routers/_app';
import { createTRPCContext } from './trpc';

const app = express();
app.use(cors());

app.use(
  '/trpc',
  createExpressMiddleware({
    router: appRouter,
    createContext: ({ req }) => createTRPCContext({ headers: new Headers(req.headers as any) }),
  }),
);

app.listen(3000, () => console.log('Server running on :3000'));
```

## Gotchas

1. **Always use `superjson` transformer.** Without it, `Date` objects serialize as ISO strings and lose their type on the client. `superjson` handles Date, Map, Set, BigInt, undefined, RegExp, and more.

2. **Input validation is required.** Never skip `.input()` — even if you think the input is simple. Zod validation is your type safety guarantee at runtime.

3. **Middleware order matters.** Middleware executes in the order you `.use()` it. Put logging first, then rate limiting, then auth.

4. **`ctx` is immutable per middleware.** Each middleware returns a new `ctx` via `next({ ctx: newCtx })`. You can't mutate the original context object.

5. **Don't import server code on the client.** The `AppRouter` type is the ONLY thing you should import from the server on the client side. Importing actual router code will bundle server dependencies into the client.

6. **Batch HTTP requests.** tRPC batches concurrent requests by default. If you call 3 queries in the same render, they go as 1 HTTP request. Don't manually batch.
