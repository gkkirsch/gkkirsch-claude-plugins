---
name: firebase-auth
description: >
  Firebase Authentication — email/password, OAuth providers, phone auth,
  custom tokens, multi-factor auth, session management, and Auth state.
  Triggers: "firebase auth", "firebase login", "firebase signup",
  "firebase oauth", "firebase phone auth", "firebase mfa",
  "firebase user management", "firebase custom token".
  NOT for: Firestore data or Cloud Functions (use firebase-firestore, firebase-hosting-functions).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Firebase Authentication

## Setup

```bash
npm install firebase
```

```typescript
// lib/firebase.ts
import { initializeApp } from "firebase/app";
import { getAuth } from "firebase/auth";

const firebaseConfig = {
  apiKey: process.env.NEXT_PUBLIC_FIREBASE_API_KEY,
  authDomain: process.env.NEXT_PUBLIC_FIREBASE_AUTH_DOMAIN,
  projectId: process.env.NEXT_PUBLIC_FIREBASE_PROJECT_ID,
  storageBucket: process.env.NEXT_PUBLIC_FIREBASE_STORAGE_BUCKET,
  messagingSenderId: process.env.NEXT_PUBLIC_FIREBASE_MESSAGING_SENDER_ID,
  appId: process.env.NEXT_PUBLIC_FIREBASE_APP_ID,
};

const app = initializeApp(firebaseConfig);
export const auth = getAuth(app);
```

## Email/Password Auth

```typescript
import {
  createUserWithEmailAndPassword,
  signInWithEmailAndPassword,
  signOut,
  sendPasswordResetEmail,
  sendEmailVerification,
  updateProfile,
} from "firebase/auth";
import { auth } from "./firebase";

// Sign up
async function signUp(email: string, password: string, displayName: string) {
  const credential = await createUserWithEmailAndPassword(auth, email, password);
  await updateProfile(credential.user, { displayName });
  await sendEmailVerification(credential.user);
  return credential.user;
}

// Sign in
async function signIn(email: string, password: string) {
  const credential = await signInWithEmailAndPassword(auth, email, password);
  return credential.user;
}

// Sign out
async function logOut() {
  await signOut(auth);
}

// Password reset
async function resetPassword(email: string) {
  await sendPasswordResetEmail(auth, email);
}
```

## OAuth Providers

```typescript
import {
  signInWithPopup,
  signInWithRedirect,
  getRedirectResult,
  GoogleAuthProvider,
  GithubAuthProvider,
  OAuthProvider,
  linkWithPopup,
} from "firebase/auth";

// Google
const googleProvider = new GoogleAuthProvider();
googleProvider.addScope("profile");
googleProvider.addScope("email");

async function signInWithGoogle() {
  const result = await signInWithPopup(auth, googleProvider);
  const credential = GoogleAuthProvider.credentialFromResult(result);
  const token = credential?.accessToken; // Google access token
  return result.user;
}

// GitHub
const githubProvider = new GithubAuthProvider();
githubProvider.addScope("repo");

async function signInWithGitHub() {
  const result = await signInWithPopup(auth, githubProvider);
  return result.user;
}

// Apple
const appleProvider = new OAuthProvider("apple.com");
appleProvider.addScope("email");
appleProvider.addScope("name");

async function signInWithApple() {
  const result = await signInWithPopup(auth, appleProvider);
  return result.user;
}

// Redirect-based (mobile-friendly)
async function signInWithGoogleRedirect() {
  await signInWithRedirect(auth, googleProvider);
}

// Handle redirect result (call on page load)
async function handleRedirectResult() {
  const result = await getRedirectResult(auth);
  if (result) {
    return result.user;
  }
  return null;
}

// Link additional providers to existing account
async function linkGoogle() {
  if (!auth.currentUser) throw new Error("Not signed in");
  const result = await linkWithPopup(auth.currentUser, googleProvider);
  return result.user;
}
```

## Auth State Management (React)

```typescript
import { onAuthStateChanged, User } from "firebase/auth";
import { createContext, useContext, useEffect, useState } from "react";

interface AuthContextType {
  user: User | null;
  loading: boolean;
}

const AuthContext = createContext<AuthContextType>({
  user: null,
  loading: true,
});

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const unsubscribe = onAuthStateChanged(auth, (user) => {
      setUser(user);
      setLoading(false);
    });
    return unsubscribe;
  }, []);

  return (
    <AuthContext.Provider value={{ user, loading }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  return useContext(AuthContext);
}

// Usage
function ProtectedPage() {
  const { user, loading } = useAuth();

  if (loading) return <div>Loading...</div>;
  if (!user) return <Navigate to="/login" />;

  return <div>Welcome, {user.displayName}</div>;
}
```

## Phone Auth

```typescript
import {
  RecaptchaVerifier,
  signInWithPhoneNumber,
  PhoneAuthProvider,
} from "firebase/auth";

// Set up reCAPTCHA (call once on component mount)
function setupRecaptcha() {
  window.recaptchaVerifier = new RecaptchaVerifier(auth, "recaptcha-container", {
    size: "normal",
    callback: () => {
      // reCAPTCHA solved — allow signInWithPhoneNumber
    },
  });
}

// Send verification code
async function sendCode(phoneNumber: string) {
  const appVerifier = window.recaptchaVerifier;
  const confirmationResult = await signInWithPhoneNumber(auth, phoneNumber, appVerifier);
  // Store confirmationResult for verification step
  return confirmationResult;
}

// Verify code
async function verifyCode(confirmationResult: any, code: string) {
  const result = await confirmationResult.confirm(code);
  return result.user;
}
```

## Custom Claims & Admin SDK

```typescript
// SERVER SIDE — Firebase Admin SDK
import { initializeApp, cert } from "firebase-admin/app";
import { getAuth } from "firebase-admin/auth";

const admin = initializeApp({
  credential: cert(JSON.parse(process.env.FIREBASE_SERVICE_ACCOUNT!)),
});
const adminAuth = getAuth(admin);

// Set custom claims (roles)
async function setUserRole(uid: string, role: string) {
  await adminAuth.setCustomUserClaims(uid, { role });
}

// Verify ID token (API routes)
async function verifyToken(idToken: string) {
  const decoded = await adminAuth.verifyIdToken(idToken);
  return decoded; // { uid, email, role, ... }
}

// Create custom token (for custom auth systems)
async function createCustomToken(uid: string, claims?: object) {
  const token = await adminAuth.createCustomToken(uid, claims);
  return token;
}

// List users
async function listUsers(maxResults = 100) {
  const result = await adminAuth.listUsers(maxResults);
  return result.users;
}

// Delete user
async function deleteUser(uid: string) {
  await adminAuth.deleteUser(uid);
}
```

## ID Token for API Calls

```typescript
// CLIENT: Get ID token to send to your backend
async function getIdToken() {
  if (!auth.currentUser) throw new Error("Not signed in");
  return auth.currentUser.getIdToken();
}

// API call with auth
async function fetchProtected(url: string) {
  const token = await getIdToken();
  const response = await fetch(url, {
    headers: { Authorization: `Bearer ${token}` },
  });
  return response.json();
}

// SERVER: Verify in Express middleware
import { getAuth } from "firebase-admin/auth";

async function authMiddleware(req, res, next) {
  const authHeader = req.headers.authorization;
  if (!authHeader?.startsWith("Bearer ")) {
    return res.status(401).json({ error: "Unauthorized" });
  }

  try {
    const token = authHeader.split("Bearer ")[1];
    const decoded = await getAuth().verifyIdToken(token);
    req.user = decoded;
    next();
  } catch (error) {
    res.status(401).json({ error: "Invalid token" });
  }
}

// Role-based middleware
function requireRole(role: string) {
  return (req, res, next) => {
    if (req.user?.role !== role) {
      return res.status(403).json({ error: "Forbidden" });
    }
    next();
  };
}
```

## Multi-Factor Authentication (MFA)

```typescript
import {
  multiFactor,
  PhoneMultiFactorGenerator,
  PhoneAuthProvider,
  getMultiFactorResolver,
} from "firebase/auth";

// Enroll MFA
async function enrollMFA(phoneNumber: string) {
  const user = auth.currentUser!;
  const session = await multiFactor(user).getSession();

  const phoneInfoOptions = {
    phoneNumber,
    session,
  };

  const phoneAuthProvider = new PhoneAuthProvider(auth);
  const verificationId = await phoneAuthProvider.verifyPhoneNumber(
    phoneInfoOptions,
    window.recaptchaVerifier
  );
  return verificationId;
}

async function confirmMFAEnrollment(verificationId: string, code: string) {
  const credential = PhoneAuthProvider.credential(verificationId, code);
  const assertion = PhoneMultiFactorGenerator.assertion(credential);
  await multiFactor(auth.currentUser!).enroll(assertion, "Phone Number");
}

// Handle MFA challenge during sign-in
async function handleMFASignIn(error: any) {
  if (error.code !== "auth/multi-factor-auth-required") throw error;

  const resolver = getMultiFactorResolver(auth, error);
  const phoneInfoOptions = {
    multiFactorHint: resolver.hints[0],
    session: resolver.session,
  };

  const phoneAuthProvider = new PhoneAuthProvider(auth);
  const verificationId = await phoneAuthProvider.verifyPhoneNumber(
    phoneInfoOptions,
    window.recaptchaVerifier
  );
  return { resolver, verificationId };
}

async function completeMFASignIn(
  resolver: any,
  verificationId: string,
  code: string
) {
  const credential = PhoneAuthProvider.credential(verificationId, code);
  const assertion = PhoneMultiFactorGenerator.assertion(credential);
  await resolver.resolveSignIn(assertion);
}
```

## Gotchas

1. **`onAuthStateChanged` fires on every page load** — the initial call with `null` doesn't mean the user is signed out. Wait for `loading` to be `false` before making auth decisions.

2. **ID tokens expire after 1 hour** — use `getIdToken(true)` to force refresh, or rely on the SDK's auto-refresh. Don't cache tokens client-side for longer than a request.

3. **Custom claims propagate on token refresh** — after setting custom claims with Admin SDK, the client won't see them until the token refreshes (up to 1 hour). Force refresh with `getIdToken(true)`.

4. **`signInWithRedirect` needs `getRedirectResult`** — call `getRedirectResult` on page load to handle the redirect callback. Missing this means silent auth failures on mobile.

5. **Phone auth requires reCAPTCHA** — even for testing. Use `auth.settings.appVerificationDisabledForTesting = true` in dev, but NEVER in production.

6. **Firebase API key is NOT secret** — it's safe to expose in client code. Security comes from Firestore rules and Auth, not API key secrecy. The key just identifies your project.
