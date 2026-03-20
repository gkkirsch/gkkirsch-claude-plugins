---
name: git-advanced
description: >
  Advanced git operations — interactive rebase, bisect, worktrees, reflog,
  cherry-pick, stash management, submodules, and recovery techniques.
  Triggers: "interactive rebase", "git bisect", "git worktree", "git reflog",
  "cherry-pick", "git stash", "git recovery", "git submodule", "git advanced".
  NOT for: Branching strategies or workflows (use git-workflows).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Advanced Git Operations

## Interactive Rebase

```bash
# Rewrite last 5 commits
git rebase -i HEAD~5

# Rebase onto a branch
git rebase -i main
```

### Rebase Commands

| Command | Short | What It Does |
|---------|-------|-------------|
| `pick` | `p` | Keep commit as-is |
| `reword` | `r` | Keep commit, edit message |
| `edit` | `e` | Pause here, amend commit |
| `squash` | `s` | Merge into previous commit, combine messages |
| `fixup` | `f` | Merge into previous, discard this message |
| `drop` | `d` | Delete commit entirely |
| `exec` | `x` | Run a shell command |

### Common Rebase Patterns

```bash
# Squash WIP commits into one clean commit
# In the rebase editor:
pick abc123 feat: add user search
fixup def456 wip
fixup ghi789 fix typo
fixup jkl012 more fixes
# Result: one clean commit with the first message

# Reorder commits
pick abc123 feat: add search
pick def456 fix: resolve N+1  ← move this up
pick ghi789 feat: add filters

# Split a commit into two
# Mark commit as "edit", then:
git reset HEAD~1                    # Unstage the commit's changes
git add src/search.ts && git commit -m "feat: add search component"
git add src/api.ts && git commit -m "feat: add search API endpoint"
git rebase --continue

# Run tests after each commit during rebase
git rebase -i HEAD~5 --exec "npm test"
```

### Autosquash

```bash
# Create a fixup commit that auto-squashes during rebase
git commit --fixup abc123
# Creates: "fixup! Original commit message"

# Later, rebase with autosquash
git rebase -i --autosquash main
# Fixup commits automatically positioned after their target
```

## Git Bisect

```bash
# Find the commit that introduced a bug
git bisect start
git bisect bad                    # Current commit is broken
git bisect good v1.0.0            # This tag was working

# Git checks out a midpoint — test it
npm test
git bisect good                   # If tests pass
# or
git bisect bad                    # If tests fail

# Git narrows down, repeat until found
# "abc123 is the first bad commit"

git bisect reset                  # Return to original branch

# Automated bisect with a test script
git bisect start HEAD v1.0.0
git bisect run npm test
# Automatically finds the breaking commit
```

## Git Worktrees

```bash
# Work on multiple branches simultaneously without stashing
git worktree add ../hotfix-branch hotfix/urgent-fix
# Creates a new working directory linked to the same repo

# List worktrees
git worktree list
# /path/to/repo           abc1234 [main]
# /path/to/hotfix-branch  def5678 [hotfix/urgent-fix]

# Work in the hotfix directory
cd ../hotfix-branch
# Edit, commit, push — completely independent

# Remove when done
git worktree remove ../hotfix-branch

# Create worktree for a new branch
git worktree add -b feat/new-feature ../new-feature main
# Creates new branch based on main in a separate directory
```

### Worktree Use Cases

| Scenario | Command |
|----------|---------|
| Hotfix while working on feature | `git worktree add ../hotfix hotfix/bug` |
| Compare implementations | `git worktree add ../experiment feat/v2` |
| Run tests on another branch | `git worktree add ../test-branch release/1.0` |
| Long-running task in background | `git worktree add ../bg-task feat/migration` |

## Git Reflog — Recovery

```bash
# See all recent HEAD movements (your safety net)
git reflog
# abc123 HEAD@{0}: commit: add feature
# def456 HEAD@{1}: checkout: moving from feat to main
# ghi789 HEAD@{2}: commit (amend): fix typo
# jkl012 HEAD@{3}: reset: moving to HEAD~1

# Recover deleted branch
git reflog
# Find the commit SHA before deletion
git checkout -b recovered-branch abc123

# Undo a bad reset
git reset --hard HEAD@{2}       # Go back to state 2 steps ago

# Recover after force push
git reflog
git reset --hard HEAD@{1}       # Restore pre-push state

# Recover dropped stash
git fsck --unreachable | grep commit
# Find stash commit SHA in output
git stash apply <sha>

# Reflog entries expire after 90 days (30 for unreachable)
```

## Cherry-Pick

```bash
# Apply a specific commit to current branch
git cherry-pick abc123

# Cherry-pick multiple commits
git cherry-pick abc123 def456 ghi789

# Cherry-pick a range
git cherry-pick abc123..ghi789    # Excludes abc123
git cherry-pick abc123^..ghi789   # Includes abc123

# Cherry-pick without committing (stage changes only)
git cherry-pick --no-commit abc123

# Cherry-pick from another remote
git fetch upstream
git cherry-pick upstream/main~3

# Handle conflicts during cherry-pick
git cherry-pick abc123
# ... resolve conflicts ...
git add .
git cherry-pick --continue
# or
git cherry-pick --abort          # Give up
```

### When to Cherry-Pick

| Use Case | Example |
|----------|---------|
| Backport fix to release branch | `git checkout release/1.0 && git cherry-pick <fix-sha>` |
| Pull one commit from abandoned PR | `git cherry-pick <good-commit>` |
| Apply hotfix to multiple branches | Cherry-pick to each release branch |

## Stash Management

```bash
# Basic stash
git stash                        # Stash tracked changes
git stash -u                     # Include untracked files
git stash -a                     # Include ignored files too

# Named stash
git stash push -m "WIP: search feature"

# Stash specific files
git stash push -m "partial" src/search.ts src/api.ts

# List stashes
git stash list
# stash@{0}: On main: WIP: search feature
# stash@{1}: On feat: debug code

# Apply and keep
git stash apply stash@{0}

# Apply and remove
git stash pop

# View stash contents
git stash show -p stash@{0}      # Full diff
git stash show --stat stash@{0}  # File list only

# Create branch from stash
git stash branch new-feature stash@{0}
# Checks out the original commit, applies stash, drops it

# Drop specific stash
git stash drop stash@{1}

# Clear all stashes
git stash clear
```

## Submodules

```bash
# Add a submodule
git submodule add https://github.com/org/lib.git libs/lib
git commit -m "add lib submodule"

# Clone a repo with submodules
git clone --recurse-submodules https://github.com/org/project.git
# or after clone:
git submodule update --init --recursive

# Update submodule to latest
cd libs/lib
git checkout main && git pull
cd ../..
git add libs/lib
git commit -m "update lib submodule"

# Update all submodules
git submodule update --remote --merge

# Remove a submodule
git submodule deinit libs/lib
git rm libs/lib
rm -rf .git/modules/libs/lib
git commit -m "remove lib submodule"
```

### Submodules vs Alternatives

| Approach | Pros | Cons |
|----------|------|------|
| Submodules | Exact version pinning, true git repos | Complex commands, easy to forget update |
| npm packages | Standard tooling, versioning | Publish cycle overhead |
| Monorepo (workspaces) | Single repo, atomic commits | Large repo, slow operations |
| git subtree | Simpler than submodules, no extra commands | Messy history, hard to extract changes |

## Useful Git Commands

```bash
# Find who changed a line
git blame -L 10,20 src/auth.ts   # Lines 10-20
git blame -w src/auth.ts         # Ignore whitespace changes

# Search commit messages
git log --grep="fix auth"        # Commits mentioning "fix auth"
git log --all --grep="TICKET-123"

# Search code changes (pickaxe)
git log -S "functionName"        # Commits that added/removed "functionName"
git log -G "regex.*pattern"      # Commits matching regex in diff

# Show what changed between branches
git log main..feature            # Commits in feature not in main
git diff main...feature          # Changes since branches diverged

# Find large files in history
git rev-list --objects --all \
  | git cat-file --batch-check='%(objecttype) %(objectname) %(objectsize) %(rest)' \
  | sort -k3 -n -r | head -20

# Clean up local branches merged to main
git branch --merged main | grep -v main | xargs git branch -d

# Verify repo integrity
git fsck --full

# Garbage collect
git gc --aggressive --prune=now
```

## Git Configuration

```bash
# Essential aliases
git config --global alias.co checkout
git config --global alias.br branch
git config --global alias.st status
git config --global alias.lg "log --oneline --graph --decorate --all"
git config --global alias.last "log -1 HEAD --stat"
git config --global alias.unstage "reset HEAD --"
git config --global alias.amend "commit --amend --no-edit"
git config --global alias.wip "commit -am 'WIP'"

# Better diff
git config --global diff.algorithm histogram  # Better diff output
git config --global merge.conflictstyle zdiff3 # Show common ancestor in conflicts

# Auto-stash on rebase
git config --global rebase.autostash true

# Auto-setup remote tracking
git config --global push.autoSetupRemote true

# Sign commits
git config --global commit.gpgsign true
git config --global gpg.format ssh
git config --global user.signingkey ~/.ssh/id_ed25519.pub
```

## Gotchas

1. **`git reset --hard` destroys uncommitted work** — There is no undo for unstaged changes after a hard reset. Always `git stash` first if you have uncommitted work. Committed work can be recovered via `git reflog`.

2. **Interactive rebase rewrites ALL subsequent commits** — Editing commit #3 in a 10-commit rebase changes the SHA of commits 3-10. If those commits are pushed, you'll need `--force-with-lease` and collaborators need to reset.

3. **Cherry-pick creates a NEW commit** — The cherry-picked commit has a different SHA than the original. If you later merge the source branch, Git sees two different commits with the same changes. Use `git merge` instead when possible.

4. **Submodule update doesn't auto-pull** — `git pull` in the parent repo does NOT update submodules. You must run `git submodule update --remote` separately. Add a post-merge hook to automate this.

5. **`git stash` doesn't stash untracked files by default** — New files that haven't been `git add`ed are not stashed with plain `git stash`. Use `git stash -u` to include untracked files.

6. **Reflog is local only** — `git reflog` only exists on your machine. It's not pushed to remotes. If you delete the local repo, the reflog is gone. For team recovery, use branch protection and avoid force pushing.
