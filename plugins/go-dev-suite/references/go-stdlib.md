# Go Standard Library Deep Dive

Practical reference for the most important standard library packages with real-world patterns.

## net/http — HTTP Server and Client

### Server with Go 1.22+ Enhanced Routing

```go
mux := http.NewServeMux()

// Method-specific routing (Go 1.22+)
mux.HandleFunc("GET /api/users", listUsers)
mux.HandleFunc("POST /api/users", createUser)
mux.HandleFunc("GET /api/users/{id}", getUser)
mux.HandleFunc("PUT /api/users/{id}", updateUser)
mux.HandleFunc("DELETE /api/users/{id}", deleteUser)

// Wildcard path (Go 1.22+)
mux.HandleFunc("GET /files/{path...}", serveFile)

// Exact match vs prefix
mux.HandleFunc("GET /api/health", healthCheck)   // Exact: /api/health only
mux.HandleFunc("GET /static/", serveStatic)       // Prefix: /static/* (trailing slash)

// Host-specific routes
mux.HandleFunc("GET api.example.com/v1/data", apiHandler)

// Path value extraction (Go 1.22+)
func getUser(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")
    // ...
}

func serveFile(w http.ResponseWriter, r *http.Request) {
    path := r.PathValue("path") // Everything after /files/
    // ...
}
```

### Production Server Configuration

```go
srv := &http.Server{
    Addr:              ":8080",
    Handler:           mux,
    ReadTimeout:       5 * time.Second,   // Time to read request headers + body
    ReadHeaderTimeout: 2 * time.Second,   // Time to read just headers
    WriteTimeout:      10 * time.Second,  // Time to write response
    IdleTimeout:       120 * time.Second, // Keep-alive timeout
    MaxHeaderBytes:    1 << 20,           // 1MB max header size
}
```

### HTTP Client Best Practices

```go
// Create a shared client — never use http.DefaultClient in production
client := &http.Client{
    Timeout: 30 * time.Second,
    Transport: &http.Transport{
        DialContext: (&net.Dialer{
            Timeout:   5 * time.Second,
            KeepAlive: 30 * time.Second,
        }).DialContext,
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
        TLSHandshakeTimeout: 5 * time.Second,
        ExpectContinueTimeout: 1 * time.Second,
    },
}

// Making requests with context
func fetchData(ctx context.Context, url string) ([]byte, error) {
    req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
    if err != nil {
        return nil, fmt.Errorf("creating request: %w", err)
    }
    req.Header.Set("Accept", "application/json")
    req.Header.Set("User-Agent", "myapp/1.0")

    resp, err := client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("executing request: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
        return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, body)
    }

    // Always limit response body reads
    return io.ReadAll(io.LimitReader(resp.Body, 10<<20)) // 10MB max
}

// POST with JSON body
func postJSON(ctx context.Context, url string, payload any) (*http.Response, error) {
    body, err := json.Marshal(payload)
    if err != nil {
        return nil, fmt.Errorf("marshaling payload: %w", err)
    }

    req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
    if err != nil {
        return nil, err
    }
    req.Header.Set("Content-Type", "application/json")

    return client.Do(req)
}
```

## encoding/json — JSON Encoding/Decoding

### Struct Tags

```go
type User struct {
    ID        string    `json:"id"`
    Name      string    `json:"name"`
    Email     string    `json:"email"`
    Age       int       `json:"age,omitempty"`        // Omit if zero value
    IsAdmin   bool      `json:"is_admin"`
    CreatedAt time.Time `json:"created_at"`
    Password  string    `json:"-"`                     // Never serialize
    Metadata  any       `json:"metadata,omitempty"`
}
```

### Custom JSON Marshal/Unmarshal

```go
// Duration that serializes as string
type Duration time.Duration

func (d Duration) MarshalJSON() ([]byte, error) {
    return json.Marshal(time.Duration(d).String())
}

func (d *Duration) UnmarshalJSON(data []byte) error {
    var s string
    if err := json.Unmarshal(data, &s); err != nil {
        return err
    }
    parsed, err := time.ParseDuration(s)
    if err != nil {
        return err
    }
    *d = Duration(parsed)
    return nil
}

// Epoch timestamp
type EpochTime time.Time

func (t EpochTime) MarshalJSON() ([]byte, error) {
    return json.Marshal(time.Time(t).Unix())
}

func (t *EpochTime) UnmarshalJSON(data []byte) error {
    var epoch int64
    if err := json.Unmarshal(data, &epoch); err != nil {
        return err
    }
    *t = EpochTime(time.Unix(epoch, 0))
    return nil
}
```

### Streaming JSON

```go
// Decode array of objects one at a time (memory-efficient)
func decodeStream(r io.Reader) iter.Seq2[User, error] {
    return func(yield func(User, error) bool) {
        dec := json.NewDecoder(r)
        // Read opening bracket
        if _, err := dec.Token(); err != nil {
            yield(User{}, err)
            return
        }
        for dec.More() {
            var u User
            if err := dec.Decode(&u); err != nil {
                yield(User{}, err)
                return
            }
            if !yield(u, nil) {
                return
            }
        }
    }
}

// Encode as NDJSON (newline-delimited JSON) for streaming
func encodeNDJSON(w io.Writer, items []Event) error {
    enc := json.NewEncoder(w)
    for _, item := range items {
        if err := enc.Encode(item); err != nil {
            return err
        }
    }
    return nil
}
```

## context — Request-Scoped Data and Cancellation

### Context Best Practices

```go
// Always pass context as the first parameter
func (s *Service) GetUser(ctx context.Context, id string) (*User, error) {
    // Check if already cancelled before doing work
    if err := ctx.Err(); err != nil {
        return nil, err
    }
    return s.repo.GetByID(ctx, id)
}

// Timeout for specific operations
func (s *Service) ProcessOrder(ctx context.Context, order *Order) error {
    // Payment has a stricter timeout
    payCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    if err := s.payments.Charge(payCtx, order.Total); err != nil {
        return fmt.Errorf("charging payment: %w", err)
    }
    return nil
}

// Context values — use typed keys, never strings
type contextKey int

const (
    userIDKey contextKey = iota
    requestIDKey
    traceIDKey
)

func WithUserID(ctx context.Context, id string) context.Context {
    return context.WithValue(ctx, userIDKey, id)
}

func UserIDFrom(ctx context.Context) (string, bool) {
    id, ok := ctx.Value(userIDKey).(string)
    return id, ok
}
```

## io — Input/Output Primitives

### Composing Readers and Writers

```go
// Chain readers
combined := io.MultiReader(header, body, footer)

// Tee: read and write simultaneously
var buf bytes.Buffer
tee := io.TeeReader(resp.Body, &buf)
// Reading from tee also writes to buf

// Limit reader: prevent memory exhaustion
limited := io.LimitReader(resp.Body, 10<<20) // 10MB max

// Pipe: connect writer to reader
pr, pw := io.Pipe()
go func() {
    defer pw.Close()
    json.NewEncoder(pw).Encode(data)
}()
// pr can now be used as request body
req, _ := http.NewRequest("POST", url, pr)

// NopCloser: add Close() to a Reader
rc := io.NopCloser(strings.NewReader("hello"))

// Copy with buffer
buf := make([]byte, 32*1024) // 32KB buffer
written, err := io.CopyBuffer(dst, src, buf)

// ReadAll with limit (safer than io.ReadAll)
data, err := io.ReadAll(io.LimitReader(r, maxSize))
```

## time — Time and Duration

### Time Operations

```go
// Parse and format
t, err := time.Parse(time.RFC3339, "2024-01-15T10:30:00Z")
formatted := t.Format("2006-01-02 15:04:05") // Go's reference time

// Common format constants
time.RFC3339       // "2006-01-02T15:04:05Z07:00"
time.RFC3339Nano   // "2006-01-02T15:04:05.999999999Z07:00"
time.DateTime      // "2006-01-02 15:04:05" (Go 1.20+)
time.DateOnly      // "2006-01-02" (Go 1.20+)
time.TimeOnly      // "15:04:05" (Go 1.20+)

// Duration arithmetic
deadline := time.Now().Add(30 * time.Minute)
elapsed := time.Since(start)
remaining := time.Until(deadline)

// Ticker for periodic work
ticker := time.NewTicker(5 * time.Second)
defer ticker.Stop()
for {
    select {
    case <-ctx.Done():
        return
    case <-ticker.C:
        doPeriodicWork()
    }
}

// Timer for one-shot delayed work
timer := time.NewTimer(10 * time.Second)
defer timer.Stop()
select {
case <-ctx.Done():
    return ctx.Err()
case <-timer.C:
    return ErrTimeout
case result := <-resultCh:
    return result
}
```

### Testable Time (Inject Clock)

```go
// Clock interface for testable code
type Clock interface {
    Now() time.Time
    After(d time.Duration) <-chan time.Time
}

type RealClock struct{}

func (RealClock) Now() time.Time                         { return time.Now() }
func (RealClock) After(d time.Duration) <-chan time.Time  { return time.After(d) }

type FakeClock struct {
    current time.Time
}

func (c *FakeClock) Now() time.Time                        { return c.current }
func (c *FakeClock) After(d time.Duration) <-chan time.Time {
    ch := make(chan time.Time, 1)
    ch <- c.current.Add(d)
    return ch
}
func (c *FakeClock) Advance(d time.Duration) { c.current = c.current.Add(d) }
```

## log/slog — Structured Logging (Go 1.21+)

```go
// JSON handler for production
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level:     slog.LevelInfo,
    AddSource: true,
    ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
        // Rename "msg" to "message" for compatibility
        if a.Key == slog.MessageKey {
            a.Key = "message"
        }
        // Format time as RFC3339
        if a.Key == slog.TimeKey {
            a.Value = slog.StringValue(a.Value.Time().Format(time.RFC3339))
        }
        return a
    },
}))

// Text handler for development
devLogger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
    Level: slog.LevelDebug,
}))

// Set as default
slog.SetDefault(logger)

// Logging with typed attributes (fastest)
slog.Info("request processed",
    slog.String("method", "GET"),
    slog.String("path", "/api/users"),
    slog.Int("status", 200),
    slog.Duration("latency", 42*time.Millisecond),
    slog.Group("user",
        slog.String("id", "user-1"),
        slog.String("role", "admin"),
    ),
)

// Logger with persistent attributes
reqLogger := logger.With(
    slog.String("request_id", requestID),
    slog.String("trace_id", traceID),
)
reqLogger.Info("handling request")
reqLogger.Error("request failed", "err", err)

// LogValuer interface for custom types
type User struct {
    ID    string
    Name  string
    Email string
    // Password should never be logged
}

func (u User) LogValue() slog.Value {
    return slog.GroupValue(
        slog.String("id", u.ID),
        slog.String("name", u.Name),
        // Email and Password intentionally omitted
    )
}

// Usage: slog.Info("user created", "user", user)
// Output: {"level":"INFO","msg":"user created","user":{"id":"...","name":"..."}}
```

## sync — Synchronization Primitives

### sync.OnceValue / sync.OnceValues (Go 1.21+)

```go
// Lazy singleton initialization
var getDB = sync.OnceValues(func() (*sql.DB, error) {
    return sql.Open("postgres", os.Getenv("DATABASE_URL"))
})

// Thread-safe config loading
var loadConfig = sync.OnceValue(func() *Config {
    cfg, err := parseConfig("config.yaml")
    if err != nil {
        panic(fmt.Sprintf("failed to load config: %v", err))
    }
    return cfg
})
```

## slices — Slice Operations (Go 1.21+)

```go
import "slices"

// Sort
slices.Sort(numbers)
slices.SortFunc(users, func(a, b User) int {
    return cmp.Compare(a.Name, b.Name)
})

// Stable sort preserves order of equal elements
slices.SortStableFunc(users, func(a, b User) int {
    return cmp.Compare(a.Priority, b.Priority)
})

// Binary search in sorted slice
idx, found := slices.BinarySearch(sorted, target)
idx, found := slices.BinarySearchFunc(users, "Alice", func(u User, name string) int {
    return cmp.Compare(u.Name, name)
})

// Contains
if slices.Contains(roles, "admin") { /* ... */ }

// Index
idx := slices.Index(items, target) // -1 if not found

// Compact: remove consecutive duplicates
unique := slices.Compact(sorted)

// Reverse
slices.Reverse(items)

// Clone
copy := slices.Clone(original)

// Concat (Go 1.22+)
all := slices.Concat(slice1, slice2, slice3)

// Max/Min
biggest := slices.Max(numbers)
smallest := slices.Min(numbers)
```

## maps — Map Operations (Go 1.21+)

```go
import "maps"

// Keys and Values
keys := maps.Keys(m)   // Returns iter.Seq[K]
vals := maps.Values(m)  // Returns iter.Seq[V]

// Collect keys into slice
keySlice := slices.Collect(maps.Keys(m))

// Clone
copy := maps.Clone(original)

// Copy entries from src to dst
maps.Copy(dst, src) // Overwrites existing keys

// Delete matching entries
maps.DeleteFunc(m, func(k string, v int) bool {
    return v < 0
})

// Equal
maps.Equal(m1, m2)

// Collect from iter.Seq2
m := maps.Collect(someIterator)
```

## database/sql — Database Access

### Connection Management

```go
db, err := sql.Open("postgres", connStr)
if err != nil {
    return nil, err
}

// Configure connection pool
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(5)
db.SetConnMaxLifetime(5 * time.Minute)
db.SetConnMaxIdleTime(1 * time.Minute)

// Verify connection
if err := db.PingContext(ctx); err != nil {
    return nil, fmt.Errorf("pinging database: %w", err)
}
```

### Query Patterns

```go
// Single row
func (s *Store) GetByID(ctx context.Context, id string) (*User, error) {
    var u User
    err := s.db.QueryRowContext(ctx,
        "SELECT id, name, email, created_at FROM users WHERE id = $1", id,
    ).Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, ErrNotFound
        }
        return nil, fmt.Errorf("querying user: %w", err)
    }
    return &u, nil
}

// Multiple rows
func (s *Store) List(ctx context.Context, limit, offset int) ([]User, error) {
    rows, err := s.db.QueryContext(ctx,
        "SELECT id, name, email FROM users ORDER BY name LIMIT $1 OFFSET $2",
        limit, offset,
    )
    if err != nil {
        return nil, fmt.Errorf("querying users: %w", err)
    }
    defer rows.Close()

    var users []User
    for rows.Next() {
        var u User
        if err := rows.Scan(&u.ID, &u.Name, &u.Email); err != nil {
            return nil, fmt.Errorf("scanning user: %w", err)
        }
        users = append(users, u)
    }
    return users, rows.Err()
}

// Exec (INSERT, UPDATE, DELETE)
func (s *Store) Create(ctx context.Context, u *User) error {
    err := s.db.QueryRowContext(ctx,
        "INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id, created_at",
        u.Name, u.Email,
    ).Scan(&u.ID, &u.CreatedAt)
    if err != nil {
        // Check for unique constraint violation
        var pgErr *pgconn.PgError
        if errors.As(err, &pgErr) && pgErr.Code == "23505" {
            return ErrConflict
        }
        return fmt.Errorf("inserting user: %w", err)
    }
    return nil
}

// Transaction
func (s *Store) Transfer(ctx context.Context, fromID, toID string, amount int64) error {
    tx, err := s.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback() // No-op if committed

    // Debit
    result, err := tx.ExecContext(ctx,
        "UPDATE accounts SET balance = balance - $1 WHERE id = $2 AND balance >= $1",
        amount, fromID,
    )
    if err != nil {
        return err
    }
    rows, _ := result.RowsAffected()
    if rows == 0 {
        return ErrInsufficientFunds
    }

    // Credit
    _, err = tx.ExecContext(ctx,
        "UPDATE accounts SET balance = balance + $1 WHERE id = $2",
        amount, toID,
    )
    if err != nil {
        return err
    }

    return tx.Commit()
}

// Nullable columns
var name sql.NullString
var age sql.NullInt64
row.Scan(&name, &age)
if name.Valid {
    user.Name = name.String
}
```

## os and os/exec — System Interaction

```go
// Environment variables
val := os.Getenv("KEY")
val, ok := os.LookupEnv("KEY") // Distinguish empty from missing

// File operations
data, err := os.ReadFile("config.json")
err = os.WriteFile("output.txt", data, 0o644)

// Create temp file
f, err := os.CreateTemp("", "prefix-*.json")
defer os.Remove(f.Name())
defer f.Close()

// Command execution with context
func runCommand(ctx context.Context, name string, args ...string) (string, error) {
    cmd := exec.CommandContext(ctx, name, args...)
    cmd.Env = append(os.Environ(), "CUSTOM_VAR=value")

    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr

    if err := cmd.Run(); err != nil {
        return "", fmt.Errorf("command %s failed: %w\nstderr: %s", name, err, stderr.String())
    }
    return stdout.String(), nil
}
```

## testing — Test Framework Essentials

```go
// Subtests
func TestParse(t *testing.T) {
    t.Run("valid input", func(t *testing.T) {
        // ...
    })
    t.Run("invalid input", func(t *testing.T) {
        // ...
    })
}

// Parallel tests
func TestConcurrent(t *testing.T) {
    t.Parallel()
    // Test body
}

// Skip conditionally
func TestRequiresDocker(t *testing.T) {
    if os.Getenv("DOCKER_HOST") == "" {
        t.Skip("requires Docker")
    }
}

// Cleanup
func TestWithDB(t *testing.T) {
    db := createTestDB(t)
    t.Cleanup(func() { db.Close() })
}

// Temp directory (auto-cleaned)
dir := t.TempDir()

// Helper functions
func assertEqual[T comparable](t testing.TB, got, want T) {
    t.Helper() // Reports caller's line, not this function's line
    if got != want {
        t.Errorf("got %v, want %v", got, want)
    }
}
```

## embed — Embedding Files

```go
import "embed"

//go:embed templates/*.html
var templateFS embed.FS

//go:embed static/*
var staticFS embed.FS

//go:embed version.txt
var version string

//go:embed schema.sql
var schema []byte

// Serve embedded static files
mux.Handle("GET /static/", http.StripPrefix("/static/",
    http.FileServerFS(staticFS)))

// Parse embedded templates
tmpl := template.Must(template.ParseFS(templateFS, "templates/*.html"))
```

## crypto — Cryptographic Operations

```go
import (
    "crypto/rand"
    "crypto/sha256"
    "crypto/subtle"
    "encoding/hex"
    "golang.org/x/crypto/bcrypt"
)

// Generate random bytes
func generateToken(n int) (string, error) {
    b := make([]byte, n)
    if _, err := rand.Read(b); err != nil {
        return "", err
    }
    return hex.EncodeToString(b), nil
}

// Hash password with bcrypt
func hashPassword(password string) (string, error) {
    hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    return string(hash), err
}

func checkPassword(hash, password string) bool {
    return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// SHA-256 hash
func sha256Hash(data []byte) string {
    h := sha256.Sum256(data)
    return hex.EncodeToString(h[:])
}

// Constant-time comparison (prevents timing attacks)
func secureCompare(a, b string) bool {
    return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
```
