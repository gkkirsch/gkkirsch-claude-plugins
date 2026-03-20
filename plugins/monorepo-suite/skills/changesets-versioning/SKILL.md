---
name: changesets-versioning
description: >
  Monorepo versioning and publishing with Changesets — creating changesets,
  version bumping, changelog generation, publishing to npm, and CI/CD automation.
  Triggers: "changeset", "version bump", "publish package", "changelog",
  "monorepo versioning", "npm publish", "release workflow".
  NOT for: workspace setup (use turborepo-setup), package structure (use workspace-packages).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Changesets Versioning & Publishing

## Setup

```bash
# Install changesets
pnpm add -Dw @changesets/cli @changesets/changelog-github

# Initialize
pnpm changeset init
# Creates .changeset/ directory with config.json
```

```json
// .changeset/config.json
{
  "$schema": "https://unpkg.com/@changesets/config@3.0.0/schema.json",
  "changelog": [
    "@changesets/changelog-github",
    { "repo": "org/repo" }
  ],
  "commit": false,
  "fixed": [],
  "linked": [
    ["@repo/ui", "@repo/theme"]
  ],
  "access": "public",
  "baseBranch": "main",
  "updateInternalDependencies": "patch",
  "ignore": [
    "@repo/docs",
    "@repo/storybook"
  ],
  "privatePackages": {
    "version": true,
    "tag": false
  }
}
```

## Creating Changesets

```bash
# Interactive changeset creation
pnpm changeset
# 1. Select packages that changed
# 2. Choose bump type (major/minor/patch) for each
# 3. Write summary of changes

# Creates .changeset/<random-name>.md:
```

```markdown
---
"@repo/ui": minor
"@repo/utils": patch
---

Add new Button variant and fix utility type inference
```

```bash
# Non-interactive (CI-friendly)
pnpm changeset --empty
# Creates an empty changeset for PRs with no version bump needed
```

## Version Bumping

```bash
# Apply all pending changesets
pnpm changeset version
# 1. Reads all .changeset/*.md files
# 2. Bumps versions in package.json files
# 3. Updates CHANGELOG.md in each package
# 4. Deletes consumed changeset files
# 5. Updates internal dependency versions

# Check what would happen (dry run)
pnpm changeset status
pnpm changeset status --verbose
```

## Publishing

```bash
# Build all packages first
pnpm build

# Publish changed packages to npm
pnpm changeset publish
# Only publishes packages with version changes since last publish

# For scoped packages (@org/name), ensure access is set:
# package.json: "publishConfig": { "access": "public" }
```

## CI/CD Automation

```yaml
# .github/workflows/release.yml
name: Release
on:
  push:
    branches: [main]

concurrency: ${{ github.workflow }}-${{ github.ref }}

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: pnpm/action-setup@v4
      - uses: actions/setup-node@v4
        with:
          node-version: 20
          cache: pnpm
          registry-url: https://registry.npmjs.org

      - run: pnpm install --frozen-lockfile
      - run: pnpm build

      - name: Create Release PR or Publish
        id: changesets
        uses: changesets/action@v1
        with:
          version: pnpm changeset version
          publish: pnpm changeset publish
          title: "chore: version packages"
          commit: "chore: version packages"
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          NPM_TOKEN: ${{ secrets.NPM_TOKEN }}
          NODE_AUTH_TOKEN: ${{ secrets.NPM_TOKEN }}
```

## Snapshot Releases (Pre-releases)

```bash
# Create a snapshot for testing (no version bump, no changelog)
pnpm changeset version --snapshot canary
# Produces versions like: 1.2.3-canary-20260315120000

# Publish snapshot
pnpm changeset publish --tag canary --no-git-tag
# Install with: pnpm add @repo/ui@canary
```

## Pre-release Mode

```bash
# Enter pre-release mode
pnpm changeset pre enter beta
# All subsequent version bumps produce beta versions

# Create changesets normally
pnpm changeset
pnpm changeset version
# Produces: 2.0.0-beta.0, 2.0.0-beta.1, etc.

# Exit pre-release mode
pnpm changeset pre exit
pnpm changeset version
# Produces final: 2.0.0
```

## Linked and Fixed Packages

```json
// .changeset/config.json
{
  // "linked" — packages bump together when ANY of them changes
  // Use for packages that should always be the same version
  "linked": [
    ["@repo/ui", "@repo/theme", "@repo/icons"]
  ],

  // "fixed" — packages are ALWAYS the same version
  // Even if only one changes, all get bumped
  "fixed": [
    ["@repo/core", "@repo/react", "@repo/vue"]
  ]
}
```

## Package.json Configuration

```json
{
  "name": "@repo/ui",
  "version": "1.0.0",
  "private": false,
  "main": "./dist/index.js",
  "module": "./dist/index.mjs",
  "types": "./dist/index.d.ts",
  "exports": {
    ".": {
      "import": "./dist/index.mjs",
      "require": "./dist/index.js",
      "types": "./dist/index.d.ts"
    },
    "./button": {
      "import": "./dist/button.mjs",
      "require": "./dist/button.js",
      "types": "./dist/button.d.ts"
    }
  },
  "files": ["dist"],
  "publishConfig": {
    "access": "public",
    "registry": "https://registry.npmjs.org"
  },
  "scripts": {
    "build": "tsup src/index.ts --format cjs,esm --dts",
    "dev": "tsup src/index.ts --format cjs,esm --dts --watch",
    "prepublishOnly": "pnpm build"
  }
}
```

## Root Package.json Scripts

```json
{
  "scripts": {
    "changeset": "changeset",
    "version-packages": "changeset version",
    "release": "pnpm build && changeset publish",
    "ci:release": "pnpm build && changeset version && changeset publish"
  }
}
```

## Gotchas

1. **Changesets must be committed to git.** The `.changeset/*.md` files are part of the PR. They get consumed (deleted) when you run `changeset version`. If you don't commit them, version bumping won't work.

2. **`updateInternalDependencies: "patch"` auto-bumps dependents.** If `@repo/ui` gets a minor bump, packages depending on it get a patch bump automatically. Set to `"minor"` for stricter versioning or `"none"` to skip.

3. **Private packages aren't published by default.** Set `"privatePackages": { "version": true, "tag": false }` in config to still version them (useful for apps) without publishing to npm.

4. **`linked` vs `fixed` is confusing.** Linked: if ANY package in the group gets a changeset, they ALL bump to the same version. Fixed: they're ALWAYS the same version, even without changesets. Use linked for related packages, fixed for framework packages.

5. **The changesets GitHub Action creates a "Version Packages" PR.** Don't merge changesets and version bumps in the same PR. The action accumulates changesets on main, then opens a PR with all version bumps. Merging that PR triggers publish.

6. **`--snapshot` doesn't update CHANGELOG.** Snapshots are ephemeral test versions. They bypass the normal version/changelog flow. Don't use them for real releases.
