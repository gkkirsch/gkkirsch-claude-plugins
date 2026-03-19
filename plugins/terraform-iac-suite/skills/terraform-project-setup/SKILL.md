# Terraform Project Setup Skill

Scaffold and configure Terraform projects with module structure, backend config, CI/CD pipelines, and development tooling.

## Triggers

- "set up a terraform project"
- "create a new terraform project"
- "scaffold terraform infrastructure"
- "initialize terraform"
- "terraform project setup"
- "new infrastructure project"
- "create terraform modules"

## Workflow

### Step 1: Gather Requirements

Ask the user:
1. **Cloud provider**: AWS (default), GCP, Azure, or multi-cloud
2. **Project type**: New infrastructure, existing resource import, or module library
3. **Environment strategy**: workspace-per-env, directory-per-env, or single environment
4. **State backend**: S3 (default), Terraform Cloud, GCS, or Azure Blob
5. **CI/CD**: GitHub Actions (default), GitLab CI, Atlantis, or Terraform Cloud

### Step 2: Project Structure

Create the standard project layout. Show the directory tree, then create all files.

```
my-infra/
├── modules/
│   ├── networking/
│   │   ├── main.tf, variables.tf, outputs.tf
│   ├── compute/
│   │   ├── main.tf, variables.tf, outputs.tf
│   └── data/
│       ├── main.tf, variables.tf, outputs.tf
├── environments/
│   ├── dev.tfvars
│   ├── staging.tfvars
│   └── production.tfvars
├── main.tf
├── variables.tf
├── outputs.tf
├── versions.tf
├── providers.tf
├── backend.tf
├── locals.tf
├── .github/workflows/terraform.yml
├── .terraform-version
├── .tflint.hcl
├── .pre-commit-config.yaml
├── Makefile
└── .gitignore
```

### Step 3: Backend Configuration

**S3 backend with DynamoDB locking (default):**
```hcl
# backend.tf
terraform {
  backend "s3" {
    bucket         = "COMPANY-terraform-state"
    key            = "PROJECT_NAME/terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true
    dynamodb_table = "terraform-state-locks"
    use_lockfile   = true
  }
}
```

**Terraform Cloud workspace:**
```hcl
terraform {
  cloud {
    organization = "COMPANY"
    workspaces { tags = ["PROJECT_NAME"] }
  }
}
```

**Partial backend config (for multi-env, completed via `terraform init -backend-config="key=dev/terraform.tfstate"`):**
```hcl
terraform {
  backend "s3" {
    bucket         = "COMPANY-terraform-state"
    region         = "us-east-1"
    encrypt        = true
    dynamodb_table = "terraform-state-locks"
    use_lockfile   = true
  }
}
```

### Step 4: Provider Setup

**versions.tf:**
```hcl
terraform {
  required_version = ">= 1.9"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}
```

**providers.tf:**
```hcl
provider "aws" {
  region = var.aws_region
  default_tags {
    tags = {
      Environment = var.environment
      Project     = var.project_name
      ManagedBy   = "terraform"
    }
  }
}
```

**variables.tf (root):**
```hcl
variable "environment" {
  description = "Deployment environment"
  type        = string
  validation {
    condition     = contains(["dev", "staging", "production"], var.environment)
    error_message = "Environment must be dev, staging, or production."
  }
}
variable "project_name" { type = string }
variable "aws_region"   { type = string; default = "us-east-1" }
variable "owner"        { type = string }
variable "vpc_cidr"     { type = string; default = "10.0.0.0/16" }
variable "container_image" { type = string; default = "nginx:latest" }
variable "container_port"  { type = number; default = 8080 }
variable "desired_count"   { type = number; default = 2 }
variable "availability_zones" {
  type    = list(string)
  default = ["us-east-1a", "us-east-1b", "us-east-1c"]
}
```

**locals.tf:**
```hcl
locals {
  name_prefix = "${var.project_name}-${var.environment}"
  common_tags = { Environment = var.environment, Project = var.project_name, ManagedBy = "terraform" }
}
```

### Step 5: Module Templates

**modules/networking/main.tf:**
```hcl
resource "aws_vpc" "main" {
  cidr_block           = var.vpc_cidr
  enable_dns_hostnames = true
  enable_dns_support   = true
  tags = merge(var.tags, { Name = "${var.name_prefix}-vpc" })
}
resource "aws_subnet" "private" {
  count             = length(var.private_subnet_cidrs)
  vpc_id            = aws_vpc.main.id
  cidr_block        = var.private_subnet_cidrs[count.index]
  availability_zone = var.availability_zones[count.index]
  tags = merge(var.tags, { Name = "${var.name_prefix}-private-${count.index}" })
}
resource "aws_subnet" "public" {
  count                   = length(var.public_subnet_cidrs)
  vpc_id                  = aws_vpc.main.id
  cidr_block              = var.public_subnet_cidrs[count.index]
  availability_zone       = var.availability_zones[count.index]
  map_public_ip_on_launch = true
  tags = merge(var.tags, { Name = "${var.name_prefix}-public-${count.index}" })
}
resource "aws_internet_gateway" "main" {
  vpc_id = aws_vpc.main.id
  tags   = merge(var.tags, { Name = "${var.name_prefix}-igw" })
}
resource "aws_eip" "nat" { domain = "vpc" }
resource "aws_nat_gateway" "main" {
  allocation_id = aws_eip.nat.id
  subnet_id     = aws_subnet.public[0].id
  depends_on    = [aws_internet_gateway.main]
}
```

**modules/networking/variables.tf:**
```hcl
variable "name_prefix"         { type = string }
variable "vpc_cidr"            { type = string; default = "10.0.0.0/16" }
variable "availability_zones"  { type = list(string) }
variable "private_subnet_cidrs" { type = list(string); default = ["10.0.1.0/24","10.0.2.0/24","10.0.3.0/24"] }
variable "public_subnet_cidrs"  { type = list(string); default = ["10.0.101.0/24","10.0.102.0/24","10.0.103.0/24"] }
variable "tags"                 { type = map(string); default = {} }
```

**modules/networking/outputs.tf:**
```hcl
output "vpc_id"             { value = aws_vpc.main.id }
output "private_subnet_ids" { value = aws_subnet.private[*].id }
output "public_subnet_ids"  { value = aws_subnet.public[*].id }
output "nat_gateway_ip"     { value = aws_eip.nat.public_ip }
```

**modules/compute/main.tf (ECS Fargate):**
```hcl
resource "aws_ecs_cluster" "main" {
  name = "${var.name_prefix}-cluster"
  setting { name = "containerInsights"; value = "enabled" }
}
resource "aws_ecs_task_definition" "app" {
  family                   = "${var.name_prefix}-app"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = var.cpu
  memory                   = var.memory
  execution_role_arn       = aws_iam_role.execution.arn
  container_definitions = jsonencode([{
    name = "app", image = var.container_image, essential = true
    portMappings = [{ containerPort = var.container_port, protocol = "tcp" }]
    logConfiguration = {
      logDriver = "awslogs"
      options   = { "awslogs-group" = aws_cloudwatch_log_group.app.name, "awslogs-region" = data.aws_region.current.name, "awslogs-stream-prefix" = "ecs" }
    }
  }])
}
resource "aws_ecs_service" "app" {
  name            = "${var.name_prefix}-svc"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.app.arn
  desired_count   = var.desired_count
  launch_type     = "FARGATE"
  network_configuration { subnets = var.private_subnet_ids; security_groups = [aws_security_group.ecs.id] }
}
resource "aws_cloudwatch_log_group" "app" { name = "/ecs/${var.name_prefix}"; retention_in_days = 30 }
resource "aws_security_group" "ecs" {
  name_prefix = "${var.name_prefix}-ecs-"
  vpc_id      = var.vpc_id
  egress { from_port = 0; to_port = 0; protocol = "-1"; cidr_blocks = ["0.0.0.0/0"] }
  lifecycle { create_before_destroy = true }
}
resource "aws_iam_role" "execution" {
  name = "${var.name_prefix}-exec"
  assume_role_policy = jsonencode({ Version = "2012-10-17", Statement = [{ Action = "sts:AssumeRole", Effect = "Allow", Principal = { Service = "ecs-tasks.amazonaws.com" } }] })
}
resource "aws_iam_role_policy_attachment" "execution" {
  role       = aws_iam_role.execution.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}
data "aws_region" "current" {}
```

**modules/compute/variables.tf & outputs.tf:**
```hcl
# variables.tf
variable "name_prefix"        { type = string }
variable "vpc_id"             { type = string }
variable "private_subnet_ids" { type = list(string) }
variable "container_image"    { type = string }
variable "container_port"     { type = number; default = 8080 }
variable "cpu"                { type = number; default = 256 }
variable "memory"             { type = number; default = 512 }
variable "desired_count"      { type = number; default = 2 }
variable "tags"               { type = map(string); default = {} }
# outputs.tf
output "cluster_id"        { value = aws_ecs_cluster.main.id }
output "service_name"      { value = aws_ecs_service.app.name }
output "security_group_id" { value = aws_security_group.ecs.id }
```

**modules/data/main.tf (RDS PostgreSQL):**
```hcl
resource "aws_db_subnet_group" "main" { name = "${var.name_prefix}-db"; subnet_ids = var.private_subnet_ids }
resource "aws_db_instance" "main" {
  identifier = "${var.name_prefix}-pg"
  engine = "postgres"; engine_version = var.engine_version; instance_class = var.instance_class
  allocated_storage = var.allocated_storage; db_name = "app"; username = "dbadmin"
  manage_master_user_password = true
  db_subnet_group_name = aws_db_subnet_group.main.name; vpc_security_group_ids = [aws_security_group.db.id]
  backup_retention_period = 7; storage_encrypted = true; skip_final_snapshot = false
  final_snapshot_identifier = "${var.name_prefix}-final"
}
resource "aws_security_group" "db" {
  name_prefix = "${var.name_prefix}-db-"; vpc_id = var.vpc_id
  ingress { from_port = 5432; to_port = 5432; protocol = "tcp"; cidr_blocks = [data.aws_vpc.selected.cidr_block] }
  lifecycle { create_before_destroy = true }
}
data "aws_vpc" "selected" { id = var.vpc_id }
```

**Root main.tf wiring modules:**
```hcl
module "networking" {
  source = "./modules/networking"
  name_prefix = local.name_prefix; vpc_cidr = var.vpc_cidr; availability_zones = var.availability_zones; tags = local.common_tags
}
module "compute" {
  source = "./modules/compute"
  name_prefix = local.name_prefix; vpc_id = module.networking.vpc_id; private_subnet_ids = module.networking.private_subnet_ids
  container_image = var.container_image; container_port = var.container_port; desired_count = var.desired_count; tags = local.common_tags
}
module "data" {
  source = "./modules/data"
  name_prefix = local.name_prefix; vpc_id = module.networking.vpc_id; private_subnet_ids = module.networking.private_subnet_ids; tags = local.common_tags
}
```

**environments/dev.tfvars:**
```hcl
environment = "dev"; project_name = "myproject"; aws_region = "us-east-1"; owner = "platform-team"
vpc_cidr = "10.0.0.0/16"; container_image = "nginx:latest"; desired_count = 1
```

### Step 6: CI/CD Pipeline

**GitHub Actions (default):**
```yaml
# .github/workflows/terraform.yml
name: Terraform
on:
  push: { branches: [main] }
  pull_request: { branches: [main] }
permissions: { contents: read, pull-requests: write, id-token: write }
env:
  TF_VERSION: "1.9.8"
  TF_VAR_FILE: "environments/dev.tfvars"
jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: hashicorp/setup-terraform@v3
        with: { terraform_version: "${{ env.TF_VERSION }}" }
      - run: terraform fmt -check -recursive
      - run: terraform init -backend=false
      - run: terraform validate
  tflint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: terraform-linters/setup-tflint@v4
      - run: tflint --init && tflint --recursive
  plan:
    needs: [validate, tflint]
    if: github.event_name == 'pull_request'
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: aws-actions/configure-aws-credentials@v4
        with: { role-to-assume: "${{ secrets.AWS_ROLE_ARN }}", aws-region: us-east-1 }
      - uses: hashicorp/setup-terraform@v3
        with: { terraform_version: "${{ env.TF_VERSION }}" }
      - run: terraform init
      - id: plan
        run: terraform plan -var-file=${{ env.TF_VAR_FILE }} -no-color -out=tfplan
      - uses: actions/github-script@v7
        with:
          script: |
            github.rest.issues.createComment({
              issue_number: context.issue.number, owner: context.repo.owner, repo: context.repo.repo,
              body: `#### Terraform Plan\n\`\`\`\n${{ steps.plan.outputs.stdout }}\n\`\`\``
            });
  apply:
    needs: [validate, tflint]
    if: github.ref == 'refs/heads/main' && github.event_name == 'push'
    runs-on: ubuntu-latest
    environment: production
    steps:
      - uses: actions/checkout@v4
      - uses: aws-actions/configure-aws-credentials@v4
        with: { role-to-assume: "${{ secrets.AWS_ROLE_ARN }}", aws-region: us-east-1 }
      - uses: hashicorp/setup-terraform@v3
        with: { terraform_version: "${{ env.TF_VERSION }}" }
      - run: terraform init
      - run: terraform apply -var-file=${{ env.TF_VAR_FILE }} -auto-approve
```

**Atlantis config:**
```yaml
# atlantis.yaml
version: 3
projects:
  - name: dev
    dir: .
    workspace: dev
    terraform_version: v1.9.8
    autoplan:
      when_modified: ["*.tf", "modules/**/*.tf", "environments/dev.tfvars"]
    apply_requirements: [mergeable, approved]
    workflow: standard
workflows:
  standard:
    plan:
      steps: [init, { plan: { extra_args: ["-var-file", "environments/$WORKSPACE.tfvars"] } }]
    apply:
      steps: [apply]
```

### Step 7: Development Tooling

**.terraform-version:** `1.9.8`

**.tflint.hcl:**
```hcl
plugin "terraform" { enabled = true; preset = "recommended" }
plugin "aws" { enabled = true; version = "0.35.0"; source = "github.com/terraform-linters/tflint-ruleset-aws" }
rule "terraform_naming_convention"    { enabled = true }
rule "terraform_documented_variables" { enabled = true }
rule "terraform_documented_outputs"   { enabled = true }
```

**.pre-commit-config.yaml:**
```yaml
repos:
  - repo: https://github.com/antonbabenko/pre-commit-tf
    rev: v1.96.1
    hooks: [{ id: terraform_fmt }, { id: terraform_validate }, { id: terraform_tflint }, { id: terraform_trivy }]
  - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v5.0.0
    hooks: [{ id: check-merge-conflict }, { id: end-of-file-fixer }, { id: trailing-whitespace }]
```

**Makefile:**
```makefile
.PHONY: init plan apply destroy fmt lint validate clean
ENV ?= dev
TF_VAR_FILE := environments/$(ENV).tfvars

init:      terraform init
plan:      terraform plan -var-file=$(TF_VAR_FILE)
apply:     terraform apply -var-file=$(TF_VAR_FILE)
destroy:   terraform destroy -var-file=$(TF_VAR_FILE)
fmt:       terraform fmt -recursive
lint:      tflint --init && tflint --recursive
validate:  terraform validate
clean:     rm -rf .terraform .terraform.lock.hcl tfplan
plan-dev:  $(MAKE) plan ENV=dev
plan-prod: $(MAKE) plan ENV=production
```

**.gitignore:**
```gitignore
.terraform/
*.tfstate
*.tfstate.*
*.tfplan
tfplan
crash.log
override.tf
override.tf.json
*_override.tf
*_override.tf.json
.terraformrc
terraform.rc
.terraform.lock.hcl
.DS_Store
```

### Step 8: Verification

Run these commands to verify the scaffolded project:
```bash
terraform init -backend=false
terraform fmt -check -recursive
terraform validate
tflint --init && tflint --recursive
```

## Checklist

After scaffolding, verify:
- [ ] `terraform init -backend=false` succeeds
- [ ] `terraform validate` passes
- [ ] `terraform fmt -check -recursive` shows no issues
- [ ] All modules have `variables.tf`, `main.tf`, and `outputs.tf`
- [ ] Backend config matches selected state backend
- [ ] Provider versions use `~>` pessimistic operator
- [ ] Environment tfvars exist for each target environment
- [ ] CI/CD runs `fmt`, `validate`, `plan` on PRs and `apply` on merge
- [ ] `.gitignore` excludes `.terraform/`, state files, and plan files
- [ ] No hardcoded credentials or account IDs in any file
- [ ] Pre-commit hooks configured for `fmt`, `validate`, and `tflint`
- [ ] Makefile provides shortcuts for common operations
