# Web Performance Quick Reference

## Core Web Vitals Targets

| Metric | Good | Needs Work | Poor | Measures |
|--------|------|------------|------|----------|
| **LCP** | < 2.5s | 2.5-4s | > 4s | Largest visible element render |
| **INP** | < 200ms | 200-500ms | > 500ms | Interaction responsiveness |
| **CLS** | < 0.1 | 0.1-0.25 | > 0.25 | Visual stability |
| FCP | < 1.8s | 1.8-3s | > 3s | First content paint |
| TTFB | < 800ms | 800ms-1.8s | > 1.8s | Server response time |

## LCP Checklist

```
□ Identify LCP element (DevTools → Performance → Timings → LCP)
□ Preload LCP image: <link rel="preload" as="image" href="..." fetchpriority="high">
□ Set fetchpriority="high" on LCP <img>
□ Never lazy-load LCP image
□ Preconnect to image CDN: <link rel="preconnect" href="https://cdn...">
□ Inline critical CSS for above-the-fold content
□ Avoid CSS background-image for LCP element
□ Server-render LCP content (not client-side rendered)
```

## INP Checklist

```
□ Break long tasks (>50ms) into chunks with yield-to-main
□ Use startTransition for non-urgent React updates
□ Debounce expensive event handlers (300ms for search)
□ Move heavy computation to Web Workers
□ Avoid synchronous layout reads after DOM writes
□ Minimize JavaScript execution in event handlers
```

## CLS Checklist

```
□ Set width + height on all <img>, <video>, <iframe>
□ Use aspect-ratio for responsive media containers
□ Reserve space for dynamic content (min-h-[Xpx])
□ Use font-display: swap + size-adjust on fallback
□ Use transform for animations (not top/left/width/height)
□ Don't insert content above existing content
□ Add contain: layout to dynamic sections
```

## Bundle Size Targets

| Bundle | Target | Red Flag |
|--------|--------|----------|
| Initial JS | < 100KB gzip | > 200KB |
| Vendor chunk | < 80KB gzip | > 150KB |
| Per-route chunk | < 50KB gzip | > 100KB |
| Total CSS | < 30KB gzip | > 60KB |
| LCP image | < 100KB | > 200KB |

## Common Bundle Bloat Replacements

| Heavy Library | Size | Replacement | Size |
|--------------|------|-------------|------|
| moment.js | 72KB | date-fns | 3-10KB (tree-shaken) |
| lodash | 72KB | lodash-es (tree-shake) | 2-5KB typical |
| axios | 13KB | fetch (native) | 0KB |
| uuid | 4KB | crypto.randomUUID() | 0KB |
| classnames | 1KB | clsx | 0.5KB |
| numeral | 8KB | Intl.NumberFormat | 0KB |

## React Performance Patterns

### When to Memoize

| Pattern | Use When | Skip When |
|---------|----------|-----------|
| `React.memo` | Expensive render + stable props | Cheap render or always-new props |
| `useMemo` | Expensive computation (filter/sort 100+ items) | Simple expressions |
| `useCallback` | Passed to memoized children | Not passed down or children not memoized |

### Code Splitting Decision

```
Route-level:     React.lazy(() => import('./pages/Settings'))
Component-level: React.lazy(() => import('./HeavyChart'))    // >50KB components
Library-level:   dynamic import on interaction               // Large libs (charts, editors)
```

### Virtualization Threshold

```
< 50 items:   Render all (no virtualization needed)
50-500 items:  Consider virtualizing if render is slow
> 500 items:   Always virtualize (@tanstack/react-virtual)
```

## Cache-Control Quick Reference

| Resource Type | Header |
|--------------|--------|
| Hashed static assets (JS, CSS) | `public, max-age=31536000, immutable` |
| HTML pages | `public, max-age=0, must-revalidate` + ETag |
| Public API data | `public, max-age=60, stale-while-revalidate=300` |
| Authenticated API data | `private, no-cache` |
| Sensitive data (tokens) | `private, no-store` |
| Images (with hash) | `public, max-age=31536000, immutable` |
| Images (without hash) | `public, max-age=86400` + ETag |

### Cache-Control Directives

| Directive | Meaning |
|-----------|---------|
| `public` | Any cache (CDN, browser) can store |
| `private` | Only browser can store |
| `max-age=N` | Fresh for N seconds |
| `s-maxage=N` | CDN-specific max-age |
| `no-cache` | Cache but always revalidate (NOT "don't cache") |
| `no-store` | Never cache at all |
| `immutable` | Never changes, skip revalidation |
| `stale-while-revalidate=N` | Serve stale while fetching fresh |

## React Query / TanStack Query

```typescript
// Key concepts
staleTime:  How long data is "fresh" (won't refetch)     // Default: 0
gcTime:     How long inactive data stays in memory        // Default: 5 min
            (was cacheTime in v4)

// Common config
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 5 * 60 * 1000,     // 5 min fresh
      gcTime: 30 * 60 * 1000,       // 30 min in cache
      retry: 3,
      refetchOnWindowFocus: true,
    },
  },
});

// Invalidation after mutation
onSettled: () => queryClient.invalidateQueries({ queryKey: ['posts'] })

// Optimistic update pattern
onMutate → cancelQueries → getQueryData (snapshot) → setQueryData (optimistic)
onError → setQueryData (rollback from snapshot)
onSettled → invalidateQueries (refetch truth)
```

## Service Worker Strategies (Workbox)

| Strategy | Behavior | Use For |
|----------|----------|---------|
| CacheFirst | Cache → Network (fallback) | Static assets, fonts, images |
| StaleWhileRevalidate | Cache immediately → Update in background | API data, non-critical resources |
| NetworkFirst | Network → Cache (fallback) | HTML pages, real-time data |
| NetworkOnly | Always network | Auth endpoints, mutations |
| CacheOnly | Always cache | Precached assets only |

## Resource Hints

```html
<!-- In <head>, ordered by priority -->
<link rel="preconnect" href="https://fonts.googleapis.com">       <!-- Critical 3rd party -->
<link rel="preload" href="/hero.webp" as="image" fetchpriority="high">  <!-- LCP image -->
<link rel="preload" href="/font.woff2" as="font" crossorigin>    <!-- Primary font -->
<link rel="dns-prefetch" href="https://analytics.example.com">    <!-- Non-critical 3rd party -->
<link rel="prefetch" href="/dashboard.js">                         <!-- Next page resources -->
```

**Limit preload to 3-5 resources** — each competes for bandwidth.

## Script Loading

| Attribute | Download | Execute | Order |
|-----------|----------|---------|-------|
| (none) | Blocks parsing | Blocks parsing | In order |
| `defer` | Parallel | After DOM ready | In order |
| `async` | Parallel | Immediately | No order |
| `type="module"` | Parallel | After DOM ready | In order (deferred) |

**Rule**: Use `defer` for app scripts. Use `async` only for independent scripts (analytics). Never block parsing with undecorated `<script>`.

## Image Optimization

```html
<!-- LCP image: eager + high priority -->
<img src="hero.webp" fetchpriority="high" loading="eager" width="1200" height="600">

<!-- Below-fold: lazy load -->
<img src="photo.webp" loading="lazy" decoding="async" width="400" height="300">

<!-- Responsive -->
<img
  srcset="photo-400.webp 400w, photo-800.webp 800w, photo-1200.webp 1200w"
  sizes="(max-width: 640px) 400px, (max-width: 1024px) 800px, 1200px"
  src="photo-800.webp" loading="lazy">
```

**Format preference**: AVIF > WebP > JPEG. Use `<picture>` for format fallbacks.

## Performance Measurement Commands

```bash
# Lighthouse CLI
npx lighthouse http://localhost:3000 --output=html --output-path=./report.html
npx lighthouse http://localhost:3000 --only-categories=performance --output=json

# Bundle analysis (Vite)
npx vite-bundle-visualizer

# Bundle analysis (Webpack)
npx webpack-bundle-analyzer stats.json

# Check gzip sizes
find dist -name '*.js' -exec sh -c 'echo "$(gzip -c "$1" | wc -c) $1"' _ {} \;

# Find large dependencies
du -sh node_modules/* | sort -rh | head -20
```

## Performance Budget (CI)

```javascript
// lighthouserc.js
module.exports = {
  ci: {
    assert: {
      assertions: {
        'categories:performance': ['error', { minScore: 0.9 }],
        'largest-contentful-paint': ['error', { maxNumericValue: 2500 }],
        'cumulative-layout-shift': ['error', { maxNumericValue: 0.1 }],
        'total-byte-weight': ['error', { maxNumericValue: 500000 }],
      },
    },
  },
};
```

## Common Anti-Patterns

| Anti-Pattern | Impact | Fix |
|-------------|--------|-----|
| No code splitting | Huge initial bundle | React.lazy per route |
| Import entire lodash | +72KB | `import debounce from 'lodash/debounce'` |
| Unoptimized images | Slow LCP | WebP/AVIF + proper sizing |
| No font preload | Flash of unstyled text | `<link rel="preload" as="font">` |
| Render-blocking scripts | Delayed FCP | `defer` or `async` |
| No caching headers | Repeat downloads | Set Cache-Control per resource type |
| Layout without dimensions | CLS | width + height on media elements |
| All state in context | Cascade re-renders | Split contexts or use Zustand |
| Re-render on every keystroke | Poor INP | Debounce + startTransition |
| Loading all data upfront | Slow TTFB | Pagination + lazy loading |
