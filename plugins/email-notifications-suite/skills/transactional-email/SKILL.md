---
name: transactional-email
description: >
  Set up transactional email sending with Resend, SendGrid, Nodemailer, or AWS SES.
  Covers provider setup, template rendering, error handling, rate limiting, and
  deliverability best practices. Works with Express and Next.js.
  Triggers: "transactional email", "send email", "email setup", "email provider",
  "sendgrid", "resend", "ses", "nodemailer".
  NOT for: marketing email campaigns or newsletter tools.
version: 1.0.0
argument-hint: "[resend|sendgrid|ses|nodemailer]"
allowed-tools: Read, Grep, Glob, Write, Edit, Bash
---

# Transactional Email

Set up reliable transactional email for your application.

## Provider Comparison

| Provider | Pricing | Best For | Free Tier |
|----------|---------|----------|-----------|
| **Resend** | $0/mo (100 emails/day) | Developer DX, React Email | 100/day, 3,000/mo |
| **SendGrid** | Free (100/day) | High volume, analytics | 100/day |
| **AWS SES** | $0.10/1,000 emails | Cost at scale | 62,000/mo (from EC2) |
| **Nodemailer** | Free (self-hosted) | Full control, SMTP | Unlimited (SMTP costs) |

## Option 1: Resend (Recommended for New Projects)

```bash
npm install resend
```

```typescript
import { Resend } from 'resend';

const resend = new Resend(process.env.RESEND_API_KEY);

// Simple text email
await resend.emails.send({
  from: 'Your App <noreply@yourdomain.com>',
  to: 'user@example.com',
  subject: 'Welcome!',
  text: 'Thanks for signing up.',
});

// HTML email
await resend.emails.send({
  from: 'Your App <noreply@yourdomain.com>',
  to: 'user@example.com',
  subject: 'Order Confirmed',
  html: '<h1>Order #123</h1><p>Your order has been confirmed.</p>',
});

// With React Email component (see email-templates skill)
import { render } from '@react-email/render';
import WelcomeEmail from './emails/welcome';

const html = await render(WelcomeEmail({ name: 'Jane' }));
await resend.emails.send({
  from: 'Your App <noreply@yourdomain.com>',
  to: 'user@example.com',
  subject: 'Welcome!',
  html,
});
```

### Resend Setup

1. Sign up at resend.com
2. Add and verify your domain (DNS records)
3. Create API key
4. Set `RESEND_API_KEY` in `.env`

## Option 2: SendGrid

```bash
npm install @sendgrid/mail
```

```typescript
import sgMail from '@sendgrid/mail';

sgMail.setApiKey(process.env.SENDGRID_API_KEY!);

await sgMail.send({
  to: 'user@example.com',
  from: 'noreply@yourdomain.com',
  subject: 'Order Confirmed',
  text: 'Your order has been confirmed.',
  html: '<h1>Order Confirmed</h1>',
});

// With dynamic template (created in SendGrid dashboard)
await sgMail.send({
  to: 'user@example.com',
  from: 'noreply@yourdomain.com',
  templateId: 'd-xxxxx',
  dynamicTemplateData: {
    name: 'Jane',
    orderNumber: '12345',
  },
});

// Bulk send
await sgMail.sendMultiple({
  to: ['user1@example.com', 'user2@example.com'],
  from: 'noreply@yourdomain.com',
  subject: 'Update',
  text: 'New features available.',
});
```

## Option 3: Nodemailer (SMTP)

```bash
npm install nodemailer
```

```typescript
import nodemailer from 'nodemailer';

const transporter = nodemailer.createTransport({
  host: process.env.SMTP_HOST,       // e.g., smtp.gmail.com
  port: parseInt(process.env.SMTP_PORT || '587'),
  secure: false,                      // true for 465, false for 587
  auth: {
    user: process.env.SMTP_USER,
    pass: process.env.SMTP_PASS,
  },
});

await transporter.sendMail({
  from: '"Your App" <noreply@yourdomain.com>',
  to: 'user@example.com',
  subject: 'Welcome',
  text: 'Thanks for signing up.',
  html: '<h1>Welcome!</h1>',
});
```

## Option 4: AWS SES

```bash
npm install @aws-sdk/client-ses
```

```typescript
import { SESClient, SendEmailCommand } from '@aws-sdk/client-ses';

const ses = new SESClient({ region: 'us-east-1' });

await ses.send(new SendEmailCommand({
  Source: 'noreply@yourdomain.com',
  Destination: {
    ToAddresses: ['user@example.com'],
  },
  Message: {
    Subject: { Data: 'Welcome' },
    Body: {
      Html: { Data: '<h1>Welcome!</h1>' },
      Text: { Data: 'Welcome!' },
    },
  },
}));
```

## Email Service Layer

Wrap your provider in an abstraction so you can switch later:

```typescript
interface EmailOptions {
  to: string | string[];
  subject: string;
  html: string;
  text?: string;
  from?: string;
  replyTo?: string;
  attachments?: { filename: string; content: Buffer }[];
}

class EmailService {
  private defaultFrom = process.env.EMAIL_FROM || 'noreply@yourdomain.com';

  async send(options: EmailOptions): Promise<void> {
    const { to, subject, html, text, from, replyTo } = options;

    try {
      await resend.emails.send({
        from: from || this.defaultFrom,
        to: Array.isArray(to) ? to : [to],
        subject,
        html,
        text: text || htmlToText(html),
        reply_to: replyTo,
      });
    } catch (error) {
      console.error('Email send failed:', { to, subject, error });
      throw error;
    }
  }

  async sendTemplate(to: string, template: string, data: Record<string, any>): Promise<void> {
    const { subject, html } = await renderTemplate(template, data);
    await this.send({ to, subject, html });
  }
}

export const emailService = new EmailService();
```

## Common Transactional Emails

```typescript
// Welcome email
await emailService.sendTemplate(user.email, 'welcome', {
  name: user.name,
  loginUrl: `${APP_URL}/login`,
});

// Password reset
const resetToken = crypto.randomBytes(32).toString('hex');
await emailService.sendTemplate(user.email, 'password-reset', {
  name: user.name,
  resetUrl: `${APP_URL}/reset-password?token=${resetToken}`,
  expiresIn: '1 hour',
});

// Order confirmation
await emailService.sendTemplate(order.email, 'order-confirmation', {
  orderNumber: order.id,
  items: order.items,
  total: formatCurrency(order.total),
  shippingAddress: order.shippingAddress,
});

// Invoice
await emailService.sendTemplate(user.email, 'invoice', {
  invoiceNumber: invoice.number,
  amount: formatCurrency(invoice.amount),
  dueDate: formatDate(invoice.dueDate),
  payUrl: `${APP_URL}/invoices/${invoice.id}/pay`,
});
```

## Rate Limiting & Queuing

```typescript
import { Queue, Worker } from 'bullmq';
import Redis from 'ioredis';

const connection = new Redis(process.env.REDIS_URL);
const emailQueue = new Queue('emails', { connection });

// Add to queue instead of sending directly
export async function queueEmail(options: EmailOptions) {
  await emailQueue.add('send', options, {
    attempts: 3,
    backoff: { type: 'exponential', delay: 1000 },
    removeOnComplete: 100,
    removeOnFail: 500,
  });
}

// Worker processes the queue
const worker = new Worker('emails', async (job) => {
  await emailService.send(job.data);
}, {
  connection,
  concurrency: 5,          // Max 5 concurrent sends
  limiter: {
    max: 100,               // Max 100 per interval
    duration: 1000,          // Per second
  },
});

worker.on('failed', (job, err) => {
  console.error(`Email job ${job?.id} failed:`, err.message);
});
```

## Deliverability Best Practices

1. **Set up SPF, DKIM, and DMARC** DNS records for your domain
2. **Use a consistent "from" address** — don't change it frequently
3. **Include both HTML and plain text** versions
4. **Avoid spam trigger words** in subject lines
5. **Include unsubscribe link** in marketing emails (legally required)
6. **Monitor bounce rates** — remove invalid addresses immediately
7. **Warm up new domains** — start with low volume, increase gradually
8. **Use a subdomain** for transactional email (e.g., `mail.yourdomain.com`)
