---
name: distributed-systems-patterns
description: >
  Core distributed systems patterns — load balancing, caching, sharding,
  replication, consistency models, message queues, and service discovery.
  Triggers: "load balancing", "caching strategy", "database sharding",
  "replication", "consistency model", "cap theorem", "message queue",
  "event driven", "service discovery", "circuit breaker".
  NOT for: Frontend architecture, CSS, database schema without distributed concerns.
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Distributed Systems Patterns

## Load Balancing

```
                    ┌─────────────┐
                    │   Client    │
                    └──────┬──────┘
                           │
                    ┌──────▼──────┐
                    │Load Balancer│
                    └──────┬──────┘
              ┌────────────┼────────────┐
         ┌────▼────┐  ┌────▼────┐  ┌────▼────┐
         │Server 1 │  │Server 2 │  │Server 3 │
         └─────────┘  └─────────┘  └─────────┘
```

### Algorithms

| Algorithm | Best For | How It Works |
|-----------|----------|-------------|
| Round Robin | Equal-capacity servers | Cycles through servers sequentially |
| Weighted Round Robin | Mixed-capacity servers | More requests to higher-weight servers |
| Least Connections | Varying request duration | Sends to server with fewest active connections |
| IP Hash | Session affinity | Same client IP always goes to same server |
| Random | Simple setups | Random server selection |
| Least Response Time | Latency-sensitive | Sends to fastest-responding server |

### Nginx Configuration

```nginx
upstream backend {
    # Least connections with weights
    least_conn;

    server app1:3000 weight=3;
    server app2:3000 weight=2;
    server app3:3000 weight=1;

    # Health checks
    server app4:3000 backup;           # only used if others are down
    server app5:3000 max_fails=3 fail_timeout=30s;
}

server {
    listen 80;

    location / {
        proxy_pass http://backend;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header Host $host;

        # Sticky sessions (IP hash)
        # ip_hash;
    }
}
```

## Caching Strategies

```
┌────────┐    ┌───────┐    ┌──────────┐    ┌──────────┐
│ Client │───▶│  CDN  │───▶│App Server│───▶│ Database │
│        │    │(Edge) │    │ (Redis)  │    │          │
└────────┘    └───────┘    └──────────┘    └──────────┘
  Browser       L1 Cache     L2 Cache        Source
```

### Cache-Aside (Lazy Loading)

```typescript
async function getUser(userId: string): Promise<User> {
  const cacheKey = `user:${userId}`;

  // Check cache first
  const cached = await redis.get(cacheKey);
  if (cached) return JSON.parse(cached);

  // Cache miss — load from DB
  const user = await db.user.findUnique({ where: { id: userId } });
  if (!user) throw new NotFoundError("User", userId);

  // Populate cache with TTL
  await redis.set(cacheKey, JSON.stringify(user), "EX", 300);
  return user;
}
```

### Write-Through

```typescript
async function updateUser(userId: string, data: Partial<User>): Promise<User> {
  // Update DB first
  const user = await db.user.update({ where: { id: userId }, data });

  // Then update cache
  await redis.set(`user:${userId}`, JSON.stringify(user), "EX", 300);

  return user;
}
```

### Write-Behind (Async)

```typescript
async function updateUserAsync(userId: string, data: Partial<User>): Promise<void> {
  // Update cache immediately
  const user = { ...(await getUser(userId)), ...data };
  await redis.set(`user:${userId}`, JSON.stringify(user), "EX", 300);

  // Queue DB write for later
  await queue.add("db-write", {
    table: "users",
    id: userId,
    data,
  });
}
```

### Cache Invalidation Patterns

```typescript
// Tag-based invalidation
async function invalidateByTag(tag: string): Promise<void> {
  const keys = await redis.smembers(`tag:${tag}`);
  if (keys.length > 0) {
    await redis.del(...keys);
    await redis.del(`tag:${tag}`);
  }
}

// When caching, tag the entry
async function cacheWithTags(key: string, value: any, tags: string[], ttl: number) {
  const pipeline = redis.pipeline();
  pipeline.set(key, JSON.stringify(value), "EX", ttl);
  for (const tag of tags) {
    pipeline.sadd(`tag:${tag}`, key);
    pipeline.expire(`tag:${tag}`, ttl);
  }
  await pipeline.exec();
}

// Usage
await cacheWithTags("user:123", user, ["users", `org:${user.orgId}`], 300);
await invalidateByTag(`org:${orgId}`); // invalidates all users in org
```

## Database Sharding

### Horizontal Sharding (Range-Based)

```typescript
function getShardId(userId: string): number {
  // Simple hash-based sharding
  const hash = hashCode(userId);
  return hash % NUM_SHARDS;
}

function getShardConnection(shardId: number): PrismaClient {
  const shards: Record<number, PrismaClient> = {
    0: new PrismaClient({ datasources: { db: { url: process.env.SHARD_0_URL } } }),
    1: new PrismaClient({ datasources: { db: { url: process.env.SHARD_1_URL } } }),
    2: new PrismaClient({ datasources: { db: { url: process.env.SHARD_2_URL } } }),
  };
  return shards[shardId];
}

async function getUser(userId: string): Promise<User> {
  const shardId = getShardId(userId);
  const db = getShardConnection(shardId);
  return db.user.findUnique({ where: { id: userId } });
}
```

### Consistent Hashing

```typescript
import { createHash } from "crypto";

class ConsistentHash {
  private ring: Map<number, string> = new Map();
  private sortedKeys: number[] = [];
  private virtualNodes: number;

  constructor(nodes: string[], virtualNodes = 150) {
    this.virtualNodes = virtualNodes;
    for (const node of nodes) this.addNode(node);
  }

  private hash(key: string): number {
    return parseInt(createHash("md5").update(key).digest("hex").slice(0, 8), 16);
  }

  addNode(node: string): void {
    for (let i = 0; i < this.virtualNodes; i++) {
      const hash = this.hash(`${node}:${i}`);
      this.ring.set(hash, node);
      this.sortedKeys.push(hash);
    }
    this.sortedKeys.sort((a, b) => a - b);
  }

  removeNode(node: string): void {
    for (let i = 0; i < this.virtualNodes; i++) {
      const hash = this.hash(`${node}:${i}`);
      this.ring.delete(hash);
      this.sortedKeys = this.sortedKeys.filter((k) => k !== hash);
    }
  }

  getNode(key: string): string {
    const hash = this.hash(key);
    let idx = this.sortedKeys.findIndex((k) => k >= hash);
    if (idx === -1) idx = 0; // wrap around
    return this.ring.get(this.sortedKeys[idx])!;
  }
}
```

## Replication Patterns

### Primary-Replica with Read Routing

```typescript
class DatabaseRouter {
  private primary: PrismaClient;
  private replicas: PrismaClient[];
  private currentReplica = 0;

  constructor() {
    this.primary = new PrismaClient({
      datasources: { db: { url: process.env.PRIMARY_DB_URL } },
    });
    this.replicas = [
      new PrismaClient({ datasources: { db: { url: process.env.REPLICA_1_URL } } }),
      new PrismaClient({ datasources: { db: { url: process.env.REPLICA_2_URL } } }),
    ];
  }

  // All writes go to primary
  get writer(): PrismaClient {
    return this.primary;
  }

  // Reads are round-robin across replicas
  get reader(): PrismaClient {
    const replica = this.replicas[this.currentReplica % this.replicas.length];
    this.currentReplica++;
    return replica;
  }

  // Force read from primary (when you need latest data)
  get strongReader(): PrismaClient {
    return this.primary;
  }
}

const db = new DatabaseRouter();

// Usage
const users = await db.reader.user.findMany(); // reads from replica
const user = await db.writer.user.create({ data }); // writes to primary
const fresh = await db.strongReader.user.findUnique({ where: { id } }); // reads from primary
```

## Message Queues & Event-Driven Architecture

### Bull Queue (Redis-based)

```typescript
import { Queue, Worker, QueueScheduler } from "bullmq";
import { Redis } from "ioredis";

const connection = new Redis({ maxRetriesPerRequest: null });

// Producer
const emailQueue = new Queue("email", { connection });

await emailQueue.add("welcome", {
  to: "user@example.com",
  subject: "Welcome!",
  template: "welcome",
}, {
  delay: 5000,          // delay 5s
  attempts: 3,          // retry 3 times
  backoff: {
    type: "exponential",
    delay: 1000,
  },
  removeOnComplete: 100, // keep last 100
  removeOnFail: 500,
});

// Consumer
const worker = new Worker("email", async (job) => {
  console.log(`Processing ${job.name} for ${job.data.to}`);
  await sendEmail(job.data);
}, {
  connection,
  concurrency: 5,
  limiter: {
    max: 10,
    duration: 1000, // max 10 jobs per second
  },
});

worker.on("completed", (job) => console.log(`Completed: ${job.id}`));
worker.on("failed", (job, err) => console.error(`Failed: ${job?.id}`, err));
```

### Event Bus Pattern

```typescript
import { EventEmitter } from "events";

type Events = {
  "user.created": { userId: string; email: string };
  "user.updated": { userId: string; changes: Record<string, any> };
  "order.placed": { orderId: string; userId: string; total: number };
  "order.fulfilled": { orderId: string };
};

class TypedEventBus {
  private emitter = new EventEmitter();

  emit<K extends keyof Events>(event: K, data: Events[K]): void {
    this.emitter.emit(event, data);
  }

  on<K extends keyof Events>(event: K, handler: (data: Events[K]) => void): void {
    this.emitter.on(event, handler);
  }
}

const bus = new TypedEventBus();

// Register handlers
bus.on("user.created", async ({ userId, email }) => {
  await sendWelcomeEmail(email);
  await createDefaultSettings(userId);
  await analytics.track("user_signup", { userId });
});

bus.on("order.placed", async ({ orderId, userId, total }) => {
  await updateInventory(orderId);
  await chargePayment(userId, total);
  await notifyFulfillment(orderId);
});
```

## Circuit Breaker

```typescript
enum CircuitState {
  CLOSED = "CLOSED",     // normal operation
  OPEN = "OPEN",         // failing, reject all calls
  HALF_OPEN = "HALF_OPEN", // testing if service recovered
}

class CircuitBreaker {
  private state = CircuitState.CLOSED;
  private failureCount = 0;
  private lastFailureTime = 0;
  private successCount = 0;

  constructor(
    private readonly threshold: number = 5,    // failures before opening
    private readonly timeout: number = 30000,  // ms before trying again
    private readonly successThreshold: number = 3, // successes to close
  ) {}

  async execute<T>(fn: () => Promise<T>): Promise<T> {
    if (this.state === CircuitState.OPEN) {
      if (Date.now() - this.lastFailureTime > this.timeout) {
        this.state = CircuitState.HALF_OPEN;
        this.successCount = 0;
      } else {
        throw new Error("Circuit breaker is OPEN");
      }
    }

    try {
      const result = await fn();

      if (this.state === CircuitState.HALF_OPEN) {
        this.successCount++;
        if (this.successCount >= this.successThreshold) {
          this.state = CircuitState.CLOSED;
          this.failureCount = 0;
        }
      }
      this.failureCount = 0;
      return result;
    } catch (error) {
      this.failureCount++;
      this.lastFailureTime = Date.now();

      if (this.failureCount >= this.threshold) {
        this.state = CircuitState.OPEN;
      }
      throw error;
    }
  }

  getState(): CircuitState {
    return this.state;
  }
}

// Usage
const paymentBreaker = new CircuitBreaker(5, 30000, 3);

async function processPayment(orderId: string) {
  return paymentBreaker.execute(async () => {
    return await paymentService.charge(orderId);
  });
}
```

## Service Discovery

```typescript
// Simple service registry
class ServiceRegistry {
  private services = new Map<string, ServiceInstance[]>();
  private healthChecks = new Map<string, NodeJS.Timer>();

  register(name: string, instance: ServiceInstance): void {
    const instances = this.services.get(name) ?? [];
    instances.push({ ...instance, lastHeartbeat: Date.now() });
    this.services.set(name, instances);
  }

  deregister(name: string, instanceId: string): void {
    const instances = this.services.get(name) ?? [];
    this.services.set(name, instances.filter((i) => i.id !== instanceId));
  }

  discover(name: string): ServiceInstance | null {
    const instances = (this.services.get(name) ?? [])
      .filter((i) => Date.now() - i.lastHeartbeat < 30000); // healthy only
    if (instances.length === 0) return null;
    return instances[Math.floor(Math.random() * instances.length)];
  }

  heartbeat(name: string, instanceId: string): void {
    const instances = this.services.get(name) ?? [];
    const instance = instances.find((i) => i.id === instanceId);
    if (instance) instance.lastHeartbeat = Date.now();
  }
}

interface ServiceInstance {
  id: string;
  host: string;
  port: number;
  lastHeartbeat: number;
  metadata?: Record<string, string>;
}
```

## CAP Theorem Reference

```
         Consistency
            /\
           /  \
          / CP \
         /______\
        /\      /\
       /  \ CA /  \
      / AP \  /    \
     /______\/______\
  Availability    Partition Tolerance
```

| System | Type | Trade-off |
|--------|------|-----------|
| PostgreSQL (single) | CA | No partition tolerance |
| MongoDB (replica set) | CP | Reads may block during election |
| Cassandra | AP | Eventually consistent reads |
| Redis Cluster | AP | Async replication, possible data loss |
| DynamoDB | AP or CP | Configurable per-read consistency |
| etcd / Consul | CP | Raft consensus, slower writes |

## Gotchas

1. **Replication lag is real.** Write to primary, immediately read from replica = stale data. Use "read your own writes" — route reads to primary for the writing user for a short window after writes.

2. **Cache stampede happens under load.** When a popular cache key expires, hundreds of requests simultaneously hit the database. Use probabilistic early expiration or mutex locks to serialize cache rebuilds.

3. **Message queues need dead letter queues.** Messages that fail all retries must go somewhere for manual inspection, not be silently dropped. Always configure a DLQ.

4. **Circuit breakers need monitoring.** A silently open circuit breaker means your service is degraded and nobody knows. Alert on state transitions and track error rates that trigger opens.

5. **Consistent hashing still needs rebalancing.** Adding or removing a node only moves ~1/N of the keys, but those keys need to be migrated. Without migration logic, removed node data is simply lost.

6. **Horizontal sharding makes cross-shard queries expensive.** Queries that span multiple shards require scatter-gather. Design your shard key so that 95%+ of queries hit a single shard.
