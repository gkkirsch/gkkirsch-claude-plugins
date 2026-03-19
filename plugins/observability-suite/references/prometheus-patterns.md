# Prometheus Patterns Reference

## PromQL Fundamentals

### Data Types

| Type | Description | Example |
|------|-------------|---------|
| Instant vector | Set of time series with a single sample per series | `http_requests_total` |
| Range vector | Set of time series with a range of samples | `http_requests_total[5m]` |
| Scalar | Single numeric value | `42` or `3.14` |
| String | Single string value | `"hello"` (rarely used) |

### Selectors

```promql
# Exact match
http_requests_total{method="GET"}

# Regex match
http_requests_total{method=~"GET|POST"}

# Negative match
http_requests_total{status_code!="200"}

# Negative regex match
http_requests_total{path!~"/health.*"}

# Multiple selectors
http_requests_total{method="GET", status_code=~"2..", service="api"}
```

### Range Vector Selectors

```promql
# Last 5 minutes of data
http_requests_total[5m]

# Last 1 hour
http_requests_total[1h]

# Time units: ms, s, m, h, d, w, y
# Can combine: 1h30m

# Offset: look back in time
http_requests_total offset 1h      # Values from 1 hour ago
rate(http_requests_total[5m] offset 1h)  # Rate from 1 hour ago
```

## Essential Functions

### rate() — Per-Second Rate of Increase

The most important function for counters. Calculates the per-second average rate of increase over a range.

```promql
# Request rate per second (5-minute window)
rate(http_requests_total[5m])

# Rate by service
sum by (service) (rate(http_requests_total[5m]))

# Error rate as percentage
sum(rate(http_requests_total{status_code=~"5.."}[5m]))
/
sum(rate(http_requests_total[5m]))
* 100
```

**Important**: Always use rate() before sum() for counters. Never sum raw counters then rate.

### irate() — Instant Rate

Uses only the last two data points. More responsive but noisier than rate().

```promql
# Use for dashboards where you want to see spikes
irate(http_requests_total[5m])

# Use rate() for alerts (more stable)
# Use irate() for dashboards (more responsive)
```

### increase() — Total Increase Over Range

Extrapolated increase over the range window. Essentially `rate() * seconds`.

```promql
# Total requests in the last hour
increase(http_requests_total[1h])

# Total errors today
increase(http_requests_total{status_code=~"5.."}[24h])
```

### histogram_quantile() — Calculate Percentiles

```promql
# P50 (median) latency
histogram_quantile(0.50,
  sum by (le) (rate(http_request_duration_seconds_bucket[5m]))
)

# P90 latency by service
histogram_quantile(0.90,
  sum by (service, le) (rate(http_request_duration_seconds_bucket[5m]))
)

# P99 latency
histogram_quantile(0.99,
  sum by (le) (rate(http_request_duration_seconds_bucket[5m]))
)

# Average latency (without histogram_quantile)
sum(rate(http_request_duration_seconds_sum[5m]))
/
sum(rate(http_request_duration_seconds_count[5m]))
```

**Important**: The `le` label must be preserved in the `by` clause for `histogram_quantile()` to work.

### Aggregation Operators

```promql
# Sum across all instances
sum(http_requests_total)

# Sum by specific labels (keep those labels)
sum by (service, method) (rate(http_requests_total[5m]))

# Sum ignoring specific labels (drop those labels)
sum without (instance, pod) (rate(http_requests_total[5m]))

# Average
avg by (service) (rate(http_request_duration_seconds_sum[5m]) / rate(http_request_duration_seconds_count[5m]))

# Min / Max
min by (service) (up)
max by (service) (http_request_duration_seconds_bucket{le="+Inf"})

# Count number of time series
count by (service) (up)

# Standard deviation
stddev by (service) (rate(http_request_duration_seconds_sum[5m]))

# Top/Bottom K
topk(5, sum by (service) (rate(http_requests_total[5m])))
bottomk(3, sum by (service) (rate(http_requests_total[5m])))

# Quantile across instances (not histogram_quantile)
quantile by (service) (0.95, rate(http_request_duration_seconds_sum[5m]))
```

## RED Method Queries

Rate, Errors, Duration — the three key metrics for request-driven services.

### Rate (Request Throughput)

```promql
# Total request rate
sum(rate(http_requests_total[5m]))

# Request rate by service
sum by (service) (rate(http_requests_total[5m]))

# Request rate by endpoint
sum by (method, path) (rate(http_requests_total[5m]))

# Request rate trend (compare to yesterday)
sum(rate(http_requests_total[5m]))
/
sum(rate(http_requests_total[5m] offset 1d))
```

### Errors (Error Rate)

```promql
# Error rate (percentage)
sum by (service) (rate(http_requests_total{status_code=~"5.."}[5m]))
/
sum by (service) (rate(http_requests_total[5m]))
* 100

# Error rate including 4xx (client errors)
sum by (service) (rate(http_requests_total{status_code=~"[45].."}[5m]))
/
sum by (service) (rate(http_requests_total[5m]))
* 100

# Errors per second
sum by (service) (rate(http_requests_total{status_code=~"5.."}[5m]))

# Error rate by endpoint
sum by (method, path) (rate(http_requests_total{status_code=~"5.."}[5m]))
/
sum by (method, path) (rate(http_requests_total[5m]))
```

### Duration (Latency)

```promql
# P50 latency
histogram_quantile(0.50,
  sum by (service, le) (rate(http_request_duration_seconds_bucket[5m]))
)

# P90 latency
histogram_quantile(0.90,
  sum by (service, le) (rate(http_request_duration_seconds_bucket[5m]))
)

# P99 latency
histogram_quantile(0.99,
  sum by (service, le) (rate(http_request_duration_seconds_bucket[5m]))
)

# Average latency
sum by (service) (rate(http_request_duration_seconds_sum[5m]))
/
sum by (service) (rate(http_request_duration_seconds_count[5m]))

# Apdex score (target: 300ms, tolerating: 1.2s)
(
  sum(rate(http_request_duration_seconds_bucket{le="0.3"}[5m]))
  +
  sum(rate(http_request_duration_seconds_bucket{le="1.2"}[5m]))
)
/
(2 * sum(rate(http_request_duration_seconds_count[5m])))
```

## USE Method Queries

Utilization, Saturation, Errors — for infrastructure resources.

### CPU

```promql
# CPU utilization (%)
100 - (avg by (instance) (rate(node_cpu_seconds_total{mode="idle"}[5m])) * 100)

# CPU saturation (load average vs CPU count)
node_load1 / count without (cpu, mode) (node_cpu_seconds_total{mode="idle"})

# CPU by mode
sum by (instance, mode) (rate(node_cpu_seconds_total[5m]))
```

### Memory

```promql
# Memory utilization (%)
(1 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes)) * 100

# Memory available
node_memory_MemAvailable_bytes / 1024 / 1024 / 1024  # in GB

# Memory saturation (swap usage)
(node_memory_SwapTotal_bytes - node_memory_SwapFree_bytes)
/ node_memory_SwapTotal_bytes * 100
```

### Disk

```promql
# Disk utilization (%)
(1 - node_filesystem_avail_bytes{mountpoint="/"} / node_filesystem_size_bytes{mountpoint="/"}) * 100

# Disk I/O utilization
rate(node_disk_io_time_seconds_total[5m]) * 100

# Disk saturation (I/O wait)
rate(node_disk_io_time_weighted_seconds_total[5m])

# Disk space prediction (when will it fill?)
predict_linear(node_filesystem_avail_bytes{mountpoint="/"}[6h], 24 * 3600) < 0
```

### Network

```promql
# Network throughput (bytes/sec)
rate(node_network_receive_bytes_total{device!="lo"}[5m])
rate(node_network_transmit_bytes_total{device!="lo"}[5m])

# Network errors
rate(node_network_receive_errs_total[5m])
rate(node_network_transmit_errs_total[5m])

# Network saturation (dropped packets)
rate(node_network_receive_drop_total[5m])
rate(node_network_transmit_drop_total[5m])
```

## SLO/SLI Queries

### Availability SLO

```promql
# Availability over 30 days (target: 99.9%)
1 - (
  sum(increase(http_requests_total{status_code=~"5.."}[30d]))
  /
  sum(increase(http_requests_total[30d]))
)

# Error budget remaining (%)
1 - (
  (1 - (
    1 - sum(increase(http_requests_total{status_code=~"5.."}[30d]))
    / sum(increase(http_requests_total[30d]))
  ))
  / (1 - 0.999)  # 99.9% target
)

# Error budget burn rate
(
  sum(rate(http_requests_total{status_code=~"5.."}[1h]))
  / sum(rate(http_requests_total[1h]))
)
/ (1 - 0.999)  # How fast we're burning relative to budget
```

### Latency SLO

```promql
# Percentage of requests under 300ms (target: 99%)
sum(rate(http_request_duration_seconds_bucket{le="0.3"}[30d]))
/
sum(rate(http_request_duration_seconds_count[30d]))

# Latency budget burn rate
(1 - (
  sum(rate(http_request_duration_seconds_bucket{le="0.3"}[1h]))
  / sum(rate(http_request_duration_seconds_count[1h]))
))
/ (1 - 0.99)
```

### Multi-Window Burn Rate

```promql
# Fast burn rate (1h window) for rapid detection
(
  1 - sum(rate(http_requests_total{status_code!~"5.."}[1h]))
  / sum(rate(http_requests_total[1h]))
) / (1 - 0.999)

# Slow burn rate (6h window) for sustained issues
(
  1 - sum(rate(http_requests_total{status_code!~"5.."}[6h]))
  / sum(rate(http_requests_total[6h]))
) / (1 - 0.999)
```

## Recording Rules

Recording rules precompute expensive queries for faster dashboards and reliable alerts.

### Naming Convention

```
level:metric:operations

Examples:
  http:requests:rate5m          # Aggregated request rate
  http:errors:ratio_rate5m      # Error ratio
  http:latency:p99_5m           # P99 latency
  node:cpu:utilization_rate5m   # CPU utilization
```

### Common Recording Rules

```yaml
groups:
  - name: http_recording_rules
    interval: 30s
    rules:
      # Request rate by service
      - record: http:requests:rate5m
        expr: sum by (service) (rate(http_requests_total[5m]))

      # Error rate by service
      - record: http:errors:rate5m
        expr: sum by (service) (rate(http_requests_total{status_code=~"5.."}[5m]))

      # Error ratio by service
      - record: http:error_ratio:rate5m
        expr: |
          http:errors:rate5m / http:requests:rate5m

      # Latency percentiles
      - record: http:latency:p50_5m
        expr: |
          histogram_quantile(0.50,
            sum by (service, le) (rate(http_request_duration_seconds_bucket[5m])))

      - record: http:latency:p90_5m
        expr: |
          histogram_quantile(0.90,
            sum by (service, le) (rate(http_request_duration_seconds_bucket[5m])))

      - record: http:latency:p99_5m
        expr: |
          histogram_quantile(0.99,
            sum by (service, le) (rate(http_request_duration_seconds_bucket[5m])))

      # Average request duration
      - record: http:latency:avg_5m
        expr: |
          sum by (service) (rate(http_request_duration_seconds_sum[5m]))
          /
          sum by (service) (rate(http_request_duration_seconds_count[5m]))
```

## Federation and Long-Term Storage

### Federation Configuration

For multi-cluster Prometheus:

```yaml
# prometheus-federation.yml
scrape_configs:
  - job_name: federate-cluster-us-east
    honor_labels: true
    metrics_path: /federate
    params:
      'match[]':
        - '{__name__=~"http:.*"}'       # Recording rules
        - '{__name__=~"slo:.*"}'         # SLO metrics
        - '{__name__=~"node:.*"}'        # Node metrics
    static_configs:
      - targets:
          - prometheus-us-east.example.com:9090
        labels:
          cluster: us-east

  - job_name: federate-cluster-eu-west
    honor_labels: true
    metrics_path: /federate
    params:
      'match[]':
        - '{__name__=~"http:.*"}'
        - '{__name__=~"slo:.*"}'
        - '{__name__=~"node:.*"}'
    static_configs:
      - targets:
          - prometheus-eu-west.example.com:9090
        labels:
          cluster: eu-west
```

### Thanos Sidecar (Long-Term Storage)

```yaml
# thanos-sidecar alongside Prometheus
containers:
  - name: prometheus
    image: prom/prometheus:v2.50.0
    args:
      - '--storage.tsdb.min-block-duration=2h'
      - '--storage.tsdb.max-block-duration=2h'
      - '--storage.tsdb.retention.time=6h'

  - name: thanos-sidecar
    image: thanosio/thanos:v0.34.0
    args:
      - sidecar
      - '--tsdb.path=/prometheus/data'
      - '--prometheus.url=http://localhost:9090'
      - '--objstore.config-file=/etc/thanos/objstore.yml'
```

### Cortex/Mimir Remote Write

```yaml
# Remote write to Grafana Mimir for long-term storage
remote_write:
  - url: http://mimir:9009/api/v1/push
    queue_config:
      max_samples_per_send: 1000
      batch_send_deadline: 5s
      max_shards: 200
      capacity: 2500
    write_relabel_configs:
      # Only send recording rules and important metrics
      - source_labels: [__name__]
        regex: 'http:.*|slo:.*|node:.*|kube_.*'
        action: keep
```

## Common Anti-Patterns

### High Cardinality

```promql
# BAD: User ID as label (millions of unique values)
http_requests_total{user_id="usr_123"}

# GOOD: Use user_type or plan as label
http_requests_total{plan="enterprise"}

# BAD: Full URL path as label
http_requests_total{path="/api/users/12345"}

# GOOD: Normalized path template
http_requests_total{path="/api/users/:id"}

# BAD: Timestamp or UUID as label
http_requests_total{request_id="550e8400-e29b-41d4-a716-446655440000"}

# GOOD: These belong in logs or traces, not metrics
```

### Incorrect rate() Usage

```promql
# BAD: rate() on a gauge (gauges can decrease)
rate(temperature_celsius[5m])

# GOOD: Use deriv() for gauge rate of change
deriv(temperature_celsius[5m])

# BAD: sum() before rate() on counters
rate(sum(http_requests_total)[5m])

# GOOD: rate() then sum()
sum(rate(http_requests_total[5m]))

# BAD: rate() with too short a range (need at least 4 scrape intervals)
rate(http_requests_total[15s])  # With 15s scrape interval

# GOOD: At least 4x scrape interval
rate(http_requests_total[1m])   # With 15s scrape interval
```

### Missing Labels in histogram_quantile

```promql
# BAD: Missing le in by() clause
histogram_quantile(0.99,
  sum by (service) (rate(http_request_duration_seconds_bucket[5m]))
)

# GOOD: le must be in by() clause
histogram_quantile(0.99,
  sum by (service, le) (rate(http_request_duration_seconds_bucket[5m]))
)
```

## Useful PromQL Patterns

### Detect Stale Metrics

```promql
# Services that stopped reporting metrics
up == 0

# Metrics that haven't updated in 5 minutes
time() - http_requests_total > 300
```

### Rate of Change

```promql
# Predict when disk will be full (linear extrapolation)
predict_linear(node_filesystem_avail_bytes[6h], 24*3600) < 0

# Derivative of a gauge
deriv(temperature_celsius[15m])

# Delta (increase for gauges)
delta(temperature_celsius[1h])
```

### Label Manipulation

```promql
# Replace label values
label_replace(up, "short_instance", "$1", "instance", "(.*):.*")

# Join with info metrics
kube_pod_info * on (pod, namespace) group_left(node) kube_pod_info
```

### Absent Metrics

```promql
# Alert when metric disappears entirely
absent(up{service="api"})

# Alert when no samples for a label value
absent(http_requests_total{service="critical-service"})

# Absent with expected labels
absent(http_requests_total{service="api", environment="production"})
```

## Prometheus Storage

### Retention Configuration

```yaml
# Command-line flags
--storage.tsdb.retention.time=30d     # Time-based retention
--storage.tsdb.retention.size=50GB    # Size-based retention (whichever triggers first)
--storage.tsdb.path=/prometheus/data
--storage.tsdb.wal-compression        # Enable WAL compression
```

### Disk Usage Estimation

```
Disk per sample: ~1-2 bytes (with compression)
Samples per series per day: 5760 (at 15s interval)

Formula:
  daily_bytes = num_series * 5760 * 2 bytes
  monthly_bytes = daily_bytes * 30

Example: 100,000 series
  daily:   100,000 * 5,760 * 2 = ~1.1 GB/day
  monthly: ~33 GB/month
```

## Security Considerations

### Authentication

Prometheus does not natively support authentication. Use:
- Reverse proxy (nginx, Envoy) with basic auth or mTLS
- `--web.config.file` for basic auth and TLS

```yaml
# web.yml (Prometheus basic auth + TLS)
tls_server_config:
  cert_file: /etc/prometheus/tls/server.crt
  key_file: /etc/prometheus/tls/server.key

basic_auth_users:
  admin: $2a$12$hashed_password_here
```

### Network Security

- Never expose Prometheus directly to the internet
- Use TLS for all scrape targets
- Use network policies to restrict access to /metrics endpoints
- Restrict Alertmanager webhook URLs to trusted destinations
