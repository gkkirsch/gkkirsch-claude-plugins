# React Native Patterns Reference

Production patterns for React Native mobile development. This reference covers offline-first architecture, deep linking, push notifications, app state management, background tasks, and platform-specific code.

---

## 1. Offline-First Architecture

### Strategy Selection

| Strategy | Use When | Complexity |
|----------|----------|------------|
| Cache-first | Read-heavy, tolerance for stale data | Low |
| Network-first | Write-heavy, need freshness | Medium |
| Stale-while-revalidate | Balance of freshness and speed | Medium |
| Offline queue | Must support offline writes | High |
| CRDT-based sync | Multi-device conflict resolution | Very High |

### MMKV for Fast Key-Value Storage

```typescript
// storage/mmkv.ts
import { MMKV } from 'react-native-mmkv';

export const storage = new MMKV({
  id: 'app-storage',
  encryptionKey: 'your-encryption-key', // optional
});

// Typed storage wrapper
export const AppStorage = {
  // Auth tokens
  getAccessToken: () => storage.getString('auth.accessToken'),
  setAccessToken: (token: string) => storage.set('auth.accessToken', token),
  removeAccessToken: () => storage.delete('auth.accessToken'),

  // User preferences
  getTheme: () => (storage.getString('prefs.theme') as 'light' | 'dark') ?? 'light',
  setTheme: (theme: 'light' | 'dark') => storage.set('prefs.theme', theme),

  // Onboarding state
  hasCompletedOnboarding: () => storage.getBoolean('onboarding.completed') ?? false,
  setOnboardingCompleted: () => storage.set('onboarding.completed', true),

  // Cache with TTL
  setWithTTL: (key: string, value: string, ttlMs: number) => {
    storage.set(key, value);
    storage.set(`${key}.__expires`, Date.now() + ttlMs);
  },
  getWithTTL: (key: string): string | undefined => {
    const expires = storage.getNumber(`${key}.__expires`);
    if (expires && Date.now() > expires) {
      storage.delete(key);
      storage.delete(`${key}.__expires`);
      return undefined;
    }
    return storage.getString(key);
  },

  // Clear all data (for logout)
  clearAll: () => storage.clearAll(),
};
```

### Zustand with MMKV Persistence

```typescript
// stores/auth-store.ts
import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';
import { MMKV } from 'react-native-mmkv';

const mmkv = new MMKV({ id: 'auth-store' });

const mmkvStorage = {
  getItem: (name: string) => mmkv.getString(name) ?? null,
  setItem: (name: string, value: string) => mmkv.set(name, value),
  removeItem: (name: string) => mmkv.delete(name),
};

interface AuthState {
  user: User | null;
  accessToken: string | null;
  refreshToken: string | null;
  isAuthenticated: boolean;
  login: (credentials: LoginCredentials) => Promise<void>;
  logout: () => void;
  refreshAuth: () => Promise<void>;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      user: null,
      accessToken: null,
      refreshToken: null,
      isAuthenticated: false,

      login: async (credentials) => {
        const response = await authApi.login(credentials);
        set({
          user: response.user,
          accessToken: response.accessToken,
          refreshToken: response.refreshToken,
          isAuthenticated: true,
        });
      },

      logout: () => {
        set({
          user: null,
          accessToken: null,
          refreshToken: null,
          isAuthenticated: false,
        });
      },

      refreshAuth: async () => {
        const { refreshToken } = get();
        if (!refreshToken) {
          get().logout();
          return;
        }
        try {
          const response = await authApi.refresh(refreshToken);
          set({
            accessToken: response.accessToken,
            refreshToken: response.refreshToken,
          });
        } catch {
          get().logout();
        }
      },
    }),
    {
      name: 'auth-storage',
      storage: createJSONStorage(() => mmkvStorage),
      partialize: (state) => ({
        user: state.user,
        accessToken: state.accessToken,
        refreshToken: state.refreshToken,
        isAuthenticated: state.isAuthenticated,
      }),
    }
  )
);
```

### WatermelonDB for Complex Offline Data

```typescript
// database/schema.ts
import { appSchema, tableSchema } from '@nozbe/watermelondb';

export const schema = appSchema({
  version: 1,
  tables: [
    tableSchema({
      name: 'tasks',
      columns: [
        { name: 'title', type: 'string' },
        { name: 'body', type: 'string' },
        { name: 'is_completed', type: 'boolean' },
        { name: 'project_id', type: 'string', isIndexed: true },
        { name: 'assigned_to', type: 'string', isIndexed: true },
        { name: 'due_at', type: 'number', isOptional: true },
        { name: 'created_at', type: 'number' },
        { name: 'updated_at', type: 'number' },
      ],
    }),
    tableSchema({
      name: 'projects',
      columns: [
        { name: 'name', type: 'string' },
        { name: 'color', type: 'string' },
        { name: 'is_archived', type: 'boolean' },
        { name: 'created_at', type: 'number' },
        { name: 'updated_at', type: 'number' },
      ],
    }),
  ],
});

// database/models/Task.ts
import { Model, Q } from '@nozbe/watermelondb';
import { field, text, relation, date, readonly, writer } from '@nozbe/watermelondb/decorators';

export class Task extends Model {
  static table = 'tasks';
  static associations = {
    projects: { type: 'belongs_to' as const, key: 'project_id' },
  };

  @text('title') title!: string;
  @text('body') body!: string;
  @field('is_completed') isCompleted!: boolean;
  @relation('projects', 'project_id') project!: any;
  @field('assigned_to') assignedTo!: string;
  @date('due_at') dueAt!: Date | null;
  @readonly @date('created_at') createdAt!: Date;
  @readonly @date('updated_at') updatedAt!: Date;

  @writer async markAsCompleted() {
    await this.update((task) => {
      task.isCompleted = true;
    });
  }

  @writer async updateTitle(newTitle: string) {
    await this.update((task) => {
      task.title = newTitle;
    });
  }
}

// database/sync.ts — Sync with backend
import { synchronize } from '@nozbe/watermelondb/sync';
import { database } from './index';

export async function syncDatabase() {
  await synchronize({
    database,
    pullChanges: async ({ lastPulledAt, schemaVersion, migration }) => {
      const response = await fetch(
        `${API_URL}/sync?last_pulled_at=${lastPulledAt}&schema_version=${schemaVersion}`
      );
      if (!response.ok) throw new Error('Sync pull failed');

      const { changes, timestamp } = await response.json();
      return { changes, timestamp };
    },
    pushChanges: async ({ changes, lastPulledAt }) => {
      const response = await fetch(`${API_URL}/sync?last_pulled_at=${lastPulledAt}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(changes),
      });
      if (!response.ok) throw new Error('Sync push failed');
    },
    migrationsEnabledAtVersion: 1,
  });
}
```

### Offline Queue Pattern

```typescript
// services/offline-queue.ts
import { MMKV } from 'react-native-mmkv';
import NetInfo from '@react-native-community/netinfo';

interface QueuedAction {
  id: string;
  type: string;
  payload: unknown;
  timestamp: number;
  retryCount: number;
}

const queueStorage = new MMKV({ id: 'offline-queue' });

export class OfflineQueue {
  private isProcessing = false;

  constructor() {
    // Process queue when coming back online
    NetInfo.addEventListener((state) => {
      if (state.isConnected && !this.isProcessing) {
        this.processQueue();
      }
    });
  }

  enqueue(action: Omit<QueuedAction, 'id' | 'timestamp' | 'retryCount'>) {
    const queue = this.getQueue();
    const queuedAction: QueuedAction = {
      ...action,
      id: `${Date.now()}-${Math.random().toString(36).slice(2)}`,
      timestamp: Date.now(),
      retryCount: 0,
    };
    queue.push(queuedAction);
    queueStorage.set('queue', JSON.stringify(queue));
  }

  private getQueue(): QueuedAction[] {
    const raw = queueStorage.getString('queue');
    return raw ? JSON.parse(raw) : [];
  }

  async processQueue() {
    if (this.isProcessing) return;
    this.isProcessing = true;

    try {
      const queue = this.getQueue();
      const remaining: QueuedAction[] = [];

      for (const action of queue) {
        try {
          await this.executeAction(action);
        } catch (error) {
          if (action.retryCount < 3) {
            remaining.push({ ...action, retryCount: action.retryCount + 1 });
          }
          // Drop after 3 retries — log to error reporting
        }
      }

      queueStorage.set('queue', JSON.stringify(remaining));
    } finally {
      this.isProcessing = false;
    }
  }

  private async executeAction(action: QueuedAction): Promise<void> {
    // Route to appropriate API handler based on action type
    switch (action.type) {
      case 'CREATE_TASK':
        await api.tasks.create(action.payload as CreateTaskPayload);
        break;
      case 'UPDATE_TASK':
        await api.tasks.update(action.payload as UpdateTaskPayload);
        break;
      case 'DELETE_TASK':
        await api.tasks.delete(action.payload as DeleteTaskPayload);
        break;
      default:
        throw new Error(`Unknown action type: ${action.type}`);
    }
  }
}

export const offlineQueue = new OfflineQueue();
```

### Network Status Hook

```typescript
// hooks/useNetworkStatus.ts
import { useEffect, useState } from 'react';
import NetInfo, { NetInfoState } from '@react-native-community/netinfo';

export function useNetworkStatus() {
  const [isConnected, setIsConnected] = useState(true);
  const [connectionType, setConnectionType] = useState<string>('unknown');

  useEffect(() => {
    const unsubscribe = NetInfo.addEventListener((state: NetInfoState) => {
      setIsConnected(state.isConnected ?? false);
      setConnectionType(state.type);
    });
    return () => unsubscribe();
  }, []);

  return { isConnected, connectionType };
}

// components/OfflineBanner.tsx
import { useNetworkStatus } from '../hooks/useNetworkStatus';

export function OfflineBanner() {
  const { isConnected } = useNetworkStatus();

  if (isConnected) return null;

  return (
    <View style={styles.banner}>
      <Icon name="wifi-off" size={16} color="#fff" />
      <Text style={styles.text}>You are offline. Changes will sync when reconnected.</Text>
    </View>
  );
}
```

---

## 2. Deep Linking

### React Navigation Deep Linking Configuration

```typescript
// navigation/linking.ts
import { LinkingOptions } from '@react-navigation/native';
import { RootStackParamList } from './types';

export const linking: LinkingOptions<RootStackParamList> = {
  prefixes: [
    'myapp://',
    'https://myapp.com',
    'https://*.myapp.com',
  ],
  config: {
    screens: {
      // Auth screens
      Login: 'login',
      Register: 'register',
      ForgotPassword: 'forgot-password',
      ResetPassword: 'reset-password/:token',

      // Main tab navigator
      MainTabs: {
        screens: {
          Home: {
            path: 'home',
            screens: {
              Feed: '',
              PostDetail: 'post/:postId',
              UserProfile: 'user/:userId',
            },
          },
          Search: 'search',
          Notifications: 'notifications',
          Profile: {
            path: 'profile',
            screens: {
              MyProfile: '',
              Settings: 'settings',
              EditProfile: 'edit',
            },
          },
        },
      },

      // Modal screens
      CreatePost: 'create-post',
      ImageViewer: 'image/:imageId',
    },
  },

  // Custom URL parsing
  getStateFromPath: (path, config) => {
    // Handle special URLs
    if (path.startsWith('invite/')) {
      return {
        routes: [{ name: 'InviteAccept', params: { code: path.replace('invite/', '') } }],
      };
    }
    // Fall back to default parsing
    return undefined;
  },
};
```

### Universal Links (iOS) Setup

```xml
<!-- ios/MyApp/MyApp.entitlements -->
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>com.apple.developer.associated-domains</key>
  <array>
    <string>applinks:myapp.com</string>
    <string>applinks:*.myapp.com</string>
  </array>
</dict>
</plist>
```

```json
// apple-app-site-association (host at https://myapp.com/.well-known/)
{
  "applinks": {
    "apps": [],
    "details": [
      {
        "appID": "TEAM_ID.com.mycompany.myapp",
        "paths": [
          "/post/*",
          "/user/*",
          "/invite/*",
          "/reset-password/*"
        ]
      }
    ]
  }
}
```

### Android App Links Setup

```xml
<!-- android/app/src/main/AndroidManifest.xml -->
<activity android:name=".MainActivity"
          android:launchMode="singleTask">
  <intent-filter android:autoVerify="true">
    <action android:name="android.intent.action.VIEW" />
    <category android:name="android.intent.category.DEFAULT" />
    <category android:name="android.intent.category.BROWSABLE" />
    <data android:scheme="https" android:host="myapp.com" />
  </intent-filter>
  <intent-filter>
    <action android:name="android.intent.action.VIEW" />
    <category android:name="android.intent.category.DEFAULT" />
    <category android:name="android.intent.category.BROWSABLE" />
    <data android:scheme="myapp" />
  </intent-filter>
</activity>
```

### Deferred Deep Links

```typescript
// services/deferred-deep-link.ts
import { Linking, Platform } from 'react-native';
import AsyncStorage from '@react-native-async-storage/async-storage';

export async function handleDeferredDeepLink() {
  // Check if this is the first launch
  const hasLaunched = await AsyncStorage.getItem('hasLaunched');
  if (hasLaunched) return null;

  await AsyncStorage.setItem('hasLaunched', 'true');

  // Check for deferred deep link from install attribution
  try {
    const initialUrl = await Linking.getInitialURL();
    if (initialUrl) return initialUrl;

    // Check clipboard for deep link (with user permission on iOS 16+)
    if (Platform.OS === 'ios') {
      // Use branch.io, adjust, or custom attribution service
      const attributionLink = await fetchAttribution();
      return attributionLink;
    }

    // Android install referrer
    if (Platform.OS === 'android') {
      const referrer = await getInstallReferrer();
      return referrer?.deepLink;
    }
  } catch (error) {
    console.warn('Deferred deep link check failed:', error);
  }

  return null;
}
```

---

## 3. Push Notifications

### Firebase Cloud Messaging Setup

```typescript
// services/notifications.ts
import messaging, { FirebaseMessagingTypes } from '@react-native-firebase/messaging';
import notifee, { AndroidImportance, EventType } from '@notifee/react-native';
import { Platform } from 'react-native';

// Request permission
export async function requestNotificationPermission(): Promise<boolean> {
  if (Platform.OS === 'ios') {
    const authStatus = await messaging().requestPermission({
      alert: true,
      badge: true,
      sound: true,
      provisional: false,
    });
    return (
      authStatus === messaging.AuthorizationStatus.AUTHORIZED ||
      authStatus === messaging.AuthorizationStatus.PROVISIONAL
    );
  }

  // Android 13+ requires POST_NOTIFICATIONS permission
  if (Platform.OS === 'android' && Platform.Version >= 33) {
    const { PermissionsAndroid } = require('react-native');
    const granted = await PermissionsAndroid.request(
      PermissionsAndroid.PERMISSIONS.POST_NOTIFICATIONS
    );
    return granted === PermissionsAndroid.RESULTS.GRANTED;
  }

  return true;
}

// Get FCM token and register with backend
export async function registerForPushNotifications() {
  const hasPermission = await requestNotificationPermission();
  if (!hasPermission) return null;

  const token = await messaging().getToken();
  await registerTokenWithBackend(token);

  // Listen for token refresh
  messaging().onTokenRefresh(async (newToken) => {
    await registerTokenWithBackend(newToken);
  });

  return token;
}

// Create notification channels (Android)
export async function setupNotificationChannels() {
  if (Platform.OS !== 'android') return;

  await notifee.createChannel({
    id: 'messages',
    name: 'Messages',
    importance: AndroidImportance.HIGH,
    sound: 'notification_sound',
    vibration: true,
  });

  await notifee.createChannel({
    id: 'updates',
    name: 'App Updates',
    importance: AndroidImportance.DEFAULT,
  });

  await notifee.createChannel({
    id: 'marketing',
    name: 'Promotions',
    importance: AndroidImportance.LOW,
  });
}

// Display local notification with notifee
export async function displayNotification(
  title: string,
  body: string,
  data?: Record<string, string>,
  channelId = 'messages'
) {
  await notifee.displayNotification({
    title,
    body,
    data,
    android: {
      channelId,
      pressAction: { id: 'default' },
      smallIcon: 'ic_notification',
      color: '#4F46E5',
    },
    ios: {
      sound: 'notification_sound.wav',
      categoryId: data?.category,
    },
  });
}

// Handle foreground messages
export function setupForegroundHandler() {
  messaging().onMessage(async (remoteMessage) => {
    // Display notification using notifee (FCM doesn't auto-display in foreground)
    const { title, body } = remoteMessage.notification ?? {};
    if (title) {
      await displayNotification(title, body ?? '', remoteMessage.data as Record<string, string>);
    }
  });
}

// Handle notification interactions
export function setupNotificationInteractionHandler(
  onNotificationPress: (data: Record<string, string>) => void
) {
  // App opened from quit state via notification
  messaging()
    .getInitialNotification()
    .then((remoteMessage) => {
      if (remoteMessage?.data) {
        onNotificationPress(remoteMessage.data as Record<string, string>);
      }
    });

  // App opened from background via notification
  messaging().onNotificationOpenedApp((remoteMessage) => {
    if (remoteMessage.data) {
      onNotificationPress(remoteMessage.data as Record<string, string>);
    }
  });

  // Handle notifee foreground interactions
  notifee.onForegroundEvent(({ type, detail }) => {
    if (type === EventType.PRESS && detail.notification?.data) {
      onNotificationPress(detail.notification.data as Record<string, string>);
    }
  });
}
```

### Background Message Handler

```typescript
// index.js (must be at entry point)
import messaging from '@react-native-firebase/messaging';
import notifee from '@notifee/react-native';

// Handle background messages
messaging().setBackgroundMessageHandler(async (remoteMessage) => {
  // Process data-only messages
  if (remoteMessage.data?.type === 'chat_message') {
    await notifee.displayNotification({
      title: remoteMessage.data.senderName,
      body: remoteMessage.data.messageText,
      data: remoteMessage.data as Record<string, string>,
      android: {
        channelId: 'messages',
        pressAction: { id: 'default' },
      },
    });
  }
});

// Handle background notification actions
notifee.onBackgroundEvent(async ({ type, detail }) => {
  if (type === EventType.ACTION_PRESS) {
    switch (detail.pressAction?.id) {
      case 'reply':
        // Handle quick reply action
        break;
      case 'mark-read':
        // Handle mark as read
        break;
    }
  }
});
```

### Notification Actions (iOS & Android)

```typescript
// services/notification-actions.ts
import notifee, { AndroidAction, IOSNotificationCategory } from '@notifee/react-native';

export async function setupNotificationActions() {
  // iOS categories with actions
  await notifee.setNotificationCategories([
    {
      id: 'message',
      actions: [
        {
          id: 'reply',
          title: 'Reply',
          input: {
            placeholderText: 'Type your reply...',
          },
        },
        {
          id: 'mark-read',
          title: 'Mark as Read',
        },
      ],
    },
    {
      id: 'social',
      actions: [
        { id: 'like', title: 'Like' },
        { id: 'comment', title: 'Comment' },
      ],
    },
  ]);
}
```

---

## 4. App State Management

### App State Lifecycle

```typescript
// hooks/useAppState.ts
import { useEffect, useRef, useState } from 'react';
import { AppState, AppStateStatus } from 'react-native';

export function useAppState() {
  const appState = useRef(AppState.currentState);
  const [currentState, setCurrentState] = useState(appState.current);

  useEffect(() => {
    const subscription = AppState.addEventListener('change', (nextState: AppStateStatus) => {
      appState.current = nextState;
      setCurrentState(nextState);
    });

    return () => subscription.remove();
  }, []);

  return currentState;
}

// hooks/useAppStateCallback.ts
export function useAppStateCallback(callbacks: {
  onForeground?: () => void;
  onBackground?: () => void;
  onInactive?: () => void;
}) {
  const appState = useRef(AppState.currentState);

  useEffect(() => {
    const subscription = AppState.addEventListener('change', (nextState) => {
      const previousState = appState.current;
      appState.current = nextState;

      if (previousState.match(/inactive|background/) && nextState === 'active') {
        callbacks.onForeground?.();
      } else if (previousState === 'active' && nextState === 'background') {
        callbacks.onBackground?.();
      } else if (nextState === 'inactive') {
        callbacks.onInactive?.();
      }
    });

    return () => subscription.remove();
  }, [callbacks]);
}

// Usage: auto-refresh data when app returns to foreground
function HomeScreen() {
  const { refetch } = useQuery({ queryKey: ['feed'] });

  useAppStateCallback({
    onForeground: () => {
      refetch();
    },
    onBackground: () => {
      // Pause expensive operations
    },
  });
}
```

### Session Tracking

```typescript
// services/session.ts
import { AppState, AppStateStatus } from 'react-native';
import { MMKV } from 'react-native-mmkv';

const sessionStorage = new MMKV({ id: 'session' });

class SessionTracker {
  private sessionStartTime = 0;
  private backgroundTime = 0;
  private readonly SESSION_TIMEOUT = 30 * 60 * 1000; // 30 minutes

  constructor() {
    this.startSession();

    AppState.addEventListener('change', (state: AppStateStatus) => {
      if (state === 'background') {
        this.backgroundTime = Date.now();
      } else if (state === 'active') {
        const elapsed = Date.now() - this.backgroundTime;
        if (elapsed > this.SESSION_TIMEOUT) {
          this.endSession();
          this.startSession();
        }
      }
    });
  }

  private startSession() {
    this.sessionStartTime = Date.now();
    const sessionId = `session_${Date.now()}`;
    sessionStorage.set('currentSessionId', sessionId);
    analytics.track('session_start', { sessionId });
  }

  private endSession() {
    const duration = Date.now() - this.sessionStartTime;
    const sessionId = sessionStorage.getString('currentSessionId');
    analytics.track('session_end', { sessionId, durationMs: duration });
  }
}

export const sessionTracker = new SessionTracker();
```

---

## 5. Background Tasks

### React Native Background Fetch

```typescript
// services/background-tasks.ts
import BackgroundFetch from 'react-native-background-fetch';

export async function configureBackgroundFetch() {
  await BackgroundFetch.configure(
    {
      minimumFetchInterval: 15, // minutes (iOS minimum)
      stopOnTerminate: false,
      startOnBoot: true,
      enableHeadless: true,
      requiredNetworkType: BackgroundFetch.NETWORK_TYPE_ANY,
    },
    async (taskId) => {
      // Perform background work
      console.log('[BackgroundFetch] Task:', taskId);

      try {
        await syncPendingData();
        await prefetchContent();
        await cleanupExpiredCache();
      } catch (error) {
        console.error('[BackgroundFetch] Error:', error);
      }

      // Required: signal completion
      BackgroundFetch.finish(taskId);
    },
    (taskId) => {
      // Task timed out — clean up and finish
      console.warn('[BackgroundFetch] Timeout:', taskId);
      BackgroundFetch.finish(taskId);
    }
  );

  // Check status
  const status = await BackgroundFetch.status();
  console.log('[BackgroundFetch] Status:', status);
}

// Schedule a one-time task
export async function scheduleBackgroundTask(taskId: string, delayMs: number) {
  await BackgroundFetch.scheduleTask({
    taskId,
    delay: delayMs,
    periodic: false,
    requiresNetworkConnectivity: true,
    requiresCharging: false,
  });
}

// Headless task (runs even when app is terminated — Android only)
BackgroundFetch.registerHeadlessTask(async ({ taskId, timeout }) => {
  if (timeout) {
    BackgroundFetch.finish(taskId);
    return;
  }

  await syncPendingData();
  BackgroundFetch.finish(taskId);
});
```

### Background Location Tracking

```typescript
// services/background-location.ts
import Geolocation from 'react-native-geolocation-service';
import BackgroundGeolocation from 'react-native-background-geolocation';

export async function startBackgroundLocationTracking() {
  const state = await BackgroundGeolocation.ready({
    desiredAccuracy: BackgroundGeolocation.DESIRED_ACCURACY_HIGH,
    distanceFilter: 50, // meters
    stopTimeout: 5, // minutes
    debug: __DEV__,
    logLevel: __DEV__
      ? BackgroundGeolocation.LOG_LEVEL_VERBOSE
      : BackgroundGeolocation.LOG_LEVEL_OFF,
    stopOnTerminate: false,
    startOnBoot: true,
    enableHeadless: true,
    // HTTP sync configuration
    url: `${API_URL}/locations`,
    batchSync: true,
    maxBatchSize: 50,
    headers: {
      Authorization: `Bearer ${getAccessToken()}`,
    },
  });

  if (!state.enabled) {
    await BackgroundGeolocation.start();
  }

  // Listen for location updates
  BackgroundGeolocation.onLocation((location) => {
    console.log('[Location]', location.coords);
  });

  // Listen for motion state changes
  BackgroundGeolocation.onMotionChange((event) => {
    console.log('[Motion]', event.isMoving ? 'Moving' : 'Stationary');
  });
}
```

---

## 6. Platform-Specific Patterns

### Platform-Specific Components

```typescript
// components/DatePicker/index.tsx (shared interface)
export interface DatePickerProps {
  value: Date;
  onChange: (date: Date) => void;
  minimumDate?: Date;
  maximumDate?: Date;
  mode?: 'date' | 'time' | 'datetime';
}

// components/DatePicker/DatePicker.ios.tsx
import DateTimePicker from '@react-native-community/datetimepicker';
import { DatePickerProps } from './index';

export function DatePicker({ value, onChange, mode = 'date', ...props }: DatePickerProps) {
  return (
    <DateTimePicker
      value={value}
      mode={mode}
      display="spinner"
      onChange={(_, date) => date && onChange(date)}
      {...props}
    />
  );
}

// components/DatePicker/DatePicker.android.tsx
import { useState } from 'react';
import { Pressable, Text } from 'react-native';
import DateTimePicker from '@react-native-community/datetimepicker';
import { DatePickerProps } from './index';

export function DatePicker({ value, onChange, mode = 'date', ...props }: DatePickerProps) {
  const [show, setShow] = useState(false);

  return (
    <>
      <Pressable onPress={() => setShow(true)}>
        <Text>{value.toLocaleDateString()}</Text>
      </Pressable>
      {show && (
        <DateTimePicker
          value={value}
          mode={mode}
          display="default"
          onChange={(_, date) => {
            setShow(false);
            if (date) onChange(date);
          }}
          {...props}
        />
      )}
    </>
  );
}
```

### Platform-Specific Styles

```typescript
// utils/platform.ts
import { Platform, StyleSheet } from 'react-native';

export const isIOS = Platform.OS === 'ios';
export const isAndroid = Platform.OS === 'android';

export function platformSelect<T>(options: { ios: T; android: T; default?: T }): T {
  return Platform.select(options) ?? options.default ?? options.ios;
}

// Shadow utilities (iOS uses shadow*, Android uses elevation)
export function createShadow(elevation: number) {
  return StyleSheet.create({
    shadow: {
      ...Platform.select({
        ios: {
          shadowColor: '#000',
          shadowOffset: { width: 0, height: elevation / 2 },
          shadowOpacity: 0.1 + elevation * 0.02,
          shadowRadius: elevation,
        },
        android: {
          elevation,
        },
      }),
    },
  }).shadow;
}

// Safe area handling
import { useSafeAreaInsets } from 'react-native-safe-area-context';

export function useBottomTabBarHeight() {
  const insets = useSafeAreaInsets();
  const TAB_BAR_HEIGHT = 49;
  return TAB_BAR_HEIGHT + insets.bottom;
}
```

### Permissions Handling

```typescript
// hooks/usePermission.ts
import { useCallback, useEffect, useState } from 'react';
import { Alert, Linking, Platform } from 'react-native';
import {
  check,
  request,
  PERMISSIONS,
  RESULTS,
  Permission,
  PermissionStatus,
} from 'react-native-permissions';

const PERMISSION_MAP = {
  camera: Platform.select({
    ios: PERMISSIONS.IOS.CAMERA,
    android: PERMISSIONS.ANDROID.CAMERA,
  })!,
  photoLibrary: Platform.select({
    ios: PERMISSIONS.IOS.PHOTO_LIBRARY,
    android:
      Platform.Version >= 33
        ? PERMISSIONS.ANDROID.READ_MEDIA_IMAGES
        : PERMISSIONS.ANDROID.READ_EXTERNAL_STORAGE,
  })!,
  location: Platform.select({
    ios: PERMISSIONS.IOS.LOCATION_WHEN_IN_USE,
    android: PERMISSIONS.ANDROID.ACCESS_FINE_LOCATION,
  })!,
  microphone: Platform.select({
    ios: PERMISSIONS.IOS.MICROPHONE,
    android: PERMISSIONS.ANDROID.RECORD_AUDIO,
  })!,
} as const;

type PermissionType = keyof typeof PERMISSION_MAP;

export function usePermission(type: PermissionType) {
  const [status, setStatus] = useState<PermissionStatus | null>(null);
  const permission = PERMISSION_MAP[type];

  const checkPermission = useCallback(async () => {
    const result = await check(permission);
    setStatus(result);
    return result;
  }, [permission]);

  const requestPermission = useCallback(async () => {
    const result = await request(permission);
    setStatus(result);

    if (result === RESULTS.BLOCKED) {
      Alert.alert(
        'Permission Required',
        `Please enable ${type} access in your device settings.`,
        [
          { text: 'Cancel', style: 'cancel' },
          { text: 'Open Settings', onPress: () => Linking.openSettings() },
        ]
      );
    }

    return result;
  }, [permission, type]);

  useEffect(() => {
    checkPermission();
  }, [checkPermission]);

  return {
    status,
    isGranted: status === RESULTS.GRANTED || status === RESULTS.LIMITED,
    isDenied: status === RESULTS.DENIED,
    isBlocked: status === RESULTS.BLOCKED,
    request: requestPermission,
    check: checkPermission,
  };
}
```

---

## 7. Navigation Patterns

### Auth Flow with React Navigation

```typescript
// navigation/RootNavigator.tsx
import { NavigationContainer } from '@react-navigation/native';
import { createNativeStackNavigator } from '@react-navigation/native-stack';
import { useAuthStore } from '../stores/auth-store';
import { linking } from './linking';

const Stack = createNativeStackNavigator<RootStackParamList>();

export function RootNavigator() {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
  const isLoading = useAuthStore((s) => s.isLoading);

  if (isLoading) {
    return <SplashScreen />;
  }

  return (
    <NavigationContainer linking={linking}>
      <Stack.Navigator screenOptions={{ headerShown: false }}>
        {isAuthenticated ? (
          <>
            <Stack.Screen name="MainTabs" component={MainTabNavigator} />
            <Stack.Group screenOptions={{ presentation: 'modal' }}>
              <Stack.Screen name="CreatePost" component={CreatePostScreen} />
              <Stack.Screen name="ImageViewer" component={ImageViewerScreen} />
            </Stack.Group>
          </>
        ) : (
          <>
            <Stack.Screen name="Login" component={LoginScreen} />
            <Stack.Screen name="Register" component={RegisterScreen} />
            <Stack.Screen name="ForgotPassword" component={ForgotPasswordScreen} />
          </>
        )}
      </Stack.Navigator>
    </NavigationContainer>
  );
}
```

### Type-Safe Navigation

```typescript
// navigation/types.ts
import { NavigatorScreenParams, CompositeScreenProps } from '@react-navigation/native';
import { NativeStackScreenProps } from '@react-navigation/native-stack';
import { BottomTabScreenProps } from '@react-navigation/bottom-tabs';

// Root stack
export type RootStackParamList = {
  MainTabs: NavigatorScreenParams<MainTabParamList>;
  Login: undefined;
  Register: undefined;
  ForgotPassword: undefined;
  ResetPassword: { token: string };
  CreatePost: undefined;
  ImageViewer: { imageId: string };
};

// Tab navigator
export type MainTabParamList = {
  Home: NavigatorScreenParams<HomeStackParamList>;
  Search: undefined;
  Notifications: undefined;
  Profile: NavigatorScreenParams<ProfileStackParamList>;
};

// Nested stack inside Home tab
export type HomeStackParamList = {
  Feed: undefined;
  PostDetail: { postId: string };
  UserProfile: { userId: string };
};

// Composed screen props for nested navigators
export type HomeScreenProps<T extends keyof HomeStackParamList> = CompositeScreenProps<
  NativeStackScreenProps<HomeStackParamList, T>,
  CompositeScreenProps<
    BottomTabScreenProps<MainTabParamList, 'Home'>,
    NativeStackScreenProps<RootStackParamList>
  >
>;

// Type-safe navigation hook
import { useNavigation } from '@react-navigation/native';
import { NativeStackNavigationProp } from '@react-navigation/native-stack';

export function useAppNavigation() {
  return useNavigation<NativeStackNavigationProp<RootStackParamList>>();
}
```

---

## 8. Error Boundary Patterns

### Global Error Boundary

```typescript
// components/ErrorBoundary.tsx
import React, { Component, ErrorInfo, ReactNode } from 'react';
import { View, Text, TouchableOpacity, StyleSheet } from 'react-native';
import * as Sentry from '@sentry/react-native';

interface Props {
  children: ReactNode;
  fallback?: ReactNode;
  onError?: (error: Error, errorInfo: ErrorInfo) => void;
}

interface State {
  hasError: boolean;
  error: Error | null;
}

export class ErrorBoundary extends Component<Props, State> {
  state: State = { hasError: false, error: null };

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    Sentry.captureException(error, { extra: { componentStack: errorInfo.componentStack } });
    this.props.onError?.(error, errorInfo);
  }

  resetError = () => {
    this.setState({ hasError: false, error: null });
  };

  render() {
    if (this.state.hasError) {
      if (this.props.fallback) return this.props.fallback;

      return (
        <View style={styles.container}>
          <Text style={styles.title}>Something went wrong</Text>
          <Text style={styles.message}>{this.state.error?.message}</Text>
          <TouchableOpacity style={styles.button} onPress={this.resetError}>
            <Text style={styles.buttonText}>Try Again</Text>
          </TouchableOpacity>
        </View>
      );
    }

    return this.props.children;
  }
}

const styles = StyleSheet.create({
  container: { flex: 1, justifyContent: 'center', alignItems: 'center', padding: 24 },
  title: { fontSize: 20, fontWeight: '700', marginBottom: 8 },
  message: { fontSize: 14, color: '#666', textAlign: 'center', marginBottom: 24 },
  button: { backgroundColor: '#4F46E5', paddingHorizontal: 24, paddingVertical: 12, borderRadius: 8 },
  buttonText: { color: '#fff', fontWeight: '600' },
});
```

---

## Quick Reference: Pattern Decision Tree

```
Need offline data?
├── Simple key-value → MMKV
├── Relational with sync → WatermelonDB
├── Offline write queue → OfflineQueue + MMKV
└── Complex sync → WatermelonDB + custom sync

Need navigation?
├── Simple stack → React Navigation Stack
├── Tabs + stacks → React Navigation Bottom Tabs + Stack
├── Auth flow → Conditional screens in root navigator
├── Deep linking → linking config + universal links
└── File-based routing → Expo Router (see expo-guide.md)

Need push notifications?
├── Basic → Firebase Cloud Messaging
├── Rich notifications → FCM + notifee
├── Actions/replies → notifee categories + actions
└── Background processing → Background message handler

Need background work?
├── Periodic sync → react-native-background-fetch
├── Location tracking → react-native-background-geolocation
├── File uploads → react-native-background-upload
└── Music playback → react-native-track-player

Platform-specific UI?
├── Date picker → Platform-specific files (.ios.tsx/.android.tsx)
├── Shadows → Platform.select (shadowColor vs elevation)
├── Haptics → expo-haptics or react-native-haptic-feedback
└── Safe areas → react-native-safe-area-context
```
