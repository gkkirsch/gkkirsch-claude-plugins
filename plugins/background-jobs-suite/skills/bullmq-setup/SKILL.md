---
name: bullmq-setup
description: >
  Set up BullMQ background job processing with Redis — queues, workers,
  retry strategies, dead letter queues, and production configuration.
  Triggers: "background jobs", "job queue", "BullMQ", "task queue",
  "async processing", "worker process", "job scheduling".
  NOT for: cron jobs only (use cron-scheduling), message brokers (RabbitMQ/Kafka).
version: 1.0.0
allowed-tools: Read, Grep, Glob, Bash, Edit, Write
---

# BullMQ Background Job Processing

## Quick Start

### 1. Install Dependencies

```bash
npm install bullmq ioredis
npm install -D @types/ioredis  # if TypeScript
```

### 2. Redis Connection

```typescript
// src/lib/redis.ts
import IORedis from 'ioredis';

// Connection for BullMQ (needs specific settings)
export const connection = new IORedis(process.env.REDIS_URL || 'redis://localhost:6379', {
  maxRetriesPerRequest: null,  // Required by BullMQ
  enableReadyCheck: false,
});

// Separate connection for app cache (don't mix with BullMQ)
export const cacheRedis = new IORedis(process.env.REDIS_URL || 'redis://localhost:6379');
```

### 3. Define a Queue

```typescript
// src/queues/email.queue.ts
import { Queue } from 'bullmq';
import { connection } from '../lib/redis';

export interface EmailJobData {
  to: string;
  subject: string;
  template: string;
  variables: Record<string, string>;
  userId?: string;
}

export const emailQueue = new Queue<EmailJobData>('email', {
  connection,
  defaultJobOptions: {
    attempts: 3,
    backoff: {
      type: 'exponential',
      delay: 1000,  // 1s, 2s, 4s
    },
    removeOnComplete: {
      count: 1000,   // keep last 1000 completed
      age: 86400,    // or 24 hours
    },
    removeOnFail: {
      count: 5000,   // keep more failures for debugging
      age: 604800,   // 7 days
    },
  },
});
```

### 4. Create a Worker

```typescript
// src/workers/email.worker.ts
import { Worker, Job } from 'bullmq';
import { connection } from '../lib/redis';
import { EmailJobData } from '../queues/email.queue';
import { sendEmail } from '../lib/email';

const emailWorker = new Worker<EmailJobData>(
  'email',
  async (job: Job<EmailJobData>) => {
    const { to, subject, template, variables } = job.data;

    // Update progress
    await job.updateProgress(10);

    // Do the work
    const result = await sendEmail({ to, subject, template, variables });

    await job.updateProgress(100);

    // Return value is stored and accessible later
    return { messageId: result.messageId, sentAt: new Date().toISOString() };
  },
  {
    connection,
    concurrency: 10,           // Process 10 emails at once
    limiter: {
      max: 100,                // Max 100 jobs
      duration: 60_000,        // per minute (respect email provider limits)
    },
  }
);

// Event handlers
emailWorker.on('completed', (job) => {
  console.log(`Email sent: ${job.id} → ${job.data.to}`);
});

emailWorker.on('failed', (job, error) => {
  console.error(`Email failed: ${job?.id} → ${error.message}`);
  if (job && job.attemptsMade >= (job.opts.attempts || 3)) {
    // All retries exhausted — alert
    alertDeadLetter('email', job.id!, error.message);
  }
});

// Graceful shutdown
async function shutdown() {
  console.log('Shutting down email worker...');
  await emailWorker.close();
  process.exit(0);
}

process.on('SIGTERM', shutdown);
process.on('SIGINT', shutdown);

export default emailWorker;
```

### 5. Add Jobs from Your API

```typescript
// src/routes/auth.ts
import { emailQueue } from '../queues/email.queue';

router.post('/signup', async (req, res) => {
  const user = await createUser(req.body);

  // Don't await the email — fire and forget
  await emailQueue.add(
    'welcome-email',           // job name (for filtering/metrics)
    {
      to: user.email,
      subject: 'Welcome!',
      template: 'welcome',
      variables: { name: user.name },
      userId: user.id,
    },
    {
      priority: 1,             // 1 = highest priority
      delay: 5000,             // Wait 5 seconds (let user land on success page)
      jobId: `welcome-${user.id}`,  // Deduplicate — same user won't get 2 welcomes
    }
  );

  res.json({ success: true });
});
```

---

## Queue Patterns

### Priority Queues

```typescript
// High-priority jobs process first
await emailQueue.add('password-reset', data, { priority: 1 });  // Urgent
await emailQueue.add('welcome', data, { priority: 5 });         // Normal
await emailQueue.add('newsletter', data, { priority: 10 });     // Low
```

### Delayed Jobs

```typescript
// Process in 30 minutes
await queue.add('reminder', data, {
  delay: 30 * 60 * 1000,
});

// Process at specific time
await queue.add('scheduled-report', data, {
  delay: targetDate.getTime() - Date.now(),
});
```

### Rate-Limited Queue

```typescript
const apiQueue = new Queue('external-api', {
  connection,
  defaultJobOptions: {
    attempts: 5,
    backoff: { type: 'fixed', delay: 60_000 },  // Retry every minute
  },
});

const apiWorker = new Worker('external-api', handler, {
  connection,
  concurrency: 1,
  limiter: {
    max: 30,
    duration: 60_000,  // 30 requests per minute
  },
});
```

### Job Dependencies (Flow)

```typescript
import { FlowProducer } from 'bullmq';

const flow = new FlowProducer({ connection });

// Parent job waits for all children to complete
await flow.add({
  name: 'generate-report',
  queueName: 'reports',
  data: { reportId: '123' },
  children: [
    {
      name: 'fetch-sales-data',
      queueName: 'data-fetching',
      data: { source: 'sales', reportId: '123' },
    },
    {
      name: 'fetch-user-data',
      queueName: 'data-fetching',
      data: { source: 'users', reportId: '123' },
    },
    {
      name: 'fetch-analytics',
      queueName: 'data-fetching',
      data: { source: 'analytics', reportId: '123' },
    },
  ],
});
```

### Batch Processing

```typescript
// Add many jobs at once (much faster than individual adds)
const jobs = users.map(user => ({
  name: 'sync-user',
  data: { userId: user.id },
  opts: { jobId: `sync-${user.id}` },  // Deduplicate
}));

await queue.addBulk(jobs);
```

### Repeatable Jobs (Cron-like)

```typescript
// Run every day at 2 AM
await queue.add('daily-cleanup', {}, {
  repeat: {
    pattern: '0 2 * * *',         // Cron syntax
    tz: 'America/New_York',
  },
});

// Run every 5 minutes
await queue.add('health-check', {}, {
  repeat: {
    every: 5 * 60 * 1000,  // Every 5 minutes
  },
});

// List all repeatable jobs
const repeatableJobs = await queue.getRepeatableJobs();
console.log(repeatableJobs);

// Remove a repeatable job
await queue.removeRepeatableByKey(repeatableJobs[0].key);
```

---

## Error Handling Best Practices

### Classify Errors for Retry Decisions

```typescript
import { Worker, Job, UnrecoverableError } from 'bullmq';

const worker = new Worker('orders', async (job: Job) => {
  try {
    await processOrder(job.data);
  } catch (error: any) {
    // Permanent failure — don't retry
    if (error.code === 'INVALID_ORDER' || error.status === 404) {
      throw new UnrecoverableError(error.message);
      // Goes directly to failed state, skips remaining retries
    }

    // Rate limited — use custom backoff
    if (error.status === 429) {
      const retryAfter = parseInt(error.headers?.['retry-after'] || '60');
      throw new DelayedError(retryAfter * 1000);
    }

    // Transient — let default retry handle it
    throw error;
  }
});
```

### Custom Backoff Strategy

```typescript
const queue = new Queue('api-calls', {
  connection,
  defaultJobOptions: {
    attempts: 6,
    backoff: {
      type: 'custom',
    },
  },
});

const worker = new Worker('api-calls', handler, {
  connection,
  settings: {
    backoffStrategy: (attemptsMade: number) => {
      // Custom: 1s, 5s, 30s, 2m, 10m, 30m
      const delays = [1000, 5000, 30000, 120000, 600000, 1800000];
      const delay = delays[attemptsMade - 1] || delays[delays.length - 1];
      // Add jitter (±25%)
      const jitter = delay * 0.25 * (Math.random() * 2 - 1);
      return Math.round(delay + jitter);
    },
  },
});
```

---

## Production Configuration

### Redis for Production

```typescript
// Production Redis config
const connection = new IORedis(process.env.REDIS_URL, {
  maxRetriesPerRequest: null,
  enableReadyCheck: false,
  // TLS for managed Redis (Heroku, AWS ElastiCache, etc.)
  tls: process.env.REDIS_TLS === 'true' ? {} : undefined,
  // Connection pool for high throughput
  lazyConnect: true,
});

// Health check
connection.on('error', (err) => {
  console.error('Redis connection error:', err.message);
});

connection.on('connect', () => {
  console.log('Redis connected');
});
```

### Worker Entrypoint

```typescript
// src/worker.ts — separate process from your API server
import './workers/email.worker';
import './workers/image.worker';
import './workers/report.worker';

console.log('Workers started');

// Keep the process alive
process.on('uncaughtException', (err) => {
  console.error('Uncaught exception in worker:', err);
  process.exit(1);
});
```

### Heroku Setup

```
# Procfile
web: node dist/server.js
worker: node dist/worker.js
```

```bash
# Scale workers independently
heroku ps:scale worker=2
heroku addons:create heroku-redis:mini
```

### Docker Compose Setup

```yaml
services:
  api:
    build: .
    command: node dist/server.js
    ports: ["3000:3000"]
    environment:
      - REDIS_URL=redis://redis:6379
    depends_on: [redis]

  worker:
    build: .
    command: node dist/worker.js
    environment:
      - REDIS_URL=redis://redis:6379
    depends_on: [redis]
    deploy:
      replicas: 3

  redis:
    image: redis:7-alpine
    ports: ["6379:6379"]
    volumes:
      - redis_data:/data
    command: redis-server --appendonly yes

volumes:
  redis_data:
```

---

## Idempotency Pattern

Every job handler MUST be idempotent — safe to run more than once.

```typescript
const worker = new Worker('payments', async (job: Job) => {
  const { orderId, amount } = job.data;

  // Check if already processed (idempotency key)
  const existing = await db.payment.findFirst({
    where: { orderId, status: 'completed' },
  });

  if (existing) {
    console.log(`Payment already processed for order ${orderId}`);
    return { alreadyProcessed: true, paymentId: existing.id };
  }

  // Process payment
  const payment = await stripe.charges.create({
    amount,
    idempotencyKey: `order-${orderId}`,  // Stripe-level idempotency
  });

  await db.payment.create({
    data: { orderId, stripeId: payment.id, status: 'completed' },
  });

  return { paymentId: payment.id };
});
```

---

## Testing Workers

```typescript
// tests/workers/email.worker.test.ts
import { Queue, Worker, Job } from 'bullmq';
import IORedis from 'ioredis';

const connection = new IORedis({ maxRetriesPerRequest: null });

describe('Email Worker', () => {
  let queue: Queue;

  beforeAll(() => {
    queue = new Queue('test-email', { connection });
  });

  afterAll(async () => {
    await queue.close();
    await connection.quit();
  });

  afterEach(async () => {
    await queue.drain();  // Remove all jobs
  });

  it('should process email job', async () => {
    const sendMock = jest.fn().mockResolvedValue({ messageId: '123' });

    const worker = new Worker('test-email', async (job) => {
      return sendMock(job.data);
    }, { connection });

    await queue.add('test', {
      to: 'test@example.com',
      subject: 'Test',
      template: 'welcome',
      variables: {},
    });

    // Wait for job to complete
    await new Promise<void>((resolve) => {
      worker.on('completed', () => {
        expect(sendMock).toHaveBeenCalledWith(
          expect.objectContaining({ to: 'test@example.com' })
        );
        resolve();
      });
    });

    await worker.close();
  });

  it('should retry on transient failure', async () => {
    let attempts = 0;
    const worker = new Worker('test-email', async () => {
      attempts++;
      if (attempts < 3) throw new Error('Connection timeout');
      return { success: true };
    }, {
      connection,
      settings: { backoffStrategy: () => 100 },  // Fast retry for tests
    });

    await queue.add('retry-test', { to: 'test@example.com' }, {
      attempts: 3,
      backoff: { type: 'custom' },
    });

    await new Promise<void>((resolve) => {
      worker.on('completed', () => {
        expect(attempts).toBe(3);
        resolve();
      });
    });

    await worker.close();
  });
});
```

---

## Common Gotchas

1. **`maxRetriesPerRequest: null`** — BullMQ requires this on the Redis connection. Without it, you get cryptic errors.

2. **Don't share Redis connections** — BullMQ needs its own connection. Create separate `IORedis` instances for your app cache and BullMQ.

3. **Job data must be JSON-serializable** — No functions, no circular refs, no `Date` objects (use ISO strings), no `Buffer` (use base64 strings).

4. **Workers are separate processes** — Don't import workers in your API server. Run them as a separate `node dist/worker.js` process.

5. **Stalled jobs** — If a worker crashes mid-job, the job becomes "stalled." BullMQ auto-retries stalled jobs by default. Set `stalledInterval` to control detection frequency.

6. **Queue names are global** — Same queue name = same queue, even across different files. Use descriptive, namespaced names.

7. **Repeatable job deduplication** — Adding a repeatable job with the same pattern won't create duplicates. But changing the pattern creates a NEW repeatable job without removing the old one. Always clean up old repeatables.

8. **Memory leaks on workers** — Close workers on shutdown. Open workers hold Redis connections. Use `worker.close()` in your shutdown handler.
