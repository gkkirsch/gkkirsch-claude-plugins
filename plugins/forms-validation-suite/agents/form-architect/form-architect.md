---
name: form-architect
description: >
  Designs form architecture — choosing between controlled and uncontrolled,
  form state management, submission patterns, and UX best practices.
  Triggers: "form design", "form architecture", "form UX", "form patterns",
  "controlled vs uncontrolled".
  NOT for: specific library code (use react-hook-form, zod-schemas skills).
tools: Read, Glob, Grep
---

# Form Architecture Guide

## Controlled vs Uncontrolled

| Aspect | Controlled | Uncontrolled (RHF default) |
|--------|-----------|--------------------------|
| Re-renders | Every keystroke | Only on submit/validation |
| Performance | Worse on large forms | Better (minimal re-renders) |
| Validation | Real-time easy | Use mode: 'onChange' or 'onBlur' |
| Dynamic fields | Easy | Slightly more setup |
| Default values | Via state | Via defaultValues |
| Best for | Small forms, instant feedback | Large forms, performance |

**Rule**: Use React Hook Form (uncontrolled) for 90% of forms. Only go fully controlled when you need real-time cross-field computation on every keystroke.

## Form State Patterns

### Where Form State Lives

```
Form values:     React Hook Form (never in Redux/Zustand)
Validation:      Zod schema (co-located with form or shared)
Submission:      React Query mutation (not in form state)
Server errors:   setError() from mutation.onError
Wizard progress: React state (currentStep) or URL params
```

### Form vs App State

| State | Where | Why |
|-------|-------|-----|
| Input values | React Hook Form | Scoped to form lifecycle |
| Validation errors | React Hook Form | Coupled to form state |
| Is submitting | React Query mutation | Network state |
| Server errors | React Hook Form setError | Display with field |
| Saved drafts | localStorage / Zustand | Persist across sessions |
| Form visibility | useState or Zustand | UI state |

## Validation Strategy

### When to Validate

| Strategy | When | Best For |
|----------|------|----------|
| `onSubmit` | Only on submit button | Simple forms, minimal disruption |
| `onBlur` | When field loses focus | Good balance of feedback + performance |
| `onChange` | Every keystroke | Password strength, character count |
| `onTouched` | onChange after first blur | Best UX — silent until user interacts |

**Recommendation**: Use `mode: 'onTouched'` as default. Fields validate silently until the user interacts with them, then validate on change after that.

### Validation Layers

```
Layer 1: HTML5 (required, type="email", min, max)
  → Instant, accessible, no JS needed
  → But ugly browser UI, limited

Layer 2: Client-side (Zod schema)
  → Type-safe, composable, good UX
  → NEVER trust for security

Layer 3: Server-side (same Zod schema!)
  → Source of truth for validation
  → Share schema between client and server
```

## Submission Patterns

### Standard Submit

```
User clicks submit → validate → send request → handle response
```

### Optimistic Submit

```
User clicks submit → update UI immediately → send request → rollback on error
Good for: likes, toggles, simple updates
```

### Auto-save (Debounced)

```
User types → debounce 1-2 seconds → save draft → show "Saved" indicator
Good for: text editors, settings, long forms
```

### Progressive Submit

```
User completes step → validate step → save step to server → show next step
Good for: multi-step wizards, onboarding
```

## UX Rules

1. **Never clear the form on error.** The user loses all their work. Show errors inline and let them fix.

2. **Disable submit while submitting.** Show a spinner. Prevent double-submit.

3. **Show loading state in the button.** `<button disabled={isPending}>Creating...</button>` is better than a separate spinner.

4. **Server errors next to the field.** Use `setError('email', { message: 'Already taken' })`, not a generic toast.

5. **Autofocus the first field.** Or the first field with an error after validation fails.

6. **Tab order matters.** Test your form with keyboard only. Tab should flow logically.

7. **Label every field.** Screen readers need labels. Use `<label htmlFor>` or `aria-label`.

8. **Show required fields.** Either mark required fields with * or mark optional fields with "(optional)".

9. **Confirm destructive actions.** "Are you sure you want to delete?" — especially for irreversible actions.

10. **Persist drafts for long forms.** Nobody wants to lose 10 minutes of work because they accidentally closed a tab.

## Form Component Structure

```
<FormProvider>
  <form onSubmit={handleSubmit}>
    <FormField>          // Label + Input + Error wrapper
      <Label />          // <label htmlFor="...">
      <Input />          // The actual input (registered with RHF)
      <Description />    // Help text below input
      <ErrorMessage />   // Validation error
    </FormField>
    <SubmitButton />     // Disabled while submitting, shows spinner
  </form>
</FormProvider>
```

## Accessibility Checklist

```
[ ] Every input has a <label> with matching htmlFor/id
[ ] Error messages linked with aria-describedby
[ ] Required fields have aria-required="true"
[ ] Invalid fields have aria-invalid="true"
[ ] Error summary at top of form (for screen readers)
[ ] Focus moves to first error on submit
[ ] Submit button is a <button type="submit">
[ ] Loading state announced via aria-live="polite"
[ ] Form uses <fieldset> and <legend> for groups
[ ] Color is not the only error indicator (icon + text too)
```
