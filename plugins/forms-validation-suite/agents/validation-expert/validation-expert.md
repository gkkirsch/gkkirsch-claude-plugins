---
name: validation-expert
description: >
  Expert on validation patterns — shared client/server schemas, custom
  validators, conditional validation, and error message design.
  Triggers: "validation pattern", "schema design", "custom validator",
  "conditional validation", "error messages".
  NOT for: Zod API reference (use zod-schemas skill).
tools: Read, Glob, Grep
---

# Validation Patterns

## Shared Client/Server Schemas

The most powerful pattern: define Zod schemas once, use them on both client and server.

```
schemas/
  user.ts         → UserSchema, CreateUserSchema, UpdateUserSchema
  post.ts         → PostSchema, CreatePostSchema
  auth.ts         → LoginSchema, RegisterSchema

Client:
  Form uses CreateUserSchema with zodResolver
  TypeScript types inferred from schema

Server:
  API route validates request.body with CreateUserSchema.parse()
  Same validation rules, guaranteed in sync
```

### Schema Design Rules

1. **Base schema first.** Define the complete entity shape, then derive create/update schemas.

```typescript
// The full entity (what the DB returns)
const UserSchema = z.object({
  id: z.string().uuid(),
  name: z.string().min(1),
  email: z.string().email(),
  role: z.enum(['user', 'admin']),
  createdAt: z.coerce.date(),
});

// Create: omit server-generated fields
const CreateUserSchema = UserSchema.omit({ id: true, createdAt: true });

// Update: make everything optional
const UpdateUserSchema = CreateUserSchema.partial();
```

2. **Type inference from schemas.** Never define types manually.

```typescript
type User = z.infer<typeof UserSchema>;
type CreateUser = z.infer<typeof CreateUserSchema>;
type UpdateUser = z.infer<typeof UpdateUserSchema>;
```

3. **Validation at boundaries.** Validate at:
   - Form submission (client)
   - API route entry (server)
   - External data ingestion (webhooks, CSV imports)

## Error Message Design

### Rules

- **Be specific.** "Email is required" not "This field is required."
- **Be helpful.** "Password must be at least 8 characters" not "Invalid password."
- **Be human.** "Looks like that email is already taken" not "Duplicate key constraint violation."
- **Never blame.** "Please enter a valid email" not "You entered an invalid email."

### Patterns by Field Type

| Field | Good Error | Bad Error |
|-------|-----------|-----------|
| Email | "Please enter a valid email address" | "Invalid format" |
| Password | "Password must be at least 8 characters with one number" | "Too short" |
| Phone | "Please enter a 10-digit phone number" | "Invalid phone" |
| URL | "Please enter a full URL starting with https://" | "Invalid URL" |
| Number | "Please enter a number between 1 and 100" | "Out of range" |
| Date | "Date must be in the future" | "Invalid date" |
| File | "Please upload a PNG or JPG under 5 MB" | "Invalid file" |

### Zod Custom Error Messages

```typescript
const schema = z.object({
  email: z
    .string({ required_error: 'Email is required' })
    .email('Please enter a valid email address'),
  password: z
    .string({ required_error: 'Password is required' })
    .min(8, 'Password must be at least 8 characters')
    .regex(/[0-9]/, 'Password must include at least one number')
    .regex(/[a-z]/, 'Password must include a lowercase letter')
    .regex(/[A-Z]/, 'Password must include an uppercase letter'),
  age: z
    .number({ invalid_type_error: 'Age must be a number' })
    .int('Age must be a whole number')
    .min(13, 'You must be at least 13 years old')
    .max(120, 'Please enter a valid age'),
});
```

## Conditional Validation

### Field Depends on Another Field

```typescript
const schema = z.discriminatedUnion('accountType', [
  z.object({
    accountType: z.literal('personal'),
    name: z.string().min(1),
  }),
  z.object({
    accountType: z.literal('business'),
    name: z.string().min(1),
    companyName: z.string().min(1, 'Company name is required for business accounts'),
    taxId: z.string().min(1, 'Tax ID is required for business accounts'),
  }),
]);
```

### Refine for Cross-Field Validation

```typescript
const schema = z.object({
  password: z.string().min(8),
  confirmPassword: z.string(),
  startDate: z.coerce.date(),
  endDate: z.coerce.date(),
}).refine((data) => data.password === data.confirmPassword, {
  message: 'Passwords do not match',
  path: ['confirmPassword'],
}).refine((data) => data.endDate > data.startDate, {
  message: 'End date must be after start date',
  path: ['endDate'],
});
```

## Common Validators

```typescript
// Reusable validators
const validators = {
  email: z.string().email('Please enter a valid email'),
  password: z.string()
    .min(8, 'At least 8 characters')
    .regex(/[A-Z]/, 'Include an uppercase letter')
    .regex(/[0-9]/, 'Include a number'),
  phone: z.string().regex(/^\+?[1-9]\d{9,14}$/, 'Enter a valid phone number'),
  url: z.string().url('Enter a valid URL'),
  slug: z.string().regex(/^[a-z0-9-]+$/, 'Only lowercase letters, numbers, and hyphens'),
  hexColor: z.string().regex(/^#[0-9A-Fa-f]{6}$/, 'Enter a valid hex color'),
  uuid: z.string().uuid('Invalid ID format'),
  positiveInt: z.number().int().positive(),
  money: z.number().multipleOf(0.01).nonnegative(),
  futureDate: z.coerce.date().min(new Date(), 'Date must be in the future'),
  fileSize: (maxMB: number) =>
    z.instanceof(File).refine((f) => f.size <= maxMB * 1024 * 1024, `File must be under ${maxMB} MB`),
  imageType: z.instanceof(File).refine(
    (f) => ['image/jpeg', 'image/png', 'image/webp'].includes(f.type),
    'Only JPEG, PNG, and WebP images are allowed'
  ),
};
```

## Server-Side Error Mapping

```typescript
// Map server validation errors to RHF field errors
async function onSubmit(data: FormValues) {
  try {
    await createUser(data);
  } catch (error) {
    if (error.status === 422 && error.details) {
      // Server returns: { details: [{ field: 'email', message: 'Already taken' }] }
      for (const detail of error.details) {
        form.setError(detail.field as keyof FormValues, {
          type: 'server',
          message: detail.message,
        });
      }
    } else {
      form.setError('root', {
        type: 'server',
        message: error.message ?? 'Something went wrong. Please try again.',
      });
    }
  }
}
```
