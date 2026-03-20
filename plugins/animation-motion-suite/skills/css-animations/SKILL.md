---
name: css-animations
description: >
  CSS animations and transitions — keyframes, transitions, transform,
  scroll-driven animations, and performance-optimized motion.
  Triggers: "CSS animation", "CSS transition", "@keyframes", "transform",
  "CSS motion", "scroll animation CSS", "will-change".
  NOT for: Framer Motion API (use framer-motion skill), architecture (use motion-architect agent).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# CSS Animations

## Transitions

```css
/* Basic transition */
.button {
  background-color: #3b82f6;
  transform: scale(1);
  transition: all 0.2s ease-out;
}

.button:hover {
  background-color: #2563eb;
  transform: scale(1.05);
}

/* Individual properties (preferred — more explicit) */
.card {
  transition:
    transform 0.2s ease-out,
    box-shadow 0.2s ease-out,
    opacity 0.15s ease-out;
}

.card:hover {
  transform: translateY(-4px);
  box-shadow: 0 12px 24px rgba(0, 0, 0, 0.15);
}
```

### Transition Properties

```css
transition: <property> <duration> <timing-function> <delay>;

/* Individual */
transition-property: transform, opacity;
transition-duration: 0.3s;
transition-timing-function: ease-out;
transition-delay: 0s;

/* Common timing functions */
transition-timing-function: ease;                    /* default */
transition-timing-function: ease-in;                 /* slow start */
transition-timing-function: ease-out;                /* slow end (best for enter) */
transition-timing-function: ease-in-out;             /* slow start and end */
transition-timing-function: linear;                  /* constant speed */
transition-timing-function: cubic-bezier(0.4, 0, 0.2, 1);  /* custom */

/* Spring-like with cubic-bezier */
transition-timing-function: cubic-bezier(0.34, 1.56, 0.64, 1);  /* overshoot */
```

## @keyframes

```css
/* Fade in and slide up */
@keyframes fadeInUp {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.element {
  animation: fadeInUp 0.3s ease-out;
}

/* Multi-step */
@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.loading {
  animation: pulse 2s ease-in-out infinite;
}

/* Spin */
@keyframes spin {
  to { transform: rotate(360deg); }
}

.spinner {
  animation: spin 1s linear infinite;
}

/* Bounce */
@keyframes bounce {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-25%); }
}

/* Shake (error feedback) */
@keyframes shake {
  0%, 100% { transform: translateX(0); }
  10%, 30%, 50%, 70%, 90% { transform: translateX(-4px); }
  20%, 40%, 60%, 80% { transform: translateX(4px); }
}

.error {
  animation: shake 0.5s ease-in-out;
}
```

### Animation Properties

```css
animation: <name> <duration> <timing> <delay> <iteration> <direction> <fill> <play-state>;

/* Individual */
animation-name: fadeInUp;
animation-duration: 0.3s;
animation-timing-function: ease-out;
animation-delay: 0s;
animation-iteration-count: 1;           /* or infinite */
animation-direction: normal;            /* or reverse, alternate, alternate-reverse */
animation-fill-mode: forwards;          /* keeps end state */
animation-play-state: running;          /* or paused */

/* Common fill modes */
animation-fill-mode: none;              /* default — reverts after animation */
animation-fill-mode: forwards;          /* keeps final keyframe state */
animation-fill-mode: backwards;         /* applies first keyframe during delay */
animation-fill-mode: both;              /* applies both forwards and backwards */
```

## Transform

```css
/* Translate (move) */
transform: translateX(100px);
transform: translateY(-50%);
transform: translate(10px, 20px);
transform: translate3d(10px, 20px, 0);     /* forces GPU layer */

/* Scale */
transform: scale(1.5);                      /* uniform */
transform: scaleX(0.5);                     /* horizontal only */
transform: scale(1.1, 1.2);                 /* non-uniform */

/* Rotate */
transform: rotate(45deg);
transform: rotateX(45deg);                   /* 3D */
transform: rotateY(180deg);                  /* 3D flip */
transform: rotate3d(1, 1, 0, 45deg);        /* custom axis */

/* Skew */
transform: skewX(10deg);

/* Combined (order matters!) */
transform: translateX(100px) scale(1.2) rotate(45deg);

/* Transform origin */
transform-origin: center;                   /* default */
transform-origin: top left;
transform-origin: 50% 100%;                 /* bottom center */

/* 3D perspective */
perspective: 1000px;                         /* on parent */
transform: perspective(1000px) rotateY(30deg);  /* on element */
```

## Scroll-Driven Animations (Modern CSS)

```css
/* Scroll progress bar (no JavaScript needed) */
.progress-bar {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 3px;
  background: #3b82f6;
  transform-origin: left;
  animation: scaleX linear;
  animation-timeline: scroll();
}

@keyframes scaleX {
  from { transform: scaleX(0); }
  to { transform: scaleX(1); }
}

/* Element animation tied to scroll position */
.fade-in-on-scroll {
  animation: fadeInUp linear both;
  animation-timeline: view();
  animation-range: entry 0% entry 100%;
}

/* Parallax with scroll() */
.parallax-bg {
  animation: parallax linear;
  animation-timeline: scroll();
}

@keyframes parallax {
  from { transform: translateY(0); }
  to { transform: translateY(-30%); }
}
```

### animation-timeline Values

```css
/* Scroll of nearest scroll container */
animation-timeline: scroll();

/* Scroll of specific axis */
animation-timeline: scroll(y);
animation-timeline: scroll(x);

/* Element enters/exits viewport */
animation-timeline: view();

/* animation-range controls when animation starts/ends */
animation-range: entry 0% entry 100%;    /* as element enters */
animation-range: cover 0% cover 100%;    /* full scroll through viewport */
animation-range: contain 0% contain 100%; /* while fully visible */
animation-range: exit 0% exit 100%;      /* as element exits */
```

## Practical Patterns

### Skeleton Loading

```css
.skeleton {
  background: linear-gradient(
    90deg,
    #f0f0f0 25%,
    #e0e0e0 50%,
    #f0f0f0 75%
  );
  background-size: 200% 100%;
  animation: shimmer 1.5s ease-in-out infinite;
  border-radius: 4px;
}

@keyframes shimmer {
  0% { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}
```

### Staggered Entrance (CSS-only)

```css
.card {
  opacity: 0;
  transform: translateY(20px);
  animation: fadeInUp 0.3s ease-out forwards;
}

.card:nth-child(1) { animation-delay: 0.0s; }
.card:nth-child(2) { animation-delay: 0.05s; }
.card:nth-child(3) { animation-delay: 0.1s; }
.card:nth-child(4) { animation-delay: 0.15s; }
.card:nth-child(5) { animation-delay: 0.2s; }

@keyframes fadeInUp {
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
```

### Tooltip

```css
.tooltip {
  opacity: 0;
  transform: translateY(4px);
  pointer-events: none;
  transition: opacity 0.15s ease-out, transform 0.15s ease-out;
}

.trigger:hover .tooltip {
  opacity: 1;
  transform: translateY(0);
  pointer-events: auto;
}
```

### Smooth Accordion

```css
.accordion-content {
  display: grid;
  grid-template-rows: 0fr;
  transition: grid-template-rows 0.3s ease-out;
}

.accordion-content[data-open="true"] {
  grid-template-rows: 1fr;
}

.accordion-content > div {
  overflow: hidden;
}
```

### Card Hover (3D Tilt)

```css
.card-3d {
  transition: transform 0.3s ease-out;
  transform-style: preserve-3d;
  perspective: 1000px;
}

.card-3d:hover {
  transform: rotateY(5deg) rotateX(5deg) translateZ(10px);
  box-shadow: -5px 5px 20px rgba(0, 0, 0, 0.2);
}
```

## Performance

```css
/* Promote to GPU layer (use sparingly) */
.animated-element {
  will-change: transform, opacity;
}

/* Remove after animation completes */
.animated-element.done {
  will-change: auto;
}

/* Force GPU compositing */
.gpu-layer {
  transform: translateZ(0);         /* hack — creates GPU layer */
  backface-visibility: hidden;      /* prevents flicker on some browsers */
}

/* Contain layout for performance */
.card {
  contain: layout style;            /* limits browser reflow scope */
}
```

## Reduced Motion

```css
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

/* Or per-element */
@media (prefers-reduced-motion: reduce) {
  .parallax { transform: none; }
  .fancy-entrance { opacity: 1; transform: none; }
}
```

## Tailwind CSS Animation Utilities

```html
<!-- Built-in -->
<div class="animate-spin" />        <!-- continuous spin -->
<div class="animate-ping" />        <!-- expanding ping -->
<div class="animate-pulse" />       <!-- opacity pulse -->
<div class="animate-bounce" />      <!-- bouncing -->

<!-- Transitions -->
<div class="transition-all duration-200 ease-out hover:scale-105" />
<div class="transition-colors duration-150 hover:bg-blue-600" />
<div class="transition-transform duration-300 hover:-translate-y-1" />

<!-- Custom in tailwind.config -->
module.exports = {
  theme: {
    extend: {
      animation: {
        'fade-in': 'fadeIn 0.3s ease-out',
        'slide-up': 'slideUp 0.3s ease-out',
        'shimmer': 'shimmer 1.5s ease-in-out infinite',
      },
      keyframes: {
        fadeIn: {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' },
        },
        slideUp: {
          '0%': { opacity: '0', transform: 'translateY(10px)' },
          '100%': { opacity: '1', transform: 'translateY(0)' },
        },
      },
    },
  },
};
```

## Gotchas

1. **`transition: all` is lazy.** It transitions EVERY property change, including ones you don't intend. List specific properties: `transition: transform 0.2s, opacity 0.2s`.

2. **`display: none` can't be transitioned.** Use `opacity` + `visibility` + `pointer-events` instead, or the new `@starting-style` + `transition-behavior: allow-discrete` in modern browsers.

3. **`animation-fill-mode: forwards` can cause stacking issues.** The element keeps its animated state, which might interfere with hover states or later style changes. Consider using `both` or removing the animation class after completion.

4. **`transform` creates a new stacking context.** Adding any transform (even `translateZ(0)`) creates a new stacking context, which can break `z-index` relationships.

5. **Scroll-driven animations need browser support check.** `animation-timeline: scroll()` is supported in Chrome 115+ and Firefox 110+ but NOT Safari (as of March 2026). Use feature detection or Framer Motion as fallback.

6. **`will-change` is not free.** Every element with `will-change` gets its own GPU layer. Too many layers waste GPU memory. Only add it to elements about to animate, and remove it after.
