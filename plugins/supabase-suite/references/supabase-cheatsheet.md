# Supabase Cheatsheet

## Setup

```bash
npm install @supabase/supabase-js @supabase/ssr
npx supabase init                    # Local dev
npx supabase gen types typescript    # Generate TypeScript types
```

## Client Creation

```typescript
// Browser
import { createBrowserClient } from "@supabase/ssr";
const supabase = createBrowserClient<Database>(URL, ANON_KEY);

// Server (Next.js App Router)
import { createServerClient } from "@supabase/ssr";
import { cookies } from "next/headers";
const supabase = createServerClient<Database>(URL, ANON_KEY, { cookies: ... });

// Admin (bypasses RLS)
import { createClient } from "@supabase/supabase-js";
const admin = createClient(URL, SERVICE_ROLE_KEY);
```

## Queries

```typescript
// SELECT
.from("posts").select("*")
.from("posts").select("id, title, author:profiles(name)")
.from("posts").select("*", { count: "exact" })
.from("posts").select("*").eq("id", id).single()

// Filters
.eq("col", val)  .neq()  .gt()  .gte()  .lt()  .lte()
.like("col", "%pat%")  .ilike()  .in("col", [a,b])
.is("col", null)  .contains()  .textSearch()
.or("a.eq.1,b.eq.2")

// Order & paginate
.order("created_at", { ascending: false })
.range(0, 9)  .limit(10)
```

## Mutations

```typescript
// INSERT
.from("posts").insert({ title, body }).select().single()
// INSERT many
.from("posts").insert([...]).select()
// UPSERT
.from("posts").upsert(data, { onConflict: "id" })
// UPDATE
.from("posts").update({ title }).eq("id", id)
// DELETE
.from("posts").delete().eq("id", id)
```

## Auth

```typescript
// Sign up
supabase.auth.signUp({ email, password })
// Sign in
supabase.auth.signInWithPassword({ email, password })
// OAuth
supabase.auth.signInWithOAuth({ provider: "google" })
// Magic link
supabase.auth.signInWithOtp({ email })
// Get user (verified — use on server)
supabase.auth.getUser()
// Get session (from local storage — client only)
supabase.auth.getSession()
// Sign out
supabase.auth.signOut()
// Password reset
supabase.auth.resetPasswordForEmail(email)
// Listen for changes
supabase.auth.onAuthStateChange((event, session) => { ... })
```

## RLS Policies

```sql
alter table posts enable row level security;

-- SELECT
create policy "name" on posts for select
  using (auth.uid() = user_id);

-- INSERT
create policy "name" on posts for insert
  with check (auth.uid() = user_id);

-- UPDATE
create policy "name" on posts for update
  using (auth.uid() = user_id)
  with check (auth.uid() = user_id);

-- DELETE
create policy "name" on posts for delete
  using (auth.uid() = user_id);

-- All operations
create policy "name" on posts for all
  using (condition);
```

## Storage

```typescript
// Upload
.storage.from("bucket").upload(path, file, { upsert: true })
// Download
.storage.from("bucket").download(path)
// Public URL
.storage.from("bucket").getPublicUrl(path)
// Signed URL
.storage.from("bucket").createSignedUrl(path, 3600)
// List
.storage.from("bucket").list(folder)
// Delete
.storage.from("bucket").remove([path1, path2])
// Transform
.storage.from("bucket").getPublicUrl(path, {
  transform: { width: 200, height: 200, resize: "cover" }
})
```

## Realtime

```typescript
// Database changes
supabase.channel("name")
  .on("postgres_changes",
    { event: "INSERT", schema: "public", table: "posts" },
    (payload) => { /* payload.new */ }
  ).subscribe()

// Broadcast (no DB, peer-to-peer)
channel.send({ type: "broadcast", event: "cursor", payload: { x, y } })
channel.on("broadcast", { event: "cursor" }, ({ payload }) => { ... })

// Presence (who's online)
channel.track({ status: "online" })
channel.on("presence", { event: "sync" }, () => {
  const state = channel.presenceState()
})

// Cleanup
supabase.removeChannel(channel)
```

## RPC (Database Functions)

```typescript
const { data } = await supabase.rpc("function_name", { arg1: "value" });
```

## SQL Essentials

```sql
-- Enable realtime
alter publication supabase_realtime add table posts;

-- Auto-create profile on signup
create trigger on_auth_user_created
  after insert on auth.users
  for each row execute function handle_new_user();

-- updated_at trigger
create trigger set_updated_at
  before update on posts
  for each row execute function update_modified_column();
```

## Environment Variables

```bash
NEXT_PUBLIC_SUPABASE_URL=https://xxx.supabase.co
NEXT_PUBLIC_SUPABASE_ANON_KEY=eyJ...        # Safe for client
SUPABASE_SERVICE_ROLE_KEY=eyJ...             # Server only!
```

## Key Rules

- Always enable RLS on every table
- Use `getUser()` on server (verified), `getSession()` on client (UI only)
- Service role key = bypasses RLS, server only, never expose
- Realtime needs table publication enabled
- Broadcast is ephemeral, presence is ephemeral
- Generate types after every schema change
- Private buckets + signed URLs for sensitive files
