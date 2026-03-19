# Technical Writing Style Guide

Comprehensive style guide for writing clear, effective technical documentation. These principles apply to all documentation types — API references, READMEs, tutorials, guides, and architecture documents.

## Core Principles

### 1. Audience First

Every sentence should serve the reader. Before writing, establish:

- **Who is reading?** (junior dev, senior architect, product manager, end user)
- **What do they know?** (prerequisites, assumed knowledge)
- **What do they need?** (quick answer, deep understanding, step-by-step)
- **What will they do next?** (implement, evaluate, troubleshoot)

| Audience | Tone | Detail Level | Examples |
|----------|------|-------------|----------|
| Beginners | Welcoming, patient | High — explain everything | Full code with comments |
| Intermediate devs | Direct, informative | Medium — explain the non-obvious | Focused snippets |
| Senior engineers | Concise, technical | Low — just the facts | API signatures, config |
| Non-technical | Plain language | High — no jargon | Screenshots, analogies |

### 2. Clarity Over Cleverness

```markdown
# Bad
The system leverages a sophisticated event-driven paradigm to facilitate
real-time data synchronization across heterogeneous distributed nodes.

# Good
The system uses events to sync data between servers in real time.
```

Rules:
- Use simple words when simple words work
- One idea per sentence
- One topic per paragraph
- If a sentence needs to be read twice, rewrite it

### 3. Show, Don't Tell

```markdown
# Bad
Our API is easy to use and very flexible.

# Good
Create a user with three lines of code:

    const client = new ApiClient({ key: 'your-key' });
    const user = await client.users.create({ email: 'jane@example.com' });
    console.log(user.id); // "usr_abc123"
```

### 4. Be Specific, Not Vague

```markdown
# Bad
The function accepts various parameters and returns data.

# Good
`listUsers(options)` accepts a page number (1-indexed), page size (1-100),
and optional role filter. Returns a paginated array of User objects with
a total count.
```

### 5. Progressive Disclosure

Structure information from simple to complex:

1. **Title** — What is this? (1 line)
2. **Summary** — Why should I care? (1 paragraph)
3. **Quick Start** — How do I use it? (5-line example)
4. **Configuration** — How do I customize it? (options table)
5. **Advanced** — What about edge cases? (detailed sections)
6. **Reference** — Complete API surface (exhaustive)

## Voice and Tone

### Active Voice (Preferred)

```markdown
# Passive (avoid)
The configuration file is read by the server on startup.
Errors should be caught by the calling function.
The request will be validated by the middleware.

# Active (preferred)
The server reads the configuration file on startup.
The calling function should catch errors.
The middleware validates the request.
```

### Second Person (Preferred for Guides)

```markdown
# Third person (formal, distancing)
The developer should configure their API key before making requests.

# First person plural (acceptable)
We need to configure our API key before making requests.

# Second person (best for instructions)
Configure your API key before making requests.
```

### Imperative Mood (Preferred for Instructions)

```markdown
# Declarative (weaker)
You should run the migration script.
The next step is to install the dependencies.

# Imperative (stronger)
Run the migration script.
Install the dependencies.
```

### Present Tense (Preferred)

```markdown
# Future tense (avoid)
The function will return a promise that will resolve with the user object.

# Present tense (preferred)
The function returns a promise that resolves with the user object.
```

## Sentence Structure

### Short Sentences for Instructions

```markdown
# Too long
Before you can start using the API, you need to create an account on our
website, navigate to the settings page, and then generate an API key which
you should store securely.

# Better
1. Create an account at example.com
2. Go to Settings → API Keys
3. Click "Generate New Key"
4. Copy the key and store it securely — it won't be shown again
```

### One Idea Per Sentence

```markdown
# Overloaded
The function validates the input against the schema, transforms the data,
sends it to the API, and returns the parsed response.

# Clearer
The function performs four steps:
1. Validates input against the schema
2. Transforms the data to the API format
3. Sends the request
4. Returns the parsed response
```

### Parallel Structure

```markdown
# Inconsistent
The system can:
- Processing data in real time
- To validate incoming requests
- Store results in the database
- Users can query historical data

# Parallel
The system can:
- Process data in real time
- Validate incoming requests
- Store results in the database
- Query historical data
```

## Word Choice

### Prefer Simple Words

| Instead of | Use |
|-----------|-----|
| utilize | use |
| leverage | use |
| facilitate | help, enable |
| implement | build, create |
| terminate | stop, end |
| initiate | start |
| modify | change |
| execute | run |
| obtain | get |
| commence | begin |
| subsequent | next |
| prior to | before |
| in order to | to |
| due to the fact that | because |
| at this point in time | now |
| in the event that | if |
| a large number of | many |
| in the majority of cases | usually |
| has the ability to | can |
| it is necessary to | you must |

### Technical Term Consistency

Pick one term and use it everywhere:

| Don't alternate between | Pick one |
|------------------------|----------|
| endpoint / route / path | endpoint |
| request / call / invocation | request |
| response / result / return value | response |
| parameter / argument / option | parameter |
| field / property / attribute | field |
| error / exception / failure | error |
| repository / repo / codebase | repository |
| directory / folder | directory |

### Avoid Weasel Words

```markdown
# Vague
The API is quite fast and handles fairly large datasets reasonably well.
It's pretty simple to set up and somewhat easy to debug.

# Specific
The API handles 10,000 requests/second and processes datasets up to 1GB.
Setup requires 3 commands. All errors include request IDs for debugging.
```

### Avoid Absolutist Claims

```markdown
# Dangerous
This approach always works and never fails.
The best way to handle authentication.
This will definitely solve your problem.

# Accurate
This approach works for most standard configurations.
A recommended approach for handling authentication.
This should resolve the issue — if not, see Troubleshooting.
```

## Formatting Conventions

### Headings

```markdown
# Document Title (H1 — one per document)
## Major Section (H2)
### Subsection (H3)
#### Detail (H4 — use sparingly)
```

Rules:
- Use sentence case for headings: "API reference" not "API Reference"
- Don't skip heading levels (H1 → H3)
- Don't use headings for emphasis — use bold instead
- Keep headings short and descriptive

### Code Formatting

```markdown
Inline code: Use `backticks` for code references in text.

# Use inline code for:
- Function names: `createUser()`
- Variable names: `apiKey`
- File names: `config.json`
- CLI commands: `npm install`
- Values: `true`, `null`, `200`
- Key names: `Authorization`

# Don't use inline code for:
- Product names: GitHub (not `GitHub`)
- General technical concepts: REST API (not `REST API`)
- Emphasis: **important** (not `important`)
```

Code blocks:

````markdown
# Always specify the language for syntax highlighting
```typescript
const result = await client.users.list({ page: 1 });
```

# Use shell/bash for command-line examples
```bash
npm install @example/sdk
```

# Use plain text or no language for output
```
Success: User created (id: usr_abc123)
```

# For terminal sessions, show prompt and output
```bash
$ npm test
✓ 42 tests passed (1.2s)
```
````

### Lists

```markdown
# Use bullet lists for unordered items (features, options)
- Feature A
- Feature B
- Feature C

# Use numbered lists for ordered steps
1. Install the package
2. Configure your API key
3. Make your first request

# Use description lists for term/definition pairs (in tables)
| Term | Definition |
|------|-----------|
| JWT | JSON Web Token for stateless authentication |
| CORS | Cross-Origin Resource Sharing |
```

### Tables

```markdown
# Use tables for structured comparisons
| Feature | Free | Pro | Enterprise |
|---------|------|-----|-----------|
| API calls/month | 1,000 | 50,000 | Unlimited |
| Support | Community | Email | Dedicated |
| SLA | None | 99.9% | 99.99% |

# Use tables for parameter documentation
| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `page` | integer | No | `1` | Page number |
| `limit` | integer | No | `20` | Results per page |
```

### Admonitions / Callouts

```markdown
> **Note**: Additional context that is helpful but not critical.

> **Important**: Something the reader should pay attention to.

> **Warning**: Potential data loss, security risk, or breaking change.

> **Tip**: Helpful shortcut or best practice.

# GitHub-flavored (newer syntax)
> [!NOTE]
> Additional context that is helpful but not critical.

> [!WARNING]
> This action cannot be undone.

> [!TIP]
> Use `--dry-run` to preview changes before applying.
```

### Links

```markdown
# Descriptive link text (preferred)
See the [authentication guide](./docs/auth.md) for details.

# Avoid "click here" or "this page"
# Bad: For more information, click [here](link).
# Good: For more information, see the [API reference](link).

# Relative links for same-repository files
[Contributing guide](./CONTRIBUTING.md)
[API docs](../docs/api.md)

# Full URLs for external resources
[OpenAPI Specification](https://spec.openapis.org/oas/v3.1.0)
```

## Writing Specific Document Types

### API Endpoint Documentation

```markdown
### Create user

    POST /api/users

Creates a new user account. Returns the created user with a generated ID.

**Authentication:** Bearer token with `users:write` scope.

**Request body:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `email` | string | Yes | Valid email (must be unique) |
| `name` | string | Yes | Display name (1-255 chars) |
| `role` | string | No | `admin`, `user`, or `moderator` (default: `user`) |

**Example request:**

```bash
curl -X POST https://api.example.com/v1/users \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "jane@example.com",
    "name": "Jane Doe",
    "role": "admin"
  }'
```

**Response (201 Created):**

```json
{
  "data": {
    "id": "usr_abc123",
    "email": "jane@example.com",
    "name": "Jane Doe",
    "role": "admin",
    "createdAt": "2024-03-15T10:00:00Z"
  }
}
```

**Errors:**

| Status | Code | When |
|--------|------|------|
| 400 | `VALIDATION_ERROR` | Invalid email format or name too long |
| 409 | `DUPLICATE` | Email already registered |
| 401 | `UNAUTHORIZED` | Missing or expired token |
| 403 | `FORBIDDEN` | Token lacks `users:write` scope |
```

### Error Message Writing

```markdown
# Bad error messages
- Error occurred
- Invalid input
- Something went wrong
- Operation failed

# Good error messages
- Email "invalid@" is not a valid email address
- Name must be between 1 and 255 characters (got 0)
- API key "sk_test_..." has expired. Generate a new key at Settings → API Keys
- Cannot delete user "usr_abc123" — they own 3 active orders. Transfer or cancel orders first.
```

Error message formula:
1. **What went wrong** (specific, not generic)
2. **Why** (what caused it)
3. **How to fix it** (actionable next step)

### Tutorial Writing

Structure tutorials with clear progression:

```markdown
# Tutorial: Build a REST API with Express

## What you'll build

A REST API for a todo list application with CRUD operations,
authentication, and input validation.

## Prerequisites

- Node.js 18 or later
- A text editor
- Basic JavaScript knowledge

## Step 1: Set up the project

Create a new directory and initialize the project:

    mkdir todo-api && cd todo-api
    npm init -y
    npm install express zod

## Step 2: Create the server

Create `index.js`:

```javascript
const express = require('express');
const app = express();
app.use(express.json());

app.listen(3000, () => {
  console.log('Server running on http://localhost:3000');
});
```

Start the server:

    node index.js

Open http://localhost:3000 — you should see "Cannot GET /".
That's expected. Let's add some routes.

## Step 3: Add routes

[Continue with clear, testable steps...]

## What you learned

- How to create an Express server
- How to define REST endpoints
- How to validate input with Zod
- How to handle errors consistently

## Next steps

- Add a database with Prisma
- Deploy to production
- Add authentication
```

Tutorial rules:
- Every step must be independently testable
- Show the expected output after each step
- Include the complete file contents at each step (not just diffs)
- End with what was learned and what to do next

### Inline Code Comments

```typescript
// Good: Explains WHY, not WHAT
// Rate limit to 100 req/min per IP to prevent abuse while
// allowing normal usage patterns (avg user makes ~20 req/min)
const limiter = rateLimit({ windowMs: 60000, max: 100 });

// Bad: Describes what the code obviously does
// Create a rate limiter
const limiter = rateLimit({ windowMs: 60000, max: 100 });

// Good: Documents non-obvious behavior
// PostgreSQL JSONB indexes don't cover nested paths by default.
// This GIN index enables fast lookups on metadata.tags[].
await db.execute('CREATE INDEX idx_metadata_tags ON items USING GIN ((metadata->\'tags\'))');

// Bad: Restates the code
// Create an index on metadata tags
await db.execute('CREATE INDEX idx_metadata_tags ON items USING GIN ((metadata->\'tags\'))');
```

## Numbers and Units

```markdown
# Spell out numbers one through nine; use digits for 10+
The function accepts three parameters.
The API supports 50 concurrent connections.

# Always use digits with units
5 MB, 30 seconds, 100 requests/minute

# Use consistent units
# Bad: The timeout is 30000 ms and the max file size is 10 megabytes.
# Good: The timeout is 30 seconds and the max file size is 10 MB.

# Use tables for multiple values
| Limit | Free | Pro |
|-------|------|-----|
| Requests/min | 60 | 300 |
| Upload size | 5 MB | 50 MB |
| Storage | 1 GB | 100 GB |
```

## Common Mistakes

### 1. Writing for Yourself Instead of the Reader

You know the codebase. The reader doesn't. Explain context that seems obvious to you.

### 2. Documenting Implementation, Not Usage

```markdown
# Implementation-focused (for internal docs)
The UserService queries the PostgreSQL users table with a LEFT JOIN
on profiles, applies Zod validation, and returns a mapped DTO.

# Usage-focused (for API docs)
GET /users returns a list of users with optional profile data.
Include `?include=profile` to add profile information to each user.
```

### 3. Missing Prerequisites

Always state what the reader needs before starting:
- Required software versions
- Required accounts or API keys
- Required knowledge or skills
- Required configuration or environment variables

### 4. Outdated Examples

Code examples must compile/run. After any API change:
1. Update all examples in documentation
2. Run code examples to verify they work
3. Update screenshots and output samples

### 5. Wall of Text

Break long content with:
- Headings (every 2-3 paragraphs)
- Code examples (break up explanations)
- Tables (instead of long lists in prose)
- Bullet points (instead of comma-separated lists)
- Whitespace (breathing room between sections)

## Documentation Maintenance

### Review Triggers

Update documentation when:
- API endpoints change (params, responses, errors)
- Configuration options change
- Dependencies are added or removed
- Default values change
- Behavior changes (even without API changes)
- New features are added
- Features are deprecated or removed

### Freshness Indicators

```markdown
<!-- Last reviewed: 2024-03-15 -->
<!-- Applies to: v1.3.0+ -->
```

### Documentation Testing

| What to Test | How |
|-------------|-----|
| Code examples compile | Extract and run in CI |
| Links aren't broken | Link checker in CI |
| API docs match code | Compare with OpenAPI spec |
| Screenshots are current | Visual regression testing |
| Commands produce expected output | Run in CI |
