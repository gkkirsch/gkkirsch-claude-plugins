---
name: firebase-firestore
description: >
  Firestore database — queries, security rules, data modeling, real-time
  listeners, transactions, batch writes, pagination, and offline support.
  Triggers: "firestore", "firestore query", "firestore rules", "firestore data model",
  "firestore real-time", "firestore transaction", "firestore pagination",
  "firestore offline", "firebase database", "firebase data".
  NOT for: Firebase Auth or Cloud Functions (use firebase-auth, firebase-hosting-functions).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Firestore Database

## Setup

```typescript
import { initializeApp } from "firebase/app";
import {
  getFirestore,
  collection,
  doc,
  addDoc,
  setDoc,
  getDoc,
  getDocs,
  updateDoc,
  deleteDoc,
  query,
  where,
  orderBy,
  limit,
  startAfter,
  onSnapshot,
  writeBatch,
  runTransaction,
  serverTimestamp,
  increment,
  arrayUnion,
  arrayRemove,
  Timestamp,
} from "firebase/firestore";

const app = initializeApp(firebaseConfig);
export const db = getFirestore(app);
```

## CRUD Operations

```typescript
// CREATE — auto-generated ID
const docRef = await addDoc(collection(db, "posts"), {
  title: "Hello World",
  body: "First post",
  authorId: auth.currentUser!.uid,
  tags: ["intro", "hello"],
  createdAt: serverTimestamp(),
  updatedAt: serverTimestamp(),
  viewCount: 0,
});
console.log("Created:", docRef.id);

// CREATE — specific ID
await setDoc(doc(db, "users", auth.currentUser!.uid), {
  displayName: "Alice",
  email: "alice@example.com",
  role: "member",
  createdAt: serverTimestamp(),
});

// READ — single document
const docSnap = await getDoc(doc(db, "posts", postId));
if (docSnap.exists()) {
  const data = docSnap.data();
  console.log(data.title); // "Hello World"
} else {
  console.log("Not found");
}

// READ — collection with query
const q = query(
  collection(db, "posts"),
  where("authorId", "==", userId),
  orderBy("createdAt", "desc"),
  limit(10)
);
const snapshot = await getDocs(q);
const posts = snapshot.docs.map(doc => ({
  id: doc.id,
  ...doc.data(),
}));

// UPDATE — merge fields
await updateDoc(doc(db, "posts", postId), {
  title: "Updated Title",
  updatedAt: serverTimestamp(),
});

// UPDATE — set with merge (creates if doesn't exist)
await setDoc(doc(db, "users", uid), {
  lastLogin: serverTimestamp(),
}, { merge: true });

// UPDATE — atomic operations
await updateDoc(doc(db, "posts", postId), {
  viewCount: increment(1),
  tags: arrayUnion("featured"),
  // tags: arrayRemove("draft"),  // remove from array
});

// DELETE
await deleteDoc(doc(db, "posts", postId));
```

## Query Operators

```typescript
// Equality
where("status", "==", "published")

// Inequality
where("age", "!=", 0)
where("price", ">", 10)
where("price", ">=", 10)
where("price", "<", 100)
where("price", "<=", 100)

// In (up to 30 values)
where("status", "in", ["active", "pending"])
where("category", "not-in", ["spam", "deleted"])

// Array contains
where("tags", "array-contains", "featured")
where("tags", "array-contains-any", ["react", "vue", "angular"])

// Compound queries (requires composite index)
const q = query(
  collection(db, "products"),
  where("category", "==", "electronics"),
  where("price", "<=", 500),
  orderBy("price", "asc"),
  limit(20)
);

// Note: inequality filters on different fields require composite indexes
// Firestore will give you a link to create the index when the query fails
```

## Pagination

```typescript
// Cursor-based pagination (Firestore way)
async function getPosts(pageSize: number, lastDoc?: any) {
  let q = query(
    collection(db, "posts"),
    orderBy("createdAt", "desc"),
    limit(pageSize)
  );

  if (lastDoc) {
    q = query(q, startAfter(lastDoc));
  }

  const snapshot = await getDocs(q);
  const posts = snapshot.docs.map(doc => ({ id: doc.id, ...doc.data() }));
  const lastVisible = snapshot.docs[snapshot.docs.length - 1];

  return {
    posts,
    lastVisible, // pass this to next call
    hasMore: snapshot.docs.length === pageSize,
  };
}

// Usage
const page1 = await getPosts(10);
const page2 = await getPosts(10, page1.lastVisible);
```

## Real-Time Listeners

```typescript
// Listen to a document
const unsubscribe = onSnapshot(doc(db, "posts", postId), (doc) => {
  if (doc.exists()) {
    console.log("Updated:", doc.data());
  }
});

// Listen to a query
const q = query(
  collection(db, "messages"),
  where("roomId", "==", roomId),
  orderBy("createdAt", "desc"),
  limit(50)
);

const unsubscribe = onSnapshot(q, (snapshot) => {
  snapshot.docChanges().forEach((change) => {
    if (change.type === "added") {
      console.log("New message:", change.doc.data());
    }
    if (change.type === "modified") {
      console.log("Modified:", change.doc.data());
    }
    if (change.type === "removed") {
      console.log("Removed:", change.doc.id);
    }
  });
});

// React hook
function useCollection<T>(q: Query) {
  const [data, setData] = useState<T[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    const unsubscribe = onSnapshot(
      q,
      (snapshot) => {
        const items = snapshot.docs.map(doc => ({
          id: doc.id,
          ...doc.data(),
        })) as T[];
        setData(items);
        setLoading(false);
      },
      (err) => {
        setError(err);
        setLoading(false);
      }
    );
    return unsubscribe;
  }, []);

  return { data, loading, error };
}

// Always unsubscribe when component unmounts
// The useEffect cleanup handles this automatically
```

## Transactions & Batch Writes

```typescript
// Transaction — read then write atomically
await runTransaction(db, async (transaction) => {
  const postRef = doc(db, "posts", postId);
  const postDoc = await transaction.get(postRef);

  if (!postDoc.exists()) throw new Error("Post not found");

  const newLikeCount = postDoc.data().likeCount + 1;
  transaction.update(postRef, { likeCount: newLikeCount });
});

// Batch write — multiple writes atomically (up to 500)
const batch = writeBatch(db);

// Add multiple documents
for (const item of items) {
  const ref = doc(collection(db, "items"));
  batch.set(ref, {
    ...item,
    createdAt: serverTimestamp(),
  });
}

// Mix operations
batch.update(doc(db, "counters", "items"), {
  total: increment(items.length),
});
batch.delete(doc(db, "temp", "import-status"));

await batch.commit();
```

## Subcollections

```typescript
// Posts > Comments subcollection
const commentsRef = collection(db, "posts", postId, "comments");

// Add a comment
await addDoc(commentsRef, {
  text: "Great post!",
  authorId: auth.currentUser!.uid,
  authorName: "Alice",
  createdAt: serverTimestamp(),
});

// Query comments for a post
const q = query(commentsRef, orderBy("createdAt", "asc"));
const snapshot = await getDocs(q);

// Collection group query — query ALL comments across ALL posts
const allComments = query(
  collectionGroup(db, "comments"),
  where("authorId", "==", userId),
  orderBy("createdAt", "desc"),
  limit(20)
);
// Requires a collection group index in Firebase Console
```

## Data Modeling Patterns

```typescript
// Pattern 1: User profile with subcollections
// users/{userId}                    — profile data
// users/{userId}/posts/{postId}     — user's posts
// users/{userId}/settings/{key}     — user settings

// Pattern 2: Chat application
// rooms/{roomId}                    — room metadata, participants
// rooms/{roomId}/messages/{msgId}   — messages (subcollection)
// users/{userId}/rooms/{roomId}     — user's room list (for sidebar query)

// Pattern 3: E-commerce
// products/{productId}              — product details
// orders/{orderId}                  — order with denormalized product info
// users/{userId}/cart/{itemId}      — shopping cart
// users/{userId}/orders/{orderId}   — user's order history (duplicate ref)

// Denormalization example — store author info in post
await addDoc(collection(db, "posts"), {
  title: "My Post",
  body: "Content...",
  // Denormalized author data (avoid extra read)
  author: {
    uid: user.uid,
    displayName: user.displayName,
    photoURL: user.photoURL,
  },
  createdAt: serverTimestamp(),
});
// Trade-off: author name change requires updating all posts
// Use Cloud Functions to propagate changes
```

## Security Rules

```
rules_version = '2';

service cloud.firestore {
  match /databases/{database}/documents {

    // Helper functions
    function isSignedIn() {
      return request.auth != null;
    }

    function isOwner(userId) {
      return request.auth.uid == userId;
    }

    function hasRole(role) {
      return get(/databases/$(database)/documents/users/$(request.auth.uid)).data.role == role;
    }

    function isValidPost() {
      let data = request.resource.data;
      return data.keys().hasAll(['title', 'body', 'authorId'])
        && data.title is string
        && data.title.size() > 0
        && data.title.size() <= 200
        && data.body is string
        && data.authorId == request.auth.uid;
    }

    // Users — owner can read/write own profile
    match /users/{userId} {
      allow read: if isSignedIn();
      allow create: if isOwner(userId);
      allow update: if isOwner(userId);
      allow delete: if isOwner(userId) || hasRole('admin');
    }

    // Posts — public read, auth create, owner update/delete
    match /posts/{postId} {
      allow read: if true;
      allow create: if isSignedIn() && isValidPost();
      allow update: if isOwner(resource.data.authorId);
      allow delete: if isOwner(resource.data.authorId) || hasRole('admin');

      // Comments subcollection
      match /comments/{commentId} {
        allow read: if true;
        allow create: if isSignedIn()
          && request.resource.data.authorId == request.auth.uid
          && request.resource.data.text is string
          && request.resource.data.text.size() > 0;
        allow update: if isOwner(resource.data.authorId);
        allow delete: if isOwner(resource.data.authorId)
          || isOwner(get(/databases/$(database)/documents/posts/$(postId)).data.authorId);
      }
    }

    // Multi-tenant — organization-scoped data
    match /orgs/{orgId} {
      allow read: if request.auth.uid in resource.data.members;

      match /projects/{projectId} {
        allow read, write: if request.auth.uid in
          get(/databases/$(database)/documents/orgs/$(orgId)).data.members;
      }
    }

    // Rate limiting pattern
    match /submissions/{submissionId} {
      allow create: if isSignedIn()
        && (!exists(/databases/$(database)/documents/rateLimits/$(request.auth.uid))
            || get(/databases/$(database)/documents/rateLimits/$(request.auth.uid)).data.lastSubmission
               < request.time - duration.value(60, 's'));
    }
  }
}
```

## Offline Support

```typescript
import {
  enableIndexedDbPersistence,
  enableMultiTabIndexedDbPersistence,
  CACHE_SIZE_UNLIMITED,
} from "firebase/firestore";

// Enable offline persistence (call before any Firestore operations)
try {
  await enableIndexedDbPersistence(db);
  // OR for multi-tab support:
  // await enableMultiTabIndexedDbPersistence(db);
} catch (err: any) {
  if (err.code === "failed-precondition") {
    // Multiple tabs open — only one can enable persistence
    console.warn("Persistence failed: multiple tabs open");
  } else if (err.code === "unimplemented") {
    // Browser doesn't support IndexedDB
    console.warn("Persistence not available");
  }
}

// Firestore handles offline reads automatically from cache
// Writes are queued and synced when back online

// Listen for online/offline status
import { enableNetwork, disableNetwork } from "firebase/firestore";

// Force offline (testing)
await disableNetwork(db);

// Come back online
await enableNetwork(db);
```

## Admin SDK (Server-Side)

```typescript
import { initializeApp, cert } from "firebase-admin/app";
import { getFirestore, FieldValue } from "firebase-admin/firestore";

const admin = initializeApp({
  credential: cert(JSON.parse(process.env.FIREBASE_SERVICE_ACCOUNT!)),
});
const adminDb = getFirestore(admin);

// Admin SDK bypasses security rules
const doc = await adminDb.collection("users").doc(uid).get();

// Bulk operations
const batch = adminDb.batch();
const snapshot = await adminDb.collection("users")
  .where("status", "==", "inactive")
  .get();

snapshot.docs.forEach(doc => {
  batch.delete(doc.ref);
});
await batch.commit();

// Server timestamp
await adminDb.collection("events").add({
  type: "user_signup",
  timestamp: FieldValue.serverTimestamp(),
});
```

## Gotchas

1. **Queries can only filter on fields they order by** — if you `orderBy("createdAt")`, you can only use inequality filters (`>`, `<`) on `createdAt`, not other fields. Need multiple inequality filters? Create a composite index.

2. **`in` and `array-contains-any` are limited to 30 values** — for larger sets, break into multiple queries and merge results client-side.

3. **Document size limit is 1MB** — if you're storing arrays or maps that grow, use subcollections instead. An array of 10K items will hit performance issues even before 1MB.

4. **`serverTimestamp()` is `null` until written** — in optimistic UI with `onSnapshot`, the timestamp field will be `null` on the first local snapshot. Handle this: `doc.data().createdAt?.toDate() || new Date()`.

5. **Deleting a document does NOT delete its subcollections** — you must delete subcollection documents explicitly, either client-side or via Cloud Functions.

6. **Reads are charged per document** — a query that returns 100 documents costs 100 reads. Use `limit()` and pagination. A listener that fires 50 times costs 50 reads.
