---
name: playwright-e2e
description: >
  Playwright E2E testing patterns — page interactions, selectors, assertions,
  authentication, fixtures, network mocking, and CI configuration.
  Triggers: "playwright", "e2e test", "end to end", "browser test",
  "page test", "playwright selector", "playwright fixture".
  NOT for: unit tests (use vitest-patterns), strategy (use test-architect agent).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Playwright E2E Patterns

## Configuration

```typescript
// playwright.config.ts
import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
  testDir: './tests/e2e',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,            // Fail if test.only in CI
  retries: process.env.CI ? 2 : 0,         // Retry flaky tests in CI
  workers: process.env.CI ? 1 : undefined,  // Single worker in CI
  reporter: [
    ['html', { open: 'never' }],
    ['list'],
  ],
  use: {
    baseURL: 'http://localhost:5173',
    trace: 'on-first-retry',               // Record trace for retried tests
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
  },
  projects: [
    { name: 'chromium', use: { ...devices['Desktop Chrome'] } },
    { name: 'firefox', use: { ...devices['Desktop Firefox'] } },
    { name: 'webkit', use: { ...devices['Desktop Safari'] } },
    { name: 'mobile', use: { ...devices['iPhone 14'] } },
  ],
  webServer: {
    command: 'npm run dev',
    url: 'http://localhost:5173',
    reuseExistingServer: !process.env.CI,
  },
});
```

## Page Object Pattern

```typescript
// tests/e2e/pages/login.page.ts
import { Page, Locator } from '@playwright/test';

export class LoginPage {
  readonly page: Page;
  readonly emailInput: Locator;
  readonly passwordInput: Locator;
  readonly submitButton: Locator;
  readonly errorMessage: Locator;

  constructor(page: Page) {
    this.page = page;
    this.emailInput = page.getByLabel('Email');
    this.passwordInput = page.getByLabel('Password');
    this.submitButton = page.getByRole('button', { name: 'Sign in' });
    this.errorMessage = page.getByRole('alert');
  }

  async goto() {
    await this.page.goto('/login');
  }

  async login(email: string, password: string) {
    await this.emailInput.fill(email);
    await this.passwordInput.fill(password);
    await this.submitButton.click();
  }

  async expectError(message: string) {
    await expect(this.errorMessage).toContainText(message);
  }
}
```

## Auth Fixture (Reusable Login State)

```typescript
// tests/e2e/fixtures/auth.ts
import { test as base, expect } from '@playwright/test';
import { LoginPage } from '../pages/login.page';

type AuthFixtures = {
  authenticatedPage: Page;
};

export const test = base.extend<AuthFixtures>({
  authenticatedPage: async ({ page }, use) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.login('test@example.com', 'password123');
    await page.waitForURL('/dashboard');
    await use(page);
  },
});

export { expect };
```

```typescript
// Alternative: storageState (faster — saves cookies/localStorage)
// 1. Create auth setup
// tests/e2e/auth.setup.ts
import { test as setup, expect } from '@playwright/test';

setup('authenticate', async ({ page }) => {
  await page.goto('/login');
  await page.getByLabel('Email').fill('test@example.com');
  await page.getByLabel('Password').fill('password123');
  await page.getByRole('button', { name: 'Sign in' }).click();
  await page.waitForURL('/dashboard');

  // Save auth state
  await page.context().storageState({ path: 'tests/e2e/.auth/user.json' });
});

// 2. Use in playwright.config.ts
// projects: [
//   { name: 'setup', testMatch: /.*\.setup\.ts/ },
//   { name: 'chromium', use: { storageState: 'tests/e2e/.auth/user.json' }, dependencies: ['setup'] },
// ]
```

## Common Test Patterns

### Navigation and Assertions

```typescript
import { test, expect } from '@playwright/test';

test('should navigate to dashboard after login', async ({ page }) => {
  await page.goto('/login');
  await page.getByLabel('Email').fill('user@test.com');
  await page.getByLabel('Password').fill('password123');
  await page.getByRole('button', { name: 'Sign in' }).click();

  // Wait for navigation
  await expect(page).toHaveURL('/dashboard');
  await expect(page).toHaveTitle(/Dashboard/);
  await expect(page.getByRole('heading', { name: 'Welcome' })).toBeVisible();
});
```

### Form Interactions

```typescript
test('should create a new post', async ({ page }) => {
  await page.goto('/posts/new');

  // Text input
  await page.getByLabel('Title').fill('My New Post');

  // Rich text / textarea
  await page.getByLabel('Content').fill('This is the post content');

  // Select dropdown
  await page.getByLabel('Category').selectOption('technology');

  // Checkbox
  await page.getByLabel('Published').check();

  // Radio button
  await page.getByLabel('Public').check();

  // File upload
  await page.getByLabel('Cover Image').setInputFiles('tests/fixtures/test-image.png');

  // Submit
  await page.getByRole('button', { name: 'Create Post' }).click();

  // Verify redirect and success
  await expect(page).toHaveURL(/\/posts\/.+/);
  await expect(page.getByText('Post created successfully')).toBeVisible();
});
```

### Table and List Assertions

```typescript
test('should display user list', async ({ page }) => {
  await page.goto('/admin/users');

  // Wait for data to load
  await expect(page.getByRole('table')).toBeVisible();

  // Count rows
  const rows = page.getByRole('row');
  await expect(rows).toHaveCount(11); // Header + 10 data rows

  // Check specific cell content
  await expect(rows.nth(1)).toContainText('admin@test.com');

  // Check sorted order
  const emails = await page.getByTestId('user-email').allTextContents();
  expect(emails).toEqual([...emails].sort());
});
```

### Modal and Dialog

```typescript
test('should confirm deletion', async ({ page }) => {
  await page.goto('/posts');

  // Click delete button
  await page.getByTestId('delete-post-1').click();

  // Wait for confirmation dialog
  const dialog = page.getByRole('dialog');
  await expect(dialog).toBeVisible();
  await expect(dialog).toContainText('Are you sure?');

  // Confirm
  await dialog.getByRole('button', { name: 'Delete' }).click();

  // Verify dialog closed and item removed
  await expect(dialog).not.toBeVisible();
  await expect(page.getByTestId('post-1')).not.toBeVisible();
});
```

### Network Mocking

```typescript
test('should handle API errors gracefully', async ({ page }) => {
  // Mock API to return error
  await page.route('**/api/posts', (route) => {
    route.fulfill({
      status: 500,
      contentType: 'application/json',
      body: JSON.stringify({
        success: false,
        error: { code: 'INTERNAL_ERROR', message: 'Server error' },
      }),
    });
  });

  await page.goto('/posts');
  await expect(page.getByText('Failed to load posts')).toBeVisible();
  await expect(page.getByRole('button', { name: 'Retry' })).toBeVisible();
});

test('should show loading state', async ({ page }) => {
  // Delay API response
  await page.route('**/api/posts', async (route) => {
    await new Promise((r) => setTimeout(r, 2000));
    await route.continue();
  });

  await page.goto('/posts');
  await expect(page.getByTestId('loading-skeleton')).toBeVisible();
  await expect(page.getByTestId('loading-skeleton')).not.toBeVisible({ timeout: 5000 });
});
```

### Waiting Patterns

```typescript
// Wait for element
await page.getByRole('button').waitFor();
await page.getByRole('button').waitFor({ state: 'visible' });
await page.getByRole('button').waitFor({ state: 'hidden' });

// Wait for URL
await page.waitForURL('/dashboard');
await page.waitForURL(/\/posts\/\w+/);

// Wait for network
await page.waitForResponse('**/api/posts');
const responsePromise = page.waitForResponse('**/api/posts');
await page.getByRole('button').click();
const response = await responsePromise;
expect(response.status()).toBe(200);

// Wait for load state
await page.waitForLoadState('networkidle');
await page.waitForLoadState('domcontentloaded');

// Custom wait
await expect(async () => {
  const count = await page.getByTestId('item').count();
  expect(count).toBeGreaterThan(0);
}).toPass({ timeout: 5000 });
```

## Selector Priority

```
1. getByRole()        — accessible role (best)
2. getByLabel()       — form elements by label
3. getByPlaceholder() — input placeholder
4. getByText()        — visible text content
5. getByTestId()      — data-testid attribute (fallback)
6. locator()          — CSS/XPath (last resort)
```

```typescript
// Good selectors
page.getByRole('button', { name: 'Submit' });
page.getByLabel('Email');
page.getByText('Welcome back');
page.getByTestId('post-card');

// Avoid
page.locator('.btn-primary');         // Fragile CSS class
page.locator('#submit-btn');          // Fragile ID
page.locator('div > span:nth-child(2)'); // Position-dependent
```

## Visual Comparison

```typescript
test('should match visual snapshot', async ({ page }) => {
  await page.goto('/');
  await expect(page).toHaveScreenshot('homepage.png', {
    maxDiffPixelRatio: 0.01,
  });
});

test('should match component snapshot', async ({ page }) => {
  await page.goto('/components/button');
  const button = page.getByRole('button', { name: 'Primary' });
  await expect(button).toHaveScreenshot('primary-button.png');
});
```

## Gotchas

1. **Don't use `page.waitForTimeout()`.** Arbitrary waits cause flaky tests. Use `waitFor`, `waitForURL`, `waitForResponse`, or `expect().toPass()` instead.

2. **`toHaveText` vs `toContainText`.** `toHaveText` checks EXACT match (with normalization). `toContainText` checks partial match. Use `toContainText` for dynamic content.

3. **Locators are lazy.** `page.getByRole('button')` doesn't query the DOM immediately. It queries when you interact with it or assert on it. This is good — it avoids stale element issues.

4. **`storageState` doesn't include sessionStorage.** It saves cookies and localStorage only. For sessionStorage-based auth, use the fixture approach.

5. **Parallel tests share nothing.** Each test gets a fresh browser context. Don't rely on test execution order or shared state between tests.

6. **CI runs are slower.** Use `workers: 1` in CI to avoid resource contention. Add retries (`retries: 2`) for flaky tests. Always generate traces for debugging CI failures.
