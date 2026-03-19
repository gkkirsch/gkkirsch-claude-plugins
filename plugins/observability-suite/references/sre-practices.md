# SRE Practices Reference

## The Three Pillars of Observability

### Logs
Discrete, timestamped records of events. Answer "what happened?"

**Use for:**
- Debugging specific requests or errors
- Audit trails and compliance
- Understanding error context and stack traces
- Investigating incidents after the fact

**Key properties:**
- Structured (JSON), not unstructured text
- Correlated via trace_id and request_id
- Appropriate log levels (ERROR, WARN, INFO, DEBUG)
- Sensitive data redacted

### Metrics
Numeric measurements collected over time. Answer "how much?" and "how fast?"

**Use for:**
- Real-time service health monitoring
- Alerting on thresholds and anomalies
- Capacity planning and trend analysis
- SLO tracking and error budget calculation

**Key properties:**
- Low cardinality labels (< 10 values per label)
- Consistent naming conventions
- Appropriate metric types (counter, gauge, histogram)
- Pre-aggregated via recording rules for dashboards

### Traces
End-to-end request journey through distributed systems. Answer "where?" and "why slow?"

**Use for:**
- Diagnosing latency in microservices
- Understanding request flow and dependencies
- Identifying bottlenecks and failure points
- Correlating events across services

**Key properties:**
- Context propagation across all boundaries
- Sampling strategy appropriate for traffic volume
- Semantic conventions for consistent attributes
- Correlated with logs via trace_id

## Golden Signals

The four signals from Google's SRE book that should be monitored for every service:

### Latency
The time it takes to serve a request.

```
Distinguish between:
- Latency of SUCCESSFUL requests (the normal user experience)
- Latency of FAILED requests (often very fast — a 500 error in 5ms isn't "fast")

Key metrics:
- P50 (median): Typical user experience
- P90: Most users' worst experience
- P99: Tail latency (important for upstream services)
- P99.9: Extreme tail (important for high-traffic services)
```

### Traffic
The amount of demand being placed on the system.

```
Measure in terms meaningful to the service:
- Web service: HTTP requests per second
- Database: Transactions or queries per second
- Message queue: Messages published/consumed per second
- Streaming: Sessions active, bytes per second
- Batch: Jobs started per hour

Key metrics:
- Request rate (total and by endpoint)
- Request rate by status code
- Traffic trend comparison (vs last week/month)
```

### Errors
The rate of requests that fail.

```
Types of errors:
- Explicit: HTTP 5xx responses, gRPC error codes
- Implicit: HTTP 200 with wrong content, degraded responses
- Policy: Responses slower than SLO threshold (counted as errors)

Key metrics:
- Error rate (errors / total requests)
- Error count by type/code
- Error rate trend (is it increasing?)
```

### Saturation
How "full" the service is. The most constrained resource.

```
Measure utilization AND predict problems:
- CPU utilization and throttling
- Memory utilization and OOM risk
- Disk I/O and space
- Network bandwidth
- Connection pool usage
- Queue depth and processing lag
- Thread pool utilization

Key rule: Alert on saturation BEFORE it causes errors.
```

## SLO Framework

### Defining SLOs

#### Step 1: Identify Critical User Journeys
```
User Journey: "Customer completes a purchase"
├── Browse products (GET /api/products)
├── Add to cart (POST /api/cart)
├── View cart (GET /api/cart)
├── Enter payment (POST /api/checkout/payment)
└── Confirm order (POST /api/checkout/confirm)
```

#### Step 2: Choose SLIs for Each Journey
```
Availability SLI:
  "Proportion of requests that complete without server error"
  Good: status_code < 500
  Total: all requests
  Measurement: sum(rate(requests{status!~"5.."}[window])) / sum(rate(requests[window]))

Latency SLI:
  "Proportion of requests that complete within 300ms"
  Good: duration < 300ms
  Total: all requests
  Measurement: sum(rate(duration_bucket{le="0.3"}[window])) / sum(rate(duration_count[window]))

Correctness SLI:
  "Proportion of operations that produce correct results"
  Good: result matches expected output
  Total: all operations
  Measurement: application-specific (checksums, validation, reconciliation)
```

#### Step 3: Set Targets
```
SLO Target Selection:
├── Start with current performance baseline
├── Consider user expectations and competition
├── Account for error budget needs (deploy, experiment, maintain)
├── Don't aim for 100% (it's impossible and counterproductive)
└── Iterate based on business feedback

Common targets:
├── Internal tools: 99% - 99.5%
├── B2B SaaS: 99.9% (three nines)
├── Consumer apps: 99.9% - 99.95%
├── Infrastructure/platform: 99.99% (four nines)
└── Life-critical systems: 99.999% (five nines)
```

#### Step 4: Calculate Error Budget

```
Error Budget = 1 - SLO target

Example: 99.9% availability SLO (30-day window)
  Total minutes in 30 days: 43,200
  Error budget: 0.1% = 43.2 minutes of downtime

  Monthly allowances:
  ├── 99%    = 432 minutes (7.2 hours)
  ├── 99.5%  = 216 minutes (3.6 hours)
  ├── 99.9%  = 43.2 minutes
  ├── 99.95% = 21.6 minutes
  ├── 99.99% = 4.32 minutes
  └── 99.999% = 0.43 minutes (26 seconds)
```

### Error Budget Policies

```
Error Budget Status → Actions

> 50% remaining (healthy):
├── Normal deployment velocity
├── Experimentation allowed
├── Technical debt work encouraged
└── Feature development prioritized

25-50% remaining (caution):
├── Reduce deployment frequency
├── Require additional review for risky changes
├── Prioritize reliability improvements
└── Increase testing coverage

< 25% remaining (at risk):
├── Freeze non-critical deployments
├── Focus exclusively on reliability
├── Require VP approval for feature launches
└── Conduct reliability review of all pending changes

Budget exhausted (violated):
├── Complete deployment freeze
├── All hands on reliability
├── Incident review for contributing events
├── Executive notification
└── Recovery plan required before resuming features
```

## Monitoring Strategy

### Layer Cake Model

Monitor at every layer, with different granularity:

```
┌──────────────────────────────────────────────┐
│ Layer 5: Business Metrics                     │
│ Revenue, signups, conversions, user activity  │
│ → Track business impact of technical issues   │
├──────────────────────────────────────────────┤
│ Layer 4: Application Metrics (RED)            │
│ Request rate, error rate, duration            │
│ → Track service-level health (SLOs live here) │
├──────────────────────────────────────────────┤
│ Layer 3: Middleware / Dependencies             │
│ Database, cache, queue, external API health   │
│ → Track dependency availability and latency   │
├──────────────────────────────────────────────┤
│ Layer 2: Container / Runtime Metrics          │
│ Pod health, restarts, resource limits         │
│ → Track deployment and scheduling health      │
├──────────────────────────────────────────────┤
│ Layer 1: Infrastructure Metrics (USE)         │
│ CPU, memory, disk, network                    │
│ → Track physical resource health              │
└──────────────────────────────────────────────┘
```

### Dashboard Hierarchy

```
Level 1: Executive Overview
├── SLO status (green/yellow/red) for all services
├── Error budget remaining for each SLO
├── Active incidents count
└── Business metrics summary

Level 2: Service Overview
├── RED metrics for each service
├── Dependency health status
├── Recent deployments
├── SLO burn rate over time
└── Top errors by frequency

Level 3: Service Deep-Dive
├── Detailed RED metrics per endpoint
├── Database query performance
├── Cache hit rates
├── Queue depths and processing rates
├── Resource utilization (CPU, memory)
└── Error breakdown by type and code

Level 4: Debugging
├── Individual request traces
├── Log search with filters
├── Resource usage time series
├── Network traffic analysis
└── Thread/goroutine/worker analysis
```

## Incident Response

### Severity Classification

| Severity | Impact | Response | Examples |
|----------|--------|----------|---------|
| SEV1 | Service outage affecting all users | Immediate, all-hands | Complete API down, data loss, security breach |
| SEV2 | Service degraded for significant users | Immediate, on-call + backup | Partial outage, major feature broken, >5% errors |
| SEV3 | Minor degradation, workaround exists | Business hours | Slow performance, minor feature broken, single endpoint down |
| SEV4 | No user impact, monitoring anomaly | Next business day | Elevated error rate on non-critical path, capacity approaching limits |

### Incident Roles

```
Incident Commander (IC):
├── Drives the incident response
├── Makes decisions about mitigation strategy
├── Coordinates communication between teams
├── Decides when to escalate
└── Declares incident resolved

Communications Lead:
├── Posts status updates to stakeholders
├── Updates status page
├── Manages external communications
├── Maintains incident timeline
└── Writes initial post-mortem summary

Subject Matter Experts (SMEs):
├── Diagnose root cause
├── Implement fixes
├── Validate resolution
└── Provide technical context to IC
```

### Incident Timeline Template

```
[TIME] DETECT: Alert fired / User report received
  - What alert? What symptoms?
  - What dashboards show the problem?

[TIME] ASSESS: Initial triage
  - Severity classification
  - User impact assessment
  - Blast radius determination

[TIME] ESCALATE: Incident declared
  - Roles assigned (IC, Comms, SMEs)
  - Incident channel created
  - Status page updated

[TIME] DIAGNOSE: Root cause investigation
  - Hypotheses tested
  - Evidence gathered
  - Root cause identified

[TIME] MITIGATE: Stop the bleeding
  - Action taken (rollback, scaling, circuit breaker)
  - Impact reduced/eliminated

[TIME] RESOLVE: Service fully restored
  - Monitoring shows normal behavior
  - Alerts resolved
  - User-facing impact confirmed ended

[TIME] FOLLOW-UP: Post-incident
  - Post-mortem scheduled
  - Action items documented
  - Knowledge shared
```

## Capacity Planning

### Approach

```
1. MEASURE current usage
   ├── Peak request rate
   ├── Peak resource utilization (CPU, memory, disk)
   ├── Growth rate (week-over-week, month-over-month)
   └── Seasonal patterns

2. MODEL future demand
   ├── Organic growth rate
   ├── Planned feature launches
   ├── Marketing campaigns
   ├── Seasonal events (Black Friday, etc.)
   └── New customer onboarding

3. PLAN capacity
   ├── Headroom target: run at 50-60% utilization normally
   ├── Burst capacity: handle 2-3x normal traffic
   ├── Scaling lead time: how long to add capacity?
   ├── Cost optimization: right-size for actual needs
   └── Reserve for incidents: extra capacity for failover

4. AUTOMATE scaling
   ├── Horizontal Pod Autoscaler (Kubernetes)
   ├── Auto Scaling Groups (AWS)
   ├── Cloud Functions/Lambda (serverless)
   └── Custom scaling based on queue depth or custom metrics
```

### PromQL for Capacity Planning

```promql
# Growth rate (week over week)
(sum(rate(http_requests_total[1d]))
 - sum(rate(http_requests_total[1d] offset 7d)))
/ sum(rate(http_requests_total[1d] offset 7d))
* 100

# Predict disk full (linear extrapolation)
predict_linear(node_filesystem_avail_bytes{mountpoint="/"}[7d], 30 * 24 * 3600)

# Peak utilization in last 7 days
max_over_time(
  (1 - (avg by (instance) (rate(node_cpu_seconds_total{mode="idle"}[5m])))) [7d:]
)

# Database connection pool trend
avg_over_time(db_connection_pool_size{state="active"}[7d])
```

## Reliability Patterns

### Circuit Breaker

Prevent cascading failures by stopping calls to a failing service:

```
States:
  CLOSED (normal) → Monitor error rate
    └── If error rate exceeds threshold → OPEN

  OPEN (failing) → Reject all calls immediately
    └── After timeout period → HALF-OPEN

  HALF-OPEN (testing) → Allow limited test requests
    ├── If test requests succeed → CLOSED
    └── If test requests fail → OPEN

Metrics to track:
  - circuit_breaker_state{service="...", target="..."} (gauge: 0=closed, 1=open, 2=half-open)
  - circuit_breaker_calls_total{service="...", state="...", result="..."}
```

### Retry with Backoff

```
Strategy: Exponential backoff with jitter

Attempt 1: Immediate
Attempt 2: 100ms + random(0-50ms)
Attempt 3: 200ms + random(0-100ms)
Attempt 4: 400ms + random(0-200ms)
Attempt 5: 800ms + random(0-400ms)
Max: 5 retries, 2s total timeout

Metrics to track:
  - retries_total{service="...", attempt="..."} (counter)
  - retry_success_total{service="...", attempt="..."} (counter)
```

### Load Shedding

Reject excess requests gracefully when overloaded:

```
Triggers:
  - CPU > 80% sustained
  - Queue depth > threshold
  - Request latency > 2x normal
  - Connection pool nearly exhausted

Response:
  - Return HTTP 503 with Retry-After header
  - Prioritize authenticated > anonymous
  - Prioritize critical endpoints > non-critical
  - Track: load_shed_total{reason="..."} (counter)
```

### Graceful Degradation

Reduce functionality to maintain core service:

```
Degradation levels:
  Level 0 (normal): Full functionality
  Level 1: Disable non-essential features (recommendations, analytics)
  Level 2: Serve cached/stale data, disable writes
  Level 3: Static maintenance page with status updates

Metrics:
  - degradation_level{service="..."} (gauge: 0-3)
  - degraded_responses_total{feature="..."} (counter)
```

## On-Call Best Practices

### On-Call Health Targets

```
Per on-call shift (1 week):
  - Critical pages: < 2 per week
  - Total pages (all severities): < 5 per week
  - Night pages (10pm-6am): < 1 per week
  - Mean time to acknowledge: < 5 minutes
  - Mean time to resolve: < 30 minutes

Team-level monthly:
  - Alert noise ratio: < 20% (non-actionable alerts / total)
  - False positive rate: < 10%
  - Night page rate trending downward
  - No single engineer with > 1.5x average page count
```

### Handoff Checklist

```
Outgoing on-call:
□ Document any ongoing issues or workarounds
□ Note any recent deployments or changes
□ Highlight flapping or recurring alerts
□ Share relevant context about open incidents
□ Update any time-sensitive runbooks

Incoming on-call:
□ Review current alert state in Alertmanager/PagerDuty
□ Check error budget status for all SLOs
□ Review upcoming deployments/maintenance
□ Verify access to all necessary tools and dashboards
□ Confirm phone notifications are working
```

## Toil Reduction

### What is Toil?

```
Toil is work that:
  ✗ Is manual (requires human intervention)
  ✗ Is repetitive (same task again and again)
  ✗ Is automatable (could be done by a machine)
  ✗ Is tactical (interrupt-driven, reactive)
  ✗ Has no enduring value (doesn't improve the system)
  ✗ Scales linearly with service growth

Examples of toil:
  - Manually restarting pods that OOM
  - Manually scaling services for traffic spikes
  - Manually rotating secrets/certificates
  - Manually clearing disk space
  - Manually investigating repetitive alerts
  - Running manual deployment checklists
```

### Toil Tracking

```
Track toil metrics:
  - Hours spent on toil per engineer per week
  - Most common toil tasks (categorize)
  - Toil ratio: toil_hours / total_engineering_hours
  - Target: < 50% toil (Google SRE recommendation)
  - Ideal: < 30% toil

Automate highest-toil tasks first:
  1. Rank by frequency × time_per_occurrence
  2. Estimate automation effort
  3. Calculate ROI: time_saved / automation_effort
  4. Prioritize high-ROI automations
```

## Observability Maturity Model

### Level 1: Reactive
```
- Basic uptime monitoring (ping, HTTP checks)
- Unstructured logs (console.log, print statements)
- No metrics beyond cloud provider defaults
- No tracing
- Alert on: "is the service up?"
- Incident response: manual, ad-hoc
```

### Level 2: Basic
```
- Structured logging with log aggregation
- Basic RED metrics (request rate, errors, duration)
- Health check endpoints
- Simple threshold alerts
- Alert on: error rate, latency, resource usage
- Incident response: runbooks for common issues
```

### Level 3: Proactive
```
- Distributed tracing across services
- SLO-based alerting with error budgets
- Correlated logs, metrics, and traces
- Recording rules and precomputed dashboards
- Alert on: SLO burn rate, dependency degradation
- Incident response: structured, role-based
```

### Level 4: Advanced
```
- Tail-based sampling for intelligent trace collection
- Anomaly detection (dynamic baselines)
- Automated remediation for known issues
- Capacity planning with predictive models
- Chaos engineering for reliability validation
- Post-mortems with tracked action items
- Toil budget and reduction targets
```

### Level 5: Optimized
```
- Full automation of incident response
- ML-driven anomaly detection and root cause analysis
- Self-healing infrastructure
- Real-time SLO dashboards for all stakeholders
- Continuous reliability validation (chaos engineering in production)
- Observability-as-code (all config in version control)
- Observability cost optimization
```
