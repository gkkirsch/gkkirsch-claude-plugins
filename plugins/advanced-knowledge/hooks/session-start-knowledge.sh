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

INDEX_FILE="knowledge/wiki/index.md"

# Build the always-on guidance. This block is what teaches the agent
# where to write research/docs/etc — without it, agents tend to stash
# files under .claude/ or scatter them in unrelated directories.
COMMON='## Where research & docs go

Anything the user might read or share — research notes, audits,
compiled docs, scraped data, plans, screenshots, scratch markdown —
lives **under your cwd** (the same dir as this prompt). The Director
Library panel shows exactly this directory; files anywhere else are
invisible to the user.

Standard subdirectories (create on demand, no need to ask):

- `research/` — short-form notes, exploratory findings.
- `docs/` — finished prose deliverables.
- `notes/` — scratch and working memory.
- `audits/` — site/code/content audits.
- `data/` — CSVs, JSON dumps, scrape outputs.
- `knowledge/` — long-form compiled research (see below).

For substantial research, prefer the `/knowledge-base` skill: drop
sources in `knowledge/raw/`, run `/knowledge-base compile` to produce
a cross-referenced wiki under `knowledge/wiki/`. The wiki then shows
up in the Library and is queryable from later turns.

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
