---
name: gog
description: Google Workspace CLI for Gmail, Calendar, Drive, Contacts, Sheets, and Docs. Use the `gog` command for searching/sending email, managing calendar events, searching Drive, reading/writing Sheets, exporting Docs, and managing Contacts.
metadata:
  credentials:
    - key: GOG_ACCOUNT
      label: "Default Google Account Email"
      description: "Your Google account email (e.g. you@gmail.com). Avoids repeating --account on every command."
      required: false
  openclaw:
    emoji: "🎮"
    requires:
      bins: ["gog"]
    install:
      - id: brew
        kind: brew
        formula: steipete/tap/gogcli
        bins: ["gog"]
        label: "Install gog CLI (brew)"
---

# gog — Google Workspace CLI

Use the `gog` command for Gmail, Calendar, Drive, Contacts, Sheets, and Docs. Powered by [gogcli](https://github.com/steipete/gogcli).

## Prerequisites

Install gogcli:

```bash
brew install steipete/tap/gogcli
```

## Setup (one-time per account)

1. Create OAuth credentials in Google Cloud Console (Desktop app type)
2. Store the credentials file:
   ```bash
   gog auth credentials /path/to/client_secret.json
   ```
3. Authorize your account with desired services:
   ```bash
   gog auth add you@gmail.com --services gmail,calendar,drive,contacts,sheets,docs
   ```
4. Verify:
   ```bash
   gog auth list
   ```

## Environment

Set `GOG_ACCOUNT` to skip `--account` on every command:

```bash
export GOG_ACCOUNT=you@gmail.com
```

Read from macOS Keychain if configured via superbot2 dashboard:

```bash
GOG_ACCOUNT=$(security find-generic-password -s superbot2-plugin-credentials -a "gog/GOG_ACCOUNT" -w 2>/dev/null || echo "")
```

## Agent Usage

Always use `--json` and `--no-input` flags when running commands programmatically. This ensures machine-parseable output and no interactive prompts:

```bash
gog --json --no-input gmail search 'newer_than:7d' --max 10
```

Always confirm with the user before sending mail or creating/modifying calendar events.

## Gmail

**Search emails:**
```bash
gog --json --no-input gmail search 'newer_than:7d' --max 10
gog --json --no-input gmail search 'from:boss@company.com subject:urgent' --max 5
gog --json --no-input gmail search 'has:attachment filename:pdf' --max 10
```

**Read a thread:**
```bash
gog --json --no-input gmail thread <threadId>
gog --json --no-input gmail message <messageId> --body
```

**Send email (confirm with user first):**
```bash
gog --no-input gmail send --to recipient@example.com --subject "Subject" --body "Message body"
```

For multi-paragraph messages, use `--body-file`:
```bash
gog --no-input gmail send --to recipient@example.com --subject "Subject" --body-file /tmp/message.txt
```

**Reply to a thread:**
```bash
gog --no-input gmail reply <threadId> --body "Reply text"
```

**Create a draft:**
```bash
gog --no-input gmail draft create --to recipient@example.com --subject "Draft" --body "Content"
```

**List labels:**
```bash
gog --json --no-input gmail labels
```

## Calendar

**List events:**
```bash
gog --json --no-input calendar events primary --today
gog --json --no-input calendar events primary --from 2026-02-24T00:00:00Z --to 2026-02-28T00:00:00Z
gog --json --no-input calendar events primary --week
```

**Create event (confirm with user first):**
```bash
gog --no-input calendar create primary --title "Meeting" --start 2026-02-25T10:00:00 --end 2026-02-25T11:00:00
gog --no-input calendar create primary --title "Meeting" --start 2026-02-25T10:00:00 --end 2026-02-25T11:00:00 --attendees "alice@example.com,bob@example.com"
```

**Check availability:**
```bash
gog --json --no-input calendar freebusy --calendars "primary" --from 2026-02-25T00:00:00Z --to 2026-02-26T00:00:00Z
```

## Drive

**Search files:**
```bash
gog --json --no-input drive search "quarterly report" --max 10
gog --json --no-input drive search "name contains 'invoice'" --max 20
```

**List files:**
```bash
gog --json --no-input drive list --max 20
```

**Download a file:**
```bash
gog --no-input drive download <fileId>
gog --no-input drive download <fileId> --format pdf
```

## Sheets

**Read data:**
```bash
gog --json --no-input sheets get <spreadsheetId> "Sheet1!A1:D10"
```

**Write data:**
```bash
gog --no-input sheets update <spreadsheetId> "Sheet1!A1:B2" --values-json '[["Name","Score"],["Alice","95"]]' --input USER_ENTERED
```

**Append rows:**
```bash
gog --no-input sheets append <spreadsheetId> "Sheet1!A:C" --values-json '[["new","row","data"]]' --insert INSERT_ROWS
```

**Clear a range:**
```bash
gog --no-input sheets clear <spreadsheetId> "Sheet1!A2:Z"
```

**Get spreadsheet metadata:**
```bash
gog --json --no-input sheets metadata <spreadsheetId>
```

## Docs

**Export to text:**
```bash
gog --no-input docs export <docId> --format txt --out /tmp/doc.txt
```

**Read doc content:**
```bash
gog --no-input docs cat <docId>
```

## Contacts

**List contacts:**
```bash
gog --json --no-input contacts list --max 20
```

**Search contacts:**
```bash
gog --json --no-input contacts search "John" --max 10
```

**Create contact:**
```bash
gog --no-input contacts create --name "Jane Doe" --email jane@example.com --phone "+1234567890"
```

## Tasks

**List task lists:**
```bash
gog --json --no-input tasks lists
```

**List tasks:**
```bash
gog --json --no-input tasks list <taskListId>
```

**Create task:**
```bash
gog --no-input tasks create <taskListId> --title "Review PR" --due 2026-02-25
```

**Complete task:**
```bash
gog --no-input tasks complete <taskListId> <taskId>
```

## Tips

- Use `--max` to limit results and avoid overwhelming output.
- Use `--json` for all read operations to get structured data.
- Use `--no-input` to prevent interactive prompts in agent contexts.
- Thread search (`gmail search`) returns threads; use `gmail message <id> --body` to read individual messages.
- Sheets values can be passed via `--values-json` (recommended for agents).
- Docs supports export and cat but not in-place editing via gog.
- Always confirm with the user before sending emails or creating calendar events.
