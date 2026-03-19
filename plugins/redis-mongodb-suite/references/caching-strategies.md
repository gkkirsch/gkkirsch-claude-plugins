# Caching Strategies Reference

Comprehensive reference for caching architectures, invalidation strategies, and operational patterns.

---

## Cache Architecture Patterns

### Cache-Aside (Lazy Loading)

```
Application reads:
  1. Check cache → HIT → return cached data
  2. MISS → read from database → write to cache → return data

Application writes:
  1. Write to database
  2. Invalidate cache (delete, not update)

Pros: Only caches what's actually read, cache failure is non-fatal
Cons: Cache miss penalty (cold start), potential stale data window
Best for: General-purpose caching, read-heavy workloads
```

### Write-Through

```
Application writes:
  1. Write to cache AND database synchronously
  2. Both must succeed before returning

Application reads:
  1. Always read from cache (always fresh)

Pros: Cache always consistent with DB, no stale reads
Cons: Write latency (both cache + DB), caches data that may never be read
Best for: Read-heavy with strong consistency requirements
```

### Write-Behind (Write-Back)

```
Application writes:
  1. Write to cache immediately (user sees instant update)
  2. Queue async write to database (batch/periodic flush)

Application reads:
  1. Always read from cache

Pros: Fastest writes, batch efficiency for DB
Cons: Risk of data loss if cache fails before flush, consistency lag
Best for: Write-heavy workloads, analytics, activity tracking
```

### Read-Through

```
Application reads:
  1. Request data from cache
  2. Cache misses → cache itself fetches from DB → populates itself → returns

Application never directly reads from database.

Pros: Clean separation, cache manages its own population
Cons: Initial read penalty, cache library must support data source integration
Best for: CDN-backed content, ORM-integrated caching
```

### Refresh-Ahead

```
Cache proactively refreshes entries before they expire:
  1. Entry TTL set to 60s
  2. At 50s (before expiry), cache pre-fetches fresh data in background
  3. User always gets cached data (no miss penalty)

Pros: No cache miss latency for popular keys
Cons: Wastes resources refreshing data nobody reads, complexity
Best for: Predictable access patterns, dashboard data, popular content
```

---

## Invalidation Strategies

### Time-Based (TTL)

```
Strategy: Set expiration time on cache entries

Short TTL (5-60s):
  - Real-time data (stock prices, live scores)
  - Session-adjacent data
  - Acceptable staleness: seconds

Medium TTL (5-60 min):
  - API responses, product catalogs
  - User profiles, settings
  - Acceptable staleness: minutes

Long TTL (1-24 hours):
  - Static content metadata
  - Configuration data
  - Reference data (countries, categories)

Permanent (no TTL):
  - Immutable data (historical records, content-addressed assets)
  - Use with LRU eviction as memory safety net
```

### Event-Driven Invalidation

```
Strategy: Invalidate cache when source data changes

Patterns:
  1. Direct: Write handler deletes cache key after DB write
  2. Pub/Sub: Publish invalidation event, all instances subscribe
  3. Change Data Capture: DB changelog triggers cache invalidation
  4. Webhook: External system notifies cache on change

Best practice: Delete (invalidate) rather than update cache entries
  - Avoids race conditions between concurrent updates
  - Simpler to reason about
  - Cache repopulates on next read (cache-aside)
```

### Tag-Based Invalidation

```
Strategy: Associate cache entries with tags, invalidate all entries by tag

Example:
  Cache key: "product:123" → tags: ["products", "category:electronics"]
  Cache key: "product:456" → tags: ["products", "category:clothing"]
  Cache key: "products:list:page1" → tags: ["products"]

  invalidate("category:electronics") → deletes product:123
  invalidate("products") → deletes ALL product-related cache entries

Implementation: Redis SET per tag containing associated cache keys
  SADD "tag:products" "product:123" "product:456" "products:list:page1"
  # On invalidation:
  SMEMBERS "tag:products" → get all keys → DEL all keys + tag key
```

### Version-Based Invalidation

```
Strategy: Include version number in cache key, increment on change

Cache key: "products:v42:list:page1"
On product update: increment version → "products:v43:..."
Old keys expire via TTL (no explicit deletion needed)

Pros: No distributed cache deletion needed, atomic switchover
Cons: Memory waste during version transitions, needs version counter
Best for: CDN purge avoidance, atomic cache switchovers
```

---

## Cache Stampede Prevention

### Problem

```
When a popular cache key expires:
  100 concurrent requests all miss cache simultaneously
  → 100 identical database queries
  → Database overload
  → Cascading failure
```

### Solution 1: Mutex / Locking

```python
def get_with_lock(key, fetch_fn, ttl=300, lock_ttl=10):
    cached = redis.get(key)
    if cached:
        return json.loads(cached)

    lock_key = f"lock:{key}"
    # Try to acquire lock
    if redis.set(lock_key, "1", nx=True, ex=lock_ttl):
        try:
            data = fetch_fn()
            redis.setex(key, ttl, json.dumps(data))
            return data
        finally:
            redis.delete(lock_key)
    else:
        # Another process is fetching — wait and retry
        time.sleep(0.1)
        return get_with_lock(key, fetch_fn, ttl, lock_ttl)
```

### Solution 2: Probabilistic Early Expiration (XFetch)

```python
import math, random

def get_with_xfetch(key, fetch_fn, ttl=300, beta=1.0):
    cached = redis.get(key)
    if cached:
        data = json.loads(cached)
        delta = data["_delta"]      # Time to compute
        expiry = data["_expiry"]    # When it expires
        now = time.time()

        # Probabilistic early refresh
        # Higher beta = more aggressive early refresh
        if now - delta * beta * math.log(random.random()) < expiry:
            return data["value"]    # Still valid, return cached

    # Cache miss or probabilistic early refresh
    start = time.time()
    value = fetch_fn()
    delta = time.time() - start

    redis.setex(key, ttl, json.dumps({
        "value": value,
        "_delta": delta,
        "_expiry": time.time() + ttl,
    }))
    return value
```

### Solution 3: Stale-While-Revalidate

```python
def get_with_swr(key, fetch_fn, ttl=300, swr_window=60):
    cached = redis.get(key)
    if cached:
        data = json.loads(cached)
        age = time.time() - data["_cached_at"]

        if age < ttl:
            return data["value"]     # Fresh

        if age < ttl + swr_window:
            # Stale but within SWR window — return stale, refresh async
            Thread(target=refresh_cache, args=(key, fetch_fn, ttl)).start()
            return data["value"]

    # Miss or too stale
    return refresh_cache(key, fetch_fn, ttl)

def refresh_cache(key, fetch_fn, ttl):
    value = fetch_fn()
    redis.setex(key, ttl + 120, json.dumps({
        "value": value,
        "_cached_at": time.time(),
    }))
    return value
```

---

## Multi-Layer Cache Architecture

### L1: In-Process + L2: Distributed

```
Request Flow:
  1. Check L1 (in-process memory) → <1ms
  2. Check L2 (Redis) → 1-5ms
  3. Check L3 (database) → 10-100ms

Invalidation Flow:
  1. Delete from L1 (local)
  2. Delete from L2 (Redis)
  3. Publish invalidation event (Redis pub/sub)
  4. All app instances clear L1 for that key

L1 Characteristics:
  - LRU cache in application memory
  - Very short TTL (10-60 seconds)
  - Small size (1000-10000 entries)
  - Not shared between instances
  - Risk: stale data across instances

L2 Characteristics:
  - Redis (shared across all instances)
  - Medium TTL (5-60 minutes)
  - Large size (millions of entries)
  - Shared, consistent view
  - Risk: network latency, Redis failure
```

### CDN + Application + Database

```
Request Flow:
  1. CDN edge cache → 0ms (served from edge)
  2. CDN miss → origin server
  3. Application cache (Redis) → 1-5ms
  4. Application miss → database → 10-100ms
  5. Response flows back through all layers

CDN Configuration:
  - Cache-Control headers control CDN behavior
  - s-maxage for CDN-specific TTL
  - stale-while-revalidate for background refresh
  - Vary header for content negotiation
  - Surrogate keys for targeted purging

Invalidation:
  - Application cache: direct delete
  - CDN: purge API call or wait for TTL
  - Emergency: lower TTL + force purge
```

---

## HTTP Caching Reference

### Cache-Control Header Builder

```
Static assets (JS/CSS with content hash):
  Cache-Control: public, max-age=31536000, immutable

HTML pages:
  Cache-Control: public, max-age=0, must-revalidate

Public API (short cache):
  Cache-Control: public, max-age=60, stale-while-revalidate=300

Private API (user-specific):
  Cache-Control: private, no-cache

Sensitive data (banking, health):
  Cache-Control: no-store

CDN-optimized API:
  Cache-Control: public, s-maxage=300, max-age=60, stale-while-revalidate=86400, stale-if-error=86400
```

### ETag and Conditional Requests

```
First request:
  Response: 200 OK
  ETag: "abc123"
  Cache-Control: private, max-age=0, must-revalidate

Second request (after max-age expires):
  Request: If-None-Match: "abc123"
  Response: 304 Not Modified (no body transferred)

With Last-Modified:
  Response: Last-Modified: Wed, 15 Jan 2024 12:00:00 GMT
  Request: If-Modified-Since: Wed, 15 Jan 2024 12:00:00 GMT
  Response: 304 Not Modified
```

### Vary Header

```
Vary: Accept-Encoding
  → Separate cache entries for gzip vs brotli vs identity

Vary: Accept-Language
  → Separate cache entries per language

Vary: Authorization
  → Effectively disables shared caching (each user gets own entry)

Vary: Accept-Encoding, Accept-Language
  → Separate entries for each combination (can explode cache size)
```

---

## Cache Sizing and Capacity Planning

### Memory Estimation

```
Per cached item:
  Key overhead: ~70 bytes (Redis internal + key string)
  Value: actual data size
  Hash entry overhead: ~100 bytes (if using Hash)
  Set member overhead: ~70 bytes

Example calculation:
  1 million user profiles at 500 bytes each:
  Keys: 1M × 70 bytes = 70 MB
  Values: 1M × 500 bytes = 500 MB
  Total: ~570 MB + Redis overhead (~20%) = ~684 MB

  Recommendation: 1 GB maxmemory for this workload
```

### Hit Rate Targets

| Workload | Target Hit Rate | Notes |
|----------|----------------|-------|
| Static content | > 95% | Long TTL, rarely changes |
| Product catalog | > 90% | Moderate TTL, periodic invalidation |
| User profiles | > 80% | Per-user keys, session-length TTL |
| Search results | > 70% | High cardinality, shorter TTL |
| Real-time data | > 50% | Very short TTL, frequent invalidation |

### Warming Strategies

```
Cold Start:
  Option 1: Accept slow start, cache fills organically
  Option 2: Pre-warm popular keys on deploy

Pre-warming script:
  1. Query database for top N items (by access frequency)
  2. Batch-load into cache via pipeline
  3. Run as deploy hook before routing traffic

Continuous warming:
  Cron job refreshes popular keys before TTL expires
  Schedule: TTL minus 20% buffer (e.g., 60 min TTL → refresh at 48 min)
```

---

## Monitoring and Alerting

### Key Metrics

| Metric | How to Measure | Alert Threshold |
|--------|---------------|----------------|
| Hit rate | hits / (hits + misses) | < 80% (investigate) |
| Miss rate | misses / (hits + misses) | > 20% sustained |
| Eviction rate | evicted_keys delta | Any increase |
| Memory usage | used_memory_rss | > 80% maxmemory |
| Key count | dbsize | Unexpected growth |
| Latency p99 | client-side measurement | > 10ms (Redis), > 100ms (CDN) |
| Connection count | connected_clients | > 80% maxclients |
| Error rate | cache operations that threw | > 0.1% |
| Stale serve rate | swr responses / total | Track, no hard threshold |
| Cold start duration | time until hit rate > threshold | Varies by workload |

### Health Check Endpoint

```javascript
app.get("/health/cache", async (req, res) => {
  try {
    const start = Date.now();
    await redis.ping();
    const latency = Date.now() - start;

    const info = await redis.info("stats");
    const hits = parseInt(info.match(/keyspace_hits:(\d+)/)?.[1] || "0");
    const misses = parseInt(info.match(/keyspace_misses:(\d+)/)?.[1] || "0");
    const hitRate = hits / (hits + misses) || 0;

    const memInfo = await redis.info("memory");
    const usedMemory = parseInt(memInfo.match(/used_memory:(\d+)/)?.[1] || "0");
    const maxMemory = parseInt(memInfo.match(/maxmemory:(\d+)/)?.[1] || "0");
    const memoryPct = maxMemory > 0 ? (usedMemory / maxMemory) * 100 : 0;

    res.json({
      status: latency < 50 && hitRate > 0.7 ? "healthy" : "degraded",
      latency_ms: latency,
      hit_rate: (hitRate * 100).toFixed(1) + "%",
      memory_pct: memoryPct.toFixed(1) + "%",
      connected: true,
    });
  } catch (err) {
    res.status(503).json({
      status: "unhealthy",
      connected: false,
      error: err.message,
    });
  }
});
```

---

## Common Pitfalls

| Pitfall | Consequence | Prevention |
|---------|-------------|------------|
| No TTL on cache entries | Memory grows unbounded until OOM | Always set TTL, use LRU eviction |
| Caching errors/empty results | Users see persistent errors | Only cache successful responses |
| Cache key collisions | Wrong data served to users | Include all discriminating params in key |
| Thundering herd on expiry | Database overload | Use mutex, XFetch, or SWR |
| Updating cache instead of deleting | Race condition: stale update wins | Always delete on write, repopulate on read |
| Caching personalized content publicly | Data leak between users | Use `private` scope, include user ID in key |
| No graceful degradation | App crashes when cache is down | Wrap cache ops in try/catch, serve from DB |
| Cache warming too aggressive | Wastes memory on unused data | Only warm actually popular keys |
| Inconsistent key naming | Can't find/invalidate keys | Enforce naming convention |
| Serialization overhead | Slow cache ops | Use efficient formats (msgpack, protobuf) |
| No monitoring | Silent cache failures | Track hit rate, latency, evictions |
| Testing only with warm cache | Miss cold-start behavior | Test with empty cache in staging |
