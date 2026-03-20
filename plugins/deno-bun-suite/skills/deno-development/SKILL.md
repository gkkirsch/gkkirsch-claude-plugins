---
name: deno-development
description: >
  Deno 2 runtime development patterns and APIs.
  Use when building applications with Deno, using Deno KV, deploying to
  Deno Deploy, building with Fresh framework, or migrating from Node.js.
  Triggers: "deno", "deno 2", "deno kv", "deno deploy", "fresh framework",
  "deno permissions", "deno serve".
  NOT for: Node.js-only projects, Bun-specific features, browser JavaScript.
version: 1.0.0
allowed-tools: Read, Grep, Glob, Bash
---

# Deno 2 Development

## HTTP Server

```typescript
// Deno.serve() — the recommended HTTP server API
Deno.serve({ port: 8000 }, async (req: Request): Promise<Response> => {
  const url = new URL(req.url);

  if (url.pathname === "/api/health") {
    return Response.json({ status: "ok", timestamp: Date.now() });
  }

  if (url.pathname === "/api/users" && req.method === "POST") {
    const body = await req.json();
    if (!body.name || typeof body.name !== "string") {
      return Response.json({ error: "name is required" }, { status: 400 });
    }
    return Response.json({ id: crypto.randomUUID(), name: body.name }, { status: 201 });
  }

  if (url.pathname.startsWith("/api/users/") && req.method === "GET") {
    const id = url.pathname.split("/").pop();
    return Response.json({ id, name: "Example User" });
  }

  return new Response("Not Found", { status: 404 });
});

// With TLS
Deno.serve({
  port: 443,
  cert: await Deno.readTextFile("./cert.pem"),
  key: await Deno.readTextFile("./key.pem"),
}, handler);

// AbortController for graceful shutdown
const controller = new AbortController();
const server = Deno.serve({ signal: controller.signal, port: 8000 }, handler);

Deno.addSignalListener("SIGTERM", () => {
  console.log("Shutting down...");
  controller.abort();
});

await server.finished;
```

## Permissions Model

```bash
# Granular permissions — Deno's killer feature
deno run --allow-net=api.example.com,localhost:8000 server.ts
deno run --allow-read=./data,./config --allow-write=./data server.ts
deno run --allow-env=DATABASE_URL,API_KEY app.ts
deno run --allow-run=git,npm script.ts

# Permission flags
# --allow-read[=<paths>]    File system read
# --allow-write[=<paths>]   File system write
# --allow-net[=<hosts>]     Network access
# --allow-env[=<vars>]      Environment variables
# --allow-run[=<programs>]  Subprocess execution
# --allow-ffi               Foreign function interface
# --allow-sys               System information
# --allow-hrtime            High-resolution time
# --allow-all / -A          All permissions (development only!)

# deno.json — configure permissions for deno task
{
  "tasks": {
    "dev": "deno run --allow-net --allow-read --allow-env --watch main.ts",
    "start": "deno run --allow-net=:8000 --allow-read=./public --allow-env=DATABASE_URL main.ts"
  }
}
```

```typescript
// Runtime permission requests
const status = await Deno.permissions.request({ name: "read", path: "./data" });
if (status.state === "granted") {
  const data = await Deno.readTextFile("./data/config.json");
}

// Check permissions without prompting
const netStatus = await Deno.permissions.query({ name: "net", host: "api.example.com" });
if (netStatus.state !== "granted") {
  console.error("Network access to api.example.com is required");
  Deno.exit(1);
}
```

## Deno KV

```typescript
// Open KV store (local SQLite in development, managed in Deno Deploy)
const kv = await Deno.openKv();

// Basic CRUD
await kv.set(["users", "user-123"], { name: "Alice", email: "alice@example.com" });
const result = await kv.get<{ name: string; email: string }>(["users", "user-123"]);
console.log(result.value); // { name: "Alice", email: "alice@example.com" }
console.log(result.versionstamp); // For optimistic concurrency

await kv.delete(["users", "user-123"]);

// List with prefix
const iter = kv.list<{ name: string }>({ prefix: ["users"] });
for await (const entry of iter) {
  console.log(entry.key, entry.value);
}

// List with pagination
const page1 = kv.list({ prefix: ["users"] }, { limit: 10 });
const entries = [];
for await (const entry of page1) entries.push(entry);
const cursor = page1.cursor; // Use for next page

// Atomic transactions with optimistic concurrency
const user = await kv.get(["users", "user-123"]);
const atomicResult = await kv.atomic()
  .check(user) // Fails if versionstamp changed
  .set(["users", "user-123"], { ...user.value, name: "Updated" })
  .set(["users_by_email", "alice@example.com"], "user-123") // Secondary index
  .commit();

if (!atomicResult.ok) {
  console.error("Concurrent modification detected, retry");
}

// Atomic with sum (counters)
await kv.atomic()
  .sum(["page_views", "home"], 1n)
  .commit();
const views = await kv.get(["page_views", "home"]);
console.log(views.value); // Deno.KvU64 { value: 42n }

// Enqueue — built-in message queue
await kv.enqueue({ type: "send_email", to: "user@example.com", subject: "Welcome" });

// Listen for queued messages
kv.listenQueue(async (msg: { type: string; to: string; subject: string }) => {
  if (msg.type === "send_email") {
    await sendEmail(msg.to, msg.subject);
  }
});

// Watch for real-time changes
const stream = kv.watch([["users", "user-123"], ["config", "theme"]]);
for await (const entries of stream) {
  console.log("Changed:", entries);
}
```

## Key Design Patterns for KV

```typescript
// Pattern 1: Secondary indexes
async function createUser(kv: Deno.Kv, user: { id: string; email: string; name: string }) {
  const result = await kv.atomic()
    .check({ key: ["users", user.id], versionstamp: null }) // Must not exist
    .check({ key: ["users_by_email", user.email], versionstamp: null })
    .set(["users", user.id], user)
    .set(["users_by_email", user.email], user.id)
    .commit();
  if (!result.ok) throw new Error("User already exists");
  return user;
}

async function getUserByEmail(kv: Deno.Kv, email: string) {
  const ref = await kv.get<string>(["users_by_email", email]);
  if (!ref.value) return null;
  const user = await kv.get(["users", ref.value]);
  return user.value;
}

// Pattern 2: Time-sorted entries (reverse chronological)
function timeKey(prefix: string[]): Deno.KvKey {
  const invertedTs = Number.MAX_SAFE_INTEGER - Date.now();
  return [...prefix, invertedTs, crypto.randomUUID()];
}

await kv.set(timeKey(["posts", "user-123"]), { title: "My Post", body: "..." });

// Pattern 3: TTL with expireIn
await kv.set(
  ["sessions", sessionId],
  { userId: "user-123", createdAt: Date.now() },
  { expireIn: 24 * 60 * 60 * 1000 } // 24 hours
);
```

## Node.js Compatibility

```typescript
// Use node: prefix for built-in modules
import { readFile } from "node:fs/promises";
import { createServer } from "node:http";
import { join } from "node:path";
import process from "node:process";

// npm packages work with deno add
// $ deno add npm:express npm:zod
import express from "npm:express";
import { z } from "npm:zod";

// __dirname replacement
const __dirname = import.meta.dirname!;
const __filename = import.meta.filename!;

// process.env replacement
const dbUrl = Deno.env.get("DATABASE_URL");
```

## deno.json Configuration

```json
{
  "compilerOptions": {
    "lib": ["deno.window", "deno.unstable"],
    "strict": true
  },
  "imports": {
    "@std/assert": "jsr:@std/assert@1",
    "@std/http": "jsr:@std/http@1",
    "@std/path": "jsr:@std/path@1",
    "hono": "npm:hono@4",
    "zod": "npm:zod@3"
  },
  "tasks": {
    "dev": "deno run --watch --allow-net --allow-read --allow-env main.ts",
    "start": "deno run --allow-net=:8000 --allow-env=DATABASE_URL main.ts",
    "test": "deno test --allow-read --allow-net",
    "check": "deno check main.ts",
    "lint": "deno lint",
    "fmt": "deno fmt"
  },
  "lint": {
    "include": ["src/"],
    "exclude": ["vendor/"]
  },
  "fmt": {
    "useTabs": false,
    "lineWidth": 100,
    "indentWidth": 2,
    "singleQuote": true
  },
  "test": {
    "include": ["tests/"]
  }
}
```

## Testing

```typescript
import { assertEquals, assertRejects } from "@std/assert";

// Basic test
Deno.test("addition works", () => {
  assertEquals(1 + 1, 2);
});

// Async test
Deno.test("fetch data", async () => {
  const resp = await fetch("https://api.example.com/data");
  assertEquals(resp.status, 200);
});

// Test with setup/teardown
Deno.test("database operations", async (t) => {
  const kv = await Deno.openKv(":memory:");

  await t.step("create user", async () => {
    await kv.set(["users", "1"], { name: "Test" });
    const user = await kv.get(["users", "1"]);
    assertEquals(user.value, { name: "Test" });
  });

  await t.step("delete user", async () => {
    await kv.delete(["users", "1"]);
    const user = await kv.get(["users", "1"]);
    assertEquals(user.value, null);
  });

  kv.close();
});

// BDD style
import { describe, it, expect } from "jsr:@std/testing/bdd";

describe("UserService", () => {
  it("should create a user", async () => {
    const user = await createUser({ name: "Alice" });
    expect(user.id).toBeDefined();
  });
});
```

## Fresh Framework (Deno's Full-Stack Framework)

```typescript
// routes/index.tsx — page route
export default function Home() {
  return (
    <div>
      <h1>Welcome to Fresh</h1>
      <Counter />
    </div>
  );
}

// routes/api/users.ts — API route
import { Handlers } from "$fresh/server.ts";

export const handler: Handlers = {
  async GET(_req, _ctx) {
    const users = await getUsers();
    return Response.json(users);
  },
  async POST(req, _ctx) {
    const body = await req.json();
    const user = await createUser(body);
    return Response.json(user, { status: 201 });
  },
};

// islands/Counter.tsx — interactive island component
import { useSignal } from "@preact/signals";

export default function Counter() {
  const count = useSignal(0);
  return (
    <div>
      <p>Count: {count}</p>
      <button onClick={() => count.value++}>+1</button>
    </div>
  );
}

// routes/users/[id].tsx — dynamic route with data loading
import { PageProps, Handlers } from "$fresh/server.ts";

interface User { id: string; name: string; }

export const handler: Handlers<User> = {
  async GET(_req, ctx) {
    const user = await getUser(ctx.params.id);
    if (!user) return ctx.renderNotFound();
    return ctx.render(user);
  },
};

export default function UserPage({ data }: PageProps<User>) {
  return <h1>{data.name}</h1>;
}

// Middleware — routes/_middleware.ts
import { FreshContext } from "$fresh/server.ts";

export async function handler(req: Request, ctx: FreshContext) {
  const start = Date.now();
  const resp = await ctx.next();
  const duration = Date.now() - start;
  resp.headers.set("X-Response-Time", `${duration}ms`);
  return resp;
}
```

## Deno Deploy

```typescript
// deployctl — deploy to Deno Deploy
// $ deployctl deploy --project=my-app main.ts

// KV works automatically on Deploy (backed by FoundationDB)
const kv = await Deno.openKv(); // No path needed on Deploy

// BroadcastChannel for multi-instance communication
const channel = new BroadcastChannel("app-events");
channel.onmessage = (event) => {
  console.log("Received:", event.data);
};
channel.postMessage({ type: "cache_invalidation", key: "users" });

// Cron jobs (Deno Deploy feature)
Deno.cron("daily-cleanup", "0 0 * * *", async () => {
  console.log("Running daily cleanup...");
  const kv = await Deno.openKv();
  // Clean up expired data
});

Deno.cron("every-5-minutes", "*/5 * * * *", async () => {
  await checkHealthEndpoints();
});
```

## Standard Library Highlights

```typescript
// @std/http — HTTP utilities
import { serveDir } from "@std/http/file-server";

Deno.serve((req) => serveDir(req, { fsRoot: "./public" }));

// @std/path — Path manipulation
import { join, basename, extname, resolve } from "@std/path";
const fullPath = join(import.meta.dirname!, "data", "config.json");

// @std/async — Async utilities
import { delay, deadline, retry } from "@std/async";

await delay(1000); // Sleep 1 second

const result = await deadline(fetchData(), 5000); // Timeout after 5s

const data = await retry(async () => {
  return await unstableApiCall();
}, { maxAttempts: 3, minTimeout: 1000, multiplier: 2 });

// @std/streams — Stream utilities
import { TextLineStream } from "@std/streams";

const file = await Deno.open("large.log");
const lines = file.readable
  .pipeThrough(new TextDecoderStream())
  .pipeThrough(new TextLineStream());

for await (const line of lines) {
  if (line.includes("ERROR")) console.log(line);
}
```

## Gotchas

1. **--allow-all in production** — defeats the entire security model. Always specify exact permissions. Use `deno task` with predefined permission sets.

2. **KV key design** — keys are arrays of strings/numbers/bigints/booleans/Uint8Arrays. Use hierarchical keys (`["users", id]`) not flat strings. Keys are sorted lexicographically — numbers sort naturally, strings sort by UTF-8 bytes.

3. **npm package compatibility** — most npm packages work, but some with native C++ addons (bcrypt, canvas, sharp) may not. Check `deno.land/x` for Deno-native alternatives. Use `node:` prefix for Node.js built-ins.

4. **Import maps vs URL imports** — always use `deno.json` import maps for production. Raw URL imports (`import x from "https://..."`) work but are harder to manage and audit.

5. **KV transaction limits** — atomic operations support max 10 mutations and 10 checks per transaction. Design around this limit — batch operations need multiple transactions.

6. **Fresh island hydration cost** — only use islands for truly interactive components. Static content should stay in routes (server-rendered, zero JS). Every island adds to client bundle.
