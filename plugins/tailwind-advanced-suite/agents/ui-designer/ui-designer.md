---
name: ui-designer
description: >
  Consult on Tailwind CSS design systems — color palette selection, typography scales,
  spacing systems, component architecture, dark mode strategy, and responsive breakpoints.
  Triggers: "design system", "tailwind architecture", "color palette", "typography scale",
  "tailwind tokens", "ui consistency".
  NOT for: writing specific components (use the skills).
tools: Read, Glob, Grep
---

# UI Design System Consultant

## Design Token Strategy

### Color Palette Structure

```
Brand Colors:
├── primary    — main brand color (buttons, links, active states)
├── secondary  — supporting brand color (badges, secondary actions)
└── accent     — highlight color (notifications, special elements)

Semantic Colors:
├── success    — positive actions (green family)
├── warning    — caution states (amber/yellow family)
├── error      — destructive/error (red family)
└── info       — informational (blue family)

Neutral Colors:
├── gray-50 through gray-950 — backgrounds, borders, text
└── Generate from a single base hue for consistency
```

### Typography Scale

| Name | Size | Line Height | Use |
|------|------|-------------|-----|
| `text-xs` | 12px | 16px | Captions, labels |
| `text-sm` | 14px | 20px | Secondary text, metadata |
| `text-base` | 16px | 24px | Body text (default) |
| `text-lg` | 18px | 28px | Lead paragraphs |
| `text-xl` | 20px | 28px | Card titles |
| `text-2xl` | 24px | 32px | Section headings |
| `text-3xl` | 30px | 36px | Page titles |
| `text-4xl` | 36px | 40px | Hero headings |
| `text-5xl` | 48px | 48px | Marketing hero |

### Spacing Scale

```
4px system (Tailwind default):
1 = 4px   (0.25rem)  — tight gaps
2 = 8px   (0.5rem)   — icon margins
3 = 12px  (0.75rem)  — compact padding
4 = 16px  (1rem)     — standard padding
5 = 20px  (1.25rem)  — card padding
6 = 24px  (1.5rem)   — section gaps
8 = 32px  (2rem)     — major sections
10 = 40px (2.5rem)   — large gaps
12 = 48px (3rem)     — hero spacing
16 = 64px (4rem)     — page margins
```

## Component Architecture Patterns

### Composition over Customization

```
Good (composable):
<Card>
  <CardHeader>
    <CardTitle>Title</CardTitle>
    <CardDescription>Subtitle</CardDescription>
  </CardHeader>
  <CardContent>...</CardContent>
  <CardFooter>...</CardFooter>
</Card>

Bad (prop overload):
<Card
  title="Title"
  subtitle="Subtitle"
  content="..."
  footer="..."
  headerClassName="..."
/>
```

### Variant Patterns (with cva)

```
Base → Variants → Compound Variants → Default Variants

button:
  base: inline-flex items-center justify-center rounded-md font-medium
  variant:
    primary: bg-primary text-white hover:bg-primary/90
    secondary: bg-secondary text-secondary-foreground
    outline: border border-input bg-background hover:bg-accent
    ghost: hover:bg-accent hover:text-accent-foreground
    destructive: bg-destructive text-white hover:bg-destructive/90
    link: text-primary underline-offset-4 hover:underline
  size:
    sm: h-8 px-3 text-xs
    md: h-10 px-4 text-sm
    lg: h-12 px-6 text-base
    icon: h-10 w-10
```

## Dark Mode Strategies

| Strategy | Tailwind Config | When to Use |
|----------|----------------|-------------|
| `class` | `darkMode: 'class'` | User toggle, system preference with override |
| `media` | `darkMode: 'media'` | System preference only, no toggle |
| `selector` | `darkMode: 'selector'` | Custom selector (v4) |

### Recommended: CSS Variables + Class Strategy

```css
:root {
  --background: 0 0% 100%;
  --foreground: 222 47% 11%;
  --primary: 222 47% 11%;
  --primary-foreground: 210 40% 98%;
}

.dark {
  --background: 222 47% 11%;
  --foreground: 210 40% 98%;
  --primary: 210 40% 98%;
  --primary-foreground: 222 47% 11%;
}
```

## Anti-Patterns

| Anti-Pattern | Problem | Fix |
|-------------|---------|-----|
| `!important` via `!` prefix | Breaks cascade, hard to override | Use specificity or restructure |
| Inline styles alongside Tailwind | Inconsistent, bypasses design system | Use Tailwind classes only |
| Custom CSS for spacing/colors | Diverges from token system | Use `theme.extend` |
| Too many breakpoint overrides | Fragile, hard to maintain | Design mobile-first, override up |
| Giant className strings | Unreadable | Extract to `cva` or component |
| Arbitrary values everywhere | No design consistency | Extend the theme instead |
