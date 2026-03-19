# Core Web Vitals Reference

Comprehensive reference for measuring, understanding, and optimizing Core Web Vitals — the metrics Google uses to evaluate real-world user experience on the web.

## Overview

Core Web Vitals are a set of specific metrics that Google considers important for user experience. They are part of Google's page experience signals used for search ranking.

**The three Core Web Vitals (as of 2024+):**

| Metric | Full Name | Measures | Good | Needs Improvement | Poor |
|--------|-----------|----------|------|-------------------|------|
| **LCP** | Largest Contentful Paint | Loading performance | ≤ 2.5s | 2.5s – 4.0s | > 4.0s |
| **INP** | Interaction to Next Paint | Responsiveness | ≤ 200ms | 200ms – 500ms | > 500ms |
| **CLS** | Cumulative Layout Shift | Visual stability | ≤ 0.1 | 0.1 – 0.25 | > 0.25 |

**Supporting metrics (not Core Web Vitals but important):**

| Metric | Measures | Good Target |
|--------|----------|-------------|
| FCP (First Contentful Paint) | Time to first visible content | ≤ 1.8s |
| TTFB (Time to First Byte) | Server response time | ≤ 800ms |
| TBT (Total Blocking Time) | Main thread blocking (lab metric, correlates with INP) | ≤ 200ms |
| SI (Speed Index) | How quickly content is visually populated | ≤ 3.4s |

---

## Largest Contentful Paint (LCP)

### What It Measures

LCP measures the time from when the user initiates loading the page until the largest content element is rendered within the viewport. It represents the point at which the user feels the page has loaded its main content.

### What Counts as an LCP Element

The browser considers these element types for LCP:
- `<img>` elements (including `<img>` inside `<picture>`)
- `<image>` elements inside `<svg>`
- `<video>` elements (uses the poster image or the first frame)
- Elements with a `background-image` loaded via `url()` (not CSS gradients)
- Block-level elements containing text nodes or inline-level text children

**The LCP element is the largest by visual size (viewport area), not DOM size.**

### LCP Sub-Parts

LCP can be broken down into four sub-parts:

```
Time to First Byte (TTFB)
  + Resource load delay (time from TTFB until browser starts loading LCP resource)
  + Resource load duration (time to download the LCP resource)
  + Element render delay (time from resource loaded to element painted)
  = LCP
```

#### 1. Time to First Byte (TTFB)

Time from navigation start to the first byte of the HTML response.

**Optimization:**
- Use a CDN to serve content from edge locations
- Use HTTP/2 or HTTP/3
- Reduce server processing time
- Use early hints (103 status code)
- Minimize redirects (each redirect adds a full round-trip)

#### 2. Resource Load Delay

Time between TTFB and when the browser starts fetching the LCP resource.

**Why it's delayed:**
- LCP resource is only discoverable after JavaScript executes (not in initial HTML)
- CSS must be downloaded and parsed before CSS background-images are discovered
- JavaScript renders the LCP element dynamically

**Optimization:**
- Make LCP resources discoverable from the HTML source
- Use `<link rel="preload">` for LCP images not in HTML
- Use `fetchpriority="high"` on the LCP `<img>` element
- Avoid JavaScript-rendered LCP elements
- Inline critical CSS to avoid blocking resource discovery

#### 3. Resource Load Duration

Time to download the LCP resource (image, video poster, web font for text LCP).

**Optimization:**
- Reduce LCP resource file size (compress, optimize, use modern formats)
- Use CDN for faster delivery
- Use responsive images (don't load a 2000px image on a 375px screen)
- Enable text compression (brotli/gzip) for CSS/JS that blocks rendering

#### 4. Element Render Delay

Time between resource load completion and the element being painted.

**Why it's delayed:**
- Render-blocking CSS prevents painting
- Render-blocking JavaScript prevents painting
- Large DOM or complex CSS requiring expensive layout/paint

**Optimization:**
- Minimize render-blocking resources
- Inline critical CSS
- Defer non-critical CSS and JavaScript
- Reduce DOM size and complexity

### LCP Optimization Checklist

```
□ Identify the LCP element (inspect in DevTools > Performance > Timings > LCP)
□ Ensure LCP resource is discoverable in HTML (not JS-rendered, not CSS background)
□ Add fetchpriority="high" to LCP <img>
□ Add <link rel="preload"> for non-<img> LCP resources
□ Remove lazy loading from LCP image (loading="lazy" delays LCP)
□ Optimize LCP image: WebP/AVIF, proper sizing, compression
□ Serve responsive images with srcset/sizes
□ Use CDN for LCP resource
□ TTFB < 800ms (check server, CDN, redirects)
□ Minimize render-blocking CSS (inline critical CSS)
□ Defer non-critical JavaScript
□ Preconnect to third-party origins needed for LCP
□ Avoid CSS background-image for LCP (use <img> instead)
□ Check for long task blocking render after resource loads
```

### LCP Measurement

```javascript
// Using web-vitals library
import { onLCP } from 'web-vitals';

onLCP((metric) => {
  console.log('LCP:', metric.value, 'ms');
  console.log('LCP element:', metric.entries[metric.entries.length - 1]?.element);
  console.log('Rating:', metric.rating); // 'good', 'needs-improvement', 'poor'
});

// Using PerformanceObserver directly
new PerformanceObserver((list) => {
  const entries = list.getEntries();
  const lastEntry = entries[entries.length - 1];
  console.log('LCP:', lastEntry.startTime);
  console.log('Element:', lastEntry.element);
  console.log('Size:', lastEntry.size);
  console.log('URL:', lastEntry.url); // For image/video elements
}).observe({ type: 'largest-contentful-paint', buffered: true });
```

### Common LCP Issues and Fixes

**Issue: LCP image loaded via JavaScript**
```html
<!-- BAD: React renders the <img> after JS execution -->
<!-- Browser can't discover the image from HTML alone -->

<!-- FIX: Add preload in HTML <head> -->
<link rel="preload" as="image" href="/hero.webp" fetchpriority="high">
```

**Issue: LCP image is a CSS background-image**
```css
/* BAD: Browser discovers this only after CSS is parsed */
.hero { background-image: url('/hero.jpg'); }

/* FIX: Use <img> element instead, or add preload */
```

**Issue: LCP is text but web font hasn't loaded**
```css
/* FIX: Use font-display: optional or swap */
@font-face {
  font-family: 'Heading';
  src: url('/heading.woff2') format('woff2');
  font-display: optional; /* Won't show FOIT/FOUT, uses fallback if font loads slowly */
}
```

---

## Interaction to Next Paint (INP)

### What It Measures

INP measures the latency of all click, tap, and keyboard interactions throughout the full page lifecycle, and reports the worst (longest) interaction (approximately — technically the 98th percentile to account for outliers on pages with many interactions).

INP replaced First Input Delay (FID) as a Core Web Vital in March 2024.

### How INP Is Calculated

For each interaction:
```
INP = Input Delay + Processing Time + Presentation Delay

Input Delay: Time from user action to event handler start
  (caused by main thread being busy with other work)

Processing Time: Time to execute all event handlers
  (caused by expensive synchronous work in handlers)

Presentation Delay: Time from handler completion to next paint
  (caused by rendering, layout, paint work after handlers)
```

The reported INP value is the worst interaction, unless the page has 50+ interactions, in which case the 98th percentile is used.

### Key Differences from FID

| | FID (Deprecated) | INP (Current) |
|---|---|---|
| Measures | First interaction only | All interactions |
| Reports | Input delay only | Full interaction latency |
| Good threshold | ≤ 100ms | ≤ 200ms |
| Captures processing time | No | Yes |
| Captures presentation delay | No | Yes |

### INP Optimization Checklist

```
□ Identify the slowest interactions (use DevTools Performance panel)
□ Break up long tasks (> 50ms) with yield points
□ Use requestAnimationFrame for visual updates
□ Use requestIdleCallback for non-urgent work
□ Defer non-essential work (analytics, logging) from event handlers
□ Avoid layout thrashing (read all, then write all)
□ Avoid forced synchronous layout in handlers
□ Use CSS containment for scrollable areas
□ Virtualize long lists (react-virtuoso, @tanstack/react-virtual)
□ Debounce/throttle expensive handlers (scroll, resize, input)
□ Use web workers for heavy computation
□ Minimize third-party script impact on main thread
□ Use React.startTransition for non-urgent state updates
□ Use useTransition/useDeferredValue for non-blocking updates
```

### Breaking Up Long Tasks

```javascript
// Pattern 1: yield to main thread with scheduler.yield() (Chrome 115+)
async function processItems(items) {
  for (const item of items) {
    processItem(item);
    // Yield to main thread every iteration
    if (scheduler.yield) {
      await scheduler.yield();
    }
  }
}

// Pattern 2: yield with setTimeout (universal fallback)
function yieldToMain() {
  return new Promise(resolve => setTimeout(resolve, 0));
}

async function processItems(items) {
  for (let i = 0; i < items.length; i++) {
    processItem(items[i]);
    if (i % 10 === 0) { // Yield every 10 items
      await yieldToMain();
    }
  }
}

// Pattern 3: requestAnimationFrame for visual updates
function handleClick() {
  // Immediate visual feedback
  button.classList.add('pressed');

  // Defer heavy work to next frame
  requestAnimationFrame(() => {
    const result = expensiveComputation();
    updateUI(result);
  });

  // Defer non-visual work
  requestIdleCallback(() => {
    analytics.track('click');
  });
}

// Pattern 4: isInputPending (Chrome) to check if user is waiting
async function processItems(items) {
  for (const item of items) {
    processItem(item);
    if (navigator.scheduling?.isInputPending()) {
      await yieldToMain(); // Yield immediately if user has pending input
    }
  }
}
```

### INP Measurement

```javascript
import { onINP } from 'web-vitals';

onINP((metric) => {
  console.log('INP:', metric.value, 'ms');
  console.log('Rating:', metric.rating);

  // Get the slowest interaction
  const entry = metric.entries[0];
  console.log('Interaction type:', entry.name); // 'click', 'keydown', etc.
  console.log('Target:', entry.target); // The DOM element
  console.log('Processing start:', entry.processingStart);
  console.log('Processing end:', entry.processingEnd);
  console.log('Input delay:', entry.processingStart - entry.startTime);
  console.log('Processing time:', entry.processingEnd - entry.processingStart);
  console.log('Presentation delay:', entry.startTime + entry.duration - entry.processingEnd);
});

// PerformanceObserver for all interactions
new PerformanceObserver((list) => {
  for (const entry of list.getEntries()) {
    if (entry.interactionId) {
      const inputDelay = entry.processingStart - entry.startTime;
      const processingTime = entry.processingEnd - entry.processingStart;
      const presentationDelay = entry.startTime + entry.duration - entry.processingEnd;
      console.log({
        type: entry.name,
        duration: entry.duration,
        inputDelay,
        processingTime,
        presentationDelay,
      });
    }
  }
}).observe({ type: 'event', buffered: true, durationThreshold: 16 });
```

### Common INP Issues and Fixes

**Issue: Heavy event handler blocks interaction**
```javascript
// BAD
searchInput.addEventListener('input', (e) => {
  const results = filterAndSortLargeDataset(e.target.value); // 100ms+
  renderResults(results);
});

// GOOD: Debounce search, show immediate feedback
searchInput.addEventListener('input', (e) => {
  setSearchValue(e.target.value); // Immediate visual update
  debouncedSearch(e.target.value); // Deferred expensive work
});
```

**Issue: Third-party scripts block main thread**
```html
<!-- BAD: Synchronous third-party script -->
<script src="https://cdn.example.com/heavy-widget.js"></script>

<!-- GOOD: Async + delayed -->
<script>
  // Load after user interaction or idle time
  requestIdleCallback(() => {
    const s = document.createElement('script');
    s.src = 'https://cdn.example.com/heavy-widget.js';
    s.async = true;
    document.head.appendChild(s);
  });
</script>
```

---

## Cumulative Layout Shift (CLS)

### What It Measures

CLS measures the total of all unexpected layout shifts that occur during the entire lifespan of the page. A layout shift occurs when a visible element changes its position from one rendered frame to the next.

### How CLS Is Calculated

```
Layout Shift Score = Impact Fraction × Distance Fraction

Impact Fraction: The percentage of the viewport that was impacted by the shift
Distance Fraction: The percentage of the viewport the element moved

CLS = Sum of all unexpected layout shift scores
      (grouped into "session windows" of max 5 seconds, with max 1 second gap between shifts)
      (only the largest session window is reported)
```

**Important: Expected layout shifts are excluded.** A shift is "expected" if it occurs within 500ms of a user interaction (click, tap, keypress). This means shifts caused by user-initiated actions (like clicking a button that expands content) do not count toward CLS.

### CLS Session Windows

CLS uses "session windows" to group related shifts:
- A session window starts with the first layout shift
- It ends when there's a 1-second gap without shifts, or after 5 seconds (whichever comes first)
- The page's CLS score is the maximum session window score (not the sum of all sessions)

### What Causes Layout Shifts

1. **Images/videos/iframes without dimensions** — Browser doesn't know the size until the resource loads
2. **Dynamically injected content** — Ads, banners, cookie notices that push content down
3. **Web fonts** — FOUT (Flash of Unstyled Text) causes text to reflow when the web font loads
4. **Late-loading CSS** — Styles that change layout after initial render
5. **Animations using layout properties** — top/left/width/height/margin/padding animations
6. **Dynamic content above existing content** — Notifications, banners, error messages inserted at the top

### CLS Optimization Checklist

```
□ Set explicit width and height on all <img>, <video>, and <iframe> elements
□ Use aspect-ratio CSS for responsive containers
□ Reserve space for dynamically inserted content (ads, banners)
□ Use font-display: optional or swap with size-adjust
□ Use transform animations instead of layout property animations
□ Never inject content above existing content without reserving space
□ Use content-visibility: auto with contain-intrinsic-size for off-screen content
□ Ensure web fonts don't cause significant text reflow (use size-adjust)
□ Check for late-loading CSS that changes layout
□ Check for JavaScript that modifies element dimensions after load
□ Avoid document.write() (modifies DOM unpredictably)
```

### CLS Measurement

```javascript
import { onCLS } from 'web-vitals';

onCLS((metric) => {
  console.log('CLS:', metric.value);
  console.log('Rating:', metric.rating);

  // Log all layout shift entries
  for (const entry of metric.entries) {
    console.log('Shift value:', entry.value);
    console.log('Had recent input:', entry.hadRecentInput);
    console.log('Sources:', entry.sources?.map(s => ({
      element: s.node,
      previousRect: s.previousRect,
      currentRect: s.currentRect,
    })));
  }
});

// PerformanceObserver for real-time shift monitoring
let clsValue = 0;
let clsEntries = [];

new PerformanceObserver((list) => {
  for (const entry of list.getEntries()) {
    // Only count unexpected shifts
    if (!entry.hadRecentInput) {
      clsValue += entry.value;
      clsEntries.push(entry);
    }
  }
}).observe({ type: 'layout-shift', buffered: true });
```

### Debugging CLS

```javascript
// Find which elements are shifting
new PerformanceObserver((list) => {
  for (const entry of list.getEntries()) {
    if (entry.hadRecentInput) continue;
    for (const source of entry.sources || []) {
      console.log('Shifting element:', source.node);
      console.log('  From:', source.previousRect);
      console.log('  To:', source.currentRect);
      // Highlight the element
      if (source.node) {
        source.node.style.outline = '3px solid red';
      }
    }
  }
}).observe({ type: 'layout-shift', buffered: true });
```

### Common CLS Issues and Fixes

**Issue: Images without dimensions**
```html
<!-- BAD -->
<img src="/photo.jpg" alt="Photo">

<!-- GOOD: Explicit dimensions -->
<img src="/photo.jpg" alt="Photo" width="800" height="600">

<!-- GOOD: CSS aspect-ratio -->
<style>
  .responsive-img {
    width: 100%;
    aspect-ratio: 4 / 3;
    object-fit: cover;
  }
</style>
<img src="/photo.jpg" alt="Photo" class="responsive-img">
```

**Issue: Font swap causes text reflow**
```css
/* GOOD: Adjust fallback font metrics to match web font */
@font-face {
  font-family: 'CustomFont';
  src: url('/custom.woff2') format('woff2');
  font-display: swap;
  size-adjust: 105%;
  ascent-override: 95%;
  descent-override: 22%;
  line-gap-override: 0%;
}

/* Use a matching system font as fallback */
body {
  font-family: 'CustomFont', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
}
```

**Issue: Dynamic banner inserted above content**
```jsx
// BAD: Banner pushes content down when it loads
function Page() {
  const [showBanner, setShowBanner] = useState(false);
  useEffect(() => {
    checkBanner().then(() => setShowBanner(true));
  }, []);
  return (
    <>
      {showBanner && <Banner />} {/* CLS when this appears */}
      <Content />
    </>
  );
}

// GOOD: Reserve space for the banner
function Page() {
  const [banner, setBanner] = useState(null);
  const [loading, setLoading] = useState(true);
  useEffect(() => {
    checkBanner().then(data => {
      setBanner(data);
      setLoading(false);
    });
  }, []);
  return (
    <>
      <div style={{ minHeight: loading || banner ? 60 : 0 }}>
        {banner && <Banner data={banner} />}
      </div>
      <Content />
    </>
  );
}
```

---

## Measurement Tools

### Lab vs Field Data

| | Lab Data | Field Data |
|---|---|---|
| Environment | Controlled (Lighthouse, DevTools) | Real users (CrUX, RUM) |
| Metrics | LCP, CLS, TBT, FCP, SI | LCP, INP, CLS, FCP, TTFB |
| Use for | Development, debugging, CI | Understanding real user experience |
| Tools | Lighthouse, WebPageTest, DevTools | CrUX, web-vitals, RUM providers |

**Note:** TBT is a lab metric that correlates with INP. INP can only be measured with real user interactions.

### Chrome DevTools

```
Performance panel:
1. Open DevTools (F12)
2. Go to Performance tab
3. Click Record, interact with page, click Stop
4. Timings track shows LCP, FCP
5. Layout Shifts track shows CLS
6. Interactions track shows INP candidates

Lighthouse:
1. Open DevTools > Lighthouse tab
2. Select categories (Performance, Accessibility, etc.)
3. Choose device (Mobile or Desktop)
4. Click "Analyze page load"
5. Review scores and audit results

Web Vitals overlay (Chrome extension):
- Install "Web Vitals" Chrome extension
- Shows LCP, INP, CLS values as you browse
```

### Chrome User Experience Report (CrUX)

```
CrUX provides real-user field data from Chrome users who opt in.

Access methods:
1. PageSpeed Insights (pagespeed.web.dev)
2. CrUX API
3. BigQuery (for large-scale analysis)
4. CrUX Dashboard (Looker Studio)
5. Search Console (Core Web Vitals report)

CrUX API example:
GET https://chromeuxreport.googleapis.com/v1/records:queryRecord?key=API_KEY
{
  "url": "https://example.com/",
  "formFactor": "PHONE",
  "metrics": ["largest_contentful_paint", "interaction_to_next_paint", "cumulative_layout_shift"]
}
```

### Lighthouse Scoring

Lighthouse Performance score is a weighted average of six metrics:

| Metric | Weight |
|--------|--------|
| TBT (Total Blocking Time) | 30% |
| LCP (Largest Contentful Paint) | 25% |
| CLS (Cumulative Layout Shift) | 25% |
| FCP (First Contentful Paint) | 10% |
| SI (Speed Index) | 10% |

**Score ranges:**
- 90-100: Good (green)
- 50-89: Needs Improvement (orange)
- 0-49: Poor (red)

### Real User Monitoring (RUM)

```javascript
// Complete RUM setup with web-vitals
import { onCLS, onINP, onLCP, onFCP, onTTFB } from 'web-vitals';

function sendMetric(metric) {
  // Include useful debugging info
  const body = {
    name: metric.name,
    value: metric.value,
    rating: metric.rating,
    delta: metric.delta,
    id: metric.id,
    navigationType: metric.navigationType,
    // Page info
    url: window.location.href,
    referrer: document.referrer,
    // Device info
    connection: navigator.connection?.effectiveType,
    deviceMemory: navigator.deviceMemory,
    // Attribution (for debugging)
    ...(metric.attribution && {
      attribution: JSON.stringify(metric.attribution),
    }),
  };

  // Send reliably
  if (navigator.sendBeacon) {
    navigator.sendBeacon('/api/vitals', JSON.stringify(body));
  } else {
    fetch('/api/vitals', {
      body: JSON.stringify(body),
      method: 'POST',
      keepalive: true,
    });
  }
}

// Use attribution builds for detailed debugging info
// import { onLCP } from 'web-vitals/attribution';
onCLS(sendMetric);
onINP(sendMetric);
onLCP(sendMetric);
onFCP(sendMetric);
onTTFB(sendMetric);
```

---

## Framework-Specific Core Web Vitals Tips

### Next.js

```
LCP:
- Use next/image with priority prop for LCP images
- Use App Router with loading.tsx for instant loading states
- Use server components to reduce client-side JS
- Static generation (SSG) or ISR for fastest TTFB

INP:
- Use React Server Components to reduce client-side interactivity cost
- Use useTransition for non-urgent updates
- Lazy load client components with next/dynamic

CLS:
- next/image automatically sets dimensions (prevents CLS)
- Use loading.tsx / Suspense boundaries (content streams in without shifting)
- next/font automatically handles font loading with size-adjust
```

### React (Vite)

```
LCP:
- Preload hero image in index.html: <link rel="preload" as="image" href="...">
- Use fetchpriority="high" on LCP image
- Code split routes with React.lazy

INP:
- Use useTransition for search/filter interactions
- Use useDeferredValue for expensive re-renders
- Virtualize long lists with @tanstack/react-virtual

CLS:
- Set width/height on all images
- Reserve space for async content
- Use font-display: swap in @font-face
```

### Vue/Nuxt

```
LCP:
- Use NuxtImg with preload prop
- Use Nuxt server components
- Use asyncData/useFetch for data loading

INP:
- Use shallowRef for large reactive objects
- Use defineAsyncComponent for lazy components
- Split expensive computed properties

CLS:
- NuxtImg handles image dimensions automatically
- Use v-show instead of v-if for toggled content (preserves space)
- Use font-display: swap
```

---

## Performance Budget Template

```json
{
  "coreWebVitals": {
    "LCP": { "good": 2500, "poor": 4000, "unit": "ms" },
    "INP": { "good": 200, "poor": 500, "unit": "ms" },
    "CLS": { "good": 0.1, "poor": 0.25, "unit": "score" }
  },
  "supportingMetrics": {
    "FCP": { "target": 1800, "unit": "ms" },
    "TTFB": { "target": 800, "unit": "ms" },
    "TBT": { "target": 200, "unit": "ms" },
    "SI": { "target": 3400, "unit": "ms" }
  },
  "resourceBudgets": {
    "totalJS": { "budget": 300, "unit": "KB", "compressed": true },
    "totalCSS": { "budget": 100, "unit": "KB", "compressed": true },
    "totalImages": { "budget": 500, "unit": "KB" },
    "totalFonts": { "budget": 100, "unit": "KB" },
    "totalPage": { "budget": 1000, "unit": "KB" },
    "criticalCSS": { "budget": 14, "unit": "KB" },
    "mainBundle": { "budget": 150, "unit": "KB", "compressed": true }
  },
  "lighthouseScores": {
    "performance": 90,
    "accessibility": 95,
    "bestPractices": 95,
    "seo": 95
  }
}
```

---

## Quick Decision Guide

```
Slow LCP?
├── TTFB > 800ms?
│   ├── Yes → CDN, server optimization, reduce redirects
│   └── No → Continue
├── LCP resource discovered late?
│   ├── Yes → Add preload, fetchpriority="high", avoid JS rendering
│   └── No → Continue
├── LCP resource too large?
│   ├── Yes → Optimize image (WebP/AVIF, responsive sizes, compression)
│   └── No → Continue
└── Render blocked?
    └── Yes → Inline critical CSS, defer non-critical CSS/JS

Slow INP?
├── Input delay > 100ms?
│   ├── Yes → Long tasks blocking main thread. Break up work, defer third-party.
│   └── No → Continue
├── Processing time > 100ms?
│   ├── Yes → Optimize event handler. Use transitions, defer non-visual work.
│   └── No → Continue
└── Presentation delay > 100ms?
    └── Yes → Complex rendering. Use CSS containment, virtualize lists.

High CLS?
├── Images without dimensions?
│   ├── Yes → Add width/height or aspect-ratio
│   └── No → Continue
├── Dynamic content injection?
│   ├── Yes → Reserve space with min-height or placeholders
│   └── No → Continue
├── Font swap causing reflow?
│   ├── Yes → Use font-display: optional, or size-adjust
│   └── No → Continue
└── Animations moving layout?
    └── Yes → Use transform instead of top/left/width/height
```
