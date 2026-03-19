# Testing Patterns Reference

Comprehensive reference for test patterns, fixtures, factories, mocking strategies, and best practices across languages and frameworks.

## Test Factory Patterns

### Builder Pattern Factory (TypeScript)

```typescript
// tests/factories/order.factory.ts
import { faker } from '@faker-js/faker';
import type { Order, OrderItem, OrderStatus, ShippingAddress } from '@/types';

class OrderBuilder {
  private order: Partial<Order> = {};
  private items: OrderItem[] = [];

  withId(id: string): this {
    this.order.id = id;
    return this;
  }

  withCustomer(customerId: string): this {
    this.order.customerId = customerId;
    return this;
  }

  withStatus(status: OrderStatus): this {
    this.order.status = status;
    return this;
  }

  withItem(item: Partial<OrderItem>): this {
    this.items.push({
      productId: item.productId ?? faker.string.uuid(),
      name: item.name ?? faker.commerce.productName(),
      price: item.price ?? parseFloat(faker.commerce.price()),
      quantity: item.quantity ?? faker.number.int({ min: 1, max: 5 }),
      ...item,
    });
    return this;
  }

  withItems(count: number): this {
    for (let i = 0; i < count; i++) {
      this.withItem({});
    }
    return this;
  }

  withShipping(address: Partial<ShippingAddress>): this {
    this.order.shippingAddress = {
      street: address.street ?? faker.location.streetAddress(),
      city: address.city ?? faker.location.city(),
      state: address.state ?? faker.location.state(),
      zipCode: address.zipCode ?? faker.location.zipCode(),
      country: address.country ?? 'US',
    };
    return this;
  }

  withDiscount(amount: number): this {
    this.order.discount = amount;
    return this;
  }

  asPending(): this {
    return this.withStatus('pending');
  }

  asProcessing(): this {
    return this.withStatus('processing');
  }

  asCompleted(): this {
    return this.withStatus('completed').withCompletedAt(new Date());
  }

  asCancelled(): this {
    return this.withStatus('cancelled');
  }

  private withCompletedAt(date: Date): this {
    this.order.completedAt = date;
    return this;
  }

  build(): Order {
    const items = this.items.length > 0 ? this.items : [
      {
        productId: faker.string.uuid(),
        name: faker.commerce.productName(),
        price: parseFloat(faker.commerce.price()),
        quantity: 1,
      },
    ];

    const subtotal = items.reduce((sum, item) => sum + item.price * item.quantity, 0);
    const discount = this.order.discount ?? 0;

    return {
      id: this.order.id ?? faker.string.uuid(),
      customerId: this.order.customerId ?? faker.string.uuid(),
      status: this.order.status ?? 'pending',
      items,
      subtotal,
      discount,
      tax: Math.round(subtotal * 0.0825 * 100) / 100,
      total: Math.round((subtotal - discount) * 1.0825 * 100) / 100,
      shippingAddress: this.order.shippingAddress ?? {
        street: faker.location.streetAddress(),
        city: faker.location.city(),
        state: faker.location.state(),
        zipCode: faker.location.zipCode(),
        country: 'US',
      },
      createdAt: new Date(),
      updatedAt: new Date(),
      completedAt: this.order.completedAt ?? null,
    };
  }
}

// Usage:
export const orderFactory = {
  build: (overrides?: Partial<Order>) => new OrderBuilder().build(),
  builder: () => new OrderBuilder(),
  pending: () => new OrderBuilder().asPending().build(),
  completed: () => new OrderBuilder().asCompleted().build(),
  withItems: (count: number) => new OrderBuilder().withItems(count).build(),
};

// In tests:
// const order = orderFactory.builder()
//   .withCustomer('cust-1')
//   .withItem({ name: 'Widget', price: 9.99, quantity: 2 })
//   .withItem({ name: 'Gadget', price: 19.99, quantity: 1 })
//   .withDiscount(5)
//   .asPending()
//   .build();
```

### Sequence Factory (Python)

```python
# tests/factories/base.py
class Sequence:
    """Auto-incrementing sequence for unique values."""
    _counters: dict[str, int] = {}

    @classmethod
    def next(cls, name: str = "default") -> int:
        cls._counters[name] = cls._counters.get(name, 0) + 1
        return cls._counters[name]

    @classmethod
    def reset(cls, name: str | None = None):
        if name:
            cls._counters.pop(name, None)
        else:
            cls._counters.clear()


# tests/factories/user_factory.py
from tests.factories.base import Sequence

def build_user(**overrides):
    n = Sequence.next("user")
    defaults = {
        "id": f"user-{n}",
        "email": f"user{n}@example.com",
        "name": f"Test User {n}",
        "role": "user",
        "is_active": True,
    }
    return {**defaults, **overrides}

def build_admin(**overrides):
    return build_user(role="admin", **overrides)

def build_users(count: int, **overrides):
    return [build_user(**overrides) for _ in range(count)]
```

## Fixture Patterns

### Database Fixtures with Transactions

```typescript
// tests/fixtures/transaction-fixture.ts
import { PrismaClient } from '@prisma/client';

/**
 * Wraps each test in a transaction that rolls back automatically.
 * Much faster than truncating tables between tests.
 */
export function useTransactionalTests() {
  const prisma = new PrismaClient();
  let originalExecute: any;

  beforeAll(async () => {
    await prisma.$connect();
  });

  beforeEach(async () => {
    // Start a transaction
    await prisma.$executeRaw`BEGIN`;
    // Save original execute for cleanup
    originalExecute = prisma.$executeRaw;
  });

  afterEach(async () => {
    // Rollback the transaction
    await prisma.$executeRaw`ROLLBACK`;
  });

  afterAll(async () => {
    await prisma.$disconnect();
  });

  return prisma;
}
```

### API Test Fixture

```typescript
// tests/fixtures/api-fixture.ts
import { createServer, type Server } from 'http';
import { app } from '@/app';

export function useTestServer() {
  let server: Server;
  let baseURL: string;

  beforeAll(async () => {
    server = createServer(app);
    await new Promise<void>((resolve) => {
      server.listen(0, () => {
        const addr = server.address();
        if (typeof addr === 'object' && addr) {
          baseURL = `http://localhost:${addr.port}`;
        }
        resolve();
      });
    });
  });

  afterAll(async () => {
    await new Promise<void>((resolve) => {
      server.close(() => resolve());
    });
  });

  return {
    getBaseURL: () => baseURL,
    fetch: (path: string, init?: RequestInit) => fetch(`${baseURL}${path}`, init),
  };
}
```

### Authentication Fixture

```typescript
// tests/fixtures/auth-fixture.ts
import jwt from 'jsonwebtoken';
import { buildUser, buildAdminUser } from '../factories/user.factory';

const JWT_SECRET = process.env.JWT_SECRET || 'test-secret';

export function createAuthToken(user: { id: string; email: string; role: string }): string {
  return jwt.sign(
    { sub: user.id, email: user.email, role: user.role },
    JWT_SECRET,
    { expiresIn: '1h' }
  );
}

export function createTestUserWithToken() {
  const user = buildUser();
  const token = createAuthToken(user);
  return { user, token, authHeader: `Bearer ${token}` };
}

export function createTestAdminWithToken() {
  const admin = buildAdminUser();
  const token = createAuthToken(admin);
  return { user: admin, token, authHeader: `Bearer ${token}` };
}

// Usage in tests:
// const { user, authHeader } = createTestUserWithToken();
// const response = await request(app)
//   .get('/api/me')
//   .set('Authorization', authHeader);
```

## Mocking Patterns

### Dependency Injection for Testability

```typescript
// src/services/notification.service.ts
export interface EmailProvider {
  send(to: string, subject: string, body: string): Promise<void>;
}

export interface SMSProvider {
  send(to: string, message: string): Promise<void>;
}

export class NotificationService {
  constructor(
    private emailProvider: EmailProvider,
    private smsProvider: SMSProvider,
  ) {}

  async notifyUser(userId: string, message: string): Promise<void> {
    const user = await this.getUser(userId);
    if (user.preferences.email) {
      await this.emailProvider.send(user.email, 'Notification', message);
    }
    if (user.preferences.sms) {
      await this.smsProvider.send(user.phone, message);
    }
  }
}

// tests/services/notification.service.test.ts
describe('NotificationService', () => {
  const mockEmail: EmailProvider = { send: vi.fn() };
  const mockSMS: SMSProvider = { send: vi.fn() };
  const service = new NotificationService(mockEmail, mockSMS);

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should send email when user prefers email', async () => {
    vi.spyOn(service as any, 'getUser').mockResolvedValue({
      email: 'test@example.com',
      phone: '+1234567890',
      preferences: { email: true, sms: false },
    });

    await service.notifyUser('user-1', 'Hello');

    expect(mockEmail.send).toHaveBeenCalledWith('test@example.com', 'Notification', 'Hello');
    expect(mockSMS.send).not.toHaveBeenCalled();
  });
});
```

### Manual Mocks (Fakes)

```typescript
// tests/mocks/fake-database.ts
export class FakeDatabase {
  private store = new Map<string, Map<string, any>>();

  collection(name: string) {
    if (!this.store.has(name)) {
      this.store.set(name, new Map());
    }
    const data = this.store.get(name)!;

    return {
      async create(item: any): Promise<any> {
        const id = item.id || crypto.randomUUID();
        const record = { ...item, id, createdAt: new Date(), updatedAt: new Date() };
        data.set(id, record);
        return record;
      },

      async findById(id: string): Promise<any | null> {
        return data.get(id) ?? null;
      },

      async findAll(filter?: Record<string, any>): Promise<any[]> {
        let results = Array.from(data.values());
        if (filter) {
          results = results.filter((item) =>
            Object.entries(filter).every(([key, value]) => item[key] === value)
          );
        }
        return results;
      },

      async update(id: string, updates: any): Promise<any | null> {
        const existing = data.get(id);
        if (!existing) return null;
        const updated = { ...existing, ...updates, updatedAt: new Date() };
        data.set(id, updated);
        return updated;
      },

      async delete(id: string): Promise<boolean> {
        return data.delete(id);
      },

      async count(filter?: Record<string, any>): Promise<number> {
        const results = await this.findAll(filter);
        return results.length;
      },

      clear(): void {
        data.clear();
      },
    };
  }

  clear(): void {
    this.store.clear();
  }
}

// Usage:
// const db = new FakeDatabase();
// const users = db.collection('users');
// await users.create({ name: 'John', email: 'john@example.com' });
// const found = await users.findAll({ name: 'John' });
```

### Event Emitter Testing

```typescript
// Testing event-driven code
describe('EventBus', () => {
  let bus: EventBus;

  beforeEach(() => {
    bus = new EventBus();
  });

  it('should emit events to subscribers', () => {
    const handler = vi.fn();
    bus.on('user:created', handler);

    bus.emit('user:created', { id: '1', name: 'John' });

    expect(handler).toHaveBeenCalledWith({ id: '1', name: 'John' });
  });

  it('should support multiple subscribers', () => {
    const handler1 = vi.fn();
    const handler2 = vi.fn();
    bus.on('user:created', handler1);
    bus.on('user:created', handler2);

    bus.emit('user:created', { id: '1' });

    expect(handler1).toHaveBeenCalledTimes(1);
    expect(handler2).toHaveBeenCalledTimes(1);
  });

  it('should support unsubscribing', () => {
    const handler = vi.fn();
    const unsubscribe = bus.on('user:created', handler);

    unsubscribe();
    bus.emit('user:created', { id: '1' });

    expect(handler).not.toHaveBeenCalled();
  });

  it('should support once listeners', () => {
    const handler = vi.fn();
    bus.once('user:created', handler);

    bus.emit('user:created', { id: '1' });
    bus.emit('user:created', { id: '2' });

    expect(handler).toHaveBeenCalledTimes(1);
    expect(handler).toHaveBeenCalledWith({ id: '1' });
  });

  it('should handle async subscribers', async () => {
    const results: string[] = [];
    bus.on('user:created', async (data) => {
      await new Promise((r) => setTimeout(r, 10));
      results.push(`handler1:${data.id}`);
    });
    bus.on('user:created', async (data) => {
      results.push(`handler2:${data.id}`);
    });

    await bus.emitAsync('user:created', { id: '1' });

    expect(results).toEqual(['handler2:1', 'handler1:1']);
  });
});
```

## Testing Patterns for Common Scenarios

### State Machine Testing

```typescript
// Testing state transitions
describe('OrderStateMachine', () => {
  const validTransitions: Array<[OrderStatus, OrderStatus]> = [
    ['pending', 'processing'],
    ['pending', 'cancelled'],
    ['processing', 'shipped'],
    ['processing', 'cancelled'],
    ['shipped', 'delivered'],
    ['shipped', 'returned'],
    ['delivered', 'returned'],
  ];

  const invalidTransitions: Array<[OrderStatus, OrderStatus]> = [
    ['pending', 'shipped'],
    ['pending', 'delivered'],
    ['processing', 'pending'],
    ['processing', 'delivered'],
    ['shipped', 'pending'],
    ['shipped', 'processing'],
    ['delivered', 'pending'],
    ['delivered', 'processing'],
    ['delivered', 'shipped'],
    ['cancelled', 'pending'],
    ['cancelled', 'processing'],
    ['cancelled', 'shipped'],
    ['returned', 'pending'],
  ];

  it.each(validTransitions)(
    'should allow transition from %s to %s',
    (from, to) => {
      const machine = new OrderStateMachine(from);
      expect(() => machine.transition(to)).not.toThrow();
      expect(machine.currentState).toBe(to);
    }
  );

  it.each(invalidTransitions)(
    'should reject transition from %s to %s',
    (from, to) => {
      const machine = new OrderStateMachine(from);
      expect(() => machine.transition(to)).toThrow(
        `Invalid transition from ${from} to ${to}`
      );
      expect(machine.currentState).toBe(from);
    }
  );
});
```

### Pagination Testing

```typescript
describe('Pagination', () => {
  beforeEach(async () => {
    // Seed 25 items
    for (let i = 0; i < 25; i++) {
      await createItem({ name: `Item ${String(i).padStart(2, '0')}` });
    }
  });

  it('should return first page with default limit', async () => {
    const result = await listItems({ page: 1 });
    expect(result.data).toHaveLength(10);
    expect(result.pagination).toEqual({
      page: 1,
      limit: 10,
      total: 25,
      totalPages: 3,
      hasNext: true,
      hasPrev: false,
    });
  });

  it('should return second page', async () => {
    const result = await listItems({ page: 2 });
    expect(result.data).toHaveLength(10);
    expect(result.pagination.hasNext).toBe(true);
    expect(result.pagination.hasPrev).toBe(true);
  });

  it('should return partial last page', async () => {
    const result = await listItems({ page: 3 });
    expect(result.data).toHaveLength(5);
    expect(result.pagination.hasNext).toBe(false);
    expect(result.pagination.hasPrev).toBe(true);
  });

  it('should return empty for page beyond total', async () => {
    const result = await listItems({ page: 10 });
    expect(result.data).toHaveLength(0);
  });

  it('should support custom limit', async () => {
    const result = await listItems({ page: 1, limit: 5 });
    expect(result.data).toHaveLength(5);
    expect(result.pagination.totalPages).toBe(5);
  });

  it('should maintain consistent ordering across pages', async () => {
    const page1 = await listItems({ page: 1, limit: 5 });
    const page2 = await listItems({ page: 2, limit: 5 });

    const page1Ids = page1.data.map((i) => i.id);
    const page2Ids = page2.data.map((i) => i.id);

    // No overlap
    expect(page1Ids.filter((id) => page2Ids.includes(id))).toHaveLength(0);
  });
});
```

### Rate Limiter Testing

```typescript
describe('RateLimiter', () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('should allow requests within the limit', () => {
    const limiter = new RateLimiter({ maxRequests: 5, windowMs: 60_000 });

    for (let i = 0; i < 5; i++) {
      expect(limiter.tryAcquire('user-1')).toBe(true);
    }
  });

  it('should reject requests exceeding the limit', () => {
    const limiter = new RateLimiter({ maxRequests: 5, windowMs: 60_000 });

    for (let i = 0; i < 5; i++) {
      limiter.tryAcquire('user-1');
    }

    expect(limiter.tryAcquire('user-1')).toBe(false);
  });

  it('should reset after the time window', () => {
    const limiter = new RateLimiter({ maxRequests: 5, windowMs: 60_000 });

    for (let i = 0; i < 5; i++) {
      limiter.tryAcquire('user-1');
    }

    expect(limiter.tryAcquire('user-1')).toBe(false);

    // Advance time past the window
    vi.advanceTimersByTime(60_001);

    expect(limiter.tryAcquire('user-1')).toBe(true);
  });

  it('should track limits per user', () => {
    const limiter = new RateLimiter({ maxRequests: 2, windowMs: 60_000 });

    limiter.tryAcquire('user-1');
    limiter.tryAcquire('user-1');

    // user-1 is limited
    expect(limiter.tryAcquire('user-1')).toBe(false);

    // user-2 is not limited
    expect(limiter.tryAcquire('user-2')).toBe(true);
  });

  it('should return remaining count', () => {
    const limiter = new RateLimiter({ maxRequests: 5, windowMs: 60_000 });

    expect(limiter.remaining('user-1')).toBe(5);

    limiter.tryAcquire('user-1');
    expect(limiter.remaining('user-1')).toBe(4);

    limiter.tryAcquire('user-1');
    limiter.tryAcquire('user-1');
    expect(limiter.remaining('user-1')).toBe(2);
  });
});
```

### Cache Testing

```typescript
describe('Cache', () => {
  let cache: Cache;

  beforeEach(() => {
    vi.useFakeTimers();
    cache = new Cache({ defaultTTL: 60_000 });
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('should store and retrieve values', async () => {
    await cache.set('key', 'value');
    expect(await cache.get('key')).toBe('value');
  });

  it('should return null for missing keys', async () => {
    expect(await cache.get('nonexistent')).toBeNull();
  });

  it('should expire entries after TTL', async () => {
    await cache.set('key', 'value', { ttl: 5000 });
    expect(await cache.get('key')).toBe('value');

    vi.advanceTimersByTime(5001);
    expect(await cache.get('key')).toBeNull();
  });

  it('should support getOrSet pattern', async () => {
    const fetcher = vi.fn().mockResolvedValue('computed-value');

    const result1 = await cache.getOrSet('key', fetcher);
    expect(result1).toBe('computed-value');
    expect(fetcher).toHaveBeenCalledTimes(1);

    const result2 = await cache.getOrSet('key', fetcher);
    expect(result2).toBe('computed-value');
    expect(fetcher).toHaveBeenCalledTimes(1); // Not called again
  });

  it('should support invalidation', async () => {
    await cache.set('key1', 'value1');
    await cache.set('key2', 'value2');

    await cache.delete('key1');

    expect(await cache.get('key1')).toBeNull();
    expect(await cache.get('key2')).toBe('value2');
  });

  it('should support pattern invalidation', async () => {
    await cache.set('user:1:profile', 'data1');
    await cache.set('user:1:settings', 'data2');
    await cache.set('user:2:profile', 'data3');

    await cache.deletePattern('user:1:*');

    expect(await cache.get('user:1:profile')).toBeNull();
    expect(await cache.get('user:1:settings')).toBeNull();
    expect(await cache.get('user:2:profile')).toBe('data3');
  });

  it('should handle complex objects', async () => {
    const data = {
      user: { name: 'John', email: 'john@test.com' },
      items: [1, 2, 3],
      nested: { deep: { value: true } },
    };

    await cache.set('complex', data);
    expect(await cache.get('complex')).toEqual(data);
  });
});
```

### Retry Logic Testing

```typescript
describe('retry', () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('should succeed on first try', async () => {
    const fn = vi.fn().mockResolvedValue('success');
    const result = await retry(fn, { maxRetries: 3 });
    expect(result).toBe('success');
    expect(fn).toHaveBeenCalledTimes(1);
  });

  it('should retry on failure and eventually succeed', async () => {
    const fn = vi.fn()
      .mockRejectedValueOnce(new Error('fail 1'))
      .mockRejectedValueOnce(new Error('fail 2'))
      .mockResolvedValue('success');

    const resultPromise = retry(fn, { maxRetries: 3, delay: 1000 });

    // First retry after 1s
    await vi.advanceTimersByTimeAsync(1000);
    // Second retry after 2s (exponential backoff)
    await vi.advanceTimersByTimeAsync(2000);

    const result = await resultPromise;
    expect(result).toBe('success');
    expect(fn).toHaveBeenCalledTimes(3);
  });

  it('should throw after max retries', async () => {
    const fn = vi.fn().mockRejectedValue(new Error('always fails'));

    const resultPromise = retry(fn, { maxRetries: 2, delay: 100 });

    await vi.advanceTimersByTimeAsync(100);
    await vi.advanceTimersByTimeAsync(200);

    await expect(resultPromise).rejects.toThrow('always fails');
    expect(fn).toHaveBeenCalledTimes(3); // 1 initial + 2 retries
  });

  it('should use exponential backoff', async () => {
    const fn = vi.fn().mockRejectedValue(new Error('fail'));
    const delays: number[] = [];

    vi.spyOn(global, 'setTimeout').mockImplementation((callback: any, delay?: number) => {
      if (delay && delay > 0) delays.push(delay);
      callback();
      return 0 as any;
    });

    try {
      await retry(fn, { maxRetries: 3, delay: 1000, backoffMultiplier: 2 });
    } catch {}

    expect(delays).toEqual([1000, 2000, 4000]);
  });

  it('should not retry non-retryable errors', async () => {
    const fn = vi.fn().mockRejectedValue(new ValidationError('Invalid input'));

    await expect(
      retry(fn, { maxRetries: 3, retryableErrors: [NetworkError, TimeoutError] })
    ).rejects.toThrow('Invalid input');

    expect(fn).toHaveBeenCalledTimes(1);
  });
});
```

### Middleware Testing

```typescript
describe('Authentication Middleware', () => {
  let req: Partial<Request>;
  let res: Partial<Response>;
  let next: vi.Mock;

  beforeEach(() => {
    req = { headers: {} };
    res = {
      status: vi.fn().mockReturnThis(),
      json: vi.fn().mockReturnThis(),
    };
    next = vi.fn();
  });

  it('should call next() for valid token', async () => {
    const token = createAuthToken({ id: '1', email: 'test@test.com', role: 'user' });
    req.headers = { authorization: `Bearer ${token}` };

    await authMiddleware(req as Request, res as Response, next);

    expect(next).toHaveBeenCalledWith();
    expect((req as any).user).toBeDefined();
    expect((req as any).user.id).toBe('1');
  });

  it('should return 401 for missing token', async () => {
    await authMiddleware(req as Request, res as Response, next);

    expect(res.status).toHaveBeenCalledWith(401);
    expect(res.json).toHaveBeenCalledWith({ error: 'Authentication required' });
    expect(next).not.toHaveBeenCalled();
  });

  it('should return 401 for expired token', async () => {
    vi.useFakeTimers();
    const token = createAuthToken({ id: '1', email: 'test@test.com', role: 'user' });
    vi.advanceTimersByTime(2 * 60 * 60 * 1000); // 2 hours

    req.headers = { authorization: `Bearer ${token}` };
    await authMiddleware(req as Request, res as Response, next);

    expect(res.status).toHaveBeenCalledWith(401);
    expect(res.json).toHaveBeenCalledWith({ error: 'Token expired' });
    vi.useRealTimers();
  });

  it('should return 401 for malformed token', async () => {
    req.headers = { authorization: 'Bearer invalid.token.here' };

    await authMiddleware(req as Request, res as Response, next);

    expect(res.status).toHaveBeenCalledWith(401);
    expect(next).not.toHaveBeenCalled();
  });
});
```

### Stream/Iterator Testing

```typescript
describe('Stream Processing', () => {
  it('should process items from async iterator', async () => {
    async function* generateItems() {
      yield { id: 1, value: 'a' };
      yield { id: 2, value: 'b' };
      yield { id: 3, value: 'c' };
    }

    const results: any[] = [];
    for await (const item of processStream(generateItems())) {
      results.push(item);
    }

    expect(results).toEqual([
      { id: 1, value: 'A', processed: true },
      { id: 2, value: 'B', processed: true },
      { id: 3, value: 'C', processed: true },
    ]);
  });

  it('should handle empty streams', async () => {
    async function* emptyStream() {
      // yields nothing
    }

    const results: any[] = [];
    for await (const item of processStream(emptyStream())) {
      results.push(item);
    }

    expect(results).toEqual([]);
  });

  it('should handle stream errors gracefully', async () => {
    async function* failingStream() {
      yield { id: 1, value: 'a' };
      throw new Error('Stream interrupted');
    }

    const results: any[] = [];
    const errors: Error[] = [];

    for await (const item of processStream(failingStream(), { onError: (e) => errors.push(e) })) {
      results.push(item);
    }

    expect(results).toHaveLength(1);
    expect(errors).toHaveLength(1);
    expect(errors[0].message).toBe('Stream interrupted');
  });
});
```

### Error Boundary Testing (React)

```typescript
describe('ErrorBoundary', () => {
  const ThrowingComponent = ({ shouldThrow }: { shouldThrow: boolean }) => {
    if (shouldThrow) throw new Error('Component exploded');
    return <div>All good</div>;
  };

  it('should render children when no error', () => {
    render(
      <ErrorBoundary fallback={<div>Error</div>}>
        <ThrowingComponent shouldThrow={false} />
      </ErrorBoundary>
    );

    expect(screen.getByText('All good')).toBeInTheDocument();
  });

  it('should render fallback when child throws', () => {
    // Suppress console.error for this test
    const spy = vi.spyOn(console, 'error').mockImplementation(() => {});

    render(
      <ErrorBoundary fallback={<div>Something went wrong</div>}>
        <ThrowingComponent shouldThrow={true} />
      </ErrorBoundary>
    );

    expect(screen.getByText('Something went wrong')).toBeInTheDocument();
    expect(screen.queryByText('All good')).not.toBeInTheDocument();

    spy.mockRestore();
  });

  it('should call onError callback', () => {
    const onError = vi.fn();
    const spy = vi.spyOn(console, 'error').mockImplementation(() => {});

    render(
      <ErrorBoundary fallback={<div>Error</div>} onError={onError}>
        <ThrowingComponent shouldThrow={true} />
      </ErrorBoundary>
    );

    expect(onError).toHaveBeenCalledWith(
      expect.objectContaining({ message: 'Component exploded' }),
      expect.any(Object)
    );

    spy.mockRestore();
  });

  it('should support retry via render prop', async () => {
    let throwCount = 0;
    const ConditionalThrower = () => {
      throwCount++;
      if (throwCount <= 1) throw new Error('First render fails');
      return <div>Recovered</div>;
    };

    const spy = vi.spyOn(console, 'error').mockImplementation(() => {});
    const user = userEvent.setup();

    render(
      <ErrorBoundary
        fallback={({ retry }) => (
          <div>
            <p>Error occurred</p>
            <button onClick={retry}>Retry</button>
          </div>
        )}
      >
        <ConditionalThrower />
      </ErrorBoundary>
    );

    expect(screen.getByText('Error occurred')).toBeInTheDocument();

    await user.click(screen.getByText('Retry'));

    expect(screen.getByText('Recovered')).toBeInTheDocument();

    spy.mockRestore();
  });
});
```

## Testing Utilities

### Wait For Condition

```typescript
// tests/utils/wait-for.ts
export async function waitForCondition(
  condition: () => boolean | Promise<boolean>,
  options: { timeout?: number; interval?: number; message?: string } = {}
): Promise<void> {
  const { timeout = 5000, interval = 100, message = 'Condition not met' } = options;
  const startTime = Date.now();

  while (Date.now() - startTime < timeout) {
    if (await condition()) return;
    await new Promise((resolve) => setTimeout(resolve, interval));
  }

  throw new Error(`Timeout: ${message}`);
}

// Usage:
// await waitForCondition(
//   () => queue.length === 0,
//   { timeout: 10000, message: 'Queue did not drain' }
// );
```

### Snapshot Serializer

```typescript
// tests/utils/serializers.ts
import { expect } from 'vitest';

// Custom serializer that strips dynamic values
expect.addSnapshotSerializer({
  test(val) {
    return val && typeof val === 'object' && 'id' in val && 'createdAt' in val;
  },
  serialize(val, config, indentation, depth, refs, printer) {
    const { id, createdAt, updatedAt, ...rest } = val;
    return printer(
      {
        ...rest,
        id: '[UUID]',
        createdAt: '[DATE]',
        updatedAt: '[DATE]',
      },
      config,
      indentation,
      depth,
      refs
    );
  },
});
```

### Test Data Generators

```typescript
// tests/utils/generators.ts
import { faker } from '@faker-js/faker';

export const generators = {
  email: () => faker.internet.email().toLowerCase(),
  uuid: () => faker.string.uuid(),
  phone: () => faker.phone.number(),
  password: (length = 12) => faker.internet.password({ length }),
  slug: () => faker.helpers.slugify(faker.lorem.words(3)).toLowerCase(),
  ipAddress: () => faker.internet.ipv4(),
  url: () => faker.internet.url(),
  date: {
    past: () => faker.date.past(),
    future: () => faker.date.future(),
    recent: () => faker.date.recent(),
    between: (from: Date, to: Date) => faker.date.between({ from, to }),
  },
  number: {
    int: (min = 0, max = 1000) => faker.number.int({ min, max }),
    float: (min = 0, max = 1000) => faker.number.float({ min, max, fractionDigits: 2 }),
    price: () => parseFloat(faker.commerce.price({ min: 1, max: 999 })),
  },
  text: {
    sentence: () => faker.lorem.sentence(),
    paragraph: () => faker.lorem.paragraph(),
    words: (count = 3) => faker.lorem.words(count),
  },
};
```

## Testing Decision Matrix

```
┌───────────────────────────────┬──────────────────────────┬──────────────────────┐
│ Scenario                      │ Pattern                  │ Why                  │
├───────────────────────────────┼──────────────────────────┼──────────────────────┤
│ Create test data              │ Factory / Builder        │ Consistent, flexible │
│ External API                  │ MSW / mock fetch         │ No real HTTP calls   │
│ Database queries              │ Transaction rollback     │ Fast cleanup         │
│ Time-dependent code           │ Fake timers              │ Deterministic        │
│ File system operations        │ In-memory FS / tmpdir    │ No side effects      │
│ Authentication                │ JWT factory              │ No real auth server  │
│ Random/UUID values            │ Seed faker / fixed vals  │ Reproducible         │
│ Event-driven code             │ Spy on event handlers    │ Verify emissions     │
│ Concurrent operations         │ Promise.all assertions   │ Test race conditions │
│ Error paths                   │ Mock rejection/throw     │ Cover error handling │
│ Complex state transitions     │ Table-driven tests       │ Cover all paths      │
│ Component rendering           │ Testing Library queries  │ Test user experience │
│ Hook behavior                 │ renderHook               │ Isolate hook logic   │
│ Middleware                    │ Mock req/res/next        │ Test in isolation    │
│ Background jobs               │ Manual trigger + assert  │ No scheduler needed  │
│ WebSocket events              │ Mock WS server           │ No real connection   │
│ Cache behavior                │ Fake timers + assertions │ Test TTL, eviction   │
│ Pagination                    │ Seed data + page queries │ Verify boundaries    │
│ Search/Filter                 │ Seed varied data + query │ Test combinations    │
│ Rate limiting                 │ Fake timers + burst test │ Verify limits/reset  │
└───────────────────────────────┴──────────────────────────┴──────────────────────┘
```
