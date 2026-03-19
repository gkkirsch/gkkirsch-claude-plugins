# Scaling Real-Time Applications

## Architecture Tiers

### Tier 1: Single Server (0-10K connections)

```
Clients → Express + Socket.io → In-memory state
```

- **When**: MVP, early product, single-region
- **Limits**: ~10K concurrent WebSocket connections on a single 2GB server
- **Stack**: Express + Socket.io, in-memory rooms, in-memory event bus
- **Cost**: $5-20/mo (single VPS or Heroku dyno)

### Tier 2: Multi-Server with Redis (10K-100K connections)

```
Clients → Load Balancer (sticky sessions)
        → Server 1 ←→ Redis Pub/Sub ←→ Server 2
                                     ←→ Server 3
```

- **When**: Growing product, need horizontal scaling
- **Limits**: ~100K connections across 5-10 servers
- **Stack**: Socket.io + `@socket.io/redis-adapter`, Redis, nginx/ALB
- **Cost**: $100-500/mo

### Tier 3: Dedicated WebSocket Layer (100K-1M+ connections)

```
Clients → CDN/Edge → WebSocket Gateway
                    → App Servers → Database
                    → Redis Cluster → Worker Servers
```

- **When**: High-scale product, multi-region
- **Stack**: Dedicated WS servers, Redis Cluster, message queue, app servers
- **Cost**: $1,000+/mo

## Redis Adapter Setup (Tier 2)

### Socket.io + Redis

```typescript
import { Server } from 'socket.io';
import { createAdapter } from '@socket.io/redis-adapter';
import { createClient } from 'redis';

const io = new Server(httpServer);

const pubClient = createClient({ url: process.env.REDIS_URL });
const subClient = pubClient.duplicate();

await Promise.all([pubClient.connect(), subClient.connect()]);

io.adapter(createAdapter(pubClient, subClient));

// Now these work across all server instances:
io.to('room-123').emit('message', data);     // Room broadcast
io.emit('announcement', data);                // Global broadcast
io.to(`user:${userId}`).emit('notification'); // Targeted message
```

### Redis Streams for Event Log

```typescript
import { createClient } from 'redis';

const redis = createClient({ url: process.env.REDIS_URL });

// Publish event to stream
async function publishEvent(channel: string, event: any) {
  await redis.xAdd(channel, '*', {
    type: event.type,
    data: JSON.stringify(event.data),
    timestamp: Date.now().toString(),
  });

  // Trim to keep last 10,000 events
  await redis.xTrim(channel, 'MAXLEN', 10000);
}

// Consumer group for reliable processing
async function consumeEvents(channel: string, group: string, consumer: string) {
  // Create group if not exists
  try {
    await redis.xGroupCreate(channel, group, '0', { MKSTREAM: true });
  } catch (e) {
    // Group already exists
  }

  while (true) {
    const results = await redis.xReadGroup(group, consumer, {
      key: channel,
      id: '>',
    }, { COUNT: 10, BLOCK: 5000 });

    if (results) {
      for (const { messages } of results) {
        for (const { id, message } of messages) {
          await processEvent(message);
          await redis.xAck(channel, group, id);
        }
      }
    }
  }
}
```

## Load Balancer Configuration

### Nginx (WebSocket + Sticky Sessions)

```nginx
upstream websocket_backend {
    # IP hash for sticky sessions
    ip_hash;

    server ws-server-1:3001;
    server ws-server-2:3001;
    server ws-server-3:3001;

    # Health checks
    keepalive 64;
}

server {
    listen 443 ssl http2;
    server_name ws.example.com;

    # WebSocket location
    location /socket.io/ {
        proxy_pass http://websocket_backend;
        proxy_http_version 1.1;

        # WebSocket upgrade headers
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";

        # Preserve client info
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # Timeouts (WebSocket connections are long-lived)
        proxy_read_timeout 86400s;   # 24 hours
        proxy_send_timeout 86400s;
        proxy_connect_timeout 10s;

        # Disable buffering for real-time
        proxy_buffering off;
    }

    # SSE location
    location /api/events {
        proxy_pass http://websocket_backend;
        proxy_http_version 1.1;
        proxy_set_header Connection '';

        # Critical for SSE
        proxy_buffering off;
        proxy_cache off;
        chunked_transfer_encoding off;

        proxy_read_timeout 86400s;
    }
}
```

### AWS ALB

```
- Enable sticky sessions (application cookie or duration)
- WebSocket connections: automatically supported
- Idle timeout: set to 3600s (1 hour) for WebSocket
- Target group health check: HTTP GET /health on port 3001
```

### Heroku

```
# WebSocket works out of the box on Heroku
# but connections timeout after 55 seconds of inactivity
# Solution: implement ping/pong keepalive

const io = new Server(httpServer, {
  pingTimeout: 30000,   // 30s timeout
  pingInterval: 20000,  // Ping every 20s (under Heroku's 55s limit)
});
```

## Connection Management

### Connection Pooling

```typescript
class ConnectionManager {
  private connections = new Map<string, Set<string>>(); // userId → socketIds
  private metadata = new Map<string, any>();             // socketId → metadata

  addConnection(userId: string, socketId: string, meta: any) {
    if (!this.connections.has(userId)) {
      this.connections.set(userId, new Set());
    }
    this.connections.get(userId)!.add(socketId);
    this.metadata.set(socketId, { userId, ...meta });
  }

  removeConnection(socketId: string) {
    const meta = this.metadata.get(socketId);
    if (meta) {
      this.connections.get(meta.userId)?.delete(socketId);
      if (this.connections.get(meta.userId)?.size === 0) {
        this.connections.delete(meta.userId);
      }
      this.metadata.delete(socketId);
    }
  }

  getUserConnections(userId: string): string[] {
    return Array.from(this.connections.get(userId) || []);
  }

  isUserOnline(userId: string): boolean {
    return (this.connections.get(userId)?.size || 0) > 0;
  }

  getOnlineUserCount(): number {
    return this.connections.size;
  }

  getTotalConnections(): number {
    return this.metadata.size;
  }
}
```

### Graceful Shutdown

```typescript
async function gracefulShutdown(io: Server) {
  console.log('Starting graceful shutdown...');

  // 1. Stop accepting new connections
  io.close();

  // 2. Notify all clients
  io.emit('server:shutdown', {
    message: 'Server restarting, please reconnect shortly',
    reconnectIn: 5000,
  });

  // 3. Wait for in-flight messages (5 seconds)
  await new Promise(resolve => setTimeout(resolve, 5000));

  // 4. Disconnect remaining clients
  const sockets = await io.fetchSockets();
  for (const socket of sockets) {
    socket.disconnect(true);
  }

  // 5. Close Redis connections
  await pubClient.quit();
  await subClient.quit();

  console.log('Shutdown complete');
  process.exit(0);
}

process.on('SIGTERM', () => gracefulShutdown(io));
process.on('SIGINT', () => gracefulShutdown(io));
```

## Rate Limiting

### Per-Connection Rate Limiter

```typescript
class RateLimiter {
  private buckets = new Map<string, { count: number; resetAt: number }>();

  check(key: string, limit: number, windowMs: number): boolean {
    const now = Date.now();
    const bucket = this.buckets.get(key);

    if (!bucket || now > bucket.resetAt) {
      this.buckets.set(key, { count: 1, resetAt: now + windowMs });
      return true;
    }

    if (bucket.count >= limit) {
      return false;
    }

    bucket.count++;
    return true;
  }
}

const limiter = new RateLimiter();

io.on('connection', (socket) => {
  // Middleware for all events
  socket.use(([event, ...args], next) => {
    const userId = socket.data.user.id;

    // 30 messages per 10 seconds
    if (!limiter.check(`msg:${userId}`, 30, 10000)) {
      return next(new Error('Rate limit exceeded'));
    }

    // Specific limits for expensive operations
    if (event === 'file:upload') {
      if (!limiter.check(`upload:${userId}`, 5, 60000)) {
        return next(new Error('Upload rate limit exceeded'));
      }
    }

    next();
  });
});
```

## Monitoring

### Key Metrics to Track

```typescript
// Prometheus-style metrics
const metrics = {
  connectionsTotal: 0,
  connectionsActive: 0,
  messagesIn: 0,
  messagesOut: 0,
  messageErrors: 0,
  avgLatencyMs: 0,
  roomCount: 0,
};

// Health endpoint
app.get('/health', (req, res) => {
  res.json({
    status: 'healthy',
    connections: metrics.connectionsActive,
    uptime: process.uptime(),
    memory: process.memoryUsage(),
  });
});

// Metrics endpoint (for Prometheus scraping)
app.get('/metrics', (req, res) => {
  res.set('Content-Type', 'text/plain');
  res.send(`
# HELP ws_connections_active Current active WebSocket connections
# TYPE ws_connections_active gauge
ws_connections_active ${metrics.connectionsActive}

# HELP ws_messages_total Total messages processed
# TYPE ws_messages_total counter
ws_messages_in_total ${metrics.messagesIn}
ws_messages_out_total ${metrics.messagesOut}

# HELP ws_message_errors_total Total message processing errors
# TYPE ws_message_errors_total counter
ws_message_errors_total ${metrics.messageErrors}
  `.trim());
});
```

### Alert Thresholds

| Metric | Warning | Critical |
|--------|---------|----------|
| Connections per server | >8,000 | >12,000 |
| Message latency (p99) | >100ms | >500ms |
| Error rate | >1% | >5% |
| Memory usage | >70% | >85% |
| Redis pub/sub lag | >100 messages | >1,000 messages |
| Reconnection rate | >5%/min | >20%/min |

## Cost Estimation

### Managed Services

| Service | Free Tier | Paid | Per Connection |
|---------|-----------|------|---------------|
| **Pusher** | 200K msg/day, 100 conn | $49/mo | $0.001/msg |
| **Ably** | 6M msg/mo, 200 conn | $29/mo | Custom |
| **Liveblocks** | 300 MAU | $25/mo | Per MAU |
| **Socket.io Cloud** | 1K daily users | $9/mo | Per daily user |
| **PubNub** | 1M transactions | $98/mo | Per transaction |
| **Supabase Realtime** | 200 concurrent | $25/mo | Included |

### Self-Hosted (AWS Estimates)

| Scale | Architecture | Monthly Cost |
|-------|-------------|-------------|
| 1K connections | 1x t3.small + ElastiCache | ~$30 |
| 10K connections | 2x t3.medium + ElastiCache | ~$120 |
| 50K connections | 3x c5.large + ElastiCache cluster | ~$500 |
| 100K connections | 5x c5.xlarge + ElastiCache cluster + ALB | ~$1,500 |
| 500K connections | Auto-scaling group + ElastiCache cluster + ALB | ~$5,000 |

### Resource Planning

```
Memory per WebSocket connection: ~2-10KB
CPU per 1000 messages/sec: ~5% of a modern core
Bandwidth per connection: ~1-5KB/min (idle keepalive)
                          ~10-100KB/min (active chat)
                          ~100KB-1MB/min (collaborative editing)
```

## Checklist Before Scaling

- [ ] Redis adapter configured for Socket.io
- [ ] Sticky sessions enabled on load balancer
- [ ] WebSocket upgrade headers passing through proxy
- [ ] Keepalive/ping configured (under proxy timeout)
- [ ] Graceful shutdown handler implemented
- [ ] Connection count monitoring in place
- [ ] Rate limiting per connection
- [ ] Health check endpoint responding
- [ ] Reconnection logic tested (kill a server, verify clients reconnect)
- [ ] Message ordering verified across multiple servers
- [ ] Room state recovery after server restart
