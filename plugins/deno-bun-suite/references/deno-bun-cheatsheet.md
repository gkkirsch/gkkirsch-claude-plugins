# Deno & Bun Cheatsheet

## Runtime Comparison

| Feature | Node.js | Deno 2 | Bun |
|---------|---------|--------|-----|
| Engine | V8 | V8 | JavaScriptCore |
| TypeScript | Via tsc | Native | Native |
| Permissions | None | Granular | None |
| HTTP Server | http module | Deno.serve() | Bun.serve() |
| Package Manager | npm | deno add | bun install |
| Test Runner | node --test | deno test | bun test |
| Bundler | No | No | bun build |
| Startup | ~40ms | ~30ms | ~5ms |

## Deno Quick Reference

### HTTP Server
```typescript
Deno.serve({ port: 8000 }, (req) => {
  return Response.json({ ok: true });
});
```

### Permissions
```bash
deno run --allow-net=localhost:8000 --allow-read=./data --allow-env=DB_URL app.ts
```

### KV Store
```typescript
const kv = await Deno.openKv();
await kv.set(["users", id], user);
const result = await kv.get(["users", id]);
// Atomic with optimistic concurrency
await kv.atomic().check(result).set(["users", id], updated).commit();
// TTL
await kv.set(["sessions", id], data, { expireIn: 86400000 });
```

### KV Key Patterns
```
["users", id]              // Primary key
["users_by_email", email]  // Secondary index
["posts", invertedTs, id]  // Reverse chronological
```

### Node Compat
```typescript
import { readFile } from "node:fs/promises";  // node: prefix
import express from "npm:express";             // npm: prefix
const dir = import.meta.dirname;               // __dirname replacement
const env = Deno.env.get("KEY");               // process.env replacement
```

### deno.json
```json
{
  "imports": { "hono": "npm:hono@4", "@std/assert": "jsr:@std/assert@1" },
  "tasks": { "dev": "deno run --watch --allow-net main.ts" }
}
```

### Testing
```typescript
import { assertEquals } from "@std/assert";
Deno.test("works", () => assertEquals(1 + 1, 2));
Deno.test("async", async (t) => {
  await t.step("step 1", () => { /* ... */ });
});
```

### Fresh Framework
```
routes/index.tsx          -> Page (server-rendered)
routes/api/users.ts       -> API endpoint (Handlers)
islands/Counter.tsx        -> Interactive component (hydrated)
routes/_middleware.ts      -> Middleware
routes/users/[id].tsx      -> Dynamic route
```

### Deno Deploy
```typescript
Deno.cron("cleanup", "0 0 * * *", async () => { /* daily */ });
const channel = new BroadcastChannel("events"); // Multi-instance
```

## Bun Quick Reference

### HTTP Server
```typescript
Bun.serve({
  port: 3000,
  fetch(req) { return Response.json({ ok: true }); },
  error(err) { return new Response("Error", { status: 500 }); },
});
```

### WebSocket
```typescript
Bun.serve({
  fetch(req, server) {
    server.upgrade(req, { data: { userId: "123" } });
  },
  websocket: {
    open(ws) { ws.subscribe("chat"); },
    message(ws, msg) { ws.publish("chat", msg); },
    close(ws) { ws.unsubscribe("chat"); },
  },
});
```

### File I/O
```typescript
const file = Bun.file("./data.json");
const text = await file.text();
const json = await file.json();
await Bun.write("./out.txt", "content");
await Bun.write("./copy.txt", Bun.file("./src.txt"));
```

### SQLite
```typescript
import { Database } from "bun:sqlite";
const db = new Database("app.db");
db.run("PRAGMA journal_mode = WAL");
const stmt = db.prepare("SELECT * FROM users WHERE id = $id");
const user = stmt.get({ $id: "123" });
const tx = db.transaction((data) => { /* atomic ops */ });
```

### Bundler
```typescript
await Bun.build({
  entrypoints: ["./src/index.ts"],
  outdir: "./dist",
  target: "browser",
  minify: true,
  splitting: true,
});
```

### Testing
```typescript
import { describe, it, expect, mock } from "bun:test";
describe("suite", () => {
  it("works", () => expect(1 + 1).toBe(2));
  it("mocks", () => {
    const fn = mock(() => 42);
    expect(fn()).toBe(42);
    expect(fn).toHaveBeenCalled();
  });
});
// Run: bun test --watch --coverage
```

### Shell
```typescript
import { $ } from "bun";
const out = await $`git status`.text();
await $`echo ${safeVar}`;  // Auto-escaped
```

### Package Manager
```bash
bun install               # Install all
bun add zod               # Add dependency
bun add -d typescript     # Dev dependency
bunx create-vite app      # npx equivalent
```

### Built-in Utilities
```typescript
// Password hashing (argon2)
const hash = await Bun.password.hash("pass");
const ok = await Bun.password.verify("pass", hash);

// Fast hashing
const digest = new Bun.CryptoHasher("sha256").update("data").digest("hex");

// Compression
const gz = Bun.gzipSync(Buffer.from("text"));
```

## Decision Guide

| Scenario | Best Choice |
|----------|-------------|
| New API server | Bun (fastest HTTP) |
| Edge/serverless | Deno (Deploy + permissions) |
| Existing Node project | Node.js (no migration) |
| Security-sensitive | Deno (granular permissions) |
| Full-stack SSR | Deno (Fresh) |
| CLI tools | Bun (fastest startup) |
| Enterprise | Node.js (ecosystem maturity) |
| Embedded DB needed | Bun (bun:sqlite) |
| KV store needed | Deno (Deno KV) |

## Migration from Node.js

```
1. require() -> import (ESM)
2. __dirname -> import.meta.dirname
3. process.env -> Deno.env.get() (Deno only)
4. fs -> node:fs (add prefix)
5. Check native C++ addons (won't work)
6. Update package manager commands
7. Update CI/CD scripts
```

## Common Gotchas

- **Deno**: Always specify permissions. Don't use `--allow-all` in production
- **Deno**: KV atomic ops limited to 10 mutations per transaction
- **Deno**: Fresh islands add to client bundle — use sparingly
- **Bun**: C++ native addons (bcrypt, sharp) don't work — use alternatives
- **Bun**: `bun.lockb` is binary — not diffable in PRs
- **Bun**: `--hot` preserves state, `--watch` restarts process
- **Both**: Test every npm dependency — some Node APIs are stubs
- **Both**: Use `node:` prefix for Node.js built-in imports
