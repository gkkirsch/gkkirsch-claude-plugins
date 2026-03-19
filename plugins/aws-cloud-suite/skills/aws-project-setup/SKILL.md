# AWS Project Setup Skill

Scaffold and configure AWS projects with CDK, SAM, or Terraform. Sets up project structure, CI/CD pipelines, IAM roles, and development tooling.

## Triggers

- "set up an AWS project"
- "create a new CDK project"
- "scaffold a SAM application"
- "initialize Terraform for AWS"
- "new serverless project"
- "aws project setup"
- "create aws infrastructure"

## Workflow

### Step 1: Gather Requirements

Ask the user:
1. **Project type**: API backend, data pipeline, static website, full-stack app, infrastructure-only
2. **IaC tool**: CDK (TypeScript), SAM, Terraform, or CloudFormation
3. **Compute model**: Serverless (Lambda), Containers (ECS Fargate), EC2, or mixed
4. **Database needs**: DynamoDB, RDS Aurora, ElastiCache, none
5. **Environments**: dev only, dev+staging, dev+staging+production

### Step 2: CDK Project Setup

If CDK is selected:

```bash
# Initialize CDK project
mkdir my-aws-project && cd my-aws-project
npx cdk init app --language typescript

# Install dependencies
npm install @aws-cdk/aws-lambda-nodejs constructs
npm install -D esbuild @types/aws-lambda
```

Create the project structure:

```
my-aws-project/
├── bin/
│   └── app.ts
├── lib/
│   ├── stacks/
│   │   ├── api-stack.ts
│   │   ├── database-stack.ts
│   │   └── monitoring-stack.ts
│   ├── constructs/
│   │   └── (reusable constructs)
│   └── config/
│       └── environments.ts
├── src/
│   ├── handlers/
│   │   ├── create-item.ts
│   │   └── get-item.ts
│   └── lib/
│       ├── types.ts
│       └── utils.ts
├── test/
│   ├── stacks/
│   │   └── api-stack.test.ts
│   └── handlers/
│       └── create-item.test.ts
├── cdk.json
├── tsconfig.json
├── jest.config.js
└── package.json
```

**Standard cdk.json configuration:**
```json
{
  "app": "npx ts-node --prefer-ts-exts bin/app.ts",
  "watch": {
    "include": ["**"],
    "exclude": [
      "README.md", "cdk*.json", "**/*.d.ts", "**/*.js",
      "tsconfig.json", "package*.json", "yarn.lock",
      "node_modules", "test"
    ]
  },
  "context": {
    "@aws-cdk/aws-lambda:recognizeVersionProps": true,
    "@aws-cdk/core:stackRelativeExports": true,
    "@aws-cdk/aws-apigateway:usagePlanKeyOrderInsensitiveId": true,
    "@aws-cdk/aws-ecs:arnFormatIncludesClusterName": true,
    "@aws-cdk/aws-iam:minimizePolicies": true,
    "@aws-cdk/core:validateSnapshotRemovalPolicy": true,
    "@aws-cdk/aws-s3:createDefaultLoggingPolicy": true,
    "@aws-cdk/aws-ec2:restrictDefaultSecurityGroup": true
  }
}
```

**Standard API stack template:**
```typescript
// lib/stacks/api-stack.ts
import { Stack, StackProps, Duration, RemovalPolicy, CfnOutput } from 'aws-cdk-lib';
import { Construct } from 'constructs';
import * as lambda from 'aws-cdk-lib/aws-lambda';
import * as nodejs from 'aws-cdk-lib/aws-lambda-nodejs';
import * as dynamodb from 'aws-cdk-lib/aws-dynamodb';
import * as apigatewayv2 from 'aws-cdk-lib/aws-apigatewayv2';
import * as integrations from 'aws-cdk-lib/aws-apigatewayv2-integrations';
import * as logs from 'aws-cdk-lib/aws-logs';

export interface ApiStackProps extends StackProps {
  environment: string;
}

export class ApiStack extends Stack {
  constructor(scope: Construct, id: string, props: ApiStackProps) {
    super(scope, id, props);

    // DynamoDB table
    const table = new dynamodb.Table(this, 'ItemsTable', {
      tableName: `${props.environment}-items`,
      partitionKey: { name: 'pk', type: dynamodb.AttributeType.STRING },
      sortKey: { name: 'sk', type: dynamodb.AttributeType.STRING },
      billingMode: dynamodb.BillingMode.PAY_PER_REQUEST,
      encryption: dynamodb.TableEncryption.AWS_MANAGED,
      pointInTimeRecovery: true,
      removalPolicy: props.environment === 'production'
        ? RemovalPolicy.RETAIN
        : RemovalPolicy.DESTROY,
    });

    // Shared Lambda configuration
    const lambdaDefaults: nodejs.NodejsFunctionProps = {
      runtime: lambda.Runtime.NODEJS_22_X,
      architecture: lambda.Architecture.ARM_64,
      memorySize: 256,
      timeout: Duration.seconds(30),
      tracing: lambda.Tracing.ACTIVE,
      environment: {
        TABLE_NAME: table.tableName,
        NODE_OPTIONS: '--enable-source-maps',
        POWERTOOLS_SERVICE_NAME: 'items-api',
        POWERTOOLS_LOG_LEVEL: props.environment === 'production' ? 'WARN' : 'DEBUG',
      },
      bundling: {
        minify: true,
        sourceMap: true,
        target: 'es2022',
        externalModules: ['@aws-sdk/*'],
      },
      logRetention: logs.RetentionDays.ONE_MONTH,
    };

    // Lambda functions
    const createItemFn = new nodejs.NodejsFunction(this, 'CreateItem', {
      ...lambdaDefaults,
      entry: 'src/handlers/create-item.ts',
      functionName: `${props.environment}-create-item`,
    });

    const getItemFn = new nodejs.NodejsFunction(this, 'GetItem', {
      ...lambdaDefaults,
      entry: 'src/handlers/get-item.ts',
      functionName: `${props.environment}-get-item`,
    });

    // Grant DynamoDB access
    table.grantReadWriteData(createItemFn);
    table.grantReadData(getItemFn);

    // HTTP API
    const api = new apigatewayv2.HttpApi(this, 'HttpApi', {
      apiName: `${props.environment}-items-api`,
      corsPreflight: {
        allowOrigins: ['*'],
        allowMethods: [apigatewayv2.CorsHttpMethod.GET, apigatewayv2.CorsHttpMethod.POST],
        allowHeaders: ['Content-Type', 'Authorization'],
      },
    });

    api.addRoutes({
      path: '/items',
      methods: [apigatewayv2.HttpMethod.POST],
      integration: new integrations.HttpLambdaIntegration('CreateItemIntegration', createItemFn),
    });

    api.addRoutes({
      path: '/items/{id}',
      methods: [apigatewayv2.HttpMethod.GET],
      integration: new integrations.HttpLambdaIntegration('GetItemIntegration', getItemFn),
    });

    new CfnOutput(this, 'ApiUrl', {
      value: api.url!,
      description: 'API Gateway URL',
    });
  }
}
```

### Step 3: SAM Project Setup

If SAM is selected:

```bash
# Initialize SAM project with custom template
sam init --runtime nodejs22.x --app-template hello-world \
  --name my-serverless-api --architecture arm64 --package-type Zip

# Or use quickstart templates
sam init --location gh:aws/aws-sam-cli-app-templates --name my-api
```

Create `samconfig.toml` for multi-environment deployments:

```toml
version = 0.1

[default.global.parameters]
stack_name = "dev-my-api"

[default.build.parameters]
cached = true
parallel = true

[default.deploy.parameters]
capabilities = "CAPABILITY_IAM CAPABILITY_AUTO_EXPAND"
confirm_changeset = true
resolve_s3 = true
s3_prefix = "my-api"
region = "us-east-1"
parameter_overrides = "Stage=dev"

[default.sync.parameters]
watch = true

[staging.deploy.parameters]
stack_name = "staging-my-api"
s3_prefix = "my-api-staging"
parameter_overrides = "Stage=staging"
confirm_changeset = true

[production.deploy.parameters]
stack_name = "prod-my-api"
s3_prefix = "my-api-prod"
parameter_overrides = "Stage=prod"
confirm_changeset = true
disable_rollback = false
```

### Step 4: Terraform Project Setup

If Terraform is selected:

```bash
mkdir -p my-infra/{modules/{vpc,ecs-service,rds},environments/{dev,staging,production},global/iam}
```

**Standard backend configuration:**
```hcl
# environments/dev/backend.tf
terraform {
  required_version = ">= 1.9"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }

  backend "s3" {
    bucket         = "mycompany-terraform-state"
    key            = "dev/terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true
    dynamodb_table = "terraform-state-locks"
  }
}

provider "aws" {
  region = "us-east-1"

  default_tags {
    tags = {
      Environment = "dev"
      ManagedBy   = "terraform"
      Project     = "my-project"
    }
  }
}
```

**Create the state management infrastructure first:**
```hcl
# global/state-backend/main.tf
resource "aws_s3_bucket" "state" {
  bucket = "mycompany-terraform-state"
}

resource "aws_s3_bucket_versioning" "state" {
  bucket = aws_s3_bucket.state.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "state" {
  bucket = aws_s3_bucket.state.id
  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "aws:kms"
    }
    bucket_key_enabled = true
  }
}

resource "aws_s3_bucket_public_access_block" "state" {
  bucket                  = aws_s3_bucket.state.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_dynamodb_table" "locks" {
  name         = "terraform-state-locks"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "LockID"

  attribute {
    name = "LockID"
    type = "S"
  }
}
```

### Step 5: CI/CD Pipeline Setup

**GitHub Actions for CDK:**
```yaml
# .github/workflows/deploy.yml
name: Deploy
on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

permissions:
  id-token: write
  contents: read

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '22'
          cache: 'npm'
      - run: npm ci
      - run: npm test
      - run: npx cdk synth

  deploy-dev:
    needs: test
    if: github.ref == 'refs/heads/main'
    runs-on: ubuntu-latest
    environment: dev
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '22'
          cache: 'npm'
      - uses: aws-actions/configure-aws-credentials@v4
        with:
          role-to-assume: ${{ secrets.AWS_ROLE_ARN_DEV }}
          aws-region: us-east-1
      - run: npm ci
      - run: npx cdk deploy --all --require-approval never -c env=dev

  deploy-production:
    needs: deploy-dev
    runs-on: ubuntu-latest
    environment: production
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '22'
          cache: 'npm'
      - uses: aws-actions/configure-aws-credentials@v4
        with:
          role-to-assume: ${{ secrets.AWS_ROLE_ARN_PROD }}
          aws-region: us-east-1
      - run: npm ci
      - run: npx cdk deploy --all --require-approval never -c env=production
```

### Step 6: Development Tooling

**Configure AWS CLI profiles:**
```ini
# ~/.aws/config
[profile dev]
sso_start_url = https://mycompany.awsapps.com/start
sso_region = us-east-1
sso_account_id = 111111111111
sso_role_name = DeveloperAccess
region = us-east-1
output = json

[profile staging]
sso_start_url = https://mycompany.awsapps.com/start
sso_region = us-east-1
sso_account_id = 222222222222
sso_role_name = DeveloperAccess
region = us-east-1

[profile production]
sso_start_url = https://mycompany.awsapps.com/start
sso_region = us-east-1
sso_account_id = 333333333333
sso_role_name = ReadOnlyAccess
region = us-east-1
```

**Local development with SAM:**
```bash
# Start local API
sam build && sam local start-api --warm-containers EAGER

# Invoke single function
sam local invoke CreateItemFunction --event events/create-item.json

# Generate sample events
sam local generate-event apigateway http-api-proxy \
  --method POST --path /items --body '{"name":"test"}' > events/create-item.json
```

**Local development with CDK:**
```bash
# Synthesize and diff before deploying
npx cdk synth -c env=dev
npx cdk diff -c env=dev

# Deploy with hot-swap for fast iterations (dev only!)
npx cdk deploy --hotswap -c env=dev

# Watch mode — auto-deploys on file changes
npx cdk watch -c env=dev
```

### Step 7: Useful Scripts

Add to `package.json`:
```json
{
  "scripts": {
    "build": "tsc",
    "test": "jest",
    "test:watch": "jest --watch",
    "cdk:synth": "cdk synth -c env=dev",
    "cdk:diff": "cdk diff -c env=dev",
    "cdk:deploy:dev": "cdk deploy --all -c env=dev",
    "cdk:deploy:staging": "cdk deploy --all -c env=staging --require-approval broadening",
    "cdk:deploy:prod": "cdk deploy --all -c env=production --require-approval broadening",
    "cdk:watch": "cdk watch -c env=dev",
    "local:api": "sam local start-api --warm-containers EAGER",
    "lint": "eslint 'src/**/*.ts' 'lib/**/*.ts'",
    "lint:fix": "eslint 'src/**/*.ts' 'lib/**/*.ts' --fix"
  }
}
```

## Checklist

After scaffolding, verify:
- [ ] Project builds without errors (`npm run build`)
- [ ] Tests pass (`npm test`)
- [ ] CDK synthesizes (`npx cdk synth`)
- [ ] `.gitignore` includes `cdk.out/`, `node_modules/`, `.aws-sam/`, `.terraform/`
- [ ] No hardcoded account IDs, regions, or secrets
- [ ] Environment configuration is externalized
- [ ] CI/CD pipeline is set up
- [ ] README with setup instructions exists
