---
name: stripe-checkout
description: >
  Set up Stripe Checkout for one-time payments and subscriptions —
  hosted checkout, embedded checkout, pricing tables, customer portal,
  and success/cancel handling. Production-ready with TypeScript.
  Triggers: "stripe checkout", "payment page", "buy button", "pricing page",
  "stripe integration", "accept payments", "payment processing".
  NOT for: custom payment forms (use Stripe Elements), marketplace payments (use stripe-connect).
version: 1.0.0
allowed-tools: Read, Grep, Glob, Bash, Edit, Write
---

# Stripe Checkout Integration

## Initial Setup

### 1. Install Dependencies

```bash
npm install stripe
npm install -D @types/stripe  # if not using stripe's built-in types
```

### 2. Initialize Stripe

```typescript
// src/lib/stripe.ts
import Stripe from 'stripe';

export const stripe = new Stripe(process.env.STRIPE_SECRET_KEY!, {
  apiVersion: '2024-12-18.acacia',  // Pin API version
  typescript: true,
});

// Client-side: load Stripe.js
// Add to your HTML/layout:
// <script src="https://js.stripe.com/v3/"></script>
// Or npm install @stripe/stripe-js
```

```typescript
// src/lib/stripe-client.ts (for React/Next.js)
import { loadStripe } from '@stripe/stripe-js';

export const stripePromise = loadStripe(process.env.NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY!);
```

### 3. Create Products & Prices in Stripe Dashboard

```
Dashboard → Products → Add product
  - Name: "Pro Plan"
  - Pricing: $29/month (recurring) or $99 (one-time)
  - Save the Price ID (price_xxx)
```

Or via API:

```typescript
// One-time product
const product = await stripe.products.create({
  name: 'Digital Course',
  description: 'Complete guide to building SaaS',
});

const price = await stripe.prices.create({
  product: product.id,
  unit_amount: 9900,  // $99.00 in cents
  currency: 'usd',
});

// Subscription product
const subPrice = await stripe.prices.create({
  product: product.id,
  unit_amount: 2900,  // $29.00/month
  currency: 'usd',
  recurring: { interval: 'month' },
});
```

---

## Hosted Checkout (Recommended)

### One-Time Payment

```typescript
// POST /api/checkout
import { stripe } from '../lib/stripe';

router.post('/api/checkout', async (req, res) => {
  const { priceId, email } = req.body;

  const session = await stripe.checkout.sessions.create({
    mode: 'payment',
    line_items: [
      {
        price: priceId,
        quantity: 1,
      },
    ],
    customer_email: email,  // Pre-fill email
    success_url: `${process.env.APP_URL}/success?session_id={CHECKOUT_SESSION_ID}`,
    cancel_url: `${process.env.APP_URL}/pricing`,
    metadata: {
      userId: req.user?.id,  // Track who bought it
    },
    // Tax calculation (optional)
    automatic_tax: { enabled: true },
    // Collect shipping address (physical goods)
    // shipping_address_collection: { allowed_countries: ['US', 'CA'] },
  });

  res.json({ url: session.url });
});
```

### Subscription Checkout

```typescript
// POST /api/checkout/subscribe
router.post('/api/checkout/subscribe', async (req, res) => {
  const { priceId, email, userId } = req.body;

  // Get or create Stripe customer
  let customerId = await getStripeCustomerId(userId);
  if (!customerId) {
    const customer = await stripe.customers.create({
      email,
      metadata: { userId },
    });
    customerId = customer.id;
    await saveStripeCustomerId(userId, customerId);
  }

  const session = await stripe.checkout.sessions.create({
    mode: 'subscription',
    customer: customerId,
    line_items: [
      {
        price: priceId,
        quantity: 1,
      },
    ],
    subscription_data: {
      trial_period_days: 14,  // Optional free trial
      metadata: { userId },
    },
    success_url: `${process.env.APP_URL}/dashboard?session_id={CHECKOUT_SESSION_ID}`,
    cancel_url: `${process.env.APP_URL}/pricing`,
    allow_promotion_codes: true,  // Let users enter coupon codes
    // Tax
    automatic_tax: { enabled: true },
    tax_id_collection: { enabled: true },
  });

  res.json({ url: session.url });
});
```

### Client-Side Redirect

```typescript
// React component
function PricingCard({ priceId, name, amount }: PricingCardProps) {
  const [loading, setLoading] = useState(false);

  const handleCheckout = async () => {
    setLoading(true);
    try {
      const response = await fetch('/api/checkout/subscribe', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ priceId }),
      });
      const { url } = await response.json();
      window.location.href = url;  // Redirect to Stripe
    } catch (error) {
      console.error('Checkout error:', error);
      setLoading(false);
    }
  };

  return (
    <div className="pricing-card">
      <h3>{name}</h3>
      <p>${(amount / 100).toFixed(2)}/mo</p>
      <button onClick={handleCheckout} disabled={loading}>
        {loading ? 'Loading...' : 'Subscribe'}
      </button>
    </div>
  );
}
```

---

## Embedded Checkout

Keep users on your site with an embedded checkout form:

```typescript
// Server: create session with embedded mode
router.post('/api/checkout/embedded', async (req, res) => {
  const session = await stripe.checkout.sessions.create({
    mode: 'payment',
    line_items: [{ price: req.body.priceId, quantity: 1 }],
    ui_mode: 'embedded',
    return_url: `${process.env.APP_URL}/checkout/complete?session_id={CHECKOUT_SESSION_ID}`,
  });

  res.json({ clientSecret: session.client_secret });
});
```

```typescript
// Client: React component
import { loadStripe } from '@stripe/stripe-js';
import { EmbeddedCheckoutProvider, EmbeddedCheckout } from '@stripe/react-stripe-js';

const stripePromise = loadStripe(process.env.NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY!);

function CheckoutPage() {
  const fetchClientSecret = async () => {
    const response = await fetch('/api/checkout/embedded', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ priceId: 'price_xxx' }),
    });
    const { clientSecret } = await response.json();
    return clientSecret;
  };

  return (
    <EmbeddedCheckoutProvider stripe={stripePromise} options={{ fetchClientSecret }}>
      <EmbeddedCheckout />
    </EmbeddedCheckoutProvider>
  );
}
```

---

## Success Page

```typescript
// GET /checkout/complete
router.get('/checkout/complete', async (req, res) => {
  const sessionId = req.query.session_id as string;

  if (!sessionId) {
    return res.redirect('/pricing');
  }

  const session = await stripe.checkout.sessions.retrieve(sessionId, {
    expand: ['customer', 'line_items'],
  });

  if (session.payment_status !== 'paid') {
    return res.redirect('/pricing?error=payment-failed');
  }

  // Show success page with order details
  res.render('success', {
    customerEmail: (session.customer as Stripe.Customer)?.email,
    amountTotal: session.amount_total,
    currency: session.currency,
  });
});
```

---

## Customer Billing Portal

Let customers manage their own subscriptions:

```typescript
// POST /api/billing-portal
router.post('/api/billing-portal', requireAuth, async (req, res) => {
  const customerId = await getStripeCustomerId(req.user.id);
  if (!customerId) {
    return res.status(400).json({ error: 'No billing account found' });
  }

  const portalSession = await stripe.billingPortal.sessions.create({
    customer: customerId,
    return_url: `${process.env.APP_URL}/dashboard`,
  });

  res.json({ url: portalSession.url });
});
```

Configure portal features in Dashboard → Settings → Billing → Customer portal:
- Update payment method ✓
- View invoices ✓
- Cancel subscription ✓
- Switch plans ✓ (list allowed prices)
- Update billing address ✓

---

## Pricing Table (No-Code Option)

Embed Stripe's pre-built pricing table — zero backend code:

```html
<!-- From Dashboard → Products → Pricing table → Get embed code -->
<script async src="https://js.stripe.com/v3/pricing-table.js"></script>
<stripe-pricing-table
  pricing-table-id="prctbl_xxx"
  publishable-key="pk_live_xxx"
  client-reference-id="user_123"
>
</stripe-pricing-table>
```

---

## Payment Links (Simplest Option)

Create shareable payment links — no code at all:

```typescript
// Or create via API for dynamic links
const paymentLink = await stripe.paymentLinks.create({
  line_items: [{ price: 'price_xxx', quantity: 1 }],
  after_completion: {
    type: 'redirect',
    redirect: { url: `${process.env.APP_URL}/success` },
  },
});

console.log(paymentLink.url);
// → https://buy.stripe.com/xxx — share anywhere
```

---

## Coupon / Promotion Codes

```typescript
// Create a coupon
const coupon = await stripe.coupons.create({
  percent_off: 20,
  duration: 'once',  // 'once' | 'repeating' | 'forever'
  name: 'Launch Discount',
});

// Create a promotion code (user-facing code string)
const promoCode = await stripe.promotionCodes.create({
  coupon: coupon.id,
  code: 'LAUNCH20',
  max_redemptions: 100,
  expires_at: Math.floor(new Date('2026-12-31').getTime() / 1000),
});

// Enable in checkout:
// allow_promotion_codes: true  (already shown above)
```

---

## Environment Variables

```bash
# .env
STRIPE_SECRET_KEY=sk_test_xxx          # Server-side only
STRIPE_PUBLISHABLE_KEY=pk_test_xxx     # Client-side (NEXT_PUBLIC_ prefix for Next.js)
STRIPE_WEBHOOK_SECRET=whsec_xxx        # Webhook signature verification
APP_URL=http://localhost:3000           # For redirect URLs
```

---

## Testing with Stripe CLI

```bash
# Install
brew install stripe/stripe-cli/stripe

# Login
stripe login

# Forward webhooks to local server
stripe listen --forward-to localhost:3000/api/webhooks/stripe

# Trigger specific events
stripe trigger checkout.session.completed
stripe trigger customer.subscription.created
stripe trigger invoice.payment_failed
```

### Test Cards

| Card Number | Scenario |
|------------|----------|
| `4242 4242 4242 4242` | Successful payment |
| `4000 0000 0000 3220` | 3D Secure required |
| `4000 0000 0000 9995` | Declined |
| `4000 0000 0000 0341` | Attach fails |
| `4000 0025 0000 3155` | SCA required |

Use any future expiry date, any 3-digit CVC, any 5-digit ZIP.

---

## Common Gotchas

1. **Always use webhooks for fulfillment** — Don't rely on the redirect to your success URL. The redirect can fail. Webhooks are the source of truth.

2. **Pin your API version** — Stripe makes breaking changes. Pin the version in your Stripe initialization.

3. **Use metadata everywhere** — Set `metadata: { userId, orderId }` on sessions, subscriptions, and customers. It's free and invaluable for debugging.

4. **Test mode vs live mode** — Different API keys, different webhook secrets. Use env vars to switch.

5. **Checkout sessions expire** — After 24 hours by default. Don't store session IDs long-term.

6. **Customer email vs customer object** — Use `customer` (existing customer ID) OR `customer_email` (create new), not both.

7. **Amounts are in cents** — `unit_amount: 2900` = $29.00. Always multiply by 100.

8. **Currency matters** — JPY and other zero-decimal currencies don't use cents. Check Stripe docs for your currency.
