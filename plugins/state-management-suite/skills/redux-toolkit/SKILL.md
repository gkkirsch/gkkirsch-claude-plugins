---
name: redux-toolkit
description: >
  Redux Toolkit (RTK) state management — slices, async thunks, RTK Query,
  entity adapters, middleware, selectors, and TypeScript patterns.
  Triggers: "redux", "redux toolkit", "rtk", "rtk query",
  "redux slice", "createSlice", "createAsyncThunk".
  NOT for: Lightweight state needs (use zustand-development).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Redux Toolkit (RTK)

## Setup

```bash
npm install @reduxjs/toolkit react-redux
```

## Store Configuration

```typescript
// src/store/index.ts
import { configureStore } from "@reduxjs/toolkit";
import { authReducer } from "./slices/auth.slice";
import { cartReducer } from "./slices/cart.slice";
import { apiSlice } from "./api/api.slice";

export const store = configureStore({
  reducer: {
    auth: authReducer,
    cart: cartReducer,
    [apiSlice.reducerPath]: apiSlice.reducer,
  },
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware().concat(apiSlice.middleware),
  devTools: process.env.NODE_ENV !== "production",
});

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;
```

```typescript
// src/store/hooks.ts
import { useDispatch, useSelector } from "react-redux";
import type { RootState, AppDispatch } from "./index";

export const useAppDispatch = useDispatch.withTypes<AppDispatch>();
export const useAppSelector = useSelector.withTypes<RootState>();
```

```tsx
// src/main.tsx
import { Provider } from "react-redux";
import { store } from "./store";

<Provider store={store}><App /></Provider>
```

## Slices

```typescript
// src/store/slices/auth.slice.ts
import { createSlice, createAsyncThunk, PayloadAction } from "@reduxjs/toolkit";

interface AuthState {
  user: User | null;
  token: string | null;
  status: "idle" | "loading" | "succeeded" | "failed";
  error: string | null;
}

export const login = createAsyncThunk(
  "auth/login",
  async (creds: { email: string; password: string }, { rejectWithValue }) => {
    const res = await fetch("/api/auth/login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(creds),
    });
    if (!res.ok) return rejectWithValue((await res.json()).message);
    return res.json();
  }
);

const authSlice = createSlice({
  name: "auth",
  initialState: { user: null, token: null, status: "idle", error: null } as AuthState,
  reducers: {
    logout: (state) => { state.user = null; state.token = null; },
    updateProfile: (state, action: PayloadAction<Partial<User>>) => {
      if (state.user) Object.assign(state.user, action.payload);
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(login.pending, (state) => { state.status = "loading"; state.error = null; })
      .addCase(login.fulfilled, (state, action) => {
        state.status = "succeeded";
        state.user = action.payload.user;
        state.token = action.payload.token;
      })
      .addCase(login.rejected, (state, action) => {
        state.status = "failed";
        state.error = action.payload as string;
      });
  },
});

export const { logout, updateProfile } = authSlice.actions;
export const authReducer = authSlice.reducer;
export const selectUser = (state: RootState) => state.auth.user;
export const selectIsAuthenticated = (state: RootState) => !!state.auth.token;
```

## RTK Query

```typescript
// src/store/api/api.slice.ts
import { createApi, fetchBaseQuery } from "@reduxjs/toolkit/query/react";

export const apiSlice = createApi({
  reducerPath: "api",
  baseQuery: fetchBaseQuery({
    baseUrl: "/api",
    prepareHeaders: (headers, { getState }) => {
      const token = (getState() as RootState).auth.token;
      if (token) headers.set("Authorization", `Bearer ${token}`);
      return headers;
    },
  }),
  tagTypes: ["User", "Post"],
  endpoints: (builder) => ({
    getUsers: builder.query<User[], void>({
      query: () => "/users",
      providesTags: (result) =>
        result
          ? [...result.map(({ id }) => ({ type: "User" as const, id })), { type: "User", id: "LIST" }]
          : [{ type: "User", id: "LIST" }],
    }),
    createUser: builder.mutation<User, Partial<User>>({
      query: (body) => ({ url: "/users", method: "POST", body }),
      invalidatesTags: [{ type: "User", id: "LIST" }],
    }),
    updateUser: builder.mutation<User, { id: string; updates: Partial<User> }>({
      query: ({ id, updates }) => ({ url: `/users/${id}`, method: "PATCH", body: updates }),
      async onQueryStarted({ id, updates }, { dispatch, queryFulfilled }) {
        const patch = dispatch(apiSlice.util.updateQueryData("getUsers", undefined, (draft) => {
          const user = draft.find((u) => u.id === id);
          if (user) Object.assign(user, updates);
        }));
        try { await queryFulfilled; } catch { patch.undo(); }
      },
      invalidatesTags: (_, __, { id }) => [{ type: "User", id }],
    }),
  }),
});

export const { useGetUsersQuery, useCreateUserMutation, useUpdateUserMutation } = apiSlice;
```

## Entity Adapter

```typescript
import { createSlice, createEntityAdapter } from "@reduxjs/toolkit";

const todosAdapter = createEntityAdapter<Todo>({
  sortComparer: (a, b) => b.createdAt.localeCompare(a.createdAt),
});

const todosSlice = createSlice({
  name: "todos",
  initialState: todosAdapter.getInitialState({ filter: "all" as "all" | "active" | "completed" }),
  reducers: {
    addTodo: todosAdapter.addOne,
    updateTodo: todosAdapter.updateOne,
    removeTodo: todosAdapter.removeOne,
    toggleTodo: (state, action: PayloadAction<string>) => {
      const todo = state.entities[action.payload];
      if (todo) todo.completed = !todo.completed;
    },
  },
});

const selectors = todosAdapter.getSelectors<RootState>((s) => s.todos);
export const selectAllTodos = selectors.selectAll;
export const selectTodoById = selectors.selectById;
```

## Memoized Selectors

```typescript
import { createSelector } from "@reduxjs/toolkit";

export const selectFilteredTodos = createSelector(
  [selectAllTodos, (state: RootState) => state.todos.filter],
  (todos, filter) => {
    switch (filter) {
      case "active": return todos.filter((t) => !t.completed);
      case "completed": return todos.filter((t) => t.completed);
      default: return todos;
    }
  }
);
```

## Listener Middleware

```typescript
import { createListenerMiddleware } from "@reduxjs/toolkit";

const listenerMiddleware = createListenerMiddleware();

listenerMiddleware.startListening({
  actionCreator: authSlice.actions.logout,
  effect: async (_, api) => {
    api.dispatch(apiSlice.util.resetApiState());
  },
});

// Debounced search
listenerMiddleware.startListening({
  actionCreator: setSearchQuery,
  effect: async (action, api) => {
    api.cancelActiveListeners();
    await api.delay(300);
    api.dispatch(searchApi.endpoints.search.initiate(action.payload));
  },
});
```

## Testing

```typescript
import { configureStore } from "@reduxjs/toolkit";

function createTestStore(preloadedState?: Partial<RootState>) {
  return configureStore({
    reducer: { auth: authReducer, todos: todosReducer },
    preloadedState,
  });
}

it("handles login", () => {
  const store = createTestStore();
  store.dispatch(login.fulfilled({ user, token: "jwt" }, "", { email: "", password: "" }));
  expect(selectUser(store.getState())).toEqual(user);
});
```

## Gotchas

1. **Use `useAppSelector` and `useAppDispatch`** — Typed wrappers give full TypeScript safety. Plain hooks lose types.

2. **RTK Query auto-caches** — Two components calling `useGetUsersQuery()` fire ONE request. Cache is keyed by endpoint + args.

3. **`invalidatesTags` vs `providesTags`** — Queries PROVIDE tags. Mutations INVALIDATE tags. Plan tag structure upfront.

4. **Entity adapter `selectAll` returns new array** — Wrap filtered results in `createSelector` to memoize.

5. **Immer is built-in** — Mutate state directly in `createSlice` reducers. But ONLY inside reducers.

6. **Don't mix RTK Query with thunks** for the same data. RTK Query for CRUD, thunks for non-API async work.
