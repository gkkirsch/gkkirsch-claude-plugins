# Pricing & Monetization Strategies

## Pricing Models

### 1. Flat-Rate Pricing

One price, one product.

```
$29/month — Everything included
```

**When to use:** Simple product with clear value. Early-stage when you don't know enough for tiers.
**Examples:** Basecamp, Hey.com

### 2. Tiered Pricing

Good-better-best. The default for SaaS.

```
Free        → Basic features, limited usage
$19/mo      → More features, higher limits
$49/mo      → All features, priority support
$99/mo      → Team features, API access, SLA
```

**Best practices:**
- 3-4 tiers maximum (decision paralysis above 4)
- Name tiers by persona, not features: "Starter", "Pro", "Team", "Enterprise"
- Highlight the recommended plan (usually middle tier)
- Show annual billing with discount (15-20% off)

### 3. Usage-Based Pricing

Pay for what you use. Common for APIs and infrastructure.

```
$0.01 per API call
$0.10 per GB stored
$5 per 1,000 emails sent
```

**When to use:** Value scales with usage. Customers have wildly different usage patterns.
**Examples:** AWS, Twilio, SendGrid, Stripe itself

### 4. Per-Seat Pricing

Price per user or team member.

```
$10/user/month
```

**When to use:** Collaboration tools where value increases per user.
**Risk:** Teams find workarounds (shared logins) to avoid per-seat costs.
**Examples:** Slack, Notion, GitHub

### 5. Freemium

Free tier with paid upgrades.

```
Free   → 3 projects, 100MB storage
Pro    → Unlimited projects, 10GB, advanced features
```

**When to use:** Low marginal cost. Network effects. Wide top-of-funnel needed.
**Key metric:** Free-to-paid conversion rate (2-5% is typical, 10%+ is excellent).

### 6. Reverse Trial

Start users on the paid plan for free, then downgrade to free tier.

```
Day 1-14:  Full Pro features (free trial)
Day 15+:   Downgrade to Free tier OR subscribe to Pro
```

**Why it works:** Users experience premium features first. Loss aversion kicks in.
**Examples:** Notion, Ahrefs

## Pricing Psychology

### Anchoring

Put an expensive option first to make others seem reasonable.

```
Enterprise: $299/mo  ← Anchor
Pro:        $49/mo   ← This looks cheap by comparison
Starter:    $19/mo
```

### The Decoy Effect

Add a plan that makes one option clearly better.

```
Basic:    $9/mo  — 10 projects
Pro:      $29/mo — 100 projects   ← Most choose this
Business: $27/mo — 50 projects    ← Decoy (worse than Pro but similar price)
```

### Charm Pricing

$29 feels meaningfully cheaper than $30. Use .99 for consumer products, round numbers for premium/enterprise.

```
Consumer:   $9.99, $19.99, $29.99
B2B/SaaS:   $29, $49, $99
Enterprise: $299, $999 (or "Contact Us")
```

### Annual Discount

Offer 15-20% off for annual billing. Show the monthly equivalent.

```
Monthly: $29/mo
Annual:  $24/mo (billed $288/year) — Save $60!
```

**Implementation:**

```typescript
const prices = {
  monthly: 2900,  // $29.00
  annual: 28800,  // $288.00 ($24/mo equivalent, 17% off)
};
```

## Conversion Optimization

### Free Trial Best Practices

| Decision | Recommendation | Why |
|----------|---------------|-----|
| Trial length | 14 days | Long enough to evaluate, short enough for urgency |
| Credit card required? | Yes (for B2B SaaS) | Higher conversion rate (40-60% vs 2-5% without card) |
| Trial plan | Full-featured (reverse trial) | Users see maximum value |
| Trial ending email | 3 days before + day of | Prevent involuntary churn |

### Pricing Page Layout

```
┌─────────────────────────────────────────────────┐
│         Simple headline: "Pick your plan"        │
│         Monthly / Annual toggle                  │
├──────────┬───────────────┬──────────────────────┤
│ Starter  │  Pro ★        │  Enterprise          │
│ $19/mo   │  $49/mo       │  Contact Us          │
│          │  MOST POPULAR │                      │
│ Feature  │  Feature      │  Everything in Pro   │
│ Feature  │  Feature      │  + Custom features   │
│ Feature  │  Feature      │  + Dedicated support │
│          │  + Extra      │  + SLA               │
│          │  + Extra      │                      │
│ [Start]  │  [Start Free] │  [Contact Sales]     │
├──────────┴───────────────┴──────────────────────┤
│         Feature comparison table below           │
│         FAQ section                              │
│         Money-back guarantee badge               │
└─────────────────────────────────────────────────┘
```

### Trust Elements

Place these near the payment button:
- Money-back guarantee (30-day)
- Security badges (SSL, PCI compliant)
- Customer count ("Join 10,000+ companies")
- Social proof (logos, testimonials)
- "Cancel anytime" text

## Revenue Metrics to Track

### Must-Have Metrics

| Metric | Formula | Target |
|--------|---------|--------|
| **MRR** | Sum of all monthly subscription revenue | Growing month-over-month |
| **ARR** | MRR x 12 | For annual planning |
| **Churn Rate** | Canceled customers / Total customers (monthly) | < 5% monthly |
| **ARPU** | Total revenue / Number of customers | Growing over time |
| **LTV** | ARPU / Monthly churn rate | > 3x CAC |
| **CAC** | Total acquisition cost / New customers | < 1/3 LTV |
| **LTV:CAC** | LTV / CAC | > 3:1 |
| **Net Revenue Retention** | (Starting MRR + Expansion - Contraction - Churn) / Starting MRR | > 100% |

### Revenue Levers

```
Revenue = Customers x ARPU

To grow revenue:
1. More customers (acquisition)
2. Higher ARPU (upsells, plan upgrades, price increases)
3. Lower churn (retention, product quality)
4. Expansion revenue (usage growth, seat growth)
```

## Stripe Implementation for Pricing Pages

### Dynamic Pricing Page

```typescript
// Fetch prices from Stripe (cache this!)
app.get('/api/pricing', async (req, res) => {
  const prices = await stripe.prices.list({
    active: true,
    expand: ['data.product'],
    type: 'recurring',
  });

  const plans = prices.data
    .filter(p => (p.product as Stripe.Product).active)
    .map(p => ({
      id: p.id,
      name: (p.product as Stripe.Product).name,
      description: (p.product as Stripe.Product).description,
      price: p.unit_amount,
      interval: p.recurring?.interval,
      features: (p.product as Stripe.Product).metadata.features?.split(',') || [],
    }))
    .sort((a, b) => (a.price || 0) - (b.price || 0));

  res.json(plans);
});
```

### Coupon / Promotion Codes

```typescript
// Create a coupon
const coupon = await stripe.coupons.create({
  percent_off: 20,
  duration: 'once',  // 'once' | 'repeating' | 'forever'
  max_redemptions: 100,
  redeem_by: Math.floor(Date.now() / 1000) + 30 * 24 * 60 * 60, // 30 days
});

// Create a shareable promotion code
const promoCode = await stripe.promotionCodes.create({
  coupon: coupon.id,
  code: 'LAUNCH20',
  max_redemptions: 50,
});

// Apply in Checkout Session
const session = await stripe.checkout.sessions.create({
  // ...
  allow_promotion_codes: true,  // Let users enter codes
  // OR apply automatically:
  discounts: [{ promotion_code: promoCode.id }],
});
```

## Common Pricing Mistakes

1. **Pricing too low** — Most indie SaaS products are underpriced. If nobody complains about price, you're too cheap.
2. **Too many tiers** — 3-4 max. Beyond that, decision paralysis kills conversion.
3. **Feature-gating wrong things** — Gate by outcome/value, not by feature count.
4. **Not offering annual billing** — You lose out on cash flow and lock-in.
5. **No free trial** — "Contact sales" without a trial kills self-serve conversion.
6. **Same price for years** — Revisit pricing every 6-12 months as you add value.
7. **Hiding pricing** — "Contact us for pricing" kills trust for SMB/consumer products.
8. **Not testing** — A/B test pricing pages. Even small changes (layout, CTA text, plan naming) can move the needle.
