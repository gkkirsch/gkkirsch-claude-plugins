---
name: frontend-performance-suite
description: >
  Frontend Performance & Accessibility Suite — AI-powered toolkit for Core Web Vitals optimization,
  WCAG accessibility compliance, responsive design auditing, and bundle size optimization. Analyzes
  LCP, INP, CLS, Lighthouse scores, ARIA patterns, keyboard navigation, responsive layouts,
  container queries, webpack/Vite configurations, tree-shaking, and code splitting.
  Triggers: "frontend audit", "performance audit", "check web vitals", "lighthouse audit",
  "core web vitals", "LCP optimization", "INP optimization", "CLS fix",
  "accessibility audit", "a11y check", "WCAG compliance", "screen reader",
  "ARIA review", "keyboard navigation", "color contrast", "accessibility scan",
  "responsive audit", "responsive design", "mobile-first", "breakpoints",
  "container queries", "fluid typography",
  "bundle audit", "bundle size", "code splitting", "tree-shaking",
  "bundle optimization", "webpack optimization", "vite optimization",
  "reduce bundle size", "lazy loading", "performance budget".
  Dispatches the appropriate specialist agent: performance-auditor, accessibility-checker,
  responsive-designer, or bundle-optimizer.
  NOT for: Backend performance, database optimization, server configuration,
  network infrastructure, API response times, or server-side rendering configuration
  (except as it relates to frontend hydration and loading performance).
version: 1.0.0
argument-hint: "<perf|a11y|responsive|bundle> [path-or-scope]"
user-invocable: true
allowed-tools: Read, Grep, Glob, Bash
model: sonnet
---

# Frontend Performance & Accessibility Suite

Production-grade frontend optimization and accessibility compliance agents for Claude Code. Four specialist agents that handle Core Web Vitals auditing, WCAG accessibility compliance, responsive design patterns, and bundle size optimization — the frontend quality work that every production web application needs.

## Available Agents

### Performance Auditor (`performance-auditor`)
Analyzes Core Web Vitals (LCP, INP, CLS), Lighthouse performance scores, rendering pipeline efficiency, resource loading patterns, and third-party script impact. Provides specific optimization recommendations with expected metric improvements and code examples for React, Vue, Next.js, Nuxt, and vanilla JavaScript applications.

**Invoke**: Dispatch via Task tool with `subagent_type: "performance-auditor"`.

**Example prompts**:
- "Audit my site's Core Web Vitals"
- "Why is my LCP score so high?"
- "Optimize my Lighthouse performance score"
- "Find what's causing layout shifts on my page"
- "Reduce my Time to Interactive"
- "Analyze third-party script impact on my site"

### Accessibility Checker (`accessibility-checker`)
Audits WCAG 2.1/2.2 Level A and AA compliance, ARIA implementation correctness, keyboard navigation, screen reader compatibility, color contrast ratios, form accessibility, and dynamic content announcements. Covers semantic HTML, focus management, live regions, and assistive technology testing procedures.

**Invoke**: Dispatch via Task tool with `subagent_type: "accessibility-checker"`.

**Example prompts**:
- "Check my site for accessibility violations"
- "Audit WCAG compliance for my React app"
- "Review my ARIA implementation in these components"
- "Check color contrast across my application"
- "Are my forms accessible to screen readers?"
- "Review keyboard navigation in my modal component"

### Responsive Designer (`responsive-designer`)
Analyzes and implements responsive layouts using CSS Grid, Flexbox, container queries, fluid typography with clamp(), responsive images, and mobile-first breakpoint strategies. Covers cross-device compatibility from 320px mobile viewports to 4K desktop displays.

**Invoke**: Dispatch via Task tool with `subagent_type: "responsive-designer"`.

**Example prompts**:
- "Audit my site's responsive design"
- "My layout breaks on mobile — help me fix it"
- "Implement container queries for my card component"
- "Set up a fluid typography system"
- "Convert my desktop-first layout to mobile-first"
- "Review my breakpoint strategy"

### Bundle Optimizer (`bundle-optimizer`)
Analyzes webpack, Vite, Rollup, and other bundler configurations for size optimization. Covers code splitting strategies, tree-shaking effectiveness, dependency analysis, lazy loading patterns, CSS optimization, compression configuration, and caching strategies.

**Invoke**: Dispatch via Task tool with `subagent_type: "bundle-optimizer"`.

**Example prompts**:
- "Analyze and reduce my bundle size"
- "Set up code splitting for my React app"
- "Find heavy dependencies I can replace"
- "Optimize my Vite build configuration"
- "Why is my JavaScript bundle so large?"
- "Set up tree-shaking for my project"

## Routing Logic

When invoked, determine which specialist to dispatch based on the user's request:

### Performance Keywords → `performance-auditor`
- "performance", "web vitals", "core web vitals", "LCP", "INP", "CLS", "FCP", "TTFB"
- "lighthouse", "page speed", "loading speed", "render", "paint"
- "slow", "speed", "fast", "optimize performance"
- "third-party scripts", "blocking resources"

### Accessibility Keywords → `accessibility-checker`
- "accessibility", "a11y", "WCAG", "ADA", "Section 508"
- "screen reader", "ARIA", "keyboard", "focus", "tab order"
- "color contrast", "alt text", "form labels", "semantic HTML"
- "assistive technology", "VoiceOver", "NVDA", "JAWS"

### Responsive Keywords → `responsive-designer`
- "responsive", "mobile", "breakpoint", "media query"
- "container query", "fluid", "layout", "grid", "flexbox"
- "viewport", "tablet", "desktop", "mobile-first"
- "touch target", "orientation"

### Bundle Keywords → `bundle-optimizer`
- "bundle", "bundle size", "code splitting", "tree-shaking"
- "webpack", "vite", "rollup", "esbuild"
- "lazy loading", "dynamic import", "chunk"
- "dependency", "node_modules", "package size"

### Ambiguous or Full Audit → Run Quick Scan
If the request doesn't clearly match one specialist, or the user asks for a "full audit" or "frontend audit":

1. Quick performance scan (top 5 issues)
2. Quick accessibility scan (top 5 issues)
3. Quick responsive scan (top 5 issues)
4. Quick bundle scan (top 5 issues)
5. Prioritized combined report

## Quick Start

```
# Full frontend audit
/frontend-performance-suite

# Performance only
/frontend-performance-suite perf

# Accessibility only
/frontend-performance-suite a11y

# Responsive design
/frontend-performance-suite responsive

# Bundle optimization
/frontend-performance-suite bundle

# Scoped to specific path
/frontend-performance-suite a11y src/components/

# Mobile-focused
/frontend-performance-suite perf --target mobile
```

## References

The suite includes comprehensive reference material:

- **Core Web Vitals** (`references/core-web-vitals.md`): Complete LCP, INP, CLS measurement and optimization guide
- **WCAG Compliance** (`references/wcag-compliance.md`): Full WCAG 2.1/2.2 criteria with implementation examples
- **CSS Performance** (`references/css-performance.md`): Containment, layers, animation optimization, selector efficiency
