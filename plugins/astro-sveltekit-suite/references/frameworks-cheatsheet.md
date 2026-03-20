# Astro & SvelteKit Cheatsheet

## Astro Quick Reference

### CLI
```bash
npm create astro@latest my-site
npm create astro@latest my-site -- --template blog
npx astro add react          # Add React integration
npx astro add tailwind       # Add Tailwind
npx astro dev                # Start dev server
npx astro build              # Build for production
npx astro preview            # Preview build
npx astro check              # Type-check .astro files
```

### Component Anatomy
```astro
---
// Server-side script (runs at build/request time)
import Component from "./Component.astro";
const { title } = Astro.props;
const data = await fetch("...").then(r => r.json());
---
<h1>{title}</h1>
{data.map(item => <p>{item.name}</p>)}
<slot />

<style>
  /* Scoped to this component */
  h1 { color: navy; }
</style>

<script>
  // Client-side JS (bundled)
  document.querySelector("h1")?.addEventListener("click", () => {});
</script>
```

### Client Directives
| Directive | When JS Loads |
|-----------|--------------|
| (none) | Never (static HTML) |
| `client:load` | Immediately |
| `client:idle` | When browser idle |
| `client:visible` | When scrolled into view |
| `client:media="()"` | When media query matches |
| `client:only="react"` | Client-only (no SSR) |

### Content Collections
```typescript
// src/content/config.ts
import { defineCollection, z } from "astro:content";
const blog = defineCollection({
  type: "content",
  schema: z.object({
    title: z.string(),
    date: z.coerce.date(),
    tags: z.array(z.string()).default([]),
  }),
});
export const collections = { blog };
```

```astro
---
import { getCollection } from "astro:content";
const posts = await getCollection("blog");
---
```

### Output Modes
| Mode | Config | Behavior |
|------|--------|----------|
| Static | `output: "static"` | All pages built at build time |
| Server | `output: "server"` | All pages rendered on request |
| Hybrid | `output: "hybrid"` | Static default, opt-in SSR |

### Routing
```
pages/index.astro         → /
pages/about.astro         → /about
pages/blog/[slug].astro   → /blog/my-post
pages/[...path].astro     → /any/nested/path
pages/api/data.json.ts    → /api/data.json
```

### API Endpoints
```typescript
import type { APIRoute } from "astro";
export const GET: APIRoute = async ({ params, request }) => {
  return new Response(JSON.stringify({ ok: true }));
};
```

---

## SvelteKit Quick Reference

### CLI
```bash
npx sv create my-app
npm run dev               # Start dev server
npm run build             # Build
npm run preview           # Preview build
npx svelte-check          # Type check
```

### File Conventions
| File | Purpose |
|------|---------|
| `+page.svelte` | Page component |
| `+page.ts` | Universal load function |
| `+page.server.ts` | Server-only load + form actions |
| `+layout.svelte` | Layout component |
| `+layout.ts` | Layout load function |
| `+layout.server.ts` | Server layout load |
| `+server.ts` | API endpoint |
| `+error.svelte` | Error page |

### Load Functions
```typescript
// +page.ts (universal — runs server AND client)
export const load = async ({ fetch, params }) => {
  const res = await fetch(`/api/posts/${params.slug}`);
  return { post: await res.json() };
};

// +page.server.ts (server only — DB access, secrets)
export const load = async ({ locals }) => {
  return { user: locals.user };
};
```

### Form Actions
```typescript
// +page.server.ts
export const actions = {
  default: async ({ request }) => {
    const data = await request.formData();
    // Process...
    return fail(400, { error: "message" }); // or just return
  },
  named: async ({ request }) => { /* ... */ },
};
```

```svelte
<form method="POST" use:enhance>...</form>
<form method="POST" action="?/named" use:enhance>...</form>
```

### API Routes
```typescript
// +server.ts
import { json } from "@sveltejs/kit";
export const GET = async () => json({ data: [] });
export const POST = async ({ request }) => {
  const body = await request.json();
  return json(body, { status: 201 });
};
```

### Hooks
```typescript
// hooks.server.ts
export const handle = async ({ event, resolve }) => {
  event.locals.user = await getUser(event.cookies.get("session"));
  return resolve(event);
};
```

### Page Options
```typescript
export const prerender = true;    // Static at build time
export const ssr = false;         // SPA mode (client only)
export const csr = false;         // No client JS
export const trailingSlash = "never";
```

### Navigation
```svelte
<script>
  import { goto, invalidate } from "$app/navigation";
  import { page } from "$app/stores";

  goto("/path");                   // Navigate
  invalidate("app:data");          // Reload data
  $page.url.pathname;              // Current path
</script>
```

### Environment Variables
```bash
PUBLIC_API=https://...   # Available everywhere
DB_URL=postgres://...    # Server only (no PUBLIC_ prefix)
```

```typescript
import { PUBLIC_API } from "$env/static/public";
import { DB_URL } from "$env/static/private";
```

### Adapters
| Adapter | Install |
|---------|---------|
| Auto-detect | `@sveltejs/adapter-auto` |
| Node.js | `@sveltejs/adapter-node` |
| Static | `@sveltejs/adapter-static` |
| Vercel | `@sveltejs/adapter-vercel` |
| Cloudflare | `@sveltejs/adapter-cloudflare` |
| Netlify | `@sveltejs/adapter-netlify` |

---

## Framework Decision

| Scenario | Choose |
|----------|--------|
| Blog, docs, marketing site | **Astro** |
| Content + some interactivity | **Astro** with islands |
| Full-stack app, dashboard | **SvelteKit** |
| Need multiple UI frameworks | **Astro** |
| Progressive enhancement priority | **SvelteKit** |
| Smallest possible JS bundle | **Astro** (zero JS default) |
| Complex forms and mutations | **SvelteKit** |
| E-commerce with dynamic pricing | **SvelteKit** |
| Static docs with search | **Astro Starlight** |
