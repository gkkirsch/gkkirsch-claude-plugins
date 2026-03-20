---
name: nextjs-app-router
description: >
  Next.js 15 App Router patterns — file-based routing, layouts, loading states,
  error boundaries, route groups, dynamic routes, parallel routes, and intercepting routes.
  Triggers: "next.js routing", "app router", "next.js pages", "next.js layout",
  "loading state", "error boundary", "dynamic route", "parallel route".
  NOT for: Pages Router (legacy), API routes (use data-fetching skill).
version: 1.0.0
allowed-tools: Read, Grep, Glob, Bash, Edit, Write
---

# Next.js App Router

## File Convention

| File | Purpose | Required? |
|------|---------|-----------|
| `page.tsx` | UI for a route (makes it publicly accessible) | Yes (to make route accessible) |
| `layout.tsx` | Shared UI wrapping page + children | Root layout required |
| `loading.tsx` | Instant loading UI (Suspense boundary) | Optional |
| `error.tsx` | Error boundary for the segment | Optional |
| `not-found.tsx` | 404 UI for the segment | Optional |
| `template.tsx` | Like layout but re-mounts on navigation | Optional |
| `default.tsx` | Fallback for parallel routes | For parallel routes |
| `route.ts` | API endpoint (GET, POST, etc.) | For API routes |
| `middleware.ts` | Request middleware (project root only) | Optional |

## Route Patterns

### Static Routes

```
app/
├── page.tsx              → /
├── about/page.tsx        → /about
├── blog/page.tsx         → /blog
└── contact/page.tsx      → /contact
```

### Dynamic Routes

```tsx
// app/blog/[slug]/page.tsx
export default async function BlogPost({
  params,
}: {
  params: Promise<{ slug: string }>;
}) {
  const { slug } = await params;  // Next.js 15: params is a Promise
  const post = await getPost(slug);

  if (!post) notFound();  // Triggers not-found.tsx

  return <article>{post.content}</article>;
}

// Static generation for known slugs
export async function generateStaticParams() {
  const posts = await getAllPosts();
  return posts.map((post) => ({ slug: post.slug }));
}
```

### Catch-All Routes

```tsx
// app/docs/[...slug]/page.tsx → /docs/a, /docs/a/b, /docs/a/b/c
export default async function DocsPage({
  params,
}: {
  params: Promise<{ slug: string[] }>;
}) {
  const { slug } = await params;  // ['getting-started', 'installation']
  const doc = await getDoc(slug.join('/'));
  return <div>{doc.content}</div>;
}

// app/shop/[[...category]]/page.tsx → /shop, /shop/shoes, /shop/shoes/nike
// Optional catch-all: also matches the root (/shop)
```

### Route Groups

Group routes without affecting the URL:

```
app/
├── (marketing)/
│   ├── layout.tsx        # Marketing layout (centered, no sidebar)
│   ├── page.tsx          # / (landing page)
│   ├── about/page.tsx    # /about
│   └── pricing/page.tsx  # /pricing
├── (dashboard)/
│   ├── layout.tsx        # Dashboard layout (sidebar, header)
│   ├── dashboard/page.tsx  # /dashboard
│   └── settings/page.tsx   # /settings
└── layout.tsx            # Root layout
```

## Layouts

### Root Layout (Required)

```tsx
// app/layout.tsx
import { Inter } from 'next/font/google';
import './globals.css';

const inter = Inter({ subsets: ['latin'] });

export const metadata = {
  title: { default: 'My App', template: '%s | My App' },
  description: 'My application description',
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body className={inter.className}>
        <Providers>
          {children}
        </Providers>
      </body>
    </html>
  );
}
```

### Nested Layout

```tsx
// app/(dashboard)/layout.tsx
import { Sidebar } from '@/components/layouts/sidebar';
import { Header } from '@/components/layouts/header';
import { getSession } from '@/lib/auth';
import { redirect } from 'next/navigation';

export default async function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const session = await getSession();
  if (!session) redirect('/login');

  return (
    <div className="flex h-screen">
      <Sidebar user={session.user} />
      <div className="flex-1 flex flex-col">
        <Header user={session.user} />
        <main className="flex-1 overflow-auto p-6">
          {children}
        </main>
      </div>
    </div>
  );
}
```

### Key Layout Rules

1. Layouts **persist** across navigations — state is preserved
2. Layouts don't re-render when navigating between child pages
3. Root layout MUST contain `<html>` and `<body>` tags
4. Layouts can fetch data (they're Server Components by default)
5. You cannot pass data between parent layout and children — use shared data fetching or context

## Loading States

```tsx
// app/(dashboard)/projects/loading.tsx
export default function Loading() {
  return (
    <div className="space-y-4">
      <div className="h-8 w-48 bg-gray-200 animate-pulse rounded" />
      <div className="grid grid-cols-3 gap-4">
        {[1, 2, 3, 4, 5, 6].map((i) => (
          <div key={i} className="h-32 bg-gray-200 animate-pulse rounded-lg" />
        ))}
      </div>
    </div>
  );
}
```

`loading.tsx` wraps the page in a `<Suspense>` boundary automatically. The loading UI shows instantly while the page's async data loads.

### Granular Suspense

For more control, use `Suspense` directly:

```tsx
// app/dashboard/page.tsx
import { Suspense } from 'react';

export default function Dashboard() {
  return (
    <div>
      <h1>Dashboard</h1>
      <Suspense fallback={<StatsSkeleton />}>
        <Stats />   {/* Fetches data, streams when ready */}
      </Suspense>
      <Suspense fallback={<ChartSkeleton />}>
        <RevenueChart />  {/* Independent loading */}
      </Suspense>
      <Suspense fallback={<TableSkeleton />}>
        <RecentOrders />  {/* Streams independently */}
      </Suspense>
    </div>
  );
}
```

## Error Handling

```tsx
// app/(dashboard)/error.tsx
'use client';  // Error boundaries must be Client Components

export default function Error({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  return (
    <div className="flex flex-col items-center justify-center min-h-[400px]">
      <h2 className="text-xl font-bold">Something went wrong</h2>
      <p className="text-gray-500 mt-2">{error.message}</p>
      <button
        onClick={() => reset()}
        className="mt-4 px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600"
      >
        Try again
      </button>
    </div>
  );
}
```

### Global Error Boundary

```tsx
// app/global-error.tsx — catches errors in root layout
'use client';

export default function GlobalError({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  return (
    <html>
      <body>
        <h2>Something went wrong!</h2>
        <button onClick={() => reset()}>Try again</button>
      </body>
    </html>
  );
}
```

## Not Found

```tsx
// app/not-found.tsx
import Link from 'next/link';

export default function NotFound() {
  return (
    <div className="flex flex-col items-center justify-center min-h-screen">
      <h1 className="text-6xl font-bold">404</h1>
      <p className="text-xl text-gray-500 mt-4">Page not found</p>
      <Link href="/" className="mt-6 text-blue-500 hover:underline">
        Go home
      </Link>
    </div>
  );
}
```

Trigger with `notFound()` from `next/navigation` in any Server Component.

## Parallel Routes

Render multiple pages simultaneously in the same layout:

```
app/
├── @analytics/
│   ├── page.tsx          # Analytics panel
│   └── default.tsx       # Fallback when no match
├── @team/
│   ├── page.tsx          # Team panel
│   └── default.tsx
├── layout.tsx            # Receives both as props
└── page.tsx              # Main content
```

```tsx
// app/layout.tsx
export default function Layout({
  children,
  analytics,
  team,
}: {
  children: React.ReactNode;
  analytics: React.ReactNode;
  team: React.ReactNode;
}) {
  return (
    <div>
      {children}
      <div className="grid grid-cols-2 gap-4">
        {analytics}
        {team}
      </div>
    </div>
  );
}
```

### Modal Pattern with Parallel Routes

```
app/
├── @modal/
│   ├── (.)photo/[id]/page.tsx   # Intercepted: shows as modal
│   └── default.tsx               # No modal by default
├── photo/[id]/page.tsx           # Direct URL: full page
├── layout.tsx
└── page.tsx                      # Feed with photo links
```

```tsx
// app/@modal/(.)photo/[id]/page.tsx
'use client';
import { useRouter } from 'next/navigation';

export default function PhotoModal({ params }: { params: Promise<{ id: string }> }) {
  const router = useRouter();
  const { id } = React.use(params);

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center"
         onClick={() => router.back()}>
      <div className="bg-white rounded-lg p-4" onClick={(e) => e.stopPropagation()}>
        <PhotoDetail id={id} />
      </div>
    </div>
  );
}
```

## Metadata

```tsx
// Static metadata
export const metadata = {
  title: 'Dashboard',
  description: 'Manage your projects',
  openGraph: {
    title: 'Dashboard',
    description: 'Manage your projects',
    images: ['/og-dashboard.png'],
  },
};

// Dynamic metadata
export async function generateMetadata({
  params,
}: {
  params: Promise<{ slug: string }>;
}) {
  const { slug } = await params;
  const post = await getPost(slug);

  return {
    title: post.title,
    description: post.excerpt,
    openGraph: {
      title: post.title,
      description: post.excerpt,
      images: [post.coverImage],
    },
  };
}
```

### Title Template

```tsx
// app/layout.tsx
export const metadata = {
  title: {
    default: 'My App',           // Fallback
    template: '%s | My App',     // Other pages: "Dashboard | My App"
  },
};

// app/dashboard/page.tsx
export const metadata = {
  title: 'Dashboard',  // Renders as "Dashboard | My App"
};
```

## Navigation

```tsx
import Link from 'next/link';
import { useRouter, usePathname, useSearchParams } from 'next/navigation';

// Declarative navigation
<Link href="/dashboard">Dashboard</Link>
<Link href={`/blog/${post.slug}`}>Read more</Link>

// Programmatic navigation (Client Components only)
const router = useRouter();
router.push('/dashboard');
router.replace('/login');
router.back();
router.refresh(); // Re-fetch server data without full reload

// Active link detection
const pathname = usePathname();
<Link
  href="/dashboard"
  className={pathname === '/dashboard' ? 'text-blue-500' : 'text-gray-500'}
>
  Dashboard
</Link>
```

## Common Gotchas

1. **params is a Promise in Next.js 15** — always `await params` in page/layout components. This changed from Next.js 14.

2. **`'use client'` doesn't mean "client-only"** — it means the component boundary where Client Component tree starts. It still server-renders on first load. It just also hydrates on the client.

3. **Can't import Server Components into Client Components** — instead, pass them as `children` props. The Client Component renders the Server Component as a "slot."

4. **Layouts don't re-render on navigation** — a parent layout won't re-fetch data when you navigate between child pages. Use `usePathname()` in a Client Component if you need to react to route changes.

5. **`redirect()` throws** — it uses a special error type internally. Don't catch it in a try/catch unless you re-throw.

6. **Static vs Dynamic rendering** — using `cookies()`, `headers()`, `searchParams`, or uncached `fetch()` opts the page into dynamic rendering. Be intentional about this.

7. **Route handlers (route.ts) vs Server Actions** — route.ts for REST APIs consumed by external clients. Server Actions for form submissions and mutations from your own UI.
