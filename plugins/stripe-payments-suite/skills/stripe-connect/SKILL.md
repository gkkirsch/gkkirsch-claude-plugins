---
name: stripe-connect
description: >
  Stripe Connect for marketplace and platform payments — onboarding sellers,
  processing payments with application fees, managing transfers and payouts,
  and handling connected account lifecycles.
  Triggers: "stripe connect", "marketplace payments", "platform payments",
  "seller onboarding", "application fees", "connected accounts", "payouts".
  NOT for: basic checkout (use stripe-checkout), subscriptions (use stripe-subscriptions).
version: 1.0.0
allowed-tools: Read, Grep, Glob, Bash, Edit, Write
---

# Stripe Connect — Marketplace Payments

## When to Use Connect

Use Stripe Connect when your platform facilitates payments between buyers and sellers:
- **Marketplaces** — Etsy-like, Airbnb-like, Uber-like
- **SaaS platforms** — letting users charge their own customers
- **Crowdfunding** — collecting and distributing funds
- **On-demand services** — splitting payments between providers

## Account Types

| Type | Control | Onboarding | Best For |
|------|---------|------------|----------|
| **Standard** | Seller manages their Stripe | Stripe-hosted (redirect) | Marketplaces where sellers are businesses |
| **Express** | Platform + seller share | Stripe-hosted (simpler) | Gig platforms, smaller sellers |
| **Custom** | Platform manages everything | You build the UI | Full white-label platforms |

**Start with Standard** unless you have a specific reason not to. It's the simplest, has the least compliance burden, and sellers manage their own tax/banking.

## Setup

```bash
npm install stripe
```

```typescript
import Stripe from 'stripe';

const stripe = new Stripe(process.env.STRIPE_SECRET_KEY!, {
  apiVersion: '2024-12-18.acacia',
});
```

## Onboarding Connected Accounts

### Standard Account Onboarding

```typescript
// Step 1: Create connected account
async function createConnectedAccount(
  email: string,
  sellerId: string
): Promise<string> {
  const account = await stripe.accounts.create({
    type: 'standard',
    email,
    metadata: {
      sellerId,
      platform: 'your-platform-name',
    },
  });

  // Save the account ID
  await db.seller.update({
    where: { id: sellerId },
    data: { stripeAccountId: account.id },
  });

  return account.id;
}

// Step 2: Create onboarding link
async function createOnboardingLink(
  accountId: string
): Promise<string> {
  const accountLink = await stripe.accountLinks.create({
    account: accountId,
    refresh_url: `${process.env.APP_URL}/seller/onboarding/refresh`,
    return_url: `${process.env.APP_URL}/seller/onboarding/complete`,
    type: 'account_onboarding',
  });

  return accountLink.url; // Redirect seller here
}
```

### Express Account Onboarding

```typescript
async function createExpressAccount(
  email: string,
  sellerId: string
): Promise<string> {
  const account = await stripe.accounts.create({
    type: 'express',
    email,
    capabilities: {
      card_payments: { requested: true },
      transfers: { requested: true },
    },
    business_type: 'individual',
    metadata: { sellerId },
  });

  await db.seller.update({
    where: { id: sellerId },
    data: { stripeAccountId: account.id },
  });

  return account.id;
}
```

### Onboarding Status Check

```typescript
async function checkOnboardingStatus(accountId: string) {
  const account = await stripe.accounts.retrieve(accountId);

  return {
    chargesEnabled: account.charges_enabled,
    payoutsEnabled: account.payouts_enabled,
    detailsSubmitted: account.details_submitted,
    requirements: account.requirements?.currently_due || [],
    errors: account.requirements?.errors || [],
  };
}
```

### Onboarding Refresh Handler

Account links expire quickly. When the seller returns to `refresh_url`, generate a new link:

```typescript
app.get('/seller/onboarding/refresh', async (req, res) => {
  const seller = await db.seller.findUnique({
    where: { userId: req.user.id },
  });

  const link = await createOnboardingLink(seller!.stripeAccountId);
  res.redirect(link);
});
```

## Payment Flows

### Destination Charges (Recommended)

Platform creates the charge, Stripe automatically transfers to the connected account minus application fee.

```typescript
async function createMarketplaceCheckout(
  buyerEmail: string,
  sellerAccountId: string,
  items: CartItem[],
  platformFeePercent: number = 10
) {
  const totalAmount = items.reduce((sum, i) => sum + i.price * i.quantity, 0);
  const applicationFee = Math.round(totalAmount * (platformFeePercent / 100));

  const session = await stripe.checkout.sessions.create({
    mode: 'payment',
    customer_email: buyerEmail,
    line_items: items.map(item => ({
      price_data: {
        currency: 'usd',
        product_data: {
          name: item.name,
          images: item.images,
        },
        unit_amount: item.price,
      },
      quantity: item.quantity,
    })),
    payment_intent_data: {
      application_fee_amount: applicationFee,
      transfer_data: {
        destination: sellerAccountId,
      },
    },
    success_url: `${process.env.APP_URL}/order/success?session_id={CHECKOUT_SESSION_ID}`,
    cancel_url: `${process.env.APP_URL}/cart`,
    metadata: {
      sellerAccountId,
      platformFee: applicationFee.toString(),
    },
  });

  return session;
}
```

### Direct Charges (On Connected Account)

Charge is created on the connected account. Platform takes a fee. Use when the seller is the merchant of record.

```typescript
async function createDirectCharge(
  sellerAccountId: string,
  amount: number,
  applicationFee: number
) {
  const paymentIntent = await stripe.paymentIntents.create(
    {
      amount,
      currency: 'usd',
      application_fee_amount: applicationFee,
    },
    {
      stripeAccount: sellerAccountId, // On behalf of connected account
    }
  );

  return paymentIntent;
}
```

### Separate Charges and Transfers

Platform charges the customer, then transfers to one or more connected accounts. Use for splitting payments across multiple sellers.

```typescript
async function chargeAndSplit(
  amount: number,
  splits: Array<{ accountId: string; amount: number }>
) {
  // Step 1: Charge the customer
  const paymentIntent = await stripe.paymentIntents.create({
    amount,
    currency: 'usd',
  });

  // Step 2: Transfer to each seller (after charge succeeds)
  for (const split of splits) {
    await stripe.transfers.create({
      amount: split.amount,
      currency: 'usd',
      destination: split.accountId,
      source_transaction: paymentIntent.latest_charge as string,
    });
  }
}
```

## Application Fees

### Fee Models

```typescript
// Percentage-based (most common)
const fee = Math.round(amount * 0.10); // 10%

// Fixed fee
const fee = 200; // $2.00 flat

// Tiered (volume-based)
function calculateFee(amount: number, monthlyVolume: number): number {
  if (monthlyVolume > 100000) return Math.round(amount * 0.05);  // 5%
  if (monthlyVolume > 10000) return Math.round(amount * 0.08);   // 8%
  return Math.round(amount * 0.10);                                // 10%
}

// Percentage + fixed
const fee = Math.round(amount * 0.10) + 50; // 10% + $0.50
```

### Fee on Subscriptions

```typescript
const subscription = await stripe.subscriptions.create(
  {
    customer: customerId,
    items: [{ price: priceId }],
    application_fee_percent: 10, // 10% of each recurring payment
  },
  {
    stripeAccount: sellerAccountId,
  }
);
```

## Managing Connected Accounts

### Dashboard Login Link (Express/Custom)

```typescript
async function getSellerDashboardLink(accountId: string) {
  const loginLink = await stripe.accounts.createLoginLink(accountId);
  return loginLink.url;
}
```

### Check Balance

```typescript
async function getSellerBalance(accountId: string) {
  const balance = await stripe.balance.retrieve({
    stripeAccount: accountId,
  });

  return {
    available: balance.available.map(b => ({
      amount: b.amount / 100,
      currency: b.currency,
    })),
    pending: balance.pending.map(b => ({
      amount: b.amount / 100,
      currency: b.currency,
    })),
  };
}
```

### List Payouts

```typescript
async function getSellerPayouts(accountId: string) {
  const payouts = await stripe.payouts.list(
    { limit: 10 },
    { stripeAccount: accountId }
  );

  return payouts.data.map(p => ({
    id: p.id,
    amount: p.amount / 100,
    currency: p.currency,
    status: p.status,
    arrivalDate: new Date(p.arrival_date * 1000),
  }));
}
```

## Connect Webhooks

### Events to Listen For

```typescript
const connectHandlers: Record<string, (event: Stripe.Event) => Promise<void>> = {
  // Account lifecycle
  'account.updated': async (event) => {
    const account = event.data.object as Stripe.Account;

    await db.seller.update({
      where: { stripeAccountId: account.id },
      data: {
        chargesEnabled: account.charges_enabled,
        payoutsEnabled: account.payouts_enabled,
        onboardingComplete: account.details_submitted,
      },
    });

    // Notify seller if requirements need attention
    if (account.requirements?.currently_due?.length) {
      await notifySeller(account.id, 'requirements_due', {
        requirements: account.requirements.currently_due,
      });
    }
  },

  // Payout events
  'payout.paid': async (event) => {
    const payout = event.data.object as Stripe.Payout;
    await db.payout.update({
      where: { stripePayoutId: payout.id },
      data: { status: 'paid', paidAt: new Date() },
    });
  },

  'payout.failed': async (event) => {
    const payout = event.data.object as Stripe.Payout;
    await notifySeller(event.account!, 'payout_failed', {
      amount: payout.amount / 100,
      failureCode: payout.failure_code,
      failureMessage: payout.failure_message,
    });
  },

  // Application fee collected
  'application_fee.created': async (event) => {
    const fee = event.data.object as Stripe.ApplicationFee;
    await db.platformRevenue.create({
      data: {
        stripeFeeId: fee.id,
        amount: fee.amount,
        connectedAccountId: fee.account as string,
        chargeId: fee.charge as string,
      },
    });
  },
};
```

### Receiving Connect Webhooks

Connect events include `event.account` — the connected account ID. Register a separate endpoint or check for the `account` field:

```typescript
app.post('/api/webhooks/stripe/connect', express.raw({ type: 'application/json' }), async (req, res) => {
  const sig = req.headers['stripe-signature']!;
  let event: Stripe.Event;

  try {
    event = stripe.webhooks.constructEvent(
      req.body,
      sig,
      process.env.STRIPE_CONNECT_WEBHOOK_SECRET! // Different secret
    );
  } catch (err: any) {
    return res.status(400).send(`Webhook Error: ${err.message}`);
  }

  res.json({ received: true });

  const handler = connectHandlers[event.type];
  if (handler) await handler(event);
});
```

## Refunds on Connect

### Refund a Destination Charge

```typescript
async function refundMarketplacePurchase(
  paymentIntentId: string,
  refundApplicationFee: boolean = true
) {
  const refund = await stripe.refunds.create({
    payment_intent: paymentIntentId,
    refund_application_fee: refundApplicationFee, // Also refund platform's fee?
    reverse_transfer: true, // Reverse the transfer to seller
  });

  return refund;
}
```

### Partial Refund

```typescript
const refund = await stripe.refunds.create({
  payment_intent: paymentIntentId,
  amount: 500, // Refund $5.00 of the charge
  refund_application_fee: true,
  reverse_transfer: true,
});
```

## Seller Onboarding UI

### React Onboarding Component

```tsx
function SellerOnboarding() {
  const [status, setStatus] = useState<'idle' | 'loading' | 'error'>('idle');

  async function startOnboarding() {
    setStatus('loading');
    try {
      const res = await fetch('/api/seller/onboard', { method: 'POST' });
      const { url } = await res.json();
      window.location.href = url; // Redirect to Stripe
    } catch {
      setStatus('error');
    }
  }

  return (
    <div>
      <h2>Start Selling</h2>
      <p>Connect your bank account to receive payments</p>
      <button onClick={startOnboarding} disabled={status === 'loading'}>
        {status === 'loading' ? 'Setting up...' : 'Connect with Stripe'}
      </button>
      {status === 'error' && (
        <p className="text-red-500">Something went wrong. Please try again.</p>
      )}
    </div>
  );
}
```

### Seller Dashboard Component

```tsx
function SellerDashboard({ seller }: { seller: Seller }) {
  if (!seller.chargesEnabled) {
    return (
      <div className="bg-yellow-50 p-4 rounded">
        <h3>Complete Your Setup</h3>
        <p>Finish onboarding to start receiving payments</p>
        <a href={`/api/seller/onboard/refresh`}>Complete Setup →</a>
      </div>
    );
  }

  return (
    <div>
      <h2>Seller Dashboard</h2>
      <div className="grid grid-cols-3 gap-4">
        <StatCard label="Available" value={`$${seller.balanceAvailable}`} />
        <StatCard label="Pending" value={`$${seller.balancePending}`} />
        <StatCard label="Total Earned" value={`$${seller.totalEarned}`} />
      </div>
      <a href={seller.dashboardUrl} target="_blank">
        Open Stripe Dashboard →
      </a>
    </div>
  );
}
```

## API Routes

```typescript
// POST /api/seller/onboard — start onboarding
app.post('/api/seller/onboard', requireAuth, async (req, res) => {
  let seller = await db.seller.findUnique({
    where: { userId: req.user.id },
  });

  if (!seller?.stripeAccountId) {
    const accountId = await createConnectedAccount(req.user.email, req.user.id);
    seller = await db.seller.findUnique({ where: { userId: req.user.id } });
  }

  const link = await createOnboardingLink(seller!.stripeAccountId);
  res.json({ url: link });
});

// GET /api/seller/status — check onboarding status
app.get('/api/seller/status', requireAuth, async (req, res) => {
  const seller = await db.seller.findUnique({
    where: { userId: req.user.id },
  });

  if (!seller?.stripeAccountId) {
    return res.json({ status: 'not_started' });
  }

  const status = await checkOnboardingStatus(seller.stripeAccountId);
  res.json(status);
});

// GET /api/seller/dashboard — get dashboard login link
app.get('/api/seller/dashboard', requireAuth, async (req, res) => {
  const seller = await db.seller.findUnique({
    where: { userId: req.user.id },
  });

  const link = await getSellerDashboardLink(seller!.stripeAccountId);
  res.json({ url: link });
});

// GET /api/seller/balance — get current balance
app.get('/api/seller/balance', requireAuth, async (req, res) => {
  const seller = await db.seller.findUnique({
    where: { userId: req.user.id },
  });

  const balance = await getSellerBalance(seller!.stripeAccountId);
  res.json(balance);
});
```

## Common Gotchas

1. **Account links expire fast** — Links from `accountLinks.create()` expire in minutes. Always generate fresh on redirect.

2. **Standard accounts need their own Stripe login** — You can't manage their dashboard. Sellers log into Stripe directly. For more control, use Express.

3. **Transfer availability** — Transfers to connected accounts can take 2-7 business days to become available for payout, depending on the connected account's country and payout schedule.

4. **Negative balances** — If you refund a destination charge, the connected account's balance can go negative. Stripe recovers from future payments. Plan for this in your UI.

5. **Tax reporting** — Stripe issues 1099s automatically for US connected accounts earning > $600/year. You don't need to handle this manually for Standard/Express accounts.

6. **Cross-border payments** — Currency conversion happens automatically. The platform fee is always in the platform's default currency. Be aware of exchange rate differences.

7. **Multiple sellers per order** — Use separate charges and transfers (not destination charges) when a single order pays multiple sellers.

8. **Testing Connect** — Use `stripe trigger account.updated` and test account tokens. Connect webhook events need a separate webhook endpoint with its own signing secret.
