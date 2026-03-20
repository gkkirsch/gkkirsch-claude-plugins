---
name: type-safe-patterns
description: >
  Type-safe patterns for real-world TypeScript — API clients, event systems,
  form handling, state machines, environment variables, and configuration.
  Triggers: "type safe api", "typed fetch", "type safe event", "type safe config",
  "type safe env", "type safe form", "type safe state machine".
  NOT for: abstract type theory (use advanced-types), utility types (use utility-patterns).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Type-Safe Patterns

## Type-Safe API Client

```typescript
// Define API routes and their request/response types
interface ApiRoutes {
  'GET /users': { response: User[]; query: { page?: number; limit?: number } };
  'GET /users/:id': { response: User; params: { id: string } };
  'POST /users': { response: User; body: CreateUserInput };
  'PUT /users/:id': { response: User; params: { id: string }; body: UpdateUserInput };
  'DELETE /users/:id': { response: void; params: { id: string } };
  'GET /posts': { response: Post[]; query: { authorId?: string } };
  'POST /posts': { response: Post; body: CreatePostInput };
}

// Extract parts
type Method = 'GET' | 'POST' | 'PUT' | 'DELETE';
type RouteKey = keyof ApiRoutes;

type RouteConfig<K extends RouteKey> = ApiRoutes[K];

// Type-safe fetch wrapper
async function api<K extends RouteKey>(
  route: K,
  options?: {
    params?: RouteConfig<K> extends { params: infer P } ? P : never;
    body?: RouteConfig<K> extends { body: infer B } ? B : never;
    query?: RouteConfig<K> extends { query: infer Q } ? Q : never;
  },
): Promise<RouteConfig<K>['response']> {
  const [method, path] = (route as string).split(' ') as [Method, string];

  let url = path;
  if (options?.params) {
    for (const [key, value] of Object.entries(options.params)) {
      url = url.replace(`:${key}`, String(value));
    }
  }

  if (options?.query) {
    const searchParams = new URLSearchParams(
      Object.entries(options.query).filter(([, v]) => v !== undefined) as [string, string][],
    );
    url += `?${searchParams}`;
  }

  const response = await fetch(`/api${url}`, {
    method,
    headers: { 'Content-Type': 'application/json' },
    body: options?.body ? JSON.stringify(options.body) : undefined,
  });

  if (!response.ok) throw new Error(`API error: ${response.status}`);
  if (response.status === 204) return undefined as any;
  return response.json();
}

// Usage — fully typed
const users = await api('GET /users', { query: { page: 1 } });        // User[]
const user = await api('GET /users/:id', { params: { id: '123' } });  // User
const newUser = await api('POST /users', { body: { email: '...', name: '...' } }); // User
```

## Type-Safe Event Emitter

```typescript
type EventMap = {
  'user:login': { userId: string; timestamp: Date };
  'user:logout': { userId: string };
  'post:created': { postId: string; authorId: string };
  'post:deleted': { postId: string };
  'error': { message: string; code: number };
};

class TypedEventEmitter<T extends Record<string, unknown>> {
  private listeners = new Map<keyof T, Set<(data: any) => void>>();

  on<K extends keyof T>(event: K, handler: (data: T[K]) => void): () => void {
    if (!this.listeners.has(event)) {
      this.listeners.set(event, new Set());
    }
    this.listeners.get(event)!.add(handler);

    // Return unsubscribe function
    return () => { this.listeners.get(event)?.delete(handler); };
  }

  emit<K extends keyof T>(event: K, data: T[K]): void {
    this.listeners.get(event)?.forEach((handler) => handler(data));
  }

  once<K extends keyof T>(event: K, handler: (data: T[K]) => void): void {
    const unsubscribe = this.on(event, (data) => {
      unsubscribe();
      handler(data);
    });
  }
}

const events = new TypedEventEmitter<EventMap>();

events.on('user:login', (data) => {
  console.log(data.userId);    // Typed!
  console.log(data.timestamp); // Typed!
});

events.emit('user:login', { userId: '123', timestamp: new Date() }); // OK
// events.emit('user:login', { userId: 123 }); // Error: number not string
```

## Type-Safe Environment Variables

```typescript
// src/config/env.ts
import { z } from 'zod';

const envSchema = z.object({
  NODE_ENV: z.enum(['development', 'production', 'test']),
  PORT: z.coerce.number().default(3000),
  DATABASE_URL: z.string().url(),
  JWT_SECRET: z.string().min(32),
  REDIS_URL: z.string().url().optional(),
  STRIPE_SECRET_KEY: z.string().startsWith('sk_'),
  STRIPE_WEBHOOK_SECRET: z.string().startsWith('whsec_'),
  S3_BUCKET: z.string(),
  S3_REGION: z.string().default('us-east-1'),
});

// Validate at startup — crash early if misconfigured
const parsed = envSchema.safeParse(process.env);

if (!parsed.success) {
  console.error('Invalid environment variables:');
  console.error(parsed.error.flatten().fieldErrors);
  process.exit(1);
}

export const env = parsed.data;

// Now import env anywhere — fully typed, validated at startup
// import { env } from './config/env';
// env.PORT          // number
// env.JWT_SECRET    // string (guaranteed 32+ chars)
// env.REDIS_URL     // string | undefined
// env.STRIPE_SECRET_KEY // string (guaranteed starts with "sk_")
```

## Type-Safe State Machine

```typescript
// Define states and transitions
type State = 'idle' | 'loading' | 'success' | 'error';

type StateData = {
  idle: undefined;
  loading: { startedAt: number };
  success: { data: unknown; loadedAt: number };
  error: { error: Error; failedAt: number };
};

type Transitions = {
  idle: 'loading';
  loading: 'success' | 'error';
  success: 'idle' | 'loading';
  error: 'idle' | 'loading';
};

type MachineState<S extends State> = {
  state: S;
  data: StateData[S];
};

class StateMachine<S extends State = 'idle'> {
  private currentState: S;
  private currentData: StateData[S];

  constructor(initial: S, data: StateData[S]) {
    this.currentState = initial;
    this.currentData = data;
  }

  transition<Next extends Transitions[S]>(
    nextState: Next,
    data: StateData[Next & State],
  ): StateMachine<Next & State> {
    (this as any).currentState = nextState;
    (this as any).currentData = data;
    return this as any;
  }

  getState(): MachineState<S> {
    return { state: this.currentState, data: this.currentData };
  }
}

const machine = new StateMachine('idle', undefined);
machine.transition('loading', { startedAt: Date.now() });
// machine.transition('success', ...); // Only valid from 'loading'
```

## Type-Safe Form Builder

```typescript
import { z } from 'zod';

// Schema-first form
const loginSchema = z.object({
  email: z.string().email('Invalid email'),
  password: z.string().min(8, 'Password must be at least 8 characters'),
  rememberMe: z.boolean().default(false),
});

type LoginForm = z.infer<typeof loginSchema>;

// Type-safe form state
interface FormState<T extends z.ZodObject<any>> {
  values: z.infer<T>;
  errors: Partial<Record<keyof z.infer<T>, string>>;
  touched: Partial<Record<keyof z.infer<T>, boolean>>;
  isValid: boolean;
  isSubmitting: boolean;
}

function createForm<T extends z.ZodObject<any>>(
  schema: T,
  defaults: z.infer<T>,
): FormState<T> & {
  setValue: <K extends keyof z.infer<T>>(field: K, value: z.infer<T>[K]) => void;
  validate: () => boolean;
} {
  const state: FormState<T> = {
    values: defaults,
    errors: {},
    touched: {},
    isValid: false,
    isSubmitting: false,
  };

  return {
    ...state,
    setValue(field, value) {
      state.values[field] = value;
      state.touched[field] = true;
    },
    validate() {
      const result = schema.safeParse(state.values);
      if (result.success) {
        state.errors = {};
        state.isValid = true;
      } else {
        const fieldErrors: any = {};
        for (const issue of result.error.issues) {
          const path = issue.path[0] as string;
          fieldErrors[path] = issue.message;
        }
        state.errors = fieldErrors;
        state.isValid = false;
      }
      return state.isValid;
    },
  };
}
```

## Type-Safe Route Params (React Router)

```typescript
// Define route params for the entire app
interface RouteParams {
  '/': never;
  '/users': never;
  '/users/:userId': { userId: string };
  '/users/:userId/posts': { userId: string };
  '/users/:userId/posts/:postId': { userId: string; postId: string };
  '/settings': never;
}

// Type-safe useParams hook
function useTypedParams<T extends keyof RouteParams>(): RouteParams[T] {
  return useParams() as RouteParams[T];
}

// Usage in component
function UserPostPage() {
  const { userId, postId } = useTypedParams<'/users/:userId/posts/:postId'>();
  // Both typed as string
}

// Type-safe navigation
function useTypedNavigate() {
  const navigate = useNavigate();
  return <T extends keyof RouteParams>(
    path: T,
    ...args: RouteParams[T] extends never ? [] : [params: RouteParams[T]]
  ) => {
    let url: string = path;
    if (args[0]) {
      for (const [key, value] of Object.entries(args[0])) {
        url = url.replace(`:${key}`, String(value));
      }
    }
    navigate(url);
  };
}

const nav = useTypedNavigate();
nav('/');                                                    // OK
nav('/users/:userId', { userId: '123' });                   // OK
// nav('/users/:userId');                                    // Error: missing params
// nav('/users/:userId', { userId: '123', foo: 'bar' });    // Error: extra params
```

## Type-Safe Local Storage

```typescript
interface StorageSchema {
  theme: 'light' | 'dark';
  language: 'en' | 'es' | 'fr';
  user: { id: string; name: string } | null;
  onboardingComplete: boolean;
  recentSearches: string[];
}

class TypedStorage<T extends Record<string, unknown>> {
  get<K extends keyof T>(key: K): T[K] | null {
    const raw = localStorage.getItem(key as string);
    if (raw === null) return null;
    try { return JSON.parse(raw); }
    catch { return null; }
  }

  set<K extends keyof T>(key: K, value: T[K]): void {
    localStorage.setItem(key as string, JSON.stringify(value));
  }

  remove<K extends keyof T>(key: K): void {
    localStorage.removeItem(key as string);
  }

  has<K extends keyof T>(key: K): boolean {
    return localStorage.getItem(key as string) !== null;
  }
}

const storage = new TypedStorage<StorageSchema>();

storage.set('theme', 'dark');        // OK
storage.set('user', { id: '1', name: 'Alice' }); // OK
// storage.set('theme', 'blue');     // Error: 'blue' not in 'light' | 'dark'

const theme = storage.get('theme');  // 'light' | 'dark' | null
```

## Gotchas

1. **Type-safe API clients add compile-time safety only.** The runtime response still needs validation (Zod `parse`) for true safety. Types don't validate network responses.

2. **`as any` in generic implementations is sometimes necessary.** TypeScript can't always prove that generic transformations are sound. Use `as any` sparingly inside implementation, never at call sites.

3. **State machine transitions lose type info without builder pattern.** The `transition()` method returns a new type, but reassigning to the same variable erases it. Chain transitions or use separate variables.

4. **Template literal route types have limits.** Complex path patterns (`/:param?`, catch-all `/*`) are hard to express with template literals. Consider a simpler approach for complex routing.

5. **`satisfies` is your friend for config objects.** `{ api: '...' } satisfies Config` validates the shape but preserves literal types. Without it, you get `string` instead of `'https://...'`.

6. **Generic function inference flows left to right.** Put the type you want inferred FIRST in the parameter list. `fn<T>(x: T, y: ...(T))` infers T from x and applies to y.
