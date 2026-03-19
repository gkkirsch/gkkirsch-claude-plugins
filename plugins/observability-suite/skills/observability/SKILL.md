# Observability Skill

## Metadata
- Name: observability
- Description: Design and implement production-grade monitoring and observability — structured logging with ELK/Loki, metrics with Prometheus/Grafana, distributed tracing with OpenTelemetry/Jaeger, and alerting with PagerDuty integration
- Version: 1.0.0

## Trigger
Activate when the user asks about:
- Logging (structured logging, log aggregation, ELK, Loki, Pino, Winston, structlog, zerolog)
- Metrics (Prometheus, Grafana, prom-client, custom metrics, counters, histograms, gauges)
- Tracing (OpenTelemetry, Jaeger, Zipkin, distributed tracing, spans, trace context)
- Alerting (Alertmanager, PagerDuty, OpsGenie, alert rules, runbooks, on-call)
- Monitoring (dashboards, health checks, uptime, service health, observability)
- SLOs (service level objectives, SLIs, error budgets, burn rate, reliability)
- SRE practices (incident response, post-mortems, capacity planning, toil reduction)
- Observability infrastructure (Fluent Bit, Promtail, OTel Collector, Filebeat)

## Agents

### logging-architect
**When to use:** Structured logging setup, log aggregation, ELK Stack, Grafana Loki, Fluent Bit, log correlation, audit logging, log redaction.

**Capabilities:**
- Structured JSON logging with Pino (Node.js), structlog (Python), zerolog (Go)
- Correlation ID propagation with AsyncLocalStorage / contextvars
- ELK Stack (Elasticsearch, Logstash, Kibana) deployment and configuration
- Grafana Loki with Promtail for cost-effective log aggregation
- Fluent Bit DaemonSet for Kubernetes log collection
- Log redaction for PII, tokens, passwords (Pino redact, Logstash gsub)
- Audit logging for SOC 2 / HIPAA / GDPR compliance
- Index lifecycle management for log retention and cost control

### metrics-engineer
**When to use:** Prometheus metrics, Grafana dashboards, custom application metrics, SLO/SLI design, recording rules, capacity planning.

**Capabilities:**
- Application instrumentation with prom-client (Node.js), prometheus_client (Python), client_golang (Go)
- RED method metrics (Rate, Errors, Duration) for every HTTP endpoint
- USE method metrics (Utilization, Saturation, Errors) for infrastructure
- Custom business metrics (orders, revenue, signups, conversion rates)
- Database, cache, queue, and external API metrics
- Prometheus server configuration with Kubernetes service discovery
- Recording rules for dashboard performance optimization
- SLO/SLI definitions with error budget calculation and burn rate alerts
- Grafana dashboard design (service overview, SLO tracking, infrastructure)
- Metric naming conventions and cardinality management

### tracing-specialist
**When to use:** OpenTelemetry SDK setup, distributed tracing, trace context propagation, Jaeger/Tempo deployment, sampling strategies, trace-log correlation.

**Capabilities:**
- OpenTelemetry SDK initialization with auto-instrumentation (Node.js, Python, Go, Java)
- W3C Trace Context and B3 propagation across service boundaries
- OpenTelemetry Collector deployment with receivers, processors, and exporters
- Custom span creation for business-critical operations
- Tail-based sampling in the Collector for intelligent trace selection
- Jaeger and Grafana Tempo deployment for trace visualization
- Trace-log correlation via trace_id injection into log entries
- Metric exemplars for trace-metric correlation
- Semantic conventions for consistent span attributes
- Context propagation across queues (RabbitMQ, Kafka, SQS)

### alerting-designer
**When to use:** Alert rule design, Alertmanager configuration, PagerDuty integration, runbook creation, on-call rotation, incident response workflows.

**Capabilities:**
- SLO-based burn rate alerts with multi-window detection
- Prometheus Alertmanager routing, grouping, and inhibition
- PagerDuty integration with escalation policies and service routing
- Slack notification formatting for warning and info alerts
- Infrastructure alerts (node health, disk, CPU, memory, pods)
- Application alerts (error rate, latency, traffic anomalies, dependencies)
- Runbook templates linked from alert annotations
- On-call rotation design and handoff checklists
- Incident response workflows with role assignments
- Post-mortem templates with action item tracking
- Alert fatigue reduction through deduplication and noise analysis

## References

### opentelemetry-guide
Complete OpenTelemetry reference: architecture, SDK components (TracerProvider, MeterProvider), signals (traces, metrics, logs), context propagation formats (W3C, B3), semantic conventions, Collector configuration, auto-instrumentation packages, sampling strategies, environment variables, deployment patterns, and troubleshooting.

### prometheus-patterns
PromQL mastery: selectors, range vectors, essential functions (rate, histogram_quantile, increase), RED method queries, USE method queries, SLO/SLI queries, recording rules, federation configuration, Thanos/Mimir long-term storage, common anti-patterns (high cardinality, incorrect rate usage), and security considerations.

### sre-practices
SRE operational reference: three pillars of observability, golden signals, SLO framework (defining SLIs, setting targets, error budgets, budget policies), monitoring strategy (layer cake model, dashboard hierarchy), incident response (severity classification, roles, timeline template), capacity planning, reliability patterns (circuit breaker, retry, load shedding, graceful degradation), on-call best practices, toil reduction, and observability maturity model.

## Workflow

1. **Assess** — Detect the stack, find existing observability setup, identify gaps
2. **Instrument** — Add structured logging, metrics, and tracing to the application
3. **Collect** — Configure log aggregation, Prometheus scraping, and trace collection
4. **Visualize** — Build Grafana dashboards for logs, metrics, and traces
5. **Alert** — Define SLOs, create alert rules, set up PagerDuty routing
6. **Validate** — Verify all signals flow correctly, test alerts, check correlation

## Languages

All agents provide code in:
- Node.js (Express, Fastify) with TypeScript
- Python (FastAPI, Django) with type hints
- Go (net/http, Gin, Gorilla Mux)

## Quality Standards

- Structured JSON logging in production with sensitive data redacted
- Correlation IDs propagate across all service boundaries
- Prometheus metrics follow naming conventions with low cardinality
- OpenTelemetry traces include semantic convention attributes
- SLOs defined for all critical user journeys
- Every critical alert has a linked runbook and dashboard
- Alert rules use `for` duration to avoid transient spikes
- Tail-based sampling keeps 100% of error traces
- All observability config is version-controlled (observability-as-code)
