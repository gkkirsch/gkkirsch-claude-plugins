# REST API Design Checklist

Quick reference for reviewing and shipping REST APIs.

---

## Pre-Launch Checklist

### URL Design
- [ ] Plural nouns for collections (`/users`, `/orders`)
- [ ] Lowercase with hyphens (`/user-profiles`, not `/userProfiles`)
- [ ] No trailing slashes
- [ ] No verbs in URLs (use HTTP methods instead)
- [ ] Nesting limited to 2 levels max
- [ ] Action endpoints use POST (`POST /orders/:id/cancel`)

### HTTP Methods
- [ ] GET for reads (safe, idempotent, cacheable)
- [ ] POST for creates and actions (not idempotent)
- [ ] PUT for full replacements (idempotent)
- [ ] PATCH for partial updates
- [ ] DELETE for removals (idempotent)

### Status Codes
- [ ] 200 for successful GET/PUT/PATCH/DELETE
- [ ] 201 for successful POST creates (with `Location` header)
- [ ] 204 for successful DELETE with no body
- [ ] 400 for malformed requests
- [ ] 401 for missing/invalid authentication
- [ ] 403 for valid auth but insufficient permissions
- [ ] 404 for non-existent resources
- [ ] 409 for conflicts (duplicate email, version mismatch)
- [ ] 422 for business rule violations
- [ ] 429 for rate limits (with `Retry-After` header)
- [ ] 500 for unexpected server errors (never leak details)

### Request/Response
- [ ] `Content-Type: application/json` on all endpoints
- [ ] Consistent error response format (code, message, details, requestId)
- [ ] `X-Request-Id` header for correlation
- [ ] No stack traces in production error responses
- [ ] `--json` or `Accept` header support for machine-readable output

### Pagination
- [ ] Default limit set (e.g., 20)
- [ ] Maximum limit enforced (e.g., 100)
- [ ] `hasNextPage`/`hasPrevPage` booleans in response
- [ ] Cursor-based for public APIs / large datasets
- [ ] Total count included (offset) or omitted intentionally (cursor)

### Filtering & Sorting
- [ ] Query parameters for filtering (`?status=active&role=admin`)
- [ ] Consistent sort format (`?sort=createdAt:desc`)
- [ ] Whitelist allowed sort fields (don't let clients sort by arbitrary columns)
- [ ] Search via `?search=term` or `?q=term`

### Authentication & Security
- [ ] Bearer token authentication (`Authorization: Bearer <token>`)
- [ ] Rate limiting on all endpoints (stricter on auth endpoints)
- [ ] Input validation on all request bodies (Zod)
- [ ] SQL injection prevention (parameterized queries)
- [ ] No sensitive data in URLs (tokens, passwords)
- [ ] CORS configured appropriately

### Versioning
- [ ] Version strategy decided (URL path recommended: `/v1/`)
- [ ] Deprecation headers when sunsetting old versions
- [ ] Migration guide for version transitions

### Documentation
- [ ] OpenAPI spec generated and served
- [ ] Interactive docs available (Swagger UI, Scalar, or Redoc)
- [ ] All endpoints documented with examples
- [ ] Error codes documented
- [ ] Authentication flow documented
- [ ] Rate limits documented

---

## HTTP Method Quick Reference

```
Collection: /resources
  GET    → List (200, paginated)
  POST   → Create (201 + Location header)

Instance: /resources/:id
  GET    → Read (200 or 404)
  PUT    → Replace (200 or 404)
  PATCH  → Update (200 or 404)
  DELETE → Delete (204 or 404)

Action: /resources/:id/action
  POST   → Execute (200 or 202 for async)
```

---

## Status Code Decision Tree

```
Request succeeded?
├── Created a resource? → 201 Created
├── No response body needed? → 204 No Content
└── Returning data? → 200 OK

Request failed?
├── Client's fault?
│   ├── Bad JSON/missing fields? → 400 Bad Request
│   ├── Not authenticated? → 401 Unauthorized
│   ├── Authenticated but not allowed? → 403 Forbidden
│   ├── Resource doesn't exist? → 404 Not Found
│   ├── Duplicate/conflict? → 409 Conflict
│   ├── Valid request but business rule violated? → 422 Unprocessable
│   └── Too many requests? → 429 Too Many Requests
└── Server's fault?
    ├── Unexpected error? → 500 Internal Server Error
    ├── Upstream service down? → 502 Bad Gateway
    ├── Maintenance/overload? → 503 Service Unavailable
    └── Upstream timeout? → 504 Gateway Timeout
```

---

## Common Query Parameter Patterns

| Parameter | Format | Example |
|-----------|--------|---------|
| Pagination | `page=1&limit=20` | `GET /users?page=2&limit=50` |
| Cursor | `cursor=abc123&limit=20` | `GET /users?cursor=eyJpZCI6MTB9&limit=20` |
| Sort | `sort=field:direction` | `GET /users?sort=createdAt:desc` |
| Multi-sort | `sort=field1:dir,field2:dir` | `GET /users?sort=role:asc,name:asc` |
| Filter (equality) | `field=value` | `GET /users?role=admin` |
| Filter (range) | `minField=n&maxField=n` | `GET /products?minPrice=10&maxPrice=100` |
| Filter (multiple values) | `field=a,b,c` | `GET /users?role=admin,moderator` |
| Search | `search=term` or `q=term` | `GET /products?q=wireless+headphones` |
| Sparse fields | `fields=a,b,c` | `GET /users?fields=id,name,email` |
| Include relations | `include=relation` | `GET /users?include=orders,profile` |
| Date range | `from=date&to=date` | `GET /events?from=2026-01-01&to=2026-03-31` |

---

## Rate Limit Headers

```
# On every response:
RateLimit-Limit: 100                → Max requests per window
RateLimit-Remaining: 42             → Requests left in current window
RateLimit-Reset: 1711065600         → Unix timestamp when window resets

# On 429 responses only:
Retry-After: 30                     → Seconds until next request is allowed
```

---

## Error Response Templates

### Validation Error (422)
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "details": [
      { "field": "email", "message": "Must be a valid email address" },
      { "field": "password", "message": "Must be at least 8 characters" }
    ],
    "requestId": "req_abc123"
  }
}
```

### Authentication Error (401)
```json
{
  "error": {
    "code": "TOKEN_EXPIRED",
    "message": "Your session has expired. Please log in again.",
    "requestId": "req_def456"
  }
}
```

### Not Found (404)
```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "User 'usr_789' not found",
    "requestId": "req_ghi789"
  }
}
```

### Rate Limited (429)
```json
{
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Rate limit exceeded. Try again in 30 seconds.",
    "retryAfter": 30,
    "requestId": "req_jkl012"
  }
}
```

### Server Error (500)
```json
{
  "error": {
    "code": "INTERNAL_ERROR",
    "message": "An unexpected error occurred",
    "requestId": "req_mno345"
  }
}
```

---

## CORS Configuration

```typescript
import cors from 'cors';

// Development: allow all
app.use(cors());

// Production: explicit origins
app.use(cors({
  origin: ['https://app.example.com', 'https://admin.example.com'],
  methods: ['GET', 'POST', 'PUT', 'PATCH', 'DELETE'],
  allowedHeaders: ['Content-Type', 'Authorization', 'X-Request-Id'],
  exposedHeaders: ['RateLimit-Limit', 'RateLimit-Remaining', 'RateLimit-Reset'],
  credentials: true,
  maxAge: 86400, // Cache preflight for 24 hours
}));
```

---

## Security Headers

```typescript
import helmet from 'helmet';

app.use(helmet());

// Or manually:
app.use((req, res, next) => {
  res.set('X-Content-Type-Options', 'nosniff');
  res.set('X-Frame-Options', 'DENY');
  res.set('Strict-Transport-Security', 'max-age=31536000; includeSubDomains');
  res.set('Cache-Control', 'no-store'); // For authenticated API responses
  next();
});
```
