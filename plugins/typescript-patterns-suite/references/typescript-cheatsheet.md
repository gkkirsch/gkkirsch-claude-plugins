# TypeScript Cheat Sheet

## tsconfig.json — Strict Mode

```json
{
  "compilerOptions": {
    "strict": true,
    "noUncheckedIndexedAccess": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "exactOptionalPropertyTypes": true,
    "noImplicitReturns": true,
    "noFallthroughCasesInSwitch": true,
    "forceConsistentCasingInFileNames": true,
    "moduleResolution": "bundler",
    "module": "ESNext",
    "target": "ES2022",
    "lib": ["ES2022", "DOM", "DOM.Iterable"],
    "jsx": "react-jsx",
    "declaration": true,
    "sourceMap": true,
    "skipLibCheck": true
  }
}
```

## Type vs Interface

| Feature | `type` | `interface` |
|---------|--------|-------------|
| Object shapes | Yes | Yes |
| Extends/inheritance | `&` intersection | `extends` |
| Union types | Yes | No |
| Mapped types | Yes | No |
| Declaration merging | No | Yes |
| `implements` | Yes | Yes |
| Computed properties | Yes | No |
| Tuple types | Yes | No |

**Rule of thumb**: Use `interface` for objects that will be extended. Use `type` for unions, intersections, mapped types, and everything else.

## Essential Utility Types

| Type | What It Does | Example |
|------|-------------|---------|
| `Partial<T>` | All props optional | `Partial<User>` |
| `Required<T>` | All props required | `Required<Config>` |
| `Readonly<T>` | All props readonly | `Readonly<State>` |
| `Pick<T, K>` | Select specific props | `Pick<User, 'id' \| 'name'>` |
| `Omit<T, K>` | Remove specific props | `Omit<User, 'password'>` |
| `Record<K, V>` | Map keys to values | `Record<string, number>` |
| `Extract<T, U>` | Keep matching members | `Extract<'a' \| 'b', 'a'>` → `'a'` |
| `Exclude<T, U>` | Remove matching members | `Exclude<'a' \| 'b', 'a'>` → `'b'` |
| `NonNullable<T>` | Remove null/undefined | `NonNullable<string \| null>` → `string` |
| `ReturnType<T>` | Function return type | `ReturnType<typeof fn>` |
| `Parameters<T>` | Function param types | `Parameters<typeof fn>` |
| `Awaited<T>` | Unwrap Promise type | `Awaited<Promise<string>>` → `string` |
| `ConstructorParameters<T>` | Constructor params | `ConstructorParameters<typeof MyClass>` |
| `InstanceType<T>` | Instance type of class | `InstanceType<typeof MyClass>` |

## Generics Patterns

```typescript
// Constrained generic
function getProperty<T, K extends keyof T>(obj: T, key: K): T[K] {
  return obj[key];
}

// Default generic
function createArray<T = string>(length: number, value: T): T[] {
  return Array(length).fill(value);
}

// Generic with multiple constraints
function merge<T extends object, U extends object>(a: T, b: U): T & U {
  return { ...a, ...b };
}

// Generic factory
function createEntity<T extends { id: string }>(data: Omit<T, 'id'>): T {
  return { ...data, id: crypto.randomUUID() } as T;
}
```

## Conditional Types

```typescript
// Basic
type IsString<T> = T extends string ? true : false;

// With infer
type UnwrapPromise<T> = T extends Promise<infer U> ? U : T;
type ElementType<T> = T extends (infer U)[] ? U : T;
type ReturnOf<T> = T extends (...args: any[]) => infer R ? R : never;

// Distributive (over unions)
type ToArray<T> = T extends any ? T[] : never;
// ToArray<string | number> → string[] | number[]

// Non-distributive (wrap in tuple)
type ToArrayND<T> = [T] extends [any] ? T[] : never;
// ToArrayND<string | number> → (string | number)[]
```

## Mapped Types

```typescript
// Make all optional
type MyPartial<T> = { [K in keyof T]?: T[K] };

// Make all required
type MyRequired<T> = { [K in keyof T]-?: T[K] };

// Make all readonly
type MyReadonly<T> = { readonly [K in keyof T]: T[K] };

// Key remapping (as clause)
type Getters<T> = {
  [K in keyof T as `get${Capitalize<string & K>}`]: () => T[K];
};
// Getters<{ name: string }> → { getName: () => string }

// Filter keys by value type
type StringKeys<T> = {
  [K in keyof T as T[K] extends string ? K : never]: T[K];
};
```

## Discriminated Unions

```typescript
type Result<T> =
  | { status: 'success'; data: T }
  | { status: 'error'; error: Error }
  | { status: 'loading' };

function handle<T>(result: Result<T>) {
  switch (result.status) {
    case 'success': return result.data;     // T
    case 'error':   throw result.error;     // Error
    case 'loading': return null;
  }
  // Exhaustive check
  const _: never = result;
}
```

## Type Guards

```typescript
// typeof
function process(x: string | number) {
  if (typeof x === 'string') return x.toUpperCase(); // string
  return x.toFixed(2); // number
}

// in
function handle(shape: Circle | Square) {
  if ('radius' in shape) return Math.PI * shape.radius ** 2; // Circle
  return shape.side ** 2; // Square
}

// Custom predicate
function isUser(value: unknown): value is User {
  return typeof value === 'object' && value !== null && 'email' in value;
}

// Assertion function
function assertDefined<T>(value: T | undefined, msg?: string): asserts value is T {
  if (value === undefined) throw new Error(msg ?? 'Value is undefined');
}
```

## Branded Types

```typescript
type Brand<T, B extends string> = T & { readonly __brand: B };

type UserId = Brand<string, 'UserId'>;
type PostId = Brand<string, 'PostId'>;
type Email = Brand<string, 'Email'>;

function createUserId(id: string): UserId { return id as UserId; }
function createEmail(email: string): Email {
  if (!email.includes('@')) throw new Error('Invalid email');
  return email as Email;
}

function getUser(id: UserId): User { /* ... */ }
// getUser('raw-string')           // Error
// getUser(createUserId('abc-123')) // OK
// getUser(createPostId('abc-123'))// Error — PostId !== UserId
```

## Template Literal Types

```typescript
type EventName<T extends string> = `${T}Changed`;
// EventName<'name'> → 'nameChanged'

type CSSUnit = 'px' | 'rem' | 'em' | '%';
type CSSValue = `${number}${CSSUnit}`;
// '16px' OK, '1.5rem' OK, 'abc' Error

type HTTPMethod = 'GET' | 'POST' | 'PUT' | 'DELETE';
type APIRoute = `${HTTPMethod} /${string}`;
// 'GET /users' OK, 'PATCH /users' Error
```

## const Assertions + satisfies

```typescript
// const assertion — narrowest possible type
const config = {
  api: 'https://api.example.com',
  timeout: 5000,
  retries: 3,
} as const;
// type: { readonly api: 'https://api.example.com'; readonly timeout: 5000; ... }

// satisfies — validate shape but keep literal types
const routes = {
  home: '/',
  about: '/about',
  user: '/users/:id',
} satisfies Record<string, string>;
// routes.home is type '/', not string
// routes.typo would error (unknown key)
```

## Function Overloads

```typescript
// Overload signatures
function format(value: string): string;
function format(value: number): string;
function format(value: Date): string;
// Implementation
function format(value: string | number | Date): string {
  if (typeof value === 'string') return value.trim();
  if (typeof value === 'number') return value.toFixed(2);
  return value.toISOString();
}

// Conditional return type (alternative to overloads)
function parse<T extends 'string' | 'number'>(
  value: string,
  as: T,
): T extends 'string' ? string : number {
  if (as === 'string') return value as any;
  return Number(value) as any;
}
```

## Zod + TypeScript

```typescript
import { z } from 'zod';

const UserSchema = z.object({
  id: z.string().uuid(),
  email: z.string().email(),
  name: z.string().min(1).max(100),
  role: z.enum(['user', 'admin']),
  createdAt: z.coerce.date(),
});

type User = z.infer<typeof UserSchema>;
type CreateUser = z.input<typeof UserSchema>;

// Runtime validation
const user = UserSchema.parse(data);           // throws on failure
const result = UserSchema.safeParse(data);     // returns { success, data, error }

// Partial / Pick / Omit work on Zod schemas too
const UpdateSchema = UserSchema.partial().omit({ id: true, createdAt: true });
```

## Common Patterns

| Pattern | Code |
|---------|------|
| Exhaustive switch | `default: const _: never = value; throw new Error(\`Unhandled: ${_}\`)` |
| Type-safe object keys | `(Object.keys(obj) as Array<keyof typeof obj>)` |
| Type-safe entries | `(Object.entries(obj) as [keyof T, T[keyof T]][])` |
| Nullable check | `value != null` (checks both null and undefined) |
| Non-null assertion | `value!.property` (use sparingly, prefer type guards) |
| Optional chaining | `obj?.nested?.deep` |
| Nullish coalescing | `value ?? defaultValue` (only null/undefined, not 0 or '') |
| Index signature | `Record<string, unknown>` over `{ [key: string]: any }` |
| Enum alternative | `const Status = { Active: 'active', Inactive: 'inactive' } as const` |
| Extract array element | `type Item = ArrayType[number]` |
| Prettify intersections | `type Prettify<T> = { [K in keyof T]: T[K] } & {}` |
| Readonly array param | `function process(items: readonly string[])` |
