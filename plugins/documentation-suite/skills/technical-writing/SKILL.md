---
name: technical-writing
description: >
  Technical writing patterns for developer documentation, guides, and tutorials.
  Use when writing README files, getting started guides, architecture docs,
  troubleshooting guides, or contributing guidelines.
  Triggers: "write docs", "README", "getting started", "tutorial",
  "architecture doc", "troubleshooting guide", "contributing guide", "ADR".
  NOT for: API reference docs (see api-documentation), code comments, or marketing copy.
version: 1.0.0
allowed-tools: Read, Grep, Glob, Bash
---

# Technical Writing Patterns

## README Template

```markdown
# Project Name

One-sentence description of what this project does and why it exists.

## Quick Start

\`\`\`bash
# Install
npm install project-name

# Configure
cp .env.example .env
# Edit .env with your settings

# Run
npm start
\`\`\`

## Features

- **Feature One** — brief explanation
- **Feature Two** — brief explanation
- **Feature Three** — brief explanation

## Usage

### Basic Example

\`\`\`typescript
import { createClient } from 'project-name';

const client = createClient({ apiKey: process.env.API_KEY });
const result = await client.doThing({ param: 'value' });
console.log(result);
\`\`\`

### Advanced: Custom Configuration

\`\`\`typescript
const client = createClient({
  apiKey: process.env.API_KEY,
  timeout: 30000,
  retries: 3,
  baseUrl: 'https://custom-endpoint.example.com',
});
\`\`\`

## Configuration

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `API_KEY` | Yes | — | Your API key from [dashboard](https://example.com/keys) |
| `TIMEOUT` | No | `10000` | Request timeout in milliseconds |
| `LOG_LEVEL` | No | `"info"` | Logging verbosity: debug, info, warn, error |

## Architecture

\`\`\`
src/
├── client/          # API client and request handling
├── models/          # Data models and validation
├── utils/           # Shared utilities
└── index.ts         # Public API exports
\`\`\`

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md) for development setup and guidelines.

## License

MIT — see [LICENSE](./LICENSE) for details.
```

## Architecture Decision Record (ADR)

```markdown
# ADR-003: Use PostgreSQL over SQLite for persistence

## Status
Accepted — 2026-03-01

## Context
We need a persistence layer for the application. The two candidates are
SQLite (file-based, zero-config) and PostgreSQL (client-server, full-featured).

Our requirements:
- Multiple application instances reading/writing simultaneously
- Full-text search on user-generated content
- JSON column support for flexible metadata
- Production deployment on Heroku

## Decision
We will use PostgreSQL.

## Consequences

### Positive
- Native concurrent access (no file locking issues)
- Full-text search with tsvector/tsquery
- JSONB columns with indexing
- Heroku Postgres add-on with managed backups
- Better tooling for schema migrations (pgmigrate, prisma)

### Negative
- Requires a running database server (even for development)
- More complex local development setup
- Slightly higher operational cost ($5-9/mo on Heroku vs free for SQLite)

### Mitigations
- Docker Compose for local development includes Postgres
- Schema migrations managed via Prisma
```

## Troubleshooting Guide Structure

```markdown
# Troubleshooting

## Connection Refused (ECONNREFUSED)

**Symptom**: Error `connect ECONNREFUSED 127.0.0.1:5432` when starting the app.

**Cause**: PostgreSQL is not running or is running on a different port.

**Fix**:
1. Check if Postgres is running:
   \`\`\`bash
   pg_isready -h localhost -p 5432
   \`\`\`
2. If not running, start it:
   \`\`\`bash
   # macOS (Homebrew)
   brew services start postgresql@16

   # Docker
   docker compose up -d postgres
   \`\`\`
3. Verify the connection string in `.env` matches your Postgres configuration.

---

## Authentication Token Expired

**Symptom**: API returns `401 Unauthorized` with message "Token expired".

**Cause**: JWT token has exceeded its TTL (default: 24 hours).

**Fix**:
1. Re-authenticate to get a fresh token:
   \`\`\`bash
   curl -X POST https://api.example.com/auth/login \\
     -H 'Content-Type: application/json' \\
     -d '{"email": "you@example.com", "password": "..."}'
   \`\`\`
2. Update your environment with the new token.

**Prevention**: Implement token refresh before expiry. See the [Authentication Guide](./auth.md#token-refresh).
```

## Writing Principles

```
1. LEAD WITH THE ANSWER
   Bad:  "In order to understand how authentication works, you first need to..."
   Good: "Authentication uses JWT tokens. Pass them in the Authorization header."

2. SHOW, DON'T TELL
   Bad:  "The configuration is flexible and supports many options."
   Good: [show a configuration example with comments]

3. ONE CONCEPT PER SECTION
   Bad:  "Setup, Configuration, and Deployment" (one heading, three topics)
   Good: Three separate sections with focused content

4. COPY-PASTE READY
   Bad:  "Run the start command with the appropriate flags"
   Good: "npm start --port 3000"

5. ASSUME NOTHING, LINK EVERYTHING
   Bad:  "Set up your database" (how? which one?)
   Good: "Set up PostgreSQL ([installation guide](./postgres-setup.md))"

6. ERROR MESSAGES ARE DOCUMENTATION
   Bad:  throw new Error('Failed')
   Good: throw new Error('Database connection failed: check DATABASE_URL in .env')
```

## Gotchas

1. **README as landing page** — the README is the first thing people see. If it doesn't answer "what is this?" and "how do I use it?" in the first 30 seconds, people leave. Lead with a one-liner description and a quick-start that works.

2. **Outdated code examples** — examples that don't match the current API are worse than no examples. Run examples as part of your test suite (doctest pattern) or version them alongside the code.

3. **Missing prerequisites** — "Run npm start" assumes Node.js is installed, the repo is cloned, and dependencies are installed. List every prerequisite with version requirements and installation links.

4. **Screenshots without alt text or context** — screenshots go stale fast and are inaccessible to screen readers. Prefer text-based examples. If you must use screenshots, add alt text and the date they were captured.

5. **Single format for all audiences** — a getting-started guide, API reference, and architecture doc serve different readers. Don't combine them into one massive document. Separate by audience and link between them.

6. **No versioning on docs** — when the docs describe v2 but users are on v1, they get confused. Pin documentation to software versions, or clearly label which version each page applies to.
