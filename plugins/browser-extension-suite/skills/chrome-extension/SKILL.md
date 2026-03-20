---
name: chrome-extension
description: >
  Chrome extension development with Manifest V3 — service workers, content
  scripts, popup/sidepanel UIs, message passing, storage, permissions,
  context menus, alarms, and TypeScript patterns.
  Triggers: "chrome extension", "browser extension", "manifest v3",
  "content script", "service worker extension", "popup extension",
  "side panel extension".
  NOT for: Firefox-only addons or Web Extensions polyfill.
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Chrome Extension Development (Manifest V3)

## Project Structure

```
my-extension/
├── manifest.json
├── src/
│   ├── background/
│   │   └── service-worker.ts      # Background service worker
│   ├── content/
│   │   └── content-script.ts      # Injected into web pages
│   ├── popup/
│   │   ├── popup.html
│   │   ├── popup.tsx              # React popup UI
│   │   └── popup.css
│   ├── sidepanel/
│   │   ├── sidepanel.html
│   │   └── sidepanel.tsx
│   ├── options/
│   │   ├── options.html
│   │   └── options.tsx
│   └── shared/
│       ├── types.ts
│       ├── storage.ts             # Typed storage helpers
│       └── messages.ts            # Typed message passing
├── public/
│   └── icons/
│       ├── icon-16.png
│       ├── icon-48.png
│       └── icon-128.png
├── vite.config.ts
├── tsconfig.json
└── package.json
```

## Manifest V3

```json
{
  "manifest_version": 3,
  "name": "My Extension",
  "version": "1.0.0",
  "description": "Does useful things",

  "permissions": [
    "storage",
    "activeTab",
    "alarms",
    "contextMenus",
    "sidePanel"
  ],

  "optional_permissions": [
    "tabs",
    "history",
    "notifications"
  ],

  "host_permissions": [
    "https://api.example.com/*"
  ],

  "background": {
    "service_worker": "src/background/service-worker.js",
    "type": "module"
  },

  "content_scripts": [
    {
      "matches": ["https://*.example.com/*"],
      "js": ["src/content/content-script.js"],
      "css": ["src/content/content-style.css"],
      "run_at": "document_idle"
    }
  ],

  "action": {
    "default_popup": "src/popup/popup.html",
    "default_icon": {
      "16": "public/icons/icon-16.png",
      "48": "public/icons/icon-48.png",
      "128": "public/icons/icon-128.png"
    }
  },

  "side_panel": {
    "default_path": "src/sidepanel/sidepanel.html"
  },

  "options_page": "src/options/options.html",

  "icons": {
    "16": "public/icons/icon-16.png",
    "48": "public/icons/icon-48.png",
    "128": "public/icons/icon-128.png"
  },

  "web_accessible_resources": [
    {
      "resources": ["assets/*"],
      "matches": ["<all_urls>"]
    }
  ],

  "content_security_policy": {
    "extension_pages": "script-src 'self'; object-src 'self'"
  }
}
```

## Service Worker (Background)

```typescript
// src/background/service-worker.ts

// Service worker lifecycle
chrome.runtime.onInstalled.addListener((details) => {
  if (details.reason === "install") {
    // First install
    chrome.storage.local.set({ settings: defaultSettings });
    chrome.contextMenus.create({
      id: "my-action",
      title: "Do Something",
      contexts: ["selection"],
    });
  } else if (details.reason === "update") {
    // Extension updated
    console.log(`Updated from ${details.previousVersion}`);
  }
});

// Handle messages from content scripts, popup, sidepanel
chrome.runtime.onMessage.addListener((message, sender, sendResponse) => {
  // MUST return true for async responses
  if (message.type === "FETCH_DATA") {
    fetchData(message.payload)
      .then((data) => sendResponse({ success: true, data }))
      .catch((error) => sendResponse({ success: false, error: error.message }));
    return true; // Keep message channel open for async response
  }

  if (message.type === "GET_SETTINGS") {
    chrome.storage.local.get("settings").then((result) => {
      sendResponse(result.settings);
    });
    return true;
  }
});

// Context menu click handler
chrome.contextMenus.onClicked.addListener(async (info, tab) => {
  if (info.menuItemId === "my-action" && tab?.id) {
    const selectedText = info.selectionText;
    // Do something with selected text
    await chrome.tabs.sendMessage(tab.id, {
      type: "PROCESS_SELECTION",
      text: selectedText,
    });
  }
});

// Alarms (periodic tasks — NOT setInterval!)
chrome.alarms.create("check-updates", { periodInMinutes: 30 });

chrome.alarms.onAlarm.addListener(async (alarm) => {
  if (alarm.name === "check-updates") {
    const data = await fetch("https://api.example.com/updates").then((r) => r.json());
    await chrome.storage.local.set({ lastUpdate: data });

    // Badge notification
    if (data.hasNew) {
      chrome.action.setBadgeText({ text: String(data.count) });
      chrome.action.setBadgeBackgroundColor({ color: "#ef4444" });
    }
  }
});

// Tab events
chrome.tabs.onUpdated.addListener(async (tabId, changeInfo, tab) => {
  if (changeInfo.status === "complete" && tab.url?.includes("example.com")) {
    // Programmatic injection (alternative to manifest content_scripts)
    await chrome.scripting.executeScript({
      target: { tabId },
      func: () => {
        document.body.style.border = "2px solid red";
      },
    });
  }
});

// Side panel
chrome.sidePanel.setPanelBehavior({ openPanelOnActionClick: true });
```

## Typed Message Passing

```typescript
// src/shared/messages.ts

// Define all message types
type Messages = {
  FETCH_DATA: { url: string; method?: string };
  GET_SETTINGS: void;
  SAVE_SETTINGS: { key: string; value: unknown };
  PROCESS_SELECTION: { text: string };
  CONTENT_READY: { url: string };
};

type MessageResponse<T extends keyof Messages> =
  T extends "FETCH_DATA" ? { success: boolean; data?: unknown; error?: string } :
  T extends "GET_SETTINGS" ? Settings :
  T extends "SAVE_SETTINGS" ? { ok: boolean } :
  void;

// Type-safe message sender
async function sendMessage<T extends keyof Messages>(
  type: T,
  payload: Messages[T]
): Promise<MessageResponse<T>> {
  return chrome.runtime.sendMessage({ type, payload });
}

// Type-safe message sender to content script
async function sendToTab<T extends keyof Messages>(
  tabId: number,
  type: T,
  payload: Messages[T]
): Promise<MessageResponse<T>> {
  return chrome.tabs.sendMessage(tabId, { type, payload });
}

// Usage in popup:
const settings = await sendMessage("GET_SETTINGS", undefined);
const result = await sendMessage("FETCH_DATA", { url: "/api/data" });
```

## Typed Storage

```typescript
// src/shared/storage.ts

interface StorageSchema {
  settings: {
    theme: "light" | "dark";
    notifications: boolean;
    apiKey: string;
  };
  cache: Record<string, { data: unknown; expiresAt: number }>;
  lastSync: number;
}

// Type-safe storage helpers
async function getStorage<K extends keyof StorageSchema>(
  key: K
): Promise<StorageSchema[K] | undefined> {
  const result = await chrome.storage.local.get(key);
  return result[key];
}

async function setStorage<K extends keyof StorageSchema>(
  key: K,
  value: StorageSchema[K]
): Promise<void> {
  await chrome.storage.local.set({ [key]: value });
}

// Watch for changes
chrome.storage.onChanged.addListener((changes, area) => {
  if (area === "local" && changes.settings) {
    const { oldValue, newValue } = changes.settings;
    console.log("Settings changed:", { oldValue, newValue });
  }
});

// Cached fetch with storage
async function cachedFetch(url: string, ttlMs = 300_000): Promise<unknown> {
  const cache = await getStorage("cache") ?? {};
  const cached = cache[url];

  if (cached && cached.expiresAt > Date.now()) {
    return cached.data;
  }

  const data = await fetch(url).then((r) => r.json());
  cache[url] = { data, expiresAt: Date.now() + ttlMs };
  await setStorage("cache", cache);
  return data;
}
```

## Content Script

```typescript
// src/content/content-script.ts

// Content scripts run in the web page context but isolated JS world

// Listen for messages from service worker
chrome.runtime.onMessage.addListener((message, sender, sendResponse) => {
  if (message.type === "PROCESS_SELECTION") {
    const result = processText(message.text);
    // Modify the page DOM
    showOverlay(result);
    sendResponse({ ok: true });
  }
  return true;
});

// Send message to service worker (for API calls)
async function fetchViaBackground(url: string) {
  try {
    const response = await chrome.runtime.sendMessage({
      type: "FETCH_DATA",
      payload: { url },
    });
    return response.data;
  } catch (error) {
    // Extension context invalidated (extension updated while page open)
    if ((error as Error).message.includes("Extension context invalidated")) {
      showReloadBanner();
      return null;
    }
    throw error;
  }
}

// Inject UI into the page
function showOverlay(content: string) {
  // Remove existing overlay
  document.getElementById("my-ext-overlay")?.remove();

  const overlay = document.createElement("div");
  overlay.id = "my-ext-overlay";
  overlay.attachShadow({ mode: "closed" }); // Shadow DOM for style isolation

  const shadow = overlay.shadowRoot!;
  shadow.innerHTML = `
    <style>
      .container {
        position: fixed;
        bottom: 20px;
        right: 20px;
        background: white;
        border-radius: 12px;
        padding: 16px;
        box-shadow: 0 4px 24px rgba(0,0,0,0.15);
        z-index: 999999;
        font-family: system-ui, sans-serif;
        max-width: 360px;
      }
      .close { cursor: pointer; float: right; }
    </style>
    <div class="container">
      <span class="close" id="close">&times;</span>
      <div>${content}</div>
    </div>
  `;

  shadow.getElementById("close")?.addEventListener("click", () => overlay.remove());
  document.body.appendChild(overlay);
}

// MutationObserver for dynamic pages (SPAs)
const observer = new MutationObserver((mutations) => {
  for (const mutation of mutations) {
    for (const node of mutation.addedNodes) {
      if (node instanceof HTMLElement && node.matches(".target-selector")) {
        processElement(node);
      }
    }
  }
});

observer.observe(document.body, { childList: true, subtree: true });

// Cleanup on navigation (SPA)
window.addEventListener("beforeunload", () => {
  observer.disconnect();
});
```

## Popup with React

```tsx
// src/popup/popup.tsx
import React, { useEffect, useState } from "react";
import { createRoot } from "react-dom/client";

function Popup() {
  const [settings, setSettings] = useState<Settings | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    chrome.storage.local.get("settings").then((result) => {
      setSettings(result.settings ?? defaultSettings);
      setLoading(false);
    });
  }, []);

  const handleToggle = async (key: keyof Settings) => {
    const updated = { ...settings!, [key]: !settings![key] };
    setSettings(updated);
    await chrome.storage.local.set({ settings: updated });
  };

  const openSidePanel = () => {
    chrome.tabs.query({ active: true, currentWindow: true }).then(([tab]) => {
      chrome.sidePanel.open({ tabId: tab.id! });
      window.close(); // Close popup
    });
  };

  if (loading) return <div className="p-4">Loading...</div>;

  return (
    <div className="w-[360px] p-4">
      <h1 className="text-lg font-bold mb-4">My Extension</h1>

      <div className="space-y-3">
        <label className="flex items-center gap-2">
          <input
            type="checkbox"
            checked={settings?.notifications}
            onChange={() => handleToggle("notifications")}
          />
          Enable notifications
        </label>

        <button
          onClick={openSidePanel}
          className="w-full py-2 px-4 bg-blue-600 text-white rounded-lg"
        >
          Open Side Panel
        </button>
      </div>
    </div>
  );
}

createRoot(document.getElementById("root")!).render(<Popup />);
```

## Vite Build Config

```typescript
// vite.config.ts
import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import { resolve } from "path";

export default defineConfig({
  plugins: [react()],
  build: {
    rollupOptions: {
      input: {
        popup: resolve(__dirname, "src/popup/popup.html"),
        sidepanel: resolve(__dirname, "src/sidepanel/sidepanel.html"),
        options: resolve(__dirname, "src/options/options.html"),
        "service-worker": resolve(__dirname, "src/background/service-worker.ts"),
        "content-script": resolve(__dirname, "src/content/content-script.ts"),
      },
      output: {
        entryFileNames: "[name].js",
        chunkFileNames: "chunks/[name]-[hash].js",
        assetFileNames: "assets/[name]-[hash][extname]",
      },
    },
    outDir: "dist",
    emptyOutDir: true,
  },
  resolve: {
    alias: {
      "@": resolve(__dirname, "src"),
    },
  },
});
```

## Declarative Net Request (URL Blocking/Modifying)

```json
// manifest.json additions
{
  "permissions": ["declarativeNetRequest"],
  "declarative_net_request": {
    "rule_resources": [
      {
        "id": "rules",
        "enabled": true,
        "path": "rules.json"
      }
    ]
  }
}
```

```json
// rules.json
[
  {
    "id": 1,
    "priority": 1,
    "action": { "type": "block" },
    "condition": {
      "urlFilter": "tracking.example.com",
      "resourceTypes": ["script", "xmlhttprequest"]
    }
  },
  {
    "id": 2,
    "priority": 1,
    "action": {
      "type": "modifyHeaders",
      "requestHeaders": [
        { "header": "X-Custom", "operation": "set", "value": "extension" }
      ]
    },
    "condition": {
      "urlFilter": "api.example.com/*",
      "resourceTypes": ["xmlhttprequest"]
    }
  }
]
```

## Debugging

```bash
# Load unpacked extension
# 1. Open chrome://extensions
# 2. Enable "Developer mode"
# 3. Click "Load unpacked" -> select dist/ folder

# Service worker logs
# Click "Inspect views: service worker" on extensions page

# Content script logs
# Open DevTools on the target page -> Console
# Filter by extension name in the source dropdown

# Popup logs
# Right-click extension icon -> "Inspect popup"
```

## Gotchas

1. **Service workers terminate after ~30 seconds idle** — Never store state in variables. Use `chrome.storage.local` for persistence. Never use `setInterval` — use `chrome.alarms` (minimum 1 minute). Your code must handle cold starts gracefully.

2. **`return true` in onMessage for async responses** — If your `onMessage` handler does async work, you MUST `return true` synchronously. Otherwise the message channel closes before `sendResponse` is called. This is the #1 extension debugging headache.

3. **Content scripts can't use most chrome.\* APIs** — Content scripts only have `chrome.runtime` and `chrome.storage`. For everything else (`chrome.tabs`, `chrome.alarms`, `fetch` to external APIs), send a message to the service worker and let it do the work.

4. **Shadow DOM for content script UI** — If you inject UI elements into a web page, use Shadow DOM (`attachShadow`) to isolate your styles. Otherwise the page's CSS will break your UI and vice versa.

5. **Extension context invalidation** — When you update/reload the extension, existing content scripts become orphaned. Their `chrome.runtime.sendMessage()` calls throw "Extension context invalidated". Always wrap in try/catch and show a "please reload the page" message.

6. **`activeTab` vs `host_permissions`** — `activeTab` grants temporary access to the current tab when the user clicks your extension icon. No scary install warning. `host_permissions` grants permanent access to matching URLs. Use `activeTab` when you only need access on user click.
