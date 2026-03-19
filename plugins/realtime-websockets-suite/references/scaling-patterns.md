# Real-time Scaling Patterns

Patterns for scaling WebSocket and real-time systems from 100 to 1M+ connections.

---

## Architecture Tiers

### Tier 1: Single Server (0-10K connections)

```
┌─────────────────┐
│   Load Balancer  │  (optional at this scale)
└────────┬────────┘
         │
┌────────▼────────┐
│  App Server      │
│  Socket.io/ws    │
│  In-memory state │
└─────────────────┘
```

- In-memory rooms, presence, message routing
- No external dependencies beyond your database
- Good for MVPs and small-medium apps

### Tier 2: Horizontal with Redis (10K-100K connections)

```
┌─────────────────┐
│   Load Balancer  │  (sticky sessions required)
└────────┬────────┘
    ┌────┼────┐
    ▼    ▼    ▼
┌──────┐┌──────┐┌──────┐
│ App 1││ App 2││ App 3│
└──┬───┘└──┬───┘└──┬───┘
   └───────┼───────┘
     ┌─────▼─────┐
     │   Redis    │
     │  Pub/Sub   │
     └───────────┘
```

- Redis adapter for Socket.io cross-server messaging
- Sticky sessions (IP hash or cookie-based)
- Redis stores ephemeral state (presence, typing indicators)
- Database for persistent data (messages, rooms)

### Tier 3: Message Broker (100K-1M connections)

```
┌──────────────┐
│ Load Balancer│
└──────┬───────┘
  ┌────┼────┐
  ▼    ▼    ▼
┌────┐┌────┐┌────┐
│WS 1││WS 2││WS 3│  ← WebSocket servers (stateless routing)
└──┬─┘└──┬─┘└──┬─┘
   └──────┼──────┘
    ┌─────▼─────┐
    │ NATS / Kafka│  ← Message broker
    │ / RabbitMQ  │
    └─────┬─────┘
    ┌─────▼─────┐
    │ Worker Pool │  ← Business logic processing
    └───────────┘
```

- WebSocket servers only handle connections and routing
- Message broker handles pub/sub, ordering, durability
- Worker pool processes business logic independently
- Horizontal scaling at every layer

### Tier 4: Managed Service (1M+ connections)

| Service | Connections | Pricing | Best For |
|---------|------------|---------|----------|
| **Pusher** | 500K+ | Per-message + connections | Simple pub/sub |
| **Ably** | Millions | Per-message | Enterprise reliability |
| **AWS AppSync** | Auto-scaling | Per-operation | GraphQL subscriptions |
| **PubNub** | Millions | Per-transaction | IoT, chat |
| **Liveblocks** | Auto-scaling | Per-MAU | Collaborative apps |
| **PartyKit** | Edge-deployed | Per-request | Multiplayer, stateful rooms |

---

## Redis Adapter Patterns

### Socket.io + Redis

```typescript
import { Server } from 'socket.io';
import { createAdapter } from '@socket.io/redis-adapter';
import { createClient } from 'redis';

const pubClient = createClient({ url: process.env.REDIS_URL });
const subClient = pubClient.duplicate();
await Promise.all([pubClient.connect(), subClient.connect()]);

const io = new Server(httpServer, {
  adapter: createAdapter(pubClient, subClient),
});

// All Socket.io operations now work across servers:
// io.to('room').emit(...)  — broadcasts to all servers
// io.in('room').fetchSockets()  — aggregates from all servers
// socket.join('room')  — synced via Redis
```

### Redis Streams for Message History

```typescript
import { createClient } from 'redis';
const redis = createClient({ url: process.env.REDIS_URL });

// Store message in stream
async function storeMessage(roomId: string, message: any) {
  await redis.xAdd(`room:${roomId}:messages`, '*', {
    data: JSON.stringify(message),
  });

  // Trim to last 1000 messages
  await redis.xTrim(`room:${roomId}:messages`, 'MAXLEN', 1000);
}

// Fetch recent messages (for new connections)
async function getRecentMessages(roomId: string, count = 50) {
  const messages = await redis.xRevRange(
    `room:${roomId}:messages`,
    '+', '-',
    { COUNT: count }
  );

  return messages.reverse().map(m => JSON.parse(m.message.data));
}
```

### Redis for Presence

```typescript
// Set user online with TTL
async function setOnline(userId: string, serverId: string) {
  await redis.hSet(`presence:${userId}`, {
    serverId,
    lastSeen: Date.now().toString(),
    status: 'online',
  });
  await redis.expire(`presence:${userId}`, 60); // Auto-expire if no heartbeat
}

// Heartbeat (call every 30s per connected client)
async function heartbeat(userId: string) {
  await redis.hSet(`presence:${userId}`, 'lastSeen', Date.now().toString());
  await redis.expire(`presence:${userId}`, 60);
}

// Get online users for a room
async function getOnlineUsers(roomId: string) {
  const memberIds = await redis.sMembers(`room:${roomId}:members`);
  const pipeline = redis.multi();

  for (const id of memberIds) {
    pipeline.hGetAll(`presence:${id}`);
  }

  const results = await pipeline.exec();
  return memberIds
    .map((id, i) => ({ id, ...(results[i] as any) }))
    .filter(u => u.status === 'online');
}
```

---

## Sticky Sessions

### Why Required

Socket.io uses HTTP long-polling as initial transport before upgrading to WebSocket. The polling requests must reach the same server that holds the session. Without sticky sessions:

```
Request 1 → Server A (creates session abc)
Request 2 → Server B (no session abc → error)
```

### Nginx (IP Hash)

```nginx
upstream websocket_backend {
    ip_hash;
    server app1:3001;
    server app2:3001;
    server app3:3001;
}

server {
    location /socket.io/ {
        proxy_pass http://websocket_backend;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_read_timeout 86400;
        proxy_send_timeout 86400;
    }
}
```

### Nginx (Cookie-Based — More Reliable)

```nginx
upstream websocket_backend {
    server app1:3001;
    server app2:3001;
    server app3:3001;
    sticky cookie srv_id expires=1h domain=.example.com path=/;
}
```

### AWS ALB

- Use Application Load Balancer (not Classic or Network)
- Enable stickiness on the target group
- Set "Stickiness type" to "Application-based cookie" or "Load balancer generated cookie"
- Duration: at least as long as your longest expected session

### Skip Sticky Sessions Entirely

If you configure Socket.io to use WebSocket-only transport, sticky sessions aren't needed:

```typescript
// Client
const socket = io(url, {
  transports: ['websocket'], // Skip polling entirely
});

// Server — still works with round-robin load balancing
```

**Trade-off**: Loses HTTP polling fallback. If WebSocket is blocked (corporate proxies), connection fails entirely.

---

## Connection Management

### Graceful Shutdown

```typescript
async function gracefulShutdown(io: Server) {
  console.log('Starting graceful shutdown...');

  // 1. Stop accepting new connections
  io.close();

  // 2. Notify connected clients
  io.emit('server:shutdown', {
    message: 'Server restarting. You will be reconnected automatically.',
    reconnectIn: 5000,
  });

  // 3. Wait for in-flight messages to complete
  await new Promise(resolve => setTimeout(resolve, 5000));

  // 4. Force-close remaining connections
  const sockets = await io.fetchSockets();
  for (const socket of sockets) {
    socket.disconnect(true);
  }

  // 5. Close Redis connections
  await pubClient.quit();
  await subClient.quit();

  process.exit(0);
}

process.on('SIGTERM', () => gracefulShutdown(io));
process.on('SIGINT', () => gracefulShutdown(io));
```

### Connection Draining (Zero-Downtime Deploys)

```
1. Deploy new server instance (Server B)
2. Health check passes on Server B
3. Remove Server A from load balancer (stop new connections)
4. Server A emits "reconnect-to" event with Server B address
5. Clients reconnect to Server B via load balancer
6. Wait for Server A connections to drain (timeout: 30s)
7. Kill Server A
```

### Heartbeat Configuration

```typescript
const io = new Server(httpServer, {
  pingTimeout: 60000,    // Wait 60s for pong before considering dead
  pingInterval: 25000,   // Send ping every 25s
});
```

**Tuning**:
- `pingInterval`: Lower = faster dead detection, more overhead. 25-30s is standard.
- `pingTimeout`: Must be > client's worst-case network latency. 60s handles mobile/spotty connections.
- Total dead detection time = `pingInterval` + `pingTimeout` (85s with defaults).

---

## Message Ordering & Delivery

### At-Most-Once (Default WebSocket)

Messages can be lost if connection drops during transmission. No guarantees.

```
Client → Server: "Hello"     ← might be lost
```

### At-Least-Once (With Acknowledgments)

```typescript
// Socket.io acknowledgment
socket.emit('message', { text: 'hello' }, (response) => {
  if (response.status === 'ok') {
    // Server confirmed receipt
  } else {
    // Retry
  }
});

// Server side
socket.on('message', (data, callback) => {
  try {
    await saveMessage(data);
    callback({ status: 'ok' });
  } catch (err) {
    callback({ status: 'error' });
  }
});
```

### Exactly-Once (Idempotent Messages)

```typescript
// Client includes unique message ID
socket.emit('message', {
  id: crypto.randomUUID(),  // Client-generated
  text: 'hello',
  timestamp: Date.now(),
});

// Server deduplicates
const processed = new Set<string>(); // Use Redis in production

socket.on('message', async (data) => {
  if (processed.has(data.id)) return; // Already processed
  processed.add(data.id);

  await saveMessage(data);
  socket.to(data.roomId).emit('message', data);
});
```

### Message Ordering

WebSocket guarantees ordering within a single connection (TCP). But across reconnections or multiple servers:

```typescript
// Use sequence numbers for ordering
let seq = 0;

socket.emit('message', {
  seq: ++seq,
  text: 'hello',
});

// Server: Buffer and reorder if out-of-sequence
// Client: Request missing messages by sequence gap
```

---

## Backpressure Handling

When a client can't consume messages fast enough:

```typescript
// Server-side: Check buffered amount before sending
function safeSend(ws: WebSocket, data: string) {
  if (ws.bufferedAmount > 1024 * 1024) { // 1MB buffer
    console.warn('Client buffer full, dropping message');
    return false;
  }
  ws.send(data);
  return true;
}

// Socket.io: Use volatile emit for droppable messages
io.to('room').volatile.emit('cursor-position', { x: 100, y: 200 });
// If the client isn't ready, the message is silently dropped
```

### Client-Side Throttling

```typescript
// Throttle outgoing messages (e.g., cursor position)
function throttle(fn: Function, delay: number) {
  let timer: NodeJS.Timeout | null = null;
  let lastArgs: any[] | null = null;

  return (...args: any[]) => {
    lastArgs = args;
    if (!timer) {
      timer = setTimeout(() => {
        fn(...lastArgs!);
        timer = null;
        lastArgs = null;
      }, delay);
    }
  };
}

const sendCursorPosition = throttle((pos) => {
  socket.volatile.emit('cursor', pos);
}, 50); // Max 20 updates/sec

document.addEventListener('mousemove', (e) => {
  sendCursorPosition({ x: e.clientX, y: e.clientY });
});
```

---

## Monitoring & Observability

### Key Metrics to Track

| Metric | What It Tells You | Alert Threshold |
|--------|-------------------|----------------|
| Active connections | Current load | >80% of capacity |
| Connection rate | Growth/spikes | >2x normal |
| Disconnection rate | Client/network issues | >10% of active |
| Message throughput | Messages/sec | >80% of capacity |
| Message latency (p95) | Processing speed | >500ms |
| Error rate | System health | >1% |
| Memory usage | Per-connection overhead | >80% of available |
| Redis pub/sub lag | Cross-server latency | >100ms |

### Socket.io Instrumentation

```typescript
import { instrument } from '@socket.io/admin-ui';

// Enable admin UI
instrument(io, {
  auth: {
    type: 'basic',
    username: 'admin',
    password: '$2b$10$...',  // bcrypt hash
  },
  mode: 'development',
});

// Access at https://admin.socket.io with your server URL
```

### Custom Metrics with Prometheus

```typescript
import { Counter, Gauge, Histogram } from 'prom-client';

const connectedClients = new Gauge({
  name: 'ws_connected_clients',
  help: 'Number of connected WebSocket clients',
});

const messagesTotal = new Counter({
  name: 'ws_messages_total',
  help: 'Total WebSocket messages',
  labelNames: ['type', 'direction'],
});

const messageLatency = new Histogram({
  name: 'ws_message_latency_seconds',
  help: 'Message processing latency',
  buckets: [0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1],
});

io.on('connection', (socket) => {
  connectedClients.inc();

  socket.onAny((event, ...args) => {
    messagesTotal.inc({ type: event, direction: 'inbound' });
  });

  socket.on('disconnect', () => {
    connectedClients.dec();
  });
});
```

---

## Common Scaling Mistakes

1. **Storing connection state in memory only** — When a server dies, all state is lost. Use Redis for ephemeral state.

2. **Broadcasting to all connections** — `io.emit(...)` sends to every connected client. Use rooms to scope broadcasts.

3. **No message size limits** — A single large message can exhaust server memory. Set `maxHttpBufferSize`.

4. **Synchronous message processing** — Blocking the event loop kills all connections on that server. Use worker threads or a message queue.

5. **No backpressure** — Sending faster than clients can consume causes memory buildup. Use `volatile.emit` for droppable messages.

6. **Forgetting cleanup** — Event listeners, intervals, and subscriptions leak if not cleaned up on disconnect.

7. **No graceful shutdown** — `process.exit()` drops all connections. Drain connections first.

8. **Polling-first transport** — Socket.io defaults to polling then upgrades. For known-good environments, use `transports: ['websocket']` to skip polling.

9. **No reconnection backoff** — Clients reconnecting immediately after a server crash creates a thundering herd. Use exponential backoff with jitter.

10. **Testing with a single server** — Everything works until you add a second server. Test with at least 2 servers and Redis from the start.
