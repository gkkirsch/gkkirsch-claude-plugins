# Node.js Backend Mastery Command

## Usage
`/nodejs-backend [topic]`

## Description
Expert Node.js backend development assistance. Covers Express/Fastify architecture, performance optimization, API security, and database integration. Routes to the most appropriate specialist agent based on your request.

## Topics

### Architecture & Routing
- Express 5 app setup, middleware composition, router organization
- Fastify 4 plugin system, hooks lifecycle, schema validation
- RESTful API design, versioning, error handling
- Project structure, layered architecture, dependency injection
- **Agent**: Express Architect

### Performance
- Event loop analysis, blocking detection, profiling
- Memory leak detection, V8 heap analysis, GC tuning
- CPU profiling, flame graphs, benchmarking
- Worker threads, clustering, stream processing
- Caching strategies (LRU, Redis), database query optimization
- **Agent**: Node.js Performance Expert

### Security
- JWT authentication with refresh token rotation
- RBAC/ABAC authorization, permission middleware
- Rate limiting (token bucket, sliding window, cost-based)
- Input validation with Zod, sanitization, mass assignment prevention
- OWASP Top 10 prevention (injection, SSRF, IDOR, XSS)
- Security headers (Helmet, CORS, CSP, HSTS)
- API key management, webhook signature verification
- **Agent**: API Security Engineer

### Database
- Prisma 5 schema design, migrations, transactions, middleware
- Drizzle ORM setup, query builder, prepared statements
- TypeORM entities, repositories, query builder
- Schema design, indexing strategies, query optimization
- Connection pooling, PgBouncer, serverless connections
- Optimistic/pessimistic locking, isolation levels
- Seeding, test databases, Testcontainers
- **Agent**: Database Integration Expert

## Routing Logic

When a user invokes `/nodejs-backend`, analyze their request:

1. **Architecture/routing/middleware/error handling** → Express Architect
2. **Performance/memory/CPU/scaling/streams/caching** → Node.js Performance Expert
3. **Auth/security/rate limiting/validation/OWASP** → API Security Engineer
4. **Database/ORM/schema/migrations/queries/transactions** → Database Integration Expert
5. **General or unclear** → Ask which area they need help with

## Examples

```
/nodejs-backend set up a new Express 5 API with TypeScript
→ Routes to Express Architect

/nodejs-backend my API has high memory usage and occasional OOM kills
→ Routes to Node.js Performance Expert

/nodejs-backend implement JWT auth with refresh token rotation
→ Routes to API Security Engineer

/nodejs-backend design a Prisma schema for a blog with categories and tags
→ Routes to Database Integration Expert

/nodejs-backend add rate limiting to my authentication endpoints
→ Routes to API Security Engineer

/nodejs-backend set up worker threads for image processing
→ Routes to Node.js Performance Expert

/nodejs-backend configure Drizzle with PostgreSQL and write migrations
→ Routes to Database Integration Expert

/nodejs-backend add graceful shutdown and health checks
→ Routes to Express Architect
```

## Cross-Cutting Concerns

Some topics span multiple agents. Route to the primary expert:

| Topic | Primary Agent | Also Relevant |
|-------|---------------|---------------|
| Connection pooling | Database Integration Expert | Performance Expert |
| Request validation | API Security Engineer | Express Architect |
| Error handling | Express Architect | API Security Engineer |
| Query optimization | Database Integration Expert | Performance Expert |
| Rate limiting | API Security Engineer | Performance Expert |
| Health checks | Express Architect | Performance Expert |
| Graceful shutdown | Express Architect | Performance Expert |
| Logging | Express Architect | API Security Engineer |
| Caching | Performance Expert | Database Integration Expert |

## Quick Reference

### Project Setup Checklist
```
1. TypeScript + ESM configuration
2. Express 5 / Fastify 4 app factory
3. Middleware stack (Helmet, CORS, compression, logging)
4. Router organization with validation
5. Centralized error handling
6. Database setup (Prisma/Drizzle) with connection pooling
7. Authentication (JWT with refresh tokens)
8. Authorization (RBAC middleware)
9. Rate limiting (per-endpoint)
10. Health check + readiness endpoints
11. Graceful shutdown
12. Structured logging (Pino)
13. Environment validation (Zod)
14. Testing setup (Vitest + Supertest + Testcontainers)
```

### Key Dependencies
```json
{
  "express": "^5.0.0",
  "helmet": "^8.0.0",
  "cors": "^2.8.5",
  "compression": "^1.7.4",
  "pino": "^9.0.0",
  "pino-http": "^10.0.0",
  "zod": "^3.23.0",
  "@prisma/client": "^5.20.0",
  "ioredis": "^5.4.0",
  "jsonwebtoken": "^9.0.0",
  "@node-rs/argon2": "^2.0.0"
}
```
