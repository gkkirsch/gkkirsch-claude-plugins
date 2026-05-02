#!/bin/bash
# SessionStart hook for the tasks plugin.
#
# Tells the agent that its TaskCreate/TodoWrite list is now visible to
# the user in the Tasks panel. Encourages keeping it current — the
# user reads it as a live status display.
set -euo pipefail

CONTEXT='## Tasks panel is visible to the user

Your TaskCreate / TodoWrite list shows up in the Tasks panel of
Director. The user sees: pending → in-progress → completed, with
checkboxes they can toggle and a delete button per task.

Use it. Keep it current:

- Break multi-step work into tasks BEFORE starting (TaskCreate).
- Mark each task in_progress when you start it (TaskUpdate).
- Mark completed the moment it'"'"'s done — donedoes not mean
  "the response that mentions it is finished," it means the actual
  work is on disk.
- One task in_progress at a time, no exceptions.

The user might tick or untick tasks too — treat their state as the
current truth on the next read.'

CONTEXT="${CONTEXT//\\/\\\\}"
CONTEXT="${CONTEXT//\"/\\\"}"
CONTEXT="${CONTEXT//$'\n'/\\n}"
CONTEXT="${CONTEXT//$'\t'/\\t}"
CONTEXT="${CONTEXT//$'\r'/}"

printf '{"hookSpecificOutput":{"hookEventName":"SessionStart","additionalContext":"%s"}}' "$CONTEXT"
