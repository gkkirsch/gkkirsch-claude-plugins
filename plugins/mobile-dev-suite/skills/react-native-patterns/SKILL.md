---
name: react-native-patterns
description: >
  React Native and Expo patterns for production mobile apps.
  Use when building navigation, state management, native modules,
  platform-specific code, or Expo configurations.
  Triggers: "react native", "expo", "mobile app", "navigation",
  "native module", "platform specific", "app store".
  NOT for: web-only React, Flutter, or native iOS/Android development.
version: 1.0.0
allowed-tools: Read, Grep, Glob, Bash
---

# React Native Patterns

## Project Structure (Expo)

```
app/
├── (tabs)/              # Tab-based navigation group
│   ├── _layout.tsx      # Tab navigator config
│   ├── index.tsx         # Home tab
│   ├── search.tsx        # Search tab
│   └── profile.tsx       # Profile tab
├── (auth)/              # Auth flow group
│   ├── _layout.tsx
│   ├── login.tsx
│   └── signup.tsx
├── _layout.tsx          # Root layout (providers, fonts)
├── +not-found.tsx       # 404 screen
└── [id].tsx             # Dynamic route
components/
├── ui/                  # Reusable primitives
│   ├── Button.tsx
│   ├── Input.tsx
│   └── Card.tsx
├── features/            # Feature-specific components
hooks/
├── useAuth.ts
├── useApi.ts
└── useKeyboard.ts
lib/
├── api.ts               # API client
├── storage.ts           # Async storage wrapper
└── constants.ts
```

## Navigation (Expo Router)

```tsx
// app/_layout.tsx — Root layout with auth guard
import { Stack } from 'expo-router';
import { useAuth } from '@/hooks/useAuth';
import { Redirect } from 'expo-router';

export default function RootLayout() {
  const { isAuthenticated, isLoading } = useAuth();

  if (isLoading) return <SplashScreen />;

  return (
    <Stack screenOptions={{ headerShown: false }}>
      {isAuthenticated ? (
        <Stack.Screen name="(tabs)" />
      ) : (
        <Stack.Screen name="(auth)" />
      )}
      <Stack.Screen name="+not-found" options={{ title: 'Not Found' }} />
    </Stack>
  );
}

// app/(tabs)/_layout.tsx — Tab navigator
import { Tabs } from 'expo-router';
import { Home, Search, User } from 'lucide-react-native';

export default function TabLayout() {
  return (
    <Tabs
      screenOptions={{
        tabBarActiveTintColor: '#3b82f6',
        tabBarInactiveTintColor: '#9ca3af',
        tabBarStyle: {
          backgroundColor: '#111827',
          borderTopColor: '#374151',
        },
        headerStyle: { backgroundColor: '#111827' },
        headerTintColor: '#f9fafb',
      }}
    >
      <Tabs.Screen name="index" options={{
        title: 'Home',
        tabBarIcon: ({ color, size }) => <Home color={color} size={size} />,
      }} />
      <Tabs.Screen name="search" options={{
        title: 'Search',
        tabBarIcon: ({ color, size }) => <Search color={color} size={size} />,
      }} />
      <Tabs.Screen name="profile" options={{
        title: 'Profile',
        tabBarIcon: ({ color, size }) => <User color={color} size={size} />,
      }} />
    </Tabs>
  );
}
```

## Platform-Specific Code

```tsx
import { Platform, StyleSheet } from 'react-native';

// Platform.select for inline values
const styles = StyleSheet.create({
  shadow: Platform.select({
    ios: {
      shadowColor: '#000',
      shadowOffset: { width: 0, height: 2 },
      shadowOpacity: 0.25,
      shadowRadius: 3.84,
    },
    android: {
      elevation: 5,
    },
    default: {},
  }),
  container: {
    paddingTop: Platform.OS === 'ios' ? 44 : 0,
  },
});

// Platform-specific file imports
// Button.ios.tsx — iOS-specific implementation
// Button.android.tsx — Android-specific implementation
// Button.tsx — fallback
// React Native auto-resolves: import Button from './Button'
```

## Secure Storage

```typescript
import * as SecureStore from 'expo-secure-store';
import AsyncStorage from '@react-native-async-storage/async-storage';

// Secure storage for sensitive data (tokens, credentials)
const secureStorage = {
  async get(key: string): Promise<string | null> {
    return SecureStore.getItemAsync(key);
  },
  async set(key: string, value: string): Promise<void> {
    await SecureStore.setItemAsync(key, value);
  },
  async remove(key: string): Promise<void> {
    await SecureStore.deleteItemAsync(key);
  },
};

// Async storage for non-sensitive data (preferences, cache)
const appStorage = {
  async getJSON<T>(key: string): Promise<T | null> {
    const raw = await AsyncStorage.getItem(key);
    return raw ? JSON.parse(raw) : null;
  },
  async setJSON<T>(key: string, value: T): Promise<void> {
    await AsyncStorage.setItem(key, JSON.stringify(value));
  },
  async remove(key: string): Promise<void> {
    await AsyncStorage.removeItem(key);
  },
};
```

## API Client with Offline Support

```typescript
import NetInfo from '@react-native-community/netinfo';

class MobileApiClient {
  private baseUrl: string;
  private getToken: () => Promise<string | null>;

  constructor(baseUrl: string, getToken: () => Promise<string | null>) {
    this.baseUrl = baseUrl;
    this.getToken = getToken;
  }

  async request<T>(path: string, options: RequestInit = {}): Promise<T> {
    const netState = await NetInfo.fetch();
    if (!netState.isConnected) {
      throw new OfflineError('No internet connection');
    }

    const token = await this.getToken();
    const response = await fetch(`${this.baseUrl}${path}`, {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
        ...options.headers,
      },
    });

    if (response.status === 401) {
      await secureStorage.remove('auth_token');
      throw new AuthError('Session expired');
    }

    if (!response.ok) {
      throw new ApiError(response.status, await response.text());
    }

    return response.json();
  }
}

class OfflineError extends Error { constructor(msg: string) { super(msg); this.name = 'OfflineError'; } }
class AuthError extends Error { constructor(msg: string) { super(msg); this.name = 'AuthError'; } }
class ApiError extends Error {
  status: number;
  constructor(status: number, msg: string) { super(msg); this.name = 'ApiError'; this.status = status; }
}
```

## Gotchas

1. **FlatList keyExtractor must return a string** — `keyExtractor={(item) => item.id}` fails if `id` is a number. Always convert: `keyExtractor={(item) => String(item.id)}`. Missing keyExtractor causes "VirtualizedList: missing keys" warnings and re-render bugs.

2. **Keyboard avoidance differs by platform** — on iOS use `KeyboardAvoidingView` with `behavior="padding"`, on Android use `behavior="height"` or just `android:windowSoftInputMode="adjustResize"` in AndroidManifest.xml. Neither works universally.

3. **AsyncStorage has a 2MB default limit on Android** — large datasets silently fail. For substantial data, use SQLite (expo-sqlite) or MMKV. AsyncStorage is only for small key-value pairs.

4. **Expo EAS Build vs Expo Go** — Expo Go doesn't support custom native modules (camera, push notifications, biometrics). Test with EAS development builds for production-representative behavior. `npx expo run:ios` builds locally.

5. **React Native StyleSheet doesn't support CSS shorthand** — `padding: '10px 20px'` fails silently. Use explicit properties: `paddingVertical: 10, paddingHorizontal: 20`. No `gap` in older RN versions (use `marginBottom` on children instead).

6. **Hot reload loses navigation state and context** — during development, fast refresh preserves component state but resets navigation stack and context providers. Add `initialRouteName` in navigation config for reliable dev experience.
