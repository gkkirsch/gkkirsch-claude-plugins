# Advanced Types Reference

Quick reference for advanced TypeScript type patterns. For in-depth explanations, consult the `typescript-architect` agent.

---

## Built-in Utility Types

### Object Manipulation

```typescript
// Partial<T> — make all properties optional
type Partial<T> = { [K in keyof T]?: T[K] };
Partial<{ name: string; age: number }>
// { name?: string; age?: number }

// Required<T> — make all properties required
type Required<T> = { [K in keyof T]-?: T[K] };
Required<{ name?: string; age?: number }>
// { name: string; age: number }

// Readonly<T> — make all properties readonly
type Readonly<T> = { readonly [K in keyof T]: T[K] };

// Pick<T, K> — select properties
Pick<User, "name" | "email">
// { name: string; email: string }

// Omit<T, K> — remove properties
Omit<User, "password" | "salt">
// { id: string; name: string; email: string }

// Record<K, V> — create object type with key type K and value type V
Record<string, number>
// { [key: string]: number }

Record<"admin" | "user" | "guest", Permission[]>
// { admin: Permission[]; user: Permission[]; guest: Permission[] }
```

### Union Manipulation

```typescript
// Exclude<T, U> — remove members from union
Exclude<"a" | "b" | "c", "a" | "c">  // "b"

// Extract<T, U> — keep only matching members
Extract<string | number | boolean, string | number>  // string | number

// NonNullable<T> — remove null and undefined
NonNullable<string | null | undefined>  // string
```

### Function Types

```typescript
// Parameters<T> — extract parameter tuple
Parameters<(a: string, b: number) => void>  // [a: string, b: number]

// ReturnType<T> — extract return type
ReturnType<() => Promise<User>>  // Promise<User>

// ConstructorParameters<T> — extract constructor params
ConstructorParameters<typeof Map>  // [entries?: readonly (readonly [any, any])[] | null]

// InstanceType<T> — extract instance type from constructor
InstanceType<typeof Date>  // Date

// ThisParameterType<T> — extract this type
// OmitThisParameter<T> — remove this parameter
```

### String Types

```typescript
Uppercase<"hello">      // "HELLO"
Lowercase<"HELLO">      // "hello"
Capitalize<"hello">     // "Hello"
Uncapitalize<"Hello">   // "hello"
```

### Awaited

```typescript
// Awaited<T> — unwrap Promise (recursive)
Awaited<Promise<string>>                  // string
Awaited<Promise<Promise<number>>>         // number
Awaited<string | Promise<string>>         // string
```

### NoInfer (5.4+)

```typescript
// NoInfer<T> — prevent inference from this position
function foo<T>(value: T, defaultValue: NoInfer<T>): T {
  return value ?? defaultValue;
}
foo("hello", "default");  // T = string, not "hello" | "default"
```

---

## Recursive Types

### DeepPartial / DeepReadonly / DeepRequired

```typescript
type DeepPartial<T> = T extends object
  ? { [K in keyof T]?: DeepPartial<T[K]> }
  : T;

type DeepReadonly<T> = T extends (infer U)[]
  ? ReadonlyArray<DeepReadonly<U>>
  : T extends object
  ? { readonly [K in keyof T]: DeepReadonly<T[K]> }
  : T;

type DeepRequired<T> = T extends object
  ? { [K in keyof T]-?: DeepRequired<T[K]> }
  : T;

type DeepNullable<T> = T extends object
  ? { [K in keyof T]: DeepNullable<T[K]> | null }
  : T | null;

type DeepNonNullable<T> = T extends object
  ? { [K in keyof T]: DeepNonNullable<NonNullable<T[K]>> }
  : NonNullable<T>;
```

### Key Paths

```typescript
// Get all dot-separated key paths
type KeyPaths<T, Prefix extends string = ""> = T extends object
  ? {
      [K in keyof T & string]:
        | `${Prefix}${K}`
        | KeyPaths<T[K], `${Prefix}${K}.`>;
    }[keyof T & string]
  : never;

// Type-safe deep get
type DeepGet<T, Path extends string> =
  Path extends `${infer Key}.${infer Rest}`
    ? Key extends keyof T
      ? DeepGet<T[Key], Rest>
      : never
    : Path extends keyof T
    ? T[Path]
    : never;

// Usage
interface Config {
  db: { host: string; port: number; auth: { user: string; pass: string } };
}
type Paths = KeyPaths<Config>;
// "db" | "db.host" | "db.port" | "db.auth" | "db.auth.user" | "db.auth.pass"
type DBHost = DeepGet<Config, "db.host">; // string
```

### Flatten

```typescript
// Flatten nested arrays
type Flatten<T> = T extends Array<infer U> ? Flatten<U> : T;
type F = Flatten<number[][][]>; // number

// Flatten one level
type FlattenOnce<T> = T extends Array<infer U> ? U : T;
type F2 = FlattenOnce<number[][]>; // number[]
```

---

## Variadic Tuples

```typescript
// Concat tuples
type Concat<A extends any[], B extends any[]> = [...A, ...B];
type C = Concat<[1, 2], [3, 4]>; // [1, 2, 3, 4]

// Head, Tail, Last, Init
type Head<T extends any[]> = T extends [infer H, ...any[]] ? H : never;
type Tail<T extends any[]> = T extends [any, ...infer R] ? R : never;
type Last<T extends any[]> = T extends [...any[], infer L] ? L : never;
type Init<T extends any[]> = T extends [...infer I, any] ? I : never;

// Length
type Length<T extends any[]> = T["length"];

// Push / Unshift
type Push<T extends any[], V> = [...T, V];
type Unshift<T extends any[], V> = [V, ...T];

// Reverse
type Reverse<T extends any[]> = T extends [infer H, ...infer R]
  ? [...Reverse<R>, H]
  : [];

// Zip
type Zip<A extends any[], B extends any[]> =
  A extends [infer AH, ...infer AR]
    ? B extends [infer BH, ...infer BR]
      ? [[AH, BH], ...Zip<AR, BR>]
      : []
    : [];

type Z = Zip<[1, 2, 3], ["a", "b", "c"]>;
// [[1, "a"], [2, "b"], [3, "c"]]
```

---

## Const Assertions

```typescript
// as const — narrows to literal types and makes readonly
const colors = ["red", "green", "blue"] as const;
// readonly ["red", "green", "blue"]

type Color = (typeof colors)[number];
// "red" | "green" | "blue"

// Object as const
const config = {
  api: "https://api.example.com",
  port: 3000,
  features: { dark_mode: true, beta: false },
} as const;

type Config = typeof config;
// {
//   readonly api: "https://api.example.com";
//   readonly port: 3000;
//   readonly features: { readonly dark_mode: true; readonly beta: false };
// }

// Enum alternative with as const
const Status = {
  Active: "active",
  Inactive: "inactive",
  Pending: "pending",
} as const;
type Status = (typeof Status)[keyof typeof Status];
// "active" | "inactive" | "pending"

// const type parameter (5.0+)
function define<const T>(value: T): T { return value; }
const v = define({ x: 1, y: 2 });
// { readonly x: 1; readonly y: 2 }
```

---

## satisfies Operator

```typescript
// Validates against a type WITHOUT widening
const palette = {
  red: { r: 255, g: 0, b: 0 },
  green: "#00ff00",
} satisfies Record<string, string | { r: number; g: number; b: number }>;

palette.red.r;             // number (not string | { r: number; ... })
palette.green.toUpperCase(); // string (not string | { r: number; ... })

// Combine with as const
const routes = {
  home: "/",
  about: "/about",
} as const satisfies Record<string, string>;

routes.home; // "/" (literal)
// routes.missing; // Error: property doesn't exist
```

---

## Branded / Opaque Types

```typescript
// Brand utility
declare const __brand: unique symbol;
type Brand<T, B extends string> = T & { readonly [__brand]: B };

// Domain types
type UserId = Brand<string, "UserId">;
type Email = Brand<string, "Email">;
type Positive = Brand<number, "Positive">;
type Percentage = Brand<number, "Percentage">;

// Smart constructors
function userId(raw: string): UserId {
  if (!raw.match(/^usr_/)) throw new Error("Invalid user ID");
  return raw as UserId;
}

function email(raw: string): Email {
  if (!raw.includes("@")) throw new Error("Invalid email");
  return raw as Email;
}

function positive(n: number): Positive {
  if (n <= 0) throw new RangeError("Must be positive");
  return n as Positive;
}
```

---

## Discriminated Unions

```typescript
// Tagged union with exhaustive handling
type Result<T, E = Error> =
  | { ok: true; value: T }
  | { ok: false; error: E };

type AsyncState<T> =
  | { status: "idle" }
  | { status: "loading" }
  | { status: "success"; data: T }
  | { status: "error"; error: Error };

// Exhaustive check utility
function assertNever(x: never): never {
  throw new Error(`Unexpected: ${JSON.stringify(x)}`);
}

// Pattern: state machine
type State =
  | { state: "idle" }
  | { state: "fetching"; url: string }
  | { state: "done"; data: unknown }
  | { state: "failed"; error: Error };

function transition(current: State, event: Event): State {
  switch (current.state) {
    case "idle":
      // Only handle valid transitions from idle
      return current;
    case "fetching":
      return current;
    case "done":
      return current;
    case "failed":
      return current;
    default:
      return assertNever(current);
  }
}
```

---

## Template Literal Types

```typescript
// String manipulation
type Snake = CamelToSnake<"getUserById">;  // Custom: "get_user_by_id"
type Camel = SnakeToCamel<"get_user_by_id">; // Custom: "getUserById"

// Parse URL params
type Params<T extends string> =
  T extends `${string}:${infer P}/${infer R}`
    ? P | Params<R>
    : T extends `${string}:${infer P}`
    ? P
    : never;

type P = Params<"/users/:userId/posts/:postId">;
// "userId" | "postId"

// Event patterns
type EventName<T extends string> = `on${Capitalize<T>}`;
type Changed<T extends string> = `${T}Changed`;

// Key transformation in mapped types
type Getters<T> = {
  [K in keyof T as `get${Capitalize<string & K>}`]: () => T[K];
};

type Setters<T> = {
  [K in keyof T as `set${Capitalize<string & K>}`]: (v: T[K]) => void;
};
```

---

## Conditional Types Cheat Sheet

```typescript
// Basic conditional
type IsString<T> = T extends string ? true : false;

// Distributive (over unions)
type ToArray<T> = T extends any ? T[] : never;
ToArray<string | number>  // string[] | number[]

// Non-distributive (wrap in tuple)
type ToArrayND<T> = [T] extends [any] ? T[] : never;
ToArrayND<string | number>  // (string | number)[]

// infer — extract types
type UnpackPromise<T> = T extends Promise<infer U> ? U : T;
type ElementOf<T> = T extends (infer E)[] ? E : T;
type ReturnOf<T> = T extends (...args: any[]) => infer R ? R : never;

// Multiple infer
type FirstLast<T> = T extends [infer F, ...any[], infer L] ? [F, L] : never;

// Nested conditional
type TypeName<T> =
  T extends string ? "string" :
  T extends number ? "number" :
  T extends boolean ? "boolean" :
  T extends undefined ? "undefined" :
  T extends Function ? "function" :
  "object";
```

---

## Mapped Types Cheat Sheet

```typescript
// Basic mapped type
type Optional<T> = { [K in keyof T]?: T[K] };

// Add/remove modifiers
type Mutable<T> = { -readonly [K in keyof T]: T[K] };
type Concrete<T> = { [K in keyof T]-?: T[K] };

// Key remapping (as clause)
type Prefixed<T, P extends string> = {
  [K in keyof T as `${P}${Capitalize<string & K>}`]: T[K];
};

// Filter by value type
type StringProps<T> = {
  [K in keyof T as T[K] extends string ? K : never]: T[K];
};

// Transform values
type Promisify<T> = {
  [K in keyof T]: T[K] extends (...args: infer A) => infer R
    ? (...args: A) => Promise<R>
    : Promise<T[K]>;
};

// Homomorphic (preserves modifiers from source)
type MyPick<T, K extends keyof T> = { [P in K]: T[P] };

// Non-homomorphic (doesn't preserve modifiers)
type MyRecord<K extends keyof any, V> = { [P in K]: V };
```

---

## Type Narrowing Reference

| Technique | Example | Narrows To |
|-----------|---------|------------|
| `typeof` | `typeof x === "string"` | `string` |
| `instanceof` | `x instanceof Date` | `Date` |
| `in` | `"bark" in animal` | Type with `bark` |
| `===` / `!==` | `x === null` | `null` / non-null |
| Truthiness | `if (x)` | Non-falsy |
| Type guard | `isUser(x)` | `User` |
| Assertion | `assertDefined(x)` | Non-null onwards |
| Discriminant | `x.type === "circle"` | `Circle` variant |
| Assignment | `x = "hello"` | `string` at that point |

---

## Quick Patterns

```typescript
// Prettify — expand intersections for readable hover
type Prettify<T> = { [K in keyof T]: T[K] } & {};

// RequireAtLeastOne
type RequireAtLeastOne<T> = {
  [K in keyof T]-?: Required<Pick<T, K>> & Partial<Pick<T, Exclude<keyof T, K>>>;
}[keyof T];

// RequireExactlyOne
type RequireExactlyOne<T, Keys extends keyof T = keyof T> = Pick<
  T,
  Exclude<keyof T, Keys>
> &
  {
    [K in Keys]-?: Required<Pick<T, K>> &
      Partial<Record<Exclude<Keys, K>, undefined>>;
  }[Keys];

// XOR — exactly one of two types
type XOR<T, U> = (T | U) extends object
  ? (Without<T, U> & U) | (Without<U, T> & T)
  : T | U;
type Without<T, U> = { [K in Exclude<keyof T, keyof U>]?: never };

// StrictOmit — only allows keys that exist
type StrictOmit<T, K extends keyof T> = Pick<T, Exclude<keyof T, K>>;

// Entries / FromEntries
type Entries<T> = { [K in keyof T]: [K, T[K]] }[keyof T][];
type FromEntries<T extends [string, any]> = { [E in T as E[0]]: E[1] };

// MergeTypes
type MergeTypes<A, B> = Omit<A, keyof B> & B;

// ValueOf
type ValueOf<T> = T[keyof T];
```
