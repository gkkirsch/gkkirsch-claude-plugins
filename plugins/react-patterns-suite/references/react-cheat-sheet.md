# React Cheat Sheet

## Hook Rules

1. Only call hooks at the top level (not inside if/else, loops, or after returns)
2. Only call hooks from React function components or custom hooks
3. Custom hooks must start with `use`

## Built-in Hooks

| Hook | Purpose | When to Use |
|------|---------|-------------|
| `useState` | Component state | Simple local state (booleans, strings, numbers) |
| `useReducer` | Complex state | State machines, multiple related values |
| `useEffect` | Side effects | API calls, subscriptions, DOM manipulation |
| `useLayoutEffect` | Sync side effects | Measure DOM before paint (rare) |
| `useRef` | Mutable ref | DOM refs, previous values, stable references |
| `useMemo` | Memoize value | Expensive computations, referential stability |
| `useCallback` | Memoize function | Stable function refs for memoized children |
| `useContext` | Read context | Theme, auth, locale (low-frequency updates) |
| `useTransition` | Non-urgent updates | Slow state updates that shouldn't block input |
| `useDeferredValue` | Defer a value | Expensive renders from fast-changing input |
| `useId` | Unique IDs | SSR-safe IDs for accessibility attributes |
| `useSyncExternalStore` | External state | Subscribe to non-React stores |
| `useImperativeHandle` | Custom ref API | Expose methods on a forwarded ref (rare) |

## Component Lifecycle

```
Mount:
  1. Component function runs
  2. DOM updates
  3. useLayoutEffect runs (sync)
  4. Browser paints
  5. useEffect runs (async)

Update (state/prop change):
  1. Component function re-runs
  2. DOM updates
  3. useLayoutEffect cleanup → runs
  4. Browser paints
  5. useEffect cleanup → runs

Unmount:
  1. useLayoutEffect cleanup
  2. useEffect cleanup
  3. Component removed from DOM
```

## JSX Patterns

```tsx
// Conditional rendering
{isLoggedIn && <Dashboard />}                  // Short circuit
{isLoggedIn ? <Dashboard /> : <Login />}       // Ternary
{status === 'loading' && <Spinner />}          // Status check

// List rendering
{items.map(item => <Item key={item.id} {...item} />)}

// Dynamic element
const Tag = `h${level}` as keyof JSX.IntrinsicElements;
<Tag className="heading">{children}</Tag>

// Spread props
<input {...register('name')} className="input" />

// Children
{React.Children.map(children, child =>
  React.isValidElement(child) ? React.cloneElement(child, { extra: 'prop' }) : child
)}
```

## TypeScript Patterns

```tsx
// Component props
interface Props {
  children: React.ReactNode;          // Any renderable content
  onClick: () => void;                // Void handler
  onChange: (value: string) => void;   // Handler with argument
  style?: React.CSSProperties;        // CSS style object
  className?: string;                  // CSS class
  as?: React.ElementType;             // Polymorphic component
  ref?: React.Ref<HTMLDivElement>;    // Ref type
}

// Event handlers
const handleClick = (e: React.MouseEvent<HTMLButtonElement>) => {};
const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {};
const handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {};
const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {};

// Generic component
function List<T>({ items, renderItem }: {
  items: T[];
  renderItem: (item: T) => ReactNode;
}) {
  return <ul>{items.map(renderItem)}</ul>;
}

// Discriminated union props
type ButtonProps =
  | { variant: 'link'; href: string; onClick?: never }
  | { variant: 'button'; onClick: () => void; href?: never };
```

## Performance Quick Reference

| Technique | When | How |
|-----------|------|-----|
| `React.memo` | Child re-renders with same props | `export default memo(MyComponent)` |
| `useMemo` | Expensive computation | `useMemo(() => compute(data), [data])` |
| `useCallback` | Callback to memo'd child | `useCallback(() => doThing(id), [id])` |
| `lazy` + `Suspense` | Route/heavy component | `const Page = lazy(() => import('./Page'))` |
| Virtualization | Lists with 100+ items | `@tanstack/react-virtual` |
| Key stability | List reordering | Always use unique stable IDs, never index |

## State Management Decision

```
Is it form data?          → React Hook Form
Is it from an API?        → TanStack Query
Is it in the URL?         → URL search params
Is it shared globally?    → Zustand
Is it component-local?    → useState / useReducer
Is it rare/slow updates?  → Context
```

## Common Import Patterns

```tsx
// React
import { useState, useEffect, useCallback, useMemo, useRef } from 'react';
import { createContext, useContext } from 'react';
import { lazy, Suspense } from 'react';
import { createPortal } from 'react-dom';

// React Hook Form + Zod
import { useForm, Controller, useFieldArray, FormProvider, useFormContext } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';

// TanStack Query
import { useQuery, useMutation, useQueryClient, useInfiniteQuery } from '@tanstack/react-query';

// Zustand
import { create } from 'zustand';
import { devtools, persist } from 'zustand/middleware';
import { immer } from 'zustand/middleware/immer';
import { useShallow } from 'zustand/react/shallow';

// React Router
import { Link, useNavigate, useParams, useSearchParams, Outlet } from 'react-router-dom';

// Utility
import { cn } from '@/lib/utils';  // clsx + tailwind-merge
```
