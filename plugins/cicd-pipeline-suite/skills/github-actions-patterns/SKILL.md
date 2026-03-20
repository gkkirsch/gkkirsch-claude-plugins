---
name: github-actions-patterns
description: >
  GitHub Actions patterns — workflow design, reusable workflows, composite actions,
  matrix strategies, caching, artifacts, environment protection rules,
  secrets management, and self-hosted runners.
  Triggers: "github actions", "github workflow", "ci pipeline", "github ci",
  "workflow dispatch", "reusable workflow", "composite action".
  NOT for: Pipeline security hardening (use pipeline-security).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# GitHub Actions Patterns

## Standard CI Workflow

```yaml
# .github/workflows/ci.yml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

concurrency:
  group: ci-${{ github.ref }}
  cancel-in-progress: ${{ github.event_name == 'pull_request' }}

permissions:
  contents: read
  pull-requests: write

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: 20
          cache: 'npm'
      - run: npm ci
      - run: npm run lint
      - run: npm run typecheck

  test:
    runs-on: ubuntu-latest
    needs: lint
    strategy:
      fail-fast: false
      matrix:
        shard: [1, 2, 3, 4]
    services:
      postgres:
        image: postgres:16-alpine
        env:
          POSTGRES_DB: test
          POSTGRES_USER: test
          POSTGRES_PASSWORD: test
        ports: ['5432:5432']
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
      redis:
        image: redis:7-alpine
        ports: ['6379:6379']
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: 20
          cache: 'npm'
      - run: npm ci

      - name: Run tests (shard ${{ matrix.shard }}/4)
        run: npm test -- --shard=${{ matrix.shard }}/4 --coverage
        env:
          DATABASE_URL: postgres://test:test@localhost:5432/test
          REDIS_URL: redis://localhost:6379
          NODE_ENV: test

      - name: Upload coverage
        uses: actions/upload-artifact@v4
        with:
          name: coverage-${{ matrix.shard }}
          path: coverage/

  coverage:
    needs: test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/download-artifact@v4
        with:
          pattern: coverage-*
          merge-multiple: true
          path: coverage/

      - name: Check coverage threshold
        run: |
          COVERAGE=$(npx nyc report --reporter=text-summary | grep 'Statements' | awk '{print $3}' | tr -d '%')
          echo "Coverage: ${COVERAGE}%"
          if (( $(echo "$COVERAGE < 80" | bc -l) )); then
            echo "::error::Coverage ${COVERAGE}% is below 80% threshold"
            exit 1
          fi

  build:
    needs: test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: 20
          cache: 'npm'
      - run: npm ci
      - run: npm run build

      - uses: actions/upload-artifact@v4
        with:
          name: build
          path: dist/
          retention-days: 7
```

## Reusable Workflows

```yaml
# .github/workflows/reusable-deploy.yml
name: Reusable Deploy

on:
  workflow_call:
    inputs:
      environment:
        required: true
        type: string
      image-tag:
        required: true
        type: string
      helm-values-file:
        required: false
        type: string
        default: values.yaml
    secrets:
      KUBE_CONFIG:
        required: true
      SLACK_WEBHOOK:
        required: false
    outputs:
      deploy-url:
        description: "Deployed URL"
        value: ${{ jobs.deploy.outputs.url }}

jobs:
  deploy:
    runs-on: ubuntu-latest
    environment:
      name: ${{ inputs.environment }}
      url: ${{ steps.deploy.outputs.url }}
    outputs:
      url: ${{ steps.deploy.outputs.url }}
    steps:
      - uses: actions/checkout@v4

      - name: Configure kubectl
        run: |
          mkdir -p ~/.kube
          echo "${{ secrets.KUBE_CONFIG }}" | base64 -d > ~/.kube/config

      - name: Deploy via Helm
        id: deploy
        run: |
          helm upgrade --install api-server ./charts/api-server \
            --namespace ${{ inputs.environment }} \
            --set image.tag=${{ inputs.image-tag }} \
            --values ./charts/api-server/${{ inputs.helm-values-file }} \
            --wait --timeout 10m

          URL=$(kubectl get ingress -n ${{ inputs.environment }} -o jsonpath='{.items[0].spec.rules[0].host}')
          echo "url=https://${URL}" >> "$GITHUB_OUTPUT"

      - name: Smoke test
        run: |
          for i in {1..10}; do
            if curl -sf "${{ steps.deploy.outputs.url }}/health" > /dev/null; then
              echo "Health check passed"
              exit 0
            fi
            sleep 5
          done
          exit 1

      - name: Notify Slack
        if: always() && secrets.SLACK_WEBHOOK != ''
        run: |
          STATUS="${{ job.status }}"
          COLOR=$([[ "$STATUS" == "success" ]] && echo "good" || echo "danger")
          curl -X POST "${{ secrets.SLACK_WEBHOOK }}" \
            -H 'Content-Type: application/json' \
            -d "{\"attachments\":[{\"color\":\"$COLOR\",\"text\":\"Deploy to ${{ inputs.environment }}: $STATUS\"}]}"

# Caller workflow
# .github/workflows/deploy-production.yml
name: Deploy Production
on:
  push:
    branches: [main]
jobs:
  deploy:
    uses: ./.github/workflows/reusable-deploy.yml
    with:
      environment: production
      image-tag: ${{ github.sha }}
      helm-values-file: values-production.yaml
    secrets:
      KUBE_CONFIG: ${{ secrets.PROD_KUBE_CONFIG }}
      SLACK_WEBHOOK: ${{ secrets.SLACK_WEBHOOK }}
```

## Composite Actions

```yaml
# .github/actions/setup-project/action.yml
name: Setup Project
description: Install dependencies with caching

inputs:
  node-version:
    description: Node.js version
    required: false
    default: '20'
  install-command:
    description: Install command to run
    required: false
    default: 'npm ci'

runs:
  using: composite
  steps:
    - uses: actions/setup-node@v4
      with:
        node-version: ${{ inputs.node-version }}
        cache: 'npm'

    - name: Cache node_modules
      uses: actions/cache@v4
      id: cache
      with:
        path: node_modules
        key: node-modules-${{ runner.os }}-${{ hashFiles('package-lock.json') }}

    - name: Install dependencies
      if: steps.cache.outputs.cache-hit != 'true'
      shell: bash
      run: ${{ inputs.install-command }}

    - name: Show versions
      shell: bash
      run: |
        echo "Node: $(node --version)"
        echo "npm: $(npm --version)"

# Usage in workflows:
# - uses: ./.github/actions/setup-project
#   with:
#     node-version: '20'
```

## Matrix Strategy Patterns

```yaml
# Cross-platform and multi-version testing
jobs:
  test:
    strategy:
      fail-fast: false
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
        node: [18, 20, 22]
        exclude:
          - os: windows-latest
            node: 18  # Drop old Node on Windows
        include:
          - os: ubuntu-latest
            node: 20
            coverage: true  # Only generate coverage for one combo
    runs-on: ${{ matrix.os }}
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: ${{ matrix.node }}
      - run: npm ci
      - run: npm test ${{ matrix.coverage && '-- --coverage' || '' }}

  # Dynamic matrix from JSON
  generate-matrix:
    runs-on: ubuntu-latest
    outputs:
      matrix: ${{ steps.set.outputs.matrix }}
    steps:
      - uses: actions/checkout@v4
      - id: set
        run: |
          # Generate matrix from changed directories
          CHANGED=$(git diff --name-only HEAD~1 | grep '^packages/' | cut -d/ -f2 | sort -u)
          MATRIX=$(echo "$CHANGED" | jq -R -s 'split("\n") | map(select(. != "")) | {package: .}')
          echo "matrix=$MATRIX" >> "$GITHUB_OUTPUT"

  test-packages:
    needs: generate-matrix
    if: needs.generate-matrix.outputs.matrix != '{"package":[]}'
    strategy:
      matrix: ${{ fromJson(needs.generate-matrix.outputs.matrix) }}
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: npm test --workspace=packages/${{ matrix.package }}
```

## Caching Strategies

```yaml
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      # npm cache (built into setup-node)
      - uses: actions/setup-node@v4
        with:
          node-version: 20
          cache: 'npm'  # Caches ~/.npm

      # Docker layer cache
      - uses: docker/setup-buildx-action@v3
      - uses: docker/build-push-action@v6
        with:
          context: .
          push: false
          tags: myapp:test
          cache-from: type=gha
          cache-to: type=gha,mode=max

      # Custom cache (Turborepo, Next.js, etc.)
      - uses: actions/cache@v4
        with:
          path: |
            .next/cache
            .turbo
            **/node_modules/.cache
          key: build-cache-${{ runner.os }}-${{ hashFiles('**/*.ts', '**/*.tsx') }}
          restore-keys: |
            build-cache-${{ runner.os }}-

      # Persistent cache across runs (good for tools)
      - name: Cache Playwright browsers
        uses: actions/cache@v4
        with:
          path: ~/.cache/ms-playwright
          key: playwright-${{ runner.os }}-${{ hashFiles('package-lock.json') }}

      - run: npx playwright install --with-deps
```

## Workflow Dispatch (Manual Triggers)

```yaml
# .github/workflows/manual-deploy.yml
name: Manual Deploy

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
      version:
        description: 'Version to deploy (tag or SHA)'
        required: true
        type: string
      skip-tests:
        description: 'Skip test suite'
        required: false
        type: boolean
        default: false
      dry-run:
        description: 'Dry run (no actual deploy)'
        required: false
        type: boolean
        default: false

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - name: Validate version exists
        run: |
          if ! git rev-parse "${{ inputs.version }}" >/dev/null 2>&1; then
            echo "::error::Version '${{ inputs.version }}' not found"
            exit 1
          fi

  test:
    if: ${{ !inputs.skip-tests }}
    needs: validate
    uses: ./.github/workflows/ci.yml

  deploy:
    needs: [validate, test]
    if: always() && needs.validate.result == 'success' && (needs.test.result == 'success' || needs.test.result == 'skipped')
    uses: ./.github/workflows/reusable-deploy.yml
    with:
      environment: ${{ inputs.environment }}
      image-tag: ${{ inputs.version }}
    secrets: inherit
```

## Release Automation

```yaml
# .github/workflows/release.yml
name: Release

on:
  push:
    tags:
      - 'v*.*.*'

permissions:
  contents: write
  packages: write

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0  # Full history for changelog

      - name: Generate changelog
        id: changelog
        run: |
          PREV_TAG=$(git describe --tags --abbrev=0 HEAD^ 2>/dev/null || echo "")
          if [ -n "$PREV_TAG" ]; then
            CHANGES=$(git log ${PREV_TAG}..HEAD --pretty=format:"- %s (%h)" --no-merges)
          else
            CHANGES=$(git log --pretty=format:"- %s (%h)" --no-merges)
          fi
          echo "changes<<EOF" >> "$GITHUB_OUTPUT"
          echo "$CHANGES" >> "$GITHUB_OUTPUT"
          echo "EOF" >> "$GITHUB_OUTPUT"

      - name: Build and push Docker image
        uses: docker/build-push-action@v6
        with:
          context: .
          push: true
          tags: |
            registry.example.com/api:${{ github.ref_name }}
            registry.example.com/api:latest

      - name: Create GitHub Release
        uses: softprops/action-gh-release@v2
        with:
          body: |
            ## Changes
            ${{ steps.changelog.outputs.changes }}
          generate_release_notes: true
          files: |
            dist/*.tar.gz
```

## Gotchas

1. **`actions/checkout` doesn't fetch full history** — By default, `actions/checkout@v4` does a shallow clone (depth 1). Commands like `git log`, `git describe`, and changelog generation fail or produce wrong output. Add `fetch-depth: 0` for full history, or `fetch-depth: 2` for just the diff.

2. **Secret masking in logs breaks JSON** — GitHub masks secret values in logs by replacing them with `***`. If a secret appears inside JSON output, the masking corrupts the JSON structure, breaking downstream parsing. Use `--quiet` flags or redirect secrets to files instead of printing them.

3. **`concurrency` with `cancel-in-progress` kills deploy jobs** — Setting `cancel-in-progress: true` globally cancels running workflows, including active deployments. Only use it on PR workflows. For deploy workflows, use `cancel-in-progress: false` to let in-progress deployments finish.

4. **Matrix jobs share the same runner cache** — If matrix jobs write to the same cache key simultaneously, one wins and the others lose their work. Use `key: cache-${{ matrix.shard }}` to create per-job cache keys instead of sharing one.

5. **`needs` doesn't propagate on skip** — If job A is skipped (by `if` condition), job B with `needs: A` is also skipped. Use `if: always() && needs.A.result != 'failure'` on job B if you want it to run when A is skipped but not when A fails.

6. **Self-hosted runner state leaks between jobs** — Unlike GitHub-hosted runners, self-hosted runners persist between jobs. Docker containers, temp files, and environment variables from previous jobs leak into the next. Always clean up in `post` steps or use ephemeral runners with `--ephemeral` flag.
