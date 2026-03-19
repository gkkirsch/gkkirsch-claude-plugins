---
name: mobile-dev
description: >
  Quick mobile development command — analyzes your React Native or Expo project and helps architect, build, test, or optimize
  mobile applications. Routes to the appropriate specialist agent based on your request.
  Triggers: "/mobile-dev", "react native", "expo app", "mobile ui", "mobile testing",
  "mobile performance", "expo router", "native module", "mobile animation", "detox test".
user-invocable: true
argument-hint: "<architect|expo|ui|test> [target] [--review] [--audit]"
allowed-tools: Read, Write, Edit, Bash, Glob, Grep
model: sonnet
---

# /mobile-dev Command

One-command mobile development with React Native and Expo. Analyzes your project, identifies the framework and patterns, and routes to the appropriate specialist agent for architecture, Expo workflow, UI building, or testing.

## Usage

```
/mobile-dev                             # Auto-detect and suggest improvements
/mobile-dev architect                   # React Native architecture and navigation
/mobile-dev architect --navigation      # Focus on navigation patterns
/mobile-dev architect --native-modules  # Focus on native module bridging
/mobile-dev expo                        # Expo workflow and EAS
/mobile-dev expo --router               # Expo Router file-based routing
/mobile-dev expo --eas                  # EAS Build, Submit, Update
/mobile-dev ui                          # Mobile UI, animations, gestures
/mobile-dev ui --animations             # Reanimated animations
/mobile-dev ui --gestures               # Gesture Handler patterns
/mobile-dev test                        # Mobile testing strategies
/mobile-dev test --e2e                  # Detox or Maestro E2E testing
/mobile-dev test --ci                   # CI/CD for mobile
/mobile-dev --review                    # Full codebase review
```

## Subcommands

### `architect` — React Native Architecture

Designs and builds React Native applications with modern architecture patterns.

**What it does:**
1. Scans for existing RN setup, navigation, and state management
2. Identifies React Native version, architecture (Old vs New), and dependencies
3. Analyzes navigation structure and native module usage
4. Suggests or generates improvements

**Flags:**
- `--navigation` — React Navigation v7 patterns, deep linking, auth flows
- `--native-modules` — Native module creation, TurboModules, JSI bridging
- `--state` — Mobile state management (Zustand, MMKV, WatermelonDB)
- `--review` — Review existing architecture for best practices

**Routes to:** `react-native-architect` agent

### `expo` — Expo Workflow

Builds and optimizes Expo applications with EAS and modern Expo features.

**What it does:**
1. Analyzes `app.json`/`app.config.ts`, Expo SDK version, and config plugins
2. Reviews Expo Router setup and file-based routing
3. Checks EAS configuration and build profiles
4. Suggests or generates improvements

**Flags:**
- `--router` — Expo Router file-based routing, layouts, tabs, modals
- `--eas` — EAS Build, Submit, Update configuration
- `--config-plugins` — Custom config plugins and native code modification
- `--monorepo` — Monorepo setup with Expo

**Routes to:** `expo-engineer` agent

### `ui` — Mobile UI & Animations

Creates beautiful, performant mobile UIs with animations and gesture handling.

**What it does:**
1. Scans for existing UI libraries (NativeWind, Tamagui, Gluestack)
2. Reviews animation setup (Reanimated, Moti)
3. Checks gesture handling patterns
4. Suggests or generates components

**Flags:**
- `--animations` — Reanimated worklets, layout animations, shared transitions
- `--gestures` — Gesture Handler compositions, swipeable rows, pan interactions
- `--design-system` — Mobile design system with theming and accessibility
- `--responsive` — Responsive layouts, orientation handling, adaptive UI

**Routes to:** `mobile-ui-expert` agent

### `test` — Mobile Testing

Sets up and runs comprehensive mobile testing strategies.

**What it does:**
1. Identifies existing test setup (Jest, Detox, Maestro)
2. Reviews test coverage and patterns
3. Checks CI/CD configuration for mobile
4. Suggests or generates test improvements

**Flags:**
- `--e2e` — Detox or Maestro E2E test setup
- `--unit` — Jest + RNTL component and hook tests
- `--ci` — GitHub Actions, Fastlane, EAS Build in CI
- `--device-farms` — AWS Device Farm, BrowserStack setup

**Routes to:** `mobile-testing-expert` agent

## Auto-Detection

When no subcommand is specified, `/mobile-dev` auto-detects your setup:

1. **Expo detected** (`app.json` with expo, `expo-router`) → Routes to `expo-engineer`
2. **React Native with navigation issues** → Routes to `react-native-architect`
3. **UI/animation libraries detected** (Reanimated, NativeWind) → Routes to `mobile-ui-expert`
4. **Test setup detected** (Detox, Maestro configs) → Routes to `mobile-testing-expert`

## Agent Selection Guide

| Need | Agent | Command |
|------|-------|---------|
| RN architecture | react-native-architect | `/mobile-dev architect` |
| Navigation patterns | react-native-architect | `/mobile-dev architect --navigation` |
| Native modules | react-native-architect | `/mobile-dev architect --native-modules` |
| New Architecture | react-native-architect | `/mobile-dev architect` |
| Expo Router | expo-engineer | `/mobile-dev expo --router` |
| EAS Build/Submit | expo-engineer | `/mobile-dev expo --eas` |
| Config plugins | expo-engineer | `/mobile-dev expo --config-plugins` |
| Monorepo | expo-engineer | `/mobile-dev expo --monorepo` |
| Animations | mobile-ui-expert | `/mobile-dev ui --animations` |
| Gestures | mobile-ui-expert | `/mobile-dev ui --gestures` |
| Design system | mobile-ui-expert | `/mobile-dev ui --design-system` |
| Responsive UI | mobile-ui-expert | `/mobile-dev ui --responsive` |
| E2E testing | mobile-testing-expert | `/mobile-dev test --e2e` |
| Unit testing | mobile-testing-expert | `/mobile-dev test --unit` |
| CI/CD | mobile-testing-expert | `/mobile-dev test --ci` |
| Full review | All agents | `/mobile-dev --review` |

## Reference Materials

This suite includes comprehensive reference documents in `references/`:

- **react-native-patterns.md** — Offline-first, deep linking, push notifications, app state, background tasks, platform-specific code
- **expo-guide.md** — Project structure, config plugins, environment management, OTA updates, EAS workflows
- **mobile-performance.md** — JS thread optimization, FlatList/FlashList, image optimization, bundle size, startup time

Agents automatically consult these references when working. You can also read them directly for quick answers.

## How It Works

1. You describe what you need (e.g., "set up Expo Router with auth flow")
2. The command analyzes your project structure and existing mobile code
3. It routes to the appropriate specialist agent
4. The agent reads your code, understands your patterns, and generates solutions
5. Code is written following mobile best practices

All generated code follows these principles:
- **React Native 0.76+**: New Architecture, Fabric, TurboModules, JSI
- **Expo SDK 52+**: Expo Router v4, EAS Build/Update, config plugins
- **TypeScript**: Strict typing, platform-specific types, navigation typing
- **Performance**: Hermes engine, lazy loading, FlatList optimization, image caching
- **Accessibility**: accessibilityLabel, accessibilityRole, screen reader support, VoiceOver/TalkBack
- **Testing**: Detox E2E, Jest + RNTL, Maestro flows
