---
name: api-test-patterns
description: >
  API testing patterns — REST and GraphQL test design, request builders,
  assertion helpers, authentication testing, error scenario coverage,
  performance baseline tests, and test data management.
  Triggers: "api test", "rest test", "endpoint test", "integration test api",
  "test api endpoint", "http test", "supertest", "api assertion".
  NOT for: Contract testing or consumer-driven contracts (use contract-testing).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# API Test Patterns

## Test Structure (Jest + Supertest)

```typescript
// tests/api/users.test.ts
import request from "supertest";
import { app } from "../../src/app";
import { db } from "../../src/db";
import { createTestUser, generateAuthToken } from "../helpers";

describe("GET /api/users", () => {
  let authToken: string;

  beforeAll(async () => {
    await db.migrate.latest();
  });

  beforeEach(async () => {
    await db("users").truncate();
    const user = await createTestUser({ role: "admin" });
    authToken = generateAuthToken(user);
  });

  afterAll(async () => {
    await db.destroy();
  });

  it("returns paginated users", async () => {
    // Seed 25 users
    await Promise.all(
      Array.from({ length: 25 }, (_, i) =>
        createTestUser({ email: `user${i}@test.com` })
      )
    );

    const res = await request(app)
      .get("/api/users")
      .set("Authorization", `Bearer ${authToken}`)
      .query({ page: 1, limit: 10 })
      .expect(200);

    expect(res.body).toMatchObject({
      data: expect.arrayContaining([
        expect.objectContaining({
          id: expect.any(String),
          email: expect.any(String),
          createdAt: expect.any(String),
        }),
      ]),
      pagination: {
        page: 1,
        limit: 10,
        total: expect.any(Number),
        totalPages: 3,
      },
    });
    expect(res.body.data).toHaveLength(10);
  });

  it("filters by role", async () => {
    await createTestUser({ email: "admin@test.com", role: "admin" });
    await createTestUser({ email: "viewer@test.com", role: "viewer" });

    const res = await request(app)
      .get("/api/users")
      .set("Authorization", `Bearer ${authToken}`)
      .query({ role: "admin" })
      .expect(200);

    res.body.data.forEach((user: any) => {
      expect(user.role).toBe("admin");
    });
  });

  it("returns 401 without auth token", async () => {
    const res = await request(app).get("/api/users").expect(401);

    expect(res.body).toMatchObject({
      error: "Unauthorized",
      message: expect.any(String),
    });
  });

  it("returns 403 for non-admin users", async () => {
    const viewer = await createTestUser({ role: "viewer" });
    const viewerToken = generateAuthToken(viewer);

    await request(app)
      .get("/api/users")
      .set("Authorization", `Bearer ${viewerToken}`)
      .expect(403);
  });
});
```

## Request Builder Pattern

```typescript
// tests/helpers/request-builder.ts
import request, { Test } from "supertest";
import { app } from "../../src/app";

export class ApiRequest {
  private req: Test;
  private headers: Record<string, string> = {};

  constructor(method: "get" | "post" | "put" | "patch" | "delete", path: string) {
    this.req = request(app)[method](path);
  }

  static get(path: string) { return new ApiRequest("get", path); }
  static post(path: string) { return new ApiRequest("post", path); }
  static put(path: string) { return new ApiRequest("put", path); }
  static patch(path: string) { return new ApiRequest("patch", path); }
  static delete(path: string) { return new ApiRequest("delete", path); }

  auth(token: string) {
    this.headers["Authorization"] = `Bearer ${token}`;
    return this;
  }

  apiKey(key: string) {
    this.headers["X-API-Key"] = key;
    return this;
  }

  as(user: { token: string }) {
    return this.auth(user.token);
  }

  json(body: Record<string, any>) {
    this.req.send(body);
    return this;
  }

  query(params: Record<string, any>) {
    this.req.query(params);
    return this;
  }

  async send() {
    Object.entries(this.headers).forEach(([k, v]) => this.req.set(k, v));
    return this.req;
  }

  async expect(status: number) {
    Object.entries(this.headers).forEach(([k, v]) => this.req.set(k, v));
    return this.req.expect(status);
  }
}

// Usage
const res = await ApiRequest.post("/api/orders")
  .as(adminUser)
  .json({ productId: "abc", quantity: 2 })
  .expect(201);
```

## CRUD Endpoint Test Template

```typescript
// tests/api/resources.test.ts — generic pattern for any REST resource
describe("POST /api/resources", () => {
  const validPayload = {
    name: "Test Resource",
    type: "standard",
    config: { retries: 3 },
  };

  it("creates resource with valid data", async () => {
    const res = await request(app)
      .post("/api/resources")
      .set("Authorization", `Bearer ${token}`)
      .send(validPayload)
      .expect(201);

    expect(res.body).toMatchObject({
      id: expect.any(String),
      ...validPayload,
      createdAt: expect.any(String),
    });

    // Verify persistence
    const fetched = await request(app)
      .get(`/api/resources/${res.body.id}`)
      .set("Authorization", `Bearer ${token}`)
      .expect(200);

    expect(fetched.body.name).toBe(validPayload.name);
  });

  it("returns 400 for missing required fields", async () => {
    const res = await request(app)
      .post("/api/resources")
      .set("Authorization", `Bearer ${token}`)
      .send({ type: "standard" }) // missing name
      .expect(400);

    expect(res.body.errors).toEqual(
      expect.arrayContaining([
        expect.objectContaining({
          field: "name",
          message: expect.any(String),
        }),
      ])
    );
  });

  it("returns 409 for duplicate name", async () => {
    await request(app)
      .post("/api/resources")
      .set("Authorization", `Bearer ${token}`)
      .send(validPayload)
      .expect(201);

    await request(app)
      .post("/api/resources")
      .set("Authorization", `Bearer ${token}`)
      .send(validPayload)
      .expect(409);
  });

  it("sanitizes input (strips unknown fields)", async () => {
    const res = await request(app)
      .post("/api/resources")
      .set("Authorization", `Bearer ${token}`)
      .send({ ...validPayload, isAdmin: true, __proto__: {} })
      .expect(201);

    expect(res.body.isAdmin).toBeUndefined();
  });
});

describe("PUT /api/resources/:id", () => {
  it("updates existing resource", async () => {
    const created = await createResource(validPayload);

    const res = await request(app)
      .put(`/api/resources/${created.id}`)
      .set("Authorization", `Bearer ${token}`)
      .send({ name: "Updated Name" })
      .expect(200);

    expect(res.body.name).toBe("Updated Name");
  });

  it("returns 404 for non-existent resource", async () => {
    await request(app)
      .put("/api/resources/non-existent-id")
      .set("Authorization", `Bearer ${token}`)
      .send({ name: "Updated" })
      .expect(404);
  });

  it("returns 403 when updating another user's resource", async () => {
    const otherUser = await createTestUser({ email: "other@test.com" });
    const resource = await createResource(validPayload, otherUser);

    await request(app)
      .put(`/api/resources/${resource.id}`)
      .set("Authorization", `Bearer ${token}`)
      .send({ name: "Hijacked" })
      .expect(403);
  });
});

describe("DELETE /api/resources/:id", () => {
  it("soft-deletes resource", async () => {
    const resource = await createResource(validPayload);

    await request(app)
      .delete(`/api/resources/${resource.id}`)
      .set("Authorization", `Bearer ${token}`)
      .expect(204);

    // Verify not returned in list
    const list = await request(app)
      .get("/api/resources")
      .set("Authorization", `Bearer ${token}`)
      .expect(200);

    expect(list.body.data.find((r: any) => r.id === resource.id)).toBeUndefined();
  });

  it("is idempotent", async () => {
    const resource = await createResource(validPayload);

    await request(app)
      .delete(`/api/resources/${resource.id}`)
      .set("Authorization", `Bearer ${token}`)
      .expect(204);

    // Second delete also succeeds (idempotent)
    await request(app)
      .delete(`/api/resources/${resource.id}`)
      .set("Authorization", `Bearer ${token}`)
      .expect(204);
  });
});
```

## Error Scenario Testing

```typescript
// tests/api/error-scenarios.test.ts

describe("Error handling", () => {
  it("returns structured errors for validation failures", async () => {
    const res = await request(app)
      .post("/api/users")
      .set("Authorization", `Bearer ${token}`)
      .send({ email: "not-an-email", password: "short" })
      .expect(400);

    expect(res.body).toMatchObject({
      error: "ValidationError",
      message: expect.any(String),
      errors: expect.arrayContaining([
        { field: "email", message: expect.stringContaining("valid email") },
        { field: "password", message: expect.stringContaining("8 characters") },
      ]),
    });
  });

  it("handles malformed JSON gracefully", async () => {
    const res = await request(app)
      .post("/api/users")
      .set("Content-Type", "application/json")
      .set("Authorization", `Bearer ${token}`)
      .send("{ invalid json")
      .expect(400);

    expect(res.body.error).toBe("BadRequest");
  });

  it("returns 415 for unsupported content type", async () => {
    await request(app)
      .post("/api/users")
      .set("Content-Type", "text/xml")
      .set("Authorization", `Bearer ${token}`)
      .send("<user><email>test@test.com</email></user>")
      .expect(415);
  });

  it("returns 429 when rate limited", async () => {
    const requests = Array.from({ length: 101 }, () =>
      request(app)
        .get("/api/health")
        .set("X-Forwarded-For", "1.2.3.4")
    );

    const responses = await Promise.all(requests);
    const rateLimited = responses.filter((r) => r.status === 429);
    expect(rateLimited.length).toBeGreaterThan(0);

    const limited = rateLimited[0];
    expect(limited.headers["retry-after"]).toBeDefined();
  });

  it("handles concurrent conflicting updates", async () => {
    const resource = await createResource({ name: "Original", version: 1 });

    const [res1, res2] = await Promise.all([
      request(app)
        .put(`/api/resources/${resource.id}`)
        .set("Authorization", `Bearer ${token}`)
        .send({ name: "Update A", version: 1 }),
      request(app)
        .put(`/api/resources/${resource.id}`)
        .set("Authorization", `Bearer ${token}`)
        .send({ name: "Update B", version: 1 }),
    ]);

    // One succeeds, one gets conflict
    const statuses = [res1.status, res2.status].sort();
    expect(statuses).toEqual([200, 409]);
  });
});
```

## Test Data Factories

```typescript
// tests/helpers/factories.ts
import { faker } from "@faker-js/faker";
import { db } from "../../src/db";
import jwt from "jsonwebtoken";

interface UserOverrides {
  email?: string;
  role?: "admin" | "viewer" | "editor";
  name?: string;
}

export async function createTestUser(overrides: UserOverrides = {}) {
  const user = {
    id: faker.string.uuid(),
    email: overrides.email ?? faker.internet.email(),
    name: overrides.name ?? faker.person.fullName(),
    role: overrides.role ?? "viewer",
    passwordHash: "$2b$10$fixedhashfortesting",
    createdAt: new Date().toISOString(),
  };

  await db("users").insert(user);
  return { ...user, token: generateAuthToken(user) };
}

export function generateAuthToken(user: { id: string; role: string }) {
  return jwt.sign(
    { sub: user.id, role: user.role },
    process.env.JWT_SECRET ?? "test-secret",
    { expiresIn: "1h" }
  );
}

export async function createResource(
  data: Record<string, any>,
  owner?: { id: string }
) {
  const resource = {
    id: faker.string.uuid(),
    ...data,
    ownerId: owner?.id ?? "default-test-user",
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  };
  await db("resources").insert(resource);
  return resource;
}

// Bulk factory for performance tests
export async function seedUsers(count: number) {
  const users = Array.from({ length: count }, (_, i) => ({
    id: faker.string.uuid(),
    email: `perf-user-${i}@test.com`,
    name: faker.person.fullName(),
    role: "viewer",
    passwordHash: "$2b$10$fixedhash",
    createdAt: new Date().toISOString(),
  }));

  // Batch insert for speed
  const batchSize = 500;
  for (let i = 0; i < users.length; i += batchSize) {
    await db("users").insert(users.slice(i, i + batchSize));
  }
  return users;
}
```

## Authentication Flow Testing

```typescript
// tests/api/auth.test.ts

describe("Authentication flows", () => {
  describe("POST /api/auth/login", () => {
    it("returns tokens for valid credentials", async () => {
      await createTestUser({
        email: "user@test.com",
        // password is "TestPass123!" via bcrypt
      });

      const res = await request(app)
        .post("/api/auth/login")
        .send({ email: "user@test.com", password: "TestPass123!" })
        .expect(200);

      expect(res.body).toMatchObject({
        accessToken: expect.any(String),
        refreshToken: expect.any(String),
        expiresIn: expect.any(Number),
      });

      // Token is usable
      await request(app)
        .get("/api/me")
        .set("Authorization", `Bearer ${res.body.accessToken}`)
        .expect(200);
    });

    it("returns 401 for wrong password", async () => {
      await createTestUser({ email: "user@test.com" });

      await request(app)
        .post("/api/auth/login")
        .send({ email: "user@test.com", password: "wrong" })
        .expect(401);
    });

    it("rate limits login attempts", async () => {
      await createTestUser({ email: "target@test.com" });

      for (let i = 0; i < 5; i++) {
        await request(app)
          .post("/api/auth/login")
          .send({ email: "target@test.com", password: `wrong${i}` });
      }

      const res = await request(app)
        .post("/api/auth/login")
        .send({ email: "target@test.com", password: "anything" })
        .expect(429);

      expect(res.body.message).toContain("Too many login attempts");
    });
  });

  describe("Token refresh", () => {
    it("issues new access token with valid refresh token", async () => {
      const login = await request(app)
        .post("/api/auth/login")
        .send({ email: "user@test.com", password: "TestPass123!" })
        .expect(200);

      const res = await request(app)
        .post("/api/auth/refresh")
        .send({ refreshToken: login.body.refreshToken })
        .expect(200);

      expect(res.body.accessToken).not.toBe(login.body.accessToken);
    });

    it("rejects expired refresh token", async () => {
      const expiredToken = jwt.sign(
        { sub: "user-id", type: "refresh" },
        process.env.JWT_SECRET ?? "test-secret",
        { expiresIn: "0s" }
      );

      await request(app)
        .post("/api/auth/refresh")
        .send({ refreshToken: expiredToken })
        .expect(401);
    });
  });
});
```

## Performance Baseline Tests

```typescript
// tests/api/performance.test.ts

describe("Performance baselines", () => {
  beforeAll(async () => {
    // Seed realistic data volume
    await seedUsers(1000);
  });

  it("GET /api/users responds under 200ms with 1000 records", async () => {
    const start = Date.now();

    await request(app)
      .get("/api/users")
      .set("Authorization", `Bearer ${adminToken}`)
      .query({ limit: 50 })
      .expect(200);

    const duration = Date.now() - start;
    expect(duration).toBeLessThan(200);
  });

  it("POST /api/users responds under 100ms", async () => {
    const start = Date.now();

    await request(app)
      .post("/api/users")
      .set("Authorization", `Bearer ${adminToken}`)
      .send({
        email: `perf-${Date.now()}@test.com`,
        name: "Perf Test",
        role: "viewer",
        password: "TestPass123!",
      })
      .expect(201);

    const duration = Date.now() - start;
    expect(duration).toBeLessThan(100);
  });

  it("handles 50 concurrent requests without errors", async () => {
    const requests = Array.from({ length: 50 }, (_, i) =>
      request(app)
        .get("/api/users")
        .set("Authorization", `Bearer ${adminToken}`)
        .query({ page: (i % 5) + 1, limit: 10 })
    );

    const responses = await Promise.all(requests);
    const failures = responses.filter((r) => r.status >= 500);
    expect(failures).toHaveLength(0);
  });
});
```

## GraphQL Testing

```typescript
// tests/api/graphql.test.ts
import request from "supertest";

const QUERY = `
  query GetUsers($first: Int, $after: String) {
    users(first: $first, after: $after) {
      edges {
        node { id email name role }
        cursor
      }
      pageInfo {
        hasNextPage
        endCursor
      }
    }
  }
`;

describe("GraphQL /api/graphql", () => {
  it("returns paginated users", async () => {
    const res = await request(app)
      .post("/api/graphql")
      .set("Authorization", `Bearer ${token}`)
      .send({ query: QUERY, variables: { first: 10 } })
      .expect(200);

    expect(res.body.errors).toBeUndefined();
    expect(res.body.data.users.edges).toHaveLength(10);
    expect(res.body.data.users.pageInfo.hasNextPage).toBe(true);
  });

  it("handles cursor-based pagination", async () => {
    const page1 = await request(app)
      .post("/api/graphql")
      .set("Authorization", `Bearer ${token}`)
      .send({ query: QUERY, variables: { first: 5 } })
      .expect(200);

    const cursor = page1.body.data.users.pageInfo.endCursor;

    const page2 = await request(app)
      .post("/api/graphql")
      .set("Authorization", `Bearer ${token}`)
      .send({ query: QUERY, variables: { first: 5, after: cursor } })
      .expect(200);

    const page1Ids = page1.body.data.users.edges.map((e: any) => e.node.id);
    const page2Ids = page2.body.data.users.edges.map((e: any) => e.node.id);
    expect(page1Ids).not.toEqual(expect.arrayContaining(page2Ids));
  });

  it("returns errors for invalid queries", async () => {
    const res = await request(app)
      .post("/api/graphql")
      .set("Authorization", `Bearer ${token}`)
      .send({ query: "{ nonExistentField }" })
      .expect(200); // GraphQL returns 200 with errors

    expect(res.body.errors).toBeDefined();
    expect(res.body.errors[0].message).toContain("Cannot query field");
  });

  it("enforces query depth limit", async () => {
    const deepQuery = `{
      users(first: 1) {
        edges {
          node {
            posts {
              comments {
                author {
                  posts {
                    comments { id }
                  }
                }
              }
            }
          }
        }
      }
    }`;

    const res = await request(app)
      .post("/api/graphql")
      .set("Authorization", `Bearer ${token}`)
      .send({ query: deepQuery })
      .expect(200);

    expect(res.body.errors[0].message).toContain("depth");
  });
});
```

## Gotchas

1. **Test isolation failures** — Tests that share database state without cleanup cause flaky failures. Always truncate tables in `beforeEach`, not `beforeAll`. Use transactions with rollback for faster cleanup: wrap each test in a transaction and roll back after.

2. **Supertest doesn't follow redirects by default** — `request(app).get("/old-path")` returns the 301/302, not the final destination. Use `.redirects(1)` to follow, or assert the redirect itself with `.expect(301)` and check `res.headers.location`.

3. **Time-dependent tests are flaky** — Tests checking `createdAt`, token expiry, or rate limit windows fail when CI is slow. Use `jest.useFakeTimers()` or freeze time with libraries like `timekeeper`. Never assert exact timestamps — use `expect.any(String)` or check within a range.

4. **Port conflicts in parallel test suites** — If tests start the actual HTTP server, parallel Jest workers fight over the port. Use `supertest(app)` with the Express app directly (no `.listen()`), or assign dynamic ports with `server.listen(0)`.

5. **Forgetting to close database connections** — Missing `afterAll(() => db.destroy())` leaves connections open. Jest hangs with `--detectOpenHandles` warning. Every test file that opens a DB connection must close it.

6. **Testing against the wrong environment** — Tests accidentally hitting production APIs because `NODE_ENV` or `DATABASE_URL` wasn't set. Use a dedicated `jest.setup.ts` that enforces `NODE_ENV=test` and validates the database URL points to a test database before any queries run.
