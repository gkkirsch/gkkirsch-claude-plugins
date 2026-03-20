---
name: playwright-patterns
description: >
  Playwright test patterns — locators, assertions, interactions, API testing,
  visual testing, network mocking, and advanced patterns. Complete E2E testing recipes.
  Triggers: "playwright test", "playwright locator", "playwright assertion",
  "playwright mock", "playwright api test", "playwright visual test",
  "e2e test pattern", "end to end test", "browser test".
  NOT for: initial setup or configuration (use playwright-setup).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Playwright Test Patterns

## Locator Strategies (Best to Worst)

```typescript
// 1. Role-based (BEST — semantic, resilient)
page.getByRole("button", { name: "Submit" });
page.getByRole("heading", { level: 2 });
page.getByRole("link", { name: /sign up/i });
page.getByRole("checkbox", { name: "Accept terms" });
page.getByRole("combobox", { name: "Country" });
page.getByRole("tab", { name: "Settings" });
page.getByRole("dialog");
page.getByRole("alert");

// 2. Label/placeholder (form fields)
page.getByLabel("Email address");
page.getByPlaceholder("Search...");

// 3. Text content
page.getByText("Welcome back");
page.getByText(/total: \$\d+/i);  // regex

// 4. Test ID (stable, explicit)
page.getByTestId("submit-button");
page.getByTestId("user-avatar");

// 5. Alt text (images)
page.getByAltText("Company logo");

// 6. Title attribute
page.getByTitle("Close dialog");

// 7. CSS/XPath (LAST RESORT — brittle)
page.locator(".btn-primary");
page.locator('[data-state="open"]');
page.locator("xpath=//div[@class='card']");
```

## Locator Chaining & Filtering

```typescript
// Chain: narrow scope
page
  .getByRole("listitem")
  .filter({ hasText: "Product A" })
  .getByRole("button", { name: "Add to cart" });

// Filter by child locator
page
  .getByRole("row")
  .filter({ has: page.getByText("Active") })
  .getByRole("button", { name: "Edit" });

// Filter by NOT having
page
  .getByRole("listitem")
  .filter({ hasNot: page.getByText("Sold out") });

// nth element
page.getByRole("listitem").nth(0);   // first
page.getByRole("listitem").last();   // last

// Within a container
const sidebar = page.getByRole("complementary");
sidebar.getByRole("link", { name: "Dashboard" });

// Parent traversal (locator chaining)
const row = page.getByRole("row").filter({ hasText: "alice@example.com" });
await row.getByRole("button", { name: "Delete" }).click();
```

## Assertions

```typescript
import { expect } from "@playwright/test";

// Visibility
await expect(page.getByText("Success")).toBeVisible();
await expect(page.getByText("Loading")).toBeHidden();
await expect(page.getByText("Loading")).not.toBeVisible();

// Text content
await expect(page.getByRole("heading")).toHaveText("Dashboard");
await expect(page.getByRole("heading")).toContainText("Dash");
await expect(page.getByRole("status")).toHaveText(/\d+ items/);

// Input values
await expect(page.getByLabel("Name")).toHaveValue("Alice");
await expect(page.getByLabel("Name")).toBeEmpty();
await expect(page.getByLabel("Name")).toBeEditable();
await expect(page.getByLabel("Name")).toBeDisabled();

// Checked/selected
await expect(page.getByRole("checkbox")).toBeChecked();
await expect(page.getByRole("checkbox")).not.toBeChecked();

// Count
await expect(page.getByRole("listitem")).toHaveCount(5);

// Attributes & CSS
await expect(page.getByRole("link")).toHaveAttribute("href", "/about");
await expect(page.getByTestId("card")).toHaveClass(/active/);
await expect(page.getByTestId("card")).toHaveCSS("color", "rgb(0, 0, 0)");

// URL
await expect(page).toHaveURL("/dashboard");
await expect(page).toHaveURL(/\/posts\/\d+/);

// Title
await expect(page).toHaveTitle("My App — Dashboard");

// Page content (full page text)
await expect(page.locator("body")).toContainText("Welcome");

// Soft assertions (don't stop test on failure)
await expect.soft(page.getByText("Beta")).toBeVisible();
await expect.soft(page.getByText("v2.0")).toBeVisible();
// Test continues even if soft assertions fail
```

## Interactions

```typescript
// Click
await page.getByRole("button", { name: "Save" }).click();
await page.getByRole("button").dblclick();
await page.getByRole("button").click({ button: "right" });
await page.getByRole("button").click({ modifiers: ["Shift"] });
await page.getByRole("button").click({ force: true }); // skip actionability

// Fill (clears then types)
await page.getByLabel("Email").fill("alice@example.com");

// Type (keystroke by keystroke — for autocomplete, debounce)
await page.getByLabel("Search").pressSequentially("react hooks", { delay: 100 });

// Clear
await page.getByLabel("Email").clear();

// Select dropdown
await page.getByLabel("Country").selectOption("US");
await page.getByLabel("Country").selectOption({ label: "United States" });
await page.getByLabel("Tags").selectOption(["react", "typescript"]); // multi

// Checkbox / radio
await page.getByRole("checkbox", { name: "Accept" }).check();
await page.getByRole("checkbox", { name: "Accept" }).uncheck();
await page.getByRole("radio", { name: "Monthly" }).check();

// File upload
await page.getByLabel("Upload").setInputFiles("./test-data/photo.jpg");
await page.getByLabel("Upload").setInputFiles([]); // clear
await page.getByLabel("Upload").setInputFiles([
  "./file1.pdf",
  "./file2.pdf",
]); // multiple

// Drag and drop
await page.getByTestId("item-1").dragTo(page.getByTestId("drop-zone"));

// Hover
await page.getByText("Menu").hover();

// Focus
await page.getByLabel("Name").focus();

// Keyboard
await page.keyboard.press("Escape");
await page.keyboard.press("Control+a");
await page.getByLabel("Search").press("Enter");

// Scroll
await page.getByTestId("feed").evaluate(el => el.scrollTop = el.scrollHeight);
await page.getByText("Load more").scrollIntoViewIfNeeded();
```

## Waiting Patterns

```typescript
// GOOD: Wait for specific condition (auto-retry)
await expect(page.getByText("Saved")).toBeVisible();
await page.waitForURL("/dashboard");
await page.waitForLoadState("networkidle");

// Wait for response
const responsePromise = page.waitForResponse(
  resp => resp.url().includes("/api/save") && resp.status() === 200
);
await page.getByRole("button", { name: "Save" }).click();
const response = await responsePromise;
const data = await response.json();

// Wait for request
const requestPromise = page.waitForRequest("**/api/analytics");
await page.getByRole("button", { name: "Track" }).click();
await requestPromise;

// Wait for navigation
const navigationPromise = page.waitForURL("**/success");
await page.getByRole("button", { name: "Submit" }).click();
await navigationPromise;

// Wait for element state
await page.getByRole("button").waitFor({ state: "visible" });
await page.getByRole("spinner").waitFor({ state: "hidden" });
await page.getByRole("dialog").waitFor({ state: "detached" });

// BAD: Never use fixed timeouts
// await page.waitForTimeout(3000); // WRONG — flaky and slow
```

## Network Mocking & Interception

```typescript
// Mock API response
await page.route("**/api/users", async route => {
  await route.fulfill({
    status: 200,
    contentType: "application/json",
    body: JSON.stringify({
      users: [
        { id: 1, name: "Alice", email: "alice@test.com" },
        { id: 2, name: "Bob", email: "bob@test.com" },
      ],
    }),
  });
});

// Mock with response modification
await page.route("**/api/products", async route => {
  const response = await route.fetch(); // get real response
  const json = await response.json();
  json.products[0].price = 0; // modify
  await route.fulfill({ response, body: JSON.stringify(json) });
});

// Mock error responses
await page.route("**/api/save", route =>
  route.fulfill({ status: 500, body: "Internal Server Error" })
);

// Mock network failure
await page.route("**/api/data", route => route.abort());

// Conditional mocking
await page.route("**/api/**", async route => {
  if (route.request().method() === "DELETE") {
    await route.fulfill({ status: 403 });
  } else {
    await route.continue();
  }
});

// Remove mock
await page.unroute("**/api/users");

// HAR recording (record real responses, replay later)
// Record:
await page.routeFromHAR("tests/fixtures/api.har", {
  url: "**/api/**",
  update: true,  // records to file
});
// Replay:
await page.routeFromHAR("tests/fixtures/api.har", {
  url: "**/api/**",
  update: false, // replays from file
});
```

## API Testing (No Browser)

```typescript
import { test, expect } from "@playwright/test";

test.describe("API Tests", () => {
  let token: string;

  test.beforeAll(async ({ request }) => {
    const response = await request.post("/api/auth/login", {
      data: { email: "admin@test.com", password: "password" },
    });
    const body = await response.json();
    token = body.token;
  });

  test("create a post", async ({ request }) => {
    const response = await request.post("/api/posts", {
      headers: { Authorization: `Bearer ${token}` },
      data: {
        title: "E2E Test Post",
        body: "Created by Playwright",
        status: "draft",
      },
    });

    expect(response.ok()).toBeTruthy();
    expect(response.status()).toBe(201);

    const post = await response.json();
    expect(post.data).toMatchObject({
      title: "E2E Test Post",
      status: "draft",
    });
    expect(post.data.id).toBeTruthy();
  });

  test("list posts with pagination", async ({ request }) => {
    const response = await request.get("/api/posts", {
      headers: { Authorization: `Bearer ${token}` },
      params: { page: "1", limit: "10" },
    });

    expect(response.ok()).toBeTruthy();
    const body = await response.json();
    expect(body.data).toBeInstanceOf(Array);
    expect(body.meta.total).toBeGreaterThan(0);
  });

  test("validation errors return 400", async ({ request }) => {
    const response = await request.post("/api/posts", {
      headers: { Authorization: `Bearer ${token}` },
      data: { title: "" }, // invalid
    });

    expect(response.status()).toBe(400);
    const body = await response.json();
    expect(body.errors).toBeDefined();
  });
});
```

## Visual / Screenshot Testing

```typescript
// Full page snapshot
await expect(page).toHaveScreenshot("dashboard.png");

// Element snapshot
await expect(page.getByTestId("chart")).toHaveScreenshot("chart.png");

// With options
await expect(page).toHaveScreenshot("hero.png", {
  maxDiffPixels: 100,          // allow small differences
  maxDiffPixelRatio: 0.01,     // 1% pixel difference ok
  threshold: 0.2,              // color diff threshold (0-1)
  animations: "disabled",      // freeze animations
  mask: [page.getByTestId("timestamp")], // mask dynamic content
  fullPage: true,              // capture entire scrollable page
});

// Update baselines: npx playwright test --update-snapshots

// Clip to region
await expect(page).toHaveScreenshot("header.png", {
  clip: { x: 0, y: 0, width: 1280, height: 100 },
});
```

## Multi-Tab & Popup Handling

```typescript
// Handle new tab/window
const pagePromise = context.waitForEvent("page");
await page.getByRole("link", { name: "Open in new tab" }).click();
const newPage = await pagePromise;
await newPage.waitForLoadState();
await expect(newPage).toHaveURL(/\/details/);
await newPage.close();

// Handle popup
const popupPromise = page.waitForEvent("popup");
await page.getByRole("button", { name: "OAuth Login" }).click();
const popup = await popupPromise;
await popup.waitForLoadState();
await popup.getByLabel("Email").fill("user@test.com");
await popup.getByRole("button", { name: "Allow" }).click();
// popup closes, back to main page
await expect(page.getByText("Logged in")).toBeVisible();
```

## Dialog Handling

```typescript
// Auto-accept alert/confirm/prompt
page.on("dialog", dialog => dialog.accept());

// Custom dialog handling
page.on("dialog", async dialog => {
  expect(dialog.type()).toBe("confirm");
  expect(dialog.message()).toContain("Are you sure");
  await dialog.accept();
});
await page.getByRole("button", { name: "Delete" }).click();

// Prompt with input
page.on("dialog", dialog => dialog.accept("My answer"));
await page.getByRole("button", { name: "Rename" }).click();

// Dismiss
page.on("dialog", dialog => dialog.dismiss());
```

## iframe Handling

```typescript
// By frame locator (recommended)
const frame = page.frameLocator("#payment-iframe");
await frame.getByLabel("Card number").fill("4242424242424242");
await frame.getByLabel("Expiry").fill("12/30");
await frame.getByRole("button", { name: "Pay" }).click();

// Nested iframes
const outer = page.frameLocator("#outer");
const inner = outer.frameLocator("#inner");
await inner.getByText("Content").click();
```

## Test Organization

```typescript
import { test, expect } from "@playwright/test";

test.describe("User Management", () => {
  // Run tests in this block serially
  test.describe.configure({ mode: "serial" });

  test.beforeEach(async ({ page }) => {
    await page.goto("/admin/users");
  });

  test("displays user list", async ({ page }) => {
    await expect(page.getByRole("table")).toBeVisible();
    await expect(page.getByRole("row")).toHaveCount(11); // header + 10
  });

  test("filters by status", async ({ page }) => {
    await page.getByRole("combobox", { name: "Status" }).selectOption("active");
    await expect(page.getByRole("row")).toHaveCount(6);
  });

  test.skip("exports to CSV", async ({ page }) => {
    // TODO: implement export feature
  });

  test("pagination works", async ({ page }) => {
    await page.getByRole("button", { name: "Next" }).click();
    await expect(page).toHaveURL(/page=2/);
  });
});

// Parameterized tests
const viewports = [
  { name: "desktop", width: 1280, height: 720 },
  { name: "tablet", width: 768, height: 1024 },
  { name: "mobile", width: 375, height: 812 },
];

for (const vp of viewports) {
  test(`responsive layout — ${vp.name}`, async ({ page }) => {
    await page.setViewportSize({ width: vp.width, height: vp.height });
    await page.goto("/");
    if (vp.width < 768) {
      await expect(page.getByRole("button", { name: "Menu" })).toBeVisible();
    } else {
      await expect(page.getByRole("navigation")).toBeVisible();
    }
  });
}

// Tags for selective execution
test("critical checkout flow @smoke", async ({ page }) => {
  // npx playwright test --grep @smoke
});

test("edge case handling @regression", async ({ page }) => {
  // npx playwright test --grep @regression
});
```

## Common Test Recipes

### Login Flow

```typescript
test("login with valid credentials", async ({ page }) => {
  await page.goto("/login");
  await page.getByLabel("Email").fill("user@example.com");
  await page.getByLabel("Password").fill("password123");
  await page.getByRole("button", { name: "Sign in" }).click();

  await expect(page).toHaveURL("/dashboard");
  await expect(page.getByText("Welcome back")).toBeVisible();
});

test("login shows error for invalid credentials", async ({ page }) => {
  await page.goto("/login");
  await page.getByLabel("Email").fill("wrong@example.com");
  await page.getByLabel("Password").fill("wrong");
  await page.getByRole("button", { name: "Sign in" }).click();

  await expect(page.getByRole("alert")).toContainText("Invalid credentials");
  await expect(page).toHaveURL("/login"); // stays on login
});
```

### CRUD Operations

```typescript
test("full CRUD lifecycle", async ({ page }) => {
  await page.goto("/posts");

  // Create
  await page.getByRole("button", { name: "New Post" }).click();
  await page.getByLabel("Title").fill("Test Post");
  await page.getByLabel("Body").fill("Test content");
  await page.getByRole("button", { name: "Publish" }).click();
  await expect(page.getByText("Post created")).toBeVisible();

  // Read
  await expect(page.getByRole("heading", { name: "Test Post" })).toBeVisible();

  // Update
  await page.getByRole("button", { name: "Edit" }).click();
  await page.getByLabel("Title").fill("Updated Post");
  await page.getByRole("button", { name: "Save" }).click();
  await expect(page.getByText("Post updated")).toBeVisible();

  // Delete
  await page.getByRole("button", { name: "Delete" }).click();
  await page.getByRole("button", { name: "Confirm" }).click();
  await expect(page.getByText("Post deleted")).toBeVisible();
  await expect(page.getByText("Updated Post")).not.toBeVisible();
});
```

### Form Validation

```typescript
test("validates required fields", async ({ page }) => {
  await page.goto("/contact");
  await page.getByRole("button", { name: "Submit" }).click();

  await expect(page.getByText("Name is required")).toBeVisible();
  await expect(page.getByText("Email is required")).toBeVisible();

  // Fill one field, check other still shows error
  await page.getByLabel("Name").fill("Alice");
  await page.getByRole("button", { name: "Submit" }).click();
  await expect(page.getByText("Name is required")).not.toBeVisible();
  await expect(page.getByText("Email is required")).toBeVisible();
});

test("validates email format", async ({ page }) => {
  await page.goto("/contact");
  await page.getByLabel("Email").fill("not-an-email");
  await page.getByLabel("Email").blur();
  await expect(page.getByText("Invalid email")).toBeVisible();

  await page.getByLabel("Email").fill("valid@example.com");
  await page.getByLabel("Email").blur();
  await expect(page.getByText("Invalid email")).not.toBeVisible();
});
```

### Search & Filter

```typescript
test("search filters results in real-time", async ({ page }) => {
  await page.goto("/products");
  await expect(page.getByRole("article")).toHaveCount(20);

  await page.getByPlaceholder("Search products").fill("widget");

  // Wait for debounced search to fire
  await page.waitForResponse(resp =>
    resp.url().includes("/api/products") && resp.url().includes("q=widget")
  );

  const results = page.getByRole("article");
  await expect(results).toHaveCount(3);
  for (const result of await results.all()) {
    await expect(result).toContainText(/widget/i);
  }
});
```

### File Download

```typescript
test("downloads report as CSV", async ({ page }) => {
  await page.goto("/reports");

  const downloadPromise = page.waitForEvent("download");
  await page.getByRole("button", { name: "Export CSV" }).click();
  const download = await downloadPromise;

  expect(download.suggestedFilename()).toBe("report.csv");

  // Save and verify content
  const path = await download.path();
  const content = await fs.readFile(path!, "utf-8");
  expect(content).toContain("Name,Email,Status");
});
```

## Gotchas

1. **`toBeVisible()` vs `toBeHidden()` vs `not.toBeVisible()`**: `toBeHidden` checks element exists but is hidden. `not.toBeVisible` passes if element doesn't exist OR is hidden. Use `not.toBeVisible` when the element might be removed from DOM entirely.

2. **Auto-waiting is your friend**: Playwright auto-waits for elements to be actionable before clicking, filling, etc. Don't add manual waits before interactions. If you need to wait, wait for a specific condition, not a timeout.

3. **Locator vs ElementHandle**: Always use locators (`page.getByRole(...)`) not element handles (`page.$(...)`). Locators are lazy, auto-retry, and re-query on each use. Element handles are snapshots that go stale.

4. **Test isolation**: Each test gets a fresh page but shares the browser context. If test A sets localStorage, test B sees it. Use `test.use({ storageState: undefined })` for clean state, or rely on the auth setup pattern.

5. **`getByRole` uses accessible names**: The `name` option matches the accessible name, not necessarily the visible text. A button with `aria-label="Close"` matches `getByRole("button", { name: "Close" })` even if it only shows an X icon.

6. **Network mocking is per-page**: Routes set with `page.route()` only apply to that page instance. If a new tab opens, it won't have the mocks. Use `context.route()` for context-wide mocking.
