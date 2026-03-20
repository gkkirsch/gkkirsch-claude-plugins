---
name: system-design-interviews
description: >
  System design interview preparation — back-of-envelope calculations,
  common system designs (URL shortener, chat, newsfeed), capacity planning,
  and structured problem-solving frameworks.
  Triggers: "system design interview", "design a url shortener", "design a chat system",
  "back of envelope", "capacity planning", "design twitter", "design instagram",
  "design notification system", "qps calculation", "storage estimation".
  NOT for: Coding interviews, algorithm questions, frontend design.
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# System Design Interview Patterns

## The Framework (4 Steps)

| Step | Time | What To Do |
|------|------|-----------|
| 1. Clarify requirements | 5 min | Functional vs non-functional, scale, constraints |
| 2. Back-of-envelope | 5 min | QPS, storage, bandwidth estimates |
| 3. High-level design | 15 min | Components, data flow, API design |
| 4. Deep dive | 15 min | Scaling, bottlenecks, trade-offs |

## Back-of-Envelope Calculations

### Reference Numbers

| Metric | Value |
|--------|-------|
| 1 day | 86,400 seconds (~100K) |
| 1 month | ~2.5M seconds |
| 1 year | ~30M seconds |
| QPS from DAU | DAU x actions/user / 86,400 |
| Peak QPS | Average QPS x 2-3 |
| SSD read | 0.1 ms |
| Network round trip (same DC) | 0.5 ms |
| HDD seek | 10 ms |
| Network round trip (cross-continent) | 150 ms |

### Storage Sizes

| Data | Size |
|------|------|
| Char | 1 byte |
| Integer | 4 bytes |
| Long/Timestamp | 8 bytes |
| UUID | 16 bytes |
| Short URL (7 chars) | 7 bytes |
| Tweet (140 chars UTF-8) | ~280 bytes |
| Metadata per record | ~500 bytes |
| Small image (thumbnail) | 10-50 KB |
| Medium image | 200 KB - 1 MB |
| Short video (1 min) | 5-10 MB |

### Calculation Template

```
Given: 100M DAU, each user does X action 5 times/day

QPS:
  Total requests/day = 100M x 5 = 500M
  Average QPS = 500M / 86,400 = ~5,800
  Peak QPS = 5,800 x 3 = ~17,400

Storage (5 years):
  Per record: 500 bytes (metadata) + 200KB (media)
  New records/day: 500M
  Daily storage: 500M x 500B = 250 GB (metadata)
                 500M x 0.05 (10% have media) x 200KB = 5 TB (media)
  5-year total: 250GB x 365 x 5 = ~450 TB (metadata)
                5TB x 365 x 5 = ~9 PB (media)

Bandwidth:
  Incoming: 250 GB/day = ~3 MB/s (metadata only)
  Outgoing (read-heavy, 10:1): 2.5 TB/day = ~30 MB/s
```

## Design: URL Shortener

### Requirements
- Shorten URL: `https://example.com/very-long-path` → `https://short.ly/abc1234`
- Redirect short URL to original
- 100M new URLs/day, 10:1 read:write ratio

### Back-of-Envelope
```
Write QPS: 100M / 86,400 = ~1,160
Read QPS: 1,160 x 10 = ~11,600
Storage (5 years): 100M x 365 x 5 x 500B = ~90 TB
```

### High-Level Design

```
┌────────┐     ┌──────────┐     ┌───────────┐     ┌──────────┐
│ Client │────▶│   API    │────▶│  Service  │────▶│ Database │
│        │◀────│ Gateway  │◀────│  Layer    │◀────│(Postgres)│
└────────┘     └──────────┘     └───────────┘     └──────────┘
                                      │
                                ┌─────▼─────┐
                                │   Cache   │
                                │  (Redis)  │
                                └───────────┘
```

### Key Design Decisions

```typescript
// URL ID generation: Base62 encoding
function encode(id: bigint): string {
  const chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz";
  let result = "";
  while (id > 0) {
    result = chars[Number(id % 62n)] + result;
    id = id / 62n;
  }
  return result.padStart(7, "0"); // 7 chars = 62^7 = 3.5 trillion
}

// API
// POST /api/shorten { url: "https://..." } → { shortUrl: "https://short.ly/abc1234" }
// GET /:shortCode → 301 redirect to original URL

// Database schema
// urls: id (bigint PK), short_code (varchar(7) UNIQUE INDEX), original_url (text),
//       created_at (timestamp), expires_at (timestamp), user_id (optional)

// Caching: cache short_code → original_url in Redis (90%+ hit rate expected)
```

## Design: Chat System

### Requirements
- 1-on-1 and group chat (up to 500 members)
- Online/offline status
- Message history
- 50M DAU, each sends 40 messages/day

### Back-of-Envelope
```
Message QPS: 50M x 40 / 86,400 = ~23,000
Peak: ~70,000 QPS
Storage/day: 2B messages x 200B = ~400 GB
Concurrent WebSocket connections: ~5M (10% of DAU online)
```

### High-Level Design

```
┌────────┐     ┌──────────┐     ┌───────────┐     ┌──────────┐
│ Client │◀───▶│ WebSocket│◀───▶│  Chat     │◀───▶│ Message  │
│        │     │ Gateway  │     │  Service  │     │   Store  │
└────────┘     └──────────┘     └───────────┘     │(Cassandra│
                    │                │             └──────────┘
               ┌────▼────┐     ┌────▼────┐
               │ Presence │     │ Message │
               │ Service  │     │  Queue  │
               │ (Redis)  │     │ (Kafka) │
               └──────────┘     └─────────┘
```

### Key Design Decisions

```
Message ID: Snowflake ID (timestamp + machine + sequence)
  - Time-ordered for chronological display
  - Unique across distributed nodes
  - 64-bit: 41 bits timestamp + 10 bits machine + 13 bits sequence

Storage: Cassandra (wide-column)
  - Partition key: channel_id
  - Clustering key: message_id (sorted)
  - Supports efficient range queries for message history

Delivery: WebSocket for real-time, push notification for offline
  - Fan-out on write for small groups (< 500)
  - Fan-out on read for large channels/broadcasts

Presence: Redis with TTL
  - SET user:{id}:online "" EX 60
  - Heartbeat every 30 seconds refreshes TTL
  - Subscribe to keyspace notifications for status changes
```

## Design: News Feed

### Requirements
- Users follow other users
- Feed shows posts from followed users, ranked
- 300M MAU, 50M DAU
- Average user follows 200 people

### High-Level Design

```
       Publishing                           Reading
┌────────┐     ┌──────────┐        ┌──────────┐     ┌────────┐
│  User  │────▶│  Post    │        │   Feed   │────▶│  User  │
│  posts │     │ Service  │        │ Service  │     │  reads │
└────────┘     └─────┬────┘        └────┬─────┘     └────────┘
                     │                  │
               ┌─────▼─────┐     ┌─────▼─────┐
               │  Fan-out  │────▶│   Feed    │
               │  Service  │     │   Cache   │
               └───────────┘     │  (Redis)  │
                                 └───────────┘
```

### Key Design Decisions

```
Fan-out approach (hybrid):
  - Normal users (< 5K followers): fan-out on WRITE
    Post → write to all followers' feed caches
    Pre-computed feeds = fast reads
  - Celebrities (> 5K followers): fan-out on READ
    Post stored once, merged into feed at read time
    Avoids writing to millions of feed caches

Feed cache structure (Redis sorted set):
  ZADD feed:{userId} {timestamp} {postId}
  ZREVRANGE feed:{userId} 0 19  // latest 20 posts

Ranking:
  score = base_score + recency_bonus + engagement_bonus
  base_score = affinity(author, viewer) x type_weight
  recency_bonus = 1 / (age_hours + 1)
  engagement_bonus = log(likes + comments + shares)
```

## Design: Notification System

### Requirements
- Push notifications, email, SMS, in-app
- 100M notifications/day across all channels
- Prioritized delivery (urgent vs batched)

### Architecture

```
┌──────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐
│ Event    │────▶│ Priority │────▶│ Channel  │────▶│ Provider │
│ Producer │     │  Queue   │     │ Router   │     │  (APNS,  │
└──────────┘     │ (Kafka)  │     │          │     │  FCM,    │
                 └──────────┘     └──────────┘     │  SES)    │
                                       │           └──────────┘
                                 ┌─────▼─────┐
                                 │ User Pref │
                                 │  Service  │
                                 └───────────┘
```

### Key Decisions

```
Priority levels:
  P0 (immediate): security alerts, 2FA codes
  P1 (fast): order confirmations, DMs
  P2 (batched): social updates, marketing
  P3 (digest): weekly summaries

Deduplication:
  - idempotency key per notification
  - Redis SET with 24h TTL for recent notification hashes
  - Check before send: SADD dedup:{hash} → 0 means duplicate

Rate limiting per user:
  - Max 5 push notifications per hour
  - Max 3 emails per day
  - Respect quiet hours (user preference)

Template system:
  - Notification templates stored in DB
  - Variable interpolation: "{{user.name}} liked your post"
  - Localization per user's preferred language
```

## Scaling Checklist

| Bottleneck | Solution |
|-----------|----------|
| Single database | Read replicas, sharding, caching |
| Single server | Horizontal scaling + load balancer |
| Hot partition | Better partition key, consistent hashing |
| Slow queries | Indexes, denormalization, materialized views |
| Large payloads | CDN, compression, pagination |
| Chatty services | Batch APIs, caching, circuit breakers |
| Single point of failure | Redundancy, multi-AZ, failover |
| Data growth | TTL policies, archiving, tiered storage |

## Gotchas

1. **Don't jump to solutions.** Spend 5 minutes clarifying requirements. The difference between 1K QPS and 100K QPS changes the entire architecture. Ask about scale, consistency needs, and availability requirements.

2. **State your assumptions explicitly.** "I'll assume 100M DAU" is better than silently using a number. Interviewers want to see your reasoning, not just your answer.

3. **Don't over-engineer for scale you don't need.** If the system serves 1K users, a single PostgreSQL instance with Redis cache handles it. Only introduce sharding, message queues, and microservices when the math demands it.

4. **Trade-offs are the whole point.** Every decision has a downside. "I chose eventual consistency because strong consistency at this scale would require distributed transactions, adding 50ms latency to every write" is what interviewers want to hear.

5. **Fan-out on write vs read is the classic trade-off.** Write-heavy (pre-compute feeds) = fast reads, expensive writes. Read-heavy (compute at read time) = slow reads, cheap writes. Most systems use a hybrid based on user follower count.

6. **Monitoring and observability aren't optional.** Mention logging, metrics, alerting, and tracing. A system you can't monitor is a system you can't debug in production.
