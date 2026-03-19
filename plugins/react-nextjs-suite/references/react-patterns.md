# React Patterns Reference

Quick-reference guide for React 19+ patterns, hooks, state management, and TypeScript best practices. Consult this when building or reviewing React applications.

---

## Hooks Rules & Patterns

### Rules of Hooks
1. Only call hooks at the top level — not inside conditions, loops, or nested functions
2. Only call hooks from React function components or custom hooks
3. Exception: `use()` in React 19 CAN be called conditionally

### Common Hook Patterns

```tsx
// useState with lazy initialization
const [data, setData] = useState(() => expensiveComputation());

// useState with updater function (when next state depends on previous)
setCount(prev => prev + 1);  // GOOD
setCount(count + 1);         // BAD if called multiple times

// useReducer for complex state
type Action =
  | { type: 'ADD_TODO'; text: string }
  | { type: 'TOGGLE_TODO'; id: string }
  | { type: 'DELETE_TODO'; id: string }
  | { type: 'SET_FILTER'; filter: Filter };

function todosReducer(state: TodoState, action: Action): TodoState {
  switch (action.type) {
    case 'ADD_TODO':
      return {
        ...state,
        todos: [...state.todos, { id: crypto.randomUUID(), text: action.text, completed: false }],
      };
    case 'TOGGLE_TODO':
      return {
        ...state,
        todos: state.todos.map(t =>
          t.id === action.id ? { ...t, completed: !t.completed } : t
        ),
      };
    case 'DELETE_TODO':
      return { ...state, todos: state.todos.filter(t => t.id !== action.id) };
    case 'SET_FILTER':
      return { ...state, filter: action.filter };
  }
}

// useRef for mutable values that don't trigger re-renders
const renderCount = useRef(0);
renderCount.current += 1; // Incrementing doesn't cause re-render

// useRef for DOM elements
const inputRef = useRef<HTMLInputElement>(null);
inputRef.current?.focus();

// useId for accessible form labels
const id = useId();
<label htmlFor={`${id}-email`}>Email</label>
<input id={`${id}-email`} type="email" />

// useSyncExternalStore for subscribing to external stores
const width = useSyncExternalStore(
  (callback) => {
    window.addEventListener('resize', callback);
    return () => window.removeEventListener('resize', callback);
  },
  () => window.innerWidth,
  () => 1024 // Server snapshot
);
```

### Custom Hook Conventions

```tsx
// Naming: always prefix with "use"
function useWindowSize() { ... }
function useLocalStorage<T>(key: string, initial: T) { ... }
function useDebounce<T>(value: T, delay: number) { ... }

// Return conventions:
// Single value → return value directly
function useMediaQuery(query: string): boolean { ... }

// Two values (state + setter) → return tuple
function useToggle(initial: boolean): [boolean, () => void] { ... }

// Multiple values → return object
function useForm<T>(config: FormConfig<T>): {
  values: T;
  errors: Record<string, string>;
  handleChange: ChangeHandler;
  handleSubmit: SubmitHandler;
  isSubmitting: boolean;
  reset: () => void;
} { ... }
```

---

## State Management Decision Tree

```
Need state? →
  ├─ Local to one component? → useState / useReducer
  ├─ Shared between a few nearby components? → Lift state up
  ├─ Shared across distant components?
  │   ├─ Updates rarely? → React Context
  │   ├─ Updates frequently? → Zustand / Jotai
  │   └─ Complex with many actions? → Zustand with slices
  ├─ Server/API data? → TanStack Query / SWR
  ├─ URL state (search, filters, pagination)? → nuqs / useSearchParams
  └─ Form state? → React Hook Form / useActionState
```

### When to Use What

| Library | Best For | Avoid When |
|---------|----------|------------|
| **useState** | Component-local state | Complex state logic, shared state |
| **useReducer** | Complex state with many transitions | Simple toggle/counter state |
| **Context** | Theme, auth, locale (rarely-changing) | High-frequency updates (causes re-renders) |
| **Zustand** | Global client state, cross-component | Server state (use TanStack Query) |
| **Jotai** | Fine-grained atomic state | Simple apps (overhead not justified) |
| **TanStack Query** | Server state, caching, optimistic updates | Pure client state |
| **nuqs** | URL-synced state (filters, search) | Non-URL state |
| **React Hook Form** | Complex forms with validation | Simple single-field forms |

---

## Component Patterns Summary

### Compound Components
```tsx
<Select>
  <Select.Trigger>Choose...</Select.Trigger>
  <Select.Content>
    <Select.Item value="a">Option A</Select.Item>
    <Select.Item value="b">Option B</Select.Item>
  </Select.Content>
</Select>
```
**Use when:** Components share implicit state and belong together.

### Polymorphic Components
```tsx
<Button as="a" href="/about">Link styled as button</Button>
<Box as="section" className="...">Renders as <section></Box>
```
**Use when:** Component needs to render as different HTML elements.

### Render Props
```tsx
<DataFetcher url="/api/data">
  {({ data, isLoading }) => isLoading ? <Spinner /> : <List data={data} />}
</DataFetcher>
```
**Use when:** Consumer needs maximum flexibility over rendering.

### HOC (Higher-Order Component)
```tsx
const ProtectedPage = withAuth(DashboardPage);
```
**Use when:** Cross-cutting concern applied to many components (rare in modern React — prefer hooks).

### Controlled vs Uncontrolled
```tsx
// Controlled: parent owns the state
<Input value={email} onChange={setEmail} />

// Uncontrolled: component owns its own state
<Input defaultValue="" ref={inputRef} />
```
**Controlled when:** You need to validate, transform, or sync state. **Uncontrolled when:** You just need the final value (forms with React Hook Form).

---

## Error Handling Patterns

### Error Boundary Placement Strategy

```
App
├── ErrorBoundary (global — catches unhandled errors, shows fallback page)
│   ├── Layout
│   │   ├── ErrorBoundary (page-level — resets on navigation)
│   │   │   └── Page Content
│   │   │       ├── ErrorBoundary (widget — isolated failure)
│   │   │       │   └── Widget A
│   │   │       ├── ErrorBoundary (widget — isolated failure)
│   │   │       │   └── Widget B
│   │   │       └── Widget C (no boundary — error bubbles to page)
```

### Error Types and Handling

| Error Type | Handling Strategy |
|------------|-------------------|
| Render error | Error Boundary with retry |
| Async error in event handler | try/catch + toast notification |
| Server Action error | useActionState with error state |
| API error | TanStack Query error + retry |
| Network error | Offline detection + retry queue |
| 404 | `notFound()` in Next.js / custom fallback |

---

## Context Optimization Patterns

### Split by Update Frequency

```tsx
// SLOW: One context for everything
<AppContext.Provider value={{ user, theme, notifications, cart }}>

// FAST: Split contexts
<UserContext.Provider value={user}>
  <ThemeContext.Provider value={{ theme, setTheme }}>
    <NotificationContext.Provider value={notifications}>
      <CartContext.Provider value={cart}>
```

### Separate State from Actions

```tsx
const CountStateContext = createContext(0);
const CountDispatchContext = createContext<Dispatch>(() => {});

// Components that only dispatch don't re-render when count changes
// Components that only read count don't re-render when dispatch reference changes
```

### Context + useRef for Stable Callbacks

```tsx
function CallbackProvider({ children }: { children: ReactNode }) {
  const [state, setState] = useState(initialState);
  const stateRef = useRef(state);
  stateRef.current = state;

  // Stable reference — never changes
  const actions = useMemo(() => ({
    increment: () => setState(s => s + 1),
    getState: () => stateRef.current,
  }), []);

  return (
    <StateContext value={state}>
      <ActionsContext value={actions}>
        {children}
      </ActionsContext>
    </StateContext>
  );
}
```

---

## React 19 Quick Reference

| Feature | Description | Usage |
|---------|-------------|-------|
| `use()` | Read promise or context (can be conditional) | `const data = use(promise)` |
| `useActionState` | Form state from async action | `const [state, action, pending] = useActionState(fn, init)` |
| `useFormStatus` | Get pending status inside `<form>` | `const { pending } = useFormStatus()` |
| `useOptimistic` | Optimistic UI updates | `const [optimistic, setOptimistic] = useOptimistic(state, reducer)` |
| `<form action>` | Form actions (client or server) | `<form action={serverAction}>` |
| Server Components | No JS shipped to client | Default in Next.js App Router |
| `ref` as prop | No more `forwardRef` needed | `function Input({ ref }) {}` |
| `<Context>` shorthand | Use context directly as provider | `<ThemeContext value={theme}>` |
| Document metadata | `<title>`, `<meta>` in components | Hoisted to `<head>` automatically |

---

## TypeScript + React Cheat Sheet

```tsx
// Component props
type Props = { name: string; age?: number };

// Children
type WithChildren = { children: ReactNode };
type RenderProp = { render: (data: Data) => ReactNode };

// Event handlers
type ButtonClick = React.MouseEventHandler<HTMLButtonElement>;
type InputChange = React.ChangeEventHandler<HTMLInputElement>;
type FormSubmit = React.FormEventHandler<HTMLFormElement>;
type KeyDown = React.KeyboardEventHandler<HTMLInputElement>;

// Refs
type InputRef = React.RefObject<HTMLInputElement>;
type DivRef = React.RefObject<HTMLDivElement>;

// Extending native elements
type ButtonProps = React.ComponentPropsWithoutRef<'button'> & { variant: string };
type InputProps = React.ComponentPropsWithRef<'input'> & { error?: string };

// Generic component
function List<T>({ items, render }: { items: T[]; render: (item: T) => ReactNode }) {}

// Discriminated union props
type Props = { mode: 'edit'; onSave: () => void } | { mode: 'view' };

// Extract element type
type ElementProps<T extends keyof JSX.IntrinsicElements> = JSX.IntrinsicElements[T];
```

---

## Anti-Patterns to Avoid

| Anti-Pattern | Problem | Fix |
|-------------|---------|-----|
| Inline objects in JSX | New reference every render | Extract to constant or useMemo |
| Inline functions as props | New reference every render | useCallback |
| Derived state in useState | State gets out of sync | Compute during render |
| useEffect for data transform | Extra render cycle | useMemo |
| Index as key | Bugs with reordering | Use stable unique ID |
| Global state for local needs | Unnecessary re-renders | Colocate state |
| Prop drilling 5+ levels | Fragile, hard to maintain | Context or state management |
| useEffect for events | Race conditions, stale closures | Event handler directly |
| Fetching in useEffect | No caching, no dedup, waterfalls | TanStack Query or RSC |
| Storing computed values | Stale data, extra state | Derive from source of truth |

---

## File Organization Conventions

```
src/
├── app/                    # Next.js App Router pages
├── components/
│   ├── ui/                 # Base UI components (Button, Input, Card)
│   ├── forms/              # Form-specific components
│   ├── layout/             # Layout components (Header, Sidebar, Footer)
│   └── features/           # Feature-specific components
│       ├── auth/
│       ├── dashboard/
│       └── settings/
├── hooks/                  # Shared custom hooks
├── lib/                    # Utilities, API client, constants
│   ├── utils.ts            # cn(), formatDate(), etc.
│   ├── api.ts              # API client functions
│   └── constants.ts
├── stores/                 # Zustand/Jotai stores
├── types/                  # Shared TypeScript types
└── styles/                 # Global styles, Tailwind config
```
