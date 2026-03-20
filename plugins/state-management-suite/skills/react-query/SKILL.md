---
name: react-query
description: >
  TanStack React Query — queries, mutations, infinite queries, optimistic updates,
  prefetching, query invalidation, and suspense integration.
  Triggers: "react query", "tanstack query", "useQuery", "useMutation",
  "server state", "data fetching", "query invalidation".
  NOT for: client-only state (use zustand or jotai).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# TanStack React Query

## Quick Start

```bash
npm install @tanstack/react-query @tanstack/react-query-devtools
```

## Provider Setup

```tsx
// app/providers.tsx
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ReactQueryDevtools } from '@tanstack/react-query-devtools';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 60 * 1000,         // 1 minute before data is "stale"
      gcTime: 5 * 60 * 1000,        // 5 minutes before inactive cache is garbage collected
      retry: 3,                      // Retry failed requests 3 times
      refetchOnWindowFocus: true,    // Refetch when tab regains focus
    },
  },
});

export function Providers({ children }: { children: React.ReactNode }) {
  return (
    <QueryClientProvider client={queryClient}>
      {children}
      <ReactQueryDevtools initialIsOpen={false} />
    </QueryClientProvider>
  );
}
```

## Basic Query

```typescript
// hooks/useUsers.ts
import { useQuery } from '@tanstack/react-query';

interface User {
  id: string;
  name: string;
  email: string;
}

async function fetchUsers(): Promise<User[]> {
  const response = await fetch('/api/users');
  if (!response.ok) throw new Error('Failed to fetch users');
  return response.json();
}

export function useUsers() {
  return useQuery({
    queryKey: ['users'],
    queryFn: fetchUsers,
  });
}

// Usage
function UsersList() {
  const { data: users, isLoading, error, isRefetching } = useUsers();

  if (isLoading) return <Skeleton count={5} />;
  if (error) return <ErrorMessage error={error} />;

  return (
    <div>
      {isRefetching && <RefreshIndicator />}
      {users?.map((user) => <UserCard key={user.id} user={user} />)}
    </div>
  );
}
```

## Query with Parameters

```typescript
// hooks/useUser.ts
export function useUser(userId: string) {
  return useQuery({
    queryKey: ['users', userId],
    queryFn: async () => {
      const response = await fetch(`/api/users/${userId}`);
      if (!response.ok) throw new Error('User not found');
      return response.json() as Promise<User>;
    },
    enabled: !!userId, // Don't fetch if no userId
  });
}

// hooks/useFilteredPosts.ts
interface PostFilters {
  status?: string;
  authorId?: string;
  page?: number;
  limit?: number;
}

export function usePosts(filters: PostFilters) {
  return useQuery({
    queryKey: ['posts', filters], // Re-fetches when filters change
    queryFn: async () => {
      const params = new URLSearchParams();
      if (filters.status) params.set('status', filters.status);
      if (filters.authorId) params.set('authorId', filters.authorId);
      params.set('page', String(filters.page ?? 1));
      params.set('limit', String(filters.limit ?? 20));

      const response = await fetch(`/api/posts?${params}`);
      if (!response.ok) throw new Error('Failed to fetch posts');
      return response.json();
    },
    placeholderData: (previousData) => previousData, // Keep previous page while loading next
  });
}
```

## Mutations

```typescript
// hooks/useCreatePost.ts
import { useMutation, useQueryClient } from '@tanstack/react-query';

interface CreatePostInput {
  title: string;
  body: string;
}

export function useCreatePost() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: CreatePostInput) => {
      const response = await fetch('/api/posts', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data),
      });
      if (!response.ok) throw new Error('Failed to create post');
      return response.json() as Promise<Post>;
    },
    onSuccess: (newPost) => {
      // Invalidate and refetch the posts list
      queryClient.invalidateQueries({ queryKey: ['posts'] });

      // Or manually add to cache (avoids refetch)
      queryClient.setQueryData<Post[]>(['posts'], (old) =>
        old ? [newPost, ...old] : [newPost]
      );
    },
    onError: (error) => {
      toast.error(`Failed: ${error.message}`);
    },
  });
}

// Usage
function CreatePostForm() {
  const createPost = useCreatePost();

  const handleSubmit = (data: CreatePostInput) => {
    createPost.mutate(data, {
      onSuccess: () => {
        toast.success('Post created!');
        router.push('/posts');
      },
    });
  };

  return (
    <form onSubmit={handleSubmit}>
      {/* form fields */}
      <button type="submit" disabled={createPost.isPending}>
        {createPost.isPending ? 'Creating...' : 'Create Post'}
      </button>
    </form>
  );
}
```

## Optimistic Updates

```typescript
export function useUpdatePost() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ id, ...data }: { id: string; title: string }) => {
      const response = await fetch(`/api/posts/${id}`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data),
      });
      if (!response.ok) throw new Error('Failed to update');
      return response.json() as Promise<Post>;
    },

    onMutate: async ({ id, ...data }) => {
      // Cancel outgoing refetches
      await queryClient.cancelQueries({ queryKey: ['posts', id] });

      // Snapshot previous value
      const previousPost = queryClient.getQueryData<Post>(['posts', id]);

      // Optimistically update
      queryClient.setQueryData<Post>(['posts', id], (old) =>
        old ? { ...old, ...data } : old
      );

      return { previousPost };
    },

    onError: (error, variables, context) => {
      // Rollback on error
      if (context?.previousPost) {
        queryClient.setQueryData(['posts', variables.id], context.previousPost);
      }
      toast.error('Update failed. Changes reverted.');
    },

    onSettled: (data, error, { id }) => {
      // Always refetch after error or success to ensure consistency
      queryClient.invalidateQueries({ queryKey: ['posts', id] });
    },
  });
}
```

## Infinite Queries (Infinite Scroll)

```typescript
import { useInfiniteQuery } from '@tanstack/react-query';

interface PostsPage {
  data: Post[];
  meta: {
    nextCursor: string | null;
    hasMore: boolean;
  };
}

export function useInfinitePosts() {
  return useInfiniteQuery({
    queryKey: ['posts', 'infinite'],
    queryFn: async ({ pageParam }): Promise<PostsPage> => {
      const params = new URLSearchParams({ limit: '20' });
      if (pageParam) params.set('cursor', pageParam);

      const response = await fetch(`/api/posts?${params}`);
      if (!response.ok) throw new Error('Failed to fetch');
      return response.json();
    },
    initialPageParam: null as string | null,
    getNextPageParam: (lastPage) =>
      lastPage.meta.hasMore ? lastPage.meta.nextCursor : undefined,
  });
}

// Usage with Intersection Observer
function InfinitePostsList() {
  const {
    data,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
  } = useInfinitePosts();

  const loadMoreRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting && hasNextPage && !isFetchingNextPage) {
          fetchNextPage();
        }
      },
      { threshold: 0.5 }
    );

    if (loadMoreRef.current) observer.observe(loadMoreRef.current);
    return () => observer.disconnect();
  }, [hasNextPage, isFetchingNextPage, fetchNextPage]);

  const allPosts = data?.pages.flatMap((page) => page.data) ?? [];

  return (
    <div>
      {allPosts.map((post) => (
        <PostCard key={post.id} post={post} />
      ))}
      <div ref={loadMoreRef}>
        {isFetchingNextPage && <Spinner />}
      </div>
    </div>
  );
}
```

## Prefetching

```typescript
// Prefetch on hover
function PostLink({ postId }: { postId: string }) {
  const queryClient = useQueryClient();

  const prefetch = () => {
    queryClient.prefetchQuery({
      queryKey: ['posts', postId],
      queryFn: () => fetchPost(postId),
      staleTime: 60_000, // Don't refetch if less than 1 min old
    });
  };

  return (
    <Link href={`/posts/${postId}`} onMouseEnter={prefetch}>
      View Post
    </Link>
  );
}

// Prefetch in route loader (Next.js)
// app/posts/[id]/page.tsx
export default async function PostPage({ params }: { params: { id: string } }) {
  const queryClient = new QueryClient();

  await queryClient.prefetchQuery({
    queryKey: ['posts', params.id],
    queryFn: () => fetchPost(params.id),
  });

  return (
    <HydrationBoundary state={dehydrate(queryClient)}>
      <PostDetail id={params.id} />
    </HydrationBoundary>
  );
}
```

## Parallel Queries

```typescript
import { useQueries } from '@tanstack/react-query';

function Dashboard() {
  const results = useQueries({
    queries: [
      { queryKey: ['users'], queryFn: fetchUsers },
      { queryKey: ['posts'], queryFn: fetchPosts },
      { queryKey: ['analytics'], queryFn: fetchAnalytics },
    ],
  });

  const [users, posts, analytics] = results;
  const isLoading = results.some((r) => r.isLoading);

  if (isLoading) return <DashboardSkeleton />;

  return (
    <div>
      <UsersWidget data={users.data} />
      <PostsWidget data={posts.data} />
      <AnalyticsWidget data={analytics.data} />
    </div>
  );
}
```

## Query Invalidation Patterns

```typescript
const queryClient = useQueryClient();

// Invalidate everything
queryClient.invalidateQueries();

// Invalidate all queries starting with 'posts'
queryClient.invalidateQueries({ queryKey: ['posts'] });

// Invalidate exact query
queryClient.invalidateQueries({ queryKey: ['posts', '123'], exact: true });

// Invalidate by predicate
queryClient.invalidateQueries({
  predicate: (query) =>
    query.queryKey[0] === 'posts' &&
    (query.queryKey[1] as any)?.status === 'draft',
});

// Remove from cache entirely
queryClient.removeQueries({ queryKey: ['posts', '123'] });

// Reset to initial state
queryClient.resetQueries({ queryKey: ['posts'] });
```

## API Client Pattern

```typescript
// lib/api.ts — centralized API client
class ApiClient {
  private baseUrl: string;
  private getToken: () => string | null;

  constructor(baseUrl: string, getToken: () => string | null) {
    this.baseUrl = baseUrl;
    this.getToken = getToken;
  }

  private async request<T>(path: string, options?: RequestInit): Promise<T> {
    const token = this.getToken();
    const response = await fetch(`${this.baseUrl}${path}`, {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
        ...options?.headers,
      },
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ message: 'Request failed' }));
      throw new ApiError(response.status, error.message ?? 'Request failed');
    }

    return response.json();
  }

  // Type-safe API methods
  users = {
    list: () => this.request<User[]>('/users'),
    get: (id: string) => this.request<User>(`/users/${id}`),
    create: (data: CreateUserInput) =>
      this.request<User>('/users', { method: 'POST', body: JSON.stringify(data) }),
  };

  posts = {
    list: (filters?: PostFilters) => {
      const params = new URLSearchParams(filters as any);
      return this.request<PaginatedResponse<Post>>(`/posts?${params}`);
    },
    get: (id: string) => this.request<Post>(`/posts/${id}`),
    create: (data: CreatePostInput) =>
      this.request<Post>('/posts', { method: 'POST', body: JSON.stringify(data) }),
    update: (id: string, data: Partial<Post>) =>
      this.request<Post>(`/posts/${id}`, { method: 'PATCH', body: JSON.stringify(data) }),
    delete: (id: string) =>
      this.request<void>(`/posts/${id}`, { method: 'DELETE' }),
  };
}

export const api = new ApiClient('/api', () => useAuthStore.getState().token);

// Then in hooks:
export function usePosts(filters?: PostFilters) {
  return useQuery({
    queryKey: ['posts', filters],
    queryFn: () => api.posts.list(filters),
  });
}
```

## Gotchas

1. **`staleTime` vs `gcTime`.** `staleTime` = how long data is "fresh" (won't refetch). `gcTime` = how long inactive data stays in cache. Set `staleTime` for your use case. Default `staleTime` is 0 (always stale).

2. **Query keys must be serializable.** Objects in keys are compared by value, not reference. `['posts', { page: 1 }]` and `['posts', { page: 1 }]` are the same key.

3. **`enabled: false` skips the query entirely.** Use for conditional fetching: `enabled: !!userId`. The query won't run until enabled becomes true.

4. **Mutations don't have keys.** They don't cache. If you need to track a mutation's status globally, use `useMutationState`.

5. **`placeholderData` vs `initialData`.** `placeholderData` shows temporary data while loading (not put in cache). `initialData` is treated as real cached data.

6. **Don't await `invalidateQueries`.** It returns a promise but you usually don't need to wait. The refetch happens in the background.

7. **Background refetch shows stale data + loading.** Use `isRefetching` (not `isLoading`) to show a subtle refresh indicator while keeping stale data visible.
