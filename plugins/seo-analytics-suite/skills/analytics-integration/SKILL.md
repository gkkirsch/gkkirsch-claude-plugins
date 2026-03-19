---
name: analytics-integration
description: >
  Set up web analytics and event tracking for any web application. Covers Google Analytics 4 (GA4),
  privacy-compliant tracking, conversion funnels, custom event taxonomies, React/Next.js
  integration, and server-side analytics. Includes cookie consent implementation.
  Use when setting up analytics for a new project, adding event tracking, implementing
  conversion tracking, or building privacy-compliant measurement.
version: 1.0.0
argument-hint: "[framework-or-tool]"
allowed-tools: Read, Grep, Glob, Bash, Write, Edit
model: sonnet
---

# Analytics Integration

Set up web analytics for any project. This skill covers the full analytics stack: GA4 setup, custom event tracking, conversion funnels, privacy/consent, and framework-specific integrations.

## Step 1: Choose Your Analytics Approach

```
Decision matrix:

Need                           → Tool
─────────────────────────────────────────────────
Standard web analytics         → Google Analytics 4 (GA4)
Privacy-first (no cookies)     → Plausible, Fathom, or self-hosted
Full control / self-hosted     → Custom (PostHog, Umami, or roll your own)
E-commerce tracking            → GA4 + Enhanced E-commerce
Marketing attribution          → GA4 + UTM parameters
Product analytics              → PostHog, Mixpanel, or Amplitude

This skill focuses on GA4 (most common) and custom/privacy-first approaches.
```

## Step 2: GA4 Installation

### Basic Setup

```html
<!-- Add to <head> on every page -->
<script async src="https://www.googletagmanager.com/gtag/js?id=G-XXXXXXXXXX"></script>
<script>
  window.dataLayer = window.dataLayer || [];
  function gtag(){dataLayer.push(arguments);}

  // Default consent state (GDPR-compliant)
  gtag('consent', 'default', {
    analytics_storage: 'denied',
    ad_storage: 'denied',
    ad_user_data: 'denied',
    ad_personalization: 'denied',
    wait_for_update: 500,
  });

  gtag('js', new Date());
  gtag('config', 'G-XXXXXXXXXX', {
    send_page_view: true,
  });
</script>
```

### Next.js Setup (App Router)

```typescript
// app/layout.tsx
import Script from 'next/script';

const GA_ID = process.env.NEXT_PUBLIC_GA_ID;

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <head>
        {GA_ID && (
          <>
            <Script
              src={`https://www.googletagmanager.com/gtag/js?id=${GA_ID}`}
              strategy="afterInteractive"
            />
            <Script id="ga4-init" strategy="afterInteractive">
              {`
                window.dataLayer = window.dataLayer || [];
                function gtag(){dataLayer.push(arguments);}
                gtag('consent', 'default', {
                  analytics_storage: 'denied',
                  ad_storage: 'denied',
                  ad_user_data: 'denied',
                  ad_personalization: 'denied',
                  wait_for_update: 500,
                });
                gtag('js', new Date());
                gtag('config', '${GA_ID}');
              `}
            </Script>
          </>
        )}
      </head>
      <body>
        <AnalyticsProvider>{children}</AnalyticsProvider>
      </body>
    </html>
  );
}
```

### Route Change Tracking (SPA)

```typescript
// components/AnalyticsProvider.tsx
'use client';

import { usePathname, useSearchParams } from 'next/navigation';
import { useEffect } from 'react';

declare global {
  interface Window {
    gtag: (...args: any[]) => void;
  }
}

export function AnalyticsProvider({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const searchParams = useSearchParams();

  useEffect(() => {
    if (!window.gtag) return;

    const url = pathname + (searchParams?.toString() ? `?${searchParams}` : '');
    window.gtag('config', process.env.NEXT_PUBLIC_GA_ID!, {
      page_path: url,
    });
  }, [pathname, searchParams]);

  return <>{children}</>;
}
```

## Step 3: Event Tracking System

### Event Taxonomy

```typescript
// lib/analytics/events.ts

// Define all tracked events in one place
export const EVENTS = {
  // Navigation
  PAGE_VIEW: 'page_view',

  // Authentication
  LOGIN: 'login',
  SIGNUP: 'sign_up',
  LOGOUT: 'logout',

  // Engagement
  CTA_CLICK: 'cta_click',
  FEATURE_USED: 'feature_used',
  SEARCH: 'search',
  SHARE: 'share',

  // Content
  SCROLL_DEPTH: 'scroll_depth',
  VIDEO_PLAY: 'video_play',
  VIDEO_COMPLETE: 'video_complete',
  FILE_DOWNLOAD: 'file_download',

  // Conversion funnel
  FUNNEL_STEP: 'funnel_step',
  PRICING_VIEW: 'pricing_view',
  PLAN_SELECTED: 'plan_selected',
  CHECKOUT_START: 'checkout_start',
  PURCHASE: 'purchase',

  // Errors
  ERROR: 'error',
  FORM_ERROR: 'form_error',

  // Feedback
  NPS_RESPONSE: 'nps_response',
  FEEDBACK_SUBMIT: 'feedback_submit',
} as const;

export type EventName = (typeof EVENTS)[keyof typeof EVENTS];
```

### Analytics Hook (React)

```typescript
// lib/analytics/useAnalytics.ts
import { useCallback } from 'react';
import { EVENTS, type EventName } from './events';

type EventParams = Record<string, string | number | boolean | undefined>;

export function useAnalytics() {
  const track = useCallback((event: EventName, params?: EventParams) => {
    // Filter out undefined values
    const cleanParams = params
      ? Object.fromEntries(Object.entries(params).filter(([, v]) => v !== undefined))
      : {};

    // GA4
    if (typeof window !== 'undefined' && window.gtag) {
      window.gtag('event', event, cleanParams);
    }

    // Development logging
    if (process.env.NODE_ENV === 'development') {
      console.log(`[Analytics] ${event}`, cleanParams);
    }
  }, []);

  const trackClick = useCallback((
    buttonId: string,
    buttonText: string,
    extra?: EventParams
  ) => {
    track(EVENTS.CTA_CLICK, {
      button_id: buttonId,
      button_text: buttonText,
      page_path: window.location.pathname,
      ...extra,
    });
  }, [track]);

  const trackConversion = useCallback((
    type: string,
    value?: number,
    params?: EventParams
  ) => {
    track(type as EventName, {
      value: value ?? 0,
      currency: 'USD',
      ...params,
    });
  }, [track]);

  const trackError = useCallback((
    errorType: string,
    errorMessage: string,
    extra?: EventParams
  ) => {
    track(EVENTS.ERROR, {
      error_type: errorType,
      error_message: errorMessage,
      page_path: window.location.pathname,
      ...extra,
    });
  }, [track]);

  return { track, trackClick, trackConversion, trackError, EVENTS };
}
```

### Usage Examples

```tsx
function SignupForm() {
  const { track, trackConversion, trackError, EVENTS } = useAnalytics();

  const handleSubmit = async (data: FormData) => {
    track(EVENTS.FUNNEL_STEP, {
      funnel_name: 'signup',
      step_name: 'form_submit',
      step_number: 2,
    });

    try {
      const result = await createAccount(data);
      trackConversion(EVENTS.SIGNUP, 0, { method: 'email' });
    } catch (error) {
      trackError('signup_error', error.message, { step: 'form_submit' });
    }
  };

  return (
    <form onSubmit={handleSubmit}>
      <input
        onFocus={() => track(EVENTS.FUNNEL_STEP, {
          funnel_name: 'signup',
          step_name: 'email_focus',
          step_number: 1,
        })}
        name="email"
        type="email"
      />
      <button type="submit">Create Account</button>
    </form>
  );
}

function PricingPage() {
  const { track, EVENTS } = useAnalytics();

  useEffect(() => {
    track(EVENTS.PRICING_VIEW, { source: 'navigation' });
  }, []);

  return (
    <div>
      <button onClick={() => {
        track(EVENTS.PLAN_SELECTED, { plan: 'pro', price: 29 });
        router.push('/checkout?plan=pro');
      }}>
        Choose Pro
      </button>
    </div>
  );
}
```

## Step 4: Scroll Depth Tracking

```typescript
// lib/analytics/useScrollDepth.ts
import { useEffect, useRef } from 'react';
import { useAnalytics } from './useAnalytics';

export function useScrollDepth() {
  const { track, EVENTS } = useAnalytics();
  const tracked = useRef(new Set<number>());

  useEffect(() => {
    const thresholds = [25, 50, 75, 90, 100];

    const handleScroll = () => {
      const scrollHeight = document.documentElement.scrollHeight - window.innerHeight;
      if (scrollHeight <= 0) return;

      const scrollPercent = Math.round((window.scrollY / scrollHeight) * 100);

      for (const threshold of thresholds) {
        if (scrollPercent >= threshold && !tracked.current.has(threshold)) {
          tracked.current.add(threshold);
          track(EVENTS.SCROLL_DEPTH, {
            depth: threshold,
            page_path: window.location.pathname,
          });
        }
      }
    };

    window.addEventListener('scroll', handleScroll, { passive: true });
    return () => window.removeEventListener('scroll', handleScroll);
  }, [track]);
}
```

## Step 5: E-Commerce Event Tracking

```typescript
// GA4 Enhanced E-commerce events

// View item
function trackProductView(product: Product) {
  gtag('event', 'view_item', {
    currency: 'USD',
    value: product.price,
    items: [{
      item_id: product.id,
      item_name: product.name,
      item_category: product.category,
      price: product.price,
      quantity: 1,
    }],
  });
}

// Add to cart
function trackAddToCart(product: Product, quantity: number) {
  gtag('event', 'add_to_cart', {
    currency: 'USD',
    value: product.price * quantity,
    items: [{
      item_id: product.id,
      item_name: product.name,
      item_category: product.category,
      price: product.price,
      quantity,
    }],
  });
}

// Begin checkout
function trackCheckoutStart(cart: CartItem[]) {
  gtag('event', 'begin_checkout', {
    currency: 'USD',
    value: cart.reduce((sum, item) => sum + item.price * item.quantity, 0),
    items: cart.map(item => ({
      item_id: item.id,
      item_name: item.name,
      price: item.price,
      quantity: item.quantity,
    })),
  });
}

// Purchase complete
function trackPurchase(order: Order) {
  gtag('event', 'purchase', {
    transaction_id: order.id,
    value: order.total,
    currency: 'USD',
    tax: order.tax,
    shipping: order.shipping,
    items: order.items.map(item => ({
      item_id: item.id,
      item_name: item.name,
      item_category: item.category,
      price: item.price,
      quantity: item.quantity,
    })),
  });
}
```

## Step 6: Cookie Consent Banner

```tsx
// components/CookieConsent.tsx
'use client';

import { useState, useEffect } from 'react';

type ConsentCategories = {
  analytics: boolean;
  marketing: boolean;
};

const CONSENT_KEY = 'cookie_consent';
const CONSENT_VERSION = '1.0';

export function CookieConsent() {
  const [showBanner, setShowBanner] = useState(false);

  useEffect(() => {
    const stored = localStorage.getItem(CONSENT_KEY);
    if (!stored) {
      setShowBanner(true);
      return;
    }
    try {
      const consent = JSON.parse(stored);
      if (consent.version !== CONSENT_VERSION) {
        setShowBanner(true);
        return;
      }
      // Apply stored consent
      updateGtagConsent(consent.categories);
    } catch {
      setShowBanner(true);
    }
  }, []);

  const acceptAll = () => {
    saveConsent({ analytics: true, marketing: true });
    setShowBanner(false);
  };

  const acceptNecessary = () => {
    saveConsent({ analytics: false, marketing: false });
    setShowBanner(false);
  };

  const saveConsent = (categories: ConsentCategories) => {
    localStorage.setItem(CONSENT_KEY, JSON.stringify({
      categories,
      version: CONSENT_VERSION,
      timestamp: new Date().toISOString(),
    }));
    updateGtagConsent(categories);
  };

  const updateGtagConsent = (categories: ConsentCategories) => {
    if (typeof window.gtag === 'undefined') return;
    window.gtag('consent', 'update', {
      analytics_storage: categories.analytics ? 'granted' : 'denied',
      ad_storage: categories.marketing ? 'granted' : 'denied',
      ad_user_data: categories.marketing ? 'granted' : 'denied',
      ad_personalization: categories.marketing ? 'granted' : 'denied',
    });
  };

  if (!showBanner) return null;

  return (
    <div role="dialog" aria-label="Cookie consent" className="cookie-banner">
      <p>
        We use cookies to analyze site usage and improve your experience.
      </p>
      <div className="cookie-actions">
        <button onClick={acceptNecessary}>
          Necessary only
        </button>
        <button onClick={acceptAll}>
          Accept all
        </button>
      </div>
    </div>
  );
}
```

## Step 7: UTM Parameter Tracking

```typescript
// lib/analytics/utm.ts

interface UTMParams {
  utm_source?: string;
  utm_medium?: string;
  utm_campaign?: string;
  utm_term?: string;
  utm_content?: string;
}

export function captureUTMParams(): UTMParams {
  if (typeof window === 'undefined') return {};

  const params = new URLSearchParams(window.location.search);
  const utm: UTMParams = {};

  for (const key of ['utm_source', 'utm_medium', 'utm_campaign', 'utm_term', 'utm_content'] as const) {
    const value = params.get(key);
    if (value) utm[key] = value;
  }

  if (Object.keys(utm).length > 0) {
    // Store for attribution
    sessionStorage.setItem('utm_params', JSON.stringify(utm));

    // Set as user properties in GA4
    if (window.gtag) {
      window.gtag('set', 'user_properties', {
        first_utm_source: utm.utm_source,
        first_utm_medium: utm.utm_medium,
        first_utm_campaign: utm.utm_campaign,
      });
    }
  }

  return utm;
}

// UTM link builder
export function buildUTMUrl(
  baseUrl: string,
  source: string,
  medium: string,
  campaign: string,
  extra?: { term?: string; content?: string }
): string {
  const url = new URL(baseUrl);
  url.searchParams.set('utm_source', source);
  url.searchParams.set('utm_medium', medium);
  url.searchParams.set('utm_campaign', campaign);
  if (extra?.term) url.searchParams.set('utm_term', extra.term);
  if (extra?.content) url.searchParams.set('utm_content', extra.content);
  return url.toString();
}

// Examples:
// buildUTMUrl('https://example.com', 'twitter', 'social', 'launch-2024')
// → https://example.com?utm_source=twitter&utm_medium=social&utm_campaign=launch-2024
```

## Step 8: Privacy-First Analytics (No Cookies)

```typescript
// lib/analytics/privacy-analytics.ts
// Server-side analytics that respects privacy — no cookies, no PII

interface AnalyticsEvent {
  event: string;
  path: string;
  referrer: string;
  params?: Record<string, string | number>;
}

// Client-side beacon
export function trackPrivate(event: string, params?: Record<string, string | number>) {
  const payload: AnalyticsEvent = {
    event,
    path: window.location.pathname,
    referrer: document.referrer,
    params,
  };

  // Use sendBeacon for reliability (doesn't block page unload)
  if (navigator.sendBeacon) {
    navigator.sendBeacon('/api/analytics', JSON.stringify(payload));
  } else {
    fetch('/api/analytics', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
      keepalive: true,
    }).catch(() => {});
  }
}
```

```typescript
// Server-side handler (Express or Next.js API route)
import crypto from 'crypto';

// Next.js API Route
export async function POST(request: Request) {
  const body = await request.json();
  const ip = request.headers.get('x-forwarded-for') || 'unknown';
  const userAgent = request.headers.get('user-agent') || '';

  // Anonymize: hash IP with daily-rotating salt
  const dailySalt = new Date().toISOString().slice(0, 10);
  const visitorHash = crypto
    .createHash('sha256')
    .update(ip + dailySalt + userAgent)
    .digest('hex')
    .slice(0, 16);

  const event = {
    event: body.event,
    path: body.path,
    referrer_domain: body.referrer ? new URL(body.referrer).hostname : 'direct',
    visitor_hash: visitorHash,
    country: request.headers.get('cf-ipcountry') || 'unknown',
    device_type: /Mobile|Android|iPhone/i.test(userAgent) ? 'mobile' : 'desktop',
    browser: detectBrowser(userAgent),
    timestamp: new Date().toISOString(),
    params: body.params || {},
  };

  // Store in your database
  await db.analytics.create({ data: event });

  return new Response(null, { status: 204 });
}

function detectBrowser(ua: string): string {
  if (ua.includes('Firefox')) return 'Firefox';
  if (ua.includes('Chrome')) return 'Chrome';
  if (ua.includes('Safari')) return 'Safari';
  if (ua.includes('Edge')) return 'Edge';
  return 'Other';
}
```

## Step 9: Testing Analytics

### Debug Mode

```javascript
// Enable GA4 debug mode
gtag('config', 'G-XXXXXXXXXX', { debug_mode: true });

// View events in:
// 1. Browser console (development)
// 2. GA4 DebugView (Realtime > DebugView in GA4 admin)
// 3. Chrome extension: "Google Analytics Debugger"
```

### Automated Testing

```typescript
// __tests__/analytics.test.ts
import { renderHook, act } from '@testing-library/react-hooks';
import { useAnalytics } from '@/lib/analytics/useAnalytics';

// Mock gtag
const mockGtag = jest.fn();
Object.defineProperty(window, 'gtag', { value: mockGtag, writable: true });

describe('useAnalytics', () => {
  beforeEach(() => mockGtag.mockClear());

  it('tracks events with correct parameters', () => {
    const { result } = renderHook(() => useAnalytics());

    act(() => {
      result.current.track('cta_click', {
        button_id: 'hero-signup',
        button_text: 'Get Started',
      });
    });

    expect(mockGtag).toHaveBeenCalledWith('event', 'cta_click', {
      button_id: 'hero-signup',
      button_text: 'Get Started',
    });
  });

  it('tracks conversions with value', () => {
    const { result } = renderHook(() => useAnalytics());

    act(() => {
      result.current.trackConversion('purchase', 29.99, { plan: 'pro' });
    });

    expect(mockGtag).toHaveBeenCalledWith('event', 'purchase', {
      value: 29.99,
      currency: 'USD',
      plan: 'pro',
    });
  });

  it('filters undefined params', () => {
    const { result } = renderHook(() => useAnalytics());

    act(() => {
      result.current.track('test_event', {
        defined: 'yes',
        undef: undefined,
      });
    });

    expect(mockGtag).toHaveBeenCalledWith('event', 'test_event', {
      defined: 'yes',
    });
  });
});
```

## Checklist

- [ ] GA4 tag installed and sending data (check GA4 Realtime report)
- [ ] Cookie consent banner implemented (GDPR/CCPA)
- [ ] Default consent state set to 'denied' before user action
- [ ] Route changes tracked in SPA (usePathname + useEffect)
- [ ] Event taxonomy documented (all events in one file)
- [ ] Key conversion events firing (signup, purchase, etc.)
- [ ] Scroll depth tracking on content pages
- [ ] Error tracking implemented
- [ ] UTM parameters captured and stored
- [ ] User properties set for segmentation (plan type, etc.)
- [ ] Debug mode works in development
- [ ] No PII in analytics (no emails, names, etc. in events)
- [ ] Analytics excluded from development/staging
- [ ] E-commerce events match GA4 Enhanced E-commerce spec
