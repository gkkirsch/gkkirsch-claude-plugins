# Notification Patterns Reference

## Notification Type Catalog

### Authentication & Security

| Notification | Channels | Priority | Template Variables |
|-------------|----------|----------|-------------------|
| Welcome email | Email | Normal | name, loginUrl |
| Email verification | Email | High | name, verifyUrl |
| Password reset | Email | Urgent | name, resetUrl, expiresIn |
| 2FA code | SMS, Email | Urgent | code, expiresIn |
| New device login | Email, Push | High | device, location, time |
| Password changed | Email | High | name, time |
| Account locked | Email, SMS | Urgent | name, unlockUrl |
| Session expired | In-app | Low | — |

### Billing & Subscription

| Notification | Channels | Priority | Template Variables |
|-------------|----------|----------|-------------------|
| Trial ending | Email, Push | High | name, daysLeft, planName, billingUrl |
| Payment received | Email | Normal | amount, invoiceUrl |
| Payment failed | Email, Push, In-app | Urgent | name, nextRetry, updatePaymentUrl |
| Subscription renewed | Email | Normal | planName, nextBillingDate, amount |
| Subscription canceled | Email | Normal | name, accessEnd, reactivateUrl |
| Plan upgraded | Email, In-app | Normal | oldPlan, newPlan, effectiveDate |
| Plan downgraded | Email, In-app | Normal | oldPlan, newPlan, effectiveDate |
| Invoice ready | Email | Normal | invoiceNumber, amount, downloadUrl |
| Refund processed | Email | Normal | amount, reason |

### E-Commerce

| Notification | Channels | Priority | Template Variables |
|-------------|----------|----------|-------------------|
| Order confirmed | Email, SMS, In-app | High | orderId, items, total, orderUrl |
| Order shipped | Email, SMS, Push | High | orderId, trackingUrl, carrier, eta |
| Out for delivery | Push, SMS | Normal | orderId, eta |
| Delivered | Email, Push | Normal | orderId, reviewUrl |
| Order canceled | Email | Normal | orderId, reason, refundAmount |
| Refund issued | Email | Normal | orderId, amount, method |
| Back in stock | Email, Push | Normal | productName, productUrl |
| Price drop | Email, Push | Low | productName, oldPrice, newPrice |
| Cart abandoned | Email | Low | items, cartUrl |
| Review request | Email | Low | productName, reviewUrl |

### Social & Collaboration

| Notification | Channels | Priority | Template Variables |
|-------------|----------|----------|-------------------|
| New comment | In-app, Push | Normal | authorName, content, postUrl |
| Reply to comment | In-app, Push, Email | Normal | authorName, content, threadUrl |
| Mention | In-app, Push, Email | High | authorName, content, url |
| New follower | In-app | Low | followerName, profileUrl |
| Team invitation | Email, In-app | High | teamName, inviterName, acceptUrl |
| Shared document | Email, In-app | Normal | sharedBy, docName, docUrl |
| Task assigned | In-app, Push, Email | Normal | taskName, assignedBy, taskUrl |
| Task completed | In-app | Low | taskName, completedBy |
| Project update | In-app, Email | Normal | projectName, updateType, url |

### System & Admin

| Notification | Channels | Priority | Template Variables |
|-------------|----------|----------|-------------------|
| Maintenance scheduled | Email, In-app | Normal | startTime, duration, affectedServices |
| Service degradation | In-app, Push | High | service, status, statusPageUrl |
| Service restored | In-app, Push | Normal | service, downtime |
| Usage limit warning | Email, In-app | High | resource, currentUsage, limit |
| Usage limit exceeded | Email, In-app, Push | Urgent | resource, overage |
| New feature | In-app | Low | featureName, learnMoreUrl |
| Terms updated | Email | Normal | effectiveDate, changesUrl |
| Data export ready | Email, In-app | Normal | downloadUrl, expiresIn |

## Channel Selection Matrix

| Factor | In-App | Email | Push | SMS |
|--------|--------|-------|------|-----|
| **Urgency** | Low-Med | Med-High | High | Urgent |
| **User must see** | No | No | Likely | Yes |
| **Rich content** | Yes | Yes | No (short) | No (160 char) |
| **Cost** | Free | ~$0.001 | Free | ~$0.01 |
| **Delivery guarantee** | If online | High | Medium | High |
| **Opt-out rate** | Low | Medium | High | Very high |
| **Best for** | Activity feed | Receipts, digests | Real-time alerts | 2FA, critical |

## Notification Grouping & Batching

### Strategy: Digest Notifications

Instead of sending individual notifications for every event, batch related events:

```typescript
// Instead of 10 separate "new comment" emails:
// Send one digest: "You have 10 new comments"

interface DigestConfig {
  type: string;
  window: number;      // milliseconds to batch
  maxBatchSize: number;
  template: string;
}

const DIGEST_CONFIG: Record<string, DigestConfig> = {
  'comment.new': {
    type: 'comment',
    window: 15 * 60 * 1000, // 15 minutes
    maxBatchSize: 20,
    template: 'comment-digest',
  },
  'like': {
    type: 'social',
    window: 60 * 60 * 1000, // 1 hour
    maxBatchSize: 50,
    template: 'like-digest',
  },
};
```

### Strategy: Smart Grouping by Tag

```typescript
// Service worker groups notifications by tag
// Only the latest notification per tag is shown

await sendPushToUser(userId, {
  title: 'New message from Alice',
  body: 'Hey, are you free tomorrow?',
  tag: `chat:${conversationId}`,  // Groups by conversation
  renotify: true,                  // Still vibrate for updates
});
```

## Quiet Hours Implementation

```typescript
interface QuietHoursConfig {
  enabled: boolean;
  start: string;  // "22:00" (local time)
  end: string;    // "08:00" (local time)
  timezone: string; // "America/New_York"
  allowUrgent: boolean;
}

function isQuietHours(config: QuietHoursConfig): boolean {
  if (!config.enabled) return false;

  const now = new Date();
  const formatter = new Intl.DateTimeFormat('en-US', {
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
    timeZone: config.timezone,
  });

  const currentTime = formatter.format(now); // "14:30"
  const [startH, startM] = config.start.split(':').map(Number);
  const [endH, endM] = config.end.split(':').map(Number);
  const [nowH, nowM] = currentTime.split(':').map(Number);

  const nowMinutes = nowH * 60 + nowM;
  const startMinutes = startH * 60 + startM;
  const endMinutes = endH * 60 + endM;

  // Handle overnight quiet hours (e.g., 22:00 - 08:00)
  if (startMinutes > endMinutes) {
    return nowMinutes >= startMinutes || nowMinutes < endMinutes;
  }

  return nowMinutes >= startMinutes && nowMinutes < endMinutes;
}
```

## Notification Preferences UI Pattern

```
Channel Settings
├── Email notifications .............. [On/Off]
│   ├── Security alerts .............. Always on (cannot disable)
│   ├── Billing & receipts ........... [On] [Off]
│   ├── Product updates .............. [On] [Off]
│   ├── Comments & mentions .......... [On] [Off]
│   └── Marketing .................... [On] [Off]
├── Push notifications ............... [On/Off]
│   ├── Security alerts .............. [On] [Off]
│   ├── Messages ..................... [On] [Off]
│   ├── Comments & mentions .......... [On] [Off]
│   └── Task updates ................ [On] [Off]
├── SMS notifications ................ [On/Off]
│   ├── 2FA codes .................... Always on (cannot disable)
│   └── Order updates ................ [On] [Off]
└── Quiet hours
    ├── Enable ....................... [On/Off]
    ├── Start time ................... [22:00]
    ├── End time ..................... [08:00]
    └── Allow urgent ................. [Yes/No]
```

## Anti-Patterns to Avoid

| Anti-Pattern | Why It's Bad | Do This Instead |
|-------------|-------------|----------------|
| Notifying on every action | Notification fatigue | Batch related events |
| No unsubscribe option | Illegal (CAN-SPAM, GDPR) | Always include unsubscribe |
| Same message all channels | Redundant, annoying | Tailor per channel |
| Push on first visit | Low permission rate | Ask after value demonstrated |
| No quiet hours | Midnight notifications | Respect user time zones |
| Infinite notification history | DB bloat | Prune after 90 days |
| No rate limiting | Spam during incidents | Max 5/min per user |
| Ignoring bounces | Reputation damage | Remove on hard bounce |
| Generic notification text | Low engagement | Personalize with context |
| Missing fallback channel | Missed notifications | Escalate: in-app → email → SMS |
