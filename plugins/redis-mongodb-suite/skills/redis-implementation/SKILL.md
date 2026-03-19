---
name: redis-implementation
description: >
  Implement Redis data structures, caching layers, session management, rate limiting,
  pub/sub messaging, and Redis Streams. Production-ready patterns with connection pooling,
  error handling, and monitoring.
  Triggers: "implement Redis", "Redis cache", "Redis session", "rate limiter",
  "Redis pub/sub", "Redis Streams", "distributed lock", "Redis queue".
  NOT for: Redis cluster ops (use redis-architect agent), MongoDB schemas (use mongodb-schema-design).
version: 1.0.0
argument-hint: "[what to implement with Redis]"
allowed-tools: Read, Grep, Glob, Bash, Write, Edit
model: sonnet
---

# Redis Implementation

Implement production-ready Redis solutions with proper connection management, error handling, serialization, and monitoring.

## Connection Management

### Node.js (ioredis)

```typescript
import Redis from 'ioredis';

// Single instance connection
const redis = new Redis({
  host: process.env.REDIS_HOST || 'localhost',
  port: parseInt(process.env.REDIS_PORT || '6379'),
  password: process.env.REDIS_PASSWORD,
  db: 0,
  maxRetriesPerRequest: 3,
  retryStrategy(times) {
    const delay = Math.min(times * 100, 3000);
    return delay;
  },
  reconnectOnError(err) {
    return err.message.includes('READONLY');
  },
  lazyConnect: true,  // Don't connect until first command
});

// Connection events
redis.on('connect', () => console.log('Redis connected'));
redis.on('error', (err) => console.error('Redis error:', err));
redis.on('close', () => console.warn('Redis connection closed'));

// Cluster connection
const cluster = new Redis.Cluster([
  { host: 'redis-node-1', port: 6379 },
  { host: 'redis-node-2', port: 6379 },
  { host: 'redis-node-3', port: 6379 },
], {
  redisOptions: { password: process.env.REDIS_PASSWORD },
  scaleReads: 'slave',  // Read from replicas
  maxRedirections: 16,
});

// Graceful shutdown
async function shutdown() {
  await redis.quit();  // Wait for pending commands
  process.exit(0);
}
process.on('SIGTERM', shutdown);
process.on('SIGINT', shutdown);
```

### Python (redis-py)

```python
import redis
from redis.backoff import ExponentialBackoff
from redis.retry import Retry

# Connection pool (recommended for multi-threaded apps)
pool = redis.ConnectionPool(
    host=os.environ.get('REDIS_HOST', 'localhost'),
    port=int(os.environ.get('REDIS_PORT', 6379)),
    password=os.environ.get('REDIS_PASSWORD'),
    db=0,
    max_connections=20,
    decode_responses=True,  # Return strings instead of bytes
    socket_timeout=5,
    socket_connect_timeout=5,
    retry=Retry(ExponentialBackoff(), 3),
    retry_on_error=[redis.ConnectionError, redis.TimeoutError],
)

r = redis.Redis(connection_pool=pool)

# Async connection (aioredis)
import aioredis

async def get_async_redis():
    return await aioredis.from_url(
        f"redis://{os.environ.get('REDIS_HOST', 'localhost')}:6379",
        password=os.environ.get('REDIS_PASSWORD'),
        max_connections=20,
        decode_responses=True,
    )
```

## Data Structures by Use Case

### Strings — Simple Key-Value Cache

```python
# Basic cache
r.setex("user:123", 3600, json.dumps(user_data))   # Set with 1h TTL
data = json.loads(r.get("user:123") or "null")       # Get

# Atomic counter
r.incr("page_views:home")                            # +1
r.incrby("page_views:home", 5)                       # +5
r.incrbyfloat("balance:user123", -29.99)              # Floating point

# Conditional set
r.set("lock:resource", "owner123", nx=True, ex=30)   # SET if not exists, 30s TTL
r.set("config:version", "v2", xx=True)                # SET only if exists

# Batch operations (pipeline)
pipe = r.pipeline()
for user_id in user_ids:
    pipe.get(f"user:{user_id}")
results = pipe.execute()  # Single round-trip for all gets
```

### Hashes — Structured Objects

```python
# Store object as hash (more memory-efficient than JSON string for small objects)
r.hset("user:123", mapping={
    "name": "Alice",
    "email": "alice@example.com",
    "plan": "premium",
    "login_count": "0",
})

# Get single field
email = r.hget("user:123", "email")

# Get all fields
user = r.hgetall("user:123")

# Increment numeric field
r.hincrby("user:123", "login_count", 1)

# Check field exists
r.hexists("user:123", "email")  # True/False
```

### Lists — Queues and Activity Feeds

```python
# Simple job queue (FIFO)
r.lpush("queue:emails", json.dumps(email_job))     # Enqueue
job = json.loads(r.rpop("queue:emails") or "null")  # Dequeue

# Blocking pop (worker pattern)
result = r.brpop("queue:emails", timeout=30)  # Block up to 30s
if result:
    queue_name, job_data = result

# Activity feed (keep last 100 items)
r.lpush(f"feed:{user_id}", json.dumps(activity))
r.ltrim(f"feed:{user_id}", 0, 99)  # Keep only 100 most recent

# Get paginated feed
page = r.lrange(f"feed:{user_id}", 0, 19)  # First 20 items
```

### Sets — Unique Collections and Set Operations

```python
# Track unique visitors
r.sadd(f"visitors:{today}", user_id)
unique_count = r.scard(f"visitors:{today}")

# Tags / categories
r.sadd(f"post:{post_id}:tags", "python", "redis", "tutorial")
tags = r.smembers(f"post:{post_id}:tags")

# Find common interests between users
common = r.sinter(f"interests:{user1}", f"interests:{user2}")

# Random member (lottery, sampling)
winner = r.srandmember("contest:entries")
```

### Sorted Sets — Rankings, Time-Series, Priority Queues

```python
# Leaderboard
r.zadd("leaderboard:weekly", {"alice": 1500, "bob": 1200, "charlie": 1800})
top_10 = r.zrevrange("leaderboard:weekly", 0, 9, withscores=True)
rank = r.zrevrank("leaderboard:weekly", "alice")  # 0-based

# Delayed job queue (score = execution timestamp)
execute_at = time.time() + 300  # 5 minutes from now
r.zadd("delayed_jobs", {json.dumps(job): execute_at})

# Process due jobs
now = time.time()
due_jobs = r.zrangebyscore("delayed_jobs", 0, now)
for job in due_jobs:
    r.zrem("delayed_jobs", job)
    process_job(json.loads(job))

# Time-series with automatic cleanup
r.zadd(f"metrics:{metric_name}", {json.dumps(data_point): timestamp})
r.zremrangebyscore(f"metrics:{metric_name}", 0, time.time() - 86400)  # Keep 24h
```

### Streams — Event Log / Message Queue

```python
# Producer: append to stream
r.xadd("events:orders", {
    "order_id": "ORD-123",
    "customer_id": "CUST-456",
    "amount": "99.99",
    "status": "created",
})

# Consumer group: reliable message processing
# Create consumer group (once)
try:
    r.xgroup_create("events:orders", "order-processors", id="0", mkstream=True)
except redis.ResponseError:
    pass  # Group already exists

# Consumer: read messages
messages = r.xreadgroup(
    groupname="order-processors",
    consumername="worker-1",
    streams={"events:orders": ">"},  # ">" = only new messages
    count=10,
    block=5000,  # Block 5 seconds
)

for stream, entries in messages:
    for msg_id, data in entries:
        process_order_event(data)
        r.xack("events:orders", "order-processors", msg_id)  # Acknowledge

# Check pending messages (unacknowledged)
pending = r.xpending("events:orders", "order-processors")
# Claim abandoned messages (idle > 60s)
claimed = r.xclaim(
    "events:orders", "order-processors", "worker-1",
    min_idle_time=60000,  # 60s
    message_ids=[msg_id for msg_id in pending_ids],
)
```

## Lua Scripting for Atomic Operations

```python
# Atomic compare-and-swap
cas_script = r.register_script("""
local key = KEYS[1]
local expected = ARGV[1]
local new_value = ARGV[2]
local ttl = tonumber(ARGV[3])

local current = redis.call('GET', key)
if current == expected then
    if ttl > 0 then
        redis.call('SETEX', key, ttl, new_value)
    else
        redis.call('SET', key, new_value)
    end
    return 1
end
return 0
""")

# Usage
success = cas_script(keys=["counter"], args=["old_value", "new_value", 3600])

# Atomic rate limiter with sliding window
rate_limit_script = r.register_script("""
local key = KEYS[1]
local max_requests = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local now = tonumber(ARGV[3])

-- Remove expired entries
redis.call('ZREMRANGEBYSCORE', key, 0, now - window)

-- Count current window
local count = redis.call('ZCARD', key)

if count < max_requests then
    -- Add this request
    redis.call('ZADD', key, now, now .. ':' .. math.random(1000000))
    redis.call('EXPIRE', key, window)
    return 1  -- Allowed
end
return 0  -- Denied
""")
```

## Express.js Middleware Patterns

### Cache Middleware

```typescript
import { Request, Response, NextFunction } from 'express';
import Redis from 'ioredis';

const redis = new Redis();

function cacheMiddleware(ttl: number = 300) {
  return async (req: Request, res: Response, next: NextFunction) => {
    if (req.method !== 'GET') return next();

    const key = `cache:${req.originalUrl}`;
    const cached = await redis.get(key);

    if (cached) {
      const { statusCode, body, headers } = JSON.parse(cached);
      res.set(headers);
      res.set('X-Cache', 'HIT');
      return res.status(statusCode).json(body);
    }

    // Monkey-patch res.json to cache the response
    const originalJson = res.json.bind(res);
    res.json = (body: any) => {
      redis.setex(key, ttl, JSON.stringify({
        statusCode: res.statusCode,
        body,
        headers: { 'Content-Type': 'application/json' },
      }));
      res.set('X-Cache', 'MISS');
      return originalJson(body);
    };

    next();
  };
}

// Usage
app.get('/api/products', cacheMiddleware(300), getProducts);
app.get('/api/products/:id', cacheMiddleware(600), getProduct);

// Invalidate on mutation
app.post('/api/products', async (req, res) => {
  await createProduct(req.body);
  // Invalidate list cache
  const keys = await redis.keys('cache:/api/products*');
  if (keys.length) await redis.del(...keys);
  res.status(201).json({ success: true });
});
```

### Rate Limit Middleware

```typescript
function rateLimitMiddleware(maxRequests: number = 100, windowSeconds: number = 60) {
  return async (req: Request, res: Response, next: NextFunction) => {
    const identifier = req.ip || req.headers['x-forwarded-for'] || 'unknown';
    const key = `ratelimit:${identifier}`;
    const now = Date.now() / 1000;

    const pipe = redis.pipeline();
    pipe.zremrangebyscore(key, 0, now - windowSeconds);
    pipe.zadd(key, now, `${now}:${Math.random()}`);
    pipe.zcard(key);
    pipe.expire(key, windowSeconds);
    const results = await pipe.exec();

    const count = results![2][1] as number;
    const remaining = Math.max(0, maxRequests - count);

    res.set('X-RateLimit-Limit', String(maxRequests));
    res.set('X-RateLimit-Remaining', String(remaining));
    res.set('X-RateLimit-Reset', String(Math.ceil(now + windowSeconds)));

    if (count > maxRequests) {
      return res.status(429).json({
        error: 'Too many requests',
        retryAfter: windowSeconds,
      });
    }

    next();
  };
}

app.use('/api/', rateLimitMiddleware(100, 60));
```

## Redis Configuration for Production

```conf
# redis.conf essential settings

# Memory
maxmemory 2gb
maxmemory-policy allkeys-lru

# Persistence (RDB + AOF for durability)
save 900 1          # Snapshot if 1+ keys changed in 900s
save 300 10         # Snapshot if 10+ keys changed in 300s
save 60 10000       # Snapshot if 10000+ keys changed in 60s
appendonly yes
appendfsync everysec

# Security
requirepass your_strong_password
rename-command FLUSHDB ""     # Disable dangerous commands
rename-command FLUSHALL ""
rename-command DEBUG ""
rename-command CONFIG ""

# Performance
tcp-keepalive 300
timeout 0
hz 10
```

## Checklist Before Completing

- [ ] Connection pool configured with retry strategy
- [ ] Graceful shutdown handles pending commands
- [ ] Keys have TTL set (no unbounded growth)
- [ ] Pipeline/batch used for bulk operations
- [ ] Lua scripts used for multi-step atomic operations
- [ ] Error handling for connection failures and timeouts
- [ ] Key naming convention is consistent (colons as separators)
- [ ] Memory eviction policy matches use case
- [ ] Monitoring for hit rate, memory, and latency
