---
name: react-patterns
description: >
  Core React component patterns — composition, compound components, render
  props, error boundaries, portals, forwardRef, and TypeScript patterns.
  Triggers: "react pattern", "component pattern", "compound component",
  "error boundary", "forwardRef", "render props", "react composition".
  NOT for: state management (use zustand-state), data fetching (use tanstack-query).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# React Component Patterns

## Component with TypeScript

```tsx
// Props with children
interface CardProps {
  title: string;
  description?: string;
  variant?: 'default' | 'outlined' | 'elevated';
  className?: string;
  children: React.ReactNode;
}

function Card({ title, description, variant = 'default', className, children }: CardProps) {
  return (
    <div className={cn('rounded-lg p-4', variantStyles[variant], className)}>
      <h3 className="font-semibold">{title}</h3>
      {description && <p className="text-muted-foreground text-sm">{description}</p>}
      <div className="mt-2">{children}</div>
    </div>
  );
}

// Polymorphic component (render as different elements)
type ButtonProps<T extends React.ElementType = 'button'> = {
  as?: T;
  variant?: 'primary' | 'secondary' | 'ghost';
  size?: 'sm' | 'md' | 'lg';
  isLoading?: boolean;
} & Omit<React.ComponentPropsWithoutRef<T>, 'as'>;

function Button<T extends React.ElementType = 'button'>({
  as,
  variant = 'primary',
  size = 'md',
  isLoading,
  children,
  ...props
}: ButtonProps<T>) {
  const Component = as || 'button';
  return (
    <Component
      className={cn(buttonStyles({ variant, size }))}
      disabled={isLoading}
      {...props}
    >
      {isLoading ? <Spinner className="mr-2" /> : null}
      {children}
    </Component>
  );
}

// Usage:
<Button>Click me</Button>
<Button as="a" href="/about">Link button</Button>
<Button as={Link} to="/dashboard">Router link</Button>
```

## Compound Components

```tsx
// Context for shared state
const SelectContext = createContext<{
  value: string;
  onChange: (value: string) => void;
  open: boolean;
  setOpen: (open: boolean) => void;
} | null>(null);

function useSelectContext() {
  const context = useContext(SelectContext);
  if (!context) throw new Error('Select components must be used within <Select>');
  return context;
}

// Root component
function Select({ value, onChange, children }: {
  value: string;
  onChange: (value: string) => void;
  children: React.ReactNode;
}) {
  const [open, setOpen] = useState(false);
  return (
    <SelectContext.Provider value={{ value, onChange, open, setOpen }}>
      <div className="relative">{children}</div>
    </SelectContext.Provider>
  );
}

// Sub-components
function SelectTrigger({ children }: { children: React.ReactNode }) {
  const { open, setOpen } = useSelectContext();
  return (
    <button onClick={() => setOpen(!open)} aria-expanded={open}>
      {children}
    </button>
  );
}

function SelectContent({ children }: { children: React.ReactNode }) {
  const { open } = useSelectContext();
  if (!open) return null;
  return <div className="absolute mt-1 w-full border rounded-lg shadow-lg bg-white">{children}</div>;
}

function SelectItem({ value, children }: { value: string; children: React.ReactNode }) {
  const { value: selected, onChange, setOpen } = useSelectContext();
  return (
    <button
      className={cn('w-full text-left px-3 py-2', value === selected && 'bg-blue-50')}
      onClick={() => { onChange(value); setOpen(false); }}
    >
      {children}
    </button>
  );
}

// Attach sub-components
Select.Trigger = SelectTrigger;
Select.Content = SelectContent;
Select.Item = SelectItem;

// Usage:
<Select value={role} onChange={setRole}>
  <Select.Trigger><Select.Value /></Select.Trigger>
  <Select.Content>
    <Select.Item value="admin">Admin</Select.Item>
    <Select.Item value="user">User</Select.Item>
  </Select.Content>
</Select>
```

## Error Boundary

```tsx
import { Component, ErrorInfo, ReactNode } from 'react';

interface ErrorBoundaryProps {
  fallback?: ReactNode | ((error: Error, reset: () => void) => ReactNode);
  onError?: (error: Error, info: ErrorInfo) => void;
  children: ReactNode;
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
  }

  resetError = () => this.setState({ error: null });

  render() {
    if (this.state.error) {
      if (typeof this.props.fallback === 'function') {
        return this.props.fallback(this.state.error, this.resetError);
      }
      return this.props.fallback ?? <DefaultErrorFallback error={this.state.error} reset={this.resetError} />;
    }
    return this.props.children;
  }
}

function DefaultErrorFallback({ error, reset }: { error: Error; reset: () => void }) {
  return (
    <div className="p-6 rounded-lg border border-red-200 bg-red-50">
      <h2 className="text-red-800 font-semibold">Something went wrong</h2>
      <p className="text-red-600 text-sm mt-1">{error.message}</p>
      <button onClick={reset} className="mt-3 text-sm underline text-red-700">
        Try again
      </button>
    </div>
  );
}

// Usage:
<ErrorBoundary
  fallback={(error, reset) => <CustomError error={error} onRetry={reset} />}
  onError={(error) => reportToSentry(error)}
>
  <DashboardWidget />
</ErrorBoundary>
```

## forwardRef with TypeScript

```tsx
import { forwardRef } from 'react';

interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  label: string;
  error?: string;
}

const Input = forwardRef<HTMLInputElement, InputProps>(
  ({ label, error, className, id, ...props }, ref) => {
    const inputId = id || label.toLowerCase().replace(/\s+/g, '-');
    return (
      <div className="space-y-1">
        <label htmlFor={inputId} className="text-sm font-medium">
          {label}
        </label>
        <input
          ref={ref}
          id={inputId}
          className={cn('w-full border rounded px-3 py-2', error && 'border-red-500', className)}
          aria-invalid={!!error}
          aria-describedby={error ? `${inputId}-error` : undefined}
          {...props}
        />
        {error && (
          <p id={`${inputId}-error`} className="text-sm text-red-500" role="alert">
            {error}
          </p>
        )}
      </div>
    );
  }
);
Input.displayName = 'Input';
```

## Portal

```tsx
import { createPortal } from 'react-dom';

function Modal({ open, onClose, children }: {
  open: boolean;
  onClose: () => void;
  children: React.ReactNode;
}) {
  if (!open) return null;

  return createPortal(
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      <div className="fixed inset-0 bg-black/50" onClick={onClose} />
      <div className="relative bg-white rounded-lg shadow-xl p-6 max-w-md w-full mx-4">
        {children}
      </div>
    </div>,
    document.body
  );
}
```

## Render Props Pattern

```tsx
interface DataLoaderProps<T> {
  url: string;
  children: (data: T, isLoading: boolean) => ReactNode;
  fallback?: ReactNode;
}

function DataLoader<T>({ url, children, fallback }: DataLoaderProps<T>) {
  const { data, isLoading, error } = useQuery({ queryKey: [url], queryFn: () => fetch(url).then(r => r.json()) });

  if (error) return <ErrorMessage error={error} />;
  if (isLoading && fallback) return <>{fallback}</>;
  if (!data) return null;

  return <>{children(data, isLoading)}</>;
}

// Usage:
<DataLoader<User[]> url="/api/users" fallback={<Skeleton />}>
  {(users) => (
    <ul>
      {users.map(u => <li key={u.id}>{u.name}</li>)}
    </ul>
  )}
</DataLoader>
```

## List Rendering Patterns

```tsx
// With empty state
function UserList({ users }: { users: User[] }) {
  if (users.length === 0) {
    return <EmptyState icon={Users} message="No users found" action={{ label: 'Add user', onClick: addUser }} />;
  }

  return (
    <ul className="divide-y">
      {users.map((user) => (
        <li key={user.id}>
          <UserRow user={user} />
        </li>
      ))}
    </ul>
  );
}

// With loading skeleton
function UserListContainer() {
  const { data: users, isLoading } = useQuery({ queryKey: ['users'], queryFn: fetchUsers });

  if (isLoading) {
    return (
      <ul className="divide-y">
        {Array.from({ length: 5 }).map((_, i) => (
          <li key={i}><UserRowSkeleton /></li>
        ))}
      </ul>
    );
  }

  return <UserList users={users ?? []} />;
}
```

## Gotchas

1. **Don't create components inside render.** Each render creates a new component type, unmounting and remounting the entire subtree. Define components outside or use `useMemo` for dynamic component creation.

2. **`key` must be stable and unique.** Array index as key causes bugs with reordering, deletion, and component state. Always use a unique ID.

3. **Avoid prop drilling past 2-3 levels.** If you're passing a prop through 3+ components that don't use it, use Context, Zustand, or component composition.

4. **`children` is a valid prop.** Prefer composition (`<Card><Content /></Card>`) over config props (`<Card content={<Content />} />`).

5. **Error boundaries only catch render errors.** They don't catch event handler errors, async errors, or errors in the error boundary itself. Use try/catch in event handlers.

6. **`React.memo` does a shallow comparison.** If you pass a new object or array as a prop, memo won't help. Memoize the prop value too.
