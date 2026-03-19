# Core Web Vitals — Complete Reference

## The Three Metrics

### LCP — Largest Contentful Paint

```
What: Time for the largest visible element to render
Target: < 2.5 seconds
Measured: From navigation start to largest element painted

What counts as LCP element:
  - <img> elements
  - <image> inside <svg>
  - <video> poster image
  - Background image via url() (CSS)
  - Block-level text elements (<h1>, <p>, etc.)

What does NOT count:
  - Elements with opacity: 0
  - Elements covering the full viewport (background)
  - Placeholder/skeleton content that gets replaced
```

#### LCP Optimization Strategies

```
1. OPTIMIZE THE LCP ELEMENT
   ─────────────────────────
   If LCP is an image:
     □ Use WebP or AVIF format (30-50% smaller)
     □ Set explicit width and height
     □ Use srcset for responsive sizes
     □ Preload: <link rel="preload" as="image" href="hero.webp">
     □ Use fetchpriority="high" on the <img>
     □ Do NOT lazy-load the LCP image
     □ Serve from CDN (close to user)
     □ Use correct dimensions (don't serve 4000px image in 800px container)

   If LCP is text:
     □ Inline critical CSS (above-fold styles in <style> tag)
     □ Preload web fonts
     □ Use font-display: swap or optional
     □ Consider system fonts for above-fold text

2. REDUCE TIME TO FIRST BYTE (TTFB)
   ──────────────────────────────────
   Target: < 800ms

   □ Use a CDN (Cloudflare, Vercel, AWS CloudFront)
   □ Enable server-side caching (Redis, in-memory)
   □ Optimize database queries
   □ Use HTTP/2 or HTTP/3
   □ Enable compression (Brotli > gzip)
   □ Consider static generation (SSG) over SSR

3. ELIMINATE RENDER-BLOCKING RESOURCES
   ────────────────────────────────────
   □ Inline critical CSS (< 14KB)
   □ Defer non-critical CSS:
     <link rel="preload" href="styles.css" as="style"
           onload="this.onload=null;this.rel='stylesheet'">
   □ async or defer on <script> tags
   □ Move non-critical JS to end of <body>
   □ Code-split: only load JS needed for current page

4. PRECONNECT TO CRITICAL ORIGINS
   ───────────────────────────────
   <link rel="preconnect" href="https://fonts.googleapis.com">
   <link rel="preconnect" href="https://cdn.example.com" crossorigin>
   <link rel="dns-prefetch" href="https://www.googletagmanager.com">
```

### INP — Interaction to Next Paint

```
What: Responsiveness to user interactions (click, tap, key press)
Target: < 200ms
Measured: Time from interaction to next visual update

How it works:
  1. User clicks/taps/presses key
  2. Browser processes event handlers
  3. Browser recalculates styles and layout
  4. Browser paints the update
  Total time = INP

Rating:
  < 200ms  = Good (green)
  200-500ms = Needs Improvement (yellow)
  > 500ms  = Poor (red)
```

#### INP Optimization Strategies

```
1. BREAK UP LONG TASKS
   ────────────────────
   Problem: JavaScript tasks > 50ms block the main thread.

   // Bad: one long synchronous operation
   function processData(items) {
     items.forEach(item => heavyComputation(item)); // 200ms+
   }

   // Good: yield to main thread between chunks
   async function processData(items) {
     const CHUNK_SIZE = 50;
     for (let i = 0; i < items.length; i += CHUNK_SIZE) {
       const chunk = items.slice(i, i + CHUNK_SIZE);
       chunk.forEach(item => heavyComputation(item));
       // Yield to main thread
       await new Promise(resolve => setTimeout(resolve, 0));
     }
   }

   // Better: use scheduler.yield() (Chrome 129+)
   async function processData(items) {
     for (const item of items) {
       heavyComputation(item);
       if (navigator.scheduling?.isInputPending?.()) {
         await scheduler.yield();
       }
     }
   }

2. OPTIMIZE EVENT HANDLERS
   ────────────────────────
   □ Debounce scroll/resize/input handlers
   □ Use passive event listeners for scroll/touch:
     element.addEventListener('scroll', handler, { passive: true });
   □ Avoid layout thrashing (reading then writing DOM in a loop)
   □ Use requestAnimationFrame for visual updates
   □ Remove unused event listeners

3. MINIMIZE DOM SIZE
   ──────────────────
   Target: < 1,500 DOM elements
   □ Virtualize long lists (react-window, @tanstack/virtual)
   □ Lazy render below-fold content
   □ Remove hidden/unused DOM nodes
   □ Flatten nested DOM structures

4. REDUCE LAYOUT/STYLE RECALCULATION
   ──────────────────────────────────
   □ Avoid forced synchronous layouts:
     // Bad (layout thrashing)
     elements.forEach(el => {
       const height = el.offsetHeight; // forces layout
       el.style.height = height + 10 + 'px'; // triggers layout again
     });

     // Good (batch reads, then writes)
     const heights = elements.map(el => el.offsetHeight);
     elements.forEach((el, i) => {
       el.style.height = heights[i] + 10 + 'px';
     });

   □ Use CSS containment: contain: content;
   □ Use will-change for animated elements (sparingly)
   □ Avoid complex CSS selectors in frequently-updated areas

5. WEB WORKERS FOR HEAVY COMPUTATION
   ──────────────────────────────────
   // main.js
   const worker = new Worker('/worker.js');
   worker.postMessage({ data: largeDataset });
   worker.onmessage = (e) => updateUI(e.data);

   // worker.js
   self.onmessage = (e) => {
     const result = heavyComputation(e.data);
     self.postMessage(result);
   };

6. REACT-SPECIFIC
   ────────────────
   □ useTransition for non-urgent updates:
     const [isPending, startTransition] = useTransition();
     startTransition(() => setSearchResults(results));

   □ useDeferredValue for expensive renders:
     const deferredQuery = useDeferredValue(query);

   □ React.memo for components that re-render unnecessarily
   □ useCallback/useMemo to prevent needless recalculations
   □ Code-split with React.lazy + Suspense
```

### CLS — Cumulative Layout Shift

```
What: Visual stability — how much the page layout shifts unexpectedly
Target: < 0.1
Measured: Sum of all unexpected layout shift scores during page life

How shift score is calculated:
  Impact fraction × Distance fraction = Shift score
  (% of viewport affected) × (distance elements moved / viewport height)

What causes a layout shift:
  - Images/iframes without dimensions
  - Dynamically injected content
  - Web fonts (FOUT/FOIT)
  - Ads, embeds, iframes resizing
  - DOM updates above existing content
  - Late-loading CSS

What does NOT cause a shift:
  - Shifts within 500ms of a user interaction
  - Scroll-triggered animations
  - CSS transforms (translate, scale, etc.)
```

#### CLS Optimization Strategies

```
1. SET DIMENSIONS ON ALL MEDIA
   ────────────────────────────
   <!-- Always include width and height -->
   <img src="photo.jpg" width="800" height="600" alt="..." />
   <video width="640" height="360" poster="thumb.jpg"></video>
   <iframe width="560" height="315" src="..."></iframe>

   /* Responsive images with aspect ratio */
   img {
     max-width: 100%;
     height: auto;
     aspect-ratio: 4 / 3;
   }

   /* For containers with dynamic content */
   .video-wrapper {
     aspect-ratio: 16 / 9;
     width: 100%;
   }

2. RESERVE SPACE FOR DYNAMIC CONTENT
   ──────────────────────────────────
   /* Ad slot with fixed dimensions */
   .ad-slot {
     min-height: 250px;
     min-width: 300px;
     background: #f0f0f0; /* Placeholder color */
   }

   /* Skeleton screen for async content */
   .skeleton {
     min-height: 200px;
     background: linear-gradient(90deg, #f0f0f0 25%, #e0e0e0 50%, #f0f0f0 75%);
     background-size: 200% 100%;
     animation: shimmer 1.5s infinite;
   }

3. HANDLE WEB FONTS PROPERLY
   ──────────────────────────
   /* Prevent FOUT from causing shift */
   @font-face {
     font-family: 'CustomFont';
     src: url('/fonts/custom.woff2') format('woff2');
     font-display: swap; /* or 'optional' for less shift */
     /* Match fallback font metrics */
     size-adjust: 105%;
     ascent-override: 90%;
     descent-override: 22%;
     line-gap-override: 0%;
   }

   /* Preload critical fonts */
   <link rel="preload" href="/fonts/custom.woff2"
         as="font" type="font/woff2" crossorigin>

4. POSITION DYNAMIC UI CORRECTLY
   ──────────────────────────────
   /* Cookie banner: fixed, not flow */
   .cookie-banner {
     position: fixed;
     bottom: 0;
     left: 0;
     right: 0;
     z-index: 9999;
   }

   /* Toast notifications: fixed */
   .toast {
     position: fixed;
     top: 20px;
     right: 20px;
   }

   /* NEVER insert content above existing content */
   /* If you must, use CSS transforms instead */
   .notification-enter {
     transform: translateY(-100%);
     animation: slideDown 0.3s forwards;
   }

5. USE CSS TRANSFORMS FOR ANIMATIONS
   ──────────────────────────────────
   /* Transforms don't cause layout shifts */
   .card:hover {
     transform: scale(1.05);     /* OK */
     transform: translateY(-4px); /* OK */
   }

   /* These DO cause shifts */
   .card:hover {
     margin-top: -4px;  /* BAD - causes reflow */
     width: 110%;       /* BAD - causes reflow */
     top: -4px;         /* BAD if position: relative */
   }
```

---

## Measuring Core Web Vitals

### Field Data (Real Users)

| Source | Data Type | Access |
|--------|-----------|--------|
| Chrome UX Report (CrUX) | 28-day rolling average | PageSpeed Insights, Search Console |
| web-vitals library | Per-session metrics | Self-hosted analytics |
| GA4 (auto-collected) | Per-session via web-vitals | GA4 dashboard |

### Lab Data (Synthetic)

| Tool | LCP | INP | CLS |
|------|-----|-----|-----|
| Lighthouse | Yes | No (uses TBT as proxy) | Yes |
| Chrome DevTools (Performance) | Yes | Yes (via interactions) | Yes |
| WebPageTest | Yes | Partial | Yes |
| PageSpeed Insights | Yes | Yes (field data) | Yes |

### web-vitals Library

```typescript
// npm install web-vitals
import { onLCP, onINP, onCLS, onFCP, onTTFB } from 'web-vitals';

function sendToAnalytics(metric) {
  // Send to your analytics endpoint
  const body = JSON.stringify({
    name: metric.name,
    value: metric.value,
    rating: metric.rating,  // 'good', 'needs-improvement', 'poor'
    delta: metric.delta,
    id: metric.id,
    navigationType: metric.navigationType,
    url: window.location.href,
  });

  // Use sendBeacon for reliability
  if (navigator.sendBeacon) {
    navigator.sendBeacon('/api/vitals', body);
  } else {
    fetch('/api/vitals', { body, method: 'POST', keepalive: true });
  }
}

onLCP(sendToAnalytics);
onINP(sendToAnalytics);
onCLS(sendToAnalytics);
onFCP(sendToAnalytics);
onTTFB(sendToAnalytics);
```

---

## Quick Reference Card

```
┌─────────────────────────────────────────────────────┐
│              CORE WEB VITALS TARGETS                │
├─────────┬──────────┬───────────────┬────────────────┤
│ Metric  │ Good     │ Needs Improv. │ Poor           │
├─────────┼──────────┼───────────────┼────────────────┤
│ LCP     │ < 2.5s   │ 2.5s - 4.0s   │ > 4.0s         │
│ INP     │ < 200ms  │ 200ms - 500ms │ > 500ms        │
│ CLS     │ < 0.1    │ 0.1 - 0.25    │ > 0.25         │
├─────────┼──────────┼───────────────┼────────────────┤
│ FCP     │ < 1.8s   │ 1.8s - 3.0s   │ > 3.0s         │
│ TTFB    │ < 800ms  │ 800ms - 1.8s  │ > 1.8s         │
│ TBT     │ < 200ms  │ 200ms - 600ms │ > 600ms        │
└─────────┴──────────┴───────────────┴────────────────┘

Top fixes by metric:
  LCP  → Preload hero image, inline critical CSS, use CDN
  INP  → Break long tasks, debounce handlers, virtualize lists
  CLS  → Set image dimensions, reserve space, fix font loading
```
