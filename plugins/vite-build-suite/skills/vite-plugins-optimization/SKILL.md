---
name: vite-plugins-optimization
description: >
  Vite plugins, build optimization, code splitting, tree shaking,
  bundle analysis, and performance tuning.
  Triggers: "vite plugin", "vite optimize", "vite code splitting",
  "vite tree shaking", "vite bundle", "vite chunk", "vite performance",
  "vite build slow", "vite bundle size", "vite lazy load".
  NOT for: basic configuration (use vite-config).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Vite Plugins & Optimization

## Essential Plugins

```bash
# Framework
npm i -D @vitejs/plugin-react-swc  # React with SWC (fastest)
npm i -D @vitejs/plugin-react      # React with Babel (if you need Babel plugins)
npm i -D @vitejs/plugin-vue        # Vue
npm i -D @sveltejs/vite-plugin-svelte  # Svelte

# Build quality
npm i -D vite-plugin-dts            # Generate .d.ts for library mode
npm i -D vite-plugin-checker        # TypeScript, ESLint checks in dev
npm i -D vite-tsconfig-paths        # Auto-resolve tsconfig paths

# Assets
npm i -D vite-plugin-svgr           # SVG as React components
npm i -D vite-plugin-image-optimizer # Optimize images at build time
npm i -D vite-plugin-compression    # Gzip/Brotli compress output

# Testing
npm i -D vitest                     # Vite-native testing
npm i -D @vitest/coverage-v8        # Coverage

# PWA
npm i -D vite-plugin-pwa            # Progressive Web App

# Analysis
npm i -D vite-bundle-visualizer     # Bundle treemap
npm i -D rollup-plugin-visualizer   # Alternative visualizer
```

## Plugin Configuration Examples

```typescript
// vite.config.ts
import { defineConfig } from "vite";
import react from "@vitejs/plugin-react-swc";
import checker from "vite-plugin-checker";
import svgr from "vite-plugin-svgr";
import compression from "vite-plugin-compression";
import { visualizer } from "rollup-plugin-visualizer";

export default defineConfig({
  plugins: [
    // React with SWC (3-20x faster than Babel)
    react(),

    // Type checking + ESLint in dev overlay
    checker({
      typescript: true,
      eslint: {
        lintCommand: 'eslint "./src/**/*.{ts,tsx}"',
        useFlatConfig: true,
      },
      overlay: { initialIsOpen: false },
    }),

    // SVGs as components: import Logo from './logo.svg?react'
    svgr({
      svgrOptions: {
        plugins: ["@svgr/plugin-svgo", "@svgr/plugin-jsx"],
        svgoConfig: {
          plugins: [{ name: "removeViewBox", active: false }],
        },
      },
    }),

    // Gzip + Brotli compression
    compression({ algorithm: "gzip" }),
    compression({ algorithm: "brotliCompress" }),

    // Bundle analysis (only in build)
    visualizer({
      filename: "stats.html",
      open: true,
      gzipSize: true,
      brotliSize: true,
    }),
  ],
});
```

## Code Splitting

```typescript
// Route-based code splitting (React)
import { lazy, Suspense } from "react";

const Dashboard = lazy(() => import("./pages/Dashboard"));
const Settings = lazy(() => import("./pages/Settings"));
const Analytics = lazy(() => import("./pages/Analytics"));

function App() {
  return (
    <Suspense fallback={<LoadingSpinner />}>
      <Routes>
        <Route path="/dashboard" element={<Dashboard />} />
        <Route path="/settings" element={<Settings />} />
        <Route path="/analytics" element={<Analytics />} />
      </Routes>
    </Suspense>
  );
}

// Named chunks for better debugging
const Dashboard = lazy(() =>
  import(/* webpackChunkName: "dashboard" */ "./pages/Dashboard")
);
```

```typescript
// Manual chunk splitting
// vite.config.ts
export default defineConfig({
  build: {
    rollupOptions: {
      output: {
        manualChunks: {
          // Group vendor deps into named chunks
          "vendor-react": ["react", "react-dom", "react-router-dom"],
          "vendor-ui": ["@radix-ui/react-dialog", "@radix-ui/react-dropdown-menu"],
          "vendor-utils": ["date-fns", "zod", "clsx"],
          "vendor-charts": ["recharts", "d3"],
        },
      },
    },
  },
});

// Dynamic manual chunks (function form)
export default defineConfig({
  build: {
    rollupOptions: {
      output: {
        manualChunks(id) {
          // Put all node_modules in a vendor chunk
          if (id.includes("node_modules")) {
            // Split large deps into own chunk
            if (id.includes("@radix-ui")) return "vendor-radix";
            if (id.includes("recharts") || id.includes("d3")) return "vendor-charts";
            if (id.includes("firebase")) return "vendor-firebase";
            return "vendor"; // everything else
          }
        },
      },
    },
  },
});
```

## Dynamic Imports

```typescript
// Conditional feature loading
async function loadEditor() {
  const { Editor } = await import("./components/Editor");
  return Editor;
}

// Load heavy library only when needed
async function generatePDF() {
  const { jsPDF } = await import("jspdf");
  const doc = new jsPDF();
  doc.text("Hello", 10, 10);
  doc.save("output.pdf");
}

// Preload critical chunks
// In <head>: <link rel="modulepreload" href="/assets/dashboard-abc123.js">
// Or programmatically:
const link = document.createElement("link");
link.rel = "modulepreload";
link.href = "/assets/critical-chunk.js";
document.head.appendChild(link);
```

## Tree Shaking

```typescript
// GOOD: Named imports — tree-shakeable
import { format, parseISO } from "date-fns";

// BAD: Default import of the whole library
import _ from "lodash"; // bundles ALL of lodash (~70KB)

// GOOD: Cherry-pick from lodash
import debounce from "lodash/debounce";
import throttle from "lodash/throttle";
// Or use lodash-es (ESM build, fully tree-shakeable)
import { debounce, throttle } from "lodash-es";

// Ensure package.json has sideEffects for proper tree shaking
// package.json
{
  "sideEffects": false,
  // Or specify files with side effects:
  "sideEffects": ["*.css", "*.scss", "./src/polyfills.ts"]
}
```

```typescript
// vite.config.ts — tree shaking is on by default in production
// To debug tree shaking issues:
export default defineConfig({
  build: {
    rollupOptions: {
      treeshake: {
        moduleSideEffects: false, // treat all modules as side-effect-free
        // preset: "recommended",  // or use recommended settings
      },
    },
  },
});
```

## Bundle Analysis

```bash
# Quick analysis
npx vite-bundle-visualizer

# With rollup-plugin-visualizer (more options)
# Add to vite.config.ts plugins array, then:
npm run build
# Opens stats.html with treemap

# Check bundle size from CLI
npx vite build --report
```

### What to Look For

| Problem | Threshold | Fix |
|---------|-----------|-----|
| Total bundle > 500KB | Warning | Code split, lazy load |
| Single chunk > 250KB | Warning | Split with manualChunks |
| Duplicate dependency | Any | Check with `npm ls <dep>`, dedupe |
| moment.js | 300KB+ | Replace with date-fns or dayjs |
| lodash (full) | 70KB+ | Use lodash-es or cherry-pick |
| All icons imported | Varies | Import individual icons |
| Unused CSS | > 50KB | PurgeCSS or Tailwind JIT |
| Source maps in prod | N/A | Disable unless debugging |

## Production Optimization

```typescript
// vite.config.ts — production optimizations
export default defineConfig({
  build: {
    // Use terser for smaller output (slower build)
    minify: "terser",
    terserOptions: {
      compress: {
        drop_console: true,      // remove console.log
        drop_debugger: true,     // remove debugger statements
        pure_funcs: ["console.log", "console.info"],
      },
      mangle: {
        safari10: true,          // Safari 10 compat
      },
    },

    // Target modern browsers (smaller polyfills)
    target: "es2022",

    // Inline small assets
    assetsInlineLimit: 8192, // 8KB — inline small images as base64

    // CSS code splitting
    cssCodeSplit: true,

    // Rollup options
    rollupOptions: {
      output: {
        // Deterministic chunk names for caching
        chunkFileNames: "assets/js/[name]-[hash].js",
        entryFileNames: "assets/js/[name]-[hash].js",
        assetFileNames: "assets/[ext]/[name]-[hash].[ext]",

        // Manual chunks
        manualChunks: {
          vendor: ["react", "react-dom"],
        },
      },
    },
  },
});
```

## CSS Optimization

```typescript
// vite.config.ts
export default defineConfig({
  css: {
    // CSS Modules
    modules: {
      localsConvention: "camelCaseOnly",
      generateScopedName: "[name]__[local]___[hash:base64:5]",
    },

    // PostCSS (auto-detected from postcss.config.js)
    postcss: {
      plugins: [
        require("autoprefixer"),
        require("cssnano")({ preset: "default" }),
      ],
    },

    // Lightning CSS (faster alternative to PostCSS)
    // npm i -D lightningcss
    transformer: "lightningcss",
    lightningcss: {
      targets: { chrome: 100, firefox: 100, safari: 15 },
      drafts: { customMedia: true },
    },
  },
});
```

## Dev Server Performance

```typescript
// vite.config.ts — speed up dev server
export default defineConfig({
  // Pre-bundle problematic dependencies
  optimizeDeps: {
    include: [
      "react",
      "react-dom",
      "react-router-dom",
      // Include deps that are slow to pre-bundle
      "@radix-ui/react-dialog",
      "date-fns",
    ],
    exclude: [
      // Exclude deps that cause issues when pre-bundled
      "@local/package",
    ],
    // Force re-optimization
    force: true, // set temporarily to clear cache
  },

  // Reduce file system watching overhead
  server: {
    watch: {
      // Ignore large directories
      ignored: ["**/node_modules/**", "**/.git/**", "**/dist/**"],
    },
    // Warm up frequently used files
    warmup: {
      clientFiles: [
        "./src/components/Layout.tsx",
        "./src/components/Sidebar.tsx",
        "./src/lib/api.ts",
      ],
    },
  },
});
```

## Writing Custom Plugins

```typescript
// Simple plugin: add build timestamp
function buildTimestamp(): Plugin {
  return {
    name: "build-timestamp",
    transformIndexHtml(html) {
      return html.replace(
        "</head>",
        `<meta name="build-time" content="${new Date().toISOString()}" /></head>`
      );
    },
  };
}

// Plugin with hooks
function myPlugin(): Plugin {
  return {
    name: "my-plugin",

    // Runs once when config is resolved
    configResolved(config) {
      console.log("Mode:", config.mode);
    },

    // Runs for each module
    transform(code, id) {
      if (id.endsWith(".ts") && code.includes("__DEBUG__")) {
        return code.replace(/__DEBUG__/g, "false");
      }
    },

    // Runs during build
    generateBundle(options, bundle) {
      // Analyze or modify output chunks
      for (const [fileName, chunk] of Object.entries(bundle)) {
        if (chunk.type === "chunk") {
          console.log(`${fileName}: ${chunk.code.length} bytes`);
        }
      }
    },

    // Dev server middleware
    configureServer(server) {
      server.middlewares.use("/api/health", (req, res) => {
        res.end(JSON.stringify({ status: "ok" }));
      });
    },
  };
}
```

## Vitest Integration

```typescript
// vite.config.ts — shared config for app + tests
/// <reference types="vitest/config" />
import { defineConfig } from "vite";
import react from "@vitejs/plugin-react-swc";

export default defineConfig({
  plugins: [react()],
  test: {
    globals: true,
    environment: "jsdom",
    setupFiles: "./src/test/setup.ts",
    css: true,
    coverage: {
      provider: "v8",
      reporter: ["text", "json", "html"],
      exclude: ["node_modules/", "src/test/"],
    },
  },
});
```

## PWA Plugin

```typescript
import { VitePWA } from "vite-plugin-pwa";

export default defineConfig({
  plugins: [
    VitePWA({
      registerType: "autoUpdate",
      manifest: {
        name: "My App",
        short_name: "App",
        theme_color: "#ffffff",
        icons: [
          { src: "/icon-192.png", sizes: "192x192", type: "image/png" },
          { src: "/icon-512.png", sizes: "512x512", type: "image/png" },
        ],
      },
      workbox: {
        globPatterns: ["**/*.{js,css,html,ico,png,svg,woff2}"],
        runtimeCaching: [
          {
            urlPattern: /^https:\/\/api\.example\.com\/.*/i,
            handler: "NetworkFirst",
            options: {
              cacheName: "api-cache",
              expiration: { maxEntries: 50, maxAgeSeconds: 300 },
            },
          },
        ],
      },
    }),
  ],
});
```

## Gotchas

1. **SWC plugin vs Babel plugin** — `@vitejs/plugin-react-swc` is 3-20x faster than `@vitejs/plugin-react` (Babel). Use SWC unless you need specific Babel plugins (styled-components, relay, etc.).

2. **`manualChunks` can break lazy loading** — if you put a dependency in a manual chunk that's also lazy-imported, it forces the chunk to load eagerly. Test your code splitting with the network tab.

3. **`visualizer` plugin should only run during build** — wrap it in a condition: `mode === 'analyze' && visualizer({...})`. Otherwise it slows down dev.

4. **CSS code splitting and `@import` order** — Vite splits CSS per JS chunk. If you rely on CSS cascade order across chunks, you may get FOUC. Use CSS Modules or utility-first CSS to avoid order dependencies.

5. **`optimizeDeps.force: true` is a debugging tool** — don't leave it on in committed config. It forces re-optimization on every dev server start, which is slow.

6. **esbuild minification drops `console.log` differently** — use `terser` with `drop_console: true` for reliable console removal. esbuild's `drop` option exists but has edge cases with conditional logging.
