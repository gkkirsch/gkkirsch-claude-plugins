---
name: microservices-architecture-suite
description: >
  Microservices Architecture Suite — AI-powered toolkit for service decomposition with Domain-Driven
  Design, API gateway design with Kong/Envoy/NGINX/Traefik, event-driven architecture with Kafka/RabbitMQ/NATS,
  and service mesh operations with Istio/Linkerd/Consul Connect. Expert agents for designing, building,
  and operating microservices at scale.
  Triggers: "microservice", "microservices", "service decomposition", "bounded context", "domain driven design",
  "DDD", "api gateway", "kong", "envoy", "nginx gateway", "traefik", "event driven", "event sourcing",
  "CQRS", "kafka", "rabbitmq", "nats", "message broker", "service mesh", "istio", "linkerd", "consul connect",
  "service discovery", "circuit breaker", "saga pattern", "distributed transaction", "monolith decomposition",
  "strangler fig", "backend for frontend", "BFF", "grpc service", "service contract",
  "context map", "aggregate root", "domain event", "event schema".
  Dispatches the appropriate specialist agent: service-architect, api-gateway-designer,
  event-driven-engineer, or service-mesh-operator.
  NOT for: Frontend/UI development, database administration without microservices context,
  CI/CD pipeline setup, or general cloud infrastructure (use devops-suite).
version: 1.0.0
argument-hint: "<decompose|gateway|events|mesh> [options]"
user-invocable: true
allowed-tools: Read, Grep, Glob, Bash
model: sonnet
---

# Microservices Architecture Suite

Production-grade microservices architecture design and operations agents for Claude Code. Four specialist
agents that handle service decomposition, API gateway design, event-driven architecture, and service
mesh configuration — the architecture work that every microservices project needs.

## Available Agents

### Service Architect (`service-architect`)
Decomposes monoliths into microservices using Domain-Driven Design. Identifies bounded contexts and
aggregates, designs service boundaries with proper data ownership, creates context maps, generates
service contracts, and plans migration strategies.

**Invoke**: Dispatch via Task tool with `subagent_type: "service-architect"`.

**Example prompts**:
- "Analyze my codebase and propose microservices boundaries"
- "Identify bounded contexts using Domain-Driven Design"
- "Create a migration plan from monolith to microservices"
- "Design service contracts for my order processing domain"
- "Map the aggregate roots and domain events in my system"

### API Gateway Designer (`api-gateway-designer`)
Designs and configures API gateways using Kong, Envoy, NGINX, Traefik, and AWS API Gateway. Implements
BFF patterns, rate limiting, authentication, routing, circuit breaking, request/response transformation,
API versioning, and traffic management.

**Invoke**: Dispatch via Task tool with `subagent_type: "api-gateway-designer"`.

**Example prompts**:
- "Design an API gateway for my microservices using Kong"
- "Configure Envoy as a reverse proxy with circuit breaking"
- "Implement Backend for Frontend pattern for web and mobile clients"
- "Set up rate limiting and JWT authentication at the gateway"
- "Design canary deployment routing with traffic splitting"

### Event-Driven Engineer (`event-driven-engineer`)
Designs event-driven architectures with Kafka, RabbitMQ, NATS, and AWS SNS/SQS. Implements event
sourcing, CQRS, saga patterns, dead letter queues, idempotent consumers, and change data capture.

**Invoke**: Dispatch via Task tool with `subagent_type: "event-driven-engineer"`.

**Example prompts**:
- "Design event-driven architecture for my order processing system"
- "Implement the saga pattern for distributed transactions across services"
- "Configure Kafka topics and consumer groups for my microservices"
- "Design event schemas with versioning for backward compatibility"
- "Set up CQRS with event sourcing for my domain"

### Service Mesh Operator (`service-mesh-operator`)
Configures Istio, Linkerd, Consul Connect, and Kuma service meshes. Implements mTLS, traffic management,
circuit breakers, fault injection, canary deployments, authorization policies, and observability.

**Invoke**: Dispatch via Task tool with `subagent_type: "service-mesh-operator"`.

**Example prompts**:
- "Configure Istio service mesh for my Kubernetes cluster"
- "Set up mutual TLS between all microservices"
- "Implement canary deployments with Istio traffic splitting"
- "Configure circuit breakers and retry policies in the mesh"
- "Design authorization policies for service-to-service communication"

## Quick Start: /microservice-design

Use the `/microservice-design` command for guided microservices architecture:

```
/microservice-design                    # Full architecture analysis
/microservice-design decompose          # DDD and service boundary analysis
/microservice-design gateway            # API gateway design
/microservice-design events             # Event-driven architecture
/microservice-design mesh               # Service mesh configuration
```

The command auto-detects your codebase structure, identifies domains, and routes to the right agent.

## Agent Selection Guide

| Need | Agent | Command |
|------|-------|---------|
| Split monolith into services | service-architect | "Decompose my monolith" |
| Define bounded contexts | service-architect | "Identify bounded contexts" |
| Design service contracts | service-architect | "Generate service contracts" |
| Plan migration strategy | service-architect | "Create migration plan" |
| Configure API gateway | api-gateway-designer | "Design API gateway" |
| Set up rate limiting | api-gateway-designer | "Configure rate limiting" |
| Design BFF pattern | api-gateway-designer | "Implement BFF pattern" |
| Traffic management | api-gateway-designer | "Set up canary deployment" |
| Design event system | event-driven-engineer | "Design event architecture" |
| Implement sagas | event-driven-engineer | "Implement saga pattern" |
| Configure Kafka | event-driven-engineer | "Set up Kafka topics" |
| Event sourcing/CQRS | event-driven-engineer | "Implement event sourcing" |
| Configure service mesh | service-mesh-operator | "Set up Istio mesh" |
| mTLS between services | service-mesh-operator | "Configure mutual TLS" |
| Circuit breakers | service-mesh-operator | "Set up circuit breakers" |
| Fault injection testing | service-mesh-operator | "Configure fault injection" |

## Reference Materials

This suite includes comprehensive reference documents in `references/`:

- **distributed-patterns.md** — Saga, circuit breaker, retry, bulkhead, timeout, rate limiter,
  outbox, event sourcing, API composition, and anti-corruption layer patterns
- **event-streaming.md** — Apache Kafka, RabbitMQ, NATS JetStream, and AWS SNS/SQS deep reference
  with configuration, monitoring, and operational best practices
- **service-communication.md** — gRPC, REST, GraphQL, async messaging, WebSocket, and SSE protocol
  comparison with implementation patterns and performance tuning

Agents automatically consult these references when working. You can also read them directly for quick answers.

## How It Works

1. You describe what you need (e.g., "decompose my monolith into microservices")
2. The SKILL.md routes to the appropriate agent
3. The agent reads your code, analyzes the domain, and identifies patterns
4. Architecture documents, configurations, or code are generated
5. The agent provides results and recommended next steps

All generated artifacts follow industry best practices:
- DDD: Bounded contexts, aggregates, context maps, ubiquitous language
- Gateway: Production-ready configs, security, rate limiting, observability
- Events: Schema versioning, idempotency, DLQ, exactly-once semantics
- Mesh: mTLS, authorization policies, traffic management, fault tolerance
