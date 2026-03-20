---
name: mobile-testing
description: >
  Mobile app testing patterns for React Native with Jest and Detox.
  Use when writing unit tests, component tests, integration tests,
  or E2E tests for mobile applications.
  Triggers: "mobile test", "react native test", "detox", "jest native",
  "snapshot test", "component test", "e2e mobile", "app testing".
  NOT for: web-only testing, Playwright, or Cypress.
version: 1.0.0
allowed-tools: Read, Grep, Glob, Bash
---

# Mobile Testing Patterns

## Component Testing with React Native Testing Library

```tsx
import { render, screen, fireEvent, waitFor } from '@testing-library/react-native';
import { LoginScreen } from '../LoginScreen';

// Wrapper with providers
function renderWithProviders(ui: React.ReactElement) {
  return render(
    <AuthProvider>
      <NavigationContainer>
        {ui}
      </NavigationContainer>
    </AuthProvider>
  );
}

describe('LoginScreen', () => {
  it('shows validation errors on empty submit', async () => {
    renderWithProviders(<LoginScreen />);

    fireEvent.press(screen.getByText('Sign In'));

    await waitFor(() => {
      expect(screen.getByText('Email is required')).toBeTruthy();
      expect(screen.getByText('Password is required')).toBeTruthy();
    });
  });

  it('calls login with credentials', async () => {
    const mockLogin = jest.fn().mockResolvedValue({ token: 'abc' });
    renderWithProviders(<LoginScreen onLogin={mockLogin} />);

    fireEvent.changeText(screen.getByPlaceholderText('Email'), 'user@test.com');
    fireEvent.changeText(screen.getByPlaceholderText('Password'), 'password123');
    fireEvent.press(screen.getByText('Sign In'));

    await waitFor(() => {
      expect(mockLogin).toHaveBeenCalledWith({
        email: 'user@test.com',
        password: 'password123',
      });
    });
  });

  it('shows loading state during submission', async () => {
    const mockLogin = jest.fn(() => new Promise(resolve => setTimeout(resolve, 1000)));
    renderWithProviders(<LoginScreen onLogin={mockLogin} />);

    fireEvent.changeText(screen.getByPlaceholderText('Email'), 'user@test.com');
    fireEvent.changeText(screen.getByPlaceholderText('Password'), 'pass');
    fireEvent.press(screen.getByText('Sign In'));

    expect(screen.getByTestId('loading-spinner')).toBeTruthy();
    expect(screen.getByText('Sign In')).toBeDisabled();
  });
});
```

## Hook Testing

```tsx
import { renderHook, act, waitFor } from '@testing-library/react-native';
import { useApi } from '../useApi';

describe('useApi', () => {
  it('fetches data successfully', async () => {
    global.fetch = jest.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({ users: [{ id: 1, name: 'Test' }] }),
    });

    const { result } = renderHook(() => useApi('/users'));

    expect(result.current.loading).toBe(true);

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
      expect(result.current.data).toEqual({ users: [{ id: 1, name: 'Test' }] });
      expect(result.current.error).toBeNull();
    });
  });

  it('handles network errors', async () => {
    global.fetch = jest.fn().mockRejectedValue(new Error('Network error'));

    const { result } = renderHook(() => useApi('/users'));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
      expect(result.current.error).toBe('Network error');
    });
  });

  it('refetches on manual trigger', async () => {
    let callCount = 0;
    global.fetch = jest.fn().mockImplementation(() => {
      callCount++;
      return Promise.resolve({
        ok: true,
        json: () => Promise.resolve({ count: callCount }),
      });
    });

    const { result } = renderHook(() => useApi('/count'));

    await waitFor(() => expect(result.current.data?.count).toBe(1));

    await act(async () => {
      await result.current.refetch();
    });

    expect(result.current.data?.count).toBe(2);
  });
});
```

## Mocking Native Modules

```typescript
// jest.setup.ts
jest.mock('@react-native-async-storage/async-storage', () =>
  require('@react-native-async-storage/async-storage/jest/async-storage-mock')
);

jest.mock('expo-secure-store', () => ({
  getItemAsync: jest.fn(),
  setItemAsync: jest.fn(),
  deleteItemAsync: jest.fn(),
}));

jest.mock('@react-native-community/netinfo', () => ({
  fetch: jest.fn().mockResolvedValue({ isConnected: true, isInternetReachable: true }),
  addEventListener: jest.fn(() => jest.fn()),
}));

jest.mock('expo-camera', () => ({
  Camera: {
    useCameraPermissions: jest.fn(() => [{ granted: true }, jest.fn()]),
  },
}));

jest.mock('expo-notifications', () => ({
  getPermissionsAsync: jest.fn().mockResolvedValue({ status: 'granted' }),
  requestPermissionsAsync: jest.fn().mockResolvedValue({ status: 'granted' }),
  getExpoPushTokenAsync: jest.fn().mockResolvedValue({ data: 'ExponentPushToken[xxx]' }),
  addNotificationReceivedListener: jest.fn(() => ({ remove: jest.fn() })),
}));

// Platform mock
jest.mock('react-native/Libraries/Utilities/Platform', () => ({
  OS: 'ios',
  select: jest.fn(obj => obj.ios),
}));
```

## Detox E2E Testing

```typescript
// e2e/login.test.ts
import { device, element, by, expect } from 'detox';

describe('Login Flow', () => {
  beforeAll(async () => {
    await device.launchApp({ newInstance: true });
  });

  beforeEach(async () => {
    await device.reloadReactNative();
  });

  it('should login with valid credentials', async () => {
    await element(by.id('email-input')).typeText('user@test.com');
    await element(by.id('password-input')).typeText('password123');
    await element(by.id('login-button')).tap();

    await waitFor(element(by.id('home-screen')))
      .toBeVisible()
      .withTimeout(5000);
  });

  it('should show error for invalid credentials', async () => {
    await element(by.id('email-input')).typeText('wrong@test.com');
    await element(by.id('password-input')).typeText('wrong');
    await element(by.id('login-button')).tap();

    await waitFor(element(by.text('Invalid credentials')))
      .toBeVisible()
      .withTimeout(3000);
  });

  it('should handle keyboard dismiss', async () => {
    await element(by.id('email-input')).tap();
    await element(by.id('email-input')).typeText('test');

    // Dismiss keyboard
    await device.pressBack(); // Android
    // OR: await element(by.id('scroll-view')).tap(); // tap outside

    await expect(element(by.id('login-button'))).toBeVisible();
  });
});
```

## Snapshot Testing

```tsx
import { render } from '@testing-library/react-native';
import { ProfileCard } from '../ProfileCard';

describe('ProfileCard', () => {
  it('matches snapshot', () => {
    const tree = render(
      <ProfileCard
        name="John Doe"
        email="john@example.com"
        avatar="https://example.com/avatar.jpg"
        joinDate="2024-01-15"
      />
    );

    expect(tree.toJSON()).toMatchSnapshot();
  });

  it('matches snapshot without avatar', () => {
    const tree = render(
      <ProfileCard
        name="Jane Doe"
        email="jane@example.com"
        joinDate="2024-06-01"
      />
    );

    expect(tree.toJSON()).toMatchSnapshot();
  });
});
```

## Gotchas

1. **`fireEvent.changeText` vs `fireEvent.change`** — React Native Testing Library uses `changeText` for TextInput, not `change` (which is the web API). Using `change` silently does nothing and the test passes with stale values.

2. **Detox requires a release build for iOS** — debug builds are too slow for Detox timing. Build with `detox build --configuration ios.sim.release`. CI must also use release builds.

3. **Jest timer mocks break Animated** — `jest.useFakeTimers()` interferes with React Native's Animated module. Use `jest.useFakeTimers({ legacyFakeTimers: true })` or mock specific timers only.

4. **AsyncStorage mock must be set up before imports** — if a module imports AsyncStorage at the top level, the mock must be configured in `jest.setup.ts`, not in the test file. Module-level imports execute before `beforeAll`.

5. **Platform.OS is 'ios' in Jest by default** — all tests run as iOS unless you mock Platform. To test Android-specific behavior: `jest.mock('react-native/Libraries/Utilities/Platform', () => ({ OS: 'android', select: (obj) => obj.android }))`.

6. **Detox element matching is strict** — `by.text('Submit')` won't match 'SUBMIT' or ' Submit '. Use `by.id()` with testID props for reliable element selection across platforms and locales.
