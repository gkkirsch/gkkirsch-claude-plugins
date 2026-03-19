---
name: write-docs
description: >
  Documentation writing command — generates API documentation, READMEs, Architecture Decision Records,
  changelogs, and technical design documents. Analyzes your codebase and produces professional-grade
  documentation in the appropriate format.
  Triggers: "/write-docs", "write documentation", "generate docs", "create readme", "write api docs",
  "create adr", "generate changelog", "document this", "write release notes".
user-invocable: true
argument-hint: "<api|readme|adr|changelog> [path-or-options] [--format openapi|markdown] [--audience dev|user|stakeholder]"
allowed-tools: Read, Write, Edit, Bash, Glob, Grep
model: sonnet
---

# /write-docs Command

One-command documentation generation. Analyzes your project and produces professional documentation.

## Usage

```
/write-docs                           # Auto-detect what docs are needed
/write-docs api                       # Generate API documentation (OpenAPI 3.1)
/write-docs api --format markdown     # Generate API docs in markdown
/write-docs readme                    # Generate or update README.md
/write-docs adr "Choose database"     # Create an Architecture Decision Record
/write-docs changelog                 # Generate changelog from git history
/write-docs changelog --since v1.2.0  # Changelog since specific tag
/write-docs design "Feature name"     # Create a technical design document
/write-docs contributing              # Generate CONTRIBUTING.md
/write-docs migration v1 v2           # Create migration guide
```

## Subcommands

### `api` — API Documentation

Generates API documentation by analyzing your route handlers, middleware, validation schemas, and response types.

**Default output:** OpenAPI 3.1 YAML specification

**Options:**
| Flag | Description |
|------|-------------|
| `--format openapi` | OpenAPI 3.1 YAML (default) |
| `--format markdown` | Markdown API reference |
| `--format json` | OpenAPI 3.1 JSON |
| `--output <path>` | Output file path |
| `--include-internal` | Include internal/admin endpoints |

**Agent:** Dispatches `api-doc-writer` agent via Task tool with `subagent_type: "api-doc-writer"`.

**Examples:**
- `/write-docs api` — Generate OpenAPI spec from Express routes
- `/write-docs api --format markdown` — Generate markdown API reference
- `/write-docs api --output docs/openapi.yaml` — Write to specific path
- "Generate API documentation for my REST endpoints"
- "Create an OpenAPI spec from my FastAPI routes"
- "Document all my API endpoints with examples"

### `readme` — README Generation

Analyzes the project and generates a comprehensive README with appropriate sections for the project type.

**Agent:** Dispatches `readme-architect` agent via Task tool with `subagent_type: "readme-architect"`.

**Options:**
| Flag | Description |
|------|-------------|
| `--type library` | Library/package README template |
| `--type cli` | CLI tool README template |
| `--type webapp` | Web application README template |
| `--type api` | API service README template |
| `--update` | Update existing README (preserve custom sections) |

**Examples:**
- `/write-docs readme` — Auto-detect project type and generate README
- `/write-docs readme --type library` — Force library README template
- `/write-docs readme --update` — Update existing README with new info
- "Write a README for this project"
- "Update the README with the new API endpoints"
- "Create a professional README with badges and examples"

### `adr` — Architecture Decision Records

Creates structured ADRs documenting architectural decisions with context, alternatives, and consequences.

**Agent:** Dispatches `adr-specialist` agent via Task tool with `subagent_type: "adr-specialist"`.

**Options:**
| Flag | Description |
|------|-------------|
| `--format standard` | Nygard format (default) |
| `--format madr` | MADR format |
| `--format y-statement` | Y-statement format |
| `--dir <path>` | ADR directory (default: `docs/decisions/`) |

**Examples:**
- `/write-docs adr "Choose authentication strategy"` — Create ADR for auth decision
- `/write-docs adr "Database selection" --format madr` — MADR format
- "Create an ADR for choosing between REST and GraphQL"
- "Document the decision to use PostgreSQL"
- "Write a design decision record for our caching strategy"

### `changelog` — Changelog Generation

Generates changelogs from git commit history, classifying changes by type and determining version bumps.

**Agent:** Dispatches `changelog-manager` agent via Task tool with `subagent_type: "changelog-manager"`.

**Options:**
| Flag | Description |
|------|-------------|
| `--since <tag>` | Generate since this git tag |
| `--format keepachangelog` | Keep a Changelog format (default) |
| `--format github` | GitHub Release Notes format |
| `--format user` | User-facing release notes |
| `--audience dev` | Developer audience (default) |
| `--audience user` | End-user audience |
| `--audience stakeholder` | Executive summary |

**Examples:**
- `/write-docs changelog` — Generate changelog from last tag to HEAD
- `/write-docs changelog --since v1.0.0` — Changelog since v1.0.0
- `/write-docs changelog --format github` — GitHub release notes format
- "Generate a changelog from recent commits"
- "Write release notes for the latest version"
- "Create a user-facing changelog for the marketing team"

### `design` — Technical Design Documents

Creates structured design documents for features, with API designs, data models, and rollout plans.

**Agent:** Dispatches `adr-specialist` agent via Task tool with `subagent_type: "adr-specialist"`.

**Examples:**
- `/write-docs design "User notifications system"` — Create design doc
- "Write a design document for the search feature"
- "Create a technical spec for the payment integration"

### `contributing` — Contributing Guide

Generates CONTRIBUTING.md with development setup, coding standards, and PR guidelines.

**Agent:** Dispatches `readme-architect` agent via Task tool with `subagent_type: "readme-architect"`.

**Examples:**
- `/write-docs contributing` — Generate CONTRIBUTING.md
- "Create a contributing guide for this project"

### `migration` — Migration Guide

Creates step-by-step migration guides for version upgrades with code examples and troubleshooting.

**Agent:** Dispatches `changelog-manager` agent via Task tool with `subagent_type: "changelog-manager"`.

**Examples:**
- `/write-docs migration v1 v2` — Migration guide from v1 to v2
- "Write a migration guide for the v2 breaking changes"

## Auto-Detection

When run without a subcommand, the command analyzes the project and suggests what documentation is needed:

1. **No README.md?** → Suggest generating one
2. **API routes but no OpenAPI spec?** → Suggest API documentation
3. **No CHANGELOG.md?** → Suggest generating from git history
4. **No docs/decisions/?** → Suggest starting an ADR practice
5. **No CONTRIBUTING.md?** → Suggest generating one

## Agent Selection

| Need | Agent | Trigger |
|------|-------|---------|
| OpenAPI/Swagger docs | api-doc-writer | "api docs", "openapi", "swagger" |
| Markdown API reference | api-doc-writer | "api reference", "endpoint docs" |
| README files | readme-architect | "readme", "project docs" |
| Contributing guide | readme-architect | "contributing", "contributor guide" |
| Architecture decisions | adr-specialist | "adr", "decision record", "rfc" |
| Design documents | adr-specialist | "design doc", "technical spec" |
| Changelogs | changelog-manager | "changelog", "release notes" |
| Migration guides | changelog-manager | "migration", "upgrade guide" |

## Reference Materials

This skill includes comprehensive reference documents:

- **openapi-spec.md** — Complete OpenAPI 3.1 specification reference
- **writing-style-guide.md** — Technical writing best practices and conventions
- **doc-templates.md** — Reusable templates for all documentation types

Agents automatically consult these references when working.
