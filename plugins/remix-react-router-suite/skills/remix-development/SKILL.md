---
name: remix-development
description: >
  Production Remix and React Router v7 patterns — loaders, actions, forms,
  nested routing, streaming, defer, error boundaries, authentication,
  optimistic UI, and deployment.
  Triggers: "remix", "react router", "loader", "action", "remix form",
  "remix auth", "remix streaming", "remix deploy", "react router v7".
  NOT for: Next.js (use nextjs-fullstack), plain React SPA (use react-patterns).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Remix Development

## Project Structure

```
app/
  entry.client.tsx      # Client entry (hydration)
  entry.server.tsx      # Server entry (request handler)
  root.tsx              # Root layout (html, head, body)
  routes/
    _index.tsx           # / (home)
    about.tsx            # /about
    dashboard.tsx        # /dashboard layout
    dashboard._index.tsx # /dashboard
    dashboard.users.tsx  # /dashboard/users
    dashboard.users.$id.tsx # /dashboard/users/:id
    api.health.tsx       # /api/health (resource route)
  components/           # Shared UI components
  lib/                  # Utilities, db, auth
    db.server.ts         # Server-only database client
    auth.server.ts       # Server-only auth helpers
    session.server.ts    # Session management
  styles/               # CSS files
```

## Loaders — Server-Side Data Fetching

```typescript
// app/routes/dashboard.users.tsx
import type { LoaderFunctionArgs } from "react-router";
import { useLoaderData } from "react-router";
import { requireAuth } from "~/lib/auth.server";
import { db } from "~/lib/db.server";

export async function loader({ request }: LoaderFunctionArgs) {
  const user = await requireAuth(request); // Throws redirect if not authed

  const url = new URL(request.url);
  const search = url.searchParams.get("q") ?? "";
  const page = parseInt(url.searchParams.get("page") ?? "1");
  const limit = 20;

  const [users, total] = await Promise.all([
    db.user.findMany({
      where: search ? { name: { contains: search, mode: "insensitive" } } : {},
      skip: (page - 1) * limit,
      take: limit,
      orderBy: { createdAt: "desc" },
    }),
    db.user.count({
      where: search ? { name: { contains: search, mode: "insensitive" } } : {},
    }),
  ]);

  return { users, total, page, search };
}

export default function UsersPage() {
  const { users, total, page, search } = useLoaderData<typeof loader>();

  return (
    <div>
      <Form method="get">
        <input type="search" name="q" defaultValue={search} placeholder="Search..." />
        <button type="submit">Search</button>
      </Form>
      <ul>
        {users.map((user) => (
          <li key={user.id}>
            <Link to={user.id}>{user.name}</Link>
          </li>
        ))}
      </ul>
      <Pagination total={total} page={page} limit={20} />
    </div>
  );
}
```

## Actions — Server-Side Mutations

```typescript
// app/routes/dashboard.users.tsx (same file as loader)
import type { ActionFunctionArgs } from "react-router";
import { redirect, data } from "react-router";
import { z } from "zod";

const CreateUserSchema = z.object({
  name: z.string().min(2, "Name must be at least 2 characters"),
  email: z.string().email("Invalid email"),
  role: z.enum(["admin", "member", "viewer"]),
});

export async function action({ request }: ActionFunctionArgs) {
  const user = await requireAuth(request);

  const formData = await request.formData();
  const intent = formData.get("intent");

  switch (intent) {
    case "create": {
      const result = CreateUserSchema.safeParse(Object.fromEntries(formData));

      if (!result.success) {
        return data(
          { errors: result.error.flatten().fieldErrors },
          { status: 400 }
        );
      }

      await db.user.create({ data: result.data });
      return redirect("/dashboard/users");
    }

    case "delete": {
      const userId = formData.get("userId");
      if (typeof userId !== "string") {
        return data({ error: "Invalid user ID" }, { status: 400 });
      }
      await db.user.delete({ where: { id: userId } });
      return { ok: true };
    }

    default:
      return data({ error: "Invalid intent" }, { status: 400 });
  }
}
```

## Forms — Progressive Enhancement

```typescript
import { Form, useActionData, useNavigation } from "react-router";

export default function CreateUserForm() {
  const actionData = useActionData<typeof action>();
  const navigation = useNavigation();
  const isSubmitting = navigation.state === "submitting";

  return (
    <Form method="post">
      <input type="hidden" name="intent" value="create" />

      <div>
        <label htmlFor="name">Name</label>
        <input id="name" name="name" required />
        {actionData?.errors?.name && (
          <p className="error">{actionData.errors.name[0]}</p>
        )}
      </div>

      <div>
        <label htmlFor="email">Email</label>
        <input id="email" name="email" type="email" required />
        {actionData?.errors?.email && (
          <p className="error">{actionData.errors.email[0]}</p>
        )}
      </div>

      <div>
        <label htmlFor="role">Role</label>
        <select id="role" name="role" defaultValue="member">
          <option value="admin">Admin</option>
          <option value="member">Member</option>
          <option value="viewer">Viewer</option>
        </select>
      </div>

      <button type="submit" disabled={isSubmitting}>
        {isSubmitting ? "Creating..." : "Create User"}
      </button>
    </Form>
  );
}
```

## useFetcher — Non-Navigation Mutations

```typescript
import { useFetcher } from "react-router";

function DeleteButton({ userId }: { userId: string }) {
  const fetcher = useFetcher();
  const isDeleting = fetcher.state !== "idle";

  return (
    <fetcher.Form method="post">
      <input type="hidden" name="intent" value="delete" />
      <input type="hidden" name="userId" value={userId} />
      <button
        type="submit"
        disabled={isDeleting}
        onClick={(e) => {
          if (!confirm("Are you sure?")) e.preventDefault();
        }}
      >
        {isDeleting ? "Deleting..." : "Delete"}
      </button>
    </fetcher.Form>
  );
}

// Fetcher for inline updates (no navigation)
function ToggleFavorite({ itemId, isFavorite }: { itemId: string; isFavorite: boolean }) {
  const fetcher = useFetcher();

  // Optimistic UI: show the new state immediately
  const optimisticFavorite = fetcher.formData
    ? fetcher.formData.get("favorite") === "true"
    : isFavorite;

  return (
    <fetcher.Form method="post" action="/api/favorites">
      <input type="hidden" name="itemId" value={itemId} />
      <input type="hidden" name="favorite" value={String(!optimisticFavorite)} />
      <button type="submit">
        {optimisticFavorite ? "★" : "☆"}
      </button>
    </fetcher.Form>
  );
}
```

## Streaming with defer + Await

```typescript
import { defer } from "react-router";
import { Suspense } from "react";
import { useLoaderData, Await } from "react-router";

export async function loader({ params }: LoaderFunctionArgs) {
  const userId = params.id!;

  // Fast data — awaited, included in initial HTML
  const user = await db.user.findUnique({ where: { id: userId } });
  if (!user) throw new Response("Not Found", { status: 404 });

  // Slow data — deferred, streamed in later
  const activityPromise = db.activity.findMany({
    where: { userId },
    orderBy: { createdAt: "desc" },
    take: 50,
  });

  const statsPromise = calculateUserStats(userId); // expensive

  return defer({
    user,                     // Resolved — in initial HTML
    activity: activityPromise, // Promise — streamed later
    stats: statsPromise,       // Promise — streamed later
  });
}

export default function UserProfile() {
  const { user, activity, stats } = useLoaderData<typeof loader>();

  return (
    <div>
      {/* Renders immediately */}
      <h1>{user.name}</h1>
      <p>{user.email}</p>

      {/* Streams in when ready */}
      <Suspense fallback={<div>Loading stats...</div>}>
        <Await resolve={stats}>
          {(resolvedStats) => (
            <StatsCard stats={resolvedStats} />
          )}
        </Await>
      </Suspense>

      <Suspense fallback={<ActivitySkeleton />}>
        <Await resolve={activity} errorElement={<p>Failed to load activity</p>}>
          {(resolvedActivity) => (
            <ActivityFeed items={resolvedActivity} />
          )}
        </Await>
      </Suspense>
    </div>
  );
}
```

## Error Boundaries

```typescript
// app/routes/dashboard.users.$id.tsx
import { isRouteErrorResponse, useRouteError } from "react-router";

export function ErrorBoundary() {
  const error = useRouteError();

  if (isRouteErrorResponse(error)) {
    return (
      <div className="error-container">
        <h1>{error.status} {error.statusText}</h1>
        <p>{error.data}</p>
      </div>
    );
  }

  // Unexpected errors
  return (
    <div className="error-container">
      <h1>Something went wrong</h1>
      <p>{error instanceof Error ? error.message : "Unknown error"}</p>
    </div>
  );
}

// Throwing responses in loaders/actions
export async function loader({ params }: LoaderFunctionArgs) {
  const user = await db.user.findUnique({ where: { id: params.id } });
  if (!user) {
    throw new Response("User not found", { status: 404 });
  }
  return { user };
}
```

## Authentication

```typescript
// app/lib/session.server.ts
import { createCookieSessionStorage, redirect } from "react-router";

const sessionStorage = createCookieSessionStorage({
  cookie: {
    name: "__session",
    httpOnly: true,
    maxAge: 60 * 60 * 24 * 30, // 30 days
    path: "/",
    sameSite: "lax",
    secrets: [process.env.SESSION_SECRET!],
    secure: process.env.NODE_ENV === "production",
  },
});

export async function createUserSession(userId: string, redirectTo: string) {
  const session = await sessionStorage.getSession();
  session.set("userId", userId);
  return redirect(redirectTo, {
    headers: {
      "Set-Cookie": await sessionStorage.commitSession(session),
    },
  });
}

export async function getUserId(request: Request): Promise<string | null> {
  const session = await sessionStorage.getSession(request.headers.get("Cookie"));
  return session.get("userId") ?? null;
}

export async function requireAuth(request: Request) {
  const userId = await getUserId(request);
  if (!userId) throw redirect("/login");
  const user = await db.user.findUnique({ where: { id: userId } });
  if (!user) throw redirect("/login");
  return user;
}

export async function logout(request: Request) {
  const session = await sessionStorage.getSession(request.headers.get("Cookie"));
  return redirect("/login", {
    headers: {
      "Set-Cookie": await sessionStorage.destroySession(session),
    },
  });
}
```

```typescript
// app/routes/_auth.login.tsx
export async function action({ request }: ActionFunctionArgs) {
  const formData = await request.formData();
  const email = formData.get("email") as string;
  const password = formData.get("password") as string;

  const user = await verifyLogin(email, password);
  if (!user) {
    return data({ error: "Invalid credentials" }, { status: 400 });
  }

  return createUserSession(user.id, "/dashboard");
}
```

## Resource Routes (API Endpoints)

```typescript
// app/routes/api.health.tsx — no default export = no UI
export async function loader() {
  return Response.json({ status: "ok", timestamp: Date.now() });
}

// app/routes/[sitemap.xml].tsx
export async function loader() {
  const pages = await db.page.findMany({ select: { slug: true, updatedAt: true } });

  const sitemap = `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  ${pages.map((p) => `<url>
    <loc>https://example.com/${p.slug}</loc>
    <lastmod>${p.updatedAt.toISOString()}</lastmod>
  </url>`).join("\n")}
</urlset>`;

  return new Response(sitemap, {
    headers: {
      "Content-Type": "application/xml",
      "Cache-Control": "public, max-age=3600",
    },
  });
}
```

## Nested Layouts

```typescript
// app/routes/dashboard.tsx — layout route
import { Outlet, NavLink, useLoaderData } from "react-router";

export async function loader({ request }: LoaderFunctionArgs) {
  const user = await requireAuth(request);
  return { user };
}

export default function DashboardLayout() {
  const { user } = useLoaderData<typeof loader>();

  return (
    <div className="dashboard">
      <aside>
        <nav>
          <NavLink to="/dashboard" end className={({ isActive }) => isActive ? "active" : ""}>
            Home
          </NavLink>
          <NavLink to="/dashboard/users">Users</NavLink>
          <NavLink to="/dashboard/settings">Settings</NavLink>
        </nav>
        <p>{user.email}</p>
      </aside>
      <main>
        <Outlet /> {/* Child routes render here */}
      </main>
    </div>
  );
}
```

## Meta & Links

```typescript
import type { MetaFunction } from "react-router";

export const meta: MetaFunction<typeof loader> = ({ data }) => {
  if (!data) {
    return [{ title: "Not Found" }];
  }
  return [
    { title: `${data.user.name} | Dashboard` },
    { name: "description", content: `Profile of ${data.user.name}` },
    { property: "og:title", content: data.user.name },
  ];
};

export function links() {
  return [
    { rel: "stylesheet", href: "/styles/dashboard.css" },
    { rel: "icon", href: "/favicon.svg", type: "image/svg+xml" },
  ];
}
```

## Gotchas

1. **Loaders run in parallel for nested routes** — Parent and child loaders fire simultaneously, not sequentially. Don't depend on parent loader data in child loaders. If you need shared data, use `context` or make the child loader fetch its own data.

2. **`return` vs `throw` for responses** — `return redirect("/login")` continues execution after the return in the calling function. `throw redirect("/login")` stops execution immediately. In `requireAuth`, always `throw` to prevent the rest of the loader from running with no user.

3. **`.server.ts` is the boundary** — Any file with `.server.ts` or in a `.server/` directory is excluded from the client bundle. Without this suffix, importing `db` in a route file will try to bundle your database client for the browser.

4. **Forms revalidate all loaders** — After an action completes, Remix revalidates ALL active loaders on the page (parent + child). This is usually what you want, but for expensive loaders, use `shouldRevalidate` to opt out.

5. **`useActionData` resets on navigation** — Action data is cleared when the user navigates away. If you need persistent error messages across navigations, use session flash messages instead.

6. **Cookie session size limit** — Cookie-based sessions have a ~4KB limit. Store only the user ID in the session cookie, not the full user object. Fetch user data in loaders.
