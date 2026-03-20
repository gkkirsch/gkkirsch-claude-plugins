---
name: rust-error-handling
description: >
  Rust error handling patterns — Result/Option, custom error types, thiserror,
  anyhow, error propagation, and idiomatic error design for libraries and applications.
  Triggers: "rust error", "rust result", "rust option", "thiserror", "anyhow",
  "rust error handling", "rust unwrap", "rust expect", "rust ?".
  NOT for: Async/concurrent Rust patterns (use rust-async-patterns).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Rust Error Handling

## The Basics: Result and Option

```rust
// Result<T, E> — operation that can fail
fn parse_config(path: &str) -> Result<Config, ConfigError> {
    let content = std::fs::read_to_string(path)
        .map_err(|e| ConfigError::IoError(e))?;

    let config: Config = serde_json::from_str(&content)
        .map_err(|e| ConfigError::ParseError(e))?;

    if config.port == 0 {
        return Err(ConfigError::Validation("port must be non-zero".into()));
    }

    Ok(config)
}

// Option<T> — value that might not exist
fn find_user(users: &[User], id: &str) -> Option<&User> {
    users.iter().find(|u| u.id == id)
}

// Converting between Option and Result
let user = find_user(&users, "123")
    .ok_or_else(|| AppError::NotFound("user 123".into()))?;

// Converting Result to Option (discards error)
let maybe_config = parse_config("config.json").ok();
```

## The ? Operator

```rust
// ? propagates errors up the call stack
// The error type must implement From<SourceError> for TargetError

fn process_order(order_id: &str) -> Result<Receipt, AppError> {
    let order = db::get_order(order_id)?;          // DbError -> AppError
    let payment = stripe::charge(&order)?;          // StripeError -> AppError
    let receipt = generate_receipt(&order, &payment)?; // ReceiptError -> AppError
    email::send_receipt(&order.email, &receipt)?;   // EmailError -> AppError
    Ok(receipt)
}

// ? also works with Option in functions returning Option
fn get_street_name(user: &User) -> Option<&str> {
    let address = user.address.as_ref()?;
    let street = address.street.as_ref()?;
    Some(street.as_str())
}
```

## Custom Error Types with thiserror

```rust
// For libraries — structured, type-safe errors
use thiserror::Error;

#[derive(Error, Debug)]
pub enum ApiError {
    #[error("not found: {resource} with id {id}")]
    NotFound { resource: String, id: String },

    #[error("validation failed: {0}")]
    Validation(String),

    #[error("unauthorized: {0}")]
    Unauthorized(String),

    #[error("forbidden")]
    Forbidden,

    #[error("conflict: {0}")]
    Conflict(String),

    #[error("rate limited, retry after {retry_after}s")]
    RateLimited { retry_after: u64 },

    #[error("database error")]
    Database(#[from] sqlx::Error),

    #[error("serialization error")]
    Serialization(#[from] serde_json::Error),

    #[error("io error")]
    Io(#[from] std::io::Error),

    #[error("internal error: {0}")]
    Internal(String),
}

impl ApiError {
    pub fn status_code(&self) -> u16 {
        match self {
            Self::NotFound { .. } => 404,
            Self::Validation(_) => 422,
            Self::Unauthorized(_) => 401,
            Self::Forbidden => 403,
            Self::Conflict(_) => 409,
            Self::RateLimited { .. } => 429,
            Self::Database(_) | Self::Internal(_) => 500,
            Self::Serialization(_) | Self::Io(_) => 500,
        }
    }

    pub fn is_retryable(&self) -> bool {
        matches!(self, Self::RateLimited { .. } | Self::Database(_))
    }
}

// Axum integration
impl axum::response::IntoResponse for ApiError {
    fn into_response(self) -> axum::response::Response {
        let status = axum::http::StatusCode::from_u16(self.status_code())
            .unwrap_or(axum::http::StatusCode::INTERNAL_SERVER_ERROR);

        let body = serde_json::json!({
            "error": self.to_string(),
            "code": self.status_code(),
        });

        (status, axum::Json(body)).into_response()
    }
}
```

## Application Errors with anyhow

```rust
// For applications — flexible, context-rich errors
use anyhow::{Context, Result, bail, ensure};

fn run_migration(db_url: &str) -> Result<()> {
    let pool = PgPool::connect(db_url)
        .await
        .context("failed to connect to database")?;

    let pending = get_pending_migrations(&pool)
        .await
        .context("failed to check pending migrations")?;

    ensure!(!pending.is_empty(), "no pending migrations found");

    for migration in &pending {
        apply_migration(&pool, migration)
            .await
            .with_context(|| format!("failed to apply migration: {}", migration.name))?;

        println!("Applied: {}", migration.name);
    }

    Ok(())
}

// bail! for early returns with formatted error
fn validate_config(config: &Config) -> Result<()> {
    if config.port == 0 {
        bail!("port must be non-zero");
    }
    if config.database_url.is_empty() {
        bail!("DATABASE_URL is required");
    }
    if config.workers > 64 {
        bail!("workers must be <= 64, got {}", config.workers);
    }
    Ok(())
}

// Downcast anyhow errors to concrete types
fn handle_error(err: &anyhow::Error) {
    if let Some(api_err) = err.downcast_ref::<ApiError>() {
        match api_err {
            ApiError::NotFound { resource, id } => {
                println!("{resource} {id} not found");
            }
            _ => println!("API error: {api_err}"),
        }
    } else {
        println!("Unknown error: {err:?}");
    }
}
```

## Error Composition

```rust
// Nested errors with context chain
use thiserror::Error;

#[derive(Error, Debug)]
pub enum OrderError {
    #[error("failed to create order")]
    Creation(#[source] Box<dyn std::error::Error + Send + Sync>),

    #[error("payment failed for order {order_id}")]
    Payment {
        order_id: String,
        #[source]
        source: PaymentError,
    },

    #[error("fulfillment failed")]
    Fulfillment(#[from] FulfillmentError),
}

#[derive(Error, Debug)]
pub enum PaymentError {
    #[error("card declined: {reason}")]
    Declined { reason: String },

    #[error("payment gateway timeout")]
    Timeout,

    #[error("insufficient funds: need {required}, have {available}")]
    InsufficientFunds { required: f64, available: f64 },
}

// Walking the error chain
fn log_error_chain(err: &dyn std::error::Error) {
    let mut current = Some(err);
    let mut depth = 0;
    while let Some(e) = current {
        if depth == 0 {
            tracing::error!("Error: {e}");
        } else {
            tracing::error!("  Caused by: {e}");
        }
        current = e.source();
        depth += 1;
    }
}
```

## Result Combinators

```rust
// map — transform the Ok value
let length: Result<usize, _> = read_file("data.txt").map(|s| s.len());

// and_then — chain fallible operations (flatmap)
let user: Result<User, _> = get_user_id(token)
    .and_then(|id| find_user(&db, &id));

// or_else — provide fallback on error
let config = load_config("prod.toml")
    .or_else(|_| load_config("default.toml"))?;

// unwrap_or_default — safe default on error
let count: i32 = parse_count(input).unwrap_or_default(); // 0 on error

// map_err — transform the error type
let file = File::open(path)
    .map_err(|e| AppError::Config(format!("can't open {path}: {e}")))?;

// Collecting Results — fail fast or collect all
let results: Result<Vec<User>, _> = ids
    .iter()
    .map(|id| find_user(id))
    .collect();  // Stops at first Err

// Partition successes and failures
let (successes, failures): (Vec<_>, Vec<_>) = ids
    .iter()
    .map(|id| find_user(id))
    .partition(Result::is_ok);
let users: Vec<User> = successes.into_iter().map(Result::unwrap).collect();
let errors: Vec<_> = failures.into_iter().map(Result::unwrap_err).collect();
```

## Pattern: Typed Error Responses

```rust
// Type-safe API error responses
#[derive(Debug, Serialize)]
#[serde(tag = "type")]
pub enum ErrorResponse {
    #[serde(rename = "validation_error")]
    Validation { fields: Vec<FieldError> },

    #[serde(rename = "not_found")]
    NotFound { resource: String, id: String },

    #[serde(rename = "auth_error")]
    Auth { message: String },

    #[serde(rename = "internal_error")]
    Internal { request_id: String },
}

#[derive(Debug, Serialize)]
pub struct FieldError {
    pub field: String,
    pub message: String,
    pub code: String,
}

impl From<ApiError> for ErrorResponse {
    fn from(err: ApiError) -> Self {
        match err {
            ApiError::NotFound { resource, id } => {
                ErrorResponse::NotFound { resource, id }
            }
            ApiError::Validation(msg) => {
                ErrorResponse::Validation {
                    fields: vec![FieldError {
                        field: "unknown".into(),
                        message: msg,
                        code: "invalid".into(),
                    }],
                }
            }
            ApiError::Unauthorized(msg) => {
                ErrorResponse::Auth { message: msg }
            }
            _ => ErrorResponse::Internal {
                request_id: "unknown".into(),
            },
        }
    }
}
```

## Testing Errors

```rust
#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_parse_config_missing_file() {
        let result = parse_config("/nonexistent/path");
        assert!(result.is_err());
        assert!(matches!(
            result.unwrap_err(),
            ConfigError::IoError(_)
        ));
    }

    #[test]
    fn test_parse_config_invalid_json() {
        // Write temp file with invalid content
        let dir = tempdir().unwrap();
        let path = dir.path().join("config.json");
        std::fs::write(&path, "not json").unwrap();

        let result = parse_config(path.to_str().unwrap());
        assert!(matches!(
            result.unwrap_err(),
            ConfigError::ParseError(_)
        ));
    }

    #[test]
    fn test_error_display() {
        let err = ApiError::NotFound {
            resource: "user".into(),
            id: "123".into(),
        };
        assert_eq!(err.to_string(), "not found: user with id 123");
        assert_eq!(err.status_code(), 404);
    }

    #[test]
    fn test_error_is_retryable() {
        assert!(ApiError::RateLimited { retry_after: 60 }.is_retryable());
        assert!(!ApiError::Forbidden.is_retryable());
    }

    // Test error chain
    #[test]
    fn test_error_source_chain() {
        let payment_err = PaymentError::Declined {
            reason: "expired card".into(),
        };
        let order_err = OrderError::Payment {
            order_id: "ord-1".into(),
            source: payment_err,
        };

        // Check the chain
        assert!(order_err.source().is_some());
        let source = order_err.source().unwrap();
        assert!(source.to_string().contains("expired card"));
    }
}
```

## Gotchas

1. **`unwrap()` in production code** — `unwrap()` panics on `None`/`Err`. Use `expect("reason")` for truly impossible cases (with a message explaining why), or propagate with `?`. Reserve `unwrap()` for tests only.

2. **thiserror vs anyhow** — Use thiserror for libraries (callers need to match on error variants). Use anyhow for applications (you just need to report errors with context). Don't use anyhow in library public APIs.

3. **`#[from]` creates implicit conversions** — `#[from] sqlx::Error` means any `sqlx::Error` auto-converts to your error type via `?`. This hides context. Consider using `map_err` with `.context()` for important conversions where you want to add information.

4. **`Box<dyn Error>` loses type info** — Once you box an error, you can only downcast it. If callers need to match on specific variants, keep the concrete type. Use `Box<dyn Error>` only at application boundaries.

5. **Panics are not errors** — `panic!` unwinds the stack and is NOT meant for error handling. It's for programmer bugs (array out of bounds, impossible states). Use `Result` for anything that can legitimately fail. Don't catch panics with `catch_unwind` as a control flow mechanism.

6. **Missing `Send + Sync` bounds on error types** — If your error type will cross thread boundaries (async, channels, thread pools), it must be `Send + Sync`. `#[derive(Error)]` with thiserror handles this, but custom `std::error::Error` impls may not. Test by using your error in an async function.
