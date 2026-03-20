---
name: zustand-state
description: >
  Zustand state management — store creation, selectors, middleware (persist,
  devtools, immer), async actions, computed values, and store patterns.
  Triggers: "zustand", "zustand store", "state management", "global state",
  "zustand persist", "zustand middleware".
  NOT for: server state (use tanstack-query), form state (use react-hook-form).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Zustand State Management

## Quick Start

```bash
npm install zustand
```

## Basic Store

```typescript
import { create } from 'zustand';

interface CounterStore {
  count: number;
  increment: () => void;
  decrement: () => void;
  reset: () => void;
  incrementBy: (amount: number) => void;
}

const useCounterStore = create<CounterStore>((set) => ({
  count: 0,
  increment: () => set((state) => ({ count: state.count + 1 })),
  decrement: () => set((state) => ({ count: state.count - 1 })),
  reset: () => set({ count: 0 }),
  incrementBy: (amount) => set((state) => ({ count: state.count + amount })),
}));

// Usage:
function Counter() {
  const count = useCounterStore((state) => state.count);
  const increment = useCounterStore((state) => state.increment);
  return <button onClick={increment}>{count}</button>;
}
```

## Real-World Store Pattern

```typescript
import { create } from 'zustand';
import { devtools, persist } from 'zustand/middleware';

interface AuthStore {
  // State
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;

  // Actions
  login: (email: string, password: string) => Promise<void>;
  logout: () => void;
  updateProfile: (updates: Partial<User>) => void;
}

const useAuthStore = create<AuthStore>()(
  devtools(
    persist(
      (set, get) => ({
        user: null,
        token: null,
        isAuthenticated: false,

        login: async (email, password) => {
          const { token, user } = await api.auth.login({ email, password });
          set({ user, token, isAuthenticated: true }, false, 'auth/login');
        },

        logout: () => {
          set({ user: null, token: null, isAuthenticated: false }, false, 'auth/logout');
        },

        updateProfile: (updates) => {
          const user = get().user;
          if (!user) return;
          set({ user: { ...user, ...updates } }, false, 'auth/updateProfile');
        },
      }),
      {
        name: 'auth-storage',         // localStorage key
        partialize: (state) => ({     // only persist these fields
          token: state.token,
          user: state.user,
        }),
      }
    ),
    { name: 'AuthStore' }             // DevTools label
  )
);
```

## Selectors (Performance)

```typescript
// BAD: subscribes to entire store — re-renders on ANY change
const state = useStore();

// GOOD: subscribe to specific slice — only re-renders when count changes
const count = useStore((state) => state.count);

// GOOD: multiple values with shallow comparison
import { useShallow } from 'zustand/react/shallow';

const { name, email } = useStore(
  useShallow((state) => ({ name: state.user?.name, email: state.user?.email }))
);

// GOOD: derived/computed value
const isAdmin = useStore((state) => state.user?.role === 'admin');
const itemCount = useStore((state) => state.items.length);
const totalPrice = useStore((state) =>
  state.cart.reduce((sum, item) => sum + item.price * item.quantity, 0)
);
```

## Middleware

### Persist

```typescript
import { persist, createJSONStorage } from 'zustand/middleware';

const useStore = create<MyStore>()(
  persist(
    (set) => ({
      // state and actions
    }),
    {
      name: 'my-store',                           // storage key
      storage: createJSONStorage(() => localStorage), // default
      // storage: createJSONStorage(() => sessionStorage),  // session only

      // Only persist certain fields
      partialize: (state) => ({
        theme: state.theme,
        language: state.language,
        // Don't persist: isLoading, error, temporary UI state
      }),

      // Version + migration
      version: 2,
      migrate: (persistedState: any, version: number) => {
        if (version === 0) {
          // Migration from v0 to v1
          persistedState.theme = persistedState.darkMode ? 'dark' : 'light';
          delete persistedState.darkMode;
        }
        if (version === 1) {
          // Migration from v1 to v2
          persistedState.language = persistedState.locale ?? 'en';
          delete persistedState.locale;
        }
        return persistedState;
      },

      // Rehydration callback
      onRehydrateStorage: () => (state) => {
        console.log('Hydration finished');
      },
    }
  )
);

// Check hydration status
const hasHydrated = useStore.persist.hasHydrated();
```

### Immer

```typescript
import { immer } from 'zustand/middleware/immer';

interface TodoStore {
  todos: Todo[];
  addTodo: (text: string) => void;
  toggleTodo: (id: string) => void;
  updateTodo: (id: string, updates: Partial<Todo>) => void;
  removeTodo: (id: string) => void;
}

const useTodoStore = create<TodoStore>()(
  immer((set) => ({
    todos: [],

    addTodo: (text) => set((state) => {
      state.todos.push({ id: crypto.randomUUID(), text, done: false });
    }),

    toggleTodo: (id) => set((state) => {
      const todo = state.todos.find((t) => t.id === id);
      if (todo) todo.done = !todo.done;
    }),

    updateTodo: (id, updates) => set((state) => {
      const todo = state.todos.find((t) => t.id === id);
      if (todo) Object.assign(todo, updates);
    }),

    removeTodo: (id) => set((state) => {
      state.todos = state.todos.filter((t) => t.id !== id);
    }),
  }))
);
```

### DevTools

```typescript
import { devtools } from 'zustand/middleware';

const useStore = create<MyStore>()(
  devtools(
    (set) => ({
      count: 0,
      // Name the action for DevTools (third arg to set)
      increment: () => set(
        (state) => ({ count: state.count + 1 }),
        false,
        'counter/increment'   // ← shows in Redux DevTools
      ),
    }),
    {
      name: 'MyStore',        // DevTools instance name
      enabled: process.env.NODE_ENV !== 'production',
    }
  )
);
```

### Combining Middleware

```typescript
// Order: outermost wraps innermost
// devtools → persist → immer → store
const useStore = create<MyStore>()(
  devtools(
    persist(
      immer(
        (set, get) => ({
          // your store
        })
      ),
      { name: 'my-storage' }
    ),
    { name: 'MyStore' }
  )
);
```

## Async Actions

```typescript
interface ProductStore {
  products: Product[];
  isLoading: boolean;
  error: string | null;
  fetchProducts: () => Promise<void>;
  createProduct: (data: CreateProduct) => Promise<Product>;
}

const useProductStore = create<ProductStore>((set, get) => ({
  products: [],
  isLoading: false,
  error: null,

  fetchProducts: async () => {
    set({ isLoading: true, error: null });
    try {
      const products = await api.products.list();
      set({ products, isLoading: false });
    } catch (error) {
      set({ error: (error as Error).message, isLoading: false });
    }
  },

  createProduct: async (data) => {
    const product = await api.products.create(data);
    set((state) => ({ products: [...state.products, product] }));
    return product;
  },
}));
```

## Store Slices Pattern

```typescript
// Split large stores into slices
interface UserSlice {
  user: User | null;
  setUser: (user: User | null) => void;
}

interface SettingsSlice {
  theme: 'light' | 'dark';
  language: string;
  setTheme: (theme: 'light' | 'dark') => void;
  setLanguage: (lang: string) => void;
}

const createUserSlice: StateCreator<UserSlice & SettingsSlice, [], [], UserSlice> = (set) => ({
  user: null,
  setUser: (user) => set({ user }),
});

const createSettingsSlice: StateCreator<UserSlice & SettingsSlice, [], [], SettingsSlice> = (set) => ({
  theme: 'light',
  language: 'en',
  setTheme: (theme) => set({ theme }),
  setLanguage: (language) => set({ language }),
});

// Combine slices
const useAppStore = create<UserSlice & SettingsSlice>()((...a) => ({
  ...createUserSlice(...a),
  ...createSettingsSlice(...a),
}));
```

## Subscribe Outside React

```typescript
// Subscribe to changes outside React components
const unsub = useStore.subscribe(
  (state) => state.count,
  (count, prevCount) => {
    console.log(`Count changed: ${prevCount} → ${count}`);
  }
);

// Get state without subscribing
const currentCount = useStore.getState().count;

// Set state from outside React
useStore.getState().increment();
// or
useStore.setState({ count: 42 });
```

## Testing

```typescript
import { act, renderHook } from '@testing-library/react';

// Reset store between tests
beforeEach(() => {
  useStore.setState({ count: 0, items: [] });
});

test('increment increases count', () => {
  const { result } = renderHook(() => useStore((s) => ({ count: s.count, increment: s.increment })));

  act(() => result.current.increment());

  expect(result.current.count).toBe(1);
});

// Or test the store directly (no React needed)
test('store actions work correctly', () => {
  const { increment, getState } = useStore;
  expect(getState().count).toBe(0);
  increment();
  expect(getState().count).toBe(1);
});
```

## Gotchas

1. **Always use selectors.** `useStore()` without a selector subscribes to the entire store. Every state change causes a re-render. Always select only what you need.

2. **Shallow equality for objects.** If your selector returns a new object, use `useShallow` or `shallow` from `zustand/shallow`. Otherwise the component re-renders every time.

3. **Don't put server data in Zustand.** API responses, cached data, pagination — use TanStack Query for this. Zustand is for client-only state.

4. **Middleware order matters.** `devtools(persist(immer(store)))` — devtools wraps persist wraps immer. Put devtools outermost to see middleware actions.

5. **Persist version and migrate.** If you change the persisted shape, bump `version` and add a `migrate` function. Otherwise users get stale/broken localStorage data.

6. **`get()` vs `set()`.** Use `get()` to read current state inside actions. Don't destructure state in the action definition — it captures the initial value.
