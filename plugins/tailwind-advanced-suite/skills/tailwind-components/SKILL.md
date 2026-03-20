---
name: tailwind-components
description: >
  Tailwind CSS component architecture — class variance authority (cva), compound variants,
  slot-based components, headless UI integration, shadcn/ui patterns, and component composition.
  Triggers: "tailwind components", "cva", "class variance", "tailwind variants", "shadcn patterns",
  "tailwind slots", "headless ui tailwind", "tailwind component library".
  NOT for: design tokens (use tailwind-design-system), responsive layout (use tailwind-responsive).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Tailwind Component Architecture

## Class Variance Authority (cva)

```typescript
// lib/components.ts
import { cva, type VariantProps } from 'class-variance-authority';
import { cn } from '@/lib/utils';

// Define variant-driven component styles
const buttonVariants = cva(
  // Base styles (always applied)
  'inline-flex items-center justify-center gap-2 rounded-md font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-50',
  {
    variants: {
      variant: {
        default: 'bg-primary text-on-primary shadow-sm hover:bg-primary-hover active:bg-primary-active',
        secondary: 'bg-surface-raised text-on-surface shadow-sm border border-border hover:bg-surface-sunken',
        destructive: 'bg-error text-white shadow-sm hover:bg-error/90',
        outline: 'border border-border bg-transparent hover:bg-surface-raised',
        ghost: 'hover:bg-surface-raised',
        link: 'text-primary underline-offset-4 hover:underline',
      },
      size: {
        sm: 'h-8 px-3 text-xs',
        md: 'h-10 px-4 text-sm',
        lg: 'h-12 px-6 text-base',
        icon: 'h-10 w-10',
      },
    },
    // Compound variants (when multiple variants combine)
    compoundVariants: [
      {
        variant: 'outline',
        size: 'sm',
        className: 'h-7 px-2',
      },
      {
        variant: ['default', 'destructive'],
        size: 'lg',
        className: 'text-base font-semibold',
      },
    ],
    defaultVariants: {
      variant: 'default',
      size: 'md',
    },
  }
);

// Type-safe component
interface ButtonProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement>,
    VariantProps<typeof buttonVariants> {
  asChild?: boolean;
}

function Button({ className, variant, size, asChild, ...props }: ButtonProps) {
  const Comp = asChild ? Slot : 'button';
  return (
    <Comp
      className={cn(buttonVariants({ variant, size, className }))}
      {...props}
    />
  );
}

// Usage
<Button variant="destructive" size="lg">Delete</Button>
<Button variant="ghost" size="icon"><TrashIcon /></Button>
<Button asChild><a href="/signup">Sign Up</a></Button>
```

## Badge Variants

```typescript
const badgeVariants = cva(
  'inline-flex items-center rounded-full border px-2.5 py-0.5 text-xs font-semibold transition-colors',
  {
    variants: {
      variant: {
        default: 'border-transparent bg-primary text-on-primary',
        secondary: 'border-transparent bg-surface-raised text-on-surface',
        success: 'border-transparent bg-success/10 text-success',
        warning: 'border-transparent bg-warning/10 text-warning',
        error: 'border-transparent bg-error/10 text-error',
        outline: 'text-on-surface border-border',
      },
    },
    defaultVariants: {
      variant: 'default',
    },
  }
);

interface BadgeProps
  extends React.HTMLAttributes<HTMLDivElement>,
    VariantProps<typeof badgeVariants> {}

function Badge({ className, variant, ...props }: BadgeProps) {
  return <div className={cn(badgeVariants({ variant }), className)} {...props} />;
}
```

## Input Components

```typescript
const inputVariants = cva(
  'flex w-full rounded-md border bg-transparent text-sm transition-colors file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-on-surface-subtle focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50',
  {
    variants: {
      size: {
        sm: 'h-8 px-2.5 text-xs',
        md: 'h-10 px-3 text-sm',
        lg: 'h-12 px-4 text-base',
      },
      state: {
        default: 'border-border',
        error: 'border-error focus-visible:ring-error',
        success: 'border-success focus-visible:ring-success',
      },
    },
    defaultVariants: {
      size: 'md',
      state: 'default',
    },
  }
);

// Form field with label, input, description, error
function FormField({
  label,
  description,
  error,
  children,
}: {
  label: string;
  description?: string;
  error?: string;
  children: React.ReactNode;
}) {
  return (
    <div className="space-y-2">
      <label className="text-sm font-medium text-on-surface">{label}</label>
      {children}
      {description && !error && (
        <p className="text-xs text-on-surface-muted">{description}</p>
      )}
      {error && (
        <p className="text-xs text-error">{error}</p>
      )}
    </div>
  );
}
```

## Card Pattern

```typescript
// Composable card with named slots
function Card({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return (
    <div
      className={cn(
        'rounded-xl border border-border bg-surface shadow-soft',
        className
      )}
      {...props}
    />
  );
}

function CardHeader({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return <div className={cn('flex flex-col space-y-1.5 p-6', className)} {...props} />;
}

function CardTitle({ className, ...props }: React.HTMLAttributes<HTMLHeadingElement>) {
  return <h3 className={cn('text-xl font-semibold leading-none tracking-tight', className)} {...props} />;
}

function CardDescription({ className, ...props }: React.HTMLAttributes<HTMLParagraphElement>) {
  return <p className={cn('text-sm text-on-surface-muted', className)} {...props} />;
}

function CardContent({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return <div className={cn('p-6 pt-0', className)} {...props} />;
}

function CardFooter({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return <div className={cn('flex items-center p-6 pt-0', className)} {...props} />;
}

// Usage
<Card>
  <CardHeader>
    <CardTitle>Team Members</CardTitle>
    <CardDescription>Manage your team and permissions.</CardDescription>
  </CardHeader>
  <CardContent>
    <UserList users={users} />
  </CardContent>
  <CardFooter className="justify-between">
    <span className="text-sm text-on-surface-muted">{users.length} members</span>
    <Button size="sm">Invite</Button>
  </CardFooter>
</Card>
```

## Radix UI + Tailwind Integration

```typescript
// Dialog (modal) with Radix primitives + Tailwind styling
import * as DialogPrimitive from '@radix-ui/react-dialog';
import { X } from 'lucide-react';

const DialogOverlay = forwardRef<
  React.ElementRef<typeof DialogPrimitive.Overlay>,
  React.ComponentPropsWithoutRef<typeof DialogPrimitive.Overlay>
>(({ className, ...props }, ref) => (
  <DialogPrimitive.Overlay
    ref={ref}
    className={cn(
      'fixed inset-0 z-50 bg-black/60 backdrop-blur-sm',
      'data-[state=open]:animate-in data-[state=closed]:animate-out',
      'data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0',
      className
    )}
    {...props}
  />
));

const DialogContent = forwardRef<
  React.ElementRef<typeof DialogPrimitive.Content>,
  React.ComponentPropsWithoutRef<typeof DialogPrimitive.Content>
>(({ className, children, ...props }, ref) => (
  <DialogPrimitive.Portal>
    <DialogOverlay />
    <DialogPrimitive.Content
      ref={ref}
      className={cn(
        'fixed left-1/2 top-1/2 z-50 w-full max-w-lg -translate-x-1/2 -translate-y-1/2',
        'rounded-xl border border-border bg-surface p-6 shadow-overlay',
        'data-[state=open]:animate-in data-[state=closed]:animate-out',
        'data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0',
        'data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95',
        'data-[state=closed]:slide-out-to-left-1/2 data-[state=closed]:slide-out-to-top-[48%]',
        'data-[state=open]:slide-in-from-left-1/2 data-[state=open]:slide-in-from-top-[48%]',
        className
      )}
      {...props}
    >
      {children}
      <DialogPrimitive.Close className="absolute right-4 top-4 rounded-sm opacity-70 hover:opacity-100 transition-opacity">
        <X className="h-4 w-4" />
      </DialogPrimitive.Close>
    </DialogPrimitive.Content>
  </DialogPrimitive.Portal>
));

// Usage
<Dialog>
  <DialogTrigger asChild>
    <Button variant="outline">Edit Profile</Button>
  </DialogTrigger>
  <DialogContent>
    <DialogHeader>
      <DialogTitle>Edit Profile</DialogTitle>
      <DialogDescription>Make changes to your profile.</DialogDescription>
    </DialogHeader>
    <form>...</form>
  </DialogContent>
</Dialog>
```

## Dropdown Menu Pattern

```typescript
import * as DropdownMenuPrimitive from '@radix-ui/react-dropdown-menu';
import { Check, ChevronRight, Circle } from 'lucide-react';

const DropdownMenuContent = forwardRef<
  React.ElementRef<typeof DropdownMenuPrimitive.Content>,
  React.ComponentPropsWithoutRef<typeof DropdownMenuPrimitive.Content>
>(({ className, sideOffset = 4, ...props }, ref) => (
  <DropdownMenuPrimitive.Portal>
    <DropdownMenuPrimitive.Content
      ref={ref}
      sideOffset={sideOffset}
      className={cn(
        'z-50 min-w-[8rem] overflow-hidden rounded-lg border border-border bg-surface p-1 shadow-floating',
        'data-[state=open]:animate-in data-[state=closed]:animate-out',
        'data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0',
        'data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95',
        'data-[side=bottom]:slide-in-from-top-2 data-[side=top]:slide-in-from-bottom-2',
        className
      )}
      {...props}
    />
  </DropdownMenuPrimitive.Portal>
));

const DropdownMenuItem = forwardRef<
  React.ElementRef<typeof DropdownMenuPrimitive.Item>,
  React.ComponentPropsWithoutRef<typeof DropdownMenuPrimitive.Item> & {
    inset?: boolean;
  }
>(({ className, inset, ...props }, ref) => (
  <DropdownMenuPrimitive.Item
    ref={ref}
    className={cn(
      'relative flex cursor-pointer select-none items-center gap-2 rounded-md px-2 py-1.5 text-sm outline-none',
      'transition-colors focus:bg-surface-raised focus:text-on-surface',
      'data-[disabled]:pointer-events-none data-[disabled]:opacity-50',
      inset && 'pl-8',
      className
    )}
    {...props}
  />
));
```

## Data Display Components

```typescript
// Stat card
function StatCard({
  title,
  value,
  change,
  trend,
}: {
  title: string;
  value: string;
  change: string;
  trend: 'up' | 'down' | 'neutral';
}) {
  return (
    <Card>
      <CardContent className="p-6">
        <p className="text-sm text-on-surface-muted">{title}</p>
        <p className="mt-1 text-3xl font-semibold tracking-tight">{value}</p>
        <p className={cn(
          'mt-1 flex items-center gap-1 text-sm',
          trend === 'up' && 'text-success',
          trend === 'down' && 'text-error',
          trend === 'neutral' && 'text-on-surface-muted',
        )}>
          {trend === 'up' && '↑'}
          {trend === 'down' && '↓'}
          {change}
        </p>
      </CardContent>
    </Card>
  );
}

// Avatar with status indicator
function Avatar({
  src,
  alt,
  fallback,
  status,
  size = 'md',
}: {
  src?: string;
  alt: string;
  fallback: string;
  status?: 'online' | 'offline' | 'busy';
  size?: 'sm' | 'md' | 'lg';
}) {
  const sizes = {
    sm: 'h-8 w-8 text-xs',
    md: 'h-10 w-10 text-sm',
    lg: 'h-14 w-14 text-base',
  };

  const statusColors = {
    online: 'bg-success',
    offline: 'bg-on-surface-subtle',
    busy: 'bg-error',
  };

  return (
    <div className="relative inline-block">
      <div className={cn(
        'rounded-full overflow-hidden bg-surface-raised flex items-center justify-center font-medium',
        sizes[size],
      )}>
        {src ? (
          <img src={src} alt={alt} className="h-full w-full object-cover" />
        ) : (
          <span>{fallback}</span>
        )}
      </div>
      {status && (
        <span className={cn(
          'absolute bottom-0 right-0 block rounded-full ring-2 ring-surface',
          statusColors[status],
          size === 'sm' ? 'h-2 w-2' : 'h-3 w-3',
        )} />
      )}
    </div>
  );
}
```

## The cn() Utility

```typescript
// lib/utils.ts — the foundation of every component
import { type ClassValue, clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

// What it does:
// 1. clsx: merges strings, arrays, objects, handles conditionals
//    clsx('px-2', ['py-1'], { 'bg-red': isError }) → 'px-2 py-1 bg-red'
//
// 2. twMerge: resolves Tailwind conflicts (last wins)
//    twMerge('px-2 px-4') → 'px-4'
//    twMerge('text-red-500 text-blue-500') → 'text-blue-500'
//    twMerge('p-4 px-6') → 'p-4 px-6' (px-6 overrides p-4's x only)
//
// Without twMerge, both classes apply and first one wins (CSS specificity)
// With twMerge, last one wins (intuitive override behavior)
```

## Component Slot Pattern (Radix asChild)

```typescript
import { Slot } from '@radix-ui/react-slot';

interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  asChild?: boolean;
  variant?: 'default' | 'outline' | 'ghost';
}

function Button({ asChild, variant = 'default', className, ...props }: ButtonProps) {
  // When asChild=true, renders children directly with button styles
  // instead of wrapping in <button>
  const Comp = asChild ? Slot : 'button';
  return <Comp className={cn(buttonVariants({ variant }), className)} {...props} />;
}

// Usage: button as a link
<Button asChild variant="outline">
  <a href="/dashboard">Go to Dashboard</a>
</Button>

// The <a> receives all button styles + its own href
// Renders: <a href="/dashboard" class="inline-flex items-center ... border ...">
```

## Gotchas

1. **Always use cn() for className props.** Without it, consumer overrides can't work: `<Button className="mt-4">` would conflict with the component's padding without twMerge resolving the conflict.

2. **cva compoundVariants order matters.** Later compound variants override earlier ones when they match the same conditions. Put your most specific overrides last.

3. **Radix data-attributes for animations.** Use `data-[state=open]:animate-in` not CSS `:is()`. Radix manages state attributes. Always include both open and closed animations for smooth transitions.

4. **Don't mix Tailwind spacing with custom CSS spacing.** If you use `gap-4` in some places and `margin: 1rem` in others, the system becomes inconsistent. Pick one approach (Tailwind utilities) and stick with it.

5. **forwardRef is required for Radix primitives.** All Radix components expect refs to be forwarded. Forgetting forwardRef causes runtime warnings and broken functionality (portals, focus management).

6. **Slot merges props, not replaces.** When using `asChild`/`Slot`, event handlers are merged (both fire), classNames are merged (via cn), but non-mergeable props (like `id`) use the child's value.
