# Testing Pipeline Reference

Comprehensive reference for building testing pipelines in CI/CD — test pyramid strategy, parallel testing, flaky test management, test infrastructure, and coverage enforcement.

---

## The Test Pyramid

```
        /  E2E  \          Few, slow, expensive
       /  Tests  \         Validate full user flows
      /___________\
     / Integration \       Medium count, moderate speed
    /    Tests      \      Test component interactions
   /________________\
  /    Unit Tests    \     Many, fast, cheap
 /                    \    Test individual functions
/______________________\
```

### How Many of Each?

| Type | Count | Speed | Cost | Reliability |
|------|-------|-------|------|-------------|
| Unit | 70-80% | ms per test | Cheapest | Very High |
| Integration | 15-20% | seconds per test | Medium | High |
| E2E | 5-10% | minutes per test | Expensive | Medium |

### CI Pipeline Order

```yaml
# Run cheapest/fastest tests first → fail fast
jobs:
  lint:                    # Seconds
    runs-on: ubuntu-latest
    steps:
      - run: npm run lint

  typecheck:               # Seconds
    runs-on: ubuntu-latest
    steps:
      - run: npm run typecheck

  unit-tests:              # Seconds to minutes
    runs-on: ubuntu-latest
    steps:
      - run: npm run test:unit

  integration-tests:       # Minutes
    needs: unit-tests
    runs-on: ubuntu-latest
    services:
      postgres: [...]
    steps:
      - run: npm run test:integration

  e2e-tests:               # Minutes to tens of minutes
    needs: unit-tests
    runs-on: ubuntu-latest
    steps:
      - run: npm run test:e2e

  # Only deploy if everything passes
  deploy:
    needs: [lint, typecheck, unit-tests, integration-tests, e2e-tests]
    runs-on: ubuntu-latest
    steps:
      - run: deploy
```

---

## Unit Testing in CI

### Jest Configuration for CI

```javascript
// jest.config.ci.js
module.exports = {
  ...require('./jest.config'),
  // CI-specific overrides
  ci: true,
  bail: 1,                          // Stop on first failure in CI
  maxWorkers: '50%',                // Use half CPU cores
  coverageThreshold: {
    global: {
      branches: 80,
      functions: 80,
      lines: 80,
      statements: 80,
    },
  },
  reporters: [
    'default',
    ['jest-junit', {
      outputDirectory: '.',
      outputName: 'junit.xml',
      classNameTemplate: '{classname}',
      titleTemplate: '{title}',
    }],
  ],
};
```

### GitHub Actions: Jest with Coverage

```yaml
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: 20
          cache: 'npm'
      - run: npm ci

      - name: Run tests
        run: npm test -- --coverage --ci --reporters=default --reporters=jest-junit
        env:
          JEST_JUNIT_OUTPUT_DIR: .

      # Upload test results
      - uses: dorny/test-reporter@v1
        if: always()
        with:
          name: Jest Tests
          path: junit.xml
          reporter: jest-junit

      # Coverage comment on PR
      - uses: davelosert/vitest-coverage-report-action@v2
        if: github.event_name == 'pull_request'
        with:
          json-summary-path: coverage/coverage-summary.json
```

### Vitest Configuration for CI

```typescript
// vitest.config.ts
import { defineConfig } from 'vitest/config';

export default defineConfig({
  test: {
    reporters: process.env.CI
      ? ['default', 'junit']
      : ['default'],
    outputFile: {
      junit: './junit.xml',
    },
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json-summary', 'lcov'],
      thresholds: {
        branches: 80,
        functions: 80,
        lines: 80,
        statements: 80,
      },
    },
    pool: 'forks',           // More stable in CI than threads
    poolOptions: {
      forks: {
        maxForks: '50%',
      },
    },
  },
});
```

---

## Integration Testing in CI

### Database Testing with Service Containers

```yaml
jobs:
  integration:
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
          --health-interval 10s

    env:
      DATABASE_URL: postgresql://test:test@localhost:5432/testdb
      REDIS_URL: redis://localhost:6379

    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: 20
          cache: 'npm'

      - run: npm ci
      - run: npx prisma migrate deploy
      - run: npm run test:integration
```

### API Integration Tests

```typescript
// tests/integration/api.test.ts
import { describe, it, expect, beforeAll, afterAll } from 'vitest';
import { createApp } from '../../src/app';
import { createTestDatabase, destroyTestDatabase } from '../helpers/db';

let app: Express;
let db: Database;

beforeAll(async () => {
  db = await createTestDatabase();
  app = createApp({ database: db });
});

afterAll(async () => {
  await destroyTestDatabase(db);
});

describe('POST /api/users', () => {
  it('creates a user and returns 201', async () => {
    const response = await request(app)
      .post('/api/users')
      .send({ name: 'Test User', email: 'test@example.com' })
      .expect(201);

    expect(response.body).toMatchObject({
      name: 'Test User',
      email: 'test@example.com',
      id: expect.any(String),
    });

    // Verify in database
    const user = await db.query('SELECT * FROM users WHERE id = $1', [response.body.id]);
    expect(user.rows).toHaveLength(1);
  });
});
```

### Test Database Isolation

```typescript
// tests/helpers/db.ts
import { Pool } from 'pg';
import { randomUUID } from 'crypto';

export async function createTestDatabase(): Promise<Pool> {
  const dbName = `test_${randomUUID().replace(/-/g, '')}`;
  const adminPool = new Pool({
    connectionString: process.env.DATABASE_URL,
  });

  await adminPool.query(`CREATE DATABASE "${dbName}"`);
  await adminPool.end();

  const testPool = new Pool({
    connectionString: process.env.DATABASE_URL?.replace(/\/[^/]+$/, `/${dbName}`),
  });

  // Run migrations
  await runMigrations(testPool);

  return testPool;
}

export async function destroyTestDatabase(pool: Pool): Promise<void> {
  const dbName = (pool as any).options.database;
  await pool.end();

  const adminPool = new Pool({
    connectionString: process.env.DATABASE_URL,
  });
  await adminPool.query(`DROP DATABASE IF EXISTS "${dbName}"`);
  await adminPool.end();
}
```

---

## E2E Testing in CI

### Playwright Setup

```yaml
jobs:
  e2e:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: 20
          cache: 'npm'

      - run: npm ci

      - name: Install Playwright browsers
        run: npx playwright install --with-deps chromium

      - name: Build application
        run: npm run build

      - name: Run E2E tests
        run: npx playwright test
        env:
          BASE_URL: http://localhost:3000

      - uses: actions/upload-artifact@v4
        if: failure()
        with:
          name: playwright-report
          path: playwright-report/
          retention-days: 7

      - uses: actions/upload-artifact@v4
        if: failure()
        with:
          name: test-results
          path: test-results/
          retention-days: 7
```

### Playwright Config for CI

```typescript
// playwright.config.ts
import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
  testDir: './tests/e2e',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,        // Fail if .only() left in code
  retries: process.env.CI ? 2 : 0,     // Retry flaky tests in CI
  workers: process.env.CI ? 2 : undefined,
  reporter: process.env.CI
    ? [['html'], ['junit', { outputFile: 'e2e-results.xml' }]]
    : [['html']],

  use: {
    baseURL: process.env.BASE_URL || 'http://localhost:3000',
    trace: 'on-first-retry',           // Capture trace on retry
    screenshot: 'only-on-failure',
    video: 'on-first-retry',
  },

  projects: [
    { name: 'chromium', use: { ...devices['Desktop Chrome'] } },
    // Only test multiple browsers in CI
    ...(process.env.CI ? [
      { name: 'firefox', use: { ...devices['Desktop Firefox'] } },
      { name: 'webkit', use: { ...devices['Desktop Safari'] } },
    ] : []),
  ],

  // Start dev server before tests
  webServer: {
    command: 'npm run preview',
    port: 3000,
    reuseExistingServer: !process.env.CI,
    timeout: 120_000,
  },
});
```

### Cypress Setup for CI

```yaml
jobs:
  e2e:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: cypress-io/github-action@v6
        with:
          build: npm run build
          start: npm run preview
          wait-on: 'http://localhost:3000'
          browser: chrome
          record: true
        env:
          CYPRESS_RECORD_KEY: ${{ secrets.CYPRESS_RECORD_KEY }}

      - uses: actions/upload-artifact@v4
        if: failure()
        with:
          name: cypress-screenshots
          path: cypress/screenshots/
```

---

## Parallel Test Execution

### Jest Parallelization

```yaml
jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        shard: [1, 2, 3, 4]
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: 20
          cache: 'npm'
      - run: npm ci
      - run: npx jest --shard=${{ matrix.shard }}/${{ strategy.job-total }}
```

### Playwright Sharding

```yaml
jobs:
  e2e:
    strategy:
      matrix:
        shard: [1/4, 2/4, 3/4, 4/4]
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: 20
          cache: 'npm'
      - run: npm ci
      - run: npx playwright install --with-deps chromium
      - run: npx playwright test --shard=${{ matrix.shard }}

      - uses: actions/upload-artifact@v4
        if: always()
        with:
          name: blob-report-${{ strategy.job-index }}
          path: blob-report/
          retention-days: 1

  merge-reports:
    if: always()
    needs: e2e
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: 20
      - run: npm ci

      - uses: actions/download-artifact@v4
        with:
          pattern: blob-report-*
          path: all-blob-reports/
          merge-multiple: true

      - run: npx playwright merge-reports --reporter html ./all-blob-reports

      - uses: actions/upload-artifact@v4
        with:
          name: html-report
          path: playwright-report/
```

### pytest Parallelization

```yaml
jobs:
  test:
    strategy:
      matrix:
        group: [1, 2, 3, 4]
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-python@v5
        with:
          python-version: '3.12'
      - run: pip install -r requirements-test.txt
      - run: |
          pytest --splits 4 --group ${{ matrix.group }} \
            --splitting-algorithm least_duration \
            --store-durations --durations-path .test_durations
```

---

## Flaky Test Management

### Identifying Flaky Tests

A test is flaky if it passes and fails without code changes. Common causes:

1. **Timing dependencies** — Tests that depend on setTimeout, animations, or race conditions
2. **Shared state** — Tests that read/write shared databases, files, or global variables
3. **Non-deterministic data** — Tests using random data, Date.now(), or UUIDs
4. **External dependencies** — Tests hitting real APIs, DNS, or network resources
5. **Order dependency** — Tests that only pass when run in a specific order

### Detecting Flakes in CI

```yaml
# Re-run failed tests to detect flakes
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - run: npm test || npm test --only-failures
      # If first run fails but retry passes, it's a flake
```

### Jest: Retry Flaky Tests

```javascript
// jest.config.js
module.exports = {
  // Retry failed tests once (CI only)
  ...(process.env.CI && { retryTimes: 1 }),
};
```

### Playwright: Built-in Retry

```typescript
// playwright.config.ts
export default defineConfig({
  retries: process.env.CI ? 2 : 0,
  // Tests that pass on retry are marked as "flaky" in reports
});
```

### Quarantine Flaky Tests

```typescript
// Mark known flaky tests
describe.skip('flaky: Payment integration', () => {
  // TODO: Fix race condition in webhook handler
  // Tracking: JIRA-1234
  it('processes webhook correctly', () => {
    // ...
  });
});

// Or use a custom tag
test('processes webhook', { tag: '@flaky' }, () => {
  // ...
});
```

```yaml
# Run quarantined tests separately (non-blocking)
jobs:
  stable-tests:
    runs-on: ubuntu-latest
    steps:
      - run: npx jest --testPathIgnorePatterns='flaky'

  flaky-tests:
    runs-on: ubuntu-latest
    continue-on-error: true  # Don't block deployment
    steps:
      - run: npx jest --testPathPattern='flaky'
```

### Fixing Common Flakes

```typescript
// FLAKY: Race condition with setTimeout
test('debounced search', async () => {
  fireEvent.change(input, { target: { value: 'test' } });
  await new Promise(r => setTimeout(r, 500));  // Brittle!
  expect(results).toBeVisible();
});

// FIXED: Use waitFor
test('debounced search', async () => {
  fireEvent.change(input, { target: { value: 'test' } });
  await waitFor(() => {
    expect(results).toBeVisible();
  });
});

// FLAKY: Shared database state
test('creates user', async () => {
  await db.query("INSERT INTO users (email) VALUES ('test@example.com')");
  // Fails if another test already inserted this email
});

// FIXED: Unique data per test
test('creates user', async () => {
  const email = `test-${randomUUID()}@example.com`;
  await db.query('INSERT INTO users (email) VALUES ($1)', [email]);
});

// FLAKY: Date-dependent test
test('shows today badge', () => {
  expect(isToday(new Date())).toBe(true);
  // Fails at midnight boundary
});

// FIXED: Mock the clock
test('shows today badge', () => {
  vi.useFakeTimers();
  vi.setSystemTime(new Date('2024-06-15T12:00:00Z'));
  expect(isToday(new Date())).toBe(true);
  vi.useRealTimers();
});
```

---

## Coverage Enforcement

### Coverage Thresholds in CI

```yaml
- name: Check coverage
  run: |
    npx vitest run --coverage

    # Extract coverage percentage
    COVERAGE=$(node -e "
      const summary = require('./coverage/coverage-summary.json');
      console.log(summary.total.lines.pct);
    ")

    echo "Coverage: ${COVERAGE}%"

    # Fail if coverage dropped
    if (( $(echo "$COVERAGE < 80" | bc -l) )); then
      echo "Coverage below threshold: ${COVERAGE}% < 80%"
      exit 1
    fi
```

### Coverage Diff on PRs

```yaml
# Show coverage changes on pull requests
- uses: davelosert/vitest-coverage-report-action@v2
  if: github.event_name == 'pull_request'
  with:
    json-summary-path: coverage/coverage-summary.json
    json-final-path: coverage/coverage-final.json

# Or with Codecov
- uses: codecov/codecov-action@v4
  with:
    token: ${{ secrets.CODECOV_TOKEN }}
    files: coverage/lcov.info
    fail_ci_if_error: true
```

### Ratcheting Coverage (Never Decrease)

```yaml
- name: Check coverage didn't decrease
  run: |
    # Get current coverage from main branch
    git fetch origin main
    git checkout origin/main -- coverage/coverage-summary.json 2>/dev/null || true

    if [ -f coverage/coverage-summary.json ]; then
      MAIN_COVERAGE=$(node -e "console.log(require('./coverage/coverage-summary.json').total.lines.pct)")
    else
      MAIN_COVERAGE=0
    fi

    git checkout - -- coverage/coverage-summary.json

    PR_COVERAGE=$(node -e "console.log(require('./coverage/coverage-summary.json').total.lines.pct)")

    echo "Main: ${MAIN_COVERAGE}%, PR: ${PR_COVERAGE}%"

    if (( $(echo "$PR_COVERAGE < $MAIN_COVERAGE" | bc -l) )); then
      echo "Coverage decreased! ${PR_COVERAGE}% < ${MAIN_COVERAGE}%"
      exit 1
    fi
```

---

## Test Data Management

### Fixtures and Factories

```typescript
// tests/factories/user.ts
import { faker } from '@faker-js/faker';

export function createUser(overrides: Partial<User> = {}): User {
  return {
    id: faker.string.uuid(),
    name: faker.person.fullName(),
    email: faker.internet.email(),
    createdAt: faker.date.past(),
    ...overrides,
  };
}

// Usage in tests
const user = createUser({ name: 'Alice' });
const users = Array.from({ length: 10 }, () => createUser());
```

### Database Seeding for CI

```typescript
// tests/seed.ts
export async function seedTestData(db: Database) {
  await db.query(`
    INSERT INTO users (id, name, email) VALUES
    ('user-1', 'Alice', 'alice@example.com'),
    ('user-2', 'Bob', 'bob@example.com')
  `);

  await db.query(`
    INSERT INTO orders (id, user_id, total) VALUES
    ('order-1', 'user-1', 99.99),
    ('order-2', 'user-2', 149.99)
  `);
}

// Clean up between tests
export async function cleanTestData(db: Database) {
  await db.query('TRUNCATE users, orders CASCADE');
}
```

### Snapshot Testing

```typescript
// Good use of snapshots: stable outputs
test('renders user profile', () => {
  const { container } = render(<UserProfile user={testUser} />);
  expect(container).toMatchSnapshot();
});

// Bad use: frequently changing outputs
// Avoid snapshotting timestamps, random IDs, etc.

// Better: inline snapshots for small outputs
test('formats currency', () => {
  expect(formatCurrency(1234.5)).toMatchInlineSnapshot('"$1,234.50"');
});
```

---

## Testing Performance in CI

### Lighthouse CI

```yaml
jobs:
  lighthouse:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: 20
          cache: 'npm'
      - run: npm ci && npm run build

      - name: Run Lighthouse CI
        run: |
          npm install -g @lhci/cli
          lhci autorun
        env:
          LHCI_GITHUB_APP_TOKEN: ${{ secrets.LHCI_GITHUB_APP_TOKEN }}
```

```javascript
// lighthouserc.js
module.exports = {
  ci: {
    collect: {
      startServerCommand: 'npm run preview',
      url: ['http://localhost:3000', 'http://localhost:3000/dashboard'],
      numberOfRuns: 3,
    },
    assert: {
      assertions: {
        'categories:performance': ['error', { minScore: 0.9 }],
        'categories:accessibility': ['error', { minScore: 0.95 }],
        'categories:best-practices': ['error', { minScore: 0.9 }],
        'categories:seo': ['error', { minScore: 0.9 }],
        'first-contentful-paint': ['warn', { maxNumericValue: 2000 }],
        'largest-contentful-paint': ['error', { maxNumericValue: 2500 }],
        'cumulative-layout-shift': ['error', { maxNumericValue: 0.1 }],
        'total-blocking-time': ['error', { maxNumericValue: 300 }],
      },
    },
    upload: {
      target: 'temporary-public-storage',
    },
  },
};
```

### Bundle Size Tracking

```yaml
- name: Check bundle size
  uses: andresz1/size-limit-action@v1
  with:
    github_token: ${{ secrets.GITHUB_TOKEN }}
```

```json
// .size-limit.js
module.exports = [
  {
    path: 'dist/index.js',
    limit: '50 KB',
    gzip: true,
  },
  {
    path: 'dist/index.css',
    limit: '20 KB',
    gzip: true,
  },
];
```

---

## Security Testing in CI

### Dependency Vulnerability Scanning

```yaml
- uses: actions/dependency-review-action@v4
  if: github.event_name == 'pull_request'
  with:
    fail-on-severity: high
    deny-licenses: GPL-3.0, AGPL-3.0

# npm audit
- run: npm audit --audit-level=high

# Snyk
- uses: snyk/actions/node@master
  env:
    SNYK_TOKEN: ${{ secrets.SNYK_TOKEN }}
```

### SAST (Static Application Security Testing)

```yaml
# CodeQL
- uses: github/codeql-action/init@v3
  with:
    languages: javascript
- uses: github/codeql-action/autobuild@v3
- uses: github/codeql-action/analyze@v3

# Semgrep
- uses: returntocorp/semgrep-action@v1
  with:
    config: >-
      p/security-audit
      p/secrets
      p/owasp-top-ten
```

### Secret Detection

```yaml
# Gitleaks — detect hardcoded secrets
- uses: gitleaks/gitleaks-action@v2
  env:
    GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

# TruffleHog
- uses: trufflesecurity/trufflehog@main
  with:
    extra_args: --only-verified
```

### Container Image Scanning

```yaml
# Trivy
- uses: aquasecurity/trivy-action@master
  with:
    image-ref: 'myapp:${{ github.sha }}'
    format: 'sarif'
    output: 'trivy-results.sarif'
    severity: 'CRITICAL,HIGH'

- uses: github/codeql-action/upload-sarif@v3
  with:
    sarif_file: 'trivy-results.sarif'
```

---

## Test Reporting

### JUnit XML Reports

Most CI systems understand JUnit XML format for test reporting.

```yaml
# GitHub Actions: Display test results
- uses: dorny/test-reporter@v1
  if: always()
  with:
    name: Test Results
    path: '**/junit.xml'
    reporter: jest-junit

# GitLab CI: Built-in JUnit support
test:
  script:
    - npm test -- --ci --reporters=jest-junit
  artifacts:
    reports:
      junit: junit.xml
```

### PR Comments with Test Results

```yaml
- uses: marocchino/sticky-pull-request-comment@v2
  if: always()
  with:
    header: test-results
    message: |
      ## Test Results

      | Metric | Value |
      |--------|-------|
      | Total | ${{ steps.tests.outputs.total }} |
      | Passed | ${{ steps.tests.outputs.passed }} |
      | Failed | ${{ steps.tests.outputs.failed }} |
      | Skipped | ${{ steps.tests.outputs.skipped }} |
      | Coverage | ${{ steps.tests.outputs.coverage }}% |
      | Duration | ${{ steps.tests.outputs.duration }}s |
```

---

## CI Test Optimization Tips

### 1. Cache Dependencies

```yaml
# Use built-in cache in setup actions
- uses: actions/setup-node@v4
  with:
    cache: 'npm'
```

### 2. Run Tests in Parallel

```yaml
strategy:
  matrix:
    shard: [1, 2, 3, 4]
steps:
  - run: npx jest --shard=${{ matrix.shard }}/4
```

### 3. Only Run Affected Tests

```yaml
# With Nx
- run: npx nx affected --target=test --base=origin/main

# With Turborepo
- run: npx turbo run test --filter='...[origin/main]'

# With jest --changedSince
- run: npx jest --changedSince=origin/main
```

### 4. Split Expensive Tests

```yaml
# Run unit tests on every push, E2E only on main
jobs:
  unit:
    runs-on: ubuntu-latest
    steps:
      - run: npm run test:unit

  e2e:
    if: github.ref == 'refs/heads/main' || contains(github.event.pull_request.labels.*.name, 'run-e2e')
    runs-on: ubuntu-latest
    steps:
      - run: npm run test:e2e
```

### 5. Use Test Duration Data

```bash
# Sort tests by duration to find slow tests
npx jest --verbose 2>&1 | sort -t'(' -k2 -rn | head -20

# Use duration-based splitting for parallel runs
npx jest --shard=1/4 --shard-split=duration
```

### 6. Pre-Build Test Dependencies

```yaml
# Build once, test many
jobs:
  build:
    steps:
      - run: npm run build
      - uses: actions/upload-artifact@v4
        with:
          name: build
          path: dist/

  test-unit:
    needs: build
    steps:
      - uses: actions/download-artifact@v4
        with:
          name: build
      - run: npm run test:unit

  test-e2e:
    needs: build
    steps:
      - uses: actions/download-artifact@v4
        with:
          name: build
      - run: npm run test:e2e
```
