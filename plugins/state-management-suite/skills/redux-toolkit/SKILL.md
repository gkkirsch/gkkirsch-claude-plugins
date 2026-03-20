---
name: redux-toolkit
description: >
  Redux Toolkit (RTK) — createSlice, configureStore, createAsyncThunk,
  RTK Query, entity adapter, and listener middleware.
  Triggers: "redux", "redux toolkit", "rtk", "createSlice", "rtk query",
  "redux store", "redux middleware".
  NOT for: simple state (use zustand), server-only state (use react-query).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Redux Toolkit

## Quick Start

```bash
npm install @reduxjs/toolkit react-redux
```

## Store Setup

```typescript
// store/index.ts
import { configureStore } from '@reduxjs/toolkit';
import { authSlice } from './slices/authSlice';
import { postsSlice } from './slices/postsSlice';
import { uiSlice } from './slices/uiSlice';
import { apiSlice } from './api/apiSlice';

export const store = configureStore({
  reducer: {
    auth: authSlice.reducer,
    posts: postsSlice.reducer,
    ui: uiSlice.reducer,
    [apiSlice.reducerPath]: apiSlice.reducer,
  },
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware().concat(apiSlice.middleware),
});

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;
```

```typescript
// store/hooks.ts — typed hooks (use these, not raw useSelector/useDispatch)
import { useDispatch, useSelector } from 'react-redux';
import type { RootState, AppDispatch } from './index';

export const useAppDispatch = useDispatch.withTypes<AppDispatch>();
export const useAppSelector = useSelector.withTypes<RootState>();
```

```tsx
// app/providers.tsx
import { Provider } from 'react-redux';
import { store } from '@/store';

export function Providers({ children }: { children: React.ReactNode }) {
  return <Provider store={store}>{children}</Provider>;
}
```

## createSlice

```typescript
// store/slices/postsSlice.ts
import { createSlice, createAsyncThunk, PayloadAction } from '@reduxjs/toolkit';

interface Post {
  id: string;
  title: string;
  body: string;
  authorId: string;
  status: 'draft' | 'published';
  createdAt: string;
}

interface PostsState {
  items: Post[];
  selectedId: string | null;
  status: 'idle' | 'loading' | 'succeeded' | 'failed';
  error: string | null;
}

const initialState: PostsState = {
  items: [],
  selectedId: null,
  status: 'idle',
  error: null,
};

// Async thunk
export const fetchPosts = createAsyncThunk(
  'posts/fetchAll',
  async (_, { rejectWithValue }) => {
    try {
      const response = await fetch('/api/posts');
      if (!response.ok) throw new Error('Failed to fetch');
      return (await response.json()) as Post[];
    } catch (error) {
      return rejectWithValue((error as Error).message);
    }
  }
);

export const createPost = createAsyncThunk(
  'posts/create',
  async (data: { title: string; body: string }, { getState, rejectWithValue }) => {
    const state = getState() as RootState;
    const token = state.auth.token;

    const response = await fetch('/api/posts', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify(data),
    });

    if (!response.ok) {
      return rejectWithValue('Failed to create post');
    }

    return (await response.json()) as Post;
  }
);

export const postsSlice = createSlice({
  name: 'posts',
  initialState,
  reducers: {
    selectPost: (state, action: PayloadAction<string>) => {
      state.selectedId = action.payload;
    },
    clearSelection: (state) => {
      state.selectedId = null;
    },
    updatePostLocally: (state, action: PayloadAction<{ id: string; changes: Partial<Post> }>) => {
      const post = state.items.find((p) => p.id === action.payload.id);
      if (post) {
        Object.assign(post, action.payload.changes);
      }
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(fetchPosts.pending, (state) => {
        state.status = 'loading';
        state.error = null;
      })
      .addCase(fetchPosts.fulfilled, (state, action) => {
        state.status = 'succeeded';
        state.items = action.payload;
      })
      .addCase(fetchPosts.rejected, (state, action) => {
        state.status = 'failed';
        state.error = action.payload as string;
      })
      .addCase(createPost.fulfilled, (state, action) => {
        state.items.unshift(action.payload);
      });
  },
});

export const { selectPost, clearSelection, updatePostLocally } = postsSlice.actions;

// Selectors
export const selectAllPosts = (state: RootState) => state.posts.items;
export const selectPostById = (state: RootState, id: string) =>
  state.posts.items.find((p) => p.id === id);
export const selectPostsStatus = (state: RootState) => state.posts.status;
export const selectPublishedPosts = (state: RootState) =>
  state.posts.items.filter((p) => p.status === 'published');
```

## Entity Adapter (Normalized State)

```typescript
import { createSlice, createEntityAdapter } from '@reduxjs/toolkit';

interface User {
  id: string;
  name: string;
  email: string;
}

const usersAdapter = createEntityAdapter<User>({
  sortComparer: (a, b) => a.name.localeCompare(b.name),
});

// State shape: { ids: ['1', '2'], entities: { '1': {...}, '2': {...} } }
const usersSlice = createSlice({
  name: 'users',
  initialState: usersAdapter.getInitialState({
    loading: false,
  }),
  reducers: {
    addUser: usersAdapter.addOne,
    addUsers: usersAdapter.addMany,
    updateUser: usersAdapter.updateOne,   // { id: '1', changes: { name: 'New' } }
    upsertUser: usersAdapter.upsertOne,   // Add or update
    removeUser: usersAdapter.removeOne,
    setAllUsers: usersAdapter.setAll,
  },
});

// Auto-generated selectors
export const {
  selectAll: selectAllUsers,
  selectById: selectUserById,
  selectIds: selectUserIds,
  selectTotal: selectUserCount,
} = usersAdapter.getSelectors((state: RootState) => state.users);
```

## RTK Query (Built-in API Caching)

```typescript
// store/api/apiSlice.ts
import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react';

interface Post {
  id: string;
  title: string;
  body: string;
}

interface PaginatedResponse<T> {
  data: T[];
  meta: { total: number; page: number; limit: number };
}

export const apiSlice = createApi({
  reducerPath: 'api',
  baseQuery: fetchBaseQuery({
    baseUrl: '/api',
    prepareHeaders: (headers, { getState }) => {
      const token = (getState() as RootState).auth.token;
      if (token) headers.set('Authorization', `Bearer ${token}`);
      return headers;
    },
  }),
  tagTypes: ['Post', 'User'],
  endpoints: (builder) => ({
    // GET /api/posts
    getPosts: builder.query<PaginatedResponse<Post>, { page?: number; limit?: number }>({
      query: ({ page = 1, limit = 20 } = {}) => `/posts?page=${page}&limit=${limit}`,
      providesTags: (result) =>
        result
          ? [
              ...result.data.map(({ id }) => ({ type: 'Post' as const, id })),
              { type: 'Post', id: 'LIST' },
            ]
          : [{ type: 'Post', id: 'LIST' }],
    }),

    // GET /api/posts/:id
    getPost: builder.query<Post, string>({
      query: (id) => `/posts/${id}`,
      providesTags: (result, error, id) => [{ type: 'Post', id }],
    }),

    // POST /api/posts
    createPost: builder.mutation<Post, { title: string; body: string }>({
      query: (body) => ({ url: '/posts', method: 'POST', body }),
      invalidatesTags: [{ type: 'Post', id: 'LIST' }],
    }),

    // PATCH /api/posts/:id
    updatePost: builder.mutation<Post, { id: string; title?: string; body?: string }>({
      query: ({ id, ...patch }) => ({ url: `/posts/${id}`, method: 'PATCH', body: patch }),
      invalidatesTags: (result, error, { id }) => [{ type: 'Post', id }],
    }),

    // DELETE /api/posts/:id
    deletePost: builder.mutation<void, string>({
      query: (id) => ({ url: `/posts/${id}`, method: 'DELETE' }),
      invalidatesTags: (result, error, id) => [
        { type: 'Post', id },
        { type: 'Post', id: 'LIST' },
      ],
    }),
  }),
});

export const {
  useGetPostsQuery,
  useGetPostQuery,
  useCreatePostMutation,
  useUpdatePostMutation,
  useDeletePostMutation,
} = apiSlice;
```

### Using RTK Query in Components

```tsx
function PostsList() {
  const { data, isLoading, error } = useGetPostsQuery({ page: 1, limit: 20 });
  const [createPost, { isLoading: isCreating }] = useCreatePostMutation();

  if (isLoading) return <Skeleton />;
  if (error) return <ErrorDisplay error={error} />;

  return (
    <div>
      <button
        onClick={() => createPost({ title: 'New Post', body: 'Content' })}
        disabled={isCreating}
      >
        Create Post
      </button>
      {data?.data.map((post) => (
        <PostCard key={post.id} post={post} />
      ))}
    </div>
  );
}
```

### RTK Query Optimistic Update

```typescript
updatePost: builder.mutation<Post, { id: string; title: string }>({
  query: ({ id, ...patch }) => ({
    url: `/posts/${id}`,
    method: 'PATCH',
    body: patch,
  }),
  async onQueryStarted({ id, ...patch }, { dispatch, queryFulfilled }) {
    // Optimistic update
    const patchResult = dispatch(
      apiSlice.util.updateQueryData('getPost', id, (draft) => {
        Object.assign(draft, patch);
      })
    );
    try {
      await queryFulfilled;
    } catch {
      patchResult.undo(); // Rollback on failure
    }
  },
}),
```

## Listener Middleware (Side Effects)

```typescript
import { createListenerMiddleware, isAnyOf } from '@reduxjs/toolkit';

const listenerMiddleware = createListenerMiddleware();

// Track auth changes
listenerMiddleware.startListening({
  actionCreator: authSlice.actions.logout,
  effect: async (action, listenerApi) => {
    // Clear all cached API data on logout
    listenerApi.dispatch(apiSlice.util.resetApiState());
    // Redirect
    window.location.href = '/login';
  },
});

// Debounced search
listenerMiddleware.startListening({
  actionCreator: searchSlice.actions.setQuery,
  effect: async (action, listenerApi) => {
    // Cancel if another setQuery fires within 300ms
    listenerApi.cancelActiveListeners();
    await listenerApi.delay(300);

    // Dispatch search
    listenerApi.dispatch(
      searchApi.endpoints.search.initiate(action.payload)
    );
  },
});

// Add to store
export const store = configureStore({
  reducer: { /* ... */ },
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware()
      .prepend(listenerMiddleware.middleware)
      .concat(apiSlice.middleware),
});
```

## Gotchas

1. **Always use the typed hooks.** `useAppSelector` and `useAppDispatch` from `store/hooks.ts`, never raw `useSelector`/`useDispatch`.

2. **Immer is built-in.** Inside `createSlice` reducers, you can mutate `state` directly — RTK uses Immer under the hood. But only inside reducers.

3. **RTK Query auto-caches.** If two components call `useGetPostQuery('123')`, only ONE request fires. The second component gets the cached result.

4. **Tag invalidation is how you refetch.** When a mutation invalidates a tag, all queries providing that tag automatically refetch. Design your tag system carefully.

5. **Don't mix RTK Query with createAsyncThunk** for the same data. Pick one approach per resource. RTK Query is better for CRUD, createAsyncThunk for non-API async work.

6. **Avoid storing derived data.** Don't store `filteredPosts` in the slice. Use a selector: `createSelector` from `@reduxjs/toolkit` for memoized derivation.
