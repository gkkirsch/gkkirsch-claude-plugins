---
name: react-hooks-advanced
description: >
  Advanced React hooks — custom hook patterns, useReducer, useRef tricks,
  useSyncExternalStore, useTransition, useDeferredValue, and hook composition.
  Triggers: "custom hook", "useReducer", "useRef", "useTransition",
  "useDeferredValue", "advanced hooks", "hook pattern".
  NOT for: basic useState/useEffect (too simple), form hooks (use react-hook-form skill).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Advanced React Hooks

## Custom Hook Patterns

### Data Fetching Hook

```tsx
function useApi<T>(url: string | null, options?: RequestInit) {
  const [data, setData] = useState<T | null>(null);
  const [error, setError] = useState<Error | null>(null);
  const [isLoading, setIsLoading] = useState(false);

  useEffect(() => {
    if (!url) return;

    const controller = new AbortController();
    setIsLoading(true);

    fetch(url, { ...options, signal: controller.signal })
      .then((res) => {
        if (!res.ok) throw new Error(`HTTP ${res.status}`);
        return res.json();
      })
      .then(setData)
      .catch((err) => {
        if (err.name !== 'AbortError') setError(err);
      })
      .finally(() => setIsLoading(false));

    return () => controller.abort();
  }, [url]);

  return { data, error, isLoading };
}

// Usage:
const { data: users, isLoading } = useApi<User[]>('/api/users');
const { data: user } = useApi<User>(userId ? `/api/users/${userId}` : null);
```

### Debounce Hook

```tsx
function useDebounce<T>(value: T, delay: number): T {
  const [debouncedValue, setDebouncedValue] = useState(value);

  useEffect(() => {
    const timer = setTimeout(() => setDebouncedValue(value), delay);
    return () => clearTimeout(timer);
  }, [value, delay]);

  return debouncedValue;
}

// Usage:
function SearchInput() {
  const [query, setQuery] = useState('');
  const debouncedQuery = useDebounce(query, 300);

  const { data } = useQuery({
    queryKey: ['search', debouncedQuery],
    queryFn: () => searchApi(debouncedQuery),
    enabled: debouncedQuery.length > 0,
  });

  return <input value={query} onChange={(e) => setQuery(e.target.value)} />;
}
```

### Local Storage Hook

```tsx
function useLocalStorage<T>(key: string, initialValue: T) {
  const [storedValue, setStoredValue] = useState<T>(() => {
    try {
      const item = localStorage.getItem(key);
      return item ? JSON.parse(item) : initialValue;
    } catch {
      return initialValue;
    }
  });

  const setValue = useCallback((value: T | ((prev: T) => T)) => {
    setStoredValue((prev) => {
      const next = value instanceof Function ? value(prev) : value;
      localStorage.setItem(key, JSON.stringify(next));
      return next;
    });
  }, [key]);

  const removeValue = useCallback(() => {
    localStorage.removeItem(key);
    setStoredValue(initialValue);
  }, [key, initialValue]);

  return [storedValue, setValue, removeValue] as const;
}

// Usage:
const [theme, setTheme] = useLocalStorage('theme', 'light');
const [favorites, setFavorites] = useLocalStorage<string[]>('favorites', []);
```

### Media Query Hook

```tsx
function useMediaQuery(query: string): boolean {
  const [matches, setMatches] = useState(() =>
    typeof window !== 'undefined' ? window.matchMedia(query).matches : false
  );

  useEffect(() => {
    const mediaQuery = window.matchMedia(query);
    const handler = (e: MediaQueryListEvent) => setMatches(e.matches);
    mediaQuery.addEventListener('change', handler);
    setMatches(mediaQuery.matches);
    return () => mediaQuery.removeEventListener('change', handler);
  }, [query]);

  return matches;
}

// Usage:
const isMobile = useMediaQuery('(max-width: 768px)');
const prefersReducedMotion = useMediaQuery('(prefers-reduced-motion: reduce)');
const isDarkMode = useMediaQuery('(prefers-color-scheme: dark)');
```

### Click Outside Hook

```tsx
function useClickOutside<T extends HTMLElement>(handler: () => void) {
  const ref = useRef<T>(null);

  useEffect(() => {
    const listener = (event: MouseEvent | TouchEvent) => {
      if (!ref.current || ref.current.contains(event.target as Node)) return;
      handler();
    };

    document.addEventListener('mousedown', listener);
    document.addEventListener('touchstart', listener);
    return () => {
      document.removeEventListener('mousedown', listener);
      document.removeEventListener('touchstart', listener);
    };
  }, [handler]);

  return ref;
}

// Usage:
function Dropdown() {
  const [open, setOpen] = useState(false);
  const ref = useClickOutside<HTMLDivElement>(() => setOpen(false));

  return (
    <div ref={ref}>
      <button onClick={() => setOpen(!open)}>Menu</button>
      {open && <DropdownContent />}
    </div>
  );
}
```

### Intersection Observer Hook

```tsx
function useIntersectionObserver(options?: IntersectionObserverInit) {
  const ref = useRef<HTMLElement>(null);
  const [isIntersecting, setIsIntersecting] = useState(false);

  useEffect(() => {
    const element = ref.current;
    if (!element) return;

    const observer = new IntersectionObserver(
      ([entry]) => setIsIntersecting(entry.isIntersecting),
      options
    );

    observer.observe(element);
    return () => observer.disconnect();
  }, [options?.threshold, options?.root, options?.rootMargin]);

  return [ref, isIntersecting] as const;
}

// Usage — lazy load images:
function LazyImage({ src, alt }: { src: string; alt: string }) {
  const [ref, isVisible] = useIntersectionObserver({ threshold: 0.1 });
  return (
    <div ref={ref}>
      {isVisible ? <img src={src} alt={alt} /> : <Skeleton className="w-full h-48" />}
    </div>
  );
}
```

## useReducer

```tsx
// State machine for async operations
type FetchState<T> =
  | { status: 'idle' }
  | { status: 'loading' }
  | { status: 'success'; data: T }
  | { status: 'error'; error: Error };

type FetchAction<T> =
  | { type: 'fetch' }
  | { type: 'success'; data: T }
  | { type: 'error'; error: Error }
  | { type: 'reset' };

function fetchReducer<T>(state: FetchState<T>, action: FetchAction<T>): FetchState<T> {
  switch (action.type) {
    case 'fetch': return { status: 'loading' };
    case 'success': return { status: 'success', data: action.data };
    case 'error': return { status: 'error', error: action.error };
    case 'reset': return { status: 'idle' };
  }
}

// Usage:
function UserProfile({ userId }: { userId: string }) {
  const [state, dispatch] = useReducer(fetchReducer<User>, { status: 'idle' });

  useEffect(() => {
    dispatch({ type: 'fetch' });
    fetchUser(userId)
      .then((data) => dispatch({ type: 'success', data }))
      .catch((error) => dispatch({ type: 'error', error }));
  }, [userId]);

  switch (state.status) {
    case 'idle': return null;
    case 'loading': return <Spinner />;
    case 'error': return <ErrorMessage error={state.error} />;
    case 'success': return <ProfileCard user={state.data} />;
  }
}
```

## useRef Advanced

```tsx
// Previous value
function usePrevious<T>(value: T): T | undefined {
  const ref = useRef<T>();
  useEffect(() => { ref.current = value; });
  return ref.current;
}

// Stable callback (latest ref pattern)
function useStableCallback<T extends (...args: any[]) => any>(callback: T): T {
  const ref = useRef(callback);
  useLayoutEffect(() => { ref.current = callback; });
  return useCallback((...args: Parameters<T>) => ref.current(...args), []) as T;
}

// Interval with cleanup
function useInterval(callback: () => void, delay: number | null) {
  const savedCallback = useRef(callback);
  useEffect(() => { savedCallback.current = callback; });
  useEffect(() => {
    if (delay === null) return;
    const id = setInterval(() => savedCallback.current(), delay);
    return () => clearInterval(id);
  }, [delay]);
}

// Measure element dimensions
function useMeasure<T extends HTMLElement>() {
  const ref = useRef<T>(null);
  const [bounds, setBounds] = useState({ width: 0, height: 0, top: 0, left: 0 });

  useEffect(() => {
    if (!ref.current) return;
    const observer = new ResizeObserver(([entry]) => {
      setBounds(entry.contentRect);
    });
    observer.observe(ref.current);
    return () => observer.disconnect();
  }, []);

  return [ref, bounds] as const;
}
```

## useTransition & useDeferredValue

```tsx
// useTransition — mark updates as non-urgent
function FilterableList({ items }: { items: Item[] }) {
  const [query, setQuery] = useState('');
  const [isPending, startTransition] = useTransition();

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    // Urgent: update the input immediately
    setQuery(e.target.value);

    // Non-urgent: filtering can wait
    startTransition(() => {
      setFilteredItems(items.filter(item =>
        item.name.toLowerCase().includes(e.target.value.toLowerCase())
      ));
    });
  };

  return (
    <>
      <input value={query} onChange={handleChange} />
      {isPending && <Spinner />}
      <ItemList items={filteredItems} />
    </>
  );
}

// useDeferredValue — defer an expensive render
function SearchResults({ query }: { query: string }) {
  const deferredQuery = useDeferredValue(query);
  const isStale = query !== deferredQuery;

  const results = useMemo(
    () => heavySearch(deferredQuery),
    [deferredQuery]
  );

  return (
    <div style={{ opacity: isStale ? 0.7 : 1 }}>
      {results.map(r => <ResultItem key={r.id} result={r} />)}
    </div>
  );
}
```

## Hook Composition

```tsx
// Compose multiple hooks into a feature hook
function useAuthenticatedUser() {
  const { data: session } = useQuery({ queryKey: ['session'], queryFn: getSession });
  const { data: user } = useQuery({
    queryKey: ['user', session?.userId],
    queryFn: () => fetchUser(session!.userId),
    enabled: !!session?.userId,
  });

  const logout = useMutation({
    mutationFn: logoutApi,
    onSuccess: () => queryClient.clear(),
  });

  return {
    user,
    isAuthenticated: !!session,
    isLoading: !session,
    logout: logout.mutate,
    isLoggingOut: logout.isPending,
  };
}

// Usage:
function Header() {
  const { user, logout, isLoggingOut } = useAuthenticatedUser();
  // Clean, declarative, all auth logic encapsulated
}
```

## Gotchas

1. **Hooks must be called in the same order.** No hooks inside if/else, loops, or after early returns. This is the #1 hook rule.

2. **useEffect cleanup runs before every re-execution.** Not just on unmount. If your effect depends on `userId`, cleanup runs when `userId` changes, then the effect runs again.

3. **useRef doesn't trigger re-renders.** Changing `ref.current` does not cause the component to re-render. Use state if you need the UI to update.

4. **Empty dependency array means "run once."** `useEffect(() => {...}, [])` runs on mount only. But if the effect uses variables from the component scope, they'll be stale.

5. **Custom hooks should start with `use`.** This isn't just convention — the React linter uses it to enforce the rules of hooks.

6. **useCallback doesn't memoize the result.** It memoizes the function reference. The function still runs every time it's called. For memoizing return values, use `useMemo`.
