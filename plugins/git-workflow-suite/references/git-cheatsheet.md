# Git Cheatsheet

## Daily Commands

```bash
# Status and info
git status                    # Working tree status
git status --short            # Compact status
git diff                      # Unstaged changes
git diff --staged             # Staged changes
git log --oneline -10         # Recent commits
git log --oneline --graph     # Visual branch history

# Branching
git branch                    # List local branches
git branch -a                 # List all branches
git branch -d feature/done    # Delete merged branch
git branch -D feature/stale   # Force delete branch
git checkout -b feat/new      # Create and switch
git switch feat/new           # Switch branch (modern)
git switch -c feat/new        # Create and switch (modern)

# Staging
git add file.ts               # Stage specific file
git add src/                  # Stage directory
git add -p                    # Stage interactively (hunk by hunk)
git restore --staged file.ts  # Unstage file
git restore file.ts           # Discard changes

# Committing
git commit -m "message"       # Commit with message
git commit -am "message"      # Stage tracked + commit
git commit --amend            # Amend last commit
git commit --amend --no-edit  # Amend without changing message

# Pushing/Pulling
git push origin branch        # Push branch
git push -u origin branch     # Push + set upstream
git pull origin main          # Pull with merge
git pull --rebase origin main # Pull with rebase
git fetch origin              # Fetch without merge
```

## Undo Reference

| Situation | Command |
|-----------|---------|
| Undo last commit (keep changes staged) | `git reset --soft HEAD~1` |
| Undo last commit (keep changes unstaged) | `git reset HEAD~1` |
| Undo last commit (discard all) | `git reset --hard HEAD~1` |
| Revert a specific commit (safe) | `git revert <sha>` |
| Discard all local changes | `git checkout -- .` |
| Unstage a file | `git restore --staged <file>` |
| Discard changes to a file | `git restore <file>` |
| Recover deleted branch | `git reflog` then `git checkout -b name <sha>` |
| Undo a bad rebase | `git reflog` then `git reset --hard HEAD@{n}` |
| Undo a merge | `git revert -m 1 <merge-sha>` |

## Commit Message Format

```
type(scope): description        # 72 chars max

[body]                          # Explain WHY, not WHAT

[footer]                        # BREAKING CHANGE:, Fixes #123
```

| Type | When |
|------|------|
| `feat` | New feature |
| `fix` | Bug fix |
| `docs` | Documentation only |
| `style` | Formatting, no logic change |
| `refactor` | Code change, no feat/fix |
| `perf` | Performance improvement |
| `test` | Add/update tests |
| `build` | Build system, deps |
| `ci` | CI/CD config |
| `chore` | Maintenance |

## Merge Strategies

| Strategy | Command | Result |
|----------|---------|--------|
| Merge commit | `git merge --no-ff branch` | Preserves full history |
| Squash merge | `git merge --squash branch` | Single clean commit |
| Fast-forward | `git merge --ff-only branch` | Linear history |
| Rebase + FF | `git rebase main && git merge --ff-only` | Linear, rebased |

## Stash Quick Reference

```bash
git stash                     # Stash tracked changes
git stash push -u             # Include untracked files
git stash push -m "label"     # With description
git stash list                # List all stashes
git stash pop                 # Apply + remove latest
git stash apply               # Apply, keep in list
git stash show -p             # Show stash diff
git stash drop stash@{1}      # Remove specific stash
git stash clear               # Remove all stashes
```

## Interactive Rebase

```bash
git rebase -i HEAD~5          # Last 5 commits
git rebase -i main            # Since diverging from main

# pick   = keep commit
# reword = edit message
# squash = merge with previous (keep both messages)
# fixup  = merge with previous (discard this message)
# drop   = delete commit
```

## Bisect (Find Bad Commit)

```bash
git bisect start
git bisect bad                # Current is broken
git bisect good v1.0.0        # This version worked
# Test each checkout, mark good/bad
git bisect good|bad
# When found:
git bisect reset

# Automated:
git bisect start HEAD v1.0.0
git bisect run npm test
```

## Diff Tricks

```bash
git diff --stat               # File-level summary
git diff --name-only          # Just filenames
git diff --word-diff          # Word-level changes
git diff main..HEAD           # Changes since branching
git diff main...HEAD          # Changes on branch only
git diff HEAD~3..HEAD         # Last 3 commits
git diff --cached             # Same as --staged
```

## Log Tricks

```bash
git log --oneline --graph --all         # Visual graph
git log --author="name"                 # By author
git log --since="2 weeks ago"           # Time-based
git log --grep="auth"                   # Search messages
git log -S "functionName"               # Search code changes
git log --follow -- file.ts             # File history (tracks renames)
git log --pretty=format:"%h %an %s"     # Custom format
git log main..feature                   # Commits on feature not on main
git shortlog -sn                        # Commit count by author
```

## Clean Up

```bash
# Remove untracked files (dry run first!)
git clean -n                  # Show what would be deleted
git clean -fd                 # Delete untracked files + dirs
git clean -fX                 # Delete only ignored files

# Prune remote tracking branches
git remote prune origin
git fetch --prune

# Find large files in history
git rev-list --objects --all | git cat-file --batch-check='%(objecttype) %(objectname) %(objectsize) %(rest)' | sort -k3 -rn | head -20
```

## .gitignore Patterns

```bash
# Common patterns
node_modules/
dist/
build/
.env
.env.local
.env.*.local
*.log
.DS_Store
coverage/
.turbo/
.next/
.vercel/
*.tsbuildinfo

# Negate (include despite parent ignore)
!.gitkeep
!important.log
```

## GitHub CLI (gh) Quick Reference

```bash
# PRs
gh pr create --title "..." --body "..."
gh pr list
gh pr view 123
gh pr merge 123 --squash
gh pr checkout 123

# Issues
gh issue create --title "..." --body "..."
gh issue list
gh issue close 123

# Repos
gh repo clone org/repo
gh repo create name --public

# Actions
gh run list
gh run view 12345
gh run watch
```

## Config

```bash
# Identity
git config --global user.name "Name"
git config --global user.email "email@example.com"

# Defaults
git config --global init.defaultBranch main
git config --global pull.rebase true
git config --global push.autoSetupRemote true
git config --global rerere.enabled true

# Performance
git config --global core.fsmonitor true
git config --global core.untrackedCache true

# Show config
git config --list --show-origin
```
