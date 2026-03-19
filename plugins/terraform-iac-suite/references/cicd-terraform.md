# CI/CD for Terraform

Complete guide to automating Terraform workflows in continuous integration and continuous delivery pipelines. This reference covers GitHub Actions, GitLab CI, Atlantis, Terraform Cloud, cost estimation, security scanning, testing, and production-ready patterns for teams managing infrastructure as code at scale. Every example uses real, working configuration that can be adapted to your environment.

---

## GitHub Actions for Terraform

### Basic Plan/Apply Workflow

A minimal but functional workflow that runs `terraform plan` on pull requests and `terraform apply` on merges to main.

```yaml
# .github/workflows/terraform.yml
name: Terraform

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

permissions:
  contents: read
  pull-requests: write

env:
  TF_VERSION: "1.7.4"
  TF_WORKING_DIR: "infrastructure"

jobs:
  terraform:
    name: Terraform Plan & Apply
    runs-on: ubuntu-latest
    defaults:
      run:
        working-directory: ${{ env.TF_WORKING_DIR }}

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Terraform
        uses: hashicorp/setup-terraform@v3
        with:
          terraform_version: ${{ env.TF_VERSION }}
          terraform_wrapper: true  # enables stdout capture for plan output

      - name: Terraform Init
        id: init
        run: terraform init -input=false

      - name: Terraform Validate
        id: validate
        run: terraform validate -no-color

      - name: Terraform Plan
        id: plan
        if: github.event_name == 'pull_request'
        run: terraform plan -no-color -input=false -out=tfplan
        continue-on-error: true

      - name: Terraform Apply
        if: github.ref == 'refs/heads/main' && github.event_name == 'push'
        run: terraform apply -auto-approve -input=false
```

### PR Comment with Plan Output

Post the Terraform plan as a comment on the pull request so reviewers can see exactly what will change.

```yaml
      # Add this step after the plan step in the workflow above
      - name: Comment Plan on PR
        if: github.event_name == 'pull_request'
        uses: actions/github-script@v7
        env:
          PLAN: ${{ steps.plan.outputs.stdout }}
        with:
          github-token: ${{ secrets.GITHUB_TOKEN }}
          script: |
            const output = `#### Terraform Format and Style: \`${{ steps.fmt.outcome }}\`
            #### Terraform Initialization: \`${{ steps.init.outcome }}\`
            #### Terraform Validation: \`${{ steps.validate.outcome }}\`
            #### Terraform Plan: \`${{ steps.plan.outcome }}\`

            <details><summary>Show Plan Output</summary>

            \`\`\`terraform
            ${process.env.PLAN}
            \`\`\`

            </details>

            *Pushed by: @${{ github.actor }}, Action: \`${{ github.event_name }}\`*`;

            // Find existing bot comment and update it, or create new one
            const { data: comments } = await github.rest.issues.listComments({
              owner: context.repo.owner,
              repo: context.repo.repo,
              issue_number: context.issue.number,
            });

            const botComment = comments.find(comment =>
              comment.user.type === 'Bot' &&
              comment.body.includes('Terraform Plan')
            );

            if (botComment) {
              await github.rest.issues.updateComment({
                owner: context.repo.owner,
                repo: context.repo.repo,
                comment_id: botComment.id,
                body: output,
              });
            } else {
              await github.rest.issues.createComment({
                owner: context.repo.owner,
                repo: context.repo.repo,
                issue_number: context.issue.number,
                body: output,
              });
            }
```

### Environment-Specific Deploys (dev, staging, production)

Use GitHub Actions environments with separate state files and variable sets per environment.

```yaml
# .github/workflows/terraform-environments.yml
name: Terraform Multi-Environment

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  plan:
    name: Plan - ${{ matrix.environment }}
    runs-on: ubuntu-latest
    strategy:
      matrix:
        environment: [dev, staging, production]
    defaults:
      run:
        working-directory: infrastructure

    steps:
      - uses: actions/checkout@v4
      - uses: hashicorp/setup-terraform@v3
        with:
          terraform_version: "1.7.4"

      - name: Terraform Init
        run: |
          terraform init \
            -backend-config="environments/${{ matrix.environment }}/backend.tfvars" \
            -input=false

      - name: Terraform Plan
        run: |
          terraform plan \
            -var-file="environments/${{ matrix.environment }}/terraform.tfvars" \
            -out=tfplan-${{ matrix.environment }} \
            -input=false

      - name: Upload Plan Artifact
        uses: actions/upload-artifact@v4
        with:
          name: tfplan-${{ matrix.environment }}
          path: infrastructure/tfplan-${{ matrix.environment }}
          retention-days: 5

  deploy-dev:
    name: Apply - dev
    needs: plan
    if: github.ref == 'refs/heads/develop' || github.ref == 'refs/heads/main'
    runs-on: ubuntu-latest
    environment: dev
    defaults:
      run:
        working-directory: infrastructure
    steps:
      - uses: actions/checkout@v4
      - uses: hashicorp/setup-terraform@v3
        with:
          terraform_version: "1.7.4"
      - uses: actions/download-artifact@v4
        with:
          name: tfplan-dev
          path: infrastructure

      - name: Terraform Init
        run: terraform init -backend-config="environments/dev/backend.tfvars" -input=false

      - name: Terraform Apply
        run: terraform apply -auto-approve -input=false tfplan-dev

  deploy-staging:
    name: Apply - staging
    needs: deploy-dev
    if: github.ref == 'refs/heads/main'
    runs-on: ubuntu-latest
    environment: staging
    defaults:
      run:
        working-directory: infrastructure
    steps:
      - uses: actions/checkout@v4
      - uses: hashicorp/setup-terraform@v3
        with:
          terraform_version: "1.7.4"
      - uses: actions/download-artifact@v4
        with:
          name: tfplan-staging
          path: infrastructure

      - name: Terraform Init
        run: terraform init -backend-config="environments/staging/backend.tfvars" -input=false

      - name: Terraform Apply
        run: terraform apply -auto-approve -input=false tfplan-staging

  deploy-production:
    name: Apply - production
    needs: deploy-staging
    if: github.ref == 'refs/heads/main'
    runs-on: ubuntu-latest
    environment: production  # requires manual approval configured in repo settings
    defaults:
      run:
        working-directory: infrastructure
    steps:
      - uses: actions/checkout@v4
      - uses: hashicorp/setup-terraform@v3
        with:
          terraform_version: "1.7.4"
      - uses: actions/download-artifact@v4
        with:
          name: tfplan-production
          path: infrastructure

      - name: Terraform Init
        run: terraform init -backend-config="environments/production/backend.tfvars" -input=false

      - name: Terraform Apply
        run: terraform apply -auto-approve -input=false tfplan-production
```

### Manual Approval Gates

Configure GitHub Environments to require manual approval before production deploys. This is set in the repository settings, but the workflow references it by name.

```yaml
  # The 'environment: production' key triggers the approval gate
  deploy-production:
    name: Apply - production
    needs: deploy-staging
    runs-on: ubuntu-latest
    environment:
      name: production
      url: https://console.aws.amazon.com  # optional: link to deployed environment
    steps:
      # Steps run only after a designated reviewer approves in the GitHub UI
      - uses: actions/checkout@v4
      # ... remaining apply steps
```

### Terraform Format and Validate Checks

Run formatting and validation as a fast first gate on every PR.

```yaml
# .github/workflows/terraform-checks.yml
name: Terraform Checks

on:
  pull_request:
    paths:
      - '**.tf'
      - '**.tfvars'

jobs:
  checks:
    name: Format & Validate
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: hashicorp/setup-terraform@v3
        with:
          terraform_version: "1.7.4"

      - name: Terraform Format Check
        id: fmt
        run: terraform fmt -check -recursive -diff
        continue-on-error: true

      - name: Terraform Init (validate needs providers)
        run: terraform init -backend=false

      - name: Terraform Validate
        id: validate
        run: terraform validate -no-color

      - name: Fail if format check failed
        if: steps.fmt.outcome == 'failure'
        run: |
          echo "::error::Terraform files are not formatted. Run 'terraform fmt -recursive' locally."
          exit 1
```

### Security Scanning in CI (tfsec, checkov)

Integrate static analysis tools that catch security misconfigurations before they reach production.

```yaml
# .github/workflows/terraform-security.yml
name: Terraform Security Scan

on:
  pull_request:
    paths: ['**.tf']

jobs:
  tfsec:
    name: tfsec Security Scan
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Run tfsec
        uses: aquasecurity/tfsec-action@v1.0.3
        with:
          working_directory: infrastructure
          soft_fail: false  # set true to warn without blocking
          format: sarif
          additional_args: "--minimum-severity HIGH"

      - name: Upload SARIF results
        if: always()
        uses: github/codeql-action/upload-sarif@v3
        with:
          sarif_file: results.sarif

  checkov:
    name: Checkov Policy Scan
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Run Checkov
        uses: bridgecrewio/checkov-action@v12
        with:
          directory: infrastructure
          framework: terraform
          output_format: cli,sarif
          output_file_path: console,checkov-results.sarif
          soft_fail: false
          skip_check: CKV_AWS_18,CKV_AWS_21  # skip specific checks if needed
          quiet: true  # only show failed checks
```

### OIDC Authentication (No Stored AWS Keys)

Use OpenID Connect so GitHub Actions can assume an AWS IAM role without storing long-lived credentials.

```yaml
# Prerequisites: Create an IAM OIDC provider and role in AWS
# See: https://docs.github.com/en/actions/deployment/security-hardening-your-deployments

permissions:
  id-token: write   # required for OIDC
  contents: read

jobs:
  terraform:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Configure AWS Credentials via OIDC
        uses: aws-actions/configure-aws-credentials@v4
        with:
          role-to-assume: arn:aws:iam::123456789012:role/GitHubActionsRole
          role-session-name: github-actions-terraform
          aws-region: us-east-1

      - uses: hashicorp/setup-terraform@v3
        with:
          terraform_version: "1.7.4"

      - name: Terraform Init & Plan
        working-directory: infrastructure
        run: |
          terraform init -input=false
          terraform plan -input=false
```

The IAM role trust policy for the OIDC provider:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Federated": "arn:aws:iam::123456789012:oidc-provider/token.actions.githubusercontent.com"
      },
      "Action": "sts:AssumeRoleWithWebIdentity",
      "Condition": {
        "StringEquals": {
          "token.actions.githubusercontent.com:aud": "sts.amazonaws.com"
        },
        "StringLike": {
          "token.actions.githubusercontent.com:sub": "repo:myorg/myrepo:*"
        }
      }
    }
  ]
}
```

### Caching Terraform Providers

Speed up CI runs by caching downloaded providers between workflow executions.

```yaml
      - name: Cache Terraform providers
        uses: actions/cache@v4
        with:
          path: |
            ~/.terraform.d/plugin-cache
            infrastructure/.terraform/providers
          key: terraform-providers-${{ runner.os }}-${{ hashFiles('infrastructure/.terraform.lock.hcl') }}
          restore-keys: |
            terraform-providers-${{ runner.os }}-

      - name: Configure provider plugin cache
        run: |
          mkdir -p ~/.terraform.d/plugin-cache
          echo 'plugin_cache_dir = "$HOME/.terraform.d/plugin-cache"' > ~/.terraformrc

      - name: Terraform Init
        run: terraform init -input=false
        working-directory: infrastructure
```

### Matrix Strategy for Multiple Workspaces

Run plan/apply across multiple Terraform workspaces or root modules in parallel.

```yaml
jobs:
  terraform:
    name: ${{ matrix.workspace }}
    runs-on: ubuntu-latest
    strategy:
      fail-fast: false
      max-parallel: 4
      matrix:
        workspace:
          - networking
          - compute
          - database
          - monitoring
    defaults:
      run:
        working-directory: infrastructure/${{ matrix.workspace }}

    steps:
      - uses: actions/checkout@v4
      - uses: hashicorp/setup-terraform@v3
        with:
          terraform_version: "1.7.4"

      - name: Terraform Init
        run: terraform init -input=false

      - name: Terraform Plan
        run: terraform plan -no-color -input=false
```

### Full Production-Ready Workflow

A comprehensive workflow combining all best practices: OIDC auth, caching, security scanning, plan comments, environment gates, and notifications.

```yaml
# .github/workflows/terraform-production.yml
name: Terraform Production Pipeline

on:
  push:
    branches: [main]
    paths: ['infrastructure/**']
  pull_request:
    branches: [main]
    paths: ['infrastructure/**']

permissions:
  id-token: write
  contents: read
  pull-requests: write

env:
  TF_VERSION: "1.7.4"
  TF_WORKING_DIR: "infrastructure"
  AWS_REGION: "us-east-1"

jobs:
  # Stage 1: Validate and scan
  validate:
    name: Validate & Security Scan
    runs-on: ubuntu-latest
    defaults:
      run:
        working-directory: ${{ env.TF_WORKING_DIR }}
    steps:
      - uses: actions/checkout@v4
      - uses: hashicorp/setup-terraform@v3
        with:
          terraform_version: ${{ env.TF_VERSION }}

      - name: Terraform Format Check
        run: terraform fmt -check -recursive

      - name: Terraform Init
        run: terraform init -backend=false

      - name: Terraform Validate
        run: terraform validate

      - name: Run tfsec
        uses: aquasecurity/tfsec-action@v1.0.3
        with:
          working_directory: ${{ env.TF_WORKING_DIR }}
          soft_fail: false

  # Stage 2: Plan
  plan:
    name: Terraform Plan
    needs: validate
    runs-on: ubuntu-latest
    defaults:
      run:
        working-directory: ${{ env.TF_WORKING_DIR }}
    steps:
      - uses: actions/checkout@v4
      - uses: hashicorp/setup-terraform@v3
        with:
          terraform_version: ${{ env.TF_VERSION }}
          terraform_wrapper: true

      - name: Configure AWS Credentials
        uses: aws-actions/configure-aws-credentials@v4
        with:
          role-to-assume: ${{ secrets.AWS_ROLE_ARN }}
          aws-region: ${{ env.AWS_REGION }}

      - name: Terraform Init
        run: terraform init -input=false

      - name: Terraform Plan
        id: plan
        run: terraform plan -no-color -input=false -out=tfplan

      - name: Upload Plan
        uses: actions/upload-artifact@v4
        with:
          name: tfplan
          path: ${{ env.TF_WORKING_DIR }}/tfplan

  # Stage 3: Apply (only on main, after approval)
  apply:
    name: Terraform Apply
    needs: plan
    if: github.ref == 'refs/heads/main' && github.event_name == 'push'
    runs-on: ubuntu-latest
    environment: production
    defaults:
      run:
        working-directory: ${{ env.TF_WORKING_DIR }}
    steps:
      - uses: actions/checkout@v4
      - uses: hashicorp/setup-terraform@v3
        with:
          terraform_version: ${{ env.TF_VERSION }}

      - name: Configure AWS Credentials
        uses: aws-actions/configure-aws-credentials@v4
        with:
          role-to-assume: ${{ secrets.AWS_ROLE_ARN }}
          aws-region: ${{ env.AWS_REGION }}

      - name: Terraform Init
        run: terraform init -input=false

      - name: Download Plan
        uses: actions/download-artifact@v4
        with:
          name: tfplan
          path: ${{ env.TF_WORKING_DIR }}

      - name: Terraform Apply
        run: terraform apply -auto-approve -input=false tfplan
```

---

## Atlantis

### What Atlantis Is and When to Use It

Atlantis is a self-hosted application that listens for Terraform pull request events via webhooks. When a PR is opened or updated, Atlantis runs `terraform plan` and posts the output as a PR comment. Team members review the plan and then comment `atlantis apply` to apply it. Atlantis is ideal when you want PR-based Terraform workflows without paying for Terraform Cloud, or when you need deep customization of the plan/apply lifecycle.

### Installation and Configuration

```bash
# Install Atlantis via Docker
docker run -d \
  --name atlantis \
  -p 4141:4141 \
  -e ATLANTIS_GH_USER="atlantis-bot" \
  -e ATLANTIS_GH_TOKEN="ghp_xxxxxxxxxxxx" \
  -e ATLANTIS_GH_WEBHOOK_SECRET="your-webhook-secret" \
  -e ATLANTIS_REPO_ALLOWLIST="github.com/myorg/*" \
  -e ATLANTIS_ATLANTIS_URL="https://atlantis.example.com" \
  ghcr.io/runatlantis/atlantis:v0.28.0 \
  server

# Or install via Helm for Kubernetes
helm repo add runatlantis https://runatlantis.github.io/helm-charts
helm install atlantis runatlantis/atlantis \
  --set orgAllowlist="github.com/myorg/*" \
  --set github.user="atlantis-bot" \
  --set github.token="ghp_xxxxxxxxxxxx" \
  --set github.secret="your-webhook-secret"
```

### atlantis.yaml Configuration

Place this in the root of your Terraform repository to control how Atlantis discovers and runs projects.

```yaml
# atlantis.yaml (repo root)
version: 3
automerge: false
delete_source_branch_on_merge: true
parallel_plan: true
parallel_apply: false  # apply sequentially for safety

projects:
  - name: networking
    dir: infrastructure/networking
    workspace: default
    terraform_version: v1.7.4
    autoplan:
      when_modified:
        - "*.tf"
        - "*.tfvars"
        - "../modules/vpc/**"  # replan when shared module changes
      enabled: true
    apply_requirements:
      - approved  # PR must be approved before apply
      - mergeable  # branch protection checks must pass

  - name: compute
    dir: infrastructure/compute
    workspace: default
    terraform_version: v1.7.4
    autoplan:
      when_modified:
        - "*.tf"
        - "*.tfvars"
      enabled: true
    apply_requirements:
      - approved
      - mergeable

  - name: database-prod
    dir: infrastructure/database
    workspace: production
    terraform_version: v1.7.4
    autoplan:
      enabled: false  # manual plan only for production database
    apply_requirements:
      - approved
      - mergeable
```

### Server-Side Repo Config

Server-side configuration allows global policies across all repos. Place this on the Atlantis server.

```yaml
# repos.yaml (server-side config)
repos:
  # Default settings for all repos
  - id: /.*/
    allowed_overrides:
      - apply_requirements
      - workflow
      - delete_source_branch_on_merge
    apply_requirements:
      - approved
      - mergeable
    allow_custom_workflows: true

  # Stricter settings for production repos
  - id: github.com/myorg/infra-production
    apply_requirements:
      - approved
      - mergeable
    allowed_overrides: []  # no overrides allowed
    allow_custom_workflows: false

workflows:
  custom:
    plan:
      steps:
        - run: terraform fmt -check -diff
        - init
        - plan:
            extra_args: ["-var-file", "prod.tfvars"]
    apply:
      steps:
        - apply
```

### Workflow Customization

Define custom workflows that add pre/post steps around plan and apply.

```yaml
# atlantis.yaml with custom workflow
version: 3
projects:
  - name: main
    dir: infrastructure
    workflow: security-scan

workflows:
  security-scan:
    plan:
      steps:
        - run: |
            # Run tfsec before planning
            tfsec . --minimum-severity HIGH --soft-fail
        - init
        - plan
        - run: |
            # Run infracost after planning
            infracost breakdown --path=$PLANFILE --format=json > /tmp/infracost.json
            infracost output --path=/tmp/infracost.json --format=table
    apply:
      steps:
        - apply
        - run: |
            # Send Slack notification after apply
            curl -X POST "$SLACK_WEBHOOK" \
              -H 'Content-type: application/json' \
              -d "{\"text\":\"Terraform applied for $PROJECT_NAME in $DIR\"}"
```

### PR-Based Plan and Apply

The Atlantis workflow in pull requests:

```
1. Developer opens PR with Terraform changes
2. Atlantis webhook fires, runs `terraform plan` automatically
3. Plan output posted as PR comment
4. Reviewer reviews code AND plan output
5. Reviewer approves PR
6. Developer comments: atlantis apply -p networking
7. Atlantis runs `terraform apply`, posts result
8. PR auto-merged (if automerge enabled)
```

Commands available in PR comments:

```
atlantis plan                    # Re-run plan for all projects
atlantis plan -p networking      # Plan specific project
atlantis plan -d infrastructure  # Plan specific directory
atlantis apply                   # Apply all planned projects
atlantis apply -p networking     # Apply specific project
atlantis unlock                  # Release locks without applying
```

### Locking and Concurrent Plan Handling

Atlantis locks a project directory when a plan is run. Other PRs modifying the same directory will see a lock conflict.

```
# If another PR tries to plan a locked project:
> atlantis plan
> Error: This project is currently locked by PR #42.
> To unlock, either merge/close PR #42 or comment `atlantis unlock`.

# Force unlock from the locking PR:
> atlantis unlock

# Server-side lock management:
curl -X DELETE https://atlantis.example.com/api/locks/networking
```

### Security Considerations

```bash
# Restrict webhook to Atlantis IP only (nginx example)
location /events {
    allow 10.0.1.50;  # Atlantis server IP
    deny all;
    proxy_pass http://atlantis:4141;
}

# Use GitHub App authentication instead of personal access tokens
# for better audit trails and fine-grained permissions
docker run -d \
  --name atlantis \
  -e ATLANTIS_GH_APP_ID="12345" \
  -e ATLANTIS_GH_APP_KEY_FILE="/etc/atlantis/app-key.pem" \
  -e ATLANTIS_GH_WEBHOOK_SECRET="$WEBHOOK_SECRET" \
  -e ATLANTIS_REPO_ALLOWLIST="github.com/myorg/*" \
  ghcr.io/runatlantis/atlantis:v0.28.0 server
```

---

## Terraform Cloud / Terraform Enterprise

### Workspace Configuration

```hcl
# Configure the Terraform Cloud backend in your root module
terraform {
  cloud {
    organization = "my-org"
    workspaces {
      name = "app-production"
    }
  }
}

# Or use tags for dynamic workspace selection
terraform {
  cloud {
    organization = "my-org"
    workspaces {
      tags = ["app", "production"]
    }
  }
}
```

### VCS-Driven Workflow

Terraform Cloud watches a VCS repository and automatically triggers runs when changes are pushed.

```hcl
# Configure workspace via the TFE provider
resource "tfe_workspace" "app_production" {
  name              = "app-production"
  organization      = "my-org"
  terraform_version = "1.7.4"
  working_directory = "infrastructure"
  queue_all_runs    = false  # only trigger on changes to working directory

  vcs_repo {
    identifier     = "myorg/myrepo"
    branch         = "main"
    oauth_token_id = tfe_oauth_client.github.oauth_token_id
  }

  # Trigger patterns: only run when these files change
  trigger_prefixes = [
    "infrastructure/",
    "modules/",
  ]
}

resource "tfe_variable" "aws_region" {
  key          = "AWS_REGION"
  value        = "us-east-1"
  category     = "env"
  workspace_id = tfe_workspace.app_production.id
}

resource "tfe_variable" "environment" {
  key          = "environment"
  value        = "production"
  category     = "terraform"  # terraform variable, not env var
  workspace_id = tfe_workspace.app_production.id
}
```

### API-Driven Workflow

Trigger runs programmatically via the Terraform Cloud API.

```bash
#!/bin/bash
# trigger-run.sh - trigger a Terraform Cloud run via API

TFC_ORG="my-org"
TFC_WORKSPACE="app-production"
TFC_TOKEN="${TFC_API_TOKEN}"

# Get workspace ID
WORKSPACE_ID=$(curl -s \
  --header "Authorization: Bearer ${TFC_TOKEN}" \
  --header "Content-Type: application/vnd.api+json" \
  "https://app.terraform.io/api/v2/organizations/${TFC_ORG}/workspaces/${TFC_WORKSPACE}" \
  | jq -r '.data.id')

# Upload configuration version
UPLOAD_URL=$(curl -s \
  --header "Authorization: Bearer ${TFC_TOKEN}" \
  --header "Content-Type: application/vnd.api+json" \
  --request POST \
  --data '{
    "data": {
      "type": "configuration-versions",
      "attributes": {
        "auto-queue-runs": true
      }
    }
  }' \
  "https://app.terraform.io/api/v2/workspaces/${WORKSPACE_ID}/configuration-versions" \
  | jq -r '.data.attributes."upload-url"')

# Create tarball and upload
tar -czf config.tar.gz -C infrastructure .
curl -s \
  --header "Content-Type: application/octet-stream" \
  --request PUT \
  --data-binary @config.tar.gz \
  "${UPLOAD_URL}"

echo "Configuration uploaded. Run will start automatically."
```

### CLI-Driven Workflow

Run Terraform commands locally but execute them remotely on Terraform Cloud.

```bash
# Log in to Terraform Cloud
terraform login

# Initialize with cloud backend
terraform init

# Plan remotely (output shown locally)
terraform plan

# Apply remotely (requires confirmation)
terraform apply
```

### Run Triggers

Chain workspaces so that applying one triggers a run in another.

```hcl
# When networking workspace applies, trigger compute workspace
resource "tfe_run_trigger" "compute_after_networking" {
  workspace_id  = tfe_workspace.compute.id
  sourceable_id = tfe_workspace.networking.id
}
```

### Policy Enforcement (Sentinel)

Sentinel policies enforce governance rules on every Terraform run.

```python
# policy: restrict-instance-types.sentinel
# Ensure only approved EC2 instance types are used

import "tfplan/v2" as tfplan

allowed_types = ["t3.micro", "t3.small", "t3.medium", "m5.large", "m5.xlarge"]

ec2_instances = filter tfplan.resource_changes as _, rc {
    rc.type is "aws_instance" and
    (rc.change.actions contains "create" or rc.change.actions contains "update")
}

instance_type_allowed = rule {
    all ec2_instances as _, instance {
        instance.change.after.instance_type in allowed_types
    }
}

main = rule {
    instance_type_allowed
}
```

```python
# policy: require-tags.sentinel
# All resources must have required tags

import "tfplan/v2" as tfplan

required_tags = ["Environment", "Team", "CostCenter"]

all_resources = filter tfplan.resource_changes as _, rc {
    rc.change.actions contains "create" or rc.change.actions contains "update"
}

taggable_resources = filter all_resources as _, rc {
    rc.change.after.tags is not undefined and
    rc.change.after.tags is not null
}

tags_present = rule {
    all taggable_resources as _, resource {
        all required_tags as tag {
            resource.change.after.tags contains tag
        }
    }
}

main = rule {
    tags_present
}
```

### Cost Estimation

Terraform Cloud provides built-in cost estimation for AWS, Azure, and GCP resources. Enable it in organization settings. Cost estimates appear automatically on every run in the UI and API responses. No additional configuration is needed.

### Agent Pools for Private Networking

When Terraform Cloud needs to reach private infrastructure (e.g., on-prem vSphere, private VPC endpoints), use agent pools.

```bash
# Run a Terraform Cloud Agent on an internal network
docker run -d \
  --name tfc-agent \
  -e TFC_AGENT_TOKEN="your-agent-token" \
  -e TFC_AGENT_NAME="internal-agent-01" \
  hashicorp/tfc-agent:latest
```

```hcl
# Assign workspace to an agent pool
resource "tfe_workspace" "private_infra" {
  name         = "private-infra"
  organization = "my-org"
  agent_pool_id   = tfe_agent_pool.internal.id
  execution_mode  = "agent"
}

resource "tfe_agent_pool" "internal" {
  name         = "internal-network"
  organization = "my-org"
}
```

---

## GitLab CI for Terraform

### Basic Pipeline

```yaml
# .gitlab-ci.yml
image:
  name: hashicorp/terraform:1.7.4
  entrypoint: [""]

variables:
  TF_ROOT: "infrastructure"
  TF_STATE_NAME: "default"

cache:
  key: terraform-providers
  paths:
    - ${TF_ROOT}/.terraform/providers/

stages:
  - validate
  - plan
  - apply

validate:
  stage: validate
  script:
    - cd ${TF_ROOT}
    - terraform init -backend=false
    - terraform fmt -check -recursive
    - terraform validate
  rules:
    - if: $CI_MERGE_REQUEST_IID
    - if: $CI_COMMIT_BRANCH == $CI_DEFAULT_BRANCH

plan:
  stage: plan
  script:
    - cd ${TF_ROOT}
    - terraform init -input=false
    - terraform plan -no-color -input=false -out=tfplan
  artifacts:
    paths:
      - ${TF_ROOT}/tfplan
    expire_in: 1 week
  rules:
    - if: $CI_MERGE_REQUEST_IID
    - if: $CI_COMMIT_BRANCH == $CI_DEFAULT_BRANCH

apply:
  stage: apply
  script:
    - cd ${TF_ROOT}
    - terraform init -input=false
    - terraform apply -auto-approve -input=false tfplan
  dependencies:
    - plan
  rules:
    - if: $CI_COMMIT_BRANCH == $CI_DEFAULT_BRANCH
      when: manual  # manual click to apply
  environment:
    name: production
```

### Environment-Specific Jobs

```yaml
# .gitlab-ci.yml - multi-environment
stages:
  - validate
  - plan
  - deploy-dev
  - deploy-staging
  - deploy-production

.plan_template: &plan_template
  stage: plan
  script:
    - cd ${TF_ROOT}
    - terraform init -backend-config="environments/${ENVIRONMENT}/backend.tfvars" -input=false
    - terraform plan -var-file="environments/${ENVIRONMENT}/terraform.tfvars" -out=tfplan-${ENVIRONMENT} -input=false
  artifacts:
    paths:
      - ${TF_ROOT}/tfplan-${ENVIRONMENT}
    expire_in: 1 week

plan-dev:
  <<: *plan_template
  variables:
    ENVIRONMENT: dev

plan-staging:
  <<: *plan_template
  variables:
    ENVIRONMENT: staging

plan-production:
  <<: *plan_template
  variables:
    ENVIRONMENT: production
```

### Terraform State in GitLab (HTTP Backend)

GitLab provides a built-in HTTP backend for Terraform state.

```hcl
# backend.tf
terraform {
  backend "http" {
    # These values are set via environment variables in GitLab CI:
    # TF_HTTP_ADDRESS, TF_HTTP_LOCK_ADDRESS, TF_HTTP_UNLOCK_ADDRESS
  }
}
```

```yaml
# .gitlab-ci.yml - configure HTTP backend variables
variables:
  TF_STATE_NAME: "production"
  TF_HTTP_ADDRESS: "${CI_API_V4_URL}/projects/${CI_PROJECT_ID}/terraform/state/${TF_STATE_NAME}"
  TF_HTTP_LOCK_ADDRESS: "${CI_API_V4_URL}/projects/${CI_PROJECT_ID}/terraform/state/${TF_STATE_NAME}/lock"
  TF_HTTP_UNLOCK_ADDRESS: "${CI_API_V4_URL}/projects/${CI_PROJECT_ID}/terraform/state/${TF_STATE_NAME}/lock"
  TF_HTTP_LOCK_METHOD: "POST"
  TF_HTTP_UNLOCK_METHOD: "DELETE"
  TF_HTTP_USERNAME: "gitlab-ci-token"
  TF_HTTP_PASSWORD: "${CI_JOB_TOKEN}"
```

### Manual Apply with Approval

```yaml
apply-production:
  stage: deploy-production
  script:
    - cd ${TF_ROOT}
    - terraform init -backend-config="environments/production/backend.tfvars" -input=false
    - terraform apply -auto-approve -input=false tfplan-production
  dependencies:
    - plan-production
  rules:
    - if: $CI_COMMIT_BRANCH == $CI_DEFAULT_BRANCH
      when: manual
      allow_failure: false  # pipeline blocks until manually triggered
  environment:
    name: production
    url: https://app.example.com
```

---

## Plan/Apply Workflow Patterns

### Two-Stage (Plan in PR, Apply on Merge)

The most common pattern: plan runs on every PR push, apply runs when the PR is merged to main.

```yaml
# Two-stage workflow summary:
# 1. PR opened/updated -> terraform plan (posted as PR comment)
# 2. PR merged to main -> terraform apply (using saved plan artifact)

# Key consideration: the plan artifact may be stale by the time the PR is
# merged if other changes landed first. Mitigations:
# - Re-plan on merge before apply
# - Use workspace locking (Atlantis handles this automatically)
# - Require linear history (no merge commits)
```

### Three-Stage (Plan, Approve, Apply)

```yaml
# Three-stage adds an explicit approval step between plan and apply.
# GitHub Environments with required reviewers provide this natively.

# Flow:
# 1. PR push -> terraform plan (automatic)
# 2. Reviewer approves in GitHub Environment UI
# 3. terraform apply runs (automatic after approval)
```

### Drift Detection Scheduled Runs

Run `terraform plan` on a schedule to detect configuration drift caused by manual changes.

```yaml
# .github/workflows/drift-detection.yml
name: Terraform Drift Detection

on:
  schedule:
    - cron: '0 8 * * 1-5'  # weekdays at 8 AM UTC
  workflow_dispatch: {}     # allow manual trigger

permissions:
  id-token: write
  contents: read
  issues: write

jobs:
  drift-detect:
    name: Detect Drift
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: hashicorp/setup-terraform@v3
        with:
          terraform_version: "1.7.4"

      - name: Configure AWS Credentials
        uses: aws-actions/configure-aws-credentials@v4
        with:
          role-to-assume: ${{ secrets.AWS_ROLE_ARN }}
          aws-region: us-east-1

      - name: Terraform Init
        working-directory: infrastructure
        run: terraform init -input=false

      - name: Terraform Plan (detect drift)
        id: plan
        working-directory: infrastructure
        run: terraform plan -detailed-exitcode -input=false -no-color 2>&1 | tee plan-output.txt
        continue-on-error: true

      # Exit code 2 = changes detected (drift)
      - name: Create Issue on Drift
        if: steps.plan.outputs.exitcode == '2'
        uses: actions/github-script@v7
        with:
          script: |
            const fs = require('fs');
            const planOutput = fs.readFileSync('infrastructure/plan-output.txt', 'utf8');
            await github.rest.issues.create({
              owner: context.repo.owner,
              repo: context.repo.repo,
              title: `Terraform Drift Detected - ${new Date().toISOString().split('T')[0]}`,
              body: `## Drift Detected\n\nTerraform plan shows differences between state and real infrastructure.\n\n<details><summary>Plan Output</summary>\n\n\`\`\`\n${planOutput.substring(0, 60000)}\n\`\`\`\n\n</details>`,
              labels: ['infrastructure', 'drift'],
            });
```

### Automated Rollback Strategies

```bash
#!/bin/bash
# rollback.sh - revert to a previous Terraform state version
# Use with caution: this replaces current state with a previous snapshot

set -euo pipefail

BUCKET="my-terraform-state"
KEY="infrastructure/terraform.tfstate"
REGION="us-east-1"

# List recent state versions
echo "Recent state versions:"
aws s3api list-object-versions \
  --bucket "$BUCKET" \
  --prefix "$KEY" \
  --max-items 5 \
  --query 'Versions[].{VersionId:VersionId,LastModified:LastModified,Size:Size}' \
  --output table

# Roll back to a specific version
read -p "Enter VersionId to restore: " VERSION_ID
aws s3api get-object \
  --bucket "$BUCKET" \
  --key "$KEY" \
  --version-id "$VERSION_ID" \
  restored-state.tfstate

echo "Downloaded state version ${VERSION_ID}."
echo "Review it, then push with: terraform state push restored-state.tfstate"
```

### Plan Artifacts and Security

```yaml
      # Plan files contain sensitive data (resource attributes, outputs).
      # Always treat plan artifacts as secrets.

      - name: Upload Plan (encrypted)
        uses: actions/upload-artifact@v4
        with:
          name: tfplan
          path: infrastructure/tfplan
          retention-days: 1  # short retention for security

      # Never log plan files to stdout in CI without -no-color
      # Never store plan files in long-lived artifact storage
      # Consider encrypting plan files at rest:
      # terraform plan -out=tfplan && gpg --symmetric --batch --passphrase "$PLAN_KEY" tfplan
```

---

## Cost Estimation

### Infracost Integration

Infracost provides cost estimates for Terraform changes before they are applied.

```bash
# Install Infracost locally
brew install infracost

# Authenticate
infracost auth login

# Generate cost breakdown from HCL
infracost breakdown --path=infrastructure/

# Generate cost diff between current and planned changes
infracost diff --path=infrastructure/
```

### GitHub Actions with Infracost

```yaml
# .github/workflows/infracost.yml
name: Infracost

on:
  pull_request:
    paths: ['infrastructure/**']

permissions:
  contents: read
  pull-requests: write

jobs:
  infracost:
    name: Cost Estimation
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup Infracost
        uses: infracost/actions/setup@v3
        with:
          api-key: ${{ secrets.INFRACOST_API_KEY }}

      # Generate cost estimate for the base branch
      - name: Checkout base branch
        uses: actions/checkout@v4
        with:
          ref: ${{ github.event.pull_request.base.ref }}
          path: base

      - name: Generate Infracost baseline
        run: |
          infracost breakdown \
            --path=base/infrastructure \
            --format=json \
            --out-file=/tmp/infracost-base.json

      # Generate cost estimate for the PR branch
      - name: Checkout PR branch
        uses: actions/checkout@v4
        with:
          path: pr

      - name: Generate Infracost diff
        run: |
          infracost diff \
            --path=pr/infrastructure \
            --compare-to=/tmp/infracost-base.json \
            --format=json \
            --out-file=/tmp/infracost-diff.json

      - name: Post Infracost comment
        run: |
          infracost comment github \
            --path=/tmp/infracost-diff.json \
            --repo=${{ github.repository }} \
            --pull-request=${{ github.event.pull_request.number }} \
            --github-token=${{ secrets.GITHUB_TOKEN }} \
            --behavior=update
```

### Cost Policies and Thresholds

```yaml
# infracost.yml - policy file
version: "0.1"
policies:
  - name: "Monthly cost limit"
    description: "Block PRs that increase monthly cost by more than $500"
    resource_type: "total"
    condition:
      monthly_cost_change: "> 500"
    action: "deny"

  - name: "Warn on large instances"
    description: "Warn when using expensive instance types"
    resource_type: "aws_instance"
    condition:
      monthly_cost: "> 200"
    action: "warn"
```

```bash
# Enforce cost thresholds in CI
infracost diff \
  --path=infrastructure/ \
  --compare-to=infracost-base.json \
  --format=json \
  --out-file=infracost-diff.json

# Check if monthly cost increase exceeds threshold
COST_CHANGE=$(jq '.diffTotalMonthlyCost | tonumber' infracost-diff.json)
THRESHOLD=500

if (( $(echo "$COST_CHANGE > $THRESHOLD" | bc -l) )); then
  echo "::error::Monthly cost increase of \$${COST_CHANGE} exceeds threshold of \$${THRESHOLD}"
  exit 1
fi
```

---

## Secrets in CI/CD

### OIDC for AWS (No Long-Lived Keys)

```hcl
# Create the OIDC provider in AWS (one-time setup)
resource "aws_iam_openid_connect_provider" "github" {
  url             = "https://token.actions.githubusercontent.com"
  client_id_list  = ["sts.amazonaws.com"]
  thumbprint_list = ["6938fd4d98bab03faadb97b34396831e3780aea1"]
}

# IAM role for GitHub Actions
resource "aws_iam_role" "github_actions" {
  name = "github-actions-terraform"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Federated = aws_iam_openid_connect_provider.github.arn
        }
        Action = "sts:AssumeRoleWithWebIdentity"
        Condition = {
          StringEquals = {
            "token.actions.githubusercontent.com:aud" = "sts.amazonaws.com"
          }
          StringLike = {
            "token.actions.githubusercontent.com:sub" = "repo:myorg/myrepo:ref:refs/heads/main"
          }
        }
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "terraform_permissions" {
  role       = aws_iam_role.github_actions.name
  policy_arn = "arn:aws:iam::aws:policy/AdministratorAccess"  # scope down in production
}
```

### GitHub Secrets Management

```yaml
# Reference secrets in workflows (set via repo Settings > Secrets)
env:
  AWS_ROLE_ARN: ${{ secrets.AWS_ROLE_ARN }}
  TF_VAR_db_password: ${{ secrets.DB_PASSWORD }}
  INFRACOST_API_KEY: ${{ secrets.INFRACOST_API_KEY }}

# Use environment-scoped secrets for different environments
jobs:
  deploy-production:
    environment: production
    env:
      # This pulls from the 'production' environment secrets
      TF_VAR_db_password: ${{ secrets.DB_PASSWORD }}
```

### Vault Integration in CI

```yaml
      - name: Retrieve secrets from Vault
        uses: hashicorp/vault-action@v3
        with:
          url: https://vault.example.com
          method: jwt
          role: github-actions-terraform
          jwtGithubAudience: https://vault.example.com
          secrets: |
            secret/data/terraform/aws access_key | AWS_ACCESS_KEY_ID ;
            secret/data/terraform/aws secret_key | AWS_SECRET_ACCESS_KEY ;
            secret/data/terraform/database password | TF_VAR_db_password

      - name: Terraform Apply
        run: terraform apply -auto-approve
        env:
          AWS_ACCESS_KEY_ID: ${{ env.AWS_ACCESS_KEY_ID }}
          AWS_SECRET_ACCESS_KEY: ${{ env.AWS_SECRET_ACCESS_KEY }}
```

### Environment Variable Patterns

```bash
# Pattern 1: TF_VAR_ prefix (automatically picked up by Terraform)
export TF_VAR_environment="production"
export TF_VAR_instance_type="m5.large"
export TF_VAR_db_password="${DB_PASSWORD}"  # from secrets manager

# Pattern 2: .auto.tfvars for non-sensitive values
# CI generates this file from environment config
cat > infrastructure/ci.auto.tfvars <<EOF
environment    = "${ENVIRONMENT}"
commit_sha     = "${GITHUB_SHA}"
deploy_version = "${GITHUB_RUN_NUMBER}"
EOF

# Pattern 3: Backend config from environment
terraform init \
  -backend-config="bucket=${TF_STATE_BUCKET}" \
  -backend-config="key=${TF_STATE_KEY}" \
  -backend-config="region=${AWS_REGION}"
```

---

## Testing in CI

### terraform validate

```bash
# Validate checks syntax and internal consistency
# Does NOT access remote state or providers (fast)
terraform init -backend=false
terraform validate

# In CI:
# - Run on every PR
# - No credentials needed
# - Catches syntax errors, type mismatches, missing required arguments
```

### terraform fmt -check

```bash
# Check formatting without modifying files
terraform fmt -check -recursive -diff

# -check:     exit 1 if files need formatting
# -recursive: check all subdirectories
# -diff:      show the formatting differences
```

### tflint

```bash
# Install tflint
brew install tflint

# Initialize with ruleset
tflint --init

# Run linting
tflint --recursive --format=compact
```

```hcl
# .tflint.hcl - configuration file
plugin "aws" {
  enabled = true
  version = "0.30.0"
  source  = "github.com/terraform-linters/tflint-ruleset-aws"
}

plugin "terraform" {
  enabled = true
  preset  = "recommended"
}

rule "terraform_naming_convention" {
  enabled = true
  format  = "snake_case"
}

rule "terraform_documented_outputs" {
  enabled = true
}

rule "terraform_documented_variables" {
  enabled = true
}
```

### terraform test (TF 1.6+)

Native Terraform testing framework introduced in Terraform 1.6.

```hcl
# tests/vpc.tftest.hcl
# Test that the VPC module creates expected resources

variables {
  environment = "test"
  vpc_cidr    = "10.99.0.0/16"
}

run "create_vpc" {
  command = plan  # or apply for integration tests

  assert {
    condition     = aws_vpc.main.cidr_block == "10.99.0.0/16"
    error_message = "VPC CIDR block does not match expected value"
  }

  assert {
    condition     = aws_vpc.main.tags["Environment"] == "test"
    error_message = "VPC Environment tag is incorrect"
  }

  assert {
    condition     = length(aws_subnet.private) == 3
    error_message = "Expected 3 private subnets"
  }
}

run "verify_nacl" {
  command = plan

  assert {
    condition     = aws_network_acl.main.ingress[0].action == "allow"
    error_message = "Default NACL should allow ingress"
  }
}
```

```bash
# Run tests
terraform test

# Run specific test file
terraform test -filter=tests/vpc.tftest.hcl

# Run with verbose output
terraform test -verbose
```

### Terratest in CI

```yaml
# .github/workflows/terratest.yml
name: Terratest

on:
  pull_request:
    paths: ['modules/**']

jobs:
  test:
    name: Integration Tests
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - name: Setup Terraform
        uses: hashicorp/setup-terraform@v3
        with:
          terraform_version: "1.7.4"
          terraform_wrapper: false  # terratest needs raw terraform output

      - name: Configure AWS Credentials
        uses: aws-actions/configure-aws-credentials@v4
        with:
          role-to-assume: ${{ secrets.AWS_TEST_ROLE_ARN }}
          aws-region: us-east-1

      - name: Run Terratest
        working-directory: test
        run: |
          go test -v -timeout 30m -run TestVpcModule ./...
        env:
          TF_VAR_environment: "ci-test-${{ github.run_id }}"
```

### Pre-commit Hooks

```yaml
# .pre-commit-config.yaml
repos:
  - repo: https://github.com/antonbabenko/pre-commit-tf-docs
    rev: v0.3.0
    hooks:
      - id: terraform-docs-go
        args: ['--args=--output-file=README.md']

  - repo: https://github.com/terraform-linters/tflint
    rev: v0.50.3
    hooks:
      - id: tflint
        args: ['--recursive']

  - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v4.6.0
    hooks:
      - id: trailing-whitespace
      - id: end-of-file-fixer
      - id: check-merge-conflict

  - repo: https://github.com/gruntwork-io/pre-commit
    rev: v0.1.23
    hooks:
      - id: terraform-fmt
      - id: terraform-validate
      - id: tflint

  - repo: https://github.com/aquasecurity/tfsec
    rev: v1.28.6
    hooks:
      - id: tfsec
```

```bash
# Install and activate pre-commit hooks
pip install pre-commit
pre-commit install

# Run against all files (useful in CI)
pre-commit run --all-files

# CI integration:
# .github/workflows/pre-commit.yml
# - uses: pre-commit/action@v3.0.1
```

---

## Monorepo Strategies

### Detecting Changed Directories

```yaml
# .github/workflows/detect-changes.yml
name: Detect Terraform Changes

on:
  pull_request:
    paths: ['infrastructure/**']

jobs:
  detect:
    name: Detect Changed Modules
    runs-on: ubuntu-latest
    outputs:
      changed_dirs: ${{ steps.changes.outputs.dirs }}
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0  # full history for accurate diff

      - name: Find changed Terraform directories
        id: changes
        run: |
          # Get list of changed .tf files
          CHANGED_FILES=$(git diff --name-only origin/${{ github.base_ref }}...HEAD -- '*.tf' '*.tfvars')

          # Extract unique directories
          DIRS=$(echo "$CHANGED_FILES" | xargs -I{} dirname {} | sort -u | jq -R -s -c 'split("\n") | map(select(. != ""))')

          echo "Changed directories: $DIRS"
          echo "dirs=$DIRS" >> "$GITHUB_OUTPUT"
```

### Selective Plan/Apply

```yaml
  plan:
    name: Plan Changed Modules
    needs: detect
    if: needs.detect.outputs.changed_dirs != '[]'
    runs-on: ubuntu-latest
    strategy:
      fail-fast: false
      matrix:
        dir: ${{ fromJson(needs.detect.outputs.changed_dirs) }}
    steps:
      - uses: actions/checkout@v4
      - uses: hashicorp/setup-terraform@v3
        with:
          terraform_version: "1.7.4"

      - name: Terraform Init
        run: terraform init -input=false
        working-directory: ${{ matrix.dir }}

      - name: Terraform Plan
        run: terraform plan -no-color -input=false
        working-directory: ${{ matrix.dir }}
```

### Dependency-Aware Ordering

```bash
#!/bin/bash
# deploy-in-order.sh - apply Terraform modules respecting dependency order
# Uses a simple dependency file to determine ordering

set -euo pipefail

# Define dependency order (earlier items must be applied first)
declare -a ORDERED_MODULES=(
  "infrastructure/networking"
  "infrastructure/security"
  "infrastructure/database"
  "infrastructure/compute"
  "infrastructure/monitoring"
)

CHANGED_DIRS="$1"  # comma-separated list of changed directories

for module in "${ORDERED_MODULES[@]}"; do
  if echo "$CHANGED_DIRS" | grep -q "$module"; then
    echo "=== Applying: $module ==="
    cd "$module"
    terraform init -input=false
    terraform apply -auto-approve -input=false
    cd -
  else
    echo "--- Skipping: $module (no changes) ---"
  fi
done
```

### Parallel Execution

```yaml
# Use matrix strategy for independent modules
jobs:
  plan:
    strategy:
      fail-fast: false
      max-parallel: 5  # limit concurrent runs to avoid API rate limits
      matrix:
        module:
          - infrastructure/networking
          - infrastructure/compute
          - infrastructure/database
    steps:
      - uses: actions/checkout@v4
      - uses: hashicorp/setup-terraform@v3

      - name: Plan
        working-directory: ${{ matrix.module }}
        run: |
          terraform init -input=false
          terraform plan -input=false -no-color
```

---

## Notifications and Reporting

### Slack Notifications on Apply

```yaml
      - name: Notify Slack on Success
        if: success()
        uses: slackapi/slack-github-action@v1.26.0
        with:
          channel-id: 'C0123456789'  # infrastructure-deploys channel
          payload: |
            {
              "blocks": [
                {
                  "type": "header",
                  "text": {
                    "type": "plain_text",
                    "text": "Terraform Apply Succeeded"
                  }
                },
                {
                  "type": "section",
                  "fields": [
                    {
                      "type": "mrkdwn",
                      "text": "*Repository:*\n${{ github.repository }}"
                    },
                    {
                      "type": "mrkdwn",
                      "text": "*Branch:*\n${{ github.ref_name }}"
                    },
                    {
                      "type": "mrkdwn",
                      "text": "*Actor:*\n${{ github.actor }}"
                    },
                    {
                      "type": "mrkdwn",
                      "text": "*Commit:*\n<${{ github.server_url }}/${{ github.repository }}/commit/${{ github.sha }}|${{ github.sha }}>"
                    }
                  ]
                }
              ]
            }
        env:
          SLACK_BOT_TOKEN: ${{ secrets.SLACK_BOT_TOKEN }}

      - name: Notify Slack on Failure
        if: failure()
        uses: slackapi/slack-github-action@v1.26.0
        with:
          channel-id: 'C0123456789'
          payload: |
            {
              "blocks": [
                {
                  "type": "header",
                  "text": {
                    "type": "plain_text",
                    "text": "Terraform Apply FAILED"
                  }
                },
                {
                  "type": "section",
                  "text": {
                    "type": "mrkdwn",
                    "text": "Apply failed for *${{ github.repository }}* on branch `${{ github.ref_name }}`.\n<${{ github.server_url }}/${{ github.repository }}/actions/runs/${{ github.run_id }}|View Run>"
                  }
                }
              ]
            }
        env:
          SLACK_BOT_TOKEN: ${{ secrets.SLACK_BOT_TOKEN }}
```

### PR Status Checks

```yaml
# Configure branch protection rules to require these status checks:
# - Terraform Format & Validate
# - Terraform Plan
# - Security Scan (tfsec)
# - Cost Estimation (Infracost)

# Each job in your workflow becomes a status check. Use consistent job names
# so branch protection rules remain stable.

jobs:
  terraform-fmt:
    name: "Terraform Format"  # this is the status check name
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: hashicorp/setup-terraform@v3
      - run: terraform fmt -check -recursive

  terraform-plan:
    name: "Terraform Plan"  # required status check
    needs: terraform-fmt
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: hashicorp/setup-terraform@v3
      - run: |
          terraform init -input=false
          terraform plan -input=false -no-color
        working-directory: infrastructure
```

### Apply Logs and Audit Trail

```yaml
      - name: Save Apply Output
        id: apply
        working-directory: infrastructure
        run: |
          terraform apply -auto-approve -input=false -no-color 2>&1 | tee apply-output.txt
          echo "apply_output<<EOF" >> "$GITHUB_OUTPUT"
          cat apply-output.txt >> "$GITHUB_OUTPUT"
          echo "EOF" >> "$GITHUB_OUTPUT"

      - name: Upload Apply Log
        uses: actions/upload-artifact@v4
        with:
          name: apply-log-${{ github.run_number }}
          path: infrastructure/apply-output.txt
          retention-days: 90  # keep for audit purposes

      # Optional: push apply metadata to a central audit log
      - name: Record Audit Event
        if: always()
        run: |
          curl -X POST "${{ secrets.AUDIT_LOG_URL }}" \
            -H "Authorization: Bearer ${{ secrets.AUDIT_TOKEN }}" \
            -H "Content-Type: application/json" \
            -d '{
              "event": "terraform_apply",
              "status": "${{ job.status }}",
              "repository": "${{ github.repository }}",
              "actor": "${{ github.actor }}",
              "commit": "${{ github.sha }}",
              "run_id": "${{ github.run_id }}",
              "timestamp": "'$(date -u +%Y-%m-%dT%H:%M:%SZ)'"
            }'
```

### Dashboard Integration

```bash
#!/bin/bash
# post-apply-metrics.sh - send apply metrics to monitoring/dashboard
# Called as a post-apply step in CI

set -euo pipefail

APPLY_LOG="$1"

# Extract resource counts from apply output
ADDED=$(grep -oP 'added' "$APPLY_LOG" | head -1 | grep -oP '\d+' || echo "0")
CHANGED=$(grep -oP 'changed' "$APPLY_LOG" | head -1 | grep -oP '\d+' || echo "0")
DESTROYED=$(grep -oP 'destroyed' "$APPLY_LOG" | head -1 | grep -oP '\d+' || echo "0")

# Send to Datadog (example)
curl -X POST "https://api.datadoghq.com/api/v1/series" \
  -H "Content-Type: application/json" \
  -H "DD-API-KEY: ${DATADOG_API_KEY}" \
  -d "{
    \"series\": [
      {
        \"metric\": \"terraform.resources.added\",
        \"points\": [[$(date +%s), ${ADDED}]],
        \"type\": \"count\",
        \"tags\": [\"env:${ENVIRONMENT}\", \"repo:${GITHUB_REPOSITORY}\"]
      },
      {
        \"metric\": \"terraform.resources.changed\",
        \"points\": [[$(date +%s), ${CHANGED}]],
        \"type\": \"count\",
        \"tags\": [\"env:${ENVIRONMENT}\", \"repo:${GITHUB_REPOSITORY}\"]
      },
      {
        \"metric\": \"terraform.resources.destroyed\",
        \"points\": [[$(date +%s), ${DESTROYED}]],
        \"type\": \"count\",
        \"tags\": [\"env:${ENVIRONMENT}\", \"repo:${GITHUB_REPOSITORY}\"]
      }
    ]
  }"

echo "Metrics posted: added=${ADDED}, changed=${CHANGED}, destroyed=${DESTROYED}"
```
