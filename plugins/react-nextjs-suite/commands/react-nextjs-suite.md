---
name: react-dev
description: >
  Quick React & Next.js development command — analyzes your codebase and helps design, build, review, or optimize
  React components, Next.js routes, UI design systems, and performance. Routes to the appropriate specialist agent
  based on your request.
  Triggers: "/react-dev", "react component", "next.js app", "react hooks", "react performance",
  "ui component", "nextjs route", "react architecture", "design system", "tailwind component".
user-invocable: true
argument-hint: "<component|nextjs|ui|perf> [target] [--review] [--audit]"
allowed-tools: Read, Write, Edit, Bash, Glob, Grep
model: sonnet
---

# /react-dev Command

One-command React and Next.js development. Analyzes your project, identifies the framework and patterns, and routes to the appropriate specialist agent for component architecture, Next.js features, UI building, or performance optimization.

## Usage

```
/react-dev                             # Auto-detect and suggest improvements
/react-dev component                   # Design or review React components
/react-dev component --hooks           # Focus on custom hook design
/react-dev component --state           # Focus on state management setup
/react-dev nextjs                      # Next.js App Router features
/react-dev nextjs --actions            # Server Actions and form handling
/react-dev nextjs --routes             # Route design and data fetching
/react-dev ui                          # Build accessible UI components
/react-dev ui --design-system          # Design system and tokens setup
/react-dev ui --forms                  # Complex form components
/react-dev perf                        # Performance audit and optimization
/react-dev perf --bundle               # Bundle size analysis
/react-dev perf --renders              # Re-render optimization
/react-dev --review                    # Full codebase review
```

## Subcommands

### `component` — React Component Architecture

Designs and builds React components with modern patterns, hooks, and state management.

**What it does:**
1. Scans for existing components, hooks, and state management
2. Identifies React version, state library, and styling approach
3. Analyzes component hierarchy and data flow
4. Suggests or generates improvements

**Flags:**
- `--hooks` — Focus on custom hook design and composition
- `--state` — State management setup (Zustand, Jotai, TanStack Query)
- `--patterns` — Apply patterns (compound components, render props, polymorphic)
- `--review` — Review existing components for best practices

**Routes to:** `react-architect` agent

### `nextjs` — Next.js App Router

Builds and optimizes Next.js 15+ applications with App Router, RSC, and Server Actions.

**What it does:**
1. Analyzes `app/` directory, layouts, and route structure
2. Reviews Server vs Client component boundaries
3. Checks data fetching and caching strategies
4. Suggests or generates improvements

**Flags:**
- `--actions` — Server Actions and form handling
- `--routes` — Route design, parallel routes, intercepting routes
- `--data` — Data fetching and caching optimization
- `--middleware` — Middleware and authentication setup

**Routes to:** `nextjs-engineer` agent

### `ui` — UI Component Building

Creates accessible, reusable UI components with proper ARIA, keyboard navigation, and design tokens.

**What it does:**
1. Scans for existing UI library (shadcn/ui, Radix, Headless UI)
2. Reviews accessibility and ARIA usage
3. Checks design token consistency
4. Suggests or generates components

**Flags:**
- `--design-system` — Set up design tokens and theme system
- `--forms` — Build complex accessible forms
- `--animation` — Add Framer Motion or CSS animations
- `--a11y` — Accessibility audit and fixes

**Routes to:** `ui-component-builder` agent

### `perf` — Performance Optimization

Diagnoses and fixes performance issues in React and Next.js applications.

**What it does:**
1. Analyzes bundle size and heavy dependencies
2. Reviews re-render patterns and memoization
3. Checks Core Web Vitals optimization
4. Provides prioritized recommendations

**Flags:**
- `--bundle` — Bundle size analysis and code splitting
- `--renders` — Re-render detection and prevention
- `--vitals` — Core Web Vitals (LCP, INP, CLS) optimization
- `--audit` — Full performance audit

**Routes to:** `react-performance-expert` agent

## Auto-Detection

When no subcommand is specified, `/react-dev` auto-detects your setup:

1. **Next.js detected** (`next.config.*`, `app/` directory) → Routes to `nextjs-engineer`
2. **React with performance issue** (large bundle, slow renders) → Routes to `react-performance-expert`
3. **UI library detected** (shadcn, Radix, Tailwind components) → Routes to `ui-component-builder`
4. **Generic React** → Routes to `react-architect`

## Agent Selection Guide

| Need | Agent | Command |
|------|-------|---------|
| Component architecture | react-architect | `/react-dev component` |
| Custom hooks | react-architect | `/react-dev component --hooks` |
| State management | react-architect | `/react-dev component --state` |
| React 19 features | react-architect | `/react-dev component --patterns` |
| App Router setup | nextjs-engineer | `/react-dev nextjs` |
| Server Actions | nextjs-engineer | `/react-dev nextjs --actions` |
| Data fetching | nextjs-engineer | `/react-dev nextjs --data` |
| Middleware | nextjs-engineer | `/react-dev nextjs --middleware` |
| Accessible components | ui-component-builder | `/react-dev ui` |
| Design system | ui-component-builder | `/react-dev ui --design-system` |
| Form building | ui-component-builder | `/react-dev ui --forms` |
| Animations | ui-component-builder | `/react-dev ui --animation` |
| Bundle optimization | react-performance-expert | `/react-dev perf --bundle` |
| Re-render fixes | react-performance-expert | `/react-dev perf --renders` |
| Core Web Vitals | react-performance-expert | `/react-dev perf --vitals` |
| Full review | All agents | `/react-dev --review` |

## Reference Materials

This suite includes comprehensive reference documents in `references/`:

- **react-patterns.md** — Hooks patterns, state management approaches, context optimization, error boundaries, compound components, TypeScript patterns
- **nextjs-app-router.md** — Routing conventions, layouts, loading states, parallel and intercepting routes, data fetching, caching, ISR, middleware
- **react-testing-guide.md** — Testing Library, Vitest, MSW, component tests, hook tests, integration tests, accessibility testing

Agents automatically consult these references when working. You can also read them directly for quick answers.

## How It Works

1. You describe what you need (e.g., "build a data table component with sorting")
2. The command analyzes your project structure and existing React/Next.js code
3. It routes to the appropriate specialist agent
4. The agent reads your code, understands your patterns, and generates solutions
5. Code is written following best practices with proper patterns

All generated code follows these principles:
- **React 19+**: Server Components, Actions, use(), useOptimistic, useActionState
- **TypeScript**: Strict typing, generic components, discriminated unions
- **Accessibility**: ARIA attributes, keyboard navigation, focus management, screen reader support
- **Performance**: Memoization where needed, code splitting, virtual scrolling, optimistic updates
- **Testing**: Testing Library patterns, MSW for API mocking, accessibility testing
