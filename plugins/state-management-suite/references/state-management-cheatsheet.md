# State Management Cheatsheet

## Zustand Quick Reference

### Setup
```bash
npm install zustand
```

### Basic Store
```typescript
import { create } from "zustand";

const useStore = create<State>((set, get) => ({
  count: 0,
  increment: () => set((s) => ({ count: s.count + 1 })),
  reset: () => set({ count: 0 }),
  getDouble: () => get().count * 2,
}));
```

### Selectors (prevent re-renders)
```tsx
// GOOD: subscribe to single value
const count = useStore((s) => s.count);

// GOOD: multiple values with shallow
import { useShallow } from "zustand/react/shallow";
const { a, b } = useStore(useShallow((s) => ({ a: s.a, b: s.b })));

// BAD: subscribes to entire store
const store = useStore();
```

### Middleware Stack
```typescript
import { devtools, persist } from "zustand/middleware";
import { immer } from "zustand/middleware/immer";

const useStore = create<State>()(
  devtools(persist(immer((set) => ({ ... })), { name: "storage-key" }))
);
```

### Outside React
```typescript
useStore.getState().increment();
useStore.setState({ count: 0 });
useStore.subscribe((state) => console.log(state));
```

---

## Redux Toolkit Quick Reference

### Setup
```bash
npm install @reduxjs/toolkit react-redux
```

### Store
```typescript
import { configureStore } from "@reduxjs/toolkit";
const store = configureStore({
  reducer: { auth: authReducer, cart: cartReducer },
});
export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;
```

### Typed Hooks
```typescript
import { useDispatch, useSelector } from "react-redux";
export const useAppDispatch = useDispatch.withTypes<AppDispatch>();
export const useAppSelector = useSelector.withTypes<RootState>();
```

### Slice
```typescript
import { createSlice, PayloadAction } from "@reduxjs/toolkit";

const slice = createSlice({
  name: "counter",
  initialState: { count: 0 },
  reducers: {
    increment: (state) => { state.count++ },  // Immer mutation OK
    incrementBy: (state, action: PayloadAction<number>) => {
      state.count += action.payload;
    },
  },
});

export const { increment, incrementBy } = slice.actions;
export default slice.reducer;
```

### Async Thunk
```typescript
import { createAsyncThunk } from "@reduxjs/toolkit";

export const fetchUser = createAsyncThunk(
  "user/fetch",
  async (id: string, { rejectWithValue }) => {
    const res = await fetch(`/api/users/${id}`);
    if (!res.ok) return rejectWithValue("Not found");
    return res.json();
  }
);

// In slice:
extraReducers: (builder) => {
  builder
    .addCase(fetchUser.pending, (state) => { state.status = "loading" })
    .addCase(fetchUser.fulfilled, (state, action) => {
      state.status = "succeeded";
      state.user = action.payload;
    })
    .addCase(fetchUser.rejected, (state, action) => {
      state.status = "failed";
      state.error = action.payload as string;
    });
}
```

### RTK Query
```typescript
import { createApi, fetchBaseQuery } from "@reduxjs/toolkit/query/react";

const api = createApi({
  baseQuery: fetchBaseQuery({ baseUrl: "/api" }),
  tagTypes: ["User"],
  endpoints: (builder) => ({
    getUsers: builder.query<User[], void>({
      query: () => "/users",
      providesTags: ["User"],
    }),
    createUser: builder.mutation<User, Partial<User>>({
      query: (body) => ({ url: "/users", method: "POST", body }),
      invalidatesTags: ["User"],
    }),
  }),
});

export const { useGetUsersQuery, useCreateUserMutation } = api;
```

### Memoized Selectors
```typescript
import { createSelector } from "@reduxjs/toolkit";

const selectItems = (state: RootState) => state.cart.items;
export const selectTotal = createSelector(
  [selectItems],
  (items) => items.reduce((sum, i) => sum + i.price * i.qty, 0)
);
```

---

## When to Use What

| Need | Solution |
|------|----------|
| Simple shared state (theme, auth) | **Zustand** |
| Complex domain logic, many devs | **Redux Toolkit** |
| API data caching & sync | **RTK Query** or **TanStack Query** |
| Form state | **React Hook Form** or **useState** |
| URL state (filters, pagination) | **URL params** (router) |
| Component-local UI state | **useState** / **useReducer** |
| Fine-grained atoms | **Jotai** |
| Mutable-style updates | **Valtio** |

---

## State Location Checklist

- [ ] Is it used by only one component? -> **useState**
- [ ] Is it used by a parent and its children? -> **Props** or **Context**
- [ ] Is it URL-representable (filters, page)? -> **URL params**
- [ ] Is it from an API? -> **TanStack Query / RTK Query**
- [ ] Is it form data? -> **React Hook Form**
- [ ] Is it shared across distant components? -> **Zustand / Redux**
- [ ] Does it need to survive page refresh? -> **persist middleware**

---

## Common Patterns

### Zustand: Persist + DevTools
```typescript
create(devtools(persist(immer((set) => ({...})), { name: "key" })))
```

### Redux: Tag-based cache invalidation
```
Query provides: [{ type: "User", id: "1" }, { type: "User", id: "LIST" }]
Mutation invalidates: [{ type: "User", id: "LIST" }]  // Refetches all user queries
```

### Both: Async loading pattern
```typescript
interface AsyncState<T> {
  data: T | null;
  status: "idle" | "loading" | "succeeded" | "failed";
  error: string | null;
}
```
