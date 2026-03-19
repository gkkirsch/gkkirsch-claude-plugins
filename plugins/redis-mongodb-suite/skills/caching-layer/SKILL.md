---
name: caching-layer
description: >
  Add a complete caching layer to an existing application. Generates Redis-backed cache
  middleware, cache key strategies, invalidation hooks, and monitoring for Express, FastAPI,
  Django, or any web framework.
  Triggers: "add caching", "cache layer", "cache middleware", "speed up API",
  "Redis cache integration", "cache response", "API caching".
  NOT for: CDN configuration, browser caching headers, database query optimization.
version: 1.0.0
argument-hint: "[framework-or-endpoint]"
allowed-tools: Read, Grep, Glob, Bash, Write, Edit
model: sonnet
---

# Caching Layer

Add a production-ready caching layer to your application. Detects your framework, generates appropriate middleware/decorators, and includes cache key strategies with invalidation.

## Step 1: Analyze the Application

Before generating cache code, determine:
1. **Framework** — Express, FastAPI, Django, Flask, NestJS, Spring Boot
2. **Cache targets** — which endpoints/functions to cache
3. **Invalidation triggers** — which write operations should bust the cache
4. **Data characteristics** — how often does data change, read:write ratio

## Express.js Cache Middleware

```javascript
const redis = require("redis");

// Redis client singleton
const cacheClient = redis.createClient({
  url: process.env.REDIS_URL || "redis://localhost:6379",
  socket: {
    reconnectStrategy: (retries) => Math.min(retries * 100, 5000),
  },
});

cacheClient.on("error", (err) => console.error("Redis error:", err));
cacheClient.connect();

/**
 * Cache middleware factory.
 * @param {object} options
 * @param {number} options.ttl - Cache TTL in seconds (default: 300)
 * @param {function} options.keyFn - Custom key generator (req) => string
 * @param {string[]} options.varyBy - Headers to include in cache key
 * @param {string[]} options.tags - Cache tags for group invalidation
 */
function cacheMiddleware(options = {}) {
  const {
    ttl = 300,
    keyFn = defaultKeyFn,
    varyBy = [],
    tags = [],
  } = options;

  return async (req, res, next) => {
    // Skip cache for non-GET requests
    if (req.method !== "GET") return next();

    // Skip if cache-control: no-cache
    if (req.headers["cache-control"] === "no-cache") return next();

    const cacheKey = keyFn(req);

    try {
      const cached = await cacheClient.get(cacheKey);
      if (cached) {
        const { body, contentType, statusCode } = JSON.parse(cached);
        res.set("Content-Type", contentType);
        res.set("X-Cache", "HIT");
        return res.status(statusCode).send(body);
      }
    } catch (err) {
      console.error("Cache read error:", err);
      // Proceed without cache on error
    }

    // Intercept response to cache it
    const originalJson = res.json.bind(res);
    const originalSend = res.send.bind(res);

    const cacheResponse = async (body) => {
      if (res.statusCode >= 200 && res.statusCode < 300) {
        const cacheData = JSON.stringify({
          body: typeof body === "string" ? body : JSON.stringify(body),
          contentType: res.get("Content-Type") || "application/json",
          statusCode: res.statusCode,
        });

        try {
          const pipe = cacheClient.multi();
          pipe.setEx(cacheKey, ttl, cacheData);

          // Associate with tags for group invalidation
          for (const tag of tags) {
            pipe.sAdd(`cache:tag:${tag}`, cacheKey);
            pipe.expire(`cache:tag:${tag}`, ttl + 3600);
          }

          await pipe.exec();
        } catch (err) {
          console.error("Cache write error:", err);
        }
      }
    };

    res.json = (body) => {
      cacheResponse(body);
      res.set("X-Cache", "MISS");
      return originalJson(body);
    };

    res.send = (body) => {
      cacheResponse(body);
      res.set("X-Cache", "MISS");
      return originalSend(body);
    };

    next();
  };
}

function defaultKeyFn(req) {
  const url = req.originalUrl || req.url;
  return `cache:${req.method}:${url}`;
}

/**
 * Invalidate cache by tag.
 */
async function invalidateTag(tag) {
  const tagKey = `cache:tag:${tag}`;
  const keys = await cacheClient.sMembers(tagKey);
  if (keys.length > 0) {
    await cacheClient.del([...keys, tagKey]);
  }
}

/**
 * Invalidate specific cache keys by pattern.
 */
async function invalidatePattern(pattern) {
  let cursor = 0;
  do {
    const result = await cacheClient.scan(cursor, {
      MATCH: pattern,
      COUNT: 100,
    });
    cursor = result.cursor;
    if (result.keys.length > 0) {
      await cacheClient.del(result.keys);
    }
  } while (cursor !== 0);
}


// === Usage ===

const app = require("express")();

// Cache product listings for 5 minutes
app.get("/api/products",
  cacheMiddleware({
    ttl: 300,
    tags: ["products"],
  }),
  async (req, res) => {
    const products = await db.getProducts(req.query);
    res.json(products);
  }
);

// Cache individual product for 10 minutes
app.get("/api/products/:id",
  cacheMiddleware({
    ttl: 600,
    keyFn: (req) => `cache:product:${req.params.id}`,
    tags: ["products", `product:${req.params?.id}`],
  }),
  async (req, res) => {
    const product = await db.getProduct(req.params.id);
    res.json(product);
  }
);

// Write endpoint invalidates cache
app.put("/api/products/:id", async (req, res) => {
  await db.updateProduct(req.params.id, req.body);

  // Invalidate related caches
  await invalidateTag("products");  // All product listings
  await invalidateTag(`product:${req.params.id}`);  // Specific product

  res.json({ success: true });
});

module.exports = { cacheMiddleware, invalidateTag, invalidatePattern, cacheClient };
```

## FastAPI Cache Decorator (Python)

```python
import redis.asyncio as aioredis
import json
import hashlib
from functools import wraps
from fastapi import FastAPI, Request, Response
from typing import Optional, Callable

app = FastAPI()

# Async Redis client
redis_pool = aioredis.ConnectionPool.from_url(
    "redis://localhost:6379",
    max_connections=20,
    decode_responses=True,
)
cache_redis = aioredis.Redis(connection_pool=redis_pool)


def cached(
    ttl: int = 300,
    key_prefix: str = "",
    tags: list[str] = None,
    key_fn: Optional[Callable] = None,
):
    """Cache decorator for FastAPI route handlers."""

    def decorator(func):
        @wraps(func)
        async def wrapper(*args, **kwargs):
            # Extract request from args
            request = kwargs.get("request") or next(
                (a for a in args if isinstance(a, Request)), None
            )

            # Build cache key
            if key_fn:
                cache_key = f"cache:{key_prefix}:{key_fn(*args, **kwargs)}"
            elif request:
                url = str(request.url)
                cache_key = f"cache:{key_prefix}:{hashlib.md5(url.encode()).hexdigest()}"
            else:
                param_hash = hashlib.md5(
                    json.dumps(kwargs, default=str, sort_keys=True).encode()
                ).hexdigest()
                cache_key = f"cache:{key_prefix}:{param_hash}"

            # Check cache
            try:
                cached_value = await cache_redis.get(cache_key)
                if cached_value:
                    return json.loads(cached_value)
            except Exception as e:
                logger.warning(f"Cache read error: {e}")

            # Execute handler
            result = await func(*args, **kwargs)

            # Cache result
            try:
                pipe = cache_redis.pipeline()
                pipe.setex(cache_key, ttl, json.dumps(result, default=str))

                if tags:
                    for tag in tags:
                        pipe.sadd(f"cache:tag:{tag}", cache_key)
                        pipe.expire(f"cache:tag:{tag}", ttl + 3600)

                await pipe.execute()
            except Exception as e:
                logger.warning(f"Cache write error: {e}")

            return result

        return wrapper
    return decorator


async def invalidate_tags(*tags: str):
    """Invalidate all cache keys associated with given tags."""
    for tag in tags:
        tag_key = f"cache:tag:{tag}"
        keys = await cache_redis.smembers(tag_key)
        if keys:
            await cache_redis.delete(*keys, tag_key)


# === Usage ===

@app.get("/api/products")
@cached(ttl=300, key_prefix="products", tags=["products"])
async def list_products(request: Request, category: str = None, page: int = 1):
    return await db.get_products(category=category, page=page)


@app.get("/api/products/{product_id}")
@cached(
    ttl=600,
    key_prefix="product",
    key_fn=lambda product_id, **kw: product_id,
    tags=["products"],
)
async def get_product(product_id: str):
    return await db.get_product(product_id)


@app.put("/api/products/{product_id}")
async def update_product(product_id: str, data: dict):
    await db.update_product(product_id, data)
    await invalidate_tags("products")
    return {"success": True}
```

## Cache Key Design

### Key Naming Convention

```
Pattern: {service}:{entity}:{identifier}:{variant}

Examples:
  api:user:123                    # User by ID
  api:user:123:profile            # User profile subset
  api:products:list:cat=elec:p=1  # Paginated product list
  api:search:q=redis:p=1          # Search results
  api:config:feature_flags        # Application config
  agg:daily_stats:2024-01-15      # Aggregated data
  sess:abc123                     # Session
  lock:process_orders             # Distributed lock
```

### Key Generation Best Practices

```javascript
// Deterministic key from query parameters
function buildCacheKey(prefix, params) {
  // Sort keys for consistency
  const sorted = Object.keys(params)
    .filter((k) => params[k] !== undefined && params[k] !== null)
    .sort()
    .map((k) => `${k}=${params[k]}`)
    .join(":");

  return `${prefix}:${sorted}`;
}

// Usage
buildCacheKey("api:products", { category: "electronics", page: 1, sort: "price" });
// → "api:products:category=electronics:page=1:sort=price"
```

## Cache Monitoring Middleware

```javascript
// Express middleware to track cache hit/miss rates
const cacheMetrics = {
  hits: 0,
  misses: 0,
  errors: 0,

  get hitRate() {
    const total = this.hits + this.misses;
    return total > 0 ? ((this.hits / total) * 100).toFixed(1) : 0;
  },

  reset() {
    this.hits = 0;
    this.misses = 0;
    this.errors = 0;
  },
};

// Expose metrics endpoint
app.get("/api/cache/stats", (req, res) => {
  res.json({
    hit_rate: `${cacheMetrics.hitRate}%`,
    hits: cacheMetrics.hits,
    misses: cacheMetrics.misses,
    errors: cacheMetrics.errors,
  });
});

// Reset metrics every hour
setInterval(() => cacheMetrics.reset(), 3600 * 1000);
```

## Gotchas

- Always handle Redis connection errors gracefully — app should work without cache
- Set `X-Cache: HIT/MISS` header for debugging cache behavior
- Never cache responses with Set-Cookie headers
- Don't cache authenticated/personalized responses with shared keys
- Use `Vary` header correctly when caching responds differently by header
- Cache invalidation on writes is critical — stale data is worse than no cache
- Monitor cache memory usage — set `maxmemory` and `maxmemory-policy` in Redis
- Don't cache errors (5xx responses) — they should be retried, not served from cache
- Test cache behavior with `Cache-Control: no-cache` header to bypass
