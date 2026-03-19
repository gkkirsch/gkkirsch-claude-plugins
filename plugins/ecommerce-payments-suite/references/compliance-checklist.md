# E-commerce & Payment Compliance Checklist

## PCI DSS Compliance

PCI DSS (Payment Card Industry Data Security Standard) applies to anyone handling card data.

### Using Stripe Elements or Checkout = SAQ A

If you use Stripe.js Elements, Checkout Sessions, or the Payment Element, card data never touches your server. You qualify for **SAQ A** (simplest self-assessment, ~20 questions).

**What you must still do:**

- [ ] Serve your payment pages over HTTPS (TLS 1.2+)
- [ ] Never log, store, or transmit raw card numbers
- [ ] Load Stripe.js from `js.stripe.com` (never self-host)
- [ ] Keep Stripe SDK updated
- [ ] Use CSP headers to restrict script sources on payment pages
- [ ] Complete SAQ A annually (via Stripe Dashboard → Compliance)

### What Breaks SAQ A Eligibility

- Accepting card numbers in your own form fields (not Elements)
- Storing card data in your database
- Logging request bodies that contain card numbers
- Using `<input>` fields for card data instead of Stripe Elements
- Proxying or intercepting card data on your server

### Key PCI Rules

| Rule | What It Means |
|------|--------------|
| Never store CVV/CVC | After authorization, CVV must be deleted. Stripe handles this. |
| Never store full PAN in logs | Don't `console.log(req.body)` on payment endpoints |
| Use tokenization | Stripe Elements tokenize cards client-side |
| Encrypt transmission | Always HTTPS. No HTTP fallback. |
| Restrict access | Stripe API keys in env vars, not code. Rotate keys periodically. |

## SCA / PSD2 (European Customers)

Strong Customer Authentication (SCA) is required for European payments.

### What Triggers SCA

- Card payments in the EEA (European Economic Area)
- Transactions over 30 EUR (though banks can request for any amount)

### How Stripe Handles It

Stripe automatically triggers 3D Secure when required:

```typescript
// Use automatic_payment_methods for SCA compliance
const paymentIntent = await stripe.paymentIntents.create({
  amount: 2000,
  currency: 'eur',
  automatic_payment_methods: { enabled: true }, // Handles SCA automatically
});
```

**Checkout Sessions and Payment Element handle SCA automatically.** If you use the Card Element, you must handle `requires_action` status:

```typescript
const { error, paymentIntent } = await stripe.confirmCardPayment(clientSecret);

if (error) {
  // 3D Secure authentication failed or was abandoned
} else if (paymentIntent.status === 'requires_action') {
  // Stripe.js handles the 3D Secure popup automatically
} else if (paymentIntent.status === 'succeeded') {
  // Payment complete
}
```

### SCA Exemptions

| Exemption | When It Applies |
|-----------|----------------|
| Low value | Transactions < 30 EUR (up to cumulative 100 EUR or 5 transactions) |
| Low risk (TRA) | Stripe Radar's risk assessment qualifies the transaction |
| Recurring | Subsequent payments on a subscription (first payment needs SCA) |
| Merchant-initiated | Off-session charges using saved payment methods |
| Corporate cards | B2B corporate payment cards |

## Tax Compliance

### Sales Tax (US)

In the US, sales tax rules vary by state (nexus, rates, taxability).

#### Stripe Tax (Recommended)

```typescript
const session = await stripe.checkout.sessions.create({
  line_items: [{ price: 'price_xxx', quantity: 1 }],
  automatic_tax: { enabled: true },  // Stripe calculates tax
  // ...
});
```

Stripe Tax automatically:
- Determines if tax applies based on customer location
- Calculates the correct rate
- Handles digital goods vs physical goods
- Generates tax reports

**Cost:** 0.5% of transaction volume (on top of payment processing fees).

#### Manual Tax Calculation

If not using Stripe Tax, you need a tax service:

| Service | Pricing | Best For |
|---------|---------|----------|
| **Stripe Tax** | 0.5% of volume | Already on Stripe, simplest integration |
| **TaxJar** | From $19/mo | US-focused, Shopify integration |
| **Avalara** | Enterprise pricing | Complex scenarios, international |

### VAT (EU / International)

For digital products sold to EU customers:

```typescript
// Collect and validate VAT ID
const session = await stripe.checkout.sessions.create({
  // ...
  customer_update: { address: 'auto' },
  tax_id_collection: { enabled: true },  // Let customers enter VAT ID
  automatic_tax: { enabled: true },
});
```

**B2B:** Valid EU VAT ID = reverse charge (no VAT collected).
**B2C:** Charge VAT at customer's country rate.

## Refund Policies

### Legal Requirements

| Region | Requirement |
|--------|------------|
| **US** | No federal requirement (state laws vary). Must honor stated policy. |
| **EU** | 14-day cooling-off period for online purchases (digital goods exempt once "consumed") |
| **UK** | 14-day cancellation right (Consumer Contracts Regulations) |
| **Australia** | Refund for faulty/not-as-described products (Australian Consumer Law) |

### Best Practice Refund Policy

```markdown
## Refund Policy

**30-Day Money-Back Guarantee**

If you're not satisfied, request a refund within 30 days of purchase
for a full refund. No questions asked.

**Subscriptions:** Cancel anytime. You'll retain access until the end
of your current billing period. We don't issue partial-month refunds
for subscription cancellations.

**How to request a refund:** Email support@example.com with your
order number.
```

### Stripe Refund Implementation

```typescript
// Full refund
await stripe.refunds.create({
  payment_intent: 'pi_xxx',
  reason: 'requested_by_customer',
});

// Partial refund
await stripe.refunds.create({
  payment_intent: 'pi_xxx',
  amount: 500,  // $5.00 partial refund
});
```

**Note:** Stripe does NOT refund processing fees. On a $100 charge with $3.20 in fees, a full refund returns $100 to the customer, but you lose the $3.20 in fees.

## Data Privacy (GDPR / CCPA)

### Required for E-commerce

- [ ] **Privacy policy** — What data you collect, why, how long, who has access
- [ ] **Cookie consent** — Banner/popup for non-essential cookies (EU)
- [ ] **Terms of service** — Purchase terms, liability, dispute resolution
- [ ] **Data deletion** — Ability for users to request data deletion
- [ ] **Data portability** — Ability for users to export their data
- [ ] **Consent for marketing** — Explicit opt-in for email marketing (EU)
- [ ] **Data breach notification** — Process to notify users within 72 hours (GDPR)

### Stripe-Specific GDPR

Stripe is a data processor (you are the controller). Stripe's DPA covers GDPR compliance for payment data. You must still:

- Have your own privacy policy mentioning Stripe as a processor
- Honor deletion requests (delete customer from your DB AND Stripe):

```typescript
// Delete customer data
await stripe.customers.del('cus_xxx');
await db.user.delete({ where: { stripeCustomerId: 'cus_xxx' } });
```

## Email Requirements

### Transactional Emails (Required)

| Email | When | Required Content |
|-------|------|-----------------|
| Order confirmation | After purchase | Order details, total, items |
| Payment receipt | After charge | Amount, last 4 digits, date |
| Shipping confirmation | After dispatch | Tracking number, carrier |
| Subscription renewal | After recurring charge | Amount, next renewal date |
| Failed payment | After payment failure | What happened, how to update payment |
| Refund confirmation | After refund processed | Amount, timeline for funds |
| Account deletion | After data deletion | Confirmation, what was deleted |

### CAN-SPAM / GDPR Email Rules

- [ ] Transactional emails don't need unsubscribe (but it's good practice)
- [ ] Marketing emails MUST have unsubscribe link
- [ ] Include physical mailing address in marketing emails (CAN-SPAM)
- [ ] Honor unsubscribe within 10 business days
- [ ] Don't pre-check marketing consent boxes (GDPR)
- [ ] Keep consent records (when, how, what they agreed to)

## Accessibility (Payment UX)

- [ ] Payment forms work with keyboard only (tab through all fields)
- [ ] Error messages are announced to screen readers
- [ ] Color is not the only indicator of errors (use icons + text)
- [ ] Loading states are announced (`aria-live="polite"`)
- [ ] Stripe Elements have proper labels
- [ ] Price displays use proper currency formatting
- [ ] Order summary is readable at 200% zoom

## Dispute / Chargeback Handling

### Prevention

```typescript
// 1. Clear billing descriptor
const paymentIntent = await stripe.paymentIntents.create({
  amount: 2000,
  currency: 'usd',
  statement_descriptor: 'MYSHOP ORDER',      // 22 char max
  statement_descriptor_suffix: '#12345',      // Appended
});

// 2. Send receipt immediately
// 3. Clear refund policy on checkout page
// 4. Deliver what was promised, when promised
```

### Responding to Disputes

```typescript
// Webhook: charge.dispute.created
app.post('/webhook', (req, res) => {
  if (event.type === 'charge.dispute.created') {
    const dispute = event.data.object as Stripe.Dispute;
    // Gather evidence and respond within 7-21 days
    // Evidence: receipt, delivery proof, customer communication, terms of service
  }
});
```

**Dispute evidence to collect:**
- Customer email + name
- IP address + geolocation at purchase time
- Order confirmation email sent
- Delivery tracking/confirmation
- Product description matching what was delivered
- Terms of service the customer agreed to
- Any customer communication acknowledging receipt

## Launch Checklist

### Before Going Live

- [ ] Switch from Stripe test keys to live keys
- [ ] Verify webhook endpoints are receiving live events
- [ ] Test a real $1 purchase (then refund)
- [ ] Privacy policy published and linked from checkout
- [ ] Terms of service published and linked from checkout
- [ ] Refund policy published and linked from checkout
- [ ] SSL certificate valid and auto-renewing
- [ ] Error monitoring set up (Sentry, etc.)
- [ ] Payment failure alerts configured
- [ ] Dispute alerts configured in Stripe Dashboard
- [ ] Tax calculation verified for your markets
- [ ] Receipt emails sending correctly
- [ ] Customer support email/channel working
- [ ] Stripe Radar enabled for fraud detection
- [ ] Rate limiting on payment endpoints
- [ ] CORS configured to allow only your domain
