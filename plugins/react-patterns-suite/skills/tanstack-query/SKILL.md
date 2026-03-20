---
name: tanstack-query
description: >
  TanStack Query (React Query) — queries, mutations, caching, infinite scroll,
  prefetching, optimistic updates, and query invalidation patterns.
  Triggers: "react query", "tanstack query", "useQuery", "useMutation",
  "data fetching", "query invalidation", "infinite scroll", "prefetch".
  NOT for: client state (use zustand-state), form state (use react-hook-form).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# TanStack Query (React Query)

## Quick Start

```bash
npm install @tanstack/react-query @tanstack/react-query-devtools
```

```tsx
// main.tsx
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ReactQueryDevtools } from '@tanstack/react-query-devtools';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000 * 60,              // 1 minute before refetch
      gcTime: 1000 * 60 * 5,             // 5 minutes in cache after inactive
      retry: 1,                            // retry failed queries once
      refetchOnWindowFocus: false,         // don't refetch on tab focus
    },
  },
});

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <Router />
      <ReactQueryDevtools initialIsOpen={false} />
    </QueryClientProvider>
  );
}
```

## Queries

### Basic Query

```tsx
import { useQuery } from '@tanstack/react-query';

function UserList() {
  const { data, isLoading, error, isError, isFetching } = useQuery({
    queryKey: ['users'],
    queryFn: async () => {
      const res = await fetch('/api/users');
      if (!res.ok) throw new Error('Failed to fetch users');
      return res.json() as Promise<User[]>;
    },
  });

  if (isLoading) return <Skeleton />;           // First load
  if (isError) return <ErrorMessage error={error} />;

  return (
    <div>
      {isFetching && <RefreshIndicator />}       {/* Background refetch */}
      {data.map((user) => <UserCard key={user.id} user={user} />)}
    </div>
  );
}
```

### Parameterized Query

```tsx
function UserProfile({ userId }: { userId: string }) {
  const { data: user } = useQuery({
    queryKey: ['users', userId],                 // Unique per userId
    queryFn: () => fetchUser(userId),
    enabled: !!userId,                           // Don't fetch if no userId
  });

  return user ? <ProfileCard user={user} /> : null;
}
```

### Dependent Queries

```tsx
function UserPosts({ userId }: { userId: string }) {
  // First: fetch user
  const { data: user } = useQuery({
    queryKey: ['users', userId],
    queryFn: () => fetchUser(userId),
  });

  // Then: fetch their posts (waits for user)
  const { data: posts } = useQuery({
    queryKey: ['posts', { authorId: user?.id }],
    queryFn: () => fetchPostsByAuthor(user!.id),
    enabled: !!user?.id,                         // Only runs after user loads
  });

  return <PostList posts={posts ?? []} />;
}
```

### Parallel Queries

```tsx
import { useQueries } from '@tanstack/react-query';

function Dashboard() {
  const results = useQueries({
    queries: [
      { queryKey: ['stats'], queryFn: fetchStats },
      { queryKey: ['notifications'], queryFn: fetchNotifications },
      { queryKey: ['recent-activity'], queryFn: fetchRecentActivity },
    ],
  });

  const [stats, notifications, activity] = results;
  const isLoading = results.some((r) => r.isLoading);

  if (isLoading) return <DashboardSkeleton />;

  return (
    <>
      <StatsGrid data={stats.data} />
      <NotificationList items={notifications.data} />
      <ActivityFeed items={activity.data} />
    </>
  );
}
```

## Mutations

### Basic Mutation

```tsx
import { useMutation, useQueryClient } from '@tanstack/react-query';

function CreateUserForm() {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (newUser: CreateUser) => {
      return fetch('/api/users', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(newUser),
      }).then((res) => {
        if (!res.ok) throw new Error('Failed to create user');
        return res.json() as Promise<User>;
      });
    },
    onSuccess: () => {
      // Invalidate and refetch user list
      queryClient.invalidateQueries({ queryKey: ['users'] });
      toast.success('User created!');
    },
    onError: (error) => {
      toast.error(error.message);
    },
  });

  return (
    <form onSubmit={handleSubmit((data) => mutation.mutate(data))}>
      {/* fields */}
      <button disabled={mutation.isPending}>
        {mutation.isPending ? 'Creating...' : 'Create User'}
      </button>
    </form>
  );
}
```

### Optimistic Updates

```tsx
const toggleTodo = useMutation({
  mutationFn: ({ id, done }: { id: string; done: boolean }) =>
    api.todos.update(id, { done }),

  // Optimistic update: change UI immediately
  onMutate: async ({ id, done }) => {
    // Cancel outgoing refetches
    await queryClient.cancelQueries({ queryKey: ['todos'] });

    // Snapshot previous value
    const previousTodos = queryClient.getQueryData<Todo[]>(['todos']);

    // Optimistically update
    queryClient.setQueryData<Todo[]>(['todos'], (old) =>
      old?.map((t) => (t.id === id ? { ...t, done } : t))
    );

    // Return snapshot for rollback
    return { previousTodos };
  },

  // Rollback on error
  onError: (_error, _variables, context) => {
    queryClient.setQueryData(['todos'], context?.previousTodos);
  },

  // Refetch after settle (success or error)
  onSettled: () => {
    queryClient.invalidateQueries({ queryKey: ['todos'] });
  },
});
```

### Delete with Optimistic Update

```tsx
const deleteMutation = useMutation({
  mutationFn: (id: string) => api.users.delete(id),
  onMutate: async (id) => {
    await queryClient.cancelQueries({ queryKey: ['users'] });
    const previous = queryClient.getQueryData<User[]>(['users']);
    queryClient.setQueryData<User[]>(['users'], (old) =>
      old?.filter((u) => u.id !== id)
    );
    return { previous };
  },
  onError: (_err, _id, context) => {
    queryClient.setQueryData(['users'], context?.previous);
  },
  onSettled: () => {
    queryClient.invalidateQueries({ queryKey: ['users'] });
  },
});
```

## Query Invalidation

```tsx
const queryClient = useQueryClient();

// Invalidate a specific query
queryClient.invalidateQueries({ queryKey: ['users'] });

// Invalidate all queries starting with 'users'
queryClient.invalidateQueries({ queryKey: ['users'], exact: false });

// Invalidate everything
queryClient.invalidateQueries();

// Remove from cache entirely
queryClient.removeQueries({ queryKey: ['users', userId] });

// Manually update cache (without refetch)
queryClient.setQueryData(['users', userId], updatedUser);

// Invalidation patterns after mutations:
// Created a user   → invalidate ['users'] (list)
// Updated a user   → invalidate ['users'] + ['users', userId]
// Deleted a user   → invalidate ['users'] (list)
// Changed settings → invalidate ['settings']
```

## Infinite Scroll / Pagination

```tsx
import { useInfiniteQuery } from '@tanstack/react-query';

function InfinitePostList() {
  const {
    data,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
    isLoading,
  } = useInfiniteQuery({
    queryKey: ['posts'],
    queryFn: ({ pageParam }) =>
      fetch(`/api/posts?cursor=${pageParam}&limit=20`).then((r) => r.json()),
    initialPageParam: '',
    getNextPageParam: (lastPage) => lastPage.nextCursor ?? undefined,
  });

  const posts = data?.pages.flatMap((page) => page.items) ?? [];

  return (
    <>
      {isLoading && <PostListSkeleton />}
      {posts.map((post) => <PostCard key={post.id} post={post} />)}
      {hasNextPage && (
        <button
          onClick={() => fetchNextPage()}
          disabled={isFetchingNextPage}
        >
          {isFetchingNextPage ? 'Loading...' : 'Load More'}
        </button>
      )}
    </>
  );
}

// With intersection observer (auto-load on scroll)
function InfiniteScrollTrigger({ onIntersect }: { onIntersect: () => void }) {
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const observer = new IntersectionObserver(
      ([entry]) => { if (entry.isIntersecting) onIntersect(); },
      { threshold: 0.1 }
    );
    if (ref.current) observer.observe(ref.current);
    return () => observer.disconnect();
  }, [onIntersect]);

  return <div ref={ref} className="h-4" />;
}
```

## Prefetching

```tsx
// Prefetch on hover (instant navigation)
function UserLink({ userId, children }: { userId: string; children: ReactNode }) {
  const queryClient = useQueryClient();

  const prefetchUser = () => {
    queryClient.prefetchQuery({
      queryKey: ['users', userId],
      queryFn: () => fetchUser(userId),
      staleTime: 1000 * 60,  // don't refetch if less than 1 min old
    });
  };

  return (
    <Link
      to={`/users/${userId}`}
      onMouseEnter={prefetchUser}
      onFocus={prefetchUser}
    >
      {children}
    </Link>
  );
}

// Prefetch on route load (loader pattern)
// In React Router:
const userRoute = {
  path: '/users/:id',
  loader: ({ params }) => {
    queryClient.prefetchQuery({
      queryKey: ['users', params.id],
      queryFn: () => fetchUser(params.id!),
    });
    return null;
  },
  element: <UserPage />,
};
```

## API Client Pattern

```typescript
// api/client.ts
const API_BASE = '/api';

class ApiError extends Error {
  constructor(public status: number, message: string, public details?: any) {
    super(message);
  }
}

async function apiFetch<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
  });

  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new ApiError(res.status, body.message ?? 'Request failed', body.details);
  }

  return res.json();
}

// api/users.ts
export const usersApi = {
  list: (params?: SearchParams) =>
    apiFetch<{ items: User[]; total: number }>(`/users?${new URLSearchParams(params)}`),
  get: (id: string) => apiFetch<User>(`/users/${id}`),
  create: (data: CreateUser) => apiFetch<User>('/users', { method: 'POST', body: JSON.stringify(data) }),
  update: (id: string, data: UpdateUser) => apiFetch<User>(`/users/${id}`, { method: 'PATCH', body: JSON.stringify(data) }),
  delete: (id: string) => apiFetch<void>(`/users/${id}`, { method: 'DELETE' }),
};

// hooks/useUsers.ts
export function useUsers(params?: SearchParams) {
  return useQuery({
    queryKey: ['users', params],
    queryFn: () => usersApi.list(params),
  });
}

export function useUser(id: string) {
  return useQuery({
    queryKey: ['users', id],
    queryFn: () => usersApi.get(id),
    enabled: !!id,
  });
}

export function useCreateUser() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: usersApi.create,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['users'] }),
  });
}
```

## Gotchas

1. **`queryKey` must include all variables.** If your query depends on `userId`, include it: `['users', userId]`. Otherwise the cache won't update when the variable changes.

2. **`staleTime` vs `gcTime`.** `staleTime` = how long data is "fresh" (no refetch). `gcTime` = how long inactive data stays in cache. Set `staleTime` for data that doesn't change often.

3. **Don't put mutations in query keys.** Query keys are for reading. Mutations use `mutationFn` directly.

4. **`enabled: false` doesn't mean "never".** It means "don't auto-fetch." You can still trigger with `refetch()`. Use it for dependent queries.

5. **Invalidation is async.** `invalidateQueries` triggers a background refetch. If you need the fresh data immediately, use `await queryClient.fetchQuery(...)` instead.

6. **Error boundaries catch query errors.** If `throwOnError: true` (or the default `suspense: true`), query errors propagate to error boundaries. Handle errors either in the component or at the boundary, not both.
