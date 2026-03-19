# Backend Testing Guide Reference

Comprehensive guide to testing Node.js backend applications with Vitest, Supertest, Testcontainers, and integration testing patterns.

---

## Test Pyramid for Backend APIs

```
            ┌───────────┐
           │    E2E     │  Few: Full API flows, real database
          │   Tests     │  Slow but high confidence
         ├─────────────┤
        │  Integration  │  Many: API endpoints, database queries
       │    Tests       │  Medium speed, good confidence
      ├─────────────────┤
     │    Unit Tests     │  Most: Services, utilities, validators
    │                     │  Fast, isolated, focused
   └───────────────────────┘
```

---

## Vitest Setup for Backend

### Configuration

```typescript
// vitest.config.ts
import { defineConfig } from 'vitest/config';

export default defineConfig({
  test: {
    // Global test config
    globals: true,
    root: '.',

    // Environment
    env: {
      NODE_ENV: 'test',
      DATABASE_URL: 'postgresql://localhost:5432/myapp_test',
      JWT_SECRET: 'test-secret-key-at-least-32-chars-long',
      LOG_LEVEL: 'error',  // Quiet during tests
    },

    // Coverage
    coverage: {
      provider: 'v8',
      reporter: ['text', 'lcov', 'html'],
      include: ['src/**/*.ts'],
      exclude: [
        'src/**/*.d.ts',
        'src/**/*.test.ts',
        'src/types/**',
        'src/generated/**',
      ],
      thresholds: {
        lines: 80,
        functions: 80,
        branches: 75,
        statements: 80,
      },
    },

    // Test organization
    include: ['**/*.test.ts', '**/*.spec.ts'],
    exclude: ['node_modules', 'dist', 'e2e'],

    // Timeouts
    testTimeout: 10_000,
    hookTimeout: 30_000,

    // Parallel execution
    pool: 'forks',
    poolOptions: {
      forks: {
        singleFork: false,  // Each test file in its own process
      },
    },

    // Setup files
    setupFiles: ['./tests/setup.ts'],
    globalSetup: ['./tests/global-setup.ts'],
  },
});
```

### Global Setup (Database)

```typescript
// tests/global-setup.ts
import { execSync } from 'node:child_process';

export async function setup() {
  // Push schema to test database
  execSync('npx prisma db push --force-reset --skip-generate', {
    env: {
      ...process.env,
      DATABASE_URL: 'postgresql://localhost:5432/myapp_test',
    },
    stdio: 'pipe',
  });
}

export async function teardown() {
  // Optional: drop test database
}
```

### Per-Test Setup

```typescript
// tests/setup.ts
import { beforeAll, afterAll, afterEach } from 'vitest';
import { PrismaClient } from '@prisma/client';

export const testDb = new PrismaClient({
  datasources: {
    db: { url: process.env.DATABASE_URL },
  },
});

beforeAll(async () => {
  await testDb.$connect();
});

afterEach(async () => {
  // Clean all tables between tests
  const tables = await testDb.$queryRaw<{ tablename: string }[]>`
    SELECT tablename FROM pg_tables
    WHERE schemaname = 'public'
    AND tablename != '_prisma_migrations'
  `;

  for (const { tablename } of tables) {
    await testDb.$executeRawUnsafe(
      `TRUNCATE TABLE "${tablename}" RESTART IDENTITY CASCADE`
    );
  }
});

afterAll(async () => {
  await testDb.$disconnect();
});
```

---

## Unit Testing

### Service Layer Testing

```typescript
// src/services/user.service.test.ts
import { describe, it, expect, beforeEach, vi } from 'vitest';
import { UserService } from './user.service.js';
import { NotFoundError, ConflictError } from '../errors/index.js';

// Mock the database
vi.mock('../lib/database.js', () => ({
  db: {
    user: {
      findUnique: vi.fn(),
      findMany: vi.fn(),
      create: vi.fn(),
      update: vi.fn(),
      delete: vi.fn(),
      count: vi.fn(),
    },
  },
}));

import { db } from '../lib/database.js';

describe('UserService', () => {
  let service: UserService;

  beforeEach(() => {
    service = new UserService();
    vi.clearAllMocks();
  });

  describe('findById', () => {
    it('returns user when found', async () => {
      const mockUser = { id: '1', email: 'test@example.com', name: 'Test' };
      vi.mocked(db.user.findUnique).mockResolvedValue(mockUser as any);

      const result = await service.findById('1');

      expect(result).toEqual(mockUser);
      expect(db.user.findUnique).toHaveBeenCalledWith({
        where: { id: '1' },
        select: expect.any(Object),
      });
    });

    it('throws NotFoundError when user not found', async () => {
      vi.mocked(db.user.findUnique).mockResolvedValue(null);

      await expect(service.findById('999')).rejects.toThrow(NotFoundError);
    });
  });

  describe('create', () => {
    it('creates user with hashed password', async () => {
      const input = { email: 'new@example.com', name: 'New', password: 'password123' };
      const mockUser = { id: '2', email: 'new@example.com', name: 'New' };

      vi.mocked(db.user.findUnique).mockResolvedValue(null); // No existing user
      vi.mocked(db.user.create).mockResolvedValue(mockUser as any);

      const result = await service.create(input);

      expect(result).toEqual(mockUser);
      expect(db.user.create).toHaveBeenCalledWith({
        data: expect.objectContaining({
          email: 'new@example.com',
          name: 'New',
          passwordHash: expect.any(String),
        }),
        select: expect.any(Object),
      });
    });

    it('throws ConflictError when email exists', async () => {
      const existing = { id: '1', email: 'exists@example.com' };
      vi.mocked(db.user.findUnique).mockResolvedValue(existing as any);

      await expect(
        service.create({ email: 'exists@example.com', name: 'Dup', password: 'pass123' })
      ).rejects.toThrow(ConflictError);
    });
  });

  describe('list', () => {
    it('returns paginated results', async () => {
      const mockUsers = [
        { id: '1', email: 'a@example.com', name: 'A' },
        { id: '2', email: 'b@example.com', name: 'B' },
      ];

      vi.mocked(db.user.findMany).mockResolvedValue(mockUsers as any);
      vi.mocked(db.user.count).mockResolvedValue(50);

      const result = await service.list({ page: 1, limit: 20 });

      expect(result.items).toHaveLength(2);
      expect(result.total).toBe(50);
      expect(result.totalPages).toBe(3);
    });
  });
});
```

### Utility Function Testing

```typescript
// src/utils/pagination.test.ts
import { describe, it, expect } from 'vitest';
import { buildPaginationMeta, parsePaginationQuery } from './pagination.js';

describe('buildPaginationMeta', () => {
  it('calculates correct total pages', () => {
    expect(buildPaginationMeta(100, 1, 20)).toEqual({
      total: 100,
      page: 1,
      limit: 20,
      totalPages: 5,
    });
  });

  it('handles zero items', () => {
    expect(buildPaginationMeta(0, 1, 20)).toEqual({
      total: 0,
      page: 1,
      limit: 20,
      totalPages: 0,
    });
  });

  it('rounds up total pages', () => {
    const meta = buildPaginationMeta(21, 1, 20);
    expect(meta.totalPages).toBe(2);
  });
});

describe('parsePaginationQuery', () => {
  it('returns defaults for empty query', () => {
    expect(parsePaginationQuery({})).toEqual({
      page: 1,
      limit: 20,
      sort: 'createdAt',
      order: 'desc',
    });
  });

  it('clamps limit to max 100', () => {
    expect(parsePaginationQuery({ limit: '500' }).limit).toBe(100);
  });

  it('rejects negative page', () => {
    expect(() => parsePaginationQuery({ page: '-1' })).toThrow();
  });
});
```

### Validation Schema Testing

```typescript
// src/schemas/user.schemas.test.ts
import { describe, it, expect } from 'vitest';
import { createUserSchema, updateUserSchema } from './user.schemas.js';

describe('createUserSchema', () => {
  it('accepts valid input', () => {
    const result = createUserSchema.safeParse({
      email: 'test@example.com',
      password: 'StrongPass1',
      name: 'Test User',
    });

    expect(result.success).toBe(true);
  });

  it('requires minimum password length', () => {
    const result = createUserSchema.safeParse({
      email: 'test@example.com',
      password: 'short',
      name: 'Test',
    });

    expect(result.success).toBe(false);
  });

  it('rejects invalid email', () => {
    const result = createUserSchema.safeParse({
      email: 'not-an-email',
      password: 'StrongPass1',
      name: 'Test',
    });

    expect(result.success).toBe(false);
  });

  it('trims and lowercases email', () => {
    const result = createUserSchema.parse({
      email: '  Test@EXAMPLE.com  ',
      password: 'StrongPass1',
      name: 'Test',
    });

    expect(result.email).toBe('test@example.com');
  });

  it('rejects unknown fields (mass assignment)', () => {
    const result = createUserSchema.safeParse({
      email: 'test@example.com',
      password: 'StrongPass1',
      name: 'Test',
      role: 'admin',      // Should be stripped or rejected
      isAdmin: true,       // Should be stripped or rejected
    });

    // Depending on Zod config: strict() rejects, strip() removes
    if (result.success) {
      expect(result.data).not.toHaveProperty('role');
      expect(result.data).not.toHaveProperty('isAdmin');
    }
  });
});
```

---

## Integration Testing

### API Endpoint Testing with Supertest

```typescript
// tests/integration/users.test.ts
import { describe, it, expect, beforeAll, afterAll, beforeEach } from 'vitest';
import request from 'supertest';
import { createApp } from '../../src/app.js';
import { testDb } from '../setup.js';
import { createTestUser, createAuthToken } from '../helpers/auth.js';

const app = createApp();

describe('Users API', () => {
  let adminToken: string;
  let userToken: string;
  let testUserId: string;

  beforeEach(async () => {
    // Create test users
    const admin = await createTestUser(testDb, { role: 'ADMIN' });
    const user = await createTestUser(testDb, { role: 'USER' });

    adminToken = createAuthToken(admin);
    userToken = createAuthToken(user);
    testUserId = user.id;
  });

  describe('GET /api/users', () => {
    it('returns paginated user list for admin', async () => {
      // Create additional test users
      await Promise.all(
        Array.from({ length: 5 }, (_, i) =>
          createTestUser(testDb, { email: `user${i}@test.com` })
        )
      );

      const res = await request(app)
        .get('/api/users')
        .set('Authorization', `Bearer ${adminToken}`)
        .query({ page: 1, limit: 3 })
        .expect(200);

      expect(res.body.data).toHaveLength(3);
      expect(res.body.meta).toEqual({
        total: expect.any(Number),
        page: 1,
        limit: 3,
        totalPages: expect.any(Number),
      });
      expect(res.body.meta.total).toBeGreaterThanOrEqual(7); // 2 + 5
    });

    it('filters by search query', async () => {
      await createTestUser(testDb, { name: 'Alice Smith', email: 'alice@test.com' });
      await createTestUser(testDb, { name: 'Bob Jones', email: 'bob@test.com' });

      const res = await request(app)
        .get('/api/users')
        .set('Authorization', `Bearer ${adminToken}`)
        .query({ search: 'alice' })
        .expect(200);

      expect(res.body.data).toHaveLength(1);
      expect(res.body.data[0].name).toBe('Alice Smith');
    });

    it('returns 401 without auth token', async () => {
      const res = await request(app)
        .get('/api/users')
        .expect(401);

      expect(res.body.error.code).toBe('UNAUTHORIZED');
    });

    it('returns 403 for non-admin users', async () => {
      await request(app)
        .get('/api/users')
        .set('Authorization', `Bearer ${userToken}`)
        .expect(403);
    });
  });

  describe('GET /api/users/:id', () => {
    it('returns user by ID', async () => {
      const res = await request(app)
        .get(`/api/users/${testUserId}`)
        .set('Authorization', `Bearer ${adminToken}`)
        .expect(200);

      expect(res.body.data.id).toBe(testUserId);
      expect(res.body.data).not.toHaveProperty('passwordHash');
    });

    it('returns 404 for non-existent user', async () => {
      const fakeId = '00000000-0000-0000-0000-000000000000';

      const res = await request(app)
        .get(`/api/users/${fakeId}`)
        .set('Authorization', `Bearer ${adminToken}`)
        .expect(404);

      expect(res.body.error.code).toBe('NOT_FOUND');
    });

    it('returns 422 for invalid UUID', async () => {
      await request(app)
        .get('/api/users/not-a-uuid')
        .set('Authorization', `Bearer ${adminToken}`)
        .expect(422);
    });
  });

  describe('POST /api/users', () => {
    it('creates user with valid data', async () => {
      const res = await request(app)
        .post('/api/users')
        .set('Authorization', `Bearer ${adminToken}`)
        .send({
          email: 'new@example.com',
          name: 'New User',
          password: 'StrongPass1',
        })
        .expect(201);

      expect(res.body.data).toMatchObject({
        email: 'new@example.com',
        name: 'New User',
      });
      expect(res.body.data.id).toBeDefined();
      expect(res.body.data).not.toHaveProperty('passwordHash');
      expect(res.body.data).not.toHaveProperty('password');
    });

    it('returns 409 for duplicate email', async () => {
      const existing = await createTestUser(testDb, { email: 'dup@test.com' });

      const res = await request(app)
        .post('/api/users')
        .set('Authorization', `Bearer ${adminToken}`)
        .send({
          email: 'dup@test.com',
          name: 'Duplicate',
          password: 'StrongPass1',
        })
        .expect(409);

      expect(res.body.error.code).toBe('CONFLICT');
    });

    it('validates required fields', async () => {
      const res = await request(app)
        .post('/api/users')
        .set('Authorization', `Bearer ${adminToken}`)
        .send({})
        .expect(422);

      expect(res.body.error.code).toBe('VALIDATION_ERROR');
    });

    it('prevents mass assignment', async () => {
      const res = await request(app)
        .post('/api/users')
        .set('Authorization', `Bearer ${adminToken}`)
        .send({
          email: 'hacker@test.com',
          name: 'Hacker',
          password: 'StrongPass1',
          role: 'ADMIN',      // Should be ignored
          isActive: true,     // Should be ignored
        })
        .expect(201);

      // User should be created with default role, not admin
      const user = await testDb.user.findUnique({
        where: { id: res.body.data.id },
      });
      expect(user?.role).toBe('USER');
    });
  });

  describe('PATCH /api/users/:id', () => {
    it('updates user fields', async () => {
      const res = await request(app)
        .patch(`/api/users/${testUserId}`)
        .set('Authorization', `Bearer ${adminToken}`)
        .send({ name: 'Updated Name' })
        .expect(200);

      expect(res.body.data.name).toBe('Updated Name');
    });
  });

  describe('DELETE /api/users/:id', () => {
    it('deletes user (soft delete)', async () => {
      await request(app)
        .delete(`/api/users/${testUserId}`)
        .set('Authorization', `Bearer ${adminToken}`)
        .expect(204);

      // User should still exist in database with deletedAt set
      const user = await testDb.user.findUnique({
        where: { id: testUserId },
      });
      // Depending on soft delete implementation:
      // expect(user?.deletedAt).toBeDefined();
    });
  });
});
```

### Test Helpers

```typescript
// tests/helpers/auth.ts
import jwt from 'jsonwebtoken';
import { hash } from '@node-rs/argon2';
import { randomUUID } from 'node:crypto';
import { type PrismaClient } from '@prisma/client';

interface TestUserOptions {
  email?: string;
  name?: string;
  role?: 'USER' | 'EDITOR' | 'ADMIN';
  password?: string;
}

let userCounter = 0;

export async function createTestUser(
  db: PrismaClient,
  options: TestUserOptions = {}
) {
  const count = ++userCounter;

  const passwordHash = await hash(options.password ?? 'TestPassword1', {
    memoryCost: 1024,  // Low for speed in tests
    timeCost: 1,
    outputLen: 32,
    parallelism: 1,
  });

  return db.user.create({
    data: {
      email: options.email ?? `testuser${count}@example.com`,
      name: options.name ?? `Test User ${count}`,
      role: options.role ?? 'USER',
      passwordHash,
      emailVerified: true,
      isActive: true,
    },
  });
}

export function createAuthToken(user: { id: string; email: string; role: string }) {
  return jwt.sign(
    { sub: user.id, email: user.email, role: user.role },
    process.env.JWT_SECRET!,
    { expiresIn: '1h', issuer: 'my-api', audience: 'my-app' }
  );
}

// tests/helpers/fixtures.ts
export function buildPost(overrides: Partial<PostInput> = {}): PostInput {
  const count = ++postCounter;
  return {
    title: `Test Post ${count}`,
    slug: `test-post-${count}`,
    content: 'Test content for post',
    status: 'DRAFT',
    ...overrides,
  };
}
```

---

## Testcontainers

### PostgreSQL Container

```typescript
// tests/containers/postgres.ts
import { PostgreSqlContainer, type StartedPostgreSqlContainer } from '@testcontainers/postgresql';
import { PrismaClient } from '@prisma/client';
import { execSync } from 'node:child_process';

let container: StartedPostgreSqlContainer;
let db: PrismaClient;

export async function startPostgres() {
  container = await new PostgreSqlContainer('postgres:16-alpine')
    .withDatabase('test')
    .withUsername('test')
    .withPassword('test')
    .withExposedPorts(5432)
    .start();

  const databaseUrl = container.getConnectionUri();

  // Run migrations
  execSync('npx prisma db push --skip-generate', {
    env: { ...process.env, DATABASE_URL: databaseUrl },
    stdio: 'pipe',
  });

  db = new PrismaClient({
    datasources: { db: { url: databaseUrl } },
  });

  await db.$connect();

  return { db, databaseUrl, container };
}

export async function stopPostgres() {
  await db?.$disconnect();
  await container?.stop();
}
```

### Redis Container

```typescript
// tests/containers/redis.ts
import { GenericContainer, type StartedTestContainer } from 'testcontainers';
import { Redis } from 'ioredis';

let container: StartedTestContainer;
let redis: Redis;

export async function startRedis() {
  container = await new GenericContainer('redis:7-alpine')
    .withExposedPorts(6379)
    .start();

  const host = container.getHost();
  const port = container.getMappedPort(6379);

  redis = new Redis({ host, port });

  return { redis, host, port, container };
}

export async function stopRedis() {
  await redis?.quit();
  await container?.stop();
}
```

### Docker Compose for Multiple Services

```typescript
// tests/containers/compose.ts
import { DockerComposeEnvironment, Wait } from 'testcontainers';
import path from 'node:path';

export async function startServices() {
  const environment = await new DockerComposeEnvironment(
    path.resolve(__dirname, '../../'),
    'docker-compose.test.yml'
  )
    .withWaitStrategy('postgres-1', Wait.forHealthCheck())
    .withWaitStrategy('redis-1', Wait.forLogMessage('Ready to accept connections'))
    .up();

  const postgresHost = environment.getContainer('postgres-1').getHost();
  const postgresPort = environment.getContainer('postgres-1').getMappedPort(5432);
  const redisHost = environment.getContainer('redis-1').getHost();
  const redisPort = environment.getContainer('redis-1').getMappedPort(6379);

  return {
    environment,
    databaseUrl: `postgresql://test:test@${postgresHost}:${postgresPort}/test`,
    redisUrl: `redis://${redisHost}:${redisPort}`,
  };
}
```

---

## Testing Patterns

### Testing Error Handling

```typescript
describe('Error handling', () => {
  it('returns 500 with generic message for unexpected errors', async () => {
    // Mock a service to throw an unexpected error
    vi.spyOn(userService, 'findById').mockRejectedValue(
      new Error('Database connection lost')
    );

    const res = await request(app)
      .get('/api/users/1')
      .set('Authorization', `Bearer ${token}`)
      .expect(500);

    expect(res.body.error.code).toBe('INTERNAL_ERROR');
    expect(res.body.error.message).toBe('An unexpected error occurred');
    // Should NOT contain the real error message
    expect(JSON.stringify(res.body)).not.toContain('Database connection lost');
  });

  it('returns proper error for invalid JSON body', async () => {
    const res = await request(app)
      .post('/api/users')
      .set('Authorization', `Bearer ${token}`)
      .set('Content-Type', 'application/json')
      .send('{ invalid json }')
      .expect(400);

    expect(res.body.error.code).toBe('INVALID_JSON');
  });
});
```

### Testing Rate Limiting

```typescript
describe('Rate limiting', () => {
  it('allows requests within limit', async () => {
    for (let i = 0; i < 5; i++) {
      await request(app)
        .post('/api/auth/login')
        .send({ email: 'test@test.com', password: 'wrong' })
        .expect((res) => {
          expect(res.status).not.toBe(429);
        });
    }
  });

  it('blocks requests exceeding limit', async () => {
    // Exceed the rate limit
    for (let i = 0; i < 11; i++) {
      await request(app)
        .post('/api/auth/login')
        .send({ email: 'test@test.com', password: 'wrong' });
    }

    const res = await request(app)
      .post('/api/auth/login')
      .send({ email: 'test@test.com', password: 'wrong' })
      .expect(429);

    expect(res.body.error.code).toBe('TOO_MANY_REQUESTS');
    expect(res.headers['retry-after']).toBeDefined();
  });
});
```

### Testing Authentication

```typescript
describe('Authentication', () => {
  it('rejects expired tokens', async () => {
    const expiredToken = jwt.sign(
      { sub: 'user-1', email: 'test@test.com', role: 'USER' },
      process.env.JWT_SECRET!,
      { expiresIn: '-1h' } // Already expired
    );

    const res = await request(app)
      .get('/api/users')
      .set('Authorization', `Bearer ${expiredToken}`)
      .expect(401);

    expect(res.body.error.message).toContain('expired');
  });

  it('rejects malformed tokens', async () => {
    await request(app)
      .get('/api/users')
      .set('Authorization', 'Bearer malformed.token.here')
      .expect(401);
  });

  it('rejects missing Authorization header', async () => {
    await request(app)
      .get('/api/users')
      .expect(401);
  });

  it('rejects non-Bearer auth scheme', async () => {
    await request(app)
      .get('/api/users')
      .set('Authorization', 'Basic dXNlcjpwYXNz')
      .expect(401);
  });
});
```

### Testing Database Transactions

```typescript
describe('Order creation (transactional)', () => {
  it('creates order and decrements stock atomically', async () => {
    const product = await createTestProduct(testDb, { stock: 10, price: 2500 });
    const user = await createTestUser(testDb);

    const res = await request(app)
      .post('/api/orders')
      .set('Authorization', `Bearer ${createAuthToken(user)}`)
      .send({ productId: product.id, quantity: 3 })
      .expect(201);

    // Verify stock decreased
    const updatedProduct = await testDb.product.findUnique({
      where: { id: product.id },
    });
    expect(updatedProduct?.stock).toBe(7);

    // Verify order created
    expect(res.body.data.totalPrice).toBe(7500);
  });

  it('rolls back on insufficient stock', async () => {
    const product = await createTestProduct(testDb, { stock: 2 });
    const user = await createTestUser(testDb);

    await request(app)
      .post('/api/orders')
      .set('Authorization', `Bearer ${createAuthToken(user)}`)
      .send({ productId: product.id, quantity: 5 })
      .expect(400);

    // Stock should be unchanged
    const unchangedProduct = await testDb.product.findUnique({
      where: { id: product.id },
    });
    expect(unchangedProduct?.stock).toBe(2);
  });
});
```

---

## Mocking Patterns

### Mocking External Services

```typescript
import { vi, beforeEach } from 'vitest';

// Mock an HTTP client
vi.mock('../lib/http-client.js', () => ({
  httpClient: {
    get: vi.fn(),
    post: vi.fn(),
  },
}));

import { httpClient } from '../lib/http-client.js';

describe('PaymentService', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('processes payment via Stripe', async () => {
    vi.mocked(httpClient.post).mockResolvedValue({
      data: { id: 'pi_123', status: 'succeeded' },
    });

    const result = await paymentService.charge({
      amount: 2500,
      currency: 'usd',
      customerId: 'cus_123',
    });

    expect(result.status).toBe('succeeded');
    expect(httpClient.post).toHaveBeenCalledWith(
      expect.stringContaining('/payment_intents'),
      expect.objectContaining({ amount: 2500 })
    );
  });

  it('handles payment failure gracefully', async () => {
    vi.mocked(httpClient.post).mockRejectedValue(
      new Error('Card declined')
    );

    await expect(
      paymentService.charge({ amount: 2500, currency: 'usd', customerId: 'cus_123' })
    ).rejects.toThrow('Payment failed');
  });
});
```

### Time Mocking

```typescript
import { vi, beforeEach, afterEach } from 'vitest';

describe('Token expiry', () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('expires refresh token after 7 days', async () => {
    const { refreshToken } = await authService.login('test@test.com', 'password');

    // Advance time by 8 days
    vi.advanceTimersByTime(8 * 24 * 60 * 60 * 1000);

    await expect(authService.refresh(refreshToken)).rejects.toThrow('Token expired');
  });
});
```

---

## Performance Testing

### Load Testing with autocannon

```typescript
// tests/load/users.load.ts
import autocannon from 'autocannon';

async function runLoadTest() {
  const result = await autocannon({
    url: 'http://localhost:3000/api/users',
    connections: 100,      // Concurrent connections
    pipelining: 10,        // Requests per connection
    duration: 30,          // Seconds
    headers: {
      'Authorization': `Bearer ${token}`,
    },
  });

  console.log('Results:', {
    requests: result.requests.total,
    throughput: result.throughput.average,
    latency: {
      p50: result.latency.p50,
      p99: result.latency.p99,
      max: result.latency.max,
    },
    errors: result.errors,
    timeouts: result.timeouts,
  });

  // Assertions
  expect(result.latency.p99).toBeLessThan(200); // p99 < 200ms
  expect(result.errors).toBe(0);
}
```

---

## Test Organization

### Recommended Structure

```
tests/
├── setup.ts                    # Per-test setup (cleanup, etc.)
├── global-setup.ts             # One-time setup (containers, migrations)
├── helpers/
│   ├── auth.ts                 # Auth helpers (createTestUser, tokens)
│   ├── fixtures.ts             # Data fixtures and builders
│   └── database.ts             # Database helpers
├── unit/
│   ├── services/
│   │   ├── user.service.test.ts
│   │   └── auth.service.test.ts
│   ├── utils/
│   │   ├── pagination.test.ts
│   │   └── crypto.test.ts
│   └── schemas/
│       └── user.schemas.test.ts
├── integration/
│   ├── auth.test.ts
│   ├── users.test.ts
│   └── posts.test.ts
├── load/
│   └── api.load.ts
└── containers/
    ├── postgres.ts
    └── redis.ts
```

### Test Naming Convention

```typescript
// Pattern: "it [action] [condition] [expected result]"
it('returns 404 when user does not exist');
it('creates user with valid data');
it('rejects expired tokens');
it('returns paginated results with correct meta');
it('handles concurrent requests without data corruption');
```

---

## CI/CD Testing

### GitHub Actions Pipeline

```yaml
# .github/workflows/test.yml
name: Tests
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:16-alpine
        env:
          POSTGRES_USER: test
          POSTGRES_PASSWORD: test
          POSTGRES_DB: test
        ports: ['5432:5432']
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

      redis:
        image: redis:7-alpine
        ports: ['6379:6379']
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    env:
      DATABASE_URL: postgresql://test:test@localhost:5432/test
      REDIS_URL: redis://localhost:6379
      JWT_SECRET: test-secret-key-at-least-32-characters-long
      NODE_ENV: test

    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: 22
          cache: npm
      - run: npm ci
      - run: npx prisma db push
      - run: npm test -- --coverage
      - uses: codecov/codecov-action@v4
        with:
          file: ./coverage/lcov.info
```
