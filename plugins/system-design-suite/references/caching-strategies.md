# Caching Strategies Reference

Quick reference for caching patterns and strategies. For full architectural context, see the scalability-engineer agent.

---

## Caching Patterns

### Cache-Aside (Lazy Loading)

Application manages the cache explicitly. Most common pattern.

```
Read:
  1. Check cache
  2. Cache hit → return cached data
  3. Cache miss → read from DB → write to cache → return

Write:
  1. Write to DB
  2. Invalidate cache (DELETE, not UPDATE)

┌───────┐  1.get  ┌───────┐
│  App  │────────►│ Cache │
│       │◄────────│       │
└───┬───┘  2.hit  └───────┘
    │
    │ 3.miss
    ▼
┌───────┐
│  DB   │
└───────┘
```

**Pros**: Only caches what's actually requested. Cache failure doesn't break reads (falls through to DB).
**Cons**: Cache miss = 3 round trips (check cache, read DB, write cache). Cold cache is slow.

**Implementation**:
```javascript
async function getUser(userId) {
  const key = `user:${userId}`;

  // 1. Try cache
  const cached = await redis.get(key);
  if (cached) return JSON.parse(cached);

  // 2. Cache miss — fetch from DB
  const user = await db.query('SELECT * FROM users WHERE id = $1', [userId]);
  if (!user) return null;

  // 3. Populate cache with TTL
  await redis.set(key, JSON.stringify(user), 'EX', 300);

  return user;
}

async function updateUser(userId, data) {
  await db.query('UPDATE users SET name = $2 WHERE id = $1', [userId, data.name]);

  // Invalidate — don't update (avoids race conditions)
  await redis.del(`user:${userId}`);
}
```

**Why invalidate instead of update?**
```
Race condition with cache update:
  Thread A: Update DB (name = "Alice")
  Thread B: Update DB (name = "Bob")
  Thread B: Update cache (name = "Bob")
  Thread A: Update cache (name = "Alice")  ← STALE! DB has "Bob"

With invalidation:
  Thread A: Update DB (name = "Alice")
  Thread B: Update DB (name = "Bob")
  Thread A: Delete cache
  Thread B: Delete cache
  Next read: Fetches "Bob" from DB ← CORRECT
```

### Read-Through

Cache sits between application and database. Cache loads data from DB on miss.

```
┌───────┐  get  ┌───────┐  miss  ┌───────┐
│  App  │──────►│ Cache │───────►│  DB   │
│       │◄──────│(loads │◄───────│       │
└───────┘       │ on    │        └───────┘
                │ miss) │
                └───────┘
```

**Pros**: Simpler application code (no cache miss handling). Cache manages its own population.
**Cons**: First request always slow (cold start). Limited to key-value lookups.

**Difference from cache-aside**: In cache-aside, the application manages cache population. In read-through, the cache layer itself fetches from the database on a miss.

### Write-Through

Every write goes through the cache to the database. Cache is always up to date.

```
┌───────┐  write  ┌───────┐  write  ┌───────┐
│  App  │────────►│ Cache │────────►│  DB   │
│       │◄────────│       │◄────────│       │
└───────┘         └───────┘         └───────┘

App writes to cache → cache writes to DB → both confirm
```

**Pros**: Cache is always consistent with DB. Reads are always fast (data in cache).
**Cons**: Higher write latency (write to cache + DB). Caches data that may never be read.

**Best combined with**: Read-through (for complete cache management) or TTL-based eviction (to avoid caching unused data).

### Write-Behind (Write-Back)

Application writes to cache; cache asynchronously writes to database.

```
┌───────┐  write  ┌───────┐  async  ┌───────┐
│  App  │────────►│ Cache │ ·····►│  DB   │
│       │◄────────│       │  batch  │       │
└───────┘ (fast)  └───────┘  write  └───────┘
```

**Pros**: Very fast writes (only hits cache). Batching reduces DB write load. Can absorb traffic spikes.
**Cons**: Risk of data loss if cache fails before writing to DB. Complexity of the write queue. Eventual consistency between cache and DB.

**Use when**: Write-heavy workloads where write latency matters more than durability (analytics, view counts, real-time scores).

**Implementation considerations**:
```
Write buffer:
  - Batch writes every 100ms or every 100 items
  - Use a persistent queue (Redis streams, Kafka) as write buffer
  - Retry failed writes with backoff
  - Monitor queue depth (growing queue = DB can't keep up)

Data loss protection:
  - Use Redis with AOF persistence
  - Or write to both cache AND a write-ahead log
  - Accept that crash = loss of unflushed writes
```

### Refresh-Ahead

Proactively refresh cache entries before they expire.

```
Entry created with TTL = 60s
At 48s (80% of TTL): Background refresh triggered
  Cache → DB (async fetch)
  Cache updated → TTL reset

User always gets cached data (no miss penalty)
```

**Pros**: No cache miss latency for hot keys. Always-fresh data.
**Cons**: Wastes resources refreshing data that may not be requested. Complex to implement correctly.

**Use when**: Hot keys with predictable access patterns. Latency-sensitive endpoints where cache misses are unacceptable.

---

## Cache Invalidation Strategies

Cache invalidation is famously one of the two hard problems in computer science.

### TTL-Based (Time to Live)

```
redis.set('user:123', data, 'EX', 300);  // Expires in 5 minutes

Short TTL (1-60 seconds):
  - Near-real-time freshness
  - Higher DB load (more misses)
  - Use for: stock prices, inventory counts, live scores

Medium TTL (1-60 minutes):
  - Good balance of freshness and hit rate
  - Acceptable staleness for most applications
  - Use for: user profiles, product details, search results

Long TTL (1-24 hours):
  - High cache hit rate
  - Significant staleness possible
  - Use for: configuration, reference data, category trees

TTL selection framework:
  How often does this data change? → Set TTL to fraction of change frequency
  What's the cost of stale data? → Lower cost = longer TTL
  What's the cost of a cache miss? → Higher cost = longer TTL
```

### Event-Based Invalidation

```
When data changes, explicitly invalidate related cache entries.

Order created:
  → Invalidate user's order list: redis.del('user:123:orders')
  → Invalidate order count: redis.del('user:123:order_count')
  → Invalidate product inventory: redis.del('product:456:stock')

Implementation:
  // Define invalidation rules
  const invalidationRules = {
    'order.created': (event) => [
      `user:${event.user_id}:orders`,
      `user:${event.user_id}:order_count`,
      `product:${event.product_id}:stock`,
    ],
    'user.updated': (event) => [
      `user:${event.user_id}`,
      `user:${event.user_id}:profile`,
    ],
  };

  // On event, invalidate related keys
  function onEvent(eventType, event) {
    const keysToInvalidate = invalidationRules[eventType]?.(event) || [];
    if (keysToInvalidate.length > 0) {
      redis.del(...keysToInvalidate);
    }
  }
```

**Pros**: Immediate freshness when data changes. No unnecessary staleness.
**Cons**: Must track all dependencies (which cache keys depend on which data). Missing an invalidation = stale data bug.

### Tag-Based Invalidation

```
Tag cache entries with metadata. Invalidate all entries with a specific tag.

// Storing with tags
cache.set('product:123', data, {
  tags: ['products', 'category:electronics', 'brand:sony'],
  ttl: 3600,
});

cache.set('product:456', data, {
  tags: ['products', 'category:electronics', 'brand:samsung'],
  ttl: 3600,
});

// Invalidate by tag
cache.invalidateTag('category:electronics');
// → Both product:123 and product:456 invalidated

cache.invalidateTag('brand:sony');
// → Only product:123 invalidated
```

**Implementation** (Redis):
```
# Store tag → key mappings
SADD tag:category:electronics product:123 product:456
SADD tag:brand:sony product:123

# Invalidate tag
SMEMBERS tag:category:electronics → [product:123, product:456]
DEL product:123 product:456 tag:category:electronics
```

### Version-Based Invalidation

```
Include a version in the cache key. When data changes, increment version.
Old cache entries naturally expire (not accessed → TTL eviction).

Cache key: user:123:v5:profile
On update: increment version → user:123:v6:profile
Old key (v5) is never accessed again → TTL eviction removes it

Where to store version:
  - Redis: INCR user:123:version
  - Database column: users.cache_version
  - Application memory (if single server)

Pros: No explicit invalidation needed. Simple. No race conditions.
Cons: Old entries waste memory until TTL expires. Extra lookup for version.
```

---

## Cache Stampede Prevention

When a popular cache entry expires, many concurrent requests all try to rebuild it simultaneously, overwhelming the database.

### Locking

Only one request rebuilds the cache. Others wait.

```javascript
async function getWithLock(key, fetchFn, ttl) {
  // 1. Try cache
  let value = await redis.get(key);
  if (value) return JSON.parse(value);

  // 2. Try to acquire lock
  const lockKey = `lock:${key}`;
  const locked = await redis.set(lockKey, '1', 'NX', 'EX', 10);

  if (locked) {
    // 3. We got the lock — rebuild cache
    try {
      value = await fetchFn();
      await redis.set(key, JSON.stringify(value), 'EX', ttl);
      return value;
    } finally {
      await redis.del(lockKey);
    }
  } else {
    // 4. Someone else is rebuilding — wait and retry
    await sleep(100);
    return getWithLock(key, fetchFn, ttl);
  }
}
```

### Stale-While-Revalidate

Serve stale data while refreshing in the background.

```javascript
async function getWithStaleRefresh(key, fetchFn, ttl, staleTtl) {
  const entry = await redis.get(key);

  if (entry) {
    const { value, expiresAt } = JSON.parse(entry);

    if (Date.now() < expiresAt) {
      // Fresh — return immediately
      return value;
    }

    // Stale but available — return stale, refresh in background
    refreshInBackground(key, fetchFn, ttl, staleTtl);
    return value;
  }

  // No cached data at all — must fetch synchronously
  const value = await fetchFn();
  await cacheWithMeta(key, value, ttl, staleTtl);
  return value;
}

async function cacheWithMeta(key, value, ttl, staleTtl) {
  const entry = {
    value,
    expiresAt: Date.now() + ttl * 1000,
  };
  // Store with extended TTL (fresh TTL + stale TTL)
  await redis.set(key, JSON.stringify(entry), 'EX', ttl + staleTtl);
}
```

### Probabilistic Early Expiration (XFetch)

Each request has an increasing probability of refreshing the cache as it approaches expiry. Spreads refresh load over time.

```
probability = delta * beta * log(random()) > (expiry - now)

delta: time to recompute the value
beta: tuning parameter (typically 1.0)

At 80% of TTL: ~5% chance of refresh
At 90% of TTL: ~20% chance of refresh
At 95% of TTL: ~50% chance of refresh
At 100% of TTL: Must refresh

Result: Cache is refreshed by early requests, avoiding stampede at expiry
```

---

## Multi-Layer Caching

```
Request flow:
  Browser cache → CDN → Reverse proxy → App cache → DB cache → DB

Layer 1: Browser Cache (Cache-Control headers)
  - Static assets: immutable, max-age=31536000
  - API data: max-age=60 or no-store
  - User-specific: private

Layer 2: CDN (Cloudflare, CloudFront)
  - Static assets: edge cache, origin shield
  - API responses: s-maxage for public data
  - Surrogate-Key for tag-based invalidation

Layer 3: Reverse Proxy (Nginx, Varnish)
  - Full page caching for anonymous users
  - Fragment caching (ESI)
  - Micro-caching (1-5 seconds for dynamic pages)

Layer 4: Application Cache (Redis)
  - Database query results
  - Computed values (aggregations, rankings)
  - Session data
  - Rate limit counters

Layer 5: Database Cache (Buffer Pool)
  - Automatic (managed by DB engine)
  - shared_buffers in PostgreSQL (25% of RAM)
  - innodb_buffer_pool_size in MySQL (70-80% of RAM)
```

### Cache Warming

Pre-populate the cache before it's needed.

```
When to warm:
  - After deployment (new cache instances)
  - After cache flush
  - Before expected traffic spike (Super Bowl, Black Friday)
  - After maintenance window

How to warm:
  1. From access logs: Replay top-N most accessed keys
  2. From analytics: Pre-cache known hot data
  3. From database: Query and cache anticipated data
  4. Gradual: Route small % of traffic to new cache, let it fill naturally

// Warm cache from top accessed keys
async function warmCache(topKeys) {
  const BATCH_SIZE = 100;
  for (let i = 0; i < topKeys.length; i += BATCH_SIZE) {
    const batch = topKeys.slice(i, i + BATCH_SIZE);
    await Promise.all(batch.map(async (key) => {
      const data = await db.query('SELECT * FROM ... WHERE id = $1', [key]);
      await redis.set(key, JSON.stringify(data), 'EX', 3600);
    }));
    // Rate limit to avoid overwhelming DB
    await sleep(100);
  }
}
```

---

## Redis-Specific Patterns

### Data Structures for Caching

```
Strings:
  Simple key-value. Use for: serialized objects, counters, flags.
  SET user:123 '{"name":"Alice"}' EX 300

Hashes:
  Field-value pairs. Use for: objects with partial updates.
  HSET user:123 name "Alice" email "alice@example.com"
  HGET user:123 name  ← get single field without deserializing entire object

Sorted Sets:
  Ordered by score. Use for: leaderboards, time-series, rate limiting.
  ZADD leaderboard 100 "user:123"
  ZRANGEBYSCORE leaderboard 0 +inf LIMIT 0 10  ← top 10

Sets:
  Unordered unique values. Use for: tags, memberships, deduplication.
  SADD user:123:roles admin editor
  SISMEMBER user:123:roles admin  ← O(1) membership check

Lists:
  Ordered. Use for: queues, recent items, activity feeds.
  LPUSH user:123:recent_views product:456
  LTRIM user:123:recent_views 0 49  ← keep last 50
```

### Redis Cluster vs Sentinel

```
Redis Sentinel (HA for single-master):
  ┌──────────┐
  │  Master  │ ← writes
  └────┬─────┘
  ┌────┴────┐
  │         │
┌─▼───────┐ ┌▼────────┐
│Replica 1│ │Replica 2│ ← reads
└─────────┘ └─────────┘
  Sentinel monitors, auto-failover on master death
  Scale reads with replicas
  No sharding (single master for writes)
  Use when: < 25 GB data, write throughput manageable by one node

Redis Cluster (sharding + HA):
  ┌──────────┐ ┌──────────┐ ┌──────────┐
  │ Master 1 │ │ Master 2 │ │ Master 3 │
  │ Slot 0-  │ │ Slot     │ │ Slot     │
  │ 5460     │ │ 5461-    │ │ 10923-   │
  └────┬─────┘ │ 10922    │ │ 16383    │
  ┌────▼─────┐ └────┬─────┘ └────┬─────┘
  │Replica 1a│ ┌────▼─────┐ ┌────▼─────┐
  └──────────┘ │Replica 2a│ │Replica 3a│
               └──────────┘ └──────────┘
  Data sharded across masters (16384 hash slots)
  Each master has replicas for HA
  Use when: > 25 GB data or need to scale write throughput
```

### Memory Management

```
Eviction policies (when memory is full):

noeviction:       Return errors on writes (no data loss, but writes fail)
allkeys-lru:      Evict least recently used key (general purpose)
allkeys-lfu:      Evict least frequently used key (better for skewed access)
volatile-lru:     Evict LRU among keys with TTL (protect non-expiring keys)
volatile-lfu:     Evict LFU among keys with TTL
volatile-ttl:     Evict key with shortest remaining TTL
allkeys-random:   Random eviction (when all keys are equally important)

Recommended:
  Cache use case → allkeys-lfu (evict rarely used items)
  Mixed use (cache + persistent) → volatile-lfu (only evict TTL keys)

Memory optimization:
  - Use hashes for small objects (Redis optimizes small hashes into zipmaps)
  - Use short key names for high-cardinality keys
  - Set maxmemory to 75% of available RAM (leave room for fragmentation)
  - Monitor: INFO memory, memory usage sampling
```

---

## Strategy Selection Guide

```
Pattern               When to Use                          Tradeoffs
──────────────────────────────────────────────────────────────────────────
Cache-aside           Most common, general purpose          3 round trips on miss
Read-through          Simpler app code, uniform access      Cold start penalty
Write-through         Consistency critical                  Higher write latency
Write-behind          Write-heavy, latency critical         Data loss risk
Refresh-ahead         Hot keys, latency sensitive           Wasted refreshes

TTL invalidation      Simple, acceptable staleness          Not immediately fresh
Event invalidation    Need immediate freshness              Complex dependency tracking
Tag invalidation      Group invalidation needed             Extra storage for tags
Version invalidation  Simple, no explicit invalidation      Wasted memory until TTL

Locking (stampede)    Prevent DB overload on miss           Added latency for waiters
Stale-while-revalidate  User experience over freshness      Occasionally stale
XFetch                High-traffic, prevent synchronized    Approximate
```
