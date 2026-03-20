---
name: react-hook-form
description: >
  React Hook Form — form registration, validation modes, Controller for
  UI libraries, dynamic fields with useFieldArray, and submission handling.
  Triggers: "react hook form", "useForm", "form validation", "register",
  "useFieldArray", "form submission".
  NOT for: Zod schema design (use zod-schemas), file uploads (use file-uploads).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# React Hook Form

## Quick Start

```bash
npm install react-hook-form @hookform/resolvers zod
```

## Basic Form

```tsx
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';

const schema = z.object({
  name: z.string().min(1, 'Name is required'),
  email: z.string().email('Invalid email address'),
  message: z.string().min(10, 'Message must be at least 10 characters'),
});

type FormValues = z.infer<typeof schema>;

function ContactForm() {
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
    reset,
  } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      name: '',
      email: '',
      message: '',
    },
  });

  const onSubmit = async (data: FormValues) => {
    await submitContactForm(data);
    reset();
  };

  return (
    <form onSubmit={handleSubmit(onSubmit)}>
      <div>
        <label htmlFor="name">Name</label>
        <input id="name" {...register('name')} />
        {errors.name && <p className="text-red-500">{errors.name.message}</p>}
      </div>

      <div>
        <label htmlFor="email">Email</label>
        <input id="email" type="email" {...register('email')} />
        {errors.email && <p className="text-red-500">{errors.email.message}</p>}
      </div>

      <div>
        <label htmlFor="message">Message</label>
        <textarea id="message" {...register('message')} />
        {errors.message && <p className="text-red-500">{errors.message.message}</p>}
      </div>

      <button type="submit" disabled={isSubmitting}>
        {isSubmitting ? 'Sending...' : 'Send'}
      </button>
    </form>
  );
}
```

## Validation Modes

```typescript
const form = useForm<FormValues>({
  resolver: zodResolver(schema),
  mode: 'onTouched',  // Best UX — validates after first interaction
  // mode: 'onBlur',   // Validates when field loses focus
  // mode: 'onChange',  // Validates every keystroke (can be slow)
  // mode: 'onSubmit',  // Only validates on submit (default)
  // mode: 'all',       // Validates on blur AND change
});
```

## Controller (For UI Libraries)

Use `Controller` for components that don't expose a `ref` (Radix, shadcn/ui, MUI, etc.):

```tsx
import { Controller, useForm } from 'react-hook-form';

function SettingsForm() {
  const { control, handleSubmit } = useForm<SettingsValues>({
    resolver: zodResolver(settingsSchema),
  });

  return (
    <form onSubmit={handleSubmit(onSubmit)}>
      {/* Select */}
      <Controller
        name="role"
        control={control}
        render={({ field, fieldState: { error } }) => (
          <div>
            <label>Role</label>
            <Select value={field.value} onValueChange={field.onChange}>
              <SelectTrigger>
                <SelectValue placeholder="Select role" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="user">User</SelectItem>
                <SelectItem value="admin">Admin</SelectItem>
              </SelectContent>
            </Select>
            {error && <p className="text-red-500">{error.message}</p>}
          </div>
        )}
      />

      {/* Switch/Toggle */}
      <Controller
        name="notifications"
        control={control}
        render={({ field }) => (
          <div className="flex items-center gap-2">
            <Switch checked={field.value} onCheckedChange={field.onChange} />
            <label>Enable notifications</label>
          </div>
        )}
      />

      {/* Date Picker */}
      <Controller
        name="birthday"
        control={control}
        render={({ field, fieldState: { error } }) => (
          <div>
            <label>Birthday</label>
            <DatePicker value={field.value} onChange={field.onChange} />
            {error && <p className="text-red-500">{error.message}</p>}
          </div>
        )}
      />
    </form>
  );
}
```

## useFieldArray (Dynamic Fields)

```tsx
import { useForm, useFieldArray } from 'react-hook-form';

const schema = z.object({
  items: z.array(z.object({
    name: z.string().min(1, 'Item name required'),
    quantity: z.number().min(1, 'At least 1'),
    price: z.number().min(0, 'Price must be positive'),
  })).min(1, 'At least one item required'),
});

function OrderForm() {
  const { control, register, handleSubmit } = useForm({
    resolver: zodResolver(schema),
    defaultValues: {
      items: [{ name: '', quantity: 1, price: 0 }],
    },
  });

  const { fields, append, remove, move } = useFieldArray({
    control,
    name: 'items',
  });

  return (
    <form onSubmit={handleSubmit(onSubmit)}>
      {fields.map((field, index) => (
        <div key={field.id} className="flex gap-2">
          <input {...register(`items.${index}.name`)} placeholder="Item name" />
          <input
            type="number"
            {...register(`items.${index}.quantity`, { valueAsNumber: true })}
          />
          <input
            type="number"
            step="0.01"
            {...register(`items.${index}.price`, { valueAsNumber: true })}
          />
          <button type="button" onClick={() => remove(index)}>Remove</button>
        </div>
      ))}

      <button type="button" onClick={() => append({ name: '', quantity: 1, price: 0 })}>
        Add Item
      </button>

      <button type="submit">Submit Order</button>
    </form>
  );
}
```

## Server Errors

```tsx
function SignupForm() {
  const form = useForm<SignupValues>({
    resolver: zodResolver(signupSchema),
  });

  const onSubmit = async (data: SignupValues) => {
    try {
      await api.auth.signup(data);
    } catch (error) {
      if (error.status === 409) {
        // Field-level server error
        form.setError('email', {
          type: 'server',
          message: 'This email is already registered',
        });
      } else {
        // Form-level error (shown at top)
        form.setError('root', {
          type: 'server',
          message: 'Something went wrong. Please try again.',
        });
      }
    }
  };

  return (
    <form onSubmit={form.handleSubmit(onSubmit)}>
      {form.formState.errors.root && (
        <div className="bg-red-50 p-3 rounded text-red-600">
          {form.formState.errors.root.message}
        </div>
      )}
      {/* ... fields ... */}
    </form>
  );
}
```

## Watch and Conditional Fields

```tsx
function CheckoutForm() {
  const { register, watch, control } = useForm<CheckoutValues>();

  const paymentMethod = watch('paymentMethod');
  const shippingAddress = watch('shippingAddress');

  return (
    <form>
      <select {...register('paymentMethod')}>
        <option value="card">Credit Card</option>
        <option value="paypal">PayPal</option>
        <option value="bank">Bank Transfer</option>
      </select>

      {/* Conditional fields based on payment method */}
      {paymentMethod === 'card' && (
        <>
          <input {...register('cardNumber')} placeholder="Card Number" />
          <input {...register('expiry')} placeholder="MM/YY" />
          <input {...register('cvv')} placeholder="CVV" />
        </>
      )}

      {paymentMethod === 'bank' && (
        <>
          <input {...register('routingNumber')} placeholder="Routing Number" />
          <input {...register('accountNumber')} placeholder="Account Number" />
        </>
      )}
    </form>
  );
}
```

## Form with React Query Mutation

```tsx
import { useForm } from 'react-hook-form';
import { useMutation, useQueryClient } from '@tanstack/react-query';

function CreatePostForm() {
  const queryClient = useQueryClient();
  const form = useForm<CreatePostValues>({
    resolver: zodResolver(createPostSchema),
  });

  const mutation = useMutation({
    mutationFn: (data: CreatePostValues) => api.posts.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['posts'] });
      form.reset();
      toast.success('Post created!');
    },
    onError: (error) => {
      form.setError('root', { message: error.message });
    },
  });

  return (
    <form onSubmit={form.handleSubmit((data) => mutation.mutate(data))}>
      {/* fields */}
      <button type="submit" disabled={mutation.isPending}>
        {mutation.isPending ? 'Creating...' : 'Create Post'}
      </button>
    </form>
  );
}
```

## Reusable Form Field Component

```tsx
// components/FormField.tsx
import { useFormContext, Controller } from 'react-hook-form';

interface FormFieldProps {
  name: string;
  label: string;
  type?: string;
  placeholder?: string;
  description?: string;
}

function FormField({ name, label, type = 'text', placeholder, description }: FormFieldProps) {
  const { register, formState: { errors } } = useFormContext();
  const error = errors[name];

  return (
    <div className="space-y-1">
      <label htmlFor={name} className="text-sm font-medium">
        {label}
      </label>
      <input
        id={name}
        type={type}
        placeholder={placeholder}
        className={cn('input', error && 'border-red-500')}
        aria-invalid={!!error}
        aria-describedby={error ? `${name}-error` : description ? `${name}-desc` : undefined}
        {...register(name, type === 'number' ? { valueAsNumber: true } : {})}
      />
      {description && !error && (
        <p id={`${name}-desc`} className="text-sm text-muted-foreground">{description}</p>
      )}
      {error && (
        <p id={`${name}-error`} className="text-sm text-red-500" role="alert">
          {error.message as string}
        </p>
      )}
    </div>
  );
}
```

## Gotchas

1. **`register` returns a ref.** Don't also add your own ref to the same element. Use `ref` callback merging if needed.

2. **`defaultValues` must match schema shape.** Missing fields cause `undefined` which may not validate correctly.

3. **Numbers need `valueAsNumber: true`.** HTML inputs always return strings. Add `{ valueAsNumber: true }` to `register` for number inputs.

4. **`handleSubmit` prevents default.** Don't add `e.preventDefault()` yourself — `handleSubmit` does it.

5. **`reset()` resets to `defaultValues`.** Not to empty strings. Pass new values to `reset()` to override.

6. **Form state is a ref.** `formState` properties like `isDirty`, `isValid` are only tracked when you destructure them. Destructure what you need at the top.
