# OAuth2 Flows Reference

Complete reference for all OAuth2 authorization flows, including protocol diagrams, code examples, security considerations, and implementation guidance.

---

## Flow Selection Guide

```
┌─────────────────────────────────────────────────────────────────────┐
│                    OAuth2 Flow Selection Matrix                     │
├────────────────────┬──────────────────────┬─────────────────────────┤
│ Application Type   │ Recommended Flow     │ Reason                  │
├────────────────────┼──────────────────────┼─────────────────────────┤
│ Server-rendered    │ Authorization Code   │ Can store client secret │
│ web app            │                      │ securely                │
├────────────────────┼──────────────────────┼─────────────────────────┤
│ SPA (React, Vue,   │ Authorization Code   │ No client secret        │
│ Angular)           │ + PKCE               │ needed                  │
├────────────────────┼──────────────────────┼─────────────────────────┤
│ Native mobile app  │ Authorization Code   │ Use custom URL scheme   │
│ (iOS, Android)     │ + PKCE               │ for redirect            │
├────────────────────┼──────────────────────┼─────────────────────────┤
│ Desktop app        │ Authorization Code   │ Use loopback redirect   │
│ (Electron)         │ + PKCE               │ (127.0.0.1)             │
├────────────────────┼──────────────────────┼─────────────────────────┤
│ CLI tool           │ Device Authorization │ No browser redirect     │
│                    │ Grant                │ possible                │
├────────────────────┼──────────────────────┼─────────────────────────┤
│ IoT device         │ Device Authorization │ Limited input           │
│                    │ Grant                │ capabilities            │
├────────────────────┼──────────────────────┼─────────────────────────┤
│ API-to-API         │ Client Credentials   │ No user context         │
│ (microservices)    │                      │                         │
├────────────────────┼──────────────────────┼─────────────────────────┤
│ Trusted first-     │ Authorization Code   │ ROPC is deprecated;     │
│ party app          │ + PKCE               │ use PKCE instead        │
├────────────────────┼──────────────────────┼─────────────────────────┤
│ NEVER use          │ Implicit Flow        │ Tokens exposed in URL   │
│                    │ (deprecated)         │ fragment — insecure     │
└────────────────────┴──────────────────────┴─────────────────────────┘
```

---

## 1. Authorization Code Flow

The most secure flow for applications with a backend server.

### Protocol Diagram

```
     ┌──────┐          ┌──────────┐          ┌────────────┐
     │      │          │          │          │            │
     │ User │          │  Client  │          │ Auth Server│
     │      │          │ (Server) │          │            │
     └──┬───┘          └────┬─────┘          └─────┬──────┘
        │                   │                      │
        │ 1. Click "Login"  │                      │
        │──────────────────>│                      │
        │                   │                      │
        │                   │ 2. Generate state    │
        │                   │    Store in session   │
        │                   │                      │
        │ 3. 302 Redirect   │                      │
        │<──────────────────│                      │
        │                   │                      │
        │ 4. GET /authorize?response_type=code     │
        │   &client_id=...&redirect_uri=...        │
        │   &scope=openid profile email            │
        │   &state=random123                       │
        │─────────────────────────────────────────>│
        │                   │                      │
        │ 5. Login page     │                      │
        │<─────────────────────────────────────────│
        │                   │                      │
        │ 6. Enter credentials                     │
        │─────────────────────────────────────────>│
        │                   │                      │
        │ 7. Consent screen │                      │
        │<─────────────────────────────────────────│
        │                   │                      │
        │ 8. Approve        │                      │
        │─────────────────────────────────────────>│
        │                   │                      │
        │ 9. 302 Redirect to redirect_uri          │
        │    ?code=AUTH_CODE&state=random123        │
        │<─────────────────────────────────────────│
        │                   │                      │
        │ 10. Follow redirect                      │
        │──────────────────>│                      │
        │                   │                      │
        │                   │ 11. Verify state     │
        │                   │                      │
        │                   │ 12. POST /token      │
        │                   │  grant_type=          │
        │                   │  authorization_code   │
        │                   │  code=AUTH_CODE       │
        │                   │  client_id=...        │
        │                   │  client_secret=...    │
        │                   │  redirect_uri=...     │
        │                   │─────────────────────>│
        │                   │                      │
        │                   │ 13. Validate code,   │
        │                   │  client credentials  │
        │                   │                      │
        │                   │ 14. Token Response   │
        │                   │ {                    │
        │                   │   access_token,      │
        │                   │   refresh_token,     │
        │                   │   id_token,          │
        │                   │   expires_in: 3600   │
        │                   │ }                    │
        │                   │<─────────────────────│
        │                   │                      │
        │ 15. Set session   │                      │
        │<──────────────────│                      │
        │                   │                      │
```

### Implementation (Node.js Express)

```javascript
// routes/auth.js — Authorization Code Flow
const express = require('express');
const crypto = require('crypto');
const router = express.Router();

const OAUTH_CONFIG = {
  authorizationUrl: 'https://auth.example.com/authorize',
  tokenUrl: 'https://auth.example.com/token',
  clientId: process.env.OAUTH_CLIENT_ID,
  clientSecret: process.env.OAUTH_CLIENT_SECRET,
  redirectUri: process.env.OAUTH_REDIRECT_URI,
  scopes: ['openid', 'profile', 'email'],
};

// Step 1: Initiate login
router.get('/login', (req, res) => {
  const state = crypto.randomBytes(32).toString('hex');
  req.session.oauthState = state;
  req.session.returnTo = req.query.returnTo || '/';

  const params = new URLSearchParams({
    response_type: 'code',
    client_id: OAUTH_CONFIG.clientId,
    redirect_uri: OAUTH_CONFIG.redirectUri,
    scope: OAUTH_CONFIG.scopes.join(' '),
    state,
  });

  res.redirect(`${OAUTH_CONFIG.authorizationUrl}?${params}`);
});

// Step 2: Handle callback
router.get('/callback', async (req, res) => {
  const { code, state, error } = req.query;

  if (error) {
    return res.redirect(`/login?error=${error}`);
  }

  // Validate state
  if (state !== req.session.oauthState) {
    return res.status(403).send('State mismatch');
  }
  delete req.session.oauthState;

  // Exchange code for tokens
  const tokenResponse = await fetch(OAUTH_CONFIG.tokenUrl, {
    method: 'POST',
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    body: new URLSearchParams({
      grant_type: 'authorization_code',
      code,
      client_id: OAUTH_CONFIG.clientId,
      client_secret: OAUTH_CONFIG.clientSecret,
      redirect_uri: OAUTH_CONFIG.redirectUri,
    }),
  });

  const tokens = await tokenResponse.json();

  if (tokens.error) {
    return res.redirect(`/login?error=${tokens.error}`);
  }

  // Store tokens in session
  req.session.accessToken = tokens.access_token;
  req.session.refreshToken = tokens.refresh_token;
  req.session.tokenExpiry = Date.now() + tokens.expires_in * 1000;

  const returnTo = req.session.returnTo || '/';
  delete req.session.returnTo;
  res.redirect(returnTo);
});
```

### Security Checklist

```
✅ Authorization Code Flow Security
   □ Generate cryptographically random state (32+ bytes)
   □ Validate state in callback (exact match)
   □ Use HTTPS for all redirect URIs
   □ Register exact redirect URIs (no wildcards)
   □ Store client_secret server-side only
   □ Exchange code immediately (codes expire fast)
   □ Use POST for token exchange (not GET)
   □ Validate token response before using
   □ Store tokens in server-side session, not cookies
   □ Add PKCE even for confidential clients (defense-in-depth)
```

---

## 2. Authorization Code Flow with PKCE

Required for public clients (SPAs, mobile apps). Recommended for all clients.

### PKCE Protocol

```
PKCE adds proof of possession to prevent authorization code interception:

1. Client generates random code_verifier (43-128 characters)
2. Client computes code_challenge = BASE64URL(SHA256(code_verifier))
3. Client sends code_challenge in authorization request
4. Auth server stores code_challenge
5. Client sends code_verifier in token request
6. Auth server verifies: SHA256(code_verifier) == stored code_challenge

This ensures:
- Only the client that initiated the flow can exchange the code
- Intercepted codes are useless without the code_verifier
- No client_secret required (safe for public clients)
```

### Flow Diagram

```
     ┌──────┐                              ┌────────────┐
     │      │                              │            │
     │ SPA  │                              │ Auth Server│
     │      │                              │            │
     └──┬───┘                              └─────┬──────┘
        │                                        │
        │ 1. Generate code_verifier              │
        │    (cryptographically random)          │
        │                                        │
        │ 2. code_challenge =                    │
        │    BASE64URL(SHA256(code_verifier))     │
        │                                        │
        │ 3. GET /authorize                      │
        │    ?response_type=code                 │
        │    &client_id=SPA_CLIENT_ID            │
        │    &redirect_uri=https://spa.com/cb    │
        │    &scope=openid profile               │
        │    &state=xyz                          │
        │    &code_challenge=CHALLENGE           │
        │    &code_challenge_method=S256         │
        │───────────────────────────────────────>│
        │                                        │
        │ 4. User authenticates + consents       │
        │                                        │
        │ 5. Redirect with code                  │
        │    ?code=AUTH_CODE&state=xyz            │
        │<───────────────────────────────────────│
        │                                        │
        │ 6. POST /token                         │
        │    grant_type=authorization_code        │
        │    code=AUTH_CODE                       │
        │    client_id=SPA_CLIENT_ID             │
        │    code_verifier=VERIFIER              │
        │    redirect_uri=https://spa.com/cb     │
        │    (NO client_secret)                  │
        │───────────────────────────────────────>│
        │                                        │
        │ 7. Server verifies:                    │
        │    BASE64URL(SHA256(VERIFIER))          │
        │    == stored CHALLENGE                  │
        │                                        │
        │ 8. Token Response                      │
        │    { access_token, id_token }          │
        │<───────────────────────────────────────│
        │                                        │
```

### Implementation (TypeScript SPA)

```typescript
// auth/pkce-flow.ts — Complete PKCE implementation for SPAs

function base64UrlEncode(buffer: ArrayBuffer): string {
  const bytes = new Uint8Array(buffer);
  let str = '';
  bytes.forEach(byte => str += String.fromCharCode(byte));
  return btoa(str).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
}

function generateCodeVerifier(): string {
  const array = new Uint8Array(64);
  crypto.getRandomValues(array);
  return base64UrlEncode(array.buffer);
}

async function generateCodeChallenge(verifier: string): Promise<string> {
  const encoder = new TextEncoder();
  const data = encoder.encode(verifier);
  const digest = await crypto.subtle.digest('SHA-256', data);
  return base64UrlEncode(digest);
}

export async function initiateLogin(returnTo?: string) {
  const codeVerifier = generateCodeVerifier();
  const codeChallenge = await generateCodeChallenge(codeVerifier);
  const state = crypto.randomUUID();

  // Store in sessionStorage (cleared on tab close)
  sessionStorage.setItem('pkce_verifier', codeVerifier);
  sessionStorage.setItem('pkce_state', state);
  if (returnTo) sessionStorage.setItem('pkce_return', returnTo);

  const params = new URLSearchParams({
    response_type: 'code',
    client_id: import.meta.env.VITE_OAUTH_CLIENT_ID,
    redirect_uri: `${window.location.origin}/auth/callback`,
    scope: 'openid profile email',
    state,
    code_challenge: codeChallenge,
    code_challenge_method: 'S256',
  });

  window.location.href = `${import.meta.env.VITE_AUTH_URL}/authorize?${params}`;
}

export async function handleCallback(): Promise<{ accessToken: string }> {
  const params = new URLSearchParams(window.location.search);

  // Validate state
  const storedState = sessionStorage.getItem('pkce_state');
  if (params.get('state') !== storedState) {
    throw new Error('State mismatch — possible CSRF attack');
  }

  const code = params.get('code');
  const codeVerifier = sessionStorage.getItem('pkce_verifier');

  // Clean up
  sessionStorage.removeItem('pkce_verifier');
  sessionStorage.removeItem('pkce_state');

  if (!code || !codeVerifier) {
    throw new Error('Missing code or verifier');
  }

  // Exchange code for tokens
  const response = await fetch(`${import.meta.env.VITE_AUTH_URL}/token`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    body: new URLSearchParams({
      grant_type: 'authorization_code',
      code,
      client_id: import.meta.env.VITE_OAUTH_CLIENT_ID,
      code_verifier: codeVerifier,
      redirect_uri: `${window.location.origin}/auth/callback`,
    }),
  });

  if (!response.ok) throw new Error('Token exchange failed');

  // Clean URL
  window.history.replaceState({}, '', window.location.pathname);

  return response.json();
}
```

### Python PKCE Implementation

```python
# auth/pkce.py — PKCE utilities
import secrets
import hashlib
import base64

def generate_code_verifier() -> str:
    """Generate a cryptographically random code verifier (43-128 chars)."""
    return secrets.token_urlsafe(64)

def generate_code_challenge(verifier: str) -> str:
    """Generate S256 code challenge from verifier."""
    digest = hashlib.sha256(verifier.encode("ascii")).digest()
    return base64.urlsafe_b64encode(digest).rstrip(b"=").decode("ascii")
```

---

## 3. Client Credentials Flow

For server-to-server (machine-to-machine) communication with no user context.

### Flow Diagram

```
     ┌──────────┐                        ┌────────────┐
     │          │                        │            │
     │  Client  │                        │ Auth Server│
     │ (Server) │                        │            │
     └────┬─────┘                        └─────┬──────┘
          │                                    │
          │ 1. POST /token                     │
          │    grant_type=client_credentials    │
          │    client_id=SERVICE_A_ID          │
          │    client_secret=SERVICE_A_SECRET  │
          │    scope=api:read api:write        │
          │───────────────────────────────────>│
          │                                    │
          │ 2. Validate credentials            │
          │    Check allowed scopes            │
          │                                    │
          │ 3. Token Response                  │
          │    {                               │
          │      access_token: "...",          │
          │      token_type: "Bearer",         │
          │      expires_in: 3600,             │
          │      scope: "api:read api:write"   │
          │    }                               │
          │    (NO refresh_token)              │
          │<───────────────────────────────────│
          │                                    │
          │ 4. Use access_token to call APIs   │
          │                                    │
```

### Implementation (Node.js)

```javascript
// auth/client-credentials.js — Service-to-service authentication
class ServiceAuth {
  constructor(config) {
    this.tokenUrl = config.tokenUrl;
    this.clientId = config.clientId;
    this.clientSecret = config.clientSecret;
    this.scopes = config.scopes || [];
    this._token = null;
    this._expiry = 0;
  }

  async getToken() {
    // Return cached token if still valid (with 60s buffer)
    if (this._token && this._expiry > Date.now() + 60000) {
      return this._token;
    }

    const response = await fetch(this.tokenUrl, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded',
        'Authorization': 'Basic ' + Buffer.from(
          `${this.clientId}:${this.clientSecret}`
        ).toString('base64'),
      },
      body: new URLSearchParams({
        grant_type: 'client_credentials',
        scope: this.scopes.join(' '),
      }),
    });

    if (!response.ok) {
      throw new Error(`Token request failed: ${response.status}`);
    }

    const data = await response.json();
    this._token = data.access_token;
    this._expiry = Date.now() + data.expires_in * 1000;

    return this._token;
  }

  // Make authenticated API call
  async fetch(url, options = {}) {
    const token = await this.getToken();
    return fetch(url, {
      ...options,
      headers: {
        ...options.headers,
        'Authorization': `Bearer ${token}`,
      },
    });
  }
}
```

### Python Implementation

```python
# auth/client_credentials.py
import time
import httpx

class ServiceAuth:
    def __init__(self, token_url: str, client_id: str, client_secret: str, scopes: list[str] = None):
        self.token_url = token_url
        self.client_id = client_id
        self.client_secret = client_secret
        self.scopes = scopes or []
        self._token = None
        self._expiry = 0

    async def get_token(self) -> str:
        if self._token and self._expiry > time.time() + 60:
            return self._token

        async with httpx.AsyncClient() as client:
            response = await client.post(
                self.token_url,
                data={
                    "grant_type": "client_credentials",
                    "scope": " ".join(self.scopes),
                },
                auth=(self.client_id, self.client_secret),
            )
            response.raise_for_status()
            data = response.json()

        self._token = data["access_token"]
        self._expiry = time.time() + data["expires_in"]
        return self._token
```

### Go Implementation

```go
// auth/client_credentials.go
package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type ServiceAuth struct {
	TokenURL     string
	ClientID     string
	ClientSecret string
	Scopes       []string
	token        string
	expiry       time.Time
	mu           sync.Mutex
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

func (s *ServiceAuth) GetToken() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.token != "" && time.Now().Add(60*time.Second).Before(s.expiry) {
		return s.token, nil
	}

	data := url.Values{
		"grant_type": {"client_credentials"},
		"scope":      {strings.Join(s.Scopes, " ")},
	}

	req, _ := http.NewRequest("POST", s.TokenURL, strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(s.ClientID, s.ClientSecret)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("token request returned %d", resp.StatusCode)
	}

	var tr tokenResponse
	json.NewDecoder(resp.Body).Decode(&tr)

	s.token = tr.AccessToken
	s.expiry = time.Now().Add(time.Duration(tr.ExpiresIn) * time.Second)

	return s.token, nil
}
```

---

## 4. Device Authorization Grant

For devices with limited input capability (CLIs, smart TVs, IoT).

### Flow Diagram

```
     ┌──────────┐    ┌──────┐        ┌────────────┐
     │  Device  │    │ User │        │ Auth Server│
     │ (CLI/TV) │    │      │        │            │
     └────┬─────┘    └──┬───┘        └─────┬──────┘
          │             │                  │
          │ 1. POST /device/code           │
          │    client_id=CLI_APP           │
          │    scope=openid profile        │
          │────────────────────────────────>│
          │             │                  │
          │ 2. Device Code Response        │
          │    {                           │
          │      device_code: "GmRh...",   │
          │      user_code: "WDJB-MJHT",  │
          │      verification_uri:         │
          │        "https://auth.com/device"│
          │      expires_in: 600,          │
          │      interval: 5               │
          │    }                           │
          │<────────────────────────────────│
          │             │                  │
          │ 3. Display to user:            │
          │    "Go to auth.com/device"     │
          │    "Enter code: WDJB-MJHT"     │
          │             │                  │
          │             │ 4. Navigate to   │
          │             │    verification  │
          │             │    URI           │
          │             │─────────────────>│
          │             │                  │
          │             │ 5. Enter code    │
          │             │ 6. Authenticate  │
          │             │ 7. Authorize     │
          │             │─────────────────>│
          │             │                  │
          │ 8. Poll: POST /token           │
          │    grant_type=                 │
          │    urn:ietf:params:oauth:       │
          │    grant-type:device_code      │
          │    device_code=GmRh...         │
          │    client_id=CLI_APP           │
          │────────────────────────────────>│
          │             │                  │
          │ 9a. "authorization_pending"    │
          │<────────────────────────────────│
          │             │                  │
          │ (wait interval, retry)         │
          │             │                  │
          │ 10. Poll again                 │
          │────────────────────────────────>│
          │             │                  │
          │ 11. Token Response             │
          │    { access_token, ... }       │
          │<────────────────────────────────│
          │             │                  │
```

### Implementation (Node.js CLI)

```javascript
// auth/device-flow.js — Device Authorization Grant for CLIs
const readline = require('readline');

class DeviceFlowAuth {
  constructor(config) {
    this.deviceCodeUrl = config.deviceCodeUrl;
    this.tokenUrl = config.tokenUrl;
    this.clientId = config.clientId;
    this.scopes = config.scopes || ['openid', 'profile'];
  }

  async login() {
    // Step 1: Request device code
    const deviceResponse = await fetch(this.deviceCodeUrl, {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: new URLSearchParams({
        client_id: this.clientId,
        scope: this.scopes.join(' '),
      }),
    });

    const deviceData = await deviceResponse.json();
    const {
      device_code,
      user_code,
      verification_uri,
      verification_uri_complete,
      expires_in,
      interval = 5,
    } = deviceData;

    // Step 2: Display instructions
    console.log('\n  To authenticate, visit:');
    console.log(`  ${verification_uri_complete || verification_uri}`);
    if (!verification_uri_complete) {
      console.log(`\n  And enter code: ${user_code}`);
    }
    console.log(`\n  Waiting for authorization...`);

    // Try to open browser automatically
    try {
      const open = require('open');
      await open(verification_uri_complete || verification_uri);
    } catch {}

    // Step 3: Poll for token
    const startTime = Date.now();
    const expiryMs = expires_in * 1000;
    let pollInterval = interval * 1000;

    while (Date.now() - startTime < expiryMs) {
      await new Promise(resolve => setTimeout(resolve, pollInterval));

      const tokenResponse = await fetch(this.tokenUrl, {
        method: 'POST',
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
        body: new URLSearchParams({
          grant_type: 'urn:ietf:params:oauth:grant-type:device_code',
          device_code,
          client_id: this.clientId,
        }),
      });

      const tokenData = await tokenResponse.json();

      if (tokenData.error) {
        switch (tokenData.error) {
          case 'authorization_pending':
            continue; // Keep polling
          case 'slow_down':
            pollInterval += 5000; // Increase interval
            continue;
          case 'expired_token':
            throw new Error('Device code expired. Please try again.');
          case 'access_denied':
            throw new Error('Authorization denied by user.');
          default:
            throw new Error(`OAuth error: ${tokenData.error}`);
        }
      }

      // Success
      console.log('  Authenticated successfully!\n');
      return {
        accessToken: tokenData.access_token,
        refreshToken: tokenData.refresh_token,
        expiresIn: tokenData.expires_in,
      };
    }

    throw new Error('Authorization timed out. Please try again.');
  }
}

module.exports = { DeviceFlowAuth };
```

---

## 5. Token Refresh Flow

### Flow Diagram

```
     ┌──────────┐                        ┌────────────┐
     │          │                        │            │
     │  Client  │                        │ Auth Server│
     │          │                        │            │
     └────┬─────┘                        └─────┬──────┘
          │                                    │
          │ Access token expired               │
          │                                    │
          │ 1. POST /token                     │
          │    grant_type=refresh_token         │
          │    refresh_token=RT_1              │
          │    client_id=...                   │
          │    client_secret=... (if confid.)  │
          │───────────────────────────────────>│
          │                                    │
          │ 2. Validate refresh token          │
          │    Check expiry                    │
          │    Check revocation                │
          │                                    │
          │ 3. Token Response                  │
          │    {                               │
          │      access_token: "NEW_AT",       │
          │      refresh_token: "NEW_RT",      │
          │      expires_in: 3600              │
          │    }                               │
          │                                    │
          │    OLD refresh_token invalidated   │
          │    (refresh token rotation)        │
          │<───────────────────────────────────│
          │                                    │
```

### Implementation

```javascript
// auth/token-refresh.js — Token refresh with rotation
async function refreshAccessToken(refreshToken, config) {
  const response = await fetch(config.tokenUrl, {
    method: 'POST',
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    body: new URLSearchParams({
      grant_type: 'refresh_token',
      refresh_token: refreshToken,
      client_id: config.clientId,
      ...(config.clientSecret && { client_secret: config.clientSecret }),
    }),
  });

  if (!response.ok) {
    if (response.status === 401 || response.status === 400) {
      // Refresh token expired or revoked — require re-authentication
      throw new TokenRefreshError('Refresh token invalid — re-authentication required');
    }
    throw new Error(`Token refresh failed: ${response.status}`);
  }

  return response.json();
}

// Axios interceptor for automatic token refresh
function setupTokenRefreshInterceptor(axiosInstance, tokenStore, config) {
  let isRefreshing = false;
  let failedQueue = [];

  const processQueue = (error, token) => {
    failedQueue.forEach(({ resolve, reject }) => {
      if (error) reject(error);
      else resolve(token);
    });
    failedQueue = [];
  };

  axiosInstance.interceptors.response.use(
    response => response,
    async error => {
      const originalRequest = error.config;

      if (error.response?.status !== 401 || originalRequest._retry) {
        return Promise.reject(error);
      }

      if (isRefreshing) {
        return new Promise((resolve, reject) => {
          failedQueue.push({ resolve, reject });
        }).then(token => {
          originalRequest.headers.Authorization = `Bearer ${token}`;
          return axiosInstance(originalRequest);
        });
      }

      originalRequest._retry = true;
      isRefreshing = true;

      try {
        const tokens = await refreshAccessToken(tokenStore.getRefreshToken(), config);
        tokenStore.setTokens(tokens);
        processQueue(null, tokens.access_token);
        originalRequest.headers.Authorization = `Bearer ${tokens.access_token}`;
        return axiosInstance(originalRequest);
      } catch (refreshError) {
        processQueue(refreshError, null);
        tokenStore.clearTokens();
        window.location.href = '/login';
        return Promise.reject(refreshError);
      } finally {
        isRefreshing = false;
      }
    }
  );
}
```

---

## OIDC (OpenID Connect) Extensions

### ID Token Claims

```
Standard OIDC Claims:
┌──────────────────┬─────────────────────────────────────────┐
│ Claim            │ Description                             │
├──────────────────┼─────────────────────────────────────────┤
│ iss              │ Issuer URL                              │
│ sub              │ Subject (unique user identifier)        │
│ aud              │ Audience (client_id)                    │
│ exp              │ Expiration time                         │
│ iat              │ Issued at time                          │
│ auth_time        │ Time of authentication                  │
│ nonce            │ Nonce for replay prevention             │
│ acr              │ Authentication Context Class Reference  │
│ amr              │ Authentication Methods References       │
│ azp              │ Authorized Party                        │
├──────────────────┼─────────────────────────────────────────┤
│ email            │ User's email address                    │
│ email_verified   │ Whether email is verified               │
│ name             │ Full name                               │
│ given_name       │ First name                              │
│ family_name      │ Last name                               │
│ picture          │ Profile picture URL                     │
│ locale           │ User's locale (e.g., "en-US")           │
│ updated_at       │ Last profile update time                │
└──────────────────┴─────────────────────────────────────────┘
```

### OIDC Discovery

```
GET https://auth.example.com/.well-known/openid-configuration

Response:
{
  "issuer": "https://auth.example.com",
  "authorization_endpoint": "https://auth.example.com/authorize",
  "token_endpoint": "https://auth.example.com/token",
  "userinfo_endpoint": "https://auth.example.com/userinfo",
  "jwks_uri": "https://auth.example.com/.well-known/jwks.json",
  "end_session_endpoint": "https://auth.example.com/logout",
  "revocation_endpoint": "https://auth.example.com/revoke",
  "introspection_endpoint": "https://auth.example.com/introspect",
  "device_authorization_endpoint": "https://auth.example.com/device/code",
  "scopes_supported": ["openid", "profile", "email", "offline_access"],
  "response_types_supported": ["code"],
  "grant_types_supported": [
    "authorization_code",
    "refresh_token",
    "client_credentials",
    "urn:ietf:params:oauth:grant-type:device_code"
  ],
  "code_challenge_methods_supported": ["S256"],
  "token_endpoint_auth_methods_supported": [
    "client_secret_basic",
    "client_secret_post",
    "private_key_jwt"
  ],
  "id_token_signing_alg_values_supported": ["RS256", "ES256"],
  "subject_types_supported": ["public"]
}
```

---

## OAuth2 Security Vulnerabilities

### Common Attacks and Mitigations

```
┌──────────────────────┬────────────────────────────────────────────┐
│ Attack               │ Mitigation                                 │
├──────────────────────┼────────────────────────────────────────────┤
│ Authorization code   │ Use PKCE (code_challenge + code_verifier)  │
│ interception         │                                            │
├──────────────────────┼────────────────────────────────────────────┤
│ CSRF on callback     │ Use state parameter with random value      │
├──────────────────────┼────────────────────────────────────────────┤
│ Open redirect        │ Exact redirect_uri matching on server      │
│                      │ Never use wildcards in redirect URIs       │
├──────────────────────┼────────────────────────────────────────────┤
│ Token leakage via    │ Use Authorization Code (not Implicit)      │
│ URL fragment         │ Never put tokens in URLs                   │
├──────────────────────┼────────────────────────────────────────────┤
│ Token theft (XSS)    │ Store tokens in httpOnly cookies           │
│                      │ Never use localStorage for tokens          │
├──────────────────────┼────────────────────────────────────────────┤
│ Mix-up attack        │ Validate issuer in ID token                │
│ (multi-provider)     │ Match state to provider                    │
├──────────────────────┼────────────────────────────────────────────┤
│ Refresh token theft  │ Rotation + reuse detection                 │
│                      │ Bind to client + IP when possible          │
├──────────────────────┼────────────────────────────────────────────┤
│ Client impersonation │ Use client authentication (secret/mTLS)    │
│                      │ PKCE for public clients                    │
├──────────────────────┼────────────────────────────────────────────┤
│ Scope escalation     │ Server enforces allowed scopes per client  │
│                      │ Never trust client-requested scopes        │
├──────────────────────┼────────────────────────────────────────────┤
│ ID token replay      │ Use nonce claim + validation               │
│                      │ Short token lifetime                       │
└──────────────────────┴────────────────────────────────────────────┘
```

---

## Provider-Specific Configuration

### Google OAuth2

```javascript
// config/google-oauth.js
const googleOAuth = {
  authorizationUrl: 'https://accounts.google.com/o/oauth2/v2/auth',
  tokenUrl: 'https://oauth2.googleapis.com/token',
  userinfoUrl: 'https://openidconnect.googleapis.com/v1/userinfo',
  discoveryUrl: 'https://accounts.google.com/.well-known/openid-configuration',
  scopes: ['openid', 'profile', 'email'],
  // Prompt options: none, consent, select_account
  extraParams: { prompt: 'select_account' },
};
```

### GitHub OAuth2

```javascript
// config/github-oauth.js
const githubOAuth = {
  authorizationUrl: 'https://github.com/login/oauth/authorize',
  tokenUrl: 'https://github.com/login/oauth/access_token',
  userinfoUrl: 'https://api.github.com/user',
  emailsUrl: 'https://api.github.com/user/emails', // Separate endpoint for emails
  scopes: ['user:email', 'read:user'],
  // GitHub uses 'login' param instead of 'login_hint'
  // GitHub doesn't support OIDC — plain OAuth2 only
};
```

### Microsoft (Azure AD) OAuth2

```javascript
// config/microsoft-oauth.js
const microsoftOAuth = {
  // Use 'common' for multi-tenant, or specific tenant ID
  authorizationUrl: 'https://login.microsoftonline.com/common/oauth2/v2.0/authorize',
  tokenUrl: 'https://login.microsoftonline.com/common/oauth2/v2.0/token',
  discoveryUrl: 'https://login.microsoftonline.com/common/v2.0/.well-known/openid-configuration',
  scopes: ['openid', 'profile', 'email', 'User.Read'],
  extraParams: { prompt: 'select_account' },
};
```

### Apple Sign In

```javascript
// config/apple-oauth.js
const appleOAuth = {
  authorizationUrl: 'https://appleid.apple.com/auth/authorize',
  tokenUrl: 'https://appleid.apple.com/auth/token',
  scopes: ['name', 'email'],
  responseMode: 'form_post', // Apple uses form_post
  // Apple requires JWT client authentication (not client_secret)
  // Client secret is a signed JWT using the Apple private key
};
```

---

## OAuth2 Token Storage Best Practices

```
┌────────────────────┬─────────┬──────────┬──────────────────────────┐
│ Storage Method     │ XSS     │ CSRF     │ Notes                    │
│                    │ Safe?   │ Safe?    │                          │
├────────────────────┼─────────┼──────────┼──────────────────────────┤
│ httpOnly cookie    │ Yes     │ No       │ Add SameSite + CSRF      │
│                    │         │          │ token. Best for most     │
│                    │         │          │ web apps.                │
├────────────────────┼─────────┼──────────┼──────────────────────────┤
│ In-memory variable │ Yes     │ Yes      │ Lost on page refresh.    │
│                    │         │          │ Best for SPAs with       │
│                    │         │          │ backend token proxy.     │
├────────────────────┼─────────┼──────────┼──────────────────────────┤
│ localStorage       │ No      │ Yes      │ NEVER use for tokens.    │
│                    │         │          │ Accessible by any JS.    │
├────────────────────┼─────────┼──────────┼──────────────────────────┤
│ sessionStorage     │ No      │ Yes      │ Slightly better than     │
│                    │         │          │ localStorage (tab-only). │
│                    │         │          │ Still XSS vulnerable.    │
├────────────────────┼─────────┼──────────┼──────────────────────────┤
│ Server-side session│ Yes     │ Depends  │ Most secure. Tokens      │
│                    │         │          │ never leave server.      │
│                    │         │          │ Use with session cookie. │
└────────────────────┴─────────┴──────────┴──────────────────────────┘

Recommendation:
- Server-rendered apps → Server-side session (tokens never reach browser)
- SPAs → httpOnly cookie for refresh token + in-memory for access token
- Mobile → Secure Keychain (iOS) / Keystore (Android)
- Desktop → OS credential manager
```
