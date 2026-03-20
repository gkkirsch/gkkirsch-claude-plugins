# i18n & Localization Cheatsheet

## react-i18next Quick Reference

### Setup
```bash
npm install i18next react-i18next i18next-http-backend i18next-browser-languagedetector
```

### Basic Usage
```tsx
import { useTranslation } from "react-i18next";

function MyComponent() {
  const { t, i18n } = useTranslation("namespace");

  return (
    <>
      <h1>{t("key")}</h1>
      <p>{t("greeting", { name: "Alice" })}</p>
      <p>{t("items", { count: 5 })}</p>
      <button onClick={() => i18n.changeLanguage("es")}>ES</button>
    </>
  );
}
```

### Pluralization
```json
{
  "items_zero": "No items",
  "items_one": "{{count}} item",
  "items_other": "{{count}} items"
}
```

### JSX in Translations
```tsx
import { Trans } from "react-i18next";
<Trans i18nKey="terms">
  Agree to <a href="/terms">Terms</a>
</Trans>
// Key: "Agree to <1>Terms</1>"
```

---

## next-intl Quick Reference

### Setup
```bash
npm install next-intl
```

### Server Component
```tsx
import { useTranslations } from "next-intl/server";
const t = useTranslations("Namespace");
return <h1>{t("key", { name: "Alice" })}</h1>;
```

### Client Component
```tsx
"use client";
import { useTranslations } from "next-intl";
const t = useTranslations("Namespace");
```

### ICU Plurals
```json
{
  "items": "{count, plural, =0 {None} one {# item} other {# items}}"
}
```

### ICU Select
```json
{
  "pronoun": "{gender, select, male {He} female {She} other {They}}"
}
```

### Navigation
```tsx
import { Link, useRouter, usePathname } from "@/i18n/navigation";
<Link href="/about">About</Link>
router.replace(pathname, { locale: "es" });
```

### Formatting
```tsx
import { useFormatter } from "next-intl";
const format = useFormatter();
format.number(1234, { style: "currency", currency: "USD" });
format.dateTime(date, { dateStyle: "medium" });
format.relativeTime(date);
format.list(["a", "b", "c"], { type: "conjunction" });
```

---

## Intl API (Built-in Browser)

### Number Formatting
```typescript
new Intl.NumberFormat(locale, { style: "currency", currency: "USD" }).format(1234.56);
new Intl.NumberFormat(locale, { notation: "compact" }).format(1500000); // "1.5M"
new Intl.NumberFormat(locale, { style: "percent" }).format(0.42);      // "42%"
```

### Date Formatting
```typescript
new Intl.DateTimeFormat(locale, { dateStyle: "full" }).format(date);
new Intl.DateTimeFormat(locale, { year: "numeric", month: "long", day: "numeric" }).format(date);
```

### Relative Time
```typescript
new Intl.RelativeTimeFormat(locale, { numeric: "auto" }).format(-1, "day");  // "yesterday"
new Intl.RelativeTimeFormat(locale).format(-3, "hour");                       // "3 hours ago"
```

### List
```typescript
new Intl.ListFormat(locale, { type: "conjunction" }).format(["a", "b", "c"]); // "a, b, and c"
new Intl.ListFormat(locale, { type: "disjunction" }).format(["a", "b"]);      // "a or b"
```

### Collation (Sorting)
```typescript
["ä", "a", "z"].sort(new Intl.Collator(locale).compare); // Locale-aware sort
```

---

## ICU Message Format Reference

| Pattern | Example | Output |
|---------|---------|--------|
| Simple | `Hello, {name}!` | Hello, Alice! |
| Number | `{n, number}` | 1,234 |
| Currency | `{n, number, currency}` | $1,234.00 |
| Percent | `{n, number, percent}` | 42% |
| Date | `{d, date, medium}` | Mar 19, 2026 |
| Time | `{d, time, short}` | 3:30 PM |
| Plural | `{n, plural, one {# item} other {# items}}` | 5 items |
| Select | `{g, select, male {He} female {She} other {They}}` | She |
| Ordinal | `{n, selectordinal, one {#st} two {#nd} few {#rd} other {#th}}` | 3rd |

---

## Translation File Organization

### By Feature (Recommended)
```
locales/en/
├── common.json      # Shared UI: buttons, labels, status
├── auth.json        # Login, signup, password reset
├── dashboard.json   # Dashboard-specific text
├── settings.json    # Settings page
└── errors.json      # Error messages
```

### Key Naming Convention
```json
{
  "feature.component.element": "Text",
  "auth.login.title": "Sign In",
  "auth.login.submitButton": "Sign In",
  "auth.errors.invalidEmail": "Please enter a valid email"
}
```

---

## RTL Support Checklist

- [ ] Set `dir="rtl"` on `<html>` for RTL languages
- [ ] Use CSS logical properties (`margin-inline-start` not `margin-left`)
- [ ] Use `text-align: start` not `text-align: left`
- [ ] Mirror icons that indicate direction (arrows, chevrons)
- [ ] Don't mirror: logos, numbers, media controls, checkmarks
- [ ] Test with actual RTL content, not just `dir="rtl"`
- [ ] RTL languages: Arabic, Hebrew, Farsi, Urdu

## Quick Decision

| Scenario | Library |
|----------|---------|
| Next.js App Router + RSC | **next-intl** |
| React SPA or Vite | **react-i18next** |
| Need ICU + smallest bundle | **Lingui** |
| Enterprise + complex plural | **FormatJS** |
