# OpenTelemetry Reference Guide

## Overview

OpenTelemetry (OTel) is the industry-standard framework for collecting telemetry data — traces, metrics, and logs — from applications and infrastructure. It provides vendor-neutral APIs, SDKs, and tools for instrumentation, collection, and export.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Application Code                          │
│                                                              │
│  ┌──────────┐   ┌──────────┐   ┌──────────┐                │
│  │  Traces   │   │  Metrics  │   │   Logs   │                │
│  │   API     │   │   API     │   │   API    │                │
│  └────┬─────┘   └────┬─────┘   └────┬─────┘                │
│       │              │              │                        │
│  ┌────┴──────────────┴──────────────┴─────┐                 │
│  │         OpenTelemetry SDK               │                 │
│  │  ┌───────────┐  ┌───────────┐          │                 │
│  │  │  Samplers  │  │ Processors │          │                 │
│  │  └───────────┘  └───────────┘          │                 │
│  │  ┌───────────────────────────┐          │                 │
│  │  │       Exporters           │          │                 │
│  │  │  OTLP | Jaeger | Zipkin   │          │                 │
│  │  └─────────────┬─────────────┘          │                 │
│  └────────────────┼────────────────────────┘                │
└───────────────────┼──────────────────────────────────────────┘
                    │
                    ▼
┌───────────────────────────────────────┐
│       OpenTelemetry Collector          │
│  ┌──────────┐ ┌──────────┐ ┌────────┐│
│  │ Receivers │→│Processors│→│Exporters││
│  │ OTLP     │ │ Batch    │ │ Jaeger ││
│  │ Jaeger   │ │ Filter   │ │ Tempo  ││
│  │ Zipkin   │ │ Sampling │ │ Datadog││
│  │ Kafka    │ │ Attributes│ │ OTLP  ││
│  └──────────┘ └──────────┘ └────────┘│
└───────────────────────────────────────┘
                    │
         ┌──────────┴──────────┐
         ▼                     ▼
┌──────────────┐    ┌──────────────┐
│    Jaeger    │    │ Grafana Tempo│
│   (Traces)   │    │   (Traces)   │
└──────────────┘    └──────────────┘
```

## SDK Components

### TracerProvider
The entry point for the tracing API. Configures:
- **Resource**: Metadata about the entity producing telemetry (service name, version)
- **Sampler**: Decides which traces to keep
- **SpanProcessor**: Processes spans before export (batching, filtering)
- **Exporter**: Sends spans to a backend

### MeterProvider
The entry point for the metrics API. Configures:
- **Resource**: Same as TracerProvider
- **MetricReader**: Collects metrics on a schedule
- **MetricExporter**: Sends metrics to a backend
- **Views**: Customize metric output (rename, change aggregation, drop attributes)

### LoggerProvider
The entry point for the logs API. Configures:
- **Resource**: Same as TracerProvider
- **LogRecordProcessor**: Processes log records before export
- **LogRecordExporter**: Sends logs to a backend

## Signals

### Traces
Represent the journey of a request through a distributed system.

**Components:**
- **Trace**: A collection of spans forming a DAG (directed acyclic graph)
- **Span**: A single operation within a trace
- **SpanContext**: Carries trace_id, span_id, and trace_flags across boundaries
- **Attributes**: Key-value metadata on spans
- **Events**: Timestamped annotations within a span
- **Links**: References to spans in other traces
- **Status**: OK, ERROR, or UNSET

**Span Kinds:**
| Kind | Description | Example |
|------|-------------|---------|
| SERVER | Handles incoming request | HTTP handler, gRPC service |
| CLIENT | Makes outgoing request | HTTP client, database query |
| PRODUCER | Sends message to queue | Kafka producer, SQS sender |
| CONSUMER | Receives message from queue | Kafka consumer, SQS processor |
| INTERNAL | Internal operation | Business logic, computation |

### Metrics
Numeric measurements collected over time.

**Instruments:**
| Instrument | Kind | Example |
|------------|------|---------|
| Counter | Monotonic, additive | request_count, bytes_sent |
| UpDownCounter | Non-monotonic, additive | active_connections, queue_depth |
| Histogram | Distribution | request_duration, response_size |
| Gauge | Non-additive | cpu_temperature, memory_usage |
| ObservableCounter | Async monotonic | total_page_faults |
| ObservableUpDownCounter | Async non-monotonic | thread_count |
| ObservableGauge | Async non-additive | cpu_utilization |

### Logs
Timestamped text or structured records.

**Log Record Fields:**
- Timestamp
- ObservedTimestamp
- SeverityNumber (1-24)
- SeverityText (TRACE, DEBUG, INFO, WARN, ERROR, FATAL)
- Body (string or structured)
- Attributes
- TraceId / SpanId (for correlation)
- Resource

## Context Propagation

### W3C Trace Context (Default)

```
traceparent: 00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01
             ^^  ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^  ^^^^^^^^^^^^^^^^  ^^
             │           trace-id (128-bit)          span-id (64-bit)  │
             version                                              trace-flags
                                                                  01 = sampled

tracestate: vendor1=opaque_value,vendor2=opaque_value
```

### B3 Propagation (Zipkin)

**Multi-header format:**
```
X-B3-TraceId: 4bf92f3577b34da6a3ce929d0e0e4736
X-B3-SpanId: 00f067aa0ba902b7
X-B3-ParentSpanId: e457b5a2e4d86bd1
X-B3-Sampled: 1
```

**Single-header format:**
```
b3: 4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-1-e457b5a2e4d86bd1
    {TraceId}-{SpanId}-{SamplingState}-{ParentSpanId}
```

### Baggage

Key-value pairs that propagate across service boundaries:
```
baggage: userId=alice,tenantId=acme,featureFlag=dark-mode
```

Use cases:
- Propagate user ID for access logging
- Propagate tenant ID for multi-tenant routing
- Propagate feature flags for testing
- Propagate deployment info for canary analysis

## Semantic Conventions

### HTTP Spans

```
# Client span attributes
http.request.method          = "GET"
url.full                     = "https://api.example.com/users/123"
server.address               = "api.example.com"
server.port                  = 443
http.response.status_code    = 200
network.protocol.version     = "1.1"

# Server span attributes
http.request.method          = "GET"
url.path                     = "/users/123"
url.scheme                   = "https"
http.response.status_code    = 200
http.route                   = "/users/:id"
client.address               = "203.0.113.42"
user_agent.original          = "Mozilla/5.0..."
```

### Database Spans

```
db.system                    = "postgresql"
db.namespace                 = "users_db"
db.operation.name            = "SELECT"
db.query.text                = "SELECT * FROM users WHERE id = $1"
server.address               = "db.example.com"
server.port                  = 5432
db.response.status_code      = "00000"
```

### Messaging Spans

```
messaging.system             = "kafka"
messaging.destination.name   = "orders-topic"
messaging.operation.type     = "publish" | "process"
messaging.message.id         = "msg-123"
messaging.kafka.consumer.group = "order-processors"
messaging.kafka.message.offset = 42
```

### RPC Spans

```
rpc.system                   = "grpc"
rpc.service                  = "myapp.UserService"
rpc.method                   = "GetUser"
rpc.grpc.status_code         = 0
server.address               = "user-service.default.svc"
server.port                  = 50051
```

## Collector Receivers

| Receiver | Protocol | Port | Use Case |
|----------|----------|------|----------|
| otlp (gRPC) | gRPC | 4317 | Primary for OTel SDKs |
| otlp (HTTP) | HTTP | 4318 | Primary for browser/serverless |
| jaeger (thrift) | HTTP | 14268 | Jaeger client compatibility |
| jaeger (gRPC) | gRPC | 14250 | Jaeger agent compatibility |
| zipkin | HTTP | 9411 | Zipkin client compatibility |
| kafka | Kafka | - | Queue-based ingestion |
| filelog | File | - | Log file collection |
| hostmetrics | - | - | System metrics (CPU, memory, disk) |
| prometheus | HTTP | - | Scrape Prometheus endpoints |

## Collector Processors

| Processor | Purpose |
|-----------|---------|
| batch | Batches data for efficient export |
| memory_limiter | Prevents OOM by dropping data |
| filter | Drops unwanted spans/metrics/logs |
| attributes | Add, update, delete, hash attributes |
| resource | Modify resource attributes |
| tail_sampling | Keep traces based on complete trace data |
| probabilistic_sampler | Random sampling by trace ID |
| span | Modify span names and attributes |
| transform | OTTL-based transformation |
| k8sattributes | Add Kubernetes metadata |
| redaction | Remove sensitive data |
| groupbytrace | Buffer spans to group by trace |

## Collector Exporters

| Exporter | Backend | Protocol |
|----------|---------|----------|
| otlp | Any OTLP backend | gRPC/HTTP |
| otlphttp | Any OTLP backend | HTTP |
| jaeger | Jaeger | gRPC/Thrift |
| zipkin | Zipkin | HTTP |
| prometheus | Prometheus | Pull (scrape) |
| prometheusremotewrite | Prometheus/Mimir/Cortex | Push |
| datadog | Datadog | HTTP |
| elasticsearch | Elasticsearch | HTTP |
| loki | Grafana Loki | HTTP |
| googlecloud | Google Cloud | gRPC |
| awsxray | AWS X-Ray | HTTP |
| azuremonitor | Azure Monitor | HTTP |

## Sampling Strategies

### Head-Based Sampling

Decisions made at the start of a trace. Simple and efficient but may miss interesting traces.

| Sampler | Description | When to Use |
|---------|-------------|-------------|
| AlwaysOn | Sample 100% | Development, low traffic |
| AlwaysOff | Sample 0% | Disable tracing |
| TraceIdRatio | Sample X% | General purpose |
| ParentBased | Respect parent decision | Multi-service (recommended) |

### Tail-Based Sampling (Collector)

Decisions made after the trace is complete. More intelligent but requires buffering.

| Policy | Description |
|--------|-------------|
| always_sample | Keep all traces (baseline) |
| latency | Keep traces exceeding duration threshold |
| status_code | Keep traces with ERROR status |
| probabilistic | Random sample of remaining traces |
| string_attribute | Keep traces with specific attributes |
| rate_limiting | Keep up to N traces per second |
| composite | Combine multiple policies |

**Recommended tail-based configuration:**
1. Keep 100% of error traces
2. Keep 100% of slow traces (>2s)
3. Keep 100% of traces with `important=true` attribute
4. Sample 5-10% of remaining successful traces

## Environment Variables

Standard OTel environment variables for configuration:

| Variable | Default | Description |
|----------|---------|-------------|
| OTEL_SERVICE_NAME | unknown_service | Service name in resource |
| OTEL_RESOURCE_ATTRIBUTES | - | Additional resource attributes |
| OTEL_EXPORTER_OTLP_ENDPOINT | http://localhost:4317 (gRPC), http://localhost:4318 (HTTP) | Collector endpoint |
| OTEL_EXPORTER_OTLP_HEADERS | - | Additional headers for exporter |
| OTEL_EXPORTER_OTLP_PROTOCOL | grpc | Protocol: grpc or http/protobuf |
| OTEL_TRACES_EXPORTER | otlp | Trace exporter: otlp, jaeger, zipkin, console, none |
| OTEL_METRICS_EXPORTER | otlp | Metrics exporter: otlp, prometheus, console, none |
| OTEL_LOGS_EXPORTER | otlp | Logs exporter: otlp, console, none |
| OTEL_TRACES_SAMPLER | parentbased_always_on | Sampler type |
| OTEL_TRACES_SAMPLER_ARG | - | Sampler argument (e.g., 0.5 for 50%) |
| OTEL_PROPAGATORS | tracecontext,baggage | Context propagators |
| OTEL_LOG_LEVEL | info | SDK log level |
| OTEL_ATTRIBUTE_VALUE_LENGTH_LIMIT | - | Max attribute value length |
| OTEL_SPAN_ATTRIBUTE_COUNT_LIMIT | 128 | Max attributes per span |
| OTEL_SPAN_EVENT_COUNT_LIMIT | 128 | Max events per span |
| OTEL_SPAN_LINK_COUNT_LIMIT | 128 | Max links per span |
| OTEL_BSP_SCHEDULE_DELAY | 5000ms | Batch span processor delay |
| OTEL_BSP_MAX_QUEUE_SIZE | 2048 | Max spans in queue |
| OTEL_BSP_MAX_EXPORT_BATCH_SIZE | 512 | Max spans per batch |
| OTEL_BSP_EXPORT_TIMEOUT | 30000ms | Export timeout |

## Auto-Instrumentation Packages

### Node.js
```bash
npm install @opentelemetry/auto-instrumentations-node
```

Includes instrumentation for:
- `http` / `https` — HTTP client and server
- `express` — Express.js framework
- `fastify` — Fastify framework
- `koa` — Koa framework
- `pg` — PostgreSQL
- `mysql` / `mysql2` — MySQL
- `mongodb` — MongoDB
- `redis` / `ioredis` — Redis
- `amqplib` — RabbitMQ
- `kafkajs` — Kafka
- `graphql` — GraphQL
- `grpc-js` — gRPC
- `aws-sdk` — AWS SDK
- `undici` — Node.js HTTP client
- `fs` — File system (usually disabled)
- `dns` — DNS resolution (usually disabled)

### Python
```bash
pip install opentelemetry-distro
opentelemetry-bootstrap -a install
```

Includes instrumentation for:
- `requests` — HTTP client
- `httpx` — Async HTTP client
- `aiohttp` — Async HTTP client/server
- `fastapi` — FastAPI framework
- `django` — Django framework
- `flask` — Flask framework
- `sqlalchemy` — Database ORM
- `psycopg2` — PostgreSQL
- `pymysql` — MySQL
- `redis` — Redis
- `celery` — Task queue
- `boto3` — AWS SDK
- `grpc` — gRPC

### Go
```bash
go get go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp
go get go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc
go get go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux
```

Available contrib packages:
- `net/http` — Standard library HTTP
- `gorilla/mux` — Gorilla Mux router
- `gin-gonic/gin` — Gin framework
- `go-chi/chi` — Chi router
- `google.golang.org/grpc` — gRPC
- `database/sql` — Standard database interface
- `go-redis/redis` — Redis
- `aws/aws-sdk-go-v2` — AWS SDK
- `segmentio/kafka-go` — Kafka

### Java
```bash
# Zero-code instrumentation via Java agent
java -javaagent:opentelemetry-javaagent.jar -jar myapp.jar
```

Automatically instruments 100+ libraries including:
- Spring Boot, Spring MVC, Spring WebFlux
- JAX-RS, Servlet
- JDBC, Hibernate, JPA
- Jedis, Lettuce (Redis)
- Kafka, RabbitMQ
- gRPC, OkHttp, Apache HttpClient
- AWS SDK, Google Cloud SDK

## Collector Deployment Patterns

### Sidecar Pattern
- One Collector per pod/container
- Minimal network hops
- Per-service configuration
- Higher resource overhead
- Best for: Tail-based sampling per service, strict isolation

### Gateway Pattern
- Shared Collector cluster
- Centralized configuration
- Resource efficient
- Single point of failure (mitigate with HA)
- Best for: Most deployments, centralized processing

### Agent + Gateway Pattern
- Lightweight agent per node (Fluent Bit, OTel Collector)
- Heavy processing at gateway
- Best of both worlds
- Best for: Large-scale deployments, multi-cluster

```
┌──────────────────────────────────────┐
│            Node                       │
│  ┌─────────┐  ┌─────────┐           │
│  │  App 1   │  │  App 2   │           │
│  │ OTel SDK │  │ OTel SDK │           │
│  └────┬─────┘  └────┬─────┘           │
│       │              │                 │
│  ┌────┴──────────────┴────┐           │
│  │  OTel Collector Agent  │           │ ← Lightweight, per-node
│  │  (batch, compress)     │           │
│  └───────────┬────────────┘           │
└──────────────┼─────────────────────────┘
               │
               ▼
┌──────────────────────────────┐
│  OTel Collector Gateway      │ ← Centralized processing
│  (sample, filter, route)     │
│  (HA: 3+ replicas)          │
└──────────┬───────────────────┘
           │
    ┌──────┴──────┐
    ▼             ▼
┌────────┐  ┌────────┐
│ Jaeger │  │  Tempo │
└────────┘  └────────┘
```

## Performance Considerations

### SDK Overhead Targets
- **CPU**: < 3% additional CPU usage
- **Memory**: < 50MB additional memory
- **Latency**: < 1ms added to request latency
- **Throughput**: < 2% reduction in request throughput

### Optimization Tips
1. Use BatchSpanProcessor, never SimpleSpanProcessor in production
2. Configure appropriate batch sizes (512 default is good)
3. Use head-based sampling in SDK to reduce spans generated
4. Use tail-based sampling in Collector for intelligent selection
5. Disable noisy auto-instrumentations (fs, dns, net)
6. Limit attribute count and value length
7. Use async exporters (gRPC is faster than HTTP)
8. Set memory limits on Collector to prevent OOM
9. Use compression (gzip) for OTLP exports
10. Monitor Collector health metrics at :8888

### Collector Sizing Guidelines
| Traffic | CPU | Memory | Instances |
|---------|-----|--------|-----------|
| < 1K spans/s | 0.5 core | 512MB | 1-2 |
| 1K-10K spans/s | 1-2 cores | 1-2GB | 2-3 |
| 10K-100K spans/s | 4-8 cores | 4-8GB | 3-5 |
| > 100K spans/s | 8+ cores | 8-16GB | 5+ (HA) |

## Troubleshooting

### Common Issues

**No traces appearing in backend:**
1. Check OTEL_EXPORTER_OTLP_ENDPOINT is correct
2. Verify Collector is running: `curl http://collector:13133` (health check)
3. Check SDK logs: set `OTEL_LOG_LEVEL=debug`
4. Verify traces are sampled: check `OTEL_TRACES_SAMPLER` config
5. Check Collector logs for export errors

**Missing spans in a trace:**
1. Verify context propagation headers are forwarded
2. Check all services use the same propagation format
3. Ensure middleware order is correct (tracing before handlers)
4. Look for async operations that lose context

**High memory usage in Collector:**
1. Add `memory_limiter` processor
2. Reduce `tail_sampling.num_traces` buffer
3. Increase batch processor timeout to reduce buffering
4. Check for cardinality explosion in attributes

**Traces are incomplete (missing child spans):**
1. Ensure tracing SDK initializes before any other imports
2. Verify auto-instrumentation covers the libraries in use
3. Check for context loss in async operations (Promise chains, goroutines)
4. Ensure HTTP clients propagate headers correctly
