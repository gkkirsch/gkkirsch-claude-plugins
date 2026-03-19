---
name: caching-patterns
description: >
  Implement multi-layer caching strategies for web applications. Covers HTTP caching
  (Cache-Control, ETags), CDN configuration, application-level caches, Redis caching,
  and cache invalidation patterns.
  Triggers: "caching strategy", "cache invalidation", "Cache-Control headers", "CDN caching",
  "cache layer", "improve response time", "reduce database load".
  NOT for: Redis data structures (use redis-implementation), database query optimization.
version: 1.0.0
argument-hint: "[what to cache or optimize]"
allowed-tools: Read, Grep, Glob, Bash, Write, Edit
model: sonnet
---

# Caching Patterns

Implement effective caching strategies across all layers of a web application — from HTTP headers to CDN to application to database.

## Cache Strategy Selection

### Decision Tree

```
What are you caching?
  │
  ├─ Static assets (JS, CSS, images, fonts)
  │   └─ CDN + immutable Cache-Control + content hash filenames
  │       Cache-Control: public, max-age=31536000, immutable
  │
  ├─ API responses (JSON)
  │   ├─ Same for all users (public data)?
  │   │   └─ CDN + stale-while-revalidate
  │   │       Cache-Control: public, max-age=60, stale-while-revalidate=300
  │   ├─ Per-user but changes rarely?
  │   │   └─ Redis + ETag validation
  │   │       Cache-Control: private, max-age=0, must-revalidate
  │   └─ Per-user and changes often?
  │       └─ In-process cache or no cache
  │
  ├─ Database query results
  │   └─ Redis or in-process cache + event-driven invalidation
  │
  ├─ Computed/aggregated data
  │   └─ Redis with TTL + background refresh
  │
  └─ Session data
      └─ Redis with sliding TTL
```

## HTTP Caching

### Cache-Control Headers

```
# Directive reference
public          — CDN and browser can cache
private         — Browser only (user-specific content)
no-cache        — Must validate with server before using cached copy
no-store        — Never cache (sensitive data: banking, health)
max-age=N       — Fresh for N seconds
s-maxage=N      — CDN-specific max-age (overrides max-age for CDNs)
immutable       — Never changes (use with content-hashed filenames)
must-revalidate — After max-age expires, MUST revalidate (no stale serving)
stale-while-revalidate=N  — Serve stale while refreshing in background for N seconds
stale-if-error=N          — Serve stale if origin returns error for N seconds
```

### Common Recipes

```javascript
// Express.js cache middleware
function setCacheHeaders(options) {
  return (req, res, next) => {
    const { scope, maxAge, swr, sie, immutable, noStore } = options;

    if (noStore) {
      res.set('Cache-Control', 'no-store');
      return next();
    }

    const directives = [];
    directives.push(scope || 'public');
    if (maxAge !== undefined) directives.push(`max-age=${maxAge}`);
    if (swr !== undefined) directives.push(`stale-while-revalidate=${swr}`);
    if (sie !== undefined) directives.push(`stale-if-error=${sie}`);
    if (immutable) directives.push('immutable');

    res.set('Cache-Control', directives.join(', '));
    next();
  };
}

// Static assets: cache forever (filename has content hash)
app.use('/assets', setCacheHeaders({
  scope: 'public',
  maxAge: 31536000,    // 1 year
  immutable: true,
}), express.static('dist/assets'));

// HTML pages: always revalidate
app.get('/', setCacheHeaders({
  scope: 'public',
  maxAge: 0,
  swr: 60,  // Serve stale for 60s while refreshing
}));

// Public API: short cache + background refresh
app.get('/api/products', setCacheHeaders({
  scope: 'public',
  maxAge: 60,          // Fresh for 1 minute
  swr: 300,            // Serve stale up to 5 minutes while refreshing
  sie: 86400,          // Serve stale up to 1 day if origin errors
}));

// User-specific API: no CDN caching
app.get('/api/me', setCacheHeaders({
  scope: 'private',
  maxAge: 0,
}));

// Sensitive data: never cache
app.get('/api/billing', setCacheHeaders({ noStore: true }));
```

### ETag Validation

```javascript
import crypto from 'crypto';

function etagMiddleware() {
  return (req, res, next) => {
    const originalJson = res.json.bind(res);

    res.json = (body) => {
      // Generate ETag from response body
      const etag = `"${crypto.createHash('md5').update(JSON.stringify(body)).digest('hex')}"`;
      res.set('ETag', etag);

      // Check if client's cached version matches
      const ifNoneMatch = req.headers['if-none-match'];
      if (ifNoneMatch === etag) {
        return res.status(304).end();  // Not Modified
      }

      return originalJson(body);
    };

    next();
  };
}

app.use(etagMiddleware());
```

## Application-Level Caching

### In-Process Cache (Node.js)

```typescript
import { LRUCache } from 'lru-cache';

// Memory-bounded LRU cache
const cache = new LRUCache<string, any>({
  max: 500,              // Max entries
  maxSize: 50_000_000,   // 50MB max total size
  sizeCalculation: (value) => JSON.stringify(value).length,
  ttl: 5 * 60 * 1000,   // 5 minute TTL
  updateAgeOnGet: true,  // Refresh TTL on read (sliding window)
  allowStale: true,      // Return stale while fetching fresh
  fetchMethod: async (key) => {
    // Auto-fetch on miss
    return await fetchFromDatabase(key);
  },
});

// Usage
const user = await cache.fetch(`user:${userId}`);

// Manual set
cache.set(`product:${productId}`, productData);

// Delete (invalidation)
cache.delete(`user:${userId}`);

// Stats
console.log({
  size: cache.size,
  hits: cache.hits,
  misses: cache.misses,
  hitRate: cache.hits / (cache.hits + cache.misses),
});
```

### Multi-Level Cache (L1 + L2)

```typescript
import { LRUCache } from 'lru-cache';
import Redis from 'ioredis';

const l1 = new LRUCache<string, any>({ max: 1000, ttl: 60_000 });  // 1 min
const l2 = new Redis();

class MultiLevelCache {
  async get<T>(key: string): Promise<T | null> {
    // L1: In-process (fastest)
    const l1Result = l1.get(key);
    if (l1Result !== undefined) return l1Result as T;

    // L2: Redis (shared across instances)
    const l2Result = await l2.get(key);
    if (l2Result) {
      const parsed = JSON.parse(l2Result) as T;
      l1.set(key, parsed);  // Backfill L1
      return parsed;
    }

    return null;
  }

  async set(key: string, value: any, ttlSeconds: number = 3600): Promise<void> {
    const serialized = JSON.stringify(value);
    l1.set(key, value);
    await l2.setex(key, ttlSeconds, serialized);
  }

  async invalidate(key: string): Promise<void> {
    l1.delete(key);
    await l2.del(key);

    // Notify other instances to clear L1
    await l2.publish('cache:invalidate', JSON.stringify({ key }));
  }

  // Call this on startup for each instance
  startInvalidationListener(): void {
    const sub = l2.duplicate();
    sub.subscribe('cache:invalidate');
    sub.on('message', (_channel: string, message: string) => {
      const { key } = JSON.parse(message);
      l1.delete(key);
    });
  }
}

const cache = new MultiLevelCache();
cache.startInvalidationListener();
```

## CDN Caching Patterns

### Cloudflare Configuration

```javascript
// Cloudflare Workers — custom cache logic at the edge
export default {
  async fetch(request, env) {
    const url = new URL(request.url);
    const cacheKey = new Request(url.toString(), request);
    const cache = caches.default;

    // Check cache first
    let response = await cache.match(cacheKey);
    if (response) {
      return response;
    }

    // Fetch from origin
    response = await fetch(request);

    // Clone response for caching
    response = new Response(response.body, response);

    // Set cache headers based on path
    if (url.pathname.startsWith('/api/products')) {
      response.headers.set('Cache-Control', 'public, s-maxage=60, stale-while-revalidate=300');
    } else if (url.pathname.startsWith('/api/')) {
      response.headers.set('Cache-Control', 'private, no-cache');
    }

    // Store in edge cache
    if (response.ok) {
      await cache.put(cacheKey, response.clone());
    }

    return response;
  },
};
```

### Cache Key Strategies

```
Default: URL (path + query string)
  /api/products?page=1&sort=price → unique cache entry

Problems with query string caching:
  /api/products?page=1&sort=price
  /api/products?sort=price&page=1  ← Different cache key! Same content.

Solutions:
  1. Normalize query parameters (sort alphabetically)
  2. Use only specific parameters in cache key
  3. Use Vary header for content negotiation

// Vary header — separate cache per value of this header
app.get('/api/data', (req, res) => {
  res.set('Vary', 'Accept-Language, Accept-Encoding');
  // Separate cache entries for each language + encoding combo
});
```

## Cache Warming

### Background Refresh Pattern

```typescript
import cron from 'node-cron';

// Warm cache before it expires
async function warmCache() {
  const popularProducts = await getPopularProductIds();  // Top 100

  for (const productId of popularProducts) {
    const product = await fetchProductFromDB(productId);
    await cache.set(`product:${productId}`, product, 3600);
  }

  console.log(`Warmed cache with ${popularProducts.length} products`);
}

// Run every 50 minutes (cache TTL is 60 minutes)
cron.schedule('*/50 * * * *', warmCache);

// Also warm on startup
warmCache().catch(console.error);
```

### Stale-While-Revalidate in Application Code

```typescript
async function getWithSWR<T>(
  key: string,
  fetchFn: () => Promise<T>,
  options: { ttl: number; swrWindow: number } = { ttl: 300, swrWindow: 60 }
): Promise<T> {
  const cached = await redis.get(key);

  if (cached) {
    const { data, cachedAt } = JSON.parse(cached);
    const age = (Date.now() - cachedAt) / 1000;

    if (age < options.ttl) {
      return data;  // Fresh — return immediately
    }

    if (age < options.ttl + options.swrWindow) {
      // Stale but within SWR window — return stale, refresh in background
      refreshInBackground(key, fetchFn, options.ttl);
      return data;
    }
  }

  // Miss or too stale — fetch synchronously
  const data = await fetchFn();
  await redis.set(key, JSON.stringify({ data, cachedAt: Date.now() }));
  await redis.expire(key, options.ttl + options.swrWindow + 60);  // Extra buffer
  return data;
}

function refreshInBackground(key: string, fetchFn: () => Promise<any>, ttl: number) {
  // Don't await — fire and forget
  fetchFn()
    .then(data => redis.set(key, JSON.stringify({ data, cachedAt: Date.now() })))
    .then(() => redis.expire(key, ttl + 120))
    .catch(err => console.error(`Background refresh failed for ${key}:`, err));
}
```

## Cache Invalidation Patterns

### Tag-Based Invalidation

```typescript
// Tag cache entries for group invalidation
class TaggedCache {
  private redis: Redis;

  async set(key: string, value: any, tags: string[], ttl: number = 3600) {
    const pipe = this.redis.pipeline();

    // Store the value
    pipe.setex(key, ttl, JSON.stringify(value));

    // Associate key with each tag
    for (const tag of tags) {
      pipe.sadd(`tag:${tag}`, key);
      pipe.expire(`tag:${tag}`, ttl + 60);
    }

    await pipe.exec();
  }

  async invalidateByTag(tag: string) {
    const keys = await this.redis.smembers(`tag:${tag}`);
    if (keys.length > 0) {
      await this.redis.del(...keys, `tag:${tag}`);
    }
  }
}

// Usage
const taggedCache = new TaggedCache(redis);

// Cache a product listing tagged by category
await taggedCache.set(
  'products:electronics:page1',
  productData,
  ['category:electronics', 'products:list'],
  300
);

// When a product in electronics changes:
await taggedCache.invalidateByTag('category:electronics');
// Clears ALL cache entries tagged with this category
```

### Write-Behind (Async Write) Pattern

```typescript
// Batch writes to database, serve from cache immediately
class WriteBehindCache {
  private writeQueue: Map<string, any> = new Map();
  private flushInterval: NodeJS.Timeout;

  constructor(private redis: Redis, flushMs: number = 5000) {
    this.flushInterval = setInterval(() => this.flush(), flushMs);
  }

  async update(key: string, value: any) {
    // Update cache immediately (users see instant update)
    await this.redis.setex(key, 3600, JSON.stringify(value));

    // Queue database write
    this.writeQueue.set(key, value);
  }

  private async flush() {
    if (this.writeQueue.size === 0) return;

    const batch = new Map(this.writeQueue);
    this.writeQueue.clear();

    try {
      await batchWriteToDatabase(batch);
    } catch (err) {
      // Re-queue failed writes
      for (const [key, value] of batch) {
        this.writeQueue.set(key, value);
      }
      console.error('Write-behind flush failed:', err);
    }
  }

  async shutdown() {
    clearInterval(this.flushInterval);
    await this.flush();  // Flush remaining writes
  }
}
```

## Monitoring and Debugging

### Cache Health Dashboard Metrics

```typescript
// Track cache performance
class CacheMetrics {
  private hits = 0;
  private misses = 0;
  private errors = 0;
  private latencies: number[] = [];

  recordHit(latencyMs: number) { this.hits++; this.latencies.push(latencyMs); }
  recordMiss(latencyMs: number) { this.misses++; this.latencies.push(latencyMs); }
  recordError() { this.errors++; }

  getStats() {
    const total = this.hits + this.misses;
    const sorted = [...this.latencies].sort((a, b) => a - b);
    return {
      hitRate: total > 0 ? this.hits / total : 0,
      missRate: total > 0 ? this.misses / total : 0,
      errorRate: total > 0 ? this.errors / total : 0,
      p50: sorted[Math.floor(sorted.length * 0.5)] || 0,
      p95: sorted[Math.floor(sorted.length * 0.95)] || 0,
      p99: sorted[Math.floor(sorted.length * 0.99)] || 0,
      total,
    };
  }

  reset() {
    this.hits = this.misses = this.errors = 0;
    this.latencies = [];
  }
}
```

## Checklist Before Completing

- [ ] Cache-Control headers set appropriately for each route
- [ ] Static assets use content-hash filenames + immutable
- [ ] API responses use appropriate scope (public vs private)
- [ ] Cache stampede prevention implemented (lock or SWR)
- [ ] Invalidation strategy defined for each cached entity
- [ ] Cache key naming is consistent and collision-free
- [ ] TTL values documented with rationale
- [ ] Cache monitoring tracks hit rate, latency, and size
- [ ] Graceful degradation when cache is unavailable
- [ ] No sensitive data in public caches
