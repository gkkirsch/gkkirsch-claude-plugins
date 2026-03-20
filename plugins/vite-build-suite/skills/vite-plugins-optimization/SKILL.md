---
name: vite-plugins-optimization
description: >
  Vite plugins and build optimization — official plugins, popular community
  plugins, code splitting, tree shaking, bundle analysis, lazy loading,
  caching, and production build performance.
  Triggers: "vite plugin", "vite optimization", "code splitting", "tree shaking",
  "bundle analysis", "lazy loading", "vite performance", "chunk splitting",
  "manual chunks", "bundle size".
  NOT for: basic vite.config.ts setup (use vite-config).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Vite Plugins & Build Optimization

## Official Plugins

```bash
# Framework plugins
npm install @vitejs/plugin-react-swc    # React with SWC (fastest)
npm install @vitejs/plugin-react        # React with Babel
npm install @vitejs/plugin-vue          # Vue 3
npm install @vitejs/plugin-vue-jsx      # Vue JSX
npm install @vitejs/plugin-legacy       # Legacy browser support
```

```typescript
// @vitejs/plugin-react-swc (recommended for React)
import react from "@vitejs/plugin-react-swc";

export default defineConfig({
  plugins: [react()],
});

// @vitejs/plugin-legacy (IE11 / old browsers)
import legacy from "@vitejs/plugin-legacy";

export default defineConfig({
  plugins: [
    legacy({
      targets: ["defaults", "not IE 11"],
      // generates legacy chunks + polyfills
    }),
  ],
});
```

## Essential Community Plugins

```typescript
// vite-plugin-dts — TypeScript declaration generation for libraries
import dts from "vite-plugin-dts";

export default defineConfig({
  plugins: [dts({ include: ["src"] })],
});

// vite-plugin-svgr — Import SVGs as React components
import svgr from "vite-plugin-svgr";

export default defineConfig({
  plugins: [svgr()],
});
// Usage: import { ReactComponent as Logo } from './logo.svg?react';
//    or: import Logo from './logo.svg?react';

// vite-plugin-pwa — Progressive Web App support
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
          { src: "pwa-192x192.png", sizes: "192x192", type: "image/png" },
          { src: "pwa-512x512.png", sizes: "512x512", type: "image/png" },
        ],
      },
      workbox: {
        globPatterns: ["**/*.{js,css,html,ico,png,svg}"],
      },
    }),
  ],
});

// vite-plugin-compression — gzip/brotli compressed assets
import viteCompression from "vite-plugin-compression";

export default defineConfig({
  plugins: [
    viteCompression({ algorithm: "gzip" }),
    viteCompression({ algorithm: "brotliCompress" }),
  ],
});

// @vitejs/plugin-basic-ssl — HTTPS in dev
import basicSsl from "@vitejs/plugin-basic-ssl";

export default defineConfig({
  plugins: [basicSsl()],
  server: { https: true },
});

// vite-tsconfig-paths — Auto-resolve tsconfig paths (alternative to manual aliases)
import tsconfigPaths from "vite-tsconfig-paths";

export default defineConfig({
  plugins: [tsconfigPaths()],
});
```

## Code Splitting

```typescript
// Route-based code splitting with React.lazy
import { lazy, Suspense } from "react";

const Dashboard = lazy(() => import("./pages/Dashboard"));
const Settings = lazy(() => import("./pages/Settings"));
const Profile = lazy(() => import("./pages/Profile"));

function App() {
  return (
    <Suspense fallback={<Loading />}>
      <Routes>
        <Route path="/dashboard" element={<Dashboard />} />
        <Route path="/settings" element={<Settings />} />
        <Route path="/profile" element={<Profile />} />
      </Routes>
    </Suspense>
  );
}

// Named chunks for debugging
const Admin = lazy(() => import(/* webpackChunkName: "admin" */ "./pages/Admin"));
```

```typescript
// Manual chunks — group vendor dependencies
// vite.config.ts
export default defineConfig({
  build: {
    rollupOptions: {
      output: {
        manualChunks: {
          // Group React ecosystem
          "vendor-react": ["react", "react-dom", "react-router-dom"],
          // Group UI library
          "vendor-ui": ["@radix-ui/react-dialog", "@radix-ui/react-dropdown-menu"],
          // Group utilities
          "vendor-utils": ["date-fns", "zod", "clsx"],
          // Group data fetching
          "vendor-data": ["@tanstack/react-query", "axios"],
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
          if (id.includes("node_modules")) {
            // Split by package name
            const packageName = id.split("node_modules/")[1].split("/")[0];

            // Group large packages individually
            if (["react", "react-dom"].includes(packageName)) {
              return "vendor-react";
            }
            if (packageName.startsWith("@radix-ui")) {
              return "vendor-ui";
            }
            // Everything else in a shared vendor chunk
            return "vendor";
          }
        },
      },
    },
  },
});
```

## Bundle Analysis

```bash
# Visual treemap of bundle contents
npx vite-bundle-visualizer

# Rollup plugin for analysis
npm install -D rollup-plugin-visualizer
```

```typescript
// vite.config.ts
import { visualizer } from "rollup-plugin-visualizer";

export default defineConfig({
  plugins: [
    visualizer({
      filename: "stats.html",
      open: true,         // auto-open in browser
      gzipSize: true,     // show gzip sizes
      brotliSize: true,   // show brotli sizes
      template: "treemap", // 'treemap', 'sunburst', 'network'
    }),
  ],
});
```

```bash
# Check bundle size from CLI
npx vite build && ls -la dist/assets/*.js | sort -k 5 -n

# Analyze with source-map-explorer
npx source-map-explorer dist/assets/*.js
```

## Tree Shaking

```typescript
// Package.json sideEffects for tree shaking
{
  "sideEffects": false,  // All files are pure (safe to tree shake)
  // Or specify files with side effects:
  "sideEffects": ["*.css", "*.scss", "./src/polyfills.ts"]
}

// Named exports are tree-shakeable, default exports are not
// GOOD — tree shakeable:
export function formatDate(d: Date) { ... }
export function parseDate(s: string) { ... }

// BAD — entire module included:
export default { formatDate, parseDate };

// GOOD — import only what you need:
import { formatDate } from "date-fns";

// BAD — imports everything:
import * as dateFns from "date-fns";
import _ from "lodash"; // Use lodash-es or individual imports
```

## Image & Asset Optimization

```typescript
// vite-plugin-imagemin — compress images at build time
import viteImagemin from "vite-plugin-imagemin";

export default defineConfig({
  plugins: [
    viteImagemin({
      gifsicle: { optimizationLevel: 3 },
      mozjpeg: { quality: 80 },
      pngquant: { quality: [0.65, 0.8] },
      svgo: {
        plugins: [
          { name: "removeViewBox", active: false },
          { name: "removeEmptyAttrs", active: false },
        ],
      },
      webp: { quality: 80 }, // also generate WebP versions
    }),
  ],
});
```

```typescript
// Asset handling in vite.config.ts
export default defineConfig({
  build: {
    assetsInlineLimit: 4096, // Inline assets < 4KB as base64
    assetsDir: "assets",
  },
  // Custom asset handling
  assetsInclude: ["**/*.glb", "**/*.gltf"], // Treat as assets (not code)
});

// Import assets in code
import logoUrl from "./logo.svg";           // URL string
import logoRaw from "./logo.svg?raw";       // Raw SVG string
import workerUrl from "./worker.js?worker"; // Web Worker
import shaderCode from "./shader.glsl?raw"; // Raw text
```

## Performance Optimizations

```typescript
// vite.config.ts — production optimizations
export default defineConfig({
  build: {
    // Use esbuild for faster minification (default)
    minify: "esbuild",
    // Or terser for smaller output (slower build)
    // minify: "terser",
    // terserOptions: {
    //   compress: { drop_console: true, drop_debugger: true },
    // },

    // Target modern browsers only
    target: "es2020",
    cssTarget: "chrome80",

    // Sourcemaps only when needed
    sourcemap: false, // or 'hidden' for error tracking without exposing

    // CSS code splitting
    cssCodeSplit: true, // true = per-chunk CSS files (default)

    // Chunk size warnings
    chunkSizeWarningLimit: 500, // KB
  },

  // Dependency pre-bundling
  optimizeDeps: {
    include: [
      "react",
      "react-dom",
      "react-router-dom",
      "@tanstack/react-query",
    ],
    exclude: ["@vite/client"], // Vite internals
  },

  // Experimental features
  experimental: {
    renderBuiltUrl(filename, { type }) {
      // Custom CDN URL for assets
      if (type === "asset") {
        return `https://cdn.example.com/${filename}`;
      }
    },
  },
});
```

## Custom Plugin Development

```typescript
// Minimal Vite plugin
function myPlugin(): Plugin {
  return {
    name: "my-plugin",
    enforce: "pre", // or 'post'

    // Config hook — modify Vite config
    config(config, { command }) {
      if (command === "build") {
        return { build: { sourcemap: true } };
      }
    },

    // Transform hook — modify file contents
    transform(code, id) {
      if (id.endsWith(".md")) {
        return `export default ${JSON.stringify(marked(code))}`;
      }
    },

    // Build hooks
    buildStart() {
      console.log("Build starting...");
    },

    // Dev server hooks
    configureServer(server) {
      server.middlewares.use("/my-api", (req, res) => {
        res.end(JSON.stringify({ ok: true }));
      });
    },

    // HMR handling
    handleHotUpdate({ file, server }) {
      if (file.endsWith(".md")) {
        console.log("Markdown file changed:", file);
        // Custom HMR logic
      }
    },
  };
}
```

## Gotchas

1. **`manualChunks` can break dynamic imports.** If a dynamically imported module shares a dependency with a manual chunk, Vite may create circular chunk references. Test with `vite build && vite preview` to verify.

2. **`import.meta.glob` is eagerly evaluated.** `import.meta.glob('./modules/*.ts')` creates code-split points for each file. Use `{ eager: true }` only if you need all modules loaded immediately. The default lazy loading is usually what you want.

3. **Tree shaking doesn't work with CommonJS.** If a dependency only publishes CJS (not ESM), Vite can't tree shake it. Check if an ESM alternative exists (e.g., `lodash-es` instead of `lodash`).

4. **`vite-plugin-legacy` nearly doubles build time.** It generates two bundles (modern + legacy). Only add it if you actually need to support old browsers. Check your analytics first.

5. **Pre-bundling runs once and caches.** When you add a new dependency, Vite may not detect it immediately. Run `npx vite --force` to clear the pre-bundle cache and rebuild.

6. **CSS `@import` order matters for specificity.** Vite processes CSS imports in the order they appear. If your global styles load after component styles, specificity issues arise. Import globals first in your entry file.
