# Ownership & Borrowing Patterns — Rust Reference

---

## 1. The Three Ownership Rules

Rust's ownership system is enforced entirely at compile time with zero runtime cost.
Every design decision in a Rust API traces back to these three rules:

**Rule 1: Each value in Rust has exactly one owner.**

There is always a single variable binding that "owns" a piece of data. When you assign
a value to a new variable or pass it to a function, ownership transfers (moves) unless
the type implements `Copy`. This means you cannot accidentally create two parts of your
program that both think they are responsible for freeing the same allocation.

**Rule 2: There can only be one owner at a time.**

Ownership is exclusive. When a value moves, the previous owner becomes invalid. The
compiler enforces this statically — attempting to use a moved value is a compile-time
error, not a runtime crash. This eliminates use-after-free and double-free bugs entirely.

**Rule 3: When the owner goes out of scope, the value is dropped.**

Rust calls `Drop::drop` automatically at the end of the owning scope. This is
deterministic destruction — you know exactly when memory is freed, file handles are
closed, and locks are released. There is no garbage collector and no ref-counting
overhead unless you opt into it explicitly with `Rc<T>` or `Arc<T>`.

**Implications for API design:**

- Functions that need to store data long-term should take ownership (`T`, not `&T`).
- Functions that only read data should borrow it (`&T`).
- Functions that need to mutate data should take an exclusive borrow (`&mut T`).
- If callers sometimes have owned data and sometimes borrowed data, consider
  accepting `impl Into<T>` or `Cow<'a, T>` to let the caller choose.

---

## 2. Borrowing Patterns

### Shared References (`&T`)

A shared reference grants read-only access to data. Any number of shared references
can coexist simultaneously, but none of them may modify the underlying value.

```rust
fn print_length(s: &str) {
    println!("Length: {}", s.len());
}

fn main() {
    let name = String::from("Rust");
    let r1 = &name;
    let r2 = &name; // Multiple shared borrows: fine
    print_length(r1);
    print_length(r2);
    println!("{name}"); // Original owner still valid
}
```

Use shared references when:
- You need to read data without taking ownership.
- Multiple parts of your code need simultaneous access.
- You want to avoid cloning for performance.

### Exclusive References (`&mut T`)

An exclusive reference grants read-write access, but enforces that no other references
(shared or exclusive) exist for the same data at the same time.

```rust
fn push_greeting(buf: &mut String) {
    buf.push_str(", world!");
}

fn main() {
    let mut msg = String::from("Hello");
    push_greeting(&mut msg);
    println!("{msg}"); // "Hello, world!"
}
```

Use exclusive references when:
- You need to mutate borrowed data in place.
- You want to guarantee no other code can observe the data mid-mutation.

### Reborrowing

When you pass `&mut T` to a function expecting `&mut T`, the compiler automatically
creates a shorter-lived reborrow rather than moving the mutable reference. This lets
you keep using the original `&mut` after the function returns.

```rust
fn increment(val: &mut i32) {
    *val += 1;
}

fn main() {
    let mut x = 10;
    let r = &mut x;
    increment(r);   // Reborrow: r is temporarily "lent" to increment
    increment(r);   // r is usable again because the reborrow ended
    println!("{r}"); // 12
}
```

The compiler inserts `increment(&mut *r)` implicitly. The original reference `r` is
not moved — it is reborrowed for the duration of the call.

### Two-Phase Borrows and Non-Lexical Lifetimes (NLL)

Before NLL (Rust 2018+), borrows lasted until the end of their lexical scope. NLL
shortened borrow lifetimes to the last point where the reference is actually used.

```rust
fn main() {
    let mut v = vec![1, 2, 3];
    let first = &v[0]; // Shared borrow starts
    println!("{first}"); // Shared borrow's last use — ends here under NLL
    v.push(4); // Mutable borrow: OK because the shared borrow is no longer live
}
```

Two-phase borrows further refine this for method calls like `v.push(v.len())`.
The compiler splits the outer `&mut self` borrow into a reservation phase (where
shared borrows are still allowed) and an activation phase (where exclusive access
begins). This allows code that would otherwise be rejected:

```rust
fn main() {
    let mut v = vec![1, 2, 3];
    // The &mut self for push is reserved, then v.len() takes &self, then push activates
    v.push(v.len()); // Compiles thanks to two-phase borrows
}
```

---

## 3. Common Ownership Patterns

### Owned Fields

Structs own their data. This is the default and usually the correct choice.

```rust
struct User {
    name: String,     // Owns the string data
    email: String,    // Owns the string data
    login_count: u64,
}

impl User {
    fn new(name: String, email: String) -> Self {
        Self { name, email, login_count: 0 }
    }
}
```

The `User` struct has a clear, self-contained lifetime. When a `User` is dropped,
its `name` and `email` strings are freed automatically.

### Borrowed Parameters

Functions that only need to inspect data should borrow it.

```rust
fn is_valid_email(email: &str) -> bool {
    email.contains('@') && email.contains('.')
}

fn full_name(first: &str, last: &str) -> String {
    format!("{first} {last}") // Borrows input, returns owned output
}
```

Accept `&str` rather than `&String` — it is strictly more general because `String`
dereferences to `str`, and string literals are already `&str`.

### Returned References

Functions can return references into their inputs. The compiler infers that the
output lifetime is tied to the input.

```rust
fn first_word(s: &str) -> &str {
    let bytes = s.as_bytes();
    for (i, &byte) in bytes.iter().enumerate() {
        if byte == b' ' {
            return &s[..i];
        }
    }
    s
}

fn main() {
    let sentence = String::from("hello world");
    let word = first_word(&sentence); // word borrows from sentence
    println!("{word}"); // "hello"
}
```

### Temporary Ownership

Take ownership, transform the data, and return the new owned value. This is the
"consume and produce" pattern.

```rust
fn normalize_email(mut email: String) -> String {
    email.make_ascii_lowercase();
    email.trim().to_string()
}

fn main() {
    let raw = String::from("  Alice@Example.COM  ");
    let clean = normalize_email(raw);
    // raw is moved; only clean exists now
    println!("{clean}"); // "alice@example.com"
}
```

This pattern is efficient when the caller no longer needs the original data.

### Split Borrows

The borrow checker allows simultaneous mutable borrows of different fields in the
same struct, because they reference disjoint memory.

```rust
struct Database {
    users: Vec<String>,
    logs: Vec<String>,
}

fn process(db: &mut Database) {
    let users = &mut db.users; // Borrows one field
    let logs = &mut db.logs;   // Borrows a different field — OK
    users.push("alice".into());
    logs.push("added alice".into());
}
```

This also works with slice splitting:

```rust
fn main() {
    let mut data = vec![1, 2, 3, 4, 5];
    let (left, right) = data.split_at_mut(3);
    left[0] = 10;
    right[0] = 40; // Disjoint mutable borrows into the same Vec
    assert_eq!(data, vec![10, 2, 3, 40, 5]);
}
```

### Deferred Ownership with `Cow`

`Cow<'a, T>` (Clone on Write) holds either a borrowed reference or an owned value.
It clones only when mutation is needed, avoiding unnecessary allocations.

```rust
use std::borrow::Cow;

fn ensure_trailing_slash(path: &str) -> Cow<'_, str> {
    if path.ends_with('/') {
        Cow::Borrowed(path) // No allocation
    } else {
        Cow::Owned(format!("{path}/")) // Allocates only when needed
    }
}

fn main() {
    let a = ensure_trailing_slash("/home/user/");  // Borrowed — zero cost
    let b = ensure_trailing_slash("/home/user");    // Owned — one allocation
    println!("{a} {b}");
}
```

Use `Cow` when most inputs pass through unmodified but some need transformation.

---

## 4. Lifetime Elision Rules

Rust applies three elision rules so that most functions do not need explicit lifetime
annotations. When the rules are insufficient, you must annotate manually.

### Rule 1: Each Reference Parameter Gets Its Own Lifetime

The compiler assigns a distinct lifetime to every reference in the parameter list.

```rust
// What you write:
fn first(s: &str) -> &str { &s[..1] }

// What the compiler sees:
fn first<'a>(s: &'a str) -> &str { &s[..1] }
// (Rule 2 then assigns the output lifetime)
```

For multiple parameters:

```rust
// What you write:
fn longest(a: &str, b: &str) -> &str { if a.len() > b.len() { a } else { b } }

// What the compiler sees after Rule 1:
fn longest<'a, 'b>(a: &'a str, b: &'b str) -> &str { ... }
// Rules 2 and 3 cannot resolve the output — explicit annotation required.
```

### Rule 2: Single Input Lifetime Propagates to Output

If there is exactly one input lifetime parameter, that lifetime is assigned to all
output references.

```rust
// Elision works — one input reference:
fn first_byte(s: &[u8]) -> &u8 {
    &s[0]
}
// Equivalent to:
fn first_byte<'a>(s: &'a [u8]) -> &'a u8 {
    &s[0]
}
```

### Rule 3: `&self` / `&mut self` Lifetime Propagates to Output

In methods, if one of the parameters is `&self` or `&mut self`, the lifetime of
`self` is assigned to all output references.

```rust
struct Buffer {
    data: Vec<u8>,
}

impl Buffer {
    // Elision works — output borrows from self:
    fn as_slice(&self) -> &[u8] {
        &self.data
    }
    // Equivalent to:
    // fn as_slice<'a>(&'a self) -> &'a [u8]
}
```

### When Elision Fails

When there are multiple input lifetimes and no `self`, you must annotate:

```rust
// Does NOT compile without explicit lifetimes:
// fn longest(a: &str, b: &str) -> &str { ... }

// Fix: tell the compiler both inputs and the output share a lifetime:
fn longest<'a>(a: &'a str, b: &'a str) -> &'a str {
    if a.len() >= b.len() { a } else { b }
}
```

Another case — returning a reference from a struct with a lifetime:

```rust
struct Excerpt<'a> {
    text: &'a str,
}

impl<'a> Excerpt<'a> {
    fn text(&self) -> &str {
        self.text // Elision assigns &self's lifetime, which is correct here
    }
}
```

---

## 5. Interior Mutability

Interior mutability lets you mutate data through a shared reference (`&T`). This is
safe because the mutation rules are enforced at runtime or through atomic operations
rather than at compile time.

### `Cell<T>` — Copy Types, Single-Threaded

`Cell<T>` stores a value and allows get/set through `&self`. Works only for `Copy`
types because it copies values in and out.

```rust
use std::cell::Cell;

struct Counter {
    count: Cell<u32>,
}

impl Counter {
    fn increment(&self) {
        self.count.set(self.count.get() + 1);
    }
}
```

Zero overhead — no runtime borrow tracking. Cannot hand out references to the inner
value, so there is no risk of dangling pointers.

### `RefCell<T>` — Non-Copy Types, Single-Threaded

`RefCell<T>` provides runtime borrow checking. You call `.borrow()` for shared access
and `.borrow_mut()` for exclusive access. Violating the rules panics at runtime.

```rust
use std::cell::RefCell;

let data = RefCell::new(vec![1, 2, 3]);
data.borrow_mut().push(4);           // Exclusive borrow
println!("{:?}", data.borrow());      // Shared borrow: [1, 2, 3, 4]

// This would panic at runtime:
// let r1 = data.borrow();
// let r2 = data.borrow_mut(); // PANIC: already borrowed
```

### `Mutex<T>` — Multi-Threaded, Blocking

`Mutex<T>` provides exclusive access across threads. The lock guard implements
`DerefMut`, giving you `&mut T` access. The lock is released when the guard drops.

```rust
use std::sync::Mutex;

let counter = Mutex::new(0);
{
    let mut num = counter.lock().unwrap();
    *num += 1;
} // Lock released here
```

### `RwLock<T>` — Multi-Threaded, Read-Heavy

`RwLock<T>` allows multiple concurrent readers or one exclusive writer. Prefer it
over `Mutex` when reads vastly outnumber writes.

```rust
use std::sync::RwLock;

let config = RwLock::new(String::from("v1"));
// Many threads can read simultaneously:
let val = config.read().unwrap();
// One thread can write exclusively:
// let mut val = config.write().unwrap();
```

### Atomic Types — Lock-Free Primitives

For simple counters and flags, atomic types (`AtomicBool`, `AtomicU64`, etc.) provide
lock-free thread-safe mutation.

```rust
use std::sync::atomic::{AtomicU64, Ordering};

static REQUEST_COUNT: AtomicU64 = AtomicU64::new(0);

fn handle_request() {
    REQUEST_COUNT.fetch_add(1, Ordering::Relaxed);
}
```

### `OnceLock` — Lazy Initialization

`OnceLock` initializes a value exactly once and then provides shared access forever
after. Thread-safe and zero-cost after initialization.

```rust
use std::sync::OnceLock;

static CONFIG: OnceLock<String> = OnceLock::new();

fn get_config() -> &'static str {
    CONFIG.get_or_init(|| {
        std::fs::read_to_string("config.toml").unwrap_or_default()
    })
}
```

### Decision Table

| Need                          | Single-Threaded | Multi-Threaded   |
|-------------------------------|-----------------|------------------|
| Mutate Copy type              | `Cell<T>`       | `Atomic*`        |
| Mutate non-Copy type          | `RefCell<T>`    | `Mutex<T>`       |
| Many readers, rare writes     | `RefCell<T>`    | `RwLock<T>`      |
| Simple counter / flag         | `Cell<T>`       | `AtomicU64` etc. |
| Initialize once, read forever | `OnceCell`      | `OnceLock`       |
| Cheapest possible             | `Cell<T>`       | `Atomic*`        |

---

## 6. Smart Pointer Selection

### `Box<T>` — Heap Allocation

Use `Box<T>` when you need:
- A value with a known size on the heap (e.g., large arrays that would overflow the stack).
- Recursive types where the compiler cannot determine size at compile time.
- Trait objects (`Box<dyn Trait>`) for dynamic dispatch.

```rust
// Recursive type — Box breaks the infinite size:
enum List<T> {
    Cons(T, Box<List<T>>),
    Nil,
}

// Trait object:
fn make_writer() -> Box<dyn std::io::Write> {
    Box::new(std::io::stdout())
}
```

`Box<T>` has zero overhead beyond the heap allocation itself — it is a thin pointer
with no reference counting or metadata.

### `Rc<T>` — Shared Ownership, Single-Threaded

`Rc<T>` enables multiple owners of the same heap-allocated value. A reference count
tracks how many `Rc` clones exist; the value is dropped when the count reaches zero.

```rust
use std::rc::Rc;

let shared = Rc::new(vec![1, 2, 3]);
let a = Rc::clone(&shared); // Increments ref count (cheap)
let b = Rc::clone(&shared);
println!("References: {}", Rc::strong_count(&shared)); // 3
```

`Rc<T>` is not `Send` — it cannot cross thread boundaries. Use `Arc<T>` for that.

### `Arc<T>` — Shared Ownership, Multi-Threaded

`Arc<T>` is `Rc<T>` with atomic reference counting, making it safe to share across
threads. The atomic operations add slight overhead compared to `Rc`.

```rust
use std::sync::Arc;
use std::thread;

let data = Arc::new(vec![1, 2, 3]);
let handles: Vec<_> = (0..4).map(|_| {
    let data = Arc::clone(&data);
    thread::spawn(move || {
        println!("Sum: {}", data.iter().sum::<i32>());
    })
}).collect();

for h in handles {
    h.join().unwrap();
}
```

Combine `Arc<Mutex<T>>` when you need shared mutable state across threads.

### `Cow<'a, T>` — Clone on Write

As covered in Section 3, `Cow` defers cloning until mutation is required. It is
particularly useful in function signatures that may or may not need to allocate.

```rust
use std::borrow::Cow;

fn process_name(name: &str) -> Cow<'_, str> {
    if name.contains(' ') {
        Cow::Owned(name.replace(' ', "_"))
    } else {
        Cow::Borrowed(name)
    }
}
```

### `Pin<Box<T>>` — Self-Referential Types and Futures

`Pin` guarantees that a value will not be moved in memory. This is essential for:
- Async futures that hold references to their own fields across `.await` points.
- Self-referential structs that contain pointers to their own data.

```rust
use std::pin::Pin;
use std::future::Future;

fn make_future() -> Pin<Box<dyn Future<Output = i32>>> {
    Box::pin(async {
        42
    })
}
```

Most users encounter `Pin` through async runtimes rather than constructing it
manually. If you are not writing an async runtime or a self-referential struct,
you likely do not need `Pin` directly.

### Smart Pointer Selection Guide

| Situation                              | Use                |
|----------------------------------------|--------------------|
| Single owner, heap allocation needed   | `Box<T>`           |
| Trait object with dynamic dispatch     | `Box<dyn Trait>`   |
| Multiple owners, single thread         | `Rc<T>`            |
| Multiple owners, multiple threads      | `Arc<T>`           |
| Shared + mutable, single thread        | `Rc<RefCell<T>>`   |
| Shared + mutable, multiple threads     | `Arc<Mutex<T>>`    |
| Avoid cloning when not always needed   | `Cow<'a, T>`       |
| Async futures / self-referential types | `Pin<Box<T>>`      |
| Recursive data structure               | `Box<T>`           |

---

## 7. Common Borrow Checker Errors & Solutions

### Error: "cannot borrow X as mutable because it is also borrowed as immutable"

```rust
fn main() {
    let mut v = vec![1, 2, 3];
    let first = &v[0];     // Shared borrow
    v.push(4);             // Mutable borrow — ERROR
    println!("{first}");
}
```

```
error[E0502]: cannot borrow `v` as mutable because it is also borrowed as immutable
 --> src/main.rs:4:5
  |
3 |     let first = &v[0];
  |                  - immutable borrow occurs here
4 |     v.push(4);
  |     ^^^^^^^^^ mutable borrow occurs here
5 |     println!("{first}");
  |               ------- immutable borrow later used here
```

**Why:** `push` may reallocate the `Vec`'s backing storage, invalidating `first`.
The borrow checker prevents this dangling pointer.

**Fix 1:** Finish using the shared borrow before mutating.

```rust
fn main() {
    let mut v = vec![1, 2, 3];
    let first = v[0]; // Copy the value instead of borrowing
    v.push(4);
    println!("{first}");
}
```

**Fix 2:** Restructure to separate the read and write phases.

```rust
fn main() {
    let mut v = vec![1, 2, 3];
    let first = v[0];
    v.push(4);
    println!("{first}"); // Uses the copied value, not a reference
}
```

### Error: "borrowed value does not live long enough"

```rust
fn longest_line() -> &str {
    let content = String::from("hello\nworld");
    content.lines().next().unwrap()
}
```

```
error[E0515]: cannot return reference to local variable `content`
 --> src/main.rs:3:5
  |
3 |     content.lines().next().unwrap()
  |     ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^ returns a reference to data owned by the current function
```

**Why:** `content` is dropped at the end of `longest_line`, so any reference into it
would dangle. The compiler catches this.

**Fix 1:** Return an owned value.

```rust
fn longest_line() -> String {
    let content = String::from("hello\nworld");
    content.lines().next().unwrap().to_string()
}
```

**Fix 2:** Accept the data as a parameter so the caller controls the lifetime.

```rust
fn longest_line(content: &str) -> &str {
    content.lines().next().unwrap()
}
```

### Error: "cannot move out of borrowed content"

```rust
fn take_name(user: &User) -> String {
    user.name // Attempts to move name out of a borrow
}
```

```
error[E0507]: cannot move out of `user.name` which is behind a shared reference
 --> src/main.rs:2:5
  |
2 |     user.name
  |     ^^^^^^^^^ move occurs because `user.name` has type `String`,
  |               which does not implement the `Copy` trait
```

**Why:** You have a `&User`, which means you are borrowing the struct. Moving a
field out would leave the struct in a partially initialized state.

**Fix 1:** Clone the value.

```rust
fn take_name(user: &User) -> String {
    user.name.clone()
}
```

**Fix 2:** Take ownership of the whole struct.

```rust
fn take_name(user: User) -> String {
    user.name
}
```

**Fix 3:** Return a reference instead.

```rust
fn get_name(user: &User) -> &str {
    &user.name
}
```

### Error: "closure may outlive the current function"

```rust
fn spawn_printer(msg: &str) {
    std::thread::spawn(|| {
        println!("{msg}");
    });
}
```

```
error[E0373]: closure may outlive the current function, but it borrows `msg`,
              which is owned by the current function
```

**Why:** `thread::spawn` requires `'static` because the thread may outlive the
calling function. `msg` is a local borrow that will be invalid once the function
returns.

**Fix 1:** Move an owned value into the closure.

```rust
fn spawn_printer(msg: String) {
    std::thread::spawn(move || {
        println!("{msg}");
    });
}
```

**Fix 2:** Use scoped threads (Rust 1.63+), which guarantee the thread joins
before the borrow expires.

```rust
fn scoped_printer(msg: &str) {
    std::thread::scope(|s| {
        s.spawn(|| {
            println!("{msg}"); // OK — scoped thread borrows are valid
        });
    }); // Thread is joined here, before msg's borrow expires
}
```

### Error: "cannot return reference to temporary value"

```rust
fn greeting() -> &str {
    &String::from("hello")
}
```

```
error[E0515]: cannot return reference to temporary value
 --> src/main.rs:2:5
  |
2 |     &String::from("hello")
  |     ^--------------------- temporary value created here
  |     |
  |     returns a reference to data owned by the current function
```

**Why:** The `String` is created as a temporary inside the function. It will be
dropped at the end of the function, leaving the reference dangling.

**Fix 1:** Return an owned `String`.

```rust
fn greeting() -> String {
    String::from("hello")
}
```

**Fix 2:** Return a string literal, which has `'static` lifetime.

```rust
fn greeting() -> &'static str {
    "hello"
}
```

---

## 8. API Design Patterns

### Accept Broad, Return Narrow

Design function parameters to accept the widest useful type. Return the most
concrete type. This maximizes flexibility for callers while keeping the return
type predictable.

```rust
use std::path::Path;

// Bad: forces caller to construct a &String
fn read_config_bad(path: &String) -> String { todo!() }

// Good: accepts &str, &String, &Path, &OsStr, etc.
fn read_config(path: impl AsRef<Path>) -> String {
    let path = path.as_ref();
    std::fs::read_to_string(path).unwrap_or_default()
}
```

### The `Into` Pattern for Flexible Constructors

Use `Into<T>` bounds to let callers pass various types that can be converted
into the target type.

```rust
struct Endpoint {
    url: String,
}

impl Endpoint {
    fn new(url: impl Into<String>) -> Self {
        Self { url: url.into() }
    }
}

fn main() {
    let a = Endpoint::new("https://example.com"); // &str -> String
    let b = Endpoint::new(String::from("https://example.com")); // String -> String (no clone)
}
```

This avoids forcing callers to call `.to_string()` or `.into()` at every call site
while remaining zero-cost when the caller already has a `String`.

### Builder Pattern with Ownership Transfer

Builders consume `self` at each step to enforce a linear construction flow and
prevent reuse of partially configured builders.

```rust
struct RequestBuilder {
    url: String,
    timeout_ms: u64,
    headers: Vec<(String, String)>,
}

impl RequestBuilder {
    fn new(url: impl Into<String>) -> Self {
        Self { url: url.into(), timeout_ms: 5000, headers: Vec::new() }
    }

    fn timeout(mut self, ms: u64) -> Self {
        self.timeout_ms = ms;
        self
    }

    fn header(mut self, key: impl Into<String>, val: impl Into<String>) -> Self {
        self.headers.push((key.into(), val.into()));
        self
    }

    fn build(self) -> Request {
        Request {
            url: self.url,
            timeout_ms: self.timeout_ms,
            headers: self.headers,
        }
    }
}

struct Request {
    url: String,
    timeout_ms: u64,
    headers: Vec<(String, String)>,
}
```

The consuming `self` pattern ensures the builder is moved into `build()` and cannot
be accidentally reused afterward.

### Iterator-Based APIs for Zero-Copy Processing

Return iterators instead of collected `Vec`s to let callers decide when to allocate.

```rust
struct LogFile {
    content: String,
}

impl LogFile {
    fn errors(&self) -> impl Iterator<Item = &str> {
        self.content
            .lines()
            .filter(|line| line.starts_with("ERROR"))
    }
}

fn main() {
    let log = LogFile { content: String::from("INFO ok\nERROR fail\nINFO ok") };

    // Zero allocation — just iterates and borrows:
    for err in log.errors() {
        println!("{err}");
    }

    // Caller decides to collect only when needed:
    let errors: Vec<&str> = log.errors().collect();
    println!("Total errors: {}", errors.len());
}
```

### Entry API for Maps

The entry API eliminates redundant lookups when inserting or updating map values.
It splits the operation into `Occupied` and `Vacant` variants.

```rust
use std::collections::HashMap;

fn count_words(text: &str) -> HashMap<&str, u32> {
    let mut counts = HashMap::new();
    for word in text.split_whitespace() {
        *counts.entry(word).or_insert(0) += 1;
    }
    counts
}
```

For more complex initialization:

```rust
use std::collections::HashMap;

fn group_by_length(words: &[&str]) -> HashMap<usize, Vec<&str>> {
    let mut groups: HashMap<usize, Vec<&str>> = HashMap::new();
    for &word in words {
        groups.entry(word.len()).or_default().push(word);
    }
    groups
}
```

The entry API performs a single hash lookup rather than a `contains_key` check
followed by a separate `insert`. This is both more efficient and more idiomatic.
The `or_default()` call uses `Default::default()` — for `Vec`, that is an empty
vector — avoiding the need to specify the initial value explicitly.
