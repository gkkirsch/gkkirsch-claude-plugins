---
name: state-architect
description: >
  Helps choose the right state management solution and design state architecture.
  Triggers: "state management", "which state library", "state architecture",
  "global state", "client state vs server state".
  NOT for: specific library implementation (use zustand, redux-toolkit, react-query, or jotai skills).
tools: Read, Glob, Grep
---

# State Architecture Guide

## The State Categories

Every piece of state falls into one of these categories. Identify the category first, then pick the tool.

| Category | What It Is | Examples | Best Tool |
|----------|-----------|----------|-----------|
| **Server state** | Data from an API/database | Users list, product details, notifications | React Query / TanStack Query |
| **Client state** | UI state that doesn't come from a server | Modal open/close, sidebar collapsed, theme | Zustand, Jotai, or React state |
| **Form state** | Input values, validation, dirty tracking | Login form, checkout, settings | React Hook Form + Zod |
| **URL state** | State encoded in the URL | Filters, pagination, search query | nuqs, useSearchParams |
| **Derived state** | Computed from other state | Filtered list, totals, formatted dates | useMemo, Zustand selectors |

## Decision Matrix: Which Library?

```
How much state do you have?
├── Just a few values shared between components
│   ├── Simple (boolean, string) → React Context + useReducer
│   └── Object with actions → Zustand (simplest global state)
├── Moderate state with several slices
│   ├── Mostly server data → React Query (it IS your state manager)
│   └── Mostly client state → Zustand
├── Complex app state (many slices, middleware, devtools)
│   └── Redux Toolkit (RTK)
└── Atomic/granular state (many independent pieces)
    └── Jotai
```

### When to Use Each

**Zustand** (most projects):
- Simple API (no boilerplate, no providers)
- Small to medium apps
- When you want `useState` but global
- 1-3 stores covering your whole app
- Great devtools, persist middleware, immer built-in

**Redux Toolkit** (large teams/apps):
- Large apps with 5+ state slices
- Need middleware (thunks, sagas, listeners)
- Team already knows Redux
- Complex state transitions
- Need time-travel debugging
- RTK Query for API caching (alternative to React Query)

**React Query / TanStack Query** (any app with an API):
- ANY app that fetches data from a server
- Eliminates 80% of "global state" (it was server state all along)
- Automatic caching, deduplication, background refetching
- Optimistic updates, infinite scroll, pagination
- Use alongside Zustand or Jotai for remaining client state

**Jotai** (atomic patterns):
- Many independent pieces of state
- Bottom-up state design (atoms compose into molecules)
- When Context causes too many re-renders
- Derived state is central to your app
- Good for complex forms or dashboards

### What NOT to Use

- **Plain React Context** for frequently changing state (causes re-renders in all consumers)
- **Redux** without Toolkit (too much boilerplate)
- **MobX** for new projects (smaller ecosystem, less hiring pool)
- **Recoil** (Meta abandoned it, use Jotai instead)

## State Architecture Patterns

### Pattern 1: Zustand + React Query (Most Common)

```
Server State: React Query
  └── API data, caching, background sync

Client State: Zustand
  └── UI state, preferences, auth status

Form State: React Hook Form
  └── Form values, validation, submission

URL State: nuqs or useSearchParams
  └── Filters, pagination, search
```

### Pattern 2: Redux Toolkit Only (Large Apps)

```
Server State: RTK Query
  └── API caching, auto-generated hooks

Client State: Redux Toolkit slices
  └── Auth, UI, feature flags, notifications

Side Effects: Listener middleware
  └── Complex async workflows

Form State: React Hook Form (external)
  └── Forms are still better outside Redux
```

### Pattern 3: Jotai + React Query (Atomic)

```
Server State: React Query (or Jotai + async atoms)
  └── API data with automatic caching

Client State: Jotai atoms
  └── Each piece of state is an atom
  └── Derived atoms compute from other atoms

URL State: jotai-location
  └── URL params as atoms
```

## State Design Rules

1. **Start with React Query.** If data comes from an API, it's server state. Don't put it in Redux/Zustand. React Query handles caching, loading states, error states, refetching, and deduplication.

2. **Don't duplicate server state.** If React Query already caches `users`, don't copy it into a Zustand store. Access it via `useQuery` wherever you need it.

3. **Keep state close to where it's used.** If only one component needs it, use `useState`. If a subtree needs it, use composition or context. Only elevate to global state when truly needed.

4. **Normalize complex relational data.** If entities reference each other (users have posts, posts have comments), normalize into flat maps keyed by ID. Redux Toolkit has `createEntityAdapter` for this.

5. **Separate concerns into stores/slices.** Don't put everything in one mega-store. Split by domain: `useAuthStore`, `useUIStore`, `useCartStore`.

6. **Derive, don't store.** If you can compute it from existing state, compute it. Don't store `filteredUsers` — derive it from `users` + `filter` with a selector or `useMemo`.

7. **URL is state too.** Filters, pagination, search queries, selected tabs — put these in the URL. Users can bookmark and share. Use `nuqs` for type-safe URL state.

## Anti-Patterns to Avoid

| Anti-Pattern | Problem | Fix |
|-------------|---------|-----|
| Storing server data in Redux | Duplicated state, stale data, manual cache invalidation | Use React Query |
| Single global store for everything | God object, tightly coupled, excessive re-renders | Split into domain-specific stores |
| Storing derived values | Can get out of sync with source | Use selectors or useMemo |
| useEffect to sync state | Effect chains, infinite loops, race conditions | Derive state or use event handlers |
| Context for frequently changing state | Every consumer re-renders on every change | Use Zustand/Jotai (subscription-based) |
| Prop drilling avoidance via global state | Over-engineering, everything becomes global | Use composition, then Context, then global |
