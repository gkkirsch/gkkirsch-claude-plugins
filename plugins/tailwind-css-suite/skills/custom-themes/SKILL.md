---
name: custom-themes
description: >
  Custom Tailwind CSS theming — tailwind.config.js customization, CSS variables,
  design tokens, multi-theme support, and shadcn/ui theming.
  Triggers: "tailwind theme", "tailwind config", "tailwind custom colors", "tailwind extend",
  "tailwind design tokens", "shadcn theme", "tailwind css variables", "tailwind fonts".
  NOT for: basic utilities (use tailwind-fundamentals), component building (use component-patterns).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Custom Tailwind Themes

## Tailwind Config (v3)

```javascript
// tailwind.config.js
/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ['./src/**/*.{js,ts,jsx,tsx}'],
  darkMode: 'class',
  theme: {
    // Override defaults entirely
    screens: {
      sm: '640px',
      md: '768px',
      lg: '1024px',
      xl: '1280px',
      '2xl': '1536px',
    },

    // Extend defaults (add without replacing)
    extend: {
      colors: {
        // Brand colors
        brand: {
          50: '#eff6ff',
          100: '#dbeafe',
          200: '#bfdbfe',
          300: '#93c5fd',
          400: '#60a5fa',
          500: '#3b82f6',
          600: '#2563eb',
          700: '#1d4ed8',
          800: '#1e40af',
          900: '#1e3a8a',
          950: '#172554',
        },
        // Semantic colors via CSS variables
        background: 'hsl(var(--background))',
        foreground: 'hsl(var(--foreground))',
        primary: {
          DEFAULT: 'hsl(var(--primary))',
          foreground: 'hsl(var(--primary-foreground))',
        },
        secondary: {
          DEFAULT: 'hsl(var(--secondary))',
          foreground: 'hsl(var(--secondary-foreground))',
        },
        muted: {
          DEFAULT: 'hsl(var(--muted))',
          foreground: 'hsl(var(--muted-foreground))',
        },
        accent: {
          DEFAULT: 'hsl(var(--accent))',
          foreground: 'hsl(var(--accent-foreground))',
        },
        destructive: {
          DEFAULT: 'hsl(var(--destructive))',
          foreground: 'hsl(var(--destructive-foreground))',
        },
        border: 'hsl(var(--border))',
        input: 'hsl(var(--input))',
        ring: 'hsl(var(--ring))',
      },
      fontFamily: {
        sans: ['Inter', 'system-ui', 'sans-serif'],
        display: ['Cal Sans', 'Inter', 'sans-serif'],
        mono: ['JetBrains Mono', 'monospace'],
      },
      fontSize: {
        '2xs': ['0.625rem', { lineHeight: '1rem' }],
      },
      spacing: {
        '18': '4.5rem',
        '88': '22rem',
        '128': '32rem',
      },
      borderRadius: {
        '4xl': '2rem',
      },
      boxShadow: {
        'soft': '0 2px 15px -3px rgba(0, 0, 0, 0.07), 0 10px 20px -2px rgba(0, 0, 0, 0.04)',
        'glow': '0 0 15px rgba(59, 130, 246, 0.5)',
      },
      animation: {
        'fade-in': 'fadeIn 0.5s ease-in-out',
        'slide-up': 'slideUp 0.3s ease-out',
        'slide-down': 'slideDown 0.3s ease-out',
      },
      keyframes: {
        fadeIn: {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' },
        },
        slideUp: {
          '0%': { transform: 'translateY(10px)', opacity: '0' },
          '100%': { transform: 'translateY(0)', opacity: '1' },
        },
        slideDown: {
          '0%': { transform: 'translateY(-10px)', opacity: '0' },
          '100%': { transform: 'translateY(0)', opacity: '1' },
        },
      },
    },
  },
  plugins: [
    require('@tailwindcss/forms'),
    require('@tailwindcss/typography'),
    require('@tailwindcss/container-queries'),
  ],
};
```

## CSS Variables Theme System

```css
/* src/styles/globals.css */
@tailwind base;
@tailwind components;
@tailwind utilities;

@layer base {
  :root {
    --background: 0 0% 100%;
    --foreground: 222.2 84% 4.9%;
    --primary: 221.2 83.2% 53.3%;
    --primary-foreground: 210 40% 98%;
    --secondary: 210 40% 96.1%;
    --secondary-foreground: 222.2 47.4% 11.2%;
    --muted: 210 40% 96.1%;
    --muted-foreground: 215.4 16.3% 46.9%;
    --accent: 210 40% 96.1%;
    --accent-foreground: 222.2 47.4% 11.2%;
    --destructive: 0 84.2% 60.2%;
    --destructive-foreground: 210 40% 98%;
    --border: 214.3 31.8% 91.4%;
    --input: 214.3 31.8% 91.4%;
    --ring: 221.2 83.2% 53.3%;
    --radius: 0.5rem;
  }

  .dark {
    --background: 222.2 84% 4.9%;
    --foreground: 210 40% 98%;
    --primary: 217.2 91.2% 59.8%;
    --primary-foreground: 222.2 47.4% 11.2%;
    --secondary: 217.2 32.6% 17.5%;
    --secondary-foreground: 210 40% 98%;
    --muted: 217.2 32.6% 17.5%;
    --muted-foreground: 215 20.2% 65.1%;
    --accent: 217.2 32.6% 17.5%;
    --accent-foreground: 210 40% 98%;
    --destructive: 0 62.8% 30.6%;
    --destructive-foreground: 210 40% 98%;
    --border: 217.2 32.6% 17.5%;
    --input: 217.2 32.6% 17.5%;
    --ring: 224.3 76.3% 48%;
  }
}
```

## Tailwind v4 Theme (CSS-first)

```css
/* Tailwind v4: CSS-only configuration */
@import "tailwindcss";

@theme {
  /* Colors */
  --color-brand-50: #eff6ff;
  --color-brand-500: #3b82f6;
  --color-brand-600: #2563eb;
  --color-brand-700: #1d4ed8;

  /* From CSS variables */
  --color-background: hsl(var(--background));
  --color-foreground: hsl(var(--foreground));
  --color-primary: hsl(var(--primary));

  /* Fonts */
  --font-sans: 'Inter', system-ui, sans-serif;
  --font-display: 'Cal Sans', 'Inter', sans-serif;
  --font-mono: 'JetBrains Mono', monospace;

  /* Spacing */
  --spacing-18: 4.5rem;

  /* Shadows */
  --shadow-soft: 0 2px 15px -3px rgba(0, 0, 0, 0.07);

  /* Animations */
  --animate-fade-in: fade-in 0.5s ease-in-out;

  /* Breakpoints */
  --breakpoint-sm: 640px;
  --breakpoint-md: 768px;
  --breakpoint-lg: 1024px;
  --breakpoint-xl: 1280px;
  --breakpoint-2xl: 1536px;
}

@keyframes fade-in {
  from { opacity: 0; }
  to { opacity: 1; }
}
```

## Custom Google Fonts Setup

```html
<!-- In <head> or layout -->
<link rel="preconnect" href="https://fonts.googleapis.com" />
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin />
<link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap" rel="stylesheet" />
```

```javascript
// tailwind.config.js
module.exports = {
  theme: {
    extend: {
      fontFamily: {
        sans: ['Inter', 'system-ui', '-apple-system', 'sans-serif'],
      },
    },
  },
};
```

```css
/* Apply globally */
@layer base {
  body {
    @apply font-sans antialiased;
  }
}
```

## shadcn/ui Theme Integration

```typescript
// lib/utils.ts — the cn() helper
import { type ClassValue, clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

// Usage in components
function Button({ className, variant, size, ...props }) {
  return (
    <button
      className={cn(
        buttonVariants({ variant, size }),
        className, // allows consumer overrides
      )}
      {...props}
    />
  );
}
```

```bash
# Install shadcn/ui
npx shadcn@latest init

# Add components
npx shadcn@latest add button
npx shadcn@latest add card
npx shadcn@latest add input
npx shadcn@latest add dialog
npx shadcn@latest add dropdown-menu
npx shadcn@latest add table
npx shadcn@latest add tabs
npx shadcn@latest add toast
```

## Multi-Theme Support

```css
/* Define multiple themes via CSS variables */
@layer base {
  :root { /* Default light */ }
  .dark { /* Dark theme */ }

  /* Additional themes */
  .theme-ocean {
    --background: 200 30% 98%;
    --foreground: 200 80% 10%;
    --primary: 200 80% 50%;
    --primary-foreground: 0 0% 100%;
  }

  .theme-forest {
    --background: 140 20% 97%;
    --foreground: 140 60% 10%;
    --primary: 140 60% 40%;
    --primary-foreground: 0 0% 100%;
  }
}
```

```typescript
// Theme switcher
type Theme = 'light' | 'dark' | 'ocean' | 'forest';

function setTheme(theme: Theme) {
  const root = document.documentElement;
  root.classList.remove('dark', 'theme-ocean', 'theme-forest');

  if (theme === 'dark') root.classList.add('dark');
  else if (theme !== 'light') root.classList.add(`theme-${theme}`);

  localStorage.setItem('theme', theme);
}
```

## Custom Plugin

```javascript
// tailwind.config.js
const plugin = require('tailwindcss/plugin');

module.exports = {
  plugins: [
    plugin(function({ addUtilities, addComponents, matchUtilities, theme }) {
      // Custom utilities
      addUtilities({
        '.text-balance': { 'text-wrap': 'balance' },
        '.text-pretty': { 'text-wrap': 'pretty' },
        '.scrollbar-hide': {
          '-ms-overflow-style': 'none',
          'scrollbar-width': 'none',
          '&::-webkit-scrollbar': { display: 'none' },
        },
      });

      // Dynamic utilities (with arbitrary values)
      matchUtilities(
        { 'grid-auto-fill': (value) => ({
            'grid-template-columns': `repeat(auto-fill, minmax(${value}, 1fr))`,
          }),
        },
        { values: theme('spacing') },
      );

      // Custom components
      addComponents({
        '.card': {
          '@apply rounded-xl border border-gray-200 bg-white p-6 shadow-sm': {},
        },
        '.input-field': {
          '@apply block w-full rounded-lg border border-gray-300 px-3 py-2 text-sm shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500': {},
        },
      });
    }),
  ],
};
```

## Prose / Typography Plugin

```html
<!-- Rich text content styling -->
<article class="prose prose-lg prose-blue mx-auto dark:prose-invert">
  <h1>Article Title</h1>
  <p>This content gets automatic typography styles.</p>
  <pre><code>Code blocks too.</code></pre>
</article>
```

```javascript
// Customizing prose styles
module.exports = {
  theme: {
    extend: {
      typography: (theme) => ({
        DEFAULT: {
          css: {
            maxWidth: '75ch',
            a: { color: theme('colors.blue.600'), '&:hover': { color: theme('colors.blue.800') } },
            'code::before': { content: '""' },
            'code::after': { content: '""' },
            code: { backgroundColor: theme('colors.gray.100'), padding: '2px 6px', borderRadius: '4px', fontSize: '0.875em' },
          },
        },
      }),
    },
  },
  plugins: [require('@tailwindcss/typography')],
};
```

## Gotchas

1. **`extend` vs top-level `theme` keys.** Putting colors at `theme.colors` REPLACES all defaults (no more gray, blue, etc.). Putting them at `theme.extend.colors` ADDS to defaults. Almost always use `extend`.

2. **CSS variable colors need special format.** When using `hsl(var(--color))`, define without the `hsl()` wrapper: `--color: 222 84% 5%`. The `hsl()` is added in the Tailwind config. This enables opacity modifiers like `bg-primary/50`.

3. **`@apply` can't use responsive/state variants.** You can't write `@apply hover:bg-blue-500` in CSS. Use the utility class directly in HTML, or write the CSS property manually.

4. **`tailwind-merge` is essential for component libraries.** Without `twMerge`, `cn('px-4', 'px-6')` keeps both classes (first wins). With `twMerge`, it keeps only `px-6` (last wins). Always use `cn()` for mergeable class props.

5. **Content path must include all files with Tailwind classes.** Missing a file path means those classes get purged in production. Common miss: forgetting to include `./node_modules/@my-org/ui/**/*.{js,ts,jsx,tsx}` for shared component libraries.

6. **Custom fonts need `preconnect` for performance.** Without `<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>`, Google Fonts adds 100-300ms to load time. Always add preconnect hints for external font sources.
