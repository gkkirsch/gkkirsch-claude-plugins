---
name: kubernetes-debugging
description: >
  Debug Kubernetes issues — pod crashes, OOMKilled, ImagePullBackOff, networking
  problems, stuck deployments, resource exhaustion, and log analysis.
  Triggers: "kubernetes debug", "pod crash", "CrashLoopBackOff", "OOMKilled",
  "ImagePullBackOff", "pod pending", "kubectl debug", "k8s troubleshoot".
  NOT for: creating manifests (use kubernetes-deployments), Dockerfiles (use dockerfile-patterns).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Kubernetes Debugging

## Quick Diagnosis Flowchart

```
Pod not running?
│
├── Status: Pending
│   ├── Insufficient resources → Scale cluster or reduce requests
│   ├── Unschedulable (taints) → Add tolerations or untaint node
│   ├── PVC not bound → Check StorageClass, PV availability
│   └── Node selector no match → Check node labels
│
├── Status: ImagePullBackOff / ErrImagePull
│   ├── Image doesn't exist → Verify image:tag in registry
│   ├── Private registry → Create/check imagePullSecrets
│   └── Rate limited → Use authenticated pulls, mirror images
│
├── Status: CrashLoopBackOff
│   ├── Exit code 1 → App error (check logs)
│   ├── Exit code 137 → OOMKilled (increase memory limit)
│   ├── Exit code 127 → Command not found (check image/entrypoint)
│   ├── Exit code 126 → Permission denied (check file permissions)
│   └── Exit code 0 → Container exits normally (check command)
│
├── Status: CreateContainerConfigError
│   ├── ConfigMap not found → Check name matches exactly
│   ├── Secret not found → Check name and namespace
│   └── Key not in ConfigMap/Secret → Check data keys
│
├── Status: Running but not Ready
│   ├── Readiness probe failing → Check probe path/port
│   ├── App not listening → Check container port matches probe
│   └── Dependency not ready → Check depends-on services
│
└── Status: Evicted
    ├── Node disk pressure → Clean up, add ephemeral-storage limits
    ├── Node memory pressure → Right-size pods, add memory limits
    └── Node PID pressure → Fix PID leaks, increase limit
```

## Essential Commands

### Pod Investigation

```bash
# Get pod status with details
kubectl get pods -n <namespace> -o wide

# Full pod description (events, conditions, volumes)
kubectl describe pod <pod> -n <namespace>

# Current logs
kubectl logs <pod> -n <namespace> --tail=100

# Previous crash logs (the crash BEFORE the current restart)
kubectl logs <pod> -n <namespace> --previous

# Logs from specific container (multi-container pods)
kubectl logs <pod> -c <container> -n <namespace>

# Follow logs in real-time
kubectl logs <pod> -n <namespace> -f

# Logs from all pods matching a label
kubectl logs -l app=myapp -n <namespace> --tail=50

# Get pod YAML (see resolved env vars, volumes, etc.)
kubectl get pod <pod> -n <namespace> -o yaml
```

### Interactive Debugging

```bash
# Shell into running pod
kubectl exec -it <pod> -n <namespace> -- /bin/sh
# If sh doesn't exist (distroless):
kubectl exec -it <pod> -n <namespace> -- /bin/bash

# Debug with ephemeral container (K8s 1.25+)
kubectl debug <pod> -n <namespace> -it --image=busybox

# Debug with network tools
kubectl debug <pod> -n <namespace> -it --image=nicolaka/netshoot

# Debug node-level issues
kubectl debug node/<node> -it --image=busybox

# Port forward to pod (bypass service/ingress)
kubectl port-forward <pod> 8080:3000 -n <namespace>

# Port forward to service
kubectl port-forward svc/<service> 8080:80 -n <namespace>
```

### Resource Investigation

```bash
# Pod resource usage (requires metrics-server)
kubectl top pods -n <namespace> --sort-by=memory
kubectl top pods -n <namespace> --sort-by=cpu

# Node resource usage
kubectl top nodes

# Check resource quotas
kubectl describe resourcequota -n <namespace>

# Check limit ranges
kubectl describe limitrange -n <namespace>

# Events (sorted by time, most recent last)
kubectl get events -n <namespace> --sort-by='.lastTimestamp'

# Events for a specific resource
kubectl get events -n <namespace> --field-selector involvedObject.name=<pod>
```

## Debugging Common Issues

### OOMKilled (Exit Code 137)

```bash
# Verify OOMKilled
kubectl describe pod <pod> | grep -A 5 "Last State"
# Look for: Reason: OOMKilled

# Check current memory usage
kubectl top pod <pod>

# Check memory limits
kubectl get pod <pod> -o jsonpath='{.spec.containers[*].resources}'

# Check node memory pressure
kubectl describe node <node> | grep -A 5 "Conditions"
```

**Fixes:**
1. Increase `resources.limits.memory` in deployment
2. Fix memory leak in application
3. Reduce worker thread count / connection pool size
4. Add memory profiling to the app

### CrashLoopBackOff

```bash
# Check restart count and reason
kubectl get pod <pod> -o jsonpath='{.status.containerStatuses[0].restartCount}'
kubectl get pod <pod> -o jsonpath='{.status.containerStatuses[0].lastState}'

# Check previous logs (the crash output)
kubectl logs <pod> --previous

# Check if it's a startup issue
kubectl describe pod <pod> | grep -A 10 "Events"

# Check if the command exists in the image
kubectl run debug --rm -it --image=<your-image> -- which <command>
```

**Common causes:**
- Missing environment variables (check ConfigMap/Secret refs)
- Database not reachable (check service names, network policies)
- File not found (check volume mounts, image contents)
- Permission denied (check security context, file permissions)

### ImagePullBackOff

```bash
# Check the exact error
kubectl describe pod <pod> | grep -A 5 "Events"
# Look for: "Failed to pull image" message

# Verify image exists
docker pull <image>:<tag>

# Check imagePullSecrets
kubectl get pod <pod> -o jsonpath='{.spec.imagePullSecrets}'

# Create registry secret
kubectl create secret docker-registry regcred \
  --docker-server=ghcr.io \
  --docker-username=<user> \
  --docker-password=<token> \
  -n <namespace>
```

### Service Discovery / Networking

```bash
# Check service has endpoints
kubectl get endpoints <service> -n <namespace>
# If endpoints list is empty: selector doesn't match any pods

# Check pod labels match service selector
kubectl get pods -l app=myapp -n <namespace>
kubectl get svc <service> -o jsonpath='{.spec.selector}'

# DNS resolution test
kubectl run dns-test --rm -it --image=busybox -- nslookup <service>.<namespace>.svc.cluster.local

# Full network debug
kubectl run netshoot --rm -it --image=nicolaka/netshoot -- bash
# Inside: curl http://service:port/health
# Inside: dig service.namespace.svc.cluster.local
# Inside: traceroute service

# Check network policies
kubectl get networkpolicy -n <namespace>
kubectl describe networkpolicy <policy> -n <namespace>
```

### Stuck Deployment / Rollout

```bash
# Check rollout status
kubectl rollout status deployment/<name> -n <namespace>

# Check rollout history
kubectl rollout history deployment/<name> -n <namespace>

# Rollback to previous version
kubectl rollout undo deployment/<name> -n <namespace>

# Rollback to specific revision
kubectl rollout undo deployment/<name> --to-revision=3 -n <namespace>

# Check why new pods aren't starting
kubectl get replicasets -n <namespace>
kubectl describe replicaset <new-rs> -n <namespace>

# Force restart all pods (rolling)
kubectl rollout restart deployment/<name> -n <namespace>
```

### PVC Issues

```bash
# Check PVC status
kubectl get pvc -n <namespace>
# STATUS should be "Bound"

# Check why PVC is Pending
kubectl describe pvc <name> -n <namespace>

# Check available PVs
kubectl get pv

# Check storage classes
kubectl get storageclass

# Check if volume is mounted
kubectl describe pod <pod> | grep -A 10 "Volumes"
```

## Debugging Probes

```bash
# Check probe config
kubectl get pod <pod> -o jsonpath='{.spec.containers[0].readinessProbe}'
kubectl get pod <pod> -o jsonpath='{.spec.containers[0].livenessProbe}'

# Test probe endpoint from inside the cluster
kubectl run probe-test --rm -it --image=busybox -- wget -qO- http://<pod-ip>:3000/health

# Check probe failure events
kubectl describe pod <pod> | grep -i "unhealthy\|probe"
```

**Probe debugging checklist:**
1. Is the path correct? (`/health` vs `/healthz` vs `/`)
2. Is the port correct? (containerPort, not service port)
3. Is `initialDelaySeconds` long enough for startup?
4. Does the health endpoint check dependencies that might be slow?
5. Is the app actually listening? (check with port-forward + curl)

## Log Aggregation Patterns

### Stream Logs from Multiple Pods

```bash
# All pods with a label
kubectl logs -l app=myapp -f --max-log-requests=10

# Using stern (better multi-pod logging)
stern myapp -n production --tail 50

# Using kubetail
kubetail myapp -n production
```

### JSON Log Parsing

```bash
# Extract specific fields from JSON logs
kubectl logs <pod> | jq -r 'select(.level == "error") | "\(.timestamp) \(.message)"'

# Count errors by type
kubectl logs <pod> | jq -r '.error_type // empty' | sort | uniq -c | sort -rn

# Find slow requests
kubectl logs <pod> | jq -r 'select(.duration_ms > 1000) | "\(.method) \(.path) \(.duration_ms)ms"'
```

## Resource Debugging Checklist

```
Before deploying to production, verify:

[ ] Resource requests set (CPU + memory) — scheduling depends on this
[ ] Resource limits set (memory required, CPU recommended)
[ ] Readiness probe configured — prevents traffic to unready pods
[ ] Liveness probe configured — restarts stuck processes
[ ] Startup probe configured — prevents false kills during slow startup
[ ] Pod disruption budget set — prevents all pods being killed at once
[ ] Anti-affinity rules — spread pods across nodes
[ ] Security context — runAsNonRoot, read-only filesystem
[ ] Image pull policy — IfNotPresent for tagged images
[ ] Graceful shutdown — terminationGracePeriodSeconds matches app shutdown time
```

## Gotchas

1. **`kubectl logs --previous` only shows the LAST crash** — if you need earlier crashes, you need a log aggregation solution (Loki, CloudWatch, Datadog).

2. **`kubectl top` requires metrics-server** — if you get "metrics not available", install metrics-server in the cluster.

3. **Port-forward uses pod IP, not service** — `kubectl port-forward` bypasses the Service, so you're testing the pod directly. Service-level issues (selector mismatch, port mapping) won't be caught.

4. **Ephemeral debug containers can't see the filesystem** — `kubectl debug` creates a new container in the pod's namespace. It shares network but NOT the filesystem of the target container.

5. **Events expire after 1 hour** — Kubernetes events are garbage-collected. If a pod failed 2 hours ago, the events are gone. Check logs or external monitoring.

6. **DNS resolution takes time after service creation** — CoreDNS caches entries. A newly created service might not resolve for 5-30 seconds. Don't test DNS immediately after creation.

7. **Resource limits != resource usage** — `kubectl top` shows actual usage. `kubectl describe` shows limits. A pod can be OOMKilled even if `top` shows low memory (brief spike).

8. **Rollback doesn't change the image tag** — `kubectl rollout undo` reverts to the previous ReplicaSet's pod template. If both revisions use `:latest`, rollback does nothing because the image hasn't changed.
