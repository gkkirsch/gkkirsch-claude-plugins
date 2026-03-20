---
name: supabase-auth
description: >
  Supabase authentication — email/password, OAuth providers, magic links,
  MFA, session management, auth hooks, and Next.js SSR integration.
  Triggers: "supabase auth", "supabase login", "supabase signup", "supabase oauth",
  "supabase magic link", "supabase mfa", "supabase session", "supabase middleware".
  NOT for: database queries (use supabase-database), storage (use supabase-storage-realtime).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Supabase Authentication

## Setup

```bash
npm install @supabase/supabase-js
# For Next.js SSR:
npm install @supabase/ssr
```

```typescript
// lib/supabase/client.ts — Browser client
import { createBrowserClient } from "@supabase/ssr";

export function createClient() {
  return createBrowserClient(
    process.env.NEXT_PUBLIC_SUPABASE_URL!,
    process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY!
  );
}
```

```typescript
// lib/supabase/server.ts — Server client (Next.js App Router)
import { createServerClient } from "@supabase/ssr";
import { cookies } from "next/headers";

export async function createClient() {
  const cookieStore = await cookies();

  return createServerClient(
    process.env.NEXT_PUBLIC_SUPABASE_URL!,
    process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY!,
    {
      cookies: {
        getAll() {
          return cookieStore.getAll();
        },
        setAll(cookiesToSet) {
          try {
            cookiesToSet.forEach(({ name, value, options }) =>
              cookieStore.set(name, value, options)
            );
          } catch {
            // Called from Server Component — ignore
          }
        },
      },
    }
  );
}
```

```typescript
// lib/supabase/middleware.ts — Auth middleware
import { createServerClient } from "@supabase/ssr";
import { NextResponse, type NextRequest } from "next/server";

export async function updateSession(request: NextRequest) {
  let supabaseResponse = NextResponse.next({ request });

  const supabase = createServerClient(
    process.env.NEXT_PUBLIC_SUPABASE_URL!,
    process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY!,
    {
      cookies: {
        getAll() {
          return request.cookies.getAll();
        },
        setAll(cookiesToSet) {
          cookiesToSet.forEach(({ name, value }) =>
            request.cookies.set(name, value)
          );
          supabaseResponse = NextResponse.next({ request });
          cookiesToSet.forEach(({ name, value, options }) =>
            supabaseResponse.cookies.set(name, value, options)
          );
        },
      },
    }
  );

  // Refresh session — IMPORTANT: don't remove this
  const { data: { user } } = await supabase.auth.getUser();

  // Protect routes
  if (!user && !request.nextUrl.pathname.startsWith("/auth")) {
    const url = request.nextUrl.clone();
    url.pathname = "/auth/login";
    return NextResponse.redirect(url);
  }

  return supabaseResponse;
}
```

```typescript
// middleware.ts
import { updateSession } from "@/lib/supabase/middleware";
import type { NextRequest } from "next/server";

export async function middleware(request: NextRequest) {
  return await updateSession(request);
}

export const config = {
  matcher: [
    // Skip static files and API routes
    "/((?!_next/static|_next/image|favicon.ico|api|auth).*)",
  ],
};
```

## Email/Password Auth

```typescript
// Sign up
const { data, error } = await supabase.auth.signUp({
  email: "user@example.com",
  password: "secure-password-123",
  options: {
    data: {
      full_name: "Jane Smith",
      avatar_url: "https://example.com/avatar.jpg",
    },
    emailRedirectTo: `${window.location.origin}/auth/callback`,
  },
});
// data.user exists but data.session is null until email confirmed

// Sign in
const { data, error } = await supabase.auth.signInWithPassword({
  email: "user@example.com",
  password: "secure-password-123",
});

// Sign out
await supabase.auth.signOut();

// Get current user (client-side)
const { data: { user } } = await supabase.auth.getUser();

// Get current session
const { data: { session } } = await supabase.auth.getSession();
```

## OAuth Providers

```typescript
// Sign in with OAuth (Google, GitHub, etc.)
const { data, error } = await supabase.auth.signInWithOAuth({
  provider: "google",
  options: {
    redirectTo: `${window.location.origin}/auth/callback`,
    queryParams: {
      access_type: "offline",
      prompt: "consent",
    },
    scopes: "email profile",
  },
});

// Auth callback route — handle the code exchange
// app/auth/callback/route.ts
import { createClient } from "@/lib/supabase/server";
import { NextResponse } from "next/server";

export async function GET(request: Request) {
  const { searchParams, origin } = new URL(request.url);
  const code = searchParams.get("code");
  const next = searchParams.get("next") ?? "/dashboard";

  if (code) {
    const supabase = await createClient();
    const { error } = await supabase.auth.exchangeCodeForSession(code);
    if (!error) {
      return NextResponse.redirect(`${origin}${next}`);
    }
  }

  return NextResponse.redirect(`${origin}/auth/error`);
}
```

## Magic Link (Passwordless)

```typescript
// Send magic link
const { data, error } = await supabase.auth.signInWithOtp({
  email: "user@example.com",
  options: {
    emailRedirectTo: `${window.location.origin}/auth/callback`,
    shouldCreateUser: true, // Create account if doesn't exist
  },
});

// Phone OTP
const { data, error } = await supabase.auth.signInWithOtp({
  phone: "+1234567890",
});

// Verify phone OTP
const { data, error } = await supabase.auth.verifyOtp({
  phone: "+1234567890",
  token: "123456",
  type: "sms",
});
```

## Password Reset

```typescript
// Request reset
const { error } = await supabase.auth.resetPasswordForEmail(
  "user@example.com",
  {
    redirectTo: `${window.location.origin}/auth/reset-password`,
  }
);

// Update password (on the reset page, after redirect)
const { error } = await supabase.auth.updateUser({
  password: "new-secure-password",
});
```

## Auth State Listener

```typescript
// React hook for auth state
import { useEffect, useState } from "react";
import type { User } from "@supabase/supabase-js";

export function useUser() {
  const supabase = createClient();
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // Get initial user
    supabase.auth.getUser().then(({ data: { user } }) => {
      setUser(user);
      setLoading(false);
    });

    // Listen for changes
    const { data: { subscription } } = supabase.auth.onAuthStateChange(
      (_event, session) => {
        setUser(session?.user ?? null);
      }
    );

    return () => subscription.unsubscribe();
  }, []);

  return { user, loading };
}
```

## User Profiles Pattern

```sql
-- Create profiles table linked to auth.users
create table public.profiles (
  id uuid references auth.users(id) on delete cascade primary key,
  full_name text,
  avatar_url text,
  bio text,
  created_at timestamptz default now(),
  updated_at timestamptz default now()
);

-- Enable RLS
alter table public.profiles enable row level security;

-- Policies
create policy "Public profiles are viewable by everyone"
  on profiles for select using (true);

create policy "Users can update own profile"
  on profiles for update using (auth.uid() = id);

-- Auto-create profile on signup
create or replace function public.handle_new_user()
returns trigger as $$
begin
  insert into public.profiles (id, full_name, avatar_url)
  values (
    new.id,
    new.raw_user_meta_data->>'full_name',
    new.raw_user_meta_data->>'avatar_url'
  );
  return new;
end;
$$ language plpgsql security definer;

create trigger on_auth_user_created
  after insert on auth.users
  for each row execute function public.handle_new_user();
```

## Server-Side Auth (Next.js App Router)

```typescript
// app/dashboard/page.tsx — Server Component
import { createClient } from "@/lib/supabase/server";
import { redirect } from "next/navigation";

export default async function DashboardPage() {
  const supabase = await createClient();
  const { data: { user } } = await supabase.auth.getUser();

  if (!user) {
    redirect("/auth/login");
  }

  const { data: profile } = await supabase
    .from("profiles")
    .select("*")
    .eq("id", user.id)
    .single();

  return <Dashboard user={user} profile={profile} />;
}
```

```typescript
// app/api/protected/route.ts — API Route
import { createClient } from "@/lib/supabase/server";
import { NextResponse } from "next/server";

export async function GET() {
  const supabase = await createClient();
  const { data: { user } } = await supabase.auth.getUser();

  if (!user) {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
  }

  const { data } = await supabase
    .from("items")
    .select("*")
    .eq("user_id", user.id);

  return NextResponse.json({ data });
}
```

## MFA (Multi-Factor Authentication)

```typescript
// Enroll MFA (TOTP)
const { data, error } = await supabase.auth.mfa.enroll({
  factorType: "totp",
  friendlyName: "My Authenticator",
});
// data.totp.qr_code — show QR code to user
// data.totp.uri — TOTP URI for manual entry
// data.id — factor ID (save this)

// Verify MFA challenge
const { data: challenge } = await supabase.auth.mfa.challenge({
  factorId: data.id,
});

const { data: verify, error } = await supabase.auth.mfa.verify({
  factorId: data.id,
  challengeId: challenge.id,
  code: "123456", // From authenticator app
});

// Check MFA status
const { data: { totp } } = await supabase.auth.mfa.getAuthenticatorAssuranceLevel();
// totp.currentLevel: "aal1" (password only) or "aal2" (MFA verified)

// Unenroll
await supabase.auth.mfa.unenroll({ factorId: "..." });
```

## Gotchas

1. **`getSession()` vs `getUser()`.** `getSession()` reads from local storage — it can be tampered with. `getUser()` makes a request to Supabase to verify the token. Always use `getUser()` on the server for security. `getSession()` is fine for client-side UI state.

2. **Cookies must be set in middleware.** The `@supabase/ssr` package manages auth tokens via cookies. The middleware must refresh the session on every request. Without it, sessions expire silently and server-side auth fails.

3. **Email confirmation is enabled by default.** New signups via email/password get a confirmation email. The session is null until they confirm. Disable in Supabase Dashboard > Auth > Settings if you don't want this.

4. **OAuth redirect URL must be whitelisted.** Add your callback URL (`http://localhost:3000/auth/callback` for dev, production URL for prod) in Dashboard > Auth > URL Configuration. Forgot URLs result in silent failures.

5. **Auth metadata is immutable after signup.** `raw_user_meta_data` set during `signUp` can only be updated via `updateUser()`. It's NOT the same as your profiles table. Use the profiles table for all mutable user data.

6. **Service role client bypasses RLS.** Creating a Supabase client with the service role key skips all RLS policies. Never use this on the client side. On the server, use it only when you intentionally need admin access (user management, background jobs).
