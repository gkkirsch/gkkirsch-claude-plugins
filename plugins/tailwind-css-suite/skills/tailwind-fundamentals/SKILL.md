---
name: tailwind-fundamentals
description: >
  Core Tailwind CSS utility patterns — layout, spacing, typography, colors, responsive design,
  dark mode, and the utility-first workflow.
  Triggers: "tailwind", "tailwind css", "utility classes", "tailwind layout", "tailwind responsive",
  "tailwind dark mode", "tailwind spacing", "tailwind flex", "tailwind grid".
  NOT for: component building patterns (use component-patterns), custom themes (use custom-themes).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Tailwind CSS Fundamentals

## Layout

```html
<!-- Flexbox -->
<div class="flex items-center justify-between gap-4">
  <div class="flex-1">Grows to fill</div>
  <div class="shrink-0">Fixed size</div>
</div>

<!-- Flex wrapping -->
<div class="flex flex-wrap gap-4">
  <div class="w-full sm:w-1/2 lg:w-1/3">Responsive columns</div>
  <div class="w-full sm:w-1/2 lg:w-1/3">Wraps on small screens</div>
  <div class="w-full sm:w-1/2 lg:w-1/3">Third column</div>
</div>

<!-- CSS Grid -->
<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
  <div>Card 1</div>
  <div>Card 2</div>
  <div>Card 3</div>
</div>

<!-- Grid with spanning -->
<div class="grid grid-cols-4 gap-4">
  <div class="col-span-2">Half width</div>
  <div class="col-span-1">Quarter</div>
  <div class="col-span-1">Quarter</div>
  <div class="col-span-4">Full width</div>
</div>

<!-- Auto-fit grid (responsive without breakpoints) -->
<div class="grid grid-cols-[repeat(auto-fit,minmax(250px,1fr))] gap-6">
  <div>Auto-sizes!</div>
</div>

<!-- Container -->
<div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
  Centered, responsive padding
</div>

<!-- Stack (vertical spacing) -->
<div class="space-y-4">
  <div>Item 1</div>
  <div>Item 2</div>
  <div>Item 3</div>
</div>

<!-- Centering -->
<div class="flex min-h-screen items-center justify-center">Centered</div>
<div class="grid min-h-screen place-items-center">Also centered</div>

<!-- Sticky header -->
<header class="sticky top-0 z-50 bg-white/80 backdrop-blur-sm border-b">
  Navigation
</header>

<!-- Sidebar layout -->
<div class="flex min-h-screen">
  <aside class="hidden w-64 shrink-0 border-r lg:block">Sidebar</aside>
  <main class="flex-1 overflow-auto">Content</main>
</div>
```

## Spacing Scale

| Class | Size | Pixels |
|-------|------|--------|
| `p-0` | 0 | 0px |
| `p-0.5` | 0.125rem | 2px |
| `p-1` | 0.25rem | 4px |
| `p-1.5` | 0.375rem | 6px |
| `p-2` | 0.5rem | 8px |
| `p-3` | 0.75rem | 12px |
| `p-4` | 1rem | 16px |
| `p-5` | 1.25rem | 20px |
| `p-6` | 1.5rem | 24px |
| `p-8` | 2rem | 32px |
| `p-10` | 2.5rem | 40px |
| `p-12` | 3rem | 48px |
| `p-16` | 4rem | 64px |
| `p-20` | 5rem | 80px |
| `p-24` | 6rem | 96px |

Same scale for margin (`m-`), gap, width (`w-`), height (`h-`), etc.

## Typography

```html
<!-- Font sizes -->
<p class="text-xs">12px</p>     <!-- 0.75rem -->
<p class="text-sm">14px</p>     <!-- 0.875rem -->
<p class="text-base">16px</p>   <!-- 1rem -->
<p class="text-lg">18px</p>     <!-- 1.125rem -->
<p class="text-xl">20px</p>     <!-- 1.25rem -->
<p class="text-2xl">24px</p>    <!-- 1.5rem -->
<p class="text-3xl">30px</p>    <!-- 1.875rem -->
<p class="text-4xl">36px</p>    <!-- 2.25rem -->

<!-- Font weight -->
<p class="font-light">300</p>
<p class="font-normal">400</p>
<p class="font-medium">500</p>
<p class="font-semibold">600</p>
<p class="font-bold">700</p>

<!-- Text styling -->
<p class="text-gray-600 leading-relaxed tracking-wide">
  Styled paragraph
</p>

<!-- Truncation -->
<p class="truncate">Single line truncation with ellipsis...</p>
<p class="line-clamp-3">Multi-line truncation after 3 lines...</p>

<!-- Responsive text -->
<h1 class="text-2xl font-bold sm:text-3xl lg:text-4xl xl:text-5xl">
  Responsive heading
</h1>
```

## Colors

```html
<!-- Gray scale (most used) -->
<div class="bg-gray-50">Lightest</div>
<div class="bg-gray-100">Light</div>
<div class="bg-gray-200">...</div>
<div class="bg-gray-500">Medium</div>
<div class="bg-gray-700">Dark text</div>
<div class="bg-gray-900">Darkest</div>
<div class="bg-gray-950">Near black</div>

<!-- Opacity modifier -->
<div class="bg-blue-500/50">50% opacity blue</div>
<div class="bg-black/80">80% opacity black</div>
<div class="text-white/90">90% opacity white text</div>

<!-- Semantic color pattern -->
<div class="bg-green-50 text-green-700 border border-green-200">Success</div>
<div class="bg-red-50 text-red-700 border border-red-200">Error</div>
<div class="bg-yellow-50 text-yellow-700 border border-yellow-200">Warning</div>
<div class="bg-blue-50 text-blue-700 border border-blue-200">Info</div>

<!-- Gradient -->
<div class="bg-gradient-to-r from-blue-500 to-purple-600 text-white">
  Gradient background
</div>
<div class="bg-gradient-to-b from-transparent via-black/50 to-black">
  Overlay gradient
</div>
```

## Responsive Design

```html
<!-- Mobile-first breakpoints -->
<!-- No prefix = mobile (all sizes) -->
<!-- sm: = 640px+ -->
<!-- md: = 768px+ -->
<!-- lg: = 1024px+ -->
<!-- xl: = 1280px+ -->
<!-- 2xl: = 1536px+ -->

<!-- Example: stack on mobile, row on tablet+ -->
<div class="flex flex-col gap-4 md:flex-row md:items-center">
  <div class="w-full md:w-1/2">Left</div>
  <div class="w-full md:w-1/2">Right</div>
</div>

<!-- Show/hide by breakpoint -->
<div class="block md:hidden">Mobile only</div>
<div class="hidden md:block">Tablet and up</div>
<div class="hidden lg:block">Desktop only</div>

<!-- Responsive padding -->
<section class="px-4 py-8 sm:px-6 sm:py-12 lg:px-8 lg:py-16">
  Progressive spacing
</section>

<!-- Responsive grid -->
<div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
  <!-- 1 col mobile → 2 tablet → 3 desktop → 4 wide -->
</div>
```

## Dark Mode

```html
<!-- Class strategy (recommended for apps) -->
<!-- Add "dark" class to <html> element -->
<div class="bg-white text-gray-900 dark:bg-gray-900 dark:text-gray-100">
  <h1 class="text-gray-900 dark:text-white">Title</h1>
  <p class="text-gray-600 dark:text-gray-400">Body text</p>
  <div class="border-gray-200 dark:border-gray-700">Border</div>
  <button class="bg-blue-600 hover:bg-blue-700 dark:bg-blue-500 dark:hover:bg-blue-400">
    Action
  </button>
</div>

<!-- Toggle script -->
<script>
  // Check preference or stored value
  if (localStorage.theme === 'dark' ||
      (!('theme' in localStorage) && window.matchMedia('(prefers-color-scheme: dark)').matches)) {
    document.documentElement.classList.add('dark');
  }

  // Toggle function
  function toggleDark() {
    document.documentElement.classList.toggle('dark');
    localStorage.theme = document.documentElement.classList.contains('dark') ? 'dark' : 'light';
  }
</script>
```

## Interactive States

```html
<!-- Hover, focus, active -->
<button class="
  bg-blue-600 text-white
  hover:bg-blue-700
  focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2
  active:bg-blue-800
  disabled:opacity-50 disabled:cursor-not-allowed
  transition-colors duration-150
">
  Button
</button>

<!-- Group hover (parent hover affects children) -->
<div class="group cursor-pointer rounded-lg border p-4 hover:border-blue-500">
  <h3 class="font-semibold group-hover:text-blue-600">Title</h3>
  <p class="text-gray-500 group-hover:text-gray-700">Description</p>
</div>

<!-- Focus-within -->
<div class="rounded-lg border focus-within:border-blue-500 focus-within:ring-2 focus-within:ring-blue-200">
  <input class="w-full border-0 bg-transparent p-3 focus:outline-none" />
</div>

<!-- Peer (sibling state) -->
<input type="checkbox" class="peer hidden" id="toggle" />
<label for="toggle" class="cursor-pointer">Toggle</label>
<div class="hidden peer-checked:block">Visible when checked</div>
```

## Animations & Transitions

```html
<!-- Transitions -->
<div class="transition-all duration-300 ease-in-out hover:scale-105 hover:shadow-lg">
  Smooth hover
</div>

<!-- Built-in animations -->
<div class="animate-spin">Spinner</div>
<div class="animate-ping">Ping effect</div>
<div class="animate-pulse">Skeleton loading</div>
<div class="animate-bounce">Bouncing</div>

<!-- Transform -->
<div class="hover:-translate-y-1 hover:shadow-xl transition-all duration-200">
  Card lift on hover
</div>
```

## Gotchas

1. **Mobile-first means `sm:` is NOT "small screens".** `sm:` means 640px AND UP. Unprefixed classes are for all sizes (starting from mobile). `text-sm md:text-base` means small text everywhere, base text from 768px up.

2. **`space-y-*` adds margin between children, not padding.** It uses `> * + *` selector. Doesn't work with absolutely positioned or hidden children. Use `gap-*` with flex/grid instead for more predictable behavior.

3. **`w-full` vs `max-w-full` matter differently.** `w-full` sets width to 100% of parent. `max-w-full` caps at parent width but allows smaller. For images, usually want `max-w-full h-auto`.

4. **Arbitrary values use square brackets.** `w-[137px]`, `bg-[#1a1a2e]`, `grid-cols-[200px_1fr_100px]`. These bypass the design system — use sparingly.

5. **`overflow-hidden` clips content AND creates a new stacking context.** Fixed-position children inside `overflow-hidden` parents can be clipped. Use `overflow-clip` instead if you only need clipping without the stacking context change (Tailwind v3.3+).

6. **Dark mode classes don't cascade from `<html>` by default in v4.** In Tailwind v4, dark mode uses `prefers-color-scheme` by default. For class-based toggling, configure `@custom-variant dark (&:where(.dark, .dark *))` in your CSS.
