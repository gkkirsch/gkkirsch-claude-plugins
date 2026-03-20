---
name: edge-patterns
description: >
  Common edge computing patterns — request routing, caching strategies,
  feature flags, bot detection, and content transformation at the edge.
  Triggers: "edge computing patterns", "edge caching", "CDN logic",
  "edge feature flags", "request routing".
  NOT for: platform-specific setup (use aws-lambda, vercel-serverless, or
  cloudflare-workers skills for that).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Edge Computing Patterns

Framework-agnostic patterns that work on Cloudflare Workers, Vercel Edge, Deno Deploy, and any V8-based edge runtime.

## Pattern 1: Smart Caching

### Cache-Aside at the Edge

```typescript
async function handleRequest(request: Request, env: Env): Promise<Response> {
  const cacheKey = new URL(request.url).pathname;

  // 1. Check edge cache (KV or Cache API)
  const cached = await env.KV.get(cacheKey, { type: 'json' });
  if (cached) {
    return new Response(JSON.stringify(cached), {
      headers: {
        'Content-Type': 'application/json',
        'X-Cache': 'HIT',
        'Cache-Control': 'public, max-age=60',
      },
    });
  }

  // 2. Fetch from origin
  const originResponse = await fetch(`${env.ORIGIN_URL}${cacheKey}`);
  const data = await originResponse.json();

  // 3. Store in edge cache (non-blocking)
  ctx.waitUntil(
    env.KV.put(cacheKey, JSON.stringify(data), { expirationTtl: 300 })
  );

  return new Response(JSON.stringify(data), {
    headers: {
      'Content-Type': 'application/json',
      'X-Cache': 'MISS',
      'Cache-Control': 'public, max-age=60',
    },
  });
}
```

### Stale-While-Revalidate

```typescript
async function staleWhileRevalidate(
  request: Request,
  env: Env,
  ctx: ExecutionContext
): Promise<Response> {
  const cacheKey = request.url;

  const { value: cached, metadata } = await env.KV.getWithMetadata(cacheKey, {
    type: 'text',
  });

  if (cached) {
    const age = Date.now() - (metadata?.cachedAt ?? 0);
    const maxAge = 60_000;      // 1 minute "fresh"
    const staleAge = 300_000;   // 5 minutes "stale but serveable"

    if (age < maxAge) {
      // Fresh — serve directly
      return new Response(cached, {
        headers: { 'X-Cache': 'HIT', 'X-Cache-Age': String(age) },
      });
    }

    if (age < staleAge) {
      // Stale — serve immediately, revalidate in background
      ctx.waitUntil(revalidate(request, env, cacheKey));
      return new Response(cached, {
        headers: { 'X-Cache': 'STALE', 'X-Cache-Age': String(age) },
      });
    }
  }

  // Expired or no cache — fetch and cache
  return revalidate(request, env, cacheKey);
}

async function revalidate(request: Request, env: Env, key: string): Promise<Response> {
  const response = await fetch(request);
  const body = await response.text();

  await env.KV.put(key, body, {
    expirationTtl: 600,
    metadata: { cachedAt: Date.now() },
  });

  return new Response(body, {
    headers: { ...Object.fromEntries(response.headers), 'X-Cache': 'MISS' },
  });
}
```

## Pattern 2: Feature Flags at the Edge

```typescript
interface FeatureFlags {
  newCheckout: boolean;
  darkMode: boolean;
  betaApi: boolean;
  maintenanceMode: boolean;
}

async function getFeatureFlags(env: Env): Promise<FeatureFlags> {
  const flags = await env.KV.get('feature-flags', { type: 'json' });
  return flags ?? {
    newCheckout: false,
    darkMode: false,
    betaApi: false,
    maintenanceMode: false,
  };
}

async function handleRequest(request: Request, env: Env): Promise<Response> {
  const flags = await getFeatureFlags(env);

  // Maintenance mode — block all traffic
  if (flags.maintenanceMode) {
    return new Response(
      JSON.stringify({ message: 'Service is undergoing maintenance' }),
      { status: 503, headers: { 'Content-Type': 'application/json', 'Retry-After': '300' } }
    );
  }

  // Route to new checkout if flag is enabled
  const url = new URL(request.url);
  if (flags.newCheckout && url.pathname.startsWith('/checkout')) {
    url.pathname = url.pathname.replace('/checkout', '/checkout-v2');
    return fetch(url.toString(), request);
  }

  return fetch(request);
}

// Update flags via KV write (from admin API or CLI):
// await env.KV.put('feature-flags', JSON.stringify({ newCheckout: true, ... }))
```

### Percentage Rollout

```typescript
function isInRollout(userId: string, percentage: number): boolean {
  // Deterministic hash — same user always gets same result
  let hash = 0;
  for (let i = 0; i < userId.length; i++) {
    hash = ((hash << 5) - hash) + userId.charCodeAt(i);
    hash |= 0;
  }
  return (Math.abs(hash) % 100) < percentage;
}

async function handleRequest(request: Request, env: Env) {
  const userId = getUserIdFromCookie(request) ?? request.headers.get('cf-connecting-ip')!;
  const flags = await getFeatureFlags(env);

  if (flags.newCheckout && isInRollout(userId, 25)) {
    // 25% of users get new checkout
    return fetch(new URL('/checkout-v2', request.url), request);
  }

  return fetch(request);
}
```

## Pattern 3: Rate Limiting at the Edge

```typescript
async function rateLimitAtEdge(
  request: Request,
  env: Env,
  limit: number = 100,
  windowSeconds: number = 60
): Promise<Response | null> {
  const ip = request.headers.get('cf-connecting-ip')
    ?? request.headers.get('x-forwarded-for')
    ?? 'unknown';

  const key = `ratelimit:${ip}`;
  const current = await env.KV.get(key, { type: 'json' }) as {
    count: number;
    resetAt: number;
  } | null;

  const now = Date.now();

  if (current && now < current.resetAt) {
    if (current.count >= limit) {
      const retryAfter = Math.ceil((current.resetAt - now) / 1000);
      return new Response(
        JSON.stringify({ error: 'Rate limit exceeded' }),
        {
          status: 429,
          headers: {
            'Content-Type': 'application/json',
            'Retry-After': String(retryAfter),
            'RateLimit-Limit': String(limit),
            'RateLimit-Remaining': '0',
            'RateLimit-Reset': String(Math.ceil(current.resetAt / 1000)),
          },
        }
      );
    }

    // Increment
    await env.KV.put(key, JSON.stringify({
      count: current.count + 1,
      resetAt: current.resetAt,
    }), { expirationTtl: windowSeconds });
  } else {
    // New window
    await env.KV.put(key, JSON.stringify({
      count: 1,
      resetAt: now + (windowSeconds * 1000),
    }), { expirationTtl: windowSeconds });
  }

  return null; // Not rate limited — continue to handler
}
```

## Pattern 4: Request Signing and Validation

```typescript
async function verifyWebhook(request: Request, secret: string): Promise<boolean> {
  const signature = request.headers.get('x-signature-256');
  if (!signature) return false;

  const body = await request.text();
  const encoder = new TextEncoder();

  const key = await crypto.subtle.importKey(
    'raw',
    encoder.encode(secret),
    { name: 'HMAC', hash: 'SHA-256' },
    false,
    ['verify']
  );

  const sigBuffer = hexToArrayBuffer(signature.replace('sha256=', ''));

  return crypto.subtle.verify(
    'HMAC',
    key,
    sigBuffer,
    encoder.encode(body)
  );
}

function hexToArrayBuffer(hex: string): ArrayBuffer {
  const bytes = new Uint8Array(hex.length / 2);
  for (let i = 0; i < hex.length; i += 2) {
    bytes[i / 2] = parseInt(hex.substr(i, 2), 16);
  }
  return bytes.buffer;
}
```

## Pattern 5: Edge-Side Includes (Content Assembly)

```typescript
async function assemblePageAtEdge(request: Request, env: Env): Promise<Response> {
  // Fetch multiple fragments in parallel
  const [header, content, footer, sidebar] = await Promise.all([
    fetch(`${env.ORIGIN}/fragments/header`).then(r => r.text()),
    fetch(`${env.ORIGIN}${new URL(request.url).pathname}`).then(r => r.text()),
    fetch(`${env.ORIGIN}/fragments/footer`).then(r => r.text()),
    fetch(`${env.ORIGIN}/fragments/sidebar`).then(r => r.text()),
  ]);

  const html = `
    <!DOCTYPE html>
    <html>
      <body>
        ${header}
        <div class="layout">
          <main>${content}</main>
          <aside>${sidebar}</aside>
        </div>
        ${footer}
      </body>
    </html>
  `;

  return new Response(html, {
    headers: {
      'Content-Type': 'text/html',
      'Cache-Control': 'public, max-age=60, stale-while-revalidate=300',
    },
  });
}
```

## Pattern 6: API Gateway at the Edge

```typescript
const ROUTES: Record<string, string> = {
  '/api/users': 'https://users-service.internal',
  '/api/orders': 'https://orders-service.internal',
  '/api/products': 'https://products-service.internal',
};

async function apiGateway(request: Request, env: Env): Promise<Response> {
  const url = new URL(request.url);

  // Find matching backend
  const matchedPath = Object.keys(ROUTES).find(path =>
    url.pathname.startsWith(path)
  );

  if (!matchedPath) {
    return new Response(JSON.stringify({ error: 'Not found' }), {
      status: 404,
      headers: { 'Content-Type': 'application/json' },
    });
  }

  // Forward to backend
  const backendUrl = `${ROUTES[matchedPath]}${url.pathname}${url.search}`;
  const backendRequest = new Request(backendUrl, {
    method: request.method,
    headers: request.headers,
    body: request.body,
  });

  // Add tracing headers
  backendRequest.headers.set('X-Request-Id', crypto.randomUUID());
  backendRequest.headers.set('X-Forwarded-For', request.headers.get('cf-connecting-ip') ?? '');

  const start = Date.now();
  const response = await fetch(backendRequest);
  const duration = Date.now() - start;

  // Add response headers
  const newResponse = new Response(response.body, response);
  newResponse.headers.set('X-Response-Time', `${duration}ms`);
  newResponse.headers.set('X-Backend', matchedPath);

  return newResponse;
}
```

## Pattern 7: Bot Detection

```typescript
function isLikelyBot(request: Request): boolean {
  const ua = request.headers.get('user-agent')?.toLowerCase() ?? '';

  // Known bot patterns
  const botPatterns = [
    'bot', 'crawl', 'spider', 'scrape', 'curl', 'wget',
    'python-requests', 'go-http-client', 'java/', 'httpclient',
  ];

  if (botPatterns.some(p => ua.includes(p))) return true;

  // No user agent
  if (!ua) return true;

  // Cloudflare Bot Management score (if available)
  const botScore = (request as any).cf?.botManagement?.score;
  if (botScore !== undefined && botScore < 30) return true;

  return false;
}

async function handleRequest(request: Request, env: Env): Promise<Response> {
  if (isLikelyBot(request)) {
    // Serve static/cached version (cheaper, no dynamic features)
    return fetch(`${env.STATIC_ORIGIN}${new URL(request.url).pathname}`);
  }

  // Serve full dynamic version for real users
  return fetch(request);
}
```

## Gotchas

- **KV is eventually consistent.** Rate limiting with KV works at edge scale but isn't perfectly accurate. For strict rate limiting, use Durable Objects or a Redis-backed system.
- **Edge functions can't connect to databases directly.** Use HTTP-based database drivers (Neon, PlanetScale, Turso, Supabase) or proxy through your origin server.
- **`crypto.subtle` is async.** All Web Crypto API operations return Promises. Don't forget `await`.
- **Response bodies can only be read once.** If you need to read and forward a response body, use `response.clone()` before reading.
- **Edge cache != CDN cache.** Your edge function can cache in KV, but the CDN cache is separate (controlled by `Cache-Control` headers). Both layers can cache independently.
- **Cold starts don't exist but initialization does.** V8 isolate creation is ~0ms, but your code still needs to parse and execute on first request. Keep bundle size small.
