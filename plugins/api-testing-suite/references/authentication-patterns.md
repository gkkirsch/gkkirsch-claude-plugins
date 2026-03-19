# API Authentication and Authorization Patterns Reference

Comprehensive reference for API authentication and authorization mechanisms, including
implementation patterns, testing strategies, security considerations, and best practices.

---

## Table of Contents

- [Overview](#overview)
- [API Keys](#api-keys)
- [Bearer Tokens](#bearer-tokens)
- [JWT (JSON Web Tokens)](#jwt-json-web-tokens)
- [OAuth 2.0](#oauth-20)
- [OpenID Connect](#openid-connect)
- [Session-Based Authentication](#session-based-authentication)
- [Basic Authentication](#basic-authentication)
- [HMAC Signatures](#hmac-signatures)
- [Mutual TLS (mTLS)](#mutual-tls-mtls)
- [SAML](#saml)
- [Rate Limiting by Auth Type](#rate-limiting-by-auth-type)
- [Security Best Practices](#security-best-practices)
- [Testing Patterns](#testing-patterns)

---

## Overview

### Authentication vs Authorization

| Concept | Question | Example |
|---------|----------|---------|
| **Authentication** (AuthN) | "Who are you?" | Login with email/password, API key |
| **Authorization** (AuthZ) | "What can you do?" | Admin can delete users, users can only view own data |

### Authentication Methods Comparison

| Method | Best For | Security | Complexity | Stateless |
|--------|----------|----------|------------|-----------|
| API Keys | Service-to-service, simple integrations | Medium | Low | Yes |
| Bearer Tokens (JWT) | SPAs, mobile apps, microservices | High | Medium | Yes |
| OAuth 2.0 | Third-party access, delegated auth | High | High | Yes |
| Session/Cookie | Traditional web apps, SSR | Medium | Low | No |
| Basic Auth | Internal APIs, dev/staging | Low | Low | Yes |
| HMAC Signatures | Webhooks, high-security APIs | Very High | High | Yes |
| mTLS | Zero-trust, service mesh | Very High | Very High | Yes |

### Choosing the Right Method

```
Is this a public-facing API?
├── Yes → Do third parties need access?
│   ├── Yes → OAuth 2.0 (+ OpenID Connect for user identity)
│   └── No → Is it a SPA or mobile app?
│       ├── Yes → JWT Bearer tokens (with refresh tokens)
│       └── No → Is it a traditional web app (server-rendered)?
│           ├── Yes → Session-based (cookies)
│           └── No → API Keys (for server-to-server)
├── No → Is it service-to-service?
│   ├── Yes → Is it within a trusted network?
│   │   ├── Yes → mTLS or API Keys
│   │   └── No → mTLS or HMAC signatures
│   └── No → Is it an internal tool?
│       ├── Yes → API Keys or Basic Auth (over HTTPS)
│       └── No → JWT Bearer tokens
```

---

## API Keys

### How It Works

A unique string identifier issued to each client. Sent with every request in a header
or query parameter.

```http
# In header (preferred)
GET /api/products HTTP/1.1
Host: api.example.com
X-API-Key: sk_live_abc123def456ghi789

# In query parameter (less secure — visible in logs/URLs)
GET /api/products?api_key=sk_live_abc123def456ghi789 HTTP/1.1
Host: api.example.com

# In Authorization header (some APIs)
GET /api/products HTTP/1.1
Authorization: ApiKey sk_live_abc123def456ghi789
```

### Implementation

```typescript
// Express middleware for API key authentication
import { Request, Response, NextFunction } from 'express';
import { db } from './database';

interface ApiKeyRecord {
  id: string;
  key: string;
  name: string;
  ownerId: string;
  permissions: string[];
  rateLimit: number;
  lastUsedAt: Date | null;
  expiresAt: Date | null;
  isActive: boolean;
}

async function apiKeyAuth(req: Request, res: Response, next: NextFunction) {
  const apiKey = req.headers['x-api-key'] as string;

  if (!apiKey) {
    return res.status(401).json({
      error: 'Unauthorized',
      message: 'API key is required. Include X-API-Key header.',
      statusCode: 401,
    });
  }

  // Hash the key before lookup (store hashed keys, not plaintext)
  const hashedKey = hashApiKey(apiKey);

  const record = await db.apiKey.findFirst({
    where: { keyHash: hashedKey, isActive: true },
    include: { owner: true },
  });

  if (!record) {
    return res.status(401).json({
      error: 'Unauthorized',
      message: 'Invalid API key',
      statusCode: 401,
    });
  }

  // Check expiration
  if (record.expiresAt && record.expiresAt < new Date()) {
    return res.status(401).json({
      error: 'Unauthorized',
      message: 'API key has expired',
      statusCode: 401,
    });
  }

  // Attach key info to request
  req.apiKey = record;
  req.user = record.owner;

  // Update last used timestamp (async, don't block the request)
  db.apiKey.update({
    where: { id: record.id },
    data: { lastUsedAt: new Date() },
  }).catch(() => {}); // Fire and forget

  next();
}
```

### Key Generation Best Practices

```typescript
import { randomBytes, createHash } from 'crypto';

// Generate a secure API key with prefix
function generateApiKey(prefix: string = 'sk'): { key: string; hash: string } {
  const entropy = randomBytes(32).toString('base64url'); // 256 bits of entropy
  const key = `${prefix}_${entropy}`;
  const hash = createHash('sha256').update(key).digest('hex');

  return { key, hash };
  // Store `hash` in database, return `key` to user ONCE
  // You cannot recover the key from the hash
}

// Naming convention:
// sk_live_... — Secret key, production
// pk_live_... — Public key, production (limited scope)
// sk_test_... — Secret key, test/sandbox
// pk_test_... — Public key, test/sandbox
```

### Testing API Key Authentication

```typescript
describe('API Key Authentication', () => {
  const validApiKey = 'sk_test_abc123';

  it('should authenticate with valid API key in header', async () => {
    await request(app)
      .get('/api/products')
      .set('X-API-Key', validApiKey)
      .expect(200);
  });

  it('should reject missing API key', async () => {
    const res = await request(app)
      .get('/api/products')
      .expect(401);

    expect(res.body.message).toContain('API key is required');
  });

  it('should reject invalid API key', async () => {
    await request(app)
      .get('/api/products')
      .set('X-API-Key', 'invalid_key')
      .expect(401);
  });

  it('should reject expired API key', async () => {
    await request(app)
      .get('/api/products')
      .set('X-API-Key', 'sk_test_expired_key')
      .expect(401);
  });

  it('should reject revoked API key', async () => {
    await request(app)
      .get('/api/products')
      .set('X-API-Key', 'sk_test_revoked_key')
      .expect(401);
  });

  it('should enforce key permissions', async () => {
    // Read-only key trying to write
    await request(app)
      .post('/api/products')
      .set('X-API-Key', 'sk_test_readonly_key')
      .send({ name: 'Test Product', price: 10 })
      .expect(403);
  });
});
```

---

## Bearer Tokens

### How It Works

A token (usually JWT) sent in the Authorization header with the "Bearer" scheme.

```http
GET /api/users/me HTTP/1.1
Host: api.example.com
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ1c2VyLTEyMyIsInJvbGUiOiJhZG1pbiIsImlhdCI6MTcxMDUwMDAwMH0.signature
```

### Token Lifecycle

```
1. Login       → POST /auth/login with credentials → get {accessToken, refreshToken}
2. Use         → Include accessToken in Authorization header for all API calls
3. Expiry      → Access token expires (1 hour typical)
4. Refresh     → POST /auth/refresh with refreshToken → get new {accessToken, refreshToken}
5. Logout      → POST /auth/logout → invalidate all tokens
```

---

## JWT (JSON Web Tokens)

### Structure

A JWT consists of three parts separated by dots: `header.payload.signature`

```
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.     ← Header (algorithm + type)
eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWV9.  ← Payload (claims)
TJVA95OrM7E2cBab30RMHrHDcEfxjoYZgeFONFh7HgQ  ← Signature
```

**Header (decoded):**
```json
{
  "alg": "HS256",   // Signing algorithm (HS256, RS256, ES256)
  "typ": "JWT"      // Token type
}
```

**Payload (decoded):**
```json
{
  "sub": "user-123",              // Subject (user ID)
  "email": "jane@example.com",    // Custom claim
  "name": "Jane Doe",             // Custom claim
  "role": "admin",                // Custom claim
  "iat": 1710500000,              // Issued at (Unix timestamp)
  "exp": 1710503600,              // Expiration (1 hour later)
  "nbf": 1710500000,              // Not before
  "iss": "https://api.example.com", // Issuer
  "aud": "https://app.example.com", // Audience
  "jti": "unique-token-id"          // JWT ID (for revocation)
}
```

### Standard Claims (RFC 7519)

| Claim | Full Name | Description | Required |
|-------|-----------|-------------|----------|
| `sub` | Subject | Who the token represents (user ID) | Recommended |
| `iss` | Issuer | Who issued the token | Recommended |
| `aud` | Audience | Intended recipient(s) | Recommended |
| `exp` | Expiration | When the token expires (Unix timestamp) | Recommended |
| `nbf` | Not Before | Token not valid before this time | Optional |
| `iat` | Issued At | When the token was created | Recommended |
| `jti` | JWT ID | Unique token identifier | Optional |

### Signing Algorithms

| Algorithm | Type | Key | Best For |
|-----------|------|-----|----------|
| HS256 | Symmetric | Shared secret | Single service (same signer and verifier) |
| HS384 | Symmetric | Shared secret | Single service |
| HS512 | Symmetric | Shared secret | Single service |
| RS256 | Asymmetric | RSA keypair | Multi-service (auth server signs, APIs verify with public key) |
| RS384 | Asymmetric | RSA keypair | Multi-service |
| RS512 | Asymmetric | RSA keypair | Multi-service |
| ES256 | Asymmetric | ECDSA P-256 | Multi-service (smaller keys than RSA) |
| ES384 | Asymmetric | ECDSA P-384 | Multi-service |
| ES512 | Asymmetric | ECDSA P-521 | Multi-service |
| PS256 | Asymmetric | RSA-PSS | Multi-service (probabilistic signatures) |
| EdDSA | Asymmetric | Ed25519/Ed448 | Modern, fastest asymmetric |

### JWT Implementation

```typescript
// Token generation
import jwt from 'jsonwebtoken';

const JWT_SECRET = process.env.JWT_SECRET!; // At least 256 bits for HS256
const ACCESS_TOKEN_TTL = '1h';
const REFRESH_TOKEN_TTL = '30d';

function generateAccessToken(user: User): string {
  return jwt.sign(
    {
      sub: user.id,
      email: user.email,
      role: user.role,
    },
    JWT_SECRET,
    {
      expiresIn: ACCESS_TOKEN_TTL,
      issuer: 'https://api.example.com',
      audience: 'https://app.example.com',
    }
  );
}

function generateRefreshToken(user: User): string {
  const token = jwt.sign(
    { sub: user.id, type: 'refresh' },
    JWT_SECRET,
    { expiresIn: REFRESH_TOKEN_TTL }
  );

  // Store refresh token hash in database for revocation
  db.refreshToken.create({
    data: {
      tokenHash: hashToken(token),
      userId: user.id,
      expiresAt: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000),
    },
  });

  return token;
}

// Token verification middleware
function verifyAccessToken(req: Request, res: Response, next: NextFunction) {
  const authHeader = req.headers.authorization;

  if (!authHeader?.startsWith('Bearer ')) {
    return res.status(401).json({
      error: 'Unauthorized',
      message: 'Bearer token required',
      statusCode: 401,
    });
  }

  const token = authHeader.slice(7);

  try {
    const payload = jwt.verify(token, JWT_SECRET, {
      issuer: 'https://api.example.com',
      audience: 'https://app.example.com',
      algorithms: ['HS256'], // IMPORTANT: Explicitly set allowed algorithms
    }) as JwtPayload;

    req.user = {
      id: payload.sub!,
      email: payload.email,
      role: payload.role,
    };

    next();
  } catch (error) {
    if (error instanceof jwt.TokenExpiredError) {
      return res.status(401).json({
        error: 'Unauthorized',
        message: 'Token has expired. Use /auth/refresh to get a new token.',
        statusCode: 401,
      });
    }
    if (error instanceof jwt.JsonWebTokenError) {
      return res.status(401).json({
        error: 'Unauthorized',
        message: 'Invalid token',
        statusCode: 401,
      });
    }
    return res.status(401).json({
      error: 'Unauthorized',
      message: 'Token verification failed',
      statusCode: 401,
    });
  }
}
```

### Refresh Token Rotation

```typescript
// Refresh token endpoint with rotation
async function refreshToken(req: Request, res: Response) {
  const { refreshToken } = req.body;

  if (!refreshToken) {
    return res.status(400).json({
      error: 'Bad Request',
      message: 'Refresh token is required',
      statusCode: 400,
    });
  }

  try {
    const payload = jwt.verify(refreshToken, JWT_SECRET) as JwtPayload;

    if (payload.type !== 'refresh') {
      return res.status(401).json({
        error: 'Unauthorized',
        message: 'Invalid token type',
        statusCode: 401,
      });
    }

    // Check if refresh token exists and hasn't been used
    const tokenHash = hashToken(refreshToken);
    const storedToken = await db.refreshToken.findFirst({
      where: { tokenHash, isUsed: false },
      include: { user: true },
    });

    if (!storedToken) {
      // SECURITY: If a used token is presented, someone may have stolen it
      // Revoke ALL tokens for this user as a precaution
      await db.refreshToken.updateMany({
        where: { userId: payload.sub },
        data: { isUsed: true },
      });

      return res.status(401).json({
        error: 'Unauthorized',
        message: 'Refresh token has been revoked. Please log in again.',
        statusCode: 401,
      });
    }

    // Mark old refresh token as used (rotation)
    await db.refreshToken.update({
      where: { id: storedToken.id },
      data: { isUsed: true, usedAt: new Date() },
    });

    // Generate new token pair
    const newAccessToken = generateAccessToken(storedToken.user);
    const newRefreshToken = generateRefreshToken(storedToken.user);

    return res.json({
      accessToken: newAccessToken,
      refreshToken: newRefreshToken,
      expiresIn: 3600,
    });
  } catch (error) {
    return res.status(401).json({
      error: 'Unauthorized',
      message: 'Invalid or expired refresh token',
      statusCode: 401,
    });
  }
}
```

### JWT Testing

```typescript
describe('JWT Authentication', () => {
  it('should accept valid JWT', async () => {
    const token = generateAccessToken(testUser);
    await request(app)
      .get('/api/users/me')
      .set('Authorization', `Bearer ${token}`)
      .expect(200);
  });

  it('should reject expired JWT', async () => {
    const token = jwt.sign(
      { sub: 'user-1', role: 'user' },
      JWT_SECRET,
      { expiresIn: '-1h' } // Already expired
    );

    const res = await request(app)
      .get('/api/users/me')
      .set('Authorization', `Bearer ${token}`)
      .expect(401);

    expect(res.body.message).toContain('expired');
  });

  it('should reject JWT signed with wrong secret', async () => {
    const token = jwt.sign(
      { sub: 'user-1', role: 'admin' },
      'wrong-secret'
    );

    await request(app)
      .get('/api/users/me')
      .set('Authorization', `Bearer ${token}`)
      .expect(401);
  });

  it('should reject JWT with none algorithm', async () => {
    // Algorithm "none" attack — attacker tries to bypass signature
    const header = Buffer.from('{"alg":"none","typ":"JWT"}').toString('base64url');
    const payload = Buffer.from('{"sub":"user-1","role":"admin"}').toString('base64url');
    const token = `${header}.${payload}.`;

    await request(app)
      .get('/api/users/me')
      .set('Authorization', `Bearer ${token}`)
      .expect(401);
  });

  it('should reject JWT with modified payload', async () => {
    // Sign a valid token, then modify the payload
    const token = generateAccessToken({ ...testUser, role: 'user' });
    const parts = token.split('.');

    // Modify payload to claim admin role
    const payload = JSON.parse(Buffer.from(parts[1], 'base64url').toString());
    payload.role = 'admin';
    parts[1] = Buffer.from(JSON.stringify(payload)).toString('base64url');

    const tamperedToken = parts.join('.');

    await request(app)
      .get('/api/admin/users')
      .set('Authorization', `Bearer ${tamperedToken}`)
      .expect(401);
  });

  it('should reject missing Bearer prefix', async () => {
    const token = generateAccessToken(testUser);
    await request(app)
      .get('/api/users/me')
      .set('Authorization', token) // Missing "Bearer "
      .expect(401);
  });

  it('should handle refresh token rotation', async () => {
    // Login to get tokens
    const loginRes = await request(app)
      .post('/api/auth/login')
      .send({ email: 'user@test.com', password: 'Password123!' })
      .expect(200);

    const { refreshToken } = loginRes.body;

    // Refresh once — should succeed
    const refresh1 = await request(app)
      .post('/api/auth/refresh')
      .send({ refreshToken })
      .expect(200);

    expect(refresh1.body.accessToken).toBeDefined();
    expect(refresh1.body.refreshToken).toBeDefined();
    expect(refresh1.body.refreshToken).not.toBe(refreshToken); // New token

    // Try to use old refresh token again — should fail (already used)
    const refresh2 = await request(app)
      .post('/api/auth/refresh')
      .send({ refreshToken }) // Old token
      .expect(401);

    expect(refresh2.body.message).toContain('revoked');
  });
});
```

---

## OAuth 2.0

### Grant Types

#### Authorization Code (with PKCE)

The most secure flow for SPAs, mobile apps, and server-side applications.

```
┌──────────┐                                ┌───────────────┐
│  Client   │                                │  Auth Server  │
│ (Browser) │                                │               │
└─────┬─────┘                                └───────┬───────┘
      │                                              │
      │ 1. Generate code_verifier + code_challenge    │
      │ 2. Redirect to auth server                    │
      │────────────────────────────────────────────▶│
      │    GET /authorize?                            │
      │      response_type=code&                      │
      │      client_id=...&                           │
      │      redirect_uri=...&                        │
      │      scope=read+write&                        │
      │      state=random-state&                      │
      │      code_challenge=...&                      │
      │      code_challenge_method=S256               │
      │                                              │
      │ 3. User logs in and consents                 │
      │                                              │
      │ 4. Redirect back with authorization code     │
      │◀────────────────────────────────────────────│
      │    302 redirect_uri?code=AUTH_CODE&state=...  │
      │                                              │
      │ 5. Exchange code for tokens                   │
      │────────────────────────────────────────────▶│
      │    POST /token                                │
      │      grant_type=authorization_code&           │
      │      code=AUTH_CODE&                          │
      │      redirect_uri=...&                        │
      │      client_id=...&                           │
      │      code_verifier=...                        │
      │                                              │
      │ 6. Receive tokens                             │
      │◀────────────────────────────────────────────│
      │    { access_token, refresh_token, ... }       │
      │                                              │
```

#### Client Credentials

For service-to-service authentication (no user involved).

```http
# Request
POST /oauth/token HTTP/1.1
Host: auth.example.com
Content-Type: application/x-www-form-urlencoded

grant_type=client_credentials&
client_id=service-abc&
client_secret=secret-xyz&
scope=api:read api:write

# Response
HTTP/1.1 200 OK
Content-Type: application/json

{
  "access_token": "eyJhbGciOiJSUzI1NiIs...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "scope": "api:read api:write"
}
```

#### Resource Owner Password Credentials (ROPC)

Legacy flow — the client collects the user's credentials directly. NOT recommended
for new applications but still encountered in migrations.

```http
POST /oauth/token HTTP/1.1
Content-Type: application/x-www-form-urlencoded

grant_type=password&
username=jane@example.com&
password=SecurePass123!&
client_id=mobile-app&
scope=profile email

HTTP/1.1 200 OK
{
  "access_token": "eyJhbGciOiJSUzI1NiIs...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "refresh_token": "dGhpcyBpcyBhIHJl..."
}
```

#### Device Authorization (Device Code)

For devices with limited input (Smart TVs, CLI tools, IoT).

```http
# Step 1: Device requests a code
POST /oauth/device/code HTTP/1.1
Content-Type: application/x-www-form-urlencoded

client_id=tv-app&scope=profile

# Response
HTTP/1.1 200 OK
{
  "device_code": "DEVICE_CODE",
  "user_code": "WDJB-MJHT",
  "verification_uri": "https://auth.example.com/device",
  "expires_in": 1800,
  "interval": 5
}

# Step 2: User goes to verification_uri and enters user_code
# Step 3: Device polls for token
POST /oauth/token HTTP/1.1
Content-Type: application/x-www-form-urlencoded

grant_type=urn:ietf:params:oauth:grant-type:device_code&
device_code=DEVICE_CODE&
client_id=tv-app

# Response (while user hasn't authorized yet)
HTTP/1.1 400 Bad Request
{"error": "authorization_pending"}

# Response (after user authorizes)
HTTP/1.1 200 OK
{"access_token": "...", "token_type": "Bearer", "expires_in": 3600}
```

### OAuth 2.0 Testing

```typescript
describe('OAuth 2.0 Flows', () => {
  describe('Authorization Code + PKCE', () => {
    it('should generate valid authorization URL', () => {
      const codeVerifier = generateCodeVerifier();
      const codeChallenge = generateCodeChallenge(codeVerifier);
      const state = generateRandomState();

      const authUrl = buildAuthorizationUrl({
        clientId: 'test-client',
        redirectUri: 'http://localhost:3000/callback',
        scope: 'openid profile email',
        state,
        codeChallenge,
        codeChallengeMethod: 'S256',
      });

      expect(authUrl).toContain('response_type=code');
      expect(authUrl).toContain('code_challenge=');
      expect(authUrl).toContain('code_challenge_method=S256');
    });

    it('should exchange authorization code for tokens', async () => {
      const res = await request(authServer)
        .post('/oauth/token')
        .send({
          grant_type: 'authorization_code',
          code: 'test-auth-code',
          redirect_uri: 'http://localhost:3000/callback',
          client_id: 'test-client',
          code_verifier: 'test-code-verifier',
        })
        .expect(200);

      expect(res.body.access_token).toBeDefined();
      expect(res.body.token_type).toBe('Bearer');
      expect(res.body.expires_in).toBeGreaterThan(0);
    });

    it('should reject invalid code_verifier', async () => {
      await request(authServer)
        .post('/oauth/token')
        .send({
          grant_type: 'authorization_code',
          code: 'test-auth-code',
          redirect_uri: 'http://localhost:3000/callback',
          client_id: 'test-client',
          code_verifier: 'wrong-verifier',
        })
        .expect(400);
    });

    it('should reject expired authorization code', async () => {
      await request(authServer)
        .post('/oauth/token')
        .send({
          grant_type: 'authorization_code',
          code: 'expired-auth-code',
          redirect_uri: 'http://localhost:3000/callback',
          client_id: 'test-client',
          code_verifier: 'valid-verifier',
        })
        .expect(400);
    });

    it('should reject replay of authorization code', async () => {
      // First use — should succeed
      await request(authServer)
        .post('/oauth/token')
        .send({
          grant_type: 'authorization_code',
          code: 'one-time-code',
          redirect_uri: 'http://localhost:3000/callback',
          client_id: 'test-client',
          code_verifier: 'valid-verifier',
        })
        .expect(200);

      // Second use — should fail
      await request(authServer)
        .post('/oauth/token')
        .send({
          grant_type: 'authorization_code',
          code: 'one-time-code',
          redirect_uri: 'http://localhost:3000/callback',
          client_id: 'test-client',
          code_verifier: 'valid-verifier',
        })
        .expect(400);
    });
  });

  describe('Client Credentials', () => {
    it('should issue token for valid client credentials', async () => {
      const res = await request(authServer)
        .post('/oauth/token')
        .send({
          grant_type: 'client_credentials',
          client_id: 'service-client',
          client_secret: 'service-secret',
          scope: 'api:read',
        })
        .expect(200);

      expect(res.body.access_token).toBeDefined();
      expect(res.body.token_type).toBe('Bearer');
      // Client credentials should NOT return a refresh token
      expect(res.body.refresh_token).toBeUndefined();
    });

    it('should reject invalid client secret', async () => {
      await request(authServer)
        .post('/oauth/token')
        .send({
          grant_type: 'client_credentials',
          client_id: 'service-client',
          client_secret: 'wrong-secret',
        })
        .expect(401);
    });

    it('should restrict scopes to what the client is allowed', async () => {
      const res = await request(authServer)
        .post('/oauth/token')
        .send({
          grant_type: 'client_credentials',
          client_id: 'read-only-client',
          client_secret: 'read-only-secret',
          scope: 'api:read api:write api:admin',
        })
        .expect(200);

      // Token should only have the scopes the client is allowed
      expect(res.body.scope).toBe('api:read');
      expect(res.body.scope).not.toContain('api:write');
      expect(res.body.scope).not.toContain('api:admin');
    });
  });
});
```

---

## OpenID Connect

### How It Works

OpenID Connect (OIDC) is an identity layer on top of OAuth 2.0. It adds:
- **ID Token** — JWT containing user identity claims
- **UserInfo endpoint** — Get additional user information
- **Discovery** — `.well-known/openid-configuration` for auto-configuration

### ID Token

```json
{
  "iss": "https://auth.example.com",
  "sub": "user-123",
  "aud": "client-id",
  "exp": 1710503600,
  "iat": 1710500000,
  "nonce": "random-nonce",
  "at_hash": "abc123",
  "email": "jane@example.com",
  "email_verified": true,
  "name": "Jane Doe",
  "picture": "https://cdn.example.com/avatars/jane.png"
}
```

### Standard Scopes

| Scope | Claims Returned |
|-------|----------------|
| `openid` | `sub` (required for OIDC) |
| `profile` | `name`, `family_name`, `given_name`, `middle_name`, `nickname`, `picture`, `gender`, `birthdate`, `zoneinfo`, `locale`, `updated_at` |
| `email` | `email`, `email_verified` |
| `address` | `address` (formatted, street_address, locality, region, postal_code, country) |
| `phone` | `phone_number`, `phone_number_verified` |

### Discovery Document

```
GET /.well-known/openid-configuration HTTP/1.1
Host: auth.example.com

HTTP/1.1 200 OK
{
  "issuer": "https://auth.example.com",
  "authorization_endpoint": "https://auth.example.com/authorize",
  "token_endpoint": "https://auth.example.com/token",
  "userinfo_endpoint": "https://auth.example.com/userinfo",
  "jwks_uri": "https://auth.example.com/.well-known/jwks.json",
  "scopes_supported": ["openid", "profile", "email"],
  "response_types_supported": ["code"],
  "grant_types_supported": ["authorization_code", "client_credentials", "refresh_token"],
  "id_token_signing_alg_values_supported": ["RS256"],
  "code_challenge_methods_supported": ["S256"]
}
```

---

## Session-Based Authentication

### How It Works

Traditional web authentication using server-side sessions and cookies.

```
┌──────────┐                    ┌────────────┐                    ┌────────────┐
│  Browser  │                    │   Server   │                    │  Session   │
│           │                    │            │                    │   Store    │
└─────┬─────┘                    └─────┬──────┘                    └─────┬──────┘
      │ 1. POST /login                 │                                 │
      │    { email, password }         │                                 │
      │───────────────────────────────▶│                                 │
      │                                │ 2. Validate credentials          │
      │                                │ 3. Create session                │
      │                                │────────────────────────────────▶│
      │                                │    store session data            │
      │                                │◀────────────────────────────────│
      │ 4. Set-Cookie: sid=abc123      │                                 │
      │◀───────────────────────────────│                                 │
      │                                │                                 │
      │ 5. GET /api/users/me           │                                 │
      │    Cookie: sid=abc123          │                                 │
      │───────────────────────────────▶│                                 │
      │                                │ 6. Lookup session                │
      │                                │────────────────────────────────▶│
      │                                │◀────────────────────────────────│
      │ 7. Response with user data     │                                 │
      │◀───────────────────────────────│                                 │
```

### Cookie Configuration

```typescript
// Session cookie best practices
app.use(session({
  name: 'sid',                    // Custom name (not 'connect.sid')
  secret: process.env.SESSION_SECRET!,
  resave: false,
  saveUninitialized: false,
  cookie: {
    httpOnly: true,               // Prevent XSS access to cookie
    secure: process.env.NODE_ENV === 'production',  // HTTPS only in prod
    sameSite: 'lax',              // CSRF protection
    maxAge: 24 * 60 * 60 * 1000, // 24 hours
    domain: '.example.com',       // Allow subdomains
    path: '/',
  },
  store: new RedisStore({ client: redisClient }),  // Server-side storage
}));
```

---

## Basic Authentication

### How It Works

Credentials (username:password) are Base64-encoded and sent in the Authorization header.

```http
# Encode: base64("username:password") → "dXNlcm5hbWU6cGFzc3dvcmQ="
GET /api/data HTTP/1.1
Authorization: Basic dXNlcm5hbWU6cGFzc3dvcmQ=
```

**IMPORTANT:** Base64 is encoding, NOT encryption. Always use Basic Auth over HTTPS.

### Testing

```typescript
describe('Basic Authentication', () => {
  it('should authenticate with valid credentials', async () => {
    const credentials = Buffer.from('admin:secret').toString('base64');
    await request(app)
      .get('/api/data')
      .set('Authorization', `Basic ${credentials}`)
      .expect(200);
  });

  it('should reject invalid credentials', async () => {
    const credentials = Buffer.from('admin:wrong').toString('base64');
    const res = await request(app)
      .get('/api/data')
      .set('Authorization', `Basic ${credentials}`)
      .expect(401);

    expect(res.headers['www-authenticate']).toContain('Basic');
  });
});
```

---

## HMAC Signatures

### How It Works

The client creates a signature by hashing the request details with a shared secret.
The server recreates the signature and compares it.

```
Client:
  1. Build canonical string: method + path + timestamp + body_hash
  2. Sign with HMAC-SHA256 using shared secret
  3. Send signature in header

Server:
  1. Rebuild canonical string from received request
  2. Sign with same secret
  3. Compare signatures (timing-safe)
  4. Check timestamp is within tolerance (prevent replay)
```

### Implementation

```typescript
import { createHmac, timingSafeEqual } from 'crypto';

// Client-side: Generate signature
function signRequest(
  method: string,
  path: string,
  body: string | null,
  secret: string,
  timestamp: string
): string {
  const bodyHash = body
    ? createHmac('sha256', secret).update(body).digest('hex')
    : '';

  const canonical = [method.toUpperCase(), path, timestamp, bodyHash].join('\n');

  return createHmac('sha256', secret).update(canonical).digest('hex');
}

// Server-side: Verify signature
function verifySignature(req: Request, secret: string): boolean {
  const receivedSignature = req.headers['x-signature'] as string;
  const timestamp = req.headers['x-timestamp'] as string;

  if (!receivedSignature || !timestamp) return false;

  // Check timestamp freshness (prevent replay attacks)
  const requestTime = parseInt(timestamp, 10);
  const now = Math.floor(Date.now() / 1000);
  const tolerance = 300; // 5 minutes

  if (Math.abs(now - requestTime) > tolerance) return false;

  // Rebuild and compare signature
  const body = req.body ? JSON.stringify(req.body) : null;
  const expectedSignature = signRequest(
    req.method,
    req.path,
    body,
    secret,
    timestamp
  );

  // Timing-safe comparison (prevents timing attacks)
  const expected = Buffer.from(expectedSignature, 'hex');
  const received = Buffer.from(receivedSignature, 'hex');

  if (expected.length !== received.length) return false;
  return timingSafeEqual(expected, received);
}
```

### Common Use: Webhook Verification

```typescript
// Verify Stripe webhook signature
function verifyStripeWebhook(req: Request): boolean {
  const signature = req.headers['stripe-signature'] as string;
  const endpointSecret = process.env.STRIPE_WEBHOOK_SECRET!;

  try {
    stripe.webhooks.constructEvent(req.body, signature, endpointSecret);
    return true;
  } catch {
    return false;
  }
}

// Verify GitHub webhook signature
function verifyGithubWebhook(req: Request): boolean {
  const signature = req.headers['x-hub-signature-256'] as string;
  const secret = process.env.GITHUB_WEBHOOK_SECRET!;

  const expected = 'sha256=' +
    createHmac('sha256', secret).update(JSON.stringify(req.body)).digest('hex');

  return timingSafeEqual(Buffer.from(signature), Buffer.from(expected));
}
```

---

## Mutual TLS (mTLS)

### How It Works

Both client and server present X.509 certificates during the TLS handshake.
The server verifies the client's certificate, and the client verifies the server's.

```
Standard TLS: Client verifies Server only
  Client ←─── Server Certificate ──── Server
  Client ──── "OK, I trust you" ─────▶ Server

Mutual TLS: Both sides verify
  Client ←─── Server Certificate ──── Server
  Client ──── Client Certificate ────▶ Server
  Client ←─── "OK, I trust you" ──── Server
```

### Configuration

```typescript
// Server-side mTLS (Node.js)
import https from 'https';
import { readFileSync } from 'fs';

const server = https.createServer({
  key: readFileSync('server-key.pem'),
  cert: readFileSync('server-cert.pem'),
  ca: readFileSync('client-ca.pem'),         // CA that signed client certs
  requestCert: true,                          // Request client certificate
  rejectUnauthorized: true,                   // Reject if no valid client cert
}, app);
```

```bash
# Client-side mTLS (curl)
curl --cert client.pem --key client-key.pem \
     --cacert server-ca.pem \
     https://api.example.com/data
```

---

## SAML

### Overview

Security Assertion Markup Language — XML-based authentication primarily used in enterprise
Single Sign-On (SSO). The SAML assertion is like a JWT but in XML format.

SAML is less common in modern APIs but still widely used in enterprise environments,
especially with identity providers like Okta, Azure AD, and OneLogin.

### Flow (SP-Initiated)

```
1. User visits Service Provider (your app)
2. SP generates SAML AuthnRequest, redirects to Identity Provider (IdP)
3. User authenticates at IdP
4. IdP generates SAML Response with Assertion
5. IdP redirects back to SP's ACS (Assertion Consumer Service) URL
6. SP validates the assertion and creates a session
```

---

## Rate Limiting by Auth Type

Different auth types typically have different rate limits because they represent
different trust levels and use cases.

| Auth Type | Typical Rate Limit | Reason |
|-----------|-------------------|--------|
| No auth (public) | 20-60 req/min per IP | Lowest trust, highest abuse risk |
| API Key | 100-1000 req/min per key | Identifiable, accountable |
| Bearer Token (user) | 100-300 req/min per user | Authenticated human user |
| Bearer Token (service) | 1000-10000 req/min | Machine-to-machine, higher needs |
| OAuth (3rd party) | 50-500 req/min per app | Delegated access, quota per app |
| mTLS (internal) | 10000+ req/min | Trusted internal service |
| Admin endpoints | 30-100 req/min | Sensitive operations, lower volume expected |
| Auth endpoints (login) | 5-10 req/min per IP | Brute force protection |

### Testing Rate Limits

```typescript
describe('Rate Limiting', () => {
  it('should enforce rate limits on auth endpoints', async () => {
    const requests = Array.from({ length: 15 }, () =>
      request(app)
        .post('/api/auth/login')
        .send({ email: 'user@test.com', password: 'wrong' })
    );

    const responses = await Promise.all(requests);
    const rateLimited = responses.filter(r => r.status === 429);

    expect(rateLimited.length).toBeGreaterThan(0);
    expect(rateLimited[0].headers['retry-after']).toBeDefined();
  });

  it('should have separate limits per auth type', async () => {
    // API key should have higher limit than unauthenticated
    const apiKeyRequests = Array.from({ length: 100 }, () =>
      request(app)
        .get('/api/products')
        .set('X-API-Key', 'sk_test_valid')
    );

    const publicRequests = Array.from({ length: 100 }, () =>
      request(app).get('/api/products')
    );

    const apiKeyResults = await Promise.all(apiKeyRequests);
    const publicResults = await Promise.all(publicRequests);

    const apiKey429s = apiKeyResults.filter(r => r.status === 429).length;
    const public429s = publicResults.filter(r => r.status === 429).length;

    // Public should hit rate limit sooner
    expect(public429s).toBeGreaterThan(apiKey429s);
  });
});
```

---

## Security Best Practices

### Token Storage

| Platform | Access Token | Refresh Token |
|----------|-------------|---------------|
| SPA (browser) | Memory only (not localStorage!) | HttpOnly Secure cookie or memory |
| Mobile app | Secure storage (Keychain/Keystore) | Secure storage |
| Server-side | Environment variable or secret manager | Database (hashed) |
| CLI tool | OS credential store | OS credential store |

### Password Storage

```typescript
import bcrypt from 'bcrypt';

// Hashing (never store plaintext passwords!)
const SALT_ROUNDS = 12; // Higher = slower = more secure (but slower login)

async function hashPassword(password: string): Promise<string> {
  return bcrypt.hash(password, SALT_ROUNDS);
}

async function verifyPassword(password: string, hash: string): Promise<boolean> {
  return bcrypt.compare(password, hash);
}

// Password requirements
const PASSWORD_REGEX = /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&])[A-Za-z\d@$!%*?&]{8,128}$/;
```

### Common Vulnerabilities

| Vulnerability | Risk | Prevention |
|--------------|------|------------|
| Token in URL | Leaks via referer header, logs | Use Authorization header |
| Token in localStorage | XSS can steal it | Store in memory or HttpOnly cookie |
| No token expiration | Stolen token is valid forever | Set short exp (1h for access tokens) |
| No token rotation | Refresh token reuse undetected | Rotate refresh tokens on use |
| JWT alg:none | Bypasses signature verification | Explicitly set allowed algorithms |
| JWT key confusion | RS256 treated as HS256 | Validate algorithm before verifying |
| CORS wildcard + creds | Allows any origin with cookies | Never use * with credentials |
| Missing CSRF | Forged requests with cookies | Use SameSite cookies or CSRF tokens |
| Timing attacks | Password/token comparison leaks info | Use timingSafeEqual |
| Brute force login | Unlimited login attempts | Rate limit auth endpoints |
| Session fixation | Attacker sets session before login | Regenerate session ID after login |

---

## Testing Patterns

### Authentication Test Matrix

For every protected endpoint, test these scenarios:

```
Auth Testing Matrix:
  □ No auth header → 401
  □ Empty auth header → 401
  □ Wrong auth scheme (Basic instead of Bearer) → 401
  □ Invalid token format → 401
  □ Expired token → 401
  □ Revoked token → 401
  □ Token with wrong signature → 401
  □ Token with "none" algorithm → 401
  □ Token for deleted user → 401
  □ Valid token, insufficient role → 403
  □ Valid token, valid role → 200
  □ Valid token, own resource → 200
  □ Valid token, someone else's resource → 403
  □ Admin token, any resource → 200
```

### Authorization Test Matrix

```
Authorization Testing Matrix:
  □ User can access own resources → 200
  □ User cannot access other user's resources → 403
  □ User cannot access admin endpoints → 403
  □ Admin can access all user resources → 200
  □ Admin can access admin endpoints → 200
  □ Service token has correct scopes → 200
  □ Service token denied for unauthorized scopes → 403
  □ Read-only token cannot write → 403
  □ Write token can read and write → 200
```

### Security Headers Test

```typescript
describe('Security Headers', () => {
  it('should not expose server technology', async () => {
    const res = await request(app).get('/api/products').expect(200);
    expect(res.headers['x-powered-by']).toBeUndefined();
    expect(res.headers['server']).toBeUndefined();
  });

  it('should include security headers', async () => {
    const res = await request(app).get('/api/products').expect(200);
    expect(res.headers['x-content-type-options']).toBe('nosniff');
    expect(res.headers['x-frame-options']).toMatch(/DENY|SAMEORIGIN/);
  });

  it('should set HSTS header', async () => {
    const res = await request(app).get('/api/products').expect(200);
    if (process.env.NODE_ENV === 'production') {
      expect(res.headers['strict-transport-security']).toContain('max-age=');
    }
  });

  it('should not allow CORS from unauthorized origins', async () => {
    const res = await request(app)
      .options('/api/products')
      .set('Origin', 'https://evil.example.com')
      .set('Access-Control-Request-Method', 'GET');

    expect(res.headers['access-control-allow-origin']).not.toBe('https://evil.example.com');
  });
});
```

---

## Multi-Factor Authentication (MFA)

### TOTP (Time-Based One-Time Password)

The most common MFA method for APIs. Uses RFC 6238 to generate time-based codes.

```typescript
// TOTP Setup flow
// Step 1: Generate secret for user
app.post('/api/auth/mfa/setup', auth, async (req, res) => {
  const secret = authenticator.generateSecret();

  // Store secret (encrypted) in database
  await db.user.update({
    where: { id: req.user.id },
    data: { mfaSecret: encrypt(secret), mfaPending: true },
  });

  // Generate QR code URI for authenticator apps
  const otpauthUri = authenticator.keyuri(
    req.user.email,
    'MyApp',
    secret
  );

  res.json({
    secret, // Show once — user should save backup codes
    qrCodeUri: otpauthUri,
    backupCodes: generateBackupCodes(8), // 8 single-use backup codes
  });
});

// Step 2: Verify TOTP code to activate MFA
app.post('/api/auth/mfa/verify', auth, async (req, res) => {
  const { code } = req.body;
  const user = await db.user.findUnique({ where: { id: req.user.id } });

  const secret = decrypt(user.mfaSecret);
  const isValid = authenticator.verify({ token: code, secret });

  if (!isValid) {
    return res.status(400).json({
      error: 'Invalid Code',
      message: 'The verification code is incorrect. Try again.',
    });
  }

  await db.user.update({
    where: { id: req.user.id },
    data: { mfaEnabled: true, mfaPending: false },
  });

  res.json({ message: 'MFA enabled successfully' });
});

// Step 3: Login with MFA
app.post('/api/auth/login', async (req, res) => {
  const { email, password, mfaCode } = req.body;

  const user = await validateCredentials(email, password);
  if (!user) return res.status(401).json({ error: 'Invalid credentials' });

  if (user.mfaEnabled) {
    if (!mfaCode) {
      return res.status(200).json({
        requiresMfa: true,
        mfaToken: generateMfaToken(user.id), // Short-lived token for MFA step
      });
    }

    const secret = decrypt(user.mfaSecret);
    const isValid = authenticator.verify({ token: mfaCode, secret });

    if (!isValid) {
      // Check backup codes
      const backupValid = await checkBackupCode(user.id, mfaCode);
      if (!backupValid) {
        return res.status(401).json({ error: 'Invalid MFA code' });
      }
    }
  }

  const tokens = generateTokens(user);
  res.json(tokens);
});
```

### WebAuthn / Passkeys

Modern passwordless authentication using public-key cryptography.

```typescript
// WebAuthn Registration (simplified)
app.post('/api/auth/webauthn/register-options', auth, async (req, res) => {
  const options = await generateRegistrationOptions({
    rpName: 'My App',
    rpID: 'example.com',
    userID: req.user.id,
    userName: req.user.email,
    authenticatorSelection: {
      residentKey: 'preferred',
      userVerification: 'preferred',
    },
  });

  // Store challenge for verification
  await db.user.update({
    where: { id: req.user.id },
    data: { currentChallenge: options.challenge },
  });

  res.json(options);
});

app.post('/api/auth/webauthn/register-verify', auth, async (req, res) => {
  const user = await db.user.findUnique({ where: { id: req.user.id } });

  const verification = await verifyRegistrationResponse({
    response: req.body,
    expectedChallenge: user.currentChallenge,
    expectedOrigin: 'https://example.com',
    expectedRPID: 'example.com',
  });

  if (verification.verified) {
    // Store credential for future authentication
    await db.credential.create({
      data: {
        userId: req.user.id,
        credentialId: verification.registrationInfo.credentialID,
        publicKey: verification.registrationInfo.credentialPublicKey,
        counter: verification.registrationInfo.counter,
      },
    });

    res.json({ verified: true });
  } else {
    res.status(400).json({ error: 'Verification failed' });
  }
});
```

### MFA Testing

```typescript
describe('Multi-Factor Authentication', () => {
  it('should set up TOTP MFA', async () => {
    const token = await getAuthToken('user');

    const setup = await authApi(token)
      .post('/api/auth/mfa/setup')
      .expect(200);

    expect(setup.body.secret).toBeDefined();
    expect(setup.body.qrCodeUri).toContain('otpauth://totp/');
    expect(setup.body.backupCodes).toHaveLength(8);
  });

  it('should require MFA code after enabling', async () => {
    // First login — should indicate MFA required
    const login = await request(app)
      .post('/api/auth/login')
      .send({ email: 'mfa-user@test.com', password: 'Password123!' })
      .expect(200);

    expect(login.body.requiresMfa).toBe(true);
    expect(login.body.mfaToken).toBeDefined();
    expect(login.body.accessToken).toBeUndefined(); // No access token yet
  });

  it('should accept valid TOTP code', async () => {
    const mfaCode = authenticator.generate(testMfaSecret);

    const login = await request(app)
      .post('/api/auth/login')
      .send({
        email: 'mfa-user@test.com',
        password: 'Password123!',
        mfaCode,
      })
      .expect(200);

    expect(login.body.accessToken).toBeDefined();
  });

  it('should reject expired TOTP code', async () => {
    // TOTP codes are valid for 30 seconds — use an old one
    const oldCode = '000000'; // Will be invalid

    await request(app)
      .post('/api/auth/login')
      .send({
        email: 'mfa-user@test.com',
        password: 'Password123!',
        mfaCode: oldCode,
      })
      .expect(401);
  });

  it('should accept backup code', async () => {
    const backupCode = testBackupCodes[0];

    const login = await request(app)
      .post('/api/auth/login')
      .send({
        email: 'mfa-user@test.com',
        password: 'Password123!',
        mfaCode: backupCode,
      })
      .expect(200);

    expect(login.body.accessToken).toBeDefined();
  });

  it('should invalidate backup code after use', async () => {
    const backupCode = testBackupCodes[0]; // Already used

    await request(app)
      .post('/api/auth/login')
      .send({
        email: 'mfa-user@test.com',
        password: 'Password123!',
        mfaCode: backupCode,
      })
      .expect(401);
  });
});
```

---

## Role-Based Access Control (RBAC)

### RBAC Implementation

```typescript
// Role and permission definitions
const ROLES = {
  USER: {
    permissions: [
      'users:read:own',
      'orders:read:own',
      'orders:create',
      'products:read',
      'cart:manage:own',
    ],
  },
  MODERATOR: {
    permissions: [
      'users:read:own',
      'users:read:any',
      'orders:read:own',
      'orders:create',
      'products:read',
      'products:create',
      'products:update',
      'reviews:moderate',
      'cart:manage:own',
    ],
  },
  ADMIN: {
    permissions: [
      'users:read:any',
      'users:create',
      'users:update:any',
      'users:delete:any',
      'orders:read:any',
      'orders:update:any',
      'orders:create',
      'products:read',
      'products:create',
      'products:update',
      'products:delete',
      'reviews:moderate',
      'reviews:delete',
      'settings:manage',
      'analytics:view',
      'cart:manage:own',
    ],
  },
};

// Permission check middleware
function requirePermission(permission: string) {
  return (req: Request, res: Response, next: NextFunction) => {
    const userRole = req.user?.role;

    if (!userRole || !ROLES[userRole]) {
      return res.status(403).json({
        error: 'Forbidden',
        message: 'Insufficient permissions',
        statusCode: 403,
      });
    }

    const userPermissions = ROLES[userRole].permissions;

    // Check exact permission or wildcard
    const hasPermission = userPermissions.some(p => {
      if (p === permission) return true;
      // Check wildcard: 'users:*' matches 'users:read:any'
      const parts = p.split(':');
      const reqParts = permission.split(':');
      return parts.every((part, i) => part === '*' || part === reqParts[i]);
    });

    if (!hasPermission) {
      return res.status(403).json({
        error: 'Forbidden',
        message: `Requires permission: ${permission}`,
        statusCode: 403,
      });
    }

    next();
  };
}

// Usage
app.get('/api/users', auth, requirePermission('users:read:any'), listUsers);
app.delete('/api/users/:id', auth, requirePermission('users:delete:any'), deleteUser);
app.get('/api/analytics', auth, requirePermission('analytics:view'), getAnalytics);
```

### RBAC Testing

```typescript
describe('Role-Based Access Control', () => {
  let userToken: string;
  let moderatorToken: string;
  let adminToken: string;

  beforeAll(async () => {
    userToken = await getAuthToken('user');
    moderatorToken = await getAuthToken('moderator');
    adminToken = await getAuthToken('admin');
  });

  const accessMatrix = [
    // [endpoint, method, user, moderator, admin]
    ['GET /api/users', 'get', 403, 200, 200],
    ['POST /api/users', 'post', 403, 403, 201],
    ['DELETE /api/users/:id', 'delete', 403, 403, 204],
    ['GET /api/products', 'get', 200, 200, 200],
    ['POST /api/products', 'post', 403, 201, 201],
    ['DELETE /api/products/:id', 'delete', 403, 403, 204],
    ['GET /api/analytics', 'get', 403, 403, 200],
    ['GET /api/orders (own)', 'get', 200, 200, 200],
  ];

  for (const [endpoint, method, userExpected, modExpected, adminExpected] of accessMatrix) {
    it(`${endpoint} — user:${userExpected}, mod:${modExpected}, admin:${adminExpected}`, async () => {
      // Test each role against the endpoint
      // (simplified — actual paths need to be constructed)
    });
  }

  it('should not allow privilege escalation via request body', async () => {
    // User tries to make themselves admin
    await authApi(userToken)
      .patch('/api/users/me')
      .send({ role: 'ADMIN' })
      .expect(200); // Request succeeds but role doesn't change

    const profile = await authApi(userToken).get('/api/users/me').expect(200);
    expect(profile.body.role).toBe('USER');
  });
});
```

---

## Scope-Based Authorization (OAuth Scopes)

### Scope Definitions

```
Scope definitions for API access:

users:read         — Read user profiles
users:write        — Create and update users
users:delete       — Delete user accounts
products:read      — Read product catalog
products:write     — Create and update products
orders:read        — Read order history
orders:write       — Create and manage orders
admin:read         — Read admin dashboards
admin:write        — Modify admin settings
```

### Scope Validation

```typescript
// Middleware to check OAuth scopes
function requireScope(...requiredScopes: string[]) {
  return (req: Request, res: Response, next: NextFunction) => {
    const tokenScopes = req.auth?.scopes || [];

    const hasAllScopes = requiredScopes.every(scope =>
      tokenScopes.includes(scope)
    );

    if (!hasAllScopes) {
      const missing = requiredScopes.filter(s => !tokenScopes.includes(s));
      return res.status(403).json({
        error: 'Insufficient Scope',
        message: `Required scopes: ${missing.join(', ')}`,
        statusCode: 403,
      });
    }

    next();
  };
}

// Usage
app.get('/api/users', auth, requireScope('users:read'), listUsers);
app.post('/api/users', auth, requireScope('users:write'), createUser);
app.delete('/api/users/:id', auth, requireScope('users:delete'), deleteUser);
```

### Scope Testing

```typescript
describe('OAuth Scope Authorization', () => {
  it('should allow access with correct scope', async () => {
    const token = await getTokenWithScopes(['users:read']);
    await request(app)
      .get('/api/users')
      .set('Authorization', `Bearer ${token}`)
      .expect(200);
  });

  it('should deny access with insufficient scope', async () => {
    const token = await getTokenWithScopes(['users:read']); // Read-only
    const res = await request(app)
      .post('/api/users')
      .set('Authorization', `Bearer ${token}`)
      .send({ email: 'test@example.com', name: 'Test' })
      .expect(403);

    expect(res.body.error).toBe('Insufficient Scope');
    expect(res.body.message).toContain('users:write');
  });

  it('should allow access with multiple scopes', async () => {
    const token = await getTokenWithScopes(['users:read', 'users:write']);
    await request(app)
      .post('/api/users')
      .set('Authorization', `Bearer ${token}`)
      .send({ email: 'test@example.com', name: 'Test' })
      .expect(201);
  });

  it('should not allow scope escalation', async () => {
    // Token with limited scopes shouldn't be able to access admin resources
    const token = await getTokenWithScopes(['users:read']);
    await request(app)
      .get('/api/admin/settings')
      .set('Authorization', `Bearer ${token}`)
      .expect(403);
  });
});
```

---

## Token Revocation

### Token Blocklist

```typescript
// Redis-based token blocklist for JWT revocation
import Redis from 'ioredis';

const redis = new Redis(process.env.REDIS_URL);

// Add token to blocklist (on logout or compromise)
async function revokeToken(jti: string, expiresIn: number): Promise<void> {
  // Store in Redis with TTL matching token expiry
  // (no need to keep it longer — expired tokens are already rejected)
  await redis.set(`blocklist:${jti}`, '1', 'EX', expiresIn);
}

// Check if token is revoked
async function isTokenRevoked(jti: string): Promise<boolean> {
  const result = await redis.get(`blocklist:${jti}`);
  return result !== null;
}

// Middleware integration
async function checkTokenRevocation(req: Request, res: Response, next: NextFunction) {
  const jti = req.auth?.jti;

  if (jti && await isTokenRevoked(jti)) {
    return res.status(401).json({
      error: 'Unauthorized',
      message: 'Token has been revoked',
      statusCode: 401,
    });
  }

  next();
}
```

### Testing Token Revocation

```typescript
describe('Token Revocation', () => {
  it('should reject revoked access tokens', async () => {
    // Login
    const login = await request(app)
      .post('/api/auth/login')
      .send({ email: 'user@test.com', password: 'Password123!' })
      .expect(200);

    const { accessToken } = login.body;

    // Token works before revocation
    await request(app)
      .get('/api/users/me')
      .set('Authorization', `Bearer ${accessToken}`)
      .expect(200);

    // Logout (revokes token)
    await request(app)
      .post('/api/auth/logout')
      .set('Authorization', `Bearer ${accessToken}`)
      .expect(200);

    // Token should be rejected after revocation
    await request(app)
      .get('/api/users/me')
      .set('Authorization', `Bearer ${accessToken}`)
      .expect(401);
  });
});
```
