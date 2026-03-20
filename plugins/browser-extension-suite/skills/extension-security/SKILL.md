---
name: extension-security
description: >
  Browser extension security patterns, CSP configuration, and secure messaging.
  Use when hardening extension permissions, implementing content security policies,
  securing message passing, or handling sensitive data in extensions.
  Triggers: "extension security", "CSP extension", "content_security_policy",
  "extension permissions", "message passing security", "chrome extension XSS",
  "extension storage security", "manifest permissions".
  NOT for: web app security (see security-audit-suite), server-side security, OAuth flows.
version: 1.0.0
allowed-tools: Read, Grep, Glob, Bash
---

# Browser Extension Security

## Manifest V3 Permission Model

```json
{
  "manifest_version": 3,
  "permissions": [
    "storage",
    "alarms",
    "notifications"
  ],
  "optional_permissions": [
    "tabs",
    "bookmarks",
    "history"
  ],
  "host_permissions": [
    "https://api.myservice.com/*"
  ],
  "optional_host_permissions": [
    "https://*/*"
  ],
  "content_security_policy": {
    "extension_pages": "script-src 'self'; object-src 'none'; style-src 'self' 'unsafe-inline'",
    "sandbox": "sandbox allow-scripts; script-src 'self' https://cdn.example.com"
  }
}
```

```typescript
// runtime-permission-request.ts — Request optional permissions only when needed
async function requestTabsPermission(): Promise<boolean> {
  // Check if already granted
  const hasPermission = await chrome.permissions.contains({
    permissions: ['tabs'],
  });
  if (hasPermission) return true;

  // Request with user gesture (must be in click handler)
  try {
    const granted = await chrome.permissions.request({
      permissions: ['tabs'],
    });
    if (granted) {
      console.log('Tabs permission granted');
    }
    return granted;
  } catch (error) {
    // User denied or not triggered from user gesture
    console.error('Permission request failed:', error);
    return false;
  }
}

// Remove permissions when no longer needed
async function revokeTabsPermission(): Promise<void> {
  await chrome.permissions.remove({ permissions: ['tabs'] });
}
```

## Secure Message Passing

```typescript
// background.ts — Validate ALL incoming messages
interface MessageSchema {
  type: string;
  payload?: unknown;
}

const ALLOWED_ACTIONS = new Set([
  'GET_SETTINGS',
  'SAVE_SETTINGS',
  'FETCH_DATA',
  'LOG_EVENT',
]);

chrome.runtime.onMessage.addListener(
  (message: unknown, sender: chrome.runtime.MessageSender, sendResponse) => {
    // 1. Validate sender origin
    if (!sender.id || sender.id !== chrome.runtime.id) {
      console.warn('Message from unknown extension:', sender.id);
      return false;
    }

    // 2. Validate message structure
    if (!isValidMessage(message)) {
      console.warn('Invalid message format:', message);
      sendResponse({ error: 'Invalid message format' });
      return false;
    }

    // 3. Validate action is allowed
    if (!ALLOWED_ACTIONS.has(message.type)) {
      console.warn('Unknown action:', message.type);
      sendResponse({ error: 'Unknown action' });
      return false;
    }

    // 4. Handle with type-safe dispatch
    handleMessage(message, sender)
      .then(result => sendResponse({ data: result }))
      .catch(err => sendResponse({ error: err.message }));

    return true; // Keep message channel open for async response
  }
);

function isValidMessage(msg: unknown): msg is MessageSchema {
  return (
    typeof msg === 'object' &&
    msg !== null &&
    'type' in msg &&
    typeof (msg as MessageSchema).type === 'string'
  );
}

// External message handling (from web pages)
chrome.runtime.onMessageExternal.addListener(
  (message, sender, sendResponse) => {
    // STRICT: Only allow from specific origins
    const ALLOWED_ORIGINS = [
      'https://myapp.com',
      'https://dashboard.myapp.com',
    ];

    if (!sender.origin || !ALLOWED_ORIGINS.includes(sender.origin)) {
      console.warn('External message from unauthorized origin:', sender.origin);
      return false;
    }

    // Only expose a minimal API to external callers
    if (message.type === 'GET_VERSION') {
      sendResponse({ version: chrome.runtime.getManifest().version });
    }

    return false;
  }
);
```

## Content Script Isolation

```typescript
// content-script.ts — Safe DOM interaction from content scripts

// NEVER inject raw HTML from untrusted sources
function injectUI(data: { title: string; count: number }): void {
  const container = document.createElement('div');
  container.id = 'my-extension-root';

  // Use Shadow DOM for style isolation
  const shadow = container.attachShadow({ mode: 'closed' });

  // Create elements safely — NEVER use innerHTML with untrusted data
  const title = document.createElement('h3');
  title.textContent = sanitizeText(data.title); // textContent is safe
  shadow.appendChild(title);

  const count = document.createElement('span');
  count.textContent = String(data.count);
  shadow.appendChild(count);

  // Inject scoped styles
  const style = document.createElement('style');
  style.textContent = `
    h3 { color: #333; font-size: 16px; }
    span { color: #666; }
  `;
  shadow.appendChild(style);

  document.body.appendChild(container);
}

// Sanitize text content (defense in depth)
function sanitizeText(input: string): string {
  const div = document.createElement('div');
  div.textContent = input;
  return div.textContent || '';
}

// NEVER do this:
// element.innerHTML = `<div>${untrustedData}</div>`; // XSS vulnerability

// NEVER evaluate page scripts from content script context:
// eval(pageData); // Code injection

// Safe way to read page data
function readPageData(selector: string): string | null {
  const element = document.querySelector(selector);
  return element?.textContent?.trim() ?? null;
}
```

## Secure Storage Patterns

```typescript
// storage.ts — Encrypted storage wrapper

// Use chrome.storage.session for sensitive ephemeral data (MV3)
// Data is cleared when browser closes, never written to disk
async function storeSessionToken(token: string): Promise<void> {
  await chrome.storage.session.set({ authToken: token });
}

async function getSessionToken(): Promise<string | null> {
  const result = await chrome.storage.session.get('authToken');
  return result.authToken ?? null;
}

// For persistent sensitive data, encrypt before storing
import { webcrypto } from 'crypto';

class SecureStorage {
  private key: CryptoKey | null = null;

  async init(passphrase: string): Promise<void> {
    const encoder = new TextEncoder();
    const keyMaterial = await crypto.subtle.importKey(
      'raw',
      encoder.encode(passphrase),
      'PBKDF2',
      false,
      ['deriveKey']
    );

    this.key = await crypto.subtle.deriveKey(
      {
        name: 'PBKDF2',
        salt: encoder.encode('extension-salt-v1'),
        iterations: 100000,
        hash: 'SHA-256',
      },
      keyMaterial,
      { name: 'AES-GCM', length: 256 },
      false,
      ['encrypt', 'decrypt']
    );
  }

  async encrypt(data: string): Promise<string> {
    if (!this.key) throw new Error('SecureStorage not initialized');
    const encoder = new TextEncoder();
    const iv = crypto.getRandomValues(new Uint8Array(12));
    const encrypted = await crypto.subtle.encrypt(
      { name: 'AES-GCM', iv },
      this.key,
      encoder.encode(data)
    );

    // Combine IV + ciphertext for storage
    const combined = new Uint8Array(iv.length + new Uint8Array(encrypted).length);
    combined.set(iv);
    combined.set(new Uint8Array(encrypted), iv.length);
    return btoa(String.fromCharCode(...combined));
  }

  async decrypt(encoded: string): Promise<string> {
    if (!this.key) throw new Error('SecureStorage not initialized');
    const combined = new Uint8Array(
      atob(encoded).split('').map(c => c.charCodeAt(0))
    );
    const iv = combined.slice(0, 12);
    const data = combined.slice(12);

    const decrypted = await crypto.subtle.decrypt(
      { name: 'AES-GCM', iv },
      this.key,
      data
    );
    return new TextDecoder().decode(decrypted);
  }
}

// Usage:
// const storage = new SecureStorage();
// await storage.init(userPassphrase);
// await chrome.storage.local.set({ apiKey: await storage.encrypt(key) });
```

## Network Request Security

```typescript
// api.ts — Secure API communication from extension

async function secureFetch<T>(
  url: string,
  options: RequestInit = {}
): Promise<T> {
  // 1. Validate URL against allowlist
  const allowedHosts = ['api.myservice.com', 'auth.myservice.com'];
  const parsed = new URL(url);

  if (!allowedHosts.includes(parsed.hostname)) {
    throw new Error(`Blocked request to unauthorized host: ${parsed.hostname}`);
  }

  if (parsed.protocol !== 'https:') {
    throw new Error('HTTPS required for all API requests');
  }

  // 2. Add auth token from session storage
  const token = await getSessionToken();
  const headers = new Headers(options.headers);
  if (token) {
    headers.set('Authorization', `Bearer ${token}`);
  }
  headers.set('X-Extension-Version', chrome.runtime.getManifest().version);

  // 3. Make request with timeout
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), 10000);

  try {
    const response = await fetch(url, {
      ...options,
      headers,
      signal: controller.signal,
    });

    if (response.status === 401) {
      // Token expired — clear and notify
      await chrome.storage.session.remove('authToken');
      throw new Error('Authentication expired');
    }

    if (!response.ok) {
      throw new Error(`API error: ${response.status} ${response.statusText}`);
    }

    return await response.json() as T;
  } finally {
    clearTimeout(timeout);
  }
}
```

## Gotchas

1. **chrome.storage is not encrypted** -- `chrome.storage.local` and `chrome.storage.sync` store data in plain text on disk. Anyone with filesystem access can read it. Use `chrome.storage.session` for sensitive ephemeral data (never touches disk), and encrypt anything sensitive before writing to local/sync storage.

2. **Content scripts share the page DOM** -- Content scripts run in an isolated JS world but share the same DOM as the page. A malicious page can modify DOM elements your content script reads. Never trust `element.dataset`, `element.getAttribute()`, or DOM-sourced data for security decisions. Validate everything.

3. **innerHTML in content scripts is XSS** -- Setting `innerHTML` with any data sourced from the page, messages, or API responses creates XSS vulnerabilities. Use `textContent` for text, `createElement` + `appendChild` for structure, or a trusted sanitizer library. Shadow DOM helps but doesn't prevent innerHTML injection within the shadow root.

4. **Optional permissions UX** -- `chrome.permissions.request()` must be called from a user gesture (click handler). Calling it from `setTimeout`, `fetch.then()`, or page load fails silently. The request dialog appears only once per permission per session — if denied, you can't re-prompt immediately.

5. **Message passing has no built-in auth** -- Any content script or extension page can send messages via `chrome.runtime.sendMessage`. The background script must validate `sender.id === chrome.runtime.id` for internal messages and check `sender.origin` against an allowlist for external messages. Without validation, malicious extensions or pages can trigger your background actions.

6. **Host permissions expose all page data** -- A host permission like `https://*/*` gives your extension access to ALL HTTPS sites. Chrome Web Store reviewers flag broad permissions. Request the minimum necessary, use `optional_host_permissions` for on-demand access, and document why each permission is needed.
