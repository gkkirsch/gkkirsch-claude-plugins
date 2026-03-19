---
name: color-contrast
description: >
  Check and fix color contrast issues for WCAG 2.2 compliance. Analyzes CSS
  color values, identifies contrast failures, suggests fixes with optimal
  accessible alternatives. Handles text contrast, UI component contrast,
  and non-text graphical objects.
  Triggers: "color contrast", "contrast check", "contrast ratio",
  "accessible colors", "fix contrast".
  NOT for: color palette design, brand guidelines.
version: 1.0.0
argument-hint: "[file or component to check]"
allowed-tools: Read, Grep, Glob, Bash, Write, Edit
---

# Color Contrast Checker

Analyze and fix color contrast issues for WCAG 2.2 AA compliance.

## WCAG Contrast Requirements

| Category | Ratio Required | Examples |
|----------|---------------|----------|
| Normal text (<18pt / <14pt bold) | ≥ 4.5:1 | Body copy, labels, form text |
| Large text (≥18pt / ≥14pt bold) | ≥ 3:1 | Headings, large buttons |
| UI components | ≥ 3:1 | Borders, icons, focus rings |
| Non-text graphics | ≥ 3:1 | Charts, graphs, diagrams |
| Incidental / decorative | None | Disabled controls, logos |

**Font size note**: 18pt = 24px, 14pt bold = 18.67px (at default browser settings).

## Step 1: Extract Color Pairs

### Find color definitions

```
# CSS color properties
Grep: "color:\|background-color:\|background:\|border-color:\|outline-color:" in **/*.{css,scss,less}

# Tailwind classes
Grep: "text-\|bg-\|border-\|ring-\|outline-" in **/*.{tsx,jsx,vue,html}

# CSS variables / tokens
Grep: "--color\|--text\|--bg\|--border\|--surface\|--primary\|--secondary" in **/*.css

# Inline styles
Grep: "style=.*color\|style=.*background" in **/*.{tsx,jsx,vue,html}
```

### Common failing pairs (memorize these)

| Foreground | Background | Ratio | Verdict |
|-----------|------------|-------|---------|
| `#999999` | `#ffffff` | 2.85:1 | ❌ Fails all |
| `#888888` | `#ffffff` | 3.54:1 | ❌ Fails normal text |
| `#777777` | `#ffffff` | 4.48:1 | ❌ Barely fails |
| `#767676` | `#ffffff` | 4.54:1 | ✅ Minimum pass (AA normal) |
| `#757575` | `#ffffff` | 4.60:1 | ✅ Passes AA normal |
| `#595959` | `#ffffff` | 7.01:1 | ✅ Passes AAA normal |
| `#ffffff` | `#0066cc` | 5.12:1 | ✅ Passes AA normal |
| `#ffffff` | `#0052a3` | 7.06:1 | ✅ Passes AAA normal |
| `#ffffff` | `#ff6600` | 3.00:1 | ❌ Fails normal text |
| `#000000` | `#ffff00` | 19.56:1 | ✅ Maximum contrast (yellow bg) |
| `#ffffff` | `#ff0000` | 4.00:1 | ❌ Fails normal text |
| `#ffffff` | `#cc0000` | 5.17:1 | ✅ Passes AA normal |

## Step 2: Calculate Contrast Ratios

### Contrast ratio formula

The contrast ratio between two colors is calculated as:

```
(L1 + 0.05) / (L2 + 0.05)

where L1 = relative luminance of the lighter color
      L2 = relative luminance of the darker color
```

### Relative luminance calculation

```javascript
function relativeLuminance(r, g, b) {
  // Convert 0-255 to 0-1
  let [rs, gs, bs] = [r / 255, g / 255, b / 255];

  // Apply gamma correction
  const linearize = (c) => c <= 0.03928 ? c / 12.92 : Math.pow((c + 0.055) / 1.055, 2.4);

  return 0.2126 * linearize(rs) + 0.7152 * linearize(gs) + 0.0722 * linearize(bs);
}

function contrastRatio(color1, color2) {
  const l1 = relativeLuminance(...color1);
  const l2 = relativeLuminance(...color2);
  const lighter = Math.max(l1, l2);
  const darker = Math.min(l1, l2);
  return (lighter + 0.05) / (darker + 0.05);
}
```

### Quick contrast check via Node.js

```bash
node -e "
const hex2rgb = (h) => [parseInt(h.slice(1,3),16), parseInt(h.slice(3,5),16), parseInt(h.slice(5,7),16)];
const lum = (r,g,b) => { const l = c => c<=0.03928?c/12.92:((c+0.055)/1.055)**2.4; return 0.2126*l(r/255)+0.7152*l(g/255)+0.0722*l(b/255); };
const cr = (a,b) => { const [l1,l2]=[lum(...a),lum(...b)].sort().reverse(); return (l1+0.05)/(l2+0.05); };
const fg = hex2rgb('$1'), bg = hex2rgb('$2');
const ratio = cr(fg,bg).toFixed(2);
console.log('Contrast ratio: ' + ratio + ':1');
console.log('Normal text (4.5:1): ' + (ratio >= 4.5 ? '✅ PASS' : '❌ FAIL'));
console.log('Large text (3.0:1): ' + (ratio >= 3.0 ? '✅ PASS' : '❌ FAIL'));
console.log('UI components (3.0:1): ' + (ratio >= 3.0 ? '✅ PASS' : '❌ FAIL'));
"
```

Replace `$1` with foreground hex and `$2` with background hex.

## Step 3: Find Accessible Alternatives

### Strategy: Darken or lighten to meet ratio

When a color fails contrast:
1. Keep the hue and saturation
2. Adjust lightness until the ratio passes
3. Prefer darker foreground over lighter background (more readable)
4. Stay as close to the original color as possible

### Common fix patterns

```
// Gray text on white background
#999 → #767676  (minimum AA pass for normal text)
#999 → #595959  (comfortable AAA pass)

// Light blue on white
#66b3ff → #0066cc  (AA pass)

// Orange on white
#ff6600 → #cc5200  (AA pass for large text)
#ff6600 → #8a3700  (AA pass for normal text)

// Green on white (status messages)
#00cc00 → #008000  (AA pass for normal text)

// Red on white (error messages)
#ff0000 → #cc0000  (AA pass for normal text)
#ff0000 → #a30000  (comfortable pass)
```

### Accessible color palette generator

For each brand color, generate accessible variants:

```
Given: Primary blue #0066cc on white (#ffffff)
  Ratio: 5.12:1 ✅ AA normal text

Need darker variant for small text:
  #004d99 → 7.50:1 ✅ AAA normal text

Need lighter variant for backgrounds:
  Blue background #0066cc with white text → 5.12:1 ✅
  Light blue background #e6f0ff with #004d99 text → 8.93:1 ✅
```

## Step 4: Specific Component Checks

### Focus Indicators

```
Focus indicator must have ≥ 3:1 contrast against:
1. The element's background
2. Adjacent colors
3. The unfocused state of the element

/* GOOD focus indicator */
:focus-visible {
  outline: 2px solid #005fcc;  /* 5.12:1 on white */
  outline-offset: 2px;
}

/* BAD — light blue barely visible */
:focus-visible {
  outline: 2px solid #99ccff;  /* 1.96:1 on white — fails */
}
```

### Placeholder Text

```
Placeholder text must meet 4.5:1 contrast (it IS text).

/* BAD */
input::placeholder { color: #ccc; }  /* 1.61:1 — fails */

/* GOOD */
input::placeholder { color: #767676; }  /* 4.54:1 — passes */

/* REMINDER: placeholder is NOT a substitute for <label> */
```

### Disabled States

Disabled controls are exempt from contrast requirements (incidental text). But the disabled state itself should be perceivable — users need to know the control exists.

### Links within Text

Links within paragraphs need to be distinguishable from surrounding text by more than just color:
- Underline (most reliable)
- Bold + 3:1 contrast difference between link and body text
- Underline on hover/focus + 3:1 contrast difference

### Dark Mode

When implementing dark mode, check ALL color pairs again:
```
Light mode: #333 on #fff → 12.63:1 ✅
Dark mode: #ddd on #1a1a1a → 12.48:1 ✅

Light mode: #0066cc on #fff → 5.12:1 ✅
Dark mode: #66b3ff on #1a1a1a → 7.81:1 ✅ (lighter blue needed for dark bg)
```

## Step 5: Tailwind CSS Contrast Guide

### Safe text colors on white background

| Tailwind Class | Hex | Ratio | AA Normal | AA Large |
|----------------|-----|-------|-----------|----------|
| `text-gray-500` | #6b7280 | 5.03:1 | ✅ | ✅ |
| `text-gray-600` | #4b5563 | 7.45:1 | ✅ | ✅ |
| `text-gray-700` | #374151 | 10.31:1 | ✅ | ✅ |
| `text-gray-800` | #1f2937 | 14.72:1 | ✅ | ✅ |
| `text-gray-900` | #111827 | 17.44:1 | ✅ | ✅ |
| `text-gray-400` | #9ca3af | 3.01:1 | ❌ | ✅ |
| `text-gray-300` | #d1d5db | 1.76:1 | ❌ | ❌ |

**Rule of thumb**: On white, use `text-gray-500` or darker for body text. `text-gray-400` only for large text or UI components.

### Safe text colors on dark backgrounds

| Tailwind Class | On `bg-gray-900` | Ratio |
|----------------|-------------------|-------|
| `text-gray-100` | #f3f4f6 | 15.99:1 ✅ |
| `text-gray-200` | #e5e7eb | 13.56:1 ✅ |
| `text-gray-300` | #d1d5db | 10.19:1 ✅ |
| `text-gray-400` | #9ca3af | 5.62:1 ✅ |
| `text-gray-500` | #6b7280 | 3.46:1 ⚠️ Large only |

## Step 6: Generate Report

For each failing color pair, report:

```markdown
### ❌ [Location]
- **Foreground**: #999999
- **Background**: #ffffff
- **Contrast ratio**: 2.85:1
- **Requirement**: 4.5:1 (AA normal text)
- **Fix**: Change foreground to #767676 (4.54:1) or #595959 (7.01:1)
- **In Tailwind**: Change `text-gray-400` to `text-gray-500`
```