# JWT Security Reference

Complete reference for JSON Web Token security — structure, signing algorithms, claim validation, key rotation, storage strategies, and common vulnerabilities with mitigations.

---

## JWT Structure

```
A JWT consists of three Base64URL-encoded parts separated by dots:

HEADER.PAYLOAD.SIGNATURE

┌─────────────────────────────────────────────────────────────────┐
│                         JWT Structure                           │
├───────────────┬─────────────────┬───────────────────────────────┤
│    Header     │    Payload      │         Signature             │
│  (metadata)   │   (claims)      │    (cryptographic proof)      │
├───────────────┼─────────────────┼───────────────────────────────┤
│ {             │ {               │ RS256(                        │
│   "alg":"RS256"│   "iss":"auth" │   base64url(header) + "." +  │
│   "typ":"JWT" │   "sub":"u123" │   base64url(payload),         │
│   "kid":"k1"  │   "exp":169... │   privateKey                  │
│ }             │   "iat":169... │ )                              │
│               │   "roles":["a"]│                                │
│               │ }               │                                │
└───────────────┴─────────────────┴───────────────────────────────┘

Total size: typically 300-800 bytes
Max recommended: < 4KB (cookie size limit)
```

---

## Signing Algorithms

### Algorithm Comparison

```
┌─────────────┬─────────────────────┬───────┬────────────────────────┐
│ Algorithm   │ Type                │ Keys  │ Best For               │
├─────────────┼─────────────────────┼───────┼────────────────────────┤
│ RS256       │ RSA + SHA-256       │ Asym  │ Most production apps   │
│ RS384       │ RSA + SHA-384       │ Asym  │ Higher security needs  │
│ RS512       │ RSA + SHA-512       │ Asym  │ Maximum RSA security   │
├─────────────┼─────────────────────┼───────┼────────────────────────┤
│ ES256       │ ECDSA + P-256       │ Asym  │ Mobile, IoT (smaller  │
│ ES384       │ ECDSA + P-384       │ Asym  │ keys, faster verify)  │
│ ES512       │ ECDSA + P-521       │ Asym  │ High security + small │
├─────────────┼─────────────────────┼───────┼────────────────────────┤
│ PS256       │ RSA-PSS + SHA-256   │ Asym  │ Modern RSA (better    │
│ PS384       │ RSA-PSS + SHA-384   │ Asym  │ than PKCS#1 v1.5)     │
│ PS512       │ RSA-PSS + SHA-512   │ Asym  │                       │
├─────────────┼─────────────────────┼───────┼────────────────────────┤
│ HS256       │ HMAC + SHA-256      │ Sym   │ Single-service only    │
│ HS384       │ HMAC + SHA-384      │ Sym   │ (both sides share key) │
│ HS512       │ HMAC + SHA-512      │ Sym   │                       │
├─────────────┼─────────────────────┼───────┼────────────────────────┤
│ EdDSA       │ Ed25519/Ed448       │ Asym  │ Cutting edge, fastest  │
├─────────────┼─────────────────────┼───────┼────────────────────────┤
│ none        │ No signature        │ None  │ NEVER use in production│
└─────────────┴─────────────────────┴───────┴────────────────────────┘

Recommendations:
1. RS256 — safest default, widely supported, JWKS-compatible
2. ES256 — if you need smaller tokens or faster verification
3. HS256 — ONLY for single-service scenarios with a strong secret (256+ bits)
4. NEVER use 'none' algorithm — always verify signatures
```

### Key Size Requirements

```
┌─────────────┬───────────────────┬──────────────────────────────┐
│ Algorithm   │ Minimum Key Size  │ Recommended Key Size         │
├─────────────┼───────────────────┼──────────────────────────────┤
│ RS256/384/512│ 2048 bits        │ 2048 bits (4096 for highest) │
│ ES256       │ P-256 (256 bits)  │ P-256 (fixed)                │
│ ES384       │ P-384 (384 bits)  │ P-384 (fixed)                │
│ ES512       │ P-521 (521 bits)  │ P-521 (fixed)                │
│ HS256       │ 256 bits (32 B)   │ 512 bits (64 bytes)          │
│ HS384       │ 384 bits (48 B)   │ 512 bits (64 bytes)          │
│ HS512       │ 512 bits (64 B)   │ 512 bits (64 bytes)          │
│ EdDSA       │ 256 bits          │ Ed25519 (256 bits)           │
└─────────────┴───────────────────┴──────────────────────────────┘

IMPORTANT: For HMAC (HS*), the secret MUST be at least as long as the hash output.
A short string like "my-secret" is NOT acceptable — use crypto.randomBytes(64).
```

---

## Standard Claims

### Registered Claims (RFC 7519)

```
┌──────┬────────────────┬──────────────────────────────────────────┐
│ Claim│ Name           │ Description + Validation                 │
├──────┼────────────────┼──────────────────────────────────────────┤
│ iss  │ Issuer         │ Who created the token.                   │
│      │                │ MUST match expected issuer URL.           │
│      │                │ Example: "https://auth.example.com"       │
├──────┼────────────────┼──────────────────────────────────────────┤
│ sub  │ Subject        │ Who the token is about (user ID).        │
│      │                │ Use opaque IDs, not emails.              │
│      │                │ Example: "user-uuid-123"                  │
├──────┼────────────────┼──────────────────────────────────────────┤
│ aud  │ Audience       │ Intended recipient(s) of the token.      │
│      │                │ MUST match your API's identifier.         │
│      │                │ Example: "https://api.example.com"        │
├──────┼────────────────┼──────────────────────────────────────────┤
│ exp  │ Expiration     │ Token expiry (Unix timestamp).           │
│      │                │ MUST be validated on every request.       │
│      │                │ Access tokens: 5-15 minutes.             │
│      │                │ ID tokens: 5-60 minutes.                 │
├──────┼────────────────┼──────────────────────────────────────────┤
│ nbf  │ Not Before     │ Token not valid before this time.        │
│      │                │ Usually same as iat.                     │
├──────┼────────────────┼──────────────────────────────────────────┤
│ iat  │ Issued At      │ When the token was created.              │
│      │                │ Useful for detecting old tokens.         │
├──────┼────────────────┼──────────────────────────────────────────┤
│ jti  │ JWT ID         │ Unique token identifier.                 │
│      │                │ Used for revocation (blacklisting).      │
│      │                │ Example: UUID v4                         │
└──────┴────────────────┴──────────────────────────────────────────┘
```

### Custom Claims Best Practices

```javascript
// Good: minimal, flat custom claims
{
  "sub": "user-123",
  "roles": ["editor", "viewer"],
  "scope": "read write",
  "org_id": "org-456",
  "plan": "pro"
}

// Bad: too much data, nested objects, sensitive info
{
  "sub": "user-123",
  "email": "user@example.com",        // PII — avoid if possible
  "address": { "street": "123 Main" }, // Never put PII in JWT
  "ssn": "123-45-6789",               // NEVER put sensitive data
  "permissions": ["document:read", "document:write", "user:read",
    "user:write", "billing:read", "billing:manage", "admin:access"],
  // Too many permissions — use roles instead
}

Rules for custom claims:
1. Keep tokens small (< 1KB payload)
2. Use roles, not individual permissions
3. Avoid PII (name, email, address) — fetch from API instead
4. Use namespaced claims for custom data: "https://myapp.com/roles"
5. Never include secrets, passwords, or credit card numbers
6. Remember: JWT payload is Base64-encoded, NOT encrypted
```

---

## Token Validation

### Complete Validation Checklist

```
Step-by-step JWT validation:

1. Parse the token (split by dots, decode Base64URL)
   □ Token has exactly 3 parts
   □ Each part is valid Base64URL

2. Verify the header
   □ "alg" is in your whitelist (e.g., ["RS256", "ES256"])
   □ "alg" is NOT "none"
   □ "typ" is "JWT" (if present)
   □ "kid" matches a known key (for key rotation)

3. Verify the signature
   □ Use the correct public key (from JWKS or local store)
   □ Signature is cryptographically valid
   □ Use constant-time comparison

4. Validate registered claims
   □ "iss" matches your expected issuer
   □ "aud" matches your API's identifier
   □ "exp" is in the future (with small clock skew tolerance, ~5s)
   □ "nbf" is in the past (if present)
   □ "iat" is reasonable (not too far in the past)
   □ "jti" has not been used before (if checking for replay)

5. Validate custom claims
   □ Required claims are present (e.g., "roles", "scope")
   □ Claim values are valid (e.g., roles exist in your system)

6. Context validation
   □ Token type matches expected use (access vs refresh)
   □ Scopes/permissions are sufficient for the requested action
```

### Validation Implementation

```javascript
// auth/jwt-validator.js — Complete JWT validation
const jose = require('jose');

class JWTValidator {
  constructor(options) {
    this.issuer = options.issuer;
    this.audience = options.audience;
    this.algorithms = options.algorithms || ['RS256', 'ES256'];
    this.clockTolerance = options.clockTolerance || 5; // seconds
    this.maxTokenAge = options.maxTokenAge || '24h';
    this.jwksUrl = options.jwksUrl;
    this._jwks = null;
  }

  async getJWKS() {
    if (!this._jwks) {
      this._jwks = jose.createRemoteJWKSet(new URL(this.jwksUrl), {
        cooldownDuration: 30000,  // Cache JWKS for 30 seconds minimum
        cacheMaxAge: 600000,      // Max cache: 10 minutes
      });
    }
    return this._jwks;
  }

  async validate(token) {
    // Step 1: Decode header first (without verification)
    const header = jose.decodeProtectedHeader(token);

    // Step 2: Check algorithm whitelist
    if (!this.algorithms.includes(header.alg)) {
      throw new JWTValidationError(
        `Algorithm ${header.alg} not allowed. Expected: ${this.algorithms.join(', ')}`
      );
    }

    // Step 3: Get the signing key
    const jwks = await this.getJWKS();

    // Step 4: Verify signature + standard claims
    try {
      const { payload, protectedHeader } = await jose.jwtVerify(token, jwks, {
        issuer: this.issuer,
        audience: this.audience,
        clockTolerance: this.clockTolerance,
        maxTokenAge: this.maxTokenAge,
        algorithms: this.algorithms,
        requiredClaims: ['sub', 'iat', 'exp'],
      });

      return {
        valid: true,
        payload,
        header: protectedHeader,
      };
    } catch (error) {
      throw new JWTValidationError(this.mapError(error));
    }
  }

  mapError(error) {
    const errorMap = {
      'ERR_JWT_EXPIRED': 'Token has expired',
      'ERR_JWT_CLAIM_VALIDATION_FAILED': `Claim validation failed: ${error.message}`,
      'ERR_JWS_SIGNATURE_VERIFICATION_FAILED': 'Invalid signature',
      'ERR_JWKS_NO_MATCHING_KEY': 'No matching key found (key may have been rotated)',
      'ERR_JWT_INVALID': 'Token is malformed',
    };
    return errorMap[error.code] || `JWT validation failed: ${error.message}`;
  }
}

class JWTValidationError extends Error {
  constructor(message) {
    super(message);
    this.name = 'JWTValidationError';
    this.statusCode = 401;
  }
}

module.exports = { JWTValidator, JWTValidationError };
```

### Express Middleware

```javascript
// middleware/jwt-auth.js — JWT authentication middleware
function jwtAuthMiddleware(validator) {
  return async (req, res, next) => {
    const authHeader = req.headers.authorization;

    if (!authHeader || !authHeader.startsWith('Bearer ')) {
      return res.status(401).json({
        error: 'Missing or invalid Authorization header',
        hint: 'Expected: Authorization: Bearer <token>',
      });
    }

    const token = authHeader.substring(7);

    try {
      const result = await validator.validate(token);
      req.user = {
        id: result.payload.sub,
        roles: result.payload.roles || [],
        scope: result.payload.scope,
        tokenId: result.payload.jti,
      };
      next();
    } catch (error) {
      if (error.message.includes('expired')) {
        return res.status(401).json({
          error: 'Token expired',
          code: 'TOKEN_EXPIRED',
        });
      }
      return res.status(401).json({
        error: 'Invalid token',
        code: 'TOKEN_INVALID',
      });
    }
  };
}

module.exports = { jwtAuthMiddleware };
```

---

## Key Rotation

### JWKS (JSON Web Key Set)

```
JWKS Endpoint: GET /.well-known/jwks.json

Response:
{
  "keys": [
    {
      "kty": "RSA",
      "kid": "key-2024-01",         ← Current signing key
      "use": "sig",
      "alg": "RS256",
      "n": "0vx7agoebGc...",       ← RSA modulus (public)
      "e": "AQAB"                   ← RSA exponent (public)
    },
    {
      "kty": "RSA",
      "kid": "key-2023-12",         ← Previous key (still validates)
      "use": "sig",
      "alg": "RS256",
      "n": "1b7aHkx...",
      "e": "AQAB"
    }
  ]
}
```

### Key Rotation Strategy

```
Timeline for key rotation:

Day 0:      Generate Key A, sign with Key A
            JWKS: [Key A]

Day 30:     Generate Key B, sign with Key B
            JWKS: [Key B, Key A]  ← Key A still validates old tokens

Day 60:     Generate Key C, sign with Key C
            JWKS: [Key C, Key B]  ← Key A removed (all tokens expired)

Rules:
1. Always keep at least 2 keys in JWKS (current + previous)
2. New key becomes active immediately for signing
3. Old keys remain for validation until all tokens signed with them expire
4. Remove old keys after: max_token_lifetime + clock_skew_tolerance
5. Rotate every 30-90 days (or immediately if compromised)
6. Never remove the current signing key from JWKS

Emergency rotation (key compromise):
1. Generate new key immediately
2. Remove compromised key from JWKS
3. All tokens signed with compromised key become invalid
4. Users must re-authenticate
```

### Key Rotation Implementation

```javascript
// auth/key-rotation.js — Automated key rotation
const jose = require('jose');

class KeyRotationManager {
  constructor(keyStore) {
    this.keyStore = keyStore; // Database or Redis
    this.rotationInterval = 30 * 24 * 60 * 60 * 1000; // 30 days
    this.maxKeyAge = 90 * 24 * 60 * 60 * 1000; // 90 days
    this.algorithm = 'RS256';
  }

  async initialize() {
    const keys = await this.keyStore.listKeys();
    if (keys.length === 0) {
      await this.generateKey();
    }
  }

  async generateKey() {
    const { publicKey, privateKey } = await jose.generateKeyPair(this.algorithm, {
      extractable: true,
    });

    const kid = `key-${Date.now()}`;

    // Export keys for storage
    const publicJWK = await jose.exportJWK(publicKey);
    const privateJWK = await jose.exportJWK(privateKey);

    await this.keyStore.saveKey({
      kid,
      algorithm: this.algorithm,
      publicKey: publicJWK,
      privateKey: privateJWK, // Encrypt before storing!
      createdAt: new Date(),
      isActive: true,
    });

    // Deactivate previous active key (but keep for validation)
    const previousKeys = await this.keyStore.getActiveKeys();
    for (const key of previousKeys) {
      if (key.kid !== kid) {
        await this.keyStore.deactivateKey(key.kid);
      }
    }

    return kid;
  }

  async getCurrentSigningKey() {
    const key = await this.keyStore.getActiveKey();
    if (!key) throw new Error('No active signing key');

    const privateKey = await jose.importJWK(key.privateKey, this.algorithm);
    return { privateKey, kid: key.kid };
  }

  getJWKS() {
    return async () => {
      const keys = await this.keyStore.listKeys();
      const jwks = [];

      for (const key of keys) {
        // Only include non-expired keys
        const age = Date.now() - key.createdAt.getTime();
        if (age < this.maxKeyAge) {
          jwks.push({
            ...key.publicKey,
            kid: key.kid,
            use: 'sig',
            alg: this.algorithm,
          });
        }
      }

      return { keys: jwks };
    };
  }

  async rotateIfNeeded() {
    const activeKey = await this.keyStore.getActiveKey();
    if (!activeKey) {
      return this.generateKey();
    }

    const age = Date.now() - activeKey.createdAt.getTime();
    if (age >= this.rotationInterval) {
      console.log(`Rotating signing key (current key age: ${Math.floor(age / 86400000)} days)`);
      return this.generateKey();
    }

    return null; // No rotation needed
  }

  async cleanupExpiredKeys() {
    const keys = await this.keyStore.listKeys();
    for (const key of keys) {
      const age = Date.now() - key.createdAt.getTime();
      if (age >= this.maxKeyAge && !key.isActive) {
        await this.keyStore.deleteKey(key.kid);
        console.log(`Removed expired key: ${key.kid}`);
      }
    }
  }
}

module.exports = { KeyRotationManager };
```

---

## Token Revocation

### Revocation Strategies

```
┌──────────────────────┬───────────────┬──────────────┬──────────────┐
│ Strategy             │ Latency       │ Scalability  │ Complexity   │
├──────────────────────┼───────────────┼──────────────┼──────────────┤
│ Short-lived tokens   │ Up to TTL     │ Excellent    │ Low          │
│ (no revocation)      │ (5-15 min)    │              │              │
├──────────────────────┼───────────────┼──────────────┼──────────────┤
│ Token blocklist      │ Immediate     │ Good         │ Medium       │
│ (Redis set)          │               │ (needs Redis)│              │
├──────────────────────┼───────────────┼──────────────┼──────────────┤
│ Token versioning     │ Immediate     │ Good         │ Medium       │
│ (user token version) │               │ (needs DB)   │              │
├──────────────────────┼───────────────┼──────────────┼──────────────┤
│ Introspection        │ Immediate     │ Poor         │ High         │
│ (check auth server)  │               │ (per-request)│              │
├──────────────────────┼───────────────┼──────────────┼──────────────┤
│ Event-driven         │ Near-instant  │ Excellent    │ High         │
│ (pub/sub revocation) │               │              │              │
└──────────────────────┴───────────────┴──────────────┴──────────────┘

Recommendation:
- Short-lived access tokens (5-15 min) + refresh token rotation
- Add Redis blocklist for immediate revocation when needed
- Use token versioning for "logout everywhere" functionality
```

### Token Blocklist Implementation

```javascript
// auth/token-blocklist.js — Redis-based token blocklist
class TokenBlocklist {
  constructor(redis) {
    this.redis = redis;
    this.prefix = 'blocked:';
  }

  // Block a specific token (by jti)
  async blockToken(jti, expiresAt) {
    const ttl = Math.ceil((expiresAt * 1000 - Date.now()) / 1000);
    if (ttl <= 0) return; // Token already expired, no need to block

    await this.redis.set(`${this.prefix}${jti}`, '1', 'EX', ttl);
  }

  // Check if a token is blocked
  async isBlocked(jti) {
    const result = await this.redis.get(`${this.prefix}${jti}`);
    return result !== null;
  }

  // Block all tokens for a user (by incrementing version)
  async blockAllUserTokens(userId) {
    await this.redis.incr(`${this.prefix}user_version:${userId}`);
  }

  // Get current token version for user
  async getUserTokenVersion(userId) {
    const version = await this.redis.get(`${this.prefix}user_version:${userId}`);
    return parseInt(version || '0');
  }
}

// Middleware that checks blocklist
function blocklistMiddleware(blocklist) {
  return async (req, res, next) => {
    if (!req.user?.tokenId) return next();

    const isBlocked = await blocklist.isBlocked(req.user.tokenId);
    if (isBlocked) {
      return res.status(401).json({ error: 'Token has been revoked' });
    }

    next();
  };
}

module.exports = { TokenBlocklist, blocklistMiddleware };
```

---

## Common JWT Vulnerabilities

### 1. Algorithm Confusion Attack

```
Attack: Attacker changes alg from RS256 to HS256 and signs
        with the public key (which is publicly available).

Vulnerable code:
  jwt.verify(token, publicKey)  // Accepts both RS256 and HS256!

The library uses publicKey as HMAC secret when alg=HS256.

Fix: ALWAYS specify allowed algorithms explicitly:
  jwt.verify(token, publicKey, { algorithms: ['RS256'] })
```

### 2. None Algorithm Attack

```
Attack: Set alg to "none" and remove signature.

Vulnerable token:
  eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.
  eyJzdWIiOiJhZG1pbiIsInJvbGUiOiJhZG1pbiJ9.
  (empty signature)

Fix: NEVER accept "none" algorithm:
  jwt.verify(token, key, { algorithms: ['RS256'] })
  // Do NOT include 'none' in the algorithms array
```

### 3. Key ID (kid) Injection

```
Attack: Manipulate the kid header to point to a different key,
        a file path, or a SQL injection payload.

Vulnerable: Using kid directly in a file path or SQL query
  const key = fs.readFileSync(`/keys/${header.kid}.pem`);
  // kid = "../../etc/passwd" → path traversal!

  const key = db.query(`SELECT key FROM keys WHERE kid = '${header.kid}'`);
  // kid = "' OR '1'='1" → SQL injection!

Fix:
1. Validate kid against a whitelist of known key IDs
2. Use parameterized queries if looking up in database
3. Never use kid in file paths
```

### 4. JWT Token Leakage

```
Common leakage vectors:
1. Tokens in URL query parameters → visible in server logs, Referer header
2. Tokens in localStorage → accessible via XSS
3. Tokens in error messages → visible to attackers
4. Tokens in browser history → persisted after logout

Fix:
1. Use Authorization header (not URL params)
2. Use httpOnly cookies or in-memory storage
3. Sanitize error responses (never include tokens)
4. Clear tokens on logout
```

### 5. Excessive Token Lifetime

```
Problem: Access tokens with 24h or 7d expiry are essentially
         session tokens without revocation capability.

If an access token with 7-day expiry is stolen, the attacker
has 7 days of unrestricted access with no way to revoke it.

Fix:
- Access tokens: 5-15 minutes maximum
- Use refresh tokens for extending sessions
- Implement refresh token rotation
- Add token blocklist for emergency revocation
```

---

## Token Lifetime Recommendations

```
┌─────────────────────┬─────────────────┬──────────────────────────┐
│ Token Type          │ Recommended TTL │ Notes                    │
├─────────────────────┼─────────────────┼──────────────────────────┤
│ Access Token        │ 5-15 minutes    │ Short-lived, frequently  │
│                     │                 │ refreshed                │
├─────────────────────┼─────────────────┼──────────────────────────┤
│ ID Token (OIDC)     │ 5-60 minutes    │ Only used at login to    │
│                     │                 │ establish identity       │
├─────────────────────┼─────────────────┼──────────────────────────┤
│ Refresh Token       │ 7-30 days       │ Rotate on each use;      │
│                     │                 │ revocable                │
├─────────────────────┼─────────────────┼──────────────────────────┤
│ Sliding Refresh     │ 15 min idle     │ Extends if user is       │
│ Token               │ 30 day absolute │ active; hard limit       │
├─────────────────────┼─────────────────┼──────────────────────────┤
│ MFA Challenge Token │ 5 minutes       │ Only valid for MFA step  │
├─────────────────────┼─────────────────┼──────────────────────────┤
│ Email Verification  │ 24 hours        │ One-time use             │
├─────────────────────┼─────────────────┼──────────────────────────┤
│ Password Reset      │ 1 hour          │ One-time use             │
├─────────────────────┼─────────────────┼──────────────────────────┤
│ API Key Token       │ 90-365 days     │ Revocable, scoped        │
│                     │ (or no expiry)  │                          │
├─────────────────────┼─────────────────┼──────────────────────────┤
│ Service-to-Service  │ 1 hour          │ Auto-renewed by client   │
│ (Client Credentials)│                 │ credentials flow         │
└─────────────────────┴─────────────────┴──────────────────────────┘
```

---

## JWT vs Opaque Tokens

```
┌──────────────────────┬─────────────────────┬──────────────────────┐
│ Feature              │ JWT (Self-contained)│ Opaque Token         │
├──────────────────────┼─────────────────────┼──────────────────────┤
│ Validation           │ Local (no DB/API)   │ Requires server call │
│ Size                 │ 300-800 bytes       │ 32-64 bytes          │
│ Contains user data   │ Yes (claims)        │ No (reference only)  │
│ Revocation           │ Hard (need blocklist)│ Easy (delete from DB)│
│ Scalability          │ Excellent           │ Depends on store     │
│ Stateless            │ Yes                 │ No                   │
│ Cross-service        │ Easy (verify sig)   │ Needs shared store   │
│ Information leakage  │ Risk (if not HTTPS) │ Minimal              │
│ Best for             │ Microservices, APIs │ Server-side sessions │
└──────────────────────┴─────────────────────┴──────────────────────┘

Hybrid approach (recommended for most apps):
- JWT access tokens: stateless verification, short-lived
- Opaque refresh tokens: stored in DB, revocable, rotated
```

---

## JWT in Different Frameworks

### Python (PyJWT)

```python
import jwt
from cryptography.hazmat.primitives.asymmetric import rsa

# Generate keys
private_key = rsa.generate_private_key(public_exponent=65537, key_size=2048)
public_key = private_key.public_key()

# Sign
token = jwt.encode(
    {"sub": "user-123", "exp": datetime.utcnow() + timedelta(minutes=15)},
    private_key,
    algorithm="RS256",
    headers={"kid": "key-1"},
)

# Verify (with algorithm whitelist!)
payload = jwt.decode(
    token,
    public_key,
    algorithms=["RS256"],  # ALWAYS whitelist algorithms
    issuer="https://auth.example.com",
    audience="https://api.example.com",
    options={"require": ["exp", "sub", "iss", "aud"]},
)
```

### Go (golang-jwt)

```go
import (
    "github.com/golang-jwt/jwt/v5"
)

// Sign
claims := jwt.MapClaims{
    "sub": "user-123",
    "exp": time.Now().Add(15 * time.Minute).Unix(),
    "iss": "https://auth.example.com",
}
token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
token.Header["kid"] = "key-1"
tokenString, _ := token.SignedString(privateKey)

// Verify (with algorithm whitelist!)
parsed, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
    if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
        return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
    }
    return publicKey, nil
}, jwt.WithIssuer("https://auth.example.com"),
   jwt.WithExpirationRequired())
```

---

## Security Checklist

```
✅ JWT Security Checklist

Signing:
□ Use RS256 or ES256 (asymmetric) for multi-service systems
□ Use HS256 ONLY for single-service with strong secret (32+ bytes)
□ NEVER use "none" algorithm
□ ALWAYS whitelist allowed algorithms in verification
□ Include "kid" header for key rotation support
□ Rotate signing keys every 30-90 days
□ Keep JWKS endpoint available with current + previous keys

Claims:
□ Always include: iss, sub, aud, exp, iat, jti
□ Validate ALL registered claims on every request
□ Set short expiry for access tokens (5-15 minutes)
□ Use "jti" for revocation tracking
□ Minimize custom claims (keep tokens small)
□ Never include sensitive data (passwords, PII, secrets)
□ Use "type" claim to distinguish access/refresh/id tokens

Storage:
□ Never store in localStorage (XSS-vulnerable)
□ Use httpOnly cookies for refresh tokens
□ Keep access tokens in memory only (SPA)
□ Use server-side sessions for server-rendered apps
□ Clear all tokens on logout

Validation:
□ Always verify signature before trusting any claims
□ Reject tokens with unknown algorithms
□ Validate issuer, audience, and expiration
□ Allow small clock skew tolerance (5 seconds)
□ Check token blocklist for revoked tokens
□ Validate kid against known keys only

Infrastructure:
□ Serve JWKS endpoint with proper caching headers
□ Monitor for signature verification failures (attack indicator)
□ Log token creation and validation events
□ Implement rate limiting on token endpoints
□ Use HTTPS everywhere — JWTs are not encrypted
```
