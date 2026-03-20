---
name: playwright-visual-testing
description: >
  Visual regression testing and screenshot comparison with Playwright.
  Use when implementing visual testing, screenshot assertions,
  component visual tests, or full-page visual regression pipelines.
  Triggers: "visual testing", "screenshot testing", "visual regression",
  "playwright screenshots", "toHaveScreenshot", "pixel comparison",
  "visual diff", "component visual test".
  NOT for: Playwright setup (see playwright-setup), functional E2E tests (see playwright-patterns), accessibility testing.
version: 1.0.0
allowed-tools: Read, Grep, Glob, Bash
---

# Playwright Visual Testing

## Basic Screenshot Assertions

```typescript
// tests/visual/homepage.spec.ts
import { test, expect } from '@playwright/test';

test.describe('Homepage visual tests', () => {
  test('full page screenshot', async ({ page }) => {
    await page.goto('/');
    // Wait for all content to load
    await page.waitForLoadState('networkidle');

    // Full page screenshot comparison
    await expect(page).toHaveScreenshot('homepage-full.png', {
      fullPage: true,
      maxDiffPixelRatio: 0.01, // Allow 1% pixel difference
    });
  });

  test('hero section screenshot', async ({ page }) => {
    await page.goto('/');
    const hero = page.locator('[data-testid="hero-section"]');
    await hero.waitFor({ state: 'visible' });

    // Element-level screenshot
    await expect(hero).toHaveScreenshot('hero-section.png', {
      maxDiffPixels: 100, // Allow up to 100 pixels different
    });
  });

  test('navigation in different states', async ({ page }) => {
    await page.goto('/');
    const nav = page.locator('nav');

    // Default state
    await expect(nav).toHaveScreenshot('nav-default.png');

    // Hover state on menu item
    await page.locator('nav a:first-child').hover();
    await expect(nav).toHaveScreenshot('nav-hover.png');

    // Mobile hamburger menu
    await page.setViewportSize({ width: 375, height: 667 });
    await expect(nav).toHaveScreenshot('nav-mobile.png');
  });
});
```

## Configuration for Visual Testing

```typescript
// playwright.config.ts
import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
  testDir: './tests/visual',
  snapshotDir: './tests/visual/__snapshots__',
  snapshotPathTemplate: '{snapshotDir}/{testFilePath}/{testName}/{projectName}{ext}',

  // Update snapshots: npx playwright test --update-snapshots
  updateSnapshots: 'none', // 'all' | 'none' | 'missing'

  expect: {
    toHaveScreenshot: {
      // Default comparison settings
      maxDiffPixelRatio: 0.01,
      threshold: 0.2, // Per-pixel color threshold (0-1)
      animations: 'disabled', // Freeze CSS animations
    },
  },

  // Test across multiple browsers and viewports
  projects: [
    {
      name: 'desktop-chrome',
      use: {
        ...devices['Desktop Chrome'],
        viewport: { width: 1280, height: 720 },
      },
    },
    {
      name: 'desktop-firefox',
      use: {
        ...devices['Desktop Firefox'],
        viewport: { width: 1280, height: 720 },
      },
    },
    {
      name: 'mobile-safari',
      use: {
        ...devices['iPhone 14'],
      },
    },
    {
      name: 'tablet',
      use: {
        viewport: { width: 768, height: 1024 },
        deviceScaleFactor: 2,
      },
    },
    // Dark mode variant
    {
      name: 'dark-mode',
      use: {
        ...devices['Desktop Chrome'],
        colorScheme: 'dark',
      },
    },
  ],
});
```

## Component Visual Testing

```typescript
// tests/visual/components.spec.ts
import { test, expect } from '@playwright/test';

// Test component variants in isolation
test.describe('Button component', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate to component preview page or Storybook
    await page.goto('/storybook/iframe.html?id=components-button--default');
  });

  const variants = ['primary', 'secondary', 'danger', 'ghost'];
  const sizes = ['sm', 'md', 'lg'];
  const states = ['default', 'hover', 'focus', 'disabled'];

  for (const variant of variants) {
    for (const size of sizes) {
      test(`${variant} ${size} button`, async ({ page }) => {
        await page.goto(
          `/storybook/iframe.html?id=components-button--default&args=variant:${variant};size:${size}`
        );
        const button = page.locator('button');
        await expect(button).toHaveScreenshot(`button-${variant}-${size}.png`);
      });
    }
  }

  test('button states', async ({ page }) => {
    const button = page.locator('button');

    // Default
    await expect(button).toHaveScreenshot('button-state-default.png');

    // Hover
    await button.hover();
    await expect(button).toHaveScreenshot('button-state-hover.png');

    // Focus
    await button.focus();
    await expect(button).toHaveScreenshot('button-state-focus.png');

    // Disabled
    await page.goto('/storybook/iframe.html?id=components-button--default&args=disabled:true');
    await expect(page.locator('button')).toHaveScreenshot('button-state-disabled.png');
  });
});
```

## Handling Dynamic Content

```typescript
// tests/visual/dynamic-content.spec.ts
import { test, expect } from '@playwright/test';

test.describe('Pages with dynamic content', () => {
  test('mask dynamic elements', async ({ page }) => {
    await page.goto('/dashboard');

    // Mask elements that change between runs
    await expect(page).toHaveScreenshot('dashboard.png', {
      mask: [
        page.locator('[data-testid="current-time"]'),
        page.locator('[data-testid="user-avatar"]'),
        page.locator('[data-testid="live-stats"]'),
        page.locator('.advertisement'),
      ],
      maskColor: '#FF00FF', // Visible magenta mask for debugging
    });
  });

  test('replace dynamic text before screenshot', async ({ page }) => {
    await page.goto('/profile');

    // Replace dynamic content with stable placeholders
    await page.evaluate(() => {
      // Replace dates
      document.querySelectorAll('[data-testid="date"]').forEach(el => {
        el.textContent = '2026-01-01';
      });
      // Replace random IDs
      document.querySelectorAll('[data-testid="user-id"]').forEach(el => {
        el.textContent = 'USR-XXXXX';
      });
    });

    await expect(page).toHaveScreenshot('profile-stable.png');
  });

  test('wait for fonts and images', async ({ page }) => {
    await page.goto('/landing');

    // Wait for web fonts to load
    await page.evaluate(() => document.fonts.ready);

    // Wait for all images to load
    await page.waitForFunction(() => {
      const images = Array.from(document.querySelectorAll('img'));
      return images.every(img => img.complete && img.naturalWidth > 0);
    });

    // Wait for animations to finish
    await page.evaluate(() => {
      return new Promise<void>(resolve => {
        const animations = document.getAnimations();
        if (animations.length === 0) return resolve();
        Promise.all(animations.map(a => a.finished)).then(() => resolve());
      });
    });

    await expect(page).toHaveScreenshot('landing-loaded.png', {
      animations: 'disabled',
    });
  });
});
```

## CI Pipeline for Visual Tests

```yaml
# .github/workflows/visual-tests.yml
name: Visual Tests
on: [pull_request]

jobs:
  visual:
    runs-on: ubuntu-latest
    container:
      image: mcr.microsoft.com/playwright:v1.50.0-jammy
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with: { node-version: 20 }
      - run: npm ci

      # Start app server
      - run: npm run build && npm run start &
      - run: npx wait-on http://localhost:3000

      # Run visual tests
      - run: npx playwright test --project=desktop-chrome tests/visual/
        env:
          CI: true

      # Upload diff artifacts on failure
      - uses: actions/upload-artifact@v4
        if: failure()
        with:
          name: visual-test-results
          path: |
            test-results/
            tests/visual/__snapshots__/
          retention-days: 7

      # Comment on PR with diff images
      - name: Comment visual diffs
        if: failure()
        uses: actions/github-script@v7
        with:
          script: |
            const fs = require('fs');
            const diffs = fs.readdirSync('test-results')
              .filter(f => f.includes('-diff'));
            if (diffs.length > 0) {
              await github.rest.issues.createComment({
                owner: context.repo.owner,
                repo: context.repo.repo,
                issue_number: context.issue.number,
                body: `## Visual Regression Detected\n\n${diffs.length} screenshot(s) changed. Check the [test artifacts](${context.serverUrl}/${context.repo.owner}/${context.repo.repo}/actions/runs/${context.runId}) for diff images.`
              });
            }
```

## Gotchas

1. **Font rendering differs across OS** -- The same font renders differently on macOS, Windows, and Linux. CI (usually Linux) will always produce different screenshots than local development (usually macOS). Generate baseline screenshots in CI, not locally. Or use Docker containers with identical font configurations for both.

2. **Anti-aliasing causes flaky tests** -- Sub-pixel rendering and anti-aliasing vary between GPU drivers and browser versions. A threshold of `0` (exact match) will produce constant false positives. Start with `threshold: 0.2` and `maxDiffPixelRatio: 0.01`, then tighten only if you're getting false negatives.

3. **CSS animations cause non-deterministic screenshots** -- A screenshot captured mid-animation differs from one captured after. Use `animations: 'disabled'` in config or `page.evaluate(() => document.getAnimations().forEach(a => a.finish()))` to force-finish all animations. Don't forget CSS transitions triggered by hover/focus states.

4. **Lazy-loaded content below the fold** -- `fullPage: true` scrolls the page to capture everything, which triggers lazy loading. But lazy-loaded images might not finish loading before the screenshot. After scrolling, wait for `networkidle` and verify images are loaded with `img.complete && img.naturalWidth > 0`.

5. **Updating snapshots in PRs** -- Running `--update-snapshots` locally and committing produces platform-specific baselines. The correct workflow: update snapshots in CI (same environment as tests), download the new snapshot artifacts, and commit those. Never update snapshots locally unless you test locally with the same Docker image as CI.

6. **Screenshot file naming collisions** -- Without `snapshotPathTemplate`, screenshots from different projects (browsers) overwrite each other. Always include `{projectName}` in the template: `{snapshotDir}/{testFilePath}/{testName}/{projectName}{ext}`. This keeps desktop-chrome and mobile-safari snapshots separate.
