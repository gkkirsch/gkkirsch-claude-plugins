# /write-observability Command

Design and implement monitoring, logging, metrics, tracing, and alerting for your applications. This command activates the Monitoring & Observability Suite agents to help you build production-grade observability.

## Usage

```
/write-observability [subcommand] [options]
```

## Subcommands

### `logging`
Set up structured logging with log aggregation.

```
/write-observability logging
```

Activates the **logging-architect** agent to help you:
- Replace console.log/print with structured JSON logging (Pino, structlog, zerolog)
- Set up correlation ID propagation across services
- Configure log aggregation (ELK Stack or Grafana Loki)
- Implement log redaction for sensitive data (passwords, tokens, PII)
- Configure Fluent Bit or Promtail for log collection
- Set up log-based alerting for error patterns
- Create Kibana/Grafana dashboards for log analysis

### `metrics`
Instrument your application with Prometheus metrics and Grafana dashboards.

```
/write-observability metrics
```

Activates the **metrics-engineer** agent to help you:
- Instrument HTTP endpoints with RED metrics (Rate, Errors, Duration)
- Add custom business metrics (orders, revenue, signups)
- Track database query performance and connection pools
- Monitor cache hit rates and external API latency
- Configure Prometheus server and scrape targets
- Create recording rules for dashboard performance
- Build Grafana dashboards for service overview
- Design SLOs with error budget tracking

### `tracing`
Add distributed tracing with OpenTelemetry.

```
/write-observability tracing
```

Activates the **tracing-specialist** agent to help you:
- Install and configure OpenTelemetry SDK with auto-instrumentation
- Set up context propagation (W3C Trace Context, B3)
- Deploy OpenTelemetry Collector with processors and exporters
- Add custom spans for business-critical operations
- Configure tail-based sampling for intelligent trace selection
- Deploy Jaeger or Grafana Tempo for trace visualization
- Set up trace-log-metric correlation
- Implement trace context propagation across queues and async operations

### `alerting`
Design alerting rules, PagerDuty integration, and runbooks.

```
/write-observability alerting
```

Activates the **alerting-designer** agent to help you:
- Create SLO-based burn rate alerts (critical + warning)
- Configure Prometheus Alertmanager with routing and grouping
- Integrate with PagerDuty for on-call paging
- Set up Slack notifications for warning-level alerts
- Write runbooks for every critical alert
- Design on-call rotation schedules
- Configure alert inhibition to prevent storms
- Build incident response workflows

### `full`
Set up complete observability for your application from scratch.

```
/write-observability full
```

Activates all agents in sequence to build your entire observability stack:
1. **Logging Architect** — Structured logging and log aggregation
2. **Metrics Engineer** — Application metrics and Grafana dashboards
3. **Tracing Specialist** — Distributed tracing with OpenTelemetry
4. **Alerting Designer** — Alert rules, PagerDuty, and runbooks

### `audit`
Review existing observability implementation for gaps.

```
/write-observability audit
```

Reviews your codebase for common observability issues:
- console.log/print statements in production code
- Missing correlation IDs in log entries
- No /metrics endpoint or metrics instrumentation
- Missing distributed tracing setup
- No structured logging configuration
- Sensitive data in logs (passwords, tokens, PII)
- High-cardinality metric labels
- Missing health check endpoints
- No alert rules defined
- Missing SLO definitions

### `slo`
Define and implement SLOs for your services.

```
/write-observability slo
```

Combines **metrics-engineer** and **alerting-designer** to help you:
- Define SLIs for critical user journeys
- Set SLO targets with error budgets
- Create recording rules for SLO calculation
- Build burn rate alert rules
- Create Grafana SLO dashboards with budget tracking
- Design error budget policies

## Examples

```
# Set up structured logging for a Node.js Express app
/write-observability logging

# Add Prometheus metrics to a Python FastAPI service
/write-observability metrics

# Add OpenTelemetry tracing across microservices
/write-observability tracing

# Create alerting rules and PagerDuty integration
/write-observability alerting

# Build complete observability stack for a new service
/write-observability full

# Audit existing observability setup for gaps
/write-observability audit

# Define SLOs and error budget alerting
/write-observability slo
```

## Reference Files

The suite includes detailed reference documents:
- **opentelemetry-guide.md** — OTel SDK, Collector, exporters, semantic conventions, sampling
- **prometheus-patterns.md** — PromQL queries, recording rules, federation, anti-patterns
- **sre-practices.md** — SLO framework, error budgets, incident response, capacity planning

## Supported Languages

All agents provide production-ready code in:
- **Node.js** (Express, Fastify) with TypeScript
- **Python** (FastAPI, Django) with type hints
- **Go** (net/http, Gin, Gorilla Mux)

## What This Suite Covers

```
Logging                       Metrics
├── Structured JSON logging   ├── Prometheus instrumentation
├── Pino / structlog / zerolog├── RED method (Rate, Errors, Duration)
├── Correlation IDs           ├── USE method (Utilization, Saturation, Errors)
├── Log aggregation (ELK/Loki)├── Custom business metrics
├── Fluent Bit / Promtail     ├── Recording rules
├── Log redaction              ├── SLO/SLI tracking
├── Audit logging              ├── Grafana dashboards
└── Log-based alerting         └── Error budget calculation

Tracing                       Alerting
├── OpenTelemetry SDK          ├── Prometheus Alertmanager
├── Auto-instrumentation       ├── SLO burn rate alerts
├── W3C Trace Context          ├── PagerDuty integration
├── OTel Collector             ├── Slack notifications
├── Jaeger / Tempo             ├── Alert routing & grouping
├── Tail-based sampling        ├── Inhibition rules
├── Custom spans               ├── Runbook design
├── Trace-log correlation      ├── On-call rotation
└── Semantic conventions       └── Incident response workflow
```
