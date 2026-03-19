---
name: react-nextjs-suite
description: >
  React & Next.js Mastery Suite — complete development toolkit for building production-grade React 19+
  and Next.js 15+ applications. Component architecture with hooks, state management, and compound patterns,
  Next.js App Router with RSC, Server Actions, and streaming SSR, accessible UI components with Tailwind
  and design systems, and React performance optimization with profiling and bundle analysis.
  Triggers: "react component", "react hooks", "react state management", "zustand", "jotai",
  "tanstack query", "react query", "react architecture", "compound component", "custom hook",
  "next.js app router", "nextjs route", "server component", "server action", "rsc", "ssr", "ssg",
  "isr", "nextjs middleware", "nextjs api", "parallel routes", "intercepting routes",
  "ui component", "accessible component", "shadcn", "radix ui", "tailwind component",
  "design system", "design tokens", "framer motion", "react animation", "form component",
  "react hook form", "keyboard navigation", "aria", "wcag",
  "react performance", "react memo", "usememo", "usecallback", "code splitting",
  "react lazy", "bundle size", "virtual scroll", "core web vitals", "lcp", "inp", "cls",
  "react profiler", "re-render", "react optimization".
  Dispatches the appropriate specialist agent: react-architect, nextjs-engineer,
  ui-component-builder, or react-performance-expert.
  NOT for: Backend-only development without React, mobile development (React Native),
  non-React frameworks (Vue, Svelte, Angular), or infrastructure/DevOps.
version: 1.0.0
argument-hint: "<component|nextjs|ui|perf> [target]"
user-invocable: true
allowed-tools: Read, Grep, Glob, Bash
model: sonnet
---

# React & Next.js Mastery Suite

Production-grade React and Next.js development agents for Claude Code. Four specialist agents that handle component architecture, Next.js App Router, accessible UI building, and performance optimization — the complete React development lifecycle.

## Available Agents

### React Architect (`react-architect`)
Designs and builds React applications from the ground up. Component architecture with compound components, custom hooks, state management (Zustand, Jotai, TanStack Query), React 19 features (use, Actions, useOptimistic), error boundaries, context optimization, and TypeScript patterns.

**Invoke**: Dispatch via Task tool with `subagent_type: "react-architect"`.

**Example prompts**:
- "Design the component architecture for my dashboard"
- "Set up Zustand state management with slices"
- "Create a custom hook for infinite scroll"
- "Implement compound components for an accordion"

### Next.js Engineer (`nextjs-engineer`)
Builds production-grade Next.js 15+ applications with App Router. React Server Components, Server Actions, streaming SSR, ISR/on-demand revalidation, parallel routes, intercepting routes, middleware, route handlers, and deployment optimization.

**Invoke**: Dispatch via Task tool with `subagent_type: "nextjs-engineer"`.

**Example prompts**:
- "Set up App Router with authentication and layouts"
- "Implement Server Actions for my forms"
- "Add parallel routes for the dashboard sidebar"
- "Optimize data fetching with proper caching"

### UI Component Builder (`ui-component-builder`)
Creates accessible, reusable UI components and design systems. Radix UI primitives, shadcn/ui patterns, Tailwind CSS design tokens, Framer Motion animations, accessible forms with React Hook Form, keyboard navigation, and dark mode.

**Invoke**: Dispatch via Task tool with `subagent_type: "ui-component-builder"`.

**Example prompts**:
- "Build an accessible combobox component"
- "Set up a design system with Tailwind v4 tokens"
- "Create a data table with sorting and pagination"
- "Add Framer Motion page transitions"

### React Performance Expert (`react-performance-expert`)
Diagnoses and fixes performance issues. React DevTools Profiler analysis, React.memo/useMemo/useCallback optimization, code splitting with React.lazy and Suspense, bundle analysis, virtual scrolling, optimistic updates, and Core Web Vitals optimization.

**Invoke**: Dispatch via Task tool with `subagent_type: "react-performance-expert"`.

**Example prompts**:
- "Find and fix unnecessary re-renders in my app"
- "Implement virtual scrolling for a large list"
- "Optimize my bundle size — it's 500KB"
- "Improve Core Web Vitals scores"

## Quick Start: /react-dev

Use the `/react-dev` command for guided React and Next.js development:

```
/react-dev                          # Auto-detect and suggest improvements
/react-dev component                # React component architecture
/react-dev component --hooks        # Custom hook design
/react-dev component --state        # State management setup
/react-dev nextjs                   # Next.js App Router features
/react-dev nextjs --actions         # Server Actions and forms
/react-dev ui                       # Build accessible UI components
/react-dev ui --design-system       # Design tokens and theme
/react-dev perf                     # Performance optimization
/react-dev perf --bundle            # Bundle size analysis
/react-dev --review                 # Full codebase review
```

The `/react-dev` command auto-detects your framework, discovers your setup, and routes to the right agent.

## Agent Selection Guide

| Need | Agent | Trigger |
|------|-------|---------|
| Component architecture | react-architect | "Design component hierarchy" |
| Custom hooks | react-architect | "Create a custom hook" |
| State management | react-architect | "Set up Zustand/Jotai" |
| React 19 features | react-architect | "Use Actions and useOptimistic" |
| App Router setup | nextjs-engineer | "Set up Next.js routes" |
| Server Actions | nextjs-engineer | "Implement form actions" |
| Data fetching/caching | nextjs-engineer | "Optimize data fetching" |
| Middleware | nextjs-engineer | "Add auth middleware" |
| Accessible components | ui-component-builder | "Build accessible dialog" |
| Design system | ui-component-builder | "Set up design tokens" |
| Form building | ui-component-builder | "Create multi-step form" |
| Animations | ui-component-builder | "Add page transitions" |
| Re-render fixes | react-performance-expert | "Fix re-render issues" |
| Bundle optimization | react-performance-expert | "Reduce bundle size" |
| Core Web Vitals | react-performance-expert | "Improve LCP/INP" |
| Code splitting | react-performance-expert | "Add code splitting" |

## Reference Materials

This skill includes comprehensive reference documents in `references/`:

- **react-patterns.md** — Hooks patterns, state management decision tree, component patterns (compound, polymorphic, render props), context optimization, React 19 features, TypeScript patterns, anti-patterns
- **nextjs-app-router.md** — File conventions, route segments, server vs client components, data fetching strategies, caching layers, revalidation, metadata API, middleware, route handlers, parallel/intercepting routes
- **react-testing-guide.md** — Testing Library query priority, component tests, form tests, hook tests, MSW setup, accessibility testing with axe, integration tests, Vitest configuration

Agents automatically consult these references when working. You can also read them directly for quick answers.

## How It Works

1. You describe what you need (e.g., "build a filterable data table")
2. The SKILL.md routes to the appropriate agent
3. The agent reads your code, discovers your stack and patterns
4. Solutions are designed and implemented following best practices
5. The agent provides results and next steps

All generated artifacts follow industry best practices:
- **React 19+**: Server Components, Actions, use(), useOptimistic, useActionState
- **TypeScript**: Strict typing, generic components, discriminated unions
- **Accessibility**: WAI-ARIA, keyboard navigation, focus management, screen reader support
- **Performance**: Memoization, code splitting, virtual scrolling, optimistic updates
- **Testing**: Testing Library, MSW, accessibility testing, integration tests
