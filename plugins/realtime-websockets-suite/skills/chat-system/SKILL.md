---
name: chat-system
description: >
  Build real-time chat systems with WebSocket — direct messages, group chat,
  typing indicators, read receipts, message history, file sharing, and
  presence detection. Covers Socket.io and native WebSocket implementations.
  Triggers: "chat system", "messaging app", "direct messages", "group chat",
  "chat feature", "real-time messaging", "typing indicator".
  NOT for: chatbots/AI chat (use AI integration), or email/notifications.
version: 1.0.0
argument-hint: "[dm|group|typing|presence]"
allowed-tools: Read, Grep, Glob, Write, Edit, Bash
---

# Chat System

Build real-time chat with WebSocket.

## Database Schema

```prisma
model Conversation {
  id        String   @id @default(cuid())
  type      ConversationType @default(DIRECT)
  name      String?  // for group chats
  avatar    String?
  createdAt DateTime @default(now())
  updatedAt DateTime @updatedAt
  members   ConversationMember[]
  messages  Message[]

  @@index([updatedAt])
}

model ConversationMember {
  id             String       @id @default(cuid())
  conversationId String
  conversation   Conversation @relation(fields: [conversationId], references: [id], onDelete: Cascade)
  userId         String
  user           User         @relation(fields: [userId], references: [id])
  role           MemberRole   @default(MEMBER)
  lastReadAt     DateTime?
  joinedAt       DateTime     @default(now())
  muted          Boolean      @default(false)

  @@unique([conversationId, userId])
  @@index([userId])
}

model Message {
  id             String       @id @default(cuid())
  conversationId String
  conversation   Conversation @relation(fields: [conversationId], references: [id], onDelete: Cascade)
  senderId       String
  sender         User         @relation(fields: [senderId], references: [id])
  content        String
  type           MessageType  @default(TEXT)
  replyToId      String?
  replyTo        Message?     @relation("MessageReplies", fields: [replyToId], references: [id])
  replies        Message[]    @relation("MessageReplies")
  attachments    Json?        // [{url, name, size, type}]
  editedAt       DateTime?
  deletedAt      DateTime?
  createdAt      DateTime     @default(now())

  @@index([conversationId, createdAt])
  @@index([senderId])
}

enum ConversationType {
  DIRECT
  GROUP
}

enum MemberRole {
  ADMIN
  MEMBER
}

enum MessageType {
  TEXT
  IMAGE
  FILE
  SYSTEM
}
```

## Server Implementation (Socket.io)

```typescript
// server/chat.ts
import { Server } from 'socket.io';
import { Server as HTTPServer } from 'http';

export function setupChat(httpServer: HTTPServer) {
  const io = new Server(httpServer, {
    cors: { origin: process.env.CLIENT_URL, credentials: true },
    pingInterval: 25000,
    pingTimeout: 60000,
  });

  // Auth middleware
  io.use(async (socket, next) => {
    const token = socket.handshake.auth.token;
    try {
      const user = await verifyToken(token);
      socket.data.userId = user.id;
      socket.data.user = user;
      next();
    } catch (err) {
      next(new Error('Authentication failed'));
    }
  });

  // Track online users
  const onlineUsers = new Map<string, Set<string>>(); // userId -> Set<socketId>

  io.on('connection', async (socket) => {
    const userId = socket.data.userId;

    // Track user online status
    if (!onlineUsers.has(userId)) {
      onlineUsers.set(userId, new Set());
    }
    onlineUsers.get(userId)!.add(socket.id);

    // Join all conversation rooms
    const memberships = await prisma.conversationMember.findMany({
      where: { userId },
      select: { conversationId: true },
    });
    for (const m of memberships) {
      socket.join(`conv:${m.conversationId}`);
    }

    // Broadcast online status
    socket.broadcast.emit('user:online', { userId });

    // --- Send Message ---
    socket.on('message:send', async (data, callback) => {
      try {
        const { conversationId, content, type = 'TEXT', replyToId, attachments } = data;

        // Verify membership
        const member = await prisma.conversationMember.findUnique({
          where: { conversationId_userId: { conversationId, userId } },
        });
        if (!member) return callback({ error: 'Not a member' });

        // Create message
        const message = await prisma.message.create({
          data: { conversationId, senderId: userId, content, type, replyToId, attachments },
          include: {
            sender: { select: { id: true, name: true, avatar: true } },
            replyTo: { select: { id: true, content: true, senderId: true } },
          },
        });

        // Update conversation timestamp
        await prisma.conversation.update({
          where: { id: conversationId },
          data: { updatedAt: new Date() },
        });

        // Broadcast to room
        io.to(`conv:${conversationId}`).emit('message:new', message);

        callback({ success: true, message });
      } catch (error) {
        callback({ error: 'Failed to send message' });
      }
    });

    // --- Typing Indicator ---
    socket.on('typing:start', (data) => {
      socket.to(`conv:${data.conversationId}`).emit('typing:update', {
        conversationId: data.conversationId,
        userId,
        isTyping: true,
      });
    });

    socket.on('typing:stop', (data) => {
      socket.to(`conv:${data.conversationId}`).emit('typing:update', {
        conversationId: data.conversationId,
        userId,
        isTyping: false,
      });
    });

    // --- Read Receipts ---
    socket.on('message:read', async (data) => {
      const { conversationId } = data;

      await prisma.conversationMember.update({
        where: { conversationId_userId: { conversationId, userId } },
        data: { lastReadAt: new Date() },
      });

      socket.to(`conv:${conversationId}`).emit('message:read', {
        conversationId,
        userId,
        readAt: new Date(),
      });
    });

    // --- Edit Message ---
    socket.on('message:edit', async (data, callback) => {
      const { messageId, content } = data;

      const message = await prisma.message.findUnique({ where: { id: messageId } });
      if (!message || message.senderId !== userId) {
        return callback({ error: 'Cannot edit this message' });
      }

      const updated = await prisma.message.update({
        where: { id: messageId },
        data: { content, editedAt: new Date() },
      });

      io.to(`conv:${message.conversationId}`).emit('message:edited', updated);
      callback({ success: true });
    });

    // --- Delete Message ---
    socket.on('message:delete', async (data, callback) => {
      const { messageId } = data;

      const message = await prisma.message.findUnique({ where: { id: messageId } });
      if (!message || message.senderId !== userId) {
        return callback({ error: 'Cannot delete this message' });
      }

      await prisma.message.update({
        where: { id: messageId },
        data: { deletedAt: new Date(), content: '' },
      });

      io.to(`conv:${message.conversationId}`).emit('message:deleted', {
        messageId,
        conversationId: message.conversationId,
      });
      callback({ success: true });
    });

    // --- Disconnect ---
    socket.on('disconnect', () => {
      const userSockets = onlineUsers.get(userId);
      if (userSockets) {
        userSockets.delete(socket.id);
        if (userSockets.size === 0) {
          onlineUsers.delete(userId);
          socket.broadcast.emit('user:offline', { userId });
        }
      }
    });
  });

  return io;
}
```

## REST API Endpoints

```typescript
// GET /api/conversations
app.get('/api/conversations', async (req, res) => {
  const conversations = await prisma.conversation.findMany({
    where: {
      members: { some: { userId: req.user.id } },
    },
    include: {
      members: {
        include: { user: { select: { id: true, name: true, avatar: true } } },
      },
      messages: {
        orderBy: { createdAt: 'desc' },
        take: 1,
        include: { sender: { select: { name: true } } },
      },
    },
    orderBy: { updatedAt: 'desc' },
  });

  // Add unread counts
  const result = await Promise.all(
    conversations.map(async (conv) => {
      const membership = conv.members.find(m => m.userId === req.user.id);
      const unreadCount = await prisma.message.count({
        where: {
          conversationId: conv.id,
          createdAt: { gt: membership?.lastReadAt || new Date(0) },
          senderId: { not: req.user.id },
        },
      });

      return {
        ...conv,
        unreadCount,
        lastMessage: conv.messages[0] || null,
      };
    })
  );

  res.json(result);
});

// GET /api/conversations/:id/messages
app.get('/api/conversations/:id/messages', async (req, res) => {
  const cursor = req.query.cursor as string | undefined;
  const limit = 50;

  const messages = await prisma.message.findMany({
    where: {
      conversationId: req.params.id,
      deletedAt: null,
    },
    include: {
      sender: { select: { id: true, name: true, avatar: true } },
      replyTo: { select: { id: true, content: true, senderId: true } },
    },
    orderBy: { createdAt: 'desc' },
    take: limit + 1,
    ...(cursor && { cursor: { id: cursor }, skip: 1 }),
  });

  const hasMore = messages.length > limit;
  if (hasMore) messages.pop();

  res.json({
    messages: messages.reverse(), // chronological order
    hasMore,
    nextCursor: hasMore ? messages[0].id : null,
  });
});

// POST /api/conversations (create DM or group)
app.post('/api/conversations', async (req, res) => {
  const { type = 'DIRECT', memberIds, name } = req.body;

  // For DMs, check if conversation already exists
  if (type === 'DIRECT' && memberIds.length === 1) {
    const existing = await prisma.conversation.findFirst({
      where: {
        type: 'DIRECT',
        AND: [
          { members: { some: { userId: req.user.id } } },
          { members: { some: { userId: memberIds[0] } } },
        ],
      },
    });
    if (existing) return res.json(existing);
  }

  const conversation = await prisma.conversation.create({
    data: {
      type,
      name: type === 'GROUP' ? name : null,
      members: {
        create: [
          { userId: req.user.id, role: type === 'GROUP' ? 'ADMIN' : 'MEMBER' },
          ...memberIds.map((id: string) => ({ userId: id })),
        ],
      },
    },
    include: {
      members: { include: { user: { select: { id: true, name: true, avatar: true } } } },
    },
  });

  res.json(conversation);
});
```

## React Chat Component

```tsx
// components/Chat.tsx
'use client';
import { useState, useEffect, useRef, useCallback } from 'react';
import { io, Socket } from 'socket.io-client';

interface Message {
  id: string;
  content: string;
  senderId: string;
  sender: { id: string; name: string; avatar?: string };
  createdAt: string;
  editedAt?: string;
  replyTo?: { id: string; content: string };
}

export function Chat({ conversationId, currentUser }: {
  conversationId: string;
  currentUser: { id: string; name: string };
}) {
  const [messages, setMessages] = useState<Message[]>([]);
  const [input, setInput] = useState('');
  const [typing, setTyping] = useState<string[]>([]);
  const [socket, setSocket] = useState<Socket | null>(null);
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const typingTimeoutRef = useRef<NodeJS.Timeout>();

  // Connect socket
  useEffect(() => {
    const s = io(process.env.NEXT_PUBLIC_WS_URL!, {
      auth: { token: localStorage.getItem('token') },
    });

    s.on('message:new', (message: Message) => {
      if (message.conversationId === conversationId) {
        setMessages(prev => [...prev, message]);
      }
    });

    s.on('typing:update', ({ userId, isTyping }) => {
      setTyping(prev =>
        isTyping ? [...new Set([...prev, userId])] : prev.filter(id => id !== userId)
      );
    });

    s.on('message:edited', (updated) => {
      setMessages(prev => prev.map(m => m.id === updated.id ? { ...m, ...updated } : m));
    });

    s.on('message:deleted', ({ messageId }) => {
      setMessages(prev => prev.filter(m => m.id !== messageId));
    });

    setSocket(s);
    return () => { s.disconnect(); };
  }, [conversationId]);

  // Load history
  useEffect(() => {
    fetch(`/api/conversations/${conversationId}/messages`)
      .then(r => r.json())
      .then(data => setMessages(data.messages));
  }, [conversationId]);

  // Auto-scroll
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  // Send message
  const sendMessage = useCallback(() => {
    if (!input.trim() || !socket) return;

    socket.emit('message:send', {
      conversationId,
      content: input.trim(),
    }, (response: any) => {
      if (response.error) console.error(response.error);
    });

    setInput('');
    socket.emit('typing:stop', { conversationId });
  }, [input, socket, conversationId]);

  // Typing indicator
  const handleInputChange = (value: string) => {
    setInput(value);
    if (!socket) return;

    socket.emit('typing:start', { conversationId });
    clearTimeout(typingTimeoutRef.current);
    typingTimeoutRef.current = setTimeout(() => {
      socket.emit('typing:stop', { conversationId });
    }, 2000);
  };

  // Mark as read
  useEffect(() => {
    if (socket && messages.length > 0) {
      socket.emit('message:read', { conversationId });
    }
  }, [messages.length, socket, conversationId]);

  return (
    <div className="flex flex-col h-full">
      {/* Messages */}
      <div className="flex-1 overflow-y-auto p-4 space-y-3">
        {messages.map(msg => (
          <div key={msg.id}
            className={`flex ${msg.senderId === currentUser.id ? 'justify-end' : 'justify-start'}`}>
            <div className={`max-w-[70%] rounded-2xl px-4 py-2 ${
              msg.senderId === currentUser.id
                ? 'bg-blue-600 text-white'
                : 'bg-gray-100 text-gray-900'
            }`}>
              {msg.senderId !== currentUser.id && (
                <p className="text-xs font-semibold mb-1 opacity-70">{msg.sender.name}</p>
              )}
              <p className="text-sm">{msg.content}</p>
              <p className="text-xs opacity-50 mt-1">
                {new Date(msg.createdAt).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                {msg.editedAt && ' (edited)'}
              </p>
            </div>
          </div>
        ))}
        <div ref={messagesEndRef} />
      </div>

      {/* Typing indicator */}
      {typing.length > 0 && (
        <div className="px-4 py-1 text-xs text-gray-500">
          {typing.length === 1 ? 'Someone is typing...' : `${typing.length} people are typing...`}
        </div>
      )}

      {/* Input */}
      <div className="border-t p-4">
        <div className="flex gap-2">
          <input
            type="text"
            value={input}
            onChange={(e) => handleInputChange(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && !e.shiftKey && sendMessage()}
            placeholder="Type a message..."
            className="flex-1 border rounded-full px-4 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
          <button onClick={sendMessage}
            className="bg-blue-600 text-white rounded-full px-6 py-2 font-medium hover:bg-blue-700">
            Send
          </button>
        </div>
      </div>
    </div>
  );
}
```

## File Upload in Chat

```typescript
// POST /api/messages/upload
app.post('/api/messages/upload', upload.single('file'), async (req, res) => {
  if (!req.file) return res.status(400).json({ error: 'No file' });

  const maxSize = 10 * 1024 * 1024; // 10MB
  if (req.file.size > maxSize) {
    return res.status(400).json({ error: 'File too large (max 10MB)' });
  }

  // Upload to storage (S3, Cloudinary, etc.)
  const url = await uploadToStorage(req.file);

  res.json({
    url,
    name: req.file.originalname,
    size: req.file.size,
    type: req.file.mimetype,
  });
});
```

## Best Practices

1. **Paginate message history** — load 50 messages at a time, older on scroll
2. **Optimistic updates** — show sent message immediately, mark as pending
3. **Reconnection handling** — Socket.io handles this, but fetch missed messages on reconnect
4. **Typing debounce** — 2-second timeout after last keystroke to stop typing indicator
5. **Unread counts** — track per-conversation with lastReadAt timestamps
6. **Message deduplication** — use message IDs to prevent duplicates on reconnect
7. **Rate limiting** — max 5 messages per second per user
8. **Content moderation** — filter profanity, detect spam patterns
