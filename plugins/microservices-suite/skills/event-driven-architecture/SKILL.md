---
name: event-driven-architecture
description: >
  Event-driven microservices — message brokers, event sourcing, CQRS, saga
  patterns, eventual consistency, and async communication patterns.
  Triggers: "event driven", "message broker", "event sourcing", "cqrs",
  "saga pattern", "rabbitmq", "kafka", "pub sub", "eventual consistency".
  NOT for: API gateway or HTTP patterns (use api-gateway-patterns).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Event-Driven Architecture

## Message Broker Patterns

### Publish/Subscribe

```typescript
// RabbitMQ pub/sub with amqplib
import amqp from "amqplib";

class EventBus {
  private connection: amqp.Connection | null = null;
  private channel: amqp.Channel | null = null;

  async connect(url: string) {
    this.connection = await amqp.connect(url);
    this.channel = await this.connection.createChannel();
    // Prefetch limits concurrent processing
    await this.channel.prefetch(10);
  }

  async publish(exchange: string, routingKey: string, event: object) {
    await this.channel!.assertExchange(exchange, "topic", { durable: true });
    this.channel!.publish(
      exchange, routingKey,
      Buffer.from(JSON.stringify({
        id: crypto.randomUUID(),
        type: routingKey,
        timestamp: new Date().toISOString(),
        data: event,
      })),
      { persistent: true, contentType: "application/json" }
    );
  }

  async subscribe(exchange: string, routingKey: string, handler: (msg: any) => Promise<void>) {
    await this.channel!.assertExchange(exchange, "topic", { durable: true });
    const { queue } = await this.channel!.assertQueue("", { exclusive: true });
    await this.channel!.bindQueue(queue, exchange, routingKey);

    this.channel!.consume(queue, async (msg) => {
      if (!msg) return;
      try {
        const event = JSON.parse(msg.content.toString());
        await handler(event);
        this.channel!.ack(msg);
      } catch (error) {
        // Nack with requeue=false sends to dead letter queue
        this.channel!.nack(msg, false, false);
      }
    });
  }
}

// Usage
const bus = new EventBus();
await bus.connect(process.env.RABBITMQ_URL!);

// Publisher (order service)
await bus.publish("orders", "order.created", {
  orderId: "123", userId: "456", total: 99.99,
});

// Subscriber (notification service)
await bus.subscribe("orders", "order.*", async (event) => {
  if (event.type === "order.created") {
    await sendConfirmationEmail(event.data.userId, event.data.orderId);
  }
});
```

### Work Queue (Competing Consumers)

```typescript
// Multiple workers processing from same queue
async function createWorkerQueue(queueName: string, handler: (data: any) => Promise<void>) {
  const conn = await amqp.connect(process.env.RABBITMQ_URL!);
  const ch = await conn.createChannel();

  await ch.assertQueue(queueName, {
    durable: true,
    deadLetterExchange: "dlx",
    deadLetterRoutingKey: `${queueName}.dead`,
  });

  // Process one at a time per worker
  await ch.prefetch(1);

  ch.consume(queueName, async (msg) => {
    if (!msg) return;
    try {
      const data = JSON.parse(msg.content.toString());
      await handler(data);
      ch.ack(msg);
    } catch (error) {
      const retries = (msg.properties.headers?.["x-retry-count"] || 0) + 1;
      if (retries < 3) {
        // Retry with backoff
        ch.publish("", queueName, msg.content, {
          headers: { "x-retry-count": retries },
          persistent: true,
        });
      }
      ch.nack(msg, false, false); // Send to DLQ
    }
  });
}
```

## Event Sourcing

```typescript
// Event store
interface DomainEvent {
  id: string;
  aggregateId: string;
  aggregateType: string;
  type: string;
  data: Record<string, unknown>;
  version: number;
  timestamp: Date;
}

class EventStore {
  async append(events: DomainEvent[]): Promise<void> {
    await db.transaction(async (tx) => {
      for (const event of events) {
        // Optimistic concurrency — fail if version conflict
        const current = await tx.query(
          "SELECT MAX(version) as v FROM events WHERE aggregate_id = $1",
          [event.aggregateId]
        );
        if (current.rows[0]?.v >= event.version) {
          throw new Error("Concurrency conflict");
        }
        await tx.query(
          "INSERT INTO events (id, aggregate_id, aggregate_type, type, data, version, timestamp) VALUES ($1,$2,$3,$4,$5,$6,$7)",
          [event.id, event.aggregateId, event.aggregateType, event.type, event.data, event.version, event.timestamp]
        );
      }
    });
  }

  async getEvents(aggregateId: string): Promise<DomainEvent[]> {
    const { rows } = await db.query(
      "SELECT * FROM events WHERE aggregate_id = $1 ORDER BY version",
      [aggregateId]
    );
    return rows;
  }
}

// Aggregate rebuilt from events
class Order {
  id: string = "";
  status: string = "pending";
  items: OrderItem[] = [];
  total: number = 0;

  static fromEvents(events: DomainEvent[]): Order {
    const order = new Order();
    for (const event of events) {
      order.apply(event);
    }
    return order;
  }

  private apply(event: DomainEvent) {
    switch (event.type) {
      case "OrderCreated":
        this.id = event.aggregateId;
        this.status = "pending";
        this.items = event.data.items as OrderItem[];
        this.total = event.data.total as number;
        break;
      case "OrderPaid":
        this.status = "paid";
        break;
      case "OrderShipped":
        this.status = "shipped";
        break;
      case "OrderCancelled":
        this.status = "cancelled";
        break;
    }
  }
}
```

## CQRS (Command Query Responsibility Segregation)

```typescript
// Write side — commands
interface Command { type: string; payload: Record<string, unknown>; }

class CommandHandler {
  async handle(command: Command): Promise<void> {
    switch (command.type) {
      case "CreateOrder": {
        const events = [
          { type: "OrderCreated", data: command.payload, version: 1 },
        ];
        await eventStore.append(events);
        await eventBus.publish("orders", "order.created", command.payload);
        break;
      }
    }
  }
}

// Read side — projections (denormalized read models)
class OrderProjection {
  async handleEvent(event: DomainEvent): Promise<void> {
    switch (event.type) {
      case "OrderCreated":
        await db.query(
          "INSERT INTO order_summary (id, user_id, total, status, item_count, created_at) VALUES ($1,$2,$3,$4,$5,$6)",
          [event.aggregateId, event.data.userId, event.data.total, "pending", (event.data.items as any[]).length, event.timestamp]
        );
        break;
      case "OrderShipped":
        await db.query(
          "UPDATE order_summary SET status = 'shipped', shipped_at = $2 WHERE id = $1",
          [event.aggregateId, event.timestamp]
        );
        break;
    }
  }
}

// Query side — optimized reads
class OrderQuery {
  async getUserOrders(userId: string, page: number) {
    return db.query(
      "SELECT * FROM order_summary WHERE user_id = $1 ORDER BY created_at DESC LIMIT 20 OFFSET $2",
      [userId, (page - 1) * 20]
    );
  }
}
```

## Saga Pattern (Distributed Transactions)

```typescript
// Orchestration-based saga
class OrderSaga {
  async execute(orderId: string, userId: string, items: OrderItem[]) {
    const steps: SagaStep[] = [
      { action: () => paymentService.charge(userId, total), compensate: () => paymentService.refund(userId, total) },
      { action: () => inventoryService.reserve(items),      compensate: () => inventoryService.release(items) },
      { action: () => shippingService.create(orderId),      compensate: () => shippingService.cancel(orderId) },
    ];

    const completed: SagaStep[] = [];

    for (const step of steps) {
      try {
        await step.action();
        completed.push(step);
      } catch (error) {
        // Compensate in reverse order
        for (const done of completed.reverse()) {
          try { await done.compensate(); }
          catch (compError) {
            logger.error("Compensation failed:", compError);
            // Manual intervention needed
            await alertOps(orderId, compError);
          }
        }
        throw error;
      }
    }
  }
}
```

## Communication Patterns

| Pattern | Use When | Example |
|---------|----------|---------|
| Request/Reply | Need immediate response | GET /api/orders/123 |
| Async Request/Reply | Need response but can wait | Submit job → poll status |
| Fire and Forget | Don't need confirmation | Log event, send analytics |
| Pub/Sub | Multiple consumers need same event | Order created → notify, invoice, ship |
| Work Queue | Distribute load across workers | Process uploaded images |

## Gotchas

1. **Message ordering is not guaranteed** — Most brokers deliver messages in approximate order, but failures, retries, and competing consumers break ordering. Design handlers to be idempotent and order-independent. Use sequence numbers if ordering matters.

2. **Exactly-once delivery is a myth** — Brokers guarantee at-least-once or at-most-once. Your consumers must handle duplicates. Use an idempotency key (event ID) and check before processing: `INSERT ... ON CONFLICT (event_id) DO NOTHING`.

3. **Event sourcing read performance** — Rebuilding an aggregate from 10,000 events is slow. Use snapshots: periodically save the current state and replay only events after the snapshot.

4. **Saga compensation is hard** — When step 3 of 5 fails, compensating steps 1-2 may also fail. You need a dead letter queue and manual intervention workflow for unrecoverable compensation failures.

5. **CQRS eventual consistency confuses users** — After creating an order (write), the read model may not be updated yet. Users see "order not found" immediately after creating it. Use read-your-writes: return the created entity from the write side.

6. **Dead letter queues fill up silently** — Failed messages go to DLQ, but nobody monitors it. Set up alerts on DLQ depth. Process DLQ messages periodically — they represent real failures that need investigation.
