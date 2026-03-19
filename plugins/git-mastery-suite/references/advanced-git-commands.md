# Advanced Git Commands Reference

Complete reference for advanced git commands with practical examples, common patterns,
and gotchas. Organized by operation category for quick lookup.

---

## Table of Contents

- [Commit Operations](#commit-operations)
- [Branch Operations](#branch-operations)
- [Merge and Rebase](#merge-and-rebase)
- [Stash Operations](#stash-operations)
- [Remote Operations](#remote-operations)
- [History and Investigation](#history-and-investigation)
- [Diff and Patch](#diff-and-patch)
- [Reset and Restore](#reset-and-restore)
- [Tag Operations](#tag-operations)
- [Submodule and Subtree](#submodule-and-subtree)
- [Worktree Operations](#worktree-operations)
- [Reflog Operations](#reflog-operations)
- [Bisect Operations](#bisect-operations)
- [Clean and Maintenance](#clean-and-maintenance)
- [Configuration](#configuration)
- [Plumbing Commands](#plumbing-commands)

---

## Commit Operations

### Creating Commits

```bash
# Standard commit
git commit -m "message"

# Multi-line commit message
git commit -m "subject line" -m "body paragraph"

# Commit with HEREDOC (preserves formatting)
git commit -m "$(cat <<'EOF'
feat(auth): add OAuth2 login flow

Implements the full OAuth2 authorization code flow with PKCE.
Supports Google, GitHub, and Microsoft providers.

Closes #123
EOF
)"

# Commit all tracked changes
git commit -am "message"

# Amend last commit (change message)
git commit --amend -m "new message"

# Amend last commit (add staged files, keep message)
git commit --amend --no-edit

# Amend last commit (change author)
git commit --amend --author="Name <email@example.com>"

# Amend last commit (change date)
GIT_COMMITTER_DATE="2024-01-15 10:30:00" git commit --amend --date="2024-01-15 10:30:00" --no-edit

# Create empty commit (useful for triggering CI)
git commit --allow-empty -m "chore: trigger CI rebuild"

# Create fixup commit (for autosquash)
git commit --fixup=abc1234

# Create squash commit (fixup + edit message)
git commit --squash=abc1234

# Sign commit with GPG/SSH
git commit -S -m "signed commit"

# Commit specific hunks interactively
git commit -p
```

### Staging Operations

```bash
# Stage specific file
git add file.ts

# Stage all changes in directory
git add src/

# Stage all tracked file changes
git add -u

# Stage everything (including untracked)
git add -A

# Interactive staging (select hunks)
git add -p
# y = stage hunk, n = skip, s = split, e = edit, q = quit

# Stage with intent to add (track file without content)
git add -N file.ts

# Unstage file
git restore --staged file.ts
# OR (older syntax)
git reset HEAD file.ts

# Unstage everything
git restore --staged .

# Discard working tree changes
git restore file.ts
# OR (older syntax)
git checkout -- file.ts

# Discard all working tree changes
git restore .
```

## Branch Operations

### Creating and Switching

```bash
# Create and switch to new branch
git checkout -b feature/name
# OR (newer syntax)
git switch -c feature/name

# Create branch from specific commit
git checkout -b feature/name abc1234
git switch -c feature/name abc1234

# Create branch from tag
git checkout -b release/v2.0 v2.0.0

# Create branch from remote branch
git checkout -b local-name origin/remote-name
git switch -c local-name origin/remote-name

# Switch to existing branch
git checkout main
git switch main

# Switch to previous branch
git checkout -
git switch -

# Create branch without switching
git branch new-branch
git branch new-branch abc1234

# Rename current branch
git branch -m new-name

# Rename any branch
git branch -m old-name new-name

# Copy a branch
git branch -c source-branch copy-branch
```

### Listing and Filtering

```bash
# List local branches
git branch

# List all branches (local + remote)
git branch -a

# List remote branches
git branch -r

# List branches with last commit info
git branch -v
git branch -vv  # Include tracking info

# List merged branches
git branch --merged main

# List unmerged branches
git branch --no-merged main

# List branches sorted by last commit
git branch --sort=-committerdate

# List branches containing a specific commit
git branch --contains abc1234

# List branches matching pattern
git branch --list "feature/*"

# Show branch in a graph
git log --oneline --graph --all --decorate
```

### Deleting Branches

```bash
# Delete merged branch
git branch -d feature/done

# Force delete unmerged branch
git branch -D feature/abandoned

# Delete remote branch
git push origin --delete feature/done
# OR
git push origin :feature/done

# Delete all merged local branches (except main/develop)
git branch --merged main | grep -vE '^\*|main|develop' | xargs git branch -d

# Prune stale remote-tracking branches
git remote prune origin
# OR during fetch
git fetch --prune
```

## Merge and Rebase

### Merge Operations

```bash
# Standard merge (fast-forward if possible)
git merge feature-branch

# Force merge commit (no fast-forward)
git merge --no-ff feature-branch

# Fast-forward only (fail if not possible)
git merge --ff-only feature-branch

# Squash merge (combine all commits, don't auto-commit)
git merge --squash feature-branch

# Merge with specific strategy
git merge -s recursive feature-branch
git merge -s ours abandoned-branch  # Discard their changes

# Merge with strategy option
git merge -X theirs feature-branch  # Prefer their changes on conflict
git merge -X ours feature-branch    # Prefer our changes on conflict

# Abort merge
git merge --abort

# Continue merge after conflict resolution
git add resolved-file.ts
git merge --continue
# OR
git commit

# Check if branch can merge cleanly
git merge --no-commit --no-ff feature-branch && git merge --abort
```

### Rebase Operations

```bash
# Rebase current branch onto main
git rebase main

# Rebase onto specific commit
git rebase abc1234

# Interactive rebase (last N commits)
git rebase -i HEAD~5

# Interactive rebase onto branch
git rebase -i main

# Rebase with autosquash (process fixup!/squash! commits)
git rebase -i --autosquash main

# Rebase preserving merges
git rebase --rebase-merges main

# Rebase onto (move branch base)
git rebase --onto new-base old-base branch
# Example: Move feature2 from feature1 base to main
git rebase --onto main feature1 feature2

# Abort rebase
git rebase --abort

# Continue rebase
git rebase --continue

# Skip current commit
git rebase --skip

# Rebase with autostash
git rebase --autostash main
```

### Cherry-Pick

```bash
# Cherry-pick single commit
git cherry-pick abc1234

# Cherry-pick without committing
git cherry-pick --no-commit abc1234

# Cherry-pick range (exclusive start)
git cherry-pick abc1234..def5678

# Cherry-pick range (inclusive)
git cherry-pick abc1234^..def5678

# Cherry-pick with mainline (for merge commits)
git cherry-pick -m 1 <merge-commit>

# Abort cherry-pick
git cherry-pick --abort

# Continue after conflict
git cherry-pick --continue

# Skip current commit
git cherry-pick --skip
```

## Stash Operations

```bash
# Stash working directory changes
git stash
git stash push

# Stash with message
git stash push -m "WIP: feature implementation"

# Stash specific files
git stash push -m "models only" src/models/

# Stash including untracked files
git stash push -u

# Stash including ignored files
git stash push -a

# Stash interactively (select hunks)
git stash push -p

# List stashes
git stash list

# Show stash diff
git stash show              # Summary
git stash show -p            # Full diff
git stash show stash@{2}     # Specific stash

# Apply stash (keep in list)
git stash apply
git stash apply stash@{2}

# Pop stash (remove from list)
git stash pop
git stash pop stash@{2}

# Drop specific stash
git stash drop stash@{1}

# Clear all stashes
git stash clear

# Create branch from stash
git stash branch new-branch stash@{0}
```

## Remote Operations

```bash
# List remotes
git remote -v

# Add remote
git remote add upstream https://github.com/original/repo.git

# Remove remote
git remote remove upstream

# Rename remote
git remote rename origin github

# Change remote URL
git remote set-url origin git@github.com:user/repo.git

# Fetch from remote
git fetch origin
git fetch --all
git fetch --prune  # Remove stale tracking branches

# Pull (fetch + merge)
git pull origin main
git pull --rebase origin main  # Fetch + rebase
git pull --ff-only             # Only if fast-forward possible

# Push
git push origin main
git push -u origin feature-branch  # Set upstream tracking
git push --force-with-lease        # Safe force push
git push --tags                    # Push all tags
git push origin --delete branch    # Delete remote branch

# Show remote info
git remote show origin

# Track remote branch
git branch --set-upstream-to=origin/main main

# Push to multiple remotes
git remote add all-remotes https://github.com/user/repo.git
git remote set-url --add --push all-remotes https://gitlab.com/user/repo.git
git remote set-url --add --push all-remotes https://github.com/user/repo.git
git push all-remotes main
```

## History and Investigation

### Log Operations

```bash
# Basic log
git log --oneline
git log --oneline -20  # Last 20 commits

# Detailed log
git log --stat         # Show files changed
git log -p             # Show patches (diffs)

# Graph view
git log --oneline --graph --all --decorate

# Custom format
git log --format="%h %an %ar %s"
git log --format="%C(yellow)%h%Creset %C(blue)%an%Creset %C(green)%ar%Creset %s"

# Filter by author
git log --author="name"

# Filter by date
git log --since="2024-01-01" --until="2024-06-30"
git log --after="2 weeks ago"

# Filter by message
git log --grep="fix" --grep="auth" --all-match

# Filter by file
git log -- src/api/auth.ts
git log --follow -- src/api/auth.ts  # Follow renames

# Filter by change content (pickaxe)
git log -S "apiKey"                  # String search
git log -G "apiKey.*=.*process"      # Regex search

# Filter by diff
git log --diff-filter=A -- "*.ts"    # Added files
git log --diff-filter=D -- "*.ts"    # Deleted files
git log --diff-filter=M -- "*.ts"    # Modified files

# Commits between refs
git log main..feature               # In feature, not in main
git log main...feature              # In either, not in both

# Shortlog (summary by author)
git shortlog -sn                    # Commit counts
git shortlog -sn --all              # All branches
git shortlog -sn --since="6 months ago"
```

### Blame

```bash
# Basic blame
git blame src/file.ts

# Blame specific lines
git blame -L 50,80 src/file.ts

# Ignore whitespace
git blame -w src/file.ts

# Detect moved/copied lines
git blame -C src/file.ts       # Detect within same commit
git blame -C -C src/file.ts    # Detect across commits
git blame -C -C -C src/file.ts # Detect across all files

# Blame at specific revision
git blame abc1234 -- src/file.ts

# Show email instead of name
git blame -e src/file.ts

# Use .git-blame-ignore-revs
git blame --ignore-revs-file .git-blame-ignore-revs src/file.ts
```

## Diff and Patch

```bash
# Working tree vs staged
git diff

# Staged vs HEAD
git diff --staged
git diff --cached  # Same as --staged

# Working tree vs HEAD
git diff HEAD

# Between commits
git diff abc1234..def5678
git diff abc1234 def5678

# Between branches
git diff main..feature
git diff main...feature  # Since diverge point

# Specific file
git diff -- src/file.ts

# Stats only
git diff --stat
git diff --shortstat

# Word diff (inline changes)
git diff --word-diff
git diff --color-words

# Name only
git diff --name-only
git diff --name-status  # With status (A/M/D)

# Create patch
git diff > changes.patch
git format-patch main..feature  # One patch per commit
git format-patch -3              # Last 3 commits

# Apply patch
git apply changes.patch
git am < patch-file.patch        # Apply format-patch output
git am --3way < patch-file.patch # Three-way merge on conflict
```

## Reset and Restore

```bash
# Reset HEAD (keep changes staged)
git reset --soft HEAD~1

# Reset HEAD (keep changes unstaged)
git reset HEAD~1
git reset --mixed HEAD~1  # Same

# Reset HEAD (discard changes - DESTRUCTIVE)
git reset --hard HEAD~1

# Reset to specific commit
git reset --hard abc1234

# Reset single file
git reset HEAD -- file.ts

# Restore file from commit
git restore --source=abc1234 -- file.ts

# Restore file from another branch
git restore --source=main -- file.ts

# Restore deleted file
git checkout HEAD~1 -- deleted-file.ts

# Revert commit (safe undo - creates new commit)
git revert abc1234

# Revert merge commit
git revert -m 1 <merge-commit-hash>

# Revert range
git revert abc1234..def5678

# Revert without committing
git revert --no-commit abc1234
```

## Tag Operations

```bash
# Create annotated tag (recommended)
git tag -a v1.0.0 -m "Release v1.0.0"

# Create lightweight tag
git tag v1.0.0

# Tag specific commit
git tag -a v1.0.0 abc1234 -m "Release v1.0.0"

# List tags
git tag
git tag -l "v1.*"
git tag -l --sort=-creatordate  # Newest first
git tag -n                       # With annotations

# Show tag details
git show v1.0.0

# Push single tag
git push origin v1.0.0

# Push all tags
git push origin --tags

# Delete local tag
git tag -d v1.0.0

# Delete remote tag
git push origin --delete v1.0.0

# Checkout tag
git checkout v1.0.0              # Detached HEAD
git checkout -b branch v1.0.0    # New branch from tag

# Verify signed tag
git tag -v v1.0.0

# Find tag containing commit
git tag --contains abc1234
```

## Submodule and Subtree

### Submodule Commands

```bash
# Add submodule
git submodule add URL path
git submodule add -b branch URL path

# Initialize submodules
git submodule init
git submodule update
git submodule update --init --recursive

# Clone with submodules
git clone --recurse-submodules URL

# Update submodule to latest
git submodule update --remote
git submodule update --remote --merge

# Status
git submodule status
git submodule summary

# Run command in all submodules
git submodule foreach 'git pull origin main'
git submodule foreach --recursive 'git status'

# Remove submodule
git submodule deinit path
git rm path
rm -rf .git/modules/path
```

### Subtree Commands

```bash
# Add subtree
git subtree add --prefix=lib/name URL branch --squash

# Pull updates
git subtree pull --prefix=lib/name URL branch --squash

# Push changes back
git subtree push --prefix=lib/name URL branch

# Split into branch
git subtree split --prefix=lib/name -b split-branch
```

## Worktree Operations

```bash
# Add worktree
git worktree add ../path branch
git worktree add ../path -b new-branch

# Add detached worktree
git worktree add --detach ../path HEAD

# List worktrees
git worktree list

# Remove worktree
git worktree remove ../path

# Prune stale entries
git worktree prune

# Lock worktree (prevent pruning)
git worktree lock ../path

# Unlock worktree
git worktree unlock ../path
```

## Reflog Operations

```bash
# Show HEAD reflog
git reflog
git reflog show HEAD

# Show branch reflog
git reflog show main

# Reflog with dates
git reflog --date=relative
git reflog --date=iso

# Reflog for specific ref
git reflog show stash

# Expire old reflog entries
git reflog expire --expire=90.days.ago --all

# Delete all reflog entries
git reflog expire --expire=now --all
```

## Bisect Operations

```bash
# Start bisect
git bisect start

# Mark commits
git bisect bad                 # Current is bad
git bisect good v1.0.0         # v1.0.0 was good
git bisect good abc1234

# Automated bisect
git bisect run npm test
git bisect run ./test-script.sh

# Skip untestable commit
git bisect skip

# View log
git bisect log

# Replay session
git bisect replay logfile

# Visualize
git bisect visualize

# End bisect
git bisect reset
```

## Clean and Maintenance

```bash
# Preview files to clean (dry run)
git clean -n

# Remove untracked files
git clean -f

# Remove untracked files and directories
git clean -fd

# Remove ignored files too
git clean -fdx

# Interactive clean
git clean -i

# Garbage collection
git gc
git gc --aggressive --prune=now

# Verify integrity
git fsck
git fsck --full

# Repack
git repack -a -d

# Enable maintenance
git maintenance start
git maintenance run --task=gc

# Count objects
git count-objects -vH

# Prune unreachable objects
git prune
git prune --dry-run
```

## Configuration

```bash
# List all config
git config --list
git config --list --show-origin  # Show where each setting is defined

# Set config
git config --global user.name "Name"
git config --local core.autocrlf input

# Unset config
git config --unset key

# Edit config file directly
git config --global --edit

# Useful settings
git config --global init.defaultBranch main
git config --global pull.rebase true
git config --global push.autoSetupRemote true
git config --global rebase.autoSquash true
git config --global rebase.autoStash true
git config --global merge.conflictstyle zdiff3
git config --global diff.algorithm histogram
git config --global rerere.enabled true
git config --global core.fsmonitor true
git config --global core.untrackedcache true
git config --global fetch.prune true
```

## Plumbing Commands

Low-level commands for scripting and understanding git internals:

```bash
# Show object type
git cat-file -t abc1234

# Show object content
git cat-file -p abc1234

# Show tree object
git ls-tree HEAD
git ls-tree -r HEAD  # Recursive

# Show file at specific revision
git show main:src/file.ts

# Find commits that modify a file
git rev-list HEAD -- src/file.ts

# Get current commit hash
git rev-parse HEAD
git rev-parse --short HEAD

# Get branch name
git rev-parse --abbrev-ref HEAD
git symbolic-ref --short HEAD

# Check if inside git repo
git rev-parse --is-inside-work-tree

# Get repo root
git rev-parse --show-toplevel

# Hash an object without storing
echo "content" | git hash-object --stdin

# List tracked files
git ls-files
git ls-files -m  # Modified
git ls-files -d  # Deleted
git ls-files -o  # Untracked
git ls-files -s  # With staging info

# Show merge base
git merge-base main feature

# Check if commit is ancestor
git merge-base --is-ancestor abc1234 def5678 && echo "yes" || echo "no"

# Show refnames for a commit
git name-rev abc1234

# Verify commit signature
git verify-commit abc1234

# Verify tag signature
git verify-tag v1.0.0
```
