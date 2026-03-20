---
name: framer-motion
description: >
  Framer Motion for React — animate, AnimatePresence, layout animations,
  gestures, scroll-triggered animations, variants, and orchestration.
  Triggers: "framer motion", "motion.div", "AnimatePresence", "layout animation",
  "drag", "gesture", "variants", "useScroll", "useTransform".
  NOT for: CSS-only animations (use css-animations skill), architecture decisions (use motion-architect agent).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Framer Motion

## Quick Start

```bash
npm install framer-motion
```

## Basic Animation

```tsx
import { motion } from 'framer-motion';

// Animate on mount
<motion.div
  initial={{ opacity: 0, y: 20 }}
  animate={{ opacity: 1, y: 0 }}
  transition={{ duration: 0.3 }}
>
  Hello
</motion.div>

// Animate between states
function Toggle() {
  const [isOpen, setIsOpen] = useState(false);

  return (
    <motion.div
      animate={{
        height: isOpen ? 'auto' : 0,
        opacity: isOpen ? 1 : 0,
      }}
      transition={{ duration: 0.2 }}
    >
      Content
    </motion.div>
  );
}
```

## AnimatePresence (Enter/Exit)

```tsx
import { AnimatePresence, motion } from 'framer-motion';

function Notifications({ items }: { items: Notification[] }) {
  return (
    <AnimatePresence>
      {items.map((item) => (
        <motion.div
          key={item.id}
          initial={{ opacity: 0, x: 50 }}
          animate={{ opacity: 1, x: 0 }}
          exit={{ opacity: 0, x: -50 }}
          transition={{ duration: 0.2 }}
        >
          {item.message}
        </motion.div>
      ))}
    </AnimatePresence>
  );
}

// Mode controls how enter/exit overlap
<AnimatePresence mode="wait">     {/* exit completes before enter starts */}
<AnimatePresence mode="popLayout"> {/* exiting elements removed from flow */}
<AnimatePresence mode="sync">      {/* enter and exit happen simultaneously */}
```

## Variants (Orchestrated Animations)

```tsx
const container = {
  hidden: { opacity: 0 },
  show: {
    opacity: 1,
    transition: {
      staggerChildren: 0.05,
      delayChildren: 0.1,
    },
  },
  exit: {
    opacity: 0,
    transition: { staggerChildren: 0.03, staggerDirection: -1 },
  },
};

const item = {
  hidden: { opacity: 0, y: 20 },
  show: { opacity: 1, y: 0 },
  exit: { opacity: 0, y: -10 },
};

function CardGrid({ cards }: { cards: Card[] }) {
  return (
    <motion.div
      variants={container}
      initial="hidden"
      animate="show"
      exit="exit"
      className="grid grid-cols-3 gap-4"
    >
      {cards.map((card) => (
        <motion.div key={card.id} variants={item} className="p-4 rounded-lg border">
          {card.title}
        </motion.div>
      ))}
    </motion.div>
  );
}
```

## Layout Animations

```tsx
// Automatic FLIP animation when layout changes
function ExpandableCard() {
  const [isExpanded, setIsExpanded] = useState(false);

  return (
    <motion.div
      layout
      onClick={() => setIsExpanded(!isExpanded)}
      className={isExpanded ? 'w-full h-96' : 'w-48 h-48'}
      transition={{ type: 'spring', stiffness: 300, damping: 30 }}
    >
      <motion.h2 layout="position">Title</motion.h2>
      {isExpanded && (
        <motion.p
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ delay: 0.2 }}
        >
          Expanded content here
        </motion.p>
      )}
    </motion.div>
  );
}

// Shared layout animation (element moves between components)
function TabContent({ activeTab }: { activeTab: string }) {
  return (
    <div className="flex gap-4">
      {tabs.map((tab) => (
        <button key={tab.id} onClick={() => setActiveTab(tab.id)}>
          {tab.label}
          {activeTab === tab.id && (
            <motion.div
              layoutId="activeTab"
              className="absolute bottom-0 left-0 right-0 h-0.5 bg-blue-600"
              transition={{ type: 'spring', stiffness: 400, damping: 30 }}
            />
          )}
        </button>
      ))}
    </div>
  );
}
```

## Gestures

```tsx
// Hover and tap
<motion.button
  whileHover={{ scale: 1.05 }}
  whileTap={{ scale: 0.95 }}
  transition={{ type: 'spring', stiffness: 400, damping: 15 }}
>
  Click me
</motion.button>

// Drag
<motion.div
  drag                          // enable both axes
  drag="x"                     // horizontal only
  dragConstraints={{ left: -100, right: 100 }}
  dragElastic={0.2}            // 0 = hard stop, 1 = full elastic
  dragSnapToOrigin             // snap back when released
  onDragEnd={(e, info) => {
    if (info.offset.x > 100) handleSwipeRight();
    if (info.offset.x < -100) handleSwipeLeft();
  }}
  whileDrag={{ scale: 1.1, cursor: 'grabbing' }}
>
  Drag me
</motion.div>

// Drag to reorder
import { Reorder } from 'framer-motion';

function SortableList() {
  const [items, setItems] = useState(['Item 1', 'Item 2', 'Item 3']);

  return (
    <Reorder.Group axis="y" values={items} onReorder={setItems}>
      {items.map((item) => (
        <Reorder.Item key={item} value={item}>
          {item}
        </Reorder.Item>
      ))}
    </Reorder.Group>
  );
}
```

## Scroll Animations

```tsx
import { motion, useScroll, useTransform, useSpring, useInView } from 'framer-motion';

// Scroll-linked progress bar
function ScrollProgressBar() {
  const { scrollYProgress } = useScroll();

  return (
    <motion.div
      className="fixed top-0 left-0 right-0 h-1 bg-blue-600 origin-left z-50"
      style={{ scaleX: scrollYProgress }}
    />
  );
}

// Parallax effect
function ParallaxHero() {
  const { scrollY } = useScroll();
  const y = useTransform(scrollY, [0, 500], [0, -150]);
  const opacity = useTransform(scrollY, [0, 300], [1, 0]);

  return (
    <motion.div style={{ y, opacity }} className="h-screen">
      <h1>Hero Content</h1>
    </motion.div>
  );
}

// Scroll-triggered (animate when element enters viewport)
function FadeInWhenVisible({ children }: { children: React.ReactNode }) {
  const ref = useRef(null);
  const isInView = useInView(ref, { once: true, margin: '-100px' });

  return (
    <motion.div
      ref={ref}
      initial={{ opacity: 0, y: 50 }}
      animate={isInView ? { opacity: 1, y: 0 } : { opacity: 0, y: 50 }}
      transition={{ duration: 0.5 }}
    >
      {children}
    </motion.div>
  );
}

// Scroll-linked element animation
function ScrollScaleCard() {
  const ref = useRef(null);
  const { scrollYProgress } = useScroll({
    target: ref,
    offset: ['start end', 'end start'], // when element enters and exits
  });

  const scale = useTransform(scrollYProgress, [0, 0.5, 1], [0.8, 1, 0.8]);
  const opacity = useTransform(scrollYProgress, [0, 0.3, 0.7, 1], [0, 1, 1, 0]);

  return (
    <motion.div ref={ref} style={{ scale, opacity }}>
      Content
    </motion.div>
  );
}
```

## Spring Animations

```tsx
// Spring transition (most natural feel)
<motion.div
  animate={{ x: 100 }}
  transition={{
    type: 'spring',
    stiffness: 300,  // how snappy
    damping: 30,     // how much friction
    mass: 1,         // how heavy
  }}
/>

// Spring presets
const springs = {
  snappy: { type: 'spring', stiffness: 400, damping: 30 },
  gentle: { type: 'spring', stiffness: 120, damping: 14 },
  bouncy: { type: 'spring', stiffness: 300, damping: 10 },
  stiff:  { type: 'spring', stiffness: 500, damping: 35 },
};

// useSpring for smooth scroll-linked values
const { scrollYProgress } = useScroll();
const smoothProgress = useSpring(scrollYProgress, {
  stiffness: 100,
  damping: 30,
  restDelta: 0.001,
});
```

## SVG Animations

```tsx
// Path drawing
<motion.svg viewBox="0 0 100 100" className="w-24 h-24">
  <motion.circle
    cx="50"
    cy="50"
    r="40"
    stroke="currentColor"
    strokeWidth="3"
    fill="none"
    initial={{ pathLength: 0 }}
    animate={{ pathLength: 1 }}
    transition={{ duration: 1.5, ease: 'easeInOut' }}
  />
</motion.svg>

// Animated checkmark
<motion.svg viewBox="0 0 24 24" className="w-6 h-6 text-green-500">
  <motion.path
    d="M5 13l4 4L19 7"
    fill="none"
    stroke="currentColor"
    strokeWidth="2"
    strokeLinecap="round"
    initial={{ pathLength: 0, opacity: 0 }}
    animate={{ pathLength: 1, opacity: 1 }}
    transition={{ duration: 0.4, delay: 0.2 }}
  />
</motion.svg>
```

## Reduced Motion Support

```tsx
import { useReducedMotion } from 'framer-motion';

function AnimatedCard({ children }: { children: React.ReactNode }) {
  const shouldReduceMotion = useReducedMotion();

  return (
    <motion.div
      initial={{ opacity: 0, y: shouldReduceMotion ? 0 : 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{
        duration: shouldReduceMotion ? 0 : 0.3,
      }}
    >
      {children}
    </motion.div>
  );
}

// Or use a global wrapper
const reducedMotionVariants = {
  hidden: { opacity: 0 },
  visible: { opacity: 1 },
};

const fullMotionVariants = {
  hidden: { opacity: 0, y: 20 },
  visible: { opacity: 1, y: 0 },
};
```

## Common Patterns

### Modal

```tsx
function Modal({ isOpen, onClose, children }) {
  return (
    <AnimatePresence>
      {isOpen && (
        <>
          <motion.div
            className="fixed inset-0 bg-black/50 z-40"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            onClick={onClose}
          />
          <motion.div
            className="fixed inset-x-4 top-[20%] max-w-lg mx-auto bg-white rounded-xl p-6 z-50"
            initial={{ opacity: 0, scale: 0.95, y: 10 }}
            animate={{ opacity: 1, scale: 1, y: 0 }}
            exit={{ opacity: 0, scale: 0.95, y: 10 }}
            transition={{ type: 'spring', stiffness: 300, damping: 30 }}
          >
            {children}
          </motion.div>
        </>
      )}
    </AnimatePresence>
  );
}
```

### Accordion

```tsx
function Accordion({ title, children }) {
  const [isOpen, setIsOpen] = useState(false);

  return (
    <div className="border rounded-lg">
      <button onClick={() => setIsOpen(!isOpen)} className="w-full p-4 flex justify-between">
        {title}
        <motion.span animate={{ rotate: isOpen ? 180 : 0 }}>▼</motion.span>
      </button>
      <AnimatePresence>
        {isOpen && (
          <motion.div
            initial={{ height: 0, opacity: 0 }}
            animate={{ height: 'auto', opacity: 1 }}
            exit={{ height: 0, opacity: 0 }}
            transition={{ duration: 0.2 }}
            className="overflow-hidden"
          >
            <div className="p-4 pt-0">{children}</div>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
}
```

### Toast Notifications

```tsx
function Toasts({ toasts, removeToast }) {
  return (
    <div className="fixed bottom-4 right-4 flex flex-col gap-2 z-50">
      <AnimatePresence>
        {toasts.map((toast) => (
          <motion.div
            key={toast.id}
            layout
            initial={{ opacity: 0, x: 50, scale: 0.95 }}
            animate={{ opacity: 1, x: 0, scale: 1 }}
            exit={{ opacity: 0, x: 50, scale: 0.95 }}
            transition={{ type: 'spring', stiffness: 400, damping: 30 }}
            className="bg-white shadow-lg rounded-lg p-4 min-w-[300px]"
          >
            {toast.message}
          </motion.div>
        ))}
      </AnimatePresence>
    </div>
  );
}
```

## Gotchas

1. **`AnimatePresence` needs direct children with `key`.** The animated elements must be direct children of AnimatePresence, each with a unique key. Don't wrap them in an extra div.

2. **`layout` causes flashes on first render.** If elements flash or jump on mount, add `initial={false}` to skip the initial animation, or use `layoutId` for shared layout transitions.

3. **`height: auto` doesn't animate smoothly.** Use `height: "auto"` with `overflow: hidden` on the container. Framer Motion measures the auto height and animates to it.

4. **Spring animations have no fixed duration.** Springs animate until they settle. Use `stiffness` and `damping` to control timing. If you need exact duration, use `type: "tween"` instead.

5. **`useScroll` returns 0-1 progress values.** Use `useTransform` to map progress to actual pixel values or other ranges.

6. **Large lists with layout animations are expensive.** The `layout` prop measures every sibling on every frame. For lists with 50+ items, limit layout animations to visible items only.
