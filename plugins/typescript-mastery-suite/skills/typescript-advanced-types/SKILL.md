---
name: typescript-advanced-types
description: >
  Advanced TypeScript type system — mapped types, conditional types, template
  literal types, branded types, type guards, discriminated unions, satisfies,
  const assertions, infer keyword, and utility type patterns.
  Triggers: "typescript types", "mapped types", "conditional types",
  "branded types", "type guards", "discriminated unions", "typescript generics".
  NOT for: runtime patterns (use typescript-patterns).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Advanced TypeScript Types

## Mapped Types

```typescript
// Make all properties optional
type Partial<T> = { [K in keyof T]?: T[K] };

// Make all properties required
type Required<T> = { [K in keyof T]-?: T[K] };

// Make all properties readonly
type Readonly<T> = { readonly [K in keyof T]: T[K] };

// Transform property types
type Nullable<T> = { [K in keyof T]: T[K] | null };

// Rename keys with template literals
type Getters<T> = {
  [K in keyof T as `get${Capitalize<string & K>}`]: () => T[K];
};

interface User {
  name: string;
  age: number;
}

type UserGetters = Getters<User>;
// { getName: () => string; getAge: () => number }

// Filter properties by type
type PickByType<T, ValueType> = {
  [K in keyof T as T[K] extends ValueType ? K : never]: T[K];
};

type StringProps = PickByType<User, string>;
// { name: string }

// Map over union
type EventMap<Events extends string> = {
  [E in Events as `on${Capitalize<E>}`]: (event: E) => void;
};

type Handlers = EventMap<"click" | "hover" | "focus">;
// { onClick: (event: "click") => void; onHover: ...; onFocus: ... }
```

## Conditional Types

```typescript
// Basic conditional
type IsString<T> = T extends string ? true : false;
type A = IsString<"hello">;  // true
type B = IsString<42>;       // false

// Extract return type
type ReturnType<T> = T extends (...args: any[]) => infer R ? R : never;

// Extract promise value
type Awaited<T> = T extends Promise<infer U> ? Awaited<U> : T;
type Result = Awaited<Promise<Promise<string>>>;  // string

// Extract array element
type ElementOf<T> = T extends (infer E)[] ? E : never;
type Item = ElementOf<string[]>;  // string

// Distributive conditional types (unions distribute automatically)
type ToArray<T> = T extends any ? T[] : never;
type Distributed = ToArray<string | number>;  // string[] | number[]

// Prevent distribution with [T]
type ToArrayNonDist<T> = [T] extends [any] ? T[] : never;
type NonDistributed = ToArrayNonDist<string | number>;  // (string | number)[]

// Extract from discriminated unions
type Extract<T, U> = T extends U ? T : never;
type Exclude<T, U> = T extends U ? never : T;

type Shape = { kind: "circle"; radius: number } | { kind: "square"; side: number };
type Circle = Extract<Shape, { kind: "circle" }>;
// { kind: "circle"; radius: number }
```

## Template Literal Types

```typescript
// String manipulation types
type EventName = `on${Capitalize<"click" | "hover">}`;
// "onClick" | "onHover"

// CSS units
type CSSUnit = "px" | "rem" | "em" | "%";
type CSSValue = `${number}${CSSUnit}`;
const width: CSSValue = "100px";    // OK
// const bad: CSSValue = "100vw";   // Error: not assignable

// Route params extraction
type ExtractParams<T extends string> =
  T extends `${string}:${infer Param}/${infer Rest}`
    ? Param | ExtractParams<Rest>
    : T extends `${string}:${infer Param}`
    ? Param
    : never;

type Params = ExtractParams<"/users/:userId/posts/:postId">;
// "userId" | "postId"

// HTTP method + path
type ApiRoute = `${"GET" | "POST" | "PUT" | "DELETE"} /${string}`;
const route: ApiRoute = "GET /users";  // OK

// Deep key paths
type PathKeys<T, Prefix extends string = ""> = T extends object
  ? {
      [K in keyof T & string]: T[K] extends object
        ? `${Prefix}${K}` | PathKeys<T[K], `${Prefix}${K}.`>
        : `${Prefix}${K}`;
    }[keyof T & string]
  : never;

interface Config {
  db: { host: string; port: number };
  cache: { ttl: number };
}

type ConfigPaths = PathKeys<Config>;
// "db" | "db.host" | "db.port" | "cache" | "cache.ttl"
```

## Branded Types

```typescript
// Nominal typing via branding
declare const __brand: unique symbol;
type Brand<T, B extends string> = T & { [__brand]: B };

type UserId = Brand<string, "UserId">;
type PostId = Brand<string, "PostId">;
type Email = Brand<string, "Email">;

// Constructor functions enforce validation
function UserId(id: string): UserId {
  if (!id.match(/^usr_[a-z0-9]+$/)) throw new Error("Invalid user ID format");
  return id as UserId;
}

function Email(email: string): Email {
  if (!email.includes("@")) throw new Error("Invalid email");
  return email as Email;
}

// Cannot accidentally swap branded types
function getUser(id: UserId): User { /* ... */ }
function getPost(id: PostId): Post { /* ... */ }

const userId = UserId("usr_abc123");
const postId = PostId("post_xyz789"); // assuming PostId constructor exists

getUser(userId);   // OK
// getUser(postId); // Error: PostId not assignable to UserId

// Branded primitives for units
type Kilometers = Brand<number, "Kilometers">;
type Miles = Brand<number, "Miles">;

function kmToMiles(km: Kilometers): Miles {
  return (km * 0.621371) as Miles;
}
```

## Type Guards

```typescript
// User-defined type guard with `is`
function isString(value: unknown): value is string {
  return typeof value === "string";
}

// Discriminated union guard
interface ApiSuccess<T> { status: "success"; data: T }
interface ApiError { status: "error"; message: string; code: number }
type ApiResponse<T> = ApiSuccess<T> | ApiError;

function isSuccess<T>(res: ApiResponse<T>): res is ApiSuccess<T> {
  return res.status === "success";
}

// Assertion function (throws instead of returning boolean)
function assertDefined<T>(value: T | null | undefined, name: string): asserts value is T {
  if (value === null || value === undefined) {
    throw new Error(`Expected ${name} to be defined`);
  }
}

const user = getUser(id);
assertDefined(user, "user");  // After this line, user is narrowed to User (non-null)
console.log(user.name);       // No null check needed

// Array type guard with filter
const items: (string | null)[] = ["a", null, "b", null, "c"];
const strings: string[] = items.filter((x): x is string => x !== null);

// in operator narrowing
interface Dog { bark(): void }
interface Cat { meow(): void }

function speak(animal: Dog | Cat) {
  if ("bark" in animal) {
    animal.bark();  // TypeScript knows it's Dog
  } else {
    animal.meow();  // TypeScript knows it's Cat
  }
}
```

## Discriminated Unions

```typescript
// State machine pattern
type RequestState<T> =
  | { status: "idle" }
  | { status: "loading" }
  | { status: "success"; data: T }
  | { status: "error"; error: Error };

function renderState<T>(state: RequestState<T>) {
  switch (state.status) {
    case "idle":
      return <div>Ready</div>;
    case "loading":
      return <Spinner />;
    case "success":
      return <DataView data={state.data} />;  // data is narrowed
    case "error":
      return <ErrorView error={state.error} />;  // error is narrowed
  }
}

// Exhaustive check helper
function assertNever(x: never): never {
  throw new Error(`Unexpected value: ${x}`);
}

// Event system
type AppEvent =
  | { type: "USER_LOGIN"; userId: string; timestamp: Date }
  | { type: "USER_LOGOUT"; userId: string }
  | { type: "PAGE_VIEW"; path: string; referrer?: string }
  | { type: "ERROR"; error: Error; context: string };

function handleEvent(event: AppEvent) {
  switch (event.type) {
    case "USER_LOGIN":
      analytics.track("login", { userId: event.userId });
      break;
    case "PAGE_VIEW":
      analytics.track("pageview", { path: event.path });
      break;
    case "ERROR":
      errorReporter.capture(event.error, { context: event.context });
      break;
    case "USER_LOGOUT":
      session.clear();
      break;
    default:
      assertNever(event);  // Compile error if a case is missed
  }
}
```

## The `satisfies` Operator

```typescript
// satisfies validates WITHOUT widening the type
type ColorMap = Record<string, string | [number, number, number]>;

// With `as const` + `satisfies`: validated AND precise
const colors = {
  red: [255, 0, 0],
  green: "#00ff00",
  blue: [0, 0, 255],
} as const satisfies ColorMap;

// colors.red is readonly [255, 0, 0] — tuple, not string | [number, number, number]
// colors.green is "#00ff00" — literal, not string | [number, number, number]
colors.red[0];     // 255 (precise)
colors.green.toUpperCase();  // OK (known to be string)

// Without satisfies — type is widened
const colorsWide: ColorMap = {
  red: [255, 0, 0],
  green: "#00ff00",
  blue: [0, 0, 255],
};
// colorsWide.green is string | [number, number, number] — lost precision

// Route configuration
type Route = { path: string; component: React.ComponentType; auth?: boolean };

const routes = {
  home: { path: "/", component: HomePage },
  dashboard: { path: "/dashboard", component: Dashboard, auth: true },
  settings: { path: "/settings", component: Settings, auth: true },
} satisfies Record<string, Route>;

// routes.home.path is "/" (literal), not just string
// Object.keys(routes) gives ("home" | "dashboard" | "settings")[]
```

## Utility Type Patterns

```typescript
// Deep partial (recursive)
type DeepPartial<T> = T extends object
  ? { [K in keyof T]?: DeepPartial<T[K]> }
  : T;

// Deep readonly
type DeepReadonly<T> = T extends object
  ? { readonly [K in keyof T]: DeepReadonly<T[K]> }
  : T;

// Make specific properties required
type RequireKeys<T, K extends keyof T> = T & Required<Pick<T, K>>;
type UserWithEmail = RequireKeys<Partial<User>, "email">;

// Mutable (remove readonly)
type Mutable<T> = { -readonly [K in keyof T]: T[K] };

// UnionToIntersection
type UnionToIntersection<U> = (U extends any ? (k: U) => void : never) extends (
  k: infer I
) => void
  ? I
  : never;

// Strict omit (only allows keys that exist)
type StrictOmit<T, K extends keyof T> = Pick<T, Exclude<keyof T, K>>;

// Function overloading types
type Overloads<T> = T extends {
  (...args: infer A1): infer R1;
  (...args: infer A2): infer R2;
}
  ? [(...args: A1) => R1, (...args: A2) => R2]
  : never;

// Builder pattern types
type Builder<T> = {
  [K in keyof T as `set${Capitalize<string & K>}`]: (value: T[K]) => Builder<T>;
} & { build(): T };
```

## Const Assertions & Enums

```typescript
// const assertion — narrowest possible type
const config = {
  endpoint: "https://api.example.com",
  retries: 3,
  methods: ["GET", "POST"],
} as const;
// typeof config.endpoint = "https://api.example.com" (not string)
// typeof config.retries = 3 (not number)
// typeof config.methods = readonly ["GET", "POST"] (not string[])

// String union enum (prefer over enum keyword)
const Status = {
  Active: "active",
  Inactive: "inactive",
  Pending: "pending",
} as const;
type Status = (typeof Status)[keyof typeof Status];
// "active" | "inactive" | "pending"

// Derive types from data
const ROLES = ["admin", "editor", "viewer"] as const;
type Role = (typeof ROLES)[number];  // "admin" | "editor" | "viewer"

function isRole(value: string): value is Role {
  return (ROLES as readonly string[]).includes(value);
}
```

## Gotchas

1. **`any` defeats the type system silently** — Use `unknown` for values of unknown type. `unknown` requires narrowing before use; `any` lets anything through without errors.

2. **`Record<string, T>` allows any string key** — If you want specific keys, use `Record<"key1" | "key2", T>` or a mapped type. `Record<string, T>` is too permissive.

3. **Generic defaults don't constrain** — `function fn<T = string>(x: T)` doesn't mean T must be string. It means T defaults to string if not inferred. Use `extends` to constrain: `<T extends string>`.

4. **Distributive conditional types are surprising** — `ToArray<string | number>` becomes `string[] | number[]`, not `(string | number)[]`. Wrap in tuple `[T]` to prevent distribution.

5. **`readonly` arrays need explicit typing** — `as const` makes arrays `readonly`. You can't pass a `readonly string[]` where `string[]` is expected. Accept `readonly string[]` in function parameters.

6. **`satisfies` doesn't add to the type** — `x satisfies T` validates x against T but x retains its inferred type. If you need x to BE type T, use `: T` annotation. Use `satisfies` when you want validation without widening.
