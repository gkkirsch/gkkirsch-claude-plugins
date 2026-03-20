# Vite Cheatsheet

## Quick Start

```bash
npm create vite@latest my-app -- --template react-ts
cd my-app && npm i && npm run dev
```

## Config Essentials

```typescript
// vite.config.ts
import { defineConfig } from "vite"
import react from "@vitejs/plugin-react-swc"
import path from "path"

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: { "@": path.resolve(__dirname, "./src") },
  },
  server: {
    port: 3000,
    proxy: { "/api": { target: "http://localhost:8080", changeOrigin: true } },
  },
  build: {
    target: "es2022",
    sourcemap: false,
    minify: "esbuild",
  },
})
```

## Environment Variables

```bash
# .env — only VITE_ prefix exposed to client
VITE_API_URL=https://api.example.com
DATABASE_URL=postgres://...  # server-only (vite.config.ts)
```

```typescript
const url = import.meta.env.VITE_API_URL
const isDev = import.meta.env.DEV
const isProd = import.meta.env.PROD
const mode = import.meta.env.MODE
```

## Path Aliases

```typescript
// vite.config.ts
resolve: { alias: { "@": path.resolve(__dirname, "./src") } }

// tsconfig.json (must match!)
{ "compilerOptions": { "paths": { "@/*": ["./src/*"] } } }
```

## Code Splitting

```typescript
// Route-based (React)
const Dashboard = lazy(() => import("./pages/Dashboard"))
const Settings = lazy(() => import("./pages/Settings"))

// Manual chunks
build: {
  rollupOptions: {
    output: {
      manualChunks: {
        "vendor-react": ["react", "react-dom"],
        "vendor-ui": ["@radix-ui/react-dialog"],
      },
    },
  },
},
```

## Essential Plugins

```bash
npm i -D @vitejs/plugin-react-swc    # React (fastest)
npm i -D vite-plugin-checker          # TS + ESLint in dev
npm i -D vite-plugin-svgr             # SVGs as components
npm i -D vite-plugin-compression      # Gzip/Brotli
npm i -D rollup-plugin-visualizer     # Bundle analysis
npm i -D vite-plugin-dts              # .d.ts for libraries
npm i -D vite-plugin-pwa              # PWA support
```

## Library Mode

```typescript
build: {
  lib: {
    entry: resolve(__dirname, "src/index.ts"),
    formats: ["es", "cjs"],
    fileName: (fmt) => `lib.${fmt === "es" ? "mjs" : "cjs"}`,
  },
  rollupOptions: {
    external: ["react", "react-dom"],
  },
},
```

## CSS

```typescript
css: {
  modules: { localsConvention: "camelCase" },
  preprocessorOptions: {
    scss: { additionalData: `@use "@/styles/vars" as *;` },
  },
  // Lightning CSS (faster than PostCSS)
  transformer: "lightningcss",
},
```

## Dev Performance

```typescript
optimizeDeps: {
  include: ["react", "react-dom"],  // pre-bundle
  exclude: ["@local/pkg"],          // skip
},
server: {
  warmup: { clientFiles: ["./src/App.tsx"] },
},
```

## Production Optimization

```typescript
build: {
  target: "es2022",
  minify: "terser",
  terserOptions: { compress: { drop_console: true } },
  cssCodeSplit: true,
  assetsInlineLimit: 8192,
  rollupOptions: {
    output: {
      manualChunks(id) {
        if (id.includes("node_modules")) return "vendor"
      },
    },
  },
},
```

## SSR

```typescript
// vite.config.ts
ssr: {
  noExternal: ["some-cjs-package"],
  external: ["express"],
  target: "node",
},
```

## Bundle Analysis

```bash
npx vite-bundle-visualizer      # opens treemap
npm run build -- --report        # CLI report
```

## Vitest

```typescript
// vite.config.ts
test: {
  globals: true,
  environment: "jsdom",
  setupFiles: "./src/test/setup.ts",
  coverage: { provider: "v8" },
},
```

## CLI

```bash
npm run dev           # start dev server
npm run build         # production build
npm run preview       # preview production build
npx vite --host       # expose to network
npx vite --port 4000  # custom port
npx vite build --mode staging  # custom mode
npx vite optimize     # force dep optimization
```

## Gotchas

1. Only `VITE_` prefixed env vars reach client code
2. Path aliases need BOTH vite.config.ts AND tsconfig.json
3. `import.meta.env` is statically replaced — no dynamic keys
4. `public/` files served as-is, no processing — reference as `/file.png`
5. SWC plugin is 3-20x faster than Babel — use unless you need Babel plugins
6. `optimizeDeps.force: true` is for debugging — don't commit it
7. `manualChunks` can break lazy loading if deps overlap
8. CSS order across chunks can cause FOUC — use CSS Modules
