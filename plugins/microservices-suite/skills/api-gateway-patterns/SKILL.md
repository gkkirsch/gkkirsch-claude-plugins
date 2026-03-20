---
name: api-gateway-patterns
description: >
  API gateway and microservice communication — routing, rate limiting,
  circuit breakers, service discovery, health checks, and inter-service auth.
  Triggers: "api gateway", "circuit breaker", "service discovery",
  "rate limiting", "health check", "microservice communication",
  "service mesh", "load balancing".
  NOT for: Event-driven or async patterns (use event-driven-architecture).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# API Gateway & Microservice Communication

## API Gateway Pattern

```typescript
// Express-based API gateway
import express from "express";
import { createProxyMiddleware } from "http-proxy-middleware";
import rateLimit from "express-rate-limit";

const app = express();

// Rate limiting per client
const limiter = rateLimit({
  windowMs: 60 * 1000,
  max: 100,
  keyGenerator: (req) => req.headers["x-api-key"] as string || req.ip,
  standardHeaders: true,
});
app.use(limiter);

// Route to microservices
app.use("/api/users", createProxyMiddleware({
  target: process.env.USER_SERVICE_URL,
  pathRewrite: { "^/api/users": "" },
  changeOrigin: true,
}));

app.use("/api/orders", createProxyMiddleware({
  target: process.env.ORDER_SERVICE_URL,
  pathRewrite: { "^/api/orders": "" },
  changeOrigin: true,
}));

app.use("/api/products", createProxyMiddleware({
  target: process.env.PRODUCT_SERVICE_URL,
  pathRewrite: { "^/api/products": "" },
  changeOrigin: true,
}));

// Health check aggregation
app.get("/health", async (_, res) => {
  const services = ["USER_SERVICE_URL", "ORDER_SERVICE_URL", "PRODUCT_SERVICE_URL"];
  const checks = await Promise.allSettled(
    services.map((env) =>
      fetch(`${process.env[env]}/health`, { signal: AbortSignal.timeout(3000) })
    )
  );

  const status = checks.every((c) => c.status === "fulfilled" && c.value.ok) ? "healthy" : "degraded";
  res.status(status === "healthy" ? 200 : 503).json({
    status,
    services: Object.fromEntries(
      services.map((s, i) => [s, checks[i].status === "fulfilled" ? "up" : "down"])
    ),
  });
});
```

## Circuit Breaker

```typescript
enum CircuitState { CLOSED, OPEN, HALF_OPEN }

class CircuitBreaker {
  private state = CircuitState.CLOSED;
  private failureCount = 0;
  private lastFailure = 0;
  private successCount = 0;

  constructor(
    private readonly threshold: number = 5,      // Failures before opening
    private readonly timeout: number = 30_000,   // Time before half-open
    private readonly halfOpenMax: number = 3,     // Successes to close
  ) {}

  async call<T>(fn: () => Promise<T>, fallback?: () => T): Promise<T> {
    if (this.state === CircuitState.OPEN) {
      if (Date.now() - this.lastFailure > this.timeout) {
        this.state = CircuitState.HALF_OPEN;
        this.successCount = 0;
      } else {
        if (fallback) return fallback();
        throw new Error("Circuit breaker is OPEN");
      }
    }

    try {
      const result = await fn();
      this.onSuccess();
      return result;
    } catch (error) {
      this.onFailure();
      if (fallback) return fallback();
      throw error;
    }
  }

  private onSuccess() {
    if (this.state === CircuitState.HALF_OPEN) {
      this.successCount++;
      if (this.successCount >= this.halfOpenMax) {
        this.state = CircuitState.CLOSED;
        this.failureCount = 0;
      }
    } else {
      this.failureCount = 0;
    }
  }

  private onFailure() {
    this.failureCount++;
    this.lastFailure = Date.now();
    if (this.failureCount >= this.threshold) {
      this.state = CircuitState.OPEN;
    }
  }
}

// Usage
const userServiceBreaker = new CircuitBreaker(5, 30_000);

async function getUser(id: string) {
  return userServiceBreaker.call(
    () => fetch(`${USER_SERVICE_URL}/users/${id}`).then((r) => r.json()),
    () => ({ id, name: "Unknown", cached: true }) // Fallback
  );
}
```

## Service Communication

### HTTP Client with Retry

```typescript
async function serviceCall<T>(
  url: string,
  options: RequestInit = {},
  retries = 3,
  backoffMs = 1000
): Promise<T> {
  for (let attempt = 1; attempt <= retries; attempt++) {
    try {
      const response = await fetch(url, {
        ...options,
        headers: {
          "Content-Type": "application/json",
          "X-Request-Id": crypto.randomUUID(),
          "X-Service-Name": process.env.SERVICE_NAME!,
          ...options.headers,
        },
        signal: AbortSignal.timeout(5000),
      });

      if (response.status >= 500 && attempt < retries) {
        await new Promise((r) => setTimeout(r, backoffMs * attempt));
        continue;
      }

      if (!response.ok) {
        throw new ServiceError(response.status, await response.text());
      }

      return response.json();
    } catch (error) {
      if (attempt === retries) throw error;
      if (error instanceof TypeError) {
        // Network error — retry
        await new Promise((r) => setTimeout(r, backoffMs * attempt));
        continue;
      }
      throw error; // Non-retryable error
    }
  }
  throw new Error("Exhausted retries");
}
```

### gRPC Service Definition

```protobuf
// user.proto
syntax = "proto3";

package user;

service UserService {
  rpc GetUser (GetUserRequest) returns (UserResponse);
  rpc ListUsers (ListUsersRequest) returns (ListUsersResponse);
  rpc CreateUser (CreateUserRequest) returns (UserResponse);
}

message GetUserRequest { string id = 1; }
message CreateUserRequest { string name = 1; string email = 2; }
message UserResponse { string id = 1; string name = 2; string email = 3; }
message ListUsersRequest { int32 page = 1; int32 limit = 2; }
message ListUsersResponse { repeated UserResponse users = 1; int32 total = 2; }
```

```typescript
// gRPC client
import * as grpc from "@grpc/grpc-js";
import * as protoLoader from "@grpc/proto-loader";

const packageDef = protoLoader.loadSync("./proto/user.proto");
const proto = grpc.loadPackageDefinition(packageDef) as any;

const client = new proto.user.UserService(
  process.env.USER_SERVICE_GRPC!,
  grpc.credentials.createInsecure()
);

// Promisified call
function getUser(id: string): Promise<User> {
  return new Promise((resolve, reject) => {
    client.GetUser({ id }, (err: Error | null, response: User) => {
      if (err) reject(err);
      else resolve(response);
    });
  });
}
```

## Health Checks

```typescript
// Comprehensive health check endpoint
app.get("/health", async (_, res) => {
  const checks = {
    database: checkDatabase(),
    redis: checkRedis(),
    diskSpace: checkDiskSpace(),
    memory: checkMemory(),
  };

  const results = await Promise.allSettled(Object.values(checks));
  const names = Object.keys(checks);

  const health: Record<string, { status: string; latency?: number; error?: string }> = {};
  let overall = true;

  results.forEach((result, i) => {
    if (result.status === "fulfilled") {
      health[names[i]] = result.value;
      if (result.value.status !== "healthy") overall = false;
    } else {
      health[names[i]] = { status: "unhealthy", error: result.reason.message };
      overall = false;
    }
  });

  res.status(overall ? 200 : 503).json({
    status: overall ? "healthy" : "unhealthy",
    uptime: process.uptime(),
    timestamp: new Date().toISOString(),
    checks: health,
  });
});

async function checkDatabase(): Promise<{ status: string; latency: number }> {
  const start = Date.now();
  await db.query("SELECT 1");
  return { status: "healthy", latency: Date.now() - start };
}

async function checkRedis(): Promise<{ status: string; latency: number }> {
  const start = Date.now();
  await redis.ping();
  return { status: "healthy", latency: Date.now() - start };
}
```

### Kubernetes Health Probes

```yaml
# deployment.yaml
spec:
  containers:
    - name: api
      livenessProbe:        # Is the process alive?
        httpGet:
          path: /health/live
          port: 3000
        initialDelaySeconds: 15
        periodSeconds: 10
      readinessProbe:       # Can it serve traffic?
        httpGet:
          path: /health/ready
          port: 3000
        initialDelaySeconds: 5
        periodSeconds: 5
      startupProbe:         # Has it started?
        httpGet:
          path: /health/startup
          port: 3000
        failureThreshold: 30
        periodSeconds: 2
```

## Inter-Service Authentication

```typescript
// JWT-based service-to-service auth
function generateServiceToken(serviceName: string): string {
  return jwt.sign(
    { sub: serviceName, type: "service" },
    process.env.SERVICE_JWT_SECRET!,
    { expiresIn: "5m", issuer: process.env.SERVICE_NAME }
  );
}

// Middleware — accept both user and service tokens
function requireServiceAuth(req: Request, res: Response, next: NextFunction) {
  const token = req.headers["x-service-token"] as string;
  if (!token) return res.status(401).json({ error: "Service token required" });

  try {
    const payload = jwt.verify(token, process.env.SERVICE_JWT_SECRET!);
    if (payload.type !== "service") {
      return res.status(403).json({ error: "Invalid token type" });
    }
    req.service = payload.sub;
    next();
  } catch {
    res.status(401).json({ error: "Invalid service token" });
  }
}
```

## Gotchas

1. **Cascading failures without circuit breakers** — Service A calls B, B calls C. If C is down, B times out, which makes A time out, which kills all user requests. Circuit breakers on every inter-service call prevent cascading failure.

2. **Retry storms amplify outages** — If a service is overloaded and 100 clients all retry 3 times, the service gets 300 requests instead of 100. Use exponential backoff with jitter: `delay * 2^attempt + random(0, delay)`.

3. **Health checks that lie** — A health endpoint returning 200 while the database is down means the load balancer keeps routing traffic. Health checks must verify ALL critical dependencies (DB, cache, queues).

4. **Synchronous call chains create tight coupling** — `API → User Service → Auth Service → DB` is a chain where any link can break the whole request. Use async events where possible. Reserve synchronous calls for queries that truly need immediate responses.

5. **Missing request IDs** — Without a correlation ID flowing through all services, debugging a distributed failure means searching logs across every service with only a timestamp to go on. Generate a request ID at the gateway and propagate it through all calls.

6. **Service discovery hardcoded** — Environment variables for service URLs work for 3 services. At 20 services, you need a registry (Consul, Kubernetes DNS) or service mesh. Plan for this early.
