#!/bin/bash
# SessionStart hook for the tasks plugin.
#
# The Tasks panel is the user's live status display. This hook makes
# task-tracking mandatory rather than encouraged, and reminds the
# agent that a periodic review schedule is checking on it.
set -euo pipefail

CONTEXT='## Tasks panel: your accountability surface

The user sees your TaskCreate / TodoWrite list as a live panel in
Director. They see exactly what is pending, in_progress, and completed.
Treat this as the source of truth for what work is happening.

## Mandatory task discipline

- **Anything more than a single read/answer call MUST start with
  TaskCreate(s).** Before doing the work, write the steps. The user
  sees the plan as you build it; you keep yourself honest.
- **Mark in_progress when you start.** TaskUpdate the moment you
  begin executing on a task. Exactly one in_progress at a time.
- **Mark completed when the work is on disk / sent / posted.** Not
  when the response that mentions it is finished. The action must
  actually have happened.
- **Delete or downgrade tasks the user no longer wants.** If they
  pivot, your task list pivots too. Stale tasks on the panel are
  worse than no tasks.
- **The user can tick / untick / delete tasks themselves.** Their
  state is the current truth on the next read. If they untick
  something to in_progress, that means you are not done with this.

## Periodic review

A scheduled task (Review tasks every 20 minutes) wakes you to
audit your own list. When that fires:

- Anything in_progress without movement since last review → take
  the next concrete action now, or downgrade to pending with a
  one-line reason in the description so the user knows what is
  blocking it.
- Anything pending that is no longer relevant → delete.
- Anything completed that is actually still pending → re-open.
- Anything the user is waiting on that you finished → report it.

Skipping the review or going through it without acting is worse
than not having the schedule at all. The point is to convert
"I will get to it" into "I am doing it now or it is no longer mine."'

CONTEXT="${CONTEXT//\\/\\\\}"
CONTEXT="${CONTEXT//\"/\\\"}"
CONTEXT="${CONTEXT//$'\n'/\\n}"
CONTEXT="${CONTEXT//$'\t'/\\t}"
CONTEXT="${CONTEXT//$'\r'/}"

printf '{"hookSpecificOutput":{"hookEventName":"SessionStart","additionalContext":"%s"}}' "$CONTEXT"
