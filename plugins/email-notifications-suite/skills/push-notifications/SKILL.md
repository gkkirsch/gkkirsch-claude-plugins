---
name: push-notifications
description: >
  Implement push notifications — Web Push API with service workers, Firebase
  Cloud Messaging (FCM), in-app notification center, and SMS via Twilio.
  Covers subscription management, permission UX, and multi-channel delivery.
  Triggers: "push notifications", "web push", "service worker notifications",
  "firebase notifications", "in-app notifications", "notification center",
  "twilio sms", "send sms".
  NOT for: email sending (use transactional-email) or email templates.
version: 1.0.0
argument-hint: "[web-push|fcm|in-app|sms]"
allowed-tools: Read, Grep, Glob, Write, Edit, Bash
---

# Push Notifications

Implement push notifications across web, mobile, and SMS channels.

## Web Push API

### Setup

```bash
npm install web-push
npx web-push generate-vapid-keys  # Generate VAPID keys
```

```bash
# .env
VAPID_PUBLIC_KEY=BLxx...
VAPID_PRIVATE_KEY=yy...
VAPID_SUBJECT=mailto:admin@yourdomain.com
```

### Server Setup

```typescript
// lib/push/web-push.ts
import webpush from 'web-push';

webpush.setVapidDetails(
  process.env.VAPID_SUBJECT!,
  process.env.VAPID_PUBLIC_KEY!,
  process.env.VAPID_PRIVATE_KEY!
);

export { webpush };
```

### Database Schema

```prisma
model PushSubscription {
  id        String   @id @default(cuid())
  userId    String
  user      User     @relation(fields: [userId], references: [id], onDelete: Cascade)
  endpoint  String   @unique
  p256dh    String   // encryption key
  auth      String   // auth secret
  userAgent String?
  createdAt DateTime @default(now())

  @@index([userId])
}
```

### Subscribe Endpoint

```typescript
// POST /api/push/subscribe
app.post('/api/push/subscribe', async (req, res) => {
  const { subscription } = req.body;
  const userId = req.user.id;

  await prisma.pushSubscription.upsert({
    where: { endpoint: subscription.endpoint },
    create: {
      userId,
      endpoint: subscription.endpoint,
      p256dh: subscription.keys.p256dh,
      auth: subscription.keys.auth,
      userAgent: req.headers['user-agent'],
    },
    update: {
      userId,
      p256dh: subscription.keys.p256dh,
      auth: subscription.keys.auth,
    },
  });

  res.json({ success: true });
});

// DELETE /api/push/unsubscribe
app.delete('/api/push/unsubscribe', async (req, res) => {
  const { endpoint } = req.body;

  await prisma.pushSubscription.deleteMany({
    where: { endpoint },
  });

  res.json({ success: true });
});
```

### Send Push Notification

```typescript
// lib/push/send.ts
import { webpush } from './web-push';

interface PushPayload {
  title: string;
  body: string;
  icon?: string;
  badge?: string;
  url?: string;
  tag?: string;       // group notifications
  renotify?: boolean; // vibrate even if same tag
  data?: Record<string, any>;
}

export async function sendPushToUser(userId: string, payload: PushPayload) {
  const subscriptions = await prisma.pushSubscription.findMany({
    where: { userId },
  });

  const results = await Promise.allSettled(
    subscriptions.map(async (sub) => {
      try {
        await webpush.sendNotification(
          {
            endpoint: sub.endpoint,
            keys: { p256dh: sub.p256dh, auth: sub.auth },
          },
          JSON.stringify(payload),
          { TTL: 86400 } // 24 hours
        );
      } catch (error: any) {
        // Remove expired subscriptions
        if (error.statusCode === 404 || error.statusCode === 410) {
          await prisma.pushSubscription.delete({ where: { id: sub.id } });
          console.log(`Removed expired subscription: ${sub.id}`);
        }
        throw error;
      }
    })
  );

  const sent = results.filter(r => r.status === 'fulfilled').length;
  const failed = results.filter(r => r.status === 'rejected').length;

  return { sent, failed, total: subscriptions.length };
}

// Send to multiple users
export async function sendPushBroadcast(
  userIds: string[],
  payload: PushPayload
) {
  return Promise.allSettled(
    userIds.map(id => sendPushToUser(id, payload))
  );
}
```

### Service Worker

```javascript
// public/sw.js
self.addEventListener('push', (event) => {
  if (!event.data) return;

  const payload = event.data.json();

  const options = {
    body: payload.body,
    icon: payload.icon || '/icon-192.png',
    badge: payload.badge || '/badge-72.png',
    tag: payload.tag,
    renotify: payload.renotify || false,
    data: {
      url: payload.url || '/',
      ...payload.data,
    },
    actions: payload.actions || [],
    vibrate: [100, 50, 100],
    requireInteraction: payload.requireInteraction || false,
  };

  event.waitUntil(
    self.registration.showNotification(payload.title, options)
  );
});

// Handle notification click
self.addEventListener('notificationclick', (event) => {
  event.notification.close();

  const url = event.notification.data?.url || '/';

  event.waitUntil(
    clients.matchAll({ type: 'window', includeUncontrolled: true }).then((windowClients) => {
      // Focus existing window if open
      for (const client of windowClients) {
        if (client.url === url && 'focus' in client) {
          return client.focus();
        }
      }
      // Open new window
      return clients.openWindow(url);
    })
  );
});

// Handle notification action buttons
self.addEventListener('notificationclick', (event) => {
  if (event.action === 'reply') {
    // Handle reply action
    clients.openWindow(`/messages?reply=${event.notification.data?.messageId}`);
  } else if (event.action === 'dismiss') {
    event.notification.close();
  }
});
```

### Client-Side Subscription

```typescript
// lib/push/client.ts
export async function subscribeToPush(): Promise<PushSubscription | null> {
  // Check support
  if (!('serviceWorker' in navigator) || !('PushManager' in window)) {
    console.warn('Push notifications not supported');
    return null;
  }

  // Register service worker
  const registration = await navigator.serviceWorker.register('/sw.js');
  await navigator.serviceWorker.ready;

  // Check permission
  const permission = await Notification.requestPermission();
  if (permission !== 'granted') {
    console.log('Push permission denied');
    return null;
  }

  // Subscribe
  const subscription = await registration.pushManager.subscribe({
    userVisibleOnly: true,
    applicationServerKey: urlBase64ToUint8Array(process.env.NEXT_PUBLIC_VAPID_PUBLIC_KEY!),
  });

  // Send to server
  await fetch('/api/push/subscribe', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ subscription }),
  });

  return subscription;
}

export async function unsubscribeFromPush() {
  const registration = await navigator.serviceWorker.ready;
  const subscription = await registration.pushManager.getSubscription();

  if (subscription) {
    await subscription.unsubscribe();
    await fetch('/api/push/unsubscribe', {
      method: 'DELETE',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ endpoint: subscription.endpoint }),
    });
  }
}

// Helper: Convert VAPID key
function urlBase64ToUint8Array(base64String: string): Uint8Array {
  const padding = '='.repeat((4 - (base64String.length % 4)) % 4);
  const base64 = (base64String + padding).replace(/-/g, '+').replace(/_/g, '/');
  const rawData = window.atob(base64);
  return Uint8Array.from([...rawData].map(char => char.charCodeAt(0)));
}
```

### Permission UX Component

```tsx
// components/PushPermission.tsx
import { useState, useEffect } from 'react';
import { subscribeToPush, unsubscribeFromPush } from '@/lib/push/client';

export function PushPermission() {
  const [permission, setPermission] = useState<NotificationPermission>('default');
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if ('Notification' in window) {
      setPermission(Notification.permission);
    }
  }, []);

  if (!('Notification' in window)) return null;
  if (permission === 'denied') return null; // Can't ask again

  const handleEnable = async () => {
    setLoading(true);
    try {
      const sub = await subscribeToPush();
      setPermission(sub ? 'granted' : 'denied');
    } finally {
      setLoading(false);
    }
  };

  const handleDisable = async () => {
    setLoading(true);
    try {
      await unsubscribeFromPush();
      setPermission('default');
    } finally {
      setLoading(false);
    }
  };

  if (permission === 'granted') {
    return (
      <div className="flex items-center justify-between p-4 border rounded-lg">
        <div>
          <p className="font-medium">Push notifications enabled</p>
          <p className="text-sm text-gray-500">You'll receive notifications in this browser</p>
        </div>
        <button onClick={handleDisable} disabled={loading}
          className="text-sm text-red-600 hover:text-red-800">
          Disable
        </button>
      </div>
    );
  }

  return (
    <div className="flex items-center justify-between p-4 border rounded-lg bg-blue-50">
      <div>
        <p className="font-medium">Enable push notifications</p>
        <p className="text-sm text-gray-600">Get notified about important updates</p>
      </div>
      <button onClick={handleEnable} disabled={loading}
        className="px-4 py-2 bg-blue-600 text-white rounded-lg text-sm font-medium hover:bg-blue-700">
        {loading ? 'Enabling...' : 'Enable'}
      </button>
    </div>
  );
}
```

## In-App Notification Center

### Database Schema

```prisma
model Notification {
  id        String    @id @default(cuid())
  userId    String
  user      User      @relation(fields: [userId], references: [id], onDelete: Cascade)
  type      String    // "comment", "mention", "system", "billing"
  title     String
  body      String
  url       String?   // link when clicked
  read      Boolean   @default(false)
  seen      Boolean   @default(false)
  createdAt DateTime  @default(now())

  @@index([userId, read])
  @@index([userId, createdAt])
}
```

### API Endpoints

```typescript
// GET /api/notifications
app.get('/api/notifications', async (req, res) => {
  const cursor = req.query.cursor as string | undefined;
  const limit = Math.min(parseInt(req.query.limit as string) || 20, 50);

  const notifications = await prisma.notification.findMany({
    where: { userId: req.user.id },
    orderBy: { createdAt: 'desc' },
    take: limit + 1,
    ...(cursor && { cursor: { id: cursor }, skip: 1 }),
  });

  const hasMore = notifications.length > limit;
  if (hasMore) notifications.pop();

  const unreadCount = await prisma.notification.count({
    where: { userId: req.user.id, read: false },
  });

  res.json({ notifications, unreadCount, hasMore, nextCursor: hasMore ? notifications[notifications.length - 1].id : null });
});

// PATCH /api/notifications/:id/read
app.patch('/api/notifications/:id/read', async (req, res) => {
  await prisma.notification.update({
    where: { id: req.params.id, userId: req.user.id },
    data: { read: true },
  });
  res.json({ success: true });
});

// POST /api/notifications/read-all
app.post('/api/notifications/read-all', async (req, res) => {
  await prisma.notification.updateMany({
    where: { userId: req.user.id, read: false },
    data: { read: true },
  });
  res.json({ success: true });
});

// POST /api/notifications/seen
app.post('/api/notifications/seen', async (req, res) => {
  await prisma.notification.updateMany({
    where: { userId: req.user.id, seen: false },
    data: { seen: true },
  });
  res.json({ success: true });
});
```

### Real-Time with SSE

```typescript
// GET /api/notifications/stream
app.get('/api/notifications/stream', (req, res) => {
  res.writeHead(200, {
    'Content-Type': 'text/event-stream',
    'Cache-Control': 'no-cache',
    Connection: 'keep-alive',
  });

  // Send keepalive every 30s
  const keepalive = setInterval(() => res.write(':keepalive\n\n'), 30000);

  // Subscribe to user's notifications
  const handler = (notification: any) => {
    res.write(`data: ${JSON.stringify(notification)}\n\n`);
  };

  notificationBus.on(`user:${req.user.id}`, handler);

  req.on('close', () => {
    clearInterval(keepalive);
    notificationBus.off(`user:${req.user.id}`, handler);
  });
});

// Event bus (in-memory — use Redis pub/sub for multi-process)
import { EventEmitter } from 'events';
export const notificationBus = new EventEmitter();
notificationBus.setMaxListeners(1000);
```

### Create Notification Helper

```typescript
// lib/notifications/create.ts
export async function createNotification({
  userId,
  type,
  title,
  body,
  url,
  push = true,
}: {
  userId: string;
  type: string;
  title: string;
  body: string;
  url?: string;
  push?: boolean;
}) {
  // Save to database
  const notification = await prisma.notification.create({
    data: { userId, type, title, body, url },
  });

  // Emit for SSE
  notificationBus.emit(`user:${userId}`, notification);

  // Send push notification
  if (push) {
    await sendPushToUser(userId, {
      title,
      body,
      url: url || '/',
      tag: type,
    }).catch(err => console.error('Push failed:', err));
  }

  return notification;
}
```

### React Notification Center

```tsx
// components/NotificationCenter.tsx
'use client';

import { useState, useEffect, useRef } from 'react';

interface Notification {
  id: string;
  type: string;
  title: string;
  body: string;
  url?: string;
  read: boolean;
  createdAt: string;
}

export function NotificationCenter() {
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [unreadCount, setUnreadCount] = useState(0);
  const [open, setOpen] = useState(false);
  const panelRef = useRef<HTMLDivElement>(null);

  // Fetch notifications
  useEffect(() => {
    fetch('/api/notifications')
      .then(r => r.json())
      .then(data => {
        setNotifications(data.notifications);
        setUnreadCount(data.unreadCount);
      });
  }, []);

  // SSE real-time updates
  useEffect(() => {
    const es = new EventSource('/api/notifications/stream');
    es.onmessage = (event) => {
      const notification = JSON.parse(event.data);
      setNotifications(prev => [notification, ...prev]);
      setUnreadCount(prev => prev + 1);
    };
    return () => es.close();
  }, []);

  // Close on outside click
  useEffect(() => {
    const handler = (e: MouseEvent) => {
      if (panelRef.current && !panelRef.current.contains(e.target as Node)) {
        setOpen(false);
      }
    };
    document.addEventListener('mousedown', handler);
    return () => document.removeEventListener('mousedown', handler);
  }, []);

  const markRead = async (id: string) => {
    await fetch(`/api/notifications/${id}/read`, { method: 'PATCH' });
    setNotifications(prev => prev.map(n => n.id === id ? { ...n, read: true } : n));
    setUnreadCount(prev => Math.max(0, prev - 1));
  };

  const markAllRead = async () => {
    await fetch('/api/notifications/read-all', { method: 'POST' });
    setNotifications(prev => prev.map(n => ({ ...n, read: true })));
    setUnreadCount(0);
  };

  const timeAgo = (date: string) => {
    const seconds = Math.floor((Date.now() - new Date(date).getTime()) / 1000);
    if (seconds < 60) return 'just now';
    if (seconds < 3600) return `${Math.floor(seconds / 60)}m ago`;
    if (seconds < 86400) return `${Math.floor(seconds / 3600)}h ago`;
    return `${Math.floor(seconds / 86400)}d ago`;
  };

  return (
    <div className="relative" ref={panelRef}>
      {/* Bell icon */}
      <button onClick={() => setOpen(!open)} className="relative p-2">
        <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
            d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
        </svg>
        {unreadCount > 0 && (
          <span className="absolute -top-1 -right-1 bg-red-500 text-white text-xs rounded-full w-5 h-5 flex items-center justify-center">
            {unreadCount > 9 ? '9+' : unreadCount}
          </span>
        )}
      </button>

      {/* Dropdown panel */}
      {open && (
        <div className="absolute right-0 mt-2 w-96 bg-white rounded-lg shadow-xl border z-50 max-h-[480px] overflow-hidden flex flex-col">
          <div className="flex items-center justify-between px-4 py-3 border-b">
            <h3 className="font-semibold">Notifications</h3>
            {unreadCount > 0 && (
              <button onClick={markAllRead} className="text-sm text-blue-600 hover:text-blue-800">
                Mark all read
              </button>
            )}
          </div>

          <div className="overflow-y-auto flex-1">
            {notifications.length === 0 ? (
              <p className="text-center text-gray-500 py-8">No notifications yet</p>
            ) : (
              notifications.map(n => (
                <div key={n.id}
                  className={`px-4 py-3 border-b hover:bg-gray-50 cursor-pointer ${!n.read ? 'bg-blue-50' : ''}`}
                  onClick={() => {
                    markRead(n.id);
                    if (n.url) window.location.href = n.url;
                  }}
                >
                  <div className="flex justify-between items-start">
                    <p className="font-medium text-sm">{n.title}</p>
                    <span className="text-xs text-gray-400 whitespace-nowrap ml-2">
                      {timeAgo(n.createdAt)}
                    </span>
                  </div>
                  <p className="text-sm text-gray-600 mt-1">{n.body}</p>
                </div>
              ))
            )}
          </div>
        </div>
      )}
    </div>
  );
}
```

## SMS via Twilio

### Setup

```bash
npm install twilio
```

```typescript
// lib/sms/twilio.ts
import twilio from 'twilio';

const client = twilio(
  process.env.TWILIO_ACCOUNT_SID,
  process.env.TWILIO_AUTH_TOKEN
);

export async function sendSMS({
  to,
  body,
  from = process.env.TWILIO_PHONE_NUMBER!,
}: {
  to: string;
  body: string;
  from?: string;
}) {
  const message = await client.messages.create({
    to,
    from,
    body,
  });

  return { sid: message.sid, status: message.status };
}
```

### Common SMS Patterns

```typescript
// 2FA verification code
export async function send2FACode(phoneNumber: string): Promise<string> {
  const code = Math.floor(100000 + Math.random() * 900000).toString();

  // Store code with expiry (5 minutes)
  await redis.setex(`2fa:${phoneNumber}`, 300, code);

  await sendSMS({
    to: phoneNumber,
    body: `Your verification code is: ${code}. It expires in 5 minutes.`,
  });

  return code;
}

// Verify 2FA code
export async function verify2FACode(phoneNumber: string, code: string): Promise<boolean> {
  const stored = await redis.get(`2fa:${phoneNumber}`);
  if (stored === code) {
    await redis.del(`2fa:${phoneNumber}`);
    return true;
  }
  return false;
}

// Order status update
export async function sendOrderStatusSMS(
  phoneNumber: string,
  orderId: string,
  status: string,
  trackingUrl?: string
) {
  const messages: Record<string, string> = {
    shipped: `Your order #${orderId} has shipped! ${trackingUrl ? `Track it: ${trackingUrl}` : ''}`,
    delivered: `Your order #${orderId} has been delivered. Enjoy!`,
    refunded: `Your refund for order #${orderId} has been processed.`,
  };

  const body = messages[status] || `Order #${orderId} update: ${status}`;

  await sendSMS({ to: phoneNumber, body });
}
```

## Multi-Channel Notification Orchestrator

```typescript
// lib/notifications/orchestrator.ts
type Channel = 'in-app' | 'email' | 'push' | 'sms';

interface NotificationConfig {
  type: string;
  channels: Channel[];
  emailTemplate?: string;
  priority?: 'low' | 'normal' | 'high' | 'urgent';
}

const NOTIFICATION_ROUTING: Record<string, NotificationConfig> = {
  'comment.new': {
    type: 'comment',
    channels: ['in-app', 'push'],
    priority: 'normal',
  },
  'mention': {
    type: 'mention',
    channels: ['in-app', 'push', 'email'],
    emailTemplate: 'mention',
    priority: 'high',
  },
  'order.confirmed': {
    type: 'order',
    channels: ['in-app', 'email', 'sms'],
    emailTemplate: 'order-confirmation',
    priority: 'high',
  },
  'payment.failed': {
    type: 'billing',
    channels: ['in-app', 'email', 'push'],
    emailTemplate: 'payment-failed',
    priority: 'urgent',
  },
  'security.login': {
    type: 'security',
    channels: ['email', 'push'],
    emailTemplate: 'new-login',
    priority: 'urgent',
  },
};

export async function notify(
  event: string,
  userId: string,
  data: Record<string, any>
) {
  const config = NOTIFICATION_ROUTING[event];
  if (!config) {
    console.warn(`No notification config for event: ${event}`);
    return;
  }

  // Check user preferences
  const prefs = await getUserNotificationPreferences(userId);

  // Check quiet hours
  if (prefs.quietHours && isQuietHours(prefs)) {
    // Queue for later delivery (except urgent)
    if (config.priority !== 'urgent') {
      await queueForLater(event, userId, data, prefs.quietHours.end);
      return;
    }
  }

  const promises: Promise<any>[] = [];

  for (const channel of config.channels) {
    // Skip if user disabled this channel for this type
    if (prefs.disabled?.[config.type]?.includes(channel)) continue;

    switch (channel) {
      case 'in-app':
        promises.push(createNotification({
          userId,
          type: config.type,
          title: data.title,
          body: data.body,
          url: data.url,
          push: false, // handled separately
        }));
        break;

      case 'email':
        if (config.emailTemplate) {
          promises.push(queueEmail(config.emailTemplate, data.email, data));
        }
        break;

      case 'push':
        promises.push(sendPushToUser(userId, {
          title: data.title,
          body: data.body,
          url: data.url,
          tag: config.type,
        }));
        break;

      case 'sms':
        if (data.phone) {
          promises.push(sendSMS({ to: data.phone, body: data.smsBody || data.body }));
        }
        break;
    }
  }

  await Promise.allSettled(promises);
}
```

### Notification Preferences

```typescript
// GET /api/notification-preferences
app.get('/api/notification-preferences', async (req, res) => {
  const prefs = await prisma.notificationPreference.findUnique({
    where: { userId: req.user.id },
  });

  res.json(prefs || {
    email: true,
    push: true,
    sms: false,
    disabled: {},
    quietHours: null,
    digest: 'none',
  });
});

// PATCH /api/notification-preferences
app.patch('/api/notification-preferences', async (req, res) => {
  const prefs = await prisma.notificationPreference.upsert({
    where: { userId: req.user.id },
    create: { userId: req.user.id, ...req.body },
    update: req.body,
  });

  res.json(prefs);
});
```

## Best Practices

1. **Ask permission at the right time** — don't prompt on first visit. Wait until the user takes an action that benefits from notifications.
2. **Explain why** — "Get notified when someone replies to your comment" is better than a generic browser dialog.
3. **Respect preferences** — always let users choose channels per notification type.
4. **Quiet hours** — respect time zones and user-defined quiet hours.
5. **Batch related notifications** — "3 new comments" instead of 3 separate notifications.
6. **Handle failures gracefully** — expired push subscriptions, invalid phone numbers, bounced emails.
7. **Rate limit** — never send more than a few notifications per minute to a single user.
8. **Clean up** — remove expired push subscriptions, prune old notification records.
