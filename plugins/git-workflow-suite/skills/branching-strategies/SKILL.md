---
name: branching-strategies
description: >
  Git branching strategies — trunk-based development, GitHub Flow, GitFlow,
  branch protection rules, merge strategies, and release management.
  Triggers: "branching strategy", "GitFlow", "trunk-based", "GitHub Flow",
  "release branch", "merge strategy", "branch protection".
  NOT for: commit messages (use conventional-commits), advanced git (use git-advanced).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Git Branching Strategies

## Trunk-Based Development

```bash
# The simplest workflow — everyone commits to main
# Short-lived feature branches (< 1 day) are optional

# Start work
git checkout main
git pull origin main
git checkout -b feat/quick-change

# Work, commit, push
git add -A && git commit -m "feat: add user avatar"
git push origin feat/quick-change

# Create PR, get review, merge immediately
# Branch lives < 1 day

# Feature flags for incomplete features
if (featureFlags.newDashboard) {
  return <NewDashboard />;
}
return <OldDashboard />;
```

### When to Use
- Mature CI/CD pipeline with automated tests
- Team deploys multiple times per day
- Strong code review culture
- Feature flags infrastructure available

## GitHub Flow

```bash
# Simple: main + feature branches
# 1. Create feature branch from main
git checkout main
git pull origin main
git checkout -b feat/user-authentication

# 2. Work on feature (multiple commits OK)
git add -A && git commit -m "feat(auth): add login form"
git add -A && git commit -m "feat(auth): add JWT middleware"
git add -A && git commit -m "test(auth): add login tests"

# 3. Push and create PR
git push origin feat/user-authentication
gh pr create --title "feat: add user authentication" --body "..."

# 4. Review, discuss, iterate
# 5. Merge to main (squash or merge commit)
# 6. Deploy from main
```

### When to Use
- Most web applications
- Teams of 2-15 developers
- Continuous deployment or weekly releases
- No need for multiple release versions

## GitFlow

```bash
# For scheduled releases with multiple environments

# Feature development
git checkout develop
git pull origin develop
git checkout -b feature/user-dashboard

# Work on feature...
git add -A && git commit -m "feat: add dashboard layout"

# Merge feature into develop
git checkout develop
git merge --no-ff feature/user-dashboard
git branch -d feature/user-dashboard

# Create release branch
git checkout -b release/1.2.0 develop

# Fix bugs on release branch only
git commit -m "fix: correct date format in release"

# Merge release to main AND develop
git checkout main
git merge --no-ff release/1.2.0
git tag -a v1.2.0 -m "Release 1.2.0"

git checkout develop
git merge --no-ff release/1.2.0
git branch -d release/1.2.0

# Hotfix (from main)
git checkout -b hotfix/1.2.1 main
git commit -m "fix: critical auth bypass"
git checkout main
git merge --no-ff hotfix/1.2.1
git tag -a v1.2.1 -m "Hotfix 1.2.1"
git checkout develop
git merge --no-ff hotfix/1.2.1
```

### When to Use
- Mobile apps with app store review cycles
- Enterprise software with scheduled release windows
- Multiple versions maintained in production
- Separate QA/staging environments

## Branch Protection Rules

```yaml
# GitHub branch protection (via gh CLI)
gh api repos/{owner}/{repo}/branches/main/protection -X PUT -f '{
  "required_status_checks": {
    "strict": true,
    "contexts": ["ci/build", "ci/test", "ci/lint"]
  },
  "enforce_admins": true,
  "required_pull_request_reviews": {
    "required_approving_review_count": 1,
    "dismiss_stale_reviews": true,
    "require_code_owner_reviews": true
  },
  "restrictions": null,
  "required_linear_history": true,
  "allow_force_pushes": false,
  "allow_deletions": false
}'
```

## Merge Strategies

```bash
# Merge commit (preserves history)
git merge --no-ff feature/branch
# Creates: merge commit with full branch history
# Best for: GitFlow, long-lived branches

# Squash merge (clean history)
git merge --squash feature/branch
git commit -m "feat: add complete auth system"
# Creates: single commit with all changes
# Best for: GitHub Flow, feature branches

# Rebase merge (linear history)
git rebase main
git checkout main
git merge --ff-only feature/branch
# Creates: linear commit history, no merge commits
# Best for: Trunk-based, when commit history matters

# Rebase interactively before merging
git rebase -i main
# pick, squash, fixup, reword commits
```

## CODEOWNERS

```bash
# .github/CODEOWNERS
# Global owners
* @team-lead @senior-dev

# Frontend
/src/components/ @frontend-team
/src/pages/ @frontend-team
*.css @frontend-team
*.tsx @frontend-team

# Backend
/src/api/ @backend-team
/src/db/ @backend-team @dba-team
/prisma/ @backend-team @dba-team

# DevOps
/Dockerfile @devops-team
/.github/ @devops-team
/terraform/ @devops-team

# Docs
/docs/ @docs-team
*.md @docs-team
```

## Gotchas

1. **GitFlow is overkill for most web apps.** If you deploy on every merge to main, you don't need develop, release, or hotfix branches. GitHub Flow is simpler and works better for continuous deployment.

2. **Long-lived feature branches cause pain.** The longer a branch lives, the harder it is to merge. Rebase regularly (`git rebase main`) or split the feature into smaller PRs.

3. **Squash merge loses commit granularity.** Individual commits in the feature branch are gone after squash. If you need that history, use merge commits instead.

4. **"Rebase and merge" can rewrite history unexpectedly.** GitHub's "Rebase and merge" button re-creates commits with new SHAs. If someone based work on the original commits, they'll have conflicts.

5. **Branch protection doesn't prevent force-push unless explicitly configured.** Enable "Do not allow force pushes" separately from requiring PR reviews.

6. **CODEOWNERS requires the file to be on the default branch.** Changes to CODEOWNERS on a feature branch don't take effect until merged. Plan CODEOWNERS changes ahead of time.
