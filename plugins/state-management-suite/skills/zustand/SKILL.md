---
name: zustand
description: >
  Zustand state management — store creation, selectors, actions, middleware
  (persist, immer, devtools), async actions, and TypeScript patterns.
  Triggers: "zustand", "zustand store", "zustand persist", "zustand middleware",
  "simple global state".
  NOT for: Redux (use redux-toolkit), server state (use react-query), atomic state (use jotai).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Zustand

## Quick Start

```bash
npm install zustand
```

## Basic Store

```typescript
// stores/useCounterStore.ts
import { create } from 'zustand';

interface CounterState {
  count: number;
  increment: () => void;
  decrement: () => void;
  reset: () => void;
  incrementBy: (amount: number) => void;
}

export const useCounterStore = create<CounterState>((set) => ({
  count: 0,
  increment: () => set((state) => ({ count: state.count + 1 })),
  decrement: () => set((state) => ({ count: state.count - 1 })),
  reset: () => set({ count: 0 }),
  incrementBy: (amount) => set((state) => ({ count: state.count + amount })),
}));
```

```tsx
// Usage in component
function Counter() {
  const count = useCounterStore((state) => state.count);
  const increment = useCounterStore((state) => state.increment);

  return <button onClick={increment}>{count}</button>;
}
```

## Real-World Store Pattern

```typescript
// stores/useAuthStore.ts
import { create } from 'zustand';
import { persist } from 'zustand/middleware';

interface User {
  id: string;
  email: string;
  name: string;
  role: 'user' | 'admin';
}

interface AuthState {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;

  login: (email: string, password: string) => Promise<void>;
  logout: () => void;
  updateProfile: (updates: Partial<User>) => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      user: null,
      token: null,
      isAuthenticated: false,

      login: async (email, password) => {
        const response = await fetch('/api/auth/login', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ email, password }),
        });

        if (!response.ok) {
          throw new Error('Login failed');
        }

        const { user, token } = await response.json();
        set({ user, token, isAuthenticated: true });
      },

      logout: () => {
        set({ user: null, token: null, isAuthenticated: false });
      },

      updateProfile: (updates) => {
        const currentUser = get().user;
        if (currentUser) {
          set({ user: { ...currentUser, ...updates } });
        }
      },
    }),
    {
      name: 'auth-storage',
      partialize: (state) => ({
        user: state.user,
        token: state.token,
        isAuthenticated: state.isAuthenticated,
      }),
    }
  )
);
```

## Selectors (Prevent Unnecessary Re-renders)

```typescript
// BAD — re-renders on ANY state change
const state = useAuthStore();

// GOOD — only re-renders when user changes
const user = useAuthStore((state) => state.user);

// GOOD — multiple selectors, each triggers independently
const user = useAuthStore((state) => state.user);
const isAuthenticated = useAuthStore((state) => state.isAuthenticated);

// Derived selector with shallow comparison
import { useShallow } from 'zustand/react/shallow';

const { user, isAuthenticated } = useAuthStore(
  useShallow((state) => ({
    user: state.user,
    isAuthenticated: state.isAuthenticated,
  }))
);

// Computed selector (derived state)
const isAdmin = useAuthStore((state) => state.user?.role === 'admin');
```

## Middleware

### Persist (LocalStorage / SessionStorage)

```typescript
import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';

const useSettingsStore = create<SettingsState>()(
  persist(
    (set) => ({
      theme: 'light' as const,
      language: 'en',
      setTheme: (theme) => set({ theme }),
      setLanguage: (language) => set({ language }),
    }),
    {
      name: 'settings',
      storage: createJSONStorage(() => sessionStorage), // Default: localStorage
      partialize: (state) => ({ theme: state.theme, language: state.language }),
      version: 1,
      migrate: (persisted, version) => {
        if (version === 0) {
          // Migration from v0 to v1
          return { ...persisted, language: 'en' };
        }
        return persisted as SettingsState;
      },
    }
  )
);
```

### Immer (Mutable-Style Updates)

```typescript
import { create } from 'zustand';
import { immer } from 'zustand/middleware/immer';

interface TodoState {
  todos: { id: string; text: string; done: boolean }[];
  addTodo: (text: string) => void;
  toggleTodo: (id: string) => void;
  removeTodo: (id: string) => void;
}

const useTodoStore = create<TodoState>()(
  immer((set) => ({
    todos: [],
    addTodo: (text) =>
      set((state) => {
        state.todos.push({ id: crypto.randomUUID(), text, done: false });
      }),
    toggleTodo: (id) =>
      set((state) => {
        const todo = state.todos.find((t) => t.id === id);
        if (todo) todo.done = !todo.done;
      }),
    removeTodo: (id) =>
      set((state) => {
        state.todos = state.todos.filter((t) => t.id !== id);
      }),
  }))
);
```

### DevTools

```typescript
import { create } from 'zustand';
import { devtools } from 'zustand/middleware';

const useStore = create<StoreState>()(
  devtools(
    (set) => ({
      // ... state and actions
      increment: () =>
        set(
          (state) => ({ count: state.count + 1 }),
          undefined,
          'counter/increment' // Action name in DevTools
        ),
    }),
    { name: 'MyStore' } // Store name in DevTools
  )
);
```

### Combining Middleware

```typescript
// Stack: devtools → persist → immer
const useStore = create<StoreState>()(
  devtools(
    persist(
      immer((set) => ({
        // ... your state
      })),
      { name: 'my-store' }
    ),
    { name: 'MyStore' }
  )
);
```

## Async Actions

```typescript
interface ProductState {
  products: Product[];
  isLoading: boolean;
  error: string | null;

  fetchProducts: () => Promise<void>;
  createProduct: (data: CreateProductDTO) => Promise<Product>;
}

const useProductStore = create<ProductState>((set, get) => ({
  products: [],
  isLoading: false,
  error: null,

  fetchProducts: async () => {
    set({ isLoading: true, error: null });
    try {
      const response = await fetch('/api/products');
      const products = await response.json();
      set({ products, isLoading: false });
    } catch (error) {
      set({ error: (error as Error).message, isLoading: false });
    }
  },

  createProduct: async (data) => {
    const response = await fetch('/api/products', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    });
    const product = await response.json();

    // Optimistic-style: add to existing list
    set((state) => ({ products: [...state.products, product] }));

    return product;
  },
}));
```

## Store Slices Pattern

```typescript
// slices/authSlice.ts
export interface AuthSlice {
  user: User | null;
  login: (email: string, password: string) => Promise<void>;
  logout: () => void;
}

export const createAuthSlice: StateCreator<
  AuthSlice & UISlice,  // Combined type
  [],
  [],
  AuthSlice
> = (set) => ({
  user: null,
  login: async (email, password) => { /* ... */ },
  logout: () => set({ user: null }),
});

// slices/uiSlice.ts
export interface UISlice {
  sidebarOpen: boolean;
  toggleSidebar: () => void;
}

export const createUISlice: StateCreator<
  AuthSlice & UISlice,
  [],
  [],
  UISlice
> = (set) => ({
  sidebarOpen: true,
  toggleSidebar: () => set((s) => ({ sidebarOpen: !s.sidebarOpen })),
});

// stores/useAppStore.ts
import { create } from 'zustand';
import { createAuthSlice } from './slices/authSlice';
import { createUISlice } from './slices/uiSlice';

export const useAppStore = create<AuthSlice & UISlice>()((...a) => ({
  ...createAuthSlice(...a),
  ...createUISlice(...a),
}));
```

## Outside React (Vanilla)

```typescript
// Access store outside components
const { user, logout } = useAuthStore.getState();

// Subscribe to changes
const unsubscribe = useAuthStore.subscribe(
  (state) => console.log('State changed:', state)
);

// Subscribe to specific slice
const unsubscribe = useAuthStore.subscribe(
  (state) => state.user,
  (user, prevUser) => {
    console.log('User changed:', prevUser, '->', user);
  },
  { equalityFn: Object.is }
);

// Use in API client (e.g., attach token to fetch)
async function apiFetch(url: string, options?: RequestInit) {
  const token = useAuthStore.getState().token;
  return fetch(url, {
    ...options,
    headers: {
      ...options?.headers,
      Authorization: token ? `Bearer ${token}` : '',
    },
  });
}
```

## Testing

```typescript
import { act, renderHook } from '@testing-library/react';
import { useCounterStore } from './useCounterStore';

// Reset store between tests
beforeEach(() => {
  useCounterStore.setState({ count: 0 });
});

test('increment increases count', () => {
  const { result } = renderHook(() => useCounterStore());

  act(() => {
    result.current.increment();
  });

  expect(result.current.count).toBe(1);
});

test('set initial state for test', () => {
  useCounterStore.setState({ count: 10 });

  const { result } = renderHook(() =>
    useCounterStore((state) => state.count)
  );

  expect(result.current).toBe(10);
});
```

## Gotchas

1. **Always use selectors.** `const state = useStore()` subscribes to everything. Always select what you need: `const count = useStore(s => s.count)`.

2. **`set` does a shallow merge.** `set({ count: 1 })` keeps all other state. But nested objects need spreading: `set((s) => ({ user: { ...s.user, name: 'new' } }))` — or use immer middleware.

3. **Don't call `set` in render.** Actions should be called from event handlers or effects, not during render. This causes infinite loops.

4. **Persist hydration is async.** On first render, persisted state isn't loaded yet. Use `useStore.persist.onFinishHydration()` or check `useStore.persist.hasHydrated()`.

5. **Multiple stores > one mega store.** Split by domain (`useAuthStore`, `useCartStore`). Stores are cheap. One big store means more re-renders.
