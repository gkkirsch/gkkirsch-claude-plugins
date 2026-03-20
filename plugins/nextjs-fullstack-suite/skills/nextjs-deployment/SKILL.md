---
name: nextjs-deployment
description: >
  Deploying Next.js applications — Vercel, Docker, Node.js standalone, static export,
  environment variables, build optimization, and production configuration.
  Triggers: "deploy next.js", "vercel deployment", "next.js docker", "next.js production",
  "next.js build", "static export", "standalone output", "next.config".
  NOT for: routing (use app-router skill), data fetching (use data-fetching skill).
version: 1.0.0
allowed-tools: Read, Grep, Glob, Bash, Edit, Write
---

# Next.js Deployment

## Deployment Options

| Platform | Best For | SSR | Edge | Cost |
|----------|----------|-----|------|------|
| **Vercel** | Fastest path to production | Full | Yes | Free tier, then usage-based |
| **Docker** | Full control, any cloud | Full | No | VPS or container service |
| **Node.js standalone** | Traditional hosting | Full | No | Any Node.js host |
| **Static export** | Static sites only | No | N/A | Any static host (S3, Cloudflare, Netlify) |
| **Cloudflare Pages** | Edge-first, static + functions | Workers | Yes | Generous free tier |
| **Railway / Fly.io** | Simple container hosting | Full | No | Usage-based, cheap |

## Vercel Deployment

### Quick Deploy

```bash
npm i -g vercel
vercel          # Deploy preview
vercel --prod   # Deploy to production
```

### Project Configuration

```json
// vercel.json (usually not needed — Vercel auto-detects Next.js)
{
  "framework": "nextjs",
  "buildCommand": "next build",
  "outputDirectory": ".next",
  "regions": ["iad1"],
  "headers": [
    {
      "source": "/api/(.*)",
      "headers": [
        { "key": "Access-Control-Allow-Origin", "value": "*" }
      ]
    }
  ],
  "rewrites": [
    { "source": "/blog/:slug", "destination": "/posts/:slug" }
  ],
  "redirects": [
    { "source": "/old-page", "destination": "/new-page", "permanent": true }
  ]
}
```

### Environment Variables

```bash
# Set via CLI
vercel env add STRIPE_SECRET_KEY production
vercel env add DATABASE_URL production preview

# Or via Dashboard: Settings → Environment Variables
```

Variables prefixed with `NEXT_PUBLIC_` are exposed to the browser. Everything else stays server-side.

### Preview Deployments

Every push to a branch creates a preview deployment with a unique URL. Use this for:
- PR reviews
- Staging environments
- Testing with preview environment variables

## Docker Deployment

### Optimized Dockerfile

```dockerfile
# Stage 1: Install dependencies
FROM node:20-alpine AS deps
RUN apk add --no-cache libc6-compat
WORKDIR /app
COPY package.json package-lock.json* ./
RUN npm ci --omit=dev

# Stage 2: Build
FROM node:20-alpine AS builder
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .

# Generate Prisma client if using Prisma
# RUN npx prisma generate

ENV NEXT_TELEMETRY_DISABLED=1
RUN npm run build

# Stage 3: Production
FROM node:20-alpine AS runner
WORKDIR /app

ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1

RUN addgroup --system --gid 1001 nodejs
RUN adduser --system --uid 1001 nextjs

# Copy standalone output
COPY --from=builder /app/public ./public
COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/.next/static ./.next/static

USER nextjs

EXPOSE 3000
ENV PORT=3000
ENV HOSTNAME="0.0.0.0"

CMD ["node", "server.js"]
```

### Required next.config.js

```javascript
/** @type {import('next').NextConfig} */
const nextConfig = {
  output: 'standalone', // Required for Docker — bundles node_modules into standalone
};

module.exports = nextConfig;
```

### Docker Compose

```yaml
# docker-compose.yml
version: '3.8'

services:
  web:
    build: .
    ports:
      - "3000:3000"
    environment:
      - DATABASE_URL=postgresql://user:pass@db:5432/myapp
      - STRIPE_SECRET_KEY=${STRIPE_SECRET_KEY}
    depends_on:
      - db

  db:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: user
      POSTGRES_PASSWORD: pass
      POSTGRES_DB: myapp
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"

volumes:
  postgres_data:
```

### Build and Run

```bash
docker build -t myapp .
docker run -p 3000:3000 --env-file .env.production myapp
```

## Node.js Standalone

For traditional hosting (Heroku, Railway, DigitalOcean, etc.):

```javascript
// next.config.js
module.exports = {
  output: 'standalone',
};
```

```bash
# Build
next build

# The standalone output is at .next/standalone/
# Copy static assets
cp -r public .next/standalone/public
cp -r .next/static .next/standalone/.next/static

# Run
cd .next/standalone
node server.js
# Server starts on port 3000
```

### Heroku

```json
// package.json
{
  "scripts": {
    "build": "next build",
    "start": "next start -p $PORT"
  },
  "engines": {
    "node": "20.x"
  }
}
```

```bash
# Procfile
web: npm start
```

```bash
heroku create myapp
heroku config:set NODE_ENV=production
heroku config:set DATABASE_URL=...
git push heroku main
```

### Railway / Fly.io

Both auto-detect Next.js and configure accordingly:

```bash
# Railway
railway init
railway up

# Fly.io (uses Dockerfile)
fly launch
fly deploy
```

## Static Export

For sites with no server-side features:

```javascript
// next.config.js
module.exports = {
  output: 'export',
  // Optional: custom output directory
  // distDir: 'out',
};
```

```bash
next build
# Output in 'out/' directory — deploy to any static host
```

### Limitations of Static Export

- No Server Components with dynamic data
- No API Routes
- No Middleware
- No ISR / revalidation
- No Image Optimization (use `unoptimized: true` or external provider)
- No `headers()`, `cookies()`, `searchParams`

### Deploy Static to Cloudflare Pages

```bash
npx wrangler pages deploy out --project-name=myapp
```

## Environment Variables

### File-Based

```bash
# .env              — All environments (committed, no secrets)
# .env.local        — Local overrides (gitignored, secrets OK)
# .env.development  — Development only
# .env.production   — Production only
# .env.test         — Test only
```

### Load Order (Highest Priority First)

1. `process.env` (runtime)
2. `.env.$(NODE_ENV).local`
3. `.env.local` (not in test)
4. `.env.$(NODE_ENV)`
5. `.env`

### Server vs Client

```bash
# Server-only (never sent to browser)
DATABASE_URL=postgresql://localhost:5432/myapp
STRIPE_SECRET_KEY=sk_live_...

# Client-accessible (bundled into JS)
NEXT_PUBLIC_APP_URL=https://myapp.com
NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY=pk_live_...
```

### Runtime Environment Variables

For Docker/containers where env vars are set at runtime (not build time):

```typescript
// lib/env.ts
export const env = {
  databaseUrl: process.env.DATABASE_URL!,
  stripeKey: process.env.STRIPE_SECRET_KEY!,
  appUrl: process.env.NEXT_PUBLIC_APP_URL!,
};

// Validate at startup
const required = ['DATABASE_URL', 'STRIPE_SECRET_KEY'];
for (const key of required) {
  if (!process.env[key]) {
    throw new Error(`Missing required env var: ${key}`);
  }
}
```

## next.config.js Production Settings

```javascript
/** @type {import('next').NextConfig} */
const nextConfig = {
  // Docker / standalone deployment
  output: 'standalone',

  // Image optimization
  images: {
    remotePatterns: [
      { protocol: 'https', hostname: '**.amazonaws.com' },
    ],
    formats: ['image/avif', 'image/webp'],
  },

  // Redirects (runs at edge, before page renders)
  async redirects() {
    return [
      { source: '/blog', destination: '/posts', permanent: true },
    ];
  },

  // Headers
  async headers() {
    return [
      {
        source: '/(.*)',
        headers: [
          { key: 'X-Frame-Options', value: 'DENY' },
          { key: 'X-Content-Type-Options', value: 'nosniff' },
          { key: 'Referrer-Policy', value: 'strict-origin-when-cross-origin' },
        ],
      },
    ];
  },

  // Webpack customization
  webpack: (config, { isServer }) => {
    // Example: handle .svg as React components
    config.module.rules.push({
      test: /\.svg$/,
      use: ['@svgr/webpack'],
    });
    return config;
  },

  // Experimental features
  experimental: {
    // Partial Prerendering (Next.js 15)
    ppr: true,
    // Server Actions
    serverActions: {
      bodySizeLimit: '2mb',
    },
  },

  // Disable telemetry
  // (or set NEXT_TELEMETRY_DISABLED=1)
};

module.exports = nextConfig;
```

## Build Optimization

### Analyze Bundle Size

```bash
npm install @next/bundle-analyzer

# next.config.js
const withBundleAnalyzer = require('@next/bundle-analyzer')({
  enabled: process.env.ANALYZE === 'true',
});
module.exports = withBundleAnalyzer(nextConfig);

# Run
ANALYZE=true next build
```

### Reduce Build Time

```javascript
// Skip TypeScript and ESLint during build (if CI handles them separately)
module.exports = {
  typescript: { ignoreBuildErrors: true },
  eslint: { ignoreDuringBuilds: true },
};
```

### Caching Build Artifacts

```bash
# GitHub Actions
- uses: actions/cache@v4
  with:
    path: |
      ~/.npm
      ${{ github.workspace }}/.next/cache
    key: ${{ runner.os }}-nextjs-${{ hashFiles('**/package-lock.json') }}
```

## Health Check Endpoint

```typescript
// app/api/health/route.ts
import { NextResponse } from 'next/server';
import { db } from '@/lib/db';

export const dynamic = 'force-dynamic';

export async function GET() {
  try {
    // Check database
    await db.$queryRaw`SELECT 1`;

    return NextResponse.json({
      status: 'healthy',
      timestamp: new Date().toISOString(),
      version: process.env.NEXT_PUBLIC_APP_VERSION || 'unknown',
    });
  } catch (err) {
    return NextResponse.json(
      { status: 'unhealthy', error: 'Database connection failed' },
      { status: 503 }
    );
  }
}
```

## Common Gotchas

1. **`output: 'standalone'` is required for Docker** — without it, the build output still depends on `node_modules`. Standalone bundles everything into `.next/standalone/`.

2. **Static assets need manual copy for standalone** — the `public/` and `.next/static/` dirs aren't included in standalone. Copy them in your Dockerfile.

3. **NEXT_PUBLIC_ vars are baked at build time** — they can't be changed at runtime. For truly dynamic config, use an API route or `__NEXT_DATA__` injection.

4. **Prisma in Docker needs `prisma generate` in build** — the Prisma client must be generated during the build stage, not just at install time.

5. **Image optimization needs sharp in production** — `next/image` uses sharp for optimization. It's auto-installed but verify it's in your production deps.

6. **Vercel auto-sets NODE_ENV=production** — don't set it manually in Vercel. Do set it for Docker/standalone deployments.

7. **Build cache matters** — `.next/cache` significantly speeds up rebuilds. Persist it in CI (GitHub Actions cache, Docker layer cache).

8. **Port configuration** — Vercel manages ports automatically. For Docker/standalone: `next start -p $PORT` or set `PORT` env var. Default is 3000.
