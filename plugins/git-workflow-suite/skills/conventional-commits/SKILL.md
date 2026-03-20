---
name: conventional-commits
description: >
  Conventional Commits standard — commit message formatting, automated changelogs,
  semantic versioning from commits, commitlint setup, and Husky git hooks.
  Triggers: "conventional commits", "commit message", "commitlint", "semantic release",
  "changelog", "husky", "git hooks", "commit format".
  NOT for: branching (use branching-strategies), advanced git (use git-advanced).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Conventional Commits

## Format

```
<type>(<scope>): <description>

[optional body]

[optional footer(s)]
```

## Types

| Type | Purpose | Semver | Example |
|------|---------|--------|---------|
| `feat` | New feature | Minor | `feat(auth): add Google OAuth` |
| `fix` | Bug fix | Patch | `fix(api): handle null response` |
| `docs` | Documentation | None | `docs: update API reference` |
| `style` | Formatting (no logic change) | None | `style: fix indentation` |
| `refactor` | Code change (no feat/fix) | None | `refactor: extract validation` |
| `perf` | Performance improvement | Patch | `perf(db): add query index` |
| `test` | Add/update tests | None | `test: add auth unit tests` |
| `build` | Build system/deps | None | `build: upgrade to Node 20` |
| `ci` | CI/CD config | None | `ci: add deploy workflow` |
| `chore` | Maintenance tasks | None | `chore: update lockfile` |
| `revert` | Revert a commit | Patch | `revert: feat(auth) add OAuth` |

## Breaking Changes

```bash
# Method 1: Footer
git commit -m "feat(api): change pagination format

BREAKING CHANGE: Response now uses cursor-based pagination instead of offset.
Clients must update from { page, limit } to { cursor, limit }."

# Method 2: Bang notation
git commit -m "feat(api)!: change pagination format"
```

## Setup: Commitlint + Husky

```bash
# Install dependencies
pnpm add -D @commitlint/cli @commitlint/config-conventional husky

# Create commitlint config
cat > commitlint.config.js << 'EOF'
module.exports = {
  extends: ['@commitlint/config-conventional'],
  rules: {
    'type-enum': [2, 'always', [
      'feat', 'fix', 'docs', 'style', 'refactor',
      'perf', 'test', 'build', 'ci', 'chore', 'revert',
    ]],
    'scope-case': [2, 'always', 'kebab-case'],
    'subject-case': [2, 'never', ['start-case', 'pascal-case', 'upper-case']],
    'subject-max-length': [2, 'always', 72],
    'body-max-line-length': [2, 'always', 100],
  },
};
EOF

# Initialize Husky
npx husky init

# Add commit-msg hook
echo 'npx --no -- commitlint --edit "$1"' > .husky/commit-msg

# Add pre-commit hook (lint + format)
cat > .husky/pre-commit << 'EOF'
npx lint-staged
EOF
```

## Lint-Staged Configuration

```json
// package.json
{
  "lint-staged": {
    "*.{ts,tsx}": [
      "eslint --fix",
      "prettier --write"
    ],
    "*.{json,md,yml}": [
      "prettier --write"
    ],
    "*.css": [
      "prettier --write"
    ]
  }
}
```

## Automated Changelog Generation

```bash
# Install standard-version (or use changesets for monorepos)
pnpm add -D standard-version

# Generate changelog + bump version + create tag
npx standard-version

# First release
npx standard-version --first-release

# Pre-release
npx standard-version --prerelease alpha
# 1.0.0 → 1.0.1-alpha.0

# Specific bump
npx standard-version --release-as minor
```

```json
// .versionrc.json
{
  "types": [
    { "type": "feat", "section": "Features" },
    { "type": "fix", "section": "Bug Fixes" },
    { "type": "perf", "section": "Performance" },
    { "type": "refactor", "section": "Code Refactoring" },
    { "type": "docs", "hidden": true },
    { "type": "style", "hidden": true },
    { "type": "test", "hidden": true },
    { "type": "build", "hidden": true },
    { "type": "ci", "hidden": true },
    { "type": "chore", "hidden": true }
  ]
}
```

## Semantic Release (CI/CD)

```bash
pnpm add -D semantic-release @semantic-release/changelog @semantic-release/git
```

```json
// .releaserc.json
{
  "branches": ["main"],
  "plugins": [
    "@semantic-release/commit-analyzer",
    "@semantic-release/release-notes-generator",
    ["@semantic-release/changelog", { "changelogFile": "CHANGELOG.md" }],
    ["@semantic-release/npm", { "npmPublish": true }],
    ["@semantic-release/git", {
      "assets": ["CHANGELOG.md", "package.json"],
      "message": "chore(release): ${nextRelease.version}"
    }],
    "@semantic-release/github"
  ]
}
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
        with:
          persist-credentials: false
      - uses: actions/setup-node@v4
        with:
          node-version: 20
      - run: npm ci
      - run: npx semantic-release
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          NPM_TOKEN: ${{ secrets.NPM_TOKEN }}
```

## Commit Message Templates

```bash
# Set up a commit template
cat > ~/.gitmessage << 'EOF'
# <type>(<scope>): <description>
#
# [optional body]
#
# [optional footer(s)]
#
# Types: feat fix docs style refactor perf test build ci chore revert
# Scope: component or module name (e.g., auth, api, ui)
# Description: imperative mood, lowercase, no period
#
# Body: explain WHY, not WHAT (the diff shows what)
# Footer: BREAKING CHANGE: description
#         Fixes #123
#         Co-Authored-By: Name <email>
EOF

git config --global commit.template ~/.gitmessage
```

## Multi-Line Commit Messages

```bash
# Using heredoc (recommended for scripts)
git commit -m "$(cat <<'EOF'
feat(auth): add refresh token rotation

Implement refresh token rotation with family tracking to detect
token reuse attacks. Each refresh token use generates a new pair
and invalidates the previous token.

- Add token family tracking to prevent replay attacks
- Rotate both access and refresh tokens on refresh
- Invalidate all family tokens on reuse detection
- Add 7-day absolute expiry on refresh tokens

BREAKING CHANGE: Refresh endpoint now returns both access and
refresh tokens. Clients must update stored refresh token on each use.

Fixes #234
EOF
)"
```

## Gotchas

1. **Scope is optional but valuable.** `feat: add thing` is valid but `feat(auth): add thing` is much more useful for changelog grouping and searching commit history.

2. **Description must be imperative mood.** "add feature" not "added feature" or "adds feature". Think of it as completing the sentence "This commit will...".

3. **BREAKING CHANGE must be uppercase in the footer.** Lowercase `breaking change:` or `Breaking Change:` won't be detected by semantic-release or standard-version.

4. **Husky hooks don't run in CI by default.** Set `HUSKY=0` in CI environment to skip hooks, or use `--no-verify` in CI scripts. Hooks are for local development safety.

5. **Commitlint checks the FINAL message.** If you use `--amend`, commitlint validates the amended message, not the original. This is correct behavior but can be confusing.

6. **Standard-version and semantic-release serve different needs.** Standard-version is manual (you decide when to release). Semantic-release is automated (every push to main triggers a release if there are releasable commits). Don't use both.
