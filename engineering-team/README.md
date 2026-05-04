# Engineering Team marketplace

Plugins for everyday engineering work — installed as a set on any orchestrator that ships code.

## Plugins

| Name | What it does |
|---|---|
| **slack** | Send and read Slack messages as yourself (no bot badge). DMs, channels, threads, reactions, file uploads, search. Single static Go binary. |
| **pr-merge** | Drive a GitHub PR to merged: address review comments, fix CI, keep branch updated with base, set auto-merge, monitor in a loop. Includes `/merge-pr` slash command. |
| **code-review** | (scaffold) Review a PR or local diff for correctness, design, and project-convention fit. Structured feedback by severity. |

## Install

Add this marketplace to claude-code:

```bash
claude plugin marketplace add https://github.com/gkkirsch/gkkirsch-claude-plugins/tree/main/engineering-team
```

Then install the plugins you want:

```bash
claude plugin install slack@engineering-team
claude plugin install pr-merge@engineering-team
claude plugin install code-review@engineering-team
```

Or install all of them:

```bash
claude plugin marketplace install engineering-team
```

## Why a separate marketplace?

These three plugins share a context: an engineer writing code, communicating about it, and shepherding it to ship. They don't belong in `director-core` (which is the default-bundled set every orchestrator gets) — only orchs that actually do engineering work need them. Splitting them into their own marketplace lets you opt in per orch.
