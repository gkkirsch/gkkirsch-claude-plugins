# Node.js Backend Checklist

## Project Setup

- [ ] `npm init -y` + TypeScript (`tsx` for dev, `tsc` for build)
- [ ] Environment validation with Zod (`src/config/env.ts`)
- [ ] Structured logging with `pino` + `pino-pretty` for dev
- [ ] Error handling middleware (last `app.use()`)
- [ ] 404 handler for unknown routes
- [ ] Graceful shutdown (SIGTERM, SIGINT)
- [ ] `.env` in `.gitignore`, `.env.example` committed
- [ ] Health check endpoint (`GET /health`)

## Architecture Decision

| Question | JWT | Sessions |
|----------|-----|----------|
| Client is SPA/mobile? | ✅ | ⚠️ Cookie issues with CORS |
| Need SSR with cookies? | ⚠️ | ✅ |
| Stateless scaling? | ✅ No server state | ❌ Shared session store |
| Token revocation? | ❌ Wait for expiry or blocklist | ✅ Delete from store |
| Mobile app? | ✅ | ❌ No cookie jar |
| Simple setup? | ❌ Access + refresh + rotation | ✅ express-session + store |

## Express Middleware Order

```
1. helmet()                    — Security headers
2. cors()                      — CORS (before routes!)
3. express.json({ limit })     — Body parser
4. express.urlencoded()        — Form data
5. requestLogger()             — Request logging
6. rateLimiter()               — Global rate limit
7. routes                      — Your API routes
8. 404 handler                 — Unknown routes
9. errorHandler                — Error middleware (MUST BE LAST)
```

## Per-Endpoint Checklist

- [ ] Input validated (Zod schema on body/query/params)
- [ ] Authentication checked (JWT/session middleware)
- [ ] Authorization checked (role + resource ownership)
- [ ] Rate limited (stricter for auth endpoints)
- [ ] Errors forwarded to error handler (`next(err)` or `asyncHandler`)
- [ ] Response follows standard format (`{ success, data }` or `{ success, error }`)
- [ ] No sensitive data in response (password, tokens filtered out)

## Database (Prisma)

- [ ] `prisma/schema.prisma` with proper relations and indexes
- [ ] Global PrismaClient with `globalThis` pattern
- [ ] Foreign keys have `@@index`
- [ ] Frequently filtered fields have `@@index`
- [ ] `onDelete` behavior explicit on relations
- [ ] Seed script for development data
- [ ] Migration files committed to git
- [ ] `prisma generate` in build step

## Authentication

| Item | JWT Flow | Session Flow |
|------|----------|--------------|
| Password hashing | bcrypt (cost 12) | bcrypt (cost 12) |
| Token storage (client) | Memory + httpOnly cookie | httpOnly cookie (auto) |
| Token storage (server) | Refresh tokens in DB | Session store (Postgres) |
| Expiry | Access: 15m, Refresh: 7d | Session: 24h |
| Rotation | Refresh on every use | Session ID on login |
| Revocation | Delete refresh from DB | Destroy session |

## Security Essentials

- [ ] `helmet()` enabled
- [ ] CORS with specific origins (not `*`)
- [ ] Rate limiting on auth endpoints (10/15min)
- [ ] Request body size limit (`express.json({ limit: '1mb' })`)
- [ ] Passwords hashed with bcrypt (12 rounds)
- [ ] JWT secrets from env vars (256+ bit)
- [ ] No sensitive data in JWT payload
- [ ] SQL injection prevention (parameterized queries / Prisma)
- [ ] XSS prevention (output encoding / React handles this)
- [ ] CSRF protection (SameSite cookies or CSRF tokens)
- [ ] `npm audit` clean (no critical vulnerabilities)
- [ ] Secrets not logged (pino `redact` option)

## Error Response Format

```json
{
  "success": false,
  "error": {
    "code": "MACHINE_READABLE_CODE",
    "message": "Human readable message",
    "details": {}
  }
}
```

| HTTP Status | Error Code | When |
|-------------|-----------|------|
| 400 | `VALIDATION_ERROR` | Invalid input |
| 400 | `INVALID_JSON` | Malformed request body |
| 401 | `UNAUTHORIZED` | Missing/invalid auth |
| 401 | `TOKEN_EXPIRED` | JWT expired |
| 403 | `FORBIDDEN` | Insufficient permissions |
| 404 | `NOT_FOUND` | Resource doesn't exist |
| 409 | `CONFLICT` / `DUPLICATE_ENTRY` | Already exists |
| 429 | `RATE_LIMIT_EXCEEDED` | Too many requests |
| 500 | `INTERNAL_ERROR` | Unexpected server error |

## Common npm Packages

| Category | Package | Purpose |
|----------|---------|---------|
| Framework | `express` | HTTP server |
| Validation | `zod` | Schema validation |
| Auth | `bcrypt` | Password hashing |
| Auth | `jsonwebtoken` | JWT tokens |
| Auth | `google-auth-library` | Google OAuth |
| Sessions | `express-session` | Session management |
| Sessions | `connect-pg-simple` | Postgres session store |
| Database | `@prisma/client` | ORM |
| Security | `helmet` | Security headers |
| Security | `cors` | CORS handling |
| Security | `express-rate-limit` | Rate limiting |
| Logging | `pino` | Structured logging |
| Logging | `pino-pretty` | Dev log formatting |
| Upload | `multer` | File uploads |
| Email | `nodemailer` | Send emails |
| Dev | `tsx` | TypeScript execution |
| Dev | `nodemon` | Auto-restart |

## Production Checklist

- [ ] `NODE_ENV=production` set
- [ ] Database migrations applied (`prisma migrate deploy`)
- [ ] All env vars set (validated by Zod on startup)
- [ ] HTTPS enforced
- [ ] Error messages don't leak internals
- [ ] Logging level set to `info` or `warn`
- [ ] Health check endpoint accessible
- [ ] Graceful shutdown handles in-flight requests
- [ ] Process manager (pm2, Docker, Heroku) restarts on crash
- [ ] Database connection pooling configured
- [ ] File uploads use cloud storage (S3/R2), not local disk
