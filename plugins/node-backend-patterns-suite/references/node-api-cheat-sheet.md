# Node.js API Cheat Sheet

## Express Middleware Order

```
1. helmet()              — Security headers
2. cors()                — CORS configuration
3. express.json()        — Body parsing (set limit!)
4. express.urlencoded()  — Form parsing
5. requestId()           — Request tracking
6. requestLogger()       — Request logging
7. rateLimiter()         — Rate limiting
8. routes                — Your API routes
9. notFoundHandler       — 404 catch-all
10. errorHandler         — Global error handler (MUST be last, MUST have 4 params)
```

## HTTP Status Codes

| Code | Name | When to Use |
|------|------|-------------|
| 200 | OK | Successful GET, PUT, PATCH |
| 201 | Created | Successful POST that creates a resource |
| 204 | No Content | Successful DELETE (no body) |
| 400 | Bad Request | Invalid input, malformed JSON |
| 401 | Unauthorized | Missing or invalid authentication |
| 403 | Forbidden | Authenticated but not authorized |
| 404 | Not Found | Resource doesn't exist |
| 409 | Conflict | Duplicate resource (unique constraint) |
| 422 | Unprocessable | Valid JSON but invalid data |
| 429 | Too Many Requests | Rate limit exceeded |
| 500 | Internal Error | Unexpected server error |
| 503 | Unavailable | Database down, external service down |

## RESTful Route Naming

```
GET    /api/posts              List posts (paginated)
GET    /api/posts/:id          Get single post
POST   /api/posts              Create post
PUT    /api/posts/:id          Replace post (full update)
PATCH  /api/posts/:id          Partial update
DELETE /api/posts/:id          Delete post

GET    /api/posts/:id/comments      List comments for post
POST   /api/posts/:id/comments      Add comment to post
DELETE /api/posts/:id/comments/:cid Delete specific comment

GET    /api/posts?page=1&limit=20&sort=createdAt&order=desc   Pagination
GET    /api/posts?search=nodejs&category=tech                 Filtering
```

**Naming rules:**
- Plural nouns (`/posts` not `/post`)
- No verbs (`POST /posts` not `POST /createPost`)
- Lowercase, hyphenated (`/user-profiles` not `/userProfiles`)
- Nesting max 2 levels deep (`/posts/:id/comments` is fine, `/posts/:id/comments/:cid/likes` is too deep — use `/comments/:cid/likes`)

## Zod Schema Patterns

```typescript
// String validations
z.string().min(1)              // Required (empty string fails)
z.string().email()             // Email format
z.string().url()               // URL format
z.string().uuid()              // UUID format
z.string().max(200)            // Max length
z.string().trim()              // Auto-trim whitespace
z.string().regex(/^[a-z]+$/)   // Custom regex

// Number coercion (from query params)
z.coerce.number().int().positive()    // "42" → 42
z.coerce.number().min(1).max(100)     // Range

// Optional vs nullable
z.string().optional()          // string | undefined
z.string().nullable()          // string | null
z.string().nullish()           // string | null | undefined

// Defaults
z.string().default('untitled')
z.boolean().default(false)
z.number().default(1)

// Transform
z.string().transform(s => s.toLowerCase())
z.string().transform(s => slugify(s))

// Enum
z.enum(['draft', 'published', 'archived'])

// Partial (all fields optional)
createSchema.partial()

// Pick / Omit
schema.pick({ title: true, content: true })
schema.omit({ password: true })

// Infer TypeScript type
type CreatePost = z.infer<typeof createPostSchema>;
```

## JWT Token Flow

```
1. Client sends credentials → POST /api/auth/login
2. Server validates → returns { accessToken, refreshToken }
3. Client stores accessToken in memory, refreshToken in httpOnly cookie
4. Client sends accessToken in Authorization: Bearer <token>
5. When accessToken expires (401) → POST /api/auth/refresh with refreshToken
6. Server rotates refreshToken → returns new { accessToken, refreshToken }
7. Logout → POST /api/auth/logout → server revokes refreshToken
```

| Token | Lifetime | Storage | Purpose |
|-------|----------|---------|---------|
| Access | 15 min | Memory | API authentication |
| Refresh | 7 days | httpOnly cookie or DB | Get new access tokens |
| API Key | Long-lived | Env var | Server-to-server auth |

## Password Hashing

```typescript
import bcrypt from 'bcrypt';

// Hash (signup)
const hash = await bcrypt.hash(password, 12);  // cost factor 12

// Compare (login)
const valid = await bcrypt.compare(password, hash);  // true/false

// Never: store plaintext, use MD5/SHA for passwords, use cost < 10
```

## Prisma Quick Reference

```bash
npx prisma init                     # Initialize Prisma
npx prisma generate                 # Generate client from schema
npx prisma migrate dev --name init  # Create + apply migration
npx prisma migrate deploy           # Apply in production
npx prisma db push                  # Push schema (no migration)
npx prisma db pull                  # Pull schema from DB
npx prisma db seed                  # Run seed script
npx prisma studio                   # Visual DB browser
npx prisma migrate reset            # Reset DB (dev only!)
```

```typescript
// CRUD
prisma.post.findMany({ where, select, include, orderBy, skip, take })
prisma.post.findUnique({ where: { id } })        // Requires unique field
prisma.post.findFirst({ where })                  // First match
prisma.post.create({ data })
prisma.post.update({ where: { id }, data })
prisma.post.delete({ where: { id } })
prisma.post.upsert({ where, create, update })
prisma.post.count({ where })

// Bulk
prisma.post.createMany({ data: [...] })
prisma.post.updateMany({ where, data })
prisma.post.deleteMany({ where })

// Relations
include: { author: true }                        // Full relation
select: { author: { select: { name: true } } }   // Partial relation
include: { _count: { select: { comments: true } } }  // Count

// Filters
where: { title: { contains: 'node', mode: 'insensitive' } }
where: { createdAt: { gte: startDate, lte: endDate } }
where: { OR: [{ title: {...} }, { content: {...} }] }
where: { tags: { some: { name: 'typescript' } } }
where: { NOT: { published: false } }

// Transaction
prisma.$transaction([query1, query2])             // Batch
prisma.$transaction(async (tx) => { ... })        // Interactive
```

## Error Classes Quick Reference

| Error Class | Status | Code | Use For |
|-------------|--------|------|---------|
| BadRequestError | 400 | BAD_REQUEST | Invalid request format |
| ValidationError | 400 | VALIDATION_ERROR | Zod/schema failures |
| UnauthorizedError | 401 | UNAUTHORIZED | Missing/invalid auth |
| ForbiddenError | 403 | FORBIDDEN | Insufficient permissions |
| NotFoundError | 404 | NOT_FOUND | Resource doesn't exist |
| ConflictError | 409 | CONFLICT | Duplicate resource |
| TooManyRequestsError | 429 | RATE_LIMIT | Rate limit exceeded |

## Rate Limiting Presets

| Endpoint Type | Window | Max Requests |
|---------------|--------|-------------|
| General API | 15 min | 100 |
| Auth (login/signup) | 15 min | 10 |
| Password reset | 1 hour | 3 |
| File upload | 1 hour | 20 |
| Public read | 1 min | 60 |

## Security Headers (Helmet Defaults)

| Header | Value | Purpose |
|--------|-------|---------|
| X-Content-Type-Options | nosniff | Prevent MIME sniffing |
| X-Frame-Options | SAMEORIGIN | Prevent clickjacking |
| X-XSS-Protection | 0 | Disable broken XSS filter |
| Strict-Transport-Security | max-age=15552000 | Force HTTPS |
| Content-Security-Policy | default-src 'self' | Prevent XSS |
| X-DNS-Prefetch-Control | off | Privacy |
| Referrer-Policy | no-referrer | Privacy |

## Project Scripts

```json
{
  "scripts": {
    "dev": "tsx watch src/server.ts",
    "build": "tsc",
    "start": "node dist/server.js",
    "db:migrate": "prisma migrate dev",
    "db:deploy": "prisma migrate deploy",
    "db:seed": "prisma db seed",
    "db:studio": "prisma studio",
    "db:reset": "prisma migrate reset",
    "lint": "eslint src/",
    "test": "vitest",
    "test:e2e": "vitest --config vitest.e2e.config.ts"
  }
}
```

## Essential Dependencies

```
Production:
  express           — Web framework
  helmet            — Security headers
  cors              — CORS middleware
  @prisma/client    — Database ORM
  bcrypt            — Password hashing
  jsonwebtoken      — JWT auth
  zod               — Validation
  pino              — Logging
  express-rate-limit — Rate limiting

Development:
  typescript        — TypeScript compiler
  tsx               — TS runner with watch mode
  @types/express    — Express types
  @types/bcrypt     — bcrypt types
  @types/jsonwebtoken — JWT types
  prisma            — Prisma CLI
  vitest            — Testing
  supertest         — HTTP testing
```
