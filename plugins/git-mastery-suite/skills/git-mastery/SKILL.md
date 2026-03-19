---
name: git-mastery-suite
description: >
  Git & Version Control Mastery Suite — expert-level agents for branching strategies and PR workflows,
  monorepo architecture with Nx/Turborepo/Lerna, advanced git operations and recovery, and CI/CD git
  integration with GitHub Actions and conventional commits.
  Triggers: "git workflow", "branching strategy", "gitflow", "trunk-based", "github flow", "pr workflow",
  "branch protection", "merge strategy", "monorepo", "nx", "turborepo", "lerna", "workspace",
  "pnpm workspaces", "monorepo setup", "git rebase", "git cherry-pick", "git bisect", "git reflog",
  "git recovery", "lost commits", "fix git", "undo git", "git stash", "git worktree", "git lfs",
  "git hooks", "husky", "commitlint", "conventional commits", "semantic release", "changelog",
  "ci cd git", "github actions", "gitlab ci", "merge queue", "release automation",
  "git submodule", "git subtree", "git operations".
  Dispatches the appropriate specialist agent: git-workflow-expert, monorepo-architect,
  git-operations-specialist, or ci-git-integrator.
  NOT for: Non-git version control (SVN, Mercurial), general DevOps without git focus,
  IDE-specific git GUIs, or basic git tutorials (add/commit/push).
version: 1.0.0
argument-hint: "<workflow|monorepo|ops|ci> [options]"
user-invocable: true
allowed-tools: Read, Grep, Glob, Bash
model: sonnet
---

# Git & Version Control Mastery Suite

Production-grade git workflow and monorepo agents for Claude Code. Four specialist agents covering
branching strategies, monorepo architecture, advanced git operations, and CI/CD integration — the
version control mastery that every development team needs.

## Available Agents

### Git Workflow Expert (`git-workflow-expert`)
Designs and implements branching strategies and PR workflows. Supports trunk-based development,
GitHub Flow, GitFlow, GitLab Flow, and release branching. Configures branch protection rules,
CODEOWNERS, PR templates, merge strategies, commit conventions, and review processes.

**Invoke**: Dispatch via Task tool with `subagent_type: "git-workflow-expert"`.

**Example prompts**:
- "Set up a branching strategy for our team of 8 developers"
- "Migrate from GitFlow to GitHub Flow"
- "Configure branch protection and PR reviews on our repo"
- "Set up conventional commits with commitlint and Husky"
- "Design a PR workflow with CODEOWNERS and automated checks"

### Monorepo Architect (`monorepo-architect`)
Designs and manages monorepo structures with Turborepo, Nx, Lerna, and pnpm workspaces.
Handles dependency graphs, build orchestration, affected-based CI, package publishing,
shared configuration packages, and Docker builds in monorepos.

**Invoke**: Dispatch via Task tool with `subagent_type: "monorepo-architect"`.

**Example prompts**:
- "Set up a Turborepo monorepo with a Next.js app and shared UI library"
- "Convert our polyrepo into a monorepo with Nx"
- "Configure affected-only CI for our monorepo"
- "Set up Changesets for publishing our npm packages"
- "Optimize our monorepo build performance with caching"

### Git Operations Specialist (`git-operations-specialist`)
Handles advanced git operations and recovery. Interactive rebase, cherry-pick, bisect, reflog,
stash management, worktrees, submodules, blame analysis, history rewriting, sensitive data
removal, and Git LFS configuration.

**Invoke**: Dispatch via Task tool with `subagent_type: "git-operations-specialist"`.

**Example prompts**:
- "I lost commits after a rebase — help me recover them"
- "Use git bisect to find which commit introduced this bug"
- "Remove a secret that was accidentally committed to history"
- "Set up Git LFS for our large assets"
- "Help me squash and reorder my last 10 commits"
- "Set up git worktrees for parallel development"

### CI/CD Git Integrator (`ci-git-integrator`)
Configures CI/CD pipelines deeply integrated with git workflows. GitHub Actions with branch
triggers, conventional commit validation, semantic release, changelog generation, merge queues,
affected-only monorepo CI, and deployment promotion strategies.

**Invoke**: Dispatch via Task tool with `subagent_type: "ci-git-integrator"`.

**Example prompts**:
- "Set up GitHub Actions with PR checks and automatic deployment"
- "Configure semantic-release with conventional commits"
- "Set up a merge queue for our main branch"
- "Add affected-only CI for our Turborepo monorepo"
- "Configure tag-based releases with changelog generation"

## Quick Start: /git-mastery

Use the `/git-mastery` command for guided git mastery:

```
/git-mastery                            # Analyze repo and recommend
/git-mastery workflow                   # Set up branching strategy
/git-mastery workflow --strategy trunk  # Trunk-based development
/git-mastery monorepo                   # Set up monorepo tooling
/git-mastery monorepo --tool turbo      # Use Turborepo
/git-mastery ops                        # Git operations / recovery
/git-mastery ci                         # CI/CD git integration
```

The `/git-mastery` command auto-detects your repository setup and routes to the right agent.

## Agent Selection Guide

| Need | Agent | Prompt Example |
|------|-------|----------------|
| Choose a branching model | git-workflow-expert | "Which strategy for our team?" |
| Set up PR reviews | git-workflow-expert | "Configure branch protection" |
| Commit conventions | git-workflow-expert | "Set up conventional commits" |
| Set up monorepo | monorepo-architect | "Create Turborepo workspace" |
| Optimize monorepo CI | monorepo-architect | "Affected-only builds" |
| Publish npm packages | monorepo-architect | "Set up Changesets" |
| Recover lost work | git-operations-specialist | "Lost commits after rebase" |
| Find bug introduction | git-operations-specialist | "Which commit broke tests?" |
| Clean git history | git-operations-specialist | "Remove secrets from history" |
| Set up GitHub Actions | ci-git-integrator | "PR pipeline with deploy" |
| Automate releases | ci-git-integrator | "Semantic release setup" |
| Changelog generation | ci-git-integrator | "Auto-generate changelog" |

## Reference Materials

This skill includes comprehensive reference documents in `references/`:

- **advanced-git-commands.md** — Complete git command reference organized by category: commits,
  branches, merges, stashes, remotes, history investigation, diffs, resets, tags, submodules,
  worktrees, reflog, bisect, maintenance, and plumbing commands
- **branching-strategies.md** — In-depth comparison of trunk-based, GitHub Flow, GitFlow, GitLab
  Flow, release branching, and forking workflows with decision frameworks and migration guides
- **monorepo-tools.md** — Detailed comparison of Turborepo, Nx, Lerna, pnpm/Yarn/npm workspaces,
  and Bazel/Pants with feature matrices, benchmarks, and migration paths

Agents automatically consult these references when working. You can also read them directly for
quick answers.

## How It Works

1. You describe your git challenge (workflow, monorepo, operation, or CI)
2. The SKILL.md analyzes your repository and routes to the appropriate agent
3. The agent reads your codebase, understands your setup, and implements the solution
4. Configuration files, workflows, and documentation are created
5. The agent provides results and next steps

All generated artifacts follow industry best practices:
- Workflows: Branch protection, PR templates, CODEOWNERS, merge strategies
- Monorepos: Incremental builds, caching, affected-only CI, dependency management
- Operations: Safety-first recovery, minimal intervention, reflog-based undo
- CI/CD: Fast feedback, cached builds, conventional commits, semantic versioning
