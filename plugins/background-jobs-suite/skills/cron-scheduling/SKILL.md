---
name: cron-scheduling
description: >
  Set up cron-based job scheduling — node-cron, BullMQ repeatables,
  Heroku Scheduler, and serverless cron patterns. Database cleanup,
  report generation, data sync, and recurring task automation.
  Triggers: "cron job", "scheduled task", "recurring job", "periodic",
  "run every", "daily task", "weekly job", "scheduler".
  NOT for: one-time delayed jobs (use bullmq-setup), real-time events.
version: 1.0.0
allowed-tools: Read, Grep, Glob, Bash, Edit, Write
---

# Cron & Scheduled Job Patterns

## Cron Expression Reference

```
┌───────────── minute (0-59)
│ ┌───────────── hour (0-23)
│ │ ┌───────────── day of month (1-31)
│ │ │ ┌───────────── month (1-12)
│ │ │ │ ┌───────────── day of week (0-7, 0 and 7 = Sunday)
│ │ │ │ │
* * * * *
```

### Common Patterns

| Pattern | Schedule | Use Case |
|---------|----------|----------|
| `* * * * *` | Every minute | Health checks, queue monitoring |
| `*/5 * * * *` | Every 5 minutes | Polling, lightweight sync |
| `0 * * * *` | Every hour | Data aggregation, cache refresh |
| `0 */6 * * *` | Every 6 hours | Report generation |
| `0 2 * * *` | Daily at 2 AM | Cleanup, backups |
| `0 9 * * 1-5` | Weekdays at 9 AM | Business reports |
| `0 0 * * 0` | Weekly (Sunday midnight) | Analytics rollup |
| `0 0 1 * *` | Monthly (1st at midnight) | Billing, invoicing |
| `0 0 1 1 *` | Yearly (Jan 1 midnight) | Annual archival |
| `30 8 * * 1` | Monday 8:30 AM | Weekly digest email |

### Special Characters

| Char | Meaning | Example |
|------|---------|---------|
| `*` | Any value | `* * * * *` = every minute |
| `,` | List | `1,15 * * * *` = minute 1 and 15 |
| `-` | Range | `0 9-17 * * *` = every hour 9 AM to 5 PM |
| `/` | Step | `*/10 * * * *` = every 10 minutes |

---

## Option 1: node-cron (In-Process Scheduler)

Best for: Simple apps, single-server deployments, development.

```bash
npm install node-cron
```

```typescript
// src/scheduler.ts
import cron from 'node-cron';

// Validate expression before scheduling
if (!cron.validate('0 2 * * *')) {
  throw new Error('Invalid cron expression');
}

// Daily cleanup at 2 AM
const cleanupJob = cron.schedule('0 2 * * *', async () => {
  console.log('Running daily cleanup...');
  try {
    const deleted = await db.session.deleteMany({
      where: { expiresAt: { lt: new Date() } },
    });
    console.log(`Cleaned up ${deleted.count} expired sessions`);
  } catch (error) {
    console.error('Cleanup failed:', error);
    // Don't throw — node-cron swallows errors silently
    alertOps('cleanup-failed', error);
  }
}, {
  timezone: 'America/New_York',
  scheduled: true,  // Start immediately (default)
});

// Hourly stats aggregation
cron.schedule('0 * * * *', async () => {
  await aggregateHourlyStats();
});

// Every 5 minutes — check for stuck jobs
cron.schedule('*/5 * * * *', async () => {
  const stuck = await db.job.findMany({
    where: {
      status: 'processing',
      updatedAt: { lt: new Date(Date.now() - 30 * 60 * 1000) },
    },
  });
  if (stuck.length > 0) {
    console.warn(`Found ${stuck.length} stuck jobs, resetting...`);
    await db.job.updateMany({
      where: { id: { in: stuck.map(j => j.id) } },
      data: { status: 'pending', attempts: { increment: 1 } },
    });
  }
});

// Stop a job
cleanupJob.stop();

// Graceful shutdown
process.on('SIGTERM', () => {
  cleanupJob.stop();
  process.exit(0);
});
```

### node-cron Gotchas

1. **In-process only** — if your app restarts, schedules reset. No persistence.
2. **Single server** — running 3 replicas = job runs 3 times. Use a distributed lock.
3. **Silent failures** — errors in handlers don't crash the process. Always add error handling.
4. **No job history** — no record of what ran or failed. Add your own logging.

---

## Option 2: BullMQ Repeatable Jobs (Distributed)

Best for: Multi-server deployments, production apps with Redis.

```typescript
// src/schedulers/setup-schedules.ts
import { Queue } from 'bullmq';
import { connection } from '../lib/redis';

const schedulerQueue = new Queue('scheduled-tasks', { connection });

export async function setupSchedules() {
  // Remove old schedules first (prevents duplicates on restart)
  const existing = await schedulerQueue.getRepeatableJobs();
  for (const job of existing) {
    await schedulerQueue.removeRepeatableByKey(job.key);
  }

  // Daily cleanup at 2 AM ET
  await schedulerQueue.add('cleanup-expired', {}, {
    repeat: { pattern: '0 2 * * *', tz: 'America/New_York' },
    jobId: 'cleanup-expired',  // Prevents duplicates
  });

  // Hourly stats
  await schedulerQueue.add('aggregate-stats', {}, {
    repeat: { pattern: '0 * * * *' },
    jobId: 'aggregate-stats',
  });

  // Every 5 minutes — health check
  await schedulerQueue.add('health-check', {}, {
    repeat: { every: 5 * 60 * 1000 },
    jobId: 'health-check',
  });

  // Weekly digest every Monday at 8:30 AM
  await schedulerQueue.add('weekly-digest', {}, {
    repeat: { pattern: '30 8 * * 1', tz: 'America/New_York' },
    jobId: 'weekly-digest',
  });

  console.log('Schedules configured');
}
```

```typescript
// src/workers/scheduler.worker.ts
import { Worker, Job } from 'bullmq';
import { connection } from '../lib/redis';

const schedulerWorker = new Worker('scheduled-tasks', async (job: Job) => {
  switch (job.name) {
    case 'cleanup-expired':
      return await cleanupExpiredSessions();
    case 'aggregate-stats':
      return await aggregateHourlyStats();
    case 'health-check':
      return await runHealthChecks();
    case 'weekly-digest':
      return await sendWeeklyDigest();
    default:
      console.warn(`Unknown scheduled job: ${job.name}`);
  }
}, { connection, concurrency: 1 });

// Call setupSchedules() on app startup
```

### BullMQ vs node-cron

| Feature | node-cron | BullMQ Repeatable |
|---------|-----------|-------------------|
| Persistence | No | Yes (Redis) |
| Multi-server safe | No | Yes (distributed lock) |
| Job history | No | Yes |
| Retry on failure | No | Yes |
| Monitoring UI | No | Yes (Bull Board) |
| Setup complexity | Low | Medium |
| Dependencies | None | Redis |

---

## Option 3: External Schedulers

### Heroku Scheduler

```bash
heroku addons:create scheduler:standard
heroku addons:open scheduler
# Configure via web UI — runs as one-off dyno
```

```json
// package.json
{
  "scripts": {
    "job:cleanup": "node dist/jobs/cleanup.js",
    "job:digest": "node dist/jobs/weekly-digest.js"
  }
}
```

```typescript
// src/jobs/cleanup.ts — standalone script
import { db } from '../lib/db';

async function main() {
  console.log('Starting cleanup job...');
  const result = await db.session.deleteMany({
    where: { expiresAt: { lt: new Date() } },
  });
  console.log(`Deleted ${result.count} expired sessions`);
  process.exit(0);
}

main().catch((err) => {
  console.error('Cleanup failed:', err);
  process.exit(1);
});
```

### GitHub Actions (Free Cron)

```yaml
# .github/workflows/daily-check.yml
name: Daily Health Check
on:
  schedule:
    - cron: '0 14 * * *'  # 2 PM UTC = 9 AM ET
  workflow_dispatch: {}     # Also allow manual trigger

jobs:
  health-check:
    runs-on: ubuntu-latest
    steps:
      - name: Check API health
        run: |
          STATUS=$(curl -s -o /dev/null -w "%{http_code}" https://api.example.com/health)
          if [ "$STATUS" != "200" ]; then
            echo "Health check failed: HTTP $STATUS"
            exit 1
          fi
```

### Vercel Cron

```json
// vercel.json
{
  "crons": [
    {
      "path": "/api/cron/cleanup",
      "schedule": "0 2 * * *"
    }
  ]
}
```

```typescript
// app/api/cron/cleanup/route.ts
import { NextResponse } from 'next/server';

export const runtime = 'nodejs';

export async function GET(request: Request) {
  // Verify the request is from Vercel Cron
  const authHeader = request.headers.get('authorization');
  if (authHeader !== `Bearer ${process.env.CRON_SECRET}`) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
  }

  await cleanupExpiredData();
  return NextResponse.json({ success: true });
}
```

---

## Distributed Lock Pattern

Prevent duplicate execution across multiple servers:

```typescript
// src/lib/distributed-lock.ts
import IORedis from 'ioredis';

const redis = new IORedis(process.env.REDIS_URL);

export async function withLock<T>(
  lockKey: string,
  ttlMs: number,
  fn: () => Promise<T>
): Promise<T | null> {
  const lockValue = `${process.pid}-${Date.now()}`;

  // Try to acquire lock (SET NX EX)
  const acquired = await redis.set(
    `lock:${lockKey}`,
    lockValue,
    'PX', ttlMs,
    'NX'
  );

  if (!acquired) {
    console.log(`Lock ${lockKey} already held, skipping`);
    return null;
  }

  try {
    return await fn();
  } finally {
    // Release lock (only if we still hold it)
    const current = await redis.get(`lock:${lockKey}`);
    if (current === lockValue) {
      await redis.del(`lock:${lockKey}`);
    }
  }
}

// Usage with node-cron
cron.schedule('0 2 * * *', async () => {
  await withLock('daily-cleanup', 60_000, async () => {
    await cleanupExpiredSessions();
  });
});
```

---

## Monitoring Scheduled Jobs

### Simple Logging Table

```typescript
// src/lib/job-log.ts
// Prisma schema:
// model JobLog {
//   id        String   @id @default(cuid())
//   name      String
//   status    String   // 'started' | 'completed' | 'failed'
//   duration  Int?     // milliseconds
//   error     String?
//   result    Json?
//   createdAt DateTime @default(now())
// }

export async function trackJob<T>(name: string, fn: () => Promise<T>): Promise<T> {
  const start = Date.now();
  await db.jobLog.create({ data: { name, status: 'started' } });

  try {
    const result = await fn();
    await db.jobLog.create({
      data: {
        name,
        status: 'completed',
        duration: Date.now() - start,
        result: result as any,
      },
    });
    return result;
  } catch (error: any) {
    await db.jobLog.create({
      data: {
        name,
        status: 'failed',
        duration: Date.now() - start,
        error: error.message,
      },
    });
    throw error;
  }
}

// Usage
cron.schedule('0 2 * * *', () => {
  trackJob('daily-cleanup', cleanupExpiredSessions);
});
```

### Health Check Endpoint

```typescript
// GET /api/cron/status
router.get('/cron/status', async (req, res) => {
  const jobs = ['daily-cleanup', 'hourly-stats', 'health-check'];

  const statuses = await Promise.all(
    jobs.map(async (name) => {
      const lastRun = await db.jobLog.findFirst({
        where: { name },
        orderBy: { createdAt: 'desc' },
      });
      const lastSuccess = await db.jobLog.findFirst({
        where: { name, status: 'completed' },
        orderBy: { createdAt: 'desc' },
      });
      return {
        name,
        lastRun: lastRun?.createdAt,
        lastStatus: lastRun?.status,
        lastSuccess: lastSuccess?.createdAt,
        lastDuration: lastSuccess?.duration,
      };
    })
  );

  res.json(statuses);
});
```
