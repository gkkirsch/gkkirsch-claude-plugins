# Desktop App Development Cheatsheet

## Electron Quick Reference

### Setup
```bash
npm init electron-app@latest my-app -- --template=vite-typescript
cd my-app && npm start
```

### Architecture
```
Main Process (Node.js) ←→ IPC ←→ Renderer Process (Chromium)
                        ↑
                   Preload Script
                  (contextBridge)
```

### Window Creation
```typescript
const win = new BrowserWindow({
  width: 1200, height: 800,
  webPreferences: {
    preload: path.join(__dirname, "preload.js"),
    contextIsolation: true,   // Always true
    nodeIntegration: false,   // Always false
  },
});
win.loadFile("index.html");
// or win.loadURL("http://localhost:5173");
```

### IPC (Main ↔ Renderer)
```typescript
// Main process
ipcMain.handle("channel", async (event, arg) => {
  return result; // Returned to renderer
});

// Preload (bridge)
contextBridge.exposeInMainWorld("api", {
  doThing: (arg) => ipcRenderer.invoke("channel", arg),
});

// Renderer
const result = await window.api.doThing(arg);
```

### Key APIs
| API | Purpose |
|-----|---------|
| `dialog.showOpenDialog()` | File picker |
| `dialog.showSaveDialog()` | Save dialog |
| `shell.openExternal(url)` | Open in browser |
| `Notification` | System notification |
| `Tray` | System tray icon |
| `Menu` | App/context menus |
| `autoUpdater` | Auto-updates |
| `nativeTheme` | Dark/light mode |
| `app.getPath("userData")` | User data directory |

### Build
```bash
npx electron-builder --mac --win --linux
# Or with Forge:
npx electron-forge make
```

---

## Tauri 2.0 Quick Reference

### Setup
```bash
npm create tauri-app@latest my-app
cd my-app
npm run tauri dev
```

### Architecture
```
Rust Core ←→ Commands/Events ←→ System Webview (frontend)
           ↑
      Plugins (fs, dialog, etc.)
```

### Commands
```rust
// Rust (src-tauri/src/lib.rs)
#[tauri::command]
fn greet(name: &str) -> String {
    format!("Hello, {}!", name)
}

// Register:
.invoke_handler(tauri::generate_handler![greet])
```

```typescript
// Frontend
import { invoke } from "@tauri-apps/api/core";
const result = await invoke<string>("greet", { name: "World" });
```

### Events
```rust
// Rust → Frontend
window.emit("event-name", payload).unwrap();
```

```typescript
// Frontend
import { listen, emit } from "@tauri-apps/api/event";
const unlisten = await listen("event-name", (e) => console.log(e.payload));
await emit("to-rust", { data: "hello" }); // Frontend → Rust
```

### Key Plugins
| Plugin | npm Package | Capability |
|--------|-------------|------------|
| File System | `@tauri-apps/plugin-fs` | `fs:default` |
| Dialog | `@tauri-apps/plugin-dialog` | `dialog:default` |
| Shell | `@tauri-apps/plugin-shell` | `shell:default` |
| HTTP | `@tauri-apps/plugin-http` | `http:default` |
| Store | `@tauri-apps/plugin-store` | `store:default` |
| Updater | `@tauri-apps/plugin-updater` | `updater:default` |
| Notification | `@tauri-apps/plugin-notification` | `notification:default` |
| SQL | `@tauri-apps/plugin-sql` | — |

### Capabilities
```json
// src-tauri/capabilities/default.json
{
  "identifier": "default",
  "windows": ["main"],
  "permissions": ["core:default", "fs:default", "dialog:default"]
}
```

### Build
```bash
npm run tauri build
# Outputs: .dmg/.app (mac), .msi/.exe (win), .deb/.AppImage (linux)
npx tauri icon app-icon.png  # Generate all icon sizes
```

---

## Quick Decision

| Need | Choose |
|------|--------|
| Node.js npm packages in backend | **Electron** |
| Smallest possible binary (2-10MB) | **Tauri** |
| Lowest memory usage | **Tauri** |
| Mobile support (iOS/Android) | **Tauri 2.0** |
| Team only knows JavaScript | **Electron** |
| Maximum security | **Tauri** |
| Pixel-perfect cross-platform | **Electron** |
| Fastest time to ship | **Electron** |

## Code Signing

### macOS
```bash
# Requires Apple Developer account ($99/year)
# Set in electron-builder.yml or tauri.conf.json:
# - signingIdentity / macOS.signingIdentity
# Notarize: xcrun notarytool submit app.dmg --apple-id ... --team-id ...
```

### Windows
```bash
# Requires code signing certificate
# Options: DigiCert, Sectigo (~$200-400/year)
# Or use Azure Trusted Signing (cheaper)
```

### Linux
```bash
# No code signing required for distribution
# AppImage, .deb, .rpm, Flatpak, Snap
```
