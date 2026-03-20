# Tailwind CSS Advanced Cheatsheet

## v4 @theme Quick Reference

```css
@import "tailwindcss";
@theme {
  --color-brand-500: oklch(0.55 0.20 250);
  --font-sans: 'Inter', system-ui, sans-serif;
  --spacing-18: 4.5rem;
  --breakpoint-xs: 475px;
  --shadow-soft: 0 1px 3px oklch(0 0 0 / 0.04);
  --animate-fade-in: fade-in 0.3s ease-out;
  --radius-4xl: 2rem;
}
@theme extend { /* adds to defaults instead of replacing */ }
@plugin "@tailwindcss/typography";
@source "../shared/**/*.tsx";
@source inline("bg-red-500 bg-blue-500");
```

## OKLCH Color Formula

```
oklch(lightness chroma hue)
  lightness: 0 (black) → 1 (white)
  chroma:    0 (gray) → ~0.3 (vivid)
  hue:       red=25, orange=70, yellow=90, green=145, blue=250, purple=300

Full palette from single hue:
  50:  oklch(0.97 0.01 HUE)
  100: oklch(0.93 0.03 HUE)
  200: oklch(0.87 0.06 HUE)
  300: oklch(0.78 0.10 HUE)
  400: oklch(0.68 0.15 HUE)
  500: oklch(0.55 0.20 HUE)   ← base
  600: oklch(0.48 0.20 HUE)
  700: oklch(0.40 0.18 HUE)
  800: oklch(0.33 0.14 HUE)
  900: oklch(0.25 0.10 HUE)
  950: oklch(0.18 0.07 HUE)
```

## Semantic Token Pattern

```css
:root {
  --surface: oklch(1 0 0);
  --on-surface: oklch(0.15 0.01 250);
  --primary: oklch(0.55 0.20 250);
  --on-primary: oklch(1 0 0);
  --error: oklch(0.55 0.20 25);
  --border: oklch(0.90 0.01 250);
}
.dark {
  --surface: oklch(0.15 0.01 250);
  --on-surface: oklch(0.93 0.01 250);
  --primary: oklch(0.68 0.18 250);
  --border: oklch(0.25 0.01 250);
}
```

## cva Pattern

```typescript
import { cva, type VariantProps } from 'class-variance-authority';
const button = cva('base-classes', {
  variants: {
    variant: { default: '...', outline: '...' },
    size: { sm: 'h-8 px-3', md: 'h-10 px-4', lg: 'h-12 px-6' },
  },
  compoundVariants: [
    { variant: 'outline', size: 'sm', className: 'h-7 px-2' },
  ],
  defaultVariants: { variant: 'default', size: 'md' },
});
```

## cn() Utility

```typescript
import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';
export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}
// twMerge('px-2 px-4') → 'px-4' (last wins)
```

## Container Queries

```html
<div class="@container">
  <div class="flex-col @md:flex-row @lg:grid-cols-3">
  </div>
</div>
<!-- Named: @container/sidebar → @md/sidebar:hidden -->
```

## Responsive Grid Patterns

```html
<!-- Auto-fill cards -->
<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">

<!-- CSS auto-fill (no breakpoints) -->
<div style="grid-template-columns: repeat(auto-fill, minmax(280px, 1fr))">

<!-- Dashboard: sidebar + main -->
<div class="grid grid-cols-1 md:grid-cols-[280px_1fr]">

<!-- Holy grail -->
<div class="grid min-h-dvh grid-rows-[auto_1fr_auto]">

<!-- Masonry (CSS columns) -->
<div class="columns-1 sm:columns-2 lg:columns-3 gap-6 space-y-6">
```

## Viewport Units

```
dvh = dynamic viewport height (accounts for mobile browser chrome)
svh = small viewport height (URL bar visible)
lvh = large viewport height (URL bar hidden)
Use: min-h-dvh (not h-dvh to avoid jank)
```

## Sticky Patterns

```html
<!-- Blur header -->
<header class="sticky top-0 z-40 bg-surface/80 backdrop-blur-sm border-b">

<!-- Sticky sidebar -->
<aside class="sticky top-20 self-start max-h-[calc(100dvh-5rem)] overflow-y-auto">

<!-- Mobile sticky CTA -->
<div class="fixed bottom-0 inset-x-0 z-40 md:hidden p-4 pb-[env(safe-area-inset-bottom)]">
```

## Radix Data Attributes

```html
data-[state=open]:animate-in
data-[state=closed]:animate-out
data-[state=closed]:fade-out-0
data-[state=open]:fade-in-0
data-[state=closed]:zoom-out-95
data-[state=open]:zoom-in-95
data-[side=bottom]:slide-in-from-top-2
data-[disabled]:pointer-events-none
data-[disabled]:opacity-50
```

## Accessibility Modifiers

```html
motion-reduce:transition-none    — prefers-reduced-motion
motion-reduce:animate-none
contrast-more:border-2          — prefers-contrast
contrast-more:font-semibold
focus-visible:ring-2            — keyboard focus only
focus-visible:ring-ring
```

## Performance Tips

```
backdrop-blur fallback:
  bg-surface/95 supports-[backdrop-filter]:bg-surface/60 backdrop-blur-sm

Animated elements:
  will-change-transform (add before animation, remove after)

Off-screen content:
  content-visibility: auto (CSS property, not a Tailwind class)

Dynamic classes:
  ✗ `bg-${color}-500`          (purged at build time)
  ✓ { red: 'bg-red-500' }[c]  (complete strings preserved)
```

## v3 → v4 Migration Map

```
tailwind.config.js              →  @theme { }
theme.extend.colors.x           →  --color-x
theme.extend.fontFamily.sans    →  --font-sans
theme.extend.spacing['18']      →  --spacing-18
theme.screens.xl                →  --breakpoint-xl
darkMode: 'class'               →  (default in v4)
content: ['./src/**']           →  (auto-detected in v4)
plugins: [require('@tw/forms')] →  @plugin "@tailwindcss/forms"
```

## Scroll Snap

```html
<!-- Horizontal gallery -->
<div class="flex snap-x snap-mandatory overflow-x-auto gap-4 scrollbar-hide">
  <div class="snap-center shrink-0">...</div>
</div>

<!-- Full-page sections -->
<div class="h-dvh snap-y snap-mandatory overflow-y-auto">
  <section class="snap-start h-dvh">...</section>
</div>
```

## Common Component Classes

```
Card:     rounded-xl border border-border bg-surface shadow-soft
Input:    h-10 w-full rounded-md border bg-transparent px-3 text-sm
Badge:    inline-flex rounded-full border px-2.5 py-0.5 text-xs font-semibold
Avatar:   h-10 w-10 rounded-full overflow-hidden bg-surface-raised
Overlay:  fixed inset-0 z-50 bg-black/60 backdrop-blur-sm
Modal:    fixed left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 rounded-xl border p-6 shadow-overlay
Toast:    rounded-lg border bg-surface p-4 shadow-elevated
```
