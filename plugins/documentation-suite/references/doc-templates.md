# Documentation Templates

Reusable templates for common documentation types. Copy, adapt, and fill in project-specific details.

## README Templates

### Library/Package README

```markdown
# package-name

> One-line description of what this library does and why it exists.

[![npm version](https://img.shields.io/npm/v/package-name.svg)](https://www.npmjs.com/package/package-name)
[![CI](https://github.com/USER/REPO/actions/workflows/ci.yml/badge.svg)](https://github.com/USER/REPO/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](./LICENSE)

## Why package-name?

Explain the problem this library solves. Show a "before" (pain) and
"after" (solution) if possible. Keep it to 2-3 sentences.

## Features

- **Feature One** — Brief explanation of what it does
- **Feature Two** — Brief explanation of what it does
- **Feature Three** — Brief explanation of what it does

## Install

```bash
npm install package-name
```

## Quick Start

```typescript
import { mainFunction } from 'package-name';

const result = await mainFunction({
  option: 'value',
});

console.log(result);
```

## API

### `mainFunction(options)`

Description of what this function does.

**Parameters:**

| Name | Type | Default | Description |
|------|------|---------|-------------|
| `option` | `string` | — | Required. What this controls. |
| `flag` | `boolean` | `false` | Optional. What this toggles. |

**Returns:** `Promise<Result>`

**Example:**

```typescript
const result = await mainFunction({
  option: 'value',
  flag: true,
});
```

### `helperFunction(input)`

Description of this helper.

**Parameters:**

| Name | Type | Description |
|------|------|-------------|
| `input` | `string` | What to process |

**Returns:** `string`

## Configuration

Create a config file `.package-namerc.json`:

```json
{
  "option1": "value",
  "option2": true
}
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `option1` | string | `"default"` | Description |
| `option2` | boolean | `true` | Description |

## FAQ

**Q: Common question about this library?**

A: Clear, helpful answer.

**Q: Another common question?**

A: Clear, helpful answer.

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md) for development setup and guidelines.

## License

[MIT](./LICENSE)
```

### CLI Tool README

```markdown
# tool-name

> One-line description of this CLI tool.

[![CI](https://github.com/USER/REPO/actions/workflows/ci.yml/badge.svg)](https://github.com/USER/REPO/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](./LICENSE)

## Install

```bash
# npm
npm install -g tool-name

# Homebrew
brew install tool-name

# Binary (macOS/Linux)
curl -fsSL https://get.tool-name.dev | sh
```

## Quick Start

```bash
# Initialize
tool-name init my-project

# Run
tool-name run

# Check status
tool-name status
```

## Commands

### `tool-name init [dir]`

Initialize a new project.

```bash
tool-name init my-project
tool-name init my-project --template typescript
tool-name init . --force
```

| Flag | Short | Description |
|------|-------|-------------|
| `--template` | `-t` | Project template to use |
| `--force` | `-f` | Overwrite existing files |
| `--no-git` | — | Skip git initialization |

### `tool-name run [script]`

Run a project script.

```bash
tool-name run          # Run default script
tool-name run build    # Run specific script
tool-name run --watch  # Run with file watching
```

| Flag | Short | Description |
|------|-------|-------------|
| `--watch` | `-w` | Re-run on file changes |
| `--verbose` | `-v` | Show detailed output |

### `tool-name config`

View or modify configuration.

```bash
tool-name config list          # Show all settings
tool-name config set key value # Set a value
tool-name config get key       # Get a value
```

## Configuration

Create `.tool-namerc` in your project root or home directory:

```json
{
  "defaultTemplate": "typescript",
  "verbose": false,
  "parallel": 4
}
```

Environment variables override config file values:

| Variable | Config Key | Description |
|----------|-----------|-------------|
| `TOOL_TEMPLATE` | `defaultTemplate` | Default project template |
| `TOOL_VERBOSE` | `verbose` | Enable verbose output |
| `TOOL_PARALLEL` | `parallel` | Parallel task count |

## Shell Completion

```bash
# Bash
tool-name completion bash >> ~/.bashrc

# Zsh
tool-name completion zsh >> ~/.zshrc

# Fish
tool-name completion fish > ~/.config/fish/completions/tool-name.fish
```

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md).

## License

[MIT](./LICENSE)
```

### Web Application README

```markdown
# App Name

> One-line description.

[![CI](https://github.com/USER/REPO/actions/workflows/ci.yml/badge.svg)](https://github.com/USER/REPO/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](./LICENSE)

[Demo](https://demo.example.com) | [Docs](./docs) | [Changelog](./CHANGELOG.md)

## Features

- **Feature One** — Description
- **Feature Two** — Description
- **Feature Three** — Description

## Getting Started

### Prerequisites

- Node.js 18+
- PostgreSQL 14+

### Setup

```bash
git clone https://github.com/USER/REPO.git
cd REPO
cp .env.example .env    # Edit with your values
npm install
npm run db:migrate
npm run dev
```

Open [http://localhost:3000](http://localhost:3000).

### Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `DATABASE_URL` | Yes | — | PostgreSQL connection string |
| `JWT_SECRET` | Yes | — | JWT signing secret (min 32 chars) |
| `PORT` | No | `3000` | Server port |
| `LOG_LEVEL` | No | `info` | Logging verbosity |

### Docker

```bash
docker compose up
```

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Frontend | React, TypeScript, Tailwind CSS |
| Backend | Node.js, Express |
| Database | PostgreSQL, Prisma |
| Testing | Vitest, Playwright |

## Project Structure

```
src/
├── app/           # Page routes and layouts
├── components/    # UI components
├── lib/           # Shared utilities and helpers
├── server/        # API routes and business logic
└── db/            # Database schema and migrations
```

## Development

```bash
npm run dev        # Start development server
npm test           # Run tests
npm run lint       # Lint code
npm run build      # Production build
npm run db:studio  # Open Prisma Studio
```

## Deployment

See [docs/deployment.md](./docs/deployment.md).

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md).

## License

[MIT](./LICENSE)
```

## Architecture Decision Record Templates

### Standard ADR

```markdown
# ADR-XXXX: [Title]

## Status

Proposed | Accepted | Deprecated | Superseded by [ADR-YYYY](link)

## Date

YYYY-MM-DD

## Context

[Describe the situation, forces, and constraints that led to this decision.
Include technical context, business requirements, and team considerations.]

## Decision

[State the decision clearly. Use active voice: "We will..."]

## Consequences

### Positive
- [Benefit 1]
- [Benefit 2]

### Negative
- [Tradeoff 1 — and how we mitigate it]
- [Tradeoff 2 — and its acceptable impact]

### Neutral
- [Change that is neither good nor bad]

## Alternatives Considered

### [Alternative A]
- Description: [what it is]
- Pros: [advantages]
- Cons: [disadvantages]
- Why rejected: [specific reason]

### [Alternative B]
- Description: [what it is]
- Pros: [advantages]
- Cons: [disadvantages]
- Why rejected: [specific reason]
```

### MADR (Markdown Any Decision Record)

```markdown
# ADR-XXXX: [Short title of solved problem and solution]

## Status

Proposed | Accepted | Deprecated | Superseded by [ADR-YYYY](link)

## Context and Problem Statement

[Describe the context and problem in 2-3 sentences. What needs to be decided?]

## Decision Drivers

* [Driver 1 — e.g., security requirements]
* [Driver 2 — e.g., developer experience]
* [Driver 3 — e.g., performance needs]
* [Driver 4 — e.g., cost constraints]

## Considered Options

1. [Option 1]
2. [Option 2]
3. [Option 3]

## Decision Outcome

Chosen option: **"[Option N]"**, because [justification — reference decision drivers].

### Consequences

* Good, because [positive consequence]
* Good, because [positive consequence]
* Bad, because [negative consequence]
* Neutral, because [neutral consequence]

### Confirmation

[How will you confirm the decision was correctly implemented? Metric, review, test?]

## Pros and Cons of the Options

### [Option 1]

* Good, because [advantage]
* Good, because [advantage]
* Neutral, because [neutral point]
* Bad, because [disadvantage]

### [Option 2]

* Good, because [advantage]
* Bad, because [disadvantage]
* Bad, because [disadvantage]

### [Option 3]

* Good, because [advantage]
* Good, because [advantage]
* Bad, because [disadvantage]

## More Information

[Links to related ADRs, documentation, research, benchmarks, etc.]
```

## Changelog Templates

### Keep a Changelog

```markdown
# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- [New feature description] ([#PR](link))

### Changed
- [Changed behavior description] ([#PR](link))

### Deprecated
- [Deprecated feature] — use [alternative] instead. Will be removed in vX.0.0.

### Removed
- [Removed feature description] ([#PR](link))

### Fixed
- [Bug fix description] ([#PR](link))

### Security
- [Security fix description] ([CVE-XXXX-XXXXX](link))

## [X.Y.Z] - YYYY-MM-DD

### Added
- [Feature description] ([#PR](link))

### Fixed
- [Bug fix description] ([#PR](link))

[Unreleased]: https://github.com/USER/REPO/compare/vX.Y.Z...HEAD
[X.Y.Z]: https://github.com/USER/REPO/compare/vA.B.C...vX.Y.Z
```

### GitHub Release Notes

```markdown
## Highlights

Brief 1-2 sentence summary of the most important change(s).

### New Features
- **Feature Name** — Description of what it does and why it matters. (#PR)
- **Feature Name** — Description. (#PR)

### Improvements
- Description of improvement. (#PR)

### Bug Fixes
- Fixed [description of bug]. (#PR)
- Fixed [description of bug]. (#PR)

### Breaking Changes
- **[What changed]**: [What was the old behavior]. [What is the new behavior]. See [migration guide](link).

### Deprecations
- `oldFunction()` is deprecated. Use `newFunction()` instead. Will be removed in vX.0.0.

---

**Full Changelog**: https://github.com/USER/REPO/compare/vPREV...vCURR
```

## API Documentation Templates

### REST Endpoint

```markdown
### [Action] [Resource]

    [METHOD] /api/v1/[resource]

[One-line description of what this endpoint does.]

**Authentication:** [Required/Optional]. [Scheme and scope.]

**Rate Limit:** [N] requests per [time window].

**Query Parameters:**

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `param` | type | Yes/No | `default` | Description |

**Request Body:**

```json
{
  "field": "value"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `field` | type | Yes/No | Description |

**Response ([status code]):**

```json
{
  "data": {}
}
```

**Error Responses:**

| Status | Code | When |
|--------|------|------|
| 400 | `ERROR_CODE` | Condition |
| 404 | `NOT_FOUND` | Resource doesn't exist |

**Example:**

```bash
curl -X [METHOD] https://api.example.com/v1/[resource] \
  -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"field": "value"}'
```
```

### GraphQL Operation

```markdown
### [Operation Name]

**Type:** Query | Mutation | Subscription

**Authentication:** [Required scope]

```graphql
query OperationName($param: Type!) {
  field(param: $param) {
    id
    name
    nestedField {
      subField
    }
  }
}
```

**Variables:**

| Variable | Type | Required | Description |
|----------|------|----------|-------------|
| `$param` | Type! | Yes | Description |

**Response:**

```json
{
  "data": {
    "field": {
      "id": "123",
      "name": "Example",
      "nestedField": {
        "subField": "value"
      }
    }
  }
}
```

**Errors:**

| Code | Message | When |
|------|---------|------|
| `NOT_FOUND` | "Resource not found" | ID doesn't exist |
```

## Contributing Guide Template

```markdown
# Contributing to [Project Name]

Thank you for contributing! This guide helps you get started.

## Quick Start

```bash
git clone https://github.com/USER/REPO.git
cd REPO
npm install
npm test
```

## Development Workflow

1. Fork the repository
2. Create a branch: `git checkout -b feature/my-feature`
3. Make changes and add tests
4. Run tests: `npm test`
5. Run linter: `npm run lint`
6. Commit: `git commit -m "feat: add my feature"`
7. Push: `git push origin feature/my-feature`
8. Open a Pull Request

## Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: add user avatar upload
fix: resolve pagination bug
docs: update API examples
test: add auth integration tests
chore: update dependencies
```

## Code Style

- Run `npm run lint:fix` before committing
- Follow existing patterns in the codebase
- Add tests for new features
- Update docs if behavior changes

## Reporting Bugs

Include:
- Steps to reproduce
- Expected vs actual behavior
- Environment (OS, Node version, package version)
- Error messages or screenshots

## Questions?

- [GitHub Discussions](link) — questions and ideas
- [Issues](link) — bugs and feature requests
```

## Design Document Template

```markdown
# Design Doc: [Feature Name]

**Author:** [Name]
**Date:** YYYY-MM-DD
**Status:** Draft | In Review | Approved | Implemented
**Approver:** [Name]

## Goal

[What are we trying to achieve? 2-3 sentences.]

## Non-Goals

- [What is explicitly out of scope]
- [What we're not trying to solve]

## Background

[Context the reader needs. Current state, relevant history, why now.]

## Proposed Design

### Overview

[High-level description of the approach.]

### API

[New or changed API surfaces — endpoints, functions, config.]

### Data Model

[Database schema changes, new tables, migrations.]

### Security

[Authentication, authorization, input validation, data protection.]

### Performance

[Expected load, caching strategy, scaling considerations.]

## Alternatives Considered

### [Alternative]
[Description, pros, cons, why not chosen.]

## Testing

[How this will be tested — unit, integration, e2e.]

## Rollout Plan

[How this will be deployed — feature flags, phased rollout, monitoring.]

## Open Questions

- [ ] [Question that needs resolution]
- [ ] [Question that needs resolution]
```

## Migration Guide Template

```markdown
# Migration Guide: v[OLD] to v[NEW]

## Overview

Version [NEW] introduces [brief summary of changes]. This guide covers
all breaking changes with step-by-step migration instructions.

**Estimated time:** [X] minutes for most projects.

## Prerequisites

- [Project Name] v[OLD] or later
- Node.js [VERSION]+
- [Other requirements]

## Breaking Changes

### 1. [Change Title]

**Before (v[OLD]):**

```typescript
// Old code
```

**After (v[NEW]):**

```typescript
// New code
```

**Migration steps:**

1. [Step 1]
2. [Step 2]
3. [Step 3]

### 2. [Change Title]

[Repeat format...]

## Deprecations

| Deprecated | Replacement | Removal Version |
|-----------|-------------|-----------------|
| `oldThing` | `newThing` | v[NEXT_MAJOR] |

## New Features

- **[Feature]** — [Brief description and link to docs]

## Troubleshooting

### [Common error after upgrade]

**Cause:** [Why this happens]
**Fix:** [How to resolve]

## Need Help?

- [Migration support thread](link)
- [Documentation](link)
- [Report an issue](link)
```

## Security Policy Template

```markdown
# Security Policy

## Supported Versions

| Version | Supported |
|---------|-----------|
| 2.x.x | Yes |
| 1.x.x | Security fixes only |
| < 1.0 | No |

## Reporting a Vulnerability

**Do not open a public issue for security vulnerabilities.**

Email security@example.com with:

1. Description of the vulnerability
2. Steps to reproduce
3. Potential impact
4. Suggested fix (if any)

We will:
- Acknowledge receipt within 48 hours
- Provide an initial assessment within 1 week
- Release a fix within 30 days (or sooner for critical issues)
- Credit you in the security advisory (unless you prefer anonymity)

## Security Practices

- Dependencies are audited weekly via `npm audit`
- All PRs require security review for auth/data changes
- We follow OWASP guidelines for web security
- Secrets are never committed to the repository
```

## Code of Conduct Template

```markdown
# Code of Conduct

## Our Pledge

We pledge to make participation in our project a harassment-free experience
for everyone, regardless of age, body size, disability, ethnicity, gender
identity, level of experience, nationality, personal appearance, race,
religion, or sexual identity and orientation.

## Our Standards

**Positive behavior:**
- Using welcoming and inclusive language
- Being respectful of differing viewpoints
- Gracefully accepting constructive criticism
- Focusing on what is best for the community

**Unacceptable behavior:**
- Trolling, insulting/derogatory comments
- Public or private harassment
- Publishing others' private information
- Other conduct which could be considered inappropriate

## Enforcement

Instances of unacceptable behavior may be reported to [email].
All complaints will be reviewed and investigated.

## Attribution

This Code of Conduct is adapted from the
[Contributor Covenant](https://www.contributor-covenant.org/version/2/1/code_of_conduct/).
```

## Template Selection Guide

| You need to... | Use this template |
|----------------|-------------------|
| Document a new library/package | Library/Package README |
| Document a CLI tool | CLI Tool README |
| Document a web application | Web Application README |
| Record an architectural decision | Standard ADR or MADR |
| Track changes between releases | Keep a Changelog |
| Announce a release | GitHub Release Notes |
| Document a REST API endpoint | REST Endpoint template |
| Document a GraphQL operation | GraphQL Operation template |
| Guide contributors | Contributing Guide |
| Design a new feature | Design Document |
| Help users upgrade | Migration Guide |
| Handle security reports | Security Policy |
| Set community standards | Code of Conduct |
