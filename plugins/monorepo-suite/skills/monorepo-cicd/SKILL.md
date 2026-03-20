---
name: monorepo-cicd
description: >
  CI/CD for monorepos — GitHub Actions with Turborepo, Docker builds with pruning,
  selective deployments, changesets for versioning, and Vercel/Fly.io deployment.
  Triggers: "monorepo ci", "monorepo deploy", "turborepo github actions", "monorepo docker",
  "changesets", "monorepo versioning", "turbo prune", "selective deployment".
  NOT for: Turborepo config (use turborepo-setup), package code (use workspace-packages).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Monorepo CI/CD

## GitHub Actions — Full Pipeline

```yaml
# .github/workflows/ci.yml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

env:
  TURBO_TOKEN: ${{ secrets.TURBO_TOKEN }}
  TURBO_TEAM: ${{ secrets.TURBO_TEAM }}

jobs:
  build:
    name: Build & Test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 2   # Needed for turbo change detection

      - uses: pnpm/action-setup@v4

      - uses: actions/setup-node@v4
        with:
          node-version: 20
          cache: "pnpm"

      - run: pnpm install --frozen-lockfile

      - name: Build
        run: pnpm turbo build

      - name: Lint
        run: pnpm turbo lint

      - name: Type Check
        run: pnpm turbo typecheck

      - name: Test
        run: pnpm turbo test
        env:
          DATABASE_URL_TEST: ${{ secrets.DATABASE_URL_TEST }}
```

## Selective CI — Only Build What Changed

```yaml
# .github/workflows/ci-selective.yml
name: Selective CI

on:
  pull_request:
    branches: [main]

jobs:
  detect-changes:
    runs-on: ubuntu-latest
    outputs:
      web: ${{ steps.filter.outputs.web }}
      api: ${{ steps.filter.outputs.api }}
      packages: ${{ steps.filter.outputs.packages }}
    steps:
      - uses: actions/checkout@v4
      - uses: dorny/paths-filter@v3
        id: filter
        with:
          filters: |
            web:
              - 'apps/web/**'
              - 'packages/ui/**'
              - 'packages/config-ts/**'
            api:
              - 'apps/api/**'
              - 'packages/db/**'
              - 'packages/utils/**'
            packages:
              - 'packages/**'

  build-web:
    needs: detect-changes
    if: needs.detect-changes.outputs.web == 'true'
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: pnpm/action-setup@v4
      - uses: actions/setup-node@v4
        with: { node-version: 20, cache: "pnpm" }
      - run: pnpm install --frozen-lockfile
      - run: pnpm turbo build --filter=web...
      - run: pnpm turbo test --filter=web

  build-api:
    needs: detect-changes
    if: needs.detect-changes.outputs.api == 'true'
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: pnpm/action-setup@v4
      - uses: actions/setup-node@v4
        with: { node-version: 20, cache: "pnpm" }
      - run: pnpm install --frozen-lockfile
      - run: pnpm turbo build --filter=api...
      - run: pnpm turbo test --filter=api
```

## Docker Builds with Turbo Prune

```dockerfile
# apps/web/Dockerfile
FROM node:20-slim AS base
ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"
RUN corepack enable

# Stage 1: Prune the monorepo
FROM base AS pruner
WORKDIR /app
RUN pnpm add -g turbo
COPY . .
RUN turbo prune web --docker

# Stage 2: Install dependencies
FROM base AS installer
WORKDIR /app

# First install dependencies (for better caching)
COPY --from=pruner /app/out/json/ .
RUN pnpm install --frozen-lockfile

# Then copy source and build
COPY --from=pruner /app/out/full/ .
RUN pnpm turbo build --filter=web

# Stage 3: Production image
FROM base AS runner
WORKDIR /app
ENV NODE_ENV=production

RUN addgroup --system --gid 1001 nodejs
RUN adduser --system --uid 1001 nextjs

# Copy built app
COPY --from=installer /app/apps/web/.next/standalone ./
COPY --from=installer /app/apps/web/.next/static ./apps/web/.next/static
COPY --from=installer /app/apps/web/public ./apps/web/public

USER nextjs
EXPOSE 3000
ENV PORT=3000
CMD ["node", "apps/web/server.js"]
```

```dockerfile
# apps/api/Dockerfile
FROM node:20-slim AS base
ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"
RUN corepack enable

FROM base AS pruner
WORKDIR /app
RUN pnpm add -g turbo
COPY . .
RUN turbo prune api --docker

FROM base AS installer
WORKDIR /app
COPY --from=pruner /app/out/json/ .
RUN pnpm install --frozen-lockfile
COPY --from=pruner /app/out/full/ .

# Generate Prisma Client
RUN pnpm --filter=@repo/db exec prisma generate
RUN pnpm turbo build --filter=api

FROM base AS runner
WORKDIR /app
ENV NODE_ENV=production

COPY --from=installer /app/apps/api/dist ./dist
COPY --from=installer /app/apps/api/package.json ./
COPY --from=installer /app/node_modules ./node_modules
COPY --from=installer /app/packages/db/prisma ./prisma

EXPOSE 3001
CMD ["sh", "-c", "npx prisma migrate deploy && node dist/server.js"]
```

## Changesets — Versioning & Publishing

```bash
# Install changesets
pnpm add -Dw @changesets/cli @changesets/changelog-github
pnpm changeset init
```

```json
// .changeset/config.json
{
  "$schema": "https://unpkg.com/@changesets/config@3.0.0/schema.json",
  "changelog": [
    "@changesets/changelog-github",
    { "repo": "your-org/your-repo" }
  ],
  "commit": false,
  "fixed": [],
  "linked": [],
  "access": "restricted",
  "baseBranch": "main",
  "updateInternalDependencies": "patch",
  "ignore": []
}
```

```bash
# Developer workflow
pnpm changeset                    # Interactive: select packages, bump type, description
pnpm changeset version            # Apply changesets: bump versions, update changelogs
pnpm changeset publish            # Publish to npm (for public packages)

# For internal monorepos (no npm publish):
pnpm changeset version            # Still useful for changelogs and version tracking
git add . && git commit -m "Version packages"
```

```yaml
# .github/workflows/release.yml
name: Release

on:
  push:
    branches: [main]

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: pnpm/action-setup@v4
      - uses: actions/setup-node@v4
        with:
          node-version: 20
          cache: "pnpm"

      - run: pnpm install --frozen-lockfile

      - name: Create Release PR or Publish
        uses: changesets/action@v1
        with:
          version: pnpm changeset version
          publish: pnpm changeset publish
          title: "chore: version packages"
          commit: "chore: version packages"
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          NPM_TOKEN: ${{ secrets.NPM_TOKEN }}
```

## Vercel Deployment

```json
// apps/web/vercel.json (optional — most config is in dashboard)
{
  "framework": "nextjs",
  "installCommand": "pnpm install --frozen-lockfile",
  "buildCommand": "pnpm turbo build --filter=web"
}
```

```bash
# Vercel project settings (set in dashboard):
# Root Directory: apps/web
# Build Command: cd ../.. && pnpm turbo build --filter=web
# Output Directory: apps/web/.next
# Install Command: pnpm install --frozen-lockfile

# Vercel automatically detects monorepo and installs from root.
# Environment variables: set in Vercel dashboard per project.
```

## Deploy Pipeline — Multiple Apps

```yaml
# .github/workflows/deploy.yml
name: Deploy

on:
  push:
    branches: [main]

jobs:
  detect-changes:
    runs-on: ubuntu-latest
    outputs:
      web: ${{ steps.filter.outputs.web }}
      api: ${{ steps.filter.outputs.api }}
    steps:
      - uses: actions/checkout@v4
        with: { fetch-depth: 2 }
      - uses: dorny/paths-filter@v3
        id: filter
        with:
          filters: |
            web:
              - 'apps/web/**'
              - 'packages/**'
            api:
              - 'apps/api/**'
              - 'packages/**'

  deploy-web:
    needs: [detect-changes]
    if: needs.detect-changes.outputs.web == 'true'
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: pnpm/action-setup@v4
      - uses: actions/setup-node@v4
        with: { node-version: 20, cache: "pnpm" }
      - run: pnpm install --frozen-lockfile
      - run: pnpm turbo build --filter=web
      - name: Deploy to Vercel
        run: pnpm vercel deploy --prod --token=${{ secrets.VERCEL_TOKEN }}
        working-directory: apps/web

  deploy-api:
    needs: [detect-changes]
    if: needs.detect-changes.outputs.api == 'true'
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: pnpm/action-setup@v4
      - uses: actions/setup-node@v4
        with: { node-version: 20, cache: "pnpm" }
      - run: pnpm install --frozen-lockfile

      - name: Build and push Docker image
        run: |
          docker build -f apps/api/Dockerfile -t api:${{ github.sha }} .
          # Push to your registry (ECR, GCR, GHCR, etc.)

      - name: Deploy to Fly.io
        uses: superfly/flyctl-actions/setup-flyctl@master
      - run: flyctl deploy --image api:${{ github.sha }}
        working-directory: apps/api
        env:
          FLY_API_TOKEN: ${{ secrets.FLY_API_TOKEN }}
```

## Database Migrations in CI

```yaml
# Migration job — runs before app deployments
migrate-db:
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - uses: pnpm/action-setup@v4
    - uses: actions/setup-node@v4
      with: { node-version: 20, cache: "pnpm" }
    - run: pnpm install --frozen-lockfile

    - name: Generate Prisma Client
      run: pnpm --filter=@repo/db exec prisma generate

    - name: Apply Migrations
      run: pnpm --filter=@repo/db exec prisma migrate deploy
      env:
        DATABASE_URL: ${{ secrets.DATABASE_URL }}
```

## Monorepo Package Scripts

```json
// Root package.json — convenience scripts
{
  "scripts": {
    "dev": "turbo dev",
    "dev:web": "turbo dev --filter=web",
    "dev:api": "turbo dev --filter=api",
    "build": "turbo build",
    "build:web": "turbo build --filter=web...",
    "build:api": "turbo build --filter=api...",
    "test": "turbo test",
    "test:web": "turbo test --filter=web",
    "lint": "turbo lint",
    "typecheck": "turbo typecheck",
    "clean": "turbo clean && rm -rf node_modules",
    "db:generate": "pnpm --filter=@repo/db exec prisma generate",
    "db:migrate": "pnpm --filter=@repo/db exec prisma migrate dev",
    "db:studio": "pnpm --filter=@repo/db exec prisma studio",
    "format": "prettier --write \"**/*.{ts,tsx,js,jsx,json,md}\"",
    "changeset": "changeset",
    "version-packages": "changeset version",
    "release": "changeset publish"
  }
}
```

## Gotchas

1. **`fetch-depth: 2` is required for change detection.** GitHub Actions checks out only the latest commit by default. Turbo and path filters need at least 2 commits to detect changes. For changesets, you may need `fetch-depth: 0` (full history).

2. **Docker COPY invalidates cache for the entire layer.** In the pruned Dockerfile, copy `out/json/` first (package.json files), install deps, THEN copy `out/full/` (source). This way, `pnpm install` is cached unless dependencies change.

3. **Vercel auto-detects but doesn't auto-scope.** When deploying a monorepo on Vercel, set the Root Directory to the app directory. Vercel installs from the workspace root but builds from the specified root directory.

4. **Changesets need `updateInternalDependencies`.** When package A bumps, and package B depends on A, `updateInternalDependencies: "patch"` auto-bumps B too. Without this, published packages can have stale internal dependency versions.

5. **pnpm `--frozen-lockfile` fails on CI if lockfile is outdated.** This is intentional — it prevents drift between local and CI. Fix by running `pnpm install` locally and committing the updated `pnpm-lock.yaml`.

6. **Turbo remote cache tokens are per-team.** If you have multiple Vercel teams, each needs its own `TURBO_TOKEN` and `TURBO_TEAM`. Misconfigured tokens result in cache misses, not errors — hard to debug.
