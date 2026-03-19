---
name: collaborative-editing
description: >
  Build collaborative real-time editing with conflict resolution — shared
  documents, multiplayer cursors, presence awareness, and operational
  transforms or CRDTs. Covers Yjs, Liveblocks, and custom implementations.
  Triggers: "collaborative editing", "real-time editing", "multiplayer editor",
  "shared document", "concurrent editing", "crdt", "operational transform",
  "yjs", "liveblocks", "cursor presence".
  NOT for: chat systems or simple live updates.
version: 1.0.0
argument-hint: "[yjs|liveblocks|custom]"
allowed-tools: Read, Grep, Glob, Write, Edit, Bash
---

# Collaborative Editing

Build real-time collaborative editing with conflict resolution.

## Approach Selection

| Approach | Best For | Complexity | Cost |
|----------|----------|-----------|------|
| **Yjs** | Custom editors, self-hosted | Medium | Free (OSS) |
| **Liveblocks** | Quick setup, managed service | Low | Paid (free tier) |
| **Tiptap Collaboration** | Rich text editors | Medium | Paid (Cloud) |
| **Custom OT/CRDT** | Full control, learning | Very High | Free |

## Yjs — Open Source CRDT Framework

### Setup

```bash
npm install yjs y-websocket y-prosemirror
# or for CodeMirror: y-codemirror.next
# or for Monaco: y-monaco
```

### Server (y-websocket)

```typescript
// server/collaboration.ts
import { WebSocketServer } from 'ws';
import { setupWSConnection } from 'y-websocket/bin/utils';

const wss = new WebSocketServer({ port: 1234 });

wss.on('connection', (ws, req) => {
  // Auth check
  const token = new URL(req.url!, `http://localhost`).searchParams.get('token');
  if (!verifyToken(token)) {
    ws.close(4001, 'Unauthorized');
    return;
  }

  setupWSConnection(ws, req);
});

console.log('Yjs WebSocket server running on :1234');
```

### With Persistence (LevelDB)

```bash
npm install y-leveldb
```

```typescript
// server/collaboration-persistent.ts
import { LeveldbPersistence } from 'y-leveldb';

const persistence = new LeveldbPersistence('./yjs-data');

// In your y-websocket setup:
const HOST = process.env.HOST || 'localhost';
const PORT = parseInt(process.env.PORT || '1234');

// Start with persistence enabled
// y-websocket reads YPERSISTENCE env var
process.env.YPERSISTENCE = './yjs-data';
```

### Client — Tiptap + Yjs

```bash
npm install @tiptap/core @tiptap/starter-kit @tiptap/extension-collaboration \
  @tiptap/extension-collaboration-cursor yjs y-websocket
```

```tsx
// components/CollaborativeEditor.tsx
'use client';
import { useEditor, EditorContent } from '@tiptap/react';
import StarterKit from '@tiptap/starter-kit';
import Collaboration from '@tiptap/extension-collaboration';
import CollaborationCursor from '@tiptap/extension-collaboration-cursor';
import * as Y from 'yjs';
import { WebsocketProvider } from 'y-websocket';
import { useEffect, useMemo } from 'react';

interface Props {
  documentId: string;
  user: { name: string; color: string };
}

export function CollaborativeEditor({ documentId, user }: Props) {
  // Create Yjs document and provider
  const { ydoc, provider } = useMemo(() => {
    const ydoc = new Y.Doc();
    const provider = new WebsocketProvider(
      process.env.NEXT_PUBLIC_YJS_WS_URL!,
      documentId,
      ydoc,
      { params: { token: localStorage.getItem('token') || '' } }
    );
    return { ydoc, provider };
  }, [documentId]);

  // Cleanup
  useEffect(() => {
    return () => {
      provider.destroy();
      ydoc.destroy();
    };
  }, [ydoc, provider]);

  const editor = useEditor({
    extensions: [
      StarterKit.configure({ history: false }), // Disable default history (Yjs handles it)
      Collaboration.configure({ document: ydoc }),
      CollaborationCursor.configure({
        provider,
        user: { name: user.name, color: user.color },
      }),
    ],
  });

  return (
    <div className="border rounded-lg">
      {/* Connection status */}
      <div className="flex items-center gap-2 px-4 py-2 border-b bg-gray-50 text-sm">
        <span className={`w-2 h-2 rounded-full ${
          provider.wsconnected ? 'bg-green-500' : 'bg-red-500'
        }`} />
        <span className="text-gray-500">
          {provider.wsconnected ? 'Connected' : 'Reconnecting...'}
        </span>

        {/* Online users */}
        <div className="ml-auto flex -space-x-2">
          {/* Render awareness states as user avatars */}
        </div>
      </div>

      {/* Editor */}
      <EditorContent editor={editor} className="prose max-w-none p-4 min-h-[300px]" />
    </div>
  );
}
```

### Awareness (Presence & Cursors)

```typescript
// Awareness shows who's online and where their cursor is
import { Awareness } from 'y-protocols/awareness';

// Set local user state
provider.awareness.setLocalStateField('user', {
  name: 'Alice',
  color: '#ff0000',
  cursor: null,
});

// Listen for remote user changes
provider.awareness.on('change', () => {
  const states = provider.awareness.getStates();
  states.forEach((state, clientId) => {
    console.log(`User ${state.user?.name} is online`);
  });
});
```

## Liveblocks — Managed Service

### Setup

```bash
npm install @liveblocks/client @liveblocks/react @liveblocks/yjs
```

```typescript
// liveblocks.config.ts
import { createClient } from '@liveblocks/client';
import { createRoomContext } from '@liveblocks/react';

const client = createClient({
  publicApiKey: process.env.NEXT_PUBLIC_LIVEBLOCKS_KEY!,
  // or authEndpoint for production:
  // authEndpoint: '/api/liveblocks-auth',
});

type Presence = {
  cursor: { x: number; y: number } | null;
  name: string;
  color: string;
};

type Storage = {
  content: any; // Yjs document
};

export const {
  RoomProvider,
  useOthers,
  useMyPresence,
  useSelf,
  useStorage,
} = createRoomContext<Presence, Storage>(client);
```

### Liveblocks + Tiptap

```tsx
// components/LiveblocksEditor.tsx
'use client';
import { useEditor, EditorContent } from '@tiptap/react';
import StarterKit from '@tiptap/starter-kit';
import Collaboration from '@tiptap/extension-collaboration';
import CollaborationCursor from '@tiptap/extension-collaboration-cursor';
import LiveblocksProvider from '@liveblocks/yjs';
import * as Y from 'yjs';
import { RoomProvider, useRoom } from '../liveblocks.config';

function Editor() {
  const room = useRoom();

  const { ydoc, provider } = useMemo(() => {
    const ydoc = new Y.Doc();
    const provider = new LiveblocksProvider(room, ydoc);
    return { ydoc, provider };
  }, [room]);

  const editor = useEditor({
    extensions: [
      StarterKit.configure({ history: false }),
      Collaboration.configure({ document: ydoc }),
      CollaborationCursor.configure({
        provider,
        user: { name: 'Alice', color: '#ff0000' },
      }),
    ],
  });

  return <EditorContent editor={editor} />;
}

// Wrap with RoomProvider
export function LiveblocksEditor({ documentId }: { documentId: string }) {
  return (
    <RoomProvider id={documentId} initialPresence={{ cursor: null, name: '', color: '' }}>
      <Editor />
    </RoomProvider>
  );
}
```

### Cursor Presence (Non-Editor)

```tsx
// Shared canvas, whiteboard, or any spatial UI
function CollaborativeCanvas() {
  const others = useOthers();
  const [myPresence, updateMyPresence] = useMyPresence();

  return (
    <div
      className="relative w-full h-[600px] bg-gray-50"
      onPointerMove={(e) => {
        const rect = e.currentTarget.getBoundingClientRect();
        updateMyPresence({
          cursor: { x: e.clientX - rect.left, y: e.clientY - rect.top },
        });
      }}
      onPointerLeave={() => updateMyPresence({ cursor: null })}
    >
      {/* Other users' cursors */}
      {others.map(({ connectionId, presence }) => {
        if (!presence.cursor) return null;
        return (
          <div
            key={connectionId}
            className="absolute pointer-events-none"
            style={{
              left: presence.cursor.x,
              top: presence.cursor.y,
              transform: 'translate(-4px, -4px)',
            }}
          >
            <svg width="24" height="24" viewBox="0 0 24 24" fill={presence.color}>
              <path d="M5 3l14 8-8 3-3 8z" />
            </svg>
            <span className="absolute left-5 top-4 text-xs bg-black/70 text-white px-1.5 py-0.5 rounded whitespace-nowrap">
              {presence.name}
            </span>
          </div>
        );
      })}
    </div>
  );
}
```

## Conflict Resolution Concepts

### CRDT (Conflict-Free Replicated Data Types)

CRDTs guarantee that concurrent edits will converge to the same result without coordination:

```
User A types "Hello" at position 0
User B types "World" at position 0
           ↓
CRDT resolves to "HelloWorld" or "WorldHello" (deterministic)
Both users see the same result — no conflicts, no data loss
```

**Used by**: Yjs, Automerge, Liveblocks

### Operational Transform (OT)

OT transforms operations against each other to resolve conflicts:

```
User A: insert "x" at position 5
User B: delete character at position 3
           ↓
OT transforms A's operation: insert "x" at position 4 (shifted by B's delete)
```

**Used by**: Google Docs, Firepad

### When to Use Which

| Factor | CRDT | OT |
|--------|------|-----|
| Offline support | Excellent | Poor |
| Peer-to-peer | Yes | No (needs server) |
| Complexity | Medium | High |
| Memory usage | Higher | Lower |
| Maturity | Growing | Proven |
| Best for | New projects | Existing infrastructure |

## Database Integration

```typescript
// Save Yjs document state to database periodically
import * as Y from 'yjs';

async function saveDocument(documentId: string, ydoc: Y.Doc) {
  const state = Y.encodeStateAsUpdate(ydoc);
  const stateVector = Y.encodeStateVector(ydoc);

  await prisma.document.upsert({
    where: { id: documentId },
    create: {
      id: documentId,
      state: Buffer.from(state),
      stateVector: Buffer.from(stateVector),
    },
    update: {
      state: Buffer.from(state),
      stateVector: Buffer.from(stateVector),
      updatedAt: new Date(),
    },
  });
}

// Load document from database
async function loadDocument(documentId: string): Promise<Y.Doc> {
  const ydoc = new Y.Doc();

  const saved = await prisma.document.findUnique({
    where: { id: documentId },
  });

  if (saved?.state) {
    Y.applyUpdate(ydoc, new Uint8Array(saved.state));
  }

  return ydoc;
}
```

## Best Practices

1. **Use Yjs for self-hosted, Liveblocks for managed** — don't build your own CRDT
2. **Disable editor history** — the CRDT handles undo/redo with full awareness
3. **Debounce persistence** — save document state every 2-5 seconds, not every keystroke
4. **Show connection status** — users need to know if they're online/synced
5. **Handle reconnection** — Yjs syncs missed changes automatically, but show loading state
6. **Limit document size** — large documents degrade performance. Split into sections.
7. **Permission levels** — read-only viewers shouldn't generate cursor presence
8. **Version history** — store periodic snapshots for time-travel and recovery
