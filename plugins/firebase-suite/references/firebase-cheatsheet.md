# Firebase Cheatsheet

## Auth

```typescript
import { getAuth, createUserWithEmailAndPassword, signInWithEmailAndPassword,
  signOut, signInWithPopup, GoogleAuthProvider, onAuthStateChanged } from "firebase/auth"

const auth = getAuth()

// Email/password
await createUserWithEmailAndPassword(auth, email, password)
await signInWithEmailAndPassword(auth, email, password)
await signOut(auth)

// Google OAuth
await signInWithPopup(auth, new GoogleAuthProvider())

// Auth state listener
onAuthStateChanged(auth, user => { /* user or null */ })

// Get ID token for API
const token = await auth.currentUser!.getIdToken()
```

## Firestore — CRUD

```typescript
import { getFirestore, collection, doc, addDoc, setDoc, getDoc, getDocs,
  updateDoc, deleteDoc, serverTimestamp, increment, arrayUnion } from "firebase/firestore"

const db = getFirestore()

// Create (auto ID)
await addDoc(collection(db, "posts"), { title: "Hello", createdAt: serverTimestamp() })

// Create (specific ID)
await setDoc(doc(db, "users", uid), { name: "Alice" })

// Read one
const snap = await getDoc(doc(db, "posts", id))
if (snap.exists()) snap.data()

// Read collection
const snaps = await getDocs(collection(db, "posts"))
snaps.docs.map(d => ({ id: d.id, ...d.data() }))

// Update
await updateDoc(doc(db, "posts", id), { title: "Updated", updatedAt: serverTimestamp() })

// Atomic
await updateDoc(doc(db, "posts", id), { views: increment(1), tags: arrayUnion("new") })

// Delete
await deleteDoc(doc(db, "posts", id))
```

## Firestore — Queries

```typescript
import { query, where, orderBy, limit, startAfter } from "firebase/firestore"

const q = query(
  collection(db, "posts"),
  where("status", "==", "published"),
  where("category", "in", ["tech", "science"]),
  orderBy("createdAt", "desc"),
  limit(10)
)
const snap = await getDocs(q)

// Operators: ==, !=, <, <=, >, >=, in, not-in, array-contains, array-contains-any

// Pagination
const next = query(q, startAfter(lastDoc))
```

## Firestore — Real-Time

```typescript
import { onSnapshot } from "firebase/firestore"

// Document
const unsub = onSnapshot(doc(db, "posts", id), snap => {
  console.log(snap.data())
})

// Query
const unsub = onSnapshot(q, snap => {
  snap.docChanges().forEach(change => {
    // change.type: "added" | "modified" | "removed"
  })
})

unsub() // cleanup
```

## Firestore — Batch & Transaction

```typescript
import { writeBatch, runTransaction } from "firebase/firestore"

// Batch (up to 500 ops)
const batch = writeBatch(db)
batch.set(doc(db, "posts", "1"), { title: "A" })
batch.update(doc(db, "counters", "posts"), { total: increment(1) })
batch.delete(doc(db, "old", "1"))
await batch.commit()

// Transaction (read-then-write)
await runTransaction(db, async tx => {
  const snap = await tx.get(doc(db, "posts", id))
  tx.update(doc(db, "posts", id), { likes: snap.data()!.likes + 1 })
})
```

## Cloud Functions

```typescript
// HTTP
export const api = onRequest({ cors: true }, async (req, res) => {
  res.json({ ok: true })
})

// Callable (with auth)
export const doThing = onCall(async (request) => {
  if (!request.auth) throw new HttpsError("unauthenticated", "Sign in")
  return { result: "done" }
})

// Firestore trigger
export const onPost = onDocumentCreated("posts/{id}", async (event) => {
  const data = event.data?.data()
})

export const onUpdate = onDocumentUpdated("posts/{id}", async (event) => {
  const before = event.data?.before.data()
  const after = event.data?.after.data()
})

// Scheduled
export const cleanup = onSchedule("every 24 hours", async () => { /* ... */ })

// Storage trigger
export const onUpload = onObjectFinalized(async (event) => {
  const path = event.data.name // "uploads/photo.jpg"
})
```

## Client Callable

```typescript
import { getFunctions, httpsCallable } from "firebase/functions"
const functions = getFunctions()
const doThing = httpsCallable(functions, "doThing")
const result = await doThing({ input: "data" })
```

## Storage

```typescript
import { getStorage, ref, uploadBytes, getDownloadURL, deleteObject } from "firebase/storage"

const storage = getStorage()

// Upload
const snap = await uploadBytes(ref(storage, "path/file.jpg"), file)
const url = await getDownloadURL(snap.ref)

// Delete
await deleteObject(ref(storage, "path/file.jpg"))
```

## Security Rules — Firestore

```
match /posts/{postId} {
  allow read: if true;
  allow create: if request.auth != null
    && request.resource.data.authorId == request.auth.uid;
  allow update, delete: if request.auth.uid == resource.data.authorId;
}
```

## Security Rules — Storage

```
match /uploads/{userId}/{file} {
  allow read: if request.auth != null;
  allow write: if request.auth.uid == userId
    && request.resource.size < 10 * 1024 * 1024;
}
```

## Admin SDK (Server)

```typescript
import { initializeApp, cert } from "firebase-admin/app"
import { getFirestore } from "firebase-admin/firestore"
import { getAuth } from "firebase-admin/auth"

initializeApp({ credential: cert(JSON.parse(process.env.FIREBASE_SA!)) })

// Firestore (bypasses rules)
const db = getFirestore()
await db.collection("users").doc(uid).set({ role: "admin" })

// Auth
await getAuth().setCustomUserClaims(uid, { role: "admin" })
const decoded = await getAuth().verifyIdToken(idToken)
```

## CLI

```bash
firebase init                        # project setup
firebase emulators:start             # local dev
firebase deploy                      # deploy all
firebase deploy --only functions     # functions only
firebase deploy --only hosting       # hosting only
firebase deploy --only firestore:rules
firebase hosting:channel:deploy preview  # preview URL
firebase functions:log               # view logs
```

## Emulator Connect

```typescript
if (process.env.NODE_ENV === "development") {
  connectAuthEmulator(auth, "http://localhost:9099")
  connectFirestoreEmulator(db, "localhost", 8080)
  connectFunctionsEmulator(functions, "localhost", 5001)
  connectStorageEmulator(storage, "localhost", 9199)
}
```

## Gotchas

1. Firestore charges per read/write — use `limit()`, batch reads
2. `serverTimestamp()` is null until written — handle in UI
3. Deleting a doc does NOT delete subcollections
4. `in`/`array-contains-any` limited to 30 values
5. Cloud Functions have cold starts (1-10s) — use `minInstances: 1` for critical paths
6. Firestore triggers can fire multiple times — make handlers idempotent
7. Firebase API key is NOT secret — security comes from rules
8. `onAuthStateChanged` fires with null initially — wait for loading state
