---
name: advanced-types
description: >
  Advanced TypeScript types — generics, conditional types, mapped types, template literal types,
  infer keyword, recursive types, and type-level programming.
  Triggers: "generic type", "conditional type", "mapped type", "template literal type",
  "infer keyword", "recursive type", "type gymnastics", "advanced typescript".
  NOT for: basic TypeScript (use type-safe-patterns), utility types (use utility-patterns).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Advanced TypeScript Types

## Generics

### Constrained Generics

```typescript
// Basic constraint
function getLength<T extends { length: number }>(item: T): number {
  return item.length;
}
getLength('hello');      // 5
getLength([1, 2, 3]);    // 3
getLength({ length: 10 }); // 10

// keyof constraint
function getProperty<T, K extends keyof T>(obj: T, key: K): T[K] {
  return obj[key];
}
const user = { name: 'Alice', age: 30 };
getProperty(user, 'name');  // string
getProperty(user, 'age');   // number
// getProperty(user, 'foo'); // Error: 'foo' not assignable to 'name' | 'age'

// Multiple constraints
function merge<T extends object, U extends object>(a: T, b: U): T & U {
  return { ...a, ...b };
}

// Default type parameter
function createState<T = string>(initial: T): { value: T; set: (v: T) => void } {
  let value = initial;
  return { value, set: (v) => { value = v; } };
}
const state = createState('hello');     // T = string (inferred)
const numState = createState<number>(0); // T = number (explicit)
```

### Generic Patterns

```typescript
// Factory pattern
function createCollection<T>() {
  const items: T[] = [];
  return {
    add: (item: T) => { items.push(item); },
    get: (index: number): T | undefined => items[index],
    getAll: (): readonly T[] => items,
    filter: (predicate: (item: T) => boolean): T[] => items.filter(predicate),
  };
}
const users = createCollection<{ name: string; age: number }>();
users.add({ name: 'Alice', age: 30 });

// Builder pattern with generics
class QueryBuilder<T> {
  private filters: Partial<T> = {};
  private sortField?: keyof T;

  where<K extends keyof T>(field: K, value: T[K]): this {
    this.filters[field] = value;
    return this;
  }

  orderBy(field: keyof T): this {
    this.sortField = field;
    return this;
  }

  build(): { filters: Partial<T>; sort?: keyof T } {
    return { filters: this.filters, sort: this.sortField };
  }
}

const query = new QueryBuilder<{ name: string; age: number; active: boolean }>()
  .where('active', true)
  .where('age', 30)
  .orderBy('name')
  .build();
```

## Conditional Types

```typescript
// Basic conditional type
type IsString<T> = T extends string ? true : false;
type A = IsString<string>;  // true
type B = IsString<number>;  // false

// Extract and Exclude (built-in, but here's how they work)
type MyExtract<T, U> = T extends U ? T : never;
type MyExclude<T, U> = T extends U ? never : T;

type Numbers = Extract<string | number | boolean, number>; // number
type NoStrings = Exclude<string | number | boolean, string>; // number | boolean

// Distributive conditional types (unions distribute)
type ToArray<T> = T extends any ? T[] : never;
type Result = ToArray<string | number>; // string[] | number[]

// Non-distributive (wrap in tuple to prevent)
type ToArrayNonDist<T> = [T] extends [any] ? T[] : never;
type Result2 = ToArrayNonDist<string | number>; // (string | number)[]
```

### The `infer` Keyword

```typescript
// Extract return type
type MyReturnType<T> = T extends (...args: any[]) => infer R ? R : never;
type Returned = MyReturnType<() => string>;  // string

// Extract promise value
type Awaited<T> = T extends Promise<infer U> ? Awaited<U> : T;
type Value = Awaited<Promise<Promise<string>>>;  // string

// Extract array element
type ElementType<T> = T extends (infer U)[] ? U : never;
type Elem = ElementType<string[]>;  // string

// Extract first argument
type FirstArg<T> = T extends (first: infer F, ...rest: any[]) => any ? F : never;
type First = FirstArg<(name: string, age: number) => void>;  // string

// Extract function parameters
type MyParameters<T> = T extends (...args: infer P) => any ? P : never;
type Params = MyParameters<(a: string, b: number) => void>;  // [string, number]

// Infer from object shape
type PropType<T, K extends string> =
  T extends { [key in K]: infer V } ? V : never;
type NameType = PropType<{ name: string; age: number }, 'name'>;  // string
```

## Mapped Types

```typescript
// Basic mapped type
type Readonly<T> = { readonly [K in keyof T]: T[K] };
type Optional<T> = { [K in keyof T]?: T[K] };
type Nullable<T> = { [K in keyof T]: T[K] | null };

// Key remapping (as clause)
type Getters<T> = {
  [K in keyof T as `get${Capitalize<string & K>}`]: () => T[K];
};

interface Person { name: string; age: number; }
type PersonGetters = Getters<Person>;
// { getName: () => string; getAge: () => number; }

// Filter keys by value type
type OnlyStrings<T> = {
  [K in keyof T as T[K] extends string ? K : never]: T[K];
};

type StringFields = OnlyStrings<{ name: string; age: number; email: string }>;
// { name: string; email: string }

// Modify value types
type Promised<T> = { [K in keyof T]: Promise<T[K]> };
type AsyncPerson = Promised<Person>;
// { name: Promise<string>; age: Promise<number>; }

// Remove readonly
type Mutable<T> = { -readonly [K in keyof T]: T[K] };

// Required (remove optional)
type MyRequired<T> = { [K in keyof T]-?: T[K] };
```

## Template Literal Types

```typescript
// Basic template literals
type EventName = `on${Capitalize<'click' | 'focus' | 'blur'>}`;
// 'onClick' | 'onFocus' | 'onBlur'

type HTTPMethod = 'GET' | 'POST' | 'PUT' | 'DELETE';
type Endpoint = `/${string}`;
type Route = `${HTTPMethod} ${Endpoint}`;
// 'GET /...' | 'POST /...' | 'PUT /...' | 'DELETE /...'

// CSS units
type CSSLength = `${number}${'px' | 'rem' | 'em' | '%' | 'vh' | 'vw'}`;
const padding: CSSLength = '16px';  // OK
// const bad: CSSLength = '16';     // Error

// Type-safe event emitter
type EventMap = {
  click: { x: number; y: number };
  focus: { target: HTMLElement };
  change: { value: string };
};

type EventHandler<T extends keyof EventMap> = (event: EventMap[T]) => void;

class TypedEmitter {
  on<T extends keyof EventMap>(event: T, handler: EventHandler<T>): void { /* ... */ }
  emit<T extends keyof EventMap>(event: T, data: EventMap[T]): void { /* ... */ }
}

const emitter = new TypedEmitter();
emitter.on('click', (event) => {
  console.log(event.x, event.y); // Fully typed!
});
```

## Discriminated Unions

```typescript
// State machine
type RequestState<T> =
  | { status: 'idle' }
  | { status: 'loading' }
  | { status: 'success'; data: T }
  | { status: 'error'; error: Error };

function renderState<T>(state: RequestState<T>) {
  switch (state.status) {
    case 'idle':
      return 'Ready';
    case 'loading':
      return 'Loading...';
    case 'success':
      return `Got: ${state.data}`;  // data is available here
    case 'error':
      return `Error: ${state.error.message}`;  // error is available here
  }
}

// API response
type ApiResponse =
  | { success: true; data: unknown }
  | { success: false; error: { code: string; message: string } };

// Action types (Redux-style)
type Action =
  | { type: 'ADD_TODO'; payload: { text: string } }
  | { type: 'TOGGLE_TODO'; payload: { id: string } }
  | { type: 'DELETE_TODO'; payload: { id: string } }
  | { type: 'CLEAR_COMPLETED' };

function reducer(state: State, action: Action): State {
  switch (action.type) {
    case 'ADD_TODO':
      return { ...state, todos: [...state.todos, { text: action.payload.text }] };
    case 'TOGGLE_TODO':
      return { ...state, /* toggle by action.payload.id */ };
    // TypeScript ensures all cases are handled with exhaustive check
  }
}

// Exhaustive check helper
function assertNever(x: never): never {
  throw new Error(`Unexpected value: ${x}`);
}
```

## Recursive Types

```typescript
// Deep readonly
type DeepReadonly<T> = {
  readonly [K in keyof T]: T[K] extends object ? DeepReadonly<T[K]> : T[K];
};

// Deep partial
type DeepPartial<T> = {
  [K in keyof T]?: T[K] extends object ? DeepPartial<T[K]> : T[K];
};

// JSON type
type JSONValue =
  | string
  | number
  | boolean
  | null
  | JSONValue[]
  | { [key: string]: JSONValue };

// Nested path type
type Path<T, K extends keyof T = keyof T> = K extends string
  ? T[K] extends object
    ? K | `${K}.${Path<T[K]>}`
    : K
  : never;

type UserPaths = Path<{
  name: string;
  address: { city: string; zip: string };
}>;
// 'name' | 'address' | 'address.city' | 'address.zip'
```

## Gotchas

1. **Conditional types distribute over unions.** `ToArray<string | number>` becomes `string[] | number[]`, not `(string | number)[]`. Wrap in tuple `[T]` to prevent distribution.

2. **`infer` only works in conditional type `extends` clause.** You can't use `infer` outside of a conditional type expression.

3. **Mapped types lose methods.** `{ [K in keyof T]: T[K] }` maps properties but can behave differently with class methods. Use `Pick` + `Omit` for selective mapping.

4. **Template literal types can create combinatorial explosion.** `${A | B | C}${D | E | F}` creates 9 types. Watch union sizes.

5. **Recursive types have a depth limit.** TypeScript limits type instantiation depth (~50 levels). Add a base case or simplify recursion.

6. **`keyof` on unions gives intersection of keys.** `keyof (A | B)` returns only keys that exist on BOTH A and B. For all keys, use `keyof A | keyof B`.
