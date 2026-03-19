# Session Management Reference

Complete reference for secure session management — cookie configuration, server-side storage, session lifecycle, security hardening, and implementation patterns for Node.js, Python, and Go.

---

## Session Architecture Overview

```
┌──────────────────────────────────────────────────────────────────┐
│                    Session Management Architecture               │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Browser                          Server                         │
│  ┌──────────────────┐            ┌─────────────────────────┐    │
│  │                  │            │                         │    │
│  │  Session Cookie  │◄──────────►│  Session Middleware     │    │
│  │  __Host-sid=abc  │  (cookie)  │  ┌───────────────────┐ │    │
│  │                  │            │  │ Parse cookie       │ │    │
│  │  Properties:     │            │  │ Load session data  │ │    │
│  │  - httpOnly      │            │  │ Attach to request  │ │    │
│  │  - secure        │            │  │ Save on response   │ │    │
│  │  - sameSite=lax  │            │  └─────────┬─────────┘ │    │
│  │  - path=/        │            │            │           │    │
│  │                  │            │            ▼           │    │
│  └──────────────────┘            │  ┌───────────────────┐ │    │
│                                  │  │  Session Store    │ │    │
│                                  │  │  ┌─────────────┐  │ │    │
│                                  │  │  │ Redis       │  │ │    │
│                                  │  │  │ PostgreSQL  │  │ │    │
│                                  │  │  │ Memory      │  │ │    │
│                                  │  │  └─────────────┘  │ │    │
│                                  │  └───────────────────┘ │    │
│                                  └─────────────────────────┘    │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘
```

---

## Cookie Security Configuration

### Cookie Attribute Reference

```
┌──────────────────┬──────────────────────────────────────────────┐
│ Attribute        │ Description + Recommendation                 │
├──────────────────┼──────────────────────────────────────────────┤
│ Name             │ Use "__Host-" prefix for maximum security.   │
│                  │ __Host-sid forces: Secure, no Domain, Path=/│
│                  │ Example: __Host-session                      │
├──────────────────┼──────────────────────────────────────────────┤
│ HttpOnly         │ ALWAYS true. Prevents JavaScript access.     │
│                  │ Blocks XSS from stealing session cookies.    │
├──────────────────┼──────────────────────────────────────────────┤
│ Secure           │ ALWAYS true in production. Cookie only sent  │
│                  │ over HTTPS. Required for __Host- prefix.     │
├──────────────────┼──────────────────────────────────────────────┤
│ SameSite         │ "Lax" — sent on top-level navigations only.  │
│                  │ "Strict" — never sent cross-site (breaks     │
│                  │ some flows like OAuth callbacks).             │
│                  │ "None" — sent cross-site (needs Secure).     │
│                  │ Recommendation: "Lax" for most apps.         │
├──────────────────┼──────────────────────────────────────────────┤
│ Path             │ "/" — available on all paths.                │
│                  │ Restrict if session is only needed on        │
│                  │ specific paths (rare).                       │
├──────────────────┼──────────────────────────────────────────────┤
│ Domain           │ Don't set (defaults to exact current host).  │
│                  │ Setting Domain enables subdomain access      │
│                  │ (usually not wanted for session cookies).    │
│                  │ __Host- prefix prevents setting Domain.      │
├──────────────────┼──────────────────────────────────────────────┤
│ Max-Age          │ Session duration in seconds.                 │
│                  │ 86400 = 24 hours (recommended max).         │
│                  │ Omit for session cookies (cleared on close). │
├──────────────────┼──────────────────────────────────────────────┤
│ Expires          │ Alternative to Max-Age (absolute date).     │
│                  │ Max-Age takes precedence if both set.        │
└──────────────────┴──────────────────────────────────────────────┘
```

### Cookie Prefix Security

```
Cookie Prefixes (enforced by browsers):

__Host- prefix:
  Requirements: Secure=true, Path=/, no Domain attribute
  Purpose: Prevents subdomain cookie overwriting
  Example: __Host-sid=abc123
  Use for: Session cookies, CSRF tokens

__Secure- prefix:
  Requirements: Secure=true
  Purpose: Ensures cookie only sent over HTTPS
  Example: __Secure-token=xyz789
  Use for: Less restrictive than __Host-, allows Domain

Recommendation: Always use __Host- for session cookies
```

---

## Session Store Comparison

```
┌──────────────────┬──────────┬───────────┬────────────┬──────────┐
│ Store            │ Speed    │ Scaling   │ Persistence│ Cost     │
├──────────────────┼──────────┼───────────┼────────────┼──────────┤
│ Redis            │ ~1ms     │ Cluster,  │ Optional   │ Moderate │
│                  │          │ sentinel  │ (RDB/AOF)  │          │
├──────────────────┼──────────┼───────────┼────────────┼──────────┤
│ PostgreSQL       │ ~5-20ms  │ Read      │ Always     │ Low      │
│                  │          │ replicas  │            │          │
├──────────────────┼──────────┼───────────┼────────────┼──────────┤
│ MongoDB          │ ~5-15ms  │ Sharding, │ Always     │ Moderate │
│                  │          │ replica   │            │          │
├──────────────────┼──────────┼───────────┼────────────┼──────────┤
│ DynamoDB         │ ~5-10ms  │ Automatic │ Always     │ Pay/use  │
├──────────────────┼──────────┼───────────┼────────────┼──────────┤
│ Memcached        │ ~1ms     │ Sharding  │ None       │ Low      │
├──────────────────┼──────────┼───────────┼────────────┼──────────┤
│ In-Memory        │ <1ms     │ None      │ None       │ Free     │
│ (process)        │          │ (single   │ (lost on   │          │
│                  │          │ process)  │ restart)   │          │
└──────────────────┴──────────┴───────────┴────────────┴──────────┘

Recommendations:
- Development: In-memory (no setup needed)
- Small production: PostgreSQL (already have it for users)
- Medium-large production: Redis (fast, TTL support, clustering)
- Serverless: DynamoDB or Redis (managed, auto-scaling)
- NEVER use in-memory for production (lost on restart, can't scale)
```

---

## Node.js Session Implementation

### Express + Redis Sessions

```javascript
// session/config.js — Production session configuration
const session = require('express-session');
const RedisStore = require('connect-redis').default;
const { createClient } = require('redis');
const crypto = require('crypto');

async function setupSession(app) {
  const redisClient = createClient({
    url: process.env.REDIS_URL || 'redis://localhost:6379',
    socket: {
      reconnectStrategy: (retries) => {
        if (retries > 10) return new Error('Redis retry limit reached');
        return Math.min(retries * 100, 5000);
      },
    },
  });

  redisClient.on('error', (err) => console.error('Redis error:', err));
  await redisClient.connect();

  const store = new RedisStore({
    client: redisClient,
    prefix: 'sess:',
    ttl: 86400, // 24 hours (seconds)
  });

  app.use(session({
    store,
    name: '__Host-sid',
    secret: process.env.SESSION_SECRET, // Use a strong random string (64+ chars)
    resave: false,
    saveUninitialized: false,
    rolling: true, // Reset maxAge on every request

    cookie: {
      secure: process.env.NODE_ENV === 'production',
      httpOnly: true,
      sameSite: 'lax',
      maxAge: 24 * 60 * 60 * 1000, // 24 hours (milliseconds)
      path: '/',
    },

    genid: () => crypto.randomBytes(32).toString('hex'),
  }));

  return store;
}

module.exports = { setupSession };
```

### Express + PostgreSQL Sessions

```javascript
// session/pg-store.js — PostgreSQL session store
const session = require('express-session');
const pgSession = require('connect-pg-simple')(session);
const { Pool } = require('pg');

function setupPgSession(app) {
  const pool = new Pool({
    connectionString: process.env.DATABASE_URL,
  });

  app.use(session({
    store: new pgSession({
      pool,
      tableName: 'sessions',
      createTableIfMissing: true,
      pruneSessionInterval: 300, // Clean expired sessions every 5 min
      errorLog: console.error,
    }),
    name: '__Host-sid',
    secret: process.env.SESSION_SECRET,
    resave: false,
    saveUninitialized: false,
    rolling: true,

    cookie: {
      secure: process.env.NODE_ENV === 'production',
      httpOnly: true,
      sameSite: 'lax',
      maxAge: 24 * 60 * 60 * 1000,
      path: '/',
    },
  }));
}

// SQL schema for sessions table (auto-created by connect-pg-simple)
/*
CREATE TABLE sessions (
    sid VARCHAR NOT NULL PRIMARY KEY,
    sess JSON NOT NULL,
    expire TIMESTAMP(6) NOT NULL
);
CREATE INDEX idx_sessions_expire ON sessions(expire);
*/
```

---

## Python Session Implementation

### FastAPI + Redis Sessions

```python
# session/config.py — FastAPI session with Redis
from starlette.middleware.sessions import SessionMiddleware
from itsdangerous import URLSafeTimedSerializer
import redis.asyncio as aioredis
import json
import secrets

class RedisSessionMiddleware:
    """Custom session middleware using Redis backend."""

    def __init__(
        self,
        app,
        redis_url: str = "redis://localhost:6379",
        session_cookie: str = "__Host-sid",
        max_age: int = 86400,
        secret_key: str = None,
    ):
        self.app = app
        self.redis_url = redis_url
        self.session_cookie = session_cookie
        self.max_age = max_age
        self.serializer = URLSafeTimedSerializer(secret_key or secrets.token_hex(32))
        self.redis = None

    async def __call__(self, scope, receive, send):
        if scope["type"] not in ("http", "websocket"):
            await self.app(scope, receive, send)
            return

        if not self.redis:
            self.redis = aioredis.from_url(self.redis_url)

        # Load session
        session_id = self._get_session_id(scope)
        session_data = {}

        if session_id:
            raw = await self.redis.get(f"sess:{session_id}")
            if raw:
                session_data = json.loads(raw)

        scope["session"] = session_data
        scope["session_id"] = session_id

        async def send_wrapper(message):
            if message["type"] == "http.response.start":
                # Save session
                if scope.get("session"):
                    sid = scope.get("session_id") or secrets.token_hex(32)
                    await self.redis.set(
                        f"sess:{sid}",
                        json.dumps(scope["session"]),
                        ex=self.max_age,
                    )
                    # Set cookie
                    cookie = (
                        f"{self.session_cookie}={sid}; "
                        f"Path=/; HttpOnly; Secure; SameSite=Lax; "
                        f"Max-Age={self.max_age}"
                    )
                    headers = list(message.get("headers", []))
                    headers.append((b"set-cookie", cookie.encode()))
                    message["headers"] = headers

            await send(message)

        await self.app(scope, receive, send_wrapper)

    def _get_session_id(self, scope):
        headers = dict(scope.get("headers", []))
        cookie_header = headers.get(b"cookie", b"").decode()
        for cookie in cookie_header.split(";"):
            cookie = cookie.strip()
            if cookie.startswith(f"{self.session_cookie}="):
                return cookie.split("=", 1)[1]
        return None
```

---

## Go Session Implementation

### Go + Redis Sessions

```go
// session/session.go — Go session management with Redis
package session

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

type SessionStore struct {
	redis      *redis.Client
	cookieName string
	maxAge     time.Duration
	prefix     string
}

type Session struct {
	ID   string
	Data map[string]interface{}
	store *SessionStore
	dirty bool
}

func NewSessionStore(redisClient *redis.Client) *SessionStore {
	return &SessionStore{
		redis:      redisClient,
		cookieName: "__Host-sid",
		maxAge:     24 * time.Hour,
		prefix:     "sess:",
	}
}

func generateSessionID() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (s *SessionStore) Get(r *http.Request) (*Session, error) {
	cookie, err := r.Cookie(s.cookieName)
	session := &Session{
		Data:  make(map[string]interface{}),
		store: s,
	}

	if err != nil || cookie.Value == "" {
		session.ID = generateSessionID()
		return session, nil
	}

	session.ID = cookie.Value

	ctx := context.Background()
	data, err := s.redis.Get(ctx, s.prefix+session.ID).Result()
	if err == redis.Nil {
		session.ID = generateSessionID()
		return session, nil
	}
	if err != nil {
		return nil, err
	}

	json.Unmarshal([]byte(data), &session.Data)
	return session, nil
}

func (s *SessionStore) Save(w http.ResponseWriter, session *Session) error {
	ctx := context.Background()
	data, err := json.Marshal(session.Data)
	if err != nil {
		return err
	}

	err = s.redis.Set(ctx, s.prefix+session.ID, data, s.maxAge).Err()
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     s.cookieName,
		Value:    session.ID,
		Path:     "/",
		MaxAge:   int(s.maxAge.Seconds()),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	return nil
}

func (s *SessionStore) Destroy(w http.ResponseWriter, session *Session) error {
	ctx := context.Background()
	s.redis.Del(ctx, s.prefix+session.ID)

	http.SetCookie(w, &http.Cookie{
		Name:     s.cookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	return nil
}

// Middleware
func (s *SessionStore) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := s.Get(r)
		if err != nil {
			http.Error(w, "Session error", http.StatusInternalServerError)
			return
		}

		ctx := context.WithValue(r.Context(), "session", session)
		r = r.WithContext(ctx)

		// Wrap response writer to save session
		srw := &sessionResponseWriter{
			ResponseWriter: w,
			session:        session,
			store:          s,
		}

		next.ServeHTTP(srw, r)

		// Save session after handler runs
		if session.dirty {
			s.Save(w, session)
		}
	})
}

type sessionResponseWriter struct {
	http.ResponseWriter
	session *Session
	store   *SessionStore
}

func (s *Session) Set(key string, value interface{}) {
	s.Data[key] = value
	s.dirty = true
}

func (s *Session) Get(key string) interface{} {
	return s.Data[key]
}

func (s *Session) Delete(key string) {
	delete(s.Data, key)
	s.dirty = true
}
```

---

## Session Security Hardening

### Session Fixation Prevention

```
Attack: Attacker sets victim's session ID before they log in.
1. Attacker gets a valid session ID from the app
2. Attacker tricks victim into using that session ID
3. Victim logs in — session is now authenticated
4. Attacker uses the known session ID to access victim's account

Prevention: ALWAYS regenerate session ID after authentication.
```

```javascript
// security/session-fixation.js — Session fixation prevention
function regenerateSession(req) {
  return new Promise((resolve, reject) => {
    const oldData = { ...req.session };
    const oldId = req.session.id;

    req.session.regenerate((err) => {
      if (err) return reject(err);

      // Restore session data to new session
      Object.assign(req.session, oldData);
      req.session.lastRegeneration = Date.now();

      // Log the regeneration
      console.log(`Session regenerated: ${oldId.substring(0, 8)}... → ${req.session.id.substring(0, 8)}...`);

      req.session.save(resolve);
    });
  });
}

// Use after login:
router.post('/login', async (req, res) => {
  // ... validate credentials ...
  await regenerateSession(req);
  req.session.userId = user.id;
  req.session.loginTime = Date.now();
  res.json({ success: true });
});
```

### Session Hijacking Prevention

```javascript
// security/session-hijacking.js — Detect session hijacking
function sessionBindingMiddleware(req, res, next) {
  if (!req.session.userId) return next();

  // Bind 1: User-Agent fingerprint
  const currentUA = req.headers['user-agent'] || '';
  if (req.session.userAgent && req.session.userAgent !== currentUA) {
    console.warn(`Session hijack attempt: UA mismatch for user ${req.session.userId}`);
    return destroyAndRedirect(req, res);
  }
  req.session.userAgent = currentUA;

  // Bind 2: Accept-Language header (stable per browser)
  const currentLang = req.headers['accept-language'] || '';
  if (req.session.acceptLanguage && req.session.acceptLanguage !== currentLang) {
    // Log but don't block (users might change language settings)
    console.warn(`Session binding: Accept-Language changed for user ${req.session.userId}`);
  }
  req.session.acceptLanguage = currentLang;

  // Note: Don't bind to IP — users on mobile frequently change IPs
  // IP binding causes false positives with VPNs and mobile networks

  next();
}

function destroyAndRedirect(req, res) {
  req.session.destroy(() => {
    res.clearCookie('__Host-sid');
    if (req.xhr) {
      res.status(401).json({ error: 'Session invalidated' });
    } else {
      res.redirect('/login?reason=session_invalidated');
    }
  });
}
```

### Session Timeout Strategies

```
┌──────────────────────┬──────────────────────────────────────────┐
│ Timeout Type         │ Implementation                           │
├──────────────────────┼──────────────────────────────────────────┤
│ Absolute timeout     │ Session expires N hours after creation,  │
│ (24 hours)           │ regardless of activity. Forces re-auth.  │
├──────────────────────┼──────────────────────────────────────────┤
│ Idle timeout         │ Session expires after N minutes of       │
│ (30 minutes)         │ inactivity. Resets on each request.     │
├──────────────────────┼──────────────────────────────────────────┤
│ Rolling timeout      │ Cookie maxAge resets on every request.   │
│ (sliding window)     │ User stays logged in while active.      │
├──────────────────────┼──────────────────────────────────────────┤
│ Privileged timeout   │ Require re-auth for sensitive actions    │
│ (15 minutes)         │ if last auth was > 15 minutes ago.      │
└──────────────────────┴──────────────────────────────────────────┘
```

```javascript
// security/session-timeout.js — Multiple timeout strategies
function sessionTimeoutMiddleware(options = {}) {
  const absoluteTimeout = options.absoluteTimeout || 24 * 60 * 60 * 1000; // 24h
  const idleTimeout = options.idleTimeout || 30 * 60 * 1000; // 30min
  const privilegedTimeout = options.privilegedTimeout || 15 * 60 * 1000; // 15min

  return (req, res, next) => {
    if (!req.session.userId) return next();

    const now = Date.now();

    // 1. Absolute timeout — force re-auth after 24 hours
    if (req.session.loginTime && now - req.session.loginTime > absoluteTimeout) {
      console.log(`Absolute timeout for user ${req.session.userId}`);
      return destroySession(req, res, 'session_expired');
    }

    // 2. Idle timeout — expire after 30 minutes of inactivity
    if (req.session.lastActivity && now - req.session.lastActivity > idleTimeout) {
      console.log(`Idle timeout for user ${req.session.userId}`);
      return destroySession(req, res, 'session_idle');
    }

    // Update last activity
    req.session.lastActivity = now;

    // 3. Privileged timeout — flag if last auth was > 15 minutes ago
    const lastAuth = req.session.lastAuthTime || req.session.loginTime || 0;
    req.isRecentlyAuthenticated = now - lastAuth < privilegedTimeout;

    next();
  };
}

// Middleware: require recent authentication for sensitive operations
function requireRecentAuth(req, res, next) {
  if (!req.isRecentlyAuthenticated) {
    return res.status(403).json({
      error: 'Recent authentication required',
      code: 'REQUIRE_RECENT_AUTH',
      message: 'Please re-enter your password to continue',
    });
  }
  next();
}

function destroySession(req, res, reason) {
  req.session.destroy(() => {
    res.clearCookie('__Host-sid');
    if (req.xhr || req.headers.accept?.includes('json')) {
      res.status(401).json({ error: 'Session expired', reason });
    } else {
      res.redirect(`/login?reason=${reason}`);
    }
  });
}

module.exports = { sessionTimeoutMiddleware, requireRecentAuth };
```

---

## Concurrent Session Management

```javascript
// security/concurrent-sessions.js — Limit active sessions per user
class ConcurrentSessionManager {
  constructor(redis, maxSessions = 5) {
    this.redis = redis;
    this.maxSessions = maxSessions;
    this.prefix = 'user_sessions:';
  }

  // Register a new session for a user
  async registerSession(userId, sessionId, metadata = {}) {
    const key = `${this.prefix}${userId}`;
    const sessionData = JSON.stringify({
      sessionId,
      createdAt: Date.now(),
      userAgent: metadata.userAgent,
      ip: metadata.ip,
    });

    // Add to sorted set (score = creation time)
    await this.redis.zadd(key, Date.now(), sessionData);

    // Check if over limit
    const count = await this.redis.zcard(key);
    if (count > this.maxSessions) {
      // Remove oldest sessions
      const toRemove = count - this.maxSessions;
      const oldSessions = await this.redis.zrange(key, 0, toRemove - 1);

      // Destroy old sessions
      for (const sessionJson of oldSessions) {
        const session = JSON.parse(sessionJson);
        await this.redis.del(`sess:${session.sessionId}`);
      }

      // Remove from user's session set
      await this.redis.zremrangebyrank(key, 0, toRemove - 1);
    }

    // Set TTL on the user sessions key
    await this.redis.expire(key, 86400 * 30); // 30 days
  }

  // Get all active sessions for a user
  async getUserSessions(userId) {
    const key = `${this.prefix}${userId}`;
    const sessions = await this.redis.zrange(key, 0, -1);
    return sessions.map(s => JSON.parse(s));
  }

  // Terminate a specific session
  async terminateSession(userId, sessionId) {
    const key = `${this.prefix}${userId}`;
    const sessions = await this.redis.zrange(key, 0, -1);

    for (const sessionJson of sessions) {
      const session = JSON.parse(sessionJson);
      if (session.sessionId === sessionId) {
        await this.redis.zrem(key, sessionJson);
        await this.redis.del(`sess:${sessionId}`);
        break;
      }
    }
  }

  // Terminate all sessions except current
  async terminateOtherSessions(userId, currentSessionId) {
    const sessions = await this.getUserSessions(userId);
    for (const session of sessions) {
      if (session.sessionId !== currentSessionId) {
        await this.terminateSession(userId, session.sessionId);
      }
    }
  }

  // Terminate all sessions (logout everywhere)
  async terminateAllSessions(userId) {
    const sessions = await this.getUserSessions(userId);
    for (const session of sessions) {
      await this.redis.del(`sess:${session.sessionId}`);
    }
    await this.redis.del(`${this.prefix}${userId}`);
  }
}

module.exports = { ConcurrentSessionManager };
```

---

## Session Data Best Practices

```
What to store in sessions:
✅ User ID (reference to database)
✅ Authentication time (for timeout checks)
✅ Last activity timestamp
✅ CSRF token
✅ OAuth access/refresh tokens (server-side only)
✅ User agent fingerprint (session binding)
✅ Flash messages (temporary notifications)
✅ Form wizard state (multi-step forms)

What NOT to store in sessions:
❌ Full user objects (becomes stale)
❌ Passwords or hashes
❌ Large data sets (keep sessions small)
❌ Sensitive PII (encrypt if necessary)
❌ Shopping cart items (use database)
❌ Anything > 1MB (Redis/cookie size limits)

Session data structure:
{
  "userId": "uuid-123",
  "loginTime": 1700000000000,
  "lastActivity": 1700003600000,
  "lastRegeneration": 1700003500000,
  "userAgent": "Mozilla/5.0...",
  "csrfToken": "random-csrf-token",
  "accessToken": "encrypted-oauth-token",
  "refreshToken": "encrypted-refresh-token",
  "tokenExpiry": 1700007200
}
```

---

## Session Cleanup

### Automatic Session Cleanup

```javascript
// session/cleanup.js — Automated session cleanup
class SessionCleanup {
  constructor(redis, options = {}) {
    this.redis = redis;
    this.interval = options.interval || 5 * 60 * 1000; // Every 5 minutes
    this.prefix = options.prefix || 'sess:';
    this.timer = null;
  }

  start() {
    this.timer = setInterval(() => this.cleanup(), this.interval);
    console.log(`Session cleanup started (every ${this.interval / 1000}s)`);
  }

  stop() {
    if (this.timer) {
      clearInterval(this.timer);
      this.timer = null;
    }
  }

  async cleanup() {
    try {
      // Redis handles TTL-based expiry automatically for individual keys.
      // This cleanup handles edge cases:

      // 1. Clean up user session tracking sets with no active sessions
      const userKeys = [];
      let cursor = '0';
      do {
        const result = await this.redis.scan(cursor, 'MATCH', 'user_sessions:*', 'COUNT', 100);
        cursor = result[0];
        userKeys.push(...result[1]);
      } while (cursor !== '0');

      for (const key of userKeys) {
        const count = await this.redis.zcard(key);
        if (count === 0) {
          await this.redis.del(key);
        }
      }
    } catch (error) {
      console.error('Session cleanup error:', error);
    }
  }
}

module.exports = { SessionCleanup };
```

### PostgreSQL Cleanup

```sql
-- Cleanup expired sessions (run periodically via cron or pg_cron)
DELETE FROM sessions WHERE expire < NOW();

-- Create pg_cron job for automatic cleanup
SELECT cron.schedule('cleanup-sessions', '*/5 * * * *',
  $$DELETE FROM sessions WHERE expire < NOW()$$
);

-- Monitor session count
SELECT
  COUNT(*) as total_sessions,
  COUNT(*) FILTER (WHERE expire > NOW()) as active_sessions,
  COUNT(*) FILTER (WHERE expire <= NOW()) as expired_sessions
FROM sessions;
```

---

## Security Checklist

```
✅ Session Management Security Checklist

Cookie Configuration:
□ Use __Host- cookie prefix
□ Set HttpOnly = true (prevents JS access)
□ Set Secure = true (HTTPS only)
□ Set SameSite = Lax (CSRF protection)
□ Set appropriate Max-Age (24h max recommended)
□ Don't set Domain attribute (use __Host- prefix)
□ Set Path = /

Session ID:
□ Generate cryptographically random IDs (32+ bytes)
□ Use crypto.randomBytes() or equivalent
□ Never use sequential or predictable IDs
□ Never include session ID in URLs
□ Never log full session IDs

Session Lifecycle:
□ Regenerate session ID after login (fixation prevention)
□ Regenerate session ID periodically (every 15 minutes)
□ Implement absolute timeout (24 hours max)
□ Implement idle timeout (30 minutes)
□ Destroy session completely on logout
□ Clear cookie on logout (set Max-Age = -1)

Session Binding:
□ Bind session to User-Agent (detect hijacking)
□ Don't bind to IP (causes false positives on mobile)
□ Require re-auth for sensitive operations
□ Log session creation and destruction

Server-Side:
□ Use Redis or PostgreSQL for session storage
□ Never use in-memory store in production
□ Encrypt sensitive session data at rest
□ Set TTL on session store entries
□ Clean up expired sessions automatically
□ Limit concurrent sessions per user (5 max)
□ Provide "active sessions" UI for users

Infrastructure:
□ Use HTTPS everywhere
□ Set proper HSTS headers
□ Monitor for session anomalies
□ Log authentication events
□ Implement rate limiting on login
```

---

## Session vs JWT Decision Matrix

```
Use sessions when:
✅ Server-rendered application
✅ Need immediate session revocation
✅ Session data changes frequently
✅ Single-origin application
✅ Simplicity is important

Use JWT when:
✅ Microservices architecture
✅ API consumed by multiple clients
✅ Cross-origin requests needed
✅ Stateless verification required
✅ Mobile app backend

Use both (hybrid):
✅ SPA with API backend
   → Session cookie for refresh token
   → JWT for API access token
✅ Microservices with user-facing frontend
   → Session on BFF (Backend for Frontend)
   → JWT between internal services
```
