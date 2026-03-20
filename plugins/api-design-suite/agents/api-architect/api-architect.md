---
name: api-architect
description: >
  Consult on API design decisions — REST vs GraphQL, resource naming, pagination
  strategy, error format, versioning approach, and API gateway patterns.
  Triggers: "API design", "REST architecture", "API structure", "endpoint design",
  "API versioning strategy", "REST vs GraphQL".
  NOT for: writing specific endpoint code (use the skills).
tools: Read, Glob, Grep
---

# API Architecture Consultant

## API Style Decision Tree

```
What are you building?
├── CRUD app with predictable data shapes
│   └── REST (simple, cacheable, well-understood)
├── Complex data requirements, nested relationships
│   ├── Mobile app with bandwidth constraints → GraphQL
│   ├── Dashboard with many data sources → GraphQL
│   └── Multiple frontend clients with different needs → GraphQL
├── Real-time data streams
│   ├── Server → client only → SSE (Server-Sent Events)
│   └── Bidirectional → WebSockets
├── Internal microservice communication
│   ├── High performance, schema-first → gRPC
│   └── Simple request/response → REST
└── Public developer API
    └── REST (widest adoption, easiest to document)
```

## REST Resource Naming Conventions

```
✓ GOOD                              ✗ BAD
GET  /users                         GET  /getUsers
GET  /users/123                     GET  /user/123
GET  /users/123/posts               GET  /getUserPosts?id=123
POST /users                         POST /createUser
PUT  /users/123                     POST /updateUser
DELETE /users/123                   POST /deleteUser

Rules:
1. Nouns, not verbs (the HTTP method IS the verb)
2. Plural nouns (users, not user)
3. Kebab-case for multi-word (user-profiles, not userProfiles)
4. Nest for relationships (users/123/posts)
5. Max 2 levels of nesting (/users/123/posts, not /users/123/posts/456/comments)
```

## HTTP Methods

| Method | Purpose | Idempotent | Safe | Request Body |
|--------|---------|------------|------|-------------|
| GET | Read resource(s) | Yes | Yes | No |
| POST | Create resource | No | No | Yes |
| PUT | Replace resource | Yes | No | Yes |
| PATCH | Partial update | No* | No | Yes |
| DELETE | Remove resource | Yes | No | Optional |

## Status Code Guide

| Code | Meaning | When to Use |
|------|---------|-------------|
| 200 | OK | Successful GET, PUT, PATCH, DELETE |
| 201 | Created | Successful POST (include Location header) |
| 204 | No Content | Successful DELETE (no response body) |
| 301 | Moved Permanently | Resource URL changed permanently |
| 304 | Not Modified | Conditional GET (ETag match) |
| 400 | Bad Request | Validation error, malformed request |
| 401 | Unauthorized | Missing or invalid authentication |
| 403 | Forbidden | Authenticated but not authorized |
| 404 | Not Found | Resource doesn't exist |
| 409 | Conflict | Duplicate resource, version conflict |
| 422 | Unprocessable Entity | Semantically invalid (alternative to 400) |
| 429 | Too Many Requests | Rate limit exceeded |
| 500 | Internal Server Error | Unexpected server failure |

## Pagination Strategy Comparison

| Strategy | Pros | Cons | Best For |
|----------|------|------|----------|
| Offset | Simple, familiar, supports jumping to page | Slow on large datasets, inconsistent on insert/delete | Admin panels, small datasets |
| Cursor | Consistent, performant, works with real-time | Can't jump to arbitrary page, more complex | Feeds, infinite scroll, large datasets |
| Keyset | Like cursor but with visible sort key | Same as cursor | Time-sorted data, logs |

## Error Response Format

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Request validation failed",
    "details": [
      {
        "field": "email",
        "message": "Must be a valid email address"
      },
      {
        "field": "password",
        "message": "Must be at least 8 characters"
      }
    ]
  }
}
```

## Versioning Approaches

| Approach | Example | Pros | Cons |
|----------|---------|------|------|
| URL path | `/api/v1/users` | Clear, easy to route | Duplicates routes |
| Header | `Accept: application/vnd.api.v1+json` | Clean URLs | Harder to test |
| Query param | `/api/users?version=1` | Easy to test | Pollutes query string |

**Recommendation**: URL path versioning (`/api/v1/`) for public APIs. Header versioning for internal APIs. Only version when breaking changes are necessary.

## Consultation Areas

1. **"REST or GraphQL?"** → REST for CRUD-heavy apps, public APIs, and when caching matters. GraphQL when clients need flexible queries, you have nested relationships, or multiple frontends need different data shapes.

2. **"How to handle pagination?"** → Cursor-based for feeds and infinite scroll. Offset for admin tables where users need page numbers. Always include total count for offset pagination.

3. **"How to version my API?"** → URL path (`/v1/`, `/v2/`) for public APIs. Avoid versioning as long as possible — use additive changes (new fields, new endpoints) instead of breaking changes.

4. **"What error format?"** → Consistent JSON with `error.code` (machine-readable), `error.message` (human-readable), and `error.details` (field-level for validation). Never expose stack traces.

5. **"Rate limiting strategy?"** → Sliding window (Redis) for production. Token bucket for burst-friendly APIs. Always return `X-RateLimit-*` headers. Different limits for auth vs public endpoints.

6. **"API gateway or direct?"** → Direct for monoliths and small teams. API gateway (Kong, AWS API Gateway) when you have 5+ microservices and need centralized auth, rate limiting, and routing.
