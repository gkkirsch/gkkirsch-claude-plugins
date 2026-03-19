# Distributed Systems Patterns Reference

Comprehensive reference for distributed systems patterns used in microservices architectures. Covers
reliability patterns (circuit breaker, retry, bulkhead), data consistency patterns (saga, outbox,
two-phase commit), and communication patterns (API composition, CQRS, event sourcing).

## Reliability Patterns

### Circuit Breaker Pattern

Prevents cascading failures by detecting when a service is failing and short-circuiting calls to it.

**States:**

```
CLOSED (normal operation)
  ↓ failure threshold reached
OPEN (all calls fail fast)
  ↓ timeout expires
HALF-OPEN (probe with limited calls)
  ↓ probe succeeds → CLOSED
  ↓ probe fails → OPEN
```

**Implementation:**

```typescript
enum CircuitState {
  CLOSED = 'CLOSED',
  OPEN = 'OPEN',
  HALF_OPEN = 'HALF_OPEN',
}

interface CircuitBreakerConfig {
  failureThreshold: number;      // Failures before opening (default: 5)
  successThreshold: number;      // Successes in half-open before closing (default: 3)
  timeout: number;               // Ms before trying half-open (default: 30000)
  monitorInterval: number;       // Ms to check failure rate (default: 10000)
  halfOpenMaxCalls: number;      // Max calls in half-open state (default: 3)
}

class CircuitBreaker {
  private state = CircuitState.CLOSED;
  private failureCount = 0;
  private successCount = 0;
  private lastFailureTime = 0;
  private halfOpenCallCount = 0;

  constructor(private config: CircuitBreakerConfig) {}

  async execute<T>(fn: () => Promise<T>): Promise<T> {
    if (this.state === CircuitState.OPEN) {
      if (Date.now() - this.lastFailureTime >= this.config.timeout) {
        this.state = CircuitState.HALF_OPEN;
        this.halfOpenCallCount = 0;
        this.successCount = 0;
      } else {
        throw new CircuitOpenError('Circuit breaker is OPEN');
      }
    }

    if (this.state === CircuitState.HALF_OPEN) {
      if (this.halfOpenCallCount >= this.config.halfOpenMaxCalls) {
        throw new CircuitOpenError('Circuit breaker HALF_OPEN: max probes reached');
      }
      this.halfOpenCallCount++;
    }

    try {
      const result = await fn();
      this.onSuccess();
      return result;
    } catch (error) {
      this.onFailure();
      throw error;
    }
  }

  private onSuccess(): void {
    if (this.state === CircuitState.HALF_OPEN) {
      this.successCount++;
      if (this.successCount >= this.config.successThreshold) {
        this.state = CircuitState.CLOSED;
        this.failureCount = 0;
      }
    } else {
      this.failureCount = 0;
    }
  }

  private onFailure(): void {
    this.failureCount++;
    this.lastFailureTime = Date.now();

    if (this.state === CircuitState.HALF_OPEN) {
      this.state = CircuitState.OPEN;
    } else if (this.failureCount >= this.config.failureThreshold) {
      this.state = CircuitState.OPEN;
    }
  }

  getState(): CircuitState {
    return this.state;
  }
}
```

**When to use:**
- Calling external services (APIs, databases, message brokers)
- Any operation that can fail due to external dependencies
- Protecting against slow responses that consume resources

**Configuration guidelines:**

| Service Type | Failure Threshold | Timeout | Half-Open Max |
|-------------|-------------------|---------|---------------|
| Critical (payments) | 3 | 60s | 1 |
| Standard (catalog) | 5 | 30s | 3 |
| Non-critical (analytics) | 10 | 15s | 5 |
| External API | 3 | 120s | 1 |

### Retry Pattern

Automatically retries failed operations with configurable backoff strategies.

**Backoff strategies:**

| Strategy | Formula | Use Case |
|----------|---------|----------|
| Fixed | `delay` | Simple, predictable retries |
| Linear | `delay * attempt` | Gradual backoff |
| Exponential | `delay * 2^attempt` | Standard for most services |
| Exponential + jitter | `delay * 2^attempt * random(0.5, 1.5)` | Prevents thundering herd |
| Decorrelated jitter | `min(cap, random(base, prev * 3))` | AWS recommended |

**Implementation:**

```typescript
interface RetryConfig {
  maxRetries: number;
  baseDelay: number;        // ms
  maxDelay: number;          // ms
  backoff: 'fixed' | 'linear' | 'exponential' | 'jitter';
  retryableErrors?: string[];
  nonRetryableErrors?: string[];
}

async function withRetry<T>(
  fn: () => Promise<T>,
  config: RetryConfig
): Promise<T> {
  let lastError: Error;

  for (let attempt = 0; attempt <= config.maxRetries; attempt++) {
    try {
      return await fn();
    } catch (error) {
      lastError = error as Error;

      if (isNonRetryable(error, config)) {
        throw error;
      }

      if (attempt === config.maxRetries) {
        throw error;
      }

      const delay = calculateDelay(attempt, config);
      await new Promise(resolve => setTimeout(resolve, delay));
    }
  }

  throw lastError!;
}

function calculateDelay(attempt: number, config: RetryConfig): number {
  let delay: number;

  switch (config.backoff) {
    case 'fixed':
      delay = config.baseDelay;
      break;
    case 'linear':
      delay = config.baseDelay * (attempt + 1);
      break;
    case 'exponential':
      delay = config.baseDelay * Math.pow(2, attempt);
      break;
    case 'jitter':
      delay = config.baseDelay * Math.pow(2, attempt) * (0.5 + Math.random());
      break;
  }

  return Math.min(delay, config.maxDelay);
}

function isNonRetryable(error: any, config: RetryConfig): boolean {
  // HTTP 4xx (except 429) are not retryable
  if (error.status >= 400 && error.status < 500 && error.status !== 429) {
    return true;
  }

  if (config.nonRetryableErrors?.some(e => error.message.includes(e))) {
    return true;
  }

  return false;
}
```

**Retry decision matrix:**

| Error Type | Retry? | Strategy |
|-----------|--------|----------|
| Connection timeout | Yes | Exponential + jitter |
| Connection refused | Yes | Exponential + jitter |
| 429 Too Many Requests | Yes | Use Retry-After header |
| 500 Internal Server Error | Yes | Exponential + jitter |
| 502 Bad Gateway | Yes | Exponential + jitter |
| 503 Service Unavailable | Yes | Exponential + jitter |
| 504 Gateway Timeout | Yes | Exponential + jitter |
| 400 Bad Request | No | Fix the request |
| 401 Unauthorized | No | Refresh token, then retry once |
| 403 Forbidden | No | Check permissions |
| 404 Not Found | No | Resource doesn't exist |
| 409 Conflict | Maybe | Re-read, merge, retry |
| 422 Unprocessable Entity | No | Fix the request |
| Network error (DNS) | Yes | Exponential + jitter |
| Parse error | No | Fix the response handler |

### Bulkhead Pattern

Isolates failures by partitioning resources so that a failure in one partition doesn't affect others.

**Types of bulkheads:**

```
Thread Pool Bulkhead:
┌────────────────────────────────────────────┐
│              Application                    │
│                                            │
│  ┌──────────────┐  ┌──────────────┐       │
│  │ Order Pool   │  │ Payment Pool │       │
│  │ (10 threads) │  │ (5 threads)  │       │
│  │              │  │              │       │
│  │ ████████░░   │  │ ███░░        │       │
│  └──────────────┘  └──────────────┘       │
│                                            │
│  ┌──────────────┐  ┌──────────────┐       │
│  │ Catalog Pool │  │ Search Pool  │       │
│  │ (8 threads)  │  │ (4 threads)  │       │
│  │              │  │              │       │
│  │ ██████░░     │  │ ██░░         │       │
│  └──────────────┘  └──────────────┘       │
│                                            │
│  If Payment Pool exhausted, Order Pool     │
│  and Catalog Pool continue working.        │
└────────────────────────────────────────────┘
```

**Semaphore bulkhead implementation:**

```typescript
class Bulkhead {
  private activeCount = 0;
  private queue: Array<{ resolve: () => void; reject: (err: Error) => void }> = [];

  constructor(
    private maxConcurrent: number,
    private maxQueue: number = 100,
    private queueTimeout: number = 30000
  ) {}

  async execute<T>(fn: () => Promise<T>): Promise<T> {
    if (this.activeCount < this.maxConcurrent) {
      return this.run(fn);
    }

    if (this.queue.length >= this.maxQueue) {
      throw new BulkheadFullError(
        `Bulkhead full: ${this.activeCount} active, ${this.queue.length} queued`
      );
    }

    // Wait for a slot
    await new Promise<void>((resolve, reject) => {
      const timer = setTimeout(() => {
        const idx = this.queue.findIndex(item => item.resolve === resolve);
        if (idx >= 0) this.queue.splice(idx, 1);
        reject(new BulkheadTimeoutError('Queue timeout'));
      }, this.queueTimeout);

      this.queue.push({
        resolve: () => {
          clearTimeout(timer);
          resolve();
        },
        reject,
      });
    });

    return this.run(fn);
  }

  private async run<T>(fn: () => Promise<T>): Promise<T> {
    this.activeCount++;
    try {
      return await fn();
    } finally {
      this.activeCount--;
      this.releaseNext();
    }
  }

  private releaseNext(): void {
    if (this.queue.length > 0 && this.activeCount < this.maxConcurrent) {
      const next = this.queue.shift();
      next?.resolve();
    }
  }
}
```

### Timeout Pattern

Set explicit timeouts on all remote calls to prevent resource exhaustion.

**Timeout guidelines:**

| Operation | Timeout | Rationale |
|-----------|---------|-----------|
| Health check | 3s | Must be fast |
| Database query (simple) | 5s | Most queries <1s |
| Database query (complex) | 30s | Reports, aggregations |
| HTTP API call | 10s | Standard service call |
| File upload | 60s | Large payload transfer |
| Batch operation | 300s | Bulk processing |
| Message publish | 5s | Broker should respond fast |
| Message consume + process | 30s | Processing takes time |
| gRPC unary call | 10s | Similar to HTTP |
| gRPC streaming | 300s | Long-lived connections |

**Cascading timeout pattern:**

```
Client timeout: 30s
  └── Gateway timeout: 25s (< client)
        └── Service A timeout: 20s (< gateway)
              └── Service B call: 10s (< Service A)
              └── Database query: 5s (< Service A)
```

Always set inner timeouts shorter than outer timeouts to prevent resource waste.

### Rate Limiter Pattern

Controls the rate of requests to prevent overload.

**Token bucket implementation:**

```typescript
class TokenBucket {
  private tokens: number;
  private lastRefillTime: number;

  constructor(
    private capacity: number,
    private refillRate: number,  // tokens per second
  ) {
    this.tokens = capacity;
    this.lastRefillTime = Date.now();
  }

  tryConsume(tokens: number = 1): boolean {
    this.refill();

    if (this.tokens >= tokens) {
      this.tokens -= tokens;
      return true;
    }

    return false;
  }

  private refill(): void {
    const now = Date.now();
    const elapsed = (now - this.lastRefillTime) / 1000;
    const newTokens = elapsed * this.refillRate;
    this.tokens = Math.min(this.capacity, this.tokens + newTokens);
    this.lastRefillTime = now;
  }
}
```

## Data Consistency Patterns

### Saga Pattern

Manages distributed transactions across multiple services using a sequence of local transactions.

**Choreography saga (event-driven):**

```
Order Service          Payment Service        Inventory Service
     │                      │                      │
     │ OrderSubmitted       │                      │
     │─────────────────────→│                      │
     │                      │ PaymentRequested     │
     │                      │─────────────────────→│
     │                      │                      │ InventoryReserved
     │                      │←─────────────────────│
     │ PaymentCompleted     │                      │
     │←─────────────────────│                      │
     │                      │                      │
     │ (success path)       │                      │
     │                      │                      │
     │ --- COMPENSATION --- │                      │
     │                      │                      │
     │ PaymentFailed        │                      │
     │←─────────────────────│                      │
     │ OrderCancelled       │                      │
     │─────────────────────→│                      │
     │                      │ ReleaseInventory     │
     │                      │─────────────────────→│
```

**Orchestration saga (coordinator):**

```
       Saga Orchestrator
             │
    Step 1:  │──→ Reserve Inventory
             │←── InventoryReserved
    Step 2:  │──→ Process Payment
             │←── PaymentCompleted
    Step 3:  │──→ Create Shipment
             │←── ShipmentCreated
             │
    DONE     │──→ Confirm Order
             │
    --- ON FAILURE ---
             │
    Step 2 fails:
             │──→ Release Inventory (compensate step 1)
             │──→ Cancel Order
```

**Saga state machine:**

```
          ┌─────────┐
          │ STARTED │
          └────┬────┘
               │
     ┌─────────┴─────────┐
     ↓                    ↓
┌─────────┐         ┌──────────┐
│RESERVING│         │  FAILED  │
│INVENTORY│         │(no comp) │
└────┬────┘         └──────────┘
     │
     ├── success ──→ ┌──────────┐
     │               │PROCESSING│
     │               │ PAYMENT  │
     │               └────┬─────┘
     │                    │
     │               ├── success ──→ ┌─────────┐
     │               │               │CREATING │
     │               │               │SHIPMENT │
     │               │               └────┬────┘
     │               │                    │
     │               │               ├── success ──→ ┌───────────┐
     │               │               │               │ COMPLETED │
     │               │               │               └───────────┘
     │               │               │
     │               │               └── fail ──→ ┌──────────────┐
     │               │                            │ COMPENSATING │
     │               │                            │ (refund)     │
     │               │                            └──────┬───────┘
     │               │                                   │
     │               └── fail ──→ ┌──────────────┐       │
     │                            │ COMPENSATING │       │
     │                            │ (release inv)│       │
     │                            └──────────────┘       │
     │                                                   │
     └── fail ──→ ┌──────────┐                          │
                  │  FAILED  │←─────────────────────────┘
                  └──────────┘
```

### Outbox Pattern

Ensures reliable event publishing by storing events in a database table alongside the business data
in the same transaction, then publishing asynchronously.

**How it works:**

```
1. Business operation + event stored in same DB transaction
   BEGIN TRANSACTION
     INSERT INTO orders (...) VALUES (...)
     INSERT INTO outbox (event_type, payload) VALUES ('OrderCreated', '...')
   COMMIT

2. Background process polls outbox table
   SELECT * FROM outbox WHERE published = false ORDER BY created_at LIMIT 100

3. Publish each event to message broker
   kafka.publish(event.topic, event.payload)

4. Mark as published
   UPDATE outbox SET published = true WHERE id = ?
```

**Outbox table schema:**

```sql
CREATE TABLE outbox (
  id              BIGSERIAL PRIMARY KEY,
  aggregate_type  VARCHAR(255) NOT NULL,
  aggregate_id    VARCHAR(255) NOT NULL,
  event_type      VARCHAR(255) NOT NULL,
  payload         JSONB NOT NULL,
  metadata        JSONB DEFAULT '{}',
  created_at      TIMESTAMP NOT NULL DEFAULT NOW(),
  published       BOOLEAN NOT NULL DEFAULT FALSE,
  published_at    TIMESTAMP,
  retry_count     INT NOT NULL DEFAULT 0,
  last_error      TEXT
);

CREATE INDEX idx_outbox_unpublished ON outbox (published, created_at)
  WHERE published = FALSE;
CREATE INDEX idx_outbox_aggregate ON outbox (aggregate_type, aggregate_id);
```

**Outbox publisher implementation:**

```typescript
class OutboxPublisher {
  private running = false;

  constructor(
    private db: PrismaClient,
    private eventPublisher: EventPublisher,
    private batchSize: number = 100,
    private pollInterval: number = 1000
  ) {}

  async start(): Promise<void> {
    this.running = true;

    while (this.running) {
      try {
        const published = await this.publishBatch();
        if (published === 0) {
          await this.delay(this.pollInterval);
        }
      } catch (error) {
        logger.error({ err: error }, 'Outbox publisher error');
        await this.delay(this.pollInterval * 5);
      }
    }
  }

  private async publishBatch(): Promise<number> {
    const events = await this.db.outbox.findMany({
      where: { published: false },
      orderBy: { createdAt: 'asc' },
      take: this.batchSize,
    });

    for (const event of events) {
      try {
        await this.eventPublisher.publish(
          `${event.aggregateType}.events`,
          event.eventType,
          event.aggregateId,
          event.aggregateType,
          JSON.parse(event.payload as string)
        );

        await this.db.outbox.update({
          where: { id: event.id },
          data: { published: true, publishedAt: new Date() },
        });
      } catch (error) {
        await this.db.outbox.update({
          where: { id: event.id },
          data: {
            retryCount: { increment: 1 },
            lastError: (error as Error).message,
          },
        });
      }
    }

    return events.length;
  }

  stop(): void {
    this.running = false;
  }

  private delay(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms));
  }
}
```

### Two-Phase Commit (2PC)

Coordinates distributed transactions across multiple resources.

**When to use (rarely in microservices):**
- Distributed database transactions within the same service
- Cross-database operations (e.g., PostgreSQL + MongoDB)
- Legacy system integration

**When NOT to use:**
- Cross-service transactions (use sagas instead)
- High-throughput systems (2PC is blocking)
- Systems requiring high availability (coordinator is SPOF)

### Event Sourcing Pattern

Store state as a sequence of events rather than current state.

**Benefits:**
- Complete audit trail
- Time travel (replay to any point)
- Event replay for new consumers
- Natural fit for CQRS

**Drawbacks:**
- Increased complexity
- Event schema evolution is hard
- Querying current state requires projection
- Storage grows unbounded without snapshots

**Snapshot strategy:**

```typescript
// Take a snapshot every N events to speed up aggregate reconstruction
const SNAPSHOT_INTERVAL = 100;

async function loadAggregate(aggregateId: string): Promise<Order> {
  // Try to load from snapshot first
  const snapshot = await snapshotStore.getLatest(aggregateId);
  const fromVersion = snapshot ? snapshot.version : 0;
  const order = snapshot ? Order.fromSnapshot(snapshot.data) : Order.empty();

  // Apply events since snapshot
  const events = await eventStore.getEvents(aggregateId, fromVersion);
  for (const event of events) {
    order.apply(event);
  }

  // Take new snapshot if needed
  if (events.length >= SNAPSHOT_INTERVAL) {
    await snapshotStore.save(aggregateId, order.version, order.toSnapshot());
  }

  return order;
}
```

## Communication Patterns

### API Composition Pattern

Aggregate data from multiple services in a single API call.

```
Client → API Composer → Service A (get user)
                      → Service B (get orders)
                      → Service C (get recommendations)
         ← Combined response
```

**Implementation:**

```typescript
async function getUserDashboard(userId: string) {
  const [user, orders, recommendations] = await Promise.allSettled([
    userService.getUser(userId),
    orderService.getRecentOrders(userId),
    recommendationService.getForUser(userId),
  ]);

  return {
    user: user.status === 'fulfilled' ? user.value : null,
    orders: orders.status === 'fulfilled' ? orders.value : [],
    recommendations: recommendations.status === 'fulfilled' ? recommendations.value : [],
    errors: [user, orders, recommendations]
      .filter(r => r.status === 'rejected')
      .map(r => (r as PromiseRejectedResult).reason.message),
  };
}
```

### Backend for Frontend (BFF) Pattern

Separate API gateways tailored to each frontend client type.

| Client | BFF | Optimizations |
|--------|-----|--------------|
| Web SPA | Web BFF | Full data, images, pagination |
| Mobile | Mobile BFF | Minimal payload, offline support |
| IoT | IoT BFF | Binary protocols, batch updates |
| Partner API | Partner BFF | Rate limiting, API keys, webhooks |

### Strangler Fig Pattern

Gradually replace a monolith by routing traffic to new microservices.

```
Phase 1: All traffic → Monolith
Phase 2: /products → New Service, /* → Monolith
Phase 3: /products, /orders → New Services, /* → Monolith
Phase 4: All traffic → New Services (monolith retired)
```

### Anti-Corruption Layer (ACL) Pattern

Translate between different domain models at service boundaries.

**Use when:**
- Integrating with legacy systems
- Consuming third-party APIs with poor models
- Preventing external model leakage into your domain

### Ambassador Pattern

Deploy a helper service alongside your main service to handle cross-cutting concerns.

```
┌─────────────────────┐
│     Application     │
│  ┌───────────────┐  │
│  │  Main Service │  │
│  └───────┬───────┘  │
│          │          │
│  ┌───────┴───────┐  │
│  │  Ambassador   │  │ Handles: TLS, retries, circuit breaking,
│  │  (sidecar)    │  │ monitoring, logging, auth
│  └───────────────┘  │
└─────────────────────┘
```

## Pattern Selection Guide

| Problem | Primary Pattern | Secondary Pattern |
|---------|----------------|-------------------|
| Service keeps failing | Circuit Breaker | Retry + Bulkhead |
| Need distributed transaction | Saga | Outbox + Events |
| Query spans multiple services | API Composition | CQRS |
| Need complete audit trail | Event Sourcing | Outbox |
| Different clients, different needs | BFF | API Gateway |
| Replacing legacy system | Strangler Fig | ACL |
| Preventing cascade failure | Bulkhead | Circuit Breaker + Timeout |
| External API integration | ACL | Circuit Breaker + Retry |
| High-throughput events | Event Streaming | CQRS |
| Complex workflow | Orchestration Saga | State Machine |

## Anti-Patterns to Avoid

| Anti-Pattern | Problem | Solution |
|-------------|---------|----------|
| Distributed monolith | All services deploy together | True service independence |
| Shared database | Multiple services access same DB | Database per service |
| Synchronous saga | Chained blocking calls | Async events or orchestration |
| Chatty services | Too many inter-service calls | Merge services or use events |
| God service | One service knows everything | Proper bounded contexts |
| Nano services | Too many tiny services | Right-size services |
| Manual retry | Copy-paste retry logic | Retry library/pattern |
| No timeout | Calls hang forever | Explicit timeouts everywhere |
| Ignoring idempotency | Duplicate processing | Idempotency keys |
| No dead letter queue | Lost messages | DLQ for every consumer |
