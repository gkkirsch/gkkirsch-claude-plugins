---
name: k8s-monitoring
description: >
  Kubernetes monitoring, observability, and alerting patterns.
  Use when setting up Prometheus, Grafana, configuring alerts,
  implementing service meshes, or building observability stacks.
  Triggers: "kubernetes monitoring", "prometheus", "grafana", "k8s alerts",
  "service mesh", "istio", "pod metrics", "cluster monitoring", "SLO", "SLI".
  NOT for: application-level logging (see observability-suite), cloud-specific monitoring (CloudWatch, Datadog).
version: 1.0.0
allowed-tools: Read, Grep, Glob, Bash
---

# Kubernetes Monitoring

## Prometheus Stack Setup

```yaml
# prometheus-stack.yaml — kube-prometheus-stack via Helm
# helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
# helm install monitoring prometheus-community/kube-prometheus-stack -n monitoring -f values.yaml

# values.yaml
prometheus:
  prometheusSpec:
    retention: 30d
    storageSpec:
      volumeClaimTemplate:
        spec:
          storageClassName: gp3
          resources:
            requests:
              storage: 50Gi
    resources:
      requests:
        cpu: 500m
        memory: 2Gi
      limits:
        cpu: 2000m
        memory: 8Gi
    # Scrape all ServiceMonitors in all namespaces
    serviceMonitorSelectorNilUsesHelmValues: false
    podMonitorSelectorNilUsesHelmValues: false

grafana:
  adminPassword: "change-me-in-production"
  persistence:
    enabled: true
    size: 10Gi
  dashboardProviders:
    dashboardproviders.yaml:
      apiVersion: 1
      providers:
        - name: default
          folder: ''
          type: file
          options:
            path: /var/lib/grafana/dashboards/default

alertmanager:
  alertmanagerSpec:
    storage:
      volumeClaimTemplate:
        spec:
          storageClassName: gp3
          resources:
            requests:
              storage: 5Gi
  config:
    global:
      resolve_timeout: 5m
    route:
      receiver: 'slack-notifications'
      group_by: ['alertname', 'namespace']
      group_wait: 30s
      group_interval: 5m
      repeat_interval: 4h
      routes:
        - receiver: 'pagerduty-critical'
          match:
            severity: critical
          repeat_interval: 1h
        - receiver: 'slack-notifications'
          match:
            severity: warning
    receivers:
      - name: 'slack-notifications'
        slack_configs:
          - api_url: 'https://hooks.slack.com/services/...'
            channel: '#alerts'
            title: '{{ .GroupLabels.alertname }}'
            text: '{{ range .Alerts }}{{ .Annotations.summary }}{{ end }}'
      - name: 'pagerduty-critical'
        pagerduty_configs:
          - service_key: 'your-pagerduty-key'
```

## ServiceMonitor for Custom Apps

```yaml
# servicemonitor.yaml — tell Prometheus to scrape your app
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: api-server
  namespace: production
  labels:
    release: monitoring  # Must match Prometheus selector
spec:
  selector:
    matchLabels:
      app: api-server
  endpoints:
    - port: metrics       # Named port from Service
      path: /metrics
      interval: 15s
      scrapeTimeout: 10s
  namespaceSelector:
    matchNames:
      - production
```

## PrometheusRule (Alerting)

```yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: api-server-alerts
  namespace: monitoring
  labels:
    release: monitoring
spec:
  groups:
    - name: api-server.rules
      rules:
        # High error rate
        - alert: HighErrorRate
          expr: |
            sum(rate(http_requests_total{job="api-server", status=~"5.."}[5m]))
            /
            sum(rate(http_requests_total{job="api-server"}[5m]))
            > 0.05
          for: 5m
          labels:
            severity: critical
          annotations:
            summary: "API error rate above 5%"
            description: "{{ $value | humanizePercentage }} of requests are failing"
            runbook_url: "https://wiki.internal/runbooks/high-error-rate"

        # High latency
        - alert: HighLatencyP99
          expr: |
            histogram_quantile(0.99,
              sum(rate(http_request_duration_seconds_bucket{job="api-server"}[5m])) by (le)
            ) > 2
          for: 10m
          labels:
            severity: warning
          annotations:
            summary: "API p99 latency above 2 seconds"
            description: "p99 latency is {{ $value | humanizeDuration }}"

        # Pod restarts
        - alert: HighPodRestarts
          expr: |
            increase(kube_pod_container_status_restarts_total{namespace="production"}[1h]) > 5
          for: 0m
          labels:
            severity: warning
          annotations:
            summary: "Pod {{ $labels.pod }} restarting frequently"
            description: "{{ $value }} restarts in the last hour"

        # Memory pressure
        - alert: ContainerMemoryNearLimit
          expr: |
            container_memory_working_set_bytes{namespace="production"}
            /
            container_spec_memory_limit_bytes{namespace="production"}
            > 0.85
          for: 5m
          labels:
            severity: warning
          annotations:
            summary: "Container {{ $labels.container }} using >85% of memory limit"

        # PVC nearly full
        - alert: PersistentVolumeNearlyFull
          expr: |
            kubelet_volume_stats_used_bytes
            /
            kubelet_volume_stats_capacity_bytes
            > 0.85
          for: 5m
          labels:
            severity: warning
          annotations:
            summary: "PVC {{ $labels.persistentvolumeclaim }} is {{ $value | humanizePercentage }} full"
```

## SLO Definitions

```yaml
# slo.yaml — Service Level Objectives
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: api-server-slos
  namespace: monitoring
spec:
  groups:
    - name: slo.api-server
      rules:
        # SLI: Availability (% of successful requests)
        - record: sli:api_availability:ratio_rate5m
          expr: |
            sum(rate(http_requests_total{job="api-server", status!~"5.."}[5m]))
            /
            sum(rate(http_requests_total{job="api-server"}[5m]))

        # SLI: Latency (% of requests under threshold)
        - record: sli:api_latency:ratio_rate5m
          expr: |
            sum(rate(http_request_duration_seconds_bucket{job="api-server", le="0.5"}[5m]))
            /
            sum(rate(http_request_duration_seconds_count{job="api-server"}[5m]))

        # Error budget: 99.9% availability = 0.1% error budget
        - record: slo:api_error_budget:remaining
          expr: |
            1 - (
              (1 - sli:api_availability:ratio_rate5m)
              / (1 - 0.999)
            )

        # Alert when error budget is burning fast
        - alert: ErrorBudgetBurnRateHigh
          expr: slo:api_error_budget:remaining < 0.5
          for: 5m
          labels:
            severity: critical
          annotations:
            summary: "API error budget more than 50% consumed"
            description: "Error budget remaining: {{ $value | humanizePercentage }}"
```

## Useful PromQL Queries

```bash
# Request rate by status code
sum by (status) (rate(http_requests_total{job="api-server"}[5m]))

# p50/p95/p99 latency
histogram_quantile(0.50, sum(rate(http_request_duration_seconds_bucket[5m])) by (le))
histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket[5m])) by (le))
histogram_quantile(0.99, sum(rate(http_request_duration_seconds_bucket[5m])) by (le))

# CPU usage by pod (percentage of requests)
sum by (pod) (rate(container_cpu_usage_seconds_total{namespace="production"}[5m]))
/
sum by (pod) (container_spec_cpu_quota{namespace="production"} / container_spec_cpu_period{namespace="production"})
* 100

# Memory usage vs limit
container_memory_working_set_bytes{namespace="production"}
/ container_spec_memory_limit_bytes{namespace="production"} * 100

# Top 10 pods by CPU
topk(10, sum by (pod) (rate(container_cpu_usage_seconds_total{namespace="production"}[5m])))

# Disk I/O by pod
sum by (pod) (rate(container_fs_writes_bytes_total{namespace="production"}[5m]))

# Network traffic by pod
sum by (pod) (rate(container_network_receive_bytes_total{namespace="production"}[5m]))
```

## Gotchas

1. **Prometheus retention vs storage** -- Default retention is 15 days. Setting 90-day retention on a busy cluster can consume hundreds of GB. Use recording rules to pre-aggregate high-cardinality metrics, and Thanos/Cortex for long-term storage instead of extending local retention.

2. **Cardinality explosion** -- Adding labels with high cardinality (user IDs, request paths, UUIDs) to metrics creates millions of time series and kills Prometheus. Use histogram buckets for latency, not per-request labels. Keep label cardinality under 1,000 per metric.

3. **Alert fatigue from noisy alerts** -- Alerts that fire every day get ignored. Every alert must have: (1) a `for` duration (not instant), (2) a runbook URL, (3) clear ownership, (4) an actionable response. If you can't define the action, it's a dashboard metric, not an alert.

4. **ServiceMonitor label mismatch** -- If Prometheus doesn't scrape your app, check that the ServiceMonitor's `labels` match the Prometheus `serviceMonitorSelector`. The default kube-prometheus-stack requires `release: monitoring`. Missing this label = silent scrape failure.

5. **histogram_quantile on aggregated data** -- `histogram_quantile` must operate on the raw `_bucket` metric with `rate()` applied first. Applying `sum()` before `rate()` or using the wrong `by()` clause gives mathematically incorrect percentiles. Always: `histogram_quantile(0.99, sum(rate(..._bucket[5m])) by (le))`.

6. **Monitoring the monitoring** -- If Prometheus goes down, you have no alerts. Set up a dead man's switch (a heartbeat alert that fires when Prometheus is NOT sending alerts). Alertmanager should have its own watchdog. Use an external service (PagerDuty heartbeat, Healthchecks.io) as the final safety net.
