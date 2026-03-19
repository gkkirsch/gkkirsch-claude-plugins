---
name: job-patterns
description: >
  Production job processing patterns — email queues, image processing
  pipelines, webhook delivery, CSV export, data sync, and notification
  fanout. Complete implementations with error handling and monitoring.
  Triggers: "email queue", "image processing pipeline", "webhook delivery",
  "CSV export job", "data sync worker", "notification system", "job pipeline".
  NOT for: basic BullMQ setup (use bullmq-setup), cron scheduling (use cron-scheduling).
version: 1.0.0
allowed-tools: Read, Grep, Glob, Bash, Edit, Write
---

# Production Job Processing Patterns

## Pattern 1: Email Queue with Templates

```typescript
// src/queues/email.queue.ts
import { Queue, Worker, Job } from 'bullmq';
import { connection } from '../lib/redis';
import { Resend } from 'resend';

const resend = new Resend(process.env.RESEND_API_KEY);

interface EmailJob {
  to: string | string[];
  subject: string;
  template: 'welcome' | 'reset-password' | 'invoice' | 'notification';
  variables: Record<string, string>;
  replyTo?: string;
  attachments?: Array<{ filename: string; content: string }>;
}

export const emailQueue = new Queue<EmailJob>('email', {
  connection,
  defaultJobOptions: {
    attempts: 3,
    backoff: { type: 'exponential', delay: 2000 },
    removeOnComplete: { count: 500, age: 86400 },
    removeOnFail: { count: 2000, age: 604800 },
  },
});

// Template renderer
function renderTemplate(template: string, vars: Record<string, string>): string {
  const templates: Record<string, string> = {
    welcome: `<h1>Welcome, ${vars.name}!</h1><p>Thanks for signing up.</p>`,
    'reset-password': `<p>Click <a href="${vars.resetUrl}">here</a> to reset your password.</p>`,
    invoice: `<h2>Invoice #${vars.invoiceNumber}</h2><p>Amount: $${vars.amount}</p>`,
    notification: `<p>${vars.message}</p>`,
  };
  return templates[template] || vars.message || '';
}

const emailWorker = new Worker<EmailJob>('email', async (job: Job<EmailJob>) => {
  const { to, subject, template, variables, replyTo, attachments } = job.data;

  const html = renderTemplate(template, variables);

  const result = await resend.emails.send({
    from: process.env.FROM_EMAIL || 'noreply@example.com',
    to: Array.isArray(to) ? to : [to],
    subject,
    html,
    replyTo,
    attachments: attachments?.map(a => ({
      filename: a.filename,
      content: Buffer.from(a.content, 'base64'),
    })),
  });

  return { messageId: result.data?.id, sentAt: new Date().toISOString() };
}, {
  connection,
  concurrency: 5,
  limiter: { max: 50, duration: 60_000 },  // 50 emails per minute
});

// Helper functions
export async function sendWelcomeEmail(user: { email: string; name: string }) {
  return emailQueue.add('welcome', {
    to: user.email,
    subject: `Welcome to the platform, ${user.name}!`,
    template: 'welcome',
    variables: { name: user.name },
  }, {
    priority: 2,
    jobId: `welcome-${user.email}`,
  });
}

export async function sendPasswordReset(email: string, resetUrl: string) {
  return emailQueue.add('reset-password', {
    to: email,
    subject: 'Password Reset Request',
    template: 'reset-password',
    variables: { resetUrl },
  }, {
    priority: 1,  // High priority — user is waiting
  });
}
```

---

## Pattern 2: Image Processing Pipeline

```typescript
// src/queues/image.queue.ts
import { Queue, Worker, Job, FlowProducer } from 'bullmq';
import sharp from 'sharp';
import { S3Client, PutObjectCommand, GetObjectCommand } from '@aws-sdk/client-s3';
import { connection } from '../lib/redis';

interface ImageJob {
  imageId: string;
  originalKey: string;
  bucket: string;
  variants: Array<{
    name: string;
    width: number;
    height?: number;
    format: 'webp' | 'avif' | 'jpeg';
    quality: number;
  }>;
}

const s3 = new S3Client({ region: process.env.AWS_REGION });

export const imageQueue = new Queue<ImageJob>('image-processing', {
  connection,
  defaultJobOptions: {
    attempts: 2,
    backoff: { type: 'fixed', delay: 5000 },
    removeOnComplete: { count: 200 },
    removeOnFail: { count: 500 },
  },
});

const imageWorker = new Worker<ImageJob>('image-processing', async (job: Job<ImageJob>) => {
  const { imageId, originalKey, bucket, variants } = job.data;

  // 1. Download original
  await job.updateProgress(10);
  const original = await s3.send(new GetObjectCommand({
    Bucket: bucket,
    Key: originalKey,
  }));
  const buffer = Buffer.from(await original.Body!.transformToByteArray());

  // 2. Generate variants
  const results: Record<string, string> = {};

  for (let i = 0; i < variants.length; i++) {
    const variant = variants[i];
    const outputKey = `processed/${imageId}/${variant.name}.${variant.format}`;

    let pipeline = sharp(buffer).resize(variant.width, variant.height, {
      fit: 'cover',
      withoutEnlargement: true,
    });

    // Apply format conversion
    switch (variant.format) {
      case 'webp':
        pipeline = pipeline.webp({ quality: variant.quality });
        break;
      case 'avif':
        pipeline = pipeline.avif({ quality: variant.quality });
        break;
      case 'jpeg':
        pipeline = pipeline.jpeg({ quality: variant.quality, mozjpeg: true });
        break;
    }

    // Strip EXIF but keep orientation
    pipeline = pipeline.rotate();  // Auto-rotate from EXIF then strip

    const processed = await pipeline.toBuffer();

    // 3. Upload variant
    await s3.send(new PutObjectCommand({
      Bucket: bucket,
      Key: outputKey,
      Body: processed,
      ContentType: `image/${variant.format}`,
      CacheControl: 'public, max-age=31536000, immutable',
    }));

    results[variant.name] = outputKey;
    await job.updateProgress(10 + Math.round((i + 1) / variants.length * 80));
  }

  // 4. Update database
  await db.image.update({
    where: { id: imageId },
    data: {
      status: 'READY',
      variants: results,
      processedAt: new Date(),
    },
  });

  await job.updateProgress(100);
  return results;
}, {
  connection,
  concurrency: 2,  // CPU-bound — keep low
});

// Trigger from upload handler
export async function processUploadedImage(imageId: string, s3Key: string) {
  return imageQueue.add('process-image', {
    imageId,
    originalKey: s3Key,
    bucket: process.env.S3_BUCKET!,
    variants: [
      { name: 'thumb', width: 200, format: 'webp', quality: 75 },
      { name: 'small', width: 400, format: 'webp', quality: 80 },
      { name: 'medium', width: 800, format: 'webp', quality: 82 },
      { name: 'large', width: 1600, format: 'webp', quality: 85 },
      { name: 'og', width: 1200, height: 630, format: 'jpeg', quality: 80 },
    ],
  });
}
```

---

## Pattern 3: Webhook Delivery with Retry

```typescript
// src/queues/webhook.queue.ts
import { Queue, Worker, Job, UnrecoverableError } from 'bullmq';
import { connection } from '../lib/redis';
import crypto from 'crypto';

interface WebhookJob {
  webhookId: string;
  url: string;
  event: string;
  payload: Record<string, unknown>;
  secret: string;
}

export const webhookQueue = new Queue<WebhookJob>('webhooks', {
  connection,
  defaultJobOptions: {
    attempts: 5,
    backoff: {
      type: 'custom',
    },
    removeOnComplete: { count: 1000 },
    removeOnFail: { count: 5000 },
  },
});

const webhookWorker = new Worker<WebhookJob>('webhooks', async (job: Job<WebhookJob>) => {
  const { webhookId, url, event, payload, secret } = job.data;

  // Sign the payload
  const body = JSON.stringify(payload);
  const timestamp = Math.floor(Date.now() / 1000);
  const signature = crypto
    .createHmac('sha256', secret)
    .update(`${timestamp}.${body}`)
    .digest('hex');

  // Deliver
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), 10_000);  // 10s timeout

  try {
    const response = await fetch(url, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-Webhook-Event': event,
        'X-Webhook-Signature': `t=${timestamp},v1=${signature}`,
        'X-Webhook-Id': webhookId,
        'User-Agent': 'MyApp-Webhooks/1.0',
      },
      body,
      signal: controller.signal,
    });

    clearTimeout(timeout);

    // 2xx = success
    if (response.ok) {
      return { status: response.status, deliveredAt: new Date().toISOString() };
    }

    // 4xx (except 429) = permanent failure, don't retry
    if (response.status >= 400 && response.status < 500 && response.status !== 429) {
      throw new UnrecoverableError(
        `Webhook rejected: HTTP ${response.status} ${response.statusText}`
      );
    }

    // 429 or 5xx = transient, retry
    throw new Error(`Webhook failed: HTTP ${response.status}`);
  } catch (error: any) {
    clearTimeout(timeout);
    if (error.name === 'AbortError') {
      throw new Error('Webhook delivery timed out (10s)');
    }
    throw error;
  }
}, {
  connection,
  concurrency: 10,
  settings: {
    backoffStrategy: (attemptsMade: number) => {
      // 1 min, 5 min, 30 min, 2 hours, 12 hours
      const delays = [60_000, 300_000, 1_800_000, 7_200_000, 43_200_000];
      return delays[attemptsMade - 1] || delays[delays.length - 1];
    },
  },
});

// Log delivery attempts
webhookWorker.on('completed', async (job) => {
  await db.webhookDelivery.create({
    data: {
      webhookId: job.data.webhookId,
      event: job.data.event,
      status: 'delivered',
      attempt: job.attemptsMade,
      response: job.returnvalue,
    },
  });
});

webhookWorker.on('failed', async (job, error) => {
  if (job) {
    await db.webhookDelivery.create({
      data: {
        webhookId: job.data.webhookId,
        event: job.data.event,
        status: job.attemptsMade >= (job.opts.attempts || 5) ? 'failed' : 'retrying',
        attempt: job.attemptsMade,
        error: error.message,
      },
    });
  }
});

// Public API to send webhooks
export async function dispatchWebhook(event: string, payload: Record<string, unknown>) {
  const webhooks = await db.webhook.findMany({
    where: { events: { has: event }, active: true },
  });

  const jobs = webhooks.map(wh => ({
    name: event,
    data: {
      webhookId: wh.id,
      url: wh.url,
      event,
      payload: { ...payload, event, timestamp: new Date().toISOString() },
      secret: wh.secret,
    },
  }));

  if (jobs.length > 0) {
    await webhookQueue.addBulk(jobs);
  }
}
```

---

## Pattern 4: CSV Export with Streaming

```typescript
// src/queues/export.queue.ts
import { Queue, Worker, Job } from 'bullmq';
import { connection } from '../lib/redis';
import { createObjectCsvStringifier } from 'csv-writer';
import { S3Client, PutObjectCommand } from '@aws-sdk/client-s3';
import { getSignedUrl } from '@aws-sdk/s3-request-presigner';

interface ExportJob {
  exportId: string;
  userId: string;
  type: 'users' | 'orders' | 'transactions';
  filters: Record<string, unknown>;
}

export const exportQueue = new Queue<ExportJob>('exports', {
  connection,
  defaultJobOptions: {
    attempts: 2,
    backoff: { type: 'fixed', delay: 10_000 },
    removeOnComplete: { count: 100 },
    removeOnFail: { count: 200 },
  },
});

const BATCH_SIZE = 1000;

const exportWorker = new Worker<ExportJob>('exports', async (job: Job<ExportJob>) => {
  const { exportId, userId, type, filters } = job.data;
  const s3 = new S3Client({ region: process.env.AWS_REGION });

  // Build query based on type
  const queryFn = getQueryForType(type, filters);

  // Count total rows for progress
  const total = await queryFn.count();
  await job.updateProgress(5);

  if (total === 0) {
    throw new Error('No data matches the export criteria');
  }

  // Stream rows in batches to build CSV
  const csvParts: string[] = [];
  let processed = 0;
  let cursor: string | undefined;

  const csvWriter = createObjectCsvStringifier({
    header: getHeadersForType(type),
  });

  // Add CSV header
  csvParts.push(csvWriter.getHeaderString()!);

  while (processed < total) {
    const batch = await queryFn.findMany({
      take: BATCH_SIZE,
      ...(cursor ? { skip: 1, cursor: { id: cursor } } : {}),
      orderBy: { id: 'asc' },
    });

    if (batch.length === 0) break;

    csvParts.push(csvWriter.stringifyRecords(batch));
    processed += batch.length;
    cursor = batch[batch.length - 1].id;

    await job.updateProgress(5 + Math.round((processed / total) * 85));
  }

  const csvContent = csvParts.join('');

  // Upload to S3
  const s3Key = `exports/${userId}/${exportId}.csv`;
  await s3.send(new PutObjectCommand({
    Bucket: process.env.S3_BUCKET!,
    Key: s3Key,
    Body: csvContent,
    ContentType: 'text/csv',
    ContentDisposition: `attachment; filename="${type}-export-${new Date().toISOString().slice(0, 10)}.csv"`,
  }));

  // Generate download URL (1 hour expiry)
  const downloadUrl = await getSignedUrl(
    s3,
    new GetObjectCommand({ Bucket: process.env.S3_BUCKET!, Key: s3Key }),
    { expiresIn: 3600 }
  );

  // Update export record
  await db.export.update({
    where: { id: exportId },
    data: {
      status: 'completed',
      rowCount: processed,
      fileUrl: downloadUrl,
      completedAt: new Date(),
    },
  });

  await job.updateProgress(100);

  // Notify user
  await emailQueue.add('export-ready', {
    to: (await db.user.findUnique({ where: { id: userId } }))!.email,
    subject: `Your ${type} export is ready`,
    template: 'notification',
    variables: {
      message: `Your export of ${processed} ${type} is ready. <a href="${downloadUrl}">Download CSV</a> (link expires in 1 hour).`,
    },
  });

  return { rowCount: processed, downloadUrl };
}, {
  connection,
  concurrency: 3,
});
```

---

## Pattern 5: Data Sync Worker

```typescript
// src/queues/sync.queue.ts
import { Queue, Worker, Job, UnrecoverableError } from 'bullmq';
import { connection } from '../lib/redis';

interface SyncJob {
  provider: 'stripe' | 'hubspot' | 'salesforce';
  operation: 'full' | 'incremental';
  since?: string;  // ISO date for incremental
}

export const syncQueue = new Queue<SyncJob>('data-sync', {
  connection,
  defaultJobOptions: {
    attempts: 3,
    backoff: { type: 'exponential', delay: 30_000 },
    removeOnComplete: { count: 50 },
    removeOnFail: { count: 100 },
  },
});

const syncWorker = new Worker<SyncJob>('data-sync', async (job: Job<SyncJob>) => {
  const { provider, operation, since } = job.data;

  const syncer = getSyncer(provider);
  let totalSynced = 0;
  let cursor: string | undefined;

  const startFrom = operation === 'incremental' && since
    ? new Date(since)
    : new Date(0);

  // Paginate through external API
  while (true) {
    let page;
    try {
      page = await syncer.fetchPage(cursor, startFrom);
    } catch (error: any) {
      if (error.status === 401) {
        throw new UnrecoverableError(`${provider} auth expired — reauthorize`);
      }
      throw error;  // Transient — retry
    }

    // Upsert into local database
    for (const record of page.data) {
      await db[provider + 'Record'].upsert({
        where: { externalId: record.id },
        create: {
          externalId: record.id,
          data: record,
          syncedAt: new Date(),
        },
        update: {
          data: record,
          syncedAt: new Date(),
        },
      });
    }

    totalSynced += page.data.length;
    cursor = page.nextCursor;

    await job.updateProgress(
      page.nextCursor ? Math.min(90, totalSynced) : 100
    );

    if (!page.nextCursor) break;

    // Respect rate limits
    await new Promise(r => setTimeout(r, 200));
  }

  // Record sync completion
  await db.syncLog.create({
    data: {
      provider,
      operation,
      recordCount: totalSynced,
      completedAt: new Date(),
    },
  });

  return { provider, totalSynced };
}, {
  connection,
  concurrency: 1,  // One sync at a time per queue
  limiter: { max: 1, duration: 60_000 },
});
```

---

## Pattern 6: Notification Fanout

```typescript
// src/queues/notification.queue.ts
import { Queue, Worker, Job, FlowProducer } from 'bullmq';
import { connection } from '../lib/redis';

interface NotificationJob {
  userId: string;
  type: 'order_shipped' | 'payment_received' | 'mention' | 'system';
  title: string;
  body: string;
  data?: Record<string, string>;
}

// One queue per channel
const pushQueue = new Queue('push-notifications', { connection });
const inAppQueue = new Queue('in-app-notifications', { connection });

export async function notify(notification: NotificationJob) {
  const { userId } = notification;

  // Check user preferences
  const prefs = await db.notificationPreference.findUnique({
    where: { userId },
  });

  const jobs: Promise<any>[] = [];

  // Always create in-app notification
  jobs.push(inAppQueue.add('in-app', notification));

  // Push notification if enabled
  if (prefs?.pushEnabled !== false) {
    jobs.push(pushQueue.add('push', notification));
  }

  // Email for important notifications
  if (prefs?.emailEnabled !== false && ['order_shipped', 'payment_received'].includes(notification.type)) {
    const user = await db.user.findUnique({ where: { id: userId } });
    if (user?.email) {
      jobs.push(emailQueue.add('notification-email', {
        to: user.email,
        subject: notification.title,
        template: 'notification',
        variables: { message: notification.body },
      }));
    }
  }

  await Promise.all(jobs);
}

// In-app notification worker (fast — just DB write)
new Worker('in-app-notifications', async (job: Job<NotificationJob>) => {
  await db.notification.create({
    data: {
      userId: job.data.userId,
      type: job.data.type,
      title: job.data.title,
      body: job.data.body,
      data: job.data.data || {},
      read: false,
    },
  });
}, { connection, concurrency: 20 });

// Push notification worker
new Worker('push-notifications', async (job: Job<NotificationJob>) => {
  const tokens = await db.pushToken.findMany({
    where: { userId: job.data.userId },
  });

  for (const token of tokens) {
    try {
      await sendPushNotification(token.token, {
        title: job.data.title,
        body: job.data.body,
        data: job.data.data,
      });
    } catch (error: any) {
      if (error.code === 'INVALID_TOKEN') {
        // Token expired — remove it
        await db.pushToken.delete({ where: { id: token.id } });
      }
    }
  }
}, { connection, concurrency: 5 });
```

---

## Anti-Patterns to Avoid

### 1. Non-Idempotent Jobs

```typescript
// BAD — running twice charges twice
await stripe.charges.create({ amount: 1000, customer: customerId });

// GOOD — idempotency key prevents double-charge
await stripe.charges.create({
  amount: 1000,
  customer: customerId,
}, {
  idempotencyKey: `charge-${orderId}`,
});
```

### 2. Huge Job Payloads

```typescript
// BAD — storing entire file in Redis
await queue.add('process', { file: hugeBuffer.toString('base64') });

// GOOD — store reference, fetch when processing
await queue.add('process', { fileId: 'abc123', bucket: 'uploads' });
```

### 3. Missing Timeout

```typescript
// BAD — job runs forever if API hangs
await fetch(externalApi);

// GOOD — abort after 30 seconds
const controller = new AbortController();
setTimeout(() => controller.abort(), 30_000);
await fetch(externalApi, { signal: controller.signal });
```

### 4. Catching and Swallowing Errors

```typescript
// BAD — job appears successful but didn't actually work
try { await processOrder(data); }
catch (e) { console.log('failed, oh well'); }

// GOOD — let BullMQ handle retry
await processOrder(data);  // Throws on failure → BullMQ retries
```
