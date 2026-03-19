# Async Rust Guide — Reference

---

## 1. Async Runtime Architecture

An async runtime drives `async fn` to completion. Rust ships no runtime — you choose one.
Every runtime provides four core components:

| Component | Responsibility |
|-----------|---------------|
| **Task scheduler** | Decides which futures to poll and on which threads |
| **I/O driver** | Integrates with OS async I/O (epoll, kqueue, io_uring, IOCP) |
| **Timer driver** | Manages `sleep`, `timeout`, and `interval` futures |
| **Reactor / Poller** | Wakes tasks when I/O readiness or timer expiry occurs |

### Tokio

Tokio offers two scheduler flavors:

- **Multi-threaded** (`#[tokio::main]`): Work-stealing across a thread pool (one thread per core). Tasks must be `Send`.
- **Current-thread** (`#[tokio::main(flavor = "current_thread")]`): Single thread. Tasks need not be `Send`. Good for CLIs and WASM.

```rust
#[tokio::main]                                    // multi-threaded (default)
async fn main() { serve().await; }

#[tokio::main(flavor = "current_thread")]          // single-threaded
async fn main() { serve().await; }
```

### async-std and smol

**async-std** mirrors the standard library API with async equivalents. Global thread-pool executor. Smaller ecosystem than Tokio.
**smol** is minimal (~1500 lines across sub-crates: `async-io`, `async-executor`, `blocking`). Best for embedded or small-dependency builds.

### Runtime Selection Guide

| Criterion | Tokio | async-std | smol |
|-----------|-------|-----------|------|
| Ecosystem size | Largest | Moderate | Small |
| Maturity | Production-proven | Stable | Stable |
| Work-stealing | Yes | Yes | Yes (async-executor) |
| io_uring support | tokio-uring (separate) | No | async-io polling |
| WASM support | Limited (current-thread) | Limited | Yes |
| Dependency weight | ~30 crates | ~20 crates | ~10 crates |
| Best for | Servers, gRPC, HTTP | std-like API fans | Embedded, minimal builds |

When in doubt, choose Tokio. Its ecosystem (hyper, tonic, axum, tower) is unmatched.

---

## 2. Futures & the Poll Model

### The Future Trait

```rust
pub trait Future {
    type Output;
    fn poll(self: Pin<&mut Self>, cx: &mut Context<'_>) -> Poll<Self::Output>;
}

pub enum Poll<T> {
    Ready(T),
    Pending,
}
```

- `poll` is called by the runtime, never by user code directly.
- `Poll::Ready(value)` completes the future. `Poll::Pending` means "wake me later."
- The `Context` carries a `Waker` that the future must use to request re-polling.

### How async/await Desugars

The compiler transforms each `async fn` into a state machine implementing `Future`. Each `.await` becomes a state transition:

```rust
async fn fetch_and_parse(url: &str) -> Result<Data> {
    let response = reqwest::get(url).await?;     // state 0 -> 1
    let body = response.text().await?;            // state 1 -> 2
    serde_json::from_str(&body)                   // state 2 -> Ready
}
// Compiles to an enum with one variant per state, each capturing live locals.
// This is why async futures can be self-referential.
```

### Waker and Wake-up Notifications

When a future returns `Poll::Pending`, it *must* arrange for `cx.waker().wake()` to be called later. Without this, the task stalls silently. The runtime's wakers enqueue the task for re-polling.

### Why Futures Are Lazy

Unlike JavaScript Promises, Rust futures do nothing until polled. `let fut = async { work().await };` allocates the state machine but performs no work until `fut.await`. This enables zero-cost cancellation (just drop the future) and prevents accidental concurrency.

### Manual Future Implementation

```rust
use std::{future::Future, pin::Pin, task::{Context, Poll}, time::Instant};

struct Delay { when: Instant }

impl Future for Delay {
    type Output = ();
    fn poll(self: Pin<&mut Self>, cx: &mut Context<'_>) -> Poll<()> {
        if Instant::now() >= self.when {
            return Poll::Ready(());
        }
        let when = self.when;
        let waker = cx.waker().clone();
        std::thread::spawn(move || { std::thread::sleep(when - Instant::now()); waker.wake(); });
        Poll::Pending
    }
}
// Note: spawning a thread per poll is illustrative. Production timers use the runtime's driver.
```

---

## 3. Tokio Concurrency Primitives

### `tokio::spawn` — Fire-and-Forget Tasks

```rust
let handle = tokio::spawn(async { do_work().await });
let value = handle.await?; // JoinError on panic or cancellation
```

The future must be `Send + 'static`.

### `JoinSet` — Managing Multiple Spawned Tasks

```rust
use tokio::task::JoinSet;
let mut set = JoinSet::new();
for url in urls {
    set.spawn(async move { fetch(url).await });
}
while let Some(result) = set.join_next().await {
    match result {
        Ok(data) => process(data),
        Err(e) => tracing::error!("Task failed: {e}"),
    }
}
```

### `tokio::select!` — Racing Multiple Futures

Polls multiple futures concurrently; executes the first branch that completes. All others are **dropped** (cancelled):

```rust
tokio::select! {
    val = future_a => { /* a completed first */ }
    val = future_b => { /* b completed first */ }
}
```

**Cancellation safety warning**: losing branches are dropped mid-execution. See Section 5.

### `tokio::join!` — Awaiting Multiple Futures Concurrently

```rust
let (users, orders, inventory) = tokio::join!(
    fetch_users(), fetch_orders(), fetch_inventory(),
);
```

No future is cancelled. All results returned in a tuple.

### Channel Primitives

**`mpsc`** — multi-producer, single-consumer with backpressure:
```rust
let (tx, mut rx) = mpsc::channel::<Command>(100); // buffer size 100
tokio::spawn(async move { tx.send(Command::Process(data)).await.unwrap(); });
while let Some(cmd) = rx.recv().await { handle(cmd); } // None when all senders drop
```

**`oneshot`** — single-use request-response:
```rust
let (tx, rx) = oneshot::channel();
tokio::spawn(async move { let _ = tx.send(compute().await); });
let value = rx.await?;
```

**`broadcast`** — every receiver gets every message:
```rust
let (tx, _) = broadcast::channel::<Event>(16);
let mut rx1 = tx.subscribe();
tx.send(Event::Tick)?; // rx1.recv().await returns Event::Tick
```

**`watch`** — latest-value only, intermediate values may be skipped:
```rust
let (tx, mut rx) = watch::channel(AppConfig::default());
tx.send(new_config)?;
loop { rx.changed().await?; apply_config(rx.borrow().clone()); }
```

### Synchronization Primitives

**`Mutex`** — use when holding a lock across `.await` (prefer `std::sync::Mutex` for short sync sections):
```rust
let data = Arc::new(tokio::sync::Mutex::new(HashMap::new()));
let mut map = data.lock().await;
map.insert("key", expensive_async_lookup().await); // held across .await — OK
```

**`Semaphore`** — limit concurrency:
```rust
let sem = Arc::new(Semaphore::new(10));
for req in requests {
    let permit = sem.clone().acquire_owned().await?;
    tokio::spawn(async move { process(req).await; drop(permit); });
}
```

**`Notify`** — lightweight signal with no data:
```rust
let notify = Arc::new(Notify::new());
let n = notify.clone();
tokio::spawn(async move { work().await; n.notify_one(); });
notify.notified().await;
```

---

## 4. Pinning

### Why Pinning Exists

Async state machines can be self-referential: a borrow in one state may point to data in the same struct. If the struct moves, the pointer dangles. `Pin<P>` wraps a pointer and guarantees the pointee will not move.

```rust
async fn example() {
    let data = vec![1, 2, 3];
    let reference = &data;
    some_async_op().await; // generated struct holds both `data` and `reference`
    println!("{reference:?}");
}
```

### `Unpin`, `Pin<Box<T>>`, and the `pin!` Macro

Most types implement `Unpin` automatically — pinning has no effect, they move freely. Only compiler-generated futures and types using `PhantomPinned` are `!Unpin`.

```rust
let boxed: Pin<Box<MyFuture>> = Box::pin(MyFuture::new()); // heap-pinned
let fut = std::pin::pin!(async { work().await });           // stack-pinned (Rust 1.68+)
```

### When You Encounter Pinning

1. **`Stream` implementations** — `poll_next` takes `Pin<&mut Self>`.
2. **Manual `Future` implementations** — `poll` takes `Pin<&mut Self>`.
3. **`select!` with reused futures** — must be pinned to survive loop iterations.

### Common Error: "the trait `Unpin` is not implemented"

```rust
// Fix 1: Box::pin (heap allocation)
let fut: Pin<Box<dyn Future<Output = ()>>> = Box::pin(my_future);

// Fix 2: pin! macro (stack, no heap cost)
let fut = std::pin::pin!(my_future);

// Fix 3: add Unpin bound (only if caller's futures are Unpin)
async fn run<F: Future + Unpin>(f: F) -> F::Output { f.await }
```

---

## 5. Cancellation Safety

### What Cancellation Means

In `tokio::select!`, when one branch completes, all others are dropped. If a future held partial state (e.g., half-read bytes), that state is lost.

### Cancellation Safety Table

| Operation | Safe? | Notes |
|-----------|-------|-------|
| `mpsc::Receiver::recv()` | Yes | Message stays in channel |
| `oneshot::Receiver` (`.await`) | Yes | Value not consumed until Ready |
| `broadcast::Receiver::recv()` | Yes | Message stays in channel |
| `watch::Receiver::changed()` | Yes | Value persists |
| `TcpStream::read()` | **No** | Partial reads lost |
| `TcpStream::write()` | **No** | Partial writes lost |
| `AsyncReadExt::read_exact()` | **No** | Partial progress lost |
| `AsyncBufReadExt::read_line()` | **No** | Partial line lost |
| `Mutex::lock()` | Yes | No state change until acquired |
| `Semaphore::acquire()` | Yes | Permit not consumed until returned |
| `tokio::time::sleep()` | Yes | Stateless |
| `JoinHandle` (`.await`) | Yes | Task continues; result cached |
| `Stream::next()` | Depends | Safe if `poll_next` is safe |

### Making Operations Cancellation-Safe

Preserve partial state outside `select!` so it survives cancellation:

```rust
let mut buf = Vec::new(); // persists across loop iterations
loop { tokio::select! {
    result = reader.read_buf(&mut buf) => {
        if result? == 0 { break; }
        if let Some(msg) = try_parse(&mut buf) { process(msg); }
    }
    _ = shutdown.recv() => { break; }
}}
```

### Pinning Futures for Reuse Across `select!` Iterations

Pin a long-running future outside the loop to avoid restarting it each iteration:

```rust
let operation = some_long_operation();
tokio::pin!(operation);
loop {
    tokio::select! {
        result = &mut operation => { handle(result); break; }
        msg = rx.recv() => { if let Some(m) = msg { process(m); } }
    }
}
```

### Structured Concurrency

Prefer explicit cancellation over implicit `select!` drops. Dedicate a task with a shutdown channel so `cleanup()` always runs:

```rust
let (shutdown_tx, mut shutdown_rx) = mpsc::channel::<()>(1);
let worker = tokio::spawn(async move {
    loop { tokio::select! {
        Some(msg) = rx.recv() => { process(msg).await; }
        _ = shutdown_rx.recv() => { break; }
    }}
    cleanup().await;
});
drop(shutdown_tx); worker.await?;
```

---

## 6. Common Async Patterns

### Graceful Shutdown

Broadcast channel + `tokio::signal` coordinates shutdown across all tasks:

```rust
async fn run_server(mut shutdown_rx: broadcast::Receiver<()>) {
    let listener = TcpListener::bind("0.0.0.0:8080").await.unwrap();
    loop {
        tokio::select! {
            Ok((stream, addr)) = listener.accept() => {
                tokio::spawn(handle_connection(stream, addr, shutdown_rx.resubscribe()));
            }
            _ = shutdown_rx.recv() => { tracing::info!("Shutting down"); break; }
        }
    }
}

#[tokio::main]
async fn main() {
    let (tx, rx) = broadcast::channel(1);
    let server = tokio::spawn(run_server(rx));
    tokio::signal::ctrl_c().await.unwrap();
    let _ = tx.send(());
    server.await.unwrap();
}
```

### Rate Limiting with Token Bucket

```rust
fn spawn_rate_limiter(rps: usize) -> Arc<Semaphore> {
    let sem = Arc::new(Semaphore::new(0));
    let s = sem.clone();
    tokio::spawn(async move {
        let mut tick = tokio::time::interval(Duration::from_secs(1));
        loop { tick.tick().await; s.add_permits(rps - s.available_permits().min(rps)); }
    });
    sem
}
```

### Retry with Exponential Backoff

```rust
async fn retry_with_backoff<F, Fut, T, E>(mut op: F, max: u32) -> Result<T, E>
where F: FnMut() -> Fut, Fut: Future<Output = Result<T, E>>, E: std::fmt::Display {
    let mut attempt = 0;
    loop {
        match op().await {
            Ok(v) => return Ok(v),
            Err(e) if attempt < max => {
                let wait = Duration::from_millis(100 * 2u64.pow(attempt))
                    + Duration::from_millis(rand::random::<u64>() % 100);
                tracing::warn!("Attempt {} failed: {e}. Retrying in {wait:?}", attempt + 1);
                tokio::time::sleep(wait).await;
                attempt += 1;
            }
            Err(e) => return Err(e),
        }
    }
}
```

### Fan-Out / Fan-In

Spawn work in parallel with bounded concurrency, collect results via `JoinSet`:

```rust
async fn fan_out_fan_in(urls: Vec<String>) -> Vec<Result<Response, Error>> {
    let mut set = JoinSet::new();
    let sem = Arc::new(Semaphore::new(20));
    for url in urls {
        let s = sem.clone();
        set.spawn(async move { let _p = s.acquire().await.unwrap(); reqwest::get(&url).await });
    }
    let mut results = Vec::new();
    while let Some(res) = set.join_next().await {
        match res { Ok(v) => results.push(v), Err(e) => tracing::error!("Panic: {e}") }
    }
    results
}
```

### Async Streaming

The `Stream` trait (`futures`/`tokio-stream`) is the async `Iterator`. Use `async-stream` for ergonomic creation:

```rust
use tokio_stream::{StreamExt, wrappers::ReceiverStream};
let results: Vec<_> = ReceiverStream::new(rx)
    .filter(|i| i.is_valid()).map(transform).take(50).collect().await;

fn heartbeats() -> impl Stream<Item = Event> {
    async_stream::stream! {
        let mut tick = tokio::time::interval(Duration::from_secs(1));
        loop { tick.tick().await; yield Event::Heartbeat; }
    }
}
```

### Background Worker with Message Channel

Use a command enum with `oneshot` reply channels for request-response:

```rust
enum Cmd {
    Process { data: Vec<u8>, reply: oneshot::Sender<ProcessResult> },
    Shutdown,
}

async fn worker_loop(mut rx: mpsc::Receiver<Cmd>) {
    while let Some(cmd) = rx.recv().await {
        match cmd {
            Cmd::Process { data, reply } => { let _ = reply.send(do_work(data)); }
            Cmd::Shutdown => break,
        }
    }
}

// Ergonomic handle:
#[derive(Clone)]
struct WorkerHandle { tx: mpsc::Sender<Cmd> }
impl WorkerHandle {
    fn new() -> Self {
        let (tx, rx) = mpsc::channel(32);
        tokio::spawn(worker_loop(rx));
        Self { tx }
    }
    async fn process(&self, data: Vec<u8>) -> Result<ProcessResult, Error> {
        let (reply, resp) = oneshot::channel();
        self.tx.send(Cmd::Process { data, reply }).await.map_err(|_| Error::WorkerGone)?;
        resp.await.map_err(|_| Error::WorkerGone)
    }
}
```

### Connection Pooling with `deadpool`

Implement `deadpool::managed::Manager` with `create` and `recycle`, then build a pool:

```rust
use deadpool::managed::{Manager, Pool, RecycleResult};
struct DbMgr { url: String }
impl Manager for DbMgr {
    type Type = DbConn; type Error = DbError;
    async fn create(&self) -> Result<DbConn, DbError> { DbConn::connect(&self.url).await }
    async fn recycle(&self, c: &mut DbConn, _: &deadpool::managed::Metrics)
        -> RecycleResult<DbError> { c.ping().await.map_err(Into::into) }
}
let pool = Pool::builder(DbMgr { url: "host=localhost".into() }).max_size(16).build()?;
let conn = pool.get().await?; // returned to pool on drop
```

---

## 7. Send + Sync in Async

### Why Spawned Tasks Require `Send`

`tokio::spawn` moves the future to a thread pool where any worker may poll it. The future and all data it captures across `.await` points must be `Send`:

```rust
// Compiles — String: Send
tokio::spawn(async { let s = String::from("hello"); work().await; println!("{s}"); });

// Does NOT compile — Rc: !Send
let rc = Rc::new(42);
tokio::spawn(async move { work().await; println!("{rc}"); });
```

### Why Holding a Non-Send Type Across `.await` Is an Error

The compiler checks which types are alive at each `.await`. A `!Send` type in scope across `.await` makes the entire generated future `!Send`.

### Common Fixes

```rust
// Fix 1: Drop before .await
async fn fixed() {
    { let data = Rc::new(vec![1, 2, 3]); println!("{data:?}"); } // Rc dropped
    something().await; // future is Send
}

// Fix 2: Use Arc instead of Rc
let data = Arc::new(vec![1, 2, 3]);
something().await;
println!("{data:?}"); // Arc: Send + Sync

// Fix 3: Extract the value before the yield point
let value = { let data = Rc::new(vec![1, 2, 3]); data.iter().sum::<i32>() };
something().await;
```

### Reading the "future is not Send" Diagnostic

Read from the bottom up. The compiler names the `!Send` type and the `.await` where it is held. Fix whichever variable the diagnostic identifies using the strategies above.

---

## 8. Common Pitfalls

### Holding `std::sync::Mutex` Across `.await`

Blocks the OS thread, starving other tasks. Fix: use `tokio::sync::Mutex`, or minimize the critical section so no `.await` is held:

```rust
// FIX 1: async mutex              // FIX 2: short sync critical section
let g = tok_mutex.lock().await;    let val = { let g = std_mutex.lock().unwrap(); g.clone() };
async_op().await; drop(g);         async_op_with(val).await;
```

### Blocking the Runtime with CPU-Heavy Work

```rust
// BAD: starves other tasks
async fn hash_password(pw: String) -> String { argon2::hash(pw) }

// FIX: offload to blocking thread pool
async fn hash_password(pw: String) -> String {
    tokio::task::spawn_blocking(move || argon2::hash(pw)).await.unwrap()
}
```

### Creating Too Many Tasks

Spawning millions of tasks without backpressure exhausts memory. Fix: use a `Semaphore`:

```rust
let sem = Arc::new(Semaphore::new(1000));
for item in huge_list {
    let permit = sem.clone().acquire_owned().await.unwrap();
    tokio::spawn(async move { process(item).await; drop(permit); });
}
```

### Forgetting to `.await`

A future does nothing until polled. This compiles but performs no work:

```rust
async fn setup_db() {
    db.run_migrations();        // returns a Future — does nothing!
    db.run_migrations().await;  // actually runs
}
```

Enable `#[warn(unused_must_use)]` (on by default) and watch for `Future` return types.

### Mixing Runtimes

Nesting a Tokio runtime inside another panics ("Cannot start a runtime from within a runtime"). If you need async from sync code inside an async context:

```rust
use tokio::runtime::Handle;
async fn call_sync_that_needs_async() {
    let handle = Handle::current();
    tokio::task::spawn_blocking(move || {
        handle.block_on(async { async_work().await })
    }).await.unwrap();
}
```
