# Idiomatic Go Patterns

Practical patterns for writing clean, maintainable, production-quality Go code.

## Error Handling Patterns

### Sentinel Errors

Define package-level error values for expected conditions callers need to check:

```go
package auth

import "errors"

var (
    ErrInvalidToken  = errors.New("auth: invalid token")
    ErrExpiredToken  = errors.New("auth: token expired")
    ErrInvalidClaims = errors.New("auth: invalid claims")
)

// Callers check with errors.Is:
if errors.Is(err, auth.ErrExpiredToken) {
    // refresh the token
}
```

### Error Types with Context

When callers need to extract information from errors:

```go
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation: %s: %s", e.Field, e.Message)
}

type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
    var b strings.Builder
    for i, ve := range e {
        if i > 0 {
            b.WriteString("; ")
        }
        b.WriteString(ve.Error())
    }
    return b.String()
}

// Usage in service:
func (s *Service) CreateUser(ctx context.Context, req CreateUserRequest) (*User, error) {
    var errs ValidationErrors
    if req.Name == "" {
        errs = append(errs, ValidationError{Field: "name", Message: "is required"})
    }
    if !isValidEmail(req.Email) {
        errs = append(errs, ValidationError{Field: "email", Message: "is invalid"})
    }
    if len(errs) > 0 {
        return nil, errs
    }
    // ...
}

// Callers extract details with errors.As:
var valErrs ValidationErrors
if errors.As(err, &valErrs) {
    for _, ve := range valErrs {
        fmt.Printf("field %s: %s\n", ve.Field, ve.Message)
    }
}
```

### Error Wrapping with fmt.Errorf

```go
func (s *Store) GetUser(ctx context.Context, id string) (*User, error) {
    row := s.db.QueryRowContext(ctx, "SELECT id, name, email FROM users WHERE id = $1", id)
    var u User
    if err := row.Scan(&u.ID, &u.Name, &u.Email); err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, fmt.Errorf("get user %s: %w", id, ErrNotFound)
        }
        return nil, fmt.Errorf("get user %s: %w", id, err)
    }
    return &u, nil
}
```

### errors.Join for Multiple Errors (Go 1.20+)

```go
func (c *Config) Validate() error {
    var errs []error
    if c.Addr == "" {
        errs = append(errs, errors.New("addr is required"))
    }
    if c.Port < 1 || c.Port > 65535 {
        errs = append(errs, fmt.Errorf("port %d is out of range", c.Port))
    }
    if c.Timeout <= 0 {
        errs = append(errs, errors.New("timeout must be positive"))
    }
    return errors.Join(errs...)
}
```

### Defer Cleanup with Named Return

```go
func (s *Store) WithTransaction(ctx context.Context, fn func(tx *sql.Tx) error) (err error) {
    tx, err := s.db.BeginTx(ctx, nil)
    if err != nil {
        return fmt.Errorf("begin transaction: %w", err)
    }
    defer func() {
        if err != nil {
            if rbErr := tx.Rollback(); rbErr != nil {
                err = fmt.Errorf("%w (rollback failed: %v)", err, rbErr)
            }
            return
        }
        err = tx.Commit()
    }()

    return fn(tx)
}
```

## Interface Patterns

### Implicit Interface Satisfaction

Go interfaces are satisfied implicitly. Define interfaces at the point of consumption:

```go
// In the consumer package:
package notification

// Sender is defined where it's used, not where it's implemented.
type Sender interface {
    Send(ctx context.Context, to, subject, body string) error
}

type Service struct {
    sender Sender
}

// In the implementation package:
package email

// Client satisfies notification.Sender without importing it.
type Client struct {
    smtpHost string
    from     string
}

func (c *Client) Send(ctx context.Context, to, subject, body string) error {
    // SMTP implementation
    return nil
}
```

### Interface Compliance Check

Compile-time verification that a type implements an interface:

```go
// Compile-time check
var _ io.ReadWriteCloser = (*MyConn)(nil)
var _ http.Handler = (*APIHandler)(nil)
var _ fmt.Stringer = (*Status)(nil)
```

### Embedding Interfaces for Extension

```go
type ReadStore interface {
    Get(ctx context.Context, key string) ([]byte, error)
    List(ctx context.Context, prefix string) ([][]byte, error)
}

type WriteStore interface {
    Set(ctx context.Context, key string, value []byte) error
    Delete(ctx context.Context, key string) error
}

type Store interface {
    ReadStore
    WriteStore
}

// Most consumers only need ReadStore. Only write paths need the full Store.
```

### Interface Guards Against Nil

```go
type Logger interface {
    Info(msg string, args ...any)
    Error(msg string, args ...any)
}

// NopLogger satisfies Logger but does nothing — useful for tests or optional logging.
type NopLogger struct{}

func (NopLogger) Info(msg string, args ...any)  {}
func (NopLogger) Error(msg string, args ...any) {}

// Usage with nil safety:
func NewService(repo Repository, logger Logger) *Service {
    if logger == nil {
        logger = NopLogger{}
    }
    return &Service{repo: repo, logger: logger}
}
```

## Generics Patterns (Go 1.18+)

### Type Constraints

```go
// Built-in constraints
type Number interface {
    ~int | ~int8 | ~int16 | ~int32 | ~int64 |
    ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
    ~float32 | ~float64
}

func Sum[T Number](values []T) T {
    var total T
    for _, v := range values {
        total += v
    }
    return total
}

// Using cmp.Ordered for comparable types
func Max[T cmp.Ordered](a, b T) T {
    if a > b {
        return a
    }
    return b
}

func Min[T cmp.Ordered](values ...T) T {
    m := values[0]
    for _, v := range values[1:] {
        if v < m {
            m = v
        }
    }
    return m
}
```

### Generic Data Structures

```go
// Stack
type Stack[T any] struct {
    items []T
}

func (s *Stack[T]) Push(item T) {
    s.items = append(s.items, item)
}

func (s *Stack[T]) Pop() (T, bool) {
    if len(s.items) == 0 {
        var zero T
        return zero, false
    }
    item := s.items[len(s.items)-1]
    s.items = s.items[:len(s.items)-1]
    return item, true
}

func (s *Stack[T]) Peek() (T, bool) {
    if len(s.items) == 0 {
        var zero T
        return zero, false
    }
    return s.items[len(s.items)-1], true
}

func (s *Stack[T]) Len() int { return len(s.items) }

// Set
type Set[T comparable] struct {
    m map[T]struct{}
}

func NewSet[T comparable](items ...T) *Set[T] {
    s := &Set[T]{m: make(map[T]struct{}, len(items))}
    for _, item := range items {
        s.m[item] = struct{}{}
    }
    return s
}

func (s *Set[T]) Add(item T)           { s.m[item] = struct{}{} }
func (s *Set[T]) Remove(item T)        { delete(s.m, item) }
func (s *Set[T]) Contains(item T) bool { _, ok := s.m[item]; return ok }
func (s *Set[T]) Len() int             { return len(s.m) }

func (s *Set[T]) Union(other *Set[T]) *Set[T] {
    result := NewSet[T]()
    for k := range s.m {
        result.Add(k)
    }
    for k := range other.m {
        result.Add(k)
    }
    return result
}

func (s *Set[T]) Intersection(other *Set[T]) *Set[T] {
    result := NewSet[T]()
    smaller, larger := s, other
    if smaller.Len() > larger.Len() {
        smaller, larger = larger, smaller
    }
    for k := range smaller.m {
        if larger.Contains(k) {
            result.Add(k)
        }
    }
    return result
}
```

### Generic Functional Helpers

```go
func Map[T, U any](items []T, fn func(T) U) []U {
    result := make([]U, len(items))
    for i, item := range items {
        result[i] = fn(item)
    }
    return result
}

func Filter[T any](items []T, pred func(T) bool) []T {
    result := make([]T, 0, len(items)/2)
    for _, item := range items {
        if pred(item) {
            result = append(result, item)
        }
    }
    return result
}

func Reduce[T, U any](items []T, initial U, fn func(U, T) U) U {
    acc := initial
    for _, item := range items {
        acc = fn(acc, item)
    }
    return acc
}

func GroupBy[T any, K comparable](items []T, key func(T) K) map[K][]T {
    groups := make(map[K][]T)
    for _, item := range items {
        k := key(item)
        groups[k] = append(groups[k], item)
    }
    return groups
}

func Chunk[T any](items []T, size int) [][]T {
    var chunks [][]T
    for size < len(items) {
        items, chunks = items[size:], append(chunks, items[:size])
    }
    return append(chunks, items)
}

// Usage:
names := Map(users, func(u User) string { return u.Name })
adults := Filter(users, func(u User) bool { return u.Age >= 18 })
total := Reduce(orders, 0, func(acc int, o Order) int { return acc + o.Total })
byDept := GroupBy(employees, func(e Employee) string { return e.Department })
```

## Builder Pattern

```go
type QueryBuilder struct {
    table      string
    conditions []string
    args       []any
    orderBy    string
    limit      int
    offset     int
}

func Select(table string) *QueryBuilder {
    return &QueryBuilder{table: table}
}

func (q *QueryBuilder) Where(condition string, args ...any) *QueryBuilder {
    q.conditions = append(q.conditions, condition)
    q.args = append(q.args, args...)
    return q
}

func (q *QueryBuilder) OrderBy(field string) *QueryBuilder {
    q.orderBy = field
    return q
}

func (q *QueryBuilder) Limit(n int) *QueryBuilder {
    q.limit = n
    return q
}

func (q *QueryBuilder) Offset(n int) *QueryBuilder {
    q.offset = n
    return q
}

func (q *QueryBuilder) Build() (string, []any) {
    var b strings.Builder
    b.WriteString("SELECT * FROM ")
    b.WriteString(q.table)

    if len(q.conditions) > 0 {
        b.WriteString(" WHERE ")
        b.WriteString(strings.Join(q.conditions, " AND "))
    }
    if q.orderBy != "" {
        b.WriteString(" ORDER BY ")
        b.WriteString(q.orderBy)
    }
    if q.limit > 0 {
        fmt.Fprintf(&b, " LIMIT %d", q.limit)
    }
    if q.offset > 0 {
        fmt.Fprintf(&b, " OFFSET %d", q.offset)
    }
    return b.String(), q.args
}

// Usage:
query, args := Select("users").
    Where("status = $1", "active").
    Where("created_at > $2", cutoffDate).
    OrderBy("name ASC").
    Limit(20).
    Offset(40).
    Build()
```

## Iterator Patterns (Go 1.22+)

### Range Over Function

```go
// Iter over map entries sorted by key
func SortedKeys[K cmp.Ordered, V any](m map[K]V) iter.Seq2[K, V] {
    return func(yield func(K, V) bool) {
        keys := make([]K, 0, len(m))
        for k := range m {
            keys = append(keys, k)
        }
        slices.Sort(keys)
        for _, k := range keys {
            if !yield(k, m[k]) {
                return
            }
        }
    }
}

// Usage:
for key, value := range SortedKeys(myMap) {
    fmt.Printf("%s: %v\n", key, value)
}

// Iter over lines in a file
func Lines(r io.Reader) iter.Seq2[string, error] {
    return func(yield func(string, error) bool) {
        scanner := bufio.NewScanner(r)
        for scanner.Scan() {
            if !yield(scanner.Text(), nil) {
                return
            }
        }
        if err := scanner.Err(); err != nil {
            yield("", err)
        }
    }
}

// Iter with filtering
func Where[T any](seq iter.Seq[T], pred func(T) bool) iter.Seq[T] {
    return func(yield func(T) bool) {
        for v := range seq {
            if pred(v) {
                if !yield(v) {
                    return
                }
            }
        }
    }
}

// Iter with transformation
func Transform[T, U any](seq iter.Seq[T], fn func(T) U) iter.Seq[U] {
    return func(yield func(U) bool) {
        for v := range seq {
            if !yield(fn(v)) {
                return
            }
        }
    }
}

// Take first N items
func Take[T any](seq iter.Seq[T], n int) iter.Seq[T] {
    return func(yield func(T) bool) {
        i := 0
        for v := range seq {
            if i >= n {
                return
            }
            if !yield(v) {
                return
            }
            i++
        }
    }
}
```

## Enum Pattern with iota

```go
type Status int

const (
    StatusPending   Status = iota + 1 // Start at 1 to distinguish from zero value
    StatusActive
    StatusSuspended
    StatusClosed
)

var statusNames = map[Status]string{
    StatusPending:   "pending",
    StatusActive:    "active",
    StatusSuspended: "suspended",
    StatusClosed:    "closed",
}

var statusValues = map[string]Status{
    "pending":   StatusPending,
    "active":    StatusActive,
    "suspended": StatusSuspended,
    "closed":    StatusClosed,
}

func (s Status) String() string {
    if name, ok := statusNames[s]; ok {
        return name
    }
    return fmt.Sprintf("Status(%d)", s)
}

func ParseStatus(s string) (Status, error) {
    if v, ok := statusValues[strings.ToLower(s)]; ok {
        return v, nil
    }
    return 0, fmt.Errorf("unknown status: %q", s)
}

func (s Status) MarshalJSON() ([]byte, error) {
    return json.Marshal(s.String())
}

func (s *Status) UnmarshalJSON(data []byte) error {
    var str string
    if err := json.Unmarshal(data, &str); err != nil {
        return err
    }
    parsed, err := ParseStatus(str)
    if err != nil {
        return err
    }
    *s = parsed
    return nil
}

// For database scanning:
func (s *Status) Scan(src any) error {
    switch v := src.(type) {
    case string:
        parsed, err := ParseStatus(v)
        if err != nil {
            return err
        }
        *s = parsed
    case int64:
        *s = Status(v)
    default:
        return fmt.Errorf("cannot scan %T into Status", src)
    }
    return nil
}

func (s Status) Value() (driver.Value, error) {
    return s.String(), nil
}
```

## Structured Logging with slog

```go
// Setup
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level:     slog.LevelInfo,
    AddSource: true,
}))

// Contextual logging — add fields for the request lifecycle
func requestLogger(logger *slog.Logger, r *http.Request) *slog.Logger {
    return logger.With(
        "request_id", r.Header.Get("X-Request-ID"),
        "method", r.Method,
        "path", r.URL.Path,
        "remote_addr", r.RemoteAddr,
    )
}

// Usage in handlers:
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
    log := requestLogger(h.logger, r)

    id := r.PathValue("id")
    log.Info("fetching user", "user_id", id)

    user, err := h.service.GetUser(r.Context(), id)
    if err != nil {
        log.Error("failed to get user", "user_id", id, "err", err)
        writeError(w, err)
        return
    }

    log.Info("user fetched", "user_id", id, "user_name", user.Name)
    writeJSON(w, http.StatusOK, user)
}

// Typed attributes for performance-critical logging
slog.Info("processed request",
    slog.String("method", "GET"),
    slog.Int("status", 200),
    slog.Duration("latency", elapsed),
    slog.Group("user",
        slog.String("id", userID),
        slog.String("role", "admin"),
    ),
)
```

## HTTP Patterns

### JSON Response Helpers

```go
func writeJSON(w http.ResponseWriter, status int, data any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    if err := json.NewEncoder(w).Encode(data); err != nil {
        slog.Error("failed to encode response", "err", err)
    }
}

func writeError(w http.ResponseWriter, err error) {
    status := apperr.HTTPStatus(err)
    writeJSON(w, status, map[string]string{
        "error": err.Error(),
    })
}

func readJSON[T any](r *http.Request, maxBytes int64) (T, error) {
    var v T
    r.Body = http.MaxBytesReader(nil, r.Body, maxBytes)
    dec := json.NewDecoder(r.Body)
    dec.DisallowUnknownFields()
    if err := dec.Decode(&v); err != nil {
        return v, fmt.Errorf("decoding request body: %w", err)
    }
    return v, nil
}
```

### Pagination Pattern

```go
type Page[T any] struct {
    Items      []T    `json:"items"`
    NextCursor string `json:"next_cursor,omitempty"`
    HasMore    bool   `json:"has_more"`
}

type PageRequest struct {
    Cursor string
    Limit  int
}

func ParsePageRequest(r *http.Request) PageRequest {
    cursor := r.URL.Query().Get("cursor")
    limit := 20
    if l, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && l > 0 && l <= 100 {
        limit = l
    }
    return PageRequest{Cursor: cursor, Limit: limit}
}

func (s *Store) ListUsers(ctx context.Context, req PageRequest) (*Page[User], error) {
    // Fetch one extra to determine if there are more
    query := "SELECT id, name FROM users WHERE id > $1 ORDER BY id LIMIT $2"
    rows, err := s.db.QueryContext(ctx, query, req.Cursor, req.Limit+1)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var users []User
    for rows.Next() {
        var u User
        if err := rows.Scan(&u.ID, &u.Name); err != nil {
            return nil, err
        }
        users = append(users, u)
    }

    page := &Page[User]{Items: users}
    if len(users) > req.Limit {
        page.Items = users[:req.Limit]
        page.HasMore = true
        page.NextCursor = users[req.Limit-1].ID
    }
    return page, rows.Err()
}
```

## Resource Management

### Closer Pattern

```go
type App struct {
    db     *sql.DB
    cache  *redis.Client
    server *http.Server
}

func (a *App) Close() error {
    var errs []error
    if a.server != nil {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        if err := a.server.Shutdown(ctx); err != nil {
            errs = append(errs, fmt.Errorf("server: %w", err))
        }
    }
    if a.cache != nil {
        if err := a.cache.Close(); err != nil {
            errs = append(errs, fmt.Errorf("cache: %w", err))
        }
    }
    if a.db != nil {
        if err := a.db.Close(); err != nil {
            errs = append(errs, fmt.Errorf("database: %w", err))
        }
    }
    return errors.Join(errs...)
}
```

### Retry Pattern

```go
type RetryConfig struct {
    MaxAttempts int
    BaseDelay   time.Duration
    MaxDelay    time.Duration
    Retryable   func(error) bool
}

func Retry(ctx context.Context, cfg RetryConfig, fn func(ctx context.Context) error) error {
    var lastErr error
    for attempt := range cfg.MaxAttempts {
        lastErr = fn(ctx)
        if lastErr == nil {
            return nil
        }
        if cfg.Retryable != nil && !cfg.Retryable(lastErr) {
            return lastErr
        }
        if attempt == cfg.MaxAttempts-1 {
            break
        }

        delay := cfg.BaseDelay * time.Duration(1<<attempt)
        if delay > cfg.MaxDelay {
            delay = cfg.MaxDelay
        }
        // Add jitter
        jitter := time.Duration(rand.Int64N(int64(delay) / 2))
        delay = delay/2 + jitter

        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-time.After(delay):
        }
    }
    return fmt.Errorf("after %d attempts: %w", cfg.MaxAttempts, lastErr)
}

// Usage:
err := Retry(ctx, RetryConfig{
    MaxAttempts: 3,
    BaseDelay:   100 * time.Millisecond,
    MaxDelay:    5 * time.Second,
    Retryable: func(err error) bool {
        return !errors.Is(err, ErrNotFound)
    },
}, func(ctx context.Context) error {
    return client.Call(ctx, request)
})
```

## Go Proverbs Applied

1. **Don't panic** — Reserve `panic` for truly unrecoverable programmer errors, never for runtime conditions
2. **Make the zero value useful** — `sync.Mutex{}`, `bytes.Buffer{}` all work without initialization
3. **A little copying is better than a little dependency** — Don't import a package for one function
4. **Clear is better than clever** — Write obvious code, not clever code
5. **Reflection is never clear** — Avoid `reflect` package in application code
6. **Errors are values** — Use them as control flow, not as exceptions
7. **Don't just check errors, handle them gracefully** — Add context, decide what to do
8. **Design the architecture, name the components, document the details** — Package names are part of the API
