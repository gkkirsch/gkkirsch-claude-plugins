# GitHub Actions Reference

Comprehensive reference for GitHub Actions — the most widely used CI/CD platform. Covers workflow syntax, reusable patterns, caching, secrets, security, and advanced features.

## Workflow File Basics

Workflows live in `.github/workflows/` as YAML files. Each workflow has triggers, jobs, and steps.

```yaml
name: CI                           # Workflow name (shown in UI)

on:                                # Triggers
  push:
    branches: [main]
  pull_request:
    branches: [main]

permissions:                       # Workflow-level permissions
  contents: read

jobs:                              # One or more jobs
  build:                           # Job ID
    runs-on: ubuntu-latest         # Runner
    steps:                         # Sequential steps
      - uses: actions/checkout@v4  # Use an action
      - run: echo "Hello"         # Run a shell command
```

---

## Trigger Events (on:)

### Push and Pull Request

```yaml
on:
  push:
    branches:
      - main
      - 'release/**'           # Wildcard matching
    branches-ignore:
      - 'dependabot/**'
    tags:
      - 'v*'                  # Tag pattern
    paths:
      - 'src/**'
      - '*.json'
    paths-ignore:
      - '**.md'
      - 'docs/**'

  pull_request:
    types: [opened, synchronize, reopened, ready_for_review]
    branches: [main]
    paths:
      - 'src/**'

  pull_request_target:          # Runs in context of BASE branch (for forks)
    types: [opened, synchronize]
```

### Schedule (Cron)

```yaml
on:
  schedule:
    - cron: '0 6 * * 1-5'    # Weekdays at 6 AM UTC
    - cron: '0 0 * * 0'      # Sundays at midnight UTC
    # Note: scheduled runs may be delayed during high load
    # Note: scheduled workflows only run on the default branch
```

### Manual Trigger (workflow_dispatch)

```yaml
on:
  workflow_dispatch:
    inputs:
      environment:
        description: 'Target environment'
        required: true
        type: choice
        options:
          - staging
          - production
      debug:
        description: 'Enable debug logging'
        type: boolean
        default: false
      version:
        description: 'Version to deploy'
        type: string
        required: false

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - run: echo "Deploying to ${{ inputs.environment }}"
      - run: echo "Debug: ${{ inputs.debug }}"
```

### Workflow Run (Chain Workflows)

```yaml
on:
  workflow_run:
    workflows: [CI, Build]
    types: [completed]
    branches: [main]

jobs:
  deploy:
    if: ${{ github.event.workflow_run.conclusion == 'success' }}
    runs-on: ubuntu-latest
    steps:
      - run: echo "CI passed, deploying..."
```

### Release

```yaml
on:
  release:
    types: [published, created, prereleased]

jobs:
  publish:
    runs-on: ubuntu-latest
    steps:
      - run: echo "Publishing ${{ github.event.release.tag_name }}"
```

### Other Useful Triggers

```yaml
on:
  issues:
    types: [opened, labeled]
  issue_comment:
    types: [created]
  deployment_status:
  discussion:
    types: [created]
  repository_dispatch:       # External webhook trigger
    types: [deploy]
```

---

## Jobs

### Job Dependencies

```yaml
jobs:
  lint:
    runs-on: ubuntu-latest
    steps: [...]

  test:
    runs-on: ubuntu-latest
    steps: [...]

  build:
    needs: [lint, test]       # Wait for both lint and test
    runs-on: ubuntu-latest
    steps: [...]

  deploy:
    needs: build              # Wait for build
    if: github.ref == 'refs/heads/main'
    runs-on: ubuntu-latest
    steps: [...]
```

### Job Outputs

```yaml
jobs:
  setup:
    runs-on: ubuntu-latest
    outputs:
      version: ${{ steps.ver.outputs.version }}
      matrix: ${{ steps.mat.outputs.matrix }}
    steps:
      - id: ver
        run: echo "version=1.2.3" >> "$GITHUB_OUTPUT"
      - id: mat
        run: |
          echo 'matrix={"node":[18,20,22]}' >> "$GITHUB_OUTPUT"

  build:
    needs: setup
    strategy:
      matrix: ${{ fromJson(needs.setup.outputs.matrix) }}
    runs-on: ubuntu-latest
    steps:
      - run: echo "Building v${{ needs.setup.outputs.version }} on Node ${{ matrix.node }}"
```

### Matrix Strategy

```yaml
jobs:
  test:
    strategy:
      fail-fast: false         # Don't cancel other jobs if one fails
      max-parallel: 3          # Limit concurrent jobs
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
        node: [18, 20, 22]
        exclude:
          - os: windows-latest
            node: 18
        include:
          - os: ubuntu-latest
            node: 22
            upload-coverage: true
    runs-on: ${{ matrix.os }}
    steps:
      - uses: actions/setup-node@v4
        with:
          node-version: ${{ matrix.node }}
      - run: npm test
      - if: matrix.upload-coverage
        run: upload-coverage
```

### Concurrency

```yaml
# Workflow-level concurrency
concurrency:
  group: ${{ github.workflow }}-${{ github.ref }}
  cancel-in-progress: true    # Cancel previous runs

# Job-level concurrency
jobs:
  deploy:
    concurrency:
      group: deploy-${{ github.ref }}
      cancel-in-progress: false  # Queue deploys, don't cancel
```

### Environments

```yaml
jobs:
  deploy-staging:
    runs-on: ubuntu-latest
    environment:
      name: staging
      url: https://staging.example.com
    steps:
      - run: deploy --staging

  deploy-production:
    needs: deploy-staging
    runs-on: ubuntu-latest
    environment:
      name: production           # Requires approval in GitHub settings
      url: https://example.com
    steps:
      - run: deploy --production
```

### Timeouts

```yaml
jobs:
  test:
    runs-on: ubuntu-latest
    timeout-minutes: 30          # Job timeout
    steps:
      - run: npm test
        timeout-minutes: 10      # Step timeout
```

### Conditional Execution

```yaml
jobs:
  deploy:
    if: github.ref == 'refs/heads/main' && github.event_name == 'push'
    runs-on: ubuntu-latest
    steps:
      - run: deploy

  # Skip draft PRs
  test:
    if: github.event.pull_request.draft == false
    runs-on: ubuntu-latest
    steps:
      - run: npm test

  # Run on specific labels
  e2e:
    if: contains(github.event.pull_request.labels.*.name, 'run-e2e')
    runs-on: ubuntu-latest
    steps:
      - run: npm run test:e2e
```

---

## Steps

### Common Step Patterns

```yaml
steps:
  # Checkout code
  - uses: actions/checkout@v4
    with:
      fetch-depth: 0            # Full history (needed for git describe)
      submodules: 'recursive'   # Checkout submodules

  # Setup runtime
  - uses: actions/setup-node@v4
    with:
      node-version-file: '.node-version'  # Read from file
      cache: 'npm'

  # Run command
  - name: Install dependencies
    run: npm ci

  # Multi-line command
  - name: Build and test
    run: |
      npm run build
      npm test

  # Set environment variables
  - run: echo "MY_VAR=value" >> "$GITHUB_ENV"

  # Set step output
  - id: my-step
    run: echo "result=hello" >> "$GITHUB_OUTPUT"

  # Use step output
  - run: echo "${{ steps.my-step.outputs.result }}"

  # Conditional step
  - if: github.event_name == 'push'
    run: deploy

  # Continue on error
  - run: npm run lint
    continue-on-error: true

  # Working directory
  - run: npm test
    working-directory: packages/api
```

### Shell Selection

```yaml
steps:
  - run: echo "Bash"
    shell: bash

  - run: Write-Output "PowerShell"
    shell: pwsh

  - run: print("Python")
    shell: python

  - run: echo "Custom shell"
    shell: bash --noprofile --norc -eo pipefail {0}
```

---

## Actions

### Using Actions

```yaml
steps:
  # Public action with version tag
  - uses: actions/checkout@v4

  # Public action pinned to SHA (most secure)
  - uses: actions/checkout@b4ffde65f46336ab88eb53be808477a3936bae11

  # Action from a different repository
  - uses: owner/repo@v1

  # Action from a subdirectory
  - uses: owner/repo/path/to/action@v1

  # Local action (in same repository)
  - uses: ./.github/actions/my-action

  # Docker action
  - uses: docker://alpine:3.19
    with:
      args: echo "Hello from Docker"
```

### Creating a Composite Action

```yaml
# .github/actions/setup-project/action.yml
name: 'Setup Project'
description: 'Install Node.js and project dependencies'

inputs:
  node-version:
    description: 'Node.js version'
    required: false
    default: '20'
  install-playwright:
    description: 'Install Playwright browsers'
    required: false
    default: 'false'

outputs:
  cache-hit:
    description: 'Whether npm cache was hit'
    value: ${{ steps.cache.outputs.cache-hit }}

runs:
  using: 'composite'
  steps:
    - uses: actions/setup-node@v4
      id: cache
      with:
        node-version: ${{ inputs.node-version }}
        cache: 'npm'

    - run: npm ci
      shell: bash

    - if: inputs.install-playwright == 'true'
      run: npx playwright install --with-deps chromium
      shell: bash
```

### Creating a JavaScript Action

```yaml
# action.yml
name: 'PR Size Label'
description: 'Label PRs by size'
inputs:
  github-token:
    required: true
runs:
  using: 'node20'
  main: 'dist/index.js'
```

```javascript
// index.js
const core = require('@actions/core');
const github = require('@actions/github');

async function run() {
  const token = core.getInput('github-token');
  const octokit = github.getOctokit(token);
  const { context } = github;

  const { data: files } = await octokit.rest.pulls.listFiles({
    owner: context.repo.owner,
    repo: context.repo.repo,
    pull_number: context.payload.pull_request.number,
  });

  const changes = files.reduce((sum, f) => sum + f.changes, 0);
  let label = 'size/S';
  if (changes > 500) label = 'size/XL';
  else if (changes > 200) label = 'size/L';
  else if (changes > 50) label = 'size/M';

  await octokit.rest.issues.addLabels({
    owner: context.repo.owner,
    repo: context.repo.repo,
    issue_number: context.payload.pull_request.number,
    labels: [label],
  });

  core.setOutput('label', label);
}

run().catch((err) => core.setFailed(err.message));
```

---

## Reusable Workflows

### Defining a Reusable Workflow

```yaml
# .github/workflows/reusable-deploy.yml
name: Deploy

on:
  workflow_call:
    inputs:
      environment:
        required: true
        type: string
      image-tag:
        required: true
        type: string
      replicas:
        required: false
        type: number
        default: 3
    secrets:
      KUBE_CONFIG:
        required: true
      SLACK_WEBHOOK:
        required: false
    outputs:
      deploy-url:
        description: 'Deployment URL'
        value: ${{ jobs.deploy.outputs.url }}

jobs:
  deploy:
    runs-on: ubuntu-latest
    outputs:
      url: ${{ steps.deploy.outputs.url }}
    environment:
      name: ${{ inputs.environment }}
      url: ${{ steps.deploy.outputs.url }}
    steps:
      - uses: actions/checkout@v4
      - id: deploy
        run: |
          echo "Deploying ${{ inputs.image-tag }} to ${{ inputs.environment }}"
          echo "url=https://${{ inputs.environment }}.example.com" >> "$GITHUB_OUTPUT"
```

### Calling a Reusable Workflow

```yaml
# .github/workflows/ci-cd.yml
name: CI/CD

on:
  push:
    branches: [main]

jobs:
  test:
    uses: ./.github/workflows/reusable-test.yml
    with:
      node-version: '20'

  deploy-staging:
    needs: test
    uses: ./.github/workflows/reusable-deploy.yml
    with:
      environment: staging
      image-tag: ${{ github.sha }}
    secrets: inherit                 # Pass all secrets

  deploy-production:
    needs: deploy-staging
    uses: ./.github/workflows/reusable-deploy.yml
    with:
      environment: production
      image-tag: ${{ github.sha }}
      replicas: 5
    secrets:
      KUBE_CONFIG: ${{ secrets.PROD_KUBE_CONFIG }}
```

### Cross-Repository Reusable Workflows

```yaml
jobs:
  deploy:
    uses: myorg/shared-workflows/.github/workflows/deploy.yml@v2
    with:
      environment: production
    secrets: inherit
```

---

## Caching

### Built-in Setup Action Caching

```yaml
# Node.js
- uses: actions/setup-node@v4
  with:
    node-version: 20
    cache: 'npm'        # Also: 'yarn', 'pnpm'

# Python
- uses: actions/setup-python@v5
  with:
    python-version: '3.12'
    cache: 'pip'        # Also: 'pipenv', 'poetry'

# Go
- uses: actions/setup-go@v5
  with:
    go-version: '1.22'
    cache: true

# Rust
- uses: dtolnay/rust-toolchain@stable
- uses: Swatinem/rust-cache@v2
```

### Manual Cache with actions/cache

```yaml
- uses: actions/cache@v4
  id: cache
  with:
    path: |
      ~/.npm
      node_modules
      .next/cache
    key: ${{ runner.os }}-node-${{ hashFiles('**/package-lock.json') }}
    restore-keys: |
      ${{ runner.os }}-node-

- if: steps.cache.outputs.cache-hit != 'true'
  run: npm ci
```

### Cache Strategies

```yaml
# Exact match only (strictest)
key: deps-${{ hashFiles('package-lock.json') }}

# Prefix matching (more forgiving)
key: deps-${{ hashFiles('package-lock.json') }}
restore-keys: |
  deps-

# OS-specific cache
key: ${{ runner.os }}-deps-${{ hashFiles('package-lock.json') }}
restore-keys: |
  ${{ runner.os }}-deps-

# Branch-specific with fallback
key: ${{ runner.os }}-${{ github.ref }}-deps-${{ hashFiles('package-lock.json') }}
restore-keys: |
  ${{ runner.os }}-${{ github.ref }}-deps-
  ${{ runner.os }}-refs/heads/main-deps-
  ${{ runner.os }}-
```

### Docker Layer Caching

```yaml
- uses: docker/build-push-action@v6
  with:
    context: .
    push: true
    tags: myapp:latest
    cache-from: type=gha
    cache-to: type=gha,mode=max

# Or using registry cache
    cache-from: type=registry,ref=myapp:buildcache
    cache-to: type=registry,ref=myapp:buildcache,mode=max
```

---

## Secrets and Variables

### Using Secrets

```yaml
steps:
  - run: deploy --token "$TOKEN"
    env:
      TOKEN: ${{ secrets.DEPLOY_TOKEN }}

  # NEVER do this (secret in command line, visible in logs)
  # - run: deploy --token ${{ secrets.DEPLOY_TOKEN }}
```

### Environment Secrets

```yaml
jobs:
  deploy:
    environment: production  # Uses production-scoped secrets
    steps:
      - run: deploy
        env:
          API_KEY: ${{ secrets.API_KEY }}  # Production-specific
```

### Variables (Non-Secret Configuration)

```yaml
# Set in GitHub Settings > Variables
steps:
  - run: echo "Deploying to ${{ vars.DEPLOY_URL }}"
  - run: echo "Region: ${{ vars.AWS_REGION }}"
```

### GITHUB_TOKEN Permissions

```yaml
permissions:
  contents: read        # Read repo contents
  issues: write         # Create/update issues
  pull-requests: write  # Comment on PRs
  packages: write       # Push to GHCR
  deployments: write    # Create deployments
  id-token: write       # OIDC token for cloud auth
  actions: read         # Read workflow runs
  checks: write         # Create check runs
  statuses: write       # Set commit statuses
  security-events: write # Upload SARIF
```

---

## Artifacts

### Upload and Download

```yaml
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - run: npm run build
      - uses: actions/upload-artifact@v4
        with:
          name: dist
          path: dist/
          retention-days: 5
          compression-level: 6    # 0-9, default 6

  deploy:
    needs: build
    runs-on: ubuntu-latest
    steps:
      - uses: actions/download-artifact@v4
        with:
          name: dist
          path: dist/
      - run: deploy dist/
```

### Multiple Artifacts

```yaml
# Upload multiple artifacts
- uses: actions/upload-artifact@v4
  with:
    name: test-results-${{ matrix.os }}
    path: test-results/

# Download all artifacts
- uses: actions/download-artifact@v4
  with:
    path: all-artifacts/
    # Creates: all-artifacts/test-results-ubuntu/
    #          all-artifacts/test-results-macos/
```

---

## Service Containers

```yaml
jobs:
  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:16-alpine
        env:
          POSTGRES_USER: test
          POSTGRES_PASSWORD: test
          POSTGRES_DB: testdb
        ports:
          - 5432:5432
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

      redis:
        image: redis:7-alpine
        ports:
          - 6379:6379
        options: >-
          --health-cmd "redis-cli ping"

      elasticsearch:
        image: elasticsearch:8.12.0
        ports:
          - 9200:9200
        env:
          discovery.type: single-node
          xpack.security.enabled: false
        options: >-
          --health-cmd "curl -f http://localhost:9200/_cluster/health"
          --health-interval 10s
          --health-timeout 5s

    env:
      DATABASE_URL: postgresql://test:test@localhost:5432/testdb
      REDIS_URL: redis://localhost:6379
      ELASTICSEARCH_URL: http://localhost:9200
    steps:
      - uses: actions/checkout@v4
      - run: npm ci
      - run: npm test
```

---

## Security Best Practices

### Pin Actions by SHA

```yaml
# Instead of mutable tags:
- uses: actions/checkout@v4

# Pin to specific commit SHA:
- uses: actions/checkout@b4ffde65f46336ab88eb53be808477a3936bae11  # v4.1.1

# Use Dependabot to auto-update pinned SHAs:
# .github/dependabot.yml
version: 2
updates:
  - package-ecosystem: "github-actions"
    directory: "/"
    schedule:
      interval: "weekly"
```

### Minimal Permissions

```yaml
# Start with read-only at workflow level
permissions:
  contents: read

# Grant additional permissions only where needed
jobs:
  publish:
    permissions:
      contents: read
      packages: write
      id-token: write
```

### Secure Script Injection Prevention

```yaml
# VULNERABLE — PR title injected into shell
- run: echo "PR: ${{ github.event.pull_request.title }}"

# SAFE — Use environment variable
- run: echo "PR: $TITLE"
  env:
    TITLE: ${{ github.event.pull_request.title }}

# SAFE — Use intermediate step
- id: sanitize
  uses: actions/github-script@v7
  with:
    result-encoding: string
    script: |
      return context.payload.pull_request.title.replace(/[^a-zA-Z0-9 ]/g, '')
```

### Dependency Review

```yaml
# Block PRs that introduce vulnerable dependencies
- uses: actions/dependency-review-action@v4
  with:
    fail-on-severity: moderate
    deny-licenses: GPL-3.0, AGPL-3.0
    comment-summary-in-pr: always
```

### CodeQL Analysis

```yaml
name: CodeQL

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]
  schedule:
    - cron: '0 6 * * 1'

jobs:
  analyze:
    runs-on: ubuntu-latest
    permissions:
      security-events: write
    strategy:
      matrix:
        language: [javascript, python]
    steps:
      - uses: actions/checkout@v4
      - uses: github/codeql-action/init@v3
        with:
          languages: ${{ matrix.language }}
      - uses: github/codeql-action/autobuild@v3
      - uses: github/codeql-action/analyze@v3
```

---

## Expression Syntax

### Context Objects

```yaml
# GitHub context
${{ github.actor }}                    # User who triggered
${{ github.repository }}               # owner/repo
${{ github.ref }}                      # refs/heads/main
${{ github.sha }}                      # Full commit SHA
${{ github.event_name }}               # push, pull_request, etc.
${{ github.run_number }}               # Incrementing run number
${{ github.workflow }}                 # Workflow name

# Event payload
${{ github.event.pull_request.number }}
${{ github.event.pull_request.head.sha }}
${{ github.event.release.tag_name }}

# Runner context
${{ runner.os }}                       # Linux, macOS, Windows
${{ runner.arch }}                     # X64, ARM64
${{ runner.temp }}                     # Temp directory

# Secrets and vars
${{ secrets.MY_SECRET }}
${{ vars.MY_VARIABLE }}
```

### Functions

```yaml
# String functions
${{ contains(github.event.pull_request.labels.*.name, 'deploy') }}
${{ startsWith(github.ref, 'refs/tags/') }}
${{ endsWith(github.repository, '-api') }}
${{ format('Hello {0}', github.actor) }}

# JSON functions
${{ fromJson(needs.setup.outputs.matrix) }}
${{ toJson(github.event) }}

# Hash functions
${{ hashFiles('**/package-lock.json') }}
${{ hashFiles('**/Cargo.lock', '**/Cargo.toml') }}

# Status functions (in if:)
${{ success() }}    # Previous steps succeeded
${{ failure() }}    # Any previous step failed
${{ cancelled() }}  # Workflow was cancelled
${{ always() }}     # Always run (even on failure/cancel)
```

---

## Useful Workflow Patterns

### Required Status Checks with Path Filtering

```yaml
# Problem: If CI only runs on src/ changes, docs-only PRs can't merge
# Solution: Use a "gate" job that always runs

name: CI

on:
  pull_request:

jobs:
  changes:
    runs-on: ubuntu-latest
    outputs:
      src: ${{ steps.filter.outputs.src }}
    steps:
      - uses: dorny/paths-filter@v3
        id: filter
        with:
          filters: |
            src:
              - 'src/**'
              - 'package.json'

  test:
    needs: changes
    if: needs.changes.outputs.src == 'true'
    runs-on: ubuntu-latest
    steps:
      - run: npm test

  # This job always runs — use as required status check
  ci-gate:
    if: always()
    needs: [test]
    runs-on: ubuntu-latest
    steps:
      - if: needs.test.result == 'failure'
        run: exit 1
      - run: echo "All checks passed"
```

### Auto-Merge Dependabot PRs

```yaml
name: Auto-merge Dependabot

on:
  pull_request:

permissions:
  contents: write
  pull-requests: write

jobs:
  auto-merge:
    if: github.actor == 'dependabot[bot]'
    runs-on: ubuntu-latest
    steps:
      - uses: dependabot/fetch-metadata@v2
        id: metadata

      - if: steps.metadata.outputs.update-type == 'version-update:semver-patch'
        run: gh pr merge --auto --squash "$PR_URL"
        env:
          PR_URL: ${{ github.event.pull_request.html_url }}
          GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

### Deploy Preview for PRs

```yaml
name: Deploy Preview

on:
  pull_request:
    types: [opened, synchronize]

permissions:
  pull-requests: write
  deployments: write

jobs:
  preview:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: npm ci && npm run build

      - name: Deploy preview
        id: deploy
        run: |
          URL=$(npx vercel deploy --token=${{ secrets.VERCEL_TOKEN }})
          echo "url=$URL" >> "$GITHUB_OUTPUT"

      - uses: marocchino/sticky-pull-request-comment@v2
        with:
          message: |
            ## Deploy Preview
            ${{ steps.deploy.outputs.url }}
```

### Notification on Failure

```yaml
jobs:
  notify:
    if: failure()
    needs: [lint, test, build]
    runs-on: ubuntu-latest
    steps:
      - uses: slackapi/slack-github-action@v1
        with:
          payload: |
            {
              "text": ":x: *${{ github.workflow }}* failed on `${{ github.ref_name }}`\n<${{ github.server_url }}/${{ github.repository }}/actions/runs/${{ github.run_id }}|View Run>",
              "unfurl_links": false
            }
        env:
          SLACK_WEBHOOK_URL: ${{ secrets.SLACK_WEBHOOK_URL }}
```

---

## Runner Types and Cost

| Runner | vCPU | RAM | Cost/min (private) |
|--------|------|-----|--------------------|
| ubuntu-latest | 4 | 16GB | $0.008 |
| ubuntu-latest-4-cores | 4 | 16GB | $0.016 |
| ubuntu-latest-8-cores | 8 | 32GB | $0.032 |
| ubuntu-latest-16-cores | 16 | 64GB | $0.064 |
| macos-latest (M1) | 3 | 7GB | $0.08 |
| macos-latest-xlarge | 12 | 30GB | $0.12 |
| windows-latest | 4 | 16GB | $0.016 |

Free tier: 2,000 minutes/month for public repos (unlimited), private repos.

### Self-Hosted Runners

```yaml
jobs:
  build:
    runs-on: [self-hosted, linux, x64, gpu]
    steps:
      - uses: actions/checkout@v4
      - run: make build
```

---

## Debugging Workflows

### Enable Debug Logging

```yaml
# Set secret ACTIONS_STEP_DEBUG = true for step-level debug logs
# Set secret ACTIONS_RUNNER_DEBUG = true for runner-level debug logs

# Or per-run:
env:
  ACTIONS_STEP_DEBUG: true
```

### Local Testing with act

```bash
# Install act
brew install act

# Run default event (push)
act

# Run specific event
act pull_request

# Run specific job
act -j test

# With secrets
act --secret-file .env.secrets

# List available workflows
act -l

# Dry run (show what would run)
act -n
```

### actionlint

```bash
# Install
brew install actionlint

# Lint all workflows
actionlint

# Common issues it catches:
# - Invalid YAML syntax
# - Unknown action inputs
# - Type mismatches in expressions
# - Missing required permissions
# - Deprecated features
```
