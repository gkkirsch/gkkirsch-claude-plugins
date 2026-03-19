# BullMQ Patterns & API Reference

## Queue Options Reference

```typescript
new Queue('name', {
  connection: IORedis,           // Required: Redis connection
  prefix: 'bull',               // Key prefix in Redis (default: 'bull')
  defaultJobOptions: {
    attempts: 3,                 // Max retry attempts
    backoff: {
      type: 'exponential',      // 'exponential' | 'fixed' | 'custom'
      delay: 1000,              // Base delay in ms
    },
    delay: 0,                   // Initial delay before first attempt
    priority: 0,                // Lower = higher priority
    lifo: false,                // Last-in-first-out
    removeOnComplete: true,     // true | false | { count, age }
    removeOnFail: false,        // true | false | { count, age }
    timestamp: Date.now(),      // Job creation timestamp
    stackTraceLimit: 0,         // Stack trace lines to keep
  },
});
```

## Worker Options Reference

```typescript
new Worker('name', handler, {
  connection: IORedis,
  concurrency: 1,               // Parallel jobs (default: 1)
  maxStalledCount: 1,           // Max stall retries before fail
  stalledInterval: 30000,       // Stall check interval (ms)
  lockDuration: 30000,          // Job lock duration (ms)
  lockRenewTime: 15000,         // Lock renewal interval (ms)
  drainDelay: 5,                // Delay between empty polls (ms)
  limiter: {
    max: 100,                   // Max jobs per duration
    duration: 60000,            // Duration window (ms)
    groupKey: 'userId',         // Group rate limit by field
  },
  settings: {
    backoffStrategy: (attempt) => delay,  // Custom backoff function
  },
  autorun: true,                // Start processing immediately
  runRetryDelay: 15000,         // Delay before retrying run errors
  skipLockRenewal: false,       // Skip lock renewal (for fast jobs)
});
```

## Job Methods

```typescript
const job = await queue.add('name', data, options);

// State management
await job.getState();           // 'waiting' | 'active' | 'completed' | 'failed' | 'delayed'
await job.moveToFailed(error);  // Force fail
await job.moveToCompleted(result); // Force complete
await job.remove();             // Delete from queue
await job.retry();              // Retry a failed job
await job.promote();            // Move delayed job to waiting

// Progress
await job.updateProgress(50);   // Update progress (0-100 or object)
await job.log('Processing step 2...');  // Add log entry

// Data access
job.data;                       // Job payload
job.returnvalue;                // Return value after completion
job.failedReason;               // Error message after failure
job.attemptsMade;               // Number of attempts so far
job.timestamp;                  // Creation timestamp
job.processedOn;                // Start processing timestamp
job.finishedOn;                 // Completion timestamp
```

## Queue Methods

```typescript
// Adding jobs
await queue.add('name', data, options);
await queue.addBulk([{ name, data, opts }, ...]);

// Job retrieval
await queue.getJob(id);
await queue.getJobs(['waiting', 'active'], 0, 100);
await queue.getJobCounts('waiting', 'active', 'completed', 'failed', 'delayed');

// Queue management
await queue.pause();            // Pause processing
await queue.resume();           // Resume processing
await queue.drain();            // Remove all waiting jobs
await queue.clean(grace, limit, type);  // Remove old jobs
await queue.obliterate();       // Remove queue entirely

// Repeatable jobs
await queue.getRepeatableJobs();
await queue.removeRepeatableByKey(key);
```

## Event Reference

### Worker Events

| Event | Args | When |
|-------|------|------|
| `completed` | `(job, result)` | Job finished successfully |
| `failed` | `(job, error)` | Job failed (may retry) |
| `progress` | `(job, progress)` | Job progress updated |
| `active` | `(job)` | Job started processing |
| `stalled` | `(jobId)` | Job stalled (worker crashed) |
| `error` | `(error)` | Worker error |
| `drained` | — | Queue is empty |
| `closing` | — | Worker is shutting down |
| `closed` | — | Worker has shut down |

### QueueEvents Events

| Event | Args | When |
|-------|------|------|
| `waiting` | `{ jobId }` | Job added to queue |
| `active` | `{ jobId, prev }` | Job started processing |
| `completed` | `{ jobId, returnvalue }` | Job completed |
| `failed` | `{ jobId, failedReason }` | Job failed |
| `delayed` | `{ jobId, delay }` | Job delayed |
| `removed` | `{ jobId }` | Job removed |
| `progress` | `{ jobId, data }` | Progress updated |
| `stalled` | `{ jobId }` | Job stalled |

## Flow Producer Patterns

### Sequential Steps

```typescript
const flow = new FlowProducer({ connection });

await flow.add({
  name: 'final-step',
  queueName: 'pipeline',
  data: { reportId: '123' },
  children: [
    {
      name: 'step-2',
      queueName: 'pipeline',
      data: { step: 2 },
      children: [
        {
          name: 'step-1',
          queueName: 'pipeline',
          data: { step: 1 },
          // step-1 runs first (leaf), then step-2, then final-step
        },
      ],
    },
  ],
});
```

### Parallel Fan-Out

```typescript
await flow.add({
  name: 'aggregate-results',
  queueName: 'pipeline',
  data: {},
  children: [
    { name: 'fetch-a', queueName: 'fetching', data: { source: 'a' } },
    { name: 'fetch-b', queueName: 'fetching', data: { source: 'b' } },
    { name: 'fetch-c', queueName: 'fetching', data: { source: 'c' } },
    // All three run in parallel, aggregate-results runs after all complete
  ],
});
```

### Accessing Parent/Child Data

```typescript
// In the parent job handler, access children's return values
const worker = new Worker('pipeline', async (job) => {
  // Get children results
  const childValues = await job.getChildrenValues();
  // { 'bull:fetching:job-id-a': { data: 'from-a' }, ... }

  // Or get dependency info
  const deps = await job.getDependencies();
  // { processed: { ... }, nextProcessedCursor: 0, unprocessed: [] }
});
```

## Sandboxed Workers

Run job handlers in separate Node.js processes (crash isolation):

```typescript
// src/workers/heavy-worker.ts
import { Worker } from 'bullmq';

// Handler runs in a child process
const worker = new Worker('heavy-processing', './dist/handlers/heavy.js', {
  connection,
  useWorkerThreads: true,  // Use worker threads instead of child processes
});
```

```typescript
// src/handlers/heavy.js (separate file)
import { SandboxedJob } from 'bullmq';

export default async function (job: SandboxedJob) {
  // This runs in isolation — if it crashes, the main worker survives
  const result = await heavyComputation(job.data);
  return result;
}
```

## Common Patterns

### Job Deduplication

```typescript
// jobId prevents duplicate jobs
await queue.add('sync-user', { userId: '123' }, {
  jobId: 'sync-user-123',  // Same ID = same job (won't duplicate)
});

// Multiple adds with same jobId are no-ops
await queue.add('sync-user', { userId: '123' }, { jobId: 'sync-user-123' });
// ^ This does nothing — job already exists
```

### Job Timeout

```typescript
const worker = new Worker('api-calls', async (job) => {
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), 30_000);

  try {
    const result = await fetch(url, { signal: controller.signal });
    clearTimeout(timeout);
    return result;
  } catch (e: any) {
    clearTimeout(timeout);
    if (e.name === 'AbortError') throw new Error('Job timed out after 30s');
    throw e;
  }
});
```

### Group Rate Limiting

```typescript
// Rate limit per user (not global)
const worker = new Worker('notifications', handler, {
  connection,
  limiter: {
    max: 5,
    duration: 60_000,
    groupKey: 'userId',  // Each userId gets its own rate limit
  },
});

// This user gets max 5 notifications per minute
await queue.add('notify', { userId: 'abc', message: '...' });
```

### Pause/Resume Per Queue

```typescript
// Pause a specific queue (workers stop picking up new jobs)
await emailQueue.pause();

// Resume
await emailQueue.resume();

// Check if paused
const isPaused = await emailQueue.isPaused();
```
