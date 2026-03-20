---
name: structural-patterns
description: >
  Structural design patterns for organizing code and object composition.
  Use when implementing adapter, facade, proxy, decorator, composite,
  bridge, or flyweight patterns in TypeScript or JavaScript.
  Triggers: "structural pattern", "adapter pattern", "facade pattern",
  "proxy pattern", "decorator pattern", "composite pattern", "bridge pattern",
  "flyweight pattern", "wrapper pattern".
  NOT for: creational patterns (see creational-patterns), behavioral patterns (see behavioral-patterns).
version: 1.0.0
allowed-tools: Read, Grep, Glob, Bash
---

# Structural Design Patterns

## Adapter Pattern

```typescript
// Unify incompatible interfaces behind a common contract

// Target interface your code expects
interface PaymentProcessor {
  charge(amount: number, currency: string, token: string): Promise<{
    id: string;
    status: 'success' | 'failed';
    amount: number;
  }>;
  refund(transactionId: string, amount?: number): Promise<{ id: string; status: string }>;
}

// Adaptee: Stripe SDK (different interface)
class StripeAdapter implements PaymentProcessor {
  constructor(private stripe: Stripe) {}

  async charge(amount: number, currency: string, token: string) {
    const intent = await this.stripe.paymentIntents.create({
      amount: Math.round(amount * 100), // Stripe uses cents
      currency: currency.toLowerCase(),
      payment_method: token,
      confirm: true,
    });
    return {
      id: intent.id,
      status: intent.status === 'succeeded' ? 'success' as const : 'failed' as const,
      amount: intent.amount / 100,
    };
  }

  async refund(transactionId: string, amount?: number) {
    const refund = await this.stripe.refunds.create({
      payment_intent: transactionId,
      ...(amount && { amount: Math.round(amount * 100) }),
    });
    return { id: refund.id, status: refund.status ?? 'unknown' };
  }
}

// Adaptee: PayPal SDK (completely different interface)
class PayPalAdapter implements PaymentProcessor {
  constructor(private paypal: PayPalClient) {}

  async charge(amount: number, currency: string, token: string) {
    const order = await this.paypal.orders.capture(token);
    const capture = order.purchase_units[0].payments.captures[0];
    return {
      id: capture.id,
      status: capture.status === 'COMPLETED' ? 'success' as const : 'failed' as const,
      amount: parseFloat(capture.amount.value),
    };
  }

  async refund(transactionId: string, amount?: number) {
    const refund = await this.paypal.payments.capturesRefund(transactionId, {
      ...(amount && { amount: { value: amount.toFixed(2), currency_code: 'USD' } }),
    });
    return { id: refund.id, status: refund.status?.toLowerCase() ?? 'unknown' };
  }
}

// Usage: swap implementations without changing business logic
function createPaymentProcessor(provider: 'stripe' | 'paypal'): PaymentProcessor {
  if (provider === 'stripe') return new StripeAdapter(new Stripe(process.env.STRIPE_KEY!));
  return new PayPalAdapter(new PayPalClient(process.env.PAYPAL_ID!, process.env.PAYPAL_SECRET!));
}
```

## Facade Pattern

```typescript
// Simplify a complex subsystem behind a clean interface

// Complex subsystems
class VideoEncoder { encode(file: Buffer, format: string): Buffer { /* ... */ return file; } }
class ThumbnailGenerator { generate(video: Buffer, timestamp: number): Buffer { /* ... */ return Buffer.alloc(0); } }
class StorageService { upload(key: string, data: Buffer): Promise<string> { /* ... */ return Promise.resolve(''); } }
class CDNService { invalidate(key: string): Promise<void> { /* ... */ } }
class MetadataDB { save(id: string, meta: Record<string, unknown>): Promise<void> { /* ... */ } }

// Facade: one method hides 5 subsystems
class VideoUploadFacade {
  private encoder = new VideoEncoder();
  private thumbnails = new ThumbnailGenerator();
  private storage = new StorageService();
  private cdn = new CDNService();
  private db = new MetadataDB();

  async uploadVideo(file: Buffer, options: {
    title: string;
    format?: string;
    thumbnailAt?: number;
  }): Promise<{ videoUrl: string; thumbnailUrl: string }> {
    // 1. Encode video
    const encoded = this.encoder.encode(file, options.format ?? 'mp4');

    // 2. Generate thumbnail
    const thumbnail = this.thumbnails.generate(encoded, options.thumbnailAt ?? 1);

    // 3. Upload both to storage
    const [videoUrl, thumbnailUrl] = await Promise.all([
      this.storage.upload(`videos/${options.title}.mp4`, encoded),
      this.storage.upload(`thumbs/${options.title}.jpg`, thumbnail),
    ]);

    // 4. Invalidate CDN cache
    await this.cdn.invalidate(`videos/${options.title}.*`);

    // 5. Save metadata
    await this.db.save(options.title, {
      videoUrl,
      thumbnailUrl,
      uploadedAt: new Date().toISOString(),
      format: options.format ?? 'mp4',
    });

    return { videoUrl, thumbnailUrl };
  }
}

// Consumer code stays clean:
// const uploader = new VideoUploadFacade();
// const result = await uploader.uploadVideo(fileBuffer, { title: 'demo-video' });
```

## Proxy Pattern

```typescript
// Control access to an object with an intermediary

// Virtual Proxy: lazy-load expensive resources
class HeavyReport {
  private data: unknown[] = [];
  constructor(private reportId: string) {
    // Expensive operation: loads millions of rows
    console.log(`Loading report ${reportId}...`);
    this.data = this.loadFromDatabase();
  }
  private loadFromDatabase(): unknown[] { return []; }
  getSummary(): string { return `Report with ${this.data.length} rows`; }
}

class LazyReportProxy {
  private report: HeavyReport | null = null;

  constructor(private reportId: string) {
    // No expensive loading in constructor
  }

  getSummary(): string {
    if (!this.report) {
      this.report = new HeavyReport(this.reportId); // Load only when needed
    }
    return this.report.getSummary();
  }
}

// Caching Proxy: avoid redundant work
class CachingApiProxy<T> {
  private cache = new Map<string, { data: T; expiresAt: number }>();

  constructor(
    private fetcher: (url: string) => Promise<T>,
    private ttlMs: number = 60_000,
  ) {}

  async fetch(url: string): Promise<T> {
    const cached = this.cache.get(url);
    if (cached && cached.expiresAt > Date.now()) {
      return cached.data;
    }

    const data = await this.fetcher(url);
    this.cache.set(url, { data, expiresAt: Date.now() + this.ttlMs });
    return data;
  }

  invalidate(url: string): void {
    this.cache.delete(url);
  }
}

// Protection Proxy: access control
class AdminOnlyProxy<T extends Record<string, (...args: unknown[]) => unknown>> {
  constructor(
    private target: T,
    private adminMethods: Set<string>,
    private isAdmin: () => boolean,
  ) {
    return new Proxy(this, {
      get: (_obj, prop: string) => {
        if (this.adminMethods.has(prop) && !this.isAdmin()) {
          throw new Error(`Access denied: ${prop} requires admin privileges`);
        }
        const value = this.target[prop];
        return typeof value === 'function' ? value.bind(this.target) : value;
      },
    }) as unknown as AdminOnlyProxy<T>;
  }
}
```

## Decorator Pattern

```typescript
// Add behavior to objects dynamically without subclassing

// Base interface
interface Logger {
  log(level: string, message: string, meta?: Record<string, unknown>): void;
}

// Concrete base
class ConsoleLogger implements Logger {
  log(level: string, message: string, meta?: Record<string, unknown>): void {
    console.log(`[${level.toUpperCase()}] ${message}`, meta ?? '');
  }
}

// Decorator: adds timestamps
class TimestampDecorator implements Logger {
  constructor(private wrapped: Logger) {}
  log(level: string, message: string, meta?: Record<string, unknown>): void {
    this.wrapped.log(level, message, {
      ...meta,
      timestamp: new Date().toISOString(),
    });
  }
}

// Decorator: adds request context
class ContextDecorator implements Logger {
  constructor(private wrapped: Logger, private context: Record<string, string>) {}
  log(level: string, message: string, meta?: Record<string, unknown>): void {
    this.wrapped.log(level, message, { ...this.context, ...meta });
  }
}

// Decorator: filters by log level
class LevelFilterDecorator implements Logger {
  private levels = { debug: 0, info: 1, warn: 2, error: 3 };
  constructor(private wrapped: Logger, private minLevel: string) {}

  log(level: string, message: string, meta?: Record<string, unknown>): void {
    const current = this.levels[level as keyof typeof this.levels] ?? 0;
    const minimum = this.levels[this.minLevel as keyof typeof this.levels] ?? 0;
    if (current >= minimum) {
      this.wrapped.log(level, message, meta);
    }
  }
}

// Compose decorators:
const logger: Logger = new LevelFilterDecorator(
  new TimestampDecorator(
    new ContextDecorator(
      new ConsoleLogger(),
      { service: 'api', version: '2.1.0' }
    )
  ),
  'info' // Filter out debug logs
);

logger.log('debug', 'Skipped');        // Filtered out
logger.log('info', 'User logged in');  // Logged with timestamp + context
```

## Composite Pattern

```typescript
// Treat individual objects and compositions uniformly

interface FileSystemNode {
  name: string;
  getSize(): number;
  print(indent?: number): string;
  find(predicate: (node: FileSystemNode) => boolean): FileSystemNode[];
}

class File implements FileSystemNode {
  constructor(public name: string, private sizeBytes: number) {}

  getSize(): number {
    return this.sizeBytes;
  }

  print(indent = 0): string {
    return `${'  '.repeat(indent)}${this.name} (${this.formatSize()})`;
  }

  find(predicate: (node: FileSystemNode) => boolean): FileSystemNode[] {
    return predicate(this) ? [this] : [];
  }

  private formatSize(): string {
    if (this.sizeBytes < 1024) return `${this.sizeBytes}B`;
    if (this.sizeBytes < 1024 * 1024) return `${(this.sizeBytes / 1024).toFixed(1)}KB`;
    return `${(this.sizeBytes / (1024 * 1024)).toFixed(1)}MB`;
  }
}

class Directory implements FileSystemNode {
  private children: FileSystemNode[] = [];

  constructor(public name: string) {}

  add(node: FileSystemNode): this {
    this.children.push(node);
    return this;
  }

  remove(name: string): void {
    this.children = this.children.filter(c => c.name !== name);
  }

  getSize(): number {
    return this.children.reduce((sum, child) => sum + child.getSize(), 0);
  }

  print(indent = 0): string {
    const lines = [`${'  '.repeat(indent)}${this.name}/`];
    for (const child of this.children) {
      lines.push(child.print(indent + 1));
    }
    return lines.join('\n');
  }

  find(predicate: (node: FileSystemNode) => boolean): FileSystemNode[] {
    const results: FileSystemNode[] = predicate(this) ? [this] : [];
    for (const child of this.children) {
      results.push(...child.find(predicate));
    }
    return results;
  }
}

// Usage:
// const root = new Directory('project')
//   .add(new File('package.json', 1200))
//   .add(new Directory('src')
//     .add(new File('index.ts', 500))
//     .add(new File('utils.ts', 800)));
// console.log(root.print());       // Prints tree
// console.log(root.getSize());     // Sums all files
// root.find(n => n.name.endsWith('.ts')); // Finds recursively
```

## Gotchas

1. **Adapter hides breaking changes** -- An adapter makes two interfaces compatible, but it can mask API changes that should propagate. If the adapted service changes its error codes, rate limits, or response semantics, the adapter might silently swallow the difference. Always log the original response alongside the adapted one during development.

2. **Facade becomes a god object** -- A facade that grows to 20+ methods is no longer simplifying — it's becoming the system. Keep facades focused on one workflow. If you need multiple simplified APIs, create multiple facades instead of one mega-facade.

3. **Decorator order affects behavior** -- `new LevelFilter(new Timestamp(base))` and `new Timestamp(new LevelFilter(base))` produce different results. The level filter might strip the timestamp, or the timestamp might be added to filtered-out messages. Order decorators outside-in: filters first, then enrichers, then the base.

4. **Proxy performance overhead** -- JavaScript `Proxy` objects add a measurable overhead to every property access and method call. For hot paths (called thousands of times per second), a Proxy-based pattern can cause noticeable performance degradation. Use class-based proxies for performance-critical code.

5. **Composite depth limits** -- Recursive `getSize()` or `find()` on deeply nested composites can cause stack overflows. Set a maximum depth limit or use iterative traversal (BFS with a queue) for composites that could be arbitrarily deep. File systems, org charts, and UI component trees are all vulnerable.

6. **Leaky adapter abstractions** -- An adapter that exposes provider-specific error types, rate limit headers, or pagination cursors from the underlying service defeats the purpose. The adapter's interface should be fully self-contained. Map provider errors to your own error types, and handle pagination internally.
