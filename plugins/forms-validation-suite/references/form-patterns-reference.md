# Form Patterns Reference

## Form Type Decision Matrix

| Form Type | Complexity | Recommended Pattern |
|-----------|-----------|-------------------|
| Login/signup | Simple | Single form, `mode: 'onTouched'`, zodResolver |
| Contact/feedback | Simple | Single form, textarea, optional file upload |
| Settings/profile | Medium | Tabbed sections, auto-save on blur |
| Checkout | Medium | Multi-step wizard with step validation |
| Data entry (table) | Complex | useFieldArray, inline editing, bulk operations |
| Survey/quiz | Complex | Dynamic fields, conditional logic, progress bar |
| Admin CRUD | Complex | Shared schemas client/server, server errors |

## Schema Composition Patterns

### Base → Create → Update

```typescript
// Full entity (what the database returns)
const UserSchema = z.object({
  id: z.string().uuid(),
  name: z.string().min(1),
  email: z.string().email(),
  role: z.enum(['user', 'admin']),
  createdAt: z.coerce.date(),
  updatedAt: z.coerce.date(),
});

// Create: omit auto-generated fields
const CreateUserSchema = UserSchema.omit({
  id: true,
  createdAt: true,
  updatedAt: true,
});

// Update: make everything optional
const UpdateUserSchema = CreateUserSchema.partial();

// Types inferred from schemas
type User = z.infer<typeof UserSchema>;
type CreateUser = z.infer<typeof CreateUserSchema>;
type UpdateUser = z.infer<typeof UpdateUserSchema>;
```

### Shared API Request/Response

```typescript
// Shared between client and server
const PaginationSchema = z.object({
  page: z.coerce.number().int().min(1).default(1),
  limit: z.coerce.number().int().min(1).max(100).default(20),
  sortBy: z.string().optional(),
  sortOrder: z.enum(['asc', 'desc']).default('desc'),
});

const SearchSchema = PaginationSchema.extend({
  query: z.string().optional(),
  status: z.enum(['active', 'inactive', 'all']).default('all'),
});

// Server: validate query params
app.get('/api/users', validate({ query: SearchSchema }), handler);

// Client: build query string from form
const params = SearchSchema.parse(formValues);
```

### Discriminated Union for Polymorphic Forms

```typescript
const NotificationSchema = z.discriminatedUnion('channel', [
  z.object({
    channel: z.literal('email'),
    emailAddress: z.string().email(),
    frequency: z.enum(['instant', 'daily', 'weekly']),
  }),
  z.object({
    channel: z.literal('sms'),
    phoneNumber: z.string().regex(/^\+[1-9]\d{9,14}$/),
  }),
  z.object({
    channel: z.literal('webhook'),
    url: z.string().url(),
    secret: z.string().min(16),
  }),
]);
```

## Multi-Step Wizard Pattern

```tsx
const steps = [
  { id: 'account', schema: accountSchema, fields: ['email', 'password'] },
  { id: 'profile', schema: profileSchema, fields: ['name', 'bio'] },
  { id: 'preferences', schema: preferencesSchema, fields: ['theme', 'notifications'] },
];

function Wizard() {
  const [step, setStep] = useState(0);
  const [formData, setFormData] = useState({});
  const currentStep = steps[step];

  const form = useForm({
    resolver: zodResolver(currentStep.schema),
    defaultValues: formData,
  });

  const onStepSubmit = (data) => {
    const merged = { ...formData, ...data };
    setFormData(merged);

    if (step < steps.length - 1) {
      setStep(step + 1);
    } else {
      submitFinalForm(merged);
    }
  };

  return (
    <>
      <ProgressBar current={step} total={steps.length} />
      <form onSubmit={form.handleSubmit(onStepSubmit)}>
        {/* Render fields for currentStep */}
        <div className="flex justify-between">
          {step > 0 && (
            <button type="button" onClick={() => setStep(step - 1)}>
              Back
            </button>
          )}
          <button type="submit">
            {step === steps.length - 1 ? 'Submit' : 'Next'}
          </button>
        </div>
      </form>
    </>
  );
}
```

## File Upload Pattern

```tsx
const FileSchema = z.object({
  file: z
    .instanceof(File)
    .refine((f) => f.size <= 5 * 1024 * 1024, 'File must be under 5 MB')
    .refine(
      (f) => ['image/jpeg', 'image/png', 'image/webp'].includes(f.type),
      'Only JPEG, PNG, and WebP images allowed'
    ),
});

function FileUpload() {
  const form = useForm({ resolver: zodResolver(FileSchema) });
  const [preview, setPreview] = useState<string | null>(null);

  return (
    <Controller
      name="file"
      control={form.control}
      render={({ field, fieldState: { error } }) => (
        <div>
          <input
            type="file"
            accept="image/jpeg,image/png,image/webp"
            onChange={(e) => {
              const file = e.target.files?.[0];
              if (file) {
                field.onChange(file);
                setPreview(URL.createObjectURL(file));
              }
            }}
          />
          {preview && <img src={preview} alt="Preview" className="w-32 h-32 object-cover" />}
          {error && <p className="text-red-500">{error.message}</p>}
        </div>
      )}
    />
  );
}
```

## Auto-Save Pattern

```tsx
function AutoSaveForm({ initialData }) {
  const form = useForm({
    resolver: zodResolver(settingsSchema),
    defaultValues: initialData,
  });

  const mutation = useMutation({
    mutationFn: saveSettings,
  });

  // Watch all fields and auto-save on change (debounced)
  useEffect(() => {
    const subscription = form.watch(
      debounce((values) => {
        if (form.formState.isDirty) {
          mutation.mutate(values);
        }
      }, 1000)
    );
    return () => subscription.unsubscribe();
  }, [form.watch]);

  return (
    <form>
      {/* fields */}
      <span className="text-sm text-muted-foreground">
        {mutation.isPending ? 'Saving...' : mutation.isSuccess ? 'Saved' : ''}
      </span>
    </form>
  );
}
```

## Inline Editable Field Pattern

```tsx
function InlineEdit({ value, onSave, schema }) {
  const [editing, setEditing] = useState(false);
  const form = useForm({
    resolver: zodResolver(schema),
    defaultValues: { value },
  });

  if (!editing) {
    return (
      <span onClick={() => setEditing(true)} className="cursor-pointer hover:bg-muted px-1 rounded">
        {value}
      </span>
    );
  }

  return (
    <form
      onSubmit={form.handleSubmit(async (data) => {
        await onSave(data.value);
        setEditing(false);
      })}
      onBlur={form.handleSubmit(async (data) => {
        await onSave(data.value);
        setEditing(false);
      })}
    >
      <input {...form.register('value')} autoFocus className="border px-1 rounded" />
      {form.formState.errors.value && (
        <p className="text-red-500 text-xs">{form.formState.errors.value.message}</p>
      )}
    </form>
  );
}
```

## Common Validators Library

```typescript
export const validators = {
  // Strings
  required: z.string().min(1, 'Required'),
  email: z.string().min(1, 'Email is required').email('Invalid email address'),
  url: z.string().url('Enter a valid URL starting with https://'),
  slug: z.string().regex(/^[a-z0-9-]+$/, 'Only lowercase letters, numbers, and hyphens'),
  phone: z.string().regex(/^\+?[1-9]\d{9,14}$/, 'Enter a valid phone number'),

  // Passwords
  password: z.string()
    .min(8, 'At least 8 characters')
    .regex(/[A-Z]/, 'Include an uppercase letter')
    .regex(/[a-z]/, 'Include a lowercase letter')
    .regex(/[0-9]/, 'Include a number'),

  // Numbers
  positiveInt: z.coerce.number().int('Must be a whole number').positive('Must be positive'),
  money: z.coerce.number().multipleOf(0.01).nonnegative('Must be zero or positive'),
  percentage: z.coerce.number().min(0, 'Min 0%').max(100, 'Max 100%'),

  // Dates
  futureDate: z.coerce.date().min(new Date(), 'Must be in the future'),
  pastDate: z.coerce.date().max(new Date(), 'Must be in the past'),

  // Files
  image: (maxMB = 5) => z.instanceof(File)
    .refine((f) => f.size <= maxMB * 1024 * 1024, `Must be under ${maxMB} MB`)
    .refine(
      (f) => ['image/jpeg', 'image/png', 'image/webp'].includes(f.type),
      'Only JPEG, PNG, and WebP allowed'
    ),
  pdf: (maxMB = 10) => z.instanceof(File)
    .refine((f) => f.size <= maxMB * 1024 * 1024, `Must be under ${maxMB} MB`)
    .refine((f) => f.type === 'application/pdf', 'Only PDF files allowed'),
};
```

## Error Display Components

### Field Error

```tsx
function FieldError({ name }: { name: string }) {
  const { formState: { errors } } = useFormContext();
  const error = errors[name];
  if (!error) return null;
  return (
    <p id={`${name}-error`} className="text-sm text-red-500 mt-1" role="alert">
      {error.message as string}
    </p>
  );
}
```

### Form Error Summary

```tsx
function ErrorSummary() {
  const { formState: { errors } } = useFormContext();
  const errorList = Object.entries(errors).filter(([key]) => key !== 'root');
  if (errorList.length === 0) return null;
  return (
    <div className="bg-red-50 border border-red-200 rounded-lg p-4 mb-4" role="alert">
      <h3 className="text-red-800 font-medium mb-2">
        Please fix {errorList.length} error{errorList.length > 1 ? 's' : ''}:
      </h3>
      <ul className="list-disc list-inside text-red-600 text-sm">
        {errorList.map(([name, error]) => (
          <li key={name}>
            <a href={`#${name}`} className="underline">
              {(error as any).message}
            </a>
          </li>
        ))}
      </ul>
    </div>
  );
}
```
