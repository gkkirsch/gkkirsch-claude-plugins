# React Testing Guide

Quick-reference guide for testing React and Next.js applications with Testing Library, Vitest, MSW, and accessibility testing. Consult this when writing or reviewing tests.

---

## Testing Stack

| Tool | Purpose |
|------|---------|
| **Vitest** | Test runner (fast, Vite-native, Jest-compatible API) |
| **@testing-library/react** | Component rendering and querying |
| **@testing-library/user-event** | Simulating real user interactions |
| **@testing-library/jest-dom** | Custom DOM matchers (toBeVisible, toHaveTextContent) |
| **MSW** (Mock Service Worker) | Mocking API requests at the network level |
| **vitest-axe** / **jest-axe** | Automated accessibility checks |
| **Playwright** | End-to-end testing |

---

## Testing Library Principles

1. **Query by what the user sees** — text, labels, roles, placeholders
2. **Avoid testing implementation details** — no querying by class name, internal state, or component instance
3. **Test behavior, not structure** — what happens when user clicks, types, submits
4. **Priority of queries** (most to least preferred):

```
1. getByRole          — Accessible role (button, heading, textbox, etc.)
2. getByLabelText     — Form elements by their label
3. getByPlaceholderText — Inputs by placeholder
4. getByText          — Non-interactive elements by text content
5. getByDisplayValue  — Inputs by current value
6. getByAltText       — Images by alt text
7. getByTitle         — Elements by title attribute
8. getByTestId        — Last resort — data-testid attribute
```

---

## Query Variants

```tsx
// getBy*  — Returns element or throws (synchronous)
// Use for: elements that should be present
const button = screen.getByRole('button', { name: 'Submit' });

// queryBy* — Returns element or null (synchronous)
// Use for: asserting elements are NOT present
expect(screen.queryByText('Error')).not.toBeInTheDocument();

// findBy* — Returns promise, waits for element (async)
// Use for: elements that appear after async operations
const message = await screen.findByText('Success');

// *AllBy* variants — Return array of elements
const items = screen.getAllByRole('listitem');
expect(items).toHaveLength(3);
```

---

## Component Testing Patterns

### Basic Component Test

```tsx
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, it, expect, vi } from 'vitest';
import { Counter } from './Counter';

describe('Counter', () => {
  it('renders with initial count', () => {
    render(<Counter initialCount={5} />);
    expect(screen.getByText('Count: 5')).toBeInTheDocument();
  });

  it('increments count on click', async () => {
    const user = userEvent.setup();
    render(<Counter initialCount={0} />);

    await user.click(screen.getByRole('button', { name: 'Increment' }));
    expect(screen.getByText('Count: 1')).toBeInTheDocument();
  });

  it('calls onChange when count changes', async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(<Counter initialCount={0} onChange={onChange} />);

    await user.click(screen.getByRole('button', { name: 'Increment' }));
    expect(onChange).toHaveBeenCalledWith(1);
  });
});
```

### Form Testing

```tsx
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

describe('LoginForm', () => {
  it('submits with valid credentials', async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn();
    render(<LoginForm onSubmit={onSubmit} />);

    await user.type(screen.getByLabelText('Email'), 'user@example.com');
    await user.type(screen.getByLabelText('Password'), 'password123');
    await user.click(screen.getByRole('button', { name: 'Sign In' }));

    await waitFor(() => {
      expect(onSubmit).toHaveBeenCalledWith({
        email: 'user@example.com',
        password: 'password123',
      });
    });
  });

  it('shows validation errors for empty fields', async () => {
    const user = userEvent.setup();
    render(<LoginForm onSubmit={vi.fn()} />);

    await user.click(screen.getByRole('button', { name: 'Sign In' }));

    expect(await screen.findByText('Email is required')).toBeInTheDocument();
    expect(screen.getByText('Password is required')).toBeInTheDocument();
  });

  it('shows error for invalid email', async () => {
    const user = userEvent.setup();
    render(<LoginForm onSubmit={vi.fn()} />);

    await user.type(screen.getByLabelText('Email'), 'not-an-email');
    await user.click(screen.getByRole('button', { name: 'Sign In' }));

    expect(await screen.findByText('Please enter a valid email')).toBeInTheDocument();
  });

  it('disables submit button while submitting', async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn(() => new Promise(resolve => setTimeout(resolve, 100)));
    render(<LoginForm onSubmit={onSubmit} />);

    await user.type(screen.getByLabelText('Email'), 'user@example.com');
    await user.type(screen.getByLabelText('Password'), 'password123');
    await user.click(screen.getByRole('button', { name: 'Sign In' }));

    expect(screen.getByRole('button', { name: /signing in/i })).toBeDisabled();

    await waitFor(() => {
      expect(screen.getByRole('button', { name: 'Sign In' })).toBeEnabled();
    });
  });
});
```

### Testing with Context/Providers

```tsx
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

function renderWithProviders(
  ui: React.ReactElement,
  options?: { user?: User; theme?: string }
) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });

  return render(
    <QueryClientProvider client={queryClient}>
      <AuthContext.Provider value={{ user: options?.user ?? null }}>
        <ThemeContext.Provider value={options?.theme ?? 'light'}>
          {ui}
        </ThemeContext.Provider>
      </AuthContext.Provider>
    </QueryClientProvider>
  );
}

// Usage
it('shows user name when authenticated', () => {
  renderWithProviders(<Header />, {
    user: { id: '1', name: 'Alice', email: 'alice@example.com' },
  });

  expect(screen.getByText('Alice')).toBeInTheDocument();
});

it('shows login button when unauthenticated', () => {
  renderWithProviders(<Header />);
  expect(screen.getByRole('link', { name: 'Sign In' })).toBeInTheDocument();
});
```

---

## Custom Hook Testing

```tsx
import { renderHook, act } from '@testing-library/react';
import { useCounter } from './useCounter';

describe('useCounter', () => {
  it('starts with initial value', () => {
    const { result } = renderHook(() => useCounter(10));
    expect(result.current.count).toBe(10);
  });

  it('increments count', () => {
    const { result } = renderHook(() => useCounter(0));

    act(() => {
      result.current.increment();
    });

    expect(result.current.count).toBe(1);
  });

  it('resets to initial value', () => {
    const { result } = renderHook(() => useCounter(5));

    act(() => {
      result.current.increment();
      result.current.increment();
    });
    expect(result.current.count).toBe(7);

    act(() => {
      result.current.reset();
    });
    expect(result.current.count).toBe(5);
  });
});

// Testing hooks with dependencies
describe('useDebounce', () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('debounces value changes', () => {
    const { result, rerender } = renderHook(
      ({ value, delay }) => useDebounce(value, delay),
      { initialProps: { value: 'hello', delay: 300 } }
    );

    expect(result.current).toBe('hello');

    rerender({ value: 'world', delay: 300 });
    expect(result.current).toBe('hello'); // Not yet

    act(() => {
      vi.advanceTimersByTime(300);
    });
    expect(result.current).toBe('world'); // Now updated
  });
});
```

---

## MSW (Mock Service Worker)

### Setup

```tsx
// mocks/handlers.ts
import { http, HttpResponse } from 'msw';

export const handlers = [
  // GET request
  http.get('/api/users', () => {
    return HttpResponse.json([
      { id: '1', name: 'Alice', email: 'alice@example.com' },
      { id: '2', name: 'Bob', email: 'bob@example.com' },
    ]);
  }),

  // GET with params
  http.get('/api/users/:id', ({ params }) => {
    const { id } = params;
    return HttpResponse.json({ id, name: 'Alice', email: 'alice@example.com' });
  }),

  // POST request
  http.post('/api/users', async ({ request }) => {
    const body = await request.json();
    return HttpResponse.json(
      { id: '3', ...body },
      { status: 201 }
    );
  }),

  // Error response
  http.delete('/api/users/:id', () => {
    return HttpResponse.json(
      { error: 'Not found' },
      { status: 404 }
    );
  }),
];

// mocks/server.ts
import { setupServer } from 'msw/node';
import { handlers } from './handlers';

export const server = setupServer(...handlers);

// vitest.setup.ts
import { server } from './mocks/server';

beforeAll(() => server.listen({ onUnhandledRequest: 'error' }));
afterEach(() => server.resetHandlers());
afterAll(() => server.close());
```

### Per-Test Overrides

```tsx
import { server } from '../mocks/server';
import { http, HttpResponse } from 'msw';

it('shows error when API fails', async () => {
  // Override handler for this test only
  server.use(
    http.get('/api/users', () => {
      return HttpResponse.json(
        { error: 'Internal Server Error' },
        { status: 500 }
      );
    })
  );

  render(<UserList />);

  expect(await screen.findByText('Failed to load users')).toBeInTheDocument();
});

it('shows empty state when no users', async () => {
  server.use(
    http.get('/api/users', () => {
      return HttpResponse.json([]);
    })
  );

  render(<UserList />);

  expect(await screen.findByText('No users found')).toBeInTheDocument();
});
```

---

## Accessibility Testing

### Automated Checks with axe

```tsx
import { axe, toHaveNoViolations } from 'jest-axe';

expect.extend(toHaveNoViolations);

it('has no accessibility violations', async () => {
  const { container } = render(<LoginForm />);
  const results = await axe(container);
  expect(results).toHaveNoViolations();
});

// Check specific rules
it('form labels are associated with inputs', async () => {
  const { container } = render(<LoginForm />);
  const results = await axe(container, {
    rules: {
      'label': { enabled: true },
      'color-contrast': { enabled: true },
    },
  });
  expect(results).toHaveNoViolations();
});
```

### Manual Accessibility Checks

```tsx
it('has proper ARIA attributes on dialog', async () => {
  const user = userEvent.setup();
  render(<DialogComponent />);

  await user.click(screen.getByRole('button', { name: 'Open Dialog' }));

  const dialog = screen.getByRole('dialog');
  expect(dialog).toHaveAttribute('aria-labelledby');
  expect(dialog).toHaveAttribute('aria-describedby');

  // Check focus is trapped
  expect(dialog).toContainElement(document.activeElement as HTMLElement);
});

it('supports keyboard navigation', async () => {
  const user = userEvent.setup();
  render(<DropdownMenu items={['Edit', 'Delete', 'Share']} />);

  // Open with Enter
  await user.tab();
  await user.keyboard('{Enter}');

  // Navigate with arrow keys
  expect(screen.getByRole('menuitem', { name: 'Edit' })).toHaveFocus();

  await user.keyboard('{ArrowDown}');
  expect(screen.getByRole('menuitem', { name: 'Delete' })).toHaveFocus();

  // Close with Escape
  await user.keyboard('{Escape}');
  expect(screen.queryByRole('menu')).not.toBeInTheDocument();
});

it('announces live region updates to screen readers', async () => {
  const user = userEvent.setup();
  render(<NotificationSystem />);

  await user.click(screen.getByRole('button', { name: 'Save' }));

  const alert = await screen.findByRole('alert');
  expect(alert).toHaveTextContent('Changes saved successfully');
});
```

---

## Integration Testing

### Testing Data Fetching Components

```tsx
describe('UserList', () => {
  it('loads and displays users', async () => {
    render(<UserList />);

    // Shows loading state
    expect(screen.getByText('Loading...')).toBeInTheDocument();

    // Waits for data
    expect(await screen.findByText('Alice')).toBeInTheDocument();
    expect(screen.getByText('Bob')).toBeInTheDocument();
  });

  it('supports search filtering', async () => {
    const user = userEvent.setup();
    render(<UserList />);

    // Wait for initial load
    await screen.findByText('Alice');

    // Type in search
    await user.type(screen.getByPlaceholderText('Search users...'), 'Alice');

    // Only Alice should be visible
    expect(screen.getByText('Alice')).toBeInTheDocument();
    expect(screen.queryByText('Bob')).not.toBeInTheDocument();
  });

  it('handles pagination', async () => {
    const user = userEvent.setup();
    render(<UserList />);

    await screen.findByText('Alice');

    // Click next page
    await user.click(screen.getByRole('button', { name: 'Next page' }));

    // Shows page 2 users
    expect(await screen.findByText('Charlie')).toBeInTheDocument();
  });
});
```

### Testing User Flows

```tsx
describe('Shopping Cart Flow', () => {
  it('adds item to cart and completes checkout', async () => {
    const user = userEvent.setup();
    render(<App />);

    // Browse products
    await screen.findByText('Product A');
    await user.click(screen.getByRole('button', { name: 'Add Product A to cart' }));

    // Cart updates
    expect(screen.getByText('Cart (1)')).toBeInTheDocument();

    // Go to cart
    await user.click(screen.getByRole('link', { name: /cart/i }));
    expect(screen.getByText('Product A')).toBeInTheDocument();
    expect(screen.getByText('$29.99')).toBeInTheDocument();

    // Proceed to checkout
    await user.click(screen.getByRole('button', { name: 'Checkout' }));

    // Fill shipping info
    await user.type(screen.getByLabelText('Full Name'), 'John Doe');
    await user.type(screen.getByLabelText('Address'), '123 Main St');
    await user.click(screen.getByRole('button', { name: 'Place Order' }));

    // Confirmation
    expect(await screen.findByText('Order Confirmed!')).toBeInTheDocument();
  });
});
```

---

## Vitest Configuration

```tsx
// vitest.config.ts
import { defineConfig } from 'vitest/config';
import react from '@vitejs/plugin-react';
import path from 'path';

export default defineConfig({
  plugins: [react()],
  test: {
    environment: 'jsdom',
    globals: true,
    setupFiles: ['./vitest.setup.ts'],
    include: ['**/*.{test,spec}.{ts,tsx}'],
    coverage: {
      provider: 'v8',
      reporter: ['text', 'html', 'lcov'],
      include: ['src/**/*.{ts,tsx}'],
      exclude: ['**/*.test.*', '**/*.spec.*', '**/types/**'],
    },
  },
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
});

// vitest.setup.ts
import '@testing-library/jest-dom/vitest';
import { cleanup } from '@testing-library/react';
import { afterEach } from 'vitest';

afterEach(() => {
  cleanup();
});
```

---

## Testing Patterns Cheat Sheet

| What to Test | How |
|-------------|-----|
| Component renders | `render()` + `getByText/getByRole` |
| User clicks | `userEvent.click()` |
| User types | `userEvent.type()` |
| Form submission | Fill fields + click submit + check result |
| Async content appears | `findByText()` / `waitFor()` |
| Content disappears | `queryByText()` + `not.toBeInTheDocument()` |
| API call made | MSW handler + check rendered data |
| Error state | MSW error handler + check error UI |
| Loading state | Check loading indicator before data |
| Accessibility | `axe()` + role queries + keyboard tests |
| Custom hooks | `renderHook()` + `act()` |
| Navigation | Check URL / rendered page content |
| Responsive | Set viewport + check visibility |

---

## Common Mistakes to Avoid

| Mistake | Fix |
|---------|-----|
| Using `getBy` for absent elements | Use `queryBy` instead |
| Not awaiting `findBy` | Always `await screen.findByText(...)` |
| Using `fireEvent` instead of `userEvent` | `userEvent.setup()` simulates real user behavior |
| Testing internal state | Test what the user sees/experiences |
| Mocking too much | Use MSW for API mocking, render real components |
| No cleanup between tests | Use `afterEach(() => cleanup())` |
| Hardcoded test data everywhere | Use factory functions or fixtures |
| Testing CSS classes | Test visible behavior (visibility, content, role) |
| Wrapping every assertion in waitFor | Only wrap assertions that need to wait for async ops |
| Not testing error/loading states | Always test the unhappy path |
