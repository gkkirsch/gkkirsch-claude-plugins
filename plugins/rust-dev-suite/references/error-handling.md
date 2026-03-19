# Error Handling in Rust — Reference

---

## 1. The Error Handling Spectrum

**When to `panic!`**: unrecoverable bugs, violated invariants, setup failures, tests/prototypes.

**When to use `Result<T, E>`**: recoverable errors, I/O, parsing, validation — anything expected to fail.

**When to use `Option<T>`**: absence of a value (not an error), lookups like `HashMap::get`.

| Context            | Recommended approach                                  |
| ------------------ | ----------------------------------------------------- |
| Library crate      | Define typed errors with `thiserror`, return `Result` |
| Application binary | Use `anyhow::Result` for convenience                  |
| Boundary code      | Convert between error types explicitly                |
| FFI boundary       | Catch panics with `catch_unwind`, return error codes  |
| Tests              | `unwrap()`, `expect()`, or `-> Result<()>` returns    |

Libraries should never panic on bad input. Applications may panic at the outermost
boundary when setup fails irrecoverably.

---

## 2. Result & Option Combinators

### `Result<T, E>` Combinators

| Combinator          | Signature (simplified)                        | Purpose                    |
| ------------------- | --------------------------------------------- | -------------------------- |
| `map`               | `(T->U) -> Result<U,E>`                      | Transform success value    |
| `map_err`           | `(E->F) -> Result<T,F>`                      | Transform error value      |
| `and_then`          | `(T->Result<U,E>) -> Result<U,E>`            | Chain fallible operations  |
| `or_else`           | `(E->Result<T,F>) -> Result<T,F>`            | Recover from error         |
| `unwrap_or`         | `T -> T`                                      | Default on error           |
| `unwrap_or_else`    | `(E->T) -> T`                                 | Compute default from error |
| `unwrap_or_default` | `-> T` (where `T: Default`)                   | `T::default()` on error   |
| `ok` / `err`        | `-> Option<T>` / `-> Option<E>`               | Convert to Option          |
| `transpose`         | `Option<Result<T,E>> -> Result<Option<T>,E>`  | Swap nesting order         |

### `Option<T>` Key Combinators

`map`, `and_then`, `or_else`, `unwrap_or[_else]`, `ok_or[_else]`, `transpose`.

### Examples

```rust
fn find_user_email(db: &Database, user_id: u64) -> Result<String, AppError> {
    db.find_user(user_id)
        .map_err(AppError::Database)?
        .ok_or(AppError::NotFound("user"))?
        .email
        .ok_or(AppError::NotFound("email"))
}
```

```rust
fn read_config(path: &str) -> Result<Config, AppError> {
    let contents = std::fs::read_to_string(path).map_err(AppError::Io)?;
    toml::from_str(&contents).map_err(AppError::Parse)
}
```

```rust
// transpose: Option<Result<T, E>> -> Result<Option<T>, E>
fn maybe_parse(input: Option<&str>) -> Result<Option<i64>, ParseIntError> {
    input.map(|s| s.parse::<i64>()).transpose()
}
```

```rust
// or_else: fallback strategies
fn load_setting(key: &str) -> Result<String, ConfigError> {
    load_from_env(key)
        .or_else(|_| load_from_file(key))
        .or_else(|_| load_from_defaults(key))
}
```

---

## 3. The `?` Operator

### How `?` Desugars

```rust
let value = some_operation()?;
// Desugars to:
let value = match some_operation() {
    Ok(v) => v,
    Err(e) => return Err(From::from(e)),
};
```

The `From::from(e)` call converts the error type. Your return type must implement
`From` for each source error, otherwise `?` will not compile.

### `?` with Option

```rust
fn first_even(numbers: &[i32]) -> Option<i32> {
    let first = numbers.first()?;  // returns None if empty
    if first % 2 == 0 { Some(*first) } else { None }
}
```

### Manual `From` Implementation

```rust
impl From<std::io::Error> for AppError {
    fn from(err: std::io::Error) -> Self { AppError::Io(err) }
}

fn read_data() -> Result<Vec<u8>, AppError> {
    let data = std::fs::read("data.bin")?;  // io::Error -> AppError via From
    Ok(data)
}
```

### `?` in `main()`

```rust
fn main() -> ExitCode {
    if let Err(e) = run() {
        eprintln!("Error: {e:#}");
        return ExitCode::FAILURE;
    }
    ExitCode::SUCCESS
}

fn run() -> anyhow::Result<()> {
    let config = load_config()?;
    start_server(config)?;
    Ok(())
}
```

---

## 4. `thiserror` — Library Error Types

Generates `Display`, `Error`, and `From` impls via derive macros. Use in **library
crates** where callers match on specific error variants.

### Full Example

```rust
use thiserror::Error;

#[derive(Debug, Error)]
pub enum StorageError {
    #[error("connection failed: {0}")]
    Connection(#[from] std::io::Error),

    #[error("query failed: {0}")]
    Query(#[from] sqlx::Error),

    #[error("record not found: {entity} with id {id}")]
    NotFound { entity: &'static str, id: String },

    #[error("constraint violation: {0}")]
    Constraint(String),

    #[error(transparent)]
    Other(#[from] anyhow::Error),
}
```

### Attribute Reference

| Attribute              | Effect                                                 |
| ---------------------- | ------------------------------------------------------ |
| `#[error("...")]`      | Generates `Display` with the format string             |
| `#[from]`              | Generates `From<SourceError>` impl for this variant    |
| `#[source]`            | Marks field as source error (no `From` impl generated) |
| `#[error(transparent)]`| Delegates both `Display` and `source()` to inner error |

### `#[from]` — Automatic Conversion

```rust
#[derive(Debug, Error)]
pub enum ApiError {
    #[error("serialization failed")]
    Serialization(#[from] serde_json::Error),
    #[error("HTTP request failed")]
    Http(#[from] reqwest::Error),
}
```

### `#[source]` — Chain Without `From`

```rust
#[derive(Debug, Error)]
#[error("failed to initialize module {name}")]
pub struct InitError {
    pub name: String,
    #[source]
    pub cause: std::io::Error,
}
```

### Struct vs Enum Errors

**Enum**: callers distinguish failure modes. **Struct**: single-cause with context.

```rust
#[derive(Debug, Error)]
#[error("rate limit exceeded for {endpoint}: retry after {retry_after_secs}s")]
pub struct RateLimitError {
    pub endpoint: String,
    pub retry_after_secs: u64,
}

#[derive(Debug, Error)]
pub enum AuthError {
    #[error("invalid credentials")]
    InvalidCredentials,
    #[error("token expired at {0}")]
    TokenExpired(chrono::DateTime<chrono::Utc>),
    #[error("insufficient permissions: requires {required}")]
    Forbidden { required: String },
}
```

Define focused errors per module, convert at boundaries with `#[from]`.

---

## 5. `anyhow` — Application Error Handling

Single flexible error type for **application code**. Propagate with context,
no need for callers to match variants.

### `.context()` and `.with_context()`

```rust
use anyhow::{Context, Result, bail, ensure};

fn read_config(path: &str) -> Result<Config> {
    let contents = std::fs::read_to_string(path)
        .context("failed to read configuration file")?;
    toml::from_str(&contents).context("failed to parse configuration")
}

async fn process_order(order_id: &str) -> Result<Receipt> {
    let order = db.find_order(order_id)
        .await
        .context("failed to fetch order from database")?
        .ok_or_else(|| anyhow::anyhow!("order {order_id} not found"))?;

    ensure!(order.status != Status::Cancelled, "order {order_id} is cancelled");
    if order.total <= 0 {
        bail!("order {order_id} has invalid total: {}", order.total);
    }

    let payment = payment_service
        .charge(order.total)
        .await
        .with_context(|| format!("failed to charge ${} for order {order_id}", order.total))?;
    Ok(Receipt { order, payment })
}
```

`bail!(...)` = `return Err(anyhow!(...))`. `ensure!(cond, ...)` = `if !cond { bail!(...) }`.

### Downcasting

```rust
fn handle_error(err: &anyhow::Error) {
    if let Some(db_err) = err.downcast_ref::<StorageError>() {
        match db_err {
            StorageError::NotFound { entity, id } => log::warn!("Missing {entity}/{id}"),
            _ => log::error!("Storage failure: {db_err}"),
        }
    }
}
```

### `anyhow` vs `thiserror`

| Criteria           | `thiserror`                  | `anyhow`                       |
| ------------------ | ---------------------------- | ------------------------------ |
| Crate type         | Library                      | Application                    |
| Match on variants  | Yes                          | No (unless downcasting)        |
| Context chaining   | Manual (new variant)         | Built-in `.context()`          |
| Macros             | None                         | `bail!`, `ensure!`, `anyhow!`  |
| Performance        | Zero-cost (static dispatch)  | Allocation per error (dynamic) |

### Combining Both

```rust
// lib.rs — thiserror
#[derive(Debug, thiserror::Error)]
pub enum CoreError {
    #[error("invalid input: {0}")]
    InvalidInput(String),
}

// main.rs — anyhow
fn main() -> anyhow::Result<()> {
    my_lib::do_thing().context("failed during startup")?;
    Ok(())
}
```

---

## 6. Error Hierarchies in Multi-Crate Projects

Each crate owns its error type. Convert at crate boundaries via `From`.

```
my-workspace/
  crates/
    core/   -> CoreError
    db/     -> DbError (wraps CoreError, sqlx::Error)
    api/    -> ApiError (wraps DbError, auth errors)
  src/main.rs -> uses anyhow
```

### Boundary Conversion

```rust
// crates/db
#[derive(Debug, Error)]
pub enum DbError {
    #[error("SQL error: {0}")]
    Sql(#[from] sqlx::Error),
    #[error("migration failed: {0}")]
    Migration(String),
    #[error(transparent)]
    Core(#[from] core::CoreError),
}

// crates/api
#[derive(Debug, Error)]
pub enum ApiError {
    #[error("database error: {0}")]
    Database(#[from] db::DbError),
    #[error("unauthorized: {0}")]
    Auth(String),
}
```

Re-export errors from public API crates: `pub use crate::error::ApiError;`

**Wrapping** (`#[from] DbError`) preserves the chain for downcasting.
**Flattening** (`Database(String)`) loses type info but simplifies. Prefer
wrapping in libraries; flattening is acceptable at the application boundary.

---

## 7. Error Handling in Async Code

### `JoinError` from `tokio::spawn`

```rust
async fn run_task() -> Result<(), AppError> {
    let handle = tokio::spawn(async { heavy_computation().await });
    match handle.await {
        Ok(Ok(result)) => process(result),
        Ok(Err(app_err)) => return Err(app_err),
        Err(join_err) if join_err.is_panic() => {
            return Err(AppError::Internal("task panic".into()));
        }
        Err(_) => return Err(AppError::Internal("task cancelled".into())),
    }
    Ok(())
}
```

### Timeout Errors

```rust
async fn fetch_with_timeout(url: &str) -> Result<Response, AppError> {
    tokio::time::timeout(Duration::from_secs(30), reqwest::get(url))
        .await
        .map_err(|_| AppError::Timeout)?
        .map_err(AppError::Http)
}
```

### `select!`, Streams, and `try_join!`

```rust
async fn race() -> Result<Data, AppError> {
    tokio::select! {
        r = fetch_primary() => r.context("primary failed"),
        r = fetch_fallback() => r.context("fallback failed"),
    }
}

async fn process_stream(s: impl Stream<Item = Result<Item, DbError>>) -> Result<Vec<Item>> {
    s.collect::<Vec<_>>().await.into_iter()
        .collect::<Result<Vec<_>, _>>()
        .map_err(|e| anyhow::anyhow!("stream failed: {e}"))
}

async fn orchestrate() -> anyhow::Result<()> {
    let (user, stock) = tokio::try_join!(fetch_user(uid), check_inventory(iid))?;
    Ok(())
}
```

---

## 8. Panic Strategies

```toml
[profile.release]
panic = "abort"    # smaller binary, no catch_unwind support
```

| Strategy | Binary size | Can catch panics | Use case            |
| -------- | ----------- | ---------------- | ------------------- |
| `unwind` | Larger      | Yes              | Libraries, servers  |
| `abort`  | Smaller     | No               | Embedded, CLI tools |

### `catch_unwind` — FFI and Thread Pool Boundaries Only

```rust
fn call_plugin(plugin: &dyn Plugin) -> Result<Output, PluginError> {
    match std::panic::catch_unwind(std::panic::AssertUnwindSafe(|| plugin.execute())) {
        Ok(result) => result,
        Err(_) => Err(PluginError::Panicked),
    }
}
```

Libraries that panic on invalid input force callers into `catch_unwind` or crash.
Always return `Result` instead.

---

## 9. Production Patterns

### Logging with `tracing`

```rust
#[instrument(skip(db))]
async fn get_user(db: &Pool, id: u64) -> Result<User, ApiError> {
    db.find_user(id).await.map_err(|e| {
        error!(user_id = id, error = %e, "database lookup failed");
        ApiError::Internal
    })
}
```

### Error Response Mapping (Axum)

```rust
impl IntoResponse for ApiError {
    fn into_response(self) -> Response {
        let (status, msg) = match &self {
            ApiError::NotFound(m) => (StatusCode::NOT_FOUND, m.clone()),
            ApiError::Validation(m) => (StatusCode::BAD_REQUEST, m.clone()),
            ApiError::Auth(_) => (StatusCode::UNAUTHORIZED, "unauthorized".into()),
            ApiError::Internal => (StatusCode::INTERNAL_SERVER_ERROR, "internal error".into()),
        };
        (status, Json(json!({ "error": msg }))).into_response()
    }
}
```

### Retryable vs Non-Retryable

```rust
impl ApiError {
    pub fn is_retryable(&self) -> bool {
        matches!(self, ApiError::Database(DbError::Sql(_)) | ApiError::Timeout)
    }
}

async fn with_retry<F, Fut, T>(attempts: u32, f: F) -> Result<T, ApiError>
where F: Fn() -> Fut, Fut: Future<Output = Result<T, ApiError>> {
    let mut last_err = None;
    for i in 0..attempts {
        match f().await {
            Ok(v) => return Ok(v),
            Err(e) if e.is_retryable() && i + 1 < attempts => {
                tokio::time::sleep(Duration::from_millis(100 * 2u64.pow(i))).await;
                last_err = Some(e);
            }
            Err(e) => return Err(e),
        }
    }
    Err(last_err.unwrap())
}
```

### Error Codes for API Consumers

```rust
impl ApiError {
    pub fn error_code(&self) -> &'static str {
        match self {
            ApiError::NotFound(_) => "NOT_FOUND",
            ApiError::Validation(_) => "VALIDATION_ERROR",
            ApiError::Auth(_) => "UNAUTHORIZED",
            ApiError::Internal => "INTERNAL_ERROR",
        }
    }
}
```

```json
{ "error_code": "VALIDATION_ERROR", "message": "email field is required" }
```
