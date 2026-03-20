---
name: react-router-patterns
description: >
  React Router v7 patterns for SPA and hybrid apps — createBrowserRouter,
  route configuration, data APIs, lazy loading, protected routes,
  search params, and migration from v5/v6.
  Triggers: "react router", "createBrowserRouter", "route config",
  "protected routes", "lazy routes", "react router migration".
  NOT for: Remix full-stack (use remix-development).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# React Router v7 Patterns

## Router Setup

```typescript
// app/router.tsx
import { createBrowserRouter, RouterProvider } from "react-router";

const router = createBrowserRouter([
  {
    path: "/",
    element: <RootLayout />,
    errorElement: <RootError />,
    children: [
      { index: true, element: <Home /> },
      { path: "about", element: <About /> },
      {
        path: "dashboard",
        element: <DashboardLayout />,
        loader: dashboardLoader,
        children: [
          { index: true, element: <DashboardHome />, loader: dashboardHomeLoader },
          { path: "users", element: <Users />, loader: usersLoader },
          { path: "users/:id", element: <UserDetail />, loader: userDetailLoader },
          { path: "settings", element: <Settings />, action: settingsAction },
        ],
      },
      { path: "*", element: <NotFound /> },
    ],
  },
]);

export default function App() {
  return <RouterProvider router={router} />;
}
```

## Lazy Loading Routes

```typescript
const router = createBrowserRouter([
  {
    path: "/",
    element: <RootLayout />,
    children: [
      { index: true, lazy: () => import("./routes/home") },
      {
        path: "dashboard",
        lazy: () => import("./routes/dashboard"),
        children: [
          { index: true, lazy: () => import("./routes/dashboard/home") },
          { path: "users", lazy: () => import("./routes/dashboard/users") },
          {
            path: "users/:id",
            lazy: () => import("./routes/dashboard/user-detail"),
          },
        ],
      },
      {
        path: "admin",
        lazy: () => import("./routes/admin"),
      },
    ],
  },
]);

// routes/dashboard/users.tsx — lazy module exports
export function loader() {
  return fetch("/api/users").then((r) => r.json());
}

export function Component() {
  const users = useLoaderData();
  return <UserList users={users} />;
}

// Optional: custom error for this route
export function ErrorBoundary() {
  return <div>Failed to load users</div>;
}
```

## Protected Routes

```typescript
// lib/auth.ts
import { redirect } from "react-router";

export async function requireAuth() {
  const token = localStorage.getItem("token");
  if (!token) throw redirect("/login");

  try {
    const res = await fetch("/api/me", {
      headers: { Authorization: `Bearer ${token}` },
    });
    if (!res.ok) throw redirect("/login");
    return res.json();
  } catch {
    throw redirect("/login");
  }
}

// Route config with auth loader
{
  path: "dashboard",
  loader: async () => {
    const user = await requireAuth();
    return { user };
  },
  element: <DashboardLayout />,
  children: [/* ... */],
}
```

## Search Params

```typescript
import { useSearchParams, Form } from "react-router";

function UserFilters() {
  const [searchParams, setSearchParams] = useSearchParams();
  const role = searchParams.get("role") ?? "all";
  const sort = searchParams.get("sort") ?? "name";

  // Declarative — Form submits update URL search params
  return (
    <Form>
      <select name="role" defaultValue={role}>
        <option value="all">All Roles</option>
        <option value="admin">Admin</option>
        <option value="member">Member</option>
      </select>
      <select name="sort" defaultValue={sort}>
        <option value="name">Name</option>
        <option value="date">Date</option>
      </select>
      <button type="submit">Filter</button>
    </Form>
  );
}

// Imperative — programmatic search param updates
function SortButton({ field }: { field: string }) {
  const [searchParams, setSearchParams] = useSearchParams();

  return (
    <button
      onClick={() => {
        setSearchParams((prev) => {
          prev.set("sort", field);
          prev.set("page", "1"); // Reset page on sort change
          return prev;
        });
      }}
    >
      Sort by {field}
    </button>
  );
}
```

## Data APIs (Loader + Action in SPA)

```typescript
// routes/dashboard/users.tsx
import { useLoaderData, useFetcher, useNavigation } from "react-router";

export async function loader({ request }: { request: Request }) {
  const url = new URL(request.url);
  const page = url.searchParams.get("page") ?? "1";

  const res = await fetch(`/api/users?page=${page}`);
  if (!res.ok) throw new Response("Failed to load", { status: res.status });
  return res.json();
}

export async function action({ request }: { request: Request }) {
  const formData = await request.formData();
  const intent = formData.get("intent");

  if (intent === "delete") {
    const id = formData.get("id");
    const res = await fetch(`/api/users/${id}`, { method: "DELETE" });
    if (!res.ok) throw new Response("Failed to delete", { status: res.status });
    return { ok: true };
  }

  throw new Response("Invalid intent", { status: 400 });
}

export function Component() {
  const { users, total } = useLoaderData();
  const navigation = useNavigation();
  const isLoading = navigation.state === "loading";

  return (
    <div style={{ opacity: isLoading ? 0.5 : 1 }}>
      {users.map((user: any) => (
        <UserRow key={user.id} user={user} />
      ))}
    </div>
  );
}
```

## Navigation State

```typescript
import { useNavigation, useLocation } from "react-router";

function GlobalLoadingIndicator() {
  const navigation = useNavigation();

  // navigation.state: "idle" | "loading" | "submitting"
  if (navigation.state === "idle") return null;

  return (
    <div className="global-spinner">
      {navigation.state === "submitting" ? "Saving..." : "Loading..."}
    </div>
  );
}

// Progress bar pattern
function NProgressBar() {
  const navigation = useNavigation();

  useEffect(() => {
    if (navigation.state !== "idle") NProgress.start();
    else NProgress.done();
  }, [navigation.state]);

  return null;
}
```

## Migration from v5/v6

### v5 → v7

```typescript
// v5 (class-based, render props)
<Switch>
  <Route exact path="/" component={Home} />
  <Route path="/users/:id" render={(props) => <User id={props.match.params.id} />} />
  <Redirect from="/old" to="/new" />
</Switch>

// v7 (object-based, data APIs)
createBrowserRouter([
  { path: "/", element: <Home /> },
  { path: "/users/:id", element: <User />, loader: userLoader },
  { path: "/old", loader: () => redirect("/new") },
])
```

### v6 → v7

```typescript
// v6 (JSX routes)
<Routes>
  <Route path="/" element={<Home />} />
  <Route path="/dashboard" element={<Dashboard />}>
    <Route index element={<DashboardHome />} />
  </Route>
</Routes>

// v7 (same JSX works, but prefer object config for data APIs)
createBrowserRouter(
  createRoutesFromElements(
    <Route path="/" element={<RootLayout />}>
      <Route index element={<Home />} />
      <Route path="dashboard" element={<Dashboard />} loader={dashboardLoader}>
        <Route index element={<DashboardHome />} />
      </Route>
    </Route>
  )
)
```

## Gotchas

1. **`useLoaderData` only works with data router** — You must use `createBrowserRouter` (not `<BrowserRouter>`). The JSX-only `<BrowserRouter>` doesn't support loaders, actions, or fetchers.

2. **Loader errors propagate up** — If a child loader throws and has no `errorElement`, the error bubbles to the parent's error boundary. Add `errorElement` to routes that should contain their own errors.

3. **`lazy()` must export `Component`** — Not `default`. The lazy module should export named `Component`, `loader`, `action`, `ErrorBoundary`. Using `export default` won't work.

4. **Search params are strings** — `searchParams.get("page")` returns a string, not a number. Always parse: `parseInt(searchParams.get("page") ?? "1")`. Don't compare with `===` against numbers.

5. **`redirect()` throws, not returns** — In loaders and actions, `redirect("/path")` creates a Response. But in auth guards, you must `throw redirect(...)` to stop execution. If you `return redirect(...)`, code after the return still executes (in the calling function).

6. **Revalidation after actions** — After an action completes, ALL active loaders re-run by default. Use `shouldRevalidate` on specific routes to prevent unnecessary refetches.
