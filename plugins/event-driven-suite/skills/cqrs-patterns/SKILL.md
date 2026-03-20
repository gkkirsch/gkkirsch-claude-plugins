---
name: cqrs-patterns
description: >
  CQRS (Command Query Responsibility Segregation) implementation — command
  handlers, query handlers, read model projections, eventual consistency,
  saga orchestration, and production CQRS with TypeScript.
  Triggers: "CQRS", "command handler", "query handler", "read model",
  "saga pattern", "process manager", "eventual consistency".
  NOT for: event store setup (use event-sourcing), message broker config.
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# CQRS Patterns

## Command Side

### Command Bus

```typescript
// Command definition
interface Command {
  type: string;
  payload: Record<string, any>;
  metadata: {
    correlationId: string;
    userId: string;
    timestamp: string;
  };
}

// Typed commands
interface CreateOrderCommand extends Command {
  type: "CreateOrder";
  payload: {
    customerId: string;
    items: { productId: string; quantity: number }[];
  };
}

interface ConfirmOrderCommand extends Command {
  type: "ConfirmOrder";
  payload: { orderId: string };
}

// Command handler registry
type CommandHandler<T extends Command = Command> = (command: T) => Promise<void>;

class CommandBus {
  private handlers = new Map<string, CommandHandler>();

  register<T extends Command>(type: string, handler: CommandHandler<T>) {
    if (this.handlers.has(type)) {
      throw new Error(`Handler already registered for ${type}`);
    }
    this.handlers.set(type, handler as CommandHandler);
  }

  async dispatch(command: Command): Promise<void> {
    const handler = this.handlers.get(command.type);
    if (!handler) {
      throw new Error(`No handler for command: ${command.type}`);
    }

    // Validate command
    await this.validate(command);

    // Execute with error handling
    try {
      await handler(command);
    } catch (error) {
      if (error instanceof ConcurrencyError) {
        // Retry once on optimistic concurrency conflict
        await handler(command);
      } else {
        throw error;
      }
    }
  }

  private async validate(command: Command) {
    if (!command.metadata?.correlationId) {
      throw new ValidationError("correlationId is required");
    }
    if (!command.metadata?.userId) {
      throw new ValidationError("userId is required");
    }
  }
}

// Command handler implementation
const createOrderHandler: CommandHandler<CreateOrderCommand> = async (cmd) => {
  const order = new OrderAggregate(crypto.randomUUID());

  // Validate items exist and have stock
  const items = await Promise.all(
    cmd.payload.items.map(async (item) => {
      const product = await productRepo.findById(item.productId);
      if (!product) throw new ValidationError(`Product ${item.productId} not found`);
      if (product.stock < item.quantity) throw new ValidationError(`Insufficient stock for ${product.name}`);
      return {
        productId: product.id,
        name: product.name,
        quantity: item.quantity,
        priceUsd: product.priceUsd,
      };
    })
  );

  order.create(cmd.payload.customerId, items);
  await eventStore.save(order);
};

// Register handlers
const bus = new CommandBus();
bus.register("CreateOrder", createOrderHandler);
bus.register("ConfirmOrder", confirmOrderHandler);
```

### Express API with Commands

```typescript
import { Router } from "express";

const router = Router();

router.post("/orders", async (req, res) => {
  const correlationId = req.headers["x-correlation-id"] as string ?? crypto.randomUUID();

  try {
    await commandBus.dispatch({
      type: "CreateOrder",
      payload: {
        customerId: req.body.customerId,
        items: req.body.items,
      },
      metadata: {
        correlationId,
        userId: req.user!.id,
        timestamp: new Date().toISOString(),
      },
    });

    // Return 202 Accepted — state change is async
    res.status(202).json({ correlationId });
  } catch (error) {
    if (error instanceof ValidationError) {
      res.status(400).json({ error: error.message });
    } else {
      res.status(500).json({ error: "Internal error" });
    }
  }
});

router.post("/orders/:id/confirm", async (req, res) => {
  await commandBus.dispatch({
    type: "ConfirmOrder",
    payload: { orderId: req.params.id },
    metadata: {
      correlationId: req.headers["x-correlation-id"] as string ?? crypto.randomUUID(),
      userId: req.user!.id,
      timestamp: new Date().toISOString(),
    },
  });
  res.status(202).json({ ok: true });
});
```

## Query Side

### Query Bus

```typescript
interface Query {
  type: string;
  params: Record<string, any>;
}

type QueryHandler<T extends Query = Query, R = any> = (query: T) => Promise<R>;

class QueryBus {
  private handlers = new Map<string, QueryHandler>();

  register<T extends Query, R>(type: string, handler: QueryHandler<T, R>) {
    this.handlers.set(type, handler as QueryHandler);
  }

  async execute<R>(query: Query): Promise<R> {
    const handler = this.handlers.get(query.type);
    if (!handler) throw new Error(`No handler for query: ${query.type}`);
    return handler(query);
  }
}

// Query handlers read from optimized read models
const getOrderListHandler: QueryHandler = async (query) => {
  const { customerId, status, page = 1, pageSize = 20 } = query.params;

  let sql = "SELECT * FROM order_list WHERE 1=1";
  const params: any[] = [];

  if (customerId) {
    params.push(customerId);
    sql += ` AND customer_id = $${params.length}`;
  }
  if (status) {
    params.push(status);
    sql += ` AND status = $${params.length}`;
  }

  sql += ` ORDER BY created_at DESC LIMIT $${params.length + 1} OFFSET $${params.length + 2}`;
  params.push(pageSize, (page - 1) * pageSize);

  const result = await readDb.query(sql, params);
  return { orders: result.rows, page, pageSize };
};

// Express API for queries
router.get("/orders", async (req, res) => {
  const result = await queryBus.execute({
    type: "GetOrderList",
    params: {
      customerId: req.query.customerId,
      status: req.query.status,
      page: parseInt(req.query.page as string) || 1,
    },
  });
  res.json(result);
});
```

### Read Model Projections

```typescript
// Event processor that updates read models
class ReadModelProcessor {
  private projections: Map<string, Projection[]> = new Map();
  private lastProcessedId: number = 0;

  register(eventType: string, projection: Projection) {
    const existing = this.projections.get(eventType) ?? [];
    existing.push(projection);
    this.projections.set(eventType, existing);
  }

  // Poll-based processing
  async processNewEvents(): Promise<number> {
    const events = await eventStore.query(
      `SELECT * FROM events WHERE id > $1 ORDER BY id ASC LIMIT 100`,
      [this.lastProcessedId]
    );

    for (const event of events.rows) {
      const projections = this.projections.get(event.event_type) ?? [];
      for (const projection of projections) {
        try {
          await projection.handle(event);
        } catch (error) {
          console.error(`Projection error for event ${event.id}:`, error);
          // Log to dead letter, don't stop processing
          await this.deadLetter(event, error);
        }
      }
      this.lastProcessedId = event.id;
      await this.saveCheckpoint(this.lastProcessedId);
    }

    return events.rows.length;
  }

  // Continuous processing loop
  async start() {
    this.lastProcessedId = await this.loadCheckpoint();
    while (true) {
      const count = await this.processNewEvents();
      if (count === 0) {
        await new Promise((r) => setTimeout(r, 500)); // Poll interval
      }
    }
  }

  private async saveCheckpoint(id: number) {
    await eventStore.query(
      `INSERT INTO projection_checkpoints (projection_name, last_event_id)
       VALUES ('read_model', $1)
       ON CONFLICT (projection_name) DO UPDATE SET last_event_id = $1`,
      [id]
    );
  }

  private async loadCheckpoint(): Promise<number> {
    const result = await eventStore.query(
      "SELECT last_event_id FROM projection_checkpoints WHERE projection_name = 'read_model'"
    );
    return result.rows[0]?.last_event_id ?? 0;
  }
}
```

## Saga Pattern (Orchestration)

```typescript
interface SagaStep<TState> {
  name: string;
  execute: (state: TState) => Promise<TState>;
  compensate: (state: TState) => Promise<TState>;
}

class SagaOrchestrator<TState extends { status: string }> {
  private steps: SagaStep<TState>[] = [];

  addStep(step: SagaStep<TState>): this {
    this.steps.push(step);
    return this;
  }

  async execute(initialState: TState): Promise<TState> {
    let state = { ...initialState };
    const completedSteps: SagaStep<TState>[] = [];

    for (const step of this.steps) {
      try {
        state = await step.execute(state);
        completedSteps.push(step);
      } catch (error) {
        console.error(`Saga step "${step.name}" failed:`, error);

        // Compensate in reverse order
        for (const completed of completedSteps.reverse()) {
          try {
            state = await completed.compensate(state);
          } catch (compensateError) {
            console.error(`Compensation for "${completed.name}" failed:`, compensateError);
            // Log for manual intervention
            await this.logCompensationFailure(completed.name, compensateError);
          }
        }

        state.status = "failed";
        return state;
      }
    }

    state.status = "completed";
    return state;
  }

  private async logCompensationFailure(stepName: string, error: any) {
    // Store for manual resolution
    await db.query(
      `INSERT INTO saga_failures (step_name, error, created_at) VALUES ($1, $2, NOW())`,
      [stepName, JSON.stringify(error)]
    );
  }
}

// Usage: Order fulfillment saga
interface OrderSagaState {
  orderId: string;
  customerId: string;
  items: OrderItem[];
  totalUsd: number;
  status: string;
  paymentId?: string;
  reservationId?: string;
  shipmentId?: string;
}

const orderFulfillmentSaga = new SagaOrchestrator<OrderSagaState>()
  .addStep({
    name: "reserveInventory",
    execute: async (state) => {
      const reservationId = await inventoryService.reserve(state.items);
      return { ...state, reservationId };
    },
    compensate: async (state) => {
      if (state.reservationId) {
        await inventoryService.release(state.reservationId);
      }
      return { ...state, reservationId: undefined };
    },
  })
  .addStep({
    name: "processPayment",
    execute: async (state) => {
      const paymentId = await paymentService.charge(
        state.customerId,
        state.totalUsd
      );
      return { ...state, paymentId };
    },
    compensate: async (state) => {
      if (state.paymentId) {
        await paymentService.refund(state.paymentId);
      }
      return { ...state, paymentId: undefined };
    },
  })
  .addStep({
    name: "createShipment",
    execute: async (state) => {
      const shipmentId = await shippingService.create(
        state.orderId,
        state.items
      );
      return { ...state, shipmentId };
    },
    compensate: async (state) => {
      if (state.shipmentId) {
        await shippingService.cancel(state.shipmentId);
      }
      return { ...state, shipmentId: undefined };
    },
  });

// Execute the saga
const result = await orderFulfillmentSaga.execute({
  orderId: "order-123",
  customerId: "cust-456",
  items: [...],
  totalUsd: 99.99,
  status: "pending",
});
```

## Idempotency

```typescript
// Idempotency middleware for event handlers
class IdempotentHandler {
  constructor(private pool: Pool) {}

  async handle(
    eventId: string,
    handlerName: string,
    handler: () => Promise<void>
  ): Promise<void> {
    // Check if already processed
    const existing = await this.pool.query(
      `SELECT 1 FROM processed_events WHERE event_id = $1 AND handler = $2`,
      [eventId, handlerName]
    );

    if (existing.rows.length > 0) {
      return; // Already processed, skip
    }

    const client = await this.pool.connect();
    try {
      await client.query("BEGIN");

      // Mark as processed (with advisory lock to prevent races)
      await client.query(
        `INSERT INTO processed_events (event_id, handler, processed_at)
         VALUES ($1, $2, NOW())
         ON CONFLICT DO NOTHING`,
        [eventId, handlerName]
      );

      // Check if we actually inserted (another worker might have beaten us)
      const inserted = await client.query(
        `SELECT 1 FROM processed_events WHERE event_id = $1 AND handler = $2 AND processed_at > NOW() - INTERVAL '1 second'`,
        [eventId, handlerName]
      );

      if (inserted.rows.length > 0) {
        await handler();
      }

      await client.query("COMMIT");
    } catch (error) {
      await client.query("ROLLBACK");
      throw error;
    } finally {
      client.release();
    }
  }
}
```

## Eventual Consistency Handling

```typescript
// Client-side: poll until read model catches up
async function waitForConsistency(
  orderId: string,
  expectedVersion: number,
  maxWaitMs: number = 5000
): Promise<Order | null> {
  const start = Date.now();

  while (Date.now() - start < maxWaitMs) {
    const order = await queryBus.execute({
      type: "GetOrder",
      params: { id: orderId },
    });

    if (order && order.version >= expectedVersion) {
      return order;
    }

    await new Promise((r) => setTimeout(r, 200));
  }

  return null; // Timed out
}

// API pattern: return correlationId + version for polling
router.post("/orders", async (req, res) => {
  const correlationId = crypto.randomUUID();

  await commandBus.dispatch({
    type: "CreateOrder",
    payload: req.body,
    metadata: { correlationId, userId: req.user!.id, timestamp: new Date().toISOString() },
  });

  // Return 202 with polling info
  res.status(202).json({
    correlationId,
    statusUrl: `/orders/status/${correlationId}`,
  });
});

// Polling endpoint
router.get("/orders/status/:correlationId", async (req, res) => {
  const result = await queryBus.execute({
    type: "GetByCorrelationId",
    params: { correlationId: req.params.correlationId },
  });

  if (result) {
    res.json({ status: "completed", data: result });
  } else {
    res.status(202).json({ status: "processing" });
  }
});
```

## Gotchas

1. **CQRS doesn't require event sourcing** — You can have separate read/write models with a regular database. Event sourcing adds complexity; use it only when you need the audit trail.

2. **Eventual consistency is unavoidable** — The read model always lags behind the write model. Design your UI for this: show optimistic updates, use polling or WebSockets for real-time updates.

3. **Saga compensation isn't rollback** — Compensation creates new events that undo previous effects. It's not a database rollback. Some effects (emails sent, SMS delivered) can't be compensated — design for this.

4. **Command handlers should be thin** — Validate input, load aggregate, call methods, save. Don't put domain logic in the handler — put it in the aggregate.

5. **One command = one aggregate** — A single command should only modify one aggregate. If you need to modify multiple aggregates, use a saga.

6. **Read models are cheap** — Create as many as you need. One per UI view is fine. They're just database tables built from events — rebuild them whenever the schema changes.
