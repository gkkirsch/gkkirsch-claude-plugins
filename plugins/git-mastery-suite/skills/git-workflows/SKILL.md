---
name: git-workflows
description: >
  Git branching strategies and team workflows — trunk-based development,
  GitHub Flow, Git Flow, conventional commits, PR best practices, and
  merge strategies.
  Triggers: "git workflow", "branching strategy", "git flow", "trunk based",
  "conventional commits", "pr workflow", "merge strategy".
  NOT for: Advanced git operations (use git-advanced).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Git Workflows

## Branching Strategies

### Trunk-Based Development (Recommended for Most Teams)

```
main ──●──●──●──●──●──●──●──●──●──●──
       │        ↑     │        ↑
       └─feat─A─┘     └─feat─B─┘
       (< 1 day)      (< 1 day)
```

- Short-lived branches (hours to 1-2 days max)
- Merge to main frequently
- Feature flags for incomplete work
- CI/CD deploys from main
- Best for: small teams, continuous deployment, microservices

```bash
# Daily workflow
git checkout main && git pull
git checkout -b feat/add-user-search
# ... work for a few hours ...
git push -u origin feat/add-user-search
gh pr create --fill
# Merge same day, delete branch
```

### GitHub Flow

```
main ──●──●──────●──●──────●──●──
       │         ↑  │         ↑
       └─feat─A──┘  └─feat─B──┘
       (1-3 days)   (1-5 days)
```

- Branch from main, PR back to main
- No release branches (deploy from main)
- Code review required before merge
- Best for: SaaS, web apps, continuous deployment

### Git Flow (Complex Projects)

```
main ────●────────────────●────────
         ↑                ↑
develop ──●──●──●──●──●──●──●──●──
          │     ↑  │     ↑
          └─feat┘  └─feat┘
               ↓
         release/1.0 ──●──●
```

- `main` = production, `develop` = integration
- Feature branches from `develop`
- Release branches for stabilization
- Hotfix branches from `main`
- Best for: versioned software, mobile apps, packages

## Conventional Commits

```
<type>(<scope>): <description>

[optional body]

[optional footer(s)]
```

### Types

| Type | When | Example |
|------|------|---------|
| `feat` | New feature | `feat(auth): add OAuth2 login` |
| `fix` | Bug fix | `fix(cart): correct total calculation` |
| `docs` | Documentation | `docs: update API reference` |
| `style` | Formatting (no logic change) | `style: fix indentation in utils` |
| `refactor` | Code change (no feat/fix) | `refactor(db): extract query builder` |
| `perf` | Performance improvement | `perf: cache user lookup results` |
| `test` | Add/update tests | `test(auth): add login edge cases` |
| `build` | Build system / dependencies | `build: upgrade vite to 6.x` |
| `ci` | CI/CD changes | `ci: add playwright to pipeline` |
| `chore` | Maintenance | `chore: clean up unused imports` |

### Breaking Changes

```
feat(api)!: change authentication to JWT

BREAKING CHANGE: The /auth/login endpoint now returns a JWT token
instead of a session cookie. All clients must update to send the
token in the Authorization header.
```

### Enforce with Commitlint

```bash
# Install
npm install -D @commitlint/cli @commitlint/config-conventional

# commitlint.config.js
echo "module.exports = { extends: ['@commitlint/config-conventional'] };" > commitlint.config.js

# Husky hook
npx husky add .husky/commit-msg 'npx --no -- commitlint --edit $1'
```

## PR Best Practices

### PR Template

```markdown
<!-- .github/pull_request_template.md -->
## Summary
<!-- What does this PR do? Why? -->

## Changes
- [ ] Change 1
- [ ] Change 2

## Testing
- [ ] Unit tests added/updated
- [ ] Manual testing completed
- [ ] Edge cases considered

## Screenshots
<!-- If UI changes, include before/after screenshots -->
```

### PR Size Guidelines

| Lines Changed | Verdict | Action |
|--------------|---------|--------|
| < 50 | Small | Quick review, merge fast |
| 50-200 | Medium | Normal review cycle |
| 200-500 | Large | Split if possible |
| > 500 | Too large | Must split into smaller PRs |

### Stacked PRs

```bash
# Feature too large for one PR? Stack them.

# PR 1: Database schema changes
git checkout -b feat/user-roles-schema
# ... schema + migration ...
gh pr create --base main

# PR 2: API endpoints (depends on PR 1)
git checkout -b feat/user-roles-api
# ... endpoints ...
gh pr create --base feat/user-roles-schema

# PR 3: UI (depends on PR 2)
git checkout -b feat/user-roles-ui
# ... components ...
gh pr create --base feat/user-roles-api

# Review and merge in order: schema → api → ui
```

## Merge Strategies

### Merge Commit (Default)

```bash
git merge feature-branch
# Creates a merge commit, preserves full branch history
```

```
main:  A───B───C───────M───
                      / \
feature: D───E───F──┘
```

- Preserves complete history
- Easy to revert entire feature (revert the merge commit)
- Can make history noisy

### Squash Merge

```bash
git merge --squash feature-branch
git commit -m "feat(auth): add OAuth2 login (#123)"
```

```
main:  A───B───C───S───
                   ↑
                   (D+E+F squashed into one commit)
```

- Clean linear history on main
- One commit = one feature/PR
- Loses individual commit detail
- **Best for most teams** — clean main, detailed work in PR

### Rebase

```bash
git checkout feature-branch
git rebase main
git checkout main
git merge --ff-only feature-branch
```

```
Before: main: A───B───C
        feat:      └──D───E

After:  main: A───B───C───D'───E'
```

- Linear history, no merge commits
- Rewrites commit SHAs (never rebase shared branches)
- Requires force push after rebase

### Which to Use?

| Strategy | History | Best For |
|----------|---------|----------|
| Merge commit | Full history preserved | Open source, large teams |
| Squash merge | Clean one-commit-per-PR | Product teams, SaaS |
| Rebase + FF | Clean linear commits | Solo/pair, libraries |

## Branch Protection

```bash
# GitHub CLI — set up branch protection
gh api repos/{owner}/{repo}/branches/main/protection \
  --method PUT \
  --field required_status_checks='{"strict":true,"contexts":["ci/test","ci/lint"]}' \
  --field enforce_admins=true \
  --field required_pull_request_reviews='{"required_approving_review_count":1}' \
  --field restrictions=null
```

| Rule | Purpose |
|------|---------|
| Require PR reviews (1+) | No solo merges to main |
| Require status checks | CI must pass before merge |
| Require up-to-date branch | Must rebase on latest main |
| No force push | Prevent history rewriting |
| Require linear history | Enforce squash or rebase |

## Git Hooks (Local)

```bash
# .husky/pre-commit
npx lint-staged

# .husky/commit-msg
npx --no -- commitlint --edit $1

# .husky/pre-push
npm run test -- --bail
```

```json
// package.json — lint-staged config
{
  "lint-staged": {
    "*.{ts,tsx}": ["eslint --fix", "prettier --write"],
    "*.{css,scss}": ["prettier --write"],
    "*.{json,md}": ["prettier --write"]
  }
}
```

## Gotchas

1. **Never rebase a shared branch** — `git rebase` rewrites commit SHAs. If others have pulled the branch, their history diverges and they need to force-reset. Only rebase local/unshared branches.

2. **Squash merge loses co-author credit** — When you squash, individual commits (and their authors) are lost. Add `Co-authored-by:` trailers to the squash commit message to preserve attribution.

3. **Long-lived branches cause painful merges** — A branch open for 2 weeks accumulates conflicts. Merge main into your branch daily, or rebase on main frequently. Short branches are the best conflict prevention.

4. **Force push overwrites remote history** — `git push --force` replaces the remote branch. Anyone who pulled it now has orphaned commits. Use `--force-with-lease` instead — it fails if someone else pushed since your last fetch.

5. **`.gitignore` doesn't remove tracked files** — Adding a file to `.gitignore` after it's committed doesn't remove it. Run `git rm --cached <file>` to untrack it while keeping the local copy.

6. **Git hooks aren't shared by default** — Hooks in `.git/hooks/` aren't committed. Use Husky or lefthook to commit hooks as part of the project so the whole team runs them.
