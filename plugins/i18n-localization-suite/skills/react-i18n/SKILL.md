---
name: react-i18n
description: >
  React internationalization with react-i18next — setup, namespaces,
  interpolation, pluralization, lazy loading, date/number formatting,
  language detection, and translation management.
  Triggers: "i18n", "react-i18next", "internationalization", "translations",
  "multi-language", "localization react", "translate app", "i18next".
  NOT for: Next.js App Router (use nextjs-i18n with next-intl).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# React i18n with react-i18next

## Setup

```bash
npm install i18next react-i18next i18next-http-backend i18next-browser-languagedetector
```

## Configuration

```typescript
// src/i18n/config.ts
import i18n from "i18next";
import { initReactI18next } from "react-i18next";
import HttpBackend from "i18next-http-backend";
import LanguageDetector from "i18next-browser-languagedetector";

i18n
  .use(HttpBackend) // Load translations from /public/locales/
  .use(LanguageDetector) // Detect user language
  .use(initReactI18next) // React bindings
  .init({
    // Languages
    supportedLngs: ["en", "es", "fr", "de", "ja", "zh"],
    fallbackLng: "en",

    // Namespaces
    ns: ["common", "auth", "dashboard"],
    defaultNS: "common",

    // Backend (where to load translations)
    backend: {
      loadPath: "/locales/{{lng}}/{{ns}}.json",
    },

    // Detection (order matters)
    detection: {
      order: ["querystring", "cookie", "localStorage", "navigator"],
      lookupQuerystring: "lang",
      lookupCookie: "i18next",
      lookupLocalStorage: "i18nextLng",
      caches: ["localStorage", "cookie"],
    },

    // Interpolation
    interpolation: {
      escapeValue: false, // React already escapes
    },

    // React options
    react: {
      useSuspense: true, // Use Suspense for loading
    },
  });

export default i18n;
```

```typescript
// src/main.tsx
import React, { Suspense } from "react";
import ReactDOM from "react-dom/client";
import App from "./App";
import "./i18n/config"; // Initialize i18n

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <Suspense fallback={<div>Loading...</div>}>
      <App />
    </Suspense>
  </React.StrictMode>
);
```

## Translation Files

```
public/
└── locales/
    ├── en/
    │   ├── common.json
    │   ├── auth.json
    │   └── dashboard.json
    ├── es/
    │   ├── common.json
    │   ├── auth.json
    │   └── dashboard.json
    └── fr/
        ├── common.json
        ├── auth.json
        └── dashboard.json
```

```json
// public/locales/en/common.json
{
  "app": {
    "title": "My Application",
    "tagline": "Build something great"
  },
  "nav": {
    "home": "Home",
    "about": "About",
    "settings": "Settings",
    "logout": "Log Out"
  },
  "actions": {
    "save": "Save",
    "cancel": "Cancel",
    "delete": "Delete",
    "confirm": "Confirm",
    "back": "Back"
  },
  "status": {
    "loading": "Loading...",
    "error": "Something went wrong",
    "empty": "No results found",
    "success": "Changes saved successfully"
  }
}
```

```json
// public/locales/en/auth.json
{
  "login": {
    "title": "Sign In",
    "email": "Email Address",
    "password": "Password",
    "submit": "Sign In",
    "forgotPassword": "Forgot your password?",
    "noAccount": "Don't have an account?",
    "signUp": "Sign Up"
  },
  "errors": {
    "invalidCredentials": "Invalid email or password",
    "emailRequired": "Email is required",
    "passwordRequired": "Password is required",
    "passwordMinLength": "Password must be at least {{min}} characters"
  },
  "welcome": "Welcome back, {{name}}!"
}
```

```json
// public/locales/es/auth.json
{
  "login": {
    "title": "Iniciar Sesión",
    "email": "Correo Electrónico",
    "password": "Contraseña",
    "submit": "Iniciar Sesión",
    "forgotPassword": "¿Olvidaste tu contraseña?",
    "noAccount": "¿No tienes una cuenta?",
    "signUp": "Registrarse"
  },
  "errors": {
    "invalidCredentials": "Correo o contraseña inválidos",
    "emailRequired": "El correo es obligatorio",
    "passwordRequired": "La contraseña es obligatoria",
    "passwordMinLength": "La contraseña debe tener al menos {{min}} caracteres"
  },
  "welcome": "¡Bienvenido de nuevo, {{name}}!"
}
```

## Using Translations

```tsx
// Hook-based (recommended)
import { useTranslation } from "react-i18next";

function LoginPage() {
  const { t } = useTranslation("auth"); // Use "auth" namespace

  return (
    <div>
      <h1>{t("login.title")}</h1>

      {/* Interpolation */}
      <p>{t("welcome", { name: "Alice" })}</p>
      {/* Output: "Welcome back, Alice!" */}

      {/* With count for pluralization */}
      <p>{t("items", { count: 5 })}</p>

      <form>
        <label>{t("login.email")}</label>
        <input type="email" />

        <label>{t("login.password")}</label>
        <input type="password" />

        <button type="submit">{t("login.submit")}</button>
      </form>

      <a href="/forgot">{t("login.forgotPassword")}</a>
    </div>
  );
}
```

```tsx
// Multiple namespaces
function Dashboard() {
  const { t } = useTranslation(["dashboard", "common"]);

  return (
    <div>
      {/* Namespace prefix */}
      <h1>{t("dashboard:title")}</h1>
      <button>{t("common:actions.save")}</button>
    </div>
  );
}
```

```tsx
// Trans component (for JSX interpolation)
import { Trans } from "react-i18next";

function Terms() {
  return (
    <Trans i18nKey="terms.agreement" t={t}>
      By clicking Sign Up, you agree to our
      <a href="/terms">Terms of Service</a> and
      <a href="/privacy">Privacy Policy</a>.
    </Trans>
  );
}

// Translation key:
// "terms.agreement": "By clicking Sign Up, you agree to our <1>Terms of Service</1> and <3>Privacy Policy</3>."
```

## Pluralization

```json
// English
{
  "items_zero": "No items",
  "items_one": "{{count}} item",
  "items_other": "{{count}} items",

  "messages": "You have {{count}} new message",
  "messages_plural": "You have {{count}} new messages"
}

// Arabic (6 plural forms)
{
  "items_zero": "لا عناصر",
  "items_one": "عنصر واحد",
  "items_two": "عنصران",
  "items_few": "{{count}} عناصر",
  "items_many": "{{count}} عنصرًا",
  "items_other": "{{count}} عنصر"
}
```

```tsx
// Usage
t("items", { count: 0 });  // "No items"
t("items", { count: 1 });  // "1 item"
t("items", { count: 42 }); // "42 items"
```

## Formatting (Intl)

```tsx
function PriceDisplay({ price, date }: { price: number; date: Date }) {
  const { t, i18n } = useTranslation();
  const locale = i18n.language;

  // Number formatting
  const formatted = new Intl.NumberFormat(locale, {
    style: "currency",
    currency: "USD",
  }).format(price);

  // Date formatting
  const dateStr = new Intl.DateTimeFormat(locale, {
    year: "numeric",
    month: "long",
    day: "numeric",
  }).format(date);

  // Relative time
  const rtf = new Intl.RelativeTimeFormat(locale, { numeric: "auto" });
  const daysAgo = Math.floor(
    (Date.now() - date.getTime()) / (1000 * 60 * 60 * 24)
  );
  const relative = rtf.format(-daysAgo, "day");

  return (
    <div>
      <p>{formatted}</p>     {/* $1,234.56 or 1.234,56 $ */}
      <p>{dateStr}</p>        {/* March 19, 2026 or 19 mars 2026 */}
      <p>{relative}</p>       {/* 3 days ago or hace 3 días */}
    </div>
  );
}
```

## Language Switcher

```tsx
function LanguageSwitcher() {
  const { i18n } = useTranslation();

  const languages = [
    { code: "en", label: "English", flag: "🇺🇸" },
    { code: "es", label: "Español", flag: "🇪🇸" },
    { code: "fr", label: "Français", flag: "🇫🇷" },
    { code: "de", label: "Deutsch", flag: "🇩🇪" },
    { code: "ja", label: "日本語", flag: "🇯🇵" },
  ];

  return (
    <select
      value={i18n.language}
      onChange={(e) => i18n.changeLanguage(e.target.value)}
    >
      {languages.map((lang) => (
        <option key={lang.code} value={lang.code}>
          {lang.flag} {lang.label}
        </option>
      ))}
    </select>
  );
}
```

## RTL Support

```tsx
// src/App.tsx
import { useTranslation } from "react-i18next";
import { useEffect } from "react";

const RTL_LANGUAGES = ["ar", "he", "fa", "ur"];

function App() {
  const { i18n } = useTranslation();
  const isRTL = RTL_LANGUAGES.includes(i18n.language);

  useEffect(() => {
    document.documentElement.dir = isRTL ? "rtl" : "ltr";
    document.documentElement.lang = i18n.language;
  }, [i18n.language, isRTL]);

  return <div className={isRTL ? "rtl" : "ltr"}>...</div>;
}
```

```css
/* Tailwind RTL utilities */
/* Use logical properties: */
.container {
  margin-inline-start: 1rem;  /* margin-left in LTR, margin-right in RTL */
  padding-inline-end: 1rem;   /* padding-right in LTR, padding-left in RTL */
}
```

## Translation Extraction

```bash
# Install extraction tool
npm install -D i18next-parser

# Create config
# i18next-parser.config.ts
export default {
  locales: ["en", "es", "fr"],
  output: "public/locales/$LOCALE/$NAMESPACE.json",
  input: ["src/**/*.{ts,tsx}"],
  defaultNamespace: "common",
  keySeparator: ".",
  namespaceSeparator: ":",
  sort: true,
};

# Run extraction
npx i18next-parser
```

## Gotchas

1. **`useSuspense: true` requires a `<Suspense>` boundary** — Without it, you'll get "Missing Suspense boundary" errors when translations are loading. Wrap your app or pages in `<Suspense fallback={<Loading />}>`.

2. **Namespace must be loaded before use** — If you `useTranslation("dashboard")` but that namespace hasn't been loaded, you'll see the raw key. Either preload namespaces or add them to the default loading list.

3. **Language detection runs once** — `i18next-browser-languagedetector` detects the language at init. If you change `navigator.language`, it won't auto-update. Call `i18n.changeLanguage()` explicitly.

4. **Translation keys are case-sensitive** — `t("Login")` and `t("login")` are different keys. Use consistent casing (lowercase with dots is most common).

5. **`<Trans>` indexes are positional** — `<1>`, `<3>` in the translation string correspond to the 2nd and 4th children of the `<Trans>` component. Adding or removing children shifts indexes and breaks translations.

6. **Don't use template literals as keys** — `` t(`status.${status}`) `` makes keys impossible to extract automatically. Use explicit keys with a mapping object instead.
