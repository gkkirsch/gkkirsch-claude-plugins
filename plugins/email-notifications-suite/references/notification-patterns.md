# Notification System Design Patterns

## Notification Architecture

```
┌──────────────┐     ┌──────────────┐     ┌──────────────────┐
│  App Event   │────→│  Dispatcher  │────→│  Channel Router  │
│ (order.paid) │     │  (decides    │     │                  │
└──────────────┘     │   who/what)  │     │  ├─ Email        │
                     └──────────────┘     │  ├─ Push         │
                                          │  ├─ In-App       │
                                          │  ├─ SMS          │
                                          │  └─ Slack/Webhook │
                                          └──────────────────┘
```

## Notification Preferences Schema

```prisma
model NotificationPreference {
  id        String  @id @default(cuid())
  userId    String
  channel   String  // 'email', 'push', 'in_app', 'sms'
  category  String  // 'marketing', 'order_updates', 'messages', 'security'
  enabled   Boolean @default(true)

  @@unique([userId, channel, category])
  @@index([userId])
}
```

### Preference Categories

| Category | Default Channels | Can Disable? |
|----------|-----------------|--------------|
| `security` | Email + Push + In-App | No (always on) |
| `order_updates` | Email + In-App | Email: No, Push: Yes |
| `messages` | Email + Push + In-App | Yes |
| `marketing` | Email | Yes |
| `product_updates` | Email + In-App | Yes |
| `billing` | Email + In-App | Email: No |

### Implementation

```typescript
async function getNotificationChannels(
  userId: string,
  category: string
): Promise<string[]> {
  const prefs = await db.notificationPreference.findMany({
    where: { userId, category },
  });

  // Security notifications always go everywhere
  if (category === 'security') {
    return ['email', 'push', 'in_app'];
  }

  return prefs.filter(p => p.enabled).map(p => p.channel);
}

async function notify(userId: string, notification: {
  category: string;
  title: string;
  body: string;
  data?: Record<string, any>;
  actionUrl?: string;
}) {
  const channels = await getNotificationChannels(userId, notification.category);

  const promises = channels.map(channel => {
    switch (channel) {
      case 'email':
        return sendEmail(userId, notification);
      case 'push':
        return sendPushNotification(userId, notification);
      case 'in_app':
        return createInAppNotification(userId, notification);
      case 'sms':
        return sendSMS(userId, notification);
      default:
        return Promise.resolve();
    }
  });

  await Promise.allSettled(promises);
}
```

## Event-Driven Notifications

```typescript
import { EventEmitter } from 'events';

class NotificationEvents extends EventEmitter {}
const events = new NotificationEvents();

// Register handlers
events.on('order.confirmed', async ({ orderId, userId }) => {
  const order = await db.order.findUnique({ where: { id: orderId } });
  await notify(userId, {
    category: 'order_updates',
    title: 'Order Confirmed',
    body: `Order #${orderId} has been confirmed`,
    actionUrl: `/orders/${orderId}`,
    data: { orderId },
  });
});

events.on('message.received', async ({ messageId, recipientId, senderName }) => {
  await notify(recipientId, {
    category: 'messages',
    title: `New message from ${senderName}`,
    body: 'You have a new message',
    actionUrl: `/messages/${messageId}`,
    data: { messageId },
  });
});

events.on('payment.failed', async ({ userId, amount }) => {
  await notify(userId, {
    category: 'billing',
    title: 'Payment Failed',
    body: `Your payment of $${(amount / 100).toFixed(2)} was declined`,
    actionUrl: '/settings/billing',
    data: { amount },
  });
});

// Emit from anywhere in your app
events.emit('order.confirmed', { orderId: 'ord_123', userId: 'usr_456' });
```

## Notification Batching

Don't send 10 notifications in 1 minute. Batch them.

```typescript
class NotificationBatcher {
  private pending = new Map<string, any[]>(); // userId -> notifications
  private timers = new Map<string, NodeJS.Timeout>();

  add(userId: string, notification: any) {
    const existing = this.pending.get(userId) || [];
    existing.push(notification);
    this.pending.set(userId, existing);

    // Reset or start batch timer (5 minutes)
    const existingTimer = this.timers.get(userId);
    if (existingTimer) clearTimeout(existingTimer);

    this.timers.set(userId, setTimeout(() => {
      this.flush(userId);
    }, 5 * 60 * 1000));

    // Immediate flush if batch is large
    if (existing.length >= 10) {
      this.flush(userId);
    }
  }

  private async flush(userId: string) {
    const notifications = this.pending.get(userId) || [];
    this.pending.delete(userId);
    this.timers.delete(userId);

    if (notifications.length === 0) return;

    if (notifications.length === 1) {
      await notify(userId, notifications[0]);
    } else {
      // Send batched digest
      await notify(userId, {
        category: notifications[0].category,
        title: `${notifications.length} new notifications`,
        body: notifications.map(n => n.title).join('\n'),
        actionUrl: '/notifications',
      });
    }
  }
}
```

## Rate Limiting Notifications

```typescript
import { RateLimiterRedis } from 'rate-limiter-flexible';

const notificationLimiter = new RateLimiterRedis({
  storeClient: redisClient,
  keyPrefix: 'notif_limit',
  points: 20,           // Max 20 notifications
  duration: 3600,        // Per hour
});

async function rateLimitedNotify(userId: string, notification: any) {
  try {
    await notificationLimiter.consume(userId);
    await notify(userId, notification);
  } catch (rejRes) {
    // Rate limit exceeded — queue for later or skip
    console.warn(`Notification rate limit hit for user ${userId}`);
    // Optionally queue for batch digest
    batcher.add(userId, notification);
  }
}
```

## Do Not Disturb (DND)

```typescript
interface DNDSettings {
  enabled: boolean;
  startHour: number;  // 0-23
  endHour: number;    // 0-23
  timezone: string;
  allowUrgent: boolean;
}

function isInDNDPeriod(settings: DNDSettings): boolean {
  if (!settings.enabled) return false;

  const now = new Date();
  const userTime = new Date(now.toLocaleString('en-US', { timeZone: settings.timezone }));
  const currentHour = userTime.getHours();

  if (settings.startHour <= settings.endHour) {
    return currentHour >= settings.startHour && currentHour < settings.endHour;
  } else {
    // Overnight DND (e.g., 22:00 - 08:00)
    return currentHour >= settings.startHour || currentHour < settings.endHour;
  }
}
```

## Notification Templates

### Template System

```typescript
const templates: Record<string, {
  title: (data: any) => string;
  body: (data: any) => string;
  email?: (data: any) => { subject: string; html: string };
}> = {
  'order.confirmed': {
    title: (d) => 'Order Confirmed',
    body: (d) => `Your order #${d.orderNumber} has been confirmed and is being prepared.`,
    email: (d) => ({
      subject: `Order #${d.orderNumber} Confirmed`,
      html: render(OrderConfirmedEmail(d)),
    }),
  },
  'payment.received': {
    title: (d) => 'Payment Received',
    body: (d) => `We received your payment of $${(d.amount / 100).toFixed(2)}.`,
  },
  'shipping.shipped': {
    title: (d) => 'Order Shipped!',
    body: (d) => `Your order #${d.orderNumber} is on its way. Tracking: ${d.trackingNumber}`,
  },
};

function resolveTemplate(type: string, data: any) {
  const template = templates[type];
  if (!template) throw new Error(`Unknown notification template: ${type}`);
  return {
    title: template.title(data),
    body: template.body(data),
    email: template.email?.(data),
  };
}
```

## Multi-Tenant Notification Branding

```typescript
interface TenantBranding {
  name: string;
  logo: string;
  primaryColor: string;
  fromEmail: string;
  fromName: string;
}

// Use tenant branding in email templates
function renderBrandedEmail(tenant: TenantBranding, content: { title: string; body: string }) {
  return `
    <div style="font-family: sans-serif; max-width: 560px; margin: 0 auto;">
      <img src="${tenant.logo}" alt="${tenant.name}" height="40" />
      <h1 style="color: ${tenant.primaryColor}">${content.title}</h1>
      <p>${content.body}</p>
      <footer style="color: #999; font-size: 12px; margin-top: 40px;">
        Sent by ${tenant.name}
      </footer>
    </div>
  `;
}
```

## Webhook Notifications (For Integrations)

```typescript
// Allow users to register webhook URLs
app.post('/api/webhooks/register', async (req, res) => {
  const { url, events, secret } = req.body;

  const webhook = await db.webhook.create({
    data: {
      userId: req.user.id,
      url,
      events,  // ['order.created', 'payment.received']
      secret: secret || crypto.randomBytes(32).toString('hex'),
      active: true,
    },
  });

  res.json({ id: webhook.id, secret: webhook.secret });
});

// Deliver webhook
async function deliverWebhook(userId: string, event: string, payload: any) {
  const webhooks = await db.webhook.findMany({
    where: {
      userId,
      active: true,
      events: { has: event },
    },
  });

  for (const webhook of webhooks) {
    const signature = crypto
      .createHmac('sha256', webhook.secret)
      .update(JSON.stringify(payload))
      .digest('hex');

    try {
      const response = await fetch(webhook.url, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-Webhook-Signature': `sha256=${signature}`,
          'X-Webhook-Event': event,
        },
        body: JSON.stringify(payload),
        signal: AbortSignal.timeout(10000),
      });

      await db.webhookDelivery.create({
        data: {
          webhookId: webhook.id,
          event,
          statusCode: response.status,
          success: response.ok,
        },
      });
    } catch (error) {
      await db.webhookDelivery.create({
        data: {
          webhookId: webhook.id,
          event,
          statusCode: 0,
          success: false,
          error: error.message,
        },
      });
    }
  }
}
```

## Checklist: Building a Notification System

- [ ] Define notification categories and default channels
- [ ] Build user preference UI (per-category, per-channel toggles)
- [ ] Implement event-driven dispatch (not inline notification calls)
- [ ] Add rate limiting per user
- [ ] Implement batching/digest for high-frequency events
- [ ] Support Do Not Disturb hours
- [ ] Handle delivery failures gracefully (retry, fallback channel)
- [ ] Clean up expired push subscriptions
- [ ] Log all notification deliveries for debugging
- [ ] Build notification history/inbox in the app
- [ ] Add "mark all as read" functionality
- [ ] Respect "security" notifications as mandatory (can't opt out)
- [ ] Include unsubscribe links in all marketing emails
