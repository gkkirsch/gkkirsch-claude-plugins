---
name: cloud-cost-optimization
description: >
  Cloud cost optimization — right-sizing instances, reserved/spot pricing,
  storage tiering, data transfer optimization, cost allocation tags,
  budget alerts, waste detection, and FinOps practices.
  Triggers: "cloud cost", "aws cost", "right sizing", "reserved instance",
  "spot instance", "savings plan", "cost optimization", "finops", "cloud budget".
  NOT for: Network architecture or VPC design (use cloud-networking).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Cloud Cost Optimization

## Right-Sizing Instances

```bash
# Find underutilized EC2 instances (CPU < 10% average over 14 days)
aws cloudwatch get-metric-statistics \
  --namespace AWS/EC2 \
  --metric-name CPUUtilization \
  --dimensions Name=InstanceId,Value=i-abc123 \
  --start-time $(date -v-14d +%Y-%m-%dT%H:%M:%S) \
  --end-time $(date +%Y-%m-%dT%H:%M:%S) \
  --period 86400 \
  --statistics Average

# AWS Compute Optimizer recommendations
aws compute-optimizer get-ec2-instance-recommendations \
  --instance-arns arn:aws:ec2:us-east-1:123456789:instance/i-abc123

# RDS right-sizing
aws cloudwatch get-metric-statistics \
  --namespace AWS/RDS \
  --metric-name CPUUtilization \
  --dimensions Name=DBInstanceIdentifier,Value=production-db \
  --start-time $(date -v-14d +%Y-%m-%dT%H:%M:%S) \
  --end-time $(date +%Y-%m-%dT%H:%M:%S) \
  --period 86400 \
  --statistics Average,Maximum
```

```typescript
// scripts/right-sizing-report.ts
import {
  CloudWatchClient,
  GetMetricStatisticsCommand,
} from "@aws-sdk/client-cloudwatch";
import { EC2Client, DescribeInstancesCommand } from "@aws-sdk/client-ec2";

interface InstanceMetrics {
  instanceId: string;
  instanceType: string;
  avgCpu: number;
  maxCpu: number;
  avgMemory?: number;
  recommendation: string;
  estimatedSavings: string;
}

async function analyzeInstances(): Promise<InstanceMetrics[]> {
  const ec2 = new EC2Client({});
  const cw = new CloudWatchClient({});

  const { Reservations } = await ec2.send(
    new DescribeInstancesCommand({
      Filters: [{ Name: "instance-state-name", Values: ["running"] }],
    })
  );

  const results: InstanceMetrics[] = [];

  for (const reservation of Reservations ?? []) {
    for (const instance of reservation.Instances ?? []) {
      const stats = await cw.send(
        new GetMetricStatisticsCommand({
          Namespace: "AWS/EC2",
          MetricName: "CPUUtilization",
          Dimensions: [
            { Name: "InstanceId", Value: instance.InstanceId },
          ],
          StartTime: new Date(Date.now() - 14 * 86400000),
          EndTime: new Date(),
          Period: 86400,
          Statistics: ["Average", "Maximum"],
        })
      );

      const avgCpu =
        stats.Datapoints?.reduce((sum, dp) => sum + (dp.Average ?? 0), 0) /
        (stats.Datapoints?.length || 1);
      const maxCpu = Math.max(
        ...(stats.Datapoints?.map((dp) => dp.Maximum ?? 0) ?? [0])
      );

      let recommendation = "OK";
      let estimatedSavings = "$0";

      if (avgCpu < 5) {
        recommendation = "TERMINATE or downsize 2 sizes";
        estimatedSavings = "~70%";
      } else if (avgCpu < 20 && maxCpu < 40) {
        recommendation = "Downsize 1 size";
        estimatedSavings = "~40%";
      } else if (avgCpu > 80) {
        recommendation = "Upsize — near capacity";
      }

      results.push({
        instanceId: instance.InstanceId!,
        instanceType: instance.InstanceType!,
        avgCpu: Math.round(avgCpu * 10) / 10,
        maxCpu: Math.round(maxCpu * 10) / 10,
        recommendation,
        estimatedSavings,
      });
    }
  }

  return results;
}
```

## Reserved Instances and Savings Plans

```hcl
# Terraform: Savings Plan (preferred over Reserved Instances)
# Note: Usually purchased via AWS Console, documented here for reference

# Cost allocation tags — essential for tracking savings
resource "aws_instance" "app" {
  instance_type = "t3.large"
  # ...

  tags = {
    Name        = "api-server"
    Environment = "production"
    Team        = "platform"
    CostCenter  = "engineering"
    Service     = "api"
    # These tags appear in Cost Explorer
  }
}
```

```bash
# Compare pricing options
# On-Demand: t3.large = $0.0832/hr = ~$60/month
# 1-yr Reserved (no upfront): ~$42/month (30% savings)
# 1-yr Reserved (all upfront): ~$38/month (37% savings)
# 3-yr Reserved (all upfront): ~$25/month (58% savings)
# Savings Plan (1-yr, compute): ~$40/month (33% savings, flexible)

# Check current Reserved Instance utilization
aws ce get-reservation-utilization \
  --time-period Start=$(date -v-30d +%Y-%m-%d),End=$(date +%Y-%m-%d) \
  --granularity MONTHLY

# Get Savings Plan recommendations
aws ce get-savings-plans-purchase-recommendation \
  --savings-plans-type COMPUTE_SP \
  --term-in-years ONE_YEAR \
  --payment-option NO_UPFRONT \
  --lookback-period-in-days SIXTY_DAYS
```

## Spot Instances

```hcl
# ECS with spot capacity provider
resource "aws_ecs_capacity_provider" "spot" {
  name = "spot"

  auto_scaling_group_provider {
    auto_scaling_group_arn         = aws_autoscaling_group.spot.arn
    managed_termination_protection = "ENABLED"

    managed_scaling {
      maximum_scaling_step_size = 5
      minimum_scaling_step_size = 1
      status                   = "ENABLED"
      target_capacity           = 100
    }
  }
}

# Mixed instances ASG (on-demand base + spot scaling)
resource "aws_autoscaling_group" "mixed" {
  min_size         = 2
  max_size         = 20
  desired_capacity = 4

  mixed_instances_policy {
    instances_distribution {
      on_demand_base_capacity                  = 2   # 2 on-demand minimum
      on_demand_percentage_above_base_capacity = 25  # 25% on-demand, 75% spot
      spot_allocation_strategy                 = "capacity-optimized"
    }

    launch_template {
      launch_template_specification {
        launch_template_id = aws_launch_template.app.id
        version            = "$Latest"
      }

      # Multiple instance types for spot diversity
      override {
        instance_type     = "t3.large"
        weighted_capacity = "1"
      }
      override {
        instance_type     = "t3a.large"
        weighted_capacity = "1"
      }
      override {
        instance_type     = "m5.large"
        weighted_capacity = "1"
      }
      override {
        instance_type     = "m5a.large"
        weighted_capacity = "1"
      }
    }
  }
}

# Spot interruption handler (Lambda)
resource "aws_cloudwatch_event_rule" "spot_interruption" {
  name = "spot-interruption-handler"

  event_pattern = jsonencode({
    source      = ["aws.ec2"]
    detail-type = ["EC2 Spot Instance Interruption Warning"]
  })
}
```

## Storage Cost Optimization

```hcl
# S3 lifecycle policies
resource "aws_s3_bucket_lifecycle_configuration" "assets" {
  bucket = aws_s3_bucket.assets.id

  rule {
    id     = "archive-old-objects"
    status = "Enabled"

    transition {
      days          = 30
      storage_class = "STANDARD_IA"  # $0.0125/GB (vs $0.023/GB standard)
    }

    transition {
      days          = 90
      storage_class = "GLACIER_IR"    # $0.004/GB, instant retrieval
    }

    transition {
      days          = 365
      storage_class = "DEEP_ARCHIVE"  # $0.00099/GB
    }

    expiration {
      days = 2555  # 7 years
    }
  }

  rule {
    id     = "cleanup-multipart-uploads"
    status = "Enabled"

    abort_incomplete_multipart_upload {
      days_after_initiation = 7
    }
  }

  rule {
    id     = "expire-old-versions"
    status = "Enabled"

    noncurrent_version_transition {
      noncurrent_days = 30
      storage_class   = "STANDARD_IA"
    }

    noncurrent_version_expiration {
      noncurrent_days = 90
    }
  }
}

# EBS volume optimization
resource "aws_ebs_volume" "data" {
  availability_zone = "us-east-1a"
  size              = 100
  type              = "gp3"  # gp3 is 20% cheaper than gp2, with better baseline
  iops              = 3000   # gp3 includes 3000 IOPS free
  throughput        = 125    # gp3 includes 125 MB/s free

  tags = { Name = "data-volume" }
}
```

## Data Transfer Optimization

```hcl
# VPC endpoints — avoid NAT Gateway charges for AWS services
resource "aws_vpc_endpoint" "s3" {
  vpc_id       = aws_vpc.main.id
  service_name = "com.amazonaws.us-east-1.s3"
  vpc_endpoint_type = "Gateway"

  route_table_ids = aws_route_table.private[*].id
  # S3 traffic now stays in AWS network — $0.00/GB instead of $0.045/GB via NAT
}

resource "aws_vpc_endpoint" "ecr_api" {
  vpc_id              = aws_vpc.main.id
  service_name        = "com.amazonaws.us-east-1.ecr.api"
  vpc_endpoint_type   = "Interface"
  subnet_ids          = aws_subnet.private[*].id
  security_group_ids  = [aws_security_group.vpc_endpoints.id]
  private_dns_enabled = true
  # ECR image pulls don't go through NAT — saves on data transfer
}

resource "aws_vpc_endpoint" "ecr_dkr" {
  vpc_id              = aws_vpc.main.id
  service_name        = "com.amazonaws.us-east-1.ecr.dkr"
  vpc_endpoint_type   = "Interface"
  subnet_ids          = aws_subnet.private[*].id
  security_group_ids  = [aws_security_group.vpc_endpoints.id]
  private_dns_enabled = true
}

# CloudFront for API responses — reduces origin data transfer
# API origin data transfer: $0.09/GB
# CloudFront to internet: $0.085/GB (first 10 TB)
# CloudFront to origin: $0.00/GB (free)
# Net savings on cache hits
```

## Budget Alerts

```hcl
resource "aws_budgets_budget" "monthly" {
  name         = "monthly-total"
  budget_type  = "COST"
  limit_amount = "5000"
  limit_unit   = "USD"
  time_unit    = "MONTHLY"

  notification {
    comparison_operator        = "GREATER_THAN"
    threshold                  = 80
    threshold_type             = "PERCENTAGE"
    notification_type          = "ACTUAL"
    subscriber_email_addresses = ["engineering@example.com"]
  }

  notification {
    comparison_operator        = "GREATER_THAN"
    threshold                  = 100
    threshold_type             = "PERCENTAGE"
    notification_type          = "FORECASTED"
    subscriber_email_addresses = ["engineering@example.com", "finance@example.com"]
  }
}

# Per-service budget
resource "aws_budgets_budget" "rds" {
  name         = "rds-budget"
  budget_type  = "COST"
  limit_amount = "1000"
  limit_unit   = "USD"
  time_unit    = "MONTHLY"

  cost_filter {
    name   = "Service"
    values = ["Amazon Relational Database Service"]
  }

  notification {
    comparison_operator        = "GREATER_THAN"
    threshold                  = 90
    threshold_type             = "PERCENTAGE"
    notification_type          = "ACTUAL"
    subscriber_email_addresses = ["dba@example.com"]
  }
}
```

## Waste Detection Script

```bash
#!/bin/bash
# scripts/find-waste.sh — Identify cloud waste

echo "=== UNATTACHED EBS VOLUMES ==="
aws ec2 describe-volumes \
  --filters Name=status,Values=available \
  --query 'Volumes[].{ID:VolumeId,Size:Size,Type:VolumeType,Created:CreateTime}' \
  --output table

echo ""
echo "=== UNUSED ELASTIC IPS ==="
aws ec2 describe-addresses \
  --query 'Addresses[?AssociationId==null].{IP:PublicIp,AllocId:AllocationId}' \
  --output table

echo ""
echo "=== OLD SNAPSHOTS (>90 days) ==="
CUTOFF=$(date -v-90d +%Y-%m-%dT%H:%M:%S)
aws ec2 describe-snapshots --owner-ids self \
  --query "Snapshots[?StartTime<'${CUTOFF}'].{ID:SnapshotId,Size:VolumeSize,Date:StartTime}" \
  --output table

echo ""
echo "=== IDLE LOAD BALANCERS (no targets) ==="
for arn in $(aws elbv2 describe-load-balancers --query 'LoadBalancers[].LoadBalancerArn' --output text); do
  TG_COUNT=$(aws elbv2 describe-target-groups --load-balancer-arn "$arn" --query 'TargetGroups | length(@)')
  if [ "$TG_COUNT" -eq 0 ]; then
    NAME=$(aws elbv2 describe-load-balancers --load-balancer-arns "$arn" --query 'LoadBalancers[0].LoadBalancerName' --output text)
    echo "  $NAME ($arn) — no target groups"
  fi
done

echo ""
echo "=== OVERSIZED RDS INSTANCES ==="
for db in $(aws rds describe-db-instances --query 'DBInstances[].DBInstanceIdentifier' --output text); do
  CPU=$(aws cloudwatch get-metric-statistics \
    --namespace AWS/RDS --metric-name CPUUtilization \
    --dimensions Name=DBInstanceIdentifier,Value=$db \
    --start-time $(date -v-7d +%Y-%m-%dT%H:%M:%S) --end-time $(date +%Y-%m-%dT%H:%M:%S) \
    --period 604800 --statistics Average \
    --query 'Datapoints[0].Average' --output text 2>/dev/null)
  if [ "$CPU" != "None" ] && (( $(echo "$CPU < 10" | bc -l 2>/dev/null || echo 0) )); then
    CLASS=$(aws rds describe-db-instances --db-instance-identifier "$db" --query 'DBInstances[0].DBInstanceClass' --output text)
    echo "  $db ($CLASS) — avg CPU: ${CPU}%"
  fi
done
```

## Cost Allocation Tags

```hcl
# Enforce tagging with AWS Config rule
resource "aws_config_config_rule" "required_tags" {
  name = "required-tags"

  source {
    owner             = "AWS"
    source_identifier = "REQUIRED_TAGS"
  }

  input_parameters = jsonencode({
    tag1Key   = "Environment"
    tag2Key   = "Team"
    tag3Key   = "Service"
    tag4Key   = "CostCenter"
  })

  scope {
    compliance_resource_types = [
      "AWS::EC2::Instance",
      "AWS::RDS::DBInstance",
      "AWS::S3::Bucket",
      "AWS::ElasticLoadBalancingV2::LoadBalancer",
    ]
  }
}

# Default tags in provider (apply to all resources)
provider "aws" {
  region = "us-east-1"

  default_tags {
    tags = {
      ManagedBy   = "terraform"
      Environment = var.environment
      Project     = var.project_name
    }
  }
}
```

## Gotchas

1. **Savings Plans lock you in, Reserved Instances lock you down** — Savings Plans are flexible (any instance type in a region), but you commit to a $/hour spend for 1-3 years. Reserved Instances lock you to a specific instance type. If you right-size later, the reservation doesn't move. Start with Savings Plans unless you have very predictable, unchanging workloads.

2. **Data transfer is the hidden cost** — AWS doesn't charge for inbound data, but outbound costs $0.09/GB. Cross-AZ traffic is $0.01/GB each way. NAT Gateway processing is $0.045/GB. A busy microservices architecture with cross-AZ calls, external API integrations, and NAT Gateway traffic can easily spend more on data transfer than compute.

3. **gp2 to gp3 migration is free performance** — gp3 volumes are 20% cheaper than gp2 and include 3000 IOPS and 125 MB/s baseline (gp2 baseline depends on size). Most gp2 volumes can be converted to gp3 with `aws ec2 modify-volume --volume-type gp3` with zero downtime. This is the easiest cost optimization available.

4. **CloudWatch costs scale with metrics** — Custom metrics cost $0.30/metric/month. Detailed monitoring (1-minute intervals) costs 7x more than basic (5-minute). A fleet of 100 instances with 20 custom metrics each at detailed monitoring = $600/month just for CloudWatch. Use metric filters and aggregation instead of per-instance custom metrics.

5. **Spot interruptions cascade in stateful workloads** — Spot instances save 60-90% but get terminated with 2 minutes warning. Databases, message queues, and single-instance services should NEVER run on spot. Use spot only for stateless, horizontally scalable workloads (web servers, batch jobs, CI runners) with proper interruption handling.

6. **Unused Reserved Instances still cost money** — If you reserved a `c5.2xlarge` but later downsized to `c5.large`, you pay for the full reservation AND the new on-demand instance. Monitor RI utilization weekly with `aws ce get-reservation-utilization`. Sell unused reservations on the EC2 Reserved Instance Marketplace before they expire.
