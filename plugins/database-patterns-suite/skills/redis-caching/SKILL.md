---
name: redis-caching
description: >
  Redis caching patterns — cache-aside, write-through, write-behind, cache
  invalidation, TTL strategies, session storage, rate limiting, and pub/sub
  with Node.js (ioredis).
  Triggers: "redis", "caching", "cache pattern", "cache invalidation",
  "rate limiting redis", "session store redis", "pub sub", "redis pub/sub".
  NOT for: SQL queries (use sql-patterns), schema design (use database-architect).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Redis Caching Patterns

## Setup

```bash
# Install Redis client
npm install ioredis

# Install Redis server (local dev)
brew install redis
brew services start redis

# Verify connection
redis-cli ping  # Should return PONG
```

```typescript
// src/lib/redis.ts
import Redis from "ioredis";

const redis = new Redis({
  host: process.env.REDIS_HOST || "localhost",
  port: parseInt(process.env.REDIS_PORT || "6379"),
  password: process.env.REDIS_PASSWORD,
  maxRetriesPerRequest: 3,
  retryStrategy(times) {
    const delay = Math.min(times * 50, 2000);
    return delay;
  },
  // Connection pool settings
  lazyConnect: true,
  enableReadyCheck: true,
});

redis.on("error", (err) => console.error("Redis error:", err));
redis.on("connect", () => console.log("Redis connected"));

export default redis;
```

## Cache-Aside Pattern (Lazy Loading)

The most common caching pattern. Read from cache first, fall back to database.

```typescript
// Generic cache-aside wrapper
async function cacheAside<T>(
  key: string,
  fetchFn: () => Promise<T>,
  ttlSeconds: number = 300
): Promise<T> {
  // 1. Check cache
  const cached = await redis.get(key);
  if (cached) {
    return JSON.parse(cached) as T;
  }

  // 2. Cache miss — fetch from source
  const data = await fetchFn();

  // 3. Store in cache (don't await — fire and forget)
  redis.setex(key, ttlSeconds, JSON.stringify(data));

  return data;
}

// Usage
async function getUser(id: string) {
  return cacheAside(
    `user:${id}`,
    () => db.user.findUnique({ where: { id } }),
    600 // 10 minutes
  );
}

async function getProduct(slug: string) {
  return cacheAside(
    `product:${slug}`,
    () => db.product.findUnique({ where: { slug } }),
    3600 // 1 hour
  );
}
```

## Write-Through Pattern

Write to cache AND database simultaneously. Ensures cache is always up to date.

```typescript
async function writeThrough<T>(
  key: string,
  data: T,
  writeFn: (data: T) => Promise<T>,
  ttlSeconds: number = 300
): Promise<T> {
  // 1. Write to database
  const result = await writeFn(data);

  // 2. Update cache
  await redis.setex(key, ttlSeconds, JSON.stringify(result));

  return result;
}

// Usage
async function updateUser(id: string, updates: Partial<User>) {
  return writeThrough(
    `user:${id}`,
    updates,
    (data) => db.user.update({ where: { id }, data }),
    600
  );
}
```

## Cache Invalidation

```typescript
// Single key invalidation
async function invalidateUser(id: string) {
  await redis.del(`user:${id}`);
}

// Pattern-based invalidation
async function invalidateUserCache(userId: string) {
  const keys = await redis.keys(`user:${userId}:*`);
  if (keys.length > 0) {
    await redis.del(...keys);
  }
}

// Tag-based invalidation (using Sets)
async function cacheWithTags<T>(
  key: string,
  tags: string[],
  data: T,
  ttl: number
): Promise<void> {
  const pipeline = redis.pipeline();
  pipeline.setex(key, ttl, JSON.stringify(data));
  for (const tag of tags) {
    pipeline.sadd(`tag:${tag}`, key);
    pipeline.expire(`tag:${tag}`, ttl + 60);
  }
  await pipeline.exec();
}

async function invalidateByTag(tag: string): Promise<void> {
  const keys = await redis.smembers(`tag:${tag}`);
  if (keys.length > 0) {
    await redis.del(...keys, `tag:${tag}`);
  }
}

// Usage
await cacheWithTags(
  `product:${id}`,
  ["products", `category:${categoryId}`],
  product,
  3600
);

// Invalidate all products in a category
await invalidateByTag(`category:${categoryId}`);
```

## TTL Strategies

```typescript
// TTL by data type
const TTL = {
  // Frequently changing data
  session: 1800,           // 30 min
  rateLimitWindow: 60,     // 1 min
  realTimeMetrics: 10,     // 10 sec

  // Moderately changing data
  userProfile: 300,        // 5 min
  searchResults: 120,      // 2 min
  apiResponse: 60,         // 1 min

  // Rarely changing data
  productCatalog: 3600,    // 1 hour
  staticContent: 86400,    // 1 day
  configSettings: 900,     // 15 min
} as const;

// Jittered TTL to prevent thundering herd
function jitteredTTL(baseTTL: number): number {
  const jitter = Math.floor(baseTTL * 0.1 * Math.random());
  return baseTTL + jitter;
}

// Stale-while-revalidate pattern
async function staleWhileRevalidate<T>(
  key: string,
  fetchFn: () => Promise<T>,
  ttl: number,
  staleTTL: number
): Promise<T> {
  const cached = await redis.get(key);
  if (cached) {
    const { data, expiresAt } = JSON.parse(cached);

    if (Date.now() > expiresAt) {
      // Stale — serve stale data but refresh in background
      fetchFn().then((fresh) => {
        redis.setex(key, staleTTL, JSON.stringify({
          data: fresh,
          expiresAt: Date.now() + ttl * 1000,
        }));
      });
    }

    return data as T;
  }

  const data = await fetchFn();
  await redis.setex(key, staleTTL, JSON.stringify({
    data,
    expiresAt: Date.now() + ttl * 1000,
  }));
  return data;
}
```

## Session Storage

```typescript
// Express session with Redis
import session from "express-session";
import RedisStore from "connect-redis";

app.use(session({
  store: new RedisStore({ client: redis }),
  secret: process.env.SESSION_SECRET!,
  resave: false,
  saveUninitialized: false,
  cookie: {
    secure: process.env.NODE_ENV === "production",
    httpOnly: true,
    maxAge: 1800000, // 30 min
    sameSite: "lax",
  },
}));

// Custom session management
class SessionManager {
  private prefix = "sess:";

  async create(userId: string, metadata: object): Promise<string> {
    const sessionId = crypto.randomUUID();
    const session = {
      userId,
      ...metadata,
      createdAt: Date.now(),
    };

    await redis.setex(
      `${this.prefix}${sessionId}`,
      1800, // 30 min
      JSON.stringify(session)
    );

    // Track user's active sessions
    await redis.sadd(`user-sessions:${userId}`, sessionId);

    return sessionId;
  }

  async get(sessionId: string) {
    const data = await redis.get(`${this.prefix}${sessionId}`);
    return data ? JSON.parse(data) : null;
  }

  async refresh(sessionId: string): Promise<boolean> {
    return (await redis.expire(`${this.prefix}${sessionId}`, 1800)) === 1;
  }

  async destroy(sessionId: string, userId: string): Promise<void> {
    await redis.del(`${this.prefix}${sessionId}`);
    await redis.srem(`user-sessions:${userId}`, sessionId);
  }

  async destroyAllForUser(userId: string): Promise<void> {
    const sessions = await redis.smembers(`user-sessions:${userId}`);
    if (sessions.length > 0) {
      await redis.del(
        ...sessions.map((s) => `${this.prefix}${s}`),
        `user-sessions:${userId}`
      );
    }
  }
}
```

## Rate Limiting

```typescript
// Sliding window rate limiter
async function slidingWindowRateLimit(
  key: string,
  limit: number,
  windowSeconds: number
): Promise<{ allowed: boolean; remaining: number; resetAt: number }> {
  const now = Date.now();
  const windowStart = now - windowSeconds * 1000;
  const member = `${now}:${Math.random()}`;

  const pipeline = redis.pipeline();
  pipeline.zremrangebyscore(key, 0, windowStart);  // Remove old entries
  pipeline.zadd(key, now, member);                  // Add current request
  pipeline.zcard(key);                              // Count requests in window
  pipeline.expire(key, windowSeconds);              // Auto-cleanup

  const results = await pipeline.exec();
  const count = results![2][1] as number;

  return {
    allowed: count <= limit,
    remaining: Math.max(0, limit - count),
    resetAt: Math.ceil((windowStart + windowSeconds * 1000) / 1000),
  };
}

// Token bucket rate limiter
async function tokenBucket(
  key: string,
  maxTokens: number,
  refillRate: number, // tokens per second
  tokensRequired: number = 1
): Promise<{ allowed: boolean; tokens: number }> {
  const now = Date.now();
  const script = `
    local key = KEYS[1]
    local max_tokens = tonumber(ARGV[1])
    local refill_rate = tonumber(ARGV[2])
    local now = tonumber(ARGV[3])
    local requested = tonumber(ARGV[4])

    local bucket = redis.call('HMGET', key, 'tokens', 'last_refill')
    local tokens = tonumber(bucket[1]) or max_tokens
    local last_refill = tonumber(bucket[2]) or now

    local elapsed = (now - last_refill) / 1000
    tokens = math.min(max_tokens, tokens + elapsed * refill_rate)

    local allowed = 0
    if tokens >= requested then
      tokens = tokens - requested
      allowed = 1
    end

    redis.call('HMSET', key, 'tokens', tokens, 'last_refill', now)
    redis.call('EXPIRE', key, math.ceil(max_tokens / refill_rate) + 1)

    return {allowed, math.floor(tokens)}
  `;

  const [allowed, tokens] = (await redis.eval(
    script, 1, key, maxTokens, refillRate, now, tokensRequired
  )) as [number, number];

  return { allowed: allowed === 1, tokens };
}

// Express middleware
function rateLimitMiddleware(limit: number, windowSeconds: number) {
  return async (req: Request, res: Response, next: NextFunction) => {
    const key = `ratelimit:${req.ip}:${req.path}`;
    const result = await slidingWindowRateLimit(key, limit, windowSeconds);

    res.set({
      "X-RateLimit-Limit": String(limit),
      "X-RateLimit-Remaining": String(result.remaining),
      "X-RateLimit-Reset": String(result.resetAt),
    });

    if (!result.allowed) {
      return res.status(429).json({
        error: "Too many requests",
        retryAfter: result.resetAt - Math.ceil(Date.now() / 1000),
      });
    }

    next();
  };
}
```

## Pub/Sub

```typescript
// Publisher
const pub = new Redis();

async function publishEvent(channel: string, event: object) {
  await pub.publish(channel, JSON.stringify(event));
}

// Subscriber
const sub = new Redis();

sub.subscribe("orders", "notifications", (err, count) => {
  if (err) throw err;
  console.log(`Subscribed to ${count} channels`);
});

sub.on("message", (channel, message) => {
  const event = JSON.parse(message);

  switch (channel) {
    case "orders":
      handleOrderEvent(event);
      break;
    case "notifications":
      handleNotification(event);
      break;
  }
});

// Pattern subscribe (wildcard channels)
sub.psubscribe("events:*", (err) => {
  if (err) throw err;
});

sub.on("pmessage", (pattern, channel, message) => {
  // channel = "events:user:signup", "events:order:placed", etc.
  console.log(`${channel}: ${message}`);
});
```

## Redis Data Structure Patterns

```typescript
// Sorted Set — Leaderboard
async function updateScore(userId: string, score: number) {
  await redis.zadd("leaderboard", score, userId);
}

async function getTopPlayers(count: number = 10) {
  return redis.zrevrange("leaderboard", 0, count - 1, "WITHSCORES");
}

async function getRank(userId: string): Promise<number | null> {
  const rank = await redis.zrevrank("leaderboard", userId);
  return rank !== null ? rank + 1 : null;
}

// Hash — User profiles (partial updates)
async function setUserFields(userId: string, fields: Record<string, string>) {
  await redis.hmset(`user:${userId}`, fields);
  await redis.expire(`user:${userId}`, 3600);
}

async function getUserField(userId: string, field: string) {
  return redis.hget(`user:${userId}`, field);
}

// List — Job queue (simple)
async function enqueue(queue: string, job: object) {
  await redis.lpush(queue, JSON.stringify(job));
}

async function dequeue(queue: string): Promise<object | null> {
  const job = await redis.rpop(queue);
  return job ? JSON.parse(job) : null;
}

// Blocking dequeue (worker pattern)
async function blockingDequeue(queue: string, timeoutSec: number = 0) {
  const result = await redis.brpop(queue, timeoutSec);
  return result ? JSON.parse(result[1]) : null;
}

// Set — Unique visitors tracking
async function trackVisitor(page: string, visitorId: string) {
  await redis.sadd(`visitors:${page}:${today()}`, visitorId);
}

async function uniqueVisitorCount(page: string): Promise<number> {
  return redis.scard(`visitors:${page}:${today()}`);
}

// HyperLogLog — Approximate unique counts (memory efficient)
async function trackPageView(page: string, visitorId: string) {
  await redis.pfadd(`pageviews:${page}`, visitorId);
}

async function approximateUniqueViews(page: string): Promise<number> {
  return redis.pfcount(`pageviews:${page}`);
}
```

## Gotchas

1. **`redis.keys()` is O(N) and blocks Redis.** Never use in production on large databases. Use `SCAN` instead for iterating keys: `for await (const key of redis.scanStream({ match: 'user:*' })) { ... }`.

2. **Pub/Sub subscribers can't run other commands.** Once a Redis client subscribes, it can ONLY receive messages. Create a separate Redis instance for pub/sub. That's why the examples above use `new Redis()` twice.

3. **Cache stampede / thundering herd.** When a popular cache key expires, hundreds of requests hit the database simultaneously. Fix: use jittered TTLs, stale-while-revalidate, or a mutex lock (`SET key:lock NX EX 10`) so only one request refreshes.

4. **JSON.stringify/parse is slow for large objects.** For frequently accessed large data, consider MessagePack (`msgpackr`) or Protocol Buffers instead of JSON serialization.

5. **Pipeline vs individual commands.** Individual `redis.get()` calls have network round-trip overhead. Use `redis.pipeline()` to batch multiple commands into a single round-trip. 10 pipelined commands are faster than 10 sequential ones.

6. **Redis memory can grow unbounded.** Always set TTLs on cache keys. Configure `maxmemory` and `maxmemory-policy` (usually `allkeys-lru`) in redis.conf. Without a policy, Redis returns errors when memory is full.
