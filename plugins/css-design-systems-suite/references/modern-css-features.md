# Modern CSS Features Reference (2024–2026)

Quick reference for modern CSS features, browser support status, and usage patterns.

---

## Feature Support Matrix

| Feature | Chrome | Firefox | Safari | Status |
|---------|--------|---------|--------|--------|
| Cascade Layers (`@layer`) | 99+ | 97+ | 15.4+ | Baseline |
| Container Queries (size) | 105+ | 110+ | 16+ | Baseline |
| Container Query Units (cqi) | 105+ | 110+ | 16+ | Baseline |
| Style Container Queries | 111+ | ❌ | ❌ | Limited |
| `:has()` | 105+ | 121+ | 15.4+ | Baseline |
| CSS Nesting | 120+ | 117+ | 17.2+ | Baseline |
| `@scope` | 118+ | ❌ | 17.4+ | Limited |
| Anchor Positioning | 125+ | ❌ | ❌ | Limited |
| `oklch()` / `oklab()` | 111+ | 113+ | 15.4+ | Baseline |
| Relative color syntax | 119+ | 128+ | 16.4+ | Baseline |
| `color-mix()` | 111+ | 113+ | 16.2+ | Baseline |
| Subgrid | 117+ | 71+ | 16+ | Baseline |
| `text-wrap: balance` | 114+ | 121+ | 17.5+ | Baseline |
| `text-wrap: pretty` | 117+ | ❌ | ❌ | Limited |
| View Transitions (same-doc) | 111+ | ❌ | 18+ | Limited |
| View Transitions (cross-doc) | 126+ | ❌ | 18+ | Limited |
| Scroll-driven animations | 115+ | ❌ | ❌ | Limited |
| Popover API | 114+ | 125+ | 17+ | Baseline |
| `@starting-style` | 117+ | ❌ | 17.5+ | Limited |
| `field-sizing: content` | 123+ | ❌ | ❌ | Limited |
| `light-dark()` | 123+ | 120+ | 17.5+ | Baseline |
| `@property` | 85+ | 128+ | 15.4+ | Baseline |
| Logical Properties | 87+ | 66+ | 15+ | Baseline |
| `dvh` / `svh` / `lvh` | 108+ | 101+ | 15.4+ | Baseline |
| `initial-letter` | ❌ | ❌ | 9+ | Limited |
| Masonry layout | ❌ (flag) | ❌ (flag) | ❌ | Experimental |
| `interpolate-size: allow-keywords` | 129+ | ❌ | ❌ | Limited |

---

## Cascade Layers

```css
/* Declare layer order */
@layer reset, base, tokens, components, utilities, overrides;

/* Assign styles to layers */
@layer components {
  .card { /* ... */ }
}

/* Import external CSS into a layer */
@import url("lib.css") layer(vendor);

/* Anonymous layer (always lowest priority) */
@layer {
  /* These styles have lowest layer priority */
}

/* Nested layers */
@layer components.buttons {
  .btn { /* ... */ }
}

/* Un-layered styles ALWAYS win over layered styles */
```

Key facts:
- Layer order is determined by the FIRST `@layer` declaration
- Later-declared layers have HIGHER priority
- Un-layered styles beat all layered styles
- `!important` reverses layer order (earlier layers win)
- Layers can be nested: `components.buttons`

---

## Container Queries

```css
/* Establish containment */
.container {
  container-type: inline-size;        /* Query inline dimension */
  container-type: size;               /* Query both dimensions (rare) */
  container-name: card;               /* Name for targeted queries */
  container: card / inline-size;      /* Shorthand */
}

/* Size queries */
@container (inline-size > 400px) { }
@container (inline-size < 200px) { }
@container (400px <= inline-size <= 800px) { }

/* Named container queries */
@container card (inline-size > 500px) { }

/* Container query units */
/* cqi = 1% of container inline size */
/* cqb = 1% of container block size */
/* cqw = 1% of container width */
/* cqh = 1% of container height */
/* cqmin = smaller of cqi/cqb */
/* cqmax = larger of cqi/cqb */
.text { font-size: clamp(1rem, 3cqi, 2rem); }

/* Style queries (Chrome only) */
@container style(--variant: compact) { }
@container card style(--theme: dark) { }
```

---

## :has() Selector

```css
/* Parent selector */
.card:has(img) { }                     /* Card that contains an image */
.card:has(> img) { }                   /* Card with direct child image */

/* Sibling awareness */
h2:has(+ p) { }                        /* h2 directly followed by p */

/* Negation */
.form:not(:has(:invalid)) { }          /* Form with no invalid inputs */

/* Quantity queries */
.list:has(:nth-child(4)) { }           /* List with 4+ children */
.list:not(:has(:nth-child(2))) { }     /* List with exactly 1 child */

/* Complex patterns */
.card:has(.badge[data-type="new"]) { } /* Card containing a "new" badge */
:has(dialog[open]) body { }            /* Body when dialog is open (won't work — :has can't go up past subject) */

/* State-based */
.group:has(:focus-visible) { }         /* Group containing focused element */
.form-group:has(input:checked) { }     /* Group with checked input */
```

---

## CSS Nesting

```css
.component {
  /* Nested selector with & */
  & .child { }

  /* Pseudo-classes */
  &:hover { }
  &:focus-visible { }

  /* Pseudo-elements */
  &::before { }
  &::after { }

  /* Combinators */
  & > .direct-child { }
  & + .adjacent-sibling { }
  & ~ .general-sibling { }

  /* Media queries */
  @media (width >= 768px) { }

  /* Container queries */
  @container (inline-size > 500px) { }

  /* Nested nesting */
  & .child {
    & .grandchild { }
  }

  /* Reverse nesting — select ancestor context */
  .dark & { }           /* When inside .dark */
  [dir="rtl"] & { }     /* When inside RTL context */
}
```

---

## @scope

```css
/* Basic scope */
@scope (.card) {
  p { color: gray; }        /* Only p inside .card */
  .title { font-weight: 700; }
}

/* Scope with lower boundary (donut scope) */
@scope (.card) to (.card__footer) {
  /* Styles apply inside .card but NOT inside .card__footer */
  a { color: blue; }
}

/* Scope proximity — nearest scope wins */
@scope (.light) {
  p { color: black; }
}
@scope (.dark) {
  p { color: white; }
}
/* A <p> inside .dark inside .light gets white (nearest scope) */

/* Implicit scope (in <style> within a component) */
@scope {
  /* Scoped to the parent element of this <style> */
  p { color: gray; }
}
```

---

## Anchor Positioning

```css
/* Define an anchor */
.trigger {
  anchor-name: --my-anchor;
}

/* Position relative to anchor */
.popover {
  position: fixed;
  position-anchor: --my-anchor;

  /* Position at anchor edges */
  top: anchor(bottom);           /* Below the anchor */
  left: anchor(left);            /* Aligned to anchor's left */
  right: anchor(right);          /* Aligned to anchor's right */
  bottom: anchor(top);           /* Above the anchor */

  /* Center alignment */
  justify-self: anchor-center;   /* Horizontally centered on anchor */
  align-self: anchor-center;     /* Vertically centered on anchor */

  /* Offset from anchor */
  top: calc(anchor(bottom) + 8px);

  /* Fallback positioning */
  position-try-fallbacks: flip-block, flip-inline;
}

/* Named position-try */
@position-try --below-right {
  top: anchor(bottom);
  left: anchor(right);
}
```

---

## Color Functions

```css
/* oklch(lightness chroma hue) */
color: oklch(55% 0.25 260);
color: oklch(55% 0.25 260 / 0.5);    /* With alpha */

/* Relative color syntax */
color: oklch(from var(--base) calc(l + 0.1) c h);    /* Lighten */
color: oklch(from var(--base) calc(l - 0.1) c h);    /* Darken */
color: oklch(from var(--base) l calc(c * 0.5) h);    /* Desaturate */
color: oklch(from var(--base) l c calc(h + 180));     /* Complement */
color: oklch(from var(--base) l c h / 0.5);           /* Semi-transparent */

/* color-mix() */
color: color-mix(in oklch, red, blue);                /* 50/50 mix */
color: color-mix(in oklch, red 30%, blue);            /* 30% red, 70% blue */
color: color-mix(in oklch, var(--brand), transparent 80%); /* Ghost color */

/* light-dark() — auto-switches based on color-scheme */
color-scheme: light dark;
color: light-dark(#333, #ccc);
background: light-dark(white, #1a1a1a);
```

---

## View Transitions API

```css
/* Enable cross-document transitions */
@view-transition {
  navigation: auto;
}

/* Name elements for shared transitions */
.element {
  view-transition-name: hero;
}

/* Style transition pseudo-elements */
::view-transition-group(hero) {
  animation-duration: 0.4s;
}

::view-transition-old(hero) {
  animation: fade-out 0.2s ease;
}

::view-transition-new(hero) {
  animation: fade-in 0.3s ease;
}

/* Root page transition */
::view-transition-old(root) { }
::view-transition-new(root) { }
```

```js
// Same-document transition
document.startViewTransition(() => {
  updateDOM();
});

// With types
document.startViewTransition({
  update: () => updateDOM(),
  types: ['slide-left'],
});
```

---

## Scroll-Driven Animations

```css
/* Scroll progress animation */
animation-timeline: scroll();                   /* Nearest scroll ancestor */
animation-timeline: scroll(root);               /* Document scroll */
animation-timeline: scroll(root block);         /* Block direction */
animation-timeline: scroll(nearest inline);     /* Inline direction */

/* View progress animation (element visibility) */
animation-timeline: view();
animation-range: entry 0% entry 100%;          /* During entry */
animation-range: exit 0% exit 100%;            /* During exit */
animation-range: cover 0% cover 100%;          /* Full cover */
animation-range: contain 0% contain 100%;      /* Contained */

/* Named scroll timeline */
.scroller {
  scroll-timeline: --my-scroll block;
}

.animated {
  animation-timeline: --my-scroll;
}

/* Named view timeline */
.observed {
  view-timeline: --my-view block;
}

.animated {
  animation-timeline: --my-view;
}
```

---

## @starting-style

Enables transitions for elements entering the DOM or changing to `display: block`:

```css
.popover {
  opacity: 1;
  transform: scale(1);
  transition: opacity 0.3s, transform 0.3s, display 0.3s allow-discrete;

  /* Starting state — applied when element first renders */
  @starting-style {
    opacity: 0;
    transform: scale(0.95);
  }

  /* Exit state */
  &:not(:popover-open) {
    opacity: 0;
    transform: scale(0.95);
  }
}

/* Dialog enter animation */
dialog[open] {
  opacity: 1;
  transform: translateY(0);
  transition: opacity 0.3s, transform 0.3s, display 0.3s allow-discrete;

  @starting-style {
    opacity: 0;
    transform: translateY(20px);
  }
}
```

---

## Popover API

```html
<button popovertarget="menu">Open Menu</button>
<div id="menu" popover>
  <!-- Popover content — auto-dismissed on outside click/Escape -->
</div>

<!-- Manual popover (no auto-dismiss) -->
<div id="dialog" popover="manual">
  <button popovertarget="dialog" popovertargetaction="hide">Close</button>
</div>
```

```css
/* Style popover */
[popover] {
  margin: auto;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: var(--space-m);
  box-shadow: var(--shadow-lg);
}

/* Backdrop */
[popover]::backdrop {
  background: oklch(0% 0 0 / 0.3);
}

/* Entry/exit animation */
[popover] {
  opacity: 1;
  transform: scale(1);
  transition: opacity 0.2s, transform 0.2s, display 0.2s allow-discrete;

  @starting-style {
    opacity: 0;
    transform: scale(0.95);
  }

  &:not(:popover-open) {
    opacity: 0;
    transform: scale(0.95);
  }
}
```

---

## @property (Registered Custom Properties)

```css
/* Register a custom property with type, inheritance, and initial value */
@property --hue {
  syntax: "<number>";
  inherits: true;
  initial-value: 260;
}

@property --progress {
  syntax: "<percentage>";
  inherits: false;
  initial-value: 0%;
}

@property --gradient-angle {
  syntax: "<angle>";
  inherits: false;
  initial-value: 0deg;
}

/* Now you can ANIMATE custom properties! */
.animated-gradient {
  --gradient-angle: 0deg;
  background: conic-gradient(from var(--gradient-angle), red, blue, red);
  animation: rotate-gradient 3s linear infinite;
}

@keyframes rotate-gradient {
  to { --gradient-angle: 360deg; }
}

/* Animated progress ring */
.progress-ring {
  --progress: 0%;
  background: conic-gradient(
    var(--color-brand) var(--progress),
    var(--color-gray-200) var(--progress)
  );
  transition: --progress 1s ease;
}
```

---

## Logical Properties Quick Reference

| Physical | Logical (Inline/Block) |
|----------|----------------------|
| `width` | `inline-size` |
| `height` | `block-size` |
| `min-width` | `min-inline-size` |
| `max-height` | `max-block-size` |
| `margin-left` | `margin-inline-start` |
| `margin-right` | `margin-inline-end` |
| `padding-top` | `padding-block-start` |
| `padding-bottom` | `padding-block-end` |
| `border-left` | `border-inline-start` |
| `top` | `inset-block-start` |
| `right` | `inset-inline-end` |
| `bottom` | `inset-block-end` |
| `left` | `inset-inline-start` |
| `text-align: left` | `text-align: start` |
| `text-align: right` | `text-align: end` |
| `float: left` | `float: inline-start` |
| `border-top-left-radius` | `border-start-start-radius` |
| `overflow-x` | `overflow-inline` |
| `overflow-y` | `overflow-block` |

Shorthand:
```css
margin-inline: 1rem 2rem;    /* start end */
margin-block: 1rem;          /* top and bottom */
padding-inline: 1rem;
inset-inline: 0;             /* left: 0; right: 0 */
inset-block: 0;
border-inline: 1px solid;
```

---

## Progressive Enhancement Patterns

```css
/* Feature detection with @supports */
@supports (container-type: inline-size) {
  .card-grid { container-type: inline-size; }
}

@supports selector(:has(*)) {
  .form:has(:invalid) .submit { opacity: 0.5; }
}

@supports (anchor-name: --x) {
  .tooltip { position-anchor: --trigger; }
}

/* Fallback-first pattern */
.layout {
  /* Fallback: flexbox */
  display: flex;
  flex-wrap: wrap;
  gap: 1rem;
}

@supports (grid-template-rows: subgrid) {
  .layout {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
  }

  .layout > .card {
    display: grid;
    grid-template-rows: subgrid;
    grid-row: span 3;
  }
}
```
