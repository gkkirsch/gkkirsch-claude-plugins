# HTTP Methods and Status Codes Reference

Comprehensive reference for HTTP methods, status codes, headers, and protocol semantics
for API testing and development. Covers HTTP/1.1, HTTP/2, and HTTP/3 with real-world examples.

---

## Table of Contents

- [HTTP Methods](#http-methods)
  - [GET](#get)
  - [POST](#post)
  - [PUT](#put)
  - [PATCH](#patch)
  - [DELETE](#delete)
  - [HEAD](#head)
  - [OPTIONS](#options)
  - [TRACE](#trace)
- [Method Properties](#method-properties)
- [Status Codes](#status-codes)
  - [1xx Informational](#1xx-informational)
  - [2xx Success](#2xx-success)
  - [3xx Redirection](#3xx-redirection)
  - [4xx Client Errors](#4xx-client-errors)
  - [5xx Server Errors](#5xx-server-errors)
- [Headers](#headers)
  - [Request Headers](#request-headers)
  - [Response Headers](#response-headers)
  - [Content Negotiation](#content-negotiation)
  - [Caching Headers](#caching-headers)
  - [Security Headers](#security-headers)
  - [CORS Headers](#cors-headers)
- [HTTP/2 Considerations](#http2-considerations)
- [HTTP/3 Considerations](#http3-considerations)
- [Testing Patterns](#testing-patterns)

---

## HTTP Methods

### GET

**Purpose:** Retrieve a resource or collection of resources.

**Properties:**
- Safe: Yes (does not modify server state)
- Idempotent: Yes (multiple identical requests produce the same result)
- Cacheable: Yes
- Request body: Not expected (some servers ignore it)

**Semantics:**
- Used for reading data only — MUST NOT cause side effects
- The response is a representation of the target resource
- Can be conditional (If-None-Match, If-Modified-Since)
- Results can be cached by intermediaries (CDNs, proxies)

**Common patterns:**

```http
# Get a single resource
GET /api/users/123 HTTP/1.1
Host: api.example.com
Accept: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...

# Response
HTTP/1.1 200 OK
Content-Type: application/json
ETag: "abc123"
Cache-Control: private, max-age=60

{
  "id": "123",
  "name": "Jane Doe",
  "email": "jane@example.com"
}
```

```http
# Get a collection with filtering, sorting, and pagination
GET /api/products?category=electronics&sort=price&order=asc&page=2&pageSize=20 HTTP/1.1
Host: api.example.com
Accept: application/json

# Response
HTTP/1.1 200 OK
Content-Type: application/json
X-Total-Count: 142
Link: </api/products?page=3&pageSize=20>; rel="next", </api/products?page=1&pageSize=20>; rel="prev"

{
  "data": [...],
  "pagination": {
    "page": 2,
    "pageSize": 20,
    "totalItems": 142,
    "totalPages": 8
  }
}
```

```http
# Conditional GET (returns 304 if unchanged)
GET /api/products/456 HTTP/1.1
Host: api.example.com
If-None-Match: "abc123"

# Response (resource hasn't changed)
HTTP/1.1 304 Not Modified
ETag: "abc123"
```

**Testing considerations:**
- Verify the response body matches the expected schema
- Test with query parameters (filtering, sorting, pagination)
- Test conditional requests (If-None-Match, If-Modified-Since)
- Verify caching headers are set correctly
- Test with invalid/missing path parameters (should return 400 or 404)
- Verify authorization — can the user access this resource?
- Check that GET requests do not modify state

---

### POST

**Purpose:** Create a new resource or trigger a server-side action.

**Properties:**
- Safe: No (modifies server state)
- Idempotent: No (multiple requests may create multiple resources)
- Cacheable: Only if explicit freshness info is included
- Request body: Expected

**Semantics:**
- Creates a new subordinate resource under the target collection
- The server assigns the resource's URI and returns it in the Location header
- Can also be used for actions that don't fit other methods (RPC-style)
- The response typically includes the created resource representation

**Common patterns:**

```http
# Create a new resource
POST /api/users HTTP/1.1
Host: api.example.com
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...

{
  "email": "john@example.com",
  "name": "John Doe",
  "password": "SecurePass123!"
}

# Response
HTTP/1.1 201 Created
Content-Type: application/json
Location: /api/users/789

{
  "id": "789",
  "email": "john@example.com",
  "name": "John Doe",
  "createdAt": "2025-03-15T10:00:00Z"
}
```

```http
# Trigger an action (non-CRUD)
POST /api/emails/send HTTP/1.1
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...

{
  "to": "customer@example.com",
  "template": "welcome",
  "data": { "name": "John" }
}

# Response
HTTP/1.1 202 Accepted
Content-Type: application/json

{
  "messageId": "msg-abc123",
  "status": "queued",
  "estimatedDelivery": "2025-03-15T10:01:00Z"
}
```

```http
# Idempotent POST with Idempotency-Key
POST /api/payments HTTP/1.1
Content-Type: application/json
Idempotency-Key: payment-abc-123
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...

{
  "amount": 99.99,
  "currency": "USD",
  "customerId": "cust-456"
}

# Second request with same Idempotency-Key returns same result
# without creating a duplicate payment
```

**Testing considerations:**
- Verify 201 Created with Location header for resource creation
- Verify the response body matches the created resource
- Test with invalid request body (missing fields, wrong types)
- Test duplicate creation (unique constraint violations → 409)
- Test without authentication (→ 401)
- Test without authorization (→ 403)
- Verify the resource actually exists after creation (GET it)
- Test idempotency key behavior if supported
- Verify request body size limits (413 Payload Too Large)

---

### PUT

**Purpose:** Replace a resource entirely with the provided representation.

**Properties:**
- Safe: No
- Idempotent: Yes (multiple identical PUTs produce the same result)
- Cacheable: No
- Request body: Expected (full resource representation)

**Semantics:**
- Replaces the ENTIRE resource — all fields must be provided
- If the resource doesn't exist, the server may create it (201) or return 404
- The result of a PUT is the same regardless of how many times it's called
- Missing optional fields in the request are set to their defaults (not preserved)

**Common patterns:**

```http
# Full resource replacement
PUT /api/products/456 HTTP/1.1
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...

{
  "name": "Updated Product Name",
  "price": 49.99,
  "description": "New description",
  "category": "electronics",
  "stock": 100,
  "tags": ["updated", "electronics"]
}

# Response
HTTP/1.1 200 OK
Content-Type: application/json

{
  "id": "456",
  "name": "Updated Product Name",
  "price": 49.99,
  "description": "New description",
  "category": "electronics",
  "stock": 100,
  "tags": ["updated", "electronics"],
  "updatedAt": "2025-03-15T10:30:00Z"
}
```

```http
# Conditional PUT (optimistic concurrency)
PUT /api/products/456 HTTP/1.1
Content-Type: application/json
If-Match: "etag-abc123"
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...

{
  "name": "Updated Name",
  "price": 49.99,
  ...
}

# Response if ETag matches
HTTP/1.1 200 OK

# Response if ETag doesn't match (concurrent modification)
HTTP/1.1 412 Precondition Failed
```

**Testing considerations:**
- Verify it's a FULL replacement (omitted optional fields should reset to defaults)
- Test idempotency (same PUT twice gives same result)
- Test with If-Match for optimistic concurrency
- Test on non-existent resource (404 or 201 depending on API design)
- Test that read-only fields (id, createdAt) are ignored in the request body
- Compare PUT vs PATCH behavior — PUT replaces, PATCH merges

---

### PATCH

**Purpose:** Partially modify a resource.

**Properties:**
- Safe: No
- Idempotent: Not necessarily (depends on the patch format)
- Cacheable: No
- Request body: Expected (partial representation or patch document)

**Semantics:**
- Modifies only the fields included in the request body
- Fields not included are left unchanged
- Two formats: JSON Merge Patch (RFC 7396) and JSON Patch (RFC 6902)
- May not be idempotent (e.g., "increment counter by 1")

**Common patterns:**

```http
# JSON Merge Patch (most common in REST APIs)
PATCH /api/users/123 HTTP/1.1
Content-Type: application/merge-patch+json
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...

{
  "name": "Jane Smith",
  "preferences": {
    "theme": "dark"
  }
}

# Response — only name and preferences.theme changed
HTTP/1.1 200 OK
Content-Type: application/json

{
  "id": "123",
  "email": "jane@example.com",
  "name": "Jane Smith",
  "preferences": {
    "theme": "dark",
    "language": "en"
  }
}
```

```http
# JSON Patch (RFC 6902) — more powerful, more complex
PATCH /api/products/456 HTTP/1.1
Content-Type: application/json-patch+json
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...

[
  { "op": "replace", "path": "/price", "value": 39.99 },
  { "op": "add", "path": "/tags/-", "value": "sale" },
  { "op": "remove", "path": "/tags/0" },
  { "op": "test", "path": "/stock", "value": 100 }
]

# Response
HTTP/1.1 200 OK
```

**Testing considerations:**
- Verify only specified fields are modified
- Test that unspecified fields remain unchanged
- Test setting a field to null (vs omitting it)
- Test nested partial updates
- Verify that read-only fields cannot be modified via PATCH
- Test with empty object {} (should succeed with no changes, or 400)

---

### DELETE

**Purpose:** Remove a resource.

**Properties:**
- Safe: No
- Idempotent: Yes (deleting an already-deleted resource is the same as deleting it once)
- Cacheable: No
- Request body: Usually not expected

**Semantics:**
- Removes the resource at the given URI
- The resource may be soft-deleted or hard-deleted (implementation dependent)
- Subsequent GET requests should return 404 (or 410 Gone)
- Second DELETE of same resource should return 404 (already deleted)

**Common patterns:**

```http
# Delete a resource
DELETE /api/products/456 HTTP/1.1
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...

# Response (no body)
HTTP/1.1 204 No Content
```

```http
# Delete with confirmation body (some APIs)
DELETE /api/users/123 HTTP/1.1
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...

# Response with deleted resource (some APIs return the deleted resource)
HTTP/1.1 200 OK
Content-Type: application/json

{
  "id": "123",
  "email": "jane@example.com",
  "deletedAt": "2025-03-15T10:00:00Z"
}
```

```http
# Bulk delete (some APIs support this)
DELETE /api/products HTTP/1.1
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...

{
  "ids": ["456", "789", "101"]
}

# Response
HTTP/1.1 200 OK
Content-Type: application/json

{
  "deleted": 3,
  "failed": 0
}
```

**Testing considerations:**
- Verify the resource is no longer accessible via GET (404)
- Test deleting a non-existent resource (404)
- Test deleting an already-deleted resource (404 — idempotency check)
- Test cascade behavior (does deleting a user delete their orders?)
- Test authorization (can a non-admin delete this?)
- Test soft delete vs hard delete behavior
- Verify related resources are handled correctly

---

### HEAD

**Purpose:** Same as GET but returns only headers, no body.

**Properties:**
- Safe: Yes
- Idempotent: Yes
- Cacheable: Yes
- Request body: Not expected

**Semantics:**
- Identical to GET but the server MUST NOT return a message body
- Response headers MUST be identical to what GET would return
- Useful for checking if a resource exists without downloading it
- Useful for checking Content-Length before downloading large files

**Common patterns:**

```http
# Check if a resource exists
HEAD /api/products/456 HTTP/1.1
Host: api.example.com

# Response (exists)
HTTP/1.1 200 OK
Content-Type: application/json
Content-Length: 1234
ETag: "abc123"
Last-Modified: Sat, 15 Mar 2025 10:00:00 GMT

# Response (doesn't exist)
HTTP/1.1 404 Not Found
```

```http
# Check file size before download
HEAD /api/files/report.pdf HTTP/1.1
Host: api.example.com
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...

# Response
HTTP/1.1 200 OK
Content-Type: application/pdf
Content-Length: 5242880
Accept-Ranges: bytes
```

**Testing considerations:**
- Verify no body is returned
- Verify status code matches what GET would return
- Verify Content-Length is accurate
- Verify ETag and other metadata headers match GET

---

### OPTIONS

**Purpose:** Describe the communication options for a resource.

**Properties:**
- Safe: Yes
- Idempotent: Yes
- Cacheable: No
- Request body: Usually not expected

**Semantics:**
- Returns the allowed HTTP methods for the resource
- Primary use in modern web: CORS preflight requests
- Can be used for feature discovery (WebDAV, etc.)

**Common patterns:**

```http
# Simple OPTIONS request
OPTIONS /api/products HTTP/1.1
Host: api.example.com

# Response
HTTP/1.1 204 No Content
Allow: GET, POST, HEAD, OPTIONS
```

```http
# CORS preflight request (browser sends this automatically)
OPTIONS /api/users HTTP/1.1
Host: api.example.com
Origin: https://app.example.com
Access-Control-Request-Method: POST
Access-Control-Request-Headers: Content-Type, Authorization

# CORS preflight response
HTTP/1.1 204 No Content
Access-Control-Allow-Origin: https://app.example.com
Access-Control-Allow-Methods: GET, POST, PUT, DELETE, PATCH
Access-Control-Allow-Headers: Content-Type, Authorization
Access-Control-Max-Age: 86400
```

**Testing considerations:**
- Verify Allow header lists correct methods
- Test CORS preflight for your frontend's origin
- Verify CORS headers are correct (not wildcard * with credentials)
- Test with disallowed methods (should be listed/unlisted correctly)

---

### TRACE

**Purpose:** Echoes the received request for debugging (rarely used in APIs).

**Properties:**
- Safe: Yes
- Idempotent: Yes
- Cacheable: No
- Request body: Not expected

**Semantics:**
- Returns the request as received by the server (loop-back diagnostic)
- SHOULD be disabled in production (security risk — XST attacks)
- Useful for debugging proxy chains

**Security note:** TRACE should be disabled on production APIs. If enabled, it can be
exploited for Cross-Site Tracing (XST) attacks to steal credentials.

**Testing considerations:**
- Verify TRACE is disabled (should return 405 Method Not Allowed)
- If it must be enabled, ensure sensitive headers (Authorization, Cookie) are stripped

---

## Method Properties

### Quick Reference Table

| Method  | Safe | Idempotent | Cacheable | Request Body | Response Body |
|---------|------|------------|-----------|-------------|---------------|
| GET     | Yes  | Yes        | Yes       | No          | Yes           |
| HEAD    | Yes  | Yes        | Yes       | No          | No            |
| POST    | No   | No         | Rarely    | Yes         | Yes           |
| PUT     | No   | Yes        | No        | Yes         | Yes           |
| PATCH   | No   | No*        | No        | Yes         | Yes           |
| DELETE  | No   | Yes        | No        | Rarely      | Optional      |
| OPTIONS | Yes  | Yes        | No        | Rarely      | Yes           |
| TRACE   | Yes  | Yes        | No        | No          | Yes           |

*PATCH can be idempotent depending on the patch format (JSON Merge Patch is, JSON Patch may not be)

### Safe Methods

A method is **safe** if it does not modify server state. Clients can send safe requests
without worrying about unintended side effects (prefetching, caching, retrying).

Safe methods: GET, HEAD, OPTIONS, TRACE

Important: "Safe" doesn't mean "no side effects at all" — a server may log the request,
update analytics counters, etc. It means the client doesn't request or expect state change.

### Idempotent Methods

A method is **idempotent** if making the same request multiple times has the same effect
as making it once. The server state after N identical requests is the same as after 1 request.

Idempotent methods: GET, HEAD, PUT, DELETE, OPTIONS, TRACE

Important for testing:
- PUT /users/123 with the same body N times → same result
- DELETE /users/123 N times → first succeeds, rest return 404 (but server state is same)
- POST /users N times → may create N different users (NOT idempotent)

### Cacheable Methods

A method's response is **cacheable** if it can be stored and reused for subsequent requests.

Cacheable by default: GET, HEAD
Cacheable with explicit headers: POST (rarely)
Not cacheable: PUT, PATCH, DELETE, OPTIONS, TRACE

---

## Status Codes

### 1xx Informational

These indicate the server has received the request and the client should continue.

| Code | Name | When to Use |
|------|------|-------------|
| 100 | Continue | Server has received request headers, client should send body |
| 101 | Switching Protocols | Server is switching to a different protocol (e.g., WebSocket upgrade) |
| 102 | Processing | Server is processing the request (WebDAV; prevents client timeout) |
| 103 | Early Hints | Preload resources before the final response (Link headers) |

**WebSocket upgrade example:**

```http
# Client request
GET /ws HTTP/1.1
Host: api.example.com
Upgrade: websocket
Connection: Upgrade
Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==
Sec-WebSocket-Version: 13

# Server response
HTTP/1.1 101 Switching Protocols
Upgrade: websocket
Connection: Upgrade
Sec-WebSocket-Accept: s3pPLMBiTxaQ9kYGzzhZRbK+xOo=
```

**Early Hints example:**

```http
HTTP/1.1 103 Early Hints
Link: </style.css>; rel=preload; as=style
Link: </script.js>; rel=preload; as=script

# ... then the final response follows
HTTP/1.1 200 OK
Content-Type: text/html
```

---

### 2xx Success

The request was successfully received, understood, and accepted.

| Code | Name | When to Use |
|------|------|-------------|
| 200 | OK | Successful GET, PUT, PATCH, or POST (action) |
| 201 | Created | Resource created via POST (include Location header) |
| 202 | Accepted | Request accepted for async processing (not yet complete) |
| 203 | Non-Authoritative | Response from a transforming proxy, not the origin |
| 204 | No Content | Successful DELETE or PUT with no response body |
| 205 | Reset Content | Success — client should reset the document view |
| 206 | Partial Content | Range request fulfilled (Content-Range header) |
| 207 | Multi-Status | Multiple resources with different status codes (WebDAV) |
| 208 | Already Reported | Members already enumerated (WebDAV) |
| 226 | IM Used | Server fulfilled a GET with instance-manipulations |

**200 OK — The workhorse:**

```http
# Successful GET
HTTP/1.1 200 OK
Content-Type: application/json

{"id": "123", "name": "Jane Doe"}

# Successful PUT/PATCH
HTTP/1.1 200 OK
Content-Type: application/json

{"id": "123", "name": "Jane Smith", "updatedAt": "2025-03-15T10:00:00Z"}
```

**201 Created — Resource creation:**

```http
HTTP/1.1 201 Created
Content-Type: application/json
Location: /api/users/789

{"id": "789", "email": "john@example.com", "name": "John Doe"}
```

**202 Accepted — Async processing:**

```http
# Start a long-running job
POST /api/reports/generate HTTP/1.1
Content-Type: application/json

{"type": "annual", "year": 2024}

# Response
HTTP/1.1 202 Accepted
Content-Type: application/json
Location: /api/reports/jobs/abc123

{
  "jobId": "abc123",
  "status": "queued",
  "estimatedCompletion": "2025-03-15T10:05:00Z",
  "statusUrl": "/api/reports/jobs/abc123"
}
```

**204 No Content — Successful with no body:**

```http
# Successful DELETE
HTTP/1.1 204 No Content

# Successful PUT with no response needed
HTTP/1.1 204 No Content
```

**206 Partial Content — Range requests:**

```http
# Request a range of bytes
GET /api/files/large-report.csv HTTP/1.1
Range: bytes=0-999

# Response
HTTP/1.1 206 Partial Content
Content-Type: text/csv
Content-Range: bytes 0-999/50000
Content-Length: 1000

[first 1000 bytes of the file]
```

---

### 3xx Redirection

The resource has moved; the client should follow the new location.

| Code | Name | When to Use |
|------|------|-------------|
| 300 | Multiple Choices | Multiple representations available (rare) |
| 301 | Moved Permanently | Resource permanently at new URI (cacheable) |
| 302 | Found | Resource temporarily at different URI (legacy; use 303 or 307) |
| 303 | See Other | Response to POST; redirect to GET another resource |
| 304 | Not Modified | Conditional GET — resource unchanged (no body) |
| 307 | Temporary Redirect | Like 302 but MUST NOT change method |
| 308 | Permanent Redirect | Like 301 but MUST NOT change method |

**301 Moved Permanently:**

```http
# Old URL
GET /api/v1/users HTTP/1.1

# Response
HTTP/1.1 301 Moved Permanently
Location: /api/v2/users
```

**304 Not Modified:**

```http
# Conditional request
GET /api/products/456 HTTP/1.1
If-None-Match: "etag-abc123"

# Response (resource unchanged — use cached version)
HTTP/1.1 304 Not Modified
ETag: "etag-abc123"
Cache-Control: private, max-age=60
```

**307 vs 308 — Preserving method:**

```http
# 307: Temporary redirect, KEEP the original method (POST stays POST)
HTTP/1.1 307 Temporary Redirect
Location: /api/v2/orders

# 308: Permanent redirect, KEEP the original method (POST stays POST)
HTTP/1.1 308 Permanent Redirect
Location: /api/v2/orders

# Compare with 301/302 which may change POST to GET (per old browser behavior)
```

---

### 4xx Client Errors

The client sent a request the server cannot process.

| Code | Name | When to Use |
|------|------|-------------|
| 400 | Bad Request | Malformed request syntax, invalid parameters |
| 401 | Unauthorized | Authentication required or failed |
| 402 | Payment Required | Reserved for future use (some APIs use it) |
| 403 | Forbidden | Authenticated but lacks permission |
| 404 | Not Found | Resource does not exist |
| 405 | Method Not Allowed | HTTP method not supported for this endpoint |
| 406 | Not Acceptable | Cannot produce response matching Accept header |
| 407 | Proxy Authentication Required | Proxy requires authentication |
| 408 | Request Timeout | Server tired of waiting for the client |
| 409 | Conflict | Request conflicts with current state (duplicate, concurrent) |
| 410 | Gone | Resource permanently deleted (unlike 404 — it existed before) |
| 411 | Length Required | Content-Length header is required |
| 412 | Precondition Failed | If-Match, If-Unmodified-Since condition not met |
| 413 | Payload Too Large | Request body exceeds server's size limit |
| 414 | URI Too Long | Request URI exceeds server's length limit |
| 415 | Unsupported Media Type | Content-Type not supported |
| 416 | Range Not Satisfiable | Range header value is out of bounds |
| 417 | Expectation Failed | Expect header cannot be met |
| 418 | I'm a teapot | RFC 2324 joke; SHOULD NOT be used seriously |
| 421 | Misdirected Request | Request directed at wrong server (HTTP/2) |
| 422 | Unprocessable Entity | Valid syntax but semantic errors (WebDAV, widely used) |
| 423 | Locked | Resource is locked (WebDAV; used for account lockout) |
| 424 | Failed Dependency | Dependent request failed (WebDAV) |
| 425 | Too Early | Server won't process request that might be replayed (TLS) |
| 426 | Upgrade Required | Client must switch to a different protocol |
| 428 | Precondition Required | Server requires conditional request headers |
| 429 | Too Many Requests | Rate limit exceeded |
| 431 | Request Header Fields Too Large | Headers exceed size limit |
| 451 | Unavailable For Legal Reasons | Blocked by legal demand (censorship, DMCA) |

**400 Bad Request — Validation errors:**

```http
POST /api/users HTTP/1.1
Content-Type: application/json

{"email": "not-an-email", "name": ""}

# Response
HTTP/1.1 400 Bad Request
Content-Type: application/json

{
  "error": "Validation Error",
  "message": "Request body contains invalid fields",
  "statusCode": 400,
  "details": [
    {"field": "email", "message": "Invalid email format", "code": "invalid_format"},
    {"field": "name", "message": "Name is required", "code": "required"}
  ]
}
```

**401 Unauthorized — Authentication failed:**

```http
GET /api/users/me HTTP/1.1
# No Authorization header

# Response
HTTP/1.1 401 Unauthorized
WWW-Authenticate: Bearer realm="api.example.com"
Content-Type: application/json

{
  "error": "Unauthorized",
  "message": "Authentication required. Include a valid Bearer token.",
  "statusCode": 401
}
```

**403 Forbidden — Insufficient permissions:**

```http
DELETE /api/users/123 HTTP/1.1
Authorization: Bearer user-token-not-admin

# Response
HTTP/1.1 403 Forbidden
Content-Type: application/json

{
  "error": "Forbidden",
  "message": "Admin role required to delete users",
  "statusCode": 403
}
```

**404 Not Found:**

```http
GET /api/users/non-existent-id HTTP/1.1
Authorization: Bearer valid-token

# Response
HTTP/1.1 404 Not Found
Content-Type: application/json

{
  "error": "Not Found",
  "message": "User not found",
  "statusCode": 404
}
```

**405 Method Not Allowed:**

```http
PATCH /api/auth/login HTTP/1.1

# Response
HTTP/1.1 405 Method Not Allowed
Allow: POST
Content-Type: application/json

{
  "error": "Method Not Allowed",
  "message": "PATCH is not supported for /api/auth/login. Allowed methods: POST",
  "statusCode": 405
}
```

**409 Conflict — Duplicate resource:**

```http
POST /api/users HTTP/1.1
Content-Type: application/json

{"email": "existing@example.com", "name": "Duplicate"}

# Response
HTTP/1.1 409 Conflict
Content-Type: application/json

{
  "error": "Conflict",
  "message": "A user with email existing@example.com already exists",
  "statusCode": 409
}
```

**412 Precondition Failed — Optimistic concurrency:**

```http
PUT /api/products/456 HTTP/1.1
If-Match: "stale-etag"
Content-Type: application/json

{"name": "Updated", "price": 39.99}

# Response (someone else modified the resource)
HTTP/1.1 412 Precondition Failed
Content-Type: application/json

{
  "error": "Precondition Failed",
  "message": "Resource was modified by another request. Refresh and try again.",
  "statusCode": 412,
  "currentETag": "new-etag"
}
```

**415 Unsupported Media Type:**

```http
POST /api/users HTTP/1.1
Content-Type: text/xml

<user><email>john@example.com</email></user>

# Response
HTTP/1.1 415 Unsupported Media Type
Content-Type: application/json

{
  "error": "Unsupported Media Type",
  "message": "Content-Type must be application/json",
  "statusCode": 415
}
```

**422 Unprocessable Entity — Semantic errors:**

```http
POST /api/orders HTTP/1.1
Content-Type: application/json

{
  "productId": "prod-123",
  "quantity": 100
}

# Response (valid JSON, but product only has 5 in stock)
HTTP/1.1 422 Unprocessable Entity
Content-Type: application/json

{
  "error": "Unprocessable Entity",
  "message": "Insufficient stock. Product has 5 available, 100 requested.",
  "statusCode": 422
}
```

**429 Too Many Requests — Rate limited:**

```http
HTTP/1.1 429 Too Many Requests
Content-Type: application/json
Retry-After: 60
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1710504060

{
  "error": "Too Many Requests",
  "message": "Rate limit exceeded. Try again in 60 seconds.",
  "statusCode": 429,
  "retryAfter": 60
}
```

---

### 5xx Server Errors

The server failed to fulfill a valid request.

| Code | Name | When to Use |
|------|------|-------------|
| 500 | Internal Server Error | Generic server error (unhandled exception) |
| 501 | Not Implemented | Server doesn't recognize the request method |
| 502 | Bad Gateway | Upstream server returned an invalid response |
| 503 | Service Unavailable | Server temporarily overloaded or down for maintenance |
| 504 | Gateway Timeout | Upstream server didn't respond in time |
| 505 | HTTP Version Not Supported | HTTP version not supported |
| 506 | Variant Also Negotiates | Content negotiation configuration error |
| 507 | Insufficient Storage | Cannot store the representation (WebDAV) |
| 508 | Loop Detected | Infinite loop in server processing (WebDAV) |
| 510 | Not Extended | Further extensions needed to fulfill request |
| 511 | Network Authentication Required | Network-level authentication needed (captive portal) |

**500 Internal Server Error:**

```http
HTTP/1.1 500 Internal Server Error
Content-Type: application/json

{
  "error": "Internal Server Error",
  "message": "An unexpected error occurred. Please try again later.",
  "statusCode": 500,
  "requestId": "req-abc123"
}
```

IMPORTANT: Never expose stack traces, SQL queries, file paths, or internal details in 500 responses.
Include a request ID so developers can reference it when contacting support.

**502 Bad Gateway:**

```http
HTTP/1.1 502 Bad Gateway
Content-Type: application/json

{
  "error": "Bad Gateway",
  "message": "The payment service returned an unexpected response",
  "statusCode": 502
}
```

**503 Service Unavailable:**

```http
HTTP/1.1 503 Service Unavailable
Retry-After: 300
Content-Type: application/json

{
  "error": "Service Unavailable",
  "message": "Service is temporarily unavailable due to maintenance. Expected recovery: 10:30 UTC.",
  "statusCode": 503,
  "retryAfter": 300
}
```

---

## Headers

### Request Headers

| Header | Purpose | Example |
|--------|---------|---------|
| Accept | Preferred response media types | `application/json` |
| Accept-Encoding | Compression algorithms supported | `gzip, deflate, br` |
| Accept-Language | Preferred response language | `en-US,en;q=0.9` |
| Authorization | Authentication credentials | `Bearer eyJhbGci...` |
| Content-Type | Media type of request body | `application/json` |
| Content-Length | Size of request body in bytes | `256` |
| Cookie | Session cookies | `session_id=abc123` |
| Host | Target host and port | `api.example.com` |
| If-Match | Conditional (must match ETag) | `"etag-abc123"` |
| If-None-Match | Conditional (must NOT match ETag) | `"etag-abc123"` |
| If-Modified-Since | Conditional (modified after date) | `Sat, 15 Mar 2025 10:00:00 GMT` |
| If-Unmodified-Since | Conditional (not modified after date) | `Sat, 15 Mar 2025 10:00:00 GMT` |
| Origin | Request origin (for CORS) | `https://app.example.com` |
| Range | Request partial content | `bytes=0-999` |
| Referer | Previous page URL | `https://app.example.com/products` |
| User-Agent | Client identification | `Mozilla/5.0...` |
| X-Request-ID | Client-generated request tracking ID | `req-abc-123` |
| Idempotency-Key | Prevent duplicate operations | `idem-abc-123` |

### Response Headers

| Header | Purpose | Example |
|--------|---------|---------|
| Content-Type | Media type of response body | `application/json; charset=utf-8` |
| Content-Length | Size of response body | `1024` |
| Content-Encoding | Compression applied | `gzip` |
| ETag | Resource version identifier | `"abc123"` |
| Last-Modified | When resource was last changed | `Sat, 15 Mar 2025 10:00:00 GMT` |
| Location | URI of created or redirected resource | `/api/users/789` |
| Allow | Supported HTTP methods | `GET, POST, HEAD, OPTIONS` |
| Retry-After | When to retry (429/503) | `60` |
| Link | Pagination and related resources | `</page2>; rel="next"` |
| X-Request-ID | Server request tracking ID | `req-abc-123` |
| X-RateLimit-Limit | Max requests in window | `100` |
| X-RateLimit-Remaining | Remaining requests | `95` |
| X-RateLimit-Reset | When limit resets (Unix timestamp) | `1710504060` |

### Content Negotiation

```http
# Client prefers JSON, will accept XML
GET /api/products HTTP/1.1
Accept: application/json;q=1.0, application/xml;q=0.5

# Server responds with JSON (higher quality)
HTTP/1.1 200 OK
Content-Type: application/json
Vary: Accept

# Quality values (q-values) range from 0 to 1:
# q=1.0 — highest preference (default if omitted)
# q=0.5 — acceptable but not preferred
# q=0   — not acceptable
```

### Caching Headers

```http
# Cache-Control directive reference
Cache-Control: public                  # CDN and browser can cache
Cache-Control: private                 # Only browser can cache (user-specific data)
Cache-Control: no-cache                # Must revalidate before using cached copy
Cache-Control: no-store                # Don't cache at all (sensitive data)
Cache-Control: max-age=3600            # Fresh for 1 hour
Cache-Control: s-maxage=86400          # CDN can cache for 24 hours
Cache-Control: must-revalidate         # Must revalidate when stale
Cache-Control: immutable               # Never changes (use for versioned assets)
Cache-Control: stale-while-revalidate=60  # Serve stale while revalidating
Cache-Control: stale-if-error=3600     # Serve stale if origin is down

# Common combinations for APIs
Cache-Control: public, max-age=60, s-maxage=300          # Public data, short cache
Cache-Control: private, no-cache                          # User-specific, always revalidate
Cache-Control: no-store                                   # Never cache (auth tokens, PII)
Cache-Control: public, max-age=31536000, immutable        # Versioned static asset
```

### Security Headers

```http
# Essential security headers for APIs
X-Content-Type-Options: nosniff          # Prevent MIME type sniffing
X-Frame-Options: DENY                   # Prevent clickjacking
Strict-Transport-Security: max-age=31536000; includeSubDomains  # Force HTTPS
Content-Security-Policy: default-src 'none'                     # No embedded content
X-XSS-Protection: 0                     # Disable (CSP is better, this can introduce issues)
Referrer-Policy: strict-origin-when-cross-origin                # Control referer leakage
Permissions-Policy: camera=(), microphone=(), geolocation=()    # Restrict browser features

# Headers to REMOVE
# X-Powered-By: Express    ← Remove this! Reveals server technology
# Server: Apache/2.4.51    ← Minimize this! Don't reveal version
```

### CORS Headers

```http
# CORS response headers
Access-Control-Allow-Origin: https://app.example.com    # Specific origin (preferred)
Access-Control-Allow-Origin: *                          # Any origin (NO credentials!)
Access-Control-Allow-Methods: GET, POST, PUT, DELETE, PATCH
Access-Control-Allow-Headers: Content-Type, Authorization, X-Request-ID
Access-Control-Expose-Headers: X-RateLimit-Limit, X-RateLimit-Remaining, X-Total-Count
Access-Control-Allow-Credentials: true                  # Allow cookies/auth headers
Access-Control-Max-Age: 86400                          # Cache preflight for 24 hours

# IMPORTANT: Access-Control-Allow-Origin: * and Access-Control-Allow-Credentials: true
# are MUTUALLY EXCLUSIVE. Browsers reject this combination.
# If you need credentials, specify the exact origin.
```

---

## HTTP/2 Considerations

### Key Differences from HTTP/1.1

| Feature | HTTP/1.1 | HTTP/2 |
|---------|----------|--------|
| Multiplexing | One request per connection (or pipelining) | Multiple concurrent streams on one connection |
| Header compression | None | HPACK compression |
| Server push | Not supported | Server can push resources proactively |
| Binary protocol | Text-based | Binary framing |
| Connection reuse | Keep-Alive (limited) | Single connection, unlimited streams |
| Priority | None | Stream prioritization |

### Impact on API Testing

1. **Multiplexing** — Multiple requests fly over a single TCP connection simultaneously.
   Load tests may see different connection patterns than HTTP/1.1.

2. **Header compression** — HPACK compresses repetitive headers. Large Authorization
   headers are efficiently compressed on subsequent requests.

3. **No head-of-line blocking** — One slow response doesn't block others.
   Performance tests should measure per-stream latency.

4. **Flow control** — Per-stream and per-connection flow control prevents any single
   stream from monopolizing bandwidth.

### Testing HTTP/2 Specifically

```bash
# Test with curl (HTTP/2)
curl --http2 -v https://api.example.com/products

# Force HTTP/2 over cleartext (h2c)
curl --http2-prior-knowledge http://localhost:3000/api/products

# k6 supports HTTP/2 by default when connecting to HTTPS
k6 run load-test.js  # Uses HTTP/2 for HTTPS targets automatically
```

---

## HTTP/3 Considerations

### Key Differences from HTTP/2

| Feature | HTTP/2 | HTTP/3 |
|---------|--------|--------|
| Transport | TCP | QUIC (UDP-based) |
| TLS | TLS 1.2+ (separate) | TLS 1.3 (built-in) |
| Head-of-line blocking | Connection-level (TCP) | None (per-stream) |
| Connection establishment | TCP + TLS handshake | 0-RTT or 1-RTT |
| Connection migration | Re-establish on IP change | Seamless (connection ID) |

### Impact on API Testing

1. **Faster connection setup** — 0-RTT reduces latency for repeat connections.
   Performance tests may see lower TTFB on QUIC/HTTP/3.

2. **No TCP head-of-line blocking** — Packet loss on one stream doesn't affect others.
   Under lossy conditions, HTTP/3 performs significantly better.

3. **Connection migration** — Mobile clients switching from WiFi to cellular maintain
   the connection. Important for mobile API testing.

### Testing HTTP/3

```bash
# Test with curl (HTTP/3)
curl --http3 -v https://api.example.com/products

# Check if server supports HTTP/3
curl -I https://api.example.com/ | grep -i alt-svc
# Response: alt-svc: h3=":443"; ma=86400
```

---

## Testing Patterns

### Method-Based Test Matrix

For each endpoint, test these scenarios:

| Scenario | Methods | Expected |
|----------|---------|----------|
| Happy path | All | 2xx success |
| Wrong method | All wrong ones | 405 Method Not Allowed |
| No auth | Authenticated endpoints | 401 Unauthorized |
| Wrong role | Role-restricted | 403 Forbidden |
| Not found | GET, PUT, PATCH, DELETE | 404 Not Found |
| Invalid input | POST, PUT, PATCH | 400 Bad Request |
| Duplicate | POST (create) | 409 Conflict |
| Rate limited | All (high volume) | 429 Too Many Requests |
| Server error | All (trigger error) | 500 Internal Server Error |

### Status Code Decision Tree

```
Is the request valid?
├── No → Is the syntax valid?
│   ├── No → 400 Bad Request
│   └── Yes → Is the client authenticated?
│       ├── No → 401 Unauthorized
│       └── Yes → Is the client authorized?
│           ├── No → 403 Forbidden
│           └── Yes → Does the resource exist?
│               ├── No → 404 Not Found
│               └── Yes → Is there a conflict?
│                   ├── Yes → 409 Conflict
│                   └── No → Are semantics valid?
│                       ├── No → 422 Unprocessable Entity
│                       └── Yes → (should have been valid — check again)
└── Yes → Can the server fulfill it?
    ├── No → 500/502/503/504
    └── Yes → What was the action?
        ├── Read → 200 OK (or 304 Not Modified)
        ├── Create → 201 Created (or 202 Accepted)
        ├── Update → 200 OK (or 204 No Content)
        └── Delete → 204 No Content (or 200 OK)
```

### Comprehensive Header Testing Checklist

```
For every endpoint, verify:

Security Headers:
  □ X-Content-Type-Options: nosniff
  □ X-Frame-Options: DENY or SAMEORIGIN
  □ No X-Powered-By header
  □ No Server version disclosure
  □ Content-Type matches actual content

Caching Headers:
  □ GET endpoints have appropriate Cache-Control
  □ Mutation responses have no-store or no-cache
  □ ETag present for cacheable resources
  □ Vary header set when response varies by header

CORS Headers (if applicable):
  □ Access-Control-Allow-Origin is specific (not *)
  □ Credentials flag consistent with origin
  □ Exposed headers include rate limit headers
  □ Preflight max-age is reasonable

Rate Limit Headers (if rate limiting is enabled):
  □ X-RateLimit-Limit present
  □ X-RateLimit-Remaining decrements
  □ X-RateLimit-Reset is valid timestamp
  □ Retry-After present on 429 responses

Content Headers:
  □ Content-Type includes charset (utf-8)
  □ Content-Length matches body size
  □ Content-Encoding present when compressed
```

## Request/Response Lifecycle

### Complete HTTP Request Anatomy

```
┌─────────────────────────────────────────────────────────────┐
│                    HTTP REQUEST                              │
├─────────────────────────────────────────────────────────────┤
│ Request Line:                                               │
│   POST /api/users HTTP/1.1                                  │
│                                                             │
│ Headers:                                                    │
│   Host: api.example.com                                     │
│   Authorization: Bearer eyJhbGciOiJIUzI1NiIs...            │
│   Content-Type: application/json                            │
│   Content-Length: 89                                         │
│   Accept: application/json                                  │
│   Accept-Encoding: gzip, br                                 │
│   User-Agent: MyApp/1.0                                     │
│   X-Request-ID: req-abc-123                                 │
│   X-Idempotency-Key: idem-xyz-456                          │
│                                                             │
│ Empty Line (separates headers from body)                    │
│                                                             │
│ Body:                                                       │
│   {"email":"jane@example.com","name":"Jane","role":"USER"}  │
└─────────────────────────────────────────────────────────────┘

                           │
                           ▼

┌─────────────────────────────────────────────────────────────┐
│                    HTTP RESPONSE                             │
├─────────────────────────────────────────────────────────────┤
│ Status Line:                                                │
│   HTTP/1.1 201 Created                                      │
│                                                             │
│ Headers:                                                    │
│   Content-Type: application/json; charset=utf-8             │
│   Content-Length: 195                                        │
│   Location: /api/users/789                                  │
│   X-Request-ID: req-abc-123                                 │
│   X-RateLimit-Limit: 100                                    │
│   X-RateLimit-Remaining: 99                                 │
│   X-Content-Type-Options: nosniff                           │
│   Cache-Control: no-store                                   │
│   Date: Sat, 15 Mar 2025 10:30:00 GMT                      │
│                                                             │
│ Empty Line                                                  │
│                                                             │
│ Body:                                                       │
│   {"id":"789","email":"jane@example.com","name":"Jane",...}  │
└─────────────────────────────────────────────────────────────┘
```

### Timing Breakdown

```
DNS Lookup    ──────▶  The time to resolve the domain to an IP address
                       Typical: 0-50ms (cached: 0ms)
                       Test: First request may be slower

TCP Connect   ──────▶  Three-way handshake to establish connection
                       Typical: 1-50ms (local: <1ms)
                       HTTP/2: Reuses connection (0ms after first)

TLS Handshake ──────▶  Negotiate encryption (HTTPS only)
                       Typical: 10-100ms
                       HTTP/2: Session resumption reduces this
                       HTTP/3: 0-RTT eliminates this on repeat connections

Time to First ──────▶  Server processing time (most important metric!)
Byte (TTFB)            This is where your API code runs:
                         - Authentication check
                         - Database queries
                         - Business logic
                         - Response serialization
                       Typical: 10-500ms depending on complexity

Content       ──────▶  Time to transfer the response body
Transfer                Typical: 1-50ms for JSON API responses
                       Large responses or slow networks increase this

Total         = DNS + TCP + TLS + TTFB + Content Transfer
```

## URL Structure for APIs

### RESTful URL Design

```
Base URL:
  https://api.example.com/v1

Resource Collections:
  GET    /users                   List users
  POST   /users                   Create user

Individual Resources:
  GET    /users/123               Get user 123
  PUT    /users/123               Replace user 123
  PATCH  /users/123               Partially update user 123
  DELETE /users/123               Delete user 123

Nested Resources:
  GET    /users/123/orders        List user 123's orders
  POST   /users/123/orders        Create order for user 123
  GET    /users/123/orders/456    Get specific order

Sub-Resources:
  GET    /users/123/avatar        Get user's avatar
  PUT    /users/123/avatar        Replace avatar
  DELETE /users/123/avatar        Remove avatar

Actions (non-CRUD):
  POST   /users/123/verify-email  Trigger email verification
  POST   /users/123/reset-password Request password reset
  POST   /orders/456/cancel       Cancel an order
  POST   /payments/789/refund     Refund a payment

Search and Filtering:
  GET    /products?category=electronics&minPrice=20&maxPrice=100
  GET    /products?search=wireless+headphones
  GET    /products?sort=price&order=asc&page=2&pageSize=20

Versioning:
  /v1/products                    Version 1
  /v2/products                    Version 2
```

### Query Parameter Conventions

```
Pagination:
  ?page=2&pageSize=20             Offset-based pagination
  ?cursor=abc123&limit=20         Cursor-based pagination
  ?offset=40&limit=20             Offset/limit pagination

Filtering:
  ?category=electronics           Exact match
  ?minPrice=10&maxPrice=100       Range filter
  ?status=active,pending          Multiple values (comma-separated)
  ?createdAfter=2025-01-01        Date filter
  ?search=keyword                 Full-text search

Sorting:
  ?sort=price&order=asc           Single field sort
  ?sort=price,-name               Multi-field sort (- for descending)
  ?sortBy=price:asc,name:desc     Alternative format

Field Selection (sparse fieldsets):
  ?fields=id,name,price           Only return specified fields
  ?include=reviews,category       Include nested/related resources
  ?exclude=description            Exclude specific fields

Expansion:
  ?expand=author,comments         Expand references inline
  ?depth=2                        Control nesting depth
```

## Content Types for APIs

### Common Content Types

```
application/json                    Standard JSON (most common for APIs)
  Content-Type: application/json; charset=utf-8

application/json-patch+json         JSON Patch (RFC 6902)
  Content-Type: application/json-patch+json

application/merge-patch+json        JSON Merge Patch (RFC 7396)
  Content-Type: application/merge-patch+json

application/x-www-form-urlencoded   Form data (used by OAuth token endpoint)
  Content-Type: application/x-www-form-urlencoded

multipart/form-data                 File uploads
  Content-Type: multipart/form-data; boundary=----FormBoundary

application/octet-stream            Binary data (file download)
  Content-Type: application/octet-stream

text/event-stream                   Server-Sent Events (SSE)
  Content-Type: text/event-stream

application/graphql                 GraphQL queries (alternative to JSON)
  Content-Type: application/graphql

application/xml                     XML (legacy APIs)
  Content-Type: application/xml

application/problem+json            RFC 7807 Problem Details
  Content-Type: application/problem+json
```

### RFC 7807 Problem Details

A standardized error format recommended for REST APIs:

```http
HTTP/1.1 403 Forbidden
Content-Type: application/problem+json

{
  "type": "https://api.example.com/errors/insufficient-funds",
  "title": "Insufficient Funds",
  "status": 403,
  "detail": "Your account balance of $10.00 is insufficient for this $25.00 purchase.",
  "instance": "/api/orders/req-abc-123",
  "balance": 10.00,
  "required": 25.00
}
```

| Field | Required | Description |
|-------|----------|-------------|
| `type` | Yes | URI identifying the error type (can be a documentation link) |
| `title` | Yes | Short, human-readable summary |
| `status` | Yes | HTTP status code |
| `detail` | No | Human-readable explanation specific to this occurrence |
| `instance` | No | URI identifying the specific occurrence |
| Custom fields | No | Additional machine-readable details |

## Compression

### Request Compression

```http
# Client sends compressed request body
POST /api/data HTTP/1.1
Content-Type: application/json
Content-Encoding: gzip
Content-Length: 234

[gzip-compressed JSON body]
```

### Response Compression

```http
# Client indicates it accepts compressed responses
GET /api/products HTTP/1.1
Accept-Encoding: gzip, br, zstd

# Server responds with compressed body
HTTP/1.1 200 OK
Content-Type: application/json
Content-Encoding: br
Vary: Accept-Encoding
Transfer-Encoding: chunked

[brotli-compressed JSON body]
```

### Compression Algorithms Comparison

| Algorithm | Speed | Ratio | Support | Best For |
|-----------|-------|-------|---------|----------|
| gzip | Fast | Good | Universal | Default choice — works everywhere |
| Brotli (br) | Slower | Better | Modern browsers + Node.js | Static assets, large JSON responses |
| zstd | Very Fast | Very Good | Emerging | High-throughput APIs, streaming |
| deflate | Fast | Good | Universal | Legacy (prefer gzip) |

## Webhooks

### Webhook HTTP Patterns

```http
# Webhook delivery (server → client)
POST /webhooks/orders HTTP/1.1
Host: customer-app.example.com
Content-Type: application/json
X-Webhook-ID: wh-abc-123
X-Webhook-Signature: sha256=5d7a8f3b9c2e1d4a6f8b0c3e5d7a9f1b
X-Webhook-Timestamp: 1710500000

{
  "event": "order.completed",
  "data": {
    "orderId": "order-456",
    "total": 99.99,
    "status": "completed"
  },
  "timestamp": "2025-03-15T10:00:00Z"
}

# Expected response
HTTP/1.1 200 OK

# If non-2xx response, retry with exponential backoff:
# Retry 1: after 1 minute
# Retry 2: after 5 minutes
# Retry 3: after 30 minutes
# Retry 4: after 2 hours
# Retry 5: after 12 hours (final attempt)
```

### Testing Webhook Delivery

```typescript
describe('Webhook Delivery', () => {
  it('should deliver webhook on order completion', async () => {
    // Register a webhook endpoint
    const webhookUrl = 'https://test-webhook-receiver.example.com/hooks';

    await authApi(adminToken)
      .post('/api/webhooks')
      .send({ url: webhookUrl, events: ['order.completed'] })
      .expect(201);

    // Trigger the event (complete an order)
    await authApi(adminToken)
      .patch('/api/orders/order-123')
      .send({ status: 'completed' })
      .expect(200);

    // Verify webhook was sent (check your webhook receiver/mock)
    // In practice, use a tool like webhook.site or a test server
  });

  it('should include valid signature', async () => {
    // Verify the X-Webhook-Signature header
    const body = '{"event":"order.completed","data":{}}';
    const secret = 'webhook-secret';
    const timestamp = '1710500000';

    const expectedSignature = createHmac('sha256', secret)
      .update(`${timestamp}.${body}`)
      .digest('hex');

    // The signature header should be: sha256=<expected>
    expect(`sha256=${expectedSignature}`).toBeDefined();
  });

  it('should retry on failure with exponential backoff', async () => {
    // Register webhook to a URL that returns 500
    // Verify retries happen at increasing intervals
  });
});
```

## Server-Sent Events (SSE)

### SSE HTTP Pattern

```http
# Client initiates SSE connection
GET /api/events/orders HTTP/1.1
Accept: text/event-stream
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
Cache-Control: no-cache

# Server responds with event stream
HTTP/1.1 200 OK
Content-Type: text/event-stream
Cache-Control: no-cache
Connection: keep-alive

event: connected
data: {"status":"connected","timestamp":"2025-03-15T10:00:00Z"}

event: order.created
data: {"orderId":"order-789","total":49.99}
id: evt-1

event: order.shipped
data: {"orderId":"order-789","trackingNumber":"1Z999AA1"}
id: evt-2

: heartbeat (comment — keeps connection alive)

event: order.delivered
data: {"orderId":"order-789","deliveredAt":"2025-03-15T14:00:00Z"}
id: evt-3
```

### Testing SSE

```typescript
describe('Server-Sent Events', () => {
  it('should establish SSE connection', async () => {
    const response = await fetch('/api/events/orders', {
      headers: {
        'Accept': 'text/event-stream',
        'Authorization': `Bearer ${token}`,
      },
    });

    expect(response.status).toBe(200);
    expect(response.headers.get('content-type')).toContain('text/event-stream');
  });

  it('should support Last-Event-ID for reconnection', async () => {
    const response = await fetch('/api/events/orders', {
      headers: {
        'Accept': 'text/event-stream',
        'Authorization': `Bearer ${token}`,
        'Last-Event-ID': 'evt-2', // Resume from event 2
      },
    });

    // Should receive events after evt-2
    expect(response.status).toBe(200);
  });
});
```
