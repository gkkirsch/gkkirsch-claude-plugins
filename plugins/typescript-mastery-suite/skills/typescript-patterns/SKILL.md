---
name: typescript-patterns
description: >
  TypeScript runtime patterns — Result type, builder pattern, dependency
  injection, repository pattern, event emitters, functional error handling,
  Zod validation, and type-safe API clients.
  Triggers: "typescript patterns", "result type", "builder pattern typescript",
  "dependency injection typescript", "zod validation", "typescript api client".
  NOT for: type system features (use typescript-advanced-types).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# TypeScript Runtime Patterns

## Result Type (No Exceptions)

```typescript
// Explicit success/failure instead of try/catch
type Result<T, E = Error> =
  | { ok: true; value: T }
  | { ok: false; error: E };

function Ok<T>(value: T): Result<T, never> {
  return { ok: true, value };
}

function Err<E>(error: E): Result<never, E> {
  return { ok: false, error };
}

// Usage in business logic
interface ValidationError {
  field: string;
  message: string;
}

function validateEmail(email: string): Result<string, ValidationError> {
  const trimmed = email.trim().toLowerCase();
  if (!trimmed.includes("@")) {
    return Err({ field: "email", message: "Invalid email format" });
  }
  if (trimmed.length > 254) {
    return Err({ field: "email", message: "Email too long" });
  }
  return Ok(trimmed);
}

function createUser(input: { email: string; name: string }): Result<User, ValidationError> {
  const emailResult = validateEmail(input.email);
  if (!emailResult.ok) return emailResult;  // Forward the error

  return Ok({
    id: crypto.randomUUID(),
    email: emailResult.value,
    name: input.name,
  });
}

// Chaining results
function map<T, U, E>(result: Result<T, E>, fn: (value: T) => U): Result<U, E> {
  return result.ok ? Ok(fn(result.value)) : result;
}

function flatMap<T, U, E>(result: Result<T, E>, fn: (value: T) => Result<U, E>): Result<U, E> {
  return result.ok ? fn(result.value) : result;
}
```

## Builder Pattern

```typescript
class QueryBuilder<T extends Record<string, unknown>> {
  private filters: string[] = [];
  private sortField?: string;
  private sortDir: "asc" | "desc" = "asc";
  private limitVal?: number;
  private offsetVal?: number;
  private selectedFields?: (keyof T)[];

  where(field: keyof T & string, op: "=" | "!=" | ">" | "<" | "like", value: unknown): this {
    this.filters.push(`${field} ${op} ${JSON.stringify(value)}`);
    return this;
  }

  orderBy(field: keyof T & string, direction: "asc" | "desc" = "asc"): this {
    this.sortField = field;
    this.sortDir = direction;
    return this;
  }

  limit(n: number): this {
    this.limitVal = n;
    return this;
  }

  offset(n: number): this {
    this.offsetVal = n;
    return this;
  }

  select(...fields: (keyof T & string)[]): this {
    this.selectedFields = fields;
    return this;
  }

  build(): { sql: string; params: unknown[] } {
    const parts: string[] = [];
    const params: unknown[] = [];

    const fields = this.selectedFields?.join(", ") || "*";
    parts.push(`SELECT ${fields} FROM table`);

    if (this.filters.length > 0) {
      parts.push(`WHERE ${this.filters.join(" AND ")}`);
    }

    if (this.sortField) {
      parts.push(`ORDER BY ${this.sortField} ${this.sortDir.toUpperCase()}`);
    }

    if (this.limitVal !== undefined) {
      parts.push(`LIMIT ${this.limitVal}`);
    }

    if (this.offsetVal !== undefined) {
      parts.push(`OFFSET ${this.offsetVal}`);
    }

    return { sql: parts.join(" "), params };
  }
}

// Fluent usage
const query = new QueryBuilder<{ name: string; age: number; active: boolean }>()
  .where("active", "=", true)
  .where("age", ">", 18)
  .orderBy("name", "asc")
  .limit(20)
  .offset(40)
  .select("name", "age")
  .build();
```

## Dependency Injection

```typescript
// Token-based DI (no decorators, no reflect-metadata)
class Container {
  private instances = new Map<string, unknown>();
  private factories = new Map<string, () => unknown>();

  register<T>(token: string, factory: () => T): void {
    this.factories.set(token, factory);
  }

  registerSingleton<T>(token: string, factory: () => T): void {
    this.factories.set(token, () => {
      if (!this.instances.has(token)) {
        this.instances.set(token, factory());
      }
      return this.instances.get(token);
    });
  }

  resolve<T>(token: string): T {
    const factory = this.factories.get(token);
    if (!factory) throw new Error(`No registration for: ${token}`);
    return factory() as T;
  }
}

// Tokens as constants (type-safe keys)
const TOKENS = {
  Database: "Database",
  UserRepo: "UserRepository",
  AuthService: "AuthService",
  Logger: "Logger",
} as const;

// Registration
const container = new Container();

container.registerSingleton(TOKENS.Logger, () => new ConsoleLogger());
container.registerSingleton(TOKENS.Database, () => new PostgresClient(process.env.DATABASE_URL!));
container.register(TOKENS.UserRepo, () =>
  new UserRepository(container.resolve(TOKENS.Database), container.resolve(TOKENS.Logger))
);
container.register(TOKENS.AuthService, () =>
  new AuthService(container.resolve(TOKENS.UserRepo), container.resolve(TOKENS.Logger))
);

// Usage
const authService = container.resolve<AuthService>(TOKENS.AuthService);
```

## Repository Pattern

```typescript
// Generic repository interface
interface Repository<T, CreateInput, UpdateInput> {
  findById(id: string): Promise<T | null>;
  findMany(options?: FindManyOptions): Promise<T[]>;
  create(data: CreateInput): Promise<T>;
  update(id: string, data: UpdateInput): Promise<T>;
  delete(id: string): Promise<void>;
  count(filter?: Record<string, unknown>): Promise<number>;
}

interface FindManyOptions {
  filter?: Record<string, unknown>;
  sort?: { field: string; direction: "asc" | "desc" };
  pagination?: { page: number; limit: number };
}

// Concrete implementation with Prisma
class PrismaUserRepository implements Repository<User, CreateUserInput, UpdateUserInput> {
  constructor(private prisma: PrismaClient) {}

  async findById(id: string): Promise<User | null> {
    return this.prisma.user.findUnique({ where: { id } });
  }

  async findMany(options?: FindManyOptions): Promise<User[]> {
    return this.prisma.user.findMany({
      where: options?.filter,
      orderBy: options?.sort
        ? { [options.sort.field]: options.sort.direction }
        : undefined,
      skip: options?.pagination
        ? (options.pagination.page - 1) * options.pagination.limit
        : undefined,
      take: options?.pagination?.limit,
    });
  }

  async create(data: CreateUserInput): Promise<User> {
    return this.prisma.user.create({ data });
  }

  async update(id: string, data: UpdateUserInput): Promise<User> {
    return this.prisma.user.update({ where: { id }, data });
  }

  async delete(id: string): Promise<void> {
    await this.prisma.user.delete({ where: { id } });
  }

  async count(filter?: Record<string, unknown>): Promise<number> {
    return this.prisma.user.count({ where: filter });
  }
}
```

## Type-Safe Event Emitter

```typescript
type EventMap = Record<string, unknown[]>;

class TypedEventEmitter<Events extends EventMap> {
  private listeners = new Map<keyof Events, Set<Function>>();

  on<E extends keyof Events>(event: E, listener: (...args: Events[E]) => void): () => void {
    if (!this.listeners.has(event)) {
      this.listeners.set(event, new Set());
    }
    this.listeners.get(event)!.add(listener);

    // Return unsubscribe function
    return () => this.listeners.get(event)?.delete(listener);
  }

  emit<E extends keyof Events>(event: E, ...args: Events[E]): void {
    this.listeners.get(event)?.forEach((listener) => listener(...args));
  }

  once<E extends keyof Events>(event: E, listener: (...args: Events[E]) => void): void {
    const unsub = this.on(event, (...args) => {
      unsub();
      listener(...args);
    });
  }

  removeAllListeners(event?: keyof Events): void {
    if (event) {
      this.listeners.delete(event);
    } else {
      this.listeners.clear();
    }
  }
}

// Define events with typed payloads
interface AppEvents extends EventMap {
  "user:login": [user: User];
  "user:logout": [userId: string];
  "order:created": [order: Order, notifyUser: boolean];
  "error": [error: Error, context: string];
}

const events = new TypedEventEmitter<AppEvents>();

// Type-safe: args are checked
events.on("user:login", (user) => {
  console.log(`${user.name} logged in`);  // user is typed as User
});

events.emit("order:created", order, true);  // Both args required
// events.emit("order:created", order);     // Error: missing second arg
```

## Zod Validation

```typescript
import { z } from "zod";

// Schema definition
const UserSchema = z.object({
  name: z.string().min(1, "Name is required").max(100),
  email: z.string().email("Invalid email"),
  age: z.number().int().min(13, "Must be at least 13").max(150),
  role: z.enum(["admin", "editor", "viewer"]),
  bio: z.string().max(500).optional(),
  tags: z.array(z.string()).max(10).default([]),
  preferences: z.object({
    theme: z.enum(["light", "dark"]).default("light"),
    notifications: z.boolean().default(true),
  }).default({}),
});

// Infer TypeScript type from schema
type User = z.infer<typeof UserSchema>;
// { name: string; email: string; age: number; role: "admin" | "editor" | "viewer"; ... }

// Validation
function createUser(input: unknown): Result<User, z.ZodError> {
  const result = UserSchema.safeParse(input);
  if (!result.success) return Err(result.error);
  return Ok(result.data);  // result.data is typed as User
}

// Transform & preprocess
const ApiResponseSchema = z.object({
  id: z.string(),
  created_at: z.string().transform((s) => new Date(s)),  // string → Date
  amount: z.string().transform(Number),  // "42.50" → 42.5
  tags: z.preprocess(
    (val) => (typeof val === "string" ? val.split(",") : val),
    z.array(z.string())
  ),
});

// Discriminated union schemas
const EventSchema = z.discriminatedUnion("type", [
  z.object({ type: z.literal("click"), x: z.number(), y: z.number() }),
  z.object({ type: z.literal("keypress"), key: z.string() }),
  z.object({ type: z.literal("scroll"), deltaY: z.number() }),
]);

// API request validation middleware
function validateBody<T extends z.ZodSchema>(schema: T) {
  return (req: Request, res: Response, next: NextFunction) => {
    const result = schema.safeParse(req.body);
    if (!result.success) {
      return res.status(400).json({
        error: "Validation failed",
        details: result.error.flatten().fieldErrors,
      });
    }
    req.body = result.data;
    next();
  };
}

// Usage: app.post("/users", validateBody(UserSchema), handler);
```

## Type-Safe API Client

```typescript
// Define API contract
interface ApiEndpoints {
  "GET /users": { query: { page?: number; limit?: number }; response: User[] };
  "GET /users/:id": { params: { id: string }; response: User };
  "POST /users": { body: CreateUserInput; response: User };
  "PUT /users/:id": { params: { id: string }; body: UpdateUserInput; response: User };
  "DELETE /users/:id": { params: { id: string }; response: void };
}

type Method = "GET" | "POST" | "PUT" | "DELETE";

class ApiClient {
  constructor(private baseUrl: string, private token?: string) {}

  async request<K extends keyof ApiEndpoints>(
    endpoint: K,
    options?: Omit<ApiEndpoints[K], "response">
  ): Promise<ApiEndpoints[K]["response"]> {
    const [method, path] = (endpoint as string).split(" ") as [Method, string];

    // Replace path params
    let url = path;
    if (options && "params" in options) {
      for (const [key, value] of Object.entries(options.params as Record<string, string>)) {
        url = url.replace(`:${key}`, encodeURIComponent(value));
      }
    }

    // Add query params
    if (options && "query" in options) {
      const params = new URLSearchParams();
      for (const [key, value] of Object.entries(options.query as Record<string, unknown>)) {
        if (value !== undefined) params.set(key, String(value));
      }
      if (params.toString()) url += `?${params}`;
    }

    const res = await fetch(`${this.baseUrl}${url}`, {
      method,
      headers: {
        "Content-Type": "application/json",
        ...(this.token ? { Authorization: `Bearer ${this.token}` } : {}),
      },
      body: options && "body" in options ? JSON.stringify(options.body) : undefined,
    });

    if (!res.ok) throw new Error(`API error: ${res.status}`);
    if (res.status === 204) return undefined as any;
    return res.json();
  }
}

// Usage — fully typed
const api = new ApiClient("https://api.example.com", "token");
const users = await api.request("GET /users", { query: { page: 1, limit: 20 } });
// users is User[]

const user = await api.request("GET /users/:id", { params: { id: "123" } });
// user is User
```

## Functional Pipe & Compose

```typescript
// Pipe: left to right
function pipe<A, B>(fn1: (a: A) => B): (a: A) => B;
function pipe<A, B, C>(fn1: (a: A) => B, fn2: (b: B) => C): (a: A) => C;
function pipe<A, B, C, D>(fn1: (a: A) => B, fn2: (b: B) => C, fn3: (c: C) => D): (a: A) => D;
function pipe(...fns: Function[]) {
  return (x: unknown) => fns.reduce((v, f) => f(v), x);
}

// Usage
const processUser = pipe(
  (input: RawUserInput) => validateInput(input),
  (validated) => normalizeEmail(validated),
  (normalized) => hashPassword(normalized),
  (hashed) => saveToDatabase(hashed)
);

// Async pipe
async function asyncPipe<T>(value: T, ...fns: ((v: any) => any | Promise<any>)[]): Promise<any> {
  let result: any = value;
  for (const fn of fns) {
    result = await fn(result);
  }
  return result;
}
```

## Gotchas

1. **Zod `.parse()` throws, `.safeParse()` returns Result** — In production, always use `.safeParse()` and handle the error case explicitly. `.parse()` throws a `ZodError` that propagates uncaught.

2. **Generic functions lose inference with default values** — `function fn<T = string>(x: T)` doesn't constrain T. If you call `fn(42)`, T is inferred as `number`, not `string`. Defaults only apply when T can't be inferred.

3. **Event emitter memory leaks** — Always store the unsubscribe function and call it on cleanup. In React: `useEffect(() => { const unsub = events.on("x", handler); return unsub; }, [])`.

4. **Builder pattern loses type safety without `this` return** — Each method must return `this`, not the class type. Otherwise, subclasses lose their extended methods in the chain.

5. **DI containers hide type errors at registration** — Type mismatches between what's registered and what's resolved only surface at runtime. Use typed tokens and validate eagerly on startup.

6. **`as const` makes arrays readonly** — `["a", "b"] as const` is `readonly ["a", "b"]`. Functions accepting `string[]` won't accept it. Use `readonly string[]` in function parameters.
