---
name: git-advanced
description: >
  Advanced git techniques — interactive rebase, bisect debugging, worktrees,
  cherry-pick, stash management, reflog recovery, submodules, and performance.
  Triggers: "git rebase", "git bisect", "git worktree", "git cherry-pick",
  "git stash", "git reflog", "git submodule", "git reset", "undo git".
  NOT for: branching strategy (use branching-strategies), commit format (use conventional-commits).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Advanced Git Techniques

## Interactive Rebase

```bash
# Rebase last 5 commits
git rebase -i HEAD~5

# Rebase onto main
git rebase -i main

# Commands in the editor:
# pick   — keep commit as-is
# reword — keep commit, edit message
# edit   — pause at commit for changes
# squash — meld into previous commit (keep message)
# fixup  — meld into previous commit (discard message)
# drop   — remove commit entirely

# Example: Clean up feature branch before PR
# Original:
#   pick abc123 feat: add login form
#   pick def456 fix typo
#   pick ghi789 more fixes
#   pick jkl012 feat: add auth middleware
#
# Cleaned:
#   pick abc123 feat: add login form
#   fixup def456 fix typo
#   fixup ghi789 more fixes
#   pick jkl012 feat: add auth middleware

# Auto-squash fixup commits
git commit --fixup abc123
git rebase -i --autosquash main
```

## Git Bisect (Find Bug-Introducing Commit)

```bash
# Start bisect
git bisect start

# Mark current commit as bad
git bisect bad

# Mark a known good commit
git bisect good v1.0.0
# or: git bisect good abc123

# Git checks out a middle commit — test it, then:
git bisect good  # if this commit works
git bisect bad   # if this commit has the bug

# Repeat until git finds the first bad commit

# End bisect
git bisect reset

# Automated bisect with a test script
git bisect start HEAD v1.0.0
git bisect run npm test
# Git automatically finds the commit that broke tests
```

## Git Worktrees

```bash
# Create a worktree for parallel work
git worktree add ../feature-auth feat/auth
# Creates a checkout at ../feature-auth on the feat/auth branch

# Create worktree with new branch
git worktree add -b hotfix/urgent ../hotfix main
# New branch 'hotfix/urgent' from main, checked out at ../hotfix

# List worktrees
git worktree list

# Remove worktree
git worktree remove ../feature-auth

# Prune stale worktrees
git worktree prune

# Use case: Review a PR while working on something else
git worktree add ../pr-review pr-branch
cd ../pr-review
# Review, test, come back to main worktree
```

## Cherry-Pick

```bash
# Apply a specific commit to current branch
git cherry-pick abc123

# Cherry-pick without committing (stage only)
git cherry-pick --no-commit abc123

# Cherry-pick a range of commits
git cherry-pick abc123..def456

# Cherry-pick from another branch
git cherry-pick feature/branch~3  # 3 commits back from branch tip

# Resolve conflicts during cherry-pick
git cherry-pick abc123
# (resolve conflicts)
git add .
git cherry-pick --continue

# Abort cherry-pick
git cherry-pick --abort
```

## Stash Management

```bash
# Stash current changes
git stash
git stash push -m "WIP: auth middleware"

# Stash specific files
git stash push -m "partial work" src/auth.ts src/middleware.ts

# Stash including untracked files
git stash push -u -m "including new files"

# Stash including ignored files
git stash push -a -m "everything"

# List stashes
git stash list

# Apply most recent stash (keep in stash list)
git stash apply

# Apply and remove from stash list
git stash pop

# Apply specific stash
git stash apply stash@{2}

# Show stash contents
git stash show -p stash@{0}

# Create branch from stash
git stash branch new-feature stash@{0}

# Drop a specific stash
git stash drop stash@{1}

# Clear all stashes
git stash clear
```

## Reflog (Undo Almost Anything)

```bash
# Show reflog (history of HEAD movements)
git reflog

# Recover a deleted branch
git reflog
# Find the commit: abc123 HEAD@{5}: checkout: moving from deleted-branch to main
git checkout -b recovered-branch abc123

# Undo a hard reset
git reflog
# Find the commit before reset: def456 HEAD@{3}: commit: important work
git reset --hard def456

# Recover after a bad rebase
git reflog
# Find pre-rebase state: ghi789 HEAD@{7}: rebase (start)
git reset --hard HEAD@{8}  # One before the rebase started

# Recover a dropped stash
git fsck --no-reflog | grep commit
# Find the dangling commit and cherry-pick it
```

## Undo Operations

```bash
# Undo last commit (keep changes staged)
git reset --soft HEAD~1

# Undo last commit (keep changes unstaged)
git reset HEAD~1

# Undo last commit (discard changes — DANGEROUS)
git reset --hard HEAD~1

# Undo a specific commit (create reverse commit)
git revert abc123

# Undo a merge commit
git revert -m 1 abc123
# -m 1 means keep the first parent (usually main)

# Unstage a file
git restore --staged file.ts

# Discard changes to a file
git restore file.ts

# Restore a deleted file from a specific commit
git checkout abc123 -- path/to/file.ts
```

## Submodules

```bash
# Add a submodule
git submodule add https://github.com/org/repo.git libs/repo

# Clone repo with submodules
git clone --recurse-submodules https://github.com/org/main-repo.git

# Initialize submodules after clone
git submodule update --init --recursive

# Update submodule to latest
cd libs/repo
git pull origin main
cd ../..
git add libs/repo
git commit -m "chore: update repo submodule"

# Update all submodules
git submodule update --remote --merge

# Remove a submodule
git submodule deinit libs/repo
git rm libs/repo
rm -rf .git/modules/libs/repo
```

## Git Aliases

```bash
# Add useful aliases
git config --global alias.lg "log --oneline --graph --decorate -20"
git config --global alias.st "status --short"
git config --global alias.co "checkout"
git config --global alias.br "branch"
git config --global alias.unstage "restore --staged"
git config --global alias.last "log -1 HEAD --stat"
git config --global alias.diff-words "diff --word-diff"
git config --global alias.branches "branch -a --sort=-committerdate"
git config --global alias.stashes "stash list"
git config --global alias.whoami "config user.email"
```

## Performance for Large Repos

```bash
# Partial clone (download objects on demand)
git clone --filter=blob:none https://github.com/org/large-repo.git

# Sparse checkout (only check out specific paths)
git sparse-checkout init --cone
git sparse-checkout set src/my-package docs

# Shallow clone (limited history)
git clone --depth 1 https://github.com/org/repo.git

# Deepen later if needed
git fetch --deepen 100

# Speed up git status
git config core.fsmonitor true          # File system monitor
git config core.untrackedCache true     # Cache untracked files
git config feature.manyFiles true       # Optimize for large repos
```

## Gotchas

1. **Never rebase commits that others have based work on.** Rebasing rewrites commit SHAs. If someone branched off your original commits, their branch will diverge. Only rebase local, unpushed commits.

2. **`git reset --hard` is not in the reflog permanently.** Reflog entries expire (default: 90 days for reachable, 30 for unreachable). If you need to recover something, do it soon.

3. **Cherry-pick creates a NEW commit with a different SHA.** The cherry-picked commit is a copy, not a move. If you later merge the source branch, git might try to apply the change again.

4. **Submodules are pinned to a specific commit.** They don't auto-update. You must explicitly update and commit the new submodule reference. Most modern projects prefer monorepos or package managers over submodules.

5. **`git stash` doesn't stash untracked files by default.** Use `git stash push -u` to include untracked files, or `-a` to include ignored files too.

6. **Worktrees share the same `.git` directory.** You can't check out the same branch in two worktrees simultaneously. Each worktree must be on a different branch.
