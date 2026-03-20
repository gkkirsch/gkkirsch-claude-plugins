# Monorepo Cheatsheet

## pnpm Workspace Commands

```bash
# Run command in specific package
pnpm --filter <package-name> <command>
pnpm --filter @repo/ui build

# Run command in all packages
pnpm -r <command>
pnpm -r build

# Run command in all packages that match a pattern
pnpm --filter "./packages/*" build

# Run command in a package and its dependencies
pnpm --filter <package>... build

# Run command in a package and its dependents
pnpm --filter ...<package> build

# Add dependency to specific package
pnpm add <dep> --filter <package>

# Add workspace dependency
pnpm add @repo/utils --filter @repo/web --workspace

# Add root dev dependency
pnpm add -Dw <dep>

# Remove dependency
pnpm remove <dep> --filter <package>

# Install all dependencies
pnpm install

# Update lockfile without modifying node_modules
pnpm install --lockfile-only

# Deduplicate packages
pnpm dedupe
```

## pnpm-workspace.yaml

```yaml
packages:
  - "apps/*"
  - "packages/*"
  - "tooling/*"
```

## Turborepo Commands

```bash
# Run a pipeline task
turbo run build
turbo run build --filter=@repo/web

# Run with dependency awareness
turbo run build --filter=@repo/web...  # Package + its deps
turbo run test --filter=...@repo/ui    # Package + its dependents

# Run multiple tasks
turbo run build test lint

# Dry run (show what would execute)
turbo run build --dry-run

# Show task graph
turbo run build --graph

# Run with specific concurrency
turbo run build --concurrency=4

# Ignore cache
turbo run build --force

# Show cache status
turbo run build --summarize
```

## turbo.json Configuration

```json
{
  "$schema": "https://turbo.build/schema.json",
  "globalDependencies": ["**/.env.*local", ".env"],
  "globalEnv": ["NODE_ENV", "CI"],
  "tasks": {
    "build": {
      "dependsOn": ["^build"],
      "inputs": ["src/**", "tsconfig.json", "package.json"],
      "outputs": ["dist/**", ".next/**"],
      "env": ["DATABASE_URL"]
    },
    "dev": {
      "dependsOn": ["^build"],
      "cache": false,
      "persistent": true
    },
    "test": {
      "dependsOn": ["build"],
      "inputs": ["src/**", "test/**", "vitest.config.*"],
      "outputs": ["coverage/**"]
    },
    "lint": {
      "dependsOn": ["^build"],
      "inputs": ["src/**", ".eslintrc.*"]
    },
    "typecheck": {
      "dependsOn": ["^build"],
      "inputs": ["src/**", "tsconfig.json"]
    },
    "clean": {
      "cache": false
    }
  }
}
```

## Task Dependencies

| Syntax | Meaning |
|--------|---------|
| `"^build"` | Run `build` in all dependency packages first |
| `"build"` | Run `build` in the same package first |
| `"@repo/db#build"` | Run `build` in a specific package first |
| `[]` | No dependencies, can run immediately |

## Cache Configuration

| Field | Purpose | Example |
|-------|---------|---------|
| `inputs` | Files that affect this task | `["src/**", "tsconfig.json"]` |
| `outputs` | Files produced by this task | `["dist/**"]` |
| `cache` | Enable/disable caching | `false` for `dev` tasks |
| `persistent` | Long-running task (dev servers) | `true` for `dev` |
| `env` | Env vars that affect output | `["DATABASE_URL"]` |
| `passThroughEnv` | Env vars passed but don't affect cache | `["AWS_REGION"]` |

## Shared TypeScript Config

```json
// packages/config-ts/tsconfig.base.json
{
  "compilerOptions": {
    "strict": true,
    "target": "ES2022",
    "module": "ESNext",
    "moduleResolution": "bundler",
    "esModuleInterop": true,
    "skipLibCheck": true,
    "forceConsistentCasingInFileNames": true,
    "resolveJsonModule": true,
    "isolatedModules": true,
    "declaration": true,
    "declarationMap": true,
    "sourceMap": true
  }
}

// packages/config-ts/tsconfig.react.json
{
  "extends": "./tsconfig.base.json",
  "compilerOptions": {
    "jsx": "react-jsx",
    "lib": ["DOM", "DOM.Iterable", "ES2022"]
  }
}

// packages/config-ts/tsconfig.node.json
{
  "extends": "./tsconfig.base.json",
  "compilerOptions": {
    "module": "CommonJS",
    "moduleResolution": "node",
    "lib": ["ES2022"]
  }
}
```

## Shared ESLint Config

```javascript
// packages/config-eslint/base.js
module.exports = {
  extends: ["eslint:recommended", "prettier"],
  env: { node: true, es2022: true },
  parserOptions: { ecmaVersion: "latest", sourceType: "module" },
  rules: { "no-console": "warn" },
};

// packages/config-eslint/react.js
module.exports = {
  extends: [
    "./base.js",
    "plugin:react/recommended",
    "plugin:react-hooks/recommended",
  ],
  settings: { react: { version: "detect" } },
  rules: { "react/react-in-jsx-scope": "off" },
};
```

## Common Package Types

| Package | Purpose | Example Deps |
|---------|---------|-------------|
| `@repo/ui` | Shared React components | react, tailwind, class-variance-authority |
| `@repo/config-ts` | TypeScript configs | (none — just .json files) |
| `@repo/config-eslint` | ESLint configs | eslint, eslint-config-* |
| `@repo/db` | Database schema + client | prisma, @prisma/client |
| `@repo/utils` | Shared utilities | zod, date-fns |
| `@repo/types` | Shared TypeScript types | (none — just .d.ts files) |
| `@repo/email` | Email templates | react-email, @react-email/components |
| `@repo/auth` | Auth utilities | jose, bcrypt |

## CI/CD Pattern

```yaml
# .github/workflows/ci.yml
name: CI
on: [push, pull_request]
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 2  # Needed for turbo cache comparison

      - uses: pnpm/action-setup@v4
      - uses: actions/setup-node@v4
        with:
          node-version: 20
          cache: pnpm

      - run: pnpm install --frozen-lockfile

      - run: turbo run build test lint typecheck
        env:
          TURBO_TOKEN: ${{ secrets.TURBO_TOKEN }}
          TURBO_TEAM: ${{ vars.TURBO_TEAM }}
```

## Remote Caching

```bash
# Login to Vercel for remote cache
npx turbo login
npx turbo link

# Or set environment variables
export TURBO_TOKEN="your-token"
export TURBO_TEAM="your-team"

# Self-hosted cache server (ducktors/turborepo-remote-cache)
export TURBO_API="https://your-cache-server.com"
export TURBO_TOKEN="your-token"
export TURBO_TEAM="your-team"
```

## Troubleshooting

| Problem | Solution |
|---------|----------|
| `Cannot find module @repo/ui` | Run `turbo build` to build dependencies first |
| Task runs but cache never hits | Check `inputs` and `outputs` in turbo.json |
| `ERR_PNPM_PEER_DEP_ISSUES` | Add `peerDependencyRules.ignoreMissing` to .npmrc |
| Slow `pnpm install` | Use `--frozen-lockfile` in CI, `pnpm dedupe` locally |
| TypeScript can't resolve workspace package | Add `"references"` to tsconfig.json |
| Dev server doesn't pick up package changes | Use `"dev"` script with `--watch` in the package |
| `turbo prune` missing files | Check `outputs` includes all needed build artifacts |
| Cache invalidated too often | Remove volatile files from `inputs` (e.g., .env) |

## Package.json Workspace Protocol

```json
// apps/web/package.json
{
  "dependencies": {
    "@repo/ui": "workspace:*",     // Latest local version
    "@repo/utils": "workspace:^",  // Local version (caret range on publish)
    "@repo/db": "workspace:~"      // Local version (tilde range on publish)
  }
}
```

## Scaffold a New Package

```bash
# Quick package scaffold
mkdir -p packages/new-pkg/src
cat > packages/new-pkg/package.json << 'EOF'
{
  "name": "@repo/new-pkg",
  "version": "0.0.0",
  "private": true,
  "main": "./dist/index.js",
  "types": "./dist/index.d.ts",
  "scripts": {
    "build": "tsup src/index.ts --format cjs,esm --dts",
    "dev": "tsup src/index.ts --format cjs,esm --dts --watch"
  }
}
EOF
echo 'export {}' > packages/new-pkg/src/index.ts
cat > packages/new-pkg/tsconfig.json << 'EOF'
{ "extends": "@repo/config-ts/tsconfig.base.json", "include": ["src"] }
EOF
pnpm install
```
