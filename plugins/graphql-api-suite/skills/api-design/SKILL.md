---
name: graphql-api-suite
description: >
  GraphQL & API Design Suite — comprehensive toolkit for designing, building, and optimizing
  GraphQL and REST APIs. Schema architecture with resolvers and subscriptions, REST API design
  with HATEOAS and versioning, query optimization with DataLoader and caching, Apollo Federation,
  and API versioning with migration strategies.
  Triggers: "graphql schema", "graphql api", "design graphql", "schema design", "graphql resolver",
  "graphql subscription", "graphql dataloader", "n+1 query", "graphql optimize", "graphql federation",
  "rest api design", "rest best practices", "hateoas", "api pagination", "api versioning",
  "api deprecation", "breaking changes", "api migration", "openapi", "rest endpoint",
  "api design", "schema optimization", "persisted queries", "api error handling",
  "cursor pagination", "relay connection", "graphql yoga", "apollo server", "pothos",
  "graphql codegen", "api rate limiting", "query complexity", "graphql security".
  Dispatches the appropriate specialist agent: graphql-architect, rest-api-designer,
  schema-optimizer, or api-versioning-expert.
  NOT for: Frontend/UI development, database schema design without API layer,
  generic backend development without API focus, or network infrastructure.
version: 1.0.0
argument-hint: "<graphql|rest|optimize|version> [target]"
user-invocable: true
allowed-tools: Read, Grep, Glob, Bash
model: sonnet
---

# GraphQL & API Design Suite

Production-grade API design and optimization agents for Claude Code. Four specialist agents that handle GraphQL schema design, REST API architecture, performance optimization, and API versioning — the complete API development lifecycle.

## Available Agents

### GraphQL Architect (`graphql-architect`)
Designs and builds GraphQL APIs from the ground up. Schema-first and code-first design, resolver architecture, DataLoader integration, subscription setup, custom scalars, directives, error handling, and testing. Supports Apollo Server, GraphQL Yoga, Mercurius, Pothos, Nexus, and TypeGraphQL.

**Invoke**: Dispatch via Task tool with `subagent_type: "graphql-architect"`.

**Example prompts**:
- "Design a GraphQL schema for my e-commerce app"
- "Add subscriptions for real-time notifications"
- "Write resolvers with DataLoader for my blog API"
- "Set up a code-first schema with Pothos and Prisma"

### REST API Designer (`rest-api-designer`)
Designs RESTful APIs with best practices. Resource modeling, HATEOAS links, pagination, filtering, sorting, RFC 7807 error handling, content negotiation, OpenAPI 3.1 documentation, and security configuration.

**Invoke**: Dispatch via Task tool with `subagent_type: "rest-api-designer"`.

**Example prompts**:
- "Design REST endpoints for user management"
- "Add HATEOAS links and cursor pagination to my API"
- "Generate an OpenAPI 3.1 spec from my Express routes"
- "Review my REST API for best practice compliance"

### Schema Optimizer (`schema-optimizer`)
Optimizes GraphQL API performance. N+1 query detection, DataLoader implementation, response caching, persisted queries, query complexity analysis, depth limiting, database optimization, and monitoring setup.

**Invoke**: Dispatch via Task tool with `subagent_type: "schema-optimizer"`.

**Example prompts**:
- "Find and fix N+1 queries in my GraphQL API"
- "Add Redis caching to my GraphQL responses"
- "Set up query complexity analysis and depth limiting"
- "Performance audit my GraphQL API"

### API Versioning Expert (`api-versioning-expert`)
Plans and implements API versioning strategies. URL/header/content-negotiation versioning, breaking change detection, deprecation workflows, sunset planning, migration guides, and monitoring dashboards.

**Invoke**: Dispatch via Task tool with `subagent_type: "api-versioning-expert"`.

**Example prompts**:
- "What versioning strategy should I use for my public API?"
- "Check my schema changes for breaking changes"
- "Create a deprecation plan for the v1 API"
- "Generate a migration guide from v1 to v2"

## Quick Start: /api-design

Use the `/api-design` command for guided API design and optimization:

```
/api-design                        # Auto-detect and suggest improvements
/api-design graphql                # Design or review GraphQL schema
/api-design graphql --federation   # Design federated architecture
/api-design rest                   # Design or review REST endpoints
/api-design rest --openapi         # Generate OpenAPI documentation
/api-design optimize               # Full performance optimization
/api-design optimize --n1          # Fix N+1 query problems
/api-design version                # Plan versioning strategy
/api-design version --breaking     # Check for breaking changes
```

The `/api-design` command auto-detects your framework, discovers your schema, and routes to the right agent.

## Agent Selection Guide

| Need | Agent | Command |
|------|-------|---------|
| Design GraphQL schema | graphql-architect | "Design a GraphQL schema" |
| Write resolvers | graphql-architect | "Write resolvers for my types" |
| Set up subscriptions | graphql-architect | "Add real-time subscriptions" |
| DataLoader setup | graphql-architect | "Add DataLoader to my resolvers" |
| Design REST endpoints | rest-api-designer | "Design REST API for orders" |
| Add HATEOAS | rest-api-designer | "Add HATEOAS links" |
| Generate OpenAPI | rest-api-designer | "Generate OpenAPI spec" |
| Fix N+1 queries | schema-optimizer | "Find N+1 problems" |
| Add caching | schema-optimizer | "Set up response caching" |
| Performance audit | schema-optimizer | "Audit API performance" |
| Federation design | graphql-architect | "Design federated subgraphs" |
| Optimize federation | schema-optimizer | "Optimize federated queries" |
| Plan versioning | api-versioning-expert | "Plan versioning strategy" |
| Check breaking changes | api-versioning-expert | "Check for breaking changes" |
| Deprecation workflow | api-versioning-expert | "Deprecate old endpoints" |
| Migration guide | api-versioning-expert | "Create migration guide" |

## Reference Materials

This skill includes comprehensive reference documents in `references/`:

- **graphql-spec-guide.md** — Complete GraphQL spec reference: type system, scalars, enums, interfaces, unions, directives, introspection, execution semantics, and naming conventions
- **api-patterns.md** — Universal API patterns: cursor and offset pagination, filtering, sorting, field selection, error handling (RFC 7807), authentication, file uploads, batch operations, real-time (SSE, WebSocket, webhooks), idempotency, rate limiting, and health checks
- **schema-federation.md** — Apollo Federation v2: all directives (@key, @external, @requires, @provides, @shareable, @override, @inaccessible), entity resolution, subgraph design, gateway/router configuration, schema stitching comparison, composition validation, and monolith-to-federation migration

Agents automatically consult these references when working. You can also read them directly for quick answers.

## How It Works

1. You describe what you need (e.g., "design a GraphQL API for my blog")
2. The SKILL.md routes to the appropriate agent
3. The agent reads your code, discovers your schema and framework, understands relationships
4. Solutions are designed and implemented following best practices
5. The agent provides results and next steps

All generated artifacts follow industry best practices:
- **GraphQL**: Relay spec pagination, proper nullability, DataLoader patterns, typed resolvers
- **REST**: RFC 7807 errors, HATEOAS discoverability, cursor pagination, OpenAPI 3.1
- **Performance**: N+1 prevention, response caching, complexity limits, monitoring
- **Security**: Auth middleware, rate limiting, depth/complexity limits, input validation
- **Versioning**: Breaking change detection, deprecation headers, migration tooling
