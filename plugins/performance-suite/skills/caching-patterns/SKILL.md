---
name: caching-patterns
description: >
  Caching strategies — HTTP caching, CDN configuration, service workers,
  React Query/SWR data caching, and browser storage patterns.
  Triggers: "caching", "cache strategy", "HTTP cache", "CDN", "service worker",
  "react query cache", "SWR", "stale while revalidate", "cache control".
  NOT for: React rendering (use react-performance), Core Web Vitals (use web-vitals).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Caching Patterns

## HTTP Caching Headers

```typescript
// Express.js cache headers
app.use('/assets', express.static('dist/assets', {
  maxAge: '1y',             // Cache for 1 year
  immutable: true,          // Never revalidate (hashed filenames)
}));

// API responses
app.get('/api/posts', (req, res) => {
  res.set({
    'Cache-Control': 'public, max-age=60, stale-while-revalidate=300',
    'ETag': generateETag(data),
  });
  res.json(data);
});

// No-cache (always revalidate)
app.get('/api/user/me', (req, res) => {
  res.set('Cache-Control', 'private, no-cache');
  res.json(user);
});
```

### Cache-Control Cheat Sheet

| Directive | Meaning | Use Case |
|-----------|---------|----------|
| `public` | Any cache can store | CDN-cacheable content |
| `private` | Only browser can store | User-specific data |
| `max-age=N` | Fresh for N seconds | Static resources, API data |
| `s-maxage=N` | CDN-specific max-age | CDN has different TTL than browser |
| `no-cache` | Must revalidate every time | Dynamic content, auth-dependent |
| `no-store` | Never cache at all | Sensitive data (tokens, passwords) |
| `immutable` | Never changes (skip revalidation) | Hashed filenames (app.abc123.js) |
| `stale-while-revalidate=N` | Serve stale while fetching fresh | API data, background updates |
| `must-revalidate` | Don't serve stale even if offline | Critical data integrity |

### Caching Strategy by Resource Type

```
Static assets (JS, CSS, images with hash):
  Cache-Control: public, max-age=31536000, immutable

HTML pages:
  Cache-Control: public, max-age=0, must-revalidate
  ETag: "abc123"

API data (public):
  Cache-Control: public, max-age=60, stale-while-revalidate=300

API data (private/auth):
  Cache-Control: private, no-cache

User session data:
  Cache-Control: private, no-store
```

## React Query (TanStack Query) Caching

```typescript
import { QueryClient, QueryClientProvider, useQuery, useMutation } from '@tanstack/react-query';

// Configure global defaults
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 5 * 60 * 1000,     // Data fresh for 5 minutes
      gcTime: 30 * 60 * 1000,       // Keep in cache for 30 minutes (was cacheTime)
      retry: 3,                       // Retry failed requests 3 times
      refetchOnWindowFocus: true,     // Refetch when tab gets focus
      refetchOnReconnect: true,       // Refetch on network reconnect
    },
  },
});

// Basic query with caching
function useUsers() {
  return useQuery({
    queryKey: ['users'],
    queryFn: () => fetch('/api/users').then(r => r.json()),
    staleTime: 5 * 60 * 1000,
  });
}

// Query with parameters
function useUser(id: string) {
  return useQuery({
    queryKey: ['users', id],
    queryFn: () => fetch(`/api/users/${id}`).then(r => r.json()),
    enabled: !!id,  // Don't run until id exists
  });
}

// Prefetch (e.g., on hover)
function UserLink({ id }: { id: string }) {
  const queryClient = useQueryClient();

  const prefetch = () => {
    queryClient.prefetchQuery({
      queryKey: ['users', id],
      queryFn: () => fetch(`/api/users/${id}`).then(r => r.json()),
    });
  };

  return <a href={`/users/${id}`} onMouseEnter={prefetch}>View User</a>;
}

// Optimistic updates with mutation
function useUpdateUser() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: UpdateUser) =>
      fetch(`/api/users/${data.id}`, {
        method: 'PUT',
        body: JSON.stringify(data),
        headers: { 'Content-Type': 'application/json' },
      }).then(r => r.json()),

    onMutate: async (newData) => {
      await queryClient.cancelQueries({ queryKey: ['users', newData.id] });
      const previous = queryClient.getQueryData(['users', newData.id]);
      queryClient.setQueryData(['users', newData.id], newData);
      return { previous };
    },

    onError: (err, newData, context) => {
      queryClient.setQueryData(['users', newData.id], context?.previous);
    },

    onSettled: (data, error, variables) => {
      queryClient.invalidateQueries({ queryKey: ['users', variables.id] });
    },
  });
}

// Infinite scroll with caching
function useInfinitePosts() {
  return useInfiniteQuery({
    queryKey: ['posts'],
    queryFn: ({ pageParam }) =>
      fetch(`/api/posts?cursor=${pageParam}`).then(r => r.json()),
    initialPageParam: '',
    getNextPageParam: (lastPage) => lastPage.nextCursor ?? undefined,
  });
}
```

## SWR Caching

```typescript
import useSWR, { SWRConfig } from 'swr';

const fetcher = (url: string) => fetch(url).then(r => r.json());

// Global config
function App() {
  return (
    <SWRConfig value={{
      fetcher,
      revalidateOnFocus: true,
      dedupingInterval: 2000,
      errorRetryCount: 3,
    }}>
      <Dashboard />
    </SWRConfig>
  );
}

// Basic usage (automatic caching + revalidation)
function UserProfile({ id }: { id: string }) {
  const { data, error, isLoading, mutate } = useSWR(`/api/users/${id}`);

  if (isLoading) return <Skeleton />;
  if (error) return <Error />;
  return <div>{data.name}</div>;
}

// Optimistic update
async function updateUser(id: string, name: string) {
  await mutate(
    `/api/users/${id}`,
    async (current: User) => {
      await fetch(`/api/users/${id}`, { method: 'PUT', body: JSON.stringify({ name }) });
      return { ...current, name };
    },
    { optimisticData: (current: User) => ({ ...current, name }), rollbackOnError: true },
  );
}

// Conditional fetching
const { data } = useSWR(userId ? `/api/users/${userId}` : null);

// Dependent queries
const { data: user } = useSWR('/api/user');
const { data: projects } = useSWR(() => `/api/users/${user.id}/projects`);
```

## Service Worker Caching

```typescript
// service-worker.ts (using Workbox)
import { precacheAndRoute } from 'workbox-precaching';
import { registerRoute } from 'workbox-routing';
import { CacheFirst, StaleWhileRevalidate, NetworkFirst } from 'workbox-strategies';
import { ExpirationPlugin } from 'workbox-expiration';

// Precache build assets
precacheAndRoute(self.__WB_MANIFEST);

// Cache-first for static assets (images, fonts)
registerRoute(
  ({ request }) => request.destination === 'image' || request.destination === 'font',
  new CacheFirst({
    cacheName: 'static-assets',
    plugins: [new ExpirationPlugin({ maxEntries: 100, maxAgeSeconds: 30 * 24 * 60 * 60 })],
  }),
);

// Stale-while-revalidate for API data
registerRoute(
  ({ url }) => url.pathname.startsWith('/api/'),
  new StaleWhileRevalidate({
    cacheName: 'api-cache',
    plugins: [new ExpirationPlugin({ maxEntries: 50, maxAgeSeconds: 5 * 60 })],
  }),
);

// Network-first for HTML pages
registerRoute(
  ({ request }) => request.mode === 'navigate',
  new NetworkFirst({
    cacheName: 'pages',
    plugins: [new ExpirationPlugin({ maxEntries: 25 })],
  }),
);
```

## CDN Configuration

```nginx
# Nginx CDN/reverse proxy caching
location /assets/ {
    expires 1y;
    add_header Cache-Control "public, immutable";
    add_header X-Cache-Status $upstream_cache_status;
}

location /api/ {
    proxy_cache api_cache;
    proxy_cache_valid 200 60s;
    proxy_cache_use_stale error timeout updating;
    add_header X-Cache-Status $upstream_cache_status;

    # Vary by auth header (private caching per user)
    proxy_cache_key "$request_uri|$http_authorization";
}
```

```typescript
// Vercel edge caching
export async function GET(request: Request) {
  const data = await fetchData();
  return new Response(JSON.stringify(data), {
    headers: {
      'Content-Type': 'application/json',
      // CDN caches for 60s, browser for 0s
      'Cache-Control': 'public, s-maxage=60, stale-while-revalidate=300',
      'CDN-Cache-Control': 'public, s-maxage=60',
      'Vercel-CDN-Cache-Control': 'public, s-maxage=3600',
    },
  });
}
```

## Browser Storage Caching

```typescript
// LocalStorage cache with TTL
class LocalCache {
  get<T>(key: string): T | null {
    const raw = localStorage.getItem(key);
    if (!raw) return null;

    const { data, expiry } = JSON.parse(raw);
    if (expiry && Date.now() > expiry) {
      localStorage.removeItem(key);
      return null;
    }
    return data as T;
  }

  set<T>(key: string, data: T, ttlMs?: number): void {
    const entry = { data, expiry: ttlMs ? Date.now() + ttlMs : null };
    localStorage.setItem(key, JSON.stringify(entry));
  }

  remove(key: string): void {
    localStorage.removeItem(key);
  }
}

const cache = new LocalCache();
cache.set('user-prefs', { theme: 'dark' }, 7 * 24 * 60 * 60 * 1000); // 7 days
const prefs = cache.get<{ theme: string }>('user-prefs');
```

## Cache Invalidation Patterns

```typescript
// 1. Time-based (TTL)
// Simplest. Set max-age and let it expire.
res.set('Cache-Control', 'public, max-age=300'); // 5 min

// 2. Event-based (invalidate on write)
// After mutation, invalidate affected queries
await queryClient.invalidateQueries({ queryKey: ['posts'] });

// 3. Tag-based (CDN purge)
// Cloudflare/Fastly: tag responses, purge by tag
res.set('Cache-Tag', 'posts, user-123');
// Purge: DELETE /purge?tag=posts

// 4. Version-based (filename hash)
// Build tool adds hash: app.abc123.js
// New deploy = new hash = new file = cache miss

// 5. Stale-while-revalidate
// Serve cached version immediately, fetch fresh in background
res.set('Cache-Control', 'public, max-age=60, stale-while-revalidate=3600');
```

## Gotchas

1. **`staleTime` vs `gcTime` in React Query.** `staleTime` = how long data is considered fresh (won't refetch). `gcTime` = how long inactive data stays in memory. Set `staleTime` based on how fresh data needs to be. Default `staleTime: 0` means every mount triggers a refetch.

2. **Service workers cache aggressively.** A cached service worker serves old HTML even after deploy. Use `skipWaiting()` + `clientsClaim()` for immediate activation, but be careful with breaking changes between old HTML and new JS.

3. **Browser limits localStorage to ~5-10MB.** Exceeding it throws. Always wrap `localStorage.setItem` in try-catch. For larger data, use IndexedDB (via `idb` library).

4. **CDN caching auth endpoints = security breach.** Never cache responses that contain user-specific data with `public` cache-control. Use `private` or `no-store` for authenticated endpoints. Add `Vary: Authorization` if caching per-user at the CDN.

5. **`no-cache` does NOT mean "don't cache".** It means "cache it but always revalidate with the server before using." `no-store` is what actually prevents caching. This is the most common HTTP caching mistake.

6. **React Query deduplication requires stable query keys.** `['users', { page: 1 }]` and `['users', { page: 1 }]` are the SAME key (deep equality). But `['users', new Date()]` creates a new key every render. Keep keys deterministic.
