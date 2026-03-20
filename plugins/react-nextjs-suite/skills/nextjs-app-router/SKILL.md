---
name: nextjs-app-router
description: >
  Next.js App Router deep dive — layouts, loading states, error handling,
  route handlers, server actions, middleware, caching, ISR, parallel routes,
  intercepting routes, and metadata.
  Triggers: "next.js app router", "nextjs routing", "server actions",
  "nextjs middleware", "nextjs caching", "nextjs ISR", "parallel routes".
  NOT for: Pages Router (legacy). NOT for: general React patterns (use react-patterns).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Next.js App Router

## Project Structure

```
app/
├── layout.tsx              # Root layout (required)
├── page.tsx                # Home page (/)
├── loading.tsx             # Loading UI for /
├── error.tsx               # Error boundary for /
├── not-found.tsx           # 404 page
├── global-error.tsx        # Root error boundary
├── dashboard/
│   ├── layout.tsx          # Nested layout
│   ├── page.tsx            # /dashboard
│   ├── loading.tsx         # Loading for /dashboard
│   ├── settings/
│   │   └── page.tsx        # /dashboard/settings
│   └── [teamId]/
│       └── page.tsx        # /dashboard/:teamId
├── blog/
│   ├── page.tsx            # /blog
│   └── [slug]/
│       ├── page.tsx        # /blog/:slug
│       └── opengraph-image.tsx  # Dynamic OG image
├── api/
│   └── webhooks/
│       └── route.ts        # API route handler
├── (marketing)/            # Route group (no URL segment)
│   ├── about/page.tsx      # /about
│   └── pricing/page.tsx    # /pricing
└── @modal/                 # Parallel route
    └── login/page.tsx      # Intercepted modal
```

## Layouts

```tsx
// app/layout.tsx — Root layout (wraps entire app)
import type { Metadata } from "next";
import { Inter } from "next/font/google";
import "./globals.css";

const inter = Inter({ subsets: ["latin"] });

export const metadata: Metadata = {
  title: { default: "My App", template: "%s | My App" },
  description: "Built with Next.js",
  metadataBase: new URL("https://myapp.com"),
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body className={inter.className}>
        <nav>{/* Shared navigation */}</nav>
        <main>{children}</main>
      </body>
    </html>
  );
}
```

```tsx
// app/dashboard/layout.tsx — Nested layout (persists across dashboard pages)
import { redirect } from "next/navigation";
import { getSession } from "@/lib/auth";

export default async function DashboardLayout({ children }: { children: React.ReactNode }) {
  const session = await getSession();
  if (!session) redirect("/login");

  return (
    <div className="flex">
      <aside className="w-64">
        <DashboardNav user={session.user} />
      </aside>
      <div className="flex-1 p-6">{children}</div>
    </div>
  );
}
```

## Dynamic Routes

```tsx
// app/blog/[slug]/page.tsx
import { notFound } from "next/navigation";

interface Props {
  params: Promise<{ slug: string }>;
}

// Generate static pages at build time
export async function generateStaticParams() {
  const posts = await db.post.findMany({ select: { slug: true } });
  return posts.map((post) => ({ slug: post.slug }));
}

// Dynamic metadata per page
export async function generateMetadata({ params }: Props) {
  const { slug } = await params;
  const post = await db.post.findUnique({ where: { slug } });
  if (!post) return {};

  return {
    title: post.title,
    description: post.excerpt,
    openGraph: { title: post.title, description: post.excerpt, type: "article" },
  };
}

export default async function BlogPost({ params }: Props) {
  const { slug } = await params;
  const post = await db.post.findUnique({ where: { slug } });
  if (!post) notFound();

  return (
    <article>
      <h1>{post.title}</h1>
      <time>{post.publishedAt.toLocaleDateString()}</time>
      <div dangerouslySetInnerHTML={{ __html: post.contentHtml }} />
    </article>
  );
}
```

## Loading & Error States

```tsx
// app/dashboard/loading.tsx — Shows while page.tsx loads
export default function DashboardLoading() {
  return (
    <div className="animate-pulse space-y-4">
      <div className="h-8 bg-gray-200 rounded w-1/3" />
      <div className="grid grid-cols-3 gap-4">
        {[1, 2, 3].map((i) => (
          <div key={i} className="h-32 bg-gray-200 rounded" />
        ))}
      </div>
    </div>
  );
}

// app/dashboard/error.tsx — Error boundary (must be "use client")
"use client";

export default function DashboardError({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  return (
    <div className="p-6 bg-red-50 rounded-lg">
      <h2 className="text-red-800 font-bold">Something went wrong</h2>
      <p className="text-red-600">{error.message}</p>
      <button
        onClick={reset}
        className="mt-4 px-4 py-2 bg-red-600 text-white rounded"
      >
        Try again
      </button>
    </div>
  );
}

// app/not-found.tsx — Custom 404
import Link from "next/link";

export default function NotFound() {
  return (
    <div className="text-center py-20">
      <h1 className="text-6xl font-bold">404</h1>
      <p className="mt-4 text-gray-600">Page not found</p>
      <Link href="/" className="mt-6 inline-block text-blue-600 underline">
        Go home
      </Link>
    </div>
  );
}
```

## Server Actions

```tsx
// app/dashboard/actions.ts
"use server";

import { revalidatePath } from "next/cache";
import { redirect } from "next/navigation";
import { z } from "zod";

const CreatePostSchema = z.object({
  title: z.string().min(1).max(200),
  content: z.string().min(1),
  published: z.coerce.boolean().default(false),
});

export async function createPost(formData: FormData) {
  const session = await getSession();
  if (!session) throw new Error("Unauthorized");

  const parsed = CreatePostSchema.safeParse({
    title: formData.get("title"),
    content: formData.get("content"),
    published: formData.get("published"),
  });

  if (!parsed.success) {
    return { error: parsed.error.flatten().fieldErrors };
  }

  const post = await db.post.create({
    data: { ...parsed.data, authorId: session.user.id },
  });

  revalidatePath("/dashboard/posts");
  redirect(`/blog/${post.slug}`);
}

export async function deletePost(postId: string) {
  const session = await getSession();
  if (!session) throw new Error("Unauthorized");

  await db.post.delete({ where: { id: postId, authorId: session.user.id } });
  revalidatePath("/dashboard/posts");
}
```

```tsx
// Using server actions in a form
"use client";

import { useActionState } from "react";
import { createPost } from "./actions";

export function CreatePostForm() {
  const [state, formAction, isPending] = useActionState(createPost, null);

  return (
    <form action={formAction}>
      <input name="title" required />
      {state?.error?.title && <p className="text-red-500">{state.error.title}</p>}

      <textarea name="content" required />
      {state?.error?.content && <p className="text-red-500">{state.error.content}</p>}

      <label>
        <input type="checkbox" name="published" />
        Publish immediately
      </label>

      <button type="submit" disabled={isPending}>
        {isPending ? "Creating..." : "Create Post"}
      </button>
    </form>
  );
}
```

## Route Handlers (API Routes)

```tsx
// app/api/webhooks/route.ts
import { NextRequest, NextResponse } from "next/server";
import { headers } from "next/headers";
import crypto from "crypto";

// GET handler
export async function GET(request: NextRequest) {
  const searchParams = request.nextUrl.searchParams;
  const page = parseInt(searchParams.get("page") || "1");
  const limit = parseInt(searchParams.get("limit") || "20");

  const posts = await db.post.findMany({
    skip: (page - 1) * limit,
    take: limit,
    orderBy: { createdAt: "desc" },
  });

  return NextResponse.json({ data: posts, page, limit });
}

// POST handler with webhook signature verification
export async function POST(request: NextRequest) {
  const body = await request.text();
  const headersList = await headers();
  const signature = headersList.get("x-webhook-signature");

  // Verify webhook signature
  const expectedSig = crypto
    .createHmac("sha256", process.env.WEBHOOK_SECRET!)
    .update(body)
    .digest("hex");

  if (signature !== expectedSig) {
    return NextResponse.json({ error: "Invalid signature" }, { status: 401 });
  }

  const payload = JSON.parse(body);
  await processWebhook(payload);

  return NextResponse.json({ received: true });
}

// Dynamic route handler
// app/api/posts/[id]/route.ts
export async function PATCH(
  request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const { id } = await params;
  const body = await request.json();

  const updated = await db.post.update({
    where: { id },
    data: body,
  });

  return NextResponse.json(updated);
}
```

## Middleware

```tsx
// middleware.ts (project root)
import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;

  // Auth check
  const token = request.cookies.get("session-token")?.value;
  if (pathname.startsWith("/dashboard") && !token) {
    const loginUrl = new URL("/login", request.url);
    loginUrl.searchParams.set("callbackUrl", pathname);
    return NextResponse.redirect(loginUrl);
  }

  // Add headers
  const response = NextResponse.next();
  response.headers.set("x-request-id", crypto.randomUUID());

  // Geolocation-based redirect
  const country = request.geo?.country;
  if (pathname === "/" && country === "DE") {
    return NextResponse.redirect(new URL("/de", request.url));
  }

  return response;
}

export const config = {
  matcher: [
    // Match all paths except static files and API
    "/((?!_next/static|_next/image|favicon.ico|api).*)",
  ],
};
```

## Caching & Revalidation

```tsx
// Fetch with caching (default: cached indefinitely)
const data = await fetch("https://api.example.com/posts", {
  cache: "force-cache",  // Default — cached until revalidated
});

// Time-based revalidation (ISR)
const data = await fetch("https://api.example.com/posts", {
  next: { revalidate: 3600 },  // Revalidate every hour
});

// No caching
const data = await fetch("https://api.example.com/posts", {
  cache: "no-store",  // Always fresh
});

// Page-level revalidation
export const revalidate = 60;  // Revalidate this page every 60 seconds

// Dynamic rendering (opt out of static)
export const dynamic = "force-dynamic";

// On-demand revalidation (in server actions or route handlers)
import { revalidatePath, revalidateTag } from "next/cache";

// Revalidate a specific path
revalidatePath("/blog");
revalidatePath("/blog/[slug]", "page");

// Revalidate by tag
const data = await fetch("https://api.example.com/posts", {
  next: { tags: ["posts"] },
});
// Later: revalidateTag("posts");
```

### Caching Rules

| Method | Default | Override |
|--------|---------|---------|
| `fetch()` in Server Component | Cached | `cache: "no-store"` or `next: { revalidate: N }` |
| `fetch()` in Route Handler (GET) | Cached | Same as above |
| `fetch()` in Route Handler (POST) | Not cached | N/A |
| `fetch()` in Server Action | Not cached | N/A |
| Database queries (no fetch) | Not cached | Wrap in `unstable_cache()` |
| `cookies()`, `headers()` | Forces dynamic | Cannot override |

## Parallel Routes

```tsx
// app/layout.tsx with parallel routes
export default function Layout({
  children,
  modal,
}: {
  children: React.ReactNode;
  modal: React.ReactNode;
}) {
  return (
    <>
      {children}
      {modal}
    </>
  );
}

// app/@modal/default.tsx — shown when no modal is active
export default function Default() {
  return null;
}

// app/@modal/(.)photo/[id]/page.tsx — intercepted route (shows as modal)
import { Modal } from "@/components/Modal";

export default async function PhotoModal({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const photo = await getPhoto(id);

  return (
    <Modal>
      <img src={photo.url} alt={photo.alt} />
    </Modal>
  );
}

// app/photo/[id]/page.tsx — direct navigation (full page)
export default async function PhotoPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const photo = await getPhoto(id);

  return (
    <div className="max-w-4xl mx-auto">
      <img src={photo.url} alt={photo.alt} />
      <h1>{photo.title}</h1>
    </div>
  );
}
```

## Image & Font Optimization

```tsx
import Image from "next/image";
import { Inter, Roboto_Mono } from "next/font/google";
import localFont from "next/font/local";

// Google Fonts — automatically self-hosted, no external requests
const inter = Inter({ subsets: ["latin"], display: "swap" });
const robotoMono = Roboto_Mono({ subsets: ["latin"], variable: "--font-mono" });

// Local font
const customFont = localFont({
  src: "./fonts/CustomFont.woff2",
  display: "swap",
  variable: "--font-custom",
});

// Optimized images
function Hero() {
  return (
    <div>
      {/* Remote image — must configure domains in next.config.ts */}
      <Image
        src="https://cdn.example.com/hero.jpg"
        alt="Hero"
        width={1200}
        height={630}
        priority  // Preload above-the-fold images
        className="object-cover"
      />

      {/* Fill mode — image fills container */}
      <div className="relative h-96">
        <Image
          src="/background.jpg"
          alt="Background"
          fill
          sizes="100vw"
          className="object-cover"
        />
      </div>
    </div>
  );
}
```

## Environment Variables

```bash
# .env.local (gitignored, local dev)
DATABASE_URL=postgresql://...
API_SECRET=secret123

# .env (committed, shared defaults)
NEXT_PUBLIC_APP_URL=http://localhost:3000

# .env.production (committed, production overrides)
NEXT_PUBLIC_APP_URL=https://myapp.com
```

| Prefix | Access | Usage |
|--------|--------|-------|
| `NEXT_PUBLIC_` | Client + Server | `process.env.NEXT_PUBLIC_APP_URL` anywhere |
| No prefix | Server only | `process.env.DATABASE_URL` in Server Components, Route Handlers, middleware |

## Gotchas

1. **`params` is now a Promise in Next.js 15** — You must `await params` before accessing properties. `const { slug } = await params;` not `const { slug } = params;`.

2. **`"use client"` doesn't mean "client-only rendering"** — It means the component CAN use hooks and event handlers. It still renders on the server (SSR) first, then hydrates on the client.

3. **Server Actions must be `async`** — Even if they don't await anything. The `"use server"` directive requires the function to be async.

4. **`cookies()` and `headers()` force dynamic rendering** — If any page reads cookies or headers, the entire route becomes dynamic. Move cookie/header reads to the smallest possible component and wrap with Suspense.

5. **Route Handlers GET requests are cached by default** — If your GET handler returns user-specific data, add `export const dynamic = "force-dynamic"` or use `cookies()`/`headers()` to opt out of caching.

6. **Parallel routes need a `default.tsx`** — Without it, parallel route slots show 404 on soft navigation. The `default.tsx` defines what to render when the slot doesn't have a matching page.
