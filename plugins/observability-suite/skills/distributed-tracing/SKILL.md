---
name: distributed-tracing
description: >
  Distributed tracing and metrics — OpenTelemetry setup, spans, context
  propagation, custom metrics, dashboards, and alerting patterns.
  Triggers: "distributed tracing", "opentelemetry", "otel", "spans",
  "metrics", "prometheus", "grafana", "apm", "observability".
  NOT for: Logging patterns (use structured-logging).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Distributed Tracing & Metrics

## OpenTelemetry Setup

```typescript
// src/instrumentation.ts — MUST be imported before all other imports
import { NodeSDK } from "@opentelemetry/sdk-node";
import { getNodeAutoInstrumentations } from "@opentelemetry/auto-instrumentations-node";
import { OTLPTraceExporter } from "@opentelemetry/exporter-trace-otlp-http";
import { OTLPMetricExporter } from "@opentelemetry/exporter-metrics-otlp-http";
import { PeriodicExportingMetricReader } from "@opentelemetry/sdk-metrics";
import { Resource } from "@opentelemetry/resources";
import { ATTR_SERVICE_NAME, ATTR_SERVICE_VERSION } from "@opentelemetry/semantic-conventions";

const sdk = new NodeSDK({
  resource: new Resource({
    [ATTR_SERVICE_NAME]: process.env.SERVICE_NAME || "api",
    [ATTR_SERVICE_VERSION]: process.env.SERVICE_VERSION || "1.0.0",
    "deployment.environment": process.env.NODE_ENV || "development",
  }),

  traceExporter: new OTLPTraceExporter({
    url: process.env.OTEL_EXPORTER_OTLP_ENDPOINT || "http://localhost:4318/v1/traces",
  }),

  metricReader: new PeriodicExportingMetricReader({
    exporter: new OTLPMetricExporter({
      url: process.env.OTEL_EXPORTER_OTLP_ENDPOINT || "http://localhost:4318/v1/metrics",
    }),
    exportIntervalMillis: 15_000,
  }),

  instrumentations: [
    getNodeAutoInstrumentations({
      "@opentelemetry/instrumentation-http": {
        ignoreIncomingPaths: ["/health", "/metrics"],
      },
      "@opentelemetry/instrumentation-express": {},
      "@opentelemetry/instrumentation-pg": {},
      "@opentelemetry/instrumentation-redis-4": {},
    }),
  ],
});

sdk.start();

process.on("SIGTERM", () => sdk.shutdown());
```

```bash
# Install OpenTelemetry packages
npm install @opentelemetry/sdk-node \
  @opentelemetry/auto-instrumentations-node \
  @opentelemetry/exporter-trace-otlp-http \
  @opentelemetry/exporter-metrics-otlp-http \
  @opentelemetry/sdk-metrics \
  @opentelemetry/api
```

```typescript
// src/index.ts — import instrumentation FIRST
import "./instrumentation";  // Must be first!
import app from "./app";
```

## Custom Spans

```typescript
import { trace, SpanStatusCode, context } from "@opentelemetry/api";

const tracer = trace.getTracer("order-service");

async function processOrder(orderId: string, items: OrderItem[]) {
  // Create a span for this operation
  return tracer.startActiveSpan("processOrder", async (span) => {
    try {
      span.setAttribute("order.id", orderId);
      span.setAttribute("order.item_count", items.length);

      // Nested span for payment
      const total = await tracer.startActiveSpan("chargePayment", async (paymentSpan) => {
        paymentSpan.setAttribute("payment.amount", calculateTotal(items));
        const result = await paymentService.charge(orderId, calculateTotal(items));
        paymentSpan.setAttribute("payment.transaction_id", result.transactionId);
        paymentSpan.end();
        return result.amount;
      });

      // Nested span for inventory
      await tracer.startActiveSpan("reserveInventory", async (inventorySpan) => {
        for (const item of items) {
          inventorySpan.addEvent("reserving_item", { "item.sku": item.sku, "item.quantity": item.quantity });
          await inventoryService.reserve(item.sku, item.quantity);
        }
        inventorySpan.end();
      });

      span.setAttribute("order.total", total);
      span.setStatus({ code: SpanStatusCode.OK });
      return { orderId, total, status: "processed" };
    } catch (error) {
      span.setStatus({ code: SpanStatusCode.ERROR, message: (error as Error).message });
      span.recordException(error as Error);
      throw error;
    } finally {
      span.end();
    }
  });
}
```

## Custom Metrics

```typescript
import { metrics } from "@opentelemetry/api";

const meter = metrics.getMeter("order-service");

// Counter — monotonically increasing value
const orderCounter = meter.createCounter("orders.created", {
  description: "Total orders created",
  unit: "orders",
});

// Histogram — distribution of values (latency, size)
const orderLatency = meter.createHistogram("orders.processing_duration", {
  description: "Order processing duration",
  unit: "ms",
});

// Gauge — point-in-time value
const activeConnections = meter.createObservableGauge("db.connections.active", {
  description: "Active database connections",
});
activeConnections.addCallback((result) => {
  result.observe(pool.totalCount);
});

// Usage in code
async function createOrder(data: OrderData) {
  const start = Date.now();

  try {
    const order = await orderService.create(data);
    orderCounter.add(1, {
      "order.type": data.type,
      "order.region": data.region,
    });
    return order;
  } finally {
    orderLatency.record(Date.now() - start, {
      "order.type": data.type,
    });
  }
}
```

## Prometheus Metrics (Alternative)

```typescript
import promClient from "prom-client";

// Default metrics (memory, CPU, event loop)
promClient.collectDefaultMetrics();

// Custom metrics
const httpDuration = new promClient.Histogram({
  name: "http_request_duration_seconds",
  help: "HTTP request duration in seconds",
  labelNames: ["method", "route", "status"],
  buckets: [0.01, 0.05, 0.1, 0.5, 1, 2, 5],
});

const activeRequests = new promClient.Gauge({
  name: "http_active_requests",
  help: "Number of active HTTP requests",
});

// Middleware
app.use((req, res, next) => {
  activeRequests.inc();
  const end = httpDuration.startTimer();

  res.on("finish", () => {
    activeRequests.dec();
    end({ method: req.method, route: req.route?.path || req.path, status: res.statusCode });
  });

  next();
});

// Metrics endpoint for Prometheus scraping
app.get("/metrics", async (_, res) => {
  res.set("Content-Type", promClient.register.contentType);
  res.send(await promClient.register.metrics());
});
```

## Context Propagation Across Services

```typescript
import { context, propagation, trace } from "@opentelemetry/api";

// Outbound HTTP — inject trace context into headers
async function callService(url: string, data: unknown) {
  const headers: Record<string, string> = { "Content-Type": "application/json" };

  // Inject W3C traceparent header automatically
  propagation.inject(context.active(), headers);

  return fetch(url, {
    method: "POST",
    headers,
    body: JSON.stringify(data),
  });
}

// Message queue — inject into message headers
function publishEvent(exchange: string, event: object) {
  const carrier: Record<string, string> = {};
  propagation.inject(context.active(), carrier);

  channel.publish(exchange, "", Buffer.from(JSON.stringify(event)), {
    headers: carrier, // traceparent, tracestate
  });
}

// Message consumer — extract context
channel.consume(queue, (msg) => {
  const parentContext = propagation.extract(context.active(), msg.properties.headers);
  context.with(parentContext, async () => {
    // This span is connected to the publisher's trace
    await tracer.startActiveSpan("processMessage", async (span) => {
      await handleMessage(JSON.parse(msg.content.toString()));
      span.end();
    });
  });
});
```

## Alerting Rules

```yaml
# Prometheus alerting rules
groups:
  - name: service-alerts
    rules:
      # High error rate
      - alert: HighErrorRate
        expr: rate(http_request_duration_seconds_count{status=~"5.."}[5m]) / rate(http_request_duration_seconds_count[5m]) > 0.05
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "High error rate (> 5%) on {{ $labels.route }}"

      # Slow responses
      - alert: SlowResponses
        expr: histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 2
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "P95 latency > 2s on {{ $labels.route }}"

      # Service down
      - alert: ServiceDown
        expr: up == 0
        for: 1m
        labels:
          severity: critical
```

## The Four Golden Signals

| Signal | What to Measure | Metric Type |
|--------|----------------|-------------|
| **Latency** | Request duration (p50, p95, p99) | Histogram |
| **Traffic** | Requests per second | Counter |
| **Errors** | Error rate (5xx / total) | Counter ratio |
| **Saturation** | Resource utilization (CPU, memory, connections) | Gauge |

## Gotchas

1. **Import instrumentation FIRST** — OpenTelemetry must be imported before any other module to monkey-patch HTTP, Express, and database libraries. If you import `express` before `instrumentation.ts`, those requests won't be traced.

2. **High-cardinality labels kill Prometheus** — Don't use user IDs, request IDs, or URLs as metric labels. Each unique label combination creates a new time series. Use bounded values: HTTP method, status code class (2xx/4xx/5xx), route pattern.

3. **Sampling is essential at scale** — Tracing every request at 10K RPS generates terabytes of data. Use head sampling (1% of requests) or tail sampling (100% of errors, 1% of successes). Configure in the OpenTelemetry collector.

4. **Missing span.end() causes memory leaks** — If you create a span but forget to call `span.end()` (e.g., early return in try/catch), the span stays in memory. Always use try/finally or `startActiveSpan` which auto-manages lifecycle.

5. **Clock skew breaks distributed traces** — If service A's clock is 5 seconds ahead of service B, trace timelines look wrong (child spans starting before parents). Use NTP time sync on all servers.

6. **Don't alert on symptoms AND causes** — If "database slow" fires alongside "API latency high", you get two pages for one problem. Alert on the cause (DB latency), not every downstream symptom.
