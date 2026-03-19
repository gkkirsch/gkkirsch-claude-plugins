# WCAG 2.1/2.2 Compliance Reference

Complete reference for Web Content Accessibility Guidelines (WCAG) 2.1 and 2.2 success criteria with practical implementation guidance for web developers.

## WCAG Overview

WCAG is organized around four principles (POUR):
1. **Perceivable** — Information must be presentable in ways users can perceive
2. **Operable** — Interface components must be operable by all users
3. **Understandable** — Information and UI operation must be understandable
4. **Robust** — Content must be robust enough for diverse user agents and assistive technologies

### Conformance Levels

| Level | Meaning | Required For |
|-------|---------|-------------|
| **A** | Minimum accessibility | All websites |
| **AA** | Standard accessibility (industry standard) | Most legal requirements, ADA, Section 508, EN 301 549 |
| **AAA** | Enhanced accessibility | Specialized audiences, government |

**Target AA for most projects.** A is insufficient; AAA is aspirational for most sites.

---

## Principle 1: Perceivable

### 1.1 Text Alternatives

#### 1.1.1 Non-text Content (Level A)

All non-text content must have a text alternative that serves the equivalent purpose.

| Content Type | Requirement | Example |
|---|---|---|
| Informative image | Descriptive alt text | `<img alt="Bar chart showing 40% growth in Q4">` |
| Decorative image | Empty alt | `<img alt="" role="presentation">` |
| Functional image (link/button) | Alt describes the function | `<img alt="Search">` (not "magnifying glass") |
| Complex image (chart/diagram) | Short alt + long description | `alt="Revenue chart" aria-describedby="chart-desc"` |
| Image of text | Text in alt matches image text | `<img alt="50% off summer sale">` |
| CAPTCHA | Alt describes purpose | `alt="Security verification image"` + audio alternative |
| Input image | Alt describes button function | `<input type="image" alt="Submit order">` |

```html
<!-- Informative image -->
<img src="team.jpg" alt="Five team members standing in front of the office building">

<!-- Decorative image -->
<img src="divider.svg" alt="">

<!-- Linked image -->
<a href="/"><img src="logo.svg" alt="ACME Corp homepage"></a>

<!-- Icon button -->
<button aria-label="Close dialog">
  <svg aria-hidden="true"><use href="#icon-x" /></svg>
</button>

<!-- Complex image with long description -->
<figure>
  <img src="org-chart.png" alt="Organization chart showing reporting structure" aria-describedby="org-desc">
  <figcaption id="org-desc">
    The CEO reports to the Board. Three VPs (Engineering, Sales, Operations) report to the CEO.
    Engineering has 4 teams: Frontend, Backend, Infrastructure, QA.
  </figcaption>
</figure>
```

### 1.2 Time-based Media

#### 1.2.1 Audio-only and Video-only (Prerecorded) (Level A)
- Audio-only: Provide a text transcript
- Video-only (no audio): Provide a text description or audio track

#### 1.2.2 Captions (Prerecorded) (Level A)
All prerecorded audio content in video must have synchronized captions.

```html
<video controls>
  <source src="tutorial.mp4" type="video/mp4">
  <track kind="captions" src="tutorial-en.vtt" srclang="en" label="English" default>
  <track kind="captions" src="tutorial-es.vtt" srclang="es" label="Español">
</video>
```

#### 1.2.3 Audio Description or Media Alternative (Level A)
Prerecorded video must have audio description or a full text alternative.

#### 1.2.4 Captions (Live) (Level AA)
Live audio content in video must have real-time captions.

#### 1.2.5 Audio Description (Prerecorded) (Level AA)
Prerecorded video must have audio description (narration of visual content during natural pauses).

```html
<video controls>
  <source src="tutorial.mp4" type="video/mp4">
  <track kind="captions" src="captions.vtt" srclang="en" label="English" default>
  <track kind="descriptions" src="descriptions.vtt" srclang="en" label="Audio descriptions">
</video>
```

### 1.3 Adaptable

#### 1.3.1 Info and Relationships (Level A)
Information, structure, and relationships conveyed visually must be programmatically determinable.

**Requirements:**
- Use semantic HTML (headings, lists, tables, landmarks)
- Don't rely on visual formatting alone to convey meaning
- Use proper table markup (th, scope, caption)
- Use fieldset/legend for related form controls

```html
<!-- BAD: Visual-only structure -->
<div style="font-size: 24px; font-weight: bold">Section Title</div>
<div style="margin-left: 20px">
  <div>• Item one</div>
  <div>• Item two</div>
</div>

<!-- GOOD: Semantic structure -->
<h2>Section Title</h2>
<ul>
  <li>Item one</li>
  <li>Item two</li>
</ul>

<!-- Required form grouping -->
<fieldset>
  <legend>Shipping Address</legend>
  <label for="street">Street</label>
  <input id="street" autocomplete="street-address">
  <label for="city">City</label>
  <input id="city" autocomplete="address-level2">
</fieldset>
```

#### 1.3.2 Meaningful Sequence (Level A)
When the sequence of content affects meaning, a correct reading sequence can be programmatically determined.

- DOM order should match visual order
- CSS should not reorder content in a way that changes meaning
- Use CSS `order` property carefully — screen readers follow DOM order

#### 1.3.3 Sensory Characteristics (Level A)
Instructions must not rely solely on shape, color, size, visual location, orientation, or sound.

```html
<!-- BAD: Relies on visual location -->
<p>Click the button on the right to continue.</p>

<!-- BAD: Relies on color -->
<p>Required fields are in red.</p>

<!-- GOOD: Non-sensory reference -->
<p>Click the "Continue" button to proceed.</p>
<p>Required fields are marked with an asterisk (*).</p>
```

#### 1.3.4 Orientation (Level AA) — WCAG 2.1
Content must not be restricted to a single display orientation (portrait/landscape) unless essential.

#### 1.3.5 Identify Input Purpose (Level AA) — WCAG 2.1
Input fields collecting user information must have programmatically determinable purpose via `autocomplete` attributes.

```html
<input type="text" autocomplete="given-name" name="firstName">
<input type="text" autocomplete="family-name" name="lastName">
<input type="email" autocomplete="email" name="email">
<input type="tel" autocomplete="tel" name="phone">
<input type="text" autocomplete="street-address" name="address">
<input type="text" autocomplete="postal-code" name="zip">
<input type="password" autocomplete="new-password" name="password">
<input type="password" autocomplete="current-password" name="loginPassword">
```

#### 1.3.6 Identify Purpose (Level AAA) — WCAG 2.1
The purpose of UI components, icons, and regions can be programmatically determined.

### 1.4 Distinguishable

#### 1.4.1 Use of Color (Level A)
Color must not be the only visual means of conveying information, indicating an action, prompting a response, or distinguishing a visual element.

```html
<!-- BAD: Error indicated only by color -->
<input style="border-color: red">

<!-- GOOD: Error with icon + text + color + aria-invalid -->
<input aria-invalid="true" aria-describedby="email-err">
<span id="email-err">
  <svg aria-hidden="true"><!-- error icon --></svg>
  Please enter a valid email address
</span>
```

#### 1.4.2 Audio Control (Level A)
If audio plays automatically for more than 3 seconds, provide a mechanism to pause/stop or control volume independently.

#### 1.4.3 Contrast (Minimum) (Level AA)
- Normal text: 4.5:1 contrast ratio minimum
- Large text (18pt/24px or 14pt/18.66px bold): 3:1 minimum

```
Common compliant color combinations (on white #FFFFFF):
- #595959 (gray) — 7.0:1 ✅ AA & AAA
- #767676 (gray) — 4.54:1 ✅ AA, ❌ AAA
- #0066CC (blue) — 5.26:1 ✅ AA
- #D93025 (red) — 4.61:1 ✅ AA for normal text

Common FAILING combinations (on white):
- #999999 — 2.85:1 ❌
- #AAAAAA — 2.32:1 ❌
- #CCCCCC — 1.61:1 ❌
- #FF6600 (orange) — 3.0:1 ❌ for normal text (OK for large text)
```

#### 1.4.4 Resize Text (Level AA)
Text can be resized up to 200% without loss of content or functionality. Use relative units (rem, em) not px for text.

#### 1.4.5 Images of Text (Level AA)
Don't use images of text when the same visual presentation can be achieved with real text. Exceptions: logos, text that is essential to the information being conveyed (like a screenshot).

#### 1.4.10 Reflow (Level AA) — WCAG 2.1
Content must reflow without requiring two-dimensional scrolling at 320px width (for vertical scrolling content) or 256px height (for horizontal scrolling content). Exceptions: data tables, toolbars, maps.

```css
/* Ensure content reflows at narrow widths */
body { overflow-x: hidden; }
img { max-width: 100%; height: auto; }
table { overflow-x: auto; display: block; }
pre { overflow-x: auto; white-space: pre-wrap; }
```

#### 1.4.11 Non-text Contrast (Level AA) — WCAG 2.1
UI components (borders, focus indicators) and meaningful graphics must have 3:1 contrast ratio against adjacent colors.

```css
/* Form field border must be 3:1 against background */
input {
  border: 1px solid #767676; /* 4.54:1 on white — passes */
}

/* Focus indicator must be 3:1 */
:focus-visible {
  outline: 2px solid #0066CC; /* clear contrast */
  outline-offset: 2px;
}

/* Icon contrast */
.icon {
  color: #595959; /* 7.0:1 on white */
}
```

#### 1.4.12 Text Spacing (Level AA) — WCAG 2.1
No loss of content or functionality when user overrides:
- Line height to 1.5× font size
- Paragraph spacing to 2× font size
- Letter spacing to 0.12× font size
- Word spacing to 0.16× font size

**Don't use fixed-height containers that clip text on spacing changes.**

#### 1.4.13 Content on Hover or Focus (Level AA) — WCAG 2.1
If additional content appears on hover/focus (tooltips, popovers):
- **Dismissible**: Can be dismissed without moving pointer/focus (Escape key)
- **Hoverable**: Pointer can move to the new content without it disappearing
- **Persistent**: Content remains visible until user dismisses, moves away, or it's no longer relevant

```css
/* GOOD: Tooltip pattern */
.tooltip-trigger:hover + .tooltip,
.tooltip-trigger:focus + .tooltip,
.tooltip:hover {  /* hoverable */
  display: block;
}
```

---

## Principle 2: Operable

### 2.1 Keyboard Accessible

#### 2.1.1 Keyboard (Level A)
All functionality must be operable through a keyboard interface (no timing for keystrokes).

**Every interactive element must be:**
- Focusable (via Tab key or arrow keys for composite widgets)
- Activatable (via Enter, Space, or appropriate key)
- Operable without requiring specific timings

```html
<!-- BAD: Not keyboard accessible -->
<div class="button" onclick="handleClick()">Click me</div>
<span onclick="toggleMenu()">Menu</span>

<!-- GOOD: Keyboard accessible -->
<button onclick="handleClick()">Click me</button>
<button onclick="toggleMenu()" aria-expanded="false">Menu</button>

<!-- If you must use div/span: add role, tabindex, and key handling -->
<div
  role="button"
  tabindex="0"
  onclick="handleClick()"
  onkeydown="if(event.key==='Enter'||event.key===' '){event.preventDefault();handleClick()}"
>
  Click me
</div>
```

#### 2.1.2 No Keyboard Trap (Level A)
If keyboard focus can be moved to a component, focus can be moved away from that component using only the keyboard. If it requires non-standard keys, the user is advised.

**Common trap locations:**
- Modals without Escape to close
- Embedded content (iframes, third-party widgets)
- Rich text editors

#### 2.1.4 Character Key Shortcuts (Level A) — WCAG 2.1
If a single character key shortcut exists (letters, numbers, punctuation), the user must be able to turn it off, remap it, or it only activates when the relevant component has focus.

### 2.2 Enough Time

#### 2.2.1 Timing Adjustable (Level A)
If there's a time limit, user must be able to: turn it off, adjust it (10× default), or extend it (warned 20 seconds before, given 10+ attempts to extend).

#### 2.2.2 Pause, Stop, Hide (Level A)
For moving, blinking, or scrolling content that starts automatically and lasts > 5 seconds: provide pause, stop, or hide mechanism. For auto-updating content: provide pause, stop, hide, or control update frequency.

```html
<!-- Carousel with pause -->
<div role="region" aria-roledescription="carousel" aria-label="Featured articles">
  <button aria-label="Pause carousel">⏸</button>
  <!-- slides -->
</div>

<!-- Auto-updating feed with controls -->
<div aria-live="polite" aria-atomic="false">
  <button>Pause updates</button>
  <!-- feed items -->
</div>
```

### 2.3 Seizures and Physical Reactions

#### 2.3.1 Three Flashes or Below Threshold (Level A)
No content flashes more than three times per second, unless the flash is below general flash and red flash thresholds.

### 2.4 Navigable

#### 2.4.1 Bypass Blocks (Level A)
Provide a mechanism to skip repeated blocks of content (skip navigation link).

```html
<body>
  <a href="#main-content" class="skip-link">Skip to main content</a>
  <header><!-- site header, nav --></header>
  <main id="main-content">
    <!-- page content -->
  </main>
</body>
```

```css
.skip-link {
  position: absolute;
  top: -40px;
  left: 0;
  background: #000;
  color: #fff;
  padding: 8px 16px;
  z-index: 100;
  transition: top 0.2s;
}
.skip-link:focus {
  top: 0;
}
```

#### 2.4.2 Page Titled (Level A)
Pages have descriptive titles that describe topic or purpose.

```html
<!-- BAD -->
<title>Page</title>
<title>Untitled</title>

<!-- GOOD -->
<title>Dashboard - MyApp</title>
<title>Edit Profile - Settings - MyApp</title>
<title>Search Results for "accessibility" - MyApp</title>
```

#### 2.4.3 Focus Order (Level A)
If content can be navigated sequentially, focus order must preserve meaning and operability. DOM order should match visual order.

#### 2.4.4 Link Purpose (In Context) (Level A)
The purpose of each link can be determined from the link text alone, or from the link text + surrounding context.

```html
<!-- BAD: Ambiguous -->
<a href="/pricing">Click here</a>
<a href="/article-123">Read more</a>

<!-- GOOD: Descriptive -->
<a href="/pricing">View pricing plans</a>
<a href="/article-123">Read more about web accessibility</a>

<!-- GOOD: Context via aria-describedby -->
<h3 id="a1-title">Understanding WCAG</h3>
<p>An overview of accessibility guidelines...</p>
<a href="/article-123" aria-describedby="a1-title">Read full article</a>
```

#### 2.4.5 Multiple Ways (Level AA)
More than one way to locate a page within a set of pages (navigation, search, site map, table of contents).

#### 2.4.6 Headings and Labels (Level AA)
Headings and labels describe topic or purpose.

#### 2.4.7 Focus Visible (Level AA)
Any keyboard-operable UI has a visible focus indicator.

```css
/* Minimum: 2px outline with 3:1 contrast */
:focus-visible {
  outline: 2px solid #0066CC;
  outline-offset: 2px;
}

/* NEVER remove focus styles globally */
/* *:focus { outline: none; } ← VIOLATION */
```

#### 2.4.11 Focus Not Obscured (Minimum) (Level AA) — WCAG 2.2
When a component receives keyboard focus, it is not entirely hidden by author-created content (sticky headers, cookie banners, chat widgets).

```css
/* Ensure sticky elements don't completely cover focused items */
.sticky-header {
  /* Use scroll-padding to account for sticky header */
}
html {
  scroll-padding-top: 80px; /* Height of sticky header */
}
```

#### 2.4.12 Focus Not Obscured (Enhanced) (Level AAA) — WCAG 2.2
No part of the focused component is hidden by author-created content.

#### 2.4.13 Focus Appearance (Level AAA) — WCAG 2.2
Focus indicator is at least 2px thick, with 3:1 contrast between focused and unfocused states, and the focus area is at least as large as a 2px perimeter.

### 2.5 Input Modalities

#### 2.5.1 Pointer Gestures (Level A) — WCAG 2.1
Functionality using multipoint or path-based gestures must have single-pointer alternatives.

```html
<!-- BAD: Pinch-to-zoom only -->
<!-- GOOD: Pinch-to-zoom + zoom buttons -->
<button aria-label="Zoom in">+</button>
<button aria-label="Zoom out">−</button>

<!-- BAD: Swipe-only carousel -->
<!-- GOOD: Swipe + previous/next buttons -->
<button aria-label="Previous slide">←</button>
<button aria-label="Next slide">→</button>
```

#### 2.5.2 Pointer Cancellation (Level A) — WCAG 2.1
For single-pointer actions: the down-event doesn't trigger the function (use click/up-event instead), or there's an abort/undo mechanism.

#### 2.5.3 Label in Name (Level A) — WCAG 2.1
If a component has a visible text label, the accessible name must contain that text.

```html
<!-- BAD: Accessible name doesn't match visible text -->
<button aria-label="Go to next page">Continue</button>

<!-- GOOD: Accessible name matches visible text -->
<button>Continue</button>
<!-- or -->
<button aria-label="Continue to payment">Continue</button>
```

#### 2.5.4 Motion Actuation (Level A) — WCAG 2.1
Functionality triggered by device motion (shake, tilt) must have UI alternatives and can be disabled.

#### 2.5.7 Dragging Movements (Level AA) — WCAG 2.2
Functionality requiring dragging must have a single-pointer alternative (buttons, menus).

```html
<!-- Drag-and-drop list with button alternatives -->
<ul>
  <li>
    <span>Item 1</span>
    <button aria-label="Move Item 1 up">↑</button>
    <button aria-label="Move Item 1 down">↓</button>
  </li>
</ul>
```

#### 2.5.8 Target Size (Minimum) (Level AA) — WCAG 2.2
Interactive targets must be at least 24×24 CSS pixels, OR have sufficient spacing from other targets. Exceptions: inline targets (in text), user-agent controlled, essential specific size.

```css
button, a, [role="button"], input, select {
  min-width: 24px;
  min-height: 24px;
}

/* Recommended: 44x44 for mobile */
@media (pointer: coarse) {
  button, a, [role="button"] {
    min-width: 44px;
    min-height: 44px;
  }
}
```

---

## Principle 3: Understandable

### 3.1 Readable

#### 3.1.1 Language of Page (Level A)
Default human language of each page must be programmatically determinable.

```html
<html lang="en">
<html lang="fr">
<html lang="ja">
```

#### 3.1.2 Language of Parts (Level AA)
Human language of each passage/phrase can be programmatically determined (when different from page language).

```html
<html lang="en">
  <body>
    <p>The French word for hello is <span lang="fr">bonjour</span>.</p>
    <blockquote lang="de">
      Die Würde des Menschen ist unantastbar.
    </blockquote>
  </body>
</html>
```

### 3.2 Predictable

#### 3.2.1 On Focus (Level A)
Receiving focus must not initiate a change of context (page navigation, form submission, significant UI change).

#### 3.2.2 On Input (Level A)
Changing a form setting must not automatically cause a change of context unless the user is warned beforehand.

```html
<!-- BAD: Auto-submit on select change -->
<select onchange="this.form.submit()">
  <option>Choose language</option>
  <option value="en">English</option>
  <option value="fr">French</option>
</select>

<!-- GOOD: Explicit submit -->
<label for="lang">Language</label>
<select id="lang" name="language">
  <option value="en">English</option>
  <option value="fr">French</option>
</select>
<button type="submit">Change language</button>
```

#### 3.2.3 Consistent Navigation (Level AA)
Navigation mechanisms repeated across pages must occur in the same relative order each time.

#### 3.2.4 Consistent Identification (Level AA)
Components with the same functionality must be identified consistently (same labels, icons, alt text across pages).

#### 3.2.6 Consistent Help (Level A) — WCAG 2.2
If a help mechanism (contact info, chat, FAQ link) appears on multiple pages, it must be in the same relative location on each page.

### 3.3 Input Assistance

#### 3.3.1 Error Identification (Level A)
If an input error is automatically detected, the error is identified and described to the user in text.

```html
<label for="email">Email</label>
<input id="email" type="email" aria-invalid="true" aria-describedby="email-err">
<p id="email-err" role="alert">Please enter a valid email address (e.g., name@example.com)</p>
```

#### 3.3.2 Labels or Instructions (Level A)
Labels or instructions are provided when content requires user input.

#### 3.3.3 Error Suggestion (Level AA)
If an input error is detected and suggestions for correction are known, provide them to the user.

```html
<!-- Suggest correction -->
<p id="date-err" role="alert">
  Invalid date format. Please use MM/DD/YYYY (e.g., 03/15/2025).
</p>
```

#### 3.3.4 Error Prevention (Legal, Financial, Data) (Level AA)
For pages that cause legal commitments or financial transactions:
- **Reversible**: Submissions can be reversed
- **Checked**: Data is checked for errors and user can correct them
- **Confirmed**: Mechanism to review, confirm, and correct before submission

```html
<!-- Order confirmation page -->
<h2>Review Your Order</h2>
<table><!-- order details --></table>
<p>Total: $99.99</p>
<form>
  <button type="button" onclick="goBack()">Edit Order</button>
  <button type="submit">Confirm and Place Order</button>
</form>
```

#### 3.3.7 Redundant Entry (Level A) — WCAG 2.2
Don't ask users to re-enter information they've already provided in the same process, unless re-entry is essential (password confirmation) or the previously entered information is no longer valid.

```html
<!-- BAD: Asking for shipping address then billing address from scratch -->
<!-- GOOD: "Same as shipping address" checkbox -->
<label>
  <input type="checkbox" id="same-address" checked>
  Billing address same as shipping address
</label>
```

#### 3.3.8 Accessible Authentication (Minimum) (Level AA) — WCAG 2.2
Don't require cognitive function tests (CAPTCHA, memorizing/transcribing passwords) for authentication unless an alternative is provided (password managers, copy-paste, biometric, or passkeys).

**Requirements:**
- Allow password managers to fill in credentials
- Don't block paste in password fields
- Provide alternative to CAPTCHA (audio, logic-based, passkey)
- Support passkeys/WebAuthn as alternative

```html
<!-- Allow password managers and paste -->
<input
  type="password"
  autocomplete="current-password"
  name="password"
>
<!-- Do NOT add oncopy/onpaste prevention -->
```

---

## Principle 4: Robust

### 4.1 Compatible

#### 4.1.2 Name, Role, Value (Level A)
For all UI components, name, role, and state must be programmatically determinable. State changes must be available to user agents (including assistive technologies).

```html
<!-- Custom toggle must expose role and state -->
<button
  role="switch"
  aria-checked="false"
  aria-label="Dark mode"
  onclick="toggle(this)"
>
  Dark Mode
</button>

<script>
function toggle(el) {
  const checked = el.getAttribute('aria-checked') === 'true';
  el.setAttribute('aria-checked', !checked);
}
</script>
```

#### 4.1.3 Status Messages (Level AA) — WCAG 2.1
Status messages can be programmatically determined through role or properties so they can be presented to the user by assistive technologies without receiving focus.

```html
<!-- Search results count -->
<div role="status" aria-live="polite">
  Showing 24 of 156 results
</div>

<!-- Form validation error summary -->
<div role="alert">
  2 errors found. Please correct the highlighted fields.
</div>

<!-- Loading indicator -->
<div role="status" aria-live="polite">
  Loading search results...
</div>

<!-- Cart update -->
<div role="status" aria-live="polite">
  Item added to cart. Cart total: 3 items.
</div>

<!-- Progress update -->
<div role="status" aria-live="polite">
  Upload 75% complete
</div>
```

---

## WCAG 2.2 New Criteria Summary

WCAG 2.2 added the following criteria (published October 2023):

| Criterion | Level | Summary |
|-----------|-------|---------|
| 2.4.11 Focus Not Obscured (Minimum) | AA | Focused element not completely hidden by sticky elements |
| 2.4.12 Focus Not Obscured (Enhanced) | AAA | No part of focused element hidden |
| 2.4.13 Focus Appearance | AAA | Focus indicator ≥ 2px thick, 3:1 contrast |
| 2.5.7 Dragging Movements | AA | Single-pointer alternative for drag operations |
| 2.5.8 Target Size (Minimum) | AA | 24×24px minimum (or sufficient spacing) |
| 3.2.6 Consistent Help | A | Help mechanisms in consistent location |
| 3.3.7 Redundant Entry | A | Don't ask for same info twice |
| 3.3.8 Accessible Authentication (Minimum) | AA | No cognitive function tests for login |
| 3.3.9 Accessible Authentication (Enhanced) | AAA | No cognitive function tests at all |

**Note:** WCAG 2.2 also removed 4.1.1 Parsing, as modern HTML parsers handle this automatically.

---

## Testing Checklist by Component Type

### Page-Level
```
□ <html lang> is set and correct
□ <title> is descriptive and unique
□ Skip navigation link present and functional
□ One <main> landmark
□ Heading hierarchy is correct (h1 → h2 → h3, no skips)
□ All landmark regions present (header, nav, main, footer)
□ Multiple navs have unique labels
□ Page is usable at 200% zoom
□ Content reflows at 320px width
□ Focus visible on all interactive elements
```

### Forms
```
□ Every input has a visible label
□ Labels are programmatically associated (for/id or wrapping)
□ Required fields indicated (not by color alone)
□ Error messages linked via aria-describedby
□ aria-invalid="true" on fields with errors
□ Error summary announced (role="alert" or aria-live)
□ Focus moves to first error on submit
□ autocomplete attributes on user data fields
□ No paste prevention on password fields
□ Form groupings use fieldset/legend
□ Submit button has descriptive text
```

### Images
```
□ Informative images have descriptive alt text
□ Decorative images have alt=""
□ Linked/button images describe the function
□ Complex images have long descriptions
□ SVGs have role="img" and title/aria-label
□ Background images with meaning have text alternative
□ No images of text (use real text)
```

### Navigation
```
□ All items keyboard accessible
□ Current page indicated (aria-current="page")
□ Focus visible on all items
□ Submenus keyboard operable (arrow keys)
□ Submenus have aria-expanded
□ Mobile menu has focus trap
□ Skip link bypasses navigation
```

### Dialogs/Modals
```
□ role="dialog" or role="alertdialog"
□ aria-modal="true"
□ aria-labelledby points to title
□ Focus moves to dialog on open
□ Focus trapped within dialog
□ Escape key closes dialog
□ Focus returns to trigger on close
□ Background content is inert (inert attribute or aria-hidden)
```

### Tables
```
□ <caption> describes the table
□ <th> with scope="col" or scope="row"
□ Complex tables use headers/id association
□ Layout tables have role="presentation" (or don't use tables for layout)
```

### Custom Widgets
```
□ Appropriate ARIA role assigned
□ All required ARIA properties set
□ States update when interaction occurs (aria-expanded, aria-checked, etc.)
□ Keyboard interaction follows ARIA Authoring Practices patterns
□ Focus management is correct (roving tabindex for composite widgets)
□ Live regions announce dynamic changes
```

---

## Common ARIA Roles Quick Reference

### Landmark Roles
| Role | HTML Equivalent | Usage |
|------|----------------|-------|
| banner | `<header>` (top-level) | Site header |
| navigation | `<nav>` | Navigation |
| main | `<main>` | Main content (one per page) |
| complementary | `<aside>` | Supporting content |
| contentinfo | `<footer>` (top-level) | Site footer |
| search | `<search>` | Search functionality |
| form | `<form>` (with accessible name) | Form landmark |
| region | `<section>` (with accessible name) | Generic landmark |

### Widget Roles
| Role | For | Required ARIA |
|------|-----|---------------|
| button | Clickable action | — |
| checkbox | Toggle option | aria-checked |
| combobox | Text input + listbox | aria-expanded, aria-controls |
| dialog | Modal dialog | aria-labelledby or aria-label |
| listbox | Selection list | — |
| menu | Action menu | — |
| menuitem | Item in menu | — |
| option | Item in listbox | aria-selected |
| progressbar | Progress indicator | aria-valuenow (or indeterminate) |
| radio | Single selection in group | aria-checked |
| slider | Range input | aria-valuenow, aria-valuemin, aria-valuemax |
| switch | On/off toggle | aria-checked |
| tab | Tab in tablist | aria-selected, aria-controls |
| tabpanel | Content for tab | aria-labelledby |
| tree | Hierarchical list | — |
| treeitem | Item in tree | aria-expanded (if has children) |

### Live Region Roles
| Role | Equivalent | When to Use |
|------|-----------|-------------|
| alert | aria-live="assertive" | Urgent errors, warnings |
| status | aria-live="polite" | Status updates, search results count |
| log | aria-live="polite" + aria-relevant="additions" | Chat messages, activity feed |
| timer | aria-live="off" | Countdown, stopwatch |

---

## Legal Landscape Quick Reference

| Regulation | Region | Requires | Standard |
|-----------|--------|----------|----------|
| ADA Title III | US | Accessible websites | WCAG 2.1 AA (de facto) |
| Section 508 | US (Federal) | Accessible ICT | WCAG 2.0 AA (EN 301 549) |
| EAA (European Accessibility Act) | EU | Accessible products/services | EN 301 549 → WCAG 2.1 AA |
| AODA | Ontario, Canada | Accessible websites | WCAG 2.0 AA |
| DDA | UK | Accessible services | WCAG 2.1 AA (recommended) |
| EN 301 549 | EU (public sector) | Accessible ICT | WCAG 2.1 AA |
