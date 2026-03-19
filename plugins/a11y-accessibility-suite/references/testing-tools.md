# Accessibility Testing Tools & Techniques

Guide to automated, semi-automated, and manual accessibility testing.

---

## Automated Testing Tools

### axe-core (Deque)

The most widely used a11y testing engine. Powers many other tools.

```bash
# Install
npm install --save-dev @axe-core/react  # React DevTools integration
npm install --save-dev @axe-core/cli    # CLI scanner
npm install --save-dev axe-core         # Core library

# CLI usage
npx @axe-core/cli https://example.com
npx @axe-core/cli https://example.com --tags wcag2a,wcag2aa
```

**React DevTools integration:**
```tsx
// src/index.tsx (development only)
if (process.env.NODE_ENV === 'development') {
  const axe = require('@axe-core/react');
  axe(React, ReactDOM, 1000);
  // Logs violations to browser console
}
```

**Unit tests with jest-axe:**
```tsx
import { render } from '@testing-library/react';
import { axe, toHaveNoViolations } from 'jest-axe';

expect.extend(toHaveNoViolations);

test('form has no a11y violations', async () => {
  const { container } = render(<LoginForm />);
  const results = await axe(container);
  expect(results).toHaveNoViolations();
});
```

**What axe catches**: ~57% of WCAG issues (all automatable ones). Missing alt text, color contrast, missing labels, invalid ARIA, duplicate IDs, keyboard traps, and more.

**What axe can't catch**: Meaningful alt text quality, logical heading hierarchy, correct focus order, usable keyboard interaction, content comprehensibility.

### eslint-plugin-jsx-a11y

Static analysis of JSX for common a11y issues.

```bash
npm install --save-dev eslint-plugin-jsx-a11y
```

```json
// .eslintrc.json
{
  "extends": ["plugin:jsx-a11y/recommended"],
  "plugins": ["jsx-a11y"]
}
```

**Key rules:**
- `alt-text` — img, area, input[type=image], object must have alt
- `anchor-has-content` — anchors must have content
- `aria-props` — valid ARIA attributes
- `aria-proptypes` — correct ARIA value types
- `aria-role` — valid role values
- `aria-unsupported-elements` — no ARIA on invalid elements
- `click-events-have-key-events` — onClick must pair with onKeyDown/onKeyUp
- `heading-has-content` — headings must have content
- `html-has-lang` — html element must have lang
- `img-redundant-alt` — alt text shouldn't contain "image" or "picture"
- `interactive-supports-focus` — interactive elements must be focusable
- `label-has-associated-control` — labels must be linked to inputs
- `no-autofocus` — no autoFocus (disorients users)
- `no-noninteractive-element-interactions` — no handlers on non-interactive elements
- `no-static-element-interactions` — no handlers on static elements without role
- `role-has-required-aria-props` — roles have required attributes
- `tabindex-no-positive` — no positive tabindex values

### Lighthouse (Google)

Built into Chrome DevTools. Audits accessibility, performance, SEO, best practices.

```bash
# CLI usage
npx lighthouse https://example.com --only-categories=accessibility --output=json

# In Chrome DevTools
# DevTools → Lighthouse tab → Check "Accessibility" → Generate report
```

**Scoring**: 0-100 based on axe-core rules. Each violation weighted by impact (critical > serious > moderate > minor).

### Pa11y

Open-source CLI accessibility checker.

```bash
npm install -g pa11y

# Basic scan
pa11y https://example.com

# WCAG 2.2 AA
pa11y --standard WCAG2AA https://example.com

# Specific actions (login, navigate)
pa11y --actions "set field #email to test@test.com" --actions "click element #submit" https://example.com

# CI integration
pa11y-ci --config .pa11yci.json
```

```json
// .pa11yci.json
{
  "defaults": {
    "standard": "WCAG2AA",
    "timeout": 10000
  },
  "urls": [
    "http://localhost:3000",
    "http://localhost:3000/about",
    "http://localhost:3000/contact",
    {
      "url": "http://localhost:3000/dashboard",
      "actions": [
        "set field #email to test@test.com",
        "set field #password to password123",
        "click element #login-button",
        "wait for url to be http://localhost:3000/dashboard"
      ]
    }
  ]
}
```

### Playwright Accessibility Testing

```typescript
import { test, expect } from '@playwright/test';
import AxeBuilder from '@axe-core/playwright';

test('homepage has no a11y violations', async ({ page }) => {
  await page.goto('http://localhost:3000');

  const accessibilityScanResults = await new AxeBuilder({ page })
    .withTags(['wcag2a', 'wcag2aa', 'wcag22aa'])
    .analyze();

  expect(accessibilityScanResults.violations).toEqual([]);
});

test('modal is accessible', async ({ page }) => {
  await page.goto('http://localhost:3000');
  await page.click('[data-testid="open-modal"]');

  const results = await new AxeBuilder({ page })
    .include('[role="dialog"]')
    .analyze();

  expect(results.violations).toEqual([]);
});
```

### Storybook Accessibility Addon

```bash
npm install --save-dev @storybook/addon-a11y
```

```javascript
// .storybook/main.js
module.exports = {
  addons: ['@storybook/addon-a11y'],
};
```

Adds an "Accessibility" panel to every story showing axe violations in real-time.

---

## Browser DevTools

### Chrome Accessibility Panel

1. DevTools → Elements → Accessibility tab
2. Shows: computed accessible name, role, state, ARIA attributes
3. Accessibility tree view: full semantic structure as screen readers see it

### Chrome Rendering Panel (Contrast)

1. DevTools → More tools → Rendering
2. "Emulate vision deficiencies": blur, protanopia, deuteranopia, tritanopia, achromatopsia
3. "Emulate CSS media feature prefers-color-scheme"
4. "Emulate CSS media feature prefers-reduced-motion"

### Firefox Accessibility Inspector

1. DevTools → Accessibility tab
2. "Check for issues" button — runs automated checks
3. Tab order visualization (shows numbered focus order overlay)
4. Contrast ratio checker with suggestions

### CSS Overview (Chrome)

1. DevTools → More tools → CSS Overview
2. Shows all color combinations and their contrast ratios
3. Quick way to find all low-contrast pairs

---

## Screen Reader Testing

### VoiceOver (macOS) — Quick Start

| Action | Shortcut |
|--------|----------|
| Turn on/off | Cmd + F5 |
| Move to next element | VO + Right (Ctrl+Opt+Right) |
| Move to previous element | VO + Left |
| Activate element | VO + Space |
| Read current element | VO + A |
| Read all from here | VO + A (hold) |
| Open rotor (landmarks, headings, links) | VO + U |
| Navigate by headings | VO + Cmd + H |
| Navigate by links | VO + Cmd + L |
| Navigate by form controls | VO + Cmd + J |

### NVDA (Windows) — Quick Start

| Action | Shortcut |
|--------|----------|
| Turn on/off | Ctrl + Alt + N |
| Move to next element | Tab or Down Arrow |
| Activate element | Enter or Space |
| Read current line | NVDA + L |
| Navigate by headings | H (browse mode) |
| Navigate by landmarks | D (browse mode) |
| Navigate by forms | F (browse mode) |
| Forms mode | Enter on form element |
| Browse mode | Escape |

### What to Test with Screen Readers

1. **Navigation**: Can you reach all content and controls?
2. **Names**: Are elements announced with correct names?
3. **Roles**: Are custom widgets announced with correct roles?
4. **States**: Do state changes get announced? (expanded, selected, checked)
5. **Live regions**: Are dynamic updates announced appropriately?
6. **Forms**: Are labels, errors, and required states communicated?
7. **Reading order**: Does content make sense when read linearly?
8. **Images**: Are images described meaningfully?

---

## Manual Testing Checklist

### Keyboard Testing (5 minutes)

- [ ] Tab through entire page — all interactive elements reachable
- [ ] Reverse Tab (Shift+Tab) — logical reverse order
- [ ] Enter/Space activate buttons and links
- [ ] Arrow keys work in tabs, menus, radio groups
- [ ] Escape closes modals, dropdowns, tooltips
- [ ] No keyboard traps (can always Tab away)
- [ ] Focus indicator visible on every focused element
- [ ] Skip link works (first Tab shows skip link)
- [ ] Modal focus trapped and restored on close

### Visual Testing (5 minutes)

- [ ] Zoom to 200% — no content lost, no horizontal scroll
- [ ] Set viewport to 320px width — content reflows properly
- [ ] High contrast mode — content still visible
- [ ] Disable CSS — content still readable and in logical order
- [ ] Motion: check for `prefers-reduced-motion` support

### Content Testing (5 minutes)

- [ ] Every page has unique, descriptive `<title>`
- [ ] Exactly one `<h1>` per page
- [ ] Heading hierarchy logical (no skipping h2→h4)
- [ ] Link text is descriptive (not "click here")
- [ ] Error messages are specific and helpful
- [ ] Form labels are visible and associated

### Programmatic Testing (5 minutes)

- [ ] No duplicate IDs (run validator)
- [ ] All `aria-*` attributes are valid
- [ ] `aria-labelledby` / `aria-describedby` point to existing IDs
- [ ] `aria-hidden="true"` not on focusable elements
- [ ] `lang` attribute on `<html>`
- [ ] All `<img>` have `alt` attribute

---

## CI/CD Integration Template

### GitHub Actions

```yaml
name: Accessibility Tests
on: [pull_request]

jobs:
  a11y:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: 20

      - run: npm ci
      - run: npm run build

      # Start server in background
      - run: npm start &
      - run: npx wait-on http://localhost:3000

      # Run axe
      - run: npx @axe-core/cli http://localhost:3000 --tags wcag2a,wcag2aa --exit

      # Run pa11y
      - run: npx pa11y-ci --config .pa11yci.json

      # Run Playwright a11y tests
      - run: npx playwright test tests/a11y/
```

### Pre-commit Hook

```json
// package.json
{
  "lint-staged": {
    "**/*.{tsx,jsx}": ["eslint --plugin jsx-a11y --rule '{\"jsx-a11y/alt-text\": \"error\", \"jsx-a11y/aria-role\": \"error\"}'"]
  }
}
```

---

## Common Automated Testing Gaps

Automated tools catch ~30-57% of WCAG issues. These REQUIRE manual testing:

| What | Why Automation Fails |
|------|---------------------|
| Alt text quality | Tools check presence, not meaningfulness |
| Heading hierarchy purpose | Tools check order, not whether headings describe content |
| Focus order correctness | Tools check focusability, not logical order |
| Keyboard interaction correctness | Tools check Tab access, not widget-specific keys |
| Color meaning | Tools check contrast, not whether color conveys information |
| Content comprehensibility | Tools can't assess reading level or clarity |
| Dynamic content announcements | Tools test static state, not runtime behavior |
| Touch target usability | Tools check size, not usability in context |
| Animation safety | Tools can't always detect problematic animation |
