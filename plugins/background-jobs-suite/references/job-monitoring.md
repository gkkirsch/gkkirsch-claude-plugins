# Job Monitoring & Observability Reference

Patterns for monitoring background job health, alerting on failures, and debugging queue issues.

---

## Key Metrics to Track

| Metric | What It Tells You | Alert Threshold |
|--------|-------------------|----------------|
| **Queue depth** (waiting) | Are workers keeping up? | > 1,000 for 5 min |
| **Active jobs** | Current parallelism | > concurrency setting |
| **Failed count** | Error rate | > 5% of completed |
| **Processing time** (p95) | Job duration trends | > 2x baseline |
| **Stalled count** | Worker crashes | > 0 |
| **DLQ depth** | Unrecoverable failures | > 0 |
| **Redis memory** | Data accumulation | > 80% of plan limit |
| **Worker uptime** | Process stability | Restart count > 3/hour |
| **Retry rate** | Transient failure frequency | > 20% of jobs retried |

## Bull Board Dashboard

The fastest way to get queue visibility:

```typescript
import { createBullBoard } from '@bull-board/api';
import { BullMQAdapter } from '@bull-board/api/bullMQAdapter';
import { ExpressAdapter } from '@bull-board/express';

const serverAdapter = new ExpressAdapter();
serverAdapter.setBasePath('/admin/queues');

createBullBoard({
  queues: [
    new BullMQAdapter(emailQueue),
    new BullMQAdapter(imageQueue),
    new BullMQAdapter(reportQueue),
  ],
  serverAdapter,
});

// Protect with auth middleware in production
app.use('/admin/queues', requireAdmin, serverAdapter.getRouter());
```

Bull Board gives you:
- Real-time queue status (waiting, active, completed, failed, delayed)
- Individual job inspection (data, result, logs, stack traces)
- Manual retry/remove/promote actions
- Job timeline view

## Custom Monitoring Endpoint

```typescript
// GET /api/admin/queue-health
app.get('/api/admin/queue-health', requireAdmin, async (req, res) => {
  const queues = [emailQueue, imageQueue, reportQueue];
  const health = await Promise.all(
    queues.map(async (queue) => {
      const [waiting, active, completed, failed, delayed] = await Promise.all([
        queue.getWaitingCount(),
        queue.getActiveCount(),
        queue.getCompletedCount(),
        queue.getFailedCount(),
        queue.getDelayedCount(),
      ]);

      const repeatableJobs = await queue.getRepeatableJobs();

      return {
        name: queue.name,
        counts: { waiting, active, completed, failed, delayed },
        repeatableJobs: repeatableJobs.length,
        healthy: waiting < 1000 && failed < 100,
      };
    })
  );

  const allHealthy = health.every((q) => q.healthy);
  res.status(allHealthy ? 200 : 503).json({
    status: allHealthy ? 'healthy' : 'degraded',
    queues: health,
    timestamp: new Date().toISOString(),
  });
});
```

## Structured Logging for Jobs

```typescript
// src/lib/job-logger.ts
interface JobLogEntry {
  queue: string;
  jobId: string;
  jobName: string;
  event: 'started' | 'completed' | 'failed' | 'retrying' | 'stalled';
  attempt?: number;
  duration?: number;
  error?: string;
  data?: Record<string, any>;
}

function logJob(entry: JobLogEntry) {
  // Structured JSON for log aggregation (Datadog, CloudWatch, etc.)
  console.log(JSON.stringify({
    level: entry.event === 'failed' ? 'error' : 'info',
    message: `job.${entry.event}`,
    ...entry,
    timestamp: new Date().toISOString(),
  }));
}

// Use in worker
const worker = new Worker('email', async (job) => {
  logJob({ queue: 'email', jobId: job.id!, jobName: job.name, event: 'started', attempt: job.attemptsMade + 1 });
  const start = Date.now();

  try {
    const result = await processEmail(job);
    logJob({ queue: 'email', jobId: job.id!, jobName: job.name, event: 'completed', duration: Date.now() - start });
    return result;
  } catch (err) {
    logJob({
      queue: 'email', jobId: job.id!, jobName: job.name,
      event: job.attemptsMade + 1 < (job.opts.attempts ?? 1) ? 'retrying' : 'failed',
      duration: Date.now() - start,
      error: err instanceof Error ? err.message : String(err),
      attempt: job.attemptsMade + 1,
    });
    throw err;
  }
}, { connection });
```

## Alerting Patterns

### Slack/Discord Webhook Alert

```typescript
async function alertOps(message: string, severity: 'info' | 'warning' | 'critical' = 'warning') {
  const emoji = { info: 'information_source', warning: 'warning', critical: 'rotating_light' };
  const webhookUrl = process.env.SLACK_WEBHOOK_URL;
  if (!webhookUrl) return;

  await fetch(webhookUrl, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      text: `:${emoji[severity]}: [${severity.toUpperCase()}] ${message}`,
    }),
  });
}

// Hook into worker events
worker.on('failed', (job, err) => {
  if (job && job.attemptsMade >= (job.opts.attempts ?? 1)) {
    alertOps(`Job ${job.name} [${job.id}] failed all ${job.attemptsMade} attempts: ${err.message}`, 'critical');
  }
});

worker.on('stalled', (jobId) => {
  alertOps(`Job ${jobId} stalled — possible worker crash`, 'warning');
});
```

### Queue Depth Polling

```typescript
setInterval(async () => {
  for (const queue of [emailQueue, imageQueue]) {
    const waiting = await queue.getWaitingCount();
    if (waiting > 1000) {
      alertOps(`Queue "${queue.name}" has ${waiting} waiting jobs — workers may need scaling`, 'warning');
    }
  }
}, 60_000); // Check every minute
```

## Debugging Failed Jobs

### Common Failure Patterns

| Symptom | Likely Cause | Fix |
|---------|-------------|-----|
| All jobs failing immediately | Worker code bug, bad config | Check worker logs, verify env vars |
| Jobs failing after timeout | Slow external API, resource exhaustion | Increase timeout, add circuit breaker |
| Intermittent failures | Network issues, rate limiting | Verify retry/backoff config |
| Jobs stuck in active state | Worker crashed mid-job, lock not released | Check stalled job recovery settings |
| Jobs pile up in waiting | Not enough workers, concurrency too low | Scale workers, increase concurrency |
| Redis OOM errors | Too many retained jobs, large payloads | Set `removeOnComplete`, reduce job data size |
| Duplicate job execution | Worker crash + restart, stalled recovery | Make handlers idempotent |

### Inspecting Failed Jobs

```typescript
// Get recent failed jobs
const failed = await queue.getFailed(0, 20); // offset, count
for (const job of failed) {
  console.log({
    id: job.id,
    name: job.name,
    data: job.data,
    failedReason: job.failedReason,
    stacktrace: job.stacktrace,
    attemptsMade: job.attemptsMade,
    timestamp: new Date(job.timestamp),
    processedOn: job.processedOn ? new Date(job.processedOn) : null,
    finishedOn: job.finishedOn ? new Date(job.finishedOn) : null,
  });
}

// Retry specific failed job
const job = await queue.getJob('failed-job-id');
if (job) await job.retry();

// Retry all failed jobs
const allFailed = await queue.getFailed();
await Promise.all(allFailed.map((job) => job.retry()));

// Clean old failed jobs
await queue.clean(7 * 24 * 3600 * 1000, 1000, 'failed'); // Older than 7 days
```

## Redis Health Checks

```bash
# Check Redis memory usage
redis-cli INFO memory | grep used_memory_human

# Check connected clients
redis-cli INFO clients | grep connected_clients

# List all BullMQ keys
redis-cli KEYS "bull:*" | head -20

# Count jobs in a queue
redis-cli XLEN "bull:email:id"

# Check queue-specific keys
redis-cli KEYS "bull:email:*" | wc -l
```

## Performance Baselines

Establish baselines for your specific workload, then alert on deviations:

| Metric | How to Measure | Baseline Example |
|--------|---------------|-----------------|
| Email send time | Worker processing duration | p50: 200ms, p95: 800ms, p99: 2s |
| Image processing | Worker processing duration | p50: 2s, p95: 8s, p99: 15s |
| Queue throughput | Jobs completed per minute | 500/min for email, 50/min for images |
| Redis latency | `redis-cli --latency` | < 1ms local, < 5ms remote |
| Worker memory | Process RSS | < 256MB per worker |

## Prometheus Metrics (Advanced)

```typescript
import { Registry, Counter, Histogram, Gauge } from 'prom-client';

const register = new Registry();

const jobsProcessed = new Counter({
  name: 'bullmq_jobs_processed_total',
  help: 'Total jobs processed',
  labelNames: ['queue', 'status'],
  registers: [register],
});

const jobDuration = new Histogram({
  name: 'bullmq_job_duration_seconds',
  help: 'Job processing duration',
  labelNames: ['queue', 'name'],
  buckets: [0.1, 0.5, 1, 2, 5, 10, 30, 60],
  registers: [register],
});

const queueDepth = new Gauge({
  name: 'bullmq_queue_depth',
  help: 'Current queue depth',
  labelNames: ['queue', 'state'],
  registers: [register],
});

// Expose /metrics endpoint
app.get('/metrics', async (req, res) => {
  // Update queue depth gauges
  for (const queue of [emailQueue, imageQueue]) {
    const [waiting, active] = await Promise.all([
      queue.getWaitingCount(),
      queue.getActiveCount(),
    ]);
    queueDepth.set({ queue: queue.name, state: 'waiting' }, waiting);
    queueDepth.set({ queue: queue.name, state: 'active' }, active);
  }

  res.set('Content-Type', register.contentType);
  res.end(await register.metrics());
});
```
