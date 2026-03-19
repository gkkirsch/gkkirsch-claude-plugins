# Common Accessibility Fixes

Quick reference for the most frequently encountered accessibility issues and their solutions.

---

## Fix 1: Missing or Bad Alt Text

### The Problem
Images without alt text force screen readers to announce the file name or URL.

### The Fix

```html
<!-- Meaningful image: describe what it shows -->
<img src="team-photo.jpg" alt="Engineering team at the 2024 offsite in Austin" />

<!-- Product image: include identifying details -->
<img src="shoe-123.jpg" alt="Nike Air Max 90 in white and red, side view" />

<!-- Chart/diagram: summarize the data -->
<img src="revenue-chart.png" alt="Revenue grew 45% year-over-year, from $2M to $2.9M" />

<!-- Decorative image: empty alt -->
<img src="divider.png" alt="" />
<img src="background-pattern.svg" alt="" role="presentation" />

<!-- Icon with text: icon is decorative -->
<button>
  <img src="trash.svg" alt="" /> Delete
</button>

<!-- Icon WITHOUT text: alt is required -->
<button aria-label="Delete item">
  <img src="trash.svg" alt="" />
</button>
<!-- OR -->
<button>
  <img src="trash.svg" alt="Delete item" />
</button>
```

### React/SVG Icons

```tsx
// Icon component with a11y built in
function Icon({ name, label, decorative = false }) {
  const SvgIcon = icons[name];
  if (decorative) {
    return <SvgIcon aria-hidden="true" focusable="false" />;
  }
  return <SvgIcon role="img" aria-label={label} />;
}

// Usage
<Icon name="search" label="Search" />
<Icon name="divider" decorative />
```

### Lucide React Icons

```tsx
import { Trash2 } from 'lucide-react';

// With visible label: icon is decorative
<button>
  <Trash2 aria-hidden="true" size={16} /> Delete
</button>

// Icon-only button: needs label
<button aria-label="Delete item">
  <Trash2 aria-hidden="true" size={16} />
</button>
```

---

## Fix 2: Missing Form Labels

### The Problem
Screen readers can't tell users what a form field is for without a label.

### The Fix

```html
<!-- Method 1: Visible <label> with for/id (preferred) -->
<label for="email">Email address</label>
<input type="email" id="email" />

<!-- Method 2: Wrapping <label> -->
<label>
  Email address
  <input type="email" />
</label>

<!-- Method 3: aria-label (when visual label would clutter UI) -->
<input type="search" aria-label="Search products" />

<!-- Method 4: aria-labelledby (using existing visible text) -->
<h2 id="billing-title">Billing Address</h2>
<input aria-labelledby="billing-title address-label" />
<span id="address-label">Street</span>

<!-- WRONG: placeholder is NOT a label -->
<input type="email" placeholder="Enter email" />
<!-- Screen reader: "edit text" — no context! -->
```

### Required Fields

```html
<label for="name">
  Full Name <span aria-hidden="true">*</span>
</label>
<input type="text" id="name" required aria-required="true" />

<!-- Or with custom message -->
<label for="name">Full Name (required)</label>
<input type="text" id="name" required aria-required="true" />
```

### Error Messages

```html
<label for="email">Email</label>
<input
  type="email"
  id="email"
  aria-invalid="true"
  aria-describedby="email-error"
/>
<p id="email-error" role="alert">
  Please enter a valid email address (e.g., name@example.com)
</p>
```

### Help Text

```html
<label for="password">Password</label>
<input
  type="password"
  id="password"
  aria-describedby="password-help"
/>
<p id="password-help">Must be at least 8 characters with one number</p>
```

---

## Fix 3: Clickable Divs

### The Problem
`<div>` and `<span>` elements are not focusable, have no role, and don't respond to keyboard by default.

### The Fix

```tsx
// BAD
<div className="card" onClick={() => navigate('/detail')}>
  <h3>Product Name</h3>
  <p>Description...</p>
</div>

// GOOD: Use a button or link
<a href="/detail" className="card">
  <h3>Product Name</h3>
  <p>Description...</p>
</a>

// GOOD: If it triggers an action (not navigation)
<button type="button" className="card" onClick={handleClick}>
  <h3>Product Name</h3>
  <p>Description...</p>
</button>

// ACCEPTABLE: If you truly can't change the element
<div
  className="card"
  role="button"
  tabIndex={0}
  onClick={handleClick}
  onKeyDown={(e) => {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      handleClick();
    }
  }}
>
  <h3>Product Name</h3>
  <p>Description...</p>
</div>
```

---

## Fix 4: Missing Focus Indicators

### The Problem
`outline: none` or `outline: 0` in CSS resets removes the visible focus indicator, making keyboard navigation impossible to track.

### The Fix

```css
/* Remove the reset */
/* ❌ *:focus { outline: none; } */

/* Add custom focus-visible styles */
:focus-visible {
  outline: 2px solid #005fcc;
  outline-offset: 2px;
}

/* If you need to remove default for mouse users only */
:focus:not(:focus-visible) {
  outline: none;
}

/* Framework-specific: Tailwind CSS */
/* <button class="focus-visible:ring-2 focus-visible:ring-blue-600 focus-visible:ring-offset-2"> */
```

### Component Library Override

```css
/* Override component library's focus reset */
.btn:focus-visible,
.link:focus-visible,
.input:focus-visible {
  outline: 2px solid #005fcc;
  outline-offset: 2px;
  box-shadow: none; /* Remove library's box-shadow focus if present */
}
```

---

## Fix 5: Color-Only Information

### The Problem
Information conveyed only by color is invisible to colorblind users and screen readers.

### The Fix

```html
<!-- BAD: Status indicated by color only -->
<span style="color: green;">Active</span>
<span style="color: red;">Inactive</span>

<!-- GOOD: Color + text -->
<span style="color: green;">✓ Active</span>
<span style="color: red;">✗ Inactive</span>

<!-- GOOD: Color + icon + text -->
<span class="status-active">
  <CheckCircleIcon aria-hidden="true" /> Active
</span>

<!-- BAD: Required field indicated by red asterisk only -->
<label style="color: red;">* Email</label>

<!-- GOOD: Required indicated by text -->
<label>Email <span>(required)</span></label>

<!-- BAD: Form errors shown by red border only -->
<input style="border-color: red;" />

<!-- GOOD: Error with icon + message -->
<input aria-invalid="true" aria-describedby="error-msg" style="border-color: red;" />
<p id="error-msg" class="error">
  <ErrorIcon aria-hidden="true" /> Please enter a valid email
</p>
```

---

## Fix 6: Missing Skip Navigation

### The Fix

```html
<!-- First element in <body> -->
<a href="#main-content" class="skip-link">Skip to main content</a>

<header>
  <nav><!-- Long navigation --></nav>
</header>

<main id="main-content" tabindex="-1">
  <!-- Page content -->
</main>
```

```css
.skip-link {
  position: absolute;
  top: -100%;
  left: 16px;
  padding: 8px 16px;
  background: #000;
  color: #fff;
  text-decoration: none;
  z-index: 9999;
  border-radius: 0 0 4px 4px;
}

.skip-link:focus {
  top: 0;
}
```

---

## Fix 7: Missing Page Language

### The Fix

```html
<!-- Set the primary language -->
<html lang="en">

<!-- For specific language content -->
<p>The French word for hello is <span lang="fr">bonjour</span>.</p>

<!-- Common language codes -->
<!-- en, en-US, en-GB, es, fr, de, ja, zh, ko, ar, pt, ru, it -->
```

---

## Fix 8: Non-Descriptive Link Text

### The Fix

```html
<!-- BAD -->
<a href="/pricing">Click here</a>
<a href="/docs">Read more</a>
<a href="/blog/post-1">Learn more</a>

<!-- GOOD: Link text describes destination -->
<a href="/pricing">View pricing plans</a>
<a href="/docs">Read the documentation</a>
<a href="/blog/post-1">How to set up authentication</a>

<!-- If you must use generic text, add hidden context -->
<a href="/blog/post-1">
  Read more<span class="sr-only"> about setting up authentication</span>
</a>

<!-- For icon links -->
<a href="https://twitter.com/handle" aria-label="Follow us on Twitter">
  <TwitterIcon aria-hidden="true" />
</a>
```

### Tailwind sr-only class

```html
<a href="/pricing">
  Learn more<span class="sr-only"> about our pricing plans</span>
</a>
```

---

## Fix 9: Missing Heading Hierarchy

### The Fix

```html
<!-- BAD: Skipping levels -->
<h1>Page Title</h1>
<h3>Section Title</h3>  <!-- Skipped h2! -->
<h5>Subsection</h5>     <!-- Skipped h4! -->

<!-- GOOD: Logical hierarchy -->
<h1>Page Title</h1>
  <h2>Section Title</h2>
    <h3>Subsection</h3>
  <h2>Another Section</h2>
    <h3>Subsection</h3>

<!-- BAD: Multiple h1 tags -->
<h1>Logo</h1>           <!-- In header -->
<h1>Page Title</h1>     <!-- In content -->

<!-- GOOD: Single h1 -->
<header>
  <a href="/"><img src="logo.png" alt="Company Name" /></a>
</header>
<main>
  <h1>Page Title</h1>
</main>

<!-- BAD: Using headings for styling -->
<h4>This is just bold text, not a heading</h4>

<!-- GOOD: Use CSS for styling, headings for structure -->
<p class="font-bold text-lg">This is styled text</p>
```

---

## Fix 10: Dynamic Content Not Announced

### The Fix

```tsx
// BAD: Content changes but screen reader doesn't know
function CartCount({ count }) {
  return <span className="badge">{count}</span>;
}

// GOOD: aria-live announces changes
function CartCount({ count }) {
  return (
    <span className="badge" role="status" aria-live="polite">
      {count} items in cart
    </span>
  );
}

// GOOD: Loading states
function LoadingSpinner({ isLoading }) {
  return (
    <div role="status" aria-live="polite">
      {isLoading ? 'Loading...' : ''}
    </div>
  );
}

// GOOD: Form submission feedback
function SubmitButton({ status }) {
  return (
    <>
      <button type="submit">Save</button>
      <div role="status" aria-live="polite">
        {status === 'saving' && 'Saving changes...'}
        {status === 'saved' && 'Changes saved successfully.'}
      </div>
      <div role="alert" aria-live="assertive">
        {status === 'error' && 'Error: Could not save changes. Please try again.'}
      </div>
    </>
  );
}
```

**Critical**: The container with `aria-live` must exist in the DOM BEFORE the content changes. Don't conditionally render the container — conditionally render the content inside it.

---

## Fix 11: Inaccessible Data Tables

### The Fix

```html
<!-- BAD: No semantic table markup -->
<div class="table">
  <div class="row">
    <div class="cell">Name</div>
    <div class="cell">Email</div>
  </div>
  <div class="row">
    <div class="cell">John</div>
    <div class="cell">john@test.com</div>
  </div>
</div>

<!-- GOOD: Semantic table -->
<table>
  <caption>User Accounts</caption>
  <thead>
    <tr>
      <th scope="col">Name</th>
      <th scope="col">Email</th>
      <th scope="col">Role</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td>John Doe</td>
      <td>john@test.com</td>
      <td>Admin</td>
    </tr>
  </tbody>
</table>

<!-- For complex tables with row headers -->
<table>
  <caption>Quarterly Results</caption>
  <thead>
    <tr>
      <th scope="col">Region</th>
      <th scope="col">Q1</th>
      <th scope="col">Q2</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <th scope="row">North America</th>
      <td>$1.2M</td>
      <td>$1.4M</td>
    </tr>
    <tr>
      <th scope="row">Europe</th>
      <td>$800K</td>
      <td>$950K</td>
    </tr>
  </tbody>
</table>
```

---

## Fix 12: Missing Autocomplete Attributes

### The Fix

```html
<!-- Personal information -->
<input type="text" autocomplete="name" />          <!-- Full name -->
<input type="text" autocomplete="given-name" />     <!-- First name -->
<input type="text" autocomplete="family-name" />    <!-- Last name -->
<input type="email" autocomplete="email" />         <!-- Email -->
<input type="tel" autocomplete="tel" />             <!-- Phone -->

<!-- Address -->
<input type="text" autocomplete="street-address" />
<input type="text" autocomplete="address-line1" />
<input type="text" autocomplete="address-line2" />
<input type="text" autocomplete="address-level2" /> <!-- City -->
<input type="text" autocomplete="address-level1" /> <!-- State/Province -->
<input type="text" autocomplete="postal-code" />
<input type="text" autocomplete="country" />

<!-- Payment -->
<input type="text" autocomplete="cc-name" />        <!-- Cardholder name -->
<input type="text" autocomplete="cc-number" />      <!-- Card number -->
<input type="text" autocomplete="cc-exp" />         <!-- Expiration -->
<input type="text" autocomplete="cc-csc" />         <!-- CVV -->

<!-- Authentication -->
<input type="text" autocomplete="username" />
<input type="password" autocomplete="current-password" />
<input type="password" autocomplete="new-password" />
```

---

## Utility: Screen-Reader-Only CSS Class

```css
/* Hide visually but keep accessible to screen readers */
.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border-width: 0;
}

/* Show when focused (for skip links) */
.sr-only-focusable:focus {
  position: static;
  width: auto;
  height: auto;
  padding: inherit;
  margin: inherit;
  overflow: visible;
  clip: auto;
  white-space: inherit;
}
```

Tailwind includes `sr-only` and `not-sr-only` utilities by default.
