---
name: web-security
description: >
  Web application security — XSS prevention, CSRF protection, SQL injection,
  Content Security Policy, CORS, rate limiting, input sanitization,
  security headers, and common vulnerability patterns.
  Triggers: "xss", "csrf", "sql injection", "security headers", "csp",
  "cors security", "rate limiting", "input sanitization", "owasp".
  NOT for: authentication/authorization (use auth-security).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Web Application Security

## XSS Prevention

```typescript
// NEVER insert user input into HTML without escaping
// BAD: element.innerHTML = userInput;
// BAD: res.send(`<div>${userInput}</div>`);

// Escape HTML entities
function escapeHtml(str: string): string {
  const map: Record<string, string> = {
    "&": "&amp;", "<": "&lt;", ">": "&gt;",
    '"': "&quot;", "'": "&#39;", "/": "&#x2F;",
  };
  return str.replace(/[&<>"'/]/g, (c) => map[c]);
}

// React is safe by default (auto-escapes JSX)
// But these are dangerous:
// - dangerouslySetInnerHTML={{ __html: userContent }}
// - href={`javascript:${userInput}`}
// - Using user input in eval(), Function(), setTimeout(string)

// Sanitize HTML when you MUST allow some markup
import DOMPurify from "dompurify";

const cleanHtml = DOMPurify.sanitize(userHtml, {
  ALLOWED_TAGS: ["b", "i", "em", "strong", "a", "p", "br"],
  ALLOWED_ATTR: ["href", "target"],
  ALLOW_DATA_ATTR: false,
});
```

## CSRF Protection

```typescript
// Express CSRF middleware
import { doubleCsrf } from "csrf-csrf";

const { doubleCsrfProtection, generateToken } = doubleCsrf({
  getSecret: () => process.env.CSRF_SECRET!,
  cookieName: "__csrf",
  cookieOptions: {
    httpOnly: true,
    sameSite: "strict",
    secure: process.env.NODE_ENV === "production",
  },
  getTokenFromRequest: (req) => req.headers["x-csrf-token"] as string,
});

// Apply to all state-changing routes
app.use("/api", doubleCsrfProtection);

// Provide token to client
app.get("/api/csrf-token", (req, res) => {
  res.json({ token: generateToken(req, res) });
});

// Client sends token in header
fetch("/api/data", {
  method: "POST",
  headers: {
    "Content-Type": "application/json",
    "X-CSRF-Token": csrfToken,
  },
  body: JSON.stringify(data),
});

// SameSite cookies provide strong CSRF protection
// Set on all auth cookies:
res.cookie("session", token, {
  httpOnly: true,
  secure: true,
  sameSite: "strict",  // or "lax" for GET navigation
  maxAge: 86400000,
});
```

## SQL Injection Prevention

```typescript
// NEVER concatenate user input into SQL
// BAD: `SELECT * FROM users WHERE id = '${userId}'`
// BAD: `SELECT * FROM users WHERE name LIKE '%${search}%'`

// Use parameterized queries (always)
// PostgreSQL (node-postgres)
const result = await pool.query(
  "SELECT * FROM users WHERE id = $1 AND status = $2",
  [userId, "active"]
);

// MySQL
const [rows] = await connection.execute(
  "SELECT * FROM users WHERE id = ? AND status = ?",
  [userId, "active"]
);

// Prisma (ORM — safe by default)
const user = await prisma.user.findUnique({
  where: { id: userId },
});

// Even Prisma raw queries need parameterization:
const result = await prisma.$queryRaw`
  SELECT * FROM users WHERE name ILIKE ${`%${search}%`}
`;
// Prisma tagged templates auto-parameterize

// BAD Prisma raw:
// prisma.$queryRawUnsafe(`SELECT * FROM users WHERE id = '${id}'`)
```

## Content Security Policy

```typescript
// Helmet.js for Express
import helmet from "helmet";

app.use(
  helmet({
    contentSecurityPolicy: {
      directives: {
        defaultSrc: ["'self'"],
        scriptSrc: ["'self'", "'strict-dynamic'"],  // No inline scripts
        styleSrc: ["'self'", "'unsafe-inline'"],     // Tailwind needs inline
        imgSrc: ["'self'", "data:", "https://cdn.example.com"],
        connectSrc: ["'self'", "https://api.example.com"],
        fontSrc: ["'self'", "https://fonts.gstatic.com"],
        objectSrc: ["'none'"],
        frameAncestors: ["'none'"],        // Prevent clickjacking
        baseUri: ["'self'"],
        formAction: ["'self'"],
        upgradeInsecureRequests: [],        // Force HTTPS
      },
    },
    crossOriginEmbedderPolicy: true,
    crossOriginOpenerPolicy: true,
    crossOriginResourcePolicy: { policy: "same-origin" },
    hsts: { maxAge: 31536000, includeSubDomains: true, preload: true },
    referrerPolicy: { policy: "strict-origin-when-cross-origin" },
  })
);

// Next.js — in next.config.ts
const securityHeaders = [
  { key: "X-Frame-Options", value: "DENY" },
  { key: "X-Content-Type-Options", value: "nosniff" },
  { key: "X-XSS-Protection", value: "0" },  // Deprecated, use CSP instead
  { key: "Referrer-Policy", value: "strict-origin-when-cross-origin" },
  { key: "Permissions-Policy", value: "camera=(), microphone=(), geolocation=()" },
  {
    key: "Content-Security-Policy",
    value: "default-src 'self'; script-src 'self' 'strict-dynamic'; style-src 'self' 'unsafe-inline';",
  },
];
```

## CORS Configuration

```typescript
import cors from "cors";

// Strict CORS (production)
app.use(cors({
  origin: (origin, callback) => {
    const allowed = [
      "https://myapp.com",
      "https://www.myapp.com",
      ...(process.env.NODE_ENV !== "production" ? ["http://localhost:3000"] : []),
    ];
    if (!origin || allowed.includes(origin)) {
      callback(null, true);
    } else {
      callback(new Error("Not allowed by CORS"));
    }
  },
  methods: ["GET", "POST", "PUT", "DELETE"],
  allowedHeaders: ["Content-Type", "Authorization", "X-CSRF-Token"],
  credentials: true,      // Allow cookies
  maxAge: 86400,           // Cache preflight for 24h
}));

// NEVER do this in production:
// app.use(cors({ origin: "*", credentials: true }));
// credentials: true with origin: "*" is blocked by browsers
```

## Rate Limiting

```typescript
import rateLimit from "express-rate-limit";
import RedisStore from "rate-limit-redis";
import Redis from "ioredis";

const redis = new Redis(process.env.REDIS_URL);

// Global rate limit
const globalLimiter = rateLimit({
  windowMs: 15 * 60 * 1000,  // 15 minutes
  max: 100,                    // 100 requests per window
  standardHeaders: true,       // Return rate limit info in headers
  legacyHeaders: false,
  store: new RedisStore({ sendCommand: (...args) => redis.call(...args) }),
  message: { error: "Too many requests, please try again later" },
});

// Strict limit for auth endpoints
const authLimiter = rateLimit({
  windowMs: 15 * 60 * 1000,
  max: 5,  // 5 login attempts per 15 minutes
  skipSuccessfulRequests: true,
  keyGenerator: (req) => req.body?.email || req.ip,
  message: { error: "Too many login attempts" },
});

app.use("/api", globalLimiter);
app.use("/api/auth/login", authLimiter);
app.use("/api/auth/register", authLimiter);
```

## Input Validation

```typescript
import { z } from "zod";

// Validate and sanitize ALL input
const CreateUserSchema = z.object({
  name: z.string()
    .trim()
    .min(1, "Name required")
    .max(100, "Name too long")
    .regex(/^[\p{L}\p{N}\s'-]+$/u, "Invalid characters"),
  email: z.string()
    .trim()
    .toLowerCase()
    .email("Invalid email")
    .max(254),
  password: z.string()
    .min(8, "At least 8 characters")
    .max(128)
    .regex(/[A-Z]/, "Need uppercase letter")
    .regex(/[0-9]/, "Need a number"),
  role: z.enum(["user", "editor"]),
  bio: z.string().max(500).optional(),
  website: z.string().url().optional().or(z.literal("")),
});

// Middleware
function validate(schema: z.ZodSchema) {
  return (req: Request, res: Response, next: NextFunction) => {
    const result = schema.safeParse(req.body);
    if (!result.success) {
      return res.status(400).json({
        error: "Validation failed",
        details: result.error.flatten().fieldErrors,
      });
    }
    req.body = result.data;  // Use sanitized data
    next();
  };
}

// URL validation — prevent SSRF
function isAllowedUrl(url: string): boolean {
  try {
    const parsed = new URL(url);
    // Block internal networks
    if (["localhost", "127.0.0.1", "0.0.0.0", "::1"].includes(parsed.hostname)) return false;
    if (parsed.hostname.endsWith(".internal")) return false;
    if (parsed.protocol !== "https:") return false;
    return true;
  } catch {
    return false;
  }
}
```

## Security Headers Checklist

| Header | Value | Purpose |
|--------|-------|---------|
| `Strict-Transport-Security` | `max-age=31536000; includeSubDomains; preload` | Force HTTPS |
| `Content-Security-Policy` | See CSP section | Prevent XSS, injection |
| `X-Content-Type-Options` | `nosniff` | Prevent MIME sniffing |
| `X-Frame-Options` | `DENY` | Prevent clickjacking |
| `Referrer-Policy` | `strict-origin-when-cross-origin` | Control referrer info |
| `Permissions-Policy` | `camera=(), microphone=()` | Restrict browser features |
| `Cross-Origin-Opener-Policy` | `same-origin` | Isolate browsing context |
| `Cross-Origin-Resource-Policy` | `same-origin` | Prevent cross-origin reads |

## Gotchas

1. **React auto-escaping doesn't protect `href` attributes** — `<a href={userInput}>` allows `javascript:alert(1)` URLs. Validate URLs: only allow `https://` and `http://` protocols.

2. **`SameSite=Lax` allows GET-based CSRF** — Lax cookies are sent on top-level GET navigations (links). If your GET endpoints have side effects (delete via GET), use `SameSite=Strict` or check the `Origin` header.

3. **Rate limiting by IP fails behind reverse proxies** — If your app is behind Nginx/CloudFlare, `req.ip` is the proxy's IP. Set `app.set('trust proxy', 1)` to read `X-Forwarded-For`.

4. **CSP `'unsafe-inline'` for styles defeats XSS protection** — Many CSS-in-JS libraries need it. Use nonces (`'nonce-abc123'`) instead. Tailwind class-based styling doesn't need `unsafe-inline`.

5. **CORS preflight caching** — Without `maxAge`, browsers send an OPTIONS preflight for every cross-origin request. Set `maxAge: 86400` to cache for 24 hours.

6. **`httpOnly` cookies are not accessible from JavaScript** — This is the point. But it means your client-side code can't read session tokens or CSRF tokens from cookies. Use a separate API endpoint to provide CSRF tokens.
