# Stripe API Quick Reference

## Authentication

```bash
# All requests need the secret key
curl https://api.stripe.com/v1/charges \
  -u sk_test_xxx:
```

```typescript
import Stripe from 'stripe';
const stripe = new Stripe(process.env.STRIPE_SECRET_KEY!);
```

## Core Objects

### Customer

```typescript
// Create
const customer = await stripe.customers.create({
  email: 'user@example.com',
  name: 'Jane Doe',
  metadata: { userId: 'usr_123' },
});

// Retrieve
const customer = await stripe.customers.retrieve('cus_xxx');

// Update
await stripe.customers.update('cus_xxx', { name: 'Jane Smith' });

// List
const customers = await stripe.customers.list({ limit: 10, email: 'user@example.com' });

// Delete
await stripe.customers.del('cus_xxx');
```

### Payment Intent

```typescript
// Create (server-side)
const paymentIntent = await stripe.paymentIntents.create({
  amount: 2000,           // $20.00 in cents
  currency: 'usd',
  customer: 'cus_xxx',
  payment_method_types: ['card'],
  metadata: { orderId: 'ord_123' },
});
// Returns: client_secret for frontend confirmation

// Confirm (usually done client-side via Elements)
const confirmed = await stripe.paymentIntents.confirm('pi_xxx', {
  payment_method: 'pm_xxx',
});

// Capture (for auth-then-capture flow)
await stripe.paymentIntents.capture('pi_xxx');

// Cancel
await stripe.paymentIntents.cancel('pi_xxx');
```

**Payment Intent Statuses:**

| Status | Meaning |
|--------|---------|
| `requires_payment_method` | Created, awaiting payment method |
| `requires_confirmation` | Has payment method, needs confirmation |
| `requires_action` | Needs customer action (3D Secure, etc.) |
| `processing` | Payment is processing |
| `requires_capture` | Authorized, awaiting capture |
| `succeeded` | Payment completed |
| `canceled` | Canceled by you |

### Checkout Session

```typescript
// Create (redirect-based, Stripe-hosted page)
const session = await stripe.checkout.sessions.create({
  mode: 'payment',  // 'payment' | 'subscription' | 'setup'
  line_items: [
    {
      price_data: {
        currency: 'usd',
        product_data: { name: 'T-shirt' },
        unit_amount: 2000,
      },
      quantity: 1,
    },
  ],
  success_url: 'https://example.com/success?session_id={CHECKOUT_SESSION_ID}',
  cancel_url: 'https://example.com/cancel',
  customer_email: 'user@example.com',  // or customer: 'cus_xxx'
  metadata: { orderId: 'ord_123' },
});
// Redirect user to session.url
```

### Subscription

```typescript
// Create
const subscription = await stripe.subscriptions.create({
  customer: 'cus_xxx',
  items: [{ price: 'price_xxx' }],
  trial_period_days: 14,
  payment_behavior: 'default_incomplete',  // don't charge until payment method works
  expand: ['latest_invoice.payment_intent'],
});

// Update (change plan)
await stripe.subscriptions.update('sub_xxx', {
  items: [{
    id: 'si_xxx',  // subscription item ID
    price: 'price_new',
  }],
  proration_behavior: 'create_prorations',
});

// Cancel
await stripe.subscriptions.update('sub_xxx', { cancel_at_period_end: true });
// or immediately:
await stripe.subscriptions.cancel('sub_xxx');
```

### Product & Price

```typescript
// Create product
const product = await stripe.products.create({
  name: 'Pro Plan',
  description: 'Full access to all features',
});

// Create recurring price
const price = await stripe.prices.create({
  product: product.id,
  unit_amount: 2900,   // $29.00
  currency: 'usd',
  recurring: { interval: 'month' },  // 'day' | 'week' | 'month' | 'year'
});

// Create one-time price
const oneTimePrice = await stripe.prices.create({
  product: product.id,
  unit_amount: 9900,
  currency: 'usd',
});
```

### Refund

```typescript
// Full refund
await stripe.refunds.create({ payment_intent: 'pi_xxx' });

// Partial refund
await stripe.refunds.create({ payment_intent: 'pi_xxx', amount: 500 }); // $5.00

// Refund reasons: 'duplicate' | 'fraudulent' | 'requested_by_customer'
await stripe.refunds.create({
  payment_intent: 'pi_xxx',
  reason: 'requested_by_customer',
});
```

### Invoice

```typescript
// Create manual invoice
const invoice = await stripe.invoices.create({
  customer: 'cus_xxx',
  collection_method: 'send_invoice',
  days_until_due: 30,
});

// Add line items
await stripe.invoiceItems.create({
  customer: 'cus_xxx',
  invoice: invoice.id,
  amount: 5000,
  currency: 'usd',
  description: 'Consulting (1 hour)',
});

// Finalize and send
await stripe.invoices.finalizeInvoice(invoice.id);
await stripe.invoices.sendInvoice(invoice.id);
```

## Webhooks

### Setup

```typescript
import express from 'express';

// IMPORTANT: Use raw body for webhook verification
app.post('/webhooks/stripe',
  express.raw({ type: 'application/json' }),
  (req, res) => {
    const sig = req.headers['stripe-signature']!;
    let event: Stripe.Event;

    try {
      event = stripe.webhooks.constructEvent(
        req.body,           // raw Buffer, NOT parsed JSON
        sig,
        process.env.STRIPE_WEBHOOK_SECRET!
      );
    } catch (err) {
      return res.status(400).send(`Webhook Error: ${err.message}`);
    }

    // Handle event
    switch (event.type) {
      case 'checkout.session.completed':
        const session = event.data.object as Stripe.Checkout.Session;
        // Fulfill the order
        break;
      case 'payment_intent.succeeded':
        const pi = event.data.object as Stripe.PaymentIntent;
        break;
      // ... more handlers
    }

    res.json({ received: true });
  }
);
```

### Essential Webhook Events

| Event | When to Handle |
|-------|---------------|
| `checkout.session.completed` | Customer completed checkout |
| `payment_intent.succeeded` | Payment confirmed |
| `payment_intent.payment_failed` | Payment failed |
| `customer.subscription.created` | New subscription started |
| `customer.subscription.updated` | Plan changed, status changed |
| `customer.subscription.deleted` | Subscription cancelled |
| `customer.subscription.trial_will_end` | Trial ending in 3 days |
| `invoice.paid` | Invoice payment succeeded |
| `invoice.payment_failed` | Invoice payment failed |
| `charge.refunded` | Refund processed |
| `charge.dispute.created` | Chargeback filed |

### Idempotency

```typescript
// Use idempotency keys to prevent duplicate charges
const paymentIntent = await stripe.paymentIntents.create(
  { amount: 2000, currency: 'usd' },
  { idempotencyKey: `order_${orderId}` }
);
```

## Frontend (Stripe.js + Elements)

### Card Element

```tsx
import { loadStripe } from '@stripe/stripe-js';
import { Elements, CardElement, useStripe, useElements } from '@stripe/react-stripe-js';

const stripePromise = loadStripe(process.env.NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY!);

function CheckoutForm() {
  const stripe = useStripe();
  const elements = useElements();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!stripe || !elements) return;

    const { error, paymentIntent } = await stripe.confirmCardPayment(clientSecret, {
      payment_method: { card: elements.getElement(CardElement)! },
    });

    if (error) {
      console.error(error.message);
    } else if (paymentIntent.status === 'succeeded') {
      // Payment successful
    }
  };

  return (
    <form onSubmit={handleSubmit}>
      <CardElement options={{
        style: {
          base: { fontSize: '16px', color: '#32325d' },
          invalid: { color: '#fa755a' },
        },
      }} />
      <button type="submit" disabled={!stripe}>Pay</button>
    </form>
  );
}

// Wrap in Elements provider
function App() {
  return (
    <Elements stripe={stripePromise} options={{ clientSecret }}>
      <CheckoutForm />
    </Elements>
  );
}
```

### Payment Element (newer, recommended)

```tsx
function PaymentForm() {
  const stripe = useStripe();
  const elements = useElements();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!stripe || !elements) return;

    const { error } = await stripe.confirmPayment({
      elements,
      confirmParams: { return_url: 'https://example.com/success' },
    });

    if (error) console.error(error.message);
  };

  return (
    <form onSubmit={handleSubmit}>
      <PaymentElement />
      <button type="submit" disabled={!stripe}>Pay</button>
    </form>
  );
}
```

## Stripe CLI (Testing)

```bash
# Install
brew install stripe/stripe-cli/stripe

# Login
stripe login

# Listen to webhooks locally
stripe listen --forward-to localhost:3000/webhooks/stripe

# Trigger test events
stripe trigger payment_intent.succeeded
stripe trigger checkout.session.completed
stripe trigger customer.subscription.created

# Create test resources
stripe customers create --email=test@example.com
stripe prices create --product=prod_xxx --unit-amount=2000 --currency=usd
```

## Test Card Numbers

| Card Number | Scenario |
|-------------|----------|
| `4242 4242 4242 4242` | Successful payment |
| `4000 0000 0000 3220` | 3D Secure required |
| `4000 0000 0000 9995` | Declined (insufficient funds) |
| `4000 0000 0000 0002` | Declined (generic) |
| `4000 0000 0000 0069` | Expired card |
| `4000 0000 0000 0127` | Incorrect CVC |
| `4000 0025 0000 3155` | SCA required (EU) |
| `4000 0000 0000 0341` | Attaching succeeds, but charge fails |

Use any future expiry date and any 3-digit CVC.

## Error Handling

```typescript
try {
  await stripe.paymentIntents.create({ amount: 2000, currency: 'usd' });
} catch (err) {
  if (err instanceof Stripe.errors.StripeCardError) {
    // Card declined
    console.error(err.message); // e.g., "Your card was declined."
    console.error(err.code);    // e.g., "card_declined"
    console.error(err.decline_code); // e.g., "insufficient_funds"
  } else if (err instanceof Stripe.errors.StripeRateLimitError) {
    // Too many requests — retry with backoff
  } else if (err instanceof Stripe.errors.StripeInvalidRequestError) {
    // Invalid parameters
  } else if (err instanceof Stripe.errors.StripeAuthenticationError) {
    // Wrong API key
  } else if (err instanceof Stripe.errors.StripeAPIError) {
    // Stripe server error — retry
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
  const list = await stripe.customers.list({
    limit: 100,
    ...(startingAfter && { starting_after: startingAfter }),
  });

  for (const customer of list.data) {
    // process
  }

  hasMore = list.has_more;
  startingAfter = list.data[list.data.length - 1]?.id;
}
```

## Metadata

Attach up to 50 key-value pairs (500 char keys, 500 char values) to any object:

```typescript
await stripe.customers.create({
  email: 'user@example.com',
  metadata: {
    userId: 'usr_123',
    plan: 'pro',
    source: 'organic',
  },
});
```

Metadata is searchable in the Stripe Dashboard and included in webhook payloads.
