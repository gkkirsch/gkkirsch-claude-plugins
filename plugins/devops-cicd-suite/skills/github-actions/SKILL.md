---
name: github-actions
description: >
  GitHub Actions CI/CD workflows — workflow syntax, triggers, jobs, steps,
  caching, matrix strategy, secrets, environments, reusable workflows,
  and composite actions.
  Triggers: "github actions", "ci/cd workflow", "github workflow", "actions yaml",
  "github actions cache", "github actions matrix", "reusable workflow",
  "composite action", "github actions deploy".
  NOT for: deployment strategies (use deployment-strategies), monitoring (use monitoring-observability).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# GitHub Actions CI/CD

## Workflow Basics

```yaml
# .github/workflows/ci.yml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

# Least-privilege permissions
permissions:
  contents: read
  pull-requests: write

# Cancel previous runs on same branch
concurrency:
  group: ci-${{ github.ref }}
  cancel-in-progress: true

jobs:
  lint-and-type-check:
    runs-on: ubuntu-22.04
    timeout-minutes: 10
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-node@v4
        with:
          node-version: 20
          cache: "npm"

      - run: npm ci

      - name: Lint
        run: npm run lint

      - name: Type check
        run: npx tsc --noEmit

  test:
    runs-on: ubuntu-22.04
    timeout-minutes: 15
    needs: lint-and-type-check
    strategy:
      matrix:
        node-version: [18, 20, 22]
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-node@v4
        with:
          node-version: ${{ matrix.node-version }}
          cache: "npm"

      - run: npm ci
      - run: npm test -- --coverage

      - name: Upload coverage
        if: matrix.node-version == 20
        uses: actions/upload-artifact@v4
        with:
          name: coverage
          path: coverage/
          retention-days: 5

  build:
    runs-on: ubuntu-22.04
    timeout-minutes: 10
    needs: test
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-node@v4
        with:
          node-version: 20
          cache: "npm"

      - run: npm ci
      - run: npm run build

      - uses: actions/upload-artifact@v4
        with:
          name: build
          path: dist/
          retention-days: 1
```

## Triggers (Complete Reference)

```yaml
on:
  # Push to specific branches
  push:
    branches: [main, "release/**"]
    paths: ["src/**", "package.json"]  # Only trigger on specific file changes
    paths-ignore: ["docs/**", "*.md"]

  # Pull request events
  pull_request:
    branches: [main]
    types: [opened, synchronize, reopened]

  # Manual trigger
  workflow_dispatch:
    inputs:
      environment:
        description: "Deploy target"
        required: true
        default: "staging"
        type: choice
        options: [staging, production]
      dry_run:
        description: "Dry run (no actual deploy)"
        type: boolean
        default: false

  # Scheduled (cron)
  schedule:
    - cron: "0 9 * * 1-5"  # Weekdays at 9am UTC

  # Called by another workflow
  workflow_call:
    inputs:
      environment:
        required: true
        type: string
    secrets:
      DEPLOY_KEY:
        required: true

  # On release
  release:
    types: [published]

  # On issue/PR comment
  issue_comment:
    types: [created]
```

## Caching

```yaml
# npm cache (built-in to setup-node)
- uses: actions/setup-node@v4
  with:
    node-version: 20
    cache: "npm"

# Custom cache
- uses: actions/cache@v4
  with:
    path: |
      ~/.npm
      node_modules
    key: ${{ runner.os }}-node-${{ hashFiles('**/package-lock.json') }}
    restore-keys: |
      ${{ runner.os }}-node-

# Turborepo remote cache
- uses: actions/cache@v4
  with:
    path: .turbo
    key: ${{ runner.os }}-turbo-${{ github.sha }}
    restore-keys: |
      ${{ runner.os }}-turbo-

# Docker layer caching
- uses: docker/build-push-action@v5
  with:
    context: .
    push: true
    tags: myapp:latest
    cache-from: type=gha
    cache-to: type=gha,mode=max

# Playwright browser caching
- uses: actions/cache@v4
  id: playwright-cache
  with:
    path: ~/.cache/ms-playwright
    key: playwright-${{ hashFiles('**/package-lock.json') }}

- if: steps.playwright-cache.outputs.cache-hit != 'true'
  run: npx playwright install --with-deps
```

## Environments and Secrets

```yaml
jobs:
  deploy-staging:
    runs-on: ubuntu-22.04
    environment:
      name: staging
      url: https://staging.myapp.com
    steps:
      - run: echo "Deploying to staging"
        env:
          DATABASE_URL: ${{ secrets.STAGING_DATABASE_URL }}
          API_KEY: ${{ secrets.API_KEY }}

  deploy-production:
    runs-on: ubuntu-22.04
    needs: deploy-staging
    environment:
      name: production
      url: https://myapp.com
    # Requires manual approval (configured in repo settings)
    steps:
      - run: echo "Deploying to production"
        env:
          DATABASE_URL: ${{ secrets.PROD_DATABASE_URL }}
```

## Matrix Strategy

```yaml
jobs:
  test:
    strategy:
      fail-fast: false  # Don't cancel other matrix jobs on failure
      matrix:
        os: [ubuntu-22.04, macos-14]
        node: [18, 20, 22]
        exclude:
          - os: macos-14
            node: 18
        include:
          - os: ubuntu-22.04
            node: 20
            coverage: true
    runs-on: ${{ matrix.os }}
    steps:
      - uses: actions/setup-node@v4
        with:
          node-version: ${{ matrix.node }}

      - run: npm test

      - if: matrix.coverage
        run: npm run test:coverage
```

## Reusable Workflows

```yaml
# .github/workflows/deploy-reusable.yml
name: Deploy (Reusable)

on:
  workflow_call:
    inputs:
      environment:
        required: true
        type: string
      app-name:
        required: true
        type: string
    secrets:
      HEROKU_API_KEY:
        required: true

jobs:
  deploy:
    runs-on: ubuntu-22.04
    environment: ${{ inputs.environment }}
    steps:
      - uses: actions/checkout@v4

      - name: Deploy to Heroku
        uses: akhileshns/heroku-deploy@v3
        with:
          heroku_api_key: ${{ secrets.HEROKU_API_KEY }}
          heroku_app_name: ${{ inputs.app-name }}

# .github/workflows/release.yml — caller workflow
name: Release
on:
  push:
    branches: [main]

jobs:
  deploy-staging:
    uses: ./.github/workflows/deploy-reusable.yml
    with:
      environment: staging
      app-name: myapp-staging
    secrets:
      HEROKU_API_KEY: ${{ secrets.HEROKU_API_KEY }}

  deploy-production:
    needs: deploy-staging
    uses: ./.github/workflows/deploy-reusable.yml
    with:
      environment: production
      app-name: myapp-prod
    secrets:
      HEROKU_API_KEY: ${{ secrets.HEROKU_API_KEY }}
```

## Composite Actions

```yaml
# .github/actions/setup-project/action.yml
name: Setup Project
description: Install Node.js, dependencies, and build tools

inputs:
  node-version:
    description: Node.js version
    default: "20"

runs:
  using: composite
  steps:
    - uses: actions/setup-node@v4
      with:
        node-version: ${{ inputs.node-version }}
        cache: "npm"

    - run: npm ci
      shell: bash

    - run: npx prisma generate
      shell: bash

# Usage in workflow
- uses: ./.github/actions/setup-project
  with:
    node-version: "20"
```

## Services (Database in CI)

```yaml
jobs:
  test:
    runs-on: ubuntu-22.04
    services:
      postgres:
        image: postgres:16-alpine
        env:
          POSTGRES_USER: test
          POSTGRES_PASSWORD: test
          POSTGRES_DB: myapp_test
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
          --health-interval 10s

    steps:
      - uses: actions/checkout@v4
      - run: npm ci
      - run: npm test
        env:
          DATABASE_URL: postgresql://test:test@localhost:5432/myapp_test
          REDIS_URL: redis://localhost:6379
```

## Deploy Patterns

```yaml
# Heroku deploy
- name: Deploy to Heroku
  env:
    HEROKU_API_KEY: ${{ secrets.HEROKU_API_KEY }}
  run: |
    git remote add heroku https://heroku:$HEROKU_API_KEY@git.heroku.com/${{ vars.HEROKU_APP }}.git
    git push heroku main

# Cloudflare Pages
- name: Deploy to Cloudflare Pages
  uses: cloudflare/wrangler-action@v3
  with:
    apiToken: ${{ secrets.CF_API_TOKEN }}
    command: pages deploy dist --project-name=myapp

# Docker + Registry
- uses: docker/login-action@v3
  with:
    registry: ghcr.io
    username: ${{ github.actor }}
    password: ${{ secrets.GITHUB_TOKEN }}

- uses: docker/build-push-action@v5
  with:
    push: true
    tags: ghcr.io/${{ github.repository }}:${{ github.sha }}

# SSH deploy
- name: Deploy via SSH
  uses: appleboy/ssh-action@v1
  with:
    host: ${{ secrets.SSH_HOST }}
    username: deploy
    key: ${{ secrets.SSH_KEY }}
    script: |
      cd /app
      git pull origin main
      npm ci --production
      pm2 restart all
```

## Gotchas

1. **`actions/checkout` only fetches 1 commit by default.** If you need git history (for changelogs, tags, blame), add `fetch-depth: 0` for full history or `fetch-depth: 2` for the diff.

2. **`npm install` vs `npm ci` in CI.** Always use `npm ci` in CI — it's faster (skips resolution), deterministic (uses lockfile exactly), and fails if lockfile is out of sync. `npm install` can modify the lockfile.

3. **Secrets are not available in PRs from forks.** For security, `secrets.*` context is empty for pull requests from forked repos. Use `pull_request_target` carefully if you need secrets for PR checks.

4. **`concurrency` without `cancel-in-progress` just queues.** Without `cancel-in-progress: true`, concurrent runs queue and all eventually execute. With it, previous runs are cancelled. Usually you want the cancellation for CI, but NOT for deploys.

5. **Matrix `fail-fast` is `true` by default.** One failing matrix combination cancels all others. Set `fail-fast: false` if you want to see all failures (useful for cross-platform testing).

6. **Self-hosted runner security.** Self-hosted runners persist state between jobs. Never use them for public repos (malicious PRs can execute arbitrary code). Always use ephemeral runners or container-based isolation for public repos.
