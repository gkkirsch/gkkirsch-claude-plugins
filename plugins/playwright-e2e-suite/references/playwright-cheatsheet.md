# Playwright Cheatsheet

## Locators (Best to Worst)

```typescript
page.getByRole("button", { name: "Save" })  // semantic, resilient
page.getByLabel("Email")                     // form fields
page.getByPlaceholder("Search...")           // inputs
page.getByText("Welcome")                   // visible text
page.getByTestId("submit-btn")              // stable test IDs
page.getByAltText("Logo")                   // images
page.locator(".css-class")                  // last resort
```

## Locator Chaining

```typescript
// Filter by text
page.getByRole("listitem").filter({ hasText: "Active" })

// Filter by child
page.getByRole("row").filter({ has: page.getByText("Alice") })

// Chain into child
page.getByRole("row").filter({ hasText: "Alice" }).getByRole("button", { name: "Edit" })

// nth / last
page.getByRole("listitem").nth(0)
page.getByRole("listitem").last()
```

## Assertions

```typescript
await expect(locator).toBeVisible()
await expect(locator).toBeHidden()
await expect(locator).toHaveText("exact text")
await expect(locator).toContainText("partial")
await expect(locator).toHaveValue("input value")
await expect(locator).toBeChecked()
await expect(locator).toBeDisabled()
await expect(locator).toBeEditable()
await expect(locator).toBeEmpty()
await expect(locator).toHaveCount(5)
await expect(locator).toHaveAttribute("href", "/about")
await expect(locator).toHaveClass(/active/)
await expect(locator).toHaveCSS("color", "rgb(0,0,0)")
await expect(page).toHaveURL("/dashboard")
await expect(page).toHaveTitle("My App")
await expect(page).toHaveScreenshot("name.png")
```

## Interactions

```typescript
await locator.click()
await locator.dblclick()
await locator.click({ button: "right" })
await locator.fill("text")                    // clear + type
await locator.pressSequentially("text", { delay: 50 })  // keystroke
await locator.clear()
await locator.selectOption("value")            // dropdown
await locator.check()                          // checkbox
await locator.uncheck()
await locator.setInputFiles("./file.pdf")      // upload
await locator.dragTo(target)
await locator.hover()
await locator.focus()
await locator.press("Enter")
await page.keyboard.press("Escape")
await locator.scrollIntoViewIfNeeded()
```

## Waiting

```typescript
// GOOD — condition-based
await expect(page.getByText("Done")).toBeVisible()
await page.waitForURL("/dashboard")
await page.waitForLoadState("networkidle")
await page.waitForResponse(r => r.url().includes("/api") && r.ok())
await page.waitForRequest("**/api/save")
await locator.waitFor({ state: "visible" })
await locator.waitFor({ state: "hidden" })

// BAD — never use
// await page.waitForTimeout(3000)
```

## Network Mocking

```typescript
// Mock response
await page.route("**/api/users", route => route.fulfill({
  status: 200,
  contentType: "application/json",
  body: JSON.stringify({ users: [] }),
}))

// Modify real response
await page.route("**/api/data", async route => {
  const resp = await route.fetch()
  const json = await resp.json()
  json.modified = true
  await route.fulfill({ response: resp, body: JSON.stringify(json) })
})

// Mock error
await page.route("**/api/save", route => route.fulfill({ status: 500 }))

// Abort
await page.route("**/api/track", route => route.abort())

// Remove mock
await page.unroute("**/api/users")
```

## API Testing (No Browser)

```typescript
test("POST /api/items", async ({ request }) => {
  const resp = await request.post("/api/items", {
    headers: { Authorization: "Bearer token" },
    data: { name: "Test" },
  })
  expect(resp.status()).toBe(201)
  const body = await resp.json()
  expect(body.id).toBeTruthy()
})
```

## Auth Setup

```typescript
// tests/auth.setup.ts
import { test as setup } from "@playwright/test"

setup("authenticate", async ({ page }) => {
  await page.goto("/login")
  await page.getByLabel("Email").fill("user@test.com")
  await page.getByLabel("Password").fill("pass")
  await page.getByRole("button", { name: "Sign in" }).click()
  await page.waitForURL("/dashboard")
  await page.context().storageState({ path: "tests/.auth/user.json" })
})
```

```typescript
// playwright.config.ts — project depends on setup
projects: [
  { name: "setup", testMatch: /.*\.setup\.ts/ },
  {
    name: "chromium",
    use: { storageState: "tests/.auth/user.json" },
    dependencies: ["setup"],
  },
]
```

## Custom Fixtures

```typescript
import { test as base } from "@playwright/test"
import { DashboardPage } from "../pages/dashboard.page"

export const test = base.extend<{ dashboard: DashboardPage }>({
  dashboard: async ({ page }, use) => {
    const dashboard = new DashboardPage(page)
    await dashboard.goto()
    await use(dashboard)
  },
})
```

## Page Object Model

```typescript
export class DashboardPage {
  constructor(private page: Page) {}

  readonly heading = this.page.getByRole("heading", { name: "Dashboard" })
  readonly createBtn = this.page.getByRole("button", { name: "Create" })

  async goto() { await this.page.goto("/dashboard") }
  async create(title: string) {
    await this.createBtn.click()
    await this.page.getByLabel("Title").fill(title)
    await this.page.getByRole("button", { name: "Save" }).click()
  }
}
```

## Visual Testing

```typescript
await expect(page).toHaveScreenshot("name.png")
await expect(page).toHaveScreenshot("name.png", {
  maxDiffPixels: 100,
  animations: "disabled",
  mask: [page.getByTestId("dynamic")],
  fullPage: true,
})
// Update: npx playwright test --update-snapshots
```

## Multi-Tab / Popup / Dialog

```typescript
// New tab
const newPage = await context.waitForEvent("page")
// after click that opens tab
await newPage.waitForLoadState()

// Dialog
page.on("dialog", d => d.accept())       // accept all
page.on("dialog", d => d.dismiss())      // dismiss all
page.on("dialog", d => d.accept("input")) // prompt

// iframe
const frame = page.frameLocator("#iframe-id")
await frame.getByLabel("Field").fill("value")
```

## File Download

```typescript
const download = await page.waitForEvent("download")
// after click
expect(download.suggestedFilename()).toBe("report.csv")
const path = await download.path()
```

## CLI Commands

```bash
npx playwright test                          # run all
npx playwright test tests/login.spec.ts      # specific file
npx playwright test -g "should login"        # by title
npx playwright test --project=chromium       # specific browser
npx playwright test --headed                 # see browser
npx playwright test --ui                     # interactive UI
npx playwright test --debug                  # step through
npx playwright test --grep @smoke            # by tag
npx playwright codegen http://localhost:3000  # record tests
npx playwright show-report                   # view report
npx playwright test --update-snapshots       # update baselines
npx playwright install --with-deps           # browsers + deps (CI)
```

## Config Quick Reference

```typescript
// playwright.config.ts
export default defineConfig({
  testDir: "./tests",
  fullyParallel: true,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  timeout: 30_000,
  expect: { timeout: 5_000 },
  use: {
    baseURL: "http://localhost:3000",
    trace: "on-first-retry",
    screenshot: "only-on-failure",
    video: "retain-on-failure",
  },
  webServer: {
    command: "npm run dev",
    url: "http://localhost:3000",
    reuseExistingServer: !process.env.CI,
  },
})
```

## Gotchas

1. `not.toBeVisible()` passes if element missing OR hidden. `toBeHidden()` requires element exists.
2. `storageState` files contain secrets — gitignore `tests/.auth/`.
3. `--with-deps` required on CI for browser system dependencies.
4. Visual snapshots are OS-specific — generate baselines on CI OS.
5. `page.route()` is per-page. Use `context.route()` for context-wide mocks.
6. Each test shares browser context (cookies/storage). Use fixtures for isolation.
7. Never use `waitForTimeout()` — always wait for a condition.
8. `getByRole` matches accessible name, not just visible text.
