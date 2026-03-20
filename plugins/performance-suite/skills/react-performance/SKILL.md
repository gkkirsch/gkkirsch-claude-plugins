---
name: react-performance
description: >
  React rendering optimization — React.memo, useMemo, useCallback, virtualization,
  code splitting, Suspense, and profiling techniques.
  Triggers: "react performance", "react memo", "useMemo", "useCallback", "react rendering",
  "react virtualization", "react lazy", "react profiler", "re-render".
  NOT for: Core Web Vitals (use web-vitals), caching strategies (use caching-patterns).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# React Performance Optimization

## Understanding Re-Renders

```
When does a component re-render?
  1. Its state changes (useState, useReducer)
  2. Its parent re-renders (unless memoized)
  3. A context it consumes changes
  4. A custom hook's state changes

When does it NOT re-render?
  - Props change alone (parent must re-render too)
  - Ref changes (useRef doesn't trigger re-render)
  - State is set to same value (React bails out)
```

## React.memo — Prevent Unnecessary Re-Renders

```typescript
// BEFORE: re-renders every time parent renders
function UserCard({ user }: { user: User }) {
  return (
    <div>
      <h3>{user.name}</h3>
      <p>{user.email}</p>
    </div>
  );
}

// AFTER: only re-renders when user prop changes (shallow compare)
const UserCard = React.memo(function UserCard({ user }: { user: User }) {
  return (
    <div>
      <h3>{user.name}</h3>
      <p>{user.email}</p>
    </div>
  );
});

// Custom comparison (use sparingly)
const UserCard = React.memo(
  function UserCard({ user }: { user: User }) { /* ... */ },
  (prevProps, nextProps) => prevProps.user.id === nextProps.user.id,
);
```

### When to use React.memo

```
USE React.memo when:
  ✓ Component renders often with same props (list items)
  ✓ Component is expensive to render (complex DOM, calculations)
  ✓ Parent re-renders frequently but child props rarely change
  ✓ Pure display components (no internal state)

SKIP React.memo when:
  ✗ Component always receives new props
  ✗ Component is cheap to render (few DOM nodes)
  ✗ Props are primitives that change every render
  ✗ Component has few siblings (overhead > benefit)
```

## useMemo — Cache Expensive Computations

```typescript
// BEFORE: filters and sorts on every render
function UserList({ users, search }: Props) {
  const filtered = users
    .filter(u => u.name.toLowerCase().includes(search.toLowerCase()))
    .sort((a, b) => a.name.localeCompare(b.name));

  return filtered.map(u => <UserCard key={u.id} user={u} />);
}

// AFTER: only recomputes when users or search changes
function UserList({ users, search }: Props) {
  const filtered = useMemo(
    () => users
      .filter(u => u.name.toLowerCase().includes(search.toLowerCase()))
      .sort((a, b) => a.name.localeCompare(b.name)),
    [users, search],
  );

  return filtered.map(u => <UserCard key={u.id} user={u} />);
}

// Memoize object/array props to prevent child re-renders
function Parent() {
  const [count, setCount] = useState(0);

  // Without useMemo: new object every render → child re-renders
  // const style = { color: 'red', fontSize: 16 };

  // With useMemo: same reference → child skips re-render
  const style = useMemo(() => ({ color: 'red', fontSize: 16 }), []);

  return <Child style={style} />;
}
```

## useCallback — Stable Function References

```typescript
// BEFORE: new function every render → breaks memo on children
function TodoList({ todos }: Props) {
  const handleToggle = (id: string) => {
    // ...
  };

  return todos.map(todo => (
    <TodoItem key={todo.id} todo={todo} onToggle={handleToggle} />
  ));
}

// AFTER: stable reference → memoized children don't re-render
function TodoList({ todos }: Props) {
  const handleToggle = useCallback((id: string) => {
    setTodos(prev => prev.map(t =>
      t.id === id ? { ...t, done: !t.done } : t
    ));
  }, []); // empty deps because we use functional setState

  return todos.map(todo => (
    <MemoizedTodoItem key={todo.id} todo={todo} onToggle={handleToggle} />
  ));
}

// useCallback + React.memo = effective pair
const TodoItem = React.memo(function TodoItem({
  todo,
  onToggle,
}: {
  todo: Todo;
  onToggle: (id: string) => void;
}) {
  return (
    <div onClick={() => onToggle(todo.id)}>
      {todo.text}
    </div>
  );
});
```

## Code Splitting with React.lazy

```typescript
import React, { Suspense } from 'react';

// Route-level splitting
const Dashboard = React.lazy(() => import('./pages/Dashboard'));
const Settings = React.lazy(() => import('./pages/Settings'));
const Analytics = React.lazy(() => import('./pages/Analytics'));

function App() {
  return (
    <Suspense fallback={<PageSkeleton />}>
      <Routes>
        <Route path="/dashboard" element={<Dashboard />} />
        <Route path="/settings" element={<Settings />} />
        <Route path="/analytics" element={<Analytics />} />
      </Routes>
    </Suspense>
  );
}

// Component-level splitting (heavy components)
const HeavyChart = React.lazy(() => import('./components/HeavyChart'));
const RichTextEditor = React.lazy(() => import('./components/RichTextEditor'));

function PostEditor() {
  return (
    <div>
      <Suspense fallback={<div className="h-96 animate-pulse bg-gray-100 rounded" />}>
        <RichTextEditor />
      </Suspense>
    </div>
  );
}

// Named export lazy loading
const MyComponent = React.lazy(() =>
  import('./components/Multi').then(module => ({ default: module.MyComponent }))
);

// Preload on hover/intent
const Settings = React.lazy(() => import('./pages/Settings'));
function NavLink() {
  const preload = () => import('./pages/Settings');
  return <a href="/settings" onMouseEnter={preload}>Settings</a>;
}
```

## Virtualization (Large Lists)

```typescript
// Using @tanstack/react-virtual
import { useVirtualizer } from '@tanstack/react-virtual';

function VirtualList({ items }: { items: Item[] }) {
  const parentRef = useRef<HTMLDivElement>(null);

  const virtualizer = useVirtualizer({
    count: items.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => 60,       // estimated row height in px
    overscan: 5,                   // render 5 extra items above/below
  });

  return (
    <div ref={parentRef} className="h-[600px] overflow-auto">
      <div style={{ height: `${virtualizer.getTotalSize()}px`, position: 'relative' }}>
        {virtualizer.getVirtualItems().map(virtualRow => (
          <div
            key={virtualRow.key}
            style={{
              position: 'absolute',
              top: 0,
              left: 0,
              width: '100%',
              height: `${virtualRow.size}px`,
              transform: `translateY(${virtualRow.start}px)`,
            }}
          >
            <ItemRow item={items[virtualRow.index]} />
          </div>
        ))}
      </div>
    </div>
  );
}
```

## Context Optimization

```typescript
// PROBLEM: one big context → every consumer re-renders on any change
const AppContext = createContext({ user: null, theme: 'light', locale: 'en' });

// SOLUTION 1: Split contexts by update frequency
const UserContext = createContext<User | null>(null);
const ThemeContext = createContext<'light' | 'dark'>('light');
const LocaleContext = createContext<string>('en');

// SOLUTION 2: Memoize context value
function ThemeProvider({ children }: { children: ReactNode }) {
  const [theme, setTheme] = useState<'light' | 'dark'>('light');

  // Without useMemo: new object every render → all consumers re-render
  // const value = { theme, setTheme };

  // With useMemo: stable reference → consumers only re-render when theme changes
  const value = useMemo(() => ({ theme, setTheme }), [theme]);

  return <ThemeContext.Provider value={value}>{children}</ThemeContext.Provider>;
}

// SOLUTION 3: Selector pattern (use-context-selector or zustand)
import { useContextSelector } from 'use-context-selector';

function UserName() {
  // Only re-renders when user.name changes, not on every context update
  const name = useContextSelector(AppContext, ctx => ctx.user?.name);
  return <span>{name}</span>;
}
```

## React Profiler

```typescript
// Programmatic profiler
import { Profiler, ProfilerOnRenderCallback } from 'react';

const onRender: ProfilerOnRenderCallback = (
  id,           // Profiler tree id
  phase,        // "mount" or "update"
  actualDuration,   // Time spent rendering
  baseDuration,     // Estimated time without memoization
  startTime,
  commitTime,
) => {
  if (actualDuration > 16) { // Longer than one frame (60fps)
    console.warn(`Slow render: ${id} took ${actualDuration.toFixed(1)}ms`);
  }
};

function App() {
  return (
    <Profiler id="Dashboard" onRender={onRender}>
      <Dashboard />
    </Profiler>
  );
}

// React DevTools Profiler (browser extension)
// 1. Open React DevTools → Profiler tab
// 2. Click record, interact with app, stop recording
// 3. Flame chart shows component render times
// 4. "Why did this render?" shows the cause
```

## Image Optimization

```typescript
// Next.js Image (automatic optimization)
import Image from 'next/image';
<Image
  src="/hero.jpg"
  alt="Hero"
  width={1200}
  height={600}
  priority        // LCP image — no lazy loading
  placeholder="blur"
  blurDataURL="data:image/..." // base64 placeholder
/>

// Native lazy loading
<img
  src="photo.jpg"
  alt="Photo"
  loading="lazy"           // Lazy load below fold
  decoding="async"         // Don't block main thread
  width="400"              // Prevent layout shift
  height="300"
/>

// Responsive images
<img
  srcSet="photo-400.webp 400w, photo-800.webp 800w, photo-1200.webp 1200w"
  sizes="(max-width: 640px) 400px, (max-width: 1024px) 800px, 1200px"
  src="photo-800.webp"
  alt="Responsive photo"
  loading="lazy"
/>
```

## State Update Batching

```typescript
// React 18: automatic batching (all updates batched by default)
function handleClick() {
  setCount(c => c + 1);
  setFlag(f => !f);
  setText('updated');
  // Only ONE re-render for all three updates ✓
}

// Even in async code (React 18+)
async function handleSubmit() {
  const data = await fetchData();
  setData(data);
  setLoading(false);
  setError(null);
  // Only ONE re-render ✓ (React 18 batches async too)
}

// Force synchronous update (rare need)
import { flushSync } from 'react-dom';
flushSync(() => setCount(c => c + 1));
// DOM is updated here
measureLayout(); // safe to read DOM
```

## Gotchas

1. **React.memo does shallow comparison.** If you pass `style={{ color: 'red' }}` inline, it creates a new object every render and memo can't help. Extract to a constant or wrap in useMemo.

2. **useMemo/useCallback are NOT free.** They have overhead (storing previous value, comparing deps). Only use when the computation is actually expensive OR when stabilizing references for memoized children.

3. **Don't wrap everything in React.memo.** Measure first. If a component renders in <1ms, the overhead of memo + shallow comparison might cost more than the re-render itself.

4. **useCallback with empty deps and setState.** Use functional updates (`setItems(prev => ...)`) instead of referencing state in deps. This keeps the callback stable without stale closures.

5. **Context re-renders ALL consumers.** Even if you only read `theme` from a context that also has `user`, changing `user` re-renders your component. Split contexts or use a selector library.

6. **Key changes force remount, not re-render.** `key={item.id}` is correct. `key={Math.random()}` remounts on every render (destroys and recreates the component). Never use random keys.
