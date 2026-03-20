---
name: trpc-patterns
description: >
  Advanced tRPC patterns — subscriptions, file uploads, testing, API versioning,
  output validation, server-sent events, and production deployment patterns.
  Triggers: "trpc subscription", "trpc websocket", "trpc upload", "trpc test",
  "trpc versioning", "trpc output", "trpc sse", "trpc production".
  NOT for: basic setup (use trpc-server), basic client usage (use trpc-client).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Advanced tRPC Patterns

## Subscriptions (WebSocket)

```typescript
// Server: WebSocket adapter
import { applyWSSHandler } from '@trpc/server/adapters/ws';
import ws from 'ws';
import { appRouter } from './routers/_app';
import { createTRPCContext } from './trpc';

const wss = new ws.Server({ port: 3001 });

const handler = applyWSSHandler({
  wss,
  router: appRouter,
  createContext: () => createTRPCContext({ headers: new Headers() }),
});

process.on('SIGTERM', () => {
  handler.broadcastReconnectNotification();
  wss.close();
});
```

```typescript
// Server: subscription procedure
import { observable } from '@trpc/server/observable';
import { EventEmitter } from 'events';

const ee = new EventEmitter();

export const chatRouter = createTRPCRouter({
  onMessage: publicProcedure
    .input(z.object({ roomId: z.string() }))
    .subscription(({ input }) => {
      return observable<{ id: string; text: string; userId: string }>((emit) => {
        const handler = (data: { roomId: string; message: any }) => {
          if (data.roomId === input.roomId) {
            emit.next(data.message);
          }
        };

        ee.on('message', handler);

        // Cleanup on unsubscribe
        return () => {
          ee.off('message', handler);
        };
      });
    }),

  sendMessage: protectedProcedure
    .input(z.object({
      roomId: z.string(),
      text: z.string().min(1).max(1000),
    }))
    .mutation(async ({ ctx, input }) => {
      const message = {
        id: crypto.randomUUID(),
        text: input.text,
        userId: ctx.session.user.id,
        createdAt: new Date(),
      };

      // Save to DB
      await ctx.db.message.create({ data: { ...message, roomId: input.roomId } });

      // Emit to subscribers
      ee.emit('message', { roomId: input.roomId, message });

      return message;
    }),
});
```

```tsx
// Client: subscribe to messages
function ChatRoom({ roomId }: { roomId: string }) {
  const [messages, setMessages] = useState<Message[]>([]);

  trpc.chat.onMessage.useSubscription(
    { roomId },
    {
      onData: (message) => {
        setMessages((prev) => [...prev, message]);
      },
      onError: (err) => {
        console.error('Subscription error:', err);
      },
    },
  );

  return (
    <div>
      {messages.map((msg) => (
        <div key={msg.id}>{msg.text}</div>
      ))}
    </div>
  );
}
```

## Server-Sent Events (SSE)

```typescript
// Server: SSE subscription (no WebSocket needed)
import { httpBatchStreamLink } from '@trpc/client';

// In Next.js route handler
// src/app/api/trpc/[trpc]/route.ts
import { fetchRequestHandler } from '@trpc/server/adapters/fetch';

const handler = (req: Request) =>
  fetchRequestHandler({
    endpoint: '/api/trpc',
    req,
    router: appRouter,
    createContext: () => createTRPCContext({ headers: req.headers }),
  });

export { handler as GET, handler as POST };
```

```typescript
// Client: use httpBatchStreamLink for SSE
import { httpBatchStreamLink } from '@trpc/client';

const trpcClient = trpc.createClient({
  links: [
    httpBatchStreamLink({
      url: '/api/trpc',
      transformer: superjson,
    }),
  ],
});
```

## Output Validation

```typescript
// Validate and narrow the output shape
const getUser = publicProcedure
  .input(z.object({ id: z.string() }))
  .output(z.object({
    id: z.string(),
    name: z.string(),
    email: z.string().email(),
    role: z.enum(['user', 'admin']),
    createdAt: z.date(),
    // Explicitly excludes password, internal fields
  }))
  .query(async ({ ctx, input }) => {
    const user = await ctx.db.user.findUnique({ where: { id: input.id } });
    if (!user) throw new TRPCError({ code: 'NOT_FOUND' });
    return user; // Zod strips extra fields (password, etc.)
  });
```

## File Uploads

```typescript
// tRPC doesn't handle file uploads natively.
// Use a separate REST endpoint and pass the URL to tRPC.

// 1. REST upload endpoint
// src/app/api/upload/route.ts
import { writeFile } from 'fs/promises';
import { nanoid } from 'nanoid';

export async function POST(req: Request) {
  const formData = await req.formData();
  const file = formData.get('file') as File;

  if (!file) return Response.json({ error: 'No file' }, { status: 400 });

  const bytes = await file.arrayBuffer();
  const buffer = Buffer.from(bytes);
  const filename = `${nanoid()}-${file.name}`;

  // Save to disk, S3, or Cloudinary
  await writeFile(`./uploads/${filename}`, buffer);

  return Response.json({ url: `/uploads/${filename}` });
}

// 2. tRPC mutation references the uploaded URL
const createPost = protectedProcedure
  .input(z.object({
    title: z.string(),
    content: z.string(),
    imageUrl: z.string().url().optional(), // From upload endpoint
  }))
  .mutation(async ({ ctx, input }) => {
    return ctx.db.post.create({ data: { ...input, authorId: ctx.session.user.id } });
  });

// 3. Client: upload file, then call tRPC
async function handleSubmit(formData: FormData) {
  // Upload file first
  const file = formData.get('image');
  let imageUrl: string | undefined;

  if (file) {
    const uploadForm = new FormData();
    uploadForm.set('file', file);
    const res = await fetch('/api/upload', { method: 'POST', body: uploadForm });
    const data = await res.json();
    imageUrl = data.url;
  }

  // Then create post with image URL
  createPost.mutate({
    title: formData.get('title') as string,
    content: formData.get('content') as string,
    imageUrl,
  });
}
```

## Testing tRPC Procedures

```typescript
// tests/routers/post.test.ts
import { describe, it, expect, beforeEach, vi } from 'vitest';
import { appRouter } from '@/server/routers/_app';
import { createCallerFactory } from '@/server/trpc';
import { db } from '@/lib/db';

const createCaller = createCallerFactory(appRouter);

describe('post router', () => {
  // Create caller with mock context
  const caller = createCaller({
    db,
    session: {
      user: { id: 'user-1', name: 'Test', email: 'test@test.com', role: 'user' },
      expires: new Date(Date.now() + 86400000).toISOString(),
    },
    headers: new Headers(),
  });

  // Unauthenticated caller
  const anonCaller = createCaller({
    db,
    session: null,
    headers: new Headers(),
  });

  beforeEach(async () => {
    await db.post.deleteMany();
    await db.user.deleteMany();
    await db.user.create({
      data: { id: 'user-1', name: 'Test', email: 'test@test.com', password: 'hash' },
    });
  });

  it('should create a post', async () => {
    const post = await caller.post.create({
      title: 'Test Post',
      content: 'Content here',
    });

    expect(post.title).toBe('Test Post');
    expect(post.slug).toBe('test-post');
    expect(post.authorId).toBe('user-1');
  });

  it('should require auth to create', async () => {
    await expect(
      anonCaller.post.create({ title: 'Test', content: 'Content' }),
    ).rejects.toThrow('UNAUTHORIZED');
  });

  it('should list posts with pagination', async () => {
    // Create 25 posts
    for (let i = 0; i < 25; i++) {
      await caller.post.create({
        title: `Post ${i}`,
        content: 'Content',
        published: true,
      });
    }

    const result = await caller.post.list({ page: 1, limit: 10 });

    expect(result.items).toHaveLength(10);
    expect(result.meta.total).toBe(25);
    expect(result.meta.totalPages).toBe(3);
  });

  it('should return NOT_FOUND for missing post', async () => {
    await expect(
      caller.post.byId({ id: 'nonexistent' }),
    ).rejects.toThrow('NOT_FOUND');
  });

  it('should prevent deleting other user posts', async () => {
    // Create post as user-1
    const post = await caller.post.create({
      title: 'Test',
      content: 'Content',
    });

    // Try deleting as user-2
    const otherCaller = createCaller({
      db,
      session: {
        user: { id: 'user-2', name: 'Other', email: 'other@test.com', role: 'user' },
        expires: new Date(Date.now() + 86400000).toISOString(),
      },
      headers: new Headers(),
    });

    await expect(
      otherCaller.post.delete({ id: post.id }),
    ).rejects.toThrow('FORBIDDEN');
  });
});
```

## API Versioning

```typescript
// src/server/routers/_app.ts
import { createTRPCRouter } from '../trpc';
import { postRouterV1 } from './v1/post';
import { postRouterV2 } from './v2/post';
import { userRouter } from './user';

export const appRouter = createTRPCRouter({
  // Current version (v2)
  post: postRouterV2,
  user: userRouter,

  // Legacy support
  v1: createTRPCRouter({
    post: postRouterV1,
  }),
});

// Client can access:
// trpc.post.list.useQuery()      — v2 (current)
// trpc.v1.post.list.useQuery()   — v1 (legacy)
```

## Shared Schemas (Server + Client)

```typescript
// src/schemas/post.ts — importable by both server and client
import { z } from 'zod';

export const createPostSchema = z.object({
  title: z.string().min(1, 'Title is required').max(200, 'Title too long'),
  content: z.string().min(1, 'Content is required'),
  published: z.boolean().default(false),
  tags: z.array(z.string()).max(10).default([]),
});

export const updatePostSchema = createPostSchema.partial().extend({
  id: z.string(),
});

export type CreatePostInput = z.infer<typeof createPostSchema>;
export type UpdatePostInput = z.infer<typeof updatePostSchema>;
```

```tsx
// Client: use the same schema for form validation
import { createPostSchema } from '@/schemas/post';

function PostForm() {
  const [errors, setErrors] = useState<Record<string, string>>({});

  const handleSubmit = (formData: FormData) => {
    const result = createPostSchema.safeParse({
      title: formData.get('title'),
      content: formData.get('content'),
    });

    if (!result.success) {
      setErrors(
        Object.fromEntries(
          Object.entries(result.error.flatten().fieldErrors)
            .map(([k, v]) => [k, v?.[0] ?? ''])
        ),
      );
      return;
    }

    // Input is validated, send to server
    createPost.mutate(result.data);
  };
}
```

## Middleware Composition

```typescript
// Composable middleware chain
const withTiming = t.middleware(async ({ path, next }) => {
  const start = performance.now();
  const result = await next();
  console.log(`${path}: ${(performance.now() - start).toFixed(1)}ms`);
  return result;
});

const withRateLimit = (max: number, windowSec: number) =>
  t.middleware(async ({ ctx, next }) => {
    // Rate limit implementation
    return next();
  });

const withOwnership = (getResourceUserId: (input: any) => Promise<string>) =>
  t.middleware(async ({ ctx, input, next }) => {
    const ownerId = await getResourceUserId(input);
    if (ownerId !== ctx.session.user.id) {
      throw new TRPCError({ code: 'FORBIDDEN' });
    }
    return next();
  });

// Compose into specialized procedures
export const publicProcedure = t.procedure.use(withTiming);

export const protectedProcedure = t.procedure
  .use(withTiming)
  .use(authMiddleware);

export const rateLimitedProcedure = t.procedure
  .use(withTiming)
  .use(withRateLimit(100, 60));

export const adminProcedure = protectedProcedure
  .use(adminMiddleware);
```

## Production Deployment Checklist

| Item | Details |
|------|---------|
| Error monitoring | Sentry, LogRocket, or similar in `onError` callback |
| Rate limiting | Per-IP and per-user limits on mutation procedures |
| Input size limits | Max string lengths, array sizes in Zod schemas |
| CORS configuration | Restrict origins in production |
| Batch limits | Set `maxBatchSize` in httpBatchLink (default: 10) |
| Request timeout | Set `requestTimeout` in the adapter |
| Health check | Add a simple `health` procedure for uptime monitoring |
| Logging | Structured logging in middleware (JSON, not console.log) |
| Cache headers | Set `Cache-Control` for public queries via `responseMeta` |

## Gotchas

1. **Subscriptions need a separate transport.** HTTP can't do subscriptions. You need WebSocket (`ws`) or SSE (`httpBatchStreamLink`). SSE is simpler but one-directional.

2. **File uploads must go through REST.** tRPC serializes everything as JSON. Binary data needs a separate upload endpoint — pass the resulting URL/key to tRPC.

3. **`createCallerFactory` is for testing and server components only.** Don't use it in API routes — use the adapter handler instead. The caller bypasses HTTP entirely.

4. **Middleware can't short-circuit with a response.** Unlike Express middleware, tRPC middleware must call `next()` or throw. You can't return a response directly.

5. **Output validation strips extra fields.** If you define `.output()`, Zod strips fields not in the schema. This is good for security (no password leaks) but surprising if you expect all DB fields.

6. **Batching can cause confusion in error handling.** When 3 queries batch into 1 HTTP request, one failure doesn't fail the others. Each query has independent error state.
