---
name: desktop-app-packaging
description: >
  Desktop application packaging, distribution, and auto-update patterns.
  Use when building installers, configuring code signing, implementing
  auto-updates, or distributing Electron or Tauri apps.
  Triggers: "desktop packaging", "electron builder", "tauri build", "code signing",
  "auto update", "app installer", "dmg", "nsis", "notarization", "app distribution".
  NOT for: Electron development (see electron-development), Tauri development (see tauri-development), web deployment.
version: 1.0.0
allowed-tools: Read, Grep, Glob, Bash
---

# Desktop App Packaging

## Electron Builder Configuration

```json
// electron-builder.yml or package.json "build" key
{
  "appId": "com.mycompany.myapp",
  "productName": "My App",
  "copyright": "Copyright 2026 My Company",
  "directories": {
    "output": "dist",
    "buildResources": "build"
  },
  "files": [
    "out/**/*",
    "package.json"
  ],
  "asar": true,
  "asarUnpack": [
    "**/*.node",
    "**/node_modules/sharp/**"
  ],

  "mac": {
    "category": "public.app-category.productivity",
    "icon": "build/icon.icns",
    "hardenedRuntime": true,
    "gatekeeperAssess": false,
    "entitlements": "build/entitlements.mac.plist",
    "entitlementsInherit": "build/entitlements.mac.plist",
    "target": [
      { "target": "dmg", "arch": ["x64", "arm64"] },
      { "target": "zip", "arch": ["x64", "arm64"] }
    ],
    "notarize": {
      "teamId": "TEAM_ID_HERE"
    }
  },

  "win": {
    "icon": "build/icon.ico",
    "target": [
      { "target": "nsis", "arch": ["x64", "arm64"] }
    ],
    "certificateSubjectName": "My Company LLC",
    "signingHashAlgorithms": ["sha256"],
    "sign": "./scripts/sign.js"
  },

  "nsis": {
    "oneClick": false,
    "perMachine": false,
    "allowToChangeInstallationDirectory": true,
    "deleteAppDataOnUninstall": false,
    "createDesktopShortcut": true,
    "createStartMenuShortcut": true,
    "shortcutName": "My App"
  },

  "linux": {
    "icon": "build/icons",
    "category": "Utility",
    "target": [
      { "target": "AppImage", "arch": ["x64"] },
      { "target": "deb", "arch": ["x64"] },
      { "target": "rpm", "arch": ["x64"] }
    ],
    "desktop": {
      "StartupWMClass": "my-app"
    }
  },

  "publish": {
    "provider": "github",
    "owner": "mycompany",
    "repo": "my-app",
    "releaseType": "release"
  }
}
```

## macOS Code Signing & Notarization

```bash
#!/bin/bash
# scripts/notarize.sh — macOS notarization workflow

# Prerequisites:
# 1. Apple Developer account ($99/year)
# 2. Developer ID Application certificate in Keychain
# 3. App-specific password for notarization

# Environment variables needed:
# APPLE_ID=developer@example.com
# APPLE_APP_SPECIFIC_PASSWORD=xxxx-xxxx-xxxx-xxxx
# APPLE_TEAM_ID=ABCDE12345

# Step 1: Build and sign
npx electron-builder --mac --publish never

# Step 2: Verify code signing
codesign --verify --deep --strict "dist/mac-arm64/My App.app"
codesign -dv --verbose=4 "dist/mac-arm64/My App.app"

# Step 3: Notarize (electron-builder does this automatically with notarize config)
# Manual notarization if needed:
xcrun notarytool submit "dist/My App-1.0.0-arm64.dmg" \
  --apple-id "$APPLE_ID" \
  --team-id "$APPLE_TEAM_ID" \
  --password "$APPLE_APP_SPECIFIC_PASSWORD" \
  --wait

# Step 4: Staple the notarization ticket
xcrun stapler staple "dist/My App-1.0.0-arm64.dmg"

# Step 5: Verify notarization
spctl --assess --type execute --verbose "dist/mac-arm64/My App.app"
```

```xml
<!-- build/entitlements.mac.plist -->
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>com.apple.security.cs.allow-jit</key>
    <true/>
    <key>com.apple.security.cs.allow-unsigned-executable-memory</key>
    <true/>
    <key>com.apple.security.cs.allow-dyld-environment-variables</key>
    <true/>
    <key>com.apple.security.network.client</key>
    <true/>
    <key>com.apple.security.files.user-selected.read-write</key>
    <true/>
</dict>
</plist>
```

## Auto-Update (Electron)

```typescript
// main/auto-updater.ts
import { autoUpdater } from 'electron-updater';
import { BrowserWindow, dialog } from 'electron';
import log from 'electron-log';

export function initAutoUpdater(mainWindow: BrowserWindow): void {
  // Configure logging
  autoUpdater.logger = log;
  autoUpdater.autoDownload = false; // Don't download without user consent
  autoUpdater.autoInstallOnAppQuit = true;

  // Check for updates periodically (every 4 hours)
  setInterval(() => autoUpdater.checkForUpdates(), 4 * 60 * 60 * 1000);
  autoUpdater.checkForUpdates(); // Check on startup

  autoUpdater.on('update-available', (info) => {
    // Notify renderer
    mainWindow.webContents.send('update-available', {
      version: info.version,
      releaseNotes: info.releaseNotes,
      releaseDate: info.releaseDate,
    });
  });

  autoUpdater.on('download-progress', (progress) => {
    mainWindow.webContents.send('update-progress', {
      percent: Math.round(progress.percent),
      transferred: progress.transferred,
      total: progress.total,
      bytesPerSecond: progress.bytesPerSecond,
    });
  });

  autoUpdater.on('update-downloaded', (info) => {
    mainWindow.webContents.send('update-downloaded', { version: info.version });

    // Option: show native dialog
    dialog.showMessageBox(mainWindow, {
      type: 'info',
      title: 'Update Ready',
      message: `Version ${info.version} has been downloaded.`,
      detail: 'The update will be installed when you restart the application.',
      buttons: ['Restart Now', 'Later'],
    }).then(({ response }) => {
      if (response === 0) {
        autoUpdater.quitAndInstall(false, true); // isSilent=false, isForceRunAfter=true
      }
    });
  });

  autoUpdater.on('error', (error) => {
    log.error('Auto-updater error:', error);
    mainWindow.webContents.send('update-error', { message: error.message });
  });
}

// IPC handlers for renderer-initiated actions
import { ipcMain } from 'electron';
ipcMain.handle('check-for-update', () => autoUpdater.checkForUpdates());
ipcMain.handle('download-update', () => autoUpdater.downloadUpdate());
ipcMain.handle('install-update', () => autoUpdater.quitAndInstall(false, true));
```

## Tauri Build Configuration

```json
// src-tauri/tauri.conf.json (Tauri v2)
{
  "productName": "My App",
  "version": "1.0.0",
  "identifier": "com.mycompany.myapp",
  "build": {
    "beforeBuildCommand": "npm run build",
    "beforeDevCommand": "npm run dev",
    "frontendDist": "../dist",
    "devUrl": "http://localhost:5173"
  },
  "bundle": {
    "active": true,
    "targets": "all",
    "icon": [
      "icons/32x32.png",
      "icons/128x128.png",
      "icons/128x128@2x.png",
      "icons/icon.icns",
      "icons/icon.ico"
    ],
    "macOS": {
      "minimumSystemVersion": "10.15",
      "signingIdentity": "-",
      "frameworks": []
    },
    "windows": {
      "certificateThumbprint": null,
      "digestAlgorithm": "sha256",
      "wix": {
        "language": "en-US"
      }
    },
    "linux": {
      "deb": { "depends": ["libwebkit2gtk-4.1-0"] },
      "appimage": { "bundleMediaFramework": true }
    }
  },
  "plugins": {
    "updater": {
      "pubkey": "YOUR_PUBLIC_KEY_HERE",
      "endpoints": [
        "https://releases.myapp.com/{{target}}/{{arch}}/{{current_version}}"
      ],
      "windows": {
        "installMode": "passive"
      }
    }
  }
}
```

## CI/CD Build Pipeline

```yaml
# .github/workflows/release.yml
name: Release
on:
  push:
    tags: ['v*']

jobs:
  build:
    strategy:
      matrix:
        include:
          - os: macos-latest
            target: mac
          - os: windows-latest
            target: win
          - os: ubuntu-latest
            target: linux
    runs-on: ${{ matrix.os }}

    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: 20
          cache: npm

      - run: npm ci

      # macOS: import signing certificate
      - name: Import macOS certificate
        if: matrix.target == 'mac'
        env:
          CERTIFICATE_BASE64: ${{ secrets.MAC_CERTIFICATE_BASE64 }}
          CERTIFICATE_PASSWORD: ${{ secrets.MAC_CERTIFICATE_PASSWORD }}
        run: |
          echo "$CERTIFICATE_BASE64" | base64 --decode > certificate.p12
          security create-keychain -p "" build.keychain
          security import certificate.p12 -k build.keychain -P "$CERTIFICATE_PASSWORD" -T /usr/bin/codesign
          security set-keychain-settings build.keychain
          security list-keychains -d user -s build.keychain login.keychain

      # Build
      - name: Build Electron app
        env:
          GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          APPLE_ID: ${{ secrets.APPLE_ID }}
          APPLE_APP_SPECIFIC_PASSWORD: ${{ secrets.APPLE_APP_SPECIFIC_PASSWORD }}
          APPLE_TEAM_ID: ${{ secrets.APPLE_TEAM_ID }}
        run: npx electron-builder --${{ matrix.target }} --publish always

      # Upload artifacts
      - uses: actions/upload-artifact@v4
        with:
          name: release-${{ matrix.target }}
          path: dist/*.{dmg,exe,AppImage,deb,rpm,zip}
```

## Gotchas

1. **Unsigned apps are blocked by default** -- macOS Gatekeeper and Windows SmartScreen block unsigned apps. On macOS, unsigned apps can't be opened without right-click > Open workaround. On Windows, SmartScreen shows a scary warning. Code signing is not optional for production distribution. Budget $99/year for Apple Developer and ~$200-400/year for a Windows code signing certificate.

2. **Notarization requires hardened runtime** -- macOS notarization fails without `hardenedRuntime: true` in electron-builder config. But hardened runtime breaks JIT compilation by default, which Electron needs. You must add `com.apple.security.cs.allow-jit` and `com.apple.security.cs.allow-unsigned-executable-memory` to your entitlements plist.

3. **Auto-update on macOS requires code signing** -- `electron-updater` and Tauri's updater both verify the code signature of downloaded updates. If your current build is unsigned, the update check may succeed but installation will fail. The update binary must be signed by the same certificate as the installed app.

4. **ASAR unpacking for native modules** -- Native Node.js modules (`.node` files) and some binaries can't run from inside an ASAR archive. Use `asarUnpack` for `**/*.node` and any native dependencies like sharp, sqlite3, or bcrypt. Forgetting this causes "module not found" errors only in production builds, not in development.

5. **Windows installer per-user vs per-machine** -- NSIS `perMachine: true` requires admin elevation (UAC prompt). Per-user installs don't need elevation but install to `%LOCALAPPDATA%` instead of `Program Files`. If your app needs system-wide services or PATH modifications, use per-machine. For most apps, per-user is less friction.

6. **Auto-update differential downloads** -- electron-updater supports differential (delta) updates via `blockmap` files, reducing download size by ~60-80%. But this only works with the `zip` target on macOS and `nsis` on Windows. DMG, AppImage, and other formats always download the full binary. Enable `"differentialDownload": true` in publish config.
