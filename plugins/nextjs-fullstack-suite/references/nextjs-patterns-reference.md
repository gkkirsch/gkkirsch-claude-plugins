# Next.js Common Patterns Reference

Practical patterns for authentication, database access, error handling, and testing in Next.js applications.

---

## Authentication with NextAuth.js (Auth.js)

### Setup

```bash
npm install next-auth@beta
npx auth secret  # Generate AUTH_SECRET
```

```typescript
// auth.ts
import NextAuth from 'next-auth';
import GitHub from 'next-auth/providers/github';
import Google from 'next-auth/providers/google';
import Credentials from 'next-auth/providers/credentials';
import { PrismaAdapter } from '@auth/prisma-adapter';
import { db } from '@/lib/db';
import bcrypt from 'bcryptjs';

export const { handlers, signIn, signOut, auth } = NextAuth({
  adapter: PrismaAdapter(db),
  providers: [
    GitHub,
    Google,
    Credentials({
      credentials: {
        email: { label: 'Email', type: 'email' },
        password: { label: 'Password', type: 'password' },
      },
      async authorize(credentials) {
        const user = await db.user.findUnique({
          where: { email: credentials.email as string },
        });
        if (!user?.passwordHash) return null;

        const valid = await bcrypt.compare(
          credentials.password as string,
          user.passwordHash
        );
        if (!valid) return null;

        return { id: user.id, email: user.email, name: user.name };
      },
    }),
  ],
  callbacks: {
    session({ session, token }) {
      if (token.sub) session.user.id = token.sub;
      return session;
    },
  },
});
```

### Route Handler

```typescript
// app/api/auth/[...nextauth]/route.ts
import { handlers } from '@/auth';
export const { GET, POST } = handlers;
```

### Protect Server Components

```tsx
import { auth } from '@/auth';
import { redirect } from 'next/navigation';

export default async function DashboardPage() {
  const session = await auth();
  if (!session) redirect('/login');

  return <div>Welcome, {session.user.name}</div>;
}
```

### Protect Server Actions

```typescript
'use server';
import { auth } from '@/auth';

export async function createPost(formData: FormData) {
  const session = await auth();
  if (!session) throw new Error('Unauthorized');

  // ... create post with session.user.id
}
```

### Sign In / Sign Out Components

```tsx
import { signIn, signOut } from '@/auth';

export function SignInButton() {
  return (
    <form action={async () => {
      'use server';
      await signIn('github');
    }}>
      <button>Sign in with GitHub</button>
    </form>
  );
}

export function SignOutButton() {
  return (
    <form action={async () => {
      'use server';
      await signOut();
    }}>
      <button>Sign out</button>
    </form>
  );
}
```

## Database Access Patterns

### Prisma Setup

```bash
npm install prisma @prisma/client
npx prisma init
```

```typescript
// lib/db.ts
import { PrismaClient } from '@prisma/client';

const globalForPrisma = globalThis as unknown as { prisma: PrismaClient };

export const db = globalForPrisma.prisma || new PrismaClient();

if (process.env.NODE_ENV !== 'production') globalForPrisma.prisma = db;
```

The `globalForPrisma` pattern prevents creating new PrismaClient instances on every hot reload in development.

### Drizzle Alternative

```typescript
// lib/db.ts
import { drizzle } from 'drizzle-orm/postgres-js';
import postgres from 'postgres';
import * as schema from './schema';

const client = postgres(process.env.DATABASE_URL!);
export const db = drizzle(client, { schema });
```

## Error Handling Patterns

### Try/Catch in Server Actions

```typescript
'use server';

type ActionResult<T = void> =
  | { success: true; data: T }
  | { success: false; error: string };

export async function createPost(formData: FormData): Promise<ActionResult<{ id: string }>> {
  try {
    // Validate, authenticate, create...
    return { success: true, data: { id: post.id } };
  } catch (err) {
    if (err instanceof Prisma.PrismaClientKnownRequestError) {
      if (err.code === 'P2002') {
        return { success: false, error: 'A post with this slug already exists' };
      }
    }
    console.error('Unexpected error:', err);
    return { success: false, error: 'Something went wrong' };
  }
}
```

### Global Error Logging

```typescript
// instrumentation.ts
export async function register() {
  if (process.env.NEXT_RUNTIME === 'nodejs') {
    // Initialize error tracking (Sentry, LogRocket, etc.)
    const Sentry = await import('@sentry/nextjs');
    Sentry.init({
      dsn: process.env.SENTRY_DSN,
      tracesSampleRate: 0.1,
    });
  }
}
```

## Testing Patterns

### Unit Testing Components

```typescript
// __tests__/components/post-card.test.tsx
import { render, screen } from '@testing-library/react';
import { PostCard } from '@/components/post-card';

describe('PostCard', () => {
  it('renders post title and excerpt', () => {
    render(
      <PostCard
        post={{
          id: '1',
          title: 'Test Post',
          excerpt: 'This is a test',
          slug: 'test-post',
        }}
      />
    );

    expect(screen.getByText('Test Post')).toBeInTheDocument();
    expect(screen.getByText('This is a test')).toBeInTheDocument();
    expect(screen.getByRole('link')).toHaveAttribute('href', '/posts/test-post');
  });
});
```

### Testing Server Actions

```typescript
// __tests__/actions/posts.test.ts
import { createPost } from '@/actions/posts';
import { db } from '@/lib/db';

// Mock auth
jest.mock('@/auth', () => ({
  auth: jest.fn(() => ({
    user: { id: 'user-1', email: 'test@example.com' },
  })),
}));

describe('createPost', () => {
  it('creates a post with valid data', async () => {
    const formData = new FormData();
    formData.set('title', 'Test Post');
    formData.set('content', 'Test content');

    const result = await createPost(formData);
    expect(result.success).toBe(true);

    const post = await db.post.findFirst({
      where: { title: 'Test Post' },
    });
    expect(post).toBeTruthy();
  });

  it('returns error for invalid data', async () => {
    const formData = new FormData();
    // Missing required title

    const result = await createPost(formData);
    expect(result.success).toBe(false);
  });
});
```

### E2E Testing with Playwright

```typescript
// e2e/posts.spec.ts
import { test, expect } from '@playwright/test';

test('create and view a post', async ({ page }) => {
  // Login
  await page.goto('/login');
  await page.fill('input[name="email"]', 'test@example.com');
  await page.fill('input[name="password"]', 'password');
  await page.click('button[type="submit"]');

  // Create post
  await page.goto('/posts/new');
  await page.fill('input[name="title"]', 'E2E Test Post');
  await page.fill('textarea[name="content"]', 'Created by Playwright');
  await page.click('button[type="submit"]');

  // Verify redirect to new post
  await expect(page).toHaveURL(/\/posts\//);
  await expect(page.locator('h1')).toHaveText('E2E Test Post');
});
```

## Internationalization (i18n)

### Middleware-Based i18n (App Router)

```typescript
// middleware.ts
import { match } from '@formatjs/intl-localematcher';
import Negotiator from 'negotiator';

const locales = ['en', 'es', 'fr', 'de'];
const defaultLocale = 'en';

function getLocale(request: NextRequest): string {
  const negotiator = new Negotiator({
    headers: { 'accept-language': request.headers.get('accept-language') || '' },
  });
  return match(negotiator.languages(), locales, defaultLocale);
}

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;

  // Check if locale is in pathname
  const hasLocale = locales.some(
    (locale) => pathname.startsWith(`/${locale}/`) || pathname === `/${locale}`
  );

  if (hasLocale) return;

  // Redirect to locale-prefixed path
  const locale = getLocale(request);
  request.nextUrl.pathname = `/${locale}${pathname}`;
  return NextResponse.redirect(request.nextUrl);
}
```

### Route Structure

```
app/
├── [locale]/
│   ├── layout.tsx
│   ├── page.tsx
│   └── about/page.tsx
└── dictionaries/
    ├── en.json
    ├── es.json
    └── fr.json
```

```typescript
// lib/dictionaries.ts
const dictionaries = {
  en: () => import('@/dictionaries/en.json').then((m) => m.default),
  es: () => import('@/dictionaries/es.json').then((m) => m.default),
};

export async function getDictionary(locale: string) {
  return dictionaries[locale as keyof typeof dictionaries]();
}

// app/[locale]/page.tsx
export default async function Home({ params }: { params: Promise<{ locale: string }> }) {
  const { locale } = await params;
  const dict = await getDictionary(locale);

  return <h1>{dict.home.title}</h1>;
}
```

## Provider Pattern

```tsx
// components/providers.tsx
'use client';

import { ThemeProvider } from 'next-themes';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { SessionProvider } from 'next-auth/react';
import { Toaster } from 'sonner';

const queryClient = new QueryClient();

export function Providers({ children }: { children: React.ReactNode }) {
  return (
    <SessionProvider>
      <QueryClientProvider client={queryClient}>
        <ThemeProvider attribute="class" defaultTheme="system" enableSystem>
          {children}
          <Toaster richColors />
        </ThemeProvider>
      </QueryClientProvider>
    </SessionProvider>
  );
}

// app/layout.tsx
import { Providers } from '@/components/providers';

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body>
        <Providers>{children}</Providers>
      </body>
    </html>
  );
}
```

## Rate Limiting API Routes

```typescript
// lib/rate-limit.ts
const rateLimit = new Map<string, { count: number; timestamp: number }>();

export function rateLimiter(
  identifier: string,
  limit: number = 10,
  windowMs: number = 60000
): { success: boolean; remaining: number } {
  const now = Date.now();
  const record = rateLimit.get(identifier);

  if (!record || now - record.timestamp > windowMs) {
    rateLimit.set(identifier, { count: 1, timestamp: now });
    return { success: true, remaining: limit - 1 };
  }

  if (record.count >= limit) {
    return { success: false, remaining: 0 };
  }

  record.count++;
  return { success: true, remaining: limit - record.count };
}

// Usage in route handler
export async function POST(request: NextRequest) {
  const ip = request.headers.get('x-forwarded-for') || 'unknown';
  const { success, remaining } = rateLimiter(ip, 10, 60000);

  if (!success) {
    return NextResponse.json(
      { error: 'Rate limit exceeded' },
      { status: 429, headers: { 'X-RateLimit-Remaining': '0' } }
    );
  }

  // ... handle request
}
```
