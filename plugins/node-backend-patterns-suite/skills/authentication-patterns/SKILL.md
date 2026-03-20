---
name: authentication-patterns
description: >
  Authentication patterns for Node.js APIs — JWT with refresh tokens, session-based auth,
  bcrypt password hashing, OAuth2, API keys, and middleware patterns.
  Triggers: "jwt auth", "authentication", "login api", "refresh token",
  "bcrypt", "session auth", "oauth", "api key auth", "passport".
  NOT for: authorization/RBAC (use security-engineer agent), frontend auth state (use React patterns).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Authentication Patterns

## JWT Authentication (Stateless)

### Auth Service

```typescript
// src/services/auth.service.ts
import bcrypt from 'bcrypt';
import jwt from 'jsonwebtoken';
import crypto from 'crypto';
import { prisma } from '../lib/prisma';
import { env } from '../config/env';
import { UnauthorizedError, ConflictError } from '../lib/errors';

const SALT_ROUNDS = 12;
const ACCESS_TOKEN_EXPIRY = '15m';
const REFRESH_TOKEN_EXPIRY = '7d';
const REFRESH_TOKEN_BYTES = 40;

interface TokenPayload {
  userId: string;
  email: string;
  role: string;
}

export class AuthService {
  async signup(email: string, password: string, name: string) {
    const existing = await prisma.user.findUnique({ where: { email } });
    if (existing) throw new ConflictError('Email already registered');

    const hashedPassword = await bcrypt.hash(password, SALT_ROUNDS);

    const user = await prisma.user.create({
      data: { email, password: hashedPassword, name },
      select: { id: true, email: true, name: true, role: true },
    });

    const tokens = await this.generateTokens(user);
    return { user, ...tokens };
  }

  async login(email: string, password: string) {
    const user = await prisma.user.findUnique({ where: { email } });
    if (!user) throw new UnauthorizedError('Invalid email or password');

    const valid = await bcrypt.compare(password, user.password);
    if (!valid) throw new UnauthorizedError('Invalid email or password');

    const tokens = await this.generateTokens(user);
    return {
      user: { id: user.id, email: user.email, name: user.name, role: user.role },
      ...tokens,
    };
  }

  async refresh(refreshToken: string) {
    const hashedToken = this.hashToken(refreshToken);

    const stored = await prisma.refreshToken.findUnique({
      where: { token: hashedToken },
      include: { user: true },
    });

    if (!stored || stored.expiresAt < new Date()) {
      if (stored) {
        // Token reuse detected — revoke all tokens for this user
        await prisma.refreshToken.deleteMany({ where: { userId: stored.userId } });
      }
      throw new UnauthorizedError('Invalid or expired refresh token');
    }

    // Rotate: delete old, create new
    await prisma.refreshToken.delete({ where: { id: stored.id } });

    const tokens = await this.generateTokens(stored.user);
    return tokens;
  }

  async logout(refreshToken: string) {
    const hashedToken = this.hashToken(refreshToken);
    await prisma.refreshToken.deleteMany({ where: { token: hashedToken } });
  }

  async logoutAll(userId: string) {
    await prisma.refreshToken.deleteMany({ where: { userId } });
  }

  private async generateTokens(user: { id: string; email: string; role: string }) {
    const payload: TokenPayload = {
      userId: user.id,
      email: user.email,
      role: user.role,
    };

    const accessToken = jwt.sign(payload, env.JWT_SECRET, {
      expiresIn: ACCESS_TOKEN_EXPIRY,
    });

    const refreshToken = crypto.randomBytes(REFRESH_TOKEN_BYTES).toString('hex');
    const hashedRefreshToken = this.hashToken(refreshToken);

    await prisma.refreshToken.create({
      data: {
        token: hashedRefreshToken,
        userId: user.id,
        expiresAt: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000),
      },
    });

    return { accessToken, refreshToken };
  }

  private hashToken(token: string): string {
    return crypto.createHash('sha256').update(token).digest('hex');
  }
}
```

### Auth Middleware

```typescript
// src/middleware/auth.ts
import { Request, Response, NextFunction } from 'express';
import jwt from 'jsonwebtoken';
import { env } from '../config/env';
import { UnauthorizedError, ForbiddenError } from '../lib/errors';

interface JwtPayload {
  userId: string;
  email: string;
  role: string;
}

export function authenticate(req: Request, _res: Response, next: NextFunction) {
  const header = req.headers.authorization;
  if (!header?.startsWith('Bearer ')) {
    return next(new UnauthorizedError('Missing or invalid authorization header'));
  }

  const token = header.slice(7);

  try {
    const payload = jwt.verify(token, env.JWT_SECRET) as JwtPayload;
    req.user = {
      id: payload.userId,
      email: payload.email,
      role: payload.role as 'user' | 'admin',
    };
    next();
  } catch (err) {
    if (err instanceof jwt.TokenExpiredError) {
      next(new UnauthorizedError('Token expired'));
    } else {
      next(new UnauthorizedError('Invalid token'));
    }
  }
}

// Optional auth — sets req.user if token present, continues if not
export function optionalAuth(req: Request, _res: Response, next: NextFunction) {
  const header = req.headers.authorization;
  if (!header?.startsWith('Bearer ')) return next();

  try {
    const token = header.slice(7);
    const payload = jwt.verify(token, env.JWT_SECRET) as JwtPayload;
    req.user = {
      id: payload.userId,
      email: payload.email,
      role: payload.role as 'user' | 'admin',
    };
  } catch {
    // Invalid token in optional auth — just continue without user
  }
  next();
}

// Role-based access control
export function requireRole(...roles: string[]) {
  return (req: Request, _res: Response, next: NextFunction) => {
    if (!req.user) return next(new UnauthorizedError('Not authenticated'));
    if (!roles.includes(req.user.role)) {
      return next(new ForbiddenError('Insufficient permissions'));
    }
    next();
  };
}
```

### Auth Routes

```typescript
// src/routes/auth.routes.ts
import { Router } from 'express';
import { AuthController } from '../controllers/auth.controller';
import { validate } from '../middleware/validate';
import { authenticate } from '../middleware/auth';
import { authLimiter } from '../middleware/rate-limit';
import { z } from 'zod';

const router = Router();
const controller = new AuthController();

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

router.post('/signup', authLimiter, validate(signupSchema), controller.signup);
router.post('/login', authLimiter, validate(loginSchema), controller.login);
router.post('/refresh', validate(refreshSchema), controller.refresh);
router.post('/logout', authenticate, controller.logout);
router.post('/logout-all', authenticate, controller.logoutAll);
router.get('/me', authenticate, controller.me);

export { router as authRoutes };
```

## Session-Based Authentication

```typescript
// src/config/session.ts
import session from 'express-session';
import connectPgSimple from 'connect-pg-simple';
import { Pool } from 'pg';

const PgSession = connectPgSimple(session);
const pool = new Pool({ connectionString: process.env.DATABASE_URL });

export const sessionConfig = session({
  store: new PgSession({
    pool,
    tableName: 'sessions',
    createTableIfMissing: true,
  }),
  secret: process.env.SESSION_SECRET!,
  resave: false,
  saveUninitialized: false,
  cookie: {
    secure: process.env.NODE_ENV === 'production',
    httpOnly: true,           // Cannot be read by JavaScript
    sameSite: 'lax',          // CSRF protection
    maxAge: 24 * 60 * 60 * 1000, // 24 hours
  },
});

// In app.ts:
// app.use(sessionConfig);

// Login:
// req.session.userId = user.id;
// req.session.role = user.role;

// Logout:
// req.session.destroy((err) => { ... });

// Auth middleware:
// if (!req.session.userId) return res.status(401).json(...)
```

## API Key Authentication

```typescript
// src/middleware/api-key.ts
import { Request, Response, NextFunction } from 'express';
import crypto from 'crypto';
import { prisma } from '../lib/prisma';
import { UnauthorizedError, ForbiddenError } from '../lib/errors';

export function apiKeyAuth(requiredScope?: string) {
  return async (req: Request, _res: Response, next: NextFunction) => {
    const apiKey = req.headers['x-api-key'] as string;
    if (!apiKey) return next(new UnauthorizedError('API key required'));

    // Hash the key to compare with stored hash
    const hashedKey = crypto.createHash('sha256').update(apiKey).digest('hex');

    const key = await prisma.apiKey.findUnique({
      where: { hashedKey },
      include: { user: true },
    });

    if (!key || key.revokedAt) {
      return next(new UnauthorizedError('Invalid API key'));
    }

    if (key.expiresAt && key.expiresAt < new Date()) {
      return next(new UnauthorizedError('API key expired'));
    }

    if (requiredScope && !key.scopes.includes(requiredScope)) {
      return next(new ForbiddenError(`API key missing scope: ${requiredScope}`));
    }

    // Update last used timestamp
    await prisma.apiKey.update({
      where: { id: key.id },
      data: { lastUsedAt: new Date() },
    });

    req.user = { id: key.user.id, email: key.user.email, role: key.user.role };
    next();
  };
}

// Generate API key for user
export async function generateApiKey(
  userId: string,
  name: string,
  scopes: string[] = ['read'],
) {
  const rawKey = `sk_${crypto.randomBytes(32).toString('hex')}`;
  const hashedKey = crypto.createHash('sha256').update(rawKey).digest('hex');
  const prefix = rawKey.slice(0, 8); // For identification without exposing full key

  await prisma.apiKey.create({
    data: { hashedKey, prefix, name, scopes, userId },
  });

  // Return raw key ONLY once — user must save it
  return { key: rawKey, prefix, name, scopes };
}
```

## Password Reset

```typescript
// src/services/password-reset.service.ts
import crypto from 'crypto';
import bcrypt from 'bcrypt';
import { prisma } from '../lib/prisma';
import { sendEmail } from '../lib/email';

const RESET_TOKEN_BYTES = 32;
const RESET_TOKEN_EXPIRY = 30 * 60 * 1000; // 30 minutes

export class PasswordResetService {
  async requestReset(email: string) {
    const user = await prisma.user.findUnique({ where: { email } });

    // Always return success (don't reveal if email exists)
    if (!user) return;

    // Invalidate any existing tokens
    await prisma.passwordReset.deleteMany({ where: { userId: user.id } });

    const token = crypto.randomBytes(RESET_TOKEN_BYTES).toString('hex');
    const hashedToken = crypto.createHash('sha256').update(token).digest('hex');

    await prisma.passwordReset.create({
      data: {
        token: hashedToken,
        userId: user.id,
        expiresAt: new Date(Date.now() + RESET_TOKEN_EXPIRY),
      },
    });

    await sendEmail({
      to: email,
      subject: 'Password Reset',
      html: `<a href="${process.env.FRONTEND_URL}/reset-password?token=${token}">Reset Password</a>`,
    });
  }

  async resetPassword(token: string, newPassword: string) {
    const hashedToken = crypto.createHash('sha256').update(token).digest('hex');

    const reset = await prisma.passwordReset.findFirst({
      where: {
        token: hashedToken,
        expiresAt: { gt: new Date() },
      },
    });

    if (!reset) throw new Error('Invalid or expired reset token');

    const hashedPassword = await bcrypt.hash(newPassword, 12);

    await prisma.$transaction([
      prisma.user.update({
        where: { id: reset.userId },
        data: { password: hashedPassword },
      }),
      prisma.passwordReset.delete({ where: { id: reset.id } }),
      // Revoke all refresh tokens (force re-login everywhere)
      prisma.refreshToken.deleteMany({ where: { userId: reset.userId } }),
    ]);
  }
}
```

## OAuth2 (Google Example)

```typescript
// src/services/oauth.service.ts
import { OAuth2Client } from 'google-auth-library';
import { prisma } from '../lib/prisma';
import { env } from '../config/env';

const client = new OAuth2Client(env.GOOGLE_CLIENT_ID);

export class OAuthService {
  async googleLogin(idToken: string) {
    const ticket = await client.verifyIdToken({
      idToken,
      audience: env.GOOGLE_CLIENT_ID,
    });

    const payload = ticket.getPayload();
    if (!payload?.email) throw new Error('Invalid Google token');

    // Find or create user
    let user = await prisma.user.findUnique({
      where: { email: payload.email },
    });

    if (!user) {
      user = await prisma.user.create({
        data: {
          email: payload.email,
          name: payload.name || payload.email.split('@')[0],
          googleId: payload.sub,
          emailVerified: payload.email_verified ?? false,
          avatar: payload.picture,
        },
      });
    } else if (!user.googleId) {
      // Link Google account to existing user
      await prisma.user.update({
        where: { id: user.id },
        data: { googleId: payload.sub, avatar: user.avatar || payload.picture },
      });
    }

    return user;
  }
}
```

## Prisma Schema (Auth Models)

```prisma
model User {
  id            String         @id @default(uuid())
  email         String         @unique
  password      String?        // Null for OAuth-only users
  name          String
  role          String         @default("user")
  googleId      String?        @unique
  avatar        String?
  emailVerified Boolean        @default(false)
  createdAt     DateTime       @default(now())
  updatedAt     DateTime       @updatedAt
  refreshTokens RefreshToken[]
  apiKeys       ApiKey[]
  passwordResets PasswordReset[]
}

model RefreshToken {
  id        String   @id @default(uuid())
  token     String   @unique  // SHA-256 hash
  userId    String
  user      User     @relation(fields: [userId], references: [id], onDelete: Cascade)
  expiresAt DateTime
  createdAt DateTime @default(now())

  @@index([userId])
}

model ApiKey {
  id         String    @id @default(uuid())
  hashedKey  String    @unique
  prefix     String    // First 8 chars for identification
  name       String
  scopes     String[]  @default(["read"])
  userId     String
  user       User      @relation(fields: [userId], references: [id], onDelete: Cascade)
  revokedAt  DateTime?
  expiresAt  DateTime?
  lastUsedAt DateTime?
  createdAt  DateTime  @default(now())

  @@index([userId])
}

model PasswordReset {
  id        String   @id @default(uuid())
  token     String   @unique
  userId    String
  user      User     @relation(fields: [userId], references: [id], onDelete: Cascade)
  expiresAt DateTime
  createdAt DateTime @default(now())

  @@index([userId])
}
```

## Gotchas

1. **Never store raw refresh tokens or API keys.** Always hash with SHA-256 before storing. Compare hashes on lookup. The raw token is returned to the user exactly once.

2. **JWT secrets must be strong.** At least 256 bits of entropy. Use `crypto.randomBytes(64).toString('hex')` to generate. Never hardcode in source.

3. **Refresh token rotation is critical.** On every refresh, delete the old token and issue a new one. If a stolen token is reused, detect it and revoke ALL tokens for that user.

4. **Password reset tokens must be single-use.** Delete immediately after use. Set short expiry (15-30 minutes). Don't reveal whether the email exists ("If an account exists, you'll receive an email").

5. **bcrypt cost factor matters.** Use 12+ in production. Below 10 is too fast for modern hardware. Test that hashing takes 200-500ms on your production hardware.

6. **Don't put sensitive data in JWT payload.** JWTs are base64-encoded, not encrypted. Anyone can decode them. Only include userId, role, and email. Never include password hashes, credit card numbers, or PII.
