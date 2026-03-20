---
name: astro-development
description: >
  Astro framework — content sites, island architecture, content collections,
  SSG/SSR/hybrid rendering, integrations, middleware, and deployment.
  Triggers: "astro", "astro project", "astro content", "astro islands",
  "content collections", "astro blog", "astro docs site", "astro components".
  NOT for: full-stack apps needing complex state (use sveltekit-development or Next.js).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Astro Development

## Quick Start

```bash
# Create new project
npm create astro@latest my-site
# Or with template:
npm create astro@latest my-site -- --template blog
npm create astro@latest my-site -- --template docs
npm create astro@latest my-site -- --template starlight  # docs framework

# Add integrations
npx astro add react        # React support
npx astro add tailwind     # Tailwind CSS
npx astro add mdx          # MDX support
npx astro add sitemap      # Auto sitemap
npx astro add image        # Image optimization (deprecated — use astro:assets)
```

## Project Structure

```
src/
├── pages/           # File-based routing (.astro, .md, .mdx)
│   ├── index.astro
│   ├── about.astro
│   ├── blog/
│   │   ├── index.astro
│   │   └── [slug].astro
│   └── api/
│       └── posts.json.ts  # API endpoint
├── layouts/         # Reusable page layouts
│   └── BaseLayout.astro
├── components/      # UI components (.astro, .tsx, .vue, .svelte)
│   ├── Header.astro
│   ├── Counter.tsx   # React island
│   └── Nav.svelte    # Svelte island
├── content/         # Content collections
│   ├── config.ts    # Collection schemas
│   ├── blog/        # Blog collection
│   │   ├── post-1.md
│   │   └── post-2.mdx
│   └── authors/     # Authors collection
│       └── alice.json
├── styles/          # Global styles
│   └── global.css
├── assets/          # Optimized assets (images, fonts)
│   └── hero.png
└── middleware.ts    # Request middleware
astro.config.mjs     # Astro configuration
```

## Astro Components (.astro)

```astro
---
// Component script (runs at build time / server)
import Layout from "../layouts/BaseLayout.astro";
import Card from "../components/Card.astro";

// Props with TypeScript
interface Props {
  title: string;
  description?: string;
}

const { title, description = "Default description" } = Astro.props;

// Fetch data at build time
const response = await fetch("https://api.example.com/posts");
const posts = await response.json();

// Access URL, params, cookies
const { pathname } = Astro.url;
const theme = Astro.cookies.get("theme")?.value || "light";
---

<Layout title={title}>
  <h1>{title}</h1>
  <p>{description}</p>

  <!-- Loop -->
  <ul>
    {posts.map((post) => (
      <li>
        <Card title={post.title} url={`/blog/${post.slug}`} />
      </li>
    ))}
  </ul>

  <!-- Conditional -->
  {posts.length === 0 && <p>No posts yet.</p>}

  <!-- Slot (children) -->
  <slot />

  <!-- Named slots -->
  <slot name="sidebar" />
</Layout>

<style>
  /* Scoped by default — only affects this component */
  h1 {
    color: navy;
    font-size: 2rem;
  }
</style>

<style is:global>
  /* Global styles */
  body { margin: 0; }
</style>

<script>
  // Client-side JavaScript (bundled, deduped)
  console.log("This runs in the browser");
</script>
```

## Island Architecture (Client Directives)

```astro
---
import Counter from "../components/Counter.tsx";
import Gallery from "../components/Gallery.svelte";
import Comments from "../components/Comments.vue";
---

<!-- No directive = server-rendered HTML only, zero JS -->
<Counter />

<!-- Load JS immediately (above the fold, critical interactivity) -->
<Counter client:load />

<!-- Load JS when visible in viewport (lazy) -->
<Gallery client:visible />

<!-- Load JS when browser is idle -->
<Comments client:idle />

<!-- Load JS only on specific media query -->
<Counter client:media="(max-width: 768px)" />

<!-- Only render on client (no SSR) -->
<Counter client:only="react" />
```

### When to use each directive

| Directive | Use For | JS Loaded |
|-----------|---------|-----------|
| (none) | Static content, no interactivity | Never |
| `client:load` | Critical above-fold interaction (nav, hero CTA) | Immediately |
| `client:idle` | Non-critical interaction (chat widget, analytics) | When idle |
| `client:visible` | Below-fold content (comments, carousels) | When scrolled to |
| `client:media` | Mobile-only or desktop-only components | When media matches |
| `client:only` | Components that can't SSR (canvas, WebGL) | Client only |

## Content Collections

```typescript
// src/content/config.ts
import { defineCollection, z, reference } from "astro:content";

const blog = defineCollection({
  type: "content", // Markdown/MDX
  schema: ({ image }) =>
    z.object({
      title: z.string(),
      description: z.string(),
      pubDate: z.coerce.date(),
      updatedDate: z.coerce.date().optional(),
      heroImage: image().optional(), // optimized image
      tags: z.array(z.string()).default([]),
      author: reference("authors"), // reference another collection
      draft: z.boolean().default(false),
    }),
});

const authors = defineCollection({
  type: "data", // JSON/YAML
  schema: z.object({
    name: z.string(),
    bio: z.string(),
    avatar: z.string().url(),
    social: z
      .object({
        twitter: z.string().optional(),
        github: z.string().optional(),
      })
      .optional(),
  }),
});

export const collections = { blog, authors };
```

```astro
---
// src/pages/blog/index.astro — List all posts
import { getCollection } from "astro:content";
import Layout from "../../layouts/BaseLayout.astro";

const posts = await getCollection("blog", ({ data }) => {
  // Filter: exclude drafts in production
  return import.meta.env.PROD ? !data.draft : true;
});

// Sort by date
const sorted = posts.sort(
  (a, b) => b.data.pubDate.valueOf() - a.data.pubDate.valueOf()
);
---

<Layout title="Blog">
  {sorted.map((post) => (
    <article>
      <a href={`/blog/${post.slug}`}>
        <h2>{post.data.title}</h2>
        <time datetime={post.data.pubDate.toISOString()}>
          {post.data.pubDate.toLocaleDateString()}
        </time>
        <p>{post.data.description}</p>
      </a>
    </article>
  ))}
</Layout>
```

```astro
---
// src/pages/blog/[...slug].astro — Individual post
import { getCollection } from "astro:content";
import Layout from "../../layouts/BaseLayout.astro";

export async function getStaticPaths() {
  const posts = await getCollection("blog");
  return posts.map((post) => ({
    params: { slug: post.slug },
    props: { post },
  }));
}

const { post } = Astro.props;
const { Content, headings } = await post.render();
---

<Layout title={post.data.title}>
  <article>
    <h1>{post.data.title}</h1>
    <time>{post.data.pubDate.toLocaleDateString()}</time>
    {post.data.heroImage && (
      <img src={post.data.heroImage.src} alt="" />
    )}
    <!-- Renders the markdown content -->
    <Content />
  </article>

  <!-- Table of contents from headings -->
  <nav>
    {headings.map((h) => (
      <a href={`#${h.slug}`} style={`margin-left: ${(h.depth - 2) * 1}rem`}>
        {h.text}
      </a>
    ))}
  </nav>
</Layout>
```

## Routing

```
src/pages/
├── index.astro          → /
├── about.astro          → /about
├── blog/
│   ├── index.astro      → /blog
│   └── [slug].astro     → /blog/my-post (dynamic)
├── [lang]/
│   └── [slug].astro     → /en/about (multiple params)
├── [...slug].astro      → /any/nested/path (catch-all)
└── 404.astro            → Custom 404 page
```

```astro
---
// Dynamic route: src/pages/blog/[slug].astro
// For SSG, must export getStaticPaths:
export async function getStaticPaths() {
  const posts = await getCollection("blog");
  return posts.map((post) => ({
    params: { slug: post.slug },
    props: { post },
  }));
}

// For SSR (output: "server"), access params directly:
const { slug } = Astro.params;
---
```

## API Endpoints

```typescript
// src/pages/api/posts.json.ts
import type { APIRoute } from "astro";
import { getCollection } from "astro:content";

export const GET: APIRoute = async ({ request, url }) => {
  const posts = await getCollection("blog");
  const tag = url.searchParams.get("tag");

  const filtered = tag
    ? posts.filter((p) => p.data.tags.includes(tag))
    : posts;

  return new Response(JSON.stringify(filtered), {
    headers: { "Content-Type": "application/json" },
  });
};

export const POST: APIRoute = async ({ request }) => {
  const body = await request.json();

  // Process form submission, webhook, etc.

  return new Response(JSON.stringify({ success: true }), {
    status: 200,
    headers: { "Content-Type": "application/json" },
  });
};
```

## Middleware

```typescript
// src/middleware.ts
import { defineMiddleware, sequence } from "astro:middleware";

const auth = defineMiddleware(async (context, next) => {
  const token = context.cookies.get("token")?.value;

  if (context.url.pathname.startsWith("/dashboard")) {
    if (!token) {
      return context.redirect("/login");
    }
    // Add user to locals (available in components via Astro.locals)
    context.locals.user = await verifyToken(token);
  }

  return next();
});

const logging = defineMiddleware(async (context, next) => {
  const start = Date.now();
  const response = await next();
  console.log(`${context.request.method} ${context.url.pathname} - ${Date.now() - start}ms`);
  return response;
});

// Chain multiple middleware
export const onRequest = sequence(logging, auth);
```

## Configuration

```javascript
// astro.config.mjs
import { defineConfig } from "astro/config";
import react from "@astrojs/react";
import tailwind from "@astrojs/tailwind";
import mdx from "@astrojs/mdx";
import sitemap from "@astrojs/sitemap";
import vercel from "@astrojs/vercel/serverless";

export default defineConfig({
  site: "https://mysite.com",

  // Rendering mode
  output: "hybrid", // "static" (default), "server", "hybrid"
  // hybrid: static by default, opt-in to SSR per page

  // Adapter for deployment target
  adapter: vercel(),
  // Other adapters: node(), cloudflare(), netlify(), deno()

  // Integrations
  integrations: [
    react(),
    tailwind(),
    mdx(),
    sitemap({
      filter: (page) => !page.includes("/admin/"),
    }),
  ],

  // Image optimization
  image: {
    service: {
      entrypoint: "astro/assets/services/sharp",
    },
    domains: ["cdn.example.com"],
    remotePatterns: [{ protocol: "https" }],
  },

  // Markdown config
  markdown: {
    shikiConfig: {
      theme: "github-dark",
      wrap: true,
    },
    remarkPlugins: [],
    rehypePlugins: [],
  },

  // Redirects
  redirects: {
    "/old-page": "/new-page",
    "/blog/[...slug]": "/articles/[...slug]",
  },

  // Dev server
  server: { port: 4321, host: true },

  // Build options
  build: {
    inlineStylesheets: "auto",
  },

  // i18n
  i18n: {
    defaultLocale: "en",
    locales: ["en", "es", "fr"],
    routing: {
      prefixDefaultLocale: false,
    },
  },
});
```

## Image Optimization

```astro
---
import { Image, getImage } from "astro:assets";
import heroImage from "../assets/hero.png";
---

<!-- Optimized image component -->
<Image
  src={heroImage}
  alt="Hero image"
  width={800}
  height={400}
  format="webp"
  quality={80}
  loading="lazy"
/>

<!-- Remote images (must be in domains/remotePatterns config) -->
<Image
  src="https://cdn.example.com/photo.jpg"
  alt="Remote photo"
  width={600}
  height={400}
  inferSize  <!-- auto-detect dimensions for remote -->
/>

<!-- Background image via getImage -->
{async () => {
  const bg = await getImage({ src: heroImage, format: "webp" });
  return <div style={`background-image: url(${bg.src})`} />;
}}
```

## View Transitions

```astro
---
// src/layouts/BaseLayout.astro
import { ViewTransitions } from "astro:transitions";
---

<html>
  <head>
    <ViewTransitions />
  </head>
  <body>
    <!-- Elements with same transition:name animate between pages -->
    <header transition:animate="none">
      <nav>...</nav>
    </header>

    <main transition:animate="slide">
      <slot />
    </main>
  </body>
</html>
```

```astro
<!-- Per-element transition names for cross-page animations -->
<img
  src={post.data.heroImage.src}
  transition:name={`hero-${post.slug}`}
  transition:animate="fade"
/>
```

## Deployment

```bash
# Vercel
npx astro add vercel
# Deploy: push to GitHub → Vercel auto-deploys

# Netlify
npx astro add netlify
# Deploy: push to GitHub → Netlify auto-deploys

# Cloudflare Pages
npx astro add cloudflare
# Deploy: npx wrangler pages deploy dist/

# Node.js server
npx astro add node
# Build: npx astro build
# Run: node dist/server/entry.mjs

# Static (any hosting)
# output: "static" in config
# Build: npx astro build
# Upload dist/ to any static host
```

## Gotchas

1. **Content collections require `astro:content` import** — don't manually read files with `fs`. Use `getCollection()` and `getEntry()` for type-safe access with schema validation.

2. **Styles are scoped by default** — CSS in `<style>` tags only applies to the current component. Use `is:global` for global styles, or put them in `src/styles/`.

3. **Component script runs on the server** — the `---` fence runs at build/request time, NOT in the browser. Use `<script>` tags for client-side code.

4. **Don't forget `getStaticPaths` for dynamic routes in static mode** — without it, Astro doesn't know which pages to generate. In server/hybrid mode, it's optional.

5. **Island hydration order matters** — `client:load` blocks rendering. For below-fold content, always use `client:visible` or `client:idle` to avoid blocking the initial paint.

6. **Image optimization needs the `astro:assets` import** — not a file path string. Import the image as a module: `import hero from '../assets/hero.png'` then pass to `<Image src={hero} />`.
