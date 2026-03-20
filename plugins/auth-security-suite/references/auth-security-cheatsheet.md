# Auth & Security Cheatsheet

## JWT Token Structure

```
Header.Payload.Signature

Header:  { "alg": "HS256", "typ": "JWT" }
Payload: { "sub": "user-id", "role": "admin", "iat": 1234567890, "exp": 1234568490 }
Signature: HMACSHA256(base64(header) + "." + base64(payload), secret)
```

## Token Lifetimes

| Token | Duration | Storage | Purpose |
|-------|----------|---------|---------|
| Access token | 15 min | HttpOnly cookie or memory | API authorization |
| Refresh token | 7 days | HttpOnly cookie + DB | Get new access token |
| CSRF token | Per session | Cookie + header | Prevent CSRF |
| Password reset | 1 hour | DB (hashed) | One-time password reset |
| Email verification | 24 hours | DB (hashed) | Verify email ownership |
| API key | No expiry | DB (hashed) | Server-to-server auth |

## Cookie Settings

```typescript
// Access token cookie
res.cookie('token', accessToken, {
  httpOnly: true,     // No JS access
  secure: true,       // HTTPS only
  sameSite: 'lax',    // CSRF protection
  maxAge: 15 * 60 * 1000,  // 15 minutes
  path: '/',
});

// Refresh token cookie
res.cookie('refreshToken', refreshToken, {
  httpOnly: true,
  secure: true,
  sameSite: 'lax',
  maxAge: 7 * 24 * 60 * 60 * 1000,  // 7 days
  path: '/api/auth/refresh',  // Only sent to refresh endpoint
});
```

## Password Hashing

| Algorithm | Cost Factor | Time (approx) | Recommendation |
|-----------|------------|----------------|----------------|
| bcrypt | 10 | ~100ms | Minimum for production |
| bcrypt | 12 | ~300ms | Recommended default |
| bcrypt | 14 | ~1s | High security |
| argon2id | default | ~250ms | Modern alternative |
| scrypt | N=2^14 | ~100ms | Node.js native |

```typescript
import { hash, compare } from 'bcrypt';
const hashed = await hash(password, 12);
const valid = await compare(password, hashed);
```

## Security Headers

| Header | Value | Purpose |
|--------|-------|---------|
| `Strict-Transport-Security` | `max-age=31536000; includeSubDomains` | Force HTTPS |
| `Content-Security-Policy` | `default-src 'self'; script-src 'self' 'nonce-xxx'` | Prevent XSS |
| `X-Content-Type-Options` | `nosniff` | Prevent MIME sniffing |
| `X-Frame-Options` | `DENY` | Prevent clickjacking |
| `Referrer-Policy` | `strict-origin-when-cross-origin` | Control referrer info |
| `Permissions-Policy` | `camera=(), microphone=(), geolocation=()` | Disable browser features |
| `X-XSS-Protection` | `0` | Disable buggy browser filter |

## CORS Configuration

```typescript
// Development
cors({ origin: 'http://localhost:5173', credentials: true })

// Production
cors({
  origin: ['https://myapp.com', 'https://admin.myapp.com'],
  credentials: true,
  methods: ['GET', 'POST', 'PUT', 'DELETE', 'PATCH'],
  allowedHeaders: ['Content-Type', 'Authorization', 'X-CSRF-Token'],
  maxAge: 86400,
})
```

## Rate Limiting Tiers

| Endpoint | Window | Max Requests | Purpose |
|----------|--------|-------------|---------|
| `/api/auth/login` | 15 min | 5 | Brute force prevention |
| `/api/auth/register` | 1 hour | 3 | Spam prevention |
| `/api/auth/forgot-password` | 1 hour | 3 | Abuse prevention |
| `/api/*` (authenticated) | 15 min | 100 | General rate limit |
| `/api/*` (unauthenticated) | 15 min | 20 | Public API limit |

## OWASP Top 10 Quick Reference

| # | Risk | Prevention |
|---|------|-----------|
| A01 | Broken Access Control | RBAC, deny by default, validate ownership |
| A02 | Cryptographic Failures | HTTPS, bcrypt, no secrets in code |
| A03 | Injection | Parameterized queries, Zod validation |
| A04 | Insecure Design | Threat modeling, security requirements |
| A05 | Security Misconfiguration | Helmet.js, disable debug, env validation |
| A06 | Vulnerable Components | `npm audit`, Dependabot, lock files |
| A07 | Auth Failures | MFA, rate limiting, secure session mgmt |
| A08 | Data Integrity Failures | Verify signatures, CI/CD security |
| A09 | Logging Failures | Log auth events, never log secrets |
| A10 | SSRF | Allowlist URLs, validate redirects |

## Input Validation Patterns

```typescript
import { z } from 'zod';

const emailSchema = z.string().email().max(255).toLowerCase();
const passwordSchema = z.string().min(8).max(128);
const usernameSchema = z.string().min(3).max(30).regex(/^[a-zA-Z0-9_-]+$/);
const uuidSchema = z.string().uuid();
const urlSchema = z.string().url().max(2048);
const phoneSchema = z.string().regex(/^\+?[1-9]\d{1,14}$/);

// Sanitize HTML content
import DOMPurify from 'isomorphic-dompurify';
const clean = DOMPurify.sanitize(userInput, { ALLOWED_TAGS: ['b', 'i', 'em', 'strong'] });
```

## OAuth 2.0 Flow Quick Reference

```
Authorization Code + PKCE (SPAs):
1. Generate code_verifier (random 43-128 chars)
2. Create code_challenge = BASE64URL(SHA256(code_verifier))
3. Redirect to /authorize?response_type=code&code_challenge=...
4. User authenticates, provider redirects with ?code=abc123
5. POST /token with code + code_verifier
6. Provider verifies SHA256(code_verifier) === code_challenge
7. Returns access_token + refresh_token

Common OAuth Scopes:
- Google: openid profile email
- GitHub: user:email read:user
- Microsoft: openid profile email User.Read
- Facebook: email public_profile
```

## Session Security Checklist

```
[ ] Regenerate session ID after login (prevent fixation)
[ ] Set secure cookie flags (httpOnly, secure, sameSite)
[ ] Implement absolute session timeout (24h max)
[ ] Implement idle session timeout (30m)
[ ] Invalidate sessions on password change
[ ] Store sessions server-side (Redis/DB, not JWT-only)
[ ] Log session creation and destruction
[ ] Limit concurrent sessions per user
```

## Common Auth Mistakes

| Mistake | Impact | Fix |
|---------|--------|-----|
| Storing JWT in localStorage | XSS can steal tokens | Use HttpOnly cookies |
| No refresh token rotation | Stolen refresh = permanent access | Rotate on each use, track families |
| Generic "Invalid credentials" missing | Username enumeration | Always return same error for login failures |
| No rate limiting on login | Brute force attacks | 5 attempts per 15 min window |
| Password reset token in URL params | Token in server logs/referrer | Use POST with token in body |
| Checking `req.headers.origin` for CORS | Can be spoofed | Use cors middleware with allowlist |
| Secrets in environment without validation | App starts with missing config | Validate env vars at startup with Zod |
| `console.log(user)` in production | Logs contain PII/tokens | Strip sensitive fields before logging |

## Crypto Quick Reference

```typescript
import { randomBytes, createHash, timingSafeEqual } from 'crypto';

// Generate random token
const token = randomBytes(32).toString('hex');

// Hash for storage (non-password)
const hash = createHash('sha256').update(token).digest('hex');

// Timing-safe comparison (prevent timing attacks)
const a = Buffer.from(hash1, 'hex');
const b = Buffer.from(hash2, 'hex');
const match = a.length === b.length && timingSafeEqual(a, b);
```

## Environment Variables Template

```bash
# Auth
JWT_SECRET="<random-64-chars>"          # openssl rand -hex 32
JWT_REFRESH_SECRET="<random-64-chars>"
SESSION_SECRET="<random-64-chars>"
BCRYPT_ROUNDS=12

# OAuth - Google
GOOGLE_CLIENT_ID=""
GOOGLE_CLIENT_SECRET=""
GOOGLE_CALLBACK_URL="http://localhost:3000/api/auth/google/callback"

# OAuth - GitHub
GITHUB_CLIENT_ID=""
GITHUB_CLIENT_SECRET=""
GITHUB_CALLBACK_URL="http://localhost:3000/api/auth/github/callback"

# Security
CORS_ORIGINS="http://localhost:5173"     # comma-separated
RATE_LIMIT_WINDOW_MS=900000             # 15 minutes
RATE_LIMIT_MAX=100
```
