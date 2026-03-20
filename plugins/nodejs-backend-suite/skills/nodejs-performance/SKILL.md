---
name: nodejs-performance
description: >
  Node.js performance optimization — clustering, worker threads, caching,
  connection pooling, streaming, memory management, and profiling.
  Triggers: "node performance", "node clustering", "worker threads",
  "node caching", "node memory", "node profiling", "node streaming".
  NOT for: Frontend performance (use frontend-performance).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Node.js Performance

## Clustering

```typescript
// cluster.ts — Use all CPU cores
import cluster from "cluster";
import os from "os";
import { logger } from "./utils/logger";

const numCPUs = parseInt(process.env.WEB_CONCURRENCY || "") || os.cpus().length;

if (cluster.isPrimary) {
  logger.info(`Primary ${process.pid} starting ${numCPUs} workers`);

  for (let i = 0; i < numCPUs; i++) {
    cluster.fork();
  }

  cluster.on("exit", (worker, code) => {
    logger.warn(`Worker ${worker.process.pid} exited (code: ${code}). Restarting...`);
    cluster.fork(); // Auto-restart crashed workers
  });

  // Graceful shutdown
  process.on("SIGTERM", () => {
    logger.info("SIGTERM — shutting down workers");
    for (const id in cluster.workers) {
      cluster.workers[id]?.process.kill("SIGTERM");
    }
  });
} else {
  // Worker process — start the Express app
  import("./index").catch((err) => {
    logger.error("Worker failed to start:", err);
    process.exit(1);
  });
}
```

## Worker Threads

```typescript
// Heavy computation off the main thread
import { Worker, isMainThread, parentPort, workerData } from "worker_threads";
import path from "path";

// Main thread — dispatch work
export function runInWorker<T>(taskFile: string, data: unknown): Promise<T> {
  return new Promise((resolve, reject) => {
    const worker = new Worker(taskFile, { workerData: data });
    worker.on("message", resolve);
    worker.on("error", reject);
    worker.on("exit", (code) => {
      if (code !== 0) reject(new Error(`Worker exited with code ${code}`));
    });
  });
}

// Usage
const result = await runInWorker<ReportResult>(
  path.resolve(__dirname, "workers/generate-report.js"),
  { startDate: "2026-01-01", endDate: "2026-03-01" }
);

// Worker file — workers/generate-report.ts
if (!isMainThread && parentPort) {
  const { startDate, endDate } = workerData;
  // CPU-intensive work here
  const report = generateReport(startDate, endDate);
  parentPort.postMessage(report);
}
```

### Worker Pool

```typescript
import { Worker } from "worker_threads";
import os from "os";

class WorkerPool {
  private workers: Worker[] = [];
  private queue: Array<{ data: unknown; resolve: (v: any) => void; reject: (e: Error) => void }> = [];
  private available: Worker[] = [];

  constructor(private taskFile: string, private size = os.cpus().length) {
    for (let i = 0; i < size; i++) {
      const worker = new Worker(taskFile);
      worker.on("message", (result) => {
        const next = this.queue.shift();
        if (next) {
          worker.postMessage(next.data);
          worker.once("message", next.resolve);
          worker.once("error", next.reject);
        } else {
          this.available.push(worker);
        }
      });
      this.available.push(worker);
    }
  }

  exec<T>(data: unknown): Promise<T> {
    return new Promise((resolve, reject) => {
      const worker = this.available.pop();
      if (worker) {
        worker.postMessage(data);
        worker.once("message", resolve);
        worker.once("error", reject);
      } else {
        this.queue.push({ data, resolve, reject });
      }
    });
  }

  async destroy() {
    await Promise.all(this.workers.map((w) => w.terminate()));
  }
}

// Reuse across requests
const imagePool = new WorkerPool("./workers/resize-image.js", 4);
const resized = await imagePool.exec<Buffer>({ path: "/uploads/photo.jpg", width: 800 });
```

## Caching

### In-Memory Cache (LRU)

```typescript
import { LRUCache } from "lru-cache";

const cache = new LRUCache<string, any>({
  max: 500,                    // Max entries
  ttl: 5 * 60 * 1000,         // 5 min TTL
  allowStale: false,
  updateAgeOnGet: true,        // Reset TTL on access
  sizeCalculation: (value) => JSON.stringify(value).length,
  maxSize: 50 * 1024 * 1024,  // 50MB max total size
});

// Cache middleware
function cacheMiddleware(ttl = 300) {
  return (req: Request, res: Response, next: NextFunction) => {
    if (req.method !== "GET") return next();

    const key = `cache:${req.originalUrl}`;
    const cached = cache.get(key);
    if (cached) {
      res.setHeader("X-Cache", "HIT");
      return res.json(cached);
    }

    const originalJson = res.json.bind(res);
    res.json = (body: any) => {
      cache.set(key, body, { ttl: ttl * 1000 });
      res.setHeader("X-Cache", "MISS");
      return originalJson(body);
    };
    next();
  };
}

// Invalidate on mutations
function invalidateCache(pattern: string) {
  for (const key of cache.keys()) {
    if (key.startsWith(pattern)) cache.delete(key);
  }
}
```

### Redis Cache

```typescript
import Redis from "ioredis";

const redis = new Redis(process.env.REDIS_URL!, {
  maxRetriesPerRequest: 3,
  retryDelayOnFailover: 100,
  lazyConnect: true,
});

async function cacheGet<T>(key: string): Promise<T | null> {
  const data = await redis.get(key);
  return data ? JSON.parse(data) : null;
}

async function cacheSet(key: string, data: unknown, ttlSeconds = 300): Promise<void> {
  await redis.set(key, JSON.stringify(data), "EX", ttlSeconds);
}

async function cacheDel(pattern: string): Promise<void> {
  const keys = await redis.keys(pattern);
  if (keys.length > 0) await redis.del(...keys);
}

// Cache-aside pattern
async function getUser(id: string): Promise<User> {
  const cacheKey = `user:${id}`;
  const cached = await cacheGet<User>(cacheKey);
  if (cached) return cached;

  const user = await db.user.findUnique({ where: { id } });
  if (!user) throw AppError.notFound("User");

  await cacheSet(cacheKey, user, 600); // 10 min
  return user;
}
```

## Connection Pooling

```typescript
// PostgreSQL with pg-pool
import { Pool } from "pg";

const pool = new Pool({
  connectionString: process.env.DATABASE_URL,
  max: 20,              // Max connections in pool
  idleTimeoutMillis: 30_000,
  connectionTimeoutMillis: 5_000,
  allowExitOnIdle: true,
});

// Monitor pool health
pool.on("error", (err) => logger.error("Pool error:", err));
pool.on("connect", () => logger.debug("New pool connection"));

// Use pool directly (auto release)
const { rows } = await pool.query("SELECT * FROM users WHERE id = $1", [userId]);

// Or checkout for transactions
const client = await pool.connect();
try {
  await client.query("BEGIN");
  await client.query("UPDATE accounts SET balance = balance - $1 WHERE id = $2", [amount, fromId]);
  await client.query("UPDATE accounts SET balance = balance + $1 WHERE id = $2", [amount, toId]);
  await client.query("COMMIT");
} catch (err) {
  await client.query("ROLLBACK");
  throw err;
} finally {
  client.release(); // Always release back to pool
}
```

## Streaming

```typescript
import { pipeline } from "stream/promises";
import { createReadStream, createWriteStream } from "fs";
import { Transform } from "stream";
import { createGzip } from "zlib";

// Stream large file download
app.get("/export/users", requireAuth, async (req, res) => {
  res.setHeader("Content-Type", "application/json");
  res.setHeader("Content-Disposition", "attachment; filename=users.json");

  const cursor = db.user.findMany({ cursor: true, batchSize: 100 });

  res.write("[");
  let first = true;
  for await (const user of cursor) {
    if (!first) res.write(",");
    res.write(JSON.stringify(user));
    first = false;
  }
  res.write("]");
  res.end();
});

// Stream file processing with backpressure
async function processLargeCSV(inputPath: string, outputPath: string) {
  const transform = new Transform({
    transform(chunk, _encoding, callback) {
      const lines = chunk.toString().split("\n");
      const processed = lines
        .filter((line: string) => line.trim())
        .map((line: string) => transformRow(line))
        .join("\n");
      callback(null, processed + "\n");
    },
  });

  await pipeline(
    createReadStream(inputPath),
    transform,
    createGzip(),
    createWriteStream(outputPath)
  );
}

// Stream JSON parsing for huge files
import { parser } from "stream-json";
import { streamArray } from "stream-json/streamers/StreamArray";

async function importLargeJSON(filePath: string) {
  const pipeline = createReadStream(filePath).pipe(parser()).pipe(streamArray());

  let batch: any[] = [];
  for await (const { value } of pipeline) {
    batch.push(value);
    if (batch.length >= 1000) {
      await db.record.createMany({ data: batch });
      batch = [];
    }
  }
  if (batch.length > 0) {
    await db.record.createMany({ data: batch });
  }
}
```

## Memory Management

```typescript
// Monitor memory usage
function logMemory() {
  const { heapUsed, heapTotal, rss, external } = process.memoryUsage();
  logger.info("Memory:", {
    heapUsed: `${(heapUsed / 1024 / 1024).toFixed(1)}MB`,
    heapTotal: `${(heapTotal / 1024 / 1024).toFixed(1)}MB`,
    rss: `${(rss / 1024 / 1024).toFixed(1)}MB`,
    external: `${(external / 1024 / 1024).toFixed(1)}MB`,
  });
}

// Periodic check
setInterval(logMemory, 60_000);

// Memory leak detection in development
if (process.env.NODE_ENV !== "production") {
  let lastHeap = 0;
  setInterval(() => {
    const { heapUsed } = process.memoryUsage();
    const growth = heapUsed - lastHeap;
    if (growth > 10 * 1024 * 1024) { // 10MB growth in 30s
      logger.warn(`Heap grew ${(growth / 1024 / 1024).toFixed(1)}MB in 30s — possible leak`);
    }
    lastHeap = heapUsed;
  }, 30_000);
}

// WeakRef for expensive cached objects
const expensiveCache = new Map<string, WeakRef<object>>();
const registry = new FinalizationRegistry<string>((key) => expensiveCache.delete(key));

function getCachedExpensive(key: string): object | undefined {
  const ref = expensiveCache.get(key);
  return ref?.deref();
}

function setCachedExpensive(key: string, value: object) {
  expensiveCache.set(key, new WeakRef(value));
  registry.register(value, key);
}
```

## Profiling

```bash
# CPU profiling with built-in inspector
node --inspect src/index.js
# Open chrome://inspect in Chrome → click "inspect" → Profiler tab

# Generate CPU profile programmatically
node --prof src/index.js
# After load test, process the log:
node --prof-process isolate-*.log > profile.txt

# Heap snapshot
node --inspect src/index.js
# Chrome DevTools → Memory tab → Take heap snapshot

# Clinic.js — automated profiling
npx clinic doctor -- node src/index.js
npx clinic flame -- node src/index.js
npx clinic bubbleprof -- node src/index.js

# 0x — flamegraph generator
npx 0x src/index.js

# autocannon — load testing
npx autocannon -c 100 -d 10 http://localhost:3000/api/users
```

### Inline Profiling

```typescript
// Measure specific operations
const { performance, PerformanceObserver } = require("perf_hooks");

const obs = new PerformanceObserver((list) => {
  for (const entry of list.getEntries()) {
    logger.info(`${entry.name}: ${entry.duration.toFixed(2)}ms`);
  }
});
obs.observe({ entryTypes: ["measure"] });

async function profiledOperation() {
  performance.mark("db-query-start");
  const result = await db.query("SELECT ...");
  performance.mark("db-query-end");
  performance.measure("db-query", "db-query-start", "db-query-end");
  return result;
}
```

## Performance Checklist

| Area | Optimization | Impact |
|------|-------------|--------|
| **CPU** | Cluster mode for multi-core | High |
| **CPU** | Worker threads for heavy computation | High |
| **I/O** | Connection pooling (DB, Redis) | High |
| **I/O** | Streaming for large data | High |
| **Memory** | LRU cache with size limits | Medium |
| **Memory** | Stream processing instead of loading all into memory | High |
| **Network** | Response compression (gzip/brotli) | Medium |
| **Network** | HTTP/2 with multiplexing | Medium |
| **Network** | Redis cache for hot data | High |
| **DB** | Query optimization, indexes | High |
| **DB** | Batch operations (createMany) | Medium |
| **DB** | Read replicas for read-heavy workloads | High |
| **Startup** | Lazy imports for cold start | Medium |
| **Monitoring** | APM (Datadog, New Relic, OpenTelemetry) | Essential |

## Gotchas

1. **Don't block the event loop** — `JSON.parse()` on a 100MB string, `crypto.pbkdf2Sync`, or `fs.readFileSync` in request handlers block ALL concurrent requests. Use async versions or worker threads.

2. **Connection pool exhaustion** — If max pool size is 20 and you have 25 concurrent queries, 5 will wait. Monitor `pool.waitingCount` and `pool.totalCount`. Set `connectionTimeoutMillis` to fail fast rather than queue indefinitely.

3. **Memory leaks from event listeners** — `emitter.on()` in request handlers without cleanup leaks. Use `emitter.once()` or store the listener and call `removeListener` when done. Node warns at 11 listeners per event.

4. **Streaming backpressure** — If you pipe a fast readable into a slow writable, Node buffers everything in memory. Use `pipeline()` from `stream/promises` which handles backpressure and error propagation automatically.

5. **Cache stampede** — When a popular cache key expires, 100 requests hit the DB simultaneously. Use a lock (Redis `SET NX`) so only one request rebuilds the cache. Others wait or return stale data.

6. **Cluster mode + WebSockets** — WebSocket connections are per-worker. Client reconnects may hit a different worker. Use Redis pub/sub or sticky sessions to route WebSocket connections to the correct worker.
