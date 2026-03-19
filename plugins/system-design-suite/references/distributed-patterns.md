# Distributed Patterns Reference

Quick reference for distributed system patterns. For full architectural context, see the system-design-architect agent.

---

## Saga Pattern

Manages distributed transactions across multiple services without a traditional two-phase commit. Each service performs its local transaction and publishes an event. If any step fails, compensating transactions undo prior steps.

### Choreography (Event-Based)

Each service listens for events and decides what to do.

```
Order Placed → Payment Service charges card
  ↓ success
Payment Completed → Inventory Service reserves items
  ↓ success
Items Reserved → Shipping Service creates shipment
  ↓ success
Shipment Created → Order Service marks as confirmed

On failure at any step → compensating events:
Inventory Reserve Failed → Payment Service refunds card
                         → Order Service marks as failed
```

```
┌─────────┐  OrderPlaced  ┌─────────┐  PaymentDone  ┌──────────┐
│  Order   │──────────────►│ Payment │──────────────►│Inventory │
│ Service  │               │ Service │               │ Service  │
└─────────┘               └─────────┘               └──────────┘
     ▲                         │                          │
     │                   PaymentFailed              ReserveFailed
     │                         │                          │
     └─────────────────────────┴──────────────────────────┘
                        Compensation
```

**Pros**: Loose coupling, each service is independent
**Cons**: Hard to track overall saga state, complex failure scenarios, difficult debugging

### Orchestration (Central Coordinator)

A central saga orchestrator tells each service what to do.

```
┌──────────────┐
│    Saga      │
│ Orchestrator │
└──────┬───────┘
       │
  1. Charge card
       ├──────────► Payment Service
       │              │
  2. Reserve items    │ success/fail
       ├──────────► Inventory Service
       │              │
  3. Create shipment  │ success/fail
       ├──────────► Shipping Service
       │              │
  4. Confirm order    │ success/fail
       └──────────► Order Service
```

**Orchestrator state machine**:
```
STARTED → PAYMENT_PENDING → PAYMENT_DONE → INVENTORY_PENDING
  → INVENTORY_RESERVED → SHIPPING_PENDING → SHIPPING_CREATED → COMPLETED

On failure:
  INVENTORY_FAILED → PAYMENT_REFUNDING → PAYMENT_REFUNDED → FAILED
```

**Pros**: Centralized logic, easy to understand flow, better observability
**Cons**: Orchestrator can become a bottleneck/SPOF, more coupling to orchestrator

### Choosing Choreography vs Orchestration

```
Choreography when:
  - 2-4 steps in the saga
  - Services are truly independent
  - Team ownership is distributed
  - Simple failure compensation

Orchestration when:
  - 5+ steps
  - Complex failure/compensation logic
  - Need centralized monitoring
  - Steps have conditional logic
```

### Saga Implementation Tips

```
1. Every step must have a compensating action
   Step: Charge card → Compensation: Refund card
   Step: Reserve inventory → Compensation: Release inventory
   Step: Send email → Compensation: Send cancellation email (or accept as non-compensable)

2. Compensations are NOT rollbacks
   They're new forward actions that semantically undo the effect
   A refund is not "undoing" a charge — it's a new transaction

3. Make steps idempotent
   Retrying a step should produce the same result
   Use idempotency keys for each saga step

4. Handle partial failures
   What if compensation itself fails?
   → Retry with backoff
   → Dead letter queue for manual intervention
   → Alert operations team

5. Track saga state persistently
   Store saga state in a database (not just in memory)
   On crash/restart, resume from last known state
```

---

## Event Sourcing

Instead of storing current state, store the sequence of events that led to current state. The current state is derived by replaying events.

```
Traditional CRUD:
  Account table: { id: 1, balance: 150 }

Event Sourcing:
  Event store:
    1. AccountCreated  { id: 1, initial_balance: 0 }
    2. MoneyDeposited  { id: 1, amount: 200 }
    3. MoneyWithdrawn  { id: 1, amount: 50 }

  Current state: replay events → balance = 0 + 200 - 50 = 150
```

### Event Store Schema

```sql
CREATE TABLE events (
  id bigserial PRIMARY KEY,
  aggregate_type varchar(100) NOT NULL,  -- 'Account', 'Order'
  aggregate_id varchar(100) NOT NULL,    -- the entity ID
  event_type varchar(100) NOT NULL,      -- 'MoneyDeposited'
  event_data jsonb NOT NULL,             -- event payload
  metadata jsonb,                        -- correlation_id, user_id, etc.
  version integer NOT NULL,              -- optimistic concurrency
  created_at timestamptz DEFAULT now(),

  UNIQUE (aggregate_id, version)         -- prevent concurrent writes
);

CREATE INDEX idx_events_aggregate ON events(aggregate_id, version);
```

### Snapshots

Replaying all events from the beginning gets slow. Take periodic snapshots.

```
Events: 1, 2, 3, ..., 997, 998, 999, [SNAPSHOT at event 1000], 1001, 1002, ...

To rebuild state:
  1. Load latest snapshot (state at event 1000)
  2. Replay events 1001, 1002, ... (only the recent ones)

Snapshot frequency: every 100-1000 events (depends on event rate)
Store snapshots separately from events
```

### When to Use Event Sourcing

```
Good fit:
  - Audit trail is a hard requirement (finance, healthcare, legal)
  - Need to replay history (debugging, analytics)
  - Complex domain with temporal queries ("what was the state on Jan 15?")
  - Event-driven architecture already in place

Bad fit:
  - Simple CRUD with no audit needs
  - Team unfamiliar with the pattern (steep learning curve)
  - Real-time queries needed without CQRS (rebuilding state is slow)
  - Simple domains where the overhead isn't justified
```

---

## CQRS (Command Query Responsibility Segregation)

Separate the read model from the write model. Commands (writes) go to one model; queries (reads) go to another, optimized for reading.

```
                    ┌──────────────┐
   Commands ───────►│  Write Model  │──── Events ────┐
   (POST, PUT,      │  (normalized, │                │
    DELETE)          │   consistent) │                │
                    └──────────────┘                │
                                                     │
                    ┌──────────────┐                │
   Queries ────────►│  Read Model   │◄── Projections─┘
   (GET)            │  (denormalized│    (async update
                    │   optimized   │     read model)
                    │   for reads)  │
                    └──────────────┘
```

### Implementation

```
Write side (PostgreSQL — normalized):
  users:  { id, name, email }
  orders: { id, user_id, status, total }
  items:  { id, order_id, product_id, quantity, price }

Read side (Elasticsearch or denormalized view — optimized for queries):
  order_view: {
    id, user_name, user_email, status, total,
    items: [{ product_name, quantity, price }],
    created_at, shipped_at
  }

Projection (event handler that updates read model):
  on OrderCreated → insert into order_view
  on OrderShipped → update order_view set shipped_at
  on ItemAdded    → append to order_view.items
```

### CQRS + Event Sourcing (Combined)

```
Command → Validate → Store Events → Publish Events → Update Read Model

User sends: "Place order for items A, B, C"
  1. Command: PlaceOrder { items: [A, B, C] }
  2. Validate: Check inventory, validate user
  3. Store events: OrderPlaced, ItemsReserved
  4. Publish to Kafka
  5. Projections update read models:
     - Order view (for API queries)
     - Analytics view (for dashboards)
     - Search index (for search)
```

### When to Use CQRS

```
Good fit:
  - Read and write patterns are very different
  - Read-heavy with complex queries
  - Different scaling needs for reads vs writes
  - Multiple read representations needed (search, analytics, API)
  - Event sourcing already in use

Bad fit:
  - Simple CRUD (massive over-engineering)
  - Strong consistency required between read and write (CQRS has eventual consistency)
  - Small team / simple domain
```

---

## Outbox Pattern

Ensures reliable event publishing alongside database writes. Solves the dual-write problem (writing to DB AND publishing to Kafka — what if one fails?).

```
Problem (Dual Write):
  1. Write to database    ✓ Success
  2. Publish to Kafka     ✗ Failure
  Result: Database updated but event never published → inconsistency

Solution (Outbox):
  1. Write to database AND outbox table in same transaction
  2. Separate process reads outbox and publishes to Kafka
  3. Mark outbox entry as published
```

### Implementation

```sql
-- Outbox table (same database as business data)
CREATE TABLE outbox (
  id bigserial PRIMARY KEY,
  aggregate_type varchar(100) NOT NULL,
  aggregate_id varchar(100) NOT NULL,
  event_type varchar(100) NOT NULL,
  payload jsonb NOT NULL,
  created_at timestamptz DEFAULT now(),
  published_at timestamptz,  -- NULL until published
  retries integer DEFAULT 0
);

CREATE INDEX idx_outbox_unpublished ON outbox(created_at)
  WHERE published_at IS NULL;
```

```javascript
// Write business data AND outbox in same transaction
async function placeOrder(orderData) {
  await db.transaction(async (tx) => {
    // 1. Insert order
    const order = await tx.query(
      'INSERT INTO orders (user_id, total, status) VALUES ($1, $2, $3) RETURNING id',
      [orderData.userId, orderData.total, 'placed']
    );

    // 2. Insert outbox event (same transaction!)
    await tx.query(
      `INSERT INTO outbox (aggregate_type, aggregate_id, event_type, payload)
       VALUES ($1, $2, $3, $4)`,
      ['Order', order.id, 'OrderPlaced', JSON.stringify({
        order_id: order.id,
        user_id: orderData.userId,
        total: orderData.total,
      })]
    );
  });
  // Both succeed or both fail — atomic!
}
```

### Outbox Publisher

```
Polling (simple):
  Every 1-5 seconds:
    SELECT * FROM outbox WHERE published_at IS NULL ORDER BY created_at LIMIT 100;
    For each event: publish to Kafka, then UPDATE outbox SET published_at = now()

CDC (Change Data Capture — better):
  Use Debezium to stream outbox table changes to Kafka
  No polling overhead, near-real-time
  Debezium reads PostgreSQL WAL → publishes to Kafka topic

  ┌──────────┐     WAL      ┌──────────┐     ┌───────┐
  │PostgreSQL│──────────────►│ Debezium │────►│ Kafka │
  │ (outbox) │               │ (CDC)    │     │       │
  └──────────┘               └──────────┘     └───────┘
```

---

## Two-Phase Commit (2PC)

Coordinates a distributed transaction across multiple participants. All commit or all abort — no partial commits.

```
Phase 1: Prepare
  Coordinator → Participant A: "Prepare to commit"
  Coordinator → Participant B: "Prepare to commit"
  Participant A → Coordinator: "Yes, ready"
  Participant B → Coordinator: "Yes, ready"

Phase 2: Commit (if all said yes)
  Coordinator → Participant A: "Commit"
  Coordinator → Participant B: "Commit"
  Done!

Phase 2: Abort (if any said no)
  Coordinator → Participant A: "Abort"
  Coordinator → Participant B: "Abort"
  Rolled back!
```

### Problems with 2PC

```
1. Blocking: If coordinator crashes after Phase 1, participants are stuck
   (holding locks, waiting for commit/abort decision)

2. Single point of failure: Coordinator is critical

3. Performance: Holding locks across prepare → commit increases latency
   and reduces throughput

4. Not partition-tolerant: Network partition can leave participants in
   uncertain state

When to use 2PC:
  - Within a single datacenter (low latency, reliable network)
  - Database-level distributed transactions (XA transactions)
  - When you need ACID across databases and can tolerate the overhead

When NOT to use 2PC:
  - Across datacenters (latency too high)
  - In microservices (use saga instead)
  - When availability > consistency
```

---

## CRDTs (Conflict-Free Replicated Data Types)

Data structures that can be replicated across multiple nodes and merged automatically without coordination. Always converge to the same state.

### Common CRDTs

```
G-Counter (Grow-only Counter):
  Each node maintains its own counter. Total = sum of all nodes.

  Node A: 5    Node B: 3    Node C: 7
  Total: 5 + 3 + 7 = 15

  Merge: take max of each node's counter
  Node A sees: {A:5, B:2, C:7}
  Node B sees: {A:4, B:3, C:6}
  Merged:      {A:5, B:3, C:7} → total: 15

PN-Counter (Positive-Negative Counter):
  Two G-Counters: one for increments, one for decrements
  Value = sum(increments) - sum(decrements)

  Use for: Like counts, inventory counts, any counter that can go up and down

G-Set (Grow-only Set):
  Elements can be added, never removed.
  Merge: union of sets

  Node A: {a, b, c}
  Node B: {b, c, d}
  Merged: {a, b, c, d}

OR-Set (Observed-Remove Set):
  Elements can be added AND removed.
  Each add has a unique tag. Remove removes specific tags.
  "Add wins" semantics: if add and remove are concurrent, element stays.

  Use for: Shopping carts, collaborator lists, tag sets

LWW-Register (Last-Writer-Wins Register):
  Single value with timestamp. Latest timestamp wins.

  Node A: { value: "Alice", ts: 100 }
  Node B: { value: "Bob", ts: 105 }
  Merged: { value: "Bob", ts: 105 }   ← Bob wins (later timestamp)

  Use for: User profile fields, configuration values

LWW-Map (Last-Writer-Wins Map):
  Each key is an LWW-Register.

  Use for: Document fields, user preferences, settings
```

### When to Use CRDTs

```
Good fit:
  - Multi-region writes (active-active replication)
  - Offline-capable applications (mobile, edge)
  - Collaborative editing (multiplayer features)
  - Counters that must be available under partition
  - Shopping carts, presence indicators

Bad fit:
  - When you need strict ordering (use consensus instead)
  - When merge semantics are complex (CRDTs can't express arbitrary business logic)
  - Simple applications that don't need multi-writer
```

### CRDTs in Practice

```
Redis CRDTs (Redis Enterprise):
  Geo-replicated counters, sets, strings with automatic conflict resolution

Riak:
  Built-in CRDT support (counters, sets, maps, flags, registers)

Automerge / Yjs:
  Libraries for building collaborative applications
  CRDT-based document editing (like Google Docs)

Amazon DynamoDB:
  Not CRDTs per se, but LWW conflict resolution by default
  Can use CRDTs on top for richer merge semantics
```

---

## Idempotent Receiver Pattern

Ensures that processing a message multiple times has the same effect as processing it once. Critical for at-least-once delivery systems.

```
Implementation:
  1. Each message has a unique message_id
  2. Before processing, check if message_id was already processed
  3. If yes → skip (return previous result)
  4. If no → process, store message_id, return result

CREATE TABLE processed_messages (
  message_id varchar(100) PRIMARY KEY,
  result jsonb,
  processed_at timestamptz DEFAULT now()
);

-- Check and process atomically
INSERT INTO processed_messages (message_id, result)
VALUES ($1, $2)
ON CONFLICT (message_id) DO NOTHING;
-- If inserted: new message, process it
-- If conflict: already processed, skip
```

---

## Transactional Outbox + CDC Pipeline

Complete pattern for reliable event-driven architecture:

```
┌──────────────────────────────────────────────────────────────┐
│                     Service A                                 │
│                                                               │
│  ┌─────────────┐  Same TX  ┌──────────────┐                 │
│  │ Business DB  │◄─────────│ Application  │                 │
│  │ (orders)     │          │              │                 │
│  ├─────────────┤          └──────────────┘                 │
│  │ Outbox Table │                                            │
│  └──────┬──────┘                                            │
│         │ WAL                                                │
│  ┌──────▼──────┐                                            │
│  │  Debezium   │                                            │
│  │  (CDC)      │                                            │
│  └──────┬──────┘                                            │
└─────────┼────────────────────────────────────────────────────┘
          │
   ┌──────▼──────┐
   │    Kafka    │
   │ (events)   │
   └──────┬──────┘
          │
   ┌──────┴──────────────────┐
   │                         │
┌──▼──────────┐  ┌──────────▼──┐
│ Service B   │  │ Service C    │
│ (consumer)  │  │ (consumer)   │
│ + idempotent│  │ + idempotent │
│   receiver  │  │   receiver   │
└─────────────┘  └──────────────┘
```

This combination provides:
1. **Atomicity**: Business write + event in same transaction (outbox)
2. **Reliability**: CDC captures all events from WAL (no polling gaps)
3. **At-least-once delivery**: Kafka retries on consumer failure
4. **Exactly-once processing**: Idempotent receivers deduplicate
5. **Ordering**: Events ordered by outbox table sequence within aggregate

---

## Pattern Selection Guide

```
Need                                    Pattern
──────────────────────────────────────────────────────────────
Distributed transaction (simple)        Saga (choreography)
Distributed transaction (complex)       Saga (orchestration)
Complete audit trail                    Event sourcing
Separate read/write optimization        CQRS
Reliable event publishing               Outbox pattern
Automatic conflict resolution           CRDTs
Strong consistency across nodes         Two-phase commit (2PC)
Exactly-once processing                 Idempotent receiver
Real-time data sync                     CDC (change data capture)
```
