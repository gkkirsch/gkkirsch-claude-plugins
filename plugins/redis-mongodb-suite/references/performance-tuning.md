# Performance Tuning Reference

> Redis and MongoDB performance optimization covering memory management, query optimization,
> connection pooling, indexing strategies, and production monitoring.

---

## Redis Performance

### Memory Optimization

#### Data Type Memory Footprint

```
Data type memory comparison (approximate, per entry):

  String key + small value:     ~90 bytes overhead + key + value
  Hash (< 128 fields, ziplist): ~50 bytes overhead + all field-value pairs
  Hash (> 128 fields, hashtable): ~800+ bytes overhead
  List (< 128 elements, listpack): ~50 bytes overhead
  Set (< 128 members, listpack): ~50 bytes overhead
  Sorted Set (< 128 members, listpack): ~50 bytes overhead

Key insight: Small hashes in ziplist encoding use 10x LESS memory than
equivalent string keys. Batch small keys into hashes.
```

#### Hash Packing (Memory Optimization)

```python
# BAD: 1 million string keys
for i in range(1_000_000):
    redis.set(f"user:{i}:name", names[i])
# Memory: ~90MB

# GOOD: Pack into hash buckets (100 entries per hash)
for i in range(1_000_000):
    bucket = i // 100
    redis.hset(f"users:{bucket}", str(i), names[i])
# Memory: ~16MB (5.6x reduction!)

# Configure ziplist thresholds
# redis.conf:
# hash-max-ziplist-entries 128    # Max entries for ziplist encoding
# hash-max-ziplist-value 64       # Max value size for ziplist
# list-max-ziplist-size -2        # Listpack max size (8kb)
# zset-max-ziplist-entries 128
# zset-max-ziplist-value 64
```

#### Memory Eviction Policies

```
maxmemory-policy options:

  noeviction          — Return errors when memory full (default)
  allkeys-lru         — Evict least recently used keys (RECOMMENDED for caches)
  allkeys-lfu         — Evict least frequently used keys (better for skewed access)
  volatile-lru        — Evict LRU keys WITH expire set
  volatile-lfu        — Evict LFU keys WITH expire set
  volatile-ttl        — Evict keys closest to expiration
  allkeys-random      — Random eviction (uniform, fast)
  volatile-random     — Random eviction from keys with expire

  Recommendation:
    Pure cache → allkeys-lru or allkeys-lfu
    Mixed (cache + persistent data) → volatile-lru (only evicts expiring keys)
    Session store → volatile-ttl

  LFU tuning:
    lfu-log-factor 10       — Higher = slower counter growth (default: 10)
    lfu-decay-time 1        — Minutes before counter halves (default: 1)
```

#### Memory Analysis

```bash
# Built-in memory analysis
redis-cli INFO memory
# used_memory:1073741824          # Bytes used by Redis
# used_memory_human:1.00G
# used_memory_rss:1207959552      # RSS (what OS reports)
# used_memory_peak:1342177280     # Peak usage
# mem_fragmentation_ratio:1.12    # RSS/used — >1.5 = fragmentation issue
# mem_allocator:jemalloc-5.2.1

# Find large keys
redis-cli --bigkeys                # Per-type largest keys
redis-cli --memkeys --samples 100  # Sample key memory usage

# Per-key memory
redis-cli MEMORY USAGE user:12345
redis-cli MEMORY USAGE user:12345 SAMPLES 0  # Exact (slower)

# Memory doctor
redis-cli MEMORY DOCTOR
# Reports issues like high fragmentation, peak memory, etc.

# Fragmentation fix
redis-cli MEMORY PURGE              # Release memory to OS (jemalloc)
# Or enable active defrag:
# activedefrag yes
# active-defrag-enabled yes
# active-defrag-threshold-lower 10   # Start defrag at 10% fragmentation
# active-defrag-threshold-upper 100  # Max effort at 100% fragmentation
```

### Slow Command Diagnosis

```bash
# Enable slow log
CONFIG SET slowlog-log-slower-than 10000    # Log commands > 10ms (microseconds)
CONFIG SET slowlog-max-len 128              # Keep last 128 entries

# View slow log
SLOWLOG GET 10                              # Last 10 slow commands
SLOWLOG LEN                                 # Total entries
SLOWLOG RESET                               # Clear

# Each entry shows:
# 1) Unique ID
# 2) Unix timestamp
# 3) Execution time (microseconds)
# 4) Command + arguments
# 5) Client IP:port
# 6) Client name

# Common slow command culprits:
# KEYS *          — O(N), blocks everything. Use SCAN instead.
# SORT            — O(N+M*log(M)). Consider pre-sorting with sorted sets.
# SMEMBERS        — O(N) for large sets. Use SSCAN for iteration.
# HGETALL         — O(N) for large hashes. Use HSCAN or HMGET specific fields.
# LRANGE 0 -1     — O(N) for large lists. Paginate with offset+count.
# FLUSHDB/FLUSHALL — O(N), blocks. Use ASYNC variant.
# DEBUG SLEEP      — Never in production.
```

### Pipeline Performance

```
Without pipeline:
  Client → SET key1 → Server → OK → Client
  Client → SET key2 → Server → OK → Client
  Client → SET key3 → Server → OK → Client
  Total: 3 round-trips × network latency

With pipeline:
  Client → SET key1, SET key2, SET key3 → Server → OK, OK, OK → Client
  Total: 1 round-trip × network latency

Performance impact (1ms network latency):
  100 commands without pipeline: ~100ms
  100 commands with pipeline:    ~1ms + command execution time

  10,000 commands without pipeline: ~10 seconds
  10,000 commands with pipeline:    ~10ms + command execution time
```

```python
# Optimal pipeline batch size: 100-1000 commands
# Above 1000: split into chunks to prevent blocking

def batch_pipeline(redis, commands, batch_size=500):
    results = []
    for i in range(0, len(commands), batch_size):
        pipe = redis.pipeline(transaction=False)
        for cmd, args in commands[i:i+batch_size]:
            getattr(pipe, cmd)(*args)
        results.extend(pipe.execute())
    return results
```

### Connection Management

```python
# Connection pool sizing
# Rule: pool_size = max_concurrent_requests × commands_per_request
# For most apps: 10-50 connections

import redis

pool = redis.ConnectionPool(
    host='localhost',
    port=6379,
    max_connections=20,          # Max connections in pool
    socket_timeout=5,            # Read timeout (seconds)
    socket_connect_timeout=2,    # Connect timeout
    socket_keepalive=True,       # TCP keepalive
    health_check_interval=30,    # Check connection health every 30s
    retry_on_timeout=True,       # Retry on timeout
)

# Connection pool monitoring
pool_info = {
    "in_use": pool._in_use_connections,
    "available": pool._available_connections,
    "max": pool.max_connections,
    "created": pool._created_connections,
}
```

### Redis Benchmark Baseline

```bash
# Built-in benchmark tool
redis-benchmark -h localhost -p 6379 -c 50 -n 100000

# Specific commands
redis-benchmark -t set,get -c 50 -n 100000 -d 256

# Pipeline benchmark
redis-benchmark -t set -c 50 -n 100000 -P 16    # 16 commands per pipeline

# Typical results (local, single thread):
# SET:  ~150,000 ops/sec
# GET:  ~150,000 ops/sec
# INCR: ~150,000 ops/sec
# LPUSH: ~150,000 ops/sec
# With pipeline (P=16): ~600,000-1,000,000 ops/sec
```

---

## MongoDB Performance

### Query Optimization

#### Explain Output Analysis

```javascript
// Always run explain on important queries
const explain = db.orders.find({
  status: "shipped",
  created_at: { $gte: new Date("2024-01-01") }
}).sort({ created_at: -1 }).explain("executionStats");

// Key metrics to check:
const stats = explain.executionStats;

// 1. Execution time
stats.executionTimeMillis;           // Total time

// 2. Documents examined vs returned (should be close to 1:1)
stats.totalDocsExamined;             // Documents loaded from disk
stats.nReturned;                     // Documents returned to client
// Ratio > 10:1 = needs better index

// 3. Keys examined vs returned
stats.totalKeysExamined;             // Index entries scanned
// Should be close to nReturned for equality queries

// 4. Stage analysis
const plan = explain.queryPlanner.winningPlan;
// COLLSCAN = no index (bad)
// IXSCAN = using index (good)
// FETCH = loading full document after index scan
// SORT = in-memory sort (bad for large results)
// SORT_MERGE = merging sorted results from multiple index scans
// PROJECTION_COVERED = answered entirely from index (best)

// 5. Rejected plans
explain.queryPlanner.rejectedPlans;  // Alternative plans considered
```

#### Covered Queries (Index-Only)

```javascript
// A covered query is answered entirely from the index — no document fetch
// This is the fastest possible query

// Index
db.orders.createIndex({ status: 1, total: 1, created_at: 1 });

// Covered query (only returns indexed fields)
db.orders.find(
  { status: "shipped" },
  { _id: 0, status: 1, total: 1, created_at: 1 }  // Projection matches index
);

// NOT covered (includes non-indexed field)
db.orders.find(
  { status: "shipped" },
  { _id: 0, status: 1, total: 1, customer_name: 1 }  // customer_name not in index
);

// Verify with explain:
// executionStats.totalDocsExamined === 0 means covered query
```

#### Index Intersection vs Compound Indexes

```
Compound index:
  { status: 1, created_at: -1 }
  → Single index scan, very efficient
  → ALWAYS prefer this for known query patterns

Index intersection (MongoDB can combine two single-field indexes):
  { status: 1 } + { created_at: -1 }
  → Two index scans + merge
  → Slower than compound, but flexible
  → MongoDB may or may not use intersection (optimizer decides)

Rule: Create compound indexes for your most common queries.
      Don't rely on index intersection for performance.
```

### Aggregation Performance

```javascript
// Aggregation pipeline optimization rules:

// 1. $match FIRST — filter early, process less data
// GOOD:
[
  { $match: { status: "active" } },  // Filter first
  { $group: { _id: "$category", count: { $sum: 1 } } }
]

// BAD:
[
  { $group: { _id: "$category", docs: { $push: "$$ROOT" } } },  // Process all docs
  { $match: { "_id": "electronics" } }  // Filter after grouping
]

// 2. $project early — reduce document size in pipeline
[
  { $match: { status: "active" } },
  { $project: { category: 1, price: 1 } },  // Drop unneeded fields early
  { $group: { _id: "$category", avg_price: { $avg: "$price" } } }
]

// 3. Use $limit after $sort to enable top-K optimization
[
  { $sort: { score: -1 } },
  { $limit: 10 }  // MongoDB optimizes this into a single bounded sort
]

// 4. $lookup optimization — add $match inside lookup pipeline
[
  {
    $lookup: {
      from: "orders",
      let: { userId: "$_id" },
      pipeline: [
        { $match: { $expr: { $eq: ["$user_id", "$$userId"] } } },
        { $match: { status: { $ne: "cancelled" } } },  // Filter in lookup
        { $limit: 5 },                                   // Limit in lookup
        { $project: { total: 1, status: 1 } }           // Project in lookup
      ],
      as: "recent_orders"
    }
  }
]

// 5. allowDiskUse for large aggregations
db.collection.aggregate(pipeline, { allowDiskUse: true });
// Default: 100MB memory limit per stage
// With allowDiskUse: spills to disk (slower but doesn't crash)
```

### Write Performance

#### Bulk Operations

```javascript
// Ordered bulk (stops on first error): ~2x faster than individual writes
// Unordered bulk (continues on error): ~3-4x faster than individual writes

// Unordered is ALWAYS faster when order doesn't matter
const result = await db.collection.bulkWrite(operations, {
  ordered: false,           // Parallel execution
  writeConcern: { w: 1 }   // Acknowledge from primary only (fastest)
});

// Write concern performance impact:
// w: 0              — Unacknowledged (fastest, no confirmation)
// w: 1              — Primary acknowledged (default, good balance)
// w: "majority"     — Majority acknowledged (safest, ~2x slower than w:1)
// j: true           — Journal write (adds ~5ms per write)
// j: true + w: "majority" — Safest (slowest)

// For bulk imports: use w:1, j:false, ordered:false
// For critical data: use w:"majority", j:true
```

#### Insert Optimization

```javascript
// Batch inserts (much faster than individual)
// Optimal batch size: 1000-5000 documents

const BATCH_SIZE = 2000;
for (let i = 0; i < documents.length; i += BATCH_SIZE) {
  const batch = documents.slice(i, i + BATCH_SIZE);
  await db.collection.insertMany(batch, {
    ordered: false,
    writeConcern: { w: 1 }
  });
}

// Pre-split for sharded collections
// Before bulk import, pre-split chunks to distribute writes
sh.splitAt("mydb.collection", { _id: splitPoint });
```

### Connection Pool Tuning

```javascript
// Node.js driver
const client = new MongoClient(uri, {
  maxPoolSize: 100,              // Max connections (default: 100)
  minPoolSize: 10,               // Min idle connections (default: 0)
  maxIdleTimeMS: 60000,          // Close idle connections after 60s
  waitQueueTimeoutMS: 10000,     // Timeout waiting for connection
  connectTimeoutMS: 10000,       // TCP connect timeout
  socketTimeoutMS: 45000,        // Socket read timeout
  serverSelectionTimeoutMS: 30000, // Server selection timeout
  heartbeatFrequencyMS: 10000,   // Replica set monitoring interval
  retryWrites: true,             // Retry failed writes once
  retryReads: true,              // Retry failed reads once
  compressors: ["snappy", "zstd"], // Network compression
  w: "majority",                 // Default write concern
  readPreference: "secondaryPreferred", // Default read preference
});

// Pool sizing formula:
// maxPoolSize = peak_concurrent_requests * avg_commands_per_request
// Usually: 50-200 for web servers, 10-50 for batch processors

// Monitor pool usage
client.on("connectionPoolCreated", (event) => { /* ... */ });
client.on("connectionCheckedOut", (event) => { /* ... */ });
client.on("connectionPoolCleared", (event) => { /* ... */ });
```

```python
# Python driver (PyMongo)
client = MongoClient(
    uri,
    maxPoolSize=100,
    minPoolSize=10,
    maxIdleTimeMS=60000,
    waitQueueTimeoutMS=10000,
    connectTimeoutMS=10000,
    socketTimeoutMS=45000,
    serverSelectionTimeoutMS=30000,
    retryWrites=True,
    retryReads=True,
    compressors="snappy,zstd",
    w="majority",
    readPreference="secondaryPreferred",
)
```

### WiredTiger Tuning

```
# mongod.conf WiredTiger settings

storage:
  wiredTiger:
    engineConfig:
      cacheSizeGB: 4              # Default: 50% of RAM - 1GB
                                  # Set to 60% of available RAM for dedicated servers
      journalCompressor: snappy   # Journal compression (snappy, zlib, zstd, none)
    collectionConfig:
      blockCompressor: zstd       # Collection compression
                                  # snappy: fast, moderate compression
                                  # zlib: better compression, slower
                                  # zstd: best balance (recommended)
                                  # none: no compression (fastest, most storage)
    indexConfig:
      prefixCompression: true     # Prefix compression for indexes (default: true)

# Cache sizing rules:
# 1. WiredTiger cache should hold your working set (hot data + indexes)
# 2. If cache < working set → excessive disk I/O
# 3. Monitor: db.serverStatus().wiredTiger.cache
#    "bytes currently in the cache" vs cacheSizeGB
#    "pages read into cache" — high = cache too small
#    "pages written from cache" — high = cache too small
#    "maximum bytes configured" — your limit
```

---

## Monitoring Dashboards

### Redis Metrics to Track

```
Category: Throughput
  instantaneous_ops_per_sec        — Current operations/sec
  total_commands_processed         — Lifetime command count
  keyspace_hits / keyspace_misses  — Cache hit rate

Category: Memory
  used_memory                      — Current memory usage
  used_memory_rss                  — OS-reported memory (includes fragmentation)
  mem_fragmentation_ratio          — RSS/used_memory (>1.5 = problem)
  used_memory_peak                 — Highest memory usage ever
  maxmemory                        — Configured limit
  evicted_keys                     — Keys evicted due to maxmemory

Category: Connections
  connected_clients                — Current client connections
  blocked_clients                  — Clients in BLPOP/BRPOP
  rejected_connections             — Connections rejected (maxclients)

Category: Persistence
  rdb_last_save_time               — Last successful RDB save
  rdb_changes_since_last_save      — Changes since last save
  aof_rewrite_in_progress          — AOF rewrite active?
  aof_last_bgrewrite_status        — Last AOF rewrite status

Category: Replication
  connected_slaves                 — Replica count
  master_repl_offset               — Replication offset
  repl_backlog_active              — Replication backlog enabled?

Category: Latency
  latency_latest / latency_history — Latency samples
  slowlog entries                  — Commands exceeding threshold

Alert thresholds:
  used_memory > 80% of maxmemory         → Warning
  mem_fragmentation_ratio > 1.5           → Investigate
  evicted_keys increasing                 → Need more memory or review TTLs
  keyspace_misses / (hits+misses) > 20%   → Cache not effective
  connected_clients > 80% of maxclients   → Scale connection pool
  rejected_connections > 0                → Increase maxclients
  instantaneous_ops_per_sec spike         → Investigate traffic
```

### MongoDB Metrics to Track

```
Category: Operations
  opcounters.query           — Queries/sec
  opcounters.insert          — Inserts/sec
  opcounters.update          — Updates/sec
  opcounters.delete          — Deletes/sec
  opcounters.command         — Commands/sec

Category: Connections
  connections.current        — Active connections
  connections.available      — Remaining available
  connections.totalCreated   — Lifetime connections

Category: Memory
  mem.resident               — Resident memory (MB)
  mem.virtual                — Virtual memory (MB)
  wiredTiger.cache."bytes currently in the cache"
  wiredTiger.cache."tracked dirty bytes in the cache"

Category: Disk I/O
  wiredTiger.cache."pages read into cache"       — Cache misses
  wiredTiger.cache."pages written from cache"     — Evictions/checkpoints
  wiredTiger.concurrentTransactions.read.out      — Active read tickets
  wiredTiger.concurrentTransactions.write.out     — Active write tickets

Category: Replication
  replSetGetStatus.members[].optimeDate          — Last oplog entry
  replication lag = primary.optimeDate - secondary.optimeDate

Category: Query Performance
  globalLock.currentQueue.readers   — Queued readers
  globalLock.currentQueue.writers   — Queued writers
  metrics.queryExecutor.scanned     — Keys examined
  metrics.queryExecutor.scannedObjects — Docs examined

Alert thresholds:
  connections.current > 80% of available          → Connection leak
  replication lag > 10 seconds                    → Replica falling behind
  globalLock.currentQueue > 0 sustained           → Lock contention
  cache dirty bytes > 20% of cache size           → Checkpoint pressure
  scannedObjects/returned ratio > 100:1           → Missing index
  oplog window < 24 hours                         → Increase oplog size
```

---

## Capacity Planning

### Redis Sizing

```
Memory estimation:

  1. Calculate per-key memory:
     key overhead + value overhead + data size
     ~90 bytes + key_length + value_length (for strings)

  2. Multiply by expected key count:
     1M keys × 200 bytes avg = ~200MB

  3. Add overhead:
     Base overhead: ~3MB
     Per-database: ~200 bytes
     Replication buffer: ~1MB
     Output buffers: client_count × buffer_size

  4. Apply safety margin:
     Recommended: set maxmemory to 75% of available RAM
     Remaining 25%: OS, fragmentation, child process (RDB/AOF), buffers

  Example sizing:
     10M sessions × 500 bytes avg = 5GB data
     + 20% overhead = 6GB
     + 25% safety margin = 8GB RAM recommended
     maxmemory = 6GB
```

### MongoDB Sizing

```
Working set estimation:

  1. Calculate data size:
     document_count × avg_document_size
     10M documents × 2KB = 20GB

  2. Calculate index size:
     Each indexed field: ~20-30 bytes per document
     Compound index: ~40-60 bytes per document
     5 indexes × 25 bytes × 10M docs = 1.25GB

  3. Working set = frequently accessed data + all indexes
     If 20% of data is hot: 4GB data + 1.25GB indexes = 5.25GB

  4. WiredTiger cache should hold the working set
     cacheSizeGB ≥ working set size

  5. Total RAM recommendation:
     WiredTiger cache (60% of RAM)
     + OS page cache (for memory-mapped files)
     + Connection overhead (~1MB per connection)
     + Application overhead

  Example:
     Working set: 5.25GB
     Cache size: 5.25GB (need 60% of RAM to be 5.25GB)
     Total RAM: 5.25 / 0.6 = ~9GB recommended

  Disk sizing:
     Data: 20GB
     Indexes: 1.25GB
     Oplog: 5% of data = 1GB
     Journal: ~1GB
     WiredTiger overhead: ~10% = 2GB
     Total: ~25GB
     With compression (zstd ~3:1): ~9GB on disk
     Recommended: 2-3x for growth = 25-35GB SSD
```

---

## Quick Checklist: Production Readiness

### Redis Production Checklist

```
[ ] maxmemory set (never rely on system OOM killer)
[ ] maxmemory-policy configured (allkeys-lru for cache, volatile-lru for mixed)
[ ] Persistence configured (RDB + AOF for durability)
[ ] Password set (requirepass)
[ ] Dangerous commands disabled (FLUSHDB, FLUSHALL, DEBUG, CONFIG)
[ ] Connection pool configured with retry strategy
[ ] Graceful shutdown handles pending commands
[ ] Slow log enabled (slowlog-log-slower-than 10000)
[ ] All keys have TTL (no unbounded growth)
[ ] No KEYS command in production code (use SCAN)
[ ] Pipeline used for batch operations
[ ] Monitoring for memory, hit rate, connections, latency
[ ] Sentinel or Cluster for high availability
[ ] Backup strategy (RDB snapshots to S3/GCS)
```

### MongoDB Production Checklist

```
[ ] Replica set deployed (minimum 3 members)
[ ] Authentication enabled (SCRAM-SHA-256)
[ ] Authorization configured (least privilege roles)
[ ] WiredTiger cache sized to working set
[ ] Compression enabled (zstd recommended)
[ ] All queries have supporting indexes (no COLLSCAN)
[ ] Compound indexes follow ESR rule
[ ] Connection pool sized appropriately
[ ] Write concern set to "majority" for critical data
[ ] Read preference configured for read scaling
[ ] Profiler enabled for slow queries (slowms: 100)
[ ] Monitoring for connections, replication lag, cache, operations
[ ] Oplog sized for 24+ hour window
[ ] Backup strategy (mongodump or Atlas backup)
[ ] Schema validation on critical collections
[ ] TTL indexes on temporary data (sessions, tokens, logs)
[ ] No unbounded arrays in document schemas
```
