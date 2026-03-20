---
name: pwa-offline-sync
description: >
  Offline data synchronization patterns for Progressive Web Apps.
  Use when implementing offline-first storage, background sync,
  IndexedDB patterns, or conflict resolution in PWAs.
  Triggers: "offline sync", "background sync", "IndexedDB", "offline first",
  "sync queue", "conflict resolution", "offline storage", "PWA data sync".
  NOT for: service worker caching (see service-worker-development), PWA setup (see pwa-setup), server-side sync.
version: 1.0.0
allowed-tools: Read, Grep, Glob, Bash
---

# PWA Offline Sync

## IndexedDB Storage Layer

```typescript
// lib/offline-store.ts — Type-safe IndexedDB wrapper
interface DBSchema {
  todos: { id: string; title: string; completed: boolean; updatedAt: number; synced: boolean };
  syncQueue: { id: string; action: 'create' | 'update' | 'delete'; data: unknown; createdAt: number; retries: number };
  metadata: { key: string; value: unknown };
}

class OfflineStore<T extends Record<string, unknown>> {
  private db: IDBDatabase | null = null;

  constructor(
    private dbName: string,
    private version: number,
    private stores: { name: string; keyPath: string; indexes?: { name: string; keyPath: string; unique?: boolean }[] }[],
  ) {}

  async open(): Promise<void> {
    return new Promise((resolve, reject) => {
      const request = indexedDB.open(this.dbName, this.version);

      request.onupgradeneeded = (event) => {
        const db = (event.target as IDBOpenDBRequest).result;
        for (const store of this.stores) {
          if (!db.objectStoreNames.contains(store.name)) {
            const objectStore = db.createObjectStore(store.name, { keyPath: store.keyPath });
            for (const index of store.indexes ?? []) {
              objectStore.createIndex(index.name, index.keyPath, { unique: index.unique });
            }
          }
        }
      };

      request.onsuccess = () => { this.db = request.result; resolve(); };
      request.onerror = () => reject(request.error);
    });
  }

  async put<K extends keyof T>(storeName: K, value: T[K]): Promise<void> {
    return this.transaction(storeName as string, 'readwrite', (store) => {
      store.put(value);
    });
  }

  async get<K extends keyof T>(storeName: K, key: IDBValidKey): Promise<T[K] | undefined> {
    return this.transaction(storeName as string, 'readonly', (store) => {
      return store.get(key);
    });
  }

  async getAll<K extends keyof T>(storeName: K): Promise<T[K][]> {
    return this.transaction(storeName as string, 'readonly', (store) => {
      return store.getAll();
    });
  }

  async delete<K extends keyof T>(storeName: K, key: IDBValidKey): Promise<void> {
    return this.transaction(storeName as string, 'readwrite', (store) => {
      store.delete(key);
    });
  }

  private transaction<R>(
    storeName: string,
    mode: IDBTransactionMode,
    callback: (store: IDBObjectStore) => IDBRequest<R> | void,
  ): Promise<R> {
    return new Promise((resolve, reject) => {
      if (!this.db) throw new Error('Database not open');
      const tx = this.db.transaction(storeName, mode);
      const store = tx.objectStore(storeName);
      const request = callback(store);

      if (request) {
        request.onsuccess = () => resolve(request.result);
        request.onerror = () => reject(request.error);
      } else {
        tx.oncomplete = () => resolve(undefined as R);
        tx.onerror = () => reject(tx.error);
      }
    });
  }
}

// Initialize
const store = new OfflineStore<DBSchema>('myapp', 1, [
  { name: 'todos', keyPath: 'id', indexes: [{ name: 'synced', keyPath: 'synced' }] },
  { name: 'syncQueue', keyPath: 'id', indexes: [{ name: 'createdAt', keyPath: 'createdAt' }] },
  { name: 'metadata', keyPath: 'key' },
]);
```

## Sync Queue with Background Sync

```typescript
// lib/sync-manager.ts — Queue mutations and sync when online

interface SyncAction {
  id: string;
  action: 'create' | 'update' | 'delete';
  endpoint: string;
  method: 'POST' | 'PUT' | 'PATCH' | 'DELETE';
  body?: unknown;
  createdAt: number;
  retries: number;
  maxRetries: number;
}

class SyncManager {
  private store: OfflineStore<DBSchema>;
  private isSyncing = false;

  constructor(store: OfflineStore<DBSchema>) {
    this.store = store;

    // Listen for online events
    window.addEventListener('online', () => this.processQueue());

    // Register for Background Sync API
    this.registerBackgroundSync();
  }

  async queueAction(action: Omit<SyncAction, 'id' | 'createdAt' | 'retries'>): Promise<void> {
    const syncAction: SyncAction = {
      ...action,
      id: crypto.randomUUID(),
      createdAt: Date.now(),
      retries: 0,
    };

    await this.store.put('syncQueue', syncAction as any);

    // Try immediate sync if online
    if (navigator.onLine) {
      await this.processQueue();
    } else {
      // Request background sync
      await this.requestBackgroundSync();
    }
  }

  async processQueue(): Promise<{ synced: number; failed: number }> {
    if (this.isSyncing) return { synced: 0, failed: 0 };
    this.isSyncing = true;

    let synced = 0, failed = 0;

    try {
      const queue = await this.store.getAll('syncQueue');
      const sorted = (queue as SyncAction[]).sort((a, b) => a.createdAt - b.createdAt);

      for (const action of sorted) {
        try {
          const response = await fetch(action.endpoint, {
            method: action.method,
            headers: { 'Content-Type': 'application/json' },
            body: action.body ? JSON.stringify(action.body) : undefined,
          });

          if (response.ok) {
            await this.store.delete('syncQueue', action.id);
            synced++;
          } else if (response.status >= 400 && response.status < 500) {
            // Client error — don't retry
            await this.store.delete('syncQueue', action.id);
            failed++;
            console.error(`Sync action ${action.id} permanently failed:`, response.status);
          } else {
            // Server error — retry later
            action.retries++;
            if (action.retries >= action.maxRetries) {
              await this.store.delete('syncQueue', action.id);
              failed++;
            } else {
              await this.store.put('syncQueue', action as any);
            }
          }
        } catch (error) {
          // Network error — stop processing, will retry later
          console.warn('Sync failed, will retry:', error);
          break;
        }
      }
    } finally {
      this.isSyncing = false;
    }

    return { synced, failed };
  }

  private async registerBackgroundSync(): Promise<void> {
    if (!('serviceWorker' in navigator) || !('SyncManager' in window)) return;
    const registration = await navigator.serviceWorker.ready;
    try {
      await (registration as any).sync.register('sync-queue');
    } catch (error) {
      console.warn('Background Sync not supported:', error);
    }
  }

  private async requestBackgroundSync(): Promise<void> {
    try {
      const registration = await navigator.serviceWorker.ready;
      await (registration as any).sync.register('sync-queue');
    } catch {
      // Fallback: retry on next online event
    }
  }
}

// Service worker side:
// self.addEventListener('sync', (event) => {
//   if (event.tag === 'sync-queue') {
//     event.waitUntil(processOfflineQueue());
//   }
// });
```

## Conflict Resolution

```typescript
// lib/conflict-resolver.ts — Handle conflicts when syncing offline changes

interface VersionedRecord {
  id: string;
  version: number;
  updatedAt: number;
  data: Record<string, unknown>;
}

type ConflictStrategy = 'client-wins' | 'server-wins' | 'last-write-wins' | 'merge';

interface ConflictResult {
  resolved: VersionedRecord;
  strategy: ConflictStrategy;
  hadConflict: boolean;
}

function resolveConflict(
  clientVersion: VersionedRecord,
  serverVersion: VersionedRecord,
  strategy: ConflictStrategy = 'last-write-wins',
): ConflictResult {
  // No conflict if versions match
  if (clientVersion.version === serverVersion.version) {
    return { resolved: clientVersion, strategy, hadConflict: false };
  }

  switch (strategy) {
    case 'client-wins':
      return {
        resolved: { ...clientVersion, version: serverVersion.version + 1 },
        strategy,
        hadConflict: true,
      };

    case 'server-wins':
      return { resolved: serverVersion, strategy, hadConflict: true };

    case 'last-write-wins':
      const winner = clientVersion.updatedAt > serverVersion.updatedAt
        ? clientVersion
        : serverVersion;
      return {
        resolved: { ...winner, version: Math.max(clientVersion.version, serverVersion.version) + 1 },
        strategy,
        hadConflict: true,
      };

    case 'merge':
      // Field-level merge: use the most recently updated value for each field
      const merged: Record<string, unknown> = {};
      const allKeys = new Set([
        ...Object.keys(clientVersion.data),
        ...Object.keys(serverVersion.data),
      ]);

      for (const key of allKeys) {
        // If only one side has the field, use that
        if (!(key in serverVersion.data)) {
          merged[key] = clientVersion.data[key];
        } else if (!(key in clientVersion.data)) {
          merged[key] = serverVersion.data[key];
        } else {
          // Both sides have the field — use client (offline changes take priority)
          merged[key] = clientVersion.data[key];
        }
      }

      return {
        resolved: {
          id: clientVersion.id,
          version: Math.max(clientVersion.version, serverVersion.version) + 1,
          updatedAt: Date.now(),
          data: merged,
        },
        strategy,
        hadConflict: true,
      };

    default:
      throw new Error(`Unknown conflict strategy: ${strategy}`);
  }
}
```

## Online Status UI

```tsx
// components/OnlineStatus.tsx
import { useState, useEffect, useSyncExternalStore } from 'react';

// Reactive online status hook
function useOnlineStatus(): boolean {
  return useSyncExternalStore(
    (callback) => {
      window.addEventListener('online', callback);
      window.addEventListener('offline', callback);
      return () => {
        window.removeEventListener('online', callback);
        window.removeEventListener('offline', callback);
      };
    },
    () => navigator.onLine,
    () => true, // SSR assumes online
  );
}

// Sync status hook
function useSyncStatus(syncManager: SyncManager) {
  const [pendingCount, setPendingCount] = useState(0);
  const [lastSyncedAt, setLastSyncedAt] = useState<Date | null>(null);

  useEffect(() => {
    const interval = setInterval(async () => {
      const queue = await store.getAll('syncQueue');
      setPendingCount(queue.length);
    }, 2000);
    return () => clearInterval(interval);
  }, []);

  return { pendingCount, lastSyncedAt };
}

// UI component
function SyncIndicator() {
  const isOnline = useOnlineStatus();
  const { pendingCount } = useSyncStatus(syncManager);

  return (
    <div role="status" aria-live="polite">
      {!isOnline && (
        <span className="offline-badge">
          Offline — changes saved locally
        </span>
      )}
      {isOnline && pendingCount > 0 && (
        <span className="syncing-badge">
          Syncing {pendingCount} change{pendingCount !== 1 ? 's' : ''}...
        </span>
      )}
    </div>
  );
}
```

## Gotchas

1. **IndexedDB storage limits** -- Browsers limit storage per origin (typically 50-80% of available disk). When storage is full, IndexedDB operations silently fail or throw QuotaExceededError. Implement storage estimation (`navigator.storage.estimate()`) and cleanup of old synced records. Never assume unlimited offline storage.

2. **Background Sync is not guaranteed** -- The Background Sync API requires a service worker and browser support. Safari has limited support. Chrome may delay sync events by minutes or hours on battery saver. Always implement a fallback: retry on `online` event + periodic polling. Don't rely solely on Background Sync.

3. **Conflict resolution needs user input for ambiguous cases** -- Automatic merge works for independent field changes, but if both client and server modified the same field, no algorithm can determine the "right" value. Surface merge conflicts to the user with a diff view. Automatic resolution of ambiguous conflicts silently loses data.

4. **Transaction ordering matters** -- Sync queue must be processed in order (FIFO). Processing "delete item X" before "create item X" fails. Out-of-order sync also causes issues with dependent records (create parent before child). Sort by `createdAt` and process sequentially, not in parallel.

5. **navigator.onLine is unreliable** -- `navigator.onLine` only detects whether a network interface is connected, not whether the internet is reachable. A connected WiFi with no internet shows `onLine: true`. Verify actual connectivity with a lightweight fetch to your API (`/api/ping`) before processing the sync queue.

6. **IndexedDB in private browsing** -- Some browsers (older Safari, Firefox) limit or disable IndexedDB in private/incognito mode. Always wrap IndexedDB operations in try-catch and fall back to in-memory storage. Check `indexedDB` existence before assuming it works.
