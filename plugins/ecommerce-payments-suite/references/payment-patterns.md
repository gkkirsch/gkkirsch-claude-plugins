# Payment Integration Patterns

## Pattern 1: Simple Checkout (Stripe Hosted)

Best for: MVPs, simple products, getting to market fast.

```
User clicks "Buy" → Server creates Checkout Session → Redirect to Stripe →
Stripe handles payment → Redirect back to success URL → Webhook confirms payment
```

**Pros:** No PCI scope. Stripe handles all UI, validation, 3D Secure, Apple Pay, Google Pay.
**Cons:** User leaves your site. Limited customization.

```typescript
// Server
app.post('/api/create-checkout', async (req, res) => {
  const session = await stripe.checkout.sessions.create({
    mode: 'payment',
    line_items: req.body.items.map(item => ({
      price_data: {
        currency: 'usd',
        product_data: { name: item.name, images: [item.image] },
        unit_amount: item.price,
      },
      quantity: item.quantity,
    })),
    success_url: `${process.env.APP_URL}/order/success?session_id={CHECKOUT_SESSION_ID}`,
    cancel_url: `${process.env.APP_URL}/cart`,
  });
  res.json({ url: session.url });
});

// Client
const handleCheckout = async () => {
  const res = await fetch('/api/create-checkout', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ items: cartItems }),
  });
  const { url } = await res.json();
  window.location.href = url;
};
```

## Pattern 2: Custom Payment Form (Elements)

Best for: Branded checkout, custom UX, staying on your site.

```
User fills form on your site → Elements tokenizes card →
Server creates PaymentIntent → Client confirms with Stripe.js →
Webhook confirms payment
```

**Pros:** Full control over UX. User stays on your site.
**Cons:** More code. Must handle 3D Secure, errors, loading states.

### The Two-Step Flow

```typescript
// Step 1: Server creates PaymentIntent
app.post('/api/create-payment-intent', async (req, res) => {
  const { amount, currency = 'usd' } = req.body;

  const paymentIntent = await stripe.paymentIntents.create({
    amount,
    currency,
    automatic_payment_methods: { enabled: true },
  });

  res.json({ clientSecret: paymentIntent.client_secret });
});

// Step 2: Client confirms payment
const { error } = await stripe.confirmPayment({
  elements,
  confirmParams: {
    return_url: `${window.location.origin}/order/success`,
  },
});
```

## Pattern 3: Auth-Then-Capture (Hold and Charge Later)

Best for: Marketplaces, hotels, pre-orders — where you authorize now but charge later.

```typescript
// Step 1: Authorize (hold funds)
const paymentIntent = await stripe.paymentIntents.create({
  amount: 10000,
  currency: 'usd',
  capture_method: 'manual',  // KEY: don't capture immediately
  payment_method: 'pm_xxx',
  confirm: true,
});
// Status: requires_capture

// Step 2: Capture later (within 7 days)
await stripe.paymentIntents.capture('pi_xxx');
// Or capture a different amount (partial capture):
await stripe.paymentIntents.capture('pi_xxx', { amount_to_capture: 8000 });

// Or cancel the hold:
await stripe.paymentIntents.cancel('pi_xxx');
```

**Important:** Uncaptured authorizations expire after 7 days. Set a reminder.

## Pattern 4: Subscription with Free Trial

Best for: SaaS products with trial periods.

```typescript
// Create subscription with trial (no payment method required initially)
const subscription = await stripe.subscriptions.create({
  customer: 'cus_xxx',
  items: [{ price: 'price_monthly_pro' }],
  trial_period_days: 14,
  trial_settings: {
    end_behavior: { missing_payment_method: 'cancel' },
  },
});

// 3 days before trial ends, Stripe sends:
// customer.subscription.trial_will_end webhook
// → Email user to add payment method

// At trial end:
// - If payment method exists → auto-charge, subscription becomes 'active'
// - If no payment method → subscription canceled (per trial_settings)
```

## Pattern 5: Metered / Usage-Based Billing

Best for: API products, cloud services, pay-as-you-go.

```typescript
// 1. Create metered price
const price = await stripe.prices.create({
  product: 'prod_xxx',
  currency: 'usd',
  recurring: {
    interval: 'month',
    usage_type: 'metered',
    aggregate_usage: 'sum',  // 'sum' | 'last_during_period' | 'last_ever' | 'max'
  },
  unit_amount: 1,  // $0.01 per unit
});

// 2. Report usage throughout the billing period
await stripe.subscriptionItems.createUsageRecord('si_xxx', {
  quantity: 100,    // 100 API calls
  timestamp: Math.floor(Date.now() / 1000),
  action: 'increment',
});

// 3. At period end, Stripe totals usage and charges automatically
```

## Pattern 6: Marketplace / Platform (Stripe Connect)

Best for: Multi-seller marketplaces, platforms taking a cut.

### Direct Charges (Platform charges, then transfers)

```typescript
// Onboard sellers (Connected Accounts)
const account = await stripe.accounts.create({
  type: 'express',  // 'standard' | 'express' | 'custom'
  country: 'US',
  email: 'seller@example.com',
  capabilities: {
    card_payments: { requested: true },
    transfers: { requested: true },
  },
});

// Create onboarding link
const accountLink = await stripe.accountLinks.create({
  account: account.id,
  refresh_url: `${APP_URL}/seller/onboarding/refresh`,
  return_url: `${APP_URL}/seller/onboarding/complete`,
  type: 'account_onboarding',
});

// Charge customer, split payment
const paymentIntent = await stripe.paymentIntents.create({
  amount: 10000,                    // $100.00
  currency: 'usd',
  application_fee_amount: 1500,     // Platform takes $15.00 (15%)
  transfer_data: {
    destination: 'acct_seller_xxx', // Seller gets $85.00
  },
});
```

### Destination Charges vs Direct Charges vs Separate Charges and Transfers

| Approach | Who Pays Stripe Fees | Refund Handling | Best For |
|----------|---------------------|-----------------|----------|
| **Destination** | Platform | Platform refunds, transfer auto-reversed | Marketplaces where platform owns relationship |
| **Direct** | Connected account | Connected account refunds | Platforms where seller owns relationship |
| **Separate** | Platform | Manual coordination | Complex multi-party payments |

## Pattern 7: One-Click Upsell (Post-Purchase)

Best for: Digital products, add-ons after initial purchase.

```typescript
// After successful checkout, offer upsell on success page
app.post('/api/upsell/accept', async (req, res) => {
  const { sessionId, upsellPriceId } = req.body;

  // Retrieve the checkout session to get customer and payment method
  const session = await stripe.checkout.sessions.retrieve(sessionId, {
    expand: ['payment_intent.payment_method'],
  });

  const paymentIntent = session.payment_intent as Stripe.PaymentIntent;
  const paymentMethodId = (paymentIntent.payment_method as Stripe.PaymentMethod).id;

  // Charge immediately using saved payment method
  const upsellPayment = await stripe.paymentIntents.create({
    amount: 2900,
    currency: 'usd',
    customer: session.customer as string,
    payment_method: paymentMethodId,
    off_session: true,
    confirm: true,
    metadata: {
      type: 'upsell',
      originalSession: sessionId,
    },
  });

  res.json({ success: true });
});
```

## Pattern 8: Saved Cards for Returning Customers

```typescript
// Save card during first purchase
const session = await stripe.checkout.sessions.create({
  mode: 'payment',
  payment_intent_data: {
    setup_future_usage: 'off_session',  // Save for future charges
  },
  // ...
});

// Charge saved card later
const paymentMethods = await stripe.paymentMethods.list({
  customer: 'cus_xxx',
  type: 'card',
});

const paymentIntent = await stripe.paymentIntents.create({
  amount: 2000,
  currency: 'usd',
  customer: 'cus_xxx',
  payment_method: paymentMethods.data[0].id,
  off_session: true,
  confirm: true,
});
```

## Anti-Patterns to Avoid

### 1. Storing Card Numbers

**Never do this.** Use Stripe Elements or Checkout. Your server should never see raw card numbers.

### 2. Not Using Webhooks

```typescript
// BAD: Relying on redirect for fulfillment
app.get('/success', (req, res) => {
  fulfillOrder(req.query.session_id); // User could fake this URL
});

// GOOD: Webhook-driven fulfillment
app.post('/webhook', (req, res) => {
  // Verified by Stripe signature
  if (event.type === 'checkout.session.completed') {
    fulfillOrder(event.data.object.id);
  }
});
```

### 3. No Idempotency

```typescript
// BAD: Double-clicking "Pay" creates duplicate charges
await stripe.paymentIntents.create({ amount: 2000, currency: 'usd' });

// GOOD: Idempotency key prevents duplicates
await stripe.paymentIntents.create(
  { amount: 2000, currency: 'usd' },
  { idempotencyKey: `checkout_${orderId}` }
);
```

### 4. Trusting Client-Side Amounts

```typescript
// BAD: Using price from request body
app.post('/pay', (req, res) => {
  stripe.paymentIntents.create({ amount: req.body.amount }); // User can modify!
});

// GOOD: Calculate price server-side
app.post('/pay', (req, res) => {
  const items = await db.product.findMany({ where: { id: { in: req.body.itemIds } } });
  const amount = items.reduce((sum, item) => sum + item.price, 0);
  stripe.paymentIntents.create({ amount });
});
```

### 5. Synchronous Fulfillment

```typescript
// BAD: Fulfill in the redirect handler
app.get('/success', async (req, res) => {
  await fulfillOrder(req.query.session_id); // Slow, might timeout
  res.render('success');
});

// GOOD: Queue fulfillment from webhook
app.post('/webhook', async (req, res) => {
  res.json({ received: true }); // Respond immediately
  // Then process asynchronously
  queue.add('fulfill-order', { sessionId: event.data.object.id });
});
```

## Security Checklist

- [ ] Never log or store raw card numbers, CVCs, or full card details
- [ ] Verify webhook signatures with `stripe.webhooks.constructEvent()`
- [ ] Use idempotency keys on all create/charge operations
- [ ] Calculate prices server-side, never trust client amounts
- [ ] Use HTTPS for all API communication
- [ ] Restrict API keys: use restricted keys with minimum permissions in production
- [ ] Store Stripe keys in environment variables, never in code
- [ ] Enable Stripe Radar for fraud detection
- [ ] Set up webhook failure alerts in Stripe Dashboard
- [ ] Log all payment events for audit trail
- [ ] Handle 3D Secure / SCA for European customers
- [ ] Use `automatic_payment_methods` instead of hardcoding payment types
