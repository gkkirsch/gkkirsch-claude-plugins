# API Design Patterns Reference

Comprehensive reference for common API patterns: pagination, filtering, error handling, authentication, file uploads, batch operations, and real-time communication. Applies to both REST and GraphQL APIs.

---

## Pagination Patterns

### Cursor-Based Pagination (Recommended)

Best for: real-time data, infinite scroll, large datasets, data that changes frequently.

```
# Request
GET /api/posts?limit=20&cursor=eyJpZCI6MTAwfQ==

# Response
{
  "data": [...],
  "meta": {
    "count": 20,
    "hasMore": true,
    "nextCursor": "eyJpZCI6ODB9",
    "prevCursor": "eyJpZCI6MTAwfQ=="
  }
}
```

Cursor encoding:

```typescript
// Simple cursor: encode the sort field value
function encodeCursor(value: string | number | Date): string {
  const payload = JSON.stringify({ v: value instanceof Date ? value.toISOString() : value });
  return Buffer.from(payload).toString('base64url');
}

function decodeCursor(cursor: string): any {
  const payload = Buffer.from(cursor, 'base64url').toString('utf-8');
  return JSON.parse(payload).v;
}

// Composite cursor: encode multiple sort fields
function encodeCompositeCursor(values: Record<string, any>): string {
  return Buffer.from(JSON.stringify(values)).toString('base64url');
}

function decodeCompositeCursor(cursor: string): Record<string, any> {
  return JSON.parse(Buffer.from(cursor, 'base64url').toString('utf-8'));
}
```

SQL implementation:

```sql
-- Forward pagination with cursor (using createdAt as cursor field)
SELECT * FROM posts
WHERE created_at < '2024-01-15T10:00:00Z'  -- cursor value
ORDER BY created_at DESC
LIMIT 21;  -- limit + 1 to check hasMore

-- Composite cursor (sort by multiple fields)
SELECT * FROM posts
WHERE (created_at, id) < ('2024-01-15T10:00:00Z', 100)
ORDER BY created_at DESC, id DESC
LIMIT 21;
```

### Offset-Based Pagination

Best for: admin dashboards, table views, search results where page jumping is needed.

```
# Request
GET /api/users?page=3&pageSize=20

# Response
{
  "data": [...],
  "meta": {
    "totalCount": 156,
    "page": 3,
    "pageSize": 20,
    "totalPages": 8
  },
  "_links": {
    "self": "/api/users?page=3&pageSize=20",
    "first": "/api/users?page=1&pageSize=20",
    "prev": "/api/users?page=2&pageSize=20",
    "next": "/api/users?page=4&pageSize=20",
    "last": "/api/users?page=8&pageSize=20"
  }
}
```

### Keyset Pagination

Best for: ordered data where cursor represents the last seen sort key.

```
# Request — get items after a specific sort key
GET /api/products?after_price=29.99&after_id=500&limit=20

# SQL
SELECT * FROM products
WHERE (price, id) > (29.99, 500)
ORDER BY price ASC, id ASC
LIMIT 21;
```

### GraphQL Relay Connection Spec

```graphql
type PostConnection {
  edges: [PostEdge!]!
  pageInfo: PageInfo!
  totalCount: Int
}

type PostEdge {
  node: Post!
  cursor: String!
}

type PageInfo {
  hasNextPage: Boolean!
  hasPreviousPage: Boolean!
  startCursor: String
  endCursor: String
}

type Query {
  posts(
    first: Int       # Forward: number of items
    after: String    # Forward: cursor after which to start
    last: Int        # Backward: number of items
    before: String   # Backward: cursor before which to start
  ): PostConnection!
}
```

### Pagination Selection Guide

| Criteria | Cursor | Offset | Keyset |
|----------|--------|--------|--------|
| Large datasets | Best | Slow | Good |
| Real-time data | Best | Inconsistent | Good |
| Jump to page | No | Yes | No |
| Simple implementation | Medium | Easy | Medium |
| Database performance | O(1) | O(n) | O(1) |
| Infinite scroll | Best | OK | Good |
| Consistent results | Yes | No | Yes |

---

## Filtering Patterns

### Simple Equality Filters

```
GET /api/users?role=ADMIN&status=active
GET /api/posts?authorId=42&status=published
```

### Comparison Operators

```
# Bracket notation
GET /api/products?price[gte]=10&price[lte]=100
GET /api/events?date[after]=2024-01-01&date[before]=2024-12-31
GET /api/orders?total[gt]=50

# Supported operators:
# eq    — equals (default)
# ne    — not equals
# gt    — greater than
# gte   — greater than or equal
# lt    — less than
# lte   — less than or equal
# in    — in list
# nin   — not in list
# like  — pattern match (SQL LIKE)
# regex — regex match
```

### Multiple Values

```
# Comma-separated (OR logic)
GET /api/posts?status=published,draft
GET /api/products?category=electronics,books,toys

# Repeated parameter (OR logic)
GET /api/posts?tag=graphql&tag=api&tag=design
```

### Complex Filters (GraphQL)

```graphql
input PostFilter {
  title: StringFilter
  status: PostStatus
  authorId: ID
  createdAt: DateTimeFilter
  tags: StringListFilter
  AND: [PostFilter!]
  OR: [PostFilter!]
  NOT: PostFilter
}

input StringFilter {
  equals: String
  contains: String
  startsWith: String
  endsWith: String
  in: [String!]
  not: StringFilter
  mode: FilterMode
}

input DateTimeFilter {
  equals: DateTime
  gt: DateTime
  gte: DateTime
  lt: DateTime
  lte: DateTime
  in: [DateTime!]
}

input StringListFilter {
  has: String          # List contains value
  hasEvery: [String!]  # List contains all values
  hasSome: [String!]   # List contains any value
  isEmpty: Boolean     # List is empty
}

enum FilterMode {
  DEFAULT
  INSENSITIVE
}

# Usage
query {
  posts(filter: {
    OR: [
      { title: { contains: "graphql" } },
      { tags: { has: "graphql" } }
    ]
    status: PUBLISHED
    createdAt: { gte: "2024-01-01" }
  }) {
    id
    title
  }
}
```

### Field Selection (Sparse Fieldsets)

```
# REST — select specific fields to return
GET /api/users?fields=id,email,displayName
GET /api/posts?fields=id,title,createdAt

# JSON:API
GET /api/users?fields[users]=id,email&fields[posts]=id,title

# GraphQL — native field selection
query {
  user(id: "1") {
    id
    email
    # Only selected fields are returned
  }
}
```

### Sorting

```
# Single field sort (prefix - for descending)
GET /api/posts?sort=-createdAt          # Newest first
GET /api/users?sort=displayName         # Alphabetical

# Multiple field sort
GET /api/products?sort=-price,name      # Price desc, then name asc

# GraphQL
query {
  posts(orderBy: { field: CREATED_AT, direction: DESC }) {
    id
    title
  }
}

# Multiple sort fields in GraphQL
query {
  posts(orderBy: [
    { field: STATUS, direction: ASC }
    { field: CREATED_AT, direction: DESC }
  ]) {
    id
    title
  }
}
```

---

## Error Handling Patterns

### RFC 7807 Problem Details (REST)

```json
{
  "type": "https://api.example.com/errors/validation-error",
  "title": "Validation Error",
  "status": 422,
  "detail": "One or more fields failed validation",
  "instance": "/api/users",
  "errors": [
    {
      "field": "email",
      "message": "Must be a valid email address",
      "code": "INVALID_FORMAT"
    },
    {
      "field": "password",
      "message": "Must be at least 8 characters",
      "code": "TOO_SHORT"
    }
  ]
}
```

Content-Type: `application/problem+json`

### GraphQL Error Extensions

```json
{
  "data": null,
  "errors": [
    {
      "message": "User not found",
      "locations": [{ "line": 2, "column": 3 }],
      "path": ["user"],
      "extensions": {
        "code": "NOT_FOUND",
        "timestamp": "2024-01-15T10:30:00Z",
        "requestId": "req-abc-123"
      }
    }
  ]
}
```

### Error Code Taxonomy

```
Authentication Errors:
  UNAUTHENTICATED          — No valid credentials provided
  TOKEN_EXPIRED            — Auth token has expired
  TOKEN_INVALID            — Auth token is malformed
  TOKEN_REVOKED            — Auth token has been revoked

Authorization Errors:
  FORBIDDEN                — Authenticated but not authorized
  INSUFFICIENT_PERMISSIONS — Missing required permission/role
  RESOURCE_OWNERSHIP       — Can only access own resources

Input Errors:
  VALIDATION_ERROR         — Input validation failed
  INVALID_FORMAT           — Field format is wrong
  REQUIRED_FIELD           — Required field is missing
  TOO_SHORT                — Value below minimum length
  TOO_LONG                 — Value above maximum length
  OUT_OF_RANGE             — Numeric value out of allowed range
  INVALID_ENUM_VALUE       — Value not in allowed set

Resource Errors:
  NOT_FOUND                — Resource doesn't exist
  ALREADY_EXISTS           — Resource with same identifier exists
  CONFLICT                 — State conflict (optimistic locking)
  GONE                     — Resource permanently deleted

Rate Limiting:
  RATE_LIMITED             — Too many requests
  QUOTA_EXCEEDED           — API quota exceeded

Server Errors:
  INTERNAL_ERROR           — Unexpected server error
  SERVICE_UNAVAILABLE      — Temporary outage
  UPSTREAM_ERROR           — Dependency failure
```

### Union-Based Errors (GraphQL)

```graphql
union CreateUserResult =
  | CreateUserSuccess
  | ValidationErrors
  | ConflictError

type CreateUserSuccess {
  user: User!
}

type ValidationErrors {
  errors: [FieldError!]!
}

type FieldError {
  field: String!
  message: String!
  code: String!
}

type ConflictError {
  message: String!
  existingResourceId: ID!
}

# Client usage
mutation {
  createUser(input: { email: "test@example.com", displayName: "Test" }) {
    ... on CreateUserSuccess {
      user { id email }
    }
    ... on ValidationErrors {
      errors { field message code }
    }
    ... on ConflictError {
      message
      existingResourceId
    }
  }
}
```

---

## Authentication Patterns

### Bearer Token (JWT)

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

```typescript
// Token structure
interface JWTPayload {
  sub: string;      // User ID
  email: string;
  role: string;
  iat: number;      // Issued at
  exp: number;      // Expiration
  jti: string;      // Unique token ID (for revocation)
}

// Token lifecycle
const accessToken = jwt.sign(payload, secret, { expiresIn: '15m' });
const refreshToken = jwt.sign({ sub: userId }, refreshSecret, { expiresIn: '7d' });
```

### API Key

```
X-API-Key: sk_live_abc123...
```

Best for: server-to-server, CI/CD, simple integrations.

### OAuth 2.0

```
# Authorization Code Flow
1. GET /oauth/authorize?client_id=X&redirect_uri=Y&scope=Z&state=random
2. User authenticates and authorizes
3. Redirect to Y?code=AUTH_CODE&state=random
4. POST /oauth/token { grant_type: "authorization_code", code: AUTH_CODE }
5. Response: { access_token, refresh_token, expires_in, token_type }
```

### GraphQL Authentication

```typescript
// HTTP context (queries, mutations)
context: async ({ req }) => {
  const token = req.headers.authorization?.replace('Bearer ', '');
  const user = token ? await verifyToken(token) : null;
  return { user };
};

// WebSocket context (subscriptions)
context: async ({ connectionParams }) => {
  const token = connectionParams?.authorization?.replace('Bearer ', '');
  const user = token ? await verifyToken(token) : null;
  return { user };
};
```

---

## File Upload Patterns

### REST Multipart Upload

```typescript
import multer from 'multer';

const upload = multer({
  storage: multer.memoryStorage(),
  limits: { fileSize: 10 * 1024 * 1024 },  // 10MB
  fileFilter: (req, file, cb) => {
    const allowed = ['image/jpeg', 'image/png', 'image/webp'];
    cb(null, allowed.includes(file.mimetype));
  },
});

router.post('/upload', authenticate, upload.single('file'), async (req, res) => {
  const { buffer, originalname, mimetype, size } = req.file!;
  const url = await storageService.upload(buffer, originalname, mimetype);
  res.status(201).json({ data: { url, filename: originalname, mimetype, size } });
});
```

### Presigned URL Upload

```typescript
// Step 1: Client requests upload URL
router.post('/upload/presign', authenticate, async (req, res) => {
  const { filename, contentType } = req.body;
  const key = `uploads/${req.user!.id}/${Date.now()}-${filename}`;
  const url = await s3.getSignedUrl('putObject', {
    Bucket: process.env.S3_BUCKET,
    Key: key,
    ContentType: contentType,
    Expires: 300,  // 5 minutes
  });
  res.json({ data: { uploadUrl: url, key } });
});

// Step 2: Client uploads directly to S3 using the presigned URL
// Step 3: Client confirms upload
router.post('/upload/confirm', authenticate, async (req, res) => {
  const { key } = req.body;
  // Verify file exists, create database record, etc.
});
```

### GraphQL File Upload

```graphql
scalar Upload

type Mutation {
  uploadFile(file: Upload!): UploadResult!
}

type UploadResult {
  url: URL!
  filename: String!
  mimetype: String!
  size: Int!
}
```

---

## Batch Operations

### REST Batch Endpoint

```typescript
// POST /api/batch — execute multiple operations
router.post('/batch', authenticate, async (req, res) => {
  const operations = req.body.operations;  // Array of operations
  const results = [];

  for (const op of operations) {
    try {
      // Route internally
      const result = await executeOperation(op.method, op.path, op.body, req.user);
      results.push({ status: result.status, body: result.body });
    } catch (error) {
      results.push({ status: 500, body: { error: error.message } });
    }
  }

  res.json({ results });
});

// Request body
{
  "operations": [
    { "method": "PATCH", "path": "/api/posts/1", "body": { "status": "published" } },
    { "method": "PATCH", "path": "/api/posts/2", "body": { "status": "published" } },
    { "method": "DELETE", "path": "/api/posts/3" }
  ]
}
```

### Batch Create/Update/Delete

```typescript
// POST /api/posts/batch — batch create
router.post('/batch', async (req, res) => {
  const items = req.body.items;  // Array of items to create
  const results = await prisma.post.createMany({ data: items });
  res.status(201).json({ data: { createdCount: results.count } });
});

// PATCH /api/posts/batch — batch update
router.patch('/batch', async (req, res) => {
  const { ids, update } = req.body;
  const result = await prisma.post.updateMany({
    where: { id: { in: ids } },
    data: update,
  });
  res.json({ data: { updatedCount: result.count } });
});

// DELETE /api/posts/batch — batch delete
router.delete('/batch', async (req, res) => {
  const { ids } = req.body;
  const result = await prisma.post.deleteMany({
    where: { id: { in: ids } },
  });
  res.json({ data: { deletedCount: result.count } });
});
```

---

## Real-Time Patterns

### Server-Sent Events (SSE)

```typescript
// REST SSE endpoint
router.get('/events', authenticate, (req, res) => {
  res.setHeader('Content-Type', 'text/event-stream');
  res.setHeader('Cache-Control', 'no-cache');
  res.setHeader('Connection', 'keep-alive');

  // Send initial connection event
  res.write(`event: connected\ndata: ${JSON.stringify({ userId: req.user!.id })}\n\n`);

  // Subscribe to events
  const unsubscribe = eventBus.subscribe(req.user!.id, (event) => {
    res.write(`event: ${event.type}\ndata: ${JSON.stringify(event.data)}\nid: ${event.id}\n\n`);
  });

  // Send keepalive every 30 seconds
  const keepalive = setInterval(() => {
    res.write(`:keepalive\n\n`);
  }, 30000);

  // Cleanup on disconnect
  req.on('close', () => {
    clearInterval(keepalive);
    unsubscribe();
  });
});
```

### WebSocket Patterns

```typescript
// GraphQL subscriptions via WebSocket
// Client sends:
{ "type": "connection_init", "payload": { "authorization": "Bearer ..." } }

// Server responds:
{ "type": "connection_ack" }

// Client subscribes:
{ "id": "1", "type": "subscribe", "payload": { "query": "subscription { postCreated { id title } }" } }

// Server sends events:
{ "id": "1", "type": "next", "payload": { "data": { "postCreated": { "id": "42", "title": "New Post" } } } }

// Client unsubscribes:
{ "id": "1", "type": "complete" }
```

### Webhooks

```typescript
// Webhook registration
router.post('/webhooks', authenticate, async (req, res) => {
  const { url, events, secret } = req.body;
  const webhook = await webhookService.register({
    userId: req.user!.id,
    url,
    events,  // ['post.created', 'post.updated', 'comment.added']
    secret,  // Client-provided secret for signature verification
  });
  res.status(201).json({ data: webhook });
});

// Webhook delivery
async function deliverWebhook(webhook: Webhook, event: string, payload: any) {
  const timestamp = Date.now();
  const signature = createHmac('sha256', webhook.secret)
    .update(`${timestamp}.${JSON.stringify(payload)}`)
    .digest('hex');

  const response = await fetch(webhook.url, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'X-Webhook-Signature': `t=${timestamp},v1=${signature}`,
      'X-Webhook-Event': event,
      'X-Webhook-ID': randomUUID(),
    },
    body: JSON.stringify({ event, data: payload, timestamp }),
  });

  // Retry with exponential backoff on failure
  if (!response.ok) {
    await retryQueue.add({ webhookId: webhook.id, event, payload }, {
      attempts: 5,
      backoff: { type: 'exponential', delay: 1000 },
    });
  }
}
```

---

## Idempotency

### Idempotency Key Pattern

```typescript
// Client sends idempotency key for non-idempotent operations
// POST /api/payments
// Idempotency-Key: unique-request-id-abc123

router.post('/payments', authenticate, async (req, res) => {
  const idempotencyKey = req.headers['idempotency-key'] as string;

  if (!idempotencyKey) {
    return res.status(400).json({
      type: 'https://api.example.com/errors/missing-idempotency-key',
      title: 'Missing Idempotency Key',
      status: 400,
      detail: 'POST requests require an Idempotency-Key header',
    });
  }

  // Check if we've seen this key before
  const existing = await redis.get(`idempotency:${idempotencyKey}`);
  if (existing) {
    const cached = JSON.parse(existing);
    return res.status(cached.status).json(cached.body);
  }

  // Process the request
  const result = await paymentService.process(req.body);
  const response = { status: 201, body: { data: result } };

  // Cache the response for 24 hours
  await redis.setex(
    `idempotency:${idempotencyKey}`,
    86400,
    JSON.stringify(response)
  );

  res.status(201).json(response.body);
});
```

---

## Rate Limiting

### Rate Limit Headers

```
# Standard headers (IETF draft)
RateLimit-Limit: 100        # Max requests per window
RateLimit-Remaining: 42     # Remaining requests in window
RateLimit-Reset: 1705312800 # Unix timestamp when window resets

# On 429 response
Retry-After: 30             # Seconds to wait before retrying
```

### Tiered Rate Limits

```typescript
const rateLimits: Record<string, { limit: number; window: number }> = {
  // By endpoint type
  'read': { limit: 1000, window: 900 },      // 1000/15min
  'write': { limit: 100, window: 900 },       // 100/15min
  'search': { limit: 30, window: 60 },        // 30/min
  'auth': { limit: 5, window: 300 },          // 5/5min
  'upload': { limit: 10, window: 3600 },      // 10/hour

  // By API tier
  'free': { limit: 100, window: 3600 },       // 100/hour
  'basic': { limit: 1000, window: 3600 },     // 1000/hour
  'pro': { limit: 10000, window: 3600 },      // 10000/hour
  'enterprise': { limit: 100000, window: 3600 }, // 100000/hour
};
```

---

## Content Negotiation

```typescript
// Accept header negotiation
router.get('/api/users/:id', (req, res) => {
  const user = getUserById(req.params.id);
  const accept = req.accepts(['json', 'xml', 'csv']);

  switch (accept) {
    case 'json':
      res.type('json').json({ data: user });
      break;
    case 'xml':
      res.type('xml').send(toXML(user));
      break;
    case 'csv':
      res.type('csv').send(toCSV(user));
      break;
    default:
      res.status(406).json({
        type: 'https://api.example.com/errors/not-acceptable',
        title: 'Not Acceptable',
        status: 406,
      });
  }
});
```

---

## Health Check Endpoints

```typescript
// GET /health — basic health check
router.get('/health', (req, res) => {
  res.json({ status: 'ok', timestamp: new Date().toISOString() });
});

// GET /health/ready — readiness check (dependencies healthy)
router.get('/health/ready', async (req, res) => {
  const checks = {
    database: await checkDatabase(),
    redis: await checkRedis(),
    storage: await checkStorage(),
  };

  const allHealthy = Object.values(checks).every(c => c.status === 'ok');

  res.status(allHealthy ? 200 : 503).json({
    status: allHealthy ? 'ok' : 'degraded',
    checks,
    timestamp: new Date().toISOString(),
  });
});

// GET /health/live — liveness check (process is running)
router.get('/health/live', (req, res) => {
  res.status(200).json({ status: 'ok' });
});
```

---

## API Discovery

### REST API Root

```json
// GET /api
{
  "name": "My API",
  "version": "2.0.0",
  "documentation": "https://docs.example.com",
  "_links": {
    "self": { "href": "/api" },
    "users": { "href": "/api/users", "title": "User management" },
    "posts": { "href": "/api/posts", "title": "Blog posts" },
    "search": { "href": "/api/search{?q,type}", "templated": true },
    "health": { "href": "/api/health" },
    "docs": { "href": "/api/docs" }
  }
}
```

### GraphQL Introspection

```graphql
# Self-documenting via introspection
{
  __schema {
    queryType { name }
    mutationType { name }
    subscriptionType { name }
    types {
      name
      description
      fields { name description type { name } }
    }
  }
}
```
