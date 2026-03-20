---
name: pwa-setup
description: >
  PWA setup and configuration — Web App Manifest, installability,
  app icons, splash screens, theme colors, display modes, and
  the install prompt UX.
  Triggers: "pwa", "progressive web app", "web manifest",
  "installable web app", "add to home screen", "pwa manifest".
  NOT for: Service worker caching (use service-worker-development).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# PWA Setup

## Web App Manifest

```json
// public/manifest.json (or manifest.webmanifest)
{
  "name": "My Progressive Web App",
  "short_name": "MyPWA",
  "description": "A fast, reliable, installable web app",
  "start_url": "/",
  "scope": "/",
  "display": "standalone",
  "orientation": "any",
  "theme_color": "#6366f1",
  "background_color": "#ffffff",
  "dir": "ltr",
  "lang": "en",

  "icons": [
    {
      "src": "/icons/icon-72.png",
      "sizes": "72x72",
      "type": "image/png"
    },
    {
      "src": "/icons/icon-96.png",
      "sizes": "96x96",
      "type": "image/png"
    },
    {
      "src": "/icons/icon-128.png",
      "sizes": "128x128",
      "type": "image/png"
    },
    {
      "src": "/icons/icon-144.png",
      "sizes": "144x144",
      "type": "image/png"
    },
    {
      "src": "/icons/icon-152.png",
      "sizes": "152x152",
      "type": "image/png"
    },
    {
      "src": "/icons/icon-192.png",
      "sizes": "192x192",
      "type": "image/png",
      "purpose": "any"
    },
    {
      "src": "/icons/icon-384.png",
      "sizes": "384x384",
      "type": "image/png"
    },
    {
      "src": "/icons/icon-512.png",
      "sizes": "512x512",
      "type": "image/png",
      "purpose": "any"
    },
    {
      "src": "/icons/icon-maskable-192.png",
      "sizes": "192x192",
      "type": "image/png",
      "purpose": "maskable"
    },
    {
      "src": "/icons/icon-maskable-512.png",
      "sizes": "512x512",
      "type": "image/png",
      "purpose": "maskable"
    }
  ],

  "screenshots": [
    {
      "src": "/screenshots/desktop.png",
      "sizes": "1280x720",
      "type": "image/png",
      "form_factor": "wide",
      "label": "Desktop view"
    },
    {
      "src": "/screenshots/mobile.png",
      "sizes": "750x1334",
      "type": "image/png",
      "form_factor": "narrow",
      "label": "Mobile view"
    }
  ],

  "categories": ["productivity", "utilities"],

  "shortcuts": [
    {
      "name": "New Task",
      "short_name": "New",
      "description": "Create a new task",
      "url": "/tasks/new",
      "icons": [{ "src": "/icons/shortcut-new.png", "sizes": "96x96" }]
    },
    {
      "name": "Dashboard",
      "url": "/dashboard",
      "icons": [{ "src": "/icons/shortcut-dashboard.png", "sizes": "96x96" }]
    }
  ],

  "share_target": {
    "action": "/share",
    "method": "POST",
    "enctype": "multipart/form-data",
    "params": {
      "title": "title",
      "text": "text",
      "url": "url",
      "files": [
        {
          "name": "media",
          "accept": ["image/*", "video/*"]
        }
      ]
    }
  }
}
```

## HTML Head Tags

```html
<head>
  <!-- Manifest -->
  <link rel="manifest" href="/manifest.json" />

  <!-- Theme color (browser toolbar) -->
  <meta name="theme-color" content="#6366f1" />
  <meta name="theme-color" content="#1e1b4b" media="(prefers-color-scheme: dark)" />

  <!-- iOS / Safari -->
  <meta name="apple-mobile-web-app-capable" content="yes" />
  <meta name="apple-mobile-web-app-status-bar-style" content="black-translucent" />
  <meta name="apple-mobile-web-app-title" content="MyPWA" />
  <link rel="apple-touch-icon" href="/icons/apple-touch-icon.png" />

  <!-- iOS splash screens -->
  <link rel="apple-touch-startup-image"
    href="/splash/splash-1170x2532.png"
    media="(device-width: 390px) and (device-height: 844px) and (-webkit-device-pixel-ratio: 3)" />
  <link rel="apple-touch-startup-image"
    href="/splash/splash-1284x2778.png"
    media="(device-width: 428px) and (device-height: 926px) and (-webkit-device-pixel-ratio: 3)" />

  <!-- Windows -->
  <meta name="msapplication-TileColor" content="#6366f1" />
  <meta name="msapplication-TileImage" content="/icons/icon-144.png" />

  <!-- Standard favicon -->
  <link rel="icon" type="image/png" sizes="32x32" href="/icons/favicon-32.png" />
  <link rel="icon" type="image/png" sizes="16x16" href="/icons/favicon-16.png" />
</head>
```

## Custom Install Prompt

```tsx
import { useEffect, useState } from "react";

interface BeforeInstallPromptEvent extends Event {
  prompt: () => Promise<void>;
  userChoice: Promise<{ outcome: "accepted" | "dismissed" }>;
}

function useInstallPrompt() {
  const [deferredPrompt, setDeferredPrompt] = useState<BeforeInstallPromptEvent | null>(null);
  const [isInstalled, setIsInstalled] = useState(false);

  useEffect(() => {
    // Check if already installed
    if (window.matchMedia("(display-mode: standalone)").matches) {
      setIsInstalled(true);
      return;
    }

    const handler = (e: Event) => {
      e.preventDefault(); // Prevent default browser prompt
      setDeferredPrompt(e as BeforeInstallPromptEvent);
    };

    window.addEventListener("beforeinstallprompt", handler);
    window.addEventListener("appinstalled", () => setIsInstalled(true));

    return () => window.removeEventListener("beforeinstallprompt", handler);
  }, []);

  const install = async () => {
    if (!deferredPrompt) return false;

    await deferredPrompt.prompt();
    const { outcome } = await deferredPrompt.userChoice;
    setDeferredPrompt(null);
    return outcome === "accepted";
  };

  return { canInstall: !!deferredPrompt && !isInstalled, isInstalled, install };
}

// Usage
function InstallButton() {
  const { canInstall, isInstalled, install } = useInstallPrompt();

  if (isInstalled) return null;
  if (!canInstall) return null;

  return (
    <button
      onClick={install}
      className="fixed bottom-4 left-4 bg-indigo-600 text-white px-4 py-2 rounded-lg shadow-lg flex items-center gap-2"
    >
      <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16v2a2 2 0 002 2h12a2 2 0 002-2v-2M12 4v12m0 0l-4-4m4 4l4-4" />
      </svg>
      Install App
    </button>
  );
}
```

## Display Modes

```css
/* Detect installed PWA mode */
@media (display-mode: standalone) {
  /* Styles for installed PWA */
  .install-banner { display: none; }
  .app-titlebar { display: flex; }
}

@media (display-mode: browser) {
  /* Styles for regular browser tab */
  .install-banner { display: flex; }
  .app-titlebar { display: none; }
}
```

```typescript
// Detect in JS
const isInstalled = window.matchMedia("(display-mode: standalone)").matches
  || (window.navigator as any).standalone === true; // iOS Safari
```

## Icon Generation

```bash
# Using sharp (Node.js)
npm install sharp
```

```typescript
// scripts/generate-icons.ts
import sharp from "sharp";
import { mkdirSync } from "fs";

const sizes = [16, 32, 72, 96, 128, 144, 152, 192, 384, 512];
const sourceIcon = "source-icon.png"; // 1024x1024 recommended

mkdirSync("public/icons", { recursive: true });

for (const size of sizes) {
  await sharp(sourceIcon)
    .resize(size, size)
    .png()
    .toFile(`public/icons/icon-${size}.png`);
}

// Maskable icon (with safe zone padding)
for (const size of [192, 512]) {
  const padding = Math.round(size * 0.1); // 10% safe zone
  const innerSize = size - padding * 2;

  await sharp(sourceIcon)
    .resize(innerSize, innerSize)
    .extend({
      top: padding,
      bottom: padding,
      left: padding,
      right: padding,
      background: { r: 99, g: 102, b: 241, alpha: 1 }, // theme color
    })
    .png()
    .toFile(`public/icons/icon-maskable-${size}.png`);
}

// Apple touch icon
await sharp(sourceIcon).resize(180, 180).png().toFile("public/icons/apple-touch-icon.png");
```

## Offline Page

```html
<!-- public/offline.html -->
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <title>Offline - MyPWA</title>
  <style>
    body {
      font-family: system-ui, sans-serif;
      display: flex;
      align-items: center;
      justify-content: center;
      min-height: 100vh;
      margin: 0;
      background: #f9fafb;
      color: #374151;
    }
    .container { text-align: center; padding: 2rem; }
    .icon { font-size: 4rem; margin-bottom: 1rem; }
    h1 { font-size: 1.5rem; margin-bottom: 0.5rem; }
    p { color: #6b7280; margin-bottom: 1.5rem; }
    button {
      background: #6366f1; color: white; border: none;
      padding: 0.75rem 1.5rem; border-radius: 0.5rem;
      font-size: 1rem; cursor: pointer;
    }
  </style>
</head>
<body>
  <div class="container">
    <div class="icon">📡</div>
    <h1>You're offline</h1>
    <p>Check your connection and try again.</p>
    <button onclick="window.location.reload()">Retry</button>
  </div>
</body>
</html>
```

## Lighthouse PWA Audit

```bash
# Run Lighthouse PWA audit
npx lighthouse https://your-app.com --only-categories=pwa --output=json --output-path=./pwa-audit.json

# Key checks:
# - Installability (manifest + SW + HTTPS)
# - PWA Optimized (theme-color, viewport, apple-touch-icon)
# - Offline support (200 on offline navigation)
# - Fast and reliable (SW-controlled, fast first paint)
```

## Gotchas

1. **Maskable icons need safe zone** — Maskable icons get cropped into circles/rounded shapes by the OS. Keep the important content within the center 80% (10% padding on each side). Use https://maskable.app to test.

2. **`start_url` must be within SW scope** — If your SW is scoped to `/app/`, then `start_url` must be `/app/` or a subpath. Mismatched scope prevents installation.

3. **iOS Safari doesn't support `beforeinstallprompt`** — On iOS, there's no programmatic install prompt. You must show manual instructions: "Tap Share > Add to Home Screen". Detect iOS: `/(iPhone|iPod|iPad)/.test(navigator.userAgent)`.

4. **Theme color changes require manifest update** — Changing `theme_color` in the manifest requires the SW to update and activate. It doesn't take effect immediately. For dynamic themes, use the `<meta name="theme-color">` tag which updates instantly.

5. **Screenshots trigger richer install UI** — Chrome 115+ shows a richer install dialog when the manifest includes `screenshots` with `form_factor` set. Without screenshots, users get the minimal "Add to Home Screen" prompt.

6. **`display: "standalone"` hides the URL bar** — This means users can't see the URL and may not trust the app. Include visible navigation and a way to share/copy the current URL within your app UI.
