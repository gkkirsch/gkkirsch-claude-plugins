---
name: zod-schemas
description: >
  Zod schema design — primitives, objects, arrays, unions, discriminated unions,
  transforms, refinements, error maps, and shared client/server validation.
  Triggers: "zod", "zod schema", "zod validation", "z.object", "z.string",
  "schema validation", "type-safe validation".
  NOT for: form-specific patterns (use react-hook-form skill).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Zod Schemas

## Quick Start

```bash
npm install zod
```

## Primitives

```typescript
import { z } from 'zod';

// Strings
z.string()                          // any string
z.string().min(1, 'Required')       // non-empty (use instead of .nonempty())
z.string().max(255)                 // max length
z.string().email('Invalid email')   // email format
z.string().url()                    // URL format
z.string().uuid()                   // UUID format
z.string().cuid()                   // CUID format
z.string().regex(/^[a-z]+$/)       // custom regex
z.string().trim()                   // trim whitespace (transform)
z.string().toLowerCase()            // lowercase (transform)
z.string().startsWith('https://')   // prefix check
z.string().includes('@')            // contains check
z.string().datetime()               // ISO datetime string
z.string().ip()                     // IP address (v4 or v6)

// Numbers
z.number()                          // any number
z.number().int()                    // integer only
z.number().positive()               // > 0
z.number().nonnegative()            // >= 0
z.number().min(1).max(100)          // range
z.number().multipleOf(0.01)         // money (2 decimal places)
z.number().finite()                 // no Infinity
z.number().safe()                   // within Number.MAX_SAFE_INTEGER

// Coercion (string → type)
z.coerce.number()                   // "42" → 42
z.coerce.boolean()                  // "true" → true
z.coerce.date()                     // "2025-01-01" → Date
z.coerce.bigint()                   // "9007199254740991" → BigInt

// Other primitives
z.boolean()
z.date()
z.bigint()
z.undefined()
z.null()
z.void()
z.any()
z.unknown()                         // safer than any — must narrow before use
```

## Objects

```typescript
// Basic object
const UserSchema = z.object({
  id: z.string().uuid(),
  name: z.string().min(1),
  email: z.string().email(),
  age: z.number().int().min(0).optional(),
  role: z.enum(['user', 'admin', 'moderator']),
  settings: z.object({
    theme: z.enum(['light', 'dark']),
    notifications: z.boolean(),
  }),
  createdAt: z.coerce.date(),
});

// Type inference
type User = z.infer<typeof UserSchema>;

// Derivations
const CreateUserSchema = UserSchema.omit({ id: true, createdAt: true });
const UpdateUserSchema = CreateUserSchema.partial();
const UserProfileSchema = UserSchema.pick({ name: true, email: true, age: true });

// Extend
const AdminSchema = UserSchema.extend({
  permissions: z.array(z.string()),
  department: z.string(),
});

// Merge two schemas
const FullProfileSchema = UserSchema.merge(AddressSchema);

// Strip unknown keys (default behavior)
UserSchema.parse(dataWithExtraKeys); // Extra keys removed

// Strict — error on unknown keys
const StrictUserSchema = UserSchema.strict();

// Passthrough — keep unknown keys
const PassthroughSchema = UserSchema.passthrough();
```

## Arrays

```typescript
z.array(z.string())                         // string[]
z.array(z.number()).min(1, 'At least one')  // non-empty array
z.array(z.string()).max(10)                 // max 10 items
z.array(z.string()).length(3)               // exactly 3 items
z.array(z.string()).nonempty()              // [string, ...string[]]

// Tuple (fixed-length array with different types)
z.tuple([z.string(), z.number(), z.boolean()])
// → [string, number, boolean]

// Set
z.set(z.string())                           // Set<string>
z.set(z.number()).min(1).max(5)
```

## Enums and Unions

```typescript
// Zod enum (string literal union)
const RoleSchema = z.enum(['user', 'admin', 'moderator']);
type Role = z.infer<typeof RoleSchema>; // 'user' | 'admin' | 'moderator'
RoleSchema.options; // ['user', 'admin', 'moderator']

// Native enum
enum Status {
  Active = 'active',
  Inactive = 'inactive',
}
const StatusSchema = z.nativeEnum(Status);

// Union
const StringOrNumber = z.union([z.string(), z.number()]);

// Discriminated union (best for tagged types)
const EventSchema = z.discriminatedUnion('type', [
  z.object({ type: z.literal('click'), x: z.number(), y: z.number() }),
  z.object({ type: z.literal('keydown'), key: z.string() }),
  z.object({ type: z.literal('scroll'), delta: z.number() }),
]);

// Literal
z.literal('active')    // exactly 'active'
z.literal(42)          // exactly 42
z.literal(true)        // exactly true
```

## Optional, Nullable, Default

```typescript
z.string().optional()                // string | undefined
z.string().nullable()                // string | null
z.string().nullish()                 // string | null | undefined
z.string().default('hello')          // defaults to 'hello' if undefined
z.string().optional().default('hi')  // optional input, defaults if not provided

// Transform null to undefined
z.string().nullable().transform((val) => val ?? undefined)
```

## Refinements and Transforms

```typescript
// Refine — custom validation
const PasswordSchema = z.string()
  .min(8, 'At least 8 characters')
  .refine((val) => /[A-Z]/.test(val), 'Must include uppercase letter')
  .refine((val) => /[0-9]/.test(val), 'Must include a number')
  .refine((val) => /[^A-Za-z0-9]/.test(val), 'Must include special character');

// Superrefine — multiple errors at once
const PasswordSchemaV2 = z.string().superRefine((val, ctx) => {
  if (val.length < 8) {
    ctx.addIssue({ code: z.ZodIssueCode.custom, message: 'At least 8 characters' });
  }
  if (!/[A-Z]/.test(val)) {
    ctx.addIssue({ code: z.ZodIssueCode.custom, message: 'Must include uppercase letter' });
  }
  if (!/[0-9]/.test(val)) {
    ctx.addIssue({ code: z.ZodIssueCode.custom, message: 'Must include a number' });
  }
});

// Transform — modify the value
const TrimmedEmail = z.string()
  .email()
  .transform((val) => val.toLowerCase().trim());

// Transform with validation
const SafeInt = z.string()
  .transform((val) => parseInt(val, 10))
  .pipe(z.number().int().positive());

// Object-level refine (cross-field)
const DateRangeSchema = z.object({
  start: z.coerce.date(),
  end: z.coerce.date(),
}).refine((data) => data.end > data.start, {
  message: 'End date must be after start date',
  path: ['end'],
});

// Password confirmation
const SignupSchema = z.object({
  password: z.string().min(8),
  confirmPassword: z.string(),
}).refine((data) => data.password === data.confirmPassword, {
  message: 'Passwords do not match',
  path: ['confirmPassword'],
});
```

## Records and Maps

```typescript
// Record (string keys, typed values)
z.record(z.string(), z.number())
// → Record<string, number> → { [key: string]: number }

// Map
z.map(z.string(), z.object({ name: z.string() }))
// → Map<string, { name: string }>

// Record with key validation
z.record(
  z.string().regex(/^[a-z_]+$/),  // key must be lowercase + underscores
  z.boolean()                      // value must be boolean
)
```

## Parsing and Validation

```typescript
const schema = z.object({
  name: z.string(),
  age: z.number(),
});

// parse — throws ZodError on failure
const result = schema.parse({ name: 'Alice', age: 30 });

// safeParse — returns { success, data } or { success, error }
const result = schema.safeParse(input);
if (result.success) {
  console.log(result.data); // typed as { name: string, age: number }
} else {
  console.log(result.error.flatten());
  // { formErrors: [], fieldErrors: { name: ['...'], age: ['...'] } }
}

// parseAsync — for schemas with async refinements
const result = await schema.parseAsync(input);
const result = await schema.safeParseAsync(input);
```

## Error Formatting

```typescript
const result = schema.safeParse(input);

if (!result.success) {
  // Flat format (best for forms)
  const flat = result.error.flatten();
  // { formErrors: string[], fieldErrors: { [field]: string[] } }

  // Formatted (nested, matches schema shape)
  const formatted = result.error.format();
  // { name: { _errors: ['...'] }, age: { _errors: ['...'] } }

  // Issues array (raw)
  result.error.issues;
  // [{ code: 'too_small', message: '...', path: ['name'], ... }]
}
```

## Express Middleware

```typescript
import { z, ZodSchema } from 'zod';
import { Request, Response, NextFunction } from 'express';

function validate(schema: {
  body?: ZodSchema;
  query?: ZodSchema;
  params?: ZodSchema;
}) {
  return (req: Request, res: Response, next: NextFunction) => {
    try {
      if (schema.body) req.body = schema.body.parse(req.body);
      if (schema.query) req.query = schema.query.parse(req.query) as any;
      if (schema.params) req.params = schema.params.parse(req.params) as any;
      next();
    } catch (error) {
      if (error instanceof z.ZodError) {
        res.status(422).json({
          error: { code: 'VALIDATION_ERROR', details: error.flatten().fieldErrors },
        });
      } else {
        next(error);
      }
    }
  };
}

// Usage
app.post('/api/users',
  validate({ body: CreateUserSchema }),
  async (req, res) => {
    // req.body is typed and validated
    const user = await db.user.create({ data: req.body });
    res.status(201).json(user);
  }
);
```

## Gotchas

1. **`z.string()` accepts empty strings.** Use `z.string().min(1, 'Required')` for required fields.

2. **HTML inputs return strings.** Use `z.coerce.number()` or `{ valueAsNumber: true }` in register for number fields.

3. **`.optional()` means `undefined`, not empty string.** An empty input is `''` (string), not `undefined`. Handle with `.transform()` or `.refine()`.

4. **`.refine()` runs after `.parse()`.** Refinements only run if the base type check passes. A `.refine()` on a string won't run if the value is a number.

5. **`z.infer` is your type source.** Never define TypeScript interfaces separately when you have a Zod schema. Always infer.

6. **Transforms change the output type.** `z.string().transform(Number)` outputs `number`. `z.infer<>` reflects the output type.
