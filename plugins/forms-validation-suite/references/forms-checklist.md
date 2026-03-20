# Forms & Validation Checklist

## Setup Checklist

| Step | Command/Action |
|------|---------------|
| Install deps | `npm install react-hook-form @hookform/resolvers zod` |
| File uploads | `npm install react-dropzone` (if needed) |
| S3 uploads | `npm install @aws-sdk/client-s3 @aws-sdk/s3-request-presigner` |
| Server uploads | `npm install multer @types/multer` |

## Form Architecture Decision

| Question | If Yes |
|----------|--------|
| More than 5 fields? | Consider multi-step wizard |
| File uploads? | Use react-dropzone + presigned URLs |
| UI library (shadcn, MUI)? | Use Controller, not register |
| Server + client validation? | Shared Zod schemas in `schemas/` directory |
| Dynamic field count? | useFieldArray |
| Cross-field validation? | `.refine()` or `.superRefine()` on the object |

## Per-Form Checklist

### Schema

- [ ] Define Zod schema with specific error messages
- [ ] Use `z.string().min(1, 'Required')` not just `z.string()` for required fields
- [ ] Use `z.coerce.number()` or `{ valueAsNumber: true }` for number inputs
- [ ] Infer TypeScript type with `z.infer<typeof schema>`
- [ ] Cross-field validation uses `.refine()` with `path` option
- [ ] Server shares the same schema (import from shared `schemas/` dir)

### Form Setup

- [ ] `useForm` with `zodResolver(schema)`
- [ ] `defaultValues` match schema shape exactly
- [ ] `mode: 'onTouched'` for best UX (validates after first interaction)
- [ ] Destructure only needed `formState` properties (tracked via refs)

### Fields

- [ ] Every field has a visible `<label>` or `aria-label`
- [ ] Error messages shown below the field
- [ ] `aria-invalid={!!error}` on inputs
- [ ] `aria-describedby` linking to error message or description
- [ ] Number inputs use `{ valueAsNumber: true }` in register options
- [ ] File inputs use `onChange` + `setValue` (not register directly)
- [ ] Select/Switch/DatePicker wrapped in Controller

### Submission

- [ ] `handleSubmit` wraps the submit handler (don't `e.preventDefault()` manually)
- [ ] Submit button disabled during `isSubmitting`
- [ ] Server errors mapped to fields with `setError('fieldName', ...)`
- [ ] Catch-all errors shown at form top with `setError('root', ...)`
- [ ] `reset()` called after successful submission (if appropriate)

### UX

- [ ] First field auto-focused on mount
- [ ] Tab order is logical (left-to-right, top-to-bottom)
- [ ] Required fields marked (asterisk or "(required)" text)
- [ ] Loading state on submit button (spinner or text change)
- [ ] Error state doesn't clear user input
- [ ] Confirm before destructive actions (delete, discard changes)
- [ ] Unsaved changes warning on navigation (`beforeunload`)

### Accessibility

- [ ] All inputs have associated labels (htmlFor + id)
- [ ] Error messages are `role="alert"` or in `aria-live` region
- [ ] Form has a heading or `aria-label`
- [ ] Color is not the only error indicator (icon + text too)
- [ ] Focus moves to first error after failed submission
- [ ] Disabled inputs have `aria-disabled` and visible styling

## Validation Strategy by Form Type

| Form Type | Mode | Validation | Notes |
|-----------|------|-----------|-------|
| Login | `onSubmit` | Email + password presence | Don't reveal which field is wrong |
| Signup | `onTouched` | Full validation + async email check | Show password requirements |
| Settings | `onTouched` | Per-field as user edits | Partial updates with `.partial()` |
| Checkout | `onTouched` | Per-step in wizard | Only validate current step |
| Search/Filter | `onChange` | Minimal, debounce heavy ops | URL state via `useSearchParams` |
| File Upload | Immediate | On drop/select, before upload | Show errors next to dropzone |
| Data Table Edit | `onBlur` | Inline editing | Validate single cell, not whole row |

## Error Message Patterns

### Good Messages

```
Email: "Please enter a valid email address"
Password: "Password must be at least 8 characters with one number"
Phone: "Please enter a 10-digit phone number"
URL: "Please enter a full URL starting with https://"
Number range: "Please enter a number between 1 and 100"
Date: "Date must be in the future"
File: "Please upload a PNG or JPG under 5 MB"
Confirm: "Passwords do not match"
Required: "First name is required" (field-specific, never generic)
```

### Bad Messages (Avoid These)

```
"Invalid input"           → What's wrong? How do I fix it?
"Error"                   → Useless
"This field is required"  → Which field? Say "Email is required"
"Too short"               → How long should it be?
"Invalid format"          → What format is expected?
"You entered an invalid…" → Don't blame the user
```

## Common Zod Patterns Quick Reference

```typescript
// Required string (not empty)
z.string().min(1, 'Required')

// Email
z.string().email('Invalid email')

// Number from HTML input
z.coerce.number().min(0).max(100)

// Optional with default
z.string().optional().default('')

// Enum
z.enum(['option1', 'option2', 'option3'])

// Password with rules
z.string()
  .min(8, 'At least 8 characters')
  .regex(/[A-Z]/, 'Include uppercase')
  .regex(/[0-9]/, 'Include a number')

// Confirm password (object-level refine)
z.object({
  password: z.string().min(8),
  confirm: z.string(),
}).refine((d) => d.password === d.confirm, {
  message: 'Passwords do not match',
  path: ['confirm'],
})

// Date range
z.object({
  start: z.coerce.date(),
  end: z.coerce.date(),
}).refine((d) => d.end > d.start, {
  message: 'End must be after start',
  path: ['end'],
})

// File upload
z.instanceof(File)
  .refine((f) => f.size <= 5 * 1024 * 1024, 'Max 5 MB')
  .refine((f) => ['image/jpeg', 'image/png'].includes(f.type), 'JPEG or PNG only')

// Discriminated union (conditional fields)
z.discriminatedUnion('type', [
  z.object({ type: z.literal('personal'), name: z.string() }),
  z.object({ type: z.literal('business'), name: z.string(), taxId: z.string() }),
])

// Create from base (omit server fields)
const CreateSchema = BaseSchema.omit({ id: true, createdAt: true });

// Update (all optional)
const UpdateSchema = CreateSchema.partial();

// Pick specific fields
const ProfileSchema = BaseSchema.pick({ name: true, email: true });
```

## Multi-Step Form Checklist

- [ ] Per-step Zod schemas defined
- [ ] `trigger(fieldNames)` validates only current step's fields
- [ ] Back button does NOT trigger validation
- [ ] Progress indicator shows completed/current/upcoming steps
- [ ] Review step shows all entered data with "Edit" links
- [ ] Draft persistence via localStorage (cleared on submit)
- [ ] Conditional steps recalculate correctly when values change
- [ ] Final submission validates the full merged schema

## File Upload Checklist

- [ ] Max file size enforced (client + server)
- [ ] Allowed file types enforced (client + server)
- [ ] Image preview uses `URL.createObjectURL()` with cleanup (`revokeObjectURL`)
- [ ] Upload progress shown for large files (XHR or chunked)
- [ ] Drag-and-drop zone with visual feedback (active, reject states)
- [ ] Multiple file management (add, remove individual files)
- [ ] Don't set `Content-Type` header manually with FormData
- [ ] Presigned URLs generated fresh (5 min expiry)
- [ ] Error handling for failed uploads (retry or user feedback)
- [ ] Server-side: multer configured with size/type limits + error handler
