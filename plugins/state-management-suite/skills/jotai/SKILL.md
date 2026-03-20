---
name: jotai
description: >
  Jotai atomic state management — primitive atoms, derived atoms, async atoms,
  atom families, persistence, and integration with React Query.
  Triggers: "jotai", "atomic state", "atom", "derived atom",
  "bottom-up state", "jotai atom".
  NOT for: large Redux-style stores (use redux-toolkit), simple global state (use zustand).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Jotai

## Quick Start

```bash
npm install jotai
```

No provider needed (uses default store). Optional provider for isolation.

## Primitive Atoms

```typescript
// atoms/counter.ts
import { atom } from 'jotai';

// Read-write atom (like useState but global)
export const countAtom = atom(0);

// Usage — exactly like useState
import { useAtom } from 'jotai';

function Counter() {
  const [count, setCount] = useAtom(countAtom);
  return <button onClick={() => setCount((c) => c + 1)}>{count}</button>;
}

// Read-only hook
import { useAtomValue } from 'jotai';
function Display() {
  const count = useAtomValue(countAtom);
  return <span>{count}</span>;
}

// Write-only hook (no re-render on change)
import { useSetAtom } from 'jotai';
function IncrementButton() {
  const setCount = useSetAtom(countAtom);
  return <button onClick={() => setCount((c) => c + 1)}>+1</button>;
}
```

## Derived Atoms

```typescript
import { atom } from 'jotai';

// Source atoms
export const itemsAtom = atom<Item[]>([]);
export const filterAtom = atom<'all' | 'active' | 'completed'>('all');
export const searchAtom = atom('');

// Derived atom (read-only, auto-updates when dependencies change)
export const filteredItemsAtom = atom((get) => {
  const items = get(itemsAtom);
  const filter = get(filterAtom);
  const search = get(searchAtom).toLowerCase();

  return items
    .filter((item) => {
      if (filter === 'active') return !item.completed;
      if (filter === 'completed') return item.completed;
      return true;
    })
    .filter((item) => item.title.toLowerCase().includes(search));
});

// Derived count
export const activeCountAtom = atom((get) => {
  return get(itemsAtom).filter((item) => !item.completed).length;
});

// Derived with write (read-write derived)
export const uppercaseAtom = atom(
  (get) => get(searchAtom).toUpperCase(),
  (get, set, newValue: string) => {
    set(searchAtom, newValue.toLowerCase());
  }
);
```

## Async Atoms

```typescript
import { atom } from 'jotai';

// Async read atom (fetches data)
export const userAtom = atom(async () => {
  const response = await fetch('/api/user');
  return response.json() as Promise<User>;
});

// Async derived (depends on another atom)
export const userPostsAtom = atom(async (get) => {
  const user = await get(userAtom);
  const response = await fetch(`/api/users/${user.id}/posts`);
  return response.json() as Promise<Post[]>;
});

// Usage with Suspense
function UserProfile() {
  const user = useAtomValue(userAtom); // Suspends until loaded

  return <div>{user.name}</div>;
}

// Wrap in Suspense boundary
function App() {
  return (
    <Suspense fallback={<Skeleton />}>
      <UserProfile />
    </Suspense>
  );
}
```

## Writable Async Atoms (CRUD)

```typescript
import { atom } from 'jotai';

interface Todo {
  id: string;
  title: string;
  completed: boolean;
}

// Base atom for todos
const todosBaseAtom = atom<Todo[]>([]);

// Async atom that fetches on first read, writable for optimistic updates
export const todosAtom = atom(
  async (get) => {
    const cached = get(todosBaseAtom);
    if (cached.length > 0) return cached;

    const response = await fetch('/api/todos');
    return response.json() as Promise<Todo[]>;
  },
  (get, set, action: { type: 'add'; todo: Todo } | { type: 'toggle'; id: string } | { type: 'remove'; id: string }) => {
    const current = get(todosBaseAtom);

    switch (action.type) {
      case 'add':
        set(todosBaseAtom, [...current, action.todo]);
        fetch('/api/todos', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(action.todo),
        });
        break;

      case 'toggle':
        set(
          todosBaseAtom,
          current.map((t) =>
            t.id === action.id ? { ...t, completed: !t.completed } : t
          )
        );
        const todo = current.find((t) => t.id === action.id);
        if (todo) {
          fetch(`/api/todos/${action.id}`, {
            method: 'PATCH',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ completed: !todo.completed }),
          });
        }
        break;

      case 'remove':
        set(todosBaseAtom, current.filter((t) => t.id !== action.id));
        fetch(`/api/todos/${action.id}`, { method: 'DELETE' });
        break;
    }
  }
);
```

## Atom Families (Dynamic Atoms)

```typescript
import { atom } from 'jotai';
import { atomFamily } from 'jotai/utils';

// Creates a unique atom for each parameter
export const userAtomFamily = atomFamily((userId: string) =>
  atom(async () => {
    const response = await fetch(`/api/users/${userId}`);
    return response.json() as Promise<User>;
  })
);

// Usage — each userId gets its own cached atom
function UserName({ userId }: { userId: string }) {
  const user = useAtomValue(userAtomFamily(userId));
  return <span>{user.name}</span>;
}

// Editable atom family
export const formFieldAtom = atomFamily((fieldName: string) =>
  atom('')
);

function FormField({ name }: { name: string }) {
  const [value, setValue] = useAtom(formFieldAtom(name));
  return <input value={value} onChange={(e) => setValue(e.target.value)} />;
}
```

## Persistence

```typescript
import { atomWithStorage } from 'jotai/utils';

// Auto-persists to localStorage
export const themeAtom = atomWithStorage<'light' | 'dark'>('theme', 'light');
export const languageAtom = atomWithStorage('language', 'en');

// sessionStorage
import { createJSONStorage } from 'jotai/utils';
export const tempDataAtom = atomWithStorage(
  'temp-data',
  null,
  createJSONStorage(() => sessionStorage)
);

// Usage — just like a regular atom, but survives page reloads
function ThemeToggle() {
  const [theme, setTheme] = useAtom(themeAtom);
  return (
    <button onClick={() => setTheme(theme === 'light' ? 'dark' : 'light')}>
      {theme}
    </button>
  );
}
```

## Utilities

```typescript
import { atom } from 'jotai';
import { atomWithReducer, atomWithDefault, loadable, selectAtom } from 'jotai/utils';

// Reducer atom (like useReducer but global)
type Action = { type: 'increment' } | { type: 'decrement' } | { type: 'reset' };

const countReducerAtom = atomWithReducer(0, (state, action: Action) => {
  switch (action.type) {
    case 'increment': return state + 1;
    case 'decrement': return state - 1;
    case 'reset': return 0;
  }
});

// Loadable (avoid Suspense, handle loading/error yourself)
const loadableUserAtom = loadable(userAtom);

function UserProfile() {
  const userLoadable = useAtomValue(loadableUserAtom);

  if (userLoadable.state === 'loading') return <Spinner />;
  if (userLoadable.state === 'hasError') return <Error error={userLoadable.error} />;

  return <div>{userLoadable.data.name}</div>;
}

// Select atom (derived slice — only re-renders when slice changes)
const userNameAtom = selectAtom(userAtom, (user) => user.name);
```

## Integration with React Query

```typescript
import { atomWithQuery, atomWithMutation } from 'jotai-tanstack-query';

// Query atom
const usersQueryAtom = atomWithQuery(() => ({
  queryKey: ['users'],
  queryFn: async () => {
    const response = await fetch('/api/users');
    return response.json() as Promise<User[]>;
  },
}));

// Mutation atom
const createUserMutationAtom = atomWithMutation(() => ({
  mutationFn: async (data: CreateUserInput) => {
    const response = await fetch('/api/users', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    });
    return response.json() as Promise<User>;
  },
}));
```

## Provider for Isolation

```tsx
import { Provider, createStore } from 'jotai';

// Isolated store (useful for testing or micro-frontends)
const myStore = createStore();

function App() {
  return (
    <Provider store={myStore}>
      <MyComponent />
    </Provider>
  );
}

// Access store outside React
myStore.get(countAtom);
myStore.set(countAtom, 42);
myStore.sub(countAtom, () => {
  console.log('Count changed:', myStore.get(countAtom));
});
```

## DevTools

```bash
npm install jotai-devtools
```

```tsx
import { DevTools } from 'jotai-devtools';
import 'jotai-devtools/styles.css';

function App() {
  return (
    <>
      <DevTools />
      <MyApp />
    </>
  );
}

// Label atoms for DevTools
const countAtom = atom(0);
countAtom.debugLabel = 'count';
```

## Gotchas

1. **Atoms are definitions, not instances.** `atom(0)` creates a config. The actual state lives in the store. Don't recreate atoms in render — define them at module level.

2. **Derived atoms auto-track dependencies.** Every `get(someAtom)` call registers a dependency. The derived atom re-computes only when dependencies change.

3. **Async atoms need Suspense.** By default, async atoms suspend. Use `loadable()` wrapper if you want to handle loading/error states manually.

4. **`atomFamily` uses reference equality.** For object parameters, the same-looking object creates a different atom. Use `atomFamily` with primitive parameters (strings, numbers) or provide a custom equality function.

5. **No provider needed by default.** Jotai uses a default store. Add a `Provider` only for isolation (testing, multiple independent widget trees).

6. **`useSetAtom` prevents re-renders.** If a component only writes (never reads), use `useSetAtom` to avoid unnecessary re-renders.
