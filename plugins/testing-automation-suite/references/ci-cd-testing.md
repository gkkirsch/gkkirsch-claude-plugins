# CI/CD Testing Best Practices Reference

Comprehensive reference for CI/CD testing strategies — parallelization, flaky test management, caching, test selection, and pipeline optimization.

## Pipeline Design Patterns

### The Testing Diamond in CI

```
┌─────────────────────────────────────────────────────────────────────┐
│                     CI Pipeline Stages                               │
│                                                                      │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌────────┐  │
│  │  Static      │→ │  Unit        │→ │  Integration │→ │  E2E   │  │
│  │  Analysis    │  │  Tests       │  │  Tests       │  │  Tests │  │
│  │              │  │              │  │              │  │        │  │
│  │  • Lint      │  │  • Fast      │  │  • Services  │  │  • UI  │  │
│  │  • Types     │  │  • Isolated  │  │  • Database  │  │  • User│  │
│  │  • Format    │  │  • Coverage  │  │  • APIs      │  │  flows │  │
│  │              │  │              │  │              │  │        │  │
│  │  ~30s-1min   │  │  ~1-3min     │  │  ~3-8min     │  │ ~5-15m │  │
│  └──────────────┘  └──────────────┘  └──────────────┘  └────────┘  │
│                                                                      │
│  Fail Fast ←─────────── Increasing cost & time ──────────→ Thorough │
└─────────────────────────────────────────────────────────────────────┘
```

### Stage Dependencies

```yaml
# GitHub Actions — optimal stage dependencies
#
# Stage 1: lint (30s)  ──┐
#                         ├── Stage 2a: unit-tests (2min)  ──┐
#                         └── Stage 2b: integration (5min) ──┼── Stage 3: e2e (10min)
#                                                            └── Stage 4: quality-gate
```

## Test Parallelization Strategies

### Vitest Sharding

```bash
# Split tests across N shards
# Each shard runs a subset of test files
npx vitest run --shard=1/4
npx vitest run --shard=2/4
npx vitest run --shard=3/4
npx vitest run --shard=4/4
```

```yaml
# GitHub Actions matrix strategy
jobs:
  test:
    strategy:
      fail-fast: false
      matrix:
        shard: [1, 2, 3, 4]
    steps:
      - run: npx vitest run --shard=${{ matrix.shard }}/4
```

### Playwright Sharding

```bash
# Playwright built-in sharding
npx playwright test --shard=1/4
npx playwright test --shard=2/4
npx playwright test --shard=3/4
npx playwright test --shard=4/4

# Merge blob reports after all shards complete
npx playwright merge-reports --reporter=html ./blob-reports
```

### pytest Sharding with pytest-split

```bash
# Install pytest-split
pip install pytest-split

# Generate test durations (run once)
pytest --store-durations

# Split by duration (optimal distribution)
pytest --splits 4 --group 1 --splitting-algorithm least_duration
pytest --splits 4 --group 2 --splitting-algorithm least_duration
pytest --splits 4 --group 3 --splitting-algorithm least_duration
pytest --splits 4 --group 4 --splitting-algorithm least_duration
```

### pytest-xdist Parallelization

```bash
# Auto-detect worker count
pytest -n auto

# Fixed worker count
pytest -n 4

# Dynamic load balancing (work stealing)
pytest -n auto --dist worksteal

# Group tests by module
pytest -n auto --dist loadgroup

# Each worker gets its own database
# conftest.py
@pytest.fixture(scope="session")
def db_name(worker_id):
    return f"test_db_{worker_id}"
```

### Go Test Parallelization

```bash
# Go tests run in parallel by default at the package level
go test ./...

# Control parallelism
go test ./... -parallel=8

# Run specific packages in parallel
go test -count=1 -race ./internal/...
```

## Flaky Test Management

### Detecting Flaky Tests

```bash
# Run tests multiple times to detect flakiness
# Vitest
for i in $(seq 1 10); do npx vitest run 2>&1 | tail -1; done

# Playwright
npx playwright test --repeat-each=5

# pytest
pytest --count=10  # with pytest-repeat
```

### Quarantine Strategy

```
┌─────────────────────────────────────────────────────┐
│              Flaky Test Lifecycle                     │
│                                                       │
│   Detected ──→ Quarantined ──→ Fixed ──→ Released    │
│                     │                                 │
│                     └── Runs in nightly pipeline      │
│                         (not blocking PRs)            │
│                                                       │
│   Quarantine rules:                                   │
│   1. Max 7 days in quarantine                        │
│   2. Must have assigned owner                        │
│   3. Tracked in issue tracker                        │
│   4. Auto-deleted if not fixed in 14 days            │
└─────────────────────────────────────────────────────┘
```

```typescript
// Playwright: Tag flaky tests
test('flaky interaction @quarantine', async ({ page }) => {
  // This test is quarantined
});

// Run quarantined tests separately
// npx playwright test --grep @quarantine       # Quarantine only
// npx playwright test --grep-invert @quarantine  # Everything else
```

```python
# pytest: Mark flaky tests
@pytest.mark.flaky(reruns=3, reason="Race condition in WebSocket handler")
def test_websocket_message():
    ...

# Or use custom marker
@pytest.mark.quarantine
def test_flaky_behavior():
    ...

# conftest.py
def pytest_collection_modifyitems(items):
    """Skip quarantined tests in CI unless explicitly running them."""
    if os.getenv("RUN_QUARANTINE") != "true":
        skip = pytest.mark.skip(reason="Quarantined test")
        for item in items:
            if "quarantine" in item.keywords:
                item.add_marker(skip)
```

### Common Flaky Test Causes and Fixes

```
┌──────────────────────┬──────────────────────────────────────────────┐
│ Cause                │ Fix                                          │
├──────────────────────┼──────────────────────────────────────────────┤
│ Race condition       │ Use waitFor/expect instead of setTimeout     │
│ Shared state         │ Isolate test data per test                  │
│ Time-dependent       │ Use fake timers / freeze time                │
│ Order-dependent      │ Ensure test independence, randomize order    │
│ Network flakiness    │ Mock external services, retry on 5xx         │
│ Port collision       │ Use port 0 (OS assigns available port)      │
│ File system races    │ Use in-memory alternatives or tmpdir         │
│ Browser timing       │ Use Playwright's auto-waiting assertions    │
│ Database deadlocks   │ Use per-test transactions, retry logic       │
│ Timezone issues      │ Set TZ=UTC in CI environment                │
│ Locale differences   │ Set LANG=en_US.UTF-8 in CI                  │
│ Floating point       │ Use toBeCloseTo instead of toBe              │
│ Random data          │ Seed faker, use deterministic data          │
│ CSS animations       │ Disable animations in test environment       │
│ Viewport differences │ Set explicit viewport size in config         │
└──────────────────────┴──────────────────────────────────────────────┘
```

## Caching Strategies

### What to Cache

```
┌──────────────────────────┬──────────────────┬────────────────────────┐
│ What                     │ Cache Key         │ Expected Savings       │
├──────────────────────────┼──────────────────┼────────────────────────┤
│ Node modules (pnpm)      │ pnpm-lock.yaml   │ 30-60 seconds         │
│ Node modules (npm)       │ package-lock.json │ 30-60 seconds         │
│ pip packages             │ requirements.txt  │ 20-40 seconds         │
│ Go modules               │ go.sum            │ 10-30 seconds         │
│ Playwright browsers      │ PW version        │ 60-120 seconds        │
│ Docker layers            │ Dockerfile hash   │ 60-180 seconds        │
│ Build artifacts          │ Source hash        │ 30-120 seconds        │
│ Next.js .next/cache      │ Source + lock      │ 30-60 seconds         │
│ Turborepo cache          │ turbo hash        │ Highly variable        │
│ pytest cache (.pytest_*)  │ requirements      │ 5-15 seconds          │
│ Rust target/             │ Cargo.lock        │ 120-300 seconds       │
│ .mypy_cache              │ requirements      │ 10-20 seconds         │
│ eslint cache             │ .eslintrc + lock  │ 5-15 seconds          │
└──────────────────────────┴──────────────────┴────────────────────────┘
```

### Cache Invalidation

```yaml
# GitHub Actions cache key patterns

# Exact match only — invalidates on any lockfile change
key: node-${{ runner.os }}-${{ hashFiles('pnpm-lock.yaml') }}

# With restore keys — falls back to partial matches
key: node-${{ runner.os }}-${{ hashFiles('pnpm-lock.yaml') }}
restore-keys: |
  node-${{ runner.os }}-

# Include OS + architecture
key: build-${{ runner.os }}-${{ runner.arch }}-${{ hashFiles('pnpm-lock.yaml') }}

# Multiple file hash
key: deps-${{ hashFiles('**/package.json', '**/pnpm-lock.yaml') }}

# With weekly rotation (prevent cache bloat)
key: node-${{ runner.os }}-week${{ github.run_number / 7 }}-${{ hashFiles('pnpm-lock.yaml') }}
```

## Test Selection and Optimization

### Changed-File-Based Test Selection

```yaml
# Only run tests affected by changed files
- name: Get changed files
  id: changed
  uses: tj-actions/changed-files@v44
  with:
    files: |
      src/**
      tests/**
      package.json
      pnpm-lock.yaml

- name: Run affected tests
  if: steps.changed.outputs.any_changed == 'true'
  run: |
    # Map changed source files to test files
    CHANGED_FILES="${{ steps.changed.outputs.all_changed_files }}"
    TEST_FILES=""
    for file in $CHANGED_FILES; do
      # Find corresponding test files
      test_file="${file%.ts}.test.ts"
      test_file="${test_file/src/tests}"
      if [ -f "$test_file" ]; then
        TEST_FILES="$TEST_FILES $test_file"
      fi
    done
    if [ -n "$TEST_FILES" ]; then
      npx vitest run $TEST_FILES
    fi
```

### Skip Conditions

```yaml
# Skip tests when only docs changed
- name: Check if tests needed
  id: check
  run: |
    CHANGED=$(git diff --name-only ${{ github.event.before }} ${{ github.sha }})
    if echo "$CHANGED" | grep -qvE '\.(md|txt|yml|yaml)$|^docs/|^\.github/'; then
      echo "run_tests=true" >> "$GITHUB_OUTPUT"
    else
      echo "run_tests=false" >> "$GITHUB_OUTPUT"
      echo "Only documentation changes, skipping tests"
    fi

- name: Run tests
  if: steps.check.outputs.run_tests == 'true'
  run: npm test
```

### Monorepo Test Selection

```yaml
# Turborepo — only test affected packages
- name: Test affected packages
  run: pnpm turbo test --filter='...[origin/${{ github.base_ref }}...HEAD]'

# Nx — affected tests
- name: Test affected
  run: npx nx affected -t test --base=origin/${{ github.base_ref }}
```

## Test Reports and Artifacts

### JUnit XML Report

```yaml
# Most CI systems understand JUnit XML
- name: Run tests
  run: npx vitest run --reporter=junit --outputFile=test-results/junit.xml

- name: Publish test results
  uses: mikepenz/action-junit-report@v4
  if: always()
  with:
    report_paths: test-results/junit.xml
    check_name: Test Results
    fail_on_failure: true
```

### Coverage Reports

```yaml
# Upload to Codecov
- name: Upload coverage
  uses: codecov/codecov-action@v4
  with:
    files: coverage/lcov.info
    token: ${{ secrets.CODECOV_TOKEN }}
    fail_ci_if_error: true

# Or Coveralls
- name: Upload coverage
  uses: coverallsapp/github-action@v2
  with:
    github-token: ${{ secrets.GITHUB_TOKEN }}
    path-to-lcov: coverage/lcov.info
```

### Playwright Report

```yaml
# Upload HTML report
- name: Upload Playwright report
  if: always()
  uses: actions/upload-artifact@v4
  with:
    name: playwright-report
    path: playwright-report/
    retention-days: 30

# Deploy report to GitHub Pages
- name: Deploy report
  if: always() && github.ref == 'refs/heads/main'
  uses: peaceiris/actions-gh-pages@v4
  with:
    github_token: ${{ secrets.GITHUB_TOKEN }}
    publish_dir: playwright-report
    destination_dir: test-reports/${{ github.run_id }}
```

### Test Timing Analysis

```yaml
# Track test durations over time
- name: Save test timings
  if: always()
  run: |
    if [ -f test-results/results.json ]; then
      # Extract slow tests (> 5 seconds)
      jq '[.testResults[].testResults[] | select(.duration > 5000) | {name: .fullName, duration: .duration}] | sort_by(.duration) | reverse' \
        test-results/results.json > slow-tests.json
      cat slow-tests.json
    fi

- name: Upload timing data
  if: always()
  uses: actions/upload-artifact@v4
  with:
    name: test-timings
    path: slow-tests.json
```

## Environment Management

### Service Containers

```yaml
# GitHub Actions services
services:
  postgres:
    image: postgres:16
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
    options: >-
      --health-cmd "redis-cli ping"
      --health-interval 10s
      --health-timeout 5s
      --health-retries 5

  elasticsearch:
    image: elasticsearch:8.12.0
    env:
      discovery.type: single-node
      xpack.security.enabled: false
    ports: ['9200:9200']
    options: >-
      --health-cmd "curl -s http://localhost:9200/_cluster/health"
      --health-interval 10s
      --health-timeout 5s
      --health-retries 10
```

### Environment Variables

```yaml
# GitHub Actions — secrets and env vars
env:
  NODE_ENV: test
  CI: true
  TZ: UTC
  DATABASE_URL: postgresql://test:test@localhost:5432/test
  REDIS_URL: redis://localhost:6379

# Use secrets for sensitive values
steps:
  - name: Run tests
    env:
      API_KEY: ${{ secrets.TEST_API_KEY }}
      STRIPE_TEST_KEY: ${{ secrets.STRIPE_TEST_KEY }}
    run: npm test
```

## Security in CI Testing

### Secrets Management

```yaml
# Never echo secrets
- run: |
    # ❌ BAD
    echo "Key is ${{ secrets.API_KEY }}"

    # ✅ GOOD — secrets are automatically masked
    curl -H "Authorization: Bearer ${{ secrets.API_KEY }}" https://api.example.com

# Limit secret exposure
- name: Run tests
  env:
    # Only expose what's needed
    DATABASE_URL: ${{ secrets.TEST_DATABASE_URL }}
  run: npm test
  # Don't pass ALL secrets to the environment
```

### Dependency Scanning

```yaml
# Run security audit as part of CI
- name: npm audit
  run: npm audit --production --audit-level=high

# Or with dedicated tools
- name: Snyk security scan
  uses: snyk/actions/node@master
  env:
    SNYK_TOKEN: ${{ secrets.SNYK_TOKEN }}
```

## Pipeline Optimization Checklist

```
Pipeline Speed Optimization:

□ Fail fast: Lint/typecheck before tests
□ Parallel stages: Independent jobs run concurrently
□ Test sharding: Split large test suites across workers
□ Caching: Dependencies, build artifacts, browser binaries
□ Incremental: Only run affected tests on PRs
□ Skip when possible: No code changes → skip tests
□ Cancel redundant: Cancel in-progress runs on new push
□ Build once: Share build artifacts across test stages
□ Right-size runners: Use appropriate runner specs
□ Minimize setup: Pre-built Docker images with deps

Pipeline Reliability:

□ Retries: E2E tests get 2 retries in CI
□ Timeouts: Every job has a timeout
□ Quarantine: Flaky tests run separately
□ Deterministic: Fake timers, seeded randoms, fixed locales
□ Isolated: Per-test data, no shared state
□ Health checks: Services ready before tests start
□ Artifacts: Screenshots/traces on failure
□ Notifications: Alert on main branch failures

Pipeline Security:

□ Secrets: Never logged, minimally exposed
□ Dependencies: Audited, lockfile committed
□ Containers: Pinned versions, scanned
□ Permissions: Minimal GITHUB_TOKEN scope
□ Fork safety: Don't run secrets on fork PRs
□ Supply chain: Verified action versions (pin to SHA)
```

## Quick Reference: CI Commands by Framework

```
┌──────────────┬──────────────────────────────────────────────────┐
│ Framework    │ CI Command                                        │
├──────────────┼──────────────────────────────────────────────────┤
│ Vitest       │ npx vitest run --coverage --reporter=junit       │
│ Jest         │ npx jest --ci --coverage --reporters=jest-junit  │
│ Playwright   │ npx playwright test --shard=N/M                  │
│ Cypress      │ npx cypress run --record --parallel              │
│ pytest       │ pytest --cov --cov-report=xml --junitxml=report  │
│ Go           │ go test ./... -v -race -coverprofile=cover.out   │
│ Rust         │ cargo test -- --test-threads=4                    │
│ .NET         │ dotnet test --logger "junit;LogFileName=test"    │
│ Ruby/RSpec   │ bundle exec rspec --format RspecJunitFormatter   │
│ PHPUnit      │ vendor/bin/phpunit --log-junit test-report.xml   │
└──────────────┴──────────────────────────────────────────────────┘
```
