# Branching Strategies Comparison Guide

Comprehensive comparison of git branching strategies with decision frameworks,
migration paths, and real-world recommendations.

---

## Table of Contents

- [Strategy Overview](#strategy-overview)
- [Trunk-Based Development](#trunk-based-development)
- [GitHub Flow](#github-flow)
- [GitFlow](#gitflow)
- [GitLab Flow](#gitlab-flow)
- [Release Branching](#release-branching)
- [Forking Workflow](#forking-workflow)
- [Decision Framework](#decision-framework)
- [Migration Guides](#migration-guides)
- [Anti-Patterns](#anti-patterns)

---

## Strategy Overview

### Quick Comparison

| Strategy | Complexity | Release | Branches | Best For |
|----------|-----------|---------|----------|----------|
| Trunk-Based | Low | Continuous | main + short-lived | SaaS, experienced teams |
| GitHub Flow | Low | Continuous | main + feature | Most web projects |
| GitFlow | High | Scheduled | main, develop, feature, release, hotfix | Enterprise, mobile |
| GitLab Flow | Medium | Environment-based | main + env branches | GitLab users, staged deploys |
| Release Branching | Medium | Multiple versions | main + release/* | Libraries, SDKs |
| Forking | Medium | Pull-request based | Forks + branches | Open source |

### Key Metrics by Strategy

| Metric | Trunk | GitHub Flow | GitFlow | GitLab Flow | Release |
|--------|-------|-------------|---------|-------------|---------|
| Time to production | Minutes | Hours | Days-weeks | Hours-days | Days |
| Merge conflict risk | Very low | Low | High | Medium | Medium |
| Process overhead | Minimal | Low | High | Medium | Medium |
| Rollback complexity | Low | Low | Medium | Low | Medium |
| Learning curve | Medium* | Low | High | Medium | Medium |
| Audit trail | Good | Good | Excellent | Good | Excellent |

*Trunk-based requires feature flag discipline, which adds learning curve.

---

## Trunk-Based Development

### Overview

All developers work on a single branch (main/trunk). Feature branches exist but
live for hours, not days. Incomplete features are hidden behind feature flags.

### Branch Model

```
main ─────●─────●─────●─────●─────●─────●─────── (always deployable)
           \   /       \   /       \   /
            ●─●         ●─●         ●
          (hours)     (hours)     (hours)
```

### Rules

1. Main is always deployable
2. Feature branches live < 24 hours (ideally < 4 hours)
3. No long-lived branches
4. Feature flags for incomplete work
5. Merge to main at least once per day
6. All commits go through CI before merge
7. Rebase merges for linear history

### Prerequisites

- Strong CI/CD pipeline (< 10 min)
- Feature flag infrastructure
- Good test coverage (> 80%)
- Team discipline for small, frequent commits
- Code review culture (pair programming or fast reviews)

### When to Use

- SaaS with continuous deployment
- Experienced teams comfortable with feature flags
- Teams practicing pair programming
- When maximum deployment speed matters

### When to Avoid

- Junior teams without CI/CD experience
- Products requiring release branches (mobile apps, libraries)
- Regulatory environments requiring release sign-off
- Teams without feature flag infrastructure

### Feature Flag Patterns

```
Release flag: Enable complete feature for all users at once
├── Usecase: Feature complete but waiting for marketing launch
└── Lifecycle: Remove flag after launch

Experiment flag: A/B test between implementations
├── Usecase: Testing new checkout flow vs old
└── Lifecycle: Remove losing variant after experiment

Ops flag: Circuit breaker or kill switch
├── Usecase: Disable feature if it causes performance issues
└── Lifecycle: Keep permanently as safety mechanism

Permission flag: Feature available to specific users
├── Usecase: Beta features, premium features
└── Lifecycle: May become permanent (premium) or temporary (beta)
```

---

## GitHub Flow

### Overview

Simple branch-based workflow. Create branch from main, work on feature, open pull
request, review, merge to main. Main is always deployable.

### Branch Model

```
main ─────●─────────●───────────●─────────●────── (always deployable)
           \       /             \       /
            ●──●──●               ●──●──●
           feature/auth         fix/login-bug
          (days to 1 week)     (hours to days)
```

### Rules

1. Main is always deployable
2. Create descriptive branch names: `type/description`
3. Open PR early (draft if work-in-progress)
4. Request reviews when ready
5. Squash merge to main
6. Auto-delete branches after merge
7. Deploy from main (or on merge)

### Branch Naming Convention

```
feature/PROJ-123-add-user-profiles
fix/PROJ-456-login-timeout
chore/update-dependencies
docs/api-reference-v2
refactor/extract-auth-service
test/payment-edge-cases
perf/optimize-search-query
```

### PR Workflow

```
1. Create branch
   git checkout -b feature/PROJ-123-description

2. Work and commit
   git commit -m "feat(auth): add login endpoint"

3. Push and create PR
   git push -u origin feature/PROJ-123-description
   gh pr create --title "feat: add user authentication" --draft

4. Mark ready when done
   gh pr ready

5. Request review
   gh pr edit --add-reviewer alice,bob

6. Address feedback
   git commit -m "fix: address review feedback"
   git push

7. Merge (squash)
   gh pr merge --squash --delete-branch
```

### When to Use

- Most web applications
- Teams of 2-20 developers
- Continuous or near-continuous deployment
- Open source projects (with forks)
- Default choice when unsure

### When to Avoid

- Need to maintain multiple production versions
- Strict release schedule with QA gates
- Very large teams (50+) without merge queue

---

## GitFlow

### Overview

Structured branching model with dedicated branches for features, releases, and
hotfixes. Two permanent branches: main (production) and develop (integration).

### Branch Model

```
main    ──●──────────────────●──────────●──── (production releases)
           \                / \        /
release     \    ●──●──●──●    \      /
             \  /               \    /
develop ──●───●──●──●──●──●──●──●──●──●──── (integration)
           \       /   \      /
feature     ●──●──●     ●──●──●
```

### Branch Types

| Branch | Purpose | From | Into | Naming |
|--------|---------|------|------|--------|
| main | Production | — | — | `main` |
| develop | Integration | main | — | `develop` |
| feature/* | New features | develop | develop | `feature/PROJ-123-desc` |
| release/* | Release prep | develop | main + develop | `release/v2.1.0` |
| hotfix/* | Production fix | main | main + develop | `hotfix/v2.1.1-desc` |

### Lifecycle

```
Feature lifecycle:
  1. Branch from develop
  2. Implement feature
  3. PR to develop (squash merge)
  4. Delete feature branch

Release lifecycle:
  1. Branch from develop
  2. Bug fixes only (no new features!)
  3. Update version numbers
  4. PR to main (merge commit)
  5. Tag main with version
  6. Merge release back to develop
  7. Delete release branch

Hotfix lifecycle:
  1. Branch from main
  2. Fix the bug (minimal change)
  3. PR to main (fast-track review)
  4. Tag main with patch version
  5. Merge hotfix back to develop
  6. Delete hotfix branch
```

### Version Tagging

```bash
# After release merge to main
git checkout main
git merge --no-ff release/v2.1.0
git tag -a v2.1.0 -m "Release v2.1.0"
git push origin main --tags

# After hotfix merge to main
git checkout main
git merge --no-ff hotfix/v2.1.1-critical-fix
git tag -a v2.1.1 -m "Hotfix v2.1.1: fix critical auth bypass"
git push origin main --tags

# Back-merge to develop
git checkout develop
git merge release/v2.1.0  # or hotfix branch
git push origin develop
```

### When to Use

- Mobile apps with app store release cycles
- Enterprise software with scheduled releases
- Projects needing QA/UAT release stages
- Teams maintaining multiple versions
- Regulatory environments needing audit trails
- Large teams (20+) with release managers

### When to Avoid

- Continuous deployment / SaaS
- Small teams (< 5 developers)
- Rapid iteration without release cycles
- When simplicity is a priority

---

## GitLab Flow

### Overview

Combines GitHub Flow simplicity with environment branches. Code flows from feature
branches through environment-specific branches (pre-production, production).

### Branch Model — Environment Branches

```
main ──●──●──●──●──●──●──●──── (latest development)
        \     \       \
staging  ●─────●───────●─────── (staging environment)
                \       \
production       ●───────●───── (production environment)
```

### Branch Model — Release Branches

```
main ──●──●──●──●──●──●──●──── (latest development)
        \           \
v2.x     ●──●──●────●──────── (v2 maintenance)
                      \
v1.x                   ●──●── (v1 LTS)
```

### Flow Rules

1. Feature branches merge into main
2. main auto-deploys to development/staging
3. Cherry-pick or merge from main to staging
4. Cherry-pick or merge from staging to production
5. Hotfixes go to main first, then cherry-pick downstream

### Promotion Process

```
1. Feature development:
   feature/auth → main (via PR)

2. Staging promotion:
   main → staging (via merge or cherry-pick)
   Automated staging tests run
   QA team verifies

3. Production promotion:
   staging → production (via merge)
   Requires deployment approval
   Auto-deploys to production

4. Hotfix:
   hotfix → main (via PR)
   Cherry-pick: main → staging → production
```

### When to Use

- GitLab CI/CD users (natural fit)
- Teams needing environment-based promotion
- When you want GitHub Flow + deployment gates
- Medium-complexity release processes

---

## Release Branching

### Overview

For projects that maintain multiple versions simultaneously. Each major/minor
version gets a release branch for long-term maintenance.

### Branch Model

```
main ──●──●──●──●──●──●──── (next version development)
        \           \
v3.x     ●──●──●────●────── (current stable)
          \    \      \
           v3.0 v3.1  v3.2  (tags)

v2.x ──●──●──●──●──●──●──── (LTS maintenance)
        \    \      \
         v2.8 v2.9  v2.10  (tags)
```

### Maintenance Policy

```
Version support tiers:

Current (v3.x):
  ✅ New features
  ✅ Bug fixes
  ✅ Security patches
  ✅ Performance improvements

LTS (v2.x):
  ❌ New features
  ✅ Bug fixes
  ✅ Security patches
  ❌ Performance improvements

End of Life (v1.x):
  ❌ New features
  ❌ Bug fixes
  ✅ Critical security patches only
  ❌ Performance improvements
```

### Backporting Process

```
1. Fix is developed and merged to main

2. Determine backport targets:
   - Security fix → All supported versions
   - Bug fix → Current + LTS
   - Feature → Current only

3. Cherry-pick to release branches:
   git checkout release/v3.x
   git cherry-pick <commit-from-main>
   # Resolve any conflicts due to version differences

4. Tag the release:
   git tag -a v3.2.1 -m "v3.2.1: security fix for CVE-2024-XXXX"

5. Repeat for each target version
```

### When to Use

- Open source libraries (React, Vue, Node.js)
- SDKs and client libraries
- Installed software (not SaaS)
- Any project with customers on multiple versions
- When backward compatibility is critical

---

## Forking Workflow

### Overview

Each contributor works on their own fork. Changes are proposed via pull requests
from fork to upstream. Common in open source projects.

### Flow

```
upstream/main ──●──●──●──●──●──── (official repository)
                 ↑      ↑
                PR     PR
                 |      |
fork/main ──●──●──●──●──●──●──── (your fork)
              \      /
               ●──●──●
             feature/contribution
```

### Workflow Steps

```bash
# 1. Fork on GitHub (web UI or gh CLI)
gh repo fork upstream/repo

# 2. Clone your fork
git clone git@github.com:you/repo.git

# 3. Add upstream remote
git remote add upstream https://github.com/upstream/repo.git

# 4. Create feature branch
git checkout -b feature/my-contribution

# 5. Work and commit

# 6. Keep in sync with upstream
git fetch upstream
git rebase upstream/main

# 7. Push to your fork
git push origin feature/my-contribution

# 8. Create PR from fork to upstream
gh pr create --repo upstream/repo

# 9. After PR is merged, sync fork
git checkout main
git fetch upstream
git merge upstream/main
git push origin main
```

### When to Use

- Open source projects
- When contributors don't have push access
- Large organizations with many external contributors
- When you want maximum isolation between contributors

---

## Decision Framework

### Quick Decision Tree

```
How often do you deploy to production?

├── Multiple times per day
│   └── Team has feature flag infrastructure?
│       ├── Yes → Trunk-Based Development
│       └── No → GitHub Flow
│
├── Daily to weekly
│   └── Need environment promotion gates?
│       ├── Yes → GitLab Flow
│       └── No → GitHub Flow
│
├── Every 2-4 weeks (sprint-based)
│   └── Team size > 15?
│       ├── Yes → GitFlow (light) or GitLab Flow
│       └── No → GitHub Flow
│
├── Monthly or quarterly
│   └── Multiple versions in production?
│       ├── Yes → Release Branching
│       └── No → GitFlow
│
└── Ad-hoc / variable
    └── External contributors?
        ├── Yes → Forking Workflow + GitHub Flow
        └── No → GitHub Flow (start simple)
```

### Scoring Matrix

Rate your project on each factor (1-5):

| Factor | Weight | Trunk | GitHub | GitFlow | GitLab | Release |
|--------|--------|-------|--------|---------|--------|---------|
| Deploy frequency | 3 | 5 | 4 | 2 | 3 | 2 |
| Team simplicity | 2 | 3 | 5 | 1 | 3 | 3 |
| Multi-version | 3 | 1 | 1 | 4 | 2 | 5 |
| Release control | 2 | 2 | 3 | 5 | 4 | 5 |
| CI/CD maturity needed | 1 | 5 | 3 | 2 | 3 | 2 |
| Rollback ease | 2 | 5 | 4 | 3 | 4 | 3 |

Multiply score x weight, sum for each strategy, pick the highest.

---

## Migration Guides

### From Nothing → GitHub Flow

```
Week 1:
- Enable branch protection on main
- Require 1 PR review
- Require CI to pass

Week 2:
- Add PR template
- Set up CODEOWNERS
- Configure squash merge as default

Week 3:
- Add commit message convention (commitlint)
- Set up auto-delete branches
- Add PR size labeling

Week 4:
- Refine and document process
- Add merge queue (if needed)
```

### From GitFlow → GitHub Flow

```
Step 1: Merge develop into main (align them)
Step 2: Set main as the deployment branch
Step 3: Delete develop branch
Step 4: Stop creating release branches
Step 5: Update CI to deploy from main
Step 6: Communicate the simpler process
```

### From GitHub Flow → Trunk-Based

```
Step 1: Set up feature flag service
Step 2: Reduce PR size targets (< 100 lines)
Step 3: Require daily merges to main
Step 4: Add merge queue for safety
Step 5: Shift to pair programming (optional)
Step 6: Measure and shorten branch lifetimes
```

---

## Anti-Patterns

### Common Mistakes

| Anti-Pattern | Problem | Fix |
|-------------|---------|-----|
| Long-lived feature branches | Merge conflicts, stale code | Break into smaller PRs, merge daily |
| Direct commits to main | No review, no CI | Enable branch protection |
| Cherry-pick everything | History confusion, double commits | Merge or rebase instead |
| Ignoring merge conflicts | Broken code | Resolve immediately, add tests |
| Too many permanent branches | Complexity, confusion | Reduce to minimum needed |
| No branch naming convention | Can't tell what branch does | Enforce type/description pattern |
| Mixing merge strategies | Inconsistent history | Pick one and configure it |
| No CI on PRs | Bad code reaches main | Require CI checks before merge |
| Manual deployments | Slow, error-prone | Automate deployment from main |
| No cleanup | 500+ stale branches | Auto-delete merged branches |

### Branch Lifetime Warnings

```
Green zone: 0-3 days
  Normal feature branch lifetime
  Low conflict risk

Yellow zone: 3-7 days
  Consider breaking into smaller PRs
  Rebase against main daily

Red zone: 7+ days
  High conflict risk
  Likely too large — split it up
  If unavoidable, rebase frequently

Emergency zone: 30+ days
  Almost certainly needs to be broken up
  Consider using feature flags
  Or create an incremental migration plan
```

### Signs You Need to Change Strategy

```
Your current strategy is wrong if:

Too simple:
- Broken code reaches production regularly
- No one reviews code before merge
- Deployments cause frequent rollbacks
- No audit trail for changes
→ Add more structure (move up complexity)

Too complex:
- Developers spend more time on git than coding
- Release process takes days of manual steps
- Merge conflicts are a daily problem
- New team members take weeks to learn the workflow
→ Simplify (move down complexity)

Complexity scale:
  Trunk-Based < GitHub Flow < GitLab Flow < GitFlow < Release Branching
```
