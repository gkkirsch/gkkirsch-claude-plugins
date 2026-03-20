---
name: stripe-webhooks
description: >
  Stripe webhook endpoint setup, signature verification, event handling patterns,
  and idempotent processing. Covers checkout fulfillment, subscription lifecycle,
  payment failures, invoice events, and dispute handling.
  Triggers: "stripe webhooks", "webhook endpoint", "stripe events", "payment webhook",
  "subscription webhook", "webhook signature verification".
  NOT for: setting up Checkout (use stripe-checkout), subscription management (use stripe-subscriptions).
version: 1.0.0
allowed-tools: Read, Grep, Glob, Bash, Edit, Write
---

# Stripe Webhooks

## Setup

```bash
npm install stripe
```

**Critical**: Use `express.raw()` for the webhook route — Stripe needs the raw body for signature verification. Do NOT use `express.json()` globally before the webhook route.

## Webhook Endpoint

### Express Setup

```typescript
import Stripe from 'stripe';
import express from 'express';

const stripe = new Stripe(process.env.STRIPE_SECRET_KEY!);
const endpointSecret = process.env.STRIPE_WEBHOOK_SECRET!;

const app = express();

// IMPORTANT: Raw body parser MUST come before json parser
// or use a route-specific parser
app.post(
  '/api/webhooks/stripe',
  express.raw({ type: 'application/json' }),
  async (req, res) => {
    const sig = req.headers['stripe-signature']!;

    let event: Stripe.Event;

    try {
      event = stripe.webhooks.constructEvent(req.body, sig, endpointSecret);
    } catch (err: any) {
      console.error(`Webhook signature verification failed:`, err.message);
      return res.status(400).send(`Webhook Error: ${err.message}`);
    }

    // Respond immediately — process async
    res.json({ received: true });

    // Handle the event
    try {
      await handleStripeEvent(event);
    } catch (err) {
      console.error(`Error handling event ${event.id}:`, err);
      // Don't throw — we already sent 200
    }
  }
);

// JSON parser for all other routes
app.use(express.json());
```

### Why Respond 200 Immediately

Stripe retries on non-2xx responses. If your handler takes > 30 seconds, Stripe times out and retries. Always:
1. Verify signature
2. Send 200 immediately
3. Process the event asynchronously

If processing fails, you'll catch it in your own error handling — not via Stripe retries of a stale event.

## Signature Verification

### How It Works

Stripe signs every webhook with HMAC-SHA256 using your endpoint secret. The `stripe-signature` header contains:

```
t=1614556828,v1=abc123...,v0=def456...
```

- `t` = timestamp (prevents replay attacks)
- `v1` = HMAC signature (current scheme)
- `v0` = legacy signature (deprecated)

`constructEvent()` checks:
1. Signature matches the raw body
2. Timestamp is within tolerance (default: 300 seconds / 5 minutes)

### Custom Tolerance

```typescript
// Increase tolerance for slow networks (in seconds)
stripe.webhooks.constructEvent(body, sig, secret, 600); // 10 minutes
```

### Manual Verification (Without SDK)

```typescript
import crypto from 'crypto';

function verifyStripeSignature(
  payload: string,
  header: string,
  secret: string,
  tolerance = 300
): boolean {
  const parts = header.split(',');
  const timestamp = parseInt(parts.find(p => p.startsWith('t='))!.slice(2));
  const signature = parts.find(p => p.startsWith('v1='))!.slice(3);

  // Check timestamp tolerance
  const now = Math.floor(Date.now() / 1000);
  if (now - timestamp > tolerance) return false;

  // Compute expected signature
  const signedPayload = `${timestamp}.${payload}`;
  const expected = crypto
    .createHmac('sha256', secret)
    .update(signedPayload)
    .digest('hex');

  return crypto.timingSafeEqual(
    Buffer.from(signature),
    Buffer.from(expected)
  );
}
```

## Event Handler Pattern

### Router Pattern (Recommended)

```typescript
type EventHandler = (event: Stripe.Event) => Promise<void>;

const handlers: Record<string, EventHandler> = {
  'checkout.session.completed': handleCheckoutCompleted,
  'customer.subscription.created': handleSubscriptionCreated,
  'customer.subscription.updated': handleSubscriptionUpdated,
  'customer.subscription.deleted': handleSubscriptionDeleted,
  'invoice.payment_succeeded': handleInvoicePaymentSucceeded,
  'invoice.payment_failed': handleInvoicePaymentFailed,
  'customer.subscription.trial_will_end': handleTrialWillEnd,
  'charge.dispute.created': handleDisputeCreated,
};

async function handleStripeEvent(event: Stripe.Event): Promise<void> {
  const handler = handlers[event.type];

  if (!handler) {
    console.log(`Unhandled event type: ${event.type}`);
    return;
  }

  console.log(`Processing ${event.type} (${event.id})`);
  await handler(event);
}
```

### Checkout Fulfillment

```typescript
async function handleCheckoutCompleted(event: Stripe.Event) {
  const session = event.data.object as Stripe.Checkout.Session;

  if (session.mode === 'subscription') {
    // Subscription checkout — provision access
    await db.user.update({
      where: { email: session.customer_email! },
      data: {
        stripeCustomerId: session.customer as string,
        subscriptionId: session.subscription as string,
        subscriptionStatus: 'active',
      },
    });
  }

  if (session.mode === 'payment') {
    // One-time payment — fulfill the order
    const lineItems = await stripe.checkout.sessions.listLineItems(session.id);

    await db.order.create({
      data: {
        userId: session.metadata!.userId,
        stripeSessionId: session.id,
        amount: session.amount_total!,
        currency: session.currency!,
        status: 'paid',
        items: lineItems.data.map(item => ({
          priceId: item.price!.id,
          quantity: item.quantity,
        })),
      },
    });

    // Send confirmation email
    await sendOrderConfirmation(session.customer_email!, session.id);
  }
}
```

### Subscription Lifecycle

```typescript
async function handleSubscriptionCreated(event: Stripe.Event) {
  const subscription = event.data.object as Stripe.Subscription;

  await db.user.update({
    where: { stripeCustomerId: subscription.customer as string },
    data: {
      subscriptionId: subscription.id,
      subscriptionStatus: subscription.status,
      currentPlan: subscription.items.data[0].price.id,
      currentPeriodEnd: new Date(subscription.current_period_end * 1000),
    },
  });
}

async function handleSubscriptionUpdated(event: Stripe.Event) {
  const subscription = event.data.object as Stripe.Subscription;
  const previousAttributes = event.data.previous_attributes as any;

  await db.user.update({
    where: { stripeCustomerId: subscription.customer as string },
    data: {
      subscriptionStatus: subscription.status,
      currentPlan: subscription.items.data[0].price.id,
      currentPeriodEnd: new Date(subscription.current_period_end * 1000),
      cancelAtPeriodEnd: subscription.cancel_at_period_end,
    },
  });

  // Detect plan change
  if (previousAttributes?.items) {
    const oldPriceId = previousAttributes.items.data[0].price.id;
    const newPriceId = subscription.items.data[0].price.id;
    if (oldPriceId !== newPriceId) {
      console.log(`Plan changed: ${oldPriceId} → ${newPriceId}`);
      // Send plan change confirmation email
    }
  }
}

async function handleSubscriptionDeleted(event: Stripe.Event) {
  const subscription = event.data.object as Stripe.Subscription;

  await db.user.update({
    where: { stripeCustomerId: subscription.customer as string },
    data: {
      subscriptionStatus: 'canceled',
      subscriptionId: null,
      currentPlan: null,
    },
  });

  // Send cancellation survey email
  await sendCancellationSurvey(subscription.customer as string);
}
```

### Payment Failures

```typescript
async function handleInvoicePaymentFailed(event: Stripe.Event) {
  const invoice = event.data.object as Stripe.Invoice;

  // Only act on subscription invoices
  if (!invoice.subscription) return;

  const attemptCount = invoice.attempt_count;

  if (attemptCount === 1) {
    // First failure — soft notice
    await sendEmail(invoice.customer_email!, 'payment-failed-soft', {
      amount: (invoice.amount_due / 100).toFixed(2),
      currency: invoice.currency,
      updateUrl: await createBillingPortalUrl(invoice.customer as string),
    });
  } else if (attemptCount >= 3) {
    // Multiple failures — urgent notice
    await sendEmail(invoice.customer_email!, 'payment-failed-urgent', {
      amount: (invoice.amount_due / 100).toFixed(2),
      daysUntilCancel: 7,
      updateUrl: await createBillingPortalUrl(invoice.customer as string),
    });
  }

  // Log for analytics
  await db.paymentEvent.create({
    data: {
      customerId: invoice.customer as string,
      type: 'payment_failed',
      attemptCount,
      amount: invoice.amount_due,
      invoiceId: invoice.id,
    },
  });
}

async function handleInvoicePaymentSucceeded(event: Stripe.Event) {
  const invoice = event.data.object as Stripe.Invoice;

  // Reset failure state if this was a retry
  if (invoice.attempt_count > 1) {
    await sendEmail(invoice.customer_email!, 'payment-recovered', {
      amount: (invoice.amount_due / 100).toFixed(2),
    });
  }

  // Generate and store receipt
  if (invoice.hosted_invoice_url) {
    await db.receipt.create({
      data: {
        userId: await getUserByCustomerId(invoice.customer as string),
        invoiceId: invoice.id,
        amount: invoice.amount_paid,
        receiptUrl: invoice.hosted_invoice_url,
        paidAt: new Date(invoice.status_transitions?.paid_at! * 1000),
      },
    });
  }
}
```

### Trial Ending

```typescript
async function handleTrialWillEnd(event: Stripe.Event) {
  const subscription = event.data.object as Stripe.Subscription;
  // Fires 3 days before trial ends

  const user = await db.user.findUnique({
    where: { stripeCustomerId: subscription.customer as string },
  });

  if (!user) return;

  // Check if payment method is on file
  const customer = await stripe.customers.retrieve(
    subscription.customer as string
  ) as Stripe.Customer;

  const hasPaymentMethod = !!customer.invoice_settings?.default_payment_method;

  await sendEmail(user.email, 'trial-ending', {
    daysLeft: 3,
    planName: subscription.items.data[0].price.nickname,
    hasPaymentMethod,
    addCardUrl: hasPaymentMethod
      ? undefined
      : await createBillingPortalUrl(subscription.customer as string),
  });
}
```

### Dispute Handling

```typescript
async function handleDisputeCreated(event: Stripe.Event) {
  const dispute = event.data.object as Stripe.Dispute;

  // Log the dispute
  await db.dispute.create({
    data: {
      stripeDisputeId: dispute.id,
      chargeId: dispute.charge as string,
      amount: dispute.amount,
      reason: dispute.reason,
      status: dispute.status,
      evidenceDueBy: new Date(dispute.evidence_details!.due_by * 1000),
    },
  });

  // Alert the team immediately
  await notifyTeam('dispute', {
    amount: (dispute.amount / 100).toFixed(2),
    reason: dispute.reason,
    chargeId: dispute.charge,
    dueBy: new Date(dispute.evidence_details!.due_by * 1000).toISOString(),
  });
}
```

## Idempotent Processing

Stripe may deliver the same event multiple times. Always make your handlers idempotent.

### Pattern: Event Log Table

```typescript
// Prisma schema
// model ProcessedEvent {
//   id        String   @id
//   type      String
//   processed Boolean  @default(false)
//   result    Json?
//   createdAt DateTime @default(now())
// }

async function processEventIdempotently(
  event: Stripe.Event,
  handler: EventHandler
): Promise<void> {
  // Check if already processed
  const existing = await db.processedEvent.findUnique({
    where: { id: event.id },
  });

  if (existing?.processed) {
    console.log(`Event ${event.id} already processed, skipping`);
    return;
  }

  // Upsert to claim the event (prevents concurrent processing)
  await db.processedEvent.upsert({
    where: { id: event.id },
    create: { id: event.id, type: event.type, processed: false },
    update: {},
  });

  // Process
  await handler(event);

  // Mark as processed
  await db.processedEvent.update({
    where: { id: event.id },
    data: { processed: true },
  });
}
```

### Use in Main Handler

```typescript
async function handleStripeEvent(event: Stripe.Event): Promise<void> {
  const handler = handlers[event.type];
  if (!handler) return;

  await processEventIdempotently(event, handler);
}
```

## Local Development with Stripe CLI

### Setup

```bash
# Install
brew install stripe/stripe-cli/stripe

# Login
stripe login

# Forward events to local server
stripe listen --forward-to localhost:3000/api/webhooks/stripe

# Copy the webhook signing secret (whsec_...) to .env
```

### Trigger Test Events

```bash
# Trigger specific events
stripe trigger checkout.session.completed
stripe trigger customer.subscription.created
stripe trigger invoice.payment_failed
stripe trigger customer.subscription.trial_will_end

# Trigger with custom data
stripe trigger checkout.session.completed \
  --override checkout_session:metadata.userId=user_123

# List recent events
stripe events list --limit 5
```

### Testing Webhook Handler

```typescript
// test/webhooks.test.ts
import request from 'supertest';
import Stripe from 'stripe';

const stripe = new Stripe(process.env.STRIPE_SECRET_KEY!);

function createTestEvent(type: string, data: any): string {
  const event = {
    id: `evt_test_${Date.now()}`,
    type,
    data: { object: data },
    created: Math.floor(Date.now() / 1000),
    livemode: false,
    api_version: '2024-12-18.acacia',
  };
  return JSON.stringify(event);
}

// For testing without signature verification
describe('Webhook handlers', () => {
  it('handles checkout.session.completed', async () => {
    const payload = createTestEvent('checkout.session.completed', {
      id: 'cs_test_123',
      mode: 'subscription',
      customer: 'cus_test_123',
      customer_email: 'test@example.com',
      subscription: 'sub_test_123',
    });

    // In test env, bypass signature verification
    // or use stripe.webhooks.generateTestHeaderString()
    const header = stripe.webhooks.generateTestHeaderString({
      payload,
      secret: process.env.STRIPE_WEBHOOK_SECRET!,
    });

    const res = await request(app)
      .post('/api/webhooks/stripe')
      .set('stripe-signature', header)
      .set('content-type', 'application/json')
      .send(payload);

    expect(res.status).toBe(200);

    const user = await db.user.findUnique({
      where: { email: 'test@example.com' },
    });
    expect(user?.subscriptionId).toBe('sub_test_123');
  });
});
```

## Production Configuration

### Register Webhook Endpoint in Stripe Dashboard

1. Go to **Developers → Webhooks → Add endpoint**
2. URL: `https://yourdomain.com/api/webhooks/stripe`
3. Select events to listen to:

**Essential events** (subscribe to these at minimum):

| Event | Why |
|-------|-----|
| `checkout.session.completed` | Fulfill purchases, provision access |
| `customer.subscription.created` | New subscription created |
| `customer.subscription.updated` | Plan change, status change, cancel scheduled |
| `customer.subscription.deleted` | Subscription cancelled |
| `invoice.payment_succeeded` | Successful recurring payment |
| `invoice.payment_failed` | Failed payment — trigger dunning |
| `customer.subscription.trial_will_end` | 3 days before trial expires |

**Optional but useful events:**

| Event | Why |
|-------|-----|
| `invoice.finalized` | Invoice ready (before payment attempt) |
| `charge.dispute.created` | Chargeback received — needs immediate response |
| `charge.refunded` | Track refunds |
| `customer.updated` | Email or payment method changed |
| `payment_intent.payment_failed` | One-time payment failure details |

### Environment Variables

```bash
# .env
STRIPE_SECRET_KEY=sk_live_...        # API key
STRIPE_WEBHOOK_SECRET=whsec_...      # Endpoint signing secret
STRIPE_PUBLISHABLE_KEY=pk_live_...   # For client-side
```

### Multiple Webhook Endpoints

You can register multiple endpoints for different concerns:

```
/api/webhooks/stripe/billing   → subscription + invoice events
/api/webhooks/stripe/checkout  → checkout.session events
/api/webhooks/stripe/disputes  → charge.dispute events
```

Each endpoint gets its own signing secret.

## Common Gotchas

1. **Raw body required** — `express.json()` before the webhook route breaks signature verification. Use `express.raw({ type: 'application/json' })` on the route.

2. **Respond before processing** — Send 200 immediately, then process. Stripe times out after 30 seconds and retries. Three failures in a row and Stripe disables the endpoint.

3. **Events arrive out of order** — `invoice.payment_succeeded` might arrive before `customer.subscription.created`. Always check for existing records and handle gracefully.

4. **Test mode vs live mode** — Webhook secrets are different for test and live mode. Use separate endpoints or check `event.livemode`.

5. **`previous_attributes` for detecting changes** — `event.data.previous_attributes` tells you what changed. Check this to avoid unnecessary work on subscription.updated events.

6. **Customer ID types** — `subscription.customer` can be a string (ID) or an expanded Customer object, depending on your API call. Always cast: `subscription.customer as string`.

7. **Idempotency is not optional** — Stripe explicitly documents that events may be delivered more than once. Always deduplicate by `event.id`.

8. **Clock skew** — If your server clock is off by more than 300 seconds (5 min), signature verification fails. Use NTP.
