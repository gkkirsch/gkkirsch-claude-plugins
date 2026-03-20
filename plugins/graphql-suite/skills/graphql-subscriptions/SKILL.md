---
name: graphql-subscriptions
description: >
  GraphQL subscriptions and real-time data patterns with WebSockets.
  Use when implementing real-time updates, live queries, subscription resolvers,
  or WebSocket transport for GraphQL APIs.
  Triggers: "graphql subscriptions", "graphql websocket", "graphql real-time",
  "graphql live queries", "subscription resolver", "graphql-ws", "PubSub graphql".
  NOT for: REST webhooks, Server-Sent Events, Socket.io without GraphQL, Apollo Client setup (see apollo-development).
version: 1.0.0
allowed-tools: Read, Grep, Glob, Bash
---

# GraphQL Subscriptions

## Server Setup with graphql-ws

```typescript
// server.ts — Apollo Server with WebSocket subscriptions
import { ApolloServer } from '@apollo/server';
import { expressMiddleware } from '@apollo/server/express4';
import { createServer } from 'http';
import express from 'express';
import { WebSocketServer } from 'ws';
import { useServer } from 'graphql-ws/lib/use/ws';
import { makeExecutableSchema } from '@graphql-tools/schema';
import { PubSub, withFilter } from 'graphql-subscriptions';

const pubsub = new PubSub(); // In-memory; use Redis PubSub for production

const typeDefs = `#graphql
  type Message {
    id: ID!
    content: String!
    author: User!
    channel: String!
    createdAt: String!
  }

  type User {
    id: ID!
    name: String!
    status: UserStatus!
  }

  enum UserStatus {
    ONLINE
    AWAY
    OFFLINE
  }

  type Query {
    messages(channel: String!, limit: Int = 50): [Message!]!
  }

  type Mutation {
    sendMessage(channel: String!, content: String!): Message!
    updateStatus(status: UserStatus!): User!
  }

  type Subscription {
    messageSent(channel: String!): Message!
    userStatusChanged(userId: ID): User!
    typingIndicator(channel: String!): TypingEvent!
  }

  type TypingEvent {
    userId: ID!
    userName: String!
    channel: String!
    isTyping: Boolean!
  }
`;

const resolvers = {
  Subscription: {
    messageSent: {
      // Filter: only receive messages for the subscribed channel
      subscribe: withFilter(
        () => pubsub.asyncIterableIterator(['MESSAGE_SENT']),
        (payload, variables) => payload.messageSent.channel === variables.channel,
      ),
    },
    userStatusChanged: {
      subscribe: withFilter(
        () => pubsub.asyncIterableIterator(['USER_STATUS_CHANGED']),
        (payload, variables) => {
          // If userId specified, filter to that user; otherwise all users
          return !variables.userId || payload.userStatusChanged.id === variables.userId;
        },
      ),
    },
    typingIndicator: {
      subscribe: withFilter(
        () => pubsub.asyncIterableIterator(['TYPING']),
        (payload, variables) => payload.typingIndicator.channel === variables.channel,
      ),
    },
  },

  Mutation: {
    sendMessage: async (_: unknown, args: { channel: string; content: string }, context: { userId: string }) => {
      const message = {
        id: crypto.randomUUID(),
        content: args.content,
        author: await getUserById(context.userId),
        channel: args.channel,
        createdAt: new Date().toISOString(),
      };

      await saveMessage(message);
      await pubsub.publish('MESSAGE_SENT', { messageSent: message });
      return message;
    },

    updateStatus: async (_: unknown, args: { status: string }, context: { userId: string }) => {
      const user = await updateUserStatus(context.userId, args.status);
      await pubsub.publish('USER_STATUS_CHANGED', { userStatusChanged: user });
      return user;
    },
  },
};

// Server setup with dual HTTP + WebSocket transport
const schema = makeExecutableSchema({ typeDefs, resolvers });
const app = express();
const httpServer = createServer(app);

const wsServer = new WebSocketServer({
  server: httpServer,
  path: '/graphql',
});

// WebSocket server with connection lifecycle
const serverCleanup = useServer({
  schema,
  context: async (ctx) => {
    // Authenticate WebSocket connections
    const token = ctx.connectionParams?.authToken as string;
    if (!token) throw new Error('Missing auth token');
    const user = await verifyToken(token);
    return { userId: user.id };
  },
  onConnect: async (ctx) => {
    console.log('Client connected:', ctx.connectionParams);
  },
  onDisconnect: async (ctx, code, reason) => {
    console.log('Client disconnected:', code, reason);
  },
  onSubscribe: (ctx, msg) => {
    console.log('New subscription:', msg.payload.query);
  },
}, wsServer);

const server = new ApolloServer({ schema });
await server.start();

app.use('/graphql', expressMiddleware(server, {
  context: async ({ req }) => {
    const token = req.headers.authorization?.replace('Bearer ', '');
    const user = token ? await verifyToken(token) : null;
    return { userId: user?.id };
  },
}));

httpServer.listen(4000, () => {
  console.log('Server ready at http://localhost:4000/graphql');
  console.log('WebSocket ready at ws://localhost:4000/graphql');
});
```

## Redis PubSub for Production

```typescript
// lib/redis-pubsub.ts — Scalable PubSub across multiple server instances
import { RedisPubSub } from 'graphql-redis-subscriptions';
import Redis from 'ioredis';

// Separate connections for pub and sub (Redis requirement)
const pubClient = new Redis(process.env.REDIS_URL!, { retryStrategy: (times) => Math.min(times * 50, 2000) });
const subClient = new Redis(process.env.REDIS_URL!, { retryStrategy: (times) => Math.min(times * 50, 2000) });

export const pubsub = new RedisPubSub({
  publisher: pubClient,
  subscriber: subClient,
  reviver: (_key, value) => {
    // Revive Date objects from JSON
    if (typeof value === 'string' && /^\d{4}-\d{2}-\d{2}T/.test(value)) {
      return new Date(value);
    }
    return value;
  },
});

// Graceful shutdown
process.on('SIGTERM', async () => {
  await pubClient.quit();
  await subClient.quit();
});
```

## Client-Side Subscription (React + Apollo)

```tsx
// components/ChatRoom.tsx
import { useSubscription, useMutation, gql } from '@apollo/client';

const MESSAGE_SUBSCRIPTION = gql`
  subscription OnMessageSent($channel: String!) {
    messageSent(channel: $channel) {
      id
      content
      author { id name }
      createdAt
    }
  }
`;

const SEND_MESSAGE = gql`
  mutation SendMessage($channel: String!, $content: String!) {
    sendMessage(channel: $channel, content: $content) {
      id
    }
  }
`;

function ChatRoom({ channel }: { channel: string }) {
  const [messages, setMessages] = useState<Message[]>([]);
  const [sendMessage] = useMutation(SEND_MESSAGE);

  // Subscribe to new messages
  const { loading, error } = useSubscription(MESSAGE_SUBSCRIPTION, {
    variables: { channel },
    onData: ({ data }) => {
      if (data.data?.messageSent) {
        setMessages(prev => [...prev, data.data.messageSent]);
      }
    },
    onError: (error) => {
      console.error('Subscription error:', error);
      // Handle reconnection logic
    },
  });

  // Alternative: update Apollo cache directly
  // useSubscription(MESSAGE_SUBSCRIPTION, {
  //   variables: { channel },
  //   onData: ({ client, data }) => {
  //     client.cache.modify({
  //       fields: {
  //         messages(existing = []) {
  //           const newRef = client.cache.writeFragment({
  //             data: data.data.messageSent,
  //             fragment: gql`fragment NewMessage on Message { id content author { id name } }`,
  //           });
  //           return [...existing, newRef];
  //         },
  //       },
  //     });
  //   },
  // });

  const handleSend = async (content: string) => {
    await sendMessage({ variables: { channel, content } });
  };

  if (error) return <div>Subscription error: {error.message}</div>;

  return (
    <div>
      {loading && <p>Connecting...</p>}
      {messages.map(msg => (
        <div key={msg.id}>
          <strong>{msg.author.name}</strong>: {msg.content}
        </div>
      ))}
      <MessageInput onSend={handleSend} />
    </div>
  );
}
```

## Apollo Client WebSocket Link Setup

```typescript
// lib/apollo-client.ts
import { ApolloClient, InMemoryCache, split, HttpLink } from '@apollo/client';
import { GraphQLWsLink } from '@apollo/client/link/subscriptions';
import { createClient } from 'graphql-ws';
import { getMainDefinition } from '@apollo/client/utilities';

const httpLink = new HttpLink({
  uri: 'http://localhost:4000/graphql',
  headers: { authorization: `Bearer ${getToken()}` },
});

const wsLink = new GraphQLWsLink(
  createClient({
    url: 'ws://localhost:4000/graphql',
    connectionParams: () => ({
      authToken: getToken(),
    }),
    shouldRetry: () => true,
    retryAttempts: Infinity,
    retryWait: (retries) => new Promise(resolve =>
      setTimeout(resolve, Math.min(1000 * 2 ** retries, 30000)) // Exponential backoff, max 30s
    ),
    on: {
      connected: () => console.log('WebSocket connected'),
      closed: (event) => console.log('WebSocket closed:', event),
      error: (error) => console.error('WebSocket error:', error),
    },
  })
);

// Route subscriptions to WebSocket, everything else to HTTP
const splitLink = split(
  ({ query }) => {
    const definition = getMainDefinition(query);
    return definition.kind === 'OperationDefinition' && definition.operation === 'subscription';
  },
  wsLink,
  httpLink,
);

export const client = new ApolloClient({
  link: splitLink,
  cache: new InMemoryCache(),
});
```

## Gotchas

1. **In-memory PubSub doesn't scale** -- `graphql-subscriptions` PubSub stores subscribers in-process memory. With 2+ server instances behind a load balancer, a message published on server A doesn't reach subscribers on server B. Always use Redis PubSub, Kafka, or another distributed backend in production.

2. **WebSocket auth is connection-level, not per-operation** -- Authentication happens once during `onConnect` via `connectionParams`. If a user's token expires mid-session, existing subscriptions continue with stale auth. Implement periodic token validation or close connections with expired tokens.

3. **withFilter runs for EVERY published event** -- The filter function in `withFilter` executes for every subscriber on every published event. With 10,000 subscribers and 100 events/second, that's 1,000,000 filter calls/second. Keep filter logic simple (direct equality checks) and avoid database calls inside filters.

4. **Subscription memory leaks** -- Every active subscription holds a reference to its async iterator. If clients disconnect without proper cleanup (browser tab closed, network drop), the iterator stays in memory. Set `connectionInitWaitTimeout` to reject slow connections and implement heartbeat-based cleanup.

5. **N+1 problem in subscription resolvers** -- Subscription payloads often need related data (e.g., message + author). If you resolve `author` inside the subscription resolver, each published message triggers a database query for the author. Pre-load related data before publishing to PubSub instead.

6. **Subscription transport confusion** -- `subscriptions-transport-ws` is deprecated. Use `graphql-ws` (the newer protocol). Apollo Client's `WebSocketLink` uses the old protocol; `GraphQLWsLink` uses the new one. Mixing protocols causes silent connection failures with no error message.
