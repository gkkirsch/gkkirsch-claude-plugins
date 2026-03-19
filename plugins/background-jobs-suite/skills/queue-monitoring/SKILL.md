---
name: queue-monitoring
description: >
  Set up job queue monitoring with Bull Board UI, Prometheus metrics,
  health checks, alerting, and dead letter queue management. Production
  observability for BullMQ-based systems.
  Triggers: "queue dashboard", "Bull Board", "job monitoring", "queue metrics",
  "dead letter queue", "DLQ", "job health check", "queue observability".
  NOT for: basic BullMQ setup (use bullmq-setup), application monitoring (use observability plugins).
version: 1.0.0
allowed-tools: Read, Grep, Glob, Bash, Edit, Write
---

# Queue Monitoring & Observability

## Bull Board — Visual Dashboard

### Setup

```bash
npm install @bull-board/express @bull-board/api
```

```typescript
// src/monitoring/bull-board.ts
import { createBullBoard } from '@bull-board/api';
import { BullMQAdapter } from '@bull-board/api/bullMQAdapter';
import { ExpressAdapter } from '@bull-board/express';
import { emailQueue } from '../queues/email.queue';
import { imageQueue } from '../queues/image.queue';
import { webhookQueue } from '../queues/webhook.queue';
import { exportQueue } from '../queues/export.queue';

const serverAdapter = new ExpressAdapter();
serverAdapter.setBasePath('/admin/queues');

createBullBoard({
  queues: [
    new BullMQAdapter(emailQueue),
    new BullMQAdapter(imageQueue),
    new BullMQAdapter(webhookQueue),
    new BullMQAdapter(exportQueue),
  ],
  serverAdapter,
});

export { serverAdapter as bullBoardAdapter };
```

```typescript
// src/server.ts
import { bullBoardAdapter } from './monitoring/bull-board';

// Protect with auth middleware
app.use('/admin/queues', requireAdmin, bullBoardAdapter.getRouter());
```

### Bull Board Features

- View job counts by status (waiting, active, completed, failed, delayed)
- Inspect individual job data and return values
- Retry failed jobs manually
- Remove completed/failed jobs
- View job progress and logs
- Pause/resume queues

---

## Prometheus Metrics

### Custom Metrics Collector

```typescript
// src/monitoring/queue-metrics.ts
import { Queue } from 'bullmq';
import { connection } from '../lib/redis';

interface QueueMetrics {
  name: string;
  waiting: number;
  active: number;
  completed: number;
  failed: number;
  delayed: number;
  paused: number;
}

const queues = [
  new Queue('email', { connection }),
  new Queue('image-processing', { connection }),
  new Queue('webhooks', { connection }),
  new Queue('exports', { connection }),
  new Queue('scheduled-tasks', { connection }),
];

export async function collectQueueMetrics(): Promise<QueueMetrics[]> {
  return Promise.all(
    queues.map(async (queue) => {
      const counts = await queue.getJobCounts(
        'waiting', 'active', 'completed', 'failed', 'delayed', 'paused'
      );
      return { name: queue.name, ...counts };
    })
  );
}

// Prometheus format endpoint
export async function prometheusMetrics(): Promise<string> {
  const metrics = await collectQueueMetrics();
  const lines: string[] = [
    '# HELP bullmq_queue_size Number of jobs in each state',
    '# TYPE bullmq_queue_size gauge',
  ];

  for (const m of metrics) {
    for (const [state, count] of Object.entries(m)) {
      if (state === 'name') continue;
      lines.push(`bullmq_queue_size{queue="${m.name}",state="${state}"} ${count}`);
    }
  }

  return lines.join('\n');
}
```

```typescript
// Expose as endpoint
app.get('/metrics', async (req, res) => {
  res.set('Content-Type', 'text/plain');
  res.send(await prometheusMetrics());
});
```

### Worker-Level Metrics

```typescript
// src/monitoring/worker-metrics.ts
import { Worker } from 'bullmq';

interface WorkerStats {
  jobsCompleted: number;
  jobsFailed: number;
  totalDuration: number;
  avgDuration: number;
  lastJobAt: Date | null;
}

const stats = new Map<string, WorkerStats>();

export function trackWorker(worker: Worker) {
  const name = worker.name;
  stats.set(name, {
    jobsCompleted: 0,
    jobsFailed: 0,
    totalDuration: 0,
    avgDuration: 0,
    lastJobAt: null,
  });

  worker.on('completed', (job) => {
    const s = stats.get(name)!;
    s.jobsCompleted++;
    if (job.processedOn && job.finishedOn) {
      const duration = job.finishedOn - job.processedOn;
      s.totalDuration += duration;
      s.avgDuration = s.totalDuration / s.jobsCompleted;
    }
    s.lastJobAt = new Date();
  });

  worker.on('failed', () => {
    const s = stats.get(name)!;
    s.jobsFailed++;
    s.lastJobAt = new Date();
  });
}

export function getWorkerStats(): Map<string, WorkerStats> {
  return stats;
}
```

---

## Health Check Endpoint

```typescript
// src/routes/health.ts
import { collectQueueMetrics } from '../monitoring/queue-metrics';
import { getWorkerStats } from '../monitoring/worker-metrics';

router.get('/health/queues', async (req, res) => {
  const metrics = await collectQueueMetrics();
  const workerStats = getWorkerStats();

  const issues: string[] = [];

  for (const queue of metrics) {
    // Alert if too many waiting jobs
    if (queue.waiting > 1000) {
      issues.push(`${queue.name}: ${queue.waiting} waiting (backlog)`);
    }

    // Alert if jobs are failing
    if (queue.failed > 100) {
      issues.push(`${queue.name}: ${queue.failed} failed jobs`);
    }
  }

  // Check worker health
  for (const [name, stats] of workerStats) {
    if (stats.lastJobAt) {
      const idleMs = Date.now() - stats.lastJobAt.getTime();
      if (idleMs > 10 * 60 * 1000) {  // 10 minutes idle
        issues.push(`Worker ${name}: idle for ${Math.round(idleMs / 60000)}min`);
      }
    }
  }

  const healthy = issues.length === 0;

  res.status(healthy ? 200 : 503).json({
    status: healthy ? 'healthy' : 'degraded',
    issues,
    queues: metrics,
    workers: Object.fromEntries(workerStats),
    timestamp: new Date().toISOString(),
  });
});
```

---

## Dead Letter Queue Management

### DLQ Setup

```typescript
// src/queues/dlq.ts
import { Queue, QueueEvents, Worker } from 'bullmq';
import { connection } from '../lib/redis';

// DLQ is just another queue
export const dlq = new Queue('dead-letter', {
  connection,
  defaultJobOptions: {
    removeOnComplete: false,   // Keep all DLQ items
    removeOnFail: false,
  },
});

// Move failed jobs to DLQ after all retries exhausted
export function setupDLQ(sourceQueue: Queue) {
  const events = new QueueEvents(sourceQueue.name, { connection });

  events.on('failed', async ({ jobId, failedReason, prev }) => {
    if (prev !== 'active') return;  // Only when transitioning from active

    const job = await sourceQueue.getJob(jobId);
    if (!job) return;

    // Check if all retries exhausted
    if (job.attemptsMade >= (job.opts.attempts || 1)) {
      await dlq.add(`dlq-${sourceQueue.name}`, {
        originalQueue: sourceQueue.name,
        originalJobId: jobId,
        originalJobName: job.name,
        originalData: job.data,
        failedReason,
        attemptsMade: job.attemptsMade,
        failedAt: new Date().toISOString(),
      });

      console.error(`Job ${jobId} moved to DLQ after ${job.attemptsMade} attempts: ${failedReason}`);
    }
  });
}

// Initialize for all queues
setupDLQ(emailQueue);
setupDLQ(imageQueue);
setupDLQ(webhookQueue);
```

### DLQ Admin API

```typescript
// src/routes/admin/dlq.ts

// List DLQ items
router.get('/admin/dlq', requireAdmin, async (req, res) => {
  const page = parseInt(req.query.page as string) || 0;
  const pageSize = 20;

  const jobs = await dlq.getJobs(['waiting', 'failed'], page * pageSize, (page + 1) * pageSize - 1);

  res.json({
    items: jobs.map(job => ({
      id: job.id,
      originalQueue: job.data.originalQueue,
      originalJobName: job.data.originalJobName,
      failedReason: job.data.failedReason,
      attemptsMade: job.data.attemptsMade,
      failedAt: job.data.failedAt,
      data: job.data.originalData,
    })),
    total: await dlq.getJobCounts('waiting', 'failed'),
  });
});

// Retry a DLQ item (re-enqueue in original queue)
router.post('/admin/dlq/:jobId/retry', requireAdmin, async (req, res) => {
  const job = await dlq.getJob(req.params.jobId);
  if (!job) return res.status(404).json({ error: 'DLQ item not found' });

  const originalQueue = new Queue(job.data.originalQueue, { connection });

  await originalQueue.add(job.data.originalJobName, job.data.originalData, {
    attempts: 3,  // Fresh retry attempts
  });

  await job.remove();

  res.json({ success: true, message: `Re-enqueued to ${job.data.originalQueue}` });
});

// Dismiss a DLQ item
router.delete('/admin/dlq/:jobId', requireAdmin, async (req, res) => {
  const job = await dlq.getJob(req.params.jobId);
  if (!job) return res.status(404).json({ error: 'DLQ item not found' });

  await job.remove();
  res.json({ success: true });
});

// Retry all DLQ items for a specific queue
router.post('/admin/dlq/retry-all', requireAdmin, async (req, res) => {
  const { queue: queueName } = req.body;
  const jobs = await dlq.getJobs(['waiting'], 0, 1000);

  const matching = jobs.filter(j => j.data.originalQueue === queueName);
  const originalQueue = new Queue(queueName, { connection });

  for (const job of matching) {
    await originalQueue.add(job.data.originalJobName, job.data.originalData);
    await job.remove();
  }

  res.json({ success: true, retried: matching.length });
});
```

---

## Alerting Patterns

### Slack Alerts for Queue Issues

```typescript
// src/monitoring/alerts.ts

interface AlertConfig {
  queueName: string;
  maxWaiting: number;      // Alert when waiting > this
  maxFailed: number;       // Alert when failed > this
  maxIdleMinutes: number;  // Alert when no jobs processed for this long
}

const alertConfigs: AlertConfig[] = [
  { queueName: 'email', maxWaiting: 500, maxFailed: 50, maxIdleMinutes: 30 },
  { queueName: 'webhooks', maxWaiting: 200, maxFailed: 20, maxIdleMinutes: 15 },
  { queueName: 'image-processing', maxWaiting: 100, maxFailed: 10, maxIdleMinutes: 60 },
];

// Run every 5 minutes
export async function checkAlerts() {
  for (const config of alertConfigs) {
    const queue = new Queue(config.queueName, { connection });
    const counts = await queue.getJobCounts('waiting', 'failed');

    if (counts.waiting > config.maxWaiting) {
      await sendSlackAlert(
        `Queue \`${config.queueName}\` has ${counts.waiting} waiting jobs (threshold: ${config.maxWaiting})`
      );
    }

    if (counts.failed > config.maxFailed) {
      await sendSlackAlert(
        `Queue \`${config.queueName}\` has ${counts.failed} failed jobs (threshold: ${config.maxFailed})`
      );
    }
  }
}

async function sendSlackAlert(message: string) {
  const webhookUrl = process.env.SLACK_WEBHOOK_URL;
  if (!webhookUrl) return;

  await fetch(webhookUrl, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      text: `🚨 Queue Alert: ${message}`,
    }),
  });
}
```

---

## Queue Cleanup Automation

```typescript
// src/jobs/queue-cleanup.ts
import { Queue } from 'bullmq';
import { connection } from '../lib/redis';

const allQueues = ['email', 'image-processing', 'webhooks', 'exports', 'scheduled-tasks'];

export async function cleanupQueues() {
  for (const name of allQueues) {
    const queue = new Queue(name, { connection });

    // Remove completed jobs older than 24 hours
    const completed = await queue.clean(24 * 60 * 60 * 1000, 1000, 'completed');

    // Remove failed jobs older than 7 days
    const failed = await queue.clean(7 * 24 * 60 * 60 * 1000, 1000, 'failed');

    if (completed.length > 0 || failed.length > 0) {
      console.log(`${name}: cleaned ${completed.length} completed, ${failed.length} failed`);
    }
  }
}

// Run via cron: schedule('0 3 * * *', cleanupQueues)
```

---

## Dashboard JSON API

```typescript
// GET /admin/queues/api — JSON summary for custom dashboards
router.get('/admin/queues/api', requireAdmin, async (req, res) => {
  const metrics = await collectQueueMetrics();
  const workerStats = getWorkerStats();

  // Get recent failures for each queue
  const recentFailures = await Promise.all(
    metrics.map(async (m) => {
      const queue = new Queue(m.name, { connection });
      const failed = await queue.getJobs(['failed'], 0, 4);
      return {
        queue: m.name,
        failures: failed.map(j => ({
          id: j.id,
          name: j.name,
          failedReason: j.failedReason,
          attemptsMade: j.attemptsMade,
          timestamp: j.timestamp,
        })),
      };
    })
  );

  res.json({
    summary: {
      totalWaiting: metrics.reduce((sum, m) => sum + m.waiting, 0),
      totalActive: metrics.reduce((sum, m) => sum + m.active, 0),
      totalFailed: metrics.reduce((sum, m) => sum + m.failed, 0),
    },
    queues: metrics,
    workers: Object.fromEntries(workerStats),
    recentFailures,
    timestamp: new Date().toISOString(),
  });
});
```
