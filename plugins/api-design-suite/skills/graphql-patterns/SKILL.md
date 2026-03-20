---
name: graphql-patterns
description: >
  GraphQL API implementation — schema design, resolvers, mutations, subscriptions,
  DataLoader for N+1 prevention, authentication, and error handling.
  Triggers: "GraphQL", "graphql schema", "graphql resolver", "graphql mutation",
  "graphql subscription", "apollo server", "DataLoader", "graphql auth".
  NOT for: REST APIs (use rest-api-patterns), API architecture decisions (use api-architect agent).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# GraphQL Patterns

## Schema Design

```graphql
# schema.graphql
type Query {
  user(id: ID!): User
  users(filter: UserFilter, pagination: PaginationInput): UserConnection!
  me: User
  post(id: ID!): Post
  posts(filter: PostFilter, first: Int, after: String): PostConnection!
}

type Mutation {
  createUser(input: CreateUserInput!): CreateUserPayload!
  updateUser(id: ID!, input: UpdateUserInput!): UpdateUserPayload!
  deleteUser(id: ID!): DeleteUserPayload!
  createPost(input: CreatePostInput!): CreatePostPayload!
  publishPost(id: ID!): PublishPostPayload!
}

type Subscription {
  postPublished: Post!
  messageReceived(channelId: ID!): Message!
}

# Types
type User {
  id: ID!
  name: String!
  email: String!
  avatar: String
  role: Role!
  posts(first: Int, after: String): PostConnection!
  createdAt: DateTime!
}

type Post {
  id: ID!
  title: String!
  body: String!
  author: User!
  tags: [String!]!
  isPublished: Boolean!
  publishedAt: DateTime
  createdAt: DateTime!
  updatedAt: DateTime!
}

# Enums
enum Role {
  USER
  EDITOR
  ADMIN
}

enum PostOrderBy {
  CREATED_AT_ASC
  CREATED_AT_DESC
  TITLE_ASC
  TITLE_DESC
}

# Input types (always use Input suffix)
input CreateUserInput {
  name: String!
  email: String!
  role: Role
}

input UpdateUserInput {
  name: String
  email: String
  role: Role
}

input CreatePostInput {
  title: String!
  body: String!
  tags: [String!]
}

input UserFilter {
  role: Role
  search: String
}

input PostFilter {
  authorId: ID
  isPublished: Boolean
  tag: String
  orderBy: PostOrderBy
}

input PaginationInput {
  page: Int
  limit: Int
}

# Mutation payloads (always return the affected object + errors)
type CreateUserPayload {
  user: User
  errors: [FieldError!]
}

type UpdateUserPayload {
  user: User
  errors: [FieldError!]
}

type DeleteUserPayload {
  success: Boolean!
}

type PublishPostPayload {
  post: Post
  errors: [FieldError!]
}

type FieldError {
  field: String!
  message: String!
}

# Relay-style connections for pagination
type PostConnection {
  edges: [PostEdge!]!
  pageInfo: PageInfo!
  totalCount: Int!
}

type PostEdge {
  node: Post!
  cursor: String!
}

type PageInfo {
  hasNextPage: Boolean!
  hasPreviousPage: Boolean!
  startCursor: String
  endCursor: String
}

# Custom scalars
scalar DateTime
scalar JSON
```

## Apollo Server Setup

```typescript
import { ApolloServer } from '@apollo/server';
import { expressMiddleware } from '@apollo/server/express4';
import { readFileSync } from 'fs';
import { resolvers } from './resolvers';

const typeDefs = readFileSync('./schema.graphql', 'utf-8');

const server = new ApolloServer({
  typeDefs,
  resolvers,
  formatError: (error) => {
    // Don't expose internal errors in production
    if (error.extensions?.code === 'INTERNAL_SERVER_ERROR') {
      console.error('GraphQL error:', error);
      return { message: 'An unexpected error occurred' };
    }
    return error;
  },
});

await server.start();

app.use('/graphql',
  express.json(),
  expressMiddleware(server, {
    context: async ({ req }) => ({
      user: req.user, // From auth middleware
      db,
      loaders: createLoaders(), // DataLoader instances
    }),
  })
);
```

## Resolvers

```typescript
import { GraphQLError } from 'graphql';

const resolvers = {
  Query: {
    me: (_: any, __: any, { user }: Context) => {
      if (!user) throw new GraphQLError('Not authenticated', {
        extensions: { code: 'UNAUTHENTICATED' },
      });
      return db.user.findUnique({ where: { id: user.id } });
    },

    user: (_: any, { id }: { id: string }) => {
      return db.user.findUnique({ where: { id } });
    },

    users: async (_: any, { filter, pagination }: any) => {
      const page = pagination?.page || 1;
      const limit = Math.min(pagination?.limit || 20, 100);
      const where: any = {};

      if (filter?.role) where.role = filter.role;
      if (filter?.search) {
        where.OR = [
          { name: { contains: filter.search, mode: 'insensitive' } },
          { email: { contains: filter.search, mode: 'insensitive' } },
        ];
      }

      const [users, total] = await Promise.all([
        db.user.findMany({ where, skip: (page - 1) * limit, take: limit }),
        db.user.count({ where }),
      ]);

      return {
        edges: users.map(u => ({ node: u, cursor: u.id })),
        pageInfo: {
          hasNextPage: page * limit < total,
          hasPreviousPage: page > 1,
        },
        totalCount: total,
      };
    },
  },

  Mutation: {
    createUser: async (_: any, { input }: any, { user }: Context) => {
      if (!user || user.role !== 'ADMIN') {
        throw new GraphQLError('Not authorized', {
          extensions: { code: 'FORBIDDEN' },
        });
      }

      try {
        const newUser = await db.user.create({ data: input });
        return { user: newUser, errors: [] };
      } catch (err: any) {
        if (err.code === 'P2002') {
          return {
            user: null,
            errors: [{ field: 'email', message: 'Email already registered' }],
          };
        }
        throw err;
      }
    },

    deleteUser: async (_: any, { id }: { id: string }, { user }: Context) => {
      if (!user || user.role !== 'ADMIN') {
        throw new GraphQLError('Not authorized', {
          extensions: { code: 'FORBIDDEN' },
        });
      }

      await db.user.delete({ where: { id } });
      return { success: true };
    },
  },

  // Field resolvers (for relationships)
  User: {
    posts: (parent: any, { first = 10, after }: any, { loaders }: Context) => {
      // Use DataLoader to avoid N+1
      return loaders.postsByAuthor.load(parent.id);
    },
  },

  Post: {
    author: (parent: any, _: any, { loaders }: Context) => {
      return loaders.userById.load(parent.authorId);
    },
  },
};
```

## DataLoader (N+1 Prevention)

```typescript
import DataLoader from 'dataloader';

function createLoaders() {
  return {
    // Batch load users by ID
    userById: new DataLoader<string, User>(async (ids) => {
      const users = await db.user.findMany({
        where: { id: { in: [...ids] } },
      });
      const userMap = new Map(users.map(u => [u.id, u]));
      return ids.map(id => userMap.get(id) || null);
    }),

    // Batch load posts by author ID (one-to-many)
    postsByAuthor: new DataLoader<string, Post[]>(async (authorIds) => {
      const posts = await db.post.findMany({
        where: { authorId: { in: [...authorIds] } },
        orderBy: { createdAt: 'desc' },
      });
      const postMap = new Map<string, Post[]>();
      posts.forEach(p => {
        const existing = postMap.get(p.authorId) || [];
        existing.push(p);
        postMap.set(p.authorId, existing);
      });
      return authorIds.map(id => postMap.get(id) || []);
    }),
  };
}

// IMPORTANT: Create new loaders per request!
// DataLoader caches results within a request lifecycle.
// Sharing across requests would serve stale data.
```

## Authentication & Authorization

```typescript
// Context creation with auth
const context = async ({ req }: any) => {
  const token = req.headers.authorization?.replace('Bearer ', '');
  let user = null;

  if (token) {
    try {
      const payload = jwt.verify(token, JWT_SECRET);
      user = await db.user.findUnique({ where: { id: payload.sub } });
    } catch {
      // Invalid token — continue with user = null
    }
  }

  return { user, db, loaders: createLoaders() };
};

// Auth directive (schema-level)
// In schema: directive @auth(requires: Role) on FIELD_DEFINITION
// In resolver: check user.role against directive arg

// Simple auth check helper
function requireAuth(user: User | null): asserts user is User {
  if (!user) {
    throw new GraphQLError('Authentication required', {
      extensions: { code: 'UNAUTHENTICATED' },
    });
  }
}

function requireRole(user: User | null, role: string): asserts user is User {
  requireAuth(user);
  if (user.role !== role) {
    throw new GraphQLError('Insufficient permissions', {
      extensions: { code: 'FORBIDDEN' },
    });
  }
}
```

## Subscriptions (Real-time)

```typescript
import { PubSub } from 'graphql-subscriptions';

const pubsub = new PubSub(); // Use Redis PubSub in production

const resolvers = {
  Mutation: {
    publishPost: async (_: any, { id }: any) => {
      const post = await db.post.update({
        where: { id },
        data: { isPublished: true, publishedAt: new Date() },
        include: { author: true },
      });

      // Publish event
      pubsub.publish('POST_PUBLISHED', { postPublished: post });

      return { post, errors: [] };
    },
  },

  Subscription: {
    postPublished: {
      subscribe: () => pubsub.asyncIterableIterator(['POST_PUBLISHED']),
    },

    messageReceived: {
      subscribe: (_: any, { channelId }: any) =>
        pubsub.asyncIterableIterator([`MESSAGE_${channelId}`]),
    },
  },
};
```

## Gotchas

1. **DataLoader must be created per request.** If you share a DataLoader across requests, the first request populates the cache, and subsequent requests get stale data. Create fresh loaders in the context function.

2. **GraphQL errors have two categories.** Use `GraphQLError` with `extensions.code` for expected errors (UNAUTHENTICATED, FORBIDDEN, BAD_USER_INPUT). Let unexpected errors bubble up — Apollo formats them as INTERNAL_SERVER_ERROR automatically.

3. **N+1 queries are the default in GraphQL.** Without DataLoader, resolving `users { posts { ... } }` executes 1 query for users + N queries for posts. DataLoader batches these into 2 queries total. Always use DataLoader for relationships.

4. **Mutation payloads should include errors.** Don't rely solely on GraphQL errors for validation. Return `{ user: null, errors: [{ field, message }] }` so the client can show field-level errors without parsing error extensions.

5. **Don't expose your entire database schema.** GraphQL makes it easy to mirror your DB tables as types. But internal fields (password hashes, internal flags, soft-delete timestamps) should be excluded. Design the schema for the client's needs, not the database's structure.

6. **`PubSub` from graphql-subscriptions is in-memory only.** It won't work across multiple server instances. Use `graphql-redis-subscriptions` or similar for production with multiple servers.
