---
name: electron-development
description: >
  Electron desktop app development — window management, IPC communication,
  menus, system tray, auto-updates, file system, packaging, and security.
  Triggers: "electron", "electron app", "desktop app electron",
  "electron ipc", "electron menu", "electron tray", "electron builder",
  "electron forge", "electron packaging".
  NOT for: lightweight desktop apps (use tauri-development).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Electron Development

## Quick Start

```bash
# Using Electron Forge (recommended)
npm init electron-app@latest my-app -- --template=vite-typescript

# Or with React
npm init electron-app@latest my-app -- --template=vite-typescript
cd my-app
npm install react react-dom
npm install -D @types/react @types/react-dom @vitejs/plugin-react

# Or manual setup
mkdir my-app && cd my-app
npm init -y
npm install electron --save-dev
npm install electron-builder --save-dev
```

## Project Structure

```
my-app/
├── src/
│   ├── main/              # Main process (Node.js)
│   │   ├── main.ts        # Entry point
│   │   ├── ipc.ts         # IPC handlers
│   │   ├── menu.ts        # Application menu
│   │   ├── tray.ts        # System tray
│   │   └── updater.ts     # Auto-updates
│   ├── preload/           # Preload scripts (bridge)
│   │   └── preload.ts     # Context bridge
│   └── renderer/          # Renderer process (Chromium)
│       ├── index.html
│       ├── App.tsx
│       └── main.tsx
├── resources/             # App icons, assets
│   ├── icon.icns          # macOS
│   ├── icon.ico           # Windows
│   └── icon.png           # Linux
├── electron-builder.yml   # Build config
├── forge.config.ts        # Forge config (if using Forge)
└── package.json
```

## Main Process

```typescript
// src/main/main.ts
import { app, BrowserWindow, ipcMain, Menu, Tray, nativeTheme } from "electron";
import path from "path";

let mainWindow: BrowserWindow | null = null;

function createWindow() {
  mainWindow = new BrowserWindow({
    width: 1200,
    height: 800,
    minWidth: 800,
    minHeight: 600,
    title: "My App",
    icon: path.join(__dirname, "../../resources/icon.png"),

    // Window style
    titleBarStyle: "hiddenInset", // macOS: hidden title bar with traffic lights
    // titleBarOverlay: true,     // Windows: overlay controls
    backgroundColor: "#1e1e1e",
    show: false, // Don't show until ready

    webPreferences: {
      preload: path.join(__dirname, "../preload/preload.js"),
      contextIsolation: true,    // REQUIRED for security
      nodeIntegration: false,    // NEVER enable in production
      sandbox: true,             // Additional security
      webSecurity: true,
    },
  });

  // Show when ready (prevents white flash)
  mainWindow.once("ready-to-show", () => {
    mainWindow?.show();
  });

  // Load content
  if (process.env.VITE_DEV_SERVER_URL) {
    mainWindow.loadURL(process.env.VITE_DEV_SERVER_URL);
    mainWindow.webContents.openDevTools();
  } else {
    mainWindow.loadFile(path.join(__dirname, "../renderer/index.html"));
  }

  // Handle external links
  mainWindow.webContents.setWindowOpenHandler(({ url }) => {
    require("electron").shell.openExternal(url);
    return { action: "deny" };
  });

  mainWindow.on("closed", () => {
    mainWindow = null;
  });
}

// App lifecycle
app.whenReady().then(() => {
  createWindow();

  // macOS: re-create window when dock icon clicked
  app.on("activate", () => {
    if (BrowserWindow.getAllWindows().length === 0) {
      createWindow();
    }
  });
});

// Quit when all windows closed (except macOS)
app.on("window-all-closed", () => {
  if (process.platform !== "darwin") {
    app.quit();
  }
});

// Security: prevent new window creation
app.on("web-contents-created", (_, contents) => {
  contents.on("will-navigate", (event) => {
    event.preventDefault();
  });
});
```

## Preload Script (Context Bridge)

```typescript
// src/preload/preload.ts
import { contextBridge, ipcRenderer } from "electron";

// Expose safe APIs to the renderer
contextBridge.exposeInMainWorld("electronAPI", {
  // File operations
  openFile: () => ipcRenderer.invoke("dialog:openFile"),
  saveFile: (content: string) => ipcRenderer.invoke("dialog:saveFile", content),
  readFile: (path: string) => ipcRenderer.invoke("fs:readFile", path),

  // App info
  getVersion: () => ipcRenderer.invoke("app:getVersion"),
  getPlatform: () => process.platform,

  // Window controls
  minimize: () => ipcRenderer.send("window:minimize"),
  maximize: () => ipcRenderer.send("window:maximize"),
  close: () => ipcRenderer.send("window:close"),

  // Theme
  getTheme: () => ipcRenderer.invoke("theme:get"),
  onThemeChange: (callback: (theme: string) => void) => {
    const handler = (_event: any, theme: string) => callback(theme);
    ipcRenderer.on("theme:changed", handler);
    return () => ipcRenderer.removeListener("theme:changed", handler);
  },

  // Notifications
  showNotification: (title: string, body: string) =>
    ipcRenderer.send("notification:show", title, body),
});

// Type declarations for renderer
declare global {
  interface Window {
    electronAPI: {
      openFile: () => Promise<string | null>;
      saveFile: (content: string) => Promise<boolean>;
      readFile: (path: string) => Promise<string>;
      getVersion: () => Promise<string>;
      getPlatform: () => string;
      minimize: () => void;
      maximize: () => void;
      close: () => void;
      getTheme: () => Promise<string>;
      onThemeChange: (callback: (theme: string) => void) => () => void;
      showNotification: (title: string, body: string) => void;
    };
  }
}
```

## IPC Communication

```typescript
// src/main/ipc.ts
import { ipcMain, dialog, app, BrowserWindow, Notification } from "electron";
import fs from "fs/promises";

export function setupIPC() {
  // Invoke pattern (async request-response)
  ipcMain.handle("dialog:openFile", async () => {
    const result = await dialog.showOpenDialog({
      properties: ["openFile"],
      filters: [
        { name: "Text Files", extensions: ["txt", "md", "json"] },
        { name: "All Files", extensions: ["*"] },
      ],
    });

    if (result.canceled || result.filePaths.length === 0) return null;

    const filePath = result.filePaths[0];
    const content = await fs.readFile(filePath, "utf-8");
    return { path: filePath, content };
  });

  ipcMain.handle("dialog:saveFile", async (_event, content: string) => {
    const result = await dialog.showSaveDialog({
      filters: [{ name: "Text Files", extensions: ["txt"] }],
    });

    if (result.canceled || !result.filePath) return false;

    await fs.writeFile(result.filePath, content, "utf-8");
    return true;
  });

  ipcMain.handle("fs:readFile", async (_event, path: string) => {
    return fs.readFile(path, "utf-8");
  });

  ipcMain.handle("app:getVersion", () => app.getVersion());

  ipcMain.handle("theme:get", () => {
    return require("electron").nativeTheme.shouldUseDarkColors ? "dark" : "light";
  });

  // Send pattern (fire-and-forget)
  ipcMain.on("window:minimize", (event) => {
    BrowserWindow.fromWebContents(event.sender)?.minimize();
  });

  ipcMain.on("window:maximize", (event) => {
    const win = BrowserWindow.fromWebContents(event.sender);
    if (win?.isMaximized()) {
      win.unmaximize();
    } else {
      win?.maximize();
    }
  });

  ipcMain.on("window:close", (event) => {
    BrowserWindow.fromWebContents(event.sender)?.close();
  });

  ipcMain.on("notification:show", (_event, title: string, body: string) => {
    new Notification({ title, body }).show();
  });
}
```

## Application Menu

```typescript
// src/main/menu.ts
import { Menu, shell, app, BrowserWindow } from "electron";

export function createMenu() {
  const isMac = process.platform === "darwin";

  const template: Electron.MenuItemConstructorOptions[] = [
    // macOS app menu
    ...(isMac
      ? [
          {
            label: app.name,
            submenu: [
              { role: "about" as const },
              { type: "separator" as const },
              { role: "services" as const },
              { type: "separator" as const },
              { role: "hide" as const },
              { role: "hideOthers" as const },
              { role: "unhide" as const },
              { type: "separator" as const },
              { role: "quit" as const },
            ],
          },
        ]
      : []),

    // File menu
    {
      label: "File",
      submenu: [
        {
          label: "Open File",
          accelerator: "CmdOrCtrl+O",
          click: () => {
            BrowserWindow.getFocusedWindow()?.webContents.send("menu:openFile");
          },
        },
        {
          label: "Save",
          accelerator: "CmdOrCtrl+S",
          click: () => {
            BrowserWindow.getFocusedWindow()?.webContents.send("menu:save");
          },
        },
        { type: "separator" },
        isMac ? { role: "close" } : { role: "quit" },
      ],
    },

    // Edit menu
    {
      label: "Edit",
      submenu: [
        { role: "undo" },
        { role: "redo" },
        { type: "separator" },
        { role: "cut" },
        { role: "copy" },
        { role: "paste" },
        { role: "selectAll" },
      ],
    },

    // View menu
    {
      label: "View",
      submenu: [
        { role: "reload" },
        { role: "forceReload" },
        { role: "toggleDevTools" },
        { type: "separator" },
        { role: "resetZoom" },
        { role: "zoomIn" },
        { role: "zoomOut" },
        { type: "separator" },
        { role: "togglefullscreen" },
      ],
    },

    // Help menu
    {
      label: "Help",
      submenu: [
        {
          label: "Documentation",
          click: () => shell.openExternal("https://myapp.com/docs"),
        },
        {
          label: "Report Issue",
          click: () => shell.openExternal("https://github.com/myapp/issues"),
        },
      ],
    },
  ];

  const menu = Menu.buildFromTemplate(template);
  Menu.setApplicationMenu(menu);
}
```

## System Tray

```typescript
// src/main/tray.ts
import { Tray, Menu, nativeImage, app, BrowserWindow } from "electron";
import path from "path";

let tray: Tray | null = null;

export function createTray(mainWindow: BrowserWindow) {
  const icon = nativeImage.createFromPath(
    path.join(__dirname, "../../resources/tray-icon.png")
  );

  // macOS: resize for menu bar (16x16 or 22x22)
  const resized = icon.resize({ width: 16, height: 16 });
  resized.setTemplateImage(true); // macOS: adapts to dark/light menu bar

  tray = new Tray(resized);
  tray.setToolTip("My App");

  const contextMenu = Menu.buildFromTemplate([
    {
      label: "Show App",
      click: () => {
        mainWindow.show();
        mainWindow.focus();
      },
    },
    {
      label: "Status",
      submenu: [
        { label: "Online", type: "radio", checked: true },
        { label: "Away", type: "radio" },
        { label: "Do Not Disturb", type: "radio" },
      ],
    },
    { type: "separator" },
    {
      label: "Quit",
      click: () => {
        app.quit();
      },
    },
  ]);

  tray.setContextMenu(contextMenu);

  // Click to show/hide window
  tray.on("click", () => {
    if (mainWindow.isVisible()) {
      mainWindow.hide();
    } else {
      mainWindow.show();
      mainWindow.focus();
    }
  });
}
```

## Auto-Updates

```typescript
// src/main/updater.ts
import { autoUpdater } from "electron-updater";
import { BrowserWindow, dialog } from "electron";
import log from "electron-log";

export function setupAutoUpdater(mainWindow: BrowserWindow) {
  autoUpdater.logger = log;
  autoUpdater.autoDownload = false;
  autoUpdater.autoInstallOnAppQuit = true;

  autoUpdater.on("update-available", (info) => {
    dialog
      .showMessageBox(mainWindow, {
        type: "info",
        title: "Update Available",
        message: `Version ${info.version} is available. Download now?`,
        buttons: ["Download", "Later"],
      })
      .then((result) => {
        if (result.response === 0) {
          autoUpdater.downloadUpdate();
        }
      });
  });

  autoUpdater.on("download-progress", (progress) => {
    mainWindow.webContents.send("update:progress", progress.percent);
    mainWindow.setProgressBar(progress.percent / 100);
  });

  autoUpdater.on("update-downloaded", () => {
    dialog
      .showMessageBox(mainWindow, {
        type: "info",
        title: "Update Ready",
        message: "Update downloaded. Restart now to install?",
        buttons: ["Restart", "Later"],
      })
      .then((result) => {
        if (result.response === 0) {
          autoUpdater.quitAndInstall();
        }
      });
  });

  autoUpdater.on("error", (error) => {
    log.error("Update error:", error);
  });

  // Check for updates on startup
  autoUpdater.checkForUpdates();

  // Check periodically (every 4 hours)
  setInterval(() => autoUpdater.checkForUpdates(), 4 * 60 * 60 * 1000);
}
```

## Packaging & Distribution

```yaml
# electron-builder.yml
appId: com.mycompany.myapp
productName: My App
copyright: Copyright © 2026 My Company

directories:
  output: dist
  buildResources: resources

files:
  - "!**/.vscode/*"
  - "!src/*"
  - "!electron.vite.config.*"
  - "!{.eslintignore,.eslintrc.cjs,.prettierignore,.prettierrc.yaml}"
  - "!{tsconfig.json,tsconfig.node.json,tsconfig.web.json}"

mac:
  target:
    - target: dmg
      arch: [universal]  # Intel + Apple Silicon
    - target: zip
      arch: [universal]
  category: public.app-category.productivity
  icon: resources/icon.icns
  hardenedRuntime: true
  gatekeeperAssess: false
  entitlements: build/entitlements.mac.plist
  entitlementsInherit: build/entitlements.mac.plist
  notarize:
    teamId: YOUR_TEAM_ID

win:
  target:
    - target: nsis
      arch: [x64, arm64]
  icon: resources/icon.ico
  certificateFile: ${env.WIN_CSC_LINK}
  certificatePassword: ${env.WIN_CSC_KEY_PASSWORD}

nsis:
  oneClick: false
  perMachine: false
  allowToChangeInstallationDirectory: true
  createDesktopShortcut: true

linux:
  target:
    - target: AppImage
      arch: [x64]
    - target: deb
      arch: [x64]
  category: Utility
  icon: resources/icon.png

publish:
  provider: github
  owner: mycompany
  repo: myapp
```

```bash
# Build commands
npx electron-builder --mac       # macOS
npx electron-builder --win       # Windows
npx electron-builder --linux     # Linux
npx electron-builder --mac --win --linux  # All platforms

# With Electron Forge
npx electron-forge make          # Build distributable
npx electron-forge publish       # Build + publish
```

## Data Storage

```typescript
// Using electron-store for persistent settings
import Store from "electron-store";

const store = new Store({
  defaults: {
    windowBounds: { width: 1200, height: 800 },
    theme: "system",
    recentFiles: [],
  },
  schema: {
    theme: { type: "string", enum: ["light", "dark", "system"] },
    recentFiles: {
      type: "array",
      items: { type: "string" },
      maxItems: 10,
    },
  },
});

// Usage
store.set("theme", "dark");
const theme = store.get("theme"); // "dark"

// Watch for changes
store.onDidChange("theme", (newValue) => {
  mainWindow.webContents.send("theme:changed", newValue);
});

// Save window position
mainWindow.on("close", () => {
  store.set("windowBounds", mainWindow.getBounds());
});
```

## Gotchas

1. **Never enable `nodeIntegration: true`** — it gives the renderer full Node.js access, which is a massive security vulnerability. Always use `contextBridge` in a preload script with `contextIsolation: true`.

2. **`__dirname` is different in production** — in development `__dirname` is your source directory, but in production it's inside the asar archive. Use `app.getPath("userData")` for writable paths and `app.isPackaged` to detect environment.

3. **macOS code signing is required for distribution** — unsigned apps trigger Gatekeeper warnings. You need an Apple Developer account ($99/year) and must notarize the app. Without it, users get "app is damaged" errors on macOS 15+.

4. **`electron-builder` vs `electron-forge`** — Forge is Electron's official tool with better plugin support. Builder has more configuration options and a larger community. Pick one and stick with it.

5. **Don't forget to handle the second-instance event** — without `app.requestSingleInstanceLock()`, users can open multiple instances. Handle the `second-instance` event to focus the existing window instead.

6. **Auto-updater needs code signing** — `electron-updater` verifies update signatures. Unsigned apps can't auto-update securely. Set up code signing before implementing auto-updates.
