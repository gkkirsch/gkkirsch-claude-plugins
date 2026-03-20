---
name: aws-lambda-serverless
description: >
  AWS serverless architecture — Lambda functions, API Gateway, DynamoDB,
  S3 operations, SQS/SNS, Step Functions, SAM templates, and CDK constructs.
  Triggers: "aws lambda", "serverless", "api gateway", "dynamodb",
  "sqs sns", "step functions", "sam template", "cdk lambda".
  NOT for: EC2, ECS, VPC, or traditional infrastructure (use aws-infrastructure).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# AWS Lambda & Serverless

## Lambda Handler Patterns

```typescript
// Basic handler with typed event
import { APIGatewayProxyHandler, APIGatewayProxyResult } from "aws-lambda";

export const handler: APIGatewayProxyHandler = async (event) => {
  try {
    const body = JSON.parse(event.body || "{}");
    const userId = event.pathParameters?.userId;
    const page = event.queryStringParameters?.page || "1";

    const result = await processRequest(userId, body);

    return {
      statusCode: 200,
      headers: {
        "Content-Type": "application/json",
        "Access-Control-Allow-Origin": "*",
      },
      body: JSON.stringify(result),
    };
  } catch (error) {
    console.error("Handler error:", error);
    return {
      statusCode: error instanceof ValidationError ? 400 : 500,
      body: JSON.stringify({
        error: error instanceof Error ? error.message : "Internal server error",
      }),
    };
  }
};

// SQS handler (batch processing)
import { SQSHandler, SQSBatchResponse } from "aws-lambda";

export const sqsHandler: SQSHandler = async (event) => {
  const batchItemFailures: SQSBatchResponse["batchItemFailures"] = [];

  for (const record of event.Records) {
    try {
      const body = JSON.parse(record.body);
      await processMessage(body);
    } catch (error) {
      console.error(`Failed record ${record.messageId}:`, error);
      batchItemFailures.push({ itemIdentifier: record.messageId });
    }
  }

  // Return partial failures (requires ReportBatchItemFailures)
  return { batchItemFailures };
};

// S3 trigger handler
import { S3Handler } from "aws-lambda";

export const s3Handler: S3Handler = async (event) => {
  for (const record of event.Records) {
    const bucket = record.s3.bucket.name;
    const key = decodeURIComponent(record.s3.object.key.replace(/\+/g, " "));
    const size = record.s3.object.size;

    console.log(`Processing: s3://${bucket}/${key} (${size} bytes)`);
    await processS3Object(bucket, key);
  }
};
```

## DynamoDB Operations

```typescript
import { DynamoDBClient } from "@aws-sdk/client-dynamodb";
import {
  DynamoDBDocumentClient,
  GetCommand,
  PutCommand,
  QueryCommand,
  UpdateCommand,
  DeleteCommand,
  BatchWriteCommand,
} from "@aws-sdk/lib-dynamodb";

const client = DynamoDBDocumentClient.from(new DynamoDBClient({}));

const TABLE = process.env.TABLE_NAME!;

// Get item
async function getUser(userId: string) {
  const { Item } = await client.send(
    new GetCommand({ TableName: TABLE, Key: { PK: `USER#${userId}`, SK: `PROFILE` } })
  );
  return Item || null;
}

// Put item (create or overwrite)
async function createUser(user: User) {
  await client.send(
    new PutCommand({
      TableName: TABLE,
      Item: {
        PK: `USER#${user.id}`,
        SK: "PROFILE",
        ...user,
        GSI1PK: `EMAIL#${user.email}`,
        GSI1SK: user.createdAt,
        TTL: Math.floor(Date.now() / 1000) + 86400 * 365,  // 1 year
      },
      ConditionExpression: "attribute_not_exists(PK)",  // Prevent overwrite
    })
  );
}

// Query (single partition)
async function getUserPosts(userId: string, limit = 20, lastKey?: Record<string, any>) {
  const { Items, LastEvaluatedKey } = await client.send(
    new QueryCommand({
      TableName: TABLE,
      KeyConditionExpression: "PK = :pk AND begins_with(SK, :sk)",
      ExpressionAttributeValues: { ":pk": `USER#${userId}`, ":sk": "POST#" },
      ScanIndexForward: false,  // Newest first
      Limit: limit,
      ExclusiveStartKey: lastKey,
    })
  );
  return { items: Items || [], nextKey: LastEvaluatedKey };
}

// Update with conditional expression
async function incrementViewCount(postId: string) {
  const { Attributes } = await client.send(
    new UpdateCommand({
      TableName: TABLE,
      Key: { PK: `POST#${postId}`, SK: "META" },
      UpdateExpression: "SET viewCount = if_not_exists(viewCount, :zero) + :inc, updatedAt = :now",
      ExpressionAttributeValues: {
        ":zero": 0,
        ":inc": 1,
        ":now": new Date().toISOString(),
      },
      ReturnValues: "ALL_NEW",
    })
  );
  return Attributes;
}

// Batch write (max 25 items per call)
async function batchCreateItems(items: Record<string, any>[]) {
  const chunks = [];
  for (let i = 0; i < items.length; i += 25) {
    chunks.push(items.slice(i, i + 25));
  }

  for (const chunk of chunks) {
    await client.send(
      new BatchWriteCommand({
        RequestItems: {
          [TABLE]: chunk.map((item) => ({
            PutRequest: { Item: item },
          })),
        },
      })
    );
  }
}
```

### Single-Table Design Patterns

| Access Pattern | PK | SK | GSI1PK | GSI1SK |
|---------------|----|----|--------|--------|
| Get user profile | `USER#<id>` | `PROFILE` | `EMAIL#<email>` | `<createdAt>` |
| Get user posts | `USER#<id>` | `POST#<timestamp>#<postId>` | `STATUS#published` | `<timestamp>` |
| Get post by ID | `POST#<id>` | `META` | | |
| Get post comments | `POST#<id>` | `COMMENT#<timestamp>#<commentId>` | | |

## API Gateway + Lambda (SAM)

```yaml
# template.yaml
AWSTemplateFormatVersion: "2010-09-09"
Transform: AWS::Serverless-2016-10-31
Description: API with Lambda and DynamoDB

Globals:
  Function:
    Runtime: nodejs20.x
    Timeout: 30
    MemorySize: 256
    Environment:
      Variables:
        TABLE_NAME: !Ref DataTable
        NODE_OPTIONS: "--enable-source-maps"
    Tracing: Active  # X-Ray tracing

Resources:
  Api:
    Type: AWS::Serverless::Api
    Properties:
      StageName: prod
      Cors:
        AllowOrigin: "'*'"
        AllowMethods: "'GET,POST,PUT,DELETE,OPTIONS'"
        AllowHeaders: "'Content-Type,Authorization'"
      Auth:
        DefaultAuthorizer: CognitoAuthorizer
        Authorizers:
          CognitoAuthorizer:
            UserPoolArn: !GetAtt UserPool.Arn

  GetUsersFunction:
    Type: AWS::Serverless::Function
    Properties:
      Handler: dist/handlers/users.list
      Events:
        Api:
          Type: Api
          Properties:
            RestApiId: !Ref Api
            Path: /users
            Method: GET
      Policies:
        - DynamoDBReadPolicy:
            TableName: !Ref DataTable

  CreateUserFunction:
    Type: AWS::Serverless::Function
    Properties:
      Handler: dist/handlers/users.create
      Events:
        Api:
          Type: Api
          Properties:
            RestApiId: !Ref Api
            Path: /users
            Method: POST
      Policies:
        - DynamoDBCrudPolicy:
            TableName: !Ref DataTable

  ProcessQueueFunction:
    Type: AWS::Serverless::Function
    Properties:
      Handler: dist/handlers/queue.process
      Timeout: 300
      Events:
        SQS:
          Type: SQS
          Properties:
            Queue: !GetAtt ProcessingQueue.Arn
            BatchSize: 10
            FunctionResponseTypes:
              - ReportBatchItemFailures

  DataTable:
    Type: AWS::DynamoDB::Table
    Properties:
      TableName: !Sub "${AWS::StackName}-data"
      BillingMode: PAY_PER_REQUEST
      AttributeDefinitions:
        - { AttributeName: PK, AttributeType: S }
        - { AttributeName: SK, AttributeType: S }
        - { AttributeName: GSI1PK, AttributeType: S }
        - { AttributeName: GSI1SK, AttributeType: S }
      KeySchema:
        - { AttributeName: PK, KeyType: HASH }
        - { AttributeName: SK, KeyType: RANGE }
      GlobalSecondaryIndexes:
        - IndexName: GSI1
          KeySchema:
            - { AttributeName: GSI1PK, KeyType: HASH }
            - { AttributeName: GSI1SK, KeyType: RANGE }
          Projection: { ProjectionType: ALL }
      TimeToLiveSpecification:
        AttributeName: TTL
        Enabled: true

  ProcessingQueue:
    Type: AWS::SQS::Queue
    Properties:
      VisibilityTimeout: 360  # 6x Lambda timeout
      RedrivePolicy:
        deadLetterTargetArn: !GetAtt DeadLetterQueue.Arn
        maxReceiveCount: 3

  DeadLetterQueue:
    Type: AWS::SQS::Queue
    Properties:
      MessageRetentionPeriod: 1209600  # 14 days

Outputs:
  ApiUrl:
    Value: !Sub "https://${Api}.execute-api.${AWS::Region}.amazonaws.com/prod"
```

## SAM CLI Commands

```bash
# Initialize new project
sam init --runtime nodejs20.x --app-template hello-world --name my-api

# Local development
sam build                              # Build the project
sam local start-api                    # Start local API Gateway
sam local invoke GetUsersFunction      # Invoke single function
sam local invoke -e events/get.json    # With test event

# Deploy
sam deploy --guided                    # First deploy (interactive)
sam deploy                             # Subsequent deploys
sam deploy --parameter-overrides Stage=prod  # With parameters

# Logs and debugging
sam logs -n GetUsersFunction --tail    # Stream CloudWatch logs
sam logs -n GetUsersFunction --filter "ERROR"  # Filter logs
```

## CDK Constructs

```typescript
import * as cdk from "aws-cdk-lib";
import * as lambda from "aws-cdk-lib/aws-lambda";
import * as apigateway from "aws-cdk-lib/aws-apigateway";
import * as dynamodb from "aws-cdk-lib/aws-dynamodb";
import * as sqs from "aws-cdk-lib/aws-sqs";
import * as lambdaEventSources from "aws-cdk-lib/aws-lambda-event-sources";
import { NodejsFunction } from "aws-cdk-lib/aws-lambda-nodejs";

export class ApiStack extends cdk.Stack {
  constructor(scope: cdk.App, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    // DynamoDB table
    const table = new dynamodb.Table(this, "DataTable", {
      partitionKey: { name: "PK", type: dynamodb.AttributeType.STRING },
      sortKey: { name: "SK", type: dynamodb.AttributeType.STRING },
      billingMode: dynamodb.BillingMode.PAY_PER_REQUEST,
      removalPolicy: cdk.RemovalPolicy.RETAIN,
      timeToLiveAttribute: "TTL",
    });

    table.addGlobalSecondaryIndex({
      indexName: "GSI1",
      partitionKey: { name: "GSI1PK", type: dynamodb.AttributeType.STRING },
      sortKey: { name: "GSI1SK", type: dynamodb.AttributeType.STRING },
    });

    // Lambda with esbuild bundling
    const apiHandler = new NodejsFunction(this, "ApiHandler", {
      entry: "src/handlers/api.ts",
      runtime: lambda.Runtime.NODEJS_20_X,
      memorySize: 256,
      timeout: cdk.Duration.seconds(30),
      environment: { TABLE_NAME: table.tableName },
      bundling: { minify: true, sourceMap: true },
      tracing: lambda.Tracing.ACTIVE,
    });

    table.grantReadWriteData(apiHandler);

    // API Gateway
    const api = new apigateway.RestApi(this, "Api", {
      restApiName: "MyApi",
      defaultCorsPreflightOptions: {
        allowOrigins: apigateway.Cors.ALL_ORIGINS,
        allowMethods: apigateway.Cors.ALL_METHODS,
      },
    });

    const users = api.root.addResource("users");
    users.addMethod("GET", new apigateway.LambdaIntegration(apiHandler));
    users.addMethod("POST", new apigateway.LambdaIntegration(apiHandler));

    // SQS queue with DLQ
    const dlq = new sqs.Queue(this, "DLQ", {
      retentionPeriod: cdk.Duration.days(14),
    });

    const queue = new sqs.Queue(this, "ProcessingQueue", {
      visibilityTimeout: cdk.Duration.seconds(180),
      deadLetterQueue: { queue: dlq, maxReceiveCount: 3 },
    });

    // Lambda consuming from SQS
    const queueHandler = new NodejsFunction(this, "QueueHandler", {
      entry: "src/handlers/queue.ts",
      timeout: cdk.Duration.seconds(30),
      environment: { TABLE_NAME: table.tableName },
    });

    queueHandler.addEventSource(
      new lambdaEventSources.SqsEventSource(queue, {
        batchSize: 10,
        reportBatchItemFailures: true,
      })
    );

    table.grantReadWriteData(queueHandler);

    // Outputs
    new cdk.CfnOutput(this, "ApiUrl", { value: api.url });
    new cdk.CfnOutput(this, "QueueUrl", { value: queue.queueUrl });
  }
}
```

## S3 Operations

```typescript
import { S3Client, GetObjectCommand, PutObjectCommand } from "@aws-sdk/client-s3";
import { getSignedUrl } from "@aws-sdk/s3-request-presigner";

const s3 = new S3Client({});

// Upload
async function uploadFile(bucket: string, key: string, body: Buffer, contentType: string) {
  await s3.send(
    new PutObjectCommand({
      Bucket: bucket,
      Key: key,
      Body: body,
      ContentType: contentType,
      ServerSideEncryption: "AES256",
    })
  );
}

// Download
async function downloadFile(bucket: string, key: string): Promise<Buffer> {
  const { Body } = await s3.send(new GetObjectCommand({ Bucket: bucket, Key: key }));
  return Buffer.from(await Body!.transformToByteArray());
}

// Pre-signed upload URL (client uploads directly to S3)
async function getUploadUrl(bucket: string, key: string): Promise<string> {
  return getSignedUrl(
    s3,
    new PutObjectCommand({
      Bucket: bucket,
      Key: key,
      ContentType: "application/octet-stream",
    }),
    { expiresIn: 3600 }  // 1 hour
  );
}
```

## Gotchas

1. **Lambda cold starts** — First invocation is slow (100ms-2s). Keep bundles small, use provisioned concurrency for latency-sensitive functions, and avoid importing unused SDK modules.

2. **DynamoDB batch operations max 25 items** — `BatchWriteCommand` accepts max 25 items per call. You MUST chunk larger batches yourself. Unprocessed items are returned in `UnprocessedItems`.

3. **SQS visibility timeout must exceed Lambda timeout** — Set queue `VisibilityTimeout` to at least 6x your function timeout. Otherwise, messages reappear and get processed twice.

4. **API Gateway has a 29-second timeout** — You cannot increase this limit. For long-running operations, accept the request, return 202, and process async via SQS or Step Functions.

5. **DynamoDB `Query` scans within ONE partition** — It doesn't search across partitions. Your partition key must match the access pattern. Use GSIs for alternative access patterns.

6. **Lambda environment variables are plaintext** — Don't store secrets in env vars visible in the console. Use AWS Secrets Manager or SSM Parameter Store with `GetSecretValue` at runtime (cache the result in a module-level variable for reuse across invocations).
