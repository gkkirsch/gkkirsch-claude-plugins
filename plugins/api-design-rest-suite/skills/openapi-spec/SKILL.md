---
name: openapi-spec
description: >
  Write and maintain OpenAPI 3.1 specifications for REST APIs. Generate specs
  from code, validate schemas, and produce documentation.
  Triggers: "OpenAPI spec", "Swagger", "API documentation", "API schema".
  NOT for: GraphQL schemas, gRPC protobuf definitions.
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# OpenAPI 3.1 Specifications

## Minimal Complete Spec

```yaml
# openapi.yaml
openapi: 3.1.0
info:
  title: My API
  version: 1.0.0
  description: A clean REST API

servers:
  - url: http://localhost:3000
    description: Local development
  - url: https://api.example.com
    description: Production

paths:
  /users:
    get:
      summary: List users
      operationId: listUsers
      tags: [Users]
      parameters:
        - $ref: '#/components/parameters/PageParam'
        - $ref: '#/components/parameters/LimitParam'
        - name: role
          in: query
          schema:
            type: string
            enum: [admin, user, moderator]
      responses:
        '200':
          description: User list
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
        '401':
          $ref: '#/components/responses/Unauthorized'

    post:
      summary: Create user
      operationId: createUser
      tags: [Users]
      security:
        - bearerAuth: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateUser'
      responses:
        '201':
          description: User created
          headers:
            Location:
              schema:
                type: string
              example: /users/123
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/User'
        '400':
          $ref: '#/components/responses/BadRequest'
        '409':
          $ref: '#/components/responses/Conflict'

  /users/{userId}:
    get:
      summary: Get user by ID
      operationId: getUser
      tags: [Users]
      parameters:
        - $ref: '#/components/parameters/UserId'
      responses:
        '200':
          description: User details
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/User'
        '404':
          $ref: '#/components/responses/NotFound'

    patch:
      summary: Update user
      operationId: updateUser
      tags: [Users]
      parameters:
        - $ref: '#/components/parameters/UserId'
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/UpdateUser'
      responses:
        '200':
          description: Updated user
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/User'
        '404':
          $ref: '#/components/responses/NotFound'

    delete:
      summary: Delete user
      operationId: deleteUser
      tags: [Users]
      parameters:
        - $ref: '#/components/parameters/UserId'
      responses:
        '204':
          description: User deleted
        '404':
          $ref: '#/components/responses/NotFound'

components:
  schemas:
    User:
      type: object
      required: [id, email, name, role, createdAt]
      properties:
        id:
          type: string
          format: uuid
        email:
          type: string
          format: email
        name:
          type: string
          minLength: 1
          maxLength: 100
        role:
          type: string
          enum: [admin, user, moderator]
        avatar:
          type: string
          format: uri
          nullable: true
        createdAt:
          type: string
          format: date-time
        updatedAt:
          type: string
          format: date-time

    CreateUser:
      type: object
      required: [email, name]
      properties:
        email:
          type: string
          format: email
        name:
          type: string
          minLength: 1
          maxLength: 100
        role:
          type: string
          enum: [admin, user, moderator]
          default: user
        password:
          type: string
          minLength: 8

    UpdateUser:
      type: object
      properties:
        name:
          type: string
          minLength: 1
          maxLength: 100
        role:
          type: string
          enum: [admin, user, moderator]
        avatar:
          type: string
          format: uri
          nullable: true

    PaginationMeta:
      type: object
      properties:
        total:
          type: integer
        page:
          type: integer
        perPage:
          type: integer
        totalPages:
          type: integer

    Error:
      type: object
      required: [code, message]
      properties:
        code:
          type: string
        message:
          type: string
        details:
          type: array
          items:
            type: object
            properties:
              field:
                type: string
              message:
                type: string

  parameters:
    UserId:
      name: userId
      in: path
      required: true
      schema:
        type: string
        format: uuid
    PageParam:
      name: page
      in: query
      schema:
        type: integer
        minimum: 1
        default: 1
    LimitParam:
      name: limit
      in: query
      schema:
        type: integer
        minimum: 1
        maximum: 100
        default: 20

  responses:
    BadRequest:
      description: Invalid request
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Error'
          example:
            code: VALIDATION_ERROR
            message: Request validation failed
            details:
              - field: email
                message: Must be a valid email address
    Unauthorized:
      description: Authentication required
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Error'
          example:
            code: UNAUTHORIZED
            message: Authentication required
    NotFound:
      description: Resource not found
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Error'
          example:
            code: NOT_FOUND
            message: Resource not found
    Conflict:
      description: Resource conflict
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Error'
          example:
            code: CONFLICT
            message: A user with this email already exists

  securitySchemes:
    bearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT

security:
  - bearerAuth: []

tags:
  - name: Users
    description: User management
```

## Generate Spec from Express Routes

Use `swagger-jsdoc` to generate specs from JSDoc comments:

```bash
npm install swagger-jsdoc swagger-ui-express
npm install -D @types/swagger-jsdoc @types/swagger-ui-express
```

```typescript
// src/swagger.ts
import swaggerJsdoc from 'swagger-jsdoc';
import swaggerUi from 'swagger-ui-express';
import type { Express } from 'express';

const options: swaggerJsdoc.Options = {
  definition: {
    openapi: '3.1.0',
    info: {
      title: 'My API',
      version: '1.0.0',
    },
    servers: [
      { url: 'http://localhost:3000', description: 'Development' },
    ],
    components: {
      securitySchemes: {
        bearerAuth: { type: 'http', scheme: 'bearer', bearerFormat: 'JWT' },
      },
    },
  },
  apis: ['./src/routes/*.ts'], // Path to annotated route files
};

const spec = swaggerJsdoc(options);

export function setupSwagger(app: Express): void {
  app.use('/docs', swaggerUi.serve, swaggerUi.setup(spec));
  app.get('/openapi.json', (req, res) => res.json(spec));
}
```

### Route Annotations

```typescript
// src/routes/users.ts

/**
 * @openapi
 * /users:
 *   get:
 *     summary: List users
 *     tags: [Users]
 *     parameters:
 *       - in: query
 *         name: page
 *         schema:
 *           type: integer
 *           default: 1
 *     responses:
 *       200:
 *         description: User list
 *         content:
 *           application/json:
 *             schema:
 *               type: array
 *               items:
 *                 $ref: '#/components/schemas/User'
 */
router.get('/users', listUsers);

/**
 * @openapi
 * components:
 *   schemas:
 *     User:
 *       type: object
 *       properties:
 *         id:
 *           type: string
 *           format: uuid
 *         name:
 *           type: string
 *         email:
 *           type: string
 *           format: email
 */
```

## Zod-to-OpenAPI (Type-Safe)

Generate OpenAPI from Zod schemas — single source of truth:

```bash
npm install @asteasolutions/zod-to-openapi zod
```

```typescript
import { z } from 'zod';
import {
  OpenAPIRegistry,
  OpenApiGeneratorV31,
  extendZodWithOpenApi,
} from '@asteasolutions/zod-to-openapi';

extendZodWithOpenApi(z);

// Define Zod schemas with OpenAPI metadata
const UserSchema = z.object({
  id: z.string().uuid().openapi({ example: '550e8400-e29b-41d4-a716-446655440000' }),
  email: z.string().email().openapi({ example: 'alice@example.com' }),
  name: z.string().min(1).max(100).openapi({ example: 'Alice Smith' }),
  role: z.enum(['admin', 'user', 'moderator']).default('user'),
  createdAt: z.string().datetime(),
}).openapi('User');

const CreateUserSchema = z.object({
  email: z.string().email(),
  name: z.string().min(1).max(100),
  role: z.enum(['admin', 'user', 'moderator']).optional(),
  password: z.string().min(8),
}).openapi('CreateUser');

// Register routes
const registry = new OpenAPIRegistry();

registry.registerPath({
  method: 'get',
  path: '/users',
  summary: 'List users',
  tags: ['Users'],
  responses: {
    200: {
      description: 'User list',
      content: {
        'application/json': {
          schema: z.array(UserSchema),
        },
      },
    },
  },
});

registry.registerPath({
  method: 'post',
  path: '/users',
  summary: 'Create user',
  tags: ['Users'],
  request: {
    body: {
      content: { 'application/json': { schema: CreateUserSchema } },
    },
  },
  responses: {
    201: {
      description: 'User created',
      content: { 'application/json': { schema: UserSchema } },
    },
  },
});

// Generate full spec
const generator = new OpenApiGeneratorV31(registry.definitions);
const spec = generator.generateDocument({
  openapi: '3.1.0',
  info: { title: 'My API', version: '1.0.0' },
});
```

**Benefits**: Zod validates at runtime AND generates OpenAPI docs. One schema, two purposes.

## Validate Existing Specs

```bash
# Install validator
npm install -D @redocly/cli

# Validate spec
npx @redocly/cli lint openapi.yaml

# Preview docs
npx @redocly/cli preview-docs openapi.yaml

# Bundle multi-file specs into one
npx @redocly/cli bundle openapi.yaml -o dist/openapi.yaml
```

## Serve Interactive Docs

### Swagger UI

```typescript
import swaggerUi from 'swagger-ui-express';
import spec from './openapi.json' assert { type: 'json' };

app.use('/docs', swaggerUi.serve, swaggerUi.setup(spec, {
  customCss: '.swagger-ui .topbar { display: none }',
  customSiteTitle: 'My API Docs',
}));
```

### Scalar (Modern Alternative)

```typescript
import { apiReference } from '@scalar/express-api-reference';

app.use('/docs', apiReference({
  spec: { url: '/openapi.json' },
  theme: 'kepler',
}));
```

### Redoc (Static-Friendly)

```html
<!-- Single HTML file, no build step -->
<script src="https://cdn.redoc.ly/redoc/latest/bundles/redoc.standalone.js"></script>
<redoc spec-url="./openapi.json"></redoc>
```

## Gotchas

- **3.1 vs 3.0**: OpenAPI 3.1 aligns with JSON Schema 2020-12. Key differences: `nullable` → `type: ["string", "null"]`, `exclusiveMinimum` is a number not boolean. Most tools now support 3.1.
- **`operationId` must be unique** across the entire spec. Use `camelCase`: `listUsers`, `getUser`, `createUser`. These become function names in generated SDKs.
- **$ref replaces the entire object.** You can't add siblings to a `$ref`. If you need to override, use `allOf: [{ $ref: '...' }, { description: 'override' }]`.
- **Spec-first vs code-first**: Spec-first (write YAML, generate code) is ideal but rare in practice. Code-first (annotate code, generate spec) is more common. Zod-to-OpenAPI gives you the best of both.
- **Don't over-specify.** Only document what clients actually need. Internal implementation details (database IDs, internal flags) should not appear in the spec.
- **Examples matter more than descriptions.** A good `example` value communicates more than a paragraph of `description` text.
