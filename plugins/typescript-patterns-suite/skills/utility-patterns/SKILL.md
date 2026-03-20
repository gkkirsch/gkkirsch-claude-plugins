---
name: utility-patterns
description: >
  TypeScript utility type patterns — built-in utilities, custom utility types,
  branded types, type guards, assertion functions, and pattern matching.
  Triggers: "utility type", "Pick Omit", "Partial Required", "branded type",
  "type guard", "assertion function", "type narrowing", "type predicate".
  NOT for: advanced generics (use advanced-types), project setup (use type-architect).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# TypeScript Utility Patterns

## Built-in Utility Types

```typescript
interface User {
  id: string;
  email: string;
  name: string;
  password: string;
  role: 'admin' | 'user';
  createdAt: Date;
}

// Pick — select specific properties
type UserPublic = Pick<User, 'id' | 'email' | 'name'>;

// Omit — exclude specific properties
type UserWithoutPassword = Omit<User, 'password'>;

// Partial — make all properties optional
type UserUpdate = Partial<User>;

// Required — make all properties required
type UserComplete = Required<User>;

// Readonly — make all properties readonly
type FrozenUser = Readonly<User>;

// Record — create object type with specific keys and value type
type UserRoles = Record<string, 'admin' | 'user' | 'guest'>;

// Extract — extract union members matching a type
type StringOrNumber = Extract<string | number | boolean, string | number>;
// string | number

// Exclude — remove union members matching a type
type OnlyStrings = Exclude<string | number | boolean, number | boolean>;
// string

// NonNullable — remove null and undefined
type Defined = NonNullable<string | null | undefined>;
// string

// ReturnType — get function return type
type FetchResult = ReturnType<typeof fetch>;
// Promise<Response>

// Parameters — get function parameter types as tuple
type FetchParams = Parameters<typeof fetch>;
// [input: RequestInfo | URL, init?: RequestInit]

// Awaited — unwrap Promise type
type ResolvedFetch = Awaited<ReturnType<typeof fetch>>;
// Response

// ConstructorParameters — get constructor parameter types
type DateParams = ConstructorParameters<typeof Date>;
```

## Custom Utility Types

```typescript
// Make specific properties optional
type PartialBy<T, K extends keyof T> = Omit<T, K> & Partial<Pick<T, K>>;
type CreateUser = PartialBy<User, 'id' | 'createdAt'>;
// { email: string; name: string; password: string; role: ... }
// & { id?: string; createdAt?: Date }

// Make specific properties required
type RequiredBy<T, K extends keyof T> = Omit<T, K> & Required<Pick<T, K>>;

// Make specific properties nullable
type NullableBy<T, K extends keyof T> = Omit<T, K> & {
  [P in K]: T[P] | null;
};

// Deep partial
type DeepPartial<T> = {
  [K in keyof T]?: T[K] extends object ? DeepPartial<T[K]> : T[K];
};

// Prettify (flatten intersection for better IntelliSense)
type Prettify<T> = { [K in keyof T]: T[K] } & {};

type UserCreateInput = Prettify<PartialBy<User, 'id' | 'createdAt'>>;
// Shows flattened type in hover, not Omit<...> & Partial<...>

// Value of object
type ValueOf<T> = T[keyof T];
type UserFieldType = ValueOf<User>;
// string | Date | 'admin' | 'user'

// Strict omit (errors if key doesn't exist)
type StrictOmit<T, K extends keyof T> = Pick<T, Exclude<keyof T, K>>;
// StrictOmit<User, 'foo'>;  // Error: 'foo' not in keyof User

// Rename key
type RenameKey<T, Old extends keyof T, New extends string> =
  Omit<T, Old> & { [K in New]: T[Old] };
type UserWithUsername = RenameKey<User, 'name', 'username'>;
```

## Branded Types

```typescript
// Prevent mixing values that share the same primitive type
declare const __brand: unique symbol;
type Brand<T, B extends string> = T & { readonly [__brand]: B };

type UserId = Brand<string, 'UserId'>;
type PostId = Brand<string, 'PostId'>;
type Email = Brand<string, 'Email'>;
type Cents = Brand<number, 'Cents'>;

// Constructor functions
function UserId(id: string): UserId { return id as UserId; }
function PostId(id: string): PostId { return id as PostId; }
function Email(email: string): Email {
  if (!email.includes('@')) throw new Error('Invalid email');
  return email as Email;
}
function Cents(amount: number): Cents {
  return Math.round(amount) as Cents;
}

// Usage — can't accidentally mix them
function getUser(id: UserId): User { /* ... */ }
function getPost(id: PostId): Post { /* ... */ }

const userId = UserId('user-123');
const postId = PostId('post-456');

getUser(userId);   // OK
// getUser(postId); // Error! PostId not assignable to UserId
// getUser('raw');  // Error! string not assignable to UserId

// Money operations
function addCents(a: Cents, b: Cents): Cents {
  return Cents(a + b);
}
```

## Type Guards

```typescript
// typeof guard
function processInput(input: string | number) {
  if (typeof input === 'string') {
    return input.toUpperCase(); // input: string
  }
  return input.toFixed(2);      // input: number
}

// in guard
interface Dog { bark(): void; breed: string; }
interface Cat { meow(): void; color: string; }

function speak(animal: Dog | Cat) {
  if ('bark' in animal) {
    animal.bark();  // animal: Dog
  } else {
    animal.meow();  // animal: Cat
  }
}

// instanceof guard
function formatError(error: unknown) {
  if (error instanceof Error) {
    return error.message;  // error: Error
  }
  return String(error);
}

// Custom type guard (type predicate)
function isString(value: unknown): value is string {
  return typeof value === 'string';
}

function isUser(value: unknown): value is User {
  return (
    typeof value === 'object' &&
    value !== null &&
    'id' in value &&
    'email' in value &&
    typeof (value as User).id === 'string' &&
    typeof (value as User).email === 'string'
  );
}

// Array type guard
function isStringArray(value: unknown): value is string[] {
  return Array.isArray(value) && value.every((item) => typeof item === 'string');
}

// Nullable guard
function isDefined<T>(value: T | null | undefined): value is T {
  return value !== null && value !== undefined;
}

// Filter with type guard
const mixed: (string | null)[] = ['a', null, 'b', null, 'c'];
const strings: string[] = mixed.filter(isDefined);
```

## Assertion Functions

```typescript
// Assert function (throws if false, narrows type after call)
function assertDefined<T>(
  value: T | null | undefined,
  message = 'Value is not defined',
): asserts value is T {
  if (value === null || value === undefined) {
    throw new Error(message);
  }
}

function assertUser(value: unknown): asserts value is User {
  if (!isUser(value)) {
    throw new Error('Invalid user data');
  }
}

// Usage
function processUser(data: unknown) {
  assertUser(data);
  // data is now typed as User — no if/else needed
  console.log(data.email);
}
```

## Const Assertions

```typescript
// as const makes literal types and readonly
const CONFIG = {
  api: 'https://api.example.com',
  timeout: 5000,
  retries: 3,
} as const;
// type: { readonly api: "https://api.example.com"; readonly timeout: 5000; readonly retries: 3 }

// Enum alternative
const STATUS = {
  Active: 'active',
  Inactive: 'inactive',
  Pending: 'pending',
} as const;

type Status = (typeof STATUS)[keyof typeof STATUS];
// 'active' | 'inactive' | 'pending'

// Tuple preservation
const ROUTES = ['/home', '/about', '/contact'] as const;
type Route = (typeof ROUTES)[number];
// '/home' | '/about' | '/contact'

// satisfies + as const (best of both worlds)
const THEME = {
  colors: {
    primary: '#3b82f6',
    secondary: '#64748b',
  },
  spacing: {
    sm: 4,
    md: 8,
    lg: 16,
  },
} as const satisfies Record<string, Record<string, string | number>>;
// Type-checks the shape AND preserves literal types
```

## Function Overloads

```typescript
// Overloaded function — different return types based on input
function createElement(tag: 'div'): HTMLDivElement;
function createElement(tag: 'span'): HTMLSpanElement;
function createElement(tag: 'input'): HTMLInputElement;
function createElement(tag: string): HTMLElement;
function createElement(tag: string): HTMLElement {
  return document.createElement(tag);
}

const div = createElement('div');   // HTMLDivElement
const span = createElement('span'); // HTMLSpanElement
const other = createElement('p');   // HTMLElement

// Conditional return types (alternative to overloads)
function fetchData<T extends 'user' | 'post'>(
  type: T,
): T extends 'user' ? User : Post {
  // implementation
}
```

## Exhaustive Pattern Matching

```typescript
// Ensure all union members are handled
type Shape =
  | { kind: 'circle'; radius: number }
  | { kind: 'rectangle'; width: number; height: number }
  | { kind: 'triangle'; base: number; height: number };

function area(shape: Shape): number {
  switch (shape.kind) {
    case 'circle':
      return Math.PI * shape.radius ** 2;
    case 'rectangle':
      return shape.width * shape.height;
    case 'triangle':
      return (shape.base * shape.height) / 2;
    default: {
      // This ensures all cases are handled at compile time
      const _exhaustive: never = shape;
      throw new Error(`Unhandled shape: ${(_exhaustive as Shape).kind}`);
    }
  }
}

// Adding a new shape kind will cause a compile error at the default case
// until you handle it. This is the whole point.
```

## Gotchas

1. **`Omit` doesn't error on invalid keys.** `Omit<User, 'foo'>` silently passes — `'foo'` is just ignored. Use `StrictOmit` pattern above for safety.

2. **`satisfies` doesn't narrow.** `satisfies` validates the type but preserves the original inferred type. It doesn't change the variable's type. Use it for validation, not narrowing.

3. **Type guards don't validate at runtime.** `value is User` is a compile-time assertion. The runtime check in the guard function is YOUR responsibility. Get it wrong and TypeScript trusts your lie.

4. **Branded types are phantom types.** The brand property doesn't exist at runtime. It's a compile-time-only tag. `JSON.stringify` won't include it.

5. **`as const` makes everything readonly.** If you need to mutate the object later, you can't use `as const`. Use `satisfies` instead for type-checking without readonly.

6. **`NonNullable<T>` vs `T & {}`.** They do the same thing. `NonNullable` is more readable. `T & {}` is a common shorthand in utility type implementations.
