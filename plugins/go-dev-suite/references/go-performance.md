# Go Performance Reference

Memory management, GC tuning, escape analysis, and optimization patterns with concrete examples.

## Memory Model and Allocation

### Stack vs Heap

Go's compiler decides where to allocate variables via escape analysis. Stack allocation is free
(just a pointer increment); heap allocation requires GC tracking.

**Stack-allocated** (fast, no GC):
- Local variables that don't escape the function
- Small structs returned by value
- Fixed-size arrays
- Variables whose lifetime is bounded by the function

**Heap-allocated** (slower, GC-tracked):
- Variables whose pointer is returned or stored beyond the function
- Variables stored in interfaces
- Variables captured by closures that outlive the stack frame
- Large allocations that exceed stack size
- Slices, maps, and channels (their backing data)

### Escape Analysis in Practice

```bash
# Run escape analysis
go build -gcflags='-m' ./...

# More verbose
go build -gcflags='-m -m' ./...

# For a specific package
go build -gcflags='-m' ./internal/handler/
```

Common escape reasons and fixes:

```go
// ESCAPES: returning pointer to local
func newConfig() *Config {
    c := Config{Port: 8080} // escapes: returned by pointer
    return &c
}
// FIX: return by value if config is small
func newConfig() Config {
    return Config{Port: 8080} // stays on stack at call site
}

// ESCAPES: assigned to interface
func process(items []any) { /* ... */ }
var x int = 42
process([]any{x}) // x escapes: interface boxing
// FIX: use concrete types or generics
func process[T any](items []T) { /* ... */ }

// ESCAPES: slice grows beyond initial capacity
func collect() []int {
    var s []int // nil slice, will grow
    for i := range 100 {
        s = append(s, i)
    }
    return s
}
// FIX: pre-allocate
func collect() []int {
    s := make([]int, 0, 100) // one allocation, right size
    for i := range 100 {
        s = append(s, i)
    }
    return s
}

// ESCAPES: closure captures variable
func makeCounter() func() int {
    n := 0 // escapes: captured by closure that outlives function
    return func() int {
        n++
        return n
    }
}

// ESCAPES: sent to channel
func send(ch chan<- *Data) {
    d := &Data{} // escapes: sent through channel
    ch <- d
}
```

## Garbage Collector Internals

### How Go's GC Works

Go uses a **concurrent, tri-color, mark-sweep** garbage collector:

1. **Mark phase**: Traverses the object graph from roots (stack, globals), marking reachable objects
2. **Sweep phase**: Reclaims memory from unmarked objects
3. **Concurrent**: Most GC work happens alongside application goroutines
4. **Stop-the-world (STW) pauses**: Brief pauses at the start and end of marking (typically <1ms)

### GC Trigger Conditions

The GC runs when:
1. Heap grows to `GOGC%` larger than after last GC (default: 100% = 2x)
2. Approaching `GOMEMLIMIT` soft limit
3. `runtime.GC()` called explicitly
4. 2 minutes since last GC (forced periodic collection)

### GOGC Tuning

```bash
# Default: GC when heap doubles (100%)
GOGC=100 ./myapp

# GC when heap grows 50% → more frequent GC, less memory
GOGC=50 ./myapp

# GC when heap grows 200% → less frequent GC, more memory
GOGC=200 ./myapp

# Disable GC entirely (dangerous — only for benchmarks)
GOGC=off ./myapp
```

```go
// Programmatic control
import "runtime/debug"

// Set GOGC at runtime
old := debug.SetGCPercent(50)
defer debug.SetGCPercent(old)
```

### GOMEMLIMIT (Go 1.19+)

Soft memory limit. GC becomes more aggressive as usage approaches the limit.

```bash
# Set limit to 1GiB
GOMEMLIMIT=1GiB ./myapp

# For containers: use ~80-90% of container memory limit
# Container limit: 2GiB → GOMEMLIMIT=1800MiB
GOMEMLIMIT=1800MiB ./myapp
```

```go
// Auto-detect in containers
func setMemoryLimit() {
    // Read cgroup memory limit
    data, err := os.ReadFile("/sys/fs/cgroup/memory.max")
    if err != nil {
        return
    }
    limit, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64)
    if err != nil {
        return
    }
    // Use 80% of container limit
    debug.SetMemoryLimit(int64(float64(limit) * 0.8))
}
```

### GOGC + GOMEMLIMIT Together

Best practice for production:
- Set `GOMEMLIMIT` to protect against OOM
- Set `GOGC=100` (default) or higher for throughput
- GOMEMLIMIT prevents OOM; GOGC controls steady-state GC frequency

```bash
# Production container config
GOGC=100 GOMEMLIMIT=1800MiB ./myapp
```

### GC Diagnostics

```bash
# GC trace output
GODEBUG=gctrace=1 ./myapp 2>&1 | head -5

# Output format:
# gc 1 @0.004s 1%: 0.020+0.82+0.006 ms clock, 0.16+0.35/0.72/0+0.052 ms cpu, 4->4->1 MB, 4 MB goal, 0 MB stacks, 0 MB globals, 8 P
#
# gc#  @elapsed  wallclock%: STW-mark-start + concurrent + STW-mark-end ms clock
# CPU breakdown: assist + background/idle + dedicated ms
# heap-before -> heap-after -> live-heap MB, goal MB, stacks, globals, num-processors
```

```go
// Runtime memory stats
func logMemStats() {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)

    slog.Info("memory stats",
        slog.Uint64("heap_alloc_mb", m.HeapAlloc/1024/1024),
        slog.Uint64("heap_sys_mb", m.HeapSys/1024/1024),
        slog.Uint64("heap_objects", m.HeapObjects),
        slog.Uint64("gc_cycles", uint64(m.NumGC)),
        slog.Float64("gc_cpu_percent", m.GCCPUFraction*100),
        slog.Uint64("total_alloc_mb", m.TotalAlloc/1024/1024),
        slog.Int("goroutines", runtime.NumGoroutine()),
    )
}
```

## Allocation Reduction Techniques

### Pre-Allocation

```go
// Slices: always pre-allocate when size is known or estimable
users := make([]User, 0, len(rows))

// Maps: pre-allocate with expected size
index := make(map[string]*User, len(users))

// strings.Builder: pre-grow
var b strings.Builder
b.Grow(estimatedSize)

// bytes.Buffer: pre-grow
buf := bytes.NewBuffer(make([]byte, 0, 4096))
```

### Avoid Unnecessary Allocations

```go
// BAD: fmt.Sprintf allocates for simple cases
key := fmt.Sprintf("%s:%s", prefix, id) // 2+ allocs

// GOOD: strings.Builder or direct concatenation
key := prefix + ":" + id // 1 alloc

// BAD: creating intermediate slices
func getNames(users []User) []string {
    names := []string{} // nil would be fine
    for _, u := range users {
        names = append(names, u.Name)
    }
    return names
}

// GOOD: pre-allocated
func getNames(users []User) []string {
    names := make([]string, len(users))
    for i, u := range users {
        names[i] = u.Name
    }
    return names
}

// BAD: string to []byte conversion in loop
for _, s := range strings {
    hash.Write([]byte(s)) // allocs each iteration
}

// GOOD: reuse buffer
buf := make([]byte, 0, 256)
for _, s := range strings {
    buf = append(buf[:0], s...)
    hash.Write(buf)
}
```

### sync.Pool for Temporary Objects

```go
var bufPool = sync.Pool{
    New: func() any {
        return new(bytes.Buffer)
    },
}

func encodeResponse(data any) ([]byte, error) {
    buf := bufPool.Get().(*bytes.Buffer)
    defer func() {
        buf.Reset()
        bufPool.Put(buf)
    }()

    if err := json.NewEncoder(buf).Encode(data); err != nil {
        return nil, err
    }

    // Copy because buf returns to pool
    result := make([]byte, buf.Len())
    copy(result, buf.Bytes())
    return result, nil
}
```

### Reducing GC Scan Overhead

The GC must scan every pointer to find live objects. Large data structures with many pointers create
GC pressure even if the objects are long-lived.

```go
// HIGH GC PRESSURE: slice of pointers (GC scans every pointer)
type Cache struct {
    entries []*Entry // GC scans all N pointers every cycle
}

// LOWER GC PRESSURE: slice of values (no pointers to scan)
type Cache struct {
    entries []Entry // GC skips this (no pointers in Entry)
}

// LOWEST GC PRESSURE: flat byte layout
type Cache struct {
    keys   []byte // Single allocation, no pointers
    values []byte // Single allocation, no pointers
    index  []uint32 // Offsets into keys/values
}
```

## Profiling Reference

### CPU Profiling

```bash
# From tests
go test -cpuprofile cpu.prof -bench=BenchmarkX ./pkg/...
go tool pprof -http=:8080 cpu.prof

# From running server
curl -o cpu.prof "http://localhost:6060/debug/pprof/profile?seconds=30"
go tool pprof -http=:8080 cpu.prof

# pprof commands:
# top20           — top 20 functions by CPU time
# top -cum 20     — top 20 by cumulative time (including callees)
# list FuncName   — line-by-line CPU breakdown
# web             — call graph in browser
# peek FuncName   — show callers and callees
```

### Memory Profiling

```bash
# Heap profile (current live allocations)
go test -memprofile mem.prof -bench=BenchmarkX ./pkg/...
go tool pprof -http=:8080 mem.prof

# From running server
curl -o heap.prof "http://localhost:6060/debug/pprof/heap"
go tool pprof -http=:8080 heap.prof

# Allocation profile (cumulative since start)
curl -o allocs.prof "http://localhost:6060/debug/pprof/allocs"
go tool pprof -http=:8080 allocs.prof

# pprof modes:
# -inuse_space    — current bytes in use (default for heap)
# -inuse_objects  — current objects in use
# -alloc_space    — total bytes allocated (cumulative)
# -alloc_objects  — total objects allocated (cumulative)
go tool pprof -alloc_space mem.prof
```

### Goroutine Profiling

```bash
# Goroutine dump (human-readable)
curl "http://localhost:6060/debug/pprof/goroutine?debug=2" > goroutines.txt

# Goroutine profile (pprof format)
curl -o goroutine.prof "http://localhost:6060/debug/pprof/goroutine"
go tool pprof -http=:8080 goroutine.prof
```

### Block and Mutex Profiling

```go
// Enable block profiling (contention on channels/mutexes)
runtime.SetBlockProfileRate(1) // Record every block event

// Enable mutex profiling (lock contention)
runtime.SetMutexProfileFraction(5) // Sample 1/5 of lock events
```

```bash
curl -o block.prof "http://localhost:6060/debug/pprof/block"
curl -o mutex.prof "http://localhost:6060/debug/pprof/mutex"
```

### Execution Trace

The trace tool shows goroutine scheduling, GC events, and system calls on a timeline:

```bash
# Capture trace
curl -o trace.out "http://localhost:6060/debug/pprof/trace?seconds=5"

# From tests
go test -trace trace.out ./pkg/...

# View in browser
go tool trace trace.out
```

Trace shows:
- Goroutine creation, blocking, and scheduling
- GC phases and STW pauses
- System calls and network IO
- Processor (P) utilization

## Benchmark Methodology

### Writing Reliable Benchmarks

```go
func BenchmarkX(b *testing.B) {
    // Setup outside the loop
    data := generateData()
    b.ResetTimer() // Don't count setup time
    b.ReportAllocs()

    for b.Loop() {
        result := processData(data)
        // Prevent compiler from optimizing away the result
        _ = result
    }
}

// Use b.Loop() (Go 1.24+) instead of for i := 0; i < b.N; i++ {}
// b.Loop() handles warmup and iteration count automatically
```

### Running Benchmarks

```bash
# Run all benchmarks
go test -bench=. -benchmem ./...

# Run specific benchmark
go test -bench=BenchmarkParseJSON -benchmem ./pkg/parser/

# Multiple iterations for statistical significance
go test -bench=. -benchmem -count=10 ./... > bench.txt

# Compare with benchstat
go install golang.org/x/perf/cmd/benchstat@latest
benchstat old.txt new.txt

# Memory-only benchmark
go test -bench=BenchmarkX -benchmem -benchtime=1x ./...
```

### Benchmark Patterns

```go
// Parameterized benchmarks
func BenchmarkBuffer(b *testing.B) {
    for _, size := range []int{64, 256, 1024, 4096, 65536} {
        b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
            data := make([]byte, size)
            b.SetBytes(int64(size)) // Reports throughput (MB/s)
            b.ReportAllocs()
            for b.Loop() {
                var buf bytes.Buffer
                buf.Write(data)
            }
        })
    }
}

// Parallel benchmark (tests throughput under contention)
func BenchmarkCacheParallel(b *testing.B) {
    cache := NewCache()
    // Pre-populate
    for i := range 10000 {
        cache.Set(fmt.Sprintf("key-%d", i), i)
    }

    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        i := 0
        for pb.Next() {
            cache.Get(fmt.Sprintf("key-%d", i%10000))
            i++
        }
    })
}
```

## Common Optimization Patterns

### Struct Field Ordering

```go
// Before: 40 bytes (with padding)
type Inefficient struct {
    a bool      // 1 byte  + 7 padding
    b int64     // 8 bytes
    c bool      // 1 byte  + 3 padding
    d int32     // 4 bytes
    e bool      // 1 byte  + 7 padding
    f int64     // 8 bytes
}

// After: 24 bytes (minimal padding)
type Efficient struct {
    b int64     // 8 bytes
    f int64     // 8 bytes
    d int32     // 4 bytes
    a bool      // 1 byte
    c bool      // 1 byte
    e bool      // 1 byte + 1 padding
}
```

```bash
# Automatic fix
go install golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment@latest
fieldalignment -fix ./...
```

### String Interning

For repeated strings (e.g., status values, country codes):

```go
var internPool sync.Map

func intern(s string) string {
    if v, ok := internPool.Load(s); ok {
        return v.(string)
    }
    // Store a copy to allow original to be GC'd
    interned := strings.Clone(s)
    v, _ := internPool.LoadOrStore(interned, interned)
    return v.(string)
}

// Usage: when parsing many records with repetitive string fields
record.Status = intern(rawStatus) // Shares one string instance
```

### Avoid Reflection in Hot Paths

```go
// SLOW: reflection-based
func encodeAny(w io.Writer, v any) error {
    return json.NewEncoder(w).Encode(v) // Uses reflect internally
}

// FAST: code-generated or manual marshaling
func (u *User) MarshalJSON() ([]byte, error) {
    var buf [256]byte
    b := buf[:0]
    b = append(b, `{"id":"`...)
    b = append(b, u.ID...)
    b = append(b, `","name":"`...)
    b = append(b, u.Name...)
    b = append(b, `","email":"`...)
    b = append(b, u.Email...)
    b = append(b, `"}`...)
    return b, nil
}
```

### IO Buffering

```go
// SLOW: many small writes
func writeRecords(w io.Writer, records []Record) error {
    for _, r := range records {
        fmt.Fprintf(w, "%s,%d\n", r.Name, r.Value) // Syscall per record
    }
    return nil
}

// FAST: buffered writes
func writeRecords(w io.Writer, records []Record) error {
    bw := bufio.NewWriterSize(w, 64*1024) // 64KB buffer
    defer bw.Flush()
    for _, r := range records {
        fmt.Fprintf(bw, "%s,%d\n", r.Name, r.Value) // Buffered
    }
    return bw.Flush()
}

// FAST: buffered reads
func readLines(r io.Reader) ([]string, error) {
    scanner := bufio.NewScanner(r)
    scanner.Buffer(make([]byte, 64*1024), 1024*1024) // 64KB initial, 1MB max
    var lines []string
    for scanner.Scan() {
        lines = append(lines, scanner.Text())
    }
    return lines, scanner.Err()
}
```

### Connection Pooling

```go
// HTTP client: reuse connections
var httpClient = &http.Client{
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 20,
        IdleConnTimeout:     90 * time.Second,
    },
    Timeout: 30 * time.Second,
}
// CRITICAL: Always drain and close response body to return connection to pool
resp, err := httpClient.Do(req)
if err != nil { return err }
defer resp.Body.Close()
io.Copy(io.Discard, resp.Body) // Drain body

// Database: tune pool settings
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(5)
db.SetConnMaxLifetime(5 * time.Minute)
```

## Performance Checklist

**Before optimizing:**
- [ ] Have benchmarks with `b.ReportAllocs()` and `b.Loop()`
- [ ] Have baseline measurements saved (`go test -bench=. -count=10 > baseline.txt`)
- [ ] Have identified the bottleneck via profiling (CPU, memory, or contention)

**Allocation reduction:**
- [ ] Pre-allocate slices and maps with known/estimated sizes
- [ ] Use `strings.Builder` with `Grow()` for string concatenation
- [ ] Return small structs by value instead of pointer
- [ ] Use `sync.Pool` for frequently allocated temporary objects
- [ ] Avoid `fmt.Sprintf` in hot paths — use direct string operations
- [ ] Minimize interface boxing in performance-critical code

**CPU optimization:**
- [ ] Use `range` loops (eliminates bounds checks)
- [ ] Consider struct-of-arrays layout for cache-friendly iteration
- [ ] Minimize virtual dispatch (interface method calls) in tight loops
- [ ] Use buffered IO for many small reads/writes

**GC pressure:**
- [ ] Order struct fields by decreasing alignment to reduce padding
- [ ] Minimize pointer-heavy data structures (GC scan overhead)
- [ ] Set `GOMEMLIMIT` in containers to prevent OOM
- [ ] Monitor GC metrics in production (`runtime.MemStats`)

**After optimizing:**
- [ ] Run benchstat to verify improvement is statistically significant
- [ ] Verify no regressions with `go test -race ./...`
- [ ] Document why the optimization matters (hot path, latency-sensitive, etc.)
