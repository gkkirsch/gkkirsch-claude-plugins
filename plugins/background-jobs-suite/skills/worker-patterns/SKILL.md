---
name: worker-patterns
description: >
  Implement robust worker processes with error handling, graceful shutdown,
  concurrency control, health checks, and monitoring.
  Triggers: "worker process", "job processor", "background worker", "consumer",
  "graceful shutdown", "error handling for jobs", "dead letter queue".
  NOT for: queue setup (use bullmq-setup), scheduling (use cron-scheduling).
version: 1.0.0
argument-hint: "[pattern-name]"
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Worker Implementation Patterns

Production-ready patterns for background job processing with proper error handling, monitoring, and deployment.

## Production Worker Template

Complete worker template with all production concerns:

```typescript
// src/workers/base.worker.ts
import { Worker, Job, MetricsTime } from 'bullmq';
import { connection } from '../lib/redis';

interface WorkerConfig {
  queueName: string;
  concurrency?: number;
  limiter?: { max: number; duration: number };
  lockDuration?: number;
}

export function createWorker(
  config: WorkerConfig,
  processor: (job: Job) => Promise<any>
) {
  const worker = new Worker(
    config.queueName,
    async (job: Job) => {
      const start = Date.now();
      const logPrefix = `[${config.queueName}:${job.id}]`;

      try {
        console.log(`${logPrefix} Starting ${job.name}`, {
          attempt: job.attemptsMade + 1,
          data: job.data,
        });

        const result = await processor(job);

        const duration = Date.now() - start;
        console.log(`${logPrefix} Completed in ${duration}ms`);

        return result;
      } catch (error) {
        const duration = Date.now() - start;
        console.error(`${logPrefix} Failed after ${duration}ms:`, error);
        throw error; // Re-throw for BullMQ retry handling
      }
    },
    {
      connection,
      concurrency: config.concurrency ?? 5,
      limiter: config.limiter,
      lockDuration: config.lockDuration ?? 30_000,
      metrics: { maxDataPoints: MetricsTime.ONE_WEEK },
    }
  );

  // Event handlers
  worker.on('completed', (job) => {
    console.log(`[${config.queueName}] Job ${job.id} completed`);
  });

  worker.on('failed', (job, err) => {
    console.error(`[${config.queueName}] Job ${job?.id} failed:`, err.message);
    if (job && job.attemptsMade >= (job.opts.attempts ?? 1)) {
      console.error(`[${config.queueName}] Job ${job.id} exhausted all retries — moving to DLQ`);
      // Alert on final failure
      alertOnFailure(config.queueName, job, err);
    }
  });

  worker.on('error', (err) => {
    console.error(`[${config.queueName}] Worker error:`, err);
  });

  worker.on('stalled', (jobId) => {
    console.warn(`[${config.queueName}] Job ${jobId} stalled — will be reprocessed`);
  });

  return worker;
}
```

## Graceful Shutdown

Every worker MUST handle graceful shutdown. Without it, jobs get lost or duplicated.

```typescript
// src/workers/shutdown.ts
const workers: Worker[] = [];

export function registerWorker(worker: Worker) {
  workers.push(worker);
}

let isShuttingDown = false;

async function gracefulShutdown(signal: string) {
  if (isShuttingDown) return;
  isShuttingDown = true;

  console.log(`${signal} received. Starting graceful shutdown...`);

  // Phase 1: Stop accepting new jobs
  console.log('Closing workers (waiting for current jobs)...');

  const shutdownPromises = workers.map(async (worker) => {
    try {
      // close() waits for current jobs to finish
      await worker.close();
      console.log(`Worker ${worker.name} closed`);
    } catch (err) {
      console.error(`Error closing worker ${worker.name}:`, err);
    }
  });

  // Phase 2: Wait with timeout
  const timeout = setTimeout(() => {
    console.error('Shutdown timeout exceeded, forcing exit');
    process.exit(1);
  }, 30_000); // 30 second timeout

  await Promise.all(shutdownPromises);
  clearTimeout(timeout);

  // Phase 3: Close connections
  console.log('All workers closed. Cleaning up connections...');
  await connection.quit();

  console.log('Graceful shutdown complete');
  process.exit(0);
}

process.on('SIGTERM', () => gracefulShutdown('SIGTERM'));
process.on('SIGINT', () => gracefulShutdown('SIGINT'));
```

## Error Classification & Retry

Not all errors should be retried. Classify them to avoid wasting resources:

```typescript
// src/lib/errors.ts

// Custom error classes for classification
export class TransientError extends Error {
  constructor(message: string) {
    super(message);
    this.name = 'TransientError';
  }
}

export class PermanentError extends Error {
  constructor(message: string) {
    super(message);
    this.name = 'PermanentError';
  }
}

export function classifyAndWrap(error: unknown): Error {
  if (error instanceof PermanentError) return error;
  if (error instanceof TransientError) return error;

  const err = error instanceof Error ? error : new Error(String(error));
  const msg = err.message.toLowerCase();

  // Transient — retry these
  if (
    msg.includes('econnrefused') ||
    msg.includes('econnreset') ||
    msg.includes('etimedout') ||
    msg.includes('timeout') ||
    msg.includes('429') ||
    msg.includes('too many requests') ||
    msg.includes('503') ||
    msg.includes('service unavailable') ||
    msg.includes('502') ||
    msg.includes('bad gateway')
  ) {
    return new TransientError(err.message);
  }

  // Permanent — never retry
  if (
    msg.includes('400') ||
    msg.includes('401') ||
    msg.includes('403') ||
    msg.includes('404') ||
    msg.includes('not found') ||
    msg.includes('invalid') ||
    msg.includes('validation') ||
    msg.includes('unauthorized')
  ) {
    return new PermanentError(err.message);
  }

  // Unknown errors are bugs — don't retry, investigate
  return new PermanentError(`[BUG] ${err.message}`);
}

// Use in job processor
async function processJob(job: Job) {
  try {
    await doWork(job.data);
  } catch (error) {
    const classified = classifyAndWrap(error);
    if (classified instanceof PermanentError) {
      // Skip retries — move to DLQ immediately
      console.error(`Permanent error, skipping retries:`, classified.message);
      throw new Error(`PERMANENT: ${classified.message}`);
      // BullMQ will exhaust attempts and move to failed
    }
    throw classified; // Transient — let BullMQ retry
  }
}
```

## Dead Letter Queue (DLQ)

Handle jobs that exhaust all retries:

```typescript
// src/queues/dlq.ts
import { Queue, QueueEvents } from 'bullmq';
import { connection } from '../lib/redis';

const dlqQueue = new Queue('dead-letter', { connection });

export function setupDLQ(sourceQueueName: string) {
  const events = new QueueEvents(sourceQueueName, { connection });

  events.on('failed', async ({ jobId, failedReason }) => {
    // Get the original job
    const sourceQueue = new Queue(sourceQueueName, { connection });
    const job = await sourceQueue.getJob(jobId);

    if (!job) return;

    // Only move to DLQ when all retries exhausted
    if (job.attemptsMade < (job.opts.attempts ?? 1)) return;

    // Add to DLQ with original context
    await dlqQueue.add('failed-job', {
      originalQueue: sourceQueueName,
      originalJobId: jobId,
      originalName: job.name,
      originalData: job.data,
      failedReason,
      failedAt: new Date().toISOString(),
      attempts: job.attemptsMade,
    }, {
      removeOnComplete: false, // Keep DLQ jobs for review
    });

    console.error(`[DLQ] Job ${jobId} from ${sourceQueueName} moved to DLQ: ${failedReason}`);
  });
}
```

## Idempotent Job Handlers

Jobs may run more than once (worker crash during processing, stalled job recovery). Design for it:

```typescript
// BAD: Not idempotent — charges twice if worker crashes after charge but before marking complete
async function processPayment(job: Job) {
  await chargeCustomer(job.data.customerId, job.data.amount);
  await markOrderPaid(job.data.orderId);
}

// GOOD: Idempotent — safe to run multiple times
async function processPayment(job: Job) {
  const order = await getOrder(job.data.orderId);

  // Guard: already processed
  if (order.status === 'paid') {
    console.log(`Order ${order.id} already paid, skipping`);
    return { skipped: true };
  }

  // Use idempotency key with payment provider
  await chargeCustomer(job.data.customerId, job.data.amount, {
    idempotencyKey: `order-${job.data.orderId}`,
  });

  await markOrderPaid(job.data.orderId);
  return { charged: true };
}
```

## Concurrency Control

```typescript
// Per-worker concurrency
const worker = new Worker('queue', processor, {
  connection,
  concurrency: 10, // 10 simultaneous jobs per worker process
});

// Global rate limiting (across all workers)
const worker = new Worker('api-calls', processor, {
  connection,
  concurrency: 5,
  limiter: {
    max: 100,        // Max 100 jobs
    duration: 60_000, // Per minute
    groupKey: 'apiKey', // Rate limit per unique apiKey in job data
  },
});

// Per-job timeout (prevent stuck jobs)
await queue.add('process', data, {
  timeout: 30_000, // Kill job if it runs longer than 30s
});
```

## Health Check Endpoint

```typescript
// src/workers/health.ts
import express from 'express';

const healthApp = express();
const HEALTH_PORT = parseInt(process.env.WORKER_HEALTH_PORT || '8081');

let lastJobProcessed = Date.now();
let jobsProcessed = 0;
let jobsFailed = 0;
let isShuttingDown = false;

export function recordJobComplete() { jobsProcessed++; lastJobProcessed = Date.now(); }
export function recordJobFailed() { jobsFailed++; lastJobProcessed = Date.now(); }
export function setShuttingDown() { isShuttingDown = true; }

healthApp.get('/health', (req, res) => {
  const idleSeconds = (Date.now() - lastJobProcessed) / 1000;
  const memMB = Math.round(process.memoryUsage().heapUsed / 1024 / 1024);

  if (isShuttingDown) {
    return res.status(503).json({ status: 'shutting_down' });
  }

  // Alert if idle too long (might indicate stalled worker)
  if (idleSeconds > 300) {
    return res.status(503).json({
      status: 'idle_too_long',
      idleSeconds,
    });
  }

  res.json({
    status: 'ok',
    uptime: Math.round(process.uptime()),
    jobsProcessed,
    jobsFailed,
    idleSeconds: Math.round(idleSeconds),
    memoryMB: memMB,
  });
});

healthApp.listen(HEALTH_PORT, () => {
  console.log(`Worker health check on :${HEALTH_PORT}/health`);
});
```

## Job Data Best Practices

```typescript
// BAD: References external state that might change
await queue.add('send-email', {
  userId: 123, // What if user is deleted before job runs?
});

// GOOD: Self-contained job data
await queue.add('send-email', {
  email: 'user@example.com',
  subject: 'Welcome!',
  templateId: 'welcome',
  templateData: {
    name: 'Jane',
    activationUrl: 'https://app.com/activate/abc123',
  },
});

// GOOD: Include enough to recover if external state changes
await queue.add('process-order', {
  orderId: 'ord_123',
  amount: 4999,      // Snapshot at time of queueing
  currency: 'usd',
  customerId: 'cus_456',
  customerEmail: 'customer@example.com',
});
```

## Monitoring Patterns

### Queue Depth Alerting

```typescript
import { Queue } from 'bullmq';

async function checkQueueHealth(queue: Queue) {
  const waiting = await queue.getWaitingCount();
  const active = await queue.getActiveCount();
  const delayed = await queue.getDelayedCount();
  const failed = await queue.getFailedCount();

  console.log(`Queue ${queue.name}: waiting=${waiting} active=${active} delayed=${delayed} failed=${failed}`);

  // Alert thresholds
  if (waiting > 1000) alertOps(`${queue.name}: ${waiting} jobs waiting — workers may be overwhelmed`);
  if (failed > 100) alertOps(`${queue.name}: ${failed} failed jobs — check DLQ`);
}
```

## Gotchas

- Never put large payloads in job data. Store in S3/database, pass a reference ID
- Set `lockDuration` longer than your longest job execution time, or BullMQ will consider the job stalled and re-assign it to another worker (causing duplicates)
- Worker `concurrency` is per process. 3 worker processes x concurrency 10 = 30 parallel jobs
- `removeOnComplete: true` deletes job data immediately — use `{ count: 1000 }` or `{ age: 86400 }` to keep some history
- Stalled jobs (worker crash during processing) are auto-retried. Make handlers idempotent
- BullMQ metrics (via `MetricsTime`) store data points in Redis — useful for dashboards but adds memory usage
