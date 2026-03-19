# NoSQL Design Patterns

> Architectural patterns for NoSQL databases covering data modeling, consistency,
> partitioning, event sourcing, and CQRS. Applicable to Redis, MongoDB, DynamoDB,
> Cassandra, and other NoSQL systems.

---

## 1. CAP Theorem in Practice

```
CAP Theorem: A distributed system can guarantee at most TWO of three:
  C — Consistency (every read gets the most recent write)
  A — Availability (every request gets a response)
  P — Partition Tolerance (system works despite network splits)

In practice, P is non-negotiable in distributed systems. The real choice is:

  CP systems: Consistent + Partition Tolerant
    → MongoDB (with majority write concern)
    → Redis Cluster (strong consistency within a shard)
    → HBase, Zookeeper
    Tradeoff: May reject writes during partitions

  AP systems: Available + Partition Tolerant
    → Cassandra
    → DynamoDB (eventual consistency mode)
    → CouchDB
    Tradeoff: May return stale data during partitions

  CA systems: Don't exist in distributed systems.
    Single-node PostgreSQL or MySQL is CA, but not distributed.
```

### Consistency Spectrum

```
Strong Consistency           Eventual Consistency
  ├────────────────────────────────────────────┤
  │                                            │
  MongoDB     Redis      DynamoDB    Cassandra  │
  (majority)  (single)   (strong)    (ONE)     │
                                               │
  Every read    Most reads  Configurable  Reads may
  sees latest   see latest  per query     be stale
  write         write                     briefly

Tunable consistency (per-operation):
  MongoDB:  writeConcern + readConcern + readPreference
  DynamoDB: ConsistentRead: true/false
  Cassandra: LOCAL_ONE, LOCAL_QUORUM, ALL, etc.
```

---

## 2. Document Design Patterns

### Polymorphic Pattern

Store different types of documents in the same collection with a type discriminator.

```javascript
// Single "notifications" collection with different shapes
{ type: "email", to: "alice@example.com", subject: "Welcome", body: "..." }
{ type: "sms", phone: "+15551234567", message: "Your code is 1234" }
{ type: "push", device_token: "abc123", title: "New message", badge: 3 }
{ type: "webhook", url: "https://api.example.com/hook", payload: {...} }

// Query by type
db.notifications.find({ type: "email" });

// Index with partial filter
db.notifications.createIndex(
  { phone: 1 },
  { partialFilterExpression: { type: "sms" } }
);
```

**When to use**: Multiple entity types share common query patterns (e.g., activity feed, notifications, events). Avoids joins across multiple collections.

### Bucket Pattern

Group related data into fixed-size buckets to reduce document count and improve query efficiency.

```javascript
// Instead of one document per reading:
{ sensor_id: "temp-1", timestamp: ISODate("2024-03-19T10:00:00"), value: 22.5 }
{ sensor_id: "temp-1", timestamp: ISODate("2024-03-19T10:01:00"), value: 22.6 }
// ... thousands of documents

// Bucket pattern: one document per hour
{
  sensor_id: "temp-1",
  bucket_start: ISODate("2024-03-19T10:00:00"),
  bucket_end: ISODate("2024-03-19T10:59:59"),
  count: 60,
  readings: [
    { timestamp: ISODate("2024-03-19T10:00:00"), value: 22.5 },
    { timestamp: ISODate("2024-03-19T10:01:00"), value: 22.6 },
    // ... up to 60 readings
  ],
  summary: {
    min: 21.8,
    max: 23.1,
    avg: 22.4
  }
}
```

**When to use**: Time-series data, IoT sensor data, event logs. Reduces document count by 10-100x. Pre-computed summaries avoid aggregation.

### Computed Pattern

Pre-compute expensive calculations and store results alongside the source data.

```javascript
// Product with pre-computed review stats
{
  _id: ObjectId("..."),
  name: "Wireless Mouse",
  // ... other fields

  // Computed on every review write (not on every read)
  review_stats: {
    count: 142,
    average_rating: 4.3,
    rating_distribution: { "5": 67, "4": 38, "3": 22, "2": 10, "1": 5 },
    last_updated: ISODate("2024-03-19")
  }
}

// Update stats when a new review is added
db.products.updateOne(
  { _id: productId },
  [
    {
      $set: {
        "review_stats.count": { $add: ["$review_stats.count", 1] },
        "review_stats.last_updated": "$$NOW"
        // Recalculate average via aggregation on reviews collection
      }
    }
  ]
);
```

**When to use**: Read-heavy workloads where computing at read time is expensive. Dashboard metrics, analytics summaries, report data.

### Extended Reference Pattern

Embed frequently-accessed fields from a related document to avoid joins.

```javascript
// Order embeds customer name and email (denormalized from users collection)
{
  order_number: "ORD-2024-001234",
  customer_ref: {
    _id: ObjectId("..."),       // Reference for full lookup
    name: "Alice Smith",         // Copied for display
    email: "alice@example.com"   // Copied for notifications
  },
  items: [...]
}

// When customer name changes, update denormalized copies
db.orders.updateMany(
  { "customer_ref._id": customerId },
  { $set: { "customer_ref.name": newName } }
);
```

**When to use**: When you almost always need specific fields from a related document but don't need the full document. Trade write complexity for read performance.

### Outlier Pattern

Handle data with occasional outliers that would otherwise cause document bloat.

```javascript
// Book with bounded reviews array + overflow flag
{
  _id: ObjectId("..."),
  title: "Popular Book",
  reviews: [...],        // Keep first 100 reviews embedded
  has_overflow: true,    // Flag indicating more reviews exist elsewhere
  review_count: 5842     // Actual total count
}

// Overflow reviews in separate collection
{
  book_id: ObjectId("..."),
  batch: 2,              // Overflow batch number
  reviews: [...]         // Reviews 101-200, 201-300, etc.
}
```

**When to use**: When most documents stay within bounds but a small percentage grow very large (e.g., viral content, celebrity accounts).

### Schema Versioning Pattern

Handle schema evolution without downtime migration.

```javascript
// Version 1
{ schema_version: 1, name: "Alice", address: "123 Main St, Portland, OR 97201" }

// Version 2 (structured address)
{ schema_version: 2, name: "Alice", address: { street: "123 Main St", city: "Portland", state: "OR", zip: "97201" } }

// Application code handles both versions
function getAddress(doc) {
  if (doc.schema_version === 1) {
    return parseAddressString(doc.address);  // Parse legacy format
  }
  return doc.address;  // Already structured
}

// Lazy migration: update to v2 on next write
function updateUser(doc, updates) {
  if (doc.schema_version < 2) {
    updates.address = parseAddressString(doc.address);
    updates.schema_version = 2;
  }
  return db.users.updateOne({ _id: doc._id }, { $set: updates });
}
```

**When to use**: Live systems that can't afford downtime for migration. Gradual rollout of schema changes. Multi-version API support.

---

## 3. Key-Value Patterns (Redis-Focused)

### Cache-Aside (Lazy Loading)

```
Read path:
  1. Check cache
  2. If hit → return cached value
  3. If miss → read from database
  4. Store in cache with TTL
  5. Return value

Write path:
  1. Write to database
  2. Invalidate cache (delete key)
  -- Do NOT update cache (stale data risk from race condition)
```

**Advantages**: Only caches what's actually read. Simple to implement.
**Disadvantages**: Cache miss = extra latency. Cold cache problem on startup.

### Read-Through / Write-Through

```
Read-Through:
  Application → Cache → Database
  Cache automatically loads from DB on miss.
  Application only talks to cache.

Write-Through:
  Application → Cache → Database (synchronous)
  Cache writes to DB before confirming to application.
  Data in cache is always consistent.

Write-Behind (Write-Back):
  Application → Cache → Database (asynchronous)
  Cache writes to DB in background (batched).
  Risk: data loss if cache crashes before flushing.
```

### Distributed Lock Pattern

```python
# Correct implementation (with fencing)
import uuid
import time

def acquire_lock(redis, resource, ttl=30):
    token = str(uuid.uuid4())
    acquired = redis.set(
        f"lock:{resource}",
        token,
        nx=True,     # Only if not exists
        ex=ttl        # Auto-expire (prevents deadlock)
    )
    return token if acquired else None

def release_lock(redis, resource, token):
    # Atomic: only release if we still own it
    script = """
    if redis.call('GET', KEYS[1]) == ARGV[1] then
        return redis.call('DEL', KEYS[1])
    end
    return 0
    """
    return redis.eval(script, 1, f"lock:{resource}", token)

# Usage
token = acquire_lock(redis, "process-orders")
if token:
    try:
        process_orders()
    finally:
        release_lock(redis, "process-orders", token)

# IMPORTANT: For multi-node Redis, use Redlock algorithm
# (acquire lock on majority of N independent Redis instances)
```

### Circuit Breaker with Redis

```python
class CircuitBreaker:
    CLOSED = "closed"      # Normal operation
    OPEN = "open"          # Failing, reject requests
    HALF_OPEN = "half_open"  # Testing recovery

    def __init__(self, redis, service, threshold=5, timeout=60):
        self.redis = redis
        self.key = f"circuit:{service}"
        self.threshold = threshold
        self.timeout = timeout

    def is_open(self):
        state = self.redis.hgetall(self.key)
        if not state:
            return False
        if state.get("state") == self.OPEN:
            opened_at = float(state.get("opened_at", 0))
            if time.time() - opened_at > self.timeout:
                self.redis.hset(self.key, "state", self.HALF_OPEN)
                return False  # Allow one test request
            return True
        return False

    def record_success(self):
        self.redis.hmset(self.key, {"state": self.CLOSED, "failures": 0})

    def record_failure(self):
        failures = self.redis.hincrby(self.key, "failures", 1)
        if failures >= self.threshold:
            self.redis.hmset(self.key, {
                "state": self.OPEN,
                "opened_at": str(time.time())
            })
```

---

## 4. Consistency Patterns

### Saga Pattern (Distributed Transactions)

```
When a single operation spans multiple services/databases,
use a saga to coordinate with compensating transactions.

Choreography-based (event-driven):
  Order Service → "OrderCreated" event
  Payment Service → processes payment → "PaymentCompleted" event
  Inventory Service → reserves stock → "StockReserved" event
  Shipping Service → creates shipment → "ShipmentCreated" event

  If Payment fails:
  Payment Service → "PaymentFailed" event
  Order Service → cancels order (compensation)

Orchestrator-based:
  Saga Orchestrator coordinates all steps sequentially.
  On failure, calls compensation functions in reverse order.
```

```javascript
// MongoDB + Redis saga implementation sketch
async function createOrderSaga(orderData) {
  const sagaId = new ObjectId();
  const steps = [];

  try {
    // Step 1: Create order (pending)
    const order = await db.orders.insertOne({
      ...orderData, status: "pending", saga_id: sagaId
    });
    steps.push({ service: "order", action: "create", id: order.insertedId });

    // Step 2: Reserve inventory
    const reserved = await db.inventory.updateOne(
      { product_id: orderData.product_id, quantity: { $gte: orderData.qty } },
      { $inc: { quantity: -orderData.qty }, $push: { reservations: sagaId } }
    );
    if (!reserved.modifiedCount) throw new Error("Insufficient inventory");
    steps.push({ service: "inventory", action: "reserve" });

    // Step 3: Charge payment
    const payment = await chargePayment(orderData.payment);
    steps.push({ service: "payment", action: "charge", id: payment.id });

    // Step 4: Confirm order
    await db.orders.updateOne({ _id: order.insertedId }, { $set: { status: "confirmed" } });

    return { success: true, orderId: order.insertedId };

  } catch (error) {
    // Compensate in reverse order
    for (const step of steps.reverse()) {
      await compensate(step, sagaId, orderData);
    }
    return { success: false, error: error.message };
  }
}

async function compensate(step, sagaId, orderData) {
  switch (`${step.service}:${step.action}`) {
    case "payment:charge":
      await refundPayment(step.id);
      break;
    case "inventory:reserve":
      await db.inventory.updateOne(
        { product_id: orderData.product_id },
        { $inc: { quantity: orderData.qty }, $pull: { reservations: sagaId } }
      );
      break;
    case "order:create":
      await db.orders.updateOne({ _id: step.id }, { $set: { status: "cancelled" } });
      break;
  }
}
```

### Eventual Consistency with Change Data Capture

```javascript
// MongoDB Change Streams for real-time data synchronization
const pipeline = [
  { $match: { operationType: { $in: ["insert", "update", "replace"] } } },
  { $match: { "ns.coll": "products" } }
];

const changeStream = db.watch(pipeline, {
  fullDocument: "updateLookup",  // Include full document on updates
  fullDocumentBeforeChange: "whenAvailable"  // MongoDB 6.0+
});

changeStream.on("change", async (change) => {
  const product = change.fullDocument;

  // Sync to Redis cache
  await redis.setex(
    `product:${product._id}`,
    3600,
    JSON.stringify(product)
  );

  // Sync to search index
  await searchIndex.update(product._id, {
    name: product.name,
    description: product.description,
    category: product.category,
    price: product.pricing.base_price
  });

  // Publish event for other services
  await redis.xadd("events:product_updated", "*", {
    product_id: String(product._id),
    action: change.operationType,
    timestamp: new Date().toISOString()
  });
});

// Resume after crash (using resume token)
const resumeToken = await redis.get("changestream:products:token");
const options = resumeToken
  ? { resumeAfter: JSON.parse(resumeToken) }
  : { startAtOperationTime: Timestamp(0, Date.now() / 1000) };

const stream = db.watch(pipeline, options);
stream.on("change", async (change) => {
  // Process change...
  // Save resume token for crash recovery
  await redis.set("changestream:products:token", JSON.stringify(change._id));
});
```

---

## 5. Event Sourcing

```
Traditional: Store current state
  users table: { id: 1, name: "Alice", email: "alice@example.com", balance: 150 }

Event Sourcing: Store every state change as an immutable event
  events: [
    { type: "UserCreated", data: { name: "Alice", email: "alice@example.com" }, timestamp: T1 },
    { type: "DepositMade", data: { amount: 200 }, timestamp: T2 },
    { type: "WithdrawalMade", data: { amount: 50 }, timestamp: T3 },
    { type: "EmailChanged", data: { email: "alice.new@example.com" }, timestamp: T4 },
  ]
  Current state = replay all events: balance = 0 + 200 - 50 = 150
```

### MongoDB Event Store

```javascript
// Events collection
const eventSchema = {
  stream_id: ObjectId,          // Aggregate root ID
  stream_type: String,          // "User", "Order", etc.
  event_type: String,           // "UserCreated", "OrderPlaced"
  event_data: Object,           // Event payload
  metadata: {
    correlation_id: String,     // Trace across events
    caused_by: String,          // What triggered this event
    user_id: String,            // Who caused the event
  },
  version: Number,              // Optimistic concurrency
  created_at: Date
};

// Indexes
db.events.createIndex({ stream_id: 1, version: 1 }, { unique: true });
db.events.createIndex({ stream_type: 1, created_at: 1 });
db.events.createIndex({ event_type: 1, created_at: 1 });

// Append event with optimistic concurrency
async function appendEvent(streamId, streamType, eventType, data, expectedVersion) {
  const event = {
    stream_id: streamId,
    stream_type: streamType,
    event_type: eventType,
    event_data: data,
    version: expectedVersion + 1,
    created_at: new Date()
  };

  try {
    await db.events.insertOne(event);
  } catch (err) {
    if (err.code === 11000) {  // Duplicate key = version conflict
      throw new Error("Concurrency conflict: stream was modified");
    }
    throw err;
  }

  // Publish to Redis Stream for projections
  await redis.xadd(`events:${streamType}`, "*", {
    stream_id: String(streamId),
    event_type: eventType,
    data: JSON.stringify(data)
  });
}

// Rebuild state from events
async function loadAggregate(streamId) {
  const events = await db.events
    .find({ stream_id: streamId })
    .sort({ version: 1 })
    .toArray();

  let state = {};
  for (const event of events) {
    state = applyEvent(state, event);
  }
  return { state, version: events.length };
}

function applyEvent(state, event) {
  switch (event.event_type) {
    case "AccountCreated":
      return { ...event.event_data, balance: 0 };
    case "DepositMade":
      return { ...state, balance: state.balance + event.event_data.amount };
    case "WithdrawalMade":
      return { ...state, balance: state.balance - event.event_data.amount };
    default:
      return state;
  }
}

// Snapshots for performance (avoid replaying all events)
async function saveSnapshot(streamId, state, version) {
  await db.snapshots.updateOne(
    { stream_id: streamId },
    { $set: { state, version, created_at: new Date() } },
    { upsert: true }
  );
}

async function loadAggregateWithSnapshot(streamId) {
  const snapshot = await db.snapshots.findOne({ stream_id: streamId });
  const fromVersion = snapshot ? snapshot.version : 0;

  const events = await db.events
    .find({ stream_id: streamId, version: { $gt: fromVersion } })
    .sort({ version: 1 })
    .toArray();

  let state = snapshot ? snapshot.state : {};
  for (const event of events) {
    state = applyEvent(state, event);
  }

  // Save new snapshot every 100 events
  if (events.length > 100) {
    await saveSnapshot(streamId, state, fromVersion + events.length);
  }

  return { state, version: fromVersion + events.length };
}
```

---

## 6. CQRS (Command Query Responsibility Segregation)

```
Traditional: Same model for reads and writes
  Application → Database (read + write)

CQRS: Separate models for reads and writes
  Commands (writes) → Write Model → Event Store (MongoDB)
  Queries (reads) → Read Model → Projections (Redis / denormalized MongoDB)

Write side:                    Read side:
  Validate command             Denormalized views
  Apply business rules         Optimized for queries
  Emit events                  Updated via event handlers
  Strong consistency           Eventually consistent
```

### Implementation Pattern

```javascript
// WRITE SIDE: Command handler
class OrderCommandHandler {
  async handle(command) {
    switch (command.type) {
      case "PlaceOrder": {
        // Load aggregate
        const { state, version } = await loadAggregate(command.customerId);

        // Business rules
        if (state.suspended) throw new Error("Account suspended");
        if (command.total > state.creditLimit) throw new Error("Credit limit exceeded");

        // Emit event
        await appendEvent(
          command.orderId, "Order", "OrderPlaced",
          { customerId: command.customerId, items: command.items, total: command.total },
          0  // New stream, version 0
        );
        break;
      }

      case "ShipOrder": {
        const { state, version } = await loadAggregate(command.orderId);
        if (state.status !== "confirmed") throw new Error("Order not confirmed");

        await appendEvent(
          command.orderId, "Order", "OrderShipped",
          { trackingNumber: command.trackingNumber, shippedAt: new Date() },
          version
        );
        break;
      }
    }
  }
}

// READ SIDE: Projection (event handler that builds read models)
class OrderProjection {
  async handle(event) {
    switch (event.event_type) {
      case "OrderPlaced": {
        // Write to denormalized read model
        await db.order_views.insertOne({
          _id: event.stream_id,
          customer_id: event.event_data.customerId,
          items: event.event_data.items,
          total: event.event_data.total,
          status: "placed",
          created_at: event.created_at
        });

        // Update customer order count (Redis for fast access)
        await redis.hincrby(`customer:${event.event_data.customerId}`, "order_count", 1);
        break;
      }

      case "OrderShipped": {
        await db.order_views.updateOne(
          { _id: event.stream_id },
          {
            $set: {
              status: "shipped",
              tracking_number: event.event_data.trackingNumber,
              shipped_at: event.event_data.shippedAt
            }
          }
        );
        break;
      }
    }
  }
}

// QUERY SIDE: Fast reads from denormalized models
app.get("/api/orders/:id", async (req, res) => {
  const order = await db.order_views.findOne({ _id: new ObjectId(req.params.id) });
  res.json(order);
});

app.get("/api/customers/:id/stats", async (req, res) => {
  const stats = await redis.hgetall(`customer:${req.params.id}`);
  res.json(stats);
});
```

---

## 7. Anti-Patterns to Avoid

### Unbounded Arrays

```javascript
// BAD: Array grows forever
{
  user_id: "123",
  followers: ["user_1", "user_2", ..., "user_1000000"]  // Document > 16MB!
}

// GOOD: Separate collection with index
// followers collection
{ user_id: "123", follower_id: "456", followed_at: ISODate("...") }
// Index: { user_id: 1, followed_at: -1 }
```

### Hot Keys (Redis)

```
BAD: All requests hit one key
  INCR global_counter        -- Every request increments same key

GOOD: Shard the counter
  INCR counter:shard:{hash(request_id) % 16}
  -- Read: SUM all 16 shards
  -- Write: distributed across 16 keys
```

### God Documents

```javascript
// BAD: Everything embedded in one massive document
{
  user_id: "123",
  profile: {...},
  settings: {...},
  orders: [{...}, {...}, ...],      // Unbounded
  notifications: [{...}, {...}],    // Unbounded
  activity_log: [{...}, {...}],     // Unbounded
  followers: [...],                 // Unbounded
  messages: [...]                   // Unbounded
}

// GOOD: Separate collections for unbounded/independent data
// users: { _id, profile, settings }          -- Bounded, always read together
// orders: { user_id, ... }                   -- Separate, paginated
// notifications: { user_id, ... }            -- Separate, TTL indexed
// activity_log: { user_id, ... }             -- Time-series collection
```

### Ignoring Read/Write Ratios

```
Before choosing a pattern, know your ratios:

  Read-heavy (100:1 read:write)
    → Denormalize aggressively
    → Pre-compute aggregations
    → Cache everything
    → Accept slower writes

  Write-heavy (1:100 read:write)
    → Normalize to minimize write amplification
    → Use append-only patterns (time-series, event log)
    → Batch writes
    → Read from replicas

  Balanced (1:1)
    → Moderate denormalization
    → Cache hot paths only
    → Use change streams for sync
```

### Using KEYS in Production (Redis)

```
NEVER:
  KEYS user:*                -- Blocks Redis for ALL clients while scanning

ALWAYS:
  SCAN 0 MATCH user:* COUNT 100    -- Non-blocking, iterative
  -- Process in batches, never blocks other clients
```

---

## 8. Data Access Patterns Summary

| Pattern | Best For | Technology |
|---------|----------|------------|
| Cache-Aside | Read-heavy, tolerates stale | Redis + any DB |
| Write-Through | Consistency critical | Redis + any DB |
| Event Sourcing | Audit trail, undo, replay | MongoDB (events) + Redis (projections) |
| CQRS | Different read/write models | MongoDB (write) + Redis/denormalized (read) |
| Saga | Distributed transactions | MongoDB + Redis Streams |
| Bucket | Time-series, IoT | MongoDB time-series collections |
| Computed | Dashboard aggregates | MongoDB (materialized views) |
| Extended Reference | Avoid joins | MongoDB (denormalized fields) |
| Change Streams | Real-time sync | MongoDB → Redis/search/cache |
| Distributed Lock | Mutual exclusion | Redis (SET NX EX) |
| Circuit Breaker | Fault tolerance | Redis (state + counters) |
| Rate Limiting | API protection | Redis (sorted set sliding window) |
