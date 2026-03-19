---
name: frontend-project-setup
description: >
  Scaffolds and configures Vue 3, Nuxt 3, and Angular 17+ projects with modern tooling.
  Triggers: vue project, angular project, nuxt project, frontend setup, scaffold vue, scaffold angular,
  create vue app, create angular app, new vue project, new angular project, nuxt setup.
  Dispatches to the appropriate agent based on the framework requested.
version: 1.0.0
argument-hint: "<vue|nuxt|angular> [project-name] [--ssr] [--pwa] [--testing]"
user-invocable: true
allowed-tools: Read, Write, Edit, Bash, Glob, Grep
model: sonnet
---

# Frontend Project Setup

This skill scaffolds and configures production-ready Vue 3, Nuxt 3, and Angular 17+ projects. It generates complete project structures with TypeScript, linting, testing, CI pipelines, and modern tooling pre-configured. The skill analyzes your request, selects the appropriate agent, and builds a fully functional starter project that follows current best practices.

## Available Agents

1. **Vue Architect** — Specializes in Vue 3 Composition API architecture, Pinia state management, Vue Router configuration, Nuxt 3 full-stack setup, Vite build tooling, and SSR strategies. Handles project scaffolding, module organization, and plugin integration.

2. **Angular Architect** — Specializes in Angular 17+ with standalone components, signals-based reactivity, RxJS integration, NgRx state management, Angular Universal SSR, and the new control flow syntax. Handles workspace configuration, lazy loading, and dependency injection patterns.

3. **Component Designer** — Specializes in reusable component patterns across both frameworks, including props/events contracts, slot and content projection patterns, design system setup, Storybook integration, and accessibility compliance. Framework-agnostic design principles applied to the target framework.

4. **Frontend Testing** — Specializes in comprehensive testing strategies using Vitest, Cypress, Playwright, Testing Library, and framework-specific test utilities. Configures unit, integration, and E2E test suites with coverage reporting and CI integration.

## Project Templates

### Vue 3 + Vite Template

#### Directory Structure

```
my-vue-app/
├── public/
│   └── favicon.ico
├── src/
│   ├── assets/
│   │   └── styles/
│   │       ├── main.css
│   │       └── variables.css
│   ├── components/
│   │   ├── common/
│   │   │   ├── AppHeader.vue
│   │   │   ├── AppFooter.vue
│   │   │   └── BaseButton.vue
│   │   └── __tests__/
│   │       └── BaseButton.spec.ts
│   ├── composables/
│   │   ├── useAuth.ts
│   │   └── useFetch.ts
│   ├── layouts/
│   │   └── DefaultLayout.vue
│   ├── pages/
│   │   ├── HomePage.vue
│   │   ├── AboutPage.vue
│   │   └── NotFoundPage.vue
│   ├── router/
│   │   └── index.ts
│   ├── stores/
│   │   ├── auth.ts
│   │   └── app.ts
│   ├── types/
│   │   └── index.ts
│   ├── utils/
│   │   └── helpers.ts
│   ├── App.vue
│   └── main.ts
├── e2e/
│   ├── fixtures/
│   ├── support/
│   └── home.spec.ts
├── .eslintrc.cjs
├── .prettierrc
├── .gitignore
├── env.d.ts
├── index.html
├── package.json
├── tsconfig.json
├── tsconfig.app.json
├── tsconfig.node.json
└── vite.config.ts
```

#### Key Configuration Files

**vite.config.ts**
```typescript
import { fileURLToPath, URL } from 'node:url'
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueDevTools from 'vite-plugin-vue-devtools'

export default defineConfig({
  plugins: [
    vue(),
    vueDevTools(),
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    }
  },
  server: {
    port: 3000,
    open: true
  },
  build: {
    target: 'esnext',
    sourcemap: true
  }
})
```

**tsconfig.json**
```json
{
  "references": [
    { "path": "./tsconfig.app.json" },
    { "path": "./tsconfig.node.json" }
  ],
  "compilerOptions": {
    "module": "NodeNext"
  }
}
```

**src/main.ts**
```typescript
import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import './assets/styles/main.css'

const app = createApp(App)

app.use(createPinia())
app.use(router)

app.mount('#app')
```

**src/App.vue**
```vue
<script setup lang="ts">
import { RouterView } from 'vue-router'
import AppHeader from '@/components/common/AppHeader.vue'
import AppFooter from '@/components/common/AppFooter.vue'
</script>

<template>
  <div id="app">
    <AppHeader />
    <main>
      <RouterView />
    </main>
    <AppFooter />
  </div>
</template>
```

#### Dependencies

```json
{
  "dependencies": {
    "vue": "^3.5",
    "vue-router": "^4.4",
    "pinia": "^2.2"
  },
  "devDependencies": {
    "@vitejs/plugin-vue": "^5.2",
    "@vue/tsconfig": "^0.7",
    "typescript": "~5.7",
    "vite": "^6.0",
    "vite-plugin-vue-devtools": "^7.6",
    "vue-tsc": "^2.2",
    "eslint": "^9.16",
    "prettier": "^3.4"
  }
}
```

#### Scripts

```json
{
  "scripts": {
    "dev": "vite",
    "build": "run-p type-check \"build-only {@}\" --",
    "preview": "vite preview",
    "build-only": "vite build",
    "type-check": "vue-tsc --build",
    "lint": "eslint . --fix",
    "format": "prettier --write src/"
  }
}
```

#### Setup Steps

1. Run the scaffolding command to create the project directory and install base dependencies
2. Add Pinia stores with typed state, getters, and actions
3. Configure Vue Router with typed route definitions and navigation guards
4. Set up ESLint with `@vue/eslint-config-typescript` and Prettier integration
5. Create base layout components and a default page structure
6. Add path aliases in both `vite.config.ts` and `tsconfig.app.json`

---

### Nuxt 3 Template

#### Directory Structure

```
my-nuxt-app/
├── assets/
│   └── css/
│       └── main.css
├── components/
│   ├── app/
│   │   ├── AppHeader.vue
│   │   └── AppFooter.vue
│   └── ui/
│       ├── BaseButton.vue
│       └── BaseCard.vue
├── composables/
│   ├── useAuth.ts
│   └── useApi.ts
├── layouts/
│   ├── default.vue
│   └── auth.vue
├── middleware/
│   └── auth.ts
├── pages/
│   ├── index.vue
│   ├── about.vue
│   └── dashboard/
│       └── index.vue
├── plugins/
│   └── init.ts
├── server/
│   ├── api/
│   │   ├── health.get.ts
│   │   └── users/
│   │       ├── index.get.ts
│   │       └── [id].get.ts
│   ├── middleware/
│   │   └── log.ts
│   └── utils/
│       └── db.ts
├── stores/
│   └── auth.ts
├── types/
│   └── index.ts
├── public/
│   └── favicon.ico
├── .eslintrc.cjs
├── .gitignore
├── app.vue
├── nuxt.config.ts
├── package.json
└── tsconfig.json
```

#### Key Configuration Files

**nuxt.config.ts**
```typescript
export default defineNuxtConfig({
  devtools: { enabled: true },

  modules: [
    '@pinia/nuxt',
    '@nuxtjs/color-mode',
    '@vueuse/nuxt',
    '@nuxt/eslint'
  ],

  css: ['~/assets/css/main.css'],

  typescript: {
    strict: true,
    typeCheck: true
  },

  runtimeConfig: {
    apiSecret: '',
    public: {
      apiBase: '/api'
    }
  },

  routeRules: {
    '/api/**': { cors: true },
    '/dashboard/**': { ssr: false }
  },

  compatibilityDate: '2025-01-01'
})
```

**app.vue**
```vue
<template>
  <NuxtLayout>
    <NuxtPage />
  </NuxtLayout>
</template>
```

**server/api/health.get.ts**
```typescript
export default defineEventHandler(() => {
  return {
    status: 'ok',
    timestamp: new Date().toISOString()
  }
})
```

#### Dependencies

```json
{
  "dependencies": {
    "nuxt": "^3.15",
    "@pinia/nuxt": "^0.9",
    "pinia": "^2.2"
  },
  "devDependencies": {
    "@nuxt/eslint": "^0.7",
    "@nuxtjs/color-mode": "^3.5",
    "@vueuse/nuxt": "^12.0",
    "typescript": "~5.7"
  }
}
```

#### Scripts

```json
{
  "scripts": {
    "dev": "nuxt dev",
    "build": "nuxt build",
    "generate": "nuxt generate",
    "preview": "nuxt preview",
    "postinstall": "nuxt prepare",
    "lint": "eslint .",
    "typecheck": "nuxt typecheck"
  }
}
```

#### Setup Steps

1. Initialize the Nuxt 3 project with `npx nuxi init`
2. Install Pinia module and configure in `nuxt.config.ts`
3. Set up auto-imported composables and components directories
4. Configure server API routes with proper HTTP method file naming
5. Add middleware for authentication and logging
6. Set up runtime config for environment variables

---

### Angular 17+ Template

#### Directory Structure

```
my-angular-app/
├── src/
│   ├── app/
│   │   ├── core/
│   │   │   ├── guards/
│   │   │   │   └── auth.guard.ts
│   │   │   ├── interceptors/
│   │   │   │   └── auth.interceptor.ts
│   │   │   ├── services/
│   │   │   │   ├── auth.service.ts
│   │   │   │   └── api.service.ts
│   │   │   └── models/
│   │   │       └── user.model.ts
│   │   ├── features/
│   │   │   ├── home/
│   │   │   │   └── home.component.ts
│   │   │   ├── dashboard/
│   │   │   │   ├── dashboard.component.ts
│   │   │   │   └── dashboard.routes.ts
│   │   │   └── shared/
│   │   │       └── not-found.component.ts
│   │   ├── shared/
│   │   │   ├── components/
│   │   │   │   ├── header/
│   │   │   │   │   └── header.component.ts
│   │   │   │   └── footer/
│   │   │   │       └── footer.component.ts
│   │   │   ├── directives/
│   │   │   ├── pipes/
│   │   │   └── ui/
│   │   │       └── button/
│   │   │           └── button.component.ts
│   │   ├── app.component.ts
│   │   ├── app.config.ts
│   │   └── app.routes.ts
│   ├── assets/
│   ├── environments/
│   │   ├── environment.ts
│   │   └── environment.prod.ts
│   ├── styles/
│   │   ├── _variables.scss
│   │   └── styles.scss
│   ├── index.html
│   └── main.ts
├── e2e/
│   └── home.spec.ts
├── .editorconfig
├── .gitignore
├── angular.json
├── package.json
├── tsconfig.json
├── tsconfig.app.json
└── tsconfig.spec.json
```

#### Key Configuration Files

**angular.json** (abbreviated)
```json
{
  "$schema": "./node_modules/@angular/cli/lib/config/schema.json",
  "version": 1,
  "newProjectRoot": "projects",
  "projects": {
    "my-angular-app": {
      "projectType": "application",
      "root": "",
      "sourceRoot": "src",
      "architect": {
        "build": {
          "builder": "@angular-devkit/build-angular:application",
          "options": {
            "outputPath": "dist/my-angular-app",
            "index": "src/index.html",
            "browser": "src/main.ts",
            "tsConfig": "tsconfig.app.json",
            "inlineStyleLanguage": "scss",
            "styles": ["src/styles/styles.scss"],
            "scripts": []
          }
        }
      }
    }
  }
}
```

**tsconfig.json**
```json
{
  "compileOnSave": false,
  "compilerOptions": {
    "outDir": "./dist/out-tsc",
    "strict": true,
    "noImplicitOverride": true,
    "noPropertyAccessFromIndexSignature": true,
    "noImplicitReturns": true,
    "noFallthroughCasesInSwitch": true,
    "sourceMap": true,
    "declaration": false,
    "downlevelIteration": true,
    "experimentalDecorators": true,
    "moduleResolution": "bundler",
    "importHelpers": true,
    "target": "ES2022",
    "module": "ES2022",
    "lib": ["ES2022", "dom"],
    "paths": {
      "@core/*": ["src/app/core/*"],
      "@shared/*": ["src/app/shared/*"],
      "@features/*": ["src/app/features/*"],
      "@env/*": ["src/environments/*"]
    }
  }
}
```

**src/main.ts**
```typescript
import { bootstrapApplication } from '@angular/platform-browser'
import { appConfig } from './app/app.config'
import { AppComponent } from './app/app.component'

bootstrapApplication(AppComponent, appConfig)
  .catch((err) => console.error(err))
```

**src/app/app.config.ts**
```typescript
import { ApplicationConfig, provideZoneChangeDetection } from '@angular/core'
import { provideRouter, withComponentInputBinding } from '@angular/router'
import { provideHttpClient, withInterceptors, withFetch } from '@angular/common/http'
import { provideClientHydration, withEventReplay } from '@angular/platform-browser'

import { routes } from './app.routes'
import { authInterceptor } from './core/interceptors/auth.interceptor'

export const appConfig: ApplicationConfig = {
  providers: [
    provideZoneChangeDetection({ eventCoalescing: true }),
    provideRouter(routes, withComponentInputBinding()),
    provideHttpClient(withInterceptors([authInterceptor]), withFetch()),
    provideClientHydration(withEventReplay())
  ]
}
```

**src/app/app.routes.ts**
```typescript
import { Routes } from '@angular/router'
import { authGuard } from './core/guards/auth.guard'

export const routes: Routes = [
  {
    path: '',
    loadComponent: () =>
      import('./features/home/home.component').then(m => m.HomeComponent)
  },
  {
    path: 'dashboard',
    canActivate: [authGuard],
    loadChildren: () =>
      import('./features/dashboard/dashboard.routes').then(m => m.DASHBOARD_ROUTES)
  },
  {
    path: '**',
    loadComponent: () =>
      import('./features/shared/not-found.component').then(m => m.NotFoundComponent)
  }
]
```

#### Dependencies

```json
{
  "dependencies": {
    "@angular/animations": "^19.0",
    "@angular/common": "^19.0",
    "@angular/compiler": "^19.0",
    "@angular/core": "^19.0",
    "@angular/forms": "^19.0",
    "@angular/platform-browser": "^19.0",
    "@angular/platform-browser-dynamic": "^19.0",
    "@angular/platform-server": "^19.0",
    "@angular/router": "^19.0",
    "@angular/ssr": "^19.0",
    "rxjs": "~7.8",
    "tslib": "^2.8",
    "zone.js": "~0.15"
  },
  "devDependencies": {
    "@angular-devkit/build-angular": "^19.0",
    "@angular/cli": "^19.0",
    "@angular/compiler-cli": "^19.0",
    "typescript": "~5.7"
  }
}
```

#### Scripts

```json
{
  "scripts": {
    "ng": "ng",
    "start": "ng serve",
    "build": "ng build",
    "watch": "ng build --watch --configuration development",
    "test": "ng test",
    "serve:ssr": "node dist/my-angular-app/server/server.mjs"
  }
}
```

#### Setup Steps

1. Generate the workspace with `npx @angular/cli new` using standalone APIs
2. Configure path aliases in `tsconfig.json` for clean imports
3. Create core, shared, and features directory structure
4. Set up functional guards and interceptors using the `inject` function
5. Configure lazy-loaded routes with `loadComponent` and `loadChildren`
6. Add environment files and wire them into the build configuration

---

## Agent Selection Guide

| Task | Agent | Example Prompt |
|------|-------|----------------|
| Vue 3 app architecture | Vue Architect | "Set up a Vue 3 app with Pinia and Vue Router" |
| Nuxt 3 full-stack app | Vue Architect | "Create a Nuxt 3 app with server routes and SSR" |
| Angular app architecture | Angular Architect | "Set up an Angular 17 app with signals and standalone" |
| Component library setup | Component Designer | "Build a Vue component library with Storybook" |
| Design system creation | Component Designer | "Create an Angular design system with tokens" |
| Testing strategy | Frontend Testing | "Set up Vitest and Playwright for my Vue app" |
| E2E test suite | Frontend Testing | "Add Cypress E2E tests to my Angular project" |
| SSR configuration | Vue Architect / Angular Architect | "Enable SSR with hydration in my Angular app" |
| State management | Vue Architect / Angular Architect | "Add NgRx signals store to my Angular app" |

## Configuration Options

### SSR Mode (--ssr)

When the `--ssr` flag is provided, the skill configures server-side rendering:

- **Vue**: Scaffolds a Nuxt 3 project with SSR enabled by default. Configures `routeRules` for hybrid rendering, sets up server middleware, and enables payload extraction for optimal hydration.
- **Angular**: Configures Angular Universal with `@angular/ssr`. Sets up `provideClientHydration(withEventReplay())` in the app config, generates `server.ts` entry point, and configures the Express server for production SSR.

### PWA Mode (--pwa)

When the `--pwa` flag is provided, the skill adds progressive web app capabilities:

- **Vue**: Installs and configures `vite-plugin-pwa` with a generated service worker, web manifest, and offline fallback page. Adds icons and configures caching strategies for assets and API routes.
- **Angular**: Adds `@angular/pwa` using `ng add`, which generates `manifest.webmanifest`, `ngsw-config.json`, and registers the Angular service worker. Configures asset groups and data groups for caching.

### Testing Mode (--testing)

When the `--testing` flag is provided, the skill sets up a complete testing infrastructure:

- **Vue**: Installs Vitest with `@vue/test-utils` for unit and component testing. Adds Playwright for E2E tests. Configures `vitest.config.ts` with Vue plugin, coverage via `@vitest/coverage-v8`, and test path aliases matching the main config.
- **Angular**: Configures the Angular CLI test runner or migrates to Vitest with `@analogjs/vitest-angular`. Adds `@testing-library/angular` for component tests. Sets up Playwright for E2E with page object patterns.

## Common Setup Commands

### Vue 3

```bash
# Create a new Vue 3 project with TypeScript, Router, and Pinia
npm create vue@latest my-app -- --typescript --router --pinia

# Navigate and install
cd my-app && npm install

# Add commonly used dev dependencies
npm install -D @vueuse/core unplugin-auto-import unplugin-vue-components

# Start development server
npm run dev
```

### Nuxt 3

```bash
# Initialize a new Nuxt 3 project
npx nuxi init my-app

# Navigate and install
cd my-app && npm install

# Add commonly used modules
npx nuxi module add pinia
npx nuxi module add vueuse
npx nuxi module add eslint

# Start development server
npm run dev
```

### Angular 17+

```bash
# Create a new Angular project with SCSS, routing, and SSR
npx @angular/cli new my-app --style=scss --routing --ssr

# Navigate into the project
cd my-app

# Add commonly used libraries
ng add @angular/material
ng add @ngrx/signals

# Start development server
ng serve
```

## Reference Materials

- **Vue 3 Documentation**: Refer to `plugins/vue-angular-suite/references/vue3-composition-api.md` for Composition API patterns, reactive primitives, and lifecycle hooks.
- **Nuxt 3 Documentation**: Refer to `plugins/vue-angular-suite/references/nuxt3-server-routes.md` for server API patterns, middleware, and runtime config.
- **Angular Documentation**: Refer to `plugins/vue-angular-suite/references/angular-signals.md` for signals, computed values, and effects.
- **Testing Guides**: Refer to `plugins/vue-angular-suite/references/testing-patterns.md` for framework-specific test setup and patterns.
- **Component Patterns**: Refer to `plugins/vue-angular-suite/references/component-design.md` for cross-framework component architecture.

## How It Works

The skill follows a five-step process to scaffold your frontend project:

### Step 1: Parse the Framework Request

The skill analyzes the user prompt to determine:
- Which framework is requested (Vue 3, Nuxt 3, or Angular 17+)
- The desired project name
- Optional flags (`--ssr`, `--pwa`, `--testing`)
- Any additional requirements mentioned in the prompt (state management, UI library, etc.)

### Step 2: Select the Appropriate Agent

Based on the parsed request, the skill dispatches to one of the four available agents. If the request spans multiple concerns (e.g., "Set up a Vue app with a component library and full testing"), the primary agent handles scaffolding while coordinating with secondary agents for their specialties.

### Step 3: Scaffold the Project Structure

The selected agent creates the full directory structure using the appropriate template. This includes:
- Creating all directories and placeholder files
- Writing configuration files (`vite.config.ts`, `nuxt.config.ts`, or `angular.json`)
- Setting up TypeScript with strict mode and path aliases
- Generating `package.json` with all required dependencies

### Step 4: Configure Tooling

The agent then configures supporting tools:
- **TypeScript**: Strict mode, path aliases, proper module resolution
- **Linting**: ESLint with framework-specific plugins and Prettier integration
- **Testing**: Unit test runner, component test utilities, and E2E framework
- **CI**: GitHub Actions workflow for lint, type-check, test, and build
- **Git**: `.gitignore` with framework-specific exclusions

### Step 5: Generate Starter Components and Tests

Finally, the agent creates functional starter code:
- A root app component with layout structure
- A home page component with basic content
- A reusable base button component demonstrating the component pattern
- An example composable or service showing the data-fetching pattern
- Unit tests for the base component and composable/service
- An E2E test for the home page navigation

All generated projects follow modern best practices including TypeScript strict mode, Composition API (Vue) or standalone components with signals (Angular), tree-shakeable imports, and comprehensive linting rules. The scaffolded code is production-ready and serves as a solid foundation for building real applications.
