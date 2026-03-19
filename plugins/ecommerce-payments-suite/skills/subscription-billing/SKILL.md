---
name: subscription-billing
description: >
  Implement subscription billing with Stripe — plans, trials, upgrades/downgrades,
  proration, usage-based billing, dunning, and customer portal. Covers the full
  subscription lifecycle for SaaS applications.
  Triggers: "subscription billing", "recurring payments", "saas billing",
  "plans and pricing", "upgrade downgrade", "free trial".
  NOT for: one-time payments (use stripe-integration).
version: 1.0.0
argument-hint: "[setup|upgrade|portal|usage]"
allowed-tools: Read, Grep, Glob, Write, Edit, Bash
---

# Subscription Billing

Implement complete subscription billing for SaaS applications.

## Subscription Lifecycle

```
Free Trial → Active → Past Due → Canceled
                ↕
           Upgraded/Downgraded
```

### States

| Status | Meaning | Action |
|--------|---------|--------|
| `trialing` | Free trial period | Full access, show days remaining |
| `active` | Paying and current | Full access |
| `past_due` | Payment failed, retrying | Show warning, restrict after grace period |
| `canceled` | Subscription ended | Revoke access at period end |
| `incomplete` | Initial payment pending | Block access until payment completes |
| `incomplete_expired` | Initial payment failed | No access, prompt to re-subscribe |
| `unpaid` | All retries exhausted | No access, prompt to update payment |
| `paused` | Temporarily paused | No access, show resume option |

## Setup: Products and Prices in Stripe

```bash
# Create product
stripe products create \
  --name="Pro Plan" \
  --description="Full access to all features"

# Create monthly price
stripe prices create \
  --product=prod_xxx \
  --unit-amount=2900 \
  --currency=usd \
  --recurring[interval]=month

# Create annual price (with discount)
stripe prices create \
  --product=prod_xxx \
  --unit-amount=29000 \
  --currency=usd \
  --recurring[interval]=year
```

## Implementation

### Create Subscription with Trial

```typescript
app.post('/api/subscribe', async (req, res) => {
  const { priceId, paymentMethodId } = req.body;
  const user = req.user;

  // Get or create Stripe customer
  let customerId = user.stripeCustomerId;
  if (!customerId) {
    const customer = await stripe.customers.create({
      email: user.email,
      name: user.name,
      metadata: { userId: user.id },
    });
    customerId = customer.id;
    await db.user.update({ where: { id: user.id }, data: { stripeCustomerId: customerId } });
  }

  // Attach payment method
  if (paymentMethodId) {
    await stripe.paymentMethods.attach(paymentMethodId, { customer: customerId });
    await stripe.customers.update(customerId, {
      invoice_settings: { default_payment_method: paymentMethodId },
    });
  }

  // Create subscription with 14-day trial
  const subscription = await stripe.subscriptions.create({
    customer: customerId,
    items: [{ price: priceId }],
    trial_period_days: 14,
    payment_settings: {
      save_default_payment_method: 'on_subscription',
    },
    expand: ['latest_invoice.payment_intent'],
  });

  // Save subscription to database
  await db.subscription.create({
    data: {
      userId: user.id,
      stripeSubscriptionId: subscription.id,
      stripePriceId: priceId,
      status: subscription.status,
      trialEnd: subscription.trial_end
        ? new Date(subscription.trial_end * 1000)
        : null,
      currentPeriodEnd: new Date(subscription.current_period_end * 1000),
    },
  });

  res.json({
    subscriptionId: subscription.id,
    status: subscription.status,
    trialEnd: subscription.trial_end,
  });
});
```

### Upgrade / Downgrade

```typescript
app.post('/api/subscription/change-plan', async (req, res) => {
  const { newPriceId } = req.body;
  const user = req.user;

  const subscription = await db.subscription.findFirst({
    where: { userId: user.id, status: { in: ['active', 'trialing'] } },
  });

  if (!subscription) return res.status(404).json({ error: 'No active subscription' });

  const stripeSubscription = await stripe.subscriptions.retrieve(subscription.stripeSubscriptionId);

  // Update the subscription item with proration
  const updated = await stripe.subscriptions.update(subscription.stripeSubscriptionId, {
    items: [{
      id: stripeSubscription.items.data[0].id,
      price: newPriceId,
    }],
    proration_behavior: 'create_prorations', // charge/credit the difference
    // Options: 'create_prorations' | 'always_invoice' | 'none'
  });

  // Update local database
  await db.subscription.update({
    where: { id: subscription.id },
    data: {
      stripePriceId: newPriceId,
      status: updated.status,
    },
  });

  res.json({ status: updated.status });
});
```

### Cancel Subscription

```typescript
// Cancel at period end (recommended — user keeps access until billing cycle ends)
app.post('/api/subscription/cancel', async (req, res) => {
  const subscription = await getUserSubscription(req.user.id);

  const updated = await stripe.subscriptions.update(subscription.stripeSubscriptionId, {
    cancel_at_period_end: true,
  });

  await db.subscription.update({
    where: { id: subscription.id },
    data: { cancelAtPeriodEnd: true },
  });

  res.json({
    cancelAt: new Date(updated.current_period_end * 1000),
    message: `Your subscription will remain active until ${new Date(updated.current_period_end * 1000).toLocaleDateString()}`,
  });
});

// Reactivate (undo cancellation before period ends)
app.post('/api/subscription/reactivate', async (req, res) => {
  const subscription = await getUserSubscription(req.user.id);

  const updated = await stripe.subscriptions.update(subscription.stripeSubscriptionId, {
    cancel_at_period_end: false,
  });

  await db.subscription.update({
    where: { id: subscription.id },
    data: { cancelAtPeriodEnd: false },
  });

  res.json({ status: updated.status });
});
```

### Customer Portal (Stripe-Hosted)

```typescript
// Let customers manage their own subscription
app.post('/api/billing/portal', async (req, res) => {
  const user = req.user;

  const session = await stripe.billingPortal.sessions.create({
    customer: user.stripeCustomerId,
    return_url: `${process.env.APP_URL}/settings/billing`,
  });

  res.json({ url: session.url });
});
```

**Portal configuration** (in Stripe Dashboard → Settings → Billing → Customer Portal):
- Allow plan changes
- Allow cancellation
- Allow payment method updates
- Show invoice history

## Webhook Handlers for Subscriptions

```typescript
// Essential webhook events for subscriptions
switch (event.type) {
  // Trial ending soon — remind user to add payment method
  case 'customer.subscription.trial_will_end':
    const trial = event.data.object as Stripe.Subscription;
    await sendTrialEndingEmail(trial.customer as string, trial.trial_end!);
    break;

  // Subscription updated — plan change, status change
  case 'customer.subscription.updated':
    const updated = event.data.object as Stripe.Subscription;
    await db.subscription.updateMany({
      where: { stripeSubscriptionId: updated.id },
      data: {
        status: updated.status,
        stripePriceId: updated.items.data[0].price.id,
        currentPeriodEnd: new Date(updated.current_period_end * 1000),
        cancelAtPeriodEnd: updated.cancel_at_period_end,
      },
    });
    break;

  // Subscription canceled — revoke access
  case 'customer.subscription.deleted':
    const canceled = event.data.object as Stripe.Subscription;
    await db.subscription.updateMany({
      where: { stripeSubscriptionId: canceled.id },
      data: { status: 'canceled' },
    });
    await revokeAccess(canceled.customer as string);
    break;

  // Invoice paid — subscription renewed
  case 'invoice.paid':
    const paidInvoice = event.data.object as Stripe.Invoice;
    if (paidInvoice.subscription) {
      await db.subscription.updateMany({
        where: { stripeSubscriptionId: paidInvoice.subscription as string },
        data: { status: 'active' },
      });
    }
    break;

  // Invoice payment failed — dunning
  case 'invoice.payment_failed':
    const failedInvoice = event.data.object as Stripe.Invoice;
    await sendPaymentFailedEmail(failedInvoice.customer as string);
    // Stripe handles retry logic automatically (Smart Retries)
    break;
}
```

## Access Control Middleware

```typescript
function requireSubscription(allowedPlans?: string[]) {
  return async (req: Request, res: Response, next: NextFunction) => {
    const subscription = await db.subscription.findFirst({
      where: {
        userId: req.user.id,
        status: { in: ['active', 'trialing'] },
      },
    });

    if (!subscription) {
      return res.status(403).json({
        error: 'subscription_required',
        message: 'This feature requires an active subscription',
        upgradeUrl: '/pricing',
      });
    }

    if (allowedPlans && !allowedPlans.includes(subscription.stripePriceId)) {
      return res.status(403).json({
        error: 'plan_upgrade_required',
        message: 'This feature requires a higher plan',
        upgradeUrl: '/pricing',
      });
    }

    req.subscription = subscription;
    next();
  };
}

// Usage
app.get('/api/premium-feature', requireSubscription(['price_pro', 'price_enterprise']), handler);
```

## Usage-Based Billing

```typescript
// Report usage to Stripe
app.post('/api/usage/report', async (req, res) => {
  const { quantity } = req.body;
  const subscription = await getUserSubscription(req.user.id);

  const stripeSubscription = await stripe.subscriptions.retrieve(subscription.stripeSubscriptionId);
  const meteredItem = stripeSubscription.items.data.find(
    item => item.price.recurring?.usage_type === 'metered'
  );

  if (!meteredItem) return res.status(400).json({ error: 'No metered item' });

  await stripe.subscriptionItems.createUsageRecord(meteredItem.id, {
    quantity,
    timestamp: Math.floor(Date.now() / 1000),
    action: 'increment', // or 'set'
  });

  res.json({ recorded: quantity });
});
```

## Best Practices

1. **Always use webhooks for status changes** — don't rely on API responses alone
2. **Prorate by default** on plan changes — customers expect fair billing
3. **Cancel at period end** not immediately — give users what they paid for
4. **Handle dunning gracefully** — Stripe retries automatically, just notify users
5. **Customer portal** saves you from building billing management UI
6. **Track MRR** — monthly recurring revenue is the key SaaS metric
7. **Offer annual billing** — better cash flow, lower churn (typically 15-20% discount)