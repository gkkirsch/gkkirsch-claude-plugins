# Infrastructure as Code Best Practices Reference

Patterns, anti-patterns, and guidelines for managing infrastructure as code across Terraform, CloudFormation, Pulumi, and CDK.

---

## Foundational Principles

### 1. Everything as Code

```
✅ DO:
  - Define ALL infrastructure in code (no manual console changes)
  - Store IaC in version control (Git)
  - Use pull requests for infrastructure changes
  - Require code review for production changes
  - Tag resources as managed by IaC tool

✗ DON'T:
  - Create resources manually then try to import
  - Use ClickOps for "quick fixes"
  - Store state files in version control
  - Hardcode account IDs, regions, or ARNs
```

### 2. Idempotency

Every apply should produce the same result regardless of how many times it runs. If infrastructure already exists in the desired state, no changes should occur.

```hcl
# Good — idempotent
resource "aws_s3_bucket" "data" {
  bucket = "myapp-data-${var.environment}"
}

# Bad — not idempotent (creates new resource each time)
resource "random_id" "bucket_suffix" {
  byte_length = 8
  keepers = {
    timestamp = timestamp()  # Changes every run!
  }
}
```

### 3. Immutability

Prefer replacing resources over modifying them in-place.

```hcl
# Immutable deployment — new launch template, rolling replacement
resource "aws_launch_template" "app" {
  name_prefix   = "app-"
  image_id      = var.ami_id

  lifecycle {
    create_before_destroy = true
  }
}

# Mutable anti-pattern — modifying running instances
# (Avoid: SSH + scripts to update running servers)
```

---

## Repository Structure Patterns

### Pattern 1: Monorepo (Recommended for Small-Medium Teams)

```
infrastructure/
├── modules/              # Shared, reusable modules
│   ├── vpc/
│   ├── ecs-service/
│   ├── rds/
│   └── monitoring/
├── environments/
│   ├── dev/
│   │   ├── main.tf
│   │   ├── variables.tf
│   │   └── terraform.tfvars
│   ├── staging/
│   └── production/
├── global/               # Cross-environment resources
│   ├── iam/
│   ├── dns/
│   └── budgets/
├── .github/
│   └── workflows/
│       └── terraform.yml
├── .tflint.hcl
├── .pre-commit-config.yaml
└── README.md
```

**Pros:** Single source of truth, easy cross-referencing, atomic changes.
**Cons:** Blast radius (all environments in one repo), longer CI runs.

### Pattern 2: Polyrepo (Recommended for Large Teams)

```
# Separate repos:
infrastructure-modules/     # Shared module library (versioned)
infrastructure-dev/         # Dev environment
infrastructure-staging/     # Staging environment
infrastructure-production/  # Production environment
```

**Pros:** Isolated blast radius, independent deployment, team ownership.
**Cons:** Module versioning complexity, cross-repo changes are harder.

### Pattern 3: Environment-per-Directory with Terragrunt

```
infrastructure/
├── terragrunt.hcl          # Root: remote state config, provider
├── _envcommon/
│   ├── vpc.hcl
│   ├── ecs.hcl
│   └── rds.hcl
├── dev/
│   ├── env.hcl             # Environment variables
│   └── us-east-1/
│       ├── region.hcl
│       ├── vpc/
│       │   └── terragrunt.hcl  # include + inputs
│       └── ecs/
│           └── terragrunt.hcl
├── staging/
└── production/
```

---

## State Management Best Practices

### Remote State Rules

```
1. ALWAYS use remote state in team environments
2. ALWAYS enable state locking (DynamoDB for S3, built-in for GCS/Azure)
3. ALWAYS encrypt state at rest (KMS for S3, default for GCS)
4. ALWAYS enable versioning on state bucket
5. NEVER commit state files to Git
6. NEVER share state files via email/Slack
7. NEVER manually edit state files
8. LIMIT access to state (contains secrets in plaintext!)
```

### State Isolation Strategies

```
Strategy 1: File-per-environment
  s3://state-bucket/dev/terraform.tfstate
  s3://state-bucket/staging/terraform.tfstate
  s3://state-bucket/production/terraform.tfstate

Strategy 2: Bucket-per-environment
  s3://state-dev/terraform.tfstate
  s3://state-staging/terraform.tfstate
  s3://state-production/terraform.tfstate

Strategy 3: Account-per-environment (most isolated)
  Account 111: s3://state/terraform.tfstate  (dev)
  Account 222: s3://state/terraform.tfstate  (staging)
  Account 333: s3://state/terraform.tfstate  (production)

Strategy 4: Component-level isolation
  s3://state/networking/terraform.tfstate
  s3://state/compute/terraform.tfstate
  s3://state/database/terraform.tfstate
```

### When to Split State

Split state when:
- Different teams own different components
- Components have different change frequencies
- You want to limit blast radius
- Plan/apply times become too long (> 5 minutes)
- You need different access controls per component

Keep together when:
- Resources are tightly coupled (VPC + subnets)
- Changes always happen together
- Cross-state references would be complex

---

## Module Design Guidelines

### Module Interface Design

```
Good module interface:
  ✅ Small number of required variables (< 5)
  ✅ Sensible defaults for optional variables
  ✅ Input validation with clear error messages
  ✅ Outputs for everything consumers might need
  ✅ Documentation (README with examples)
  ✅ Semantic versioning

Bad module interface:
  ✗ Dozens of required variables
  ✗ No defaults (forces consumers to specify everything)
  ✗ No validation (fails with cryptic provider errors)
  ✗ Missing outputs (forces consumers to use data sources)
  ✗ No documentation
```

### Module Versioning

```hcl
# Pin to exact version in production
module "vpc" {
  source  = "git::https://github.com/myorg/terraform-modules.git//vpc?ref=v2.3.1"
}

# Pin to minor version for auto-patches
module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "~> 5.5"  # >= 5.5.0, < 5.6.0
}

# Never do this in production:
module "vpc" {
  source = "git::https://github.com/myorg/terraform-modules.git//vpc"  # No version!
}
```

### Module Composition vs Inheritance

```
Prefer COMPOSITION (small, focused modules combined):

module "network"  → VPC, subnets, NAT, flow logs
module "security" → Security groups, NACLs, WAF
module "compute"  → ECS cluster, task definitions, services
module "database" → RDS instance, parameter group, subnet group
module "monitor"  → CloudWatch alarms, dashboards, SNS topics

Avoid INHERITANCE (giant module that does everything):

module "application" → VPC + security + ECS + RDS + monitoring
  (too many variables, hard to test, hard to reuse partially)
```

---

## Security Practices

### Secrets in IaC

```
NEVER store secrets in:
  ✗ terraform.tfvars
  ✗ Environment variables in Terraform config
  ✗ Hardcoded in .tf files
  ✗ Git history (even if later removed)

ALWAYS use:
  ✅ Secrets Manager / Vault data sources
  ✅ OIDC for CI/CD authentication (no long-lived keys)
  ✅ IAM roles (not access keys)
  ✅ Terraform's sensitive = true for variables
  ✅ .gitignore for .tfvars files with secrets
```

```hcl
# Good — read from Secrets Manager
data "aws_secretsmanager_secret_version" "db" {
  secret_id = "production/database"
}

# Good — mark variable as sensitive
variable "api_key" {
  type      = string
  sensitive = true
}

# Good — use managed password rotation
resource "aws_db_instance" "main" {
  manage_master_user_password = true
}
```

### Policy as Code

**Sentinel (Terraform Cloud/Enterprise):**

```python
# Enforce encryption on all S3 buckets
import "tfplan/v2" as tfplan

s3_buckets = filter tfplan.resource_changes as _, rc {
  rc.type is "aws_s3_bucket" and
  rc.mode is "managed" and
  (rc.change.actions contains "create" or rc.change.actions contains "update")
}

main = rule {
  all s3_buckets as _, bucket {
    bucket.change.after.server_side_encryption_configuration is not null
  }
}
```

**Open Policy Agent (OPA) / Conftest:**

```rego
# policy/terraform.rego
package terraform

deny[msg] {
  resource := input.resource_changes[_]
  resource.type == "aws_s3_bucket"
  not resource.change.after.server_side_encryption_configuration
  msg := sprintf("S3 bucket '%s' must have encryption enabled", [resource.address])
}

deny[msg] {
  resource := input.resource_changes[_]
  resource.type == "aws_security_group_rule"
  resource.change.after.cidr_blocks[_] == "0.0.0.0/0"
  resource.change.after.from_port == 22
  msg := sprintf("Security group '%s' must not allow SSH from 0.0.0.0/0", [resource.address])
}

deny[msg] {
  resource := input.resource_changes[_]
  resource.type == "aws_instance"
  not resource.change.after.metadata_options[0].http_tokens == "required"
  msg := sprintf("EC2 instance '%s' must require IMDSv2", [resource.address])
}
```

---

## CI/CD Patterns

### GitOps for Infrastructure

```
Developer → PR → Plan → Review → Approve → Merge → Apply

Branch Protection Rules:
  - Require PR reviews (2+ for production)
  - Require passing CI (format, validate, plan, security scan)
  - Require up-to-date branch
  - No force push
  - No direct commits to main
```

### Pipeline Stages

```
┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐
│  Format  │→│ Validate │→│ Security │→│   Plan   │→│  Apply   │
│  Check   │  │          │  │  Scan    │  │          │  │          │
│          │  │          │  │          │  │          │  │          │
│ terraform│  │ terraform│  │ tfsec    │  │ terraform│  │ terraform│
│ fmt -check│  │ validate │  │ checkov  │  │ plan     │  │ apply    │
│          │  │          │  │ trivy    │  │ infracost│  │          │
└──────────┘  └──────────┘  └──────────┘  └──────────┘  └──────────┘
   Always       Always        Always        PR only      Main only
                                           (comment)    (auto/manual)
```

### Pre-commit Hooks

```yaml
# .pre-commit-config.yaml
repos:
  - repo: https://github.com/antonbabenko/pre-commit-terraform
    rev: v1.86.0
    hooks:
      - id: terraform_fmt
      - id: terraform_validate
      - id: terraform_tflint
        args:
          - --args=--config=__GIT_WORKING_DIR__/.tflint.hcl
      - id: terraform_docs
        args:
          - --hook-config=--path-to-file=README.md
          - --hook-config=--add-to-existing-file=true
          - --hook-config=--create-file-if-not-exist=true
      - id: terraform_tfsec

  - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v4.5.0
    hooks:
      - id: check-merge-conflict
      - id: end-of-file-fixer
      - id: trailing-whitespace
      - id: no-commit-to-branch
        args: ['--branch', 'main']
```

---

## Testing Strategies

### Testing Pyramid for IaC

```
                    /\
                   /  \
                  / E2E \          Integration tests
                 / Tests  \        (real cloud resources)
                /──────────\
               /  Contract   \     Module interface tests
              /    Tests      \    (plan-based validation)
             /────────────────\
            /   Unit Tests     \   Static analysis
           /  (format, lint,    \  (tfsec, checkov, OPA)
          /    validate, plan)   \
         /────────────────────────\
```

### Static Analysis Tools

| Tool | Purpose | Speed |
|------|---------|-------|
| `terraform fmt` | Code formatting | Instant |
| `terraform validate` | Syntax and type checking | Instant |
| `tflint` | Linting, provider-specific checks | Fast |
| `tfsec` / `trivy` | Security scanning | Fast |
| `checkov` | Security + compliance | Fast |
| `infracost` | Cost estimation | Medium |
| `conftest` | Custom OPA policies | Fast |

### Plan-Based Testing

```bash
# Generate plan as JSON for testing
terraform plan -out=tfplan
terraform show -json tfplan > plan.json

# Test with conftest
conftest test plan.json --policy policy/

# Test with custom scripts
jq '.resource_changes[] | select(.type == "aws_s3_bucket") | .change.after.server_side_encryption_configuration' plan.json
```

---

## CloudFormation Patterns

### Nested Stacks

```yaml
AWSTemplateFormatVersion: '2010-09-09'
Description: Root stack

Resources:
  NetworkingStack:
    Type: AWS::CloudFormation::Stack
    Properties:
      TemplateURL: https://s3.amazonaws.com/templates/networking.yaml
      Parameters:
        VpcCidr: 10.0.0.0/16
        Environment: production

  ComputeStack:
    Type: AWS::CloudFormation::Stack
    DependsOn: NetworkingStack
    Properties:
      TemplateURL: https://s3.amazonaws.com/templates/compute.yaml
      Parameters:
        VpcId: !GetAtt NetworkingStack.Outputs.VpcId
        SubnetIds: !GetAtt NetworkingStack.Outputs.PrivateSubnetIds

  DatabaseStack:
    Type: AWS::CloudFormation::Stack
    DependsOn: NetworkingStack
    Properties:
      TemplateURL: https://s3.amazonaws.com/templates/database.yaml
      Parameters:
        VpcId: !GetAtt NetworkingStack.Outputs.VpcId
        SubnetIds: !GetAtt NetworkingStack.Outputs.DatabaseSubnetIds
        AppSecurityGroup: !GetAtt ComputeStack.Outputs.AppSecurityGroupId
```

### CloudFormation vs Terraform

| Feature | CloudFormation | Terraform |
|---------|---------------|-----------|
| Cloud support | AWS only | Multi-cloud |
| State | AWS-managed | Self-managed |
| Language | YAML/JSON | HCL |
| Drift detection | Built-in | `plan -detailed-exitcode` |
| Rollback | Automatic on failure | Manual |
| Modularity | Nested stacks, modules | Modules (better) |
| Testing | cfn-lint, taskcat | terraform test, terratest |
| Import | Yes (drift detection) | Yes (import blocks) |
| Cost | Free | Free (OSS), paid (Cloud) |

---

## Pulumi Patterns

```typescript
import * as aws from "@pulumi/aws";
import * as pulumi from "@pulumi/pulumi";

const config = new pulumi.Config();
const environment = config.require("environment");

// VPC with standard networking
const vpc = new aws.ec2.Vpc("main", {
  cidrBlock: "10.0.0.0/16",
  enableDnsHostnames: true,
  enableDnsSupport: true,
  tags: { Environment: environment },
});

// Create subnets programmatically
const azs = ["us-east-1a", "us-east-1b", "us-east-1c"];
const privateSubnets = azs.map((az, i) => new aws.ec2.Subnet(`private-${az}`, {
  vpcId: vpc.id,
  cidrBlock: `10.0.${i + 1}.0/24`,
  availabilityZone: az,
  tags: { Name: `private-${az}`, Tier: "private" },
}));

// Type-safe outputs
export const vpcId = vpc.id;
export const subnetIds = privateSubnets.map(s => s.id);
```

### Pulumi vs Terraform

| Feature | Pulumi | Terraform |
|---------|--------|-----------|
| Language | TypeScript, Python, Go, C#, Java | HCL |
| IDE support | Full (autocomplete, types) | Limited |
| Logic | Native (loops, conditions, classes) | HCL expressions |
| Testing | Native test frameworks | terraform test, terratest |
| State | Pulumi Cloud or self-managed | S3, GCS, etc. |
| Learning curve | Know a language = easy | Learn HCL |
| Ecosystem | Growing | Massive |
| Maturity | Newer | Established |

---

## Anti-Patterns to Avoid

### 1. ClickOps + Import
Creating resources manually then importing into IaC. This leads to drift, undocumented configurations, and missing dependencies.

### 2. Mega-Module
One module that creates an entire application stack. Hard to test, review, reuse, or debug.

### 3. Copy-Paste Environments
Duplicating entire directories for each environment instead of using variables, workspaces, or Terragrunt.

### 4. State Monolith
All resources in a single state file. Slow plans, large blast radius, team conflicts.

### 5. Implicit Dependencies
Relying on apply ordering instead of explicit `depends_on` or data source references.

### 6. Ignoring Drift
Not regularly checking for manual changes. Use scheduled drift detection in CI.

### 7. No Lifecycle Rules
Not using `prevent_destroy` on critical resources or `create_before_destroy` for zero-downtime updates.

### 8. Hardcoded Everything
Account IDs, region names, AMI IDs, subnet IDs directly in configurations instead of using variables and data sources.

### 9. No Cost Visibility
Not estimating costs before apply. Use Infracost to catch expensive changes in PR reviews.

### 10. Missing Rollback Plan
Not knowing how to recover when an apply goes wrong. Always have a plan: rollback commit, restore state, or manual intervention.

---

## Migration Checklist

When migrating from manual/console management to IaC:

```
Phase 1: Foundation
  □ Set up remote state backend (S3/GCS/Azure Storage)
  □ Configure CI/CD pipeline for IaC
  □ Establish module library (or use registry)
  □ Define tagging and naming conventions
  □ Set up security scanning in pipeline

Phase 2: Import Existing Resources
  □ Inventory all existing resources
  □ Write Terraform configurations
  □ Import resources into state
  □ Verify plan shows no changes
  □ Test in non-production first

Phase 3: Governance
  □ Enable drift detection (scheduled CI)
  □ Set up policy-as-code (OPA/Sentinel)
  □ Create runbooks for common operations
  □ Train team on IaC workflow
  □ Document module usage and conventions

Phase 4: Optimization
  □ Refactor into reusable modules
  □ Add cost estimation to PR workflow
  □ Implement automated testing
  □ Set up module versioning and changelog
  □ Review and optimize state boundaries
```

---

## Key Takeaways

1. **Code is the source of truth.** If it's not in code, it doesn't exist.
2. **State is sacred.** Protect it, encrypt it, lock it, back it up.
3. **Small modules, small state.** Limit blast radius, increase velocity.
4. **Test before apply.** Format, validate, lint, scan, plan, estimate — then apply.
5. **Version everything.** Providers, modules, and tools. Pin in production.
6. **Automate the pipeline.** PR → plan → review → apply. No manual applies.
7. **Detect drift.** Scheduled checks catch console cowboys.
8. **Separate secrets.** Never in code, always in secret managers.
9. **Document decisions.** ADRs for why, READMEs for how.
10. **Start simple.** Add complexity only when justified by team size or requirements.
