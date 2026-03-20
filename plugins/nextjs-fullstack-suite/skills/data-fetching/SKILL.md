---
name: nextjs-data-fetching
description: >
  Next.js data fetching patterns — Server Component fetching, caching, revalidation,
  parallel fetching, API routes, middleware, and database access patterns.
  Triggers: "next.js data fetching", "server component fetch", "api route",
  "next.js caching", "revalidate", "ISR", "SSG", "SSR", "route handler".
  NOT for: form mutations (use server-actions skill), routing (use app-router skill).
version: 1.0.0
allowed-tools: Read, Grep, Glob, Bash, Edit, Write
---

# Next.js Data Fetching

## The Golden Rule

**Fetch data in Server Components.** Not in Client Components, not in useEffect, not in client-side libraries. Server Components can directly access databases, APIs, and file systems without sending JavaScript to the client.

## Server Component Data Fetching

### Basic Pattern

```tsx
// app/posts/page.tsx (Server Component by default)
import { db } from '@/lib/db';

export default async function PostsPage() {
  const posts = await db.post.findMany({
    orderBy: { createdAt: 'desc' },
    take: 20,
  });

  return (
    <div>
      <h1>Posts</h1>
      {posts.map((post) => (
        <PostCard key={post.id} post={post} />
      ))}
    </div>
  );
}
```

No `useState`, no `useEffect`, no loading state management. The component is async, fetches data, and renders.

### Parallel Data Fetching

```tsx
// BAD: Sequential (slow) — each await blocks the next
export default async function Dashboard() {
  const user = await getUser();         // 200ms
  const posts = await getPosts();       // 300ms
  const analytics = await getAnalytics(); // 400ms
  // Total: 900ms

  return <div>...</div>;
}

// GOOD: Parallel — all fetch at the same time
export default async function Dashboard() {
  const [user, posts, analytics] = await Promise.all([
    getUser(),        // 200ms
    getPosts(),       // 300ms
    getAnalytics(),   // 400ms
  ]);
  // Total: 400ms (slowest one)

  return <div>...</div>;
}

// BEST: Streaming with Suspense — each section loads independently
export default function Dashboard() {
  return (
    <div>
      <Suspense fallback={<UserSkeleton />}>
        <UserInfo />       {/* Streams at 200ms */}
      </Suspense>
      <Suspense fallback={<PostsSkeleton />}>
        <RecentPosts />    {/* Streams at 300ms */}
      </Suspense>
      <Suspense fallback={<ChartSkeleton />}>
        <Analytics />      {/* Streams at 400ms */}
      </Suspense>
    </div>
  );
}
```

### Request Deduplication

Next.js automatically deduplicates identical `fetch()` calls within a single render:

```tsx
// Both components call the same URL — only one request is made
async function PostTitle({ id }: { id: string }) {
  const post = await fetch(`${API}/posts/${id}`).then(r => r.json());
  return <h1>{post.title}</h1>;
}

async function PostBody({ id }: { id: string }) {
  const post = await fetch(`${API}/posts/${id}`).then(r => r.json()); // Deduped!
  return <div>{post.content}</div>;
}
```

For non-fetch data sources (Prisma, etc.), use React `cache()`:

```typescript
// lib/data.ts
import { cache } from 'react';
import { db } from '@/lib/db';

export const getUser = cache(async (id: string) => {
  return db.user.findUnique({ where: { id } });
});

// Called multiple times in one render → only one DB query
```

## Caching Strategies

### Static Data (Cached Forever)

```typescript
// Default behavior for fetch() in Server Components
const data = await fetch('https://api.example.com/posts');
// Cached at build time, never refetched (unless revalidated)
```

### Time-Based Revalidation (ISR)

```typescript
// Refetch every 60 seconds
const data = await fetch('https://api.example.com/posts', {
  next: { revalidate: 60 },
});

// Or at the page level
export const revalidate = 60; // seconds
```

### On-Demand Revalidation

```typescript
// Revalidate when data changes (in a Server Action or Route Handler)
import { revalidatePath, revalidateTag } from 'next/cache';

// By path
revalidatePath('/posts');

// By tag (more precise)
revalidateTag('posts');

// Tag your fetches
const data = await fetch('https://api.example.com/posts', {
  next: { tags: ['posts'] },
});
```

### No Cache (Dynamic)

```typescript
// Never cache — always fetch fresh
const data = await fetch('https://api.example.com/posts', {
  cache: 'no-store',
});

// Or at the page level
export const dynamic = 'force-dynamic';
```

### For Non-Fetch Data Sources

```typescript
import { unstable_cache } from 'next/cache';

const getCachedPosts = unstable_cache(
  async () => {
    return db.post.findMany({ orderBy: { createdAt: 'desc' } });
  },
  ['posts'],               // Cache key
  {
    revalidate: 60,        // Seconds
    tags: ['posts'],       // For on-demand revalidation
  }
);

// Use in Server Component
const posts = await getCachedPosts();
```

## API Routes (Route Handlers)

```typescript
// app/api/posts/route.ts
import { NextRequest, NextResponse } from 'next/server';
import { db } from '@/lib/db';

// GET /api/posts
export async function GET(request: NextRequest) {
  const searchParams = request.nextUrl.searchParams;
  const page = parseInt(searchParams.get('page') ?? '1');
  const limit = parseInt(searchParams.get('limit') ?? '20');

  const posts = await db.post.findMany({
    skip: (page - 1) * limit,
    take: limit,
    orderBy: { createdAt: 'desc' },
  });

  const total = await db.post.count();

  return NextResponse.json({
    data: posts,
    pagination: { page, limit, total, pages: Math.ceil(total / limit) },
  });
}

// POST /api/posts
export async function POST(request: NextRequest) {
  try {
    const body = await request.json();

    const post = await db.post.create({
      data: {
        title: body.title,
        content: body.content,
      },
    });

    return NextResponse.json(post, { status: 201 });
  } catch (err) {
    return NextResponse.json(
      { error: 'Invalid request' },
      { status: 400 }
    );
  }
}
```

### Dynamic Route Handler

```typescript
// app/api/posts/[id]/route.ts
export async function GET(
  request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const { id } = await params;
  const post = await db.post.findUnique({ where: { id } });

  if (!post) {
    return NextResponse.json({ error: 'Not found' }, { status: 404 });
  }

  return NextResponse.json(post);
}

export async function PATCH(
  request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const { id } = await params;
  const body = await request.json();

  const post = await db.post.update({
    where: { id },
    data: body,
  });

  return NextResponse.json(post);
}

export async function DELETE(
  request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const { id } = await params;
  await db.post.delete({ where: { id } });
  return new NextResponse(null, { status: 204 });
}
```

### Webhook Route Handler

```typescript
// app/api/webhooks/stripe/route.ts
import { NextRequest, NextResponse } from 'next/server';
import Stripe from 'stripe';

const stripe = new Stripe(process.env.STRIPE_SECRET_KEY!);

export async function POST(request: NextRequest) {
  const body = await request.text(); // Raw body for signature
  const sig = request.headers.get('stripe-signature')!;

  try {
    const event = stripe.webhooks.constructEvent(
      body,
      sig,
      process.env.STRIPE_WEBHOOK_SECRET!
    );

    // Handle event...
    return NextResponse.json({ received: true });
  } catch (err: any) {
    return NextResponse.json(
      { error: err.message },
      { status: 400 }
    );
  }
}
```

## Middleware

```typescript
// middleware.ts (project root)
import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

export function middleware(request: NextRequest) {
  // Auth guard
  const token = request.cookies.get('session')?.value;

  if (request.nextUrl.pathname.startsWith('/dashboard') && !token) {
    const loginUrl = new URL('/login', request.url);
    loginUrl.searchParams.set('from', request.nextUrl.pathname);
    return NextResponse.redirect(loginUrl);
  }

  // CORS for API routes
  if (request.nextUrl.pathname.startsWith('/api/')) {
    const response = NextResponse.next();
    response.headers.set('Access-Control-Allow-Origin', '*');
    response.headers.set('Access-Control-Allow-Methods', 'GET, POST, PUT, DELETE, OPTIONS');
    response.headers.set('Access-Control-Allow-Headers', 'Content-Type, Authorization');
    return response;
  }

  // Geolocation-based redirect
  const country = request.geo?.country;
  if (country === 'DE' && !request.nextUrl.pathname.startsWith('/de')) {
    return NextResponse.redirect(new URL('/de' + request.nextUrl.pathname, request.url));
  }

  return NextResponse.next();
}

export const config = {
  matcher: [
    '/dashboard/:path*',
    '/api/:path*',
    '/((?!_next/static|_next/image|favicon.ico).*)',
  ],
};
```

## Search Params (URL State)

```tsx
// app/posts/page.tsx — Server Component
export default async function PostsPage({
  searchParams,
}: {
  searchParams: Promise<{ q?: string; page?: string; sort?: string }>;
}) {
  const { q, page = '1', sort = 'newest' } = await searchParams;

  const posts = await db.post.findMany({
    where: q ? { title: { contains: q, mode: 'insensitive' } } : undefined,
    orderBy: sort === 'newest' ? { createdAt: 'desc' } : { title: 'asc' },
    skip: (parseInt(page) - 1) * 20,
    take: 20,
  });

  return (
    <div>
      <SearchForm defaultValue={q} />
      <PostList posts={posts} />
      <Pagination currentPage={parseInt(page)} />
    </div>
  );
}
```

```tsx
// components/search-form.tsx — Client Component for interactivity
'use client';

import { useRouter, useSearchParams } from 'next/navigation';
import { useDebouncedCallback } from 'use-debounce';

export function SearchForm({ defaultValue }: { defaultValue?: string }) {
  const router = useRouter();
  const searchParams = useSearchParams();

  const handleSearch = useDebouncedCallback((term: string) => {
    const params = new URLSearchParams(searchParams);
    if (term) {
      params.set('q', term);
    } else {
      params.delete('q');
    }
    params.set('page', '1'); // Reset to page 1 on new search
    router.push(`/posts?${params.toString()}`);
  }, 300);

  return (
    <input
      type="search"
      placeholder="Search posts..."
      defaultValue={defaultValue}
      onChange={(e) => handleSearch(e.target.value)}
    />
  );
}
```

## Static Generation

```tsx
// app/blog/[slug]/page.tsx

// Generate static pages at build time
export async function generateStaticParams() {
  const posts = await db.post.findMany({
    select: { slug: true },
  });

  return posts.map((post) => ({
    slug: post.slug,
  }));
}

// dynamicParams controls behavior for unknown slugs
export const dynamicParams = true; // (default) Generate on first request, cache
// export const dynamicParams = false; // 404 for unknown slugs

export default async function BlogPost({
  params,
}: {
  params: Promise<{ slug: string }>;
}) {
  const { slug } = await params;
  const post = await db.post.findUnique({ where: { slug } });
  if (!post) notFound();

  return <article>{post.content}</article>;
}
```

## Common Gotchas

1. **Don't fetch in Client Components with useEffect** — fetch in the Server Component and pass data as props. Only use client-side fetching for real-time data (SWR/React Query for polling, WebSockets).

2. **fetch() in Server Components is extended** — Next.js adds caching and revalidation options to the native fetch. This is NOT the same as browser fetch.

3. **searchParams opts into dynamic rendering** — any page that reads `searchParams` becomes dynamic (not statically generated). This is correct for search pages but be aware of it.

4. **unstable_cache is the non-fetch equivalent** — for Prisma, direct DB calls, or any non-fetch data source that needs caching. Despite the name, it's production-ready.

5. **Route Handlers are cached by default for GET** — `GET` handlers with no dynamic functions (cookies, headers) are statically cached. Add `export const dynamic = 'force-dynamic'` to prevent this.

6. **Middleware runs on every request** — keep it lightweight. No database calls, no heavy computation. Use it for redirects, headers, and cookie checks only.

7. **params is a Promise in Next.js 15** — `await params` in all page, layout, and route handler functions. This changed from synchronous in Next.js 14.

8. **cookies() and headers() make the page dynamic** — calling these in a Server Component or layout opts the entire page into dynamic rendering. Use them intentionally.
