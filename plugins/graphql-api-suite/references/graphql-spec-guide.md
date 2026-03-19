# GraphQL Specification Guide

Comprehensive reference for GraphQL specification concepts, directives, custom scalars, type system features, and execution semantics. Use this guide when building or reviewing GraphQL schemas and resolvers.

---

## Type System

### Object Types

The fundamental building block of a GraphQL schema. Object types define a set of named fields, each with a type.

```graphql
type User {
  id: ID!
  email: String!
  displayName: String!
  bio: String             # Nullable — can return null
  age: Int
  isActive: Boolean!
  rating: Float
  createdAt: DateTime!
  metadata: JSON
}
```

### Field Types & Nullability

```graphql
# Nullable field — can return null
field: String       # Returns String or null

# Non-nullable field — never returns null
field: String!      # Always returns String, error if null

# Nullable list of nullable items
field: [String]     # Returns null, [], ["a"], ["a", null]

# Non-nullable list of nullable items
field: [String]!    # Returns [], ["a"], ["a", null] — never null list

# Nullable list of non-nullable items
field: [String!]    # Returns null, [], ["a"] — no null items

# Non-nullable list of non-nullable items
field: [String!]!   # Returns [], ["a"] — never null list, no null items
```

**Guidelines:**
- Use `!` for fields that should always have a value (IDs, required attributes)
- Use nullable for fields that may legitimately be absent
- For lists, prefer `[Type!]!` — non-null list with non-null items
- Query return types: often nullable (null = not found)
- Mutation return types: often non-nullable (always returns a result)

### Scalar Types

Built-in scalars:

```graphql
# Int — 32-bit signed integer
# Float — IEEE 754 double-precision floating-point
# String — UTF-8 character sequence
# Boolean — true or false
# ID — Unique identifier, serialized as String
```

Common custom scalars:

```graphql
scalar DateTime      # ISO 8601: "2024-01-15T10:30:00Z"
scalar Date          # ISO 8601: "2024-01-15"
scalar Time          # ISO 8601: "10:30:00Z"
scalar BigInt        # Arbitrary-precision integer
scalar Decimal       # Arbitrary-precision decimal (financial)
scalar JSON          # Arbitrary JSON value
scalar UUID          # UUID v4: "550e8400-e29b-41d4-a716-446655440000"
scalar EmailAddress  # Validated email
scalar URL           # Validated URL
scalar PhoneNumber   # E.164 format: "+14155552671"
scalar Currency      # ISO 4217: "USD", "EUR"
scalar CountryCode   # ISO 3166-1 alpha-2: "US", "DE"
scalar Void          # No return value (for mutations)
scalar Upload        # File upload (graphql-upload)
scalar Byte          # Base64-encoded binary data
scalar HexColor      # CSS hex color: "#FF5733"
scalar Latitude      # -90 to 90
scalar Longitude     # -180 to 180
scalar PostalCode    # Postal/ZIP code
scalar RGB           # "rgb(255, 87, 51)"
scalar RGBA          # "rgba(255, 87, 51, 0.5)"
scalar Timestamp     # Unix timestamp (seconds)
```

Implementation using `graphql-scalars` library:

```typescript
import {
  DateTimeResolver,
  DateResolver,
  JSONResolver,
  BigIntResolver,
  UUIDResolver,
  EmailAddressResolver,
  URLResolver,
  PhoneNumberResolver,
  NonNegativeIntResolver,
  PositiveIntResolver,
} from 'graphql-scalars';

const resolvers = {
  DateTime: DateTimeResolver,
  Date: DateResolver,
  JSON: JSONResolver,
  BigInt: BigIntResolver,
  UUID: UUIDResolver,
  EmailAddress: EmailAddressResolver,
  URL: URLResolver,
  PhoneNumber: PhoneNumberResolver,
  NonNegativeInt: NonNegativeIntResolver,
  PositiveInt: PositiveIntResolver,
};
```

### Enum Types

```graphql
# Enums define a fixed set of allowed values
enum Role {
  ADMIN
  MODERATOR
  USER
  GUEST
}

enum PostStatus {
  DRAFT
  REVIEW
  PUBLISHED
  ARCHIVED
  DELETED
}

enum SortDirection {
  ASC
  DESC
}

# Deprecated enum values
enum OrderStatus {
  PENDING
  PROCESSING
  SHIPPED
  DELIVERED
  CANCELLED
  RETURNED
  REFUND_PENDING @deprecated(reason: "Use RETURNED. Removal: 2025-06-01.")
}
```

Enum resolver mapping (when DB values differ from GraphQL values):

```typescript
const resolvers = {
  Role: {
    ADMIN: 'admin',
    MODERATOR: 'moderator',
    USER: 'user',
    GUEST: 'guest',
  },
};
```

### Input Types

Input types are used for mutation arguments and complex query parameters.

```graphql
# Input types can only contain scalar fields, enum fields, and other input types
# They CANNOT contain object types, interfaces, or unions

input CreatePostInput {
  title: String!
  content: String!
  status: PostStatus = DRAFT    # Default value
  tags: [String!]               # Optional list
  categoryId: ID
  publishAt: DateTime
}

input UpdatePostInput {
  title: String                 # All fields optional for partial updates
  content: String
  status: PostStatus
  tags: [String!]
}

# Nested input types
input AddressInput {
  street: String!
  city: String!
  state: String!
  postalCode: String!
  country: CountryCode!
}

input CreateUserInput {
  email: EmailAddress!
  displayName: String!
  password: String!
  address: AddressInput         # Nested input
}

# Filter input types
input StringFilter {
  equals: String
  contains: String
  startsWith: String
  endsWith: String
  in: [String!]
  notIn: [String!]
  mode: FilterMode
}

input DateTimeFilter {
  equals: DateTime
  gt: DateTime
  gte: DateTime
  lt: DateTime
  lte: DateTime
}

enum FilterMode {
  DEFAULT
  INSENSITIVE
}

input UserFilter {
  email: StringFilter
  displayName: StringFilter
  role: Role
  createdAt: DateTimeFilter
  AND: [UserFilter!]
  OR: [UserFilter!]
  NOT: UserFilter
}
```

### Interfaces

Interfaces define a set of fields that implementing types must include.

```graphql
# Abstract type that other types implement
interface Node {
  id: ID!
}

interface Timestamped {
  createdAt: DateTime!
  updatedAt: DateTime!
}

interface Auditable {
  createdBy: User!
  updatedBy: User
  version: Int!
}

# Types implement interfaces
type User implements Node & Timestamped {
  id: ID!
  email: String!
  displayName: String!
  createdAt: DateTime!
  updatedAt: DateTime!
}

type Post implements Node & Timestamped & Auditable {
  id: ID!
  title: String!
  content: String!
  author: User!
  createdAt: DateTime!
  updatedAt: DateTime!
  createdBy: User!
  updatedBy: User
  version: Int!
}

# Query by interface
type Query {
  node(id: ID!): Node                      # Returns any Node type
  recentActivity(first: Int): [Timestamped!]!  # Returns any Timestamped type
}
```

Interface resolver — must resolve which concrete type to return:

```typescript
const resolvers = {
  Node: {
    __resolveType(obj: any) {
      // Determine concrete type based on object shape
      if (obj.email) return 'User';
      if (obj.title) return 'Post';
      if (obj.body) return 'Comment';
      return null; // GraphQL will throw an error
    },
  },
  Timestamped: {
    __resolveType(obj: any) {
      if (obj.email) return 'User';
      if (obj.title) return 'Post';
      if (obj.body) return 'Comment';
      return null;
    },
  },
};
```

### Union Types

Unions represent a value that could be one of several object types. Unlike interfaces, union members don't need to share fields.

```graphql
# Union of unrelated types
union SearchResult = User | Post | Comment | Tag

union NotificationTarget = Post | Comment | User

union MediaContent = Image | Video | Audio | Document

type Image {
  url: URL!
  width: Int!
  height: Int!
  alt: String
}

type Video {
  url: URL!
  duration: Int!
  thumbnail: URL
}

type Audio {
  url: URL!
  duration: Int!
  format: String!
}

type Document {
  url: URL!
  filename: String!
  size: Int!
  mimeType: String!
}

# Query returning union
type Query {
  search(query: String!, types: [SearchableType!]): [SearchResult!]!
  media(id: ID!): MediaContent
}

# Querying unions — use inline fragments
# query {
#   search(query: "graphql") {
#     ... on User { id, displayName }
#     ... on Post { id, title }
#     ... on Comment { id, body }
#   }
# }
```

Union resolver:

```typescript
const resolvers = {
  SearchResult: {
    __resolveType(obj: any) {
      if ('email' in obj) return 'User';
      if ('title' in obj) return 'Post';
      if ('body' in obj) return 'Comment';
      if ('name' in obj && !('email' in obj)) return 'Tag';
      return null;
    },
  },
  MediaContent: {
    __resolveType(obj: any) {
      if ('width' in obj) return 'Image';
      if ('duration' in obj && 'thumbnail' in obj) return 'Video';
      if ('duration' in obj) return 'Audio';
      if ('mimeType' in obj) return 'Document';
      return null;
    },
  },
};
```

---

## Directives

### Built-in Directives

```graphql
# @deprecated — marks a field or enum value as deprecated
type User {
  name: String! @deprecated(reason: "Use 'displayName' instead")
  displayName: String!
}

# @skip — conditionally skip a field based on a variable
query GetUser($skipEmail: Boolean!) {
  user(id: "1") {
    id
    email @skip(if: $skipEmail)
    displayName
  }
}

# @include — conditionally include a field based on a variable
query GetUser($includeEmail: Boolean!) {
  user(id: "1") {
    id
    email @include(if: $includeEmail)
    displayName
  }
}

# @specifiedBy — specifies a URL for a custom scalar's specification
scalar DateTime @specifiedBy(url: "https://scalars.graphql.org/andimarek/date-time")
```

### Custom Directive Definitions

```graphql
# Schema directives (processed at schema build time)
directive @auth(requires: Role = USER) on FIELD_DEFINITION | OBJECT
directive @cacheControl(maxAge: Int, scope: CacheControlScope) on FIELD_DEFINITION | OBJECT
directive @deprecated(reason: String = "No longer supported") on FIELD_DEFINITION | ENUM_VALUE | ARGUMENT_DEFINITION | INPUT_FIELD_DEFINITION
directive @rateLimit(limit: Int!, duration: Int!) on FIELD_DEFINITION
directive @complexity(value: Int!, multipliers: [String!]) on FIELD_DEFINITION
directive @computed on FIELD_DEFINITION
directive @external on FIELD_DEFINITION  # Federation
directive @key(fields: String!) on OBJECT | INTERFACE  # Federation
directive @provides(fields: String!) on FIELD_DEFINITION  # Federation
directive @requires(fields: String!) on FIELD_DEFINITION  # Federation
directive @shareable on FIELD_DEFINITION | OBJECT  # Federation
directive @inaccessible on FIELD_DEFINITION | OBJECT | INTERFACE | UNION | ENUM | ENUM_VALUE | SCALAR | INPUT_OBJECT | INPUT_FIELD_DEFINITION | ARGUMENT_DEFINITION  # Federation
directive @override(from: String!) on FIELD_DEFINITION  # Federation
directive @tag(name: String!) repeatable on FIELD_DEFINITION | OBJECT | INTERFACE | UNION | ENUM | ENUM_VALUE | SCALAR | INPUT_OBJECT | INPUT_FIELD_DEFINITION | ARGUMENT_DEFINITION

# Executable directives (processed at query execution time)
directive @log(level: LogLevel = INFO) on FIELD
directive @timeout(ms: Int!) on FIELD
directive @transform(operation: TransformOp!) on FIELD

enum CacheControlScope {
  PUBLIC
  PRIVATE
}

enum LogLevel {
  DEBUG
  INFO
  WARN
  ERROR
}

enum TransformOp {
  UPPERCASE
  LOWERCASE
  TRIM
}
```

### Implementing Custom Directives

```typescript
// Using @graphql-tools/utils
import { mapSchema, getDirective, MapperKind } from '@graphql-tools/utils';
import { defaultFieldResolver, GraphQLSchema } from 'graphql';

// @auth directive
function authDirectiveTransformer(schema: GraphQLSchema): GraphQLSchema {
  return mapSchema(schema, {
    [MapperKind.OBJECT_FIELD]: (fieldConfig) => {
      const authDirective = getDirective(schema, fieldConfig, 'auth')?.[0];
      if (!authDirective) return fieldConfig;

      const { requires } = authDirective;
      const originalResolve = fieldConfig.resolve || defaultFieldResolver;

      fieldConfig.resolve = async function (source, args, context, info) {
        if (!context.currentUser) {
          throw new AuthenticationError('Authentication required');
        }
        if (requires && context.currentUser.role !== requires) {
          throw new ForbiddenError(`Requires ${requires} role`);
        }
        return originalResolve(source, args, context, info);
      };

      return fieldConfig;
    },
  });
}

// @upper directive (transforms string fields to uppercase)
function upperDirectiveTransformer(schema: GraphQLSchema): GraphQLSchema {
  return mapSchema(schema, {
    [MapperKind.OBJECT_FIELD]: (fieldConfig) => {
      const upperDirective = getDirective(schema, fieldConfig, 'upper')?.[0];
      if (!upperDirective) return fieldConfig;

      const originalResolve = fieldConfig.resolve || defaultFieldResolver;
      fieldConfig.resolve = async function (source, args, context, info) {
        const result = await originalResolve(source, args, context, info);
        if (typeof result === 'string') {
          return result.toUpperCase();
        }
        return result;
      };

      return fieldConfig;
    },
  });
}

// @rateLimit directive
function rateLimitDirectiveTransformer(schema: GraphQLSchema): GraphQLSchema {
  const limitMap = new Map<string, Map<string, number[]>>();

  return mapSchema(schema, {
    [MapperKind.OBJECT_FIELD]: (fieldConfig, fieldName, typeName) => {
      const rateLimitDirective = getDirective(schema, fieldConfig, 'rateLimit')?.[0];
      if (!rateLimitDirective) return fieldConfig;

      const { limit, duration } = rateLimitDirective;
      const originalResolve = fieldConfig.resolve || defaultFieldResolver;
      const fieldKey = `${typeName}.${fieldName}`;

      fieldConfig.resolve = async function (source, args, context, info) {
        const userId = context.currentUser?.id || context.req?.ip || 'anonymous';
        const key = `${fieldKey}:${userId}`;

        if (!limitMap.has(key)) {
          limitMap.set(key, new Map());
        }

        const timestamps = limitMap.get(key)!;
        const now = Date.now();
        const windowStart = now - duration * 1000;

        // Clean old entries
        const recentCalls = [...(timestamps.get(fieldKey) || [])].filter(t => t > windowStart);

        if (recentCalls.length >= limit) {
          throw new GraphQLError('Rate limit exceeded', {
            extensions: {
              code: 'RATE_LIMITED',
              retryAfter: Math.ceil((recentCalls[0] + duration * 1000 - now) / 1000),
            },
          });
        }

        recentCalls.push(now);
        timestamps.set(fieldKey, recentCalls);

        return originalResolve(source, args, context, info);
      };

      return fieldConfig;
    },
  });
}

// Apply directive transformers to schema
let schema = makeExecutableSchema({ typeDefs, resolvers });
schema = authDirectiveTransformer(schema);
schema = upperDirectiveTransformer(schema);
schema = rateLimitDirectiveTransformer(schema);
```

---

## Query & Mutation Design

### Query Patterns

```graphql
type Query {
  # Single resource by ID
  user(id: ID!): User
  post(id: ID!): Post

  # Single resource by unique field
  userByEmail(email: EmailAddress!): User

  # Collection with pagination
  users(
    first: Int          # Forward pagination: number of items
    after: String       # Forward pagination: cursor
    last: Int           # Backward pagination: number of items
    before: String      # Backward pagination: cursor
    filter: UserFilter  # Filtering
    orderBy: UserOrderBy  # Sorting
  ): UserConnection!

  # Search
  search(
    query: String!
    types: [SearchableType!]
    first: Int
    after: String
  ): SearchResultConnection!

  # Aggregations
  userStats: UserStats!
  postCountByStatus: [StatusCount!]!

  # Current user (auth required)
  me: User
  myPosts(first: Int, after: String): PostConnection!
  myNotifications(unreadOnly: Boolean): [Notification!]!
}
```

### Mutation Patterns

```graphql
type Mutation {
  # Create — returns payload with resource and possible errors
  createUser(input: CreateUserInput!): CreateUserPayload!
  createPost(input: CreatePostInput!): CreatePostPayload!

  # Update — returns payload with updated resource
  updateUser(id: ID!, input: UpdateUserInput!): UpdateUserPayload!
  updatePost(id: ID!, input: UpdatePostInput!): UpdatePostPayload!

  # Delete — returns success indicator
  deleteUser(id: ID!): DeleteResult!
  deletePost(id: ID!): DeleteResult!

  # Actions — domain-specific operations
  publishPost(id: ID!): PublishPostPayload!
  archivePost(id: ID!): ArchivePostPayload!
  login(email: String!, password: String!): AuthPayload!
  logout: LogoutPayload!
  resetPassword(email: String!): ResetPasswordPayload!
  changePassword(oldPassword: String!, newPassword: String!): ChangePasswordPayload!

  # Batch operations
  batchUpdatePosts(ids: [ID!]!, input: BatchUpdatePostInput!): BatchUpdateResult!
  batchDeletePosts(ids: [ID!]!): BatchDeleteResult!

  # File upload
  uploadAvatar(file: Upload!): UploadResult!
}

# Mutation payloads follow a consistent pattern
type CreateUserPayload {
  user: User
  errors: [UserError!]!
}

type UserError {
  field: String
  message: String!
  code: ErrorCode!
}

type DeleteResult {
  success: Boolean!
  message: String
}
```

### Subscription Patterns

```graphql
type Subscription {
  # Entity-level subscriptions
  postCreated: Post!
  postUpdated(id: ID): Post!
  postDeleted: DeleteEvent!

  # Filtered subscriptions
  commentAdded(postId: ID!): Comment!
  notificationReceived(userId: ID!): Notification!

  # Presence
  userPresenceChanged: PresenceEvent!

  # Real-time feeds
  liveFeed(topic: String!): FeedItem!
}

type DeleteEvent {
  id: ID!
  deletedAt: DateTime!
}

type PresenceEvent {
  userId: ID!
  status: PresenceStatus!
  lastSeen: DateTime
}

enum PresenceStatus {
  ONLINE
  AWAY
  OFFLINE
}
```

---

## Execution Semantics

### Resolver Execution Order

```
Query: {
  user(id: "1") {           # 1. Root resolver: Query.user
    id                       # 2. Field resolver: User.id (default: parent.id)
    displayName              # 3. Field resolver: User.displayName (default: parent.displayName)
    posts(first: 5) {       # 4. Field resolver: User.posts (custom resolver)
      id                     # 5. Field resolver: Post.id (for each post)
      title                  # 6. Field resolver: Post.title (for each post)
      author {               # 7. Field resolver: Post.author (for each post)
        displayName          # 8. Field resolver: User.displayName (for each author)
      }
    }
  }
}

# Execution flow:
# 1. Query.user(id: "1") → returns User object
# 2-3. Default resolvers read fields from User object
# 4. User.posts(first: 5) → returns [Post] array
# 5-8. For each Post, resolve fields (parallel within each level)
```

### Default Field Resolution

```typescript
// GraphQL's default resolver behavior
function defaultFieldResolver(source, args, context, info) {
  // 1. Check if source has a property matching the field name
  const property = source[info.fieldName];

  // 2. If it's a function, call it with args
  if (typeof property === 'function') {
    return property(args, context, info);
  }

  // 3. Otherwise, return the property value
  return property;
}

// This means for simple cases, you don't need explicit resolvers:
const user = {
  id: '1',
  email: 'test@example.com',
  displayName: 'Test User',
  // These fields resolve automatically via default resolver
};
```

### Error Handling in Execution

```graphql
# Partial success — some fields can error while others succeed
# Query:
{
  user(id: "1") {
    id
    email
    dangerousField    # This resolver throws
    safeField         # This still resolves
  }
}

# Response:
{
  "data": {
    "user": {
      "id": "1",
      "email": "test@example.com",
      "dangerousField": null,    # null due to error
      "safeField": "works!"
    }
  },
  "errors": [
    {
      "message": "Something went wrong",
      "locations": [{ "line": 5, "column": 5 }],
      "path": ["user", "dangerousField"],
      "extensions": {
        "code": "INTERNAL_SERVER_ERROR"
      }
    }
  ]
}

# Non-nullable field error bubbles up to nearest nullable parent
# If dangerousField is String! (non-nullable) and errors:
# → user becomes null (because it can't have a non-null field be null)
# → If user is User! (non-nullable), error bubbles to data
# → data becomes null if the root field is non-nullable
```

---

## Introspection

```graphql
# Query the schema itself
{
  __schema {
    types {
      name
      kind
      description
    }
    queryType { name }
    mutationType { name }
    subscriptionType { name }
    directives {
      name
      description
      locations
      args { name type { name } }
    }
  }
}

# Query a specific type
{
  __type(name: "User") {
    name
    kind
    description
    fields {
      name
      description
      type {
        name
        kind
        ofType { name kind }
      }
      args {
        name
        type { name kind }
        defaultValue
      }
      isDeprecated
      deprecationReason
    }
    interfaces { name }
    possibleTypes { name }  # For interfaces and unions
    enumValues { name description isDeprecated deprecationReason }
    inputFields { name type { name kind } defaultValue }
  }
}
```

Disabling introspection in production:

```typescript
import { NoSchemaIntrospectionCustomRule } from 'graphql';

const server = new ApolloServer({
  schema,
  validationRules: process.env.NODE_ENV === 'production'
    ? [NoSchemaIntrospectionCustomRule]
    : [],
});
```

---

## Schema Design Patterns Reference

### Relay Global Object Identification

```graphql
interface Node {
  id: ID!  # Globally unique, opaque ID
}

type Query {
  node(id: ID!): Node
  nodes(ids: [ID!]!): [Node]!
}
```

### Relay Input Object Mutations

```graphql
# Input always named *Input, contains a clientMutationId
input CreatePostInput {
  clientMutationId: String
  title: String!
  content: String!
}

# Payload always named *Payload, returns clientMutationId
type CreatePostPayload {
  clientMutationId: String
  post: Post
  errors: [UserError!]!
}
```

### Connection Pattern (Relay Pagination)

```graphql
type UserConnection {
  edges: [UserEdge!]!
  pageInfo: PageInfo!
  totalCount: Int
}

type UserEdge {
  node: User!
  cursor: String!   # Opaque cursor for this edge
}

type PageInfo {
  hasNextPage: Boolean!
  hasPreviousPage: Boolean!
  startCursor: String
  endCursor: String
}
```

### Polymorphic Result Types

```graphql
# Instead of throwing errors, return typed results
type Mutation {
  login(input: LoginInput!): LoginResult!
}

union LoginResult =
  | LoginSuccess
  | InvalidCredentialsError
  | AccountLockedError
  | TwoFactorRequiredError

type LoginSuccess {
  token: String!
  user: User!
  expiresAt: DateTime!
}

type InvalidCredentialsError {
  message: String!
}

type AccountLockedError {
  message: String!
  lockedUntil: DateTime
}

type TwoFactorRequiredError {
  message: String!
  challengeToken: String!
  methods: [TwoFactorMethod!]!
}
```

---

## GraphQL over HTTP

### Standard Request Format

```
POST /graphql HTTP/1.1
Content-Type: application/json

{
  "query": "query GetUser($id: ID!) { user(id: $id) { id email } }",
  "variables": { "id": "123" },
  "operationName": "GetUser"
}
```

### GET Request Format (for queries only)

```
GET /graphql?query={user(id:"123"){id,email}}&operationName=GetUser HTTP/1.1
```

### Batch Requests

```
POST /graphql HTTP/1.1
Content-Type: application/json

[
  { "query": "{ user(id: \"1\") { id } }" },
  { "query": "{ post(id: \"1\") { id } }" }
]
```

### Multipart Request (File Upload)

```
POST /graphql HTTP/1.1
Content-Type: multipart/form-data; boundary=----Boundary

------Boundary
Content-Disposition: form-data; name="operations"

{"query":"mutation($file: Upload!) { uploadFile(file: $file) { url } }","variables":{"file":null}}
------Boundary
Content-Disposition: form-data; name="map"

{"0":["variables.file"]}
------Boundary
Content-Disposition: form-data; name="0"; filename="photo.jpg"
Content-Type: image/jpeg

<binary data>
------Boundary--
```

### Response Format

```json
// Success
{
  "data": { "user": { "id": "1", "email": "test@example.com" } }
}

// Partial error
{
  "data": { "user": { "id": "1", "email": null } },
  "errors": [
    {
      "message": "Not authorized to view email",
      "locations": [{ "line": 1, "column": 30 }],
      "path": ["user", "email"],
      "extensions": { "code": "FORBIDDEN" }
    }
  ]
}

// Full error
{
  "errors": [
    {
      "message": "User not found",
      "locations": [{ "line": 1, "column": 3 }],
      "path": ["user"],
      "extensions": { "code": "NOT_FOUND" }
    }
  ],
  "data": null
}
```

---

## Server Libraries Comparison

| Feature | Apollo Server | GraphQL Yoga | Mercurius | Express-GraphQL |
|---------|--------------|--------------|-----------|-----------------|
| Runtime | Node.js | Node.js | Fastify | Express |
| Subscriptions | WebSocket | WebSocket/SSE | WebSocket | No |
| Federation | Yes (gateway) | Manual | Yes | No |
| Persisted Queries | APQ built-in | Plugin | Built-in | No |
| File Uploads | Plugin | Built-in | Plugin | No |
| Response Caching | Plugin | Plugin | Built-in | No |
| TypeScript | First-class | First-class | First-class | Basic |
| Plugins | Lifecycle hooks | Envelop plugins | Hooks | Middleware |
| Performance | Good | Very Good | Excellent | Good |
| Bundle Size | Large | Medium | Small | Small |
| Maintenance | Active | Active | Active | Minimal |

---

## Quick Reference

### Naming Conventions
- **Types**: PascalCase (`User`, `BlogPost`)
- **Fields**: camelCase (`displayName`, `createdAt`)
- **Arguments**: camelCase (`userId`, `orderBy`)
- **Enums**: PascalCase type, SCREAMING_SNAKE_CASE values (`PostStatus.PUBLISHED`)
- **Input types**: PascalCase + `Input` suffix (`CreateUserInput`)
- **Payload types**: PascalCase + `Payload` suffix (`CreateUserPayload`)
- **Connections**: PascalCase + `Connection` suffix (`UserConnection`)
- **Edges**: PascalCase + `Edge` suffix (`UserEdge`)

### Type Suffixes Convention
- `*Input` — Input types for mutations
- `*Payload` — Mutation return types
- `*Connection` — Paginated list types (Relay)
- `*Edge` — Edge types in connections
- `*Filter` — Filter input types
- `*OrderBy` — Sorting input types
- `*Error` — Error types in payloads
- `*Result` — Union result types for mutations
