#!/bin/bash
# SessionStart hook for the tasks plugin.
#
# The Tasks panel is how the orchestrator drives projects forward:
# every piece of work is a task with an owner and a status. This hook
# reframes the panel from "checklist nice-to-have" to "the project
# board — drive everything through it."
set -euo pipefail

CONTEXT='## Tasks panel: how you drive projects forward

The Tasks panel is your project board. Every piece of work that is
in flight or has not been done yet lives there, with an owner and a
status. The user reads it as the canonical answer to "what is being
worked on, by whom, and how is it going."

This is not a side checklist. It is how you move work.

## The discipline

**Anything more than a single read/answer call starts with TaskCreate.**
Before the work, write the steps. The user sees the plan as you build
it; you keep yourself honest about what you committed to.

**Every task has an owner.** Use the description field to record who:

- `owner: self` — you are doing this directly in this turn
- `owner: <worker-id>` — you delegated it to a roster worker
- `owner: user` — blocked on the user (input, decision, approval)
- `owner: external` — blocked on something outside the fleet (CI run,
  reviewer, third-party API, scheduled time, etc.)

When you delegate, the same task stays on YOUR list with `owner: <worker-id>`.
The worker has its own task list for its own breakdown — do not mirror
it. Your task says "what we are getting done"; the workers task list
says "what I am doing right now."

**Status reflects reality:**

- `pending` — accepted, not started
- `in_progress` — actively being worked on right now (by you or by
  the named owner). Exactly one in_progress per owner at a time.
- `completed` — the work is on disk / sent / posted / merged / shipped.
  Not "I will get to it." Not "the response that mentions it is
  finished." The actual artifact must exist.

**The user can edit too.** Their tick / untick / delete is the truth
on the next read. If they untick something to in_progress, it means
"you are not done with this." If they delete a task, drop it from
your plan.

## Periodic review (the standup)

A scheduled task ("Review tasks every 20 minutes") fires automatically.
It is your standup with yourself. Every fire, walk the list:

- in_progress with movement → keep going, no message needed
- in_progress with no movement → take the next concrete action right
  now, OR demote to pending with a one-line reason in the description
  ("blocked: waiting on PR-merge CI") so the user knows what is in
  the way
- pending no longer needed → delete
- completed but the user is waiting for the result → report it
- delegated to a worker (`owner: <worker-id>`) → check `roster
  describe <worker-id>` and `roster trace <worker-id> --tail 20`.
  If the worker is stalled or off-track, ping them or take it back.

The point of the cadence is to convert "I will get to it" into "I
am doing it now, OR it is blocked on someone specific, OR it is no
longer mine." Going through the review without acting on anything
is worse than not having the review.'

CONTEXT="${CONTEXT//\\/\\\\}"
CONTEXT="${CONTEXT//\"/\\\"}"
CONTEXT="${CONTEXT//$'\n'/\\n}"
CONTEXT="${CONTEXT//$'\t'/\\t}"
CONTEXT="${CONTEXT//$'\r'/}"

printf '{"hookSpecificOutput":{"hookEventName":"SessionStart","additionalContext":"%s"}}' "$CONTEXT"
