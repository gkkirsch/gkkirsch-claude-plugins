---
name: nextjs-server-actions
description: >
  Next.js Server Actions — form handling, mutations, optimistic updates,
  revalidation, error handling, and progressive enhancement patterns.
  Triggers: "server actions", "next.js forms", "use server", "form action",
  "revalidatePath", "optimistic update", "next.js mutations".
  NOT for: data fetching/reading (use data-fetching skill), routing (use app-router skill).
version: 1.0.0
allowed-tools: Read, Grep, Glob, Bash, Edit, Write
---

# Next.js Server Actions

Server Actions are async functions that run on the server, callable directly from Client and Server Components. They replace API routes for most mutations.

## Basic Server Action

### In a Separate File (Recommended)

```typescript
// actions/posts.ts
'use server';

import { db } from '@/lib/db';
import { revalidatePath } from 'next/cache';
import { redirect } from 'next/navigation';
import { z } from 'zod';

const CreatePostSchema = z.object({
  title: z.string().min(1, 'Title is required').max(200),
  content: z.string().min(1, 'Content is required'),
  published: z.coerce.boolean().optional().default(false),
});

export async function createPost(formData: FormData) {
  const parsed = CreatePostSchema.safeParse({
    title: formData.get('title'),
    content: formData.get('content'),
    published: formData.get('published'),
  });

  if (!parsed.success) {
    return { error: parsed.error.flatten().fieldErrors };
  }

  const post = await db.post.create({
    data: parsed.data,
  });

  revalidatePath('/posts');
  redirect(`/posts/${post.id}`);
}
```

### Using in a Form

```tsx
// app/posts/new/page.tsx
import { createPost } from '@/actions/posts';

export default function NewPostPage() {
  return (
    <form action={createPost}>
      <input name="title" placeholder="Post title" required />
      <textarea name="content" placeholder="Write your post..." required />
      <label>
        <input type="checkbox" name="published" value="true" />
        Publish immediately
      </label>
      <button type="submit">Create Post</button>
    </form>
  );
}
```

This form works without JavaScript (progressive enhancement). When JS is available, it submits without a full page reload.

## Form State with useActionState

```tsx
// components/forms/create-post-form.tsx
'use client';

import { useActionState } from 'react';
import { createPost } from '@/actions/posts';

type State = {
  error?: Record<string, string[]>;
  message?: string;
};

export function CreatePostForm() {
  const [state, formAction, isPending] = useActionState(
    async (prevState: State, formData: FormData): Promise<State> => {
      const result = await createPost(formData);
      if (result?.error) return { error: result.error };
      return { message: 'Post created!' };
    },
    { error: undefined, message: undefined }
  );

  return (
    <form action={formAction}>
      <div>
        <input name="title" placeholder="Post title" />
        {state.error?.title && (
          <p className="text-red-500 text-sm">{state.error.title[0]}</p>
        )}
      </div>

      <div>
        <textarea name="content" placeholder="Write your post..." />
        {state.error?.content && (
          <p className="text-red-500 text-sm">{state.error.content[0]}</p>
        )}
      </div>

      <button type="submit" disabled={isPending}>
        {isPending ? 'Creating...' : 'Create Post'}
      </button>

      {state.message && (
        <p className="text-green-500">{state.message}</p>
      )}
    </form>
  );
}
```

## Pending States with useFormStatus

```tsx
'use client';

import { useFormStatus } from 'react-dom';

function SubmitButton({ children }: { children: React.ReactNode }) {
  const { pending } = useFormStatus();

  return (
    <button
      type="submit"
      disabled={pending}
      className={pending ? 'opacity-50 cursor-not-allowed' : ''}
    >
      {pending ? (
        <span className="flex items-center gap-2">
          <Spinner className="w-4 h-4" />
          Saving...
        </span>
      ) : (
        children
      )}
    </button>
  );
}

// Use in any form
<form action={createPost}>
  <input name="title" />
  <SubmitButton>Create Post</SubmitButton>  {/* Auto-detects form status */}
</form>
```

## Optimistic Updates

```tsx
'use client';

import { useOptimistic } from 'react';
import { toggleLike } from '@/actions/posts';

type Post = { id: string; title: string; liked: boolean; likeCount: number };

export function PostCard({ post }: { post: Post }) {
  const [optimisticPost, setOptimisticPost] = useOptimistic(
    post,
    (current, liked: boolean) => ({
      ...current,
      liked,
      likeCount: liked ? current.likeCount + 1 : current.likeCount - 1,
    })
  );

  async function handleLike() {
    const newLiked = !optimisticPost.liked;
    setOptimisticPost(newLiked);  // Instant UI update
    await toggleLike(post.id);    // Server mutation (may revert on error)
  }

  return (
    <div>
      <h3>{optimisticPost.title}</h3>
      <button onClick={handleLike}>
        {optimisticPost.liked ? '❤️' : '🤍'} {optimisticPost.likeCount}
      </button>
    </div>
  );
}
```

## Revalidation Patterns

```typescript
'use server';

import { revalidatePath, revalidateTag } from 'next/cache';

export async function updatePost(id: string, formData: FormData) {
  await db.post.update({
    where: { id },
    data: { title: formData.get('title') as string },
  });

  // Option 1: Revalidate specific path
  revalidatePath('/posts');              // List page
  revalidatePath(`/posts/${id}`);        // Detail page

  // Option 2: Revalidate by tag (more precise)
  revalidateTag('posts');                // All fetches tagged 'posts'
  revalidateTag(`post-${id}`);           // Specific post

  // Option 3: Revalidate layout (affects all pages in segment)
  revalidatePath('/posts', 'layout');
}
```

### Tag-Based Revalidation (Recommended)

```typescript
// Data fetching with tags
async function getPost(id: string) {
  const res = await fetch(`${API_URL}/posts/${id}`, {
    next: { tags: ['posts', `post-${id}`] },
  });
  return res.json();
}

// Revalidate just this post everywhere it's fetched
export async function updatePost(id: string, data: PostData) {
  await db.post.update({ where: { id }, data });
  revalidateTag(`post-${id}`);  // Only refetches this post's data
}
```

## Authentication in Server Actions

```typescript
'use server';

import { getSession } from '@/lib/auth';

export async function createPost(formData: FormData) {
  const session = await getSession();
  if (!session) {
    throw new Error('Unauthorized');
  }

  const post = await db.post.create({
    data: {
      title: formData.get('title') as string,
      content: formData.get('content') as string,
      authorId: session.user.id,
    },
  });

  revalidatePath('/posts');
  redirect(`/posts/${post.id}`);
}
```

### Authorization Wrapper

```typescript
// lib/action-utils.ts
import { getSession } from '@/lib/auth';

export function authenticatedAction<T>(
  action: (session: Session, ...args: any[]) => Promise<T>
) {
  return async (...args: any[]): Promise<T> => {
    const session = await getSession();
    if (!session) {
      throw new Error('Unauthorized');
    }
    return action(session, ...args);
  };
}

// Usage
export const createPost = authenticatedAction(
  async (session, formData: FormData) => {
    // session is guaranteed to exist
    await db.post.create({
      data: {
        title: formData.get('title') as string,
        authorId: session.user.id,
      },
    });
    revalidatePath('/posts');
  }
);
```

## Delete Pattern with Confirmation

```tsx
'use client';

import { deletePost } from '@/actions/posts';
import { useTransition } from 'react';

export function DeleteButton({ postId }: { postId: string }) {
  const [isPending, startTransition] = useTransition();

  function handleDelete() {
    if (!confirm('Are you sure you want to delete this post?')) return;

    startTransition(async () => {
      await deletePost(postId);
    });
  }

  return (
    <button
      onClick={handleDelete}
      disabled={isPending}
      className="text-red-500 hover:text-red-700"
    >
      {isPending ? 'Deleting...' : 'Delete'}
    </button>
  );
}
```

```typescript
// actions/posts.ts
'use server';

export async function deletePost(id: string) {
  const session = await getSession();
  if (!session) throw new Error('Unauthorized');

  const post = await db.post.findUnique({ where: { id } });
  if (post?.authorId !== session.user.id) {
    throw new Error('Forbidden');
  }

  await db.post.delete({ where: { id } });
  revalidatePath('/posts');
  redirect('/posts');
}
```

## File Upload

```typescript
// actions/upload.ts
'use server';

import { writeFile } from 'fs/promises';
import { join } from 'path';
import { nanoid } from 'nanoid';

export async function uploadFile(formData: FormData) {
  const file = formData.get('file') as File;
  if (!file || file.size === 0) {
    return { error: 'No file provided' };
  }

  // Validate
  const MAX_SIZE = 5 * 1024 * 1024; // 5MB
  if (file.size > MAX_SIZE) {
    return { error: 'File too large (max 5MB)' };
  }

  const ALLOWED_TYPES = ['image/jpeg', 'image/png', 'image/webp'];
  if (!ALLOWED_TYPES.includes(file.type)) {
    return { error: 'Invalid file type' };
  }

  // Save
  const bytes = await file.arrayBuffer();
  const buffer = Buffer.from(bytes);
  const ext = file.name.split('.').pop();
  const filename = `${nanoid()}.${ext}`;
  const path = join(process.cwd(), 'public/uploads', filename);

  await writeFile(path, buffer);

  return { url: `/uploads/${filename}` };
}
```

```tsx
// For production, upload to S3/Cloudinary instead of local filesystem
'use client';

export function UploadForm() {
  const [preview, setPreview] = useState<string | null>(null);

  return (
    <form action={uploadFile}>
      <input
        type="file"
        name="file"
        accept="image/*"
        onChange={(e) => {
          const file = e.target.files?.[0];
          if (file) setPreview(URL.createObjectURL(file));
        }}
      />
      {preview && <img src={preview} alt="Preview" className="w-32 h-32 object-cover" />}
      <SubmitButton>Upload</SubmitButton>
    </form>
  );
}
```

## Error Handling Pattern

```typescript
// lib/action-result.ts
type ActionResult<T = void> =
  | { success: true; data: T }
  | { success: false; error: string; fieldErrors?: Record<string, string[]> };

// actions/posts.ts
'use server';

export async function createPost(formData: FormData): Promise<ActionResult<{ id: string }>> {
  try {
    const session = await getSession();
    if (!session) return { success: false, error: 'Please sign in' };

    const parsed = CreatePostSchema.safeParse({
      title: formData.get('title'),
      content: formData.get('content'),
    });

    if (!parsed.success) {
      return {
        success: false,
        error: 'Validation failed',
        fieldErrors: parsed.error.flatten().fieldErrors,
      };
    }

    const post = await db.post.create({
      data: { ...parsed.data, authorId: session.user.id },
    });

    revalidatePath('/posts');
    return { success: true, data: { id: post.id } };
  } catch (err) {
    console.error('createPost failed:', err);
    return { success: false, error: 'Something went wrong. Please try again.' };
  }
}
```

## Common Gotchas

1. **Server Actions must be async** — even if they don't await anything. The `'use server'` directive requires it.

2. **FormData values are strings** — `formData.get('count')` returns `string | null`, not `number`. Use Zod's `z.coerce.number()` for type conversion.

3. **Can't return non-serializable data** — Server Actions communicate via JSON. No functions, classes, Dates (use ISO strings), or circular references.

4. **redirect() throws** — don't put `redirect()` in a try/catch unless you re-throw. It uses a special error type internally.

5. **Revalidation is eventual** — `revalidatePath` marks the cache as stale. The next request fetches fresh data. It's not synchronous.

6. **Server Actions are POST requests** — they're not cached, they're not idempotent by default. Treat them like form submissions.

7. **File uploads need multipart** — forms with file inputs work natively with Server Actions. The FormData includes File objects automatically.

8. **Closure over server values** — Server Actions defined inline in Server Components can access server-side values via closure. But those values are serialized and sent to the client as part of the action reference. Don't close over secrets.
