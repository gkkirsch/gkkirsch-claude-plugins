---
name: structured-logging
description: >
  Production logging patterns — structured JSON logging, log levels, context
  propagation, sensitive data filtering, and log aggregation integration.
  Triggers: "structured logging", "json logging", "log levels", "pino",
  "winston", "log aggregation", "logging best practices".
  NOT for: Metrics or distributed tracing (use distributed-tracing).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Structured Logging

## Setup with Pino (Recommended)

```typescript
// src/utils/logger.ts
import pino from "pino";

export const logger = pino({
  level: process.env.LOG_LEVEL || "info",
  formatters: {
    level: (label) => ({ level: label }), // "info" instead of 30
  },
  // Redact sensitive fields
  redact: {
    paths: ["password", "token", "authorization", "cookie", "*.password", "*.token"],
    censor: "[REDACTED]",
  },
  // Pretty print in development only
  transport: process.env.NODE_ENV !== "production"
    ? { target: "pino-pretty", options: { colorize: true } }
    : undefined,
});

// Child logger with persistent context
export function createServiceLogger(service: string) {
  return logger.child({ service });
}
```

### Output Format

```json
{"level":"info","time":1711324567890,"service":"api","requestId":"abc-123","method":"POST","path":"/api/users","statusCode":201,"duration":45,"msg":"request completed"}
{"level":"error","time":1711324567891,"service":"api","requestId":"def-456","err":{"type":"ValidationError","message":"Invalid email","stack":"..."},"msg":"request failed"}
```

## Log Levels

| Level | When | Example |
|-------|------|---------|
| `fatal` | App cannot continue | Database connection permanently lost |
| `error` | Operation failed, needs attention | Payment processing failed |
| `warn` | Something unexpected but handled | Rate limit approaching, retry succeeded |
| `info` | Normal operations worth recording | Request completed, user logged in |
| `debug` | Detailed diagnostic information | SQL query, cache hit/miss |
| `trace` | Very fine-grained (rarely used) | Function entry/exit, loop iterations |

```typescript
// Good logging examples
logger.info({ userId, action: "login" }, "User authenticated");
logger.warn({ retries: 3, service: "payment" }, "Service call succeeded after retries");
logger.error({ err, orderId, userId }, "Payment processing failed");

// BAD: unstructured string interpolation
logger.info(`User ${userId} logged in at ${new Date()}`);  // Not searchable

// GOOD: structured context object
logger.info({ userId, loginAt: new Date() }, "User logged in");  // Searchable
```

## Request Logging Middleware

```typescript
import { randomUUID } from "crypto";
import { pinoHttp } from "pino-http";

// Express middleware
export const requestLogger = pinoHttp({
  logger,
  genReqId: (req) => req.headers["x-request-id"] as string || randomUUID(),

  // Custom log fields
  customProps: (req) => ({
    requestId: req.id,
    userId: (req as any).user?.id,
  }),

  // What to log per request
  serializers: {
    req: (req) => ({
      method: req.method,
      url: req.url,
      query: req.query,
      userAgent: req.headers["user-agent"],
    }),
    res: (res) => ({
      statusCode: res.statusCode,
    }),
  },

  // Custom success/error messages
  customSuccessMessage: (req, res) =>
    `${req.method} ${req.url} ${res.statusCode}`,
  customErrorMessage: (req, res, err) =>
    `${req.method} ${req.url} ${res.statusCode} - ${err.message}`,

  // Don't log health checks
  autoLogging: {
    ignore: (req) => req.url === "/health",
  },
});

// Usage
app.use(requestLogger);
```

## Context Propagation

```typescript
import { AsyncLocalStorage } from "async_hooks";

interface RequestContext {
  requestId: string;
  userId?: string;
  traceId?: string;
  spanId?: string;
}

const contextStore = new AsyncLocalStorage<RequestContext>();

// Middleware to set context
function contextMiddleware(req: Request, _res: Response, next: NextFunction) {
  const context: RequestContext = {
    requestId: req.headers["x-request-id"] as string || randomUUID(),
    userId: (req as any).user?.id,
    traceId: req.headers["x-trace-id"] as string || randomUUID(),
    spanId: randomUUID(),
  };

  contextStore.run(context, () => next());
}

// Logger that auto-includes context
export function getLogger() {
  const ctx = contextStore.getStore();
  return ctx ? logger.child(ctx) : logger;
}

// Usage anywhere in the request lifecycle
function processOrder(orderId: string) {
  const log = getLogger(); // Automatically has requestId, userId, traceId
  log.info({ orderId }, "Processing order");
}
```

## Sensitive Data Filtering

```typescript
// Pino redaction (configured at creation)
const logger = pino({
  redact: {
    paths: [
      "password", "newPassword", "oldPassword",
      "token", "accessToken", "refreshToken",
      "authorization", "cookie",
      "creditCard", "ssn", "dob",
      "req.headers.authorization",
      "req.headers.cookie",
      "*.password", "*.token", "*.secret",
    ],
    censor: "[REDACTED]",
  },
});

// Custom serializer for deeper control
const logger = pino({
  serializers: {
    user: (user: any) => ({
      id: user.id,
      email: user.email ? maskEmail(user.email) : undefined,
      role: user.role,
      // Exclude: password, tokens, sessions
    }),
  },
});

function maskEmail(email: string): string {
  const [local, domain] = email.split("@");
  return `${local[0]}***@${domain}`;
}
```

## Error Logging

```typescript
// Always log errors with full context
async function processPayment(orderId: string, amount: number) {
  const log = getLogger();
  try {
    log.info({ orderId, amount }, "Starting payment processing");
    const result = await stripe.charges.create({ amount, currency: "usd" });
    log.info({ orderId, chargeId: result.id }, "Payment successful");
    return result;
  } catch (error) {
    log.error(
      {
        err: error,           // Pino serializes Error objects (message, stack, type)
        orderId,
        amount,
        // Include context needed to reproduce/debug
      },
      "Payment failed"
    );
    throw error;
  }
}

// Pino error serializer extracts: type, message, stack, code, signal
// Custom error properties are also included
```

## Log Aggregation

```bash
# Ship logs to Elasticsearch (via Filebeat)
# filebeat.yml
filebeat.inputs:
  - type: container
    paths: ["/var/log/containers/*.log"]
    json.keys_under_root: true
    json.add_error_key: true

output.elasticsearch:
  hosts: ["elasticsearch:9200"]
  index: "app-logs-%{+yyyy.MM.dd}"

# Ship to Datadog
# Use pino-datadog-transport
```

```typescript
// Pino transport to multiple destinations
const logger = pino({
  transport: {
    targets: [
      { target: "pino-pretty", level: "debug", options: { colorize: true } },
      { target: "pino-datadog-transport", level: "info", options: { ddClientConf: { authMethods: { apiKeyAuth: process.env.DD_API_KEY } } } },
    ],
  },
});
```

## Gotchas

1. **Don't log request/response bodies in production** — Bodies can contain PII, passwords, tokens, and large payloads. Log only metadata (method, path, status, duration). If you must log bodies for debugging, use `debug` level and redact sensitive fields.

2. **String concatenation in log messages** — `logger.info("User " + userId + " created")` isn't searchable. Use structured context: `logger.info({ userId }, "User created")`. Structured fields can be filtered and aggregated in log tools.

3. **Logging inside loops** — `for (item of items) { logger.debug({ item }, "processing") }` can produce thousands of log lines. Log the batch: `logger.info({ count: items.length }, "Processing batch")`.

4. **Missing error context** — `logger.error(error)` alone is useless for debugging. Always include what operation failed, what input caused it, and what state the system was in. The goal is to reproduce the issue from the log alone.

5. **Log level too verbose in production** — `debug` and `trace` levels generate enormous volume and cost money in log aggregation. Use `info` in production, `debug` only when investigating specific issues (toggle via env var).

6. **Synchronous logging blocks the event loop** — `console.log` is synchronous and blocks. Pino uses async writing by default. Don't use `console.log` in production Node.js — it's a performance bottleneck under load.
