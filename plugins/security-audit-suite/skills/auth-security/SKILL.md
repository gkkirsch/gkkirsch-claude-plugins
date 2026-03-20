---
name: auth-security
description: >
  Authentication and authorization security — JWT best practices, password
  hashing, OAuth 2.0/OIDC, session management, RBAC, API key management,
  and secure token handling.
  Triggers: "jwt security", "password hashing", "oauth security",
  "session management", "rbac", "api key security", "auth best practices".
  NOT for: XSS, CSRF, or general web security (use web-security).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Authentication & Authorization Security

## Password Hashing

```typescript
import bcrypt from "bcrypt";
// Alternative: import argon2 from "argon2";

const SALT_ROUNDS = 12;  // ~250ms on modern hardware

// Hash password
async function hashPassword(plaintext: string): Promise<string> {
  return bcrypt.hash(plaintext, SALT_ROUNDS);
}

// Verify password (constant-time comparison built in)
async function verifyPassword(plaintext: string, hash: string): Promise<boolean> {
  return bcrypt.compare(plaintext, hash);
}

// With Argon2 (recommended for new projects)
// async function hashPassword(plaintext: string): Promise<string> {
//   return argon2.hash(plaintext, {
//     type: argon2.argon2id,
//     memoryCost: 65536,   // 64 MB
//     timeCost: 3,
//     parallelism: 4,
//   });
// }

// Password strength requirements
import { z } from "zod";

const PasswordSchema = z.string()
  .min(8, "At least 8 characters")
  .max(128, "Too long")
  .refine((pw) => /[A-Z]/.test(pw), "Need uppercase letter")
  .refine((pw) => /[a-z]/.test(pw), "Need lowercase letter")
  .refine((pw) => /[0-9]/.test(pw), "Need a digit")
  .refine((pw) => !/^(password|12345|qwerty)/i.test(pw), "Too common");
```

## JWT Best Practices

```typescript
import jwt from "jsonwebtoken";

const ACCESS_SECRET = process.env.JWT_ACCESS_SECRET!;
const REFRESH_SECRET = process.env.JWT_REFRESH_SECRET!;

// Short-lived access token (15 min)
function generateAccessToken(user: { id: string; role: string }): string {
  return jwt.sign(
    { sub: user.id, role: user.role },
    ACCESS_SECRET,
    {
      expiresIn: "15m",
      issuer: "myapp",
      audience: "myapp-api",
    }
  );
}

// Long-lived refresh token (7 days), stored in httpOnly cookie
function generateRefreshToken(user: { id: string }): string {
  return jwt.sign(
    { sub: user.id, type: "refresh", jti: crypto.randomUUID() },
    REFRESH_SECRET,
    { expiresIn: "7d" }
  );
}

// Verify access token
function verifyAccessToken(token: string): JwtPayload {
  return jwt.verify(token, ACCESS_SECRET, {
    issuer: "myapp",
    audience: "myapp-api",
  }) as JwtPayload;
}

// Auth middleware
async function requireAuth(req: Request, res: Response, next: NextFunction) {
  const authHeader = req.headers.authorization;
  if (!authHeader?.startsWith("Bearer ")) {
    return res.status(401).json({ error: "Missing token" });
  }

  try {
    const token = authHeader.slice(7);
    const payload = verifyAccessToken(token);

    // Check token isn't revoked
    const isRevoked = await redis.get(`revoked:${payload.jti}`);
    if (isRevoked) return res.status(401).json({ error: "Token revoked" });

    req.user = { id: payload.sub, role: payload.role };
    next();
  } catch (error) {
    if (error instanceof jwt.TokenExpiredError) {
      return res.status(401).json({ error: "Token expired" });
    }
    return res.status(401).json({ error: "Invalid token" });
  }
}

// Token refresh endpoint
app.post("/auth/refresh", async (req, res) => {
  const refreshToken = req.cookies.refreshToken;
  if (!refreshToken) return res.status(401).json({ error: "No refresh token" });

  try {
    const payload = jwt.verify(refreshToken, REFRESH_SECRET) as JwtPayload;

    // Rotate refresh token (invalidate old one)
    await redis.set(`revoked:${payload.jti}`, "1", "EX", 7 * 86400);

    const user = await db.user.findUnique({ where: { id: payload.sub } });
    if (!user) return res.status(401).json({ error: "User not found" });

    const newAccessToken = generateAccessToken(user);
    const newRefreshToken = generateRefreshToken(user);

    res.cookie("refreshToken", newRefreshToken, {
      httpOnly: true,
      secure: true,
      sameSite: "strict",
      maxAge: 7 * 86400 * 1000,
      path: "/auth/refresh",  // Only sent to refresh endpoint
    });

    res.json({ accessToken: newAccessToken });
  } catch {
    return res.status(401).json({ error: "Invalid refresh token" });
  }
});
```

### JWT Security Rules

| Do | Don't |
|----|-------|
| Store access token in memory (variable) | Store in localStorage (XSS accessible) |
| Store refresh token in httpOnly cookie | Store refresh token in localStorage |
| Use short expiry (15 min access) | Use long-lived access tokens (24h+) |
| Rotate refresh tokens on use | Reuse refresh tokens indefinitely |
| Validate `iss`, `aud`, `exp` claims | Trust token without verification |
| Use RS256 for distributed systems | Use HS256 with weak secrets |
| Revoke tokens on password change | Let old tokens live after password change |

## Session Management

```typescript
import session from "express-session";
import connectRedis from "connect-redis";
import Redis from "ioredis";

const redis = new Redis(process.env.REDIS_URL);

app.use(
  session({
    store: new connectRedis({ client: redis }),
    name: "__session",  // Don't use default "connect.sid"
    secret: process.env.SESSION_SECRET!,
    resave: false,
    saveUninitialized: false,
    cookie: {
      httpOnly: true,
      secure: process.env.NODE_ENV === "production",
      sameSite: "strict",
      maxAge: 24 * 60 * 60 * 1000,  // 24 hours
      domain: process.env.NODE_ENV === "production" ? ".myapp.com" : undefined,
    },
    rolling: true,  // Reset expiry on each request
  })
);

// Login
app.post("/auth/login", async (req, res) => {
  const { email, password } = req.body;
  const user = await db.user.findUnique({ where: { email } });

  if (!user || !(await verifyPassword(password, user.passwordHash))) {
    // Generic error — don't reveal if email exists
    return res.status(401).json({ error: "Invalid credentials" });
  }

  // Regenerate session ID to prevent fixation
  req.session.regenerate((err) => {
    if (err) return res.status(500).json({ error: "Session error" });

    req.session.userId = user.id;
    req.session.role = user.role;
    req.session.loginAt = Date.now();

    res.json({ user: { id: user.id, name: user.name, role: user.role } });
  });
});

// Logout — destroy session
app.post("/auth/logout", (req, res) => {
  req.session.destroy((err) => {
    res.clearCookie("__session");
    res.json({ ok: true });
  });
});
```

## Role-Based Access Control (RBAC)

```typescript
// Permission definitions
const PERMISSIONS = {
  "posts:read": ["admin", "editor", "viewer"],
  "posts:create": ["admin", "editor"],
  "posts:update": ["admin", "editor"],
  "posts:delete": ["admin"],
  "users:read": ["admin"],
  "users:manage": ["admin"],
  "settings:read": ["admin", "editor"],
  "settings:update": ["admin"],
} as const;

type Permission = keyof typeof PERMISSIONS;
type Role = "admin" | "editor" | "viewer";

function hasPermission(role: Role, permission: Permission): boolean {
  return PERMISSIONS[permission]?.includes(role) ?? false;
}

// Middleware
function requirePermission(permission: Permission) {
  return (req: Request, res: Response, next: NextFunction) => {
    if (!req.user) return res.status(401).json({ error: "Not authenticated" });

    if (!hasPermission(req.user.role as Role, permission)) {
      return res.status(403).json({ error: "Insufficient permissions" });
    }

    next();
  };
}

// Usage
app.get("/api/posts", requireAuth, requirePermission("posts:read"), getPostsHandler);
app.post("/api/posts", requireAuth, requirePermission("posts:create"), createPostHandler);
app.delete("/api/posts/:id", requireAuth, requirePermission("posts:delete"), deletePostHandler);

// Resource-level authorization
async function requireOwnership(req: Request, res: Response, next: NextFunction) {
  const post = await db.post.findUnique({ where: { id: req.params.id } });
  if (!post) return res.status(404).json({ error: "Not found" });

  // Admins can access any resource, others only their own
  if (req.user!.role !== "admin" && post.authorId !== req.user!.id) {
    return res.status(403).json({ error: "Not authorized" });
  }

  req.resource = post;
  next();
}
```

## API Key Management

```typescript
import crypto from "crypto";

// Generate API key (prefix for identification, random for security)
function generateApiKey(): { key: string; hash: string } {
  const prefix = "sk_live_";
  const random = crypto.randomBytes(32).toString("hex");
  const key = `${prefix}${random}`;
  const hash = crypto.createHash("sha256").update(key).digest("hex");
  return { key, hash };  // Store hash in DB, return key to user ONCE
}

// Verify API key
async function verifyApiKey(key: string): Promise<ApiKeyRecord | null> {
  const hash = crypto.createHash("sha256").update(key).digest("hex");
  return db.apiKey.findUnique({
    where: { hash, revokedAt: null },
    include: { user: true },
  });
}

// API key middleware
async function requireApiKey(req: Request, res: Response, next: NextFunction) {
  const key = req.headers["x-api-key"] as string;
  if (!key) return res.status(401).json({ error: "API key required" });

  const record = await verifyApiKey(key);
  if (!record) return res.status(401).json({ error: "Invalid API key" });

  // Check rate limit and scopes
  if (record.expiresAt && record.expiresAt < new Date()) {
    return res.status(401).json({ error: "API key expired" });
  }

  // Track usage
  await db.apiKey.update({
    where: { id: record.id },
    data: { lastUsedAt: new Date(), usageCount: { increment: 1 } },
  });

  req.user = { id: record.user.id, role: record.user.role };
  next();
}
```

## OAuth 2.0 / OIDC

```typescript
// OAuth 2.0 Authorization Code flow (server-side)
import { OAuth2Client } from "google-auth-library";

const oauth = new OAuth2Client({
  clientId: process.env.GOOGLE_CLIENT_ID,
  clientSecret: process.env.GOOGLE_CLIENT_SECRET,
  redirectUri: `${process.env.APP_URL}/auth/callback/google`,
});

// Step 1: Redirect to provider
app.get("/auth/google", (req, res) => {
  const state = crypto.randomBytes(32).toString("hex");
  req.session.oauthState = state;

  const url = oauth.generateAuthUrl({
    access_type: "offline",
    scope: ["openid", "email", "profile"],
    state,
    prompt: "consent",
  });

  res.redirect(url);
});

// Step 2: Handle callback
app.get("/auth/callback/google", async (req, res) => {
  // Verify state to prevent CSRF
  if (req.query.state !== req.session.oauthState) {
    return res.status(403).json({ error: "Invalid state" });
  }
  delete req.session.oauthState;

  const { tokens } = await oauth.getToken(req.query.code as string);
  const ticket = await oauth.verifyIdToken({
    idToken: tokens.id_token!,
    audience: process.env.GOOGLE_CLIENT_ID,
  });

  const payload = ticket.getPayload()!;
  const { sub: googleId, email, name, picture } = payload;

  // Find or create user
  let user = await db.user.findUnique({ where: { googleId } });
  if (!user) {
    user = await db.user.create({
      data: { googleId, email: email!, name: name!, avatar: picture },
    });
  }

  // Create session
  req.session.regenerate(() => {
    req.session.userId = user.id;
    res.redirect("/dashboard");
  });
});
```

## Security Checklist

| Category | Check |
|----------|-------|
| **Passwords** | Bcrypt/Argon2 with cost factor ≥ 12 |
| **Passwords** | Minimum 8 characters, check against breach lists |
| **Tokens** | Access tokens ≤ 15 min, refresh tokens ≤ 7 days |
| **Tokens** | Refresh token rotation on every use |
| **Cookies** | httpOnly, secure, sameSite=strict |
| **Sessions** | Regenerate ID on login (prevent fixation) |
| **Sessions** | Destroy on logout, not just clear |
| **API Keys** | Store hashed, never log plaintext |
| **API Keys** | Revocable, with expiration dates |
| **OAuth** | Verify state parameter (CSRF protection) |
| **OAuth** | Validate ID token audience claim |
| **Errors** | Generic auth errors ("Invalid credentials") |
| **Errors** | Don't reveal if email exists on login failure |
| **Logging** | Log auth events, never log passwords/tokens |

## Gotchas

1. **Timing attacks on password comparison** — `password === storedHash` is vulnerable to timing analysis. Always use `bcrypt.compare()` which is constant-time. Never compare password hashes with `===`.

2. **JWT in localStorage is XSS-vulnerable** — Any XSS exploit can read localStorage and steal the token. Store access tokens in JavaScript memory (variable/closure) and refresh tokens in httpOnly cookies.

3. **OAuth state parameter is critical** — Without it, an attacker can craft a callback URL that links their OAuth account to the victim's session. Always generate, store, and verify the state parameter.

4. **Password reset tokens must be single-use** — After a password reset token is used, invalidate it AND all existing sessions for that user. Otherwise, an attacker with a stolen session continues to have access.

5. **Don't use `jwt.decode()` for verification** — `jwt.decode()` parses the token WITHOUT verifying the signature. Anyone can create a valid-looking token. Always use `jwt.verify()` with the secret.

6. **Session fixation after login** — Call `req.session.regenerate()` after successful authentication. Without it, an attacker who set a session cookie before login keeps access to the authenticated session.
