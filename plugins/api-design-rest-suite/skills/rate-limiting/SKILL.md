---
name: rate-limiting
description: >
  Implement API rate limiting — sliding window, token bucket, fixed window,
  Redis-backed distributed limiting, and per-user quotas.
  Triggers: "rate limiting", "API throttling", "request limits", "429 too many requests".
  NOT for: general caching, API authentication.
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# API Rate Limiting

## Quick Start: Express Rate Limit

For single-server deployments:

```bash
npm install express-rate-limit
```

```typescript
import rateLimit from 'express-rate-limit';

// Global rate limit: 100 requests per 15 minutes per IP
const globalLimiter = rateLimit({
  windowMs: 15 * 60 * 1000, // 15 minutes
  max: 100,
  standardHeaders: 'draft-7', // RateLimit-* headers (IETF draft)
  legacyHeaders: false,       // Disable X-RateLimit-* headers
  message: {
    error: 'TOO_MANY_REQUESTS',
    message: 'Rate limit exceeded. Try again later.',
    retryAfter: 900, // seconds
  },
});

app.use(globalLimiter);
```

### Per-Route Limits

```typescript
// Strict limit for auth endpoints (prevent brute force)
const authLimiter = rateLimit({
  windowMs: 15 * 60 * 1000,
  max: 5,  // 5 attempts per 15 minutes
  message: {
    error: 'TOO_MANY_REQUESTS',
    message: 'Too many login attempts. Try again in 15 minutes.',
  },
});

// Generous limit for read endpoints
const readLimiter = rateLimit({
  windowMs: 60 * 1000,
  max: 200,  // 200 requests per minute
});

// Strict limit for write endpoints
const writeLimiter = rateLimit({
  windowMs: 60 * 1000,
  max: 30,  // 30 writes per minute
});

app.use('/auth/login', authLimiter);
app.use('/auth/signup', authLimiter);
app.get('/api/*', readLimiter);
app.post('/api/*', writeLimiter);
app.put('/api/*', writeLimiter);
app.patch('/api/*', writeLimiter);
app.delete('/api/*', writeLimiter);
```

### Per-User Limits (Authenticated)

```typescript
const userLimiter = rateLimit({
  windowMs: 60 * 1000,
  max: 100,
  keyGenerator: (req) => {
    // Rate limit by user ID (authenticated) or IP (anonymous)
    return req.user?.id ?? req.ip;
  },
  skip: (req) => {
    // Skip rate limiting for admin users
    return req.user?.role === 'admin';
  },
});
```

## Redis-Backed Rate Limiting (Distributed)

For multi-server deployments, use Redis as the backing store:

```bash
npm install express-rate-limit rate-limit-redis ioredis
```

```typescript
import rateLimit from 'express-rate-limit';
import RedisStore from 'rate-limit-redis';
import Redis from 'ioredis';

const redis = new Redis(process.env.REDIS_URL);

const limiter = rateLimit({
  windowMs: 60 * 1000,
  max: 100,
  standardHeaders: 'draft-7',
  store: new RedisStore({
    sendCommand: (...args: string[]) => redis.call(...args),
    prefix: 'rl:',  // Redis key prefix
  }),
});

app.use('/api', limiter);
```

## Sliding Window Algorithm (Custom)

The sliding window algorithm provides smoother rate limiting than fixed windows:

```typescript
import Redis from 'ioredis';

const redis = new Redis(process.env.REDIS_URL);

interface RateLimitResult {
  allowed: boolean;
  remaining: number;
  resetAt: Date;
  retryAfter?: number;
}

async function slidingWindowRateLimit(
  key: string,
  maxRequests: number,
  windowMs: number
): Promise<RateLimitResult> {
  const now = Date.now();
  const windowStart = now - windowMs;
  const redisKey = `ratelimit:${key}`;

  // Use a Redis sorted set: score = timestamp, member = unique request ID
  const pipeline = redis.pipeline();
  pipeline.zremrangebyscore(redisKey, 0, windowStart); // Remove expired
  pipeline.zadd(redisKey, now, `${now}:${Math.random()}`); // Add current
  pipeline.zcard(redisKey); // Count in window
  pipeline.expire(redisKey, Math.ceil(windowMs / 1000)); // Set TTL

  const results = await pipeline.exec();
  const count = results![2][1] as number;

  if (count > maxRequests) {
    // Over limit — find when the oldest request in the window expires
    const oldestInWindow = await redis.zrange(redisKey, 0, 0, 'WITHSCORES');
    const oldestTimestamp = parseInt(oldestInWindow[1]);
    const retryAfter = Math.ceil((oldestTimestamp + windowMs - now) / 1000);

    return {
      allowed: false,
      remaining: 0,
      resetAt: new Date(oldestTimestamp + windowMs),
      retryAfter,
    };
  }

  return {
    allowed: true,
    remaining: maxRequests - count,
    resetAt: new Date(now + windowMs),
  };
}
```

### Use as Express Middleware

```typescript
function createRateLimiter(maxRequests: number, windowMs: number) {
  return async (req: Request, res: Response, next: NextFunction) => {
    const key = req.user?.id ?? req.ip;
    const result = await slidingWindowRateLimit(key, maxRequests, windowMs);

    // Always set rate limit headers
    res.set('RateLimit-Limit', String(maxRequests));
    res.set('RateLimit-Remaining', String(result.remaining));
    res.set('RateLimit-Reset', String(Math.ceil(result.resetAt.getTime() / 1000)));

    if (!result.allowed) {
      res.set('Retry-After', String(result.retryAfter));
      return res.status(429).json({
        error: 'TOO_MANY_REQUESTS',
        message: `Rate limit exceeded. Try again in ${result.retryAfter} seconds.`,
        retryAfter: result.retryAfter,
      });
    }

    next();
  };
}

// Usage
app.use('/api', createRateLimiter(100, 60_000)); // 100 req/min
app.use('/auth', createRateLimiter(5, 900_000));  // 5 req/15min
```

## Token Bucket Algorithm

Token bucket allows bursts while maintaining an average rate:

```typescript
async function tokenBucket(
  key: string,
  maxTokens: number,      // bucket capacity
  refillRate: number,      // tokens added per second
  tokensRequired: number = 1
): Promise<{ allowed: boolean; remaining: number; retryAfter?: number }> {
  const redisKey = `bucket:${key}`;
  const now = Date.now();

  // Lua script for atomic operation
  const luaScript = `
    local key = KEYS[1]
    local maxTokens = tonumber(ARGV[1])
    local refillRate = tonumber(ARGV[2])
    local now = tonumber(ARGV[3])
    local requested = tonumber(ARGV[4])

    local data = redis.call('HMGET', key, 'tokens', 'lastRefill')
    local tokens = tonumber(data[1]) or maxTokens
    local lastRefill = tonumber(data[2]) or now

    -- Refill tokens based on elapsed time
    local elapsed = (now - lastRefill) / 1000
    tokens = math.min(maxTokens, tokens + (elapsed * refillRate))

    if tokens >= requested then
      tokens = tokens - requested
      redis.call('HMSET', key, 'tokens', tokens, 'lastRefill', now)
      redis.call('EXPIRE', key, math.ceil(maxTokens / refillRate) + 1)
      return {1, math.floor(tokens)}
    else
      redis.call('HMSET', key, 'tokens', tokens, 'lastRefill', now)
      redis.call('EXPIRE', key, math.ceil(maxTokens / refillRate) + 1)
      local waitTime = math.ceil((requested - tokens) / refillRate)
      return {0, math.floor(tokens), waitTime}
    end
  `;

  const result = await redis.eval(
    luaScript, 1, redisKey,
    maxTokens, refillRate, now, tokensRequired
  ) as number[];

  return {
    allowed: result[0] === 1,
    remaining: result[1],
    retryAfter: result[2],
  };
}
```

## Tiered Rate Limits (By Plan)

```typescript
const PLAN_LIMITS: Record<string, { rpm: number; daily: number }> = {
  free:       { rpm: 10,   daily: 1000 },
  starter:    { rpm: 60,   daily: 10000 },
  pro:        { rpm: 300,  daily: 100000 },
  enterprise: { rpm: 1000, daily: 1000000 },
};

async function tieredRateLimit(req: Request, res: Response, next: NextFunction) {
  const plan = req.user?.plan ?? 'free';
  const limits = PLAN_LIMITS[plan];
  const userId = req.user?.id ?? req.ip;

  // Check per-minute limit
  const minuteResult = await slidingWindowRateLimit(
    `${userId}:rpm`, limits.rpm, 60_000
  );

  // Check daily limit
  const dailyResult = await slidingWindowRateLimit(
    `${userId}:daily`, limits.daily, 86_400_000
  );

  // Set headers (show the most restrictive limit)
  res.set('RateLimit-Limit', String(limits.rpm));
  res.set('RateLimit-Remaining', String(Math.min(minuteResult.remaining, dailyResult.remaining)));
  res.set('X-RateLimit-Daily-Limit', String(limits.daily));
  res.set('X-RateLimit-Daily-Remaining', String(dailyResult.remaining));

  if (!minuteResult.allowed) {
    res.set('Retry-After', String(minuteResult.retryAfter));
    return res.status(429).json({
      error: 'RATE_LIMIT_EXCEEDED',
      message: `Rate limit exceeded (${limits.rpm}/min for ${plan} plan).`,
      upgrade: plan !== 'enterprise' ? 'https://example.com/pricing' : undefined,
    });
  }

  if (!dailyResult.allowed) {
    return res.status(429).json({
      error: 'DAILY_LIMIT_EXCEEDED',
      message: `Daily limit exceeded (${limits.daily}/day for ${plan} plan).`,
      resetsAt: dailyResult.resetAt.toISOString(),
    });
  }

  next();
}
```

## Response Headers (IETF Standard)

Always return rate limit info in response headers:

```
RateLimit-Limit: 100           → Maximum requests in window
RateLimit-Remaining: 42        → Requests remaining
RateLimit-Reset: 1711065600    → Unix timestamp when window resets
Retry-After: 30                → Seconds to wait (only on 429)
```

## Gotchas

- **Always use `standardHeaders: 'draft-7'`** with express-rate-limit. The legacy `X-RateLimit-*` headers are non-standard.
- **IP-based limiting behind proxies**: Set `app.set('trust proxy', 1)` for Express behind nginx/load balancers. Otherwise `req.ip` is always the proxy IP.
- **Redis is required for multi-server.** In-memory rate limiting (default express-rate-limit store) resets on restart and doesn't share state across servers.
- **Separate limits for reads vs writes.** Reads are cheap; writes are expensive. `GET /users` should have a higher limit than `POST /users`.
- **Auth endpoints need the strictest limits.** Login, signup, password reset — these are brute force targets. 5-10 attempts per 15 minutes is reasonable.
- **Rate limiting by API key, not just IP.** If your API uses API keys, rate limit by key. Multiple users behind the same corporate IP shouldn't share a limit.
- **Return `Retry-After` on 429.** Clients need to know how long to wait. Without it, they'll retry immediately and stay rate-limited.
- **Don't rate limit health checks.** `/health` and `/ready` endpoints should be excluded from all rate limits.
