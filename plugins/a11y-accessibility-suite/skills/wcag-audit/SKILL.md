---
name: wcag-audit
description: >
  Run a comprehensive WCAG 2.2 AA accessibility audit on your web application.
  Scans HTML structure, ARIA usage, keyboard navigation, color contrast, and generates
  a prioritized fix list. Works with React, Vue, Svelte, Next.js, and vanilla HTML.
  Triggers: "accessibility audit", "a11y audit", "WCAG check", "check accessibility".
  NOT for: PDF accessibility, native mobile apps.
version: 1.0.0
argument-hint: "[file or directory to audit]"
allowed-tools: Read, Grep, Glob, Bash
---

# WCAG 2.2 AA Audit

Run a systematic accessibility audit following WCAG 2.2 AA criteria.

## Step 1: Scope the Audit

Identify what to audit. If an argument was provided, audit that file/directory. Otherwise, audit the entire project.

```
Glob: **/*.{html,tsx,jsx,vue,svelte}
```

## Step 2: Run Automated Checks

### 2a. Check for a11y linting

```bash
# Check if jsx-a11y or equivalent is installed
grep -r "jsx-a11y\|eslint-plugin-vuejs-accessibility" package.json
```

If installed, run the linter first to catch low-hanging fruit:
```bash
npx eslint --no-eslintrc --plugin jsx-a11y --rule '{"jsx-a11y/alt-text": "error", "jsx-a11y/anchor-has-content": "error", "jsx-a11y/aria-props": "error", "jsx-a11y/aria-role": "error", "jsx-a11y/role-has-required-aria-props": "error"}' src/
```

### 2b. Check for axe-core

```bash
grep -r "axe-core\|@axe-core" package.json
```

## Step 3: Manual Code Review

For each file, systematically check:

### Images and Icons
```
Grep: "<img " — check each for alt attribute
Grep: "role=\"img\"" — check for aria-label
Grep: "<svg" — check for <title> or aria-label
Grep: "<Icon\|<FontAwesome\|<Lucide" — check icon components have labels
```

### Headings
```
Grep: "<h[1-6]|<Heading" — verify hierarchy (no skipping levels)
```
There should be exactly ONE h1 per page/route.

### Forms
```
Grep: "<input\|<select\|<textarea" — verify each has an associated <label> or aria-label
Grep: "placeholder=" — placeholder is NOT a substitute for label
Grep: "required\|aria-required" — required fields properly marked
Grep: "aria-describedby\|aria-errormessage" — error messages linked to fields
```

### Interactive Elements
```
Grep: "onClick\|@click\|on:click" — verify element is a <button> or <a>, not a <div>/<span>
Grep: "tabIndex\|tabindex" — check for correct usage (0 or -1, rarely positive)
Grep: "onMouseOver\|onMouseEnter\|@mouseenter" — verify keyboard equivalent exists
Grep: "outline:\s*none\|outline:\s*0" — focus visibility removed?
```

### ARIA Usage
```
Grep: "aria-hidden=\"true\"" — verify no focusable children
Grep: "role=\"" — verify roles are correct and necessary
Grep: "aria-live\|role=\"alert\"\|role=\"status\"" — live regions present
Grep: "aria-expanded" — disclosure/accordion patterns correct
Grep: "aria-modal" — dialog patterns correct
```

### Color and Contrast
```
Grep: "color:\|background-color:\|bg-" — flag low-contrast combinations
```
Common failures:
- `#999` on `#fff` (2.85:1 — fails 4.5:1)
- `#777` on `#fff` (4.48:1 — fails 4.5:1)
- `#767676` on `#fff` (4.54:1 — barely passes)

### Keyboard Navigation
```
Grep: "onKeyDown\|onKeyUp\|@keydown\|@keyup" — keyboard handlers present
Grep: "focus()\|focus-trap\|useFocusTrap" — focus management present
Grep: "e.key === \|e.keyCode\|key ===" — specific key handling
```

### Language
```
Grep: "lang=" in root HTML file — page language declared
```

### Skip Navigation
```
Grep: "skip.*nav\|skip.*main\|skip.*content" — skip link present
```

## Step 4: Generate Report

Organize findings by severity:

### Critical (WCAG A — Legal Risk)
- Missing alt text on meaningful images
- Keyboard traps
- Missing form labels
- No page language
- Interactive divs/spans without keyboard access

### Serious (WCAG AA — Standard Compliance)
- Insufficient color contrast
- Missing focus indicators
- Missing skip navigation
- Non-descriptive link text
- Missing error identification

### Moderate (Best Practices)
- Missing landmark regions
- Suboptimal heading hierarchy
- Missing aria-live for dynamic content
- Placeholder used as label

### Report Format

```markdown
# Accessibility Audit Report

## Score: X/100
- Critical: N issues
- Serious: N issues
- Moderate: N issues
- Passed: N checks

## Issues

### [CRITICAL] Missing alt text on images
**WCAG**: 1.1.1 Non-text Content (A)
**File**: src/components/ProductCard.tsx:24
**Current**: `<img src={product.image} />`
**Fix**: `<img src={product.image} alt={product.name} />`

...

## Automated Testing Setup

Add these to prevent regressions:
[Include specific package installs and config]
```

## Step 5: Fix Issues

After presenting the report, offer to fix issues automatically. Start with Critical, then Serious. Use Edit tool to apply fixes directly.