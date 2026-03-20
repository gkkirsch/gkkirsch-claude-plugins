# PWA Cheatsheet

## Installability Requirements

| Requirement | Required |
|------------|----------|
| HTTPS (or localhost) | Yes |
| Service worker with fetch handler | Yes |
| Web App Manifest | Yes |
| `name` or `short_name` | Yes |
| `start_url` | Yes |
| `display`: standalone/fullscreen/minimal-ui | Yes |
| Icon 192x192 PNG | Yes |
| Icon 512x512 PNG | Yes |

## Manifest Minimal

```json
{
  "name": "My App",
  "short_name": "App",
  "start_url": "/",
  "display": "standalone",
  "theme_color": "#6366f1",
  "background_color": "#ffffff",
  "icons": [
    { "src": "/icon-192.png", "sizes": "192x192", "type": "image/png" },
    { "src": "/icon-512.png", "sizes": "512x512", "type": "image/png" },
    { "src": "/icon-maskable.png", "sizes": "512x512", "type": "image/png", "purpose": "maskable" }
  ]
}
```

## Caching Strategies

| Strategy | Use For | Code |
|----------|---------|------|
| Cache First | CSS, JS, fonts, images | `caches.match(req) \|\| fetch(req)` |
| Network First | API calls, dynamic | `fetch(req).catch(() => caches.match(req))` |
| Stale While Revalidate | Avatars, configs | Return cached, fetch + update cache |
| Network Only | Auth, payments | `fetch(req)` |
| Cache Only | Pre-cached shell | `caches.match(req)` |

## Service Worker Lifecycle

```
Install → Waiting → Activate → Fetch/Push/Sync
```

```javascript
// Install: pre-cache
self.addEventListener("install", (e) => {
  e.waitUntil(caches.open("v1").then(c => c.addAll(["/", "/app.js"])));
});

// Activate: clean old caches
self.addEventListener("activate", (e) => {
  e.waitUntil(caches.keys().then(keys =>
    Promise.all(keys.filter(k => k !== "v1").map(k => caches.delete(k)))
  ));
  self.clients.claim();
});

// Fetch: route to strategy
self.addEventListener("fetch", (e) => {
  e.respondWith(/* caching strategy */);
});

// Skip waiting
self.addEventListener("message", (e) => {
  if (e.data?.type === "SKIP_WAITING") self.skipWaiting();
});
```

## Workbox with Vite

```bash
npm install vite-plugin-pwa
```

```typescript
// vite.config.ts
import { VitePWA } from "vite-plugin-pwa";

VitePWA({
  registerType: "prompt",
  workbox: {
    globPatterns: ["**/*.{js,css,html,ico,png,svg,woff2}"],
    runtimeCaching: [
      { urlPattern: /^https:\/\/api/, handler: "NetworkFirst",
        options: { cacheName: "api", expiration: { maxEntries: 100 } } },
      { urlPattern: /\.(?:png|jpg|svg)$/, handler: "CacheFirst",
        options: { cacheName: "images", expiration: { maxEntries: 200, maxAgeSeconds: 30*24*60*60 } } },
    ],
  },
})
```

## Install Prompt (React)

```tsx
const [prompt, setPrompt] = useState<any>(null);

useEffect(() => {
  window.addEventListener("beforeinstallprompt", (e) => {
    e.preventDefault();
    setPrompt(e);
  });
}, []);

const install = async () => {
  if (!prompt) return;
  await prompt.prompt();
  const { outcome } = await prompt.userChoice;
  setPrompt(null);
};
```

## Push Notifications

```typescript
// Subscribe
const reg = await navigator.serviceWorker.ready;
const sub = await reg.pushManager.subscribe({
  userVisibleOnly: true,
  applicationServerKey: vapidKey,
});

// In SW: handle push
self.addEventListener("push", (e) => {
  const data = e.data?.json();
  e.waitUntil(self.registration.showNotification(data.title, { body: data.body }));
});

// Handle click
self.addEventListener("notificationclick", (e) => {
  e.notification.close();
  e.waitUntil(clients.openWindow(e.notification.data?.url || "/"));
});
```

## Background Sync

```typescript
// Register
const reg = await navigator.serviceWorker.ready;
await reg.sync.register("sync-data");

// In SW: handle sync
self.addEventListener("sync", (e) => {
  if (e.tag === "sync-data") e.waitUntil(doSync());
});
```

## HTML Head Tags

```html
<link rel="manifest" href="/manifest.json" />
<meta name="theme-color" content="#6366f1" />
<meta name="apple-mobile-web-app-capable" content="yes" />
<meta name="apple-mobile-web-app-status-bar-style" content="black-translucent" />
<link rel="apple-touch-icon" href="/icons/apple-touch-icon.png" />
```

## Detect Installed

```typescript
const isInstalled = window.matchMedia("(display-mode: standalone)").matches;
```

```css
@media (display-mode: standalone) { .install-btn { display: none; } }
```

## iOS Limitations

- No `beforeinstallprompt` — show manual "Add to Home Screen" instructions
- No push notifications (iOS 16.4+ supports, earlier doesn't)
- No background sync
- Web app data can be cleared if device runs low on storage
- 50MB total storage limit per web app
