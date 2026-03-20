---
name: go-web-apis
description: >
  Production Go web API patterns — HTTP handlers, middleware, routing,
  database access, validation, authentication, and structured error handling
  with standard library and popular frameworks.
  Triggers: "go api", "go http", "go web server", "go router", "go middleware",
  "gin", "chi", "go rest api", "go handler".
  NOT for: Go CLI tools or system programming (use go-concurrency).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Go Web API Patterns

## Project Structure

```
myapi/
  cmd/
    server/
      main.go           # Entry point
  internal/
    config/
      config.go         # Configuration
    handler/
      user.go           # HTTP handlers
      middleware.go      # Middleware
    model/
      user.go           # Domain models
    repository/
      user.go           # Database access
    service/
      user.go           # Business logic
  pkg/
    response/
      response.go       # Shared response helpers
    validator/
      validator.go      # Input validation
  migrations/
  go.mod
  go.sum
```

## HTTP Server Setup

```go
// cmd/server/main.go
package main

import (
    "context"
    "log/slog"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "myapi/internal/config"
    "myapi/internal/handler"
    "myapi/internal/repository"
    "myapi/internal/service"
)

func main() {
    cfg := config.Load()

    // Structured logging
    logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level: cfg.LogLevel,
    }))
    slog.SetDefault(logger)

    // Database
    db, err := repository.NewPostgresDB(cfg.DatabaseURL)
    if err != nil {
        slog.Error("failed to connect to database", "error", err)
        os.Exit(1)
    }
    defer db.Close()

    // Wire dependencies
    userRepo := repository.NewUserRepository(db)
    userSvc := service.NewUserService(userRepo)
    userHandler := handler.NewUserHandler(userSvc)

    // Router
    mux := http.NewServeMux()
    mux.HandleFunc("GET /health", handler.HealthCheck)
    mux.HandleFunc("GET /api/users", userHandler.List)
    mux.HandleFunc("GET /api/users/{id}", userHandler.Get)
    mux.HandleFunc("POST /api/users", userHandler.Create)
    mux.HandleFunc("PUT /api/users/{id}", userHandler.Update)
    mux.HandleFunc("DELETE /api/users/{id}", userHandler.Delete)

    // Middleware stack
    wrapped := handler.Chain(mux,
        handler.Recovery,
        handler.RequestID,
        handler.Logger,
        handler.CORS(cfg.AllowedOrigins),
        handler.RateLimit(100, time.Minute),
    )

    // Server with timeouts
    srv := &http.Server{
        Addr:         ":" + cfg.Port,
        Handler:      wrapped,
        ReadTimeout:  5 * time.Second,
        WriteTimeout: 10 * time.Second,
        IdleTimeout:  120 * time.Second,
    }

    // Graceful shutdown
    go func() {
        slog.Info("server starting", "port", cfg.Port)
        if err := srv.ListenAndServe(); err != http.ErrServerClosed {
            slog.Error("server error", "error", err)
            os.Exit(1)
        }
    }()

    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    slog.Info("shutting down server")
    if err := srv.Shutdown(ctx); err != nil {
        slog.Error("forced shutdown", "error", err)
    }
}
```

## Configuration

```go
// internal/config/config.go
package config

import (
    "log/slog"
    "os"
    "strconv"
    "time"
)

type Config struct {
    Port           string
    DatabaseURL    string
    JWTSecret      string
    JWTExpiry      time.Duration
    AllowedOrigins []string
    LogLevel       slog.Level
}

func Load() *Config {
    return &Config{
        Port:           getEnv("PORT", "8080"),
        DatabaseURL:    requireEnv("DATABASE_URL"),
        JWTSecret:      requireEnv("JWT_SECRET"),
        JWTExpiry:      getDuration("JWT_EXPIRY", 24*time.Hour),
        AllowedOrigins: getEnvSlice("ALLOWED_ORIGINS", []string{"http://localhost:3000"}),
        LogLevel:       getLogLevel("LOG_LEVEL", slog.LevelInfo),
    }
}

func getEnv(key, fallback string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return fallback
}

func requireEnv(key string) string {
    v := os.Getenv(key)
    if v == "" {
        slog.Error("required env var missing", "key", key)
        os.Exit(1)
    }
    return v
}
```

## Handlers

```go
// internal/handler/user.go
package handler

import (
    "encoding/json"
    "errors"
    "net/http"
    "strconv"

    "myapi/internal/model"
    "myapi/internal/service"
    "myapi/pkg/response"
)

type UserHandler struct {
    svc *service.UserService
}

func NewUserHandler(svc *service.UserService) *UserHandler {
    return &UserHandler{svc: svc}
}

func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
    page, _ := strconv.Atoi(r.URL.Query().Get("page"))
    if page < 1 {
        page = 1
    }
    limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
    if limit < 1 || limit > 100 {
        limit = 20
    }

    users, total, err := h.svc.List(r.Context(), page, limit)
    if err != nil {
        response.Error(w, http.StatusInternalServerError, "failed to list users")
        return
    }

    response.Paginated(w, users, total, page, limit)
}

func (h *UserHandler) Get(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")

    user, err := h.svc.GetByID(r.Context(), id)
    if err != nil {
        if errors.Is(err, service.ErrNotFound) {
            response.Error(w, http.StatusNotFound, "user not found")
            return
        }
        response.Error(w, http.StatusInternalServerError, "failed to get user")
        return
    }

    response.JSON(w, http.StatusOK, user)
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
    var input model.CreateUserInput
    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        response.Error(w, http.StatusBadRequest, "invalid request body")
        return
    }

    if errs := input.Validate(); len(errs) > 0 {
        response.ValidationError(w, errs)
        return
    }

    user, err := h.svc.Create(r.Context(), &input)
    if err != nil {
        if errors.Is(err, service.ErrDuplicate) {
            response.Error(w, http.StatusConflict, "email already registered")
            return
        }
        response.Error(w, http.StatusInternalServerError, "failed to create user")
        return
    }

    response.JSON(w, http.StatusCreated, user)
}
```

## Middleware

```go
// internal/handler/middleware.go
package handler

import (
    "context"
    "log/slog"
    "net/http"
    "runtime/debug"
    "strings"
    "sync"
    "time"

    "github.com/google/uuid"
)

type contextKey string

const RequestIDKey contextKey = "requestID"

// Chain applies middleware in order (last wraps first)
func Chain(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
    for i := len(middlewares) - 1; i >= 0; i-- {
        h = middlewares[i](h)
    }
    return h
}

// Recovery catches panics and returns 500
func Recovery(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                slog.Error("panic recovered",
                    "error", err,
                    "stack", string(debug.Stack()),
                    "path", r.URL.Path,
                )
                http.Error(w, "internal server error", http.StatusInternalServerError)
            }
        }()
        next.ServeHTTP(w, r)
    })
}

// RequestID injects a unique ID into context and response header
func RequestID(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        id := r.Header.Get("X-Request-ID")
        if id == "" {
            id = uuid.NewString()
        }
        ctx := context.WithValue(r.Context(), RequestIDKey, id)
        w.Header().Set("X-Request-ID", id)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// Logger logs request method, path, status, and duration
func Logger(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        wrapped := &statusWriter{ResponseWriter: w, status: http.StatusOK}

        next.ServeHTTP(wrapped, r)

        slog.Info("request",
            "method", r.Method,
            "path", r.URL.Path,
            "status", wrapped.status,
            "duration_ms", time.Since(start).Milliseconds(),
            "request_id", r.Context().Value(RequestIDKey),
        )
    })
}

type statusWriter struct {
    http.ResponseWriter
    status int
}

func (w *statusWriter) WriteHeader(code int) {
    w.status = code
    w.ResponseWriter.WriteHeader(code)
}

// CORS middleware
func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
    origins := make(map[string]bool)
    for _, o := range allowedOrigins {
        origins[o] = true
    }

    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            origin := r.Header.Get("Origin")
            if origins[origin] || origins["*"] {
                w.Header().Set("Access-Control-Allow-Origin", origin)
                w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
                w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
                w.Header().Set("Access-Control-Max-Age", "86400")
            }

            if r.Method == http.MethodOptions {
                w.WriteHeader(http.StatusNoContent)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}

// RateLimit — token bucket per IP
func RateLimit(maxRequests int, window time.Duration) func(http.Handler) http.Handler {
    type client struct {
        count    int
        lastSeen time.Time
    }
    var (
        mu      sync.Mutex
        clients = make(map[string]*client)
    )

    // Cleanup goroutine
    go func() {
        for range time.Tick(window) {
            mu.Lock()
            for ip, c := range clients {
                if time.Since(c.lastSeen) > window {
                    delete(clients, ip)
                }
            }
            mu.Unlock()
        }
    }()

    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            ip := strings.Split(r.RemoteAddr, ":")[0]

            mu.Lock()
            c, exists := clients[ip]
            if !exists {
                c = &client{}
                clients[ip] = c
            }
            if time.Since(c.lastSeen) > window {
                c.count = 0
            }
            c.count++
            c.lastSeen = time.Now()
            count := c.count
            mu.Unlock()

            if count > maxRequests {
                http.Error(w, "too many requests", http.StatusTooManyRequests)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}
```

## Response Helpers

```go
// pkg/response/response.go
package response

import (
    "encoding/json"
    "net/http"
)

type ErrorResponse struct {
    Error   string            `json:"error"`
    Details map[string]string `json:"details,omitempty"`
}

type PaginatedResponse struct {
    Data  any `json:"data"`
    Total int `json:"total"`
    Page  int `json:"page"`
    Pages int `json:"pages"`
}

func JSON(w http.ResponseWriter, status int, data any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(data)
}

func Error(w http.ResponseWriter, status int, message string) {
    JSON(w, status, ErrorResponse{Error: message})
}

func ValidationError(w http.ResponseWriter, errors map[string]string) {
    JSON(w, http.StatusUnprocessableEntity, ErrorResponse{
        Error:   "validation failed",
        Details: errors,
    })
}

func Paginated(w http.ResponseWriter, data any, total, page, limit int) {
    pages := (total + limit - 1) / limit
    JSON(w, http.StatusOK, PaginatedResponse{
        Data:  data,
        Total: total,
        Page:  page,
        Pages: pages,
    })
}
```

## Database Access

```go
// internal/repository/user.go
package repository

import (
    "context"
    "database/sql"
    "errors"
    "fmt"

    "myapi/internal/model"
)

type UserRepository struct {
    db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
    return &UserRepository{db: db}
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
    var u model.User
    err := r.db.QueryRowContext(ctx,
        `SELECT id, name, email, created_at, updated_at FROM users WHERE id = $1`,
        id,
    ).Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt, &u.UpdatedAt)

    if errors.Is(err, sql.ErrNoRows) {
        return nil, fmt.Errorf("user %s: %w", id, ErrNotFound)
    }
    if err != nil {
        return nil, fmt.Errorf("query user: %w", err)
    }
    return &u, nil
}

func (r *UserRepository) List(ctx context.Context, offset, limit int) ([]model.User, int, error) {
    // Count
    var total int
    err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&total)
    if err != nil {
        return nil, 0, fmt.Errorf("count users: %w", err)
    }

    // Fetch page
    rows, err := r.db.QueryContext(ctx,
        `SELECT id, name, email, created_at, updated_at
         FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
        limit, offset,
    )
    if err != nil {
        return nil, 0, fmt.Errorf("list users: %w", err)
    }
    defer rows.Close()

    var users []model.User
    for rows.Next() {
        var u model.User
        if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt, &u.UpdatedAt); err != nil {
            return nil, 0, fmt.Errorf("scan user: %w", err)
        }
        users = append(users, u)
    }
    return users, total, rows.Err()
}

func (r *UserRepository) Create(ctx context.Context, u *model.User) error {
    _, err := r.db.ExecContext(ctx,
        `INSERT INTO users (id, name, email, password_hash, created_at, updated_at)
         VALUES ($1, $2, $3, $4, NOW(), NOW())`,
        u.ID, u.Name, u.Email, u.PasswordHash,
    )
    return err
}

// Transaction example
func (r *UserRepository) TransferCredits(ctx context.Context, fromID, toID string, amount int) error {
    tx, err := r.db.BeginTx(ctx, nil)
    if err != nil {
        return fmt.Errorf("begin tx: %w", err)
    }
    defer tx.Rollback() // No-op if committed

    var balance int
    err = tx.QueryRowContext(ctx,
        `SELECT credits FROM users WHERE id = $1 FOR UPDATE`, fromID,
    ).Scan(&balance)
    if err != nil {
        return fmt.Errorf("check balance: %w", err)
    }
    if balance < amount {
        return fmt.Errorf("insufficient credits: have %d, need %d", balance, amount)
    }

    if _, err := tx.ExecContext(ctx,
        `UPDATE users SET credits = credits - $1 WHERE id = $2`, amount, fromID,
    ); err != nil {
        return fmt.Errorf("debit: %w", err)
    }

    if _, err := tx.ExecContext(ctx,
        `UPDATE users SET credits = credits + $1 WHERE id = $2`, amount, toID,
    ); err != nil {
        return fmt.Errorf("credit: %w", err)
    }

    return tx.Commit()
}
```

## Input Validation

```go
// internal/model/user.go
package model

import (
    "net/mail"
    "time"
    "unicode/utf8"
)

type User struct {
    ID           string    `json:"id"`
    Name         string    `json:"name"`
    Email        string    `json:"email"`
    PasswordHash string    `json:"-"` // Never serialize
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
}

type CreateUserInput struct {
    Name     string `json:"name"`
    Email    string `json:"email"`
    Password string `json:"password"`
}

func (i *CreateUserInput) Validate() map[string]string {
    errs := make(map[string]string)

    if utf8.RuneCountInString(i.Name) < 2 {
        errs["name"] = "must be at least 2 characters"
    }
    if utf8.RuneCountInString(i.Name) > 100 {
        errs["name"] = "must be at most 100 characters"
    }
    if _, err := mail.ParseAddress(i.Email); err != nil {
        errs["email"] = "invalid email address"
    }
    if len(i.Password) < 8 {
        errs["password"] = "must be at least 8 characters"
    }

    if len(errs) == 0 {
        return nil
    }
    return errs
}
```

## JWT Authentication

```go
// internal/handler/auth.go
package handler

import (
    "context"
    "net/http"
    "strings"
    "time"

    "github.com/golang-jwt/jwt/v5"
)

type Claims struct {
    UserID string `json:"user_id"`
    Role   string `json:"role"`
    jwt.RegisteredClaims
}

func GenerateToken(userID, role, secret string, expiry time.Duration) (string, error) {
    claims := Claims{
        UserID: userID,
        Role:   role,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
        },
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(secret))
}

// Auth middleware — extracts and validates JWT
func Auth(secret string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            header := r.Header.Get("Authorization")
            if !strings.HasPrefix(header, "Bearer ") {
                http.Error(w, "missing authorization", http.StatusUnauthorized)
                return
            }

            tokenStr := strings.TrimPrefix(header, "Bearer ")
            claims := &Claims{}

            token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
                return []byte(secret), nil
            })

            if err != nil || !token.Valid {
                http.Error(w, "invalid token", http.StatusUnauthorized)
                return
            }

            ctx := context.WithValue(r.Context(), "userID", claims.UserID)
            ctx = context.WithValue(ctx, "role", claims.Role)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

// RequireRole checks the user has the specified role
func RequireRole(role string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            userRole, _ := r.Context().Value("role").(string)
            if userRole != role {
                http.Error(w, "forbidden", http.StatusForbidden)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}
```

## Testing

```go
// internal/handler/user_test.go
package handler_test

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "myapi/internal/handler"
    "myapi/internal/service"
)

// Mock service for testing
type mockUserService struct {
    users []model.User
    err   error
}

func (m *mockUserService) List(ctx context.Context, page, limit int) ([]model.User, int, error) {
    if m.err != nil {
        return nil, 0, m.err
    }
    return m.users, len(m.users), nil
}

func TestUserHandler_List(t *testing.T) {
    svc := &mockUserService{
        users: []model.User{{ID: "1", Name: "Alice", Email: "alice@test.com"}},
    }
    h := handler.NewUserHandler(svc)

    req := httptest.NewRequest(http.MethodGet, "/api/users?page=1&limit=10", nil)
    rec := httptest.NewRecorder()

    h.List(rec, req)

    if rec.Code != http.StatusOK {
        t.Errorf("expected 200, got %d", rec.Code)
    }

    var resp map[string]any
    json.NewDecoder(rec.Body).Decode(&resp)
    data := resp["data"].([]any)
    if len(data) != 1 {
        t.Errorf("expected 1 user, got %d", len(data))
    }
}

func TestUserHandler_Create_Validation(t *testing.T) {
    h := handler.NewUserHandler(&mockUserService{})

    body := `{"name": "A", "email": "bad", "password": "short"}`
    req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBufferString(body))
    rec := httptest.NewRecorder()

    h.Create(rec, req)

    if rec.Code != http.StatusUnprocessableEntity {
        t.Errorf("expected 422, got %d", rec.Code)
    }
}

// Table-driven tests
func TestCreateUserInput_Validate(t *testing.T) {
    tests := []struct {
        name    string
        input   model.CreateUserInput
        wantErr bool
        field   string
    }{
        {"valid", model.CreateUserInput{Name: "Alice", Email: "a@b.com", Password: "12345678"}, false, ""},
        {"short name", model.CreateUserInput{Name: "A", Email: "a@b.com", Password: "12345678"}, true, "name"},
        {"bad email", model.CreateUserInput{Name: "Alice", Email: "bad", Password: "12345678"}, true, "email"},
        {"short password", model.CreateUserInput{Name: "Alice", Email: "a@b.com", Password: "123"}, true, "password"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            errs := tt.input.Validate()
            if tt.wantErr && errs == nil {
                t.Error("expected validation error")
            }
            if !tt.wantErr && errs != nil {
                t.Errorf("unexpected errors: %v", errs)
            }
            if tt.field != "" && errs[tt.field] == "" {
                t.Errorf("expected error on field %q", tt.field)
            }
        })
    }
}
```

## Gotchas

1. **Goroutine leaks in HTTP handlers** — If you spawn a goroutine in a handler but the request context gets canceled (client disconnects), the goroutine keeps running. Always pass `r.Context()` and select on `ctx.Done()` in long-running goroutines.

2. **`defer rows.Close()` after error check** — Always check the error from `db.QueryContext` BEFORE deferring `rows.Close()`. If `err != nil`, `rows` is nil and `rows.Close()` will panic.

3. **`json:"-"` vs omitting the tag** — `json:"-"` excludes a field from JSON serialization. No tag at all exports the field with its Go name. For passwords and secrets, always use `json:"-"`.

4. **`http.Error` doesn't return** — After calling `http.Error(w, ...)`, execution continues. You MUST `return` after it, otherwise the handler keeps running and may write additional responses, causing a superfluous WriteHeader warning.

5. **`context.Value` type assertions** — `r.Context().Value("key")` returns `any`. Always use a comma-ok assertion: `val, ok := ctx.Value(key).(string)`. Failing to check causes silent zero-value bugs.

6. **`sql.DB` is a pool, not a connection** — `sql.Open` returns a pool. Don't create one per request. Create ONE `sql.DB` at startup and share it. Set `SetMaxOpenConns`, `SetMaxIdleConns`, and `SetConnMaxLifetime` to control pool behavior.
