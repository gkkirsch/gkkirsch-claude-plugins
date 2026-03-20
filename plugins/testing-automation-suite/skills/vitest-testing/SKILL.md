---
name: vitest-testing
description: >
  Vitest unit and integration testing — setup, matchers, mocking, spies,
  snapshot testing, coverage, testing React components with Testing Library,
  testing hooks, and async testing patterns.
  Triggers: "vitest", "unit testing", "testing library", "test react component",
  "mock function", "test coverage", "snapshot test".
  NOT for: E2E testing (use playwright-e2e).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Vitest Testing

## Setup

```bash
npm install -D vitest @testing-library/react @testing-library/jest-dom @testing-library/user-event jsdom
```

```typescript
// vitest.config.ts
import { defineConfig } from "vitest/config";
import react from "@vitejs/plugin-react";
import path from "path";

export default defineConfig({
  plugins: [react()],
  test: {
    globals: true,
    environment: "jsdom",
    setupFiles: ["./src/test/setup.ts"],
    include: ["src/**/*.{test,spec}.{ts,tsx}"],
    coverage: {
      provider: "v8",
      reporter: ["text", "lcov", "html"],
      include: ["src/**/*.{ts,tsx}"],
      exclude: ["src/test/**", "src/**/*.d.ts", "src/main.tsx"],
      thresholds: { statements: 80, branches: 75, functions: 80, lines: 80 },
    },
  },
  resolve: {
    alias: { "@": path.resolve(__dirname, "./src") },
  },
});
```

```typescript
// src/test/setup.ts
import "@testing-library/jest-dom/vitest";
import { cleanup } from "@testing-library/react";
import { afterEach } from "vitest";

afterEach(() => cleanup());
```

## Basic Tests

```typescript
import { describe, it, expect } from "vitest";

describe("math utils", () => {
  it("adds numbers correctly", () => {
    expect(add(2, 3)).toBe(5);
  });

  it("handles negative numbers", () => {
    expect(add(-1, -2)).toBe(-3);
  });

  it("throws on invalid input", () => {
    expect(() => add(NaN, 1)).toThrow("Invalid number");
  });
});

// Common matchers
expect(value).toBe(expected);           // Strict equality (===)
expect(value).toEqual(expected);        // Deep equality
expect(value).toStrictEqual(expected);  // Deep equality + undefined properties
expect(value).toBeTruthy();
expect(value).toBeFalsy();
expect(value).toBeNull();
expect(value).toBeDefined();
expect(value).toBeGreaterThan(3);
expect(value).toContain("substring");
expect(array).toHaveLength(3);
expect(obj).toHaveProperty("key", "value");
expect(fn).toThrow(/error message/);
expect(value).toMatchObject({ name: "Alice" });  // Partial match
```

## Mocking

```typescript
import { vi, describe, it, expect, beforeEach } from "vitest";

// Mock a module
vi.mock("@/lib/api", () => ({
  fetchUsers: vi.fn(),
  createUser: vi.fn(),
}));

import { fetchUsers, createUser } from "@/lib/api";

describe("UserService", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("fetches users", async () => {
    const mockUsers = [{ id: "1", name: "Alice" }];
    vi.mocked(fetchUsers).mockResolvedValue(mockUsers);

    const result = await fetchUsers();

    expect(fetchUsers).toHaveBeenCalledOnce();
    expect(result).toEqual(mockUsers);
  });

  it("handles API errors", async () => {
    vi.mocked(fetchUsers).mockRejectedValue(new Error("Network error"));

    await expect(fetchUsers()).rejects.toThrow("Network error");
  });
});

// Spy on object methods
const obj = { greet: (name: string) => `Hello, ${name}` };
const spy = vi.spyOn(obj, "greet");

obj.greet("Alice");

expect(spy).toHaveBeenCalledWith("Alice");
expect(spy).toHaveReturnedWith("Hello, Alice");

// Mock timers
vi.useFakeTimers();

setTimeout(callback, 1000);
vi.advanceTimersByTime(1000);
expect(callback).toHaveBeenCalled();

vi.useRealTimers();

// Mock environment variables
vi.stubEnv("API_URL", "https://test.example.com");
// process.env.API_URL === "https://test.example.com"
vi.unstubAllEnvs();

// Partial mock (mock some exports, keep others)
vi.mock("@/lib/utils", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/lib/utils")>();
  return {
    ...actual,
    generateId: vi.fn().mockReturnValue("test-id-123"),
  };
});
```

## Testing React Components

```tsx
import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, it, expect, vi } from "vitest";
import { LoginForm } from "./LoginForm";

describe("LoginForm", () => {
  it("renders form fields", () => {
    render(<LoginForm onSubmit={vi.fn()} />);

    expect(screen.getByLabelText(/email/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/password/i)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /sign in/i })).toBeInTheDocument();
  });

  it("validates required fields", async () => {
    const user = userEvent.setup();
    render(<LoginForm onSubmit={vi.fn()} />);

    await user.click(screen.getByRole("button", { name: /sign in/i }));

    expect(screen.getByText(/email is required/i)).toBeInTheDocument();
  });

  it("submits with valid data", async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn();
    render(<LoginForm onSubmit={onSubmit} />);

    await user.type(screen.getByLabelText(/email/i), "alice@example.com");
    await user.type(screen.getByLabelText(/password/i), "password123");
    await user.click(screen.getByRole("button", { name: /sign in/i }));

    expect(onSubmit).toHaveBeenCalledWith({
      email: "alice@example.com",
      password: "password123",
    });
  });

  it("disables submit button while loading", () => {
    render(<LoginForm onSubmit={vi.fn()} isLoading />);

    expect(screen.getByRole("button", { name: /signing in/i })).toBeDisabled();
  });
});
```

### Query Priority

| Priority | Query | Use For |
|----------|-------|---------|
| 1 | `getByRole` | Buttons, links, headings, checkboxes |
| 2 | `getByLabelText` | Form inputs with labels |
| 3 | `getByPlaceholderText` | Inputs with placeholder |
| 4 | `getByText` | Non-interactive text content |
| 5 | `getByTestId` | Last resort — when no semantic query works |

### Query Variants

| Variant | Returns | Throws? | Use For |
|---------|---------|---------|---------|
| `getBy` | Element | Yes, if not found | Element must exist |
| `queryBy` | Element \| null | No | Element may not exist |
| `findBy` | Promise\<Element\> | Yes, if timeout | Async/delayed elements |
| `getAllBy` | Element[] | Yes, if none found | Multiple elements |

## Testing Hooks

```tsx
import { renderHook, act } from "@testing-library/react";
import { useCounter } from "./useCounter";

describe("useCounter", () => {
  it("initializes with default value", () => {
    const { result } = renderHook(() => useCounter());
    expect(result.current.count).toBe(0);
  });

  it("initializes with custom value", () => {
    const { result } = renderHook(() => useCounter(10));
    expect(result.current.count).toBe(10);
  });

  it("increments counter", () => {
    const { result } = renderHook(() => useCounter());

    act(() => {
      result.current.increment();
    });

    expect(result.current.count).toBe(1);
  });

  it("respects max value", () => {
    const { result } = renderHook(() => useCounter(9, { max: 10 }));

    act(() => {
      result.current.increment();
      result.current.increment();  // Should not exceed max
    });

    expect(result.current.count).toBe(10);
  });
});
```

## Async Testing

```typescript
describe("async operations", () => {
  it("waits for async result", async () => {
    const result = await fetchData("user-123");
    expect(result).toEqual({ id: "user-123", name: "Alice" });
  });

  it("waits for element to appear", async () => {
    render(<UserProfile userId="123" />);

    // findBy waits up to 1000ms by default
    const name = await screen.findByText("Alice");
    expect(name).toBeInTheDocument();
  });

  it("waits for element to disappear", async () => {
    render(<LoadingComponent />);

    await waitForElementToBeRemoved(() => screen.queryByText("Loading..."));
    expect(screen.getByText("Content loaded")).toBeInTheDocument();
  });

  it("tests with fake timers and async", async () => {
    vi.useFakeTimers();
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });

    render(<DebouncedSearch onSearch={onSearch} />);

    await user.type(screen.getByRole("textbox"), "hello");
    await vi.advanceTimersByTimeAsync(300);

    expect(onSearch).toHaveBeenCalledWith("hello");
    vi.useRealTimers();
  });
});
```

## Snapshot Testing

```typescript
it("renders correctly", () => {
  const { container } = render(<Badge variant="success">Active</Badge>);
  expect(container.firstChild).toMatchSnapshot();
});

// Inline snapshot (stored in test file)
it("formats date correctly", () => {
  expect(formatDate(new Date("2026-03-19"))).toMatchInlineSnapshot('"March 19, 2026"');
});

// Update snapshots: vitest --update or press 'u' in watch mode
```

## CLI Commands

```bash
vitest                  # Run in watch mode
vitest run              # Run once
vitest run --coverage   # With coverage report
vitest run src/utils    # Run specific directory
vitest run -t "login"   # Run tests matching name
vitest --reporter=verbose  # Detailed output
vitest --ui             # Open browser UI
```

## Gotchas

1. **`userEvent` must use `.setup()` and `await`** — `userEvent.click(el)` without setup doesn't handle async correctly. Always: `const user = userEvent.setup()` then `await user.click(el)`.

2. **`screen.getByRole("button")` is preferred over `getByText`** — Role queries test accessibility. If your button isn't findable by role, it's probably missing accessible markup.

3. **`vi.clearAllMocks()` vs `vi.resetAllMocks()`** — `clearAllMocks` resets call history but keeps implementation. `resetAllMocks` also removes mock implementation. Use `clearAllMocks` in `beforeEach` unless you want to re-mock every test.

4. **Act warnings** — "An update was not wrapped in act()" means state updated outside React's batch. Wrap the trigger in `act()`, or use `findBy` which includes act internally.

5. **Testing Library cleanup** — If not using `globals: true` with the setup file, you must call `cleanup()` in `afterEach`. Missing cleanup causes tests to leak state between runs.

6. **Mock hoisting** — `vi.mock()` calls are hoisted to the top of the file. You can't reference variables from the test scope inside the mock factory. Use `vi.hoisted()` for shared mock values.
