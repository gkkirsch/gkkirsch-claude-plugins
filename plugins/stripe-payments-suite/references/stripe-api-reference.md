# Stripe API Quick Reference

Essential API patterns, object relationships, and common operations for Stripe integration.

---

## Object Hierarchy

```
Customer
├── PaymentMethod (card, bank account, etc.)
├── Subscription
│   ├── SubscriptionItem → Price → Product
│   ├── Invoice (auto-generated each billing cycle)
│   │   └── InvoiceLineItem
│   └── Discount (coupon applied)
├── Checkout.Session (one-time purchase flow)
│   └── LineItem → Price → Product
├── PaymentIntent (single payment)
│   └── Charge
│       └── Refund
└── BillingPortal.Session (self-service management)

Connected Account (Stripe Connect)
├── Account (Standard/Express/Custom)
├── Transfer (platform → connected account)
├── Payout (connected account → bank)
└── ApplicationFee (platform revenue)
```

## Products & Prices

### Create a Product with Prices

```typescript
// Product = what you sell
const product = await stripe.products.create({
  name: 'Pro Plan',
  description: 'Full access to all features',
  metadata: { tier: 'pro' },
});

// Price = how much it costs (can have multiple per product)
const monthlyPrice = await stripe.prices.create({
  product: product.id,
  unit_amount: 2900,          // $29.00
  currency: 'usd',
  recurring: { interval: 'month' },
  lookup_key: 'pro_monthly',  // Stable key for code references
});

const yearlyPrice = await stripe.prices.create({
  product: product.id,
  unit_amount: 29000,         // $290.00 (2 months free)
  currency: 'usd',
  recurring: { interval: 'year' },
  lookup_key: 'pro_yearly',
});
```

### Price Types

| Type | `recurring` | `unit_amount` | Use Case |
|------|-------------|---------------|----------|
| One-time | omit | set | Single purchases |
| Flat recurring | `{ interval }` | set | Fixed subscription |
| Per-seat | `{ interval }` | set + quantity on sub | Per-user pricing |
| Metered | `{ interval, usage_type: 'metered' }` | set | Pay-per-use (API calls, etc.) |
| Tiered | `{ interval }` | omit, use `tiers` | Volume discounts |

### Lookup by Key (Recommended for Code)

```typescript
// Find price by lookup_key instead of hardcoding price IDs
const prices = await stripe.prices.list({
  lookup_keys: ['pro_monthly', 'pro_yearly'],
});
```

## Customers

```typescript
// Create
const customer = await stripe.customers.create({
  email: 'user@example.com',
  name: 'Jane Smith',
  metadata: { userId: 'user_123' },
});

// Retrieve
const customer = await stripe.customers.retrieve('cus_...');

// Update
await stripe.customers.update('cus_...', {
  metadata: { plan: 'pro' },
});

// Search (useful for finding by email)
const result = await stripe.customers.search({
  query: 'email:"user@example.com"',
});
```

## Payment Methods

```typescript
// Attach payment method to customer
await stripe.paymentMethods.attach('pm_...', {
  customer: 'cus_...',
});

// Set as default
await stripe.customers.update('cus_...', {
  invoice_settings: {
    default_payment_method: 'pm_...',
  },
});

// List customer's payment methods
const methods = await stripe.paymentMethods.list({
  customer: 'cus_...',
  type: 'card',
});
```

## Subscriptions

```typescript
// Create
const subscription = await stripe.subscriptions.create({
  customer: 'cus_...',
  items: [{ price: 'price_...' }],
  trial_period_days: 14,
  payment_behavior: 'default_incomplete',  // Collect payment intent
  expand: ['latest_invoice.payment_intent'],
});

// Update (change plan)
await stripe.subscriptions.update('sub_...', {
  items: [{
    id: subscription.items.data[0].id,
    price: 'price_new_plan',
  }],
  proration_behavior: 'create_prorations',
});

// Cancel at period end
await stripe.subscriptions.update('sub_...', {
  cancel_at_period_end: true,
});

// Cancel immediately
await stripe.subscriptions.cancel('sub_...');

// Reactivate (if cancel_at_period_end was set)
await stripe.subscriptions.update('sub_...', {
  cancel_at_period_end: false,
});
```

### Subscription Statuses

| Status | Access? | What to Do |
|--------|---------|------------|
| `trialing` | Yes | Show trial days remaining |
| `active` | Yes | Normal state |
| `past_due` | Grace period | Show payment update banner |
| `incomplete` | No | Initial payment failed — retry |
| `incomplete_expired` | No | Initial payment window expired |
| `canceled` | No | Subscription ended |
| `unpaid` | No | All retry attempts failed |
| `paused` | No | Manually paused |

## Invoices

```typescript
// List customer's invoices
const invoices = await stripe.invoices.list({
  customer: 'cus_...',
  limit: 10,
});

// Upcoming invoice (preview next charge)
const upcoming = await stripe.invoices.retrieveUpcoming({
  customer: 'cus_...',
});

// Void an invoice
await stripe.invoices.voidInvoice('in_...');
```

## Coupons & Promotion Codes

```typescript
// Create coupon
const coupon = await stripe.coupons.create({
  percent_off: 20,
  duration: 'once',            // 'forever', 'once', 'repeating'
  // duration_in_months: 3,    // for 'repeating'
  max_redemptions: 100,
  redeem_by: Math.floor(Date.now() / 1000) + 30 * 86400, // 30 days
});

// Create shareable promotion code
const promoCode = await stripe.promotionCodes.create({
  coupon: coupon.id,
  code: 'LAUNCH20',
  max_redemptions: 50,
});

// Apply to checkout session
const session = await stripe.checkout.sessions.create({
  // ...
  allow_promotion_codes: true,        // Let user enter code
  // OR
  discounts: [{ promotion_code: promoCode.id }],  // Pre-apply
});
```

## Amounts & Currencies

**All amounts are in the smallest currency unit:**

| Currency | Unit | $29.99 = |
|----------|------|----------|
| USD | cents | `2999` |
| EUR | cents | `2999` |
| GBP | pence | `2999` |
| JPY | yen (no decimals) | `2999` |

```typescript
// Convert for display
const displayAmount = (amount: number, currency: string) => {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: currency.toUpperCase(),
  }).format(amount / 100);
};

// Zero-decimal currencies (JPY, KRW, etc.) — no division needed
const ZERO_DECIMAL = ['bif','clp','djf','gnf','jpy','kmf','krw','mga','pyg','rwf','ugx','vnd','vuv','xaf','xof','xpf'];
```

## Error Handling

```typescript
try {
  await stripe.paymentIntents.create({ ... });
} catch (err) {
  if (err instanceof Stripe.errors.StripeCardError) {
    // Card declined — show user the decline message
    console.log(err.message);    // "Your card was declined."
    console.log(err.code);       // "card_declined"
    console.log(err.decline_code); // "insufficient_funds"
  } else if (err instanceof Stripe.errors.StripeRateLimitError) {
    // Too many requests — retry with backoff
  } else if (err instanceof Stripe.errors.StripeInvalidRequestError) {
    // Invalid parameters — developer error
  } else if (err instanceof Stripe.errors.StripeAuthenticationError) {
    // Wrong API key
  } else if (err instanceof Stripe.errors.StripeAPIError) {
    // Stripe server error — retry
  } else if (err instanceof Stripe.errors.StripeConnectionError) {
    // Network issue — retry
  }
}
```

## Pagination

```typescript
// Auto-pagination (recommended)
for await (const customer of stripe.customers.list({ limit: 100 })) {
  console.log(customer.id);
}

// Manual pagination
let hasMore = true;
let startingAfter: string | undefined;

while (hasMore) {
  const page = await stripe.customers.list({
    limit: 100,
    starting_after: startingAfter,
  });

  for (const customer of page.data) {
    // process
  }

  hasMore = page.has_more;
  startingAfter = page.data[page.data.length - 1]?.id;
}
```

## Metadata

Every Stripe object supports `metadata` — up to 50 key-value pairs, keys up to 40 chars, values up to 500 chars.

```typescript
// Set metadata at creation
await stripe.customers.create({
  email: 'user@example.com',
  metadata: {
    userId: 'user_123',
    plan: 'pro',
    source: 'landing-page',
  },
});

// Update metadata (merges, doesn't replace)
await stripe.customers.update('cus_...', {
  metadata: { plan: 'enterprise' },
});

// Delete a metadata key
await stripe.customers.update('cus_...', {
  metadata: { source: '' },  // Empty string removes the key
});
```

**Use metadata for**: linking Stripe objects to your database IDs, tracking attribution, passing context through webhooks.

## Test Cards

| Number | Scenario |
|--------|----------|
| `4242424242424242` | Succeeds |
| `4000000000000002` | Declined (generic) |
| `4000000000009995` | Insufficient funds |
| `4000000000009987` | Lost card |
| `4000000000009979` | Stolen card |
| `4000002500003155` | Requires 3D Secure |
| `4000000000000341` | Attach succeeds, charge fails |
| `4000003560000123` | 3DS required on all transactions |

**Expiry**: Any future date. **CVC**: Any 3 digits. **ZIP**: Any 5 digits.

## API Versioning

```typescript
const stripe = new Stripe(process.env.STRIPE_SECRET_KEY!, {
  apiVersion: '2024-12-18.acacia', // Pin your version
});
```

Always pin your API version. Stripe deprecates but never removes API versions. Upgrade by:
1. Reading the changelog for breaking changes
2. Updating the version in your code
3. Testing in test mode
4. Deploying

## Idempotency

```typescript
// Prevent duplicate charges on retry
const paymentIntent = await stripe.paymentIntents.create(
  {
    amount: 2000,
    currency: 'usd',
    customer: 'cus_...',
  },
  {
    idempotencyKey: `order_${orderId}`, // Same key = same result
  }
);
```

Idempotency keys expire after 24 hours. Use deterministic keys derived from your business logic (order ID, user ID + action, etc.).

## Expanding Objects

```typescript
// By default, related objects are just IDs
const subscription = await stripe.subscriptions.retrieve('sub_...');
// subscription.customer === 'cus_123' (just an ID)

// Expand to get full objects inline
const subscription = await stripe.subscriptions.retrieve('sub_...', {
  expand: ['customer', 'latest_invoice.payment_intent'],
});
// subscription.customer === { id: 'cus_123', email: '...', ... }
```

Max 4 levels of expansion. Use expand to reduce API calls.

## Webhook Event Types (Complete List)

### Checkout
- `checkout.session.completed` — Purchase/subscription started
- `checkout.session.expired` — Session expired (24hr default)

### Subscriptions
- `customer.subscription.created` — New subscription
- `customer.subscription.updated` — Plan change, status change
- `customer.subscription.deleted` — Subscription cancelled
- `customer.subscription.paused` — Subscription paused
- `customer.subscription.resumed` — Subscription resumed
- `customer.subscription.trial_will_end` — 3 days before trial ends

### Invoices
- `invoice.created` — Invoice drafted
- `invoice.finalized` — Invoice ready for payment
- `invoice.payment_succeeded` — Payment collected
- `invoice.payment_failed` — Payment failed
- `invoice.upcoming` — Next invoice preview (7 days before)

### Payments
- `payment_intent.succeeded` — Payment completed
- `payment_intent.payment_failed` — Payment failed
- `charge.refunded` — Refund processed
- `charge.dispute.created` — Chargeback opened

### Connect
- `account.updated` — Connected account changed
- `payout.paid` — Payout delivered
- `payout.failed` — Payout failed
- `application_fee.created` — Platform fee collected
