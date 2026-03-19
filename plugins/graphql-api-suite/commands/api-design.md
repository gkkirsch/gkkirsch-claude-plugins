---
name: api-design
description: >
  Quick API design command — analyzes your codebase and helps design, build, review, or optimize
  GraphQL and REST APIs. Routes to the appropriate specialist agent based on your request.
  Triggers: "/api-design", "design api", "graphql schema", "rest api", "api review",
  "optimize graphql", "api versioning", "schema design", "api best practices".
user-invocable: true
argument-hint: "<graphql|rest|optimize|version> [target] [--review] [--federation]"
allowed-tools: Read, Write, Edit, Bash, Glob, Grep
model: sonnet
---

# /api-design Command

One-command API design and optimization. Analyzes your project, identifies the API architecture, and routes to the appropriate specialist agent for schema design, REST endpoint creation, performance optimization, or versioning strategy.

## Usage

```
/api-design                           # Auto-detect API type and suggest improvements
/api-design graphql                   # Design or review GraphQL schema
/api-design graphql --federation      # Design federated GraphQL architecture
/api-design rest                      # Design or review REST API endpoints
/api-design rest /api/users           # Design REST endpoints for a specific resource
/api-design optimize                  # Optimize existing GraphQL API performance
/api-design optimize --n1             # Detect and fix N+1 query problems
/api-design version                   # Plan API versioning strategy
/api-design version --breaking        # Analyze breaking changes
/api-design --review                  # Full API design review
```

## Subcommands

### `graphql` — GraphQL Schema Design

Designs or reviews GraphQL schemas, resolvers, subscriptions, and DataLoader patterns.

**What it does:**
1. Scans for existing `.graphql` files, SDL definitions, and GraphQL server setup
2. Identifies the framework (Apollo, Yoga, Mercurius, Pothos, Nexus, TypeGraphQL)
3. Analyzes types, resolvers, and relationships
4. Suggests or generates schema improvements

**Flags:**
- `--federation` — Design for Apollo Federation (subgraph architecture)
- `--subscriptions` — Focus on real-time subscription setup
- `--codegen` — Generate TypeScript types from schema
- `--review` — Review existing schema for best practices

**Examples:**
```
/api-design graphql                            # Analyze and improve schema
/api-design graphql --federation               # Design federated subgraphs
/api-design graphql --subscriptions            # Add subscription support
/api-design graphql User Post Comment          # Design schema for these entities
```

**Routes to:** `graphql-architect` agent

### `rest` — REST API Design

Designs or reviews REST API endpoints with proper resource modeling, HATEOAS, pagination, and error handling.

**What it does:**
1. Scans for existing route definitions and middleware
2. Identifies the framework (Express, Fastify, Koa, Hono, NestJS)
3. Analyzes endpoint patterns and response formats
4. Suggests or generates improvements following REST best practices

**Flags:**
- `--hateoas` — Include HATEOAS links in responses
- `--openapi` — Generate or update OpenAPI 3.1 spec
- `--hal` — Use HAL response format
- `--jsonapi` — Use JSON:API response format

**Examples:**
```
/api-design rest                               # Analyze and improve REST API
/api-design rest /api/users                    # Design endpoints for users resource
/api-design rest --openapi                     # Generate OpenAPI documentation
/api-design rest --hateoas /api/orders         # Add HATEOAS to orders endpoints
```

**Routes to:** `rest-api-designer` agent

### `optimize` — Performance Optimization

Analyzes and optimizes GraphQL API performance, focusing on N+1 prevention, caching, and query efficiency.

**What it does:**
1. Scans all resolver files for N+1 patterns
2. Checks DataLoader implementation and usage
3. Reviews caching configuration
4. Analyzes query complexity and depth limits
5. Provides prioritized optimization recommendations

**Flags:**
- `--n1` — Focus specifically on N+1 detection and DataLoader setup
- `--cache` — Focus on caching strategy
- `--audit` — Full performance audit with metrics
- `--federation` — Optimize federated subgraph performance

**Examples:**
```
/api-design optimize                           # Full optimization analysis
/api-design optimize --n1                      # Detect and fix N+1 problems
/api-design optimize --cache                   # Review and improve caching
/api-design optimize --audit                   # Complete performance audit
```

**Routes to:** `schema-optimizer` agent

### `version` — API Versioning

Plans and implements API versioning strategies, breaking change management, and deprecation workflows.

**What it does:**
1. Analyzes current API versioning (if any)
2. Identifies potential breaking changes
3. Recommends versioning strategy
4. Creates migration plans and deprecation timelines

**Flags:**
- `--breaking` — Analyze schema changes for breaking changes
- `--deprecate <field>` — Create deprecation plan for a field/endpoint
- `--migrate` — Generate migration guide between versions
- `--sunset` — Plan sunset timeline for old version

**Examples:**
```
/api-design version                            # Recommend versioning strategy
/api-design version --breaking                 # Check for breaking changes
/api-design version --deprecate name           # Plan deprecation of 'name' field
/api-design version --migrate v1 v2            # Generate v1→v2 migration guide
```

**Routes to:** `api-versioning-expert` agent

## Auto-Detection

When no subcommand is specified, `/api-design` auto-detects your API type:

1. **GraphQL detected** (`.graphql` files, `typeDefs`, GraphQL server imports) → Routes to `graphql-architect`
2. **REST detected** (route definitions, Express/Fastify/Koa routes) → Routes to `rest-api-designer`
3. **Both detected** → Asks which aspect to focus on
4. **Neither detected** → Asks what type of API you're building

## Agent Selection Guide

| Need | Agent | Command |
|------|-------|---------|
| Design GraphQL schema | graphql-architect | `/api-design graphql` |
| Write resolvers | graphql-architect | `/api-design graphql` |
| Set up subscriptions | graphql-architect | `/api-design graphql --subscriptions` |
| Design REST endpoints | rest-api-designer | `/api-design rest` |
| Add HATEOAS links | rest-api-designer | `/api-design rest --hateoas` |
| Generate OpenAPI spec | rest-api-designer | `/api-design rest --openapi` |
| Fix N+1 queries | schema-optimizer | `/api-design optimize --n1` |
| Add response caching | schema-optimizer | `/api-design optimize --cache` |
| Performance audit | schema-optimizer | `/api-design optimize --audit` |
| Federation design | graphql-architect | `/api-design graphql --federation` |
| Optimize federation | schema-optimizer | `/api-design optimize --federation` |
| Plan versioning | api-versioning-expert | `/api-design version` |
| Breaking change check | api-versioning-expert | `/api-design version --breaking` |
| Deprecation workflow | api-versioning-expert | `/api-design version --deprecate` |
| Full API review | All agents | `/api-design --review` |

## Reference Materials

This suite includes comprehensive reference documents in `references/`:

- **graphql-spec-guide.md** — GraphQL type system, directives, custom scalars, execution semantics, query/mutation/subscription design patterns
- **api-patterns.md** — Pagination, filtering, error handling, authentication, file uploads, batch operations, real-time patterns, idempotency
- **schema-federation.md** — Apollo Federation v2 directives, subgraph design, entity resolution, gateway configuration, schema stitching, monolith-to-federation migration

Agents automatically consult these references when working. You can also read them directly for quick answers.

## How It Works

1. You describe what you need (e.g., "design a GraphQL schema for my blog")
2. The command analyzes your project structure and existing API code
3. It routes to the appropriate specialist agent
4. The agent reads your code, understands your domain, and generates solutions
5. Code is written following best practices with proper patterns

All generated code follows these principles:
- **GraphQL**: Relay connection spec, proper nullability, DataLoader, typed resolvers
- **REST**: RFC 7807 errors, HATEOAS links, cursor pagination, OpenAPI 3.1
- **Security**: Auth middleware, rate limiting, input validation, query depth/complexity limits
- **Testing**: Schema validation tests, resolver unit tests, integration tests
