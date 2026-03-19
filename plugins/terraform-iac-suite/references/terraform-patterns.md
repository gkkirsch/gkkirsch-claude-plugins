# Terraform Patterns & Best Practices

Comprehensive reference for Terraform infrastructure-as-code patterns, module composition strategies, state management, testing, and idiomatic HCL usage. Every code example in this document uses real, working HCL syntax suitable for production use. This guide assumes Terraform 1.6+ unless otherwise noted.

---

## Module Composition

Modules are the primary mechanism for code reuse in Terraform. Choosing the right composition pattern determines how maintainable and flexible your infrastructure code will be over time.

### Flat Module Pattern

A flat module keeps all resources at a single level with no nested module calls. This is appropriate for small, focused units of infrastructure.

```hcl
# modules/s3-bucket/main.tf
# Flat module: a single-purpose module that manages one logical resource group.

resource "aws_s3_bucket" "this" {
  bucket        = var.bucket_name
  force_destroy = var.force_destroy

  tags = merge(var.tags, {
    ManagedBy = "terraform"
    Module    = "s3-bucket"
  })
}

resource "aws_s3_bucket_versioning" "this" {
  bucket = aws_s3_bucket.this.id

  versioning_configuration {
    status = var.versioning_enabled ? "Enabled" : "Suspended"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "this" {
  bucket = aws_s3_bucket.this.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm     = var.kms_key_arn != null ? "aws:kms" : "AES256"
      kms_master_key_id = var.kms_key_arn
    }
    bucket_key_enabled = var.kms_key_arn != null
  }
}

resource "aws_s3_bucket_public_access_block" "this" {
  bucket = aws_s3_bucket.this.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}
```

```hcl
# modules/s3-bucket/variables.tf

variable "bucket_name" {
  description = "Name of the S3 bucket"
  type        = string

  validation {
    condition     = can(regex("^[a-z0-9][a-z0-9.-]{1,61}[a-z0-9]$", var.bucket_name))
    error_message = "Bucket name must be 3-63 characters, lowercase, and DNS-compliant."
  }
}

variable "versioning_enabled" {
  description = "Enable versioning on the bucket"
  type        = bool
  default     = true
}

variable "kms_key_arn" {
  description = "ARN of the KMS key for server-side encryption. Null uses AES256."
  type        = string
  default     = null
}

variable "force_destroy" {
  description = "Allow bucket to be destroyed even if it contains objects"
  type        = bool
  default     = false
}

variable "tags" {
  description = "Tags to apply to all resources"
  type        = map(string)
  default     = {}
}
```

```hcl
# modules/s3-bucket/outputs.tf

output "bucket_id" {
  description = "The name of the bucket"
  value       = aws_s3_bucket.this.id
}

output "bucket_arn" {
  description = "The ARN of the bucket"
  value       = aws_s3_bucket.this.arn
}

output "bucket_regional_domain_name" {
  description = "The regional domain name of the bucket"
  value       = aws_s3_bucket.this.bucket_regional_domain_name
}
```

### Nested Module Pattern

Nested modules compose smaller modules inside a parent module to build complex infrastructure from reusable parts.

```hcl
# modules/web-app/main.tf
# Nested module: composes child modules to build a complete web application stack.

module "vpc" {
  source = "../vpc"

  cidr_block         = var.vpc_cidr
  availability_zones = var.availability_zones
  environment        = var.environment
  tags               = local.common_tags
}

module "alb" {
  source = "../alb"

  name               = "${var.app_name}-alb"
  vpc_id             = module.vpc.vpc_id
  subnet_ids         = module.vpc.public_subnet_ids
  certificate_arn    = var.certificate_arn
  health_check_path  = var.health_check_path
  tags               = local.common_tags
}

module "ecs_service" {
  source = "../ecs-service"

  cluster_name       = var.cluster_name
  service_name       = var.app_name
  container_image    = var.container_image
  container_port     = var.container_port
  desired_count      = var.desired_count
  subnet_ids         = module.vpc.private_subnet_ids
  security_group_ids = [module.vpc.app_security_group_id]
  target_group_arn   = module.alb.target_group_arn
  tags               = local.common_tags
}

locals {
  common_tags = merge(var.tags, {
    Application = var.app_name
    Environment = var.environment
  })
}
```

### Facade Module Pattern

A facade module wraps a complex third-party or community module, exposing a simplified interface that enforces organizational standards.

```hcl
# modules/rds-standard/main.tf
# Facade: wraps the community RDS module with our org's defaults and guardrails.

module "rds" {
  source  = "terraform-aws-modules/rds/aws"
  version = "~> 6.0"

  identifier = var.db_identifier

  # Engine configuration — expose only what callers need
  engine               = "postgres"
  engine_version       = var.engine_version
  family               = "postgres${split(".", var.engine_version)[0]}"
  major_engine_version = split(".", var.engine_version)[0]
  instance_class       = var.instance_class

  # Storage — enforce encryption, allow size customization
  allocated_storage     = var.allocated_storage
  max_allocated_storage = var.max_allocated_storage
  storage_encrypted     = true  # Always on — not exposed as a variable
  kms_key_id            = var.kms_key_arn

  # Networking
  db_subnet_group_name   = var.db_subnet_group_name
  vpc_security_group_ids = var.security_group_ids
  publicly_accessible    = false  # Never public — not exposed

  # Backup — enforce org minimums
  backup_retention_period = max(var.backup_retention_period, 7)
  backup_window           = "03:00-04:00"
  maintenance_window      = "Mon:04:00-Mon:05:00"

  # Monitoring
  monitoring_interval                   = 60
  monitoring_role_arn                   = var.monitoring_role_arn
  performance_insights_enabled          = true
  performance_insights_retention_period = 7

  # Deletion protection — always on in production
  deletion_protection = var.environment == "production"

  tags = merge(var.tags, {
    DataClassification = var.data_classification
  })
}
```

### Module Composition with Dependency Injection

Pass resources into modules as variables rather than creating them internally. This makes modules composable and testable.

```hcl
# Root module injects dependencies into child modules.

# Create shared resources at the root level
resource "aws_kms_key" "main" {
  description             = "Main encryption key for ${var.project_name}"
  deletion_window_in_days = 30
  enable_key_rotation     = true

  tags = local.tags
}

resource "aws_security_group" "app" {
  name_prefix = "${var.project_name}-app-"
  vpc_id      = module.vpc.vpc_id

  tags = local.tags

  lifecycle {
    create_before_destroy = true
  }
}

# Inject shared resources into modules that need them
module "app_storage" {
  source = "./modules/s3-bucket"

  bucket_name = "${var.project_name}-app-storage"
  kms_key_arn = aws_kms_key.main.arn
  tags        = local.tags
}

module "app_database" {
  source = "./modules/rds-standard"

  db_identifier      = "${var.project_name}-db"
  engine_version     = "15.4"
  instance_class     = "db.r6g.large"
  kms_key_arn        = aws_kms_key.main.arn
  security_group_ids = [aws_security_group.app.id]
  tags               = local.tags
}
```

### Root Module Orchestrating Child Modules (Full Example)

```hcl
# environments/production/main.tf
# Full root module that orchestrates multiple child modules for a production stack.

terraform {
  required_version = ">= 1.6.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.30"
    }
  }

  backend "s3" {
    bucket         = "myorg-terraform-state"
    key            = "production/infrastructure.tfstate"
    region         = "us-east-1"
    dynamodb_table = "terraform-locks"
    encrypt        = true
  }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Environment = "production"
      ManagedBy   = "terraform"
      Project     = var.project_name
    }
  }
}

locals {
  tags = {
    CostCenter = var.cost_center
    Owner      = var.owner_email
  }
}

module "networking" {
  source = "../../modules/networking"

  vpc_cidr           = "10.0.0.0/16"
  availability_zones = ["us-east-1a", "us-east-1b", "us-east-1c"]
  environment        = "production"
  enable_nat_gateway = true
  single_nat_gateway = false  # HA NAT in production
  tags               = local.tags
}

module "compute" {
  source = "../../modules/ecs-cluster"

  cluster_name       = "${var.project_name}-production"
  vpc_id             = module.networking.vpc_id
  private_subnet_ids = module.networking.private_subnet_ids
  tags               = local.tags
}

module "database" {
  source = "../../modules/rds-standard"

  db_identifier         = "${var.project_name}-production"
  engine_version        = "15.4"
  instance_class        = "db.r6g.xlarge"
  allocated_storage     = 100
  max_allocated_storage = 500
  db_subnet_group_name  = module.networking.database_subnet_group_name
  security_group_ids    = [module.networking.database_security_group_id]
  kms_key_arn           = module.encryption.key_arn
  monitoring_role_arn   = module.monitoring.rds_monitoring_role_arn
  environment           = "production"
  data_classification   = "confidential"
  tags                  = local.tags
}

module "encryption" {
  source = "../../modules/kms"

  key_alias   = "alias/${var.project_name}-production"
  environment = "production"
  tags        = local.tags
}

module "monitoring" {
  source = "../../modules/monitoring"

  project_name       = var.project_name
  environment        = "production"
  alarm_email        = var.alarm_email
  enable_dashboards  = true
  tags               = local.tags
}
```

---

## Remote State

Remote state stores the Terraform state file in a shared backend, enabling team collaboration and state locking to prevent concurrent modifications.

### S3 Backend with Locking (Full Config)

```hcl
# backend-bootstrap/main.tf
# Bootstrap the S3 backend resources. Run this ONCE, then migrate state.

resource "aws_s3_bucket" "terraform_state" {
  bucket = "myorg-terraform-state-${data.aws_caller_identity.current.account_id}"

  tags = {
    Purpose   = "Terraform State Storage"
    ManagedBy = "terraform-bootstrap"
  }
}

resource "aws_s3_bucket_versioning" "terraform_state" {
  bucket = aws_s3_bucket.terraform_state.id

  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "terraform_state" {
  bucket = aws_s3_bucket.terraform_state.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "aws:kms"
    }
    bucket_key_enabled = true
  }
}

resource "aws_s3_bucket_public_access_block" "terraform_state" {
  bucket = aws_s3_bucket.terraform_state.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_dynamodb_table" "terraform_locks" {
  name         = "terraform-locks"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "LockID"

  attribute {
    name = "LockID"
    type = "S"
  }

  tags = {
    Purpose   = "Terraform State Locking"
    ManagedBy = "terraform-bootstrap"
  }
}

data "aws_caller_identity" "current" {}
```

### Reading Remote State with terraform_remote_state

```hcl
# Read state from another Terraform configuration.
# The networking stack publishes outputs that the app stack consumes.

data "terraform_remote_state" "networking" {
  backend = "s3"

  config = {
    bucket = "myorg-terraform-state"
    key    = "production/networking.tfstate"
    region = "us-east-1"
  }
}

# Use outputs from the networking state
resource "aws_instance" "app" {
  ami           = data.aws_ami.amazon_linux.id
  instance_type = "t3.medium"
  subnet_id     = data.terraform_remote_state.networking.outputs.private_subnet_ids[0]

  vpc_security_group_ids = [
    data.terraform_remote_state.networking.outputs.app_security_group_id
  ]

  tags = {
    Name = "app-server"
  }
}
```

### Cross-Stack References

```hcl
# Stack A: networking — publishes outputs for consumers
# networking/outputs.tf

output "vpc_id" {
  description = "VPC ID for use by other stacks"
  value       = aws_vpc.main.id
}

output "private_subnet_ids" {
  description = "Private subnet IDs for workloads"
  value       = aws_subnet.private[*].id
}

output "database_subnet_group_name" {
  description = "DB subnet group for RDS instances"
  value       = aws_db_subnet_group.database.name
}

# Stack B: application — consumes networking outputs
# application/data.tf

data "terraform_remote_state" "networking" {
  backend = "s3"
  config = {
    bucket = "myorg-terraform-state"
    key    = "production/networking.tfstate"
    region = "us-east-1"
  }
}

data "terraform_remote_state" "security" {
  backend = "s3"
  config = {
    bucket = "myorg-terraform-state"
    key    = "production/security.tfstate"
    region = "us-east-1"
  }
}

# Reference both stacks
module "app" {
  source = "../../modules/ecs-service"

  subnet_ids         = data.terraform_remote_state.networking.outputs.private_subnet_ids
  security_group_ids = [data.terraform_remote_state.security.outputs.app_sg_id]
}
```

### Remote State vs Data Sources

```hcl
# APPROACH 1: Remote state — pulls from another Terraform configuration
# Best when: You control both stacks and want explicit coupling.

data "terraform_remote_state" "vpc" {
  backend = "s3"
  config = {
    bucket = "myorg-terraform-state"
    key    = "networking/vpc.tfstate"
    region = "us-east-1"
  }
}

locals {
  vpc_id_from_state = data.terraform_remote_state.vpc.outputs.vpc_id
}

# APPROACH 2: Data source — queries the AWS API directly
# Best when: The resource may have been created outside Terraform,
# or you want to decouple from another stack's implementation.

data "aws_vpc" "main" {
  filter {
    name   = "tag:Name"
    values = ["main-vpc"]
  }
}

locals {
  vpc_id_from_data = data.aws_vpc.main.id
}
```

### State Output Design for Consumers

```hcl
# Design outputs with consumers in mind. Group related values and document them.

# Good: typed, documented outputs that form a clear contract
output "vpc" {
  description = "VPC attributes for consuming stacks"
  value = {
    id         = aws_vpc.main.id
    cidr_block = aws_vpc.main.cidr_block
  }
}

output "subnets" {
  description = "Subnet groupings by tier"
  value = {
    public  = { ids = aws_subnet.public[*].id, cidrs = aws_subnet.public[*].cidr_block }
    private = { ids = aws_subnet.private[*].id, cidrs = aws_subnet.private[*].cidr_block }
    data    = { ids = aws_subnet.data[*].id, cidrs = aws_subnet.data[*].cidr_block }
  }
}

output "security_groups" {
  description = "Pre-built security groups by role"
  value = {
    alb_sg = aws_security_group.alb.id
    app_sg = aws_security_group.app.id
    db_sg  = aws_security_group.database.id
  }
}
```

---

## Data Sources

Data sources let Terraform query existing infrastructure and external systems without managing their lifecycle.

### AWS Data Sources

```hcl
# Fetch the latest Amazon Linux 2023 AMI
data "aws_ami" "amazon_linux" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["al2023-ami-2023.*-x86_64"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }

  filter {
    name   = "root-device-type"
    values = ["ebs"]
  }
}

# Look up the default VPC
data "aws_vpc" "default" {
  default = true
}

# Find all subnets in a specific VPC
data "aws_subnets" "private" {
  filter {
    name   = "vpc-id"
    values = [data.aws_vpc.default.id]
  }

  tags = {
    Tier = "private"
  }
}

# Get current AWS account and region info
data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

# Use identity and region in resource configuration
resource "aws_s3_bucket" "logs" {
  bucket = "logs-${data.aws_caller_identity.current.account_id}-${data.aws_region.current.name}"
}
```

### Data Source Filtering Patterns

```hcl
# Multiple filter criteria narrow results precisely
data "aws_ami" "ubuntu" {
  most_recent = true
  owners      = ["099720109477"]  # Canonical

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"]
  }

  filter {
    name   = "state"
    values = ["available"]
  }
}

# Filter subnets by multiple tags
data "aws_subnets" "app_tier" {
  filter {
    name   = "vpc-id"
    values = [var.vpc_id]
  }

  filter {
    name   = "availability-zone"
    values = ["us-east-1a", "us-east-1b"]
  }

  tags = {
    Tier        = "application"
    Environment = var.environment
  }
}

# Look up security groups by name pattern
data "aws_security_groups" "web" {
  filter {
    name   = "group-name"
    values = ["web-*"]
  }

  filter {
    name   = "vpc-id"
    values = [var.vpc_id]
  }
}
```

### Using Data Sources for Dynamic Configuration

```hcl
# Dynamically spread instances across all available AZs
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_subnet" "private" {
  count = length(data.aws_availability_zones.available.names)

  vpc_id            = aws_vpc.main.id
  cidr_block        = cidrsubnet(aws_vpc.main.cidr_block, 8, count.index)
  availability_zone = data.aws_availability_zones.available.names[count.index]

  tags = {
    Name = "private-${data.aws_availability_zones.available.names[count.index]}"
    Tier = "private"
  }
}
```

### External Data Source

```hcl
# Call an external script and read its JSON output.
# The script MUST return a JSON object with string values.

data "external" "git_info" {
  program = ["bash", "-c", <<-EOF
    echo '{"commit_sha": "'$(git rev-parse --short HEAD)'", "branch": "'$(git rev-parse --abbrev-ref HEAD)'"}'
  EOF
  ]

  working_dir = path.module
}

resource "aws_instance" "app" {
  ami           = data.aws_ami.amazon_linux.id
  instance_type = "t3.micro"

  tags = {
    GitCommit = data.external.git_info.result.commit_sha
    GitBranch = data.external.git_info.result.branch
  }
}
```

### HTTP Data Source

```hcl
# Fetch data from an HTTP endpoint
data "http" "my_ip" {
  url = "https://checkip.amazonaws.com"
}

# Use the response to create a security group rule
resource "aws_security_group_rule" "allow_my_ip" {
  type              = "ingress"
  from_port         = 22
  to_port           = 22
  protocol          = "tcp"
  cidr_blocks       = ["${trimspace(data.http.my_ip.response_body)}/32"]
  security_group_id = aws_security_group.bastion.id
  description       = "SSH from operator IP"
}
```

---

## Provisioners (and Why to Avoid Them)

Provisioners are a last resort. They break Terraform's declarative model, don't appear in plan output, and have no rollback mechanism. Prefer native resource arguments, `user_data`, cloud-init, or image-baking with Packer.

### local-exec Provisioner

```hcl
# Runs a command on the machine executing Terraform.
# Use case: triggering a CI/CD pipeline after infrastructure is ready.

resource "aws_eks_cluster" "main" {
  name     = var.cluster_name
  role_arn = aws_iam_role.eks.arn
  version  = "1.28"

  vpc_config {
    subnet_ids = var.subnet_ids
  }

  # Update kubeconfig after cluster creation
  provisioner "local-exec" {
    command = "aws eks update-kubeconfig --name ${self.name} --region ${var.region}"
  }
}
```

### remote-exec Provisioner

```hcl
# Runs commands on the remote resource via SSH or WinRM.
# Almost always better to use user_data or configuration management instead.

resource "aws_instance" "legacy" {
  ami           = data.aws_ami.amazon_linux.id
  instance_type = "t3.micro"
  key_name      = var.key_name

  connection {
    type        = "ssh"
    user        = "ec2-user"
    private_key = file(var.private_key_path)
    host        = self.public_ip
  }

  provisioner "remote-exec" {
    inline = [
      "sudo yum update -y",
      "sudo yum install -y httpd",
      "sudo systemctl start httpd",
      "sudo systemctl enable httpd",
    ]
  }
}
```

### file Provisioner

```hcl
# Copies files or directories to the remote resource.
resource "aws_instance" "web" {
  ami           = data.aws_ami.amazon_linux.id
  instance_type = "t3.micro"
  key_name      = var.key_name

  connection {
    type        = "ssh"
    user        = "ec2-user"
    private_key = file(var.private_key_path)
    host        = self.public_ip
  }

  provisioner "file" {
    source      = "${path.module}/scripts/setup.sh"
    destination = "/tmp/setup.sh"
  }

  provisioner "remote-exec" {
    inline = [
      "chmod +x /tmp/setup.sh",
      "sudo /tmp/setup.sh",
    ]
  }
}
```

### null_resource with Triggers

```hcl
# null_resource runs provisioners without managing a real resource.
# triggers controls when the provisioners re-run.

resource "null_resource" "db_migration" {
  triggers = {
    # Re-run migration whenever the schema version changes
    schema_version = var.schema_version
    db_endpoint    = aws_db_instance.main.endpoint
  }

  provisioner "local-exec" {
    command = <<-EOF
      DATABASE_URL="postgresql://${var.db_user}:${var.db_password}@${aws_db_instance.main.endpoint}/${var.db_name}"
      flyway -url="jdbc:${DATABASE_URL}" migrate
    EOF

    environment = {
      FLYWAY_SCHEMAS = var.db_name
    }
  }

  depends_on = [aws_db_instance.main]
}
```

### Better Alternatives (user_data, cloud-init, Packer)

```hcl
# PREFERRED: Use user_data instead of remote-exec provisioners.
# user_data runs at instance boot and is part of the resource declaration.

resource "aws_instance" "web" {
  ami           = data.aws_ami.amazon_linux.id
  instance_type = "t3.micro"

  user_data = base64encode(templatefile("${path.module}/templates/user-data.sh.tftpl", {
    app_version = var.app_version
    db_endpoint = aws_db_instance.main.endpoint
    environment = var.environment
  }))

  user_data_replace_on_change = true

  tags = {
    Name = "${var.app_name}-web"
  }
}

# PREFERRED: Use aws_launch_template with cloud-init for ASG instances.
resource "aws_launch_template" "app" {
  name_prefix   = "${var.app_name}-"
  image_id      = data.aws_ami.amazon_linux.id
  instance_type = var.instance_type

  user_data = base64encode(<<-CLOUDINIT
    #cloud-config
    package_update: true
    packages:
      - docker
      - aws-cli
    runcmd:
      - systemctl start docker
      - aws ecr get-login-password --region ${var.region} | docker login --username AWS --password-stdin ${var.ecr_registry}
      - docker run -d -p 80:8080 ${var.container_image}
  CLOUDINIT
  )

  tag_specifications {
    resource_type = "instance"
    tags = {
      Name = "${var.app_name}-instance"
    }
  }
}
```

---

## Dynamic Blocks

Dynamic blocks generate repeated nested blocks from collections, reducing repetition while keeping configurations readable.

### Dynamic Block Syntax

```hcl
# Basic dynamic block structure
resource "aws_security_group" "example" {
  name        = "dynamic-sg"
  description = "Security group with dynamic rules"
  vpc_id      = var.vpc_id

  dynamic "ingress" {
    for_each = var.ingress_rules
    content {
      description     = ingress.value.description
      from_port       = ingress.value.from_port
      to_port         = ingress.value.to_port
      protocol        = ingress.value.protocol
      cidr_blocks     = ingress.value.cidr_blocks
      security_groups = ingress.value.security_groups
    }
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
    description = "Allow all outbound"
  }
}
```

### Nested Dynamic Blocks

```hcl
# Dynamic blocks can be nested for complex resource structures.
resource "aws_lb_listener" "https" {
  load_balancer_arn = aws_lb.main.arn
  port              = 443
  protocol          = "HTTPS"
  ssl_policy        = "ELBSecurityPolicy-TLS13-1-2-2021-06"
  certificate_arn   = var.default_certificate_arn

  default_action {
    type = "fixed-response"
    fixed_response {
      content_type = "text/plain"
      message_body = "Not Found"
      status_code  = "404"
    }
  }
}

resource "aws_lb_listener_rule" "services" {
  for_each = var.service_routes

  listener_arn = aws_lb_listener.https.arn
  priority     = each.value.priority

  action {
    type             = "forward"
    target_group_arn = each.value.target_group_arn
  }

  dynamic "condition" {
    for_each = each.value.host_headers != null ? [each.value.host_headers] : []
    content {
      host_header {
        values = condition.value
      }
    }
  }

  dynamic "condition" {
    for_each = each.value.path_patterns != null ? [each.value.path_patterns] : []
    content {
      path_pattern {
        values = condition.value
      }
    }
  }
}
```

### Dynamic Security Group Rules

```hcl
variable "service_ports" {
  description = "Map of service names to their port configurations"
  type = map(object({
    port        = number
    protocol    = string
    cidr_blocks = list(string)
    description = string
  }))
  default = {
    http = {
      port        = 80
      protocol    = "tcp"
      cidr_blocks = ["0.0.0.0/0"]
      description = "HTTP from anywhere"
    }
    https = {
      port        = 443
      protocol    = "tcp"
      cidr_blocks = ["0.0.0.0/0"]
      description = "HTTPS from anywhere"
    }
    ssh = {
      port        = 22
      protocol    = "tcp"
      cidr_blocks = ["10.0.0.0/8"]
      description = "SSH from internal network"
    }
  }
}

resource "aws_security_group" "service" {
  name_prefix = "service-"
  vpc_id      = var.vpc_id

  dynamic "ingress" {
    for_each = var.service_ports
    content {
      description = ingress.value.description
      from_port   = ingress.value.port
      to_port     = ingress.value.port
      protocol    = ingress.value.protocol
      cidr_blocks = ingress.value.cidr_blocks
    }
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}
```

### Dynamic IAM Policy Statements

```hcl
variable "s3_permissions" {
  description = "Per-bucket IAM permissions"
  type = map(object({
    actions   = list(string)
    bucket_id = string
    prefix    = string
  }))
}

data "aws_iam_policy_document" "s3_access" {
  dynamic "statement" {
    for_each = var.s3_permissions
    content {
      sid    = "Access${replace(title(statement.key), "-", "")}"
      effect = "Allow"
      actions = statement.value.actions

      resources = [
        "arn:aws:s3:::${statement.value.bucket_id}",
        "arn:aws:s3:::${statement.value.bucket_id}/${statement.value.prefix}*",
      ]
    }
  }
}

resource "aws_iam_policy" "s3_access" {
  name   = "s3-granular-access"
  policy = data.aws_iam_policy_document.s3_access.json
}
```

### Dynamic Tags

```hcl
# Dynamically generate tag blocks for resources that use the tag sub-block
# syntax instead of a tags map (e.g., aws_autoscaling_group).

locals {
  asg_tags = merge(var.tags, {
    Name        = "${var.app_name}-asg"
    Environment = var.environment
  })
}

resource "aws_autoscaling_group" "app" {
  name                = "${var.app_name}-asg"
  desired_capacity    = var.desired_capacity
  max_size            = var.max_size
  min_size            = var.min_size
  vpc_zone_identifier = var.subnet_ids

  launch_template {
    id      = aws_launch_template.app.id
    version = "$Latest"
  }

  dynamic "tag" {
    for_each = local.asg_tags
    content {
      key                 = tag.key
      value               = tag.value
      propagate_at_launch = true
    }
  }
}
```

### When to Use vs When to Avoid

```hcl
# GOOD use of dynamic blocks: generating rules from a variable-length collection.
# The number of rules is determined by the caller.

resource "aws_security_group" "good_example" {
  name   = "good-dynamic"
  vpc_id = var.vpc_id

  dynamic "ingress" {
    for_each = var.allowed_ports  # Caller controls the list
    content {
      from_port   = ingress.value
      to_port     = ingress.value
      protocol    = "tcp"
      cidr_blocks = ["10.0.0.0/8"]
    }
  }
}

# BAD use of dynamic blocks: when you have a fixed, known set of blocks.
# Just write them out explicitly — it is clearer and easier to review.

# Instead of dynamic "ingress" { for_each = { http = 80, https = 443 } ... }
# just write:
resource "aws_security_group" "explicit_is_better" {
  name   = "explicit-sg"
  vpc_id = var.vpc_id

  ingress {
    description = "HTTP"
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    description = "HTTPS"
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}
```

---

## for_each vs count

Both `for_each` and `count` create multiple resource instances. They differ in how instances are identified in state, which affects behavior during changes.

### count Basics and Limitations

```hcl
# count creates instances identified by index (0, 1, 2, ...)
# Removing an item from the middle causes all subsequent items to shift.

resource "aws_subnet" "public" {
  count = length(var.availability_zones)

  vpc_id            = aws_vpc.main.id
  cidr_block        = cidrsubnet(aws_vpc.main.cidr_block, 8, count.index)
  availability_zone = var.availability_zones[count.index]

  tags = {
    Name = "public-${var.availability_zones[count.index]}"
  }
}

# count is suitable for conditional creation (0 or 1 instances)
resource "aws_eip" "nat" {
  count  = var.enable_nat_gateway ? 1 : 0
  domain = "vpc"
}
```

### for_each with Sets

```hcl
# for_each with a set of strings — each instance is keyed by the string value.
variable "iam_users" {
  type    = set(string)
  default = ["alice", "bob", "carol"]
}

resource "aws_iam_user" "this" {
  for_each = var.iam_users
  name     = each.value
  path     = "/team/"

  tags = {
    ManagedBy = "terraform"
  }
}

# Reference: aws_iam_user.this["alice"], aws_iam_user.this["bob"], etc.
```

### for_each with Maps

```hcl
# for_each with a map — each.key is the map key, each.value is the map value.
variable "buckets" {
  type = map(object({
    versioning    = bool
    force_destroy = bool
  }))
  default = {
    "app-assets" = { versioning = true, force_destroy = false }
    "app-logs"   = { versioning = false, force_destroy = true }
    "app-backup" = { versioning = true, force_destroy = false }
  }
}

resource "aws_s3_bucket" "this" {
  for_each = var.buckets
  bucket   = "${var.project_name}-${each.key}"

  force_destroy = each.value.force_destroy

  tags = {
    Purpose = each.key
  }
}

resource "aws_s3_bucket_versioning" "this" {
  for_each = var.buckets

  bucket = aws_s3_bucket.this[each.key].id

  versioning_configuration {
    status = each.value.versioning ? "Enabled" : "Suspended"
  }
}
```

### for_each with Complex Objects

```hcl
variable "services" {
  type = map(object({
    container_image = string
    container_port  = number
    cpu             = number
    memory          = number
    desired_count   = number
    environment     = map(string)
  }))
}

resource "aws_ecs_task_definition" "this" {
  for_each = var.services

  family                   = each.key
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = each.value.cpu
  memory                   = each.value.memory

  container_definitions = jsonencode([{
    name      = each.key
    image     = each.value.container_image
    essential = true
    portMappings = [{
      containerPort = each.value.container_port
      protocol      = "tcp"
    }]
    environment = [
      for k, v in each.value.environment : {
        name  = k
        value = v
      }
    ]
  }])
}

resource "aws_ecs_service" "this" {
  for_each = var.services

  name            = each.key
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.this[each.key].arn
  desired_count   = each.value.desired_count
  launch_type     = "FARGATE"

  network_configuration {
    subnets         = var.private_subnet_ids
    security_groups = [aws_security_group.ecs_tasks.id]
  }
}
```

### Migrating from count to for_each

```hcl
# BEFORE: using count — fragile index-based addressing
# resource "aws_iam_user" "team" {
#   count = length(var.team_members)
#   name  = var.team_members[count.index]
# }

# AFTER: using for_each — stable key-based addressing
resource "aws_iam_user" "team" {
  for_each = toset(var.team_members)
  name     = each.value
}

# Migration requires moved blocks to avoid destroy/recreate:
moved {
  from = aws_iam_user.team[0]
  to   = aws_iam_user.team["alice"]
}

moved {
  from = aws_iam_user.team[1]
  to   = aws_iam_user.team["bob"]
}

moved {
  from = aws_iam_user.team[2]
  to   = aws_iam_user.team["carol"]
}
```

### When to Use Each

```hcl
# Use count for:
# 1. Conditional creation (0 or 1 instance)
resource "aws_nat_gateway" "this" {
  count         = var.enable_nat ? 1 : 0
  allocation_id = aws_eip.nat[0].id
  subnet_id     = var.public_subnet_id
}

# 2. Creating N identical copies where order does not matter
resource "aws_cloudwatch_log_group" "shards" {
  count             = var.shard_count
  name              = "/app/shard-${count.index}"
  retention_in_days = 30
}

# Use for_each for:
# 1. Collections where each item is distinct and identifiable
resource "aws_route53_record" "services" {
  for_each = var.service_dns_records

  zone_id = var.zone_id
  name    = each.key
  type    = "A"

  alias {
    name                   = each.value.alb_dns_name
    zone_id                = each.value.alb_zone_id
    evaluate_target_health = true
  }
}
```

---

## Expressions and Functions

Terraform's expression language provides conditionals, iteration, and a rich function library for transforming data.

### Conditional Expressions (Ternary)

```hcl
# Syntax: condition ? true_value : false_value

resource "aws_instance" "app" {
  ami           = data.aws_ami.amazon_linux.id
  instance_type = var.environment == "production" ? "m5.xlarge" : "t3.medium"

  # Conditional association of a public IP
  associate_public_ip_address = var.is_public ? true : false

  # Conditionally attach an IAM role
  iam_instance_profile = var.enable_ssm ? aws_iam_instance_profile.ssm[0].name : null

  # Conditional monitoring
  monitoring = var.environment == "production"
}

# Conditional resource creation pattern
resource "aws_iam_instance_profile" "ssm" {
  count = var.enable_ssm ? 1 : 0
  name  = "${var.app_name}-ssm-profile"
  role  = aws_iam_role.ssm[0].name
}
```

### for Expressions (List and Map Comprehensions)

```hcl
# Transform a list into another list
locals {
  # Uppercase all names
  upper_names = [for name in var.names : upper(name)]

  # Filter a list
  production_instances = [
    for inst in var.instances : inst
    if inst.environment == "production"
  ]

  # Transform a list into a map
  instance_map = {
    for inst in var.instances : inst.name => inst.id
  }

  # Nested for: flatten a map of lists
  all_subnet_cidrs = flatten([
    for az, config in var.az_config : [
      for idx in range(config.subnet_count) :
      cidrsubnet(var.vpc_cidr, 8, config.offset + idx)
    ]
  ])

  # Group by: create a map of lists
  instances_by_az = {
    for inst in aws_instance.app : inst.availability_zone => inst.id...
  }
}
```

### Splat Expressions

```hcl
# Splat expressions are shorthand for common for expressions over lists.

# These two are equivalent:
locals {
  subnet_ids_splat = aws_subnet.private[*].id
  subnet_ids_for   = [for s in aws_subnet.private : s.id]
}

# Splat with nested attributes
output "instance_private_ips" {
  value = aws_instance.cluster[*].private_ip
}

# Note: splat only works with lists (count resources), not maps (for_each).
# For for_each resources, use values():
output "bucket_arns" {
  value = values(aws_s3_bucket.this)[*].arn
}
```

### try() and can() Functions

```hcl
# try() returns the first expression that does not produce an error.
# Useful for optional nested attributes or varying data structures.

locals {
  # Safely access a potentially missing attribute
  db_port = try(var.database_config.port, 5432)

  # Walk a nested structure with fallback
  log_level = try(
    var.app_config.logging.level,
    var.default_log_level,
    "info"
  )
}

# can() returns true if the expression evaluates without error.
# Useful in validation blocks and conditionals.

variable "cidr_block" {
  type = string
  validation {
    condition     = can(cidrnetmask(var.cidr_block))
    error_message = "Must be a valid CIDR notation (e.g., 10.0.0.0/16)."
  }
}

variable "json_config" {
  type = string
  validation {
    condition     = can(jsondecode(var.json_config))
    error_message = "Must be valid JSON."
  }
}
```

### coalesce() and coalescelist()

```hcl
# coalesce() returns the first non-null, non-empty-string argument.
locals {
  effective_region = coalesce(var.override_region, var.default_region, "us-east-1")
  instance_name    = coalesce(var.custom_name, "${var.project}-${var.environment}")
}

# coalescelist() returns the first non-empty list.
locals {
  security_groups = coalescelist(
    var.custom_security_group_ids,
    [aws_security_group.default.id]
  )

  subnet_ids = coalescelist(
    var.override_subnet_ids,
    data.aws_subnets.private.ids
  )
}
```

### merge() and lookup()

```hcl
# merge() combines maps. Later values override earlier ones.
locals {
  default_tags = {
    ManagedBy   = "terraform"
    Environment = var.environment
    Project     = var.project_name
  }

  extra_tags = {
    CostCenter = var.cost_center
    Owner      = var.owner
  }

  # Merge defaults with extras; extras win on conflict
  all_tags = merge(local.default_tags, local.extra_tags, var.additional_tags)
}

# lookup() gets a value from a map with a default fallback.
variable "instance_type_map" {
  default = {
    development = "t3.small"
    staging     = "t3.medium"
    production  = "m5.large"
  }
}

resource "aws_instance" "app" {
  ami           = data.aws_ami.amazon_linux.id
  instance_type = lookup(var.instance_type_map, var.environment, "t3.micro")
}
```

### templatefile()

```hcl
# templatefile() renders a template file with provided variables.

# templates/user-data.sh.tftpl
# #!/bin/bash
# set -euo pipefail
# echo "Deploying ${app_name} version ${app_version}"
# echo "DB_HOST=${db_host}" >> /etc/app/config.env
# %{ for key, value in env_vars ~}
# echo "${key}=${value}" >> /etc/app/config.env
# %{ endfor ~}

resource "aws_instance" "app" {
  ami           = data.aws_ami.amazon_linux.id
  instance_type = "t3.medium"

  user_data = base64encode(templatefile("${path.module}/templates/user-data.sh.tftpl", {
    app_name    = var.app_name
    app_version = var.app_version
    db_host     = aws_db_instance.main.address
    env_vars    = var.environment_variables
  }))
}

# templates/ecs-task-definition.json.tftpl — JSON template
# [
#   {
#     "name": "${name}",
#     "image": "${image}",
#     "cpu": ${cpu},
#     "memory": ${memory},
#     "essential": true,
#     "portMappings": [
#       { "containerPort": ${port}, "protocol": "tcp" }
#     ]
#   }
# ]
```

### cidrsubnet() for Network Planning

```hcl
# cidrsubnet(prefix, newbits, netnum) calculates subnet CIDRs deterministically.

locals {
  vpc_cidr = "10.0.0.0/16"

  # Public subnets: 10.0.0.0/24, 10.0.1.0/24, 10.0.2.0/24
  public_cidrs = [for i in range(3) : cidrsubnet(local.vpc_cidr, 8, i)]

  # Private subnets: 10.0.10.0/24, 10.0.11.0/24, 10.0.12.0/24
  private_cidrs = [for i in range(3) : cidrsubnet(local.vpc_cidr, 8, i + 10)]

  # Database subnets: 10.0.20.0/24, 10.0.21.0/24, 10.0.22.0/24
  database_cidrs = [for i in range(3) : cidrsubnet(local.vpc_cidr, 8, i + 20)]

  # Larger subnets (/20) for Kubernetes pods
  pod_cidrs = [for i in range(3) : cidrsubnet(local.vpc_cidr, 4, i + 8)]
}

resource "aws_subnet" "public" {
  count             = 3
  vpc_id            = aws_vpc.main.id
  cidr_block        = local.public_cidrs[count.index]
  availability_zone = data.aws_availability_zones.available.names[count.index]

  map_public_ip_on_launch = true

  tags = { Name = "public-${count.index}" }
}
```

### String Templates and Directives

```hcl
locals {
  # String interpolation
  resource_prefix = "${var.project}-${var.environment}"

  # Heredoc with interpolation
  policy_json = <<-EOT
    {
      "Version": "2012-10-17",
      "Statement": [{
        "Effect": "Allow",
        "Action": "s3:GetObject",
        "Resource": "arn:aws:s3:::${var.bucket_name}/*"
      }]
    }
  EOT

  # Directive: conditional inclusion in a template
  cloud_init = <<-EOT
    #cloud-config
    packages:
      - docker
    %{if var.install_monitoring_agent}
      - cloudwatch-agent
    %{endif}
    runcmd:
      - systemctl start docker
    %{for mount in var.ebs_mounts~}
      - mkdir -p ${mount.path}
      - mount ${mount.device} ${mount.path}
    %{endfor~}
  EOT
}
```

---

## Lifecycle Rules

Lifecycle meta-arguments control how Terraform creates, updates, and destroys resources.

### create_before_destroy

```hcl
# Create the replacement resource before destroying the old one.
# Essential for zero-downtime updates of resources like security groups and launch configs.

resource "aws_security_group" "app" {
  name_prefix = "${var.app_name}-"
  vpc_id      = var.vpc_id

  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_launch_template" "app" {
  name_prefix   = "${var.app_name}-"
  image_id      = var.ami_id
  instance_type = var.instance_type

  lifecycle {
    create_before_destroy = true
  }
}
```

### prevent_destroy

```hcl
# Prevent accidental destruction of critical resources.
# Terraform will error if a plan would destroy this resource.

resource "aws_db_instance" "production" {
  identifier     = "${var.project}-production"
  engine         = "postgres"
  engine_version = "15.4"
  instance_class = "db.r6g.xlarge"

  allocated_storage = 100
  storage_encrypted = true

  lifecycle {
    prevent_destroy = true
  }
}

resource "aws_s3_bucket" "audit_logs" {
  bucket = "${var.project}-audit-logs"

  lifecycle {
    prevent_destroy = true
  }
}
```

### ignore_changes

```hcl
# Ignore changes to specific attributes that are managed outside Terraform.
# Common for autoscaling, ECS desired count, or tags managed by other systems.

resource "aws_autoscaling_group" "app" {
  name                = "${var.app_name}-asg"
  min_size            = var.min_size
  max_size            = var.max_size
  desired_capacity    = var.desired_capacity
  vpc_zone_identifier = var.subnet_ids

  launch_template {
    id      = aws_launch_template.app.id
    version = "$Latest"
  }

  lifecycle {
    # ASG scaling policies change desired_capacity; don't revert it
    ignore_changes = [desired_capacity]
  }
}

resource "aws_ecs_service" "app" {
  name            = var.service_name
  cluster         = var.cluster_id
  task_definition = var.task_definition_arn
  desired_count   = var.desired_count

  lifecycle {
    # Application Auto Scaling manages desired_count
    ignore_changes = [desired_count, task_definition]
  }
}
```

### replace_triggered_by

```hcl
# Force replacement of a resource when a related resource changes.
# Added in Terraform 1.2.

resource "aws_launch_template" "app" {
  name_prefix   = "${var.app_name}-"
  image_id      = var.ami_id
  instance_type = var.instance_type
}

resource "aws_instance" "app" {
  ami           = aws_launch_template.app.image_id
  instance_type = aws_launch_template.app.instance_type

  lifecycle {
    # Replace the instance whenever the launch template changes
    replace_triggered_by = [
      aws_launch_template.app.id
    ]
  }
}
```

### Custom Conditions (precondition, postcondition)

```hcl
# precondition: validates assumptions before a resource is created or updated.
# postcondition: validates guarantees after a resource is applied.

resource "aws_instance" "app" {
  ami           = var.ami_id
  instance_type = var.instance_type
  subnet_id     = var.subnet_id

  lifecycle {
    precondition {
      condition     = var.instance_type != "t2.micro" || var.environment != "production"
      error_message = "t2.micro instances are not allowed in production."
    }

    postcondition {
      condition     = self.private_ip != ""
      error_message = "Instance must have a private IP address assigned."
    }
  }
}

# precondition on a data source
data "aws_ami" "app" {
  most_recent = true
  owners      = ["self"]

  filter {
    name   = "name"
    values = ["${var.app_name}-*"]
  }

  lifecycle {
    postcondition {
      condition     = self.image_id != ""
      error_message = "No AMI found matching the filter. Build one first with Packer."
    }
  }
}
```

---

## Terraform Testing

Terraform 1.6 introduced native testing with `.tftest.hcl` files, providing a built-in way to validate module behavior without external tools.

### terraform test Framework (TF 1.6+)

```hcl
# Run tests with: terraform test
# Test files live in the module directory or in a tests/ subdirectory.
# File extension must be .tftest.hcl
```

### Test File Structure (.tftest.hcl)

```hcl
# tests/s3_bucket.tftest.hcl

# Variables block sets inputs for all run blocks in this file
variables {
  bucket_name       = "test-bucket-unit-001"
  versioning_enabled = true
  kms_key_arn       = null
  force_destroy     = true
  tags = {
    Test = "true"
  }
}

# Provider configuration for tests
provider "aws" {
  region = "us-east-1"
}
```

### Run Blocks and Assertions

```hcl
# tests/s3_bucket.tftest.hcl

variables {
  bucket_name        = "test-bucket-assertions-001"
  versioning_enabled = true
  kms_key_arn        = null
  force_destroy      = true
  tags               = { Test = "true" }
}

# Plan-only test: validates plan output without creating resources
run "verify_bucket_name" {
  command = plan

  assert {
    condition     = aws_s3_bucket.this.bucket == "test-bucket-assertions-001"
    error_message = "Bucket name does not match expected value."
  }
}

# Apply test: creates real resources and validates them
run "verify_versioning_enabled" {
  command = apply

  assert {
    condition     = aws_s3_bucket_versioning.this.versioning_configuration[0].status == "Enabled"
    error_message = "Versioning should be enabled."
  }
}

run "verify_public_access_blocked" {
  command = apply

  assert {
    condition     = aws_s3_bucket_public_access_block.this.block_public_acls == true
    error_message = "Public ACLs should be blocked."
  }

  assert {
    condition     = aws_s3_bucket_public_access_block.this.restrict_public_buckets == true
    error_message = "Public bucket access should be restricted."
  }
}

run "verify_encryption_defaults_to_aes256" {
  command = apply

  assert {
    condition = (
      aws_s3_bucket_server_side_encryption_configuration.this.rule[0]
        .apply_server_side_encryption_by_default[0].sse_algorithm == "AES256"
    )
    error_message = "Default encryption should be AES256 when no KMS key is provided."
  }
}
```

### Mock Providers

```hcl
# tests/unit.tftest.hcl
# Mock providers allow plan-only testing without real API calls.

mock_provider "aws" {
  alias = "mock"
}

run "validate_resource_tags" {
  command = plan

  providers = {
    aws = aws.mock
  }

  variables {
    bucket_name = "mock-test-bucket"
    tags = {
      Environment = "test"
      ManagedBy   = "terraform"
    }
  }

  assert {
    condition     = aws_s3_bucket.this.tags["ManagedBy"] == "terraform"
    error_message = "ManagedBy tag must be set to terraform."
  }
}
```

### Terratest (Go-based Testing)

```hcl
# For reference: Terratest is a Go library for integration testing Terraform.
# Test file: test/s3_bucket_test.go
#
# func TestS3Bucket(t *testing.T) {
#   t.Parallel()
#   terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
#     TerraformDir: "../modules/s3-bucket",
#     Vars: map[string]interface{}{
#       "bucket_name":        fmt.Sprintf("test-%s", random.UniqueId()),
#       "versioning_enabled": true,
#       "force_destroy":      true,
#     },
#   })
#   defer terraform.Destroy(t, terraformOptions)
#   terraform.InitAndApply(t, terraformOptions)
#
#   bucketID := terraform.Output(t, terraformOptions, "bucket_id")
#   aws.AssertS3BucketExists(t, "us-east-1", bucketID)
#   aws.AssertS3BucketVersioningExists(t, "us-east-1", bucketID)
# }
```

### Example Test Suite

```hcl
# tests/integration.tftest.hcl
# End-to-end test that validates a full module with multiple run blocks.

variables {
  project_name = "test-project"
  environment  = "test"
  vpc_cidr     = "10.99.0.0/16"
}

run "create_networking" {
  command = apply
  module {
    source = "./modules/networking"
  }

  variables {
    vpc_cidr           = var.vpc_cidr
    availability_zones = ["us-east-1a", "us-east-1b"]
    environment        = var.environment
  }

  assert {
    condition     = length(output.private_subnet_ids) == 2
    error_message = "Expected 2 private subnets."
  }

  assert {
    condition     = output.vpc_cidr == var.vpc_cidr
    error_message = "VPC CIDR does not match input."
  }
}

run "create_database" {
  command = apply
  module {
    source = "./modules/rds-standard"
  }

  variables {
    db_identifier        = "${var.project_name}-test"
    engine_version       = "15.4"
    instance_class       = "db.t3.micro"
    allocated_storage    = 20
    db_subnet_group_name = run.create_networking.database_subnet_group_name
    security_group_ids   = [run.create_networking.database_security_group_id]
    environment          = var.environment
    data_classification  = "internal"
  }

  assert {
    condition     = output.db_endpoint != ""
    error_message = "Database endpoint should not be empty after creation."
  }
}
```

---

## Common Anti-Patterns

Recognizing and avoiding these patterns leads to more maintainable, safer Terraform configurations.

### God Module (Everything in One Module)

```hcl
# ANTI-PATTERN: One massive root module with hundreds of resources.
# This makes plans slow, blast radius huge, and code hard to review.

# BAD — everything in one main.tf:
# resource "aws_vpc" "main" { ... }
# resource "aws_subnet" "public" { ... }
# resource "aws_subnet" "private" { ... }
# resource "aws_nat_gateway" "main" { ... }
# resource "aws_ecs_cluster" "main" { ... }
# resource "aws_ecs_service" "api" { ... }
# resource "aws_ecs_service" "web" { ... }
# resource "aws_rds_cluster" "main" { ... }
# resource "aws_elasticache_cluster" "main" { ... }
# ... 200 more resources ...

# BETTER: Split into focused stacks/modules with defined interfaces.
# stacks/
#   networking/    — VPC, subnets, NAT, route tables
#   compute/       — ECS cluster, services, task definitions
#   data/          — RDS, ElastiCache, S3
#   monitoring/    — CloudWatch, alarms, dashboards
```

### Hardcoded Values

```hcl
# ANTI-PATTERN: Hardcoded values scattered throughout configurations.

# BAD:
# resource "aws_instance" "app" {
#   ami           = "ami-0abcdef1234567890"
#   instance_type = "t3.large"
#   subnet_id     = "subnet-12345"
# }

# GOOD: Use variables with validation, data sources, and locals.
data "aws_ami" "app" {
  most_recent = true
  owners      = ["self"]
  filter {
    name   = "name"
    values = ["${var.app_name}-*"]
  }
}

variable "instance_type" {
  type    = string
  default = "t3.medium"
  validation {
    condition     = can(regex("^(t3|m5|r6g)\\.", var.instance_type))
    error_message = "Only t3, m5, and r6g instance families are approved."
  }
}

resource "aws_instance" "app" {
  ami           = data.aws_ami.app.id
  instance_type = var.instance_type
  subnet_id     = var.subnet_id
}
```

### Missing State Locking

```hcl
# ANTI-PATTERN: S3 backend without DynamoDB locking.
# Two people running terraform apply simultaneously will corrupt state.

# BAD:
# terraform {
#   backend "s3" {
#     bucket = "my-state"
#     key    = "terraform.tfstate"
#     region = "us-east-1"
#   }
# }

# GOOD: Always include a DynamoDB table for locking.
terraform {
  backend "s3" {
    bucket         = "my-state"
    key            = "terraform.tfstate"
    region         = "us-east-1"
    dynamodb_table = "terraform-locks"
    encrypt        = true
  }
}
```

### Overly Complex Expressions

```hcl
# ANTI-PATTERN: Deeply nested expressions that are impossible to debug.

# BAD:
# locals {
#   result = flatten([for k, v in merge(var.a, { for x in var.b : x.name => x if can(x.enabled) && x.enabled }) : [for i in range(try(v.count, 1)) : { key = "${k}-${i}", value = try(v.values[i], v.default_value, "unknown") }] if length(try(v.values, [])) > 0 || try(v.include_default, false)])
# }

# GOOD: Break complex logic into named intermediate steps.
locals {
  # Step 1: Merge the two input sources
  combined_config = merge(var.a, local.b_as_map)

  # Step 2: Convert list b into a map keyed by name
  b_as_map = {
    for item in var.b : item.name => item
    if try(item.enabled, false)
  }

  # Step 3: Expand each config entry into indexed instances
  expanded = flatten([
    for key, config in local.combined_config : [
      for i in range(try(config.count, 1)) : {
        key   = "${key}-${i}"
        value = try(config.values[i], config.default_value, "unknown")
      }
    ]
    if length(try(config.values, [])) > 0 || try(config.include_default, false)
  ])
}
```

### Ignoring Plan Output

```hcl
# ANTI-PATTERN: Running terraform apply -auto-approve without reviewing the plan.

# GOOD PRACTICE: Always review the plan. Use a saved plan file in CI/CD.

# Step 1: Generate plan
# terraform plan -out=tfplan

# Step 2: Review plan (human or automated policy check)
# terraform show tfplan

# Step 3: Apply the reviewed plan
# terraform apply tfplan

# In CI/CD, use OPA or Sentinel to enforce policy on the plan JSON:
# terraform show -json tfplan > tfplan.json
# opa eval --data policies/ --input tfplan.json "data.terraform.deny"
```

### Not Using terraform fmt and validate

```hcl
# ANTI-PATTERN: Skipping formatting and validation in CI.

# GOOD: Run these in pre-commit hooks and CI pipelines.
# terraform fmt -check -recursive
# terraform validate

# Example pre-commit hook (.pre-commit-config.yaml):
# repos:
#   - repo: https://github.com/antonbabenko/pre-commit-terraform
#     rev: v1.83.5
#     hooks:
#       - id: terraform_fmt
#       - id: terraform_validate
#       - id: terraform_tflint
#       - id: terraform_docs
```

---

## Naming Conventions

Consistent naming makes Terraform code scannable and reduces cognitive load during reviews.

### Resource Naming

```hcl
# 1. Use "this" when a module manages a single primary resource of that type.
resource "aws_s3_bucket" "this" {
  bucket = var.bucket_name
}

# 2. Use descriptive names when a module has multiple resources of the same type.
resource "aws_security_group" "alb" {
  name_prefix = "${var.app_name}-alb-"
  vpc_id      = var.vpc_id
}

resource "aws_security_group" "ecs_tasks" {
  name_prefix = "${var.app_name}-ecs-"
  vpc_id      = var.vpc_id
}

# 3. Use snake_case for all Terraform identifiers.
resource "aws_iam_role" "ecs_task_execution" {
  name = "${var.app_name}-ecs-task-execution"
}

# 4. Name AWS resources with a consistent pattern: project-environment-purpose
resource "aws_db_instance" "main" {
  identifier = "${var.project_name}-${var.environment}-primary"
}
```

### Variable Naming

```hcl
# Use descriptive names. Prefix booleans with enable_ or is_.
variable "enable_nat_gateway" {
  description = "Whether to create a NAT gateway for private subnets"
  type        = bool
  default     = true
}

variable "is_public" {
  description = "Whether the instance should receive a public IP"
  type        = bool
  default     = false
}

# Use _id, _arn, _name suffixes for reference variables.
variable "vpc_id" {
  description = "ID of the VPC"
  type        = string
}

variable "kms_key_arn" {
  description = "ARN of the KMS key for encryption"
  type        = string
  default     = null
}

# Use _ids (plural) for lists of references.
variable "subnet_ids" {
  description = "List of subnet IDs for the service"
  type        = list(string)
}
```

### Output Naming

```hcl
# Mirror the pattern: resource_type_attribute
output "vpc_id" {
  description = "The ID of the VPC"
  value       = aws_vpc.main.id
}

output "vpc_cidr_block" {
  description = "The CIDR block of the VPC"
  value       = aws_vpc.main.cidr_block
}

output "private_subnet_ids" {
  description = "List of private subnet IDs"
  value       = aws_subnet.private[*].id
}

output "database_endpoint" {
  description = "The connection endpoint for the RDS instance"
  value       = aws_db_instance.main.endpoint
  sensitive   = false
}

output "database_password" {
  description = "The master password for the database"
  value       = aws_db_instance.main.password
  sensitive   = true  # Mark sensitive outputs
}
```

### File Organization

```hcl
# Standard file layout for a module:
#
# modules/my-module/
#   main.tf          — Primary resource definitions
#   variables.tf     — All input variables
#   outputs.tf       — All outputs
#   versions.tf      — terraform and provider version constraints
#   data.tf          — Data sources (optional, can go in main.tf for small modules)
#   locals.tf        — Local values (optional)
#   README.md        — Auto-generated with terraform-docs
#
# Root module layout:
#
# environments/production/
#   main.tf          — Module calls and provider config
#   backend.tf       — Backend configuration
#   variables.tf     — Input variables
#   outputs.tf       — Outputs
#   terraform.tfvars — Variable values (or use .auto.tfvars)
#   versions.tf      — Version constraints
```

---

## Dependency Management

Terraform automatically infers most dependencies from resource references. Explicit dependency declarations handle the remaining edge cases.

### Implicit Dependencies

```hcl
# Terraform detects dependencies from attribute references.
# No depends_on is needed here — Terraform knows the order.

resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"
}

# Implicit dependency: references aws_vpc.main.id
resource "aws_subnet" "public" {
  vpc_id     = aws_vpc.main.id
  cidr_block = "10.0.1.0/24"
}

# Implicit dependency: references aws_subnet.public.id
resource "aws_instance" "web" {
  ami           = data.aws_ami.amazon_linux.id
  instance_type = "t3.micro"
  subnet_id     = aws_subnet.public.id
}

# Chain: VPC -> Subnet -> Instance (automatically ordered)
```

### Explicit depends_on

```hcl
# Use depends_on when a dependency exists but is not captured by references.
# Common case: IAM policies must be attached before a resource uses the role.

resource "aws_iam_role" "lambda" {
  name = "${var.function_name}-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = { Service = "lambda.amazonaws.com" }
    }]
  })
}

resource "aws_iam_role_policy_attachment" "lambda_basic" {
  role       = aws_iam_role.lambda.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_lambda_function" "this" {
  function_name = var.function_name
  role          = aws_iam_role.lambda.arn
  handler       = "index.handler"
  runtime       = "nodejs20.x"
  filename      = data.archive_file.lambda.output_path

  # Without depends_on, the Lambda may be created before the policy
  # is attached, causing invocation failures.
  depends_on = [aws_iam_role_policy_attachment.lambda_basic]
}

# Another common case: VPC endpoints must exist before resources use them.
resource "aws_vpc_endpoint" "s3" {
  vpc_id       = aws_vpc.main.id
  service_name = "com.amazonaws.${var.region}.s3"
}

resource "aws_instance" "private_app" {
  ami           = data.aws_ami.amazon_linux.id
  instance_type = "t3.micro"
  subnet_id     = aws_subnet.private.id

  # Ensure the S3 VPC endpoint is ready before the instance boots
  # and tries to pull packages from S3-backed repos.
  depends_on = [aws_vpc_endpoint.s3]
}
```

### Data Source Dependencies

```hcl
# Data sources also participate in the dependency graph.
# They are read during plan (or apply for depends_on cases).

resource "aws_iam_role" "app" {
  name = "${var.app_name}-role"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

# This data source has an implicit dependency on the role
data "aws_iam_policy_document" "assume_role" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["ecs-tasks.amazonaws.com"]
    }
  }
}

# When a data source must wait for a resource, use depends_on.
resource "aws_iam_role_policy" "app" {
  name   = "app-policy"
  role   = aws_iam_role.app.id
  policy = data.aws_iam_policy_document.app_permissions.json
}

data "aws_iam_policy_document" "app_permissions" {
  statement {
    actions   = ["s3:GetObject"]
    resources = ["${aws_s3_bucket.app_data.arn}/*"]
  }

  # Data source waits for the bucket to exist
  depends_on = [aws_s3_bucket.app_data]
}
```

### Module Dependencies

```hcl
# Modules inherit dependencies from their input references.
# When module B uses an output from module A, Terraform orders them correctly.

module "networking" {
  source = "./modules/networking"
  vpc_cidr = var.vpc_cidr
}

# Implicit dependency: references module.networking outputs
module "compute" {
  source     = "./modules/compute"
  vpc_id     = module.networking.vpc_id
  subnet_ids = module.networking.private_subnet_ids
}

# Implicit dependency: references both networking and compute outputs
module "database" {
  source             = "./modules/database"
  subnet_group_name  = module.networking.database_subnet_group_name
  security_group_ids = [module.compute.db_client_security_group_id]
}

# When there is no reference but an ordering requirement exists,
# use depends_on on the module block.
module "dns" {
  source = "./modules/dns"

  zone_id     = var.zone_id
  record_name = var.app_domain
  alb_dns     = module.compute.alb_dns_name

  # Ensure the certificate is validated before creating DNS records
  # that reference it, even though there is no direct attribute reference.
  depends_on = [module.certificate_validation]
}
```

---

## Quick Reference: Common Patterns Summary

| Pattern | When to Use | Key Benefit |
|---|---|---|
| Flat Module | Single-purpose, few resources | Simplicity and clarity |
| Nested Module | Multi-tier infrastructure | Encapsulation of complex stacks |
| Facade Module | Wrapping community modules | Org-standard guardrails |
| Dependency Injection | Shared resources across modules | Testability and composability |
| `for_each` with maps | Distinct, named resources | Stable state addressing |
| `count` for conditionals | 0 or 1 resource creation | Simple on/off toggle |
| Dynamic blocks | Variable-length nested blocks | Reduced repetition |
| `create_before_destroy` | Zero-downtime replacements | Availability during updates |
| `prevent_destroy` | Critical data stores | Safety against accidental deletion |
| `terraform test` | Module contract validation | Built-in, no external tools |
| Remote state | Cross-stack references | Team collaboration |
| `moved` blocks | Refactoring without recreation | Safe state migrations |
