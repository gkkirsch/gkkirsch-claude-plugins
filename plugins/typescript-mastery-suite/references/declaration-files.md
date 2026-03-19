# Declaration Files Reference

Guide to writing, consuming, and maintaining TypeScript declaration files (`.d.ts`).

---

## What Are Declaration Files?

Declaration files (`.d.ts`) describe the shape of JavaScript code without containing implementations. They tell TypeScript what types exist, what functions are available, and what modules export — without any runtime code.

```typescript
// math.d.ts — declares what math.js exports
export declare function add(a: number, b: number): number;
export declare function multiply(a: number, b: number): number;
export declare const PI: number;
```

---

## Writing Declaration Files

### For a CommonJS Module

```typescript
// types/legacy-lib.d.ts
declare module "legacy-lib" {
  // Default export
  export default function createClient(options: ClientOptions): Client;

  // Named exports
  export function connect(url: string): Promise<Connection>;
  export const VERSION: string;

  // Types
  export interface ClientOptions {
    host: string;
    port: number;
    timeout?: number;
    retries?: number;
  }

  export interface Client {
    get(path: string): Promise<any>;
    post(path: string, body: unknown): Promise<any>;
    put(path: string, body: unknown): Promise<any>;
    delete(path: string): Promise<void>;
    close(): void;
  }

  export interface Connection {
    readonly connected: boolean;
    disconnect(): void;
  }

  // Class export
  export class EventEmitter {
    on(event: string, listener: (...args: any[]) => void): this;
    off(event: string, listener: (...args: any[]) => void): this;
    emit(event: string, ...args: any[]): boolean;
    once(event: string, listener: (...args: any[]) => void): this;
  }
}
```

### For an ESM Module

```typescript
// types/modern-lib.d.ts
declare module "modern-lib" {
  export function process<T>(data: T[]): ProcessedResult<T>;
  export function transform<T, U>(data: T, fn: (item: T) => U): U;

  export type ProcessedResult<T> = {
    items: T[];
    count: number;
    timestamp: Date;
  };

  export interface Config {
    mode: "development" | "production";
    features: string[];
    plugins?: Plugin[];
  }

  export interface Plugin {
    name: string;
    version: string;
    apply(config: Config): Config;
  }
}
```

### For a Module with Subpaths

```typescript
// types/big-lib/index.d.ts
declare module "big-lib" {
  export function init(config: Config): void;
  export interface Config { /* ... */ }
}

// types/big-lib/utils.d.ts
declare module "big-lib/utils" {
  export function formatDate(date: Date, format: string): string;
  export function parseJSON<T>(json: string): T;
}

// types/big-lib/testing.d.ts
declare module "big-lib/testing" {
  export function createMock<T>(partial?: Partial<T>): T;
  export function resetMocks(): void;
}
```

### For Global Scripts (No Module System)

```typescript
// types/global-lib.d.ts
// No import/export — declares globals

// Global variable
declare var GLOBAL_CONFIG: {
  apiUrl: string;
  debug: boolean;
};

// Global function
declare function $(selector: string): HTMLElement | null;
declare function $$(selector: string): NodeListOf<HTMLElement>;

// Global class
declare class Analytics {
  static track(event: string, properties?: Record<string, unknown>): void;
  static identify(userId: string): void;
  static page(name: string): void;
}

// Global interface
interface Window {
  __INITIAL_STATE__: Record<string, unknown>;
  gtag: (...args: any[]) => void;
}
```

---

## Ambient Modules

Ambient module declarations describe modules that exist at runtime but have no TypeScript source.

### Wildcard Module Declarations

```typescript
// types/assets.d.ts

// CSS Modules
declare module "*.module.css" {
  const classes: Record<string, string>;
  export default classes;
}

declare module "*.module.scss" {
  const classes: Record<string, string>;
  export default classes;
}

// Images
declare module "*.png" {
  const src: string;
  export default src;
}

declare module "*.jpg" {
  const src: string;
  export default src;
}

declare module "*.svg" {
  import type { FC, SVGProps } from "react";
  const ReactComponent: FC<SVGProps<SVGSVGElement>>;
  export default ReactComponent;
}

// Raw file content
declare module "*.txt" {
  const content: string;
  export default content;
}

// YAML/TOML
declare module "*.yaml" {
  const data: Record<string, unknown>;
  export default data;
}

// GraphQL
declare module "*.graphql" {
  import { DocumentNode } from "graphql";
  const value: DocumentNode;
  export default value;
}

// Web Workers
declare module "*?worker" {
  const WorkerConstructor: new () => Worker;
  export default WorkerConstructor;
}

// URL imports (Vite)
declare module "*?url" {
  const url: string;
  export default url;
}
```

---

## Module Augmentation

Extend existing module types without modifying the original declarations.

### Augmenting npm Packages

```typescript
// types/express-augment.d.ts
import "express";

declare module "express" {
  interface Request {
    user?: {
      id: string;
      email: string;
      role: "admin" | "user";
    };
    requestId: string;
    startTime: number;
  }

  interface Response {
    sendSuccess(data: unknown): Response;
    sendError(statusCode: number, message: string): Response;
  }
}

// types/express-session-augment.d.ts
import "express-session";

declare module "express-session" {
  interface SessionData {
    userId: string;
    loginAt: Date;
    preferences: {
      theme: "light" | "dark";
      language: string;
    };
  }
}
```

### Augmenting Your Own Modules

```typescript
// src/events.ts — base event definitions
export interface AppEvents {
  "app:start": { timestamp: Date };
  "app:stop": { reason: string };
}

// src/features/auth/events.ts — augment from another feature
declare module "../events" {
  interface AppEvents {
    "auth:login": { userId: string; method: string };
    "auth:logout": { userId: string };
    "auth:failed": { email: string; reason: string };
  }
}

// src/features/payments/events.ts — augment from another feature
declare module "../../events" {
  interface AppEvents {
    "payment:created": { amount: number; currency: string };
    "payment:completed": { transactionId: string };
    "payment:failed": { error: string };
  }
}

// All features' events are now merged into AppEvents
```

### Augmenting Global Types

```typescript
// types/global-augment.d.ts
export {}; // Make this a module

declare global {
  // Extend built-in types
  interface Array<T> {
    groupBy<K extends string>(fn: (item: T) => K): Record<K, T[]>;
    unique(): T[];
  }

  interface String {
    truncate(maxLength: number): string;
  }

  // Add global variables
  var __APP_VERSION__: string;
  var __BUILD_TIME__: string;

  // Extend Window
  interface Window {
    ENV: {
      API_URL: string;
      SENTRY_DSN: string;
      FEATURE_FLAGS: Record<string, boolean>;
    };
  }

  // Extend ProcessEnv
  namespace NodeJS {
    interface ProcessEnv {
      NODE_ENV: "development" | "production" | "test";
      PORT: string;
      DATABASE_URL: string;
      REDIS_URL: string;
      JWT_SECRET: string;
      AWS_REGION: string;
      S3_BUCKET: string;
    }
  }
}
```

---

## Declaration Merging

TypeScript merges multiple declarations of the same name. This is how module augmentation works.

### What Merges

| Declaration | Namespace | Type | Value |
|-------------|-----------|------|-------|
| `namespace` | X | | X |
| `class` | | X | X |
| `enum` | | X | X |
| `interface` | | X | |
| `type alias` | | X | |
| `function` | | | X |
| `variable` | | | X |

**Interfaces merge:**

```typescript
interface Box {
  width: number;
  height: number;
}

interface Box {
  depth: number;
  color: string;
}

// Result: Box has width, height, depth, and color
const box: Box = { width: 10, height: 20, depth: 5, color: "red" };
```

**Namespaces merge with classes, functions, and enums:**

```typescript
class Album {
  label: Album.AlbumLabel = { name: "default" };
}

namespace Album {
  export interface AlbumLabel {
    name: string;
  }
}

// Album is both a class and a namespace
const album = new Album();
const label: Album.AlbumLabel = { name: "My Label" };
```

**Type aliases do NOT merge** — they error on duplicate names.

---

## DefinitelyTyped Contributions

When writing types for DefinitelyTyped (`@types/` packages):

### File Structure

```
types/my-package/
├── index.d.ts          # Main type declarations
├── my-package-tests.ts # Test file (validates types compile)
├── tsconfig.json       # Config for the package
└── tslint.json         # Linting config (legacy, being migrated)
```

### index.d.ts Template

```typescript
// Type definitions for my-package 2.5
// Project: https://github.com/author/my-package
// Definitions by: Your Name <https://github.com/yourusername>
// Definitions: https://github.com/DefinitelyTyped/DefinitelyTyped

export interface Options {
  timeout?: number;
  retries?: number;
}

export function doSomething(input: string, options?: Options): Promise<Result>;

export interface Result {
  data: unknown;
  status: number;
}

// If module has a default export
export default function myPackage(config: Options): Client;

export interface Client {
  request(method: string, path: string): Promise<Result>;
  close(): void;
}
```

### Test File

```typescript
// my-package-tests.ts
import myPackage, { doSomething, Options, Result } from "my-package";

// $ExpectType Promise<Result>
doSomething("test");

// $ExpectType Promise<Result>
doSomething("test", { timeout: 5000 });

const client = myPackage({ retries: 3 });

// $ExpectType Promise<Result>
client.request("GET", "/api/data");

// @ts-expect-error — should not accept number
doSomething(42);
```

---

## Configuration

### typeRoots and types

```jsonc
{
  "compilerOptions": {
    // Where to look for type declarations (default: node_modules/@types)
    "typeRoots": [
      "./types",              // Your custom declarations
      "./node_modules/@types" // DefinitelyTyped packages
    ],

    // Specific type packages to include (from typeRoots)
    // If specified, ONLY these types are included
    "types": ["node", "vitest/globals"]
    // If omitted, ALL types from typeRoots are included
  }
}
```

### Declaration File Discovery Order

1. `"types"` in tsconfig (if specified)
2. Packages listed in `typeRoots`
3. `node_modules/@types/` (default)
4. Package's own `"types"` or `"typings"` field in package.json
5. Package's `"exports"` with `"types"` condition
6. `index.d.ts` in package root

### Package.json Types Field

```jsonc
// For library authors
{
  "name": "my-library",
  "version": "1.0.0",
  "type": "module",

  // Single entry point
  "types": "./dist/index.d.ts",
  "main": "./dist/index.js",

  // Multiple entry points (preferred)
  "exports": {
    ".": {
      "types": "./dist/index.d.ts",        // MUST be first
      "import": "./dist/index.js",
      "require": "./dist/index.cjs"
    },
    "./utils": {
      "types": "./dist/utils.d.ts",
      "import": "./dist/utils.js"
    }
  }
}
```

---

## Best Practices

### Do

- Use `export declare` for module declarations
- Use `interface` for objects that others might augment
- Use `type` for unions, intersections, and computed types
- Include JSDoc comments for public APIs
- Test declarations with a `.test.ts` file
- Use `@ts-expect-error` to test that invalid code errors

### Don't

- Don't use `namespace` for new code (use modules instead)
- Don't declare globals unless absolutely necessary
- Don't use `any` — use `unknown` for truly unknown values
- Don't export mutable variables — use functions
- Don't forget `export {}` to make a file a module when augmenting globals

### Common Mistakes

```typescript
// WRONG: Missing export {} — this becomes a script, not a module
declare global {
  interface Window { foo: string; }
}
// This might work but can cause issues

// RIGHT: Include export {} to make it a module
declare global {
  interface Window { foo: string; }
}
export {};

// WRONG: Using type where interface is needed for augmentation
type AppEvents = { click: MouseEvent };
// Can't be augmented later!

// RIGHT: Use interface for extensible types
interface AppEvents { click: MouseEvent; }
// Can be augmented by other modules

// WRONG: Forgetting that declaration files can't have implementations
// types/helpers.d.ts
export function helper(x: string): string {  // Error!
  return x.toUpperCase();
}

// RIGHT: Declaration only
export declare function helper(x: string): string;
```

---

## Quick Reference

| Goal | Pattern |
|------|---------|
| Type an npm package | `declare module "package-name" { ... }` |
| Type asset imports | `declare module "*.png" { ... }` |
| Extend Express Request | `declare module "express" { interface Request { ... } }` |
| Add global variable | `declare global { var name: Type; }` |
| Extend Window | `declare global { interface Window { ... } }` |
| Extend ProcessEnv | `declare global { namespace NodeJS { interface ProcessEnv { ... } } }` |
| Type a class | `declare class Name { ... }` |
| Type a function | `declare function name(args): return` |
| Type a constant | `declare const name: Type` |
| Make file a module | Add `export {}` or any export statement |
