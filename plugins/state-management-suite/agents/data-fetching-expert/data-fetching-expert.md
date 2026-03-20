---
name: data-fetching-expert
description: >
  Expert on data fetching patterns — caching strategies, optimistic updates,
  infinite scroll, prefetching, and background synchronization.
  Triggers: "data fetching pattern", "caching strategy", "optimistic update",
  "infinite scroll", "prefetching", "stale while revalidate".
  NOT for: specific React Query implementation (use react-query skill).
tools: Read, Glob, Grep
---

# Data Fetching Patterns

## Caching Strategies

### Stale-While-Revalidate

The default React Query strategy. Serve cached data immediately, refetch in the background.

```
Request → Cache has data?
├── Yes → Return cached data (stale)
│         └── Background refetch → Update cache → Re-render with fresh data
└── No  → Fetch from server → Cache response → Return data
```

**When to use**: Most data. Users see instant results, data stays fresh.
**Trade-off**: Brief moment of stale data (usually acceptable).

### Cache-First (Long TTL)

Data that rarely changes. Serve from cache, only refetch after long intervals.

```typescript
{
  staleTime: 24 * 60 * 60 * 1000,  // 24 hours
  gcTime: 7 * 24 * 60 * 60 * 1000, // 7 days
}
```

**When to use**: Config, feature flags, translations, user profile.
**Trade-off**: Changes take hours to propagate.

### Network-First (Short TTL)

Data that changes frequently. Always fetch fresh, but show cached during loading.

```typescript
{
  staleTime: 0,  // Always stale → always refetch
  refetchInterval: 30_000,  // Poll every 30s
}
```

**When to use**: Notifications, live dashboards, chat messages.
**Trade-off**: More network requests.

### Network-Only (No Cache)

Every request hits the server. No caching at all.

```typescript
{
  staleTime: 0,
  gcTime: 0,
  refetchOnMount: 'always',
  refetchOnWindowFocus: 'always',
}
```

**When to use**: Sensitive data (account balance, real-time pricing).
**Trade-off**: Slower UX, more load on server.

## Optimistic Updates

Show the result immediately, roll back if the server rejects it.

### The Pattern

```
1. User clicks "Like"
2. Immediately update UI (like count +1, heart filled)
3. Send request to server
4. Server responds:
   ├── Success → Cache is already correct, done
   └── Failure → Roll back to previous state, show error
```

### When to Use Optimistic Updates

| Action | Optimistic? | Why |
|--------|------------|-----|
| Like/favorite | Yes | Low risk, instant feedback matters |
| Add to cart | Yes | Common action, easy to undo |
| Send message | Yes | Users expect instant send |
| Delete item | Cautious | Show "undo" toast instead of instant delete |
| Payment | Never | Irreversible, must confirm server success |
| Create account | Never | Complex validation, server-side checks |

### Error Recovery Strategies

```
Optimistic update fails:
├── Rollback + toast → "Failed to save. Changes reverted."
├── Retry + toast → "Connection lost. Retrying..."
├── Retry silently → Background retry, user doesn't notice
└── Queue for later → "You're offline. Changes will sync when connected."
```

## Pagination Patterns

### Offset Pagination
```
Page 1: GET /api/items?page=1&limit=20
Page 2: GET /api/items?page=2&limit=20
```
- Simple, supports "jump to page 5"
- Can miss or duplicate items if data changes between pages
- Slow on large datasets (OFFSET scans rows)

### Cursor Pagination
```
First:  GET /api/items?limit=20
Next:   GET /api/items?cursor=eyJpZCI6MjB9&limit=20
```
- Consistent results even as data changes
- No page jumping (forward/back only)
- Fast on any dataset size (indexed seek)

### Infinite Scroll vs Load More

| Feature | Infinite Scroll | Load More Button |
|---------|----------------|-----------------|
| UX | Seamless (feels modern) | Explicit (user controls) |
| Performance | Risk of memory issues | Predictable memory |
| Accessibility | Harder (focus management) | Easier |
| SEO | Poor (content invisible) | Poor (same issue) |
| Best for | Social feeds, image grids | Search results, data tables |

## Prefetching

### On Hover
```
Prefetch data when user hovers over a link. ~300ms head start.
Best for: Navigation items, list items that lead to detail pages.
```

### On Route Mount
```
Parent route prefetches data the child route will need.
Best for: Dashboard → Detail page transitions.
```

### Based on Viewport
```
Prefetch data for items about to scroll into view (Intersection Observer).
Best for: Long lists, image galleries, virtual scroll.
```

### Waterfall Prevention
```
Bad:  Component mounts → fetch parent → render children → fetch child data
Good: Fetch parent AND child data in parallel at route level

Solution: Prefetch child queries when parent query succeeds,
or use React Query's useQueries for parallel fetching.
```

## Background Sync

### Polling
```typescript
// Simple interval polling
{ refetchInterval: 30_000 }  // Every 30 seconds

// Conditional polling (only when tab is visible)
{ refetchInterval: 30_000, refetchIntervalInBackground: false }
```

### Window Focus Refetch
```typescript
// Refetch when user tabs back (default in React Query)
{ refetchOnWindowFocus: true }
```

### WebSocket + Cache Invalidation
```
WebSocket event received → invalidate relevant query → auto-refetch
Best for: Real-time features (chat, notifications, live updates)
```

### Event-Driven Invalidation
```
User action (create/update/delete) → invalidate related queries

Example: Create a post → invalidate ['posts'] query → list refetches
```

## Query Key Design

### Hierarchical Keys
```typescript
// All users
['users']

// Users with filters
['users', { role: 'admin', status: 'active' }]

// Single user
['users', userId]

// User's posts
['users', userId, 'posts']

// User's single post
['users', userId, 'posts', postId]
```

### Invalidation Cascading
```typescript
// Invalidate everything under 'users'
queryClient.invalidateQueries({ queryKey: ['users'] })
// Matches: ['users'], ['users', 1], ['users', 1, 'posts'], etc.

// Invalidate only the list
queryClient.invalidateQueries({ queryKey: ['users'], exact: true })
// Matches: ['users'] only
```

## Error Handling in Data Fetching

### Retry Strategy
```
Attempt 1: Immediate
Attempt 2: Wait 1 second
Attempt 3: Wait 2 seconds
Attempt 4: Wait 4 seconds (give up)

Don't retry: 401, 403, 404, 422 (client errors)
Always retry: 500, 502, 503 (server/transient errors)
```

### Error Boundaries
```
Global: catch unhandled errors, show "something went wrong" page
Per-section: catch errors in a dashboard widget, show error + retry button
Per-query: handle specific error in component (e.g., "post not found" → redirect)
```

### Loading States
```
First load:  Show skeleton
Background refetch: Keep showing stale data (no loader)
Error + retry: Show error + stale data + retry button
Offline: Show cached data + "offline" banner
```
