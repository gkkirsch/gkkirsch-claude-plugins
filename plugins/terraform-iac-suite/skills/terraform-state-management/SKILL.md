---
name: terraform-state-management
description: >
  Terraform state management — remote backends, state locking, workspaces,
  state manipulation, import, migration, and disaster recovery patterns.
  Triggers: "terraform state", "terraform backend", "terraform workspace",
  "terraform import", "terraform migrate", "terraform lock", "tfstate".
  NOT for: Module design or resource patterns (use terraform-modules).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Terraform State Management

## Remote Backend Configuration

```hcl
# backend.tf — S3 backend (AWS)
terraform {
  backend "s3" {
    bucket         = "mycompany-terraform-state"
    key            = "production/api/terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true
    dynamodb_table = "terraform-locks"
    # Use assume_role for cross-account access
    # role_arn     = "arn:aws:iam::123456789:role/TerraformStateAccess"
  }
}

# backend.tf — GCS backend (GCP)
terraform {
  backend "gcs" {
    bucket = "mycompany-terraform-state"
    prefix = "production/api"
  }
}

# backend.tf — Azure backend
terraform {
  backend "azurerm" {
    resource_group_name  = "terraform-state-rg"
    storage_account_name = "tfstate12345"
    container_name       = "tfstate"
    key                  = "production/api/terraform.tfstate"
  }
}
```

## State Locking Setup (AWS)

```hcl
# Bootstrap — create the state infrastructure first
# Run this once with local backend, then migrate

resource "aws_s3_bucket" "terraform_state" {
  bucket = "mycompany-terraform-state"

  lifecycle {
    prevent_destroy = true
  }
}

resource "aws_s3_bucket_versioning" "terraform_state" {
  bucket = aws_s3_bucket.terraform_state.id
  versioning_configuration {
    status = "Enabled" # Enables state recovery
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "terraform_state" {
  bucket = aws_s3_bucket.terraform_state.id
  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "aws:kms"
    }
  }
}

resource "aws_s3_bucket_public_access_block" "terraform_state" {
  bucket                  = aws_s3_bucket.terraform_state.id
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

  lifecycle {
    prevent_destroy = true
  }
}
```

## State Key Organization

```
# Recommended state key hierarchy
s3://terraform-state/
  bootstrap/
    terraform.tfstate           # State infra itself
  production/
    networking/terraform.tfstate # VPC, subnets, NAT
    database/terraform.tfstate   # RDS, ElastiCache
    api/terraform.tfstate        # ECS services, ALB
    monitoring/terraform.tfstate # CloudWatch, alerts
  staging/
    networking/terraform.tfstate
    database/terraform.tfstate
    api/terraform.tfstate
  shared/
    dns/terraform.tfstate       # Route53 zones
    ecr/terraform.tfstate       # Container registries
    iam/terraform.tfstate       # IAM roles, policies
```

## Workspaces

```bash
# List workspaces
terraform workspace list

# Create and switch
terraform workspace new staging
terraform workspace new production

# Switch
terraform workspace select production

# Show current
terraform workspace show
```

```hcl
# Using workspace in configuration
locals {
  env_config = {
    staging = {
      instance_type = "t3.small"
      min_count     = 1
      max_count     = 2
    }
    production = {
      instance_type = "t3.large"
      min_count     = 3
      max_count     = 10
    }
  }

  config = local.env_config[terraform.workspace]
}

resource "aws_instance" "app" {
  instance_type = local.config.instance_type
  # ...
}

# Workspace-aware backend key
terraform {
  backend "s3" {
    bucket = "terraform-state"
    key    = "app/terraform.tfstate"
    # Each workspace gets: app/env:/staging/terraform.tfstate
  }
}
```

## State Manipulation

```bash
# List all resources in state
terraform state list

# Show a specific resource
terraform state show aws_instance.web

# Move a resource (rename or restructure)
terraform state mv aws_instance.web aws_instance.api

# Move into a module
terraform state mv aws_instance.web module.compute.aws_instance.web

# Remove from state (resource still exists in cloud)
terraform state rm aws_instance.legacy

# Pull remote state to local file
terraform state pull > backup.tfstate

# Push local state to remote (DANGEROUS)
terraform state push backup.tfstate

# Replace provider (e.g., after fork)
terraform state replace-provider hashicorp/aws registry.example.com/aws
```

## Import Existing Resources

```bash
# Import a single resource
terraform import aws_instance.web i-1234567890abcdef0

# Import with module path
terraform import module.vpc.aws_vpc.this vpc-abc123

# Import with for_each key
terraform import 'aws_subnet.public["us-east-1a"]' subnet-xyz789
```

```hcl
# Terraform 1.5+ import blocks (declarative)
import {
  to = aws_instance.web
  id = "i-1234567890abcdef0"
}

import {
  to = aws_s3_bucket.assets
  id = "my-existing-bucket"
}

# Generate config from import
# terraform plan -generate-config-out=generated.tf
```

## State Migration

```bash
# Migrate from local to S3 backend
# 1. Add backend config to backend.tf
# 2. Run init with migration
terraform init -migrate-state

# Migrate between backends
# 1. Update backend block
# 2. Run init
terraform init -migrate-state
# Terraform will ask to copy state

# Migrate state between state files (splitting a monolith)
# 1. In source project: remove resource from state
terraform state rm aws_rds_instance.db

# 2. In destination project: import the resource
terraform import aws_rds_instance.db my-db-instance

# Safer approach: state mv between backends
terraform state mv \
  -state=monolith.tfstate \
  -state-out=database.tfstate \
  aws_rds_instance.db aws_rds_instance.db
```

## Data Sources for Cross-Stack References

```hcl
# Reference another state file's outputs
data "terraform_remote_state" "networking" {
  backend = "s3"
  config = {
    bucket = "terraform-state"
    key    = "production/networking/terraform.tfstate"
    region = "us-east-1"
  }
}

# Use outputs from the other state
resource "aws_instance" "app" {
  subnet_id = data.terraform_remote_state.networking.outputs.private_subnet_ids[0]
  vpc_security_group_ids = [
    data.terraform_remote_state.networking.outputs.app_security_group_id
  ]
}

# Alternative: SSM Parameter Store for decoupling
resource "aws_ssm_parameter" "vpc_id" {
  name  = "/infra/production/vpc_id"
  type  = "String"
  value = aws_vpc.this.id
}

# In another project
data "aws_ssm_parameter" "vpc_id" {
  name = "/infra/production/vpc_id"
}
```

## Lifecycle Rules

```hcl
resource "aws_db_instance" "production" {
  # ...

  lifecycle {
    # Never destroy this resource via terraform
    prevent_destroy = true

    # Ignore changes made outside terraform
    ignore_changes = [
      tags["LastModifiedBy"],
      engine_version, # Managed by AWS auto-upgrade
    ]

    # Create replacement before destroying old
    create_before_destroy = true

    # Custom condition
    precondition {
      condition     = var.environment == "production" ? var.multi_az : true
      error_message = "Production databases must be multi-AZ."
    }

    postcondition {
      condition     = self.status == "available"
      error_message = "Database is not in available state."
    }
  }
}

# Moved blocks — refactor without destroy
moved {
  from = aws_instance.web
  to   = module.compute.aws_instance.web
}

moved {
  from = module.old_name
  to   = module.new_name
}
```

## Disaster Recovery

```bash
# State is versioned in S3 — list versions
aws s3api list-object-versions \
  --bucket terraform-state \
  --prefix production/api/terraform.tfstate

# Restore a previous version
aws s3api get-object \
  --bucket terraform-state \
  --key production/api/terraform.tfstate \
  --version-id "abc123" \
  restored.tfstate

# Force unlock (when lock is stuck)
terraform force-unlock LOCK_ID

# Refresh state from actual cloud resources
terraform apply -refresh-only

# Detect drift without making changes
terraform plan -detailed-exitcode
# Exit 0: no changes, Exit 1: error, Exit 2: changes detected
```

## CI/CD Pipeline Patterns

```yaml
# GitHub Actions workflow
name: Terraform
on:
  pull_request:
    paths: ['infra/**']
  push:
    branches: [main]
    paths: ['infra/**']

jobs:
  plan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: hashicorp/setup-terraform@v3

      - name: Terraform Init
        run: terraform init -no-color
        working-directory: infra/production

      - name: Terraform Validate
        run: terraform validate -no-color

      - name: Terraform Plan
        id: plan
        run: terraform plan -no-color -out=tfplan
        working-directory: infra/production

      - name: Comment Plan on PR
        if: github.event_name == 'pull_request'
        uses: actions/github-script@v7
        with:
          script: |
            const plan = `${{ steps.plan.outputs.stdout }}`
            github.rest.issues.createComment({
              owner: context.repo.owner,
              repo: context.repo.repo,
              issue_number: context.issue.number,
              body: `### Terraform Plan\n\`\`\`\n${plan}\n\`\`\``
            })

  apply:
    needs: plan
    if: github.ref == 'refs/heads/main'
    runs-on: ubuntu-latest
    environment: production
    steps:
      - uses: actions/checkout@v4
      - uses: hashicorp/setup-terraform@v3
      - run: terraform init -no-color
        working-directory: infra/production
      - run: terraform apply -auto-approve -no-color
        working-directory: infra/production
```

## Gotchas

1. **Never edit `.tfstate` files manually** — State files are JSON but have internal checksums and serial numbers. Manual edits corrupt the state. Use `terraform state` commands, `terraform import`, or `moved` blocks for all state changes.

2. **State lock contention in CI/CD** — If a pipeline crashes mid-apply, the DynamoDB lock stays. The next run fails with "state locked." Use `terraform force-unlock <LOCK_ID>` cautiously, and add retry logic to your CI pipeline.

3. **`terraform destroy` is permanent** — There's no undo for `terraform destroy`. Use `prevent_destroy = true` on critical resources (databases, S3 buckets with data). Consider a separate approval step for destroy operations.

4. **Workspace state isolation is partial** — Workspaces share the same backend bucket and config. They're good for environment isolation (dev/staging/prod) but NOT for team isolation. Different teams should use different state files with different backend keys.

5. **`-target` creates state drift** — `terraform apply -target=aws_instance.web` only manages that resource, leaving others potentially out of sync. It's a debugging tool, not a workflow. Always run a full `terraform plan` after using `-target`.

6. **Remote state data source caching** — `terraform_remote_state` reads the state file at plan time and caches it. If the referenced state changes between plan and apply, you get stale values. For frequently changing values, use SSM Parameter Store or Consul instead.
