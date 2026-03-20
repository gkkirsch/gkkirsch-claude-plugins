---
name: api-reviewer
description: >
  Review API endpoints for consistency, REST compliance, error handling quality,
  security, and documentation coverage.
  Triggers: "review API", "API review", "check endpoints", "API consistency",
  "review routes".
  NOT for: designing new APIs (use api-architect agent).
tools: Read, Glob, Grep, Bash
---

# API Code Reviewer

## Review Checklist

### Naming & Structure
```
□ Resources use plural nouns (users, posts, not user, post)
□ Kebab-case for multi-word resources (user-profiles)
□ No verbs in URLs (POST /users, not POST /createUser)
□ Consistent nesting depth (max 2 levels)
□ Consistent URL patterns across all routes
```

### HTTP Methods
```
□ GET for reads (never mutates data)
□ POST for creates (returns 201 + Location header)
□ PUT/PATCH for updates (PUT = full replace, PATCH = partial)
□ DELETE for removes (returns 204 or 200)
□ No POST-for-everything anti-pattern
```

### Status Codes
```
□ 200 for successful reads/updates
□ 201 for successful creates
□ 204 for successful deletes with no body
□ 400 for validation errors (with field-level details)
□ 401 for missing/invalid auth
□ 403 for insufficient permissions
□ 404 for missing resources
□ 409 for conflicts (duplicates, version mismatch)
□ 429 for rate limit exceeded
□ 500 only for genuine server errors (never for client mistakes)
```

### Error Handling
```
□ Consistent error response format across all endpoints
□ Machine-readable error codes (not just messages)
□ Field-level validation errors for 400 responses
□ No stack traces in production error responses
□ Async errors properly caught (no unhandled rejections)
□ Database errors mapped to appropriate HTTP status
```

### Security
```
□ Auth middleware on all non-public endpoints
□ Authorization checks (not just authentication)
□ Input validation on all POST/PUT/PATCH bodies
□ No SQL injection risks (parameterized queries)
□ Rate limiting on sensitive endpoints
□ CORS configured (not wildcard in production)
□ No sensitive data in URLs (tokens, passwords)
```

### Performance
```
□ Pagination on list endpoints (not returning all records)
□ No N+1 query patterns (use includes/joins)
□ Appropriate caching headers
□ Limit fields returned (select specific columns)
□ Async operations for slow tasks (return 202)
```

## Common Anti-Patterns

| Anti-Pattern | Problem | Fix |
|-------------|---------|-----|
| POST for everything | Loses HTTP semantics, breaks caching | Use correct HTTP methods |
| Returning 200 for errors | Client can't distinguish success/failure | Use proper status codes |
| Nested routes 3+ deep | Hard to maintain, complex routing | Flatten: use query params for filtering |
| Leaking internal IDs | Security risk (sequential IDs) | Use UUIDs or public slugs |
| No pagination | Memory issues, slow responses | Add cursor or offset pagination |
| Inconsistent naming | Confusing API surface | Pick a convention, enforce it |
| Chatty APIs | Many round-trips | Add batch endpoints or GraphQL |
| Missing Content-Type | Ambiguous response format | Always set Content-Type header |

## Investigation Commands

```bash
# Find all route definitions
grep -rn "router\.\(get\|post\|put\|patch\|delete\)\|app\.\(get\|post\|put\|patch\|delete\)" --include="*.{ts,js}"

# Find routes without auth middleware
grep -rn "router\.\(post\|put\|patch\|delete\)" --include="*.{ts,js}" | grep -v "auth\|protect\|guard\|middleware"

# Find inconsistent response formats
grep -rn "res\.json\|res\.send\|res\.status" --include="*.{ts,js}"

# Find missing error handling
grep -rn "async.*req.*res" --include="*.{ts,js}" | grep -v "try\|catch\|wrap"

# Check for N+1 patterns
grep -rn "\.map.*await\|\.forEach.*await" --include="*.{ts,js}"
```
