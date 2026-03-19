# Service Communication Reference

Deep reference for inter-service communication protocols: REST, gRPC, GraphQL, async messaging,
WebSockets, and Server-Sent Events. Covers protocol selection, implementation patterns, performance
characteristics, and production best practices.

## Protocol Comparison Matrix

| Feature | REST (HTTP) | gRPC | GraphQL | Async Messaging | WebSocket | SSE |
|---------|-------------|------|---------|-----------------|-----------|-----|
| Transport | HTTP/1.1, HTTP/2 | HTTP/2 | HTTP/1.1, HTTP/2 | TCP (broker) | TCP | HTTP/1.1 |
| Data format | JSON, XML | Protobuf (binary) | JSON | JSON, Avro, Protobuf | Any | Text |
| Contract | OpenAPI (opt) | .proto (required) | Schema (required) | Schema (optional) | None | None |
| Code gen | Optional | Built-in | Optional | Optional | None | None |
| Streaming | No (SSE for push) | Bidirectional | Subscriptions | N/A | Bidirectional | Server push |
| Browser support | Full | grpc-web (limited) | Full | No (needs gateway) | Full | Full |
| Coupling | Low | Medium | Low | Very low | Medium | Low |
| Performance | Good | Excellent | Good | Excellent | Good | Good |
| Best for | Public APIs | Internal services | Frontend queries | Event-driven | Real-time | Notifications |

## REST (HTTP/JSON)

### RESTful API Design Guide

**Resource naming:**

```
Good:
  GET    /api/v1/orders              # List orders
  POST   /api/v1/orders              # Create order
  GET    /api/v1/orders/{id}         # Get order
  PUT    /api/v1/orders/{id}         # Replace order
  PATCH  /api/v1/orders/{id}         # Update order fields
  DELETE /api/v1/orders/{id}         # Delete order
  GET    /api/v1/orders/{id}/items   # List order items
  POST   /api/v1/orders/{id}/submit  # Submit order (action)

Bad:
  GET    /api/v1/getOrders           # Verb in URL
  POST   /api/v1/createOrder         # Verb in URL
  GET    /api/v1/order               # Singular collection
  POST   /api/v1/orders/delete/{id}  # Wrong HTTP method
```

**HTTP status codes:**

| Code | Meaning | When to Use |
|------|---------|-------------|
| 200 | OK | GET success, PATCH success, action success |
| 201 | Created | POST success (include Location header) |
| 202 | Accepted | Async operation started |
| 204 | No Content | DELETE success, PUT with no body |
| 301 | Moved Permanently | Resource permanently relocated |
| 304 | Not Modified | Conditional GET, ETag match |
| 400 | Bad Request | Validation error, malformed request |
| 401 | Unauthorized | Missing or invalid authentication |
| 403 | Forbidden | Authenticated but not authorized |
| 404 | Not Found | Resource doesn't exist |
| 405 | Method Not Allowed | Wrong HTTP method |
| 409 | Conflict | Concurrent modification, duplicate |
| 410 | Gone | Resource permanently deleted |
| 413 | Payload Too Large | Request body exceeds limit |
| 415 | Unsupported Media Type | Wrong Content-Type |
| 422 | Unprocessable Entity | Semantic validation error |
| 429 | Too Many Requests | Rate limit exceeded |
| 500 | Internal Server Error | Unexpected server error |
| 502 | Bad Gateway | Upstream service error |
| 503 | Service Unavailable | Service temporarily down |
| 504 | Gateway Timeout | Upstream timeout |

**Standard error response format:**

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Request validation failed",
    "details": [
      {
        "field": "email",
        "message": "Must be a valid email address",
        "value": "not-an-email"
      },
      {
        "field": "quantity",
        "message": "Must be greater than 0",
        "value": -1
      }
    ],
    "requestId": "req_abc123",
    "timestamp": "2025-01-15T10:30:00Z",
    "documentation": "https://docs.example.com/errors/VALIDATION_ERROR"
  }
}
```

**Pagination patterns:**

```
Offset-based (simple, has performance issues at large offsets):
  GET /api/v1/orders?page=3&pageSize=20
  Response: { data: [...], total: 1000, page: 3, pageSize: 20, totalPages: 50 }

Cursor-based (performant, recommended):
  GET /api/v1/orders?cursor=eyJpZCI6MTIzfQ&limit=20
  Response: { data: [...], nextCursor: "eyJpZCI6MTQzfQ", hasMore: true }

Keyset-based (similar to cursor, explicit):
  GET /api/v1/orders?after_id=123&limit=20
  Response: { data: [...], has_more: true }
```

**Filtering, sorting, and field selection:**

```
Filtering:
  GET /api/v1/orders?status=shipped&customerId=abc123&createdAfter=2025-01-01

Sorting:
  GET /api/v1/orders?sort=-createdAt,+total  (- desc, + asc)
  GET /api/v1/orders?sort=createdAt:desc,total:asc

Field selection (sparse fieldsets):
  GET /api/v1/orders?fields=id,status,total,createdAt
  GET /api/v1/orders?include=items,customer  (related resources)
```

**HATEOAS links:**

```json
{
  "data": {
    "id": "order-123",
    "status": "submitted",
    "total": 99.99
  },
  "_links": {
    "self": { "href": "/api/v1/orders/order-123" },
    "cancel": { "href": "/api/v1/orders/order-123/cancel", "method": "POST" },
    "items": { "href": "/api/v1/orders/order-123/items" },
    "customer": { "href": "/api/v1/customers/cust-456" }
  }
}
```

### HTTP Client Best Practices

**Connection pooling:**

```typescript
// Node.js HTTP agent with connection pooling
import http from 'http';
import https from 'https';

const httpAgent = new http.Agent({
  keepAlive: true,
  keepAliveMsecs: 30000,
  maxSockets: 100,          // Max connections per host
  maxTotalSockets: 256,     // Max connections total
  maxFreeSockets: 10,       // Max idle connections to keep
  timeout: 30000,
  scheduling: 'fifo',
});

// Use with fetch or axios
const response = await fetch('http://order-service:3000/orders', {
  agent: httpAgent,
  signal: AbortSignal.timeout(10000),
});
```

**Request/response interceptors:**

```typescript
// Axios interceptor for correlation ID propagation
axios.interceptors.request.use(config => {
  config.headers['X-Request-ID'] = getRequestId();
  config.headers['X-Correlation-ID'] = getCorrelationId();
  config.headers['X-Source-Service'] = 'order-service';
  return config;
});

// Response interceptor for error normalization
axios.interceptors.response.use(
  response => response,
  error => {
    if (error.response) {
      // Server responded with error status
      throw new ServiceError(
        error.response.data?.error?.message || 'Service error',
        error.response.status,
        error.response.data?.error?.code
      );
    } else if (error.request) {
      // No response received
      throw new ServiceUnavailableError(
        `Service unreachable: ${error.config.url}`
      );
    }
    throw error;
  }
);
```

## gRPC

### Protocol Buffers (Protobuf)

**Message definition best practices:**

```protobuf
syntax = "proto3";

package order.v1;

import "google/protobuf/timestamp.proto";
import "google/protobuf/wrappers.proto";
import "google/protobuf/field_mask.proto";

// Service definition
service OrderService {
  // Unary RPCs
  rpc CreateOrder(CreateOrderRequest) returns (CreateOrderResponse);
  rpc GetOrder(GetOrderRequest) returns (Order);
  rpc UpdateOrder(UpdateOrderRequest) returns (Order);
  rpc DeleteOrder(DeleteOrderRequest) returns (google.protobuf.Empty);
  rpc ListOrders(ListOrdersRequest) returns (ListOrdersResponse);

  // Server streaming
  rpc WatchOrderStatus(WatchOrderRequest) returns (stream OrderStatusUpdate);

  // Client streaming
  rpc BulkCreateOrders(stream CreateOrderRequest) returns (BulkCreateResponse);

  // Bidirectional streaming
  rpc OrderChat(stream OrderMessage) returns (stream OrderMessage);
}

// Use wrapper types for optional fields
message UpdateOrderRequest {
  string order_id = 1;
  google.protobuf.StringValue notes = 2;           // Optional string
  google.protobuf.Int32Value priority = 3;          // Optional int
  google.protobuf.FieldMask update_mask = 4;        // Which fields to update
}

// Pagination
message ListOrdersRequest {
  int32 page_size = 1;
  string page_token = 2;                            // Cursor for next page
  string filter = 3;                                // e.g., "status=shipped"
  string order_by = 4;                              // e.g., "created_at desc"
}

message ListOrdersResponse {
  repeated Order orders = 1;
  string next_page_token = 2;
  int32 total_size = 3;
}
```

**gRPC error handling:**

```
gRPC Status Codes (mapped to HTTP):

| gRPC Code | HTTP | Description |
|-----------|------|-------------|
| OK (0) | 200 | Success |
| CANCELLED (1) | 499 | Client cancelled |
| UNKNOWN (2) | 500 | Unknown error |
| INVALID_ARGUMENT (3) | 400 | Bad request |
| DEADLINE_EXCEEDED (4) | 504 | Timeout |
| NOT_FOUND (5) | 404 | Resource not found |
| ALREADY_EXISTS (6) | 409 | Duplicate |
| PERMISSION_DENIED (7) | 403 | Forbidden |
| RESOURCE_EXHAUSTED (8) | 429 | Rate limited |
| FAILED_PRECONDITION (9) | 400 | State conflict |
| ABORTED (10) | 409 | Concurrent conflict |
| OUT_OF_RANGE (11) | 400 | Value out of range |
| UNIMPLEMENTED (12) | 501 | Not implemented |
| INTERNAL (13) | 500 | Internal error |
| UNAVAILABLE (14) | 503 | Service unavailable |
| DATA_LOSS (15) | 500 | Data corruption |
| UNAUTHENTICATED (16) | 401 | Not authenticated |
```

### gRPC Performance Tuning

| Parameter | Default | Recommendation |
|-----------|---------|----------------|
| Max message size | 4MB | 4-16MB |
| Keepalive time | Infinite | 30-60s |
| Keepalive timeout | 20s | 10-20s |
| Max concurrent streams | 100 | 100-1000 |
| Initial window size | 64KB | 1-4MB |
| Flow control window | 64KB | 1-4MB |
| Compression | None | gzip for large payloads |
| Connection pooling | 1 per channel | 1 (HTTP/2 multiplexing) |

### gRPC vs REST Decision

| Choose gRPC When | Choose REST When |
|-------------------|------------------|
| Internal service-to-service | Public-facing APIs |
| High performance needed | Browser clients (no grpc-web) |
| Streaming required | Simple CRUD operations |
| Strong typing needed | Team unfamiliar with Protobuf |
| Polyglot services | Cacheable responses (GET) |
| Binary data transfer | Human-readable debugging |

## GraphQL

### GraphQL for Microservices

**Apollo Federation (recommended for microservices):**

```
Client → API Gateway (Apollo Router) → Subgraph A (Order Service)
                                      → Subgraph B (Product Service)
                                      → Subgraph C (Customer Service)
```

**Subgraph schema:**

```graphql
# Order subgraph
extend schema @link(url: "https://specs.apollo.dev/federation/v2.0", import: ["@key", "@external", "@requires"])

type Order @key(fields: "id") {
  id: ID!
  customerId: ID!
  customer: Customer!
  items: [OrderItem!]!
  status: OrderStatus!
  total: Float!
  createdAt: DateTime!
}

type OrderItem {
  productId: ID!
  product: Product!
  quantity: Int!
  unitPrice: Float!
}

# Reference to Customer from Customer subgraph
type Customer @key(fields: "id") {
  id: ID! @external
}

# Reference to Product from Product subgraph
type Product @key(fields: "id") {
  id: ID! @external
}

type Query {
  order(id: ID!): Order
  orders(customerId: ID, status: OrderStatus, first: Int, after: String): OrderConnection!
}

type Mutation {
  createOrder(input: CreateOrderInput!): Order!
  submitOrder(orderId: ID!): Order!
  cancelOrder(orderId: ID!, reason: String!): Order!
}

enum OrderStatus {
  DRAFT
  SUBMITTED
  PAID
  SHIPPED
  DELIVERED
  CANCELLED
}
```

### GraphQL Anti-Patterns in Microservices

| Anti-Pattern | Problem | Solution |
|-------------|---------|----------|
| N+1 queries | Fetching related data in loops | DataLoader batching |
| Unbounded queries | `{ orders { items { product { reviews { ... } } } } }` | Query depth limiting |
| Giant mutations | Single mutation doing too much | Split into focused mutations |
| No pagination | Fetching all records | Cursor-based connections |
| Schema coupling | GraphQL schema matches DB schema | Domain-driven schema design |

## Async Messaging Patterns

### Message Types

| Type | Description | Example | Response? |
|------|-------------|---------|-----------|
| Command | Request to do something | `CreateOrder` | Expected (async) |
| Event | Notification of something that happened | `OrderCreated` | No |
| Document | Data transfer | `OrderReport` | No |
| Query | Request for information | `GetOrderStatus` | Expected (async) |

### Request/Reply Over Messaging

```
┌─────────┐  Request (correlation-id: abc)  ┌─────────┐
│ Client  │ ──────────────────────────────→ │ Server  │
│         │                                 │         │
│         │  Reply (correlation-id: abc)    │         │
│         │ ←────────────────────────────── │         │
└─────────┘  (to reply-to queue)           └─────────┘
```

```typescript
// RabbitMQ RPC pattern
async function rpcCall(request: any): Promise<any> {
  const correlationId = uuidv4();
  const replyQueue = await channel.assertQueue('', { exclusive: true, autoDelete: true });

  return new Promise((resolve, reject) => {
    const timeout = setTimeout(() => reject(new Error('RPC timeout')), 30000);

    channel.consume(replyQueue.queue, (msg) => {
      if (msg?.properties.correlationId === correlationId) {
        clearTimeout(timeout);
        resolve(JSON.parse(msg.content.toString()));
      }
    }, { noAck: true });

    channel.publish('', 'rpc.order-service', Buffer.from(JSON.stringify(request)), {
      correlationId,
      replyTo: replyQueue.queue,
      expiration: '30000',
    });
  });
}
```

### Message Delivery Guarantees

| Guarantee | Description | Implementation |
|-----------|-------------|---------------|
| At-most-once | Message may be lost, never duplicated | Fire-and-forget, no ack |
| At-least-once | Message delivered 1+ times, may duplicate | Ack after processing |
| Exactly-once | Message delivered exactly once | Idempotent consumer + dedup |

**Idempotent consumer pattern:**

```typescript
class IdempotentConsumer {
  constructor(private deduplicationStore: DeduplicationStore) {}

  async process(event: DomainEvent, handler: (e: DomainEvent) => Promise<void>): Promise<void> {
    const messageId = event.eventId;

    // Check if already processed
    if (await this.deduplicationStore.exists(messageId)) {
      logger.debug({ messageId }, 'Duplicate message skipped');
      return;
    }

    // Process the message
    await handler(event);

    // Mark as processed (with TTL for cleanup)
    await this.deduplicationStore.mark(messageId, 7 * 24 * 60 * 60); // 7 day TTL
  }
}
```

## WebSocket Communication

### WebSocket vs SSE vs Polling

| Feature | WebSocket | SSE | Long Polling |
|---------|-----------|-----|-------------|
| Direction | Bidirectional | Server → Client | Client → Server |
| Protocol | ws:// / wss:// | HTTP | HTTP |
| Reconnection | Manual | Built-in | Manual |
| Binary data | Yes | No (text only) | Yes |
| Multiplexing | Manual | HTTP/2 multiplexing | Per-request |
| Proxy support | Limited | Excellent | Excellent |
| Browser support | Full | Full (except IE) | Full |
| Best for | Chat, gaming, collaboration | Notifications, dashboards | Compatibility |

### When to Use Each Protocol

| Scenario | Protocol | Why |
|----------|----------|-----|
| Service-to-service sync call | gRPC | Performance, type safety |
| Service-to-service async | Message broker | Decoupled, reliable |
| Public API | REST | Universal support |
| Frontend data fetching | GraphQL or REST | Flexible queries |
| Real-time updates to browser | SSE or WebSocket | Push-based |
| Mobile notifications | Push notifications | Battery-efficient |
| File upload | REST multipart | Well-supported |
| Streaming data processing | gRPC streaming or Kafka | High throughput |

## Service Communication Security

### mTLS (Mutual TLS)

Both client and server authenticate each other with certificates.

```
Client                               Server
  │                                    │
  │  ClientHello                       │
  │──────────────────────────────────→│
  │                                    │
  │  ServerHello + ServerCert          │
  │  + CertificateRequest             │
  │←──────────────────────────────────│
  │                                    │
  │  ClientCert + ClientKeyExchange    │
  │──────────────────────────────────→│
  │                                    │
  │  (both verify certificates)        │
  │                                    │
  │  Encrypted Application Data        │
  │←─────────────────────────────────→│
```

**Use mTLS for:**
- All service-to-service communication in production
- Database connections
- Message broker connections

**Manage with:**
- Service mesh (Istio, Linkerd) — automatic mTLS
- cert-manager on Kubernetes — automated certificate lifecycle
- HashiCorp Vault — certificate authority and secrets management

### API Key Authentication

Best for third-party API consumers.

```
Headers:
  X-API-Key: sk_live_abc123
  # or
  Authorization: ApiKey sk_live_abc123
```

**API key best practices:**
- Prefix keys by environment: `sk_live_`, `sk_test_`
- Store hashed in database, never plaintext
- Support key rotation (allow 2 active keys per consumer)
- Associate keys with rate limits and permissions
- Log key usage for auditing
- Revoke immediately on compromise

### JWT Token Propagation

```
External Client → API Gateway → Service A → Service B
  [JWT token]     [validates]   [forwards]   [trusts mesh]
                  [adds claims] [reads claims]

Gateway responsibilities:
  - Validate JWT signature and expiration
  - Extract claims (userId, roles, permissions)
  - Forward claims as headers: X-User-ID, X-User-Roles

Internal service responsibilities:
  - Trust gateway-injected headers (mesh provides mTLS)
  - Never validate JWT again (gateway already did)
  - Use claims for authorization decisions
```

## Performance Optimization

### Connection Management

| Strategy | Protocol | Benefit |
|----------|----------|---------|
| Keep-alive | HTTP/1.1 | Reuse TCP connections |
| Multiplexing | HTTP/2, gRPC | Multiple requests on one connection |
| Connection pooling | All | Pre-established connections |
| DNS caching | All | Avoid repeated DNS lookups |
| Circuit breaking | All | Fail fast when service is down |

### Payload Optimization

| Technique | Benefit | Trade-off |
|-----------|---------|-----------|
| Compression (gzip, brotli) | 60-80% smaller payloads | CPU overhead |
| Binary format (Protobuf) | 3-10x smaller than JSON | Less readable |
| Field selection | Only requested fields | Implementation complexity |
| Pagination | Bounded response size | Multiple requests |
| Caching (ETag, Cache-Control) | Avoid redundant requests | Staleness |
| Batch APIs | Fewer round trips | Larger payloads |

### Latency Budget

```
User-facing request: 200ms budget

  API Gateway:     5ms  (routing, auth check)
  Service A:       50ms (business logic + DB query)
  Service A → B:   20ms (gRPC call)
  Service B:       30ms (business logic + DB query)
  Service A → C:   20ms (gRPC call, parallel with B)
  Service C:       30ms (business logic)
  Response:        5ms  (serialization, compression)
  ─────────────────────
  Total:           ~160ms (within 200ms budget)

  Buffer:          40ms  (for retries, spikes, network jitter)
```

## Protocol Selection Decision Tree

```
Is the consumer a browser?
├── Yes: Need real-time updates?
│   ├── Yes: Bidirectional?
│   │   ├── Yes → WebSocket
│   │   └── No → SSE
│   └── No: Complex data requirements?
│       ├── Yes → GraphQL
│       └── No → REST
└── No: Internal service-to-service?
    ├── Yes: Need streaming?
    │   ├── Yes → gRPC streaming
    │   └── No: Synchronous or async?
    │       ├── Sync → gRPC (performance) or REST (simplicity)
    │       └── Async → Message broker (Kafka/RabbitMQ/NATS)
    └── No: Third-party integration?
        ├── Inbound → REST + webhooks
        └── Outbound → REST (their API format)
```
