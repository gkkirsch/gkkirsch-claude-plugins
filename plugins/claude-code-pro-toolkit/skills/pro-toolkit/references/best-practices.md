# Best Practices Reference

Quick reference for the Pro Toolkit agents. These patterns are injected as context when agents need guidance.

---

## Security Best Practices

### Authentication
- Hash passwords with bcrypt (cost factor >= 12) or argon2id.
- JWTs: use RS256 or ES256 (not HS256 with weak secrets). Set short expiration (15 min access, 7 day refresh). Validate `iss`, `aud`, `exp` claims.
- Store refresh tokens server-side, not in localStorage. Use HttpOnly, Secure, SameSite=Strict cookies.
- Implement account lockout after 5-10 failed attempts.
- Rate limit auth endpoints: 5 attempts per minute per IP.

### Input Validation
- Validate at the boundary (API endpoints, form handlers), not deep in business logic.
- Use schema validation (Zod, Joi, Pydantic) for all request bodies.
- Validate and sanitize file uploads: check MIME type, limit size, rename files.
- Never trust client-side validation alone.

### Database
- Always use parameterized queries. Never concatenate user input into SQL.
- Apply principle of least privilege to database users.
- Encrypt sensitive data at rest (PII, financial data, health records).

### HTTP Security Headers
```
Content-Security-Policy: default-src 'self'; script-src 'self'
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
Strict-Transport-Security: max-age=31536000; includeSubDomains
Referrer-Policy: strict-origin-when-cross-origin
Permissions-Policy: camera=(), microphone=(), geolocation=()
```

---

## Testing Best Practices

### Test Pyramid
- **Unit tests** (70%): Test individual functions/methods in isolation. Fast, many.
- **Integration tests** (20%): Test module interactions. Database queries, API endpoints.
- **E2E tests** (10%): Test complete user workflows. Slow, few, but critical.

### Test Naming
- Describe the behavior, not the implementation: `it('rejects expired tokens')` not `it('calls isExpired and returns false')`.
- Use the pattern: `it('<action> when <condition>')` or `it('should <expected behavior>')`.

### Mocking Strategy
- Mock external dependencies (APIs, databases, file system, time).
- Don't mock the code under test.
- Prefer dependency injection over module mocking.
- Reset mocks between tests (`beforeEach(() => vi.clearAllMocks())`).

### Assertion Quality
- Assert on specific values, not truthiness: `expect(result).toEqual({id: 1, name: 'Alice'})` not `expect(result).toBeTruthy()`.
- Assert error messages, not just error type: `expect(() => fn()).toThrow('Invalid email format')`.
- For async errors: `await expect(fn()).rejects.toThrow(...)`.

---

## Documentation Best Practices

### When to Document
- Public API surface (exported functions, classes, types).
- Non-obvious behavior, side effects, error conditions.
- Configuration options with defaults and valid ranges.
- Architecture decisions (why, not what).

### When NOT to Document
- Self-explanatory code: `getUserById(id)` doesn't need "Gets a user by their ID."
- Implementation details that may change.
- Obvious type information already expressed in TypeScript/Python type hints.

### README Structure
1. What it does (one sentence).
2. Quick start (copy-paste to get running).
3. Usage examples (real code, not pseudo-code).
4. API reference (if library).
5. Configuration (env vars, config files).
6. Architecture (for non-trivial projects).

---

## Performance Best Practices

### Database
- Add indexes for columns in WHERE, JOIN, ORDER BY clauses.
- Use EXPLAIN ANALYZE to verify query plans.
- Paginate all list endpoints (LIMIT + OFFSET or cursor-based).
- Use connection pooling (pg-pool, SQLAlchemy pool, GORM pool).

### Frontend
- Virtualize long lists (react-virtual, vue-virtual-scroller).
- Lazy load below-the-fold components with dynamic imports.
- Memoize expensive computations and stable callbacks.
- Avoid creating new object/array references in render.

### Backend
- Use `Promise.all()` for independent async operations.
- Stream large responses instead of buffering.
- Cache expensive computations and external API responses.
- Set appropriate HTTP cache headers (ETag, Cache-Control).

### General
- Measure before optimizing. Use profilers, not intuition.
- Optimize the hot path first (per-request code, render functions).
- Prefer algorithmic improvements (O(n) vs O(n^2)) over micro-optimizations.

---

## Migration Best Practices

### Before Starting
- Read the official migration guide completely.
- Run codemods if available (they handle the mechanical work).
- Ensure tests pass BEFORE starting migration (baseline).
- Create a branch and commit after each logical step.

### During Migration
- Update one dependency at a time, not all at once.
- Fix type errors and build errors before fixing runtime errors.
- Don't "improve" code while migrating — migrate first, refactor later.
- If a step breaks more than expected, revert and try a smaller step.

### After Migration
- Run the full test suite.
- Check for deprecation warnings (they'll be breaking changes in the next version).
- Update CI/CD configuration if runtime versions changed.
- Document any manual steps for teammates.
