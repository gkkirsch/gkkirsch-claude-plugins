# AWS Services Deep Reference Guide

Comprehensive reference for core AWS services — architecture patterns, configuration examples, pricing models, and when to use each service.

---

## Compute Services

### EC2 (Elastic Compute Cloud)

**Instance Families:**

| Family | Prefix | Use Case | Example |
|--------|--------|----------|---------|
| General Purpose | m7g, m7i, m6i | Web servers, app servers | m7g.xlarge |
| Compute Optimized | c7g, c7i, c6i | Batch, gaming, ML inference | c7g.2xlarge |
| Memory Optimized | r7g, r7i, r6i | Databases, caches | r7g.4xlarge |
| Storage Optimized | i4g, i3, d3 | Data warehousing, HDFS | i4g.2xlarge |
| Accelerated | p5, g5, inf2 | ML training, graphics | p5.48xlarge |
| High Performance | hpc7g, hpc6a | Scientific computing | hpc7g.16xlarge |

**Instance Naming Convention:**

```
m7g.2xlarge
│││  └───── Size (nano → metal)
││└──────── Generation (7th)
│└───────── Processor/feature (g=Graviton, i=Intel, a=AMD)
└────────── Family (m=general purpose)
```

**Size Progression:**

```
nano → micro → small → medium → large → xlarge → 2xl → 4xl → 8xl → 12xl → 16xl → 24xl → metal
```

**Pricing Models:**

| Model | Discount | Commitment | Best For |
|-------|----------|------------|----------|
| On-Demand | 0% | None | Testing, unpredictable |
| Spot | 60-90% | None (can interrupt) | Batch, fault-tolerant |
| Reserved (1yr) | ~30-40% | 1 year | Steady-state |
| Reserved (3yr) | ~50-60% | 3 years | Long-running known |
| Savings Plan | ~25-66% | $/hr commitment | Flexible |

### Lambda

**Configuration Limits:**

| Parameter | Limit |
|-----------|-------|
| Memory | 128 MB - 10,240 MB |
| Timeout | 1s - 900s (15 min) |
| Deployment package | 50 MB (zip), 250 MB (unzipped) |
| Container image | 10 GB |
| /tmp storage | 512 MB - 10,240 MB |
| Concurrent executions | 1,000 (default, can increase) |
| Burst concurrency | 500-3000 (region-dependent) |
| Environment variables | 4 KB total |
| Layers | 5 per function |

**Supported Runtimes:**

```
Node.js:    20.x, 18.x
Python:     3.12, 3.11, 3.10, 3.9
Java:       21, 17, 11
.NET:       8, 6
Ruby:       3.3, 3.2
Go:         Custom runtime (al2023)
Rust:       Custom runtime (al2023)
Custom:     Any language via container image or custom runtime
```

**Invocation Patterns:**

```
Synchronous (RequestResponse):
  API Gateway → Lambda → Response
  SDK invoke → Lambda → Response

Asynchronous (Event):
  S3 event → Lambda (2 retries, DLQ)
  SNS → Lambda
  EventBridge → Lambda

Stream (Poll-based):
  Kinesis → Lambda (batch processing)
  DynamoDB Streams → Lambda
  SQS → Lambda (batch up to 10,000)
```

### ECS (Elastic Container Service)

**Launch Types:**

| Feature | EC2 | Fargate |
|---------|-----|---------|
| Server management | You manage | AWS manages |
| Pricing | Per instance | Per vCPU/memory/second |
| GPU support | Yes | No |
| Instance placement | Full control | AWS decides |
| Local storage | Instance store | 20-200 GB ephemeral |
| Max task size | Instance size | 16 vCPU, 120 GB memory |
| Networking | awsvpc, bridge, host | awsvpc only |

**Task Definition CPU/Memory Combinations (Fargate):**

```
CPU (vCPU)  │  Memory Options (GB)
────────────┼─────────────────────
0.25        │  0.5, 1, 2
0.5         │  1, 2, 3, 4
1           │  2, 3, 4, 5, 6, 7, 8
2           │  4-16 (1 GB increments)
4           │  8-30 (1 GB increments)
8           │  16-60 (4 GB increments)
16          │  32-120 (8 GB increments)
```

### EKS (Elastic Kubernetes Service)

**Pricing:**

```
EKS Control Plane:  $0.10/hour ($73/month)
EKS Anywhere:       Per node licensing
EKS on Fargate:     Fargate pricing (no control plane charge for pods)
EKS Auto Mode:      $0.10/hr control plane + node pricing
Worker Nodes:       EC2 pricing (on-demand, spot, RI)
```

**Add-ons (managed):**

```
Required:
  - CoreDNS (cluster DNS)
  - kube-proxy (networking)
  - VPC CNI (pod networking)

Recommended:
  - EBS CSI Driver (persistent volumes)
  - EFS CSI Driver (shared file systems)
  - Karpenter (node auto-scaling)
  - AWS Load Balancer Controller
  - Metrics Server
  - Cluster Autoscaler (legacy, prefer Karpenter)
```

---

## Database Services

### RDS (Relational Database Service)

**Supported Engines:**

| Engine | Versions | Max Storage | Multi-AZ |
|--------|----------|-------------|----------|
| PostgreSQL | 13-16 | 64 TB | Yes |
| MySQL | 8.0 | 64 TB | Yes |
| MariaDB | 10.5-10.11 | 64 TB | Yes |
| Oracle | 19c, 21c | 64 TB | Yes |
| SQL Server | 2019, 2022 | 16 TB | Yes |

**Instance Classes:**

```
db.t4g.*    — Burstable (dev/test)
db.m7g.*    — General purpose
db.r7g.*    — Memory optimized (production databases)
db.x2g.*    — Memory optimized (large in-memory workloads)
```

**Storage Types:**

| Type | IOPS | Throughput | Cost/GB | Use Case |
|------|------|-----------|---------|----------|
| gp3 | 3,000 (base) | 125 MB/s | $0.08 | General purpose |
| io2 | Up to 256K | 4,000 MB/s | $0.125 | High-performance |
| magnetic | Moderate | Moderate | $0.10 | Legacy (avoid) |

### Aurora

**Architecture:**

```
┌──────────────────────────────────────────────┐
│              Aurora Cluster                    │
│                                              │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐     │
│  │ Writer  │  │ Reader  │  │ Reader  │     │
│  │Instance │  │Instance │  │Instance │     │
│  └────┬────┘  └────┬────┘  └────┬────┘     │
│       │            │            │           │
│  ┌────▼────────────▼────────────▼────┐      │
│  │        Shared Storage (SSD)       │      │
│  │    6 copies across 3 AZs         │      │
│  │    Up to 128 TB auto-scaling     │      │
│  └───────────────────────────────────┘      │
└──────────────────────────────────────────────┘
```

**Aurora vs Standard RDS:**

| Feature | Aurora | Standard RDS |
|---------|--------|-------------|
| Storage | 6-way replication, auto-scaling | 2-way (Multi-AZ) |
| Failover | 30 seconds | 60-120 seconds |
| Read replicas | Up to 15 | Up to 5 (15 for MySQL) |
| Backtrack | Yes (rewind to point in time) | No |
| Serverless | v2 (auto-scaling ACUs) | No |
| Global database | Sub-second cross-region | Read replica (async) |
| Performance | 3-5x PostgreSQL, 5x MySQL | Standard |
| Cost | ~20% more than RDS | Baseline |

### DynamoDB

**Data Modeling — Single Table Design:**

```
┌─────────────────┬─────────────────┬──────────────────┐
│ PK              │ SK              │ Attributes       │
├─────────────────┼─────────────────┼──────────────────┤
│ USER#123        │ PROFILE         │ name, email      │
│ USER#123        │ ORDER#456       │ total, status    │
│ USER#123        │ ORDER#789       │ total, status    │
│ ORDER#456       │ ORDER#456       │ full order data  │
│ ORDER#456       │ ITEM#001        │ product, qty     │
│ ORDER#456       │ ITEM#002        │ product, qty     │
│ PRODUCT#ABC     │ PRODUCT#ABC     │ name, price      │
│ PRODUCT#ABC     │ REVIEW#001      │ rating, text     │
└─────────────────┴─────────────────┴──────────────────┘
```

**Access Patterns:**

```
Get user profile:        PK=USER#123, SK=PROFILE
Get user's orders:       PK=USER#123, SK begins_with("ORDER#")
Get order details:       PK=ORDER#456, SK=ORDER#456
Get order items:         PK=ORDER#456, SK begins_with("ITEM#")
Get product reviews:     PK=PRODUCT#ABC, SK begins_with("REVIEW#")
```

**Capacity Modes:**

| Mode | Pricing | Best For |
|------|---------|----------|
| On-Demand | $1.25/M reads, $6.25/M writes | Unpredictable traffic |
| Provisioned | $0.00065/RCU/hr, $0.00065/WCU/hr | Steady traffic |
| Reserved (1yr) | ~53% off provisioned | Known steady workloads |

**Limits:**

```
Item size:           400 KB max
Partition key:       2048 bytes max
Sort key:            1024 bytes max
Batch write:         25 items, 16 MB
Batch read:          100 items, 16 MB
Query/Scan:          1 MB per call (paginate)
GSIs per table:      20 (can increase)
LSIs per table:      5
Transaction items:   100 per transaction
```

### ElastiCache

**Redis vs Memcached:**

| Feature | Redis | Memcached |
|---------|-------|-----------|
| Data structures | Strings, lists, sets, sorted sets, hashes, streams | Strings only |
| Persistence | AOF, RDB snapshots | None |
| Replication | Multi-AZ with auto-failover | None |
| Clustering | Yes (up to 500 nodes) | Yes (up to 300 nodes) |
| Pub/Sub | Yes | No |
| Lua scripting | Yes | No |
| Transactions | Yes | No |
| Geospatial | Yes | No |
| Max item size | 512 MB | 50 MB (adjustable) |

---

## Storage Services

### S3 (Simple Storage Service)

**Storage Classes Quick Reference:**

| Class | Availability | Min Duration | Retrieval | Best For |
|-------|-------------|-------------|-----------|----------|
| Standard | 99.99% | None | Instant | Frequent access |
| Intelligent Tiering | 99.9% | None | Instant | Unknown patterns |
| Standard-IA | 99.9% | 30 days | Instant | Monthly access |
| One Zone-IA | 99.5% | 30 days | Instant | Reproducible |
| Glacier Instant | 99.9% | 90 days | Instant | Quarterly access |
| Glacier Flexible | 99.99% | 90 days | 1-12 hours | Annual archives |
| Glacier Deep | 99.99% | 180 days | 12-48 hours | Compliance |

**S3 Features:**

```
Versioning:           Track all object versions
Lifecycle:            Auto-transition between storage classes
Replication:          Cross-region (CRR) or same-region (SRR)
Event notifications:  Trigger Lambda, SQS, SNS, EventBridge
Object Lock:          WORM compliance (Governance or Compliance mode)
Access Points:        Named endpoints with custom policies
Batch Operations:     Bulk operations on billions of objects
S3 Select:            SQL queries on CSV/JSON/Parquet objects
Transfer Acceleration: Fast long-distance uploads via CloudFront edges
```

**S3 Limits:**

```
Object size:          5 TB max (5 GB per PUT, use multipart for larger)
Bucket count:         100 per account (can increase)
Object count:         Unlimited per bucket
Request rate:         5,500 GET/s and 3,500 PUT/s per prefix
Metadata:             2 KB per object (user-defined)
Tags:                 10 per object
```

### EBS (Elastic Block Store)

**Volume Types:**

| Type | IOPS (Max) | Throughput | Size | Use Case |
|------|-----------|------------|------|----------|
| gp3 | 16,000 | 1,000 MB/s | 1 GB-16 TB | General purpose |
| gp2 | 16,000 | 250 MB/s | 1 GB-16 TB | Legacy (migrate to gp3) |
| io2 Block Express | 256,000 | 4,000 MB/s | 4 GB-64 TB | Critical databases |
| io2 | 64,000 | 1,000 MB/s | 4 GB-16 TB | High-performance |
| st1 | 500 | 500 MB/s | 125 GB-16 TB | Big data throughput |
| sc1 | 250 | 250 MB/s | 125 GB-16 TB | Cold data |

### EFS (Elastic File System)

```
Performance Modes:
  General Purpose:    Lowest latency, most workloads
  Max I/O:            Higher latency, highest throughput (legacy)
  Elastic:            Auto-scales throughput (recommended)

Throughput Modes:
  Elastic:            Automatic, up to 10 GB/s reads, 3 GB/s writes
  Provisioned:        Fixed throughput, 1-3 GB/s
  Bursting:           Scales with storage size (legacy)

Storage Classes:
  Standard:           $0.30/GB/month (frequent access)
  Infrequent Access:  $0.016/GB/month + $0.01/GB read
  Archive:            $0.008/GB/month + $0.03/GB read (90-day min)
```

---

## Networking Services

### VPC (Virtual Private Cloud)

**CIDR Planning:**

```
Production VPC:      10.0.0.0/16   (65,534 IPs)
  Public subnets:    10.0.0.0/20   (4,094 per subnet × 3 AZs)
  Private subnets:   10.0.16.0/20  (4,094 per subnet × 3 AZs)
  Database subnets:  10.0.32.0/20  (4,094 per subnet × 3 AZs)
  Reserved:          10.0.48.0/20  (future use)

Staging VPC:         10.1.0.0/16
Development VPC:     10.2.0.0/16
Shared Services VPC: 10.3.0.0/16
```

**Security Group vs NACL:**

| Feature | Security Group | NACL |
|---------|---------------|------|
| Level | Instance/ENI | Subnet |
| Rules | Allow only | Allow and Deny |
| State | Stateful | Stateless |
| Evaluation | All rules | Rules in order |
| Default | Deny all inbound | Allow all |

### Route 53

**Routing Policies:**

| Policy | Use Case |
|--------|----------|
| Simple | Single resource |
| Weighted | A/B testing, gradual migration |
| Latency | Multi-region, route to closest |
| Failover | Primary/secondary DR |
| Geolocation | Compliance, content localization |
| Geoproximity | Route by geographic area with bias |
| Multivalue Answer | Simple load balancing across multiple IPs |
| IP-based | Route based on source IP ranges |

### CloudFront

**Cache Behaviors:**

```
Path Pattern          Cache Policy              Origin
/api/*               CachingDisabled            ALB
/static/*            CachingOptimized           S3
/images/*            CachingOptimized           S3
/ws/*                CachingDisabled            ALB (WebSocket)
Default (*)          Custom (TTL 60s)           ALB
```

**Edge Functions:**

| Feature | CloudFront Functions | Lambda@Edge |
|---------|---------------------|-------------|
| Runtime | JavaScript | Node.js, Python |
| Execution | Viewer events only | All events |
| Duration | < 1 ms | Up to 30s |
| Memory | 2 MB | 128-10,240 MB |
| Network | No | Yes |
| Body access | No | Yes (origin events) |
| Pricing | $0.10/M requests | $0.60/M + duration |

---

## Messaging and Integration

### SQS (Simple Queue Service)

**Standard vs FIFO:**

| Feature | Standard | FIFO |
|---------|----------|------|
| Throughput | Unlimited | 300-30,000 msg/s |
| Ordering | Best-effort | Guaranteed |
| Deduplication | At-least-once | Exactly-once |
| Cost | $0.40/M requests | $0.50/M requests |

**Configuration:**

```
Visibility Timeout:   30s default (0s - 12h)
Message retention:    4 days default (1 min - 14 days)
Max message size:     256 KB
Long polling:         Up to 20 seconds (reduces costs)
Dead letter queue:    Configure maxReceiveCount for failed messages
Delay queue:          0-900 seconds delay before visible
```

### SNS (Simple Notification Service)

**Subscription Protocols:**

```
HTTP/HTTPS:      Webhook endpoints
Email:           Notifications
SMS:             Text messages
SQS:             Fan-out to queues
Lambda:          Serverless processing
Kinesis Firehose: Streaming delivery
Application:     Mobile push
```

**Fan-Out Pattern:**

```
Producer → SNS Topic → SQS Queue 1 → Lambda (Process)
                     → SQS Queue 2 → ECS (Analyze)
                     → SQS Queue 3 → Lambda (Archive)
                     → Lambda (Real-time alert)
```

### EventBridge

**Event Pattern Matching:**

```json
{
  "source": ["myapp.orders"],
  "detail-type": ["OrderCreated", "OrderUpdated"],
  "detail": {
    "amount": [{"numeric": [">=", 100]}],
    "status": [{"anything-but": "cancelled"}],
    "region": [{"prefix": "us-"}]
  }
}
```

**Pipe Pattern (EventBridge Pipes):**

```
Source → Filter → Enrichment → Target

DynamoDB Stream → Filter: INSERT events → Lambda: enrich → SQS
Kinesis Stream → Filter: error events → Step Functions: process
SQS Queue → No filter → API Gateway: forward
```

### Step Functions

**State Types:**

| State | Purpose |
|-------|---------|
| Task | Execute work (Lambda, ECS, API call) |
| Choice | Branching logic |
| Parallel | Run branches concurrently |
| Map | Process items in array |
| Wait | Delay execution |
| Pass | Pass input to output |
| Succeed | End successfully |
| Fail | End with error |

**Pricing:**

```
Standard Workflow:
  $0.025 per 1,000 state transitions
  Best for: Long-running, auditable workflows

Express Workflow:
  $0.00001667/GB-second + $0.000001/request
  Best for: High-volume, short-duration (< 5 min)
```

---

## Security Services

### IAM (Identity and Access Management)

**Policy Evaluation Logic:**

```
1. If any explicit DENY → DENY
2. If any SCP DENY → DENY
3. If any permission boundary DENY → DENY
4. If explicit ALLOW from identity + resource policy → ALLOW
5. Default → DENY (implicit deny)
```

**Best Practices:**

```
1. Never use root account for daily operations
2. Enable MFA on all human users
3. Use IAM Identity Center (SSO) for human access
4. Use IAM roles for services, never long-term keys
5. Follow least privilege — start with no permissions, add as needed
6. Use AWS managed policies as starting point, then customize
7. Set permission boundaries on all IAM entities
8. Enable CloudTrail for audit logging
9. Review IAM Access Analyzer findings regularly
10. Rotate credentials automatically (Secrets Manager)
```

### KMS (Key Management Service)

**Key Types:**

| Type | Management | Cost | Use Case |
|------|-----------|------|----------|
| AWS managed | AWS | Free | S3-SSE, EBS default |
| Customer managed | You | $1/month + API | Custom encryption policies |
| Customer provided | You | Free | Bring your own key (BYOK) |

### WAF (Web Application Firewall)

**Managed Rule Groups:**

```
AWS Managed Rules:
  AWSManagedRulesCommonRuleSet         — OWASP Top 10 basics
  AWSManagedRulesKnownBadInputsRuleSet — Log4j, path traversal
  AWSManagedRulesSQLiRuleSet           — SQL injection
  AWSManagedRulesLinuxRuleSet          — LFI on Linux
  AWSManagedRulesAmazonIpReputationList — Known malicious IPs
  AWSManagedRulesBotControlRuleSet     — Bot detection
  AWSManagedRulesAnonymousIpList       — VPN, proxy, Tor
```

---

## Monitoring and Observability

### CloudWatch

**Key Metrics by Service:**

```
EC2:
  CPUUtilization, NetworkIn/Out, DiskReadOps, StatusCheckFailed

RDS:
  CPUUtilization, FreeableMemory, DatabaseConnections,
  ReadIOPS, WriteIOPS, ReplicaLag

ALB:
  RequestCount, TargetResponseTime, HTTPCode_Target_5XX_Count,
  HealthyHostCount, UnHealthyHostCount

Lambda:
  Invocations, Duration, Errors, Throttles, ConcurrentExecutions

ECS:
  CPUUtilization, MemoryUtilization, RunningTaskCount

DynamoDB:
  ConsumedReadCapacityUnits, ConsumedWriteCapacityUnits,
  ThrottledRequests, SystemErrors

SQS:
  ApproximateNumberOfMessagesVisible, ApproximateAgeOfOldestMessage,
  NumberOfMessagesSent, NumberOfMessagesReceived
```

**CloudWatch Alarms — Recommended:**

```yaml
Critical:
  - EC2 StatusCheckFailed > 0 for 2 minutes
  - RDS CPU > 90% for 10 minutes
  - ALB 5XX > 5% of total for 5 minutes
  - Lambda Errors > 10% of invocations
  - DynamoDB ThrottledRequests > 0
  - ECS RunningTaskCount < desired count

Warning:
  - EC2 CPU > 80% for 15 minutes
  - RDS FreeableMemory < 1 GB
  - ALB TargetResponseTime > 2 seconds (p99)
  - Lambda Duration > 80% of timeout
  - SQS ApproximateAgeOfOldestMessage > 300 seconds
  - EBS VolumeQueueLength > 10
```

### CloudTrail

**Must-Monitor Events:**

```
Security:
  ConsoleLogin (especially without MFA)
  CreateUser, CreateAccessKey, AttachUserPolicy
  StopLogging (CloudTrail disabled!)
  DeleteTrail, PutEventSelectors
  CreateNetworkAclEntry, AuthorizeSecurityGroupIngress

Cost:
  RunInstances (large instance types)
  CreateDBInstance, ModifyDBInstance
  PurchaseReservedInstancesOffering

Data:
  PutBucketPolicy, PutBucketPublicAccessBlock
  DeleteBucket, DeleteObject
  CreateSnapshot, ShareSnapshot
```

---

## Cost Reference

### Free Tier (Always Free)

```
Lambda:         1M requests/month, 400,000 GB-seconds
DynamoDB:       25 GB storage, 25 WCU, 25 RCU
S3:             N/A (12 months free tier only)
CloudWatch:     10 custom metrics, 10 alarms, 1M API requests
SNS:            1M publishes, 100K HTTP deliveries
SQS:            1M requests
API Gateway:    1M REST API calls (12 months)
CloudFront:     1 TB data transfer (12 months)
```

### Common Monthly Costs (us-east-1, approximate)

```
EC2 m7g.large (on-demand):         $59.57/month
EC2 t3.micro (on-demand):          $7.49/month
RDS db.r7g.large (Multi-AZ):       $342.72/month
RDS db.t4g.micro (Single-AZ):      $12.41/month
Aurora Serverless v2 (1 ACU idle):  $43.20/month
ElastiCache r7g.large:             $139.68/month
ALB (idle):                        $16.43/month
NAT Gateway (idle):                $32.40/month
EKS Control Plane:                 $73.00/month
S3 (100 GB Standard):              $2.30/month
CloudFront (100 GB transfer):      $8.50/month
Route 53 hosted zone:              $0.50/month
KMS customer key:                  $1.00/month
Secrets Manager secret:            $0.40/month
```

---

## Service Limits Quick Reference

```
VPCs per region:                5 (adjustable to 100)
Subnets per VPC:                200
Security groups per VPC:        2,500
Rules per security group:       60 inbound + 60 outbound
Elastic IPs per region:         5 (adjustable)
Internet Gateways per region:   5 (1 per VPC)
NAT Gateways per AZ:            5
VPC Peering per VPC:            50 (adjustable to 125)

EC2 instances per region:       Varies by type (check quotas)
EBS volumes per region:         5,000
EBS snapshots per region:       100,000
AMIs per region:                50,000

S3 buckets per account:         100 (adjustable to 1,000)
Lambda concurrent executions:   1,000 (adjustable)
API Gateway REST APIs:          600 per account
CloudFormation stacks:          2,000 per region
IAM roles per account:          1,000 (adjustable to 5,000)
IAM policies per account:       1,500
```

---

## AWS CLI Quick Reference

```bash
# Identity
aws sts get-caller-identity

# EC2
aws ec2 describe-instances --filters Name=instance-state-name,Values=running
aws ec2 describe-instances --query 'Reservations[].Instances[].{ID:InstanceId,Type:InstanceType,State:State.Name}' --output table

# S3
aws s3 ls s3://bucket-name/
aws s3 cp file.txt s3://bucket/
aws s3 sync ./dir s3://bucket/prefix/

# RDS
aws rds describe-db-instances --query 'DBInstances[].{ID:DBInstanceIdentifier,Engine:Engine,Class:DBInstanceClass,Status:DBInstanceStatus}' --output table

# ECS
aws ecs list-clusters
aws ecs list-services --cluster production
aws ecs describe-services --cluster production --services api

# Lambda
aws lambda list-functions --query 'Functions[].{Name:FunctionName,Runtime:Runtime,Memory:MemorySize}' --output table
aws lambda invoke --function-name myfunction --payload '{"key":"value"}' output.json

# CloudWatch
aws cloudwatch get-metric-statistics --namespace AWS/EC2 --metric-name CPUUtilization --dimensions Name=InstanceId,Value=i-xxx --start-time 2024-01-01T00:00:00Z --end-time 2024-01-02T00:00:00Z --period 3600 --statistics Average

# Cost Explorer
aws ce get-cost-and-usage --time-period Start=2024-01-01,End=2024-02-01 --granularity MONTHLY --metrics BlendedCost --group-by Type=DIMENSION,Key=SERVICE
```
