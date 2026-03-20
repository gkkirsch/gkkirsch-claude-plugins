---
name: apollo-development
description: >
  Apollo GraphQL development — Apollo Server for building APIs and Apollo Client
  for React apps with caching, mutations, subscriptions, and error handling.
  Triggers: "apollo", "apollo server", "apollo client", "graphql api",
  "graphql server", "graphql react", "graphql mutation", "graphql subscription".
  NOT for: Code generation (use graphql-codegen).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Apollo GraphQL

## Apollo Server

### Setup

```bash
npm install @apollo/server graphql
```

### Basic Server

```typescript
// src/server.ts
import { ApolloServer } from "@apollo/server";
import { expressMiddleware } from "@apollo/server/express4";
import express from "express";
import cors from "cors";

const typeDefs = `#graphql
  type Query {
    hello: String!
    user(id: ID!): User
    users(first: Int = 20, after: String): UserConnection!
  }

  type Mutation {
    createUser(input: CreateUserInput!): CreateUserPayload!
    updateUser(id: ID!, input: UpdateUserInput!): UpdateUserPayload!
    deleteUser(id: ID!): DeleteUserPayload!
  }

  type User {
    id: ID!
    name: String!
    email: String!
    posts: [Post!]!
    createdAt: String!
  }

  type Post {
    id: ID!
    title: String!
    content: String!
    author: User!
  }

  input CreateUserInput {
    name: String!
    email: String!
  }

  input UpdateUserInput {
    name: String
    email: String
  }

  type CreateUserPayload {
    user: User
    errors: [FieldError!]!
  }

  type UpdateUserPayload {
    user: User
    errors: [FieldError!]!
  }

  type DeleteUserPayload {
    success: Boolean!
    errors: [FieldError!]!
  }

  type FieldError {
    field: String!
    message: String!
  }

  type UserConnection {
    edges: [UserEdge!]!
    pageInfo: PageInfo!
    totalCount: Int!
  }

  type UserEdge {
    cursor: String!
    node: User!
  }

  type PageInfo {
    hasNextPage: Boolean!
    hasPreviousPage: Boolean!
    startCursor: String
    endCursor: String
  }
`;

const resolvers = {
  Query: {
    hello: () => "Hello, GraphQL!",
    user: async (_: any, { id }: { id: string }, { dataSources }: Context) => {
      return dataSources.users.findById(id);
    },
    users: async (_: any, { first, after }: { first: number; after?: string }, { dataSources }: Context) => {
      return dataSources.users.findMany({ first, after });
    },
  },
  Mutation: {
    createUser: async (_: any, { input }: { input: any }, { dataSources }: Context) => {
      try {
        const user = await dataSources.users.create(input);
        return { user, errors: [] };
      } catch (err) {
        return { user: null, errors: [{ field: "email", message: "Email already exists" }] };
      }
    },
    updateUser: async (_: any, { id, input }: any, { dataSources }: Context) => {
      const user = await dataSources.users.update(id, input);
      return { user, errors: [] };
    },
    deleteUser: async (_: any, { id }: { id: string }, { dataSources }: Context) => {
      await dataSources.users.delete(id);
      return { success: true, errors: [] };
    },
  },
  User: {
    posts: async (user: any, _: any, { dataSources }: Context) => {
      return dataSources.posts.findByAuthorId(user.id);
    },
  },
  Post: {
    author: async (post: any, _: any, { dataSources }: Context) => {
      return dataSources.users.findById(post.authorId);
    },
  },
};

interface Context {
  dataSources: {
    users: UserDataSource;
    posts: PostDataSource;
  };
  user?: { id: string; role: string };
}

const server = new ApolloServer<Context>({ typeDefs, resolvers });

const app = express();

await server.start();

app.use(
  "/graphql",
  cors(),
  express.json(),
  expressMiddleware(server, {
    context: async ({ req }) => {
      const token = req.headers.authorization?.replace("Bearer ", "");
      const user = token ? await verifyToken(token) : undefined;
      return {
        user,
        dataSources: {
          users: new UserDataSource(db),
          posts: new PostDataSource(db),
        },
      };
    },
  })
);

app.listen(4000, () => console.log("Server at http://localhost:4000/graphql"));
```

### DataLoader (Solve N+1)

```bash
npm install dataloader
```

```typescript
// src/dataloaders.ts
import DataLoader from "dataloader";

// Batch function: takes array of IDs, returns array of results IN SAME ORDER
export function createUserLoader(db: Database) {
  return new DataLoader<string, User>(async (ids) => {
    const users = await db.users.findMany({
      where: { id: { in: ids as string[] } },
    });

    // Must return results in same order as input ids
    const userMap = new Map(users.map((u) => [u.id, u]));
    return ids.map((id) => userMap.get(id) || null);
  });
}

export function createPostsByAuthorLoader(db: Database) {
  return new DataLoader<string, Post[]>(async (authorIds) => {
    const posts = await db.posts.findMany({
      where: { authorId: { in: authorIds as string[] } },
    });

    const postsByAuthor = new Map<string, Post[]>();
    for (const post of posts) {
      const existing = postsByAuthor.get(post.authorId) || [];
      existing.push(post);
      postsByAuthor.set(post.authorId, existing);
    }

    return authorIds.map((id) => postsByAuthor.get(id) || []);
  });
}

// In context factory — create new loaders per request (caching is per-request)
context: async ({ req }) => ({
  loaders: {
    user: createUserLoader(db),
    postsByAuthor: createPostsByAuthorLoader(db),
  },
}),

// In resolvers
User: {
  posts: (user, _, { loaders }) => loaders.postsByAuthor.load(user.id),
},
Post: {
  author: (post, _, { loaders }) => loaders.user.load(post.authorId),
},
```

### Authentication & Authorization

```typescript
import { GraphQLError } from "graphql";

// Auth directive (field-level)
function requireAuth(context: Context) {
  if (!context.user) {
    throw new GraphQLError("Not authenticated", {
      extensions: { code: "UNAUTHENTICATED" },
    });
  }
  return context.user;
}

function requireRole(context: Context, role: string) {
  const user = requireAuth(context);
  if (user.role !== role) {
    throw new GraphQLError("Not authorized", {
      extensions: { code: "FORBIDDEN" },
    });
  }
  return user;
}

// Usage in resolvers
const resolvers = {
  Query: {
    me: (_, __, context) => {
      const user = requireAuth(context);
      return context.dataSources.users.findById(user.id);
    },
    adminDashboard: (_, __, context) => {
      requireRole(context, "ADMIN");
      return context.dataSources.admin.getDashboard();
    },
  },
  User: {
    email: (user, _, context) => {
      // Only show email to the user themselves or admins
      if (context.user?.id === user.id || context.user?.role === "ADMIN") {
        return user.email;
      }
      return null;
    },
  },
};
```

### Error Handling

```typescript
import { GraphQLError } from "graphql";

// Custom error classes
class NotFoundError extends GraphQLError {
  constructor(entity: string, id: string) {
    super(`${entity} not found: ${id}`, {
      extensions: { code: "NOT_FOUND", entity, id },
    });
  }
}

class ValidationError extends GraphQLError {
  constructor(errors: { field: string; message: string }[]) {
    super("Validation failed", {
      extensions: { code: "VALIDATION_ERROR", errors },
    });
  }
}

// Error formatting plugin
const server = new ApolloServer({
  typeDefs,
  resolvers,
  formatError: (formattedError, error) => {
    // Don't expose internal errors in production
    if (process.env.NODE_ENV === "production") {
      if (formattedError.extensions?.code === "INTERNAL_SERVER_ERROR") {
        return { message: "Internal server error", extensions: { code: "INTERNAL_SERVER_ERROR" } };
      }
    }
    return formattedError;
  },
});
```

### Subscriptions (WebSocket)

```bash
npm install graphql-ws ws
```

```typescript
import { WebSocketServer } from "ws";
import { useServer } from "graphql-ws/lib/use/ws";
import { makeExecutableSchema } from "@graphql-tools/schema";

const schema = makeExecutableSchema({ typeDefs, resolvers });

// Add subscription type
const typeDefs = `#graphql
  type Subscription {
    messageAdded(channelId: ID!): Message!
    userStatusChanged: UserStatus!
  }
`;

// Subscription resolvers use AsyncIterator
import { PubSub } from "graphql-subscriptions";
const pubsub = new PubSub();

const resolvers = {
  Subscription: {
    messageAdded: {
      subscribe: (_, { channelId }) => {
        return pubsub.asyncIterableIterator(`MESSAGE_ADDED_${channelId}`);
      },
    },
  },
  Mutation: {
    sendMessage: async (_, { channelId, content }, context) => {
      const message = await createMessage(channelId, content, context.user.id);
      pubsub.publish(`MESSAGE_ADDED_${channelId}`, { messageAdded: message });
      return message;
    },
  },
};

// WebSocket server setup
const wsServer = new WebSocketServer({ server: httpServer, path: "/graphql" });
useServer(
  {
    schema,
    context: async (ctx) => {
      const token = ctx.connectionParams?.authToken as string;
      const user = token ? await verifyToken(token) : null;
      return { user };
    },
  },
  wsServer
);
```

## Apollo Client (React)

### Setup

```bash
npm install @apollo/client graphql
```

```typescript
// src/lib/apollo-client.ts
import { ApolloClient, InMemoryCache, createHttpLink, split } from "@apollo/client";
import { setContext } from "@apollo/client/link/context";
import { GraphQLWsLink } from "@apollo/client/link/subscriptions";
import { getMainDefinition } from "@apollo/client/utilities";
import { createClient } from "graphql-ws";

const httpLink = createHttpLink({
  uri: "/graphql",
});

const authLink = setContext((_, { headers }) => {
  const token = localStorage.getItem("token");
  return {
    headers: {
      ...headers,
      authorization: token ? `Bearer ${token}` : "",
    },
  };
});

const wsLink = new GraphQLWsLink(
  createClient({
    url: "ws://localhost:4000/graphql",
    connectionParams: () => ({
      authToken: localStorage.getItem("token"),
    }),
  })
);

// Split between HTTP (queries/mutations) and WebSocket (subscriptions)
const splitLink = split(
  ({ query }) => {
    const definition = getMainDefinition(query);
    return definition.kind === "OperationDefinition" && definition.operation === "subscription";
  },
  wsLink,
  authLink.concat(httpLink)
);

export const client = new ApolloClient({
  link: splitLink,
  cache: new InMemoryCache({
    typePolicies: {
      Query: {
        fields: {
          users: {
            // Cursor-based pagination merge
            keyArgs: false,
            merge(existing, incoming) {
              if (!existing) return incoming;
              return {
                ...incoming,
                edges: [...existing.edges, ...incoming.edges],
              };
            },
          },
        },
      },
    },
  }),
});
```

```tsx
// src/main.tsx
import { ApolloProvider } from "@apollo/client";
import { client } from "./lib/apollo-client";

function App() {
  return (
    <ApolloProvider client={client}>
      <Router />
    </ApolloProvider>
  );
}
```

### Queries

```tsx
import { useQuery, gql } from "@apollo/client";

const GET_USERS = gql`
  query GetUsers($first: Int!, $after: String) {
    users(first: $first, after: $after) {
      edges {
        cursor
        node {
          id
          name
          email
        }
      }
      pageInfo {
        hasNextPage
        endCursor
      }
      totalCount
    }
  }
`;

function UserList() {
  const { data, loading, error, fetchMore } = useQuery(GET_USERS, {
    variables: { first: 20 },
  });

  if (loading) return <p>Loading...</p>;
  if (error) return <p>Error: {error.message}</p>;

  return (
    <div>
      {data.users.edges.map(({ node }) => (
        <div key={node.id}>{node.name} — {node.email}</div>
      ))}
      {data.users.pageInfo.hasNextPage && (
        <button onClick={() => fetchMore({
          variables: { after: data.users.pageInfo.endCursor },
        })}>
          Load More
        </button>
      )}
    </div>
  );
}
```

### Mutations with Optimistic Updates

```tsx
import { useMutation, gql } from "@apollo/client";

const CREATE_USER = gql`
  mutation CreateUser($input: CreateUserInput!) {
    createUser(input: $input) {
      user {
        id
        name
        email
      }
      errors {
        field
        message
      }
    }
  }
`;

function CreateUserForm() {
  const [createUser, { loading }] = useMutation(CREATE_USER, {
    // Optimistic response — immediately update UI
    optimisticResponse: {
      createUser: {
        __typename: "CreateUserPayload",
        user: {
          __typename: "User",
          id: "temp-id",
          name: formData.name,
          email: formData.email,
        },
        errors: [],
      },
    },
    // Update cache after mutation
    update: (cache, { data }) => {
      if (data?.createUser.user) {
        cache.modify({
          fields: {
            users(existingUsers = { edges: [] }) {
              const newEdge = {
                __typename: "UserEdge",
                cursor: data.createUser.user.id,
                node: cache.writeFragment({
                  data: data.createUser.user,
                  fragment: gql`
                    fragment NewUser on User {
                      id name email
                    }
                  `,
                }),
              };
              return {
                ...existingUsers,
                edges: [newEdge, ...existingUsers.edges],
                totalCount: existingUsers.totalCount + 1,
              };
            },
          },
        });
      }
    },
    onError: (error) => {
      console.error("Create user failed:", error);
    },
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const { data } = await createUser({ variables: { input: formData } });
    if (data?.createUser.errors.length) {
      // Show field-level errors
      setErrors(data.createUser.errors);
    }
  };

  return <form onSubmit={handleSubmit}>...</form>;
}
```

### Subscriptions

```tsx
import { useSubscription, gql } from "@apollo/client";

const MESSAGE_SUBSCRIPTION = gql`
  subscription OnMessageAdded($channelId: ID!) {
    messageAdded(channelId: $channelId) {
      id
      content
      author {
        id
        name
      }
      createdAt
    }
  }
`;

function ChatMessages({ channelId }: { channelId: string }) {
  const { data: subData } = useSubscription(MESSAGE_SUBSCRIPTION, {
    variables: { channelId },
  });

  // subData.messageAdded is the latest message from the subscription

  return <div>...</div>;
}
```

## Gotchas

1. **DataLoader must be created per-request** — DataLoader caches results for the lifetime of the instance. If you create it once at startup, stale data is served. Create new loaders in the context factory for each request.

2. **Apollo Client cache requires `__typename` and `id`** — If your objects lack these, the normalized cache can't track them. Always select `id` in queries. For objects without a natural ID, use `keyFields` in type policies.

3. **`useQuery` re-renders on every cache update** — If another component writes to the same cache key, your component re-renders. Use `fetchPolicy: "cache-and-network"` for balance, or `"network-only"` to skip cache entirely.

4. **Subscription transport changed** — Apollo Client 3.x uses `graphql-ws` (not the older `subscriptions-transport-ws`). The protocols are incompatible. Server and client must use the same library.

5. **Error objects are not thrown by default** — `useQuery` and `useMutation` return errors in the `error` field, they don't throw. Check `error` explicitly. For network errors vs GraphQL errors, check `error.networkError` and `error.graphQLErrors`.

6. **Fragment colocation is critical at scale** — Each component should define a fragment for the data it needs, composed into parent queries. Without this, changing one component's data needs breaks queries across the app.
