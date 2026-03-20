---
name: container-orchestration
description: >
  Container orchestration patterns — Docker multi-stage builds, Docker Compose
  for development, Kubernetes manifests, Helm charts, health checks, resource
  limits, horizontal pod autoscaling, and container security.
  Triggers: "docker", "dockerfile", "docker compose", "kubernetes", "k8s",
  "helm chart", "container", "pod", "deployment manifest".
  NOT for: CI/CD pipeline design or deployment strategies (use deployment-strategies).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Container Orchestration

## Multi-Stage Dockerfile (Node.js)

```dockerfile
# Stage 1: Dependencies
FROM node:20-alpine AS deps
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci --production=false

# Stage 2: Build
FROM node:20-alpine AS build
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .
RUN npm run build
# Prune dev dependencies after build
RUN npm prune --production

# Stage 3: Production
FROM node:20-alpine AS production
RUN apk add --no-cache dumb-init
# Non-root user
RUN addgroup -g 1001 appgroup && adduser -u 1001 -G appgroup -s /bin/sh -D appuser

WORKDIR /app
COPY --from=build --chown=appuser:appgroup /app/dist ./dist
COPY --from=build --chown=appuser:appgroup /app/node_modules ./node_modules
COPY --from=build --chown=appuser:appgroup /app/package.json ./

USER appuser
EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
  CMD wget --quiet --tries=1 --spider http://localhost:8080/health || exit 1

ENTRYPOINT ["dumb-init", "--"]
CMD ["node", "dist/server.js"]
```

## Multi-Stage Dockerfile (Go)

```dockerfile
FROM golang:1.22-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /server ./cmd/server

FROM scratch
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /server /server

EXPOSE 8080
ENTRYPOINT ["/server"]
```

## Docker Compose (Development)

```yaml
# docker-compose.yml
services:
  app:
    build:
      context: .
      dockerfile: Dockerfile
      target: deps  # Stop at deps stage for dev
    command: npm run dev
    volumes:
      - .:/app
      - /app/node_modules  # Exclude node_modules from mount
    ports:
      - "3000:3000"
    environment:
      NODE_ENV: development
      DATABASE_URL: postgres://postgres:postgres@db:5432/myapp_dev
      REDIS_URL: redis://redis:6379
    depends_on:
      db:
        condition: service_healthy
      redis:
        condition: service_started
    healthcheck:
      test: ["CMD", "wget", "--spider", "-q", "http://localhost:3000/health"]
      interval: 10s
      timeout: 5s
      retries: 3

  db:
    image: postgres:16-alpine
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./scripts/init-db.sql:/docker-entrypoint-initdb.d/init.sql
    environment:
      POSTGRES_DB: myapp_dev
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
    ports:
      - "5432:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 5s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data

  worker:
    build:
      context: .
      target: deps
    command: npm run worker
    volumes:
      - .:/app
      - /app/node_modules
    environment:
      DATABASE_URL: postgres://postgres:postgres@db:5432/myapp_dev
      REDIS_URL: redis://redis:6379
    depends_on:
      - db
      - redis

  mailhog:
    image: mailhog/mailhog
    ports:
      - "8025:8025"  # Web UI
      - "1025:1025"  # SMTP

volumes:
  postgres_data:
  redis_data:
```

## Kubernetes Deployment

```yaml
# k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-server
  labels:
    app: api-server
    version: v1
spec:
  replicas: 3
  selector:
    matchLabels:
      app: api-server
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  template:
    metadata:
      labels:
        app: api-server
        version: v1
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "8080"
        prometheus.io/path: "/metrics"
    spec:
      serviceAccountName: api-server
      securityContext:
        runAsNonRoot: true
        runAsUser: 1001
        fsGroup: 1001
      containers:
        - name: api-server
          image: registry.example.com/api-server:1.2.3
          imagePullPolicy: IfNotPresent
          ports:
            - name: http
              containerPort: 8080
              protocol: TCP
          env:
            - name: NODE_ENV
              value: "production"
            - name: PORT
              value: "8080"
            - name: DATABASE_URL
              valueFrom:
                secretKeyRef:
                  name: api-secrets
                  key: database-url
            - name: REDIS_URL
              valueFrom:
                configMapKeyRef:
                  name: api-config
                  key: redis-url
          resources:
            requests:
              cpu: 250m
              memory: 256Mi
            limits:
              cpu: 500m
              memory: 512Mi
          readinessProbe:
            httpGet:
              path: /health/ready
              port: http
            initialDelaySeconds: 5
            periodSeconds: 10
            failureThreshold: 3
          livenessProbe:
            httpGet:
              path: /health/live
              port: http
            initialDelaySeconds: 15
            periodSeconds: 20
            failureThreshold: 3
          startupProbe:
            httpGet:
              path: /health/live
              port: http
            initialDelaySeconds: 5
            periodSeconds: 5
            failureThreshold: 30  # 5s * 30 = 150s max startup
          volumeMounts:
            - name: tmp
              mountPath: /tmp
      volumes:
        - name: tmp
          emptyDir: {}
      topologySpreadConstraints:
        - maxSkew: 1
          topologyKey: topology.kubernetes.io/zone
          whenUnsatisfiable: DoNotSchedule
          labelSelector:
            matchLabels:
              app: api-server
---
# k8s/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: api-server
spec:
  selector:
    app: api-server
  ports:
    - name: http
      port: 80
      targetPort: http
  type: ClusterIP
---
# k8s/hpa.yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: api-server
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: api-server
  minReplicas: 3
  maxReplicas: 20
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 70
    - type: Resource
      resource:
        name: memory
        target:
          type: Utilization
          averageUtilization: 80
    - type: Pods
      pods:
        metric:
          name: http_requests_per_second
        target:
          type: AverageValue
          averageValue: "100"
  behavior:
    scaleUp:
      stabilizationWindowSeconds: 60
      policies:
        - type: Percent
          value: 50
          periodSeconds: 60
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
        - type: Percent
          value: 10
          periodSeconds: 60
```

## ConfigMap and Secrets

```yaml
# k8s/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: api-config
data:
  redis-url: "redis://redis-master:6379"
  log-level: "info"
  cors-origins: "https://app.example.com,https://admin.example.com"
---
# k8s/secret.yaml (use sealed-secrets or external-secrets in production)
apiVersion: v1
kind: Secret
metadata:
  name: api-secrets
type: Opaque
stringData:
  database-url: "postgres://user:pass@db-host:5432/myapp"
  jwt-secret: "your-jwt-secret-here"
---
# External Secrets Operator (production pattern)
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: api-secrets
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: aws-secrets-manager
    kind: ClusterSecretStore
  target:
    name: api-secrets
  data:
    - secretKey: database-url
      remoteRef:
        key: production/api/database-url
    - secretKey: jwt-secret
      remoteRef:
        key: production/api/jwt-secret
```

## Helm Chart Structure

```yaml
# charts/api-server/Chart.yaml
apiVersion: v2
name: api-server
description: API server Helm chart
version: 1.0.0
appVersion: "1.2.3"

# charts/api-server/values.yaml
replicaCount: 3

image:
  repository: registry.example.com/api-server
  tag: "1.2.3"
  pullPolicy: IfNotPresent

service:
  type: ClusterIP
  port: 80

ingress:
  enabled: true
  className: nginx
  hosts:
    - host: api.example.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: api-tls
      hosts:
        - api.example.com

resources:
  requests:
    cpu: 250m
    memory: 256Mi
  limits:
    cpu: 500m
    memory: 512Mi

autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 20
  targetCPUUtilizationPercentage: 70

env:
  NODE_ENV: production

secrets:
  databaseUrl: ""  # Override in values-production.yaml
  jwtSecret: ""

# charts/api-server/templates/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "api-server.fullname" . }}
  labels:
    {{- include "api-server.labels" . | nindent 4 }}
spec:
  {{- if not .Values.autoscaling.enabled }}
  replicas: {{ .Values.replicaCount }}
  {{- end }}
  selector:
    matchLabels:
      {{- include "api-server.selectorLabels" . | nindent 6 }}
  template:
    metadata:
      labels:
        {{- include "api-server.selectorLabels" . | nindent 8 }}
    spec:
      containers:
        - name: {{ .Chart.Name }}
          image: "{{ .Values.image.repository }}:{{ .Values.image.tag }}"
          imagePullPolicy: {{ .Values.image.pullPolicy }}
          ports:
            - name: http
              containerPort: 8080
          env:
            - name: NODE_ENV
              value: {{ .Values.env.NODE_ENV | quote }}
            - name: DATABASE_URL
              valueFrom:
                secretKeyRef:
                  name: {{ include "api-server.fullname" . }}
                  key: database-url
          resources:
            {{- toYaml .Values.resources | nindent 12 }}
          readinessProbe:
            httpGet:
              path: /health/ready
              port: http
            initialDelaySeconds: 5
          livenessProbe:
            httpGet:
              path: /health/live
              port: http
            initialDelaySeconds: 15
```

## Health Check Patterns

```typescript
// src/health.ts
import { Router } from "express";
import { Pool } from "pg";
import Redis from "ioredis";

const router = Router();

interface HealthStatus {
  status: "healthy" | "degraded" | "unhealthy";
  checks: Record<string, {
    status: "pass" | "fail";
    latency?: number;
    message?: string;
  }>;
  uptime: number;
  version: string;
}

// Liveness — is the process alive?
router.get("/health/live", (req, res) => {
  res.json({ status: "ok" });
});

// Readiness — can it serve traffic?
router.get("/health/ready", async (req, res) => {
  const checks: HealthStatus["checks"] = {};

  // Database check
  const dbStart = Date.now();
  try {
    await pool.query("SELECT 1");
    checks.database = { status: "pass", latency: Date.now() - dbStart };
  } catch (err) {
    checks.database = {
      status: "fail",
      message: (err as Error).message,
      latency: Date.now() - dbStart,
    };
  }

  // Redis check
  const redisStart = Date.now();
  try {
    await redis.ping();
    checks.redis = { status: "pass", latency: Date.now() - redisStart };
  } catch (err) {
    checks.redis = {
      status: "fail",
      message: (err as Error).message,
      latency: Date.now() - redisStart,
    };
  }

  const allPassing = Object.values(checks).every((c) => c.status === "pass");
  const status: HealthStatus = {
    status: allPassing ? "healthy" : "unhealthy",
    checks,
    uptime: process.uptime(),
    version: process.env.APP_VERSION || "unknown",
  };

  res.status(allPassing ? 200 : 503).json(status);
});
```

## Container Security

```dockerfile
# Security-hardened Dockerfile
FROM node:20-alpine AS production

# Don't run as root
RUN addgroup -g 1001 appgroup && \
    adduser -u 1001 -G appgroup -s /bin/sh -D appuser

# Remove unnecessary packages
RUN apk del --purge apk-tools && \
    rm -rf /var/cache/apk/* /tmp/* /var/tmp/*

WORKDIR /app
COPY --from=build --chown=appuser:appgroup /app/dist ./dist
COPY --from=build --chown=appuser:appgroup /app/node_modules ./node_modules

# Read-only filesystem where possible
RUN chmod -R 555 /app/dist

USER appuser

# Drop all capabilities
# (requires --cap-drop=ALL in docker run or securityContext in k8s)
```

```yaml
# k8s security context
spec:
  securityContext:
    runAsNonRoot: true
    runAsUser: 1001
    fsGroup: 1001
    seccompProfile:
      type: RuntimeDefault
  containers:
    - name: app
      securityContext:
        allowPrivilegeEscalation: false
        readOnlyRootFilesystem: true
        capabilities:
          drop: ["ALL"]
      volumeMounts:
        - name: tmp
          mountPath: /tmp
  volumes:
    - name: tmp
      emptyDir:
        sizeLimit: 100Mi
```

## Gotchas

1. **Docker layer caching invalidation** — `COPY . .` before `npm ci` busts the dependency cache on every code change. Always copy `package.json` and `package-lock.json` first, run `npm ci`, then copy the rest. The build layer ordering in a Dockerfile directly controls cache efficiency.

2. **Alpine DNS resolution in Kubernetes** — Alpine uses musl libc, which handles DNS differently than glibc. Under high load, DNS lookups can fail or be extremely slow. Fix with `RUN echo "options ndots:0" >> /etc/resolv.conf` or switch to Debian-slim base images for production workloads.

3. **Container memory limits vs Node.js heap** — A container with `limits.memory: 512Mi` will be OOM-killed if Node.js tries to use more. Set `--max-old-space-size` to ~75% of the container memory limit: `CMD ["node", "--max-old-space-size=384", "dist/server.js"]`. Without this, Node.js defaults to ~1.5GB heap regardless of container limits.

4. **Liveness vs readiness probe confusion** — Liveness checks if the process is alive (restart if dead). Readiness checks if it can serve traffic (remove from load balancer if not ready). Checking the database in the liveness probe means a DB outage cascading-restarts all pods. Only check external dependencies in readiness probes.

5. **`latest` tag in production** — `image: myapp:latest` means Kubernetes doesn't know when the image changed and won't pull updates. Always use immutable tags (commit SHA or semver). Set `imagePullPolicy: IfNotPresent` with specific tags to avoid re-pulling on every pod restart.

6. **Docker Compose volumes mask build artifacts** — Mounting `.:/app` in development replaces everything including `node_modules` from the build stage. The anonymous volume trick (`/app/node_modules`) prevents this, but you must rebuild after adding new dependencies. Run `docker compose build` after `npm install` changes.
