# Schema Federation & Stitching Reference

Comprehensive reference for Apollo Federation v2, schema stitching, distributed GraphQL architectures, and supergraph composition. Use this guide when designing or implementing federated GraphQL APIs.

---

## Apollo Federation v2

### Core Concepts

**Supergraph**: The combined schema from all subgraphs, served by the gateway/router.
**Subgraph**: An individual GraphQL service that owns part of the schema.
**Gateway/Router**: The entry point that receives client queries and routes them to subgraphs.
**Entity**: A type that can be referenced and extended across subgraphs.

### Federation Directives

```graphql
# Import Federation v2 directives
extend schema @link(
  url: "https://specs.apollo.dev/federation/v2.3"
  import: [
    "@key",
    "@shareable",
    "@external",
    "@provides",
    "@requires",
    "@override",
    "@inaccessible",
    "@tag",
    "@composeDirective",
    "@interfaceObject"
  ]
)
```

### @key — Entity Identity

Defines the primary key for an entity type. Enables cross-subgraph references.

```graphql
# Single key field
type User @key(fields: "id") {
  id: ID!
  email: String!
  displayName: String!
}

# Composite key
type CartItem @key(fields: "userId productId") {
  userId: ID!
  productId: ID!
  quantity: Int!
}

# Multiple keys (either can be used to resolve)
type Product @key(fields: "id") @key(fields: "sku") {
  id: ID!
  sku: String!
  name: String!
  price: Float!
}

# Nested key fields
type Review @key(fields: "author { id } product { id }") {
  author: User!
  product: Product!
  rating: Int!
  body: String!
}

# Non-resolvable key (reference only, can't be resolved by this subgraph)
type User @key(fields: "id", resolvable: false) {
  id: ID!
}
```

### @external — Reference External Fields

Marks a field as owned by another subgraph. Used with @requires and @provides.

```graphql
# In reviews subgraph: reference user's email from users subgraph
type User @key(fields: "id") {
  id: ID!
  email: String! @external  # Owned by users subgraph
}
```

### @requires — Computed Fields

Declares that a field needs external fields to compute its value.

```graphql
# In shipping subgraph: compute shipping cost based on product weight
type Product @key(fields: "id") {
  id: ID!
  weight: Float! @external      # Owned by products subgraph
  size: String! @external       # Owned by products subgraph
  shippingCost: Float! @requires(fields: "weight size")  # Computed here
}
```

Resolver:

```typescript
const resolvers = {
  Product: {
    __resolveReference: async (ref: { id: string; weight: number; size: string }) => {
      // weight and size are provided by the gateway because of @requires
      return ref;
    },
    shippingCost: (product: { weight: number; size: string }) => {
      // Calculate shipping based on weight and size
      return calculateShipping(product.weight, product.size);
    },
  },
};
```

### @provides — Optimization Hint

Declares that a resolver provides certain fields of a referenced entity, avoiding an extra subgraph call.

```graphql
# In reviews subgraph: when fetching review.author, this subgraph
# already has the author's username, no need to call users subgraph
type Review @key(fields: "id") {
  id: ID!
  body: String!
  author: User! @provides(fields: "username")
}

type User @key(fields: "id") {
  id: ID!
  username: String! @external
}
```

### @shareable — Shared Field Resolution

Allows multiple subgraphs to resolve the same field.

```graphql
# Both products and inventory subgraphs can resolve Product.name
# In products subgraph:
type Product @key(fields: "id") {
  id: ID!
  name: String! @shareable
  price: Float!
}

# In inventory subgraph:
type Product @key(fields: "id") {
  id: ID!
  name: String! @shareable  # Can also resolve name
  inStock: Boolean!
}
```

### @override — Migrate Fields

Gradually migrate field resolution from one subgraph to another.

```graphql
# In new-users subgraph: take over the "email" field from legacy-users
type User @key(fields: "id") {
  id: ID!
  email: String! @override(from: "legacy-users")
}

# Progressive override (percentage-based)
type User @key(fields: "id") {
  id: ID!
  email: String! @override(from: "legacy-users", percent: 50)  # 50% of traffic
}
```

### @inaccessible — Hide from Supergraph

Marks a field as visible within the subgraph but hidden from the public supergraph API.

```graphql
type Product @key(fields: "id") {
  id: ID!
  name: String!
  price: Float!
  internalCost: Float! @inaccessible  # Not exposed to clients
  supplierCode: String! @inaccessible
}
```

### @tag — Metadata Annotation

Adds metadata tags for tooling and governance.

```graphql
type User @key(fields: "id") @tag(name: "pii") {
  id: ID!
  email: String! @tag(name: "pii")
  displayName: String!
  ssn: String! @tag(name: "sensitive") @inaccessible
}
```

---

## Subgraph Architecture Patterns

### Domain-Driven Subgraphs

```
                    ┌─────────────┐
                    │   Gateway    │
                    │  (Router)   │
                    └──────┬──────┘
           ┌───────────────┼───────────────┐
           │               │               │
    ┌──────┴──────┐ ┌──────┴──────┐ ┌──────┴──────┐
    │   Users     │ │   Posts     │ │  Commerce   │
    │  Subgraph   │ │  Subgraph   │ │  Subgraph   │
    └─────────────┘ └─────────────┘ └─────────────┘
    - User profiles  - Blog posts    - Products
    - Authentication - Comments      - Orders
    - Roles/perms    - Tags          - Payments
    - Preferences    - Categories    - Inventory
```

### Example: E-Commerce Federation

Users subgraph:

```graphql
extend schema @link(url: "https://specs.apollo.dev/federation/v2.3",
  import: ["@key"])

type User @key(fields: "id") {
  id: ID!
  email: String!
  displayName: String!
  role: Role!
  addresses: [Address!]!
  createdAt: DateTime!
}

type Address {
  id: ID!
  street: String!
  city: String!
  state: String!
  zipCode: String!
  country: String!
  isDefault: Boolean!
}

enum Role {
  CUSTOMER
  ADMIN
  SUPPORT
}

type Query {
  me: User
  user(id: ID!): User
}

type Mutation {
  register(input: RegisterInput!): AuthPayload!
  login(email: String!, password: String!): AuthPayload!
  updateProfile(input: UpdateProfileInput!): User!
}
```

Products subgraph:

```graphql
extend schema @link(url: "https://specs.apollo.dev/federation/v2.3",
  import: ["@key", "@shareable"])

type Product @key(fields: "id") @key(fields: "sku") {
  id: ID!
  sku: String!
  name: String! @shareable
  description: String!
  price: Float!
  compareAtPrice: Float
  images: [String!]!
  category: Category!
  variants: [ProductVariant!]!
  createdAt: DateTime!
}

type ProductVariant {
  id: ID!
  name: String!
  sku: String!
  price: Float!
  attributes: JSON
}

type Category @key(fields: "id") {
  id: ID!
  name: String!
  slug: String!
  products(first: Int, after: String): ProductConnection!
}

type Query {
  product(id: ID!): Product
  productBySku(sku: String!): Product
  products(
    first: Int
    after: String
    filter: ProductFilter
    sort: ProductSort
  ): ProductConnection!
  categories: [Category!]!
}
```

Orders subgraph:

```graphql
extend schema @link(url: "https://specs.apollo.dev/federation/v2.3",
  import: ["@key", "@external", "@requires"])

type Order @key(fields: "id") {
  id: ID!
  customer: User!
  items: [OrderItem!]!
  status: OrderStatus!
  subtotal: Float!
  tax: Float!
  shipping: Float!
  total: Float!
  shippingAddress: Address!
  createdAt: DateTime!
  updatedAt: DateTime!
}

type OrderItem {
  product: Product!
  variant: ProductVariant
  quantity: Int!
  unitPrice: Float!
  total: Float!
}

# Extend User from users subgraph
type User @key(fields: "id") {
  id: ID!
  orders(first: Int, after: String, status: OrderStatus): OrderConnection!
  orderCount: Int!
}

# Extend Product from products subgraph
type Product @key(fields: "id") {
  id: ID!
  name: String! @external
  price: Float! @external
  totalSold: Int!
}

enum OrderStatus {
  PENDING
  CONFIRMED
  PROCESSING
  SHIPPED
  DELIVERED
  CANCELLED
  RETURNED
}

type Query {
  order(id: ID!): Order
  orders(first: Int, after: String, filter: OrderFilter): OrderConnection!
}

type Mutation {
  createOrder(input: CreateOrderInput!): Order!
  cancelOrder(id: ID!): Order!
  updateOrderStatus(id: ID!, status: OrderStatus!): Order!
}
```

Inventory subgraph:

```graphql
extend schema @link(url: "https://specs.apollo.dev/federation/v2.3",
  import: ["@key", "@shareable"])

type Product @key(fields: "id") {
  id: ID!
  name: String! @shareable
  inStock: Boolean!
  stockLevel: Int!
  reservedQuantity: Int!
  availableQuantity: Int!
  lowStockThreshold: Int!
  isLowStock: Boolean!
}

type Query {
  inventoryStatus(productId: ID!): Product
}

type Mutation {
  reserveStock(productId: ID!, quantity: Int!): ReserveResult!
  releaseStock(productId: ID!, quantity: Int!): ReleaseResult!
  adjustStock(productId: ID!, adjustment: Int!, reason: String!): Product!
}

type Subscription {
  stockLevelChanged(productId: ID): StockChangeEvent!
}

type StockChangeEvent {
  productId: ID!
  previousLevel: Int!
  newLevel: Int!
  reason: String!
  timestamp: DateTime!
}
```

---

## Entity Resolution

### __resolveReference

Every entity type with `@key` must implement a `__resolveReference` resolver in the subgraph that defines it.

```typescript
// users subgraph
const resolvers = {
  User: {
    __resolveReference: async (reference: { id: string }, context) => {
      // Fetch the full user from this subgraph's database
      return context.prisma.user.findUnique({
        where: { id: reference.id },
      });
    },
  },
};

// products subgraph — supports two keys
const resolvers = {
  Product: {
    __resolveReference: async (reference: { id?: string; sku?: string }, context) => {
      if (reference.id) {
        return context.prisma.product.findUnique({ where: { id: reference.id } });
      }
      if (reference.sku) {
        return context.prisma.product.findUnique({ where: { sku: reference.sku } });
      }
      return null;
    },
  },
};
```

### Batched Entity Resolution

Use DataLoader for efficient batching when the gateway sends multiple entity references.

```typescript
const resolvers = {
  User: {
    __resolveReference: async (reference: { id: string }, context) => {
      // Uses DataLoader to batch multiple user lookups
      return context.loaders.userById.load(reference.id);
    },
  },
};
```

### Returning Entity References

When a subgraph doesn't own all fields of an entity, return a reference that the gateway will resolve.

```typescript
// In orders subgraph: return a User reference
const resolvers = {
  Order: {
    customer: (order) => {
      // Return a stub — the gateway will resolve the rest from users subgraph
      return { __typename: 'User', id: order.customerId };
    },
  },
};
```

---

## Gateway / Router Configuration

### Apollo Router (Recommended for Production)

```yaml
# router.yaml
supergraph:
  introspection: true
  listen: 0.0.0.0:4000

# CORS configuration
cors:
  origins:
    - https://myapp.com
    - https://studio.apollographql.com
  allow_headers:
    - Content-Type
    - Authorization
    - X-Request-ID

# Authentication
authentication:
  router:
    jwt:
      jwks:
        - url: https://auth.example.com/.well-known/jwks.json

# Rate limiting
limits:
  max_depth: 15
  max_height: 200
  max_aliases: 30
  max_root_fields: 20

# Caching
preview_entity_cache:
  enabled: true
  subgraph:
    all:
      enabled: true
      ttl: 60s

# Headers propagation
headers:
  all:
    request:
      - propagate:
          named: Authorization
      - propagate:
          named: X-Request-ID

# Telemetry
telemetry:
  exporters:
    tracing:
      otlp:
        enabled: true
        endpoint: http://otel-collector:4317
```

### Node.js Gateway (Development/Simple Use Cases)

```typescript
import { ApolloGateway, IntrospectAndCompose } from '@apollo/gateway';
import { ApolloServer } from '@apollo/server';
import { startStandaloneServer } from '@apollo/server/standalone';

const gateway = new ApolloGateway({
  supergraphSdl: new IntrospectAndCompose({
    subgraphs: [
      { name: 'users', url: process.env.USERS_SERVICE_URL },
      { name: 'products', url: process.env.PRODUCTS_SERVICE_URL },
      { name: 'orders', url: process.env.ORDERS_SERVICE_URL },
      { name: 'inventory', url: process.env.INVENTORY_SERVICE_URL },
    ],
    pollIntervalInMs: 10000,
  }),
});

const server = new ApolloServer({ gateway });

const { url } = await startStandaloneServer(server, {
  listen: { port: 4000 },
  context: async ({ req }) => {
    // Forward auth to subgraphs
    return {
      headers: {
        authorization: req.headers.authorization,
        'x-request-id': req.headers['x-request-id'],
      },
    };
  },
});

console.log(`Gateway running at ${url}`);
```

---

## Schema Stitching

Schema stitching is an alternative to Federation for combining schemas. Better for integrating third-party APIs or legacy services that can't adopt Federation directives.

### Basic Schema Stitching

```typescript
import { stitchSchemas } from '@graphql-tools/stitch';
import { delegateToSchema } from '@graphql-tools/delegate';

const gatewaySchema = stitchSchemas({
  subschemas: [
    {
      schema: usersSchema,
      executor: usersExecutor,
      transforms: [
        // Optional: rename types to avoid conflicts
        new RenameTypes((name) => `Users_${name}`),
      ],
    },
    {
      schema: postsSchema,
      executor: postsExecutor,
    },
  ],
  // Type merging: combine types from different subschemas
  typeMergingOptions: {
    typeCandidateMerger: (candidates) => {
      // Custom merge logic
      return candidates[0]; // Default: use first candidate
    },
  },
});
```

### Type Merging

```typescript
const gatewaySchema = stitchSchemas({
  subschemas: [
    {
      schema: usersSchema,
      executor: usersExecutor,
      merge: {
        User: {
          // How to fetch users by ID from this subschema
          selectionSet: '{ id }',
          fieldName: 'user',
          args: (representation) => ({ id: representation.id }),
        },
      },
    },
    {
      schema: postsSchema,
      executor: postsExecutor,
      merge: {
        User: {
          // How to fetch user data from posts subschema
          selectionSet: '{ id }',
          fieldName: 'userPosts',
          args: (representation) => ({ userId: representation.id }),
        },
      },
    },
  ],
});
```

---

## Federation vs Stitching

| Feature | Apollo Federation | Schema Stitching |
|---------|------------------|------------------|
| Approach | Declarative (directives) | Imperative (code) |
| Subgraph awareness | Yes (subgraphs use federation directives) | No (any GraphQL schema) |
| Entity resolution | Built-in (@key, __resolveReference) | Manual (type merging, delegation) |
| Type extension | @key + extend type | Type merging config |
| Performance | Optimized query planning | Depends on implementation |
| Third-party schemas | Requires wrapper | Direct integration |
| Learning curve | Moderate | Steeper |
| Tooling | Apollo Studio, Rover CLI | graphql-tools |
| Production readiness | Apollo Router (Rust) | Node.js only |
| Best for | Greenfield, team-owned services | Legacy integration, third-party APIs |

### When to Use Federation
- Building new microservices from scratch
- Teams own their subgraphs independently
- Need production-grade routing (Apollo Router)
- Want managed tooling (Apollo Studio)

### When to Use Stitching
- Integrating existing REST/GraphQL APIs you don't control
- Legacy services that can't add federation directives
- Need maximum flexibility in composition logic
- One team manages the gateway

---

## Composition & Validation

### Rover CLI (Federation)

```bash
# Check subgraph schema for federation compliance
rover subgraph check my-graph@production \
  --schema ./users/schema.graphql \
  --name users

# Publish subgraph schema
rover subgraph publish my-graph@production \
  --schema ./users/schema.graphql \
  --name users \
  --routing-url http://users-service:4001/graphql

# Compose supergraph locally
rover supergraph compose --config supergraph.yaml > supergraph.graphql

# Validate composition
rover supergraph compose --config supergraph.yaml --output /dev/null
```

supergraph.yaml:

```yaml
federation_version: =2.3.2
subgraphs:
  users:
    routing_url: http://users-service:4001/graphql
    schema:
      file: ./subgraphs/users/schema.graphql
  products:
    routing_url: http://products-service:4002/graphql
    schema:
      file: ./subgraphs/products/schema.graphql
  orders:
    routing_url: http://orders-service:4003/graphql
    schema:
      file: ./subgraphs/orders/schema.graphql
  inventory:
    routing_url: http://inventory-service:4004/graphql
    schema:
      file: ./subgraphs/inventory/schema.graphql
```

### Common Composition Errors

```
Error: EXTERNAL_MISSING_ON_BASE
→ Field marked @external in subgraph A but doesn't exist in any subgraph's base definition.
Fix: Ensure another subgraph defines the field without @external.

Error: KEY_FIELDS_MISSING_EXTERNAL
→ @key references a field that isn't defined or marked @external.
Fix: Define the field in the subgraph or add @external.

Error: PROVIDES_FIELDS_MISSING_EXTERNAL
→ @provides references fields not marked @external.
Fix: Add @external to the fields in @provides.

Error: REQUIRES_FIELDS_MISSING_EXTERNAL
→ @requires references fields not marked @external.
Fix: Add @external to the fields in @requires.

Error: FIELD_TYPE_MISMATCH
→ Same field name has different types across subgraphs.
Fix: Ensure consistent types or use @shareable.

Error: NON_SHAREABLE_FIELD_OVERRIDE
→ Multiple subgraphs define the same field without @shareable.
Fix: Add @shareable to both definitions or use @override.
```

---

## Migration: Monolith to Federation

### Phase 1: Identify Boundaries

```
Current monolith schema:
├── User types (User, Address, Preferences)
├── Product types (Product, Category, Variant)
├── Order types (Order, OrderItem, Payment)
└── Review types (Review, Rating)

Proposed subgraphs:
├── users → User, Address, Preferences, Auth
├── products → Product, Category, Variant
├── orders → Order, OrderItem, Payment
└── reviews → Review, Rating
```

### Phase 2: Extract First Subgraph

1. Copy types and resolvers for the domain
2. Add Federation directives (@key, @external, etc.)
3. Deploy the subgraph
4. Update the gateway to include the subgraph
5. Route traffic through the gateway
6. Remove extracted code from the monolith

### Phase 3: Migrate Remaining Domains

Repeat Phase 2 for each domain, using @override for gradual migration:

```graphql
# In new subgraph: gradually take over fields
type User @key(fields: "id") {
  id: ID!
  email: String! @override(from: "monolith", percent: 10)  # Start with 10%
}

# Increase percent over time: 10% → 25% → 50% → 75% → 100%
# When at 100%, remove from monolith
```

---

## Best Practices

### Subgraph Design
- One bounded context per subgraph
- Each entity has exactly one "owning" subgraph
- Keep subgraphs independently deployable
- Minimize cross-subgraph dependencies

### Entity Design
- Use stable, unique identifiers for @key
- Prefer single-field keys when possible
- Add multiple @key definitions for flexibility
- Use non-resolvable keys for reference-only types

### Performance
- Use DataLoader in __resolveReference for batching
- Add @provides hints to reduce cross-subgraph calls
- Cache entity resolution at the gateway level
- Monitor query plan complexity

### Schema Evolution
- Additive changes only (no breaking changes)
- Deprecate fields before removing
- Use @override for gradual field migration
- Run composition checks in CI/CD
- Version subgraph schemas independently
