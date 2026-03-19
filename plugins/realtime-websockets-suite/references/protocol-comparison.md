# Real-Time Protocol Comparison

## Protocol Matrix

| Feature | WebSocket | SSE | HTTP Long Polling | HTTP Polling | WebTransport |
|---------|-----------|-----|-------------------|-------------|--------------|
| Direction | Bidirectional | Server → Client | Server → Client | Client → Server | Bidirectional |
| Protocol | ws:// / wss:// | HTTP/1.1+ | HTTP | HTTP | QUIC/HTTP3 |
| Connection | Persistent | Persistent | Semi-persistent | New per request | Persistent |
| Latency | ~1ms | ~1ms | ~50-500ms | Interval-dependent | <1ms |
| Auto-reconnect | Manual | Built-in | Manual | N/A | Manual |
| Binary data | Yes | No (text only) | Yes | Yes | Yes |
| HTTP/2 multiplexing | No (upgrade) | Yes | Yes | Yes | N/A |
| Browser support | All modern | All modern | All | All | Chrome, Edge |
| Max connections/domain | ~6 (HTTP/1.1) | ~6 (HTTP/1.1) | ~6 | No limit | Many |
| Proxy-friendly | Sometimes | Yes | Yes | Yes | Limited |
| Load balancer support | Needs sticky | Standard | Standard | Standard | Limited |

## Decision Guide

### Use WebSocket When

- **Bidirectional communication required** — chat, gaming, collaborative editing
- **High-frequency updates** — more than 10 messages/second
- **Low latency critical** — real-time gaming, trading platforms
- **Binary data transfer** — file streaming, audio/video signaling
- **Custom protocol needed** — multiplayer games, IoT

### Use SSE When

- **Server-to-client only** — dashboards, feeds, notifications
- **Auto-reconnect matters** — EventSource handles it natively
- **HTTP/2 available** — multiplexing eliminates connection limits
- **Simple implementation** — just HTTP with special headers
- **Behind restrictive proxies** — SSE works through most corporate proxies

### Use Long Polling When

- **WebSocket/SSE blocked** — corporate firewalls, legacy proxies
- **Fallback needed** — Socket.io uses this automatically
- **Low update frequency** — updates every few seconds or less
- **Maximum compatibility** — works everywhere HTTP works

### Use HTTP Polling When

- **Updates are infrequent** — every 30s+ is acceptable
- **Simplicity is priority** — no special server infrastructure
- **Stateless required** — no persistent connections
- **Rate-limited API** — checking external service status

## Protocol Deep Dive

### WebSocket Handshake

```
Client → Server:
GET /chat HTTP/1.1
Host: example.com
Upgrade: websocket
Connection: Upgrade
Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==
Sec-WebSocket-Version: 13

Server → Client:
HTTP/1.1 101 Switching Protocols
Upgrade: websocket
Connection: Upgrade
Sec-WebSocket-Accept: s3pPLMBiTxaQ9kYGzzhZRbK+xOo=
```

After handshake, communication switches from HTTP to the WebSocket protocol on the same TCP connection. Frames are lightweight (2-14 bytes overhead).

### SSE Format

```
event: notification
data: {"type": "message", "from": "alice"}
id: 12345
retry: 5000

: this is a comment (keepalive)

data: simple message without event type
```

- `event:` — named event type (client listens with `addEventListener`)
- `data:` — payload (multiple `data:` lines concatenate with `\n`)
- `id:` — event ID (sent as `Last-Event-ID` header on reconnect)
- `retry:` — reconnection delay in milliseconds
- Lines starting with `:` — comments (used for keepalive)

### WebSocket Frame Structure

```
 0                   1                   2                   3
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-------+-+-------------+-------------------------------+
|F|R|R|R| opcode|M| Payload len |    Extended payload length    |
|I|S|S|S|  (4)  |A|     (7)     |             (16/64)           |
|N|V|V|V|       |S|             |   (if payload len==126/127)   |
| |1|2|3|       |K|             |                               |
+-+-+-+-+-------+-+-------------+-------------------------------+
```

Opcodes: `0x1` text, `0x2` binary, `0x8` close, `0x9` ping, `0xA` pong

## Connection Limits

### HTTP/1.1 Browser Limits

Browsers limit concurrent connections per domain:

| Browser | Max Connections/Domain |
|---------|----------------------|
| Chrome | 6 |
| Firefox | 6 |
| Safari | 6 |
| Edge | 6 |

This affects both WebSocket and SSE. Solutions:
1. **HTTP/2** — multiplexes over a single TCP connection (SSE benefits most)
2. **Subdomains** — `ws1.example.com`, `ws2.example.com`
3. **Single multiplexed connection** — one WebSocket carrying multiple channels

### HTTP/2 and SSE

HTTP/2 multiplexes streams over a single connection. SSE benefits greatly:
- No 6-connection limit per domain
- Lower overhead than WebSocket (no protocol upgrade)
- Standard HTTP caching and compression

## Performance Benchmarks (Typical)

| Metric | WebSocket | SSE | Long Polling |
|--------|-----------|-----|-------------|
| Connection setup | ~100ms | ~50ms | ~50ms |
| Message latency | 1-5ms | 1-5ms | 50-500ms |
| Overhead per message | 2-14 bytes | ~50 bytes | ~500 bytes |
| Memory per connection | ~2KB | ~4KB | ~8KB |
| Messages/sec/connection | 10,000+ | 1,000+ | ~2-10 |
| Connections per server | 50K-1M | 50K-500K | 10K-50K |

## Scaling Characteristics

### WebSocket Scaling

```
Challenge: Each connection is persistent → high memory per server
Solution:  Horizontal scaling with sticky sessions + Redis pub/sub

Architecture:
  Client → Load Balancer (sticky sessions)
         → Server 1 ←→ Redis Pub/Sub ←→ Server 2
                                      ←→ Server 3
```

**Sticky sessions required**: WebSocket upgrade happens on a specific server. The client must always route to the same server.

### SSE Scaling

```
Challenge: Each connection holds an HTTP response open
Solution:  HTTP/2 multiplexing + Redis pub/sub

Architecture:
  Client → Load Balancer (no sticky needed with HTTP/2)
         → Server 1 ←→ Redis Pub/Sub ←→ Server 2
```

SSE doesn't need sticky sessions because each reconnect is a new HTTP request. The `Last-Event-ID` header lets servers resume from the right position.

## Security Considerations

| Concern | WebSocket | SSE |
|---------|-----------|-----|
| Authentication | Token in handshake or first message | Token in URL or cookie |
| CORS | Not subject to same-origin (use Origin header) | Standard CORS headers |
| Encryption | wss:// (TLS) | https:// (TLS) |
| Message validation | Must validate every frame | Must validate every event |
| DoS protection | Rate limit per connection | Standard HTTP rate limiting |
| XSS impact | Persistent connection = persistent attack | Auto-reconnect = persistent |

### Authentication Patterns

```typescript
// WebSocket — token in handshake
const socket = io(url, { auth: { token: jwt } });

// WebSocket — token in first message (less common)
ws.onopen = () => ws.send(JSON.stringify({ type: 'auth', token: jwt }));

// SSE — token in URL (avoid if possible, shows in logs)
new EventSource(`/events?token=${jwt}`);

// SSE — cookie-based (preferred)
// Set HttpOnly cookie on login, browser sends automatically
new EventSource('/events', { withCredentials: true });
```

## Migration Paths

### Polling → SSE

1. Keep existing REST endpoints for initial data load
2. Add SSE endpoint for incremental updates
3. Client subscribes to SSE after initial fetch
4. Remove polling interval

### SSE → WebSocket

1. Replace EventSource with Socket.io client
2. Replace SSE endpoint with Socket.io server
3. Add bidirectional events (no more separate POST requests)
4. Implement reconnection logic (Socket.io handles this)

### Socket.io → Raw WebSocket

1. Replace Socket.io server with `ws` library
2. Implement custom protocol (event names, acknowledgments)
3. Implement reconnection, heartbeat, and room management manually
4. Only do this if you need: lower overhead, custom binary protocol, or fewer dependencies
