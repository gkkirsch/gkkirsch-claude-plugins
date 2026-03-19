---
name: frontend-audit
description: >
  Quick frontend audit command — analyzes your codebase for performance bottlenecks, accessibility violations,
  responsive design issues, and bundle size problems. Routes to specialist agents based on the audit type.
  Triggers: "/frontend-audit", "audit frontend", "check performance", "check accessibility",
  "audit a11y", "check responsive", "audit bundle size", "lighthouse audit", "web vitals",
  "WCAG check", "performance review", "accessibility scan".
user-invocable: true
argument-hint: "<perf|a11y|responsive|bundle> [path] [--target mobile|desktop]"
allowed-tools: Read, Write, Edit, Bash, Glob, Grep
model: sonnet
---

# /frontend-audit Command

One-command frontend analysis for any web application. Detects the tech stack, identifies the audit focus area, and dispatches the appropriate specialist agent for deep analysis.

## Usage

```
/frontend-audit                                      # Full audit (all areas)
/frontend-audit perf                                 # Performance audit only
/frontend-audit perf src/                            # Performance audit of specific path
/frontend-audit perf --target mobile                 # Mobile performance focus
/frontend-audit a11y                                 # Accessibility audit only
/frontend-audit a11y src/components/                 # Audit specific components
/frontend-audit responsive                           # Responsive design audit
/frontend-audit responsive --target mobile           # Mobile responsive focus
/frontend-audit bundle                               # Bundle size optimization
/frontend-audit bundle --target 200kb                # Bundle with specific budget
```

## Subcommands

| Subcommand | Agent | Description |
|------------|-------|-------------|
| `perf` | performance-auditor | Core Web Vitals, Lighthouse, rendering, paint optimization |
| `a11y` | accessibility-checker | WCAG 2.1/2.2 compliance, ARIA, keyboard nav, screen readers |
| `responsive` | responsive-designer | Breakpoints, fluid typography, container queries, mobile-first |
| `bundle` | bundle-optimizer | Webpack/Vite optimization, tree-shaking, code splitting |

## Procedure

### Step 1: Detect Technology Stack

Read the project root to understand the codebase:

1. **Read** `package.json` — determine language, framework, dependencies
2. **Glob** for build configuration files:
   - `vite.config.*`, `webpack.config.*`, `next.config.*`, `nuxt.config.*`, `astro.config.*`
   - `tailwind.config.*`, `postcss.config.*`
   - `tsconfig.json`, `.babelrc`
3. **Glob** for entry points:
   - `index.html`, `src/main.*`, `src/App.*`, `app/layout.*`, `pages/_app.*`
4. **Identify** the core stack: framework, bundler, CSS approach, component library

### Step 2: Route to Specialist

Based on the subcommand (or all areas for full audit):

**For `perf`** — Dispatch `performance-auditor`:
- Analyze Core Web Vitals (LCP, INP, CLS)
- Check resource loading (images, fonts, CSS, JS)
- Check render-blocking resources
- Check third-party script impact
- Report with Lighthouse score estimate

**For `a11y`** — Dispatch `accessibility-checker`:
- Run automated scan (search for violations)
- Check semantic HTML structure
- Check ARIA implementation
- Check keyboard navigation patterns
- Check color contrast
- Report with WCAG conformance level

**For `responsive`** — Dispatch `responsive-designer`:
- Catalog breakpoints and media queries
- Check for responsive layout patterns
- Check for mobile-first approach
- Check responsive images
- Check touch target sizes
- Report with viewport compatibility assessment

**For `bundle`** — Dispatch `bundle-optimizer`:
- Analyze bundle size and composition
- Check code splitting strategy
- Check tree-shaking effectiveness
- Identify heavy dependencies
- Report with size reduction opportunities

### Step 3: Generate Report

Combine findings into a prioritized report:

```markdown
# Frontend Audit Report

## Tech Stack
- Framework: [detected]
- Bundler: [detected]
- CSS: [detected]
- Components: [detected]

## Score Summary
| Area | Score | Status |
|------|-------|--------|
| Performance | X/100 | ✅/⚠️/❌ |
| Accessibility | X% WCAG AA | ✅/⚠️/❌ |
| Responsive | X/10 | ✅/⚠️/❌ |
| Bundle Size | XKB | ✅/⚠️/❌ |

## Critical Issues (Fix Immediately)
1. [Issue] — Impact: [who/what affected]
   Fix: [specific code change]

## Important Issues (Fix Soon)
1. ...

## Recommendations (Nice to Have)
1. ...

## Quick Wins
- [ ] [Action] — 5 min effort, high impact
- [ ] [Action] — 10 min effort, medium impact
```

### Step 4: Provide Actionable Fixes

For each issue found:
1. Explain what's wrong and why it matters
2. Show the current problematic code
3. Show the fixed code
4. Estimate the impact of the fix

## Arguments

| Argument | Description |
|----------|-------------|
| `perf` | Performance audit (Core Web Vitals, Lighthouse) |
| `a11y` | Accessibility audit (WCAG 2.1/2.2) |
| `responsive` | Responsive design audit |
| `bundle` | Bundle size optimization |
| `[path]` | Scope audit to a specific directory |
| `--target mobile` | Focus on mobile-specific issues |
| `--target desktop` | Focus on desktop-specific issues |
| `--target <size>` | Bundle budget target (e.g., `200kb`) |

## No Arguments (Full Audit)

When run without arguments, perform a quick scan across all areas:

1. **Performance**: Check top 5 performance issues (LCP image, render-blocking resources, bundle size, third-party scripts, image optimization)
2. **Accessibility**: Check top 5 a11y issues (alt text, form labels, color contrast, focus visibility, heading hierarchy)
3. **Responsive**: Check top 5 responsive issues (viewport meta, images max-width, breakpoints, touch targets, horizontal overflow)
4. **Bundle**: Check top 5 bundle issues (total size, largest dependencies, code splitting, unused CSS, compression)

Provide a summary with the most impactful fixes across all areas.
