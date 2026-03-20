---
name: rust-async-patterns
description: >
  Rust async/await patterns — Tokio runtime, async functions, futures,
  streams, channels, concurrent tasks, cancellation, and async testing.
  Triggers: "rust async", "rust await", "tokio", "rust future", "rust stream",
  "rust channel", "rust spawn", "rust concurrent", "rust async trait".
  NOT for: Error handling patterns (use rust-error-handling).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Rust Async Patterns

## Tokio Runtime Setup

```rust
// main.rs
#[tokio::main]
async fn main() -> anyhow::Result<()> {
    // Initialize tracing
    tracing_subscriber::fmt()
        .with_env_filter("myapp=debug,tower_http=info")
        .json()
        .init();

    let config = Config::from_env()?;

    // Build app state
    let pool = PgPoolOptions::new()
        .max_connections(config.db_pool_size)
        .connect(&config.database_url)
        .await?;

    let state = AppState {
        db: pool,
        redis: redis::Client::open(config.redis_url)?,
    };

    // Build router
    let app = Router::new()
        .route("/health", get(health))
        .route("/api/users", get(list_users).post(create_user))
        .route("/api/users/:id", get(get_user).put(update_user))
        .layer(TraceLayer::new_for_http())
        .with_state(state);

    // Graceful shutdown
    let listener = TcpListener::bind(&config.bind_addr).await?;
    tracing::info!("listening on {}", config.bind_addr);

    axum::serve(listener, app)
        .with_graceful_shutdown(shutdown_signal())
        .await?;

    Ok(())
}

async fn shutdown_signal() {
    let ctrl_c = async {
        tokio::signal::ctrl_c().await.expect("install ctrl-c handler");
    };

    #[cfg(unix)]
    let terminate = async {
        tokio::signal::unix::signal(tokio::signal::unix::SignalKind::terminate())
            .expect("install SIGTERM handler")
            .recv()
            .await;
    };

    #[cfg(not(unix))]
    let terminate = std::future::pending::<()>();

    tokio::select! {
        _ = ctrl_c => tracing::info!("ctrl-c received"),
        _ = terminate => tracing::info!("SIGTERM received"),
    }
}
```

## Spawning Tasks

```rust
use tokio::task;

// Spawn a background task — runs independently
let handle = task::spawn(async move {
    process_queue().await
});

// Spawn and await result
let result = task::spawn(async move {
    expensive_computation().await
}).await?; // JoinError if task panics

// Spawn blocking — for CPU-bound or sync code
let hash = task::spawn_blocking(move || {
    argon2::hash_encoded(password.as_bytes(), &salt, &config)
}).await??;

// spawn_blocking is essential — sync code in async blocks the executor
// BAD:  let hash = argon2::hash(...);  // blocks the tokio runtime
// GOOD: let hash = spawn_blocking(|| argon2::hash(...)).await?;

// Spawn on a specific runtime
let rt = tokio::runtime::Builder::new_multi_thread()
    .worker_threads(4)
    .enable_all()
    .build()?;

rt.spawn(async { background_work().await });
```

## Concurrent Operations

```rust
use tokio::try_join;
use futures::future::join_all;

// Run two futures concurrently — fail fast
async fn fetch_user_data(user_id: &str) -> Result<UserProfile> {
    let (user, orders) = try_join!(
        get_user(user_id),
        get_orders(user_id),
    )?; // Returns Err as soon as either fails

    Ok(UserProfile { user, orders })
}

// Run N futures concurrently
async fn fetch_all_users(ids: &[String]) -> Vec<Result<User>> {
    let futures: Vec<_> = ids.iter()
        .map(|id| get_user(id))
        .collect();

    join_all(futures).await
}

// Concurrent with concurrency limit
use futures::stream::{self, StreamExt};

async fn process_batch(items: Vec<Item>) -> Vec<Result<Output>> {
    stream::iter(items)
        .map(|item| async move {
            process_item(item).await
        })
        .buffer_unordered(10) // Max 10 concurrent
        .collect()
        .await
}

// Select — race futures, take first to complete
use tokio::select;
use tokio::time::{sleep, Duration};

async fn fetch_with_timeout(url: &str) -> Result<Response> {
    select! {
        result = reqwest::get(url) => {
            result.map_err(Into::into)
        }
        _ = sleep(Duration::from_secs(5)) => {
            Err(anyhow::anyhow!("request timed out"))
        }
    }
}

// Select with cancellation
async fn interruptible_work(cancel: CancellationToken) -> Result<()> {
    loop {
        select! {
            _ = cancel.cancelled() => {
                tracing::info!("work cancelled");
                return Ok(());
            }
            result = do_next_chunk() => {
                result?;
            }
        }
    }
}
```

## Channels

```rust
use tokio::sync::{mpsc, oneshot, broadcast, watch};

// mpsc — multiple producer, single consumer
async fn event_processor() {
    let (tx, mut rx) = mpsc::channel::<Event>(100);

    // Producer
    let tx2 = tx.clone();
    tokio::spawn(async move {
        tx2.send(Event::UserCreated { id: "1".into() }).await.ok();
    });

    // Consumer
    while let Some(event) = rx.recv().await {
        match event {
            Event::UserCreated { id } => handle_user_created(&id).await,
            Event::OrderPlaced { order } => handle_order(&order).await,
        }
    }
}

// oneshot — single value, single use (request/response)
async fn query_service(request: Request) -> Result<Response> {
    let (tx, rx) = oneshot::channel();

    // Send request with response channel
    service_tx.send((request, tx)).await?;

    // Wait for response
    let response = rx.await?;
    Ok(response)
}

// broadcast — multiple consumers, each gets every message
async fn event_bus() {
    let (tx, _) = broadcast::channel::<String>(100);

    let mut rx1 = tx.subscribe();
    let mut rx2 = tx.subscribe();

    tokio::spawn(async move {
        while let Ok(msg) = rx1.recv().await {
            println!("listener 1: {msg}");
        }
    });

    tokio::spawn(async move {
        while let Ok(msg) = rx2.recv().await {
            println!("listener 2: {msg}");
        }
    });

    tx.send("hello".into()).ok();
}

// watch — latest value, multiple readers
async fn config_watcher() {
    let (tx, rx) = watch::channel(Config::default());

    // Reader — always gets the latest value
    let mut rx = rx.clone();
    tokio::spawn(async move {
        while rx.changed().await.is_ok() {
            let config = rx.borrow();
            println!("config updated: {:?}", *config);
        }
    });

    // Writer — update config
    tx.send(Config::load().await?)?;
}
```

## Async Streams

```rust
use futures::stream::{self, Stream, StreamExt, TryStreamExt};
use tokio_stream::wrappers::ReceiverStream;

// Create a stream from an async function
fn user_stream(db: &PgPool) -> impl Stream<Item = Result<User>> + '_ {
    sqlx::query_as::<_, User>("SELECT * FROM users ORDER BY created_at")
        .fetch(db)
        .map_err(Into::into)
}

// Process a stream with transformations
async fn process_stream(db: &PgPool) -> Result<()> {
    user_stream(db)
        .try_filter(|user| future::ready(user.is_active))
        .try_chunks(100) // Batch into groups of 100
        .try_for_each(|batch| async move {
            update_batch(&batch).await
        })
        .await?;
    Ok(())
}

// Convert channel to stream
async fn events_as_stream(rx: mpsc::Receiver<Event>) -> impl Stream<Item = Event> {
    ReceiverStream::new(rx)
}

// SSE streaming (axum)
use axum::response::sse::{Event as SseEvent, Sse};
use std::convert::Infallible;

async fn sse_handler(
    State(state): State<AppState>,
) -> Sse<impl Stream<Item = Result<SseEvent, Infallible>>> {
    let stream = stream::unfold(state, |state| async move {
        tokio::time::sleep(Duration::from_secs(1)).await;
        let data = get_updates(&state).await;
        let event = SseEvent::default()
            .data(serde_json::to_string(&data).unwrap());
        Some((Ok(event), state))
    });

    Sse::new(stream)
}
```

## Shared State

```rust
use std::sync::Arc;
use tokio::sync::{Mutex, RwLock};

// Arc<Mutex<T>> — shared mutable state
#[derive(Clone)]
struct AppState {
    db: PgPool,
    cache: Arc<Mutex<HashMap<String, CachedValue>>>,
    config: Arc<RwLock<Config>>,
}

// Mutex — exclusive access
async fn get_cached(state: &AppState, key: &str) -> Option<String> {
    let cache = state.cache.lock().await;
    cache.get(key).map(|v| v.value.clone())
}

async fn set_cached(state: &AppState, key: String, value: String) {
    let mut cache = state.cache.lock().await;
    cache.insert(key, CachedValue {
        value,
        expires_at: Instant::now() + Duration::from_secs(300),
    });
}

// RwLock — multiple readers OR one writer
async fn get_config(state: &AppState) -> Config {
    state.config.read().await.clone()
}

async fn update_config(state: &AppState, new_config: Config) {
    let mut config = state.config.write().await;
    *config = new_config;
}

// DashMap — concurrent hashmap (no async needed)
use dashmap::DashMap;

#[derive(Clone)]
struct FastCache {
    inner: Arc<DashMap<String, CachedValue>>,
}

impl FastCache {
    fn get(&self, key: &str) -> Option<String> {
        self.inner.get(key).map(|v| v.value.clone())
    }

    fn set(&self, key: String, value: String) {
        self.inner.insert(key, CachedValue { value, expires_at: Instant::now() + Duration::from_secs(300) });
    }
}
```

## Cancellation

```rust
use tokio_util::sync::CancellationToken;

async fn run_workers(cancel: CancellationToken) -> Result<()> {
    let mut handles = Vec::new();

    for i in 0..4 {
        let cancel = cancel.clone();
        handles.push(tokio::spawn(async move {
            loop {
                tokio::select! {
                    _ = cancel.cancelled() => {
                        tracing::info!("worker {i} shutting down");
                        break;
                    }
                    _ = do_work(i) => {}
                }
            }
        }));
    }

    // Wait for shutdown signal
    shutdown_signal().await;
    cancel.cancel(); // Signal all workers

    // Wait for all workers to finish
    for handle in handles {
        handle.await?;
    }

    Ok(())
}

// Timeout pattern
use tokio::time::timeout;

async fn with_retry<T, F, Fut>(
    max_retries: u32,
    timeout_ms: u64,
    f: F,
) -> Result<T>
where
    F: Fn() -> Fut,
    Fut: std::future::Future<Output = Result<T>>,
{
    let mut last_error = None;

    for attempt in 0..max_retries {
        match timeout(Duration::from_millis(timeout_ms), f()).await {
            Ok(Ok(result)) => return Ok(result),
            Ok(Err(e)) => {
                tracing::warn!("attempt {attempt} failed: {e}");
                last_error = Some(e);
            }
            Err(_) => {
                tracing::warn!("attempt {attempt} timed out");
                last_error = Some(anyhow::anyhow!("timeout"));
            }
        }

        // Exponential backoff
        let delay = Duration::from_millis(100 * 2u64.pow(attempt));
        tokio::time::sleep(delay).await;
    }

    Err(last_error.unwrap_or_else(|| anyhow::anyhow!("all retries failed")))
}
```

## Testing Async Code

```rust
#[cfg(test)]
mod tests {
    use super::*;

    // Basic async test
    #[tokio::test]
    async fn test_fetch_user() {
        let pool = setup_test_db().await;
        let user = create_test_user(&pool).await;

        let result = get_user(&pool, &user.id).await;
        assert!(result.is_ok());
        assert_eq!(result.unwrap().email, user.email);
    }

    // Test with timeout
    #[tokio::test]
    async fn test_slow_operation_timeout() {
        let result = timeout(
            Duration::from_millis(100),
            slow_operation(),
        ).await;

        assert!(result.is_err()); // Should timeout
    }

    // Test concurrent operations
    #[tokio::test]
    async fn test_concurrent_writes() {
        let counter = Arc::new(AtomicU64::new(0));
        let mut handles = Vec::new();

        for _ in 0..100 {
            let counter = counter.clone();
            handles.push(tokio::spawn(async move {
                counter.fetch_add(1, Ordering::SeqCst);
            }));
        }

        for handle in handles {
            handle.await.unwrap();
        }

        assert_eq!(counter.load(Ordering::SeqCst), 100);
    }

    // Test channels
    #[tokio::test]
    async fn test_event_processing() {
        let (tx, rx) = mpsc::channel(10);

        let processor = tokio::spawn(async move {
            let mut count = 0;
            let mut rx = rx;
            while let Some(_event) = rx.recv().await {
                count += 1;
            }
            count
        });

        tx.send(Event::Test).await.unwrap();
        tx.send(Event::Test).await.unwrap();
        drop(tx); // Close channel

        let count = processor.await.unwrap();
        assert_eq!(count, 2);
    }

    // Mock with test fixture
    async fn setup_test_db() -> PgPool {
        let url = std::env::var("TEST_DATABASE_URL")
            .unwrap_or_else(|_| "postgres://localhost/test".into());

        let pool = PgPool::connect(&url).await.unwrap();
        sqlx::migrate!().run(&pool).await.unwrap();
        pool
    }
}
```

## Gotchas

1. **Holding a `MutexGuard` across `.await`** — `std::sync::MutexGuard` is not `Send`. Holding it across an `.await` point fails to compile or causes deadlocks. Use `tokio::sync::Mutex` for async code, or limit the guard's scope with a block: `{ let val = mutex.lock(); val.clone() }`.

2. **`tokio::spawn` requires `'static`** — Spawned tasks can't borrow local data. You must move owned data into the closure: `let id = id.clone(); tokio::spawn(async move { use(id) })`. Arc for shared ownership.

3. **Blocking in async context** — Sync operations (file I/O, CPU-heavy work, `thread::sleep`) block the tokio worker thread. Use `spawn_blocking` for sync code, `tokio::fs` for file I/O, and `tokio::time::sleep` instead of `thread::sleep`.

4. **`select!` cancels losing branches** — When one branch completes in `select!`, the other futures are DROPPED (cancelled). If a dropped future was mid-database-write, that write is lost. Use `tokio::pin!` and loop with `&mut future` for futures that must complete.

5. **Channel backpressure** — `mpsc::channel(100)` blocks the sender when the buffer is full. If the consumer is slow, the producer blocks too. Size the buffer based on expected throughput, or use `try_send` to handle backpressure explicitly.

6. **`Arc<Mutex<HashMap>>` vs `DashMap`** — `Arc<Mutex<HashMap>>` locks the entire map for any read or write. `DashMap` uses sharded locking for much better concurrent performance. Use DashMap for high-contention caches. Use Mutex only when you need transactional access to multiple keys.
