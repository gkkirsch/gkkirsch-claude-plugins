# Redis Command Reference

> Complete Redis command patterns organized by data type and use case.
> Covers Redis 7.x commands with practical examples.

---

## String Commands

### Basic Operations

```
SET key value [EX seconds] [PX milliseconds] [NX|XX] [GET]
  SET user:123 "Alice"              -- Simple set
  SET user:123 "Alice" EX 3600      -- Set with 1h TTL
  SET lock:res "owner" NX EX 30     -- Set if not exists (distributed lock)
  SET config:v "2" XX               -- Set only if exists (update)
  SET counter "10" GET              -- Set and return old value (atomic swap)

GET key
  GET user:123                      -- Returns "Alice" or nil

MSET key value [key value ...]
  MSET user:1 "Alice" user:2 "Bob"  -- Set multiple keys atomically

MGET key [key ...]
  MGET user:1 user:2 user:3         -- Get multiple keys in one round-trip

GETSET key value                     -- Deprecated in 7.x, use SET ... GET
GETDEL key                           -- Get and delete (atomic)
GETEX key [EX seconds]               -- Get and set expiration

SETNX key value                      -- SET if Not eXists (returns 1/0)
SETEX key seconds value              -- SET with EXpiration
PSETEX key milliseconds value        -- SET with millisecond expiration

APPEND key value                     -- Append to string
STRLEN key                           -- String length
GETRANGE key start end               -- Substring (0-indexed, inclusive)
SETRANGE key offset value            -- Overwrite at offset
```

### Counters

```
INCR key                             -- Increment by 1 (atomic)
INCRBY key increment                 -- Increment by integer
INCRBYFLOAT key increment            -- Increment by float
DECR key                             -- Decrement by 1
DECRBY key decrement                 -- Decrement by integer

-- Pattern: Page view counter
INCR page_views:home
INCRBY page_views:home 5

-- Pattern: Rate limiter counter
INCR ratelimit:192.168.1.1
EXPIRE ratelimit:192.168.1.1 60
```

---

## Hash Commands

```
HSET key field value [field value ...]
  HSET user:123 name "Alice" email "alice@example.com" plan "premium"

HGET key field
  HGET user:123 email               -- Returns "alice@example.com"

HMGET key field [field ...]
  HMGET user:123 name email plan     -- Returns array of values

HGETALL key
  HGETALL user:123                   -- Returns all field-value pairs

HDEL key field [field ...]
  HDEL user:123 deprecated_field

HEXISTS key field                    -- Returns 1 if field exists
HLEN key                             -- Number of fields
HKEYS key                            -- All field names
HVALS key                            -- All values
HSETNX key field value               -- Set only if field doesn't exist

HINCRBY key field increment          -- Increment hash field (integer)
HINCRBYFLOAT key field increment     -- Increment hash field (float)

HSCAN key cursor [MATCH pattern] [COUNT count]
  HSCAN user:123 0 MATCH "pref_*"   -- Scan fields matching pattern

-- Pattern: Session storage
HSET session:abc123 user_id "456" role "admin" expires_at "1711929600"
HGET session:abc123 user_id
EXPIRE session:abc123 3600

-- Pattern: Feature flags
HSET features dark_mode "1" beta_ui "0" new_checkout "1"
HGET features dark_mode
```

---

## List Commands

```
LPUSH key element [element ...]      -- Push to head (left)
RPUSH key element [element ...]      -- Push to tail (right)
LPOP key [count]                     -- Pop from head
RPOP key [count]                     -- Pop from tail
LLEN key                             -- List length
LRANGE key start stop                -- Get range (0-indexed, inclusive, -1 = last)
LINDEX key index                     -- Get element at index
LSET key index element               -- Set element at index
LINSERT key BEFORE|AFTER pivot element
LREM key count element               -- Remove count occurrences
LTRIM key start stop                 -- Trim to range (keep only these elements)
LPOS key element [RANK rank] [COUNT count]  -- Find element position(s)

-- Blocking operations (worker pattern)
BLPOP key [key ...] timeout          -- Blocking pop from head
BRPOP key [key ...] timeout          -- Blocking pop from tail
BLMOVE source destination LEFT|RIGHT LEFT|RIGHT timeout

LMOVE source destination LEFT|RIGHT LEFT|RIGHT
  LMOVE queue:pending queue:processing LEFT RIGHT  -- Atomic move between lists

-- Pattern: Job queue
LPUSH queue:jobs '{"type":"email","to":"alice@example.com"}'
BRPOP queue:jobs 30                  -- Worker blocks until job available

-- Pattern: Activity feed (bounded)
LPUSH feed:user123 '{"action":"liked","post":456}'
LTRIM feed:user123 0 99             -- Keep only last 100 items

-- Pattern: Reliable queue (RPOPLPUSH pattern)
LMOVE queue:pending queue:processing RIGHT LEFT  -- Move job to processing
-- Process the job...
LREM queue:processing 1 job_data     -- Remove from processing when done
```

---

## Set Commands

```
SADD key member [member ...]         -- Add members
SREM key member [member ...]         -- Remove members
SISMEMBER key member                 -- Check membership (O(1))
SMISMEMBER key member [member ...]   -- Check multiple memberships
SMEMBERS key                         -- Get all members
SCARD key                            -- Set cardinality (count)
SRANDMEMBER key [count]              -- Random member(s)
SPOP key [count]                     -- Remove and return random member(s)

-- Set operations
SUNION key [key ...]                 -- Union of sets
SINTER key [key ...]                 -- Intersection of sets
SDIFF key [key ...]                  -- Difference (first minus others)
SUNIONSTORE dest key [key ...]       -- Store union result
SINTERSTORE dest key [key ...]       -- Store intersection result
SDIFFSTORE dest key [key ...]        -- Store difference result
SINTERCARD numkeys key [key ...] [LIMIT limit]  -- Count of intersection

SSCAN key cursor [MATCH pattern] [COUNT count]

-- Pattern: Unique visitors
SADD visitors:2024-03-19 "user:123"
SCARD visitors:2024-03-19            -- Unique count today

-- Pattern: Tag intersection
SADD tag:python "post:1" "post:3" "post:5"
SADD tag:redis "post:2" "post:3" "post:7"
SINTER tag:python tag:redis          -- Posts tagged both python AND redis

-- Pattern: Online users
SADD online:users "user:123"
SREM online:users "user:123"
SISMEMBER online:users "user:123"    -- Is user online?

-- Pattern: Friend recommendation (mutual friends)
SINTER friends:alice friends:bob     -- Common friends
SDIFF friends:alice friends:bob      -- Alice's friends who aren't Bob's
```

---

## Sorted Set Commands

```
ZADD key [NX|XX] [GT|LT] [CH] score member [score member ...]
  ZADD leaderboard 1500 "alice" 1200 "bob"     -- Add with scores
  ZADD leaderboard NX 1000 "charlie"             -- Only add new members
  ZADD leaderboard XX GT 1600 "alice"            -- Update only if new score > current
  ZADD leaderboard CH 1600 "alice"               -- Return count of changed (not just added)

ZSCORE key member                    -- Get member's score
ZMSCORE key member [member ...]      -- Get multiple scores
ZRANK key member                     -- 0-based rank (ascending)
ZREVRANK key member                  -- 0-based rank (descending)
ZCARD key                            -- Total members
ZCOUNT key min max                   -- Count members in score range

-- Range queries
ZRANGE key min max [BYSCORE|BYLEX] [REV] [LIMIT offset count] [WITHSCORES]
  ZRANGE leaderboard 0 9 REV WITHSCORES         -- Top 10 descending
  ZRANGE leaderboard 1000 2000 BYSCORE WITHSCORES  -- Score range 1000-2000
  ZRANGE leaderboard "(alpha" "[omega" BYLEX     -- Lexicographic range

ZRANGEBYSCORE key min max [WITHSCORES] [LIMIT offset count]  -- Legacy
ZREVRANGEBYSCORE key max min [WITHSCORES] [LIMIT offset count]

-- Modifications
ZINCRBY key increment member         -- Increment score
ZREM key member [member ...]         -- Remove members
ZREMRANGEBYRANK key start stop       -- Remove by rank range
ZREMRANGEBYSCORE key min max         -- Remove by score range

-- Set operations on sorted sets
ZUNIONSTORE dest numkeys key [key ...] [WEIGHTS weight ...] [AGGREGATE SUM|MIN|MAX]
ZINTERSTORE dest numkeys key [key ...] [WEIGHTS weight ...] [AGGREGATE SUM|MIN|MAX]
ZDIFFSTORE dest numkeys key [key ...]
ZRANDMEMBER key [count [WITHSCORES]]

ZPOPMIN key [count]                  -- Pop lowest-scored member(s)
ZPOPMAX key [count]                  -- Pop highest-scored member(s)
BZPOPMIN key [key ...] timeout       -- Blocking pop min
BZPOPMAX key [key ...] timeout       -- Blocking pop max

ZSCAN key cursor [MATCH pattern] [COUNT count]

-- Pattern: Leaderboard
ZADD leaderboard 1500 "alice"
ZINCRBY leaderboard 100 "alice"      -- Alice scored 100 more points
ZREVRANGE leaderboard 0 9 WITHSCORES -- Top 10
ZREVRANK leaderboard "alice"         -- Alice's rank

-- Pattern: Delayed job queue (score = timestamp)
ZADD delayed_jobs 1711929600 '{"type":"reminder","user":"123"}'
ZRANGEBYSCORE delayed_jobs 0 <current_timestamp>  -- Get due jobs
ZREM delayed_jobs job_data           -- Remove after processing

-- Pattern: Sliding window rate limiter
ZADD ratelimit:user123 <now> "<now>:<random>"
ZREMRANGEBYSCORE ratelimit:user123 0 <now - window>
ZCARD ratelimit:user123              -- Current request count
```

---

## Stream Commands

```
-- Producer
XADD key [NOMKSTREAM] [MAXLEN|MINID [=|~] threshold] *|ID field value [field value ...]
  XADD events * action "click" page "/home"       -- Auto-generated ID
  XADD events MAXLEN ~ 10000 * action "click"     -- Approximate trim to 10k

-- Consumer (without groups)
XREAD [COUNT count] [BLOCK milliseconds] STREAMS key [key ...] ID [ID ...]
  XREAD COUNT 10 BLOCK 5000 STREAMS events $      -- Read new entries, block 5s

-- Consumer groups
XGROUP CREATE key groupname ID|$ [MKSTREAM]
  XGROUP CREATE events processors 0 MKSTREAM      -- From beginning
  XGROUP CREATE events processors $ MKSTREAM      -- From now on

XREADGROUP GROUP group consumer [COUNT count] [BLOCK milliseconds] [NOACK] STREAMS key [key ...] ID [ID ...]
  XREADGROUP GROUP processors worker-1 COUNT 10 BLOCK 5000 STREAMS events >

XACK key group ID [ID ...]           -- Acknowledge processing
  XACK events processors 1711929600000-0

-- Pending entries (unacknowledged)
XPENDING key group [[IDLE min-idle-time] start end count [consumer]]
  XPENDING events processors - + 10              -- First 10 pending
  XPENDING events processors IDLE 60000 - + 10   -- Idle > 60s

XCLAIM key group consumer min-idle-time ID [ID ...]
  XCLAIM events processors worker-2 60000 1711929600000-0  -- Claim idle message

XAUTOCLAIM key group consumer min-idle-time start [COUNT count]
  XAUTOCLAIM events processors worker-2 60000 0-0 COUNT 10  -- Auto-claim idle

-- Stream management
XLEN key                             -- Stream length
XINFO STREAM key [FULL [COUNT count]]
XINFO GROUPS key                     -- List consumer groups
XINFO CONSUMERS key groupname        -- List consumers in group
XRANGE key start end [COUNT count]   -- Read range
XREVRANGE key end start [COUNT count]
XTRIM key MAXLEN|MINID [=|~] threshold
XDEL key ID [ID ...]                 -- Delete specific entries

-- Pattern: Reliable event processing
XADD orders * order_id "ORD-123" amount "99.99" status "created"
XGROUP CREATE orders order-processors $ MKSTREAM
-- Worker loop:
--   XREADGROUP GROUP order-processors worker-1 COUNT 1 BLOCK 5000 STREAMS orders >
--   Process message...
--   XACK orders order-processors <message-id>
-- Periodic cleanup:
--   XAUTOCLAIM orders order-processors worker-1 300000 0-0
```

---

## Key Management Commands

```
-- Expiration
EXPIRE key seconds [NX|XX|GT|LT]     -- Set TTL in seconds
PEXPIRE key milliseconds [NX|XX|GT|LT]
EXPIREAT key unix-time-seconds       -- Set absolute expiration
PEXPIREAT key unix-time-milliseconds
PERSIST key                           -- Remove expiration
TTL key                               -- Remaining TTL in seconds (-1 = no expire, -2 = doesn't exist)
PTTL key                              -- Remaining TTL in milliseconds
EXPIRETIME key                        -- Absolute Unix timestamp of expiration

-- Key operations
DEL key [key ...]                    -- Delete (blocking)
UNLINK key [key ...]                 -- Delete (non-blocking, async free)
EXISTS key [key ...]                 -- Returns count of existing keys
TYPE key                             -- Returns data type
OBJECT ENCODING key                  -- Internal encoding
OBJECT IDLETIME key                  -- Seconds since last access
OBJECT FREQ key                      -- Access frequency (LFU policy)
RENAME key newkey                    -- Rename (overwrites newkey)
RENAMENX key newkey                  -- Rename only if newkey doesn't exist
COPY source destination [DB db] [REPLACE]
DUMP key                             -- Serialize value
RESTORE key ttl serialized-value     -- Deserialize

-- Key scanning
SCAN cursor [MATCH pattern] [COUNT count] [TYPE type]
  SCAN 0 MATCH "user:*" COUNT 100 TYPE string    -- Find string keys matching pattern
  -- Iterate: use returned cursor until it returns 0

KEYS pattern                         -- Find keys matching pattern (BLOCKING - never in production!)
RANDOMKEY                            -- Return random key
DBSIZE                               -- Total key count in current DB

-- Pattern: Safe key iteration in production
-- ALWAYS use SCAN, never KEYS
-- cursor = 0
-- do:
--   cursor, keys = SCAN cursor MATCH "cache:*" COUNT 100
--   for key in keys: process(key)
-- while cursor != 0
```

---

## Transaction & Scripting Commands

### Transactions (MULTI/EXEC)

```
MULTI                                -- Start transaction
-- ... queue commands ...
EXEC                                 -- Execute all queued commands atomically
DISCARD                              -- Discard queued commands

WATCH key [key ...]                  -- Optimistic locking
UNWATCH                              -- Cancel all watches

-- Pattern: Optimistic locking (CAS)
WATCH balance:user123
val = GET balance:user123
MULTI
SET balance:user123 (val - 100)
EXEC                                 -- Fails if balance:user123 changed since WATCH
```

### Lua Scripting

```
EVAL script numkeys [key ...] [arg ...]
  EVAL "return redis.call('GET', KEYS[1])" 1 mykey

EVALSHA sha1 numkeys [key ...] [arg ...]
  -- Run cached script by SHA1 hash (faster, no script transfer)

SCRIPT LOAD script                   -- Cache script, returns SHA1
SCRIPT EXISTS sha1 [sha1 ...]        -- Check if scripts are cached
SCRIPT FLUSH [ASYNC|SYNC]            -- Clear script cache

-- Pattern: Atomic rate limiter (Lua)
EVAL "
  local key = KEYS[1]
  local limit = tonumber(ARGV[1])
  local window = tonumber(ARGV[2])
  local now = tonumber(ARGV[3])

  redis.call('ZREMRANGEBYSCORE', key, 0, now - window)
  local count = redis.call('ZCARD', key)

  if count < limit then
    redis.call('ZADD', key, now, now .. ':' .. math.random(1000000))
    redis.call('EXPIRE', key, window)
    return 1
  end
  return 0
" 1 ratelimit:user123 100 60 <current_timestamp>

-- Pattern: Atomic counter with ceiling
EVAL "
  local key = KEYS[1]
  local max = tonumber(ARGV[1])
  local current = tonumber(redis.call('GET', key) or '0')
  if current < max then
    return redis.call('INCR', key)
  end
  return -1
" 1 inventory:sku001 100
```

### Functions (Redis 7+)

```
FUNCTION LOAD [REPLACE] "<library code>"
FCALL function numkeys [key ...] [arg ...]
FUNCTION LIST [LIBRARYNAME library] [WITHCODE]
FUNCTION DELETE library
FUNCTION DUMP / FUNCTION RESTORE

-- Library definition example:
-- #!lua name=mylib
-- redis.register_function('my_hgetall', function(keys, args)
--   return redis.call('HGETALL', keys[1])
-- end)
```

---

## Pub/Sub Commands

```
SUBSCRIBE channel [channel ...]      -- Subscribe to channels
UNSUBSCRIBE [channel ...]            -- Unsubscribe
PSUBSCRIBE pattern [pattern ...]     -- Pattern subscribe (e.g., "events.*")
PUNSUBSCRIBE [pattern ...]
PUBLISH channel message              -- Publish to channel (returns subscriber count)
PUBSUB CHANNELS [pattern]            -- List active channels
PUBSUB NUMSUB [channel ...]         -- Subscriber count per channel
PUBSUB NUMPAT                        -- Number of pattern subscriptions

-- Pattern: Event broadcasting
PUBLISH events:orders '{"order_id":"123","status":"shipped"}'
-- Subscriber receives: ["message", "events:orders", "{...}"]

-- Pattern: Cache invalidation across instances
PUBLISH cache:invalidate '{"key":"user:123"}'
-- All app instances subscribed to cache:invalidate clear their local cache

-- Important: Pub/Sub is fire-and-forget. No persistence. No replay.
-- For reliable messaging, use Redis Streams instead.
```

---

## Server & Administration Commands

### Info & Monitoring

```
INFO [section]                       -- Server information
  INFO memory                        -- Memory usage
  INFO stats                         -- General statistics
  INFO replication                   -- Replication info
  INFO clients                       -- Client connections
  INFO keyspace                      -- Database key counts
  INFO commandstats                  -- Command call statistics

MONITOR                              -- Real-time command stream (DEBUG ONLY - massive overhead)
SLOWLOG GET [count]                  -- Recent slow commands
SLOWLOG LEN                          -- Number of slow log entries
SLOWLOG RESET                        -- Clear slow log

CLIENT LIST [TYPE normal|master|replica|pubsub]
CLIENT GETNAME
CLIENT SETNAME name
CLIENT ID
CLIENT KILL [ID id | ADDR ip:port | ...]
CLIENT NO-EVICT ON|OFF              -- Protect client from eviction

LATENCY LATEST                       -- Latest latency samples
LATENCY HISTORY event                -- History for specific event
LATENCY RESET [event ...]
DEBUG SLEEP seconds                  -- DEBUG ONLY
```

### Memory Analysis

```
MEMORY USAGE key [SAMPLES count]     -- Memory used by key (bytes)
MEMORY DOCTOR                        -- Memory usage diagnosis
MEMORY STATS                         -- Detailed memory stats
MEMORY PURGE                         -- Release memory back to OS
MEMORY MALLOC-STATS                  -- Allocator statistics

OBJECT ENCODING key                  -- Internal encoding of value
OBJECT HELP                          -- Available OBJECT subcommands

-- Pattern: Find large keys
-- Use redis-cli --bigkeys for built-in analysis
-- Or SCAN + MEMORY USAGE for custom analysis
```

### Configuration

```
CONFIG GET parameter [parameter ...]
  CONFIG GET maxmemory
  CONFIG GET maxmemory-policy
  CONFIG GET "save"

CONFIG SET parameter value [parameter value ...]
  CONFIG SET maxmemory 2gb
  CONFIG SET maxmemory-policy allkeys-lru
  CONFIG SET slowlog-log-slower-than 10000    -- Log commands > 10ms

CONFIG REWRITE                       -- Write config to file
CONFIG RESETSTAT                     -- Reset statistics

-- Key configuration parameters:
-- maxmemory           — Maximum memory limit
-- maxmemory-policy    — Eviction policy (allkeys-lru, volatile-lru, allkeys-lfu, etc.)
-- save                — RDB snapshot intervals
-- appendonly          — Enable AOF
-- appendfsync         — AOF sync policy (always, everysec, no)
-- hz                  — Background task frequency (default: 10)
-- slowlog-log-slower-than — Slow log threshold (microseconds)
-- timeout             — Client idle timeout (0 = disabled)
-- tcp-keepalive       — TCP keepalive interval
-- maxclients          — Maximum client connections
```

### Persistence

```
BGSAVE                               -- Background RDB snapshot
BGREWRITEAOF                         -- Background AOF rewrite
LASTSAVE                             -- Timestamp of last successful save
DBSIZE                               -- Key count
FLUSHDB [ASYNC]                      -- Delete all keys in current DB
FLUSHALL [ASYNC]                     -- Delete all keys in all DBs
SELECT db                            -- Switch database (0-15)
SWAPDB db1 db2                       -- Swap two databases

-- Persistence strategy:
-- RDB only: Good for backups, faster restart, some data loss acceptable
-- AOF only: Better durability, larger files, slower restart
-- RDB + AOF: Best durability. On restart, AOF is used (more complete).
```

### Replication

```
REPLICAOF host port                  -- Make this server a replica
REPLICAOF NO ONE                     -- Promote to primary
WAIT numreplicas timeout             -- Block until N replicas acknowledge
```

### Cluster

```
CLUSTER INFO                         -- Cluster state
CLUSTER NODES                        -- List all nodes
CLUSTER SLOTS                        -- Slot-to-node mapping
CLUSTER MEET ip port                 -- Add node to cluster
CLUSTER ADDSLOTS slot [slot ...]     -- Assign hash slots
CLUSTER DELSLOTS slot [slot ...]
CLUSTER FAILOVER [FORCE|TAKEOVER]    -- Manual failover
CLUSTER RESET [HARD|SOFT]
CLUSTER KEYSLOT key                  -- Hash slot for key
CLUSTER COUNTKEYSINSLOT slot
CLUSTER GETKEYSINSLOT slot count
CLUSTER REPLICATE node-id            -- Make this node a replica
```

---

## Pipeline Best Practices

```python
# Python pipeline (ioredis/redis-py)
# Without pipeline: N round-trips
# With pipeline: 1 round-trip for N commands

pipe = redis.pipeline(transaction=False)  # No MULTI/EXEC wrapping
for user_id in user_ids:
    pipe.get(f"user:{user_id}")
    pipe.hgetall(f"profile:{user_id}")
results = pipe.execute()

# Transaction pipeline (atomic)
pipe = redis.pipeline(transaction=True)   # Wraps in MULTI/EXEC
pipe.decrby(f"balance:{sender}", amount)
pipe.incrby(f"balance:{receiver}", amount)
pipe.execute()  # Atomic — both succeed or both fail

# Batch size recommendation:
# - 100-1000 commands per pipeline
# - Above 1000: split into chunks to avoid blocking
# - Monitor with SLOWLOG to ensure pipeline isn't too large
```

---

## redis-cli Quick Reference

```bash
# Connect
redis-cli                            # localhost:6379
redis-cli -h host -p port -a password
redis-cli -u redis://user:pass@host:port/db
redis-cli --tls                      # TLS connection

# Useful flags
redis-cli --bigkeys                  # Find largest keys per type
redis-cli --memkeys                  # Find keys using most memory
redis-cli --hotkeys                  # Find most accessed keys (requires LFU)
redis-cli --stat                     # Rolling stats display
redis-cli --latency                  # Continuous latency measurement
redis-cli --latency-history          # Latency over time
redis-cli --latency-dist             # Latency distribution
redis-cli --intrinsic-latency 5      # Measure system latency (5 seconds)
redis-cli --pipe                     # Mass import (Redis protocol)
redis-cli --rdb dump.rdb             # Download RDB snapshot
redis-cli --scan --pattern "user:*"  # Safe key scanning

# One-liner commands
redis-cli PING                       # Test connectivity
redis-cli INFO memory | grep used_memory_human
redis-cli CONFIG GET maxmemory
redis-cli DBSIZE
redis-cli SLOWLOG GET 10
```
