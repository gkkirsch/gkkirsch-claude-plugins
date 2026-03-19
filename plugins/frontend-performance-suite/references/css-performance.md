# CSS Performance Reference

Comprehensive reference for CSS performance optimization techniques including containment, layers, animation performance, selector efficiency, and modern CSS features that improve rendering speed.

## CSS Rendering Pipeline

Understanding how the browser processes CSS is essential for optimizing performance.

### The Rendering Pipeline

```
Style → Layout → Paint → Composite

1. Style: Parse CSS, resolve selectors, compute styles for each element
2. Layout: Calculate element sizes and positions (also called "reflow")
3. Paint: Fill in pixels for each element (colors, images, text, shadows)
4. Composite: Combine painted layers in the correct order
```

### Properties by Pipeline Stage

Changing a CSS property triggers different stages. Fewer stages = better performance.

**Layout Properties** (triggers: Style → Layout → Paint → Composite):
```
width, height, min-width, min-height, max-width, max-height
padding, margin, border-width
top, right, bottom, left (on positioned elements)
display, position, float, clear
flex-basis, flex-grow, flex-shrink
grid-template-*, grid-column, grid-row
font-size, font-weight, font-family, line-height
text-align, vertical-align, white-space
overflow, writing-mode
```

**Paint Properties** (triggers: Style → Paint → Composite):
```
color, background-color, background-image
border-color, border-style, border-radius
box-shadow, text-shadow, text-decoration
outline, outline-color
visibility
clip-path (non-composited)
filter (non-composited)
```

**Composite Properties** (triggers: Style → Composite only):
```
transform (translate, rotate, scale, skew)
opacity
will-change
perspective
```

**Key takeaway: Use `transform` and `opacity` for animations.** They skip layout and paint, running only on the GPU compositor thread.

---

## CSS Containment

CSS Containment allows you to isolate parts of the page so the browser can optimize rendering.

### contain Property

```css
/* contain: layout — Element's layout is independent of the rest of the page */
/* Internal layout changes don't trigger layout of ancestors */
.card {
  contain: layout;
}

/* contain: paint — Element's content won't paint outside its bounds */
/* Descendants outside the element's box are not visible */
/* Creates a new stacking context, containing block, and formatting context */
.card {
  contain: paint;
}

/* contain: style — Counters and quotes don't escape this element */
.card {
  contain: style;
}

/* contain: size — Element's size is not affected by its children */
/* You MUST set explicit dimensions when using this */
.card {
  contain: size;
  width: 300px;
  height: 200px;
}

/* contain: content — Shorthand for layout + paint + style */
/* The most common and useful containment value */
.card {
  contain: content;
}

/* contain: strict — Shorthand for layout + paint + style + size */
/* Most aggressive containment — requires explicit dimensions */
.scrollable-list {
  contain: strict;
  width: 100%;
  height: 500px;
  overflow-y: auto;
}
```

### When to Use Containment

```css
/* Scrollable containers — huge performance win */
.scroll-container {
  contain: strict;
  overflow-y: auto;
  height: 600px;
}

/* Independent UI cards / widgets */
.dashboard-widget {
  contain: content;
}

/* List items in long lists */
.list-item {
  contain: content;
}

/* Off-screen content */
.below-fold {
  contain: content;
}

/* Third-party widget containers */
.ad-slot, .embed-container {
  contain: strict;
  width: 300px;
  height: 250px;
}
```

### content-visibility

`content-visibility` is a higher-level API built on containment. It lets the browser skip rendering work for off-screen content entirely.

```css
/* auto: Render when near the viewport, skip when far away */
.section {
  content-visibility: auto;
  contain-intrinsic-size: 0 500px; /* Estimated height to prevent layout shifts */
}

/* hidden: Never render (like display:none but preserves accessibility tree) */
.hidden-panel {
  content-visibility: hidden;
}

/* visible: Normal rendering (default) */
.always-visible {
  content-visibility: visible;
}
```

#### contain-intrinsic-size

When using `content-visibility: auto`, the browser needs a size estimate for off-screen elements to prevent layout shifts.

```css
/* Fixed estimate */
.section {
  content-visibility: auto;
  contain-intrinsic-size: 0 500px; /* 0 width (auto), 500px estimated height */
}

/* Auto with remembered last size (Chrome 107+) */
.section {
  content-visibility: auto;
  contain-intrinsic-size: auto 500px;
  /* Uses 500px initially, then remembers actual size once rendered */
}

/* Separate width and height */
.section {
  content-visibility: auto;
  contain-intrinsic-width: auto 300px;
  contain-intrinsic-height: auto 500px;
}
```

#### Performance Impact

```
Without content-visibility:
- 1000 items rendered, each takes 2ms layout/paint
- Total: ~2000ms rendering time

With content-visibility: auto:
- Only ~20 visible items rendered (2ms each = 40ms)
- Off-screen items get 0ms rendering
- Total: ~40ms rendering time (50x improvement!)
```

#### Caveats

- Elements with `content-visibility: auto` don't participate in `Ctrl+F` (Find in Page) search until they're rendered
- Fragment navigation (anchor links) may not work for off-screen content
- Should not be used on elements that need to be counted by `counter-increment`

---

## CSS Layers (@layer)

CSS Layers help manage specificity and avoid cascade conflicts, indirectly improving performance by reducing `!important` usage and complex selector chains.

### Basic Usage

```css
/* Define layer order (first = lowest priority) */
@layer reset, base, components, utilities;

/* Add styles to a layer */
@layer reset {
  *, *::before, *::after {
    box-sizing: border-box;
    margin: 0;
    padding: 0;
  }
}

@layer base {
  body {
    font-family: system-ui, sans-serif;
    line-height: 1.6;
    color: #333;
  }
  a {
    color: #0066cc;
  }
}

@layer components {
  .btn {
    padding: 0.5rem 1rem;
    border-radius: 0.375rem;
    font-weight: 500;
  }
  .btn-primary {
    background: #0066cc;
    color: white;
  }
}

@layer utilities {
  .sr-only {
    position: absolute;
    width: 1px;
    height: 1px;
    overflow: hidden;
    clip: rect(0, 0, 0, 0);
  }
  .hidden {
    display: none;
  }
}
```

### Layer Priority

```css
/* Layers declared first have LOWER priority */
@layer base, components, overrides;

/* Unlayered styles ALWAYS win over layered styles */
/* This is useful for one-off overrides */

/* Priority (lowest → highest):
   1. @layer base (first declared)
   2. @layer components
   3. @layer overrides (last declared)
   4. Unlayered styles (always highest)
*/
```

### Performance Benefits

```css
/* WITHOUT layers: Specificity wars lead to bloated CSS */
.sidebar .nav .nav-item .nav-link.active { /* specificity: 0,5,0 */
  color: blue;
}
.header .nav .nav-item .nav-link.active { /* specificity: 0,5,0 — same! */
  color: red !important; /* need !important to override */
}

/* WITH layers: Clean specificity, no !important needed */
@layer components {
  .nav-link.active { color: blue; }
}
@layer overrides {
  .nav-link.active { color: red; } /* Wins because layer is higher priority */
}
```

### Nested Layers

```css
@layer components {
  @layer button {
    .btn { padding: 0.5rem 1rem; }
  }
  @layer card {
    .card { border-radius: 0.5rem; }
  }
}

/* Reference nested layers with dot notation */
@layer components.button {
  .btn-lg { padding: 1rem 2rem; }
}
```

### Third-Party CSS in Layers

```css
/* Put third-party CSS in a low-priority layer */
@layer third-party, app;

@import url('https://cdn.example.com/library.css') layer(third-party);

@layer app {
  /* Your styles always override third-party */
  .component {
    /* No need for !important */
  }
}
```

---

## will-change

The `will-change` property hints to the browser that an element will change, allowing it to set up optimizations ahead of time (like promoting to a compositor layer).

### Correct Usage

```css
/* GOOD: Apply will-change just before animation starts */
.card:hover {
  will-change: transform;
}
.card:active {
  transform: scale(0.98);
}

/* GOOD: Apply via JavaScript before animation */
element.style.willChange = 'transform';
element.animate(
  [{ transform: 'translateX(0)' }, { transform: 'translateX(100px)' }],
  { duration: 300 }
).onfinish = () => {
  element.style.willChange = 'auto';
};

/* GOOD: For elements that are continuously animated */
.spinner {
  will-change: transform;
  animation: spin 1s linear infinite;
}

/* GOOD: For scroll-triggered animations, apply when element is about to enter viewport */
.animate-on-scroll.about-to-animate {
  will-change: transform, opacity;
}
```

### Misuse to Avoid

```css
/* BAD: will-change on everything */
* {
  will-change: transform;
}

/* BAD: Permanent will-change on many elements */
.card {
  will-change: transform; /* Creates GPU layer for EVERY card */
}

/* BAD: Too many properties */
.element {
  will-change: transform, opacity, color, background, border, box-shadow;
}
```

### Memory Impact

Each `will-change: transform` or `will-change: opacity` creates a separate compositor layer. Each layer consumes GPU memory:

```
Layer memory ≈ width × height × 4 bytes (RGBA)
A 1000×500 element = 2MB per layer
100 elements with will-change = 200MB GPU memory!

Guideline:
- Maximum ~20 will-change layers at any time
- Remove will-change when animation ends
- Never apply to elements larger than necessary
```

---

## Animation Performance

### GPU-Accelerated Properties

Only `transform` and `opacity` are guaranteed to run on the compositor thread (GPU-accelerated, off-main-thread):

```css
/* FAST: Compositor-only animation */
.slide-in {
  animation: slideIn 0.3s ease;
}
@keyframes slideIn {
  from {
    transform: translateX(-100%);
    opacity: 0;
  }
  to {
    transform: translateX(0);
    opacity: 1;
  }
}

/* SLOW: Layout-triggering animation */
.slide-in-bad {
  animation: slideInBad 0.3s ease;
}
@keyframes slideInBad {
  from {
    margin-left: -100%;
    height: 0;
  }
  to {
    margin-left: 0;
    height: auto;
  }
}
```

### Common Animation Patterns (Optimized)

```css
/* Fade in */
.fade-in {
  animation: fadeIn 0.3s ease;
}
@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

/* Scale on hover */
.hover-scale {
  transition: transform 0.2s ease;
}
.hover-scale:hover {
  transform: scale(1.05);
}

/* Slide from side */
.slide-from-right {
  animation: slideFromRight 0.3s ease;
}
@keyframes slideFromRight {
  from { transform: translateX(100%); }
  to { transform: translateX(0); }
}

/* Collapse/expand height — use transform instead of height */
.collapse {
  transform: scaleY(0);
  transform-origin: top;
  transition: transform 0.3s ease;
}
.collapse.open {
  transform: scaleY(1);
}

/* Card flip */
.card {
  perspective: 1000px;
}
.card-inner {
  transition: transform 0.6s;
  transform-style: preserve-3d;
}
.card:hover .card-inner {
  transform: rotateY(180deg);
}
```

### Height Animation Alternatives

Animating `height` is a common need but `height: auto` can't be transitioned and height triggers layout.

```css
/* Option 1: Use max-height (simple but has timing issues) */
.collapsible {
  max-height: 0;
  overflow: hidden;
  transition: max-height 0.3s ease;
}
.collapsible.open {
  max-height: 500px; /* Set higher than content */
}

/* Option 2: Use CSS Grid (modern, clean) */
.collapsible {
  display: grid;
  grid-template-rows: 0fr;
  transition: grid-template-rows 0.3s ease;
}
.collapsible.open {
  grid-template-rows: 1fr;
}
.collapsible > div {
  overflow: hidden;
}

/* Option 3: interpolate-size (Chrome 129+, most elegant) */
.collapsible {
  interpolate-size: allow-keywords;
  height: 0;
  overflow: hidden;
  transition: height 0.3s ease;
}
.collapsible.open {
  height: auto; /* Actually transitions to auto! */
}

/* Option 4: calc-size() (Chrome 129+) */
.collapsible.open {
  height: calc-size(auto);
}
```

### Reduced Motion

```css
/* ALWAYS provide reduced motion support */
@media (prefers-reduced-motion: reduce) {
  *,
  *::before,
  *::after {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
    scroll-behavior: auto !important;
  }
}

/* Better: Use reduce as default, enhance for motion-OK */
.element {
  /* No animation by default */
  opacity: 1;
  transform: none;
}

@media (prefers-reduced-motion: no-preference) {
  .element {
    animation: fadeSlideIn 0.3s ease;
  }
}
```

### Web Animations API

```javascript
// Better performance than CSS transitions for dynamic animations
// The browser can optimize WAAPI animations to run on the compositor

const animation = element.animate(
  [
    { transform: 'translateX(0)', opacity: 1 },
    { transform: 'translateX(100px)', opacity: 0 },
  ],
  {
    duration: 300,
    easing: 'ease-out',
    fill: 'forwards',
  }
);

// Clean up will-change after animation
animation.onfinish = () => {
  element.style.transform = 'translateX(100px)';
  element.style.opacity = '0';
};

// Cancel if needed
animation.cancel();
```

### View Transitions API

```css
/* Page transitions with View Transitions (Chrome 111+) */
::view-transition-old(root) {
  animation: fade-out 0.3s ease;
}
::view-transition-new(root) {
  animation: fade-in 0.3s ease;
}

/* Named transitions for specific elements */
.card {
  view-transition-name: card-hero;
}
::view-transition-old(card-hero) {
  animation: scale-down 0.3s ease;
}
::view-transition-new(card-hero) {
  animation: scale-up 0.3s ease;
}
```

```javascript
// Trigger a view transition
document.startViewTransition(() => {
  // Update the DOM
  updatePageContent();
});
```

---

## Selector Performance

### Selector Matching

Browsers match selectors **right to left**. The rightmost part (key selector) is matched first.

```css
/* The browser first finds ALL .active elements,
   then filters to those inside .sidebar .nav .nav-item */
.sidebar .nav .nav-item .active {
  color: blue;
}

/* This is faster because the key selector .nav-item--active is more specific */
.nav-item--active {
  color: blue;
}
```

### Selector Performance Guidelines

```css
/* FAST selectors (modern browsers handle these easily) */
.class-name { }          /* Class — very fast */
#id { }                  /* ID — very fast */
element { }              /* Tag — fast */
[attribute] { }          /* Attribute existence — fast */

/* MEDIUM selectors */
.parent .child { }       /* Descendant — browser walks up tree */
.parent > .child { }     /* Direct child — slightly faster than descendant */
.sibling + .next { }     /* Adjacent sibling — fast */

/* SLOW selectors (avoid in hot paths, long lists) */
* { }                    /* Universal — matches everything */
:nth-child(odd) { }      /* Structural pseudo-classes on large lists */
[attr^="val"] { }        /* Attribute substring matching */
.a .b .c .d .e .f { }   /* Deeply nested — many tree walks */

/* Practical guidelines:
   - Keep selectors under 3 levels deep
   - Prefer class selectors
   - Avoid universal selectors in compound selectors
   - Avoid :nth-child on lists with 1000+ items (use classes instead)
   - BEM or similar naming conventions help keep selectors flat */
```

### CSS vs selector performance reality

**Modern browsers are very fast at selector matching.** For most websites, selector performance is negligible compared to layout, paint, and JavaScript costs. Focus on:

1. Reducing total CSS size (unused CSS removal)
2. Reducing layout/paint cost (containment, composited animations)
3. Reducing style recalculation scope (containment)

---

## Critical CSS

### What Is Critical CSS

Critical CSS is the minimum CSS needed to render above-the-fold content. By inlining it in `<head>`, the browser can render the page without waiting for external stylesheet downloads.

### Target Size

```
Critical CSS budget: ≤ 14KB (compressed)

Why 14KB? The first TCP round-trip can deliver ~14KB.
Inlining critical CSS within this budget means the browser can
render the page after just one round-trip (after DNS + TCP + TLS).
```

### Implementation

```html
<head>
  <!-- Inline critical CSS -->
  <style>
    /* Only styles needed for above-the-fold content */
    body { margin: 0; font-family: system-ui, sans-serif; }
    .header { display: flex; height: 64px; background: #fff; }
    .hero { padding: 4rem 2rem; background: #f8f9fa; }
    .hero h1 { font-size: clamp(2rem, 5vw, 3rem); }
    /* ... */
  </style>

  <!-- Load full CSS asynchronously -->
  <link rel="preload" href="/styles.css" as="style" onload="this.onload=null;this.rel='stylesheet'">
  <noscript><link rel="stylesheet" href="/styles.css"></noscript>
</head>
```

### Automated Critical CSS Extraction

```javascript
// Using critters (webpack plugin)
const Critters = require('critters-webpack-plugin');
new Critters({
  preload: 'swap',
  pruneSource: true,
});

// Using critical (Node.js)
const critical = require('critical');
await critical.generate({
  inline: true,
  base: 'dist/',
  src: 'index.html',
  target: 'index.html',
  dimensions: [
    { width: 375, height: 812 },  // Mobile
    { width: 1280, height: 900 }, // Desktop
  ],
});
```

---

## CSS File Size Optimization

### Reducing CSS Size

```
1. Remove unused CSS
   - PurgeCSS for any framework
   - Tailwind JIT/v4 automatic purging
   - Coverage tool in Chrome DevTools

2. Minification
   - cssnano (PostCSS)
   - Lightning CSS (very fast)
   - esbuild (built into Vite)

3. Compression
   - gzip: ~60-70% compression
   - brotli: ~70-80% compression

4. Reduce redundancy
   - CSS custom properties instead of repeated values
   - Utility classes instead of repetitive component styles
   - CSS layers to avoid !important chains
```

### Measuring Unused CSS

```
Chrome DevTools → Coverage tab:
1. Open DevTools (F12)
2. Cmd+Shift+P → "Show Coverage"
3. Click record
4. Navigate the page
5. Review red (unused) vs blue (used) CSS

Typical findings:
- CSS frameworks: 80-95% unused (if not tree-shaken)
- Component libraries: 60-80% unused
- Custom CSS: 20-40% unused (from deleted features)
```

---

## Modern CSS Performance Features

### Scroll-Driven Animations

```css
/* Animate based on scroll position (Chrome 115+) */
/* No JavaScript needed! Runs on compositor thread. */

/* Progress bar that fills as you scroll */
.progress-bar {
  position: fixed;
  top: 0;
  left: 0;
  height: 3px;
  background: #0066cc;
  transform-origin: left;
  animation: grow-progress linear;
  animation-timeline: scroll(); /* Linked to page scroll */
}

@keyframes grow-progress {
  from { transform: scaleX(0); }
  to { transform: scaleX(1); }
}

/* Fade in elements as they scroll into view */
.fade-on-scroll {
  animation: fadeIn linear;
  animation-timeline: view(); /* Linked to element's visibility */
  animation-range: entry 0% entry 100%;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(20px); }
  to { opacity: 1; transform: translateY(0); }
}
```

### Anchor Positioning

```css
/* CSS Anchor Positioning (Chrome 125+) */
/* Position tooltips, popovers without JavaScript */

.trigger {
  anchor-name: --my-trigger;
}

.tooltip {
  position: fixed;
  position-anchor: --my-trigger;
  bottom: anchor(top);
  left: anchor(center);
  transform: translateX(-50%);
  /* Auto-flip if not enough space */
  position-try-fallbacks: flip-block;
}
```

### CSS Nesting

```css
/* Native CSS nesting (all modern browsers) */
.card {
  padding: 1rem;
  border-radius: 0.5rem;

  & .title {
    font-size: 1.25rem;
    font-weight: 600;
  }

  &:hover {
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  }

  @media (min-width: 768px) {
    padding: 2rem;
  }
}

/* Performance note: Nesting doesn't affect runtime performance.
   It's purely a developer experience feature.
   The browser parses nested rules the same as flat rules. */
```

### :has() Performance Considerations

```css
/* :has() is generally fast, but be aware of these patterns */

/* FAST: Simple child check */
.card:has(> img) { }
.form-group:has(:invalid) { }

/* MEDIUM: Descendant check (browser must walk subtree) */
.container:has(.error) { }

/* POTENTIALLY SLOW on very large DOMs: */
/* Avoid :has() that requires checking many descendants */
body:has(.deeply-nested-rare-class) { }

/* Best practice: Keep :has() selectors close to the element */
.card:has(> .badge) { }  /* GOOD: Direct child */
.card:has(.badge) { }     /* OK: But more expensive */
```

---

## Performance Debugging

### Chrome DevTools CSS Performance

```
Performance panel:
1. Record a trace
2. Look for "Recalculate Style" events
3. Long "Recalculate Style" events indicate expensive CSS

Rendering panel (Cmd+Shift+P → "Show Rendering"):
- Paint flashing: Shows areas being repainted (should be minimal)
- Layout shift regions: Shows CLS-causing layout shifts
- Layer borders: Shows compositor layers

Coverage panel:
- Shows which CSS rules are actually used
- Red = unused, Blue = used
```

### Common CSS Performance Issues

```
1. Expensive box-shadow on many elements
   Fix: Use drop-shadow filter (sometimes cheaper) or reduce shadow complexity

2. Large border-radius on oversized elements
   Fix: Simplify or use clip-path

3. Filter effects (blur, brightness) on large areas
   Fix: Apply to smaller elements, use backdrop-filter sparingly

4. Frequent style recalculations from JS
   Fix: Batch DOM reads/writes, use requestAnimationFrame

5. Complex CSS selectors in large lists
   Fix: Simplify selectors, use containment on list items

6. !important chains requiring even more !important
   Fix: Use CSS layers to manage cascade priority
```

---

## Quick Reference: Performance Checklist

```
□ Animations use only transform and opacity
□ will-change applied sparingly and removed after animation
□ contain: content on independent UI sections
□ content-visibility: auto on below-fold content
□ Critical CSS inlined (≤ 14KB compressed)
□ Non-critical CSS loaded asynchronously
□ Unused CSS removed (PurgeCSS or Tailwind JIT)
□ CSS compressed with brotli/gzip
□ No layout-triggering animations
□ @media (prefers-reduced-motion: reduce) respected
□ CSS layers managing cascade (no !important wars)
□ Selectors kept shallow (≤ 3 levels)
□ CSS code-split by route
□ Font-display set on all @font-face rules
□ No universal selectors in compound selectors
□ CSS custom properties for repeated values
□ Containment on scrollable areas
□ Total CSS ≤ 100KB compressed
```
