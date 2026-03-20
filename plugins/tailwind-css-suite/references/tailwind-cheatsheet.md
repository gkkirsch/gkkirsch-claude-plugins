# Tailwind CSS Quick Reference

## Spacing Scale

| Class | Value | px |
|-------|-------|----|
| `0` | 0 | 0 |
| `px` | 1px | 1 |
| `0.5` | 0.125rem | 2 |
| `1` | 0.25rem | 4 |
| `1.5` | 0.375rem | 6 |
| `2` | 0.5rem | 8 |
| `3` | 0.75rem | 12 |
| `4` | 1rem | 16 |
| `5` | 1.25rem | 20 |
| `6` | 1.5rem | 24 |
| `8` | 2rem | 32 |
| `10` | 2.5rem | 40 |
| `12` | 3rem | 48 |
| `16` | 4rem | 64 |
| `20` | 5rem | 80 |
| `24` | 6rem | 96 |

Used with: `p-`, `m-`, `gap-`, `space-x-`, `space-y-`, `w-`, `h-`, `top-`, `left-`, etc.

## Breakpoints

| Prefix | Min Width | CSS |
|--------|-----------|-----|
| `sm:` | 640px | `@media (min-width: 640px)` |
| `md:` | 768px | `@media (min-width: 768px)` |
| `lg:` | 1024px | `@media (min-width: 1024px)` |
| `xl:` | 1280px | `@media (min-width: 1280px)` |
| `2xl:` | 1536px | `@media (min-width: 1536px)` |

Mobile-first: base styles apply to all, prefix overrides upward.

## Flexbox Patterns

| Layout | Classes |
|--------|---------|
| Horizontal row | `flex gap-4` |
| Vertical stack | `flex flex-col gap-4` |
| Center everything | `flex items-center justify-center` |
| Space between | `flex items-center justify-between` |
| Wrap items | `flex flex-wrap gap-4` |
| Push last right | `flex gap-4` + child: `ml-auto` |
| Equal width items | `flex` + children: `flex-1` |
| Shrink-proof item | `flex-shrink-0` or `shrink-0` |

## Grid Patterns

| Layout | Classes |
|--------|---------|
| 2 columns | `grid grid-cols-2 gap-4` |
| 3 columns | `grid grid-cols-3 gap-6` |
| Auto-fit responsive | `grid grid-cols-[repeat(auto-fit,minmax(250px,1fr))] gap-4` |
| Sidebar + main | `grid grid-cols-[250px_1fr] gap-6` |
| Span 2 columns | child: `col-span-2` |
| Full width row | child: `col-span-full` |

## Typography

| Class | Effect |
|-------|--------|
| `text-xs` | 12px |
| `text-sm` | 14px |
| `text-base` | 16px |
| `text-lg` | 18px |
| `text-xl` | 20px |
| `text-2xl` | 24px |
| `text-3xl` | 30px |
| `text-4xl` | 36px |
| `font-normal` | 400 |
| `font-medium` | 500 |
| `font-semibold` | 600 |
| `font-bold` | 700 |
| `leading-tight` | 1.25 |
| `leading-normal` | 1.5 |
| `leading-relaxed` | 1.625 |
| `tracking-tight` | -0.025em |
| `tracking-wide` | 0.025em |
| `truncate` | overflow: hidden + text-overflow: ellipsis + white-space: nowrap |
| `line-clamp-2` | Clamp to 2 lines with ellipsis |

## Color Opacity

```
bg-blue-500        → solid blue
bg-blue-500/50     → 50% opacity blue
bg-blue-500/0      → transparent
text-gray-900/80   → 80% opacity text
border-red-500/30  → 30% opacity border
```

## Common Component Recipes

### Centered Page Container
```html
<div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
```

### Card
```html
<div class="rounded-xl border bg-white p-6 shadow-sm dark:border-gray-800 dark:bg-gray-900">
```

### Badge
```html
<span class="inline-flex items-center rounded-full bg-blue-50 px-2 py-1 text-xs font-medium text-blue-700">
```

### Input
```html
<input class="block w-full rounded-lg border border-gray-300 px-3 py-2 text-sm shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500" />
```

### Button Primary
```html
<button class="inline-flex items-center justify-center rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 disabled:opacity-50">
```

### Responsive Stack → Row
```html
<div class="flex flex-col gap-4 sm:flex-row sm:items-center">
```

### Sticky Header
```html
<header class="sticky top-0 z-50 border-b bg-white/80 backdrop-blur-sm">
```

### Skeleton Loader
```html
<div class="h-4 w-3/4 animate-pulse rounded bg-gray-200"></div>
```

## Dark Mode

```html
<!-- Element-level dark mode -->
<div class="bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100">

<!-- Toggle script -->
<script>
  document.documentElement.classList.toggle('dark',
    localStorage.theme === 'dark' ||
    (!localStorage.theme && matchMedia('(prefers-color-scheme: dark)').matches)
  );
</script>
```

## State Variants

| Variant | Trigger |
|---------|---------|
| `hover:` | Mouse over |
| `focus:` | Element focused |
| `focus-visible:` | Keyboard focus only |
| `focus-within:` | Child element focused |
| `active:` | Being clicked |
| `disabled:` | Disabled attribute |
| `group-hover:` | Parent with `group` class hovered |
| `peer-checked:` | Sibling with `peer` class checked |
| `first:` | First child |
| `last:` | Last child |
| `odd:` | Odd children |
| `even:` | Even children |

## Tailwind v4 vs v3

| Feature | v3 | v4 |
|---------|----|----|
| Config | `tailwind.config.js` | `@theme { }` in CSS |
| Import | `@tailwind base/components/utilities` | `@import "tailwindcss"` |
| Colors | `theme.extend.colors` | `--color-*` CSS vars |
| Fonts | `theme.extend.fontFamily` | `--font-*` CSS vars |
| Plugins | `require('@tailwindcss/forms')` | `@plugin "@tailwindcss/forms"` |
| Dark mode | `darkMode: 'class'` | Automatic with `@variant dark` |
| Breakpoints | `theme.screens` | `--breakpoint-*` CSS vars |

## Sizing Quick Reference

| Class | Value |
|-------|-------|
| `w-full` | 100% |
| `w-screen` | 100vw |
| `w-fit` | fit-content |
| `w-min` | min-content |
| `w-max` | max-content |
| `max-w-sm` | 24rem (384px) |
| `max-w-md` | 28rem (448px) |
| `max-w-lg` | 32rem (512px) |
| `max-w-xl` | 36rem (576px) |
| `max-w-2xl` | 42rem (672px) |
| `max-w-4xl` | 56rem (896px) |
| `max-w-7xl` | 80rem (1280px) |
| `max-w-screen-xl` | 1280px |
| `min-h-screen` | 100vh |
| `h-dvh` | 100dvh (dynamic viewport) |

## Class Organization Convention

Order classes for readability:

```
1. Layout:      flex, grid, block, inline, hidden
2. Position:    relative, absolute, fixed, sticky, top-0, z-10
3. Box model:   w-full, h-12, p-4, m-2, border, rounded-lg
4. Typography:  text-sm, font-medium, text-gray-900, leading-tight
5. Visual:      bg-white, shadow-sm, opacity-50
6. Interactive: hover:bg-gray-50, focus:ring-2, transition, cursor-pointer
7. Responsive:  sm:flex-row, md:grid-cols-2, lg:px-8
```

## Arbitrary Values

```html
<!-- When Tailwind doesn't have the exact value -->
<div class="w-[327px] h-[calc(100vh-64px)] bg-[#1da1f2] text-[13px] grid-cols-[200px_1fr_100px]">

<!-- Arbitrary variants -->
<div class="[&>*]:mb-4 [&_p]:text-sm [&:nth-child(3)]:col-span-2">
```
