---
name: documentation-suite
description: >
  Documentation & Technical Writing Suite — AI-powered documentation toolkit for API documentation
  with OpenAPI/Swagger, README architecture and project docs, Architecture Decision Records and
  design docs, and changelog/release note management. Generates production-quality documentation
  from codebase analysis.
  Triggers: "write docs", "documentation", "generate docs", "create readme", "api docs", "openapi",
  "swagger", "api reference", "endpoint documentation", "readme", "project docs", "badges",
  "adr", "architecture decision", "decision record", "rfc", "design doc", "technical spec",
  "changelog", "release notes", "conventional commits", "migration guide", "contributing guide",
  "write documentation", "document this project", "document my api".
  Dispatches the appropriate specialist agent: api-doc-writer, readme-architect, adr-specialist,
  or changelog-manager.
  NOT for: Code generation without documentation, blog posts, marketing copy, or non-technical writing.
version: 1.0.0
argument-hint: "<api|readme|adr|changelog> [path-or-topic]"
user-invocable: true
allowed-tools: Read, Grep, Glob, Bash
model: sonnet
---

# Documentation & Technical Writing Suite

Production-grade documentation agents for Claude Code. Four specialist agents that handle API documentation, README architecture, Architecture Decision Records, and changelog management — the documentation work that every project needs.

## Available Agents

### API Documentation Writer (`api-doc-writer`)
Generates comprehensive API documentation from codebase analysis. OpenAPI 3.1 specifications, markdown API references, endpoint documentation with request/response examples, authentication guides, and SDK documentation. Supports Express, FastAPI, Django REST Framework, Spring Boot, Gin, and more.

**Invoke**: Dispatch via Task tool with `subagent_type: "api-doc-writer"`.

**Example prompts**:
- "Generate an OpenAPI 3.1 spec from my Express routes"
- "Write a complete API reference in markdown"
- "Document all endpoints with request/response examples"
- "Create SDK documentation for my REST API"

### README Architect (`readme-architect`)
Designs and writes outstanding README files, contributing guides, and project documentation. Supports library, CLI, web app, and API service templates with badges, installation instructions, examples, architecture diagrams, and FAQ sections.

**Invoke**: Dispatch via Task tool with `subagent_type: "readme-architect"`.

**Example prompts**:
- "Write a professional README for this project"
- "Update the README with new features and badges"
- "Create a CONTRIBUTING.md with development setup"
- "Generate a README for my npm package"

### ADR Specialist (`adr-specialist`)
Creates Architecture Decision Records, RFCs, technical design documents, and technology evaluation reports. Supports Nygard, MADR, and Y-statement formats with structured alternatives analysis and trade-off documentation.

**Invoke**: Dispatch via Task tool with `subagent_type: "adr-specialist"`.

**Example prompts**:
- "Create an ADR for choosing between REST and GraphQL"
- "Write an RFC for migrating to a new database"
- "Document the decision to use JWT authentication"
- "Create a technology evaluation for state management"

### Changelog Manager (`changelog-manager`)
Generates and maintains changelogs from git history. Supports Keep a Changelog, GitHub Release Notes, user-facing announcements, and stakeholder summaries. Implements Conventional Commits classification and semantic versioning.

**Invoke**: Dispatch via Task tool with `subagent_type: "changelog-manager"`.

**Example prompts**:
- "Generate a changelog from recent commits"
- "Write release notes for version 2.0"
- "Create a migration guide for breaking changes"
- "Generate a user-facing changelog for the marketing team"

## Quick Start: /write-docs

Use the `/write-docs` command for guided documentation generation:

```
/write-docs                           # Auto-detect what docs are needed
/write-docs api                       # Generate OpenAPI documentation
/write-docs api --format markdown     # Markdown API reference
/write-docs readme                    # Generate README.md
/write-docs adr "Choose database"     # Architecture Decision Record
/write-docs changelog                 # Changelog from git history
/write-docs changelog --since v1.0.0  # Changelog since specific tag
/write-docs design "Feature name"     # Technical design document
/write-docs contributing              # CONTRIBUTING.md
/write-docs migration v1 v2           # Version migration guide
```

The `/write-docs` command auto-detects your project type, framework, and existing documentation.

## Agent Selection Guide

| Need | Agent | Command |
|------|-------|---------|
| OpenAPI spec from code | api-doc-writer | "Generate OpenAPI docs" |
| Markdown API reference | api-doc-writer | "Write API reference" |
| Endpoint documentation | api-doc-writer | "Document my endpoints" |
| Authentication guide | api-doc-writer | "Document auth flow" |
| SDK documentation | api-doc-writer | "Write SDK docs" |
| Project README | readme-architect | "Write a README" |
| Contributing guide | readme-architect | "Create CONTRIBUTING.md" |
| Badge setup | readme-architect | "Add badges to README" |
| Architecture decision | adr-specialist | "Create an ADR" |
| Design document | adr-specialist | "Write a design doc" |
| RFC / proposal | adr-specialist | "Write an RFC" |
| Tech evaluation | adr-specialist | "Evaluate technologies" |
| Changelog | changelog-manager | "Generate changelog" |
| Release notes | changelog-manager | "Write release notes" |
| Migration guide | changelog-manager | "Create migration guide" |
| Version bump | changelog-manager | "What version should this be?" |

## Reference Materials

This skill includes comprehensive reference documents in `references/`:

- **openapi-spec.md** — Complete OpenAPI 3.1 specification reference with schemas, security schemes, parameters, webhooks, polymorphism, code generation, and validation tools
- **writing-style-guide.md** — Technical writing best practices covering voice, tone, formatting, word choice, audience targeting, and documentation maintenance
- **doc-templates.md** — Reusable templates for READMEs, ADRs, changelogs, API docs, contributing guides, design documents, migration guides, and security policies

Agents automatically consult these references when working. You can also read them directly for quick answers.

## How It Works

1. You describe what documentation you need
2. The SKILL.md routes to the appropriate agent
3. The agent reads your codebase — routes, schemas, models, middleware, git history
4. Documentation is generated following industry standards
5. Files are written to the appropriate location in your project

All generated documentation follows best practices:
- API docs: OpenAPI 3.1 compliant, real examples, complete schemas
- READMEs: Project-type-appropriate sections, working code examples, badges
- ADRs: Structured decisions with alternatives and consequences
- Changelogs: Keep a Changelog format, conventional commit classification
