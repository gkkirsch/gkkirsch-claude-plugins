---
name: page-transitions
description: >
  Page transition patterns — route transitions with Framer Motion,
  View Transitions API, shared element transitions, and loading states.
  Triggers: "page transition", "route animation", "view transition",
  "shared element", "route change animation", "navigation animation".
  NOT for: component-level animation (use framer-motion or css-animations).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Page Transitions

## React Router + Framer Motion

### Basic Fade Transition

```tsx
import { AnimatePresence, motion } from 'framer-motion';
import { useLocation, Routes, Route } from 'react-router-dom';

function AnimatedRoutes() {
  const location = useLocation();

  return (
    <AnimatePresence mode="wait">
      <Routes location={location} key={location.pathname}>
        <Route path="/" element={<PageWrapper><Home /></PageWrapper>} />
        <Route path="/about" element={<PageWrapper><About /></PageWrapper>} />
        <Route path="/contact" element={<PageWrapper><Contact /></PageWrapper>} />
      </Routes>
    </AnimatePresence>
  );
}

function PageWrapper({ children }: { children: React.ReactNode }) {
  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      exit={{ opacity: 0 }}
      transition={{ duration: 0.2 }}
    >
      {children}
    </motion.div>
  );
}
```

### Slide Transition

```tsx
const slideVariants = {
  initial: { x: '100%', opacity: 0 },
  animate: { x: 0, opacity: 1 },
  exit: { x: '-100%', opacity: 0 },
};

function PageWrapper({ children }: { children: React.ReactNode }) {
  return (
    <motion.div
      variants={slideVariants}
      initial="initial"
      animate="animate"
      exit="exit"
      transition={{ type: 'spring', stiffness: 300, damping: 30 }}
    >
      {children}
    </motion.div>
  );
}
```

### Direction-Aware Transitions

```tsx
// Detect navigation direction for slide direction
function useNavigationDirection() {
  const location = useLocation();
  const [direction, setDirection] = useState(1); // 1 = forward, -1 = back
  const prevPath = useRef(location.pathname);

  useEffect(() => {
    // Simple heuristic: pages later in nav order go "forward"
    const pages = ['/', '/features', '/pricing', '/contact'];
    const prevIndex = pages.indexOf(prevPath.current);
    const currIndex = pages.indexOf(location.pathname);
    setDirection(currIndex >= prevIndex ? 1 : -1);
    prevPath.current = location.pathname;
  }, [location.pathname]);

  return direction;
}

function DirectionalPageWrapper({ children }: { children: React.ReactNode }) {
  const direction = useNavigationDirection();

  return (
    <motion.div
      initial={{ x: direction * 300, opacity: 0 }}
      animate={{ x: 0, opacity: 1 }}
      exit={{ x: direction * -300, opacity: 0 }}
      transition={{ type: 'spring', stiffness: 300, damping: 30 }}
    >
      {children}
    </motion.div>
  );
}
```

## View Transitions API (Native Browser)

### Basic Usage

```typescript
// Wrap any DOM change in startViewTransition
function navigateWithTransition(url: string) {
  if (!document.startViewTransition) {
    // Fallback for unsupported browsers
    window.location.href = url;
    return;
  }

  document.startViewTransition(() => {
    // Update the DOM (React re-render, etc.)
    navigate(url); // React Router navigate
  });
}
```

```css
/* Default cross-fade (works out of the box) */
::view-transition-old(root) {
  animation: fade-out 0.2s ease-out;
}

::view-transition-new(root) {
  animation: fade-in 0.2s ease-in;
}

/* Custom slide transition */
::view-transition-old(root) {
  animation: slide-out-left 0.3s ease-out;
}

::view-transition-new(root) {
  animation: slide-in-right 0.3s ease-out;
}

@keyframes slide-out-left {
  to { transform: translateX(-100%); opacity: 0; }
}

@keyframes slide-in-right {
  from { transform: translateX(100%); opacity: 0; }
}
```

### Shared Element Transitions (View Transitions API)

```css
/* Tag elements that should animate between pages */
.product-image {
  view-transition-name: product-hero;
}

.product-title {
  view-transition-name: product-title;
}

/* Customize the shared element animation */
::view-transition-group(product-hero) {
  animation-duration: 0.4s;
  animation-timing-function: cubic-bezier(0.4, 0, 0.2, 1);
}

/* IMPORTANT: view-transition-name must be unique on the page */
/* Dynamic names for list items: */
.product-card:nth-child(1) .image { view-transition-name: product-1; }
.product-card:nth-child(2) .image { view-transition-name: product-2; }

/* Or use inline styles in React: */
```

```tsx
function ProductCard({ product }: { product: Product }) {
  return (
    <Link to={`/product/${product.id}`}>
      <img
        src={product.image}
        style={{ viewTransitionName: `product-${product.id}` }}
      />
    </Link>
  );
}

function ProductDetail({ product }: { product: Product }) {
  return (
    <img
      src={product.image}
      style={{ viewTransitionName: `product-${product.id}` }}
      className="w-full h-96 object-cover"
    />
  );
}
```

### React Hook for View Transitions

```tsx
function useViewTransition() {
  const navigate = useNavigate();

  const transitionNavigate = useCallback(
    (to: string, options?: NavigateOptions) => {
      if (!document.startViewTransition) {
        navigate(to, options);
        return;
      }

      document.startViewTransition(() => {
        navigate(to, options);
      });
    },
    [navigate],
  );

  return { navigate: transitionNavigate };
}

// Usage
function Nav() {
  const { navigate } = useViewTransition();

  return (
    <nav>
      <button onClick={() => navigate('/')}>Home</button>
      <button onClick={() => navigate('/about')}>About</button>
    </nav>
  );
}
```

## Next.js Page Transitions

### App Router with Framer Motion

```tsx
// app/template.tsx — wraps each page (remounts on navigation)
'use client';

import { motion } from 'framer-motion';

export default function Template({ children }: { children: React.ReactNode }) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 10 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.2 }}
    >
      {children}
    </motion.div>
  );
}
```

### App Router with View Transitions

```tsx
// app/layout.tsx
'use client';

import { useRouter } from 'next/navigation';
import Link from 'next/link';

// Next.js 15+ has experimental viewTransition support
// next.config.js: { experimental: { viewTransition: true } }

// Or manual implementation:
function TransitionLink({ href, children, ...props }: React.ComponentProps<typeof Link>) {
  const router = useRouter();

  const handleClick = (e: React.MouseEvent) => {
    e.preventDefault();

    if (!document.startViewTransition) {
      router.push(href.toString());
      return;
    }

    document.startViewTransition(() => {
      router.push(href.toString());
    });
  };

  return (
    <Link href={href} onClick={handleClick} {...props}>
      {children}
    </Link>
  );
}
```

## Loading States Between Pages

### Skeleton During Navigation

```tsx
import { Suspense } from 'react';
import { motion, AnimatePresence } from 'framer-motion';

function PageSkeleton() {
  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      exit={{ opacity: 0 }}
      className="space-y-4 p-6"
    >
      <div className="h-8 w-1/3 bg-gray-200 rounded animate-pulse" />
      <div className="h-4 w-full bg-gray-200 rounded animate-pulse" />
      <div className="h-4 w-2/3 bg-gray-200 rounded animate-pulse" />
      <div className="h-64 w-full bg-gray-200 rounded animate-pulse" />
    </motion.div>
  );
}

// In Next.js App Router — loading.tsx
export default function Loading() {
  return <PageSkeleton />;
}
```

### Progress Bar (NProgress Style)

```tsx
import { useEffect, useState } from 'react';
import { motion } from 'framer-motion';
import { useNavigation } from 'react-router-dom';

function NavigationProgress() {
  const navigation = useNavigation();
  const isNavigating = navigation.state !== 'idle';

  return (
    <AnimatePresence>
      {isNavigating && (
        <motion.div
          className="fixed top-0 left-0 right-0 h-0.5 bg-blue-600 z-50 origin-left"
          initial={{ scaleX: 0 }}
          animate={{ scaleX: 0.7 }}
          exit={{ scaleX: 1, opacity: 0 }}
          transition={{
            scaleX: { duration: 5, ease: 'linear' },
            opacity: { duration: 0.2, delay: 0.1 },
          }}
        />
      )}
    </AnimatePresence>
  );
}
```

## Transition Presets

```typescript
// Reusable transition configs
export const transitions = {
  fade: {
    initial: { opacity: 0 },
    animate: { opacity: 1 },
    exit: { opacity: 0 },
    transition: { duration: 0.2 },
  },
  slideUp: {
    initial: { opacity: 0, y: 20 },
    animate: { opacity: 1, y: 0 },
    exit: { opacity: 0, y: -20 },
    transition: { duration: 0.3 },
  },
  slideRight: {
    initial: { opacity: 0, x: -20 },
    animate: { opacity: 1, x: 0 },
    exit: { opacity: 0, x: 20 },
    transition: { duration: 0.3 },
  },
  scale: {
    initial: { opacity: 0, scale: 0.95 },
    animate: { opacity: 1, scale: 1 },
    exit: { opacity: 0, scale: 0.95 },
    transition: { type: 'spring', stiffness: 300, damping: 30 },
  },
  none: {
    initial: {},
    animate: {},
    exit: {},
    transition: { duration: 0 },
  },
};

// Usage
function Page() {
  return <motion.div {...transitions.slideUp}>Content</motion.div>;
}
```

## Gotchas

1. **`AnimatePresence mode="wait"` delays new page.** The new page doesn't render until the old page's exit animation completes. Keep exit durations short (100-200ms) or use `mode="popLayout"` for overlap.

2. **`useLocation` must come from the same Router.** If AnimatePresence wraps Routes, the `location` key must come from `useLocation()` inside the same Router context.

3. **View Transitions API browser support.** Chrome 111+, Edge 111+, Safari 18+, Firefox 126+. Always include a fallback (just navigate without transition). Check with `document.startViewTransition`.

4. **`view-transition-name` must be unique.** Two elements with the same `view-transition-name` on the same page causes both to disappear. Use dynamic names for lists (`product-${id}`).

5. **Next.js App Router `template.tsx` vs `layout.tsx`.** Use `template.tsx` for page transitions — it remounts on navigation. `layout.tsx` persists across navigations and won't trigger mount animations.

6. **Scroll position with page transitions.** Browsers may restore scroll position before the enter animation starts. Use `window.scrollTo(0, 0)` in the page component's effect, or configure React Router's `scrollRestoration`.
