# Slack

Send and read Slack messages **as you** (no bot badge) using a personal OAuth user token.

Single static Go binary, zero runtime deps. No Python, no `uv`, no shebangs.

## What it does

- **Send** messages to channels, DMs, group DMs
- **Read** channel history and threads
- **Reply** to threads (with optional broadcast)
- **Search** messages across the workspace
- **React** with emoji
- **Upload** files
- **Activity** — quick view of who you've talked to recently

All actions appear from your account, not a bot.

## Layout

```
plugins/slack/
├── slack                       ← shim that picks the right pre-built binary
├── bin/
│   ├── darwin-arm64/director-slack
│   ├── darwin-amd64/director-slack
│   ├── linux-amd64/director-slack
│   └── linux-arm64/director-slack
├── skills/slack/SKILL.md
├── *.go                        ← source
├── go.mod
└── build.sh                    ← rebuild for all platforms
```

## Install

```bash
export SLACK_CLIENT_ID="your-client-id"
export SLACK_CLIENT_SECRET="your-client-secret"

./slack auth
```

The first run opens a browser, you authorize, the token lands in macOS Keychain (`director-slack-user`).

## Examples

```bash
./slack send "#general" "hello"
./slack send "Alice" "got a sec?"
./slack send "Alice, Bob" "sync at 3?"

./slack read "#general" -n 25
./slack thread read C01ABC123/1736934225.000100
./slack thread reply C01ABC123/1736934225.000100 "thanks!"

./slack search "deployment failed"
./slack activity --days 7

./slack upload "#general" /path/to/file.png
./slack react C01ABC123/1736934225.000100 thumbsup
```

## Develop

```bash
go build -o /tmp/slack-test .   # build a local binary for testing
./build.sh                       # rebuild all 4 platforms into bin/
```

`go build .` (no `-o`) drops a `director-slack` binary in this folder — that name is gitignored to avoid shadowing the shim.

## Storage

- **OAuth token** — macOS Keychain, service `director-slack-user`
- **Caches** — `~/.cache/director-slack/` (channels, recent contacts, rate state)

## Notes

- Public channels, DMs, group DMs only (private channels not supported)
- Sent messages auto-react with `:robot_face:` so recipients can tell it was AI-assisted (use `--no-react` to skip)
- After updating scopes: `./slack auth --reauth`
- Rate-limit status: `./slack cache --rate-status`
