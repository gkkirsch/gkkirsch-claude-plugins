---
name: vite-config
description: >
  Vite configuration — vite.config.ts, dev server, proxy, environment
  variables, aliases, CSS, library mode, SSR, and multi-page apps.
  Triggers: "vite config", "vite setup", "vite proxy", "vite env",
  "vite alias", "vite library", "vite ssr", "vite multi page",
  "vite css", "vite dev server".
  NOT for: plugins or optimization (use vite-plugins-optimization).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Vite Configuration

## Quick Start

```bash
# Create new project
npm create vite@latest my-app -- --template react-ts
# Templates: vanilla, vanilla-ts, react, react-ts, react-swc, react-swc-ts,
#            vue, vue-ts, svelte, svelte-ts, solid, solid-ts, preact, preact-ts

cd my-app && npm install && npm run dev
```

## Full Configuration

```typescript
// vite.config.ts
import { defineConfig, loadEnv } from "vite";
import react from "@vitejs/plugin-react-swc";
import path from "path";

export default defineConfig(({ command, mode }) => {
  // Load env files based on mode (.env, .env.local, .env.[mode])
  const env = loadEnv(mode, process.cwd(), "");

  return {
    // Framework plugin
    plugins: [react()],

    // Path aliases
    resolve: {
      alias: {
        "@": path.resolve(__dirname, "./src"),
        "@components": path.resolve(__dirname, "./src/components"),
        "@lib": path.resolve(__dirname, "./src/lib"),
        "@hooks": path.resolve(__dirname, "./src/hooks"),
        "@types": path.resolve(__dirname, "./src/types"),
      },
    },

    // Dev server
    server: {
      port: 3000,
      open: true,
      host: true, // listen on all addresses (for Docker/network)

      // API proxy
      proxy: {
        "/api": {
          target: "http://localhost:8080",
          changeOrigin: true,
          // rewrite: (path) => path.replace(/^\/api/, ""), // strip /api prefix
        },
        "/ws": {
          target: "ws://localhost:8080",
          ws: true,
        },
      },

      // CORS
      cors: true,

      // Custom headers
      headers: {
        "X-Custom-Header": "value",
      },
    },

    // Preview server (production build preview)
    preview: {
      port: 4173,
    },

    // Build options
    build: {
      outDir: "dist",
      sourcemap: mode === "development", // sourcemaps only in dev
      minify: "esbuild", // 'esbuild' (default, fast) or 'terser' (smaller, slower)
      target: "es2020", // browser target
      cssTarget: "chrome80",
      assetsInlineLimit: 4096, // inline assets < 4KB as base64
      chunkSizeWarningLimit: 500, // warn for chunks > 500KB

      rollupOptions: {
        output: {
          // Asset file naming
          assetFileNames: "assets/[name]-[hash][extname]",
          chunkFileNames: "assets/[name]-[hash].js",
          entryFileNames: "assets/[name]-[hash].js",
        },
      },
    },

    // CSS
    css: {
      modules: {
        localsConvention: "camelCase", // converts kebab-case to camelCase
        scopeBehaviour: "local",
      },
      preprocessorOptions: {
        scss: {
          additionalData: `@use "@/styles/variables" as *;`,
        },
      },
      devSourcemap: true,
    },

    // Dependency optimization
    optimizeDeps: {
      include: ["react", "react-dom", "react-router-dom"], // force pre-bundle
      exclude: ["@local/package"], // skip pre-bundling
    },

    // Define global constants
    define: {
      __APP_VERSION__: JSON.stringify(process.env.npm_package_version),
      __BUILD_TIME__: JSON.stringify(new Date().toISOString()),
    },
  };
});
```

## TypeScript Path Aliases

```json
// tsconfig.json — must match vite.config.ts aliases
{
  "compilerOptions": {
    "baseUrl": ".",
    "paths": {
      "@/*": ["./src/*"],
      "@components/*": ["./src/components/*"],
      "@lib/*": ["./src/lib/*"],
      "@hooks/*": ["./src/hooks/*"],
      "@types/*": ["./src/types/*"]
    }
  }
}
```

## Environment Variables

```bash
# .env                  — loaded in all cases
# .env.local            — loaded in all cases, gitignored
# .env.[mode]           — loaded only in specified mode
# .env.[mode].local     — loaded only in specified mode, gitignored

# Priority: .env.[mode].local > .env.[mode] > .env.local > .env

# Only VITE_ prefixed vars are exposed to client code
VITE_API_URL=https://api.example.com
VITE_APP_TITLE=My App

# Non-prefixed vars are only available in vite.config.ts (server-side)
DATABASE_URL=postgres://localhost/mydb
SECRET_KEY=my-secret
```

```typescript
// Access in code
const apiUrl = import.meta.env.VITE_API_URL;
const mode = import.meta.env.MODE;      // "development" | "production"
const isDev = import.meta.env.DEV;       // true in dev
const isProd = import.meta.env.PROD;     // true in prod
const baseUrl = import.meta.env.BASE_URL; // "/" by default

// Type declarations
// src/vite-env.d.ts
/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_API_URL: string;
  readonly VITE_APP_TITLE: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
```

## Library Mode

```typescript
// vite.config.ts — for building npm packages
import { defineConfig } from "vite";
import react from "@vitejs/plugin-react-swc";
import { resolve } from "path";
import dts from "vite-plugin-dts";

export default defineConfig({
  plugins: [
    react(),
    dts({ include: ["src"] }), // generate .d.ts files
  ],
  build: {
    lib: {
      entry: resolve(__dirname, "src/index.ts"),
      name: "MyLibrary",
      formats: ["es", "cjs"],
      fileName: (format) => `my-library.${format === "es" ? "mjs" : "cjs"}`,
    },
    rollupOptions: {
      // Don't bundle peer dependencies
      external: ["react", "react-dom", "react/jsx-runtime"],
      output: {
        globals: {
          react: "React",
          "react-dom": "ReactDOM",
        },
      },
    },
    sourcemap: true,
    minify: false, // let consumers handle minification
  },
});
```

```json
// package.json for library
{
  "name": "my-library",
  "version": "1.0.0",
  "type": "module",
  "main": "./dist/my-library.cjs",
  "module": "./dist/my-library.mjs",
  "types": "./dist/index.d.ts",
  "exports": {
    ".": {
      "import": "./dist/my-library.mjs",
      "require": "./dist/my-library.cjs",
      "types": "./dist/index.d.ts"
    },
    "./styles": "./dist/style.css"
  },
  "files": ["dist"],
  "peerDependencies": {
    "react": ">=18",
    "react-dom": ">=18"
  }
}
```

## Multi-Page App

```typescript
// vite.config.ts
import { resolve } from "path";

export default defineConfig({
  build: {
    rollupOptions: {
      input: {
        main: resolve(__dirname, "index.html"),
        admin: resolve(__dirname, "admin/index.html"),
        login: resolve(__dirname, "login/index.html"),
      },
    },
  },
});
```

```
project/
  index.html          → /
  admin/
    index.html        → /admin/
  login/
    index.html        → /login/
```

## SSR Configuration

```typescript
// vite.config.ts
export default defineConfig({
  ssr: {
    // CJS dependencies that need to be bundled
    noExternal: ["some-cjs-package"],

    // Node.js built-ins and deps to keep external
    external: ["express"],

    // SSR build target
    target: "node",
  },

  build: {
    ssr: true, // or specify entry: build.ssr = 'src/server-entry.ts'
  },
});
```

```typescript
// src/entry-server.ts
import { renderToString } from "react-dom/server";
import { App } from "./App";

export function render() {
  return renderToString(<App />);
}
```

```typescript
// server.js
import express from "express";
import { createServer as createViteServer } from "vite";

async function start() {
  const app = express();

  // Create Vite dev server in middleware mode
  const vite = await createViteServer({
    server: { middlewareMode: true },
    appType: "custom",
  });

  app.use(vite.middlewares);

  app.use("*", async (req, res) => {
    const template = await vite.transformIndexHtml(
      req.originalUrl,
      fs.readFileSync("index.html", "utf-8")
    );
    const { render } = await vite.ssrLoadModule("/src/entry-server.ts");
    const html = template.replace("<!--ssr-outlet-->", render());
    res.status(200).set({ "Content-Type": "text/html" }).end(html);
  });

  app.listen(3000);
}

start();
```

## Conditional Config

```typescript
// vite.config.ts — different config per mode
export default defineConfig(({ command, mode }) => {
  if (command === "serve") {
    // Dev-specific config
    return {
      plugins: [react()],
      server: { port: 3000 },
    };
  }

  // Production build
  return {
    plugins: [react()],
    build: {
      sourcemap: false,
      minify: "terser",
      terserOptions: {
        compress: { drop_console: true },
      },
    },
  };
});
```

## Gotchas

1. **Only `VITE_` prefixed env vars are exposed to client** — anything without the prefix stays server-side (vite.config.ts only). This is a security feature.

2. **Path aliases need BOTH vite.config.ts AND tsconfig.json** — Vite resolves at bundle time, TypeScript resolves at type-check time. They must match.

3. **`import.meta.env` is statically replaced** — `import.meta.env[dynamicKey]` won't work. The replacement happens at build time, not runtime.

4. **CSS `@import` in `<style>` tags uses Vite's resolver** — you can use aliases (`@/styles/vars.css`) directly in CSS imports inside components.

5. **The `public/` directory is served as-is** — files are NOT processed by Vite. Use it for favicon, robots.txt, etc. Reference as `/file.png` (no `/public` prefix).

6. **`optimizeDeps.include` is for CJS deps** — Vite pre-bundles CJS dependencies to ESM on first run. If a dep causes issues, try adding it to `include` or `exclude`.
