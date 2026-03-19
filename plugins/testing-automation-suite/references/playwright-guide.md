# Playwright Complete Reference Guide

Comprehensive reference for Playwright test automation — selectors, actions, assertions, configuration, fixtures, and advanced patterns.

## Installation & Setup

```bash
# Install Playwright
npm init playwright@latest

# Install specific browsers
npx playwright install chromium
npx playwright install firefox
npx playwright install webkit
npx playwright install --with-deps  # Include system dependencies

# Update Playwright
npm install -D @playwright/test@latest
npx playwright install --with-deps
```

## Selectors & Locators

### Built-in Locators (Recommended)

```typescript
// Role-based (preferred — matches ARIA roles)
page.getByRole('button', { name: 'Submit' });
page.getByRole('heading', { name: 'Welcome', level: 1 });
page.getByRole('link', { name: 'Sign up' });
page.getByRole('textbox', { name: 'Email' });
page.getByRole('checkbox', { name: 'Remember me' });
page.getByRole('combobox', { name: 'Country' });
page.getByRole('listitem');
page.getByRole('navigation');
page.getByRole('dialog');
page.getByRole('tab', { name: 'Settings' });
page.getByRole('tabpanel');
page.getByRole('alert');
page.getByRole('row', { name: /John/ });
page.getByRole('cell', { name: 'Active' });
page.getByRole('menuitem', { name: 'Delete' });

// Label-based (for form elements)
page.getByLabel('Email address');
page.getByLabel('Password');
page.getByLabel(/^First name/i);

// Placeholder-based
page.getByPlaceholder('Search...');
page.getByPlaceholder('Enter your email');

// Text-based
page.getByText('Welcome back');
page.getByText('Submit', { exact: true });
page.getByText(/sign (up|in)/i);

// Alt text (for images)
page.getByAltText('Company logo');
page.getByAltText(/profile/i);

// Title attribute
page.getByTitle('Close dialog');

// Test ID (fallback)
page.getByTestId('submit-button');
page.getByTestId('user-profile');
```

### CSS and XPath Selectors

```typescript
// CSS selectors (use as last resort)
page.locator('button.primary');
page.locator('[data-status="active"]');
page.locator('form input[type="email"]');
page.locator('#submit-btn');
page.locator('.sidebar >> .nav-item');

// XPath (avoid if possible)
page.locator('xpath=//button[contains(text(), "Submit")]');

// CSS pseudo-classes
page.locator(':has-text("Submit")');
page.locator('button:visible');
page.locator('input:enabled');
```

### Locator Chaining & Filtering

```typescript
// Chain locators
page.getByRole('list').getByRole('listitem');
page.getByTestId('user-table').getByRole('row');

// Filter locators
page.getByRole('listitem').filter({ hasText: 'Active' });
page.getByRole('listitem').filter({ has: page.getByRole('button') });
page.getByRole('row').filter({ hasText: 'John' }).getByRole('button', { name: 'Edit' });
page.getByRole('listitem').filter({ hasNot: page.getByRole('button', { name: 'Delete' }) });

// Nth element
page.getByRole('listitem').nth(0);      // First item
page.getByRole('listitem').nth(-1);     // Last item
page.getByRole('listitem').first();     // First
page.getByRole('listitem').last();      // Last

// Count
const count = await page.getByRole('listitem').count();
```

## Actions

### Click Actions

```typescript
// Basic click
await page.getByRole('button', { name: 'Submit' }).click();

// Double click
await page.getByRole('button').dblclick();

// Right click
await page.getByRole('button').click({ button: 'right' });

// Click with modifier keys
await page.getByRole('link').click({ modifiers: ['Control'] }); // New tab
await page.getByRole('link').click({ modifiers: ['Shift'] });

// Click at specific position
await page.getByTestId('canvas').click({ position: { x: 100, y: 200 } });

// Force click (bypass actionability checks)
await page.getByRole('button').click({ force: true });

// Click and hold
await page.getByRole('button').click({ delay: 1000 });

// Hover
await page.getByRole('button').hover();
```

### Input Actions

```typescript
// Fill input (clears first)
await page.getByLabel('Email').fill('test@example.com');

// Type character by character (triggers keyboard events)
await page.getByLabel('Search').pressSequentially('hello', { delay: 100 });

// Clear input
await page.getByLabel('Email').clear();

// Press key
await page.getByLabel('Email').press('Enter');
await page.keyboard.press('Escape');
await page.keyboard.press('Control+a');
await page.keyboard.press('Control+c');
await page.keyboard.press('Tab');

// Select text and replace
await page.getByLabel('Email').fill('');
await page.getByLabel('Email').fill('new@example.com');

// Checkbox
await page.getByLabel('Accept terms').check();
await page.getByLabel('Accept terms').uncheck();
await page.getByLabel('Accept terms').setChecked(true);

// Radio button
await page.getByLabel('Express shipping').check();

// Select dropdown
await page.getByLabel('Country').selectOption('US');
await page.getByLabel('Country').selectOption({ label: 'United States' });
await page.getByLabel('Colors').selectOption(['red', 'green', 'blue']); // Multi-select
```

### File Upload

```typescript
// Single file
await page.getByLabel('Upload').setInputFiles('path/to/file.pdf');

// Multiple files
await page.getByLabel('Upload').setInputFiles(['file1.pdf', 'file2.pdf']);

// Remove files
await page.getByLabel('Upload').setInputFiles([]);

// Create file from buffer
await page.getByLabel('Upload').setInputFiles({
  name: 'test.txt',
  mimeType: 'text/plain',
  buffer: Buffer.from('Hello World'),
});
```

### Drag and Drop

```typescript
// Drag and drop
await page.getByTestId('source').dragTo(page.getByTestId('target'));

// Manual drag (for complex cases)
await page.getByTestId('source').hover();
await page.mouse.down();
await page.getByTestId('target').hover();
await page.mouse.up();
```

## Assertions

### Page Assertions

```typescript
// URL
await expect(page).toHaveURL('/dashboard');
await expect(page).toHaveURL(/\/dashboard\/?$/);

// Title
await expect(page).toHaveTitle('My App - Dashboard');
await expect(page).toHaveTitle(/dashboard/i);

// Screenshot comparison
await expect(page).toHaveScreenshot('homepage.png');
await expect(page).toHaveScreenshot('homepage.png', {
  maxDiffPixelRatio: 0.01,
  fullPage: true,
});
```

### Locator Assertions

```typescript
// Visibility
await expect(page.getByText('Hello')).toBeVisible();
await expect(page.getByText('Error')).toBeHidden();

// Enabled/Disabled
await expect(page.getByRole('button')).toBeEnabled();
await expect(page.getByRole('button')).toBeDisabled();

// Checked state
await expect(page.getByLabel('Terms')).toBeChecked();
await expect(page.getByLabel('Terms')).not.toBeChecked();

// Text content
await expect(page.getByTestId('message')).toHaveText('Hello World');
await expect(page.getByTestId('message')).toContainText('Hello');
await expect(page.getByTestId('message')).toHaveText(/hello/i);

// Value
await expect(page.getByLabel('Email')).toHaveValue('test@example.com');
await expect(page.getByLabel('Email')).toHaveValue(/example\.com/);
await expect(page.getByLabel('Email')).toBeEmpty();

// Attribute
await expect(page.getByRole('link')).toHaveAttribute('href', '/about');
await expect(page.getByRole('img')).toHaveAttribute('alt', /logo/i);

// CSS class
await expect(page.getByTestId('alert')).toHaveClass('alert alert-danger');
await expect(page.getByTestId('alert')).toHaveClass(/danger/);

// CSS property
await expect(page.getByTestId('box')).toHaveCSS('color', 'rgb(255, 0, 0)');
await expect(page.getByTestId('box')).toHaveCSS('display', 'flex');

// Count
await expect(page.getByRole('listitem')).toHaveCount(5);

// Focus
await expect(page.getByLabel('Email')).toBeFocused();

// Attached
await expect(page.getByTestId('dynamic')).toBeAttached();

// Element screenshot
await expect(page.getByTestId('card')).toHaveScreenshot('card.png');

// Custom timeout
await expect(page.getByText('Loaded')).toBeVisible({ timeout: 15_000 });

// Negation
await expect(page.getByText('Error')).not.toBeVisible();
```

### Soft Assertions

```typescript
// Soft assertions don't stop the test on failure
await expect.soft(page.getByTestId('status')).toHaveText('Active');
await expect.soft(page.getByTestId('count')).toHaveText('5');
await expect.soft(page.getByTestId('date')).toContainText('2024');
// Test continues even if some assertions fail
```

## Navigation

```typescript
// Navigate to URL
await page.goto('https://example.com');
await page.goto('/dashboard');
await page.goto('/login', { waitUntil: 'networkidle' });

// Wait options
await page.goto('/page', { waitUntil: 'domcontentloaded' });
await page.goto('/page', { waitUntil: 'load' });
await page.goto('/page', { waitUntil: 'networkidle' });
await page.goto('/page', { waitUntil: 'commit' });
await page.goto('/page', { timeout: 30_000 });

// Navigation events
await page.goBack();
await page.goForward();
await page.reload();

// Wait for navigation
await Promise.all([
  page.waitForURL('/dashboard'),
  page.getByRole('button', { name: 'Login' }).click(),
]);

// Wait for specific response
const responsePromise = page.waitForResponse('**/api/users');
await page.getByRole('button').click();
const response = await responsePromise;
```

## Waiting

```typescript
// Wait for element
await page.getByText('Loaded').waitFor();
await page.getByText('Loaded').waitFor({ state: 'visible' });
await page.getByText('Spinner').waitFor({ state: 'hidden' });
await page.getByText('Removed').waitFor({ state: 'detached' });

// Wait for URL
await page.waitForURL('/dashboard');
await page.waitForURL('**/dashboard/**');

// Wait for load state
await page.waitForLoadState('networkidle');
await page.waitForLoadState('domcontentloaded');

// Wait for response
const response = await page.waitForResponse(
  (resp) => resp.url().includes('/api/data') && resp.status() === 200
);

// Wait for request
const request = await page.waitForRequest('**/api/submit');

// Wait for event
await page.waitForEvent('dialog');
await page.waitForEvent('download');
await page.waitForEvent('popup');

// Wait for function
await page.waitForFunction(() => document.querySelector('.loaded') !== null);
await page.waitForFunction(
  (selector) => document.querySelectorAll(selector).length > 5,
  '.item'
);

// Explicit timeout (avoid if possible)
await page.waitForTimeout(1000);
```

## Network Interception

### Route Handling

```typescript
// Mock API response
await page.route('**/api/users', (route) => {
  route.fulfill({
    status: 200,
    contentType: 'application/json',
    body: JSON.stringify([{ id: 1, name: 'John' }]),
  });
});

// Modify response
await page.route('**/api/users', async (route) => {
  const response = await route.fetch();
  const json = await response.json();
  json.push({ id: 999, name: 'Injected User' });
  route.fulfill({ response, body: JSON.stringify(json) });
});

// Abort request
await page.route('**/analytics/**', (route) => route.abort());
await page.route('**/*.{png,jpg,jpeg}', (route) => route.abort()); // Block images

// Simulate network error
await page.route('**/api/data', (route) => route.abort('connectionfailed'));

// Delay response
await page.route('**/api/data', async (route) => {
  await new Promise((r) => setTimeout(r, 3000));
  route.fulfill({ status: 200, body: '{}' });
});

// Conditional routing
await page.route('**/api/**', (route) => {
  if (route.request().method() === 'POST') {
    route.fulfill({ status: 201 });
  } else {
    route.continue();
  }
});

// Remove route
const handler = (route) => route.fulfill({ status: 200 });
await page.route('**/api/**', handler);
// Later:
await page.unroute('**/api/**', handler);

// Route at context level (applies to all pages)
await context.route('**/api/**', handler);
```

### HAR Recording and Playback

```typescript
// Record network to HAR file
await page.routeFromHAR('tests/e2e/fixtures/api.har', {
  url: '**/api/**',
  update: true, // Record mode
});

// Playback from HAR file
await page.routeFromHAR('tests/e2e/fixtures/api.har', {
  url: '**/api/**',
  update: false, // Playback mode
  notFound: 'fallback', // Fall through to network for unmatched requests
});
```

## Fixtures

### Built-in Fixtures

```typescript
import { test, expect } from '@playwright/test';

test('basic fixtures', async ({
  page,          // Isolated page per test
  context,       // Browser context (cookies, permissions)
  browser,       // Browser instance (shared across tests in same worker)
  browserName,   // 'chromium' | 'firefox' | 'webkit'
  request,       // API request context (no browser needed)
}) => {
  // Use fixtures directly
});
```

### Custom Fixtures

```typescript
// tests/e2e/fixtures.ts
import { test as base, expect } from '@playwright/test';

// Define fixture types
type MyFixtures = {
  todoPage: TodoPage;
  apiClient: APIClient;
  testUser: { email: string; password: string };
};

// Extend base test
export const test = base.extend<MyFixtures>({
  // Fixture with setup and teardown
  testUser: async ({ request }, use) => {
    // Setup: create user
    const response = await request.post('/api/test/users', {
      data: {
        email: `test-${Date.now()}@example.com`,
        password: 'SecurePass123!',
      },
    });
    const user = await response.json();

    // Provide to test
    await use({ email: user.email, password: 'SecurePass123!' });

    // Teardown: delete user
    await request.delete(`/api/test/users/${user.id}`);
  },

  // Page object fixture
  todoPage: async ({ page }, use) => {
    const todoPage = new TodoPage(page);
    await todoPage.goto();
    await use(todoPage);
  },

  // API client fixture
  apiClient: async ({ request }, use) => {
    await use(new APIClient(request));
  },
});

export { expect };
```

### Worker-Scoped Fixtures

```typescript
// Shared across all tests in a worker (for expensive setup)
type WorkerFixtures = {
  sharedDb: Database;
  sharedServer: TestServer;
};

export const test = base.extend<{}, WorkerFixtures>({
  sharedDb: [async ({}, use) => {
    const db = await Database.connect();
    await use(db);
    await db.disconnect();
  }, { scope: 'worker' }],

  sharedServer: [async ({}, use) => {
    const server = await TestServer.start();
    await use(server);
    await server.stop();
  }, { scope: 'worker' }],
});
```

## Dialogs

```typescript
// Handle alert
page.on('dialog', (dialog) => dialog.accept());

// Handle confirm
page.on('dialog', (dialog) => {
  expect(dialog.type()).toBe('confirm');
  expect(dialog.message()).toBe('Are you sure?');
  dialog.accept();
});

// Handle prompt
page.on('dialog', (dialog) => {
  expect(dialog.type()).toBe('prompt');
  dialog.accept('My input');
});

// Dismiss dialog
page.on('dialog', (dialog) => dialog.dismiss());

// One-time handler
page.once('dialog', (dialog) => dialog.accept());
```

## Downloads

```typescript
// Handle download
const downloadPromise = page.waitForEvent('download');
await page.getByRole('link', { name: 'Download' }).click();
const download = await downloadPromise;

// Save to file
await download.saveAs('/tmp/download.pdf');

// Get filename
const filename = download.suggestedFilename();

// Read content
const stream = await download.createReadStream();
```

## Multi-Page and Popup Handling

```typescript
// Handle popup (new tab/window)
const popupPromise = page.waitForEvent('popup');
await page.getByRole('link', { name: 'Open in new tab' }).click();
const popup = await popupPromise;
await popup.waitForLoadState();
expect(await popup.title()).toBe('New Page');

// Handle new page in context
const pagePromise = context.waitForEvent('page');
await page.getByRole('link').click();
const newPage = await pagePromise;
await newPage.waitForLoadState();
```

## iframe Handling

```typescript
// Access iframe by name
const frame = page.frame('iframe-name');

// Access iframe by URL
const frame = page.frame({ url: /embedded/ });

// Use frameLocator (recommended)
const frame = page.frameLocator('#my-iframe');
await frame.getByRole('button', { name: 'Submit' }).click();

// Nested iframes
const nestedFrame = page
  .frameLocator('#outer-frame')
  .frameLocator('#inner-frame');
await nestedFrame.getByText('Hello').click();
```

## API Testing (Without Browser)

```typescript
import { test, expect } from '@playwright/test';

test.describe('API Tests', () => {
  test('GET /api/users', async ({ request }) => {
    const response = await request.get('/api/users');
    expect(response.ok()).toBeTruthy();
    expect(response.status()).toBe(200);

    const data = await response.json();
    expect(data).toHaveProperty('data');
    expect(data.data.length).toBeGreaterThan(0);
  });

  test('POST /api/users', async ({ request }) => {
    const response = await request.post('/api/users', {
      data: {
        email: 'new@example.com',
        name: 'New User',
        password: 'secure123',
      },
    });
    expect(response.status()).toBe(201);

    const user = await response.json();
    expect(user.email).toBe('new@example.com');
  });

  test('PUT /api/users/:id', async ({ request }) => {
    const response = await request.put('/api/users/1', {
      data: { name: 'Updated Name' },
      headers: { Authorization: 'Bearer test-token' },
    });
    expect(response.ok()).toBeTruthy();
  });

  test('DELETE /api/users/:id', async ({ request }) => {
    const response = await request.delete('/api/users/1');
    expect(response.status()).toBe(204);
  });

  test('multipart form data', async ({ request }) => {
    const response = await request.post('/api/upload', {
      multipart: {
        file: {
          name: 'test.txt',
          mimeType: 'text/plain',
          buffer: Buffer.from('Hello'),
        },
        description: 'Test file',
      },
    });
    expect(response.ok()).toBeTruthy();
  });
});
```

## Visual Testing

```typescript
// Full page screenshot
await expect(page).toHaveScreenshot('full-page.png', {
  fullPage: true,
  maxDiffPixelRatio: 0.01,
});

// Element screenshot
await expect(page.getByTestId('card')).toHaveScreenshot('card.png');

// With masking (hide dynamic content)
await expect(page).toHaveScreenshot('page.png', {
  mask: [
    page.getByTestId('timestamp'),
    page.getByTestId('random-avatar'),
  ],
  maskColor: '#FF00FF',
});

// With clipping
await expect(page).toHaveScreenshot('header.png', {
  clip: { x: 0, y: 0, width: 1280, height: 100 },
});

// Threshold configuration
await expect(page).toHaveScreenshot('page.png', {
  threshold: 0.2,        // Per-pixel threshold (0-1)
  maxDiffPixels: 100,    // Max different pixels
  maxDiffPixelRatio: 0.01, // Max ratio of different pixels
});

// Update snapshots: npx playwright test --update-snapshots
```

## Accessibility Testing

```typescript
import AxeBuilder from '@axe-core/playwright';

test('accessibility scan', async ({ page }) => {
  await page.goto('/');

  const results = await new AxeBuilder({ page })
    .withTags(['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa'])
    .exclude('#third-party-widget')
    .analyze();

  expect(results.violations).toEqual([]);
});

test('specific WCAG rules', async ({ page }) => {
  await page.goto('/form');

  const results = await new AxeBuilder({ page })
    .withRules(['color-contrast', 'label', 'image-alt'])
    .analyze();

  expect(results.violations).toEqual([]);
});
```

## Test Configuration Patterns

### Project-Based Configuration

```typescript
// playwright.config.ts
export default defineConfig({
  projects: [
    // Logged out tests
    { name: 'public', testMatch: '**/*.public.spec.ts' },

    // Auth setup
    { name: 'setup', testMatch: /.*\.setup\.ts/ },

    // User tests
    {
      name: 'user',
      testMatch: '**/*.user.spec.ts',
      use: { storageState: '.auth/user.json' },
      dependencies: ['setup'],
    },

    // Admin tests
    {
      name: 'admin',
      testMatch: '**/*.admin.spec.ts',
      use: { storageState: '.auth/admin.json' },
      dependencies: ['setup'],
    },

    // API tests (no browser)
    {
      name: 'api',
      testMatch: '**/*.api.spec.ts',
      use: { baseURL: 'http://localhost:3000' },
    },

    // Visual regression
    {
      name: 'visual',
      testMatch: '**/*.visual.spec.ts',
      use: { ...devices['Desktop Chrome'] },
    },
  ],
});
```

### Tag-Based Test Selection

```typescript
// Tag tests
test('login flow @smoke @auth', async ({ page }) => { ... });
test('user profile @auth', async ({ page }) => { ... });
test('payment @smoke @billing', async ({ page }) => { ... });

// Run by tag:
// npx playwright test --grep @smoke
// npx playwright test --grep-invert @slow
// npx playwright test --grep "@smoke|@auth"
```

## CLI Commands Reference

```bash
# Run all tests
npx playwright test

# Run specific file
npx playwright test auth.spec.ts

# Run tests matching title
npx playwright test -g "login"

# Run specific project
npx playwright test --project=chromium

# Debug mode
npx playwright test --debug

# Headed mode
npx playwright test --headed

# UI mode (interactive)
npx playwright test --ui

# Trace viewer
npx playwright show-trace trace.zip

# Code generation
npx playwright codegen http://localhost:3000

# Update snapshots
npx playwright test --update-snapshots

# Show report
npx playwright show-report

# List tests without running
npx playwright test --list

# Retry failed tests
npx playwright test --retries=3

# Run in parallel
npx playwright test --workers=4

# Shard tests
npx playwright test --shard=1/3
```
