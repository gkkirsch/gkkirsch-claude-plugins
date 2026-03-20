---
name: stripe-subscriptions
description: >
  Implement Stripe subscription billing — plan management, trials,
  upgrades/downgrades, per-seat billing, metered usage, billing portals,
  cancellation flows, and revenue recovery. Production patterns with TypeScript.
  Triggers: "subscription billing", "recurring payments", "SaaS billing",
  "plan upgrade", "per-seat pricing", "metered billing", "usage-based billing",
  "trial period", "billing portal".
  NOT for: one-time payments (use stripe-checkout), marketplace payments (use stripe-connect).
version: 1.0.0
allowed-tools: Read, Grep, Glob, Bash, Edit, Write
---

# Stripe Subscription Billing

## Database Schema

```prisma
// prisma/schema.prisma
model User {
  id                String   @id @default(cuid())
  email             String   @unique
  stripeCustomerId  String?  @unique
  subscriptionId    String?
  subscriptionStatus String?  // 'trialing' | 'active' | 'past_due' | 'canceled'
  planId            String?  // Your internal plan identifier
  currentPeriodEnd  DateTime?
  trialEndsAt       DateTime?
  cancelAtPeriodEnd Boolean  @default(false)
  createdAt         DateTime @default(now())
}

model Plan {
  id             String  @id @default(cuid())
  name           String  // "Starter", "Pro", "Enterprise"
  stripePriceId  String  @unique
  monthlyPrice   Int     // in cents
  yearlyPriceId  String? // annual pricing
  features       Json    // ["Feature 1", "Feature 2"]
  limits         Json    // { "seats": 5, "projects": 10, "storage_gb": 50 }
  sortOrder      Int     @default(0)
  active         Boolean @default(true)
}
```

---

## Subscription Management

### Create Subscription via Checkout

```typescript
// POST /api/subscribe
router.post('/api/subscribe', requireAuth, async (req, res) => {
  const { priceId } = req.body;
  const user = req.user;

  // Get or create Stripe customer
  let customerId = user.stripeCustomerId;
  if (!customerId) {
    const customer = await stripe.customers.create({
      email: user.email,
      metadata: { userId: user.id },
    });
    customerId = customer.id;
    await db.user.update({
      where: { id: user.id },
      data: { stripeCustomerId: customerId },
    });
  }

  // Check for existing active subscription
  if (user.subscriptionId && user.subscriptionStatus === 'active') {
    // Redirect to plan change flow instead
    return res.status(400).json({
      error: 'Already subscribed. Use /api/subscription/change to switch plans.',
    });
  }

  const session = await stripe.checkout.sessions.create({
    mode: 'subscription',
    customer: customerId,
    line_items: [{ price: priceId, quantity: 1 }],
    subscription_data: {
      trial_period_days: 14,
      metadata: { userId: user.id },
    },
    success_url: `${process.env.APP_URL}/dashboard?subscribed=true`,
    cancel_url: `${process.env.APP_URL}/pricing`,
    allow_promotion_codes: true,
  });

  res.json({ url: session.url });
});
```

### Change Plan (Upgrade/Downgrade)

```typescript
// POST /api/subscription/change
router.post('/api/subscription/change', requireAuth, async (req, res) => {
  const { newPriceId } = req.body;
  const user = req.user;

  if (!user.subscriptionId) {
    return res.status(400).json({ error: 'No active subscription' });
  }

  const subscription = await stripe.subscriptions.retrieve(user.subscriptionId);
  const currentItem = subscription.items.data[0];

  // Determine if upgrade or downgrade
  const currentPrice = await stripe.prices.retrieve(currentItem.price.id);
  const newPrice = await stripe.prices.retrieve(newPriceId);
  const isUpgrade = (newPrice.unit_amount || 0) > (currentPrice.unit_amount || 0);

  const updated = await stripe.subscriptions.update(user.subscriptionId, {
    items: [{
      id: currentItem.id,
      price: newPriceId,
    }],
    proration_behavior: isUpgrade ? 'always_invoice' : 'create_prorations',
    // Upgrade: charge difference immediately
    // Downgrade: credit applied to next invoice
  });

  res.json({
    success: true,
    newPlan: newPriceId,
    effectiveDate: isUpgrade ? 'immediate' : 'next billing cycle',
  });
});
```

### Cancel Subscription

```typescript
// POST /api/subscription/cancel
router.post('/api/subscription/cancel', requireAuth, async (req, res) => {
  const user = req.user;

  if (!user.subscriptionId) {
    return res.status(400).json({ error: 'No active subscription' });
  }

  // Cancel at period end (user keeps access until paid period expires)
  await stripe.subscriptions.update(user.subscriptionId, {
    cancel_at_period_end: true,
  });

  await db.user.update({
    where: { id: user.id },
    data: { cancelAtPeriodEnd: true },
  });

  res.json({
    success: true,
    accessUntil: user.currentPeriodEnd,
    message: 'Your subscription will end at the current billing period.',
  });
});

// POST /api/subscription/reactivate (undo cancellation)
router.post('/api/subscription/reactivate', requireAuth, async (req, res) => {
  const user = req.user;

  if (!user.subscriptionId || !user.cancelAtPeriodEnd) {
    return res.status(400).json({ error: 'No pending cancellation' });
  }

  await stripe.subscriptions.update(user.subscriptionId, {
    cancel_at_period_end: false,
  });

  await db.user.update({
    where: { id: user.id },
    data: { cancelAtPeriodEnd: false },
  });

  res.json({ success: true, message: 'Subscription reactivated' });
});
```

---

## Per-Seat Billing

```typescript
// POST /api/subscription/seats
router.post('/api/subscription/seats', requireAuth, async (req, res) => {
  const { action } = req.body;  // 'add' | 'remove'
  const user = req.user;

  const subscription = await stripe.subscriptions.retrieve(user.subscriptionId!);
  const item = subscription.items.data[0];
  const currentQuantity = item.quantity || 1;

  if (action === 'remove' && currentQuantity <= 1) {
    return res.status(400).json({ error: 'Cannot remove last seat' });
  }

  const newQuantity = action === 'add' ? currentQuantity + 1 : currentQuantity - 1;

  await stripe.subscriptions.update(user.subscriptionId!, {
    items: [{
      id: item.id,
      quantity: newQuantity,
    }],
    proration_behavior: action === 'add' ? 'always_invoice' : 'create_prorations',
  });

  res.json({ success: true, seats: newQuantity });
});
```

---

## Metered / Usage-Based Billing

```typescript
// Report usage (call whenever a billable event occurs)
async function reportUsage(subscriptionItemId: string, quantity: number) {
  await stripe.subscriptionItems.createUsageRecord(subscriptionItemId, {
    quantity,
    timestamp: Math.floor(Date.now() / 1000),
    action: 'increment',
  });
}

// Example: track API calls per request
app.use('/api', async (req, res, next) => {
  // After successful response
  res.on('finish', async () => {
    if (res.statusCode < 400 && req.user?.subscriptionItemId) {
      await reportUsage(req.user.subscriptionItemId, 1);
    }
  });
  next();
});

// Get current usage for display
router.get('/api/usage', requireAuth, async (req, res) => {
  const summary = await stripe.subscriptionItems.listUsageRecordSummaries(
    req.user.subscriptionItemId,
    { limit: 1 }
  );

  res.json({
    currentPeriodUsage: summary.data[0]?.total_usage || 0,
    // Add your plan limit for display
    limit: req.user.planLimits?.apiCalls || Infinity,
  });
});
```

---

## Access Control Middleware

```typescript
// middleware/requireSubscription.ts
export function requireSubscription(allowedPlans?: string[]) {
  return async (req: Request, res: Response, next: NextFunction) => {
    const user = req.user;

    if (!user) {
      return res.status(401).json({ error: 'Authentication required' });
    }

    // Check subscription status
    const activeStatuses = ['trialing', 'active'];
    if (!activeStatuses.includes(user.subscriptionStatus || '')) {
      return res.status(403).json({
        error: 'Active subscription required',
        subscriptionStatus: user.subscriptionStatus,
        upgradeUrl: '/pricing',
      });
    }

    // Check specific plan if required
    if (allowedPlans && !allowedPlans.includes(user.planId || '')) {
      return res.status(403).json({
        error: 'This feature requires a higher plan',
        currentPlan: user.planId,
        requiredPlans: allowedPlans,
        upgradeUrl: '/pricing',
      });
    }

    next();
  };
}

// Usage
router.get('/api/advanced-feature',
  requireSubscription(['pro', 'enterprise']),
  handler
);
```

---

## Feature Gating

```typescript
// src/lib/features.ts
interface PlanLimits {
  seats: number;
  projects: number;
  storageGb: number;
  apiCallsPerMonth: number;
  features: string[];
}

const PLAN_LIMITS: Record<string, PlanLimits> = {
  starter: {
    seats: 1,
    projects: 3,
    storageGb: 5,
    apiCallsPerMonth: 1000,
    features: ['basic-analytics', 'email-support'],
  },
  pro: {
    seats: 10,
    projects: 50,
    storageGb: 100,
    apiCallsPerMonth: 50000,
    features: ['basic-analytics', 'advanced-analytics', 'priority-support', 'api-access'],
  },
  enterprise: {
    seats: Infinity,
    projects: Infinity,
    storageGb: Infinity,
    apiCallsPerMonth: Infinity,
    features: ['basic-analytics', 'advanced-analytics', 'priority-support', 'api-access', 'sso', 'audit-log', 'custom-branding'],
  },
};

export function getPlanLimits(planId: string): PlanLimits {
  return PLAN_LIMITS[planId] || PLAN_LIMITS.starter;
}

export function hasFeature(planId: string, feature: string): boolean {
  return getPlanLimits(planId).features.includes(feature);
}

export function isWithinLimit(planId: string, resource: keyof Omit<PlanLimits, 'features'>, current: number): boolean {
  const limit = getPlanLimits(planId)[resource] as number;
  return current < limit;
}
```

---

## Billing Portal

```typescript
// POST /api/billing-portal
router.post('/api/billing-portal', requireAuth, async (req, res) => {
  if (!req.user.stripeCustomerId) {
    return res.status(400).json({ error: 'No billing account' });
  }

  const session = await stripe.billingPortal.sessions.create({
    customer: req.user.stripeCustomerId,
    return_url: `${process.env.APP_URL}/settings/billing`,
    flow_data: req.body.flow ? {
      type: req.body.flow,  // 'payment_method_update' | 'subscription_cancel' | 'subscription_update'
    } : undefined,
  });

  res.json({ url: session.url });
});
```

Configure in Stripe Dashboard → Settings → Billing → Customer portal:
- ✅ Payment methods: customers can update cards
- ✅ Invoices: customers can view past invoices
- ✅ Subscriptions: customers can cancel
- ✅ Plan switching: list which prices they can switch between

---

## Pricing Page Component

```typescript
// components/PricingPage.tsx
'use client';

interface Plan {
  id: string;
  name: string;
  price: number;
  priceId: string;
  yearlyPriceId?: string;
  features: string[];
  popular?: boolean;
}

const plans: Plan[] = [
  {
    id: 'starter',
    name: 'Starter',
    price: 9,
    priceId: 'price_starter_monthly',
    yearlyPriceId: 'price_starter_yearly',
    features: ['1 user', '3 projects', '5GB storage', 'Email support'],
  },
  {
    id: 'pro',
    name: 'Pro',
    price: 29,
    priceId: 'price_pro_monthly',
    yearlyPriceId: 'price_pro_yearly',
    features: ['10 users', '50 projects', '100GB storage', 'Priority support', 'API access'],
    popular: true,
  },
  {
    id: 'enterprise',
    name: 'Enterprise',
    price: 99,
    priceId: 'price_enterprise_monthly',
    yearlyPriceId: 'price_enterprise_yearly',
    features: ['Unlimited users', 'Unlimited projects', 'Unlimited storage', 'SSO', 'Audit log', 'Custom branding'],
  },
];

export function PricingPage({ currentPlan }: { currentPlan?: string }) {
  const [annual, setAnnual] = useState(false);
  const [loading, setLoading] = useState<string | null>(null);

  const handleSubscribe = async (plan: Plan) => {
    setLoading(plan.id);
    const priceId = annual && plan.yearlyPriceId ? plan.yearlyPriceId : plan.priceId;

    const res = await fetch('/api/subscribe', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ priceId }),
    });

    const { url } = await res.json();
    window.location.href = url;
  };

  return (
    <div>
      {/* Annual toggle */}
      <div className="flex items-center justify-center gap-3 mb-8">
        <span>Monthly</span>
        <button
          onClick={() => setAnnual(!annual)}
          className={`relative w-14 h-7 rounded-full transition ${annual ? 'bg-blue-600' : 'bg-gray-300'}`}
        >
          <span className={`absolute top-1 w-5 h-5 bg-white rounded-full transition ${annual ? 'left-8' : 'left-1'}`} />
        </button>
        <span>Annual <span className="text-green-600 text-sm">(Save 20%)</span></span>
      </div>

      {/* Plan cards */}
      <div className="grid md:grid-cols-3 gap-6">
        {plans.map(plan => (
          <div key={plan.id} className={`border rounded-lg p-6 ${plan.popular ? 'border-blue-500 ring-2 ring-blue-500' : ''}`}>
            {plan.popular && <span className="text-sm bg-blue-500 text-white px-2 py-1 rounded">Most Popular</span>}
            <h3 className="text-xl font-bold mt-2">{plan.name}</h3>
            <p className="text-3xl font-bold mt-2">
              ${annual ? Math.round(plan.price * 0.8) : plan.price}
              <span className="text-sm font-normal">/mo</span>
            </p>
            <ul className="mt-4 space-y-2">
              {plan.features.map(f => (
                <li key={f} className="flex items-center gap-2">
                  <span className="text-green-500">✓</span> {f}
                </li>
              ))}
            </ul>
            <button
              onClick={() => handleSubscribe(plan)}
              disabled={loading === plan.id || currentPlan === plan.id}
              className="w-full mt-6 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50"
            >
              {currentPlan === plan.id ? 'Current Plan' : loading === plan.id ? 'Loading...' : 'Get Started'}
            </button>
          </div>
        ))}
      </div>
    </div>
  );
}
```
