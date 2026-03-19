# Next.js App Router Reference

Quick-reference guide for Next.js 15+ App Router conventions, data fetching, caching, and deployment patterns. Consult this when building or reviewing Next.js applications.

---

## File Conventions

| File | Purpose | Server/Client |
|------|---------|---------------|
| `layout.tsx` | Shared UI wrapper, persists across navigations | Server (default) |
| `page.tsx` | Unique UI for a route, makes route accessible | Server (default) |
| `loading.tsx` | Loading UI shown while page loads (Suspense boundary) | Server |
| `error.tsx` | Error UI shown when page throws | **Client** (must be) |
| `not-found.tsx` | 404 UI for `notFound()` calls | Server |
| `template.tsx` | Like layout but remounts on navigation | Server |
| `default.tsx` | Fallback for parallel routes | Server |
| `route.ts` | API endpoint (Route Handler) | Server |
| `global-error.tsx` | Root-level error boundary | **Client** |
| `opengraph-image.tsx` | Dynamic OG image generation | Server (Edge) |
| `sitemap.ts` | Dynamic sitemap generation | Server |
| `robots.ts` | Dynamic robots.txt | Server |
| `manifest.ts` | Web app manifest | Server |
| `middleware.ts` | Request middleware (root only) | Edge |

---

## Route Segments

| Pattern | Example | Matches |
|---------|---------|---------|
| Static | `app/about/page.tsx` | `/about` |
| Dynamic | `app/blog/[slug]/page.tsx` | `/blog/hello-world` |
| Catch-all | `app/docs/[...slug]/page.tsx` | `/docs/a`, `/docs/a/b/c` |
| Optional catch-all | `app/docs/[[...slug]]/page.tsx` | `/docs`, `/docs/a`, `/docs/a/b` |
| Route group | `app/(marketing)/about/page.tsx` | `/about` (group ignored in URL) |
| Parallel | `app/@modal/login/page.tsx` | Renders alongside sibling slots |
| Intercept same level | `app/(.)photo/[id]/page.tsx` | Intercepts `/photo/:id` |
| Intercept one up | `app/(..)photo/[id]/page.tsx` | Intercepts from parent level |
| Intercept root | `app/(...)photo/[id]/page.tsx` | Intercepts from root |

---

## Server vs Client Components

### Decision Guide

| Need | Component Type |
|------|---------------|
| Fetch data | **Server** |
| Access backend resources | **Server** |
| Keep secrets (API keys, tokens) | **Server** |
| Reduce client JS | **Server** |
| onClick, onChange, onSubmit | **Client** |
| useState, useEffect, useRef | **Client** |
| Browser APIs (localStorage, geolocation) | **Client** |
| Custom hooks with state | **Client** |
| React Context (read) | Both (use() in Server, useContext in Client) |

### Composition Rules

```
Server Component CAN import:
  ✅ Other Server Components
  ✅ Client Components (as children/props)
  ✅ Server-only packages

Client Component CAN import:
  ✅ Other Client Components
  ❌ Server Components directly
  ✅ Server Components passed as children/props (React serializes them)

Passing data across the boundary:
  ✅ Serializable props (string, number, boolean, Date, plain objects, arrays)
  ❌ Functions, classes, Symbols, Buffers, Streams
```

### Pattern: Server Component Passes Data to Client

```tsx
// page.tsx (Server)
import { ClientWidget } from './client-widget';
const data = await fetchData(); // Server-only
return <ClientWidget initialData={data} />;

// client-widget.tsx (Client)
'use client';
export function ClientWidget({ initialData }: { initialData: Data }) {
  const [data, setData] = useState(initialData);
  // Interactive from here
}
```

---

## Data Fetching Strategies

### fetch Caching

```tsx
// Static (cached indefinitely — default in production)
fetch(url)
fetch(url, { cache: 'force-cache' })

// Revalidated (cached but refreshed on schedule)
fetch(url, { next: { revalidate: 3600 } })  // Every hour

// Dynamic (never cached)
fetch(url, { cache: 'no-store' })

// Tagged (revalidated on demand)
fetch(url, { next: { tags: ['posts'] } })
// Trigger: revalidateTag('posts')
```

### Caching Layers

```
Request → Fetch Cache → Data Cache → Full Route Cache → Client Router Cache

1. React cache()          — Deduplicates within a single request
2. fetch cache            — Caches fetch responses (configurable)
3. unstable_cache()       — Caches non-fetch data (DB queries)
4. Full Route Cache       — Caches entire HTML + RSC payload at build
5. Router Cache (client)  — Caches visited routes in browser (30s dynamic, 5min static)
```

### Revalidation

```tsx
// Time-based (ISR)
export const revalidate = 3600; // Page-level: revalidate every hour

// On-demand (from Server Action or Route Handler)
import { revalidatePath, revalidateTag } from 'next/cache';

revalidatePath('/posts');           // Revalidate specific path
revalidatePath('/posts', 'page');   // Only the page, not layouts
revalidatePath('/posts', 'layout'); // Page + all layouts
revalidateTag('posts');             // All fetches tagged 'posts'
```

---

## Segment Config Options

```tsx
// Per-page or per-layout configuration
export const dynamic = 'auto' | 'force-dynamic' | 'error' | 'force-static';
export const revalidate = false | 0 | number;
export const fetchCache = 'auto' | 'default-cache' | 'only-cache' | 'force-cache' | 'default-no-store' | 'only-no-store' | 'force-no-store';
export const runtime = 'nodejs' | 'edge';
export const preferredRegion = 'auto' | 'global' | 'home' | string | string[];
export const maxDuration = number; // Function timeout in seconds
```

### Dynamic Functions (Force Dynamic Rendering)

These functions opt the route into dynamic rendering:
- `cookies()`
- `headers()`
- `searchParams` prop
- `connection()`
- `unstable_noStore()`

---

## Metadata API

```tsx
// Static metadata
export const metadata: Metadata = {
  title: 'My Page',
  description: 'Page description',
};

// Dynamic metadata
export async function generateMetadata({ params, searchParams }: Props): Promise<Metadata> {
  const { slug } = await params;
  const product = await getProduct(slug);

  return {
    title: product.name,
    description: product.description,
    openGraph: {
      images: [product.imageUrl],
    },
  };
}

// Template metadata (in layout)
export const metadata: Metadata = {
  title: {
    template: '%s | MyApp',  // Children's titles get wrapped
    default: 'MyApp',
  },
};

// Metadata files (alternative to config)
// app/opengraph-image.tsx    → Dynamic OG image
// app/icon.tsx               → Dynamic favicon
// app/sitemap.ts             → Dynamic sitemap
// app/robots.ts              → Dynamic robots.txt
```

---

## Server Actions Quick Reference

```tsx
// Define (in separate file or inline)
'use server';

export async function myAction(formData: FormData) {
  // Server-only code: DB access, auth checks, etc.
  revalidatePath('/');
  redirect('/success');
}

// Use in Server Component form
<form action={myAction}>
  <input name="title" />
  <button type="submit">Submit</button>
</form>

// Use in Client Component
'use client';
const [state, formAction, isPending] = useActionState(myAction, initialState);
<form action={formAction}>

// Use with useFormStatus (inside form)
'use client';
function SubmitButton() {
  const { pending } = useFormStatus();
  return <button disabled={pending}>{pending ? 'Saving...' : 'Save'}</button>;
}

// Call directly (not from form)
async function handleClick() {
  await myAction(data);
}
```

### Server Action Patterns

```tsx
// Return validation errors
export async function createItem(prevState: State, formData: FormData) {
  const result = schema.safeParse(Object.fromEntries(formData));
  if (!result.success) return { errors: result.error.flatten().fieldErrors };
  await db.item.create({ data: result.data });
  revalidatePath('/items');
  return { errors: {} };
}

// Redirect after mutation
export async function deleteItem(id: string) {
  await db.item.delete({ where: { id } });
  revalidatePath('/items');
  redirect('/items');
}

// With auth check
export async function updateProfile(formData: FormData) {
  const session = await getSession();
  if (!session) throw new Error('Unauthorized');
  // proceed...
}
```

---

## Middleware Reference

```tsx
// middleware.ts (root of project)
import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

export function middleware(request: NextRequest) {
  // Available:
  request.nextUrl          // URL object with pathname, searchParams
  request.cookies          // Read/set cookies
  request.headers          // Request headers
  request.geo              // Geolocation (Vercel)
  request.ip               // Client IP (Vercel)

  // Actions:
  NextResponse.next()                        // Continue
  NextResponse.redirect(new URL('/login'))   // Redirect
  NextResponse.rewrite(new URL('/proxy'))    // Rewrite (URL stays same)
  NextResponse.json({ error: 'msg' })       // JSON response

  // Modify headers
  const response = NextResponse.next();
  response.headers.set('x-custom', 'value');
  return response;
}

// Matcher — which routes middleware runs on
export const config = {
  matcher: [
    // Match all except static files
    '/((?!_next/static|_next/image|favicon.ico).*)',
    // Or specific paths
    '/dashboard/:path*',
    '/api/:path*',
  ],
};
```

---

## Route Handlers (API)

```tsx
// app/api/items/route.ts
import { NextRequest, NextResponse } from 'next/server';

// Supported methods: GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS
export async function GET(request: NextRequest) {
  const searchParams = request.nextUrl.searchParams;
  const page = searchParams.get('page') ?? '1';
  return NextResponse.json({ data: items });
}

export async function POST(request: NextRequest) {
  const body = await request.json();
  return NextResponse.json(created, { status: 201 });
}

// Dynamic route: app/api/items/[id]/route.ts
export async function GET(
  request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const { id } = await params;
  return NextResponse.json(item);
}

// Streaming response
export async function GET() {
  const stream = new ReadableStream({
    async start(controller) {
      controller.enqueue(new TextEncoder().encode('data: hello\n\n'));
      controller.close();
    },
  });
  return new Response(stream, {
    headers: { 'Content-Type': 'text/event-stream' },
  });
}

// Route Handler caching
export const dynamic = 'force-static'; // Cache GET responses
export const revalidate = 3600;        // Revalidate every hour
```

---

## Parallel Routes

```
app/dashboard/
├── layout.tsx            # Receives {children, analytics, notifications}
├── page.tsx              # children slot
├── @analytics/
│   ├── page.tsx          # analytics slot for /dashboard
│   ├── loading.tsx       # Independent loading state
│   └── default.tsx       # Fallback for non-matching sub-routes
└── @notifications/
    ├── page.tsx          # notifications slot for /dashboard
    └── default.tsx
```

```tsx
// layout.tsx
export default function Layout({
  children,
  analytics,
  notifications,
}: {
  children: React.ReactNode;
  analytics: React.ReactNode;
  notifications: React.ReactNode;
}) {
  return (
    <div className="grid grid-cols-12">
      <main className="col-span-8">{children}</main>
      <aside className="col-span-4">
        {analytics}
        {notifications}
      </aside>
    </div>
  );
}
```

**Key behaviors:**
- Each slot loads independently (parallel data fetching)
- Each slot can have its own `loading.tsx` and `error.tsx`
- `default.tsx` is required as fallback for unmatched sub-routes
- Slots are NOT URL segments — they don't affect the URL

---

## Intercepting Routes

```
Prefix  Meaning              Example
(.)     Same level           app/feed/(.)photo/[id]
(..)    One level up         app/feed/(..)photo/[id]
(..)(..) Two levels up       app/feed/(..)(..)/photo/[id]
(...)   From root            app/feed/(...)/photo/[id]
```

**Use case:** Show modal on soft navigation, full page on hard navigation (refresh/direct URL).

---

## Image Optimization

```tsx
import Image from 'next/image';

// Local image (auto width/height, blur placeholder)
import photo from './photo.jpg';
<Image src={photo} alt="Photo" placeholder="blur" />

// Remote image (must specify dimensions)
<Image src="https://..." alt="..." width={800} height={600} />

// Priority (preload for LCP images)
<Image src={hero} alt="Hero" priority />

// Responsive sizes
<Image
  src={photo}
  alt="Photo"
  sizes="(max-width: 768px) 100vw, (max-width: 1200px) 50vw, 33vw"
/>

// Fill mode (fills parent container)
<div className="relative h-64 w-full">
  <Image src={photo} alt="Photo" fill className="object-cover" />
</div>
```

**next.config.ts:**
```tsx
const config = {
  images: {
    remotePatterns: [
      { protocol: 'https', hostname: '**.example.com' },
    ],
    formats: ['image/avif', 'image/webp'],
  },
};
```

---

## Environment Variables

```
# .env.local (git-ignored, local development)
DATABASE_URL=postgres://...
SECRET_KEY=...

# .env (committed, defaults)
NEXT_PUBLIC_APP_URL=https://myapp.com

# Access:
# Server-only: process.env.DATABASE_URL
# Client+Server: process.env.NEXT_PUBLIC_APP_URL (must have NEXT_PUBLIC_ prefix)
```

---

## next.config.ts Common Options

```tsx
import type { NextConfig } from 'next';

const config: NextConfig = {
  // Redirects
  async redirects() {
    return [
      { source: '/old-page', destination: '/new-page', permanent: true },
    ];
  },

  // Rewrites
  async rewrites() {
    return [
      { source: '/api/:path*', destination: 'https://api.example.com/:path*' },
    ];
  },

  // Headers
  async headers() {
    return [
      {
        source: '/(.*)',
        headers: [
          { key: 'X-Frame-Options', value: 'DENY' },
        ],
      },
    ];
  },

  // Performance
  experimental: {
    optimizePackageImports: ['lucide-react', '@/components/ui'],
    ppr: true, // Partial Prerendering
  },

  // Images
  images: {
    remotePatterns: [{ hostname: '**.cloudinary.com' }],
  },
};

export default config;
```

---

## Deployment Checklist

- [ ] `next build` succeeds without errors
- [ ] Environment variables set in production
- [ ] Images use `next/image` with proper `sizes`
- [ ] Fonts use `next/font` (self-hosted, no layout shift)
- [ ] Metadata configured for all pages
- [ ] Sitemap generated (`app/sitemap.ts`)
- [ ] Error boundaries at page and layout level
- [ ] Loading states for all dynamic routes
- [ ] Middleware configured for auth/redirects
- [ ] Cache strategies reviewed (static vs dynamic vs ISR)
- [ ] Bundle analyzed for heavy dependencies
- [ ] Core Web Vitals passing (LCP < 2.5s, INP < 200ms, CLS < 0.1)
