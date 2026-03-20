# tRPC Cheat Sheet

## Setup

```bash
npm install @trpc/server @trpc/client @trpc/react-query @trpc/next @tanstack/react-query zod superjson
```

## Server

```typescript
// trpc.ts
import { initTRPC, TRPCError } from '@trpc/server';
import superjson from 'superjson';

const t = initTRPC.context<Context>().create({ transformer: superjson });
export const router = t.router;
export const publicProcedure = t.procedure;
export const protectedProcedure = t.procedure.use(({ ctx, next }) => {
  if (!ctx.session?.user) throw new TRPCError({ code: 'UNAUTHORIZED' });
  return next({ ctx: { session: { ...ctx.session, user: ctx.session.user } } });
});
```

```typescript
// router
export const postRouter = router({
  list: publicProcedure
    .input(z.object({ page: z.number().default(1) }))
    .query(async ({ ctx, input }) => { /* ... */ }),

  create: protectedProcedure
    .input(z.object({ title: z.string().min(1), content: z.string() }))
    .mutation(async ({ ctx, input }) => { /* ... */ }),
});
```

## Client (React)

```typescript
const trpc = createTRPCReact<AppRouter>();

// Query
const { data, isLoading } = trpc.post.list.useQuery({ page: 1 });

// Mutation
const create = trpc.post.create.useMutation({
  onSuccess: () => utils.post.list.invalidate(),
});
create.mutate({ title: 'New', content: 'Content' });

// Optimistic update
const toggle = trpc.todo.toggle.useMutation({
  onMutate: async ({ id }) => {
    await utils.todo.list.cancel();
    const prev = utils.todo.list.getData();
    utils.todo.list.setData(undefined, (old) =>
      old?.map((t) => t.id === id ? { ...t, done: !t.done } : t));
    return { prev };
  },
  onError: (_e, _v, ctx) => utils.todo.list.setData(undefined, ctx?.prev),
  onSettled: () => utils.todo.list.invalidate(),
});

// Infinite query
const { data, fetchNextPage, hasNextPage } = trpc.post.infinite.useInfiniteQuery(
  { limit: 20 },
  { getNextPageParam: (last) => last.nextCursor },
);
```

## Server Component (RSC)

```typescript
const createCaller = createCallerFactory(appRouter);
export const api = cache(async () => {
  const heads = new Headers(await headers());
  return createCaller(await createTRPCContext({ headers: heads }));
});

// Usage in page.tsx
const caller = await api();
const posts = await caller.post.list({ page: 1 });
```

## Next.js Route Handler

```typescript
// app/api/trpc/[trpc]/route.ts
import { fetchRequestHandler } from '@trpc/server/adapters/fetch';
const handler = (req: Request) =>
  fetchRequestHandler({ endpoint: '/api/trpc', req, router: appRouter, createContext: () => createTRPCContext({ headers: req.headers }) });
export { handler as GET, handler as POST };
```

## Testing

```typescript
const createCaller = createCallerFactory(appRouter);
const caller = createCaller({ db, session: mockSession, headers: new Headers() });
const post = await caller.post.create({ title: 'Test', content: 'Content' });
await expect(anonCaller.post.create({ title: 'T', content: 'C' })).rejects.toThrow('UNAUTHORIZED');
```

## Error Codes

| Code | HTTP | Use |
|------|------|-----|
| `BAD_REQUEST` | 400 | Invalid input |
| `UNAUTHORIZED` | 401 | Not logged in |
| `FORBIDDEN` | 403 | Wrong role/permission |
| `NOT_FOUND` | 404 | Resource missing |
| `CONFLICT` | 409 | Duplicate |
| `TOO_MANY_REQUESTS` | 429 | Rate limited |
| `INTERNAL_SERVER_ERROR` | 500 | Unexpected |

## Common Patterns

| Pattern | Code |
|---------|------|
| Conditional query | `useQuery({ id }, { enabled: !!id })` |
| Select/transform | `useQuery({}, { select: (d) => d.items.map(i => i.name) })` |
| Invalidate cache | `utils.post.list.invalidate()` |
| Prefetch | `await utils.post.byId.prefetch({ id })` |
| Get cached data | `utils.post.list.getData({ page: 1 })` |
| Set cached data | `utils.post.list.setData(input, updater)` |
| Cancel queries | `await utils.post.list.cancel()` |
| Shared Zod schemas | Export from `src/schemas/`, import in both server + client |

## Quick Reference

```
tRPC v11 + React Query v5 + Zod v3 + superjson

Server:  initTRPC → router → procedures (query/mutation/subscription)
Client:  createTRPCReact → useQuery/useMutation/useSubscription
RSC:     createCallerFactory → direct function calls
Testing: createCallerFactory → mock context → call procedures directly
```
