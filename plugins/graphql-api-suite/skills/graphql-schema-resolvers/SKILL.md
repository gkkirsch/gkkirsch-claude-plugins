---
name: graphql-schema-resolvers
description: >
  GraphQL schema design, type definitions, resolvers, and code-first vs
  schema-first approaches with Apollo Server, GraphQL Yoga, and Pothos.
  Triggers: "graphql schema", "graphql resolver", "graphql types",
  "code-first graphql", "schema-first", "apollo server setup",
  "graphql subscription", "graphql input types", "graphql enum".
  NOT for: REST APIs (use rest-api-design), database queries (use database-patterns).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# GraphQL Schema & Resolvers

## Setup

```bash
# Apollo Server (most popular)
npm install @apollo/server graphql

# GraphQL Yoga (lightweight alternative)
npm install graphql-yoga graphql

# Code-first with Pothos
npm install @pothos/core @pothos/plugin-prisma graphql
```

## Schema-First Approach

```typescript
// schema.graphql
import { readFileSync } from "fs";
import { ApolloServer } from "@apollo/server";
import { startStandaloneServer } from "@apollo/server/standalone";

// Type definitions (SDL)
const typeDefs = `#graphql
  enum Role {
    ADMIN
    USER
    MODERATOR
  }

  enum SortOrder {
    ASC
    DESC
  }

  type User {
    id: ID!
    email: String!
    name: String!
    role: Role!
    posts(limit: Int = 10, offset: Int = 0): [Post!]!
    postsCount: Int!
    createdAt: String!
  }

  type Post {
    id: ID!
    title: String!
    content: String!
    published: Boolean!
    author: User!
    tags: [Tag!]!
    createdAt: String!
    updatedAt: String!
  }

  type Tag {
    id: ID!
    name: String!
    posts: [Post!]!
  }

  # Pagination
  type PageInfo {
    hasNextPage: Boolean!
    hasPreviousPage: Boolean!
    startCursor: String
    endCursor: String
    totalCount: Int!
  }

  type PostEdge {
    cursor: String!
    node: Post!
  }

  type PostConnection {
    edges: [PostEdge!]!
    pageInfo: PageInfo!
  }

  # Inputs
  input CreatePostInput {
    title: String!
    content: String!
    tags: [String!]
    published: Boolean = false
  }

  input UpdatePostInput {
    title: String
    content: String
    published: Boolean
  }

  input PostFilter {
    published: Boolean
    authorId: ID
    tagName: String
    search: String
  }

  # Union & Interface
  interface Node {
    id: ID!
  }

  union SearchResult = User | Post | Tag

  # Queries
  type Query {
    user(id: ID!): User
    users(role: Role): [User!]!
    post(id: ID!): Post
    posts(
      filter: PostFilter
      first: Int
      after: String
      orderBy: SortOrder
    ): PostConnection!
    search(query: String!): [SearchResult!]!
    node(id: ID!): Node
  }

  # Mutations
  type Mutation {
    createPost(input: CreatePostInput!): Post!
    updatePost(id: ID!, input: UpdatePostInput!): Post!
    deletePost(id: ID!): Boolean!
    publishPost(id: ID!): Post!
  }

  # Subscriptions
  type Subscription {
    postPublished: Post!
    postUpdated(authorId: ID): Post!
  }
`;
```

## Resolvers

```typescript
import { PubSub, withFilter } from "graphql-subscriptions";

const pubsub = new PubSub();

const resolvers = {
  // Enum mapping
  Role: {
    ADMIN: "admin",
    USER: "user",
    MODERATOR: "moderator",
  },

  // Interface resolution
  Node: {
    __resolveType(obj: any) {
      if (obj.email) return "User";
      if (obj.title) return "Post";
      if (obj.name && !obj.email) return "Tag";
      return null;
    },
  },

  // Union resolution
  SearchResult: {
    __resolveType(obj: any) {
      if (obj.email) return "User";
      if (obj.title) return "Post";
      return "Tag";
    },
  },

  Query: {
    user: async (_: any, { id }: { id: string }, ctx: Context) => {
      return ctx.db.user.findUnique({ where: { id } });
    },

    users: async (_: any, { role }: { role?: string }, ctx: Context) => {
      return ctx.db.user.findMany({
        where: role ? { role } : undefined,
        orderBy: { createdAt: "desc" },
      });
    },

    post: async (_: any, { id }: { id: string }, ctx: Context) => {
      return ctx.db.post.findUnique({ where: { id } });
    },

    // Cursor-based pagination (Relay-style)
    posts: async (_: any, args: PostsArgs, ctx: Context) => {
      const { filter, first = 10, after, orderBy = "DESC" } = args;
      const limit = Math.min(first, 50); // cap at 50

      const where: any = {};
      if (filter?.published !== undefined) where.published = filter.published;
      if (filter?.authorId) where.authorId = filter.authorId;
      if (filter?.search) {
        where.OR = [
          { title: { contains: filter.search, mode: "insensitive" } },
          { content: { contains: filter.search, mode: "insensitive" } },
        ];
      }

      // Decode cursor
      const cursor = after ? { id: Buffer.from(after, "base64").toString() } : undefined;

      const [items, totalCount] = await Promise.all([
        ctx.db.post.findMany({
          where,
          take: limit + 1, // fetch one extra to check hasNextPage
          cursor,
          skip: cursor ? 1 : 0, // skip the cursor item
          orderBy: { createdAt: orderBy === "ASC" ? "asc" : "desc" },
        }),
        ctx.db.post.count({ where }),
      ]);

      const hasNextPage = items.length > limit;
      const edges = items.slice(0, limit).map((node) => ({
        cursor: Buffer.from(node.id).toString("base64"),
        node,
      }));

      return {
        edges,
        pageInfo: {
          hasNextPage,
          hasPreviousPage: !!after,
          startCursor: edges[0]?.cursor ?? null,
          endCursor: edges[edges.length - 1]?.cursor ?? null,
          totalCount,
        },
      };
    },

    search: async (_: any, { query }: { query: string }, ctx: Context) => {
      const [users, posts, tags] = await Promise.all([
        ctx.db.user.findMany({
          where: { name: { contains: query, mode: "insensitive" } },
          take: 5,
        }),
        ctx.db.post.findMany({
          where: { title: { contains: query, mode: "insensitive" } },
          take: 5,
        }),
        ctx.db.tag.findMany({
          where: { name: { contains: query, mode: "insensitive" } },
          take: 5,
        }),
      ]);
      return [...users, ...posts, ...tags];
    },
  },

  Mutation: {
    createPost: async (_: any, { input }: { input: CreatePostInput }, ctx: Context) => {
      requireAuth(ctx);
      const post = await ctx.db.post.create({
        data: {
          ...input,
          authorId: ctx.user!.id,
          tags: input.tags
            ? { connectOrCreate: input.tags.map((name) => ({
                where: { name },
                create: { name },
              })) }
            : undefined,
        },
        include: { tags: true, author: true },
      });
      return post;
    },

    updatePost: async (_: any, { id, input }: any, ctx: Context) => {
      requireAuth(ctx);
      await requireOwnership(ctx, id);
      const post = await ctx.db.post.update({
        where: { id },
        data: input,
        include: { author: true, tags: true },
      });
      pubsub.publish("POST_UPDATED", { postUpdated: post });
      return post;
    },

    deletePost: async (_: any, { id }: { id: string }, ctx: Context) => {
      requireAuth(ctx);
      await requireOwnership(ctx, id);
      await ctx.db.post.delete({ where: { id } });
      return true;
    },

    publishPost: async (_: any, { id }: { id: string }, ctx: Context) => {
      requireAuth(ctx);
      const post = await ctx.db.post.update({
        where: { id },
        data: { published: true },
        include: { author: true, tags: true },
      });
      pubsub.publish("POST_PUBLISHED", { postPublished: post });
      return post;
    },
  },

  Subscription: {
    postPublished: {
      subscribe: () => pubsub.asyncIterableIterator(["POST_PUBLISHED"]),
    },
    postUpdated: {
      subscribe: withFilter(
        () => pubsub.asyncIterableIterator(["POST_UPDATED"]),
        (payload, variables) => {
          if (variables.authorId) {
            return payload.postUpdated.authorId === variables.authorId;
          }
          return true;
        }
      ),
    },
  },

  // Field resolvers (runs for each User in a query)
  User: {
    posts: async (parent: User, args: { limit: number; offset: number }, ctx: Context) => {
      return ctx.db.post.findMany({
        where: { authorId: parent.id },
        take: args.limit,
        skip: args.offset,
        orderBy: { createdAt: "desc" },
      });
    },
    postsCount: async (parent: User, _: any, ctx: Context) => {
      return ctx.db.post.count({ where: { authorId: parent.id } });
    },
  },

  Post: {
    author: async (parent: Post, _: any, ctx: Context) => {
      // DataLoader handles batching (see optimization skill)
      return ctx.loaders.userLoader.load(parent.authorId);
    },
    tags: async (parent: Post, _: any, ctx: Context) => {
      return ctx.db.tag.findMany({
        where: { posts: { some: { id: parent.id } } },
      });
    },
  },
};
```

## Code-First with Pothos

```typescript
import SchemaBuilder from "@pothos/core";
import PrismaPlugin from "@pothos/plugin-prisma";
import type PrismaTypes from "@pothos/plugin-prisma/generated";
import { PrismaClient } from "@prisma/client";

const prisma = new PrismaClient();

const builder = new SchemaBuilder<{
  PrismaTypes: PrismaTypes;
  Context: { user?: { id: string; role: string } };
}>({
  plugins: [PrismaPlugin],
  prisma: { client: prisma },
});

// Object types from Prisma models
const UserType = builder.prismaObject("User", {
  fields: (t) => ({
    id: t.exposeID("id"),
    email: t.exposeString("email"),
    name: t.exposeString("name"),
    posts: t.relation("posts", {
      args: { limit: t.arg.int({ defaultValue: 10 }) },
      query: (args) => ({ take: args.limit, orderBy: { createdAt: "desc" } }),
    }),
  }),
});

const PostType = builder.prismaObject("Post", {
  fields: (t) => ({
    id: t.exposeID("id"),
    title: t.exposeString("title"),
    content: t.exposeString("content"),
    published: t.exposeBoolean("published"),
    author: t.relation("author"),
  }),
});

// Input types
const CreatePostInput = builder.inputType("CreatePostInput", {
  fields: (t) => ({
    title: t.string({ required: true }),
    content: t.string({ required: true }),
    published: t.boolean({ defaultValue: false }),
  }),
});

// Queries
builder.queryType({
  fields: (t) => ({
    user: t.prismaField({
      type: "User",
      nullable: true,
      args: { id: t.arg.id({ required: true }) },
      resolve: (query, _, args) =>
        prisma.user.findUnique({ ...query, where: { id: String(args.id) } }),
    }),
    posts: t.prismaField({
      type: ["Post"],
      args: { published: t.arg.boolean() },
      resolve: (query, _, args) =>
        prisma.post.findMany({
          ...query,
          where: args.published !== null ? { published: args.published } : undefined,
        }),
    }),
  }),
});

// Mutations
builder.mutationType({
  fields: (t) => ({
    createPost: t.prismaField({
      type: "Post",
      args: { input: t.arg({ type: CreatePostInput, required: true }) },
      resolve: (query, _, { input }, ctx) => {
        if (!ctx.user) throw new Error("Not authenticated");
        return prisma.post.create({
          ...query,
          data: { ...input, authorId: ctx.user.id },
        });
      },
    }),
  }),
});

export const schema = builder.toSchema();
```

## Server Setup

```typescript
// Apollo Server 4
import { ApolloServer } from "@apollo/server";
import { expressMiddleware } from "@apollo/server/express4";
import { ApolloServerPluginDrainHttpServer } from "@apollo/server/plugin/drainHttpServer";
import express from "express";
import http from "http";
import cors from "cors";

const app = express();
const httpServer = http.createServer(app);

const server = new ApolloServer({
  typeDefs,
  resolvers,
  plugins: [ApolloServerPluginDrainHttpServer({ httpServer })],
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

await server.start();

app.use(
  "/graphql",
  cors<cors.CorsRequest>(),
  express.json(),
  expressMiddleware(server, {
    context: async ({ req }) => {
      const token = req.headers.authorization?.replace("Bearer ", "");
      const user = token ? await verifyToken(token) : null;
      return {
        user,
        db: prisma,
        loaders: createLoaders(), // DataLoader instances per request
      };
    },
  })
);

await new Promise<void>((resolve) => httpServer.listen({ port: 4000 }, resolve));
console.log("Server ready at http://localhost:4000/graphql");
```

## Authentication & Authorization

```typescript
import { GraphQLError } from "graphql";

function requireAuth(ctx: Context) {
  if (!ctx.user) {
    throw new GraphQLError("You must be logged in", {
      extensions: { code: "UNAUTHENTICATED" },
    });
  }
}

function requireRole(ctx: Context, roles: string[]) {
  requireAuth(ctx);
  if (!roles.includes(ctx.user!.role)) {
    throw new GraphQLError("Insufficient permissions", {
      extensions: { code: "FORBIDDEN" },
    });
  }
}

async function requireOwnership(ctx: Context, postId: string) {
  const post = await ctx.db.post.findUnique({ where: { id: postId } });
  if (!post) {
    throw new GraphQLError("Post not found", {
      extensions: { code: "NOT_FOUND" },
    });
  }
  if (post.authorId !== ctx.user!.id && ctx.user!.role !== "admin") {
    throw new GraphQLError("Not authorized to modify this post", {
      extensions: { code: "FORBIDDEN" },
    });
  }
}
```

## Error Handling

```typescript
import { GraphQLError } from "graphql";

// Custom error classes
class NotFoundError extends GraphQLError {
  constructor(resource: string, id: string) {
    super(`${resource} with id ${id} not found`, {
      extensions: { code: "NOT_FOUND", resource, id },
    });
  }
}

class ValidationError extends GraphQLError {
  constructor(field: string, message: string) {
    super(`Validation error: ${message}`, {
      extensions: { code: "BAD_USER_INPUT", field },
    });
  }
}

// Input validation
function validateCreatePost(input: CreatePostInput) {
  if (input.title.length < 3) throw new ValidationError("title", "Must be at least 3 characters");
  if (input.title.length > 200) throw new ValidationError("title", "Must be under 200 characters");
  if (input.content.length < 10) throw new ValidationError("content", "Must be at least 10 characters");
}
```

## Gotchas

1. **N+1 queries kill GraphQL performance.** Every field resolver runs independently. Without DataLoader, querying 50 posts with authors = 51 SQL queries. Always use DataLoader for relation fields (see optimization skill).

2. **Subscriptions need a real PubSub in production.** The `PubSub` from `graphql-subscriptions` is in-memory only — it doesn't work across multiple server instances. Use `graphql-redis-subscriptions` or similar for production.

3. **`__resolveType` is required for interfaces and unions.** Without it, GraphQL can't determine which concrete type to use. Return the type NAME as a string, not the type object.

4. **Context is created per request.** DataLoader instances must be created per request to avoid cross-request data leaking. Never create DataLoaders at module scope.

5. **Enums need value mapping.** GraphQL enum values are strings in the schema but may map to different values in your database. Define the mapping in resolvers or use `@pothos/core` enum support.

6. **Input types can't have fields that return other types.** Inputs only support scalars, enums, and other input types. You can't nest an object type inside an input.
