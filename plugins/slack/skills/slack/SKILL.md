---
name: slack
description: "Use when sending Slack messages, uploading files, reading channels, searching conversations, or replying to threads. Sends DMs by name ('Alice'), hashtag channels ('#general'), group DMs ('Alice, Bob'). Uploads files. Reads channel history, searches messages, replies to threads, adds reactions. Keywords: slack, message, channel, DM, direct message, search, thread, react, read, send, upload, file, attachment."
---

# Slack Skill

## Overview

Full Slack integration as **you** (no APP badge). One static Go binary, no Python or runtime dependencies. OAuth user token in macOS Keychain.

**Key features:**
- Unified target syntax: `"Alice"`, `"#channel"`, `"Alice, Bob"`, `"alice@x.com"`
- Auto-resolves user names from your recent Slack activity
- Client-side rate limiting
- :robot_face: reaction added automatically to outgoing messages so recipients can tell it was AI-assisted (use `--no-react` to skip)

**Limitations:**
- Public channels, DMs, and group DMs only — private channels are not supported
- macOS or Linux (binaries are pre-built for darwin-arm64, darwin-amd64, linux-amd64, linux-arm64)

## Quick Reference

| Action | Command |
|--------|---------|
| **Send to channel** | `${CLAUDE_PLUGIN_ROOT}/../slack send "#channel" "message"` |
| **Send DM by name** | `${CLAUDE_PLUGIN_ROOT}/../slack send "Alice" "message"` |
| **Send group DM** | `${CLAUDE_PLUGIN_ROOT}/../slack send "Alice, Bob" "message"` |
| **Read channel** | `${CLAUDE_PLUGIN_ROOT}/../slack read "#channel" -n 25` |
| **Read thread** | `${CLAUDE_PLUGIN_ROOT}/../slack thread read CHANNEL_ID/TS` |
| **Reply to thread** | `${CLAUDE_PLUGIN_ROOT}/../slack thread reply CHANNEL_ID/TS "message"` |
| **Search messages** | `${CLAUDE_PLUGIN_ROOT}/../slack search "deployment failed"` |
| **Recent activity** | `${CLAUDE_PLUGIN_ROOT}/../slack activity --days 7` |
| **Upload file** | `${CLAUDE_PLUGIN_ROOT}/../slack upload "#channel" /path/to/file` |
| **Upload with msg** | `${CLAUDE_PLUGIN_ROOT}/../slack upload -m "look" "Alice" /path/to/file` |
| **Add reaction** | `${CLAUDE_PLUGIN_ROOT}/../slack react CHANNEL_ID/TS thumbsup` |
| **Debug resolution** | `${CLAUDE_PLUGIN_ROOT}/../slack resolve --debug "Alice"` |
| **Cache status** | `${CLAUDE_PLUGIN_ROOT}/../slack cache` |
| **Clear caches** | `${CLAUDE_PLUGIN_ROOT}/../slack cache --clear` |
| **Auth status** | `${CLAUDE_PLUGIN_ROOT}/../slack auth --status` |
| **Re-authenticate** | `${CLAUDE_PLUGIN_ROOT}/../slack auth --reauth` |

> Note: `${CLAUDE_PLUGIN_ROOT}` resolves to `<plugin>/skills/slack/`, so `../` walks back to the plugin root where the `slack` shim lives.

## Target Syntax

| Target | Type | Example |
|--------|------|---------|
| `#channel-name` | Channel by name | `"#general"` |
| `C01ABC123` | Channel by ID | direct |
| `Alice` | DM by name | auto-resolves |
| `Alice Smith` | DM by full name | more precise |
| `@alice.smith` | DM by handle | exact handle |
| `alice@example.com` | DM by email | email lookup |
| `Alice, Bob` | Group DM | comma-separated |

## Setup

```bash
# Required for OAuth
export SLACK_CLIENT_ID="your-client-id"
export SLACK_CLIENT_SECRET="your-client-secret"

# First run kicks off browser-based OAuth and stores the token in Keychain
slack auth
```

## Output

- **Sends, uploads, reactions, auth, cache** → JSON to stdout
- **Reads, search, activity** → human-readable text (use `--json` for raw)

Read output uses message refs in the form `CHANNEL_ID/TIMESTAMP` (e.g. `C01ABC123/1736934225.000100`). Pass these to `thread read`, `thread reply`, or `react`.

## Permission hook

Write operations (send / react / reply / upload) go through a `PreToolUse` hook so Claude can review before posting. Set the mode in `.claude/settings.local.json`:

```json
{ "env": { "SLACK_PERMISSION_MODE": "smart" } }
```

| Mode    | Behavior                                              |
|---------|-------------------------------------------------------|
| `ask`   | Confirm every write (default)                         |
| `allow` | Auto-approve everything                               |
| `smart` | Auto-approve `#channels` and recent contacts; ask otherwise |

Read-only operations never prompt.

## Storage

- **OAuth token** — macOS Keychain, service `director-slack-user`
- **Caches** — `~/.cache/director-slack/` (channels, recent contacts, user names, rate state)

## Common mistakes

### Using a first name with multiple matches
"John" in a workspace with multiple Johns will resolve to the most-recent contact. Use full name, handle, or email for precision: `"John Smith"`, `"@john.smith"`, `"john@example.com"`.

### Channel without `#`
`general` (no `#`) is treated as a DM target. Always include the `#` for channels.

### Sending to channels you haven't joined
The skill can only see channels you're a member of. Join first in the Slack UI.

### Forgetting to re-auth after scope changes
If you see `missing_scope` errors after upgrading, run `slack auth --reauth`.
