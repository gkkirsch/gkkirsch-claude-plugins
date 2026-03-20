---
name: go-concurrency
description: >
  Go concurrency patterns — goroutines, channels, sync primitives, worker pools,
  fan-out/fan-in, context cancellation, and common concurrency pitfalls.
  Triggers: "goroutine", "go channel", "go concurrency", "go sync", "go worker pool",
  "go fan out", "go context", "go mutex", "go select", "go waitgroup".
  NOT for: HTTP/web API patterns (use go-web-apis).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Go Concurrency Patterns

## Goroutine Basics

```go
// Always handle goroutine lifecycle — never fire-and-forget
func processItems(items []Item) error {
    var wg sync.WaitGroup
    errCh := make(chan error, len(items))

    for _, item := range items {
        wg.Add(1)
        go func(item Item) {
            defer wg.Done()
            if err := process(item); err != nil {
                errCh <- fmt.Errorf("process %s: %w", item.ID, err)
            }
        }(item)
    }

    wg.Wait()
    close(errCh)

    // Collect errors
    var errs []error
    for err := range errCh {
        errs = append(errs, err)
    }
    if len(errs) > 0 {
        return errors.Join(errs...)
    }
    return nil
}
```

## Channel Patterns

### Generator

```go
// Generator returns a read-only channel
func generateIDs(ctx context.Context) <-chan string {
    ch := make(chan string)
    go func() {
        defer close(ch)
        for i := 0; ; i++ {
            id := fmt.Sprintf("id-%d-%d", time.Now().UnixNano(), i)
            select {
            case ch <- id:
            case <-ctx.Done():
                return
            }
        }
    }()
    return ch
}

// Usage
ctx, cancel := context.WithCancel(context.Background())
defer cancel()
ids := generateIDs(ctx)
id1 := <-ids
id2 := <-ids
```

### Fan-Out / Fan-In

```go
// Fan-out: distribute work across multiple goroutines
// Fan-in: merge results into a single channel
func fanOutFanIn(ctx context.Context, urls []string, workers int) []Result {
    // Fan-out: send work to a shared channel
    jobs := make(chan string, len(urls))
    go func() {
        defer close(jobs)
        for _, url := range urls {
            select {
            case jobs <- url:
            case <-ctx.Done():
                return
            }
        }
    }()

    // Fan-out: start workers
    results := make([]<-chan Result, workers)
    for i := 0; i < workers; i++ {
        results[i] = worker(ctx, jobs)
    }

    // Fan-in: merge all result channels
    var all []Result
    for result := range merge(ctx, results...) {
        all = append(all, result)
    }
    return all
}

func worker(ctx context.Context, jobs <-chan string) <-chan Result {
    out := make(chan Result)
    go func() {
        defer close(out)
        for url := range jobs {
            select {
            case <-ctx.Done():
                return
            default:
                result := fetch(url)
                select {
                case out <- result:
                case <-ctx.Done():
                    return
                }
            }
        }
    }()
    return out
}

func merge[T any](ctx context.Context, channels ...<-chan T) <-chan T {
    var wg sync.WaitGroup
    merged := make(chan T)

    for _, ch := range channels {
        wg.Add(1)
        go func(c <-chan T) {
            defer wg.Done()
            for v := range c {
                select {
                case merged <- v:
                case <-ctx.Done():
                    return
                }
            }
        }(ch)
    }

    go func() {
        wg.Wait()
        close(merged)
    }()

    return merged
}
```

### Pipeline

```go
// Pipeline: chain processing stages with channels
func pipeline(ctx context.Context, input []int) <-chan int {
    // Stage 1: Generate
    gen := func() <-chan int {
        out := make(chan int)
        go func() {
            defer close(out)
            for _, n := range input {
                select {
                case out <- n:
                case <-ctx.Done():
                    return
                }
            }
        }()
        return out
    }

    // Stage 2: Square
    square := func(in <-chan int) <-chan int {
        out := make(chan int)
        go func() {
            defer close(out)
            for n := range in {
                select {
                case out <- n * n:
                case <-ctx.Done():
                    return
                }
            }
        }()
        return out
    }

    // Stage 3: Filter (keep even)
    filter := func(in <-chan int) <-chan int {
        out := make(chan int)
        go func() {
            defer close(out)
            for n := range in {
                if n%2 == 0 {
                    select {
                    case out <- n:
                    case <-ctx.Done():
                        return
                    }
                }
            }
        }()
        return out
    }

    return filter(square(gen()))
}
```

## Worker Pool

```go
type WorkerPool[T any, R any] struct {
    workers    int
    jobs       chan T
    results    chan R
    processor  func(context.Context, T) (R, error)
    errHandler func(T, error)
}

func NewWorkerPool[T any, R any](
    workers int,
    processor func(context.Context, T) (R, error),
    errHandler func(T, error),
) *WorkerPool[T, R] {
    return &WorkerPool[T, R]{
        workers:    workers,
        jobs:       make(chan T, workers*2),
        results:    make(chan R, workers*2),
        processor:  processor,
        errHandler: errHandler,
    }
}

func (p *WorkerPool[T, R]) Start(ctx context.Context) {
    var wg sync.WaitGroup

    for i := 0; i < p.workers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for job := range p.jobs {
                select {
                case <-ctx.Done():
                    return
                default:
                    result, err := p.processor(ctx, job)
                    if err != nil {
                        if p.errHandler != nil {
                            p.errHandler(job, err)
                        }
                        continue
                    }
                    p.results <- result
                }
            }
        }()
    }

    // Close results when all workers are done
    go func() {
        wg.Wait()
        close(p.results)
    }()
}

func (p *WorkerPool[T, R]) Submit(job T) {
    p.jobs <- job
}

func (p *WorkerPool[T, R]) Close() {
    close(p.jobs)
}

func (p *WorkerPool[T, R]) Results() <-chan R {
    return p.results
}

// Usage
pool := NewWorkerPool(10,
    func(ctx context.Context, url string) (Response, error) {
        return httpGet(ctx, url)
    },
    func(url string, err error) {
        slog.Error("fetch failed", "url", url, "error", err)
    },
)

ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

pool.Start(ctx)

// Submit jobs
go func() {
    for _, url := range urls {
        pool.Submit(url)
    }
    pool.Close()
}()

// Collect results
for result := range pool.Results() {
    fmt.Println(result.StatusCode)
}
```

## Context Patterns

```go
// Always propagate context — never use context.Background() mid-chain
func handleRequest(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context() // Already has request deadline

    // Add timeout for specific operation
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    // Add values for tracing
    ctx = context.WithValue(ctx, "requestID", uuid.NewString())

    result, err := fetchData(ctx)
    if err != nil {
        if errors.Is(err, context.DeadlineExceeded) {
            http.Error(w, "request timeout", http.StatusGatewayTimeout)
            return
        }
        if errors.Is(err, context.Canceled) {
            return // Client disconnected
        }
        http.Error(w, "internal error", http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(result)
}

// Cancellation-aware function
func fetchData(ctx context.Context) (*Data, error) {
    ch := make(chan *Data, 1)
    errCh := make(chan error, 1)

    go func() {
        data, err := slowOperation()
        if err != nil {
            errCh <- err
            return
        }
        ch <- data
    }()

    select {
    case data := <-ch:
        return data, nil
    case err := <-errCh:
        return nil, err
    case <-ctx.Done():
        return nil, ctx.Err()
    }
}
```

## Sync Primitives

```go
// Mutex — protect shared state
type SafeCounter struct {
    mu sync.RWMutex
    counts map[string]int
}

func (c *SafeCounter) Increment(key string) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.counts[key]++
}

func (c *SafeCounter) Get(key string) int {
    c.mu.RLock()        // Multiple readers OK
    defer c.mu.RUnlock()
    return c.counts[key]
}

// Once — initialize exactly once
type Client struct {
    once sync.Once
    conn *Connection
}

func (c *Client) getConn() *Connection {
    c.once.Do(func() {
        c.conn = connect() // Runs exactly once, even with concurrent callers
    })
    return c.conn
}

// Semaphore — limit concurrency
func processWithLimit(ctx context.Context, items []Item, maxConcurrent int) error {
    sem := make(chan struct{}, maxConcurrent)
    var wg sync.WaitGroup
    errCh := make(chan error, len(items))

    for _, item := range items {
        wg.Add(1)
        go func(item Item) {
            defer wg.Done()

            // Acquire semaphore slot
            select {
            case sem <- struct{}{}:
                defer func() { <-sem }() // Release
            case <-ctx.Done():
                errCh <- ctx.Err()
                return
            }

            if err := process(ctx, item); err != nil {
                errCh <- err
            }
        }(item)
    }

    wg.Wait()
    close(errCh)

    var errs []error
    for err := range errCh {
        errs = append(errs, err)
    }
    return errors.Join(errs...)
}

// errgroup — structured concurrency
import "golang.org/x/sync/errgroup"

func fetchAll(ctx context.Context, urls []string) ([]Response, error) {
    g, ctx := errgroup.WithContext(ctx)
    responses := make([]Response, len(urls))

    for i, url := range urls {
        i, url := i, url // Capture loop vars
        g.Go(func() error {
            resp, err := fetch(ctx, url)
            if err != nil {
                return fmt.Errorf("fetch %s: %w", url, err)
            }
            responses[i] = resp
            return nil
        })
    }

    if err := g.Wait(); err != nil {
        return nil, err
    }
    return responses, nil
}
```

## Rate Limiter

```go
import "golang.org/x/time/rate"

// Token bucket rate limiter
type RateLimiter struct {
    limiters map[string]*rate.Limiter
    mu       sync.RWMutex
    rate     rate.Limit
    burst    int
}

func NewRateLimiter(r rate.Limit, burst int) *RateLimiter {
    return &RateLimiter{
        limiters: make(map[string]*rate.Limiter),
        rate:     r,
        burst:    burst,
    }
}

func (rl *RateLimiter) GetLimiter(key string) *rate.Limiter {
    rl.mu.RLock()
    limiter, exists := rl.limiters[key]
    rl.mu.RUnlock()

    if exists {
        return limiter
    }

    rl.mu.Lock()
    defer rl.mu.Unlock()

    // Double-check after acquiring write lock
    if limiter, exists = rl.limiters[key]; exists {
        return limiter
    }

    limiter = rate.NewLimiter(rl.rate, rl.burst)
    rl.limiters[key] = limiter
    return limiter
}

func (rl *RateLimiter) Allow(key string) bool {
    return rl.GetLimiter(key).Allow()
}

func (rl *RateLimiter) Wait(ctx context.Context, key string) error {
    return rl.GetLimiter(key).Wait(ctx)
}
```

## Ticker and Periodic Tasks

```go
func startPeriodicTask(ctx context.Context, interval time.Duration, task func(context.Context) error) {
    ticker := time.NewTicker(interval)
    defer ticker.Stop()

    // Run immediately on start
    if err := task(ctx); err != nil {
        slog.Error("periodic task failed", "error", err)
    }

    for {
        select {
        case <-ticker.C:
            if err := task(ctx); err != nil {
                slog.Error("periodic task failed", "error", err)
            }
        case <-ctx.Done():
            slog.Info("periodic task stopped")
            return
        }
    }
}

// Usage
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

go startPeriodicTask(ctx, 5*time.Minute, func(ctx context.Context) error {
    return cleanupExpiredSessions(ctx)
})
```

## Gotchas

1. **Goroutine leak from unbuffered channel** — If a goroutine sends to an unbuffered channel but nobody reads, the goroutine blocks forever. Always use `select` with `ctx.Done()` or use buffered channels when the receiver might not consume.

2. **Loop variable capture in goroutines** — `go func() { use(item) }()` inside a `for _, item := range items` captures the loop variable by reference. In Go < 1.22, all goroutines see the last value. Fix: pass as parameter `go func(item Item) { ... }(item)` or use Go 1.22+ which fixes this.

3. **Closing a channel twice panics** — Only the sender should close a channel, and only once. If multiple goroutines send to the same channel, use a `sync.WaitGroup` to coordinate closing after all senders are done.

4. **`sync.Mutex` is not reentrant** — Locking a mutex twice from the same goroutine deadlocks. If method A holds the lock and calls method B which also needs the lock, refactor: extract an internal `locked_` method or restructure.

5. **`context.WithValue` is not a map** — Don't store 10 values in context. It's a linked list, O(n) lookup. Use a single struct as the value: `ctx = context.WithValue(ctx, reqKey, &RequestInfo{ID: id, UserID: uid})`.

6. **Race detector only catches races during execution** — `go test -race` only detects races in code paths actually exercised. Full coverage doesn't mean race-free. Run `-race` in CI with as much test coverage as possible, and use production workload tests.
