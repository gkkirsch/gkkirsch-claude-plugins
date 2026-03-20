---
name: sveltekit-development
description: >
  SvelteKit full-stack framework — routing, load functions, form actions,
  SSR/SSG/SPA modes, hooks, API routes, and deployment.
  Triggers: "sveltekit", "svelte kit", "svelte project", "svelte app",
  "svelte routing", "svelte forms", "svelte ssr", "svelte deployment".
  NOT for: content-heavy static sites (use astro-development).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# SvelteKit Development

## Quick Start

```bash
# Create new project
npx sv create my-app
# Or:
npm create svelte@latest my-app

# Options: Skeleton, Demo, Library
# Add-ons: TypeScript, ESLint, Prettier, Playwright, Vitest, Tailwind

cd my-app
npm install
npm run dev
```

## Project Structure

```
src/
├── routes/           # File-based routing
│   ├── +page.svelte          # / (home page)
│   ├── +page.ts              # Load data for home
│   ├── +page.server.ts       # Server-only load + form actions
│   ├── +layout.svelte        # Root layout
│   ├── +layout.ts            # Layout data
│   ├── +error.svelte         # Error page
│   ├── about/
│   │   └── +page.svelte      # /about
│   ├── blog/
│   │   ├── +page.svelte      # /blog
│   │   ├── +page.server.ts   # Load posts
│   │   └── [slug]/
│   │       ├── +page.svelte  # /blog/my-post
│   │       └── +page.ts      # Load single post
│   ├── api/
│   │   └── posts/
│   │       └── +server.ts    # API endpoint: /api/posts
│   └── (auth)/               # Route group (no URL segment)
│       ├── login/
│       └── register/
├── lib/              # Shared code ($lib alias)
│   ├── components/
│   ├── server/       # Server-only code ($lib/server)
│   └── utils.ts
├── params/           # Param matchers
│   └── integer.ts
├── hooks.server.ts   # Server hooks
├── hooks.client.ts   # Client hooks
└── app.d.ts          # Type declarations
static/               # Static assets (served as-is)
svelte.config.js      # SvelteKit configuration
vite.config.ts        # Vite configuration
```

## Routing

```
src/routes/
├── +page.svelte              → /
├── about/+page.svelte        → /about
├── blog/
│   ├── +page.svelte          → /blog
│   └── [slug]/+page.svelte   → /blog/hello-world (dynamic)
├── [category]/[id]/          → /tech/42 (multiple params)
├── [...path]/+page.svelte    → /any/nested/path (rest/catch-all)
├── (marketing)/              → Route group (no URL segment)
│   ├── pricing/+page.svelte  → /pricing
│   └── faq/+page.svelte      → /faq
└── [[lang]]/                 → Optional param: / or /en
    └── +page.svelte
```

### Param Matchers

```typescript
// src/params/integer.ts
import type { ParamMatcher } from "@sveltejs/kit";

export const match: ParamMatcher = (param) => {
  return /^\d+$/.test(param);
};

// Use in route: src/routes/items/[id=integer]/+page.svelte
// Matches /items/42 but not /items/abc
```

## Load Functions

```typescript
// src/routes/blog/+page.ts — Universal load (runs on server AND client)
import type { PageLoad } from "./$types";

export const load: PageLoad = async ({ fetch, params, url, depends }) => {
  // Use SvelteKit's fetch (handles cookies, relative URLs)
  const page = url.searchParams.get("page") || "1";
  const response = await fetch(`/api/posts?page=${page}`);
  const posts = await response.json();

  // Declare dependency for invalidation
  depends("app:posts");

  return {
    posts,
    page: parseInt(page),
  };
};
```

```typescript
// src/routes/blog/+page.server.ts — Server-only load
import type { PageServerLoad } from "./$types";
import { db } from "$lib/server/database";
import { error } from "@sveltejs/kit";

export const load: PageServerLoad = async ({ params, locals, cookies }) => {
  // Access database directly (never sent to client)
  const posts = await db.post.findMany({
    where: { published: true },
    orderBy: { createdAt: "desc" },
  });

  // Access auth from hooks
  if (!locals.user) {
    error(401, "Unauthorized");
  }

  // Read cookies
  const theme = cookies.get("theme") || "light";

  return { posts, theme };
};
```

```typescript
// src/routes/+layout.ts — Layout load (shared across child routes)
import type { LayoutLoad } from "./$types";

export const load: LayoutLoad = async ({ fetch }) => {
  const user = await fetch("/api/auth/me").then((r) =>
    r.ok ? r.json() : null
  );
  return { user };
};
```

```svelte
<!-- src/routes/blog/+page.svelte — Access loaded data -->
<script lang="ts">
  import type { PageData } from "./$types";

  // SvelteKit 2 / Svelte 5 runes syntax
  let { data }: { data: PageData } = $props();

  // Or in Svelte 4: export let data: PageData;
</script>

<h1>Blog</h1>

{#each data.posts as post}
  <article>
    <a href="/blog/{post.slug}">
      <h2>{post.title}</h2>
      <p>{post.excerpt}</p>
    </a>
  </article>
{/each}
```

## Form Actions

```typescript
// src/routes/login/+page.server.ts
import type { Actions, PageServerLoad } from "./$types";
import { fail, redirect } from "@sveltejs/kit";
import { db } from "$lib/server/database";
import bcrypt from "bcrypt";

export const load: PageServerLoad = async ({ locals }) => {
  if (locals.user) redirect(303, "/dashboard");
};

export const actions: Actions = {
  // Default action (form without action attribute)
  default: async ({ request, cookies }) => {
    const formData = await request.formData();
    const email = formData.get("email") as string;
    const password = formData.get("password") as string;

    // Validation
    if (!email || !password) {
      return fail(400, {
        email,
        error: "Email and password are required",
      });
    }

    const user = await db.user.findUnique({ where: { email } });
    if (!user || !(await bcrypt.compare(password, user.passwordHash))) {
      return fail(400, {
        email,
        error: "Invalid email or password",
      });
    }

    // Set session cookie
    const session = await db.session.create({
      data: { userId: user.id },
    });
    cookies.set("session", session.id, {
      path: "/",
      httpOnly: true,
      sameSite: "lax",
      secure: true,
      maxAge: 60 * 60 * 24 * 30,
    });

    redirect(303, "/dashboard");
  },
};
```

```svelte
<!-- src/routes/login/+page.svelte -->
<script lang="ts">
  import type { ActionData, PageData } from "./$types";
  import { enhance } from "$app/forms";

  let { form }: { form: ActionData } = $props();
</script>

<!-- use:enhance for progressive enhancement (works without JS too) -->
<form method="POST" use:enhance>
  <label>
    Email
    <input name="email" type="email" value={form?.email ?? ""} required />
  </label>

  <label>
    Password
    <input name="password" type="password" required />
  </label>

  {#if form?.error}
    <p class="error">{form.error}</p>
  {/if}

  <button type="submit">Sign In</button>
</form>
```

### Named Actions

```typescript
// +page.server.ts with named actions
export const actions: Actions = {
  create: async ({ request }) => {
    const data = await request.formData();
    await db.todo.create({ data: { text: data.get("text") as string } });
  },

  delete: async ({ request }) => {
    const data = await request.formData();
    await db.todo.delete({ where: { id: data.get("id") as string } });
  },

  toggle: async ({ request }) => {
    const data = await request.formData();
    const id = data.get("id") as string;
    const todo = await db.todo.findUnique({ where: { id } });
    await db.todo.update({
      where: { id },
      data: { done: !todo?.done },
    });
  },
};
```

```svelte
<!-- Use ?/actionName in form action attribute -->
<form method="POST" action="?/create" use:enhance>
  <input name="text" placeholder="New todo" required />
  <button>Add</button>
</form>

{#each data.todos as todo}
  <div>
    <form method="POST" action="?/toggle" use:enhance>
      <input type="hidden" name="id" value={todo.id} />
      <button>{todo.done ? "✓" : "○"} {todo.text}</button>
    </form>

    <form method="POST" action="?/delete" use:enhance>
      <input type="hidden" name="id" value={todo.id} />
      <button>Delete</button>
    </form>
  </div>
{/each}
```

## API Routes

```typescript
// src/routes/api/posts/+server.ts
import type { RequestHandler } from "./$types";
import { json, error } from "@sveltejs/kit";
import { db } from "$lib/server/database";

export const GET: RequestHandler = async ({ url, locals }) => {
  const page = parseInt(url.searchParams.get("page") || "1");
  const limit = 10;

  const posts = await db.post.findMany({
    skip: (page - 1) * limit,
    take: limit,
    orderBy: { createdAt: "desc" },
  });

  return json({ posts, page, hasMore: posts.length === limit });
};

export const POST: RequestHandler = async ({ request, locals }) => {
  if (!locals.user) error(401, "Unauthorized");

  const body = await request.json();
  const post = await db.post.create({
    data: {
      title: body.title,
      content: body.content,
      authorId: locals.user.id,
    },
  });

  return json(post, { status: 201 });
};

export const DELETE: RequestHandler = async ({ params, locals }) => {
  if (!locals.user) error(401, "Unauthorized");

  await db.post.delete({ where: { id: params.id } });
  return new Response(null, { status: 204 });
};
```

## Hooks

```typescript
// src/hooks.server.ts — Server hooks
import type { Handle, HandleServerError, HandleFetch } from "@sveltejs/kit";
import { sequence } from "@sveltejs/kit/hooks";
import { db } from "$lib/server/database";

// Authentication hook
const auth: Handle = async ({ event, resolve }) => {
  const sessionId = event.cookies.get("session");

  if (sessionId) {
    const session = await db.session.findUnique({
      where: { id: sessionId },
      include: { user: true },
    });

    if (session) {
      event.locals.user = session.user;
    }
  }

  return resolve(event);
};

// Security headers
const headers: Handle = async ({ event, resolve }) => {
  const response = await resolve(event);
  response.headers.set("X-Frame-Options", "DENY");
  response.headers.set("X-Content-Type-Options", "nosniff");
  return response;
};

// Logging
const logging: Handle = async ({ event, resolve }) => {
  const start = Date.now();
  const response = await resolve(event);
  console.log(`${event.request.method} ${event.url.pathname} - ${Date.now() - start}ms`);
  return response;
};

// Chain hooks with sequence()
export const handle = sequence(logging, auth, headers);

// Error handling
export const handleError: HandleServerError = async ({ error, event }) => {
  console.error(`Error on ${event.url.pathname}:`, error);
  return {
    message: "An unexpected error occurred",
    code: "UNEXPECTED",
  };
};

// Modify fetch requests (add headers, rewrite URLs)
export const handleFetch: HandleFetch = async ({ request, fetch }) => {
  if (request.url.startsWith("https://api.internal.com")) {
    request.headers.set("Authorization", `Bearer ${API_SECRET}`);
  }
  return fetch(request);
};
```

```typescript
// src/app.d.ts — Type declarations for locals
declare global {
  namespace App {
    interface Locals {
      user: {
        id: string;
        email: string;
        name: string;
      } | null;
    }

    interface Error {
      message: string;
      code?: string;
    }

    interface PageData {}
    interface PageState {}
    interface Platform {}
  }
}

export {};
```

## Page Options (Per-Route)

```typescript
// In any +page.ts or +layout.ts:

// Prerender this page at build time
export const prerender = true;
// Options: true, false, "auto"

// Disable SSR (SPA mode for this page)
export const ssr = false;

// Disable client-side routing (full page reloads)
export const csr = false;

// Trailing slash behavior
export const trailingSlash = "never";
// Options: "never", "always", "ignore"
```

## Navigation & Page Store

```svelte
<script>
  import { goto, invalidate, invalidateAll } from "$app/navigation";
  import { page, navigating } from "$app/stores";
  import { browser } from "$app/environment";

  // Current page info
  $: currentPath = $page.url.pathname;
  $: searchParams = $page.url.searchParams;
  $: routeParams = $page.params;
  $: pageData = $page.data;
  $: pageError = $page.error;

  // Programmatic navigation
  function goToPost(slug: string) {
    goto(`/blog/${slug}`);
    // Options: { replaceState, keepFocus, noScroll, invalidateAll }
  }

  // Invalidate and reload data
  async function refresh() {
    await invalidate("app:posts"); // invalidate by dependency
    await invalidateAll();          // reload all load functions
  }

  // Loading indicator
  $: isNavigating = !!$navigating;
</script>

{#if isNavigating}
  <div class="loading-bar" />
{/if}

<!-- Active link styling -->
<nav>
  <a href="/" class:active={$page.url.pathname === "/"}>Home</a>
  <a href="/blog" class:active={$page.url.pathname.startsWith("/blog")}>Blog</a>
</nav>
```

## Environment Variables

```bash
# .env
PUBLIC_API_URL=https://api.example.com    # Available everywhere
DATABASE_URL=postgres://...               # Server only
SECRET_KEY=abc123                         # Server only
```

```typescript
// Access in code:
import { env } from "$env/dynamic/private";     // Server: runtime vars
import { env } from "$env/dynamic/public";       // Both: runtime public vars
import { DATABASE_URL } from "$env/static/private";    // Server: build-time (tree-shakeable)
import { PUBLIC_API_URL } from "$env/static/public";   // Both: build-time

// Rule: PUBLIC_ prefix = available in client code
//       No prefix = server-only (won't compile if imported in client)
```

## Configuration

```javascript
// svelte.config.js
import adapter from "@sveltejs/adapter-auto"; // auto-detect platform
// import adapter from "@sveltejs/adapter-vercel";
// import adapter from "@sveltejs/adapter-node";
// import adapter from "@sveltejs/adapter-cloudflare";
// import adapter from "@sveltejs/adapter-static";
import { vitePreprocess } from "@sveltejs/vite-plugin-svelte";

export default {
  preprocess: vitePreprocess(), // TypeScript, SCSS, etc.

  kit: {
    adapter: adapter(),

    // Path aliases
    alias: {
      $components: "src/lib/components",
      $utils: "src/lib/utils",
    },

    // CSP headers
    csp: {
      directives: {
        "script-src": ["self"],
      },
    },

    // CSRF protection (enabled by default)
    csrf: {
      checkOrigin: true,
    },

    // Prerender options
    prerender: {
      handleMissingId: "warn",
      handleHttpError: "warn",
      entries: ["*"], // Crawl all discovered links
    },
  },
};
```

## Deployment

```bash
# Auto-detect platform
npm i -D @sveltejs/adapter-auto
# Detects: Vercel, Netlify, Cloudflare Pages, Azure

# Node.js server
npm i -D @sveltejs/adapter-node
# Build: npm run build
# Run: node build/index.js
# PORT env var controls port

# Static site
npm i -D @sveltejs/adapter-static
# Requires: export const prerender = true in root +layout.ts

# Vercel
npm i -D @sveltejs/adapter-vercel
# Push to GitHub → Vercel auto-deploys

# Cloudflare Pages
npm i -D @sveltejs/adapter-cloudflare
# Deploy: npx wrangler pages deploy .svelte-kit/cloudflare
```

## Gotchas

1. **`+page.ts` vs `+page.server.ts`** — `.ts` runs on both server and client (universal load). `.server.ts` runs only on the server. Use `.server.ts` when accessing databases, secrets, or server-only APIs. Use `.ts` when the load function can run on either side (e.g., calling a public API).

2. **Forms work without JavaScript** — SvelteKit form actions use progressive enhancement. Without `use:enhance`, forms do a full-page POST. With `use:enhance`, they submit via fetch and update the page reactively.

3. **`$lib/server` is server-only by convention** — SvelteKit enforces that files in `src/lib/server/` can only be imported from server-side code. If you try to import them in a component, you get a build error.

4. **Layout data is inherited** — Data from `+layout.ts` is available in all child routes. But `+page.ts` load function receives `parent()` to access parent data, NOT direct access. Call `const parentData = await parent()`.

5. **`goto()` only works in the browser** — Don't call `goto()` in a load function. Use `redirect(303, '/path')` from `@sveltejs/kit` instead. `goto()` is for client-side navigation from components.

6. **Cookie changes need `path: '/'`** — When setting cookies with `cookies.set()`, always specify `path: '/'`. Without it, the cookie is scoped to the current route path, causing confusing behavior across routes.
