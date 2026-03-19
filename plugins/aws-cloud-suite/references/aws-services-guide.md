# AWS Core Services Deep Dive

Practical reference for the most-used AWS services. Covers real-world usage patterns, CLI examples, CloudFormation snippets, and critical gotchas.

---

## EC2 (Elastic Compute Cloud)

### Instance Selection Guide

```
Workload Type        │ Instance Family │ Recommended                │ Why
─────────────────────┼─────────────────┼────────────────────────────┼──────────────────
General web apps     │ M7g/M7i         │ m7g.xlarge                 │ Balanced CPU/memory
Memory-intensive     │ R7g/R7i         │ r7g.2xlarge                │ High memory:CPU ratio
Compute-intensive    │ C7g/C7i         │ c7g.2xlarge                │ High CPU:memory ratio
GPU/ML training      │ P5/P4d          │ p5.48xlarge                │ NVIDIA H100 GPUs
GPU/ML inference     │ Inf2/G5         │ inf2.xlarge                │ AWS Inferentia2
Storage-optimized    │ I4i/D3          │ i4i.xlarge                 │ NVMe SSD, high IOPS
Burstable dev/test   │ T4g             │ t4g.medium                 │ Cheapest, credit-based
HPC                  │ Hpc7g           │ hpc7g.16xlarge             │ EFA networking
```

**Always prefer Graviton (g suffix) — 20-40% better price-performance than Intel/AMD equivalents.**

### Launch Template (Best Practice)

```bash
aws ec2 create-launch-template \
  --launch-template-name prod-web-server \
  --version-description "v1.0 - AL2023 with SSM" \
  --launch-template-data '{
    "ImageId": "resolve:ssm:/aws/service/ami-amazon-linux-latest/al2023-ami-kernel-default-arm64",
    "InstanceType": "m7g.large",
    "KeyName": "",
    "SecurityGroupIds": ["sg-app"],
    "IamInstanceProfile": {"Name": "EC2SSMRole"},
    "BlockDeviceMappings": [
      {
        "DeviceName": "/dev/xvda",
        "Ebs": {
          "VolumeSize": 30,
          "VolumeType": "gp3",
          "Iops": 3000,
          "Throughput": 125,
          "Encrypted": true,
          "KmsKeyId": "alias/ebs-key",
          "DeleteOnTermination": true
        }
      }
    ],
    "MetadataOptions": {
      "HttpTokens": "required",
      "HttpPutResponseHopLimit": 1,
      "HttpEndpoint": "enabled",
      "InstanceMetadataTags": "enabled"
    },
    "Monitoring": {"Enabled": true},
    "TagSpecifications": [
      {
        "ResourceType": "instance",
        "Tags": [
          {"Key": "Environment", "Value": "production"},
          {"Key": "ManagedBy", "Value": "launch-template"}
        ]
      }
    ],
    "UserData": ""
  }'
```

**Critical: `HttpTokens: required` enforces IMDSv2 — prevents SSRF attacks from accessing instance metadata.**

### Auto Scaling Group

```yaml
AutoScalingGroup:
  Type: AWS::AutoScaling::AutoScalingGroup
  Properties:
    AutoScalingGroupName: prod-web-asg
    LaunchTemplate:
      LaunchTemplateId: !Ref LaunchTemplate
      Version: !GetAtt LaunchTemplate.LatestVersionNumber
    MinSize: 2
    MaxSize: 20
    DesiredCapacity: 3
    VPCZoneIdentifier:
      - !Ref PrivateSubnet1
      - !Ref PrivateSubnet2
      - !Ref PrivateSubnet3
    TargetGroupARNs:
      - !Ref ALBTargetGroup
    HealthCheckType: ELB
    HealthCheckGracePeriod: 120
    DefaultInstanceWarmup: 60
    MixedInstancesPolicy:
      InstancesDistribution:
        OnDemandBaseCapacity: 2
        OnDemandPercentageAboveBaseCapacity: 25
        SpotAllocationStrategy: price-capacity-optimized
      LaunchTemplate:
        LaunchTemplateSpecification:
          LaunchTemplateId: !Ref LaunchTemplate
          Version: !GetAtt LaunchTemplate.LatestVersionNumber
        Overrides:
          - InstanceType: m7g.large
          - InstanceType: m6g.large
          - InstanceType: m7i-flex.large
    Tags:
      - Key: Name
        Value: prod-web
        PropagateAtLaunch: true

# Target tracking scaling policy
ScalingPolicy:
  Type: AWS::AutoScaling::ScalingPolicy
  Properties:
    AutoScalingGroupName: !Ref AutoScalingGroup
    PolicyType: TargetTrackingScaling
    TargetTrackingConfiguration:
      PredefinedMetricSpecification:
        PredefinedMetricType: ALBRequestCountPerTarget
        ResourceLabel: !Sub "${ALB.FullName}/${ALBTargetGroup.TargetGroupFullName}"
      TargetValue: 1000
      DisableScaleIn: false
```

### EBS Volume Types Decision

```
Type     │ IOPS          │ Throughput      │ $/GB/mo │ Use Case
─────────┼───────────────┼─────────────────┼─────────┼──────────────────────
gp3      │ 3,000 (free)  │ 125 MB/s (free) │ $0.08   │ Default for everything
         │ up to 16,000  │ up to 1,000MB/s │         │
gp2      │ 3 × GB size   │ 250 MB/s        │ $0.10   │ Legacy, migrate to gp3
io2 BE   │ up to 256,000 │ 4,000 MB/s      │ $0.125  │ Critical databases
st1      │ 500 baseline  │ 500 MB/s        │ $0.045  │ Sequential reads (logs)
sc1      │ 250 baseline  │ 250 MB/s        │ $0.015  │ Infrequent access
```

**gp3 is almost always the right choice.** It's cheaper than gp2 and gives you 3,000 IOPS baseline for free.

---

## S3 (Simple Storage Service)

### Storage Class Selection

```
Class                  │ Availability │ Min Duration │ Retrieval     │ $/GB/mo │ When
───────────────────────┼──────────────┼──────────────┼───────────────┼─────────┼─────────────────
Standard               │ 99.99%       │ None         │ Instant       │ $0.023  │ Active data
Intelligent-Tiering    │ 99.9%        │ None         │ Instant       │ $0.023  │ Unknown access
Standard-IA            │ 99.9%        │ 30 days      │ Instant       │ $0.0125 │ Monthly access
One Zone-IA            │ 99.5%        │ 30 days      │ Instant       │ $0.01   │ Reproducible data
Glacier Instant        │ 99.9%        │ 90 days      │ Instant       │ $0.004  │ Quarterly access
Glacier Flexible       │ 99.99%       │ 90 days      │ 1-12 hours    │ $0.0036 │ Annual access
Glacier Deep Archive   │ 99.99%       │ 180 days     │ 12-48 hours   │ $0.00099│ Compliance/archive
Express One Zone       │ 99.95%       │ None         │ Single-digit ms│ $0.16  │ ML training data
```

### Lifecycle Policy

```json
{
  "Rules": [
    {
      "ID": "optimize-storage-costs",
      "Status": "Enabled",
      "Filter": {"Prefix": "data/"},
      "Transitions": [
        {
          "Days": 30,
          "StorageClass": "STANDARD_IA"
        },
        {
          "Days": 90,
          "StorageClass": "GLACIER_IR"
        },
        {
          "Days": 365,
          "StorageClass": "DEEP_ARCHIVE"
        }
      ],
      "NoncurrentVersionTransitions": [
        {
          "NoncurrentDays": 30,
          "StorageClass": "STANDARD_IA"
        }
      ],
      "NoncurrentVersionExpiration": {
        "NoncurrentDays": 90,
        "NewerNoncurrentVersions": 3
      },
      "AbortIncompleteMultipartUpload": {
        "DaysAfterInitiation": 7
      }
    },
    {
      "ID": "expire-temp-files",
      "Status": "Enabled",
      "Filter": {"Prefix": "tmp/"},
      "Expiration": {
        "Days": 1
      }
    }
  ]
}
```

### S3 Performance Optimization

```bash
# Multipart upload (required for >5GB, recommended for >100MB)
aws s3 cp large-file.tar.gz s3://bucket/large-file.tar.gz \
  --expected-size 10737418240 \
  --storage-class INTELLIGENT_TIERING

# Transfer Acceleration for cross-region uploads
aws s3api put-bucket-accelerate-configuration \
  --bucket my-bucket \
  --accelerate-configuration Status=Enabled

# Use the accelerate endpoint
aws s3 cp file.zip s3://my-bucket/file.zip --endpoint-url https://my-bucket.s3-accelerate.amazonaws.com

# S3 Select — query CSV/JSON/Parquet in-place without downloading
aws s3api select-object-content \
  --bucket analytics-data \
  --key "logs/2026/03/access.csv.gz" \
  --expression "SELECT s.timestamp, s.status_code FROM s3object s WHERE s.status_code >= '400'" \
  --expression-type SQL \
  --input-serialization '{"CSV":{"FileHeaderInfo":"USE"},"CompressionType":"GZIP"}' \
  --output-serialization '{"JSON":{}}' \
  output.json
```

### S3 Event Notifications

```yaml
Bucket:
  Type: AWS::S3::Bucket
  Properties:
    NotificationConfiguration:
      EventBridgeConfiguration:
        EventBridgeEnabled: true  # Preferred — sends ALL events to EventBridge
      LambdaConfigurations:
        - Event: "s3:ObjectCreated:*"
          Filter:
            S3Key:
              Rules:
                - Name: prefix
                  Value: uploads/
                - Name: suffix
                  Value: .csv
          Function: !GetAtt ProcessCSVFunction.Arn
```

---

## RDS & Aurora

### Aurora PostgreSQL Setup

```bash
# Create cluster
aws rds create-db-cluster \
  --db-cluster-identifier prod-aurora \
  --engine aurora-postgresql \
  --engine-version 16.4 \
  --master-username admin \
  --manage-master-user-password \
  --master-user-secret-kms-key-id alias/rds-secrets \
  --db-subnet-group-name private-subnets \
  --vpc-security-group-ids sg-database \
  --storage-encrypted \
  --kms-key-id alias/rds-data \
  --backup-retention-period 35 \
  --preferred-backup-window "03:00-04:00" \
  --preferred-maintenance-window "sun:04:00-sun:05:00" \
  --enable-cloudwatch-logs-exports '["postgresql"]' \
  --deletion-protection \
  --copy-tags-to-snapshot \
  --serverless-v2-scaling-configuration MinCapacity=0.5,MaxCapacity=64

# Create instances (writer + reader)
aws rds create-db-instance \
  --db-instance-identifier prod-aurora-writer \
  --db-cluster-identifier prod-aurora \
  --db-instance-class db.serverless \
  --engine aurora-postgresql \
  --monitoring-interval 60 \
  --monitoring-role-arn arn:aws:iam::123456789:role/rds-monitoring \
  --enable-performance-insights \
  --performance-insights-kms-key-id alias/rds-pi \
  --performance-insights-retention-period 731

aws rds create-db-instance \
  --db-instance-identifier prod-aurora-reader \
  --db-cluster-identifier prod-aurora \
  --db-instance-class db.serverless \
  --engine aurora-postgresql \
  --monitoring-interval 60 \
  --monitoring-role-arn arn:aws:iam::123456789:role/rds-monitoring \
  --enable-performance-insights \
  --performance-insights-kms-key-id alias/rds-pi
```

### RDS Proxy for Lambda

Lambda functions can exhaust database connection limits. RDS Proxy pools and manages connections:

```yaml
RDSProxy:
  Type: AWS::RDS::DBProxy
  Properties:
    DBProxyName: my-api-proxy
    EngineFamily: POSTGRESQL
    Auth:
      - AuthScheme: SECRETS
        SecretArn: !Ref DatabaseSecret
        IAMAuth: REQUIRED
    RoleArn: !GetAtt RDSProxyRole.Arn
    VpcSubnetIds:
      - !Ref PrivateSubnet1
      - !Ref PrivateSubnet2
    VpcSecurityGroupIds:
      - !Ref DatabaseSecurityGroup
    RequireTLS: true
    IdleClientTimeout: 1800

RDSProxyTargetGroup:
  Type: AWS::RDS::DBProxyTargetGroup
  Properties:
    DBProxyName: !Ref RDSProxy
    TargetGroupName: default
    DBClusterIdentifiers:
      - !Ref AuroraCluster
    ConnectionPoolConfigurationInfo:
      MaxConnectionsPercent: 90
      MaxIdleConnectionsPercent: 50
      ConnectionBorrowTimeout: 120
```

### Aurora vs RDS Decision

```
Factor              │ Aurora                      │ RDS
────────────────────┼─────────────────────────────┼──────────────────────────
Storage             │ Auto-scales to 128 TiB      │ Manual, up to 64 TiB
Replication         │ 6 copies across 3 AZs       │ 1 standby (Multi-AZ)
Read replicas       │ Up to 15, <20ms lag          │ Up to 15, minutes lag
Failover            │ <30 seconds                 │ 60-120 seconds
Serverless          │ V2 (scales to 0.5 ACU)      │ Not available
Global              │ <1s cross-region replication│ Cross-region read replicas
Cost                │ ~20% more than RDS           │ Baseline
When to choose      │ Production workloads, HA    │ Dev/test, simple apps
```

---

## DynamoDB

### Table Design Principles

**Identify access patterns BEFORE designing the table.**

```
Access Pattern                    │ Key Design                        │ Operation
──────────────────────────────────┼───────────────────────────────────┼──────────
Get user profile                  │ PK=USER#123, SK=PROFILE           │ GetItem
Get user's orders                 │ PK=USER#123, SK begins_with ORD   │ Query
Get order details                 │ PK=ORDER#456, SK=DETAILS          │ GetItem
Get order items                   │ PK=ORDER#456, SK begins_with ITEM │ Query
Orders by date (all users)        │ GSI1: PK=STATUS#pending, SK=date  │ Query GSI1
User by email                     │ GSI2: PK=EMAIL#user@co.com        │ Query GSI2
```

### Capacity Modes

```
Mode         │ Pricing                    │ When to Use
─────────────┼────────────────────────────┼────────────────────────────
On-Demand    │ $1.25/million writes       │ Unpredictable traffic,
             │ $0.25/million reads        │ new tables, spiky workloads
Provisioned  │ $0.00065/WCU/hour          │ Steady, predictable traffic
             │ $0.00013/RCU/hour          │ (30-70% cheaper with RIs)
```

### DynamoDB Transactions

```typescript
import { TransactWriteCommand } from '@aws-sdk/lib-dynamodb';

// Atomic transfer between two accounts
await docClient.send(new TransactWriteCommand({
  TransactItems: [
    {
      Update: {
        TableName: 'accounts',
        Key: { pk: 'ACCT#source', sk: 'BALANCE' },
        UpdateExpression: 'SET balance = balance - :amount',
        ConditionExpression: 'balance >= :amount',
        ExpressionAttributeValues: { ':amount': 100 },
      },
    },
    {
      Update: {
        TableName: 'accounts',
        Key: { pk: 'ACCT#dest', sk: 'BALANCE' },
        UpdateExpression: 'SET balance = balance + :amount',
        ExpressionAttributeValues: { ':amount': 100 },
      },
    },
    {
      Put: {
        TableName: 'transactions',
        Item: {
          pk: `TXN#${crypto.randomUUID()}`,
          sk: new Date().toISOString(),
          from: 'source',
          to: 'dest',
          amount: 100,
        },
        ConditionExpression: 'attribute_not_exists(pk)',
      },
    },
  ],
}));
```

### DynamoDB Best Practices

1. **Use `begins_with` on sort key** for hierarchical queries — it's O(log n) not O(n)
2. **Avoid Scan** — always use Query with partition key. Scans read the entire table
3. **Use projection expressions** — only fetch attributes you need
4. **Batch operations** for multiple items — `BatchWriteItem` handles up to 25 items
5. **TTL for automatic expiration** — no write cost for TTL deletions
6. **Enable Point-in-Time Recovery** — 35-day continuous backups
7. **Use PartiQL for SQL-like syntax** when it improves readability
8. **Sparse GSI** — only project items that have the GSI key attributes

---

## SQS (Simple Queue Service)

### Standard vs FIFO

```
Feature           │ Standard                │ FIFO
──────────────────┼─────────────────────────┼──────────────────────
Order             │ Best-effort             │ Strict FIFO
Delivery          │ At-least-once           │ Exactly-once
Throughput        │ Unlimited               │ 3,000 msg/sec (batch)
                  │                         │ 300 msg/sec (no batch)
Price (per 1M)    │ $0.40                   │ $0.50
Deduplication     │ No                      │ 5-minute window
Use case          │ Fan-out, async          │ Order-sensitive ops
```

### Dead Letter Queue Pattern

```yaml
MainQueue:
  Type: AWS::SQS::Queue
  Properties:
    QueueName: orders-queue
    VisibilityTimeout: 900  # 6x your Lambda timeout
    MessageRetentionPeriod: 1209600  # 14 days
    ReceiveMessageWaitTimeSeconds: 20  # Long polling
    RedrivePolicy:
      deadLetterTargetArn: !GetAtt DLQ.Arn
      maxReceiveCount: 3

DLQ:
  Type: AWS::SQS::Queue
  Properties:
    QueueName: orders-dlq
    MessageRetentionPeriod: 1209600
    # Alert when messages appear in DLQ
    RedriveAllowPolicy:
      redrivePermission: byQueue
      sourceQueueArns:
        - !GetAtt MainQueue.Arn

DLQAlarm:
  Type: AWS::CloudWatch::Alarm
  Properties:
    AlarmName: orders-dlq-messages
    MetricName: ApproximateNumberOfMessagesVisible
    Namespace: AWS/SQS
    Dimensions:
      - Name: QueueName
        Value: !GetAtt DLQ.QueueName
    Statistic: Sum
    Period: 300
    EvaluationPeriods: 1
    Threshold: 1
    ComparisonOperator: GreaterThanOrEqualToThreshold
    AlarmActions:
      - !Ref AlertTopic
```

---

## SNS (Simple Notification Service)

### Fan-Out Pattern

```yaml
OrderEventsTopic:
  Type: AWS::SNS::Topic
  Properties:
    TopicName: order-events
    KmsMasterKeyId: alias/sns-key

# Each subscriber gets its own copy of every message
InventorySubscription:
  Type: AWS::SNS::Subscription
  Properties:
    TopicArn: !Ref OrderEventsTopic
    Protocol: sqs
    Endpoint: !GetAtt InventoryQueue.Arn
    FilterPolicy:
      event_type: ["order_placed", "order_cancelled"]
    FilterPolicyScope: MessageBody
    RawMessageDelivery: true

NotificationSubscription:
  Type: AWS::SNS::Subscription
  Properties:
    TopicArn: !Ref OrderEventsTopic
    Protocol: sqs
    Endpoint: !GetAtt NotificationQueue.Arn
    FilterPolicy:
      event_type: ["order_placed", "order_shipped"]
    FilterPolicyScope: MessageBody
    RawMessageDelivery: true

AnalyticsSubscription:
  Type: AWS::SNS::Subscription
  Properties:
    TopicArn: !Ref OrderEventsTopic
    Protocol: firehose
    Endpoint: !GetAtt AnalyticsDeliveryStream.Arn
    SubscriptionRoleArn: !GetAtt SNSFirehoseRole.Arn
```

---

## ECS Fargate

### Task Definition

```json
{
  "family": "my-api",
  "networkMode": "awsvpc",
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "1024",
  "memory": "2048",
  "runtimePlatform": {
    "cpuArchitecture": "ARM64",
    "operatingSystemFamily": "LINUX"
  },
  "executionRoleArn": "arn:aws:iam::123456789:role/ecsTaskExecutionRole",
  "taskRoleArn": "arn:aws:iam::123456789:role/myApiTaskRole",
  "containerDefinitions": [
    {
      "name": "app",
      "image": "123456789.dkr.ecr.us-east-1.amazonaws.com/my-api:latest",
      "portMappings": [
        {"containerPort": 8080, "protocol": "tcp"}
      ],
      "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "/ecs/my-api",
          "awslogs-region": "us-east-1",
          "awslogs-stream-prefix": "app"
        }
      },
      "environment": [
        {"name": "NODE_ENV", "value": "production"},
        {"name": "PORT", "value": "8080"}
      ],
      "secrets": [
        {
          "name": "DATABASE_URL",
          "valueFrom": "arn:aws:secretsmanager:us-east-1:123456789:secret:prod/db-url"
        }
      ],
      "healthCheck": {
        "command": ["CMD-SHELL", "curl -f http://localhost:8080/health || exit 1"],
        "interval": 30,
        "timeout": 5,
        "retries": 3,
        "startPeriod": 60
      },
      "essential": true,
      "readonlyRootFilesystem": true,
      "linuxParameters": {
        "initProcessEnabled": true
      }
    }
  ]
}
```

### ECS Service with Blue/Green Deployment

```yaml
ECSService:
  Type: AWS::ECS::Service
  Properties:
    ServiceName: my-api
    Cluster: !Ref ECSCluster
    TaskDefinition: !Ref TaskDefinition
    DesiredCount: 3
    LaunchType: FARGATE
    PlatformVersion: LATEST
    DeploymentController:
      Type: CODE_DEPLOY  # Blue/Green
    NetworkConfiguration:
      AwsvpcConfiguration:
        AssignPublicIp: DISABLED
        SecurityGroups:
          - !Ref AppSecurityGroup
        Subnets:
          - !Ref PrivateSubnet1
          - !Ref PrivateSubnet2
          - !Ref PrivateSubnet3
    LoadBalancers:
      - ContainerName: app
        ContainerPort: 8080
        TargetGroupArn: !Ref BlueTargetGroup
    EnableExecuteCommand: true  # For ECS Exec debugging
    ServiceConnectConfiguration:
      Enabled: true
      Namespace: prod
      Services:
        - PortName: http
          DiscoveryName: my-api
          ClientAliases:
            - Port: 8080
```

### ECS Exec for Debugging

```bash
# Execute a command in a running container
aws ecs execute-command \
  --cluster prod-cluster \
  --task arn:aws:ecs:us-east-1:123456789:task/prod-cluster/abc123 \
  --container app \
  --interactive \
  --command "/bin/sh"
```

---

## CloudFront

### Distribution Configuration

```yaml
Distribution:
  Type: AWS::CloudFront::Distribution
  Properties:
    DistributionConfig:
      Origins:
        - Id: S3Origin
          DomainName: !GetAtt WebBucket.RegionalDomainName
          OriginAccessControlId: !GetAtt OAC.Id
          S3OriginConfig:
            OriginAccessIdentity: ""
        - Id: ApiOrigin
          DomainName: !Sub "${HttpApi}.execute-api.${AWS::Region}.amazonaws.com"
          CustomOriginConfig:
            HTTPSPort: 443
            OriginProtocolPolicy: https-only
            OriginSSLProtocols: [TLSv1.2]
      DefaultCacheBehavior:
        TargetOriginId: S3Origin
        ViewerProtocolPolicy: redirect-to-https
        CachePolicyId: 658327ea-f89d-4fab-a63d-7e88639e58f6  # CachingOptimized
        ResponseHeadersPolicyId: 67f7725c-6f97-4210-82d7-5512b31e9d03  # SecurityHeaders
        Compress: true
      CacheBehaviors:
        - PathPattern: /api/*
          TargetOriginId: ApiOrigin
          ViewerProtocolPolicy: https-only
          CachePolicyId: 4135ea2d-6df8-44a3-9df3-4b5a84be39ad  # CachingDisabled
          OriginRequestPolicyId: b689b0a8-53d0-40ab-baf2-68738e2966ac  # AllViewerExceptHostHeader
          AllowedMethods: [GET, HEAD, OPTIONS, PUT, POST, PATCH, DELETE]
      DefaultRootObject: index.html
      CustomErrorResponses:
        - ErrorCode: 403
          ResponseCode: 200
          ResponsePagePath: /index.html  # SPA routing
        - ErrorCode: 404
          ResponseCode: 200
          ResponsePagePath: /index.html
      HttpVersion: http2and3
      PriceClass: PriceClass_100  # US, Canada, Europe only
      ViewerCertificate:
        AcmCertificateArn: !Ref Certificate
        SslSupportMethod: sni-only
        MinimumProtocolVersion: TLSv1.2_2021
      WebACLId: !GetAtt WAF.Arn
```

### CloudFront Functions for Edge Logic

```javascript
// URL rewriting for SPA + clean URLs
function handler(event) {
  var request = event.request;
  var uri = request.uri;

  // Check if URI has a file extension
  if (uri.includes('.')) {
    return request; // Serve static file as-is
  }

  // Check if URI ends with /
  if (uri.endsWith('/')) {
    request.uri += 'index.html';
  } else {
    // Rewrite to index.html for SPA routing
    request.uri = '/index.html';
  }

  return request;
}
```

---

## Service Comparison Quick Reference

### Message Queue: SQS vs SNS vs EventBridge vs Kinesis

```
Feature         │ SQS              │ SNS              │ EventBridge       │ Kinesis
────────────────┼──────────────────┼──────────────────┼───────────────────┼─────────────
Pattern         │ Queue (pull)     │ Pub/sub (push)   │ Event bus (push)  │ Stream (pull)
Ordering        │ FIFO available   │ FIFO available   │ No                │ Per-shard
Retention       │ 14 days          │ None (instant)   │ Archive (replay)  │ 365 days
Consumers       │ 1 (per message)  │ Many (fan-out)   │ Many (rules)      │ Many (shards)
Throughput      │ Unlimited        │ Unlimited        │ Unlimited         │ 1 MB/sec/shard
Filtering       │ No               │ Yes (attributes) │ Yes (rich rules)  │ No
Dead letter     │ Yes              │ Yes              │ Yes               │ No
Best for        │ Work queues,     │ Fan-out,         │ Event routing,    │ Real-time
                │ decoupling       │ notifications    │ SaaS integration  │ analytics
```

### Database Selection

```
Workload                 │ Service              │ Why
─────────────────────────┼──────────────────────┼──────────────────────────
Relational, complex SQL  │ Aurora PostgreSQL     │ Full SQL, high availability
Key-value, <10ms latency │ DynamoDB              │ Unlimited scale, serverless
Document store           │ DocumentDB            │ MongoDB-compatible
Graph relationships      │ Neptune               │ Native graph queries
In-memory cache          │ ElastiCache Redis     │ Sub-millisecond latency
Time-series (IoT/metrics)│ Timestream            │ Built-in time functions
Ledger/audit trail       │ QLDB                  │ Immutable, cryptographic
Wide column (Cassandra)  │ Keyspaces             │ Managed Cassandra
Search                   │ OpenSearch             │ Full-text, analytics
Data warehouse           │ Redshift               │ Columnar, petabyte-scale
```
