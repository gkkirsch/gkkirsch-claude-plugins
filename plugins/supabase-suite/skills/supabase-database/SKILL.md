---
name: supabase-database
description: >
  Supabase database — queries with supabase-js, Row Level Security policies,
  database functions, triggers, views, and server-side admin patterns.
  Triggers: "supabase query", "supabase select", "supabase insert", "supabase rls",
  "row level security", "supabase policy", "supabase function", "supabase trigger".
  NOT for: authentication (use supabase-auth), storage or realtime (use supabase-storage-realtime).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Supabase Database & RLS

## Query Basics

```typescript
import { createClient } from "@/lib/supabase/client";
const supabase = createClient();

// SELECT
const { data, error } = await supabase
  .from("posts")
  .select("*");

// SELECT with specific columns
const { data } = await supabase
  .from("posts")
  .select("id, title, created_at");

// SELECT with relations (joins)
const { data } = await supabase
  .from("posts")
  .select(`
    id,
    title,
    author:profiles(id, full_name, avatar_url),
    comments(id, body, created_at, user:profiles(full_name))
  `);

// SELECT single row
const { data } = await supabase
  .from("posts")
  .select("*")
  .eq("id", postId)
  .single();  // Returns object, not array. Errors if 0 or 2+ rows.

// SELECT with count
const { data, count } = await supabase
  .from("posts")
  .select("*", { count: "exact" })
  .eq("status", "published");
```

## Filters

```typescript
// Equality
.eq("status", "published")           // status = 'published'
.neq("status", "draft")              // status != 'draft'

// Comparison
.gt("price", 100)                    // price > 100
.gte("price", 100)                   // price >= 100
.lt("price", 50)                     // price < 50
.lte("price", 50)                    // price <= 50

// Pattern matching
.like("title", "%hello%")            // LIKE (case-sensitive)
.ilike("title", "%hello%")           // ILIKE (case-insensitive)

// Arrays and ranges
.in("status", ["published", "draft"])  // IN
.contains("tags", ["react", "next"])   // @> (array contains)
.containedBy("tags", ["react", "next", "ts"])  // <@ (contained by)
.overlaps("tags", ["react", "vue"])    // && (arrays overlap)

// Null
.is("deleted_at", null)              // IS NULL
.not("deleted_at", "is", null)       // IS NOT NULL

// JSON
.eq("metadata->theme", "dark")       // JSON field
.eq("metadata->>name", "John")       // JSON text field

// Logical
.or("status.eq.published,status.eq.featured")
.and("price.gt.10,price.lt.100")

// Full-text search
.textSearch("title", "hello world", { type: "websearch" })
```

## Ordering, Pagination, Limiting

```typescript
// Order
.order("created_at", { ascending: false })
.order("title", { ascending: true, nullsFirst: false })

// Pagination (offset-based)
.range(0, 9)    // First 10 rows (0-indexed, inclusive)
.range(10, 19)  // Next 10 rows

// Limit
.limit(10)

// Combined pagination pattern
const page = 1;
const pageSize = 20;
const from = (page - 1) * pageSize;
const to = from + pageSize - 1;

const { data, count } = await supabase
  .from("posts")
  .select("*", { count: "exact" })
  .order("created_at", { ascending: false })
  .range(from, to);

const totalPages = Math.ceil((count ?? 0) / pageSize);
```

## INSERT, UPDATE, DELETE

```typescript
// INSERT single
const { data, error } = await supabase
  .from("posts")
  .insert({ title: "New Post", body: "Content", user_id: userId })
  .select()         // Return the inserted row
  .single();

// INSERT multiple
const { data, error } = await supabase
  .from("posts")
  .insert([
    { title: "Post 1", body: "..." },
    { title: "Post 2", body: "..." },
  ])
  .select();

// UPSERT (insert or update on conflict)
const { data, error } = await supabase
  .from("settings")
  .upsert(
    { user_id: userId, theme: "dark", language: "en" },
    { onConflict: "user_id" }  // Conflict column
  )
  .select()
  .single();

// UPDATE
const { data, error } = await supabase
  .from("posts")
  .update({ title: "Updated Title", updated_at: new Date().toISOString() })
  .eq("id", postId)
  .select()
  .single();

// DELETE
const { error } = await supabase
  .from("posts")
  .delete()
  .eq("id", postId);

// Soft delete pattern
const { error } = await supabase
  .from("posts")
  .update({ deleted_at: new Date().toISOString() })
  .eq("id", postId);
```

## Row Level Security (RLS)

```sql
-- Always enable RLS on every table
alter table posts enable row level security;

-- SELECT: Users can read their own posts
create policy "Users can read own posts"
  on posts for select
  using (auth.uid() = user_id);

-- SELECT: Published posts are public
create policy "Published posts are public"
  on posts for select
  using (status = 'published');

-- INSERT: Authenticated users can create posts
create policy "Authenticated users can create posts"
  on posts for insert
  with check (auth.uid() = user_id);

-- UPDATE: Users can update own posts
create policy "Users can update own posts"
  on posts for update
  using (auth.uid() = user_id)          -- Which rows can be targeted
  with check (auth.uid() = user_id);    -- What the new row must satisfy

-- DELETE: Users can delete own posts
create policy "Users can delete own posts"
  on posts for delete
  using (auth.uid() = user_id);

-- Admin access: full CRUD for admins
create policy "Admins have full access"
  on posts for all
  using (
    exists (
      select 1 from profiles
      where profiles.id = auth.uid()
      and profiles.role = 'admin'
    )
  );
```

## RLS Patterns for Common Scenarios

```sql
-- Multi-tenant: org-based access
create policy "Org members can read"
  on documents for select
  using (
    org_id in (
      select org_id from org_members
      where user_id = auth.uid()
    )
  );

-- Team-based with roles
create policy "Team members can read, managers can write"
  on projects for select
  using (
    exists (
      select 1 from team_members
      where team_members.project_id = projects.id
      and team_members.user_id = auth.uid()
    )
  );

create policy "Managers can update"
  on projects for update
  using (
    exists (
      select 1 from team_members
      where team_members.project_id = projects.id
      and team_members.user_id = auth.uid()
      and team_members.role = 'manager'
    )
  );

-- Public + private content
create policy "Public items visible to all"
  on items for select
  using (is_public = true);

create policy "Private items visible to owner"
  on items for select
  using (auth.uid() = user_id);

-- Time-based access
create policy "Can only edit within 24 hours"
  on comments for update
  using (
    auth.uid() = user_id
    and created_at > now() - interval '24 hours'
  );
```

## Database Functions

```sql
-- RPC function callable from client
create or replace function get_user_stats(target_user_id uuid)
returns json as $$
  select json_build_object(
    'post_count', (select count(*) from posts where user_id = target_user_id),
    'comment_count', (select count(*) from comments where user_id = target_user_id),
    'total_likes', (select coalesce(sum(like_count), 0) from posts where user_id = target_user_id)
  );
$$ language sql security definer;  -- Runs with function owner's permissions

-- Call from client
const { data } = await supabase.rpc("get_user_stats", {
  target_user_id: userId,
});
```

```sql
-- Increment function (atomic)
create or replace function increment_view_count(post_id uuid)
returns void as $$
  update posts
  set view_count = view_count + 1
  where id = post_id;
$$ language sql security definer;

-- Search function with full-text search
create or replace function search_posts(search_query text)
returns setof posts as $$
  select *
  from posts
  where
    to_tsvector('english', title || ' ' || body)
    @@ plainto_tsquery('english', search_query)
  order by ts_rank(
    to_tsvector('english', title || ' ' || body),
    plainto_tsquery('english', search_query)
  ) desc;
$$ language sql security definer;
```

## Triggers

```sql
-- Auto-update updated_at
create or replace function update_modified_column()
returns trigger as $$
begin
  new.updated_at = now();
  return new;
end;
$$ language plpgsql;

create trigger update_posts_modtime
  before update on posts
  for each row execute function update_modified_column();

-- Notify on new comments (for realtime or webhooks)
create or replace function notify_new_comment()
returns trigger as $$
begin
  -- Update comment count on post
  update posts
  set comment_count = comment_count + 1
  where id = new.post_id;

  return new;
end;
$$ language plpgsql security definer;

create trigger on_new_comment
  after insert on comments
  for each row execute function notify_new_comment();
```

## Views

```sql
-- Create a view for common queries
create or replace view public.post_with_author as
  select
    p.id,
    p.title,
    p.body,
    p.status,
    p.created_at,
    p.updated_at,
    p.user_id,
    pr.full_name as author_name,
    pr.avatar_url as author_avatar,
    (select count(*) from comments c where c.post_id = p.id) as comment_count,
    (select count(*) from likes l where l.post_id = p.id) as like_count
  from posts p
  join profiles pr on p.user_id = pr.id;

-- Query the view like a table
const { data } = await supabase
  .from("post_with_author")
  .select("*")
  .eq("status", "published")
  .order("created_at", { ascending: false })
  .limit(10);
```

## Admin / Service Role Client

```typescript
// lib/supabase/admin.ts — Server-only, bypasses RLS
import { createClient } from "@supabase/supabase-js";

export const supabaseAdmin = createClient(
  process.env.NEXT_PUBLIC_SUPABASE_URL!,
  process.env.SUPABASE_SERVICE_ROLE_KEY!,  // Never expose this
  {
    auth: {
      autoRefreshToken: false,
      persistSession: false,
    },
  }
);

// Usage: background jobs, webhooks, admin operations
const { data } = await supabaseAdmin
  .from("users")
  .select("*");  // Bypasses all RLS policies
```

## TypeScript Types from Schema

```bash
# Generate types from your Supabase schema
npx supabase gen types typescript --project-id your-project-id > lib/database.types.ts
# Or from a local database:
npx supabase gen types typescript --local > lib/database.types.ts
```

```typescript
// lib/supabase/client.ts — Typed client
import { createBrowserClient } from "@supabase/ssr";
import type { Database } from "@/lib/database.types";

export function createClient() {
  return createBrowserClient<Database>(
    process.env.NEXT_PUBLIC_SUPABASE_URL!,
    process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY!
  );
}

// Now queries are fully typed:
const { data } = await supabase
  .from("posts")        // Autocomplete table names
  .select("id, title")  // Autocomplete column names
  .eq("status", "published");
// data is typed as { id: string; title: string }[] | null
```

## Gotchas

1. **RLS policies are additive for the same operation.** Two SELECT policies on the same table = either can grant access (OR). This is different from middleware where you chain checks. Design policies carefully to avoid accidentally granting too much access.

2. **`security definer` functions bypass RLS.** Functions with `security definer` run as the function owner (usually `postgres`), bypassing RLS. Use this intentionally for admin operations, but be careful — a client can call any function via `supabase.rpc()`.

3. **`.single()` throws if no rows match.** Use `.maybeSingle()` if the row might not exist. `.single()` returns an error with code `PGRST116` when zero rows are found.

4. **Generated types get stale.** Re-run `supabase gen types` after every schema change. Stale types cause runtime errors that TypeScript can't catch.

5. **Cascading deletes need explicit foreign keys.** Supabase doesn't auto-cascade. Define `on delete cascade` in your foreign key constraints, or handle cleanup in triggers.

6. **`auth.uid()` is null for unauthenticated requests.** If a policy uses `auth.uid() = user_id` and the request is unauthenticated, the expression evaluates to `null = user_id` which is always false. This is correct behavior (deny by default), but can confuse when debugging.
