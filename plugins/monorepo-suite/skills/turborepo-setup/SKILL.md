---
name: turborepo-setup
description: >
  Turborepo setup and pipeline configuration — turbo.json, task pipelines,
  caching, remote cache, environment variables, filtering, and watch mode.
  Triggers: "turborepo", "turbo.json", "turbo pipeline", "turbo cache",
  "monorepo setup", "pnpm workspace", "workspace setup".
  NOT for: individual package code (use workspace-packages), CI/CD pipelines (use monorepo-cicd).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Turborepo Setup & Configuration

## Initial Setup

```bash
# New monorepo from scratch
pnpm dlx create-turbo@latest my-monorepo
cd my-monorepo

# Add Turborepo to existing pnpm workspace
pnpm add -Dw turbo

# Verify
pnpm turbo --version
```

## Workspace Configuration

```yaml
# pnpm-workspace.yaml
packages:
  - "apps/*"
  - "packages/*"
  - "tooling/*"    # Optional: build tools, scripts
```

```json
// Root package.json
{
  "name": "my-monorepo",
  "private": true,
  "scripts": {
    "dev": "turbo dev",
    "build": "turbo build",
    "lint": "turbo lint",
    "test": "turbo test",
    "typecheck": "turbo typecheck",
    "clean": "turbo clean",
    "format": "prettier --write \"**/*.{ts,tsx,js,jsx,json,md}\""
  },
  "devDependencies": {
    "turbo": "^2.4",
    "prettier": "^3.4",
    "typescript": "^5.7"
  },
  "packageManager": "pnpm@9.15.0"
}
```

## turbo.json — Pipeline Configuration

```json
{
  "$schema": "https://turbo.build/schema.json",
  "ui": "tui",
  "tasks": {
    "build": {
      "dependsOn": ["^build"],
      "inputs": ["src/**", "tsconfig.json", "package.json"],
      "outputs": ["dist/**", ".next/**", "!.next/cache/**"],
      "env": ["NODE_ENV", "DATABASE_URL"]
    },
    "dev": {
      "dependsOn": ["^build"],
      "persistent": true,
      "cache": false
    },
    "lint": {
      "dependsOn": ["^build"],
      "inputs": ["src/**", ".eslintrc.*", "eslint.config.*"]
    },
    "test": {
      "dependsOn": ["build"],
      "inputs": ["src/**", "tests/**", "vitest.config.*"],
      "outputs": ["coverage/**"],
      "env": ["CI", "DATABASE_URL_TEST"]
    },
    "typecheck": {
      "dependsOn": ["^build"],
      "inputs": ["src/**", "tsconfig.json"]
    },
    "clean": {
      "cache": false
    },
    "db:generate": {
      "cache": false
    },
    "db:migrate": {
      "cache": false,
      "dependsOn": ["db:generate"]
    }
  },
  "globalDependencies": [
    ".env",
    "tsconfig.base.json"
  ],
  "globalEnv": [
    "CI",
    "NODE_ENV"
  ]
}
```

## Key Concepts

### Task Dependencies

```json
{
  "tasks": {
    "build": {
      // ^build = build dependencies first (topological)
      "dependsOn": ["^build"]
    },
    "test": {
      // build = build THIS package first (same package)
      "dependsOn": ["build"]
    },
    "deploy": {
      // Run specific package tasks first
      "dependsOn": ["build", "test", "lint"]
    }
  }
}
```

- `^task` — run task in all DEPENDENCY packages first (topological order)
- `task` — run task in the SAME package first
- `package#task` — run specific package's task first

### Caching

```json
{
  "tasks": {
    "build": {
      // Files that affect the output (cache key inputs)
      "inputs": [
        "src/**/*.ts",
        "src/**/*.tsx",
        "tsconfig.json",
        "package.json"
      ],
      // Files produced by the task (restored from cache)
      "outputs": [
        "dist/**",
        ".next/**",
        "!.next/cache/**"   // Exclude from cache
      ],
      // Env vars that affect the output (part of cache key)
      "env": ["NODE_ENV", "API_URL"],
      // Env vars included in hash but value not checked
      "passThroughEnv": ["AWS_SECRET_KEY"]
    },
    "dev": {
      "cache": false,        // Never cache dev server
      "persistent": true     // Long-running process
    }
  }
}
```

### Environment Variables

```json
{
  // Global env vars — affect ALL task hashes
  "globalEnv": ["CI", "VERCEL"],

  // Global file dependencies — changes invalidate ALL caches
  "globalDependencies": [".env", "tsconfig.base.json"],

  "tasks": {
    "build": {
      // Task-specific env vars
      "env": ["DATABASE_URL", "API_KEY"],

      // Pass through without hashing (secrets)
      "passThroughEnv": ["AWS_SECRET_ACCESS_KEY"]
    }
  }
}
```

## CLI Commands

```bash
# Run all tasks
pnpm turbo build
pnpm turbo lint test typecheck

# Filter by package
pnpm turbo build --filter=web
pnpm turbo build --filter=@repo/ui
pnpm turbo dev --filter=web...         # web + all its deps
pnpm turbo build --filter=...web       # web + all its dependents
pnpm turbo build --filter=./apps/*     # all apps
pnpm turbo build --filter=[HEAD~1]     # changed since last commit

# Combine filters
pnpm turbo build --filter=web --filter=api

# Watch mode (re-run on file changes)
pnpm turbo watch build --filter=@repo/ui

# Dry run (show what would run)
pnpm turbo build --dry-run

# Graph visualization
pnpm turbo build --graph           # Open in browser
pnpm turbo build --graph=graph.svg # Save to file

# Cache management
pnpm turbo build --force           # Ignore cache
pnpm turbo build --summarize       # Show cache hit/miss stats
pnpm turbo prune --scope=web       # Create minimal monorepo for deployment
```

## Remote Caching

```bash
# Vercel Remote Cache (recommended)
pnpm turbo login
pnpm turbo link

# Self-hosted (S3, GCS, Azure Blob)
# turbo.json:
{
  "remoteCache": {
    "enabled": true,
    "signature": true   // Verify cache integrity
  }
}

# Environment variables for CI:
# TURBO_TOKEN=your-token
# TURBO_TEAM=your-team
# TURBO_REMOTE_CACHE_SIGNATURE_KEY=your-key (if signature: true)
```

## Package-Specific turbo.json

```json
// apps/web/turbo.json — extends root
{
  "extends": ["//"],
  "tasks": {
    "build": {
      "outputs": [".next/**", "!.next/cache/**"],
      "env": ["NEXT_PUBLIC_API_URL", "NEXT_PUBLIC_ANALYTICS_ID"]
    },
    "dev": {
      "persistent": true,
      "cache": false
    }
  }
}
```

## Pruned Deployments

```bash
# Create a pruned monorepo with only what web needs
pnpm turbo prune web --out-dir=./out

# The output includes:
# out/
#   package.json        # Root with only relevant deps
#   pnpm-lock.yaml      # Pruned lockfile
#   pnpm-workspace.yaml # Only relevant packages
#   apps/web/           # The target app
#   packages/ui/        # Only deps of web
#   packages/db/        # Only deps of web

# Use in Dockerfile:
# COPY --from=pruner /app/out/json/ .
# RUN pnpm install --frozen-lockfile
# COPY --from=pruner /app/out/full/ .
# RUN pnpm turbo build --filter=web
```

## Gotchas

1. **`^build` means dependencies, not dependents.** `"dependsOn": ["^build"]` builds packages that THIS package depends on. To build packages that depend on this one, use `...package` filter syntax.

2. **Cache misses from env vars.** If builds randomly miss cache, check if an unlisted env var is changing. Add it to `env` or `globalEnv`. Common culprits: `CI`, `VERCEL`, `NODE_ENV`, timestamps in builds.

3. **`persistent: true` tasks never cache.** Dev servers, watch modes, and other long-running tasks must be marked `persistent: true` and `cache: false`. They also block other tasks from completing.

4. **pnpm hoisting can hide missing deps.** A package might work locally because pnpm hoisted a dependency it didn't declare. Use `pnpm --filter=web exec -- node -e "require('react')"` to verify a dep is properly declared. Pruned deployments will fail if deps are missing.

5. **`turbo prune` doesn't run install.** It creates the pruned workspace structure, but you still need to `pnpm install` in the output directory. This catches missing declared dependencies.

6. **Root scripts bypass Turborepo.** Running `pnpm -w run format` skips turbo entirely. Only scripts routed through `turbo` (like `pnpm turbo build`) get caching and parallelization.
