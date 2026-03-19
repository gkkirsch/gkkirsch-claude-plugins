---
name: technical-seo
description: >
  Implement technical SEO optimizations for web applications. Covers meta tags, Open Graph,
  Twitter Cards, canonical URLs, XML sitemaps, robots.txt, Core Web Vitals, structured data
  basics, and framework-specific SEO patterns (Next.js, React, Express).
  Use when setting up SEO for a new project, fixing technical SEO issues, or optimizing
  for search engine crawlability and indexing.
version: 1.0.0
argument-hint: "[url-or-component]"
allowed-tools: Read, Grep, Glob, Bash, Write, Edit
model: sonnet
---

# Technical SEO Implementation

Implement technical SEO best practices for web applications. This skill covers the foundational elements that search engines need to crawl, index, and rank your pages.

## Quick Start Checklist

Before diving in, run through these essentials:

```
□ Every page has a unique <title> (50-60 chars)
□ Every page has a unique <meta name="description"> (150-160 chars)
□ <html lang="en"> is set
□ Viewport meta tag present
□ Favicon configured
□ robots.txt exists and is correct
□ XML sitemap exists and is submitted
□ Canonical URLs set on all pages
□ Open Graph tags on all shareable pages
□ HTTPS enforced (no mixed content)
□ Mobile-responsive design
□ Core Web Vitals passing (LCP < 2.5s, INP < 200ms, CLS < 0.1)
```

## 1. HTML Head — The Foundation

### Complete SEO Head Template

```html
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />

  <!-- Primary Meta Tags -->
  <title>Page Title — Brand Name</title>
  <meta name="description" content="Concise description of page content. 150-160 characters. Include primary keyword naturally." />

  <!-- Canonical URL (prevents duplicate content) -->
  <link rel="canonical" href="https://example.com/page" />

  <!-- Favicon -->
  <link rel="icon" href="/favicon.ico" sizes="32x32" />
  <link rel="icon" href="/favicon.svg" type="image/svg+xml" />
  <link rel="apple-touch-icon" href="/apple-touch-icon.png" />
  <link rel="manifest" href="/site.webmanifest" />

  <!-- Open Graph (Facebook, LinkedIn, Discord) -->
  <meta property="og:type" content="website" />
  <meta property="og:url" content="https://example.com/page" />
  <meta property="og:title" content="Page Title — Brand Name" />
  <meta property="og:description" content="Description for social sharing. Can be longer than meta description." />
  <meta property="og:image" content="https://example.com/og-image.png" />
  <meta property="og:image:width" content="1200" />
  <meta property="og:image:height" content="630" />
  <meta property="og:site_name" content="Brand Name" />
  <meta property="og:locale" content="en_US" />

  <!-- Twitter Card -->
  <meta name="twitter:card" content="summary_large_image" />
  <meta name="twitter:site" content="@brandhandle" />
  <meta name="twitter:creator" content="@authorhandle" />
  <meta name="twitter:title" content="Page Title" />
  <meta name="twitter:description" content="Description for Twitter sharing." />
  <meta name="twitter:image" content="https://example.com/twitter-image.png" />

  <!-- Robots -->
  <meta name="robots" content="index, follow" />

  <!-- For paginated content -->
  <!-- <link rel="prev" href="https://example.com/page/2" /> -->
  <!-- <link rel="next" href="https://example.com/page/4" /> -->

  <!-- Preconnect to critical origins -->
  <link rel="preconnect" href="https://fonts.googleapis.com" />
  <link rel="preconnect" href="https://cdn.example.com" />

  <!-- DNS prefetch for third parties -->
  <link rel="dns-prefetch" href="https://www.googletagmanager.com" />
</head>
```

### Title Tag Rules

```
Format: Primary Keyword — Secondary Keyword | Brand Name
Length: 50-60 characters (Google truncates at ~60)

Good examples:
  "React Performance Guide — Optimize Rendering | DevBlog"
  "Buy Running Shoes Online — Free Shipping | ShoeStore"
  "Claude Code Hooks Tutorial — Automate Your Workflow"

Bad examples:
  "Home" (too generic)
  "Welcome to Our Amazing Website — The Best Place for Everything You Need" (too long)
  "Page 1" (meaningless)
  "keyword, keyword, keyword, keyword" (keyword stuffing)

Rules:
  ✅ Unique per page
  ✅ Primary keyword near the start
  ✅ Brand name at the end (separated by | or —)
  ✅ Readable by humans (not just keywords)
  ❌ No duplicate titles across pages
  ❌ No ALL CAPS
  ❌ No keyword stuffing
```

### Meta Description Rules

```
Length: 150-160 characters (Google truncates around 155-160)

Good examples:
  "Learn how to set up Claude Code hooks for automated testing, linting,
   and deployment. Step-by-step guide with real examples."
  "Free shipping on all running shoes. Browse 500+ styles from Nike,
   Adidas, and more. 30-day returns. Shop now."

Rules:
  ✅ Unique per page
  ✅ Contains primary keyword naturally
  ✅ Includes a call-to-action when appropriate
  ✅ Accurately describes page content
  ✅ Compelling enough to click (it's your ad copy in search results)
  ❌ No duplicate descriptions across pages
  ❌ Don't just list keywords
  ❌ Don't exceed 160 characters
```

## 2. Robots.txt

### Standard robots.txt

```
# /robots.txt
User-agent: *
Allow: /

# Block admin/private areas
Disallow: /admin/
Disallow: /api/
Disallow: /dashboard/
Disallow: /internal/
Disallow: /_next/      # Next.js internals
Disallow: /cdn-cgi/    # Cloudflare internals

# Block query parameter variations
Disallow: /*?sort=
Disallow: /*?filter=
Disallow: /*?ref=

# Sitemap location
Sitemap: https://example.com/sitemap.xml

# Crawl rate (optional, only if you're getting hammered)
# Crawl-delay: 10
```

### Next.js robots.ts (App Router)

```typescript
// app/robots.ts
import type { MetadataRoute } from 'next';

export default function robots(): MetadataRoute.Robots {
  const baseUrl = process.env.NEXT_PUBLIC_BASE_URL || 'https://example.com';

  return {
    rules: [
      {
        userAgent: '*',
        allow: '/',
        disallow: ['/admin/', '/api/', '/dashboard/', '/_next/'],
      },
    ],
    sitemap: `${baseUrl}/sitemap.xml`,
  };
}
```

## 3. XML Sitemap

### Static Sitemap

```xml
<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url>
    <loc>https://example.com/</loc>
    <lastmod>2024-01-15</lastmod>
    <changefreq>weekly</changefreq>
    <priority>1.0</priority>
  </url>
  <url>
    <loc>https://example.com/about</loc>
    <lastmod>2024-01-10</lastmod>
    <changefreq>monthly</changefreq>
    <priority>0.8</priority>
  </url>
  <url>
    <loc>https://example.com/blog</loc>
    <lastmod>2024-01-14</lastmod>
    <changefreq>daily</changefreq>
    <priority>0.9</priority>
  </url>
</urlset>
```

### Next.js Dynamic Sitemap (App Router)

```typescript
// app/sitemap.ts
import type { MetadataRoute } from 'next';
import { db } from '@/lib/db';

export default async function sitemap(): Promise<MetadataRoute.Sitemap> {
  const baseUrl = process.env.NEXT_PUBLIC_BASE_URL || 'https://example.com';

  // Static pages
  const staticPages: MetadataRoute.Sitemap = [
    { url: baseUrl, lastModified: new Date(), changeFrequency: 'weekly', priority: 1 },
    { url: `${baseUrl}/about`, lastModified: new Date(), changeFrequency: 'monthly', priority: 0.8 },
    { url: `${baseUrl}/pricing`, lastModified: new Date(), changeFrequency: 'monthly', priority: 0.9 },
    { url: `${baseUrl}/blog`, lastModified: new Date(), changeFrequency: 'daily', priority: 0.9 },
  ];

  // Dynamic pages from database
  const posts = await db.post.findMany({
    where: { published: true },
    select: { slug: true, updatedAt: true },
  });

  const blogPages: MetadataRoute.Sitemap = posts.map((post) => ({
    url: `${baseUrl}/blog/${post.slug}`,
    lastModified: post.updatedAt,
    changeFrequency: 'weekly' as const,
    priority: 0.7,
  }));

  return [...staticPages, ...blogPages];
}
```

### Express.js Sitemap Generator

```javascript
const { SitemapStream, streamToPromise } = require('sitemap');
const { Readable } = require('stream');

app.get('/sitemap.xml', async (req, res) => {
  const baseUrl = 'https://example.com';

  // Static pages
  const links = [
    { url: '/', changefreq: 'weekly', priority: 1.0 },
    { url: '/about', changefreq: 'monthly', priority: 0.8 },
    { url: '/pricing', changefreq: 'monthly', priority: 0.9 },
  ];

  // Add dynamic pages
  const posts = await Post.find({ published: true }).select('slug updatedAt');
  for (const post of posts) {
    links.push({
      url: `/blog/${post.slug}`,
      changefreq: 'weekly',
      priority: 0.7,
      lastmod: post.updatedAt.toISOString(),
    });
  }

  const stream = new SitemapStream({ hostname: baseUrl });
  const xmlString = await streamToPromise(Readable.from(links).pipe(stream));

  res.header('Content-Type', 'application/xml');
  res.send(xmlString.toString());
});
```

## 4. Canonical URLs

### When to Use Canonicals

```
ALWAYS set a canonical URL. It tells search engines which version of a page is the "real" one.

Scenarios requiring canonicals:
  1. Same content at multiple URLs:
     /products?sort=price  →  canonical: /products
     /products?page=1      →  canonical: /products
     /products/            →  canonical: /products (no trailing slash)

  2. HTTP vs HTTPS:
     http://example.com/page  →  canonical: https://example.com/page

  3. www vs non-www:
     www.example.com/page  →  canonical: example.com/page

  4. Syndicated content:
     partner-site.com/your-article  →  canonical: your-site.com/your-article

  5. Mobile vs desktop:
     m.example.com/page  →  canonical: example.com/page (with responsive design)

Implementation:
  <link rel="canonical" href="https://example.com/exact-page-url" />

Rules:
  ✅ Use absolute URLs (https://example.com/page, not /page)
  ✅ Self-referencing canonicals are fine (page points to itself)
  ✅ Canonical should be the version you want indexed
  ✅ Only one canonical per page
  ❌ Don't canonical to a 404 or redirect
  ❌ Don't canonical paginated pages to page 1 (use rel=prev/next)
```

## 5. Core Web Vitals

### Largest Contentful Paint (LCP) — Target: < 2.5s

```
What it measures: Time for the largest visible element to render.

Common culprits and fixes:

1. Hero images too large:
   ✅ Use WebP/AVIF format
   ✅ Set explicit width/height
   ✅ Use srcset for responsive images
   ✅ Preload the hero image: <link rel="preload" as="image" href="hero.webp">
   ✅ Use loading="eager" for above-fold images (not lazy)

2. Render-blocking resources:
   ✅ Inline critical CSS
   ✅ Defer non-critical CSS: <link rel="preload" as="style" onload="this.rel='stylesheet'">
   ✅ async/defer on non-critical scripts
   ✅ Preconnect to third-party origins

3. Server response time (TTFB):
   ✅ Use a CDN
   ✅ Enable compression (gzip/brotli)
   ✅ Cache HTML responses
   ✅ Optimize database queries
```

### Interaction to Next Paint (INP) — Target: < 200ms

```
What it measures: Responsiveness to user interactions (clicks, taps, key presses).

Common culprits and fixes:

1. Long JavaScript tasks:
   ✅ Break up tasks with setTimeout or requestIdleCallback
   ✅ Use web workers for heavy computation
   ✅ Debounce/throttle event handlers
   ✅ Use requestAnimationFrame for visual updates

2. Main thread blocking:
   ✅ Code-split with dynamic imports
   ✅ Lazy load below-fold components
   ✅ Minimize third-party scripts
   ✅ Use will-change CSS for animated elements

3. React-specific:
   ✅ Use React.memo for expensive components
   ✅ Use useTransition for non-urgent updates
   ✅ Virtualize long lists (react-window, tanstack-virtual)
```

### Cumulative Layout Shift (CLS) — Target: < 0.1

```
What it measures: Visual stability. How much the page layout shifts unexpectedly.

Common culprits and fixes:

1. Images without dimensions:
   ✅ Always set width and height on <img> tags
   ✅ Use aspect-ratio CSS for responsive images
   ❌ <img src="photo.jpg"> (no dimensions = CLS)
   ✅ <img src="photo.jpg" width="800" height="600" style="max-width:100%;height:auto">

2. Dynamic content injection:
   ✅ Reserve space for ads, embeds, dynamic content
   ✅ Use min-height on containers that load async content
   ✅ Skeleton screens instead of empty space

3. Web fonts causing FOUT/FOIT:
   ✅ font-display: swap (or optional)
   ✅ Preload critical fonts: <link rel="preload" as="font" type="font/woff2" href="font.woff2" crossorigin>
   ✅ Use size-adjust to match fallback font metrics

4. Injected elements:
   ❌ Cookie banners that push content down
   ✅ Use fixed/sticky positioning for banners
   ❌ Dynamically injected headers/toolbars
```

## 6. Image Optimization

### Responsive Images

```html
<!-- Responsive image with multiple sizes -->
<img
  src="photo-800.webp"
  srcset="
    photo-400.webp 400w,
    photo-800.webp 800w,
    photo-1200.webp 1200w,
    photo-1600.webp 1600w
  "
  sizes="(max-width: 640px) 100vw,
         (max-width: 1024px) 50vw,
         800px"
  alt="Descriptive alt text for SEO and accessibility"
  width="800"
  height="600"
  loading="lazy"
  decoding="async"
/>

<!-- Art direction with <picture> -->
<picture>
  <source media="(max-width: 640px)" srcset="photo-mobile.webp" type="image/webp" />
  <source media="(max-width: 640px)" srcset="photo-mobile.jpg" type="image/jpeg" />
  <source srcset="photo-desktop.webp" type="image/webp" />
  <img src="photo-desktop.jpg" alt="Alt text" width="1200" height="600" loading="lazy" />
</picture>
```

### Next.js Image Component

```tsx
import Image from 'next/image';

// Optimized image with automatic WebP/AVIF, lazy loading, blur placeholder
<Image
  src="/photos/hero.jpg"
  alt="Descriptive alt text"
  width={1200}
  height={600}
  priority  // Use for above-fold images (disables lazy loading)
  placeholder="blur"
  blurDataURL="data:image/jpeg;base64,..."  // Or import static image for auto-blur
  sizes="(max-width: 768px) 100vw, (max-width: 1200px) 50vw, 33vw"
/>
```

## 7. Internal Linking

```
Internal linking rules for SEO:

1. Use descriptive anchor text:
   ✅ <a href="/pricing">view our pricing plans</a>
   ❌ <a href="/pricing">click here</a>
   ❌ <a href="/pricing">learn more</a>

2. Link to important pages from high-authority pages:
   - Homepage → top category pages → individual pages
   - Blog posts → relevant product/feature pages
   - About page → key landing pages

3. Breadcrumbs (with structured data):
   Home > Category > Subcategory > Current Page
   Each level is a link with proper anchor text.

4. Related content:
   "Related articles" or "You might also like" sections
   help search engines understand content relationships.

5. Avoid:
   ❌ Orphan pages (pages with no internal links pointing to them)
   ❌ Too many links per page (Google says "reasonable number")
   ❌ JavaScript-only links (use real <a> tags)
   ❌ Links in iframes
```

## 8. URL Structure

```
Good URL patterns:
  ✅ /blog/how-to-set-up-claude-code-hooks
  ✅ /products/running-shoes
  ✅ /docs/api/authentication

Bad URL patterns:
  ❌ /blog/post?id=12345
  ❌ /p/1a2b3c4d
  ❌ /category/subcategory/sub-subcategory/sub-sub-subcategory/page
  ❌ /Blog/How-To-Set-Up-Claude-Code-Hooks (mixed case)

Rules:
  ✅ Lowercase only
  ✅ Hyphens between words (not underscores)
  ✅ Short and descriptive
  ✅ Include primary keyword
  ✅ No trailing slashes (pick one convention and redirect the other)
  ✅ Max 3-4 levels deep
  ❌ No query parameters for indexable content
  ❌ No session IDs in URLs
  ❌ No file extensions (.html, .php) unless necessary
```

## 9. Redirects

```
Redirect types:

301 (Permanent): Use when a page has permanently moved.
  Old URL → New URL (search engines transfer link equity)
  Example: /old-page → /new-page

302 (Temporary): Use when a page is temporarily moved.
  Example: /sale → /summer-sale-2024 (seasonal)

Common redirect scenarios:
  HTTP → HTTPS (301)
  www → non-www or vice versa (301)
  Old slug → new slug (301)
  Trailing slash normalization (301)
  Deleted page → relevant alternative (301)
  Deleted page → no alternative (410 Gone, not 404)

Express.js redirect middleware:
```

```javascript
// Enforce HTTPS
app.use((req, res, next) => {
  if (req.headers['x-forwarded-proto'] !== 'https' && process.env.NODE_ENV === 'production') {
    return res.redirect(301, `https://${req.headers.host}${req.url}`);
  }
  next();
});

// Trailing slash normalization (remove trailing slash)
app.use((req, res, next) => {
  if (req.path !== '/' && req.path.endsWith('/')) {
    const query = req.url.slice(req.path.length);
    return res.redirect(301, req.path.slice(0, -1) + query);
  }
  next();
});

// Redirect map for old URLs
const redirects = {
  '/old-page': '/new-page',
  '/blog/old-post': '/blog/new-post',
  '/products/discontinued': '/products',
};

app.use((req, res, next) => {
  const target = redirects[req.path];
  if (target) return res.redirect(301, target);
  next();
});
```

## 10. Next.js SEO — Complete Setup

### App Router Metadata API

```typescript
// app/layout.tsx — Global metadata
import type { Metadata } from 'next';

export const metadata: Metadata = {
  metadataBase: new URL('https://example.com'),
  title: {
    template: '%s | Brand Name',
    default: 'Brand Name — Tagline',
  },
  description: 'Default site description for SEO.',
  openGraph: {
    type: 'website',
    locale: 'en_US',
    siteName: 'Brand Name',
    images: [{ url: '/og-default.png', width: 1200, height: 630 }],
  },
  twitter: {
    card: 'summary_large_image',
    site: '@brandhandle',
  },
  robots: {
    index: true,
    follow: true,
    googleBot: {
      index: true,
      follow: true,
      'max-video-preview': -1,
      'max-image-preview': 'large',
      'max-snippet': -1,
    },
  },
  verification: {
    google: 'google-site-verification-code',
  },
};

// app/blog/[slug]/page.tsx — Dynamic metadata
export async function generateMetadata({ params }): Promise<Metadata> {
  const post = await getPost(params.slug);

  return {
    title: post.title,
    description: post.excerpt,
    openGraph: {
      title: post.title,
      description: post.excerpt,
      type: 'article',
      publishedTime: post.createdAt,
      modifiedTime: post.updatedAt,
      authors: [post.author.name],
      images: [{ url: post.coverImage, width: 1200, height: 630 }],
    },
    twitter: {
      title: post.title,
      description: post.excerpt,
      images: [post.coverImage],
    },
    alternates: {
      canonical: `/blog/${params.slug}`,
    },
  };
}
```

## Checklist

When done implementing, verify:

- [ ] Every page has unique title (50-60 chars) and description (150-160 chars)
- [ ] `<html lang>` is set
- [ ] Canonical URLs on all pages
- [ ] robots.txt is correct and accessible
- [ ] XML sitemap exists, is valid, and submitted to Google Search Console
- [ ] Open Graph and Twitter Card tags on all shareable pages
- [ ] All images have alt text, width, height, and use modern formats
- [ ] Internal links use descriptive anchor text
- [ ] URLs are clean, lowercase, hyphenated
- [ ] HTTPS enforced with proper redirects
- [ ] Core Web Vitals passing (check with Lighthouse or PageSpeed Insights)
- [ ] No orphan pages (every page reachable from navigation or links)
- [ ] Structured data validates (use Google Rich Results Test)
