---
name: extension-publishing
description: >
  Browser extension publishing — Chrome Web Store submission, Firefox Add-ons,
  automated builds, versioning, store listing optimization, review process,
  and update distribution.
  Triggers: "publish extension", "chrome web store", "firefox addon",
  "extension store listing", "extension review", "extension update".
  NOT for: Extension development (use chrome-extension).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Extension Publishing

## Chrome Web Store

### Developer Account Setup

1. Go to [Chrome Web Store Developer Dashboard](https://chrome.google.com/webstore/devconsole)
2. Pay one-time $5 registration fee
3. Verify your email address
4. Set up a developer profile (name, website, icon)

### Build & Package

```bash
# Build the extension
npm run build

# Create a ZIP of the dist folder
cd dist && zip -r ../my-extension.zip . && cd ..

# Or use a script:
```

```json
// package.json
{
  "scripts": {
    "build": "vite build",
    "package": "npm run build && cd dist && zip -r ../my-extension.zip . && cd ..",
    "version:patch": "npm version patch --no-git-tag-version",
    "version:minor": "npm version minor --no-git-tag-version",
    "release": "npm run version:patch && npm run package"
  }
}
```

### Store Listing Requirements

```
Required:
├── Extension name (max 75 characters)
├── Summary (max 132 characters)
├── Description (detailed, max 16K characters)
├── Category (choose one)
├── Language
├── Screenshots (1280x800 or 640x400, 1-5 required)
├── Icon (128x128 PNG, in manifest.json)
└── Privacy policy URL (if using host_permissions or remote code)

Optional but recommended:
├── Promotional images (440x280 small, 920x680 large)
├── YouTube video URL
├── Homepage URL
└── Support URL
```

### Store Listing Optimization

```markdown
## Title Formula
[Primary Keyword] - [Value Proposition] (max 75 chars)
Example: "Tab Manager - Save Memory & Organize Tabs"

## Description Structure
Paragraph 1: What it does (elevator pitch)
Paragraph 2: Key features (bullet points)
Paragraph 3: How it works (simple steps)
Paragraph 4: Privacy commitment
Paragraph 5: Support/contact info

## Keywords Strategy
- Include keywords naturally in title, summary, and description
- Don't keyword-stuff — Google reviews manually
- Focus on what users search for: "tab manager" not "chromium tab lifecycle handler"
```

### Privacy Practices Disclosure

```yaml
# Required disclosures for Chrome Web Store
data_collection:
  - category: "Web history"
    purpose: "Functionality"      # or "Analytics", "Marketing"
    is_personal: false
    storage: "Local only"         # or "Transferred off device"

  - category: "User activity"
    purpose: "Functionality"
    is_personal: false
    storage: "Local only"

# If you DON'T collect data:
single_purpose: "Manages browser tabs for better organization"
no_data_collection: true
```

### Review Process

| Review Type | Duration | When |
|------------|----------|------|
| New extension | 1-7 business days | First submission |
| Update (no permission change) | Minutes to 1 day | Code-only updates |
| Update (permission change) | 1-3 business days | Adding new permissions |
| Flagged review | 3-14 business days | Policy concern detected |

**Common rejection reasons:**
1. Missing or inadequate privacy policy
2. Requesting unnecessary permissions
3. Description doesn't match functionality
4. Obfuscated code (minification is fine, obfuscation isn't)
5. Affiliate or promotional content without disclosure
6. Missing offline functionality disclosure

### Automated Publishing

```bash
# Install chrome-webstore-upload-cli
npm install -D chrome-webstore-upload-cli

# Set up API credentials (Google Cloud Console):
# 1. Create project at console.cloud.google.com
# 2. Enable Chrome Web Store API
# 3. Create OAuth2 credentials (Desktop app type)
# 4. Get refresh token via OAuth flow
```

```json
// package.json
{
  "scripts": {
    "upload": "chrome-webstore-upload upload --source my-extension.zip",
    "publish": "chrome-webstore-upload publish"
  }
}
```

```bash
# Environment variables
export EXTENSION_ID="your-extension-id"
export CLIENT_ID="your-client-id"
export CLIENT_SECRET="your-client-secret"
export REFRESH_TOKEN="your-refresh-token"

# Upload and publish
npm run package
npm run upload
npm run publish
```

### CI/CD Pipeline

```yaml
# .github/workflows/publish.yml
name: Publish Extension
on:
  push:
    tags: ["v*"]

jobs:
  publish:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with: { node-version: 20 }
      - run: npm ci
      - run: npm run build
      - run: cd dist && zip -r ../extension.zip .
      - name: Upload to Chrome Web Store
        uses: mnao305/chrome-extension-upload@v5.0.0
        with:
          file-path: extension.zip
          extension-id: ${{ secrets.EXTENSION_ID }}
          client-id: ${{ secrets.CLIENT_ID }}
          client-secret: ${{ secrets.CLIENT_SECRET }}
          refresh-token: ${{ secrets.REFRESH_TOKEN }}
          publish: true
```

## Firefox Add-ons

### Submission

1. Go to [Firefox Add-on Developer Hub](https://addons.mozilla.org/developers/)
2. Create a Firefox account
3. Submit add-on (ZIP upload)
4. Choose: listed (public AMO) or unlisted (self-distributed)

### Cross-Browser Manifest

```json
// Firefox differences from Chrome manifest:
{
  // Firefox uses "browser_specific_settings" instead of "key"
  "browser_specific_settings": {
    "gecko": {
      "id": "my-extension@example.com",
      "strict_min_version": "109.0"
    }
  },

  // Firefox supports both "scripts" (MV2-style) and "service_worker" (MV3)
  // For best compatibility, use service_worker
  "background": {
    "service_worker": "service-worker.js"
  }
}
```

### Web Extension Polyfill

```bash
npm install webextension-polyfill @types/webextension-polyfill
```

```typescript
// Use browser.* (Promise-based) instead of chrome.* (callback-based)
import browser from "webextension-polyfill";

// Works in both Chrome and Firefox
const tabs = await browser.tabs.query({ active: true, currentWindow: true });
await browser.storage.local.set({ key: "value" });
```

## Versioning Strategy

```typescript
// manifest.json version format: "major.minor.patch"
// Chrome Web Store requires version to increase on each upload

// Semantic versioning:
// major: Breaking changes (removed features, API changes)
// minor: New features (backward compatible)
// patch: Bug fixes

// Automated version bump:
// npm version patch -> 1.0.0 -> 1.0.1
// npm version minor -> 1.0.1 -> 1.1.0
// npm version major -> 1.1.0 -> 2.0.0

// Sync manifest.json version with package.json:
```

```javascript
// scripts/sync-version.js
const pkg = JSON.parse(fs.readFileSync("package.json", "utf8"));
const manifest = JSON.parse(fs.readFileSync("manifest.json", "utf8"));
manifest.version = pkg.version;
fs.writeFileSync("manifest.json", JSON.stringify(manifest, null, 2));
```

## Update Distribution

```typescript
// Service worker: check for updates and notify
chrome.runtime.onUpdateAvailable.addListener((details) => {
  console.log(`Update available: ${details.version}`);
  // Apply update immediately (or defer)
  chrome.runtime.reload();
});

// Content script: handle post-update gracefully
try {
  await chrome.runtime.sendMessage({ type: "PING" });
} catch (error) {
  if ((error as Error).message.includes("Extension context invalidated")) {
    // Extension was updated — show "reload page" banner
    showUpdateBanner();
  }
}
```

## Gotchas

1. **Version must always increase** — Chrome Web Store rejects uploads where the manifest version is less than or equal to the currently published version. Automate version bumps in your build script.

2. **Source code may be requested** — For extensions using minified/bundled code, Google may request your source code for review. Keep your build process reproducible and documented.

3. **Privacy policy is mandatory** — Required if you use `host_permissions`, access user data, or use remote code. Host it on a public URL. Template at https://nicescale.com/privacy-policy-for-chrome-extensions/.

4. **Screenshot dimensions matter** — Must be exactly 1280x800 or 640x400 pixels. Chrome Web Store rejects other sizes. Take screenshots at the exact resolution.

5. **Firefox requires explicit extension ID** — Unlike Chrome (which auto-generates), Firefox needs `browser_specific_settings.gecko.id` in manifest for updates to work. Use email-style format: `my-extension@example.com`.

6. **Updates can take hours to propagate** — After publishing, it can take 30-60 minutes for the update to appear to all users. Don't panic if the old version is still showing.
