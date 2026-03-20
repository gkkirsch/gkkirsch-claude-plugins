---
name: trpc-architect
description: >
  Consult on tRPC architecture — router organization, middleware patterns, error handling
  strategy, Next.js integration approach, and performance optimization.
  Triggers: "trpc architecture", "trpc design", "trpc organization", "trpc vs rest",
  "trpc performance", "should I use trpc".
  NOT for: writing specific procedures (use the skills).
tools: Read, Glob, Grep
---

# tRPC Architecture Consultant

## When to Use tRPC

| Scenario | tRPC? | Why |
|----------|-------|-----|
| Full-stack TypeScript (Next.js, T3) | **Yes** | Maximum value — types flow end-to-end |
| Monorepo with shared types | **Yes** | Types shared via packages |
| Internal tools / admin panels | **Yes** | Speed of development, type safety |
| Public API for third parties | **No** | Use REST or GraphQL — clients need a spec |
| Multi-language backend | **No** | tRPC is TypeScript-only |
| Mobile app (React Native) | **Yes** | Works with `@trpc/client` |
| Non-React frontend | **Maybe** | Vanilla client works, but React hooks are the sweet spot |

## tRPC vs Alternatives

| Feature | tRPC | REST | GraphQL |
|---------|------|------|---------|
| Type safety | Full (compile-time) | Manual (OpenAPI codegen) | Partial (codegen) |
| Bundle size | ~2KB client | Varies | ~30KB+ (Apollo) |
| Learning curve | Low (if you know TS) | Low | Medium |
| Caching | Via React Query | Built-in (HTTP) | Apollo/urql cache |
| Real-time | Subscriptions (built-in) | WebSockets (manual) | Subscriptions |
| Tooling | TS compiler | Postman, curl | GraphiQL, Apollo Studio |
| Overfetching | N/A (you control the return) | Common | Solved by design |
| Public API | Not ideal | Standard | Good |

## Router Organization

### Small Apps (< 20 procedures)

```
src/server/
├── trpc.ts              # tRPC init, context, middleware
├── routers/
│   ├── _app.ts          # Root router (merges all)
│   ├── user.ts          # User procedures
│   ├── post.ts          # Post procedures
│   └── auth.ts          # Auth procedures
```

### Medium Apps (20-100 procedures)

```
src/server/
├── trpc.ts              # Base tRPC setup
├── middleware/
│   ├── auth.ts          # Authentication
│   ├── rateLimit.ts     # Rate limiting
│   └── logging.ts       # Request logging
├── routers/
│   ├── _app.ts          # Root router
│   ├── user/
│   │   ├── index.ts     # User router
│   │   ├── user.service.ts
│   │   └── user.schema.ts
│   ├── post/
│   │   ├── index.ts
│   │   ├── post.service.ts
│   │   └── post.schema.ts
```

### Large Apps (100+ procedures)

```
src/server/
├── trpc.ts
├── middleware/
├── routers/
│   ├── _app.ts          # Lazy-loaded routers
│   ├── v1/              # API versioning
│   │   ├── user.ts
│   │   └── post.ts
│   └── v2/
│       └── user.ts
├── services/            # Business logic (framework-agnostic)
├── schemas/             # Zod schemas (shared with client)
```

## Middleware Strategy

```
Request Flow:
  1. Base context (db, session)
  2. Logging middleware (all requests)
  3. Rate limiting middleware (public endpoints)
  4. Auth middleware (protected routes)
  5. Role-based middleware (admin routes)
  6. Procedure handler
```

## Error Handling Strategy

| Error Type | tRPC Code | HTTP Status | When |
|------------|-----------|-------------|------|
| Validation | `BAD_REQUEST` | 400 | Invalid input (Zod catches) |
| Auth required | `UNAUTHORIZED` | 401 | Missing or expired token |
| Permission denied | `FORBIDDEN` | 403 | Valid token, wrong role |
| Not found | `NOT_FOUND` | 404 | Resource doesn't exist |
| Conflict | `CONFLICT` | 409 | Duplicate email, slug, etc. |
| Rate limit | `TOO_MANY_REQUESTS` | 429 | Too many requests |
| Server error | `INTERNAL_SERVER_ERROR` | 500 | Unexpected errors |

## Performance Patterns

1. **Batching**: tRPC batches concurrent requests by default (single HTTP request for multiple procedures called in the same render)
2. **Prefetching**: Use `queryClient.prefetchQuery()` or tRPC's `prefetch()` for anticipated data
3. **Optimistic updates**: Via React Query's `onMutate` — update UI immediately, rollback on error
4. **Select**: Use `.select()` on tRPC queries to transform/narrow response data client-side
5. **Stale-while-revalidate**: Configure `staleTime` and `gcTime` per procedure
6. **Subscriptions**: Use WebSocket transport for real-time, not polling

## Anti-Patterns

| Anti-Pattern | Problem | Fix |
|-------------|---------|-----|
| Putting business logic in procedures | Hard to test, reuse | Extract to service layer |
| Not validating input | Runtime errors, type lies | Always use Zod `.input()` |
| Giant router files | Unmaintainable | Split by domain (user, post, etc.) |
| Returning raw DB models | Leaks internal structure | Use `.output()` or transform |
| Catching all errors silently | Hides bugs | Let tRPC error handler surface them |
| Skipping middleware for auth | Security holes | Use `protectedProcedure` consistently |
| Not using `superjson` | Dates serialize as strings | Add `superjson` transformer |
