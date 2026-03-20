# SaaS Pricing Strategy Guide

Practical pricing decisions for SaaS products, informed by real-world patterns and Stripe implementation details.

---

## Pricing Models Ranked by Simplicity

### 1. Flat-Rate Subscription

```
$29/mo — one plan, all features
```

**When to use**: Early stage, < 100 customers, still learning what users value.
**Stripe implementation**: Single Price object with `recurring: { interval: 'month' }`.
**Pros**: Zero pricing complexity, easy to explain, no billing surprises.
**Cons**: Leaves money on the table with power users, may be too expensive for small users.
**Examples**: Basecamp, Hey.com

### 2. Good / Better / Best (Tiered Plans)

```
Free    — 3 projects, basic features
Pro     — Unlimited projects, advanced features    $19/mo
Team    — Everything + collaboration               $49/mo/seat
```

**When to use**: When you have clear feature differentiation between user segments.
**Stripe implementation**: Multiple Price objects, one per plan. Feature gating in your app.
**Pros**: Captures different willingness to pay, natural upgrade path.
**Cons**: Choosing the right feature split is hard. Customers hate feeling feature-gated.
**Examples**: Notion, Linear, Figma

### 3. Per-Seat

```
$12/user/month — all features
```

**When to use**: Collaborative tools where more users = more value.
**Stripe implementation**: Price with `recurring`, then set `quantity` on subscription items.
**Pros**: Revenue scales with customer growth. Natural expansion revenue.
**Cons**: Encourages seat-sharing. Small teams feel overcharged.
**Examples**: Slack, GitHub, Jira

### 4. Usage-Based (Metered)

```
$0.01 per API call
$0.10 per GB stored
```

**When to use**: Infrastructure, APIs, storage — where usage varies 100x between customers.
**Stripe implementation**: Price with `usage_type: 'metered'`, report usage via `subscriptionItems.createUsageRecord()`.
**Pros**: Low barrier to start, revenue tracks actual value delivered.
**Cons**: Unpredictable bills frustrate customers. Hard to forecast revenue.
**Examples**: AWS, Twilio, OpenAI

### 5. Hybrid (Base + Usage)

```
$49/mo base — includes 10K API calls
$0.005 per additional call
```

**When to use**: When you want predictable base revenue plus upside from heavy users.
**Stripe implementation**: One fixed Price + one metered Price on the same subscription.
**Pros**: Predictable base revenue, captures upside, customers get a "floor" of value.
**Cons**: More complex billing, harder to communicate.
**Examples**: Vercel, Supabase, PlanetScale

### 6. Freemium

```
Free — limited usage (forever)
Pro  — full access              $X/mo
```

**When to use**: Products with viral/network effects, or where free users generate value (content, data, word-of-mouth).
**Stripe implementation**: No Stripe for free tier. Only create subscriptions on upgrade.
**Pros**: Maximum top-of-funnel, low barrier.
**Cons**: Most users never pay (2-5% conversion typical). Free users cost money to support.
**Examples**: Spotify, Dropbox, Notion

## Pricing Page Best Practices

### The Three-Plan Rule

Most pricing pages should have exactly 3 plans:
1. **Starter** — for individuals and small teams (anchor the low end)
2. **Pro** — the plan you want most people on (highlight this one)
3. **Enterprise** — for large orgs (anchor the high end, "Contact Sales")

### Annual Discount

Offer 15-20% off for annual billing. This:
- Reduces churn (committed for a year)
- Improves cash flow
- Standard discount: 2 months free on annual = 16.7% off

```
Monthly: $29/mo
Annual:  $290/yr ($24.17/mo — save $58)
```

### The Decoy Effect

Add a plan that makes the target plan look better:

```
Basic: $9/mo   — 5 projects
Pro:   $29/mo  — Unlimited projects, API access    ← Target
Team:  $39/mo  — Everything in Pro + 5 seats        ← Decoy (only $10 more for seats)
```

### Pricing Anchoring

Show the most expensive plan first (left-to-right or top-to-bottom). This makes the middle plan feel reasonable by comparison.

## Free Trial Strategy

| Strategy | Conversion Rate | Best For |
|----------|----------------|----------|
| No trial, freemium | 2-5% of free → paid | Products with viral loops |
| 7-day trial, no card | 8-12% trial → paid | Simple products, quick value |
| 14-day trial, no card | 10-15% trial → paid | Most SaaS (default choice) |
| 14-day trial, card required | 40-60% trial → paid | Higher intent, lower volume |
| 30-day trial | 5-10% trial → paid | Complex enterprise products |
| Reverse trial (start on Pro) | 15-25% trial → paid | Products where Pro features sell themselves |

**Default recommendation**: 14-day trial, no card required. It maximizes trial starts while giving enough time for the "aha moment."

### Trial Implementation

```typescript
// No card required — just track trial end
const subscription = await stripe.subscriptions.create({
  customer: customerId,
  items: [{ price: priceId }],
  trial_period_days: 14,
  trial_settings: {
    end_behavior: { missing_payment_method: 'cancel' }, // Don't charge if no card added
  },
});

// Card required — will auto-charge at trial end
const subscription = await stripe.subscriptions.create({
  customer: customerId,
  items: [{ price: priceId }],
  trial_period_days: 14,
  payment_behavior: 'default_incomplete',
  expand: ['latest_invoice.payment_intent'],
});
```

## Revenue Recovery (Dunning)

When a payment fails on a subscription renewal:

### Stripe Smart Retries

Enable in Dashboard → Settings → Billing → Subscriptions → Smart Retries. Stripe uses ML to pick the optimal retry time. Recovers ~15% of failed payments automatically.

### Manual Dunning Flow

| Day | Action |
|-----|--------|
| 0 | Payment fails → send soft email ("payment issue, please update") |
| 3 | Stripe auto-retry #1 |
| 5 | Send email #2 ("your account may be affected") |
| 7 | Stripe auto-retry #2 |
| 10 | Send email #3 with urgency ("last chance before cancellation") |
| 14 | Stripe auto-retry #3 (final) |
| 14 | If still failing → cancel subscription, send "we're sorry to see you go" |

### In-App Banner

```tsx
function PaymentBanner({ subscriptionStatus }: { subscriptionStatus: string }) {
  if (subscriptionStatus !== 'past_due') return null;

  return (
    <div className="bg-yellow-50 border-l-4 border-yellow-400 p-4">
      <p className="font-medium">Payment issue</p>
      <p className="text-sm">
        Your last payment failed. Please update your payment method to avoid
        service interruption.
      </p>
      <a href="/billing" className="text-yellow-700 font-medium">
        Update payment method →
      </a>
    </div>
  );
}
```

## Metrics to Track

| Metric | Formula | Healthy Range |
|--------|---------|---------------|
| **MRR** | Sum of all monthly subscription revenue | Growing |
| **ARPU** | MRR / active customers | $20-200 for SMB SaaS |
| **Churn rate** | Cancelled / total customers per month | < 5% monthly |
| **Net revenue retention** | (MRR + expansion - contraction - churn) / MRR | > 100% (expansion > churn) |
| **Trial conversion** | Paid / total trials | 10-25% (no card) |
| **LTV** | ARPU / monthly churn rate | > 3x CAC |
| **Expansion revenue %** | Revenue from upgrades / total new MRR | > 30% at scale |

## Common Pricing Mistakes

1. **Pricing too low** — Most indie SaaS underprices by 2-3x. If nobody complains about price, you're too cheap. Aim for 20% of prospects saying "too expensive."

2. **Too many plans** — 3-4 plans max. Every additional plan increases decision paralysis and support burden.

3. **Feature gating the wrong things** — Gate based on scale (usage limits, seats) not based on core features. Users hate paying to unlock features they can see but can't use.

4. **Not offering annual** — Annual plans reduce churn by 30-50% and improve cash flow. Always offer with a meaningful discount.

5. **Free tier too generous** — If free users never hit limits, they never convert. Set limits that let users experience value but create friction at the point of serious use.

6. **Changing prices without grandfathering** — Always grandfather existing customers on their current plan for 6-12 months. Use Stripe's `billing_cycle_anchor` and `proration_behavior` to manage transitions.

7. **No cancellation flow** — A simple "why are you leaving?" survey + a one-click pause option can save 10-15% of cancellations.

8. **Ignoring expansion revenue** — The cheapest revenue is from existing customers upgrading. Build natural upgrade triggers (usage limits, seat additions, feature unlock prompts).
