---
name: behavioral-patterns
description: >
  Behavioral design patterns in TypeScript — Strategy, Observer, Command,
  State Machine, Chain of Responsibility, and event-driven patterns
  with practical examples.
  Triggers: "strategy pattern", "observer pattern", "command pattern",
  "state machine", "event bus", "chain of responsibility", "behavioral pattern",
  "pub sub", "middleware pattern".
  NOT for: Object creation (use creational-patterns).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Behavioral Patterns

## Strategy

Replace switch/if-else chains with interchangeable algorithms.

```typescript
// Strategy interface
interface PricingStrategy {
  calculate(basePrice: number, quantity: number): number;
}

// Concrete strategies
class StandardPricing implements PricingStrategy {
  calculate(basePrice: number, quantity: number): number {
    return basePrice * quantity;
  }
}

class BulkPricing implements PricingStrategy {
  calculate(basePrice: number, quantity: number): number {
    const discount = quantity >= 100 ? 0.2 : quantity >= 50 ? 0.1 : 0;
    return basePrice * quantity * (1 - discount);
  }
}

class SubscriptionPricing implements PricingStrategy {
  constructor(private readonly monthlyDiscount: number) {}

  calculate(basePrice: number, quantity: number): number {
    return basePrice * quantity * (1 - this.monthlyDiscount);
  }
}

// Context
class OrderCalculator {
  constructor(private strategy: PricingStrategy) {}

  setStrategy(strategy: PricingStrategy): void {
    this.strategy = strategy;
  }

  calculateTotal(items: Array<{ price: number; quantity: number }>): number {
    return items.reduce(
      (total, item) => total + this.strategy.calculate(item.price, item.quantity),
      0
    );
  }
}

// Usage
const calculator = new OrderCalculator(new StandardPricing());
let total = calculator.calculateTotal(items);

// Switch strategy at runtime
if (customer.isSubscriber) {
  calculator.setStrategy(new SubscriptionPricing(0.15));
  total = calculator.calculateTotal(items);
}
```

### Function-Based Strategy (Simpler)

```typescript
type SortStrategy<T> = (a: T, b: T) => number;

const sortByName: SortStrategy<User> = (a, b) => a.name.localeCompare(b.name);
const sortByDate: SortStrategy<User> = (a, b) => b.createdAt.getTime() - a.createdAt.getTime();
const sortByScore: SortStrategy<User> = (a, b) => b.score - a.score;

function sortUsers(users: User[], strategy: SortStrategy<User>): User[] {
  return [...users].sort(strategy);
}

// Usage — strategy is just a function
const sorted = sortUsers(users, sortByScore);

// Strategy map
const sortStrategies: Record<string, SortStrategy<User>> = {
  name: sortByName,
  date: sortByDate,
  score: sortByScore,
};

const sorted2 = sortUsers(users, sortStrategies[req.query.sort ?? "name"]);
```

## Observer / Event Emitter

Decouple components that react to state changes.

```typescript
type EventMap = {
  "user:created": { user: User };
  "user:updated": { user: User; changes: Partial<User> };
  "user:deleted": { userId: string };
  "order:placed": { order: Order };
  "order:shipped": { orderId: string; trackingNumber: string };
};

type EventHandler<T> = (data: T) => void | Promise<void>;

class TypedEventBus {
  private handlers = new Map<string, Set<EventHandler<any>>>();

  on<K extends keyof EventMap>(event: K, handler: EventHandler<EventMap[K]>): () => void {
    if (!this.handlers.has(event)) {
      this.handlers.set(event, new Set());
    }
    this.handlers.get(event)!.add(handler);

    // Return unsubscribe function
    return () => this.handlers.get(event)?.delete(handler);
  }

  async emit<K extends keyof EventMap>(event: K, data: EventMap[K]): Promise<void> {
    const handlers = this.handlers.get(event);
    if (!handlers) return;

    // Run all handlers concurrently
    await Promise.allSettled(
      [...handlers].map((handler) => handler(data))
    );
  }

  once<K extends keyof EventMap>(event: K, handler: EventHandler<EventMap[K]>): void {
    const unsubscribe = this.on(event, (data) => {
      unsubscribe();
      handler(data);
    });
  }
}

// Usage
const bus = new TypedEventBus();

// Subscribe — each handler is independent
bus.on("user:created", async ({ user }) => {
  await sendWelcomeEmail(user.email);
});

bus.on("user:created", async ({ user }) => {
  await analytics.track("signup", { userId: user.id });
});

bus.on("order:placed", async ({ order }) => {
  await inventoryService.reserve(order.items);
});

// Emit — all handlers run
await bus.emit("user:created", { user: newUser });
```

## Command

Encapsulate operations as objects for undo/redo, queuing, or logging.

```typescript
interface Command {
  execute(): Promise<void>;
  undo(): Promise<void>;
  describe(): string;
}

// Concrete commands
class AddItemCommand implements Command {
  private previousState: CartItem[] | null = null;

  constructor(
    private readonly cart: Cart,
    private readonly item: CartItem,
  ) {}

  async execute(): Promise<void> {
    this.previousState = [...this.cart.items];
    this.cart.addItem(this.item);
  }

  async undo(): Promise<void> {
    if (this.previousState) {
      this.cart.items = this.previousState;
    }
  }

  describe(): string {
    return `Add ${this.item.name} to cart`;
  }
}

class UpdateQuantityCommand implements Command {
  private previousQuantity: number = 0;

  constructor(
    private readonly cart: Cart,
    private readonly itemId: string,
    private readonly newQuantity: number,
  ) {}

  async execute(): Promise<void> {
    const item = this.cart.findItem(this.itemId);
    if (!item) throw new Error(`Item ${this.itemId} not found`);
    this.previousQuantity = item.quantity;
    item.quantity = this.newQuantity;
  }

  async undo(): Promise<void> {
    const item = this.cart.findItem(this.itemId);
    if (item) item.quantity = this.previousQuantity;
  }

  describe(): string {
    return `Update quantity to ${this.newQuantity}`;
  }
}

// Command history (undo/redo manager)
class CommandHistory {
  private executed: Command[] = [];
  private undone: Command[] = [];

  async execute(command: Command): Promise<void> {
    await command.execute();
    this.executed.push(command);
    this.undone = []; // Clear redo stack on new command
  }

  async undo(): Promise<void> {
    const command = this.executed.pop();
    if (!command) return;
    await command.undo();
    this.undone.push(command);
  }

  async redo(): Promise<void> {
    const command = this.undone.pop();
    if (!command) return;
    await command.execute();
    this.executed.push(command);
  }

  get canUndo(): boolean { return this.executed.length > 0; }
  get canRedo(): boolean { return this.undone.length > 0; }
  get history(): string[] { return this.executed.map((c) => c.describe()); }
}

// Usage
const history = new CommandHistory();
await history.execute(new AddItemCommand(cart, newItem));
await history.execute(new UpdateQuantityCommand(cart, "item-1", 3));
await history.undo(); // Reverts quantity change
await history.redo(); // Re-applies quantity change
```

## State Machine

Manage complex state transitions explicitly.

```typescript
type OrderState = "draft" | "pending" | "confirmed" | "shipped" | "delivered" | "cancelled";

type OrderEvent =
  | { type: "SUBMIT" }
  | { type: "CONFIRM"; confirmedBy: string }
  | { type: "SHIP"; trackingNumber: string }
  | { type: "DELIVER" }
  | { type: "CANCEL"; reason: string };

interface StateTransition {
  from: OrderState;
  event: OrderEvent["type"];
  to: OrderState;
  guard?: (order: Order, event: OrderEvent) => boolean;
  action?: (order: Order, event: OrderEvent) => void;
}

const transitions: StateTransition[] = [
  { from: "draft", event: "SUBMIT", to: "pending" },
  { from: "pending", event: "CONFIRM", to: "confirmed",
    action: (order, event) => {
      if (event.type === "CONFIRM") order.confirmedBy = event.confirmedBy;
    }
  },
  { from: "confirmed", event: "SHIP", to: "shipped",
    action: (order, event) => {
      if (event.type === "SHIP") order.trackingNumber = event.trackingNumber;
    }
  },
  { from: "shipped", event: "DELIVER", to: "delivered" },
  // Cancel from multiple states
  { from: "draft", event: "CANCEL", to: "cancelled" },
  { from: "pending", event: "CANCEL", to: "cancelled" },
  { from: "confirmed", event: "CANCEL", to: "cancelled",
    guard: (order) => !order.trackingNumber,  // Can't cancel if shipped
    action: (order, event) => {
      if (event.type === "CANCEL") order.cancelReason = event.reason;
    }
  },
];

class OrderStateMachine {
  transition(order: Order, event: OrderEvent): Order {
    const transition = transitions.find(
      (t) => t.from === order.state && t.event === event.type
    );

    if (!transition) {
      throw new Error(`Invalid transition: ${order.state} + ${event.type}`);
    }

    if (transition.guard && !transition.guard(order, event)) {
      throw new Error(`Guard failed: ${order.state} + ${event.type}`);
    }

    const updated = { ...order, state: transition.to as OrderState };
    transition.action?.(updated, event);
    return updated;
  }

  validEvents(state: OrderState): OrderEvent["type"][] {
    return transitions
      .filter((t) => t.from === state)
      .map((t) => t.event);
  }
}

// Usage
const machine = new OrderStateMachine();
let order = { id: "1", state: "draft" as OrderState };
order = machine.transition(order, { type: "SUBMIT" });           // -> pending
order = machine.transition(order, { type: "CONFIRM", confirmedBy: "admin" }); // -> confirmed
order = machine.transition(order, { type: "SHIP", trackingNumber: "1Z..." }); // -> shipped
```

## Chain of Responsibility (Middleware)

```typescript
// Express-like middleware pattern
type Context = {
  req: Request;
  res: Response;
  user?: User;
  startTime?: number;
};

type Middleware = (ctx: Context, next: () => Promise<void>) => Promise<void>;

class Pipeline {
  private middlewares: Middleware[] = [];

  use(middleware: Middleware): this {
    this.middlewares.push(middleware);
    return this;
  }

  async execute(ctx: Context): Promise<void> {
    let index = 0;
    const next = async (): Promise<void> => {
      if (index >= this.middlewares.length) return;
      const middleware = this.middlewares[index++];
      await middleware(ctx, next);
    };
    await next();
  }
}

// Middleware implementations
const timing: Middleware = async (ctx, next) => {
  ctx.startTime = Date.now();
  await next();
  const duration = Date.now() - ctx.startTime;
  ctx.res.setHeader("X-Response-Time", `${duration}ms`);
};

const auth: Middleware = async (ctx, next) => {
  const token = ctx.req.headers.authorization?.replace("Bearer ", "");
  if (!token) { ctx.res.status(401).json({ error: "Unauthorized" }); return; }
  ctx.user = await verifyToken(token);
  await next();
};

const rateLimit: Middleware = async (ctx, next) => {
  const key = ctx.req.ip;
  const count = await redis.incr(key);
  if (count === 1) await redis.expire(key, 60);
  if (count > 100) { ctx.res.status(429).json({ error: "Too many requests" }); return; }
  await next();
};

// Build pipeline
const pipeline = new Pipeline()
  .use(timing)
  .use(rateLimit)
  .use(auth);

await pipeline.execute(ctx);
```

## Result Type (Error Handling Pattern)

```typescript
type Result<T, E = Error> =
  | { ok: true; value: T }
  | { ok: false; error: E };

function ok<T>(value: T): Result<T, never> {
  return { ok: true, value };
}

function err<E>(error: E): Result<never, E> {
  return { ok: false, error };
}

// Domain errors
type UserError =
  | { type: "NOT_FOUND"; userId: string }
  | { type: "DUPLICATE_EMAIL"; email: string }
  | { type: "INVALID_PASSWORD"; reason: string };

async function createUser(data: CreateUserDto): Promise<Result<User, UserError>> {
  const existing = await userRepo.findByEmail(data.email);
  if (existing) {
    return err({ type: "DUPLICATE_EMAIL", email: data.email });
  }

  if (data.password.length < 8) {
    return err({ type: "INVALID_PASSWORD", reason: "Too short" });
  }

  const user = await userRepo.save(User.create(data));
  return ok(user);
}

// Usage — exhaustive error handling
const result = await createUser(data);
if (!result.ok) {
  switch (result.error.type) {
    case "NOT_FOUND":
      return res.status(404).json({ error: `User ${result.error.userId} not found` });
    case "DUPLICATE_EMAIL":
      return res.status(409).json({ error: `Email ${result.error.email} already exists` });
    case "INVALID_PASSWORD":
      return res.status(400).json({ error: result.error.reason });
    default:
      result.error satisfies never; // Exhaustiveness check
  }
}

const user = result.value; // TypeScript knows this is User
```

## Gotchas

1. **Strategy pattern overkill for 2 options** — If you only have two strategies and won't add more, a simple `if/else` is clearer. Extract a Strategy when you have 3+ variants or when the algorithms are complex enough to warrant their own test files.

2. **Observer memory leaks** — Every `.on()` creates a reference. If the observer outlives the subscriber (e.g., component unmounts but event handler persists), you leak memory. Always call the unsubscribe function returned by `.on()` in cleanup.

3. **Command undo without snapshot** — `undo()` must restore the EXACT previous state. Storing `previousState` in the command is the simplest approach. Without it, undo is unreliable.

4. **State machine without exhaustive event handling** — If your state machine silently ignores invalid transitions, bugs hide. Always throw on invalid transitions and provide a `validEvents(state)` method for UI state.

5. **Middleware order matters** — In a Chain of Responsibility, each middleware can short-circuit by not calling `next()`. Auth before rate-limiting means unauthenticated users skip rate limits. Think about the correct order.

6. **EventBus hides dependencies** — Components that communicate via event bus have implicit dependencies. Debugging becomes harder because the call path isn't traceable through code. Use event bus for truly decoupled concerns (analytics, logging), not for core business logic flow.
