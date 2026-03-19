# API Security Testing Checklist

Comprehensive security testing reference for APIs, covering the OWASP API Security Top 10,
injection attacks, authentication flaws, data exposure, rate limiting, CORS, security headers,
input validation, and penetration testing methodology.

---

## Table of Contents

- [OWASP API Security Top 10 (2023)](#owasp-api-security-top-10-2023)
  - [API1: Broken Object Level Authorization (BOLA)](#api1-broken-object-level-authorization-bola)
  - [API2: Broken Authentication](#api2-broken-authentication)
  - [API3: Broken Object Property Level Authorization](#api3-broken-object-property-level-authorization)
  - [API4: Unrestricted Resource Consumption](#api4-unrestricted-resource-consumption)
  - [API5: Broken Function Level Authorization (BFLA)](#api5-broken-function-level-authorization-bfla)
  - [API6: Unrestricted Access to Sensitive Business Flows](#api6-unrestricted-access-to-sensitive-business-flows)
  - [API7: Server-Side Request Forgery (SSRF)](#api7-server-side-request-forgery-ssrf)
  - [API8: Security Misconfiguration](#api8-security-misconfiguration)
  - [API9: Improper Inventory Management](#api9-improper-inventory-management)
  - [API10: Unsafe Consumption of APIs](#api10-unsafe-consumption-of-apis)
- [Injection Attacks](#injection-attacks)
  - [SQL Injection](#sql-injection)
  - [NoSQL Injection](#nosql-injection)
  - [Command Injection](#command-injection)
  - [LDAP Injection](#ldap-injection)
  - [XSS via API](#xss-via-api)
  - [Header Injection](#header-injection)
- [Input Validation](#input-validation)
- [Output Encoding](#output-encoding)
- [API Gateway Security](#api-gateway-security)
- [Security Headers](#security-headers)
- [Penetration Testing Methodology](#penetration-testing-methodology)
- [Security Audit Report Template](#security-audit-report-template)
- [Remediation Priorities](#remediation-priorities)

---

## OWASP API Security Top 10 (2023)

### API1: Broken Object Level Authorization (BOLA)

**The #1 API vulnerability.** Users can access resources belonging to other users by
manipulating object IDs in API requests.

**How it works:**

```
Normal request (User A views their own order):
GET /api/orders/order-123    → User A's order (OK)

Attack (User A views User B's order):
GET /api/orders/order-456    → User B's order (BOLA!)

The API returns the resource without checking if the authenticated user
is authorized to access it.
```

**Vulnerable code example:**

```typescript
// VULNERABLE — no ownership check
app.get('/api/orders/:orderId', auth, async (req, res) => {
  const order = await db.order.findUnique({
    where: { id: req.params.orderId },
  });

  if (!order) return res.status(404).json({ error: 'Not found' });

  res.json(order); // Returns ANY user's order!
});
```

**Fixed code:**

```typescript
// SECURE — checks ownership
app.get('/api/orders/:orderId', auth, async (req, res) => {
  const order = await db.order.findUnique({
    where: { id: req.params.orderId },
  });

  if (!order) return res.status(404).json({ error: 'Not found' });

  // BOLA prevention: verify the authenticated user owns this resource
  if (order.userId !== req.user.id && req.user.role !== 'ADMIN') {
    return res.status(404).json({ error: 'Not found' });
    // Return 404, not 403, to avoid information disclosure
  }

  res.json(order);
});
```

**Testing checklist:**

```
BOLA Tests:
  □ User A can access their own resources → 200
  □ User A CANNOT access User B's resources → 404
  □ User A CANNOT enumerate resources (sequential IDs) → 404
  □ Admin CAN access any user's resources → 200
  □ Modify resource ID in URL path → 404
  □ Modify resource ID in request body → 404 or ignored
  □ Modify resource ID in query parameters → 404
  □ Try UUIDs of other users' resources → 404
  □ Try predictable ID patterns → 404

BOLA in nested resources:
  □ GET /users/other-user-id/orders → 404
  □ GET /users/me/orders/other-users-order-id → 404
  □ PUT /users/other-user-id/profile → 403 or 404
  □ DELETE /users/other-user-id/data → 403 or 404
```

**Test implementation:**

```typescript
describe('BOLA Prevention', () => {
  let userAToken: string;
  let userBToken: string;
  let userAOrderId: string;
  let userBOrderId: string;

  beforeAll(async () => {
    userAToken = await getAuthToken('userA');
    userBToken = await getAuthToken('userB');

    // Create orders for both users
    const orderA = await authApi(userAToken)
      .post('/api/orders')
      .send({ productId: 'prod-1', quantity: 1 })
      .expect(201);
    userAOrderId = orderA.body.id;

    const orderB = await authApi(userBToken)
      .post('/api/orders')
      .send({ productId: 'prod-2', quantity: 1 })
      .expect(201);
    userBOrderId = orderB.body.id;
  });

  it('User A can view their own order', async () => {
    await authApi(userAToken)
      .get(`/api/orders/${userAOrderId}`)
      .expect(200);
  });

  it('User A CANNOT view User B order', async () => {
    await authApi(userAToken)
      .get(`/api/orders/${userBOrderId}`)
      .expect(404); // NOT 403 — don't reveal the resource exists
  });

  it('User A CANNOT modify User B order', async () => {
    await authApi(userAToken)
      .put(`/api/orders/${userBOrderId}`)
      .send({ status: 'cancelled' })
      .expect(404);
  });

  it('User A CANNOT delete User B order', async () => {
    await authApi(userAToken)
      .delete(`/api/orders/${userBOrderId}`)
      .expect(404);
  });

  it('should not leak existence via different error codes', async () => {
    // Both non-existent and unauthorized should return 404
    const nonExistent = await authApi(userAToken)
      .get('/api/orders/non-existent-id');

    const unauthorized = await authApi(userAToken)
      .get(`/api/orders/${userBOrderId}`);

    // Both should return 404 (not 403 for unauthorized)
    expect(nonExistent.status).toBe(404);
    expect(unauthorized.status).toBe(404);

    // Response bodies should be identical (no info leak)
    expect(nonExistent.body.error).toBe(unauthorized.body.error);
  });
});
```

---

### API2: Broken Authentication

Weak authentication mechanisms that allow attackers to compromise tokens, keys,
or user credentials.

**Common vulnerabilities:**

| Vulnerability | Attack | Prevention |
|--------------|--------|------------|
| No rate limiting on login | Brute force passwords | Rate limit: 5 attempts/min per IP |
| Weak passwords allowed | Dictionary attack | Enforce password complexity |
| Token in URL | Leaked via referrer/logs | Use Authorization header |
| No token expiration | Stolen token valid forever | Short TTL (1 hour) |
| Predictable tokens | Token forging | Use cryptographic randomness |
| No account lockout | Unlimited attempts | Lock after 5 failures for 15 min |
| Password in response | Data exposure | Never return password in any response |
| No MFA support | Credential stuffing | Support TOTP/WebAuthn |
| Insecure password reset | Account takeover | Time-limited, single-use tokens |

**Testing checklist:**

```
Authentication Tests:
  □ Brute force protection — lock after N failed attempts
  □ Rate limiting on login endpoint — 5-10 req/min
  □ Weak password rejection — enforce complexity rules
  □ Password not in any response — not even in user profile
  □ Token expiration enforced — expired tokens rejected
  □ Token rotation on refresh — old refresh tokens invalidated
  □ Logout invalidates tokens — both access and refresh
  □ Session fixation prevention — new session ID after login
  □ Credential stuffing protection — rate limit per IP + account
  □ Password reset token is time-limited and single-use
  □ Email enumeration prevention — same response for valid/invalid emails
  □ JWT algorithm validation — reject "none" algorithm
  □ JWT signature verification — reject tampered tokens
```

**Test implementations:**

```typescript
describe('Authentication Security', () => {
  it('should lock account after 5 failed login attempts', async () => {
    const attempts = Array.from({ length: 6 }, () =>
      request(app)
        .post('/api/auth/login')
        .send({ email: 'target@example.com', password: 'wrong-password' })
    );

    const results = await Promise.all(attempts);

    // First 5 should be 401, 6th should be 423 (Locked)
    const locked = results.filter(r => r.status === 423);
    expect(locked.length).toBeGreaterThan(0);
    expect(locked[0].body.message).toContain('locked');
  });

  it('should not reveal whether email exists', async () => {
    const existingEmail = await request(app)
      .post('/api/auth/login')
      .send({ email: 'real-user@example.com', password: 'wrong' });

    const nonExistentEmail = await request(app)
      .post('/api/auth/login')
      .send({ email: 'nonexistent@example.com', password: 'wrong' });

    // Both should return same status and similar message
    expect(existingEmail.status).toBe(nonExistentEmail.status);
    expect(existingEmail.body.message).toBe(nonExistentEmail.body.message);
  });

  it('should enforce password complexity', async () => {
    const weakPasswords = ['password', '12345678', 'Password1', 'abcdefgh', 'short'];

    for (const password of weakPasswords) {
      const res = await request(app)
        .post('/api/auth/register')
        .send({ email: `test-${Date.now()}@example.com`, password, name: 'Test' });

      expect(res.status).toBe(400);
    }
  });

  it('should never return password in any response', async () => {
    const token = await getAuthToken('user');

    // Check user profile
    const profile = await authApi(token).get('/api/users/me').expect(200);
    expect(profile.body).not.toHaveProperty('password');
    expect(profile.body).not.toHaveProperty('passwordHash');
    expect(profile.body).not.toHaveProperty('hash');
    expect(JSON.stringify(profile.body)).not.toMatch(/\$2[aby]\$/); // bcrypt hash pattern
  });
});
```

---

### API3: Broken Object Property Level Authorization

Users can read or modify object properties they shouldn't have access to.
Combines "Excessive Data Exposure" and "Mass Assignment" from the 2019 list.

**Excessive Data Exposure:**

```typescript
// VULNERABLE — returns all database fields
app.get('/api/users/:id', auth, async (req, res) => {
  const user = await db.user.findUnique({
    where: { id: req.params.id },
  });
  res.json(user); // Includes passwordHash, internalNotes, SSN, etc.!
});

// SECURE — explicit field selection
app.get('/api/users/:id', auth, async (req, res) => {
  const user = await db.user.findUnique({
    where: { id: req.params.id },
    select: {
      id: true,
      email: true,
      name: true,
      role: true,
      avatarUrl: true,
      createdAt: true,
    },
  });
  res.json(user);
});
```

**Mass Assignment:**

```typescript
// VULNERABLE — accepts any field from request body
app.put('/api/users/:id', auth, async (req, res) => {
  const user = await db.user.update({
    where: { id: req.params.id },
    data: req.body, // Attacker can set { role: "ADMIN", isVerified: true }!
  });
  res.json(user);
});

// SECURE — whitelist allowed fields
app.put('/api/users/:id', auth, async (req, res) => {
  const allowedFields = ['name', 'avatarUrl', 'preferences'];
  const data: Record<string, unknown> = {};

  for (const field of allowedFields) {
    if (req.body[field] !== undefined) {
      data[field] = req.body[field];
    }
  }

  const user = await db.user.update({
    where: { id: req.params.id },
    data,
  });
  res.json(user);
});
```

**Testing checklist:**

```
Property-Level Authorization Tests:

Excessive Data Exposure:
  □ User responses don't include password/hash
  □ User responses don't include internal IDs/notes
  □ User responses don't include PII not needed by the consumer
  □ Admin-only fields are not returned to regular users
  □ List endpoints don't return more fields than detail endpoints
  □ Error responses don't include stack traces or SQL queries
  □ Debug information is not exposed in production

Mass Assignment:
  □ Cannot set role via user profile update
  □ Cannot set isAdmin/isVerified via profile update
  □ Cannot set createdAt/updatedAt via request body
  □ Cannot set internal fields (deletedAt, internalNotes)
  □ Cannot set pricing/billing fields on user-facing endpoints
  □ Cannot set ownership fields (userId, ownerId)
  □ Unknown fields in request body are ignored (not stored)
```

**Test implementations:**

```typescript
describe('Property-Level Authorization', () => {
  describe('Excessive Data Exposure', () => {
    it('should not expose sensitive fields in user response', async () => {
      const token = await getAuthToken('user');
      const res = await authApi(token).get('/api/users/me').expect(200);

      const sensitiveFields = [
        'password', 'passwordHash', 'hash', 'salt',
        'ssn', 'socialSecurityNumber',
        'creditCard', 'cardNumber',
        'internalNotes', 'adminNotes',
        'lastLoginIp', 'loginAttempts',
        'resetToken', 'verificationToken',
      ];

      for (const field of sensitiveFields) {
        expect(res.body).not.toHaveProperty(field);
      }
    });

    it('should return different fields for admin vs user', async () => {
      const userToken = await getAuthToken('user');
      const adminToken = await getAuthToken('admin');

      const userView = await authApi(userToken).get('/api/users/user-1').expect(200);
      const adminView = await authApi(adminToken).get('/api/users/user-1').expect(200);

      // Admin might see more fields (lastLoginAt, etc.)
      // But neither should see password
      expect(userView.body).not.toHaveProperty('password');
      expect(adminView.body).not.toHaveProperty('password');
    });
  });

  describe('Mass Assignment', () => {
    it('should not allow role escalation via profile update', async () => {
      const userToken = await getAuthToken('user');

      await authApi(userToken)
        .put('/api/users/me')
        .send({
          name: 'Hacker',
          role: 'ADMIN',        // Attempting privilege escalation
          isAdmin: true,         // Another attempt
          isVerified: true,      // Skip email verification
        })
        .expect(200);

      // Verify role didn't change
      const profile = await authApi(userToken).get('/api/users/me').expect(200);
      expect(profile.body.role).not.toBe('ADMIN');
    });

    it('should not allow setting read-only fields', async () => {
      const token = await getAuthToken('user');

      await authApi(token)
        .put('/api/users/me')
        .send({
          name: 'Updated Name',
          id: 'different-id',                    // Read-only
          email: 'stolen@example.com',           // Should require verification
          createdAt: '2020-01-01T00:00:00Z',     // Read-only
        })
        .expect(200);

      const profile = await authApi(token).get('/api/users/me').expect(200);
      expect(profile.body.id).not.toBe('different-id');
      expect(profile.body.createdAt).not.toBe('2020-01-01T00:00:00Z');
    });

    it('should ignore unknown fields in request body', async () => {
      const token = await getAuthToken('admin');

      const res = await authApi(token)
        .post('/api/products')
        .send({
          name: 'Test Product',
          price: 29.99,
          category: 'electronics',
          _internalCost: 5.00,     // Unknown field — should be ignored
          __proto__: { admin: true }, // Prototype pollution attempt
          constructor: { name: 'Object' }, // Another pollution attempt
        })
        .expect(201);

      expect(res.body).not.toHaveProperty('_internalCost');
    });
  });
});
```

---

### API4: Unrestricted Resource Consumption

No limits on the size or number of resources a client can request, leading to
denial of service, excessive costs, or performance degradation.

**Common vulnerabilities:**

| Vulnerability | Example | Prevention |
|--------------|---------|------------|
| No rate limiting | 1M requests/min from one client | Rate limit by IP/token/client |
| No pagination limits | `GET /users?pageSize=1000000` | Cap pageSize (e.g., max 100) |
| No request body size limit | 10GB JSON payload | Set max body size (e.g., 1MB) |
| No file upload size limit | Upload 100GB file | Set max file size (e.g., 10MB) |
| No query complexity limit | Deeply nested GraphQL | Limit query depth/complexity |
| No execution timeout | Query that takes 10 minutes | Set request timeout (30s) |
| No resource creation limit | Create 1M records in one request | Limit batch sizes |
| Expensive operations unbounded | RegExp DoS (ReDoS) | Use safe regex or timeout |

**Testing checklist:**

```
Resource Consumption Tests:
  □ Rate limiting enforced — returns 429 with Retry-After
  □ Pagination max size enforced — caps at max pageSize
  □ Request body size limited — returns 413 for large bodies
  □ File upload size limited — returns 413 for large files
  □ Query timeout enforced — returns 504 for slow queries
  □ Batch operation size limited — rejects over-large batches
  □ GraphQL depth limit — rejects deeply nested queries
  □ GraphQL complexity limit — rejects expensive queries
  □ No ReDoS in input validation — regex patterns are safe
  □ Concurrent request limit per user
```

---

### API5: Broken Function Level Authorization (BFLA)

Users can access API endpoints or functions intended for a different role or permission level.

**How it works:**

```
Regular user discovers admin endpoint:
  GET  /api/users          → Regular user list (OK)
  GET  /api/admin/users    → Admin user list with full details (BFLA!)
  POST /api/admin/users    → Create user as admin (BFLA!)
  DELETE /api/users/123    → Delete any user (BFLA!)

The endpoint exists and is reachable but lacks proper role checks.
```

**Testing checklist:**

```
BFLA Tests:
  □ Regular user CANNOT access admin endpoints
  □ Regular user CANNOT access moderator endpoints
  □ Unauthenticated user CANNOT access any protected endpoint
  □ Admin endpoints return 403 (not 401) for authenticated non-admins
  □ HTTP method restrictions enforced (GET-only users can't POST)
  □ Debug endpoints not accessible in production
  □ Internal endpoints not exposed publicly
  □ Role changes take effect immediately (not cached)
  □ Horizontal BFLA: User A can't use User B's admin permissions
```

**Test implementations:**

```typescript
describe('Function-Level Authorization (BFLA)', () => {
  let userToken: string;
  let adminToken: string;

  beforeAll(async () => {
    userToken = await getAuthToken('user');
    adminToken = await getAuthToken('admin');
  });

  const adminEndpoints = [
    { method: 'get', path: '/api/admin/users' },
    { method: 'get', path: '/api/admin/orders' },
    { method: 'get', path: '/api/admin/analytics' },
    { method: 'post', path: '/api/admin/users' },
    { method: 'delete', path: '/api/admin/users/any-id' },
    { method: 'put', path: '/api/admin/settings' },
  ];

  for (const endpoint of adminEndpoints) {
    it(`should deny ${endpoint.method.toUpperCase()} ${endpoint.path} for regular users`, async () => {
      const res = await (authApi(userToken) as any)[endpoint.method](endpoint.path);
      expect(res.status).toBe(403);
    });

    it(`should allow ${endpoint.method.toUpperCase()} ${endpoint.path} for admins`, async () => {
      const res = await (authApi(adminToken) as any)[endpoint.method](endpoint.path);
      expect([200, 201, 204, 404]).toContain(res.status); // Not 403
    });
  }

  it('should not expose admin endpoints via different HTTP methods', async () => {
    // Some APIs accidentally expose admin functionality via PATCH/PUT
    // when only GET was intended
    const res = await authApi(userToken)
      .put('/api/users/admin-user-id')
      .send({ role: 'ADMIN' });

    expect([403, 404]).toContain(res.status);
  });

  it('should check permissions on every request (not cached)', async () => {
    // Demote admin to user, verify immediate effect
    // (This tests that auth/role is checked per-request, not cached)
    const tempAdmin = await createTempUser('admin');

    // Should work as admin
    await authApi(tempAdmin.token)
      .get('/api/admin/users')
      .expect(200);

    // Demote to regular user
    await db.user.update({
      where: { id: tempAdmin.id },
      data: { role: 'USER' },
    });

    // Should immediately fail
    await authApi(tempAdmin.token)
      .get('/api/admin/users')
      .expect(403);
  });
});
```

---

### API6: Unrestricted Access to Sensitive Business Flows

Attackers exploit business logic by automating flows intended for humans (buying
limited items, creating fake accounts, scraping prices, spamming reviews).

**Common targets:**

| Flow | Attack | Prevention |
|------|--------|------------|
| Purchase | Bot buying limited-edition items | CAPTCHA, purchase limits, queue system |
| Registration | Mass fake account creation | Email verification, CAPTCHA, rate limiting |
| Reviews | Fake review spam | Purchase verification, rate limiting |
| Password reset | Enumerate valid emails | Constant-time responses, rate limiting |
| Coupon redemption | Automated coupon abuse | One-per-user, per-order limits |
| Content creation | Spam content | Anti-spam, content moderation |

**Testing checklist:**

```
Business Flow Tests:
  □ Purchase flow has per-user limits
  □ Registration requires email verification
  □ Review posting requires verified purchase
  □ Coupon codes are single-use or per-user
  □ Password reset doesn't reveal email existence
  □ Referral system prevents self-referral
  □ Free trial prevents repeated signups (same email, IP, device)
```

---

### API7: Server-Side Request Forgery (SSRF)

The API accepts a URL from the client and makes a server-side request to it, allowing
attackers to access internal resources.

**Vulnerable code:**

```typescript
// VULNERABLE — fetches arbitrary URLs
app.post('/api/fetch-url', async (req, res) => {
  const { url } = req.body;
  const response = await fetch(url); // Attacker: url = "http://169.254.169.254/metadata"
  res.json(await response.json());
});
```

**Attacks:**

```
# Access cloud metadata (AWS)
POST /api/fetch-url
{ "url": "http://169.254.169.254/latest/meta-data/iam/security-credentials/" }

# Access internal services
POST /api/fetch-url
{ "url": "http://internal-admin.local:8080/admin/users" }

# Port scan internal network
POST /api/fetch-url
{ "url": "http://10.0.0.1:22" }  → SSH open
{ "url": "http://10.0.0.1:3306" } → MySQL open
```

**Prevention:**

```typescript
import { URL } from 'url';
import { isIP } from 'net';
import dns from 'dns/promises';

// SECURE — URL validation and allowlisting
async function validateUrl(input: string): Promise<boolean> {
  let parsed: URL;
  try {
    parsed = new URL(input);
  } catch {
    return false;
  }

  // Only allow HTTPS
  if (parsed.protocol !== 'https:') return false;

  // Block private IP ranges
  const hostname = parsed.hostname;

  // Resolve DNS to check for SSRF via DNS rebinding
  const addresses = await dns.resolve(hostname);
  for (const addr of addresses) {
    if (isPrivateIP(addr)) return false;
  }

  // Allowlist domains (if possible)
  const allowedDomains = ['cdn.example.com', 'images.example.com'];
  if (!allowedDomains.some(d => hostname.endsWith(d))) {
    return false;
  }

  return true;
}

function isPrivateIP(ip: string): boolean {
  const parts = ip.split('.').map(Number);
  return (
    parts[0] === 10 ||                        // 10.0.0.0/8
    (parts[0] === 172 && parts[1] >= 16 && parts[1] <= 31) || // 172.16.0.0/12
    (parts[0] === 192 && parts[1] === 168) || // 192.168.0.0/16
    parts[0] === 127 ||                        // 127.0.0.0/8
    (parts[0] === 169 && parts[1] === 254) || // 169.254.0.0/16 (link-local / metadata)
    ip === '0.0.0.0'
  );
}
```

**Testing checklist:**

```
SSRF Tests:
  □ Cannot access cloud metadata endpoint (169.254.169.254)
  □ Cannot access localhost (127.0.0.1, localhost, [::1])
  □ Cannot access private IPs (10.x, 172.16-31.x, 192.168.x)
  □ Cannot access internal hostnames
  □ Cannot use non-HTTP schemes (file://, gopher://, dict://)
  □ DNS rebinding protection (resolve before connecting)
  □ Redirect following doesn't bypass validation
  □ Port restrictions (only 80, 443)
  □ Response size limits (prevent large downloads)
```

---

### API8: Security Misconfiguration

Insecure default configurations, missing security hardening, unnecessary features enabled.

**Testing checklist:**

```
Configuration Security Tests:

Transport:
  □ HTTPS enforced — HTTP redirects to HTTPS
  □ TLS 1.2+ required — TLS 1.0/1.1 rejected
  □ Strong cipher suites only
  □ HSTS header present with sufficient max-age

Headers:
  □ X-Content-Type-Options: nosniff
  □ X-Frame-Options: DENY or SAMEORIGIN
  □ Content-Security-Policy present
  □ No X-Powered-By header
  □ No Server version disclosure
  □ Referrer-Policy set

CORS:
  □ Access-Control-Allow-Origin is NOT * (with credentials)
  □ Only trusted origins are allowed
  □ Access-Control-Allow-Methods is restrictive
  □ Access-Control-Allow-Headers is restrictive
  □ Access-Control-Max-Age is set

Error Handling:
  □ Stack traces not exposed in production
  □ SQL queries not exposed in error messages
  □ File paths not exposed in errors
  □ Internal service names not exposed
  □ Debug mode disabled in production
  □ Verbose error messages only in development

Endpoints:
  □ Debug/test endpoints disabled in production
  □ Default admin credentials changed
  □ API documentation not exposed in production (or protected)
  □ GraphQL introspection disabled in production
  □ TRACE method disabled
  □ Unnecessary HTTP methods disabled
```

---

### API9: Improper Inventory Management

Unmanaged API versions, undocumented endpoints, and unmonitored API inventory.

**Testing checklist:**

```
Inventory Management Tests:
  □ Old API versions deprecated with sunset dates
  □ Beta endpoints clearly marked and access-controlled
  □ No undocumented endpoints discoverable via fuzzing
  □ API documentation matches actual implementation
  □ Debug endpoints not accessible in production
  □ Unused endpoints are removed, not just hidden
  □ All APIs behind API gateway with logging
  □ Third-party API integrations audited
```

---

### API10: Unsafe Consumption of APIs

The API trusts data from third-party APIs without validation, inheriting their
vulnerabilities.

**Vulnerable pattern:**

```typescript
// VULNERABLE — trusting third-party response
app.get('/api/enriched-user/:id', async (req, res) => {
  const user = await db.user.findUnique({ where: { id: req.params.id } });

  // Fetch from third-party API
  const enrichment = await fetch(`https://api.thirdparty.com/users/${user.email}`);
  const enrichedData = await enrichment.json();

  // Directly merge third-party data (DANGEROUS!)
  res.json({ ...user, ...enrichedData });
  // Third-party could inject: { role: "admin", isVerified: true }
});
```

**Testing checklist:**

```
Third-Party API Security:
  □ Third-party responses are validated against a schema
  □ Third-party data is sanitized before use/storage
  □ Third-party API failures are handled gracefully (circuit breaker)
  □ Third-party URLs are validated (no SSRF via redirects)
  □ Third-party data doesn't override internal fields
  □ Timeouts set for third-party API calls
  □ Third-party API keys are not exposed to clients
  □ Webhook payloads from third parties are signature-verified
```

---

## Injection Attacks

### SQL Injection

**Vulnerable code:**

```typescript
// VULNERABLE — string concatenation
app.get('/api/users', async (req, res) => {
  const query = `SELECT * FROM users WHERE name = '${req.query.name}'`;
  const users = await db.$queryRawUnsafe(query);
  res.json(users);
});

// Attack: GET /api/users?name=' OR '1'='1
// Becomes: SELECT * FROM users WHERE name = '' OR '1'='1'
```

**Prevention — always use parameterized queries:**

```typescript
// SECURE — parameterized query
app.get('/api/users', async (req, res) => {
  const users = await db.user.findMany({
    where: { name: req.query.name as string },
  });
  res.json(users);
});

// Or with raw SQL (parameterized)
const users = await db.$queryRaw`SELECT * FROM users WHERE name = ${req.query.name}`;
```

**Testing payloads:**

```
SQL Injection Test Payloads:

Basic:
  ' OR '1'='1
  ' OR '1'='1' --
  ' OR '1'='1' /*
  '; DROP TABLE users; --
  ' UNION SELECT 1,2,3 --

Blind SQL Injection:
  ' AND 1=1 --     (true — normal response)
  ' AND 1=2 --     (false — different response)
  ' AND SLEEP(5) -- (time-based)

Authentication Bypass:
  admin' --
  admin' #
  ' OR 1=1 LIMIT 1 --

Error-Based:
  ' AND (SELECT 1 FROM (SELECT COUNT(*),CONCAT(...) x FROM information_schema.tables GROUP BY x) a) --
```

### NoSQL Injection

```typescript
// VULNERABLE — MongoDB operator injection
app.post('/api/auth/login', async (req, res) => {
  const user = await db.collection('users').findOne({
    email: req.body.email,
    password: req.body.password,
  });
});

// Attack:
// POST /api/auth/login
// { "email": "admin@example.com", "password": { "$gt": "" } }
// The $gt operator matches any non-empty password!
```

**Prevention:**

```typescript
// SECURE — validate types, sanitize input
app.post('/api/auth/login', async (req, res) => {
  const { email, password } = req.body;

  // Type check — reject non-string values
  if (typeof email !== 'string' || typeof password !== 'string') {
    return res.status(400).json({ error: 'Invalid input types' });
  }

  const user = await db.collection('users').findOne({
    email: email,
    // Don't compare passwords directly — use bcrypt
  });

  if (!user || !await bcrypt.compare(password, user.passwordHash)) {
    return res.status(401).json({ error: 'Invalid credentials' });
  }
});
```

### Command Injection

```typescript
// VULNERABLE — shell command with user input
app.post('/api/tools/ping', async (req, res) => {
  const { host } = req.body;
  const result = execSync(`ping -c 4 ${host}`); // DANGEROUS!
  res.json({ output: result.toString() });
});

// Attack: { "host": "example.com; cat /etc/passwd" }
// Executes: ping -c 4 example.com; cat /etc/passwd
```

**Prevention:**

```typescript
// SECURE — use array arguments (no shell interpretation)
import { execFile } from 'child_process';

app.post('/api/tools/ping', async (req, res) => {
  const { host } = req.body;

  // Validate input
  if (!/^[a-zA-Z0-9.\-]+$/.test(host)) {
    return res.status(400).json({ error: 'Invalid hostname' });
  }

  execFile('ping', ['-c', '4', host], (error, stdout) => {
    if (error) return res.status(500).json({ error: 'Ping failed' });
    res.json({ output: stdout });
  });
});
```

### XSS via API

APIs that store user input and return it unsanitized can enable Stored XSS.

```typescript
// VULNERABLE — stores and returns unsanitized HTML
app.post('/api/comments', async (req, res) => {
  const comment = await db.comment.create({
    data: { text: req.body.text }, // <script>alert('XSS')</script>
  });
  res.json(comment);
});

// When the frontend renders this comment, the script executes
```

**Prevention:**

```typescript
import DOMPurify from 'isomorphic-dompurify';

// SECURE — sanitize on input
app.post('/api/comments', async (req, res) => {
  const sanitizedText = DOMPurify.sanitize(req.body.text, {
    ALLOWED_TAGS: [], // Strip all HTML tags
    ALLOWED_ATTR: [],
  });

  const comment = await db.comment.create({
    data: { text: sanitizedText },
  });
  res.json(comment);
});
```

### Header Injection

```typescript
// VULNERABLE — user input in response headers
app.get('/api/redirect', (req, res) => {
  const { url } = req.query;
  res.redirect(url as string); // Attacker: url = "http://evil.com\r\nSet-Cookie: session=stolen"
});

// Prevention: Validate URLs, use allowlists, encode header values
```

---

## Input Validation

### Validation Checklist

```
For every input field, validate:

Type Safety:
  □ Correct type (string, number, boolean, array, object)
  □ Reject unexpected types (e.g., object when string expected)
  □ Handle null vs undefined vs empty string

String Validation:
  □ Minimum length (where applicable)
  □ Maximum length (always — prevent DoS)
  □ Allowed character set (whitelist preferred)
  □ Format validation (email, URL, UUID, phone, etc.)
  □ No control characters (null bytes, newlines in single-line fields)
  □ UTF-8 encoding is valid

Number Validation:
  □ Minimum value
  □ Maximum value
  □ Integer vs float
  □ Positive/non-negative (where applicable)
  □ NaN, Infinity, -Infinity rejection

Array Validation:
  □ Maximum length (prevent DoS)
  □ Unique items (where applicable)
  □ Each item validates individually

Object Validation:
  □ Known properties only (reject unknown/extra fields)
  □ Required properties present
  □ No prototype pollution (__proto__, constructor, prototype)

File Validation:
  □ File size limit
  □ File type validation (MIME type + magic bytes, not just extension)
  □ Filename sanitization (no path traversal)
  □ Virus scanning (for uploaded files)
```

---

## Output Encoding

### What to Encode

```
Output Encoding Rules:
  □ HTML content → HTML entity encode
  □ JSON responses → proper JSON serialization (no raw string concat)
  □ URLs → URL encode user-supplied values
  □ HTTP headers → encode/validate header values
  □ SQL → parameterized queries (not encoding!)
  □ Command-line → use array arguments, not shell interpolation
  □ XML → XML entity encode
  □ LDAP → LDAP-specific encoding
```

---

## API Gateway Security

### Gateway Security Checklist

```
API Gateway Configuration:
  □ All APIs routed through gateway (no direct backend access)
  □ Authentication enforced at gateway level
  □ Rate limiting at gateway level
  □ Request size limits enforced
  □ TLS termination with strong ciphers
  □ IP allowlisting/blocklisting
  □ Geographic restrictions (if applicable)
  □ WAF rules (OWASP CRS)
  □ DDoS protection
  □ Request logging and monitoring
  □ API versioning enforced
  □ Health check endpoints not exposed publicly
  □ Admin endpoints on separate network/port
```

---

## Security Headers

### Complete Security Headers Reference

```http
# Transport Security
Strict-Transport-Security: max-age=31536000; includeSubDomains; preload

# Content Security
Content-Security-Policy: default-src 'none'; frame-ancestors 'none'
X-Content-Type-Options: nosniff
X-Frame-Options: DENY

# Information Leakage Prevention
Referrer-Policy: strict-origin-when-cross-origin
Permissions-Policy: camera=(), microphone=(), geolocation=()

# Cache Control (for sensitive responses)
Cache-Control: no-store, no-cache, must-revalidate
Pragma: no-cache

# Remove revealing headers
# X-Powered-By: (remove entirely)
# Server: (remove or set to generic value)
```

### Implementation

```typescript
// Express security headers middleware
import helmet from 'helmet';

app.use(helmet({
  contentSecurityPolicy: {
    directives: {
      defaultSrc: ["'none'"],
      frameAncestors: ["'none'"],
    },
  },
  crossOriginEmbedderPolicy: true,
  crossOriginOpenerPolicy: true,
  crossOriginResourcePolicy: { policy: 'same-origin' },
  dnsPrefetchControl: true,
  frameguard: { action: 'deny' },
  hsts: { maxAge: 31536000, includeSubDomains: true, preload: true },
  ieNoOpen: true,
  noSniff: true,
  referrerPolicy: { policy: 'strict-origin-when-cross-origin' },
  xssFilter: false, // Disabled — CSP is better, X-XSS-Protection can introduce issues
}));

// Remove X-Powered-By
app.disable('x-powered-by');
```

---

## Penetration Testing Methodology

### Phase 1: Reconnaissance

```
Information Gathering:
  □ Discover all API endpoints (documentation, code, fuzzing)
  □ Identify authentication mechanisms
  □ Map authorization model (roles, permissions)
  □ Identify input vectors (path params, query params, headers, body)
  □ Identify output formats (JSON, XML, HTML)
  □ Find API documentation (Swagger, Postman collections)
  □ Enumerate technologies (headers, error messages, fingerprinting)
  □ Find related APIs (v1, v2, beta, internal, admin)
```

### Phase 2: Authentication Testing

```
Authentication Attack Surface:
  □ Brute force login (with rate limit testing)
  □ Default/weak credentials
  □ Password complexity enforcement
  □ Token expiration and refresh
  □ JWT attacks (none alg, key confusion, claim tampering)
  □ Session management (fixation, expiry, logout)
  □ OAuth flow vulnerabilities
  □ API key security (leaked keys, no rotation)
  □ MFA bypass attempts
```

### Phase 3: Authorization Testing

```
Authorization Attack Surface:
  □ BOLA — access other users' resources
  □ BFLA — access admin/elevated endpoints
  □ Mass assignment — set unauthorized fields
  □ Privilege escalation — change own role
  □ Horizontal — User A accessing User B's data
  □ Vertical — User accessing Admin functions
```

### Phase 4: Input Testing

```
Input Attack Surface:
  □ SQL injection (all parameters)
  □ NoSQL injection (all parameters)
  □ Command injection (if server executes commands)
  □ XSS (stored, reflected, DOM)
  □ SSRF (any URL/webhook parameters)
  □ Path traversal (file operations)
  □ XML External Entity (if XML accepted)
  □ Template injection (if templates used)
  □ Header injection (CRLF)
  □ Integer overflow/underflow
  □ Buffer overflow (in native code)
  □ Regex DoS (if user input in regex)
```

### Phase 5: Business Logic Testing

```
Business Logic:
  □ Price manipulation (negative quantities, modified prices)
  □ Workflow bypass (skip required steps)
  □ Race conditions (double-spending, duplicate operations)
  □ Limit bypass (free trial abuse, quota circumvention)
  □ Feature abuse (automation of manual-intended flows)
```

### Phase 6: Configuration Testing

```
Configuration:
  □ HTTPS enforcement
  □ Security headers present
  □ CORS properly configured
  □ Error handling doesn't leak info
  □ Debug mode disabled
  □ Unnecessary methods disabled
  □ Default paths removed (/admin, /debug, /test)
  □ GraphQL introspection disabled
```

---

## Security Audit Report Template

```
API SECURITY AUDIT REPORT
═════════════════════════

1. EXECUTIVE SUMMARY
━━━━━━━━━━━━━━━━━━━

Scope: [API name and version]
Date: [Audit date]
Auditor: [Agent/person]
Environment: [Dev/Staging/Production]

Finding Summary:
  Critical:  X findings
  High:      X findings
  Medium:    X findings
  Low:       X findings
  Info:      X findings

Overall Risk Level: [Critical/High/Medium/Low]

2. FINDINGS
━━━━━━━━━

[For each finding:]

FINDING #1: [Title]
  Severity:    Critical/High/Medium/Low/Info
  Category:    OWASP API [number]
  Endpoint:    [Method] [Path]
  Description: [What the vulnerability is]
  Evidence:    [Request/response showing the issue]
  Impact:      [What an attacker could do]
  Remediation: [How to fix it]
  Priority:    Fix immediately / Fix this sprint / Fix next sprint / Backlog

3. METHODOLOGY
━━━━━━━━━━━━━

[Description of testing approach, tools used, scope limitations]

4. RECOMMENDATIONS
━━━━━━━━━━━━━━━━━

Immediate:
  □ [Fix critical and high findings]

Short-term:
  □ [Fix medium findings, implement missing controls]

Long-term:
  □ [Security architecture improvements, monitoring, training]
```

---

## Remediation Priorities

### Priority Matrix

| Severity | Fix Timeline | Examples |
|----------|-------------|---------|
| **Critical** | Immediately (< 24 hours) | SQL injection, auth bypass, data breach |
| **High** | This sprint (< 1 week) | BOLA, BFLA, SSRF, missing rate limiting |
| **Medium** | Next sprint (< 1 month) | XSS, weak passwords, missing headers, verbose errors |
| **Low** | Backlog | Missing HSTS preload, old TLS ciphers, info disclosure |
| **Info** | Documentation | Best practice recommendations, defense-in-depth suggestions |

### Quick Wins

These security improvements can be implemented in under an hour:

```
Quick Security Wins:
  □ Add helmet middleware (security headers)              — 5 minutes
  □ Remove X-Powered-By header                           — 1 minute
  □ Set request body size limit (1MB)                     — 5 minutes
  □ Add rate limiting to auth endpoints                   — 15 minutes
  □ Disable TRACE method                                  — 5 minutes
  □ Set Content-Type on all responses                     — 5 minutes
  □ Add CORS configuration (specific origins)             — 10 minutes
  □ Strip stack traces from production errors              — 10 minutes
  □ Disable GraphQL introspection in production           — 5 minutes
  □ Add request ID to error responses (for tracking)      — 10 minutes
```
