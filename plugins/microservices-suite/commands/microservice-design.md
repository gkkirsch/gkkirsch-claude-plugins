---
name: microservice-design
description: >
  One-command microservices architecture design — analyzes your codebase, identifies bounded contexts,
  designs service boundaries, generates architecture documentation, and configures API gateways and
  event-driven communication. Dispatches specialist agents for service decomposition, gateway design,
  event architecture, and service mesh configuration.
  Triggers: "/microservice-design", "design microservices", "decompose monolith", "service boundaries",
  "api gateway", "event driven", "service mesh", "bounded contexts", "service architecture".
user-invocable: true
argument-hint: "<decompose|gateway|events|mesh> [--domain <domain>] [--services <count>]"
allowed-tools: Read, Write, Edit, Bash, Glob, Grep
model: sonnet
---

# /microservice-design Command

One-command microservices architecture design. Analyzes your codebase, identifies domains and bounded
contexts, designs service boundaries, and generates comprehensive architecture documentation.

## Usage

```
/microservice-design                          # Full analysis: decompose + gateway + events
/microservice-design decompose                # Service decomposition and DDD analysis
/microservice-design gateway                  # API gateway design and configuration
/microservice-design events                   # Event-driven architecture design
/microservice-design mesh                     # Service mesh configuration
/microservice-design decompose --domain ecommerce  # Analyze with domain hint
```

## Subcommands

| Subcommand | Agent | Description |
|------------|-------|-------------|
| `decompose` | service-architect | DDD analysis, bounded contexts, service boundaries |
| `gateway` | api-gateway-designer | API gateway design (Kong, Envoy, NGINX, Traefik) |
| `events` | event-driven-engineer | Event sourcing, CQRS, Kafka/RabbitMQ design |
| `mesh` | service-mesh-operator | Istio, Linkerd, Consul Connect configuration |

## Procedure

### Step 1: Analyze the Codebase

Read the project root to understand the system:

1. **Read** `package.json`, `requirements.txt`, `go.mod`, `Cargo.toml`, `pom.xml` — determine language/framework
2. **Glob** for architectural files:
   - Models: `**/models/**`, `**/entities/**`, `**/domain/**`
   - Routes: `**/routes/**`, `**/controllers/**`, `**/handlers/**`
   - Services: `**/services/**`, `**/usecases/**`
   - Database: `**/migrations/**`, `**/schema.*`, `**/*.prisma`
   - Config: `docker-compose*.yml`, `**/k8s/**`, `**/terraform/**`
3. **Grep** for domain patterns:
   - Entity definitions: "class.*Model", "interface.*Entity", "@Entity"
   - API endpoints: "router.", "app.get|post", "@GetMapping", "@app.route"
   - Event patterns: "Event", "emit", "publish", "subscribe"
   - Database tables: "CREATE TABLE", "model " (Prisma)
4. **Check for existing architecture docs**: `ARCHITECTURE.md`, `docs/architecture/**`, `ADR/**`
5. **Count modules**: How many distinct domain areas exist?

Report findings:

```
Detected:
- Language: TypeScript (Node.js)
- Framework: Express + Prisma
- Domain areas: 6 (orders, products, customers, payments, shipping, auth)
- Entities: 24 (8 core, 16 supporting)
- API endpoints: 42
- Database tables: 18
- Existing services: 1 (monolith)
- Communication: Synchronous (HTTP only)
- Message broker: None detected
```

### Step 2: Route to Agent

Based on the subcommand (or full analysis), dispatch the appropriate agent:

#### `decompose` — Service Architect

```
Task tool:
  subagent_type: "service-architect"
  mode: "bypassPermissions"
  prompt: |
    Analyze this codebase and design microservices architecture.
    Stack: [detected stack]
    Domain areas: [discovered domains]
    Entities: [discovered entities]
    Current architecture: [monolith/partially decomposed/microservices]

    Deliverables:
    1. Bounded context map with relationships
    2. Service catalog (name, responsibility, data owned, team)
    3. Service contracts (API specs or event schemas)
    4. Data ownership matrix
    5. Migration plan (if decomposing monolith)
    6. Architecture Decision Records for key decisions
```

#### `gateway` — API Gateway Designer

```
Task tool:
  subagent_type: "api-gateway-designer"
  mode: "bypassPermissions"
  prompt: |
    Design an API gateway for this microservices architecture.
    Stack: [detected stack]
    Services: [discovered services]
    Clients: [web, mobile, third-party]
    Existing gateway: [none/Kong/Envoy/NGINX/Traefik]

    Deliverables:
    1. Gateway architecture (edge, BFF, federated)
    2. Route configuration for all services
    3. Authentication and authorization setup
    4. Rate limiting configuration
    5. Traffic management (canary, blue-green)
    6. Production-ready configuration files
```

#### `events` — Event-Driven Engineer

```
Task tool:
  subagent_type: "event-driven-engineer"
  mode: "bypassPermissions"
  prompt: |
    Design event-driven architecture for this system.
    Stack: [detected stack]
    Services: [discovered services]
    Current communication: [sync/async/mixed]
    Existing broker: [none/Kafka/RabbitMQ/NATS/SQS]

    Deliverables:
    1. Event catalog (all domain events with schemas)
    2. Topic/queue/exchange design
    3. Producer and consumer configurations
    4. Saga implementations for distributed transactions
    5. Dead letter queue strategy
    6. Event schema versioning approach
```

#### `mesh` — Service Mesh Operator

```
Task tool:
  subagent_type: "service-mesh-operator"
  mode: "bypassPermissions"
  prompt: |
    Configure a service mesh for this Kubernetes deployment.
    Stack: [detected stack]
    Services: [discovered services]
    K8s platform: [EKS/GKE/AKS/self-managed]
    Existing mesh: [none/Istio/Linkerd/Consul]
    Security requirements: [mTLS, authorization policies]

    Deliverables:
    1. Mesh selection recommendation (if no existing mesh)
    2. Installation and configuration
    3. Traffic management policies (retries, timeouts, circuit breakers)
    4. Security policies (mTLS, authorization)
    5. Observability setup (tracing, metrics, dashboards)
    6. Canary deployment configuration
```

### Step 3: Results Summary

After the agent completes, present results:

**For decomposition:**
```
Microservices Architecture Design:
- Bounded contexts identified: 6
- Services proposed: 7 (6 domain + 1 shared infrastructure)
- Migration strategy: Strangler Fig (4 phases, 6-12 months)
- Architecture Decision Records: 5
- Files created: docs/architecture/
  - context-map.md
  - service-catalog.md
  - data-ownership.md
  - migration-plan.md
  - ADR/*.md
```

**For gateway:**
```
API Gateway Design:
- Pattern: Backend for Frontend (BFF)
- Platform: Kong Gateway
- Routes configured: 42 across 7 services
- Auth: JWT validation at gateway
- Rate limiting: 3-tier (global, per-consumer, per-endpoint)
- Files created: infrastructure/gateway/
  - kong.yml
  - docker-compose.yml
```

**For events:**
```
Event-Driven Architecture Design:
- Events defined: 24 domain events
- Broker: Apache Kafka (12 topics, 3x replication)
- Sagas: 2 (Order Processing, Payment + Fulfillment)
- Event schemas: Avro with Schema Registry
- DLQ: Per-consumer dead letter topics
- Files created: infrastructure/kafka/
  - topics.sh
  - schemas/
  - consumer-configs/
```

**For mesh:**
```
Service Mesh Configuration:
- Platform: Istio 1.20
- mTLS: Strict (all namespaces)
- Authorization: 12 policies (deny-by-default)
- Traffic management: Retries, timeouts, circuit breakers per service
- Observability: Jaeger tracing, Prometheus metrics, Kiali dashboard
- Files created: infrastructure/istio/
  - virtual-services/
  - destination-rules/
  - authorization-policies/
```

## Error Recovery

| Error | Cause | Fix |
|-------|-------|-----|
| Too few entities found | Unconventional code structure | Provide domain model manually |
| No clear domain boundaries | Highly coupled codebase | Start with modules, extract later |
| Gateway platform unknown | No existing infrastructure | Default to Kong or NGINX |
| No K8s deployment found | Not containerized yet | Start with Docker Compose |
| Conflicting naming | Multiple domains use same terms | Event Storming workshop needed |

## Notes

- Full analysis (`/microservice-design` with no subcommand) runs decompose first, then recommends next steps
- Architecture documents are written to `docs/architecture/` by default
- Gateway configs are written to `infrastructure/gateway/`
- Event configs are written to `infrastructure/events/`
- Mesh configs are written to `infrastructure/mesh/`
- Always review generated architecture before implementing
- Start with fewer, larger services — split when complexity demands it
