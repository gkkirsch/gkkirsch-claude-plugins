---
name: playwright-setup
description: >
  Playwright setup and configuration — installation, playwright.config.ts,
  project structure, fixtures, authentication, global setup, and CI/CD.
  Triggers: "playwright setup", "playwright config", "playwright install",
  "playwright fixture", "playwright auth", "playwright ci", "playwright github actions",
  "e2e setup", "end to end testing setup".
  NOT for: writing test assertions or patterns (use playwright-patterns).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Playwright Setup & Configuration

## Installation

```bash
# Install Playwright
npm init playwright@latest
# Installs: @playwright/test, playwright.config.ts, tests/ folder, example test

# Install browsers
npx playwright install
npx playwright install chromium    # Just Chromium
npx playwright install --with-deps # With system dependencies (CI)
```

## Configuration

```typescript
// playwright.config.ts
import { defineConfig, devices } from "@playwright/test";

export default defineConfig({
  testDir: "./tests",
  testMatch: "**/*.spec.ts",

  // Parallel execution
  fullyParallel: true,
  workers: process.env.CI ? 1 : undefined,

  // Retries
  retries: process.env.CI ? 2 : 0,

  // Reporter
  reporter: process.env.CI
    ? [["html", { open: "never" }], ["github"]]
    : [["html", { open: "on-failure" }]],

  // Timeouts
  timeout: 30_000,           // Per test
  expect: {
    timeout: 5_000,          // Per assertion
  },

  // Shared settings
  use: {
    baseURL: process.env.BASE_URL || "http://localhost:3000",
    trace: "on-first-retry",
    screenshot: "only-on-failure",
    video: "retain-on-failure",

    // Browser context
    viewport: { width: 1280, height: 720 },
    locale: "en-US",
    timezoneId: "America/New_York",
  },

  // Browser projects
  projects: [
    // Setup project — runs first, saves auth state
    {
      name: "setup",
      testMatch: /.*\.setup\.ts/,
    },

    // Desktop browsers
    {
      name: "chromium",
      use: {
        ...devices["Desktop Chrome"],
        storageState: "tests/.auth/user.json",
      },
      dependencies: ["setup"],
    },
    {
      name: "firefox",
      use: {
        ...devices["Desktop Firefox"],
        storageState: "tests/.auth/user.json",
      },
      dependencies: ["setup"],
    },
    {
      name: "webkit",
      use: {
        ...devices["Desktop Safari"],
        storageState: "tests/.auth/user.json",
      },
      dependencies: ["setup"],
    },

    // Mobile browsers
    {
      name: "mobile-chrome",
      use: {
        ...devices["Pixel 7"],
        storageState: "tests/.auth/user.json",
      },
      dependencies: ["setup"],
    },
    {
      name: "mobile-safari",
      use: {
        ...devices["iPhone 14"],
        storageState: "tests/.auth/user.json",
      },
      dependencies: ["setup"],
    },
  ],

  // Dev server
  webServer: {
    command: "npm run dev",
    url: "http://localhost:3000",
    reuseExistingServer: !process.env.CI,
    timeout: 120_000,
  },
});
```

## Project Structure

```
tests/
  .auth/
    user.json              # Auth state (gitignored)
    admin.json
  fixtures/
    base.ts                # Base fixtures
    auth.ts                # Auth fixtures
    data.ts                # Test data fixtures
  pages/
    login.page.ts          # Page object
    dashboard.page.ts
    settings.page.ts
  helpers/
    api.ts                 # API helpers
    data-factory.ts        # Test data factories
  auth.setup.ts            # Global auth setup
  dashboard.spec.ts        # Test files
  settings.spec.ts
  checkout.spec.ts
playwright.config.ts
```

## Authentication Setup

```typescript
// tests/auth.setup.ts
import { test as setup, expect } from "@playwright/test";

const authFile = "tests/.auth/user.json";

setup("authenticate as user", async ({ page }) => {
  // Navigate to login
  await page.goto("/auth/login");

  // Fill credentials
  await page.getByLabel("Email").fill("test@example.com");
  await page.getByLabel("Password").fill("password123");
  await page.getByRole("button", { name: "Sign in" }).click();

  // Wait for auth to complete
  await page.waitForURL("/dashboard");
  await expect(page.getByText("Welcome")).toBeVisible();

  // Save auth state
  await page.context().storageState({ path: authFile });
});

// tests/admin.setup.ts — for admin auth
const adminFile = "tests/.auth/admin.json";

setup("authenticate as admin", async ({ page }) => {
  await page.goto("/auth/login");
  await page.getByLabel("Email").fill("admin@example.com");
  await page.getByLabel("Password").fill("admin123");
  await page.getByRole("button", { name: "Sign in" }).click();
  await page.waitForURL("/admin");
  await page.context().storageState({ path: adminFile });
});
```

```gitignore
# .gitignore
tests/.auth/
test-results/
playwright-report/
blob-report/
```

## Custom Fixtures

```typescript
// tests/fixtures/base.ts
import { test as base, expect } from "@playwright/test";
import { DashboardPage } from "../pages/dashboard.page";
import { SettingsPage } from "../pages/settings.page";

type Fixtures = {
  dashboardPage: DashboardPage;
  settingsPage: SettingsPage;
};

export const test = base.extend<Fixtures>({
  dashboardPage: async ({ page }, use) => {
    const dashboardPage = new DashboardPage(page);
    await dashboardPage.goto();
    await use(dashboardPage);
  },
  settingsPage: async ({ page }, use) => {
    const settingsPage = new SettingsPage(page);
    await settingsPage.goto();
    await use(settingsPage);
  },
});

export { expect };
```

```typescript
// tests/fixtures/data.ts — Data fixture with cleanup
import { test as base } from "./base";

type DataFixtures = {
  testPost: { id: string; title: string };
};

export const test = base.extend<DataFixtures>({
  testPost: async ({ request }, use) => {
    // Setup: create test data via API
    const response = await request.post("/api/posts", {
      data: {
        title: `Test Post ${Date.now()}`,
        body: "E2E test content",
        status: "published",
      },
    });
    const post = await response.json();

    // Provide to test
    await use(post.data);

    // Teardown: clean up
    await request.delete(`/api/posts/${post.data.id}`);
  },
});

export { expect } from "./base";
```

## Page Object Model

```typescript
// tests/pages/dashboard.page.ts
import type { Page, Locator } from "@playwright/test";

export class DashboardPage {
  readonly page: Page;
  readonly heading: Locator;
  readonly createButton: Locator;
  readonly postList: Locator;
  readonly searchInput: Locator;

  constructor(page: Page) {
    this.page = page;
    this.heading = page.getByRole("heading", { name: "Dashboard" });
    this.createButton = page.getByRole("button", { name: "Create" });
    this.postList = page.getByTestId("post-list");
    this.searchInput = page.getByPlaceholder("Search posts...");
  }

  async goto() {
    await this.page.goto("/dashboard");
    await this.heading.waitFor();
  }

  async createPost(title: string, body: string) {
    await this.createButton.click();
    await this.page.getByLabel("Title").fill(title);
    await this.page.getByLabel("Body").fill(body);
    await this.page.getByRole("button", { name: "Publish" }).click();
    await this.page.waitForURL(/\/posts\/.+/);
  }

  async searchPosts(query: string) {
    await this.searchInput.fill(query);
    await this.searchInput.press("Enter");
    // Wait for results to update
    await this.page.waitForResponse(resp =>
      resp.url().includes("/api/posts") && resp.status() === 200
    );
  }

  async getPostCount(): Promise<number> {
    return this.postList.getByRole("article").count();
  }
}
```

## CLI Commands

```bash
# Run all tests
npx playwright test

# Run specific file
npx playwright test tests/dashboard.spec.ts

# Run tests with specific title
npx playwright test -g "should create a post"

# Run in headed mode (see the browser)
npx playwright test --headed

# Run in UI mode (interactive)
npx playwright test --ui

# Run in debug mode (step through)
npx playwright test --debug

# Run specific project/browser
npx playwright test --project=chromium
npx playwright test --project=mobile-chrome

# Generate tests (codegen)
npx playwright codegen http://localhost:3000

# Show report
npx playwright show-report

# Update snapshots
npx playwright test --update-snapshots
```

## GitHub Actions CI

```yaml
# .github/workflows/e2e.yml
name: E2E Tests

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  e2e:
    runs-on: ubuntu-latest
    timeout-minutes: 15
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-node@v4
        with:
          node-version: 20
          cache: "npm"

      - run: npm ci

      - name: Install Playwright browsers
        run: npx playwright install --with-deps

      - name: Build app
        run: npm run build

      - name: Run E2E tests
        run: npx playwright test
        env:
          BASE_URL: http://localhost:3000

      - name: Upload report
        uses: actions/upload-artifact@v4
        if: ${{ !cancelled() }}
        with:
          name: playwright-report
          path: playwright-report/
          retention-days: 7

      - name: Upload test results
        uses: actions/upload-artifact@v4
        if: failure()
        with:
          name: test-results
          path: test-results/
          retention-days: 7
```

## Gotchas

1. **`storageState` files contain secrets.** Add `tests/.auth/` to `.gitignore`. These files contain cookies and local storage data that could include auth tokens.

2. **`webServer` starts before tests, but doesn't wait for app ready.** The `url` option just checks if the URL responds with 2xx. If your app takes time to initialize (database migrations, etc.), increase `timeout` or add a health check endpoint.

3. **Tests share browser context by default.** Each test gets a fresh page but shares the context (cookies, localStorage). Use `test.describe.configure({ mode: 'serial' })` if tests must run in order, or ensure complete isolation via fixtures.

4. **`--with-deps` is required on CI.** Without it, browser binaries exist but system dependencies (fonts, libraries) are missing, causing cryptic launch failures. Always use `npx playwright install --with-deps` on fresh CI environments.

5. **Visual snapshots are OS-specific.** A screenshot taken on macOS won't match one taken on Linux (different fonts, rendering). Generate baseline snapshots on the same OS as CI (usually Linux). Use Docker for local snapshot updates.

6. **`page.waitForTimeout()` is always wrong.** It makes tests slow and flaky. Use `expect(locator).toBeVisible()`, `page.waitForURL()`, `page.waitForResponse()`, or `page.waitForLoadState()` instead. Every wait should be conditional, not time-based.
