---
name: event-bus-patterns
description: >
  Event bus, pub/sub, and domain event patterns for event-driven architectures.
  Use when implementing in-process event systems, domain events, saga orchestration,
  or decoupled communication between services or modules.
  Triggers: "event bus", "pub/sub", "domain events", "event emitter",
  "saga pattern", "event dispatcher", "message broker in-process",
  "decoupled events", "EventEmitter pattern".
  NOT for: CQRS (see cqrs-patterns), event sourcing (see event-sourcing), external message queues (see message-queues-suite).
version: 1.0.0
allowed-tools: Read, Grep, Glob, Bash
---

# Event Bus Patterns

## Type-Safe Event Bus

```typescript
// lib/event-bus.ts — Strongly-typed in-process event system
type EventMap = {
  'user.created': { userId: string; email: string; plan: string };
  'user.upgraded': { userId: string; fromPlan: string; toPlan: string };
  'order.placed': { orderId: string; userId: string; total: number; items: string[] };
  'order.fulfilled': { orderId: string; shippedAt: Date; trackingNumber: string };
  'payment.received': { paymentId: string; orderId: string; amount: number };
  'payment.failed': { paymentId: string; orderId: string; reason: string };
};

type EventHandler<T> = (payload: T) => void | Promise<void>;

class TypedEventBus {
  private handlers = new Map<string, Set<EventHandler<unknown>>>();
  private middlewares: Array<(event: string, payload: unknown) => unknown> = [];

  on<K extends keyof EventMap>(event: K, handler: EventHandler<EventMap[K]>): () => void {
    if (!this.handlers.has(event)) {
      this.handlers.set(event, new Set());
    }
    this.handlers.get(event)!.add(handler as EventHandler<unknown>);

    // Return unsubscribe function
    return () => {
      this.handlers.get(event)?.delete(handler as EventHandler<unknown>);
    };
  }

  once<K extends keyof EventMap>(event: K, handler: EventHandler<EventMap[K]>): void {
    const unsubscribe = this.on(event, (payload) => {
      unsubscribe();
      return handler(payload);
    });
  }

  async emit<K extends keyof EventMap>(event: K, payload: EventMap[K]): Promise<void> {
    // Run middlewares
    let processedPayload = payload as unknown;
    for (const mw of this.middlewares) {
      processedPayload = mw(event, processedPayload);
    }

    const handlers = this.handlers.get(event);
    if (!handlers?.size) return;

    // Execute all handlers concurrently
    const promises = Array.from(handlers).map(async (handler) => {
      try {
        await handler(processedPayload);
      } catch (error) {
        console.error(`Error in handler for ${event}:`, error);
        // Don't throw — one failing handler shouldn't break others
      }
    });

    await Promise.allSettled(promises);
  }

  use(middleware: (event: string, payload: unknown) => unknown): void {
    this.middlewares.push(middleware);
  }

  removeAllListeners(event?: keyof EventMap): void {
    if (event) {
      this.handlers.delete(event);
    } else {
      this.handlers.clear();
    }
  }
}

// Singleton instance
export const eventBus = new TypedEventBus();

// Logging middleware
eventBus.use((event, payload) => {
  console.log(`[Event] ${event}`, JSON.stringify(payload).slice(0, 200));
  return payload;
});
```

## Domain Events Pattern

```typescript
// domain/events.ts — Domain events with metadata

interface DomainEvent<T = unknown> {
  eventId: string;
  eventType: string;
  aggregateId: string;
  aggregateType: string;
  payload: T;
  metadata: {
    timestamp: Date;
    correlationId: string;
    causationId: string | null;
    userId: string | null;
    version: number;
  };
}

// Base aggregate that collects domain events
abstract class AggregateRoot {
  private domainEvents: DomainEvent[] = [];
  abstract get id(): string;

  protected addEvent<T>(type: string, payload: T, correlationId: string): void {
    this.domainEvents.push({
      eventId: crypto.randomUUID(),
      eventType: type,
      aggregateId: this.id,
      aggregateType: this.constructor.name,
      payload,
      metadata: {
        timestamp: new Date(),
        correlationId,
        causationId: null,
        userId: null,
        version: this.domainEvents.length + 1,
      },
    });
  }

  pullEvents(): DomainEvent[] {
    const events = [...this.domainEvents];
    this.domainEvents = [];
    return events;
  }
}

// Example: Order aggregate
class Order extends AggregateRoot {
  constructor(
    public readonly id: string,
    private status: 'pending' | 'confirmed' | 'shipped' | 'delivered',
    private items: Array<{ productId: string; quantity: number; price: number }>,
  ) {
    super();
  }

  confirm(correlationId: string): void {
    if (this.status !== 'pending') throw new Error('Only pending orders can be confirmed');
    this.status = 'confirmed';
    this.addEvent('OrderConfirmed', {
      orderId: this.id,
      total: this.items.reduce((sum, i) => sum + i.price * i.quantity, 0),
      itemCount: this.items.length,
    }, correlationId);
  }

  ship(trackingNumber: string, correlationId: string): void {
    if (this.status !== 'confirmed') throw new Error('Only confirmed orders can be shipped');
    this.status = 'shipped';
    this.addEvent('OrderShipped', {
      orderId: this.id,
      trackingNumber,
      shippedAt: new Date().toISOString(),
    }, correlationId);
  }
}

// Usage: after saving the aggregate, dispatch its events
// const order = new Order('ord-1', 'pending', items);
// order.confirm(correlationId);
// await orderRepo.save(order);
// for (const event of order.pullEvents()) {
//   await eventBus.emit(event.eventType, event);
// }
```

## Saga Pattern (Process Manager)

```typescript
// sagas/order-saga.ts — Coordinate multi-step business process

interface SagaStep<T = unknown> {
  name: string;
  execute: (context: T) => Promise<void>;
  compensate: (context: T) => Promise<void>; // Undo on failure
}

class SagaOrchestrator<T> {
  private steps: SagaStep<T>[] = [];
  private completedSteps: SagaStep<T>[] = [];

  addStep(step: SagaStep<T>): this {
    this.steps.push(step);
    return this;
  }

  async execute(context: T): Promise<void> {
    for (const step of this.steps) {
      try {
        console.log(`[Saga] Executing: ${step.name}`);
        await step.execute(context);
        this.completedSteps.push(step);
      } catch (error) {
        console.error(`[Saga] Failed at: ${step.name}`, error);
        await this.compensate(context);
        throw new Error(`Saga failed at ${step.name}: ${(error as Error).message}`);
      }
    }
  }

  private async compensate(context: T): Promise<void> {
    // Compensate in reverse order
    for (const step of this.completedSteps.reverse()) {
      try {
        console.log(`[Saga] Compensating: ${step.name}`);
        await step.compensate(context);
      } catch (error) {
        console.error(`[Saga] Compensation failed for ${step.name}:`, error);
        // Log and continue — compensation must be best-effort
      }
    }
  }
}

// Example: order placement saga
interface OrderContext {
  orderId: string;
  userId: string;
  items: Array<{ productId: string; quantity: number }>;
  paymentId?: string;
  reservationId?: string;
}

const orderSaga = new SagaOrchestrator<OrderContext>()
  .addStep({
    name: 'Reserve Inventory',
    execute: async (ctx) => {
      ctx.reservationId = await inventoryService.reserve(ctx.items);
    },
    compensate: async (ctx) => {
      if (ctx.reservationId) await inventoryService.release(ctx.reservationId);
    },
  })
  .addStep({
    name: 'Process Payment',
    execute: async (ctx) => {
      ctx.paymentId = await paymentService.charge(ctx.userId, ctx.items);
    },
    compensate: async (ctx) => {
      if (ctx.paymentId) await paymentService.refund(ctx.paymentId);
    },
  })
  .addStep({
    name: 'Confirm Order',
    execute: async (ctx) => {
      await orderService.confirm(ctx.orderId);
    },
    compensate: async (ctx) => {
      await orderService.cancel(ctx.orderId);
    },
  })
  .addStep({
    name: 'Send Confirmation Email',
    execute: async (ctx) => {
      await emailService.sendOrderConfirmation(ctx.userId, ctx.orderId);
    },
    compensate: async (_ctx) => {
      // No compensation needed for emails — idempotent side effect
    },
  });

// Usage:
// await orderSaga.execute({ orderId: 'ord-1', userId: 'user-1', items: [...] });
```

## Gotchas

1. **Event handler errors silently swallowed** -- If one event handler throws, should it prevent other handlers from running? Most event buses use `Promise.allSettled()` to isolate failures, but this means errors are logged and ignored. Always have a dead letter queue or error tracking for failed event handlers.

2. **Memory leaks from unsubscribed handlers** -- Handlers registered with `on()` persist until manually removed. In React components, forgetting to unsubscribe in cleanup causes memory leaks and stale closures. Always capture and call the unsubscribe function in `useEffect` cleanup.

3. **Saga compensation is not atomic** -- Compensation steps can fail too. If "Refund Payment" fails during saga compensation, you have a partially compensated saga. Compensation must be idempotent and retryable. Log compensation failures to a dead letter table for manual resolution.

4. **Event ordering is not guaranteed** -- `Promise.allSettled()` runs handlers concurrently. If handler A must complete before handler B, you need explicit ordering (priority queues) or sequential execution. Don't assume handlers execute in registration order.

5. **Domain events vs integration events** -- Domain events are in-process, synchronous, and within a bounded context. Integration events cross service boundaries, are async, and use message brokers. Mixing them causes tight coupling. Use domain events inside a service and integration events between services.

6. **Circular event chains** -- Event A triggers handler that emits event B, which triggers handler that emits event A. This creates an infinite loop. Track event causation chains (correlationId/causationId) and detect cycles, or set a maximum event depth limit per correlation.
