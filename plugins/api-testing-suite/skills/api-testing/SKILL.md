---
name: api-testing-suite
description: >
  API Testing & Documentation Suite — AI-powered toolkit for HTTP testing, OpenAPI documentation,
  load testing, and API security auditing. Generates comprehensive test suites, writes OpenAPI 3.1
  specs, runs performance benchmarks with k6/Artillery, and audits for OWASP API Top 10 vulnerabilities.
  Triggers: "api test", "test api", "test endpoint", "api testing", "http test", "rest test",
  "graphql test", "openapi", "swagger", "api docs", "api documentation", "api reference",
  "load test", "performance test", "stress test", "benchmark api", "k6", "artillery",
  "api security", "security audit", "owasp", "api vulnerability", "penetration test api",
  "rate limiting", "api auth test", "jwt test", "oauth test".
  Dispatches the appropriate specialist agent: api-tester, api-doc-writer, load-tester,
  or api-security.
  NOT for: Frontend/UI testing, browser automation, database testing without API layer,
  or network infrastructure security.
version: 1.0.0
argument-hint: "<test|doc|load|security> [endpoint]"
user-invocable: true
allowed-tools: Read, Grep, Glob, Bash
model: sonnet
---

# API Testing & Documentation Suite

Production-grade API testing and documentation agents for Claude Code. Four specialist agents that handle HTTP testing, API documentation, load testing, and security auditing — the API quality work that every backend project needs.

## Available Agents

### API Tester (`api-tester`)
Creates comprehensive API test suites. HTTP request builders, response assertions, authentication flows, error scenario coverage, mock servers, and contract testing. Supports REST, GraphQL, gRPC, and WebSocket APIs.

**Invoke**: Dispatch via Task tool with `subagent_type: "api-tester"`.

**Example prompts**:
- "Generate a complete test suite for my Express REST API"
- "Write integration tests for the authentication endpoints"
- "Create contract tests between my frontend and API"
- "Test my GraphQL mutations with edge cases and error scenarios"

### API Documentation Writer (`api-doc-writer`)
Generates comprehensive API documentation. OpenAPI 3.1 specs from code analysis, markdown API references, endpoint documentation with examples, changelog generation, and SDK documentation.

**Invoke**: Dispatch via Task tool with `subagent_type: "api-doc-writer"`.

**Example prompts**:
- "Generate an OpenAPI 3.1 spec from my Express routes"
- "Write a complete API reference in markdown"
- "Create API documentation with request/response examples for every endpoint"
- "Generate a changelog from my recent API changes"

### Load Tester (`load-tester`)
Performance and load testing for APIs. Creates k6 and Artillery test scripts, simulates realistic traffic patterns, identifies bottlenecks, generates performance reports, and provides optimization recommendations.

**Invoke**: Dispatch via Task tool with `subagent_type: "load-tester"`.

**Example prompts**:
- "Load test my API with 500 concurrent users"
- "Create a k6 script that simulates realistic user behavior on my e-commerce API"
- "Benchmark my search endpoint — it feels slow under load"
- "Run a stress test to find the breaking point of my order processing pipeline"

### API Security Auditor (`api-security`)
Security auditing for APIs. Checks for OWASP API Top 10 vulnerabilities, authentication/authorization flaws, injection attacks, rate limiting gaps, CORS misconfigurations, and provides remediation guidance.

**Invoke**: Dispatch via Task tool with `subagent_type: "api-security"`.

**Example prompts**:
- "Run a security audit on my API — check for OWASP Top 10 issues"
- "Test my authentication flow for vulnerabilities"
- "Check my API for injection vulnerabilities and data exposure"
- "Audit the rate limiting and CORS configuration on my endpoints"

## Quick Start: /test-api

Use the `/test-api` command for guided API testing and documentation:

```
/test-api                        # Auto-discover and test all endpoints
/test-api test /api/users        # Test specific endpoint
/test-api doc                    # Generate OpenAPI documentation
/test-api doc --format markdown  # Generate markdown API reference
/test-api load /api/orders       # Load test specific endpoint
/test-api security               # Full security audit
```

The `/test-api` command auto-detects your framework, discovers endpoints, and routes to the right agent.

## Agent Selection Guide

| Need | Agent | Command |
|------|-------|---------|
| Test API endpoints | api-tester | "Write tests for my API" |
| Test auth flows | api-tester | "Test my login/signup endpoints" |
| Contract testing | api-tester | "Create contract tests" |
| GraphQL testing | api-tester | "Test my GraphQL API" |
| Generate OpenAPI spec | api-doc-writer | "Generate OpenAPI docs" |
| API reference docs | api-doc-writer | "Write API reference" |
| Endpoint documentation | api-doc-writer | "Document my endpoints" |
| Performance testing | load-tester | "Load test my API" |
| Find bottlenecks | load-tester | "Benchmark my slow endpoint" |
| Stress testing | load-tester | "Find breaking point" |
| Security audit | api-security | "Audit API security" |
| Auth vulnerabilities | api-security | "Test auth for vulnerabilities" |
| OWASP compliance | api-security | "Check OWASP API Top 10" |
| Rate limit testing | api-security | "Test rate limiting" |

## Reference Materials

This skill includes comprehensive reference documents in `references/`:

- **http-testing-patterns.md** — REST/GraphQL testing patterns, request builders, assertion strategies, authentication flows, pagination testing, error handling patterns, mock servers, and contract testing
- **openapi-reference.md** — Complete OpenAPI 3.1 specification guide, schema definitions, examples, validation rules, code generation, security schemes, and webhook documentation
- **api-security-checklist.md** — Comprehensive OWASP API Top 10 security checklist, JWT/OAuth2 testing, injection prevention, rate limiting, CORS configuration, and security header requirements

Agents automatically consult these references when working. You can also read them directly for quick answers.

## How It Works

1. You describe what you need (e.g., "test my API endpoints")
2. The SKILL.md routes to the appropriate agent
3. The agent reads your code, discovers endpoints, understands auth patterns, and detects the framework
4. Tests, docs, or reports are generated and written to your project
5. The agent provides results and next steps

All generated artifacts follow industry best practices:
- Testing: Comprehensive coverage, proper assertions, isolated tests, CI-ready
- Documentation: OpenAPI 3.1 compliant, real examples, proper schemas
- Performance: Realistic load patterns, statistical analysis, actionable insights
- Security: OWASP-aligned, severity-ranked findings, remediation guidance
