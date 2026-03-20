---
name: tailwind-design-system
description: >
  Tailwind CSS v4 design system architecture — @theme directive, CSS-first configuration,
  design tokens, custom properties, color system design, typography tokens, and spacing scales.
  Triggers: "tailwind v4", "tailwind design tokens", "tailwind @theme", "tailwind css config",
  "tailwind custom properties", "tailwind color system", "tailwind typography tokens".
  NOT for: basic utilities (use tailwind-fundamentals), theming with v3 config (use custom-themes).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Tailwind v4 Design System Architecture

## @theme Directive (v4 CSS-First Config)

```css
/* app.css — replaces tailwind.config.js */
@import "tailwindcss";

@theme {
  /* ---- Colors ---- */
  --color-brand-50: oklch(0.97 0.01 250);
  --color-brand-100: oklch(0.93 0.03 250);
  --color-brand-200: oklch(0.87 0.06 250);
  --color-brand-300: oklch(0.78 0.10 250);
  --color-brand-400: oklch(0.68 0.15 250);
  --color-brand-500: oklch(0.55 0.20 250);
  --color-brand-600: oklch(0.48 0.20 250);
  --color-brand-700: oklch(0.40 0.18 250);
  --color-brand-800: oklch(0.33 0.14 250);
  --color-brand-900: oklch(0.25 0.10 250);
  --color-brand-950: oklch(0.18 0.07 250);

  /* Semantic tokens referencing CSS variables */
  --color-surface: var(--surface);
  --color-on-surface: var(--on-surface);
  --color-primary: var(--primary);
  --color-on-primary: var(--on-primary);

  /* ---- Typography ---- */
  --font-sans: 'Inter Variable', 'Inter', system-ui, sans-serif;
  --font-display: 'Cal Sans', 'Inter', sans-serif;
  --font-mono: 'JetBrains Mono', 'Fira Code', monospace;

  --text-2xs: 0.625rem;
  --text-2xs--line-height: 1rem;

  --tracking-tighter: -0.04em;
  --tracking-tight: -0.02em;
  --tracking-normal: 0;
  --tracking-wide: 0.02em;

  /* ---- Spacing ---- */
  --spacing-4_5: 1.125rem;
  --spacing-18: 4.5rem;
  --spacing-88: 22rem;

  /* ---- Breakpoints ---- */
  --breakpoint-xs: 475px;
  --breakpoint-3xl: 1920px;

  /* ---- Shadows ---- */
  --shadow-soft: 0 1px 3px 0 oklch(0 0 0 / 0.04), 0 1px 2px -1px oklch(0 0 0 / 0.04);
  --shadow-elevated: 0 4px 6px -1px oklch(0 0 0 / 0.07), 0 2px 4px -2px oklch(0 0 0 / 0.05);
  --shadow-floating: 0 10px 15px -3px oklch(0 0 0 / 0.08), 0 4px 6px -4px oklch(0 0 0 / 0.04);
  --shadow-overlay: 0 20px 25px -5px oklch(0 0 0 / 0.1), 0 8px 10px -6px oklch(0 0 0 / 0.06);

  /* ---- Border Radius ---- */
  --radius-4xl: 2rem;
  --radius-5xl: 2.5rem;

  /* ---- Animations ---- */
  --animate-fade-in: fade-in 0.3s ease-out;
  --animate-fade-out: fade-out 0.2s ease-in;
  --animate-slide-up: slide-up 0.3s ease-out;
  --animate-slide-down: slide-down 0.3s ease-out;
  --animate-scale-in: scale-in 0.2s ease-out;
  --animate-accordion-down: accordion-down 0.2s ease-out;
  --animate-accordion-up: accordion-up 0.15s ease-in;

  /* ---- Easing ---- */
  --ease-spring: cubic-bezier(0.34, 1.56, 0.64, 1);
  --ease-bounce: cubic-bezier(0.68, -0.55, 0.27, 1.55);
}

/* ---- Keyframes ---- */
@keyframes fade-in {
  from { opacity: 0; }
  to { opacity: 1; }
}

@keyframes fade-out {
  from { opacity: 1; }
  to { opacity: 0; }
}

@keyframes slide-up {
  from { opacity: 0; transform: translateY(8px); }
  to { opacity: 1; transform: translateY(0); }
}

@keyframes slide-down {
  from { opacity: 0; transform: translateY(-8px); }
  to { opacity: 1; transform: translateY(0); }
}

@keyframes scale-in {
  from { opacity: 0; transform: scale(0.95); }
  to { opacity: 1; transform: scale(1); }
}

@keyframes accordion-down {
  from { height: 0; }
  to { height: var(--radix-accordion-content-height); }
}

@keyframes accordion-up {
  from { height: var(--radix-accordion-content-height); }
  to { height: 0; }
}
```

## OKLCH Color System

```css
/*
  OKLCH: perceptually uniform color space
  oklch(lightness chroma hue)

  lightness: 0 (black) → 1 (white)
  chroma: 0 (gray) → ~0.4 (most vivid)
  hue: 0-360 degrees (red=25, orange=70, yellow=90, green=145,
       cyan=195, blue=250, purple=300, pink=350)
*/

@theme {
  /* Generate a full palette from a single hue */
  /* Hue 250 = blue */
  --color-blue-50: oklch(0.97 0.01 250);   /* near white, tiny tint */
  --color-blue-100: oklch(0.93 0.03 250);  /* very light */
  --color-blue-200: oklch(0.87 0.06 250);  /* light */
  --color-blue-300: oklch(0.78 0.10 250);  /* medium-light */
  --color-blue-400: oklch(0.68 0.15 250);  /* medium */
  --color-blue-500: oklch(0.55 0.20 250);  /* base */
  --color-blue-600: oklch(0.48 0.20 250);  /* medium-dark */
  --color-blue-700: oklch(0.40 0.18 250);  /* dark */
  --color-blue-800: oklch(0.33 0.14 250);  /* very dark */
  --color-blue-900: oklch(0.25 0.10 250);  /* near black */
  --color-blue-950: oklch(0.18 0.07 250);  /* deepest */

  /*
    Pattern for consistent palettes:
    - Lightness: 0.97 → 0.18 (evenly distributed)
    - Chroma: low at extremes, peaks at 400-600
    - Hue: constant (or slight rotation for richness)
  */
}
```

## Semantic Token Layer

```css
/* tokens.css — semantic layer on top of primitives */

@layer base {
  :root {
    /* Surface hierarchy */
    --surface: oklch(1 0 0);          /* white */
    --surface-raised: oklch(0.98 0 0); /* slightly off-white */
    --surface-overlay: oklch(1 0 0);   /* modals, popovers */
    --surface-sunken: oklch(0.96 0 0); /* inputs, code blocks */

    /* Text hierarchy */
    --on-surface: oklch(0.15 0.01 250);     /* primary text */
    --on-surface-muted: oklch(0.45 0.01 250); /* secondary text */
    --on-surface-subtle: oklch(0.65 0.01 250); /* placeholder, disabled */

    /* Interactive */
    --primary: oklch(0.55 0.20 250);
    --on-primary: oklch(1 0 0);
    --primary-hover: oklch(0.48 0.20 250);
    --primary-active: oklch(0.42 0.18 250);

    /* Feedback */
    --success: oklch(0.55 0.15 145);
    --warning: oklch(0.70 0.15 85);
    --error: oklch(0.55 0.20 25);
    --info: oklch(0.55 0.15 250);

    /* Borders */
    --border: oklch(0.90 0.01 250);
    --border-strong: oklch(0.80 0.02 250);
    --ring: oklch(0.55 0.20 250);

    /* Spacing tokens */
    --space-page-x: 1rem;
    --space-section-y: 3rem;
    --space-card-padding: 1.5rem;
  }

  .dark {
    --surface: oklch(0.15 0.01 250);
    --surface-raised: oklch(0.20 0.01 250);
    --surface-overlay: oklch(0.22 0.02 250);
    --surface-sunken: oklch(0.12 0.01 250);

    --on-surface: oklch(0.93 0.01 250);
    --on-surface-muted: oklch(0.65 0.01 250);
    --on-surface-subtle: oklch(0.45 0.01 250);

    --primary: oklch(0.68 0.18 250);
    --on-primary: oklch(0.15 0.01 250);
    --primary-hover: oklch(0.73 0.16 250);
    --primary-active: oklch(0.63 0.20 250);

    --success: oklch(0.65 0.15 145);
    --warning: oklch(0.75 0.15 85);
    --error: oklch(0.65 0.18 25);
    --info: oklch(0.65 0.15 250);

    --border: oklch(0.25 0.01 250);
    --border-strong: oklch(0.35 0.02 250);
    --ring: oklch(0.68 0.18 250);
  }

  /* Responsive spacing */
  @media (min-width: 768px) {
    :root {
      --space-page-x: 2rem;
      --space-section-y: 5rem;
      --space-card-padding: 2rem;
    }
  }

  @media (min-width: 1280px) {
    :root {
      --space-page-x: 4rem;
      --space-section-y: 8rem;
    }
  }
}
```

## Typography System

```css
@import "tailwindcss";

@theme {
  /* Variable font with optical sizing */
  --font-sans: 'Inter Variable', system-ui, sans-serif;

  /* Type scale (major third ratio: 1.25) */
  --text-xs: 0.75rem;     /* 12px */
  --text-sm: 0.875rem;    /* 14px */
  --text-base: 1rem;      /* 16px */
  --text-lg: 1.125rem;    /* 18px */
  --text-xl: 1.25rem;     /* 20px */
  --text-2xl: 1.5rem;     /* 24px */
  --text-3xl: 1.875rem;   /* 30px */
  --text-4xl: 2.25rem;    /* 36px */
  --text-5xl: 3rem;       /* 48px */
  --text-6xl: 3.75rem;    /* 60px */

  /* Line heights paired to sizes */
  --text-xs--line-height: 1rem;
  --text-sm--line-height: 1.25rem;
  --text-base--line-height: 1.5rem;
  --text-lg--line-height: 1.75rem;
  --text-xl--line-height: 1.75rem;
  --text-2xl--line-height: 2rem;
  --text-3xl--line-height: 2.25rem;
  --text-4xl--line-height: 2.5rem;
  --text-5xl--line-height: 1;
  --text-6xl--line-height: 1;
}
```

```typescript
// Fluid typography utility (CSS clamp)
// Usage: text-fluid-xl → scales from text-lg to text-2xl

// In a Tailwind v4 plugin:
@utility text-fluid-sm {
  font-size: clamp(0.75rem, 0.7rem + 0.25vw, 0.875rem);
}

@utility text-fluid-base {
  font-size: clamp(0.875rem, 0.8rem + 0.375vw, 1rem);
}

@utility text-fluid-lg {
  font-size: clamp(1rem, 0.9rem + 0.5vw, 1.25rem);
}

@utility text-fluid-xl {
  font-size: clamp(1.125rem, 0.95rem + 0.875vw, 1.5rem);
}

@utility text-fluid-2xl {
  font-size: clamp(1.25rem, 0.95rem + 1.5vw, 2.25rem);
}

@utility text-fluid-3xl {
  font-size: clamp(1.5rem, 1rem + 2.5vw, 3rem);
}

@utility text-fluid-hero {
  font-size: clamp(2.25rem, 1.2rem + 5.25vw, 4.5rem);
}
```

## Design Token Migration (v3 → v4)

```
v3 tailwind.config.js          →  v4 @theme { }
─────────────────────────────────────────────────
theme.colors.brand.500          →  --color-brand-500
theme.extend.fontFamily.sans    →  --font-sans
theme.extend.spacing['18']      →  --spacing-18
theme.extend.borderRadius['4xl'] →  --radius-4xl
theme.extend.boxShadow.soft    →  --shadow-soft
theme.extend.animation['fade']  →  --animate-fade
theme.screens.xl               →  --breakpoint-xl
plugins: [require('@tw/forms')] →  @plugin "@tailwindcss/forms"
darkMode: 'class'              →  (default behavior in v4)
content: ['./src/**/*.tsx']    →  (automatic detection in v4)
```

```css
/* v4: Import plugins via CSS */
@plugin "@tailwindcss/forms";
@plugin "@tailwindcss/typography";
@plugin "@tailwindcss/container-queries";

/* v4: Source detection override (rarely needed) */
@source "../components/**/*.tsx";
@source inline("btn-primary btn-secondary badge-success");
```

## Multi-Theme Architecture

```css
/* themes.css — multiple theme definitions */
@layer base {
  :root {
    --surface: oklch(1 0 0);
    --on-surface: oklch(0.15 0 0);
    --primary: oklch(0.55 0.20 250);
    --radius-base: 0.5rem;
  }

  .dark {
    --surface: oklch(0.13 0.01 250);
    --on-surface: oklch(0.93 0 0);
    --primary: oklch(0.68 0.18 250);
  }

  .theme-warm {
    --surface: oklch(0.98 0.01 60);
    --on-surface: oklch(0.20 0.02 50);
    --primary: oklch(0.55 0.18 40);
    --radius-base: 0.75rem;
  }

  .theme-brutalist {
    --surface: oklch(1 0 0);
    --on-surface: oklch(0 0 0);
    --primary: oklch(0 0 0);
    --radius-base: 0;
  }
}
```

```typescript
// Theme switching with system preference detection
type Theme = 'light' | 'dark' | 'warm' | 'brutalist' | 'system';

function applyTheme(theme: Theme) {
  const root = document.documentElement;
  const classes = ['dark', 'theme-warm', 'theme-brutalist'];
  root.classList.remove(...classes);

  if (theme === 'system') {
    const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
    if (prefersDark) root.classList.add('dark');
  } else if (theme === 'dark') {
    root.classList.add('dark');
  } else if (theme !== 'light') {
    root.classList.add(`theme-${theme}`);
  }

  localStorage.setItem('theme', theme);
}

// Listen for system preference changes
window.matchMedia('(prefers-color-scheme: dark)')
  .addEventListener('change', (e) => {
    if (localStorage.getItem('theme') === 'system') {
      document.documentElement.classList.toggle('dark', e.matches);
    }
  });
```

## Gotchas

1. **@theme values must be valid CSS.** Unlike tailwind.config.js, you can't use JavaScript functions or computed values. Use CSS calc() and color functions instead.

2. **OKLCH chroma varies by hue.** Blue (hue ~250) supports higher chroma (~0.25) than yellow (hue ~90, max ~0.15). Don't blindly copy chroma values across hues or you'll get out-of-gamut colors.

3. **@theme replaces, @theme extend adds.** Using `@theme { --color-blue-500: ... }` replaces ALL default blue colors. Use `@theme extend { }` to keep defaults and add your tokens alongside them.

4. **CSS variable colors and opacity.** `oklch(var(--primary))` breaks opacity modifiers like `bg-primary/50`. Instead, define the full value: `--color-primary: oklch(0.55 0.20 250)` in @theme, and `--primary: 0.55 0.20 250` as a raw variable in `:root`.

5. **Font loading affects CLS.** Always pair custom fonts with `font-display: swap` and provide accurate `size-adjust` values. Use `@font-face` with `font-display: optional` for non-critical fonts.

6. **Design tokens need documentation.** A token system without documentation becomes tribal knowledge. Create a Storybook page or style guide showing every token with its visual output.
