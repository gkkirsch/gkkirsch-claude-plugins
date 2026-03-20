---
name: bun-development
description: >
  Bun runtime development patterns and APIs.
  Use when building applications with Bun, using Bun.serve, bun:sqlite,
  the Bun bundler, test runner, or migrating from Node.js to Bun.
  Triggers: "bun", "bun serve", "bun build", "bun test", "bun sqlite",
  "bun install", "bun runtime".
  NOT for: Node.js-only projects, Deno-specific features, browser JavaScript.
version: 1.0.0
allowed-tools: Read, Grep, Glob, Bash
---

# Bun Development

## HTTP Server

```typescript
// Bun.serve() — the fastest HTTP server in JavaScript
const server = Bun.serve({
  port: 3000,
  async fetch(req: Request): Promise<Response> {
    const url = new URL(req.url);

    // Routing
    if (url.pathname === "/api/health") {
      return Response.json({ status: "ok", runtime: "bun" });
    }

    if (url.pathname === "/api/users" && req.method === "POST") {
      const body = await req.json();
      return Response.json(
        { id: crypto.randomUUID(), ...body },
        { status: 201 }
      );
    }

    // Static files
    if (url.pathname.startsWith("/public/")) {
      const file = Bun.file(`./static${url.pathname}`);
      if (await file.exists()) return new Response(file);
    }

    return new Response("Not Found", { status: 404 });
  },

  // Error handler — don't leak stack traces
  error(error: Error): Response {
    console.error(error);
    return Response.json(
      { error: "Internal Server Error" },
      { status: 500 }
    );
  },
});

console.log(`Listening on ${server.hostname}:${server.port}`);

// Graceful shutdown
process.on("SIGTERM", () => {
  server.stop();
  process.exit(0);
});
```

## WebSocket Server

```typescript
const server = Bun.serve({
  port: 3000,
  fetch(req, server) {
    // Upgrade HTTP to WebSocket
    const url = new URL(req.url);
    if (url.pathname === "/ws") {
      const userId = url.searchParams.get("userId");
      const success = server.upgrade(req, {
        data: { userId, connectedAt: Date.now() },
      });
      if (success) return undefined; // Bun handles the response
      return new Response("WebSocket upgrade failed", { status: 400 });
    }
    return new Response("Hello");
  },

  websocket: {
    open(ws) {
      console.log(`Connected: ${ws.data.userId}`);
      ws.subscribe("chat"); // Subscribe to topic
    },
    message(ws, message) {
      // Broadcast to all subscribers
      ws.publish("chat", JSON.stringify({
        from: ws.data.userId,
        text: String(message),
        timestamp: Date.now(),
      }));
    },
    close(ws) {
      console.log(`Disconnected: ${ws.data.userId}`);
      ws.unsubscribe("chat");
    },
    perMessageDeflate: true, // Compression
  },
});
```

## File I/O

```typescript
// Bun.file() — lazy file reference (doesn't read until needed)
const file = Bun.file("./data.json");
console.log(file.size);  // Size without reading content
console.log(file.type);  // MIME type

// Read
const text = await file.text();
const json = await file.json();
const buffer = await file.arrayBuffer();
const stream = file.stream(); // ReadableStream

// Write
await Bun.write("./output.txt", "Hello, Bun!");
await Bun.write("./data.json", JSON.stringify({ key: "value" }));
await Bun.write("./copy.txt", Bun.file("./original.txt")); // File copy
await Bun.write("./binary.bin", new Uint8Array([0, 1, 2, 3]));

// Write with Response/Blob
await Bun.write("./fetched.html", await fetch("https://example.com"));

// Glob
const glob = new Bun.Glob("**/*.ts");
for await (const path of glob.scan({ cwd: "./src" })) {
  console.log(path);
}
```

## SQLite (Built-in)

```typescript
import { Database } from "bun:sqlite";

// Open database (creates if not exists)
const db = new Database("app.db");

// WAL mode for better concurrent performance
db.run("PRAGMA journal_mode = WAL");
db.run("PRAGMA foreign_keys = ON");

// Create tables
db.run(`
  CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    created_at TEXT DEFAULT (datetime('now'))
  )
`);

// Prepared statements (reusable, faster)
const insertUser = db.prepare(
  "INSERT INTO users (id, name, email) VALUES ($id, $name, $email)"
);

const getUser = db.prepare("SELECT * FROM users WHERE id = $id");
const getUserByEmail = db.prepare("SELECT * FROM users WHERE email = $email");
const listUsers = db.prepare("SELECT * FROM users ORDER BY created_at DESC LIMIT $limit");

// Execute
insertUser.run({
  $id: crypto.randomUUID(),
  $name: "Alice",
  $email: "alice@example.com",
});

const user = getUser.get({ $id: "some-id" }); // Single row
const users = listUsers.all({ $limit: 20 }); // All rows

// Transactions
const createUserWithProfile = db.transaction((user, profile) => {
  insertUser.run(user);
  db.prepare("INSERT INTO profiles (user_id, bio) VALUES ($userId, $bio)")
    .run({ $userId: user.$id, $bio: profile.bio });
});

createUserWithProfile(
  { $id: "123", $name: "Bob", $email: "bob@example.com" },
  { bio: "Hello world" }
);

// Close when done
db.close();
```

## Bundler

```typescript
// Bun.build() — fast JavaScript/TypeScript bundler
const result = await Bun.build({
  entrypoints: ["./src/index.ts"],
  outdir: "./dist",
  target: "browser", // "browser" | "bun" | "node"
  format: "esm",     // "esm" | "cjs" | "iife"
  minify: true,
  sourcemap: "external",
  splitting: true,    // Code splitting
  naming: "[dir]/[name]-[hash].[ext]",

  // Define compile-time constants
  define: {
    "process.env.NODE_ENV": JSON.stringify("production"),
    "API_URL": JSON.stringify("https://api.example.com"),
  },

  // External packages (don't bundle)
  external: ["react", "react-dom"],

  // Plugins
  plugins: [
    {
      name: "css-loader",
      setup(build) {
        build.onLoad({ filter: /\.css$/ }, async (args) => {
          const text = await Bun.file(args.path).text();
          return {
            contents: `export default ${JSON.stringify(text)}`,
            loader: "js",
          };
        });
      },
    },
  ],
});

if (!result.success) {
  console.error("Build failed:");
  for (const msg of result.logs) {
    console.error(msg);
  }
  process.exit(1);
}

for (const output of result.outputs) {
  console.log(`${output.path} — ${output.size} bytes`);
}
```

## Test Runner

```typescript
// test/user.test.ts
import { describe, it, expect, beforeAll, afterAll, mock, spyOn } from "bun:test";

describe("UserService", () => {
  let db: Database;

  beforeAll(() => {
    db = new Database(":memory:");
    db.run("CREATE TABLE users (id TEXT, name TEXT, email TEXT)");
  });

  afterAll(() => {
    db.close();
  });

  it("creates a user", () => {
    const user = { id: "1", name: "Alice", email: "alice@test.com" };
    db.prepare("INSERT INTO users VALUES ($id, $name, $email)").run({
      $id: user.id, $name: user.name, $email: user.email,
    });
    const found = db.prepare("SELECT * FROM users WHERE id = $id").get({ $id: "1" });
    expect(found).toEqual(user);
  });
});

// Mocking
describe("API Client", () => {
  it("fetches users", async () => {
    const mockFetch = mock(() =>
      Promise.resolve(Response.json([{ id: "1", name: "Alice" }]))
    );
    globalThis.fetch = mockFetch;

    const users = await fetchUsers();
    expect(mockFetch).toHaveBeenCalledTimes(1);
    expect(users).toHaveLength(1);
  });

  it("spies on methods", () => {
    const obj = { greet: (name: string) => `Hello, ${name}` };
    const spy = spyOn(obj, "greet");

    obj.greet("World");
    expect(spy).toHaveBeenCalledWith("World");
    expect(spy.mock.results[0].value).toBe("Hello, World");
  });
});

// Run: bun test
// Run specific: bun test test/user.test.ts
// Watch mode: bun test --watch
// Coverage: bun test --coverage
```

## Package Manager

```bash
# Install — fastest npm-compatible package manager
bun install                    # Install all dependencies
bun add express zod           # Add dependencies
bun add -d typescript @types/express  # Dev dependencies
bun add -g serve              # Global install
bun remove express            # Remove package

# Lockfile: bun.lockb (binary, faster than JSON)
# Symlinks node_modules (uses global cache — saves disk)

# Scripts
bun run dev                   # Run package.json scripts
bun run build
bunx create-vite my-app       # npx equivalent

# Workspace support (monorepo)
# package.json:
# { "workspaces": ["packages/*"] }
bun install                   # Installs all workspace deps
```

## Shell (Subprocess)

```typescript
import { $ } from "bun";

// Simple command
const output = await $`echo "Hello from Bun"`.text();

// With variables (auto-escaped)
const filename = "my file.txt";
await $`ls -la ${filename}`;

// Pipe commands
const result = await $`cat package.json | grep "name"`.text();

// Check exit code
const { exitCode } = await $`npm test`.nothrow();
if (exitCode !== 0) {
  console.error("Tests failed");
}

// Redirect output
await $`echo "log entry"`.appendTo("app.log");

// Low-level: Bun.spawn
const proc = Bun.spawn(["git", "status"], {
  cwd: "/path/to/repo",
  stdout: "pipe",
  stderr: "pipe",
});

const stdout = await new Response(proc.stdout).text();
const exitCode2 = await proc.exited;
```

## Node.js Compatibility

```typescript
// Bun supports most Node.js APIs natively
import fs from "node:fs/promises";
import path from "node:path";
import { createServer } from "node:http";
import { EventEmitter } from "node:events";
import crypto from "node:crypto";

// process works
console.log(process.env.NODE_ENV);
console.log(process.cwd());
console.log(process.argv);

// require() works (CJS compatibility)
const pkg = require("./package.json");

// __dirname and __filename work
console.log(__dirname);
console.log(__filename);

// Most popular npm packages work unchanged:
// express, fastify, hono, koa
// prisma, drizzle-orm, knex
// zod, joi, yup
// ws, socket.io
```

## Hono with Bun (Recommended Web Framework)

```typescript
import { Hono } from "hono";
import { cors } from "hono/cors";
import { logger } from "hono/logger";
import { jwt } from "hono/jwt";
import { zValidator } from "@hono/zod-validator";
import { z } from "zod";

const app = new Hono();

// Middleware
app.use("*", logger());
app.use("/api/*", cors());
app.use("/api/protected/*", jwt({ secret: Bun.env.JWT_SECRET! }));

// Validated routes
const createUserSchema = z.object({
  name: z.string().min(1).max(100),
  email: z.string().email(),
});

app.post("/api/users", zValidator("json", createUserSchema), async (c) => {
  const { name, email } = c.req.valid("json");
  const user = { id: crypto.randomUUID(), name, email };
  return c.json(user, 201);
});

app.get("/api/users/:id", async (c) => {
  const id = c.req.param("id");
  return c.json({ id, name: "Example" });
});

// Protected route
app.get("/api/protected/me", (c) => {
  const payload = c.get("jwtPayload");
  return c.json({ userId: payload.sub });
});

// Export for Bun.serve
export default {
  port: 3000,
  fetch: app.fetch,
};
```

## Performance Patterns

```typescript
// Bun.password — fast password hashing (uses argon2)
const hash = await Bun.password.hash("my-password", {
  algorithm: "argon2id",
  memoryCost: 65536,
  timeCost: 2,
});

const isValid = await Bun.password.verify("my-password", hash);

// Bun.CryptoHasher — fast hashing
const hasher = new Bun.CryptoHasher("sha256");
hasher.update("data to hash");
const digest = hasher.digest("hex");

// Bun.gzipSync / Bun.gunzipSync — fast compression
const compressed = Bun.gzipSync(Buffer.from("hello world"));
const decompressed = Bun.gunzipSync(compressed);

// Bun.Transpiler — fast TypeScript/JSX transpilation
const transpiler = new Bun.Transpiler({
  loader: "tsx",
  target: "browser",
});
const js = transpiler.transformSync(`
  const App: React.FC = () => <div>Hello</div>;
  export default App;
`);
```

## Gotchas

1. **Not all Node.js APIs are implemented** — `vm`, `worker_threads` (partial), `dgram`, some `cluster` features are missing or incomplete. Check bun.sh/docs/runtime/nodejs-apis for compatibility.

2. **C++ native addons don't work** — packages using node-gyp (bcrypt, canvas, sharp, sqlite3) need Bun-native alternatives. Use `bun:sqlite` instead of `better-sqlite3`. Use `Bun.password` instead of bcrypt.

3. **bun.lockb is binary** — not human-readable or diffable. Use `bun install --yarn` to generate a `yarn.lock` alongside it if you need readable lockfiles for code review.

4. **Hot reloading vs watch** — `bun --watch` restarts the process. `bun --hot` does HMR without restarting. Use `--hot` for servers (preserves state), `--watch` for scripts.

5. **Bun.serve() vs Express** — Bun.serve() is 5-10x faster than Express on Bun. If you're using Bun for speed, don't use Express — use Hono or Bun.serve() directly.

6. **Production readiness** — Bun is newer than Node.js and Deno. For mission-critical production services, have a Node.js fallback plan. Test edge cases thoroughly.
