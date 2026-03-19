---
name: live-updates
description: >
  Implement real-time live updates with Server-Sent Events (SSE), WebSocket
  pub/sub, and polling fallbacks. Covers live dashboards, activity feeds,
  real-time notifications, live search, and data synchronization.
  Triggers: "live updates", "real-time dashboard", "activity feed",
  "live data", "sse", "server-sent events", "live notifications",
  "real-time sync", "live feed".
  NOT for: chat systems (use chat-system) or collaborative editing.
version: 1.0.0
argument-hint: "[sse|websocket|dashboard|feed]"
allowed-tools: Read, Grep, Glob, Write, Edit, Bash
---

# Live Updates

Implement real-time live updates for dashboards, feeds, and data sync.

## Approach Selection

| Method | Best For | Pros | Cons |
|--------|----------|------|------|
| **SSE** | One-way server→client | Simple, auto-reconnect, HTTP/2 | One direction only |
| **WebSocket** | Bidirectional, high frequency | Full duplex, low latency | More complex |
| **Long Polling** | Fallback, simple clients | Universal compatibility | Higher latency, more requests |
| **Polling** | Low frequency updates | Simplest to implement | Wasteful, high latency |

**Rule of thumb**: SSE for dashboards/feeds. WebSocket for interactive features. Polling as last resort.

## Server-Sent Events (SSE)

### Server Implementation

```typescript
// server/sse.ts
import { EventEmitter } from 'events';

// Event bus (use Redis pub/sub for multi-process)
export const eventBus = new EventEmitter();
eventBus.setMaxListeners(10000);

// SSE endpoint
app.get('/api/events', (req, res) => {
  // Auth
  const userId = req.user?.id;
  if (!userId) return res.status(401).end();

  // SSE headers
  res.writeHead(200, {
    'Content-Type': 'text/event-stream',
    'Cache-Control': 'no-cache',
    'Connection': 'keep-alive',
    'X-Accel-Buffering': 'no', // Disable nginx buffering
  });

  // Send initial connection event
  res.write(`event: connected\ndata: ${JSON.stringify({ userId })}\n\n`);

  // Keepalive every 30 seconds
  const keepalive = setInterval(() => {
    res.write(`:keepalive ${Date.now()}\n\n`);
  }, 30000);

  // Subscribe to user events
  const handler = (event: { type: string; data: any }) => {
    res.write(`event: ${event.type}\ndata: ${JSON.stringify(event.data)}\n\n`);
  };

  eventBus.on(`user:${userId}`, handler);

  // Subscribe to broadcast events
  const broadcastHandler = (event: { type: string; data: any }) => {
    res.write(`event: ${event.type}\ndata: ${JSON.stringify(event.data)}\n\n`);
  };

  eventBus.on('broadcast', broadcastHandler);

  // Cleanup on disconnect
  req.on('close', () => {
    clearInterval(keepalive);
    eventBus.off(`user:${userId}`, handler);
    eventBus.off('broadcast', broadcastHandler);
  });
});
```

### Emitting Events

```typescript
// Emit to specific user
export function emitToUser(userId: string, type: string, data: any) {
  eventBus.emit(`user:${userId}`, { type, data });
}

// Emit to all connected clients
export function emitBroadcast(type: string, data: any) {
  eventBus.emit('broadcast', { type, data });
}

// Usage examples
emitToUser(userId, 'order:status', { orderId: '123', status: 'shipped' });
emitBroadcast('stats:updated', { activeUsers: 42, revenue: 15000 });
```

### Client-Side SSE

```typescript
// lib/sse.ts
export class LiveUpdates {
  private eventSource: EventSource | null = null;
  private listeners = new Map<string, Set<(data: any) => void>>();
  private reconnectDelay = 1000;

  connect(token: string) {
    this.eventSource = new EventSource(`/api/events?token=${token}`);

    this.eventSource.onopen = () => {
      this.reconnectDelay = 1000; // Reset on successful connect
    };

    this.eventSource.onerror = () => {
      this.eventSource?.close();
      setTimeout(() => this.connect(token), this.reconnectDelay);
      this.reconnectDelay = Math.min(this.reconnectDelay * 2, 30000);
    };

    // Route events to listeners
    this.eventSource.addEventListener('message', (event) => {
      const data = JSON.parse(event.data);
      this.notify('message', data);
    });

    // Register named event listeners
    for (const type of this.listeners.keys()) {
      this.eventSource.addEventListener(type, (event) => {
        const data = JSON.parse((event as MessageEvent).data);
        this.notify(type, data);
      });
    }
  }

  on(type: string, callback: (data: any) => void) {
    if (!this.listeners.has(type)) {
      this.listeners.set(type, new Set());

      // Register with EventSource if already connected
      if (this.eventSource) {
        this.eventSource.addEventListener(type, (event) => {
          const data = JSON.parse((event as MessageEvent).data);
          this.notify(type, data);
        });
      }
    }
    this.listeners.get(type)!.add(callback);

    // Return unsubscribe function
    return () => {
      this.listeners.get(type)?.delete(callback);
    };
  }

  private notify(type: string, data: any) {
    this.listeners.get(type)?.forEach(cb => cb(data));
  }

  disconnect() {
    this.eventSource?.close();
    this.eventSource = null;
  }
}
```

### React Hook

```tsx
// hooks/useLiveUpdates.ts
import { useEffect, useRef, useState } from 'react';

export function useLiveUpdates<T>(
  eventType: string,
  initialData: T
): T {
  const [data, setData] = useState<T>(initialData);

  useEffect(() => {
    const es = new EventSource('/api/events');

    es.addEventListener(eventType, (event) => {
      const parsed = JSON.parse((event as MessageEvent).data);
      setData(parsed);
    });

    es.onerror = () => {
      es.close();
      // Reconnect after 3 seconds
      setTimeout(() => {
        // Re-run effect
      }, 3000);
    };

    return () => es.close();
  }, [eventType]);

  return data;
}

// Usage
function Dashboard() {
  const stats = useLiveUpdates('stats:updated', { activeUsers: 0, revenue: 0 });

  return (
    <div>
      <p>Active Users: {stats.activeUsers}</p>
      <p>Revenue: ${stats.revenue}</p>
    </div>
  );
}
```

## Live Activity Feed

```tsx
// components/ActivityFeed.tsx
'use client';
import { useState, useEffect } from 'react';

interface Activity {
  id: string;
  type: string;
  actor: { name: string; avatar?: string };
  action: string;
  target?: string;
  targetUrl?: string;
  createdAt: string;
}

export function ActivityFeed() {
  const [activities, setActivities] = useState<Activity[]>([]);

  // Load initial activities
  useEffect(() => {
    fetch('/api/activities?limit=20')
      .then(r => r.json())
      .then(data => setActivities(data.activities));
  }, []);

  // Listen for new activities via SSE
  useEffect(() => {
    const es = new EventSource('/api/events');

    es.addEventListener('activity:new', (event) => {
      const activity = JSON.parse((event as MessageEvent).data);
      setActivities(prev => [activity, ...prev].slice(0, 50)); // Keep latest 50
    });

    return () => es.close();
  }, []);

  const timeAgo = (date: string) => {
    const seconds = Math.floor((Date.now() - new Date(date).getTime()) / 1000);
    if (seconds < 60) return 'just now';
    if (seconds < 3600) return `${Math.floor(seconds / 60)}m ago`;
    if (seconds < 86400) return `${Math.floor(seconds / 3600)}h ago`;
    return `${Math.floor(seconds / 86400)}d ago`;
  };

  return (
    <div className="space-y-1">
      {activities.map(activity => (
        <div key={activity.id} className="flex items-start gap-3 py-3 border-b">
          <div className="w-8 h-8 rounded-full bg-gray-200 flex-shrink-0" />
          <div className="flex-1 min-w-0">
            <p className="text-sm">
              <span className="font-medium">{activity.actor.name}</span>
              {' '}{activity.action}{' '}
              {activity.target && (
                activity.targetUrl
                  ? <a href={activity.targetUrl} className="font-medium text-blue-600">{activity.target}</a>
                  : <span className="font-medium">{activity.target}</span>
              )}
            </p>
            <p className="text-xs text-gray-500 mt-0.5">{timeAgo(activity.createdAt)}</p>
          </div>
        </div>
      ))}
    </div>
  );
}
```

## Live Dashboard with Auto-Refresh

```tsx
// components/LiveDashboard.tsx
'use client';
import { useState, useEffect } from 'react';

interface DashboardStats {
  activeUsers: number;
  todayRevenue: number;
  ordersToday: number;
  conversionRate: number;
}

export function LiveDashboard() {
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [lastUpdated, setLastUpdated] = useState<Date>(new Date());

  // Initial load + SSE updates
  useEffect(() => {
    // Fetch initial
    fetch('/api/dashboard/stats')
      .then(r => r.json())
      .then(data => {
        setStats(data);
        setLastUpdated(new Date());
      });

    // Listen for real-time updates
    const es = new EventSource('/api/events');

    es.addEventListener('dashboard:stats', (event) => {
      const data = JSON.parse((event as MessageEvent).data);
      setStats(data);
      setLastUpdated(new Date());
    });

    return () => es.close();
  }, []);

  if (!stats) return <div>Loading...</div>;

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-xl font-bold">Live Dashboard</h2>
        <div className="flex items-center gap-2 text-sm text-gray-500">
          <span className="w-2 h-2 bg-green-500 rounded-full animate-pulse" />
          Updated {lastUpdated.toLocaleTimeString()}
        </div>
      </div>

      <div className="grid grid-cols-4 gap-4">
        <StatCard label="Active Users" value={stats.activeUsers} live />
        <StatCard label="Revenue Today" value={`$${(stats.todayRevenue / 100).toFixed(2)}`} />
        <StatCard label="Orders Today" value={stats.ordersToday} />
        <StatCard label="Conversion" value={`${stats.conversionRate}%`} />
      </div>
    </div>
  );
}

function StatCard({ label, value, live }: { label: string; value: string | number; live?: boolean }) {
  return (
    <div className="bg-white border rounded-lg p-6">
      <p className="text-sm text-gray-600 flex items-center gap-1">
        {label}
        {live && <span className="w-1.5 h-1.5 bg-green-500 rounded-full" />}
      </p>
      <p className="text-3xl font-bold mt-2">{value}</p>
    </div>
  );
}
```

## Redis Pub/Sub for Multi-Process SSE

```typescript
// lib/redis-pubsub.ts
import IORedis from 'ioredis';

const pub = new IORedis(process.env.REDIS_URL!);
const sub = new IORedis(process.env.REDIS_URL!);

// Subscribe to channels
sub.subscribe('events:broadcast', 'events:user:*');

sub.on('message', (channel, message) => {
  const event = JSON.parse(message);

  if (channel === 'events:broadcast') {
    eventBus.emit('broadcast', event);
  } else {
    const userId = channel.replace('events:user:', '');
    eventBus.emit(`user:${userId}`, event);
  }
});

// Publish functions
export function publishToUser(userId: string, type: string, data: any) {
  pub.publish(`events:user:${userId}`, JSON.stringify({ type, data }));
}

export function publishBroadcast(type: string, data: any) {
  pub.publish('events:broadcast', JSON.stringify({ type, data }));
}
```

## Polling Fallback

```typescript
// For environments where SSE/WebSocket isn't available
function usePolling<T>(url: string, interval: number = 5000): T | null {
  const [data, setData] = useState<T | null>(null);

  useEffect(() => {
    let active = true;

    const poll = async () => {
      try {
        const res = await fetch(url);
        if (active) setData(await res.json());
      } catch (e) {
        console.error('Polling error:', e);
      }
    };

    poll(); // Initial fetch
    const timer = setInterval(poll, interval);

    return () => {
      active = false;
      clearInterval(timer);
    };
  }, [url, interval]);

  return data;
}
```

## Best Practices

1. **SSE over WebSocket for one-way data** — simpler, auto-reconnect, works with HTTP/2
2. **Keep connections lean** — send only changed data, not full state
3. **Implement reconnection** — exponential backoff with max delay
4. **Use event types** — named events let clients subscribe selectively
5. **Keepalive comments** — prevent proxy timeouts with `:keepalive\n\n` every 30s
6. **Redis pub/sub for scaling** — required when running multiple server instances
7. **Graceful degradation** — fall back to polling if SSE fails
8. **Rate limit events** — batch rapid updates to avoid flooding clients
