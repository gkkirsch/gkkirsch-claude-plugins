---
name: vitest-patterns
description: >
  Vitest testing patterns — unit tests, mocking, spies, test fixtures, parameterized tests,
  snapshot testing, async testing, and configuration.
  Triggers: "vitest", "unit test", "test mock", "test spy", "vi.fn",
  "describe test", "test setup", "vitest config".
  NOT for: E2E tests (use playwright-e2e), strategy (use test-architect agent).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Vitest Patterns

## Configuration

```typescript
// vitest.config.ts
import { defineConfig } from 'vitest/config';
import path from 'path';

export default defineConfig({
  test: {
    globals: true,                    // No need to import describe, it, expect
    environment: 'node',              // or 'jsdom' for React components
    include: ['src/**/*.test.ts', 'src/**/*.test.tsx'],
    exclude: ['tests/e2e/**'],
    coverage: {
      provider: 'v8',
      reporter: ['text', 'html', 'lcov'],
      include: ['src/**/*.ts'],
      exclude: ['src/**/*.test.ts', 'src/**/index.ts', 'src/types/**'],
      thresholds: {
        branches: 80,
        functions: 80,
        lines: 80,
        statements: 80,
      },
    },
    setupFiles: ['./tests/setup.ts'],
  },
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
});
```

## Basic Test Structure

```typescript
import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import { PostService } from '../posts.service';

describe('PostService', () => {
  let service: PostService;

  beforeEach(() => {
    service = new PostService();
  });

  describe('create', () => {
    it('should create a post with valid data', async () => {
      const post = await service.create({
        title: 'Test Post',
        content: 'Content here',
        authorId: 'user-1',
      });

      expect(post).toMatchObject({
        title: 'Test Post',
        content: 'Content here',
      });
      expect(post.id).toBeDefined();
      expect(post.createdAt).toBeInstanceOf(Date);
    });

    it('should generate slug from title', async () => {
      const post = await service.create({
        title: 'My Amazing Post!',
        content: 'Content',
        authorId: 'user-1',
      });

      expect(post.slug).toBe('my-amazing-post');
    });

    it('should throw on empty title', async () => {
      await expect(
        service.create({ title: '', content: 'Content', authorId: 'user-1' }),
      ).rejects.toThrow('Title is required');
    });
  });
});
```

## Mocking

```typescript
import { describe, it, expect, vi, beforeEach } from 'vitest';

// Mock an entire module
vi.mock('../lib/prisma', () => ({
  prisma: {
    post: {
      findMany: vi.fn(),
      findUnique: vi.fn(),
      create: vi.fn(),
      update: vi.fn(),
      delete: vi.fn(),
      count: vi.fn(),
    },
    $transaction: vi.fn((fns) => Promise.all(fns)),
  },
}));

import { prisma } from '../lib/prisma';
import { PostService } from '../posts.service';

describe('PostService', () => {
  const service = new PostService();

  beforeEach(() => {
    vi.clearAllMocks(); // Reset call counts and implementations
  });

  it('should list posts with pagination', async () => {
    const mockPosts = [{ id: '1', title: 'Post 1' }];
    vi.mocked(prisma.post.findMany).mockResolvedValue(mockPosts);
    vi.mocked(prisma.post.count).mockResolvedValue(1);
    vi.mocked(prisma.$transaction).mockResolvedValue([mockPosts, 1]);

    const result = await service.list({ page: 1, limit: 20 });

    expect(result.items).toEqual(mockPosts);
    expect(result.total).toBe(1);
    expect(prisma.$transaction).toHaveBeenCalledOnce();
  });
});
```

### Mock Types

```typescript
// vi.fn() — function spy
const mockFn = vi.fn();
const mockFnWithReturn = vi.fn().mockReturnValue(42);
const mockFnAsync = vi.fn().mockResolvedValue({ id: '1' });
const mockFnReject = vi.fn().mockRejectedValue(new Error('fail'));

// Mock implementation
const mockFn = vi.fn((x: number) => x * 2);

// One-time mock
mockFn.mockReturnValueOnce(1).mockReturnValueOnce(2);

// vi.spyOn — spy on existing method
const spy = vi.spyOn(service, 'findById');
spy.mockResolvedValue({ id: '1', title: 'Test' });
// spy preserves original type, unlike vi.fn()

// vi.stubGlobal — mock global objects
vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
  ok: true,
  json: () => Promise.resolve({ data: 'test' }),
}));

// Restore originals
vi.restoreAllMocks();  // Restore spies to original implementation
vi.clearAllMocks();     // Clear call history but keep mock implementations
vi.resetAllMocks();     // Reset to vi.fn() with no implementation
```

### Partial Mocking (Keep Real + Mock Some)

```typescript
// Mock only specific exports, keep the rest real
vi.mock('../utils', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../utils')>();
  return {
    ...actual,
    sendEmail: vi.fn(), // Only mock sendEmail, keep everything else
  };
});
```

## Testing Async Code

```typescript
// Resolved promises
it('should resolve with data', async () => {
  const result = await fetchUser('1');
  expect(result).toEqual({ id: '1', name: 'Test' });
});

// Rejected promises
it('should throw on not found', async () => {
  await expect(fetchUser('bad-id')).rejects.toThrow('Not found');
  await expect(fetchUser('bad-id')).rejects.toThrow(NotFoundError);
  await expect(fetchUser('bad-id')).rejects.toMatchObject({
    statusCode: 404,
    code: 'NOT_FOUND',
  });
});

// Timers
import { vi, describe, it, expect, beforeEach, afterEach } from 'vitest';

describe('debounce', () => {
  beforeEach(() => { vi.useFakeTimers(); });
  afterEach(() => { vi.useRealTimers(); });

  it('should debounce calls', () => {
    const fn = vi.fn();
    const debounced = debounce(fn, 300);

    debounced();
    debounced();
    debounced();

    expect(fn).not.toHaveBeenCalled();

    vi.advanceTimersByTime(300);

    expect(fn).toHaveBeenCalledOnce();
  });
});
```

## Parameterized Tests

```typescript
// test.each with array of arrays
it.each([
  ['hello-world', 'Hello World'],
  ['my-post', 'My Post'],
  ['a', 'A'],
  ['already-slug', 'already-slug'],
])('slugify("%s") should return "%s"', (expected, input) => {
  expect(slugify(input)).toBe(expected);
});

// test.each with array of objects
it.each([
  { input: 0, expected: 'free' },
  { input: 9.99, expected: '$9.99' },
  { input: 100, expected: '$100.00' },
])('formatPrice($input) should return "$expected"', ({ input, expected }) => {
  expect(formatPrice(input)).toBe(expected);
});

// describe.each for grouped parameterized tests
describe.each([
  { role: 'admin', canDelete: true },
  { role: 'user', canDelete: false },
  { role: 'guest', canDelete: false },
])('$role permissions', ({ role, canDelete }) => {
  it(`should ${canDelete ? '' : 'not '}allow delete`, () => {
    expect(checkPermission(role, 'delete')).toBe(canDelete);
  });
});
```

## React Component Testing

```typescript
// vitest.config.ts for React
export default defineConfig({
  test: {
    environment: 'jsdom',
    setupFiles: ['./tests/setup.ts'],
  },
});

// tests/setup.ts
import '@testing-library/jest-dom/vitest';
import { cleanup } from '@testing-library/react';
import { afterEach } from 'vitest';

afterEach(() => { cleanup(); });
```

```tsx
import { describe, it, expect, vi } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { PostForm } from '../PostForm';

describe('PostForm', () => {
  it('should submit form data', async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn();

    render(<PostForm onSubmit={onSubmit} />);

    await user.type(screen.getByLabelText('Title'), 'My Post');
    await user.type(screen.getByLabelText('Content'), 'Post content');
    await user.click(screen.getByRole('button', { name: /submit/i }));

    await waitFor(() => {
      expect(onSubmit).toHaveBeenCalledWith({
        title: 'My Post',
        content: 'Post content',
      });
    });
  });

  it('should show validation errors', async () => {
    const user = userEvent.setup();
    render(<PostForm onSubmit={vi.fn()} />);

    await user.click(screen.getByRole('button', { name: /submit/i }));

    expect(screen.getByText('Title is required')).toBeInTheDocument();
  });

  it('should disable submit while loading', () => {
    render(<PostForm onSubmit={vi.fn()} isLoading />);

    expect(screen.getByRole('button', { name: /submit/i })).toBeDisabled();
  });
});
```

## Custom Hook Testing

```tsx
import { describe, it, expect, vi } from 'vitest';
import { renderHook, act, waitFor } from '@testing-library/react';
import { useDebounce } from '../useDebounce';

describe('useDebounce', () => {
  beforeEach(() => { vi.useFakeTimers(); });
  afterEach(() => { vi.useRealTimers(); });

  it('should debounce value updates', () => {
    const { result, rerender } = renderHook(
      ({ value }) => useDebounce(value, 300),
      { initialProps: { value: 'initial' } },
    );

    expect(result.current).toBe('initial');

    rerender({ value: 'updated' });
    expect(result.current).toBe('initial'); // Not yet

    act(() => { vi.advanceTimersByTime(300); });
    expect(result.current).toBe('updated'); // Now
  });
});
```

## API Integration Tests

```typescript
import { describe, it, expect, beforeAll, afterAll } from 'vitest';
import request from 'supertest';
import app from '../app';
import { prisma } from '../lib/prisma';

describe('POST /api/auth/signup', () => {
  afterAll(async () => {
    await prisma.user.deleteMany({ where: { email: { contains: '@test.com' } } });
    await prisma.$disconnect();
  });

  it('should create a new user', async () => {
    const res = await request(app)
      .post('/api/auth/signup')
      .send({
        email: 'new@test.com',
        password: 'password123',
        name: 'Test User',
      });

    expect(res.status).toBe(201);
    expect(res.body.success).toBe(true);
    expect(res.body.data.user.email).toBe('new@test.com');
    expect(res.body.data.accessToken).toBeDefined();
    expect(res.body.data.user.password).toBeUndefined(); // Not leaked
  });

  it('should reject duplicate email', async () => {
    // First signup
    await request(app)
      .post('/api/auth/signup')
      .send({ email: 'dupe@test.com', password: 'password123', name: 'Test' });

    // Duplicate
    const res = await request(app)
      .post('/api/auth/signup')
      .send({ email: 'dupe@test.com', password: 'password123', name: 'Test' });

    expect(res.status).toBe(409);
    expect(res.body.error.code).toBe('CONFLICT');
  });

  it('should reject invalid email', async () => {
    const res = await request(app)
      .post('/api/auth/signup')
      .send({ email: 'not-an-email', password: 'password123', name: 'Test' });

    expect(res.status).toBe(400);
    expect(res.body.error.code).toBe('VALIDATION_ERROR');
  });
});
```

## Test Fixtures / Factories

```typescript
// tests/fixtures/users.ts
import { faker } from '@faker-js/faker';

export function createTestUser(overrides?: Partial<User>) {
  return {
    id: faker.string.uuid(),
    email: faker.internet.email(),
    name: faker.person.fullName(),
    role: 'user' as const,
    createdAt: new Date(),
    updatedAt: new Date(),
    ...overrides,
  };
}

export function createTestPost(overrides?: Partial<Post>) {
  return {
    id: faker.string.uuid(),
    title: faker.lorem.sentence(),
    slug: faker.lorem.slug(),
    content: faker.lorem.paragraphs(3),
    published: true,
    authorId: faker.string.uuid(),
    createdAt: new Date(),
    updatedAt: new Date(),
    ...overrides,
  };
}
```

## Gotchas

1. **`vi.clearAllMocks()` vs `vi.resetAllMocks()` vs `vi.restoreAllMocks()`.** `clear` resets call counts. `reset` also removes implementations. `restore` restores spies to original. Use `clearAllMocks` in `beforeEach` for most cases.

2. **Module mocks must be at top level.** `vi.mock()` is hoisted to the top of the file automatically. Don't put it inside `describe` or `it` blocks — it won't work as expected.

3. **`toEqual` vs `toBe`.** `toBe` uses `===` (reference equality). `toEqual` does deep comparison. Use `toEqual` for objects and arrays.

4. **Fake timers affect all timers.** `vi.useFakeTimers()` affects `setTimeout`, `setInterval`, `Date.now()`, etc. This can break async operations that rely on real time. Always call `vi.useRealTimers()` in `afterEach`.

5. **`cleanup()` is not automatic in Vitest.** Unlike Jest with Testing Library, Vitest doesn't auto-cleanup rendered components. Add `cleanup()` in `afterEach` or in setup file.

6. **`mockResolvedValue` vs `mockReturnValue`.** Use `mockResolvedValue` for async functions (returns a Promise). `mockReturnValue` for sync functions.
