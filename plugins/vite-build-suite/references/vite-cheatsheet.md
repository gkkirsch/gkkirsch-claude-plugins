# Vite & Build Tools Cheatsheet

## CLI Commands

| Command | Use |
|---------|-----|
| `npm create vite@latest` | Create new project |
| `npx vite` | Start dev server |
| `npx vite build` | Production build |
| `npx vite preview` | Preview production build |
| `npx vite --force` | Clear dep cache and restart |
| `npx vite --host` | Expose to network |
| `npx vite --port 4000` | Custom port |
| `npx vite optimize --force` | Force re-optimize deps |
| `npx vitest` | Run tests (watch mode) |
| `npx vitest run` | Run tests once (CI) |
| `npx vitest run --coverage` | Run with coverage |
| `npx vitest --ui` | Browser test viewer |

## Vite Templates

```bash
npm create vite@latest my-app -- --template <template>
```

| Template | Stack |
|----------|-------|
| `react-ts` | React + TypeScript |
| `react-swc-ts` | React + SWC + TypeScript (fastest) |
| `vue-ts` | Vue 3 + TypeScript |
| `svelte-ts` | Svelte + TypeScript |
| `solid-ts` | Solid + TypeScript |
| `preact-ts` | Preact + TypeScript |
| `vanilla-ts` | Vanilla + TypeScript |

## Environment Variables

| Variable | Access | Scope |
|----------|--------|-------|
| `VITE_*` | `import.meta.env.VITE_*` | Client + server |
| Non-prefixed | `loadEnv()` in config only | Server only |
| `import.meta.env.MODE` | Auto | `"development"` or `"production"` |
| `import.meta.env.DEV` | Auto | `true` in dev |
| `import.meta.env.PROD` | Auto | `true` in prod |
| `import.meta.env.BASE_URL` | Auto | Base URL from config |
| `import.meta.env.SSR` | Auto | `true` during SSR |

### .env File Priority

```
.env                  # Always loaded
.env.local            # Always loaded, gitignored
.env.[mode]           # Only in specified mode
.env.[mode].local     # Only in specified mode, gitignored

Priority: .env.[mode].local > .env.[mode] > .env.local > .env
```

## Asset Handling

| Import Syntax | Result |
|---------------|--------|
| `import img from './img.png'` | URL string |
| `import raw from './file.txt?raw'` | Raw string content |
| `import url from './file.json?url'` | URL (force) |
| `import worker from './worker.js?worker'` | Web Worker |
| `import wasm from './lib.wasm?init'` | WASM module |
| `import svg from './icon.svg?react'` | React component (with vite-plugin-svgr) |

## import.meta.glob

```typescript
// Lazy loading (default) — code-split modules
const modules = import.meta.glob('./modules/*.ts');
// Returns: { './modules/a.ts': () => import('./modules/a.ts'), ... }

// Eager loading — bundled together
const modules = import.meta.glob('./modules/*.ts', { eager: true });
// Returns: { './modules/a.ts': { default: ..., export1: ... }, ... }

// Named imports only
const modules = import.meta.glob('./modules/*.ts', { import: 'setup' });

// As strings
const modules = import.meta.glob('./modules/*.ts', { query: '?raw', import: 'default' });
```

## Config Quick Reference

```typescript
// vite.config.ts — minimal production-ready
import { defineConfig } from "vite";
import react from "@vitejs/plugin-react-swc";

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: { "@": "/src" },
  },
  server: {
    port: 3000,
    proxy: { "/api": "http://localhost:8080" },
  },
  build: {
    target: "es2020",
    sourcemap: false,
    rollupOptions: {
      output: {
        manualChunks: {
          vendor: ["react", "react-dom"],
        },
      },
    },
  },
});
```

## Build Output Structure

```
dist/
  index.html
  assets/
    index-[hash].js       # Entry chunk
    vendor-[hash].js      # Vendor chunk (manual)
    Dashboard-[hash].js   # Route chunk (lazy)
    index-[hash].css      # CSS
    logo-[hash].svg       # Static assets
```

## Plugin Hook Order

```
config          → Modify config
configResolved  → Read final config
configureServer → Add dev server middleware
buildStart      → Build starting
resolveId       → Resolve import paths
load            → Load file contents
transform       → Transform file contents
buildEnd        → Build complete
closeBundle     → After bundle written
```

## Optimization Checklist

| Check | How |
|-------|-----|
| Bundle size | `npx vite-bundle-visualizer` |
| Unused code | Check tree shaking with `sideEffects: false` in package.json |
| Large deps | Replace lodash → lodash-es, moment → date-fns |
| Image size | Use `vite-plugin-imagemin` or WebP format |
| CSS unused | PurgeCSS or Tailwind's built-in purge |
| Code splitting | `React.lazy()` for route components |
| Chunk grouping | `manualChunks` for vendor deps |
| Compression | `vite-plugin-compression` for gzip/brotli |
| Caching | Content hashes in filenames (default) |
| Preload | `<link rel="modulepreload">` (auto-generated) |

## Vitest Quick Reference

```typescript
// Test file naming: *.test.ts or *.spec.ts

// Assertions
expect(value).toBe(exact);
expect(value).toEqual(deep);
expect(value).toBeTruthy();
expect(value).toBeFalsy();
expect(value).toBeNull();
expect(value).toBeDefined();
expect(value).toContain(item);
expect(value).toHaveLength(n);
expect(value).toMatch(/regex/);
expect(fn).toThrow("message");
expect(fn).toHaveBeenCalled();
expect(fn).toHaveBeenCalledWith(arg);
expect(fn).toHaveBeenCalledTimes(n);
expect(value).toMatchSnapshot();
expect(value).toMatchInlineSnapshot();

// Mocking
vi.fn()                         // Mock function
vi.fn(() => 42)                 // Mock with implementation
vi.spyOn(obj, "method")        // Spy on method
vi.mock("./module")             // Mock module (hoisted)
vi.doMock("./module")           // Mock module (not hoisted)
vi.mocked(fn)                   // TypeScript-typed mock
vi.useFakeTimers()              // Fake timers
vi.advanceTimersByTime(ms)      // Advance fake timers
vi.useRealTimers()              // Restore real timers
vi.stubEnv("KEY", "value")     // Mock env var
vi.unstubAllEnvs()              // Restore env vars

// Lifecycle
beforeAll(() => {});
afterAll(() => {});
beforeEach(() => {});
afterEach(() => {});

// Test organization
describe("group", () => {});
it("test", () => {});
it.skip("skipped", () => {});
it.only("focused", () => {});
it.todo("future test");
it.each([1, 2, 3])("test %i", (n) => {});
```

## Common Vite Errors

| Error | Fix |
|-------|-----|
| `Failed to resolve import` | Check alias config, ensure file exists |
| `Pre-bundling... new dependencies` | Run `npx vite --force` to rebuild |
| `Cannot use import statement outside module` | Add `"type": "module"` to package.json |
| `Hydration mismatch` | SSR rendering differs from client — check dynamic content |
| `[plugin:vite:css] Preprocessor not found` | Install `sass` or `less` package |
| `process is not defined` | Use `import.meta.env` instead of `process.env` |
| `global is not defined` | Add `define: { global: "globalThis" }` to config |
| `require is not defined` | Use `import` instead, or add to `optimizeDeps.include` |
