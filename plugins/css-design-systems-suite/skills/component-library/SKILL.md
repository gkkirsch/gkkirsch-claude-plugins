---
name: component-library
description: >
  Production component library patterns with composable, accessible components.
  Use when building reusable UI components, implementing compound component
  patterns, or creating design system primitives.
  Triggers: "component library", "reusable component", "compound component",
  "design system component", "button component", "input component", "modal".
  NOT for: CSS-only styling, design tokens, or Tailwind configuration.
version: 1.0.0
allowed-tools: Read, Grep, Glob, Bash
---

# Component Library Patterns

## Compound Component Pattern

```tsx
import { createContext, useContext, useState, ReactNode } from 'react';
import { cn } from '@/lib/utils';

// Accordion — compound component with shared state
interface AccordionContextType {
  openItems: Set<string>;
  toggle: (id: string) => void;
  type: 'single' | 'multiple';
}

const AccordionContext = createContext<AccordionContextType | null>(null);

function useAccordion() {
  const ctx = useContext(AccordionContext);
  if (!ctx) throw new Error('Accordion components must be used within <Accordion>');
  return ctx;
}

interface AccordionProps {
  type?: 'single' | 'multiple';
  defaultOpen?: string[];
  children: ReactNode;
  className?: string;
}

function Accordion({ type = 'single', defaultOpen = [], children, className }: AccordionProps) {
  const [openItems, setOpenItems] = useState<Set<string>>(new Set(defaultOpen));

  const toggle = (id: string) => {
    setOpenItems(prev => {
      const next = new Set(prev);
      if (next.has(id)) {
        next.delete(id);
      } else {
        if (type === 'single') next.clear();
        next.add(id);
      }
      return next;
    });
  };

  return (
    <AccordionContext.Provider value={{ openItems, toggle, type }}>
      <div className={cn('divide-y divide-gray-800', className)}>
        {children}
      </div>
    </AccordionContext.Provider>
  );
}

function AccordionItem({ id, children, className }: { id: string; children: ReactNode; className?: string }) {
  return <div className={cn('', className)} data-state={useAccordion().openItems.has(id) ? 'open' : 'closed'}>{children}</div>;
}

function AccordionTrigger({ id, children }: { id: string; children: ReactNode }) {
  const { openItems, toggle } = useAccordion();
  const isOpen = openItems.has(id);

  return (
    <button
      onClick={() => toggle(id)}
      aria-expanded={isOpen}
      aria-controls={`accordion-content-${id}`}
      className="flex w-full items-center justify-between py-4 text-left font-medium transition-all hover:underline"
    >
      {children}
      <svg
        className={cn('h-4 w-4 shrink-0 transition-transform duration-200', isOpen && 'rotate-180')}
        fill="none" viewBox="0 0 24 24" stroke="currentColor"
      >
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
      </svg>
    </button>
  );
}

function AccordionContent({ id, children }: { id: string; children: ReactNode }) {
  const { openItems } = useAccordion();
  const isOpen = openItems.has(id);

  return (
    <div
      id={`accordion-content-${id}`}
      role="region"
      aria-labelledby={`accordion-trigger-${id}`}
      className={cn(
        'overflow-hidden transition-all duration-200',
        isOpen ? 'max-h-96 pb-4' : 'max-h-0'
      )}
    >
      {children}
    </div>
  );
}

Accordion.Item = AccordionItem;
Accordion.Trigger = AccordionTrigger;
Accordion.Content = AccordionContent;
```

## Polymorphic Component (as prop)

```tsx
import { ElementType, ComponentPropsWithoutRef, ReactNode } from 'react';

// Type-safe "as" prop — render as any HTML element or component
type PolymorphicProps<E extends ElementType, P = {}> = P & {
  as?: E;
  children?: ReactNode;
} & Omit<ComponentPropsWithoutRef<E>, keyof P | 'as' | 'children'>;

function Text<E extends ElementType = 'p'>({
  as,
  size = 'md',
  weight = 'normal',
  className,
  children,
  ...props
}: PolymorphicProps<E, { size?: 'xs' | 'sm' | 'md' | 'lg' | 'xl'; weight?: 'normal' | 'medium' | 'bold' }>) {
  const Component = as ?? 'p';

  const sizes = { xs: 'text-xs', sm: 'text-sm', md: 'text-base', lg: 'text-lg', xl: 'text-xl' };
  const weights = { normal: 'font-normal', medium: 'font-medium', bold: 'font-bold' };

  return (
    <Component className={cn(sizes[size], weights[weight], className)} {...props}>
      {children}
    </Component>
  );
}

// Usage — fully type-safe, inherits element props
// <Text as="h1" size="xl" weight="bold">Title</Text>
// <Text as="span" size="sm">Inline text</Text>
// <Text as="a" href="/link" size="md">Link text</Text>  ← knows href is valid for <a>
```

## Accessible Form Components

```tsx
import { forwardRef, useId } from 'react';

interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  label: string;
  error?: string;
  hint?: string;
}

const Input = forwardRef<HTMLInputElement, InputProps>(
  ({ label, error, hint, className, id: providedId, ...props }, ref) => {
    const generatedId = useId();
    const id = providedId ?? generatedId;
    const errorId = `${id}-error`;
    const hintId = `${id}-hint`;

    return (
      <div className="space-y-1.5">
        <label htmlFor={id} className="block text-sm font-medium text-gray-200">
          {label}
          {props.required && <span className="text-red-400 ml-1" aria-hidden="true">*</span>}
        </label>

        {hint && (
          <p id={hintId} className="text-xs text-gray-400">{hint}</p>
        )}

        <input
          ref={ref}
          id={id}
          aria-invalid={!!error}
          aria-describedby={[error ? errorId : null, hint ? hintId : null].filter(Boolean).join(' ') || undefined}
          className={cn(
            'w-full rounded-md border bg-gray-900 px-3 py-2 text-sm transition-colors',
            'focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-offset-gray-950',
            error
              ? 'border-red-500 focus:ring-red-500'
              : 'border-gray-700 focus:ring-blue-500',
            className
          )}
          {...props}
        />

        {error && (
          <p id={errorId} role="alert" className="text-xs text-red-400">
            {error}
          </p>
        )}
      </div>
    );
  }
);
Input.displayName = 'Input';
```

## Modal / Dialog

```tsx
import { useEffect, useRef, ReactNode } from 'react';

interface DialogProps {
  open: boolean;
  onClose: () => void;
  title: string;
  children: ReactNode;
}

function Dialog({ open, onClose, title, children }: DialogProps) {
  const dialogRef = useRef<HTMLDialogElement>(null);
  const previousFocus = useRef<HTMLElement | null>(null);

  useEffect(() => {
    const dialog = dialogRef.current;
    if (!dialog) return;

    if (open) {
      previousFocus.current = document.activeElement as HTMLElement;
      dialog.showModal();
    } else {
      dialog.close();
      previousFocus.current?.focus();
    }
  }, [open]);

  // Close on backdrop click
  const handleBackdropClick = (e: React.MouseEvent) => {
    if (e.target === dialogRef.current) onClose();
  };

  // Close on Escape
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && open) onClose();
    };
    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [open, onClose]);

  return (
    <dialog
      ref={dialogRef}
      onClick={handleBackdropClick}
      className="backdrop:bg-black/50 bg-gray-900 text-gray-100 rounded-lg p-0 max-w-lg w-full shadow-xl"
    >
      <div className="p-6">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg font-semibold">{title}</h2>
          <button onClick={onClose} aria-label="Close dialog" className="text-gray-400 hover:text-white">
            <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
        {children}
      </div>
    </dialog>
  );
}
```

## Gotchas

1. **forwardRef with generic components** — React's forwardRef erases generic type parameters. For polymorphic components with ref forwarding, use a wrapper function or type assertion to preserve the generic.

2. **Context value object recreated every render** — `value={{ state, toggle }}` creates a new object reference each render, causing all consumers to re-render. Memoize with `useMemo()` or split into separate contexts for state vs dispatch.

3. **useId generates different IDs on server vs client** — if you're doing SSR with React 18+, `useId()` handles this correctly. But if you're on React 17 or using custom ID generation, you'll get hydration mismatches.

4. **Dialog element focus trap is browser-native but inconsistent** — `<dialog>.showModal()` provides native focus trapping, but browser implementations differ. Test across Chrome, Firefox, and Safari. Add a manual focus trap as fallback for non-supporting browsers.

5. **Compound components break with intermediary wrappers** — if someone wraps `<Accordion.Item>` in a `<div>`, the context traversal still works (React context), but CSS selectors targeting direct children break. Use context-based state, not DOM-based.

6. **Missing displayName on forwardRef components** — React DevTools shows "ForwardRef" instead of the component name. Always add `.displayName = 'ComponentName'` after the forwardRef call.
