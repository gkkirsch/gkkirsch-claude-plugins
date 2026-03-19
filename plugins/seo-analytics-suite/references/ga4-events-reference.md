# Google Analytics 4 — Event Reference

## GA4 Event Types

### Automatically Collected Events

These are tracked by GA4 without any code. Just install the tag.

| Event | Trigger | Parameters |
|-------|---------|------------|
| `first_visit` | First time user visits | — |
| `session_start` | New session begins | — |
| `page_view` | Page loads | `page_location`, `page_referrer`, `page_title` |
| `user_engagement` | App/page in foreground ≥1s | `engagement_time_msec` |
| `scroll` | User scrolls ≥90% of page | — |
| `click` | Outbound link clicked | `link_url`, `link_domain`, `link_text`, `outbound` |
| `view_search_results` | Site search used | `search_term` |
| `file_download` | File link clicked | `file_name`, `file_extension`, `link_url` |
| `video_start` | YouTube video starts | `video_title`, `video_url`, `video_percent` |
| `video_progress` | Video reaches 10/25/50/75% | `video_title`, `video_percent` |
| `video_complete` | Video reaches end | `video_title`, `video_url` |
| `form_start` | User interacts with form | `form_id`, `form_name` |
| `form_submit` | Form submitted | `form_id`, `form_name`, `form_submit_text` |

### Enhanced Measurement Events

Enabled in GA4 admin under Data Streams > Enhanced Measurement. No code needed.

| Feature | Events Tracked |
|---------|---------------|
| Page views | `page_view` on every navigation |
| Scrolls | `scroll` at 90% depth |
| Outbound clicks | `click` on external links |
| Site search | `view_search_results` (configure URL parameter) |
| Video engagement | `video_start`, `video_progress`, `video_complete` (YouTube only) |
| File downloads | `file_download` for common file types |
| Form interactions | `form_start`, `form_submit` |

### Recommended Events

Google recommends these event names. Using them enables built-in GA4 reports.

#### E-Commerce Events (Full Flow)

```javascript
// 1. View item list (category page)
gtag('event', 'view_item_list', {
  item_list_id: 'related_products',
  item_list_name: 'Related Products',
  items: [{
    item_id: 'SKU_12345',
    item_name: 'Product Name',
    item_category: 'Category',
    price: 29.99,
    quantity: 1,
    index: 0,
  }],
});

// 2. Select item (click on product)
gtag('event', 'select_item', {
  item_list_id: 'related_products',
  items: [{ item_id: 'SKU_12345', item_name: 'Product Name', price: 29.99 }],
});

// 3. View item (product detail page)
gtag('event', 'view_item', {
  currency: 'USD',
  value: 29.99,
  items: [{
    item_id: 'SKU_12345',
    item_name: 'Product Name',
    item_brand: 'Brand',
    item_category: 'Category',
    item_category2: 'Subcategory',
    item_variant: 'Blue',
    price: 29.99,
    quantity: 1,
  }],
});

// 4. Add to cart
gtag('event', 'add_to_cart', {
  currency: 'USD',
  value: 29.99,
  items: [{ item_id: 'SKU_12345', item_name: 'Product Name', price: 29.99, quantity: 1 }],
});

// 5. View cart
gtag('event', 'view_cart', {
  currency: 'USD',
  value: 59.98,
  items: [
    { item_id: 'SKU_12345', item_name: 'Product A', price: 29.99, quantity: 1 },
    { item_id: 'SKU_67890', item_name: 'Product B', price: 29.99, quantity: 1 },
  ],
});

// 6. Begin checkout
gtag('event', 'begin_checkout', {
  currency: 'USD',
  value: 59.98,
  coupon: 'SUMMER20',
  items: [/* same as view_cart */],
});

// 7. Add shipping info
gtag('event', 'add_shipping_info', {
  currency: 'USD',
  value: 59.98,
  shipping_tier: 'Standard',
  items: [/* ... */],
});

// 8. Add payment info
gtag('event', 'add_payment_info', {
  currency: 'USD',
  value: 59.98,
  payment_type: 'Credit Card',
  items: [/* ... */],
});

// 9. Purchase
gtag('event', 'purchase', {
  transaction_id: 'T_12345',
  value: 64.98,
  tax: 5.00,
  shipping: 0,
  currency: 'USD',
  coupon: 'SUMMER20',
  items: [/* ... */],
});

// 10. Refund (partial or full)
gtag('event', 'refund', {
  transaction_id: 'T_12345',
  value: 29.99,
  currency: 'USD',
  items: [{ item_id: 'SKU_12345', quantity: 1 }], // Omit items for full refund
});
```

#### SaaS / Subscription Events

```javascript
// Signup
gtag('event', 'sign_up', {
  method: 'email',  // or 'google', 'github', etc.
});

// Login
gtag('event', 'login', {
  method: 'email',
});

// Trial start
gtag('event', 'begin_checkout', {
  currency: 'USD',
  value: 0,
  items: [{ item_id: 'trial', item_name: 'Free Trial', price: 0 }],
});

// Subscription purchase
gtag('event', 'purchase', {
  transaction_id: 'SUB_12345',
  value: 29,
  currency: 'USD',
  items: [{
    item_id: 'pro_monthly',
    item_name: 'Pro Plan (Monthly)',
    price: 29,
    quantity: 1,
  }],
});

// Feature usage (custom event)
gtag('event', 'feature_used', {
  feature_name: 'export_csv',
  plan_type: 'pro',
});

// Tutorial completed
gtag('event', 'tutorial_complete', {
  tutorial_name: 'onboarding_flow',
});

// Upgrade
gtag('event', 'purchase', {
  transaction_id: 'UPG_12345',
  value: 20,  // difference in price
  currency: 'USD',
  items: [{
    item_id: 'enterprise_monthly',
    item_name: 'Enterprise Plan (Monthly)',
    price: 49,
    quantity: 1,
  }],
});
```

#### Content / Lead Gen Events

```javascript
// Generate lead
gtag('event', 'generate_lead', {
  currency: 'USD',
  value: 50,  // estimated lead value
  lead_source: 'contact_form',
});

// Share
gtag('event', 'share', {
  method: 'twitter',
  content_type: 'article',
  item_id: 'blog-post-123',
});

// Search
gtag('event', 'search', {
  search_term: 'user query',
});

// Select content (e.g., tab, accordion, filter)
gtag('event', 'select_content', {
  content_type: 'tab',
  content_id: 'pricing_annual',
});
```

---

## Custom Events

For events not covered by recommended events, use custom names:

```javascript
// Custom event naming convention: snake_case, verb_noun
gtag('event', 'copy_code_snippet', {
  snippet_language: 'javascript',
  page_section: 'tutorial',
});

gtag('event', 'toggle_dark_mode', {
  new_state: 'dark',
});

gtag('event', 'expand_faq', {
  question_id: 'pricing-q3',
  question_text: 'Is there a free plan?',
});

gtag('event', 'invite_teammate', {
  invite_method: 'email',
  team_size: 5,
});

gtag('event', 'api_error', {
  error_code: 429,
  endpoint: '/api/users',
  error_message: 'Rate limit exceeded',
});
```

---

## User Properties

```javascript
// Set user properties for segmentation
gtag('set', 'user_properties', {
  plan_type: 'pro',           // Free, Pro, Enterprise
  signup_date: '2024-01-15',
  company_size: '11-50',
  account_age_days: 45,
  is_paying: true,
  feature_flags: 'dark_mode,beta_export',
});

// Set user ID for cross-device tracking
gtag('config', 'G-XXXXXXXXXX', {
  user_id: 'USER_12345',  // Your internal user ID (NOT PII)
});
```

---

## Consent Mode v2

```javascript
// Default: deny everything until user consents
gtag('consent', 'default', {
  analytics_storage: 'denied',    // GA cookies
  ad_storage: 'denied',           // Ads cookies
  ad_user_data: 'denied',         // Ads user data
  ad_personalization: 'denied',   // Ads personalization
  functionality_storage: 'denied', // Functional cookies
  personalization_storage: 'denied', // Personalization cookies
  security_storage: 'granted',    // Always OK (security)
  wait_for_update: 500,           // Wait 500ms for consent
});

// After user consents (e.g., clicks "Accept All")
gtag('consent', 'update', {
  analytics_storage: 'granted',
  ad_storage: 'granted',
  ad_user_data: 'granted',
  ad_personalization: 'granted',
});

// After user accepts only necessary
gtag('consent', 'update', {
  analytics_storage: 'denied',
  ad_storage: 'denied',
  ad_user_data: 'denied',
  ad_personalization: 'denied',
});
```

**What happens when consent is denied:**
- GA4 still receives pings (cookieless, for modeling)
- No cookies are set
- No user-level data is stored
- Google uses modeling to estimate denied traffic

---

## GA4 Configuration Options

```javascript
gtag('config', 'G-XXXXXXXXXX', {
  // Page tracking
  send_page_view: true,           // Auto page_view on config (default: true)
  page_title: document.title,     // Override page title
  page_location: window.location.href, // Override URL

  // Session
  session_timeout: 1800,          // Session timeout in seconds (default: 30min)

  // Content grouping
  content_group: 'Blog',          // Group pages for reporting

  // Cookie settings
  cookie_domain: 'auto',          // Cookie domain
  cookie_expires: 63072000,       // Cookie lifetime in seconds (default: 2 years)
  cookie_prefix: '_ga',           // Cookie prefix
  cookie_flags: 'SameSite=None;Secure', // Cookie flags

  // Debug
  debug_mode: true,               // Enable DebugView in GA4

  // Cross-domain
  linker: {
    domains: ['example.com', 'shop.example.com'],
    accept_incoming: true,
  },

  // User ID
  user_id: 'USER_12345',
});
```

---

## Custom Dimensions & Metrics

Configure in GA4 Admin > Custom Definitions before sending data.

```javascript
// After configuring custom dimensions in GA4 admin:
gtag('event', 'page_view', {
  // Custom dimensions (must match admin config)
  user_plan: 'pro',              // User-scoped dimension
  page_author: 'John Doe',       // Event-scoped dimension
  content_category: 'Tutorial',  // Event-scoped dimension

  // Custom metrics
  word_count: 2500,              // Event-scoped metric
  time_to_first_interaction: 3.2, // Event-scoped metric
});
```

---

## Debugging GA4

### Tools

| Tool | How to Access | What It Shows |
|------|--------------|---------------|
| GA4 DebugView | GA4 Admin > DebugView | Real-time events with all parameters |
| Chrome Tag Assistant | chrome.google.com/webstore | Tag firing, errors, data layer |
| GA Debugger Extension | Chrome Web Store | Console logging of all GA calls |
| Network tab | Chrome DevTools > Network | Raw collect requests to google-analytics.com |
| dataLayer inspection | Console: `window.dataLayer` | All events pushed to GTM/GA4 |

### Debug Mode

```javascript
// Enable debug mode for the current page
gtag('config', 'G-XXXXXXXXXX', { debug_mode: true });

// Or via URL parameter (add to any page URL):
// https://example.com/page?gtm_debug=true

// Or via Chrome extension: "Google Analytics Debugger"
```

### Common Issues

```
1. Events not appearing in GA4
   → Check: Is the GA4 tag firing? (Network tab, filter "collect")
   → Check: Is the Measurement ID correct? (G-XXXXXXXXXX)
   → Check: Is consent mode blocking? (analytics_storage: 'denied')
   → Check: Are you in a development/localhost environment? (add debug_mode: true)
   → Wait: GA4 reports take 24-48 hours to populate. Use Realtime for instant check.

2. Page views not tracking on SPA navigation
   → You need manual page_view tracking on route changes
   → Use the AnalyticsProvider pattern from the analytics-integration skill

3. Events showing but parameters missing
   → Check: Did you register custom dimensions in GA4 Admin > Custom Definitions?
   → Unregistered parameters are still collected but won't appear in reports

4. Conversions not counting
   → Check: Did you mark the event as a conversion in GA4 Admin > Events?
   → Any event can be marked as a conversion — toggle it on

5. User ID not linking sessions
   → Set user_id BEFORE sending events (in the config call, not after)
   → user_id must be consistent across devices (use your internal ID)
```

---

## Measurement Protocol (Server-Side Events)

```javascript
// Send events server-side (no client JS needed)
const MEASUREMENT_ID = 'G-XXXXXXXXXX';
const API_SECRET = 'your_api_secret'; // GA4 Admin > Data Streams > Measurement Protocol

async function sendServerEvent(clientId, events) {
  await fetch(
    `https://www.google-analytics.com/mp/collect?measurement_id=${MEASUREMENT_ID}&api_secret=${API_SECRET}`,
    {
      method: 'POST',
      body: JSON.stringify({
        client_id: clientId,  // From _ga cookie or generated
        events: events,
      }),
    }
  );
}

// Usage: track server-side purchase
await sendServerEvent('GA1.1.12345.67890', [{
  name: 'purchase',
  params: {
    transaction_id: 'T_12345',
    value: 29.99,
    currency: 'USD',
    items: [{ item_id: 'PRO_MONTHLY', item_name: 'Pro Plan' }],
  },
}]);

// Usage: track webhook event
await sendServerEvent(clientId, [{
  name: 'subscription_renewed',
  params: {
    plan: 'pro',
    value: 29,
    renewal_count: 3,
  },
}]);
```

---

## GA4 Naming Conventions

```
Events:     snake_case, max 40 characters
            verb_noun pattern: sign_up, add_to_cart, begin_checkout
            ✅ cta_click, feature_used, form_error
            ❌ CTAClick, FeatureUsed, formError

Parameters: snake_case, max 40 characters
            ✅ button_text, error_type, plan_name
            ❌ buttonText, ErrorType

Values:     No PII (emails, phone numbers, names)
            ✅ method: 'google', plan: 'pro'
            ❌ email: 'user@example.com'

Limits:
  - 500 distinct event names per property
  - 50 custom dimensions per property
  - 50 custom metrics per property
  - 25 parameters per event
  - 40 characters per event name
  - 40 characters per parameter name
  - 100 characters per parameter value
```
