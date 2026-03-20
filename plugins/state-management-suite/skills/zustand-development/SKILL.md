---
name: zustand-development
description: >
  Zustand state management — lightweight stores, selectors, middleware,
  persistence, devtools, async actions, and TypeScript patterns.
  Triggers: "zustand", "zustand store", "lightweight state",
  "react state management", "zustand middleware", "zustand persist".
  NOT for: Redux/RTK (use redux-toolkit skill).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Zustand — Lightweight State Management

## Setup

```bash
npm install zustand
```

## Basic Store

```typescript
// src/stores/counter.store.ts
import { create } from "zustand";

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
  const count = useCounterStore((s) => s.count);
  const increment = useCounterStore((s) => s.increment);

  return <button onClick={increment}>{count}</button>;
}
```

## Real-World Store Pattern

```typescript
// src/stores/auth.store.ts
import { create } from "zustand";
import { persist, devtools } from "zustand/middleware";
import { immer } from "zustand/middleware/immer";

interface User {
  id: string;
  name: string;
  email: string;
  role: "admin" | "user";
}

interface AuthState {
  user: User | null;
  token: string | null;
  isLoading: boolean;
  error: string | null;

  // Actions
  login: (email: string, password: string) => Promise<void>;
  logout: () => void;
  updateProfile: (updates: Partial<User>) => void;
}

export const useAuthStore = create<AuthState>()(
  devtools(
    persist(
      immer((set, get) => ({
        user: null,
        token: null,
        isLoading: false,
        error: null,

        login: async (email, password) => {
          set({ isLoading: true, error: null });
          try {
            const res = await fetch("/api/auth/login", {
              method: "POST",
              headers: { "Content-Type": "application/json" },
              body: JSON.stringify({ email, password }),
            });

            if (!res.ok) {
              const data = await res.json();
              set({ error: data.message, isLoading: false });
              return;
            }

            const { user, token } = await res.json();
            set({ user, token, isLoading: false });
          } catch (err) {
            set({ error: "Network error", isLoading: false });
          }
        },

        logout: () => {
          set({ user: null, token: null, error: null });
        },

        updateProfile: (updates) => {
          set((state) => {
            if (state.user) {
              Object.assign(state.user, updates); // immer allows mutation
            }
          });
        },
      })),
      {
        name: "auth-storage",
        partialize: (state) => ({
          user: state.user,
          token: state.token,
        }), // Only persist user and token, not loading/error
      }
    ),
    { name: "AuthStore" } // DevTools label
  )
);
```

## Selectors (Prevent Re-renders)

```typescript
// BAD: Re-renders on ANY state change
function UserProfile() {
  const store = useAuthStore(); // Subscribes to entire store
  return <span>{store.user?.name}</span>;
}

// GOOD: Only re-renders when user.name changes
function UserProfile() {
  const name = useAuthStore((s) => s.user?.name);
  return <span>{name}</span>;
}

// GOOD: Multiple values with shallow comparison
import { useShallow } from "zustand/react/shallow";

function UserCard() {
  const { name, email, role } = useAuthStore(
    useShallow((s) => ({
      name: s.user?.name,
      email: s.user?.email,
      role: s.user?.role,
    }))
  );

  return (
    <div>
      <h2>{name}</h2>
      <p>{email}</p>
      <span>{role}</span>
    </div>
  );
}
```

## Computed / Derived Values

```typescript
// src/stores/cart.store.ts
interface CartState {
  items: CartItem[];
  addItem: (item: CartItem) => void;
  removeItem: (id: string) => void;
  clearCart: () => void;
}

export const useCartStore = create<CartState>((set) => ({
  items: [],
  addItem: (item) =>
    set((state) => {
      const existing = state.items.find((i) => i.id === item.id);
      if (existing) {
        return {
          items: state.items.map((i) =>
            i.id === item.id ? { ...i, quantity: i.quantity + 1 } : i
          ),
        };
      }
      return { items: [...state.items, { ...item, quantity: 1 }] };
    }),
  removeItem: (id) =>
    set((state) => ({ items: state.items.filter((i) => i.id !== id) })),
  clearCart: () => set({ items: [] }),
}));

// Derived selectors (computed outside the store)
export const useCartTotal = () =>
  useCartStore((s) => s.items.reduce((sum, item) => sum + item.price * item.quantity, 0));

export const useCartItemCount = () =>
  useCartStore((s) => s.items.reduce((sum, item) => sum + item.quantity, 0));

export const useCartItem = (id: string) =>
  useCartStore((s) => s.items.find((item) => item.id === id));
```

## Slices Pattern (Large Stores)

```typescript
// src/stores/slices/user.slice.ts
import { StateCreator } from "zustand";

export interface UserSlice {
  user: User | null;
  setUser: (user: User | null) => void;
}

export const createUserSlice: StateCreator<
  UserSlice & NotificationSlice, // Combined store type
  [],
  [],
  UserSlice
> = (set) => ({
  user: null,
  setUser: (user) => set({ user }),
});
```

```typescript
// src/stores/slices/notification.slice.ts
export interface NotificationSlice {
  notifications: Notification[];
  addNotification: (n: Notification) => void;
  dismissNotification: (id: string) => void;
}

export const createNotificationSlice: StateCreator<
  UserSlice & NotificationSlice,
  [],
  [],
  NotificationSlice
> = (set) => ({
  notifications: [],
  addNotification: (n) =>
    set((state) => ({ notifications: [...state.notifications, n] })),
  dismissNotification: (id) =>
    set((state) => ({
      notifications: state.notifications.filter((n) => n.id !== id),
    })),
});
```

```typescript
// src/stores/app.store.ts — combine slices
import { create } from "zustand";
import { createUserSlice, UserSlice } from "./slices/user.slice";
import { createNotificationSlice, NotificationSlice } from "./slices/notification.slice";

type AppStore = UserSlice & NotificationSlice;

export const useAppStore = create<AppStore>()((...args) => ({
  ...createUserSlice(...args),
  ...createNotificationSlice(...args),
}));
```

## Middleware

```typescript
import { create } from "zustand";
import { devtools, persist, subscribeWithSelector } from "zustand/middleware";
import { immer } from "zustand/middleware/immer";

// Middleware stacking order: devtools(persist(immer(subscribeWithSelector(...))))
const useStore = create<State>()(
  devtools(
    persist(
      immer(
        subscribeWithSelector((set, get, api) => ({
          // Store definition
        }))
      ),
      { name: "app-storage" }
    ),
    { name: "AppStore" }
  )
);
```

### Custom Middleware (Logging)

```typescript
import { StateCreator, StoreMutatorIdentifier } from "zustand";

type Logger = <
  T,
  Mps extends [StoreMutatorIdentifier, unknown][] = [],
  Mcs extends [StoreMutatorIdentifier, unknown][] = [],
>(
  f: StateCreator<T, Mps, Mcs>,
  name?: string
) => StateCreator<T, Mps, Mcs>;

const logger: Logger = (f, name) => (set, get, store) => {
  const loggedSet: typeof set = (...args) => {
    const prev = get();
    set(...args);
    const next = get();
    console.log(`[${name || "store"}]`, { prev, next });
  };
  return f(loggedSet, get, store);
};

// Usage
const useStore = create(logger((set) => ({ ... }), "MyStore"));
```

## Subscribe to Changes (Outside React)

```typescript
// Listen to specific state changes
const unsub = useAuthStore.subscribe(
  (state) => state.token,
  (token, prevToken) => {
    if (token) {
      // User logged in — set up axios interceptor
      axios.defaults.headers.common["Authorization"] = `Bearer ${token}`;
    } else {
      delete axios.defaults.headers.common["Authorization"];
    }
  }
);

// With subscribeWithSelector middleware
const unsub = useStore.subscribe(
  (state) => state.user?.role,
  (role) => console.log("Role changed:", role),
  { equalityFn: Object.is, fireImmediately: true }
);
```

## Using Store Outside React

```typescript
// In utility functions, API interceptors, etc.
const token = useAuthStore.getState().token;
useAuthStore.getState().logout();

// Set state from anywhere
useAuthStore.setState({ error: "Session expired" });
```

## Testing

```typescript
import { renderHook, act } from "@testing-library/react";
import { useCartStore } from "./cart.store";

describe("Cart Store", () => {
  beforeEach(() => {
    // Reset store between tests
    useCartStore.setState({ items: [] });
  });

  it("adds items", () => {
    const { result } = renderHook(() => useCartStore());

    act(() => {
      result.current.addItem({ id: "1", name: "Widget", price: 10, quantity: 1 });
    });

    expect(result.current.items).toHaveLength(1);
    expect(result.current.items[0].name).toBe("Widget");
  });

  it("increments quantity for existing items", () => {
    const { result } = renderHook(() => useCartStore());

    act(() => {
      result.current.addItem({ id: "1", name: "Widget", price: 10, quantity: 1 });
      result.current.addItem({ id: "1", name: "Widget", price: 10, quantity: 1 });
    });

    expect(result.current.items).toHaveLength(1);
    expect(result.current.items[0].quantity).toBe(2);
  });
});
```

## Gotchas

1. **Don't destructure the store call** — `const { count, increment } = useStore()` subscribes to the ENTIRE store. Always use selectors: `useStore(s => s.count)`. This is the #1 performance mistake.

2. **`useShallow` is needed for object selectors** — `useStore(s => ({ a: s.a, b: s.b }))` creates a new object every render, causing infinite re-renders. Wrap with `useShallow()` or use individual selectors.

3. **Persist middleware hydration is async** — On page load, the store initializes with defaults FIRST, then hydrates from storage. Use `onRehydrateStorage` callback or `useStore.persist.hasHydrated()` to detect when hydration completes.

4. **Immer middleware changes the `set` signature** — With immer, `set((state) => { state.count++ })` works (mutate draft). Without immer, you must return a new object: `set((state) => ({ count: state.count + 1 }))`.

5. **DevTools middleware should be outermost** — Wrap order: `devtools(persist(immer(...)))`. If devtools is inside persist, state updates won't show correctly in the Redux DevTools extension.

6. **Selectors run on every render** — Even if the selected value hasn't changed, the selector function runs. Keep selectors simple and fast. For expensive derivations, use `useMemo` on the selected value.
