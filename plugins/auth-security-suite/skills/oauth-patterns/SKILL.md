---
name: oauth-patterns
description: >
  OAuth 2.0 and social login implementation — Google, GitHub, Apple sign-in,
  PKCE flow, token exchange, and account linking patterns.
  Triggers: "OAuth", "social login", "Google sign in", "GitHub login", "Apple login",
  "PKCE", "authorization code flow", "account linking", "passport.js".
  NOT for: JWT implementation details (use jwt-auth), general security (use security-hardening).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# OAuth 2.0 Patterns

## OAuth 2.0 Authorization Code Flow (with PKCE)

```typescript
// 1. Generate PKCE challenge
import { randomBytes, createHash } from 'crypto';

function generatePKCE() {
  const verifier = randomBytes(32).toString('base64url');
  const challenge = createHash('sha256').update(verifier).digest('base64url');
  return { verifier, challenge };
}

// 2. Build authorization URL
function getAuthURL(provider: 'google' | 'github') {
  const { verifier, challenge } = generatePKCE();
  const state = randomBytes(16).toString('hex');

  // Store verifier + state in session
  // req.session.oauthState = { verifier, state, provider };

  const configs = {
    google: {
      authUrl: 'https://accounts.google.com/o/oauth2/v2/auth',
      clientId: process.env.GOOGLE_CLIENT_ID!,
      scopes: 'openid email profile',
    },
    github: {
      authUrl: 'https://github.com/login/oauth/authorize',
      clientId: process.env.GITHUB_CLIENT_ID!,
      scopes: 'user:email',
    },
  };

  const config = configs[provider];
  const params = new URLSearchParams({
    client_id: config.clientId,
    redirect_uri: `${process.env.APP_URL}/api/auth/callback/${provider}`,
    response_type: 'code',
    scope: config.scopes,
    state,
    code_challenge: challenge,
    code_challenge_method: 'S256',
  });

  return `${config.authUrl}?${params}`;
}
```

## Google OAuth Implementation

```typescript
import { Router } from 'express';

const router = Router();

// Redirect to Google
router.get('/auth/google', (req, res) => {
  const { verifier, challenge } = generatePKCE();
  const state = randomBytes(16).toString('hex');
  req.session.oauthState = { verifier, state };

  const params = new URLSearchParams({
    client_id: process.env.GOOGLE_CLIENT_ID!,
    redirect_uri: `${process.env.APP_URL}/api/auth/callback/google`,
    response_type: 'code',
    scope: 'openid email profile',
    state,
    code_challenge: challenge,
    code_challenge_method: 'S256',
    access_type: 'offline', // Get refresh token
    prompt: 'consent',       // Force consent screen (for refresh token)
  });

  res.redirect(`https://accounts.google.com/o/oauth2/v2/auth?${params}`);
});

// Handle callback
router.get('/auth/callback/google', async (req, res) => {
  const { code, state } = req.query as { code: string; state: string };
  const stored = req.session.oauthState;

  // Verify state
  if (!stored || state !== stored.state) {
    return res.status(400).json({ error: 'Invalid state parameter' });
  }

  try {
    // Exchange code for tokens
    const tokenRes = await fetch('https://oauth2.googleapis.com/token', {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: new URLSearchParams({
        code,
        client_id: process.env.GOOGLE_CLIENT_ID!,
        client_secret: process.env.GOOGLE_CLIENT_SECRET!,
        redirect_uri: `${process.env.APP_URL}/api/auth/callback/google`,
        grant_type: 'authorization_code',
        code_verifier: stored.verifier,
      }),
    });

    const tokens = await tokenRes.json();
    if (tokens.error) throw new Error(tokens.error_description);

    // Get user info
    const userRes = await fetch('https://www.googleapis.com/oauth2/v3/userinfo', {
      headers: { Authorization: `Bearer ${tokens.access_token}` },
    });
    const profile = await userRes.json();

    // Find or create user
    const user = await findOrCreateOAuthUser({
      provider: 'google',
      providerId: profile.sub,
      email: profile.email,
      name: profile.name,
      avatar: profile.picture,
    });

    // Set auth cookies and redirect
    const accessToken = createAccessToken(user);
    const refreshToken = createRefreshToken(user);
    setAuthCookies(res, accessToken, refreshToken);

    delete req.session.oauthState;
    res.redirect('/dashboard');
  } catch (err) {
    console.error('Google OAuth error:', err);
    res.redirect('/login?error=oauth_failed');
  }
});
```

## GitHub OAuth Implementation

```typescript
router.get('/auth/github', (req, res) => {
  const state = randomBytes(16).toString('hex');
  req.session.oauthState = { state };

  const params = new URLSearchParams({
    client_id: process.env.GITHUB_CLIENT_ID!,
    redirect_uri: `${process.env.APP_URL}/api/auth/callback/github`,
    scope: 'user:email',
    state,
  });

  res.redirect(`https://github.com/login/oauth/authorize?${params}`);
});

router.get('/auth/callback/github', async (req, res) => {
  const { code, state } = req.query as { code: string; state: string };

  if (state !== req.session.oauthState?.state) {
    return res.status(400).json({ error: 'Invalid state' });
  }

  try {
    // Exchange code for token
    const tokenRes = await fetch('https://github.com/login/oauth/access_token', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Accept: 'application/json',
      },
      body: JSON.stringify({
        client_id: process.env.GITHUB_CLIENT_ID!,
        client_secret: process.env.GITHUB_CLIENT_SECRET!,
        code,
        redirect_uri: `${process.env.APP_URL}/api/auth/callback/github`,
      }),
    });

    const { access_token } = await tokenRes.json();

    // Get user profile
    const userRes = await fetch('https://api.github.com/user', {
      headers: { Authorization: `Bearer ${access_token}` },
    });
    const profile = await userRes.json();

    // Get primary email (may be private)
    const emailRes = await fetch('https://api.github.com/user/emails', {
      headers: { Authorization: `Bearer ${access_token}` },
    });
    const emails = await emailRes.json();
    const primaryEmail = emails.find((e: any) => e.primary)?.email;

    const user = await findOrCreateOAuthUser({
      provider: 'github',
      providerId: String(profile.id),
      email: primaryEmail,
      name: profile.name || profile.login,
      avatar: profile.avatar_url,
    });

    const accessToken = createAccessToken(user);
    const refreshToken = createRefreshToken(user);
    setAuthCookies(res, accessToken, refreshToken);

    delete req.session.oauthState;
    res.redirect('/dashboard');
  } catch (err) {
    console.error('GitHub OAuth error:', err);
    res.redirect('/login?error=oauth_failed');
  }
});
```

## Account Linking & User Resolution

```typescript
// Find existing user or create new one
// Handles: same email different provider, account linking
async function findOrCreateOAuthUser(profile: {
  provider: string;
  providerId: string;
  email: string;
  name: string;
  avatar?: string;
}) {
  // 1. Check if this OAuth account already linked
  const existingLink = await db.oauthAccounts.findFirst({
    where: { provider: profile.provider, providerId: profile.providerId },
    include: { user: true },
  });

  if (existingLink) {
    return existingLink.user;
  }

  // 2. Check if user exists with same email
  const existingUser = await db.users.findByEmail(profile.email);

  if (existingUser) {
    // Link this OAuth provider to existing account
    await db.oauthAccounts.create({
      userId: existingUser.id,
      provider: profile.provider,
      providerId: profile.providerId,
    });
    return existingUser;
  }

  // 3. Create new user + OAuth link
  const newUser = await db.users.create({
    email: profile.email,
    name: profile.name,
    avatar: profile.avatar,
    role: 'user',
  });

  await db.oauthAccounts.create({
    userId: newUser.id,
    provider: profile.provider,
    providerId: profile.providerId,
  });

  return newUser;
}

// Database schema for OAuth accounts
// oauth_accounts:
//   id, userId, provider, providerId, createdAt
//   UNIQUE(provider, providerId)
```

## Social Login Buttons (React)

```typescript
function LoginPage() {
  return (
    <div className="flex flex-col gap-3">
      <a
        href="/api/auth/google"
        className="flex items-center justify-center gap-2 rounded-lg border border-gray-300 bg-white px-4 py-2.5 text-sm font-medium text-gray-700 shadow-sm hover:bg-gray-50"
      >
        <GoogleIcon className="h-5 w-5" />
        Continue with Google
      </a>

      <a
        href="/api/auth/github"
        className="flex items-center justify-center gap-2 rounded-lg bg-gray-900 px-4 py-2.5 text-sm font-medium text-white shadow-sm hover:bg-gray-800"
      >
        <GitHubIcon className="h-5 w-5" />
        Continue with GitHub
      </a>

      <div className="relative my-2">
        <div className="absolute inset-0 flex items-center">
          <div className="w-full border-t border-gray-300" />
        </div>
        <div className="relative flex justify-center">
          <span className="bg-white px-2 text-sm text-gray-500">or</span>
        </div>
      </div>

      <EmailPasswordForm />
    </div>
  );
}
```

## Environment Variables Template

```env
# Google OAuth
GOOGLE_CLIENT_ID=your-client-id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=GOCSPX-your-secret

# GitHub OAuth
GITHUB_CLIENT_ID=your-github-client-id
GITHUB_CLIENT_SECRET=your-github-client-secret

# App
APP_URL=http://localhost:3000
```

## Gotchas

1. **Always validate the `state` parameter.** Without state validation, your callback is vulnerable to CSRF attacks. Generate a random state, store it in the session, verify it matches on callback.

2. **GitHub emails can be private.** The `/user` endpoint may return `email: null`. You must call `/user/emails` separately and find the `primary` email. Handle the case where no verified email exists.

3. **Google requires `prompt: 'consent'` for refresh tokens.** Without it, Google only returns a refresh token on the first authorization. If the user already authorized your app, you won't get one.

4. **Account linking by email has security implications.** If you auto-link accounts by email, an attacker who controls a GitHub account with someone else's email can access their Google-linked account. Require email verification or explicit user confirmation for linking.

5. **PKCE is required for public clients (SPAs).** Even for server-side apps, PKCE is recommended as defense-in-depth. Always use `code_challenge_method: 'S256'` (SHA-256), never `plain`.

6. **OAuth redirect URIs must be exact matches.** `http://localhost:3000/callback` and `http://localhost:3000/callback/` are different. Register the exact URI in the provider's console. Trailing slashes matter.
