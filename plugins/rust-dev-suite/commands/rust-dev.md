---
name: rust-dev
description: >
  Expert Rust development assistance. Get help with ownership and borrowing, lifetime annotations,
  async/await patterns, error handling strategies, unsafe code review, FFI bindings, embedded/no_std
  development, cargo workspace management, testing strategies, and production architecture with
  Actix, Axum, Diesel, SQLx, Serde, and Clap.
version: 1.0.0
argument-hint: "<topic-or-question> [--agent architect|systems|cargo|testing]"
allowed-tools: Read, Grep, Glob, Bash, Write, Edit
model: sonnet
user-invocable: true

metadata:
  superbot:
    emoji: "🦀"
---

# /rust-dev

Expert Rust development assistance covering the full spectrum of systems programming — from ownership fundamentals to production-grade async services.

## What You Can Do

- **Architecture** — Design Rust applications with correct ownership, lifetime, and trait patterns
- **Systems** — Build FFI bindings, embedded firmware, no_std libraries, and performance-critical code
- **Cargo** — Manage workspaces, feature flags, build scripts, conditional compilation, and publishing
- **Testing** — Write property tests, benchmarks, fuzz tests, and integration test harnesses

## How to Use

```
/rust-dev <your question or topic>
```

### Examples

```
/rust-dev How do I structure a multi-crate workspace with shared types?
/rust-dev Review my async code for lifetime issues
/rust-dev Help me write FFI bindings for this C library
/rust-dev Set up property-based testing with proptest
/rust-dev Design error types for my REST API
/rust-dev Optimize this hot loop — it's allocating too much
```

### Subcommands

#### `architecture` — Design & Review Rust Code

Get expert guidance on ownership patterns, lifetime annotations, trait design, async architecture, error handling strategies, and macro design.

**Handled by**: `rust-architect` agent

**Examples**:
```
/rust-dev architecture Design a plugin system using trait objects
/rust-dev architecture Review my lifetime annotations in this module
/rust-dev architecture Help me choose between Arc<Mutex<T>> and channels
/rust-dev architecture Design error types with thiserror for my crate
```

#### `systems` — Low-Level & Systems Programming

Get help with FFI/C interop, embedded development, no_std libraries, memory layout optimization, SIMD, and performance tuning.

**Handled by**: `systems-engineer` agent

**Examples**:
```
/rust-dev systems Write safe Rust bindings for this C header
/rust-dev systems Help me write a no_std driver for this sensor
/rust-dev systems Optimize memory layout of this struct
/rust-dev systems Profile and fix this allocation bottleneck
```

#### `cargo` — Build System & Package Management

Get help with cargo workspaces, feature flags, build scripts, conditional compilation, cross-compilation, and crate publishing.

**Handled by**: `cargo-expert` agent

**Examples**:
```
/rust-dev cargo Set up a workspace with shared dependencies
/rust-dev cargo Add feature flags for optional functionality
/rust-dev cargo Write a build script to generate bindings
/rust-dev cargo Prepare my crate for publishing to crates.io
```

#### `testing` — Testing & Quality Assurance

Get help with unit tests, integration tests, property-based testing, fuzzing, benchmarking, mocking, and test architecture.

**Handled by**: `rust-testing-expert` agent

**Examples**:
```
/rust-dev testing Set up proptest for my parser
/rust-dev testing Write benchmarks with criterion
/rust-dev testing Add fuzz testing with cargo-fuzz
/rust-dev testing Design an integration test harness for my API
```

## Specialist Agents

| Agent | Focus Areas |
|-------|-------------|
| **rust-architect** | Ownership, lifetimes, async/await, error handling, traits, generics, macros, unsafe review |
| **systems-engineer** | FFI, embedded, no_std, performance, memory layout, SIMD, allocation optimization |
| **cargo-expert** | Workspaces, features, build scripts, conditional compilation, publishing, CI/CD |
| **rust-testing-expert** | Property testing, benchmarks, fuzzing, mocking, integration tests, test architecture |

## Reference Materials

The following deep-dive references are available:

- **Ownership Patterns** — Common ownership and borrowing patterns, lifetime elision rules, interior mutability, and smart pointer selection
- **Async Rust Guide** — Tokio and async-std patterns, executor architecture, pinning, cancellation safety, and structured concurrency
- **Error Handling** — Error type design with thiserror and anyhow, Result combinators, error context propagation, and panic strategies

## Tips

- Be specific about your Rust edition (2018, 2021, 2024) — patterns differ across editions
- Share your `Cargo.toml` when asking about dependency or feature questions
- Include compiler error messages — they contain crucial context for diagnosing issues
- Mention your target platform for systems/embedded questions (x86_64, ARM, RISC-V, WASM)
- Specify your async runtime (Tokio, async-std, smol) for async questions
