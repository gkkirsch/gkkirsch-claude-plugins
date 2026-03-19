---
name: system-design-suite
description: >
  System Design & Architecture Suite — Expert-level guidance on distributed systems, scalability,
  reliability engineering, and API design. Four specialist agents deliver production-grade architectural
  advice, system design walkthroughs, back-of-envelope calculations, and codebase-aware recommendations.
  Triggers: "system design", "architecture review", "distributed system", "design a system",
  "scalability review", "scale my app", "horizontal scaling", "load balancing", "caching strategy",
  "database sharding", "reliability", "fault tolerance", "circuit breaker", "chaos engineering",
  "disaster recovery", "SLO", "SLI", "api design", "rest api", "grpc", "graphql design",
  "api versioning", "rate limiting", "design interview", "system design interview",
  "url shortener design", "chat system design", "newsfeed design", "notification system design".
  Dispatches the appropriate specialist agent: system-design-architect, scalability-engineer,
  reliability-engineer, or api-design-expert.
  NOT for: Frontend UI design, CSS/styling, database schema design without distributed concerns,
  or project management/planning.
version: 1.0.0
argument-hint: "<architect|scale|reliability|api> [topic-or-system]"
user-invocable: true
allowed-tools: Read, Grep, Glob, Bash
model: sonnet
---

# System Design & Architecture Suite

Production-grade system design and architecture agents for Claude Code. Four specialist agents covering distributed systems, scalability, reliability, and API design — the architectural expertise every serious engineering team needs.

## Available Agents

### System Design Architect (`system-design-architect`)
Distributed systems fundamentals, system design walkthroughs, back-of-envelope calculations, and architectural decision-making. Covers CAP theorem, consistency models, partitioning, replication, consensus algorithms, microservices decomposition, and event-driven architecture.

**Invoke**: Dispatch via Task tool with `subagent_type: "system-design-architect"`.

**Example prompts**:
- "Design a URL shortener that handles 100M daily active users"
- "Walk me through designing a real-time chat system"
- "Help me choose between eventual and strong consistency for my payment system"
- "Review my microservices architecture for a food delivery platform"

### Scalability Engineer (`scalability-engineer`)
Horizontal and vertical scaling strategies, load balancing, caching layers, database scaling, async processing, and CDN optimization. Analyzes your actual codebase and provides concrete scaling recommendations.

**Invoke**: Dispatch via Task tool with `subagent_type: "scalability-engineer"`.

**Example prompts**:
- "My API response times spike at 10K concurrent users — help me scale"
- "Design a caching strategy for my e-commerce product catalog"
- "How should I shard my PostgreSQL database for multi-tenant SaaS?"
- "Review my load balancing setup and recommend improvements"

### Reliability Engineer (`reliability-engineer`)
Fault tolerance patterns, circuit breakers, retry strategies, chaos engineering, observability, disaster recovery, and incident response. Helps you build systems that gracefully handle failures.

**Invoke**: Dispatch via Task tool with `subagent_type: "reliability-engineer"`.

**Example prompts**:
- "Add circuit breakers to my microservices communication"
- "Design a disaster recovery plan for my multi-region deployment"
- "Set up SLOs and SLIs for my payment processing service"
- "Review my retry logic — I think we have retry storms"

### API Design Expert (`api-design-expert`)
REST best practices, gRPC, GraphQL, API versioning, rate limiting, authentication, and pagination. Designs clean, consistent APIs and reviews existing ones for improvements.

**Invoke**: Dispatch via Task tool with `subagent_type: "api-design-expert"`.

**Example prompts**:
- "Design a RESTful API for a project management tool"
- "Should I use gRPC or REST for my internal service communication?"
- "Review my API for consistency and REST best practices"
- "Design a cursor-based pagination system for my feed API"

## Agent Selection Guide

| Need | Agent | Trigger |
|------|-------|---------|
| Design a complete system | system-design-architect | "Design a ..." |
| Back-of-envelope calculations | system-design-architect | "How much storage/bandwidth do I need?" |
| Consistency/partitioning choices | system-design-architect | "Which consistency model?" |
| Handle more traffic | scalability-engineer | "Scale my app" |
| Caching strategy | scalability-engineer | "Add caching" |
| Database scaling | scalability-engineer | "Shard my database" |
| Load balancing | scalability-engineer | "Load balancer setup" |
| Handle failures gracefully | reliability-engineer | "Circuit breakers" |
| Disaster recovery | reliability-engineer | "DR plan" |
| Monitoring/alerting | reliability-engineer | "SLOs and observability" |
| Incident management | reliability-engineer | "On-call and runbooks" |
| API design/review | api-design-expert | "Design my API" |
| API versioning | api-design-expert | "Version my API" |
| Rate limiting | api-design-expert | "Rate limit my API" |
| Auth strategy | api-design-expert | "API authentication" |

## Reference Materials

This suite includes comprehensive reference documents in `references/`:

- **distributed-patterns.md** — Saga pattern, event sourcing, CQRS, outbox pattern, two-phase commit, CRDTs, and other distributed system patterns with implementation guidance
- **caching-strategies.md** — Write-through, write-behind, read-through, cache-aside, TTL strategies, cache invalidation patterns, and multi-layer caching architectures
- **messaging-systems.md** — Kafka, RabbitMQ, SQS, pub/sub patterns, exactly-once delivery, dead letter queues, and message ordering guarantees

Agents automatically consult these references when working. You can also read them directly for quick answers.

## How It Works

1. You describe your architecture need (e.g., "design a notification system")
2. The SKILL.md routes to the appropriate specialist agent
3. The agent analyzes your requirements, considers tradeoffs, and references best practices
4. You receive architectural recommendations, diagrams (ASCII), calculations, and concrete implementation guidance
5. For codebase reviews, the agent reads your actual code and provides specific suggestions

All recommendations are production-grade:
- Based on real-world patterns from companies like Netflix, Uber, Stripe, and Google
- Include tradeoff analysis — not just "do X" but "X is better than Y because..."
- Provide concrete numbers: throughput estimates, storage calculations, latency targets
- Consider operational complexity, not just technical elegance
