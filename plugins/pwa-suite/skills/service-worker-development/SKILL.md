---
name: service-worker-development
description: >
  Service worker development — caching strategies, offline support,
  background sync, push notifications, Workbox integration, and
  the service worker lifecycle.
  Triggers: "service worker", "offline support", "cache strategy",
  "workbox", "push notification", "background sync", "precache".
  NOT for: Browser extension service workers (use chrome-extension).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Service Workers

## Manual Service Worker

### Registration

```typescript
// src/sw-register.ts — call from main entry point
export async function registerServiceWorker() {
  if (!("serviceWorker" in navigator)) return;

  // Register AFTER page load to not compete with critical resources
  window.addEventListener("load", async () => {
    try {
      const registration = await navigator.serviceWorker.register("/sw.js", {
        scope: "/",
      });

      // Handle updates
      registration.addEventListener("updatefound", () => {
        const newWorker = registration.installing;
        if (!newWorker) return;

        newWorker.addEventListener("statechange", () => {
          if (newWorker.state === "installed" && navigator.serviceWorker.controller) {
            // New version available — prompt user
            showUpdatePrompt(registration);
          }
        });
      });
    } catch (error) {
      console.error("SW registration failed:", error);
    }
  });
}

function showUpdatePrompt(registration: ServiceWorkerRegistration) {
  if (confirm("New version available! Reload?")) {
    registration.waiting?.postMessage({ type: "SKIP_WAITING" });
    // Reload when the new SW takes control
    navigator.serviceWorker.addEventListener("controllerchange", () => {
      window.location.reload();
    });
  }
}
```

### Service Worker File

```typescript
// public/sw.js
const CACHE_NAME = "app-cache-v1";
const STATIC_ASSETS = [
  "/",
  "/index.html",
  "/styles.css",
  "/app.js",
  "/offline.html",
];

// Install: pre-cache static assets
self.addEventListener("install", (event: ExtendableEvent) => {
  event.waitUntil(
    caches.open(CACHE_NAME).then((cache) => cache.addAll(STATIC_ASSETS))
  );
});

// Activate: clean old caches
self.addEventListener("activate", (event: ExtendableEvent) => {
  event.waitUntil(
    caches.keys().then((keys) =>
      Promise.all(
        keys
          .filter((key) => key !== CACHE_NAME)
          .map((key) => caches.delete(key))
      )
    )
  );
  // Take control of all pages immediately
  (self as any).clients.claim();
});

// Skip waiting when prompted
self.addEventListener("message", (event) => {
  if (event.data?.type === "SKIP_WAITING") {
    (self as any).skipWaiting();
  }
});
```

### Caching Strategies

```typescript
// Cache First (static assets)
async function cacheFirst(request: Request): Promise<Response> {
  const cached = await caches.match(request);
  if (cached) return cached;

  const response = await fetch(request);
  if (response.ok) {
    const cache = await caches.open(CACHE_NAME);
    cache.put(request, response.clone());
  }
  return response;
}

// Network First (API calls)
async function networkFirst(request: Request): Promise<Response> {
  try {
    const response = await fetch(request);
    if (response.ok) {
      const cache = await caches.open("api-cache-v1");
      cache.put(request, response.clone());
    }
    return response;
  } catch {
    const cached = await caches.match(request);
    if (cached) return cached;
    return new Response(JSON.stringify({ error: "Offline" }), {
      status: 503,
      headers: { "Content-Type": "application/json" },
    });
  }
}

// Stale While Revalidate (semi-dynamic)
async function staleWhileRevalidate(request: Request): Promise<Response> {
  const cache = await caches.open(CACHE_NAME);
  const cached = await cache.match(request);

  const networkPromise = fetch(request).then((response) => {
    if (response.ok) cache.put(request, response.clone());
    return response;
  });

  return cached ?? networkPromise;
}

// Fetch handler — route to strategies
self.addEventListener("fetch", (event: FetchEvent) => {
  const { request } = event;
  const url = new URL(request.url);

  // Navigation requests — network first with offline fallback
  if (request.mode === "navigate") {
    event.respondWith(
      networkFirst(request).catch(() => caches.match("/offline.html")!)
    );
    return;
  }

  // API calls — network first
  if (url.pathname.startsWith("/api/")) {
    event.respondWith(networkFirst(request));
    return;
  }

  // Static assets — cache first
  if (
    request.destination === "style" ||
    request.destination === "script" ||
    request.destination === "image" ||
    request.destination === "font"
  ) {
    event.respondWith(cacheFirst(request));
    return;
  }

  // Everything else — stale while revalidate
  event.respondWith(staleWhileRevalidate(request));
});
```

## Workbox (Recommended for Production)

```bash
npm install workbox-webpack-plugin  # or workbox-build
# or for Vite:
npm install vite-plugin-pwa
```

### Vite PWA Plugin

```typescript
// vite.config.ts
import { VitePWA } from "vite-plugin-pwa";

export default defineConfig({
  plugins: [
    react(),
    VitePWA({
      registerType: "prompt",  // or "autoUpdate"
      includeAssets: ["favicon.ico", "robots.txt", "apple-touch-icon.png"],

      manifest: {
        name: "My App",
        short_name: "MyApp",
        description: "My Progressive Web App",
        theme_color: "#6366f1",
        background_color: "#ffffff",
        display: "standalone",
        start_url: "/",
        icons: [
          { src: "/icons/icon-192.png", sizes: "192x192", type: "image/png" },
          { src: "/icons/icon-512.png", sizes: "512x512", type: "image/png" },
          { src: "/icons/icon-512.png", sizes: "512x512", type: "image/png", purpose: "maskable" },
        ],
      },

      workbox: {
        // Pre-cache app shell
        globPatterns: ["**/*.{js,css,html,ico,png,svg,woff2}"],

        // Runtime caching
        runtimeCaching: [
          {
            // API calls — network first
            urlPattern: /^https:\/\/api\.example\.com\/.*/i,
            handler: "NetworkFirst",
            options: {
              cacheName: "api-cache",
              expiration: { maxEntries: 100, maxAgeSeconds: 60 * 60 },
              networkTimeoutSeconds: 5,
            },
          },
          {
            // Images — cache first
            urlPattern: /\.(?:png|jpg|jpeg|svg|gif|webp)$/,
            handler: "CacheFirst",
            options: {
              cacheName: "images-cache",
              expiration: { maxEntries: 200, maxAgeSeconds: 30 * 24 * 60 * 60 },
            },
          },
          {
            // Google Fonts — stale while revalidate
            urlPattern: /^https:\/\/fonts\.googleapis\.com\/.*/i,
            handler: "StaleWhileRevalidate",
            options: {
              cacheName: "google-fonts",
              expiration: { maxEntries: 20 },
            },
          },
        ],

        // Offline fallback
        navigateFallback: "/index.html",
        navigateFallbackDenylist: [/^\/api/],
      },
    }),
  ],
});
```

### Update Prompt (React)

```tsx
// src/components/PWAUpdatePrompt.tsx
import { useRegisterSW } from "virtual:pwa-register/react";

function PWAUpdatePrompt() {
  const {
    needRefresh: [needRefresh, setNeedRefresh],
    updateServiceWorker,
  } = useRegisterSW();

  if (!needRefresh) return null;

  return (
    <div className="fixed bottom-4 right-4 bg-white shadow-lg rounded-lg p-4 z-50 border">
      <p className="text-sm font-medium">New version available!</p>
      <div className="mt-2 flex gap-2">
        <button
          onClick={() => updateServiceWorker(true)}
          className="px-3 py-1 bg-blue-600 text-white text-sm rounded"
        >
          Update
        </button>
        <button
          onClick={() => setNeedRefresh(false)}
          className="px-3 py-1 text-gray-600 text-sm"
        >
          Later
        </button>
      </div>
    </div>
  );
}
```

## Push Notifications

```typescript
// Request permission and subscribe
async function subscribePush(): Promise<PushSubscription | null> {
  const permission = await Notification.requestPermission();
  if (permission !== "granted") return null;

  const registration = await navigator.serviceWorker.ready;
  const subscription = await registration.pushManager.subscribe({
    userVisibleOnly: true,
    applicationServerKey: urlBase64ToUint8Array(VAPID_PUBLIC_KEY),
  });

  // Send subscription to your server
  await fetch("/api/push/subscribe", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(subscription),
  });

  return subscription;
}

function urlBase64ToUint8Array(base64String: string): Uint8Array {
  const padding = "=".repeat((4 - (base64String.length % 4)) % 4);
  const base64 = (base64String + padding).replace(/-/g, "+").replace(/_/g, "/");
  const rawData = atob(base64);
  return Uint8Array.from([...rawData].map((c) => c.charCodeAt(0)));
}

// In service worker: handle push events
self.addEventListener("push", (event: PushEvent) => {
  const data = event.data?.json() ?? {};

  event.waitUntil(
    (self as any).registration.showNotification(data.title ?? "Notification", {
      body: data.body,
      icon: "/icons/icon-192.png",
      badge: "/icons/badge-72.png",
      data: { url: data.url ?? "/" },
      actions: [
        { action: "open", title: "Open" },
        { action: "dismiss", title: "Dismiss" },
      ],
    })
  );
});

// Handle notification click
self.addEventListener("notificationclick", (event: NotificationEvent) => {
  event.notification.close();

  if (event.action === "dismiss") return;

  const url = event.notification.data?.url ?? "/";
  event.waitUntil(
    (self as any).clients.matchAll({ type: "window" }).then((clients: any[]) => {
      // Focus existing tab if open
      const existing = clients.find((c: any) => c.url === url);
      if (existing) return existing.focus();
      // Otherwise open new tab
      return (self as any).clients.openWindow(url);
    })
  );
});
```

## Background Sync

```typescript
// Register sync when offline action happens
async function saveForLater(data: any) {
  // Store in IndexedDB
  const db = await openDB("sync-queue", 1, {
    upgrade(db) { db.createObjectStore("pending", { autoIncrement: true }); },
  });
  await db.add("pending", { data, timestamp: Date.now() });

  // Request background sync
  const registration = await navigator.serviceWorker.ready;
  await registration.sync.register("sync-pending-data");
}

// In service worker: handle sync events
self.addEventListener("sync", (event: SyncEvent) => {
  if (event.tag === "sync-pending-data") {
    event.waitUntil(syncPendingData());
  }
});

async function syncPendingData() {
  const db = await openDB("sync-queue", 1);
  const tx = db.transaction("pending", "readwrite");
  const store = tx.objectStore("pending");
  const items = await store.getAll();

  for (const item of items) {
    try {
      await fetch("/api/data", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(item.data),
      });
      await store.delete(item.id);
    } catch {
      // Will retry on next sync event
      break;
    }
  }
}
```

## Gotchas

1. **Service worker scope is based on location** — A SW at `/js/sw.js` can only control pages under `/js/`. Place it at the root (`/sw.js`) or set `Service-Worker-Allowed` header to extend scope.

2. **No ES module imports in service workers** — Use `importScripts()` for dependencies or bundle your SW with Workbox/Vite. Some browsers support `type: "module"` in SW registration but it's not universal.

3. **Cache API stores Request/Response pairs** — The cache key is the full URL including query parameters. `cache.match("/api?page=1")` won't match `cache.match("/api?page=2")`. Use `ignoreSearch: true` if you want to ignore query params.

4. **`skipWaiting()` can break in-flight requests** — If a user has multiple tabs open, skipping waiting swaps the SW for all tabs simultaneously. Pages mid-request may break if the new SW handles requests differently. Use with a reload prompt.

5. **Push notifications require user-visible notification** — The `userVisibleOnly: true` option is mandatory. You cannot receive push events silently. Every push MUST show a notification to the user.

6. **Background Sync is not periodic** — `sync.register()` fires ONCE when connectivity returns. For periodic tasks, use the experimental Periodic Background Sync API (`periodicSync.register()`) with a minimum interval of 12 hours.
