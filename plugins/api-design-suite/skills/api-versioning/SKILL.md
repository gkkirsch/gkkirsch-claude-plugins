---
name: api-versioning
description: >
  API versioning and evolution — URL versioning, header versioning, deprecation
  strategy, backward compatibility, API documentation with OpenAPI/Swagger.
  Triggers: "API versioning", "API version", "API deprecation", "OpenAPI",
  "Swagger", "API documentation", "backward compatibility", "breaking changes".
  NOT for: REST endpoint implementation (use rest-api-patterns), GraphQL (use graphql-patterns).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# API Versioning & Documentation

## URL Path Versioning (Recommended)

```typescript
import { Router } from 'express';
import v1Router from './routes/v1';
import v2Router from './routes/v2';

// Mount versioned routers
app.use('/api/v1', v1Router);
app.use('/api/v2', v2Router);

// routes/v1/users.ts
const router = Router();
router.get('/users', listUsersV1);     // Original format
router.get('/users/:id', getUserV1);
export default router;

// routes/v2/users.ts
const router = Router();
router.get('/users', listUsersV2);     // New format (includes metadata)
router.get('/users/:id', getUserV2);
export default router;

// Shared logic — avoid duplicating business logic
async function listUsersV1(req: Request, res: Response) {
  const users = await getUsersFromDB(req.query);
  res.json(users.map(formatUserV1)); // V1 response shape
}

async function listUsersV2(req: Request, res: Response) {
  const users = await getUsersFromDB(req.query); // Same DB query
  res.json({
    data: users.map(formatUserV2),  // V2 response shape
    meta: { version: 2, deprecations: [] },
  });
}
```

## Header Versioning (Alternative)

```typescript
// Client sends: Accept: application/vnd.myapi.v2+json
function versionMiddleware(req: Request, res: Response, next: NextFunction) {
  const accept = req.headers.accept || '';
  const match = accept.match(/application\/vnd\.myapi\.v(\d+)\+json/);
  req.apiVersion = match ? parseInt(match[1]) : 1; // Default to v1
  next();
}

app.use(versionMiddleware);

router.get('/users', (req, res) => {
  if (req.apiVersion >= 2) {
    return listUsersV2(req, res);
  }
  return listUsersV1(req, res);
});
```

## Deprecation Strategy

```typescript
// Deprecation header middleware
function deprecationWarning(
  sunset: string,
  alternative: string
) {
  return (req: Request, res: Response, next: NextFunction) => {
    res.setHeader('Deprecation', 'true');
    res.setHeader('Sunset', sunset); // RFC 8594
    res.setHeader('Link', `<${alternative}>; rel="successor-version"`);
    next();
  };
}

// Apply to deprecated endpoints
router.get('/users',
  deprecationWarning('2026-06-01', '/api/v2/users'),
  listUsersV1
);

// Deprecation response wrapper
function withDeprecationNotice(data: any, notice: string) {
  return {
    ...data,
    _deprecation: {
      message: notice,
      sunset: '2026-06-01',
      migration: 'https://docs.example.com/migration-guide',
    },
  };
}
```

## Breaking vs Non-Breaking Changes

```
NON-BREAKING (safe to deploy without versioning):
  ✓ Adding new endpoints
  ✓ Adding optional fields to responses
  ✓ Adding optional query parameters
  ✓ Adding new enum values (if client handles unknown values)
  ✓ Increasing rate limits
  ✓ Adding new response headers

BREAKING (requires new version):
  ✗ Removing or renaming fields
  ✗ Changing field types (string → number)
  ✗ Changing URL structure
  ✗ Removing endpoints
  ✗ Making optional fields required
  ✗ Changing error response format
  ✗ Changing authentication method
  ✗ Changing pagination format
  ✗ Reducing rate limits
```

## OpenAPI / Swagger Documentation

```yaml
# openapi.yaml
openapi: 3.1.0
info:
  title: My API
  version: 2.0.0
  description: |
    API for managing users and posts.

    ## Authentication
    All endpoints require a Bearer token in the Authorization header.

    ## Rate Limiting
    - 100 requests per 15 minutes for authenticated users
    - 20 requests per 15 minutes for unauthenticated requests

servers:
  - url: https://api.example.com/v2
    description: Production
  - url: http://localhost:3000/api/v2
    description: Development

security:
  - bearerAuth: []

paths:
  /users:
    get:
      summary: List users
      operationId: listUsers
      tags: [Users]
      parameters:
        - name: page
          in: query
          schema:
            type: integer
            default: 1
            minimum: 1
        - name: limit
          in: query
          schema:
            type: integer
            default: 20
            minimum: 1
            maximum: 100
        - name: search
          in: query
          schema:
            type: string
            maxLength: 100
        - name: role
          in: query
          schema:
            $ref: '#/components/schemas/Role'
      responses:
        '200':
          description: List of users
          content:
            application/json:
              schema:
                type: object
                properties:
                  data:
                    type: array
                    items:
                      $ref: '#/components/schemas/User'
                  pagination:
                    $ref: '#/components/schemas/Pagination'
        '401':
          $ref: '#/components/responses/Unauthorized'
        '429':
          $ref: '#/components/responses/RateLimited'

    post:
      summary: Create a user
      operationId: createUser
      tags: [Users]
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateUserInput'
      responses:
        '201':
          description: User created
          headers:
            Location:
              schema:
                type: string
              description: URL of the created user
          content:
            application/json:
              schema:
                type: object
                properties:
                  data:
                    $ref: '#/components/schemas/User'
        '400':
          $ref: '#/components/responses/ValidationError'
        '409':
          description: Email already registered

  /users/{id}:
    get:
      summary: Get a user by ID
      operationId: getUser
      tags: [Users]
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
            format: uuid
      responses:
        '200':
          description: User details
          content:
            application/json:
              schema:
                type: object
                properties:
                  data:
                    $ref: '#/components/schemas/User'
        '404':
          $ref: '#/components/responses/NotFound'

components:
  securitySchemes:
    bearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT

  schemas:
    User:
      type: object
      properties:
        id:
          type: string
          format: uuid
        name:
          type: string
        email:
          type: string
          format: email
        role:
          $ref: '#/components/schemas/Role'
        createdAt:
          type: string
          format: date-time
      required: [id, name, email, role, createdAt]

    CreateUserInput:
      type: object
      properties:
        name:
          type: string
          minLength: 1
          maxLength: 100
        email:
          type: string
          format: email
        role:
          $ref: '#/components/schemas/Role'
      required: [name, email]

    Role:
      type: string
      enum: [user, editor, admin]

    Pagination:
      type: object
      properties:
        page:
          type: integer
        limit:
          type: integer
        total:
          type: integer
        totalPages:
          type: integer
        hasMore:
          type: boolean

    Error:
      type: object
      properties:
        error:
          type: object
          properties:
            code:
              type: string
            message:
              type: string
            details:
              type: object

  responses:
    Unauthorized:
      description: Authentication required
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Error'
    NotFound:
      description: Resource not found
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Error'
    ValidationError:
      description: Validation failed
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Error'
    RateLimited:
      description: Too many requests
      headers:
        Retry-After:
          schema:
            type: integer
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Error'
```

## Serving OpenAPI Docs

```typescript
import swaggerUi from 'swagger-ui-express';
import YAML from 'yaml';
import { readFileSync } from 'fs';

const spec = YAML.parse(readFileSync('./openapi.yaml', 'utf-8'));

app.use('/docs', swaggerUi.serve, swaggerUi.setup(spec, {
  customCss: '.swagger-ui .topbar { display: none }',
  customSiteTitle: 'My API Docs',
}));

// Also serve raw spec
app.get('/openapi.json', (req, res) => res.json(spec));
app.get('/openapi.yaml', (req, res) => {
  res.type('text/yaml').send(readFileSync('./openapi.yaml', 'utf-8'));
});
```

## API Evolution Best Practices

```typescript
// 1. Additive changes (no version bump needed)
// Add new optional field — existing clients ignore it
type UserV1 = { id: string; name: string; email: string };
type UserV1Plus = { id: string; name: string; email: string; avatar?: string }; // Safe

// 2. Field aliasing during migration
function formatUser(user: DBUser, version: number) {
  const base = { id: user.id, name: user.name, email: user.email };

  if (version >= 2) {
    return { ...base, displayName: user.name }; // New field name in v2
  }

  return { ...base, username: user.name }; // Old field name in v1
}

// 3. Feature flags for gradual rollout
app.get('/api/users', (req, res) => {
  const useNewFormat = req.headers['x-use-new-format'] === 'true';
  // Return new format for opted-in clients
});

// 4. Changelog in response headers
app.use((req, res, next) => {
  res.setHeader('X-API-Version', '2.3.1');
  res.setHeader('X-API-Changelog', 'https://docs.example.com/changelog');
  next();
});
```

## Gotchas

1. **Don't version too aggressively.** Every version is a maintenance burden. Use additive changes (new optional fields, new endpoints) as long as possible. Only create a new version for genuine breaking changes.

2. **Sunset headers need enforcement.** Setting `Sunset: 2026-06-01` is a promise. Build monitoring to track v1 usage and alert clients before actually shutting it down. Don't just flip the switch on the sunset date.

3. **OpenAPI and implementation can drift.** Auto-generate the spec from code (like `tsoa`, `nestjs/swagger`) or auto-generate types from the spec (`openapi-typescript`). Manual specs always get out of date.

4. **GraphQL doesn't need versioning.** GraphQL's type system handles evolution via `@deprecated` directives and additive type changes. If you're versioning a GraphQL API, you're probably doing it wrong.

5. **API keys in URLs are a security risk.** `?api_key=abc123` gets logged in server logs, browser history, and CDN caches. Use the `Authorization` header or a custom header like `X-API-Key`.

6. **CORS preflight caches expire.** `Access-Control-Max-Age` defaults to 5 seconds in some browsers. Set it explicitly (e.g., 86400 for 24 hours) to avoid excessive OPTIONS requests from browser clients.
