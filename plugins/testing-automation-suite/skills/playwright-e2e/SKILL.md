---
name: playwright-e2e
description: >
  Playwright end-to-end testing — page navigation, selectors, assertions,
  authentication, API testing, visual regression, fixtures, and CI setup.
  Triggers: "playwright", "e2e testing", "end to end test", "browser testing",
  "visual regression", "playwright fixtures", "e2e ci".
  NOT for: unit testing or component testing (use vitest-testing).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Playwright E2E Testing

## Setup

```bash
npm init playwright@latest
# or
npm install -D @playwright/test
npx playwright install
```

```typescript
// playwright.config.ts
import { defineConfig, devices } from "@playwright/test";

export default defineConfig({
  testDir: "./e2e",
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: [
    ["html", { open: "never" }],
    ["list"],
    ...(process.env.CI ? [["github"] as const] : []),
  ],
  use: {
    baseURL: process.env.BASE_URL || "http://localhost:3000",
    trace: "on-first-retry",
    screenshot: "only-on-failure",
    video: "on-first-retry",
  },
  projects: [
    { name: "chromium", use: { ...devices["Desktop Chrome"] } },
    { name: "firefox", use: { ...devices["Desktop Firefox"] } },
    { name: "mobile", use: { ...devices["iPhone 14"] } },
  ],
  webServer: {
    command: "npm run dev",
    url: "http://localhost:3000",
    reuseExistingServer: !process.env.CI,
    timeout: 30_000,
  },
});
```

## Page Navigation & Interaction

```typescript
import { test, expect } from "@playwright/test";

test("user can navigate and interact", async ({ page }) => {
  // Navigate
  await page.goto("/");
  await page.goto("/dashboard", { waitUntil: "networkidle" });

  // Click
  await page.click("button:has-text('Submit')");
  await page.getByRole("button", { name: "Submit" }).click();
  await page.getByRole("link", { name: "Dashboard" }).click();

  // Type
  await page.getByLabel("Email").fill("alice@example.com");
  await page.getByPlaceholder("Search...").fill("query");

  // Select dropdown
  await page.getByLabel("Country").selectOption("US");
  await page.getByLabel("Country").selectOption({ label: "United States" });

  // Checkbox/radio
  await page.getByLabel("Accept terms").check();
  await page.getByLabel("Premium plan").check();

  // File upload
  await page.getByLabel("Upload").setInputFiles("./fixtures/test-image.png");

  // Keyboard
  await page.keyboard.press("Enter");
  await page.keyboard.press("Control+A");
  await page.keyboard.type("Hello World");

  // Wait for navigation
  await Promise.all([
    page.waitForURL("/dashboard"),
    page.click("a[href='/dashboard']"),
  ]);
});
```

## Selectors (Priority Order)

```typescript
// 1. Role-based (preferred — tests accessibility)
page.getByRole("button", { name: "Submit" });
page.getByRole("heading", { name: "Dashboard", level: 1 });
page.getByRole("link", { name: "Settings" });
page.getByRole("textbox", { name: "Email" });
page.getByRole("checkbox", { name: "Remember me" });

// 2. Label-based (form elements)
page.getByLabel("Email address");
page.getByLabel("Password");

// 3. Placeholder
page.getByPlaceholder("Search...");

// 4. Text
page.getByText("Welcome back");
page.getByText(/welcome/i);  // Regex for case-insensitive

// 5. Test ID (last resort)
page.getByTestId("submit-button");

// Chaining and filtering
page.getByRole("listitem").filter({ hasText: "Alice" }).getByRole("button", { name: "Delete" });
page.locator("table tbody tr").nth(2).getByRole("cell").first();
```

## Assertions

```typescript
test("page content assertions", async ({ page }) => {
  await page.goto("/dashboard");

  // Visibility
  await expect(page.getByText("Welcome")).toBeVisible();
  await expect(page.getByTestId("error")).not.toBeVisible();
  await expect(page.getByTestId("error")).toBeHidden();

  // Text content
  await expect(page.getByRole("heading")).toHaveText("Dashboard");
  await expect(page.getByRole("heading")).toContainText("Dash");

  // Attributes
  await expect(page.getByRole("button")).toBeEnabled();
  await expect(page.getByRole("button")).toBeDisabled();
  await expect(page.getByLabel("Email")).toHaveValue("alice@example.com");
  await expect(page.getByRole("link")).toHaveAttribute("href", "/settings");

  // CSS
  await expect(page.getByTestId("alert")).toHaveClass(/bg-red/);
  await expect(page.getByTestId("box")).toHaveCSS("display", "flex");

  // Count
  await expect(page.getByRole("listitem")).toHaveCount(5);

  // URL
  await expect(page).toHaveURL(/\/dashboard/);
  await expect(page).toHaveTitle("Dashboard | My App");

  // Wait for element (auto-retry up to timeout)
  await expect(page.getByText("Data loaded")).toBeVisible({ timeout: 10_000 });
});
```

## Authentication Fixture

```typescript
// e2e/fixtures.ts
import { test as base, expect } from "@playwright/test";

// Save auth state to file for reuse across tests
const authFile = "e2e/.auth/user.json";

// Setup: login once, save state
base.describe("setup", () => {
  base("authenticate", async ({ page }) => {
    await page.goto("/login");
    await page.getByLabel("Email").fill("test@example.com");
    await page.getByLabel("Password").fill("password123");
    await page.getByRole("button", { name: "Sign In" }).click();
    await page.waitForURL("/dashboard");

    await page.context().storageState({ path: authFile });
  });
});

// Authenticated test fixture
export const test = base.extend<{ authenticatedPage: typeof base }>({});

// In playwright.config.ts — use storage state for authenticated tests
// projects: [
//   { name: "setup", testMatch: /.*\.setup\.ts/ },
//   {
//     name: "tests",
//     use: { storageState: "e2e/.auth/user.json" },
//     dependencies: ["setup"],
//   },
// ]
```

## API Testing

```typescript
import { test, expect } from "@playwright/test";

test.describe("API", () => {
  test("CRUD operations", async ({ request }) => {
    // POST — create
    const createRes = await request.post("/api/users", {
      data: { name: "Alice", email: "alice@example.com" },
    });
    expect(createRes.ok()).toBeTruthy();
    const user = await createRes.json();
    expect(user).toHaveProperty("id");

    // GET — read
    const getRes = await request.get(`/api/users/${user.id}`);
    expect(getRes.ok()).toBeTruthy();
    expect(await getRes.json()).toMatchObject({ name: "Alice" });

    // PUT — update
    const updateRes = await request.put(`/api/users/${user.id}`, {
      data: { name: "Alice Updated" },
    });
    expect(updateRes.ok()).toBeTruthy();

    // DELETE — cleanup
    const deleteRes = await request.delete(`/api/users/${user.id}`);
    expect(deleteRes.status()).toBe(204);
  });

  test("handles validation errors", async ({ request }) => {
    const res = await request.post("/api/users", {
      data: { name: "" },  // Missing required fields
    });
    expect(res.status()).toBe(400);
    const body = await res.json();
    expect(body.error).toBeDefined();
  });
});
```

## Mock Network Requests

```typescript
test("mocks API response", async ({ page }) => {
  // Intercept and mock
  await page.route("**/api/users", (route) => {
    route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify([
        { id: "1", name: "Mock User" },
      ]),
    });
  });

  await page.goto("/users");
  await expect(page.getByText("Mock User")).toBeVisible();
});

test("simulates network error", async ({ page }) => {
  await page.route("**/api/data", (route) => route.abort("failed"));

  await page.goto("/dashboard");
  await expect(page.getByText("Failed to load data")).toBeVisible();
});

test("delays response", async ({ page }) => {
  await page.route("**/api/data", async (route) => {
    await new Promise((r) => setTimeout(r, 3000));
    await route.continue();
  });

  await page.goto("/dashboard");
  await expect(page.getByText("Loading...")).toBeVisible();
  await expect(page.getByText("Data loaded")).toBeVisible({ timeout: 5000 });
});
```

## Visual Regression

```typescript
test("visual comparison", async ({ page }) => {
  await page.goto("/dashboard");

  // Full page screenshot comparison
  await expect(page).toHaveScreenshot("dashboard.png", {
    maxDiffPixelRatio: 0.01,  // Allow 1% pixel difference
  });

  // Element screenshot
  const card = page.getByTestId("stats-card");
  await expect(card).toHaveScreenshot("stats-card.png");

  // With masking (ignore dynamic content)
  await expect(page).toHaveScreenshot("dashboard-masked.png", {
    mask: [
      page.getByTestId("timestamp"),
      page.getByTestId("random-avatar"),
    ],
  });
});

// Update snapshots: npx playwright test --update-snapshots
```

## Page Object Model

```typescript
// e2e/pages/LoginPage.ts
import { Page, Locator } from "@playwright/test";

export class LoginPage {
  readonly page: Page;
  readonly emailInput: Locator;
  readonly passwordInput: Locator;
  readonly submitButton: Locator;
  readonly errorMessage: Locator;

  constructor(page: Page) {
    this.page = page;
    this.emailInput = page.getByLabel("Email");
    this.passwordInput = page.getByLabel("Password");
    this.submitButton = page.getByRole("button", { name: "Sign In" });
    this.errorMessage = page.getByRole("alert");
  }

  async goto() {
    await this.page.goto("/login");
  }

  async login(email: string, password: string) {
    await this.emailInput.fill(email);
    await this.passwordInput.fill(password);
    await this.submitButton.click();
  }
}

// Usage in tests
test("login flow", async ({ page }) => {
  const loginPage = new LoginPage(page);
  await loginPage.goto();
  await loginPage.login("alice@example.com", "password123");
  await expect(page).toHaveURL("/dashboard");
});
```

## CLI Commands

```bash
npx playwright test                      # Run all tests
npx playwright test --headed             # Show browser
npx playwright test --ui                 # Interactive UI mode
npx playwright test --debug              # Debug with inspector
npx playwright test e2e/login.spec.ts    # Specific file
npx playwright test -g "login"           # Grep test name
npx playwright test --project=chromium   # Specific browser
npx playwright show-report               # Open HTML report
npx playwright codegen http://localhost:3000  # Record test
npx playwright test --update-snapshots   # Update visual snapshots
```

## Gotchas

1. **Auto-waiting is built in** — Playwright automatically waits for elements to be actionable before clicking/typing. Don't add manual `waitForSelector` before `click`. It's redundant.

2. **`page.locator()` vs `page.$()`** — Always use `page.locator()` or `page.getByRole()`. The `$()` API doesn't auto-wait and returns null instead of throwing on missing elements.

3. **Network idle is fragile** — `waitUntil: "networkidle"` waits for no network requests for 500ms. Long-polling, WebSockets, or analytics requests can make this hang. Prefer `waitForURL` or `expect(...).toBeVisible()`.

4. **Test isolation** — Each test gets a fresh browser context by default. But if you're using `storageState` for auth, cookies persist. Clean up test data in `afterEach` if tests modify shared state.

5. **CI needs system dependencies** — `npx playwright install --with-deps` installs browser dependencies on CI. Without `--with-deps`, Chromium may fail on Ubuntu with missing shared libraries.

6. **Screenshot tests are OS-specific** — Font rendering differs between macOS, Linux, and Windows. Run screenshot tests on the same OS as CI, or use `maxDiffPixelRatio` to allow minor differences.
