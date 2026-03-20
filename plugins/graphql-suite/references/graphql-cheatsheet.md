# GraphQL Cheatsheet

## Schema Basics

```graphql
# Scalar types: String, Int, Float, Boolean, ID
# ! = non-null, [] = list

type User {
  id: ID!
  name: String!
  email: String!
  age: Int
  posts: [Post!]!
}

input CreateUserInput {
  name: String!
  email: String!
}

type Query {
  user(id: ID!): User
  users(first: Int!, after: String): UserConnection!
}

type Mutation {
  createUser(input: CreateUserInput!): CreateUserPayload!
}

type Subscription {
  messageAdded(channelId: ID!): Message!
}

# Pagination (Relay-style)
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

# Union types for error handling
type CreateUserPayload {
  user: User
  errors: [FieldError!]!
}

type FieldError {
  field: String!
  message: String!
}

# Enum
enum Role { ADMIN USER MODERATOR }

# Interface
interface Node { id: ID! }
type User implements Node { id: ID! name: String! }
```

---

## Apollo Server

### Setup
```bash
npm install @apollo/server graphql
```

### Resolver Pattern
```typescript
const resolvers = {
  Query: {
    user: (parent, { id }, context, info) => context.db.users.findById(id),
  },
  Mutation: {
    createUser: (_, { input }, ctx) => ctx.db.users.create(input),
  },
  User: {
    posts: (user, _, { loaders }) => loaders.postsByAuthor.load(user.id),
  },
};
```

### DataLoader (N+1 fix)
```typescript
import DataLoader from "dataloader";

// Create per request in context
const userLoader = new DataLoader(async (ids) => {
  const users = await db.users.findMany({ where: { id: { in: ids } } });
  const map = new Map(users.map(u => [u.id, u]));
  return ids.map(id => map.get(id) || null);
});
```

### Error Handling
```typescript
import { GraphQLError } from "graphql";

throw new GraphQLError("Not found", {
  extensions: { code: "NOT_FOUND" },
});
```

---

## Apollo Client

### Setup
```bash
npm install @apollo/client graphql
```

### Provider
```tsx
import { ApolloClient, InMemoryCache, ApolloProvider } from "@apollo/client";

const client = new ApolloClient({
  uri: "/graphql",
  cache: new InMemoryCache(),
});

<ApolloProvider client={client}><App /></ApolloProvider>
```

### useQuery
```tsx
const { data, loading, error, refetch, fetchMore } = useQuery(GET_USERS, {
  variables: { first: 20 },
  fetchPolicy: "cache-and-network",   // or "network-only", "cache-first"
  pollInterval: 30000,                 // Re-fetch every 30s
});
```

### useMutation
```tsx
const [createUser, { data, loading, error }] = useMutation(CREATE_USER, {
  variables: { input: { name, email } },
  optimisticResponse: { ... },
  update: (cache, { data }) => { cache.modify({ ... }); },
  refetchQueries: ["GetUsers"],       // Re-fetch after mutation
  onCompleted: (data) => { ... },
  onError: (error) => { ... },
});
```

### useSubscription
```tsx
const { data } = useSubscription(MESSAGE_ADDED, {
  variables: { channelId },
});
```

### Fetch Policies

| Policy | Behavior |
|--------|----------|
| `cache-first` | Read cache, fetch only if miss (default) |
| `cache-and-network` | Read cache immediately, then fetch and update |
| `network-only` | Always fetch, write to cache |
| `cache-only` | Only read cache, never fetch |
| `no-cache` | Always fetch, don't write to cache |

---

## graphql-codegen

### Setup
```bash
npm install -D @graphql-codegen/cli @graphql-codegen/client-preset
```

### Config (codegen.ts)
```typescript
import type { CodegenConfig } from "@graphql-codegen/cli";

const config: CodegenConfig = {
  schema: "http://localhost:4000/graphql",
  documents: ["src/**/*.{ts,tsx}"],
  generates: {
    "./src/gql/": {
      preset: "client",
      config: { enumsAsTypes: true },
    },
  },
};
export default config;
```

### Usage
```tsx
// Import gql from GENERATED output, not @apollo/client
import { gql } from "../gql";

const GET_USER = gql(`
  query GetUser($id: ID!) {
    user(id: $id) { id name email }
  }
`);

// Variables and return types are auto-typed
const { data } = useQuery(GET_USER, { variables: { id: "123" } });
```

### Commands
```bash
npx graphql-codegen          # Generate once
npx graphql-codegen --watch  # Watch mode
```

---

## Quick Decision

| Scenario | Client Library |
|----------|---------------|
| Full-featured React app | **Apollo Client** |
| Simple/medium React app | **urql** |
| Already using TanStack Query | **graphql-request** + TanStack |
| Facebook-scale app | **Relay** |

| Scenario | Server Approach |
|----------|----------------|
| Team collaboration, API-first | **Schema-first** (SDL) |
| TypeScript-heavy, rapid iteration | **Code-first** (Pothos/Nexus) |
| Existing REST API migration | **Schema-first** with REST data sources |

---

## Common Patterns

### Cursor-Based Pagination Query
```graphql
query Users($first: Int!, $after: String) {
  users(first: $first, after: $after) {
    edges { cursor node { id name } }
    pageInfo { hasNextPage endCursor }
  }
}
```

### Mutation with Error Handling
```graphql
mutation CreateUser($input: CreateUserInput!) {
  createUser(input: $input) {
    user { id name email }
    errors { field message }
  }
}
```

### Fragment Colocation
```graphql
fragment UserCard on User {
  id name email avatarUrl
}

query GetUsers {
  users { ...UserCard }
}
```
