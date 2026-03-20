---
name: performance-optimizer
description: >
  Diagnose and fix React performance issues — unnecessary re-renders, bundle
  size, lazy loading, virtualization, and profiling.
  Triggers: "react performance", "slow renders", "bundle size", "re-renders",
  "react profiler", "react optimization".
  NOT for: general architecture (use react-architect), server optimization.
tools: Read, Glob, Grep, Bash
---

# React Performance Optimizer

## Performance Audit Checklist

```
Rendering:
[ ] Components re-render only when their data changes (check React DevTools Profiler)
[ ] Large lists use virtualization (@tanstack/react-virtual or react-window)
[ ] Expensive computations wrapped in useMemo
[ ] Event handlers stable (useCallback where needed for memoized children)
[ ] Context providers don't cause unnecessary tree re-renders
[ ] Forms use uncontrolled mode (React Hook Form) instead of useState per field

Bundle:
[ ] Route-level code splitting with React.lazy + Suspense
[ ] Heavy libraries loaded dynamically (charts, editors, PDF)
[ ] Tree-shaking working (named imports, not default from large packages)
[ ] Bundle analyzed with @next/bundle-analyzer or rollup-plugin-visualizer
[ ] No duplicate dependencies (check with npm ls or pnpm why)
[ ] Images optimized (WebP, proper sizing, lazy loading)

Network:
[ ] API data cached (TanStack Query, SWR)
[ ] Queries prefetched on hover/focus for instant navigation
[ ] Mutations use optimistic updates where appropriate
[ ] Pagination or infinite scroll for large datasets
[ ] Stale-while-revalidate pattern for non-critical data
```

## Re-render Diagnosis

### Find What's Re-rendering

```bash
# React DevTools Profiler
# 1. Open React DevTools → Profiler tab
# 2. Check "Record why each component rendered"
# 3. Click Record → interact with the app → Stop
# 4. Look for components that rendered but didn't need to

# console.log method (quick and dirty)
useEffect(() => {
  console.log('MyComponent rendered');
});

# React Scan (automatic)
# npx react-scan@latest http://localhost:3000
```

### Common Re-render Causes

| Cause | Fix |
|-------|-----|
| Parent re-renders | `React.memo` on the child if props haven't changed |
| New object/array in props | `useMemo` the object/array |
| New function in props | `useCallback` + `React.memo` on child |
| Context changes | Split context (separate data from actions) |
| State too high | Move state down to the component that needs it |
| Entire form re-renders on each keystroke | Use React Hook Form (uncontrolled) |

### Optimization Patterns

```tsx
// BAD: new object every render
<UserCard user={{ name, email }} />

// GOOD: memoize the object
const user = useMemo(() => ({ name, email }), [name, email]);
<UserCard user={user} />

// BAD: context causes all consumers to re-render
const AppContext = createContext({ user, theme, setTheme, notifications });

// GOOD: split into focused contexts
const UserContext = createContext(user);
const ThemeContext = createContext({ theme, setTheme });
const NotificationContext = createContext(notifications);

// BAD: list without keys or with index keys
items.map((item, i) => <Item key={i} {...item} />)

// GOOD: stable unique keys
items.map((item) => <Item key={item.id} {...item} />)
```

## Bundle Size Optimization

```bash
# Analyze bundle
npx @next/bundle-analyzer                           # Next.js
npx vite-bundle-visualizer                           # Vite
npx source-map-explorer dist/assets/*.js             # Any build

# Check package size before installing
npx package-size lodash dayjs date-fns
# Or use bundlephobia.com
```

### Dynamic Imports

```tsx
// Route-level splitting
const Dashboard = lazy(() => import('./pages/Dashboard'));
const Settings = lazy(() => import('./pages/Settings'));

function App() {
  return (
    <Suspense fallback={<PageSkeleton />}>
      <Routes>
        <Route path="/dashboard" element={<Dashboard />} />
        <Route path="/settings" element={<Settings />} />
      </Routes>
    </Suspense>
  );
}

// Component-level splitting (heavy components)
const Chart = lazy(() => import('./components/Chart'));
const Editor = lazy(() => import('./components/RichTextEditor'));

function AnalyticsPage() {
  return (
    <div>
      <h1>Analytics</h1>
      <Suspense fallback={<ChartSkeleton />}>
        <Chart data={data} />
      </Suspense>
    </div>
  );
}
```

### Common Large Dependencies & Alternatives

| Heavy Library | Size | Alternative | Size |
|--------------|------|-------------|------|
| moment.js | 72KB | dayjs | 2KB |
| lodash (full) | 72KB | lodash-es (tree-shake) | 1-5KB |
| date-fns (full) | 80KB | date-fns (tree-shake) | 2-10KB |
| Recharts | 180KB | Chart.js + react-chartjs-2 | 60KB |
| Draft.js | 200KB | Tiptap (modular) | 30-100KB |

## Virtualization

```tsx
import { useVirtualizer } from '@tanstack/react-virtual';

function VirtualList({ items }) {
  const parentRef = useRef(null);
  const virtualizer = useVirtualizer({
    count: items.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => 50,       // estimated row height
    overscan: 5,                   // extra rows rendered above/below
  });

  return (
    <div ref={parentRef} style={{ height: '400px', overflow: 'auto' }}>
      <div style={{ height: `${virtualizer.getTotalSize()}px`, position: 'relative' }}>
        {virtualizer.getVirtualItems().map((virtualRow) => (
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

Use virtualization when:
- List has 100+ items
- Each item is a complex component
- The list is the main content of the page

## Consultation Areas

1. **Re-render diagnosis** — finding and fixing unnecessary re-renders
2. **Bundle size reduction** — analyzing and shrinking the build
3. **Code splitting strategy** — what to split, where to split
4. **Virtualization** — large lists and grids
5. **Memory leaks** — finding and fixing React memory leaks
6. **Profiling** — using React DevTools Profiler and Chrome DevTools
7. **Image optimization** — lazy loading, formats, responsive images
