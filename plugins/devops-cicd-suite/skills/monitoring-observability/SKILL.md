---
name: monitoring-observability
description: >
  Monitoring and observability — metrics collection, structured logging,
  distributed tracing, alerting, SLOs/SLIs, error tracking, and dashboard
  design for Node.js applications.
  Triggers: "monitoring", "observability", "logging", "metrics", "tracing",
  "alerting", "SLO", "SLI", "error tracking", "structured logging",
  "prometheus", "grafana", "datadog".
  NOT for: CI/CD pipelines (use github-actions), deployment (use deployment-strategies).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Monitoring & Observability

## The Three Pillars

```
┌──────────┐    ┌──────────┐    ┌──────────┐
│ Metrics  │    │   Logs   │    │  Traces  │
│ (numbers)│    │  (events)│    │  (flows) │
├──────────┤    ├──────────┤    ├──────────┤
│ Request  │    │ Structured│   │ Span A   │
│ rate: 42 │    │ JSON with │   │  ├─ Span B│
│ p99: 120 │    │ context   │   │  └─ Span C│
│ errors: 2│    │           │   │           │
└──────────┘    └──────────┘    └──────────┘
     │               │               │
     └───────────────┴───────────────┘
                     │
              ┌──────▼──────┐
              │  Dashboards │
              │   Alerts    │
              └─────────────┘
```

## Structured Logging

```typescript
// src/lib/logger.ts
import pino from "pino";

export const logger = pino({
  level: process.env.LOG_LEVEL || "info",
  ...(process.env.NODE_ENV === "production"
    ? {} // JSON output in production
    : { transport: { target: "pino-pretty" } }), // Pretty print in dev

  // Base fields on every log line
  base: {
    service: "myapp",
    version: process.env.APP_VERSION,
    environment: process.env.NODE_ENV,
  },

  // Redact sensitive fields
  redact: {
    paths: ["req.headers.authorization", "req.headers.cookie", "*.password", "*.token"],
    casing: "loose",
  },
});

// Child logger with request context
export function requestLogger(req: Request) {
  return logger.child({
    requestId: req.headers["x-request-id"] || crypto.randomUUID(),
    userId: req.user?.id,
    method: req.method,
    url: req.url,
  });
}
```

```typescript
// Usage patterns
const log = requestLogger(req);

// Good: structured data, not string interpolation
log.info({ userId, action: "login" }, "User logged in");
log.error({ err, orderId }, "Payment processing failed");
log.warn({ queueDepth: 1500, threshold: 1000 }, "Queue depth above threshold");

// Bad: string-only logs lose structure
log.info(`User ${userId} logged in`);  // Can't filter/aggregate
```

```typescript
// Express request logging middleware
import pinoHttp from "pino-http";

app.use(pinoHttp({
  logger,
  autoLogging: {
    ignore: (req) => req.url === "/health", // Don't log health checks
  },
  customSuccessMessage: (req, res) =>
    `${req.method} ${req.url} ${res.statusCode}`,
  customErrorMessage: (req, res) =>
    `${req.method} ${req.url} ${res.statusCode} failed`,
  customAttributeKeys: {
    req: "request",
    res: "response",
    err: "error",
    responseTime: "duration",
  },
}));
```

## Metrics with Prometheus

```typescript
// src/lib/metrics.ts
import { Registry, Counter, Histogram, Gauge, collectDefaultMetrics } from "prom-client";

const register = new Registry();

// Collect Node.js default metrics (memory, CPU, event loop)
collectDefaultMetrics({ register });

// Custom metrics
export const httpRequestDuration = new Histogram({
  name: "http_request_duration_seconds",
  help: "Duration of HTTP requests in seconds",
  labelNames: ["method", "route", "status_code"],
  buckets: [0.01, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10],
  registers: [register],
});

export const httpRequestTotal = new Counter({
  name: "http_requests_total",
  help: "Total number of HTTP requests",
  labelNames: ["method", "route", "status_code"],
  registers: [register],
});

export const activeConnections = new Gauge({
  name: "active_connections",
  help: "Number of active connections",
  registers: [register],
});

export const dbQueryDuration = new Histogram({
  name: "db_query_duration_seconds",
  help: "Duration of database queries in seconds",
  labelNames: ["operation", "table"],
  buckets: [0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1],
  registers: [register],
});

// Metrics endpoint
app.get("/metrics", async (req, res) => {
  res.set("Content-Type", register.contentType);
  res.send(await register.metrics());
});
```

```typescript
// Middleware to record request metrics
app.use((req, res, next) => {
  const start = process.hrtime.bigint();
  activeConnections.inc();

  res.on("finish", () => {
    const duration = Number(process.hrtime.bigint() - start) / 1e9;
    const route = req.route?.path || req.path;
    const labels = {
      method: req.method,
      route,
      status_code: String(res.statusCode),
    };
    httpRequestDuration.observe(labels, duration);
    httpRequestTotal.inc(labels);
    activeConnections.dec();
  });

  next();
});
```

## Error Tracking

```typescript
// src/lib/errors.ts

// Application error classes
class AppError extends Error {
  constructor(
    message: string,
    public statusCode: number = 500,
    public code: string = "INTERNAL_ERROR",
    public isOperational: boolean = true
  ) {
    super(message);
    this.name = this.constructor.name;
    Error.captureStackTrace(this, this.constructor);
  }
}

class NotFoundError extends AppError {
  constructor(resource: string, id: string) {
    super(`${resource} not found: ${id}`, 404, "NOT_FOUND");
  }
}

class ValidationError extends AppError {
  constructor(message: string, public fields?: Record<string, string>) {
    super(message, 400, "VALIDATION_ERROR");
  }
}

// Global error handler
app.use((err: Error, req: Request, res: Response, next: NextFunction) => {
  const log = requestLogger(req);

  if (err instanceof AppError && err.isOperational) {
    // Expected errors (validation, not found, etc.)
    log.warn({ err, statusCode: err.statusCode }, err.message);
    return res.status(err.statusCode).json({
      error: { code: err.code, message: err.message },
    });
  }

  // Unexpected errors (bugs, crashes)
  log.error({ err }, "Unhandled error");

  // Don't expose internal errors to clients
  res.status(500).json({
    error: { code: "INTERNAL_ERROR", message: "Something went wrong" },
  });
});

// Catch unhandled rejections
process.on("unhandledRejection", (reason) => {
  logger.fatal({ err: reason }, "Unhandled promise rejection");
  // Give time to flush logs, then exit
  setTimeout(() => process.exit(1), 1000);
});

process.on("uncaughtException", (err) => {
  logger.fatal({ err }, "Uncaught exception");
  setTimeout(() => process.exit(1), 1000);
});
```

## SLOs and SLIs

```
SLI (Service Level Indicator): A metric that measures service quality
SLO (Service Level Objective): A target value for an SLI
SLA (Service Level Agreement): A contract with consequences for missing SLOs
```

| SLI | How to Measure | Typical SLO |
|-----|---------------|-------------|
| Availability | `successful_requests / total_requests` | 99.9% (8.7h downtime/yr) |
| Latency (p50) | 50th percentile response time | < 100ms |
| Latency (p99) | 99th percentile response time | < 500ms |
| Error rate | `error_responses / total_responses` | < 0.1% |
| Throughput | Requests per second | > 1000 rps |
| Saturation | CPU/memory/disk utilization | < 80% |

```typescript
// SLO tracking with Prometheus
const sloAvailability = new Gauge({
  name: "slo_availability_ratio",
  help: "Current availability SLO ratio (target: 0.999)",
  registers: [register],
});

// Calculate periodically
setInterval(async () => {
  const total = await httpRequestTotal.get();
  const errors = total.values.filter(v => v.labels.status_code?.startsWith("5"));
  const totalCount = total.values.reduce((sum, v) => sum + v.value, 0);
  const errorCount = errors.reduce((sum, v) => sum + v.value, 0);

  if (totalCount > 0) {
    sloAvailability.set(1 - errorCount / totalCount);
  }
}, 60000);
```

## Alerting Rules

```yaml
# Prometheus alerting rules (alerts.yml)
groups:
  - name: application
    rules:
      # High error rate
      - alert: HighErrorRate
        expr: |
          sum(rate(http_requests_total{status_code=~"5.."}[5m]))
          / sum(rate(http_requests_total[5m])) > 0.01
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "Error rate above 1% for 5 minutes"
          description: "Current error rate: {{ $value | humanizePercentage }}"

      # High latency
      - alert: HighLatency
        expr: |
          histogram_quantile(0.99, rate(http_request_duration_seconds_bucket[5m])) > 1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "p99 latency above 1 second"

      # Service down
      - alert: ServiceDown
        expr: up == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Service {{ $labels.instance }} is down"

      # Memory usage
      - alert: HighMemoryUsage
        expr: |
          process_resident_memory_bytes / 1024 / 1024 > 512
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "Memory usage above 512MB for 10 minutes"
```

## Dashboard Design

```
Essential dashboards for a web application:

1. Overview Dashboard
   ├── Request rate (rps)
   ├── Error rate (%)
   ├── p50/p95/p99 latency
   ├── Active users
   └── Deployment markers

2. API Performance
   ├── Request duration by endpoint
   ├── Slowest endpoints (top 10)
   ├── Error rate by endpoint
   └── Request volume by endpoint

3. Infrastructure
   ├── CPU utilization
   ├── Memory usage
   ├── Disk I/O
   ├── Network traffic
   └── Container count

4. Database
   ├── Query duration (p50/p99)
   ├── Active connections
   ├── Connection pool utilization
   ├── Slow queries log
   └── Replication lag

5. Business Metrics
   ├── Signups per hour
   ├── Orders per hour
   ├── Revenue
   ├── Conversion funnel
   └── Feature adoption
```

## Gotchas

1. **Don't log everything in production.** Debug-level logs in production can generate GBs of data daily. Use `info` level by default, `debug` only when investigating issues. Structure your logs so you can grep/filter efficiently.

2. **High-cardinality labels kill Prometheus.** Labels like `userId`, `requestId`, or `URL path with IDs` create millions of time series. Use route patterns (`/users/:id`) not actual paths (`/users/12345`). Keep label cardinality under 1000 per metric.

3. **Percentiles can't be aggregated across instances.** You can't average p99 latencies from 10 servers to get the overall p99. Use histogram buckets and `histogram_quantile()` on the aggregated buckets instead.

4. **Alert fatigue is worse than no alerts.** If alerts fire too often, people ignore them. Every alert should be actionable. If you can't define what action to take when an alert fires, it shouldn't be an alert (make it a dashboard panel instead).

5. **Health check endpoints should check dependencies.** A `/health` that always returns 200 is useless. Check database connectivity, Redis connectivity, disk space. But don't make health checks too expensive (no full queries, just pings).

6. **Log rotation and retention.** Without rotation, logs fill disks. Without retention policies, storage costs grow forever. Set up logrotate for file-based logs, and configure retention in your log aggregation service (30 days for info, 90 days for errors is a common starting point).
