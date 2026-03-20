---
name: remix-data-patterns
description: >
  Remix data loading, mutations, and advanced data patterns.
  Use when implementing loaders, actions, optimistic UI, streaming,
  error boundaries, or resource routes in Remix.
  Triggers: "remix loader", "remix action", "remix optimistic UI",
  "remix streaming", "defer", "remix error boundary", "remix form",
  "useFetcher", "useLoaderData", "remix resource route".
  NOT for: React Router SPA patterns (see react-router-patterns), general React state management.
version: 1.0.0
allowed-tools: Read, Grep, Glob, Bash
---

# Remix Data Patterns

## Loader with Type Safety

```typescript
// app/routes/dashboard.tsx
import type { LoaderFunctionArgs, MetaFunction } from '@remix-run/node';
import { json, defer } from '@remix-run/node';
import { useLoaderData, Await } from '@remix-run/react';
import { Suspense } from 'react';
import { requireUser } from '~/services/auth.server';
import { getRecentActivity, getStats, getNotifications } from '~/services/dashboard.server';

// Loader: runs on the server before rendering
export async function loader({ request }: LoaderFunctionArgs) {
  const user = await requireUser(request); // Redirects if not authenticated

  // Defer non-critical data for streaming
  const stats = getStats(user.id);         // Promise — streamed later
  const notifications = getNotifications(user.id); // Promise — streamed later
  const recentActivity = await getRecentActivity(user.id, { limit: 10 }); // Awaited — blocks render

  return defer({
    user,
    recentActivity,            // Available immediately
    stats,                     // Streamed to client
    notifications,             // Streamed to client
  });
}

export const meta: MetaFunction<typeof loader> = ({ data }) => [
  { title: `Dashboard — ${data?.user.name}` },
];

export default function Dashboard() {
  const { user, recentActivity, stats, notifications } = useLoaderData<typeof loader>();

  return (
    <div>
      <h1>Welcome, {user.name}</h1>

      {/* Immediately available data */}
      <section>
        <h2>Recent Activity</h2>
        <ul>
          {recentActivity.map(item => (
            <li key={item.id}>{item.description} — {item.timestamp}</li>
          ))}
        </ul>
      </section>

      {/* Streamed data with Suspense */}
      <Suspense fallback={<div>Loading stats...</div>}>
        <Await resolve={stats} errorElement={<p>Failed to load stats</p>}>
          {(resolvedStats) => (
            <section>
              <h2>Stats</h2>
              <p>Users: {resolvedStats.userCount}</p>
              <p>Revenue: ${resolvedStats.revenue}</p>
            </section>
          )}
        </Await>
      </Suspense>

      <Suspense fallback={<div>Loading notifications...</div>}>
        <Await resolve={notifications}>
          {(resolvedNotifications) => (
            <NotificationList items={resolvedNotifications} />
          )}
        </Await>
      </Suspense>
    </div>
  );
}
```

## Actions & Form Mutations

```typescript
// app/routes/todos.tsx
import type { ActionFunctionArgs } from '@remix-run/node';
import { json, redirect } from '@remix-run/node';
import { Form, useActionData, useNavigation } from '@remix-run/react';
import { z } from 'zod';

const TodoSchema = z.object({
  title: z.string().min(1, 'Title is required').max(200),
  priority: z.enum(['low', 'medium', 'high']).default('medium'),
});

// Action: handles form submissions (POST, PUT, DELETE)
export async function action({ request }: ActionFunctionArgs) {
  const user = await requireUser(request);
  const formData = await request.formData();
  const intent = formData.get('intent');

  switch (intent) {
    case 'create': {
      const result = TodoSchema.safeParse({
        title: formData.get('title'),
        priority: formData.get('priority'),
      });

      if (!result.success) {
        return json(
          { errors: result.error.flatten().fieldErrors, intent: 'create' },
          { status: 400 }
        );
      }

      await db.todo.create({
        data: { ...result.data, userId: user.id },
      });

      return json({ success: true, intent: 'create' });
    }

    case 'toggle': {
      const id = formData.get('id') as string;
      const todo = await db.todo.findUnique({ where: { id } });
      if (!todo || todo.userId !== user.id) {
        return json({ error: 'Not found' }, { status: 404 });
      }

      await db.todo.update({
        where: { id },
        data: { completed: !todo.completed },
      });

      return json({ success: true, intent: 'toggle' });
    }

    case 'delete': {
      const id = formData.get('id') as string;
      await db.todo.deleteMany({
        where: { id, userId: user.id },
      });
      return json({ success: true, intent: 'delete' });
    }

    default:
      return json({ error: 'Unknown intent' }, { status: 400 });
  }
}

export default function Todos() {
  const { todos } = useLoaderData<typeof loader>();
  const actionData = useActionData<typeof action>();
  const navigation = useNavigation();
  const isCreating = navigation.formData?.get('intent') === 'create';

  return (
    <div>
      {/* Create form */}
      <Form method="post">
        <input type="hidden" name="intent" value="create" />
        <input
          name="title"
          placeholder="What needs doing?"
          aria-invalid={actionData?.errors?.title ? true : undefined}
          aria-errormessage="title-error"
        />
        {actionData?.errors?.title && (
          <p id="title-error" role="alert">{actionData.errors.title[0]}</p>
        )}
        <select name="priority">
          <option value="low">Low</option>
          <option value="medium">Medium</option>
          <option value="high">High</option>
        </select>
        <button type="submit" disabled={isCreating}>
          {isCreating ? 'Adding...' : 'Add Todo'}
        </button>
      </Form>

      {/* Todo list with inline actions */}
      <ul>
        {todos.map(todo => (
          <li key={todo.id}>
            <Form method="post" style={{ display: 'inline' }}>
              <input type="hidden" name="intent" value="toggle" />
              <input type="hidden" name="id" value={todo.id} />
              <button type="submit">
                {todo.completed ? '✓' : '○'} {todo.title}
              </button>
            </Form>
            <Form method="post" style={{ display: 'inline' }}>
              <input type="hidden" name="intent" value="delete" />
              <input type="hidden" name="id" value={todo.id} />
              <button type="submit" aria-label={`Delete ${todo.title}`}>✕</button>
            </Form>
          </li>
        ))}
      </ul>
    </div>
  );
}
```

## Optimistic UI with useFetcher

```typescript
// app/routes/posts.$postId.tsx
import { useFetcher } from '@remix-run/react';

function LikeButton({ postId, isLiked, likeCount }: {
  postId: string;
  isLiked: boolean;
  likeCount: number;
}) {
  const fetcher = useFetcher();

  // Optimistic state: show the expected result immediately
  const optimisticIsLiked = fetcher.formData
    ? fetcher.formData.get('intent') === 'like'
    : isLiked;
  const optimisticCount = fetcher.formData
    ? likeCount + (fetcher.formData.get('intent') === 'like' ? 1 : -1)
    : likeCount;

  return (
    <fetcher.Form method="post" action={`/api/posts/${postId}/like`}>
      <input type="hidden" name="intent" value={optimisticIsLiked ? 'unlike' : 'like'} />
      <button type="submit" aria-label={optimisticIsLiked ? 'Unlike' : 'Like'}>
        {optimisticIsLiked ? '❤️' : '🤍'} {optimisticCount}
      </button>
    </fetcher.Form>
  );
}

// Multiple fetchers for independent mutations
function CommentSection({ postId }: { postId: string }) {
  const addComment = useFetcher();
  const deleteComment = useFetcher();

  return (
    <div>
      <addComment.Form method="post" action={`/api/posts/${postId}/comments`}>
        <textarea name="body" required />
        <button type="submit" disabled={addComment.state !== 'idle'}>
          {addComment.state === 'submitting' ? 'Posting...' : 'Comment'}
        </button>
      </addComment.Form>

      {/* Each comment's delete is independent */}
      {comments.map(comment => (
        <div key={comment.id}>
          <p>{comment.body}</p>
          <deleteComment.Form method="post" action={`/api/comments/${comment.id}/delete`}>
            <button type="submit">Delete</button>
          </deleteComment.Form>
        </div>
      ))}
    </div>
  );
}
```

## Error Boundaries

```typescript
// app/routes/dashboard.tsx
import { isRouteErrorResponse, useRouteError } from '@remix-run/react';

// Route-level error boundary — catches loader and render errors
export function ErrorBoundary() {
  const error = useRouteError();

  if (isRouteErrorResponse(error)) {
    // Thrown responses from loader/action (json({}, { status: 404 }))
    return (
      <div className="error-container">
        <h1>{error.status} {error.statusText}</h1>
        <p>{error.data?.message ?? 'Something went wrong'}</p>
        {error.status === 404 && <p>The page you're looking for doesn't exist.</p>}
        {error.status === 401 && <a href="/login">Log in</a>}
      </div>
    );
  }

  // Unexpected errors
  const message = error instanceof Error ? error.message : 'Unknown error';
  console.error('Route error:', error);

  return (
    <div className="error-container">
      <h1>Oops!</h1>
      <p>Something went wrong. Our team has been notified.</p>
      {process.env.NODE_ENV === 'development' && (
        <pre>{message}</pre>
      )}
    </div>
  );
}

// Root error boundary (app/root.tsx) — last resort
// Catches errors that bubble past route-level boundaries
// Must render its own <html>, <head>, <body> since the app shell may have failed
```

## Resource Routes (API Endpoints)

```typescript
// app/routes/api.export.csv.ts — Non-UI route that returns data
import type { LoaderFunctionArgs } from '@remix-run/node';

export async function loader({ request }: LoaderFunctionArgs) {
  const user = await requireUser(request);
  const url = new URL(request.url);
  const startDate = url.searchParams.get('start');
  const endDate = url.searchParams.get('end');

  const data = await getExportData(user.id, { startDate, endDate });

  const csv = [
    'Date,Description,Amount,Category',
    ...data.map(row =>
      `${row.date},"${row.description}",${row.amount},${row.category}`
    ),
  ].join('\n');

  return new Response(csv, {
    headers: {
      'Content-Type': 'text/csv',
      'Content-Disposition': `attachment; filename="export-${startDate}-${endDate}.csv"`,
      'Cache-Control': 'no-store',
    },
  });
}

// app/routes/api.upload.ts — File upload endpoint
export async function action({ request }: ActionFunctionArgs) {
  const user = await requireUser(request);
  const formData = await request.formData();
  const file = formData.get('file') as File;

  if (!file || file.size === 0) {
    return json({ error: 'No file provided' }, { status: 400 });
  }

  if (file.size > 10 * 1024 * 1024) {
    return json({ error: 'File too large (max 10MB)' }, { status: 400 });
  }

  const buffer = Buffer.from(await file.arrayBuffer());
  const url = await uploadToStorage(buffer, file.name, file.type);

  return json({ url, filename: file.name, size: file.size });
}
```

## Gotchas

1. **Loader runs on every navigation** -- Unlike `useEffect`, loaders run on the server for every page navigation, not just on mount. This means database queries run every time a user navigates to the route. Use HTTP caching headers (`Cache-Control`) or Remix's `shouldRevalidate` to avoid redundant data fetching.

2. **Form without method="post" does a GET** -- An HTML `<Form>` without `method="post"` submits as GET, which triggers the loader, not the action. If your form mutation silently "does nothing," check the method attribute. GET forms are useful for search/filter — they update the URL query string.

3. **useFetcher doesn't update URL** -- `useFetcher` submits forms without navigation. The URL stays the same, and the page doesn't scroll to top. Use `<Form>` for mutations that should trigger navigation (create → redirect to new item). Use `useFetcher` for inline mutations (like, toggle, delete) that shouldn't change the page.

4. **defer requires Suspense boundaries** -- `defer()` returns promises that resolve asynchronously. Without `<Suspense>` + `<Await>` wrapping deferred values, accessing the data throws. The error message is not obvious. If streamed data doesn't render, check for missing Suspense boundaries.

5. **Action errors don't throw to ErrorBoundary by default** -- Returning `json({ error }, { status: 400 })` from an action does NOT trigger ErrorBoundary. It returns to the component as `actionData`. Only `throw` statements or thrown Responses trigger ErrorBoundary. Use `throw json({ message: 'Not found' }, 404)` for error boundary handling.

6. **Resource routes need explicit exports** -- A route file with only `loader` (no default export) is a resource route. But if you accidentally export a default component too, Remix treats it as a UI route and wraps it in the layout. Resource routes (API endpoints, CSV exports, webhooks) must NOT export a default component.
