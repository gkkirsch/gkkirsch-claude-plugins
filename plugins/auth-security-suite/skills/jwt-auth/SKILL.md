---
name: jwt-auth
description: >
  JWT authentication implementation — token creation, verification, refresh token
  rotation, middleware patterns, and cookie-based JWT storage.
  Triggers: "JWT", "JSON web token", "access token", "refresh token", "token auth",
  "bearer token", "jwt middleware", "token rotation".
  NOT for: OAuth flows (use oauth-patterns), session-based auth, security hardening (use security-hardening).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# JWT Authentication

## Token Creation & Verification

```typescript
import jwt from 'jsonwebtoken';

const JWT_SECRET = process.env.JWT_SECRET!; // Min 256 bits (32+ chars)
const JWT_REFRESH_SECRET = process.env.JWT_REFRESH_SECRET!;

// Create access token (short-lived)
function createAccessToken(user: { id: string; role: string }): string {
  return jwt.sign(
    { sub: user.id, role: user.role },
    JWT_SECRET,
    { expiresIn: '15m', algorithm: 'HS256' }
  );
}

// Create refresh token (long-lived)
function createRefreshToken(user: { id: string }): string {
  return jwt.sign(
    { sub: user.id, type: 'refresh' },
    JWT_REFRESH_SECRET,
    { expiresIn: '7d', algorithm: 'HS256' }
  );
}

// Verify token
function verifyAccessToken(token: string): JwtPayload {
  return jwt.verify(token, JWT_SECRET, {
    algorithms: ['HS256'], // Prevent algorithm confusion attack
  }) as JwtPayload;
}

interface JwtPayload {
  sub: string;
  role: string;
  iat: number;
  exp: number;
}
```

## Cookie-Based JWT (Recommended for Web)

```typescript
import { Request, Response } from 'express';

// Set tokens as HttpOnly cookies
function setAuthCookies(res: Response, accessToken: string, refreshToken: string) {
  res.cookie('access_token', accessToken, {
    httpOnly: true,        // Not accessible via JavaScript (XSS protection)
    secure: true,          // HTTPS only
    sameSite: 'lax',       // CSRF protection (allows top-level navigations)
    maxAge: 15 * 60 * 1000, // 15 minutes
    path: '/',
  });

  res.cookie('refresh_token', refreshToken, {
    httpOnly: true,
    secure: true,
    sameSite: 'lax',
    maxAge: 7 * 24 * 60 * 60 * 1000, // 7 days
    path: '/api/auth/refresh', // Only sent to refresh endpoint
  });
}

// Clear cookies on logout
function clearAuthCookies(res: Response) {
  res.clearCookie('access_token');
  res.clearCookie('refresh_token', { path: '/api/auth/refresh' });
}
```

## Auth Middleware

```typescript
import { Request, Response, NextFunction } from 'express';

// Extend Express Request type
declare global {
  namespace Express {
    interface Request {
      user?: JwtPayload;
    }
  }
}

// Auth middleware — extracts and verifies JWT
function authenticate(req: Request, res: Response, next: NextFunction) {
  // Try cookie first, then Authorization header
  const token = req.cookies.access_token
    || req.headers.authorization?.replace('Bearer ', '');

  if (!token) {
    return res.status(401).json({ error: 'Authentication required' });
  }

  try {
    req.user = verifyAccessToken(token);
    next();
  } catch (err) {
    if (err instanceof jwt.TokenExpiredError) {
      return res.status(401).json({ error: 'Token expired', code: 'TOKEN_EXPIRED' });
    }
    return res.status(401).json({ error: 'Invalid token' });
  }
}

// Role-based authorization
function authorize(...roles: string[]) {
  return (req: Request, res: Response, next: NextFunction) => {
    if (!req.user) {
      return res.status(401).json({ error: 'Authentication required' });
    }
    if (!roles.includes(req.user.role)) {
      return res.status(403).json({ error: 'Insufficient permissions' });
    }
    next();
  };
}

// Usage
app.get('/api/admin/users', authenticate, authorize('admin'), getUsers);
app.get('/api/profile', authenticate, getProfile);
```

## Refresh Token Rotation

```typescript
import { randomBytes } from 'crypto';

// Store refresh tokens server-side (Redis or DB)
// Key: refreshTokenId → { userId, family, isRevoked }
interface StoredRefreshToken {
  userId: string;
  family: string;     // Token family for detecting reuse
  isRevoked: boolean;
  expiresAt: Date;
}

async function refreshTokens(req: Request, res: Response) {
  const oldRefreshToken = req.cookies.refresh_token;
  if (!oldRefreshToken) {
    return res.status(401).json({ error: 'No refresh token' });
  }

  try {
    const payload = jwt.verify(oldRefreshToken, JWT_REFRESH_SECRET) as JwtPayload;
    const stored = await db.refreshTokens.findByToken(oldRefreshToken);

    if (!stored || stored.isRevoked) {
      // Token reuse detected — revoke entire family
      if (stored) {
        await db.refreshTokens.revokeFamily(stored.family);
      }
      clearAuthCookies(res);
      return res.status(401).json({ error: 'Token reuse detected' });
    }

    // Revoke old token
    await db.refreshTokens.revoke(oldRefreshToken);

    // Issue new token pair
    const user = await db.users.findById(payload.sub);
    const newAccessToken = createAccessToken(user);
    const newRefreshToken = createRefreshToken(user);

    // Store new refresh token in same family
    await db.refreshTokens.create({
      token: newRefreshToken,
      userId: user.id,
      family: stored.family,
      expiresAt: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000),
    });

    setAuthCookies(res, newAccessToken, newRefreshToken);
    return res.json({ user: sanitizeUser(user) });
  } catch {
    clearAuthCookies(res);
    return res.status(401).json({ error: 'Invalid refresh token' });
  }
}
```

## Login & Registration Flow

```typescript
import bcrypt from 'bcrypt';
import { z } from 'zod';

const loginSchema = z.object({
  email: z.string().email(),
  password: z.string().min(8),
});

async function login(req: Request, res: Response) {
  const { email, password } = loginSchema.parse(req.body);

  const user = await db.users.findByEmail(email);
  if (!user || !(await bcrypt.compare(password, user.passwordHash))) {
    // Generic error (don't reveal which field is wrong)
    return res.status(401).json({ error: 'Invalid email or password' });
  }

  const accessToken = createAccessToken(user);
  const refreshToken = createRefreshToken(user);

  // Store refresh token
  const family = randomBytes(16).toString('hex');
  await db.refreshTokens.create({
    token: refreshToken,
    userId: user.id,
    family,
    expiresAt: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000),
  });

  setAuthCookies(res, accessToken, refreshToken);
  return res.json({ user: sanitizeUser(user) });
}

async function register(req: Request, res: Response) {
  const { email, password } = loginSchema.parse(req.body);

  const existing = await db.users.findByEmail(email);
  if (existing) {
    return res.status(409).json({ error: 'Email already registered' });
  }

  const passwordHash = await bcrypt.hash(password, 12); // Cost factor 12
  const user = await db.users.create({ email, passwordHash, role: 'user' });

  const accessToken = createAccessToken(user);
  const refreshToken = createRefreshToken(user);

  const family = randomBytes(16).toString('hex');
  await db.refreshTokens.create({
    token: refreshToken,
    userId: user.id,
    family,
    expiresAt: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000),
  });

  setAuthCookies(res, accessToken, refreshToken);
  return res.status(201).json({ user: sanitizeUser(user) });
}

async function logout(req: Request, res: Response) {
  const refreshToken = req.cookies.refresh_token;
  if (refreshToken) {
    const stored = await db.refreshTokens.findByToken(refreshToken);
    if (stored) {
      await db.refreshTokens.revokeFamily(stored.family);
    }
  }
  clearAuthCookies(res);
  return res.json({ success: true });
}

// Never expose password hash
function sanitizeUser(user: User) {
  const { passwordHash, ...safe } = user;
  return safe;
}
```

## Client-Side Token Handling (React)

```typescript
// Auth context with automatic token refresh
import { createContext, useContext, useCallback, useEffect, useState } from 'react';

interface AuthContextType {
  user: User | null;
  login: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  isLoading: boolean;
}

const AuthContext = createContext<AuthContextType | null>(null);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  // Check auth status on mount (cookie-based, no token in JS)
  useEffect(() => {
    fetch('/api/auth/me', { credentials: 'include' })
      .then(r => r.ok ? r.json() : null)
      .then(data => setUser(data?.user ?? null))
      .finally(() => setIsLoading(false));
  }, []);

  const login = useCallback(async (email: string, password: string) => {
    const res = await fetch('/api/auth/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include', // Send/receive cookies
      body: JSON.stringify({ email, password }),
    });
    if (!res.ok) throw new Error('Login failed');
    const data = await res.json();
    setUser(data.user);
  }, []);

  const logout = useCallback(async () => {
    await fetch('/api/auth/logout', {
      method: 'POST',
      credentials: 'include',
    });
    setUser(null);
  }, []);

  return (
    <AuthContext.Provider value={{ user, login, logout, isLoading }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used within AuthProvider');
  return ctx;
}

// Fetch wrapper that handles 401 → refresh → retry
async function authFetch(url: string, options: RequestInit = {}): Promise<Response> {
  let res = await fetch(url, { ...options, credentials: 'include' });

  if (res.status === 401) {
    // Try refreshing the token
    const refreshRes = await fetch('/api/auth/refresh', {
      method: 'POST',
      credentials: 'include',
    });

    if (refreshRes.ok) {
      // Retry original request with new token
      res = await fetch(url, { ...options, credentials: 'include' });
    }
  }

  return res;
}
```

## Gotchas

1. **Never store JWTs in localStorage.** Any XSS vulnerability exposes the token. Use HttpOnly cookies — JavaScript can't read them, so XSS can't steal them.

2. **JWT_SECRET must be at least 256 bits (32 characters).** Shorter secrets are brute-forceable. Use `openssl rand -hex 32` to generate.

3. **Always set `algorithms: ['HS256']` when verifying.** Without this, an attacker can set the algorithm to `none` and bypass verification entirely (algorithm confusion attack).

4. **Refresh token rotation is not optional.** Without it, a stolen refresh token gives permanent access. Always rotate: old token → revoked, new token → issued. Track token families to detect reuse.

5. **"Invalid email or password" — never reveal which.** Saying "email not found" vs "wrong password" tells attackers which emails are registered (user enumeration).

6. **bcrypt cost factor matters.** Cost 10 is the minimum for production (2024). Each increment doubles the time. 12 is recommended. Never use MD5, SHA-256, or any non-adaptive hash for passwords.
