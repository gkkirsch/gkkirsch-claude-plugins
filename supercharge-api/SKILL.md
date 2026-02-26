---
name: supercharge-api
description: "Use when interacting with the superchargeclaudecode.com platform API. Covers auth, plugin management, marketplace operations, and deployment."
metadata:
  credentials:
    - key: SUPERCHARGE_EMAIL
      label: "Supercharge Claude Code Email"
      description: "Your account email at superchargeclaudecode.com"
      required: true
    - key: SUPERCHARGE_PASSWORD
      label: "Supercharge Claude Code Password"
      description: "Your account password at superchargeclaudecode.com"
      required: true
---

# Supercharge Claude Code API

Reference for the superchargeclaudecode.com platform API.

## Quick Reference

- **Plugins repo**: `~/dev/gkkirsch-claude-plugins` → `https://github.com/gkkirsch/gkkirsch-claude-plugins`
- **Login response token path**: `data.token` (NOT top-level `token`)
- **Public API** (`/api/plugins/:name`): Does NOT return plugin IDs
- **Authenticated API** (`/plugins/my-plugins`): Returns plugin IDs (needed for delete/sync)
- **Trusted publisher** (superbot2): Plugins auto-approved on import/submit

## Common Workflows

### Update an Existing Plugin (delete + re-import)

This is the most reliable way to update a plugin. Sync only works if the plugin was originally imported from GitHub (has a stored `sourceUrl`).

```bash
# Step 1: Login
TOKEN=$(curl -s -X POST https://superchargeclaudecode.com/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"superbot2@superchargeclaudecode.com","password":"superbot2"}' \
  | python3 -c "import json,sys; print(json.load(sys.stdin)['data']['token'])")

# Step 2: Push latest code to GitHub
cd ~/dev/gkkirsch-claude-plugins && git push origin main

# Step 3: Find plugin ID (public API does NOT return IDs)
PLUGIN_ID=$(curl -s https://superchargeclaudecode.com/plugins/my-plugins \
  -H "Authorization: Bearer $TOKEN" \
  | python3 -c "
import json,sys
for p in json.load(sys.stdin)['data']:
    if p['slug'] == 'PLUGIN_SLUG':
        print(p['id']); break
")

# Step 4: Try sync first (only works if plugin has sourceUrl)
curl -s -X POST "https://superchargeclaudecode.com/plugins/${PLUGIN_ID}/sync" \
  -H "Authorization: Bearer $TOKEN"
# If sync returns success → done. If not, continue to step 5.

# Step 5: Delete and re-import
curl -s -X DELETE "https://superchargeclaudecode.com/plugins/${PLUGIN_ID}" \
  -H "Authorization: Bearer $TOKEN"

curl -s -X POST https://superchargeclaudecode.com/plugins/import-url \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"url":"https://github.com/gkkirsch/gkkirsch-claude-plugins/tree/main/PLUGIN_SLUG"}'

# Step 6: Verify
curl -s https://superchargeclaudecode.com/api/plugins/PLUGIN_SLUG | python3 -m json.tool
```

Replace `PLUGIN_SLUG` with the plugin name (e.g., `superbot-browser`, `gog`, `1password`).

### Sync a GitHub-Imported Plugin (fastest)

Only works if the plugin was imported via `import-url` (has a stored `sourceUrl`). If the plugin was uploaded via file upload, sync will fail — use delete + re-import instead.

```bash
TOKEN=$(curl -s -X POST https://superchargeclaudecode.com/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"superbot2@superchargeclaudecode.com","password":"superbot2"}' \
  | python3 -c "import json,sys; print(json.load(sys.stdin)['data']['token'])")

# Get plugin ID
PLUGIN_ID=$(curl -s https://superchargeclaudecode.com/plugins/my-plugins \
  -H "Authorization: Bearer $TOKEN" \
  | python3 -c "
import json,sys
for p in json.load(sys.stdin)['data']:
    if p['slug'] == 'PLUGIN_SLUG':
        print(p['id']); break
")

# Sync
curl -s -X POST "https://superchargeclaudecode.com/plugins/${PLUGIN_ID}/sync" \
  -H "Authorization: Bearer $TOKEN" | python3 -m json.tool
```

## Authentication

### Login

```bash
curl -X POST https://superchargeclaudecode.com/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "superbot2@superchargeclaudecode.com", "password": "<password>"}'
```

Response: `{ "success": true, "data": { "user": {...}, "token": "JWT_TOKEN" } }`

**Important**: Token is at `data.token`, not top-level. Extract with:
```python
python3 -c "import json,sys; print(json.load(sys.stdin)['data']['token'])"
```

Use for all authenticated requests:

```
Authorization: Bearer <token>
```

### Signup

```bash
curl -X POST https://superchargeclaudecode.com/auth/signup \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "...", "name": "..."}'
```

## Account

- **superbot2**: superbot2@superchargeclaudecode.com (trusted publisher, `isTrustedPublisher: true`)
- **Plugins repo**: `~/dev/gkkirsch-claude-plugins` → GitHub: `https://github.com/gkkirsch/gkkirsch-claude-plugins`

## Plugins

### List All Plugins (public)

```bash
curl https://superchargeclaudecode.com/api/plugins
```

### Get Single Plugin (public)

```bash
curl https://superchargeclaudecode.com/api/plugins/:name
```

### Full Marketplace Listing (public)

Returns all approved plugins (85+) in standard marketplace.json format.

```bash
curl https://superchargeclaudecode.com/api/marketplace.json
```

### Import Plugin from GitHub (auth required)

```bash
curl -X POST https://superchargeclaudecode.com/plugins/import-url \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"url": "https://github.com/user/repo"}'
```

Also supports subdirectory URLs (e.g., `https://github.com/owner/repo/tree/main/skills/my-skill`).

For trusted publishers, imported plugins are auto-approved.

### Plugin Upload — Folder/File Upload (auth required)

Upload a local plugin folder via the multi-step API. There is no single zip endpoint — upload files individually with their relative paths.

**Step 1: Create a plugin draft**

```bash
curl -X POST https://superchargeclaudecode.com/plugins \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-plugin",
    "description": "What this plugin does",
    "version": "1.0.0",
    "tags": ["tag1", "tag2"]
  }'
```

Response includes `data.id` (the pluginId for subsequent calls) and `data.slug`.

Required fields: `name` (2-50 chars, alphanumeric/hyphens/underscores), `description`, `version` (semver).
Optional fields: `shortDesc`, `authorName`, `tags` (max 10), `repositoryUrl`.

**Step 2: Upload files one at a time**

```bash
curl -X POST https://superchargeclaudecode.com/plugins/<pluginId>/files \
  -H "Authorization: Bearer <token>" \
  -F "file=@./skills/my-skill/SKILL.md" \
  -F "relativePath=skills/my-skill/SKILL.md"
```

- `file`: The file content (multipart/form-data, max 5MB per file)
- `relativePath`: Path within the plugin directory structure (preserves folder hierarchy)

The server auto-detects file type (SKILL, COMMAND, AGENT, HOOK, MCP_CONFIG, OTHER) from the path.

System files (.DS_Store, Thumbs.db, desktop.ini, ._* files) are silently skipped.

**Step 3: Submit for review**

```bash
curl -X POST https://superchargeclaudecode.com/plugins/<pluginId>/submit \
  -H "Authorization: Bearer <token>"
```

For trusted publishers (e.g., superbot2), plugins are auto-approved on submit.

**Full example — upload a plugin folder via bash:**

```bash
# Login
TOKEN=$(curl -s -X POST https://superchargeclaudecode.com/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"$SUPERCHARGE_EMAIL","password":"$SUPERCHARGE_PASSWORD"}' \
  | python3 -c "import json,sys; print(json.load(sys.stdin)['data']['token'])")

# Create plugin
PLUGIN_ID=$(curl -s -X POST https://superchargeclaudecode.com/plugins \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"my-plugin","description":"My plugin","version":"1.0.0"}' \
  | python3 -c "import json,sys; print(json.load(sys.stdin)['data']['id'])")

# Upload each file (preserving relative paths)
for file in skills/my-skill/SKILL.md .claude-plugin/plugin.json README.md; do
  curl -s -X POST "https://superchargeclaudecode.com/plugins/${PLUGIN_ID}/files" \
    -H "Authorization: Bearer $TOKEN" \
    -F "file=@./${file}" \
    -F "relativePath=${file}"
done

# Submit for review
curl -s -X POST "https://superchargeclaudecode.com/plugins/${PLUGIN_ID}/submit" \
  -H "Authorization: Bearer $TOKEN"
```

### Delete File from Plugin (auth required)

```bash
curl -X DELETE https://superchargeclaudecode.com/plugins/<pluginId>/files/<fileId> \
  -H "Authorization: Bearer <token>"
```

### Delete Plugin (auth required)

```bash
curl -X DELETE https://superchargeclaudecode.com/plugins/<pluginId> \
  -H "Authorization: Bearer <token>"
```

### Sync Plugin from GitHub Source (auth required)

Re-fetches all files from the plugin's original GitHub source URL. Deletes existing files and replaces them with fresh ones. Only works for plugins that were imported from a GitHub URL. Owner or trusted publishers only.

```bash
curl -X POST https://superchargeclaudecode.com/plugins/<pluginId>/sync \
  -H "Authorization: Bearer <token>"
```

Returns `{ "success": true, "data": { "ok": true, "filesUpdated": N } }`.

Returns 400 if the plugin has no stored source URL.

### Get User's Plugins (auth required)

**Important**: This is the only way to get plugin IDs. The public API (`/api/plugins/:name`) does not return IDs. You need the ID for delete, sync, and file operations.

```bash
curl https://superchargeclaudecode.com/plugins/my-plugins \
  -H "Authorization: Bearer <token>"
```

Response includes `id`, `slug`, `sourceUrl`, `status`, and all file details for each plugin.

## Custom Marketplaces

### List Your Marketplaces (auth required)

```bash
curl https://superchargeclaudecode.com/marketplaces \
  -H "Authorization: Bearer <token>"
```

### Create Marketplace (auth required)

```bash
curl -X POST https://superchargeclaudecode.com/marketplaces \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"name": "My Marketplace", "slug": "my-marketplace", "description": "..."}'
```

### Get Marketplace by Slug (public)

```bash
curl https://superchargeclaudecode.com/api/marketplaces/:slug
```

Public page served at `/m/:slug`.

### Update Marketplace (auth required)

```bash
curl -X PUT https://superchargeclaudecode.com/marketplaces/:id \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"name": "Updated Name", "description": "..."}'
```

### Delete Marketplace (auth required)

```bash
curl -X DELETE https://superchargeclaudecode.com/marketplaces/:id \
  -H "Authorization: Bearer <token>"
```

### Add Plugin to Marketplace (auth required)

```bash
curl -X POST https://superchargeclaudecode.com/marketplaces/:id/plugins \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"pluginId": "<plugin-id>", "category": "Marketing"}'
```

The `category` field is optional. Used for per-plugin categorization within a marketplace.

### Remove Plugin from Marketplace (auth required)

```bash
curl -X DELETE https://superchargeclaudecode.com/marketplaces/:id/plugins/:pluginId \
  -H "Authorization: Bearer <token>"
```

### Marketplace JSON (public)

Standard marketplace.json format compatible with `claude plugin marketplace add`.

```bash
curl https://superchargeclaudecode.com/api/marketplaces/:slug/marketplace.json
```

Install a custom marketplace in Claude Code:

```bash
claude plugin marketplace add https://superchargeclaudecode.com/api/marketplaces/<slug>/marketplace.json
```

## Curated Marketplace

The main curated marketplace is **Superbot Marketplace** at `/m/superbot-marketplace`.

Categories: Marketing, Landing Pages, Web Applications, Scraping, Communication.

Install:

```bash
claude plugin marketplace add https://superchargeclaudecode.com/api/marketplaces/superbot-marketplace/marketplace.json
```

## Deployment

- **Heroku app**: supercharge-claude-code
- **Source**: ~/dev/personal/plugin-viewer
- **Deploy**: `git push heroku main` from the source directory
- **Post-build**: prisma generate, build, db push

## API Docs

Full documentation with curl examples available at [superchargeclaudecode.com/docs](https://superchargeclaudecode.com/docs) (API Reference tab).
