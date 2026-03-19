# Data Modeling Patterns Reference

Cross-database data modeling patterns for real-world applications. Covers document vs relational
modeling decisions, event sourcing, CQRS, time-series data, graph relationships, multi-tenant
isolation, caching layers, and MongoDB-specific schema patterns.

## Document vs Relational: Decision Framework

### When to Use Relational (PostgreSQL, MySQL, SQLite)

```
Choose relational when:
1. Data has clear, stable relationships (orders → items → products)
2. Strong consistency is required (financial transactions, inventory)
3. Complex queries with JOINs across multiple entities
4. ACID transactions across multiple tables
5. Data integrity through foreign keys and constraints
6. Reporting and analytics with aggregations
7. Schema is well-known upfront and doesn't change often
8. Multiple applications access the same database

Red flags for relational:
- EAV (Entity-Attribute-Value) patterns emerging
- JSON columns used for most data
- Schema changes happen weekly
- Single-entity queries dominate (no JOINs needed)
```

### When to Use Document (MongoDB)

```
Choose document when:
1. Schema varies significantly between documents
2. Hierarchical/nested data is the natural model
3. Read patterns are single-document focused
4. Rapid schema evolution during development
5. Denormalized data is acceptable (embed related data)
6. Horizontal scaling is a primary requirement
7. Real-time analytics with aggregation pipeline
8. Content management with flexible structures

Red flags for document:
- Many cross-collection lookups ($lookup)
- Need for multi-document ACID transactions frequently
- Complex referential integrity requirements
- Many-to-many relationships dominating the model
```

### Hybrid Approaches

```
PostgreSQL + JSONB:
- Relational tables for core structured data
- JSONB columns for flexible metadata/attributes
- Best of both: foreign keys + schema flexibility
- GIN indexes for JSONB queries

PostgreSQL + Redis:
- PostgreSQL for persistent, transactional data
- Redis for caching, sessions, real-time features
- Common pattern in web applications

MongoDB + Redis:
- MongoDB for persistent document storage
- Redis for caching hot data, real-time analytics
- Common in high-throughput applications
```

## Event Sourcing

### Core Concept

Instead of storing current state, store the complete history of events that led to the
current state. The current state is derived by replaying events.

```
Traditional (state-based):
  UPDATE accounts SET balance = 950 WHERE id = 1;
  -- Previous balance is lost

Event sourcing:
  INSERT INTO events (stream_id, type, data)
  VALUES ('account:1', 'MoneyWithdrawn', '{"amount": 50}');
  -- Every change is preserved forever
```

### Event Store Schema

```sql
-- PostgreSQL event store
CREATE TABLE event_store (
    -- Event identity
    id BIGSERIAL PRIMARY KEY,
    stream_id TEXT NOT NULL,           -- Aggregate identifier (e.g., 'order:42')
    stream_type TEXT NOT NULL,         -- Aggregate type (e.g., 'Order')
    event_type TEXT NOT NULL,          -- Event type (e.g., 'OrderPlaced')
    version INT NOT NULL,              -- Sequence within stream (optimistic concurrency)

    -- Event data
    data JSONB NOT NULL,               -- Event payload
    metadata JSONB NOT NULL DEFAULT '{}', -- Correlation ID, causation ID, user context

    -- Timestamp
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Optimistic concurrency: prevent concurrent writes to same stream
    UNIQUE (stream_id, version)
);

-- Indexes
CREATE INDEX idx_events_stream ON event_store(stream_id, version);
CREATE INDEX idx_events_type ON event_store(event_type);
CREATE INDEX idx_events_created ON event_store(created_at);
CREATE INDEX idx_events_correlation ON event_store((metadata->>'correlationId'));
```

### Event Types and Schemas

```sql
-- Example: Order aggregate events

-- OrderPlaced
INSERT INTO event_store (stream_id, stream_type, event_type, version, data, metadata)
VALUES (
    'order:42', 'Order', 'OrderPlaced', 1,
    '{
        "orderId": "42",
        "customerId": "17",
        "items": [
            {"productId": "101", "quantity": 2, "unitPrice": 29.99},
            {"productId": "205", "quantity": 1, "unitPrice": 49.99}
        ],
        "shippingAddress": {"street": "123 Main St", "city": "Springfield"}
    }',
    '{"correlationId": "req-abc-123", "userId": "17", "timestamp": "2024-01-15T10:30:00Z"}'
);

-- OrderItemAdded
INSERT INTO event_store (stream_id, stream_type, event_type, version, data)
VALUES ('order:42', 'Order', 'OrderItemAdded', 2,
    '{"productId": "310", "quantity": 1, "unitPrice": 19.99}');

-- OrderShipped
INSERT INTO event_store (stream_id, stream_type, event_type, version, data)
VALUES ('order:42', 'Order', 'OrderShipped', 3,
    '{"trackingNumber": "1Z999AA10123456784", "carrier": "UPS"}');
```

### Snapshots (Performance Optimization)

```sql
-- Rebuild state from events
-- For an order with 100 events, replaying all 100 is slow
-- Take snapshots periodically to start replay from a known state

CREATE TABLE event_snapshots (
    stream_id TEXT NOT NULL,
    stream_type TEXT NOT NULL,
    version INT NOT NULL,              -- Event version this snapshot represents
    state JSONB NOT NULL,              -- Full aggregate state
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (stream_id)
);

-- Rebuild current state:
-- 1. Load latest snapshot (if exists)
-- 2. Load events after snapshot version
-- 3. Apply events to snapshot state

-- Example: Rebuild order state
WITH latest_snapshot AS (
    SELECT state, version FROM event_snapshots WHERE stream_id = 'order:42'
),
new_events AS (
    SELECT * FROM event_store
    WHERE stream_id = 'order:42'
      AND version > COALESCE((SELECT version FROM latest_snapshot), 0)
    ORDER BY version
)
SELECT * FROM new_events;
-- Application code applies these events to the snapshot state
```

### Projections (Read Models)

```sql
-- Build read-optimized views from events

-- Real-time projection: Maintained by event handlers
CREATE TABLE order_summaries (
    order_id TEXT PRIMARY KEY,
    customer_id TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'placed',
    item_count INT NOT NULL DEFAULT 0,
    total DECIMAL(12, 2) NOT NULL DEFAULT 0,
    placed_at TIMESTAMPTZ,
    shipped_at TIMESTAMPTZ,
    delivered_at TIMESTAMPTZ,
    last_updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Event handler (application code) processes events and updates projections:
-- ON OrderPlaced: INSERT into order_summaries
-- ON OrderItemAdded: UPDATE item_count and total
-- ON OrderShipped: UPDATE status and shipped_at
-- ON OrderDelivered: UPDATE status and delivered_at

-- Projection can be rebuilt from scratch by replaying all events
-- This is powerful: you can add new read models at any time by replaying history
```

## CQRS (Command Query Responsibility Segregation)

### Concept

Separate write models (commands) from read models (queries). Each is optimized for its purpose.

```
Command Side (Write):
- Normalized schema
- Optimized for writes and consistency
- Event store or traditional tables
- Strong validation and business rules

Query Side (Read):
- Denormalized schema
- Optimized for specific query patterns
- May use different database technology
- Eventually consistent with write side

Synchronization:
- Event-driven: Write side publishes events, read side subscribes
- CDC (Change Data Capture): Database triggers or log reading
- Polling: Periodic refresh (simplest, least real-time)
```

### CQRS Schema Example

```sql
-- WRITE MODEL: Normalized, optimized for consistency
CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    customer_id INT NOT NULL REFERENCES customers(id),
    status TEXT NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE order_items (
    id SERIAL PRIMARY KEY,
    order_id INT NOT NULL REFERENCES orders(id),
    product_id INT NOT NULL REFERENCES products(id),
    quantity INT NOT NULL CHECK (quantity > 0),
    unit_price DECIMAL(10, 2) NOT NULL
);

-- READ MODEL: Denormalized, optimized for dashboard query
CREATE TABLE order_dashboard_view (
    order_id INT PRIMARY KEY,
    customer_name TEXT NOT NULL,
    customer_email TEXT NOT NULL,
    status TEXT NOT NULL,
    item_count INT NOT NULL,
    total DECIMAL(12, 2) NOT NULL,
    product_names TEXT[] NOT NULL,    -- Array of product names
    placed_at TIMESTAMPTZ NOT NULL,
    last_status_change TIMESTAMPTZ NOT NULL
);

-- Single query, no JOINs, covers entire dashboard row
SELECT * FROM order_dashboard_view
WHERE status = 'shipped'
ORDER BY placed_at DESC
LIMIT 50;
```

### CQRS Synchronization Patterns

```sql
-- Pattern 1: Trigger-based sync (PostgreSQL)
CREATE OR REPLACE FUNCTION sync_order_dashboard()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO order_dashboard_view (
        order_id, customer_name, customer_email, status,
        item_count, total, product_names, placed_at, last_status_change
    )
    SELECT
        o.id,
        c.name,
        c.email,
        o.status,
        COUNT(oi.id),
        SUM(oi.unit_price * oi.quantity),
        ARRAY_AGG(DISTINCT p.name),
        o.created_at,
        o.updated_at
    FROM orders o
    JOIN customers c ON c.id = o.customer_id
    JOIN order_items oi ON oi.order_id = o.id
    JOIN products p ON p.id = oi.product_id
    WHERE o.id = NEW.id
    GROUP BY o.id, c.name, c.email, o.status, o.created_at, o.updated_at
    ON CONFLICT (order_id) DO UPDATE SET
        status = EXCLUDED.status,
        item_count = EXCLUDED.item_count,
        total = EXCLUDED.total,
        product_names = EXCLUDED.product_names,
        last_status_change = EXCLUDED.last_status_change;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Pattern 2: Materialized view (simpler, batch refresh)
CREATE MATERIALIZED VIEW order_dashboard_mv AS
SELECT
    o.id AS order_id,
    c.name AS customer_name,
    c.email AS customer_email,
    o.status,
    COUNT(oi.id) AS item_count,
    SUM(oi.unit_price * oi.quantity) AS total,
    ARRAY_AGG(DISTINCT p.name) AS product_names,
    o.created_at AS placed_at,
    o.updated_at AS last_status_change
FROM orders o
JOIN customers c ON c.id = o.customer_id
JOIN order_items oi ON oi.order_id = o.id
JOIN products p ON p.id = oi.product_id
GROUP BY o.id, c.name, c.email, o.status, o.created_at, o.updated_at;

REFRESH MATERIALIZED VIEW CONCURRENTLY order_dashboard_mv;
```

## Time-Series Data

### Schema Design for Time-Series

```sql
-- Standard time-series table with partitioning
CREATE TABLE metrics (
    time TIMESTAMPTZ NOT NULL,
    metric_name TEXT NOT NULL,
    value DOUBLE PRECISION NOT NULL,
    tags JSONB NOT NULL DEFAULT '{}'
) PARTITION BY RANGE (time);

-- Create monthly partitions
CREATE TABLE metrics_2024_01 PARTITION OF metrics
    FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');
CREATE TABLE metrics_2024_02 PARTITION OF metrics
    FOR VALUES FROM ('2024-02-01') TO ('2024-03-01');
-- ... continue for each month

-- BRIN index (very efficient for time-ordered data)
CREATE INDEX idx_metrics_time ON metrics USING brin (time);

-- B-tree index for specific metric lookup
CREATE INDEX idx_metrics_name_time ON metrics(metric_name, time DESC);

-- TimescaleDB hypertable (automatic partitioning)
-- CREATE EXTENSION timescaledb;
-- SELECT create_hypertable('metrics', 'time', chunk_time_interval => INTERVAL '1 day');
```

### Time-Series Query Patterns

```sql
-- Downsampling: Aggregate to coarser time intervals
SELECT
    time_bucket('1 hour', time) AS bucket,  -- TimescaleDB function
    -- Or: date_trunc('hour', time) AS bucket,  -- Standard PostgreSQL
    metric_name,
    AVG(value) AS avg_value,
    MIN(value) AS min_value,
    MAX(value) AS max_value,
    COUNT(*) AS sample_count
FROM metrics
WHERE time >= NOW() - INTERVAL '24 hours'
  AND metric_name = 'cpu_usage'
GROUP BY bucket, metric_name
ORDER BY bucket;

-- Gap filling: Include empty time buckets
WITH time_series AS (
    SELECT generate_series(
        date_trunc('hour', NOW() - INTERVAL '24 hours'),
        date_trunc('hour', NOW()),
        '1 hour'
    ) AS bucket
)
SELECT
    ts.bucket,
    COALESCE(AVG(m.value), 0) AS avg_value
FROM time_series ts
LEFT JOIN metrics m ON date_trunc('hour', m.time) = ts.bucket
    AND m.metric_name = 'cpu_usage'
GROUP BY ts.bucket
ORDER BY ts.bucket;

-- Rolling window: Moving average
SELECT
    time,
    value,
    AVG(value) OVER (
        ORDER BY time
        ROWS BETWEEN 11 PRECEDING AND CURRENT ROW
    ) AS moving_avg_12
FROM metrics
WHERE metric_name = 'temperature'
  AND time >= NOW() - INTERVAL '7 days'
ORDER BY time;

-- Retention: Delete old data
DELETE FROM metrics WHERE time < NOW() - INTERVAL '90 days';
-- Or with partitioning: DROP TABLE metrics_2023_01;
```

### MongoDB Time-Series Collections (5.0+)

```javascript
// Create time-series collection
db.createCollection("metrics", {
  timeseries: {
    timeField: "timestamp",
    metaField: "metadata",
    granularity: "seconds"  // seconds, minutes, hours
  },
  expireAfterSeconds: 7776000  // 90 days TTL
});

// Insert time-series data
db.metrics.insertMany([
  {
    timestamp: new Date(),
    metadata: { sensor: "temp-01", location: "server-room" },
    value: 23.5
  },
  {
    timestamp: new Date(),
    metadata: { sensor: "humidity-01", location: "server-room" },
    value: 45.2
  }
]);

// Aggregation on time-series
db.metrics.aggregate([
  { $match: {
    "metadata.sensor": "temp-01",
    timestamp: { $gte: new Date(Date.now() - 86400000) }
  }},
  { $group: {
    _id: {
      $dateTrunc: { date: "$timestamp", unit: "hour" }
    },
    avgValue: { $avg: "$value" },
    minValue: { $min: "$value" },
    maxValue: { $max: "$value" },
    count: { $sum: 1 }
  }},
  { $sort: { "_id": 1 } }
]);
```

## Graph Relationships in Relational Databases

### Adjacency List (Simple Graph)

```sql
-- Social network: followers
CREATE TABLE follows (
    follower_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    followed_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (follower_id, followed_id),
    CHECK (follower_id != followed_id)
);

-- Bidirectional friendship
CREATE TABLE friendships (
    user_a INT NOT NULL REFERENCES users(id),
    user_b INT NOT NULL REFERENCES users(id),
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'blocked')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_a, user_b),
    CHECK (user_a < user_b)  -- Canonical order prevents duplicates
);

-- Find mutual friends
SELECT f1.followed_id AS mutual_friend
FROM follows f1
JOIN follows f2 ON f1.followed_id = f2.followed_id
WHERE f1.follower_id = 1  -- User 1
  AND f2.follower_id = 2  -- User 2;

-- Friends of friends (2 degrees)
WITH direct_friends AS (
    SELECT followed_id AS friend_id FROM follows WHERE follower_id = 1
)
SELECT DISTINCT f.followed_id AS friend_of_friend
FROM follows f
JOIN direct_friends df ON f.follower_id = df.friend_id
WHERE f.followed_id != 1  -- Not self
  AND f.followed_id NOT IN (SELECT friend_id FROM direct_friends);  -- Not already friends
```

### Recursive Traversal

```sql
-- Shortest path between two users (BFS with recursive CTE)
WITH RECURSIVE path AS (
    -- Start from user 1
    SELECT
        followed_id AS current_node,
        ARRAY[1, followed_id] AS path,
        1 AS depth
    FROM follows
    WHERE follower_id = 1

    UNION ALL

    -- Expand to next level
    SELECT
        f.followed_id,
        p.path || f.followed_id,
        p.depth + 1
    FROM follows f
    JOIN path p ON f.follower_id = p.current_node
    WHERE f.followed_id != ALL(p.path)  -- Avoid cycles
      AND p.depth < 6                    -- Max 6 degrees of separation
)
SELECT path, depth
FROM path
WHERE current_node = 42  -- Target user
ORDER BY depth
LIMIT 1;

-- Network reach: How many users can be reached within 3 hops
WITH RECURSIVE network AS (
    SELECT followed_id AS user_id, 1 AS depth
    FROM follows WHERE follower_id = 1

    UNION

    SELECT f.followed_id, n.depth + 1
    FROM follows f
    JOIN network n ON f.follower_id = n.user_id
    WHERE n.depth < 3
)
SELECT depth, COUNT(DISTINCT user_id) AS reachable_users
FROM network
GROUP BY depth
ORDER BY depth;
```

## Polymorphic Data Patterns

### Single Table with Discriminator

```sql
-- All notification types in one table
CREATE TABLE notifications (
    id SERIAL PRIMARY KEY,
    type TEXT NOT NULL CHECK (type IN ('email', 'sms', 'push', 'webhook')),
    recipient_id INT NOT NULL REFERENCES users(id),
    subject TEXT,
    body TEXT NOT NULL,
    -- Type-specific data in JSONB (flexible per type)
    type_data JSONB NOT NULL DEFAULT '{}',
    status TEXT NOT NULL DEFAULT 'pending',
    sent_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- type_data examples:
-- email: {"to": "user@example.com", "cc": [...], "html": true}
-- sms: {"phone": "+1234567890", "provider": "twilio"}
-- push: {"deviceToken": "...", "badge": 5, "sound": "default"}
-- webhook: {"url": "https://...", "headers": {...}, "retries": 3}

-- Partial indexes per type
CREATE INDEX idx_notifications_email ON notifications(recipient_id, created_at)
    WHERE type = 'email';
CREATE INDEX idx_notifications_pending ON notifications(type, created_at)
    WHERE status = 'pending';
```

### JSONB for Flexible Attributes

```sql
-- Products with different attribute sets per category
CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    category TEXT NOT NULL,
    base_price DECIMAL(10, 2) NOT NULL,
    attributes JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Electronics: {"brand": "Sony", "weight_grams": 350, "color": "black", "warranty_months": 24}
-- Clothing: {"brand": "Nike", "size": "XL", "color": "blue", "material": "cotton"}
-- Books: {"author": "John Doe", "isbn": "978-...", "pages": 320, "language": "en"}

-- GIN index for JSONB queries
CREATE INDEX idx_products_attrs ON products USING gin (attributes jsonb_path_ops);

-- Query any attribute
SELECT * FROM products WHERE attributes @> '{"brand": "Sony"}';
SELECT * FROM products WHERE (attributes->>'weight_grams')::int > 500;

-- Validate attributes per category with CHECK constraint
CREATE OR REPLACE FUNCTION validate_product_attrs(cat TEXT, attrs JSONB)
RETURNS BOOLEAN AS $$
BEGIN
    IF cat = 'electronics' THEN
        RETURN attrs ? 'brand' AND attrs ? 'weight_grams';
    ELSIF cat = 'clothing' THEN
        RETURN attrs ? 'size' AND attrs ? 'material';
    ELSIF cat = 'books' THEN
        RETURN attrs ? 'author' AND attrs ? 'isbn';
    END IF;
    RETURN TRUE;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

ALTER TABLE products ADD CONSTRAINT chk_product_attrs
    CHECK (validate_product_attrs(category, attributes));
```

## Multi-Tenant Isolation Patterns

### Shared Database, Row-Level Isolation

```sql
-- Every table includes tenant_id
CREATE TABLE projects (
    id SERIAL PRIMARY KEY,
    tenant_id INT NOT NULL REFERENCES tenants(id),
    name TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, name)  -- Unique per tenant, not globally
);

-- All indexes include tenant_id
CREATE INDEX idx_projects_tenant ON projects(tenant_id);
CREATE INDEX idx_projects_name ON projects(tenant_id, name);

-- Row-Level Security for automatic filtering
ALTER TABLE projects ENABLE ROW LEVEL SECURITY;
ALTER TABLE projects FORCE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON projects
    USING (tenant_id = current_setting('app.tenant_id')::int)
    WITH CHECK (tenant_id = current_setting('app.tenant_id')::int);

-- Application middleware sets tenant context per request:
-- await db.query("SET LOCAL app.tenant_id = $1", [tenantId]);
```

### Data Isolation Comparison

| Pattern | Isolation | Cost | Complexity | Scale |
|---------|-----------|------|------------|-------|
| Shared schema + RLS | Logical | Low | Low | 1000s of tenants |
| Schema per tenant | Strong | Medium | Medium | 10s-100s of tenants |
| Database per tenant | Strongest | High | High | 10s of tenants |

### Cross-Tenant Queries (Admin/Analytics)

```sql
-- Bypass RLS for admin queries
SET LOCAL app.tenant_id = '0';  -- Special admin tenant
-- Or: Connect as a user not subject to RLS policies

-- Or: Use a separate reporting database with all data
-- Replicate from tenant-filtered tables to a central analytics DB
```

## Archival Strategies

### Time-Based Partitioning with Archival

```sql
-- Active data in current partitions
-- Old partitions can be detached and archived

-- Detach old partition (instant, no data movement)
ALTER TABLE events DETACH PARTITION events_2022_q1;

-- Archive to cheap storage
COPY events_2022_q1 TO '/archive/events_2022_q1.csv' WITH (FORMAT csv, HEADER true);
-- Or: pg_dump -t events_2022_q1 > /archive/events_2022_q1.sql

-- Drop the partition after confirming archive
DROP TABLE events_2022_q1;

-- Reattach if needed (restore from archive first)
-- CREATE TABLE events_2022_q1 (LIKE events INCLUDING ALL);
-- COPY events_2022_q1 FROM '/archive/events_2022_q1.csv' WITH (FORMAT csv, HEADER true);
-- ALTER TABLE events ATTACH PARTITION events_2022_q1
--     FOR VALUES FROM ('2022-01-01') TO ('2022-04-01');
```

### Hot/Warm/Cold Storage Pattern

```sql
-- Hot: Current month (fast storage, full indexes)
CREATE TABLE orders_hot (
    LIKE orders INCLUDING ALL
) PARTITION BY RANGE (created_at);

-- Warm: Last 12 months (standard storage, reduced indexes)
CREATE TABLE orders_warm (
    LIKE orders INCLUDING DEFAULTS INCLUDING CONSTRAINTS
);
-- Only essential indexes on warm
CREATE INDEX idx_warm_orders_id ON orders_warm(id);
CREATE INDEX idx_warm_orders_customer ON orders_warm(customer_id);

-- Cold: Older than 12 months (compressed, minimal indexes)
CREATE TABLE orders_cold (
    LIKE orders INCLUDING DEFAULTS INCLUDING CONSTRAINTS
);
-- Minimal indexes on cold
CREATE INDEX idx_cold_orders_id ON orders_cold(id);

-- Migration cron job: Move data between tiers
-- Daily: Move orders older than 30 days from hot to warm
-- Monthly: Move orders older than 12 months from warm to cold
```

## Redis Caching Patterns

### Cache-Aside (Lazy Loading)

```
Most common pattern. Application manages cache explicitly.

Read:
1. Check cache: GET cache:user:42
2. If hit: return cached data
3. If miss: query database, SET cache:user:42 <data> EX 3600, return data

Write:
1. Update database
2. Invalidate cache: DEL cache:user:42
   (Don't update cache — let next read populate it)

Pros: Only caches data that's actually read. Simple.
Cons: Cache miss penalty on first read. Data can be stale if cache isn't invalidated.
```

### Write-Through

```
Write to cache AND database simultaneously.

Write:
1. Update cache: SET cache:user:42 <new_data> EX 3600
2. Update database

Read:
1. Check cache: GET cache:user:42
2. If hit: return (always fresh)
3. If miss: query database, SET cache, return

Pros: Cache is always up-to-date. No stale data.
Cons: Write overhead (two writes per mutation). Caches data that may never be read.
```

### Write-Behind (Write-Back)

```
Write to cache immediately. Database is updated asynchronously.

Write:
1. Update cache: SET cache:user:42 <new_data>
2. Add to write queue: LPUSH write_queue '{"table":"users","id":42,...}'
3. Background worker processes queue and writes to database

Read:
1. Check cache: GET cache:user:42 (always fresh)

Pros: Very fast writes. Reduces database write load.
Cons: Risk of data loss if Redis crashes before queue is drained.
      Complex error handling. Eventual consistency.
```

### Cache Invalidation Strategies

```
1. TTL (Time-To-Live): Simplest. Data expires after fixed time.
   SET cache:user:42 <data> EX 3600  # 1 hour

2. Event-based: Invalidate when data changes.
   # In application: after UPDATE users SET ... WHERE id = 42:
   DEL cache:user:42

3. Version-based: Include version in key.
   SET cache:user:42:v7 <data>
   # Increment version on write, old keys expire naturally

4. Pattern-based: Invalidate related keys.
   # Delete all cache keys for a user:
   # Store a set of related keys
   SADD user:42:cache_keys "cache:user:42:profile" "cache:user:42:prefs"
   # On invalidation:
   SMEMBERS user:42:cache_keys → DEL each key
```

### Redis Data Structure Patterns

```
# Session store (Hash)
HSET session:abc123 userId 42 email "user@example.com" role "admin" createdAt "2024-01-15"
EXPIRE session:abc123 86400

# Rate limiter (Sorted Set with sliding window)
# Add request timestamp
ZADD rate:user:42 <timestamp_ms> <unique_id>
# Remove old entries
ZREMRANGEBYSCORE rate:user:42 0 <timestamp_ms - window>
# Count requests in window
ZCARD rate:user:42
# If count > limit, reject

# Distributed lock (String with NX)
SET lock:resource:42 "owner:worker-1" NX EX 30
# NX = set only if not exists, EX 30 = expire in 30 seconds
# If SET returns OK: lock acquired
# If SET returns nil: lock held by someone else

# Pub/Sub for real-time features
SUBSCRIBE channel:notifications:user:42
PUBLISH channel:notifications:user:42 '{"type":"new_message","from":"user:17"}'

# Stream for event log (Redis 5.0+)
XADD events:orders * action "placed" orderId "42" total "99.99"
XREAD COUNT 10 BLOCK 5000 STREAMS events:orders $
# Consumer groups for distributed processing
XGROUP CREATE events:orders workers $ MKSTREAM
XREADGROUP GROUP workers worker-1 COUNT 10 BLOCK 5000 STREAMS events:orders >
```

## MongoDB Schema Design Patterns

### Subset Pattern

```javascript
// Problem: Document is too large because of a growing array
// Solution: Embed a frequently-accessed subset, reference the full set

// Product with reviews
{
  _id: ObjectId("product1"),
  name: "Wireless Mouse",
  price: 29.99,
  // Embed only top 10 reviews (frequently displayed)
  topReviews: [
    { userId: ObjectId("u1"), rating: 5, text: "Great!", date: ISODate("2024-01-15") },
    { userId: ObjectId("u2"), rating: 4, text: "Good value", date: ISODate("2024-01-14") }
  ],
  reviewStats: {
    count: 234,
    averageRating: 4.3,
    distribution: { 5: 120, 4: 60, 3: 30, 2: 14, 1: 10 }
  }
}

// Full reviews in separate collection
// db.reviews — referenced by productId
```

### Bucket Pattern

```javascript
// Problem: One document per data point = too many documents
// Solution: Group data points into time-based buckets

// Instead of:
// { sensorId: "temp-1", time: ISODate("..."), value: 22.5 }
// { sensorId: "temp-1", time: ISODate("..."), value: 22.6 }
// ... millions of documents

// Bucket by hour:
{
  sensorId: "temp-1",
  date: ISODate("2024-01-15T10:00:00Z"),  // Bucket start
  measurements: [
    { minute: 0, value: 22.5 },
    { minute: 5, value: 22.6 },
    { minute: 10, value: 22.4 },
    // ... up to 12 entries per hour (every 5 min)
  ],
  count: 12,
  sum: 270.0,
  min: 22.1,
  max: 23.2,
  avg: 22.5
}

// Benefits:
// - 12x fewer documents
// - Pre-computed aggregates (sum, min, max, avg)
// - Efficient time-range queries
// - Each document stays small and bounded
```

### Computed Pattern

```javascript
// Problem: Expensive calculations on every read
// Solution: Pre-compute and store results

{
  _id: ObjectId("product1"),
  name: "Wireless Mouse",
  price: 29.99,
  // Computed fields (updated by background job or trigger)
  computed: {
    totalSales: 1523,
    totalRevenue: 45661.77,
    averageRating: 4.3,
    reviewCount: 234,
    popularityScore: 87.5,
    lastComputed: ISODate("2024-01-15T12:00:00Z")
  }
}

// Update computed fields periodically:
db.products.updateOne(
  { _id: ObjectId("product1") },
  [{
    $set: {
      "computed.totalSales": { /* aggregation result */ },
      "computed.lastComputed": "$$NOW"
    }
  }]
);
```

### Schema Versioning Pattern

```javascript
// Problem: Schema evolves but old documents exist
// Solution: Version field + migration on read

{
  _id: ObjectId("..."),
  schemaVersion: 2,
  // V1 had: name (string)
  // V2 has: firstName, lastName (split)
  firstName: "Jane",
  lastName: "Smith",
  email: "jane@example.com"
}

// Application code handles versioning:
function normalizeUser(doc) {
  if (doc.schemaVersion === 1) {
    const [firstName, ...rest] = doc.name.split(' ');
    return { ...doc, firstName, lastName: rest.join(' '), schemaVersion: 2 };
  }
  return doc;
}

// Background migration:
db.users.find({ schemaVersion: 1 }).forEach(doc => {
  const [firstName, ...rest] = doc.name.split(' ');
  db.users.updateOne(
    { _id: doc._id },
    {
      $set: { firstName, lastName: rest.join(' '), schemaVersion: 2 },
      $unset: { name: "" }
    }
  );
});
```

### Extended Reference Pattern

```javascript
// Problem: Need data from related documents without $lookup
// Solution: Copy frequently-accessed fields from related documents

// Order document with extended customer reference
{
  _id: ObjectId("order1"),
  customerId: ObjectId("customer42"),
  // Extended reference: copy fields needed for display
  customer: {
    name: "Jane Smith",
    email: "jane@example.com",
    tier: "premium"
  },
  items: [...],
  total: 129.99,
  status: "shipped"
}

// Trade-offs:
// + No $lookup needed for order display
// + Fast reads
// - Must update copies when customer data changes
// - Data can be temporarily stale

// Update copies when source changes:
db.orders.updateMany(
  { customerId: ObjectId("customer42") },
  { $set: {
    "customer.name": "Jane Doe",  // Name changed
    "customer.email": "jane.doe@example.com"  // Email changed
  }}
);
```

### Outlier Pattern

```javascript
// Problem: Most documents are small, but some have huge arrays
// Solution: Overflow documents for outliers

// Normal book: few authors
{
  _id: ObjectId("book1"),
  title: "Database Design",
  authors: ["Alice", "Bob"],
  hasOverflow: false
}

// Outlier book: 100+ authors (anthology)
{
  _id: ObjectId("book2"),
  title: "Greatest Short Stories",
  authors: ["Author1", "Author2", /* ... first 20 */],
  hasOverflow: true
}

// Overflow collection
{
  bookId: ObjectId("book2"),
  authors: ["Author21", "Author22", /* ... remaining authors */]
}

// Query: if hasOverflow, also query overflow collection
```

## Data Lifecycle Management

### Lifecycle Stages

```
1. Active: Current, frequently accessed data
   - In main tables
   - Full indexing
   - Full backup frequency

2. Warm: Recent historical, occasionally accessed
   - In same database, possibly separate tablespace
   - Reduced indexing
   - Standard backup frequency

3. Cold: Old historical, rarely accessed
   - In separate tables or database
   - Minimal indexing
   - Compressed storage
   - Less frequent backups

4. Archive: Compliance/legal retention
   - Exported to file storage (S3, etc.)
   - No indexes
   - Compressed and possibly encrypted
   - Retained per legal requirements

5. Deleted: Past retention period
   - Permanently removed
   - Removal documented for compliance
```

### Implementing Data Lifecycle

```sql
-- Automated lifecycle with PostgreSQL partitioning

-- Active: current quarter partition
-- Warm: previous 4 quarters
-- Cold: move to compressed tablespace
-- Archive: export and drop partition

-- Create archival function
CREATE OR REPLACE FUNCTION archive_old_partitions(
    table_name TEXT,
    months_to_keep INT,
    archive_path TEXT
) RETURNS VOID AS $$
DECLARE
    partition RECORD;
    cutoff_date DATE;
BEGIN
    cutoff_date := DATE_TRUNC('month', NOW()) - (months_to_keep || ' months')::interval;

    FOR partition IN
        SELECT inhrelid::regclass AS name
        FROM pg_inherits
        WHERE inhparent = table_name::regclass
    LOOP
        -- Check if partition is older than cutoff
        -- Export, detach, and drop
        RAISE NOTICE 'Checking partition %', partition.name;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- Schedule with pg_cron
SELECT cron.schedule('archive-old-events', '0 3 1 * *',  -- 1st of each month at 3 AM
    'SELECT archive_old_partitions(''events'', 12, ''/archive/'')');
```
