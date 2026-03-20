---
name: api-documentation
description: >
  API documentation patterns with OpenAPI, code examples, and SDK generation.
  Use when writing API docs, generating OpenAPI specs, creating code examples,
  or building documentation sites for REST or GraphQL APIs.
  Triggers: "API docs", "OpenAPI", "Swagger", "API reference", "endpoint docs",
  "code example", "SDK docs", "API guide".
  NOT for: code comments, README files, or user-facing product documentation.
version: 1.0.0
allowed-tools: Read, Grep, Glob, Bash
---

# API Documentation Patterns

## OpenAPI 3.1 Specification

```yaml
# openapi.yaml
openapi: 3.1.0
info:
  title: My API
  version: 1.0.0
  description: |
    A RESTful API for managing resources.

    ## Authentication
    All endpoints require a Bearer token in the Authorization header.

    ## Rate Limits
    - 100 requests/minute per API key
    - 429 response with Retry-After header when exceeded

    ## Errors
    All errors follow RFC 7807 Problem Details format.
  contact:
    email: api-support@example.com

servers:
  - url: https://api.example.com/v1
    description: Production
  - url: https://staging-api.example.com/v1
    description: Staging

security:
  - bearerAuth: []

paths:
  /users:
    get:
      operationId: listUsers
      summary: List users
      description: Returns a paginated list of users. Results are ordered by creation date (newest first).
      tags: [Users]
      parameters:
        - name: limit
          in: query
          schema:
            type: integer
            minimum: 1
            maximum: 100
            default: 20
          description: Number of results per page
        - name: cursor
          in: query
          schema:
            type: string
          description: Pagination cursor from previous response
        - name: status
          in: query
          schema:
            type: string
            enum: [active, inactive, suspended]
          description: Filter by user status
      responses:
        '200':
          description: Paginated list of users
          content:
            application/json:
              schema:
                type: object
                properties:
                  data:
                    type: array
                    items:
                      $ref: '#/components/schemas/User'
                  meta:
                    $ref: '#/components/schemas/PaginationMeta'
              example:
                data:
                  - id: "usr_abc123"
                    email: "user@example.com"
                    name: "Jane Doe"
                    status: "active"
                    createdAt: "2026-01-15T10:30:00Z"
                meta:
                  hasMore: true
                  nextCursor: "eyJpZCI6MTAwfQ"
                  total: 1542
        '401':
          $ref: '#/components/responses/Unauthorized'
        '429':
          $ref: '#/components/responses/RateLimited'

    post:
      operationId: createUser
      summary: Create a user
      tags: [Users]
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required: [email, name]
              properties:
                email:
                  type: string
                  format: email
                  description: Must be unique across all users
                name:
                  type: string
                  minLength: 1
                  maxLength: 100
                role:
                  type: string
                  enum: [user, admin]
                  default: user
            example:
              email: "new@example.com"
              name: "New User"
              role: "user"
      responses:
        '201':
          description: User created successfully
          content:
            application/json:
              schema:
                type: object
                properties:
                  data:
                    $ref: '#/components/schemas/User'
        '409':
          description: Email already exists
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
              example:
                type: "https://api.example.com/errors/duplicate-email"
                title: "Conflict"
                status: 409
                detail: "A user with this email already exists"

components:
  schemas:
    User:
      type: object
      properties:
        id:
          type: string
          description: Unique identifier (prefixed with usr_)
          example: "usr_abc123"
        email:
          type: string
          format: email
        name:
          type: string
        status:
          type: string
          enum: [active, inactive, suspended]
        createdAt:
          type: string
          format: date-time

    PaginationMeta:
      type: object
      properties:
        hasMore:
          type: boolean
        nextCursor:
          type: string
          nullable: true
        total:
          type: integer

    Error:
      type: object
      properties:
        type:
          type: string
          format: uri
        title:
          type: string
        status:
          type: integer
        detail:
          type: string

  responses:
    Unauthorized:
      description: Missing or invalid authentication
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Error'
          example:
            type: "https://api.example.com/errors/unauthorized"
            title: "Unauthorized"
            status: 401
            detail: "Bearer token is missing or invalid"

    RateLimited:
      description: Rate limit exceeded
      headers:
        Retry-After:
          schema:
            type: integer
          description: Seconds to wait before retrying
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Error'

  securitySchemes:
    bearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT
```

## Multi-Language Code Examples

```typescript
// Code example generator for API docs
interface CodeExample {
  language: string;
  label: string;
  code: string;
}

function generateExamples(
  method: string,
  path: string,
  body?: Record<string, any>,
  headers?: Record<string, string>
): CodeExample[] {
  const url = `https://api.example.com/v1${path}`;
  const authHeader = 'Authorization: Bearer YOUR_API_KEY';

  return [
    {
      language: 'curl',
      label: 'cURL',
      code: [
        `curl -X ${method} '${url}'`,
        `  -H '${authHeader}'`,
        body ? `  -H 'Content-Type: application/json'` : '',
        body ? `  -d '${JSON.stringify(body, null, 2)}'` : '',
      ].filter(Boolean).join(' \\\n'),
    },
    {
      language: 'javascript',
      label: 'Node.js',
      code: `const response = await fetch('${url}', {
  method: '${method}',
  headers: {
    'Authorization': 'Bearer YOUR_API_KEY',
    ${body ? "'Content-Type': 'application/json'," : ''}
  },${body ? `\n  body: JSON.stringify(${JSON.stringify(body, null, 4).replace(/\n/g, '\n  ')}),` : ''}
});

const data = await response.json();
console.log(data);`,
    },
    {
      language: 'python',
      label: 'Python',
      code: `import requests

response = requests.${method.toLowerCase()}(
    '${url}',
    headers={'Authorization': 'Bearer YOUR_API_KEY'},${body ? `\n    json=${JSON.stringify(body)},` : ''}
)

data = response.json()
print(data)`,
    },
  ];
}
```

## Changelog Format

```markdown
# Changelog

## v1.3.0 — 2026-03-15

### Added
- `GET /users/:id/activity` — retrieve user activity log
- `status` filter parameter on `GET /users`
- Rate limit headers on all responses (`X-RateLimit-Remaining`, `X-RateLimit-Reset`)

### Changed
- `GET /users` now returns results sorted by `createdAt` descending (was ascending)
- Increased default pagination `limit` from 10 to 20

### Deprecated
- `GET /users?page=N` pagination — use cursor-based pagination instead. Will be removed in v2.0.

### Fixed
- `PATCH /users/:id` now returns 404 instead of 500 for non-existent users
```

## Gotchas

1. **OpenAPI examples are not validated** — the example values in your spec can be completely wrong (wrong types, missing required fields) and most tools won't warn you. Use `openapi-spec-validator` or `spectral` to catch issues.

2. **operationId must be unique across the entire spec** — duplicate operationIds break SDK generation. Use consistent naming: `listUsers`, `getUser`, `createUser`, `updateUser`, `deleteUser`.

3. **$ref cannot have sibling properties** — `$ref: '#/...'` alongside `description:` is invalid in OpenAPI 3.0 (fixed in 3.1). In 3.0, wrap in allOf: `allOf: [{ $ref: '#/...' }]` then add properties.

4. **Code examples with hardcoded tokens** — never put real API keys in documentation examples. Use `YOUR_API_KEY` placeholder and add a callout explaining where to find the key.

5. **Changelog without migration guide** — listing breaking changes without explaining how to migrate is useless. Include before/after code for every breaking change.

6. **Missing error response documentation** — documenting only happy paths leaves developers guessing about error formats. Document every possible error status code with example response bodies and troubleshooting steps.
