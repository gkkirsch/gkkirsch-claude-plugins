# AWS Cost Optimization Reference

Practical strategies for reducing AWS costs without sacrificing reliability. Covers commitment-based discounts, right-sizing, architecture optimization, and FinOps practices.

---

## Cost Visibility Foundation

### Enable Cost and Usage Report (CUR)

The single most important cost tool — provides line-item billing data:

```bash
# Create CUR 2.0 report
aws bcm-data-exports create-export \
  --export '{
    "Name": "daily-cost-report",
    "DataQuery": {
      "QueryStatement": "SELECT * FROM COST_AND_USAGE_REPORT",
      "TableConfigurations": {
        "COST_AND_USAGE_REPORT": {
          "TIME_GRANULARITY": "DAILY",
          "INCLUDE_RESOURCES": "TRUE",
          "INCLUDE_MANUAL_DISCOUNT_COMPATIBILITY": "FALSE",
          "INCLUDE_SPLIT_COST_ALLOCATION_DATA": "TRUE"
        }
      }
    },
    "DestinationConfigurations": {
      "S3Destination": {
        "S3Bucket": "company-cost-reports",
        "S3Prefix": "cur/",
        "S3Region": "us-east-1",
        "S3OutputConfigurations": {
          "OutputType": "CUSTOM",
          "Format": "PARQUET",
          "Compression": "PARQUET",
          "Overwrite": "OVERWRITE_REPORT"
        }
      }
    },
    "RefreshCadence": {
      "Frequency": "SYNCHRONOUS"
    }
  }'
```

### Cost Allocation Tags

Tags are the foundation of cost accountability. Without them, you can't attribute costs.

```bash
# Activate cost allocation tags
aws ce update-cost-allocation-tags-status \
  --cost-allocation-tags-status '[
    {"TagKey": "Environment", "Status": "Active"},
    {"TagKey": "CostCenter", "Status": "Active"},
    {"TagKey": "Owner", "Status": "Active"},
    {"TagKey": "Application", "Status": "Active"},
    {"TagKey": "Project", "Status": "Active"}
  ]'
```

### AWS Cost Anomaly Detection

```bash
# Create a cost anomaly monitor for each service
aws ce create-anomaly-monitor \
  --anomaly-monitor '{
    "MonitorName": "service-cost-monitor",
    "MonitorType": "DIMENSIONAL",
    "MonitorDimension": "SERVICE"
  }'

# Create alert subscription
aws ce create-anomaly-subscription \
  --anomaly-subscription '{
    "SubscriptionName": "cost-anomaly-alerts",
    "MonitorArnList": ["arn:aws:ce::123456789:anomalymonitor/monitor-id"],
    "Subscribers": [
      {
        "Address": "finops-team@company.com",
        "Type": "EMAIL"
      },
      {
        "Address": "arn:aws:sns:us-east-1:123456789:cost-alerts",
        "Type": "SNS"
      }
    ],
    "Threshold": 10,
    "ThresholdExpression": {
      "Dimensions": {
        "Key": "ANOMALY_TOTAL_IMPACT_PERCENTAGE",
        "Values": ["10"],
        "MatchOptions": ["GREATER_THAN_OR_EQUAL"]
      }
    },
    "Frequency": "IMMEDIATE"
  }'
```

### AWS Budgets

```bash
# Monthly budget with forecasted alerts
aws budgets create-budget \
  --account-id 123456789012 \
  --budget '{
    "BudgetName": "monthly-total",
    "BudgetLimit": {"Amount": "10000", "Unit": "USD"},
    "TimeUnit": "MONTHLY",
    "BudgetType": "COST",
    "CostFilters": {},
    "CostTypes": {
      "IncludeCredit": false,
      "IncludeRefund": false,
      "IncludeSupport": true,
      "IncludeTax": true
    }
  }' \
  --notifications-with-subscribers '[
    {
      "Notification": {
        "NotificationType": "ACTUAL",
        "ComparisonOperator": "GREATER_THAN",
        "Threshold": 80,
        "ThresholdType": "PERCENTAGE"
      },
      "Subscribers": [
        {"SubscriptionType": "EMAIL", "Address": "finops@company.com"}
      ]
    },
    {
      "Notification": {
        "NotificationType": "FORECASTED",
        "ComparisonOperator": "GREATER_THAN",
        "Threshold": 100,
        "ThresholdType": "PERCENTAGE"
      },
      "Subscribers": [
        {"SubscriptionType": "EMAIL", "Address": "finops@company.com"},
        {"SubscriptionType": "SNS", "Address": "arn:aws:sns:us-east-1:123456789:budget-alerts"}
      ]
    }
  ]'

# Per-account budget (in Organizations)
aws budgets create-budget \
  --account-id 123456789012 \
  --budget '{
    "BudgetName": "dev-account-budget",
    "BudgetLimit": {"Amount": "500", "Unit": "USD"},
    "TimeUnit": "MONTHLY",
    "BudgetType": "COST",
    "CostFilters": {
      "LinkedAccount": ["444444444444"]
    }
  }' \
  --notifications-with-subscribers '[
    {
      "Notification": {
        "NotificationType": "ACTUAL",
        "ComparisonOperator": "GREATER_THAN",
        "Threshold": 90,
        "ThresholdType": "PERCENTAGE"
      },
      "Subscribers": [
        {"SubscriptionType": "EMAIL", "Address": "dev-lead@company.com"}
      ]
    }
  ]'
```

---

## Commitment-Based Discounts

### Savings Plans

```
Type                   │ Discount │ Flexibility            │ Term     │ Best For
───────────────────────┼──────────┼────────────────────────┼──────────┼──────────────
Compute Savings Plans  │ ~30-40%  │ Any instance, Lambda,  │ 1 or 3yr │ Mixed workloads
                       │          │ Fargate, across regions│          │
EC2 Instance SP        │ ~40-50%  │ Specific family+region │ 1 or 3yr │ Stable EC2 fleets
SageMaker SP           │ ~30-40%  │ Any SageMaker instance │ 1 or 3yr │ ML workloads
```

**Sizing Strategy:**
```
1. Analyze last 3 months of usage in Cost Explorer
2. Look at the Savings Plans recommendation page
3. Start with 70-80% of your steady-state compute spend
   (leave headroom for flexibility)
4. Use hourly commitment (not upfront unless cash flow allows)
5. Review and adjust quarterly

Example:
  Current compute spend: $10,000/month
  Steady-state baseline: ~$7,000/month (70%)
  Commitment: $7,000 × 0.7 (30% discount) = $4,900/month
  Savings: ~$2,100/month on the committed portion
```

```bash
# Get Savings Plans recommendations
aws savingsplans describe-savings-plans-offering-rates \
  --savings-plan-offering-ids offering-id \
  --filters '[
    {"name": "region", "values": ["us-east-1"]},
    {"name": "instanceFamily", "values": ["m7g"]}
  ]'

# Purchase Savings Plan
aws savingsplans create-savings-plan \
  --savings-plan-offering-id offering-id \
  --commitment "10.0" \
  --savings-plan-type "ComputeSavingsPlans" \
  --term-in-years 1 \
  --payment-option "NoUpfront"
```

### Reserved Instances (RDS, ElastiCache, OpenSearch, Redshift)

Savings Plans don't cover all services. Use RIs for:

```
Service          │ RI Discount │ Notes
─────────────────┼─────────────┼────────────────────────────────
RDS              │ 30-60%      │ Size-flexible within same family
ElastiCache      │ 30-55%      │ Must match node type exactly
OpenSearch       │ 30-50%      │ Must match instance type
Redshift         │ 30-50%      │ Must match node type
DynamoDB         │ ~25%        │ Reserved capacity for provisioned
```

```bash
# Purchase RDS Reserved Instance
aws rds purchase-reserved-db-instances-offering \
  --reserved-db-instances-offering-id offering-id \
  --db-instance-count 2 \
  --reserved-db-instance-id "prod-aurora-ri"
```

---

## Compute Optimization

### EC2 Right-Sizing

```bash
# Get Compute Optimizer recommendations
aws compute-optimizer get-ec2-instance-recommendations \
  --instance-arns arn:aws:ec2:us-east-1:123456789:instance/i-1234567890abcdef0 \
  --recommendation-preferences '{
    "cpuVendorArchitectures": ["AWS_ARM64", "CURRENT"]
  }'

# Export recommendations for all instances
aws compute-optimizer export-ec2-instance-recommendations \
  --s3-destination-config '{
    "bucket": "optimizer-reports",
    "keyPrefix": "ec2-recommendations"
  }' \
  --file-format Csv \
  --include-member-accounts
```

### Spot Instances

**Up to 90% discount for fault-tolerant workloads.**

```
Good for Spot               │ Bad for Spot
────────────────────────────┼────────────────────────────
Batch processing            │ Databases
CI/CD build agents          │ Stateful applications
Data analysis (EMR, Spark)  │ Single-instance workloads
Dev/test environments       │ Long-running critical tasks
Container tasks (some %)    │ Real-time user-facing APIs
Auto Scaling Groups (mixed) │
```

**EC2 Spot with diversification:**
```bash
aws ec2 request-spot-fleet \
  --spot-fleet-request-config '{
    "AllocationStrategy": "priceCapacityOptimized",
    "TargetCapacity": 10,
    "TerminateInstancesWithExpiration": true,
    "Type": "maintain",
    "LaunchTemplateConfigs": [
      {
        "LaunchTemplateSpecification": {
          "LaunchTemplateId": "lt-123",
          "Version": "$Latest"
        },
        "Overrides": [
          {"InstanceType": "m7g.large", "SubnetId": "subnet-1a"},
          {"InstanceType": "m6g.large", "SubnetId": "subnet-1a"},
          {"InstanceType": "m7i-flex.large", "SubnetId": "subnet-1a"},
          {"InstanceType": "m7g.large", "SubnetId": "subnet-1b"},
          {"InstanceType": "m6g.large", "SubnetId": "subnet-1b"},
          {"InstanceType": "m7i-flex.large", "SubnetId": "subnet-1b"}
        ]
      }
    ]
  }'
```

### ECS Fargate Spot

```yaml
ECSService:
  Type: AWS::ECS::Service
  Properties:
    CapacityProviderStrategy:
      - CapacityProvider: FARGATE
        Weight: 1
        Base: 2          # Always have 2 on-demand tasks
      - CapacityProvider: FARGATE_SPOT
        Weight: 3         # 75% of scaling is Spot
```

### Lambda Cost Optimization

```bash
# Use AWS Lambda Power Tuning to find optimal memory
# Deploy: https://github.com/alexcasalboni/aws-lambda-power-tuning

# Run the tuning state machine
aws stepfunctions start-execution \
  --state-machine-arn arn:aws:states:us-east-1:123456789:stateMachine:powerTuning \
  --input '{
    "lambdaARN": "arn:aws:lambda:us-east-1:123456789:function:my-function",
    "powerValues": [128, 256, 512, 1024, 1769, 3008],
    "num": 50,
    "payload": {"key": "value"},
    "strategy": "cost"
  }'
```

**Lambda cost rules of thumb:**
- ARM64 is 20% cheaper than x86_64 — always use it
- More memory can be cheaper if it reduces duration (CPU scales with memory)
- Batch SQS messages: process 10 per invocation instead of 1
- Use event filtering to avoid invoking Lambda for events you'll discard
- Use Provisioned Concurrency Auto Scaling instead of always-on provisioned

---

## Storage Optimization

### S3 Cost Reduction

```
Action                          │ Savings    │ Effort │ Impact
────────────────────────────────┼────────────┼────────┼──────────────
Enable Intelligent-Tiering     │ 20-40%     │ Low    │ High
Add lifecycle policies         │ 30-60%     │ Low    │ High
Delete incomplete uploads      │ Varies     │ Low    │ Medium
Compress objects before upload │ 50-80%     │ Medium │ High
Use S3 analytics               │ Variable   │ Low    │ Medium
Move logs to Glacier           │ 80-90%     │ Low    │ High
Enable S3 Batch Operations     │ Time saved │ Medium │ Medium
```

```bash
# Find and abort incomplete multipart uploads (they cost money!)
aws s3api list-multipart-uploads --bucket my-bucket \
  --query 'Uploads[?Initiated<`2026-03-01`].[Key,UploadId]' \
  --output text

# Add lifecycle rule to auto-abort
aws s3api put-bucket-lifecycle-configuration \
  --bucket my-bucket \
  --lifecycle-configuration '{
    "Rules": [{
      "ID": "abort-incomplete-uploads",
      "Status": "Enabled",
      "Filter": {},
      "AbortIncompleteMultipartUpload": {"DaysAfterInitiation": 7}
    }]
  }'

# S3 Storage Lens for organization-wide visibility
aws s3control put-storage-lens-configuration \
  --account-id 123456789012 \
  --config-id org-storage-lens \
  --storage-lens-configuration '{
    "Id": "org-storage-lens",
    "AccountLevel": {
      "BucketLevel": {
        "ActivityMetrics": {"IsEnabled": true},
        "PrefixLevel": {
          "StorageMetrics": {"IsEnabled": true, "SelectionCriteria": {"MaxDepth": 3, "MinStorageBytesPercentage": 1.0}}
        },
        "AdvancedCostOptimizationMetrics": {"IsEnabled": true},
        "AdvancedDataProtectionMetrics": {"IsEnabled": true},
        "DetailedStatusCodesMetrics": {"IsEnabled": true}
      }
    },
    "Include": {},
    "IsEnabled": true
  }'
```

### EBS Optimization

```bash
# Find unattached EBS volumes
aws ec2 describe-volumes \
  --filters Name=status,Values=available \
  --query 'Volumes[].{ID:VolumeId,Size:Size,Type:VolumeType,Created:CreateTime}' \
  --output table

# Find volumes with low IOPS utilization (candidates for gp3 downgrade)
# Use CloudWatch: VolumeReadOps + VolumeWriteOps < 3000 = gp3 baseline is enough

# Migrate gp2 to gp3 (saves 20% and gives better baseline performance)
aws ec2 modify-volume \
  --volume-id vol-123 \
  --volume-type gp3 \
  --iops 3000 \
  --throughput 125
```

### EBS Snapshot Cleanup

```bash
# Find snapshots older than 90 days
aws ec2 describe-snapshots --owner-ids self \
  --query 'Snapshots[?StartTime<`2025-12-19`].{ID:SnapshotId,Size:VolumeSize,Date:StartTime}' \
  --output table

# Delete old snapshots (be careful — check if they're used by AMIs first)
# aws ec2 delete-snapshot --snapshot-id snap-123
```

---

## Data Transfer Optimization

### Understanding Data Transfer Costs

```
Path                                    │ Cost/GB    │ Notes
────────────────────────────────────────┼────────────┼──────────────────
Internet → AWS                          │ Free       │ Ingress is free
AWS → Internet (first 100 GB/mo)        │ Free       │ Free tier
AWS → Internet (next 10 TB)             │ $0.09      │ Tiered pricing
Same AZ, private IP                     │ Free       │ Always use private IPs
Cross-AZ (within region)                │ $0.01/GB   │ Each direction
Cross-Region                            │ $0.02/GB   │ Each direction
VPC Peering (cross-AZ)                  │ $0.01/GB   │ Same as cross-AZ
Transit Gateway (cross-AZ)              │ $0.02/GB   │ Attachment + processing
NAT Gateway (processing)                │ $0.045/GB  │ Plus standard data transfer
VPC Endpoint (interface)                │ $0.01/GB   │ But no NAT Gateway cost
CloudFront → Internet                   │ $0.085/GB  │ Cheaper than direct
CloudFront → Origin (AWS)               │ Free       │ Origin fetch is free!
```

**Key optimization strategies:**
1. **Use CloudFront** — data from CloudFront to internet is cheaper than EC2/S3 direct
2. **Use VPC Endpoints** — avoid NAT Gateway processing charges ($0.045/GB)
3. **Keep traffic in-AZ** — cross-AZ traffic costs $0.01/GB each way
4. **Use S3 Gateway Endpoint** — free data transfer vs NAT Gateway
5. **Compress payloads** — reduces bytes transferred
6. **Use private IPs** — same-AZ traffic on private IPs is free

### CloudFront Cost Comparison

```
Scenario: 10 TB/month outbound to internet from us-east-1

Direct from EC2/S3:
  10,000 GB × $0.09 = $900/month

Via CloudFront:
  10,000 GB × $0.085 = $850/month
  + Origin fetch: Free (AWS origins)
  = $850/month (6% cheaper + lower latency + DDoS protection)

Via CloudFront with 50% cache hit rate:
  5,000 GB (origin fetch) × $0.00 = $0
  10,000 GB (CloudFront to internet) × $0.085 = $850
  = $850/month but origin only serves 5 TB
```

---

## Service-Specific Optimization

### DynamoDB Cost

```
Mode         │ Write Cost           │ Read Cost             │ When to Switch
─────────────┼──────────────────────┼───────────────────────┼──────────────────
On-Demand    │ $1.25/million WRU    │ $0.25/million RRU     │ < 18% utilization
Provisioned  │ $0.00065/WCU/hour    │ $0.00013/RCU/hour     │ > 18% utilization
Prov + RI    │ ~$0.0003/WCU/hour    │ ~$0.00006/RCU/hour    │ Steady, predictable
```

**Break-even: ~18% utilization of provisioned capacity.**
If your table uses less than 18% of provisioned capacity, on-demand is cheaper.
If it uses more than 18% consistently, provisioned is cheaper.

```bash
# Check table utilization
aws cloudwatch get-metric-statistics \
  --namespace AWS/DynamoDB \
  --metric-name ConsumedWriteCapacityUnits \
  --dimensions Name=TableName,Value=my-table \
  --start-time 2026-03-12T00:00:00Z \
  --end-time 2026-03-19T00:00:00Z \
  --period 86400 \
  --statistics Average \
  --unit Count
```

### RDS Cost Optimization

```
Strategy                        │ Savings │ Notes
────────────────────────────────┼─────────┼───────────────────────────
Aurora Serverless v2            │ Variable│ Scales to 0.5 ACU ($43/mo min)
RDS Reserved Instances          │ 30-60%  │ 1 or 3-year terms
Graviton (r7g instead of r6i)  │ 20%     │ Better price-performance
Stop dev/test instances at night│ 65%     │ 10hr/day vs 24hr/day
Right-size with Perf Insights  │ 20-50%  │ Many instances are oversized
Use Aurora I/O-Optimized       │ Fixed   │ Predictable cost, no I/O charges
Read replicas instead of upsize│ Variable│ Scale reads without scaling writes
```

**Auto-stop dev/test RDS instances:**
```bash
# Use Lambda + EventBridge Scheduler
# Stop at 7 PM, start at 7 AM (Mon-Fri)
aws scheduler create-schedule \
  --name stop-dev-rds \
  --schedule-expression "cron(0 19 ? * MON-FRI *)" \
  --schedule-expression-timezone "America/New_York" \
  --target '{
    "Arn": "arn:aws:lambda:us-east-1:123456789:function:manage-rds",
    "RoleArn": "arn:aws:iam::123456789:role/SchedulerRole",
    "Input": "{\"action\":\"stop\",\"instanceId\":\"dev-database\"}"
  }' \
  --flexible-time-window '{"Mode":"OFF"}'
```

### CloudWatch Cost

CloudWatch costs can sneak up. Key cost drivers:

```
Component                      │ Cost                     │ Optimization
───────────────────────────────┼──────────────────────────┼──────────────────────
Custom metrics                 │ $0.30/metric/month       │ Use EMF instead of PutMetricData
Logs ingestion                 │ $0.50/GB                 │ Filter before sending
Logs storage                   │ $0.03/GB/month           │ Set retention periods
Dashboard                      │ $3/dashboard/month       │ Consolidate dashboards
Alarms                         │ $0.10/alarm/month        │ Use composite alarms
Metric Insights queries        │ $0.01/query              │ Cache dashboards
```

```bash
# Set log retention to reduce storage costs
aws logs put-retention-policy \
  --log-group-name /aws/lambda/my-function \
  --retention-in-days 30

# Use subscription filter to only send ERROR logs to long-term storage
aws logs put-subscription-filter \
  --log-group-name /aws/lambda/my-function \
  --filter-name errors-only \
  --filter-pattern "ERROR" \
  --destination-arn arn:aws:firehose:us-east-1:123456789:deliverystream/error-logs
```

---

## FinOps Practices

### Weekly Cost Review Process

```
1. Review Cost Explorer
   - Total spend vs budget (target: within 5%)
   - Service breakdown (top 5 services by cost)
   - Daily cost trends (look for spikes)

2. Check Anomaly Detection alerts
   - Were all anomalies investigated?
   - Were any real issues found?

3. Review Compute Optimizer
   - Any new right-sizing recommendations?
   - Any Graviton migration opportunities?

4. Check Savings Plans utilization
   - Coverage: aim for 70-80% of steady-state
   - Utilization: should be >95% (if <90%, you over-committed)

5. Review untagged resources
   - Tag compliance should be >95%
   - Action: tag or terminate untagged resources

6. Check idle resources
   - Unattached EBS volumes
   - Unused Elastic IPs ($3.60/month each!)
   - Idle NAT Gateways
   - Empty S3 buckets with logging enabled
```

### Cost Per Unit Metrics

Track cost efficiency, not just total cost:

```
Metric                          │ Formula                          │ Target
────────────────────────────────┼──────────────────────────────────┼────────
Cost per request                │ Total cost / total API requests  │ < $0.001
Cost per customer               │ Total cost / active customers    │ Decreasing
Infrastructure cost ratio       │ AWS cost / revenue               │ < 15%
Waste ratio                     │ Idle cost / total cost           │ < 5%
Commitment coverage             │ SP+RI cost / total compute       │ 70-80%
```

### Automated Cost Guardrails

```yaml
# Lambda that checks for expensive resources and alerts
CostGuardrailFunction:
  Type: AWS::Lambda::Function
  Properties:
    Runtime: python3.13
    Handler: index.handler
    Code:
      ZipFile: |
        import boto3

        def handler(event, context):
            ec2 = boto3.client('ec2')
            findings = []

            # Check for expensive instance types
            expensive_types = ['p5.', 'p4d.', 'x2idn.', 'u-']
            instances = ec2.describe_instances(
                Filters=[{'Name': 'instance-state-name', 'Values': ['running']}]
            )
            for r in instances['Reservations']:
                for i in r['Instances']:
                    for prefix in expensive_types:
                        if i['InstanceType'].startswith(prefix):
                            findings.append(
                                f"Expensive instance: {i['InstanceId']} ({i['InstanceType']})"
                            )

            # Check for unattached EBS volumes
            volumes = ec2.describe_volumes(
                Filters=[{'Name': 'status', 'Values': ['available']}]
            )
            for v in volumes['Volumes']:
                if v['Size'] > 100:
                    findings.append(
                        f"Unattached volume: {v['VolumeId']} ({v['Size']} GB, {v['VolumeType']})"
                    )

            # Check for unused Elastic IPs
            eips = ec2.describe_addresses()
            for eip in eips['Addresses']:
                if 'AssociationId' not in eip:
                    findings.append(f"Unused EIP: {eip['PublicIp']} (${0.005*730:.2f}/month)")

            if findings:
                sns = boto3.client('sns')
                sns.publish(
                    TopicArn=os.environ['ALERT_TOPIC'],
                    Subject=f"Cost Guardrail: {len(findings)} issues found",
                    Message='\n'.join(findings)
                )

            return {'findings': len(findings)}
```

---

## Quick Wins Checklist

Ordered by effort vs impact:

```
Priority │ Action                                     │ Savings    │ Effort
─────────┼────────────────────────────────────────────┼────────────┼────────
1        │ Delete unattached EBS volumes              │ Immediate  │ 5 min
2        │ Release unused Elastic IPs                  │ Immediate  │ 5 min
3        │ Set CloudWatch log retention               │ 30-80%     │ 15 min
4        │ Enable S3 Intelligent-Tiering              │ 20-40%     │ 15 min
5        │ Add S3 lifecycle policies                  │ 30-60%     │ 30 min
6        │ Delete old EBS snapshots                   │ Variable   │ 30 min
7        │ Migrate gp2 → gp3 EBS volumes              │ 20%        │ 30 min
8        │ Use Graviton instances                     │ 20-40%     │ 1 hour
9        │ Add VPC Gateway Endpoints (S3, DynamoDB)   │ NAT costs  │ 1 hour
10       │ Right-size EC2 with Compute Optimizer      │ 20-50%     │ 2 hours
11       │ Schedule dev/test RDS stop/start           │ 65%        │ 2 hours
12       │ Purchase Compute Savings Plans             │ 30-40%     │ 2 hours
13       │ Move to Fargate Spot for non-critical      │ 70%        │ 4 hours
14       │ Replace NAT Gateway with VPC Endpoints     │ $30+/mo    │ 4 hours
15       │ Use CloudFront for S3 content delivery     │ 10-50%     │ Half day
```

---

## Cost Calculator Mental Models

### Monthly Cost Estimates for Common Architectures

```
Architecture: Serverless API (small)
  Lambda (1M requests, 256MB, 200ms avg)     $3
  API Gateway HTTP API (1M requests)          $1
  DynamoDB on-demand (1M reads, 500K writes)  $1
  CloudWatch Logs (1GB)                       $0.50
  Total: ~$5.50/month

Architecture: Serverless API (medium)
  Lambda (50M requests, 512MB, 200ms avg)     $110
  API Gateway HTTP API (50M requests)          $50
  DynamoDB on-demand (50M reads, 10M writes)  $25
  CloudWatch Logs (50GB)                       $25
  S3 (100GB)                                  $2.30
  Total: ~$212/month

Architecture: Container API (ECS Fargate)
  Fargate (2 tasks, 1 vCPU, 2GB, 24/7)       $120
  ALB                                         $22
  RDS Aurora Serverless (8 ACU avg)            $550
  NAT Gateway (1, 100GB processed)             $37
  CloudWatch                                   $15
  Total: ~$744/month

Architecture: EC2-based (production)
  3x m7g.xlarge (RI, no upfront, 1yr)         $390
  ALB                                         $22
  RDS r6g.xlarge Multi-AZ (RI)               $480
  ElastiCache r6g.large (RI)                  $160
  NAT Gateway (3, HA, 500GB/mo)               $120
  S3 (1TB)                                    $23
  CloudFront (500GB/mo)                        $43
  CloudWatch                                   $30
  Total: ~$1,268/month
```

These are rough estimates — actual costs vary by usage patterns, region, and negotiated discounts. Use the AWS Pricing Calculator for accurate estimates.
