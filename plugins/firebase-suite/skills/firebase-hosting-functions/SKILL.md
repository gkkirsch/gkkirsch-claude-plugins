---
name: firebase-hosting-functions
description: >
  Firebase Cloud Functions, Hosting, and Storage — function triggers,
  callable functions, scheduled functions, Hosting configuration, Storage
  uploads and security rules.
  Triggers: "firebase functions", "cloud functions", "firebase hosting",
  "firebase storage", "firebase deploy", "firebase callable",
  "firebase trigger", "firebase scheduled function", "firebase cron".
  NOT for: Auth or Firestore queries (use firebase-auth, firebase-firestore).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Firebase Cloud Functions, Hosting & Storage

## Cloud Functions Setup

```bash
# Initialize Cloud Functions
firebase init functions

# Choose TypeScript
# Creates: functions/src/index.ts, functions/package.json, functions/tsconfig.json
```

```typescript
// functions/src/index.ts
import { onRequest, onCall, HttpsError } from "firebase-functions/v2/https";
import { onDocumentCreated, onDocumentUpdated, onDocumentDeleted } from "firebase-functions/v2/firestore";
import { onSchedule } from "firebase-functions/v2/scheduler";
import { onObjectFinalized } from "firebase-functions/v2/storage";
import { initializeApp } from "firebase-admin/app";
import { getFirestore, FieldValue } from "firebase-admin/firestore";
import { getAuth } from "firebase-admin/auth";

initializeApp();
const db = getFirestore();
```

## HTTP Functions

```typescript
// Basic HTTP endpoint
export const api = onRequest(
  { cors: true, region: "us-central1" },
  async (req, res) => {
    if (req.method !== "GET") {
      res.status(405).send("Method not allowed");
      return;
    }

    const posts = await db.collection("posts")
      .orderBy("createdAt", "desc")
      .limit(10)
      .get();

    const data = posts.docs.map(doc => ({ id: doc.id, ...doc.data() }));
    res.json({ data });
  }
);

// Express-style with middleware
import express from "express";

const app = express();
app.use(express.json());

// Auth middleware
app.use(async (req, res, next) => {
  const authHeader = req.headers.authorization;
  if (!authHeader?.startsWith("Bearer ")) {
    res.status(401).json({ error: "Unauthorized" });
    return;
  }
  try {
    const token = authHeader.split("Bearer ")[1];
    req.user = await getAuth().verifyIdToken(token);
    next();
  } catch {
    res.status(401).json({ error: "Invalid token" });
  }
});

app.get("/posts", async (req, res) => {
  const posts = await db.collection("posts")
    .where("authorId", "==", req.user.uid)
    .get();
  res.json({ data: posts.docs.map(d => ({ id: d.id, ...d.data() })) });
});

app.post("/posts", async (req, res) => {
  const ref = await db.collection("posts").add({
    ...req.body,
    authorId: req.user.uid,
    createdAt: FieldValue.serverTimestamp(),
  });
  res.status(201).json({ id: ref.id });
});

export const apiV2 = onRequest({ cors: true }, app);
```

## Callable Functions

```typescript
// SERVER: Define callable function
export const createPost = onCall(
  { enforceAppCheck: false },
  async (request) => {
    // Auth is automatic — request.auth is populated
    if (!request.auth) {
      throw new HttpsError("unauthenticated", "Must be signed in");
    }

    const { title, body } = request.data;

    if (!title || !body) {
      throw new HttpsError("invalid-argument", "Title and body required");
    }

    const ref = await db.collection("posts").add({
      title,
      body,
      authorId: request.auth.uid,
      authorName: request.auth.token.name || "Anonymous",
      createdAt: FieldValue.serverTimestamp(),
    });

    return { id: ref.id };
  }
);

// CLIENT: Call it
import { getFunctions, httpsCallable } from "firebase/functions";

const functions = getFunctions();
const createPostFn = httpsCallable(functions, "createPost");

try {
  const result = await createPostFn({ title: "Hello", body: "World" });
  console.log("Created:", result.data.id);
} catch (error: any) {
  console.error(error.code, error.message); // "invalid-argument", "Title and body required"
}
```

## Firestore Triggers

```typescript
// On document created
export const onPostCreated = onDocumentCreated(
  "posts/{postId}",
  async (event) => {
    const snapshot = event.data;
    if (!snapshot) return;

    const post = snapshot.data();

    // Update user's post count
    await db.doc(`users/${post.authorId}`).update({
      postCount: FieldValue.increment(1),
    });

    // Send notification
    console.log(`New post: ${post.title} by ${post.authorId}`);
  }
);

// On document updated
export const onPostUpdated = onDocumentUpdated(
  "posts/{postId}",
  async (event) => {
    const before = event.data?.before.data();
    const after = event.data?.after.data();

    if (!before || !after) return;

    // Detect specific field changes
    if (before.status !== after.status && after.status === "published") {
      // Post just got published — notify followers
      await notifyFollowers(after.authorId, event.params.postId);
    }
  }
);

// On document deleted
export const onPostDeleted = onDocumentDeleted(
  "posts/{postId}",
  async (event) => {
    const post = event.data?.data();
    if (!post) return;

    // Decrement user's post count
    await db.doc(`users/${post.authorId}`).update({
      postCount: FieldValue.increment(-1),
    });

    // Delete subcollection (comments)
    const comments = await db.collection(`posts/${event.params.postId}/comments`).get();
    const batch = db.batch();
    comments.docs.forEach(doc => batch.delete(doc.ref));
    await batch.commit();
  }
);
```

## Scheduled Functions

```typescript
// Run every day at midnight UTC
export const dailyCleanup = onSchedule(
  {
    schedule: "0 0 * * *",  // cron syntax
    timeZone: "America/New_York",
    retryCount: 3,
  },
  async () => {
    // Delete old temp data
    const cutoff = new Date();
    cutoff.setDate(cutoff.getDate() - 30);

    const old = await db.collection("temp")
      .where("createdAt", "<", cutoff)
      .limit(500)
      .get();

    const batch = db.batch();
    old.docs.forEach(doc => batch.delete(doc.ref));
    await batch.commit();

    console.log(`Cleaned up ${old.size} old documents`);
  }
);

// Run every 5 minutes
export const healthCheck = onSchedule("every 5 minutes", async () => {
  await db.doc("system/health").set({
    lastCheck: FieldValue.serverTimestamp(),
    status: "ok",
  });
});
```

## Storage Triggers

```typescript
// On file upload
export const onFileUploaded = onObjectFinalized(async (event) => {
  const filePath = event.data.name;      // "uploads/user123/photo.jpg"
  const contentType = event.data.contentType; // "image/jpeg"
  const bucket = event.data.bucket;

  if (!contentType?.startsWith("image/")) return;

  // Generate thumbnail using sharp
  const { getStorage } = await import("firebase-admin/storage");
  const storage = getStorage();
  const file = storage.bucket(bucket).file(filePath);

  const [buffer] = await file.download();
  const sharp = (await import("sharp")).default;

  const thumbnail = await sharp(buffer)
    .resize(200, 200, { fit: "cover" })
    .toBuffer();

  const thumbPath = filePath.replace(/(\.\w+)$/, "_thumb$1");
  await storage.bucket(bucket).file(thumbPath).save(thumbnail, {
    metadata: { contentType },
  });

  console.log(`Thumbnail created: ${thumbPath}`);
});
```

## Firebase Hosting

```json
// firebase.json
{
  "hosting": {
    "public": "dist",
    "ignore": ["firebase.json", "**/.*", "**/node_modules/**"],
    "rewrites": [
      {
        "source": "/api/**",
        "function": "apiV2"
      },
      {
        "source": "**",
        "destination": "/index.html"
      }
    ],
    "headers": [
      {
        "source": "**/*.@(js|css)",
        "headers": [
          {
            "key": "Cache-Control",
            "value": "public, max-age=31536000, immutable"
          }
        ]
      },
      {
        "source": "**/*.@(jpg|jpeg|gif|png|svg|webp)",
        "headers": [
          {
            "key": "Cache-Control",
            "value": "public, max-age=86400"
          }
        ]
      }
    ],
    "redirects": [
      {
        "source": "/old-page",
        "destination": "/new-page",
        "type": 301
      }
    ],
    "cleanUrls": true,
    "trailingSlash": false
  }
}
```

## Firebase Storage (Client)

```typescript
import {
  getStorage,
  ref,
  uploadBytes,
  uploadBytesResumable,
  getDownloadURL,
  deleteObject,
  listAll,
} from "firebase/storage";

const storage = getStorage();

// Simple upload
async function uploadFile(file: File, path: string) {
  const storageRef = ref(storage, path);
  const snapshot = await uploadBytes(storageRef, file, {
    contentType: file.type,
    customMetadata: {
      uploadedBy: auth.currentUser!.uid,
    },
  });
  const url = await getDownloadURL(snapshot.ref);
  return url;
}

// Resumable upload with progress
function uploadWithProgress(file: File, path: string) {
  const storageRef = ref(storage, path);
  const uploadTask = uploadBytesResumable(storageRef, file);

  return new Promise<string>((resolve, reject) => {
    uploadTask.on(
      "state_changed",
      (snapshot) => {
        const progress = (snapshot.bytesTransferred / snapshot.totalBytes) * 100;
        console.log(`Upload: ${progress.toFixed(0)}%`);
      },
      (error) => reject(error),
      async () => {
        const url = await getDownloadURL(uploadTask.snapshot.ref);
        resolve(url);
      }
    );
  });
}

// Delete file
async function deleteFile(path: string) {
  const storageRef = ref(storage, path);
  await deleteObject(storageRef);
}

// List files
async function listFiles(path: string) {
  const listRef = ref(storage, path);
  const result = await listAll(listRef);
  return result.items.map(item => item.fullPath);
}

// React upload component
function FileUpload() {
  const [progress, setProgress] = useState(0);
  const [url, setUrl] = useState("");

  const handleUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    const path = `uploads/${auth.currentUser!.uid}/${file.name}`;
    const storageRef = ref(storage, path);
    const uploadTask = uploadBytesResumable(storageRef, file);

    uploadTask.on("state_changed", (snap) => {
      setProgress((snap.bytesTransferred / snap.totalBytes) * 100);
    });

    await uploadTask;
    const downloadUrl = await getDownloadURL(uploadTask.snapshot.ref);
    setUrl(downloadUrl);
  };

  return (
    <div>
      <input type="file" onChange={handleUpload} />
      {progress > 0 && <progress value={progress} max={100} />}
      {url && <img src={url} alt="Uploaded" />}
    </div>
  );
}
```

## Storage Security Rules

```
rules_version = '2';

service firebase.storage {
  match /b/{bucket}/o {
    // User uploads — owner only
    match /uploads/{userId}/{allPaths=**} {
      allow read: if request.auth != null;
      allow write: if request.auth.uid == userId
        && request.resource.size < 10 * 1024 * 1024  // 10MB limit
        && request.resource.contentType.matches('image/.*');  // images only
    }

    // Public assets
    match /public/{allPaths=**} {
      allow read: if true;
      allow write: if request.auth != null
        && request.auth.token.role == 'admin';
    }

    // Profile photos
    match /profiles/{userId}/photo.{ext} {
      allow read: if true;
      allow write: if request.auth.uid == userId
        && request.resource.size < 5 * 1024 * 1024
        && ext.matches('jpg|jpeg|png|webp');
    }
  }
}
```

## Deploy

```bash
# Deploy everything
firebase deploy

# Deploy only functions
firebase deploy --only functions

# Deploy specific function
firebase deploy --only functions:createPost

# Deploy only hosting
firebase deploy --only hosting

# Deploy only rules
firebase deploy --only firestore:rules
firebase deploy --only storage

# Preview channel (temporary URL for testing)
firebase hosting:channel:deploy preview-branch

# Use emulators for local development
firebase emulators:start
firebase emulators:start --only auth,firestore,functions
```

## Local Emulator

```typescript
// Connect to emulators in development
import { connectAuthEmulator } from "firebase/auth";
import { connectFirestoreEmulator } from "firebase/firestore";
import { connectFunctionsEmulator } from "firebase/functions";
import { connectStorageEmulator } from "firebase/storage";

if (process.env.NODE_ENV === "development") {
  connectAuthEmulator(auth, "http://localhost:9099");
  connectFirestoreEmulator(db, "localhost", 8080);
  connectFunctionsEmulator(functions, "localhost", 5001);
  connectStorageEmulator(storage, "localhost", 9199);
}
```

## Gotchas

1. **Cold starts** — Cloud Functions have cold start latency (1-10s on first invocation). Use `minInstances: 1` in production for critical paths, but it costs more.

2. **Function timeout** — default 60s for HTTP, 540s for background triggers. Set `timeoutSeconds` in options if you need longer.

3. **Memory** — default 256MB. Increase with `memory: "1GiB"` for image processing or heavy computation.

4. **Region matters** — deploy functions in the same region as your Firestore database to minimize latency. Default is `us-central1`.

5. **Firestore triggers can fire multiple times** — design your trigger handlers to be idempotent. Use the event ID as a dedup key.

6. **Hosting rewrites execute in order** — put specific paths before the catch-all SPA rewrite. The first matching rule wins.

7. **Storage download URLs don't expire by default** — they contain a long-lived token. To revoke, delete and re-upload the file, or use signed URLs with expiration.
