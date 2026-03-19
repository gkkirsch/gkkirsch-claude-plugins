---
name: stripe-integration
description: >
  Set up Stripe payment processing in your web application. Covers Checkout Sessions,
  Payment Intents, Stripe Elements, webhook handling, and subscription billing.
  Works with Node.js, Express, Next.js, and React.
  Triggers: "add stripe", "stripe integration", "payment processing", "accept payments",
  "stripe checkout", "payment intent".
  NOT for: PayPal, Square, or other payment providers.
version: 1.0.0
argument-hint: "[checkout|subscription|connect|webhook]"
allowed-tools: Read, Grep, Glob, Write, Edit, Bash
---

# Stripe Integration

Set up Stripe payment processing step by step.

## Step 1: Install Dependencies

```bash
# Server
npm install stripe

# Client (React)
npm install @stripe/react-stripe-js @stripe/stripe-js
```

## Step 2: Environment Setup

```bash
# .env
STRIPE_SECRET_KEY=sk_test_...          # Server only
STRIPE_PUBLISHABLE_KEY=pk_test_...     # Client-safe
STRIPE_WEBHOOK_SECRET=whsec_...        # Webhook verification
```

**Never expose the secret key to the client.** The publishable key is safe for browsers.

## Step 3: Choose Your Integration

### Option A: Stripe Checkout (Fastest — Hosted Payment Page)

**Server:**
```typescript
import Stripe from 'stripe';

const stripe = new Stripe(process.env.STRIPE_SECRET_KEY!);

// Create Checkout Session
app.post('/api/checkout', async (req, res) => {
  const { priceId } = req.body;

  const session = await stripe.checkout.sessions.create({
    mode: 'payment',
    line_items: [{ price: priceId, quantity: 1 }],
    success_url: `${process.env.APP_URL}/success?session_id={CHECKOUT_SESSION_ID}`,
    cancel_url: `${process.env.APP_URL}/pricing`,
    customer_email: req.user?.email,
    metadata: { userId: req.user?.id },
  });

  res.json({ url: session.url });
});

// Verify session on success page
app.get('/api/checkout/verify', async (req, res) => {
  const { session_id } = req.query;
  const session = await stripe.checkout.sessions.retrieve(session_id as string);

  if (session.payment_status === 'paid') {
    res.json({ success: true, customerEmail: session.customer_email });
  } else {
    res.json({ success: false });
  }
});
```

**Client:**
```tsx
const handleCheckout = async (priceId: string) => {
  const res = await fetch('/api/checkout', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ priceId }),
  });
  const { url } = await res.json();
  window.location.href = url;
};
```

### Option B: Custom Checkout (Stripe Elements — Embedded Form)

**Server:**
```typescript
app.post('/api/payment-intent', async (req, res) => {
  const { amount, currency = 'usd' } = req.body;

  // Validate amount on server
  const validatedAmount = calculatePrice(req.body.items);

  const paymentIntent = await stripe.paymentIntents.create({
    amount: validatedAmount,
    currency,
    automatic_payment_methods: { enabled: true },
    metadata: { userId: req.user?.id },
  });

  res.json({ clientSecret: paymentIntent.client_secret });
});
```

**Client:**
```tsx
import { loadStripe } from '@stripe/stripe-js';
import { Elements, PaymentElement, useStripe, useElements } from '@stripe/react-stripe-js';

const stripePromise = loadStripe(process.env.NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY!);

function CheckoutForm() {
  const stripe = useStripe();
  const elements = useElements();
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!stripe || !elements) return;

    setLoading(true);
    setError('');

    const { error } = await stripe.confirmPayment({
      elements,
      confirmParams: {
        return_url: `${window.location.origin}/order/success`,
      },
    });

    if (error) {
      setError(error.message || 'An error occurred');
      setLoading(false);
    }
  };

  return (
    <form onSubmit={handleSubmit}>
      <PaymentElement />
      <button type="submit" disabled={!stripe || loading}>
        {loading ? 'Processing...' : 'Pay Now'}
      </button>
      {error && <p className="text-red-500">{error}</p>}
    </form>
  );
}

function PaymentPage() {
  const [clientSecret, setClientSecret] = useState('');

  useEffect(() => {
    fetch('/api/payment-intent', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ items: cart.items }),
    })
      .then((res) => res.json())
      .then(({ clientSecret }) => setClientSecret(clientSecret));
  }, []);

  if (!clientSecret) return <div>Loading...</div>;

  return (
    <Elements stripe={stripePromise} options={{ clientSecret }}>
      <CheckoutForm />
    </Elements>
  );
}
```

### Option C: Subscription Billing

```typescript
// Create subscription
app.post('/api/subscribe', async (req, res) => {
  const { priceId } = req.body;

  // Get or create Stripe customer
  let customer = await findCustomerByUserId(req.user.id);
  if (!customer) {
    const stripeCustomer = await stripe.customers.create({
      email: req.user.email,
      metadata: { userId: req.user.id },
    });
    customer = await saveCustomer(req.user.id, stripeCustomer.id);
  }

  const subscription = await stripe.subscriptions.create({
    customer: customer.stripeCustomerId,
    items: [{ price: priceId }],
    payment_behavior: 'default_incomplete',
    payment_settings: { save_default_payment_method: 'on_subscription' },
    expand: ['latest_invoice.payment_intent'],
  });

  const invoice = subscription.latest_invoice as Stripe.Invoice;
  const paymentIntent = invoice.payment_intent as Stripe.PaymentIntent;

  res.json({
    subscriptionId: subscription.id,
    clientSecret: paymentIntent?.client_secret,
  });
});

// Cancel subscription
app.post('/api/cancel-subscription', async (req, res) => {
  const { subscriptionId } = req.body;

  const subscription = await stripe.subscriptions.update(subscriptionId, {
    cancel_at_period_end: true,
  });

  res.json({ cancelAt: subscription.cancel_at });
});
```

## Step 4: Set Up Webhooks

```typescript
import { buffer } from 'micro';

export const config = { api: { bodyParser: false } };

app.post('/api/webhooks/stripe', async (req, res) => {
  const buf = await buffer(req);
  const sig = req.headers['stripe-signature']!;

  let event: Stripe.Event;
  try {
    event = stripe.webhooks.constructEvent(buf, sig, process.env.STRIPE_WEBHOOK_SECRET!);
  } catch (err) {
    return res.status(400).send(`Webhook Error: ${err}`);
  }

  // Idempotency check
  if (await isProcessed(event.id)) return res.json({ received: true });

  switch (event.type) {
    case 'checkout.session.completed':
      const session = event.data.object as Stripe.Checkout.Session;
      await fulfillOrder(session);
      break;

    case 'invoice.paid':
      const invoice = event.data.object as Stripe.Invoice;
      await activateSubscription(invoice);
      break;

    case 'invoice.payment_failed':
      const failedInvoice = event.data.object as Stripe.Invoice;
      await handleFailedPayment(failedInvoice);
      break;

    case 'customer.subscription.deleted':
      const sub = event.data.object as Stripe.Subscription;
      await deactivateSubscription(sub);
      break;
  }

  await markProcessed(event.id);
  res.json({ received: true });
});
```

## Step 5: Stripe CLI for Testing

```bash
# Install Stripe CLI
brew install stripe/stripe-cli/stripe

# Login
stripe login

# Listen for webhooks locally
stripe listen --forward-to localhost:3000/api/webhooks/stripe

# Trigger test events
stripe trigger payment_intent.succeeded
stripe trigger customer.subscription.created
stripe trigger invoice.payment_failed
```

## Step 6: Create Products and Prices

```bash
# Via Stripe Dashboard (easiest) or CLI:
stripe products create --name="Pro Plan" --description="Full access"
stripe prices create --product=prod_xxx --unit-amount=2900 --currency=usd --recurring[interval]=month
```

## Testing Cards

| Number | Scenario |
|--------|----------|
| `4242424242424242` | Successful payment |
| `4000000000003220` | 3D Secure required |
| `4000000000009995` | Declined (insufficient funds) |
| `4000000000000002` | Declined (generic) |
| `4000002500003155` | Requires authentication |

Use any future expiration date (e.g., 12/34), any 3-digit CVC, any postal code.