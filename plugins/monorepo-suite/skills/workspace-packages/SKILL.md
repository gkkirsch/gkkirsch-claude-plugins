---
name: workspace-packages
description: >
  Shared packages in a monorepo — UI component libraries, config packages,
  shared utilities, database packages, TypeScript configuration, ESLint config.
  Triggers: "shared package", "workspace package", "monorepo package", "shared ui",
  "shared config", "tsconfig base", "eslint config package", "@repo/", "internal package".
  NOT for: Turborepo pipeline config (use turborepo-setup), CI/CD (use monorepo-cicd).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Shared Workspace Packages

## TypeScript Config Package

```json
// packages/config-ts/package.json
{
  "name": "@repo/config-ts",
  "private": true,
  "files": ["*.json"]
}
```

```json
// packages/config-ts/base.json
{
  "$schema": "https://json.schemastore.org/tsconfig",
  "compilerOptions": {
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "forceConsistentCasingInFileNames": true,
    "moduleResolution": "bundler",
    "module": "esnext",
    "target": "es2022",
    "lib": ["es2022"],
    "declaration": true,
    "declarationMap": true,
    "sourceMap": true,
    "isolatedModules": true,
    "verbatimModuleSyntax": true,
    "resolveJsonModule": true,
    "noUncheckedIndexedAccess": true,
    "noImplicitOverride": true
  },
  "exclude": ["node_modules", "dist"]
}
```

```json
// packages/config-ts/nextjs.json
{
  "extends": "./base.json",
  "compilerOptions": {
    "lib": ["dom", "dom.iterable", "es2022"],
    "jsx": "preserve",
    "module": "esnext",
    "moduleResolution": "bundler",
    "allowJs": true,
    "noEmit": true,
    "incremental": true,
    "plugins": [{ "name": "next" }]
  }
}
```

```json
// packages/config-ts/library.json
{
  "extends": "./base.json",
  "compilerOptions": {
    "outDir": "dist",
    "rootDir": "src",
    "declaration": true,
    "declarationMap": true,
    "sourceMap": true
  },
  "include": ["src"],
  "exclude": ["node_modules", "dist", "**/*.test.ts"]
}
```

```json
// apps/web/tsconfig.json — consuming the config
{
  "extends": "@repo/config-ts/nextjs.json",
  "compilerOptions": {
    "paths": {
      "@/*": ["./src/*"]
    }
  },
  "include": ["next-env.d.ts", "src/**/*.ts", "src/**/*.tsx"],
  "exclude": ["node_modules"]
}
```

## ESLint Config Package

```json
// packages/config-eslint/package.json
{
  "name": "@repo/config-eslint",
  "private": true,
  "main": "index.mjs",
  "dependencies": {
    "@typescript-eslint/eslint-plugin": "^8.0",
    "@typescript-eslint/parser": "^8.0",
    "eslint-config-prettier": "^9.0",
    "eslint-plugin-import-x": "^4.0"
  },
  "peerDependencies": {
    "eslint": "^9.0"
  }
}
```

```javascript
// packages/config-eslint/index.mjs
import tsParser from "@typescript-eslint/parser";
import tsPlugin from "@typescript-eslint/eslint-plugin";
import importPlugin from "eslint-plugin-import-x";
import prettierConfig from "eslint-config-prettier";

/** @type {import("eslint").Linter.Config[]} */
export const base = [
  {
    files: ["**/*.ts", "**/*.tsx"],
    languageOptions: {
      parser: tsParser,
      parserOptions: {
        projectService: true,
      },
    },
    plugins: {
      "@typescript-eslint": tsPlugin,
      "import-x": importPlugin,
    },
    rules: {
      "@typescript-eslint/no-unused-vars": ["warn", { argsIgnorePattern: "^_" }],
      "@typescript-eslint/no-explicit-any": "warn",
      "@typescript-eslint/consistent-type-imports": "error",
      "import-x/order": ["error", {
        "groups": ["builtin", "external", "internal", "parent", "sibling"],
        "newlines-between": "always",
      }],
      "import-x/no-duplicates": "error",
    },
  },
  prettierConfig,
];

/** @type {import("eslint").Linter.Config[]} */
export const react = [
  ...base,
  {
    files: ["**/*.tsx"],
    rules: {
      // React-specific rules
    },
  },
];

/** @type {import("eslint").Linter.Config[]} */
export const node = [
  ...base,
  {
    files: ["**/*.ts"],
    rules: {
      "no-console": "off",
    },
  },
];
```

```javascript
// apps/web/eslint.config.mjs — consuming the config
import { react } from "@repo/config-eslint";

export default [
  ...react,
  { ignores: [".next/", "dist/"] },
];
```

## Shared UI Component Library

```json
// packages/ui/package.json
{
  "name": "@repo/ui",
  "private": true,
  "version": "0.0.0",
  "type": "module",
  "main": "./src/index.ts",
  "types": "./src/index.ts",
  "exports": {
    ".": {
      "types": "./src/index.ts",
      "default": "./src/index.ts"
    },
    "./button": {
      "types": "./src/components/button.tsx",
      "default": "./src/components/button.tsx"
    },
    "./card": {
      "types": "./src/components/card.tsx",
      "default": "./src/components/card.tsx"
    },
    "./globals.css": "./src/globals.css"
  },
  "scripts": {
    "typecheck": "tsc --noEmit",
    "lint": "eslint src/"
  },
  "devDependencies": {
    "@repo/config-ts": "workspace:*",
    "@repo/config-eslint": "workspace:*",
    "typescript": "^5.7"
  },
  "peerDependencies": {
    "react": "^19.0",
    "react-dom": "^19.0"
  }
}
```

```typescript
// packages/ui/src/index.ts — barrel export
export { Button, type ButtonProps } from "./components/button";
export { Card, CardHeader, CardTitle, CardContent, CardFooter } from "./components/card";
export { Input, type InputProps } from "./components/input";
export { Badge, type BadgeProps } from "./components/badge";
export { cn } from "./lib/utils";
```

```typescript
// packages/ui/src/lib/utils.ts
import { type ClassValue, clsx } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}
```

```typescript
// packages/ui/src/components/button.tsx
import { forwardRef, type ButtonHTMLAttributes } from "react";
import { cva, type VariantProps } from "class-variance-authority";
import { cn } from "../lib/utils";

const buttonVariants = cva(
  "inline-flex items-center justify-center rounded-md text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50",
  {
    variants: {
      variant: {
        default: "bg-primary text-primary-foreground hover:bg-primary/90",
        secondary: "bg-secondary text-secondary-foreground hover:bg-secondary/80",
        outline: "border border-input bg-background hover:bg-accent hover:text-accent-foreground",
        ghost: "hover:bg-accent hover:text-accent-foreground",
        destructive: "bg-destructive text-destructive-foreground hover:bg-destructive/90",
        link: "text-primary underline-offset-4 hover:underline",
      },
      size: {
        sm: "h-8 px-3 text-xs",
        default: "h-10 px-4 py-2",
        lg: "h-12 px-6 text-base",
        icon: "h-10 w-10",
      },
    },
    defaultVariants: { variant: "default", size: "default" },
  }
);

export interface ButtonProps
  extends ButtonHTMLAttributes<HTMLButtonElement>,
    VariantProps<typeof buttonVariants> {}

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant, size, ...props }, ref) => (
    <button
      className={cn(buttonVariants({ variant, size, className }))}
      ref={ref}
      {...props}
    />
  )
);
Button.displayName = "Button";
```

```typescript
// Consuming in an app:
// apps/web/package.json
{
  "dependencies": {
    "@repo/ui": "workspace:*"
  }
}

// apps/web/src/app/page.tsx
import { Button, Card, CardHeader, CardTitle, CardContent } from "@repo/ui";
// OR granular imports:
import { Button } from "@repo/ui/button";
```

## Shared Database Package

```json
// packages/db/package.json
{
  "name": "@repo/db",
  "private": true,
  "version": "0.0.0",
  "type": "module",
  "main": "./src/index.ts",
  "types": "./src/index.ts",
  "exports": {
    ".": "./src/index.ts",
    "./schema": "./src/schema.ts",
    "./client": "./src/client.ts"
  },
  "scripts": {
    "db:generate": "prisma generate",
    "db:migrate": "prisma migrate dev",
    "db:push": "prisma db push",
    "db:studio": "prisma studio",
    "typecheck": "tsc --noEmit"
  },
  "dependencies": {
    "@prisma/client": "^6.0"
  },
  "devDependencies": {
    "@repo/config-ts": "workspace:*",
    "prisma": "^6.0",
    "typescript": "^5.7"
  }
}
```

```typescript
// packages/db/src/client.ts
import { PrismaClient } from "@prisma/client";

const globalForPrisma = globalThis as unknown as {
  prisma: PrismaClient | undefined;
};

export const db =
  globalForPrisma.prisma ??
  new PrismaClient({
    log: process.env.NODE_ENV === "development" ? ["query", "warn", "error"] : ["error"],
  });

if (process.env.NODE_ENV !== "production") globalForPrisma.prisma = db;
```

```typescript
// packages/db/src/index.ts
export { db } from "./client";
export type { User, Post, Organization } from "@prisma/client";
// Re-export Prisma types for type-safe queries in apps
export { Prisma } from "@prisma/client";
```

```typescript
// Consuming in an app:
// apps/api/src/routes/users.ts
import { db, type User } from "@repo/db";

export async function getUsers(): Promise<User[]> {
  return db.user.findMany();
}
```

## Shared Utilities Package

```json
// packages/utils/package.json
{
  "name": "@repo/utils",
  "private": true,
  "version": "0.0.0",
  "type": "module",
  "main": "./src/index.ts",
  "types": "./src/index.ts",
  "exports": {
    ".": "./src/index.ts",
    "./date": "./src/date.ts",
    "./validation": "./src/validation.ts",
    "./errors": "./src/errors.ts"
  },
  "scripts": {
    "test": "vitest run",
    "typecheck": "tsc --noEmit"
  },
  "dependencies": {
    "zod": "^3.24"
  },
  "devDependencies": {
    "@repo/config-ts": "workspace:*",
    "vitest": "^3.0",
    "typescript": "^5.7"
  }
}
```

```typescript
// packages/utils/src/errors.ts
export class AppError extends Error {
  constructor(
    message: string,
    public readonly code: string,
    public readonly statusCode: number = 500,
    public readonly details?: Record<string, unknown>
  ) {
    super(message);
    this.name = "AppError";
  }

  static notFound(resource: string, id?: string) {
    return new AppError(
      id ? `${resource} not found: ${id}` : `${resource} not found`,
      "NOT_FOUND",
      404
    );
  }

  static unauthorized(message = "Authentication required") {
    return new AppError(message, "UNAUTHORIZED", 401);
  }

  static forbidden(message = "Insufficient permissions") {
    return new AppError(message, "FORBIDDEN", 403);
  }

  static badRequest(message: string, details?: Record<string, unknown>) {
    return new AppError(message, "BAD_REQUEST", 400, details);
  }
}
```

```typescript
// packages/utils/src/validation.ts
import { z } from "zod";

// Reusable schemas shared across apps
export const emailSchema = z.string().email().toLowerCase().trim();
export const passwordSchema = z.string().min(8).max(128);
export const slugSchema = z.string().regex(/^[a-z0-9-]+$/).min(1).max(100);
export const idSchema = z.string().cuid2();

export const paginationSchema = z.object({
  page: z.coerce.number().int().positive().default(1),
  limit: z.coerce.number().int().positive().max(100).default(20),
});

export type Pagination = z.infer<typeof paginationSchema>;
```

## Shared Types Package (Lightweight Alternative)

```json
// packages/types/package.json
{
  "name": "@repo/types",
  "private": true,
  "version": "0.0.0",
  "type": "module",
  "main": "./src/index.ts",
  "types": "./src/index.ts"
}
```

```typescript
// packages/types/src/index.ts
// API response wrapper — used by both API and frontend
export interface ApiResponse<T> {
  data: T;
  meta?: {
    page: number;
    limit: number;
    total: number;
    totalPages: number;
  };
}

export interface ApiError {
  error: {
    code: string;
    message: string;
    details?: Record<string, string[]>;
  };
}

// Shared domain types
export interface UserProfile {
  id: string;
  name: string;
  email: string;
  avatar?: string;
  role: "user" | "editor" | "admin";
}
```

## Internal Package Conventions

### package.json — Point to Source (Not Built)

```json
{
  "name": "@repo/ui",
  "private": true,
  "main": "./src/index.ts",
  "types": "./src/index.ts",
  "exports": {
    ".": "./src/index.ts"
  }
}
```

**Why source, not dist?** Internal packages consumed by apps that already have their own bundler (Next.js, Vite). The app's bundler transpiles the source directly. No need for a separate build step. This removes build order complexity and speeds up dev.

**When to use dist output instead:** When the package is consumed by something without a bundler (a Node.js script, a CLI tool) or when build times are too slow for the consuming app.

### Workspace Protocol

```json
// apps/web/package.json
{
  "dependencies": {
    "@repo/ui": "workspace:*",
    "@repo/db": "workspace:*",
    "@repo/utils": "workspace:*"
  },
  "devDependencies": {
    "@repo/config-ts": "workspace:*",
    "@repo/config-eslint": "workspace:*"
  }
}
```

`workspace:*` tells pnpm to always resolve to the local workspace package. During `pnpm publish`, it's replaced with the actual version. For internal packages (`"private": true`), the `*` version is fine.

## Gotchas

1. **Source imports need bundler support.** Pointing `main` to `./src/index.ts` (TypeScript source) only works if the consuming app has a bundler that handles TS. Next.js, Vite, and esbuild all do. Plain Node.js does not — use `tsx` or build to `dist/`.

2. **Tailwind CSS needs to scan package sources.** In the consuming app's Tailwind config, add the UI package to the content paths: `content: ["./src/**/*.tsx", "../../packages/ui/src/**/*.tsx"]`. Without this, utility classes from the shared UI package won't be included.

3. **`workspace:*` breaks outside the monorepo.** If you ever need to publish a package to npm, change to proper semver. Internal packages (`"private": true`) should always use `workspace:*`.

4. **Barrel exports can hurt tree-shaking.** A single `index.ts` that re-exports everything means importing one function loads the entire package. Use granular exports in `package.json` (`"./button"`, `"./card"`) for large packages.

5. **pnpm strict mode catches phantom deps.** If a package uses `react` but doesn't declare it in its own `package.json`, pnpm will catch it. This is a feature — fix it by adding the dependency. Don't disable strict mode.

6. **Hot reload across packages needs config.** Next.js: add `transpilePackages: ["@repo/ui"]` in `next.config.js`. Vite: configure `optimizeDeps.include` for linked packages. Without this, changes to shared packages may not trigger hot reload.
