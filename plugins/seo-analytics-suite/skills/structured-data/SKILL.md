---
name: structured-data
description: >
  Implement Schema.org structured data (JSON-LD) for web pages. Covers all major schema types:
  Article, Product, Organization, FAQ, HowTo, BreadcrumbList, LocalBusiness, Event, Recipe,
  SoftwareApplication, and more. Validates against Google Rich Results requirements.
  Use when adding structured data to a website, fixing rich snippet issues, or optimizing
  for Google's rich results (star ratings, FAQ dropdowns, breadcrumbs, etc.).
version: 1.0.0
argument-hint: "[schema-type-or-page]"
allowed-tools: Read, Grep, Glob, Bash, Write, Edit
model: sonnet
---

# Structured Data Implementation (JSON-LD)

Implement Schema.org structured data using JSON-LD format for Google Rich Results. Structured data helps search engines understand your content and can trigger enhanced search results (rich snippets, knowledge panels, FAQ dropdowns, star ratings, etc.).

## JSON-LD Basics

### Format

```html
<!-- Always use JSON-LD (Google's preferred format) -->
<script type="application/ld+json">
{
  "@context": "https://schema.org",
  "@type": "WebPage",
  "name": "Page Title",
  "description": "Page description"
}
</script>
```

### Rules

```
1. Place JSON-LD in <head> or <body> (both work, <head> is cleaner)
2. Multiple JSON-LD blocks per page are fine
3. Content in JSON-LD MUST match visible page content
4. Use absolute URLs for all url/image properties
5. Required properties vary by type — check Google's docs
6. Test with Google Rich Results Test: https://search.google.com/test/rich-results
7. Monitor in Google Search Console > Enhancements
```

## Schema Types — Complete Reference

### 1. Organization

```json
{
  "@context": "https://schema.org",
  "@type": "Organization",
  "name": "Company Name",
  "url": "https://example.com",
  "logo": "https://example.com/logo.png",
  "description": "Brief company description",
  "foundingDate": "2020-01-01",
  "founders": [
    {
      "@type": "Person",
      "name": "Founder Name"
    }
  ],
  "address": {
    "@type": "PostalAddress",
    "streetAddress": "123 Main St",
    "addressLocality": "San Francisco",
    "addressRegion": "CA",
    "postalCode": "94105",
    "addressCountry": "US"
  },
  "contactPoint": {
    "@type": "ContactPoint",
    "telephone": "+1-555-555-5555",
    "contactType": "customer support",
    "email": "support@example.com",
    "availableLanguage": "English"
  },
  "sameAs": [
    "https://twitter.com/company",
    "https://linkedin.com/company/company",
    "https://github.com/company"
  ]
}
```

### 2. WebSite (with Search Action)

```json
{
  "@context": "https://schema.org",
  "@type": "WebSite",
  "name": "Site Name",
  "url": "https://example.com",
  "potentialAction": {
    "@type": "SearchAction",
    "target": {
      "@type": "EntryPoint",
      "urlTemplate": "https://example.com/search?q={search_term_string}"
    },
    "query-input": "required name=search_term_string"
  }
}
```

### 3. BreadcrumbList

```json
{
  "@context": "https://schema.org",
  "@type": "BreadcrumbList",
  "itemListElement": [
    {
      "@type": "ListItem",
      "position": 1,
      "name": "Home",
      "item": "https://example.com"
    },
    {
      "@type": "ListItem",
      "position": 2,
      "name": "Blog",
      "item": "https://example.com/blog"
    },
    {
      "@type": "ListItem",
      "position": 3,
      "name": "How to Set Up Claude Code Hooks"
    }
  ]
}
```

### 4. Article / BlogPosting

```json
{
  "@context": "https://schema.org",
  "@type": "BlogPosting",
  "headline": "How to Set Up Claude Code Hooks",
  "description": "A step-by-step guide to configuring lifecycle hooks in Claude Code.",
  "image": "https://example.com/images/hooks-guide.png",
  "datePublished": "2024-01-15T08:00:00+00:00",
  "dateModified": "2024-02-01T10:30:00+00:00",
  "author": {
    "@type": "Person",
    "name": "Author Name",
    "url": "https://example.com/authors/author-name"
  },
  "publisher": {
    "@type": "Organization",
    "name": "Site Name",
    "logo": {
      "@type": "ImageObject",
      "url": "https://example.com/logo.png"
    }
  },
  "mainEntityOfPage": {
    "@type": "WebPage",
    "@id": "https://example.com/blog/claude-code-hooks"
  },
  "wordCount": 2500,
  "keywords": ["Claude Code", "hooks", "automation", "development"],
  "articleSection": "Tutorials"
}
```

### 5. Product

```json
{
  "@context": "https://schema.org",
  "@type": "Product",
  "name": "Product Name",
  "image": [
    "https://example.com/photos/product-1.jpg",
    "https://example.com/photos/product-2.jpg"
  ],
  "description": "Product description for rich results.",
  "sku": "SKU-12345",
  "brand": {
    "@type": "Brand",
    "name": "Brand Name"
  },
  "offers": {
    "@type": "Offer",
    "url": "https://example.com/product",
    "priceCurrency": "USD",
    "price": "29.99",
    "priceValidUntil": "2025-12-31",
    "availability": "https://schema.org/InStock",
    "seller": {
      "@type": "Organization",
      "name": "Store Name"
    }
  },
  "aggregateRating": {
    "@type": "AggregateRating",
    "ratingValue": "4.8",
    "reviewCount": "256",
    "bestRating": "5",
    "worstRating": "1"
  },
  "review": [
    {
      "@type": "Review",
      "reviewRating": {
        "@type": "Rating",
        "ratingValue": "5",
        "bestRating": "5"
      },
      "author": {
        "@type": "Person",
        "name": "Reviewer Name"
      },
      "reviewBody": "This product is excellent...",
      "datePublished": "2024-01-10"
    }
  ]
}
```

### 6. FAQ Page

```json
{
  "@context": "https://schema.org",
  "@type": "FAQPage",
  "mainEntity": [
    {
      "@type": "Question",
      "name": "What is Claude Code?",
      "acceptedAnswer": {
        "@type": "Answer",
        "text": "Claude Code is Anthropic's official CLI for Claude, providing an agentic coding experience in your terminal."
      }
    },
    {
      "@type": "Question",
      "name": "How much does it cost?",
      "acceptedAnswer": {
        "@type": "Answer",
        "text": "Claude Code is available through Claude Pro ($20/month), Team ($25/month per user), and Enterprise plans."
      }
    },
    {
      "@type": "Question",
      "name": "Can I use my own API key?",
      "acceptedAnswer": {
        "@type": "Answer",
        "text": "Yes, you can use Claude Code with your own Anthropic API key by setting the ANTHROPIC_API_KEY environment variable."
      }
    }
  ]
}
```

### 7. HowTo

```json
{
  "@context": "https://schema.org",
  "@type": "HowTo",
  "name": "How to Install Claude Code",
  "description": "Step-by-step guide to installing and setting up Claude Code on your machine.",
  "totalTime": "PT5M",
  "estimatedCost": {
    "@type": "MonetaryAmount",
    "currency": "USD",
    "value": "0"
  },
  "tool": [
    {
      "@type": "HowToTool",
      "name": "Terminal / Command Line"
    },
    {
      "@type": "HowToTool",
      "name": "Node.js 18+"
    }
  ],
  "step": [
    {
      "@type": "HowToStep",
      "name": "Install via npm",
      "text": "Run npm install -g @anthropic-ai/claude-code in your terminal.",
      "url": "https://example.com/docs/install#step-1",
      "image": "https://example.com/images/install-step1.png"
    },
    {
      "@type": "HowToStep",
      "name": "Authenticate",
      "text": "Run claude and follow the authentication prompts to connect your account.",
      "url": "https://example.com/docs/install#step-2"
    },
    {
      "@type": "HowToStep",
      "name": "Start using",
      "text": "Navigate to your project directory and run claude to start an interactive session.",
      "url": "https://example.com/docs/install#step-3"
    }
  ]
}
```

### 8. SoftwareApplication

```json
{
  "@context": "https://schema.org",
  "@type": "SoftwareApplication",
  "name": "App Name",
  "operatingSystem": "macOS, Windows, Linux",
  "applicationCategory": "DeveloperApplication",
  "description": "Description of the software.",
  "url": "https://example.com",
  "downloadUrl": "https://example.com/download",
  "screenshot": "https://example.com/screenshot.png",
  "softwareVersion": "2.1.0",
  "datePublished": "2024-01-01",
  "offers": {
    "@type": "Offer",
    "price": "0",
    "priceCurrency": "USD"
  },
  "aggregateRating": {
    "@type": "AggregateRating",
    "ratingValue": "4.7",
    "ratingCount": "1500"
  },
  "author": {
    "@type": "Organization",
    "name": "Company Name",
    "url": "https://example.com"
  }
}
```

### 9. LocalBusiness

```json
{
  "@context": "https://schema.org",
  "@type": "LocalBusiness",
  "name": "Business Name",
  "image": "https://example.com/photos/storefront.jpg",
  "address": {
    "@type": "PostalAddress",
    "streetAddress": "123 Main St",
    "addressLocality": "San Francisco",
    "addressRegion": "CA",
    "postalCode": "94105",
    "addressCountry": "US"
  },
  "geo": {
    "@type": "GeoCoordinates",
    "latitude": 37.7749,
    "longitude": -122.4194
  },
  "url": "https://example.com",
  "telephone": "+1-555-555-5555",
  "email": "contact@example.com",
  "priceRange": "$$",
  "openingHoursSpecification": [
    {
      "@type": "OpeningHoursSpecification",
      "dayOfWeek": ["Monday", "Tuesday", "Wednesday", "Thursday", "Friday"],
      "opens": "09:00",
      "closes": "17:00"
    },
    {
      "@type": "OpeningHoursSpecification",
      "dayOfWeek": ["Saturday"],
      "opens": "10:00",
      "closes": "14:00"
    }
  ],
  "aggregateRating": {
    "@type": "AggregateRating",
    "ratingValue": "4.6",
    "reviewCount": "89"
  }
}
```

### 10. Event

```json
{
  "@context": "https://schema.org",
  "@type": "Event",
  "name": "AI Developer Conference 2024",
  "description": "Annual conference for AI developers and engineers.",
  "startDate": "2024-06-15T09:00:00-07:00",
  "endDate": "2024-06-17T17:00:00-07:00",
  "eventAttendanceMode": "https://schema.org/MixedEventAttendanceMode",
  "eventStatus": "https://schema.org/EventScheduled",
  "location": [
    {
      "@type": "Place",
      "name": "Convention Center",
      "address": {
        "@type": "PostalAddress",
        "addressLocality": "San Francisco",
        "addressRegion": "CA",
        "addressCountry": "US"
      }
    },
    {
      "@type": "VirtualLocation",
      "url": "https://example.com/livestream"
    }
  ],
  "image": "https://example.com/event-poster.jpg",
  "organizer": {
    "@type": "Organization",
    "name": "Organizer Name",
    "url": "https://example.com"
  },
  "performer": {
    "@type": "Person",
    "name": "Keynote Speaker"
  },
  "offers": {
    "@type": "Offer",
    "url": "https://example.com/tickets",
    "price": "299",
    "priceCurrency": "USD",
    "availability": "https://schema.org/InStock",
    "validFrom": "2024-01-01T00:00:00-08:00"
  }
}
```

### 11. Course

```json
{
  "@context": "https://schema.org",
  "@type": "Course",
  "name": "Advanced Claude Code Workflows",
  "description": "Master hooks, agents, and multi-agent workflows in Claude Code.",
  "provider": {
    "@type": "Organization",
    "name": "Provider Name",
    "sameAs": "https://example.com"
  },
  "url": "https://example.com/courses/claude-code-advanced",
  "courseCode": "CC-ADV-101",
  "numberOfCredits": "0",
  "educationalLevel": "Intermediate",
  "inLanguage": "en",
  "hasCourseInstance": {
    "@type": "CourseInstance",
    "courseMode": "online",
    "courseWorkload": "PT10H",
    "instructor": {
      "@type": "Person",
      "name": "Instructor Name"
    }
  },
  "offers": {
    "@type": "Offer",
    "price": "49",
    "priceCurrency": "USD",
    "category": "Self-paced"
  }
}
```

### 12. VideoObject

```json
{
  "@context": "https://schema.org",
  "@type": "VideoObject",
  "name": "Claude Code Tutorial — Getting Started",
  "description": "Learn how to install and use Claude Code in 10 minutes.",
  "thumbnailUrl": "https://example.com/thumbnails/tutorial.jpg",
  "uploadDate": "2024-01-15T08:00:00+00:00",
  "duration": "PT10M30S",
  "contentUrl": "https://example.com/videos/tutorial.mp4",
  "embedUrl": "https://www.youtube.com/embed/VIDEO_ID",
  "interactionStatistic": {
    "@type": "InteractionCounter",
    "interactionType": "https://schema.org/WatchAction",
    "userInteractionCount": 15000
  },
  "author": {
    "@type": "Person",
    "name": "Creator Name"
  }
}
```

## Framework Implementations

### Next.js (App Router)

```typescript
// components/JsonLd.tsx
interface JsonLdProps {
  data: Record<string, any>;
}

export function JsonLd({ data }: JsonLdProps) {
  return (
    <script
      type="application/ld+json"
      dangerouslySetInnerHTML={{ __html: JSON.stringify(data) }}
    />
  );
}

// app/layout.tsx — Organization + WebSite schema
import { JsonLd } from '@/components/JsonLd';

export default function RootLayout({ children }) {
  return (
    <html lang="en">
      <body>
        <JsonLd data={{
          '@context': 'https://schema.org',
          '@type': 'Organization',
          name: 'Company Name',
          url: 'https://example.com',
          logo: 'https://example.com/logo.png',
          sameAs: ['https://twitter.com/company', 'https://github.com/company'],
        }} />
        <JsonLd data={{
          '@context': 'https://schema.org',
          '@type': 'WebSite',
          name: 'Site Name',
          url: 'https://example.com',
        }} />
        {children}
      </body>
    </html>
  );
}

// app/blog/[slug]/page.tsx — Article schema
export default async function BlogPost({ params }) {
  const post = await getPost(params.slug);

  return (
    <>
      <JsonLd data={{
        '@context': 'https://schema.org',
        '@type': 'BlogPosting',
        headline: post.title,
        description: post.excerpt,
        image: post.coverImage,
        datePublished: post.createdAt,
        dateModified: post.updatedAt,
        author: { '@type': 'Person', name: post.author.name },
        publisher: {
          '@type': 'Organization',
          name: 'Site Name',
          logo: { '@type': 'ImageObject', url: 'https://example.com/logo.png' },
        },
      }} />
      <JsonLd data={{
        '@context': 'https://schema.org',
        '@type': 'BreadcrumbList',
        itemListElement: [
          { '@type': 'ListItem', position: 1, name: 'Home', item: 'https://example.com' },
          { '@type': 'ListItem', position: 2, name: 'Blog', item: 'https://example.com/blog' },
          { '@type': 'ListItem', position: 3, name: post.title },
        ],
      }} />
      <article>{/* Post content */}</article>
    </>
  );
}
```

### React (with Helmet)

```tsx
import { Helmet } from 'react-helmet-async';

function ProductPage({ product }) {
  const schema = {
    '@context': 'https://schema.org',
    '@type': 'Product',
    name: product.name,
    image: product.images,
    description: product.description,
    sku: product.sku,
    brand: { '@type': 'Brand', name: product.brand },
    offers: {
      '@type': 'Offer',
      url: window.location.href,
      priceCurrency: 'USD',
      price: product.price,
      availability: product.inStock
        ? 'https://schema.org/InStock'
        : 'https://schema.org/OutOfStock',
    },
  };

  return (
    <>
      <Helmet>
        <script type="application/ld+json">{JSON.stringify(schema)}</script>
      </Helmet>
      <div>{/* Product page content */}</div>
    </>
  );
}
```

### Express.js Middleware

```javascript
// middleware/structuredData.js
function injectJsonLd(schemas) {
  return (req, res, next) => {
    // Store schemas for the template to render
    res.locals.jsonLd = schemas.map(s => JSON.stringify(s));
    next();
  };
}

// Route-level usage
app.get('/', injectJsonLd([
  {
    '@context': 'https://schema.org',
    '@type': 'WebSite',
    name: 'Site Name',
    url: 'https://example.com',
  }
]), (req, res) => {
  res.render('home');
});

// In your EJS/Pug/Handlebars template
// <% if (jsonLd) { jsonLd.forEach(schema => { %>
//   <script type="application/ld+json"><%- schema %></script>
// <% }) } %>
```

## Validation

### Google Rich Results Test

```
URL: https://search.google.com/test/rich-results

What it checks:
  ✅ Valid JSON-LD syntax
  ✅ Required properties present
  ✅ Correct property types
  ✅ Eligible for rich results in Google Search

What it doesn't check:
  ❌ Whether content matches page content
  ❌ Whether Google will actually show rich results
  ❌ Other search engines' requirements
```

### Schema.org Validator

```
URL: https://validator.schema.org/

More permissive than Google — validates against the full Schema.org vocabulary,
not just Google's subset.
```

### Common Validation Errors

```
1. "Missing field 'image'"
   → Add an image property. Google requires images for most rich result types.

2. "Invalid date format"
   → Use ISO 8601: "2024-01-15T08:00:00+00:00" or "2024-01-15"

3. "URL is not valid"
   → Must be absolute URL with https://

4. "Value must be one of..."
   → Check Google's docs for allowed enum values (availability, eventStatus, etc.)

5. "Missing field 'author'"
   → Articles require author. Use Person or Organization type.

6. "'price' is not a valid number"
   → Price must be a string number: "29.99", not 29.99 as JSON number
```

## Checklist

- [ ] Organization schema on homepage
- [ ] BreadcrumbList schema on all pages with breadcrumbs
- [ ] Article/BlogPosting schema on blog posts
- [ ] Product schema on product pages (with offers, ratings if available)
- [ ] FAQ schema on FAQ pages
- [ ] SoftwareApplication schema on app landing pages
- [ ] All schemas validate with Google Rich Results Test
- [ ] All URLs are absolute (https://...)
- [ ] JSON-LD content matches visible page content
- [ ] No duplicate or conflicting schemas
- [ ] Monitor Google Search Console > Enhancements for errors
