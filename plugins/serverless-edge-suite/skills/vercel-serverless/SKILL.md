---
name: vercel-serverless
description: >
  Build Vercel Serverless and Edge Functions — API routes, middleware,
  streaming, cron jobs, and deployment patterns.
  Triggers: "Vercel function", "Vercel API", "Vercel edge", "Vercel serverless",
  "Next.js API route".
  NOT for: AWS Lambda, Cloudflare Workers, standalone Express servers.
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Vercel Serverless & Edge Functions

## Two Runtimes

| Feature | Node.js Runtime | Edge Runtime |
|---------|----------------|--------------|
| Cold start | ~250ms | ~0ms |
| Max duration | 60s (Hobby), 300s (Pro) | 30s |
| Memory | 1024 MB | 128 MB |
| Node APIs | Full | Web APIs only |
| npm packages | All | Edge-compatible only |
| Database | Direct connections | HTTP-based only |
| Streaming | Yes | Yes |
| Location | Single region | 200+ edge locations |

**Rule of thumb**: Use Edge for auth checks, redirects, feature flags. Use Node.js for database queries, heavy compute, file operations.

## Next.js API Routes (App Router)

### Node.js Runtime (Default)

```typescript
// app/api/users/route.ts
import { NextRequest, NextResponse } from 'next/server';
import { db } from '@/lib/db';

export async function GET(request: NextRequest) {
  const { searchParams } = new URL(request.url);
  const page = parseInt(searchParams.get('page') ?? '1');
  const limit = parseInt(searchParams.get('limit') ?? '20');

  const users = await db.user.findMany({
    skip: (page - 1) * limit,
    take: limit,
    orderBy: { createdAt: 'desc' },
  });

  const total = await db.user.count();

  return NextResponse.json({
    data: users,
    meta: { page, limit, total, totalPages: Math.ceil(total / limit) },
  });
}

export async function POST(request: NextRequest) {
  const body = await request.json();

  // Validate with Zod
  const parsed = CreateUserSchema.safeParse(body);
  if (!parsed.success) {
    return NextResponse.json(
      { error: 'Validation failed', details: parsed.error.flatten() },
      { status: 422 }
    );
  }

  const user = await db.user.create({ data: parsed.data });

  return NextResponse.json(user, { status: 201 });
}
```

### Edge Runtime

```typescript
// app/api/hello/route.ts
export const runtime = 'edge'; // Opt into edge runtime

export async function GET(request: Request) {
  const country = request.headers.get('x-vercel-ip-country') ?? 'US';

  return new Response(JSON.stringify({
    message: `Hello from the edge! You're visiting from ${country}`,
    timestamp: Date.now(),
  }), {
    headers: { 'Content-Type': 'application/json' },
  });
}
```

### Dynamic Route Parameters

```typescript
// app/api/users/[id]/route.ts
import { NextRequest, NextResponse } from 'next/server';

export async function GET(
  request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const { id } = await params;

  const user = await db.user.findUnique({ where: { id } });
  if (!user) {
    return NextResponse.json({ error: 'Not found' }, { status: 404 });
  }

  return NextResponse.json(user);
}

export async function DELETE(
  request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const { id } = await params;

  await db.user.delete({ where: { id } });
  return new NextResponse(null, { status: 204 });
}
```

## Middleware (Edge by Default)

```typescript
// middleware.ts (project root)
import { NextRequest, NextResponse } from 'next/server';

export function middleware(request: NextRequest) {
  // Auth check
  const token = request.cookies.get('session')?.value;
  if (!token && request.nextUrl.pathname.startsWith('/dashboard')) {
    return NextResponse.redirect(new URL('/login', request.url));
  }

  // Add request ID
  const requestId = crypto.randomUUID();
  const response = NextResponse.next();
  response.headers.set('x-request-id', requestId);

  return response;
}

// Only run middleware on specific paths
export const config = {
  matcher: [
    '/dashboard/:path*',
    '/api/:path*',
    // Skip static files and images
    '/((?!_next/static|_next/image|favicon.ico).*)',
  ],
};
```

### Middleware Patterns

```typescript
// Geo-based redirects
export function middleware(request: NextRequest) {
  const country = request.geo?.country ?? 'US';

  if (country === 'DE' && !request.nextUrl.pathname.startsWith('/de')) {
    return NextResponse.redirect(new URL(`/de${request.nextUrl.pathname}`, request.url));
  }

  return NextResponse.next();
}

// A/B testing
export function middleware(request: NextRequest) {
  const variant = request.cookies.get('variant')?.value
    ?? (Math.random() < 0.5 ? 'a' : 'b');

  const response = NextResponse.next();

  if (!request.cookies.has('variant')) {
    response.cookies.set('variant', variant, { maxAge: 60 * 60 * 24 * 7 });
  }

  // Rewrite to variant page
  if (request.nextUrl.pathname === '/pricing') {
    return NextResponse.rewrite(
      new URL(`/pricing/${variant}`, request.url),
      { headers: response.headers }
    );
  }

  return response;
}
```

## Streaming Responses

```typescript
// app/api/stream/route.ts
export const runtime = 'edge';

export async function GET() {
  const encoder = new TextEncoder();

  const stream = new ReadableStream({
    async start(controller) {
      for (let i = 0; i < 10; i++) {
        controller.enqueue(
          encoder.encode(`data: ${JSON.stringify({ count: i })}\n\n`)
        );
        await new Promise(resolve => setTimeout(resolve, 500));
      }
      controller.enqueue(encoder.encode('data: [DONE]\n\n'));
      controller.close();
    },
  });

  return new Response(stream, {
    headers: {
      'Content-Type': 'text/event-stream',
      'Cache-Control': 'no-cache',
      'Connection': 'keep-alive',
    },
  });
}
```

## Cron Jobs

```json
// vercel.json
{
  "crons": [
    {
      "path": "/api/cron/cleanup",
      "schedule": "0 0 * * *"
    },
    {
      "path": "/api/cron/send-digest",
      "schedule": "0 9 * * 1"
    }
  ]
}
```

```typescript
// app/api/cron/cleanup/route.ts
import { NextRequest, NextResponse } from 'next/server';

export async function GET(request: NextRequest) {
  // Verify the request is from Vercel Cron
  const authHeader = request.headers.get('authorization');
  if (authHeader !== `Bearer ${process.env.CRON_SECRET}`) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
  }

  const deleted = await db.session.deleteMany({
    where: { expiresAt: { lt: new Date() } },
  });

  return NextResponse.json({ deleted: deleted.count });
}
```

## Standalone Serverless Functions (No Next.js)

```typescript
// api/hello.ts (in project root /api/ directory)
import type { VercelRequest, VercelResponse } from '@vercel/node';

export default function handler(req: VercelRequest, res: VercelResponse) {
  const { name = 'World' } = req.query;
  res.status(200).json({ message: `Hello, ${name}!` });
}
```

```typescript
// api/users.ts — Edge runtime standalone
export const config = { runtime: 'edge' };

export default async function handler(request: Request) {
  return new Response(JSON.stringify({ users: [] }), {
    headers: { 'Content-Type': 'application/json' },
  });
}
```

## Environment Variables

```bash
# Set via CLI
vercel env add DATABASE_URL production
vercel env add STRIPE_KEY production preview

# Or in Vercel dashboard: Settings → Environment Variables

# In code:
const dbUrl = process.env.DATABASE_URL!;
```

`.env.local` for local development (gitignored).

## Deployment

```bash
# Install Vercel CLI
npm i -g vercel

# Deploy to preview
vercel

# Deploy to production
vercel --prod

# Or auto-deploy via GitHub integration (recommended)
# Push to main → auto-deploys to production
# Push to branch → auto-deploys to preview URL
```

## Gotchas

- **Edge Runtime has no `fs` module.** Can't read files. Use `fetch()` for everything, including your own API routes.
- **Middleware runs on EVERY request.** Use the `matcher` config to limit it. Without a matcher, middleware runs on static file requests too.
- **Vercel cron requires `CRON_SECRET`.** Set it in env vars and check `Authorization: Bearer <CRON_SECRET>` in the handler.
- **API routes are serverless functions.** Each file in `app/api/` becomes a separate Lambda. Don't share state between them (no global variables that persist across requests to different routes).
- **Cold starts on free tier.** Functions go cold after ~5 minutes of inactivity. Use `edge` runtime for zero cold starts on latency-sensitive routes.
- **Body size limits.** 4.5 MB for Node.js functions, 4 MB for Edge. For larger uploads, use presigned URLs to S3/R2.
- **Duration limits are hard.** If your function hits the timeout, it's killed. For long-running work, use background functions or queue to an external service.
- **`request.json()` can only be called once.** If you need the body in middleware AND the handler, clone the request first.
