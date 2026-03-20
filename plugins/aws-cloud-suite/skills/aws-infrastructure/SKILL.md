---
name: aws-infrastructure
description: >
  AWS infrastructure patterns — VPC networking, ECS/Fargate, RDS, ElastiCache,
  CloudFront CDN, Route53 DNS, IAM best practices, and cost optimization.
  Triggers: "aws vpc", "ecs fargate", "rds aurora", "cloudfront",
  "route53", "iam policy", "aws networking", "aws cost optimization".
  NOT for: Lambda, DynamoDB, or serverless (use aws-lambda-serverless).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# AWS Infrastructure Patterns

## VPC Networking

```typescript
// CDK VPC with public + private subnets
import * as ec2 from "aws-cdk-lib/aws-ec2";

const vpc = new ec2.Vpc(this, "AppVpc", {
  maxAzs: 3,
  natGateways: 1,  // Cost optimization: 1 NAT instead of 3
  subnetConfiguration: [
    {
      name: "Public",
      subnetType: ec2.SubnetType.PUBLIC,
      cidrMask: 24,
    },
    {
      name: "Private",
      subnetType: ec2.SubnetType.PRIVATE_WITH_EGRESS,
      cidrMask: 24,
    },
    {
      name: "Isolated",
      subnetType: ec2.SubnetType.PRIVATE_ISOLATED,
      cidrMask: 24,
    },
  ],
});
```

### Subnet Strategy

| Subnet Type | Internet Access | Use For |
|-------------|----------------|---------|
| **Public** | Direct (IGW) | Load balancers, bastion hosts, NAT gateways |
| **Private with NAT** | Outbound only (NAT GW) | App servers, ECS tasks, Lambda |
| **Isolated** | None | Databases, ElastiCache, internal services |

### Security Groups

```typescript
// Web tier
const albSg = new ec2.SecurityGroup(this, "AlbSg", { vpc });
albSg.addIngressRule(ec2.Peer.anyIpv4(), ec2.Port.tcp(443), "HTTPS");
albSg.addIngressRule(ec2.Peer.anyIpv4(), ec2.Port.tcp(80), "HTTP redirect");

// App tier — only from ALB
const appSg = new ec2.SecurityGroup(this, "AppSg", { vpc });
appSg.addIngressRule(albSg, ec2.Port.tcp(3000), "From ALB");

// Database tier — only from app tier
const dbSg = new ec2.SecurityGroup(this, "DbSg", { vpc });
dbSg.addIngressRule(appSg, ec2.Port.tcp(5432), "Postgres from app");

// Redis — only from app tier
const redisSg = new ec2.SecurityGroup(this, "RedisSg", { vpc });
redisSg.addIngressRule(appSg, ec2.Port.tcp(6379), "Redis from app");
```

## ECS/Fargate

```typescript
import * as ecs from "aws-cdk-lib/aws-ecs";
import * as ecsPatterns from "aws-cdk-lib/aws-ecs-patterns";
import * as ecr from "aws-cdk-lib/aws-ecr";

// ECS Cluster
const cluster = new ecs.Cluster(this, "Cluster", {
  vpc,
  containerInsights: true,
});

// Fargate service with ALB (the 80% use case)
const service = new ecsPatterns.ApplicationLoadBalancedFargateService(this, "WebService", {
  cluster,
  cpu: 512,           // 0.5 vCPU
  memoryLimitMiB: 1024,  // 1 GB
  desiredCount: 2,
  taskImageOptions: {
    image: ecs.ContainerImage.fromEcrRepository(
      ecr.Repository.fromRepositoryName(this, "Repo", "my-app")
    ),
    containerPort: 3000,
    environment: {
      NODE_ENV: "production",
      DATABASE_URL: databaseUrl,
    },
    secrets: {
      API_SECRET: ecs.Secret.fromSecretsManager(apiSecret),
    },
  },
  publicLoadBalancer: true,
  healthCheck: {
    command: ["CMD-SHELL", "curl -f http://localhost:3000/health || exit 1"],
    interval: cdk.Duration.seconds(30),
    timeout: cdk.Duration.seconds(5),
    retries: 3,
  },
});

// Auto-scaling
const scaling = service.service.autoScaleTaskCount({ maxCapacity: 10 });
scaling.scaleOnCpuUtilization("CpuScaling", {
  targetUtilizationPercent: 70,
  scaleInCooldown: cdk.Duration.seconds(60),
  scaleOutCooldown: cdk.Duration.seconds(60),
});
scaling.scaleOnRequestCount("RequestScaling", {
  requestsPerTarget: 1000,
  targetGroup: service.targetGroup,
});
```

### ECS Task Definition (Direct)

```typescript
const taskDef = new ecs.FargateTaskDefinition(this, "TaskDef", {
  cpu: 1024,
  memoryLimitMiB: 2048,
});

// App container
const appContainer = taskDef.addContainer("app", {
  image: ecs.ContainerImage.fromEcrRepository(repo, "latest"),
  logging: ecs.LogDrivers.awsLogs({ streamPrefix: "app" }),
  healthCheck: {
    command: ["CMD-SHELL", "curl -f http://localhost:3000/health || exit 1"],
    interval: cdk.Duration.seconds(30),
    retries: 3,
  },
  environment: { NODE_ENV: "production" },
  secrets: {
    DB_PASSWORD: ecs.Secret.fromSecretsManager(dbSecret, "password"),
  },
});

appContainer.addPortMappings({ containerPort: 3000 });

// Sidecar container (e.g., log router, metrics)
taskDef.addContainer("datadog", {
  image: ecs.ContainerImage.fromRegistry("datadog/agent:latest"),
  memoryLimitMiB: 256,
  environment: { DD_API_KEY: "xxx" },
  essential: false,  // Don't kill task if sidecar fails
});
```

## RDS / Aurora

```typescript
import * as rds from "aws-cdk-lib/aws-rds";

// Aurora Serverless v2 (auto-scaling, pay-per-use)
const dbCluster = new rds.DatabaseCluster(this, "Database", {
  engine: rds.DatabaseClusterEngine.auroraPostgres({
    version: rds.AuroraPostgresEngineVersion.VER_15_4,
  }),
  serverlessV2MinCapacity: 0.5,   // Min ACUs
  serverlessV2MaxCapacity: 8,     // Max ACUs
  writer: rds.ClusterInstance.serverlessV2("writer"),
  readers: [
    rds.ClusterInstance.serverlessV2("reader", { scaleWithWriter: true }),
  ],
  vpc,
  vpcSubnets: { subnetType: ec2.SubnetType.PRIVATE_ISOLATED },
  securityGroups: [dbSg],
  defaultDatabaseName: "myapp",
  credentials: rds.Credentials.fromGeneratedSecret("postgres"),
  backup: { retention: cdk.Duration.days(7) },
  deletionProtection: true,
});

// Standard RDS (predictable workloads)
const dbInstance = new rds.DatabaseInstance(this, "Database", {
  engine: rds.DatabaseInstanceEngine.postgres({
    version: rds.PostgresEngineVersion.VER_16,
  }),
  instanceType: ec2.InstanceType.of(ec2.InstanceClass.T4G, ec2.InstanceSize.MEDIUM),
  vpc,
  vpcSubnets: { subnetType: ec2.SubnetType.PRIVATE_ISOLATED },
  securityGroups: [dbSg],
  multiAz: true,
  storageEncrypted: true,
  autoMinorVersionUpgrade: true,
  backupRetention: cdk.Duration.days(7),
  deletionProtection: true,
  credentials: rds.Credentials.fromGeneratedSecret("postgres"),
});
```

## ElastiCache (Redis)

```typescript
import * as elasticache from "aws-cdk-lib/aws-elasticache";

const subnetGroup = new elasticache.CfnSubnetGroup(this, "RedisSubnets", {
  description: "Redis subnet group",
  subnetIds: vpc.selectSubnets({ subnetType: ec2.SubnetType.PRIVATE_ISOLATED }).subnetIds,
});

const redis = new elasticache.CfnReplicationGroup(this, "Redis", {
  replicationGroupDescription: "App cache",
  engine: "redis",
  cacheNodeType: "cache.t4g.micro",
  numCacheClusters: 2,  // 1 primary + 1 replica
  automaticFailoverEnabled: true,
  multiAzEnabled: true,
  cacheSubnetGroupName: subnetGroup.ref,
  securityGroupIds: [redisSg.securityGroupId],
  atRestEncryptionEnabled: true,
  transitEncryptionEnabled: true,
  engineVersion: "7.0",
});
```

## CloudFront CDN

```typescript
import * as cloudfront from "aws-cdk-lib/aws-cloudfront";
import * as origins from "aws-cdk-lib/aws-cloudfront-origins";
import * as s3 from "aws-cdk-lib/aws-s3";
import * as acm from "aws-cdk-lib/aws-certificatemanager";

const siteBucket = new s3.Bucket(this, "SiteBucket", {
  blockPublicAccess: s3.BlockPublicAccess.BLOCK_ALL,
  encryption: s3.BucketEncryption.S3_MANAGED,
  removalPolicy: cdk.RemovalPolicy.RETAIN,
});

const certificate = acm.Certificate.fromCertificateArn(
  this, "Cert",
  "arn:aws:acm:us-east-1:123456789:certificate/abc-123"
);

const distribution = new cloudfront.Distribution(this, "CDN", {
  defaultBehavior: {
    origin: new origins.S3BucketOrigin(siteBucket),
    viewerProtocolPolicy: cloudfront.ViewerProtocolPolicy.REDIRECT_TO_HTTPS,
    cachePolicy: cloudfront.CachePolicy.CACHING_OPTIMIZED,
  },
  additionalBehaviors: {
    "/api/*": {
      origin: new origins.HttpOrigin("api.example.com"),
      cachePolicy: cloudfront.CachePolicy.CACHING_DISABLED,
      originRequestPolicy: cloudfront.OriginRequestPolicy.ALL_VIEWER_EXCEPT_HOST_HEADER,
      allowedMethods: cloudfront.AllowedMethods.ALLOW_ALL,
    },
  },
  domainNames: ["example.com", "www.example.com"],
  certificate,
  defaultRootObject: "index.html",
  errorResponses: [
    {
      httpStatus: 404,
      responsePagePath: "/index.html",  // SPA fallback
      responseHttpStatus: 200,
      ttl: cdk.Duration.seconds(0),
    },
  ],
});
```

## IAM Best Practices

```typescript
// Least privilege policy
const policy = new iam.PolicyDocument({
  statements: [
    new iam.PolicyStatement({
      effect: iam.Effect.ALLOW,
      actions: [
        "dynamodb:GetItem",
        "dynamodb:PutItem",
        "dynamodb:Query",
        "dynamodb:UpdateItem",
      ],
      resources: [
        table.tableArn,
        `${table.tableArn}/index/*`,
      ],
    }),
    new iam.PolicyStatement({
      effect: iam.Effect.ALLOW,
      actions: ["s3:GetObject", "s3:PutObject"],
      resources: [`${bucket.bucketArn}/*`],
      conditions: {
        StringEquals: { "s3:x-amz-server-side-encryption": "AES256" },
      },
    }),
    new iam.PolicyStatement({
      effect: iam.Effect.ALLOW,
      actions: ["secretsmanager:GetSecretValue"],
      resources: [secret.secretArn],
    }),
  ],
});
```

### IAM Rules

| Do | Don't |
|----|-------|
| Use IAM roles for services | Use access keys for services |
| Scope to specific resources | Use `Resource: "*"` |
| Use condition keys | Grant blanket `s3:*` |
| Rotate access keys (90 days) | Share credentials |
| Enable MFA for console users | Use root account for daily work |
| Use permission boundaries | Grant AdministratorAccess |

## Cost Optimization

| Resource | Cost Tip |
|----------|----------|
| **NAT Gateway** | $0.045/hr + $0.045/GB. Use 1 NAT (not 3). Consider NAT instances or VPC endpoints for S3/DynamoDB. |
| **RDS** | Use Aurora Serverless v2 for variable loads. Reserved instances for steady workloads (40-60% savings). |
| **Lambda** | ARM64 (Graviton2) is 20% cheaper. Use provisioned concurrency only where needed. |
| **ECS/Fargate** | Fargate Spot for non-critical tasks (70% discount). Right-size CPU/memory. |
| **S3** | Lifecycle rules: Standard → IA (30 days) → Glacier (90 days). Delete incomplete multipart uploads. |
| **CloudFront** | Cache static assets aggressively. Compress with gzip/brotli. Use Price Class 100 if US/EU only. |
| **Data Transfer** | Use VPC endpoints for S3/DynamoDB (free). CloudFront is cheaper than direct ALB for egress. |

## AWS CLI Quick Reference

```bash
# ECS
aws ecs list-services --cluster my-cluster
aws ecs update-service --cluster my-cluster --service my-svc --force-new-deployment
aws ecs describe-tasks --cluster my-cluster --tasks $(aws ecs list-tasks --cluster my-cluster --service my-svc --query 'taskArns[0]' --output text)

# RDS
aws rds describe-db-instances --query 'DBInstances[*].[DBInstanceIdentifier,DBInstanceStatus,Endpoint.Address]' --output table
aws rds create-db-snapshot --db-instance-identifier my-db --db-snapshot-identifier my-backup

# CloudWatch Logs
aws logs tail /ecs/my-app --follow --since 1h
aws logs filter-log-events --log-group-name /ecs/my-app --filter-pattern "ERROR" --start-time $(date -d '1 hour ago' +%s000)

# Secrets Manager
aws secretsmanager create-secret --name my-app/db --secret-string '{"password":"xxx"}'
aws secretsmanager get-secret-value --secret-id my-app/db --query 'SecretString' --output text

# S3
aws s3 sync ./dist s3://my-bucket --delete --cache-control "max-age=31536000"
aws s3 presign s3://my-bucket/file.pdf --expires-in 3600
```

## Gotchas

1. **NAT Gateway costs add up fast** — $0.045/hr per gateway + data processing charges. A 3-AZ VPC with 3 NAT gateways costs ~$100/mo before any traffic. Use 1 NAT gateway for non-production environments.

2. **ACM certificates for CloudFront must be in us-east-1** — CloudFront only accepts certificates from the us-east-1 region, regardless of where your stack deploys. Create the cert separately or use a cross-region reference.

3. **Security group rules are stateful** — If you allow inbound traffic on port 443, the response traffic is automatically allowed outbound. Don't add redundant outbound rules for responses.

4. **RDS deletion protection doesn't prevent CloudFormation deletion** — `deletionProtection: true` prevents manual API/console deletion. But a `cdk destroy` or stack deletion will still try to delete it. Set `removalPolicy: RETAIN` for true protection.

5. **ECS task role vs execution role** — The task role is what your application code uses (DynamoDB, S3 access). The execution role is what ECS uses to pull images and write logs. Don't put application permissions on the execution role.

6. **VPC endpoints save money AND latency** — Gateway endpoints for S3 and DynamoDB are free and route traffic within the VPC instead of through the NAT gateway. Always create them.
