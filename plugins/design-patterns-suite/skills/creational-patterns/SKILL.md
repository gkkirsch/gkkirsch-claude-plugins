---
name: creational-patterns
description: >
  Creational design patterns in TypeScript — Factory Method, Abstract Factory,
  Builder, Singleton, and dependency injection patterns with practical examples.
  Triggers: "factory pattern", "builder pattern", "singleton", "dependency injection",
  "creational pattern", "object creation", "DI container".
  NOT for: Behavioral patterns (use behavioral-patterns).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Creational Patterns

## Factory Method

Use when you need to create objects without specifying the exact class.

```typescript
// Define the product interface
interface Notification {
  send(to: string, message: string): Promise<void>;
}

// Concrete implementations
class EmailNotification implements Notification {
  constructor(private readonly smtpClient: SmtpClient) {}

  async send(to: string, message: string): Promise<void> {
    await this.smtpClient.send({ to, subject: "Notification", body: message });
  }
}

class SmsNotification implements Notification {
  constructor(private readonly twilioClient: TwilioClient) {}

  async send(to: string, message: string): Promise<void> {
    await this.twilioClient.messages.create({ to, body: message });
  }
}

class SlackNotification implements Notification {
  constructor(private readonly webhookUrl: string) {}

  async send(to: string, message: string): Promise<void> {
    await fetch(this.webhookUrl, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ channel: to, text: message }),
    });
  }
}

// Factory
type NotificationType = "email" | "sms" | "slack";

class NotificationFactory {
  constructor(
    private readonly smtpClient: SmtpClient,
    private readonly twilioClient: TwilioClient,
    private readonly slackWebhookUrl: string,
  ) {}

  create(type: NotificationType): Notification {
    switch (type) {
      case "email": return new EmailNotification(this.smtpClient);
      case "sms": return new SmsNotification(this.twilioClient);
      case "slack": return new SlackNotification(this.slackWebhookUrl);
      default: throw new Error(`Unknown notification type: ${type satisfies never}`);
    }
  }
}

// Usage
const factory = new NotificationFactory(smtp, twilio, webhookUrl);
const notifier = factory.create(user.preferredChannel);
await notifier.send(user.contact, "Your order shipped!");
```

### Registry-Based Factory (Open/Closed Principle)

```typescript
// Open for extension — add new types without modifying factory
class NotificationRegistry {
  private readonly creators = new Map<string, () => Notification>();

  register(type: string, creator: () => Notification): void {
    this.creators.set(type, creator);
  }

  create(type: string): Notification {
    const creator = this.creators.get(type);
    if (!creator) throw new Error(`Unknown type: ${type}`);
    return creator();
  }
}

// Registration (at startup)
const registry = new NotificationRegistry();
registry.register("email", () => new EmailNotification(smtp));
registry.register("sms", () => new SmsNotification(twilio));
// Add new types without touching the registry class:
registry.register("discord", () => new DiscordNotification(discordBot));
```

## Builder

Use when creating objects with many optional parameters.

```typescript
interface QueryOptions {
  table: string;
  columns: string[];
  where: Array<{ field: string; op: string; value: unknown }>;
  orderBy: Array<{ field: string; direction: "ASC" | "DESC" }>;
  limit?: number;
  offset?: number;
  joins: Array<{ table: string; on: string; type: "INNER" | "LEFT" | "RIGHT" }>;
}

class QueryBuilder {
  private options: QueryOptions;

  constructor(table: string) {
    this.options = {
      table,
      columns: ["*"],
      where: [],
      orderBy: [],
      joins: [],
    };
  }

  select(...columns: string[]): this {
    this.options.columns = columns;
    return this;
  }

  where(field: string, op: string, value: unknown): this {
    this.options.where.push({ field, op, value });
    return this;
  }

  orderBy(field: string, direction: "ASC" | "DESC" = "ASC"): this {
    this.options.orderBy.push({ field, direction });
    return this;
  }

  limit(n: number): this {
    this.options.limit = n;
    return this;
  }

  offset(n: number): this {
    this.options.offset = n;
    return this;
  }

  join(table: string, on: string, type: "INNER" | "LEFT" | "RIGHT" = "INNER"): this {
    this.options.joins.push({ table, on, type });
    return this;
  }

  build(): { sql: string; params: unknown[] } {
    const params: unknown[] = [];
    let sql = `SELECT ${this.options.columns.join(", ")} FROM ${this.options.table}`;

    for (const j of this.options.joins) {
      sql += ` ${j.type} JOIN ${j.table} ON ${j.on}`;
    }

    if (this.options.where.length > 0) {
      const clauses = this.options.where.map((w) => {
        params.push(w.value);
        return `${w.field} ${w.op} $${params.length}`;
      });
      sql += ` WHERE ${clauses.join(" AND ")}`;
    }

    if (this.options.orderBy.length > 0) {
      sql += ` ORDER BY ${this.options.orderBy.map((o) => `${o.field} ${o.direction}`).join(", ")}`;
    }

    if (this.options.limit !== undefined) sql += ` LIMIT ${this.options.limit}`;
    if (this.options.offset !== undefined) sql += ` OFFSET ${this.options.offset}`;

    return { sql, params };
  }
}

// Usage
const query = new QueryBuilder("users")
  .select("id", "name", "email")
  .join("orders", "orders.user_id = users.id", "LEFT")
  .where("status", "=", "active")
  .where("created_at", ">", "2024-01-01")
  .orderBy("name", "ASC")
  .limit(20)
  .offset(40)
  .build();
```

### Immutable Builder (Functional Style)

```typescript
interface HttpRequest {
  readonly url: string;
  readonly method: "GET" | "POST" | "PUT" | "DELETE";
  readonly headers: Readonly<Record<string, string>>;
  readonly body?: unknown;
  readonly timeout: number;
}

function createRequest(url: string): HttpRequest {
  return { url, method: "GET", headers: {}, timeout: 30_000 };
}

function withMethod(req: HttpRequest, method: HttpRequest["method"]): HttpRequest {
  return { ...req, method };
}

function withHeader(req: HttpRequest, key: string, value: string): HttpRequest {
  return { ...req, headers: { ...req.headers, [key]: value } };
}

function withBody(req: HttpRequest, body: unknown): HttpRequest {
  return { ...req, body, method: req.method === "GET" ? "POST" : req.method };
}

function withTimeout(req: HttpRequest, ms: number): HttpRequest {
  return { ...req, timeout: ms };
}

// Usage — pipe-style
const req = withTimeout(
  withHeader(
    withBody(
      withMethod(createRequest("/api/users"), "POST"),
      { name: "Alice" }
    ),
    "Authorization", "Bearer token"
  ),
  5000
);
```

## Singleton

Use sparingly — only for truly global, stateless resources.

```typescript
// Modern TypeScript Singleton (module-scoped)
class DatabasePool {
  private static instance: DatabasePool | null = null;
  private readonly pool: Pool;

  private constructor(connectionString: string) {
    this.pool = new Pool({ connectionString, max: 20 });
  }

  static getInstance(connectionString?: string): DatabasePool {
    if (!DatabasePool.instance) {
      if (!connectionString) throw new Error("Connection string required for first init");
      DatabasePool.instance = new DatabasePool(connectionString);
    }
    return DatabasePool.instance;
  }

  async query<T>(sql: string, params?: unknown[]): Promise<T[]> {
    const client = await this.pool.connect();
    try {
      const result = await client.query(sql, params);
      return result.rows;
    } finally {
      client.release();
    }
  }

  // For testing — reset the singleton
  static resetForTesting(): void {
    DatabasePool.instance = null;
  }
}
```

### Better: Module-Level Singleton (No Class Needed)

```typescript
// db.ts — module is a natural singleton in JS
import { Pool } from "pg";

const pool = new Pool({
  connectionString: process.env.DATABASE_URL,
  max: 20,
});

export async function query<T>(sql: string, params?: unknown[]): Promise<T[]> {
  const result = await pool.query(sql, params);
  return result.rows;
}

export async function getClient() {
  return pool.connect();
}

// Usage: import { query } from "./db";
// Modules are cached — same pool instance everywhere
```

## Dependency Injection

### Manual DI (No Framework)

```typescript
// Interfaces
interface UserRepository {
  findById(id: string): Promise<User | null>;
  save(user: User): Promise<void>;
}

interface EmailService {
  send(to: string, subject: string, body: string): Promise<void>;
}

interface Logger {
  info(message: string, meta?: Record<string, unknown>): void;
  error(message: string, error?: Error): void;
}

// Service depends on abstractions
class UserService {
  constructor(
    private readonly userRepo: UserRepository,
    private readonly emailService: EmailService,
    private readonly logger: Logger,
  ) {}

  async registerUser(data: CreateUserDto): Promise<User> {
    this.logger.info("Registering user", { email: data.email });

    const user = User.create(data);
    await this.userRepo.save(user);
    await this.emailService.send(user.email, "Welcome!", "Thanks for signing up.");

    this.logger.info("User registered", { userId: user.id });
    return user;
  }
}

// Composition root — wire everything together at startup
function createApp() {
  // Concrete implementations
  const logger = new PinoLogger();
  const userRepo = new PostgresUserRepository(pool);
  const emailService = new SendgridEmailService(apiKey);

  // Compose services
  const userService = new UserService(userRepo, emailService, logger);
  const authService = new AuthService(userRepo, logger);

  // Create HTTP handlers
  const userController = new UserController(userService);
  const authController = new AuthController(authService);

  return { userController, authController };
}

// Testing — inject mocks
const mockRepo: UserRepository = {
  findById: vi.fn().mockResolvedValue(testUser),
  save: vi.fn().mockResolvedValue(undefined),
};
const mockEmail: EmailService = { send: vi.fn().mockResolvedValue(undefined) };
const mockLogger: Logger = { info: vi.fn(), error: vi.fn() };

const service = new UserService(mockRepo, mockEmail, mockLogger);
```

### DI Container (Lightweight)

```typescript
// Simple typed DI container
type Factory<T> = () => T;

class Container {
  private factories = new Map<string, Factory<unknown>>();
  private singletons = new Map<string, unknown>();

  register<T>(key: string, factory: Factory<T>): void {
    this.factories.set(key, factory);
  }

  singleton<T>(key: string, factory: Factory<T>): void {
    this.factories.set(key, () => {
      if (!this.singletons.has(key)) {
        this.singletons.set(key, factory());
      }
      return this.singletons.get(key)!;
    });
  }

  resolve<T>(key: string): T {
    const factory = this.factories.get(key);
    if (!factory) throw new Error(`No registration for: ${key}`);
    return factory() as T;
  }
}

// Usage
const container = new Container();
container.singleton("logger", () => new PinoLogger());
container.singleton("db", () => new PostgresPool(process.env.DATABASE_URL!));
container.register("userRepo", () =>
  new PostgresUserRepository(container.resolve("db"))
);
container.register("userService", () =>
  new UserService(
    container.resolve("userRepo"),
    container.resolve("emailService"),
    container.resolve("logger"),
  )
);
```

## Gotchas

1. **Singleton + testing = pain** — Singletons carry state between tests. Either provide a `resetForTesting()` method, or better, use dependency injection so tests can inject their own instances. Module-level singletons can be mocked with `vi.mock()`.

2. **Builder without validation** — A Builder that doesn't validate required fields lets you build invalid objects. Either throw in `build()` if required fields are missing, or use TypeScript's type system to enforce them (generic builder pattern).

3. **Factory returning `any`** — The factory's return type should be an interface, not `any`. The whole point is that consumers don't know the concrete type. If they do, you don't need a factory.

4. **DI constructor explosion** — If a class takes 8+ dependencies via constructor, it's doing too much. Split the class. The number of constructor parameters is a proxy for the number of responsibilities.

5. **Abstract Factory without real need** — Abstract Factory (factory of factories) is rarely needed in TypeScript. Usually a simple Factory Method or a Map-based registry is sufficient. Don't add abstraction layers you won't use.

6. **Immutable builders that allocate excessively** — Each chaining call creates a new object. For hot paths, use a mutable builder class. Immutable builders are for configuration (cold path), not data transformation (hot path).
