---
name: component-patterns
description: >
  Tailwind CSS component building patterns — buttons, cards, forms, modals, navigation,
  tables, badges, and common UI components using utility-first approach.
  Triggers: "tailwind button", "tailwind card", "tailwind form", "tailwind modal",
  "tailwind navbar", "tailwind table", "tailwind component", "tailwind ui".
  NOT for: core utilities (use tailwind-fundamentals), theming (use custom-themes).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Tailwind Component Patterns

## Button Variants

```html
<!-- Primary -->
<button class="inline-flex items-center justify-center rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed transition-colors">
  Primary Button
</button>

<!-- Secondary -->
<button class="inline-flex items-center justify-center rounded-lg border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 shadow-sm hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 transition-colors">
  Secondary
</button>

<!-- Ghost -->
<button class="inline-flex items-center justify-center rounded-lg px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500 transition-colors">
  Ghost
</button>

<!-- Destructive -->
<button class="inline-flex items-center justify-center rounded-lg bg-red-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-red-500 focus:ring-offset-2 transition-colors">
  Delete
</button>

<!-- Icon button -->
<button class="inline-flex h-10 w-10 items-center justify-center rounded-lg text-gray-500 hover:bg-gray-100 hover:text-gray-700 focus:outline-none focus:ring-2 focus:ring-blue-500 transition-colors">
  <svg class="h-5 w-5" />
</button>

<!-- Button with icon -->
<button class="inline-flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 transition-colors">
  <svg class="h-4 w-4" />
  With Icon
</button>

<!-- Loading button -->
<button class="inline-flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white opacity-75 cursor-not-allowed" disabled>
  <svg class="h-4 w-4 animate-spin" viewBox="0 0 24 24">
    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" fill="none" />
    <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
  </svg>
  Loading...
</button>

<!-- Size variants -->
<button class="h-8 px-3 text-xs rounded-md ...">Small</button>
<button class="h-10 px-4 text-sm rounded-lg ...">Medium</button>
<button class="h-12 px-6 text-base rounded-lg ...">Large</button>
```

## Cards

```html
<!-- Basic card -->
<div class="rounded-xl border border-gray-200 bg-white p-6 shadow-sm">
  <h3 class="text-lg font-semibold text-gray-900">Card Title</h3>
  <p class="mt-2 text-sm text-gray-600">Card content goes here.</p>
</div>

<!-- Card with image -->
<div class="overflow-hidden rounded-xl border border-gray-200 bg-white shadow-sm">
  <img src="..." alt="..." class="h-48 w-full object-cover" />
  <div class="p-6">
    <h3 class="text-lg font-semibold text-gray-900">Title</h3>
    <p class="mt-2 text-sm text-gray-600">Description</p>
  </div>
</div>

<!-- Interactive card -->
<a href="#" class="group block rounded-xl border border-gray-200 bg-white p-6 shadow-sm transition-all hover:border-blue-300 hover:shadow-md">
  <h3 class="font-semibold text-gray-900 group-hover:text-blue-600">Clickable Card</h3>
  <p class="mt-2 text-sm text-gray-600">Hover to see the effect.</p>
  <span class="mt-4 inline-flex items-center text-sm font-medium text-blue-600">
    Learn more
    <svg class="ml-1 h-4 w-4 transition-transform group-hover:translate-x-1" />
  </span>
</a>

<!-- Stats card -->
<div class="rounded-xl border border-gray-200 bg-white p-6">
  <p class="text-sm font-medium text-gray-500">Total Revenue</p>
  <p class="mt-1 text-3xl font-bold text-gray-900">$45,231</p>
  <p class="mt-1 text-sm text-green-600">+12.5% from last month</p>
</div>
```

## Forms

```html
<!-- Text input -->
<div>
  <label for="email" class="block text-sm font-medium text-gray-700">Email</label>
  <input
    id="email"
    type="email"
    class="mt-1 block w-full rounded-lg border border-gray-300 px-3 py-2 text-sm shadow-sm placeholder:text-gray-400 focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500 disabled:bg-gray-50 disabled:text-gray-500"
    placeholder="you@example.com"
  />
</div>

<!-- Input with error -->
<div>
  <label class="block text-sm font-medium text-gray-700">Email</label>
  <input class="mt-1 block w-full rounded-lg border border-red-300 px-3 py-2 text-sm shadow-sm focus:border-red-500 focus:ring-1 focus:ring-red-500" />
  <p class="mt-1 text-sm text-red-600">Please enter a valid email address.</p>
</div>

<!-- Select -->
<select class="block w-full rounded-lg border border-gray-300 px-3 py-2 text-sm shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500">
  <option>Option 1</option>
  <option>Option 2</option>
</select>

<!-- Textarea -->
<textarea rows="4" class="block w-full rounded-lg border border-gray-300 px-3 py-2 text-sm shadow-sm placeholder:text-gray-400 focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500" placeholder="Write something..."></textarea>

<!-- Checkbox -->
<label class="flex items-center gap-2">
  <input type="checkbox" class="h-4 w-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500" />
  <span class="text-sm text-gray-700">Remember me</span>
</label>

<!-- Toggle switch -->
<button type="button" class="relative inline-flex h-6 w-11 shrink-0 cursor-pointer rounded-full border-2 border-transparent bg-gray-200 transition-colors focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2" role="switch" aria-checked="false">
  <span class="translate-x-0 pointer-events-none inline-block h-5 w-5 rounded-full bg-white shadow ring-0 transition-transform"></span>
</button>

<!-- Form layout -->
<form class="space-y-6">
  <div class="grid grid-cols-1 gap-6 sm:grid-cols-2">
    <!-- First name / Last name -->
  </div>
  <div><!-- Email --></div>
  <div><!-- Password --></div>
  <button type="submit" class="w-full rounded-lg bg-blue-600 px-4 py-2.5 text-sm font-medium text-white hover:bg-blue-700 transition-colors">
    Sign Up
  </button>
</form>
```

## Navigation

```html
<!-- Top navbar -->
<nav class="border-b bg-white">
  <div class="mx-auto flex h-16 max-w-7xl items-center justify-between px-4 sm:px-6 lg:px-8">
    <a href="/" class="text-xl font-bold text-gray-900">Logo</a>

    <!-- Desktop nav -->
    <div class="hidden items-center gap-8 md:flex">
      <a href="#" class="text-sm font-medium text-gray-900 hover:text-blue-600">Home</a>
      <a href="#" class="text-sm font-medium text-gray-500 hover:text-blue-600">Features</a>
      <a href="#" class="text-sm font-medium text-gray-500 hover:text-blue-600">Pricing</a>
      <button class="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700">
        Sign Up
      </button>
    </div>

    <!-- Mobile hamburger -->
    <button class="md:hidden rounded-lg p-2 text-gray-500 hover:bg-gray-100">
      <svg class="h-6 w-6" />
    </button>
  </div>
</nav>

<!-- Sidebar nav -->
<nav class="flex w-64 flex-col gap-1 border-r bg-gray-50 p-4">
  <a href="#" class="flex items-center gap-3 rounded-lg bg-blue-50 px-3 py-2 text-sm font-medium text-blue-700">
    <svg class="h-5 w-5" /> Dashboard
  </a>
  <a href="#" class="flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium text-gray-700 hover:bg-gray-100">
    <svg class="h-5 w-5" /> Settings
  </a>
</nav>

<!-- Breadcrumbs -->
<nav class="flex items-center gap-2 text-sm text-gray-500">
  <a href="#" class="hover:text-gray-700">Home</a>
  <span>/</span>
  <a href="#" class="hover:text-gray-700">Products</a>
  <span>/</span>
  <span class="text-gray-900 font-medium">Current Page</span>
</nav>

<!-- Tabs -->
<div class="flex border-b">
  <button class="border-b-2 border-blue-600 px-4 py-3 text-sm font-medium text-blue-600">Active</button>
  <button class="border-b-2 border-transparent px-4 py-3 text-sm font-medium text-gray-500 hover:border-gray-300 hover:text-gray-700">Tab 2</button>
  <button class="border-b-2 border-transparent px-4 py-3 text-sm font-medium text-gray-500 hover:border-gray-300 hover:text-gray-700">Tab 3</button>
</div>
```

## Modal / Dialog

```html
<!-- Modal backdrop + container -->
<div class="fixed inset-0 z-50 flex items-center justify-center">
  <!-- Backdrop -->
  <div class="fixed inset-0 bg-black/50 backdrop-blur-sm"></div>

  <!-- Modal content -->
  <div class="relative z-10 w-full max-w-md rounded-xl bg-white p-6 shadow-xl">
    <div class="flex items-center justify-between">
      <h2 class="text-lg font-semibold text-gray-900">Modal Title</h2>
      <button class="rounded-lg p-1 text-gray-400 hover:bg-gray-100 hover:text-gray-600">
        <svg class="h-5 w-5" />
      </button>
    </div>
    <div class="mt-4">
      <p class="text-sm text-gray-600">Modal body content.</p>
    </div>
    <div class="mt-6 flex justify-end gap-3">
      <button class="rounded-lg border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50">Cancel</button>
      <button class="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700">Confirm</button>
    </div>
  </div>
</div>
```

## Badges & Alerts

```html
<!-- Badges -->
<span class="inline-flex items-center rounded-full bg-green-100 px-2.5 py-0.5 text-xs font-medium text-green-800">Active</span>
<span class="inline-flex items-center rounded-full bg-yellow-100 px-2.5 py-0.5 text-xs font-medium text-yellow-800">Pending</span>
<span class="inline-flex items-center rounded-full bg-red-100 px-2.5 py-0.5 text-xs font-medium text-red-800">Inactive</span>

<!-- Alert -->
<div class="flex items-start gap-3 rounded-lg border border-blue-200 bg-blue-50 p-4">
  <svg class="mt-0.5 h-5 w-5 shrink-0 text-blue-600" />
  <div>
    <h3 class="text-sm font-medium text-blue-800">Information</h3>
    <p class="mt-1 text-sm text-blue-700">This is an informational message.</p>
  </div>
</div>

<!-- Toast notification -->
<div class="flex items-center gap-3 rounded-lg bg-gray-900 px-4 py-3 text-sm text-white shadow-lg">
  <svg class="h-5 w-5 text-green-400" />
  <p>Changes saved successfully.</p>
  <button class="ml-auto text-gray-400 hover:text-white">&times;</button>
</div>
```

## Table

```html
<div class="overflow-x-auto rounded-lg border border-gray-200">
  <table class="min-w-full divide-y divide-gray-200">
    <thead class="bg-gray-50">
      <tr>
        <th class="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">Name</th>
        <th class="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">Email</th>
        <th class="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">Role</th>
        <th class="px-6 py-3 text-right text-xs font-medium uppercase tracking-wider text-gray-500">Actions</th>
      </tr>
    </thead>
    <tbody class="divide-y divide-gray-200 bg-white">
      <tr class="hover:bg-gray-50 transition-colors">
        <td class="whitespace-nowrap px-6 py-4 text-sm font-medium text-gray-900">John Doe</td>
        <td class="whitespace-nowrap px-6 py-4 text-sm text-gray-500">john@example.com</td>
        <td class="whitespace-nowrap px-6 py-4">
          <span class="inline-flex rounded-full bg-green-100 px-2 py-0.5 text-xs font-medium text-green-800">Admin</span>
        </td>
        <td class="whitespace-nowrap px-6 py-4 text-right text-sm">
          <button class="text-blue-600 hover:text-blue-800">Edit</button>
        </td>
      </tr>
    </tbody>
  </table>
</div>
```

## Skeleton Loading

```html
<div class="animate-pulse space-y-4">
  <div class="h-48 rounded-lg bg-gray-200"></div>
  <div class="h-4 w-3/4 rounded bg-gray-200"></div>
  <div class="h-4 w-1/2 rounded bg-gray-200"></div>
  <div class="flex gap-4">
    <div class="h-10 w-10 rounded-full bg-gray-200"></div>
    <div class="flex-1 space-y-2">
      <div class="h-4 rounded bg-gray-200"></div>
      <div class="h-4 w-5/6 rounded bg-gray-200"></div>
    </div>
  </div>
</div>
```

## Gotchas

1. **Class order doesn't matter for specificity — but readability does.** Tailwind utilities all have the same specificity. But organize them logically: layout → sizing → spacing → typography → colors → effects → states → responsive.

2. **Don't mix Tailwind with custom CSS for the same property.** `class="text-red-500"` plus `.my-class { color: blue }` creates confusing specificity battles. Pick one approach per element.

3. **`divide-*` requires direct children only.** `divide-y` uses `> * + *` selector. Nested elements or fragments break the spacing. Use `space-y-*` or explicit borders if structure is complex.

4. **`ring-*` is for focus indicators, not borders.** `ring-2 ring-blue-500` creates a box-shadow-based ring outside the element. It doesn't affect layout. Use `border-*` for structural borders.

5. **Hover states on touch devices.** `hover:` states stick on mobile after touch. Consider using `@media (hover: hover)` or Tailwind's `hover:` which is already hover-media-query aware in v3.1+.

6. **`hidden` vs `invisible`.** `hidden` removes from layout (`display: none`). `invisible` keeps space but makes transparent (`visibility: hidden`). Use `hidden` for responsive show/hide, `invisible` for layout-preserving hide.
