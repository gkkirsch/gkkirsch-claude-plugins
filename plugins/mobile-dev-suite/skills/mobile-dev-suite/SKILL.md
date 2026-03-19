---
name: mobile-dev-suite
description: >
  Mobile Development Suite — complete toolkit for building production-grade React Native and Expo
  mobile applications. React Native architecture with Fabric, TurboModules, and navigation patterns,
  Expo workflow with Router, EAS Build/Submit/Update, and config plugins, mobile UI with Reanimated
  animations, Gesture Handler, and NativeWind, and mobile testing with Detox, Maestro, and CI/CD.
  Triggers: "react native", "expo", "mobile app", "mobile development", "react native navigation",
  "expo router", "eas build", "eas update", "eas submit", "config plugin", "expo modules",
  "native module", "turbomodule", "jsi", "fabric", "new architecture", "hermes",
  "react native reanimated", "reanimated", "gesture handler", "nativewind", "tamagui",
  "mobile animation", "mobile gesture", "mobile design system", "mobile responsive",
  "detox", "maestro", "react native testing library", "rntl", "mobile e2e",
  "mobile ci cd", "fastlane", "mobile testing", "device farm",
  "deep linking", "push notifications", "mobile offline", "watermelondb", "mmkv",
  "flatlist", "flashlist", "mobile performance", "app startup", "bundle size mobile".
  Dispatches the appropriate specialist agent: react-native-architect, expo-engineer,
  mobile-ui-expert, or mobile-testing-expert.
  NOT for: Web-only React development (use react-nextjs-suite), Flutter, native iOS (Swift/Kotlin),
  backend development, or infrastructure/DevOps.
version: 1.0.0
argument-hint: "<architect|expo|ui|test> [target]"
user-invocable: true
allowed-tools: Read, Grep, Glob, Bash
model: sonnet
---

# Mobile Development Suite

Production-grade React Native and Expo development agents for Claude Code. Four specialist agents that handle React Native architecture, Expo workflow, mobile UI/animations, and mobile testing — the complete mobile development lifecycle.

## Available Agents

### React Native Architect (`react-native-architect`)
Designs and builds React Native applications with modern architecture. New Architecture (Fabric, TurboModules, JSI), React Navigation v7, native module bridging, mobile state management (Zustand, MMKV, WatermelonDB), Hermes optimization, and platform-specific patterns.

**Invoke**: Dispatch via Task tool with `subagent_type: "react-native-architect"`.

**Example prompts**:
- "Design the navigation architecture for my app"
- "Migrate to React Native New Architecture"
- "Create a native module for Bluetooth"
- "Set up offline-first state with WatermelonDB"

### Expo Engineer (`expo-engineer`)
Builds production-grade Expo applications with modern workflow. Expo Router file-based routing, EAS Build/Submit/Update, config plugins for native code modification, dev builds, prebuild workflow, and monorepo setup.

**Invoke**: Dispatch via Task tool with `subagent_type: "expo-engineer"`.

**Example prompts**:
- "Set up Expo Router with auth and tab navigation"
- "Configure EAS Build with custom profiles"
- "Create a config plugin for custom native code"
- "Set up OTA updates with EAS Update channels"

### Mobile UI Expert (`mobile-ui-expert`)
Creates performant, beautiful mobile UIs with animations and gestures. React Native Reanimated worklets, Gesture Handler compositions, NativeWind/Tailwind styling, design systems, responsive layouts, and accessibility.

**Invoke**: Dispatch via Task tool with `subagent_type: "mobile-ui-expert"`.

**Example prompts**:
- "Build a swipeable card stack with spring animations"
- "Set up NativeWind with a mobile design system"
- "Create a bottom sheet with gesture interactions"
- "Build an image gallery with pinch-to-zoom"

### Mobile Testing Expert (`mobile-testing-expert`)
Comprehensive mobile testing from unit to E2E. Detox for native E2E, Maestro for declarative flows, Jest + RNTL for component tests, CI/CD with GitHub Actions and Fastlane, and device farm integration.

**Invoke**: Dispatch via Task tool with `subagent_type: "mobile-testing-expert"`.

**Example prompts**:
- "Set up Detox E2E testing with CI integration"
- "Write Maestro flows for critical user journeys"
- "Configure Jest + RNTL for component testing"
- "Set up GitHub Actions for mobile builds and tests"

## Quick Start: /mobile-dev

Use the `/mobile-dev` command for guided mobile development:

```
/mobile-dev                          # Auto-detect and suggest improvements
/mobile-dev architect                # React Native architecture
/mobile-dev architect --navigation   # Navigation patterns
/mobile-dev expo                     # Expo workflow and EAS
/mobile-dev expo --router            # Expo Router setup
/mobile-dev ui                       # Mobile UI and animations
/mobile-dev ui --animations          # Reanimated animations
/mobile-dev test                     # Mobile testing
/mobile-dev test --e2e               # E2E with Detox/Maestro
/mobile-dev --review                 # Full codebase review
```

## Reference Materials

This skill includes comprehensive reference documents in `references/`:

- **react-native-patterns.md** — Offline-first architecture, deep linking, push notifications, app state management, background tasks, platform-specific patterns
- **expo-guide.md** — Project structure, config plugins API, environment management, OTA update strategies, EAS workflows, monorepo setup
- **mobile-performance.md** — JS thread optimization, FlatList/FlashList tuning, image optimization, bundle size reduction, startup time improvement, memory management

Agents automatically consult these references when working. You can also read them directly for quick answers.

## How It Works

1. You describe what you need (e.g., "build a swipeable list with offline support")
2. The SKILL.md routes to the appropriate agent
3. The agent reads your code, discovers your stack and patterns
4. Solutions are designed and implemented following best practices
5. The agent provides results and next steps

All generated artifacts follow industry best practices:
- **React Native 0.76+**: New Architecture, Fabric, TurboModules, JSI, Hermes
- **Expo SDK 52+**: Router v4, EAS Build/Submit/Update, config plugins
- **TypeScript**: Strict typing, navigation types, platform-specific types
- **Performance**: Hermes, lazy loading, FlatList optimization, image caching
- **Accessibility**: VoiceOver, TalkBack, accessibilityLabel, roles, hints
- **Testing**: Detox E2E, Jest + RNTL, Maestro flows, CI/CD pipelines
