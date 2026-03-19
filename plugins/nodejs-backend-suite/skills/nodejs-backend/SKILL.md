# Node.js Backend Mastery Suite

Expert Node.js backend development suite with 4 specialized agents for Express/Fastify architecture, performance optimization, API security, and database integration.

## Agents

### Express Architect
Expert in Express 5 and Fastify 4 architecture. Handles:
- Application setup and middleware composition
- Router organization and RESTful API design
- Error handling (centralized, async, custom error classes)
- Graceful shutdown, health checks, readiness probes
- Project structure and layered architecture
- Request validation with Zod
- Logging with Pino
- File uploads, SSE, WebSockets
- Express 4 to 5 migration

**Invoke when**: Building APIs, designing routes, setting up middleware, handling errors, structuring projects.

### Node.js Performance Expert
Expert in Node.js runtime performance. Handles:
- Event loop analysis (phases, microtasks, blocking detection)
- Memory profiling (V8 heap, leak detection, GC monitoring)
- CPU profiling (flame graphs, V8 profiler, benchmarking)
- Clustering and scaling (node:cluster, PM2, load balancing)
- Worker threads (CPU offloading, worker pools, SharedArrayBuffer)
- Streams (Transform, backpressure, pipeline composition)
- Caching strategies (LRU, Redis, stale-while-revalidate)
- Database query optimization

**Invoke when**: Diagnosing slow responses, memory leaks, high CPU, scaling issues, or optimizing throughput.

### API Security Engineer
Expert in securing Node.js APIs. Handles:
- JWT authentication with refresh token rotation
- RBAC and ABAC authorization patterns
- Rate limiting (token bucket, sliding window, cost-based)
- Input validation and sanitization
- Security headers (Helmet, CORS, CSP, HSTS)
- OWASP Top 10 prevention (injection, SSRF, IDOR, XSS)
- API key management and webhook verification
- Password hashing with Argon2id
- Security audits and compliance

**Invoke when**: Implementing auth, securing endpoints, preventing attacks, rate limiting, or auditing security.

### Database Integration Expert
Expert in database design and ORM integration. Handles:
- Prisma 5 (schema, migrations, transactions, middleware, relations)
- Drizzle ORM (schema, queries, prepared statements, migrations)
- TypeORM (entities, repositories, query builder)
- Schema design (normalization, indexing, soft deletes, audit trails)
- Safe migration patterns (zero-downtime, backfills)
- Transactions (ACID, isolation levels, optimistic/pessimistic locking)
- Connection pooling (sizing, PgBouncer, serverless)
- Query optimization (EXPLAIN, N+1 prevention, covering indexes)
- Seeding, test databases, Testcontainers

**Invoke when**: Designing schemas, setting up ORMs, writing migrations, optimizing queries, or handling transactions.

## Quick Start

### New Express 5 API
```
/nodejs-backend set up a new Express 5 API with TypeScript, Prisma, and JWT auth
```

### Diagnose Performance
```
/nodejs-backend my API response times are high and I see memory growing over time
```

### Secure an API
```
/nodejs-backend implement JWT auth with refresh tokens and role-based access control
```

### Database Setup
```
/nodejs-backend set up Prisma with PostgreSQL, design schema for a blog app with users, posts, categories, and tags
```

## References

This suite includes detailed reference guides:

- **Express Patterns** — Middleware order, error handling, validation, response formats, deployment checklist
- **Node.js Internals** — Event loop phases, V8 memory, GC, libuv, async hooks, perf_hooks, CLI flags
- **Backend Testing Guide** — Vitest setup, Supertest, Testcontainers, mocking, CI/CD pipelines

## Stack Coverage

| Category | Technologies |
|----------|-------------|
| **Frameworks** | Express 5, Fastify 4 |
| **Language** | TypeScript (ESM) |
| **Validation** | Zod, TypeBox (Fastify) |
| **ORMs** | Prisma 5, Drizzle ORM, TypeORM |
| **Databases** | PostgreSQL, MySQL, SQLite |
| **Auth** | JWT, Argon2id, OAuth 2.0/OIDC |
| **Caching** | Redis (ioredis), LRU Cache |
| **Logging** | Pino, pino-http |
| **Testing** | Vitest, Supertest, Testcontainers |
| **Process Mgmt** | node:cluster, PM2, Docker |
| **Monitoring** | Prometheus (prom-client), OpenTelemetry |
| **Security** | Helmet, CORS, rate limiting |

## Best Practices Enforced

1. **Always validate input** — Zod schemas on every endpoint
2. **Never leak internals** — Generic error messages in production
3. **Always use parameterized queries** — No string interpolation in SQL
4. **Always hash passwords** — Argon2id with OWASP-recommended parameters
5. **Always set security headers** — Helmet with strict CSP
6. **Always implement graceful shutdown** — SIGTERM handling with connection draining
7. **Always use structured logging** — Pino, never console.log
8. **Always validate environment** — Zod schema at startup
9. **Always use connection pooling** — Sized appropriately for workload
10. **Always test** — Unit, integration, and load tests
