---
name: redis-patterns
description: >
  Implement common Redis patterns including session management, leaderboards, pub/sub messaging,
  job queues, and real-time analytics. Generates production-ready Redis code with proper
  error handling, connection pooling, and pipeline optimization.
  Triggers: "Redis session", "leaderboard", "Redis queue", "pub/sub", "Redis analytics",
  "Redis counter", "distributed lock", "rate limiter".
  NOT for: Redis cluster setup (use redis-architect agent), caching strategy (use cache-strategist agent).
version: 1.0.0
argument-hint: "[pattern-name]"
allowed-tools: Read, Grep, Glob, Bash, Write, Edit
model: sonnet
---

# Redis Patterns

Implement production-ready Redis patterns. Each pattern includes complete code with error handling, connection management, and operational considerations.

## Session Management

### Express.js + Redis Sessions

```javascript
const express = require("express");
const session = require("express-session");
const RedisStore = require("connect-redis").default;
const { createClient } = require("redis");

const app = express();

// Redis client with reconnection
const redisClient = createClient({
  url: process.env.REDIS_URL || "redis://localhost:6379",
  socket: {
    reconnectStrategy: (retries) => Math.min(retries * 100, 5000),
    connectTimeout: 10000,
  },
});

redisClient.on("error", (err) => console.error("Redis error:", err));
redisClient.on("reconnecting", () => console.log("Redis reconnecting..."));

await redisClient.connect();

app.use(
  session({
    store: new RedisStore({
      client: redisClient,
      prefix: "sess:",
      ttl: 86400, // 24 hours
      disableTouch: false, // Extend TTL on access
    }),
    secret: process.env.SESSION_SECRET,
    resave: false,
    saveUninitialized: false,
    name: "sid",
    cookie: {
      secure: process.env.NODE_ENV === "production",
      httpOnly: true,
      maxAge: 24 * 60 * 60 * 1000, // 24 hours
      sameSite: "lax",
    },
  })
);
```

### Python Flask + Redis Sessions

```python
from flask import Flask, session
from flask_session import Session
import redis

app = Flask(__name__)

app.config.update(
    SESSION_TYPE="redis",
    SESSION_REDIS=redis.Redis(
        host="localhost",
        port=6379,
        socket_timeout=5,
        socket_connect_timeout=5,
        retry_on_timeout=True,
    ),
    SESSION_PERMANENT=True,
    PERMANENT_SESSION_LIFETIME=86400,
    SESSION_KEY_PREFIX="sess:",
    SESSION_USE_SIGNER=True,
    SECRET_KEY=os.environ["SECRET_KEY"],
)

Session(app)
```

## Leaderboard

```python
class Leaderboard:
    """Redis sorted set leaderboard with ranking, pagination, and nearby players."""

    def __init__(self, redis_client, name: str):
        self.redis = redis_client
        self.key = f"leaderboard:{name}"

    def add_score(self, player_id: str, score: float):
        """Add or update a player's score."""
        self.redis.zadd(self.key, {player_id: score})

    def increment_score(self, player_id: str, delta: float) -> float:
        """Atomically increment score."""
        return self.redis.zincrby(self.key, delta, player_id)

    def get_rank(self, player_id: str) -> Optional[int]:
        """Get player's rank (0-indexed, highest score = rank 0)."""
        rank = self.redis.zrevrank(self.key, player_id)
        return rank + 1 if rank is not None else None

    def get_score(self, player_id: str) -> Optional[float]:
        return self.redis.zscore(self.key, player_id)

    def get_top(self, count: int = 10) -> list[dict]:
        """Get top N players with scores."""
        results = self.redis.zrevrange(self.key, 0, count - 1, withscores=True)
        return [
            {"rank": i + 1, "player_id": pid, "score": score}
            for i, (pid, score) in enumerate(results)
        ]

    def get_page(self, page: int, page_size: int = 20) -> list[dict]:
        """Paginated leaderboard."""
        start = (page - 1) * page_size
        end = start + page_size - 1
        results = self.redis.zrevrange(self.key, start, end, withscores=True)
        return [
            {"rank": start + i + 1, "player_id": pid, "score": score}
            for i, (pid, score) in enumerate(results)
        ]

    def get_around_player(self, player_id: str, count: int = 5) -> list[dict]:
        """Get players ranked around a specific player."""
        rank = self.redis.zrevrank(self.key, player_id)
        if rank is None:
            return []

        start = max(0, rank - count)
        end = rank + count
        results = self.redis.zrevrange(self.key, start, end, withscores=True)
        return [
            {"rank": start + i + 1, "player_id": pid, "score": score}
            for i, (pid, score) in enumerate(results)
        ]

    def total_players(self) -> int:
        return self.redis.zcard(self.key)

    def remove_player(self, player_id: str):
        self.redis.zrem(self.key, player_id)


# Usage
lb = Leaderboard(r, "weekly")
lb.add_score("player:alice", 1500)
lb.increment_score("player:alice", 100)
top_10 = lb.get_top(10)
alice_rank = lb.get_rank("player:alice")
nearby = lb.get_around_player("player:alice", count=3)
```

## Job Queue

```python
import uuid
import time

class RedisJobQueue:
    """Reliable job queue with retry, dead letter, and priorities."""

    def __init__(self, redis_client, queue_name: str):
        self.redis = redis_client
        self.queue = f"queue:{queue_name}"
        self.processing = f"queue:{queue_name}:processing"
        self.dead_letter = f"queue:{queue_name}:dead"
        self.retry_limit = 3

    def enqueue(self, job_data: dict, priority: int = 0) -> str:
        """Add a job to the queue."""
        job_id = str(uuid.uuid4())
        job = {
            "id": job_id,
            "data": json.dumps(job_data),
            "created_at": time.time(),
            "attempts": 0,
            "priority": priority,
        }

        if priority > 0:
            # High priority: push to front
            self.redis.lpush(self.queue, json.dumps(job))
        else:
            # Normal priority: push to back
            self.redis.rpush(self.queue, json.dumps(job))

        return job_id

    def dequeue(self, timeout: int = 5) -> Optional[dict]:
        """Reliably dequeue a job (BRPOPLPUSH pattern)."""
        result = self.redis.brpoplpush(
            self.queue,
            self.processing,
            timeout=timeout,
        )

        if result:
            job = json.loads(result)
            job["attempts"] += 1
            job["started_at"] = time.time()
            # Update the processing entry
            self.redis.lrem(self.processing, 1, result)
            self.redis.lpush(self.processing, json.dumps(job))
            return job

        return None

    def complete(self, job: dict):
        """Mark job as completed."""
        self.redis.lrem(self.processing, 1, json.dumps(job))

    def fail(self, job: dict, error: str):
        """Handle job failure with retry logic."""
        self.redis.lrem(self.processing, 1, json.dumps(job))

        if job["attempts"] >= self.retry_limit:
            # Move to dead letter queue
            job["error"] = error
            job["dead_at"] = time.time()
            self.redis.lpush(self.dead_letter, json.dumps(job))
        else:
            # Re-enqueue with exponential backoff
            delay = 2 ** job["attempts"]
            job["retry_after"] = time.time() + delay
            self.redis.rpush(self.queue, json.dumps(job))

    def queue_length(self) -> int:
        return self.redis.llen(self.queue)

    def processing_count(self) -> int:
        return self.redis.llen(self.processing)

    def dead_letter_count(self) -> int:
        return self.redis.llen(self.dead_letter)

    def recover_stale(self, timeout: int = 300):
        """Recover jobs stuck in processing for too long."""
        processing_jobs = self.redis.lrange(self.processing, 0, -1)
        now = time.time()

        for raw in processing_jobs:
            job = json.loads(raw)
            if now - job.get("started_at", 0) > timeout:
                self.redis.lrem(self.processing, 1, raw)
                self.redis.rpush(self.queue, raw)


# Worker
queue = RedisJobQueue(r, "emails")

while True:
    job = queue.dequeue(timeout=5)
    if job is None:
        continue

    try:
        send_email(json.loads(job["data"]))
        queue.complete(job)
    except Exception as e:
        queue.fail(job, str(e))
```

## Real-Time Analytics

### Counters and Time Series

```python
class RealtimeAnalytics:
    """Real-time event counting and time series using Redis."""

    def __init__(self, redis_client):
        self.redis = redis_client

    def track_event(self, event_type: str, dimensions: dict = None):
        """Track an event with optional dimensions."""
        now = datetime.utcnow()
        pipe = self.redis.pipeline()

        # Total counter
        pipe.incr(f"events:{event_type}:total")

        # Time-bucketed counters
        minute_key = f"events:{event_type}:min:{now.strftime('%Y%m%d%H%M')}"
        hour_key = f"events:{event_type}:hour:{now.strftime('%Y%m%d%H')}"
        day_key = f"events:{event_type}:day:{now.strftime('%Y%m%d')}"

        pipe.incr(minute_key)
        pipe.expire(minute_key, 3600)       # Keep 1 hour of minute data
        pipe.incr(hour_key)
        pipe.expire(hour_key, 86400 * 7)    # Keep 7 days of hourly data
        pipe.incr(day_key)
        pipe.expire(day_key, 86400 * 90)    # Keep 90 days of daily data

        # Dimension counters
        if dimensions:
            for dim_name, dim_value in dimensions.items():
                dim_key = f"events:{event_type}:dim:{dim_name}:{dim_value}:{now.strftime('%Y%m%d')}"
                pipe.incr(dim_key)
                pipe.expire(dim_key, 86400 * 30)

        # Unique visitors (HyperLogLog)
        if "user_id" in (dimensions or {}):
            hll_key = f"events:{event_type}:uniq:{now.strftime('%Y%m%d')}"
            pipe.pfadd(hll_key, dimensions["user_id"])
            pipe.expire(hll_key, 86400 * 90)

        pipe.execute()

    def get_counts(self, event_type: str, granularity: str, periods: int) -> list[dict]:
        """Get time series counts."""
        now = datetime.utcnow()
        results = []

        for i in range(periods):
            if granularity == "minute":
                ts = now - timedelta(minutes=i)
                key = f"events:{event_type}:min:{ts.strftime('%Y%m%d%H%M')}"
            elif granularity == "hour":
                ts = now - timedelta(hours=i)
                key = f"events:{event_type}:hour:{ts.strftime('%Y%m%d%H')}"
            elif granularity == "day":
                ts = now - timedelta(days=i)
                key = f"events:{event_type}:day:{ts.strftime('%Y%m%d')}"

            count = self.redis.get(key)
            results.append({
                "timestamp": ts.isoformat(),
                "count": int(count or 0),
            })

        return list(reversed(results))

    def get_unique_count(self, event_type: str, date_str: str) -> int:
        """Get approximate unique visitor count."""
        return self.redis.pfcount(f"events:{event_type}:uniq:{date_str}")


# Usage
analytics = RealtimeAnalytics(r)

# Track events
analytics.track_event("page_view", {
    "user_id": "u123",
    "page": "/products",
    "source": "google",
})

# Query
last_60_minutes = analytics.get_counts("page_view", "minute", 60)
today_uniques = analytics.get_unique_count("page_view", "20240115")
```

## Pub/Sub Patterns

### Event Bus

```python
import threading

class RedisEventBus:
    """Pub/Sub event bus for real-time notifications."""

    def __init__(self, redis_url: str):
        self.publisher = redis.Redis.from_url(redis_url)
        self.subscriber = redis.Redis.from_url(redis_url)
        self.handlers = {}

    def publish(self, channel: str, event: dict):
        """Publish event to a channel."""
        self.publisher.publish(channel, json.dumps(event))

    def subscribe(self, channel: str, handler):
        """Register handler for a channel."""
        if channel not in self.handlers:
            self.handlers[channel] = []
        self.handlers[channel].append(handler)

    def start(self):
        """Start listening for events in background thread."""
        pubsub = self.subscriber.pubsub()
        pubsub.subscribe(*self.handlers.keys())

        def listen():
            for message in pubsub.listen():
                if message["type"] == "message":
                    channel = message["channel"]
                    if isinstance(channel, bytes):
                        channel = channel.decode()
                    data = json.loads(message["data"])
                    for handler in self.handlers.get(channel, []):
                        try:
                            handler(data)
                        except Exception as e:
                            logger.error(f"Handler error on {channel}: {e}")

        thread = threading.Thread(target=listen, daemon=True)
        thread.start()
        return thread


# Usage
bus = RedisEventBus("redis://localhost:6379")

bus.subscribe("orders", lambda event: print(f"New order: {event['order_id']}"))
bus.subscribe("orders", lambda event: send_confirmation_email(event))
bus.subscribe("inventory", lambda event: update_stock_display(event))

bus.start()

# Publish from anywhere
bus.publish("orders", {"order_id": "123", "total": 99.99})
```

## Pipeline Optimization

```python
# BAD: 100 round trips
for user_id in user_ids:
    r.get(f"user:{user_id}")

# GOOD: 1 round trip
pipe = r.pipeline()
for user_id in user_ids:
    pipe.get(f"user:{user_id}")
results = pipe.execute()

# BETTER: pipeline with transaction (atomic)
pipe = r.pipeline(transaction=True)
pipe.multi()
pipe.decrby("inventory:product1", quantity)
pipe.incrby("sales:product1", quantity)
pipe.rpush("order_log", json.dumps(order))
pipe.execute()  # All or nothing
```

## Gotchas

- Always set TTL on cached values — without it, data stays forever and memory fills up
- Use `UNLINK` instead of `DEL` for large keys (async deletion, non-blocking)
- Pub/Sub messages are fire-and-forget — use Streams if you need persistence/replay
- Pipeline commands are not atomic by default — wrap in `MULTI/EXEC` if needed
- Don't use `KEYS *` in production — use `SCAN` instead (non-blocking iteration)
- Redis is single-threaded for commands — one slow command blocks everything
- Connection pooling is essential — don't create a new connection per request
- Monitor `evicted_keys` — if it's growing, you need more memory or shorter TTLs
