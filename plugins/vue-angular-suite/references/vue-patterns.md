# Vue 3 Patterns Quick Reference

> Claude Code plugin reference for Vue 3 + TypeScript patterns.
> Composition API, composables, Pinia, Nuxt 3, Vue Router, and more.

---

## Composition API Quick Reference

### Core Reactivity Functions

| Function | Description | Returns |
|---|---|---|
| `ref(value)` | Wraps a primitive or object in a reactive reference | `Ref<T>` |
| `reactive(obj)` | Makes a plain object deeply reactive | `T` (proxied) |
| `readonly(obj)` | Creates a read-only proxy of a reactive object | `DeepReadonly<T>` |
| `computed(getter)` | Derives a cached value from reactive dependencies | `ComputedRef<T>` |
| `computed({ get, set })` | Writable computed property | `WritableComputedRef<T>` |
| `watch(source, cb, opts?)` | Watches one or more reactive sources | `WatchStopHandle` |
| `watchEffect(cb, opts?)` | Runs a side effect immediately and re-runs on dependency change | `WatchStopHandle` |
| `watchPostEffect(cb)` | `watchEffect` with `flush: 'post'` | `WatchStopHandle` |
| `watchSyncEffect(cb)` | `watchEffect` with `flush: 'sync'` | `WatchStopHandle` |
| `shallowRef(value)` | Ref that only tracks `.value` reassignment (not deep) | `ShallowRef<T>` |
| `shallowReactive(obj)` | Reactive only at root level | `T` (proxied) |
| `triggerRef(ref)` | Force trigger watchers on a `shallowRef` | `void` |
| `toRef(obj, key)` | Creates a ref linked to a reactive object property | `Ref<T>` |
| `toRefs(obj)` | Converts every property of a reactive object to a ref | `ToRefs<T>` |
| `toValue(source)` | Unwraps a ref, getter, or returns the value as-is | `T` |
| `unref(ref)` | Unwraps a ref to get its inner value | `T` |
| `isRef(val)` | Type guard: checks if a value is a ref | `boolean` |
| `isReactive(val)` | Type guard: checks if a value is reactive | `boolean` |
| `isReadonly(val)` | Type guard: checks if a value is readonly | `boolean` |
| `isProxy(val)` | Type guard: checks if created by `reactive` or `readonly` | `boolean` |
| `markRaw(obj)` | Marks an object so it will never become reactive | `T` |
| `effectScope()` | Creates an effect scope to group reactive effects | `EffectScope` |
| `getCurrentScope()` | Returns current active effect scope | `EffectScope \| undefined` |
| `onScopeDispose(cb)` | Registers a callback when the current scope is disposed | `void` |

### Lifecycle Hooks

| Hook | Description | Equivalent Options API |
|---|---|---|
| `onBeforeMount(cb)` | Before the component is mounted to the DOM | `beforeMount` |
| `onMounted(cb)` | After the component is mounted | `mounted` |
| `onBeforeUpdate(cb)` | Before the component re-renders | `beforeUpdate` |
| `onUpdated(cb)` | After the component re-renders | `updated` |
| `onBeforeUnmount(cb)` | Before the component is destroyed | `beforeUnmount` |
| `onUnmounted(cb)` | After the component is destroyed | `unmounted` |
| `onActivated(cb)` | When a `<KeepAlive>` component is activated | `activated` |
| `onDeactivated(cb)` | When a `<KeepAlive>` component is deactivated | `deactivated` |
| `onErrorCaptured(cb)` | When an error from a descendant is captured | `errorCaptured` |
| `onRenderTracked(cb)` | (dev only) When a reactive dependency is tracked | `renderTracked` |
| `onRenderTriggered(cb)` | (dev only) When a reactive dependency triggers re-render | `renderTriggered` |
| `onServerPrefetch(cb)` | Async hook that runs before SSR rendering | -- |

### ref vs reactive Decision Table

| Scenario | Use `ref` | Use `reactive` |
|---|---|---|
| Primitive values (string, number, boolean) | Yes | No |
| Reassigning the entire value | Yes (`.value = newVal`) | No (loses reactivity) |
| Destructuring in composable returns | Yes (stays reactive) | No (loses reactivity) |
| Template auto-unwrap | Yes (auto-unwrapped) | Yes (directly accessed) |
| Complex nested objects you never reassign | Possible | Yes |
| Arrays you might replace entirely | Yes | No |
| Passing to composables | Yes (explicit) | Less portable |

**Rule of thumb:** Default to `ref` for everything. Use `reactive` only when you have a complex object with many properties that you access frequently and never reassign wholesale.

---

## Composable Patterns

### Standard Composable Structure

```ts
// composables/useFeature.ts
import { ref, computed, onMounted, onUnmounted } from 'vue'
import type { Ref, MaybeRefOrGetter } from 'vue'

interface UseFeatureOptions {
  immediate?: boolean
  timeout?: number
}

interface UseFeatureReturn {
  data: Ref<string | null>
  isLoading: Ref<boolean>
  error: Ref<Error | null>
  execute: () => Promise<void>
  reset: () => void
}

export function useFeature(
  input: MaybeRefOrGetter<string>,
  options: UseFeatureOptions = {}
): UseFeatureReturn {
  const { immediate = true, timeout = 5000 } = options

  const data = ref<string | null>(null)
  const isLoading = ref(false)
  const error = ref<Error | null>(null)

  async function execute() {
    isLoading.value = true
    error.value = null
    try {
      const resolved = toValue(input)
      data.value = await fetchData(resolved, timeout)
    } catch (e) {
      error.value = e instanceof Error ? e : new Error(String(e))
    } finally {
      isLoading.value = false
    }
  }

  function reset() {
    data.value = null
    error.value = null
    isLoading.value = false
  }

  if (immediate) {
    execute()
  }

  return { data, isLoading, error, execute, reset }
}
```

### useLocalStorage Composable

```ts
// composables/useLocalStorage.ts
import { ref, watch } from 'vue'
import type { Ref } from 'vue'

export function useLocalStorage<T>(
  key: string,
  defaultValue: T
): Ref<T> {
  const stored = localStorage.getItem(key)
  const data = ref<T>(
    stored !== null ? JSON.parse(stored) : defaultValue
  ) as Ref<T>

  watch(
    data,
    (newValue) => {
      if (newValue === null || newValue === undefined) {
        localStorage.removeItem(key)
      } else {
        localStorage.setItem(key, JSON.stringify(newValue))
      }
    },
    { deep: true }
  )

  // Listen for changes from other tabs
  function handleStorageEvent(e: StorageEvent) {
    if (e.key === key) {
      data.value = e.newValue !== null
        ? JSON.parse(e.newValue)
        : defaultValue
    }
  }

  window.addEventListener('storage', handleStorageEvent)
  onScopeDispose(() => {
    window.removeEventListener('storage', handleStorageEvent)
  })

  return data
}
```

### useFetch Composable with Abort

```ts
// composables/useFetch.ts
import { ref, toValue, watchEffect, onScopeDispose } from 'vue'
import type { Ref, MaybeRefOrGetter } from 'vue'

interface UseFetchOptions extends RequestInit {
  immediate?: boolean
  refetch?: boolean
}

interface UseFetchReturn<T> {
  data: Ref<T | null>
  error: Ref<Error | null>
  isLoading: Ref<boolean>
  statusCode: Ref<number | null>
  abort: () => void
  execute: () => Promise<void>
  canAbort: Ref<boolean>
}

export function useFetch<T = unknown>(
  url: MaybeRefOrGetter<string>,
  options: UseFetchOptions = {}
): UseFetchReturn<T> {
  const { immediate = true, refetch = false, ...fetchOptions } = options

  const data = ref<T | null>(null) as Ref<T | null>
  const error = ref<Error | null>(null)
  const isLoading = ref(false)
  const statusCode = ref<number | null>(null)
  const canAbort = ref(false)

  let abortController: AbortController | null = null

  function abort() {
    abortController?.abort()
    abortController = null
    canAbort.value = false
  }

  async function execute() {
    abort()
    abortController = new AbortController()
    canAbort.value = true

    isLoading.value = true
    error.value = null
    statusCode.value = null

    try {
      const response = await fetch(toValue(url), {
        ...fetchOptions,
        signal: abortController.signal,
      })
      statusCode.value = response.status

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`)
      }

      const contentType = response.headers.get('content-type')
      data.value = contentType?.includes('application/json')
        ? await response.json()
        : await response.text() as unknown as T
    } catch (e) {
      if (e instanceof DOMException && e.name === 'AbortError') {
        return
      }
      error.value = e instanceof Error ? e : new Error(String(e))
    } finally {
      isLoading.value = false
      canAbort.value = false
    }
  }

  if (refetch) {
    watchEffect(() => {
      toValue(url) // track the url
      execute()
    })
  } else if (immediate) {
    execute()
  }

  onScopeDispose(abort)

  return { data, error, isLoading, statusCode, abort, execute, canAbort }
}
```

### useDebounce Composable

```ts
// composables/useDebounce.ts
import { ref, watch } from 'vue'
import type { Ref, MaybeRefOrGetter } from 'vue'

export function useDebounce<T>(
  source: MaybeRefOrGetter<T>,
  delay: number = 300
): Ref<T> {
  const debounced = ref(toValue(source)) as Ref<T>
  let timeout: ReturnType<typeof setTimeout>

  watch(
    () => toValue(source),
    (newVal) => {
      clearTimeout(timeout)
      timeout = setTimeout(() => {
        debounced.value = newVal
      }, delay)
    }
  )

  onScopeDispose(() => clearTimeout(timeout))

  return debounced
}

// Usage with a search input:
// const search = ref('')
// const debouncedSearch = useDebounce(search, 500)
// watch(debouncedSearch, (val) => fetchResults(val))
```

### useIntersectionObserver Composable

```ts
// composables/useIntersectionObserver.ts
import { ref, watch, onScopeDispose } from 'vue'
import type { Ref } from 'vue'

interface UseIntersectionObserverOptions {
  root?: HTMLElement | null
  rootMargin?: string
  threshold?: number | number[]
  immediate?: boolean
}

export function useIntersectionObserver(
  target: Ref<HTMLElement | null | undefined>,
  options: UseIntersectionObserverOptions = {}
) {
  const {
    root = null,
    rootMargin = '0px',
    threshold = 0,
    immediate = true,
  } = options

  const isIntersecting = ref(false)
  const isSupported = ref('IntersectionObserver' in window)

  let observer: IntersectionObserver | null = null

  function cleanup() {
    observer?.disconnect()
    observer = null
  }

  function observe() {
    cleanup()
    if (!isSupported.value || !target.value) return

    observer = new IntersectionObserver(
      ([entry]) => {
        isIntersecting.value = entry.isIntersecting
      },
      { root, rootMargin, threshold }
    )
    observer.observe(target.value)
  }

  if (immediate) {
    watch(
      target,
      () => observe(),
      { immediate: true, flush: 'post' }
    )
  }

  onScopeDispose(cleanup)

  return { isIntersecting, isSupported, observe, stop: cleanup }
}
```

### useEventListener Composable

```ts
// composables/useEventListener.ts
import { onScopeDispose, watch, toValue } from 'vue'
import type { MaybeRefOrGetter } from 'vue'

type EventTarget = Window | Document | HTMLElement

export function useEventListener<K extends keyof WindowEventMap>(
  target: MaybeRefOrGetter<EventTarget | null | undefined>,
  event: K,
  handler: (evt: WindowEventMap[K]) => void,
  options?: boolean | AddEventListenerOptions
): () => void {
  let cleanup = () => {}

  const stopWatch = watch(
    () => toValue(target),
    (el) => {
      cleanup()
      if (!el) return

      el.addEventListener(event, handler as EventListener, options)
      cleanup = () => {
        el.removeEventListener(event, handler as EventListener, options)
      }
    },
    { immediate: true, flush: 'post' }
  )

  const stop = () => {
    stopWatch()
    cleanup()
  }

  onScopeDispose(stop)
  return stop
}

// Usage:
// useEventListener(window, 'resize', () => { ... })
// useEventListener(buttonRef, 'click', handleClick)
```

### useMediaQuery Composable

```ts
// composables/useMediaQuery.ts
import { ref, onScopeDispose } from 'vue'
import type { Ref } from 'vue'

export function useMediaQuery(query: string): Ref<boolean> {
  const matches = ref(false)

  const mediaQuery = window.matchMedia(query)
  matches.value = mediaQuery.matches

  function handler(e: MediaQueryListEvent) {
    matches.value = e.matches
  }

  mediaQuery.addEventListener('change', handler)
  onScopeDispose(() => {
    mediaQuery.removeEventListener('change', handler)
  })

  return matches
}

// Convenience composables built on useMediaQuery:
export function usePrefersDark() {
  return useMediaQuery('(prefers-color-scheme: dark)')
}

export function useBreakpoints() {
  return {
    sm: useMediaQuery('(min-width: 640px)'),
    md: useMediaQuery('(min-width: 768px)'),
    lg: useMediaQuery('(min-width: 1024px)'),
    xl: useMediaQuery('(min-width: 1280px)'),
    xxl: useMediaQuery('(min-width: 1536px)'),
  }
}
```

### usePagination Composable

```ts
// composables/usePagination.ts
import { ref, computed } from 'vue'
import type { Ref } from 'vue'

interface UsePaginationOptions {
  pageSize?: number
  initialPage?: number
}

interface UsePaginationReturn {
  currentPage: Ref<number>
  pageSize: Ref<number>
  totalItems: Ref<number>
  totalPages: Readonly<Ref<number>>
  isFirstPage: Readonly<Ref<boolean>>
  isLastPage: Readonly<Ref<boolean>>
  offset: Readonly<Ref<number>>
  next: () => void
  prev: () => void
  goTo: (page: number) => void
  setTotal: (total: number) => void
}

export function usePagination(
  options: UsePaginationOptions = {}
): UsePaginationReturn {
  const { pageSize: defaultPageSize = 10, initialPage = 1 } = options

  const currentPage = ref(initialPage)
  const pageSize = ref(defaultPageSize)
  const totalItems = ref(0)

  const totalPages = computed(() =>
    Math.max(1, Math.ceil(totalItems.value / pageSize.value))
  )
  const isFirstPage = computed(() => currentPage.value <= 1)
  const isLastPage = computed(() => currentPage.value >= totalPages.value)
  const offset = computed(() => (currentPage.value - 1) * pageSize.value)

  function next() {
    if (!isLastPage.value) currentPage.value++
  }

  function prev() {
    if (!isFirstPage.value) currentPage.value--
  }

  function goTo(page: number) {
    currentPage.value = Math.max(1, Math.min(page, totalPages.value))
  }

  function setTotal(total: number) {
    totalItems.value = total
    if (currentPage.value > totalPages.value) {
      currentPage.value = totalPages.value
    }
  }

  return {
    currentPage,
    pageSize,
    totalItems,
    totalPages,
    isFirstPage,
    isLastPage,
    offset,
    next,
    prev,
    goTo,
    setTotal,
  }
}
```

### Composable Naming Conventions

| Convention | Example | Notes |
|---|---|---|
| Prefix with `use` | `useAuth`, `useFetch` | Universal Vue convention |
| One composable per file | `useAuth.ts` | Keeps imports clean |
| File matches function name | `useAuth.ts` exports `useAuth` | Easy to locate |
| Place in `composables/` directory | `composables/useAuth.ts` | Standard project layout |
| Test file beside source | `composables/useAuth.test.ts` | Or in `__tests__/` folder |
| Group by domain if large | `composables/auth/useLogin.ts` | For large codebases |

### Return Patterns

```ts
// PREFERRED: Return individual refs (destructurable, flexible)
export function useCounter() {
  const count = ref(0)
  const doubled = computed(() => count.value * 2)
  const increment = () => count.value++
  return { count, doubled, increment }
}

// ALTERNATIVE: Return a reactive object (not destructurable)
export function useCounter() {
  const state = reactive({
    count: 0,
    get doubled() { return this.count * 2 },
    increment() { this.count++ },
  })
  return state
}
// Destructuring reactive objects loses reactivity:
// const { count } = useCounter() // count is NOT reactive
// Use toRefs to destructure: const { count } = toRefs(useCounter())
```

### Error Handling in Composables

```ts
// Pattern: expose an error ref + retry mechanism
export function useApi<T>(endpoint: MaybeRefOrGetter<string>) {
  const data = ref<T | null>(null) as Ref<T | null>
  const error = ref<Error | null>(null)
  const isLoading = ref(false)
  const retryCount = ref(0)

  async function execute() {
    isLoading.value = true
    error.value = null
    try {
      const res = await fetch(toValue(endpoint))
      if (!res.ok) throw new Error(`HTTP ${res.status}`)
      data.value = await res.json()
      retryCount.value = 0
    } catch (e) {
      error.value = e instanceof Error ? e : new Error(String(e))
    } finally {
      isLoading.value = false
    }
  }

  async function retry() {
    retryCount.value++
    await execute()
  }

  return { data, error, isLoading, retryCount, execute, retry }
}
```

---

## Component Patterns

### `<script setup>` Syntax Reference

```vue
<script setup lang="ts">
// All top-level bindings are exposed to the template automatically.
// No need for return statements.
import { ref, computed } from 'vue'
import MyComponent from './MyComponent.vue'

const count = ref(0)
const doubled = computed(() => count.value * 2)

function increment() {
  count.value++
}
</script>

<template>
  <MyComponent />
  <button @click="increment">{{ count }} ({{ doubled }})</button>
</template>
```

### defineProps Patterns with TypeScript

```vue
<script setup lang="ts">
// Basic typed props
const props = defineProps<{
  title: string
  count?: number
  items: string[]
}>()

// With defaults using withDefaults
const props = withDefaults(
  defineProps<{
    title: string
    count?: number
    items?: string[]
    theme?: 'light' | 'dark'
  }>(),
  {
    count: 0,
    items: () => [],
    theme: 'light',
  }
)

// Reusable prop types via interface
interface CardProps {
  title: string
  subtitle?: string
  image?: string
  href?: string
  variant?: 'default' | 'outlined' | 'elevated'
}

const props = withDefaults(defineProps<CardProps>(), {
  variant: 'default',
})
</script>
```

### defineEmits Patterns with TypeScript

```vue
<script setup lang="ts">
// Typed emit declarations
const emit = defineEmits<{
  (e: 'update', id: number, value: string): void
  (e: 'delete', id: number): void
  (e: 'close'): void
}>()

// Shorthand syntax (Vue 3.3+)
const emit = defineEmits<{
  update: [id: number, value: string]
  delete: [id: number]
  close: []
}>()

// Usage
emit('update', 1, 'new value')
emit('delete', 42)
emit('close')
</script>
```

### defineModel (Vue 3.4+)

```vue
<!-- ToggleSwitch.vue -->
<script setup lang="ts">
// Basic v-model binding
const modelValue = defineModel<boolean>({ required: true })

// Named models for multiple v-model bindings
const firstName = defineModel<string>('firstName', { required: true })
const lastName = defineModel<string>('lastName', { default: '' })
</script>

<template>
  <input v-model="firstName" placeholder="First name" />
  <input v-model="lastName" placeholder="Last name" />
  <label>
    <input type="checkbox" v-model="modelValue" />
    Toggle
  </label>
</template>

<!-- Parent usage: -->
<!-- <ToggleSwitch v-model="isActive" v-model:firstName="first" v-model:lastName="last" /> -->
```

### defineExpose

```vue
<script setup lang="ts">
import { ref } from 'vue'

const count = ref(0)
const internalState = ref('secret')

function increment() {
  count.value++
}

function reset() {
  count.value = 0
}

// Only these are accessible via template ref from the parent
defineExpose({
  count,
  increment,
  reset,
})
</script>
```

### defineSlots for Typed Slots

```vue
<script setup lang="ts">
// Type-safe slot definitions
const slots = defineSlots<{
  default(props: { item: string; index: number }): any
  header(props: { title: string }): any
  footer(): any
}>()
</script>

<template>
  <div>
    <header>
      <slot name="header" :title="'My Header'" />
    </header>
    <main>
      <slot v-for="(item, index) in items" :item="item" :index="index" />
    </main>
    <footer>
      <slot name="footer" />
    </footer>
  </div>
</template>
```

### defineOptions

```vue
<script setup lang="ts">
// Set component options without a separate <script> block
defineOptions({
  name: 'MyCustomComponent',
  inheritAttrs: false,
})
</script>
```

### Slot Patterns

```vue
<!-- Default slot -->
<template>
  <div class="card">
    <slot />
  </div>
</template>

<!-- Named slots -->
<template>
  <div class="layout">
    <header><slot name="header" /></header>
    <main><slot /></main>
    <footer><slot name="footer" /></footer>
  </div>
</template>

<!-- Scoped slots (passing data to the parent) -->
<template>
  <ul>
    <li v-for="item in items" :key="item.id">
      <slot name="item" :item="item" :isActive="item.id === activeId">
        {{ item.name }} <!-- fallback content -->
      </slot>
    </li>
  </ul>
</template>

<!-- Parent consuming scoped slot -->
<template>
  <ItemList :items="todos">
    <template #item="{ item, isActive }">
      <span :class="{ bold: isActive }">{{ item.name }}</span>
    </template>
  </ItemList>
</template>

<!-- Dynamic slot names -->
<template>
  <template v-for="section in sections" :key="section">
    <slot :name="section" />
  </template>
</template>
```

### Compound Component Pattern with Provide/Inject

```vue
<!-- Tabs.vue -->
<script setup lang="ts">
import { ref, provide } from 'vue'
import type { InjectionKey, Ref } from 'vue'

export interface TabsContext {
  activeTab: Ref<string>
  setActive: (id: string) => void
}

export const TabsKey: InjectionKey<TabsContext> = Symbol('Tabs')

const activeTab = ref('')

function setActive(id: string) {
  activeTab.value = id
}

provide(TabsKey, { activeTab, setActive })
</script>

<template>
  <div class="tabs" role="tablist">
    <slot />
  </div>
</template>

<!-- Tab.vue -->
<script setup lang="ts">
import { inject, computed } from 'vue'
import { TabsKey } from './Tabs.vue'

const props = defineProps<{ id: string; label: string }>()
const tabs = inject(TabsKey)!

const isActive = computed(() => tabs.activeTab.value === props.id)
</script>

<template>
  <button
    role="tab"
    :aria-selected="isActive"
    @click="tabs.setActive(id)"
  >
    {{ label }}
  </button>
  <div v-if="isActive" role="tabpanel">
    <slot />
  </div>
</template>
```

### Renderless Component Pattern

```vue
<!-- RenderlessCounter.vue -->
<script setup lang="ts">
import { ref } from 'vue'

const count = ref(0)
const increment = () => count.value++
const decrement = () => count.value--
const reset = () => (count.value = 0)
</script>

<template>
  <slot
    :count="count"
    :increment="increment"
    :decrement="decrement"
    :reset="reset"
  />
</template>

<!-- Usage: caller provides all rendering -->
<!-- <RenderlessCounter v-slot="{ count, increment, decrement }">
  <p>{{ count }}</p>
  <button @click="increment">+</button>
  <button @click="decrement">-</button>
</RenderlessCounter> -->
```

---

## Provide/Inject Patterns

### Basic Provide/Inject

```vue
<!-- Parent.vue -->
<script setup lang="ts">
import { provide, ref } from 'vue'

const theme = ref<'light' | 'dark'>('light')
provide('theme', theme)
</script>

<!-- Child.vue (any depth) -->
<script setup lang="ts">
import { inject } from 'vue'
import type { Ref } from 'vue'

const theme = inject<Ref<'light' | 'dark'>>('theme', ref('light'))
</script>
```

### InjectionKey for Type Safety

```ts
// keys.ts
import type { InjectionKey, Ref } from 'vue'

export interface UserContext {
  user: Ref<User | null>
  login: (credentials: Credentials) => Promise<void>
  logout: () => void
}

export const UserKey: InjectionKey<UserContext> = Symbol('UserContext')
```

```vue
<!-- Provider.vue -->
<script setup lang="ts">
import { provide, ref } from 'vue'
import { UserKey } from './keys'

const user = ref<User | null>(null)
async function login(creds: Credentials) { /* ... */ }
function logout() { user.value = null }

provide(UserKey, { user, login, logout })
</script>

<!-- Consumer.vue -->
<script setup lang="ts">
import { inject } from 'vue'
import { UserKey } from './keys'

// Fully typed: UserContext | undefined
const ctx = inject(UserKey)

// With required assertion (throws if missing):
function useRequiredInject<T>(key: InjectionKey<T>, errorMsg?: string): T {
  const value = inject(key)
  if (value === undefined) {
    throw new Error(errorMsg ?? `Missing injection: ${String(key)}`)
  }
  return value
}

const userCtx = useRequiredInject(UserKey)
</script>
```

### Plugin-Style Injection

```ts
// plugins/notification.ts
import type { App, InjectionKey, Ref } from 'vue'

interface Notification {
  id: number
  message: string
  type: 'info' | 'success' | 'error'
}

interface NotificationService {
  notifications: Ref<Notification[]>
  notify: (msg: string, type?: Notification['type']) => void
  dismiss: (id: number) => void
}

export const NotificationKey: InjectionKey<NotificationService> =
  Symbol('NotificationService')

export function createNotificationPlugin() {
  return {
    install(app: App) {
      const notifications = ref<Notification[]>([])
      let nextId = 0

      const service: NotificationService = {
        notifications,
        notify(msg, type = 'info') {
          notifications.value.push({ id: nextId++, message: msg, type })
        },
        dismiss(id) {
          notifications.value = notifications.value.filter(n => n.id !== id)
        },
      }

      app.provide(NotificationKey, service)
    },
  }
}
```

### App-Level Provide

```ts
// main.ts
import { createApp } from 'vue'
import App from './App.vue'

const app = createApp(App)

// Available to all components via inject('appVersion')
app.provide('appVersion', '2.1.0')
app.provide('apiBaseUrl', import.meta.env.VITE_API_URL)

app.mount('#app')
```

### Nested Provide Override

```vue
<!-- GrandParent.vue — provides theme: 'light' -->
<script setup lang="ts">
import { provide, ref } from 'vue'
provide('theme', ref('light'))
</script>

<!-- Parent.vue — overrides theme for its subtree -->
<script setup lang="ts">
import { provide, ref } from 'vue'
provide('theme', ref('dark'))
</script>

<!-- Child.vue — receives 'dark' from nearest provider -->
<script setup lang="ts">
const theme = inject('theme') // 'dark'
</script>
```

---

## Pinia Store Patterns

### Setup Store (Preferred)

```ts
// stores/useAuthStore.ts
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useAuthStore = defineStore('auth', () => {
  // State
  const user = ref<User | null>(null)
  const token = ref<string | null>(localStorage.getItem('token'))
  const isLoading = ref(false)

  // Getters
  const isAuthenticated = computed(() => !!token.value)
  const displayName = computed(() => user.value?.name ?? 'Guest')

  // Actions
  async function login(credentials: { email: string; password: string }) {
    isLoading.value = true
    try {
      const res = await api.post('/auth/login', credentials)
      token.value = res.data.token
      user.value = res.data.user
      localStorage.setItem('token', res.data.token)
    } finally {
      isLoading.value = false
    }
  }

  function logout() {
    user.value = null
    token.value = null
    localStorage.removeItem('token')
  }

  return { user, token, isLoading, isAuthenticated, displayName, login, logout }
})
```

### Options Store

```ts
// stores/useCounterStore.ts
import { defineStore } from 'pinia'

export const useCounterStore = defineStore('counter', {
  state: () => ({
    count: 0,
    history: [] as number[],
  }),

  getters: {
    doubled: (state) => state.count * 2,
    lastEntry: (state) => state.history.at(-1) ?? 0,
  },

  actions: {
    increment() {
      this.history.push(this.count)
      this.count++
    },
    async fetchCount() {
      const res = await fetch('/api/count')
      this.count = await res.json()
    },
  },
})
```

### Composing Stores

```ts
// stores/useCartStore.ts
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { useAuthStore } from './useAuthStore'
import { useProductStore } from './useProductStore'

export const useCartStore = defineStore('cart', () => {
  const auth = useAuthStore()
  const products = useProductStore()

  const items = ref<CartItem[]>([])

  const total = computed(() =>
    items.value.reduce((sum, item) => {
      const product = products.getById(item.productId)
      return sum + (product?.price ?? 0) * item.quantity
    }, 0)
  )

  async function checkout() {
    if (!auth.isAuthenticated) throw new Error('Must be logged in')
    await api.post('/orders', { items: items.value })
    items.value = []
  }

  return { items, total, checkout }
})
```

### Store Plugins

```ts
// plugins/piniaLogger.ts
import type { PiniaPlugin } from 'pinia'

export const piniaLoggerPlugin: PiniaPlugin = ({ store }) => {
  store.$onAction(({ name, args, after, onError }) => {
    console.log(`[Pinia] ${store.$id}.${name}`, args)

    after((result) => {
      console.log(`[Pinia] ${store.$id}.${name} completed`, result)
    })

    onError((error) => {
      console.error(`[Pinia] ${store.$id}.${name} failed`, error)
    })
  })
}

// main.ts
import { createPinia } from 'pinia'
import { piniaLoggerPlugin } from './plugins/piniaLogger'

const pinia = createPinia()
pinia.use(piniaLoggerPlugin)
```

### Store Subscriptions

```ts
const store = useAuthStore()

// Watch state changes
store.$subscribe(
  (mutation, state) => {
    console.log('Mutation type:', mutation.type) // 'direct' | 'patch object' | 'patch function'
    console.log('Store ID:', mutation.storeId)
    localStorage.setItem('auth', JSON.stringify(state))
  },
  { detached: true } // survives component unmount
)

// Watch actions
store.$onAction(({ name, store, args, after, onError }) => {
  const start = Date.now()

  after((result) => {
    console.log(`${name} took ${Date.now() - start}ms`)
  })

  onError((error) => {
    console.error(`${name} failed after ${Date.now() - start}ms`)
  })
})
```

### Reset Store State

```ts
// Options stores: built-in $reset()
const counter = useCounterStore()
counter.$reset()

// Setup stores: implement your own reset
export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  const token = ref<string | null>(null)

  function $reset() {
    user.value = null
    token.value = null
  }

  return { user, token, $reset }
})
```

### SSR-Safe Stores

```ts
// In Nuxt 3 or any SSR context, always use the pinia instance
// from the current request to avoid cross-request state pollution.

// Nuxt plugin (auto-available via useNuxtApp):
export const useAuthStore = defineStore('auth', () => {
  // useState is SSR-safe in Nuxt
  const user = ref<User | null>(null)
  return { user }
})

// In a Nuxt page:
// <script setup>
// const auth = useAuthStore()  // automatically uses the correct pinia instance
// </script>
```

### Testing Stores

```ts
// stores/__tests__/useAuthStore.test.ts
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useAuthStore } from '../useAuthStore'

describe('useAuthStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('starts unauthenticated', () => {
    const store = useAuthStore()
    expect(store.isAuthenticated).toBe(false)
    expect(store.user).toBeNull()
  })

  it('logs in successfully', async () => {
    vi.spyOn(global, 'fetch').mockResolvedValueOnce(
      new Response(JSON.stringify({ token: 'abc', user: { name: 'Alice' } }))
    )

    const store = useAuthStore()
    await store.login({ email: 'a@b.com', password: '123' })
    expect(store.isAuthenticated).toBe(true)
    expect(store.displayName).toBe('Alice')
  })
})
```

---

## Nuxt 3 Patterns

### File-Based Routing Conventions

| File Path | Generated Route | Notes |
|---|---|---|
| `pages/index.vue` | `/` | Home page |
| `pages/about.vue` | `/about` | Static route |
| `pages/blog/index.vue` | `/blog` | Nested index |
| `pages/blog/[slug].vue` | `/blog/:slug` | Dynamic param |
| `pages/blog/[...slug].vue` | `/blog/:slug(.*)` | Catch-all |
| `pages/users/[id]/posts.vue` | `/users/:id/posts` | Nested dynamic |
| `pages/[[optional]].vue` | `/:optional?` | Optional param |
| `pages/products.vue` + `pages/products/[id].vue` | Nested route with `<NuxtPage>` | Parent layout |

### Data Fetching Comparison

| Feature | `useFetch` | `useAsyncData` | `$fetch` |
|---|---|---|---|
| SSR support | Yes | Yes | Yes (but no SSR dedup) |
| Auto deduplication | Yes | Yes | No |
| Caching | Yes (key-based) | Yes (key-based) | No |
| Watch reactive sources | `watch` option | `watch` option | Manual |
| Request interceptors | Via `onRequest`, `onResponse` | Custom handler | Via `onRequest`, `onResponse` |
| Lazy loading | `lazy: true` | `lazy: true` | N/A |
| Best for | API routes / external APIs | Complex async operations | Event handlers, mutations |

```vue
<script setup lang="ts">
// useFetch — most common, wraps useAsyncData + $fetch
const { data: posts, pending, error, refresh } = await useFetch('/api/posts', {
  query: { page: 1 },
  pick: ['id', 'title'],       // only serialize these fields
  transform: (data) => data.results,
  watch: [page],               // re-fetch when page changes
  default: () => [],
})

// useAsyncData — custom async logic
const { data: combined } = await useAsyncData('dashboard', async () => {
  const [users, stats] = await Promise.all([
    $fetch('/api/users'),
    $fetch('/api/stats'),
  ])
  return { users, stats }
})

// $fetch — use in event handlers (not SSR data fetching)
async function createPost(title: string) {
  const post = await $fetch('/api/posts', {
    method: 'POST',
    body: { title },
  })
  await refreshNuxtData('posts')
}
</script>
```

### Server Routes Structure

```
server/
  api/
    users/
      index.get.ts        GET  /api/users
      index.post.ts       POST /api/users
      [id].get.ts          GET  /api/users/:id
      [id].put.ts          PUT  /api/users/:id
      [id].delete.ts       DELETE /api/users/:id
  middleware/
    auth.ts               Server middleware (runs on every request)
  plugins/
    db.ts                 Server plugin (runs on server startup)
  utils/
    db.ts                 Server-only utility (auto-imported)
```

```ts
// server/api/users/index.get.ts
export default defineEventHandler(async (event) => {
  const query = getQuery(event) // { page?: string, limit?: string }
  const users = await db.user.findMany({
    skip: Number(query.page ?? 0) * Number(query.limit ?? 10),
    take: Number(query.limit ?? 10),
  })
  return users
})

// server/api/users/index.post.ts
export default defineEventHandler(async (event) => {
  const body = await readBody(event)
  const user = await db.user.create({ data: body })
  setResponseStatus(event, 201)
  return user
})

// server/api/users/[id].get.ts
export default defineEventHandler(async (event) => {
  const id = getRouterParam(event, 'id')
  const user = await db.user.findUnique({ where: { id: Number(id) } })
  if (!user) throw createError({ statusCode: 404, message: 'User not found' })
  return user
})
```

### Middleware Types

```ts
// middleware/auth.ts — Named middleware (must be applied explicitly)
export default defineNuxtRouteMiddleware((to, from) => {
  const auth = useAuthStore()
  if (!auth.isAuthenticated) {
    return navigateTo('/login')
  }
})

// middleware/analytics.global.ts — Global middleware (runs on every route)
export default defineNuxtRouteMiddleware((to, from) => {
  trackPageView(to.fullPath)
})

// Inline middleware in a page:
// <script setup>
// definePageMeta({
//   middleware: [
//     function (to, from) {
//       if (to.params.id === '0') return abortNavigation()
//     },
//     'auth', // also apply named middleware
//   ],
// })
// </script>
```

### Plugin Types

```ts
// plugins/api.ts — Universal plugin (runs on both client and server)
export default defineNuxtPlugin((nuxtApp) => {
  const api = $fetch.create({
    baseURL: useRuntimeConfig().public.apiBase,
    onRequest({ options }) {
      const token = useCookie('token')
      if (token.value) {
        options.headers = {
          ...options.headers,
          Authorization: `Bearer ${token.value}`,
        }
      }
    },
  })

  return { provide: { api } }
})

// plugins/chart.client.ts — Client-only plugin (file suffix .client.ts)
export default defineNuxtPlugin(async () => {
  const { Chart } = await import('chart.js')
  Chart.register(/* ... */)
})

// plugins/db.server.ts — Server-only plugin (file suffix .server.ts)
export default defineNuxtPlugin(() => {
  const db = new PrismaClient()
  return { provide: { db } }
})
```

### Runtime Config

```ts
// nuxt.config.ts
export default defineNuxtConfig({
  runtimeConfig: {
    // Private keys (server only, never exposed to client)
    databaseUrl: process.env.DATABASE_URL,
    secretKey: process.env.SECRET_KEY,

    // Public keys (exposed to client via public sub-object)
    public: {
      apiBase: process.env.NUXT_PUBLIC_API_BASE || 'http://localhost:3000',
      appName: 'My App',
    },
  },
})
```

```vue
<!-- In components / pages (client + server) -->
<script setup lang="ts">
const config = useRuntimeConfig()
// config.public.apiBase — accessible
// config.secretKey — undefined on client, accessible on server
</script>
```

```ts
// In server routes (server only)
export default defineEventHandler((event) => {
  const config = useRuntimeConfig(event)
  // config.databaseUrl — accessible
  // config.public.apiBase — accessible
})
```

### useHead and useSeoMeta

```vue
<script setup lang="ts">
// useHead — full control over <head>
useHead({
  title: 'Product Page',
  meta: [
    { name: 'description', content: 'Buy this product' },
  ],
  link: [
    { rel: 'canonical', href: 'https://example.com/product/1' },
  ],
  script: [
    { src: 'https://analytics.example.com/script.js', defer: true },
  ],
})

// useSeoMeta — type-safe SEO meta tags
useSeoMeta({
  title: 'Product Page',
  description: 'Buy this product',
  ogTitle: 'Product Page',
  ogDescription: 'Buy this product',
  ogImage: 'https://example.com/og.png',
  ogUrl: 'https://example.com/product/1',
  twitterCard: 'summary_large_image',
  twitterTitle: 'Product Page',
  twitterDescription: 'Buy this product',
  twitterImage: 'https://example.com/og.png',
})
</script>
```

### useState for SSR-Safe Shared State

```vue
<script setup lang="ts">
// useState is SSR-safe and shared across components.
// It replaces the need for Pinia in simple cases within Nuxt.
const counter = useState<number>('counter', () => 0)
const user = useState<User | null>('user', () => null)

// Can also be wrapped in a composable:
function useCounter() {
  const count = useState<number>('counter', () => 0)
  const increment = () => count.value++
  return { count, increment }
}
</script>
```

### Nuxt Layers and Extends

```ts
// nuxt.config.ts — extend from a layer
export default defineNuxtConfig({
  extends: [
    '../base-layer',               // local path
    'github:org/nuxt-layer',       // GitHub
    '@my-company/nuxt-base-layer', // npm package
  ],
})

// The layer provides its own:
// - components/
// - composables/
// - layouts/
// - pages/
// - plugins/
// - server/
// - nuxt.config.ts (merged)
// - app.config.ts (merged)
```

---

## Vue Router Patterns

### Route Definition Patterns

```ts
// router/index.ts
import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    component: () => import('@/layouts/DefaultLayout.vue'),
    children: [
      {
        path: '',
        name: 'home',
        component: () => import('@/pages/Home.vue'),
      },
      {
        path: 'about',
        name: 'about',
        component: () => import('@/pages/About.vue'),
      },
    ],
  },
  {
    path: '/dashboard',
    component: () => import('@/layouts/DashboardLayout.vue'),
    meta: { requiresAuth: true },
    children: [
      {
        path: '',
        name: 'dashboard',
        component: () => import('@/pages/Dashboard.vue'),
      },
      {
        path: 'settings',
        name: 'dashboard-settings',
        component: () => import('@/pages/DashboardSettings.vue'),
      },
    ],
  },
  {
    path: '/users/:id(\\d+)',
    name: 'user-detail',
    component: () => import('@/pages/UserDetail.vue'),
    props: true, // pass route params as props
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'not-found',
    component: () => import('@/pages/NotFound.vue'),
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior(to, from, savedPosition) {
    if (savedPosition) return savedPosition
    if (to.hash) return { el: to.hash }
    return { top: 0 }
  },
})

export default router
```

### Navigation Guard Patterns

```ts
// Global before guard
router.beforeEach(async (to, from) => {
  const auth = useAuthStore()

  if (to.meta.requiresAuth && !auth.isAuthenticated) {
    return { name: 'login', query: { redirect: to.fullPath } }
  }

  if (to.meta.requiresAdmin && !auth.isAdmin) {
    return { name: 'forbidden' }
  }
})

// Global after hook (no navigation control)
router.afterEach((to, from) => {
  document.title = (to.meta.title as string) ?? 'My App'
})

// Per-route guard
{
  path: '/admin',
  beforeEnter: (to, from) => {
    const auth = useAuthStore()
    if (!auth.isAdmin) return { name: 'forbidden' }
  },
  component: () => import('@/pages/Admin.vue'),
}

// In-component guard (Composition API)
<script setup lang="ts">
import { onBeforeRouteLeave, onBeforeRouteUpdate } from 'vue-router'

onBeforeRouteLeave((to, from) => {
  if (hasUnsavedChanges.value) {
    const answer = window.confirm('Discard unsaved changes?')
    if (!answer) return false
  }
})

onBeforeRouteUpdate((to, from) => {
  // Runs when the route changes but the component is reused
  // e.g., /users/1 -> /users/2
  loadUser(to.params.id as string)
})
</script>
```

### Route Meta Typing

```ts
// router/types.ts
import 'vue-router'

declare module 'vue-router' {
  interface RouteMeta {
    requiresAuth?: boolean
    requiresAdmin?: boolean
    title?: string
    transition?: string
    layout?: 'default' | 'dashboard' | 'auth'
  }
}
```

### Programmatic Navigation

```vue
<script setup lang="ts">
import { useRouter, useRoute } from 'vue-router'

const router = useRouter()
const route = useRoute()

// Push (adds to history)
router.push('/about')
router.push({ name: 'user-detail', params: { id: '42' } })
router.push({ path: '/search', query: { q: 'vue' } })

// Replace (replaces current entry)
router.replace({ name: 'login' })

// Go back/forward
router.go(-1) // back
router.go(1)  // forward
router.back() // shorthand for go(-1)

// Access current route
console.log(route.params.id)
console.log(route.query.q)
console.log(route.meta.requiresAuth)
</script>
```

### Named Routes

```ts
// Define
{ path: '/users/:id', name: 'user-detail', component: UserDetail }

// Navigate
router.push({ name: 'user-detail', params: { id: '42' } })

// In template
// <RouterLink :to="{ name: 'user-detail', params: { id: user.id } }">
//   {{ user.name }}
// </RouterLink>
```

### Nested Routes

```ts
const routes = [
  {
    path: '/settings',
    component: () => import('@/layouts/SettingsLayout.vue'),
    children: [
      { path: '', redirect: { name: 'settings-profile' } },
      { path: 'profile', name: 'settings-profile', component: () => import('@/pages/settings/Profile.vue') },
      { path: 'security', name: 'settings-security', component: () => import('@/pages/settings/Security.vue') },
      { path: 'notifications', name: 'settings-notifications', component: () => import('@/pages/settings/Notifications.vue') },
    ],
  },
]
```

```vue
<!-- SettingsLayout.vue -->
<template>
  <div class="settings-layout">
    <nav>
      <RouterLink :to="{ name: 'settings-profile' }">Profile</RouterLink>
      <RouterLink :to="{ name: 'settings-security' }">Security</RouterLink>
      <RouterLink :to="{ name: 'settings-notifications' }">Notifications</RouterLink>
    </nav>
    <main>
      <RouterView />
    </main>
  </div>
</template>
```

### Route Transitions

```vue
<!-- App.vue -->
<template>
  <RouterView v-slot="{ Component, route }">
    <Transition :name="route.meta.transition || 'fade'" mode="out-in">
      <component :is="Component" :key="route.path" />
    </Transition>
  </RouterView>
</template>

<style>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

.slide-enter-active,
.slide-leave-active {
  transition: transform 0.3s ease;
}
.slide-enter-from {
  transform: translateX(100%);
}
.slide-leave-to {
  transform: translateX(-100%);
}
</style>
```

---

## TypeScript Patterns

### Complex Prop Types

```vue
<script setup lang="ts">
interface Column<T = unknown> {
  key: keyof T & string
  label: string
  sortable?: boolean
  render?: (value: T[keyof T], row: T) => string
}

interface TableProps<T = Record<string, unknown>> {
  columns: Column<T>[]
  rows: T[]
  loading?: boolean
  selectable?: boolean
  onRowClick?: (row: T, index: number) => void
}

const props = withDefaults(defineProps<TableProps>(), {
  loading: false,
  selectable: false,
})

// Union type props
interface ButtonProps {
  variant: 'primary' | 'secondary' | 'danger' | 'ghost'
  size: 'sm' | 'md' | 'lg'
  disabled?: boolean
  loading?: boolean
  as?: 'button' | 'a' | 'router-link'
}

// Discriminated union props
type ModalProps =
  | { type: 'alert'; message: string; onConfirm: () => void }
  | { type: 'confirm'; message: string; onConfirm: () => void; onCancel: () => void }
  | { type: 'prompt'; message: string; onSubmit: (value: string) => void }
</script>
```

### Event Handler Types

```vue
<script setup lang="ts">
// Typed event handlers for native DOM events
function handleInput(event: Event) {
  const target = event.target as HTMLInputElement
  console.log(target.value)
}

function handleKeydown(event: KeyboardEvent) {
  if (event.key === 'Enter') submit()
}

function handleSubmit(event: SubmitEvent) {
  event.preventDefault()
  const formData = new FormData(event.target as HTMLFormElement)
}

function handleDrag(event: DragEvent) {
  event.dataTransfer?.setData('text/plain', 'data')
}
</script>

<template>
  <input @input="handleInput" @keydown="handleKeydown" />
  <form @submit="handleSubmit">
    <div @dragstart="handleDrag" draggable="true" />
  </form>
</template>
```

### Template Ref Types

```vue
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import type { ComponentPublicInstance } from 'vue'
import MyInput from './MyInput.vue'

// DOM element ref
const divRef = ref<HTMLDivElement | null>(null)
const canvasRef = ref<HTMLCanvasElement | null>(null)

// Component ref (access exposed properties)
const inputRef = ref<InstanceType<typeof MyInput> | null>(null)

onMounted(() => {
  divRef.value?.focus()
  canvasRef.value?.getContext('2d')
  inputRef.value?.validate() // if expose'd
})

// Array of refs (v-for)
const itemRefs = ref<HTMLLIElement[]>([])

function setItemRef(el: Element | ComponentPublicInstance | null, index: number) {
  if (el) itemRefs.value[index] = el as HTMLLIElement
}
</script>

<template>
  <div ref="divRef" tabindex="0" />
  <canvas ref="canvasRef" />
  <MyInput ref="inputRef" />
  <ul>
    <li v-for="(item, i) in items" :key="i" :ref="(el) => setItemRef(el, i)">
      {{ item }}
    </li>
  </ul>
</template>
```

### Slots Type Inference

```vue
<script setup lang="ts">
// Using defineSlots for type-safe scoped slots
const slots = defineSlots<{
  default(props: { item: Product; index: number }): any
  empty(): any
  loading(): any
}>()

// The parent now gets type checking:
// <ProductList :items="products">
//   <template #default="{ item, index }">  <!-- typed! -->
//     {{ item.name }}
//   </template>
// </ProductList>
</script>
```

### Generic Components

```vue
<!-- GenericList.vue -->
<script setup lang="ts" generic="T extends { id: string | number }">
// Vue 3.3+ generic components
const props = defineProps<{
  items: T[]
  selected?: T
}>()

const emit = defineEmits<{
  select: [item: T]
  delete: [item: T]
}>()

defineSlots<{
  default(props: { item: T; index: number }): any
  empty(): any
}>()
</script>

<template>
  <ul v-if="items.length">
    <li
      v-for="(item, index) in items"
      :key="item.id"
      :class="{ active: selected?.id === item.id }"
      @click="emit('select', item)"
    >
      <slot :item="item" :index="index" />
    </li>
  </ul>
  <div v-else>
    <slot name="empty">
      <p>No items.</p>
    </slot>
  </div>
</template>

<!-- Usage: T is inferred from the items prop -->
<!-- <GenericList :items="users" @select="handleSelect">
  <template #default="{ item }">
    {{ item.name }}  <-- item is typed as User
  </template>
</GenericList> -->
```

### Augmenting Module Declarations

```ts
// types/vue-shims.d.ts
import 'vue'

// Add global properties to all components
declare module 'vue' {
  interface ComponentCustomProperties {
    $formatDate: (date: Date | string) => string
    $formatCurrency: (amount: number, currency?: string) => string
    $api: ReturnType<typeof createApiClient>
  }
}

// Augment global components for type checking in templates
declare module 'vue' {
  interface GlobalComponents {
    RouterLink: typeof import('vue-router')['RouterLink']
    RouterView: typeof import('vue-router')['RouterView']
    BaseButton: typeof import('@/components/BaseButton.vue')['default']
    BaseInput: typeof import('@/components/BaseInput.vue')['default']
  }
}

export {} // ensure this is treated as a module
```

### Type-Safe Router (Typed Route Names)

```ts
// router/typed.ts
import type { RouteLocationRaw } from 'vue-router'

// Define a union of all route names
type RouteName =
  | 'home'
  | 'about'
  | 'user-detail'
  | 'dashboard'
  | 'dashboard-settings'
  | 'login'
  | 'not-found'

// Route params map
interface RouteParamsMap {
  'home': Record<string, never>
  'about': Record<string, never>
  'user-detail': { id: string }
  'dashboard': Record<string, never>
  'dashboard-settings': Record<string, never>
  'login': Record<string, never>
  'not-found': Record<string, never>
}

// Type-safe navigation helper
export function typedPush<T extends RouteName>(
  router: ReturnType<typeof useRouter>,
  name: T,
  params: RouteParamsMap[T],
  query?: Record<string, string>
) {
  return router.push({
    name,
    params,
    query,
  } as RouteLocationRaw)
}

// Usage:
// typedPush(router, 'user-detail', { id: '42' })
// typedPush(router, 'home', {})
// typedPush(router, 'user-detail', {})  // TS error: missing 'id'
```

---

## Anti-Patterns

| Pattern | Problem | Solution |
|---|---|---|
| Mutating props directly | Violates one-way data flow; Vue warns in dev mode | Emit an event or use `defineModel` for two-way binding |
| Using `reactive` for primitives | `reactive` only works with objects; primitives lose reactivity | Use `ref` for primitives |
| Destructuring `reactive` without `toRefs` | Destructured values are plain (non-reactive) copies | Use `toRefs(state)` or stick to `ref` |
| Accessing `ref` in template with `.value` | Template auto-unwraps refs; `.value` is redundant / incorrect | Omit `.value` in templates |
| Not using `.value` in `<script>` for refs | Accessing the wrapper instead of the inner value | Always use `.value` in script blocks |
| Reassigning a `reactive` variable | `let state = reactive({...}); state = reactive({...})` loses reactivity | Use `ref` and reassign `.value`, or use `Object.assign` |
| Putting non-serializable values in Pinia state | Breaks SSR hydration and devtools | Keep state JSON-serializable; use getters for derived values |
| Using `v-if` with `v-for` on the same element | `v-if` has higher priority (Vue 3); confusing behavior | Wrap in `<template v-for>` with `v-if` on child, or use computed to filter |
| Watching a reactive object without deep option | `watch(reactiveObj, cb)` is deep by default but `watch(() => reactiveObj.prop, cb)` is not | Be explicit: add `{ deep: true }` when watching nested getters |
| Not cleaning up side effects | Event listeners, timers, and subscriptions leak memory | Use `onUnmounted` or `onScopeDispose` for cleanup |
| Global mutable state outside stores | Creates cross-request state pollution in SSR | Use Pinia stores or Nuxt `useState` for shared state |
| Overly large components (>300 lines) | Hard to maintain, test, and reuse | Extract logic into composables and child components |
| Using `this` in Composition API | `this` is undefined in `<script setup>` | Use local variables and imported functions |
| Not providing a key on `v-for` | Vue cannot efficiently patch the DOM | Always use `:key` with a unique identifier |
| Using array index as key with reorderable lists | DOM elements are reused incorrectly after reorder | Use a stable unique ID from the data |
| Calling composables conditionally | Composables rely on the setup execution order | Always call composables at the top level of `<script setup>` |
| Circular provide/inject dependencies | Hard to debug, implicit coupling | Use a shared store or event bus instead |
| Forgetting `await` with `useFetch` / `useAsyncData` | Data is not available during SSR | Always `await` at the top level of `<script setup>` |
| Over-using watchers for derived state | Extra complexity when `computed` suffices | Use `computed` for derived values; use `watch` for side effects |

---

## File Organization

### Feature-Based Structure

```
src/
  features/
    auth/
      components/
        LoginForm.vue
        RegisterForm.vue
        AuthGuard.vue
      composables/
        useAuth.ts
        usePermissions.ts
      stores/
        useAuthStore.ts
      types/
        auth.ts
      utils/
        tokenStorage.ts
      __tests__/
        useAuth.test.ts
        useAuthStore.test.ts
    products/
      components/
        ProductCard.vue
        ProductList.vue
        ProductFilters.vue
      composables/
        useProductSearch.ts
        useCart.ts
      stores/
        useProductStore.ts
        useCartStore.ts
      types/
        product.ts
      api/
        products.ts
  components/
    base/
      BaseButton.vue
      BaseInput.vue
      BaseModal.vue
      BaseSpinner.vue
    layout/
      AppHeader.vue
      AppSidebar.vue
      AppFooter.vue
  composables/
    useLocalStorage.ts
    useFetch.ts
    useDebounce.ts
    useMediaQuery.ts
  layouts/
    DefaultLayout.vue
    DashboardLayout.vue
    AuthLayout.vue
  pages/
    index.vue
    about.vue
    login.vue
    dashboard/
      index.vue
      settings.vue
  plugins/
    api.ts
    analytics.ts
  router/
    index.ts
    guards.ts
  stores/
    index.ts
  types/
    global.d.ts
    env.d.ts
  utils/
    formatters.ts
    validators.ts
  App.vue
  main.ts
```

### Component Naming Conventions

| Convention | Example | Notes |
|---|---|---|
| Multi-word names | `UserProfile.vue` | Avoids conflict with HTML elements |
| PascalCase filenames | `BaseButton.vue` | Standard Vue style guide recommendation |
| `Base` prefix for base components | `BaseButton.vue`, `BaseInput.vue` | Generic reusable components |
| `App` prefix for singleton components | `AppHeader.vue`, `AppSidebar.vue` | Only one instance per app |
| `The` prefix (alternative singleton) | `TheNavbar.vue`, `TheFooter.vue` | Alternative naming for singletons |
| Tight coupling prefix | `TodoList.vue`, `TodoListItem.vue` | Shows parent-child relationship |
| Full words over abbreviations | `UserDashboard.vue` not `UsrDash.vue` | Clarity over brevity |

### Composable File Naming

| Convention | Example |
|---|---|
| `use` prefix, camelCase | `useAuth.ts` |
| Match export name to filename | `useAuth.ts` exports `useAuth` |
| One composable per file | `useLocalStorage.ts` |
| Group in `composables/` directory | `composables/useAuth.ts` |
| Domain grouping for large projects | `composables/auth/useLogin.ts` |
| Test file beside composable | `composables/useAuth.test.ts` |

### Store File Naming

| Convention | Example |
|---|---|
| `use...Store` suffix | `useAuthStore.ts` |
| Match Pinia ID to name | `defineStore('auth', ...)` in `useAuthStore.ts` |
| One store per file | `stores/useAuthStore.ts` |
| Group in `stores/` directory | `stores/useAuthStore.ts` |
| Domain grouping for large projects | `stores/auth/useAuthStore.ts` |
| Test file beside store | `stores/__tests__/useAuthStore.test.ts` |

---

*End of Vue 3 Patterns Quick Reference.*
