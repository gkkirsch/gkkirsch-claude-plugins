---
name: authentication
description: >
  Node.js authentication patterns — JWT access/refresh tokens, bcrypt password hashing,
  session-based auth, OAuth 2.0/Google sign-in, middleware guards, and role-based access control.
  Triggers: "jwt auth", "node authentication", "login endpoint", "password hashing",
  "refresh token", "oauth node", "session auth", "rbac", "protect routes".
  NOT for: frontend auth UI (use react-hook-form), API architecture (use api-architect agent).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Authentication Patterns

## JWT Access + Refresh Token Flow

```typescript
// src/lib/tokens.ts
import jwt from 'jsonwebtoken';
import { env } from '../config/env';

interface TokenPayload {
  userId: string;
  email: string;
  role: 'user' | 'admin';
}

export function generateAccessToken(payload: TokenPayload): string {
  return jwt.sign(payload, env.JWT_SECRET, { expiresIn: '15m' });
}

export function generateRefreshToken(payload: TokenPayload): string {
  return jwt.sign(payload, env.JWT_REFRESH_SECRET, { expiresIn: '7d' });
}

export function verifyAccessToken(token: string): TokenPayload {
  return jwt.verify(token, env.JWT_SECRET) as TokenPayload;
}

export function verifyRefreshToken(token: string): TokenPayload {
  return jwt.verify(token, env.JWT_REFRESH_SECRET) as TokenPayload;
}

export function generateTokenPair(payload: TokenPayload) {
  return {
    accessToken: generateAccessToken(payload),
    refreshToken: generateRefreshToken(payload),
    expiresIn: 900, // 15 minutes in seconds
  };
}
```

## Password Hashing

```typescript
// src/lib/password.ts
import bcrypt from 'bcrypt';

const SALT_ROUNDS = 12;

export async function hashPassword(password: string): Promise<string> {
  return bcrypt.hash(password, SALT_ROUNDS);
}

export async function verifyPassword(
  password: string,
  hash: string,
): Promise<boolean> {
  return bcrypt.compare(password, hash);
}
```

## Auth Controller

```typescript
// src/controllers/auth.controller.ts
import { Request, Response, NextFunction } from 'express';
import { UserService } from '../services/user.service';
import { hashPassword, verifyPassword } from '../lib/password';
import { generateTokenPair, verifyRefreshToken } from '../lib/tokens';
import { UnauthorizedError, ConflictError } from '../lib/errors';
import { prisma } from '../lib/prisma';

const userService = new UserService();

export async function signup(req: Request, res: Response, next: NextFunction) {
  try {
    const { email, password, name } = req.body;

    const existing = await userService.findByEmail(email);
    if (existing) throw new ConflictError('Email already registered');

    const hashedPassword = await hashPassword(password);
    const user = await userService.create({ email, password: hashedPassword, name });

    const tokens = generateTokenPair({
      userId: user.id,
      email: user.email,
      role: user.role,
    });

    // Store refresh token in database
    await prisma.refreshToken.create({
      data: {
        token: tokens.refreshToken,
        userId: user.id,
        expiresAt: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000),
      },
    });

    res.status(201).json({
      success: true,
      data: {
        user: { id: user.id, email: user.email, name: user.name },
        ...tokens,
      },
    });
  } catch (err) {
    next(err);
  }
}

export async function login(req: Request, res: Response, next: NextFunction) {
  try {
    const { email, password } = req.body;

    const user = await userService.findByEmail(email);
    if (!user) throw new UnauthorizedError('Invalid credentials');

    const valid = await verifyPassword(password, user.password);
    if (!valid) throw new UnauthorizedError('Invalid credentials');

    const tokens = generateTokenPair({
      userId: user.id,
      email: user.email,
      role: user.role,
    });

    // Store refresh token
    await prisma.refreshToken.create({
      data: {
        token: tokens.refreshToken,
        userId: user.id,
        expiresAt: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000),
      },
    });

    res.json({ success: true, data: tokens });
  } catch (err) {
    next(err);
  }
}

export async function refresh(req: Request, res: Response, next: NextFunction) {
  try {
    const { refreshToken } = req.body;
    if (!refreshToken) throw new UnauthorizedError('Refresh token required');

    // Verify token signature
    const payload = verifyRefreshToken(refreshToken);

    // Check token exists in database (not revoked)
    const stored = await prisma.refreshToken.findUnique({
      where: { token: refreshToken },
    });
    if (!stored || stored.expiresAt < new Date()) {
      throw new UnauthorizedError('Invalid refresh token');
    }

    // Rotate: delete old, create new
    await prisma.refreshToken.delete({ where: { token: refreshToken } });

    const tokens = generateTokenPair({
      userId: payload.userId,
      email: payload.email,
      role: payload.role,
    });

    await prisma.refreshToken.create({
      data: {
        token: tokens.refreshToken,
        userId: payload.userId,
        expiresAt: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000),
      },
    });

    res.json({ success: true, data: tokens });
  } catch (err) {
    next(err);
  }
}

export async function logout(req: Request, res: Response, next: NextFunction) {
  try {
    const { refreshToken } = req.body;
    if (refreshToken) {
      await prisma.refreshToken.deleteMany({ where: { token: refreshToken } });
    }
    res.json({ success: true, data: { message: 'Logged out' } });
  } catch (err) {
    next(err);
  }
}
```

## Auth Middleware

```typescript
// src/middleware/auth.ts
import { Request, Response, NextFunction } from 'express';
import { verifyAccessToken } from '../lib/tokens';
import { UnauthorizedError, ForbiddenError } from '../lib/errors';

export function authenticate(req: Request, _res: Response, next: NextFunction) {
  try {
    const header = req.headers.authorization;
    if (!header?.startsWith('Bearer ')) {
      throw new UnauthorizedError('Missing authorization header');
    }

    const token = header.slice(7);
    const payload = verifyAccessToken(token);

    req.user = {
      id: payload.userId,
      email: payload.email,
      role: payload.role,
    };

    next();
  } catch (err) {
    if (err instanceof UnauthorizedError) return next(err);
    next(new UnauthorizedError('Invalid or expired token'));
  }
}

// Role-based access control
export function authorize(...roles: Array<'user' | 'admin'>) {
  return (req: Request, _res: Response, next: NextFunction) => {
    if (!req.user) {
      return next(new UnauthorizedError('Not authenticated'));
    }
    if (!roles.includes(req.user.role)) {
      return next(new ForbiddenError('Insufficient permissions'));
    }
    next();
  };
}

// Resource ownership check
export function ownerOnly(getResourceUserId: (req: Request) => Promise<string | null>) {
  return async (req: Request, _res: Response, next: NextFunction) => {
    try {
      if (!req.user) return next(new UnauthorizedError('Not authenticated'));
      if (req.user.role === 'admin') return next(); // Admins bypass

      const resourceUserId = await getResourceUserId(req);
      if (!resourceUserId || resourceUserId !== req.user.id) {
        return next(new ForbiddenError('You do not own this resource'));
      }
      next();
    } catch (err) {
      next(err);
    }
  };
}
```

## Auth Routes

```typescript
// src/routes/auth.routes.ts
import { Router } from 'express';
import { signup, login, refresh, logout } from '../controllers/auth.controller';
import { validate } from '../middleware/validate';
import { authLimiter } from '../middleware/rate-limit';
import { z } from 'zod';

const router = Router();

const signupSchema = z.object({
  email: z.string().email(),
  password: z.string().min(8).max(128),
  name: z.string().min(1).max(100),
});

const loginSchema = z.object({
  email: z.string().email(),
  password: z.string().min(1),
});

const refreshSchema = z.object({
  refreshToken: z.string().min(1),
});

router.post('/signup', authLimiter, validate(signupSchema), signup);
router.post('/login', authLimiter, validate(loginSchema), login);
router.post('/refresh', validate(refreshSchema), refresh);
router.post('/logout', logout);

export { router as authRoutes };
```

## Session-Based Auth (Alternative)

```typescript
// src/middleware/session-auth.ts
import session from 'express-session';
import connectPgSimple from 'connect-pg-simple';
import { pool } from '../lib/db';

const PgStore = connectPgSimple(session);

export const sessionMiddleware = session({
  store: new PgStore({ pool, tableName: 'sessions' }),
  secret: process.env.SESSION_SECRET!,
  resave: false,
  saveUninitialized: false,
  cookie: {
    secure: process.env.NODE_ENV === 'production',
    httpOnly: true,
    sameSite: 'lax',
    maxAge: 24 * 60 * 60 * 1000, // 24 hours
  },
});

// Session login
export async function sessionLogin(req: Request, res: Response) {
  const user = await authenticateUser(req.body.email, req.body.password);
  req.session.userId = user.id;
  req.session.role = user.role;
  res.json({ success: true });
}

// Session auth check
export function requireSession(req: Request, _res: Response, next: NextFunction) {
  if (!req.session.userId) {
    return next(new UnauthorizedError('Not authenticated'));
  }
  req.user = { id: req.session.userId, role: req.session.role };
  next();
}
```

## Google OAuth 2.0

```typescript
// src/lib/google-oauth.ts
import { OAuth2Client } from 'google-auth-library';

const client = new OAuth2Client(
  process.env.GOOGLE_CLIENT_ID,
  process.env.GOOGLE_CLIENT_SECRET,
  process.env.GOOGLE_REDIRECT_URI,
);

export function getGoogleAuthUrl(): string {
  return client.generateAuthUrl({
    access_type: 'offline',
    scope: ['openid', 'email', 'profile'],
    prompt: 'consent',
  });
}

export async function getGoogleUser(code: string) {
  const { tokens } = await client.getToken(code);
  client.setCredentials(tokens);

  const ticket = await client.verifyIdToken({
    idToken: tokens.id_token!,
    audience: process.env.GOOGLE_CLIENT_ID,
  });

  const payload = ticket.getPayload()!;
  return {
    googleId: payload.sub,
    email: payload.email!,
    name: payload.name || '',
    picture: payload.picture || '',
  };
}

// Route handler
export async function googleCallback(req: Request, res: Response, next: NextFunction) {
  try {
    const { code } = req.query;
    const googleUser = await getGoogleUser(code as string);

    // Find or create user
    let user = await prisma.user.findUnique({
      where: { googleId: googleUser.googleId },
    });

    if (!user) {
      user = await prisma.user.create({
        data: {
          email: googleUser.email,
          name: googleUser.name,
          googleId: googleUser.googleId,
          avatar: googleUser.picture,
        },
      });
    }

    const tokens = generateTokenPair({
      userId: user.id,
      email: user.email,
      role: user.role,
    });

    // Redirect to frontend with tokens
    const params = new URLSearchParams({
      accessToken: tokens.accessToken,
      refreshToken: tokens.refreshToken,
    });
    res.redirect(`${process.env.FRONTEND_URL}/auth/callback?${params}`);
  } catch (err) {
    next(err);
  }
}
```

## API Key Authentication

```typescript
// src/middleware/api-key.ts
import { Request, Response, NextFunction } from 'express';
import { prisma } from '../lib/prisma';
import { UnauthorizedError } from '../lib/errors';
import crypto from 'crypto';

export async function authenticateApiKey(
  req: Request,
  _res: Response,
  next: NextFunction,
) {
  try {
    const apiKey = req.headers['x-api-key'] as string;
    if (!apiKey) throw new UnauthorizedError('API key required');

    // Hash the key to compare with stored hash
    const keyHash = crypto.createHash('sha256').update(apiKey).digest('hex');

    const key = await prisma.apiKey.findUnique({
      where: { keyHash },
      include: { user: true },
    });

    if (!key || key.revokedAt) {
      throw new UnauthorizedError('Invalid API key');
    }

    // Update last used timestamp
    await prisma.apiKey.update({
      where: { id: key.id },
      data: { lastUsedAt: new Date() },
    });

    req.user = {
      id: key.user.id,
      email: key.user.email,
      role: key.scope === 'admin' ? 'admin' : 'user',
    };

    next();
  } catch (err) {
    next(err);
  }
}

// Generate a new API key (return raw key once, store hash)
export async function createApiKey(userId: string, name: string, scope: string) {
  const rawKey = `sk_${crypto.randomBytes(32).toString('hex')}`;
  const keyHash = crypto.createHash('sha256').update(rawKey).digest('hex');

  await prisma.apiKey.create({
    data: { keyHash, name, scope, userId },
  });

  return rawKey; // Only returned once — user must save it
}
```

## Gotchas

1. **Never store raw refresh tokens.** Hash them or store server-side in DB. If an attacker dumps your database, unhashed refresh tokens give them persistent access.

2. **JWT_SECRET must be strong.** Minimum 256-bit (32+ characters). Use `node -e "console.log(require('crypto').randomBytes(64).toString('hex'))"` to generate.

3. **`bcrypt` cost factor matters.** 12 is the current recommendation. 10 is too fast (brute-forceable), 14+ is slow for users. Re-evaluate yearly as hardware improves.

4. **Refresh token rotation is critical.** When a refresh token is used, delete it and issue a new one. If a stolen refresh token is used after the real user rotated it, both tokens are invalidated (detect reuse attack).

5. **Don't put sensitive data in JWT payload.** JWTs are base64-encoded, NOT encrypted. Anyone can decode them. Never include passwords, SSN, or other secrets. Keep payload minimal: userId, email, role.

6. **Session fixation with express-session.** Call `req.session.regenerate()` after login to prevent session fixation attacks. The session ID changes but session data persists.

7. **OAuth state parameter.** Always include a `state` parameter in OAuth flows to prevent CSRF. Generate a random string, store in session, verify on callback.
