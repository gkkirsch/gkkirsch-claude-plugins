# State Management Patterns & Quick Reference

## Library Comparison

| Feature | Zustand | Redux Toolkit | React Query | Jotai |
|---------|---------|--------------|-------------|-------|
| **Bundle size** | ~1 KB | ~11 KB | ~13 KB | ~3 KB |
| **Boilerplate** | Minimal | Moderate | Minimal | Minimal |
| **Learning curve** | Low | Medium | Low | Low |
| **DevTools** | Redux DevTools | Redux DevTools | React Query DevTools | Jotai DevTools |
| **Provider needed** | No | Yes | Yes | No (optional) |
| **TypeScript** | Excellent | Excellent | Excellent | Excellent |
| **Middleware** | persist, immer, devtools | Built-in (thunk, listener) | N/A | Utilities |
| **SSR support** | Manual | Manual | HydrationBoundary | Manual |
| **React Native** | Yes | Yes | Yes | Yes |
| **Best for** | Simple-medium apps | Large apps, teams | Server state | Atomic/granular |

## When to Use What — Quick Decision

```
Your data comes from an API?
  → React Query (always, for server state)

You need simple shared UI state?
  → Zustand (2-3 stores covers most apps)

You have a large app with 5+ state slices?
  → Redux Toolkit (battle-tested at scale)

You need many independent state atoms?
  → Jotai (bottom-up, granular)

You need forms?
  → React Hook Form + Zod (not a state manager)

You need URL-synced state?
  → nuqs or useSearchParams (not a state manager)
```

## Common Combinations

### Most Apps: Zustand + React Query
```
Zustand:        Auth, UI preferences, feature flags
React Query:    All API data (users, posts, etc.)
React state:    Component-local state (form inputs, toggles)
URL:            Filters, pagination, search
```

### Large Apps: Redux Toolkit + RTK Query
```
RTK slices:     Auth, notifications, UI, complex client state
RTK Query:      All API data with auto-caching
React state:    Component-local state
URL:            Filters, pagination, search
```

### Atomic Apps: Jotai + React Query
```
Jotai atoms:    All client state (granular, composable)
React Query:    API data (or jotai-tanstack-query)
React state:    Trivial component state
```

## State Categorization Checklist

Before choosing a tool, categorize each piece of state:

| State | Category | Tool |
|-------|----------|------|
| User profile from API | Server | React Query |
| List of products from API | Server | React Query |
| Current logged-in user | Client (auth) | Zustand/Redux |
| Sidebar open/closed | Client (UI) | Zustand/useState |
| Dark mode preference | Client (persisted) | Zustand persist / atomWithStorage |
| Form field values | Form | React Hook Form |
| Search query in URL | URL | nuqs / useSearchParams |
| Shopping cart | Client (persisted) | Zustand persist |
| Notification count | Server (polling) | React Query refetchInterval |
| Selected table row | Client (ephemeral) | useState |
| Filters applied to list | URL + Server | nuqs + React Query |

## Performance Patterns

### Prevent Re-renders

```typescript
// Zustand: Always use selectors
const count = useStore((s) => s.count);  // Only re-renders when count changes

// Jotai: Atoms are naturally granular
const count = useAtomValue(countAtom);   // Only re-renders when this atom changes

// Redux: Use typed selectors
const count = useAppSelector((s) => s.counter.count);

// React Query: Returned data is referentially stable
const { data } = useQuery({ queryKey: ['user'], queryFn: fetchUser });
// data doesn't change reference if the content is the same
```

### Selector Memoization

```typescript
// Zustand: useShallow for object/array selectors
import { useShallow } from 'zustand/react/shallow';
const { name, email } = useStore(useShallow((s) => ({ name: s.name, email: s.email })));

// Redux: createSelector for derived data
import { createSelector } from '@reduxjs/toolkit';
const selectActiveUsers = createSelector(
  (state: RootState) => state.users.items,
  (users) => users.filter((u) => u.active)
);

// Jotai: selectAtom for derived slices
import { selectAtom } from 'jotai/utils';
const userNameAtom = selectAtom(userAtom, (user) => user.name);
```

## Migration Paths

### Context + useReducer → Zustand

```typescript
// Before (Context)
const ThemeContext = createContext<ThemeContextType>(null!);
function ThemeProvider({ children }) {
  const [theme, setTheme] = useState<'light' | 'dark'>('light');
  return <ThemeContext.Provider value={{ theme, setTheme }}>{children}</ThemeContext.Provider>;
}
const useTheme = () => useContext(ThemeContext);

// After (Zustand)
const useThemeStore = create<ThemeState>((set) => ({
  theme: 'light' as const,
  setTheme: (theme) => set({ theme }),
}));
// No provider needed. Just import and use.
```

### Redux (legacy) → Redux Toolkit

```typescript
// Before (legacy Redux)
const ADD_TODO = 'ADD_TODO';
function addTodo(text) { return { type: ADD_TODO, payload: text }; }
function todosReducer(state = [], action) {
  switch (action.type) {
    case ADD_TODO: return [...state, { text: action.payload, completed: false }];
    default: return state;
  }
}

// After (RTK)
const todosSlice = createSlice({
  name: 'todos',
  initialState: [],
  reducers: {
    addTodo: (state, action) => { state.push({ text: action.payload, completed: false }); },
  },
});
export const { addTodo } = todosSlice.actions;
```

### useEffect data fetching → React Query

```typescript
// Before (useEffect)
function UserProfile({ userId }) {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    fetch(`/api/users/${userId}`)
      .then((r) => r.json())
      .then((data) => { if (!cancelled) setUser(data); })
      .catch((e) => { if (!cancelled) setError(e); })
      .finally(() => { if (!cancelled) setLoading(false); });
    return () => { cancelled = true; };
  }, [userId]);
  // ...
}

// After (React Query)
function UserProfile({ userId }) {
  const { data: user, isLoading, error } = useQuery({
    queryKey: ['users', userId],
    queryFn: () => fetch(`/api/users/${userId}`).then((r) => r.json()),
  });
  // Automatic caching, deduplication, refetching, error retry
}
```

## Anti-Pattern Reference

| Anti-Pattern | Why It's Bad | Solution |
|-------------|-------------|----------|
| Storing API data in Redux/Zustand | Manual cache management, stale data, loading states | React Query |
| `useEffect` → `setState` for fetching | Race conditions, no caching, manual cleanup | React Query |
| One mega Context for everything | All consumers re-render on every change | Split contexts or use Zustand/Jotai |
| Prop drilling 5+ levels | Fragile, verbose, hard to refactor | Composition first, then Zustand/Context |
| `useEffect` to sync two states | Effect chains, infinite loops, timing bugs | Derive state or use event handlers |
| `useState` + lifting state for 3+ levels | Components become coupled to parent structure | Zustand/Jotai for truly shared state |
| Fetching in every component that needs data | Duplicate requests, inconsistent data | React Query (deduplicates automatically) |
| Global state for form values | Over-engineering, unnecessary re-renders | React Hook Form (form-scoped state) |
