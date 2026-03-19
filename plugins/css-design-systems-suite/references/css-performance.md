# CSS Performance Reference

Comprehensive guide to CSS performance optimization: rendering pipeline, compositor-friendly animations, containment, content-visibility, and measurement techniques.

---

## The Rendering Pipeline

Understanding the browser rendering pipeline is essential for writing performant CSS:

```
Style → Layout → Paint → Composite
```

1. **Style**: Calculate computed styles for every element
2. **Layout**: Calculate size and position of every element
3. **Paint**: Fill pixels for each element (backgrounds, borders, text, shadows)
4. **Composite**: Combine painted layers and draw to screen

### Cost of CSS Properties

| Triggers | Properties | Cost |
|----------|-----------|------|
| Style only | `color`, `visibility`, `opacity` (sometimes) | Low |
| Layout + Paint + Composite | `width`, `height`, `margin`, `padding`, `display`, `position`, `top/left/right/bottom`, `font-size`, `line-height`, `flex`, `grid` | High |
| Paint + Composite | `background`, `border`, `box-shadow`, `border-radius`, `outline`, `text-decoration` | Medium |
| Composite only | `transform`, `opacity`, `filter`, `backdrop-filter`, `clip-path` (some) | Very Low |

### Golden Rule

**Animate only compositor-friendly properties**: `transform`, `opacity`, `filter`.

```css
/* GOOD — compositor only, runs on GPU */
.card {
  transition: transform 0.3s, opacity 0.3s;
}
.card:hover {
  transform: translateY(-4px) scale(1.02);
}

/* BAD — triggers layout on every frame */
.card {
  transition: margin-top 0.3s, width 0.3s;
}
.card:hover {
  margin-top: -4px;
  width: 102%;
}

/* ACCEPTABLE — triggers paint, fine for short hover transitions */
.btn {
  transition: background-color 0.15s, border-color 0.15s;
}
.btn:hover {
  background-color: var(--color-brand-hover);
}
```

---

## CSS Containment

`contain` tells the browser that an element's internals are independent from the rest of the page, enabling optimizations.

### Containment Types

```css
/* Layout containment — element's layout doesn't affect siblings */
.widget {
  contain: layout;
}

/* Paint containment — element's paint doesn't overflow bounds */
.card {
  contain: paint;
}

/* Size containment — element's size doesn't depend on children */
.fixed-size-widget {
  contain: size;
  width: 300px;
  height: 200px;
}

/* Style containment — counters and quotes don't leak */
.component {
  contain: style;
}

/* Shorthand combinations */
.optimized-card {
  contain: layout paint;        /* Most common for cards/widgets */
}

.optimized-section {
  contain: layout style paint;  /* Full containment except size */
}

/* strict = size + layout + style + paint */
.fully-contained {
  contain: strict;
}

/* content = layout + style + paint (most useful shorthand) */
.content-contained {
  contain: content;
}
```

### When to Use Containment

```css
/* Cards in a grid — layout changes in one card don't affect others */
.card-grid > .card {
  contain: layout paint;
}

/* Off-screen or below-fold content */
.below-fold {
  contain: layout style paint;
}

/* Fixed-size widgets (ad slots, video embeds) */
.ad-slot {
  contain: strict;
  width: 300px;
  height: 250px;
}

/* Third-party content */
.third-party-embed {
  contain: layout style paint;
}
```

---

## content-visibility

`content-visibility` is the highest-impact CSS performance property. It skips rendering (layout, paint) for off-screen content.

### content-visibility: auto

```css
/* Skip rendering for off-screen sections */
.page-section {
  content-visibility: auto;
  contain-intrinsic-size: auto 500px; /* Estimated height for scrollbar accuracy */
}

/* Long list items */
.list-item {
  content-visibility: auto;
  contain-intrinsic-size: auto 80px;
}

/* Data table rows */
.table-row {
  content-visibility: auto;
  contain-intrinsic-size: auto 48px;
}

/* Blog post cards */
.article-card {
  content-visibility: auto;
  contain-intrinsic-size: auto 350px;
}
```

### contain-intrinsic-size

```css
/* Fixed estimate */
.section {
  contain-intrinsic-size: 300px 500px; /* width height */
}

/* Auto — remembers the last rendered size */
.section {
  contain-intrinsic-size: auto 500px; /* auto with 500px initial estimate */
}

/* Block-only (most common — width is usually known) */
.section {
  contain-intrinsic-block-size: auto 500px;
}
```

### Impact

content-visibility can reduce initial render time by 50-80% on long pages. It's the single most impactful CSS optimization for content-heavy pages.

**Caveats:**
- Don't use on above-the-fold content (it hides content until rendered)
- Elements with `content-visibility: auto` may cause layout shifts when scrolled into view
- `contain-intrinsic-size` should closely match actual size to minimize CLS
- Not suitable for elements that need to be found by in-page search (Chrome handles this, but other browsers may not)

---

## will-change

`will-change` hints to the browser that a property will change, allowing it to set up optimizations (like promoting to a compositor layer) ahead of time.

### Correct Usage

```css
/* Apply JUST BEFORE the animation, not permanently */

/* Approach 1: Apply on interaction trigger */
.card:hover {
  will-change: transform;
  transform: translateY(-4px);
}

/* Approach 2: Apply on parent hover, animate child */
.card-container:hover .card {
  will-change: transform;
}
.card-container:hover .card {
  transform: translateY(-4px);
}
```

```js
// Approach 3: Apply via JS before animation, remove after
element.addEventListener('mouseenter', () => {
  element.style.willChange = 'transform';
});

element.addEventListener('transitionend', () => {
  element.style.willChange = 'auto';
});
```

### Anti-Patterns

```css
/* DON'T: Apply to all elements */
* {
  will-change: transform, opacity;  /* Wastes massive GPU memory */
}

/* DON'T: Apply permanently to many elements */
.card {
  will-change: transform;  /* Creates a compositor layer for EVERY card */
}

/* DON'T: Use for non-animated properties */
.element {
  will-change: color, background-color;  /* These aren't compositor-friendly anyway */
}
```

### When to Use

- Elements that WILL animate (not might animate)
- Complex animations where you notice jank without it
- Elements with `position: fixed` that scroll with the page
- Remove with `will-change: auto` when animation completes

---

## Selector Performance

### Selector Matching Order

Browsers match selectors **right to left**. The rightmost part (key selector) is checked first:

```css
/* Browser first finds ALL .title elements, then checks if parent is .card */
.card .title { }

/* Browser first finds ALL p elements in the entire document */
.sidebar .widget .content p { }
```

### Performance Guidelines

```css
/* GOOD: Low specificity, specific key selector */
.card-title { }
.nav-link { }
[data-component="card"] { }

/* FINE: Shallow nesting */
.card > .title { }
.nav .link { }

/* AVOID: Deep nesting (but impact is usually negligible in practice) */
.page .main .content .article .card .body .text p { }

/* AVOID: Universal key selector */
.card * { }

/* AVOID in large DOM: Attribute selectors with complex matching */
[class*="icon-"] { }
[href$=".pdf"] { }
```

### Practical Reality

In modern browsers, selector performance is rarely a bottleneck unless:
- You have 10,000+ DOM elements
- You're using very complex selectors on very large DOMs
- You're triggering style recalculation frequently (JS-driven animations)

**Focus on layout/paint cost, not selector performance.**

---

## Reducing Layout Thrashing

Layout thrashing occurs when JavaScript reads layout properties then writes style changes in a tight loop:

```js
// BAD: Forces synchronous layout on every iteration
items.forEach((item) => {
  const height = item.offsetHeight;        // Read → forces layout
  item.style.height = `${height * 2}px`;   // Write → invalidates layout
  // Next read will force layout AGAIN
});

// GOOD: Batch reads, then batch writes
const heights = items.map((item) => item.offsetHeight); // All reads first
items.forEach((item, i) => {
  item.style.height = `${heights[i] * 2}px`;            // All writes second
});

// BEST: Use requestAnimationFrame or CSS
// Let CSS handle sizing with aspect-ratio, flex, grid, etc.
```

### Layout-Triggering Properties (Reads)

These properties force synchronous layout when read:

```
offsetTop, offsetLeft, offsetWidth, offsetHeight
scrollTop, scrollLeft, scrollWidth, scrollHeight
clientTop, clientLeft, clientWidth, clientHeight
getComputedStyle()
getBoundingClientRect()
```

---

## Font Performance

### Font Loading Strategy

```css
/* Use font-display to control loading behavior */
@font-face {
  font-family: 'Inter';
  src: url('/fonts/inter-var.woff2') format('woff2');
  font-weight: 100 900;
  font-display: swap;          /* Show fallback immediately, swap when loaded */
  unicode-range: U+0000-00FF;  /* Only load Latin characters */
}

/* Preload critical fonts */
/* In HTML <head>: */
/* <link rel="preload" href="/fonts/inter-var.woff2" as="font" type="font/woff2" crossorigin> */
```

### font-display Values

| Value | Behavior | Best For |
|-------|----------|----------|
| `swap` | Show fallback immediately, swap when loaded | Body text |
| `optional` | Show fallback, only swap if loaded quickly | Non-critical fonts |
| `fallback` | Short blank period, then fallback, then swap | Headings |
| `block` | Brief blank period, then show when loaded | Icon fonts |
| `auto` | Browser decides | Default |

### Reducing Layout Shift from Fonts

```css
/* Use size-adjust and ascent/descent-override to match fallback metrics */
@font-face {
  font-family: 'Inter';
  src: url('/fonts/inter-var.woff2') format('woff2');
  font-display: swap;
}

/* Adjusted fallback that matches Inter's metrics */
@font-face {
  font-family: 'Inter Fallback';
  src: local('Arial');
  ascent-override: 90%;
  descent-override: 22%;
  line-gap-override: 0%;
  size-adjust: 107%;
}

body {
  font-family: 'Inter', 'Inter Fallback', system-ui, sans-serif;
}
```

---

## Critical CSS

### Inline Critical CSS

Extract and inline styles needed for above-the-fold content:

```html
<head>
  <!-- Critical CSS inlined -->
  <style>
    /* Only styles for above-the-fold content */
    body { margin: 0; font-family: system-ui; }
    .header { display: flex; align-items: center; height: 64px; }
    .hero { min-height: 100vh; display: grid; place-items: center; }
  </style>

  <!-- Full CSS loaded async -->
  <link rel="preload" href="/styles.css" as="style" onload="this.onload=null;this.rel='stylesheet'">
  <noscript><link rel="stylesheet" href="/styles.css"></noscript>
</head>
```

### CSS Splitting Strategy

```css
/* critical.css — inlined in <head> */
/* Contains: reset, layout, header, hero, above-fold components */

/* main.css — loaded async */
/* Contains: below-fold components, utilities, print styles */

/* component.css — loaded on demand */
/* Contains: dialog, tooltip, drawer, other interactive components */
```

---

## Measuring CSS Performance

### Chrome DevTools

1. **Performance panel**: Record and analyze rendering performance
   - Look for long "Recalculate Style" tasks (>5ms)
   - Look for "Layout" events during animations
   - Check "Paint" events — should be minimal during scroll/animation

2. **Rendering panel** (More tools → Rendering):
   - **Paint flashing**: Green highlights show areas being repainted
   - **Layout shift regions**: Blue highlights show layout shifts
   - **Layer borders**: Orange borders show compositor layers
   - **FPS meter**: Real-time frame rate

3. **Coverage panel** (More tools → Coverage):
   - Shows how much CSS is actually used on the current page
   - Helps identify CSS that can be deferred or removed

### Core Web Vitals CSS Impact

| Metric | CSS Impact | Optimization |
|--------|-----------|-------------|
| **LCP** (Largest Contentful Paint) | Render-blocking CSS delays LCP | Inline critical CSS, async load rest |
| **CLS** (Cumulative Layout Shift) | Font loading, dynamic content, images without dimensions | font-display, size-adjust, aspect-ratio |
| **INP** (Interaction to Next Paint) | Long style recalculations block main thread | contain, content-visibility, reduce selector complexity |

### Performance Budget

| Metric | Target | Warning |
|--------|--------|---------|
| Total CSS size | <50KB (compressed) | >100KB |
| Critical CSS | <14KB (first TCP round-trip) | >20KB |
| Unused CSS | <20% of total | >40% |
| Style recalc per frame | <5ms | >10ms |
| Layout per frame | <5ms | >10ms |
| Compositor layers | <30 | >50 |

---

## Optimization Checklist

### High Impact
- [ ] Use `content-visibility: auto` on below-fold sections
- [ ] Inline critical CSS for above-fold content
- [ ] Preload critical fonts with `font-display: swap`
- [ ] Animate only `transform`, `opacity`, `filter`
- [ ] Use `contain: layout paint` on independent components
- [ ] Remove unused CSS (check with Coverage panel)

### Medium Impact
- [ ] Use `aspect-ratio` on images/videos to prevent layout shift
- [ ] Use `size-adjust` and metric overrides for font fallbacks
- [ ] Split CSS by route/feature for code splitting
- [ ] Use native CSS features over JavaScript alternatives
- [ ] Use `@layer` to avoid `!important` chains
- [ ] Minimize use of `box-shadow` in animations (use `filter: drop-shadow` or pseudo-elements)

### Low Impact (but good practice)
- [ ] Keep selectors shallow (3 levels max)
- [ ] Prefer class selectors over attribute selectors
- [ ] Avoid `*` as key selector
- [ ] Use `will-change` sparingly and remove after animation
- [ ] Avoid `@import` in CSS (use build tool bundling instead)
- [ ] Minimize CSS custom property inheritance chains

---

## Quick Wins

### Replace JS Animations with CSS

```css
/* Before: JS-based scroll handler */
/* window.addEventListener('scroll', updateHeader) */

/* After: CSS scroll-driven animation */
.header {
  animation: header-bg linear both;
  animation-timeline: scroll(root);
  animation-range: 0px 100px;
}
```

### Replace JS Layout with CSS

```css
/* Before: JS-based masonry layout */
/* After: CSS columns (imperfect but no JS) */
.masonry {
  columns: 250px;
  column-gap: 1rem;
}
.masonry > * {
  break-inside: avoid;
  margin-bottom: 1rem;
}
```

### Replace JS Intersection Observer with CSS

```css
/* Before: IntersectionObserver for reveal-on-scroll */
/* After: CSS scroll-driven animation */
.reveal {
  animation: reveal 1s both;
  animation-timeline: view();
  animation-range: entry 10% entry 50%;
}
```

---

## Anti-Patterns

1. **Don't use `transition: all`** — transition specific properties to avoid unintended transitions
2. **Don't apply `will-change` globally** — each use creates a compositor layer consuming GPU memory
3. **Don't animate `width`/`height`** — use `transform: scale()` instead
4. **Don't animate `top`/`left`** — use `transform: translate()` instead
5. **Don't use `@import` in CSS files** — it creates sequential loading; use build tools instead
6. **Don't load all CSS upfront** — split and lazy-load non-critical CSS
7. **Don't use complex selectors in hot paths** — simple class selectors during frequent updates
8. **Don't forget `contain-intrinsic-size`** with `content-visibility: auto` — prevents scrollbar jumping
9. **Don't use `box-shadow` in scroll animations** — it triggers paint; use `filter: drop-shadow()` or overlays
10. **Don't ignore the Coverage panel** — unused CSS is wasted bytes and parsing time
