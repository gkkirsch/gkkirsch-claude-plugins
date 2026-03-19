---
name: email-queue
description: >
  Implement reliable email delivery with queuing, templates, and retry handling.
  Triggers: "email queue", "send emails in background", "email worker",
  "transactional email", "email delivery", "email templates", "nodemailer queue".
  NOT for: marketing email campaigns (use a service like Mailchimp/SendGrid marketing),
  general queue setup (use bullmq-setup).
version: 1.0.0
argument-hint: "[email-type]"
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Email Queue Implementation

Set up reliable email delivery with BullMQ queuing, template rendering, provider failover, and delivery tracking.

## Email Provider Decision

| Provider | Free Tier | Price After | Best For |
|----------|----------|-------------|----------|
| **Resend** | 100/day | $20/mo (50K) | Modern API, React Email templates |
| **SendGrid** | 100/day | $20/mo (50K) | Mature, marketing + transactional |
| **Postmark** | 100/mo | $15/mo (10K) | Transactional only, best deliverability |
| **AWS SES** | 62K/mo (from EC2) | $0.10/1K | Cheapest at scale |
| **Mailgun** | 100/day (trial) | $35/mo (50K) | Developer-friendly, good logs |

**Recommendation**: Resend for new projects (modern DX, React Email). AWS SES for high volume. Postmark for deliverability-critical apps.

## Step 1: Install Dependencies

```bash
npm install bullmq ioredis resend
# OR for nodemailer:
npm install bullmq ioredis nodemailer
npm install -D @types/nodemailer
```

## Step 2: Email Queue Setup

```typescript
// src/queues/email.queue.ts
import { Queue } from 'bullmq';
import { connection } from '../lib/redis';

export const emailQueue = new Queue('email', {
  connection,
  defaultJobOptions: {
    attempts: 3,
    backoff: {
      type: 'exponential',
      delay: 5000, // 5s, 10s, 20s
    },
    removeOnComplete: { age: 7 * 24 * 3600 },  // Keep 7 days
    removeOnFail: { age: 30 * 24 * 3600 },      // Keep 30 days
  },
});
```

## Step 3: Email Service

```typescript
// src/services/email.service.ts
import { emailQueue } from '../queues/email.queue';

// Type-safe email job data
interface EmailJob {
  to: string | string[];
  subject: string;
  template: string;
  data: Record<string, any>;
  replyTo?: string;
  priority?: 'high' | 'normal' | 'low';
  scheduledFor?: string; // ISO date
}

const PRIORITY_MAP = { high: 1, normal: 5, low: 10 };

export async function sendEmail(params: EmailJob) {
  const delay = params.scheduledFor
    ? new Date(params.scheduledFor).getTime() - Date.now()
    : undefined;

  await emailQueue.add(params.template, {
    to: params.to,
    subject: params.subject,
    template: params.template,
    data: params.data,
    replyTo: params.replyTo,
  }, {
    priority: PRIORITY_MAP[params.priority ?? 'normal'],
    delay: delay && delay > 0 ? delay : undefined,
    jobId: params.data.idempotencyKey, // Prevent duplicate sends
  });
}

// Convenience methods
export async function sendWelcomeEmail(user: { email: string; name: string }) {
  await sendEmail({
    to: user.email,
    subject: `Welcome to the platform, ${user.name}!`,
    template: 'welcome',
    data: {
      name: user.name,
      loginUrl: `${process.env.APP_URL}/login`,
      idempotencyKey: `welcome-${user.email}`,
    },
  });
}

export async function sendPasswordResetEmail(email: string, token: string) {
  await sendEmail({
    to: email,
    subject: 'Reset your password',
    template: 'password-reset',
    data: {
      resetUrl: `${process.env.APP_URL}/reset-password?token=${token}`,
      expiresIn: '1 hour',
      idempotencyKey: `reset-${token}`,
    },
    priority: 'high',
  });
}

export async function sendOrderConfirmation(order: {
  id: string;
  email: string;
  items: Array<{ name: string; price: number; qty: number }>;
  total: number;
}) {
  await sendEmail({
    to: order.email,
    subject: `Order confirmed: #${order.id}`,
    template: 'order-confirmation',
    data: {
      orderId: order.id,
      items: order.items,
      total: (order.total / 100).toFixed(2),
      idempotencyKey: `order-confirm-${order.id}`,
    },
  });
}
```

## Step 4: Email Worker with Resend

```typescript
// src/workers/email.worker.ts
import { Worker, Job } from 'bullmq';
import { Resend } from 'resend';
import { connection } from '../lib/redis';
import { renderTemplate } from '../lib/email-templates';

const resend = new Resend(process.env.RESEND_API_KEY);

const FROM_ADDRESS = process.env.EMAIL_FROM || 'noreply@yourdomain.com';

const worker = new Worker(
  'email',
  async (job: Job) => {
    const { to, subject, template, data, replyTo } = job.data;

    // Render template to HTML
    const html = renderTemplate(template, data);

    // Send via Resend
    const { data: result, error } = await resend.emails.send({
      from: FROM_ADDRESS,
      to: Array.isArray(to) ? to : [to],
      subject,
      html,
      reply_to: replyTo,
      tags: [
        { name: 'template', value: template },
        { name: 'jobId', value: job.id ?? 'unknown' },
      ],
    });

    if (error) {
      // Classify error for retry decision
      if (error.message.includes('rate limit') || error.message.includes('429')) {
        throw new Error(`TRANSIENT: ${error.message}`);
      }
      if (error.message.includes('invalid') || error.message.includes('not verified')) {
        throw new Error(`PERMANENT: ${error.message}`);
      }
      throw error;
    }

    return { messageId: result?.id, sentAt: new Date().toISOString() };
  },
  {
    connection,
    concurrency: 10,
    limiter: {
      max: 50,
      duration: 1000, // 50 emails per second max
    },
  }
);

// Graceful shutdown
process.on('SIGTERM', async () => {
  await worker.close();
  process.exit(0);
});
```

## Step 5: Alternative — Nodemailer Worker

```typescript
// src/workers/email.worker.ts (Nodemailer variant)
import { Worker, Job } from 'bullmq';
import nodemailer from 'nodemailer';
import { connection } from '../lib/redis';
import { renderTemplate } from '../lib/email-templates';

// SMTP transport (SendGrid, Mailgun, SES, etc.)
const transporter = nodemailer.createTransport({
  host: process.env.SMTP_HOST || 'smtp.sendgrid.net',
  port: parseInt(process.env.SMTP_PORT || '587'),
  auth: {
    user: process.env.SMTP_USER || 'apikey',
    pass: process.env.SMTP_PASS,
  },
});

// Verify connection on startup
transporter.verify().then(() => {
  console.log('SMTP connection verified');
}).catch((err) => {
  console.error('SMTP connection failed:', err);
  process.exit(1);
});

const worker = new Worker(
  'email',
  async (job: Job) => {
    const { to, subject, template, data, replyTo } = job.data;
    const html = renderTemplate(template, data);

    const info = await transporter.sendMail({
      from: process.env.EMAIL_FROM || '"App" <noreply@yourdomain.com>',
      to: Array.isArray(to) ? to.join(', ') : to,
      subject,
      html,
      replyTo,
    });

    return { messageId: info.messageId };
  },
  { connection, concurrency: 10 }
);

process.on('SIGTERM', async () => {
  await worker.close();
  transporter.close();
  process.exit(0);
});
```

## Step 6: Simple Template System

```typescript
// src/lib/email-templates.ts
import fs from 'fs';
import path from 'path';

const TEMPLATE_DIR = path.join(__dirname, '../../templates/email');

// Simple template variable replacement
export function renderTemplate(templateName: string, data: Record<string, any>): string {
  const filePath = path.join(TEMPLATE_DIR, `${templateName}.html`);

  if (!fs.existsSync(filePath)) {
    throw new Error(`Email template not found: ${templateName}`);
  }

  let html = fs.readFileSync(filePath, 'utf-8');

  // Replace {{variable}} with data values
  html = html.replace(/\{\{(\w+)\}\}/g, (match, key) => {
    return data[key] !== undefined ? String(data[key]) : match;
  });

  // Replace {{#each items}}...{{/each}} for simple loops
  html = html.replace(
    /\{\{#each (\w+)\}\}([\s\S]*?)\{\{\/each\}\}/g,
    (match, arrayKey, itemTemplate) => {
      const items = data[arrayKey];
      if (!Array.isArray(items)) return '';
      return items.map((item: any) => {
        let rendered = itemTemplate;
        for (const [key, value] of Object.entries(item)) {
          rendered = rendered.replace(
            new RegExp(`\\{\\{${key}\\}\\}`, 'g'),
            String(value)
          );
        }
        return rendered;
      }).join('');
    }
  );

  return html;
}
```

### Example Template

```html
<!-- templates/email/welcome.html -->
<!DOCTYPE html>
<html>
<head>
  <style>
    body { font-family: -apple-system, sans-serif; max-width: 600px; margin: 0 auto; }
    .button { background: #000; color: #fff; padding: 12px 24px; text-decoration: none;
              border-radius: 6px; display: inline-block; }
  </style>
</head>
<body>
  <h1>Welcome, {{name}}!</h1>
  <p>Thanks for signing up. Get started by logging in:</p>
  <a href="{{loginUrl}}" class="button">Log In</a>
  <p style="color: #666; font-size: 14px; margin-top: 40px;">
    If you didn't create this account, you can ignore this email.
  </p>
</body>
</html>
```

## Delivery Tracking

```typescript
// Track email status in your database
import { QueueEvents } from 'bullmq';

const emailEvents = new QueueEvents('email', { connection });

emailEvents.on('completed', async ({ jobId, returnvalue }) => {
  const result = JSON.parse(returnvalue);
  await db.emailLog.create({
    data: {
      jobId,
      messageId: result.messageId,
      status: 'sent',
      sentAt: new Date(result.sentAt),
    },
  });
});

emailEvents.on('failed', async ({ jobId, failedReason }) => {
  await db.emailLog.create({
    data: {
      jobId,
      status: 'failed',
      error: failedReason,
      failedAt: new Date(),
    },
  });
});
```

## Provider Failover

```typescript
// src/lib/email-sender.ts
async function sendWithFailover(params: {
  from: string; to: string[]; subject: string; html: string;
}): Promise<{ provider: string; messageId: string }> {
  // Try primary (Resend)
  try {
    const { data, error } = await resend.emails.send(params);
    if (error) throw error;
    return { provider: 'resend', messageId: data!.id };
  } catch (err) {
    console.warn('Primary email provider failed, trying fallback:', err);
  }

  // Fallback (Nodemailer/SMTP)
  try {
    const info = await transporter.sendMail({
      from: params.from,
      to: params.to.join(', '),
      subject: params.subject,
      html: params.html,
    });
    return { provider: 'smtp-fallback', messageId: info.messageId };
  } catch (err) {
    console.error('All email providers failed');
    throw err;
  }
}
```

## Gotchas

- Always use idempotency keys (jobId) to prevent duplicate emails. Users getting 2 welcome emails is a support nightmare
- Rate limit your email worker to match your provider's limits. Resend: 10/sec free, 100/sec paid. SendGrid: varies by plan
- Never put HTML email content in the job data — store templates on disk and pass template name + data variables
- Email templates MUST be tested across clients (Gmail, Outlook, Apple Mail). Use Litmus or Email on Acid
- For transactional email (password reset, order confirmation), use a dedicated sending domain — don't share with marketing email
- Set `replyTo` to a monitored inbox. Users will reply to transactional emails with support questions
- Store email logs with the messageId from your provider — essential for debugging delivery issues
- AWS SES requires domain/email verification before sending. Don't discover this in production
- Heroku scheduler + BullMQ is the recommended pattern for digest/batch emails. Don't use node-cron on Heroku (dyno cycling kills it)
