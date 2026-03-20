# Remix & React Router Cheatsheet

## Remix Loader (GET data)

```typescript
export async function loader({ request, params }: LoaderFunctionArgs) {
  const url = new URL(request.url);
  const page = url.searchParams.get("page") ?? "1";
  const userId = params.id; // from route path ":id"

  const data = await db.query(...);
  if (!data) throw new Response("Not Found", { status: 404 });
  return json({ data });
}

// In component:
const { data } = useLoaderData<typeof loader>();
```

## Remix Action (POST/PUT/DELETE)

```typescript
export async function action({ request }: ActionFunctionArgs) {
  const form = await request.formData();
  const intent = form.get("intent");

  if (intent === "create") {
    const name = form.get("name") as string;
    // validate with Zod
    const result = schema.safeParse({ name });
    if (!result.success) return json({ errors: result.error.flatten() }, 400);
    await db.create(result.data);
    return redirect("/items");
  }

  if (intent === "delete") {
    await db.delete(form.get("id") as string);
    return json({ ok: true });
  }
}
```

## Forms

```tsx
// Standard form (full page navigation)
<Form method="post">
  <input name="title" />
  <button name="intent" value="create">Create</button>
</Form>

// Fetcher form (no navigation, inline mutation)
const fetcher = useFetcher();
<fetcher.Form method="post" action="/api/favorite">
  <button name="intent" value="toggle">
    {fetcher.state === "submitting" ? "Saving..." : "Favorite"}
  </button>
</fetcher.Form>

// Action data (validation errors)
const actionData = useActionData<typeof action>();
{actionData?.errors && <p>{actionData.errors.fieldErrors.name}</p>}
```

## Navigation State

```typescript
const navigation = useNavigation();
// navigation.state: "idle" | "loading" | "submitting"
// navigation.formData — submitted form data (optimistic UI)

const isSubmitting = navigation.state === "submitting";
const isLoading = navigation.state === "loading";
```

## useFetcher (Non-Navigation Mutations)

```typescript
const fetcher = useFetcher();
fetcher.submit({ intent: "toggle", id: "123" }, { method: "post", action: "/api/favorite" });
// fetcher.state: "idle" | "submitting" | "loading"
// fetcher.data — response data
```

## Streaming / Defer

```typescript
export async function loader() {
  const fastData = await getFastData();          // Await (blocking)
  const slowPromise = getSlowData();             // Don't await (streaming)
  return defer({ fast: fastData, slow: slowPromise });
}

// Component:
const { fast, slow } = useLoaderData<typeof loader>();
<div>{fast.title}</div>
<Suspense fallback={<Spinner />}>
  <Await resolve={slow}>{(data) => <SlowContent data={data} />}</Await>
</Suspense>
```

## Error Boundaries

```typescript
export function ErrorBoundary() {
  const error = useRouteError();
  if (isRouteErrorResponse(error)) {
    return <div>{error.status}: {error.statusText}</div>; // Thrown Response
  }
  return <div>Unexpected error</div>; // Thrown Error
}
```

## Authentication (Cookie Sessions)

```typescript
// sessions.server.ts
import { createCookieSessionStorage, redirect } from "@remix-run/node";

const { getSession, commitSession, destroySession } = createCookieSessionStorage({
  cookie: { name: "__session", httpOnly: true, secure: true, sameSite: "lax", maxAge: 86400, secrets: [process.env.SESSION_SECRET!] },
});

async function requireAuth(request: Request) {
  const session = await getSession(request.headers.get("Cookie"));
  const userId = session.get("userId");
  if (!userId) throw redirect("/login");
  return userId;
}
```

## Resource Routes (API endpoints)

```typescript
// routes/api.users.tsx — no default export = resource route
export async function loader({ request }: LoaderFunctionArgs) {
  const users = await db.getUsers();
  return json(users); // JSON API endpoint
}

// routes/sitemap[.]xml.tsx
export async function loader() {
  const xml = generateSitemap();
  return new Response(xml, { headers: { "Content-Type": "application/xml" } });
}
```

## React Router v7 (SPA Mode)

```typescript
import { createBrowserRouter, RouterProvider } from "react-router";

const router = createBrowserRouter([
  {
    path: "/",
    element: <Layout />,
    errorElement: <Error />,
    children: [
      { index: true, element: <Home /> },
      { path: "users", lazy: () => import("./routes/users") }, // Code split
      { path: "users/:id", lazy: () => import("./routes/user-detail") },
    ],
  },
]);

// Lazy module must export named Component (NOT default):
export function Component() { ... }
export function loader() { ... }
export function ErrorBoundary() { ... }
```

## Protected Routes (React Router SPA)

```typescript
export async function loader() {
  const token = localStorage.getItem("token");
  if (!token) throw redirect("/login"); // Must THROW, not return
  const res = await fetch("/api/me", { headers: { Authorization: `Bearer ${token}` } });
  if (!res.ok) throw redirect("/login");
  return res.json();
}
```

## Search Params

```typescript
const [searchParams, setSearchParams] = useSearchParams();
const page = parseInt(searchParams.get("page") ?? "1"); // Always strings!

// Update programmatically:
setSearchParams((prev) => { prev.set("sort", "date"); return prev; });

// Or use <Form> for declarative updates:
<Form><select name="sort">...</select><button type="submit">Apply</button></Form>
```

## Meta & Links

```typescript
export function meta({ data }: MetaArgs) {
  return [
    { title: data.title },
    { name: "description", content: data.description },
    { property: "og:image", content: data.ogImage },
  ];
}

export function links() {
  return [
    { rel: "stylesheet", href: styles },
    { rel: "canonical", href: "https://example.com/page" },
  ];
}
```

## Quick Gotchas

```
1. useLoaderData needs createBrowserRouter, NOT <BrowserRouter>
2. lazy() modules export { Component }, not default
3. searchParams.get() returns string | null, never number
4. redirect() must be THROWN in auth guards (not returned)
5. After action completes, ALL active loaders re-run (use shouldRevalidate to optimize)
6. .server.ts files are NEVER sent to the browser (server-only code)
7. Resource routes = no default export component
8. Cookie session secrets must be in env vars, never hardcoded
9. useFetcher for inline mutations, <Form> for page-level navigation
10. defer + Await for non-blocking slow data (fast data is awaited, slow is streamed)
```

## File Convention (Remix v2 Flat Routes)

```
routes/
  _index.tsx           → /
  about.tsx            → /about
  dashboard.tsx        → /dashboard (layout)
  dashboard._index.tsx → /dashboard (index)
  dashboard.users.tsx  → /dashboard/users
  dashboard.$id.tsx    → /dashboard/:id
  api.users.tsx        → /api/users (resource route, no UI)
  $.tsx                → /* (catch-all / 404)
```

## Remix vs React Router v7 Decision

```
Full-stack (SSR, server loaders, cookie sessions, streaming)?
  → Remix

SPA with client-side data fetching?
  → React Router v7 with createBrowserRouter

Migrating from React Router v5/v6?
  → React Router v7 first, then add Remix later if needed
```
