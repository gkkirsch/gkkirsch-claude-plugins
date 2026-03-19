---
name: go-project-setup
description: >
  Scaffold a new Go project with module initialization, directory structure, CI configuration,
  linting setup, Dockerfile, Makefile, and development tooling. Supports CLI tools, HTTP APIs,
  gRPC services, and library packages.
---

# Go Project Setup Skill

Scaffold a new Go project with proper structure, tooling, and CI configuration.

## Gather Requirements

Before scaffolding, determine:

1. **Project type**: CLI tool, HTTP API, gRPC service, library, or worker
2. **Module path**: e.g., `github.com/org/project`
3. **Go version**: Default to latest stable (1.22+)
4. **Database**: PostgreSQL (default), SQLite, or none
5. **Features needed**: Authentication, background jobs, caching, etc.

## Project Templates

### HTTP API Service

```bash
mkdir -p cmd/server internal/{auth,config,middleware,platform/{database,cache}} migrations api
```

**cmd/server/main.go**:
```go
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"MODULE_PATH/internal/config"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("loading config", "err", err)
		os.Exit(1)
	}

	if err := run(context.Background(), cfg, logger); err != nil {
		logger.Error("application error", "err", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, cfg *config.Config, logger *slog.Logger) error {
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"status":"ok"}`)
	})

	srv := &http.Server{
		Addr:         cfg.Server.Addr,
		Handler:      mux,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("starting server", "addr", cfg.Server.Addr)
		errCh <- srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			return err
		}
	case <-ctx.Done():
		logger.Info("shutting down")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	}
	return nil
}
```

**internal/config/config.go**:
```go
package config

import (
	"errors"
	"os"
	"time"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
}

type ServerConfig struct {
	Addr         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type DatabaseConfig struct {
	URL          string
	MaxOpenConns int
}

func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Addr:         envOr("SERVER_ADDR", ":8080"),
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
		},
		Database: DatabaseConfig{
			URL:          os.Getenv("DATABASE_URL"),
			MaxOpenConns: 25,
		},
	}

	if cfg.Database.URL == "" {
		return nil, errors.New("DATABASE_URL is required")
	}

	return cfg, nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
```

### CLI Tool

```bash
mkdir -p cmd/mytool internal
```

**cmd/mytool/main.go**:
```go
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
)

var version = "dev"

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	flags := flag.NewFlagSet("mytool", flag.ExitOnError)
	showVersion := flags.Bool("version", false, "print version")
	verbose := flags.Bool("verbose", false, "enable verbose output")

	if err := flags.Parse(args); err != nil {
		return err
	}

	if *showVersion {
		fmt.Printf("mytool %s\n", version)
		return nil
	}

	_ = ctx
	_ = verbose
	return nil
}
```

### Library Package

```bash
mkdir -p internal testdata
```

Minimal structure — just the package files, tests, and a go.mod.

## Configuration Files

### Makefile

```makefile
.PHONY: all build test lint fmt vet run clean

BINARY := myapp
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X main.version=$(VERSION)"

all: lint test build

build:
	go build $(LDFLAGS) -o bin/$(BINARY) ./cmd/server

run:
	go run ./cmd/server

test:
	go test -race -count=1 ./...

test-coverage:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

bench:
	go test -bench=. -benchmem -count=5 ./...

lint:
	golangci-lint run ./...

fmt:
	gofumpt -w .
	goimports -w .

vet:
	go vet ./...

clean:
	rm -rf bin/ coverage.out coverage.html

# Database
migrate-up:
	migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path migrations -database "$(DATABASE_URL)" down 1

migrate-create:
	migrate create -ext sql -dir migrations -seq $(name)

# Docker
docker-build:
	docker build -t $(BINARY):$(VERSION) .

docker-run:
	docker run --rm -p 8080:8080 --env-file .env $(BINARY):$(VERSION)
```

### Dockerfile (Multi-Stage)

```dockerfile
# Build stage
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X main.version=$(git describe --tags --always)" \
    -o /app/bin/server \
    ./cmd/server

# Runtime stage
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

COPY --from=builder /app/bin/server /usr/local/bin/server

EXPOSE 8080

ENTRYPOINT ["server"]
```

### .golangci.yml

```yaml
run:
  timeout: 5m
  go: "1.22"

linters:
  enable:
    - errcheck
    - govet
    - staticcheck
    - unused
    - gosimple
    - ineffassign
    - typecheck
    - gofumpt
    - goimports
    - misspell
    - unconvert
    - unparam
    - nilerr
    - errname
    - errorlint
    - exhaustive
    - prealloc
    - revive

linters-settings:
  errcheck:
    check-type-assertions: true
    check-blank: true
  govet:
    enable-all: true
  revive:
    rules:
      - name: exported
        arguments:
          - "checkPrivateReceivers"
      - name: unexported-return
      - name: blank-imports
      - name: context-as-argument
      - name: error-return
      - name: error-naming
      - name: increment-decrement
      - name: var-declaration
      - name: range
      - name: receiver-naming
      - name: time-naming
      - name: indent-error-flow
  exhaustive:
    default-signifies-exhaustive: true

issues:
  exclude-dirs:
    - vendor
  max-issues-per-linter: 0
  max-same-issues: 0
```

### GitHub Actions CI

```yaml
# .github/workflows/ci.yml
name: CI

on:
  push:
    branches: [main]
  pull_request:

jobs:
  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:16-alpine
        env:
          POSTGRES_DB: testdb
          POSTGRES_USER: test
          POSTGRES_PASSWORD: test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 5432:5432

    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: "1.22"
          cache: true

      - name: Download dependencies
        run: go mod download

      - name: Vet
        run: go vet ./...

      - name: Test
        run: go test -race -coverprofile=coverage.out ./...
        env:
          DATABASE_URL: postgres://test:test@localhost:5432/testdb?sslmode=disable

      - name: Build
        run: go build ./cmd/...

  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: "1.22"
          cache: true

      - name: golangci-lint
        uses: golangci/golangci-lint-action@v4
        with:
          version: latest
```

### .gitignore

```
# Binaries
bin/
*.exe

# Test
coverage.out
coverage.html
*.test
*.prof

# IDE
.idea/
.vscode/
*.swp
*.swo

# OS
.DS_Store
Thumbs.db

# Environment
.env
.env.local

# Vendor (uncomment if vendoring)
# vendor/

# Build
dist/
```

## Setup Checklist

After scaffolding, verify:

1. `go mod tidy` — clean dependencies
2. `go vet ./...` — passes
3. `go test ./...` — passes
4. `go build ./cmd/...` — compiles
5. `.gitignore` covers binaries and env files
6. `Makefile` targets work
7. `Dockerfile` builds successfully
8. CI configuration is valid

## Development Tooling Setup

```bash
# Install development tools
go install mvdan.cc/gofumpt@latest
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install golang.org/x/perf/cmd/benchstat@latest
go install github.com/air-verse/air@latest  # Hot reload for development

# Database migration tool
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

### Air configuration for hot reload (.air.toml):

```toml
root = "."
tmp_dir = "tmp"

[build]
  bin = "./tmp/main"
  cmd = "go build -o ./tmp/main ./cmd/server"
  delay = 1000
  exclude_dir = ["vendor", "tmp", "testdata"]
  exclude_regex = ["_test.go"]
  include_ext = ["go", "html", "tmpl"]
  kill_delay = "0s"

[log]
  time = false

[misc]
  clean_on_exit = true
```
