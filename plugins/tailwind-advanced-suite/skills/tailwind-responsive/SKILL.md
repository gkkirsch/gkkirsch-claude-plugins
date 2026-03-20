---
name: tailwind-responsive
description: >
  Tailwind CSS responsive design — container queries, fluid layouts, responsive grids,
  mobile-first patterns, dynamic viewport units, and production optimization.
  Triggers: "tailwind responsive", "tailwind container queries", "tailwind grid layout",
  "tailwind mobile first", "tailwind fluid", "tailwind @container", "tailwind production".
  NOT for: design tokens (use tailwind-design-system), component variants (use tailwind-components).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Tailwind Responsive & Production Patterns

## Container Queries

```typescript
// Parent: mark as container
<div className="@container">
  <div className="flex flex-col @md:flex-row @lg:grid @lg:grid-cols-3 gap-4">
    <Card />
    <Card />
    <Card />
  </div>
</div>

// Named containers (when nesting)
<div className="@container/sidebar">
  <div className="@container/main">
    <div className="@md/sidebar:hidden @lg/main:grid-cols-2">
      {/* Responds to sidebar container width */}
    </div>
  </div>
</div>
```

```css
/* Container query breakpoints (Tailwind v4) */
@theme {
  --container-3xs: 16rem;   /* 256px */
  --container-2xs: 18rem;   /* 288px */
  --container-xs: 20rem;    /* 320px */
  --container-sm: 24rem;    /* 384px */
  --container-md: 28rem;    /* 448px */
  --container-lg: 32rem;    /* 512px */
  --container-xl: 36rem;    /* 576px */
  --container-2xl: 42rem;   /* 672px */
}
```

```typescript
// Responsive card that adapts to its container, not the viewport
function ProductCard({ product }: { product: Product }) {
  return (
    <div className="@container">
      <div className={cn(
        'flex flex-col rounded-xl border border-border overflow-hidden',
        '@sm:flex-row',        // side-by-side when container >= 384px
        '@lg:flex-col',        // back to stacked when container >= 512px (grid layout)
      )}>
        <div className={cn(
          'aspect-square @sm:aspect-auto @sm:w-48 @lg:aspect-video @lg:w-full',
          'bg-surface-sunken overflow-hidden',
        )}>
          <img
            src={product.image}
            alt={product.name}
            className="h-full w-full object-cover"
          />
        </div>
        <div className="flex flex-col gap-2 p-4 @sm:p-5">
          <h3 className="font-semibold text-on-surface @sm:text-lg">{product.name}</h3>
          <p className="text-sm text-on-surface-muted line-clamp-2 @sm:line-clamp-3">
            {product.description}
          </p>
          <p className="text-lg font-bold mt-auto">${product.price}</p>
        </div>
      </div>
    </div>
  );
}
```

## Responsive Grid Patterns

```typescript
// Auto-fill grid (cards fill available space)
<div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
  {items.map(item => <Card key={item.id} item={item} />)}
</div>

// CSS Grid auto-fill (no breakpoints needed)
<div className="grid gap-6" style={{
  gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))'
}}>
  {items.map(item => <Card key={item.id} item={item} />)}
</div>

// Dashboard layout with sidebar
<div className="grid min-h-screen grid-cols-1 md:grid-cols-[280px_1fr]">
  <aside className="hidden md:block border-r border-border p-6">
    <Sidebar />
  </aside>
  <main className="p-6 md:p-8 lg:p-10">
    <Outlet />
  </main>
</div>

// Holy grail layout
<div className="grid min-h-screen grid-rows-[auto_1fr_auto]">
  <header className="sticky top-0 z-40 border-b border-border bg-surface/80 backdrop-blur-sm">
    <nav className="mx-auto flex h-16 max-w-7xl items-center px-4 sm:px-6 lg:px-8">
      <Logo />
      <NavLinks className="hidden md:flex" />
      <MobileMenu className="md:hidden" />
    </nav>
  </header>
  <main className="mx-auto w-full max-w-7xl px-4 sm:px-6 lg:px-8 py-8">
    <Outlet />
  </main>
  <footer className="border-t border-border bg-surface-raised">
    <FooterContent />
  </footer>
</div>

// Masonry-like layout (CSS columns)
<div className="columns-1 sm:columns-2 lg:columns-3 xl:columns-4 gap-6 space-y-6">
  {items.map(item => (
    <div key={item.id} className="break-inside-avoid">
      <Card item={item} />
    </div>
  ))}
</div>
```

## Mobile-First Patterns

```typescript
// Navigation: mobile hamburger → desktop inline
function Navigation() {
  const [open, setOpen] = useState(false);

  return (
    <nav className="flex items-center justify-between h-16 px-4">
      <Logo />

      {/* Desktop nav — hidden on mobile */}
      <div className="hidden md:flex items-center gap-6">
        <NavLink href="/features">Features</NavLink>
        <NavLink href="/pricing">Pricing</NavLink>
        <NavLink href="/docs">Docs</NavLink>
        <Button size="sm">Sign Up</Button>
      </div>

      {/* Mobile menu button — hidden on desktop */}
      <button
        className="md:hidden p-2 rounded-md hover:bg-surface-raised"
        onClick={() => setOpen(!open)}
      >
        {open ? <X className="h-5 w-5" /> : <Menu className="h-5 w-5" />}
      </button>

      {/* Mobile menu panel */}
      {open && (
        <div className="absolute inset-x-0 top-16 z-50 md:hidden border-b border-border bg-surface shadow-elevated">
          <div className="flex flex-col p-4 gap-2">
            <NavLink href="/features" className="px-3 py-2 rounded-md hover:bg-surface-raised">
              Features
            </NavLink>
            <NavLink href="/pricing" className="px-3 py-2 rounded-md hover:bg-surface-raised">
              Pricing
            </NavLink>
            <NavLink href="/docs" className="px-3 py-2 rounded-md hover:bg-surface-raised">
              Docs
            </NavLink>
            <Button className="mt-2">Sign Up</Button>
          </div>
        </div>
      )}
    </nav>
  );
}

// Responsive table → card list on mobile
function DataTable({ rows }: { rows: Row[] }) {
  return (
    <>
      {/* Table view (desktop) */}
      <table className="hidden md:table w-full">
        <thead>
          <tr className="border-b border-border text-left text-sm text-on-surface-muted">
            <th className="pb-3 font-medium">Name</th>
            <th className="pb-3 font-medium">Status</th>
            <th className="pb-3 font-medium">Amount</th>
            <th className="pb-3 font-medium">Date</th>
          </tr>
        </thead>
        <tbody>
          {rows.map(row => (
            <tr key={row.id} className="border-b border-border last:border-0">
              <td className="py-3">{row.name}</td>
              <td className="py-3"><Badge variant={row.status}>{row.status}</Badge></td>
              <td className="py-3">{row.amount}</td>
              <td className="py-3 text-on-surface-muted">{row.date}</td>
            </tr>
          ))}
        </tbody>
      </table>

      {/* Card view (mobile) */}
      <div className="md:hidden space-y-3">
        {rows.map(row => (
          <div key={row.id} className="rounded-lg border border-border p-4 space-y-2">
            <div className="flex justify-between items-center">
              <span className="font-medium">{row.name}</span>
              <Badge variant={row.status}>{row.status}</Badge>
            </div>
            <div className="flex justify-between text-sm text-on-surface-muted">
              <span>{row.amount}</span>
              <span>{row.date}</span>
            </div>
          </div>
        ))}
      </div>
    </>
  );
}
```

## Dynamic Viewport Units

```css
/* Modern viewport units for mobile browsers */
.hero {
  /* dvh accounts for mobile browser chrome (URL bar) */
  min-height: 100dvh;
}

.sticky-footer {
  /* svh = smallest viewport (URL bar visible) */
  min-height: 100svh;
}

.full-screen-modal {
  /* lvh = largest viewport (URL bar hidden) */
  height: 100lvh;
}
```

```typescript
// Hero section with safe viewport height
<section className="min-h-dvh flex flex-col items-center justify-center px-4 text-center">
  <h1 className="text-fluid-hero font-bold tracking-tight">
    Build faster with AI
  </h1>
  <p className="mt-6 max-w-2xl text-lg text-on-surface-muted">
    Ship production-ready code in minutes, not weeks.
  </p>
  <div className="mt-10 flex flex-col sm:flex-row gap-4">
    <Button size="lg">Get Started</Button>
    <Button size="lg" variant="outline">Watch Demo</Button>
  </div>
</section>
```

## Scroll Snap

```typescript
// Horizontal scroll snap gallery
<div className="flex snap-x snap-mandatory overflow-x-auto gap-4 pb-4 -mx-4 px-4 scrollbar-hide">
  {images.map((img, i) => (
    <div
      key={i}
      className="snap-center shrink-0 first:pl-4 last:pr-4"
    >
      <img
        src={img.src}
        alt={img.alt}
        className="h-64 w-80 rounded-xl object-cover"
      />
    </div>
  ))}
</div>

// Vertical page-snap sections
<div className="h-dvh snap-y snap-mandatory overflow-y-auto">
  <section className="snap-start h-dvh flex items-center justify-center">
    <Hero />
  </section>
  <section className="snap-start h-dvh flex items-center justify-center">
    <Features />
  </section>
  <section className="snap-start h-dvh flex items-center justify-center">
    <Pricing />
  </section>
</div>
```

## Sticky Patterns

```typescript
// Sticky header with blur
<header className="sticky top-0 z-40 border-b border-border bg-surface/80 backdrop-blur-sm">
  <Nav />
</header>

// Sticky sidebar with scroll
<div className="grid grid-cols-[280px_1fr] gap-8">
  <aside className="sticky top-20 self-start max-h-[calc(100dvh-5rem)] overflow-y-auto">
    <TableOfContents />
  </aside>
  <main>
    <Content />
  </main>
</div>

// Sticky footer CTA on mobile
<div className="fixed bottom-0 inset-x-0 z-40 md:hidden border-t border-border bg-surface p-4 safe-area-pb">
  <Button className="w-full" size="lg">Add to Cart — $29.99</Button>
</div>

// Safe area for notched devices
<div className="pb-[env(safe-area-inset-bottom)]">
  <BottomNav />
</div>
```

## Production Optimization

```css
/* Tailwind v4: automatic content detection */
/* No content config needed — scans project files automatically */

/* Explicit source paths (when auto-detection isn't enough) */
@source "../node_modules/@my-org/ui/dist/**/*.js";

/* Safelist classes that appear in dynamic strings */
@source inline("bg-red-500 bg-blue-500 bg-green-500 text-red-500 text-blue-500 text-green-500");
```

```typescript
// Dynamic classes: use complete class names
// BAD: Tailwind can't find these
const color = 'red';
<div className={`bg-${color}-500`} />  // purged!

// GOOD: complete class strings
const bgColors: Record<string, string> = {
  red: 'bg-red-500',
  blue: 'bg-blue-500',
  green: 'bg-green-500',
};
<div className={bgColors[color]} />  // preserved!

// GOOD: use @source inline to safelist
// In your CSS: @source inline("bg-red-500 bg-blue-500 bg-green-500");
```

## CSS Layers (v4)

```css
@import "tailwindcss";

/* Tailwind v4 layer order:
   @layer theme    — design tokens (@theme)
   @layer base     — resets, defaults
   @layer components — .card, .btn patterns
   @layer utilities  — Tailwind utility classes (highest specificity)
*/

@layer base {
  body {
    @apply bg-surface text-on-surface antialiased;
  }

  /* Focus visible for keyboard navigation only */
  :focus-visible {
    @apply outline-2 outline-offset-2 outline-ring;
  }

  /* Remove focus ring for mouse users */
  :focus:not(:focus-visible) {
    outline: none;
  }
}

@layer components {
  /* Custom components sit between base and utilities */
  .prose-custom {
    @apply prose prose-neutral dark:prose-invert max-w-none;
    @apply prose-headings:scroll-mt-20;
    @apply prose-a:text-primary prose-a:no-underline hover:prose-a:underline;
    @apply prose-code:before:content-[''] prose-code:after:content-[''];
    @apply prose-code:bg-surface-sunken prose-code:rounded prose-code:px-1.5 prose-code:py-0.5;
  }
}
```

## Performance Checklist

```
Build Time:
  □ Use Tailwind v4 (Oxide engine — 5-10x faster builds)
  □ Remove unused @source paths
  □ Avoid @apply in hot paths (generates extra CSS)

Runtime:
  □ Use container queries (@container) over resize observers
  □ Prefer CSS transitions over JS animations
  □ Use will-change-transform on animated elements
  □ Avoid layout thrashing (transform > top/left)
  □ Use content-visibility: auto on off-screen sections

File Size:
  □ Don't safelist classes you don't need
  □ Use @source inline sparingly
  □ Split CSS with @layer for critical/non-critical
  □ Enable gzip/brotli compression on server

Accessibility:
  □ prefers-reduced-motion: motion-reduce:transition-none
  □ prefers-contrast: contrast-more:border-2
  □ Touch targets: min 44x44px on mobile (min-h-11 min-w-11)
  □ Focus indicators: focus-visible:ring-2
```

```typescript
// Reduced motion support
<div className={cn(
  'transition-all duration-300',
  'motion-reduce:transition-none motion-reduce:animate-none',
)}>
  <AnimatedContent />
</div>

// High contrast support
<button className={cn(
  'border border-transparent',
  'contrast-more:border-on-surface contrast-more:font-semibold',
)}>
  Submit
</button>
```

## Gotchas

1. **Container queries require @container on parent.** The `@md:flex-row` won't work unless an ancestor has `@container`. Easy to miss when refactoring — the query silently fails.

2. **dvh on desktop Safari can be janky.** Use `min-h-dvh` not `h-dvh` for full-page sections. The dynamic unit changes as the browser chrome hides/shows, causing repaints on scroll.

3. **Dynamic class names are purged.** Template literals like `` `bg-${color}-500` `` don't work — Tailwind needs to see the complete class string at build time. Use a lookup object or @source inline.

4. **@apply doesn't work with responsive/state variants.** You can't write `@apply hover:bg-blue-500` in a @layer. Write the raw CSS property or use the utility class in HTML.

5. **snap-mandatory + overflow-auto can trap scroll.** On mobile, if a snap section is taller than viewport, users can't scroll past it. Use `snap-proximity` instead, or ensure sections fit viewport.

6. **backdrop-blur has performance cost.** On lower-end devices, `backdrop-blur-sm` on sticky headers can cause scroll jank. Use a solid background fallback: `bg-surface/95 supports-[backdrop-filter]:bg-surface/60 backdrop-blur-sm`.
