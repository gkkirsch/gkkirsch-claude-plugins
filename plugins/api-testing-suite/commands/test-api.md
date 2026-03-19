---
name: test-api
description: >
  Quick API testing command — analyzes your API codebase, discovers endpoints, generates test suites,
  runs them, and provides coverage reports. Can also generate OpenAPI docs from your code.
  Triggers: "/test-api", "test my api", "test endpoints", "api tests", "generate api docs".
user-invocable: true
argument-hint: "<test|doc|load|security> [endpoint-or-path] [--verbose] [--format openapi|markdown]"
allowed-tools: Read, Write, Edit, Bash, Glob, Grep
model: sonnet
---

# /test-api Command

One-command API testing and documentation. Analyzes your API, discovers endpoints, generates tests, runs them, and optionally generates OpenAPI documentation.

## Usage

```
/test-api                       # Auto-discover and test all endpoints
/test-api test                  # Run API test suite
/test-api test /api/users       # Test specific endpoint
/test-api doc                   # Generate OpenAPI documentation
/test-api doc --format markdown # Generate markdown API reference
/test-api load /api/users       # Load test specific endpoint
/test-api security              # Run API security audit
/test-api security /api/auth    # Security audit specific endpoint
```

## Subcommands

| Subcommand | Agent | Description |
|------------|-------|-------------|
| `test` | api-tester | Discover endpoints, generate & run test suites |
| `doc` | api-doc-writer | Generate OpenAPI specs or markdown API docs |
| `load` | load-tester | Performance & load testing with k6/Artillery |
| `security` | api-security | OWASP API Top 10 security audit |

## Procedure

### Step 1: Detect API Framework

Read the project root to understand the API stack:

1. **Read** `package.json`, `requirements.txt`, `go.mod`, `Cargo.toml`, `Gemfile`, `pom.xml` — determine language/framework
2. **Glob** for route/controller files:
   - Express: `**/routes/**`, `**/controllers/**`, `app.{js,ts}`, `server.{js,ts}`
   - FastAPI: `**/routers/**`, `main.py`, `app.py`
   - Django: `**/urls.py`, `**/views.py`
   - Rails: `config/routes.rb`, `app/controllers/**`
   - Spring: `**/*Controller.java`, `**/*Resource.java`
   - Go: `**/handler*.go`, `**/router*.go`, `main.go`
3. **Grep** for route definitions:
   - Express: `router.get|post|put|delete|patch`, `app.get|post|put|delete|patch`
   - FastAPI: `@app.get|post|put|delete|patch`, `@router.get|post|put|delete|patch`
   - Django: `path(`, `re_path(`
   - Rails: `get '`, `post '`, `resources :`
4. **Check for existing tests**: `**/*.test.*`, `**/*.spec.*`, `**/test_*`, `**/*_test.*`
5. **Check for existing docs**: `**/openapi.*`, `**/swagger.*`, `**/api-docs/**`

Report findings:

```
Detected:
- Language: TypeScript (Node.js)
- Framework: Express
- Endpoints found: 24 (8 GET, 6 POST, 5 PUT, 3 DELETE, 2 PATCH)
- Existing tests: 12 (50% coverage)
- Existing docs: None
- Auth: JWT (Bearer token)
- Database: PostgreSQL (Prisma ORM)
```

### Step 2: Route to Agent

Based on the subcommand, dispatch the appropriate agent:

#### `test` — API Tester Agent

```
Task tool:
  subagent_type: "api-tester"
  mode: "bypassPermissions"
  prompt: |
    Test the API endpoints for this project.
    Stack: [detected stack]
    Endpoints: [discovered endpoints]
    Auth pattern: [detected auth]
    Existing tests: [what already exists]
    Target: [specific endpoint if provided, or all]
    Generate comprehensive test suites covering:
    - Happy path tests for each endpoint
    - Error handling (400, 401, 403, 404, 422, 500)
    - Edge cases (empty body, missing fields, invalid types)
    - Auth flows (valid token, expired token, no token, wrong role)
    - Write tests using the project's existing test framework
```

#### `doc` — API Documentation Writer

```
Task tool:
  subagent_type: "api-doc-writer"
  mode: "bypassPermissions"
  prompt: |
    Generate API documentation for this project.
    Stack: [detected stack]
    Endpoints: [discovered endpoints]
    Format: [openapi or markdown, from --format flag]
    Existing docs: [what already exists]
    Generate comprehensive documentation including:
    - All endpoints with request/response schemas
    - Authentication requirements
    - Error response formats
    - Example requests and responses
    - Write to docs/api/ directory
```

#### `load` — Load Tester Agent

```
Task tool:
  subagent_type: "load-tester"
  mode: "bypassPermissions"
  prompt: |
    Create and run load tests for this API.
    Stack: [detected stack]
    Target endpoint: [specific endpoint or all]
    Auth pattern: [detected auth]
    Generate k6 or Artillery load test scripts that:
    - Simulate realistic user patterns
    - Test endpoint response times under load
    - Identify bottlenecks and breaking points
    - Generate a performance report
```

#### `security` — API Security Agent

```
Task tool:
  subagent_type: "api-security"
  mode: "bypassPermissions"
  prompt: |
    Run a security audit on this API.
    Stack: [detected stack]
    Endpoints: [discovered endpoints]
    Auth pattern: [detected auth]
    Target: [specific endpoint or all]
    Check for OWASP API Top 10 vulnerabilities:
    - Broken authentication and authorization
    - Injection vulnerabilities
    - Mass assignment / excessive data exposure
    - Rate limiting gaps
    - Security misconfiguration
    - Generate a security report with findings and fixes
```

### Step 3: Results Summary

After the agent completes, present results:

**For tests:**
```
API Test Results:
- Tests generated: 48 (24 endpoints x 2 avg tests each)
- Tests passing: 45
- Tests failing: 3
  - POST /api/users: Missing validation for email format
  - PUT /api/orders/:id: Returns 500 instead of 404 for non-existent order
  - DELETE /api/admin/users/:id: No role check — any authenticated user can delete
- Coverage: 87% of endpoints
- Files created: tests/api/*.test.ts
```

**For docs:**
```
API Documentation Generated:
- Format: OpenAPI 3.1
- Endpoints documented: 24
- Schemas defined: 15
- Files created: docs/api/openapi.yaml, docs/api/README.md
- View: npx @redocly/cli preview-docs docs/api/openapi.yaml
```

**For load tests:**
```
Load Test Results:
- Endpoint: POST /api/orders
- Virtual users: 100 concurrent
- Duration: 60s
- Avg response time: 145ms
- P95 response time: 380ms
- P99 response time: 720ms
- Errors: 2.1% (timeout at high concurrency)
- Bottleneck: Database query in order validation (avg 89ms)
```

**For security:**
```
Security Audit Results:
- Critical: 1 (SQL injection in search endpoint)
- High: 2 (Missing rate limiting on auth, CORS wildcard in production)
- Medium: 3 (Verbose error messages, missing security headers, no input size limits)
- Low: 4 (No request ID tracking, missing API versioning, ...)
- Report: docs/security-audit.md
```

## Error Recovery

| Error | Cause | Fix |
|-------|-------|-----|
| No endpoints found | Non-standard routing | Provide route files manually |
| Tests can't connect | Server not running | Start dev server first |
| Auth failures in tests | Missing test credentials | Set up test env vars |
| Load test timeout | Server overwhelmed | Reduce virtual users |
| False security positives | Framework handles it | Review and dismiss |

## Notes

- Always run tests against a development/test environment, never production
- Load tests can stress your database — use a test database
- Security audit findings should be triaged by severity before fixing
- Generated OpenAPI docs can be used with Swagger UI, Redoc, or Postman
