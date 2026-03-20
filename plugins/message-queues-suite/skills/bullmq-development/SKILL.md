---
name: bullmq-development
description: >
  BullMQ job queue development — Redis-backed queues with retries, scheduling,
  rate limiting, priorities, flows, and worker management.
  Triggers: "bullmq", "bull queue", "job queue", "background jobs",
  "redis queue", "task queue", "cron jobs node", "delayed jobs".
  NOT for: Kafka event streaming (use kafka-development).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# BullMQ — Redis-Backed Job Queues

## Setup

```bash
npm install bullmq ioredis
```

## Connection

```typescript
// src/lib/queue-connection.ts
import { ConnectionOptions } from "bullmq";

export const connection: ConnectionOptions = {
  host: process.env.REDIS_HOST || "localhost",
  port: parseInt(process.env.REDIS_PORT || "6379"),
  password: process.env.REDIS_PASSWORD,
  maxRetriesPerRequest: null, // Required for BullMQ workers
};
```

## Queue + Worker Pattern

```typescript
// src/queues/email.queue.ts
import { Queue } from "bullmq";
import { connection } from "../lib/queue-connection";

interface EmailJobData {
  to: string;
  subject: string;
  template: string;
  variables: Record<string, string>;
}

export const emailQueue = new Queue<EmailJobData>("email", {
  connection,
  defaultJobOptions: {
    attempts: 3,
    backoff: {
      type: "exponential",
      delay: 1000, // 1s, 2s, 4s
    },
    removeOnComplete: {
      age: 24 * 3600,   // Keep completed jobs for 24h
      count: 1000,       // Keep last 1000
    },
    removeOnFail: {
      age: 7 * 24 * 3600, // Keep failed jobs for 7 days
    },
  },
});
```

```typescript
// src/workers/email.worker.ts
import { Worker, Job } from "bullmq";
import { connection } from "../lib/queue-connection";

const emailWorker = new Worker(
  "email",
  async (job: Job) => {
    const { to, subject, template, variables } = job.data;

    // Update progress
    await job.updateProgress(10);

    // Send email (your email service)
    const result = await sendEmail({ to, subject, template, variables });

    await job.updateProgress(100);

    // Return value is stored as job.returnvalue
    return { messageId: result.id, sentAt: new Date().toISOString() };
  },
  {
    connection,
    concurrency: 5,          // Process 5 jobs simultaneously
    limiter: {
      max: 100,              // Max 100 jobs
      duration: 60_000,      // Per minute (rate limiting)
    },
  }
);

// Event handlers
emailWorker.on("completed", (job) => {
  console.log(`Email job ${job.id} completed: ${job.returnvalue?.messageId}`);
});

emailWorker.on("failed", (job, err) => {
  console.error(`Email job ${job?.id} failed: ${err.message}`);
  if (job && job.attemptsMade >= (job.opts.attempts ?? 0)) {
    // All retries exhausted — alert or move to DLQ
    notifyDLQ("email", job.id!, err.message);
  }
});

emailWorker.on("error", (err) => {
  console.error("Worker error:", err);
});

export default emailWorker;
```

## Adding Jobs

```typescript
// src/services/notification.service.ts
import { emailQueue } from "../queues/email.queue";

// Simple job
await emailQueue.add("welcome-email", {
  to: "user@example.com",
  subject: "Welcome!",
  template: "welcome",
  variables: { name: "Alice" },
});

// Delayed job (send in 1 hour)
await emailQueue.add("reminder", data, {
  delay: 60 * 60 * 1000,
});

// Priority job (lower number = higher priority)
await emailQueue.add("urgent", data, {
  priority: 1,
});

// Unique job (deduplicate by jobId)
await emailQueue.add("daily-digest", data, {
  jobId: `digest-${userId}-${date}`,
});

// Bulk add
await emailQueue.addBulk([
  { name: "welcome", data: userData1 },
  { name: "welcome", data: userData2 },
  { name: "welcome", data: userData3 },
]);
```

## Repeatable / Cron Jobs

```typescript
// Run every day at 9 AM
await emailQueue.add("daily-report", { type: "daily" }, {
  repeat: {
    pattern: "0 9 * * *", // Cron syntax
    tz: "America/New_York",
  },
});

// Run every 5 minutes
await emailQueue.add("health-check", {}, {
  repeat: {
    every: 5 * 60 * 1000, // 5 minutes in ms
  },
});

// List repeatable jobs
const repeatableJobs = await emailQueue.getRepeatableJobs();

// Remove a repeatable job
await emailQueue.removeRepeatableByKey(repeatableJobs[0].key);
```

## Job Flows (Parent-Child Dependencies)

```typescript
import { FlowProducer } from "bullmq";
import { connection } from "../lib/queue-connection";

const flowProducer = new FlowProducer({ connection });

// Parent job waits for all children to complete
const flow = await flowProducer.add({
  name: "process-order",
  queueName: "orders",
  data: { orderId: "123" },
  children: [
    {
      name: "validate-payment",
      queueName: "payments",
      data: { orderId: "123", amount: 99.99 },
    },
    {
      name: "reserve-inventory",
      queueName: "inventory",
      data: { orderId: "123", items: ["SKU-001", "SKU-002"] },
    },
    {
      name: "send-confirmation",
      queueName: "email",
      data: { orderId: "123", to: "buyer@example.com" },
    },
  ],
});

// In the parent worker, access children results:
const orderWorker = new Worker("orders", async (job) => {
  const childrenValues = await job.getChildrenValues();
  // childrenValues = { "bull:payments:jobId": result, "bull:inventory:jobId": result, ... }
  console.log("All children completed:", childrenValues);
});
```

## Queue Events & Monitoring

```typescript
import { QueueEvents } from "bullmq";

const queueEvents = new QueueEvents("email", { connection });

queueEvents.on("completed", ({ jobId, returnvalue }) => {
  console.log(`Job ${jobId} completed with:`, returnvalue);
});

queueEvents.on("failed", ({ jobId, failedReason }) => {
  console.error(`Job ${jobId} failed:`, failedReason);
});

queueEvents.on("progress", ({ jobId, data: progress }) => {
  console.log(`Job ${jobId} progress:`, progress);
});

queueEvents.on("stalled", ({ jobId }) => {
  console.warn(`Job ${jobId} stalled — will be reprocessed`);
});
```

## Dashboard (Bull Board)

```bash
npm install @bull-board/express @bull-board/api
```

```typescript
import { createBullBoard } from "@bull-board/api";
import { BullMQAdapter } from "@bull-board/api/bullMQAdapter";
import { ExpressAdapter } from "@bull-board/express";
import { emailQueue } from "./queues/email.queue";
import { orderQueue } from "./queues/order.queue";

const serverAdapter = new ExpressAdapter();
serverAdapter.setBasePath("/admin/queues");

createBullBoard({
  queues: [
    new BullMQAdapter(emailQueue),
    new BullMQAdapter(orderQueue),
  ],
  serverAdapter,
});

// Mount in Express
app.use("/admin/queues", serverAdapter.getRouter());
```

## Graceful Shutdown

```typescript
// src/server.ts
import emailWorker from "./workers/email.worker";
import orderWorker from "./workers/order.worker";

const workers = [emailWorker, orderWorker];

async function gracefulShutdown(signal: string) {
  console.log(`Received ${signal}. Shutting down gracefully...`);

  // Stop accepting new jobs, finish active ones
  await Promise.all(workers.map((w) => w.close()));

  // Close queue connections
  await emailQueue.close();
  await orderQueue.close();

  console.log("All workers shut down. Exiting.");
  process.exit(0);
}

process.on("SIGTERM", () => gracefulShutdown("SIGTERM"));
process.on("SIGINT", () => gracefulShutdown("SIGINT"));
```

## Dead Letter Queue Pattern

```typescript
// src/queues/dlq.ts
import { Queue, Worker } from "bullmq";
import { connection } from "../lib/queue-connection";

// Create a DLQ for failed jobs
export const deadLetterQueue = new Queue("dead-letter", { connection });

// In your main worker's failed handler:
emailWorker.on("failed", async (job, err) => {
  if (job && job.attemptsMade >= (job.opts.attempts ?? 0)) {
    await deadLetterQueue.add("failed-email", {
      originalQueue: "email",
      originalJobId: job.id,
      originalData: job.data,
      error: err.message,
      stack: err.stack,
      failedAt: new Date().toISOString(),
      attempts: job.attemptsMade,
    });
  }
});

// DLQ worker — alert and store for review
const dlqWorker = new Worker("dead-letter", async (job) => {
  console.error(`DLQ: ${job.data.originalQueue}/${job.data.originalJobId}`, job.data.error);
  // Send alert (Slack, PagerDuty, etc.)
  // Store in database for admin review
}, { connection, concurrency: 1 });
```

## Sandboxed Processors (Separate Process)

```typescript
// src/workers/heavy.worker.ts
import { Worker } from "bullmq";
import { connection } from "../lib/queue-connection";
import path from "path";

// Run processor in a separate child process
// Prevents CPU-heavy jobs from blocking the event loop
const heavyWorker = new Worker(
  "image-processing",
  path.join(__dirname, "processors/image.processor.js"),
  {
    connection,
    concurrency: 2,
    useWorkerThreads: true, // Use worker threads instead of child processes
  }
);
```

```typescript
// src/workers/processors/image.processor.ts
import { Job } from "bullmq";

export default async function (job: Job) {
  const { imageUrl, operations } = job.data;

  await job.updateProgress(0);
  const image = await downloadImage(imageUrl);

  await job.updateProgress(30);
  const processed = await applyOperations(image, operations);

  await job.updateProgress(70);
  const url = await uploadToS3(processed);

  await job.updateProgress(100);
  return { url };
}
```

## Testing

```typescript
import { Queue, Worker, Job } from "bullmq";
import IORedis from "ioredis";

describe("Email Worker", () => {
  let queue: Queue;
  let worker: Worker;
  let connection: IORedis;

  beforeAll(() => {
    connection = new IORedis({ maxRetriesPerRequest: null });
    queue = new Queue("test-email", { connection });
  });

  afterAll(async () => {
    await queue.obliterate({ force: true }); // Clean up test queue
    await queue.close();
    await connection.quit();
  });

  it("processes email jobs", async () => {
    const completedPromise = new Promise<Job>((resolve) => {
      worker = new Worker("test-email", async (job) => {
        return { sent: true };
      }, { connection });
      worker.on("completed", resolve);
    });

    await queue.add("test", {
      to: "test@example.com",
      subject: "Test",
      template: "test",
      variables: {},
    });

    const job = await completedPromise;
    expect(job.returnvalue).toEqual({ sent: true });
    await worker.close();
  });
});
```

## Gotchas

1. **`maxRetriesPerRequest: null` is required** — BullMQ workers need this Redis option or they'll throw `MaxRetriesPerRequestError`. The default `ioredis` setting of 20 retries conflicts with BullMQ's blocking commands.

2. **Jobs are NOT removed by default** — Completed and failed jobs accumulate in Redis forever. Always set `removeOnComplete` and `removeOnFail` to prevent memory leaks.

3. **Stalled jobs happen when workers crash** — If a worker process dies mid-job, the job is "stalled." BullMQ auto-detects and re-enqueues stalled jobs (default: 30s check interval). Make handlers idempotent.

4. **Rate limiter is per-worker, not global** — `limiter: { max: 100, duration: 60000 }` limits each worker instance to 100/min. With 3 workers, you get 300/min total. For global rate limiting, use a shared Redis counter.

5. **Repeatable jobs need a running Queue instance** — The repeat schedule is stored in Redis, but a Queue instance must be active to trigger jobs. If your server restarts, repeatable jobs resume automatically once the Queue reconnects.

6. **Flow children run in parallel** — FlowProducer children are NOT sequential. They all start simultaneously. For sequential execution, chain them: child1 -> child2 -> child3 -> parent.
