---
name: terraform-modules
description: >
  Terraform module patterns — reusable modules, input/output design,
  resource composition, conditional resources, for_each patterns,
  and module testing with Terratest.
  Triggers: "terraform module", "terraform reusable", "terraform for_each",
  "terraform conditional", "terraform composition", "terratest".
  NOT for: State management or backend config (use terraform-state-management).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Terraform Module Patterns

## Module Structure

```
modules/
  vpc/
    main.tf           # Resources
    variables.tf      # Input variables
    outputs.tf        # Output values
    versions.tf       # Provider requirements
    README.md
  rds/
    main.tf
    variables.tf
    outputs.tf
    versions.tf
  ecs-service/
    main.tf
    variables.tf
    outputs.tf
    versions.tf
    templates/
      task-definition.json.tpl
environments/
  production/
    main.tf           # Root module
    variables.tf
    outputs.tf
    terraform.tfvars
    backend.tf
  staging/
    main.tf
    variables.tf
    outputs.tf
    terraform.tfvars
    backend.tf
```

## Variable Design

```hcl
# modules/vpc/variables.tf

variable "name" {
  description = "Name prefix for all VPC resources"
  type        = string

  validation {
    condition     = length(var.name) > 0 && length(var.name) <= 32
    error_message = "Name must be 1-32 characters."
  }
}

variable "cidr_block" {
  description = "CIDR block for the VPC"
  type        = string
  default     = "10.0.0.0/16"

  validation {
    condition     = can(cidrhost(var.cidr_block, 0))
    error_message = "Must be a valid CIDR block."
  }
}

variable "availability_zones" {
  description = "List of AZs to deploy into"
  type        = list(string)

  validation {
    condition     = length(var.availability_zones) >= 2
    error_message = "At least 2 AZs required for high availability."
  }
}

variable "enable_nat_gateway" {
  description = "Deploy NAT gateways for private subnet internet access"
  type        = bool
  default     = true
}

variable "single_nat_gateway" {
  description = "Use a single NAT gateway (cost saving, less HA)"
  type        = bool
  default     = false
}

variable "tags" {
  description = "Tags to apply to all resources"
  type        = map(string)
  default     = {}
}

# Complex variable with object type
variable "private_subnet_config" {
  description = "Configuration for private subnets"
  type = object({
    cidrs           = list(string)
    enable_flow_log = optional(bool, false)
    nacl_rules = optional(list(object({
      rule_number = number
      protocol    = string
      action      = string
      cidr_block  = string
      from_port   = number
      to_port     = number
    })), [])
  })
}
```

## Resource Composition

```hcl
# modules/vpc/main.tf

resource "aws_vpc" "this" {
  cidr_block           = var.cidr_block
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = merge(var.tags, {
    Name = "${var.name}-vpc"
  })
}

# Public subnets
resource "aws_subnet" "public" {
  for_each = { for idx, az in var.availability_zones : az => idx }

  vpc_id                  = aws_vpc.this.id
  cidr_block              = cidrsubnet(var.cidr_block, 8, each.value)
  availability_zone       = each.key
  map_public_ip_on_launch = true

  tags = merge(var.tags, {
    Name = "${var.name}-public-${each.key}"
    Tier = "public"
  })
}

# Private subnets
resource "aws_subnet" "private" {
  for_each = { for idx, az in var.availability_zones : az => idx }

  vpc_id            = aws_vpc.this.id
  cidr_block        = cidrsubnet(var.cidr_block, 8, each.value + length(var.availability_zones))
  availability_zone = each.key

  tags = merge(var.tags, {
    Name = "${var.name}-private-${each.key}"
    Tier = "private"
  })
}

# Conditional NAT gateway
resource "aws_eip" "nat" {
  for_each = var.enable_nat_gateway ? (
    var.single_nat_gateway ? { single = 0 } : { for idx, az in var.availability_zones : az => idx }
  ) : {}

  domain = "vpc"

  tags = merge(var.tags, {
    Name = "${var.name}-nat-eip-${each.key}"
  })
}

resource "aws_nat_gateway" "this" {
  for_each = aws_eip.nat

  allocation_id = each.value.id
  subnet_id     = var.single_nat_gateway ? values(aws_subnet.public)[0].id : aws_subnet.public[each.key].id

  tags = merge(var.tags, {
    Name = "${var.name}-nat-${each.key}"
  })

  depends_on = [aws_internet_gateway.this]
}
```

## Outputs

```hcl
# modules/vpc/outputs.tf

output "vpc_id" {
  description = "The ID of the VPC"
  value       = aws_vpc.this.id
}

output "vpc_cidr" {
  description = "The CIDR block of the VPC"
  value       = aws_vpc.this.cidr_block
}

output "public_subnet_ids" {
  description = "List of public subnet IDs"
  value       = [for subnet in aws_subnet.public : subnet.id]
}

output "private_subnet_ids" {
  description = "List of private subnet IDs"
  value       = [for subnet in aws_subnet.private : subnet.id]
}

output "nat_gateway_ips" {
  description = "Elastic IPs of NAT gateways"
  value       = [for eip in aws_eip.nat : eip.public_ip]
}

# Structured output for downstream modules
output "subnet_groups" {
  description = "Subnet groups organized by tier"
  value = {
    public = {
      ids  = [for s in aws_subnet.public : s.id]
      cidrs = [for s in aws_subnet.public : s.cidr_block]
    }
    private = {
      ids  = [for s in aws_subnet.private : s.id]
      cidrs = [for s in aws_subnet.private : s.cidr_block]
    }
  }
}
```

## Root Module (Environment)

```hcl
# environments/production/main.tf

module "vpc" {
  source = "../../modules/vpc"

  name               = "prod"
  cidr_block         = "10.0.0.0/16"
  availability_zones = ["us-east-1a", "us-east-1b", "us-east-1c"]
  enable_nat_gateway = true
  single_nat_gateway = false

  private_subnet_config = {
    cidrs           = ["10.0.10.0/24", "10.0.11.0/24", "10.0.12.0/24"]
    enable_flow_log = true
  }

  tags = local.common_tags
}

module "rds" {
  source = "../../modules/rds"

  name              = "prod-db"
  engine            = "postgres"
  engine_version    = "16.1"
  instance_class    = "db.r6g.large"
  allocated_storage = 100

  vpc_id             = module.vpc.vpc_id
  subnet_ids         = module.vpc.private_subnet_ids
  allowed_cidr_blocks = module.vpc.subnet_groups.private.cidrs

  multi_az               = true
  backup_retention_period = 7
  deletion_protection     = true

  tags = local.common_tags
}

module "api_service" {
  source = "../../modules/ecs-service"

  name        = "api"
  environment = "production"

  cluster_id  = module.ecs_cluster.cluster_id
  vpc_id      = module.vpc.vpc_id
  subnet_ids  = module.vpc.private_subnet_ids

  container_image = "${var.ecr_repository_url}:${var.api_version}"
  container_port  = 8080
  cpu             = 512
  memory          = 1024
  desired_count   = 3

  environment_variables = {
    DATABASE_URL = module.rds.connection_string
    REDIS_URL    = module.elasticache.endpoint
    LOG_LEVEL    = "info"
  }

  secrets = {
    JWT_SECRET = aws_ssm_parameter.jwt_secret.arn
    API_KEY    = aws_secretsmanager_secret.api_key.arn
  }

  health_check_path = "/health"

  tags = local.common_tags
}

locals {
  common_tags = {
    Environment = "production"
    ManagedBy   = "terraform"
    Project     = "myapp"
  }
}
```

## Dynamic Blocks and for_each Patterns

```hcl
# Security group with dynamic rules
resource "aws_security_group" "this" {
  name_prefix = "${var.name}-"
  vpc_id      = var.vpc_id

  dynamic "ingress" {
    for_each = var.ingress_rules
    content {
      from_port   = ingress.value.from_port
      to_port     = ingress.value.to_port
      protocol    = ingress.value.protocol
      cidr_blocks = ingress.value.cidr_blocks
      description = ingress.value.description
    }
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  lifecycle {
    create_before_destroy = true
  }

  tags = merge(var.tags, { Name = var.name })
}

# for_each with map transformation
variable "services" {
  type = map(object({
    port        = number
    protocol    = string
    health_path = string
  }))
}

resource "aws_lb_target_group" "services" {
  for_each = var.services

  name        = "${var.name}-${each.key}"
  port        = each.value.port
  protocol    = each.value.protocol
  vpc_id      = var.vpc_id
  target_type = "ip"

  health_check {
    path     = each.value.health_path
    port     = each.value.port
    protocol = each.value.protocol
  }
}

# Conditional resources
resource "aws_cloudwatch_log_group" "this" {
  count = var.enable_logging ? 1 : 0

  name              = "/app/${var.name}"
  retention_in_days = var.log_retention_days
}
```

## Testing with Terratest

```go
// test/vpc_test.go
package test

import (
    "testing"

    "github.com/gruntwork-io/terratest/modules/terraform"
    "github.com/stretchr/testify/assert"
)

func TestVpcModule(t *testing.T) {
    t.Parallel()

    opts := &terraform.Options{
        TerraformDir: "../modules/vpc",
        Vars: map[string]interface{}{
            "name":               "test",
            "cidr_block":         "10.99.0.0/16",
            "availability_zones": []string{"us-east-1a", "us-east-1b"},
            "enable_nat_gateway":  false,
            "tags":               map[string]string{"Test": "true"},
            "private_subnet_config": map[string]interface{}{
                "cidrs":           []string{"10.99.10.0/24", "10.99.11.0/24"},
                "enable_flow_log": false,
            },
        },
    }

    defer terraform.Destroy(t, opts)
    terraform.InitAndApply(t, opts)

    vpcId := terraform.Output(t, opts, "vpc_id")
    assert.NotEmpty(t, vpcId)

    publicSubnets := terraform.OutputList(t, opts, "public_subnet_ids")
    assert.Len(t, publicSubnets, 2)

    privateSubnets := terraform.OutputList(t, opts, "private_subnet_ids")
    assert.Len(t, privateSubnets, 2)
}
```

## Gotchas

1. **`count` vs `for_each`** — `count` uses index-based addressing (`resource[0]`, `resource[1]`). Removing an item shifts all indexes, causing unnecessary destroys. `for_each` uses key-based addressing — removing an item only destroys that specific resource. Always prefer `for_each` for sets of similar resources.

2. **Sensitive outputs leak in state** — Marking an output `sensitive = true` hides it from CLI output, but it's still stored in plaintext in the state file. Encrypt your state backend (S3 with SSE, GCS with CMEK) and restrict access to the state bucket.

3. **`depends_on` on modules is a sledgehammer** — `depends_on` on a module creates a dependency on ALL resources in that module. If module A `depends_on` module B, changing ANY resource in B forces ALL resources in A to be re-evaluated. Use data sources or explicit resource references instead.

4. **Terraform plan doesn't catch IAM permission errors** — `terraform plan` validates syntax and provider API calls, but doesn't verify the executing role has permission to create the planned resources. The error only appears during `apply`.

5. **`create_before_destroy` with unique name constraints** — Resources with unique names (security groups, IAM roles) can't use `create_before_destroy` because the new resource can't be created while the old one exists with the same name. Use `name_prefix` instead of `name` for these resources.

6. **Module source pinning** — `source = "git::https://..."` without `?ref=v1.2.3` pulls the latest commit, which can break your infrastructure on the next `init`. Always pin module sources to a specific tag, commit SHA, or version constraint.
