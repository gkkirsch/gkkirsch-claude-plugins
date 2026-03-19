# Monorepo Tools Comparison Reference

Comprehensive comparison of Nx, Turborepo, Lerna, and native workspace tools for
JavaScript/TypeScript monorepo management.

---

## Table of Contents

- [Quick Comparison](#quick-comparison)
- [Turborepo](#turborepo)
- [Nx](#nx)
- [Lerna](#lerna)
- [Package Manager Workspaces](#package-manager-workspaces)
- [Bazel and Pants](#bazel-and-pants)
- [Feature Matrix](#feature-matrix)
- [Performance Benchmarks](#performance-benchmarks)
- [Decision Guide](#decision-guide)
- [Migration Paths](#migration-paths)
- [Common Configurations](#common-configurations)

---

## Quick Comparison

| Tool | Philosophy | Complexity | Best For |
|------|-----------|-----------|----------|
| **Turborepo** | Simple, fast, convention-based | Low | Most JS/TS monorepos |
| **Nx** | Comprehensive, plugin-based | High | Large teams, enterprise |
| **Lerna** | Publishing-focused, now Nx-powered | Medium | npm package publishing |
| **pnpm workspaces** | Minimal, strict dependencies | Low | Small-medium monorepos |
| **Yarn workspaces** | PnP support, Yarn ecosystem | Low | Existing Yarn projects |
| **npm workspaces** | Zero-config, built-in | Low | Simple monorepos |
| **Bazel** | Hermetic, multi-language | Very High | Google-scale, multi-lang |
| **Pants** | Python-first, modern | High | Python-heavy projects |

---

## Turborepo

### Overview

Build system for JavaScript/TypeScript monorepos by Vercel. Focuses on speed
through intelligent caching and parallel execution. Minimal configuration.

### Key Features

- **Incremental builds** — Only rebuild what changed
- **Content-aware hashing** — Cache based on file content, not timestamps
- **Parallel execution** — Run tasks across packages concurrently
- **Remote caching** — Share cache across team and CI via Vercel
- **Task pipelines** — Define task dependencies declaratively
- **Pruning** — Generate minimal Docker contexts per package
- **Watch mode** — Efficient dev mode for changed packages

### Configuration (turbo.json)

```json
{
  "$schema": "https://turbo.build/schema.json",
  "tasks": {
    "build": {
      "dependsOn": ["^build"],
      "inputs": ["src/**", "tsconfig.json"],
      "outputs": ["dist/**", ".next/**"],
      "cache": true
    },
    "test": {
      "dependsOn": ["^build"],
      "inputs": ["src/**", "tests/**"],
      "outputs": ["coverage/**"],
      "cache": true
    },
    "lint": {
      "inputs": ["src/**", ".eslintrc.*"],
      "outputs": [],
      "cache": true
    },
    "dev": {
      "cache": false,
      "persistent": true
    }
  }
}
```

### Commands

```bash
# Run task across all packages
turbo run build

# Run task for specific package + dependencies
turbo run build --filter=@myorg/web...

# Run task for affected packages (since main)
turbo run test --filter=...[main]

# Run multiple tasks
turbo run build lint test

# Dry run (show what would run)
turbo run build --dry

# Show task graph
turbo run build --graph

# Prune for Docker
turbo prune @myorg/api --docker
```

### Strengths

- Simplest setup of any monorepo tool
- Excellent caching (local + remote)
- Great Vercel/Next.js integration
- Small learning curve
- Fast — written in Rust
- Good documentation

### Weaknesses

- JS/TS only (no multi-language support)
- No code generators
- No module boundary enforcement
- No dependency graph visualization UI
- Limited plugin ecosystem
- No migration/schematic tooling

### Best For

- Web teams using Vercel/Next.js
- Teams wanting minimal tooling overhead
- Small to medium monorepos (5-30 packages)
- Teams new to monorepo tooling

---

## Nx

### Overview

Full-featured build system and monorepo toolkit by Nrwl. Plugin-based architecture
with code generators, module boundaries, dependency graph visualization, and cloud
caching. Supports JS/TS, Go, Rust, Java, and more.

### Key Features

- **Affected commands** — Run tasks only for projects affected by changes
- **Computation caching** — Local and cloud (Nx Cloud)
- **Code generators** — Scaffold apps, libraries, components
- **Module boundary enforcement** — ESLint rules for import restrictions
- **Dependency graph UI** — Interactive visualization
- **Plugin ecosystem** — React, Angular, Node, Next.js, Vite, Jest, Cypress
- **Distributed task execution** — Split CI across machines
- **Project inference** — Auto-detect projects and targets

### Configuration (nx.json)

```json
{
  "targetDefaults": {
    "build": {
      "dependsOn": ["^build"],
      "inputs": ["production", "^production"],
      "cache": true
    },
    "test": {
      "inputs": ["default", "^production"],
      "cache": true
    },
    "lint": {
      "inputs": ["default"],
      "cache": true
    }
  },
  "namedInputs": {
    "default": ["{projectRoot}/**/*"],
    "production": [
      "default",
      "!{projectRoot}/**/*.spec.ts",
      "!{projectRoot}/**/*.test.ts"
    ]
  },
  "defaultBase": "main"
}
```

### Commands

```bash
# Run task for specific project
nx build web

# Run task for affected projects
nx affected -t build
nx affected -t test

# Run task for all projects
nx run-many -t build
nx run-many -t build --parallel=5

# Generate code
nx generate @nx/react:application admin
nx generate @nx/react:library ui --publishable

# Show dependency graph
nx graph

# Show affected graph
nx affected:graph

# Show project details
nx show project web

# Migrate Nx version
nx migrate latest
nx migrate --run-migrations
```

### Strengths

- Most comprehensive feature set
- Module boundary enforcement (architectural guardrails)
- Code generators for consistent project scaffolding
- Interactive dependency graph
- Multi-language support (Go, Rust, Java, etc.)
- Distributed task execution for CI
- Large plugin ecosystem
- Enterprise-grade with Nx Cloud

### Weaknesses

- Steeper learning curve
- More configuration required
- Can feel heavyweight for small projects
- Plugin abstraction can obscure what's happening
- Breaking changes between major versions
- Nx Cloud required for some features

### Best For

- Large teams (15+) needing architectural governance
- Enterprise monorepos with many packages (30+)
- Multi-framework projects (React + Angular + Node)
- Teams wanting code generation and scaffolding
- Projects needing module boundary enforcement

---

## Lerna

### Overview

Originally the first JavaScript monorepo tool. Now maintained by Nx team and
uses Nx under the hood for task running. Primarily focused on npm package
versioning and publishing.

### Key Features

- **Version management** — Semantic versioning with conventional commits
- **Publishing** — Publish to npm with proper dependency resolution
- **Changelog generation** — Automatic changelogs from conventional commits
- **Independent versioning** — Each package versioned separately
- **Fixed versioning** — All packages share one version
- **Nx integration** — Task running powered by Nx
- **Canary releases** — Pre-release publishing

### Configuration (lerna.json)

```json
{
  "version": "independent",
  "npmClient": "pnpm",
  "useWorkspaces": true,
  "useNx": true,
  "command": {
    "version": {
      "conventionalCommits": true,
      "createRelease": "github",
      "message": "chore(release): version packages"
    },
    "publish": {
      "conventionalCommits": true,
      "yes": true
    }
  }
}
```

### Commands

```bash
# List packages
lerna list
lerna list --all --long

# Check which packages changed
lerna changed

# Version packages
lerna version
lerna version --conventional-commits

# Publish to npm
lerna publish
lerna publish from-git

# Run script in all packages
lerna run build
lerna run test --scope=@myorg/ui

# Execute command in all packages
lerna exec -- rm -rf dist

# Bootstrap (install + link) — legacy
lerna bootstrap

# Clean
lerna clean
```

### Strengths

- Best-in-class npm publishing workflow
- Conventional commit → version bump automation
- Automatic changelog generation
- GitHub/GitLab release creation
- Mature and widely used
- Good for library authoring

### Weaknesses

- Publishing-focused (limited build orchestration without Nx)
- Now essentially a Nx wrapper for task running
- Learning both Lerna and Nx configs
- Historical baggage from v4 → v5 → v6 changes
- Less standalone value since Nx integration

### Best For

- Publishing npm packages from a monorepo
- Library authors maintaining multiple packages
- Projects needing automated versioning and changelogs
- Teams already using Lerna wanting to upgrade

---

## Package Manager Workspaces

### pnpm Workspaces

```yaml
# pnpm-workspace.yaml
packages:
  - 'apps/*'
  - 'packages/*'
```

**Key features:**
- Strict dependency isolation (no phantom dependencies)
- Content-addressable storage (faster installs, less disk)
- Fastest package manager for monorepos
- Built-in filtering and parallel execution

**Commands:**
```bash
pnpm install                            # Install all
pnpm add react --filter @myorg/web      # Add to specific package
pnpm -r run build                       # Run in all packages
pnpm --filter @myorg/web dev            # Run in specific package
pnpm --filter @myorg/web... build       # Build with dependencies
```

### Yarn Workspaces

```json
{
  "workspaces": ["apps/*", "packages/*"]
}
```

**Key features:**
- Plug'n'Play (PnP) — no node_modules
- Zero-installs (check in cache)
- Constraints for dependency policies
- Good for existing Yarn projects

### npm Workspaces

```json
{
  "workspaces": ["apps/*", "packages/*"]
}
```

**Key features:**
- Zero additional tooling
- Built into npm 7+
- Simplest possible setup

**Commands:**
```bash
npm install                                    # Install all
npm run build -w @myorg/web                    # Run in workspace
npm run build --workspaces                     # Run in all
npm run build --workspaces --if-present        # Skip if no script
```

### Workspace Comparison

| Feature | pnpm | Yarn | npm |
|---------|------|------|-----|
| Install speed | Fastest | Fast (PnP) | Moderate |
| Disk usage | Lowest | Low (PnP) | Highest |
| Strict deps | Yes (default) | Yes (PnP) | No |
| Filtering | Excellent | Good | Basic |
| Phantom deps | Prevented | Prevented (PnP) | Allowed |
| Maturity | High | High | Moderate |

**Recommendation:** Use pnpm workspaces as the foundation, then add Turborepo or Nx on top for build orchestration.

---

## Bazel and Pants

### Bazel

Google's build system. Hermetic, reproducible builds across any language.

**When to consider:**
- 500+ packages or 1M+ lines of code
- Multi-language (Java + Python + Go + JS in one repo)
- Need hermetic builds (identical output regardless of machine)
- Google/Meta-scale engineering

**When to avoid:**
- JS/TS only projects (Turborepo/Nx are simpler)
- Teams without dedicated build infrastructure engineers
- Projects under 50 packages
- When simplicity matters

### Pants

Modern build system focused on Python, with growing support for other languages.

**When to consider:**
- Python-heavy monorepo
- Need fine-grained dependency tracking for Python
- Want Bazel-like features with simpler setup
- Mixed Python + Docker + Shell projects

---

## Feature Matrix

| Feature | Turborepo | Nx | Lerna | pnpm WS | Yarn WS | npm WS |
|---------|-----------|-----|-------|---------|---------|--------|
| Task orchestration | Yes | Yes | Via Nx | Basic | Basic | Basic |
| Local caching | Yes | Yes | Via Nx | No | No | No |
| Remote caching | Yes (Vercel) | Yes (Nx Cloud) | Via Nx | No | No | No |
| Affected detection | Yes | Yes | Yes | No | No | No |
| Code generators | No | Yes | No | No | No | No |
| Module boundaries | No | Yes | No | No | No | No |
| Dep graph UI | No | Yes | No | No | No | No |
| Publishing | No | No | Yes | No | No | No |
| Multi-language | No | Yes | No | No | No | No |
| Distributed CI | No | Yes | No | No | No | No |
| Docker pruning | Yes | No | No | No | No | No |
| Watch mode | Yes | Yes | No | No | No | No |
| Plugins | No | Yes | No | No | No | No |
| Config complexity | Low | High | Medium | Minimal | Minimal | Minimal |
| Learning curve | Low | High | Medium | Low | Low | Low |

---

## Performance Benchmarks

Approximate times for a 20-package monorepo:

### Cold Build (No Cache)

| Tool | Time | Notes |
|------|------|-------|
| Turborepo | ~45s | Parallel execution, Rust-based |
| Nx | ~50s | Parallel, computation graph |
| npm workspaces | ~120s | Sequential by default |
| pnpm -r run build | ~90s | Basic parallelism |

### Warm Build (Full Cache)

| Tool | Time | Notes |
|------|------|-------|
| Turborepo | ~0.5s | Content hash match, skip all |
| Nx | ~0.8s | Computation cache hit |
| npm workspaces | ~120s | No caching, full rebuild |
| pnpm -r run build | ~90s | No caching, full rebuild |

### Affected Build (1 Package Changed)

| Tool | Time | Notes |
|------|------|-------|
| Turborepo | ~8s | Only rebuild changed + dependents |
| Nx | ~10s | Affected graph computation |
| npm workspaces | ~120s | No affected detection, full rebuild |

### Install Time

| Package Manager | Cold | Warm (Cache) |
|----------------|------|-------------|
| pnpm | ~15s | ~3s |
| Yarn (PnP) | ~20s | ~0.5s (zero-install) |
| npm | ~30s | ~8s |

---

## Decision Guide

### Choose Turborepo When

- You want the simplest monorepo setup
- Your team is small to medium (2-20)
- You use Vercel for deployment
- You only have JS/TS packages
- You want fast results with minimal config
- You're new to monorepo tooling

### Choose Nx When

- You need module boundary enforcement
- You want code generators and scaffolding
- Your team is large (15+)
- You have 30+ packages
- You need multi-language support
- You want distributed CI execution
- You need the dependency graph UI
- You're in an enterprise environment

### Choose Lerna (+ Nx) When

- You publish npm packages
- You need automated versioning with conventional commits
- You want automatic changelog generation
- You maintain a library ecosystem

### Choose pnpm Workspaces (No Orchestrator) When

- You have < 10 packages
- You want zero additional tooling
- Strict dependency isolation is a priority
- You'll add Turborepo/Nx later if needed

### Combination Recommendations

```
Most common setups:

1. pnpm + Turborepo (recommended default)
   Simple, fast, good caching, easy setup

2. pnpm + Nx (large projects)
   Comprehensive, architectural governance, generators

3. pnpm + Lerna + Nx (library publishing)
   Publishing workflow + build orchestration

4. Yarn PnP + Turborepo (Yarn teams)
   Zero-install + caching

5. pnpm only (minimal)
   Just workspace linking, no orchestration
```

---

## Migration Paths

### Adding Turborepo to Existing Workspaces

```bash
# 1. Install Turborepo
npm install turbo --save-dev

# 2. Create turbo.json
# Define your task pipeline

# 3. Update scripts in root package.json
# "build": "turbo run build"

# 4. Test
turbo run build

# 5. Set up remote caching (optional)
npx turbo login
npx turbo link
```

### Adding Nx to Existing Workspaces

```bash
# 1. Initialize Nx
npx nx init

# 2. Follow the prompts
# Nx detects existing packages and creates config

# 3. Test
nx run-many -t build

# 4. Add plugins for frameworks
npm install @nx/react @nx/vite
```

### Migrating from Lerna v4 to v7

```bash
# 1. Update Lerna
npm install lerna@latest --save-dev

# 2. Update lerna.json
# Add "useNx": true

# 3. Remove bootstrap (use workspace installs)
# Remove "lerna bootstrap" from scripts

# 4. Test versioning and publishing
lerna version --dry-run
```

### Migrating from Turborepo to Nx

```bash
# 1. Install Nx
npm install nx @nx/workspace --save-dev
npx nx init

# 2. Remove turbo.json (gradually)
# Map turbo.json tasks to nx.json targetDefaults

# 3. Update scripts
# "build": "nx run-many -t build"

# 4. Add Nx plugins for additional features
# @nx/eslint, @nx/react, etc.

# 5. Remove Turborepo
npm uninstall turbo
rm turbo.json
```

---

## Common Configurations

### Shared TypeScript Config Package

```
packages/tsconfig/
├── package.json       # name: "@myorg/tsconfig"
├── base.json          # Shared compiler options
├── react.json         # React-specific (extends base)
├── node.json          # Node.js-specific (extends base)
└── library.json       # Library-specific (extends base)
```

### Shared ESLint Config Package

```
packages/eslint-config/
├── package.json       # name: "@myorg/eslint-config"
├── index.js           # Base rules
├── react.js           # React rules (extends index)
├── node.js            # Node rules (extends index)
└── library.js         # Library rules (extends index)
```

### Recommended Package Structure

```
packages/my-package/
├── src/
│   ├── index.ts       # Public API exports
│   └── ...
├── tests/
│   └── ...
├── package.json
├── tsconfig.json      # Extends @myorg/tsconfig/...
├── .eslintrc.js       # Extends @myorg/eslint-config/...
├── vitest.config.ts   # Or jest.config.ts
└── README.md
```

### Internal Package package.json

```json
{
  "name": "@myorg/utils",
  "version": "0.0.0",
  "private": true,
  "main": "./src/index.ts",
  "types": "./src/index.ts",
  "exports": {
    ".": "./src/index.ts"
  },
  "scripts": {
    "build": "tsup src/index.ts --format esm,cjs --dts",
    "lint": "eslint .",
    "test": "vitest run",
    "typecheck": "tsc --noEmit"
  }
}
```
