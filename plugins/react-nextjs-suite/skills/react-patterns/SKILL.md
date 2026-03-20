---
name: react-patterns
description: >
  Advanced React patterns — compound components, custom hooks, render props,
  error boundaries, React Server Components, Suspense, transitions, and
  performance optimization with memo/useMemo/useCallback.
  Triggers: "react patterns", "compound components", "custom hooks",
  "error boundary", "react server components", "suspense", "react performance".
  NOT for: Next.js-specific patterns (use nextjs-app-router).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Advanced React Patterns

## Compound Components

```tsx
import { createContext, useContext, useState, ReactNode } from "react";

// Context for internal state
interface AccordionContext {
  openItems: Set<string>;
  toggle: (id: string) => void;
}

const AccordionCtx = createContext<AccordionContext | null>(null);

function useAccordion() {
  const ctx = useContext(AccordionCtx);
  if (!ctx) throw new Error("Accordion components must be used within <Accordion>");
  return ctx;
}

// Parent component owns the state
function Accordion({ children, multiple = false }: { children: ReactNode; multiple?: boolean }) {
  const [openItems, setOpenItems] = useState<Set<string>>(new Set());

  const toggle = (id: string) => {
    setOpenItems((prev) => {
      const next = new Set(multiple ? prev : []);
      if (prev.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  };

  return (
    <AccordionCtx.Provider value={{ openItems, toggle }}>
      <div role="region">{children}</div>
    </AccordionCtx.Provider>
  );
}

// Child components consume context
function AccordionItem({ id, title, children }: { id: string; title: string; children: ReactNode }) {
  const { openItems, toggle } = useAccordion();
  const isOpen = openItems.has(id);

  return (
    <div>
      <button
        onClick={() => toggle(id)}
        aria-expanded={isOpen}
        aria-controls={`panel-${id}`}
      >
        {title}
        <span>{isOpen ? "▲" : "▼"}</span>
      </button>
      {isOpen && (
        <div id={`panel-${id}`} role="region">
          {children}
        </div>
      )}
    </div>
  );
}

// Attach as static properties
Accordion.Item = AccordionItem;

// Usage: <Accordion><Accordion.Item id="1" title="FAQ">Answer</Accordion.Item></Accordion>
```

## Custom Hooks

### Data Fetching Hook

```tsx
import { useState, useEffect, useCallback, useRef } from "react";

interface UseFetchResult<T> {
  data: T | null;
  error: Error | null;
  isLoading: boolean;
  refetch: () => void;
}

function useFetch<T>(url: string, options?: RequestInit): UseFetchResult<T> {
  const [data, setData] = useState<T | null>(null);
  const [error, setError] = useState<Error | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const abortRef = useRef<AbortController | null>(null);

  const fetchData = useCallback(async () => {
    // Cancel previous request
    abortRef.current?.abort();
    const controller = new AbortController();
    abortRef.current = controller;

    setIsLoading(true);
    setError(null);

    try {
      const res = await fetch(url, { ...options, signal: controller.signal });
      if (!res.ok) throw new Error(`HTTP ${res.status}: ${res.statusText}`);
      const json = await res.json();
      setData(json);
    } catch (err) {
      if (err instanceof Error && err.name !== "AbortError") {
        setError(err);
      }
    } finally {
      setIsLoading(false);
    }
  }, [url]);

  useEffect(() => {
    fetchData();
    return () => abortRef.current?.abort();
  }, [fetchData]);

  return { data, error, isLoading, refetch: fetchData };
}
```

### Local Storage Hook

```tsx
function useLocalStorage<T>(key: string, initialValue: T) {
  const [storedValue, setStoredValue] = useState<T>(() => {
    try {
      const item = window.localStorage.getItem(key);
      return item ? (JSON.parse(item) as T) : initialValue;
    } catch {
      return initialValue;
    }
  });

  const setValue = useCallback(
    (value: T | ((prev: T) => T)) => {
      setStoredValue((prev) => {
        const next = value instanceof Function ? value(prev) : value;
        window.localStorage.setItem(key, JSON.stringify(next));
        return next;
      });
    },
    [key]
  );

  const removeValue = useCallback(() => {
    window.localStorage.removeItem(key);
    setStoredValue(initialValue);
  }, [key, initialValue]);

  return [storedValue, setValue, removeValue] as const;
}

// Usage: const [theme, setTheme] = useLocalStorage("theme", "dark");
```

### Debounced Value Hook

```tsx
function useDebounce<T>(value: T, delayMs: number): T {
  const [debouncedValue, setDebouncedValue] = useState(value);

  useEffect(() => {
    const timer = setTimeout(() => setDebouncedValue(value), delayMs);
    return () => clearTimeout(timer);
  }, [value, delayMs]);

  return debouncedValue;
}

// Usage in search:
function SearchInput() {
  const [query, setQuery] = useState("");
  const debouncedQuery = useDebounce(query, 300);

  useEffect(() => {
    if (debouncedQuery) {
      // Fetch search results
      fetch(`/api/search?q=${encodeURIComponent(debouncedQuery)}`);
    }
  }, [debouncedQuery]);

  return <input value={query} onChange={(e) => setQuery(e.target.value)} />;
}
```

## Error Boundaries

```tsx
import { Component, ErrorInfo, ReactNode } from "react";

interface ErrorBoundaryProps {
  children: ReactNode;
  fallback?: ReactNode | ((error: Error, reset: () => void) => ReactNode);
  onError?: (error: Error, info: ErrorInfo) => void;
}

interface ErrorBoundaryState {
  error: Error | null;
}

class ErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  state: ErrorBoundaryState = { error: null };

  static getDerivedStateFromError(error: Error) {
    return { error };
  }

  componentDidCatch(error: Error, info: ErrorInfo) {
    this.props.onError?.(error, info);
    // Log to error tracking service
    console.error("ErrorBoundary caught:", error, info.componentStack);
  }

  reset = () => this.setState({ error: null });

  render() {
    const { error } = this.state;
    const { fallback, children } = this.props;

    if (error) {
      if (typeof fallback === "function") return fallback(error, this.reset);
      return fallback ?? <div>Something went wrong</div>;
    }

    return children;
  }
}

// Usage with retry button:
<ErrorBoundary
  fallback={(error, reset) => (
    <div>
      <p>Error: {error.message}</p>
      <button onClick={reset}>Try Again</button>
    </div>
  )}
  onError={(error) => Sentry.captureException(error)}
>
  <UserProfile />
</ErrorBoundary>
```

## React Server Components (RSC)

```tsx
// Server Component (default in Next.js App Router)
// Can: read files, query DB, use secrets, await async
// Cannot: useState, useEffect, event handlers, browser APIs
async function ProductList() {
  const products = await db.product.findMany({
    where: { active: true },
    orderBy: { createdAt: "desc" },
  });

  return (
    <ul>
      {products.map((p) => (
        <li key={p.id}>
          <span>{p.name}</span>
          <span>${p.price}</span>
          {/* Client component for interactivity */}
          <AddToCartButton productId={p.id} />
        </li>
      ))}
    </ul>
  );
}

// Client Component — interactivity required
"use client";
import { useState, useTransition } from "react";

function AddToCartButton({ productId }: { productId: string }) {
  const [isPending, startTransition] = useTransition();
  const [added, setAdded] = useState(false);

  const handleClick = () => {
    startTransition(async () => {
      await fetch("/api/cart", {
        method: "POST",
        body: JSON.stringify({ productId }),
      });
      setAdded(true);
    });
  };

  return (
    <button onClick={handleClick} disabled={isPending}>
      {isPending ? "Adding..." : added ? "Added ✓" : "Add to Cart"}
    </button>
  );
}
```

### RSC Rules

| Can Do (Server) | Cannot Do (Server) | Can Do (Client) |
|------------------|--------------------|-----------------|
| `await` at component level | `useState` / `useEffect` | All hooks |
| Direct DB queries | Event handlers (onClick) | Event handlers |
| `fs.readFile` | `useContext` | Browser APIs |
| Access env vars / secrets | `window` / `document` | `localStorage` |
| Import server-only libs | Import client-only libs | WebSocket |

## Suspense & Transitions

```tsx
import { Suspense, useTransition, useState } from "react";

// Suspense wraps async server components or lazy-loaded components
function Dashboard() {
  return (
    <div>
      <h1>Dashboard</h1>

      {/* Each Suspense shows its own loading state */}
      <Suspense fallback={<Skeleton type="chart" />}>
        <RevenueChart />
      </Suspense>

      <Suspense fallback={<Skeleton type="table" />}>
        <RecentOrders />
      </Suspense>

      {/* Nested Suspense for progressive loading */}
      <Suspense fallback={<Skeleton type="list" />}>
        <UserList>
          <Suspense fallback={<Skeleton type="detail" />}>
            <UserDetails />
          </Suspense>
        </UserList>
      </Suspense>
    </div>
  );
}

// useTransition for non-blocking updates
function TabContainer() {
  const [tab, setTab] = useState("home");
  const [isPending, startTransition] = useTransition();

  const selectTab = (nextTab: string) => {
    startTransition(() => {
      setTab(nextTab);  // This update won't block the UI
    });
  };

  return (
    <div>
      <nav style={{ opacity: isPending ? 0.7 : 1 }}>
        {["home", "profile", "settings"].map((t) => (
          <button key={t} onClick={() => selectTab(t)} disabled={tab === t}>
            {t}
          </button>
        ))}
      </nav>
      <Suspense fallback={<Spinner />}>
        <TabContent tab={tab} />
      </Suspense>
    </div>
  );
}
```

## Performance Optimization

```tsx
import { memo, useMemo, useCallback, lazy, Suspense } from "react";

// memo — skip re-render if props haven't changed
const ExpensiveList = memo(function ExpensiveList({
  items,
  onSelect,
}: {
  items: Item[];
  onSelect: (id: string) => void;
}) {
  return (
    <ul>
      {items.map((item) => (
        <li key={item.id} onClick={() => onSelect(item.id)}>
          {item.name}
        </li>
      ))}
    </ul>
  );
});

// Parent must stabilize props for memo to work
function ParentComponent({ data }: { data: Item[] }) {
  const [selected, setSelected] = useState<string | null>(null);

  // useCallback — stable function reference across renders
  const handleSelect = useCallback((id: string) => {
    setSelected(id);
  }, []);

  // useMemo — expensive computation cached
  const sortedItems = useMemo(
    () => [...data].sort((a, b) => a.name.localeCompare(b.name)),
    [data]
  );

  return <ExpensiveList items={sortedItems} onSelect={handleSelect} />;
}

// Lazy loading — code-split large components
const HeavyChart = lazy(() => import("./HeavyChart"));

function AnalyticsPage() {
  return (
    <Suspense fallback={<div>Loading chart...</div>}>
      <HeavyChart />
    </Suspense>
  );
}
```

### When to Optimize

| Technique | When to Use | When NOT to Use |
|-----------|------------|-----------------|
| `memo` | Large lists, expensive renders, stable parents | Small components, frequently changing props |
| `useMemo` | Expensive calculations, referential equality | Simple math, primitives |
| `useCallback` | Passing callbacks to memoized children | Simple handlers, no memoized children |
| `lazy` | Large route-level components, rarely visited pages | Small components, above-the-fold content |

## Context Performance Pattern

```tsx
// Split context to avoid unnecessary re-renders
const UserDataCtx = createContext<User | null>(null);
const UserActionsCtx = createContext<UserActions | null>(null);

function UserProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);

  // Actions object is stable — doesn't change on re-render
  const actions = useMemo(
    () => ({
      login: async (creds: Credentials) => {
        const user = await authApi.login(creds);
        setUser(user);
      },
      logout: () => setUser(null),
      updateProfile: async (data: Partial<User>) => {
        const updated = await authApi.updateProfile(data);
        setUser(updated);
      },
    }),
    []
  );

  return (
    <UserActionsCtx.Provider value={actions}>
      <UserDataCtx.Provider value={user}>
        {children}
      </UserDataCtx.Provider>
    </UserActionsCtx.Provider>
  );
}

// Components reading only actions don't re-render when user data changes
function LogoutButton() {
  const { logout } = useContext(UserActionsCtx)!;
  return <button onClick={logout}>Log Out</button>;
}
```

## Gotchas

1. **`memo` is useless if parent creates new objects/arrays each render** — `<List items={data.filter(x => x.active)} />` creates a new array every render, defeating memo. Move the filter into `useMemo`.

2. **`useEffect` cleanup runs before the next effect, not on unmount only** — If your effect runs on every render, cleanup runs every time too. Use the dependency array to control when effects fire.

3. **Server Components can't be children of Client Components directly** — Pass them as `children` props or use composition. `"use client"` in a file makes everything in that file client-side.

4. **`useCallback` without dependencies array doesn't memoize** — `useCallback(fn)` is the same as just writing `fn`. You must pass `[deps]` as the second argument.

5. **Context causes all consumers to re-render** — Even if only one field in the context object changed, every `useContext` consumer re-renders. Split contexts by update frequency.

6. **Lazy components must be default exports** — `lazy(() => import("./Chart"))` expects the module to have `export default`. Named exports need: `lazy(() => import("./Chart").then(m => ({ default: m.Chart })))`.
