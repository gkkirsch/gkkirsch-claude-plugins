---
name: trpc-client
description: >
  tRPC client integration — React Query hooks, Next.js App Router RSC, provider setup,
  optimistic updates, prefetching, error handling, and infinite queries.
  Triggers: "trpc client", "trpc react", "trpc hook", "trpc query", "trpc mutation",
  "trpc provider", "trpc next.js client", "trpc react query".
  NOT for: server-side setup (use trpc-server), advanced patterns (use trpc-patterns).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# tRPC Client

## Provider Setup (Next.js App Router)

```typescript
// src/lib/trpc.ts — client-side tRPC setup
'use client';

import { createTRPCReact } from '@trpc/react-query';
import type { AppRouter } from '@/server/routers/_app';

export const trpc = createTRPCReact<AppRouter>();
```

```typescript
// src/lib/trpc-provider.tsx
'use client';

import { useState } from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { httpBatchLink } from '@trpc/client';
import { trpc } from './trpc';
import superjson from 'superjson';

export function TRPCProvider({ children }: { children: React.ReactNode }) {
  const [queryClient] = useState(
    () =>
      new QueryClient({
        defaultOptions: {
          queries: {
            staleTime: 5 * 60 * 1000, // 5 minutes
            gcTime: 10 * 60 * 1000,   // 10 minutes
            refetchOnWindowFocus: false,
          },
        },
      }),
  );

  const [trpcClient] = useState(() =>
    trpc.createClient({
      links: [
        httpBatchLink({
          url: '/api/trpc',
          transformer: superjson,
          headers() {
            return {
              // Add any headers (auth tokens, etc.)
            };
          },
        }),
      ],
    }),
  );

  return (
    <trpc.Provider client={trpcClient} queryClient={queryClient}>
      <QueryClientProvider client={queryClient}>
        {children}
      </QueryClientProvider>
    </trpc.Provider>
  );
}
```

```tsx
// src/app/layout.tsx
import { TRPCProvider } from '@/lib/trpc-provider';

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html>
      <body>
        <TRPCProvider>{children}</TRPCProvider>
      </body>
    </html>
  );
}
```

## Queries

```tsx
'use client';

import { trpc } from '@/lib/trpc';

// Basic query
function PostList() {
  const { data, isLoading, error } = trpc.post.list.useQuery({
    page: 1,
    limit: 20,
  });

  if (isLoading) return <div>Loading...</div>;
  if (error) return <div>Error: {error.message}</div>;

  return (
    <ul>
      {data.items.map((post) => (
        <li key={post.id}>{post.title}</li>
      ))}
    </ul>
  );
}

// Query with parameters
function PostDetail({ id }: { id: string }) {
  const { data: post, isLoading } = trpc.post.byId.useQuery({ id });

  if (isLoading) return <div>Loading...</div>;
  if (!post) return <div>Not found</div>;

  return <article>{post.title}</article>;
}

// Conditional query (don't run until condition is met)
function UserProfile({ userId }: { userId?: string }) {
  const { data } = trpc.user.byId.useQuery(
    { id: userId! },
    { enabled: !!userId },
  );

  return data ? <div>{data.name}</div> : null;
}

// Query with select (transform data client-side)
function PostTitles() {
  const { data: titles } = trpc.post.list.useQuery(
    { page: 1, limit: 100 },
    { select: (data) => data.items.map((p) => p.title) },
  );

  return <ul>{titles?.map((t) => <li key={t}>{t}</li>)}</ul>;
}
```

## Mutations

```tsx
'use client';

import { trpc } from '@/lib/trpc';

function CreatePostForm() {
  const utils = trpc.useUtils();

  const createPost = trpc.post.create.useMutation({
    onSuccess: () => {
      // Invalidate the posts list so it refetches
      utils.post.list.invalidate();
    },
    onError: (error) => {
      if (error.data?.zodError) {
        // Handle validation errors
        const fieldErrors = error.data.zodError.fieldErrors;
        console.log('Validation errors:', fieldErrors);
      } else {
        console.error('Error:', error.message);
      }
    },
  });

  const handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const formData = new FormData(e.currentTarget);

    createPost.mutate({
      title: formData.get('title') as string,
      content: formData.get('content') as string,
    });
  };

  return (
    <form onSubmit={handleSubmit}>
      <input name="title" placeholder="Title" required />
      <textarea name="content" placeholder="Content" required />
      <button type="submit" disabled={createPost.isPending}>
        {createPost.isPending ? 'Creating...' : 'Create Post'}
      </button>
      {createPost.error && (
        <p className="text-red-500">{createPost.error.message}</p>
      )}
    </form>
  );
}
```

## Optimistic Updates

```tsx
function TodoList() {
  const utils = trpc.useUtils();

  const toggleTodo = trpc.todo.toggle.useMutation({
    // Optimistic update: update UI immediately
    onMutate: async ({ id }) => {
      // Cancel outgoing queries (prevent overwrite)
      await utils.todo.list.cancel();

      // Get current data
      const previousTodos = utils.todo.list.getData();

      // Optimistically update
      utils.todo.list.setData(undefined, (old) =>
        old?.map((todo) =>
          todo.id === id ? { ...todo, completed: !todo.completed } : todo,
        ),
      );

      // Return previous data for rollback
      return { previousTodos };
    },

    // Rollback on error
    onError: (_err, _vars, context) => {
      if (context?.previousTodos) {
        utils.todo.list.setData(undefined, context.previousTodos);
      }
    },

    // Refetch after mutation settles (success or error)
    onSettled: () => {
      utils.todo.list.invalidate();
    },
  });

  // ...
}
```

## Infinite Queries (Cursor Pagination)

```tsx
function InfinitePostList() {
  const {
    data,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
    isLoading,
  } = trpc.post.infiniteList.useInfiniteQuery(
    { limit: 20 },
    {
      getNextPageParam: (lastPage) => lastPage.nextCursor,
    },
  );

  if (isLoading) return <div>Loading...</div>;

  return (
    <>
      {data?.pages.map((page) =>
        page.items.map((post) => (
          <PostCard key={post.id} post={post} />
        )),
      )}

      {hasNextPage && (
        <button
          onClick={() => fetchNextPage()}
          disabled={isFetchingNextPage}
        >
          {isFetchingNextPage ? 'Loading more...' : 'Load More'}
        </button>
      )}
    </>
  );
}
```

Server-side for infinite queries:

```typescript
// Server router
infiniteList: publicProcedure
  .input(z.object({
    limit: z.number().min(1).max(100).default(20),
    cursor: z.string().optional(), // Last item ID
  }))
  .query(async ({ ctx, input }) => {
    const { limit, cursor } = input;

    const items = await ctx.db.post.findMany({
      take: limit + 1, // Fetch one extra to check for next page
      where: { published: true },
      cursor: cursor ? { id: cursor } : undefined,
      orderBy: { createdAt: 'desc' },
    });

    let nextCursor: string | undefined;
    if (items.length > limit) {
      const nextItem = items.pop(); // Remove extra item
      nextCursor = nextItem!.id;
    }

    return { items, nextCursor };
  }),
```

## Server-Side Calls (RSC / Server Components)

```typescript
// src/lib/trpc-server.ts
import 'server-only';
import { createCallerFactory } from '@/server/trpc';
import { appRouter } from '@/server/routers/_app';
import { createTRPCContext } from '@/server/trpc';
import { headers } from 'next/headers';
import { cache } from 'react';

const createCaller = createCallerFactory(appRouter);

export const api = cache(async () => {
  const heads = new Headers(await headers());
  return createCaller(await createTRPCContext({ headers: heads }));
});
```

```tsx
// src/app/posts/page.tsx — Server Component
import { api } from '@/lib/trpc-server';

export default async function PostsPage() {
  const caller = await api();
  const { items: posts } = await caller.post.list({ page: 1, limit: 20 });

  return (
    <ul>
      {posts.map((post) => (
        <li key={post.id}>{post.title}</li>
      ))}
    </ul>
  );
}
```

## Error Handling on Client

```tsx
function PostDetail({ id }: { id: string }) {
  const { data, error, isLoading } = trpc.post.byId.useQuery({ id });

  if (isLoading) return <Skeleton />;

  if (error) {
    // tRPC errors have structured data
    switch (error.data?.code) {
      case 'NOT_FOUND':
        return <NotFound message="Post not found" />;
      case 'UNAUTHORIZED':
        return <LoginPrompt />;
      case 'FORBIDDEN':
        return <AccessDenied />;
      default:
        return <ErrorMessage message={error.message} />;
    }
  }

  return <Post data={data} />;
}

// Handle mutation validation errors
function CreateForm() {
  const create = trpc.post.create.useMutation();

  // Access Zod field errors
  const fieldErrors = create.error?.data?.zodError?.fieldErrors;

  return (
    <form>
      <input name="title" />
      {fieldErrors?.title && (
        <span className="text-red-500">{fieldErrors.title[0]}</span>
      )}
    </form>
  );
}
```

## useUtils — Cache Management

```tsx
const utils = trpc.useUtils();

// Invalidate (refetch) queries
utils.post.list.invalidate();              // Invalidate all post.list queries
utils.post.byId.invalidate({ id: '1' });  // Invalidate specific query

// Prefetch data
await utils.post.byId.prefetch({ id: '1' });

// Get cached data
const cachedPosts = utils.post.list.getData({ page: 1, limit: 20 });

// Set cached data manually
utils.post.list.setData({ page: 1, limit: 20 }, (old) => ({
  ...old!,
  items: [...old!.items, newPost],
}));

// Cancel in-flight queries
await utils.post.list.cancel();

// Reset query to initial state
utils.post.list.reset();
```

## Gotchas

1. **Provider must wrap the app.** The `TRPCProvider` must be a client component and wrap everything that uses tRPC hooks. In Next.js App Router, put it in `layout.tsx`.

2. **`trpc.createClient` and `QueryClient` must use `useState`.** Creating them outside `useState` causes hydration mismatches in Next.js. Always `const [client] = useState(() => ...)`.

3. **Don't import server types at runtime.** `import type { AppRouter }` is safe (compile-time only). `import { appRouter }` pulls server code into the client bundle.

4. **`superjson` must match server and client.** If the server uses `superjson` transformer, the client must too. Mismatched transformers cause silent data corruption.

5. **`invalidate()` triggers a refetch, not an immediate update.** If you need instant UI feedback, use optimistic updates via `setData` in `onMutate`.

6. **Infinite query cursor must match the server.** `getNextPageParam` must return the exact cursor type the server expects. `undefined` means no more pages.
