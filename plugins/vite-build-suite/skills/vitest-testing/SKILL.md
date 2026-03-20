---
name: vitest-testing
description: >
  Vitest testing framework — configuration, unit tests, component testing,
  mocking, coverage, snapshots, and integration with Vite projects.
  Triggers: "vitest", "vitest config", "vitest setup", "vitest mock",
  "vitest coverage", "vitest snapshot", "unit testing vite", "component testing vitest".
  NOT for: Playwright E2E tests (use playwright-e2e-suite), Jest (use testing-suite).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Vitest Testing Framework

## Setup

```bash
npm install -D vitest @vitest/coverage-v8
# For React component testing:
npm install -D @testing-library/react @testing-library/jest-dom @testing-library/user-event jsdom
# For Vue component testing:
npm install -D @testing-library/vue @vue/test-utils jsdom
```

## Configuration

```typescript
// vitest.config.ts (or inline in vite.config.ts)
import { defineConfig } from "vitest/config";
import react from "@vitejs/plugin-react-swc";

export default defineConfig({
  plugins: [react()],
  test: {
    // Environment
    environment: "jsdom", // 'jsdom', 'happy-dom', 'node'
    globals: true,        // Use describe/it/expect without imports

    // Setup files
    setupFiles: ["./tests/setup.ts"],

    // File patterns
    include: ["src/**/*.{test,spec}.{ts,tsx}"],
    exclude: ["node_modules", "dist", "e2e"],

    // Coverage
    coverage: {
      provider: "v8",
      reporter: ["text", "html", "lcov"],
      include: ["src/**/*.{ts,tsx}"],
      exclude: [
        "src/**/*.test.{ts,tsx}",
        "src/**/*.d.ts",
        "src/types/",
        "src/main.tsx",
      ],
      thresholds: {
        statements: 80,
        branches: 80,
        functions: 80,
        lines: 80,
      },
    },

    // Performance
    pool: "forks",        // 'forks' (default), 'threads', 'vmThreads'
    poolOptions: {
      forks: { singleFork: false },
    },

    // Timeouts
    testTimeout: 10000,
    hookTimeout: 10000,

    // Watch mode settings
    watchExclude: ["node_modules", "dist"],
  },
});
```

```typescript
// tests/setup.ts
import "@testing-library/jest-dom/vitest";
import { cleanup } from "@testing-library/react";
import { afterEach } from "vitest";

// Cleanup after each test
afterEach(() => {
  cleanup();
});

// Mock IntersectionObserver
class MockIntersectionObserver {
  observe = vi.fn();
  unobserve = vi.fn();
  disconnect = vi.fn();
}
Object.defineProperty(window, "IntersectionObserver", {
  value: MockIntersectionObserver,
});

// Mock matchMedia
Object.defineProperty(window, "matchMedia", {
  value: vi.fn().mockImplementation((query) => ({
    matches: false,
    media: query,
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
  })),
});
```

## Writing Tests

```typescript
// src/utils/math.test.ts
import { describe, it, expect } from "vitest";
import { add, divide, clamp } from "./math";

describe("math utilities", () => {
  describe("add", () => {
    it("adds two positive numbers", () => {
      expect(add(2, 3)).toBe(5);
    });

    it("handles negative numbers", () => {
      expect(add(-1, 1)).toBe(0);
    });
  });

  describe("divide", () => {
    it("divides two numbers", () => {
      expect(divide(10, 2)).toBe(5);
    });

    it("throws on division by zero", () => {
      expect(() => divide(1, 0)).toThrow("Cannot divide by zero");
    });
  });

  describe("clamp", () => {
    it.each([
      { value: 5, min: 0, max: 10, expected: 5 },
      { value: -5, min: 0, max: 10, expected: 0 },
      { value: 15, min: 0, max: 10, expected: 10 },
    ])("clamp($value, $min, $max) = $expected", ({ value, min, max, expected }) => {
      expect(clamp(value, min, max)).toBe(expected);
    });
  });
});
```

## Component Testing

```typescript
// src/components/Counter.test.tsx
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, it, expect } from "vitest";
import { Counter } from "./Counter";

describe("Counter", () => {
  it("renders initial count", () => {
    render(<Counter initialCount={5} />);
    expect(screen.getByText("Count: 5")).toBeInTheDocument();
  });

  it("increments count on button click", async () => {
    const user = userEvent.setup();
    render(<Counter initialCount={0} />);

    await user.click(screen.getByRole("button", { name: "Increment" }));

    expect(screen.getByText("Count: 1")).toBeInTheDocument();
  });

  it("calls onChange callback", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(<Counter initialCount={0} onChange={onChange} />);

    await user.click(screen.getByRole("button", { name: "Increment" }));

    expect(onChange).toHaveBeenCalledWith(1);
  });

  it("disables decrement at zero", () => {
    render(<Counter initialCount={0} />);
    expect(screen.getByRole("button", { name: "Decrement" })).toBeDisabled();
  });
});

// Testing forms
describe("LoginForm", () => {
  it("submits with email and password", async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn();
    render(<LoginForm onSubmit={onSubmit} />);

    await user.type(screen.getByLabelText("Email"), "test@example.com");
    await user.type(screen.getByLabelText("Password"), "password123");
    await user.click(screen.getByRole("button", { name: "Sign in" }));

    expect(onSubmit).toHaveBeenCalledWith({
      email: "test@example.com",
      password: "password123",
    });
  });

  it("shows validation errors", async () => {
    const user = userEvent.setup();
    render(<LoginForm onSubmit={vi.fn()} />);

    await user.click(screen.getByRole("button", { name: "Sign in" }));

    expect(screen.getByText("Email is required")).toBeInTheDocument();
    expect(screen.getByText("Password is required")).toBeInTheDocument();
  });
});
```

## Mocking

```typescript
// Mock modules
vi.mock("./api", () => ({
  fetchUser: vi.fn(),
  updateUser: vi.fn(),
}));

import { fetchUser } from "./api";

it("fetches user data", async () => {
  vi.mocked(fetchUser).mockResolvedValue({ id: "1", name: "Alice" });

  const user = await fetchUser("1");
  expect(user.name).toBe("Alice");
  expect(fetchUser).toHaveBeenCalledWith("1");
});

// Mock with factory (auto-mock all exports)
vi.mock("./analytics", () => ({
  track: vi.fn(),
  identify: vi.fn(),
  page: vi.fn(),
}));

// Mock with importOriginal (partial mock)
vi.mock("./utils", async (importOriginal) => {
  const actual = await importOriginal<typeof import("./utils")>();
  return {
    ...actual,
    generateId: vi.fn(() => "mock-id-123"),
  };
});

// Spy on methods
const consoleSpy = vi.spyOn(console, "error").mockImplementation(() => {});
// ... test ...
expect(consoleSpy).toHaveBeenCalledWith("Error message");
consoleSpy.mockRestore();

// Mock timers
vi.useFakeTimers();

it("debounces search", async () => {
  const onSearch = vi.fn();
  render(<SearchInput onSearch={onSearch} debounceMs={300} />);

  await userEvent.type(screen.getByRole("textbox"), "hello");

  expect(onSearch).not.toHaveBeenCalled();
  vi.advanceTimersByTime(300);
  expect(onSearch).toHaveBeenCalledWith("hello");
});

vi.useRealTimers();

// Mock fetch
globalThis.fetch = vi.fn();

it("handles API response", async () => {
  vi.mocked(fetch).mockResolvedValue(
    new Response(JSON.stringify({ data: "test" }), {
      status: 200,
      headers: { "Content-Type": "application/json" },
    })
  );

  const result = await myApiCall();
  expect(result.data).toBe("test");
});

// Mock environment variables
it("uses production API in prod", () => {
  vi.stubEnv("VITE_API_URL", "https://api.prod.com");

  expect(getApiUrl()).toBe("https://api.prod.com");

  vi.unstubAllEnvs();
});
```

## Snapshots

```typescript
// Inline snapshots (auto-updated by Vitest)
it("formats user display name", () => {
  expect(formatDisplayName({ first: "Alice", last: "Smith" }))
    .toMatchInlineSnapshot(`"Alice Smith"`);
});

// File snapshots
it("renders correctly", () => {
  const { container } = render(<UserCard user={mockUser} />);
  expect(container.innerHTML).toMatchSnapshot();
});

// Update snapshots: npx vitest --update
```

## Async Testing

```typescript
// Testing async functions
it("fetches data", async () => {
  const data = await fetchData();
  expect(data).toEqual({ items: [] });
});

// Testing React hooks
import { renderHook, waitFor } from "@testing-library/react";

it("useUsers returns user list", async () => {
  const { result } = renderHook(() => useUsers());

  await waitFor(() => {
    expect(result.current.isLoading).toBe(false);
  });

  expect(result.current.users).toHaveLength(3);
});

// Testing with React Query
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
    },
  });

  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
}

it("useUser fetches user", async () => {
  const { result } = renderHook(() => useUser("1"), {
    wrapper: createWrapper(),
  });

  await waitFor(() => expect(result.current.isSuccess).toBe(true));
  expect(result.current.data?.name).toBe("Alice");
});
```

## CLI Commands

```bash
# Run all tests
npx vitest

# Run once (CI mode)
npx vitest run

# Run specific file
npx vitest src/utils/math.test.ts

# Run tests matching name
npx vitest -t "should add"

# Watch mode (default)
npx vitest --watch

# Coverage
npx vitest run --coverage

# UI mode (browser-based test viewer)
npx vitest --ui

# Type checking
npx vitest typecheck

# Update snapshots
npx vitest --update

# Specific reporter
npx vitest run --reporter=json --outputFile=results.json
```

## Gotchas

1. **`globals: true` needs a TypeScript config update.** Add `"types": ["vitest/globals"]` to `tsconfig.json` `compilerOptions` so TypeScript recognizes `describe`, `it`, `expect` without imports.

2. **React Testing Library's `render` doesn't auto-cleanup in Vitest.** Unlike Jest, Vitest doesn't auto-cleanup. Add `afterEach(() => cleanup())` in your setup file, or import `@testing-library/react/pure` and call `cleanup()` manually.

3. **`vi.mock()` is hoisted to the top of the file.** Even if you write it inside a `describe` block, it applies to the entire file. Use `vi.doMock()` for non-hoisted mocks (requires dynamic `import()`).

4. **Vitest reuses Vite's config.** Path aliases, plugins, and defines from `vite.config.ts` work automatically in tests. No need to duplicate config. But if you have a separate `vitest.config.ts`, it REPLACES `vite.config.ts` (use `mergeConfig` to combine).

5. **`happy-dom` is faster than `jsdom` but less complete.** Use `happy-dom` for simple component tests (2-5x faster). Switch to `jsdom` if you need `ResizeObserver`, `IntersectionObserver`, or complex DOM APIs.

6. **`vi.stubEnv()` only works in the current test.** It doesn't affect imports that already captured `import.meta.env` at module load time. If a module reads env vars at import time, you need `vi.mock()` or set env before the import.
