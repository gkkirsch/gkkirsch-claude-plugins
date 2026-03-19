# Expo Guide Reference

Production patterns for Expo development. This reference covers project structure, config plugins, environment management, OTA updates, EAS workflows, and monorepo setup.

---

## 1. Project Structure

### Recommended Expo Router Project Layout

```
my-app/
├── app/                        # Expo Router file-based routes
│   ├── _layout.tsx             # Root layout (providers, fonts, splash)
│   ├── index.tsx               # Home screen (/)
│   ├── +not-found.tsx          # 404 handler
│   ├── +html.tsx               # Custom HTML wrapper (web)
│   ├── (auth)/                 # Auth group (unauthed only)
│   │   ├── _layout.tsx         # Auth layout
│   │   ├── login.tsx           # /login
│   │   ├── register.tsx        # /register
│   │   └── forgot-password.tsx # /forgot-password
│   ├── (tabs)/                 # Main tab group (authed only)
│   │   ├── _layout.tsx         # Tab bar layout
│   │   ├── index.tsx           # Home tab
│   │   ├── search.tsx          # Search tab
│   │   ├── notifications.tsx   # Notifications tab
│   │   └── profile/            # Profile tab with nested stack
│   │       ├── _layout.tsx     # Profile stack layout
│   │       ├── index.tsx       # Profile screen
│   │       ├── edit.tsx        # Edit profile
│   │       └── settings.tsx    # Settings
│   ├── post/
│   │   └── [id].tsx            # /post/:id (dynamic route)
│   ├── user/
│   │   └── [id].tsx            # /user/:id
│   └── (modals)/               # Modal group
│       ├── _layout.tsx         # Modal presentation config
│       ├── create-post.tsx     # Create post modal
│       └── image/[id].tsx      # Image viewer modal
├── src/
│   ├── components/             # Shared components
│   │   ├── ui/                 # Design system primitives
│   │   ├── forms/              # Form components
│   │   └── layout/             # Layout components
│   ├── hooks/                  # Custom hooks
│   ├── stores/                 # Zustand stores
│   ├── services/               # API clients, services
│   ├── utils/                  # Utility functions
│   ├── types/                  # TypeScript types
│   └── constants/              # App constants
├── assets/                     # Static assets (images, fonts)
├── plugins/                    # Custom config plugins
├── app.config.ts               # Dynamic Expo config
├── eas.json                    # EAS Build/Submit/Update config
├── metro.config.js             # Metro bundler config
├── babel.config.js             # Babel config
└── tsconfig.json               # TypeScript config
```

### app.config.ts (Dynamic Config)

```typescript
// app.config.ts
import { ExpoConfig, ConfigContext } from 'expo/config';

const IS_DEV = process.env.APP_VARIANT === 'development';
const IS_PREVIEW = process.env.APP_VARIANT === 'preview';

const getAppName = () => {
  if (IS_DEV) return 'MyApp (Dev)';
  if (IS_PREVIEW) return 'MyApp (Preview)';
  return 'MyApp';
};

const getBundleId = () => {
  if (IS_DEV) return 'com.mycompany.myapp.dev';
  if (IS_PREVIEW) return 'com.mycompany.myapp.preview';
  return 'com.mycompany.myapp';
};

export default ({ config }: ConfigContext): ExpoConfig => ({
  ...config,
  name: getAppName(),
  slug: 'my-app',
  version: '1.0.0',
  orientation: 'portrait',
  icon: './assets/icon.png',
  scheme: 'myapp',
  userInterfaceStyle: 'automatic',
  splash: {
    image: './assets/splash.png',
    resizeMode: 'contain',
    backgroundColor: '#ffffff',
  },
  assetBundlePatterns: ['**/*'],
  ios: {
    supportsTablet: true,
    bundleIdentifier: getBundleId(),
    associatedDomains: ['applinks:myapp.com'],
    infoPlist: {
      NSCameraUsageDescription: 'Take photos for your posts',
      NSPhotoLibraryUsageDescription: 'Select photos for your posts',
      NSLocationWhenInUseUsageDescription: 'Show nearby content',
    },
  },
  android: {
    adaptiveIcon: {
      foregroundImage: './assets/adaptive-icon.png',
      backgroundColor: '#ffffff',
    },
    package: getBundleId(),
    intentFilters: [
      {
        action: 'VIEW',
        autoVerify: true,
        data: [{ scheme: 'https', host: 'myapp.com', pathPrefix: '/' }],
        category: ['BROWSABLE', 'DEFAULT'],
      },
    ],
  },
  web: {
    bundler: 'metro',
    output: 'static',
    favicon: './assets/favicon.png',
  },
  plugins: [
    'expo-router',
    'expo-secure-store',
    'expo-font',
    ['expo-camera', { cameraPermission: 'Allow $(PRODUCT_NAME) to access your camera' }],
    ['expo-image-picker', { photosPermission: 'Allow $(PRODUCT_NAME) to access your photos' }],
    ['expo-location', { isAndroidBackgroundLocationEnabled: false }],
    './plugins/with-custom-splash',
  ],
  experiments: {
    typedRoutes: true,
  },
  extra: {
    eas: { projectId: 'your-eas-project-id' },
    apiUrl: process.env.API_URL ?? 'https://api.myapp.com',
  },
  updates: {
    url: 'https://u.expo.dev/your-eas-project-id',
  },
  runtimeVersion: {
    policy: 'appVersion',
  },
});
```

---

## 2. Config Plugins

### What Config Plugins Do

Config plugins modify native project configuration during `expo prebuild` without requiring you to maintain native code directly. They can modify:

- `Info.plist` (iOS), `AndroidManifest.xml` (Android)
- `build.gradle`, `Podfile`, `AppDelegate`
- Add native files, frameworks, dependencies
- Modify Xcode project settings

### Creating a Config Plugin

```typescript
// plugins/with-custom-splash.ts
import { ConfigPlugin, withInfoPlist, withAndroidManifest } from 'expo/config-plugins';

const withCustomSplash: ConfigPlugin<{ backgroundColor?: string }> = (config, props = {}) => {
  const bgColor = props.backgroundColor ?? '#FFFFFF';

  // Modify iOS Info.plist
  config = withInfoPlist(config, (config) => {
    config.modResults.UIViewControllerBasedStatusBarAppearance = false;
    return config;
  });

  // Modify AndroidManifest.xml
  config = withAndroidManifest(config, (config) => {
    const mainApplication = config.modResults.manifest.application?.[0];
    if (mainApplication) {
      mainApplication.$['android:theme'] = '@style/AppTheme';
    }
    return config;
  });

  return config;
};

export default withCustomSplash;
```

### Common Config Plugin Patterns

```typescript
// plugins/with-google-services.ts — Add Firebase
import { ConfigPlugin, withAppBuildGradle, withProjectBuildGradle } from 'expo/config-plugins';

const withGoogleServices: ConfigPlugin = (config) => {
  // Add classpath to project build.gradle
  config = withProjectBuildGradle(config, (config) => {
    if (config.modResults.contents.includes('com.google.gms:google-services')) {
      return config;
    }
    config.modResults.contents = config.modResults.contents.replace(
      /dependencies\s*{/,
      `dependencies {\n        classpath 'com.google.gms:google-services:4.4.0'`
    );
    return config;
  });

  // Apply plugin in app build.gradle
  config = withAppBuildGradle(config, (config) => {
    if (config.modResults.contents.includes("id 'com.google.gms.google-services'")) {
      return config;
    }
    config.modResults.contents = config.modResults.contents.replace(
      /plugins\s*{/,
      `plugins {\n    id 'com.google.gms.google-services'`
    );
    return config;
  });

  return config;
};

export default withGoogleServices;
```

### Expo Modules API

```typescript
// modules/my-native-module/index.ts
import { NativeModule, requireNativeModule } from 'expo-modules-core';

interface MyModuleEvents {
  onDataReceived: (data: { value: string }) => void;
}

declare class MyNativeModule extends NativeModule<MyModuleEvents> {
  hello(name: string): string;
  fetchDataAsync(): Promise<string>;
}

export default requireNativeModule<MyNativeModule>('MyNativeModule');

// modules/my-native-module/ios/MyNativeModule.swift
import ExpoModulesCore

public class MyNativeModule: Module {
  public func definition() -> ModuleDefinition {
    Name("MyNativeModule")

    Function("hello") { (name: String) -> String in
      return "Hello, \(name)!"
    }

    AsyncFunction("fetchDataAsync") { (promise: Promise) in
      // Async native work
      DispatchQueue.global().async {
        let result = "native data"
        promise.resolve(result)
      }
    }

    Events("onDataReceived")
  }
}
```

---

## 3. Environment Management

### Multiple App Variants

```json
// eas.json
{
  "cli": { "version": ">= 12.0.0" },
  "build": {
    "development": {
      "developmentClient": true,
      "distribution": "internal",
      "env": {
        "APP_VARIANT": "development",
        "API_URL": "https://dev-api.myapp.com"
      },
      "ios": {
        "simulator": true
      }
    },
    "preview": {
      "distribution": "internal",
      "env": {
        "APP_VARIANT": "preview",
        "API_URL": "https://staging-api.myapp.com"
      }
    },
    "production": {
      "env": {
        "APP_VARIANT": "production",
        "API_URL": "https://api.myapp.com"
      },
      "autoIncrement": true
    }
  },
  "submit": {
    "production": {
      "ios": {
        "appleId": "your@email.com",
        "ascAppId": "1234567890",
        "appleTeamId": "ABC123"
      },
      "android": {
        "serviceAccountKeyPath": "./google-service-account.json",
        "track": "internal"
      }
    }
  }
}
```

### Environment Variables in Code

```typescript
// src/constants/env.ts
import Constants from 'expo-constants';

interface AppConfig {
  apiUrl: string;
  appVariant: 'development' | 'preview' | 'production';
  sentryDsn: string;
  analyticsKey: string;
}

function getConfig(): AppConfig {
  const extra = Constants.expoConfig?.extra;

  return {
    apiUrl: extra?.apiUrl ?? 'https://api.myapp.com',
    appVariant: (process.env.APP_VARIANT as AppConfig['appVariant']) ?? 'production',
    sentryDsn: process.env.SENTRY_DSN ?? '',
    analyticsKey: process.env.ANALYTICS_KEY ?? '',
  };
}

export const config = getConfig();
```

---

## 4. OTA Updates with EAS Update

### Update Strategy

| Strategy | When to Use | Risk |
|----------|-------------|------|
| Automatic on launch | Most apps, background updates | Low — users get updates naturally |
| Check + prompt | Important updates the user should know about | Low — user controls timing |
| Force update | Critical bug fixes, security patches | Medium — interrupts user |
| Channel-based rollout | Gradual rollout to % of users | Low — can roll back |

### Implementing EAS Update

```typescript
// services/updates.ts
import * as Updates from 'expo-updates';
import { Alert } from 'react-native';

export async function checkForUpdates(options: { force?: boolean } = {}) {
  if (__DEV__) return; // No updates in dev

  try {
    const update = await Updates.checkForUpdateAsync();

    if (update.isAvailable) {
      if (options.force) {
        await forceUpdate();
      } else {
        promptForUpdate();
      }
    }
  } catch (error) {
    console.error('Update check failed:', error);
  }
}

async function forceUpdate() {
  try {
    const result = await Updates.fetchUpdateAsync();
    if (result.isNew) {
      await Updates.reloadAsync();
    }
  } catch (error) {
    console.error('Update fetch failed:', error);
  }
}

function promptForUpdate() {
  Alert.alert(
    'Update Available',
    'A new version is available. Would you like to update now?',
    [
      { text: 'Later', style: 'cancel' },
      {
        text: 'Update',
        onPress: async () => {
          await forceUpdate();
        },
      },
    ]
  );
}

// Auto-check on app launch (in root layout)
export function useUpdateCheck() {
  useEffect(() => {
    checkForUpdates();
  }, []);
}
```

### EAS Update Channels and Branches

```bash
# Create update channels
eas channel:create production
eas channel:create staging
eas channel:create beta

# Link branches to channels
eas channel:edit production --branch production
eas channel:edit staging --branch staging

# Publish update to a branch
eas update --branch production --message "Fix checkout bug"

# Rollback: point channel to previous branch
eas channel:rollback production
```

### Runtime Version Policy

```typescript
// app.config.ts — Runtime version strategies
export default {
  // Option 1: Tied to app version (most common)
  runtimeVersion: {
    policy: 'appVersion', // "1.0.0" → only updates for 1.0.0 builds
  },

  // Option 2: SDK-based (for Expo Go compatibility)
  runtimeVersion: {
    policy: 'sdkVersion', // Updates compatible with same SDK
  },

  // Option 3: Fingerprint-based (recommended for production)
  runtimeVersion: {
    policy: 'fingerprint', // Hash of native dependencies — auto-detects compatibility
  },

  // Option 4: Custom string (full control)
  runtimeVersion: '1.0.0',
};
```

---

## 5. EAS Workflows

### Build Workflow

```bash
# Development build (with dev client)
eas build --profile development --platform ios
eas build --profile development --platform android

# Preview build (internal distribution)
eas build --profile preview --platform all

# Production build
eas build --profile production --platform all

# Local build (no EAS servers)
eas build --local --profile development --platform ios
```

### Submit Workflow

```bash
# Submit to App Store
eas submit --platform ios --latest

# Submit to Google Play
eas submit --platform android --latest

# Submit specific build
eas submit --platform ios --id BUILD_ID
```

### CI/CD Integration

```yaml
# .github/workflows/eas-build.yml
name: EAS Build & Submit
on:
  push:
    branches: [main]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-node@v4
        with:
          node-version: 20
          cache: npm

      - run: npm ci

      - uses: expo/expo-github-action@v8
        with:
          eas-version: latest
          token: ${{ secrets.EXPO_TOKEN }}

      - name: Build for all platforms
        run: eas build --platform all --non-interactive --no-wait

      # For tagged releases, also submit
      - name: Submit to stores
        if: startsWith(github.ref, 'refs/tags/v')
        run: |
          eas submit --platform ios --latest --non-interactive
          eas submit --platform android --latest --non-interactive
```

---

## 6. Development Workflow

### Dev Builds vs Expo Go

| Feature | Expo Go | Dev Build |
|---------|---------|-----------|
| Custom native code | No | Yes |
| Config plugins | No | Yes |
| Third-party native libs | Limited | Full |
| Setup time | Instant | 5-10 min first build |
| Updates | OTA from Expo | OTA from your project |
| Best for | Prototyping, learning | Production development |

### Prebuild Workflow

```bash
# Generate native projects from config
npx expo prebuild

# Clean and regenerate
npx expo prebuild --clean

# Platform-specific
npx expo prebuild --platform ios
npx expo prebuild --platform android

# Run with native build tools
npx expo run:ios
npx expo run:android

# Run on specific device/simulator
npx expo run:ios --device "iPhone 15 Pro"
npx expo run:android --device "Pixel_7_API_34"
```

---

## 7. Monorepo with Expo

### Workspace Setup

```
my-monorepo/
├── apps/
│   ├── mobile/                 # Expo app
│   │   ├── app/               # Expo Router routes
│   │   ├── app.config.ts
│   │   └── package.json
│   └── web/                    # Next.js app (optional)
│       └── package.json
├── packages/
│   ├── ui/                     # Shared UI components
│   │   ├── src/
│   │   └── package.json
│   ├── api-client/             # Shared API client
│   │   ├── src/
│   │   └── package.json
│   └── shared/                 # Shared utilities, types
│       ├── src/
│       └── package.json
├── package.json                # Workspace root
└── turbo.json                  # Turborepo config (optional)
```

```json
// Root package.json
{
  "private": true,
  "workspaces": ["apps/*", "packages/*"],
  "scripts": {
    "mobile": "cd apps/mobile && npx expo start",
    "mobile:ios": "cd apps/mobile && npx expo run:ios",
    "mobile:android": "cd apps/mobile && npx expo run:android"
  }
}
```

### Metro Config for Monorepo

```javascript
// apps/mobile/metro.config.js
const { getDefaultConfig } = require('expo/metro-config');
const { FileStore } = require('metro-cache');
const path = require('path');

const projectRoot = __dirname;
const monorepoRoot = path.resolve(projectRoot, '../..');

const config = getDefaultConfig(projectRoot);

// Watch all files in the monorepo
config.watchFolders = [monorepoRoot];

// Let Metro resolve packages from monorepo root
config.resolver.nodeModulesPaths = [
  path.resolve(projectRoot, 'node_modules'),
  path.resolve(monorepoRoot, 'node_modules'),
];

// Force resolving nested modules to the app's node_modules
config.resolver.disableHierarchicalLookup = true;

// Use file-based cache to avoid conflicts between apps
config.cacheStores = [
  new FileStore({ root: path.join(projectRoot, 'node_modules', '.cache', 'metro') }),
];

module.exports = config;
```

---

## Quick Reference: Expo Decision Tree

```
Starting a new project?
├── Prototype/learning → expo init with Expo Go
├── Production app → expo init with dev builds
└── Adding to monorepo → workspace package with metro config

Need native code?
├── Expo module exists → Install from expo ecosystem
├── Third-party RN lib → Use dev builds + config plugins
├── Custom native code → Expo Modules API or config plugins
└── Heavy native work → Consider bare workflow

Deploying?
├── Internal testing → eas build --profile preview
├── Beta testing → eas build --profile production + TestFlight/Internal Track
├── Production → eas build --profile production + eas submit
└── Quick fix → eas update (OTA, no store review)

Managing environments?
├── 2 envs (dev/prod) → APP_VARIANT in eas.json
├── 3+ envs → APP_VARIANT + channel-based EAS Update
└── Feature flags → Use remote config (Firebase, LaunchDarkly)
```
