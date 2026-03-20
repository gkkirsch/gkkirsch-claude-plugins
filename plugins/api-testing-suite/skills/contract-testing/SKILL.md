---
name: contract-testing
description: >
  Contract testing patterns — consumer-driven contracts with Pact,
  provider verification, schema validation, API versioning strategies,
  backward compatibility checks, and OpenAPI contract enforcement.
  Triggers: "contract test", "pact test", "consumer driven", "api contract",
  "schema validation", "api versioning", "backward compatible", "openapi test".
  NOT for: General API endpoint testing (use api-test-patterns).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Contract Testing

## Consumer-Driven Contracts with Pact (TypeScript)

```typescript
// consumer/tests/pact/user-service.pact.test.ts
import { PactV4, MatchersV3 } from "@pact-foundation/pact";
import { UserApiClient } from "../../src/clients/user-api";

const { like, eachLike, regex, integer, string, timestamp } = MatchersV3;

const provider = new PactV4({
  consumer: "OrderService",
  provider: "UserService",
  logLevel: "warn",
});

describe("UserService API Contract", () => {
  describe("GET /api/users/:id", () => {
    it("returns user details when user exists", async () => {
      await provider
        .addInteraction()
        .given("user abc-123 exists")
        .uponReceiving("a request for user abc-123")
        .withRequest("GET", "/api/users/abc-123", (builder) => {
          builder.headers({
            Authorization: like("Bearer token-123"),
            Accept: "application/json",
          });
        })
        .willRespondWith(200, (builder) => {
          builder
            .headers({ "Content-Type": "application/json" })
            .jsonBody({
              id: string("abc-123"),
              email: regex(/^.+@.+\..+$/, "user@example.com"),
              name: string("Jane Doe"),
              role: regex(/^(admin|editor|viewer)$/, "editor"),
              createdAt: timestamp(
                "yyyy-MM-dd'T'HH:mm:ss.SSS'Z'",
                "2024-01-15T10:30:00.000Z"
              ),
              subscription: {
                plan: regex(/^(free|pro|enterprise)$/, "pro"),
                expiresAt: timestamp(
                  "yyyy-MM-dd'T'HH:mm:ss.SSS'Z'",
                  "2025-01-15T10:30:00.000Z"
                ),
              },
            });
        })
        .executeTest(async (mockServer) => {
          const client = new UserApiClient(mockServer.url);
          const user = await client.getUser("abc-123");

          expect(user.id).toBe("abc-123");
          expect(user.email).toContain("@");
          expect(user.subscription.plan).toBeDefined();
        });
    });

    it("returns 404 when user not found", async () => {
      await provider
        .addInteraction()
        .given("user xyz-999 does not exist")
        .uponReceiving("a request for non-existent user xyz-999")
        .withRequest("GET", "/api/users/xyz-999", (builder) => {
          builder.headers({
            Authorization: like("Bearer token-123"),
          });
        })
        .willRespondWith(404, (builder) => {
          builder.jsonBody({
            error: string("NotFound"),
            message: string("User not found"),
          });
        })
        .executeTest(async (mockServer) => {
          const client = new UserApiClient(mockServer.url);
          await expect(client.getUser("xyz-999")).rejects.toThrow("User not found");
        });
    });
  });

  describe("POST /api/users", () => {
    it("creates a new user", async () => {
      await provider
        .addInteraction()
        .uponReceiving("a request to create a user")
        .withRequest("POST", "/api/users", (builder) => {
          builder
            .headers({
              "Content-Type": "application/json",
              Authorization: like("Bearer admin-token"),
            })
            .jsonBody({
              email: "new@example.com",
              name: "New User",
              role: "viewer",
            });
        })
        .willRespondWith(201, (builder) => {
          builder.jsonBody({
            id: like("new-user-id"),
            email: string("new@example.com"),
            name: string("New User"),
            role: string("viewer"),
            createdAt: timestamp(
              "yyyy-MM-dd'T'HH:mm:ss.SSS'Z'",
              "2024-06-01T12:00:00.000Z"
            ),
          });
        })
        .executeTest(async (mockServer) => {
          const client = new UserApiClient(mockServer.url);
          const user = await client.createUser({
            email: "new@example.com",
            name: "New User",
            role: "viewer",
          });

          expect(user.id).toBeDefined();
          expect(user.email).toBe("new@example.com");
        });
    });
  });

  describe("GET /api/users (list)", () => {
    it("returns paginated user list", async () => {
      await provider
        .addInteraction()
        .given("multiple users exist")
        .uponReceiving("a request for the user list")
        .withRequest("GET", "/api/users", (builder) => {
          builder
            .headers({ Authorization: like("Bearer token") })
            .query({ page: "1", limit: "10" });
        })
        .willRespondWith(200, (builder) => {
          builder.jsonBody({
            data: eachLike({
              id: string("user-1"),
              email: string("user@example.com"),
              name: string("User One"),
              role: string("viewer"),
            }),
            pagination: {
              page: integer(1),
              limit: integer(10),
              total: integer(25),
              totalPages: integer(3),
            },
          });
        })
        .executeTest(async (mockServer) => {
          const client = new UserApiClient(mockServer.url);
          const result = await client.listUsers({ page: 1, limit: 10 });

          expect(result.data.length).toBeGreaterThan(0);
          expect(result.pagination.page).toBe(1);
        });
    });
  });
});
```

## Provider Verification

```typescript
// provider/tests/pact/verify.test.ts
import { Verifier } from "@pact-foundation/pact";
import { app } from "../../src/app";
import { db } from "../../src/db";
import http from "http";

describe("Provider verification", () => {
  let server: http.Server;
  let port: number;

  beforeAll(async () => {
    await db.migrate.latest();
    server = app.listen(0);
    port = (server.address() as any).port;
  });

  afterAll(async () => {
    server.close();
    await db.destroy();
  });

  it("validates contracts against OrderService consumer", async () => {
    await new Verifier({
      providerBaseUrl: `http://localhost:${port}`,
      pactUrls: ["./pacts/OrderService-UserService.json"],
      // Or from Pact Broker:
      // pactBrokerUrl: "https://pact-broker.example.com",
      // provider: "UserService",
      // consumerVersionSelectors: [
      //   { mainBranch: true },
      //   { deployedOrReleased: true },
      // ],
      stateHandlers: {
        "user abc-123 exists": async () => {
          await db("users").insert({
            id: "abc-123",
            email: "user@example.com",
            name: "Jane Doe",
            role: "editor",
            subscription_plan: "pro",
            subscription_expires: "2025-01-15T10:30:00.000Z",
            created_at: "2024-01-15T10:30:00.000Z",
          });
        },
        "user xyz-999 does not exist": async () => {
          await db("users").where({ id: "xyz-999" }).delete();
        },
        "multiple users exist": async () => {
          await db("users").insert([
            { id: "u1", email: "a@test.com", name: "A", role: "viewer" },
            { id: "u2", email: "b@test.com", name: "B", role: "editor" },
          ]);
        },
      },
      beforeEach: async () => {
        await db("users").truncate();
      },
    }).verifyProvider();
  });
});
```

## OpenAPI Schema Validation

```typescript
// tests/middleware/schema-validation.test.ts
import SwaggerParser from "@apidevtools/swagger-parser";
import request from "supertest";
import { app } from "../../src/app";

let schema: any;

beforeAll(async () => {
  schema = await SwaggerParser.dereference("./openapi.yaml");
});

describe("OpenAPI contract compliance", () => {
  it("GET /api/users response matches schema", async () => {
    const res = await request(app)
      .get("/api/users")
      .set("Authorization", `Bearer ${token}`)
      .query({ page: 1, limit: 10 })
      .expect(200);

    const responseSchema =
      schema.paths["/api/users"].get.responses["200"].content[
        "application/json"
      ].schema;

    // Validate structure matches
    expect(res.body).toHaveProperty("data");
    expect(res.body).toHaveProperty("pagination");

    // Check required fields exist on each item
    const requiredFields =
      responseSchema.properties.data.items.required || [];
    res.body.data.forEach((item: any) => {
      requiredFields.forEach((field: string) => {
        expect(item).toHaveProperty(field);
      });
    });
  });
});

// Runtime middleware for schema validation
// src/middleware/validate-response.ts
import { OpenApiValidator } from "express-openapi-validator";

export function setupResponseValidation(app: Express) {
  if (process.env.NODE_ENV === "test") {
    app.use(
      OpenApiValidator.middleware({
        apiSpec: "./openapi.yaml",
        validateRequests: true,
        validateResponses: true, // Catches contract violations at runtime
      })
    );
  }
}
```

## Backward Compatibility Checking

```typescript
// tests/compatibility/breaking-changes.test.ts
import SwaggerParser from "@apidevtools/swagger-parser";

describe("API backward compatibility", () => {
  let currentSpec: any;
  let previousSpec: any;

  beforeAll(async () => {
    currentSpec = await SwaggerParser.dereference("./openapi.yaml");
    previousSpec = await SwaggerParser.dereference("./openapi.previous.yaml");
  });

  it("does not remove existing endpoints", () => {
    const previousPaths = Object.keys(previousSpec.paths);
    const currentPaths = Object.keys(currentSpec.paths);

    const removedPaths = previousPaths.filter(
      (p) => !currentPaths.includes(p)
    );
    expect(removedPaths).toEqual([]);
  });

  it("does not remove required response fields", () => {
    for (const [path, methods] of Object.entries(previousSpec.paths) as any) {
      for (const [method, operation] of Object.entries(methods) as any) {
        if (!currentSpec.paths[path]?.[method]) continue;

        const prevResponse =
          operation.responses?.["200"]?.content?.["application/json"]?.schema;
        const currResponse =
          currentSpec.paths[path][method].responses?.["200"]?.content?.[
            "application/json"
          ]?.schema;

        if (!prevResponse || !currResponse) continue;

        const prevRequired = prevResponse.required || [];
        const currRequired = currResponse.required || [];

        // Current must include all previously required fields
        prevRequired.forEach((field: string) => {
          expect(currRequired).toContain(field);
        });
      }
    }
  });

  it("does not change existing field types", () => {
    for (const [path, methods] of Object.entries(previousSpec.paths) as any) {
      for (const [method, operation] of Object.entries(methods) as any) {
        if (!currentSpec.paths[path]?.[method]) continue;

        const prevSchema =
          operation.responses?.["200"]?.content?.["application/json"]?.schema;
        const currSchema =
          currentSpec.paths[path][method].responses?.["200"]?.content?.[
            "application/json"
          ]?.schema;

        if (!prevSchema?.properties || !currSchema?.properties) continue;

        for (const [field, def] of Object.entries(prevSchema.properties) as any) {
          if (currSchema.properties[field]) {
            expect(currSchema.properties[field].type).toBe(def.type);
          }
        }
      }
    }
  });

  it("new required request fields have defaults", () => {
    for (const [path, methods] of Object.entries(currentSpec.paths) as any) {
      for (const [method, operation] of Object.entries(methods) as any) {
        const prevOp = previousSpec.paths[path]?.[method];
        if (!prevOp) continue; // new endpoint, skip

        const currBody =
          operation.requestBody?.content?.["application/json"]?.schema;
        const prevBody =
          prevOp.requestBody?.content?.["application/json"]?.schema;

        if (!currBody?.required || !prevBody) continue;

        const newRequired = (currBody.required || []).filter(
          (f: string) => !(prevBody.required || []).includes(f)
        );

        // New required fields must have a default value
        newRequired.forEach((field: string) => {
          expect(currBody.properties[field]).toHaveProperty("default");
        });
      }
    }
  });
});
```

## API Versioning Patterns

```typescript
// src/middleware/api-version.ts

// URL-based versioning
// /api/v1/users, /api/v2/users
import { Router } from "express";

const v1Router = Router();
const v2Router = Router();

v1Router.get("/users", (req, res) => {
  // V1: returns flat user object
  res.json({ id: "123", email: "user@test.com", name: "Test" });
});

v2Router.get("/users", (req, res) => {
  // V2: returns structured response with metadata
  res.json({
    data: { id: "123", email: "user@test.com", profile: { name: "Test" } },
    meta: { version: "2", deprecations: [] },
  });
});

app.use("/api/v1", v1Router);
app.use("/api/v2", v2Router);

// Header-based versioning
// Accept: application/vnd.myapi.v2+json
function versionMiddleware(req: Request, res: Response, next: NextFunction) {
  const accept = req.headers.accept || "";
  const match = accept.match(/application\/vnd\.myapi\.v(\d+)\+json/);
  req.apiVersion = match ? parseInt(match[1]) : 1;
  next();
}

// Test versioned endpoints
describe("API versioning", () => {
  it("v1 returns flat user object", async () => {
    const res = await request(app)
      .get("/api/v1/users/123")
      .set("Authorization", `Bearer ${token}`)
      .expect(200);

    expect(res.body).toHaveProperty("name");
    expect(res.body).not.toHaveProperty("profile");
  });

  it("v2 returns structured response", async () => {
    const res = await request(app)
      .get("/api/v2/users/123")
      .set("Authorization", `Bearer ${token}`)
      .expect(200);

    expect(res.body.data).toHaveProperty("profile.name");
    expect(res.body.meta.version).toBe("2");
  });

  it("v1 includes deprecation header", async () => {
    const res = await request(app)
      .get("/api/v1/users/123")
      .set("Authorization", `Bearer ${token}`)
      .expect(200);

    expect(res.headers["deprecation"]).toBeDefined();
    expect(res.headers["sunset"]).toBeDefined();
  });
});
```

## CI Pipeline Contract Verification

```yaml
# .github/workflows/contract-tests.yml
name: Contract Tests

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  consumer-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with: { node-version: 20 }

      - run: npm ci
      - run: npm run test:pact

      # Publish pact to broker
      - name: Publish pacts
        if: github.ref == 'refs/heads/main'
        run: |
          npx pact-broker publish ./pacts \
            --consumer-app-version=${{ github.sha }} \
            --branch=${{ github.ref_name }} \
            --broker-base-url=${{ secrets.PACT_BROKER_URL }} \
            --broker-token=${{ secrets.PACT_BROKER_TOKEN }}

  provider-verification:
    needs: consumer-tests
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:16
        env:
          POSTGRES_DB: test
          POSTGRES_PASSWORD: test
        ports: ['5432:5432']
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with: { node-version: 20 }

      - run: npm ci
      - run: npm run test:pact:verify
        env:
          DATABASE_URL: postgres://postgres:test@localhost:5432/test

      # Record verification result
      - name: Record verification
        if: github.ref == 'refs/heads/main'
        run: |
          npx pact-broker record-deployment \
            --pacticipant=UserService \
            --version=${{ github.sha }} \
            --environment=production \
            --broker-base-url=${{ secrets.PACT_BROKER_URL }} \
            --broker-token=${{ secrets.PACT_BROKER_TOKEN }}

  breaking-change-check:
    runs-on: ubuntu-latest
    if: github.event_name == 'pull_request'
    steps:
      - uses: actions/checkout@v4
        with: { fetch-depth: 0 }

      - name: Get previous OpenAPI spec
        run: git show origin/main:openapi.yaml > openapi.previous.yaml

      - run: npm ci
      - run: npm run test:compatibility
```

## Gotchas

1. **Pact matchers are not assertions** — `like("hello")` means "any string", not "must be hello". The exact value is only used for the mock response. Use matchers to define the shape/type, then assert business logic in your consumer test's `executeTest` callback.

2. **Provider state names must match exactly** — `"user exists"` in the consumer and `"User exists"` in the provider are different states. The provider won't find the handler and the test silently passes with no data setup. Use a shared constants file or copy-paste the exact string.

3. **Pact tests are not integration tests** — Pact verifies the contract (shape, status codes, headers), not business logic. A passing Pact test doesn't mean the feature works end-to-end. You still need integration tests that hit real databases and external services.

4. **OpenAPI spec drift** — The spec and implementation can diverge silently. Use `express-openapi-validator` with `validateResponses: true` in your test environment to catch drift automatically. Run schema validation as part of CI, not just in contract tests.

5. **Breaking changes hide in optional fields** — Adding a new field is safe. Making a previously optional field required, changing a field's type from `string` to `number`, or renaming a field are all breaking changes that automated checks might miss if you only compare required fields. Check all property types and names.

6. **Consumer tests pass with stale pacts** — If you change the provider API but don't regenerate consumer pacts, old pacts still pass on the consumer side (they test against the old mock). Always run `can-i-deploy` via Pact Broker before deploying to catch consumer/provider version mismatches.
