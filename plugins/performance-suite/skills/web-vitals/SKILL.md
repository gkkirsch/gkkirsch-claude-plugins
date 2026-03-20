---
name: web-vitals
description: >
  Core Web Vitals optimization — LCP, INP, CLS measurement and improvement,
  font loading, resource hints, critical CSS, and performance monitoring.
  Triggers: "web vitals", "core web vitals", "LCP", "INP", "CLS", "FCP", "TTFB",
  "lighthouse", "page speed", "performance score", "layout shift".
  NOT for: React-specific optimization (use react-performance), caching (use caching-patterns).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Core Web Vitals Optimization

## Metrics Overview

| Metric | Good | Needs Work | Poor | What It Measures |
|--------|------|------------|------|------------------|
| **LCP** | < 2.5s | 2.5-4s | > 4s | Largest visible element render time |
| **INP** | < 200ms | 200-500ms | > 500ms | Interaction to Next Paint (responsiveness) |
| **CLS** | < 0.1 | 0.1-0.25 | > 0.25 | Visual stability (layout shift score) |
| FCP | < 1.8s | 1.8-3s | > 3s | First content paint |
| TTFB | < 800ms | 800ms-1.8s | > 1.8s | Server response time |

## Measuring Web Vitals

```typescript
// Using web-vitals library
import { onCLS, onINP, onLCP, onFCP, onTTFB } from 'web-vitals';

function sendToAnalytics(metric: { name: string; value: number; id: string }) {
  console.log(metric);
  // Send to your analytics endpoint
  fetch('/api/vitals', {
    method: 'POST',
    body: JSON.stringify(metric),
    headers: { 'Content-Type': 'application/json' },
  });
}

onCLS(sendToAnalytics);
onINP(sendToAnalytics);
onLCP(sendToAnalytics);
onFCP(sendToAnalytics);
onTTFB(sendToAnalytics);
```

```bash
# Lighthouse CLI
npx lighthouse http://localhost:3000 --output=html --output-path=./report.html
npx lighthouse http://localhost:3000 --only-categories=performance --output=json

# Chrome DevTools
# Performance tab → Record → Interact → Stop → Analyze waterfall
# Lighthouse tab → Generate report
```

## LCP Optimization (Largest Contentful Paint)

```html
<!-- 1. Preload LCP image -->
<link rel="preload" as="image" href="/hero-image.webp" fetchpriority="high" />

<!-- 2. Use fetchpriority on LCP element -->
<img src="/hero.webp" alt="Hero" fetchpriority="high" loading="eager"
     width="1200" height="600" />

<!-- 3. Preconnect to image CDN -->
<link rel="preconnect" href="https://cdn.example.com" />
<link rel="dns-prefetch" href="https://cdn.example.com" />

<!-- 4. Inline critical CSS (above-the-fold styles) -->
<style>
  /* Critical CSS for LCP element */
  .hero { display: flex; align-items: center; min-height: 60vh; }
  .hero img { width: 100%; height: auto; }
</style>

<!-- 5. Defer non-critical CSS -->
<link rel="preload" href="/styles.css" as="style" onload="this.onload=null;this.rel='stylesheet'" />
<noscript><link rel="stylesheet" href="/styles.css" /></noscript>
```

```typescript
// Next.js: priority prop on LCP image
import Image from 'next/image';
<Image src="/hero.webp" alt="Hero" width={1200} height={600} priority />

// Avoid LCP killers:
// ✗ CSS background-image for LCP element (can't preload)
// ✗ Lazy-loaded LCP image
// ✗ Client-side rendered LCP content
// ✗ Render-blocking scripts before LCP element
```

### Common LCP Elements

```
1. <img> elements (most common)
2. <video> poster images
3. Elements with CSS background-image
4. Block-level text elements (<h1>, <p>)
5. <svg> elements
```

## INP Optimization (Interaction to Next Paint)

```typescript
// 1. Break up long tasks (> 50ms)
// BEFORE: blocks main thread
function processLargeList(items: Item[]) {
  items.forEach(item => heavyComputation(item));
}

// AFTER: yield to main thread between chunks
async function processLargeList(items: Item[]) {
  const CHUNK_SIZE = 50;
  for (let i = 0; i < items.length; i += CHUNK_SIZE) {
    const chunk = items.slice(i, i + CHUNK_SIZE);
    chunk.forEach(item => heavyComputation(item));
    await yieldToMain(); // Let browser handle interactions
  }
}

function yieldToMain(): Promise<void> {
  return new Promise(resolve => {
    if ('scheduler' in globalThis && 'yield' in (globalThis as any).scheduler) {
      (globalThis as any).scheduler.yield().then(resolve);
    } else {
      setTimeout(resolve, 0);
    }
  });
}

// 2. Use startTransition for non-urgent updates
import { startTransition } from 'react';

function SearchPage() {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState([]);

  function handleInput(e: ChangeEvent<HTMLInputElement>) {
    setQuery(e.target.value); // Urgent: update input immediately

    startTransition(() => {
      setResults(filterResults(e.target.value)); // Non-urgent: can be deferred
    });
  }
}

// 3. Debounce expensive handlers
function SearchInput() {
  const debouncedSearch = useMemo(
    () => debounce((query: string) => {
      startTransition(() => setResults(search(query)));
    }, 300),
    [],
  );

  return <input onChange={e => debouncedSearch(e.target.value)} />;
}

// 4. Move heavy work off main thread
const worker = new Worker(new URL('./heavy-worker.ts', import.meta.url));

function processData(data: Data) {
  return new Promise<Result>(resolve => {
    worker.postMessage(data);
    worker.onmessage = (e) => resolve(e.data);
  });
}
```

## CLS Optimization (Cumulative Layout Shift)

```html
<!-- 1. Always set dimensions on images/video -->
<img src="photo.jpg" width="800" height="600" alt="Photo" />
<video width="640" height="360" />
<iframe width="560" height="315" />

<!-- 2. Use aspect-ratio for responsive media -->
<div class="aspect-video">
  <img src="photo.jpg" class="w-full h-full object-cover" alt="" />
</div>

<!-- 3. Reserve space for dynamic content -->
<div class="min-h-[250px]"><!-- Ad slot with reserved height --></div>
<div class="min-h-[100px]"><!-- Dynamic banner --></div>

<!-- 4. Use CSS containment -->
<div style="contain: layout;">
  <!-- Content that might change size -->
</div>
```

```css
/* 5. Font loading without layout shift */
@font-face {
  font-family: 'Inter';
  src: url('/fonts/Inter.woff2') format('woff2');
  font-display: swap;  /* Show fallback immediately, swap when loaded */
  /* Or: optional — only use if cached (no swap flash) */
  /* font-display: optional; */
}

/* Size-adjust fallback to match custom font metrics */
@font-face {
  font-family: 'Inter Fallback';
  src: local('Arial');
  size-adjust: 107%;        /* Match Inter's character width */
  ascent-override: 90%;
  descent-override: 22%;
  line-gap-override: 0%;
}

body { font-family: 'Inter', 'Inter Fallback', sans-serif; }
```

```typescript
// 6. Avoid inserting content above existing content
// BAD: prepending a banner pushes everything down
document.body.prepend(banner);

// GOOD: use fixed/sticky positioning or reserve space
<div class="sticky top-0 z-50">{banner}</div>

// 7. Use transform for animations instead of top/left/width/height
// BAD: animating layout properties causes shifts
element.style.top = `${newTop}px`;

// GOOD: transforms don't cause layout shift
element.style.transform = `translateY(${offset}px)`;
```

## Resource Hints

```html
<head>
  <!-- DNS prefetch: resolve DNS early for third-party domains -->
  <link rel="dns-prefetch" href="https://cdn.example.com" />
  <link rel="dns-prefetch" href="https://fonts.googleapis.com" />

  <!-- Preconnect: DNS + TCP + TLS handshake (use for critical third-party) -->
  <link rel="preconnect" href="https://cdn.example.com" />
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin />

  <!-- Preload: fetch critical resources early -->
  <link rel="preload" href="/fonts/Inter.woff2" as="font" type="font/woff2" crossorigin />
  <link rel="preload" href="/hero.webp" as="image" fetchpriority="high" />
  <link rel="preload" href="/critical.css" as="style" />

  <!-- Prefetch: fetch resources for NEXT navigation (low priority) -->
  <link rel="prefetch" href="/dashboard.js" />
  <link rel="prefetch" href="/api/user" />

  <!-- Modulepreload: preload ES modules -->
  <link rel="modulepreload" href="/src/utils.js" />
</head>
```

## Script Loading

```html
<!-- Render-blocking (default — avoid for non-critical scripts) -->
<script src="critical.js"></script>

<!-- Defer: download during parse, execute after DOM ready (in order) -->
<script defer src="app.js"></script>
<script defer src="analytics.js"></script>

<!-- Async: download during parse, execute immediately (no order guarantee) -->
<script async src="analytics.js"></script>

<!-- Dynamic import (most flexible — load on demand) -->
<script>
  button.addEventListener('click', async () => {
    const { showModal } = await import('./modal.js');
    showModal();
  });
</script>
```

## Performance Budget

```json
// bundlesize config in package.json
{
  "bundlesize": [
    { "path": "dist/assets/index-*.js", "maxSize": "150kB" },
    { "path": "dist/assets/vendor-*.js", "maxSize": "100kB" },
    { "path": "dist/assets/index-*.css", "maxSize": "30kB" },
    { "path": "dist/assets/*.jpg", "maxSize": "100kB" }
  ]
}
```

```typescript
// Lighthouse CI (in CI/CD)
// lighthouserc.js
module.exports = {
  ci: {
    collect: { url: ['http://localhost:3000'] },
    assert: {
      assertions: {
        'categories:performance': ['error', { minScore: 0.9 }],
        'first-contentful-paint': ['error', { maxNumericValue: 1800 }],
        'largest-contentful-paint': ['error', { maxNumericValue: 2500 }],
        'cumulative-layout-shift': ['error', { maxNumericValue: 0.1 }],
        'interactive': ['error', { maxNumericValue: 3800 }],
      },
    },
  },
};
```

## Gotchas

1. **Lighthouse lab scores != real user experience.** Lab tests run on simulated conditions. Use Real User Monitoring (RUM) with `web-vitals` library for actual field data. PageSpeed Insights shows both.

2. **`fetchpriority="high"` only works on the LCP element.** Using it on multiple images defeats the purpose. Only the single LCP element should get high priority. Other above-fold images: `loading="eager"` is sufficient.

3. **`font-display: swap` can cause CLS.** The font swap from fallback to custom font shifts text. Use `size-adjust` on the fallback font-face or `font-display: optional` (only uses custom font if already cached).

4. **Preloading too many resources hurts.** Each preload competes for bandwidth. Limit to 3-5 critical resources: LCP image, primary font, critical CSS. More than that and you're slowing down what matters.

5. **Third-party scripts are the #1 performance killer.** Analytics, chat widgets, ad scripts, A/B testing — each adds 50-200ms. Audit with `performance.getEntriesByType('resource')`. Load non-essential third parties with `defer` or after `load` event.

6. **CLS is measured for the entire page lifetime.** A layout shift 30 seconds after load still counts. Infinite scroll, lazy-loaded images without dimensions, and dynamic content injection all contribute to CLS long after initial paint.
