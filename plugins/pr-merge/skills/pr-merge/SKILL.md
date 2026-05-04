---
name: pr-merge
description: "Use to shepherd a GitHub PR all the way to merged. Addresses review comments (either by editing code or replying for clarification), fixes failing CI by reading the actual logs, keeps the branch updated with base, enables auto-merge, and monitors in a loop until the PR is closed/merged. Activates when the user asks to merge / land / ship / unblock a PR, gives a PR URL or number, or says 'watch this PR'. Keywords: github, gh, pr, pull request, merge, auto-merge, ci, checks, review, comments, rebase, branch, ship, land."
---

# PR Merge: drive a PR to merged

## When this activates

The user has a PR in flight and wants it merged. Maybe it has open review comments. Maybe CI is red. Maybe the branch is behind. Your job is to walk it through every blocker until GitHub reports `MERGED`.

This is not a one-shot review skill — it's a loop. You'll re-check the PR's state every few minutes (or on event), act on whatever's now blocking, and keep going.

## Required tools

- `gh` CLI (already on PATH for any orch with our tooling)
- The repo cloned and writable (you'll be pushing commits)
- Auth: `gh auth status` should be green

If `gh auth status` is red, surface that to the user and stop — you can't drive the PR without auth.

## The loop

```
                ┌─────────────────────┐
        ┌──────►│  fetch PR state     │
        │       └──────────┬──────────┘
        │                  │
        │       ┌──────────▼──────────┐
        │       │   merged or closed? │──── yes ──► report + stop
        │       └──────────┬──────────┘
        │                no│
        │       ┌──────────▼──────────┐
        │       │ identify top blocker│
        │       └──────────┬──────────┘
        │                  │
        │  ┌───────────────┼───────────────┐
        │  │               │               │
        │  ▼               ▼               ▼
        │ behind base    CI failing    open review
        │  │               │            comments
        │  │               │               │
        │  ▼               ▼               ▼
        │ pull base      read logs +    address each
        │ into branch    fix + push     (edit or reply)
        │  │               │               │
        │  └───────────────┼───────────────┘
        │                  │
        │       ┌──────────▼──────────┐
        │       │ auto-merge enabled? │ no → enable it
        │       └──────────┬──────────┘
        │                  │
        │                wait
        │                  │
        └──────────────────┘
```

Always work the **top** blocker — don't try to fix three things at once. If the branch is behind, the CI you'd read is stale; pull base first, then re-run the loop.

## Step 1 — fetch state

One `gh` call gives you almost everything:

```bash
gh pr view "$PR" --json \
  number,state,mergeable,mergeStateStatus,reviewDecision,\
  isDraft,baseRefName,headRefName,headRefOid,\
  statusCheckRollup,reviews,comments,autoMergeRequest,url,title
```

(`$PR` is a URL or number; `gh` accepts both.)

Key fields:

| Field | What to look for |
|---|---|
| `state` | `OPEN` / `CLOSED` / `MERGED` |
| `mergeStateStatus` | `BEHIND` / `BLOCKED` / `CLEAN` / `DIRTY` / `HAS_HOOKS` / `UNSTABLE` |
| `reviewDecision` | `APPROVED` / `CHANGES_REQUESTED` / `REVIEW_REQUIRED` |
| `statusCheckRollup` | array — any `conclusion: FAILURE`? |
| `autoMergeRequest` | null if not enabled |
| `isDraft` | true → don't try to merge; surface to user |

## Step 2 — pick the blocker

Priority order (do the highest-priority blocker first; ignore the rest until next iteration):

1. **PR is draft** → tell user, stop. Drafts are intentional.
2. **PR is closed** → stop.
3. **Branch is behind base** (`mergeStateStatus == "BEHIND"`)
4. **CI is failing** (any check has `conclusion == "FAILURE"`)
5. **Open review comments** (unresolved threads)
6. **Auto-merge not yet enabled, but everything is green**
7. **Auto-merge enabled, waiting for the actual merge** → just sleep + re-check

## Step 3 — playbooks per blocker

### Behind base

```bash
gh pr update-branch "$PR"
```

If GitHub refuses (e.g. conflict), pull locally and resolve:

```bash
git fetch origin
git checkout <head-branch>
git merge origin/<base-branch>   # or rebase, per repo convention
# resolve conflicts, commit
git push
```

### CI failing

Pull the failing run and read its logs — *read them*, don't guess.

```bash
gh pr checks "$PR"

# Identify the failing run id from the output (or the JSON form):
gh pr checks "$PR" --json name,conclusion,detailsUrl,workflow

# For each FAILED run:
RUN_ID=<id>
gh run view "$RUN_ID" --log-failed | head -200
```

Read enough log to know **what** failed and **why**. Then:

- **Code bug** → fix the code, push.
- **Flaky test** → re-run the workflow (`gh run rerun "$RUN_ID" --failed`), and tell the user "rerunning a flake."
- **Infra failure** (auth issue, dependency download timeout) → re-run; if it persists, surface.
- **New test the change should add** → add it.

After pushing a fix, **wait** for the new check run to start before re-fetching state — racing GitHub gives you stale data.

### Open review comments

Get the comment threads:

```bash
# Inline review comments (PR comments tied to specific lines):
gh api "repos/{owner}/{repo}/pulls/{pr}/comments" --jq \
  '[.[] | {id, path, line, body, user: .user.login, in_reply_to_id, created_at}]'

# General PR conversation comments:
gh api "repos/{owner}/{repo}/issues/{pr}/comments" --jq \
  '[.[] | {id, body, user: .user.login, created_at}]'

# Review summary verdicts:
gh api "repos/{owner}/{repo}/pulls/{pr}/reviews" --jq \
  '[.[] | {id, state, body, user: .user.login, submitted_at}]'
```

For each comment, decide:

**Address it** (edit code) when:
- The reviewer points out a real issue
- The reviewer asks for a clearly-defined change
- It's a project-convention fit

**Reply to it** (don't change code) when:
- The comment was a question — answer it
- You disagree — explain why politely
- The comment is about something already changed in a later commit — point at the commit
- The reviewer is asking for a follow-up that's out of scope — propose a follow-up issue/PR

Reply to a specific inline thread:

```bash
gh api "repos/{owner}/{repo}/pulls/{pr}/comments" \
  -X POST \
  -f body="$REPLY" \
  -F in_reply_to=$ORIGINAL_COMMENT_ID
```

General PR comment:

```bash
gh pr comment "$PR" --body "$BODY"
```

After every code change in response to a review:
1. Push.
2. Reply to the comment you addressed, naming the commit SHA: e.g.  
   `Fixed in abc1234 — moved the validation up to the handler.`
3. Mark the conversation resolved if you have permission:  
   `gh api graphql -f query=...resolveReviewThread...`  
   (else leave it for the reviewer)

### Auto-merge

Once the branch is up-to-date and CI is green and reviews are approved, enable auto-merge so GitHub completes the merge for you and you don't have to babysit the click:

```bash
# Check if it's already enabled (autoMergeRequest is non-null in the JSON view)
# If not:
gh pr merge "$PR" --auto --squash       # or --merge / --rebase per repo convention
```

Pick the merge style the repo uses — look at recent merged PRs:
```bash
gh pr list --state merged --limit 5 --json mergeCommit,title,number
```
or read `.github/pull_request_template.md` and CONTRIBUTING for hints.

### Waiting for the actual merge

Once auto-merge is set + everything is green, GitHub will merge in seconds-to-minutes. You don't need to babysit aggressively — sleep 30-60s and re-fetch. If `state == "MERGED"`, report success.

## Step 4 — pacing the loop

Two reasonable cadences depending on what you're waiting on:

| Waiting on | Cadence |
|---|---|
| CI to start a run after a push | poll every 15-30s for ~3 min |
| CI to finish | every 60-120s |
| A reviewer to come back to address a reply | every 5-10 min for ~1 hour, then back off |
| Auto-merge to land | every 30-60s |

**Use `/loop`** for the actual cadence — it's already wired and handles both fixed-interval and self-paced. Drive the iteration with:

```
/loop check status of PR <url> using the pr-merge skill
```

Or do it inline with `ScheduleWakeup` if you're already in a session and just want a heartbeat.

## Step 5 — when to stop

Stop the loop when any of these happens — and report which:

- ✅ `state == "MERGED"` — celebrate.
- ❌ `state == "CLOSED"` and no merge — explain to user.
- ⚠️ A blocker requires user input you don't have authority for — pause + ask:
  - Reviewer with hard `CHANGES_REQUESTED` who hasn't replied in 24h
  - A test failure that requires non-obvious domain knowledge
  - Auto-merge requires admin approval
  - Repo settings forbid auto-merge → tell user, ask if they want to merge manually

Always say **why** you stopped, with the most recent state snapshot.

## Reporting

Each iteration, surface a one-line status update:

> `iter 7 — CI green, 1 unresolved thread (lint), auto-merge set, waiting`

Final report:

> ✅ Merged at <sha>. Branch <head> → <base>. 4 iterations, 2 CI fixes (typo + missing import), 3 review comments addressed, 1 reply ("test rationale").

## Common mistakes to avoid

- ❌ **Fixing two things in one push** when CI is the bottleneck. Each push triggers its own CI run; sequential one-thing-per-push lets you isolate cause.
- ❌ **Replying "fixed!" without a commit SHA**. Reviewers can't tell what you fixed without the pointer.
- ❌ **Pushing to a branch the user is also pushing to** without coordinating — pull first, every time.
- ❌ **Ignoring `mergeStateStatus`**. `BEHIND` means the CI on the PR is stale; fix that before reading checks.
- ❌ **Force-pushing over reviewer comments** that point at a specific commit SHA — your push invalidates the SHA they referenced. Prefer fixup commits + squash-on-merge.
- ❌ **Treating auto-merge as "done"**. It's "set"; you still verify it actually merged before reporting success.

## Test the integration after a code change (uses dogfood)

Before you push your fix, exercise it locally — at least at the test level:

```bash
# Whatever the project's test command is — find it in package.json / Makefile / etc.
npm test -- --filter "<the failing test>"
```

If you can't reproduce locally what's failing in CI, try to (different node version, different env). You'll iterate faster on the fix when the local feedback loop matches CI.

This is the dogfood principle applied to PR work: don't push and pray. Run it locally, see it pass, *then* push.
