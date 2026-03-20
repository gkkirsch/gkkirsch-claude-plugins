---
name: nextjs-i18n
description: >
  Next.js internationalization with next-intl — App Router, Server Components,
  middleware routing, message formatting, type-safe translations.
  Triggers: "next-intl", "nextjs i18n", "nextjs translations",
  "nextjs localization", "nextjs multi-language", "app router i18n".
  NOT for: non-Next.js React apps (use react-i18n with react-i18next).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Next.js i18n with next-intl

## Setup

```bash
npm install next-intl
```

## Project Structure

```
src/
├── i18n/
│   ├── request.ts          # Server-side i18n config
│   ├── routing.ts           # Routing config
│   └── navigation.ts        # Localized navigation
├── messages/                # Translation files
│   ├── en.json
│   ├── es.json
│   └── fr.json
├── middleware.ts             # Locale detection + routing
└── app/
    └── [locale]/            # Locale-prefixed routes
        ├── layout.tsx
        ├── page.tsx
        └── dashboard/
            └── page.tsx
```

## Configuration

```typescript
// src/i18n/routing.ts
import { defineRouting } from "next-intl/routing";

export const routing = defineRouting({
  locales: ["en", "es", "fr", "de", "ja"],
  defaultLocale: "en",
  localePrefix: "as-needed", // "always" | "as-needed" | "never"
  // "as-needed" = no prefix for default locale (/ instead of /en)
});
```

```typescript
// src/i18n/navigation.ts
import { createNavigation } from "next-intl/navigation";
import { routing } from "./routing";

// Type-safe navigation helpers
export const { Link, redirect, usePathname, useRouter, getPathname } =
  createNavigation(routing);
```

```typescript
// src/i18n/request.ts
import { getRequestConfig } from "next-intl/server";
import { routing } from "./routing";

export default getRequestConfig(async ({ requestLocale }) => {
  let locale = await requestLocale;

  // Validate locale
  if (!locale || !routing.locales.includes(locale as any)) {
    locale = routing.defaultLocale;
  }

  return {
    locale,
    messages: (await import(`../messages/${locale}.json`)).default,
    // Optional: timezone and date/time formatting
    timeZone: "America/New_York",
    now: new Date(),
  };
});
```

```typescript
// src/middleware.ts
import createMiddleware from "next-intl/middleware";
import { routing } from "./i18n/routing";

export default createMiddleware(routing);

export const config = {
  matcher: [
    // Match all pathnames except API routes, static files, etc.
    "/((?!api|_next|_vercel|.*\\..*).*)",
  ],
};
```

```typescript
// next.config.ts
import createNextIntlPlugin from "next-intl/plugin";

const withNextIntl = createNextIntlPlugin("./src/i18n/request.ts");

const nextConfig = {
  // ... your config
};

export default withNextIntl(nextConfig);
```

## Translation Files (ICU Format)

```json
// src/messages/en.json
{
  "Navigation": {
    "home": "Home",
    "about": "About",
    "dashboard": "Dashboard",
    "settings": "Settings"
  },
  "HomePage": {
    "title": "Welcome to {appName}",
    "description": "Build something amazing with Next.js"
  },
  "Auth": {
    "login": "Sign In",
    "logout": "Sign Out",
    "welcome": "Welcome back, {name}!"
  },
  "Dashboard": {
    "title": "Dashboard",
    "stats": {
      "users": "Total Users",
      "revenue": "Revenue",
      "orders": "Orders"
    },
    "items": "{count, plural, =0 {No items} one {# item} other {# items}}",
    "lastUpdated": "Last updated {date, date, medium} at {date, time, short}"
  },
  "Common": {
    "save": "Save",
    "cancel": "Cancel",
    "delete": "Delete",
    "loading": "Loading...",
    "error": "Something went wrong"
  }
}
```

```json
// src/messages/es.json
{
  "Navigation": {
    "home": "Inicio",
    "about": "Acerca de",
    "dashboard": "Panel",
    "settings": "Configuración"
  },
  "HomePage": {
    "title": "Bienvenido a {appName}",
    "description": "Construye algo increíble con Next.js"
  },
  "Auth": {
    "login": "Iniciar Sesión",
    "logout": "Cerrar Sesión",
    "welcome": "¡Bienvenido de nuevo, {name}!"
  },
  "Dashboard": {
    "title": "Panel",
    "stats": {
      "users": "Usuarios Totales",
      "revenue": "Ingresos",
      "orders": "Pedidos"
    },
    "items": "{count, plural, =0 {Sin elementos} one {# elemento} other {# elementos}}",
    "lastUpdated": "Última actualización {date, date, medium} a las {date, time, short}"
  },
  "Common": {
    "save": "Guardar",
    "cancel": "Cancelar",
    "delete": "Eliminar",
    "loading": "Cargando...",
    "error": "Algo salió mal"
  }
}
```

## Layout & Root Setup

```tsx
// src/app/[locale]/layout.tsx
import { NextIntlClientProvider } from "next-intl";
import { getMessages, setRequestLocale } from "next-intl/server";
import { routing } from "@/i18n/routing";
import { notFound } from "next/navigation";

// Generate static params for all locales
export function generateStaticParams() {
  return routing.locales.map((locale) => ({ locale }));
}

export default async function LocaleLayout({
  children,
  params,
}: {
  children: React.ReactNode;
  params: Promise<{ locale: string }>;
}) {
  const { locale } = await params;

  // Validate locale
  if (!routing.locales.includes(locale as any)) {
    notFound();
  }

  // Enable static rendering
  setRequestLocale(locale);

  // Load messages for client components
  const messages = await getMessages();

  return (
    <html lang={locale} dir={locale === "ar" ? "rtl" : "ltr"}>
      <body>
        <NextIntlClientProvider messages={messages}>
          {children}
        </NextIntlClientProvider>
      </body>
    </html>
  );
}
```

## Server Components (Zero Bundle Impact)

```tsx
// src/app/[locale]/page.tsx (Server Component — default)
import { useTranslations, setRequestLocale } from "next-intl/server";
import { Link } from "@/i18n/navigation";

export default async function HomePage({
  params,
}: {
  params: Promise<{ locale: string }>;
}) {
  const { locale } = await params;
  setRequestLocale(locale);

  const t = useTranslations("HomePage");

  return (
    <main>
      <h1>{t("title", { appName: "My App" })}</h1>
      <p>{t("description")}</p>
      <Link href="/dashboard">{t("goToDashboard")}</Link>
    </main>
  );
}
```

## Client Components

```tsx
// src/components/LanguageSwitcher.tsx
"use client";

import { useLocale } from "next-intl";
import { useRouter, usePathname } from "@/i18n/navigation";
import { routing } from "@/i18n/routing";

const labels: Record<string, string> = {
  en: "English",
  es: "Español",
  fr: "Français",
  de: "Deutsch",
  ja: "日本語",
};

export function LanguageSwitcher() {
  const locale = useLocale();
  const router = useRouter();
  const pathname = usePathname();

  function onChange(newLocale: string) {
    router.replace(pathname, { locale: newLocale });
  }

  return (
    <select value={locale} onChange={(e) => onChange(e.target.value)}>
      {routing.locales.map((loc) => (
        <option key={loc} value={loc}>
          {labels[loc] || loc}
        </option>
      ))}
    </select>
  );
}
```

```tsx
// src/components/ItemCount.tsx
"use client";

import { useTranslations } from "next-intl";

export function ItemCount({ count }: { count: number }) {
  const t = useTranslations("Dashboard");

  return (
    <div>
      {/* ICU plural rules */}
      <p>{t("items", { count })}</p>
      {/* count=0: "No items" */}
      {/* count=1: "1 item" */}
      {/* count=5: "5 items" */}
    </div>
  );
}
```

## ICU Message Format

```json
{
  "greeting": "Hello, {name}!",

  "items": "{count, plural, =0 {No items} one {# item} other {# items}}",

  "gender": "{gender, select, male {He} female {She} other {They}} liked your post",

  "richText": "Please <link>read the docs</link> before continuing",

  "ordinal": "{position, selectordinal, one {#st} two {#nd} few {#rd} other {#th}} place",

  "nested": "{count, plural, one {You have # new {count, select, 1 {notification} other {notification}}} other {You have # new notifications}}"
}
```

```tsx
// Using rich text with components
const t = useTranslations("Onboarding");

return (
  <p>
    {t.rich("richText", {
      link: (chunks) => <a href="/docs">{chunks}</a>,
    })}
  </p>
);
```

## Date, Time & Number Formatting

```tsx
import { useFormatter } from "next-intl";

function StatsCard({ revenue, lastUpdated }: { revenue: number; lastUpdated: Date }) {
  const format = useFormatter();

  return (
    <div>
      {/* Currency */}
      <p>{format.number(revenue, { style: "currency", currency: "USD" })}</p>
      {/* en: $1,234.56 | de: 1.234,56 $ | ja: ¥1,235 */}

      {/* Date */}
      <p>{format.dateTime(lastUpdated, { dateStyle: "medium" })}</p>
      {/* en: Mar 19, 2026 | es: 19 mar 2026 | ja: 2026/03/19 */}

      {/* Relative time */}
      <p>{format.relativeTime(lastUpdated)}</p>
      {/* en: 3 hours ago | es: hace 3 horas */}

      {/* Number */}
      <p>{format.number(1234567)}</p>
      {/* en: 1,234,567 | de: 1.234.567 */}

      {/* List */}
      <p>{format.list(["Alice", "Bob", "Charlie"], { type: "conjunction" })}</p>
      {/* en: Alice, Bob, and Charlie | es: Alice, Bob y Charlie */}
    </div>
  );
}
```

## Type Safety

```bash
# Generate types from translation files
npx next-intl typegen ./src/messages/en.json --out ./src/types/messages.d.ts
```

```typescript
// src/types/messages.d.ts (auto-generated)
// This file is auto-generated by next-intl
type Messages = typeof import("../messages/en.json");
declare interface IntlMessages extends Messages {}
```

```typescript
// global.d.ts
import en from "./messages/en.json";

type Messages = typeof en;

declare global {
  interface IntlMessages extends Messages {}
}
```

## Metadata & SEO

```tsx
// src/app/[locale]/layout.tsx
import { getTranslations, setRequestLocale } from "next-intl/server";
import { routing } from "@/i18n/routing";

export async function generateMetadata({
  params,
}: {
  params: Promise<{ locale: string }>;
}) {
  const { locale } = await params;
  const t = await getTranslations({ locale, namespace: "Metadata" });

  return {
    title: t("title"),
    description: t("description"),
    alternates: {
      canonical: `/${locale}`,
      languages: Object.fromEntries(
        routing.locales.map((l) => [l, `/${l}`])
      ),
    },
    openGraph: {
      title: t("title"),
      description: t("description"),
      locale,
    },
  };
}
```

## Gotchas

1. **`useTranslations` is different in server vs client** — In server components, import from `next-intl/server`. In client components (`"use client"`), import from `next-intl`. Using the wrong import silently fails.

2. **`setRequestLocale` is required for static rendering** — Without it, pages using translations can't be statically generated. Call it at the top of every page and layout that uses `useTranslations`.

3. **`NextIntlClientProvider` is required for client components** — Server components can use translations directly. Client components need the provider in a parent layout with `messages` passed down.

4. **ICU format uses `#` for the count value** — In `"{count, plural, one {# item} other {# items}}"`, `#` is replaced by the formatted count. Don't use `{count}` inside the plural branches.

5. **Middleware must exclude API routes and static files** — Without the `matcher` config, the middleware will try to add locale prefixes to `/api/`, `/_next/`, and static files, causing 404s.

6. **`localePrefix: "as-needed"` can cause hydration mismatches** — The server renders `/` for the default locale, but the client might expect `/en`. Use consistent paths and test with `localePrefix: "always"` if you see hydration errors.
