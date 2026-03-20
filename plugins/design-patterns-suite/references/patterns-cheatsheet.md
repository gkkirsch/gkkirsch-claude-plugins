# Design Patterns Cheatsheet

## Creational Patterns

### Factory Method

```typescript
interface Product { execute(): void; }
class ConcreteA implements Product { execute() { /* ... */ } }
class ConcreteB implements Product { execute() { /* ... */ } }

// Registry-based (open/closed)
const registry = new Map<string, () => Product>();
registry.set("a", () => new ConcreteA());
registry.set("b", () => new ConcreteB());

function create(type: string): Product {
  const factory = registry.get(type);
  if (!factory) throw new Error(`Unknown: ${type}`);
  return factory();
}
```

### Builder

```typescript
class QueryBuilder {
  private parts: string[] = [];
  select(cols: string) { this.parts.push(`SELECT ${cols}`); return this; }
  from(table: string) { this.parts.push(`FROM ${table}`); return this; }
  where(cond: string) { this.parts.push(`WHERE ${cond}`); return this; }
  build() { return this.parts.join(" "); }
}

const sql = new QueryBuilder().select("*").from("users").where("active = true").build();
```

### Singleton

```typescript
// Module-level (preferred in TS)
const pool = new Pool({ max: 20 });
export const db = { query: (sql: string, params?: any[]) => pool.query(sql, params) };

// Class-based
class Database {
  private static instance: Database;
  static getInstance(): Database {
    return (Database.instance ??= new Database());
  }
}
```

### Dependency Injection

```typescript
interface Logger { log(msg: string): void; }
interface UserRepo { findById(id: string): Promise<User | null>; }

class UserService {
  constructor(private repo: UserRepo, private logger: Logger) {}
  async getUser(id: string) {
    this.logger.log(`Fetching user ${id}`);
    return this.repo.findById(id);
  }
}

// Compose at entry point
const service = new UserService(new PostgresUserRepo(db), new PinoLogger());
```

## Behavioral Patterns

### Strategy

```typescript
// Function-based (simplest)
type Strategy<T> = (a: T, b: T) => number;
const byName: Strategy<User> = (a, b) => a.name.localeCompare(b.name);
const byDate: Strategy<User> = (a, b) => b.createdAt.getTime() - a.createdAt.getTime();

const strategies: Record<string, Strategy<User>> = { name: byName, date: byDate };
const sorted = [...users].sort(strategies[sortKey]);
```

### Observer / Event Bus

```typescript
type EventMap = { "user:created": { user: User }; "order:placed": { order: Order }; };
type Handler<T> = (data: T) => void | Promise<void>;

class EventBus {
  private handlers = new Map<string, Set<Handler<any>>>();
  on<K extends keyof EventMap>(event: K, handler: Handler<EventMap[K]>): () => void {
    if (!this.handlers.has(event)) this.handlers.set(event, new Set());
    this.handlers.get(event)!.add(handler);
    return () => this.handlers.get(event)?.delete(handler); // unsubscribe
  }
  async emit<K extends keyof EventMap>(event: K, data: EventMap[K]) {
    await Promise.allSettled([...(this.handlers.get(event) ?? [])].map(h => h(data)));
  }
}
```

### Command (Undo/Redo)

```typescript
interface Command { execute(): Promise<void>; undo(): Promise<void>; describe(): string; }

class History {
  private done: Command[] = [];
  private undone: Command[] = [];
  async execute(cmd: Command) { await cmd.execute(); this.done.push(cmd); this.undone = []; }
  async undo() { const c = this.done.pop(); if (c) { await c.undo(); this.undone.push(c); } }
  async redo() { const c = this.undone.pop(); if (c) { await c.execute(); this.done.push(c); } }
}
```

### State Machine

```typescript
type State = "draft" | "pending" | "confirmed" | "shipped" | "cancelled";
type Event = { type: "SUBMIT" } | { type: "CONFIRM" } | { type: "SHIP" } | { type: "CANCEL" };

const transitions: Array<{ from: State; event: string; to: State }> = [
  { from: "draft", event: "SUBMIT", to: "pending" },
  { from: "pending", event: "CONFIRM", to: "confirmed" },
  { from: "confirmed", event: "SHIP", to: "shipped" },
  { from: "draft", event: "CANCEL", to: "cancelled" },
  { from: "pending", event: "CANCEL", to: "cancelled" },
];

function transition(state: State, event: Event): State {
  const t = transitions.find(t => t.from === state && t.event === event.type);
  if (!t) throw new Error(`Invalid: ${state} + ${event.type}`);
  return t.to;
}
```

### Middleware / Chain of Responsibility

```typescript
type Ctx = { req: Request; res: Response; user?: User };
type Middleware = (ctx: Ctx, next: () => Promise<void>) => Promise<void>;

class Pipeline {
  private mw: Middleware[] = [];
  use(m: Middleware) { this.mw.push(m); return this; }
  async run(ctx: Ctx) {
    let i = 0;
    const next = async () => { if (i < this.mw.length) await this.mw[i++](ctx, next); };
    await next();
  }
}
```

### Result Type

```typescript
type Result<T, E = Error> = { ok: true; value: T } | { ok: false; error: E };
const ok = <T>(value: T): Result<T, never> => ({ ok: true, value });
const err = <E>(error: E): Result<never, E> => ({ ok: false, error });

// Usage
type UserError = { type: "NOT_FOUND" } | { type: "DUPLICATE_EMAIL"; email: string };
async function createUser(data: Dto): Promise<Result<User, UserError>> {
  if (await exists(data.email)) return err({ type: "DUPLICATE_EMAIL", email: data.email });
  return ok(await save(data));
}
```

## SOLID Quick Reference

| Principle | Meaning | Smell | Fix |
|-----------|---------|-------|-----|
| **S**ingle Responsibility | One reason to change | God class, 500+ line file | Extract focused classes/modules |
| **O**pen/Closed | Extend without modifying | Switch on type, if/else chains | Strategy, registry, polymorphism |
| **L**iskov Substitution | Subtypes are substitutable | Override that throws, empty impl | Composition over inheritance |
| **I**nterface Segregation | Small focused interfaces | Unused methods, fat interfaces | Split into role-specific interfaces |
| **D**ependency Inversion | Depend on abstractions | `new ConcreteClass()` in business logic | Constructor injection, interfaces |

## Code Smell to Pattern

| Smell | Pattern |
|-------|---------|
| Switch on type/string | Strategy or Factory |
| God class (does everything) | Extract + Facade |
| Copy-paste with variations | Template Method or Strategy |
| Complex object construction | Builder |
| Global mutable state | Singleton (module) + DI |
| Nested callbacks | Command or Pipeline |
| Implicit state transitions | State Machine |
| Tight coupling between modules | Observer / Event Bus |
| Deep inheritance hierarchy | Composition + DI |
| Primitive obsession (string types) | Value objects + branded types |

## When NOT to Use Patterns

- **2 options** — `if/else` beats Strategy
- **No variation planned** — direct code beats Factory
- **Single implementation** — interface without multiple implementations is overhead
- **Small codebase** — patterns add indirection; only worth it when complexity justifies it
- **Premature abstraction** — wait for the third occurrence before extracting a pattern
