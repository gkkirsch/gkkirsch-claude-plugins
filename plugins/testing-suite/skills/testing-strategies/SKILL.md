---
name: testing-strategies
description: >
  Testing strategies and patterns — test data management, mocking boundaries,
  CI test configuration, database testing, and test organization patterns.
  Triggers: "test data", "test database", "test ci", "mock boundary",
  "test isolation", "test factory", "test fixture", "integration test setup".
  NOT for: writing specific Vitest tests (use vitest-patterns), Playwright tests (use playwright-e2e).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Testing Strategies

## Test Database Setup

```typescript
// tests/helpers/db.ts
import { PrismaClient } from '@prisma/client';
import { execSync } from 'child_process';

const TEST_DATABASE_URL = process.env.TEST_DATABASE_URL
  || 'postgresql://postgres:postgres@localhost:5432/myapp_test';

// Create a separate Prisma client for tests
export const testPrisma = new PrismaClient({
  datasources: { db: { url: TEST_DATABASE_URL } },
});

// Reset database between test suites
export async function resetDatabase() {
  const tables = await testPrisma.$queryRaw<{ tablename: string }[]>`
    SELECT tablename FROM pg_tables WHERE schemaname = 'public'
    AND tablename != '_prisma_migrations'
  `;

  for (const { tablename } of tables) {
    await testPrisma.$executeRawUnsafe(`TRUNCATE TABLE "${tablename}" CASCADE`);
  }
}

// Setup test database (run once before all tests)
export async function setupTestDatabase() {
  process.env.DATABASE_URL = TEST_DATABASE_URL;
  execSync('npx prisma migrate deploy', {
    env: { ...process.env, DATABASE_URL: TEST_DATABASE_URL },
  });
}
```

```typescript
// tests/setup.ts (Vitest globalSetup)
import { beforeAll, afterAll, beforeEach } from 'vitest';
import { setupTestDatabase, resetDatabase, testPrisma } from './helpers/db';

beforeAll(async () => {
  await setupTestDatabase();
});

beforeEach(async () => {
  await resetDatabase();
});

afterAll(async () => {
  await testPrisma.$disconnect();
});
```

## Test Data Factories

```typescript
// tests/factories/index.ts
import { faker } from '@faker-js/faker';
import { prisma } from '../helpers/db';
import bcrypt from 'bcrypt';

export async function createUser(overrides?: Partial<{
  email: string;
  password: string;
  name: string;
  role: string;
}>) {
  const data = {
    email: overrides?.email || faker.internet.email(),
    password: await bcrypt.hash(overrides?.password || 'password123', 4), // Low rounds for speed
    name: overrides?.name || faker.person.fullName(),
    role: overrides?.role || 'user',
  };

  return prisma.user.create({ data });
}

export async function createPost(
  authorId: string,
  overrides?: Partial<{
    title: string;
    content: string;
    published: boolean;
  }>,
) {
  const title = overrides?.title || faker.lorem.sentence();

  return prisma.post.create({
    data: {
      title,
      slug: title.toLowerCase().replace(/\W+/g, '-'),
      content: overrides?.content || faker.lorem.paragraphs(3),
      published: overrides?.published ?? true,
      authorId,
    },
  });
}

// Helper: create user with token (for API tests)
export async function createAuthenticatedUser(overrides?: Parameters<typeof createUser>[0]) {
  const user = await createUser(overrides);
  const token = jwt.sign(
    { userId: user.id, email: user.email, role: user.role },
    process.env.JWT_SECRET!,
    { expiresIn: '1h' },
  );
  return { user, token };
}
```

## API Testing with Supertest

```typescript
// tests/integration/posts.test.ts
import { describe, it, expect, beforeEach } from 'vitest';
import request from 'supertest';
import app from '../../src/app';
import { createAuthenticatedUser, createPost } from '../factories';
import { resetDatabase } from '../helpers/db';

describe('Posts API', () => {
  let authUser: { user: any; token: string };

  beforeEach(async () => {
    await resetDatabase();
    authUser = await createAuthenticatedUser();
  });

  describe('GET /api/posts', () => {
    it('should return paginated posts', async () => {
      // Create test data
      await Promise.all(
        Array.from({ length: 25 }, () => createPost(authUser.user.id)),
      );

      const res = await request(app)
        .get('/api/posts?page=1&limit=10')
        .set('Authorization', `Bearer ${authUser.token}`);

      expect(res.status).toBe(200);
      expect(res.body.success).toBe(true);
      expect(res.body.data).toHaveLength(10);
      expect(res.body.meta.total).toBe(25);
      expect(res.body.meta.totalPages).toBe(3);
    });

    it('should filter by search term', async () => {
      await createPost(authUser.user.id, { title: 'TypeScript Guide' });
      await createPost(authUser.user.id, { title: 'Python Tutorial' });

      const res = await request(app)
        .get('/api/posts?search=typescript')
        .set('Authorization', `Bearer ${authUser.token}`);

      expect(res.body.data).toHaveLength(1);
      expect(res.body.data[0].title).toBe('TypeScript Guide');
    });
  });

  describe('POST /api/posts', () => {
    it('should require authentication', async () => {
      const res = await request(app)
        .post('/api/posts')
        .send({ title: 'Test', content: 'Content' });

      expect(res.status).toBe(401);
    });

    it('should validate input', async () => {
      const res = await request(app)
        .post('/api/posts')
        .set('Authorization', `Bearer ${authUser.token}`)
        .send({ title: '', content: '' });

      expect(res.status).toBe(400);
      expect(res.body.error.code).toBe('VALIDATION_ERROR');
      expect(res.body.error.details.title).toBeDefined();
    });
  });

  describe('DELETE /api/posts/:id', () => {
    it('should prevent deleting other users posts', async () => {
      const otherUser = await createAuthenticatedUser();
      const post = await createPost(otherUser.user.id);

      const res = await request(app)
        .delete(`/api/posts/${post.id}`)
        .set('Authorization', `Bearer ${authUser.token}`);

      expect(res.status).toBe(403);
    });
  });
});
```

## Mocking Boundaries

```
What to mock:
├── External APIs (Stripe, SendGrid, S3)
├── Time/dates (use vi.useFakeTimers())
├── Random values (seed or mock)
├── File system (when testing logic, not IO)
└── Environment variables

What NOT to mock:
├── Your own database (use test database)
├── Your own services (test the real thing)
├── Express middleware (test through HTTP)
├── Validation schemas (test with real data)
└── Internal utility functions (test the real implementation)
```

```typescript
// Good: Mock external service
vi.mock('../lib/stripe', () => ({
  stripe: {
    customers: { create: vi.fn().mockResolvedValue({ id: 'cus_test' }) },
    subscriptions: { create: vi.fn().mockResolvedValue({ id: 'sub_test' }) },
  },
}));

// Good: Mock email sending
vi.mock('../lib/email', () => ({
  sendEmail: vi.fn().mockResolvedValue(undefined),
}));

// Bad: Don't mock your own service
// vi.mock('../services/posts.service'); // ← testing nothing
```

## CI Configuration

```yaml
# .github/workflows/test.yml
name: Tests
on: [push, pull_request]

jobs:
  unit-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: 20
          cache: 'npm'
      - run: npm ci
      - run: npx vitest --coverage --reporter=default

  integration-tests:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:16
        env:
          POSTGRES_USER: postgres
          POSTGRES_PASSWORD: postgres
          POSTGRES_DB: myapp_test
        ports:
          - 5432:5432
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    env:
      DATABASE_URL: postgresql://postgres:postgres@localhost:5432/myapp_test
      JWT_SECRET: test-secret-at-least-32-characters-long
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with: { node-version: 20, cache: 'npm' }
      - run: npm ci
      - run: npx prisma migrate deploy
      - run: npx vitest --config vitest.integration.config.ts

  e2e-tests:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:16
        env:
          POSTGRES_USER: postgres
          POSTGRES_PASSWORD: postgres
          POSTGRES_DB: myapp_test
        ports: ['5432:5432']
        options: --health-cmd pg_isready --health-interval 10s --health-timeout 5s --health-retries 5
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with: { node-version: 20, cache: 'npm' }
      - run: npm ci
      - run: npx playwright install --with-deps
      - run: npx prisma migrate deploy
        env:
          DATABASE_URL: postgresql://postgres:postgres@localhost:5432/myapp_test
      - run: npx playwright test
        env:
          DATABASE_URL: postgresql://postgres:postgres@localhost:5432/myapp_test
          JWT_SECRET: test-secret-at-least-32-characters-long
      - uses: actions/upload-artifact@v4
        if: failure()
        with:
          name: playwright-report
          path: playwright-report/
```

## Test Scripts

```json
{
  "scripts": {
    "test": "vitest",
    "test:run": "vitest run",
    "test:coverage": "vitest run --coverage",
    "test:integration": "vitest --config vitest.integration.config.ts",
    "test:e2e": "playwright test",
    "test:e2e:ui": "playwright test --ui",
    "test:e2e:headed": "playwright test --headed",
    "test:e2e:debug": "playwright test --debug"
  }
}
```

## Gotchas

1. **Test database must be separate.** NEVER run tests against your development database. Use a dedicated test database and reset between suites.

2. **bcrypt is slow in tests.** Use cost factor 4 (not 12) in test factories. It's still bcrypt, just faster. Or mock bcrypt entirely for unit tests.

3. **Parallel tests + shared database = flaky.** Either use `--workers=1` for integration tests, or isolate data per test (use unique identifiers, not shared rows).

4. **`faker` generates random data.** Tests using faker won't reproduce the same data on re-run. For deterministic tests, use `faker.seed(12345)` in your setup file, or use fixed test data for assertions.

5. **Supertest doesn't start the server.** `request(app)` makes in-process requests without listening on a port. Don't call `app.listen()` in your test setup.

6. **Environment variables in CI.** Set `NODE_ENV=test`, `JWT_SECRET`, and `DATABASE_URL` in CI environment. Don't rely on `.env` files — they're gitignored.
