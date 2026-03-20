# Event-Driven Architecture Cheatsheet

## Pattern Decision

```
Simple decoupling between services?       → Domain Events (pub/sub)
Need full audit trail / undo?             → Event Sourcing
Read-heavy with complex queries?          → CQRS (separate read model)
Multi-service transaction?                → Saga (orchestration or choreography)
Reliable event publishing from DB?        → Outbox Pattern
Sync legacy systems?                      → Change Data Capture (CDC)
```

## Event Schema

```typescript
interface DomainEvent {
  eventId: string;          // UUID, unique
  eventType: string;        // "OrderCreated"
  aggregateId: string;      // Entity ID
  aggregateType: string;    // "Order"
  version: number;          // Schema version
  timestamp: string;        // ISO-8601
  correlationId: string;    // Traces workflow
  causationId: string;      // What caused this
  payload: Record<string, any>;
}
```

## Event Sourcing Aggregate

```
Command → Validate Rules → Emit Event → Mutate State
                                            ↑
                                    Replay from events
```

```typescript
class Aggregate {
  // Commands: validate business rules, call apply()
  create(data) { if (exists) throw; this.apply("Created", data); }

  // Apply: create full event, call mutate, add to uncommitted
  private apply(type, payload) { this.mutate(event); this.uncommitted.push(event); }

  // Mutate: pure state change (used by both apply AND replay)
  private mutate(event) { switch(event.type) { ... } }

  // Hydrate from events
  static fromEvents(events) { for (e of events) agg.mutate(e); }
}
```

## CQRS Architecture

```
Write Side:                          Read Side:
  Command → Handler → Aggregate       Event → Projection → Read Model
              ↓                                                ↓
         Event Store                                      Query Handler
                                                               ↓
                                                           API Response
```

Commands modify. Queries never modify. Different databases OK.

## Saga Pattern

### Orchestration (Central)
```
Saga Orchestrator
  → Step 1 → success → Step 2 → success → Step 3 → FAIL
  ← Compensate 2 ← Compensate 1
```

### Choreography (Decentralized)
```
OrderCreated → InventoryService (listen) → InventoryReserved
  → PaymentService (listen) → PaymentProcessed
  → ShippingService (listen) → Shipped
```

## Outbox Pattern

```sql
-- Same transaction: save to DB + save to outbox
BEGIN;
  INSERT INTO orders (...) VALUES (...);
  INSERT INTO outbox (event_type, payload) VALUES ('OrderCreated', '...');
COMMIT;

-- Separate process polls outbox and publishes to message broker
SELECT * FROM outbox WHERE published_at IS NULL FOR UPDATE SKIP LOCKED;
-- Publish → UPDATE outbox SET published_at = NOW()
```

## Idempotency

```typescript
// Before processing any event:
const already = await db.query(
  "SELECT 1 FROM processed_events WHERE event_id = $1 AND handler = $2",
  [eventId, handlerName]
);
if (already.rows.length > 0) return; // Skip duplicate

// Process, then mark as done (same transaction)
await handler();
await db.query(
  "INSERT INTO processed_events (event_id, handler) VALUES ($1, $2)",
  [eventId, handlerName]
);
```

## Eventual Consistency

```
Client sends command → 202 Accepted + correlationId
Client polls read model → 200 (ready) or 202 (still processing)

// Or: WebSocket/SSE push when projection catches up
```

## Event Store (PostgreSQL)

```sql
CREATE TABLE events (
  id BIGSERIAL PRIMARY KEY,
  aggregate_id UUID NOT NULL,
  aggregate_type VARCHAR(100),
  event_type VARCHAR(100) NOT NULL,
  event_data JSONB NOT NULL,
  version INT NOT NULL,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  UNIQUE(aggregate_id, version)  -- Optimistic concurrency
);
```

## Concurrency Control

```
UNIQUE(aggregate_id, version)
  → Concurrent writes get constraint violation
  → Catch error → reload aggregate → retry command
```

## Schema Evolution (Upcasting)

```typescript
// Old event v1: { amount: 100 }
// New event v2: { items: [{ price: 100, qty: 1 }] }

if (event.version === 1) {
  event.payload = {
    items: [{ price: event.payload.amount, qty: 1 }]
  };
}
```

Never change old events. Transform on read.

## Projection Rebuild

```
Projections are disposable:
  1. Fix the projection code
  2. TRUNCATE the read model table
  3. Replay all events from the beginning
  4. Read model is now correct
```

## Infrastructure

| Tool | Type | Best For |
|------|------|----------|
| Kafka | Streaming | High throughput, replay |
| RabbitMQ | Broker | Complex routing, RPC |
| EventStoreDB | Event store | Native event sourcing |
| PostgreSQL | Outbox | Small-medium, simple |
| Redis Streams | Streaming | Low latency |
| SQS + SNS | AWS | Serverless fan-out |

## Anti-Patterns

```
1. Dual Write      → Use outbox pattern (one transaction)
2. Fat Events      → Only changed fields + ID (not entire entity)
3. No Idempotency  → Duplicate delivery WILL happen
4. Sync Publishing → Async events should not block requests
5. No Versioning   → Events without version break consumers
6. No Correlation  → Can't trace workflows across services
7. Logic in Handlers → Put business rules in aggregates
8. Modifying Events → Event store is append-only
```

## Checklist

```
[ ] Events have: eventId, eventType, aggregateId, version, timestamp, correlationId
[ ] Idempotency keys on all handlers
[ ] Schema version on all events
[ ] Upcasters for old event versions
[ ] Outbox pattern (not dual write)
[ ] Dead letter queue for failed events
[ ] Projection checkpoint tracking
[ ] Compensation logic for sagas
[ ] Optimistic concurrency on aggregate save
[ ] Monitoring: event lag, processing time, DLQ depth
```
