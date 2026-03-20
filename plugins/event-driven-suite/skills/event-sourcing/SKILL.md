---
name: event-sourcing
description: >
  Event sourcing implementation — event stores, aggregates, projections,
  snapshots, event replay, schema evolution, and production event-sourced
  systems with TypeScript examples.
  Triggers: "event sourcing", "event store", "aggregate", "projection",
  "event replay", "snapshot", "domain events".
  NOT for: CQRS read models (use cqrs-patterns), message queues (use message-queues-suite).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Event Sourcing Implementation

## Core Concepts

### Event Store (PostgreSQL)

```sql
CREATE TABLE events (
  id BIGSERIAL PRIMARY KEY,
  aggregate_id UUID NOT NULL,
  aggregate_type VARCHAR(100) NOT NULL,
  event_type VARCHAR(100) NOT NULL,
  event_data JSONB NOT NULL,
  metadata JSONB DEFAULT '{}',
  version INT NOT NULL,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  UNIQUE(aggregate_id, version)  -- Optimistic concurrency
);

CREATE INDEX idx_events_aggregate ON events(aggregate_id, version);
CREATE INDEX idx_events_type ON events(event_type);
CREATE INDEX idx_events_created ON events(created_at);

-- Snapshots for performance
CREATE TABLE snapshots (
  aggregate_id UUID PRIMARY KEY,
  aggregate_type VARCHAR(100) NOT NULL,
  version INT NOT NULL,
  state JSONB NOT NULL,
  created_at TIMESTAMPTZ DEFAULT NOW()
);
```

### Domain Events

```typescript
// Base event interface
interface DomainEvent {
  eventId: string;
  eventType: string;
  aggregateId: string;
  aggregateType: string;
  version: number;
  timestamp: string;
  correlationId: string;
  causationId: string;
  metadata: Record<string, any>;
  payload: Record<string, any>;
}

// Typed events for an Order aggregate
type OrderEvent =
  | { eventType: "OrderCreated"; payload: { customerId: string; items: OrderItem[] } }
  | { eventType: "ItemAdded"; payload: { item: OrderItem } }
  | { eventType: "ItemRemoved"; payload: { itemId: string } }
  | { eventType: "OrderConfirmed"; payload: { confirmedAt: string } }
  | { eventType: "OrderShipped"; payload: { trackingNumber: string; carrier: string } }
  | { eventType: "OrderCancelled"; payload: { reason: string } };

interface OrderItem {
  productId: string;
  name: string;
  quantity: number;
  priceUsd: number;
}
```

### Aggregate

```typescript
// Order aggregate
interface OrderState {
  id: string;
  customerId: string;
  items: OrderItem[];
  status: "draft" | "confirmed" | "shipped" | "cancelled";
  totalUsd: number;
  version: number;
}

class OrderAggregate {
  private state: OrderState;
  private uncommittedEvents: DomainEvent[] = [];

  constructor(id: string) {
    this.state = {
      id,
      customerId: "",
      items: [],
      status: "draft",
      totalUsd: 0,
      version: 0,
    };
  }

  // ---- Commands (validate business rules, emit events) ----

  create(customerId: string, items: OrderItem[]) {
    if (this.state.version > 0) {
      throw new Error("Order already exists");
    }
    if (items.length === 0) {
      throw new Error("Order must have at least one item");
    }
    this.apply({
      eventType: "OrderCreated",
      payload: { customerId, items },
    });
  }

  addItem(item: OrderItem) {
    if (this.state.status !== "draft") {
      throw new Error("Can only add items to draft orders");
    }
    if (this.state.items.some((i) => i.productId === item.productId)) {
      throw new Error("Item already in order");
    }
    this.apply({
      eventType: "ItemAdded",
      payload: { item },
    });
  }

  confirm() {
    if (this.state.status !== "draft") {
      throw new Error("Can only confirm draft orders");
    }
    if (this.state.items.length === 0) {
      throw new Error("Cannot confirm empty order");
    }
    this.apply({
      eventType: "OrderConfirmed",
      payload: { confirmedAt: new Date().toISOString() },
    });
  }

  cancel(reason: string) {
    if (this.state.status === "shipped") {
      throw new Error("Cannot cancel shipped orders");
    }
    if (this.state.status === "cancelled") {
      throw new Error("Order already cancelled");
    }
    this.apply({
      eventType: "OrderCancelled",
      payload: { reason },
    });
  }

  // ---- Event application (mutate state, no validation) ----

  private apply(event: Partial<DomainEvent>) {
    const fullEvent: DomainEvent = {
      eventId: crypto.randomUUID(),
      eventType: event.eventType!,
      aggregateId: this.state.id,
      aggregateType: "Order",
      version: this.state.version + 1,
      timestamp: new Date().toISOString(),
      correlationId: event.correlationId ?? crypto.randomUUID(),
      causationId: event.causationId ?? "",
      metadata: event.metadata ?? {},
      payload: event.payload ?? {},
    };

    this.mutate(fullEvent);
    this.uncommittedEvents.push(fullEvent);
  }

  // Pure state mutation — called for both new events AND replay
  private mutate(event: DomainEvent) {
    switch (event.eventType) {
      case "OrderCreated":
        this.state.customerId = event.payload.customerId;
        this.state.items = event.payload.items;
        this.state.totalUsd = this.calculateTotal(event.payload.items);
        this.state.status = "draft";
        break;

      case "ItemAdded":
        this.state.items.push(event.payload.item);
        this.state.totalUsd = this.calculateTotal(this.state.items);
        break;

      case "ItemRemoved":
        this.state.items = this.state.items.filter(
          (i) => i.productId !== event.payload.itemId
        );
        this.state.totalUsd = this.calculateTotal(this.state.items);
        break;

      case "OrderConfirmed":
        this.state.status = "confirmed";
        break;

      case "OrderShipped":
        this.state.status = "shipped";
        break;

      case "OrderCancelled":
        this.state.status = "cancelled";
        break;
    }
    this.state.version = event.version;
  }

  private calculateTotal(items: OrderItem[]): number {
    return items.reduce((sum, i) => sum + i.priceUsd * i.quantity, 0);
  }

  // ---- Hydration ----

  static fromEvents(id: string, events: DomainEvent[]): OrderAggregate {
    const aggregate = new OrderAggregate(id);
    for (const event of events) {
      aggregate.mutate(event);
    }
    return aggregate;
  }

  static fromSnapshot(snapshot: OrderState, events: DomainEvent[]): OrderAggregate {
    const aggregate = new OrderAggregate(snapshot.id);
    aggregate.state = { ...snapshot };
    for (const event of events) {
      aggregate.mutate(event);
    }
    return aggregate;
  }

  getUncommittedEvents(): DomainEvent[] {
    return [...this.uncommittedEvents];
  }

  clearUncommittedEvents() {
    this.uncommittedEvents = [];
  }

  getState(): Readonly<OrderState> {
    return { ...this.state };
  }
}
```

### Event Store Repository

```typescript
class EventStoreRepository {
  constructor(private pool: Pool) {}

  async save(aggregate: OrderAggregate): Promise<void> {
    const events = aggregate.getUncommittedEvents();
    if (events.length === 0) return;

    const client = await this.pool.connect();
    try {
      await client.query("BEGIN");

      for (const event of events) {
        await client.query(
          `INSERT INTO events (aggregate_id, aggregate_type, event_type, event_data, metadata, version)
           VALUES ($1, $2, $3, $4, $5, $6)`,
          [
            event.aggregateId,
            event.aggregateType,
            event.eventType,
            JSON.stringify(event.payload),
            JSON.stringify(event.metadata),
            event.version,
          ]
        );
      }

      await client.query("COMMIT");
      aggregate.clearUncommittedEvents();
    } catch (error: any) {
      await client.query("ROLLBACK");
      if (error.code === "23505") {
        // Unique violation on (aggregate_id, version) = concurrency conflict
        throw new ConcurrencyError(
          `Aggregate ${events[0].aggregateId} was modified concurrently`
        );
      }
      throw error;
    } finally {
      client.release();
    }
  }

  async load(aggregateId: string): Promise<OrderAggregate> {
    // Check for snapshot first
    const snapshot = await this.pool.query(
      "SELECT * FROM snapshots WHERE aggregate_id = $1",
      [aggregateId]
    );

    let fromVersion = 0;
    let aggregate: OrderAggregate;

    if (snapshot.rows.length > 0) {
      const snap = snapshot.rows[0];
      aggregate = OrderAggregate.fromSnapshot(snap.state, []);
      fromVersion = snap.version;
    } else {
      aggregate = new OrderAggregate(aggregateId);
    }

    // Load events after snapshot
    const result = await this.pool.query(
      `SELECT * FROM events
       WHERE aggregate_id = $1 AND version > $2
       ORDER BY version ASC`,
      [aggregateId, fromVersion]
    );

    if (result.rows.length === 0 && fromVersion === 0) {
      throw new AggregateNotFoundError(aggregateId);
    }

    const events = result.rows.map(this.rowToEvent);
    return OrderAggregate.fromEvents(aggregateId, events);
  }

  async saveSnapshot(aggregate: OrderAggregate): Promise<void> {
    const state = aggregate.getState();
    await this.pool.query(
      `INSERT INTO snapshots (aggregate_id, aggregate_type, version, state)
       VALUES ($1, $2, $3, $4)
       ON CONFLICT (aggregate_id) DO UPDATE SET
         version = EXCLUDED.version,
         state = EXCLUDED.state,
         created_at = NOW()`,
      [state.id, "Order", state.version, JSON.stringify(state)]
    );
  }

  private rowToEvent(row: any): DomainEvent {
    return {
      eventId: row.id.toString(),
      eventType: row.event_type,
      aggregateId: row.aggregate_id,
      aggregateType: row.aggregate_type,
      version: row.version,
      timestamp: row.created_at.toISOString(),
      correlationId: row.metadata?.correlationId ?? "",
      causationId: row.metadata?.causationId ?? "",
      metadata: row.metadata ?? {},
      payload: row.event_data,
    };
  }
}

class ConcurrencyError extends Error {}
class AggregateNotFoundError extends Error {}
```

## Projections (Read Models)

```typescript
// Projection that builds a read model from events
class OrderListProjection {
  constructor(private pool: Pool) {}

  async handleEvent(event: DomainEvent): Promise<void> {
    switch (event.eventType) {
      case "OrderCreated":
        await this.pool.query(
          `INSERT INTO order_list (id, customer_id, status, total_usd, item_count, created_at)
           VALUES ($1, $2, 'draft', $3, $4, $5)`,
          [
            event.aggregateId,
            event.payload.customerId,
            event.payload.items.reduce((s: number, i: any) => s + i.priceUsd * i.quantity, 0),
            event.payload.items.length,
            event.timestamp,
          ]
        );
        break;

      case "ItemAdded":
        await this.pool.query(
          `UPDATE order_list SET
            item_count = item_count + 1,
            total_usd = total_usd + $2
           WHERE id = $1`,
          [event.aggregateId, event.payload.item.priceUsd * event.payload.item.quantity]
        );
        break;

      case "OrderConfirmed":
        await this.pool.query(
          "UPDATE order_list SET status = 'confirmed' WHERE id = $1",
          [event.aggregateId]
        );
        break;

      case "OrderCancelled":
        await this.pool.query(
          "UPDATE order_list SET status = 'cancelled' WHERE id = $1",
          [event.aggregateId]
        );
        break;
    }
  }

  // Rebuild projection from scratch
  async rebuild(): Promise<void> {
    await this.pool.query("TRUNCATE order_list");
    const events = await this.pool.query(
      "SELECT * FROM events WHERE aggregate_type = 'Order' ORDER BY id ASC"
    );
    for (const row of events.rows) {
      await this.handleEvent(this.rowToEvent(row));
    }
  }
}
```

## Schema Evolution

```typescript
// Event upcaster — transforms old event versions to current
interface EventUpcaster {
  eventType: string;
  fromVersion: number;
  toVersion: number;
  upcast(payload: any): any;
}

const upcasters: EventUpcaster[] = [
  {
    eventType: "OrderCreated",
    fromVersion: 1,
    toVersion: 2,
    upcast(payload) {
      // v1 had "amount" as a single number, v2 has "items" array
      return {
        ...payload,
        items: payload.items ?? [
          { productId: "unknown", name: "Legacy item", quantity: 1, priceUsd: payload.amount ?? 0 },
        ],
      };
    },
  },
  {
    eventType: "OrderShipped",
    fromVersion: 1,
    toVersion: 2,
    upcast(payload) {
      // v1 didn't have carrier field
      return {
        ...payload,
        carrier: payload.carrier ?? "unknown",
      };
    },
  },
];

function upcastEvent(event: DomainEvent): DomainEvent {
  let payload = event.payload;
  const version = event.metadata?.schemaVersion ?? 1;

  for (const upcaster of upcasters) {
    if (
      upcaster.eventType === event.eventType &&
      upcaster.fromVersion >= version
    ) {
      payload = upcaster.upcast(payload);
    }
  }

  return { ...event, payload };
}
```

## Outbox Pattern (Reliable Publishing)

```typescript
// Save events AND outbox entries in one transaction
async function saveWithOutbox(
  client: PoolClient,
  aggregate: OrderAggregate
): Promise<void> {
  const events = aggregate.getUncommittedEvents();

  for (const event of events) {
    // Save to event store
    await client.query(
      `INSERT INTO events (aggregate_id, aggregate_type, event_type, event_data, metadata, version)
       VALUES ($1, $2, $3, $4, $5, $6)`,
      [event.aggregateId, event.aggregateType, event.eventType, event.payload, event.metadata, event.version]
    );

    // Save to outbox (same transaction!)
    await client.query(
      `INSERT INTO outbox (event_id, event_type, payload, created_at)
       VALUES ($1, $2, $3, NOW())`,
      [event.eventId, event.eventType, JSON.stringify(event)]
    );
  }
}

// Outbox publisher — polls and publishes
async function publishOutbox(pool: Pool, publisher: EventPublisher) {
  while (true) {
    const client = await pool.connect();
    try {
      await client.query("BEGIN");

      // Lock and fetch unpublished events
      const result = await client.query(
        `SELECT * FROM outbox
         WHERE published_at IS NULL
         ORDER BY created_at ASC
         LIMIT 100
         FOR UPDATE SKIP LOCKED`
      );

      for (const row of result.rows) {
        await publisher.publish(row.event_type, JSON.parse(row.payload));
        await client.query(
          "UPDATE outbox SET published_at = NOW() WHERE id = $1",
          [row.id]
        );
      }

      await client.query("COMMIT");
    } catch (error) {
      await client.query("ROLLBACK");
      console.error("Outbox publish error:", error);
    } finally {
      client.release();
    }

    await new Promise((r) => setTimeout(r, 1000)); // Poll interval
  }
}
```

## Gotchas

1. **Snapshots are optimization, not requirement** — Start without snapshots. Add them when replay gets slow (> 1000 events per aggregate). Snapshot every N events (e.g., every 100).

2. **Event store is append-only** — Never update or delete events. If you need to "undo" something, emit a compensating event (OrderCancelled, not DELETE FROM events).

3. **Projections are disposable** — Read models can be rebuilt from events at any time. This means projection bugs are fixable: fix the code, rebuild, done.

4. **Concurrency via optimistic locking** — The UNIQUE(aggregate_id, version) constraint catches concurrent writes. Catch the constraint violation, reload, and retry.

5. **Don't put business logic in event handlers** — Event handlers (projections) should only transform and store data. Business rules belong in the aggregate's command methods.

6. **Global event ordering is expensive** — Within an aggregate, events are strictly ordered. Across aggregates, use timestamps + correlation IDs for causal ordering, not global sequence numbers.
