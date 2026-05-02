#!/bin/bash
# SessionStart hook for advanced-knowledge.
#
# Injects an "additionalContext" block into the agent's system prompt
# every session. Two responsibilities:
#
#   1. Tell the agent that research/docs/notes/audits go under its
#      cwd ($DIRECTOR_SPACE for orch agents) — that's the directory
#      the user sees in Director's Library panel. Files written
#      anywhere else are invisible to the user.
#
#   2. If a knowledge wiki exists for this agent, surface its index so
#      the agent can reference past research instead of re-deriving it.
#
# The hook runs with PWD set to the agent's cwd, so all paths are
# relative.
set -euo pipefail

INDEX_FILE="library/wiki/index.md"

# Build the always-on guidance. This block teaches the agent where to
# write artifacts — without it, agents tend to stash files under
# .claude/ or scatter them across unrelated directories.
COMMON='## Where everything you make goes: library/

Every artifact you produce that the user might read or share —
research, audits, compiled docs, scraped data, plans, screenshots,
scratch markdown — goes directly under `library/` in your cwd. That
single directory is what the Director Library panel shows the user.

**Keep `library/` flat.** Drop files at the root —
`library/<some-thing>.md`, `library/<scraped-data>.csv`,
`library/<screenshot>.png`. Do not invent topical subdirectories
(no `library/audits/`, no `library/research/`, no
`library/notes/`). A flat list is what the user sees.

The one exception is the wiki, which has structure:

- `library/wiki/` — compiled knowledge base built by the
  `/knowledge-base` skill. Concepts, summaries, connections, the
  index, and filed Q&A queries (`library/wiki/queries/`). Use this
  for substantial research you want cross-referenced and reusable.

**Never write to `.claude/`** — claude-code blocks writes there.'

if [ -f "$INDEX_FILE" ]; then
  WIKI=$(head -50 "$INDEX_FILE")
  WIKI_BLOCK=$'\n\n## Your existing knowledge base\n\nWiki index (top of '"$INDEX_FILE"$'):\n\n'"$WIKI"
else
  WIKI_BLOCK=$'\n\n## Your existing knowledge base\n\n(empty — no wiki compiled yet)'
fi

CONTEXT="${COMMON}${WIKI_BLOCK}"

# Escape for JSON embedding. Order matters: backslash first.
CONTEXT="${CONTEXT//\\/\\\\}"
CONTEXT="${CONTEXT//\"/\\\"}"
CONTEXT="${CONTEXT//$'\n'/\\n}"
CONTEXT="${CONTEXT//$'\t'/\\t}"
CONTEXT="${CONTEXT//$'\r'/}"

printf '{"hookSpecificOutput":{"hookEventName":"SessionStart","additionalContext":"%s"}}' "$CONTEXT"
