---
name: bundle-optimization
description: >
  JavaScript bundle optimization — code splitting, tree shaking, dynamic imports,
  dependency analysis, Vite/webpack configuration, and size reduction strategies.
  Triggers: "bundle size", "code splitting", "tree shaking", "dynamic import",
  "webpack optimization", "vite optimization", "reduce bundle".
  NOT for: Runtime performance or Web Vitals (use web-vitals-optimization).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Bundle Optimization

## Analyze First

```bash
# Vite bundle analysis
npx vite-bundle-visualizer
# Opens interactive treemap showing every module's size

# Webpack bundle analysis
npx webpack-bundle-analyzer stats.json
# Generate stats: npx webpack --profile --json > stats.json

# Quick size check
npx bundlephobia <package-name>
# Shows minified + gzipped size before you install

# Source map explorer
npx source-map-explorer dist/assets/*.js
```

## Code Splitting

### Route-Based Splitting (React)

```typescript
import { lazy, Suspense } from "react";
import { Routes, Route } from "react-router-dom";

// Each route becomes its own chunk
const Home = lazy(() => import("./pages/Home"));
const Dashboard = lazy(() => import("./pages/Dashboard"));
const Settings = lazy(() => import("./pages/Settings"));
const AdminPanel = lazy(() => import("./pages/AdminPanel"));

function App() {
  return (
    <Suspense fallback={<PageSkeleton />}>
      <Routes>
        <Route path="/" element={<Home />} />
        <Route path="/dashboard" element={<Dashboard />} />
        <Route path="/settings" element={<Settings />} />
        <Route path="/admin/*" element={<AdminPanel />} />
      </Routes>
    </Suspense>
  );
}
```

### Component-Level Splitting

```typescript
// Heavy components loaded on demand
const MarkdownEditor = lazy(() => import("./components/MarkdownEditor"));
const ChartDashboard = lazy(() => import("./components/ChartDashboard"));
const PDFViewer = lazy(() => import("./components/PDFViewer"));

function PostEditor({ mode }: { mode: "simple" | "markdown" }) {
  return (
    <Suspense fallback={<EditorSkeleton />}>
      {mode === "markdown" ? <MarkdownEditor /> : <SimpleEditor />}
    </Suspense>
  );
}

// Preload on hover/focus for instant navigation
function NavLink({ to, children }: { to: string; children: React.ReactNode }) {
  const preload = () => {
    if (to === "/dashboard") import("./pages/Dashboard");
    if (to === "/settings") import("./pages/Settings");
  };

  return (
    <Link to={to} onMouseEnter={preload} onFocus={preload}>
      {children}
    </Link>
  );
}
```

### Dynamic Imports for Libraries

```typescript
// BAD: Import heavy library at top level
import { marked } from "marked";       // 40KB gzipped, loaded on every page
import hljs from "highlight.js";       // 290KB, loaded even if unused

// GOOD: Import only when needed
async function renderMarkdown(content: string): Promise<string> {
  const { marked } = await import("marked");
  const hljs = await import("highlight.js");

  marked.setOptions({
    highlight: (code, lang) => hljs.highlight(code, { language: lang }).value,
  });

  return marked(content);
}

// GOOD: Conditional heavy imports
async function exportToPDF(data: ReportData) {
  const { jsPDF } = await import("jspdf");
  const doc = new jsPDF();
  // ... generate PDF
}
```

## Tree Shaking

```typescript
// BAD: Barrel imports defeat tree shaking
import { Button, Icon, Modal, Tooltip, Dropdown } from "./components";
// Imports ALL components even if you only use Button

// GOOD: Direct imports — only Button is bundled
import { Button } from "./components/Button";

// BAD: lodash full import (71KB)
import _ from "lodash";
_.debounce(fn, 300);

// GOOD: Cherry-pick (4KB)
import debounce from "lodash/debounce";

// BEST: Use lodash-es for tree shaking (only debounce is bundled)
import { debounce } from "lodash-es";

// BAD: Importing entire icon library
import { FaHome, FaUser } from "react-icons/fa";
// Actually imports ALL Font Awesome icons

// GOOD: Direct icon imports
import { FaHome } from "react-icons/fa/FaHome";
import { FaUser } from "react-icons/fa/FaUser";
```

### Ensure Tree Shaking Works

```json
// package.json — mark your package as side-effect free
{
  "sideEffects": false
}

// Or specify files with side effects
{
  "sideEffects": ["*.css", "*.scss", "./src/polyfills.ts"]
}
```

```typescript
// vite.config.ts — verify tree shaking
export default defineConfig({
  build: {
    rollupOptions: {
      output: {
        // Analyze what's included
        experimentalMinChunkSize: 10_000,
      },
    },
  },
});
```

## Vite Configuration

```typescript
// vite.config.ts — production optimization
import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  build: {
    target: "es2020",
    minify: "esbuild",       // Fastest minifier (default)
    sourcemap: true,          // For debugging, remove in prod if needed
    cssMinify: "lightningcss", // Faster CSS minification

    rollupOptions: {
      output: {
        // Manual chunk splitting
        manualChunks: {
          // Vendor chunks — cached separately from app code
          "react-vendor": ["react", "react-dom", "react-router-dom"],
          "ui-vendor": ["@radix-ui/react-dialog", "@radix-ui/react-dropdown-menu"],
          "query": ["@tanstack/react-query"],
        },
      },
    },

    // Chunk size warnings
    chunkSizeWarningLimit: 500, // KB
  },
});
```

### Advanced Vite Optimizations

```typescript
export default defineConfig({
  // Pre-bundle dependencies for faster dev startup
  optimizeDeps: {
    include: ["react", "react-dom", "axios"],
    exclude: ["@my/local-package"],
  },

  build: {
    rollupOptions: {
      output: {
        manualChunks(id) {
          // Split node_modules into separate vendor chunks
          if (id.includes("node_modules")) {
            const name = id.split("node_modules/")[1].split("/")[0];
            // Group small deps together, split large ones
            if (["react", "react-dom", "react-router-dom"].includes(name)) {
              return "react-vendor";
            }
            if (name.startsWith("@radix-ui") || name.startsWith("@headlessui")) {
              return "ui-vendor";
            }
            return "vendor"; // Everything else
          }
        },
      },
    },
  },
});
```

## Dependency Optimization

### Replace Heavy Libraries

| Heavy Library | Size (gzip) | Lightweight Alternative | Size (gzip) |
|--------------|-------------|------------------------|-------------|
| moment | 72KB | date-fns | 5-15KB (tree-shakeable) |
| moment | 72KB | dayjs | 2KB |
| lodash | 71KB | lodash-es | tree-shakeable |
| lodash | 71KB | Native JS | 0KB |
| axios | 13KB | fetch (native) | 0KB |
| uuid | 3KB | crypto.randomUUID() | 0KB |
| classnames | 1KB | clsx | 0.5KB |
| numeral | 17KB | Intl.NumberFormat | 0KB |

### Audit Dependencies

```bash
# Find duplicate packages
npx depcheck

# Check for unused dependencies
npx depcheck --ignores="@types/*,eslint-*"

# Find duplicate versions of the same package
npx npm-dedupe
# or
npx yarn-deduplicate

# Check total install size
npx cost-of-modules
```

## Compression

```typescript
// Vite: Enable brotli + gzip compression
import viteCompression from "vite-plugin-compression";

export default defineConfig({
  plugins: [
    react(),
    viteCompression({ algorithm: "brotliCompress" }),  // .br files
    viteCompression({ algorithm: "gzip" }),             // .gz files
  ],
});
```

```nginx
# Nginx: Serve pre-compressed files
location /assets/ {
  gzip_static on;           # Serve .gz files if they exist
  brotli_static on;         # Serve .br files if they exist
  expires 1y;               # Cache immutable assets
  add_header Cache-Control "public, immutable";
}
```

| Algorithm | Compression Ratio | Speed | Support |
|-----------|------------------|-------|---------|
| Brotli | Best (20-25% smaller than gzip) | Slower to compress | All modern browsers |
| Gzip | Good | Fast | Universal |
| None | Baseline | N/A | N/A |

## Performance Budget

```json
// bundlesize in package.json
{
  "bundlesize": [
    { "path": "dist/assets/index-*.js", "maxSize": "150 kB" },
    { "path": "dist/assets/vendor-*.js", "maxSize": "100 kB" },
    { "path": "dist/assets/index-*.css", "maxSize": "30 kB" }
  ]
}
```

```typescript
// Vite: Fail build on oversized chunks
export default defineConfig({
  build: {
    chunkSizeWarningLimit: 200, // Warn at 200KB
  },
  plugins: [{
    name: "size-limit",
    closeBundle() {
      // Custom size check in CI
      const stats = fs.statSync("dist/assets/index.js");
      if (stats.size > 200 * 1024) {
        throw new Error(`Bundle too large: ${(stats.size / 1024).toFixed(0)}KB`);
      }
    },
  }],
});
```

## Quick Wins Checklist

| Optimization | Effort | Impact |
|-------------|--------|--------|
| Route-based code splitting | Low | High |
| Replace moment with dayjs | Low | Medium |
| Direct imports (no barrel files) | Low | Medium |
| Dynamic import for heavy components | Low | High |
| manualChunks for vendor splitting | Low | Medium |
| Brotli compression | Low | Medium |
| Remove unused dependencies (depcheck) | Low | Medium |
| Tree-shake lodash → lodash-es | Low | Medium |
| Preload critical chunks | Medium | Medium |
| Component-level lazy loading | Medium | High |
| Replace axios with fetch | Medium | Low |
| Image optimization (AVIF/WebP) | Medium | High |

## Gotchas

1. **Barrel files (`index.ts` re-exports) kill tree shaking** — `export * from "./Button"` in an index file forces bundlers to include everything. Import directly from the component file, not the barrel.

2. **Dynamic imports create network waterfalls** — `lazy(() => import("./Page"))` fetches the chunk only when needed. If the chunk imports another chunk, you get a waterfall. Use `modulePreload` or preload on hover to mitigate.

3. **`manualChunks` can create circular dependencies** — If chunk A imports from chunk B and vice versa, the bundler may fail or create unnecessary chunks. Test with `vite build --debug` and check the chunk graph.

4. **CSS-in-JS libraries add to JS bundle** — styled-components, Emotion, etc. ship their runtime in your JS bundle (5-15KB). Tailwind CSS and CSS Modules compile to static CSS with zero runtime cost.

5. **Source maps double your deploy size** — Source maps are useful for debugging but shouldn't be served to users. Generate them for error tracking (Sentry) but configure your CDN to not serve `.map` files publicly.

6. **Compression doesn't help already-compressed assets** — Images (JPEG, WebP, AVIF), fonts (WOFF2), and videos are already compressed. Only enable gzip/brotli for text assets (JS, CSS, HTML, JSON, SVG).
