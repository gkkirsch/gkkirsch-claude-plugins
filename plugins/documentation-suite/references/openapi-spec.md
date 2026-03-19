# OpenAPI 3.1 Complete Reference

Comprehensive reference for the OpenAPI 3.1 specification. Use this when generating, validating, or understanding OpenAPI documents.

## OpenAPI 3.1 vs 3.0

OpenAPI 3.1 is fully compatible with JSON Schema 2020-12. Key differences from 3.0:

| Feature | 3.0 | 3.1 |
|---------|-----|-----|
| JSON Schema version | Draft 04 (subset) | 2020-12 (full) |
| `nullable` keyword | `nullable: true` | `type: ['string', 'null']` |
| `exclusiveMinimum` | Boolean | Number |
| `exclusiveMaximum` | Boolean | Number |
| `example` keyword | OpenAPI-specific | JSON Schema `examples` array |
| `if/then/else` | Not supported | Supported |
| `prefixItems` | Not supported | Supported (tuple validation) |
| `$dynamicRef` | Not supported | Supported |
| Webhooks | Via callbacks only | First-class `webhooks` object |
| `type` as array | Not supported | Supported |
| `const` keyword | Not supported | Supported |
| `contentEncoding` | Not supported | Supported |
| `contentMediaType` | Not supported | Supported |

## Document Structure

```yaml
openapi: 3.1.0                    # Required. OpenAPI version

info:                              # Required. API metadata
  title: string                    # Required
  version: string                  # Required. API version (not OpenAPI version)
  summary: string                  # Short summary
  description: string              # Markdown description
  termsOfService: string           # URL to terms
  contact:                         # Contact information
    name: string
    url: string
    email: string
  license:                         # License information
    name: string                   # Required
    identifier: string             # SPDX expression (3.1 only)
    url: string

jsonSchemaDialect: string          # Default JSON Schema dialect URI

servers:                           # Array of server objects
  - url: string                    # Required. Server URL (can use variables)
    description: string
    variables:                     # URL template variables
      variableName:
        default: string            # Required
        enum: [string]
        description: string

paths:                             # API endpoints
  /path:                           # Path template
    $ref: string                   # External path item reference
    summary: string
    description: string
    get: Operation
    put: Operation
    post: Operation
    delete: Operation
    options: Operation
    head: Operation
    patch: Operation
    trace: Operation
    parameters: [Parameter]        # Common parameters for all operations
    servers: [Server]              # Path-specific servers

webhooks:                          # 3.1: First-class webhook definitions
  eventName:
    post: Operation                # Webhook callback operation

components:                        # Reusable components
  schemas: {}                      # Schema objects
  responses: {}                    # Response objects
  parameters: {}                   # Parameter objects
  examples: {}                     # Example objects
  requestBodies: {}                # Request body objects
  headers: {}                      # Header objects
  securitySchemes: {}              # Security scheme objects
  links: {}                        # Link objects
  callbacks: {}                    # Callback objects
  pathItems: {}                    # 3.1: Reusable path items

security: []                       # Global security requirements
tags: []                           # Tag definitions for grouping
externalDocs:                      # External documentation link
  url: string                      # Required
  description: string
```

## Operation Object

```yaml
/users:
  get:
    operationId: listUsers         # Unique operation identifier
    summary: string                # Short summary (shown in UI)
    description: string            # Detailed markdown description
    tags: [string]                 # Grouping tags
    deprecated: boolean            # Mark as deprecated
    externalDocs:                  # External documentation
      url: string
      description: string

    parameters:                    # Request parameters
      - $ref: '#/components/parameters/PageParam'
      - name: string              # Required
        in: query|header|path|cookie  # Required
        required: boolean          # Required for path params
        deprecated: boolean
        allowEmptyValue: boolean   # Query params only
        description: string
        schema: Schema
        style: string              # Serialization style
        explode: boolean           # Explode arrays/objects
        allowReserved: boolean     # Allow reserved chars in query
        example: any
        examples: {}
        content: {}                # Alternative to schema (for complex types)

    requestBody:                   # Request body
      $ref: string                 # Or inline:
      required: boolean
      description: string
      content:
        application/json:
          schema: Schema
          example: any
          examples: {}
          encoding: {}             # Encoding for multipart
        multipart/form-data:
          schema:
            type: object
            properties:
              file:
                type: string
                format: binary
              metadata:
                type: object
          encoding:
            file:
              contentType: image/png, image/jpeg
            metadata:
              contentType: application/json

    responses:                     # Required. Response definitions
      '200':
        description: string        # Required
        headers: {}
        content:
          application/json:
            schema: Schema
            example: any
            examples: {}
        links: {}                  # Hypermedia links
      '4XX':                       # Range wildcards
        $ref: '#/components/responses/ClientError'
      default:                     # Default response
        $ref: '#/components/responses/UnexpectedError'

    callbacks:                     # Async callbacks
      onEvent:
        '{$request.body#/callbackUrl}':
          post:
            requestBody:
              content:
                application/json:
                  schema:
                    $ref: '#/components/schemas/Event'
            responses:
              '200':
                description: Callback processed

    security:                      # Operation-level security
      - bearerAuth: []
      - apiKey: []
      - oauth2: [users:read]

    servers:                       # Operation-level servers
      - url: https://special.api.example.com
```

## Schema Object (JSON Schema 2020-12)

OpenAPI 3.1 supports the full JSON Schema 2020-12 vocabulary:

### Basic Types

```yaml
# String
type: string
minLength: 1
maxLength: 255
pattern: "^[a-zA-Z0-9]+$"
format: date-time|date|time|duration|email|idn-email|hostname|
        idn-hostname|ipv4|ipv6|uri|uri-reference|iri|iri-reference|
        uuid|uri-template|json-pointer|relative-json-pointer|regex|
        password|byte|binary|int32|int64|float|double

# Number
type: number
minimum: 0
maximum: 100
exclusiveMinimum: 0           # 3.1: number, not boolean
exclusiveMaximum: 100         # 3.1: number, not boolean
multipleOf: 0.01

# Integer
type: integer
format: int32|int64
minimum: 0
maximum: 2147483647

# Boolean
type: boolean

# Null (3.1)
type: 'null'

# Nullable (3.1 style)
type: [string, 'null']        # Instead of 3.0's nullable: true

# Const (3.1)
const: "fixed_value"

# Enum
type: string
enum: [active, inactive, suspended]
```

### Object Types

```yaml
type: object
required: [id, name, email]
properties:
  id:
    type: string
    format: uuid
    readOnly: true              # Only in responses
  name:
    type: string
    minLength: 1
    maxLength: 255
  email:
    type: string
    format: email
  metadata:
    type: object
    additionalProperties: true  # Allow any extra fields
  tags:
    type: object
    additionalProperties:
      type: string              # Values must be strings
  settings:
    type: object
    propertyNames:
      pattern: "^[a-z_]+$"     # Key format validation
minProperties: 1
maxProperties: 50
additionalProperties: false     # Reject unknown fields
```

### Array Types

```yaml
# Basic array
type: array
items:
  $ref: '#/components/schemas/User'
minItems: 0
maxItems: 100
uniqueItems: true

# Tuple validation (3.1)
type: array
prefixItems:                    # Fixed items at specific positions
  - type: number                # First item: latitude
  - type: number                # Second item: longitude
items: false                    # No additional items allowed

# Mixed-type array
type: array
items:
  oneOf:
    - $ref: '#/components/schemas/TextBlock'
    - $ref: '#/components/schemas/ImageBlock'

# Array with contains (3.1)
type: array
contains:
  type: object
  properties:
    type:
      const: "primary"
minContains: 1
maxContains: 1
```

### Composition

```yaml
# allOf — Combine schemas (AND)
allOf:
  - $ref: '#/components/schemas/BaseEntity'
  - type: object
    properties:
      specificField:
        type: string

# oneOf — Exactly one must match (XOR)
oneOf:
  - $ref: '#/components/schemas/CreditCard'
  - $ref: '#/components/schemas/BankTransfer'
  - $ref: '#/components/schemas/PayPal'
discriminator:
  propertyName: paymentType
  mapping:
    credit_card: '#/components/schemas/CreditCard'
    bank_transfer: '#/components/schemas/BankTransfer'
    paypal: '#/components/schemas/PayPal'

# anyOf — One or more must match (OR)
anyOf:
  - type: string
  - type: number

# not — Must not match
not:
  type: string
  pattern: "^admin_"

# if/then/else (3.1)
type: object
properties:
  type:
    type: string
    enum: [personal, business]
  taxId:
    type: string
if:
  properties:
    type:
      const: business
then:
  required: [taxId]
else:
  properties:
    taxId: false                # Disallow taxId for personal
```

### References

```yaml
# Local reference
$ref: '#/components/schemas/User'

# File reference
$ref: './schemas/user.yaml'

# URL reference
$ref: 'https://api.example.com/schemas/user.yaml'

# Reference with description override (3.1)
$ref: '#/components/schemas/User'
description: "The user who created this order"
summary: "Creator"

# Dynamic reference (3.1)
$dynamicRef: '#node'
```

## Security Schemes

```yaml
components:
  securitySchemes:
    # Bearer Token (JWT)
    bearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT
      description: JWT obtained from POST /auth/login

    # Basic Auth
    basicAuth:
      type: http
      scheme: basic

    # API Key in Header
    apiKeyHeader:
      type: apiKey
      in: header
      name: X-API-Key

    # API Key in Query
    apiKeyQuery:
      type: apiKey
      in: query
      name: api_key

    # API Key in Cookie
    apiKeyCookie:
      type: apiKey
      in: cookie
      name: session_id

    # OAuth 2.0
    oauth2:
      type: oauth2
      flows:
        authorizationCode:
          authorizationUrl: https://auth.example.com/authorize
          tokenUrl: https://auth.example.com/token
          refreshUrl: https://auth.example.com/refresh
          scopes:
            users:read: Read user data
            users:write: Modify user data
            admin: Full admin access
        clientCredentials:
          tokenUrl: https://auth.example.com/token
          scopes:
            service:read: Service-to-service read
            service:write: Service-to-service write
        implicit:
          authorizationUrl: https://auth.example.com/authorize
          scopes:
            users:read: Read user data
        password:
          tokenUrl: https://auth.example.com/token
          scopes:
            users:read: Read user data

    # OpenID Connect
    openIdConnect:
      type: openIdConnect
      openIdConnectUrl: https://auth.example.com/.well-known/openid-configuration

    # Mutual TLS
    mutualTLS:
      type: mutualTLS
      description: Client certificate authentication

# Applying security
security:
  - bearerAuth: []              # Global default
  - {}                          # Empty = no auth (optional)

paths:
  /public:
    get:
      security: []              # Override: no auth required
  /admin:
    get:
      security:
        - bearerAuth: []
        - oauth2: [admin]       # Require admin scope
```

## Parameter Styles

Control how parameters are serialized:

### Query Parameters

```yaml
# style: form (default for query)
# GET /users?role=admin
- name: role
  in: query
  schema:
    type: string

# explode: true (default for form)
# GET /users?role=admin&role=user
- name: role
  in: query
  explode: true
  schema:
    type: array
    items:
      type: string

# explode: false
# GET /users?role=admin,user
- name: role
  in: query
  explode: false
  schema:
    type: array
    items:
      type: string

# style: deepObject (nested objects)
# GET /users?filter[role]=admin&filter[status]=active
- name: filter
  in: query
  style: deepObject
  explode: true
  schema:
    type: object
    properties:
      role:
        type: string
      status:
        type: string

# style: pipeDelimited
# GET /events?ids=1|2|3
- name: ids
  in: query
  style: pipeDelimited
  schema:
    type: array
    items:
      type: integer

# style: spaceDelimited
# GET /events?ids=1%202%203
- name: ids
  in: query
  style: spaceDelimited
  schema:
    type: array
    items:
      type: integer
```

### Path Parameters

```yaml
# style: simple (default for path)
# GET /users/123
- name: id
  in: path
  required: true
  schema:
    type: string

# Array in path
# GET /users/123,456,789
- name: ids
  in: path
  required: true
  style: simple
  schema:
    type: array
    items:
      type: string

# style: label
# GET /users/.123
- name: id
  in: path
  required: true
  style: label
  schema:
    type: string

# style: matrix
# GET /users/;id=123
- name: id
  in: path
  required: true
  style: matrix
  schema:
    type: string
```

### Header Parameters

```yaml
# style: simple (default and only option for headers)
- name: X-Request-ID
  in: header
  required: true
  schema:
    type: string
    format: uuid
  description: Unique request identifier for tracing
```

## Content Negotiation

```yaml
requestBody:
  content:
    application/json:
      schema:
        $ref: '#/components/schemas/User'
    application/xml:
      schema:
        $ref: '#/components/schemas/User'
    application/x-www-form-urlencoded:
      schema:
        type: object
        properties:
          name:
            type: string
          email:
            type: string
        required: [name, email]
    multipart/form-data:
      schema:
        type: object
        properties:
          profile:
            $ref: '#/components/schemas/UserProfile'
          avatar:
            type: string
            format: binary
            contentMediaType: image/png    # 3.1
      encoding:
        profile:
          contentType: application/json
        avatar:
          contentType: image/png, image/jpeg, image/webp
          headers:
            X-Custom-Header:
              schema:
                type: string
    application/octet-stream:
      schema:
        type: string
        format: binary
```

## Webhooks (3.1)

```yaml
webhooks:
  userCreated:
    post:
      summary: User created event
      operationId: onUserCreated
      tags:
        - Webhooks
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/WebhookEvent'
            example:
              id: "evt_abc123"
              type: "user.created"
              data:
                id: "usr_def456"
                email: "jane@example.com"
      responses:
        '200':
          description: Event processed
        '400':
          description: Invalid payload
      security:
        - webhookSignature: []

  orderPaid:
    post:
      summary: Order payment confirmed
      operationId: onOrderPaid
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/WebhookEvent'
      responses:
        '200':
          description: Event processed
```

## Links (Hypermedia)

```yaml
paths:
  /users:
    post:
      operationId: createUser
      responses:
        '201':
          description: User created
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/User'
          links:
            GetUser:
              operationId: getUser
              parameters:
                userId: '$response.body#/id'
              description: Get the created user
            ListUserOrders:
              operationId: listUserOrders
              parameters:
                userId: '$response.body#/id'
              description: List orders for this user

  /users/{userId}:
    get:
      operationId: getUser
      parameters:
        - name: userId
          in: path
          required: true
          schema:
            type: string
      responses:
        '200':
          description: User details
          links:
            UpdateUser:
              operationId: updateUser
              parameters:
                userId: '$response.body#/id'
            DeleteUser:
              operationId: deleteUser
              parameters:
                userId: '$response.body#/id'
```

## Common Patterns

### Pagination

```yaml
components:
  parameters:
    PageParam:
      name: page
      in: query
      schema:
        type: integer
        minimum: 1
        default: 1
      description: Page number (1-indexed)
    LimitParam:
      name: limit
      in: query
      schema:
        type: integer
        minimum: 1
        maximum: 100
        default: 20
      description: Results per page
    CursorParam:
      name: cursor
      in: query
      schema:
        type: string
      description: Pagination cursor from previous response

  schemas:
    PaginationMeta:
      type: object
      properties:
        page:
          type: integer
        limit:
          type: integer
        total:
          type: integer
        pages:
          type: integer
        hasNext:
          type: boolean
        hasPrev:
          type: boolean

    CursorPaginationMeta:
      type: object
      properties:
        nextCursor:
          type: [string, 'null']
        hasMore:
          type: boolean

  headers:
    X-Total-Count:
      schema:
        type: integer
      description: Total items matching query
    X-Page-Count:
      schema:
        type: integer
      description: Total pages
    Link:
      schema:
        type: string
      description: "RFC 8288 pagination links"
```

### Filtering and Sorting

```yaml
parameters:
  - name: sort
    in: query
    schema:
      type: string
      enum: [created_at, updated_at, name, email]
      default: created_at
    description: Field to sort by
  - name: order
    in: query
    schema:
      type: string
      enum: [asc, desc]
      default: desc
    description: Sort direction
  - name: search
    in: query
    schema:
      type: string
      minLength: 2
    description: Full-text search query
  - name: created_after
    in: query
    schema:
      type: string
      format: date-time
    description: Filter by creation date (after)
  - name: created_before
    in: query
    schema:
      type: string
      format: date-time
    description: Filter by creation date (before)
```

### File Upload

```yaml
/upload:
  post:
    summary: Upload a file
    requestBody:
      required: true
      content:
        multipart/form-data:
          schema:
            type: object
            required: [file]
            properties:
              file:
                type: string
                format: binary
                description: File to upload (max 10MB)
              description:
                type: string
                maxLength: 500
          encoding:
            file:
              contentType: >-
                image/png, image/jpeg, image/webp,
                application/pdf, text/csv
    responses:
      '201':
        description: File uploaded
        content:
          application/json:
            schema:
              type: object
              properties:
                id:
                  type: string
                  format: uuid
                url:
                  type: string
                  format: uri
                size:
                  type: integer
                  description: File size in bytes
                contentType:
                  type: string
      '413':
        description: File too large
```

### Polymorphism with Discriminator

```yaml
components:
  schemas:
    Notification:
      oneOf:
        - $ref: '#/components/schemas/EmailNotification'
        - $ref: '#/components/schemas/SmsNotification'
        - $ref: '#/components/schemas/PushNotification'
      discriminator:
        propertyName: channel
        mapping:
          email: '#/components/schemas/EmailNotification'
          sms: '#/components/schemas/SmsNotification'
          push: '#/components/schemas/PushNotification'

    NotificationBase:
      type: object
      required: [channel, recipient, message]
      properties:
        channel:
          type: string
        recipient:
          type: string
        message:
          type: string
        scheduledAt:
          type: string
          format: date-time

    EmailNotification:
      allOf:
        - $ref: '#/components/schemas/NotificationBase'
        - type: object
          required: [subject]
          properties:
            channel:
              const: email
            subject:
              type: string
            htmlBody:
              type: string
            attachments:
              type: array
              items:
                type: string
                format: uri

    SmsNotification:
      allOf:
        - $ref: '#/components/schemas/NotificationBase'
        - type: object
          properties:
            channel:
              const: sms
            phoneNumber:
              type: string
              pattern: "^\\+[1-9]\\d{1,14}$"

    PushNotification:
      allOf:
        - $ref: '#/components/schemas/NotificationBase'
        - type: object
          properties:
            channel:
              const: push
            deviceToken:
              type: string
            badge:
              type: integer
            sound:
              type: string
```

## Validation Tools

| Tool | Language | Purpose |
|------|---------|---------|
| `@readme/openapi-parser` | JavaScript | Parse and validate |
| `swagger-parser` | JavaScript | Parse, validate, dereference |
| `openapi-spec-validator` | Python | Validate against spec |
| `vacuum` | Go | Fast validation and linting |
| `spectral` | JavaScript | Linting with custom rules |
| `redocly` | JavaScript | Validate, bundle, lint |

### Spectral Ruleset Example

```yaml
# .spectral.yaml
extends: ["spectral:oas"]
rules:
  operation-operationId: error
  operation-tags: error
  operation-description: warn
  info-description: error
  oas3-api-servers: error
  no-eval-in-markdown: error
  typed-enum: error
  # Custom rules
  must-have-examples:
    given: "$.paths.*.*.responses.*.content.application/json"
    then:
      field: example
      function: truthy
    severity: warn
    message: "Response should include an example"
```

## Code Generation

Generate client libraries and server stubs from OpenAPI specs:

| Generator | Output | Command |
|-----------|--------|---------|
| openapi-generator | 50+ languages | `openapi-generator generate -i spec.yaml -g typescript-fetch -o ./client` |
| openapi-typescript | TypeScript types | `npx openapi-typescript spec.yaml -o types.ts` |
| orval | React Query/SWR hooks | `npx orval --input spec.yaml --output ./api` |
| swagger-codegen | 40+ languages | `swagger-codegen generate -i spec.yaml -l python -o ./client` |
| oapi-codegen | Go server/client | `oapi-codegen -package api spec.yaml > api.gen.go` |

## Documentation Renderers

| Renderer | Type | Features |
|----------|------|----------|
| Swagger UI | Interactive | Try-it-out, auth support |
| Redoc | Static | Beautiful layout, deep linking |
| Stoplight Elements | Interactive | Modern UI, mock server |
| RapiDoc | Interactive | Customizable, dark mode |
| Scalar | Interactive | Modern design, themes |
