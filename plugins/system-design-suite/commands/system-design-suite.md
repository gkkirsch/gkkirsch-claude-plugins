---
name: system-design-suite
description: >
  System design and architecture command — get expert guidance on distributed systems,
  scalability, reliability, and API design. Dispatches specialized agents for design reviews,
  architecture planning, and system design interview preparation.
  Triggers: "/system-design", "system design", "design system", "architecture review",
  "distributed system", "scalability", "reliability", "api design".
user-invocable: true
argument-hint: "<architect|scale|reliability|api> [topic-or-system]"
allowed-tools: Read, Write, Edit, Bash, Glob, Grep
model: sonnet
---

# /system-design Command

One-command system design and architecture guidance. Routes to the appropriate specialist agent based on your need.

## Usage

```
/system-design                          # General architecture guidance
/system-design architect url-shortener  # Design a URL shortener system
/system-design scale my-api             # Scalability review of your codebase
/system-design reliability              # Reliability audit of current system
/system-design api                      # API design review
```

## Subcommands

| Subcommand | Agent | Description |
|------------|-------|-------------|
| `architect` | system-design-architect | Distributed systems design, CAP theorem, consistency, partitioning |
| `scale` | scalability-engineer | Horizontal/vertical scaling, caching, load balancing, database sharding |
| `reliability` | reliability-engineer | Fault tolerance, circuit breakers, chaos engineering, disaster recovery |
| `api` | api-design-expert | REST, gRPC, GraphQL, API versioning, rate limiting, authentication |

## Procedure

1. Parse the subcommand and topic from user input
2. If no subcommand, analyze the request and route to the best-fit agent
3. Dispatch the appropriate agent via Task tool with the user's context
4. Present findings, diagrams, and recommendations

## Notes

- Agents reference materials in `references/` automatically
- Each agent produces actionable recommendations, not just theory
- For system design interviews, use `architect` with a system name
- For production systems, agents analyze your actual codebase
