# AWS Resource Patterns for Terraform

Ready-to-use Terraform patterns for common AWS resources. Each block is production-oriented,
copy-paste-ready HCL with inline comments explaining important settings. Use these as starting
points and adapt values (CIDR ranges, instance sizes, naming conventions) to your environment.

---

## VPC & Networking

### Complete VPC (VPC + Subnets + IGW + NAT + Route Tables)

```hcl
locals {
  az_count = 3
  azs      = slice(data.aws_availability_zones.available.names, 0, local.az_count)
}

data "aws_availability_zones" "available" {
  state = "available"
}

resource "aws_vpc" "main" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_support   = true   # Required for private hosted zones and VPC endpoints
  enable_dns_hostnames = true   # Required for ECS service discovery and RDS DNS

  tags = {
    Name = "${var.project}-vpc"
  }
}

resource "aws_internet_gateway" "main" {
  vpc_id = aws_vpc.main.id

  tags = {
    Name = "${var.project}-igw"
  }
}

# Public subnets — one per AZ for ALBs, NAT Gateways, bastion hosts
resource "aws_subnet" "public" {
  count                   = local.az_count
  vpc_id                  = aws_vpc.main.id
  cidr_block              = cidrsubnet(aws_vpc.main.cidr_block, 8, count.index)       # /24 subnets
  availability_zone       = local.azs[count.index]
  map_public_ip_on_launch = true  # Instances launched here get public IPs automatically

  tags = {
    Name = "${var.project}-public-${local.azs[count.index]}"
    Tier = "public"
  }
}

# Private subnets — one per AZ for ECS tasks, RDS, Lambda
resource "aws_subnet" "private" {
  count             = local.az_count
  vpc_id            = aws_vpc.main.id
  cidr_block        = cidrsubnet(aws_vpc.main.cidr_block, 8, count.index + 10)       # Offset by 10
  availability_zone = local.azs[count.index]

  tags = {
    Name = "${var.project}-private-${local.azs[count.index]}"
    Tier = "private"
  }
}

# Elastic IP for NAT Gateway — one per AZ for high availability
resource "aws_eip" "nat" {
  count  = local.az_count
  domain = "vpc"

  tags = {
    Name = "${var.project}-nat-eip-${local.azs[count.index]}"
  }
}

# NAT Gateway — one per AZ so private subnets survive a single-AZ outage
resource "aws_nat_gateway" "main" {
  count         = local.az_count
  allocation_id = aws_eip.nat[count.index].id
  subnet_id     = aws_subnet.public[count.index].id

  tags = {
    Name = "${var.project}-nat-${local.azs[count.index]}"
  }

  depends_on = [aws_internet_gateway.main]
}

# Public route table — routes internet traffic through the IGW
resource "aws_route_table" "public" {
  vpc_id = aws_vpc.main.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.main.id
  }

  tags = {
    Name = "${var.project}-public-rt"
  }
}

resource "aws_route_table_association" "public" {
  count          = local.az_count
  subnet_id      = aws_subnet.public[count.index].id
  route_table_id = aws_route_table.public.id
}

# Private route tables — each AZ routes through its own NAT Gateway
resource "aws_route_table" "private" {
  count  = local.az_count
  vpc_id = aws_vpc.main.id

  route {
    cidr_block     = "0.0.0.0/0"
    nat_gateway_id = aws_nat_gateway.main[count.index].id
  }

  tags = {
    Name = "${var.project}-private-rt-${local.azs[count.index]}"
  }
}

resource "aws_route_table_association" "private" {
  count          = local.az_count
  subnet_id      = aws_subnet.private[count.index].id
  route_table_id = aws_route_table.private[count.index].id
}
```

### VPC Endpoints (S3 Gateway, ECR, CloudWatch)

```hcl
# S3 Gateway Endpoint — free, avoids NAT Gateway data processing charges for S3 traffic
resource "aws_vpc_endpoint" "s3" {
  vpc_id       = aws_vpc.main.id
  service_name = "com.amazonaws.${var.region}.s3"
  vpc_endpoint_type = "Gateway"
  route_table_ids   = aws_route_table.private[*].id

  tags = {
    Name = "${var.project}-s3-endpoint"
  }
}

# Security group for interface endpoints
resource "aws_security_group" "vpc_endpoints" {
  name_prefix = "${var.project}-vpce-"
  vpc_id      = aws_vpc.main.id

  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = [aws_vpc.main.cidr_block]
  }

  tags = {
    Name = "${var.project}-vpce-sg"
  }
}

# ECR API endpoint — required for docker pull in private subnets without NAT
resource "aws_vpc_endpoint" "ecr_api" {
  vpc_id              = aws_vpc.main.id
  service_name        = "com.amazonaws.${var.region}.ecr.api"
  vpc_endpoint_type   = "Interface"
  private_dns_enabled = true
  subnet_ids          = aws_subnet.private[*].id
  security_group_ids  = [aws_security_group.vpc_endpoints.id]

  tags = {
    Name = "${var.project}-ecr-api-endpoint"
  }
}

# ECR DKR endpoint — required alongside ecr.api for image layer downloads
resource "aws_vpc_endpoint" "ecr_dkr" {
  vpc_id              = aws_vpc.main.id
  service_name        = "com.amazonaws.${var.region}.ecr.dkr"
  vpc_endpoint_type   = "Interface"
  private_dns_enabled = true
  subnet_ids          = aws_subnet.private[*].id
  security_group_ids  = [aws_security_group.vpc_endpoints.id]

  tags = {
    Name = "${var.project}-ecr-dkr-endpoint"
  }
}

# CloudWatch Logs endpoint — lets Fargate tasks push logs without NAT
resource "aws_vpc_endpoint" "logs" {
  vpc_id              = aws_vpc.main.id
  service_name        = "com.amazonaws.${var.region}.logs"
  vpc_endpoint_type   = "Interface"
  private_dns_enabled = true
  subnet_ids          = aws_subnet.private[*].id
  security_group_ids  = [aws_security_group.vpc_endpoints.id]

  tags = {
    Name = "${var.project}-logs-endpoint"
  }
}
```

### Transit Gateway

```hcl
resource "aws_ec2_transit_gateway" "main" {
  description                     = "Central hub for VPC-to-VPC and on-premises connectivity"
  default_route_table_association = "enable"
  default_route_table_propagation = "enable"
  dns_support                     = "enable"
  auto_accept_shared_attachments  = "disable"  # Require explicit acceptance for security

  tags = {
    Name = "${var.project}-tgw"
  }
}

resource "aws_ec2_transit_gateway_vpc_attachment" "main" {
  transit_gateway_id = aws_ec2_transit_gateway.main.id
  vpc_id             = aws_vpc.main.id
  subnet_ids         = aws_subnet.private[*].id

  tags = {
    Name = "${var.project}-tgw-attachment"
  }
}
```

---

## ECS (Elastic Container Service)

### ECS Cluster with Fargate

```hcl
resource "aws_ecs_cluster" "main" {
  name = "${var.project}-cluster"

  setting {
    name  = "containerInsights"
    value = "enabled"            # Enables CloudWatch Container Insights metrics
  }

  configuration {
    execute_command_configuration {
      logging = "OVERRIDE"

      log_configuration {
        cloud_watch_log_group_name = aws_cloudwatch_log_group.ecs_exec.name
      }
    }
  }

  tags = {
    Name = "${var.project}-cluster"
  }
}

resource "aws_ecs_cluster_capacity_providers" "main" {
  cluster_name       = aws_ecs_cluster.main.name
  capacity_providers = ["FARGATE", "FARGATE_SPOT"]

  default_capacity_provider_strategy {
    capacity_provider = "FARGATE"
    weight            = 1
    base              = 1     # At least 1 task always runs on regular Fargate
  }

  default_capacity_provider_strategy {
    capacity_provider = "FARGATE_SPOT"
    weight            = 3     # 75% of additional tasks use Spot for cost savings
  }
}

resource "aws_cloudwatch_log_group" "ecs_exec" {
  name              = "/ecs/${var.project}/exec"
  retention_in_days = 30
}
```

### Task Definition (with Container Definitions, Logging, Secrets)

```hcl
resource "aws_ecs_task_definition" "app" {
  family                   = "${var.project}-app"
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"         # Required for Fargate
  cpu                      = 512              # 0.5 vCPU
  memory                   = 1024             # 1 GB
  execution_role_arn       = aws_iam_role.ecs_execution.arn   # Pulls images, writes logs
  task_role_arn            = aws_iam_role.ecs_task.arn         # App-level AWS API calls

  container_definitions = jsonencode([
    {
      name      = "app"
      image     = "${var.ecr_repo_url}:${var.image_tag}"
      essential = true

      portMappings = [
        {
          containerPort = 8080
          protocol      = "tcp"
        }
      ]

      environment = [
        { name = "NODE_ENV", value = "production" },
        { name = "PORT", value = "8080" }
      ]

      # Secrets pulled from SSM Parameter Store or Secrets Manager at launch
      secrets = [
        {
          name      = "DATABASE_URL"
          valueFrom = aws_ssm_parameter.database_url.arn
        },
        {
          name      = "API_KEY"
          valueFrom = "${aws_secretsmanager_secret.api_key.arn}:api_key::"
        }
      ]

      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-group"         = aws_cloudwatch_log_group.app.name
          "awslogs-region"        = var.region
          "awslogs-stream-prefix" = "app"
        }
      }

      healthCheck = {
        command     = ["CMD-SHELL", "curl -f http://localhost:8080/health || exit 1"]
        interval    = 30
        timeout     = 5
        retries     = 3
        startPeriod = 60    # Grace period during startup
      }
    }
  ])

  runtime_platform {
    operating_system_family = "LINUX"
    cpu_architecture        = "ARM64"   # Graviton — ~20% cheaper than x86
  }

  tags = {
    Name = "${var.project}-app-task"
  }
}

resource "aws_cloudwatch_log_group" "app" {
  name              = "/ecs/${var.project}/app"
  retention_in_days = 90
}
```

### ECS Service with Load Balancer

```hcl
resource "aws_ecs_service" "app" {
  name            = "${var.project}-app"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.app.arn
  desired_count   = 2
  launch_type     = "FARGATE"

  # Deploy new tasks before draining old ones — zero-downtime deployments
  deployment_minimum_healthy_percent = 100
  deployment_maximum_percent         = 200

  deployment_circuit_breaker {
    enable   = true
    rollback = true   # Automatically roll back failed deployments
  }

  network_configuration {
    subnets          = aws_subnet.private[*].id
    security_groups  = [aws_security_group.ecs_tasks.id]
    assign_public_ip = false    # Tasks live in private subnets behind NAT/VPC endpoints
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.app.arn
    container_name   = "app"
    container_port   = 8080
  }

  # Ignore desired_count changes from auto scaling
  lifecycle {
    ignore_changes = [desired_count]
  }

  depends_on = [aws_lb_listener.https]
}

resource "aws_security_group" "ecs_tasks" {
  name_prefix = "${var.project}-ecs-tasks-"
  vpc_id      = aws_vpc.main.id

  ingress {
    from_port       = 8080
    to_port         = 8080
    protocol        = "tcp"
    security_groups = [aws_security_group.alb.id]   # Only accept traffic from ALB
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "${var.project}-ecs-tasks-sg"
  }
}
```

### ECS Auto Scaling (Target Tracking and Step Scaling)

```hcl
resource "aws_appautoscaling_target" "ecs" {
  max_capacity       = 10
  min_capacity       = 2
  resource_id        = "service/${aws_ecs_cluster.main.name}/${aws_ecs_service.app.name}"
  scalable_dimension = "ecs:service:DesiredCount"
  service_namespace  = "ecs"
}

# Target tracking — maintain 70% average CPU utilization
resource "aws_appautoscaling_policy" "ecs_cpu" {
  name               = "${var.project}-cpu-tracking"
  policy_type        = "TargetTrackingScaling"
  resource_id        = aws_appautoscaling_target.ecs.resource_id
  scalable_dimension = aws_appautoscaling_target.ecs.scalable_dimension
  service_namespace  = aws_appautoscaling_target.ecs.service_namespace

  target_tracking_scaling_policy_configuration {
    predefined_metric_specification {
      predefined_metric_type = "ECSServiceAverageCPUUtilization"
    }
    target_value       = 70.0
    scale_in_cooldown  = 300    # Wait 5 minutes before scaling in
    scale_out_cooldown = 60     # Scale out quickly
  }
}

# Target tracking — scale based on ALB request count per target
resource "aws_appautoscaling_policy" "ecs_requests" {
  name               = "${var.project}-request-tracking"
  policy_type        = "TargetTrackingScaling"
  resource_id        = aws_appautoscaling_target.ecs.resource_id
  scalable_dimension = aws_appautoscaling_target.ecs.scalable_dimension
  service_namespace  = aws_appautoscaling_target.ecs.service_namespace

  target_tracking_scaling_policy_configuration {
    predefined_metric_specification {
      predefined_metric_type = "ALBRequestCountPerTarget"
      resource_label         = "${aws_lb.main.arn_suffix}/${aws_lb_target_group.app.arn_suffix}"
    }
    target_value       = 1000   # 1000 requests per target per minute
    scale_in_cooldown  = 300
    scale_out_cooldown = 60
  }
}
```

### ECS Service Discovery

```hcl
resource "aws_service_discovery_private_dns_namespace" "main" {
  name        = "${var.project}.local"
  vpc         = aws_vpc.main.id
  description = "Service discovery namespace for internal service-to-service communication"
}

resource "aws_service_discovery_service" "app" {
  name = "app"

  dns_config {
    namespace_id = aws_service_discovery_private_dns_namespace.main.id

    dns_records {
      ttl  = 10     # Low TTL for fast failover
      type = "A"
    }

    routing_policy = "MULTIVALUE"   # Returns all healthy IPs
  }

  health_check_custom_config {
    failure_threshold = 1    # Deregister unhealthy tasks after 1 failed check
  }
}
```

---

## RDS (Relational Database Service)

### RDS Aurora PostgreSQL Cluster

```hcl
resource "aws_rds_cluster" "main" {
  cluster_identifier     = "${var.project}-aurora"
  engine                 = "aurora-postgresql"
  engine_version         = "15.4"
  database_name          = var.db_name
  master_username        = var.db_username
  manage_master_user_password = true  # Let RDS manage the password in Secrets Manager

  db_subnet_group_name   = aws_db_subnet_group.main.name
  vpc_security_group_ids = [aws_security_group.rds.id]

  storage_encrypted = true
  kms_key_id        = aws_kms_key.rds.arn

  backup_retention_period      = 35     # Maximum retention
  preferred_backup_window      = "03:00-04:00"
  preferred_maintenance_window = "sun:04:00-sun:05:00"
  copy_tags_to_snapshot        = true
  deletion_protection          = true
  skip_final_snapshot          = false
  final_snapshot_identifier    = "${var.project}-aurora-final"

  enabled_cloudwatch_logs_exports = ["postgresql"]

  tags = {
    Name = "${var.project}-aurora"
  }
}

resource "aws_rds_cluster_instance" "main" {
  count              = 2    # Writer + one reader
  identifier         = "${var.project}-aurora-${count.index}"
  cluster_identifier = aws_rds_cluster.main.id
  instance_class     = "db.r6g.large"
  engine             = aws_rds_cluster.main.engine
  engine_version     = aws_rds_cluster.main.engine_version

  performance_insights_enabled    = true
  performance_insights_kms_key_id = aws_kms_key.rds.arn

  monitoring_interval = 60    # Enhanced Monitoring every 60 seconds
  monitoring_role_arn = aws_iam_role.rds_monitoring.arn

  tags = {
    Name = "${var.project}-aurora-${count.index}"
  }
}
```

### Aurora Serverless v2

```hcl
resource "aws_rds_cluster" "serverless" {
  cluster_identifier     = "${var.project}-serverless"
  engine                 = "aurora-postgresql"
  engine_mode            = "provisioned"       # Serverless v2 uses provisioned mode
  engine_version         = "15.4"
  database_name          = var.db_name
  master_username        = var.db_username
  manage_master_user_password = true

  db_subnet_group_name   = aws_db_subnet_group.main.name
  vpc_security_group_ids = [aws_security_group.rds.id]
  storage_encrypted      = true

  serverlessv2_scaling_configuration {
    min_capacity = 0.5    # Minimum 0.5 ACU — scales to near zero during idle
    max_capacity = 16     # Maximum 16 ACUs during peak load
  }

  tags = {
    Name = "${var.project}-serverless"
  }
}

resource "aws_rds_cluster_instance" "serverless" {
  cluster_identifier = aws_rds_cluster.serverless.id
  identifier         = "${var.project}-serverless-0"
  instance_class     = "db.serverless"     # Required for Serverless v2
  engine             = aws_rds_cluster.serverless.engine
  engine_version     = aws_rds_cluster.serverless.engine_version
}
```

### Parameter Groups and Subnet Groups

```hcl
resource "aws_rds_cluster_parameter_group" "main" {
  name   = "${var.project}-aurora-pg15"
  family = "aurora-postgresql15"

  parameter {
    name  = "log_min_duration_statement"
    value = "1000"   # Log queries slower than 1 second
  }

  parameter {
    name  = "shared_preload_libraries"
    value = "pg_stat_statements"
  }

  parameter {
    name         = "max_connections"
    value        = "500"
    apply_method = "pending-reboot"   # Requires reboot to take effect
  }

  tags = {
    Name = "${var.project}-aurora-pg15"
  }
}

resource "aws_db_subnet_group" "main" {
  name       = "${var.project}-db-subnets"
  subnet_ids = aws_subnet.private[*].id

  tags = {
    Name = "${var.project}-db-subnets"
  }
}
```

### RDS Proxy

```hcl
resource "aws_db_proxy" "main" {
  name                   = "${var.project}-rds-proxy"
  debug_logging          = false
  engine_family          = "POSTGRESQL"
  idle_client_timeout    = 1800
  require_tls            = true
  role_arn               = aws_iam_role.rds_proxy.arn
  vpc_security_group_ids = [aws_security_group.rds.id]
  vpc_subnet_ids         = aws_subnet.private[*].id

  auth {
    auth_scheme = "SECRETS"
    description = "RDS managed master user secret"
    iam_auth    = "REQUIRED"
    secret_arn  = aws_rds_cluster.main.master_user_secret[0].secret_arn
  }

  tags = {
    Name = "${var.project}-rds-proxy"
  }
}

resource "aws_db_proxy_default_target_group" "main" {
  db_proxy_name = aws_db_proxy.main.name

  connection_pool_config {
    max_connections_percent      = 100
    max_idle_connections_percent = 50
    connection_borrow_timeout    = 120
  }
}

resource "aws_db_proxy_target" "main" {
  db_proxy_name         = aws_db_proxy.main.name
  target_group_name     = aws_db_proxy_default_target_group.main.name
  db_cluster_identifier = aws_rds_cluster.main.id
}
```

---

## Lambda

### Lambda Function (with Zip Deployment)

```hcl
data "archive_file" "lambda" {
  type        = "zip"
  source_dir  = "${path.module}/src/lambda"
  output_path = "${path.module}/build/lambda.zip"
}

resource "aws_lambda_function" "main" {
  function_name    = "${var.project}-handler"
  role             = aws_iam_role.lambda.arn
  handler          = "index.handler"
  runtime          = "nodejs20.x"
  architectures    = ["arm64"]       # Graviton — better price-performance
  filename         = data.archive_file.lambda.output_path
  source_code_hash = data.archive_file.lambda.output_base64sha256
  timeout          = 30
  memory_size      = 256

  environment {
    variables = {
      TABLE_NAME  = aws_dynamodb_table.main.name
      LOG_LEVEL   = "info"
    }
  }

  tracing_config {
    mode = "Active"   # Enable X-Ray tracing
  }

  dead_letter_config {
    target_arn = aws_sqs_queue.lambda_dlq.arn
  }

  tags = {
    Name = "${var.project}-handler"
  }
}
```

### Lambda with Container Image

```hcl
resource "aws_lambda_function" "container" {
  function_name = "${var.project}-container-handler"
  role          = aws_iam_role.lambda.arn
  package_type  = "Image"
  image_uri     = "${aws_ecr_repository.lambda.repository_url}:latest"
  timeout       = 60
  memory_size   = 512
  architectures = ["arm64"]

  image_config {
    command           = ["app.handler"]
    entry_point       = ["/lambda-entrypoint.sh"]
    working_directory = "/var/task"
  }

  environment {
    variables = {
      ENV = "production"
    }
  }

  tags = {
    Name = "${var.project}-container-handler"
  }
}
```

### Lambda Layers

```hcl
resource "aws_lambda_layer_version" "shared_deps" {
  layer_name          = "${var.project}-shared-deps"
  filename            = "${path.module}/build/layer.zip"
  source_code_hash    = filebase64sha256("${path.module}/build/layer.zip")
  compatible_runtimes = ["nodejs20.x", "nodejs18.x"]
  compatible_architectures = ["arm64", "x86_64"]
  description         = "Shared dependencies: AWS SDK extensions, logging utilities"
}

# Reference the layer in a function
resource "aws_lambda_function" "with_layer" {
  function_name = "${var.project}-layered"
  role          = aws_iam_role.lambda.arn
  handler       = "index.handler"
  runtime       = "nodejs20.x"
  filename      = data.archive_file.lambda.output_path
  source_code_hash = data.archive_file.lambda.output_base64sha256
  timeout       = 30
  memory_size   = 256

  layers = [
    aws_lambda_layer_version.shared_deps.arn,
  ]

  tags = {
    Name = "${var.project}-layered"
  }
}
```

### Lambda Event Source Mappings (SQS, DynamoDB Streams, Kinesis)

```hcl
# SQS event source — Lambda polls the queue and invokes with batches
resource "aws_lambda_event_source_mapping" "sqs" {
  event_source_arn                   = aws_sqs_queue.events.arn
  function_name                      = aws_lambda_function.main.arn
  batch_size                         = 10     # Up to 10 messages per invocation
  maximum_batching_window_in_seconds = 5      # Wait up to 5s to fill the batch
  function_response_types            = ["ReportBatchItemFailures"]  # Partial batch failure

  scaling_config {
    maximum_concurrency = 10   # Limit concurrent Lambda executions
  }
}

# DynamoDB Streams event source
resource "aws_lambda_event_source_mapping" "dynamodb" {
  event_source_arn  = aws_dynamodb_table.main.stream_arn
  function_name     = aws_lambda_function.main.arn
  starting_position = "LATEST"
  batch_size        = 100
  maximum_batching_window_in_seconds   = 10
  maximum_retry_attempts               = 3
  bisect_batch_on_function_error       = true   # Split failed batches to isolate bad records
  maximum_record_age_in_seconds        = 86400  # Skip records older than 24 hours
  parallelization_factor               = 2      # 2 concurrent batches per shard

  destination_config {
    on_failure {
      destination_arn = aws_sqs_queue.stream_dlq.arn
    }
  }

  filter_criteria {
    filter {
      pattern = jsonencode({
        eventName = ["INSERT", "MODIFY"]
      })
    }
  }
}
```

### Lambda with API Gateway v2 (HTTP API)

```hcl
resource "aws_apigatewayv2_api" "main" {
  name          = "${var.project}-api"
  protocol_type = "HTTP"

  cors_configuration {
    allow_origins = ["https://${var.domain_name}"]
    allow_methods = ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
    allow_headers = ["Content-Type", "Authorization"]
    max_age       = 3600
  }
}

resource "aws_apigatewayv2_stage" "default" {
  api_id      = aws_apigatewayv2_api.main.id
  name        = "$default"
  auto_deploy = true

  access_log_settings {
    destination_arn = aws_cloudwatch_log_group.api_gw.arn
    format = jsonencode({
      requestId      = "$context.requestId"
      ip             = "$context.identity.sourceIp"
      httpMethod     = "$context.httpMethod"
      path           = "$context.path"
      status         = "$context.status"
      latency        = "$context.responseLatency"
      integrationErr = "$context.integrationErrorMessage"
    })
  }
}

resource "aws_apigatewayv2_integration" "lambda" {
  api_id                 = aws_apigatewayv2_api.main.id
  integration_type       = "AWS_PROXY"
  integration_uri        = aws_lambda_function.main.invoke_arn
  payload_format_version = "2.0"      # Use v2 payload format for HTTP APIs
}

resource "aws_apigatewayv2_route" "default" {
  api_id    = aws_apigatewayv2_api.main.id
  route_key = "$default"               # Catch-all route
  target    = "integrations/${aws_apigatewayv2_integration.lambda.id}"
}

resource "aws_lambda_permission" "api_gw" {
  statement_id  = "AllowAPIGateway"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.main.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.main.execution_arn}/*/*"
}

resource "aws_cloudwatch_log_group" "api_gw" {
  name              = "/aws/apigateway/${var.project}"
  retention_in_days = 30
}
```

### Lambda VPC Configuration

```hcl
resource "aws_lambda_function" "vpc_lambda" {
  function_name = "${var.project}-vpc-handler"
  role          = aws_iam_role.lambda.arn
  handler       = "index.handler"
  runtime       = "nodejs20.x"
  filename      = data.archive_file.lambda.output_path
  source_code_hash = data.archive_file.lambda.output_base64sha256
  timeout       = 30
  memory_size   = 256

  vpc_config {
    subnet_ids         = aws_subnet.private[*].id
    security_group_ids = [aws_security_group.lambda.id]
  }

  tags = {
    Name = "${var.project}-vpc-handler"
  }
}

resource "aws_security_group" "lambda" {
  name_prefix = "${var.project}-lambda-"
  vpc_id      = aws_vpc.main.id

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "${var.project}-lambda-sg"
  }
}
```

---

## S3

### S3 Bucket with Best Practices (Encryption, Versioning, Logging)

```hcl
resource "aws_s3_bucket" "main" {
  bucket = "${var.project}-${var.environment}-${data.aws_caller_identity.current.account_id}"

  tags = {
    Name = "${var.project}-main"
  }
}

resource "aws_s3_bucket_versioning" "main" {
  bucket = aws_s3_bucket.main.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "main" {
  bucket = aws_s3_bucket.main.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm     = "aws:kms"
      kms_master_key_id = aws_kms_key.s3.arn
    }
    bucket_key_enabled = true   # Reduces KMS API calls and cost
  }
}

resource "aws_s3_bucket_public_access_block" "main" {
  bucket                  = aws_s3_bucket.main.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket_logging" "main" {
  bucket        = aws_s3_bucket.main.id
  target_bucket = aws_s3_bucket.logs.id
  target_prefix = "s3-access-logs/${aws_s3_bucket.main.id}/"
}
```

### S3 Lifecycle Rules

```hcl
resource "aws_s3_bucket_lifecycle_configuration" "main" {
  bucket = aws_s3_bucket.main.id

  rule {
    id     = "transition-to-ia"
    status = "Enabled"

    transition {
      days          = 30
      storage_class = "STANDARD_IA"    # Infrequent Access after 30 days
    }

    transition {
      days          = 90
      storage_class = "GLACIER_IR"     # Glacier Instant Retrieval after 90 days
    }

    transition {
      days          = 365
      storage_class = "DEEP_ARCHIVE"   # Deep Archive after 1 year
    }
  }

  rule {
    id     = "expire-noncurrent"
    status = "Enabled"

    noncurrent_version_expiration {
      noncurrent_days = 90    # Delete old versions after 90 days
    }

    noncurrent_version_transition {
      noncurrent_days = 30
      storage_class   = "STANDARD_IA"
    }
  }

  rule {
    id     = "abort-multipart"
    status = "Enabled"

    abort_incomplete_multipart_upload {
      days_after_initiation = 7   # Clean up incomplete uploads
    }
  }
}
```

### S3 Bucket Policy

```hcl
resource "aws_s3_bucket_policy" "main" {
  bucket = aws_s3_bucket.main.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid       = "EnforceTLS"
        Effect    = "Deny"
        Principal = "*"
        Action    = "s3:*"
        Resource = [
          aws_s3_bucket.main.arn,
          "${aws_s3_bucket.main.arn}/*"
        ]
        Condition = {
          Bool = {
            "aws:SecureTransport" = "false"
          }
        }
      },
      {
        Sid       = "EnforceMinimumTLSVersion"
        Effect    = "Deny"
        Principal = "*"
        Action    = "s3:*"
        Resource = [
          aws_s3_bucket.main.arn,
          "${aws_s3_bucket.main.arn}/*"
        ]
        Condition = {
          NumericLessThan = {
            "s3:TlsVersion" = 1.2
          }
        }
      }
    ]
  })
}
```

### S3 Replication Configuration

```hcl
resource "aws_s3_bucket" "replica" {
  provider = aws.replica
  bucket   = "${var.project}-replica-${data.aws_caller_identity.current.account_id}"
}

resource "aws_s3_bucket_versioning" "replica" {
  provider = aws.replica
  bucket   = aws_s3_bucket.replica.id
  versioning_configuration {
    status = "Enabled"   # Versioning required on both source and destination
  }
}

resource "aws_s3_bucket_replication_configuration" "main" {
  bucket = aws_s3_bucket.main.id
  role   = aws_iam_role.s3_replication.arn

  rule {
    id     = "replicate-all"
    status = "Enabled"

    filter {}   # Empty filter replicates all objects

    destination {
      bucket        = aws_s3_bucket.replica.arn
      storage_class = "STANDARD_IA"   # Save cost on replica

      encryption_configuration {
        replica_kms_key_id = aws_kms_key.s3_replica.arn
      }
    }

    delete_marker_replication {
      status = "Enabled"
    }
  }

  depends_on = [aws_s3_bucket_versioning.main]
}
```

### S3 Notification Configuration (to SQS, SNS, Lambda)

```hcl
resource "aws_s3_bucket_notification" "main" {
  bucket = aws_s3_bucket.main.id

  lambda_function {
    lambda_function_arn = aws_lambda_function.processor.arn
    events              = ["s3:ObjectCreated:*"]
    filter_prefix       = "uploads/"
    filter_suffix       = ".csv"
  }

  queue {
    queue_arn     = aws_sqs_queue.s3_events.arn
    events        = ["s3:ObjectCreated:*"]
    filter_prefix = "data/"
  }

  topic {
    topic_arn     = aws_sns_topic.s3_events.arn
    events        = ["s3:ObjectRemoved:*"]
  }
}

resource "aws_lambda_permission" "s3_invoke" {
  statement_id  = "AllowS3Invoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.processor.function_name
  principal     = "s3.amazonaws.com"
  source_arn    = aws_s3_bucket.main.arn
}
```

---

## CloudFront

### CloudFront Distribution with S3 Origin

```hcl
resource "aws_cloudfront_distribution" "main" {
  enabled             = true
  is_ipv6_enabled     = true
  default_root_object = "index.html"
  aliases             = [var.domain_name]
  price_class         = "PriceClass_100"   # US, Canada, Europe only — cheapest
  comment             = "${var.project} CDN"

  origin {
    domain_name              = aws_s3_bucket.static.bucket_regional_domain_name
    origin_id                = "s3-static"
    origin_access_control_id = aws_cloudfront_origin_access_control.s3.id
  }

  default_cache_behavior {
    allowed_methods        = ["GET", "HEAD", "OPTIONS"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = "s3-static"
    viewer_protocol_policy = "redirect-to-https"
    compress               = true    # Enable automatic gzip/brotli compression

    cache_policy_id            = aws_cloudfront_cache_policy.static.id
    origin_request_policy_id   = data.aws_cloudfront_origin_request_policy.cors_s3.id
    response_headers_policy_id = aws_cloudfront_response_headers_policy.security.id
  }

  # Custom error response — serve SPA index.html for 404s
  custom_error_response {
    error_code            = 404
    response_code         = 200
    response_page_path    = "/index.html"
    error_caching_min_ttl = 10
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    acm_certificate_arn      = aws_acm_certificate.main.arn
    ssl_support_method       = "sni-only"
    minimum_protocol_version = "TLSv1.2_2021"
  }

  tags = {
    Name = "${var.project}-cdn"
  }
}
```

### Origin Access Control (OAC)

```hcl
resource "aws_cloudfront_origin_access_control" "s3" {
  name                              = "${var.project}-s3-oac"
  description                       = "OAC for S3 static bucket"
  origin_access_control_origin_type = "s3"
  signing_behavior                  = "always"
  signing_protocol                  = "sigv4"
}

# S3 bucket policy to allow CloudFront OAC access
resource "aws_s3_bucket_policy" "cloudfront" {
  bucket = aws_s3_bucket.static.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid       = "AllowCloudFrontOAC"
        Effect    = "Allow"
        Principal = { Service = "cloudfront.amazonaws.com" }
        Action    = "s3:GetObject"
        Resource  = "${aws_s3_bucket.static.arn}/*"
        Condition = {
          StringEquals = {
            "AWS:SourceArn" = aws_cloudfront_distribution.main.arn
          }
        }
      }
    ]
  })
}
```

### CloudFront with ALB Origin

```hcl
resource "aws_cloudfront_distribution" "api" {
  enabled         = true
  is_ipv6_enabled = true
  aliases         = ["api.${var.domain_name}"]
  comment         = "${var.project} API CDN"

  origin {
    domain_name = aws_lb.main.dns_name
    origin_id   = "alb-api"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "https-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }

    # Custom header to verify requests come through CloudFront
    custom_header {
      name  = "X-Origin-Verify"
      value = var.cloudfront_shared_secret
    }
  }

  default_cache_behavior {
    allowed_methods        = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = "alb-api"
    viewer_protocol_policy = "redirect-to-https"

    # Disable caching for API — forward everything to origin
    cache_policy_id          = data.aws_cloudfront_cache_policy.disabled.id
    origin_request_policy_id = data.aws_cloudfront_origin_request_policy.all_viewer.id
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    acm_certificate_arn      = aws_acm_certificate.main.arn
    ssl_support_method       = "sni-only"
    minimum_protocol_version = "TLSv1.2_2021"
  }
}
```

### Custom Domain with ACM Certificate

```hcl
# ACM certificate must be in us-east-1 for CloudFront
resource "aws_acm_certificate" "main" {
  provider          = aws.us_east_1
  domain_name       = var.domain_name
  subject_alternative_names = ["*.${var.domain_name}"]
  validation_method = "DNS"

  lifecycle {
    create_before_destroy = true   # Avoid downtime during certificate renewal
  }

  tags = {
    Name = "${var.project}-cert"
  }
}

resource "aws_route53_record" "cert_validation" {
  for_each = {
    for dvo in aws_acm_certificate.main.domain_validation_options : dvo.domain_name => {
      name   = dvo.resource_record_name
      type   = dvo.resource_record_type
      record = dvo.resource_record_value
    }
  }

  zone_id = aws_route53_zone.main.zone_id
  name    = each.value.name
  type    = each.value.type
  records = [each.value.record]
  ttl     = 60
}

resource "aws_acm_certificate_validation" "main" {
  provider                = aws.us_east_1
  certificate_arn         = aws_acm_certificate.main.arn
  validation_record_fqdns = [for record in aws_route53_record.cert_validation : record.fqdn]
}
```

---

## Route53

### Hosted Zone and A Record (Alias to ALB, CloudFront)

```hcl
resource "aws_route53_zone" "main" {
  name    = var.domain_name
  comment = "${var.project} public hosted zone"
}

# Alias to CloudFront distribution — no TTL needed for alias records
resource "aws_route53_record" "root" {
  zone_id = aws_route53_zone.main.zone_id
  name    = var.domain_name
  type    = "A"

  alias {
    name                   = aws_cloudfront_distribution.main.domain_name
    zone_id                = aws_cloudfront_distribution.main.hosted_zone_id
    evaluate_target_health = false   # CloudFront handles health internally
  }
}

# Alias to ALB
resource "aws_route53_record" "api" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "api.${var.domain_name}"
  type    = "A"

  alias {
    name                   = aws_lb.main.dns_name
    zone_id                = aws_lb.main.zone_id
    evaluate_target_health = true   # Route53 checks ALB health
  }
}
```

### Health Checks and Failover Routing

```hcl
resource "aws_route53_health_check" "primary" {
  fqdn              = "api.${var.domain_name}"
  port               = 443
  type               = "HTTPS"
  resource_path      = "/health"
  request_interval   = 30
  failure_threshold  = 3
  enable_sni         = true
  regions            = ["us-east-1", "eu-west-1", "ap-southeast-1"]

  tags = {
    Name = "${var.project}-primary-health"
  }
}

# Failover routing — primary record
resource "aws_route53_record" "failover_primary" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "app.${var.domain_name}"
  type    = "A"

  failover_routing_policy {
    type = "PRIMARY"
  }

  alias {
    name                   = aws_lb.primary.dns_name
    zone_id                = aws_lb.primary.zone_id
    evaluate_target_health = true
  }

  set_identifier  = "primary"
  health_check_id = aws_route53_health_check.primary.id
}

# Failover routing — secondary record
resource "aws_route53_record" "failover_secondary" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "app.${var.domain_name}"
  type    = "A"

  failover_routing_policy {
    type = "SECONDARY"
  }

  alias {
    name                   = aws_s3_bucket_website_configuration.failover.website_domain
    zone_id                = aws_s3_bucket.failover.hosted_zone_id
    evaluate_target_health = false
  }

  set_identifier = "secondary"
}
```

### Weighted Routing

```hcl
resource "aws_route53_record" "weighted_blue" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "api.${var.domain_name}"
  type    = "A"

  weighted_routing_policy {
    weight = 90    # 90% of traffic to blue (stable)
  }

  alias {
    name                   = aws_lb.blue.dns_name
    zone_id                = aws_lb.blue.zone_id
    evaluate_target_health = true
  }

  set_identifier = "blue"
}

resource "aws_route53_record" "weighted_green" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "api.${var.domain_name}"
  type    = "A"

  weighted_routing_policy {
    weight = 10    # 10% of traffic to green (canary)
  }

  alias {
    name                   = aws_lb.green.dns_name
    zone_id                = aws_lb.green.zone_id
    evaluate_target_health = true
  }

  set_identifier = "green"
}
```

---

## ALB (Application Load Balancer)

### ALB with HTTPS Listener and Target Groups

```hcl
resource "aws_lb" "main" {
  name               = "${var.project}-alb"
  internal           = false
  load_balancer_type = "application"
  security_groups    = [aws_security_group.alb.id]
  subnets            = aws_subnet.public[*].id

  enable_deletion_protection = true
  drop_invalid_header_fields = true   # Protect against HTTP request smuggling

  access_logs {
    bucket  = aws_s3_bucket.alb_logs.id
    prefix  = "alb"
    enabled = true
  }

  tags = {
    Name = "${var.project}-alb"
  }
}

resource "aws_security_group" "alb" {
  name_prefix = "${var.project}-alb-"
  vpc_id      = aws_vpc.main.id

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "${var.project}-alb-sg"
  }
}

# Redirect HTTP to HTTPS
resource "aws_lb_listener" "http" {
  load_balancer_arn = aws_lb.main.arn
  port              = 80
  protocol          = "HTTP"

  default_action {
    type = "redirect"
    redirect {
      port        = "443"
      protocol    = "HTTPS"
      status_code = "HTTP_301"
    }
  }
}

resource "aws_lb_listener" "https" {
  load_balancer_arn = aws_lb.main.arn
  port              = 443
  protocol          = "HTTPS"
  ssl_policy        = "ELBSecurityPolicy-TLS13-1-2-2021-06"   # TLS 1.3 preferred
  certificate_arn   = aws_acm_certificate.regional.arn

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.app.arn
  }
}

resource "aws_lb_target_group" "app" {
  name        = "${var.project}-app-tg"
  port        = 8080
  protocol    = "HTTP"
  vpc_id      = aws_vpc.main.id
  target_type = "ip"    # Required for Fargate

  health_check {
    enabled             = true
    path                = "/health"
    port                = "traffic-port"
    healthy_threshold   = 2
    unhealthy_threshold = 3
    timeout             = 5
    interval            = 30
    matcher             = "200"
  }

  deregistration_delay = 30   # Faster draining for quicker deployments

  tags = {
    Name = "${var.project}-app-tg"
  }
}
```

### Path-Based and Host-Based Routing Rules

```hcl
# Path-based routing — send /api/* to API target group
resource "aws_lb_listener_rule" "api_path" {
  listener_arn = aws_lb_listener.https.arn
  priority     = 100

  action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.api.arn
  }

  condition {
    path_pattern {
      values = ["/api/*", "/graphql"]
    }
  }
}

# Host-based routing — send admin.example.com to admin target group
resource "aws_lb_listener_rule" "admin_host" {
  listener_arn = aws_lb_listener.https.arn
  priority     = 50

  action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.admin.arn
  }

  condition {
    host_header {
      values = ["admin.${var.domain_name}"]
    }
  }
}

# Fixed response — maintenance page
resource "aws_lb_listener_rule" "maintenance" {
  listener_arn = aws_lb_listener.https.arn
  priority     = 1     # Highest priority — uncomment condition to activate

  action {
    type = "fixed-response"
    fixed_response {
      content_type = "text/html"
      message_body = "<h1>Under Maintenance</h1><p>We will be back shortly.</p>"
      status_code  = "503"
    }
  }

  # Condition matches a header that is never sent — effectively disabled
  # Change to path_pattern { values = ["/*"] } to activate maintenance mode
  condition {
    http_header {
      http_header_name = "X-Maintenance-Mode"
      values           = ["enabled"]
    }
  }
}
```

### WAF Association

```hcl
resource "aws_wafv2_web_acl_association" "alb" {
  resource_arn = aws_lb.main.arn
  web_acl_arn  = aws_wafv2_web_acl.main.arn
}

resource "aws_wafv2_web_acl" "main" {
  name  = "${var.project}-waf"
  scope = "REGIONAL"   # Use CLOUDFRONT for CloudFront distributions

  default_action {
    allow {}
  }

  # AWS Managed Rules — Core Rule Set
  rule {
    name     = "aws-managed-common"
    priority = 1

    override_action {
      none {}   # Use rule group actions as-is
    }

    statement {
      managed_rule_group_statement {
        name        = "AWSManagedRulesCommonRuleSet"
        vendor_name = "AWS"
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = true
      metric_name                = "aws-common-rules"
      sampled_requests_enabled   = true
    }
  }

  # Rate limiting — 2000 requests per 5 minutes per IP
  rule {
    name     = "rate-limit"
    priority = 2

    action {
      block {}
    }

    statement {
      rate_based_statement {
        limit              = 2000
        aggregate_key_type = "IP"
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = true
      metric_name                = "rate-limit"
      sampled_requests_enabled   = true
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = true
    metric_name                = "${var.project}-waf"
    sampled_requests_enabled   = true
  }
}
```

---

## IAM Essentials

### IAM Role with Trust Policy

```hcl
resource "aws_iam_role" "ecs_execution" {
  name = "${var.project}-ecs-execution"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect    = "Allow"
        Principal = { Service = "ecs-tasks.amazonaws.com" }
        Action    = "sts:AssumeRole"
      }
    ]
  })

  tags = {
    Name = "${var.project}-ecs-execution"
  }
}

# Attach the AWS managed policy for ECS task execution
resource "aws_iam_role_policy_attachment" "ecs_execution" {
  role       = aws_iam_role.ecs_execution.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

# Additional policy for pulling secrets from SSM and Secrets Manager
resource "aws_iam_role_policy" "ecs_execution_secrets" {
  name = "secrets-access"
  role = aws_iam_role.ecs_execution.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "ssm:GetParameters",
          "secretsmanager:GetSecretValue"
        ]
        Resource = [
          "arn:aws:ssm:${var.region}:${data.aws_caller_identity.current.account_id}:parameter/${var.project}/*",
          "arn:aws:secretsmanager:${var.region}:${data.aws_caller_identity.current.account_id}:secret:${var.project}/*"
        ]
      },
      {
        Effect   = "Allow"
        Action   = "kms:Decrypt"
        Resource = aws_kms_key.secrets.arn
      }
    ]
  })
}
```

### IAM Role for Lambda

```hcl
resource "aws_iam_role" "lambda" {
  name = "${var.project}-lambda"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect    = "Allow"
        Principal = { Service = "lambda.amazonaws.com" }
        Action    = "sts:AssumeRole"
      }
    ]
  })
}

# Managed policy for basic execution (CloudWatch Logs)
resource "aws_iam_role_policy_attachment" "lambda_basic" {
  role       = aws_iam_role.lambda.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

# Managed policy for VPC access (ENI management)
resource "aws_iam_role_policy_attachment" "lambda_vpc" {
  role       = aws_iam_role.lambda.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaVPCAccessExecutionRole"
}

# Custom policy scoped to specific resources
resource "aws_iam_role_policy" "lambda_app" {
  name = "app-permissions"
  role = aws_iam_role.lambda.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect   = "Allow"
        Action   = ["dynamodb:GetItem", "dynamodb:PutItem", "dynamodb:Query"]
        Resource = aws_dynamodb_table.main.arn
      },
      {
        Effect   = "Allow"
        Action   = ["sqs:SendMessage"]
        Resource = aws_sqs_queue.events.arn
      }
    ]
  })
}
```

### OIDC Provider for GitHub Actions

```hcl
data "tls_certificate" "github" {
  url = "https://token.actions.githubusercontent.com/.well-known/openid-configuration"
}

resource "aws_iam_openid_connect_provider" "github" {
  url             = "https://token.actions.githubusercontent.com"
  client_id_list  = ["sts.amazonaws.com"]
  thumbprint_list = [data.tls_certificate.github.certificates[0].sha1_fingerprint]
}

resource "aws_iam_role" "github_actions" {
  name = "${var.project}-github-actions"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect    = "Allow"
        Principal = { Federated = aws_iam_openid_connect_provider.github.arn }
        Action    = "sts:AssumeRoleWithWebIdentity"
        Condition = {
          StringEquals = {
            "token.actions.githubusercontent.com:aud" = "sts.amazonaws.com"
          }
          StringLike = {
            # Restrict to specific repo and branch
            "token.actions.githubusercontent.com:sub" = "repo:${var.github_org}/${var.github_repo}:ref:refs/heads/main"
          }
        }
      }
    ]
  })

  tags = {
    Name = "${var.project}-github-actions"
  }
}

# Grant GitHub Actions permission to push to ECR and update ECS
resource "aws_iam_role_policy" "github_actions_deploy" {
  name = "deploy-permissions"
  role = aws_iam_role.github_actions.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "ecr:GetAuthorizationToken",
          "ecr:BatchCheckLayerAvailability",
          "ecr:PutImage",
          "ecr:InitiateLayerUpload",
          "ecr:UploadLayerPart",
          "ecr:CompleteLayerUpload"
        ]
        Resource = "*"
      },
      {
        Effect = "Allow"
        Action = [
          "ecs:UpdateService",
          "ecs:DescribeServices",
          "ecs:DescribeTaskDefinition",
          "ecs:RegisterTaskDefinition"
        ]
        Resource = "*"
      },
      {
        Effect   = "Allow"
        Action   = "iam:PassRole"
        Resource = [
          aws_iam_role.ecs_execution.arn,
          aws_iam_role.ecs_task.arn
        ]
      }
    ]
  })
}
```

---

## SQS & SNS

### SQS Queue with Dead Letter Queue

```hcl
resource "aws_sqs_queue" "events" {
  name                       = "${var.project}-events"
  visibility_timeout_seconds = 60     # Must be >= Lambda timeout
  message_retention_seconds  = 1209600  # 14 days
  receive_wait_time_seconds  = 20     # Long polling — reduces empty receives and cost
  max_message_size           = 262144   # 256 KB

  # Server-side encryption with SQS-managed key
  sqs_managed_sse_enabled = true

  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.events_dlq.arn
    maxReceiveCount     = 3    # Move to DLQ after 3 failed processing attempts
  })

  tags = {
    Name = "${var.project}-events"
  }
}

resource "aws_sqs_queue" "events_dlq" {
  name                      = "${var.project}-events-dlq"
  message_retention_seconds = 1209600   # Keep DLQ messages for 14 days
  sqs_managed_sse_enabled   = true

  tags = {
    Name = "${var.project}-events-dlq"
  }
}

# Redrive allow policy — only the source queue can use this DLQ
resource "aws_sqs_queue_redrive_allow_policy" "events_dlq" {
  queue_url = aws_sqs_queue.events_dlq.id

  redrive_allow_policy = jsonencode({
    redrivePermission = "byQueue"
    sourceQueueArns   = [aws_sqs_queue.events.arn]
  })
}
```

### SNS Topic with Subscriptions

```hcl
resource "aws_sns_topic" "alerts" {
  name              = "${var.project}-alerts"
  kms_master_key_id = "alias/aws/sns"   # Encrypt at rest

  tags = {
    Name = "${var.project}-alerts"
  }
}

resource "aws_sns_topic_subscription" "email" {
  topic_arn = aws_sns_topic.alerts.arn
  protocol  = "email"
  endpoint  = var.alert_email
}

resource "aws_sns_topic_subscription" "sqs" {
  topic_arn            = aws_sns_topic.alerts.arn
  protocol             = "sqs"
  endpoint             = aws_sqs_queue.alert_processing.arn
  raw_message_delivery = true    # Deliver raw message without SNS envelope
}

resource "aws_sns_topic_subscription" "lambda" {
  topic_arn = aws_sns_topic.alerts.arn
  protocol  = "lambda"
  endpoint  = aws_lambda_function.alert_handler.arn
}
```

### SQS Queue Policy

```hcl
resource "aws_sqs_queue_policy" "s3_events" {
  queue_url = aws_sqs_queue.s3_events.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid       = "AllowS3SendMessage"
        Effect    = "Allow"
        Principal = { Service = "s3.amazonaws.com" }
        Action    = "sqs:SendMessage"
        Resource  = aws_sqs_queue.s3_events.arn
        Condition = {
          ArnEquals = {
            "aws:SourceArn" = aws_s3_bucket.main.arn
          }
        }
      },
      {
        Sid       = "AllowSNSPublish"
        Effect    = "Allow"
        Principal = { Service = "sns.amazonaws.com" }
        Action    = "sqs:SendMessage"
        Resource  = aws_sqs_queue.s3_events.arn
        Condition = {
          ArnEquals = {
            "aws:SourceArn" = aws_sns_topic.alerts.arn
          }
        }
      }
    ]
  })
}
```

---

## Monitoring

### CloudWatch Log Group

```hcl
resource "aws_cloudwatch_log_group" "application" {
  name              = "/app/${var.project}"
  retention_in_days = 90     # Balance cost vs. compliance requirements
  kms_key_id        = aws_kms_key.logs.arn

  tags = {
    Name = "${var.project}-app-logs"
  }
}

# Metric filter — extract error count from application logs
resource "aws_cloudwatch_log_metric_filter" "errors" {
  name           = "${var.project}-error-count"
  log_group_name = aws_cloudwatch_log_group.application.name
  pattern        = "{ $.level = \"ERROR\" }"   # JSON structured logs

  metric_transformation {
    name          = "ErrorCount"
    namespace     = "${var.project}/Application"
    value         = "1"
    default_value = "0"
  }
}
```

### CloudWatch Metric Alarm

```hcl
# ECS CPU alarm
resource "aws_cloudwatch_metric_alarm" "ecs_cpu_high" {
  alarm_name          = "${var.project}-ecs-cpu-high"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 3
  metric_name         = "CPUUtilization"
  namespace           = "AWS/ECS"
  period              = 60
  statistic           = "Average"
  threshold           = 85
  alarm_description   = "ECS CPU utilization above 85% for 3 consecutive minutes"
  alarm_actions       = [aws_sns_topic.alerts.arn]
  ok_actions          = [aws_sns_topic.alerts.arn]

  dimensions = {
    ClusterName = aws_ecs_cluster.main.name
    ServiceName = aws_ecs_service.app.name
  }

  tags = {
    Name = "${var.project}-ecs-cpu-high"
  }
}

# ALB 5xx error rate alarm
resource "aws_cloudwatch_metric_alarm" "alb_5xx" {
  alarm_name          = "${var.project}-alb-5xx"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 2
  threshold           = 10
  alarm_description   = "ALB returning more than 10 5xx errors per minute"
  alarm_actions       = [aws_sns_topic.alerts.arn]

  metric_query {
    id          = "e1"
    expression  = "m1"
    label       = "5xx Error Count"
    return_data = true
  }

  metric_query {
    id = "m1"
    metric {
      metric_name = "HTTPCode_ELB_5XX_Count"
      namespace   = "AWS/ApplicationELB"
      period      = 60
      stat        = "Sum"
      dimensions = {
        LoadBalancer = aws_lb.main.arn_suffix
      }
    }
  }

  tags = {
    Name = "${var.project}-alb-5xx"
  }
}

# Application error rate from log metric filter
resource "aws_cloudwatch_metric_alarm" "app_errors" {
  alarm_name          = "${var.project}-app-errors"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "ErrorCount"
  namespace           = "${var.project}/Application"
  period              = 300
  statistic           = "Sum"
  threshold           = 50
  alarm_description   = "More than 50 application errors in 5 minutes"
  alarm_actions       = [aws_sns_topic.alerts.arn]
  treat_missing_data  = "notBreaching"   # No data = healthy

  tags = {
    Name = "${var.project}-app-errors"
  }
}
```

### CloudWatch Dashboard

```hcl
resource "aws_cloudwatch_dashboard" "main" {
  dashboard_name = "${var.project}-overview"

  dashboard_body = jsonencode({
    widgets = [
      {
        type   = "metric"
        x      = 0
        y      = 0
        width  = 12
        height = 6
        properties = {
          title   = "ECS CPU & Memory"
          metrics = [
            ["AWS/ECS", "CPUUtilization", "ClusterName", aws_ecs_cluster.main.name,
             "ServiceName", aws_ecs_service.app.name, { stat = "Average" }],
            ["AWS/ECS", "MemoryUtilization", "ClusterName", aws_ecs_cluster.main.name,
             "ServiceName", aws_ecs_service.app.name, { stat = "Average" }]
          ]
          period = 300
          region = var.region
          view   = "timeSeries"
        }
      },
      {
        type   = "metric"
        x      = 12
        y      = 0
        width  = 12
        height = 6
        properties = {
          title   = "ALB Request Count & Latency"
          metrics = [
            ["AWS/ApplicationELB", "RequestCount", "LoadBalancer", aws_lb.main.arn_suffix,
             { stat = "Sum" }],
            ["AWS/ApplicationELB", "TargetResponseTime", "LoadBalancer", aws_lb.main.arn_suffix,
             { stat = "p99", yAxis = "right" }]
          ]
          period = 60
          region = var.region
          view   = "timeSeries"
        }
      },
      {
        type   = "metric"
        x      = 0
        y      = 6
        width  = 12
        height = 6
        properties = {
          title   = "RDS Connections & CPU"
          metrics = [
            ["AWS/RDS", "DatabaseConnections", "DBClusterIdentifier", aws_rds_cluster.main.cluster_identifier,
             { stat = "Average" }],
            ["AWS/RDS", "CPUUtilization", "DBClusterIdentifier", aws_rds_cluster.main.cluster_identifier,
             { stat = "Average", yAxis = "right" }]
          ]
          period = 300
          region = var.region
          view   = "timeSeries"
        }
      },
      {
        type   = "metric"
        x      = 12
        y      = 6
        width  = 12
        height = 6
        properties = {
          title   = "SQS Queue Depth"
          metrics = [
            ["AWS/SQS", "ApproximateNumberOfMessagesVisible", "QueueName", aws_sqs_queue.events.name,
             { stat = "Maximum" }],
            ["AWS/SQS", "ApproximateAgeOfOldestMessage", "QueueName", aws_sqs_queue.events.name,
             { stat = "Maximum", yAxis = "right" }]
          ]
          period = 60
          region = var.region
          view   = "timeSeries"
        }
      }
    ]
  })
}
```

### SNS Alert Topic for Alarms

```hcl
resource "aws_sns_topic" "alarm_notifications" {
  name              = "${var.project}-alarm-notifications"
  kms_master_key_id = "alias/aws/sns"

  tags = {
    Name = "${var.project}-alarm-notifications"
  }
}

# SNS topic policy — allow CloudWatch Alarms to publish
resource "aws_sns_topic_policy" "alarm_notifications" {
  arn = aws_sns_topic.alarm_notifications.arn

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid       = "AllowCloudWatchAlarms"
        Effect    = "Allow"
        Principal = { Service = "cloudwatch.amazonaws.com" }
        Action    = "sns:Publish"
        Resource  = aws_sns_topic.alarm_notifications.arn
        Condition = {
          ArnLike = {
            "aws:SourceArn" = "arn:aws:cloudwatch:${var.region}:${data.aws_caller_identity.current.account_id}:alarm:*"
          }
        }
      }
    ]
  })
}

resource "aws_sns_topic_subscription" "alarm_email" {
  topic_arn = aws_sns_topic.alarm_notifications.arn
  protocol  = "email"
  endpoint  = var.ops_email
}

# Optional: Slack integration via Lambda
resource "aws_sns_topic_subscription" "alarm_slack" {
  topic_arn = aws_sns_topic.alarm_notifications.arn
  protocol  = "lambda"
  endpoint  = aws_lambda_function.slack_notifier.arn
}

resource "aws_lambda_permission" "sns_slack" {
  statement_id  = "AllowSNSInvoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.slack_notifier.function_name
  principal     = "sns.amazonaws.com"
  source_arn    = aws_sns_topic.alarm_notifications.arn
}
```
