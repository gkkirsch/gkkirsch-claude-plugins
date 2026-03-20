---
name: graphql-optimization
description: >
  GraphQL performance optimization — DataLoader for N+1 queries,
  query complexity analysis, persisted queries, caching strategies,
  and Apollo Federation for microservices.
  Triggers: "graphql n+1", "dataloader", "graphql cache", "graphql performance",
  "query complexity", "persisted queries", "apollo federation", "graphql rate limit".
  NOT for: Schema design (use graphql-schema-resolvers), REST APIs.
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# GraphQL Performance Optimization

## DataLoader — Solving N+1 Queries

```bash
npm install dataloader
```

```typescript
import DataLoader from "dataloader";
import { PrismaClient } from "@prisma/client";

const prisma = new PrismaClient();

// Basic DataLoader
function createUserLoader() {
  return new DataLoader<string, User | null>(async (userIds) => {
    const users = await prisma.user.findMany({
      where: { id: { in: [...userIds] } },
    });
    // MUST return in same order as input IDs
    const userMap = new Map(users.map((u) => [u.id, u]));
    return userIds.map((id) => userMap.get(id) ?? null);
  });
}

// DataLoader with relations
function createPostsByAuthorLoader() {
  return new DataLoader<string, Post[]>(async (authorIds) => {
    const posts = await prisma.post.findMany({
      where: { authorId: { in: [...authorIds] } },
      orderBy: { createdAt: "desc" },
    });
    const grouped = new Map<string, Post[]>();
    for (const post of posts) {
      const existing = grouped.get(post.authorId) ?? [];
      existing.push(post);
      grouped.set(post.authorId, existing);
    }
    return authorIds.map((id) => grouped.get(id) ?? []);
  });
}

// DataLoader with composite keys
function createPostTagsLoader() {
  return new DataLoader<string, Tag[]>(async (postIds) => {
    const postTags = await prisma.tag.findMany({
      where: { posts: { some: { id: { in: [...postIds] } } } },
      include: { posts: { select: { id: true } } },
    });
    const grouped = new Map<string, Tag[]>();
    for (const tag of postTags) {
      for (const post of tag.posts) {
        const existing = grouped.get(post.id) ?? [];
        existing.push(tag);
        grouped.set(post.id, existing);
      }
    }
    return postIds.map((id) => grouped.get(id) ?? []);
  });
}

// Create all loaders per request
function createLoaders() {
  return {
    userLoader: createUserLoader(),
    postsByAuthorLoader: createPostsByAuthorLoader(),
    postTagsLoader: createPostTagsLoader(),
  };
}

// Use in resolvers
const resolvers = {
  Post: {
    author: (parent: Post, _: any, ctx: Context) => {
      return ctx.loaders.userLoader.load(parent.authorId);
    },
    tags: (parent: Post, _: any, ctx: Context) => {
      return ctx.loaders.postTagsLoader.load(parent.id);
    },
  },
  User: {
    posts: (parent: User, _: any, ctx: Context) => {
      return ctx.loaders.postsByAuthorLoader.load(parent.id);
    },
  },
};
```

## Query Complexity & Depth Limiting

```bash
npm install graphql-query-complexity graphql-depth-limit
```

```typescript
import { createComplexityLimitRule } from "graphql-validation-complexity";
import depthLimit from "graphql-depth-limit";
import {
  getComplexity,
  simpleEstimator,
  fieldExtensionsEstimator,
} from "graphql-query-complexity";

// Apollo Server plugin for complexity analysis
const complexityPlugin = {
  requestDidStart: async () => ({
    async didResolveOperation({ request, document, schema }: any) {
      const complexity = getComplexity({
        schema,
        operationName: request.operationName,
        query: document,
        variables: request.variables,
        estimators: [
          fieldExtensionsEstimator(),
          simpleEstimator({ defaultComplexity: 1 }),
        ],
      });

      const maxComplexity = 1000;
      if (complexity > maxComplexity) {
        throw new GraphQLError(
          `Query complexity ${complexity} exceeds maximum ${maxComplexity}`,
          { extensions: { code: "QUERY_TOO_COMPLEX", complexity, maxComplexity } }
        );
      }
    },
  }),
};

// Schema with complexity hints
const typeDefs = `#graphql
  type Query {
    # @complexity(value: 1)
    user(id: ID!): User

    # @complexity(value: 5, multipliers: ["first"])
    posts(first: Int = 10): PostConnection!

    # @complexity(value: 10)
    search(query: String!): [SearchResult!]!
  }
`;

// Depth limiting
const server = new ApolloServer({
  typeDefs,
  resolvers,
  validationRules: [depthLimit(7)], // max 7 levels deep
  plugins: [complexityPlugin],
});
```

## Persisted Queries

```typescript
// Automatic Persisted Queries (APQ) — Apollo built-in
import { ApolloServer } from "@apollo/server";

const server = new ApolloServer({
  typeDefs,
  resolvers,
  // APQ is enabled by default in Apollo Server 4
  // First request sends query hash + full query
  // Subsequent requests send only the hash
});

// Client-side: Apollo Client handles APQ automatically
// import { createPersistedQueryLink } from "@apollo/client/link/persisted-queries";
// import { sha256 } from "crypto-hash";
//
// const link = createPersistedQueryLink({ sha256 });

// Manual persisted queries (allowlist)
const ALLOWED_QUERIES: Record<string, string> = {
  "get-user": `query GetUser($id: ID!) { user(id: $id) { id name email } }`,
  "list-posts": `query ListPosts($first: Int) { posts(first: $first) { edges { node { id title } } } }`,
};

// Plugin to enforce persisted queries only
const persistedQueriesPlugin = {
  requestDidStart: async () => ({
    async didResolveOperation({ request }: any) {
      if (process.env.NODE_ENV === "production") {
        const queryId = request.extensions?.persistedQuery?.id;
        if (!queryId || !ALLOWED_QUERIES[queryId]) {
          throw new GraphQLError("Only persisted queries allowed in production", {
            extensions: { code: "PERSISTED_QUERY_REQUIRED" },
          });
        }
      }
    },
  }),
};
```

## Response Caching

```typescript
import responseCachePlugin from "@apollo/server-plugin-response-cache";
import KeyvRedis from "@keyv/redis";
import { KeyvAdapter } from "@apollo/utils.keyvadapter";

// In-memory cache (default)
const server = new ApolloServer({
  typeDefs,
  resolvers,
  plugins: [
    responseCachePlugin({
      // Cache control per session
      sessionId: (ctx) => ctx.user?.id ?? null,
    }),
  ],
});

// Redis-backed cache (production)
const server = new ApolloServer({
  typeDefs,
  resolvers,
  cache: new KeyvAdapter(new KeyvRedis("redis://localhost:6379")),
  plugins: [responseCachePlugin()],
});

// Schema-level cache hints
const typeDefs = `#graphql
  type Post @cacheControl(maxAge: 60) {
    id: ID!
    title: String!
    author: User! @cacheControl(maxAge: 300)
    viewCount: Int! @cacheControl(maxAge: 0) # never cache
  }

  type Query {
    posts: [Post!]! @cacheControl(maxAge: 30)
    currentUser: User @cacheControl(maxAge: 0, scope: PRIVATE)
  }
`;

// Programmatic cache hints in resolvers
const resolvers = {
  Query: {
    posts: async (_: any, args: any, ctx: Context, info: GraphQLResolveInfo) => {
      info.cacheControl.setCacheHint({ maxAge: 60, scope: "PUBLIC" });
      return ctx.db.post.findMany({ where: { published: true } });
    },
  },
};
```

## Rate Limiting

```typescript
import { GraphQLError } from "graphql";
import { Redis } from "ioredis";

const redis = new Redis();

// Rate limit by user/IP
async function checkRateLimit(key: string, limit: number, windowSec: number): Promise<void> {
  const current = await redis.incr(key);
  if (current === 1) {
    await redis.expire(key, windowSec);
  }
  if (current > limit) {
    const ttl = await redis.ttl(key);
    throw new GraphQLError("Rate limit exceeded", {
      extensions: {
        code: "RATE_LIMITED",
        retryAfter: ttl,
        limit,
        remaining: 0,
      },
    });
  }
}

// Apollo Server plugin
const rateLimitPlugin = {
  requestDidStart: async () => ({
    async didResolveOperation({ request, contextValue }: any) {
      const key = contextValue.user?.id
        ? `gql:rate:user:${contextValue.user.id}`
        : `gql:rate:ip:${contextValue.ip}`;

      // 100 requests per minute for authenticated, 20 for anonymous
      const limit = contextValue.user ? 100 : 20;
      await checkRateLimit(key, limit, 60);
    },
  }),
};

// Field-level rate limiting
const resolvers = {
  Mutation: {
    createPost: async (_: any, args: any, ctx: Context) => {
      await checkRateLimit(`gql:create:${ctx.user!.id}`, 10, 3600); // 10 posts/hour
      return ctx.db.post.create({ data: { ...args.input, authorId: ctx.user!.id } });
    },
  },
};
```

## Apollo Federation (Microservices)

```bash
npm install @apollo/subgraph @apollo/gateway @apollo/server
```

```typescript
// Users subgraph (users-service)
import { buildSubgraphSchema } from "@apollo/subgraph";
import gql from "graphql-tag";

const typeDefs = gql`
  extend schema @link(url: "https://specs.apollo.dev/federation/v2.0", import: ["@key", "@external", "@requires"])

  type User @key(fields: "id") {
    id: ID!
    name: String!
    email: String!
  }

  type Query {
    user(id: ID!): User
    users: [User!]!
  }
`;

const resolvers = {
  User: {
    __resolveReference: async (ref: { id: string }) => {
      return db.users.findById(ref.id);
    },
  },
  Query: {
    user: (_, { id }) => db.users.findById(id),
    users: () => db.users.findAll(),
  },
};

const server = new ApolloServer({
  schema: buildSubgraphSchema({ typeDefs, resolvers }),
});

// Posts subgraph (posts-service)
const typeDefs = gql`
  extend schema @link(url: "https://specs.apollo.dev/federation/v2.0", import: ["@key", "@external"])

  type Post @key(fields: "id") {
    id: ID!
    title: String!
    content: String!
    author: User!
  }

  extend type User @key(fields: "id") {
    id: ID! @external
    posts: [Post!]!
  }

  type Query {
    post(id: ID!): Post
    posts: [Post!]!
  }
`;

// Gateway (API gateway)
import { ApolloGateway, IntrospectAndCompose } from "@apollo/gateway";

const gateway = new ApolloGateway({
  supergraphSdl: new IntrospectAndCompose({
    subgraphs: [
      { name: "users", url: "http://users-service:4001/graphql" },
      { name: "posts", url: "http://posts-service:4002/graphql" },
    ],
  }),
});

const server = new ApolloServer({ gateway });
```

## Query Optimization Patterns

```typescript
// Select only requested fields (Prisma)
import { GraphQLResolveInfo } from "graphql";
import { parseResolveInfo, simplifyParsedResolveInfoFragmentType } from "graphql-parse-resolve-info";

function getRequestedFields(info: GraphQLResolveInfo): string[] {
  const parsedInfo = parseResolveInfo(info);
  if (!parsedInfo) return [];
  const simplified = simplifyParsedResolveInfoFragmentType(parsedInfo);
  return Object.keys(simplified.fields);
}

const resolvers = {
  Query: {
    users: async (_: any, __: any, ctx: Context, info: GraphQLResolveInfo) => {
      const fields = getRequestedFields(info);
      return ctx.db.user.findMany({
        select: Object.fromEntries(fields.map((f) => [f, true])),
      });
    },
  },
};

// Batch mutations
const resolvers = {
  Mutation: {
    createPosts: async (_: any, { inputs }: { inputs: CreatePostInput[] }, ctx: Context) => {
      return ctx.db.$transaction(
        inputs.map((input) =>
          ctx.db.post.create({
            data: { ...input, authorId: ctx.user!.id },
          })
        )
      );
    },
  },
};
```

## Gotchas

1. **DataLoader batch function must return results in the same order as input keys.** If you return `[userB, userA]` for keys `[idA, idB]`, data gets swapped between users. Always build a Map and re-order.

2. **DataLoader caches per request by default.** This is correct — do NOT share DataLoader instances across requests. Create new loaders in the context factory function for each request.

3. **Query complexity estimators run before resolvers.** They use the schema and query AST, not actual data. A query requesting `posts(first: 1000)` gets high complexity even if only 3 posts exist. Design your estimators around worst-case scenarios.

4. **Response caching doesn't work with mutations.** Only queries are cacheable. Mutations always bypass the response cache. Use `@cacheControl(maxAge: 0)` on fields that depend on mutation results.

5. **Federation `__resolveReference` is called for every entity.** Without DataLoader in `__resolveReference`, you get N+1 queries across service boundaries. Batch reference resolution is critical for federation performance.

6. **Persisted queries require client-side coordination.** The server needs the query hash mapping, and the client must send hashes instead of full queries. Mismatched hash algorithms cause "PersistedQueryNotFound" errors.
