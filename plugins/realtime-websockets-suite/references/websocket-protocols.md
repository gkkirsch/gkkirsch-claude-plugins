# WebSocket Protocol Reference

Quick reference for WebSocket protocols, transports, and connection lifecycle.

---

## WebSocket Protocol (RFC 6455)

### Connection Lifecycle

```
Client                                Server
  |                                     |
  |  HTTP GET /ws (Upgrade request)     |
  |  Connection: Upgrade               |
  |  Upgrade: websocket                |
  |  Sec-WebSocket-Key: dGhlIHNh...    |
  |  Sec-WebSocket-Version: 13         |
  | ---------------------------------→ |
  |                                     |
  |  HTTP 101 Switching Protocols      |
  |  Connection: Upgrade               |
  |  Upgrade: websocket                |
  |  Sec-WebSocket-Accept: s3pPLM...   |
  | ←--------------------------------- |
  |                                     |
  |  ← WebSocket frames (bidirectional) → |
  |                                     |
  |  Close frame (opcode 0x8)          |
  | ←→ Close handshake                 |
```

### Frame Opcodes

| Opcode | Type | Description |
|--------|------|-------------|
| `0x0` | Continuation | Fragment continuation |
| `0x1` | Text | UTF-8 text data |
| `0x2` | Binary | Binary data |
| `0x8` | Close | Connection close |
| `0x9` | Ping | Heartbeat request |
| `0xA` | Pong | Heartbeat response |

### Close Codes

| Code | Name | Meaning |
|------|------|---------|
| `1000` | Normal | Clean close |
| `1001` | Going Away | Server shutting down or client navigating away |
| `1002` | Protocol Error | Protocol violation |
| `1003` | Unsupported | Unexpected data type |
| `1006` | Abnormal | No close frame received (connection dropped) |
| `1008` | Policy Violation | Message violates policy |
| `1009` | Too Large | Message too big |
| `1011` | Internal Error | Server error |
| `1012` | Service Restart | Server restarting |
| `1013` | Try Again | Temporary condition, retry |
| `4000-4999` | Application | Reserved for application use |

### URL Schemes

| Scheme | Port | Encryption |
|--------|------|------------|
| `ws://` | 80 | None |
| `wss://` | 443 | TLS (always use in production) |

---

## Transport Comparison

### WebSocket vs SSE vs HTTP Polling

| Feature | WebSocket | SSE | Long Polling | Short Polling |
|---------|-----------|-----|-------------|---------------|
| Direction | Bidirectional | Server → Client | Server → Client | Server → Client |
| Connection | Persistent | Persistent | Per-request (held) | Per-request |
| Protocol | ws:// / wss:// | HTTP | HTTP | HTTP |
| Auto-reconnect | Manual | Built-in | Manual | N/A (always new) |
| Binary support | Yes | No (text only) | Yes | Yes |
| HTTP/2 multiplexing | No (separate TCP) | Yes | Yes | Yes |
| Browser support | All modern | All modern | All | All |
| Through firewalls | Sometimes blocked | Always works | Always works | Always works |
| Max connections | ~6 per domain | ~6 per domain (HTTP/1.1) | ~6 per domain | N/A |
| Latency | Lowest (~ms) | Low (~ms) | Medium (~100ms) | High (poll interval) |
| Server resources | 1 conn per client | 1 conn per client | High (frequent requests) | Medium |

### When to Use Each

| Use Case | Best Transport | Why |
|----------|---------------|-----|
| Chat, gaming | WebSocket | Bidirectional, low latency |
| Live dashboard, feeds | SSE | One-way, auto-reconnect, simpler |
| Notifications | SSE | One-way push, auto-reconnect |
| Collaborative editing | WebSocket | Bidirectional, binary CRDT sync |
| File upload progress | SSE | Server → client progress updates |
| Real-time search | WebSocket | Query + results both directions |
| IoT sensor data | WebSocket | High-frequency bidirectional |
| Stock ticker | SSE or WebSocket | One-way is fine, but WebSocket if client filters |
| Fallback / corporate | Long Polling | Works through restrictive proxies |

---

## Socket.io Protocol

### Transport Upgrade Flow

```
1. Client connects via HTTP long-polling (transport: "polling")
2. Server responds with session ID (sid) and available upgrades
3. Client sends WebSocket upgrade probe
4. Server confirms probe
5. Client switches to WebSocket transport
6. Polling transport closed
```

This is why Socket.io works through restrictive firewalls — it starts with HTTP and upgrades.

### Packet Types

| Type | ID | Description |
|------|-----|-------------|
| CONNECT | `0` | Connect to namespace |
| DISCONNECT | `1` | Disconnect from namespace |
| EVENT | `2` | Event with data |
| ACK | `3` | Event acknowledgment |
| CONNECT_ERROR | `4` | Connection error |
| BINARY_EVENT | `5` | Binary event data |
| BINARY_ACK | `6` | Binary acknowledgment |

### Engine.io Packet Types (Transport Layer)

| Type | ID | Description |
|------|-----|-------------|
| open | `0` | Session handshake |
| close | `1` | Close session |
| ping | `2` | Heartbeat request |
| pong | `3` | Heartbeat response |
| message | `4` | Actual data (Socket.io packets) |
| upgrade | `5` | Transport upgrade |
| noop | `6` | No-op (used in upgrades) |

### Socket.io vs Raw WebSocket

| Feature | Socket.io | Raw WebSocket (ws) |
|---------|-----------|-------------------|
| Reconnection | Built-in | Manual |
| Rooms/namespaces | Built-in | Manual |
| Broadcasting | Built-in | Manual |
| Fallback transports | Polling → WebSocket | WebSocket only |
| Binary support | Auto-detected | Manual |
| Acknowledgments | Built-in callbacks | Manual |
| Middleware | Express-style | Manual |
| Scaling (Redis) | @socket.io/redis-adapter | Manual pub/sub |
| Bundle size | ~45KB (client) | 0 (browser native) |
| Overhead per message | ~10-20 bytes | ~2-6 bytes |
| Latency | Slightly higher (protocol) | Lowest possible |

**Use Socket.io** when: You need rooms, reconnection, fallbacks, or rapid development.
**Use raw ws** when: Maximum performance matters, you control both ends, or minimal overhead is critical.

---

## Message Serialization

### JSON (Default)

```typescript
// Most common — simple, debuggable, universal
socket.emit('message', { type: 'chat', text: 'hello', ts: Date.now() });
```

### MessagePack (Binary JSON)

```bash
npm install @msgpack/msgpack socket.io-msgpack-parser
```

```typescript
// Server
import { Server } from 'socket.io';
import { createParser } from 'socket.io-msgpack-parser';

const io = new Server(httpServer, {
  parser: createParser(),
});

// Client
import { io } from 'socket.io-client';
import { createParser } from 'socket.io-msgpack-parser';

const socket = io(url, {
  parser: createParser(),
});
```

**~30-50% smaller** than JSON for typical payloads. Worth it for high-frequency data.

### Protocol Buffers

```bash
npm install protobufjs
```

```typescript
import protobuf from 'protobufjs';

// Define schema
const root = await protobuf.load('messages.proto');
const ChatMessage = root.lookupType('ChatMessage');

// Encode
const buffer = ChatMessage.encode({ text: 'hello', userId: 123 }).finish();
ws.send(buffer);

// Decode
ws.on('message', (data: Buffer) => {
  const message = ChatMessage.decode(new Uint8Array(data));
});
```

**~80-90% smaller** than JSON. Best for high-throughput binary protocols (gaming, IoT).

---

## Connection Limits

### Browser Limits

| Browser | WebSocket connections per domain | SSE connections per domain (HTTP/1.1) | SSE with HTTP/2 |
|---------|--------------------------------|--------------------------------------|----------------|
| Chrome | 255 | 6 | 100 |
| Firefox | 200 | 6 | 100 |
| Safari | 255 | 6 | ~100 |
| Edge | 255 | 6 | 100 |

**Key insight**: SSE over HTTP/1.1 is limited to 6 connections per domain (shared with other HTTP requests). Use HTTP/2 or WebSocket for multiple real-time connections.

### Server Limits

| Resource | Default Limit | Tuning |
|----------|---------------|--------|
| File descriptors (Linux) | 1024 | `ulimit -n 65535` or `/etc/security/limits.conf` |
| Ephemeral ports | ~28K | `sysctl net.ipv4.ip_local_port_range` |
| Memory per connection | ~10-50KB | Depends on buffers |
| Backlog queue | 128 | `net.core.somaxconn = 65535` |

**Rule of thumb**: A single server process can handle 10K-100K WebSocket connections depending on message frequency and processing. Memory is usually the bottleneck.

### Linux Tuning for High Connection Count

```bash
# /etc/sysctl.conf
net.core.somaxconn = 65535
net.ipv4.ip_local_port_range = 1024 65535
net.ipv4.tcp_tw_reuse = 1
net.core.rmem_max = 16777216
net.core.wmem_max = 16777216

# /etc/security/limits.conf
* soft nofile 65535
* hard nofile 65535
```

---

## Security

### Authentication Patterns

```typescript
// Pattern 1: Token in handshake (Socket.io)
const socket = io(url, {
  auth: { token: 'jwt-token-here' },
});

// Server middleware
io.use((socket, next) => {
  const token = socket.handshake.auth.token;
  try {
    socket.data.user = verifyJWT(token);
    next();
  } catch {
    next(new Error('Authentication failed'));
  }
});

// Pattern 2: Token in URL query (raw WebSocket)
const ws = new WebSocket(`wss://example.com/ws?token=${token}`);
// CAUTION: Token visible in server logs and browser history

// Pattern 3: Cookie-based (best for same-origin)
const ws = new WebSocket('wss://example.com/ws');
// Browser automatically sends cookies
// Server reads session cookie from upgrade request

// Pattern 4: First-message auth
const ws = new WebSocket('wss://example.com/ws');
ws.onopen = () => {
  ws.send(JSON.stringify({ type: 'auth', token: 'jwt-token' }));
};
// Server closes connection if first message isn't valid auth
```

### Rate Limiting

```typescript
// Per-connection rate limiter
const rateLimits = new Map<string, { count: number; resetAt: number }>();

function checkRateLimit(socketId: string, limit = 100, windowMs = 60000): boolean {
  const now = Date.now();
  const entry = rateLimits.get(socketId);

  if (!entry || now > entry.resetAt) {
    rateLimits.set(socketId, { count: 1, resetAt: now + windowMs });
    return true;
  }

  if (entry.count >= limit) return false;
  entry.count++;
  return true;
}

// Usage in message handler
socket.on('message', (data) => {
  if (!checkRateLimit(socket.id)) {
    socket.emit('error', { message: 'Rate limit exceeded' });
    return;
  }
  // Process message
});
```

### Input Validation

```typescript
import { z } from 'zod';

const MessageSchema = z.object({
  type: z.enum(['chat', 'typing', 'read']),
  roomId: z.string().uuid(),
  text: z.string().max(2000).optional(),
  timestamp: z.number().optional(),
});

socket.on('message', (raw) => {
  const result = MessageSchema.safeParse(raw);
  if (!result.success) {
    socket.emit('error', { message: 'Invalid message format' });
    return;
  }
  handleMessage(socket, result.data);
});
```

---

## Testing WebSocket Connections

### wscat (CLI)

```bash
npm install -g wscat

# Connect
wscat -c wss://example.com/ws

# With headers
wscat -c wss://example.com/ws -H "Authorization: Bearer token"

# Send message
> {"type": "ping"}
```

### Browser DevTools

1. Open DevTools → Network tab
2. Filter by "WS"
3. Click the WebSocket connection
4. "Messages" tab shows all frames (sent in green, received in white)

### Postman

Postman supports WebSocket connections natively — create a new WebSocket Request, set the URL, and send/receive messages.

### Artillery (Load Testing)

```bash
npm install -g artillery
```

```yaml
# artillery-ws-test.yml
config:
  target: "wss://example.com"
  phases:
    - duration: 60
      arrivalRate: 10  # 10 new connections per second
  engines:
    ws: {}

scenarios:
  - engine: ws
    flow:
      - send: '{"type": "join", "room": "test"}'
      - think: 1
      - send: '{"type": "message", "text": "hello"}'
      - think: 2
```

```bash
artillery run artillery-ws-test.yml
```
