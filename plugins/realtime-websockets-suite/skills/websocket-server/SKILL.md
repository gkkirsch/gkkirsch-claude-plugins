---
name: websocket-server
description: >
  Set up WebSocket servers with Socket.io or raw ws library. Covers server setup,
  client connection, rooms, namespaces, authentication, reconnection, and scaling
  with Redis adapter. Works with Express and Next.js.
  Triggers: "websocket server", "socket.io setup", "real-time server",
  "websocket connection", "ws server".
  NOT for: SSE-only use cases (use live-updates skill).
version: 1.0.0
argument-hint: "[socket.io|ws|setup|scale]"
allowed-tools: Read, Grep, Glob, Write, Edit, Bash
---

# WebSocket Server

Set up a WebSocket server for real-time communication.

## Option 1: Socket.io (Recommended)

```bash
npm install socket.io socket.io-client
```

### Server Setup (Express)

```typescript
import express from 'express';
import { createServer } from 'http';
import { Server } from 'socket.io';

const app = express();
const httpServer = createServer(app);
const io = new Server(httpServer, {
  cors: {
    origin: process.env.CLIENT_URL || 'http://localhost:3000',
    methods: ['GET', 'POST'],
  },
  pingTimeout: 60000,      // Time before considering connection dead
  pingInterval: 25000,     // How often to ping
  maxHttpBufferSize: 1e6,  // 1MB max message size
});

// Middleware: Authentication
io.use((socket, next) => {
  const token = socket.handshake.auth.token;
  try {
    const user = verifyJWT(token);
    socket.data.user = user;
    next();
  } catch (err) {
    next(new Error('Authentication failed'));
  }
});

// Connection handler
io.on('connection', (socket) => {
  const user = socket.data.user;
  console.log(`User connected: ${user.id}`);

  // Join user's personal room (for targeted messages)
  socket.join(`user:${user.id}`);

  // Handle events
  socket.on('message', (data) => {
    // Validate input
    if (!data.text || typeof data.text !== 'string') return;
    if (data.text.length > 1000) return;

    // Broadcast to room
    io.to(data.roomId).emit('message', {
      id: generateId(),
      userId: user.id,
      userName: user.name,
      text: data.text,
      timestamp: Date.now(),
    });
  });

  // Join a room
  socket.on('join-room', (roomId: string) => {
    socket.join(roomId);
    socket.to(roomId).emit('user-joined', {
      userId: user.id,
      userName: user.name,
    });
  });

  // Leave a room
  socket.on('leave-room', (roomId: string) => {
    socket.leave(roomId);
    socket.to(roomId).emit('user-left', {
      userId: user.id,
    });
  });

  // Typing indicator
  socket.on('typing', (roomId: string) => {
    socket.to(roomId).emit('user-typing', {
      userId: user.id,
      userName: user.name,
    });
  });

  // Disconnect
  socket.on('disconnect', (reason) => {
    console.log(`User disconnected: ${user.id}, reason: ${reason}`);
  });
});

httpServer.listen(3001, () => {
  console.log('WebSocket server running on port 3001');
});
```

### Client Connection (React)

```tsx
import { io, Socket } from 'socket.io-client';
import { useEffect, useRef, useState, useCallback } from 'react';

function useSocket(url: string, token: string) {
  const socketRef = useRef<Socket | null>(null);
  const [connected, setConnected] = useState(false);

  useEffect(() => {
    const socket = io(url, {
      auth: { token },
      reconnection: true,
      reconnectionAttempts: 10,
      reconnectionDelay: 1000,
      reconnectionDelayMax: 10000,
      timeout: 20000,
    });

    socket.on('connect', () => setConnected(true));
    socket.on('disconnect', () => setConnected(false));
    socket.on('connect_error', (err) => {
      console.error('Connection error:', err.message);
    });

    socketRef.current = socket;

    return () => {
      socket.disconnect();
    };
  }, [url, token]);

  const emit = useCallback((event: string, data: any) => {
    socketRef.current?.emit(event, data);
  }, []);

  const on = useCallback((event: string, handler: (...args: any[]) => void) => {
    socketRef.current?.on(event, handler);
    return () => { socketRef.current?.off(event, handler); };
  }, []);

  return { socket: socketRef.current, connected, emit, on };
}

// Usage
function ChatRoom({ roomId }: { roomId: string }) {
  const { connected, emit, on } = useSocket('http://localhost:3001', authToken);
  const [messages, setMessages] = useState<Message[]>([]);

  useEffect(() => {
    emit('join-room', roomId);
    const cleanup = on('message', (msg: Message) => {
      setMessages(prev => [...prev, msg]);
    });
    return () => {
      cleanup();
      emit('leave-room', roomId);
    };
  }, [roomId, emit, on]);

  const sendMessage = (text: string) => {
    emit('message', { roomId, text });
  };

  return (
    <div>
      <div className={`text-sm ${connected ? 'text-green-500' : 'text-red-500'}`}>
        {connected ? 'Connected' : 'Reconnecting...'}
      </div>
      {/* Message list and input */}
    </div>
  );
}
```

### Namespaces

```typescript
// Separate namespaces for different features
const chatNsp = io.of('/chat');
const notifNsp = io.of('/notifications');
const adminNsp = io.of('/admin');

// Each namespace has its own middleware
adminNsp.use((socket, next) => {
  if (socket.data.user?.role !== 'admin') {
    return next(new Error('Admin access required'));
  }
  next();
});

chatNsp.on('connection', (socket) => {
  // Chat-specific handlers
});

notifNsp.on('connection', (socket) => {
  // Notification-specific handlers
});
```

## Option 2: Raw WebSocket (ws library)

```bash
npm install ws
```

```typescript
import { WebSocketServer, WebSocket } from 'ws';
import { createServer } from 'http';

const server = createServer();
const wss = new WebSocketServer({ server });

// Connection handling
wss.on('connection', (ws: WebSocket, req) => {
  // Auth from query string or headers
  const token = new URL(req.url!, 'http://localhost').searchParams.get('token');
  const user = verifyJWT(token);
  if (!user) { ws.close(4001, 'Unauthorized'); return; }

  // Attach user data
  (ws as any).userId = user.id;

  // Message handling
  ws.on('message', (data: Buffer) => {
    try {
      const message = JSON.parse(data.toString());
      handleMessage(ws, user, message);
    } catch (e) {
      ws.send(JSON.stringify({ error: 'Invalid message format' }));
    }
  });

  // Heartbeat
  (ws as any).isAlive = true;
  ws.on('pong', () => { (ws as any).isAlive = true; });

  ws.on('close', () => {
    console.log(`User ${user.id} disconnected`);
  });
});

// Heartbeat interval (detect dead connections)
const heartbeat = setInterval(() => {
  wss.clients.forEach((ws) => {
    if (!(ws as any).isAlive) return ws.terminate();
    (ws as any).isAlive = false;
    ws.ping();
  });
}, 30000);

wss.on('close', () => clearInterval(heartbeat));

// Broadcast to all connected clients
function broadcast(data: any, exclude?: WebSocket) {
  const message = JSON.stringify(data);
  wss.clients.forEach((client) => {
    if (client !== exclude && client.readyState === WebSocket.OPEN) {
      client.send(message);
    }
  });
}

server.listen(3001);
```

## Scaling with Redis Adapter

```bash
npm install @socket.io/redis-adapter redis
```

```typescript
import { createAdapter } from '@socket.io/redis-adapter';
import { createClient } from 'redis';

const pubClient = createClient({ url: process.env.REDIS_URL });
const subClient = pubClient.duplicate();

await pubClient.connect();
await subClient.connect();

io.adapter(createAdapter(pubClient, subClient));

// Now socket.io works across multiple server instances
// Rooms, broadcasts, and targeted messages all work transparently
```

### Sticky Sessions (Required for Scaling)

```nginx
# Nginx config for WebSocket sticky sessions
upstream websocket_servers {
    ip_hash;  # Sticky sessions based on client IP
    server ws1:3001;
    server ws2:3001;
    server ws3:3001;
}

server {
    location /socket.io/ {
        proxy_pass http://websocket_servers;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_read_timeout 86400;  # 24h for long-lived connections
    }
}
```

## Error Handling & Reconnection

```typescript
// Client-side robust connection
const socket = io(url, {
  auth: { token },
  reconnection: true,
  reconnectionAttempts: Infinity,
  reconnectionDelay: 1000,
  reconnectionDelayMax: 30000,
  randomizationFactor: 0.5,
  transports: ['websocket', 'polling'], // WebSocket first, fallback to polling
});

// Handle reconnection
socket.on('reconnect', (attemptNumber) => {
  console.log(`Reconnected after ${attemptNumber} attempts`);
  // Re-join rooms, re-sync state
  socket.emit('rejoin', { rooms: currentRooms });
});

socket.on('reconnect_error', (error) => {
  console.error('Reconnection error:', error);
});

socket.on('reconnect_failed', () => {
  // All attempts exhausted
  showReconnectionFailedUI();
});
```

## Best Practices

1. **Always authenticate** before allowing WebSocket communication
2. **Validate every message** — never trust client input
3. **Set message size limits** — prevent memory exhaustion
4. **Implement heartbeat/ping** — detect dead connections
5. **Handle reconnection** — clients will disconnect (mobile, network changes)
6. **Use rooms** for scoped broadcasting (don't broadcast to everyone)
7. **Rate limit messages** per connection to prevent abuse
8. **Graceful shutdown** — drain connections before killing server
9. **Monitor connection count** — alert on unusual spikes
10. **Use binary for large data** — MessagePack or protobuf for performance
