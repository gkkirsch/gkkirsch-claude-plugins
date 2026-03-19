# WCAG 2.2 AA Checklist

Complete checklist for WCAG 2.2 Level AA compliance. Each criterion includes its level (A or AA), a plain-English description, and what to check.

---

## Principle 1: Perceivable

### 1.1 Text Alternatives

| # | Criterion | Level | Check |
|---|-----------|-------|-------|
| 1.1.1 | Non-text Content | A | Every `<img>` has meaningful `alt`. Decorative images have `alt=""`. SVGs have `<title>` or `aria-label`. Icon buttons have accessible names. `<canvas>` has fallback. CAPTCHAs have alternatives. |

### 1.2 Time-Based Media

| # | Criterion | Level | Check |
|---|-----------|-------|-------|
| 1.2.1 | Audio-only / Video-only (Prerecorded) | A | Audio has transcript. Video-only has text description or audio track. |
| 1.2.2 | Captions (Prerecorded) | A | All prerecorded video with audio has synchronized captions. |
| 1.2.3 | Audio Description or Media Alternative | A | Video has audio description or full text alternative. |
| 1.2.4 | Captions (Live) | AA | Live video with audio has real-time captions. |
| 1.2.5 | Audio Description (Prerecorded) | AA | Prerecorded video has audio description track. |

### 1.3 Adaptable

| # | Criterion | Level | Check |
|---|-----------|-------|-------|
| 1.3.1 | Info and Relationships | A | Semantic HTML used. Headings, lists, tables have proper markup. Form inputs have labels. Landmark regions present. |
| 1.3.2 | Meaningful Sequence | A | Reading order (DOM order) matches visual order. CSS doesn't create confusing sequence. |
| 1.3.3 | Sensory Characteristics | A | Instructions don't rely solely on shape, color, size, position, or sound. ("Click the round button" → also label the button.) |
| 1.3.4 | Orientation | AA | Content not restricted to single orientation (portrait/landscape) unless essential. |
| 1.3.5 | Identify Input Purpose | AA | Form inputs with common purposes have `autocomplete` attributes. (name, email, address, tel, etc.) |

### 1.4 Distinguishable

| # | Criterion | Level | Check |
|---|-----------|-------|-------|
| 1.4.1 | Use of Color | A | Color is not the only visual means of conveying information. Error fields have icons/text, not just red border. Links have underline, not just color. |
| 1.4.2 | Audio Control | A | Auto-playing audio > 3 seconds has pause/stop/volume control. |
| 1.4.3 | Contrast (Minimum) | AA | Normal text: ≥ 4.5:1 contrast ratio. Large text (≥18pt / ≥14pt bold): ≥ 3:1. |
| 1.4.4 | Resize Text | AA | Text resizable to 200% without loss of content or function. |
| 1.4.5 | Images of Text | AA | Real text used instead of images of text (except logos). |
| 1.4.10 | Reflow | AA | Content viewable at 320px width without horizontal scrolling (except data tables, toolbars). |
| 1.4.11 | Non-text Contrast | AA | UI components and graphical objects: ≥ 3:1 contrast against adjacent colors. (Borders, icons, focus indicators, chart segments.) |
| 1.4.12 | Text Spacing | AA | No loss of content when user adjusts: line-height to 1.5x, paragraph spacing to 2x font size, letter spacing to 0.12em, word spacing to 0.16em. |
| 1.4.13 | Content on Hover or Focus | AA | Tooltips/popovers: dismissable (Escape), hoverable (mouse can move to tooltip), persistent (stays visible until dismissed). |

---

## Principle 2: Operable

### 2.1 Keyboard Accessible

| # | Criterion | Level | Check |
|---|-----------|-------|-------|
| 2.1.1 | Keyboard | A | All functionality available via keyboard. No mouse-only interactions. |
| 2.1.2 | No Keyboard Trap | A | Focus can always be moved away from any component using keyboard. |
| 2.1.4 | Character Key Shortcuts | A | Single-character keyboard shortcuts can be remapped, turned off, or only activate on focus. |

### 2.2 Enough Time

| # | Criterion | Level | Check |
|---|-----------|-------|-------|
| 2.2.1 | Timing Adjustable | A | Time limits can be turned off, adjusted, or extended (20-second warning with option to extend). |
| 2.2.2 | Pause, Stop, Hide | A | Moving, blinking, scrolling content can be paused. Auto-updating content can be paused. |

### 2.3 Seizures and Physical Reactions

| # | Criterion | Level | Check |
|---|-----------|-------|-------|
| 2.3.1 | Three Flashes or Below | A | No content flashes more than 3 times per second. |

### 2.4 Navigable

| # | Criterion | Level | Check |
|---|-----------|-------|-------|
| 2.4.1 | Bypass Blocks | A | Skip navigation link present. Landmarks used (`<main>`, `<nav>`, `<aside>`). |
| 2.4.2 | Page Titled | A | Each page has descriptive, unique `<title>`. |
| 2.4.3 | Focus Order | A | Tab order follows logical reading/use order. |
| 2.4.4 | Link Purpose (In Context) | A | Link text + surrounding context makes purpose clear. (Not "click here".) |
| 2.4.5 | Multiple Ways | AA | Two or more ways to find each page (nav, search, sitemap, A-Z index). |
| 2.4.6 | Headings and Labels | AA | Headings and form labels describe topic/purpose. |
| 2.4.7 | Focus Visible | AA | Keyboard focus indicator is visible on all interactive elements. |
| 2.4.11 | Focus Not Obscured (Minimum) | AA | Focused element is not entirely hidden by author-created content (sticky headers, modals, cookie banners). |

### 2.5 Input Modalities

| # | Criterion | Level | Check |
|---|-----------|-------|-------|
| 2.5.1 | Pointer Gestures | A | Multi-point gestures (pinch, multi-finger) have single-pointer alternative. |
| 2.5.2 | Pointer Cancellation | A | For down-event actions: up-event completes or aborts. Or: undo mechanism. |
| 2.5.3 | Label in Name | A | Visible label text is included in the accessible name. (Button says "Search" → accessible name includes "Search".) |
| 2.5.4 | Motion Actuation | A | Motion-triggered functions (shake, tilt) have UI alternative and can be disabled. |
| 2.5.7 | Dragging Movements | AA | Drag-and-drop has non-dragging alternative (buttons, menus). |
| 2.5.8 | Target Size (Minimum) | AA | Interactive targets ≥ 24×24 CSS pixels, OR have sufficient spacing, OR have alternative target. |

---

## Principle 3: Understandable

### 3.1 Readable

| # | Criterion | Level | Check |
|---|-----------|-------|-------|
| 3.1.1 | Language of Page | A | `<html lang="en">` (or appropriate language code) present. |
| 3.1.2 | Language of Parts | AA | Content in a different language has `lang` attribute. (`<span lang="fr">Bonjour</span>`) |

### 3.2 Predictable

| # | Criterion | Level | Check |
|---|-----------|-------|-------|
| 3.2.1 | On Focus | A | Receiving focus doesn't trigger unexpected context change (page navigation, form submission, dialog opening). |
| 3.2.2 | On Input | A | Changing a form input doesn't trigger unexpected context change unless user is warned beforehand. |
| 3.2.3 | Consistent Navigation | AA | Navigation menus are in same order across pages. |
| 3.2.4 | Consistent Identification | AA | Components with same function have same labels across pages. |

### 3.3 Input Assistance

| # | Criterion | Level | Check |
|---|-----------|-------|-------|
| 3.3.1 | Error Identification | A | Errors are identified in text and describe the specific error. Not just "invalid input". |
| 3.3.2 | Labels or Instructions | A | Form inputs have labels. Required fields indicated. Expected format shown. |
| 3.3.3 | Error Suggestion | AA | When errors are detected and suggestions known, provide them. ("Must be at least 8 characters.") |
| 3.3.4 | Error Prevention (Legal, Financial) | AA | Submissions with legal/financial commitments: reversible, checked for errors, or confirmed before final submit. |
| 3.3.7 | Redundant Entry | A | Info already provided in current session is auto-populated or available for selection (don't make users re-enter address, name, etc.). |
| 3.3.8 | Accessible Authentication (Minimum) | AA | Login doesn't require cognitive function test (CAPTCHA) unless alternative provided, or mechanism assisted (paste into password field, password manager support). |

---

## Principle 4: Robust

### 4.1 Compatible

| # | Criterion | Level | Check |
|---|-----------|-------|-------|
| 4.1.2 | Name, Role, Value | A | Custom widgets have correct ARIA roles, accessible names, and state values that update programmatically. |
| 4.1.3 | Status Messages | AA | Status messages (success, error, progress) use `aria-live`, `role="alert"`, or `role="status"` — not just visual changes. |

---

## Quick-Reference: Most Common Failures

These are the issues found most frequently in web accessibility audits:

| Rank | Issue | WCAG | Prevalence |
|------|-------|------|------------|
| 1 | Low contrast text | 1.4.3 | ~84% of sites |
| 2 | Missing alt text | 1.1.1 | ~58% of sites |
| 3 | Missing form labels | 1.3.1 | ~54% of sites |
| 4 | Empty links | 2.4.4 | ~49% of sites |
| 5 | Empty buttons | 4.1.2 | ~28% of sites |
| 6 | Missing document language | 3.1.1 | ~18% of sites |

Source: WebAIM Million annual study (percentages approximate).

---

## New in WCAG 2.2 (Added Criteria)

| # | Name | Level | What's New |
|---|------|-------|------------|
| 2.4.11 | Focus Not Obscured (Minimum) | AA | Focused element not entirely hidden by sticky headers, etc. |
| 2.4.13 | Focus Appearance | AAA | Focus indicator is ≥ 2px thick and 3:1 contrast (AAA only) |
| 2.5.7 | Dragging Movements | AA | Drag-and-drop has non-dragging alternative |
| 2.5.8 | Target Size (Minimum) | AA | Touch targets ≥ 24×24px |
| 3.2.6 | Consistent Help | A | Help mechanisms in same location across pages |
| 3.3.7 | Redundant Entry | A | Don't ask for same info twice in a session |
| 3.3.8 | Accessible Authentication (Minimum) | AA | No cognitive function tests for login |
| 3.3.9 | Accessible Authentication (Enhanced) | AAA | Stricter version of 3.3.8 |

**Removed from WCAG 2.2** (was in 2.1):
- 4.1.1 Parsing — removed because modern browsers handle malformed HTML gracefully.
