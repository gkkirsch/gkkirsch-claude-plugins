---
description: "Drive a GitHub PR to merged: address comments, fix CI, keep branch updated, set auto-merge, monitor."
argument-hint: "<pr-url-or-number>"
allowed-tools: ["Bash", "Read", "Edit", "Write", "Grep", "Glob"]
---

# /merge-pr

Shepherd PR `$ARGUMENTS` through every blocker until GitHub reports it `MERGED`. Read the **pr-merge** skill in this plugin for the full methodology — the structure below is the bootstrap.

## Bootstrap (run these in order)

### 1. Verify auth and PR target

```bash
gh auth status
gh pr view "$ARGUMENTS" --json number,state,url,title,headRefName,baseRefName,isDraft
```

If auth is broken, stop and tell the user. If the PR is already `MERGED` or `CLOSED`, stop and tell the user. If `isDraft: true`, stop unless the user explicitly said to merge a draft.

### 2. Snapshot full state

```bash
gh pr view "$ARGUMENTS" --json \
  state,mergeable,mergeStateStatus,reviewDecision,isDraft,\
  baseRefName,headRefName,headRefOid,statusCheckRollup,\
  reviews,autoMergeRequest,url,title
```

### 3. Pick the top blocker (priority order)

1. Branch behind base → `gh pr update-branch "$ARGUMENTS"` (or local merge + push if conflicts)
2. CI failing → `gh pr checks "$ARGUMENTS"`, `gh run view <id> --log-failed`, fix, push
3. Open review threads → fetch via `gh api .../pulls/.../comments`, address or reply per the skill
4. Auto-merge not set + everything green → `gh pr merge "$ARGUMENTS" --auto --squash` (match the repo's merge style)
5. Auto-merge set, just waiting → sleep ~60s and re-check

### 4. Loop until `state == "MERGED"` (or you hit a stopping condition)

Use the `/loop` skill to schedule subsequent iterations. The pacing depends on what you're waiting on (CI takes minutes; reviewer responses take hours):

```
/loop check the status of PR $ARGUMENTS via the pr-merge skill and act on the top blocker. exit when the PR is merged or stuck on input you don't have authority for.
```

That delegates the cadence to /loop's dynamic-mode self-pacing — typically 1–2 minutes when CI is in flight, 5–10 minutes when waiting on a human reviewer.

### 5. Report each iteration in one line

Format: `iter N — <state-summary>, <next-action>`

Examples:
- `iter 3 — CI green, 1 unresolved thread (lint nit), pushing reply, waiting`
- `iter 7 — auto-merge set, all checks pass, waiting for GitHub to land it`
- `iter 12 — MERGED ✅`

### 6. Stopping conditions

- ✅ Merged → final report, done
- ❌ Closed without merge → explain
- ⚠️ Need user input (admin merge approval, hard-changes-requested with no reply, repo forbids auto-merge) → pause and ask

## Tips

- **Don't fix two things in one push.** Each push restarts CI; one-issue-per-push isolates cause and saves time.
- **When you reply to a comment with a fix, name the commit SHA.** "Fixed in abc1234 — pulled validation up." Reviewers can't tell otherwise.
- **Read the actual CI log** (`--log-failed`). Don't guess based on the check name.
- **`mergeStateStatus: BEHIND`** means CI on the PR is stale — pull base first, then re-evaluate.

See `skills/pr-merge/SKILL.md` for the full per-blocker playbook.
