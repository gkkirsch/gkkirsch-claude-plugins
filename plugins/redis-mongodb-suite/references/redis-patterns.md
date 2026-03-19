# Redis Patterns Reference

Comprehensive reference for Redis data structures, patterns, configuration, and operations.

---

## Data Structure Selection Guide

| Use Case | Data Structure | Why |
|----------|---------------|-----|
| Simple cache | String | Fast get/set, TTL support |
| Object with fields | Hash | Memory-efficient, partial reads |
| Queue / activity feed | List | O(1) push/pop, blocking pop |
| Unique collection | Set | O(1) membership test, set ops |
| Ranking / time-series | Sorted Set | Score-based ordering, range queries |
| Event log / message bus | Stream | Consumer groups, persistence |
| Counting unique items | HyperLogLog | ~0.81% error, 12KB per key |
| Geospatial | Sorted Set (GEO) | Radius search, distance calc |
| Bitmap flags | String (BITFIELD) | Compact boolean arrays |

---

## String Patterns

### Atomic Counters

```
INCR page_views:homepage           # +1
INCRBY page_views:homepage 5       # +5
DECR stock:SKU-001                  # -1
INCRBYFLOAT balance:user123 -29.99  # Float arithmetic

# Get-and-set (atomic swap)
GETSET lock:resource "new_owner"    # Returns old value

# Conditional set
SET lock:resource "owner1" NX EX 30  # SET if Not eXists, 30s TTL
SET config:v "2" XX                   # SET only if eXists
```

### Expiration Patterns

```
SET session:abc123 "{...}" EX 3600     # Set with TTL (seconds)
SET session:abc123 "{...}" PX 3600000  # Set with TTL (milliseconds)
EXPIRE key 3600                         # Add TTL to existing key
EXPIREAT key 1735689600                 # Expire at Unix timestamp
PERSIST key                             # Remove TTL (make permanent)
TTL key                                 # Check remaining TTL (-1=no TTL, -2=doesn't exist)
```

### Batch Operations (Pipeline)

```python
# Python — single round-trip for multiple commands
pipe = redis.pipeline()
pipe.set("key1", "val1")
pipe.set("key2", "val2")
pipe.get("key3")
pipe.incr("counter")
results = pipe.execute()  # [True, True, "val3", 42]
```

```typescript
// Node.js (ioredis)
const pipe = redis.pipeline();
pipe.set("key1", "val1");
pipe.set("key2", "val2");
pipe.get("key3");
pipe.incr("counter");
const results = await pipe.exec();
// [[null, "OK"], [null, "OK"], [null, "val3"], [null, 42]]
```

---

## Hash Patterns

### Object Storage

```
HSET user:123 name "Alice" email "alice@example.com" plan "premium"
HGET user:123 email                    # Single field
HMGET user:123 name email              # Multiple fields
HGETALL user:123                       # All fields
HINCRBY user:123 login_count 1         # Atomic increment
HDEL user:123 deprecated_field         # Remove field
HEXISTS user:123 email                 # Check field existence
HLEN user:123                          # Count of fields
```

### When to Use Hash vs String

| Criterion | Hash | String (JSON) |
|-----------|------|---------------|
| Partial read/update | Yes (HGET/HSET) | No (full deserialize) |
| Nested objects | No (flat only) | Yes |
| Memory (small objects) | Better (ziplist encoding) | Worse |
| Memory (large objects) | Similar | Similar |
| Atomic field increment | Yes (HINCRBY) | No |
| TTL per field | No (only per key) | N/A |

**Rule of thumb**: Use Hash when you need partial field access. Use String+JSON when you always read/write the whole object or need nested structure.

---

## List Patterns

### FIFO Queue

```
LPUSH queue:emails '{"to":"alice@...","subject":"..."}'   # Enqueue (left)
RPOP queue:emails                                          # Dequeue (right)
BRPOP queue:emails 30                                      # Blocking dequeue (30s timeout)
LLEN queue:emails                                          # Queue depth
```

### Reliable Queue (RPOPLPUSH)

```
# Move item from main queue to processing queue atomically
RPOPLPUSH queue:emails queue:emails:processing

# On success: remove from processing queue
LREM queue:emails:processing 1 '{"to":"alice@..."}'

# On failure: items stay in processing queue for retry
# Periodic cleanup job moves stale items back to main queue
```

### Capped List (Activity Feed)

```
LPUSH feed:user123 '{"action":"liked","post":"..."}'
LTRIM feed:user123 0 99                    # Keep only 100 most recent
LRANGE feed:user123 0 19                   # Get page 1 (20 items)
LRANGE feed:user123 20 39                  # Get page 2
```

---

## Set Patterns

### Unique Tracking

```
SADD visitors:2024-01-15 "user:123" "user:456"
SCARD visitors:2024-01-15              # Count unique visitors
SISMEMBER visitors:2024-01-15 "user:123"  # Check membership

# Set operations
SINTER interests:user1 interests:user2     # Common interests
SUNION tags:post1 tags:post2               # All tags from both
SDIFF followers:user1 followers:user2      # Followers of user1 but not user2
SRANDMEMBER contest:entries                # Random pick
SPOP contest:entries                       # Random pick + remove
```

### Tag System

```
# Add tags to items
SADD item:123:tags "electronics" "wireless" "bluetooth"
SADD tag:electronics item:123 item:456 item:789

# Find items with ALL tags (intersection)
SINTER tag:electronics tag:wireless     # Items tagged both

# Find items with ANY tag (union)
SUNION tag:electronics tag:wireless     # Items tagged either
```

---

## Sorted Set Patterns

### Leaderboard

```
ZADD leaderboard:weekly alice 1500 bob 1200 charlie 1800
ZREVRANGE leaderboard:weekly 0 9 WITHSCORES    # Top 10
ZREVRANK leaderboard:weekly alice                # Alice's rank (0-based)
ZSCORE leaderboard:weekly alice                  # Alice's score
ZINCRBY leaderboard:weekly 100 alice             # Add 100 to Alice's score
ZCOUNT leaderboard:weekly 1000 2000              # Count scores between 1000-2000
```

### Delayed Job Queue

```python
import time, json

# Schedule job for 5 minutes from now
execute_at = time.time() + 300
redis.zadd("delayed_jobs", {json.dumps(job_data): execute_at})

# Worker: poll for due jobs
while True:
    now = time.time()
    due = redis.zrangebyscore("delayed_jobs", 0, now, start=0, num=10)
    for job in due:
        if redis.zrem("delayed_jobs", job):  # Atomic claim
            process(json.loads(job))
    time.sleep(1)
```

### Sliding Window Rate Limiter

```python
import time

def is_allowed(redis, key, max_requests, window_seconds):
    now = time.time()
    pipe = redis.pipeline()
    pipe.zremrangebyscore(key, 0, now - window_seconds)  # Remove expired
    pipe.zadd(key, {f"{now}:{id(now)}": now})             # Add current
    pipe.zcard(key)                                        # Count in window
    pipe.expire(key, window_seconds)                       # Auto-cleanup
    results = pipe.execute()
    return results[2] <= max_requests
```

---

## Stream Patterns

### Producer-Consumer with Consumer Groups

```
# Create stream + consumer group
XGROUP CREATE events:orders order-processors $ MKSTREAM

# Produce event
XADD events:orders * order_id ORD-123 amount 99.99 status created

# Consume (worker-1 in order-processors group)
XREADGROUP GROUP order-processors worker-1 COUNT 10 BLOCK 5000 STREAMS events:orders >

# Acknowledge processing
XACK events:orders order-processors 1234567890-0

# Check pending (unacknowledged) messages
XPENDING events:orders order-processors - + 10

# Claim abandoned messages (idle > 60s)
XCLAIM events:orders order-processors worker-2 60000 1234567890-0

# Trim stream to max length
XTRIM events:orders MAXLEN ~ 10000   # Approximate trim (faster)
```

### Stream vs Pub/Sub vs List

| Feature | Stream | Pub/Sub | List |
|---------|--------|---------|------|
| Message persistence | Yes | No | Yes |
| Consumer groups | Yes | No | No (manual) |
| Message acknowledgment | Yes | No | No |
| Replay/reread | Yes | No | Yes (destructive) |
| Fan-out | Yes (groups) | Yes (channels) | No |
| Blocking read | Yes | Yes (subscribe) | Yes (BRPOP) |
| Best for | Event sourcing, reliable queues | Real-time notifications | Simple FIFO queues |

---

## Pub/Sub Patterns

### Channel-Based Messaging

```python
# Publisher
redis.publish("notifications:user123", json.dumps({
    "type": "new_message",
    "from": "bob",
    "preview": "Hey, are you free..."
}))

# Subscriber (blocking — needs dedicated connection)
pubsub = redis.pubsub()
pubsub.subscribe("notifications:user123")

for message in pubsub.listen():
    if message["type"] == "message":
        data = json.loads(message["data"])
        handle_notification(data)

# Pattern subscribe (wildcard channels)
pubsub.psubscribe("notifications:*")
```

### Cache Invalidation via Pub/Sub

```python
# When cache entry is invalidated, notify all app instances
def invalidate_cache(key):
    local_cache.delete(key)
    redis.delete(key)
    redis.publish("cache:invalidate", json.dumps({"key": key}))

# Each app instance subscribes
def start_invalidation_listener():
    sub = redis.pubsub()
    sub.subscribe("cache:invalidate")
    for msg in sub.listen():
        if msg["type"] == "message":
            data = json.loads(msg["data"])
            local_cache.delete(data["key"])  # Clear local L1 cache
```

---

## Distributed Lock (Redlock Pattern)

### Single-Instance Lock

```python
import uuid, time

def acquire_lock(redis, lock_name, ttl=10):
    token = str(uuid.uuid4())
    acquired = redis.set(f"lock:{lock_name}", token, nx=True, ex=ttl)
    return token if acquired else None

def release_lock(redis, lock_name, token):
    # Lua script for atomic check-and-delete
    script = """
    if redis.call('GET', KEYS[1]) == ARGV[1] then
        return redis.call('DEL', KEYS[1])
    end
    return 0
    """
    return redis.eval(script, 1, f"lock:{lock_name}", token)

# Usage
token = acquire_lock(redis, "process_payment")
if token:
    try:
        process_payment()
    finally:
        release_lock(redis, "process_payment", token)
```

### Lock with Auto-Extension

```python
import threading

class DistributedLock:
    def __init__(self, redis, name, ttl=10):
        self.redis = redis
        self.name = f"lock:{name}"
        self.ttl = ttl
        self.token = str(uuid.uuid4())
        self._renewal_thread = None
        self._stop_renewal = threading.Event()

    def acquire(self, timeout=30):
        deadline = time.time() + timeout
        while time.time() < deadline:
            if self.redis.set(self.name, self.token, nx=True, ex=self.ttl):
                self._start_renewal()
                return True
            time.sleep(0.1)
        return False

    def release(self):
        self._stop_renewal.set()
        if self._renewal_thread:
            self._renewal_thread.join()
        release_lock(self.redis, self.name.replace("lock:", ""), self.token)

    def _start_renewal(self):
        def renew():
            while not self._stop_renewal.wait(self.ttl / 3):
                self.redis.expire(self.name, self.ttl)
        self._renewal_thread = threading.Thread(target=renew, daemon=True)
        self._renewal_thread.start()

    def __enter__(self):
        if not self.acquire():
            raise TimeoutError("Could not acquire lock")
        return self

    def __exit__(self, *args):
        self.release()
```

---

## Session Management

```python
import json, uuid, time

class RedisSessionStore:
    def __init__(self, redis, prefix="sess:", ttl=3600):
        self.redis = redis
        self.prefix = prefix
        self.ttl = ttl

    def create(self, user_id, data=None):
        session_id = str(uuid.uuid4())
        session = {
            "user_id": user_id,
            "created_at": time.time(),
            "data": data or {},
        }
        self.redis.setex(
            f"{self.prefix}{session_id}",
            self.ttl,
            json.dumps(session)
        )
        return session_id

    def get(self, session_id):
        raw = self.redis.get(f"{self.prefix}{session_id}")
        if not raw:
            return None
        # Sliding expiration — reset TTL on read
        self.redis.expire(f"{self.prefix}{session_id}", self.ttl)
        return json.loads(raw)

    def update(self, session_id, data):
        session = self.get(session_id)
        if not session:
            return False
        session["data"].update(data)
        self.redis.setex(
            f"{self.prefix}{session_id}",
            self.ttl,
            json.dumps(session)
        )
        return True

    def destroy(self, session_id):
        return self.redis.delete(f"{self.prefix}{session_id}")

    def destroy_all_for_user(self, user_id):
        # Requires a secondary index
        sessions = self.redis.smembers(f"user_sessions:{user_id}")
        if sessions:
            pipe = self.redis.pipeline()
            for sid in sessions:
                pipe.delete(f"{self.prefix}{sid}")
            pipe.delete(f"user_sessions:{user_id}")
            pipe.execute()
```

---

## Key Naming Conventions

```
Pattern: {service}:{entity}:{id}:{subfield}

Examples:
  cache:user:123                  # Cached user object
  cache:api:GET:/products?p=1     # Cached API response
  sess:abc123def456               # Session
  lock:process_orders             # Distributed lock
  queue:emails                    # Job queue
  ratelimit:ip:192.168.1.1        # Rate limit counter
  feed:user:123                   # Activity feed
  leaderboard:weekly              # Sorted set leaderboard
  metrics:api_latency             # Time-series metric
  tag:electronics                 # Tag → item mapping
  pub:notifications:user123       # Pub/Sub channel (conceptual)
  stream:events:orders            # Stream name
```

**Rules**:
- Colons (`:`) as separators (Redis convention)
- Lowercase, no spaces
- Specific → general (entity type before ID)
- Consistent prefix per use case (`cache:`, `sess:`, `lock:`, etc.)
- Keep keys short — they consume memory too

---

## Memory Optimization

### Encoding Thresholds

Redis uses compact encodings for small values:

| Type | Compact Encoding | Threshold | Config Directive |
|------|-----------------|-----------|-----------------|
| Hash | ziplist | <= 128 fields, values <= 64 bytes | `hash-max-ziplist-entries`, `hash-max-ziplist-value` |
| List | ziplist → quicklist | <= 128 elements, values <= 64 bytes | `list-max-ziplist-size` |
| Set | intset (integers only) | <= 512 elements | `set-max-intset-entries` |
| Sorted Set | ziplist | <= 128 elements, values <= 64 bytes | `zset-max-ziplist-entries`, `zset-max-ziplist-value` |

**Optimization tips**:
- Keep Hash values under 64 bytes to use ziplist encoding (10x less memory)
- Use integer IDs in Sets to use intset encoding
- Batch small objects into a single Hash (key = object ID, field = serialized object)

### Memory Analysis Commands

```
MEMORY USAGE key                    # Bytes used by specific key
INFO memory                         # Overall memory stats
MEMORY DOCTOR                       # Memory health check
DEBUG OBJECT key                    # Internal encoding info
OBJECT ENCODING key                 # Current encoding type
OBJECT IDLETIME key                 # Seconds since last access
```

---

## Eviction Policies

| Policy | Behavior | Best For |
|--------|----------|----------|
| `noeviction` | Return error when full | Data must not be lost |
| `allkeys-lru` | Evict least recently used | General cache |
| `allkeys-lfu` | Evict least frequently used | Hot/cold data pattern |
| `allkeys-random` | Random eviction | Uniform access |
| `volatile-lru` | LRU among keys with TTL | Mixed cache + persistent |
| `volatile-lfu` | LFU among keys with TTL | Mixed with hot/cold |
| `volatile-random` | Random among TTL keys | Mixed, simple |
| `volatile-ttl` | Evict nearest expiry | Temporal data |

**Recommendation**: `allkeys-lru` for pure cache. `volatile-lru` for mixed cache + persistent data.

---

## Production Configuration Reference

```conf
# === Memory ===
maxmemory 4gb
maxmemory-policy allkeys-lru
maxmemory-samples 10              # LRU precision (higher = more accurate, slower)

# === Persistence ===
# RDB snapshots
save 900 1                         # Snapshot if 1+ keys changed in 15 min
save 300 10                        # Snapshot if 10+ keys changed in 5 min
save 60 10000                      # Snapshot if 10000+ keys changed in 1 min
rdbcompression yes
rdbchecksum yes
dbfilename dump.rdb

# AOF (Append Only File)
appendonly yes
appendfsync everysec               # fsync every second (recommended)
auto-aof-rewrite-percentage 100    # Rewrite when AOF is 2x baseline
auto-aof-rewrite-min-size 64mb     # Minimum size before rewrite

# === Networking ===
bind 127.0.0.1 -::1               # Listen on localhost only
port 6379
tcp-keepalive 300                  # TCP keepalive in seconds
timeout 0                          # No idle timeout (0 = disabled)
tcp-backlog 511                    # Connection backlog

# === Security ===
requirepass your_strong_password_here
# Disable dangerous commands
rename-command FLUSHDB ""
rename-command FLUSHALL ""
rename-command DEBUG ""
rename-command CONFIG "CONFIG_SECRET_CMD"
rename-command KEYS ""             # KEYS is O(N) — use SCAN instead

# === Performance ===
hz 10                              # Server tick rate (10-100)
dynamic-hz yes                     # Adjust hz based on connected clients
lazyfree-lazy-eviction yes         # Async eviction (no main thread blocking)
lazyfree-lazy-expire yes           # Async expiration
lazyfree-lazy-server-del yes       # Async DEL for large keys

# === Clients ===
maxclients 10000                   # Max concurrent connections

# === Slow Log ===
slowlog-log-slower-than 10000      # Log commands slower than 10ms
slowlog-max-len 128                # Keep last 128 slow entries
```

---

## Monitoring Commands

```
# Real-time stats
INFO all                           # Full server info
INFO stats                         # Hit/miss counts
INFO memory                        # Memory usage
INFO clients                       # Connected clients
INFO replication                   # Replication status

# Performance
SLOWLOG GET 10                     # Last 10 slow commands
SLOWLOG LEN                        # Number of slow log entries
SLOWLOG RESET                      # Clear slow log
LATENCY LATEST                     # Latest latency events
LATENCY HISTORY event-name         # History for specific event

# Key analysis (production-safe)
SCAN 0 MATCH "cache:*" COUNT 100   # Iterate keys by pattern
TYPE key                            # Get key type
OBJECT ENCODING key                 # Internal encoding
OBJECT IDLETIME key                 # Last access time
OBJECT FREQ key                     # Access frequency (LFU mode)

# Dangerous (never in production)
KEYS *                              # O(N) — blocks server. Use SCAN instead
DEBUG SLEEP 5                       # Blocks server for 5 seconds
MONITOR                             # Dumps all commands (high overhead)
```

### Key Metrics to Track

| Metric | Source | Alert Threshold |
|--------|--------|----------------|
| Hit rate | `INFO stats` → keyspace_hits / (hits+misses) | < 80% |
| Memory usage | `INFO memory` → used_memory_rss | > 80% of maxmemory |
| Connected clients | `INFO clients` → connected_clients | > 80% of maxclients |
| Evicted keys | `INFO stats` → evicted_keys | Any increase (investigate) |
| Blocked clients | `INFO clients` → blocked_clients | > 0 sustained |
| Replication lag | `INFO replication` → master_repl_offset diff | > 1MB |
| Slow commands | SLOWLOG LEN | Increasing trend |
| CPU usage | `INFO cpu` → used_cpu_sys + used_cpu_user | > 70% |

---

## Cluster Operations Reference

### Cluster Setup

```
# Create 6-node cluster (3 masters + 3 replicas)
redis-cli --cluster create \
  node1:6379 node2:6379 node3:6379 \
  node4:6379 node5:6379 node6:6379 \
  --cluster-replicas 1

# Check cluster health
redis-cli --cluster check node1:6379

# Reshard (move hash slots between nodes)
redis-cli --cluster reshard node1:6379

# Add node to cluster
redis-cli --cluster add-node new_node:6379 existing_node:6379

# Remove node
redis-cli --cluster del-node node1:6379 node_id
```

### Cluster Client Configuration

```typescript
// ioredis cluster client
const cluster = new Redis.Cluster([
  { host: "node1", port: 6379 },
  { host: "node2", port: 6379 },
  { host: "node3", port: 6379 },
], {
  redisOptions: {
    password: process.env.REDIS_PASSWORD,
  },
  scaleReads: "slave",           // Read from replicas
  maxRedirections: 16,           // Follow MOVED/ASK redirections
  retryDelayOnClusterDown: 300,  // ms to wait when cluster is down
  clusterRetryStrategy: (times) => Math.min(times * 100, 3000),
});
```

### Hash Tags (Force Keys to Same Slot)

```
# Keys with same hash tag go to same slot → can be used in multi-key commands
SET {user:123}:profile "..."
SET {user:123}:settings "..."
# Both keys hash on "user:123" → same slot → MGET works

MGET {user:123}:profile {user:123}:settings   # OK in cluster
MGET user:123:profile user:456:settings        # ERROR — cross-slot
```

---

## Anti-Patterns

| Anti-Pattern | Problem | Solution |
|-------------|---------|----------|
| `KEYS *` in production | O(N), blocks server | Use `SCAN` with cursor |
| No TTL on cache keys | Memory grows unbounded | Always set TTL |
| Large values (>100KB) | Blocks event loop | Compress or split |
| Hot key | Single shard overloaded | Add random suffix, client-side cache |
| MONITOR in production | Massive overhead | Use SLOWLOG instead |
| Missing retry strategy | Silent failures | Configure retryStrategy |
| Storing secrets without encryption | Plaintext in memory | Encrypt before storing |
| Using SELECT (multiple databases) | No cluster support, confusing | Use key prefixes |
| Unbounded lists/sets | Memory growth | Use LTRIM, MAXLEN |
| Synchronous FLUSHALL | Blocks for seconds | Use ASYNC flag |
