# Secure Coding Patterns — Language-Specific Reference

Comprehensive secure coding patterns for building resilient applications. Covers input validation, output encoding, parameterized queries, session management, CORS, CSP, cookie security, file operations, deserialization, cryptography, and secret management — with production-ready code examples in JavaScript/TypeScript, Python, Java, Go, and Ruby.

---

## 1. Input Validation

### Philosophy: Allowlisting Over Denylisting

**Allowlisting** (accept known good): Define exactly what valid input looks like.
**Denylisting** (reject known bad): Try to block malicious patterns.

Always prefer allowlisting. Denylists are inherently incomplete — there's always a new bypass.

### JavaScript/TypeScript — Zod

```typescript
import { z } from 'zod';

// User registration schema
const registrationSchema = z.object({
  username: z.string()
    .min(3, 'Username must be at least 3 characters')
    .max(30, 'Username must be at most 30 characters')
    .regex(/^[a-zA-Z0-9_-]+$/, 'Username can only contain letters, numbers, hyphens, and underscores'),

  email: z.string()
    .email('Invalid email address')
    .max(255)
    .transform(v => v.toLowerCase().trim()),

  password: z.string()
    .min(12, 'Password must be at least 12 characters')
    .max(128, 'Password must be at most 128 characters'),

  age: z.number()
    .int('Age must be a whole number')
    .min(13, 'Must be at least 13 years old')
    .max(150, 'Invalid age')
    .optional(),

  role: z.enum(['user', 'moderator']).default('user'),
}).strict(); // Reject any unknown fields

// Usage in Express route
app.post('/register', async (req, res) => {
  const result = registrationSchema.safeParse(req.body);
  if (!result.success) {
    return res.status(400).json({
      errors: result.error.flatten().fieldErrors,
    });
  }
  // result.data is typed and validated
  const user = await createUser(result.data);
  res.status(201).json(user);
});

// URL parameter validation
const idSchema = z.string().uuid('Invalid ID format');

app.get('/users/:id', async (req, res) => {
  const idResult = idSchema.safeParse(req.params.id);
  if (!idResult.success) {
    return res.status(400).json({ error: 'Invalid user ID' });
  }
  const user = await User.findByPk(idResult.data);
  if (!user) return res.status(404).json({ error: 'Not found' });
  res.json(user);
});

// Query parameter validation
const searchSchema = z.object({
  q: z.string().max(200).trim(),
  page: z.coerce.number().int().min(1).default(1),
  limit: z.coerce.number().int().min(1).max(100).default(20),
  sort: z.enum(['name', 'date', 'relevance']).default('relevance'),
});

app.get('/search', async (req, res) => {
  const result = searchSchema.safeParse(req.query);
  if (!result.success) return res.status(400).json({ error: 'Invalid parameters' });
  // Use result.data.q, result.data.page, etc.
});
```

### Python — Pydantic

```python
from pydantic import BaseModel, EmailStr, Field, field_validator
from typing import Optional
from enum import Enum

class UserRole(str, Enum):
    user = "user"
    moderator = "moderator"

class RegistrationRequest(BaseModel):
    username: str = Field(min_length=3, max_length=30, pattern=r'^[a-zA-Z0-9_-]+$')
    email: EmailStr
    password: str = Field(min_length=12, max_length=128)
    age: Optional[int] = Field(None, ge=13, le=150)
    role: UserRole = UserRole.user

    model_config = {"extra": "forbid"}  # Reject unknown fields

    @field_validator('email')
    @classmethod
    def normalize_email(cls, v):
        return v.lower().strip()

# Usage in Flask
@app.route('/register', methods=['POST'])
def register():
    try:
        data = RegistrationRequest.model_validate(request.json)
    except ValidationError as e:
        return jsonify(errors=e.errors()), 400
    user = create_user(data)
    return jsonify(user.dict()), 201
```

### Java — Jakarta Bean Validation

```java
import jakarta.validation.constraints.*;

public record RegistrationRequest(
    @NotBlank @Size(min = 3, max = 30) @Pattern(regexp = "^[a-zA-Z0-9_-]+$")
    String username,

    @NotBlank @Email @Size(max = 255)
    String email,

    @NotBlank @Size(min = 12, max = 128)
    String password,

    @Min(13) @Max(150)
    Integer age
) {}

// Usage in Spring Boot
@PostMapping("/register")
public ResponseEntity<?> register(@Valid @RequestBody RegistrationRequest request) {
    User user = userService.createUser(request);
    return ResponseEntity.status(201).body(user);
}
```

### Go — struct tags + validator

```go
import "github.com/go-playground/validator/v10"

type RegistrationRequest struct {
    Username string `json:"username" validate:"required,min=3,max=30,alphanumdash"`
    Email    string `json:"email" validate:"required,email,max=255"`
    Password string `json:"password" validate:"required,min=12,max=128"`
    Age      *int   `json:"age,omitempty" validate:"omitempty,min=13,max=150"`
}

var validate = validator.New()

func registerHandler(w http.ResponseWriter, r *http.Request) {
    var req RegistrationRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }
    if err := validate.Struct(req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    // req is validated
}
```

---

## 2. Output Encoding

### Context-Specific Encoding

Different output contexts require different encoding strategies. Using the wrong encoding for a context is still vulnerable.

```
Context          | Encoding Type         | Example
-----------------|-----------------------|--------
HTML body        | HTML entity encoding  | &lt;script&gt;
HTML attributes  | HTML attribute encode | &quot;onclick&quot;
JavaScript       | JavaScript encoding   | \x3Cscript\x3E
URL parameters   | URL encoding          | %3Cscript%3E
CSS values       | CSS encoding          | \3C script\3E
JSON             | JSON serialization    | proper escaping
```

### JavaScript — Template Engine Configuration

```javascript
// Express + EJS — use <%= (escaped) not <%- (unescaped)
// template.ejs
// SAFE:
<p>Hello, <%= username %></p>

// VULNERABLE:
<p>Hello, <%- username %></p>

// Express + Handlebars — use {{ }} (escaped) not {{{ }}} (unescaped)
// SAFE:
<p>Hello, {{username}}</p>

// VULNERABLE:
<p>Hello, {{{username}}}</p>

// React — JSX auto-escapes by default
// SAFE:
function UserGreeting({ name }) {
  return <p>Hello, {name}</p>;  // Auto-escaped
}

// VULNERABLE — explicit opt-out:
function RichContent({ html }) {
  return <div dangerouslySetInnerHTML={{ __html: html }} />;  // NOT escaped
}

// SAFE with sanitization:
import DOMPurify from 'dompurify';

function RichContent({ html }) {
  const clean = DOMPurify.sanitize(html, {
    ALLOWED_TAGS: ['b', 'i', 'em', 'strong', 'a', 'p', 'br', 'ul', 'ol', 'li'],
    ALLOWED_ATTR: ['href', 'title'],
  });
  return <div dangerouslySetInnerHTML={{ __html: clean }} />;
}
```

### URL Encoding

```javascript
// JavaScript — URL encoding for user data in URLs
const searchUrl = `/search?q=${encodeURIComponent(userInput)}`;

// For full URLs with multiple parameters:
const url = new URL('https://example.com/search');
url.searchParams.set('q', userInput);  // Automatically encoded
url.searchParams.set('page', '1');
const safeUrl = url.toString();
```

```python
# Python — URL encoding
from urllib.parse import urlencode, quote

# For query parameters
params = urlencode({'q': user_input, 'page': 1})
url = f'https://example.com/search?{params}'

# For URL path components
safe_path = quote(user_input, safe='')
url = f'https://example.com/files/{safe_path}'
```

### JSON Encoding

```javascript
// ALWAYS use JSON.stringify, never string concatenation
// VULNERABLE:
const response = `{"name": "${userName}", "role": "${userRole}"}`;

// SAFE:
const response = JSON.stringify({ name: userName, role: userRole });
```

---

## 3. Parameterized Queries & Prepared Statements

### Node.js + PostgreSQL (pg)

```javascript
import pg from 'pg';
const pool = new pg.Pool({ connectionString: process.env.DATABASE_URL });

// Parameterized query — positional parameters
async function getUserByEmail(email) {
  const result = await pool.query(
    'SELECT id, name, email FROM users WHERE email = $1',
    [email]
  );
  return result.rows[0];
}

// Parameterized INSERT
async function createUser(name, email, passwordHash) {
  const result = await pool.query(
    'INSERT INTO users (name, email, password_hash) VALUES ($1, $2, $3) RETURNING id',
    [name, email, passwordHash]
  );
  return result.rows[0];
}

// Parameterized query with LIKE (safe way to handle wildcards)
async function searchUsers(term) {
  const result = await pool.query(
    'SELECT id, name, email FROM users WHERE name ILIKE $1 LIMIT 50',
    [`%${term}%`]  // Wildcards go in the parameter value, not the query
  );
  return result.rows;
}

// Parameterized IN clause
async function getUsersByIds(ids) {
  const placeholders = ids.map((_, i) => `$${i + 1}`).join(', ');
  const result = await pool.query(
    `SELECT id, name, email FROM users WHERE id IN (${placeholders})`,
    ids
  );
  return result.rows;
}
```

### Python + SQLAlchemy

```python
from sqlalchemy import text

# Parameterized query with named parameters
def get_user_by_email(session, email):
    result = session.execute(
        text("SELECT id, name, email FROM users WHERE email = :email"),
        {"email": email}
    )
    return result.fetchone()

# ORM — automatically parameterized
def get_user_by_email_orm(session, email):
    return session.query(User).filter(User.email == email).first()

# ORM — safe LIKE query
def search_users(session, term):
    return session.query(User).filter(User.name.ilike(f"%{term}%")).limit(50).all()
```

### Java + JDBC

```java
// PreparedStatement — always use for user input
public User getUserByEmail(String email) throws SQLException {
    String sql = "SELECT id, name, email FROM users WHERE email = ?";
    try (PreparedStatement stmt = connection.prepareStatement(sql)) {
        stmt.setString(1, email);
        try (ResultSet rs = stmt.executeQuery()) {
            if (rs.next()) {
                return new User(rs.getLong("id"), rs.getString("name"), rs.getString("email"));
            }
        }
    }
    return null;
}
```

---

## 4. Secure Session Management

### Express + express-session

```javascript
import session from 'express-session';
import RedisStore from 'connect-redis';
import { createClient } from 'redis';

// Redis client for session store
const redisClient = createClient({ url: process.env.REDIS_URL });
await redisClient.connect();

app.use(session({
  store: new RedisStore({ client: redisClient }),
  name: '__Host-sid',                   // __Host- prefix enforces Secure + path=/
  secret: process.env.SESSION_SECRET,   // Strong random secret from env
  resave: false,                        // Don't save session if unmodified
  saveUninitialized: false,             // Don't create session until stored
  rolling: true,                        // Reset expiry on each request
  cookie: {
    httpOnly: true,                     // Not accessible via JavaScript
    secure: true,                       // Only sent over HTTPS
    sameSite: 'strict',                 // Not sent on cross-origin requests
    maxAge: 3600000,                    // 1 hour absolute timeout
    path: '/',
  },
}));

// Session regeneration after authentication
app.post('/login', async (req, res) => {
  const user = await authenticate(req.body.email, req.body.password);
  if (!user) return res.status(401).json({ error: 'Invalid credentials' });

  // CRITICAL: Regenerate session ID to prevent session fixation
  req.session.regenerate((err) => {
    if (err) return next(err);
    req.session.userId = user.id;
    req.session.loginTime = Date.now();
    req.session.save(() => res.json({ success: true }));
  });
});

// Complete session destruction on logout
app.post('/logout', (req, res) => {
  const sessionId = req.session.id;
  req.session.destroy(async (err) => {
    if (err) console.error('Session destruction error:', err);
    // Also delete from Redis to ensure server-side invalidation
    await redisClient.del(`sess:${sessionId}`);
    res.clearCookie('__Host-sid');
    res.json({ success: true });
  });
});

// Idle timeout middleware
function checkIdleTimeout(req, res, next) {
  if (req.session.lastActivity) {
    const idleTime = Date.now() - req.session.lastActivity;
    if (idleTime > 15 * 60 * 1000) { // 15 minutes idle
      return req.session.destroy(() => {
        res.clearCookie('__Host-sid');
        res.status(401).json({ error: 'Session expired due to inactivity' });
      });
    }
  }
  req.session.lastActivity = Date.now();
  next();
}
```

---

## 5. CORS Configuration

### Correct CORS Setup

```javascript
import cors from 'cors';

// Production CORS — strict origin allowlist
const corsOptions = {
  origin: (origin, callback) => {
    const allowedOrigins = [
      'https://myapp.com',
      'https://admin.myapp.com',
      'https://staging.myapp.com',
    ];

    // Allow requests with no origin (mobile apps, curl, server-to-server)
    if (!origin) return callback(null, true);

    if (allowedOrigins.includes(origin)) {
      callback(null, true);
    } else {
      callback(new Error('Not allowed by CORS'));
    }
  },
  methods: ['GET', 'POST', 'PUT', 'DELETE', 'PATCH'],
  allowedHeaders: ['Content-Type', 'Authorization', 'X-Request-ID'],
  exposedHeaders: ['X-Request-ID', 'X-RateLimit-Remaining'],
  credentials: true,       // Allow cookies
  maxAge: 86400,           // Cache preflight for 24 hours
  optionsSuccessStatus: 200,
};

app.use(cors(corsOptions));
```

### CORS Anti-Patterns

```javascript
// DANGEROUS — wildcard origin
app.use(cors());  // Allows ALL origins

// DANGEROUS — reflecting any origin with credentials
app.use(cors({
  origin: true,         // Reflects the requesting origin
  credentials: true,    // With cookies = any site can make authenticated requests
}));

// DANGEROUS — wildcard with credentials
res.setHeader('Access-Control-Allow-Origin', '*');
res.setHeader('Access-Control-Allow-Credentials', 'true');
// Browsers block this combination, but misconfigured proxies may not
```

---

## 6. Content Security Policy (CSP)

### Strict CSP Configuration

```javascript
// Express CSP with nonce-based script loading
import crypto from 'crypto';

app.use((req, res, next) => {
  // Generate a unique nonce for each request
  const nonce = crypto.randomBytes(16).toString('base64');
  res.locals.nonce = nonce;

  res.setHeader('Content-Security-Policy', [
    `default-src 'self'`,
    `script-src 'self' 'nonce-${nonce}'`,           // Only scripts with our nonce
    `style-src 'self' 'nonce-${nonce}'`,             // Only styles with our nonce
    `img-src 'self' data: https:`,                   // Images from self, data URIs, HTTPS
    `font-src 'self' https://fonts.gstatic.com`,     // Google Fonts
    `connect-src 'self' https://api.myapp.com`,      // API connections
    `frame-ancestors 'none'`,                        // No framing (clickjacking protection)
    `base-uri 'self'`,                               // Restrict <base> tag
    `form-action 'self'`,                            // Forms only submit to self
    `object-src 'none'`,                             // No plugins (Flash, Java)
    `upgrade-insecure-requests`,                     // Auto-upgrade HTTP to HTTPS
  ].join('; '));

  next();
});

// In templates, use the nonce:
// <script nonce="<%= nonce %>">...</script>
```

### CSP for React/SPA

```javascript
// For React apps with inline scripts (create-react-app hashes)
const csp = [
  "default-src 'self'",
  "script-src 'self'",              // No inline scripts in production build
  "style-src 'self' 'unsafe-inline'", // CSS-in-JS may need unsafe-inline
  "img-src 'self' data: https:",
  "connect-src 'self' https://api.myapp.com wss://ws.myapp.com",
  "font-src 'self'",
  "frame-ancestors 'none'",
  "base-uri 'self'",
  "form-action 'self'",
].join('; ');
```

---

## 7. Cookie Security

### Secure Cookie Configuration

```javascript
// Set cookies with all security attributes
res.cookie('preference', 'dark-mode', {
  httpOnly: false,        // OK for non-sensitive cookies accessed by JS
  secure: true,           // Only sent over HTTPS
  sameSite: 'lax',        // Sent on top-level navigations
  maxAge: 30 * 24 * 3600000, // 30 days
  path: '/',
  domain: '.myapp.com',  // Available on subdomains
});

// Session/auth cookies — maximum security
res.cookie('session', sessionToken, {
  httpOnly: true,         // MUST be true — not accessible via JS
  secure: true,           // MUST be true — HTTPS only
  sameSite: 'strict',     // MUST be strict — no cross-site sending
  maxAge: 3600000,        // 1 hour
  path: '/',
  // Use __Host- prefix for maximum cookie security:
  // - Requires Secure attribute
  // - Requires Path=/
  // - Must not have Domain attribute
  // - Cannot be overridden by subdomains
});

// CSRF token cookie — accessible to JS but SameSite protected
res.cookie('csrf-token', csrfToken, {
  httpOnly: false,        // JS needs to read this for CSRF headers
  secure: true,
  sameSite: 'strict',
  maxAge: 3600000,
  path: '/',
});
```

### Cookie Security Attributes Reference

```
Attribute  | Purpose                          | Recommended
-----------|----------------------------------|-----------------------------
HttpOnly   | Block JavaScript access          | true for auth/session cookies
Secure     | HTTPS only                       | Always true in production
SameSite   | Cross-site request control        | strict for auth, lax for others
Path       | URL path scope                   | / (unless scoping needed)
Domain     | Domain scope                     | Omit (defaults to exact origin)
Max-Age    | Lifetime in seconds              | As short as practical
__Host-    | Name prefix enforcing security   | Use for critical cookies
__Secure-  | Name prefix requiring Secure     | Use for important cookies
```

---

## 8. Secure File Operations

### Path Traversal Prevention

```javascript
import path from 'path';
import fs from 'fs/promises';

const UPLOAD_DIR = path.resolve('/app/uploads');

// SECURE file access
async function getFile(userFilename) {
  // Resolve the full path
  const filePath = path.resolve(UPLOAD_DIR, userFilename);

  // CRITICAL: Verify the resolved path is within the allowed directory
  if (!filePath.startsWith(UPLOAD_DIR + path.sep) && filePath !== UPLOAD_DIR) {
    throw new Error('Access denied: path traversal attempt');
  }

  // Verify file exists
  try {
    await fs.access(filePath);
  } catch {
    throw new Error('File not found');
  }

  return filePath;
}

// SECURE file upload
async function saveUpload(file) {
  // Generate random filename — never use user-supplied filename
  const ext = path.extname(file.originalname).toLowerCase();
  const allowedExtensions = ['.jpg', '.jpeg', '.png', '.gif', '.pdf'];

  if (!allowedExtensions.includes(ext)) {
    throw new Error('File type not allowed');
  }

  const filename = `${crypto.randomUUID()}${ext}`;
  const filepath = path.join(UPLOAD_DIR, filename);

  // Verify we're still in the upload directory
  if (!filepath.startsWith(UPLOAD_DIR + path.sep)) {
    throw new Error('Invalid filename');
  }

  await fs.writeFile(filepath, file.buffer);
  return filename;
}
```

```python
# Python — safe file access
import os

UPLOAD_DIR = os.path.realpath('/app/uploads')

def safe_file_path(user_filename):
    """Resolve user filename and verify it's within the upload directory."""
    # Resolve to absolute path
    filepath = os.path.realpath(os.path.join(UPLOAD_DIR, user_filename))

    # Verify containment
    if not filepath.startswith(UPLOAD_DIR + os.sep) and filepath != UPLOAD_DIR:
        raise ValueError("Path traversal attempt detected")

    if not os.path.exists(filepath):
        raise FileNotFoundError("File not found")

    return filepath
```

---

## 9. Safe Deserialization

### JSON — Safe by Default

```javascript
// JSON.parse is safe — it cannot execute code
const data = JSON.parse(userInput);

// But VALIDATE the structure after parsing
const schema = z.object({
  name: z.string(),
  age: z.number().int().positive(),
});
const validated = schema.parse(data);
```

### YAML — Use Safe Load

```python
# VULNERABLE — yaml.load can execute arbitrary Python
import yaml
data = yaml.load(user_input)  # RCE possible!

# SECURE — yaml.safe_load only creates basic Python types
data = yaml.safe_load(user_input)  # No code execution
```

```javascript
// Node.js — js-yaml safe by default since v4
import yaml from 'js-yaml';
const data = yaml.load(userInput);  // Safe in v4+

// For v3: explicitly use safeLoad
const data = yaml.safeLoad(userInput);
```

### Never Deserialize Untrusted Data With Unsafe Formats

```
NEVER use with untrusted input:
- Python: pickle, marshal, shelve
- Ruby: Marshal.load, YAML.load (use YAML.safe_load)
- Java: ObjectInputStream (without type filtering)
- PHP: unserialize (use json_decode)
- Node.js: node-serialize (removed, but may exist in legacy code)

SAFE alternatives:
- JSON (all languages)
- Protocol Buffers
- MessagePack (with type restrictions)
- YAML safe_load
```

---

## 10. Cryptographic Best Practices

### Password Hashing

```javascript
// Node.js — bcrypt
import bcrypt from 'bcrypt';

const COST_FACTOR = 12; // Adjust based on server speed (target: 250ms-500ms)

async function hashPassword(password) {
  return bcrypt.hash(password, COST_FACTOR);
}

async function verifyPassword(password, hash) {
  return bcrypt.compare(password, hash);
}
```

```python
# Python — argon2-cffi (recommended by OWASP)
from argon2 import PasswordHasher

ph = PasswordHasher(
    time_cost=2,         # Number of iterations
    memory_cost=19456,   # Memory in KiB (19 MiB)
    parallelism=1,       # Number of threads
)

hashed = ph.hash("password")
try:
    ph.verify(hashed, "password")
except argon2.exceptions.VerifyMismatchError:
    pass  # Wrong password
```

### Symmetric Encryption — AES-256-GCM

```javascript
import crypto from 'crypto';

const ALGORITHM = 'aes-256-gcm';
const KEY_LENGTH = 32;  // 256 bits
const IV_LENGTH = 12;   // 96 bits for GCM (NIST recommended)
const TAG_LENGTH = 16;  // 128 bits auth tag

function encrypt(plaintext, key) {
  const iv = crypto.randomBytes(IV_LENGTH);
  const cipher = crypto.createCipheriv(ALGORITHM, key, iv, { authTagLength: TAG_LENGTH });

  let encrypted = cipher.update(plaintext, 'utf8');
  encrypted = Buffer.concat([encrypted, cipher.final()]);
  const tag = cipher.getAuthTag();

  // Concatenate: IV + Tag + Ciphertext
  return Buffer.concat([iv, tag, encrypted]);
}

function decrypt(data, key) {
  const iv = data.subarray(0, IV_LENGTH);
  const tag = data.subarray(IV_LENGTH, IV_LENGTH + TAG_LENGTH);
  const ciphertext = data.subarray(IV_LENGTH + TAG_LENGTH);

  const decipher = crypto.createDecipheriv(ALGORITHM, key, iv, { authTagLength: TAG_LENGTH });
  decipher.setAuthTag(tag);

  let decrypted = decipher.update(ciphertext);
  decrypted = Buffer.concat([decrypted, decipher.final()]);
  return decrypted.toString('utf8');
}
```

### Random Number Generation

```javascript
// Node.js — ALWAYS use crypto module for security
import crypto from 'crypto';

// Random bytes
const randomBytes = crypto.randomBytes(32);

// Random UUID
const uuid = crypto.randomUUID();

// Random integer in range [min, max)
function randomInt(min, max) {
  return crypto.randomInt(min, max);
}

// Random token (URL-safe)
function generateToken(length = 32) {
  return crypto.randomBytes(length).toString('base64url');
}

// NEVER use for security:
// Math.random()              ← Predictable
// Date.now()                 ← Predictable
// process.pid                ← Predictable
// Sequential counters        ← Predictable
```

```python
# Python — use secrets module
import secrets

# Random bytes
random_bytes = secrets.token_bytes(32)

# Random URL-safe token
token = secrets.token_urlsafe(32)

# Random hex token
hex_token = secrets.token_hex(32)

# Random integer
random_int = secrets.randbelow(1000000)

# NEVER use for security:
# random.random()             ← Predictable
# random.randint()            ← Predictable
# uuid.uuid1()                ← Contains MAC address and timestamp
```

### Digital Signatures

```javascript
// Ed25519 — fast, secure, recommended
import crypto from 'crypto';

// Generate key pair
const { publicKey, privateKey } = crypto.generateKeyPairSync('ed25519');

// Sign
const signature = crypto.sign(null, Buffer.from(message), privateKey);

// Verify
const isValid = crypto.verify(null, Buffer.from(message), publicKey, signature);
```

### HMAC for Message Authentication

```javascript
import crypto from 'crypto';

function createHMAC(message, key) {
  return crypto.createHmac('sha256', key).update(message).digest('hex');
}

function verifyHMAC(message, key, expectedHmac) {
  const computed = createHMAC(message, key);
  // CRITICAL: Use timing-safe comparison to prevent timing attacks
  return crypto.timingSafeEqual(
    Buffer.from(computed, 'hex'),
    Buffer.from(expectedHmac, 'hex')
  );
}
```

---

## 11. Secret Management

### Environment Variables — Basic Pattern

```javascript
// Load from .env in development only
if (process.env.NODE_ENV !== 'production') {
  const dotenv = await import('dotenv');
  dotenv.config();
}

// Validate required secrets at startup
const requiredSecrets = [
  'DATABASE_URL',
  'SESSION_SECRET',
  'JWT_SECRET',
  'ENCRYPTION_KEY',
  'STRIPE_SECRET_KEY',
];

for (const secret of requiredSecrets) {
  if (!process.env[secret]) {
    console.error(`FATAL: Missing required environment variable: ${secret}`);
    process.exit(1);
  }
}
```

### .env File Security

```bash
# .gitignore — NEVER commit .env files
.env
.env.local
.env.production
.env.*.local

# Provide a template for developers
# .env.example (commit this)
DATABASE_URL=postgresql://user:password@localhost:5432/dbname
SESSION_SECRET=generate-a-random-64-char-string
JWT_SECRET=generate-a-random-64-char-string
ENCRYPTION_KEY=generate-a-random-64-hex-char-string
STRIPE_SECRET_KEY=sk_test_your_stripe_key
```

### Secret Rotation

```javascript
// Support multiple key versions for rotation
const ENCRYPTION_KEYS = {
  v1: Buffer.from(process.env.ENCRYPTION_KEY_V1, 'hex'),
  v2: Buffer.from(process.env.ENCRYPTION_KEY_V2, 'hex'),  // Current
};
const CURRENT_KEY_VERSION = 'v2';

function encrypt(plaintext) {
  const key = ENCRYPTION_KEYS[CURRENT_KEY_VERSION];
  const encrypted = doEncrypt(plaintext, key);
  return `${CURRENT_KEY_VERSION}:${encrypted}`;  // Prefix with version
}

function decrypt(data) {
  const [version, ciphertext] = data.split(':');
  const key = ENCRYPTION_KEYS[version];
  if (!key) throw new Error(`Unknown key version: ${version}`);
  return doDecrypt(ciphertext, key);
}

// Re-encryption migration (run periodically)
async function migrateEncryptedData() {
  const records = await db.query('SELECT id, data FROM secrets WHERE data NOT LIKE $1',
    [`${CURRENT_KEY_VERSION}:%`]);
  for (const record of records) {
    const decrypted = decrypt(record.data);
    const reEncrypted = encrypt(decrypted);
    await db.query('UPDATE secrets SET data = $1 WHERE id = $2',
      [reEncrypted, record.id]);
  }
}
```

---

## 12. Security Headers Quick Reference

```javascript
// Complete security headers middleware
function securityHeaders(req, res, next) {
  // Prevent XSS
  res.setHeader('X-Content-Type-Options', 'nosniff');
  res.setHeader('X-XSS-Protection', '0');  // Disabled — CSP is better

  // Prevent clickjacking
  res.setHeader('X-Frame-Options', 'DENY');

  // Enforce HTTPS
  res.setHeader('Strict-Transport-Security', 'max-age=31536000; includeSubDomains; preload');

  // Control referrer information
  res.setHeader('Referrer-Policy', 'strict-origin-when-cross-origin');

  // Disable dangerous browser features
  res.setHeader('Permissions-Policy',
    'camera=(), microphone=(), geolocation=(), payment=(), usb=(), bluetooth=()');

  // Content Security Policy (customize per app)
  res.setHeader('Content-Security-Policy',
    "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; " +
    "img-src 'self' data: https:; font-src 'self'; connect-src 'self'; " +
    "frame-ancestors 'none'; base-uri 'self'; form-action 'self'");

  // Prevent MIME type confusion
  res.setHeader('X-Content-Type-Options', 'nosniff');

  // Cross-Origin policies
  res.setHeader('Cross-Origin-Opener-Policy', 'same-origin');
  res.setHeader('Cross-Origin-Resource-Policy', 'same-origin');

  next();
}

app.use(securityHeaders);

// Or use helmet.js which handles most of this:
import helmet from 'helmet';
app.use(helmet());
```

---

## 13. Rate Limiting Patterns

```javascript
import rateLimit from 'express-rate-limit';

// General API rate limit
const apiLimiter = rateLimit({
  windowMs: 60 * 1000,     // 1 minute
  max: 100,                 // 100 requests per minute
  standardHeaders: true,    // Return rate limit info in RateLimit-* headers
  legacyHeaders: false,
  message: { error: 'Too many requests' },
});
app.use('/api/', apiLimiter);

// Strict rate limit for authentication
const authLimiter = rateLimit({
  windowMs: 15 * 60 * 1000,  // 15 minutes
  max: 5,                     // 5 attempts
  message: { error: 'Too many login attempts. Try again in 15 minutes.' },
  keyGenerator: (req) => req.body?.email || req.ip,
  skipSuccessfulRequests: true,  // Don't count successful logins
});
app.use('/api/login', authLimiter);
app.use('/api/register', authLimiter);

// Strict rate limit for password reset
const resetLimiter = rateLimit({
  windowMs: 60 * 60 * 1000,  // 1 hour
  max: 3,                     // 3 attempts per hour
  message: { error: 'Too many password reset attempts' },
  keyGenerator: (req) => req.body?.email || req.ip,
});
app.use('/api/password-reset', resetLimiter);
```

---

## 14. Constant-Time Comparison

Timing attacks can leak secret values by measuring response times. Always use constant-time comparison for security-critical values.

```javascript
import crypto from 'crypto';

// VULNERABLE — early return leaks information via timing
function unsafeCompare(a, b) {
  if (a.length !== b.length) return false;
  for (let i = 0; i < a.length; i++) {
    if (a[i] !== b[i]) return false;  // Returns early on first mismatch
  }
  return true;
}

// SECURE — constant-time comparison
function safeCompare(a, b) {
  if (typeof a !== 'string' || typeof b !== 'string') return false;
  const bufA = Buffer.from(a);
  const bufB = Buffer.from(b);
  if (bufA.length !== bufB.length) {
    // Still do comparison to prevent length-based timing leak
    crypto.timingSafeEqual(bufA, bufA);
    return false;
  }
  return crypto.timingSafeEqual(bufA, bufB);
}
```

```python
# Python — use hmac.compare_digest
import hmac

# SECURE — constant-time comparison
def safe_compare(a: str, b: str) -> bool:
    return hmac.compare_digest(a.encode(), b.encode())
```
