# OWASP Top 10 (2021) — Complete Technical Reference

The OWASP Top 10 is the industry-standard awareness document for web application security. This reference covers each category with deep technical detail: description, real-world attack scenarios, detection techniques, prevention controls, code examples in multiple languages, and testing methodology.

---

## A01:2021 — Broken Access Control

**Moved up from #5 in 2017. The most common vulnerability category.**

### Description

Access control enforces policy such that users cannot act outside of their intended permissions. Failures typically lead to unauthorized information disclosure, modification, or destruction of data, or performing a business function outside of the user's limits.

### CWEs Mapped

- CWE-200: Exposure of Sensitive Information to an Unauthorized Actor
- CWE-201: Insertion of Sensitive Information Into Sent Data
- CWE-352: Cross-Site Request Forgery
- CWE-284: Improper Access Control
- CWE-285: Improper Authorization
- CWE-639: Authorization Bypass Through User-Controlled Key (IDOR)
- CWE-862: Missing Authorization
- CWE-863: Incorrect Authorization

### Attack Scenarios

**Scenario 1: IDOR — Accessing other users' data**

The application uses unverified data in a SQL call to access account information:
```
https://example.com/api/accounts?id=12345
```

An attacker simply modifies the `id` parameter to access any account:
```
https://example.com/api/accounts?id=12346
```

Vulnerable code (Node.js):
```javascript
// VULNERABLE — no ownership check
app.get('/api/accounts', async (req, res) => {
  const account = await Account.findByPk(req.query.id);
  res.json(account);
});
```

**Scenario 2: Force browsing to admin pages**

An attacker navigates to admin URLs:
```
https://example.com/admin/deleteUser?id=123
```

If the application only hides the admin menu but doesn't check authorization server-side, the admin function executes for a regular user.

**Scenario 3: Metadata manipulation**

An attacker modifies the JWT token's role claim:
```json
// Original token payload
{ "sub": "user123", "role": "user", "exp": 1700000000 }

// Modified token payload
{ "sub": "user123", "role": "admin", "exp": 1700000000 }
```

If the application uses a weak or predictable signing key, or if the `none` algorithm is accepted, the attacker gains admin access.

**Scenario 4: CORS misconfiguration**

```
Access-Control-Allow-Origin: *
Access-Control-Allow-Credentials: true
```

This allows any website to make authenticated requests on behalf of the user, accessing their data.

### Detection Techniques

```
Manual testing:
1. Access a resource as User A
2. Copy the request (including resource ID)
3. Replay the request as User B
4. If User B can access User A's resource → IDOR

Automated patterns:
- Grep for routes without auth middleware
- Grep for findById without userId filter
- Grep for admin routes accessible without role check
- Test for path traversal: /../admin/, /./admin/
- Test for HTTP method tampering: change POST to PUT/DELETE
- Test for parameter pollution: ?id=1&id=2
```

### Prevention Controls

```javascript
// 1. Default deny — require explicit authorization
function requireAuth(req, res, next) {
  if (!req.user) return res.status(401).json({ error: 'Unauthorized' });
  next();
}

function requireRole(...roles) {
  return (req, res, next) => {
    if (!roles.includes(req.user.role)) {
      return res.status(403).json({ error: 'Forbidden' });
    }
    next();
  };
}

// 2. Object-level authorization — always scope to user
app.get('/api/orders/:id', requireAuth, async (req, res) => {
  const order = await Order.findOne({
    where: { id: req.params.id, userId: req.user.id }
  });
  if (!order) return res.status(404).json({ error: 'Not found' });
  res.json(order);
});

// 3. Rate limiting to prevent enumeration
import rateLimit from 'express-rate-limit';
const apiLimiter = rateLimit({ windowMs: 60000, max: 100 });
app.use('/api/', apiLimiter);

// 4. Disable directory listing
app.use(express.static('public', { dotfiles: 'deny', index: false }));

// 5. Disable CORS for sensitive endpoints
app.use(cors({
  origin: ['https://myapp.com'],
  credentials: true,
  methods: ['GET', 'POST', 'PUT', 'DELETE'],
}));
```

```python
# Django — object-level permissions
from rest_framework import permissions

class IsOwner(permissions.BasePermission):
    def has_object_permission(self, request, view, obj):
        return obj.owner == request.user

class OrderViewSet(viewsets.ModelViewSet):
    permission_classes = [permissions.IsAuthenticated, IsOwner]

    def get_queryset(self):
        # Always scope to authenticated user
        return Order.objects.filter(owner=self.request.user)
```

```java
// Spring Boot — method-level security
@PreAuthorize("hasRole('ADMIN')")
@DeleteMapping("/api/admin/users/{id}")
public ResponseEntity<?> deleteUser(@PathVariable Long id) {
    userService.deleteUser(id);
    return ResponseEntity.ok().build();
}

@GetMapping("/api/orders/{id}")
public ResponseEntity<Order> getOrder(@PathVariable Long id, Authentication auth) {
    Order order = orderService.findByIdAndUserId(id, auth.getName());
    if (order == null) return ResponseEntity.notFound().build();
    return ResponseEntity.ok(order);
}
```

### Testing Methodology

```
1. Map all endpoints and their required roles
2. For each endpoint, test:
   - Access without authentication
   - Access with wrong role (user → admin endpoints)
   - Access to other users' resources (change IDs)
   - Access via direct URL (bypassing UI navigation)
   - Access after logout (replay old session tokens)
   - HTTP method switching (GET → DELETE)
   - URL manipulation (path traversal, encoding)
3. Verify CORS configuration
4. Verify CSRF protection on state-changing endpoints
```

---

## A02:2021 — Cryptographic Failures

**Previously "Sensitive Data Exposure" — renamed to focus on the root cause.**

### Description

Failures related to cryptography which often lead to sensitive data exposure. This includes: not encrypting sensitive data, using weak algorithms, using default/hard-coded keys, not enforcing encryption, and improper certificate validation.

### CWEs Mapped

- CWE-259: Use of Hard-coded Password
- CWE-261: Weak Encoding for Password
- CWE-296: Improper Following of a Certificate's Chain of Trust
- CWE-310: Cryptographic Issues
- CWE-319: Cleartext Transmission of Sensitive Information
- CWE-321: Use of Hard-coded Cryptographic Key
- CWE-327: Use of a Broken or Risky Cryptographic Algorithm
- CWE-328: Use of Weak Hash
- CWE-330: Use of Insufficiently Random Values
- CWE-331: Insufficient Entropy
- CWE-335: Incorrect Usage of Seeds in Pseudo-Random Number Generator

### Attack Scenarios

**Scenario 1: Weak password hashing**

An application stores passwords using unsalted SHA-256:
```python
# VULNERABLE — fast hash, no salt
password_hash = hashlib.sha256(password.encode()).hexdigest()
```

An attacker who gains database access can crack passwords using rainbow tables or GPU-accelerated brute force. SHA-256 can compute billions of hashes per second on modern GPUs.

**Scenario 2: Missing TLS**

A financial application transmits data over HTTP. An attacker on the same network (coffee shop WiFi, compromised router) can intercept the traffic:
```
GET http://bank.example.com/api/account/balance HTTP/1.1
Cookie: session=abc123def456
```

All session cookies and account data are visible in plaintext.

**Scenario 3: Hardcoded encryption key**

```javascript
// VULNERABLE — key in source code
const ENCRYPTION_KEY = 'MySecretKey12345MySecretKey12345'; // 256-bit
const cipher = crypto.createCipheriv('aes-256-cbc', ENCRYPTION_KEY, iv);
```

Anyone with access to the source code (including git history) can decrypt all data.

### Prevention Controls

```javascript
// Node.js — proper encryption with AES-256-GCM
import crypto from 'crypto';

const ALGORITHM = 'aes-256-gcm';
const KEY = Buffer.from(process.env.ENCRYPTION_KEY, 'hex'); // 32 bytes from env
const IV_LENGTH = 12; // 96 bits for GCM
const AUTH_TAG_LENGTH = 16; // 128 bits

function encrypt(plaintext) {
  const iv = crypto.randomBytes(IV_LENGTH);
  const cipher = crypto.createCipheriv(ALGORITHM, KEY, iv, { authTagLength: AUTH_TAG_LENGTH });
  let encrypted = cipher.update(plaintext, 'utf8');
  encrypted = Buffer.concat([encrypted, cipher.final()]);
  const authTag = cipher.getAuthTag();
  // Return IV + AuthTag + Ciphertext (all needed for decryption)
  return Buffer.concat([iv, authTag, encrypted]).toString('base64');
}

function decrypt(encryptedBase64) {
  const data = Buffer.from(encryptedBase64, 'base64');
  const iv = data.subarray(0, IV_LENGTH);
  const authTag = data.subarray(IV_LENGTH, IV_LENGTH + AUTH_TAG_LENGTH);
  const ciphertext = data.subarray(IV_LENGTH + AUTH_TAG_LENGTH);
  const decipher = crypto.createDecipheriv(ALGORITHM, KEY, iv, { authTagLength: AUTH_TAG_LENGTH });
  decipher.setAuthTag(authTag);
  let decrypted = decipher.update(ciphertext);
  decrypted = Buffer.concat([decrypted, decipher.final()]);
  return decrypted.toString('utf8');
}
```

```python
# Python — proper password hashing with argon2
from argon2 import PasswordHasher

ph = PasswordHasher(
    time_cost=2,        # iterations
    memory_cost=19456,  # 19 MiB
    parallelism=1,
    hash_len=32,
    salt_len=16,
)

# Hash
hashed = ph.hash(password)

# Verify
try:
    ph.verify(hashed, password)
    if ph.check_needs_rehash(hashed):
        # Re-hash with current parameters if they've changed
        new_hash = ph.hash(password)
except argon2.exceptions.VerifyMismatchError:
    # Invalid password
    pass
```

---

## A03:2021 — Injection

### Description

An application is vulnerable to injection when user-supplied data is not validated, filtered, or sanitized; dynamic queries or non-parameterized calls without context-aware escaping are used directly in the interpreter; or hostile data is used within ORM search parameters to extract additional sensitive records.

### CWEs Mapped

- CWE-77: Command Injection
- CWE-78: OS Command Injection
- CWE-79: Cross-site Scripting (XSS)
- CWE-89: SQL Injection
- CWE-90: LDAP Injection
- CWE-91: XML Injection (Blind XPath Injection)
- CWE-564: SQL Injection: Hibernate
- CWE-917: Expression Language Injection

### Types of Injection

**SQL Injection**: User input in SQL queries
```javascript
// VULNERABLE
const query = `SELECT * FROM users WHERE name = '${req.body.name}'`;

// SECURE
const query = 'SELECT * FROM users WHERE name = $1';
const result = await db.query(query, [req.body.name]);
```

**NoSQL Injection**: User input in NoSQL queries
```javascript
// VULNERABLE — MongoDB operator injection
const user = await User.findOne({ username: req.body.username, password: req.body.password });
// Attacker sends: { "username": "admin", "password": { "$ne": "" } }
// This matches any non-empty password → authentication bypass

// SECURE — validate types
const username = String(req.body.username);
const password = String(req.body.password);
const user = await User.findOne({ username, password: await hash(password) });
```

**Command Injection**: User input in shell commands
```javascript
// VULNERABLE
exec(`nslookup ${req.body.domain}`);

// SECURE
execFile('nslookup', [req.body.domain]); // No shell interpretation
```

**LDAP Injection**: User input in LDAP queries
```javascript
// VULNERABLE
const filter = `(&(uid=${username})(userPassword=${password}))`;

// SECURE — escape special characters
const escapedUsername = username.replace(/[\\*()]/g, '\\$&');
const filter = `(&(uid=${escapedUsername})(userPassword=${escapedPassword}))`;
```

**Template Injection (SSTI)**: User input in template engines
```python
# VULNERABLE — Server-Side Template Injection
return render_template_string(user_input)
# Payload: {{ config.__class__.__init__.__globals__['os'].popen('id').read() }}

# SECURE — pass data as variables
return render_template('template.html', user_data=user_input)
```

### Prevention Matrix

| Injection Type | Primary Defense | Secondary Defense |
|----------------|-----------------|-------------------|
| SQL | Parameterized queries | Input validation |
| NoSQL | Type checking | Schema validation |
| OS Command | Avoid shell; use execFile | Input allowlisting |
| LDAP | Parameterized LDAP | Character escaping |
| XPath | Parameterized XPath | Input validation |
| Template (SSTI) | Never render user input as template | Sandbox template engine |
| Expression Language | Avoid user input in EL | Sandboxed evaluation |

---

## A04:2021 — Insecure Design

### Description

Insecure design is a broad category representing different weaknesses as "missing or ineffective control design." It is not the source of all other Top 10 risk categories. There is a difference between insecure design and insecure implementation.

### Key Concepts

**Threat Modeling**: Identify what can go wrong, then design controls.

```
STRIDE model:
- Spoofing: Can attackers impersonate users? → Authentication
- Tampering: Can attackers modify data? → Integrity controls
- Repudiation: Can users deny actions? → Audit logging
- Information Disclosure: Can data leak? → Encryption, access control
- Denial of Service: Can the system be overwhelmed? → Rate limiting
- Elevation of Privilege: Can users gain unauthorized access? → Authorization
```

**Secure Design Patterns:**
```
1. Fail secure (deny by default)
   - If auth check fails → deny access (don't default to allow)
   - If validation fails → reject input (don't try to sanitize)

2. Defense in depth
   - Multiple layers: WAF → rate limiting → auth → authorization → input validation
   - If one layer fails, others catch it

3. Least privilege
   - Users get minimum necessary permissions
   - Services run with minimum necessary OS permissions
   - Database connections use least-privilege accounts

4. Separation of duties
   - Same person shouldn't create and approve transactions
   - Different credentials for different environments

5. Trust boundaries
   - Validate data at every trust boundary crossing
   - Never trust data from client, even from your own UI
```

### Examples of Insecure Design

**Example 1: No rate limiting on credential recovery**
A password recovery flow sends a reset code via email but has no rate limiting:
- Attacker can brute-force 6-digit reset codes (1M possibilities)
- Fix: Rate limit to 3 attempts per hour + exponential backoff + longer codes

**Example 2: No transaction limits**
A transfer feature has no daily limits:
- Attacker with stolen credentials transfers entire balance in one request
- Fix: Implement transaction limits, require MFA for large transfers

**Example 3: Predictable resource IDs**
Using sequential IDs for resources:
```
/api/invoices/1001
/api/invoices/1002
/api/invoices/1003
```
- Attacker can enumerate all invoices by incrementing
- Fix: Use UUIDs or random IDs + authorization checks

---

## A05:2021 — Security Misconfiguration

### Description

The application might be vulnerable if it is missing appropriate security hardening, has unnecessary features enabled, has default accounts and passwords, reveals overly informative error messages, or has unpatched systems.

### Common Misconfigurations

**1. Missing security headers:**
```
Required headers:
- Strict-Transport-Security: max-age=31536000; includeSubDomains
- X-Content-Type-Options: nosniff
- X-Frame-Options: DENY
- Content-Security-Policy: default-src 'self'
- Referrer-Policy: strict-origin-when-cross-origin
- Permissions-Policy: camera=(), microphone=(), geolocation=()
```

**2. CORS wildcard with credentials:**
```
Access-Control-Allow-Origin: *
Access-Control-Allow-Credentials: true
// This is a CRITICAL misconfiguration — browsers block this, but misconfigured
// proxies may not. The real risk is reflecting the Origin header:
Access-Control-Allow-Origin: [attacker-origin]
Access-Control-Allow-Credentials: true
```

**3. Debug mode in production:**
```python
# Django — NEVER in production
DEBUG = True  # Exposes stack traces, settings, SQL queries

# Flask
app.debug = True  # Full debugger with code execution

# Express
app.set('env', 'development');  # Verbose error pages
```

**4. Default credentials:**
```
admin / admin
admin / password
root / root
sa / (blank)
postgres / postgres
```

**5. Unnecessary services exposed:**
```
- /graphql with introspection enabled in production
- /swagger or /api-docs in production
- /debug, /status with sensitive information
- Database ports (3306, 5432, 27017) on public internet
- Redis port (6379) without authentication
```

### Prevention Checklist

```
[ ] Security headers configured (use helmet.js for Express)
[ ] CORS properly restricted
[ ] Debug mode disabled in production
[ ] Default credentials changed
[ ] Unnecessary features/endpoints disabled in production
[ ] Error pages don't reveal sensitive information
[ ] Directory listing disabled
[ ] Server version headers removed
[ ] GraphQL introspection disabled in production
[ ] API documentation behind authentication in production
[ ] Database ports not publicly accessible
[ ] Redis/Memcached require authentication
[ ] TLS 1.2+ only (no TLS 1.0/1.1, no SSL)
[ ] Strong cipher suites only
```

---

## A06:2021 — Vulnerable and Outdated Components

### Description

Components such as libraries, frameworks, and other software modules run with the same privileges as the application. If a vulnerable component is exploited, such an attack can facilitate serious data loss or server takeover.

### Prevention

```bash
# Automated dependency scanning
npm audit                    # npm
pip-audit                    # Python
cargo audit                  # Rust
govulncheck ./...           # Go
mvn dependency-check:check  # Maven

# Automated updates
# Dependabot (GitHub)
# Renovate Bot
# npm-check-updates

# CI/CD integration
# Run dependency audit in CI pipeline
# Fail build on critical/high vulnerabilities
# Automated PR creation for dependency updates
```

### Known Critical Vulnerabilities

```
Log4Shell (CVE-2021-44228) — CVSS 10.0
- log4j-core 2.0-beta9 to 2.14.1
- Remote Code Execution via JNDI lookup
- Fix: Upgrade to 2.17.1+

Spring4Shell (CVE-2022-22965) — CVSS 9.8
- Spring Framework < 5.3.18
- Remote Code Execution via data binding
- Fix: Upgrade Spring Framework

Prototype Pollution (multiple CVEs)
- lodash, minimist, json5, qs, and many more
- Can lead to RCE, DoS, or property injection
- Fix: Upgrade to patched versions
```

---

## A07:2021 — Identification and Authentication Failures

### Description

Confirmation of the user's identity, authentication, and session management is critical to protect against authentication-related attacks.

### Common Failures

```
1. Permits brute force or credential stuffing
   - No rate limiting on login
   - No account lockout
   - No CAPTCHA after failed attempts

2. Permits weak passwords
   - No minimum length requirement
   - No breached password check
   - No complexity requirements (NIST says length > complexity)

3. Uses weak credential recovery
   - Security questions (easily researched)
   - SMS-based recovery (SIM swapping)
   - Password reset links that don't expire

4. Stores passwords insecurely
   - Plain text, MD5, SHA-1, unsalted hashes
   - Should use bcrypt, argon2id, or scrypt

5. Missing or ineffective MFA
   - No MFA option
   - MFA only on initial login, not on sensitive actions
   - MFA bypass available via API

6. Session management failures
   - Session ID in URL
   - Long-lived sessions (no timeout)
   - Session not invalidated on logout
   - No session regeneration after login
```

### Prevention Controls

```javascript
// Complete secure authentication flow (Express + bcrypt + session)

// 1. Registration with strong password requirements
const passwordSchema = z.string()
  .min(12, 'Password must be at least 12 characters')
  .max(128, 'Password must be at most 128 characters');

app.post('/register', async (req, res) => {
  const { email, password } = req.body;

  // Validate password
  const parsed = passwordSchema.safeParse(password);
  if (!parsed.success) return res.status(400).json({ error: parsed.error });

  // Check breached passwords (haveibeenpwned)
  const isBreached = await checkBreachedPassword(password);
  if (isBreached) return res.status(400).json({ error: 'This password has been compromised' });

  // Hash with bcrypt
  const hash = await bcrypt.hash(password, 12);
  await User.create({ email, passwordHash: hash });
  res.status(201).json({ success: true });
});

// 2. Login with rate limiting and session regeneration
const loginLimiter = rateLimit({ windowMs: 900000, max: 5 });

app.post('/login', loginLimiter, async (req, res) => {
  const user = await User.findOne({ where: { email: req.body.email } });
  const isValid = user && await bcrypt.compare(req.body.password, user.passwordHash);

  if (!isValid) {
    return res.status(401).json({ error: 'Invalid email or password' });
  }

  req.session.regenerate((err) => {
    req.session.userId = user.id;
    req.session.save(() => res.json({ success: true }));
  });
});

// 3. Secure session configuration
app.use(session({
  secret: process.env.SESSION_SECRET,
  name: '__Host-sid',
  resave: false,
  saveUninitialized: false,
  cookie: { httpOnly: true, secure: true, sameSite: 'strict', maxAge: 3600000 },
}));

// 4. Logout with session destruction
app.post('/logout', (req, res) => {
  req.session.destroy(() => {
    res.clearCookie('__Host-sid');
    res.json({ success: true });
  });
});
```

---

## A08:2021 — Software and Data Integrity Failures

### Description

Software and data integrity failures relate to code and infrastructure that does not protect against integrity violations. This includes using untrusted plugins, libraries, or modules from untrusted sources, and insecure CI/CD pipelines.

### Attack Vectors

```
1. Insecure deserialization
   - pickle.loads(), Marshal.load(), unserialize(), ObjectInputStream
   - Can lead to Remote Code Execution

2. Supply chain attacks
   - Compromised npm/pip packages
   - Typosquatting (lodahs instead of lodash)
   - Dependency confusion (internal package name published publicly)

3. Insecure CI/CD pipeline
   - Secrets in CI configuration
   - Untrusted code in PR can access CI secrets
   - Missing integrity verification of build artifacts

4. Auto-update without integrity verification
   - Software updates without signature verification
   - CDN resources without Subresource Integrity (SRI)
```

### Prevention

```html
<!-- Subresource Integrity for CDN resources -->
<script src="https://cdn.example.com/lib.js"
        integrity="sha384-oqVuAfXRKap7fdgcCY5uykM6+R9GqQ8K/uxy9rx7HNQlGYl1kPzQho1wx4JwY8w"
        crossorigin="anonymous"></script>
```

```json
// package-lock.json integrity hashes
{
  "packages": {
    "node_modules/express": {
      "version": "4.18.2",
      "resolved": "https://registry.npmjs.org/express/-/express-4.18.2.tgz",
      "integrity": "sha512-5/PsL6iGPdfQ/lKM1UuielYgv3BUoJfz1aUwU9vHZ+J7gyvwdQXFEBIEIaxeGf0GIcreATNyBExtalisDbuMg=="
    }
  }
}
```

---

## A09:2021 — Security Logging and Monitoring Failures

### Description

Without logging and monitoring, breaches cannot be detected. Insufficient logging, detection, monitoring, and active response occurs any time.

### What to Log

```
ALWAYS log:
- Authentication successes and failures
- Authorization failures (403)
- Input validation failures
- Server errors (500)
- Administrative actions
- Data access (especially sensitive data)
- Configuration changes
- Suspicious activity (multiple failed logins, unusual access patterns)

NEVER log:
- Passwords (even hashed)
- Session tokens / API keys
- Credit card numbers
- Social security numbers
- Personal health information
- Any data protected by regulation
```

### Log Format

```json
{
  "timestamp": "2024-01-15T10:30:00.000Z",
  "level": "warn",
  "event": "authentication_failed",
  "source": "auth-service",
  "requestId": "req-abc123",
  "userId": null,
  "ip": "192.168.1.100",
  "userAgent": "Mozilla/5.0...",
  "path": "/api/login",
  "method": "POST",
  "details": {
    "email": "j***@example.com",
    "reason": "invalid_password",
    "failedAttempts": 3
  }
}
```

---

## A10:2021 — Server-Side Request Forgery (SSRF)

### Description

SSRF flaws occur whenever a web application is fetching a remote resource without validating the user-supplied URL. It allows an attacker to coerce the application to send a crafted request to an unexpected destination.

### Attack Targets

```
1. Cloud metadata services:
   - AWS: http://169.254.169.254/latest/meta-data/
   - GCP: http://metadata.google.internal/computeMetadata/v1/
   - Azure: http://169.254.169.254/metadata/instance?api-version=2021-02-01

2. Internal services:
   - Redis: http://localhost:6379/
   - Elasticsearch: http://localhost:9200/
   - MongoDB: http://localhost:27017/
   - Internal admin panels

3. Cloud service APIs (using stolen credentials from metadata):
   - AWS S3, SQS, Lambda
   - GCP Storage, Pub/Sub
   - Azure Blob, Queue
```

### Prevention

```javascript
import { URL } from 'url';
import dns from 'dns/promises';
import ipaddr from 'ipaddr.js';

async function validateUrl(urlString) {
  let url;
  try {
    url = new URL(urlString);
  } catch {
    throw new Error('Invalid URL');
  }

  // Protocol check
  if (!['http:', 'https:'].includes(url.protocol)) {
    throw new Error('Only HTTP(S) allowed');
  }

  // Block known internal hosts
  const blockedHosts = ['localhost', '127.0.0.1', '::1', '0.0.0.0',
                        'metadata.google.internal', '169.254.169.254'];
  if (blockedHosts.includes(url.hostname)) {
    throw new Error('Internal hosts not allowed');
  }

  // Resolve DNS and check for private IPs
  const addresses = await dns.resolve4(url.hostname).catch(() => []);
  for (const addr of addresses) {
    const parsed = ipaddr.parse(addr);
    const range = parsed.range();
    if (range !== 'unicast') {
      throw new Error('Private/internal IPs not allowed');
    }
  }

  return url;
}
```

---

## Testing Tools Reference

### Static Analysis (SAST)

```
Language-specific:
- JavaScript/TypeScript: ESLint (eslint-plugin-security), Semgrep
- Python: Bandit, Semgrep
- Java: SpotBugs (with FindSecBugs), Semgrep
- Go: gosec, Semgrep
- Ruby: Brakeman
- PHP: PHPStan (with security rules)
- C#: Security Code Scan

Multi-language:
- Semgrep (free, 1000+ rules)
- SonarQube (community edition free)
- CodeQL (free for open source)
```

### Dynamic Analysis (DAST)

```
- OWASP ZAP (free, open-source)
- Burp Suite (professional — industry standard)
- Nuclei (free, template-based scanner)
```

### Dependency Scanning (SCA)

```
- npm audit / yarn audit (Node.js)
- pip-audit (Python)
- cargo audit (Rust)
- govulncheck (Go)
- OWASP Dependency-Check (Java)
- Snyk (multi-language, free tier)
- Trivy (container + dependency scanning)
- Grype (container + dependency scanning)
```

### Secret Detection

```
- TruffleHog (git history scanning)
- GitLeaks (git history scanning)
- detect-secrets (pre-commit hook)
- GitHub Secret Scanning (built into GitHub)
```

---

## OWASP Top 10 Quick Reference Card

| # | Category | Primary Defense | Key CWE |
|---|----------|-----------------|---------|
| A01 | Broken Access Control | Authorization middleware, IDOR checks | CWE-284, CWE-639 |
| A02 | Cryptographic Failures | Strong algorithms, key management | CWE-327, CWE-330 |
| A03 | Injection | Parameterized queries, input validation | CWE-79, CWE-89 |
| A04 | Insecure Design | Threat modeling, secure design patterns | CWE-840 |
| A05 | Security Misconfiguration | Hardening checklist, security headers | CWE-16 |
| A06 | Vulnerable Components | Dependency scanning, automated updates | CWE-1104 |
| A07 | Auth Failures | Strong passwords, MFA, session management | CWE-287 |
| A08 | Integrity Failures | SRI, signed artifacts, secure CI/CD | CWE-502 |
| A09 | Logging Failures | Comprehensive logging, monitoring, alerting | CWE-778 |
| A10 | SSRF | URL validation, allowlisting, network segmentation | CWE-918 |
