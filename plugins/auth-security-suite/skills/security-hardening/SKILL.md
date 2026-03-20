---
name: security-hardening
description: >
  Security hardening for web applications — CSRF protection, XSS prevention, rate
  limiting, input validation, security headers, CORS configuration, and encryption.
  Triggers: "CSRF", "XSS prevention", "rate limiting", "security headers", "CORS",
  "input validation", "helmet", "sanitize", "encrypt", "hash".
  NOT for: JWT auth (use jwt-auth), OAuth flows (use oauth-patterns).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Security Hardening

## Security Headers (Helmet.js)

```typescript
import helmet from 'helmet';

app.use(helmet({
  contentSecurityPolicy: {
    directives: {
      defaultSrc: ["'self'"],
      scriptSrc: ["'self'", "'unsafe-inline'"],  // Tighten in production
      styleSrc: ["'self'", "'unsafe-inline'", "https://fonts.googleapis.com"],
      imgSrc: ["'self'", "data:", "https:"],
      fontSrc: ["'self'", "https://fonts.gstatic.com"],
      connectSrc: ["'self'", "https://api.example.com"],
      frameSrc: ["'none'"],
      objectSrc: ["'none'"],
      upgradeInsecureRequests: [],
    },
  },
  crossOriginEmbedderPolicy: false, // Set true if using SharedArrayBuffer
  hsts: { maxAge: 31536000, includeSubDomains: true },
}));

// Manual security headers (without helmet)
app.use((req, res, next) => {
  res.setHeader('X-Content-Type-Options', 'nosniff');
  res.setHeader('X-Frame-Options', 'DENY');
  res.setHeader('Referrer-Policy', 'strict-origin-when-cross-origin');
  res.setHeader('Permissions-Policy', 'camera=(), microphone=(), geolocation=()');
  next();
});
```

## CORS Configuration

```typescript
import cors from 'cors';

// Development
app.use(cors({
  origin: 'http://localhost:5173',
  credentials: true, // Allow cookies
}));

// Production — whitelist specific origins
app.use(cors({
  origin: (origin, callback) => {
    const allowedOrigins = [
      'https://myapp.com',
      'https://www.myapp.com',
      'https://admin.myapp.com',
    ];

    // Allow requests with no origin (mobile apps, Postman)
    if (!origin || allowedOrigins.includes(origin)) {
      callback(null, true);
    } else {
      callback(new Error('Not allowed by CORS'));
    }
  },
  credentials: true,
  methods: ['GET', 'POST', 'PUT', 'DELETE', 'PATCH'],
  allowedHeaders: ['Content-Type', 'Authorization', 'X-CSRF-Token'],
  maxAge: 86400, // Preflight cache: 24 hours
}));

// NEVER do this in production:
// app.use(cors({ origin: '*', credentials: true }));
// Browsers block credentials: true with origin: '*'
```

## CSRF Protection

```typescript
import { randomBytes } from 'crypto';

// Generate CSRF token per session
function generateCsrfToken(req: Request): string {
  if (!req.session.csrfToken) {
    req.session.csrfToken = randomBytes(32).toString('hex');
  }
  return req.session.csrfToken;
}

// Middleware to validate CSRF token on mutations
function validateCsrf(req: Request, res: Response, next: NextFunction) {
  if (['GET', 'HEAD', 'OPTIONS'].includes(req.method)) {
    return next(); // Safe methods don't need CSRF
  }

  const token = req.headers['x-csrf-token'] || req.body._csrf;
  if (!token || token !== req.session.csrfToken) {
    return res.status(403).json({ error: 'Invalid CSRF token' });
  }
  next();
}

// Endpoint to get CSRF token
app.get('/api/csrf-token', (req, res) => {
  res.json({ token: generateCsrfToken(req) });
});

// Double Submit Cookie pattern (alternative, stateless)
app.use((req, res, next) => {
  if (!req.cookies['csrf-token']) {
    const token = randomBytes(32).toString('hex');
    res.cookie('csrf-token', token, {
      httpOnly: false, // JS needs to read this
      secure: true,
      sameSite: 'strict',
    });
  }
  next();
});
```

## Rate Limiting

```typescript
import rateLimit from 'express-rate-limit';
import RedisStore from 'rate-limit-redis';
import { createClient } from 'redis';

// General API rate limit
const apiLimiter = rateLimit({
  windowMs: 15 * 60 * 1000, // 15 minutes
  max: 100,                   // 100 requests per window
  standardHeaders: true,      // Return rate limit info in headers
  legacyHeaders: false,
  message: { error: 'Too many requests, try again later' },
});

// Strict rate limit for auth endpoints
const authLimiter = rateLimit({
  windowMs: 15 * 60 * 1000,
  max: 5,                     // 5 attempts per 15 min
  skipSuccessfulRequests: true, // Only count failures
  message: { error: 'Too many login attempts' },
});

// Redis store for distributed rate limiting
const redisClient = createClient({ url: process.env.REDIS_URL });
const redisLimiter = rateLimit({
  windowMs: 15 * 60 * 1000,
  max: 100,
  store: new RedisStore({
    sendCommand: (...args: string[]) => redisClient.sendCommand(args),
  }),
});

app.use('/api/', apiLimiter);
app.use('/api/auth/login', authLimiter);
app.use('/api/auth/register', authLimiter);
```

## Input Validation & Sanitization

```typescript
import { z } from 'zod';
import DOMPurify from 'isomorphic-dompurify';

// Schema validation with Zod
const createPostSchema = z.object({
  title: z.string().min(1).max(200).trim(),
  body: z.string().min(1).max(10000),
  tags: z.array(z.string().max(50)).max(10).optional(),
  isPublished: z.boolean().default(false),
});

// Validation middleware factory
function validate(schema: z.ZodSchema) {
  return (req: Request, res: Response, next: NextFunction) => {
    const result = schema.safeParse(req.body);
    if (!result.success) {
      return res.status(400).json({
        error: 'Validation failed',
        details: result.error.flatten().fieldErrors,
      });
    }
    req.body = result.data; // Use parsed/transformed data
    next();
  };
}

app.post('/api/posts', authenticate, validate(createPostSchema), createPost);

// HTML sanitization (when you MUST accept HTML)
function sanitizeHtml(dirty: string): string {
  return DOMPurify.sanitize(dirty, {
    ALLOWED_TAGS: ['b', 'i', 'em', 'strong', 'a', 'p', 'br', 'ul', 'ol', 'li'],
    ALLOWED_ATTR: ['href', 'target', 'rel'],
  });
}

// SQL parameterization (never concatenate!)
// BAD:  db.query(`SELECT * FROM users WHERE id = '${userId}'`)
// GOOD: db.query('SELECT * FROM users WHERE id = $1', [userId])
```

## XSS Prevention

```typescript
// React: JSX auto-escapes by default ✓
// This is SAFE:
<p>{userInput}</p>  // Auto-escaped

// This is DANGEROUS:
<div dangerouslySetInnerHTML={{ __html: userInput }} />  // XSS!

// If you must render HTML, sanitize first:
<div dangerouslySetInnerHTML={{ __html: DOMPurify.sanitize(userInput) }} />

// Content Security Policy nonce for inline scripts
import { randomBytes } from 'crypto';

app.use((req, res, next) => {
  const nonce = randomBytes(16).toString('base64');
  res.locals.nonce = nonce;
  res.setHeader(
    'Content-Security-Policy',
    `script-src 'self' 'nonce-${nonce}'`
  );
  next();
});

// In template: <script nonce="<%= nonce %>">...</script>
```

## Password Hashing

```typescript
import bcrypt from 'bcrypt';
import { randomBytes } from 'crypto';

// Hash password (on registration/change)
async function hashPassword(password: string): Promise<string> {
  return bcrypt.hash(password, 12); // Cost factor 12
}

// Verify password (on login)
async function verifyPassword(password: string, hash: string): Promise<boolean> {
  return bcrypt.compare(password, hash);
}

// Password requirements
const passwordSchema = z.string()
  .min(8, 'Password must be at least 8 characters')
  .max(128, 'Password must be at most 128 characters')
  .regex(/[A-Z]/, 'Must contain uppercase letter')
  .regex(/[a-z]/, 'Must contain lowercase letter')
  .regex(/[0-9]/, 'Must contain number');

// Generate secure random tokens
function generateToken(bytes = 32): string {
  return randomBytes(bytes).toString('hex');
}

// Password reset token (time-limited)
async function createPasswordResetToken(userId: string) {
  const token = generateToken();
  const expiresAt = new Date(Date.now() + 60 * 60 * 1000); // 1 hour

  await db.passwordResets.create({
    userId,
    tokenHash: await hashPassword(token), // Hash the token too!
    expiresAt,
  });

  return token; // Send this in the reset email
}
```

## Environment Variable Security

```typescript
// .env.example (commit this, not .env)
DATABASE_URL=postgresql://user:pass@localhost:5432/mydb
JWT_SECRET=generate-with-openssl-rand-hex-32
SESSION_SECRET=generate-with-openssl-rand-hex-32
GOOGLE_CLIENT_ID=your-client-id
GOOGLE_CLIENT_SECRET=your-client-secret

// Validate required env vars at startup
import { z } from 'zod';

const envSchema = z.object({
  NODE_ENV: z.enum(['development', 'production', 'test']).default('development'),
  DATABASE_URL: z.string().url(),
  JWT_SECRET: z.string().min(32),
  PORT: z.coerce.number().default(3000),
});

// Throws at startup if env is misconfigured
export const env = envSchema.parse(process.env);
```

## Error Handling (Security-Aware)

```typescript
// Global error handler — never leak stack traces
app.use((err: Error, req: Request, res: Response, next: NextFunction) => {
  // Log full error for debugging
  console.error('Error:', {
    message: err.message,
    stack: err.stack,
    url: req.url,
    method: req.method,
    // Never log: req.body (may contain passwords)
  });

  // Send safe error to client
  if (err.name === 'ZodError') {
    return res.status(400).json({ error: 'Validation failed' });
  }

  if (err.name === 'UnauthorizedError') {
    return res.status(401).json({ error: 'Authentication required' });
  }

  // Generic 500 — never expose internals
  res.status(500).json({
    error: process.env.NODE_ENV === 'production'
      ? 'Internal server error'
      : err.message,
  });
});
```

## Gotchas

1. **`SameSite=Lax` vs `Strict` vs `None`.** `Lax` allows cookies on top-level navigations (link clicks) but blocks cross-site POSTs — good default. `Strict` blocks all cross-site requests including link clicks (can break OAuth redirects). `None` requires `Secure` and allows all cross-site requests.

2. **CORS does NOT prevent server-side requests.** CORS is enforced by browsers only. APIs are always accessible via curl/Postman. Don't rely on CORS as your only access control — always authenticate and authorize.

3. **`express-rate-limit` uses in-memory storage by default.** This means rate limits reset on restart and don't work across multiple server instances. Use Redis store in production.

4. **`helmet()` sets CSP to `default-src 'self'` by default.** This blocks inline styles, external fonts, images from CDNs, etc. You'll need to customize `contentSecurityPolicy.directives` for your specific app needs.

5. **Environment variables in Next.js are tricky.** Only `NEXT_PUBLIC_*` variables are available client-side. Server-only secrets must NOT have the `NEXT_PUBLIC_` prefix. Leaking a `NEXT_PUBLIC_JWT_SECRET` exposes it in the browser bundle.

6. **HTML sanitization libraries can be bypassed.** DOMPurify is the most battle-tested, but new bypass vectors are discovered regularly. Keep it updated. If you don't need to accept HTML, don't — use Markdown and render it safely.
