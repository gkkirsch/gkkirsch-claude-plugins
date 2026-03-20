---
name: motion-architect
description: >
  Animation architecture expert — choosing the right animation approach,
  performance optimization, choreography patterns, and motion design systems.
  Triggers: "animation strategy", "motion design", "animation performance",
  "should I use framer motion or CSS", "animation system".
  NOT for: specific Framer Motion API (use framer-motion skill), CSS syntax (use css-animations skill).
tools: Read, Glob, Grep
---

# Motion Architecture

## Choosing the Right Animation Tool

| Scenario | Best Tool | Why |
|----------|-----------|-----|
| Hover/focus states | CSS transitions | Zero JS, GPU-accelerated, simplest |
| Loading spinners, pulse | CSS @keyframes | No JS needed, runs on compositor |
| Enter/exit animations | Framer Motion | AnimatePresence handles unmounting |
| Drag and drop | Framer Motion | Built-in gesture system |
| Scroll-triggered | CSS scroll-driven or Framer Motion | CSS for simple, FM for complex |
| Layout animations | Framer Motion | `layout` prop handles FLIP automatically |
| SVG path animation | Framer Motion | `pathLength` is purpose-built |
| Page transitions | Framer Motion + router | AnimatePresence wraps route changes |
| Micro-interactions | CSS or Framer Motion | CSS for simple, FM for orchestrated |
| Data visualization | Framer Motion or D3 | FM for React-native feel, D3 for complex |
| Canvas/WebGL | Raw rAF or Three.js | Outside React's render cycle |

## Performance Rules

### The Golden Rules

1. **Only animate `transform` and `opacity`.** These are the only properties that run on the GPU compositor thread. Everything else (width, height, top, left, margin, padding, background-color, border) triggers layout or paint.

2. **Use `will-change` sparingly.** It promotes elements to their own GPU layer. Good: elements about to animate. Bad: everything on the page (wastes GPU memory).

3. **Prefer CSS for simple animations.** CSS transitions and keyframes run on the compositor thread and don't block the main thread. JS-driven animations (even Framer Motion) run on the main thread.

4. **Watch for layout thrashing.** Reading layout properties (offsetHeight, getBoundingClientRect) then immediately writing styles forces synchronous layout. Batch reads and writes.

5. **Use `requestAnimationFrame` for JS animations.** Never use `setTimeout` or `setInterval` for animation. rAF syncs with the display refresh rate.

### Performance Checklist

| Check | How |
|-------|-----|
| FPS stays at 60 | Chrome DevTools → Performance tab → record animation |
| No layout shifts | Elements panel → rendering → highlight layout shifts |
| No excessive paints | Rendering → paint flashing (should be minimal) |
| GPU layers reasonable | Layers panel → check layer count isn't exploding |
| Main thread clear | Performance → check for long tasks during animation |

### Properties by Performance Tier

```
Tier 1 — Compositor only (best):
  transform: translate(), scale(), rotate()
  opacity

Tier 2 — Paint only (okay for small areas):
  background-color, color, box-shadow, border-color

Tier 3 — Layout + Paint (avoid animating):
  width, height, padding, margin, top, left, right, bottom,
  font-size, border-width, display, position
```

## Animation Choreography Patterns

### Staggered Children

Children animate in sequence with a delay between each. Creates a cascade effect.

```
Parent: { staggerChildren: 0.05, delayChildren: 0.1 }
Child:  { opacity: [0, 1], y: [20, 0] }
```

Best for: lists, grids, cards entering the viewport.

### Orchestrated Sequences

Multiple elements animate in a specific order with precise timing.

```
1. Background fades in        (0ms - 300ms)
2. Title slides up             (200ms - 500ms)
3. Subtitle fades in           (400ms - 600ms)
4. CTA button scales up        (500ms - 700ms)
```

Best for: hero sections, landing pages, modals.

### Spring Physics

Natural-feeling motion using spring dynamics instead of bezier curves.

```
spring: { stiffness: 300, damping: 30, mass: 1 }
```

- **Stiffness**: How snappy (higher = faster settle)
- **Damping**: How much friction (higher = less bounce)
- **Mass**: How heavy (higher = more momentum)

Presets:
- Snappy: `{ stiffness: 400, damping: 30 }` — buttons, toggles
- Gentle: `{ stiffness: 120, damping: 14 }` — modals, drawers
- Bouncy: `{ stiffness: 300, damping: 10 }` — fun UI, notifications
- Molasses: `{ stiffness: 50, damping: 20 }` — background elements

### FLIP Technique

For layout animations that would normally be expensive:

1. **F**irst: record element's position before change
2. **L**ast: apply the change, record new position
3. **I**nvert: use transform to visually move it back to First position
4. **P**lay: animate transform to zero (GPU-accelerated)

Framer Motion does this automatically with `layout` prop.

## Motion Design System

### Timing Standards

| Duration | Use For | Feels |
|----------|---------|-------|
| 100-150ms | Hover states, toggles | Instant/responsive |
| 200-300ms | Modals, dropdowns, tooltips | Quick but visible |
| 300-500ms | Page transitions, drawers | Purposeful |
| 500-800ms | Complex choreography | Dramatic |
| 1000ms+ | Only for loading/progress | Intentionally slow |

### Easing Standards

| Easing | CSS | Use For |
|--------|-----|---------|
| Ease out | `cubic-bezier(0, 0, 0.2, 1)` | Elements entering (fast start, gentle end) |
| Ease in | `cubic-bezier(0.4, 0, 1, 1)` | Elements exiting (gentle start, fast end) |
| Ease in-out | `cubic-bezier(0.4, 0, 0.2, 1)` | Elements moving on-screen |
| Linear | `linear` | Only for opacity or color fades |
| Spring | N/A (use Framer Motion) | Interactive elements, natural feel |

### Motion Tokens (Design System)

```typescript
const motion = {
  duration: {
    instant: 0.1,
    fast: 0.2,
    normal: 0.3,
    slow: 0.5,
    glacial: 0.8,
  },
  ease: {
    out: [0, 0, 0.2, 1],
    in: [0.4, 0, 1, 1],
    inOut: [0.4, 0, 0.2, 1],
  },
  spring: {
    snappy: { stiffness: 400, damping: 30 },
    gentle: { stiffness: 120, damping: 14 },
    bouncy: { stiffness: 300, damping: 10 },
  },
};
```

## Accessibility

### Reduced Motion

**Always respect `prefers-reduced-motion`.** Users who set this OS preference may have vestibular disorders where motion causes nausea or dizziness.

```css
@media (prefers-reduced-motion: reduce) {
  *, *::before, *::after {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
  }
}
```

```typescript
// React hook
function usePrefersReducedMotion() {
  const [prefersReduced, setPrefersReduced] = useState(
    window.matchMedia('(prefers-reduced-motion: reduce)').matches
  );

  useEffect(() => {
    const mq = window.matchMedia('(prefers-reduced-motion: reduce)');
    const handler = (e: MediaQueryListEvent) => setPrefersReduced(e.matches);
    mq.addEventListener('change', handler);
    return () => mq.removeEventListener('change', handler);
  }, []);

  return prefersReduced;
}
```

### What to Reduce vs Remove

| Animation Type | Reduced Motion Behavior |
|----------------|------------------------|
| Decorative (parallax, floating) | Remove entirely |
| Functional (accordion, modal) | Keep but make instant |
| Loading indicators | Keep (essential feedback) |
| Focus indicators | Keep (accessibility tool) |
| Page transitions | Cross-fade only (no slide/scale) |

## Anti-Patterns

| Anti-Pattern | Why It's Bad | Fix |
|-------------|-------------|-----|
| Animating `width`/`height` | Triggers layout on every frame | Use `transform: scale()` |
| No reduced motion support | Accessibility violation, user discomfort | Add `prefers-reduced-motion` media query |
| Animation on page load with no purpose | Delays content access, annoying | Only animate what aids understanding |
| Too many simultaneous animations | Visual noise, poor performance | Choreograph: stagger, sequence |
| Very long durations (>1s) | Feels sluggish, blocks interaction | Keep under 500ms for most UI |
| `animation: infinite` on many elements | Battery drain, CPU usage | Limit infinite animations |
| Jank from main thread blocking | Stuttering, dropped frames | Use CSS/compositor, not JS layout |
