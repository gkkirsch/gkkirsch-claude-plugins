---
name: i18n-testing
description: >
  Testing patterns for internationalized applications.
  Use when writing tests for translated content, RTL layouts,
  locale-specific formatting, pluralization rules, or ICU messages.
  Triggers: "i18n testing", "test translations", "RTL testing",
  "locale testing", "pluralization test", "ICU message test",
  "internationalization testing", "translation coverage".
  NOT for: setting up i18n (see react-i18n, nextjs-i18n), translation management platforms, general testing.
version: 1.0.0
allowed-tools: Read, Grep, Glob, Bash
---

# i18n Testing Patterns

## Translation Completeness Checker

```typescript
// scripts/check-translations.ts — Verify all locales have all keys
import fs from 'fs';
import path from 'path';

interface TranslationFile {
  locale: string;
  keys: Set<string>;
  filePath: string;
}

function flattenKeys(obj: Record<string, unknown>, prefix = ''): string[] {
  const keys: string[] = [];
  for (const [key, value] of Object.entries(obj)) {
    const fullKey = prefix ? `${prefix}.${key}` : key;
    if (typeof value === 'object' && value !== null && !Array.isArray(value)) {
      keys.push(...flattenKeys(value as Record<string, unknown>, fullKey));
    } else {
      keys.push(fullKey);
    }
  }
  return keys;
}

function checkTranslations(localesDir: string): {
  missing: Map<string, string[]>;
  extra: Map<string, string[]>;
  empty: Map<string, string[]>;
} {
  const files: TranslationFile[] = [];
  const allKeys = new Set<string>();

  // Load all translation files
  for (const file of fs.readdirSync(localesDir)) {
    if (!file.endsWith('.json')) continue;
    const locale = path.basename(file, '.json');
    const content = JSON.parse(fs.readFileSync(path.join(localesDir, file), 'utf-8'));
    const keys = new Set(flattenKeys(content));
    keys.forEach(k => allKeys.add(k));
    files.push({ locale, keys, filePath: path.join(localesDir, file) });
  }

  const missing = new Map<string, string[]>();
  const extra = new Map<string, string[]>();
  const empty = new Map<string, string[]>();

  const baseLocale = files.find(f => f.locale === 'en')!;

  for (const file of files) {
    // Keys in base but not in this locale
    const missingKeys = [...baseLocale.keys].filter(k => !file.keys.has(k));
    if (missingKeys.length) missing.set(file.locale, missingKeys);

    // Keys in this locale but not in base (probably stale)
    const extraKeys = [...file.keys].filter(k => !baseLocale.keys.has(k));
    if (extraKeys.length) extra.set(file.locale, extraKeys);

    // Empty string values
    const content = JSON.parse(fs.readFileSync(file.filePath, 'utf-8'));
    const emptyKeys = flattenKeys(content).filter(key => {
      const value = key.split('.').reduce((obj: any, k) => obj?.[k], content);
      return value === '' || value === null;
    });
    if (emptyKeys.length) empty.set(file.locale, emptyKeys);
  }

  return { missing, extra, empty };
}

// Usage in CI:
// const result = checkTranslations('./public/locales');
// if (result.missing.size > 0) process.exit(1);
```

## ICU Message Format Testing

```typescript
// __tests__/icu-messages.test.ts
import { IntlMessageFormat } from 'intl-messageformat';

describe('ICU Message Format', () => {
  // Pluralization rules
  describe('plural rules', () => {
    const message = new IntlMessageFormat(
      `{count, plural,
        =0 {No items}
        one {# item}
        other {# items}
      }`,
      'en'
    );

    it.each([
      [0, 'No items'],
      [1, '1 item'],
      [2, '2 items'],
      [100, '100 items'],
    ])('formats count=%i as "%s"', (count, expected) => {
      expect(message.format({ count })).toBe(expected);
    });
  });

  // Gender-based messages
  describe('select rules', () => {
    const message = new IntlMessageFormat(
      `{gender, select,
        male {He joined}
        female {She joined}
        other {They joined}
      } the team.`,
      'en'
    );

    it.each([
      ['male', 'He joined the team.'],
      ['female', 'She joined the team.'],
      ['nonbinary', 'They joined the team.'],
    ])('gender=%s → "%s"', (gender, expected) => {
      expect(message.format({ gender })).toBe(expected);
    });
  });

  // Number formatting varies by locale
  describe('number formatting by locale', () => {
    const template = '{price, number, ::currency/USD}';

    it('formats USD in en-US', () => {
      const msg = new IntlMessageFormat(template, 'en-US');
      expect(msg.format({ price: 1234.5 })).toContain('1,234.50');
    });

    it('formats USD in de-DE', () => {
      const msg = new IntlMessageFormat(template, 'de-DE');
      const result = msg.format({ price: 1234.5 }) as string;
      // German uses comma for decimal, period for thousands
      expect(result).toMatch(/1\.?234,50/);
    });
  });

  // Date formatting varies by locale
  describe('date formatting by locale', () => {
    const date = new Date('2026-03-15T10:30:00Z');

    it('en-US: month/day/year', () => {
      const formatted = new Intl.DateTimeFormat('en-US').format(date);
      expect(formatted).toBe('3/15/2026');
    });

    it('de-DE: day.month.year', () => {
      const formatted = new Intl.DateTimeFormat('de-DE').format(date);
      expect(formatted).toBe('15.3.2026');
    });

    it('ja-JP: year/month/day', () => {
      const formatted = new Intl.DateTimeFormat('ja-JP').format(date);
      expect(formatted).toBe('2026/3/15');
    });
  });
});
```

## RTL Layout Testing

```typescript
// __tests__/rtl-layout.test.tsx
import { render, screen } from '@testing-library/react';
import { I18nextProvider } from 'react-i18next';

// Helper: render component with specific locale and direction
function renderWithLocale(ui: React.ReactElement, locale: string) {
  const dir = ['ar', 'he', 'fa', 'ur'].includes(locale) ? 'rtl' : 'ltr';

  return render(
    <I18nextProvider i18n={createI18nInstance(locale)}>
      <div dir={dir} lang={locale}>
        {ui}
      </div>
    </I18nextProvider>
  );
}

describe('RTL Layout', () => {
  it('sets dir="rtl" for Arabic', () => {
    const { container } = renderWithLocale(<App />, 'ar');
    expect(container.firstChild).toHaveAttribute('dir', 'rtl');
  });

  it('sets dir="ltr" for English', () => {
    const { container } = renderWithLocale(<App />, 'en');
    expect(container.firstChild).toHaveAttribute('dir', 'ltr');
  });

  it('mirrors navigation for RTL', () => {
    renderWithLocale(<Navigation />, 'ar');
    const nav = screen.getByRole('navigation');
    const style = window.getComputedStyle(nav);
    // Logical properties should be used instead of left/right
    expect(style.paddingInlineStart).toBeDefined();
  });
});

// Visual regression test for RTL
// playwright.config.ts
// projects: [
//   { name: 'ltr', use: { locale: 'en-US' } },
//   { name: 'rtl', use: { locale: 'ar-SA' } },
// ]
```

## Snapshot Testing Translations

```typescript
// __tests__/translation-snapshots.test.ts
import i18n from '../i18n';

const LOCALES = ['en', 'es', 'fr', 'de', 'ja', 'ar'];
const KEY_GROUPS = [
  'common',     // Shared UI strings
  'auth',       // Login/signup
  'dashboard',  // Main app
  'errors',     // Error messages
];

describe('Translation snapshots', () => {
  for (const locale of LOCALES) {
    describe(`locale: ${locale}`, () => {
      beforeAll(async () => {
        await i18n.changeLanguage(locale);
      });

      for (const group of KEY_GROUPS) {
        it(`${group} translations match snapshot`, () => {
          const bundle = i18n.getResourceBundle(locale, group);
          expect(bundle).toMatchSnapshot(`${locale}-${group}`);
        });
      }
    });
  }
});

// Snapshot testing catches:
// 1. Accidental key deletions
// 2. Changed translations without review
// 3. Missing interpolation variables
// 4. Broken ICU message syntax
// Run: npx jest --updateSnapshot to accept changes after review
```

## CI Pipeline for i18n

```yaml
# .github/workflows/i18n-check.yml
name: i18n Quality
on: [pull_request]

jobs:
  translations:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with: { node-version: 20 }
      - run: npm ci

      # Check for missing translation keys
      - name: Check translation completeness
        run: npx ts-node scripts/check-translations.ts

      # Validate ICU message syntax
      - name: Validate ICU messages
        run: |
          npx ts-node -e "
            const files = require('glob').sync('public/locales/**/*.json');
            const { IntlMessageFormat } = require('intl-messageformat');
            let errors = 0;
            for (const file of files) {
              const locale = file.split('/')[2];
              const data = require('./' + file);
              for (const [key, value] of Object.entries(flattenKeys(data))) {
                try { new IntlMessageFormat(value, locale); }
                catch (e) { console.error(file, key, e.message); errors++; }
              }
            }
            if (errors) process.exit(1);
          "

      # Check for hardcoded strings in components
      - name: Check for hardcoded strings
        run: |
          # Find JSX text content that isn't wrapped in t() or <Trans>
          grep -rn --include='*.tsx' --include='*.jsx' \
            -P '>\s*[A-Z][a-z]+(\s+[a-z]+){2,}\s*<' src/ \
            && echo "Found potential hardcoded strings" && exit 1 \
            || echo "No hardcoded strings found"

      # Run translation snapshot tests
      - name: Snapshot tests
        run: npx jest --testPathPattern=translation-snapshots
```

## Gotchas

1. **Pluralization rules differ wildly by language** -- English has 2 forms (one, other). Arabic has 6 (zero, one, two, few, many, other). Polish has 4. Testing only English plural rules misses bugs in other locales. Test at least one language from each CLDR plural rule group: English (one/other), French (one/other with different threshold), Arabic (all 6 forms), Polish (few/many distinction).

2. **String concatenation breaks translation** -- `t('hello') + ' ' + t('world')` prevents translators from reordering words. Different languages have different word orders. Use a single key with interpolation: `t('greeting', { name })` → "Hello, {{name}}" (en) → "{{name}}さん、こんにちは" (ja). Never concatenate translated strings.

3. **Hardcoded date/number formats** -- `date.toLocaleDateString()` without a locale argument uses the runtime's locale, not the user's selected locale. Always pass the locale: `date.toLocaleDateString(currentLocale)`. Same for `Intl.NumberFormat`, `Intl.DateTimeFormat`, and `Intl.RelativeTimeFormat`.

4. **Text expansion breaks layouts** -- German text is ~30% longer than English. Finnish can be 40% longer. Russian uses wide Cyrillic characters. Test with pseudo-localization (e.g., "Login" → "[Llloogggiiin]") to catch layout overflow before translators start. Avoid fixed-width containers for translated text.

5. **RTL is more than just dir="rtl"** -- Right-to-left languages require mirrored icons (arrows, progress bars), swapped margin/padding, and logical CSS properties (`margin-inline-start` instead of `margin-left`). CSS `direction: rtl` alone doesn't fix layout issues. Use CSS logical properties throughout and test with real RTL content, not just `dir` attribute.

6. **Translation files bloat bundle size** -- Loading all locales on startup defeats code splitting. Use dynamic imports (`import(`./locales/${locale}.json`)`) and load only the active locale. For large apps, split translations by route or feature. A 50KB translation file per locale adds up fast with 20+ supported languages.
