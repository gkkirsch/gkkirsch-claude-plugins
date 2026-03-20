# API Design Cheatsheet

## HTTP Methods

| Method | Purpose | Idempotent | Request Body | Response |
|--------|---------|-----------|-------------|----------|
| GET | Read resource(s) | Yes | No | 200 + data |
| POST | Create resource | No | Yes | 201 + Location header |
| PUT | Full replace | Yes | Yes | 200 + updated data |
| PATCH | Partial update | No* | Yes | 200 + updated data |
| DELETE | Remove resource | Yes | No | 204 (no body) |

## Status Codes

| Code | When to Use | Response Body |
|------|------------|---------------|
| 200 | Success (GET, PUT, PATCH) | Resource data |
| 201 | Created (POST) | Created resource + Location header |
| 204 | Success, no content (DELETE) | Empty |
| 301 | Permanent redirect | Empty + Location header |
| 304 | Not modified (caching) | Empty |
| 400 | Bad request / validation error | Error details with field-level errors |
| 401 | Not authenticated | `{ error: { code: "UNAUTHENTICATED" } }` |
| 403 | Authenticated but not authorized | `{ error: { code: "FORBIDDEN" } }` |
| 404 | Resource not found | `{ error: { code: "NOT_FOUND" } }` |
| 409 | Conflict (duplicate, version mismatch) | `{ error: { code: "CONFLICT" } }` |
| 422 | Unprocessable entity (semantic error) | Error details |
| 429 | Rate limited | `Retry-After` header + error |
| 500 | Server error | Generic message (no stack traces) |

## URL Naming Conventions

```
GOOD:
  GET    /api/v1/users              # List users
  POST   /api/v1/users              # Create user
  GET    /api/v1/users/:id          # Get user
  PUT    /api/v1/users/:id          # Replace user
  PATCH  /api/v1/users/:id          # Update user
  DELETE /api/v1/users/:id          # Delete user
  GET    /api/v1/users/:id/posts    # User's posts (nested)
  POST   /api/v1/users/:id/posts    # Create post for user

BAD:
  GET    /api/v1/getUsers           # Verb in URL
  POST   /api/v1/user/create        # Verb + singular
  GET    /api/v1/User/:id           # PascalCase
  DELETE /api/v1/users/:id/delete   # Redundant verb
  GET    /api/v1/users_list         # Snake_case
```

## Error Response Format

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Request validation failed",
    "details": [
      { "field": "email", "message": "Invalid email format" },
      { "field": "name", "message": "Name is required" }
    ],
    "requestId": "req_abc123"
  }
}
```

## Pagination Patterns

### Offset-Based (Simple)

```
GET /api/users?page=2&limit=20

Response:
{
  "data": [...],
  "pagination": {
    "page": 2,
    "limit": 20,
    "total": 150,
    "totalPages": 8,
    "hasMore": true
  }
}
```

### Cursor-Based (Scalable)

```
GET /api/users?limit=20&after=eyJpZCI6MTAwfQ

Response:
{
  "data": [...],
  "pagination": {
    "hasNextPage": true,
    "hasPreviousPage": true,
    "startCursor": "eyJpZCI6ODF9",
    "endCursor": "eyJpZCI6MTAwfQ"
  }
}
```

| Factor | Offset | Cursor |
|--------|--------|--------|
| Implementation | Simple | Complex |
| "Jump to page N" | Yes | No |
| Consistent with inserts | No (page drift) | Yes |
| Performance at scale | Degrades | Constant |
| Best for | Admin panels, small datasets | Feeds, infinite scroll |

## Filtering & Sorting

```
GET /api/users?role=admin&status=active           # Exact match
GET /api/users?search=john                        # Text search
GET /api/users?created_after=2026-01-01           # Date range
GET /api/users?sort=created_at:desc,name:asc      # Multi-sort
GET /api/users?fields=id,name,email               # Sparse fields
GET /api/users?include=posts,profile              # Eager load relations
```

## Versioning Strategies

| Strategy | Format | Pros | Cons |
|----------|--------|------|------|
| URL path | `/api/v1/users` | Simple, clear, cacheable | URL changes |
| Header | `Accept: application/vnd.api.v2+json` | Clean URLs | Hidden, harder to test |
| Query param | `/api/users?version=2` | Easy to test | Caching issues |

**Recommendation**: URL path versioning. It's the most explicit and debuggable.

## Rate Limiting Headers

```
X-RateLimit-Limit: 100        # Max requests in window
X-RateLimit-Remaining: 42     # Requests left
X-RateLimit-Reset: 1234567890 # Window reset (Unix timestamp)
Retry-After: 30               # Seconds until retry (on 429)
```

## Content Negotiation

```
# Request
Accept: application/json          # Client wants JSON
Content-Type: application/json    # Client sends JSON

# Response
Content-Type: application/json; charset=utf-8

# File upload
Content-Type: multipart/form-data

# Conditional requests (caching)
If-None-Match: "etag-value"       # Client has cached version
If-Modified-Since: Wed, 01 Jan 2026 00:00:00 GMT
```

## Authentication Patterns

```
# Bearer token (JWT)
Authorization: Bearer eyJhbGciOi...

# API key (header)
X-API-Key: sk_live_abc123

# API key (query — NOT recommended, gets logged)
GET /api/data?api_key=sk_live_abc123

# Basic auth (Base64)
Authorization: Basic dXNlcjpwYXNz
```

## GraphQL vs REST Decision

| Factor | REST | GraphQL |
|--------|------|---------|
| Data shape | Server decides | Client decides |
| Over/under-fetching | Common | Solved |
| Caching | HTTP caching (simple) | Custom (complex) |
| File uploads | Native multipart | Requires workaround |
| Real-time | SSE / WebSockets | Subscriptions |
| Learning curve | Low | Medium |
| Tooling | curl, Postman | Apollo, Relay |
| N+1 problem | ORM handles | DataLoader needed |
| Best for | CRUD, public APIs | Complex UIs, mobile |

## OpenAPI Quick Reference

```yaml
openapi: 3.1.0
info:
  title: My API
  version: 1.0.0

paths:
  /users:
    get:
      summary: List users
      operationId: listUsers
      tags: [Users]
      parameters:
        - name: page
          in: query
          schema: { type: integer, default: 1 }
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/UserList'

components:
  schemas:
    User:
      type: object
      required: [id, name, email]
      properties:
        id: { type: string, format: uuid }
        name: { type: string }
        email: { type: string, format: email }
```

## Common Anti-Patterns

| Anti-Pattern | Problem | Better Approach |
|-------------|---------|-----------------|
| `POST /getUsers` | Verb in URL | `GET /users` |
| `200` for everything | Hides errors | Use appropriate status codes |
| Returning `null` for not found | Ambiguous | Return `404` with error body |
| Nested resources > 2 levels | Complex URLs | `/posts/:id` not `/users/:uid/posts/:pid/comments` |
| Different error formats | Hard to parse | Consistent error envelope |
| No pagination on list endpoints | Memory issues | Always paginate, default limit 20 |
| Exposing DB IDs | Security leak | Use UUIDs or slugs |
| Ignoring `Accept` header | Bad API citizenship | Content negotiation |

## Webhook Design

```json
// Webhook payload format
{
  "id": "evt_abc123",
  "type": "user.created",
  "created_at": "2026-01-15T10:30:00Z",
  "data": {
    "id": "usr_xyz",
    "email": "new@user.com"
  }
}

// Webhook headers
X-Webhook-Signature: sha256=abc123  // HMAC signature for verification
X-Webhook-ID: evt_abc123           // Idempotency key
X-Webhook-Timestamp: 1234567890    // Prevent replay attacks
```

## HATEOAS Links (Optional)

```json
{
  "data": { "id": "123", "name": "Alice" },
  "links": {
    "self": "/api/users/123",
    "posts": "/api/users/123/posts",
    "avatar": "/api/users/123/avatar"
  }
}
```

## API Testing Commands

```bash
# GET with auth
curl -H "Authorization: Bearer $TOKEN" https://api.example.com/users

# POST with JSON
curl -X POST https://api.example.com/users \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"name":"Alice","email":"alice@example.com"}'

# PUT (full replace)
curl -X PUT https://api.example.com/users/123 \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice Updated","email":"alice@new.com"}'

# DELETE
curl -X DELETE https://api.example.com/users/123 \
  -H "Authorization: Bearer $TOKEN" -v

# Check response headers
curl -I https://api.example.com/users
```
