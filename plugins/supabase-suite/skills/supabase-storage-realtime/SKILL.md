---
name: supabase-storage-realtime
description: >
  Supabase Storage and Realtime — file uploads, signed URLs, image transforms,
  storage policies, realtime subscriptions, presence channels, and broadcast.
  Triggers: "supabase storage", "supabase upload", "supabase file", "supabase bucket",
  "supabase realtime", "supabase subscribe", "supabase presence", "supabase broadcast",
  "supabase channel", "signed url".
  NOT for: auth (use supabase-auth), database queries (use supabase-database).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Supabase Storage & Realtime

## Storage Setup

```sql
-- Create a private bucket (RLS-protected)
insert into storage.buckets (id, name, public)
values ('avatars', 'avatars', false);

-- Create a public bucket (no auth needed for reads)
insert into storage.buckets (id, name, public, file_size_limit, allowed_mime_types)
values (
  'public-assets',
  'public-assets',
  true,
  5242880,  -- 5MB limit
  array['image/jpeg', 'image/png', 'image/webp', 'image/gif']
);
```

## Storage Policies

```sql
-- Users can upload their own avatar
create policy "Users can upload avatar"
  on storage.objects for insert
  with check (
    bucket_id = 'avatars'
    and auth.uid()::text = (storage.foldername(name))[1]
  );

-- Users can update their own avatar
create policy "Users can update avatar"
  on storage.objects for update
  using (
    bucket_id = 'avatars'
    and auth.uid()::text = (storage.foldername(name))[1]
  );

-- Users can read any avatar
create policy "Anyone can read avatars"
  on storage.objects for select
  using (bucket_id = 'avatars');

-- Users can delete their own avatar
create policy "Users can delete own avatar"
  on storage.objects for delete
  using (
    bucket_id = 'avatars'
    and auth.uid()::text = (storage.foldername(name))[1]
  );

-- Org-based storage access
create policy "Org members can access files"
  on storage.objects for all
  using (
    bucket_id = 'org-files'
    and (storage.foldername(name))[1] in (
      select org_id::text from org_members
      where user_id = auth.uid()
    )
  );
```

## File Upload

```typescript
// Upload a file
const file = event.target.files[0];
const fileExt = file.name.split(".").pop();
const fileName = `${user.id}/${Date.now()}.${fileExt}`;

const { data, error } = await supabase.storage
  .from("avatars")
  .upload(fileName, file, {
    cacheControl: "3600",
    upsert: true,   // Overwrite if exists
    contentType: file.type,
  });

// Get public URL (for public buckets)
const { data: { publicUrl } } = supabase.storage
  .from("public-assets")
  .getPublicUrl("path/to/file.jpg");

// Get signed URL (for private buckets)
const { data: { signedUrl } } = await supabase.storage
  .from("avatars")
  .createSignedUrl("user-id/avatar.jpg", 3600); // 1 hour expiry

// Download a file
const { data: blob } = await supabase.storage
  .from("documents")
  .download("report.pdf");
```

## Image Transforms

```typescript
// Get transformed image URL (on-the-fly)
const { data: { publicUrl } } = supabase.storage
  .from("avatars")
  .getPublicUrl("user-id/avatar.jpg", {
    transform: {
      width: 200,
      height: 200,
      resize: "cover",  // 'cover' | 'contain' | 'fill'
      quality: 80,
      format: "origin",  // 'origin' to keep original format
    },
  });

// Signed URL with transform
const { data } = await supabase.storage
  .from("photos")
  .createSignedUrl("vacation/beach.jpg", 3600, {
    transform: {
      width: 800,
      height: 600,
      resize: "contain",
    },
  });
```

## File Management

```typescript
// List files in a folder
const { data: files } = await supabase.storage
  .from("documents")
  .list("user-id/reports", {
    limit: 100,
    offset: 0,
    sortBy: { column: "created_at", order: "desc" },
  });

// Move/rename a file
const { error } = await supabase.storage
  .from("documents")
  .move("old/path/file.pdf", "new/path/file.pdf");

// Copy a file
const { error } = await supabase.storage
  .from("documents")
  .copy("original.pdf", "copy.pdf");

// Delete files
const { error } = await supabase.storage
  .from("documents")
  .remove(["file1.pdf", "file2.pdf"]);
```

## React Upload Component

```typescript
"use client";
import { useState } from "react";
import { createClient } from "@/lib/supabase/client";

export function AvatarUpload({ userId }: { userId: string }) {
  const supabase = createClient();
  const [uploading, setUploading] = useState(false);
  const [avatarUrl, setAvatarUrl] = useState<string | null>(null);

  async function handleUpload(event: React.ChangeEvent<HTMLInputElement>) {
    const file = event.target.files?.[0];
    if (!file) return;

    // Validate
    if (file.size > 2 * 1024 * 1024) {
      alert("File must be under 2MB");
      return;
    }

    setUploading(true);
    const fileExt = file.name.split(".").pop();
    const filePath = `${userId}/avatar.${fileExt}`;

    const { error: uploadError } = await supabase.storage
      .from("avatars")
      .upload(filePath, file, { upsert: true });

    if (uploadError) {
      console.error(uploadError);
      setUploading(false);
      return;
    }

    // Get public URL
    const { data: { publicUrl } } = supabase.storage
      .from("avatars")
      .getPublicUrl(filePath);

    // Update profile
    await supabase
      .from("profiles")
      .update({ avatar_url: publicUrl })
      .eq("id", userId);

    setAvatarUrl(publicUrl);
    setUploading(false);
  }

  return (
    <div>
      {avatarUrl && <img src={avatarUrl} alt="Avatar" className="w-20 h-20 rounded-full" />}
      <input
        type="file"
        accept="image/*"
        onChange={handleUpload}
        disabled={uploading}
      />
      {uploading && <p>Uploading...</p>}
    </div>
  );
}
```

---

## Realtime — Database Changes

```typescript
// Subscribe to INSERT events on a table
const channel = supabase
  .channel("posts-changes")
  .on(
    "postgres_changes",
    {
      event: "INSERT",
      schema: "public",
      table: "posts",
      filter: "status=eq.published",  // Optional filter
    },
    (payload) => {
      console.log("New post:", payload.new);
      setPosts(prev => [payload.new as Post, ...prev]);
    }
  )
  .subscribe();

// Subscribe to all changes (INSERT, UPDATE, DELETE)
const channel = supabase
  .channel("posts-all")
  .on(
    "postgres_changes",
    { event: "*", schema: "public", table: "posts" },
    (payload) => {
      switch (payload.eventType) {
        case "INSERT":
          setPosts(prev => [payload.new as Post, ...prev]);
          break;
        case "UPDATE":
          setPosts(prev =>
            prev.map(p => p.id === payload.new.id ? payload.new as Post : p)
          );
          break;
        case "DELETE":
          setPosts(prev => prev.filter(p => p.id !== payload.old.id));
          break;
      }
    }
  )
  .subscribe();

// Unsubscribe
supabase.removeChannel(channel);
```

## Realtime — Broadcast (Custom Events)

```typescript
// Broadcast is peer-to-peer via Supabase — no database involved.
// Great for: cursors, typing indicators, game state, notifications.

// Send a broadcast event
const channel = supabase.channel("room:123");

channel.subscribe((status) => {
  if (status === "SUBSCRIBED") {
    channel.send({
      type: "broadcast",
      event: "cursor-move",
      payload: { x: 100, y: 200, userId: "user-1" },
    });
  }
});

// Listen for broadcast events
channel
  .on("broadcast", { event: "cursor-move" }, ({ payload }) => {
    updateCursor(payload.userId, payload.x, payload.y);
  })
  .subscribe();
```

## Realtime — Presence (Who's Online)

```typescript
// Track who's online in a room
const channel = supabase.channel("room:123", {
  config: {
    presence: { key: userId },
  },
});

// Track your own presence
channel
  .on("presence", { event: "sync" }, () => {
    const state = channel.presenceState();
    // state = { "user-1": [{ online_at: "...", status: "active" }], ... }
    setOnlineUsers(Object.keys(state));
  })
  .on("presence", { event: "join" }, ({ key, newPresences }) => {
    console.log(`${key} joined`, newPresences);
  })
  .on("presence", { event: "leave" }, ({ key, leftPresences }) => {
    console.log(`${key} left`, leftPresences);
  })
  .subscribe(async (status) => {
    if (status === "SUBSCRIBED") {
      await channel.track({
        online_at: new Date().toISOString(),
        status: "active",
      });
    }
  });

// Update your presence (e.g., idle/active status)
await channel.track({
  online_at: new Date().toISOString(),
  status: "idle",
});

// Untrack (go offline)
await channel.untrack();
```

## React Realtime Hook

```typescript
"use client";
import { useEffect, useState } from "react";
import { createClient } from "@/lib/supabase/client";
import type { RealtimeChannel } from "@supabase/supabase-js";

export function useRealtimePosts(initialPosts: Post[]) {
  const [posts, setPosts] = useState(initialPosts);
  const supabase = createClient();

  useEffect(() => {
    const channel = supabase
      .channel("posts-realtime")
      .on(
        "postgres_changes",
        { event: "*", schema: "public", table: "posts" },
        (payload) => {
          if (payload.eventType === "INSERT") {
            setPosts(prev => [payload.new as Post, ...prev]);
          } else if (payload.eventType === "UPDATE") {
            setPosts(prev =>
              prev.map(p => p.id === payload.new.id ? payload.new as Post : p)
            );
          } else if (payload.eventType === "DELETE") {
            setPosts(prev => prev.filter(p => p.id !== payload.old.id));
          }
        }
      )
      .subscribe();

    return () => {
      supabase.removeChannel(channel);
    };
  }, []);

  return posts;
}

// Usage in a page:
export default function PostsPage({ initialPosts }: { initialPosts: Post[] }) {
  const posts = useRealtimePosts(initialPosts);
  return posts.map(post => <PostCard key={post.id} post={post} />);
}
```

## Realtime Setup (SQL)

```sql
-- Enable realtime for specific tables
-- In Supabase Dashboard: Database > Replication > enable tables

-- Or via SQL:
alter publication supabase_realtime add table posts;
alter publication supabase_realtime add table comments;

-- Disable realtime for a table
alter publication supabase_realtime drop table posts;
```

## Gotchas

1. **Realtime requires table publication.** Tables must be added to the `supabase_realtime` publication before changes are broadcast. Do this in Dashboard > Database > Replication, or via `ALTER PUBLICATION`. Without it, subscriptions connect but never fire.

2. **Realtime respects RLS.** Database change events are filtered by the subscribing user's RLS policies. If a user can't SELECT a row, they won't receive its INSERT/UPDATE event. This is a security feature.

3. **Broadcast is not persisted.** Broadcast events are fire-and-forget. If a client is offline when an event is sent, they miss it. For persistent messages, write to the database and use postgres_changes instead.

4. **Presence state is ephemeral.** When all clients disconnect, presence state is lost. It's not stored in the database. For persistent online status, write to a database table.

5. **Storage policies use `storage.foldername()`.** The `name` column in `storage.objects` is the full path. Use `storage.foldername(name)` to get path segments for folder-based access control. `(storage.foldername(name))[1]` gets the first folder (often the user ID).

6. **Public bucket URLs are permanent and guessable.** Anyone with the URL can access public bucket files. For sensitive files, use private buckets with signed URLs that expire. Don't put user-uploaded content in public buckets unless it's truly meant to be public.
