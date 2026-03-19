# Structured Data Schemas — Quick Reference

## Google Rich Results Support Matrix

Which Schema.org types trigger rich results in Google Search:

| Schema Type | Rich Result | Requirements |
|-------------|------------|--------------|
| Article / BlogPosting | Article carousel, author info | headline, image, datePublished, author |
| BreadcrumbList | Breadcrumb trail in SERP | itemListElement with name + item |
| Course | Course listing | name, provider, description |
| Event | Event listing with date/location | name, startDate, location |
| FAQ | Expandable Q&A in SERP | Question + Answer pairs |
| HowTo | Step-by-step instructions | name, step[] |
| JobPosting | Job listing | title, description, datePosted, hiringOrganization |
| LocalBusiness | Knowledge panel, map pack | name, address, telephone |
| Organization | Knowledge panel, logo | name, url, logo |
| Product | Price, availability, ratings | name, image, offers |
| Recipe | Recipe card with image/time | name, image, recipeIngredient |
| Review | Star rating in SERP | reviewRating, author, itemReviewed |
| SoftwareApplication | App info panel | name, offers, operatingSystem |
| VideoObject | Video thumbnail in SERP | name, thumbnailUrl, uploadDate |
| WebSite | Sitelinks search box | name, url, potentialAction |

---

## Required vs Recommended Properties

### Article / BlogPosting

```
Required:
  @type: "Article" or "BlogPosting" or "NewsArticle"
  headline: string (max 110 chars)
  image: URL or ImageObject (at least one)
  datePublished: ISO 8601 date
  author: Person or Organization

Recommended:
  dateModified: ISO 8601 date
  publisher: Organization (with name + logo)
  description: string
  mainEntityOfPage: WebPage
  wordCount: integer
  articleSection: string
```

### Product

```
Required:
  @type: "Product"
  name: string
  image: URL (at least one)

Required for rich results:
  offers: Offer or AggregateOffer
    - price: string number ("29.99")
    - priceCurrency: ISO 4217 ("USD")
    - availability: Schema.org URL
    - url: product page URL

Recommended:
  description: string
  sku: string
  brand: Brand
  aggregateRating: AggregateRating
  review: Review[]
  gtin / gtin13 / mpn: product identifiers
```

### FAQ Page

```
Required:
  @type: "FAQPage"
  mainEntity: Question[]
    - name: the question text
    - acceptedAnswer: Answer
      - text: the answer text (can include HTML links)

Rules:
  - Questions must be visible on the page
  - Answers must fully answer the question
  - FAQ schema is for pages whose PRIMARY purpose is FAQ
  - Don't use on pages where FAQ is a minor section
  - HTML in answer text: <a>, <b>, <strong>, <i>, <em>, <br>, <ol>, <ul>, <li>, <p>, <h2>-<h6>
```

### LocalBusiness

```
Required:
  @type: "LocalBusiness" (or subtype like "Restaurant", "Dentist")
  name: string
  address: PostalAddress

Recommended:
  telephone: string
  url: string
  image: URL
  openingHoursSpecification: OpeningHoursSpecification[]
  geo: GeoCoordinates (latitude + longitude)
  priceRange: string ("$", "$$", "$$$")
  aggregateRating: AggregateRating
  review: Review[]
  sameAs: [social media URLs]

LocalBusiness subtypes:
  Restaurant, Dentist, AutoRepair, BarberShop, BeautySalon,
  Florist, GasStation, GolfCourse, HealthClub, HotelRoom,
  LegalService, Library, MedicalClinic, Pharmacy, RealEstateAgent,
  SportsActivityLocation, Store, TouristInformationCenter, etc.
```

### Event

```
Required:
  @type: "Event"
  name: string
  startDate: ISO 8601 datetime
  location: Place or VirtualLocation

Recommended:
  endDate: ISO 8601 datetime
  description: string
  image: URL
  offers: Offer (ticket info)
  organizer: Organization or Person
  performer: Organization or Person
  eventStatus: EventScheduled / EventCancelled / EventPostponed / EventRescheduled
  eventAttendanceMode: OfflineEventAttendanceMode / OnlineEventAttendanceMode / MixedEventAttendanceMode
```

### HowTo

```
Required:
  @type: "HowTo"
  name: string
  step: HowToStep[]
    - text: instruction text

Recommended:
  description: string
  totalTime: ISO 8601 duration ("PT30M")
  estimatedCost: MonetaryAmount
  supply: HowToSupply[]
  tool: HowToTool[]
  image: URL per step
  video: VideoObject

Duration format (ISO 8601):
  PT5M = 5 minutes
  PT1H30M = 1 hour 30 minutes
  P2D = 2 days
  PT45S = 45 seconds
```

### VideoObject

```
Required:
  @type: "VideoObject"
  name: string
  thumbnailUrl: URL
  uploadDate: ISO 8601 date

Recommended:
  description: string
  duration: ISO 8601 duration
  contentUrl: URL (direct video file)
  embedUrl: URL (embed player URL)
  interactionStatistic: InteractionCounter
  expires: ISO 8601 date (if video expires)

For video rich results:
  Either contentUrl or embedUrl is required
  Google prefers contentUrl when available
```

---

## Availability Values (for Product offers)

```
https://schema.org/InStock
https://schema.org/OutOfStock
https://schema.org/PreOrder
https://schema.org/SoldOut
https://schema.org/OnlineOnly
https://schema.org/InStoreOnly
https://schema.org/LimitedAvailability
https://schema.org/Discontinued
https://schema.org/BackOrder
```

---

## Common Patterns

### Multiple Schemas on One Page

```html
<!-- Page with Article + BreadcrumbList + Organization -->
<script type="application/ld+json">
{
  "@context": "https://schema.org",
  "@type": "BlogPosting",
  "headline": "Article Title",
  "datePublished": "2024-01-15",
  "author": {"@type": "Person", "name": "Author"}
}
</script>

<script type="application/ld+json">
{
  "@context": "https://schema.org",
  "@type": "BreadcrumbList",
  "itemListElement": [
    {"@type": "ListItem", "position": 1, "name": "Home", "item": "https://example.com"},
    {"@type": "ListItem", "position": 2, "name": "Blog", "item": "https://example.com/blog"},
    {"@type": "ListItem", "position": 3, "name": "Article Title"}
  ]
}
</script>
```

### Nested Schemas with @id

```json
{
  "@context": "https://schema.org",
  "@graph": [
    {
      "@type": "Organization",
      "@id": "https://example.com/#organization",
      "name": "Company Name",
      "url": "https://example.com",
      "logo": "https://example.com/logo.png"
    },
    {
      "@type": "WebSite",
      "@id": "https://example.com/#website",
      "name": "Site Name",
      "url": "https://example.com",
      "publisher": {"@id": "https://example.com/#organization"}
    },
    {
      "@type": "WebPage",
      "@id": "https://example.com/about/#webpage",
      "url": "https://example.com/about",
      "name": "About Us",
      "isPartOf": {"@id": "https://example.com/#website"}
    }
  ]
}
```

### AggregateOffer (Price Range)

```json
{
  "@type": "Product",
  "name": "Product with Variants",
  "offers": {
    "@type": "AggregateOffer",
    "lowPrice": "19.99",
    "highPrice": "99.99",
    "priceCurrency": "USD",
    "offerCount": 5,
    "availability": "https://schema.org/InStock"
  }
}
```

### Review with Pros/Cons

```json
{
  "@type": "Review",
  "reviewRating": {
    "@type": "Rating",
    "ratingValue": "4",
    "bestRating": "5"
  },
  "author": {"@type": "Person", "name": "Reviewer"},
  "reviewBody": "Full review text...",
  "positiveNotes": {
    "@type": "ItemList",
    "itemListElement": [
      {"@type": "ListItem", "position": 1, "name": "Fast performance"},
      {"@type": "ListItem", "position": 2, "name": "Easy to use"}
    ]
  },
  "negativeNotes": {
    "@type": "ItemList",
    "itemListElement": [
      {"@type": "ListItem", "position": 1, "name": "Limited free plan"}
    ]
  }
}
```

---

## Testing & Validation

### Google Rich Results Test

```
URL: https://search.google.com/test/rich-results

How to use:
  1. Paste URL or code snippet
  2. Click "Test URL" or "Test Code"
  3. Review detected schemas
  4. Fix any errors or warnings
  5. Preview how rich result will appear

Common errors:
  - "Missing field 'image'" → Add image property
  - "Invalid date" → Use ISO 8601 format
  - "Value does not match" → Check enum values (availability, etc.)
  - "'price' is not a valid number" → Use string: "29.99" not number: 29.99
```

### Schema Markup Validator

```
URL: https://validator.schema.org/

More permissive than Google — validates against full Schema.org vocabulary.
Good for finding structural issues Google's tool might not catch.
```

### Google Search Console

```
Admin > Enhancements

Shows:
  - Which rich result types are detected on your site
  - Valid items count
  - Items with errors
  - Items with warnings
  - Specific URLs and error details

Check weekly for new errors.
```

---

## Quick Decision Guide

```
What type of page?          → Schema to use
──────────────────────────────────────────────
Homepage                    → Organization + WebSite
About page                  → Organization + BreadcrumbList
Blog index                  → BreadcrumbList
Blog post                   → Article/BlogPosting + BreadcrumbList
Product page                → Product + BreadcrumbList
Category page               → BreadcrumbList + CollectionPage
FAQ page                    → FAQPage + BreadcrumbList
Contact page                → LocalBusiness + BreadcrumbList
Pricing page                → Product/SoftwareApplication
Tutorial/Guide              → HowTo + BreadcrumbList
Event page                  → Event + BreadcrumbList
Recipe page                 → Recipe + BreadcrumbList
Video page                  → VideoObject + BreadcrumbList
Job listing                 → JobPosting + BreadcrumbList
Course/Learning             → Course + BreadcrumbList
SaaS landing page           → SoftwareApplication + FAQPage
```

---

## ISO 8601 Date Formats

```
Date only:          2024-01-15
Date + time:        2024-01-15T08:00:00
Date + time + TZ:   2024-01-15T08:00:00-08:00
Date + time UTC:    2024-01-15T16:00:00Z

Duration:
  PT5M              = 5 minutes
  PT1H              = 1 hour
  PT1H30M           = 1 hour 30 minutes
  PT2H15M30S        = 2 hours 15 minutes 30 seconds
  P1D               = 1 day
  P7D               = 7 days
  P1Y               = 1 year
```

---

## Currency Codes (ISO 4217)

```
USD  US Dollar
EUR  Euro
GBP  British Pound
JPY  Japanese Yen
CAD  Canadian Dollar
AUD  Australian Dollar
CHF  Swiss Franc
CNY  Chinese Yuan
INR  Indian Rupee
BRL  Brazilian Real
```
