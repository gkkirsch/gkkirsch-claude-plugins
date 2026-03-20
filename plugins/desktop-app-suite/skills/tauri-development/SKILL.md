---
name: tauri-development
description: >
  Tauri 2.0 desktop and mobile app development — commands, events, plugins,
  window management, file system, system tray, auto-updates, and distribution.
  Triggers: "tauri", "tauri app", "tauri 2", "tauri desktop",
  "tauri commands", "tauri plugins", "tauri mobile",
  "tauri build", "tauri packaging".
  NOT for: apps needing Node.js backend (use electron-development).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Tauri 2.0 Development

## Prerequisites

```bash
# macOS
xcode-select --install
brew install rustup
rustup-init

# Windows
# Install Build Tools for Visual Studio 2022
# Install Rust via rustup.rs
# Install WebView2 Runtime (usually pre-installed on Win 10+)

# Linux (Ubuntu/Debian)
sudo apt install libwebkit2gtk-4.1-dev build-essential curl wget file \
  libxdo-dev libssl-dev libayatana-appindicator3-dev librsvg2-dev
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
```

## Quick Start

```bash
# Create new project with a web framework
npm create tauri-app@latest my-app
# Choose: React/Vue/Svelte/Solid/Vanilla + TypeScript

# Or add Tauri to an existing web project
cd existing-project
npm install @tauri-apps/cli@latest
npx tauri init

# Development
npm run tauri dev

# Build
npm run tauri build
```

## Project Structure

```
my-app/
├── src/                   # Frontend (React/Vue/Svelte/etc.)
│   ├── App.tsx
│   └── main.tsx
├── src-tauri/             # Rust backend
│   ├── src/
│   │   ├── main.rs        # Entry point
│   │   ├── lib.rs         # Commands and setup
│   │   └── commands/      # Custom commands
│   │       ├── mod.rs
│   │       └── files.rs
│   ├── capabilities/      # Permission configs
│   │   └── default.json
│   ├── icons/             # App icons (auto-generated)
│   ├── Cargo.toml         # Rust dependencies
│   └── tauri.conf.json    # Tauri configuration
├── package.json
└── vite.config.ts
```

## Configuration

```json
// src-tauri/tauri.conf.json
{
  "$schema": "https://raw.githubusercontent.com/tauri-apps/tauri/dev/crates/tauri-cli/schema.json",
  "productName": "My App",
  "version": "1.0.0",
  "identifier": "com.mycompany.myapp",
  "build": {
    "frontendDist": "../dist",
    "devUrl": "http://localhost:5173",
    "beforeDevCommand": "npm run dev",
    "beforeBuildCommand": "npm run build"
  },
  "app": {
    "windows": [
      {
        "title": "My App",
        "width": 1200,
        "height": 800,
        "minWidth": 800,
        "minHeight": 600,
        "resizable": true,
        "fullscreen": false,
        "decorations": true,
        "transparent": false
      }
    ],
    "security": {
      "csp": "default-src 'self'; img-src 'self' https:; style-src 'self' 'unsafe-inline'"
    },
    "trayIcon": {
      "iconPath": "icons/tray-icon.png",
      "tooltip": "My App"
    }
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
      "signingIdentity": null,
      "providerShortName": null
    },
    "windows": {
      "certificateThumbprint": null,
      "digestAlgorithm": "sha256"
    }
  }
}
```

## Commands (Rust → Frontend Communication)

```rust
// src-tauri/src/lib.rs
use tauri::Manager;

// Simple command
#[tauri::command]
fn greet(name: &str) -> String {
    format!("Hello, {}!", name)
}

// Async command with error handling
#[tauri::command]
async fn read_file(path: String) -> Result<String, String> {
    std::fs::read_to_string(&path)
        .map_err(|e| format!("Failed to read file: {}", e))
}

// Command with state
#[tauri::command]
fn get_count(state: tauri::State<'_, AppState>) -> i32 {
    let count = state.count.lock().unwrap();
    *count
}

#[tauri::command]
fn increment(state: tauri::State<'_, AppState>) -> i32 {
    let mut count = state.count.lock().unwrap();
    *count += 1;
    *count
}

// Managed state
struct AppState {
    count: std::sync::Mutex<i32>,
    db: std::sync::Mutex<Option<Database>>,
}

// Register commands
pub fn run() {
    tauri::Builder::default()
        .manage(AppState {
            count: std::sync::Mutex::new(0),
            db: std::sync::Mutex::new(None),
        })
        .invoke_handler(tauri::generate_handler![
            greet,
            read_file,
            get_count,
            increment,
        ])
        .setup(|app| {
            // Initialization code runs once at startup
            let window = app.get_webview_window("main").unwrap();
            println!("Window label: {}", window.label());
            Ok(())
        })
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
```

```typescript
// Frontend: invoke commands
import { invoke } from "@tauri-apps/api/core";

// Simple call
const greeting = await invoke<string>("greet", { name: "World" });

// With error handling
try {
  const content = await invoke<string>("read_file", { path: "/tmp/test.txt" });
  console.log(content);
} catch (error) {
  console.error("Error:", error); // Error string from Rust Err()
}

// State commands
const count = await invoke<number>("increment");
```

## Events (Bidirectional)

```rust
// src-tauri/src/lib.rs
use tauri::Emitter;

// Emit event from Rust to frontend
#[tauri::command]
fn start_processing(window: tauri::Window) {
    std::thread::spawn(move || {
        for i in 0..100 {
            window.emit("progress", i).unwrap();
            std::thread::sleep(std::time::Duration::from_millis(50));
        }
        window.emit("complete", "Done!").unwrap();
    });
}

// Listen to frontend events in Rust
fn setup(app: &mut tauri::App) -> Result<(), Box<dyn std::error::Error>> {
    let handle = app.handle().clone();
    app.listen("frontend-event", move |event| {
        println!("Got event: {:?}", event.payload());
    });
    Ok(())
}
```

```typescript
// Frontend: listen and emit events
import { listen, emit } from "@tauri-apps/api/event";

// Listen to events from Rust
const unlisten = await listen<number>("progress", (event) => {
  console.log(`Progress: ${event.payload}%`);
});

// Listen to completion
await listen<string>("complete", (event) => {
  console.log(event.payload);
});

// Emit event to Rust
await emit("frontend-event", { message: "Hello from frontend" });

// Cleanup listener when component unmounts
unlisten();
```

## Plugins (Tauri 2.0)

```bash
# Official plugins
npm install @tauri-apps/plugin-fs        # File system
npm install @tauri-apps/plugin-dialog    # File/save dialogs
npm install @tauri-apps/plugin-shell     # Shell commands
npm install @tauri-apps/plugin-http      # HTTP client
npm install @tauri-apps/plugin-store     # Persistent key-value store
npm install @tauri-apps/plugin-updater   # Auto-updates
npm install @tauri-apps/plugin-notification  # System notifications
npm install @tauri-apps/plugin-clipboard-manager  # Clipboard
npm install @tauri-apps/plugin-os        # OS info
npm install @tauri-apps/plugin-process   # Process control
npm install @tauri-apps/plugin-sql       # SQLite/MySQL/PostgreSQL

# Also add Rust dependencies in Cargo.toml:
# [dependencies]
# tauri-plugin-fs = "2"
# tauri-plugin-dialog = "2"
# tauri-plugin-shell = "2"
# etc.
```

```rust
// Register plugins in lib.rs
pub fn run() {
    tauri::Builder::default()
        .plugin(tauri_plugin_fs::init())
        .plugin(tauri_plugin_dialog::init())
        .plugin(tauri_plugin_shell::init())
        .plugin(tauri_plugin_store::Builder::default().build())
        .plugin(tauri_plugin_updater::Builder::default().build())
        .invoke_handler(tauri::generate_handler![/* commands */])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
```

```typescript
// Frontend: use plugins
import { open, save } from "@tauri-apps/plugin-dialog";
import { readTextFile, writeTextFile, BaseDirectory } from "@tauri-apps/plugin-fs";
import { Store } from "@tauri-apps/plugin-store";
import { check } from "@tauri-apps/plugin-updater";

// File dialog
const filePath = await open({
  filters: [{ name: "Text", extensions: ["txt", "md"] }],
  multiple: false,
});

if (filePath) {
  const content = await readTextFile(filePath);
  console.log(content);
}

// Save dialog
const savePath = await save({
  filters: [{ name: "Text", extensions: ["txt"] }],
});
if (savePath) {
  await writeTextFile(savePath, "Hello, world!");
}

// App data directory
await writeTextFile("config.json", JSON.stringify(config), {
  baseDir: BaseDirectory.AppData,
});

// Persistent store (like electron-store)
const store = await Store.load("settings.json");
await store.set("theme", "dark");
const theme = await store.get<string>("theme");
await store.save(); // Persist to disk

// Auto-update check
const update = await check();
if (update) {
  console.log(`Update ${update.version} available`);
  await update.downloadAndInstall();
}
```

## Capabilities (Permissions)

```json
// src-tauri/capabilities/default.json
{
  "$schema": "https://raw.githubusercontent.com/tauri-apps/tauri/dev/crates/tauri-utils/schema.json",
  "identifier": "default",
  "description": "Default capabilities for the main window",
  "windows": ["main"],
  "permissions": [
    "core:default",
    "dialog:default",
    {
      "identifier": "fs:default",
      "allow": [
        {
          "path": "$APPDATA/**"
        },
        {
          "path": "$DOCUMENT/**"
        }
      ]
    },
    "shell:default",
    "store:default",
    "notification:default",
    "clipboard-manager:default",
    "updater:default"
  ]
}
```

## Window Management

```rust
// Create multiple windows
use tauri::{WebviewUrl, WebviewWindowBuilder};

#[tauri::command]
async fn open_settings(app: tauri::AppHandle) -> Result<(), String> {
    let _settings_window = WebviewWindowBuilder::new(
        &app,
        "settings",
        WebviewUrl::App("/settings".into()),
    )
    .title("Settings")
    .inner_size(600.0, 400.0)
    .resizable(false)
    .build()
    .map_err(|e| e.to_string())?;

    Ok(())
}
```

```typescript
// Frontend: window operations
import { getCurrentWebviewWindow } from "@tauri-apps/api/webviewWindow";
import { WebviewWindow } from "@tauri-apps/api/webviewWindow";

const appWindow = getCurrentWebviewWindow();

// Window controls
await appWindow.minimize();
await appWindow.toggleMaximize();
await appWindow.close();
await appWindow.setTitle("New Title");
await appWindow.setFullscreen(true);

// Create new window
const settingsWindow = new WebviewWindow("settings", {
  url: "/settings",
  title: "Settings",
  width: 600,
  height: 400,
});
```

## System Tray

```rust
// src-tauri/src/lib.rs
use tauri::{
    menu::{Menu, MenuItem},
    tray::{MouseButton, MouseButtonState, TrayIconBuilder, TrayIconEvent},
    Manager,
};

pub fn run() {
    tauri::Builder::default()
        .setup(|app| {
            let quit = MenuItem::with_id(app, "quit", "Quit", true, None::<&str>)?;
            let show = MenuItem::with_id(app, "show", "Show", true, None::<&str>)?;
            let menu = Menu::with_items(app, &[&show, &quit])?;

            let _tray = TrayIconBuilder::new()
                .icon(app.default_window_icon().unwrap().clone())
                .menu(&menu)
                .on_menu_event(|app, event| match event.id.as_ref() {
                    "quit" => app.exit(0),
                    "show" => {
                        if let Some(window) = app.get_webview_window("main") {
                            window.show().unwrap();
                            window.set_focus().unwrap();
                        }
                    }
                    _ => {}
                })
                .on_tray_icon_event(|tray, event| {
                    if let TrayIconEvent::Click {
                        button: MouseButton::Left,
                        button_state: MouseButtonState::Up,
                        ..
                    } = event
                    {
                        let app = tray.app_handle();
                        if let Some(window) = app.get_webview_window("main") {
                            window.show().unwrap();
                            window.set_focus().unwrap();
                        }
                    }
                })
                .build(app)?;

            Ok(())
        })
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
```

## Building & Distribution

```bash
# Build for current platform
npm run tauri build

# macOS: produces .dmg and .app in src-tauri/target/release/bundle/
# Windows: produces .msi and .exe in src-tauri/target/release/bundle/
# Linux: produces .deb, .rpm, .AppImage

# Cross-compilation (requires targets installed)
rustup target add aarch64-apple-darwin  # Apple Silicon
npm run tauri build -- --target aarch64-apple-darwin

# Universal macOS binary (Intel + Apple Silicon)
npm run tauri build -- --target universal-apple-darwin

# Debug build (faster, larger)
npm run tauri build -- --debug

# Generate icons from a single 1024x1024 PNG
npx tauri icon src-tauri/app-icon.png
```

## Gotchas

1. **Tauri 2.0 uses a capability system for permissions** — plugins don't work until you add their permissions to `src-tauri/capabilities/default.json`. "Permission denied" errors usually mean a missing capability, not a code bug.

2. **Rust commands must be registered** — adding a `#[tauri::command]` function isn't enough. You must add it to `tauri::generate_handler![command_name]` in the builder. Missing commands cause "command not found" errors from `invoke()`.

3. **System webview differences** — Tauri uses the OS webview (WebView2 on Windows, WebKit on macOS/Linux), not bundled Chromium. CSS/JS behavior can differ slightly across platforms. Test on all target platforms.

4. **Frontend dev server must be running** — `npm run tauri dev` expects the frontend dev server at the URL in `tauri.conf.json`. If Vite isn't running on port 5173, the app shows a blank window. The `beforeDevCommand` config automates this.

5. **Rust compile times** — first build downloads and compiles all Rust dependencies (several minutes). Subsequent builds are incremental and fast. Use `cargo install sccache` to cache compilations across projects.

6. **Mobile requires additional setup** — Tauri 2.0 supports iOS and Android, but you need Xcode (iOS) or Android Studio (Android) installed. Run `npx tauri android init` or `npx tauri ios init` to set up mobile targets.
