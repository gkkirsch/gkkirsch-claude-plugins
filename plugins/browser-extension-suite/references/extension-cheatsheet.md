# Browser Extension Cheatsheet

## Manifest V3 Minimal

```json
{
  "manifest_version": 3,
  "name": "My Extension",
  "version": "1.0.0",
  "permissions": ["storage", "activeTab"],
  "background": { "service_worker": "service-worker.js", "type": "module" },
  "action": { "default_popup": "popup.html" },
  "content_scripts": [{
    "matches": ["https://*.example.com/*"],
    "js": ["content.js"]
  }]
}
```

## Service Worker Patterns

```typescript
// Install
chrome.runtime.onInstalled.addListener(({ reason }) => {
  if (reason === "install") { /* first install */ }
});

// Message handler (MUST return true for async)
chrome.runtime.onMessage.addListener((msg, sender, sendResponse) => {
  doAsyncWork().then(sendResponse);
  return true; // <-- CRITICAL for async
});

// Alarms (NOT setInterval — worker terminates!)
chrome.alarms.create("check", { periodInMinutes: 30 });
chrome.alarms.onAlarm.addListener((alarm) => { /* handle */ });

// Badge
chrome.action.setBadgeText({ text: "3" });
chrome.action.setBadgeBackgroundColor({ color: "#ef4444" });

// Context menu
chrome.contextMenus.create({ id: "action", title: "Do Thing", contexts: ["selection"] });
chrome.contextMenus.onClicked.addListener((info, tab) => { /* handle */ });

// Programmatic injection
chrome.scripting.executeScript({
  target: { tabId },
  func: () => { document.title; },
});

// Storage (NOT localStorage — no DOM in service worker)
await chrome.storage.local.set({ key: "value" });
const { key } = await chrome.storage.local.get("key");
```

## Content Script Patterns

```typescript
// Send to service worker
const response = await chrome.runtime.sendMessage({ type: "FETCH", url });

// Listen from service worker
chrome.runtime.onMessage.addListener((msg, sender, sendResponse) => {
  if (msg.type === "DO_THING") { /* modify DOM */ }
  return true;
});

// Shadow DOM for injected UI
const host = document.createElement("div");
const shadow = host.attachShadow({ mode: "closed" });
shadow.innerHTML = `<style>...</style><div>...</div>`;
document.body.appendChild(host);

// Handle extension updates
try {
  await chrome.runtime.sendMessage({ type: "PING" });
} catch (e) {
  if (e.message.includes("Extension context invalidated")) {
    showReloadBanner();
  }
}
```

## Message Passing

| From | To | Method |
|------|-----|--------|
| Content Script | Service Worker | `chrome.runtime.sendMessage()` |
| Service Worker | Content Script | `chrome.tabs.sendMessage(tabId, msg)` |
| Popup/Panel | Service Worker | `chrome.runtime.sendMessage()` |
| Service Worker | Popup/Panel | `chrome.runtime.sendMessage()` |

## Permissions Quick Reference

| Permission | Warning Level | Use Case |
|-----------|---------------|----------|
| `storage` | None | Store settings/data |
| `alarms` | None | Periodic tasks |
| `contextMenus` | None | Right-click menu |
| `sidePanel` | None | Side panel UI |
| `activeTab` | Low | Current tab on click |
| `notifications` | Low | Desktop notifications |
| `tabs` | Medium | Tab URLs and titles |
| `history` | Medium | Browsing history |
| `host_permissions` | High | Access specific sites |
| `<all_urls>` | High | Access all sites |

**Strategy**: Start with `activeTab` + `optional_permissions`. Request more via `chrome.permissions.request()` when needed.

## Storage Options

| API | Limit | Sync | Use For |
|-----|-------|------|---------|
| `chrome.storage.local` | 10 MB | No | Large data, cache |
| `chrome.storage.sync` | 100 KB | Yes (across devices) | Settings, preferences |
| `chrome.storage.session` | 10 MB | No (cleared on close) | Temporary state |

## Chrome Web Store Publishing

```bash
# Package
cd dist && zip -r ../extension.zip .

# Upload via CLI
npm install -D chrome-webstore-upload-cli
chrome-webstore-upload upload --source extension.zip
chrome-webstore-upload publish

# Required env vars
EXTENSION_ID, CLIENT_ID, CLIENT_SECRET, REFRESH_TOKEN
```

## Debugging

| What | Where |
|------|-------|
| Service worker | chrome://extensions → "Inspect views: service worker" |
| Content script | Page DevTools → Console (filter by extension) |
| Popup | Right-click icon → "Inspect popup" |
| Storage | DevTools → Application → Extension Storage |
| Network | DevTools → Network (filter by extension origin) |

## MV2 → MV3 Migration

| MV2 | MV3 |
|-----|-----|
| `chrome.browserAction` | `chrome.action` |
| `background.scripts` | `background.service_worker` |
| `chrome.tabs.executeScript` | `chrome.scripting.executeScript` |
| `chrome.tabs.insertCSS` | `chrome.scripting.insertCSS` |
| `webRequestBlocking` | `declarativeNetRequest` |
| `localStorage` (background) | `chrome.storage.local` |
| Persistent background page | Event-driven service worker |

## Cross-Browser Compatibility

```bash
npm install webextension-polyfill
```

```typescript
import browser from "webextension-polyfill";
// Use browser.* everywhere — works in Chrome and Firefox
const tabs = await browser.tabs.query({ active: true });
```

Firefox manifest addition:
```json
{
  "browser_specific_settings": {
    "gecko": { "id": "my-ext@example.com", "strict_min_version": "109.0" }
  }
}
```
