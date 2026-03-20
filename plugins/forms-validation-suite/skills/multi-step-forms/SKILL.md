---
name: multi-step-forms
description: >
  Multi-step form wizard patterns — step validation, progress tracking,
  data persistence, conditional steps, and navigation.
  Triggers: "multi-step form", "wizard", "stepper", "form steps",
  "onboarding flow", "checkout flow", "form wizard".
  NOT for: single-page forms (use react-hook-form), file uploads (use file-uploads).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Multi-Step Forms

## Architecture Decision

| Approach | When | Pros | Cons |
|----------|------|------|------|
| Single `useForm`, render steps | Simple wizards (3-5 steps) | Shared state, easy back/forward | All validation schemas loaded upfront |
| Separate `useForm` per step | Complex flows, independent steps | Step-level validation, lazy loading | Must merge data manually |
| URL-based steps (route per step) | Checkout, onboarding | Deep-linkable, browser back works | More boilerplate, state in URL/context |

**Recommendation**: Single `useForm` with per-step validation for most cases.

## Single useForm Wizard

```tsx
import { useForm, FormProvider } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useState } from 'react';

// Per-step schemas
const stepSchemas = [
  z.object({
    firstName: z.string().min(1, 'First name is required'),
    lastName: z.string().min(1, 'Last name is required'),
    email: z.string().email('Invalid email address'),
  }),
  z.object({
    company: z.string().min(1, 'Company is required'),
    role: z.enum(['developer', 'designer', 'manager', 'other']),
    teamSize: z.coerce.number().min(1).max(10000),
  }),
  z.object({
    plan: z.enum(['free', 'pro', 'enterprise']),
    agreeToTerms: z.literal(true, {
      errorMap: () => ({ message: 'You must agree to the terms' }),
    }),
  }),
];

// Combined schema for final submission
const fullSchema = stepSchemas.reduce((acc, schema) => acc.merge(schema));
type FormValues = z.infer<typeof fullSchema>;

const STEPS = ['Personal Info', 'Company', 'Plan'] as const;

function OnboardingWizard() {
  const [step, setStep] = useState(0);

  const methods = useForm<FormValues>({
    resolver: zodResolver(fullSchema),
    defaultValues: {
      firstName: '',
      lastName: '',
      email: '',
      company: '',
      role: 'developer',
      teamSize: 1,
      plan: 'free',
      agreeToTerms: false as any,
    },
    mode: 'onTouched',
  });

  const next = async () => {
    // Validate only current step's fields
    const fields = Object.keys(stepSchemas[step].shape) as (keyof FormValues)[];
    const valid = await methods.trigger(fields);
    if (valid) setStep((s) => Math.min(s + 1, STEPS.length - 1));
  };

  const back = () => setStep((s) => Math.max(s - 1, 0));

  const onSubmit = async (data: FormValues) => {
    await createAccount(data);
  };

  return (
    <FormProvider {...methods}>
      <form onSubmit={methods.handleSubmit(onSubmit)}>
        {/* Progress bar */}
        <StepProgress steps={STEPS} currentStep={step} />

        {/* Step content */}
        {step === 0 && <PersonalInfoStep />}
        {step === 1 && <CompanyStep />}
        {step === 2 && <PlanStep />}

        {/* Navigation */}
        <div className="flex justify-between mt-6">
          <button type="button" onClick={back} disabled={step === 0}>
            Back
          </button>

          {step < STEPS.length - 1 ? (
            <button type="button" onClick={next}>
              Next
            </button>
          ) : (
            <button type="submit" disabled={methods.formState.isSubmitting}>
              {methods.formState.isSubmitting ? 'Creating...' : 'Create Account'}
            </button>
          )}
        </div>
      </form>
    </FormProvider>
  );
}
```

## Step Components

```tsx
import { useFormContext } from 'react-hook-form';

function PersonalInfoStep() {
  const { register, formState: { errors } } = useFormContext<FormValues>();

  return (
    <div className="space-y-4">
      <h2 className="text-lg font-semibold">Personal Information</h2>

      <div>
        <label htmlFor="firstName">First Name</label>
        <input id="firstName" {...register('firstName')} />
        {errors.firstName && <p className="text-red-500">{errors.firstName.message}</p>}
      </div>

      <div>
        <label htmlFor="lastName">Last Name</label>
        <input id="lastName" {...register('lastName')} />
        {errors.lastName && <p className="text-red-500">{errors.lastName.message}</p>}
      </div>

      <div>
        <label htmlFor="email">Email</label>
        <input id="email" type="email" {...register('email')} />
        {errors.email && <p className="text-red-500">{errors.email.message}</p>}
      </div>
    </div>
  );
}

function CompanyStep() {
  const { register, formState: { errors } } = useFormContext<FormValues>();

  return (
    <div className="space-y-4">
      <h2 className="text-lg font-semibold">Company Details</h2>

      <div>
        <label htmlFor="company">Company Name</label>
        <input id="company" {...register('company')} />
        {errors.company && <p className="text-red-500">{errors.company.message}</p>}
      </div>

      <div>
        <label htmlFor="role">Your Role</label>
        <select id="role" {...register('role')}>
          <option value="developer">Developer</option>
          <option value="designer">Designer</option>
          <option value="manager">Manager</option>
          <option value="other">Other</option>
        </select>
      </div>

      <div>
        <label htmlFor="teamSize">Team Size</label>
        <input
          id="teamSize"
          type="number"
          {...register('teamSize', { valueAsNumber: true })}
        />
        {errors.teamSize && <p className="text-red-500">{errors.teamSize.message}</p>}
      </div>
    </div>
  );
}
```

## Progress Indicator

```tsx
function StepProgress({ steps, currentStep }: { steps: readonly string[]; currentStep: number }) {
  return (
    <div className="flex items-center gap-2 mb-8">
      {steps.map((label, i) => (
        <div key={label} className="flex items-center gap-2">
          <div
            className={cn(
              'flex h-8 w-8 items-center justify-center rounded-full text-sm font-medium',
              i < currentStep && 'bg-green-500 text-white',
              i === currentStep && 'bg-blue-600 text-white',
              i > currentStep && 'bg-gray-200 text-gray-500',
            )}
          >
            {i < currentStep ? '✓' : i + 1}
          </div>
          <span
            className={cn(
              'text-sm hidden sm:block',
              i === currentStep ? 'font-medium' : 'text-gray-500',
            )}
          >
            {label}
          </span>
          {i < steps.length - 1 && (
            <div className={cn('h-px w-8', i < currentStep ? 'bg-green-500' : 'bg-gray-200')} />
          )}
        </div>
      ))}
    </div>
  );
}
```

## Data Persistence (Draft Saving)

```tsx
function useFormPersistence<T extends Record<string, any>>(
  key: string,
  methods: ReturnType<typeof useForm<T>>,
) {
  const { watch, reset } = methods;

  // Load saved data on mount
  useEffect(() => {
    const saved = localStorage.getItem(key);
    if (saved) {
      try {
        const parsed = JSON.parse(saved);
        reset(parsed, { keepDefaultValues: true });
      } catch {}
    }
  }, [key, reset]);

  // Auto-save on change (debounced)
  useEffect(() => {
    const subscription = watch((data) => {
      localStorage.setItem(key, JSON.stringify(data));
    });
    return () => subscription.unsubscribe();
  }, [key, watch]);

  const clearSaved = () => localStorage.removeItem(key);

  return { clearSaved };
}

// Usage
function OnboardingWizard() {
  const methods = useForm<FormValues>({ ... });
  const { clearSaved } = useFormPersistence('onboarding-draft', methods);

  const onSubmit = async (data: FormValues) => {
    await createAccount(data);
    clearSaved(); // Clear draft on successful submission
  };
}
```

## Conditional Steps

```tsx
function useConditionalSteps(watchValues: FormValues) {
  return useMemo(() => {
    const steps = [
      { id: 'personal', label: 'Personal Info', component: PersonalInfoStep },
      { id: 'company', label: 'Company', component: CompanyStep },
    ];

    // Only show billing step for paid plans
    if (watchValues.plan !== 'free') {
      steps.push({ id: 'billing', label: 'Billing', component: BillingStep });
    }

    // Only show team step for larger teams
    if (watchValues.teamSize > 1) {
      steps.push({ id: 'team', label: 'Team Members', component: TeamStep });
    }

    steps.push({ id: 'review', label: 'Review', component: ReviewStep });

    return steps;
  }, [watchValues.plan, watchValues.teamSize]);
}
```

## URL-Based Steps (React Router)

```tsx
import { Outlet, useNavigate, useLocation } from 'react-router-dom';

const WizardContext = createContext<{
  data: Partial<FormValues>;
  updateData: (step: Partial<FormValues>) => void;
}>({ data: {}, updateData: () => {} });

function WizardLayout() {
  const [data, setData] = useState<Partial<FormValues>>({});
  const location = useLocation();

  const steps = [
    { path: '/onboard/personal', label: 'Personal' },
    { path: '/onboard/company', label: 'Company' },
    { path: '/onboard/plan', label: 'Plan' },
  ];

  const currentStep = steps.findIndex((s) => s.path === location.pathname);

  return (
    <WizardContext.Provider
      value={{
        data,
        updateData: (step) => setData((prev) => ({ ...prev, ...step })),
      }}
    >
      <StepProgress steps={steps.map((s) => s.label)} currentStep={currentStep} />
      <Outlet />
    </WizardContext.Provider>
  );
}

function PersonalStep() {
  const { data, updateData } = useContext(WizardContext);
  const navigate = useNavigate();
  const methods = useForm({
    resolver: zodResolver(personalSchema),
    defaultValues: {
      firstName: data.firstName ?? '',
      lastName: data.lastName ?? '',
      email: data.email ?? '',
    },
  });

  const onSubmit = (values: PersonalValues) => {
    updateData(values);
    navigate('/onboard/company');
  };

  return (
    <form onSubmit={methods.handleSubmit(onSubmit)}>
      {/* fields */}
      <button type="submit">Next</button>
    </form>
  );
}
```

## Review Step

```tsx
function ReviewStep({ onEditStep }: { onEditStep: (step: number) => void }) {
  const { getValues } = useFormContext<FormValues>();
  const data = getValues();

  const sections = [
    {
      title: 'Personal Info',
      fields: [
        { label: 'Name', value: `${data.firstName} ${data.lastName}` },
        { label: 'Email', value: data.email },
      ],
      editStep: 0,
    },
    {
      title: 'Company',
      fields: [
        { label: 'Company', value: data.company },
        { label: 'Role', value: data.role },
        { label: 'Team Size', value: String(data.teamSize) },
      ],
      editStep: 1,
    },
  ];

  return (
    <div className="space-y-6">
      <h2 className="text-lg font-semibold">Review your information</h2>
      {sections.map((section) => (
        <div key={section.title} className="rounded-lg border p-4">
          <div className="flex justify-between items-center mb-2">
            <h3 className="font-medium">{section.title}</h3>
            <button
              type="button"
              onClick={() => onEditStep(section.editStep)}
              className="text-sm text-blue-600 hover:underline"
            >
              Edit
            </button>
          </div>
          <dl className="space-y-1">
            {section.fields.map(({ label, value }) => (
              <div key={label} className="flex gap-2 text-sm">
                <dt className="text-gray-500 w-24">{label}:</dt>
                <dd>{value}</dd>
              </div>
            ))}
          </dl>
        </div>
      ))}
    </div>
  );
}
```

## Reusable Wizard Hook

```tsx
function useWizard(totalSteps: number) {
  const [step, setStep] = useState(0);
  const [maxVisited, setMaxVisited] = useState(0);

  const next = () => {
    const nextStep = Math.min(step + 1, totalSteps - 1);
    setStep(nextStep);
    setMaxVisited((max) => Math.max(max, nextStep));
  };

  const back = () => setStep((s) => Math.max(s - 1, 0));

  const goTo = (target: number) => {
    if (target <= maxVisited) setStep(target);
  };

  return {
    step,
    next,
    back,
    goTo,
    isFirst: step === 0,
    isLast: step === totalSteps - 1,
    maxVisited,
    progress: ((step + 1) / totalSteps) * 100,
  };
}
```

## Gotchas

1. **`trigger()` with field names for per-step validation.** Don't validate the entire form on each step — only trigger the current step's fields. Pass an array of field names to `trigger()`.

2. **Back button should NOT validate.** Going back is a navigation action, not a submission. Don't run validation on `back()`.

3. **`useFormContext` requires `FormProvider`.** Step components using `useFormContext` must be wrapped in `<FormProvider {...methods}>`. Easy to forget when splitting steps into separate files.

4. **Conditional steps change total count.** If steps are dynamic (based on form values), the step index can become stale. Use the step `id` or `path` instead of numeric index for navigation state.

5. **Draft persistence needs cleanup.** Always clear saved drafts after successful submission. Otherwise users see stale data on their next visit.

6. **Animations between steps.** If adding CSS transitions, use `key={step}` on the step container to trigger mount/unmount animations. Consider `framer-motion`'s `AnimatePresence` for smooth direction-aware transitions.
