# Design Token Patterns Reference

Comprehensive reference for design token naming, multi-brand systems, Style Dictionary workflows, and Figma integration.

---

## Token Naming Conventions

### CTI (Category–Type–Item) Convention

The most widely adopted naming structure:

```
{category}.{type}.{item}.{subitem}.{state}
```

Examples:
```
color.bg.default
color.bg.brand.hover
color.text.primary
color.text.link.visited
color.border.default
color.border.focus

font.family.sans
font.size.base
font.weight.bold
font.lineHeight.tight

space.inline.page
space.block.section
space.gap.card

radius.sm
radius.lg
radius.full

shadow.sm
shadow.lg

motion.duration.fast
motion.duration.normal
motion.easing.spring
motion.easing.smooth

z.dropdown
z.modal
z.toast
```

### Semantic vs Primitive Token Names

```
PRIMITIVE (Tier 1)          SEMANTIC (Tier 2)              COMPONENT (Tier 3)
─────────────────           ──────────────────             ──────────────────
color.blue.500         →    color.bg.brand.default    →    button.primary.bg
color.blue.600         →    color.bg.brand.hover      →    button.primary.bg.hover
color.gray.900         →    color.text.default        →    card.title.color
color.gray.200         →    color.border.default      →    input.border.color
space.4                →    space.inline.component    →    button.padding.inline
radius.lg              →    radius.interactive        →    button.radius
```

### Rules for Naming

1. **Always use semantic names in components** — never `color-blue-500`, always `color-bg-brand`
2. **Use consistent hierarchy** — `{what}.{where}.{which}.{when}`
3. **Include state in the name** — `hover`, `active`, `disabled`, `focus`
4. **Keep primitives for internal use only** — they're the building blocks, not the API
5. **Name by purpose, not appearance** — `color-text-secondary` not `color-gray-600`

---

## Token Types

### Color Tokens

```json
{
  "color": {
    "primitive": {
      "blue": {
        "50":  { "value": "oklch(97% 0.02 250)", "type": "color" },
        "100": { "value": "oklch(93% 0.04 250)", "type": "color" },
        "200": { "value": "oklch(87% 0.08 250)", "type": "color" },
        "300": { "value": "oklch(78% 0.12 250)", "type": "color" },
        "400": { "value": "oklch(68% 0.17 250)", "type": "color" },
        "500": { "value": "oklch(55% 0.22 250)", "type": "color" },
        "600": { "value": "oklch(47% 0.2 250)",  "type": "color" },
        "700": { "value": "oklch(40% 0.18 250)", "type": "color" },
        "800": { "value": "oklch(33% 0.14 250)", "type": "color" },
        "900": { "value": "oklch(25% 0.1 250)",  "type": "color" }
      }
    },
    "semantic": {
      "bg": {
        "default":    { "value": "{color.primitive.neutral.0}" },
        "subtle":     { "value": "{color.primitive.neutral.50}" },
        "muted":      { "value": "{color.primitive.neutral.100}" },
        "inverse":    { "value": "{color.primitive.neutral.900}" },
        "brand": {
          "default":  { "value": "{color.primitive.blue.500}" },
          "hover":    { "value": "{color.primitive.blue.600}" },
          "active":   { "value": "{color.primitive.blue.700}" },
          "subtle":   { "value": "{color.primitive.blue.50}" }
        }
      }
    }
  }
}
```

### Spacing Tokens

```json
{
  "space": {
    "primitive": {
      "0":    { "value": "0",       "type": "dimension" },
      "px":   { "value": "1px",     "type": "dimension" },
      "0.5":  { "value": "0.125rem","type": "dimension" },
      "1":    { "value": "0.25rem", "type": "dimension" },
      "2":    { "value": "0.5rem",  "type": "dimension" },
      "3":    { "value": "0.75rem", "type": "dimension" },
      "4":    { "value": "1rem",    "type": "dimension" },
      "6":    { "value": "1.5rem",  "type": "dimension" },
      "8":    { "value": "2rem",    "type": "dimension" },
      "12":   { "value": "3rem",    "type": "dimension" },
      "16":   { "value": "4rem",    "type": "dimension" },
      "24":   { "value": "6rem",    "type": "dimension" }
    },
    "semantic": {
      "page": {
        "inline": { "value": "{space.primitive.6}" },
        "block":  { "value": "{space.primitive.16}" }
      },
      "section": {
        "gap": { "value": "{space.primitive.12}" }
      },
      "component": {
        "inline": { "value": "{space.primitive.4}" },
        "block":  { "value": "{space.primitive.3}" },
        "gap":    { "value": "{space.primitive.3}" }
      },
      "element": {
        "inline": { "value": "{space.primitive.2}" },
        "block":  { "value": "{space.primitive.1}" },
        "gap":    { "value": "{space.primitive.2}" }
      }
    }
  }
}
```

### Typography Composite Tokens

```json
{
  "typography": {
    "display": {
      "xl": {
        "value": {
          "fontFamily": "{font.family.display}",
          "fontSize": "{font.size.5xl}",
          "fontWeight": "{font.weight.bold}",
          "lineHeight": "{font.lineHeight.tight}",
          "letterSpacing": "-0.03em"
        },
        "type": "typography"
      }
    },
    "heading": {
      "lg": {
        "value": {
          "fontFamily": "{font.family.sans}",
          "fontSize": "{font.size.3xl}",
          "fontWeight": "{font.weight.bold}",
          "lineHeight": "{font.lineHeight.tight}",
          "letterSpacing": "-0.02em"
        },
        "type": "typography"
      }
    },
    "body": {
      "md": {
        "value": {
          "fontFamily": "{font.family.sans}",
          "fontSize": "{font.size.base}",
          "fontWeight": "{font.weight.regular}",
          "lineHeight": "{font.lineHeight.normal}"
        },
        "type": "typography"
      }
    }
  }
}
```

### Shadow Tokens

```json
{
  "shadow": {
    "primitive": {
      "sm": {
        "value": "0 1px 2px 0 oklch(0% 0 0 / 0.05)",
        "type": "shadow"
      },
      "md": {
        "value": [
          { "x": 0, "y": 4, "blur": 6, "spread": -1, "color": "oklch(0% 0 0 / 0.1)" },
          { "x": 0, "y": 2, "blur": 4, "spread": -2, "color": "oklch(0% 0 0 / 0.1)" }
        ],
        "type": "shadow"
      },
      "lg": {
        "value": [
          { "x": 0, "y": 10, "blur": 15, "spread": -3, "color": "oklch(0% 0 0 / 0.1)" },
          { "x": 0, "y": 4, "blur": 6, "spread": -4, "color": "oklch(0% 0 0 / 0.1)" }
        ],
        "type": "shadow"
      }
    },
    "semantic": {
      "card": { "value": "{shadow.primitive.sm}" },
      "dropdown": { "value": "{shadow.primitive.lg}" },
      "modal": { "value": "{shadow.primitive.xl}" }
    }
  }
}
```

---

## Multi-Brand Token Strategy

### File Structure

```
tokens/
├── core/                    # Shared across all brands
│   ├── spacing.json
│   ├── radius.json
│   ├── motion.json
│   └── z-index.json
├── brands/
│   ├── acme/
│   │   ├── color.json      # Brand-specific color primitives
│   │   ├── typography.json  # Brand fonts
│   │   └── overrides.json   # Any semantic overrides
│   └── beta/
│       ├── color.json
│       ├── typography.json
│       └── overrides.json
├── semantic/
│   ├── light.json           # Light theme semantic tokens
│   └── dark.json            # Dark theme semantic tokens
└── component/
    ├── button.json
    ├── input.json
    └── card.json
```

### Brand Configuration

```json
// tokens/brands/acme/color.json
{
  "color": {
    "brand": {
      "50":  { "value": "oklch(97% 0.02 280)" },
      "100": { "value": "oklch(93% 0.05 280)" },
      "500": { "value": "oklch(55% 0.25 280)" },
      "600": { "value": "oklch(47% 0.22 280)" },
      "700": { "value": "oklch(40% 0.2 280)" }
    }
  }
}

// tokens/brands/beta/color.json
{
  "color": {
    "brand": {
      "50":  { "value": "oklch(97% 0.02 145)" },
      "100": { "value": "oklch(93% 0.04 145)" },
      "500": { "value": "oklch(55% 0.2 145)" },
      "600": { "value": "oklch(47% 0.18 145)" },
      "700": { "value": "oklch(40% 0.16 145)" }
    }
  }
}
```

### Build Configuration for Multiple Brands

```js
// style-dictionary.config.mjs
const brands = ['acme', 'beta'];
const themes = ['light', 'dark'];

export default brands.flatMap((brand) =>
  themes.map((theme) => ({
    source: [
      'tokens/core/**/*.json',
      `tokens/brands/${brand}/**/*.json`,
      `tokens/semantic/${theme}.json`,
      'tokens/component/**/*.json',
    ],
    platforms: {
      css: {
        transformGroup: 'css',
        buildPath: `dist/${brand}/`,
        files: [
          {
            destination: `${theme}.css`,
            format: 'css/variables',
            options: {
              selector: theme === 'light'
                ? `:root, [data-theme="light"]`
                : `[data-theme="dark"]`,
            },
          },
        ],
      },
    },
  }))
);
```

Output:
```
dist/
├── acme/
│   ├── light.css
│   └── dark.css
└── beta/
    ├── light.css
    └── dark.css
```

---

## Style Dictionary Deep Dive

### Custom Transforms

```js
import StyleDictionary from 'style-dictionary';

// Transform oklch to hex for legacy platforms
StyleDictionary.registerTransform({
  name: 'color/oklchToHex',
  type: 'value',
  filter: (token) => token.type === 'color',
  transform: (token) => {
    // Use a library like culori for conversion
    // return convertOklchToHex(token.value);
    return token.value;
  },
});

// Transform px values to rem
StyleDictionary.registerTransform({
  name: 'size/pxToRem',
  type: 'value',
  filter: (token) =>
    token.type === 'dimension' &&
    typeof token.value === 'string' &&
    token.value.endsWith('px'),
  transform: (token) => {
    const px = parseFloat(token.value);
    return `${px / 16}rem`;
  },
});

// Transform composite typography token to CSS
StyleDictionary.registerTransform({
  name: 'typography/css',
  type: 'value',
  filter: (token) => token.type === 'typography',
  transform: (token) => {
    const { fontWeight, fontSize, lineHeight, fontFamily } = token.value;
    return `${fontWeight} ${fontSize}/${lineHeight} ${fontFamily}`;
  },
});

// Transform token name to CSS custom property
StyleDictionary.registerTransform({
  name: 'name/kebab',
  type: 'name',
  transform: (token) => {
    return token.path.join('-').toLowerCase();
  },
});
```

### Custom Formats

```js
// CSS with @layer
StyleDictionary.registerFormat({
  name: 'css/layered-variables',
  format: ({ dictionary, options }) => {
    const { layer = 'tokens', selector = ':root' } = options;
    const vars = dictionary.allTokens
      .map((token) => `  --${token.name}: ${token.value};`)
      .join('\n');
    return `@layer ${layer} {\n  ${selector} {\n${vars}\n  }\n}\n`;
  },
});

// TypeScript constants
StyleDictionary.registerFormat({
  name: 'typescript/constants',
  format: ({ dictionary }) => {
    const tokens = dictionary.allTokens
      .map((token) => {
        const name = token.path
          .map((p) => p.charAt(0).toUpperCase() + p.slice(1))
          .join('');
        return `export const ${name} = '${token.value}' as const;`;
      })
      .join('\n');
    return `// Auto-generated — do not edit\n${tokens}\n`;
  },
});

// Tailwind theme config
StyleDictionary.registerFormat({
  name: 'tailwind/theme',
  format: ({ dictionary }) => {
    const theme = {};
    dictionary.allTokens.forEach((token) => {
      let obj = theme;
      const path = token.path.slice(0, -1);
      const key = token.path[token.path.length - 1];
      path.forEach((segment) => {
        obj[segment] = obj[segment] || {};
        obj = obj[segment];
      });
      obj[key] = token.value;
    });
    return `// Auto-generated Tailwind theme\nexport default ${JSON.stringify(theme, null, 2)};\n`;
  },
});
```

### Filters

```js
// Only export semantic tokens (not primitives)
StyleDictionary.registerFilter({
  name: 'semantic-only',
  filter: (token) => !token.path.includes('primitive'),
});

// Only color tokens
StyleDictionary.registerFilter({
  name: 'is-color',
  filter: (token) => token.type === 'color',
});
```

---

## Figma Integration

### Tokens Studio for Figma

#### Token Structure Mapping

| Tokens Studio | Style Dictionary | CSS Output |
|---------------|-----------------|------------|
| `colors/brand/500` | `color.brand.500` | `--color-brand-500` |
| `typography/heading/lg` | `typography.heading.lg` | `--typography-heading-lg` |
| `spacing/4` | `space.4` | `--space-4` |

#### Theme Configuration

```json
{
  "$themes": [
    {
      "id": "light",
      "name": "Light",
      "selectedTokenSets": {
        "core/color-primitives": "source",
        "core/spacing": "source",
        "semantic/light": "enabled",
        "component/button": "enabled"
      }
    },
    {
      "id": "dark",
      "name": "Dark",
      "selectedTokenSets": {
        "core/color-primitives": "source",
        "core/spacing": "source",
        "semantic/dark": "enabled",
        "component/button": "enabled"
      }
    }
  ]
}
```

- **`source`**: Token set provides values but tokens aren't applied directly
- **`enabled`**: Tokens are active and applied
- **`disabled`**: Token set is ignored

### Sync Workflow

```
1. Design in Figma → tokens defined in Tokens Studio
2. Push tokens → Git repository (tokens/ directory)
3. CI runs Style Dictionary → generates CSS, JS, JSON
4. Developers import generated tokens
5. Changes flow both ways via pull requests
```

#### GitHub Actions Sync

```yaml
# .github/workflows/tokens.yml
name: Build Tokens

on:
  push:
    paths:
      - 'tokens/**'

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: 20
      - run: npm ci
      - run: npx style-dictionary build
      - uses: actions/upload-artifact@v4
        with:
          name: tokens
          path: dist/
```

---

## Token Validation

### Schema Validation

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "properties": {
    "value": {
      "oneOf": [
        { "type": "string" },
        { "type": "number" },
        { "type": "object" }
      ]
    },
    "type": {
      "enum": [
        "color",
        "dimension",
        "fontFamily",
        "fontWeight",
        "fontSize",
        "lineHeight",
        "letterSpacing",
        "shadow",
        "borderRadius",
        "typography",
        "opacity",
        "duration",
        "cubicBezier"
      ]
    },
    "description": { "type": "string" }
  },
  "required": ["value"]
}
```

### Contrast Validation

```js
// Validate that text/background token combinations meet WCAG contrast
function validateContrast(tokens) {
  const pairs = [
    ['color.text.default', 'color.bg.default'],
    ['color.text.secondary', 'color.bg.default'],
    ['color.text.muted', 'color.bg.default'],
    ['color.text.inverse', 'color.bg.inverse'],
    ['color.text.brand', 'color.bg.default'],
    ['color.text.danger', 'color.bg.default'],
  ];

  const results = pairs.map(([fg, bg]) => {
    const fgValue = resolveToken(tokens, fg);
    const bgValue = resolveToken(tokens, bg);
    const ratio = calculateContrastRatio(fgValue, bgValue);
    const passes = ratio >= 4.5; // WCAG AA for normal text

    return { fg, bg, ratio: ratio.toFixed(2), passes };
  });

  return results;
}
```

---

## Anti-Patterns

1. **Don't use primitive tokens in components** — always go through semantic layer
2. **Don't create one-off tokens** — if a value is used once, it might not need to be a token
3. **Don't skip the naming convention** — inconsistent names make tokens unusable
4. **Don't hardcode values alongside tokens** — either everything is tokenized or nothing is
5. **Don't forget dark mode when creating semantic tokens** — every semantic color needs both light and dark values
6. **Don't mix token formats** — pick one format (DTCG, Tokens Studio, custom) and stick with it
7. **Don't version tokens independently from components** — they ship together
8. **Don't generate tokens without validation** — check contrast, completeness, and references before publishing
