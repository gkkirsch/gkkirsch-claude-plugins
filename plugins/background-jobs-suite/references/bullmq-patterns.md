# BullMQ Patterns & Configuration Reference

Quick reference for BullMQ configuration, job options, and common patterns.

---

## Job Options Reference

```typescript
await queue.add('job-name', { /* data */ }, {
  // Retry
  attempts: 5,                               // Max attempts (including first)
  backoff: { type: 'exponential', delay: 1000 }, // 1s, 2s, 4s, 8s, 16s
  // backoff: { type: 'fixed', delay: 5000 },  // 5s, 5s, 5s, 5s

  // Scheduling
  delay: 60_000,                             // Wait 60s before first processing
  priority: 1,                               // Lower = higher priority (1-2097152)

  // Deduplication
  jobId: 'unique-key',                       // Prevents duplicate jobs with same ID

  // Cleanup
  removeOnComplete: true,                    // Remove when done
  removeOnComplete: { count: 1000 },         // Keep last 1000
  removeOnComplete: { age: 24 * 3600 },      // Keep for 24 hours
  removeOnFail: true,                        // Remove on failure
  removeOnFail: { count: 5000 },             // Keep last 5000 failed
  removeOnFail: { age: 7 * 24 * 3600 },     // Keep failed for 7 days

  // Execution
  timeout: 30_000,                           // Kill if running > 30s
  timestamp: Date.now(),                     // Custom timestamp

  // Repeatable
  repeat: {
    pattern: '0 9 * * *',                   // Cron pattern
    tz: 'America/New_York',                  // Timezone
    // OR:
    every: 5 * 60 * 1000,                   // Every 5 minutes (ms)
    limit: 100,                              // Max 100 repetitions
  },
});
```

## Worker Options Reference

```typescript
const worker = new Worker('queue-name', processor, {
  connection,
  concurrency: 10,                           // Parallel jobs per worker

  limiter: {
    max: 100,                                // Max jobs
    duration: 60_000,                        // Per time window (ms)
    groupKey: 'tenantId',                    // Rate limit per group
  },

  lockDuration: 30_000,                      // Job lock (ms), default 30s
  lockRenewTime: 15_000,                     // Lock renewal interval
  stalledInterval: 30_000,                   // Check stalled jobs interval
  maxStalledCount: 1,                        // Max stalls before failing

  autorun: true,                             // Start processing immediately
  runRetryDelay: 15_000,                     // Delay before retrying after error

  metrics: {
    maxDataPoints: MetricsTime.ONE_WEEK,     // Store metrics for 1 week
  },
});
```

## Queue Methods Cheat Sheet

```typescript
const queue = new Queue('my-queue', { connection });

// Add jobs
await queue.add('name', data);                        // Single job
await queue.addBulk([{ name, data }, ...]);           // Bulk add

// Query jobs
await queue.getJob('job-id');                         // Get specific job
await queue.getJobs(['waiting', 'active']);            // Get by state
await queue.getWaitingCount();                        // Count waiting
await queue.getActiveCount();                         // Count active
await queue.getCompletedCount();                      // Count completed
await queue.getFailedCount();                         // Count failed
await queue.getDelayedCount();                        // Count delayed

// Repeatable jobs
await queue.getRepeatableJobs();                      // List all repeatable
await queue.removeRepeatableByKey(key);               // Remove specific

// Management
await queue.pause();                                  // Pause processing
await queue.resume();                                 // Resume processing
await queue.clean(3600 * 1000, 1000, 'completed');    // Clean old jobs
await queue.obliterate({ force: true });              // Delete queue + all data
await queue.drain();                                  // Remove all waiting jobs
```

## Job Methods Cheat Sheet

```typescript
// Inside worker processor
async (job: Job) => {
  job.id;                    // Job ID
  job.name;                  // Job name
  job.data;                  // Job payload
  job.opts;                  // Job options
  job.attemptsMade;          // Current attempt number
  job.timestamp;             // When job was created

  await job.updateProgress(50);          // Report progress (0-100)
  await job.updateData({ ...newData });  // Update job data
  await job.log('Processing step 3');    // Add log entry

  return { result: 'value' };            // Return value stored as result
};

// Outside worker (query existing jobs)
const job = await queue.getJob('job-id');
await job.retry();                       // Retry a failed job
await job.remove();                      // Delete job
await job.moveToFailed(err, token);      // Manually fail
const state = await job.getState();      // 'waiting'|'active'|'completed'|'failed'|'delayed'
const logs = await queue.getJobLogs(job.id); // Get job logs
```

## Flows (Job Dependencies)

```typescript
import { FlowProducer } from 'bullmq';

const flow = new FlowProducer({ connection });

// Parent waits for all children to complete
const tree = await flow.add({
  name: 'send-invoice',
  queueName: 'email',
  data: { invoiceId: 123 },
  children: [
    {
      name: 'generate-pdf',
      queueName: 'documents',
      data: { invoiceId: 123 },
    },
    {
      name: 'calculate-tax',
      queueName: 'billing',
      data: { invoiceId: 123 },
    },
  ],
});

// In parent worker, access children results
async (job: Job) => {
  const childrenValues = await job.getChildrenValues();
  // { 'bull:documents:generate-pdf:123': { pdfUrl: '...' }, ... }
};
```

## Event Reference

### Worker Events

| Event | Payload | When |
|-------|---------|------|
| `completed` | `(job, returnvalue)` | Job finished successfully |
| `failed` | `(job, error)` | Job threw an error |
| `progress` | `(job, progress)` | `job.updateProgress()` called |
| `error` | `(error)` | Worker-level error |
| `stalled` | `(jobId)` | Job lock expired, will be reprocessed |
| `active` | `(job)` | Job started processing |
| `drained` | — | Queue is empty |
| `closing` | — | `worker.close()` called |
| `closed` | — | Worker fully closed |
| `ready` | — | Worker connected to Redis |

### QueueEvents Events

| Event | Payload | When |
|-------|---------|------|
| `completed` | `{ jobId, returnvalue }` | Job completed (any worker) |
| `failed` | `{ jobId, failedReason }` | Job failed (any worker) |
| `progress` | `{ jobId, data }` | Progress updated |
| `waiting` | `{ jobId }` | Job added to queue |
| `active` | `{ jobId }` | Job started |
| `delayed` | `{ jobId, delay }` | Job delayed |
| `stalled` | `{ jobId }` | Job stalled |
| `removed` | `{ jobId }` | Job removed |
| `duplicated` | `{ jobId }` | Duplicate job rejected |

## Redis Memory Estimation

| Jobs Stored | Avg Job Size | Estimated Redis Memory |
|-------------|-------------|----------------------|
| 1,000 | 1 KB | ~5 MB |
| 10,000 | 1 KB | ~50 MB |
| 100,000 | 1 KB | ~500 MB |
| 1,000,000 | 1 KB | ~5 GB |

Factors that increase memory:
- Large job data payloads
- Job logs (`job.log()`)
- Long `removeOnComplete`/`removeOnFail` retention
- Many repeatable job patterns
- Flow/dependency metadata

**Rule of thumb**: Set `removeOnComplete: { age: 86400 }` (1 day) and `removeOnFail: { age: 604800 }` (7 days) unless you need longer retention.

## Heroku-Specific Configuration

```typescript
// Redis connection for Heroku Redis
const connection = new IORedis(process.env.REDIS_URL, {
  maxRetriesPerRequest: null,    // Required by BullMQ
  enableReadyCheck: false,
  tls: process.env.REDIS_URL?.startsWith('rediss://') ? {
    rejectUnauthorized: false,   // Heroku Redis uses self-signed certs
  } : undefined,
});
```

```
# Procfile
web: node dist/server.js
worker: node dist/workers/index.js
```

```bash
# Scale workers independently
heroku ps:scale worker=2
heroku ps:scale web=1 worker=3
```

Heroku Redis plans:
| Plan | Connections | Memory | Price |
|------|------------|--------|-------|
| Mini | 20 | 25 MB | $3/mo |
| Premium 0 | 40 | 50 MB | $15/mo |
| Premium 1 | 80 | 100 MB | $30/mo |
| Premium 2 | 120 | 250 MB | $60/mo |

## Testing Workers

```typescript
// test/workers/email.worker.test.ts
import { Queue, Worker, Job } from 'bullmq';
import IORedis from 'ioredis';

const testConnection = new IORedis({ maxRetriesPerRequest: null });
const testQueue = new Queue('test-email', { connection: testConnection });

afterAll(async () => {
  await testQueue.obliterate({ force: true });
  await testConnection.quit();
});

test('processes email job', async () => {
  const results: any[] = [];

  const worker = new Worker('test-email', async (job) => {
    results.push(job.data);
    return { sent: true };
  }, { connection: testConnection });

  await testQueue.add('welcome', { to: 'test@test.com', template: 'welcome' });

  // Wait for processing
  await new Promise<void>((resolve) => {
    worker.on('completed', () => {
      resolve();
    });
  });

  expect(results).toHaveLength(1);
  expect(results[0].to).toBe('test@test.com');

  await worker.close();
});
```
