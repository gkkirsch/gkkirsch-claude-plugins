---
name: git-mastery
description: >
  Git mastery command — analyzes your repository, recommends branching strategies, sets up PR workflows,
  configures monorepo tooling, performs advanced git operations, and integrates CI/CD with git.
  Triggers: "/git-mastery", "git workflow", "branching strategy", "monorepo setup", "git recovery",
  "git ci", "fix my git".
user-invocable: true
argument-hint: "<workflow|monorepo|ops|ci> [--strategy github-flow|gitflow|trunk] [--tool turbo|nx|lerna]"
allowed-tools: Read, Write, Edit, Bash, Glob, Grep
model: sonnet
---

# /git-mastery Command

One-command git mastery. Analyzes your repository, recommends strategies, and dispatches the right
specialist agent for branching workflows, monorepo architecture, git operations, or CI/CD integration.

## Usage

```
/git-mastery                            # Auto-detect and recommend
/git-mastery workflow                   # Set up branching strategy and PR workflow
/git-mastery workflow --strategy trunk  # Use trunk-based development
/git-mastery monorepo                   # Set up or optimize monorepo
/git-mastery monorepo --tool turbo      # Use Turborepo specifically
/git-mastery ops                        # Advanced git operation or recovery
/git-mastery ci                         # Set up CI/CD git integration
```

## Subcommands

| Subcommand | Agent | Description |
|------------|-------|-------------|
| `workflow` | git-workflow-expert | Branching strategies, PR workflows, branch protection |
| `monorepo` | monorepo-architect | Nx, Turborepo, Lerna, workspace management |
| `ops` | git-operations-specialist | Rebase, cherry-pick, bisect, reflog, recovery |
| `ci` | ci-git-integrator | GitHub Actions, git hooks, conventional commits, releases |

## Procedure

### Step 1: Analyze Repository

Read the project to understand the git setup:

1. **Check git configuration:**
   ```bash
   git branch -a                  # All branches
   git remote -v                  # Remotes
   git log --oneline -20          # Recent history
   git tag -l --sort=-creatordate | head -10  # Recent tags
   ```

2. **Check for existing CI:**
   ```
   Glob: .github/workflows/*.yml, .gitlab-ci.yml, Jenkinsfile,
         .circleci/config.yml, bitbucket-pipelines.yml
   ```

3. **Check for git tooling:**
   ```
   Glob: .husky/*, .commitlintrc*, .releaserc*, cliff.toml,
         .git-blame-ignore-revs, .gitattributes, .gitmodules
   ```

4. **Check for monorepo indicators:**
   ```
   Glob: turbo.json, nx.json, lerna.json, pnpm-workspace.yaml
   ```

5. **Check project structure:**
   ```
   Glob: package.json, apps/*/package.json, packages/*/package.json
   ```

Report findings:

```
Repository Analysis:
- Branches: 12 (main + 11 feature branches)
- Strategy: GitHub Flow (informal)
- CI: GitHub Actions (basic lint + test)
- Monorepo: No (single package)
- Hooks: None configured
- Commit convention: Inconsistent
- Tags: v1.0.0 - v1.5.2 (SemVer)
- Contributors: 4 active
```

### Step 2: Route to Agent

Based on the subcommand, dispatch the appropriate agent:

#### `workflow` — Git Workflow Expert

```
Task tool:
  subagent_type: "git-workflow-expert"
  mode: "bypassPermissions"
  prompt: |
    Set up a git branching strategy and PR workflow.
    Current setup: [detected configuration]
    Team size: [estimated from contributors]
    Strategy requested: [if --strategy flag provided]
    Goals:
    - Configure appropriate branching strategy
    - Set up branch protection rules
    - Create PR template and CODEOWNERS
    - Configure merge strategy
    - Set up commit message conventions
    - Add automated PR checks
```

#### `monorepo` — Monorepo Architect

```
Task tool:
  subagent_type: "monorepo-architect"
  mode: "bypassPermissions"
  prompt: |
    Set up or optimize monorepo architecture.
    Current setup: [detected monorepo config]
    Tool requested: [if --tool flag provided]
    Package structure: [detected apps/packages]
    Goals:
    - Choose and configure monorepo tool
    - Set up workspace dependencies
    - Configure build pipeline and caching
    - Set up affected-only CI
    - Configure shared tooling packages
```

#### `ops` — Git Operations Specialist

```
Task tool:
  subagent_type: "git-operations-specialist"
  mode: "bypassPermissions"
  prompt: |
    Perform advanced git operation or recovery.
    Current state: [git status, log, reflog]
    Problem: [user's description]
    Goals:
    - Diagnose the current state
    - Plan the safest recovery/operation
    - Execute with safety backups
    - Verify the result
```

#### `ci` — CI/CD Git Integrator

```
Task tool:
  subagent_type: "ci-git-integrator"
  mode: "bypassPermissions"
  prompt: |
    Set up CI/CD git integration.
    CI platform: [detected or requested]
    Current CI: [existing configuration]
    Project type: [detected stack]
    Goals:
    - Configure CI pipeline with proper git triggers
    - Set up conventional commit validation
    - Configure automated releases
    - Add changelog generation
    - Optimize with caching and affected-only builds
```

### Step 3: Results Summary

After the agent completes, present results:

**For workflow:**
```
Git Workflow Setup Complete:
- Strategy: GitHub Flow
- Branch protection: Configured on main (2 reviewers, CI required)
- PR template: .github/pull_request_template.md
- CODEOWNERS: .github/CODEOWNERS
- Merge strategy: Squash merge
- Commit convention: Conventional Commits (commitlint configured)
- Files created/modified: [list]
```

**For monorepo:**
```
Monorepo Setup Complete:
- Tool: Turborepo
- Packages: 3 apps + 4 shared packages
- Build pipeline: build → lint → test (with caching)
- Remote cache: Configured (Vercel)
- CI: Affected-only builds in PRs
- Files created/modified: [list]
```

**For ops:**
```
Git Operation Complete:
- Operation: Recovered 3 lost commits from failed rebase
- Before: HEAD at abc1234 (after bad rebase)
- After: HEAD at def5678 (all commits restored)
- Backup branch: backup-1234567890 (safe to delete)
- Verification: All tests pass, history is clean
```

**For ci:**
```
CI/CD Setup Complete:
- Platform: GitHub Actions
- PR pipeline: lint → typecheck → test → build → security (5 min avg)
- Deploy pipeline: test → staging → production (with approval)
- Release: Semantic release with conventional commits
- Changelog: git-cliff configured
- Hooks: Husky + lint-staged + commitlint
- Files created/modified: [list]
```

## Error Recovery

| Error | Cause | Fix |
|-------|-------|-----|
| Can't determine strategy | New/empty repo | Specify --strategy flag |
| Branch protection fails | Insufficient permissions | Need admin access |
| Monorepo tool conflict | Multiple tools detected | Specify --tool flag |
| Recovery impossible | Reflog expired | Check remote for backup |
| CI setup fails | Missing secrets | Configure secrets first |

## Notes

- Always review branch protection changes before applying
- Monorepo setup may require dependency reinstallation
- Git recovery operations create backup branches — clean up when satisfied
- CI changes should be tested on a feature branch first
- Semantic release requires specific branch naming to work correctly
