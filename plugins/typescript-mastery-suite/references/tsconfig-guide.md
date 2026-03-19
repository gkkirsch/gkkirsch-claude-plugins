# tsconfig.json Reference Guide

Complete guide to every important tsconfig option with recommended configurations for different project types.

---

## Top-Level Options

```jsonc
{
  // Compiler options — the main configuration
  "compilerOptions": { ... },

  // Files to include (glob patterns)
  "include": ["src/**/*"],

  // Files to exclude (glob patterns)
  "exclude": ["node_modules", "dist", "**/*.test.ts"],

  // Explicit file list (overrides include/exclude)
  "files": ["src/index.ts"],

  // Extend another config
  "extends": "./tsconfig.base.json",

  // Project references for monorepos
  "references": [
    { "path": "./packages/shared" },
    { "path": "./packages/api" }
  ]
}
```

---

## Compiler Options by Category

### Type Checking

```jsonc
{
  "compilerOptions": {
    // Master strict switch — enables all strict flags below
    "strict": true,

    // === Flags enabled by strict: true ===

    // Error on implicit any types
    "noImplicitAny": true,

    // null and undefined are distinct types
    "strictNullChecks": true,

    // Contravariant function parameter checking
    "strictFunctionTypes": true,

    // Type-check bind, call, apply
    "strictBindCallApply": true,

    // Class properties must be initialized
    "strictPropertyInitialization": true,

    // Error on implicit this: any
    "noImplicitThis": true,

    // catch variables are unknown (not any)
    "useUnknownInCatchVariables": true,

    // Emit "use strict" in output
    "alwaysStrict": true,

    // === Additional strictness (NOT part of strict) ===

    // Index signatures include undefined
    // HIGHLY RECOMMENDED
    "noUncheckedIndexedAccess": true,

    // Distinguishes undefined from missing properties
    "exactOptionalProperties": true,

    // Error on missing return statements
    "noImplicitReturns": true,

    // Error on fallthrough in switch
    "noFallthroughCasesInSwitch": true,

    // Require override keyword for overridden methods
    "noImplicitOverride": true,

    // Require bracket notation for index signatures
    "noPropertyAccessFromIndexSignature": true,

    // Error on case sensitivity mismatches
    "forceConsistentCasingInFileNames": true,

    // Mark unused locals as errors
    "noUnusedLocals": true,

    // Mark unused parameters as errors
    "noUnusedParameters": true
  }
}
```

### Module System

```jsonc
{
  "compilerOptions": {
    // What module system to emit
    // "ESNext" — ES modules (import/export)
    // "CommonJS" — require/module.exports
    // "NodeNext" — Node.js ESM with package.json "type" awareness
    // "Preserve" (5.4+) — keep import/export as-is
    "module": "ESNext",

    // How to resolve import paths
    // "bundler" — modern bundler resolution (Vite, esbuild, webpack)
    // "node16" / "nodenext" — Node.js ESM resolution
    // "node10" — legacy Node.js resolution (CJS only)
    "moduleResolution": "bundler",

    // Allow .ts extensions in imports (requires noEmit or emitDeclarationOnly)
    "allowImportingTsExtensions": true,

    // Enforce type-only import/export syntax
    // import type { X } from "..." required for types
    "verbatimModuleSyntax": true,

    // Allow default import from modules without default export
    "esModuleInterop": true,

    // Allow importing .json files
    "resolveJsonModule": true,

    // Allow importing .js files
    "allowJs": true,

    // Type-check .js files
    "checkJs": false,

    // Import path aliases
    "baseUrl": ".",
    "paths": {
      "@/*": ["./src/*"],
      "@components/*": ["./src/components/*"]
    },

    // Custom type roots (default: node_modules/@types)
    "typeRoots": ["./types", "./node_modules/@types"],

    // Specific type packages to include
    "types": ["node", "vitest/globals"]
  }
}
```

### Emit Options

```jsonc
{
  "compilerOptions": {
    // Output directory
    "outDir": "./dist",

    // Source root directory
    "rootDir": "./src",

    // Don't emit output files (type-check only)
    "noEmit": true,

    // Only emit .d.ts files (let bundler handle JS)
    "emitDeclarationOnly": true,

    // Generate .d.ts declaration files
    "declaration": true,

    // Generate .d.ts.map files (for go-to-definition)
    "declarationMap": true,

    // Output directory for declarations
    "declarationDir": "./dist/types",

    // Generate source maps
    "sourceMap": true,

    // Inline source maps in output files
    "inlineSourceMap": false,

    // Include source code in source maps
    "inlineSources": true,

    // JS compilation target
    // ES2022 — modern (class fields, top-level await, etc.)
    // ES2020 — good compatibility
    // ES2015 — legacy browser support
    "target": "ES2022",

    // Library type definitions to include
    "lib": ["ES2022", "DOM", "DOM.Iterable"],

    // JSX handling
    // "react-jsx" — React 17+ automatic runtime
    // "react" — React.createElement (legacy)
    // "preserve" — keep JSX as-is (for another tool)
    "jsx": "react-jsx",

    // Remove comments from output
    "removeComments": false,

    // Import helpers from tslib instead of inlining
    "importHelpers": true,

    // Generate helper code for for...of on older targets
    "downlevelIteration": false,

    // Preserve const enum declarations
    "preserveConstEnums": false,

    // Strip internal declarations
    "stripInternal": true
  }
}
```

### Performance Options

```jsonc
{
  "compilerOptions": {
    // Enable incremental compilation
    "incremental": true,

    // Location of incremental cache
    "tsBuildInfoFile": "./dist/.tsbuildinfo",

    // Skip type-checking .d.ts files
    // RECOMMENDED: almost always enable
    "skipLibCheck": true,

    // Each file can be transpiled independently
    // Required for: esbuild, swc, Vite, Babel
    "isolatedModules": true,

    // Enable parallel .d.ts generation (5.5+)
    "isolatedDeclarations": true,

    // Required for project references
    "composite": true,

    // Disable size limit on generated files
    "disableSizeLimit": false
  }
}
```

---

## Recommended Configurations

### Node.js API Server

```jsonc
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "NodeNext",
    "moduleResolution": "nodenext",
    "outDir": "./dist",
    "rootDir": "./src",

    "strict": true,
    "noUncheckedIndexedAccess": true,
    "forceConsistentCasingInFileNames": true,
    "noImplicitReturns": true,

    "declaration": true,
    "declarationMap": true,
    "sourceMap": true,
    "incremental": true,
    "tsBuildInfoFile": "./dist/.tsbuildinfo",
    "skipLibCheck": true,
    "isolatedModules": true,

    "esModuleInterop": true,
    "resolveJsonModule": true,
    "verbatimModuleSyntax": true,

    "lib": ["ES2022"],
    "types": ["node"]
  },
  "include": ["src/**/*"],
  "exclude": ["node_modules", "dist", "**/*.test.ts"]
}
```

### React (Vite) Application

```jsonc
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "ESNext",
    "moduleResolution": "bundler",
    "noEmit": true,

    "strict": true,
    "noUncheckedIndexedAccess": true,
    "forceConsistentCasingInFileNames": true,

    "skipLibCheck": true,
    "isolatedModules": true,

    "jsx": "react-jsx",
    "lib": ["ES2022", "DOM", "DOM.Iterable"],

    "esModuleInterop": true,
    "resolveJsonModule": true,
    "verbatimModuleSyntax": true,
    "allowImportingTsExtensions": true,

    "baseUrl": ".",
    "paths": {
      "@/*": ["./src/*"]
    }
  },
  "include": ["src/**/*", "vite-env.d.ts"],
  "exclude": ["node_modules"]
}
```

### Next.js Application

```jsonc
{
  "compilerOptions": {
    "target": "ES2017",
    "module": "ESNext",
    "moduleResolution": "bundler",
    "noEmit": true,

    "strict": true,
    "noUncheckedIndexedAccess": true,
    "forceConsistentCasingInFileNames": true,

    "skipLibCheck": true,
    "isolatedModules": true,

    "jsx": "preserve",
    "lib": ["DOM", "DOM.Iterable", "ES2022"],

    "esModuleInterop": true,
    "resolveJsonModule": true,
    "allowJs": true,
    "incremental": true,

    "plugins": [{ "name": "next" }],

    "baseUrl": ".",
    "paths": {
      "@/*": ["./src/*"]
    }
  },
  "include": ["next-env.d.ts", "**/*.ts", "**/*.tsx", ".next/types/**/*.ts"],
  "exclude": ["node_modules"]
}
```

### Library / npm Package

```jsonc
{
  "compilerOptions": {
    "target": "ES2020",
    "module": "ESNext",
    "moduleResolution": "bundler",

    "outDir": "./dist",
    "rootDir": "./src",
    "declaration": true,
    "declarationMap": true,
    "sourceMap": true,

    "strict": true,
    "noUncheckedIndexedAccess": true,
    "forceConsistentCasingInFileNames": true,

    "skipLibCheck": true,
    "isolatedModules": true,
    "isolatedDeclarations": true,

    "esModuleInterop": true,
    "verbatimModuleSyntax": true,

    "stripInternal": true,
    "importHelpers": true
  },
  "include": ["src/**/*"],
  "exclude": ["node_modules", "dist", "**/*.test.ts", "**/*.spec.ts"]
}

// Corresponding package.json:
// {
//   "type": "module",
//   "exports": {
//     ".": {
//       "types": "./dist/index.d.ts",
//       "import": "./dist/index.js"
//     }
//   },
//   "files": ["dist"],
//   "peerDependencies": { "typescript": ">=5.0" }
// }
```

### Monorepo Root

```jsonc
// tsconfig.base.json — shared settings
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "ESNext",
    "moduleResolution": "bundler",

    "strict": true,
    "noUncheckedIndexedAccess": true,
    "forceConsistentCasingInFileNames": true,

    "skipLibCheck": true,
    "isolatedModules": true,

    "esModuleInterop": true,
    "resolveJsonModule": true,
    "verbatimModuleSyntax": true,

    "declaration": true,
    "declarationMap": true,
    "composite": true,
    "incremental": true
  }
}

// tsconfig.json — project references
{
  "files": [],
  "references": [
    { "path": "./packages/shared" },
    { "path": "./packages/api" },
    { "path": "./packages/web" }
  ]
}

// packages/shared/tsconfig.json
{
  "extends": "../../tsconfig.base.json",
  "compilerOptions": {
    "outDir": "./dist",
    "rootDir": "./src",
    "tsBuildInfoFile": "./dist/.tsbuildinfo"
  },
  "include": ["src/**/*"]
}
```

### Test Configuration

```jsonc
// tsconfig.test.json
{
  "extends": "./tsconfig.json",
  "compilerOptions": {
    "noEmit": true,
    "types": ["vitest/globals", "@testing-library/jest-dom"]
  },
  "include": [
    "src/**/*",
    "tests/**/*",
    "**/*.test.ts",
    "**/*.test.tsx",
    "**/*.spec.ts"
  ]
}
```

---

## Common Patterns

### Extending Configs

```jsonc
// tsconfig.base.json — shared across projects
{
  "compilerOptions": {
    "strict": true,
    "skipLibCheck": true,
    "isolatedModules": true
  }
}

// tsconfig.json — extends base
{
  "extends": "./tsconfig.base.json",
  "compilerOptions": {
    "outDir": "./dist"
    // Inherits strict, skipLibCheck, isolatedModules
  }
}

// Can also extend from npm packages
{
  "extends": "@tsconfig/node20/tsconfig.json"
  // Popular bases: @tsconfig/node20, @tsconfig/strictest, @tsconfig/vite-react
}
```

### Multiple Build Targets

```jsonc
// tsconfig.json — type checking (editor, CI)
{
  "compilerOptions": {
    "noEmit": true
  }
}

// tsconfig.build.json — production build
{
  "extends": "./tsconfig.json",
  "compilerOptions": {
    "noEmit": false,
    "outDir": "./dist",
    "declaration": true
  },
  "exclude": ["**/*.test.ts"]
}
```

```bash
# Type check
tsc --noEmit

# Build
tsc --project tsconfig.build.json
```

---

## Flag Decision Matrix

| Question | Recommended Flag |
|----------|-----------------|
| New project? | `strict: true` |
| Migrating from JS? | Start with `allowJs: true, strict: false` |
| Using Vite/esbuild/swc? | `isolatedModules: true, noEmit: true` |
| Building a library? | `declaration: true, declarationMap: true, stripInternal: true` |
| Monorepo? | `composite: true, references: [...]` |
| Slow builds? | `incremental: true, skipLibCheck: true` |
| Server-side only? | Remove `DOM` from `lib` |
| Using path aliases? | `baseUrl: ".", paths: {...}` |
| Want strict null checks without full strict? | `strictNullChecks: true` individually |
| CI only type-check? | `noEmit: true` in main config |
| Need source maps? | `sourceMap: true` or `inlineSourceMap: true` |

---

## Troubleshooting

### "Cannot find module"

```jsonc
// Check moduleResolution matches your bundler
"moduleResolution": "bundler"  // For Vite, esbuild, webpack
"moduleResolution": "nodenext" // For Node.js ESM
"moduleResolution": "node10"   // For legacy Node.js CJS

// Check paths configuration
"baseUrl": ".",
"paths": { "@/*": ["./src/*"] }

// Check types are installed
// npm install --save-dev @types/express
```

### "implicitly has an 'any' type"

```jsonc
// Either annotate the type or disable the check
"noImplicitAny": false  // During migration only

// Or install type declarations
// npm install --save-dev @types/package-name
```

### "not assignable to type"

Common with strictNullChecks — value might be null/undefined:
```typescript
// Add null check
const el = document.getElementById("app");
if (!el) throw new Error("Missing element");
el.innerHTML = "hello"; // Now guaranteed non-null
```

### Build too slow

```jsonc
{
  "compilerOptions": {
    "incremental": true,
    "skipLibCheck": true,
    "isolatedModules": true
  },
  "exclude": ["**/*.test.ts", "**/*.spec.ts"]
}
```
