---
name: web-vitals-optimization
description: >
  Core Web Vitals optimization — LCP, INP, CLS fixes, image optimization,
  font loading, critical CSS, preloading, and performance measurement.
  Triggers: "web vitals", "lcp optimization", "cls fix", "inp optimization",
  "page speed", "core web vitals", "lighthouse score".
  NOT for: Bundle size optimization (use bundle-optimization).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Core Web Vitals Optimization

## The Three Metrics

| Metric | What It Measures | Good | Needs Work | Poor |
|--------|-----------------|------|-----------|------|
| **LCP** (Largest Contentful Paint) | Loading speed — when largest element renders | < 2.5s | 2.5-4.0s | > 4.0s |
| **INP** (Interaction to Next Paint) | Responsiveness — input delay to visual update | < 200ms | 200-500ms | > 500ms |
| **CLS** (Cumulative Layout Shift) | Visual stability — unexpected layout movement | < 0.1 | 0.1-0.25 | > 0.25 |

## LCP Optimization

### Identify the LCP Element

```javascript
// Find what your LCP element is
new PerformanceObserver((list) => {
  const entries = list.getEntries();
  const lastEntry = entries[entries.length - 1];
  console.log("LCP element:", lastEntry.element);
  console.log("LCP time:", lastEntry.startTime);
}).observe({ type: "largest-contentful-paint", buffered: true });
```

Common LCP elements: hero images, large text blocks, video poster images, background images.

### Fix Slow LCP

```html
<!-- 1. Preload the LCP image -->
<link rel="preload" as="image" href="/hero.webp" fetchpriority="high" />

<!-- 2. Set fetchpriority on the LCP image -->
<img src="/hero.webp" alt="Hero" fetchpriority="high" loading="eager"
     width="1200" height="600" />

<!-- 3. Preconnect to external origins -->
<link rel="preconnect" href="https://cdn.example.com" />
<link rel="dns-prefetch" href="https://cdn.example.com" />
```

```typescript
// Next.js: priority prop on LCP image
import Image from "next/image";

<Image src="/hero.webp" alt="Hero" width={1200} height={600}
       priority  // Sets fetchpriority="high" + disables lazy loading
       sizes="100vw" />
```

### Critical CSS

```html
<!-- Inline critical CSS in <head> -->
<style>
  /* Only styles needed for above-the-fold content */
  .hero { display: flex; min-height: 60vh; }
  .hero img { width: 100%; object-fit: cover; }
  .nav { display: flex; gap: 1rem; padding: 1rem; }
</style>

<!-- Defer non-critical CSS -->
<link rel="preload" href="/styles.css" as="style" onload="this.onload=null;this.rel='stylesheet'" />
<noscript><link rel="stylesheet" href="/styles.css" /></noscript>
```

### Font Loading

```css
/* Use font-display: swap for text visibility during load */
@font-face {
  font-family: "Inter";
  src: url("/fonts/inter.woff2") format("woff2");
  font-display: swap;      /* Show fallback immediately, swap when loaded */
  font-weight: 400;
  font-style: normal;
}

/* Reduce layout shift with size-adjust on fallback */
@font-face {
  font-family: "Inter Fallback";
  src: local("Arial");
  size-adjust: 107%;        /* Match metrics of web font */
  ascent-override: 90%;
  descent-override: 22%;
  line-gap-override: 0%;
}

body { font-family: "Inter", "Inter Fallback", sans-serif; }
```

```html
<!-- Preload the font file -->
<link rel="preload" href="/fonts/inter.woff2" as="font" type="font/woff2" crossorigin />
```

## INP Optimization

### Break Up Long Tasks

```typescript
// BAD: One long task blocking the main thread
function processLargeList(items: Item[]) {
  items.forEach((item) => {
    heavyComputation(item); // 500ms total blocks all interactions
  });
}

// GOOD: Yield to the browser between chunks
async function processLargeList(items: Item[]) {
  const CHUNK_SIZE = 50;
  for (let i = 0; i < items.length; i += CHUNK_SIZE) {
    const chunk = items.slice(i, i + CHUNK_SIZE);
    chunk.forEach((item) => heavyComputation(item));

    // Yield to browser — allows paint and input handling
    await new Promise((resolve) => setTimeout(resolve, 0));
  }
}

// BEST: Use scheduler.yield() (modern browsers)
async function processLargeList(items: Item[]) {
  for (const item of items) {
    heavyComputation(item);
    if (navigator.scheduling?.isInputPending?.()) {
      await scheduler.yield();
    }
  }
}
```

### Optimize Event Handlers

```typescript
// BAD: Heavy synchronous work in click handler
button.addEventListener("click", () => {
  const result = expensiveCalculation();  // Blocks paint update
  updateDOM(result);
});

// GOOD: Defer non-critical work
button.addEventListener("click", () => {
  // Immediate visual feedback
  button.classList.add("active");

  // Defer heavy work
  requestAnimationFrame(() => {
    const result = expensiveCalculation();
    updateDOM(result);
  });
});

// React: Use useTransition for non-urgent updates
function SearchResults() {
  const [query, setQuery] = useState("");
  const [results, setResults] = useState<Item[]>([]);
  const [isPending, startTransition] = useTransition();

  function handleSearch(value: string) {
    setQuery(value); // Urgent: update input immediately

    startTransition(() => {
      setResults(filterItems(value)); // Non-urgent: can be interrupted
    });
  }

  return (
    <>
      <input value={query} onChange={(e) => handleSearch(e.target.value)} />
      {isPending ? <Spinner /> : <ResultList items={results} />}
    </>
  );
}
```

### Debounce Input Handlers

```typescript
// Debounce expensive operations (search, filter, resize)
function useDebouncedValue<T>(value: T, delay = 300): T {
  const [debounced, setDebounced] = useState(value);

  useEffect(() => {
    const timer = setTimeout(() => setDebounced(value), delay);
    return () => clearTimeout(timer);
  }, [value, delay]);

  return debounced;
}

// Usage
function Search() {
  const [query, setQuery] = useState("");
  const debouncedQuery = useDebouncedValue(query, 300);

  // Only fetch when debounced value changes
  const { data } = useQuery({
    queryKey: ["search", debouncedQuery],
    queryFn: () => searchApi(debouncedQuery),
    enabled: debouncedQuery.length > 2,
  });
}
```

## CLS Optimization

### Always Set Dimensions

```html
<!-- BAD: No dimensions — image loads and pushes content down -->
<img src="/photo.jpg" alt="Photo" />

<!-- GOOD: Explicit dimensions reserve space -->
<img src="/photo.jpg" alt="Photo" width="800" height="600" />

<!-- GOOD: CSS aspect-ratio for responsive -->
<img src="/photo.jpg" alt="Photo" style="aspect-ratio: 4/3; width: 100%;" />
```

### Prevent Dynamic Content Shifts

```css
/* Reserve space for ads, embeds, or async content */
.ad-slot {
  min-height: 250px; /* Reserve space for standard ad */
  contain: layout;   /* Prevent layout recalculation */
}

/* Skeleton loading preserves layout */
.skeleton {
  background: linear-gradient(90deg, #f0f0f0 25%, #e0e0e0 50%, #f0f0f0 75%);
  background-size: 200% 100%;
  animation: shimmer 1.5s infinite;
  border-radius: 4px;
}

@keyframes shimmer {
  0% { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}
```

```typescript
// React: Skeleton component
function CardSkeleton() {
  return (
    <div className="card">
      <div className="skeleton" style={{ width: "100%", height: 200 }} />
      <div className="skeleton" style={{ width: "60%", height: 24, marginTop: 12 }} />
      <div className="skeleton" style={{ width: "40%", height: 16, marginTop: 8 }} />
    </div>
  );
}

function CardList() {
  const { data, isLoading } = useQuery({ queryKey: ["cards"], queryFn: fetchCards });

  if (isLoading) {
    return Array.from({ length: 6 }, (_, i) => <CardSkeleton key={i} />);
  }

  return data.map((card) => <Card key={card.id} data={card} />);
}
```

### Font-Induced CLS

```css
/* Match fallback font metrics to web font */
@font-face {
  font-family: "Custom Font Fallback";
  src: local("Arial");
  size-adjust: 105%;
  ascent-override: 95%;
}

/* Use CSS containment */
.text-container {
  contain: layout style;
  content-visibility: auto; /* Lazy render off-screen content */
}
```

## Image Optimization

```html
<!-- Modern format with fallbacks -->
<picture>
  <source srcset="/hero.avif" type="image/avif" />
  <source srcset="/hero.webp" type="image/webp" />
  <img src="/hero.jpg" alt="Hero" width="1200" height="600"
       loading="lazy" decoding="async" />
</picture>

<!-- Responsive images with srcset -->
<img srcset="/photo-400.webp 400w, /photo-800.webp 800w, /photo-1200.webp 1200w"
     sizes="(max-width: 600px) 400px, (max-width: 1024px) 800px, 1200px"
     src="/photo-800.webp" alt="Photo"
     loading="lazy" decoding="async" width="1200" height="800" />
```

| Format | Compression | Browser Support | Use For |
|--------|------------|----------------|---------|
| AVIF | Best (50% smaller than JPEG) | Chrome, Firefox, Safari 16.1+ | Photos, hero images |
| WebP | Great (25-35% smaller than JPEG) | All modern browsers | General purpose |
| JPEG | Baseline | Universal | Fallback |
| PNG | Lossless | Universal | Screenshots, icons with transparency |
| SVG | Vector | Universal | Icons, logos, illustrations |

## Measurement

```typescript
// web-vitals library
import { onLCP, onINP, onCLS, onFCP, onTTFB } from "web-vitals";

function sendToAnalytics(metric: { name: string; value: number; id: string }) {
  fetch("/api/vitals", {
    method: "POST",
    body: JSON.stringify(metric),
    keepalive: true, // Survives page unload
  });
}

onLCP(sendToAnalytics);
onINP(sendToAnalytics);
onCLS(sendToAnalytics);
onFCP(sendToAnalytics);
onTTFB(sendToAnalytics);
```

```bash
# Lighthouse CLI
npx lighthouse https://example.com --output=json --output-path=./report.json

# Chrome DevTools Performance panel
# 1. Open DevTools → Performance tab
# 2. Click Record → interact with page → Stop
# 3. Look for "Long Tasks" (red flags) in the timeline
```

## Gotchas

1. **`loading="lazy"` on LCP image kills performance** — The LCP image must load immediately. Never lazy-load it. Use `loading="eager"` (default) and `fetchpriority="high"` on the LCP element.

2. **Third-party scripts are the #1 LCP killer** — Analytics, chat widgets, and ad scripts block the main thread. Load them with `async` or `defer`, or use a facade pattern (show a static placeholder, load the real widget on interaction).

3. **CSS `@import` serializes loading** — `@import url("other.css")` inside a CSS file creates a waterfall (file A must load before file B starts). Use multiple `<link>` tags in HTML instead for parallel loading.

4. **Layout shifts from web fonts** — When a web font loads and replaces the fallback, text reflows. Use `font-display: swap` + `size-adjust` on the fallback font to minimize the shift.

5. **`content-visibility: auto` can cause CLS** — If you use `content-visibility: auto` without setting `contain-intrinsic-size`, the browser renders elements with zero height initially, then shifts when they become visible. Always pair with `contain-intrinsic-size`.

6. **INP counts ALL interactions, not just the worst** — INP is the 98th percentile interaction latency, not the single worst one. One slow handler that fires on every keypress will dominate your INP score. Focus on the most frequent interactions first.
