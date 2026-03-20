---
name: queue-monitoring
description: >
  Message queue monitoring, dead letter queues, and retry patterns.
  Use when implementing DLQ handling, monitoring queue health,
  building retry strategies, or debugging failed queue jobs.
  Triggers: "dead letter queue", "DLQ", "queue monitoring", "retry pattern",
  "job failure", "queue health", "queue observability", "failed jobs",
  "queue backpressure", "message retry".
  NOT for: BullMQ setup (see bullmq-development), Kafka setup (see kafka-development), in-process event buses.
version: 1.0.0
allowed-tools: Read, Grep, Glob, Bash
---

# Queue Monitoring & Reliability

## Dead Letter Queue Pattern (BullMQ)

```typescript
// queues/dlq-handler.ts — Automatic DLQ routing for failed jobs
import { Queue, Worker, Job } from 'bullmq';
import Redis from 'ioredis';

const connection = new Redis(process.env.REDIS_URL!, { maxRetriesPerRequest: null });

// Main processing queue
const emailQueue = new Queue('email-sending', { connection });

// Dead letter queue for permanently failed jobs
const emailDLQ = new Queue('email-sending-dlq', { connection });

// Configure retry with exponential backoff
await emailQueue.add('send-welcome', { userId: 'user-1', template: 'welcome' }, {
  attempts: 5,
  backoff: {
    type: 'exponential',
    delay: 1000, // 1s, 2s, 4s, 8s, 16s
  },
  removeOnComplete: { age: 86400, count: 1000 }, // Keep 1000 or 24h
  removeOnFail: false, // Keep failed jobs for inspection
});

// Worker with DLQ routing
const worker = new Worker('email-sending', async (job: Job) => {
  const { userId, template } = job.data;

  try {
    await sendEmail(userId, template);
  } catch (error) {
    // Classify error: retryable vs permanent
    if (isRetryableError(error)) {
      throw error; // Let BullMQ retry
    }

    // Permanent failure → route to DLQ
    await emailDLQ.add('failed-email', {
      originalJob: {
        id: job.id,
        data: job.data,
        attemptsMade: job.attemptsMade,
        failedReason: (error as Error).message,
        timestamp: job.timestamp,
      },
      failedAt: new Date().toISOString(),
      errorType: classifyError(error),
    });

    // Don't throw — mark as completed to stop retries
    return { status: 'routed-to-dlq', reason: (error as Error).message };
  }
}, { connection, concurrency: 10 });

// Error classification
function isRetryableError(error: unknown): boolean {
  const msg = (error as Error).message?.toLowerCase() ?? '';
  return (
    msg.includes('timeout') ||
    msg.includes('econnrefused') ||
    msg.includes('rate limit') ||
    msg.includes('503') ||
    msg.includes('429')
  );
}

function classifyError(error: unknown): string {
  const msg = (error as Error).message?.toLowerCase() ?? '';
  if (msg.includes('invalid email')) return 'invalid_recipient';
  if (msg.includes('unsubscribed')) return 'unsubscribed';
  if (msg.includes('bounced')) return 'hard_bounce';
  if (msg.includes('spam')) return 'spam_complaint';
  return 'unknown';
}

// Event listeners for monitoring
worker.on('completed', (job) => {
  metrics.increment('email.sent.success');
});

worker.on('failed', (job, error) => {
  metrics.increment('email.sent.failed');
  console.error(`Job ${job?.id} failed after ${job?.attemptsMade} attempts:`, error.message);
});

worker.on('error', (error) => {
  console.error('Worker error:', error);
});
```

## Queue Health Dashboard

```typescript
// api/queue-health.ts — Expose queue metrics
import { Queue } from 'bullmq';

interface QueueHealth {
  name: string;
  counts: {
    waiting: number;
    active: number;
    completed: number;
    failed: number;
    delayed: number;
    paused: number;
  };
  throughput: {
    processed1h: number;
    failed1h: number;
    successRate: number;
  };
  latency: {
    avgProcessingMs: number;
    avgWaitMs: number;
    oldestWaitingAge: number | null;
  };
  workers: {
    count: number;
    utilization: number;
  };
  alerts: string[];
}

async function getQueueHealth(queue: Queue): Promise<QueueHealth> {
  const counts = await queue.getJobCounts(
    'waiting', 'active', 'completed', 'failed', 'delayed', 'paused'
  );

  // Calculate throughput from recent completed/failed jobs
  const oneHourAgo = Date.now() - 3600_000;
  const recentCompleted = await queue.getCompleted(0, 100);
  const recentFailed = await queue.getFailed(0, 100);

  const processed1h = recentCompleted.filter(j => j.finishedOn && j.finishedOn > oneHourAgo).length;
  const failed1h = recentFailed.filter(j => j.finishedOn && j.finishedOn > oneHourAgo).length;

  // Calculate latency from recent jobs
  const processingTimes = recentCompleted
    .filter(j => j.finishedOn && j.processedOn)
    .map(j => j.finishedOn! - j.processedOn!);
  const waitTimes = recentCompleted
    .filter(j => j.processedOn)
    .map(j => j.processedOn! - j.timestamp);

  // Find oldest waiting job
  const waiting = await queue.getWaiting(0, 1);
  const oldestWaitingAge = waiting.length > 0 ? Date.now() - waiting[0].timestamp : null;

  // Worker info
  const workers = await queue.getWorkers();

  // Generate alerts
  const alerts: string[] = [];
  if (counts.waiting > 10000) alerts.push('Queue depth exceeds 10,000');
  if (counts.failed > 100) alerts.push('More than 100 failed jobs');
  if (oldestWaitingAge && oldestWaitingAge > 300_000) alerts.push('Oldest job waiting >5 minutes');
  if (processed1h > 0 && failed1h / processed1h > 0.1) alerts.push('Failure rate >10% in last hour');
  if (workers.length === 0) alerts.push('No active workers');

  return {
    name: queue.name,
    counts,
    throughput: {
      processed1h,
      failed1h,
      successRate: processed1h > 0 ? Math.round((1 - failed1h / (processed1h + failed1h)) * 100) : 100,
    },
    latency: {
      avgProcessingMs: processingTimes.length > 0
        ? Math.round(processingTimes.reduce((a, b) => a + b, 0) / processingTimes.length)
        : 0,
      avgWaitMs: waitTimes.length > 0
        ? Math.round(waitTimes.reduce((a, b) => a + b, 0) / waitTimes.length)
        : 0,
      oldestWaitingAge,
    },
    workers: {
      count: workers.length,
      utilization: workers.length > 0
        ? Math.round((counts.active / (workers.length * 10)) * 100) // Assuming concurrency=10
        : 0,
    },
    alerts,
  };
}

// Express endpoint
app.get('/api/queues/health', async (req, res) => {
  const queues = [emailQueue, paymentQueue, notificationQueue];
  const health = await Promise.all(queues.map(getQueueHealth));
  res.json({
    timestamp: new Date().toISOString(),
    queues: health,
    summary: {
      totalFailed: health.reduce((sum, q) => sum + q.counts.failed, 0),
      totalWaiting: health.reduce((sum, q) => sum + q.counts.waiting, 0),
      hasAlerts: health.some(q => q.alerts.length > 0),
    },
  });
});
```

## Retry Strategies

```typescript
// lib/retry-strategies.ts

interface RetryConfig {
  maxAttempts: number;
  strategy: 'fixed' | 'exponential' | 'linear' | 'custom';
  baseDelay: number;
  maxDelay?: number;
  jitter?: boolean;
  retryableErrors?: string[];
}

const STRATEGIES: Record<string, RetryConfig> = {
  // Critical path: aggressive retries with short delays
  payment: {
    maxAttempts: 3,
    strategy: 'fixed',
    baseDelay: 2000,      // 2s between each retry
    retryableErrors: ['timeout', 'rate_limit', 'gateway_error'],
  },

  // Idempotent: many retries with exponential backoff
  email: {
    maxAttempts: 5,
    strategy: 'exponential',
    baseDelay: 1000,
    maxDelay: 60_000,     // Cap at 1 minute
    jitter: true,         // Random jitter to prevent thundering herd
  },

  // External API: linear backoff respecting rate limits
  webhook: {
    maxAttempts: 10,
    strategy: 'linear',
    baseDelay: 5000,      // 5s, 10s, 15s, 20s...
    maxDelay: 300_000,    // Cap at 5 minutes
  },

  // Low priority: slow retries over hours
  report: {
    maxAttempts: 3,
    strategy: 'exponential',
    baseDelay: 60_000,    // 1min, 2min, 4min
    maxDelay: 3600_000,   // Cap at 1 hour
  },
};

function calculateDelay(config: RetryConfig, attempt: number): number {
  let delay: number;

  switch (config.strategy) {
    case 'fixed':
      delay = config.baseDelay;
      break;
    case 'exponential':
      delay = config.baseDelay * Math.pow(2, attempt - 1);
      break;
    case 'linear':
      delay = config.baseDelay * attempt;
      break;
    default:
      delay = config.baseDelay;
  }

  // Apply max delay cap
  delay = Math.min(delay, config.maxDelay ?? Infinity);

  // Apply jitter (0.5x to 1.5x randomization)
  if (config.jitter) {
    delay = delay * (0.5 + Math.random());
  }

  return Math.round(delay);
}
```

## DLQ Reprocessing

```typescript
// scripts/reprocess-dlq.ts — Retry failed DLQ items
import { Queue, Job } from 'bullmq';

async function reprocessDLQ(
  dlqName: string,
  targetQueueName: string,
  options: {
    filter?: (job: Job) => boolean;
    limit?: number;
    dryRun?: boolean;
  } = {}
): Promise<{ reprocessed: number; skipped: number; errors: number }> {
  const dlq = new Queue(dlqName, { connection });
  const targetQueue = new Queue(targetQueueName, { connection });

  const failedJobs = await dlq.getWaiting(0, options.limit ?? 100);
  let reprocessed = 0, skipped = 0, errors = 0;

  for (const job of failedJobs) {
    // Apply filter
    if (options.filter && !options.filter(job)) {
      skipped++;
      continue;
    }

    if (options.dryRun) {
      console.log(`[DRY RUN] Would reprocess: ${job.id}`, job.data.originalJob?.data);
      reprocessed++;
      continue;
    }

    try {
      // Re-add to original queue with fresh retry count
      await targetQueue.add(
        job.data.originalJob?.name ?? 'reprocessed',
        job.data.originalJob?.data ?? job.data,
        {
          attempts: 3,
          backoff: { type: 'exponential', delay: 2000 },
        }
      );

      // Remove from DLQ
      await job.remove();
      reprocessed++;
    } catch (error) {
      console.error(`Failed to reprocess ${job.id}:`, error);
      errors++;
    }
  }

  return { reprocessed, skipped, errors };
}

// Usage:
// await reprocessDLQ('email-sending-dlq', 'email-sending', {
//   filter: (job) => job.data.errorType !== 'invalid_recipient', // Don't retry invalid emails
//   limit: 50,
//   dryRun: true, // Preview first
// });
```

## Gotchas

1. **Retry without idempotency causes duplicates** -- If a job partially succeeds (email sent, but status update fails), retrying sends the email again. Every retryable operation must be idempotent. Use idempotency keys, database upserts, or "check then act" guards before performing side effects.

2. **Exponential backoff without cap** -- `2^n * 1000ms` with 20 attempts means the 20th retry waits 12 days. Always set `maxDelay` to a reasonable cap (5 minutes for user-facing, 1 hour for background). Without a cap, jobs effectively never retry after a few attempts.

3. **DLQ without alerting** -- A dead letter queue that silently accumulates failed jobs is worse than no DLQ. Set up alerts when DLQ depth exceeds zero (for critical queues) or exceeds a threshold (for high-volume queues). Check DLQ as part of daily operational health.

4. **Monitoring only counts, not latency** -- Queue depth tells you how many jobs are waiting, but not how long. 100 jobs waiting could be fine (burst) or terrible (stuck consumer). Track both queue depth AND oldest job age. Alert on age, not just count.

5. **No graceful shutdown** -- Killing a worker mid-job leaves jobs in a "stuck active" state. They won't be retried until the stale lock expires (default 30s in BullMQ). Implement `SIGTERM` handlers that call `worker.close()` to finish active jobs before exiting. Set `stalledInterval` appropriately.

6. **Testing with real Redis** -- Unit tests that depend on a real Redis instance are slow and flaky. Use `ioredis-mock` for unit tests and reserve real Redis for integration tests. BullMQ's internal state machine has many edge cases that mocks don't cover, so integration tests are still essential.
