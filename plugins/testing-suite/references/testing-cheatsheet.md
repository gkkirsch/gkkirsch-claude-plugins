# Testing Cheat Sheet

## Vitest Quick Reference

```bash
npx vitest                     # Run in watch mode
npx vitest run                 # Run once
npx vitest run --coverage      # With coverage
npx vitest run src/auth        # Filter by path
npx vitest run -t "login"      # Filter by test name
npx vitest --ui                # Open browser UI
```

### Config

```typescript
// vitest.config.ts
import { defineConfig } from 'vitest/config';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: ['./src/test/setup.ts'],
    coverage: {
      provider: 'v8',
      reporter: ['text', 'lcov'],
      thresholds: { statements: 80, branches: 75, functions: 80, lines: 80 },
    },
    include: ['src/**/*.{test,spec}.{ts,tsx}'],
  },
});
```

### Assertions

| Assertion | Use |
|-----------|-----|
| `expect(x).toBe(y)` | Strict equality (`===`) |
| `expect(x).toEqual(y)` | Deep equality |
| `expect(x).toStrictEqual(y)` | Deep + no extra properties |
| `expect(x).toBeTruthy()` | Truthy check |
| `expect(x).toBeNull()` | `=== null` |
| `expect(x).toBeUndefined()` | `=== undefined` |
| `expect(x).toBeDefined()` | `!== undefined` |
| `expect(x).toContain(y)` | Array/string contains |
| `expect(x).toHaveLength(n)` | Length check |
| `expect(x).toMatch(/regex/)` | Regex match |
| `expect(x).toThrow()` | Throws error |
| `expect(x).toThrow('msg')` | Throws with message |
| `expect(x).toHaveBeenCalled()` | Mock was called |
| `expect(x).toHaveBeenCalledWith(a)` | Mock called with args |
| `expect(x).toHaveBeenCalledTimes(n)` | Call count |
| `expect(x).resolves.toBe(y)` | Promise resolves to |
| `expect(x).rejects.toThrow()` | Promise rejects |

### Mocking

```typescript
// Mock a module
vi.mock('./db', () => ({
  db: { query: { users: { findFirst: vi.fn() } } },
}));

// Mock a function
const handler = vi.fn().mockReturnValue('result');
const asyncHandler = vi.fn().mockResolvedValue({ id: '1' });

// Spy on existing method
const spy = vi.spyOn(service, 'create');
spy.mockResolvedValueOnce(mockUser);

// Reset
vi.clearAllMocks();   // Clear call history
vi.resetAllMocks();   // Clear history + implementation
vi.restoreAllMocks(); // Restore original implementation

// Timers
vi.useFakeTimers();
vi.advanceTimersByTime(1000);
vi.runAllTimers();
vi.useRealTimers();
```

### React Testing Library

```typescript
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

// Setup
const user = userEvent.setup();

// Queries (priority order)
screen.getByRole('button', { name: /submit/i });  // Accessible name
screen.getByLabelText(/email/i);                   // Form labels
screen.getByPlaceholderText(/search/i);            // Placeholder
screen.getByText(/welcome/i);                      // Visible text
screen.getByTestId('custom-element');               // Last resort

// Query variants
screen.getByRole(...)     // Throws if not found
screen.queryByRole(...)   // Returns null if not found
screen.findByRole(...)    // Async, waits for element

// Events
await user.click(button);
await user.type(input, 'text');
await user.clear(input);
await user.selectOptions(select, 'value');
await user.keyboard('{Enter}');

// Async
await waitFor(() => expect(screen.getByText('loaded')).toBeInTheDocument());
await screen.findByText('loaded'); // shorthand
```

---

## Playwright Quick Reference

```bash
npx playwright test                    # Run all
npx playwright test --headed           # With browser
npx playwright test --ui               # Interactive UI
npx playwright test --debug            # Step-through debugger
npx playwright test auth.spec.ts      # Single file
npx playwright test -g "login"         # Filter by title
npx playwright show-report             # View HTML report
npx playwright codegen localhost:3000  # Record actions
```

### Config

```typescript
// playwright.config.ts
import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
  testDir: './e2e',
  timeout: 30_000,
  retries: process.env.CI ? 2 : 0,
  reporter: process.env.CI ? 'html' : 'list',
  use: {
    baseURL: 'http://localhost:3000',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
  },
  projects: [
    { name: 'chromium', use: { ...devices['Desktop Chrome'] } },
    { name: 'firefox', use: { ...devices['Desktop Firefox'] } },
    { name: 'mobile', use: { ...devices['iPhone 14'] } },
  ],
  webServer: {
    command: 'npm run dev',
    port: 3000,
    reuseExistingServer: !process.env.CI,
  },
});
```

### Page Actions

```typescript
// Navigation
await page.goto('/login');
await page.goBack();
await page.reload();

// Locators (recommended)
page.getByRole('button', { name: 'Submit' });
page.getByLabel('Email');
page.getByPlaceholder('Search...');
page.getByText('Welcome');
page.getByTestId('sidebar');
page.locator('.class-name');        // CSS selector
page.locator('xpath=//div[@id]');   // XPath

// Actions
await page.getByLabel('Email').fill('test@test.com');
await page.getByRole('button', { name: 'Submit' }).click();
await page.getByRole('combobox').selectOption('value');
await page.getByLabel('File').setInputFiles('test.pdf');

// Assertions
await expect(page).toHaveURL('/dashboard');
await expect(page).toHaveTitle(/Dashboard/);
await expect(page.getByText('Welcome')).toBeVisible();
await expect(page.getByRole('alert')).toHaveText('Success');
await expect(page.getByRole('button')).toBeEnabled();
await expect(page.getByRole('list')).toHaveCount(5);

// Waiting
await page.waitForURL('/dashboard');
await page.waitForResponse('**/api/users');
await page.waitForLoadState('networkidle');
await expect(page.getByText('Loaded')).toBeVisible({ timeout: 10_000 });

// Network
await page.route('**/api/users', (route) =>
  route.fulfill({ json: [{ id: '1', name: 'Mock' }] })
);
```

### Auth Fixture

```typescript
// e2e/fixtures.ts
import { test as base } from '@playwright/test';

export const test = base.extend<{ authenticatedPage: Page }>({
  authenticatedPage: async ({ browser }, use) => {
    const context = await browser.newContext({
      storageState: 'e2e/.auth/user.json',
    });
    const page = await context.newPage();
    await use(page);
    await context.close();
  },
});
```

---

## Testing Decision Matrix

| Code Type | Test Type | Tool | Priority |
|-----------|-----------|------|----------|
| Utility function | Unit test | Vitest | High |
| React component | Component test | Vitest + RTL | High |
| Custom hook | Hook test | Vitest + renderHook | Medium |
| API endpoint | Integration test | Vitest + supertest | High |
| Service layer | Integration test | Vitest + test DB | High |
| User flow (login, checkout) | E2E test | Playwright | High |
| Visual design | Visual test | Playwright screenshots | Low |
| Error states | Unit + E2E | Both | Medium |

## Coverage Targets

| Layer | Target | Notes |
|-------|--------|-------|
| Utils/helpers | 95%+ | Pure functions, easy to test |
| Services/business logic | 85%+ | Core value, must be tested |
| API routes | 80%+ | Integration tests preferred |
| Components | 75%+ | Behavior over implementation |
| E2E happy paths | 100% | Every critical user flow |
| E2E error paths | Key ones | Auth failures, network errors |

## Test File Organization

```
src/
  services/
    auth.service.ts
    auth.service.test.ts      # Co-located unit/integration
  components/
    LoginForm.tsx
    LoginForm.test.tsx         # Co-located component test
  test/
    setup.ts                   # Global test setup
    factories.ts               # Test data factories
    helpers.ts                 # Shared test utilities
e2e/
  auth.spec.ts                 # E2E test files
  dashboard.spec.ts
  fixtures.ts                  # Shared fixtures
  pages/                       # Page Objects
    login.page.ts
    dashboard.page.ts
  .auth/                       # Auth state (gitignored)
    user.json
```

## CI Pipeline

```yaml
# .github/workflows/test.yml
jobs:
  unit:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with: { node-version: 20 }
      - run: npm ci
      - run: npx vitest run --coverage

  e2e:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:16
        env:
          POSTGRES_DB: test
          POSTGRES_PASSWORD: test
        ports: ['5432:5432']
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with: { node-version: 20 }
      - run: npm ci
      - run: npx playwright install --with-deps
      - run: npx playwright test
      - uses: actions/upload-artifact@v4
        if: failure()
        with:
          name: playwright-report
          path: playwright-report/
```

## Common Patterns

| Pattern | Code |
|---------|------|
| Test factory | `createTestUser({ role: 'admin' })` — faker defaults + overrides |
| Authenticated request | `request(app).get('/api/me').set('Authorization', \`Bearer ${token}\`)` |
| Wait for async | `await waitFor(() => expect(screen.getByText('Done')).toBeVisible())` |
| Mock API in E2E | `page.route('**/api/data', route => route.fulfill({ json: mock }))` |
| Reset DB between tests | `beforeEach(() => db.execute(sql\`TRUNCATE ... CASCADE\`))` |
| Snapshot test | `expect(tree).toMatchSnapshot()` — use sparingly |
| Error boundary test | Render with error, assert fallback UI appears |
| Hook test | `const { result } = renderHook(() => useMyHook())` |
| Debounce test | `vi.useFakeTimers(); await act(() => vi.advanceTimersByTime(300))` |
