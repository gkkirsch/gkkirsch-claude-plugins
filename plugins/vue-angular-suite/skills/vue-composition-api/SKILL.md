---
name: vue-composition-api
description: >
  Vue 3 Composition API patterns and Nuxt 3 development.
  Use when building Vue 3 applications with Composition API, composables,
  Pinia state management, or Nuxt 3 server routes and auto-imports.
  Triggers: "vue 3", "composition api", "composable", "pinia", "nuxt 3",
  "vue router", "vue reactive", "vue ref", "defineProps", "defineEmits".
  NOT for: Vue 2 Options API, React, Angular, Svelte.
version: 1.0.0
allowed-tools: Read, Grep, Glob, Bash
---

# Vue 3 Composition API

## Component Patterns

```vue
<script setup lang="ts">
// <script setup> — the standard for Vue 3 components
import { ref, computed, watch, onMounted } from 'vue'

// Props with defaults and validation
const props = withDefaults(defineProps<{
  title: string
  count?: number
  items: string[]
  variant?: 'primary' | 'secondary'
}>(), {
  count: 0,
  variant: 'primary',
})

// Events with typed payloads
const emit = defineEmits<{
  'update:count': [value: number]
  'submit': [data: { title: string; items: string[] }]
  'close': []
}>()

// Reactive state
const searchQuery = ref('')
const isLoading = ref(false)

// Computed
const filteredItems = computed(() =>
  props.items.filter(item =>
    item.toLowerCase().includes(searchQuery.value.toLowerCase())
  )
)

// Watchers
watch(searchQuery, (newVal, oldVal) => {
  console.log(`Search changed: ${oldVal} → ${newVal}`)
}, { immediate: false })

// Deep watch on reactive objects
watch(
  () => props.items,
  (newItems) => { console.log('Items changed:', newItems.length) },
  { deep: true }
)

// Lifecycle
onMounted(async () => {
  isLoading.value = true
  // fetch initial data
  isLoading.value = false
})

// Methods
function handleSubmit() {
  emit('submit', { title: props.title, items: filteredItems.value })
}

// Expose for template refs (parent access)
defineExpose({ handleSubmit })
</script>

<template>
  <div :class="['card', `card--${variant}`]">
    <h2>{{ title }} ({{ filteredItems.length }})</h2>

    <input
      v-model="searchQuery"
      placeholder="Search..."
      type="search"
    />

    <ul v-if="filteredItems.length">
      <li v-for="item in filteredItems" :key="item">
        {{ item }}
      </li>
    </ul>
    <p v-else>No items found</p>

    <button @click="handleSubmit" :disabled="isLoading">
      Submit
    </button>
  </div>
</template>
```

## Composables (Reusable Logic)

```typescript
// composables/useApi.ts — generic API composable
import { ref, type Ref } from 'vue'

interface UseApiOptions<T> {
  immediate?: boolean
  initialData?: T
  onError?: (error: Error) => void
}

interface UseApiReturn<T> {
  data: Ref<T | null>
  error: Ref<Error | null>
  isLoading: Ref<boolean>
  execute: (...args: unknown[]) => Promise<T>
}

export function useApi<T>(
  fetcher: (...args: unknown[]) => Promise<T>,
  options: UseApiOptions<T> = {}
): UseApiReturn<T> {
  const data = ref<T | null>(options.initialData ?? null) as Ref<T | null>
  const error = ref<Error | null>(null)
  const isLoading = ref(false)

  async function execute(...args: unknown[]): Promise<T> {
    isLoading.value = true
    error.value = null
    try {
      const result = await fetcher(...args)
      data.value = result
      return result
    } catch (e) {
      const err = e instanceof Error ? e : new Error(String(e))
      error.value = err
      options.onError?.(err)
      throw err
    } finally {
      isLoading.value = false
    }
  }

  if (options.immediate) {
    execute()
  }

  return { data, error, isLoading, execute }
}

// Usage:
// const { data: users, isLoading, execute: fetchUsers } = useApi(
//   () => fetch('/api/users').then(r => r.json()),
//   { immediate: true }
// )
```

```typescript
// composables/useLocalStorage.ts — reactive localStorage
import { ref, watch, type Ref } from 'vue'

export function useLocalStorage<T>(key: string, defaultValue: T): Ref<T> {
  const stored = localStorage.getItem(key)
  const data = ref<T>(
    stored ? JSON.parse(stored) : defaultValue
  ) as Ref<T>

  watch(data, (newVal) => {
    localStorage.setItem(key, JSON.stringify(newVal))
  }, { deep: true })

  // Listen for changes from other tabs
  window.addEventListener('storage', (event) => {
    if (event.key === key && event.newValue) {
      data.value = JSON.parse(event.newValue)
    }
  })

  return data
}
```

```typescript
// composables/useDebounce.ts
import { ref, watch, type Ref } from 'vue'

export function useDebounce<T>(source: Ref<T>, delay: number = 300): Ref<T> {
  const debounced = ref(source.value) as Ref<T>
  let timeout: ReturnType<typeof setTimeout>

  watch(source, (newVal) => {
    clearTimeout(timeout)
    timeout = setTimeout(() => {
      debounced.value = newVal
    }, delay)
  })

  return debounced
}
```

## Pinia State Management

```typescript
// stores/auth.ts — Pinia store with composition style
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

interface User {
  id: string
  name: string
  email: string
  role: 'admin' | 'user'
}

export const useAuthStore = defineStore('auth', () => {
  // State
  const user = ref<User | null>(null)
  const token = ref<string | null>(localStorage.getItem('token'))
  const isLoading = ref(false)

  // Getters
  const isAuthenticated = computed(() => !!token.value && !!user.value)
  const isAdmin = computed(() => user.value?.role === 'admin')
  const displayName = computed(() => user.value?.name ?? 'Guest')

  // Actions
  async function login(email: string, password: string) {
    isLoading.value = true
    try {
      const response = await fetch('/api/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, password }),
      })
      if (!response.ok) throw new Error('Login failed')

      const data = await response.json()
      token.value = data.token
      user.value = data.user
      localStorage.setItem('token', data.token)
    } finally {
      isLoading.value = false
    }
  }

  function logout() {
    user.value = null
    token.value = null
    localStorage.removeItem('token')
  }

  async function fetchProfile() {
    if (!token.value) return
    const response = await fetch('/api/auth/me', {
      headers: { Authorization: `Bearer ${token.value}` },
    })
    if (response.ok) {
      user.value = await response.json()
    } else {
      logout()
    }
  }

  return {
    user, token, isLoading,
    isAuthenticated, isAdmin, displayName,
    login, logout, fetchProfile,
  }
}, {
  // Pinia persist plugin (optional)
  persist: {
    key: 'auth',
    pick: ['token'],
  },
})
```

## Vue Router Patterns

```typescript
// router/index.ts
import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      component: () => import('@/layouts/DefaultLayout.vue'),
      children: [
        { path: '', name: 'home', component: () => import('@/pages/Home.vue') },
        { path: 'about', name: 'about', component: () => import('@/pages/About.vue') },
      ],
    },
    {
      path: '/dashboard',
      component: () => import('@/layouts/DashboardLayout.vue'),
      meta: { requiresAuth: true },
      children: [
        { path: '', name: 'dashboard', component: () => import('@/pages/Dashboard.vue') },
        {
          path: 'users/:id',
          name: 'user-detail',
          component: () => import('@/pages/UserDetail.vue'),
          props: true, // pass route params as props
        },
        {
          path: 'admin',
          name: 'admin',
          component: () => import('@/pages/Admin.vue'),
          meta: { requiresAdmin: true },
        },
      ],
    },
    { path: '/login', name: 'login', component: () => import('@/pages/Login.vue') },
    { path: '/:pathMatch(.*)*', name: 'not-found', component: () => import('@/pages/NotFound.vue') },
  ],
})

// Navigation guards
router.beforeEach((to) => {
  const auth = useAuthStore()

  if (to.meta.requiresAuth && !auth.isAuthenticated) {
    return { name: 'login', query: { redirect: to.fullPath } }
  }

  if (to.meta.requiresAdmin && !auth.isAdmin) {
    return { name: 'dashboard' }
  }
})

export default router
```

## Nuxt 3 Patterns

```typescript
// server/api/users/index.get.ts — Nuxt server route
export default defineEventHandler(async (event) => {
  const query = getQuery(event)
  const page = Number(query.page) || 1
  const limit = Number(query.limit) || 20

  // Access runtime config
  const config = useRuntimeConfig()

  const users = await $fetch(`${config.apiBaseUrl}/users`, {
    query: { page, limit },
  })

  return users
})

// server/api/users/index.post.ts
export default defineEventHandler(async (event) => {
  const body = await readBody(event)

  if (!body.name || !body.email) {
    throw createError({
      statusCode: 400,
      statusMessage: 'Name and email are required',
    })
  }

  // Create user...
  return { id: 'new-id', ...body }
})

// server/middleware/auth.ts — server middleware
export default defineEventHandler((event) => {
  const protectedPaths = ['/api/admin', '/api/users']
  const path = getRequestURL(event).pathname

  if (protectedPaths.some(p => path.startsWith(p))) {
    const token = getHeader(event, 'authorization')?.replace('Bearer ', '')
    if (!token) {
      throw createError({ statusCode: 401, statusMessage: 'Unauthorized' })
    }
    // Verify token and attach user to event context
    event.context.user = verifyToken(token)
  }
})
```

```vue
<!-- pages/users.vue — Nuxt page with useAsyncData -->
<script setup lang="ts">
// Auto-imported: useAsyncData, useFetch, navigateTo, useRoute

const route = useRoute()
const page = computed(() => Number(route.query.page) || 1)

// useAsyncData — fetches on server, caches on client
const { data: users, pending, error, refresh } = await useAsyncData(
  `users-page-${page.value}`,
  () => $fetch('/api/users', { query: { page: page.value } }),
  { watch: [page] } // Re-fetch when page changes
)

// useFetch — shorthand for useAsyncData + $fetch
const { data: stats } = await useFetch('/api/stats', {
  pick: ['totalUsers', 'activeUsers'], // Only serialize these fields
  transform: (data) => ({
    ...data,
    activeRate: (data.activeUsers / data.totalUsers * 100).toFixed(1),
  }),
})

// SEO
useHead({
  title: 'Users',
  meta: [{ name: 'description', content: 'User management page' }],
})

// Middleware (page-level)
definePageMeta({
  middleware: 'auth',
  layout: 'dashboard',
})
</script>
```

## Gotchas

1. **ref() vs reactive()** -- `ref()` works for primitives AND objects (access via `.value`). `reactive()` only works for objects and loses reactivity if destructured. Stick to `ref()` for consistency. Use `toRefs()` when destructuring reactive objects: `const { name, email } = toRefs(state)`.

2. **watchEffect cleanup** -- `watchEffect` runs immediately and tracks dependencies. If it creates side effects (timers, subscriptions), return a cleanup function: `watchEffect((onCleanup) => { const id = setInterval(...); onCleanup(() => clearInterval(id)) })`.

3. **Pinia store outside setup** -- Calling `useAuthStore()` outside of `<script setup>` or a `setup()` function throws "getActivePinia() was called but there was no active Pinia." Pass the Pinia instance: `const store = useAuthStore(pinia)` in router guards and plugins.

4. **Template ref typing** -- `const inputRef = ref<HTMLInputElement | null>(null)` must include `null` because the element doesn't exist until mounted. Always check `if (inputRef.value)` before accessing DOM methods.

5. **Nuxt auto-imports and tree shaking** -- Nuxt auto-imports from `composables/`, `utils/`, and `server/utils/`. Files NOT in these directories won't auto-import. Don't `import { ref } from 'vue'` in Nuxt -- it's auto-imported and the explicit import can cause issues with SSR.

6. **v-model with Composition API** -- Custom v-model requires `defineModel()` (Vue 3.4+): `const modelValue = defineModel<string>()`. Before 3.4, use `defineProps(['modelValue'])` + `defineEmits(['update:modelValue'])` manually.
