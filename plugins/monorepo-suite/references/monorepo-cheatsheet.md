# Monorepo Cheatsheet

## Project Structure

```
my-monorepo/
  apps/
    web/          # Next.js
    api/          # Express/Fastify
    docs/         # Docs site
  packages/
    ui/           # Shared React components
    db/           # Prisma schema + client
    utils/        # Shared utilities
    types/        # Shared TypeScript types
    config-ts/    # tsconfig bases
    config-eslint/# ESLint configs
  turbo.json
  pnpm-workspace.yaml
  package.json
```

## Setup

```bash
pnpm dlx create-turbo@latest my-monorepo   # New monorepo
pnpm add -Dw turbo                          # Add to existing

# pnpm-workspace.yaml
packages:
  - "apps/*"
  - "packages/*"
```

## turbo.json Essentials

```json
{
  "tasks": {
    "build": {
      "dependsOn": ["^build"],
      "inputs": ["src/**"],
      "outputs": ["dist/**", ".next/**"],
      "env": ["NODE_ENV"]
    },
    "dev": { "persistent": true, "cache": false },
    "lint": { "dependsOn": ["^build"] },
    "test": { "dependsOn": ["build"] }
  },
  "globalEnv": ["CI"]
}
```

## CLI Quick Reference

```bash
pnpm turbo build                     # Build all
pnpm turbo build --filter=web        # Build one package
pnpm turbo build --filter=web...     # Package + all deps
pnpm turbo build --filter=...web     # Package + all dependents
pnpm turbo build --filter=./apps/*   # All apps
pnpm turbo build --filter=[HEAD~1]   # Changed since last commit
pnpm turbo build --force             # Skip cache
pnpm turbo build --dry-run           # Show what would run
pnpm turbo build --graph             # Dependency graph
pnpm turbo prune web --docker        # Prune for Docker
pnpm turbo watch build               # Watch mode
```

## Internal Package Pattern

```json
// packages/ui/package.json
{
  "name": "@repo/ui",
  "private": true,
  "main": "./src/index.ts",
  "types": "./src/index.ts",
  "exports": {
    ".": "./src/index.ts",
    "./button": "./src/components/button.tsx"
  }
}

// apps/web/package.json — consuming
{
  "dependencies": {
    "@repo/ui": "workspace:*"
  }
}
```

## Config Package Pattern

```json
// packages/config-ts/base.json
{
  "compilerOptions": {
    "strict": true,
    "moduleResolution": "bundler",
    "module": "esnext",
    "target": "es2022",
    "verbatimModuleSyntax": true
  }
}

// apps/web/tsconfig.json
{ "extends": "@repo/config-ts/nextjs.json" }
```

## Docker Prune Pattern

```dockerfile
FROM node:20-slim AS pruner
RUN corepack enable && pnpm add -g turbo
COPY . .
RUN turbo prune web --docker

FROM node:20-slim AS installer
COPY --from=pruner /app/out/json/ .    # package.jsons first (cache)
RUN pnpm install --frozen-lockfile
COPY --from=pruner /app/out/full/ .    # source code second
RUN pnpm turbo build --filter=web
```

## Remote Cache

```bash
pnpm turbo login && pnpm turbo link    # Vercel cache
# CI: set TURBO_TOKEN + TURBO_TEAM
```

## Changesets

```bash
pnpm add -Dw @changesets/cli
pnpm changeset init
pnpm changeset               # Create changeset
pnpm changeset version       # Apply: bump versions + changelogs
pnpm changeset publish       # Publish to npm
```

## GitHub Actions Template

```yaml
steps:
  - uses: actions/checkout@v4
    with: { fetch-depth: 2 }
  - uses: pnpm/action-setup@v4
  - uses: actions/setup-node@v4
    with: { node-version: 20, cache: "pnpm" }
  - run: pnpm install --frozen-lockfile
  - run: pnpm turbo build lint test typecheck
```

## Task Dependencies

```
^task  = run in DEPENDENCY packages first (topological)
task   = run in SAME package first
pkg#task = run specific package's task first
```

## Key Rules

- `workspace:*` for all internal deps
- `"private": true` for all internal packages
- Point `main` to source (not dist) when apps have bundlers
- Each package declares its own deps (no hoisting reliance)
- `persistent: true` + `cache: false` for dev servers
- Config packages: tsconfig, eslint, prettier — always worth extracting
- Tailwind content paths must include `../../packages/ui/src/**`
- Next.js: add `transpilePackages: ["@repo/ui"]`
