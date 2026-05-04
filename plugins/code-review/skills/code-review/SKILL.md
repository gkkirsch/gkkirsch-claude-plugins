---
name: code-review
description: "Use to review a PR or a local diff for correctness, design, and project-convention fit. Produces structured, actionable feedback grouped by severity (Critical / Important / Suggestions). Activates when the user says 'review this PR', 'review my changes', shares a PR URL with intent to review, or finishes a chunk of work and asks for a sanity-check. Keywords: review, code review, pr review, feedback, critique, sanity check."
---

# Code Review (scaffold)

> **Status: scaffold.** This skill is a sketch — fill in the per-language, per-project conventions you actually want enforced before relying on it.

## When to use

The user (or an agent finishing a chunk of work) wants a second pair of eyes on a diff before merging. Inputs:

- A PR URL or number (review against GitHub state)
- "the diff in this branch" (review `git diff main...HEAD`)
- "what I just changed" (review uncommitted + recently-committed work)

## Output shape

Always group findings by severity. An author should be able to scan and act:

```
## Critical (must fix before merge)
- file.ts:42 — <issue> — <suggested fix>

## Important (should address)
- ...

## Suggestions (nice-to-have)
- ...

## Strengths (call out what's good)
- ...

## Recommended action
1. Fix critical
2. Address important
3. Consider suggestions
4. Re-run review or merge
```

If there are zero findings, say so plainly — don't pad with theatre.

## Phases (TODO: flesh these out per project)

### 1. Get the diff

```bash
# PR
gh pr diff "$PR"

# Local branch vs main
git diff main...HEAD

# Uncommitted
git diff
```

### 2. Run focused passes

The general rule: **one pass per concern**. Mixing all concerns into one review means each one gets shallow attention. Suggested passes:

- **Correctness** — does it do what it claims? Are there off-by-ones, race conditions, error swallowing, untested edge cases?
- **Design** — leaky abstractions, premature abstractions, parameter sprawl, copy-paste-with-variation
- **Project fit** — does it match the conventions of the surrounding code? (TODO: list project-specific conventions here once known)
- **Testing** — what does the test suite actually cover? Missing tests for the new branch?
- **Comments / docs** — are comments saying *why* (good) or *what* (delete)?
- **Security** — input validation at boundaries, injection, secrets leaks, dependency CVEs
- **Performance** — obvious N+1s, hot-path bloat, unbounded memory

(TODO: add a project-specific pass — e.g. "i18n strings extracted", "feature flags wired correctly", "telemetry events fire", whatever you actually care about)

### 3. De-duplicate and prioritize

- Merge findings that say the same thing in different files
- Anything that breaks correctness → Critical
- Anything that hurts maintainability or violates convention → Important
- Style / aesthetic preferences → Suggestions
- Don't bikeshed; if a thing is fine-but-different, leave it

### 4. Deliver

- For PR reviews: post as `gh pr review --comment` (general) + inline review comments on specific lines
- For local diffs: print the report directly

## Useful tooling

- `gh pr diff` / `gh pr view --json files` — diff + metadata
- `gh pr checks` — what CI is saying
- `git log -p path/to/file` — historical context for a touched line
- `git blame -L x,y file` — who last touched these lines + what commit message
- `Grep` — find existing patterns the new code may have ignored

## Anti-patterns

- ❌ "LGTM" without doing the passes
- ❌ "Looks fine" with no structured findings — devalues the review
- ❌ Posting a wall of nits as "Critical"
- ❌ Reviewing the diff without the surrounding file (read for context)

## TODO before relying on this

This scaffold is a starting point. To make it actually good for our codebases, fill in:

- [ ] Project-specific conventions (file layout, naming, error handling style, test patterns)
- [ ] Per-language gotchas we keep hitting (Go, TS, Python — whichever the team uses)
- [ ] Severity calibration (e.g. "in this codebase, X is always critical because Y")
- [ ] Templates for common review verdicts (so the output reads consistent)
- [ ] Hook for posting inline GitHub comments at specific lines, not just the summary

Once those are filled in, this becomes a real review skill instead of a generic checklist.
