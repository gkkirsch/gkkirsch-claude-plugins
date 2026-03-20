# Animation & Motion Cheat Sheet

## Decision: CSS vs Framer Motion

| Need | Use CSS | Use Framer Motion |
|------|---------|-------------------|
| Hover effects | ✅ `transition` | Overkill |
| Loading spinners | ✅ `@keyframes` | Overkill |
| Enter animation | ✅ If no exit needed | ✅ If exit needed |
| Exit animation | ❌ Can't animate unmount | ✅ `AnimatePresence` |
| Layout changes | ❌ Expensive to animate | ✅ `layout` prop (FLIP) |
| Drag/swipe | ❌ No built-in support | ✅ `drag` prop |
| Scroll-linked | ✅ `animation-timeline` | ✅ `useScroll` |
| Page transitions | ⚠️ View Transitions API | ✅ More control |
| Reorder lists | ❌ No built-in support | ✅ `Reorder` component |
| SVG path drawing | ❌ Limited | ✅ `pathLength` |
| Spring physics | ❌ Approximation only | ✅ Native springs |

## Framer Motion Quick Reference

```tsx
// Mount animation
<motion.div
  initial={{ opacity: 0, y: 20 }}
  animate={{ opacity: 1, y: 0 }}
  transition={{ duration: 0.3 }}
/>

// Enter + Exit
<AnimatePresence mode="wait">
  {isVisible && (
    <motion.div
      key="unique"
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      exit={{ opacity: 0 }}
    />
  )}
</AnimatePresence>

// Hover + Tap
<motion.button
  whileHover={{ scale: 1.05 }}
  whileTap={{ scale: 0.95 }}
/>

// Layout animation
<motion.div layout />

// Shared layout (element moves between positions)
<motion.div layoutId="shared-id" />

// Staggered children
<motion.div
  variants={{
    show: { transition: { staggerChildren: 0.05 } },
  }}
  initial="hidden"
  animate="show"
>
  <motion.div variants={{ hidden: { opacity: 0 }, show: { opacity: 1 } }} />
</motion.div>

// Scroll-linked
const { scrollYProgress } = useScroll();
<motion.div style={{ scaleX: scrollYProgress }} />

// In viewport
const ref = useRef(null);
const isInView = useInView(ref, { once: true });
<motion.div ref={ref} animate={isInView ? { opacity: 1 } : { opacity: 0 }} />

// Drag
<motion.div drag="x" dragConstraints={{ left: -100, right: 100 }} />

// Reorder
<Reorder.Group values={items} onReorder={setItems}>
  {items.map(item => <Reorder.Item key={item} value={item} />)}
</Reorder.Group>
```

## CSS Quick Reference

```css
/* Transition */
.element {
  transition: transform 0.2s ease-out, opacity 0.2s ease-out;
}
.element:hover {
  transform: translateY(-4px);
}

/* Keyframe animation */
@keyframes fadeIn {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}
.element { animation: fadeIn 0.3s ease-out; }

/* Infinite */
@keyframes spin { to { transform: rotate(360deg); } }
.spinner { animation: spin 1s linear infinite; }

/* Scroll-driven (Chrome 115+) */
.progress { animation: grow linear; animation-timeline: scroll(); }
@keyframes grow { from { transform: scaleX(0); } to { transform: scaleX(1); } }
```

## Timing Guide

| Element | Duration | Easing |
|---------|----------|--------|
| Hover state | 100-150ms | ease-out |
| Button feedback | 100ms | ease-out |
| Tooltip | 150ms | ease-out |
| Dropdown/Menu | 200ms | ease-out |
| Modal enter | 200-300ms | spring or ease-out |
| Modal exit | 150-200ms | ease-in |
| Drawer/Sheet | 300ms | spring |
| Page transition | 200-300ms | ease-in-out |
| Stagger delay | 30-80ms | — |
| Scroll parallax | Continuous | linear |

## Spring Presets

| Name | Stiffness | Damping | Use For |
|------|-----------|---------|---------|
| Snappy | 400 | 30 | Buttons, toggles, tabs |
| Gentle | 120 | 14 | Modals, drawers, sheets |
| Bouncy | 300 | 10 | Fun UI, notifications |
| Stiff | 500 | 35 | Precise, minimal overshoot |

## Performance Checklist

- [ ] Only animate `transform` and `opacity` (compositor properties)
- [ ] `will-change` only on elements about to animate
- [ ] No `transition: all` (list specific properties)
- [ ] Reduced motion support (`prefers-reduced-motion`)
- [ ] No layout thrashing (batch reads/writes)
- [ ] Spring animations for interactive elements (natural feel)
- [ ] Stagger delay under 80ms (faster than that feels simultaneous)
- [ ] Exit animations shorter than enter (100-200ms)
- [ ] Test on low-end devices (4x CPU throttle in DevTools)

## Reduced Motion

```css
@media (prefers-reduced-motion: reduce) {
  *, *::before, *::after {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
  }
}
```

```tsx
import { useReducedMotion } from 'framer-motion';
const shouldReduce = useReducedMotion();
// Use shouldReduce to skip non-essential animations
```

## Common Easing Curves

```
ease-out   (enter):    cubic-bezier(0, 0, 0.2, 1)
ease-in    (exit):     cubic-bezier(0.4, 0, 1, 1)
ease-in-out (move):    cubic-bezier(0.4, 0, 0.2, 1)
overshoot  (playful):  cubic-bezier(0.34, 1.56, 0.64, 1)
```

## View Transitions API

```tsx
// Navigate with transition
document.startViewTransition(() => navigate('/page'));

// Shared element (same view-transition-name on both pages)
<img style={{ viewTransitionName: `hero-${id}` }} />

// CSS customization
::view-transition-old(root) { animation: fade-out 0.2s; }
::view-transition-new(root) { animation: fade-in 0.2s; }
```

Browser support: Chrome 111+, Edge 111+, Safari 18+, Firefox 126+.
