---
name: design-tokens
description: >
  Design token systems for scalable, consistent UI theming.
  Use when setting up CSS custom properties, theme systems, dark mode,
  color scales, typography scales, or spacing systems.
  Triggers: "design tokens", "CSS variables", "theme", "dark mode",
  "color scale", "typography scale", "spacing system", "theming".
  NOT for: individual component styling, Tailwind utility classes, or CSS layouts.
version: 1.0.0
allowed-tools: Read, Grep, Glob, Bash
---

# Design Token Systems

## CSS Custom Properties Foundation

```css
/* tokens.css — the single source of truth */

:root {
  /* === Color Primitives (raw palette, never use directly in components) === */
  --color-gray-50: #f9fafb;
  --color-gray-100: #f3f4f6;
  --color-gray-200: #e5e7eb;
  --color-gray-300: #d1d5db;
  --color-gray-400: #9ca3af;
  --color-gray-500: #6b7280;
  --color-gray-600: #4b5563;
  --color-gray-700: #374151;
  --color-gray-800: #1f2937;
  --color-gray-900: #111827;
  --color-gray-950: #030712;

  --color-blue-500: #3b82f6;
  --color-blue-600: #2563eb;
  --color-blue-700: #1d4ed8;

  --color-green-500: #22c55e;
  --color-red-500: #ef4444;
  --color-amber-500: #f59e0b;

  /* === Semantic Tokens (USE THESE in components) === */

  /* Backgrounds */
  --bg-primary: var(--color-gray-950);
  --bg-secondary: var(--color-gray-900);
  --bg-tertiary: var(--color-gray-800);
  --bg-elevated: var(--color-gray-800);
  --bg-overlay: rgb(0 0 0 / 0.5);

  /* Text */
  --text-primary: var(--color-gray-50);
  --text-secondary: var(--color-gray-400);
  --text-tertiary: var(--color-gray-500);
  --text-inverse: var(--color-gray-950);

  /* Borders */
  --border-primary: var(--color-gray-700);
  --border-secondary: var(--color-gray-800);
  --border-focus: var(--color-blue-500);

  /* Interactive */
  --interactive-primary: var(--color-blue-600);
  --interactive-primary-hover: var(--color-blue-700);
  --interactive-danger: var(--color-red-500);
  --interactive-success: var(--color-green-500);
  --interactive-warning: var(--color-amber-500);

  /* === Typography Scale === */
  --font-sans: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
  --font-mono: 'JetBrains Mono', 'Fira Code', 'Consolas', monospace;

  --text-xs: 0.75rem;    /* 12px */
  --text-sm: 0.875rem;   /* 14px */
  --text-base: 1rem;     /* 16px */
  --text-lg: 1.125rem;   /* 18px */
  --text-xl: 1.25rem;    /* 20px */
  --text-2xl: 1.5rem;    /* 24px */
  --text-3xl: 1.875rem;  /* 30px */
  --text-4xl: 2.25rem;   /* 36px */

  --leading-tight: 1.25;
  --leading-normal: 1.5;
  --leading-relaxed: 1.75;

  --font-normal: 400;
  --font-medium: 500;
  --font-semibold: 600;
  --font-bold: 700;

  /* === Spacing Scale (4px base) === */
  --space-0: 0;
  --space-1: 0.25rem;   /* 4px */
  --space-2: 0.5rem;    /* 8px */
  --space-3: 0.75rem;   /* 12px */
  --space-4: 1rem;      /* 16px */
  --space-5: 1.25rem;   /* 20px */
  --space-6: 1.5rem;    /* 24px */
  --space-8: 2rem;      /* 32px */
  --space-10: 2.5rem;   /* 40px */
  --space-12: 3rem;     /* 48px */
  --space-16: 4rem;     /* 64px */
  --space-20: 5rem;     /* 80px */

  /* === Border Radius === */
  --radius-sm: 0.25rem;  /* 4px */
  --radius-md: 0.375rem; /* 6px */
  --radius-lg: 0.5rem;   /* 8px */
  --radius-xl: 0.75rem;  /* 12px */
  --radius-full: 9999px;

  /* === Shadows === */
  --shadow-sm: 0 1px 2px 0 rgb(0 0 0 / 0.25);
  --shadow-md: 0 4px 6px -1px rgb(0 0 0 / 0.3), 0 2px 4px -2px rgb(0 0 0 / 0.3);
  --shadow-lg: 0 10px 15px -3px rgb(0 0 0 / 0.4), 0 4px 6px -4px rgb(0 0 0 / 0.4);
  --shadow-xl: 0 20px 25px -5px rgb(0 0 0 / 0.5), 0 8px 10px -6px rgb(0 0 0 / 0.5);

  /* === Transitions === */
  --duration-fast: 100ms;
  --duration-normal: 200ms;
  --duration-slow: 300ms;
  --ease-default: cubic-bezier(0.4, 0, 0.2, 1);
  --ease-in: cubic-bezier(0.4, 0, 1, 1);
  --ease-out: cubic-bezier(0, 0, 0.2, 1);

  /* === Z-Index Scale === */
  --z-dropdown: 50;
  --z-sticky: 100;
  --z-overlay: 200;
  --z-modal: 300;
  --z-popover: 400;
  --z-toast: 500;
}
```

## Dark/Light Theme Toggle

```css
/* Light theme overrides (dark is default above) */
[data-theme="light"] {
  --bg-primary: var(--color-gray-50);
  --bg-secondary: white;
  --bg-tertiary: var(--color-gray-100);
  --bg-elevated: white;

  --text-primary: var(--color-gray-900);
  --text-secondary: var(--color-gray-600);
  --text-tertiary: var(--color-gray-500);

  --border-primary: var(--color-gray-200);
  --border-secondary: var(--color-gray-100);

  --shadow-sm: 0 1px 2px 0 rgb(0 0 0 / 0.05);
  --shadow-md: 0 4px 6px -1px rgb(0 0 0 / 0.1);
}

/* System preference auto-detect */
@media (prefers-color-scheme: light) {
  :root:not([data-theme]) {
    --bg-primary: var(--color-gray-50);
    --bg-secondary: white;
    /* ... same overrides as [data-theme="light"] */
  }
}
```

```typescript
// Theme toggle with system preference detection
function useTheme() {
  const [theme, setTheme] = useState<'light' | 'dark' | 'system'>(() => {
    if (typeof window === 'undefined') return 'system';
    return (localStorage.getItem('theme') as 'light' | 'dark' | 'system') ?? 'system';
  });

  useEffect(() => {
    const root = document.documentElement;

    if (theme === 'system') {
      root.removeAttribute('data-theme');
    } else {
      root.setAttribute('data-theme', theme);
    }

    localStorage.setItem('theme', theme);
  }, [theme]);

  const toggle = () => {
    setTheme(t => t === 'dark' ? 'light' : t === 'light' ? 'system' : 'dark');
  };

  return { theme, setTheme, toggle };
}
```

## Tailwind v4 Integration

```css
/* app.css — Tailwind v4 with design tokens */
@import "tailwindcss";

@theme {
  /* Map tokens to Tailwind utilities */
  --color-bg-primary: var(--bg-primary);
  --color-bg-secondary: var(--bg-secondary);
  --color-bg-tertiary: var(--bg-tertiary);

  --color-text-primary: var(--text-primary);
  --color-text-secondary: var(--text-secondary);

  --color-border-primary: var(--border-primary);

  --color-interactive: var(--interactive-primary);
  --color-interactive-hover: var(--interactive-primary-hover);
  --color-danger: var(--interactive-danger);
  --color-success: var(--interactive-success);

  --font-family-sans: var(--font-sans);
  --font-family-mono: var(--font-mono);

  --radius-default: var(--radius-md);
}

/* Usage in components:
   bg-bg-primary text-text-primary border-border-primary
   bg-interactive hover:bg-interactive-hover
   font-sans font-mono
   rounded-default
*/
```

## Component Token Scoping

```css
/* Component-level token overrides */
.btn {
  --btn-bg: var(--interactive-primary);
  --btn-bg-hover: var(--interactive-primary-hover);
  --btn-text: white;
  --btn-radius: var(--radius-md);
  --btn-padding-x: var(--space-4);
  --btn-padding-y: var(--space-2);

  background: var(--btn-bg);
  color: var(--btn-text);
  border-radius: var(--btn-radius);
  padding: var(--btn-padding-y) var(--btn-padding-x);
  transition: background var(--duration-fast) var(--ease-default);
}

.btn:hover { background: var(--btn-bg-hover); }

/* Variants override component tokens, not raw values */
.btn--danger {
  --btn-bg: var(--interactive-danger);
  --btn-bg-hover: #dc2626; /* darken */
}

.btn--ghost {
  --btn-bg: transparent;
  --btn-bg-hover: var(--bg-tertiary);
  --btn-text: var(--text-secondary);
}

.btn--sm {
  --btn-padding-x: var(--space-3);
  --btn-padding-y: var(--space-1);
  font-size: var(--text-sm);
}
```

## Responsive Token Overrides

```css
/* Adjust spacing scale for mobile */
@media (max-width: 640px) {
  :root {
    --space-8: 1.5rem;    /* 24px instead of 32px */
    --space-10: 2rem;     /* 32px instead of 40px */
    --space-12: 2.5rem;   /* 40px instead of 48px */
    --space-16: 3rem;     /* 48px instead of 64px */

    --text-3xl: 1.5rem;   /* 24px instead of 30px */
    --text-4xl: 1.875rem; /* 30px instead of 36px */
  }
}
```

## Gotchas

1. **Don't use primitive tokens in components** — `var(--color-gray-700)` in a component means it won't respond to theme changes. Always use semantic tokens: `var(--border-primary)` not `var(--color-gray-700)`. Primitives are for defining semantics only.

2. **CSS custom properties are not statically analyzable** — unlike Sass variables, CSS custom properties can't be tree-shaken or validated at build time. Typos like `var(--bg-primry)` silently resolve to nothing. Use a linter (stylelint-value-no-unknown-custom-properties) to catch these.

3. **Custom property fallback values cascade differently** — `var(--missing, red)` uses the fallback only if the property is literally not defined, NOT if it's defined as `initial` or empty string. Test undefined vs empty states.

4. **Performance with hundreds of custom properties** — each custom property on `:root` is inherited by every element in the DOM. For very large token sets (500+), scope tokens to the components that use them rather than declaring everything on `:root`.

5. **Dark mode flash on page load** — if theme preference is stored in localStorage and applied via JavaScript, there's a flash of the default theme before JS runs. Fix with a blocking `<script>` in `<head>` that sets `data-theme` before any rendering occurs.

6. **HSL is better than hex for programmatic color manipulation** — with HSL tokens (`--color-blue-h: 217; --color-blue-s: 91%; --color-blue-l: 60%`), you can derive hover states, disabled states, and opacity variants with `calc()` instead of defining separate tokens for each variant.
