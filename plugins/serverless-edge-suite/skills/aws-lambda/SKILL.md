---
name: aws-lambda
description: >
  Build and deploy AWS Lambda functions with Node.js/TypeScript — handlers,
  API Gateway integration, environment variables, layers, and deployment.
  Triggers: "AWS Lambda", "Lambda function", "serverless AWS", "API Gateway Lambda".
  NOT for: Cloudflare Workers, Vercel functions, non-AWS serverless.
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# AWS Lambda with Node.js/TypeScript

## Basic Lambda Handler

```typescript
// src/handlers/hello.ts
import type { APIGatewayProxyEvent, APIGatewayProxyResult } from 'aws-lambda';

export const handler = async (
  event: APIGatewayProxyEvent
): Promise<APIGatewayProxyResult> => {
  const name = event.queryStringParameters?.name ?? 'World';

  return {
    statusCode: 200,
    headers: {
      'Content-Type': 'application/json',
      'Access-Control-Allow-Origin': '*',
    },
    body: JSON.stringify({ message: `Hello, ${name}!` }),
  };
};
```

## REST API with API Gateway

```typescript
// src/handlers/users.ts
import type { APIGatewayProxyEvent, APIGatewayProxyResult } from 'aws-lambda';

export const handler = async (
  event: APIGatewayProxyEvent
): Promise<APIGatewayProxyResult> => {
  const { httpMethod, pathParameters, body } = event;

  try {
    switch (httpMethod) {
      case 'GET':
        if (pathParameters?.id) {
          return await getUser(pathParameters.id);
        }
        return await listUsers(event.queryStringParameters);

      case 'POST':
        return await createUser(JSON.parse(body ?? '{}'));

      case 'PUT':
        return await updateUser(pathParameters?.id!, JSON.parse(body ?? '{}'));

      case 'DELETE':
        return await deleteUser(pathParameters?.id!);

      default:
        return response(405, { error: 'Method not allowed' });
    }
  } catch (err) {
    console.error('Handler error:', err);
    return response(500, { error: 'Internal server error' });
  }
};

// Response helper
function response(statusCode: number, body: object): APIGatewayProxyResult {
  return {
    statusCode,
    headers: {
      'Content-Type': 'application/json',
      'Access-Control-Allow-Origin': '*',
      'Access-Control-Allow-Methods': 'GET,POST,PUT,DELETE,OPTIONS',
    },
    body: JSON.stringify(body),
  };
}
```

## Event Sources

### SQS Queue Handler

```typescript
import type { SQSEvent, SQSRecord } from 'aws-lambda';

export const handler = async (event: SQSEvent): Promise<void> => {
  const failed: string[] = [];

  for (const record of event.Records) {
    try {
      const body = JSON.parse(record.body);
      await processMessage(body);
    } catch (err) {
      console.error(`Failed to process ${record.messageId}:`, err);
      failed.push(record.messageId);
    }
  }

  // Partial batch failure reporting
  if (failed.length > 0) {
    return {
      batchItemFailures: failed.map(id => ({ itemIdentifier: id })),
    } as any;
  }
};
```

### S3 Event Handler

```typescript
import type { S3Event } from 'aws-lambda';

export const handler = async (event: S3Event): Promise<void> => {
  for (const record of event.Records) {
    const bucket = record.s3.bucket.name;
    const key = decodeURIComponent(record.s3.object.key.replace(/\+/g, ' '));
    const size = record.s3.object.size;

    console.log(`New file: s3://${bucket}/${key} (${size} bytes)`);
    await processFile(bucket, key);
  }
};
```

### Scheduled Event (Cron)

```typescript
import type { ScheduledEvent } from 'aws-lambda';

export const handler = async (event: ScheduledEvent): Promise<void> => {
  console.log('Cron triggered at:', event.time);
  await runDailyCleanup();
};

// In serverless.yml or SAM template:
// Schedule: rate(1 day)  or  cron(0 9 * * ? *)
```

## Middleware Pattern (Middy)

```bash
npm install @middy/core @middy/http-json-body-parser @middy/http-error-handler @middy/validator
```

```typescript
import middy from '@middy/core';
import httpJsonBodyParser from '@middy/http-json-body-parser';
import httpErrorHandler from '@middy/http-error-handler';
import createError from 'http-errors';
import type { APIGatewayProxyEvent, APIGatewayProxyResult } from 'aws-lambda';

const baseHandler = async (
  event: APIGatewayProxyEvent & { body: { name: string; email: string } }
): Promise<APIGatewayProxyResult> => {
  const { name, email } = event.body; // Already parsed by middleware

  if (!name || !email) {
    throw createError(400, 'Name and email are required');
  }

  const user = await createUser({ name, email });

  return {
    statusCode: 201,
    body: JSON.stringify(user),
  };
};

export const handler = middy(baseHandler)
  .use(httpJsonBodyParser())      // Parse JSON body
  .use(httpErrorHandler());       // Format error responses
```

### Custom Middleware

```typescript
const authMiddleware = (): middy.MiddlewareObj => ({
  before: async (request) => {
    const token = request.event.headers?.Authorization?.replace('Bearer ', '');
    if (!token) {
      throw createError(401, 'Authentication required');
    }

    const user = await verifyToken(token);
    request.event.requestContext.authorizer = { user };
  },
});

export const handler = middy(baseHandler)
  .use(httpJsonBodyParser())
  .use(authMiddleware())
  .use(httpErrorHandler());
```

## Environment Variables and Secrets

```typescript
// Access env vars directly
const DB_URL = process.env.DATABASE_URL!;
const API_KEY = process.env.STRIPE_API_KEY!;

// For secrets, use AWS Secrets Manager or SSM Parameter Store
import { SSMClient, GetParameterCommand } from '@aws-sdk/client-ssm';

const ssm = new SSMClient({});

// Cache secrets outside handler (persists across warm invocations)
let cachedSecret: string | null = null;

async function getSecret(name: string): Promise<string> {
  if (cachedSecret) return cachedSecret;

  const result = await ssm.send(new GetParameterCommand({
    Name: name,
    WithDecryption: true,
  }));

  cachedSecret = result.Parameter!.Value!;
  return cachedSecret;
}
```

## Deployment with SST (Recommended)

```bash
# Install
npx create-sst@latest my-api
cd my-api && npm install
```

```typescript
// sst.config.ts
export default $config({
  app(input) {
    return {
      name: 'my-api',
      removal: input?.stage === 'production' ? 'retain' : 'remove',
    };
  },
  async run() {
    const api = new sst.aws.ApiGatewayV2('Api');

    api.route('GET /users', 'src/handlers/users.handler');
    api.route('POST /users', 'src/handlers/users.handler');
    api.route('GET /users/{id}', 'src/handlers/users.handler');

    return { url: api.url };
  },
});
```

```bash
# Deploy
npx sst deploy --stage dev
npx sst deploy --stage production
```

## Deployment with SAM

```yaml
# template.yaml
AWSTemplateFormatVersion: '2010-09-09'
Transform: AWS::Serverless-2016-10-31

Globals:
  Function:
    Runtime: nodejs20.x
    Timeout: 30
    MemorySize: 256
    Environment:
      Variables:
        NODE_ENV: production

Resources:
  UsersFunction:
    Type: AWS::Serverless::Function
    Properties:
      Handler: dist/handlers/users.handler
      Events:
        ListUsers:
          Type: Api
          Properties:
            Path: /users
            Method: GET
        CreateUser:
          Type: Api
          Properties:
            Path: /users
            Method: POST
    Metadata:
      BuildMethod: esbuild
      BuildProperties:
        Minify: true
        Target: es2022
```

```bash
sam build && sam deploy --guided
```

## Cold Start Optimization

```typescript
// 1. Initialize SDK clients OUTSIDE the handler (reused across invocations)
import { DynamoDBClient } from '@aws-sdk/client-dynamodb';
import { DynamoDBDocumentClient } from '@aws-sdk/lib-dynamodb';

const client = new DynamoDBClient({});
const db = DynamoDBDocumentClient.from(client);

// 2. Use AWS SDK v3 (tree-shakeable, smaller bundles)
// BAD: import AWS from 'aws-sdk';              // ~60MB
// GOOD: import { S3Client } from '@aws-sdk/client-s3'; // ~2MB

// 3. Bundle with esbuild (in SAM, SST, or manual)
// Result: single .js file, no node_modules, fast cold start

export const handler = async (event: any) => {
  // db is already initialized (warm instance reuse)
  const result = await db.send(new GetCommand({ TableName: 'users', Key: { id: '123' } }));
  return { statusCode: 200, body: JSON.stringify(result.Item) };
};
```

## Gotchas

- **Handler must return or throw.** Lambda waits for the promise to resolve. If you forget to return, Lambda times out.
- **`/tmp` is your only writable filesystem.** Limited to 512 MB (configurable up to 10 GB). Cleared between invocations (sometimes).
- **Connection pooling across invocations.** Initialize DB connections outside the handler. They persist across warm invocations.
- **API Gateway has a 30-second timeout.** Even if Lambda allows 15 minutes, API Gateway will 504 after 30s. Use async patterns for long-running tasks.
- **`event.body` is a string.** API Gateway sends the body as a JSON string. Parse it with `JSON.parse()` or use Middy's `httpJsonBodyParser`.
- **CORS must be returned in headers.** API Gateway doesn't add CORS headers automatically (unless configured). Include them in every response.
- **Cold starts compound with VPC.** Lambda in a VPC adds 1-10 seconds to cold start (ENI attachment). Use VPC only if you need private resources.
- **Provisioned concurrency is expensive.** Only use it for latency-critical endpoints. $0.015 per GB-hour of provisioned concurrency.
