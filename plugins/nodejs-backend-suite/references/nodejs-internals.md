# Node.js Internals Reference

Deep reference for Node.js runtime internals, V8 engine, libuv, and system-level understanding.

---

## Event Loop Architecture

### Event Loop Phases (Detailed)

```
┌───────────────────────────────────────────┐
│                 Timers                     │
│  Execute callbacks from setTimeout and    │
│  setInterval. Checks timer list, runs     │
│  callbacks whose threshold has elapsed.   │
├───────────────────────────────────────────┤
│            Pending Callbacks               │
│  Execute I/O callbacks deferred to next   │
│  loop iteration (e.g., TCP errors).       │
├───────────────────────────────────────────┤
│             Idle, Prepare                  │
│  Internal use only. Prepare for polling.  │
├───────────────────────────────────────────┤
│                  Poll                      │
│  Retrieve new I/O events. Execute I/O     │
│  callbacks (almost all except close,      │
│  timers, and setImmediate). Will block    │
│  here when appropriate.                   │
├───────────────────────────────────────────┤
│                 Check                      │
│  Execute setImmediate() callbacks.        │
│  Always runs after poll phase completes.  │
├───────────────────────────────────────────┤
│             Close Callbacks                │
│  Execute close event callbacks like       │
│  socket.on('close', ...).                 │
└───────────────────────────────────────────┘

Between EVERY phase transition:
  → process.nextTick() queue is drained
  → Promise microtask queue is drained
  → queueMicrotask() queue is drained

Priority order within microtasks:
  1. process.nextTick (highest)
  2. Promise.then / queueMicrotask
```

### Timer Phase Details

```typescript
// Timer precision: setTimeout(fn, 0) actually fires after ~1ms minimum
// Timer coalescing: Multiple timers with similar deadlines may fire together
// Max timer delay: 2^31 - 1 ms (~24.8 days). Larger values fire immediately.

// Timer implementation:
// - Stored in a min-heap (priority queue) sorted by expiry time
// - Event loop checks the heap root on each timer phase
// - If root has expired, pop and execute callback

// setImmediate vs setTimeout(fn, 0):
// - Inside I/O callback: setImmediate ALWAYS fires first
// - Outside I/O callback: order is non-deterministic
// - setImmediate runs in the "check" phase (after poll)
// - setTimeout(fn, 0) runs in the "timer" phase (before poll)
```

### Poll Phase Behavior

```
When the event loop enters the poll phase:

1. If the poll queue is NOT empty:
   → Execute callbacks synchronously until queue is empty
     or system-dependent hard limit is reached

2. If the poll queue IS empty:
   a. If setImmediate() scripts are scheduled:
      → End poll phase, move to check phase
   b. If setImmediate() is NOT scheduled:
      → Wait for callbacks to be added to queue
      → Execute them immediately when they arrive
      → But also check if any timers have expired
         → If yes, wrap back to timers phase
```

---

## libuv Thread Pool

```
Default thread pool size: 4
Set via: UV_THREADPOOL_SIZE environment variable
Maximum: 1024 (compile-time limit)

Operations that use the thread pool:
- DNS lookups (dns.lookup, NOT dns.resolve*)
- File system operations (fs.*)
- crypto.pbkdf2, crypto.scrypt, crypto.randomBytes (large)
- zlib compression/decompression
- User-created worker threads do NOT use this pool

Operations that do NOT use the thread pool:
- Network I/O (TCP, UDP, HTTP) — uses OS async I/O (epoll/kqueue/IOCP)
- DNS resolve (dns.resolve*) — uses c-ares library
- Child process operations — OS-level process management
- Timers — managed by libuv event loop directly
```

### Thread Pool Saturation

```typescript
// If all 4 threads are busy with fs operations,
// new fs operations will QUEUE and wait.
// This causes unexpected latency.

// Diagnosis:
// 1. High event loop lag despite low CPU usage
// 2. File operations taking much longer than expected
// 3. dns.lookup calls are slow (they share the thread pool)

// Solutions:
// 1. Increase pool size: UV_THREADPOOL_SIZE=16
// 2. Use dns.resolve instead of dns.lookup
// 3. Use worker_threads for CPU-bound work (doesn't use UV pool)
// 4. Use streaming instead of bulk file operations
```

---

## V8 Engine Internals

### Memory Layout

```
V8 Heap Structure:
┌──────────────────────────────────────────┐
│           New Space (Young Gen)          │
│  ┌─────────────┐  ┌─────────────┐       │
│  │ Semi-space A │  │ Semi-space B │       │
│  │  (from)      │  │  (to)        │       │
│  └─────────────┘  └─────────────┘       │
│  Size: 1-8 MB each (--max-semi-space-size)│
│  GC: Scavenge (minor GC, very fast)     │
├──────────────────────────────────────────┤
│           Old Space (Old Gen)            │
│  ┌──────────────────────────────────┐    │
│  │   Old Pointer Space              │    │
│  │   (objects with pointers to      │    │
│  │    other objects)                 │    │
│  ├──────────────────────────────────┤    │
│  │   Old Data Space                 │    │
│  │   (objects with only data,       │    │
│  │    no pointers: strings, numbers)│    │
│  └──────────────────────────────────┘    │
│  GC: Mark-Sweep-Compact (major GC, slow)│
├──────────────────────────────────────────┤
│          Large Object Space              │
│  Objects larger than kMaxRegularHeapObj  │
│  (~256KB). Never moved by GC.           │
├──────────────────────────────────────────┤
│            Code Space                    │
│  JIT-compiled machine code              │
│  (TurboFan optimized code)              │
├──────────────────────────────────────────┤
│             Map Space                    │
│  Hidden classes / object shapes         │
│  (V8's internal type system)            │
└──────────────────────────────────────────┘

Total heap limit: --max-old-space-size (default: ~1.5GB on 64-bit)
```

### Garbage Collection

```
Minor GC (Scavenge):
- Operates on New Space only
- Very fast: 1-10ms typically
- Copies surviving objects from one semi-space to another
- Objects surviving 2 scavenges are promoted to Old Space
- Triggered when semi-space is full

Major GC (Mark-Sweep-Compact):
- Operates on Old Space
- Much slower: 50-200ms possible
- Three phases:
  1. Mark: Walk from roots, mark all reachable objects
  2. Sweep: Free unmarked (unreachable) objects
  3. Compact: Move surviving objects together (defragment)
- Triggered when Old Space grows beyond threshold

Incremental Marking:
- Major GC marking is split into small increments
- Interleaved with JavaScript execution
- Reduces individual pause times

Concurrent Sweeping/Compaction:
- Sweeping and compaction run on background threads
- Main thread can continue executing JavaScript
- V8 uses write barriers to track cross-generation pointers

Idle-Time GC:
- V8 uses idle time (when event loop is waiting)
- to perform GC work without pausing JavaScript
```

### Hidden Classes (Maps)

```typescript
// V8 uses hidden classes (internally called "Maps") for fast property access
// Objects with the same shape share the same hidden class

// GOOD — same shape, same hidden class:
function Point(x, y) {
  this.x = x;  // Transition to Map1
  this.y = y;  // Transition to Map2
}
const p1 = new Point(1, 2);
const p2 = new Point(3, 4);
// p1 and p2 share Map2 → fast property access

// BAD — different property order:
const a = {};
a.x = 1;
a.y = 2;

const b = {};
b.y = 1;  // Different order!
b.x = 2;
// a and b have DIFFERENT hidden classes → slower

// BAD — adding properties conditionally:
function createUser(name, email, admin) {
  const user = { name, email };
  if (admin) user.permissions = ['all']; // Conditional shape → megamorphic
  return user;
}

// GOOD — always same shape:
function createUser(name, email, admin) {
  return {
    name,
    email,
    permissions: admin ? ['all'] : null, // Always present
  };
}
```

### JIT Compilation Pipeline

```
Source Code
    │
    ▼
┌─────────────────┐
│    Ignition      │  Bytecode interpreter (fast startup)
│   (Interpreter)  │  Collects type feedback
└────────┬────────┘
         │ Hot code (called many times)
         ▼
┌─────────────────┐
│   TurboFan       │  Optimizing compiler
│  (JIT Compiler)  │  Produces optimized machine code
└────────┬────────┘
         │
         ▼
  Machine Code (fast execution)

Deoptimization:
- If type assumptions are wrong, TurboFan bails out
- Falls back to Ignition (interpreter)
- Expensive: avoid changing types of variables

Common deoptimization triggers:
- Changing object shape after creation
- Using arguments object in certain ways
- try/catch in hot functions (improved in newer V8)
- Mixing types in the same variable
- Deleting object properties (delete obj.key)
```

---

## process Module

### Memory Monitoring

```typescript
// process.memoryUsage()
const usage = process.memoryUsage();
// {
//   rss: 35651584,        // Resident Set Size: total memory allocated
//   heapTotal: 7708672,   // V8 heap total allocated
//   heapUsed: 5765728,    // V8 heap actually used
//   external: 1089739,    // Memory for C++ objects bound to JS
//   arrayBuffers: 11158,  // Memory for ArrayBuffers & SharedArrayBuffers
// }

// RSS includes:
// - V8 heap (heapTotal)
// - C++ objects (external)
// - Stack
// - Code segments
// - Memory-mapped files

// RSS > heapTotal because:
// - Stack memory
// - Native code (compiled JavaScript)
// - Native add-on memory
// - OS-level buffers
```

### CPU Usage

```typescript
// process.cpuUsage()
const startUsage = process.cpuUsage();
// ... some work ...
const endUsage = process.cpuUsage(startUsage);
// { user: 56789, system: 12345 }  // Microseconds
// user: time in user mode
// system: time in kernel mode

// High system time indicates:
// - Lots of I/O operations
// - Context switching
// - Network operations
```

### Resource Limits

```typescript
// Check/set limits
process.getActiveResourcesInfo();
// Returns array of active async resource types

// Report generation
process.report.getReport(); // JSON diagnostic report
process.report.writeReport(); // Write to file
// Includes: JavaScript stack, native stack, heap stats, system info, etc.
```

---

## Async Resource Tracking

### AsyncLocalStorage (Context Propagation)

```typescript
import { AsyncLocalStorage } from 'node:async_hooks';

interface RequestContext {
  requestId: string;
  userId?: string;
  startTime: number;
}

export const requestContext = new AsyncLocalStorage<RequestContext>();

// Middleware to set context
app.use((req, res, next) => {
  const context: RequestContext = {
    requestId: req.id,
    userId: req.user?.id,
    startTime: Date.now(),
  };

  requestContext.run(context, () => next());
});

// Access context anywhere without passing it
function getRequestId(): string {
  return requestContext.getStore()?.requestId ?? 'unknown';
}

// Logger with automatic request context
const logger = {
  info(msg: string, data?: object) {
    const ctx = requestContext.getStore();
    console.log(JSON.stringify({
      level: 'info',
      msg,
      requestId: ctx?.requestId,
      userId: ctx?.userId,
      ...data,
    }));
  },
};
```

### Async Hooks (Advanced Debugging)

```typescript
import { createHook, executionAsyncId, triggerAsyncId } from 'node:async_hooks';

// Track all async operations (expensive — for debugging only)
const activeResources = new Map<number, { type: string; stack: string }>();

const hook = createHook({
  init(asyncId, type, triggerAsyncId, resource) {
    activeResources.set(asyncId, {
      type,
      stack: new Error().stack!,
    });
  },
  destroy(asyncId) {
    activeResources.delete(asyncId);
  },
});

// Enable only during debugging
// hook.enable();

// Dump active resources (find leaks)
function dumpActiveResources() {
  for (const [id, { type, stack }] of activeResources) {
    console.log(`Async ID: ${id}, Type: ${type}`);
    console.log(stack);
  }
}
```

---

## Module System Internals

### ESM vs CJS

```
CommonJS (CJS):
- require() is synchronous
- Module is evaluated on first require()
- Cached in require.cache
- Circular dependencies: partial exports
- Can require() conditionally (dynamic)
- File extensions: .js (with "type": "commonjs"), .cjs

ESM (ECMAScript Modules):
- import is asynchronous (statically analyzed)
- Module is evaluated lazily
- Live bindings (not copies)
- Circular dependencies: may get TDZ errors
- Dynamic import: await import('./module.js')
- File extensions: .js (with "type": "module"), .mjs
- Top-level await supported
- import.meta.url replaces __filename
- import.meta.dirname replaces __dirname (Node 21+)
```

### Module Resolution

```
ESM Resolution Order:
1. Exact path match (with extension required)
2. Package.json "exports" field
3. Package.json "main" field
4. index.js in directory

CJS Resolution Order:
1. Exact file match
2. file.js
3. file.json
4. file.node
5. directory/package.json "main"
6. directory/index.js
7. directory/index.json
8. directory/index.node
9. node_modules (walk up directory tree)
```

---

## Networking Internals

### HTTP Keep-Alive

```typescript
// Node.js HTTP server keep-alive defaults:
server.keepAliveTimeout = 5000;   // Close idle keep-alive connections after 5s
server.headersTimeout = 60000;    // Timeout for receiving headers

// IMPORTANT for production behind load balancers:
// keepAliveTimeout MUST be greater than the load balancer's idle timeout
// AWS ALB idle timeout: 60s (default)
// So: keepAliveTimeout >= 61s

server.keepAliveTimeout = 65_000;  // 65s > ALB's 60s
server.headersTimeout = 66_000;    // Must be > keepAliveTimeout
```

### Connection Handling

```
TCP Connection Lifecycle in Node.js:

1. Client connects → server 'connection' event
2. HTTP request received → server 'request' event
3. Response sent
4. If keep-alive: connection stays open for next request
5. If no activity for keepAliveTimeout: server closes connection
6. socket.on('close') fires

Connection States:
- Established: Active TCP connection
- Half-open: One side closed (FIN sent/received)
- Time-wait: Post-close cleanup (2 * MSL)

Tracking active connections:
const connections = new Set();
server.on('connection', (socket) => {
  connections.add(socket);
  socket.on('close', () => connections.delete(socket));
});
// connections.size = current active connections
```

---

## Diagnostics and Debugging

### Diagnostic Channels (Node.js 20+)

```typescript
import diagnostics_channel from 'node:diagnostics_channel';

// Subscribe to built-in channels
const httpChannel = diagnostics_channel.channel('http.server.request.start');
httpChannel.subscribe((message) => {
  const { request, response, server, socket } = message as any;
  console.log(`HTTP ${request.method} ${request.url}`);
});

// Create custom diagnostic channels
const dbChannel = diagnostics_channel.channel('app:database:query');

// Publisher (in your database layer)
dbChannel.publish({
  query: sql,
  params,
  duration: elapsed,
});

// Subscriber (in your monitoring layer)
dbChannel.subscribe((msg) => {
  const { query, duration } = msg as any;
  if (duration > 100) {
    logger.warn({ query, duration }, 'Slow query');
  }
});
```

### Inspector Protocol

```typescript
// Programmatic debugging
import { Session } from 'node:inspector/promises';

const session = new Session();
session.connect();

// CPU profile
await session.post('Profiler.enable');
await session.post('Profiler.start');
// ... code runs ...
const { profile } = await session.post('Profiler.stop');

// Heap snapshot
await session.post('HeapProfiler.enable');
const chunks: string[] = [];
session.on('HeapProfiler.addHeapSnapshotChunk', (m) => {
  chunks.push(m.params.chunk);
});
await session.post('HeapProfiler.takeHeapSnapshot');
fs.writeFileSync('heap.heapsnapshot', chunks.join(''));
```

---

## Performance Counters

### perf_hooks Module

```typescript
import {
  performance,
  PerformanceObserver,
  monitorEventLoopDelay,
  createHistogram,
} from 'node:perf_hooks';

// 1. High-resolution timing
const start = performance.now(); // Microsecond precision
// ... work ...
const elapsed = performance.now() - start;

// 2. Performance marks and measures
performance.mark('db-query-start');
await db.query('SELECT ...');
performance.mark('db-query-end');
performance.measure('db-query', 'db-query-start', 'db-query-end');

// 3. Observe measurements
const obs = new PerformanceObserver((list) => {
  for (const entry of list.getEntries()) {
    logger.info({
      name: entry.name,
      duration: entry.duration,
      type: entry.entryType,
    }, 'Performance measure');
  }
});
obs.observe({ entryTypes: ['measure', 'function'] });

// 4. Event loop delay monitoring
const h = monitorEventLoopDelay({ resolution: 20 });
h.enable();

setInterval(() => {
  console.log({
    min: h.min / 1e6,
    max: h.max / 1e6,
    mean: h.mean / 1e6,
    p50: h.percentile(50) / 1e6,
    p99: h.percentile(99) / 1e6,
  });
  h.reset();
}, 5000);

// 5. Histogram (Node.js 19+)
const histogram = createHistogram();
// Record values in nanoseconds
histogram.record(1000);
histogram.record(2000);
console.log(histogram.percentile(99));
```

---

## Node.js Version Features

### Node.js 20 LTS

```
- Stable test runner (node:test)
- Permission model (--experimental-permission)
- Stable single executable applications
- V8 11.3 (improved performance)
- import.meta.resolve() stable
- Custom ESM loader hooks (--loader flag)
```

### Node.js 22 LTS

```
- require(ESM) support (loading ESM from CJS)
- WebSocket client (stable)
- V8 12.4
- Glob support in node:fs
- watch mode (node --watch) stable
- Strip types (--experimental-strip-types)
- SQLite module (node:sqlite)
```

### Node.js 23+

```
- import.meta.dirname and import.meta.filename
- Unflagged --experimental-strip-types
- Improved WebSocket support
- Enhanced SQLite integration
- Type stripping for TypeScript (run .ts files directly)
```

---

## CLI Flags Reference

```bash
# Memory
--max-old-space-size=4096    # Max heap size in MB
--max-semi-space-size=64     # Max young generation semi-space in MB
--initial-heap-size=512      # Initial heap size in MB

# GC
--expose-gc                   # Enable global.gc()
--trace-gc                    # Log GC events to stderr
--trace-gc-verbose            # Detailed GC logging
--gc-interval=100             # Force GC every N allocations

# Profiling
--prof                        # Generate V8 tick processor profile
--prof-process                # Process --prof output
--inspect                     # Enable Chrome DevTools debugging
--inspect-brk                 # Same but break on first line
--inspect=0.0.0.0:9229        # Listen on all interfaces

# Diagnostics
--report-on-fatalerror        # Generate diagnostic report on crash
--report-on-signal=SIGUSR2    # Generate report on signal
--heapsnapshot-signal=SIGUSR2 # Heap snapshot on signal
--diagnostic-dir=./diag       # Output directory for diagnostics
--trace-warnings              # Show stack traces for process warnings
--trace-deprecation            # Show stack traces for deprecations

# Module System
--experimental-strip-types    # Run TypeScript directly (Node 22+)
--conditions=development      # Set package.json export conditions
--enable-source-maps          # Enable source map support

# Performance
--v8-pool-size=0              # Disable V8 background threads (debugging)
--zero-fill-buffers           # Zero-fill all new Buffer allocations

# Security
--experimental-permission     # Enable permission model
--allow-fs-read=/path         # Allow file system reads
--allow-fs-write=/path        # Allow file system writes
```
