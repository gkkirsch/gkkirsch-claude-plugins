# SEO Checklist — Complete Reference

## Pre-Launch SEO Checklist

### Technical Foundation

```
□ HTTPS enforced (no mixed content)
□ www vs non-www redirected (pick one, 301 the other)
□ Trailing slash normalized (pick one convention)
□ robots.txt present and correct
□ XML sitemap generated and submitted to Google Search Console
□ Google Search Console verified
□ Bing Webmaster Tools verified (optional but recommended)
□ 404 page exists and is helpful (includes navigation, search)
□ Server returns proper HTTP status codes (200, 301, 404, 410, 500)
□ No redirect chains (A→B→C should be A→C)
□ No redirect loops
□ Page load time < 3 seconds
□ Mobile-responsive design
□ Core Web Vitals passing (LCP < 2.5s, INP < 200ms, CLS < 0.1)
```

### On-Page SEO (Every Page)

```
□ Unique <title> tag (50-60 characters)
□ Unique <meta description> (150-160 characters)
□ <html lang="xx"> attribute set
□ <meta charset="utf-8">
□ <meta name="viewport" content="width=device-width, initial-scale=1">
□ Canonical URL set (<link rel="canonical">)
□ One H1 per page (contains primary keyword)
□ Heading hierarchy (H1 → H2 → H3, no skipping)
□ Images have descriptive alt text
□ Images have width and height attributes
□ Images use modern formats (WebP/AVIF with fallbacks)
□ Internal links use descriptive anchor text
□ External links use rel="noopener" (security, not SEO)
□ No broken links (internal or external)
```

### Social & Sharing

```
□ Open Graph tags (og:title, og:description, og:image, og:url, og:type)
□ OG image is 1200x630px
□ Twitter Card tags (twitter:card, twitter:title, twitter:description, twitter:image)
□ Twitter Card type is "summary_large_image"
□ Favicon configured (favicon.ico + favicon.svg + apple-touch-icon)
□ Site manifest (site.webmanifest)
```

### Structured Data

```
□ Organization schema on homepage
□ WebSite schema with SearchAction (if site has search)
□ BreadcrumbList schema on pages with breadcrumbs
□ Article/BlogPosting schema on blog posts
□ Product schema on product pages (with offers)
□ FAQ schema on FAQ pages
□ LocalBusiness schema (if applicable)
□ All schemas validate with Google Rich Results Test
```

### Content Quality

```
□ Every page has at least 300 words of unique content
□ No duplicate content across pages
□ No thin content pages (minimal value, mostly boilerplate)
□ No keyword stuffing
□ Content matches search intent for target keywords
□ Spelling and grammar checked
□ Content is up-to-date and accurate
□ Author information present on articles (E-E-A-T)
□ Sources cited where appropriate
```

### Navigation & Architecture

```
□ Flat site architecture (important pages within 3 clicks of homepage)
□ Breadcrumb navigation on all pages (except homepage)
□ Main navigation includes key pages
□ Footer navigation includes secondary pages
□ HTML sitemap page exists (for users, not just search engines)
□ No orphan pages (every page linked from at least one other page)
□ Category/tag pages have unique content (not just a list of links)
□ Pagination uses rel="next" and rel="prev" (or load-more patterns)
```

### Performance

```
□ Google PageSpeed Insights score > 90 (mobile)
□ First Contentful Paint (FCP) < 1.8s
□ Largest Contentful Paint (LCP) < 2.5s
□ Interaction to Next Paint (INP) < 200ms
□ Cumulative Layout Shift (CLS) < 0.1
□ Time to First Byte (TTFB) < 800ms
□ Critical CSS inlined
□ JavaScript deferred or async (non-critical)
□ Images lazy-loaded (below the fold)
□ Hero image preloaded
□ Web fonts preloaded with font-display: swap
□ Compression enabled (gzip or Brotli)
□ CDN configured for static assets
□ Browser caching headers set
```

### Security & Trust

```
□ SSL certificate valid and not expiring soon
□ HSTS header enabled
□ Content Security Policy header set
□ No mixed content warnings
□ Privacy policy page exists
□ Terms of service page exists
□ Cookie consent banner (GDPR/CCPA compliance)
□ Contact information accessible
```

### Analytics & Monitoring

```
□ Google Analytics 4 installed and tracking
□ Cookie consent before analytics tracking
□ Google Search Console monitoring
□ Conversion goals configured in GA4
□ 404 error monitoring in place
□ Uptime monitoring configured
□ Core Web Vitals monitoring (CrUX or lab data)
```

---

## Ongoing SEO Tasks

### Weekly

```
□ Check Google Search Console for errors
□ Review 404 errors and fix or redirect
□ Monitor Core Web Vitals
□ Check for new crawl issues
□ Review and respond to reviews (if applicable)
```

### Monthly

```
□ Publish new content (blog posts, guides, case studies)
□ Update old content with new information
□ Build internal links to new content
□ Audit and fix broken links
□ Review keyword rankings for target terms
□ Check competitors' new content
□ Review and optimize underperforming pages
□ Submit new/updated pages to Google for indexing
```

### Quarterly

```
□ Full technical SEO audit
□ Content audit (thin content, duplicates, outdated)
□ Backlink profile review
□ Competitor analysis
□ Keyword research refresh
□ Schema markup audit
□ Site speed optimization review
□ Review analytics goals and conversion tracking
```

---

## SEO Audit Template

Use this template when running a full SEO audit:

```markdown
# SEO Audit Report — [Site Name]
# Date: YYYY-MM-DD

## Executive Summary
- Overall health: [Good / Needs Work / Critical Issues]
- Pages audited: N
- Critical issues: N
- Warnings: N
- Passed checks: N

## Technical SEO
### Crawlability
- [ ] robots.txt: [OK / Issue]
- [ ] XML sitemap: [OK / Issue]
- [ ] Crawl errors: [N errors]
- [ ] Redirect issues: [N chains/loops]

### Indexability
- [ ] Pages indexed vs total: [N / N]
- [ ] Noindex pages (intentional): [list]
- [ ] Duplicate content issues: [list]
- [ ] Canonical URL issues: [list]

### Performance
- [ ] Average page load: [N seconds]
- [ ] LCP: [N seconds]
- [ ] INP: [N ms]
- [ ] CLS: [N]
- [ ] Mobile score: [N/100]

## On-Page SEO
### Meta Tags
- Pages missing title: [N]
- Pages with duplicate titles: [N]
- Pages missing meta description: [N]
- Pages with duplicate descriptions: [N]
- Pages missing H1: [N]
- Pages with multiple H1: [N]

### Content
- Thin content pages (< 300 words): [N]
- Pages without images: [N]
- Images missing alt text: [N]

### Internal Linking
- Orphan pages: [N]
- Pages with 0 internal links: [N]
- Broken internal links: [N]

## Structured Data
- Pages with schema: [N / total]
- Schema validation errors: [N]
- Rich results eligible: [N pages]

## Recommendations (Priority Order)
1. [Critical] ...
2. [High] ...
3. [Medium] ...
4. [Low] ...
```

---

## Common SEO Mistakes

### Title Tags

| Mistake | Fix |
|---------|-----|
| Same title on every page | Unique title per page with template: "Page Name \| Brand" |
| Title too long (> 60 chars) | Keep to 50-60 characters |
| No keyword in title | Include primary keyword near the start |
| Keyword stuffing in title | One primary keyword, natural language |
| Missing brand name | Add brand at end: "Page Title \| Brand" |

### Meta Descriptions

| Mistake | Fix |
|---------|-----|
| No meta description | Write unique 150-160 char description |
| Duplicate across pages | Unique description per page |
| Too short (< 120 chars) | Aim for 150-160 characters |
| No call-to-action | Include actionable language |
| Just a keyword list | Write compelling, descriptive copy |

### Images

| Mistake | Fix |
|---------|-----|
| No alt text | Descriptive alt text for every image |
| alt="image" or alt="photo" | Describe what the image shows |
| No width/height | Set dimensions to prevent CLS |
| Huge file sizes | Compress and use WebP/AVIF |
| No lazy loading | Add loading="lazy" to below-fold images |

### URLs

| Mistake | Fix |
|---------|-----|
| Query parameters for content | Clean URLs: /blog/post-title |
| Mixed case | Lowercase only |
| Underscores | Use hyphens |
| Too many levels deep | Keep to 3-4 levels max |
| Non-descriptive | Include keywords in URL path |

### Internal Linking

| Mistake | Fix |
|---------|-----|
| "Click here" anchor text | Descriptive: "view our pricing plans" |
| Orphan pages | Link from navigation, related content, or sitemap |
| JavaScript-only links | Use real `<a>` tags with href |
| Too many links per page | Keep reasonable (~100 max) |
| No breadcrumbs | Add breadcrumb navigation with schema |

---

## SEO Tools Reference

### Free Tools

| Tool | Purpose |
|------|---------|
| Google Search Console | Index status, crawl errors, search analytics |
| Google PageSpeed Insights | Core Web Vitals, performance audit |
| Google Rich Results Test | Structured data validation |
| Google Mobile-Friendly Test | Mobile responsiveness check |
| Bing Webmaster Tools | Bing-specific indexing and SEO |
| Lighthouse (Chrome DevTools) | Performance, accessibility, SEO audit |
| Schema.org Validator | Full Schema.org validation |
| Screaming Frog (free up to 500 URLs) | Technical SEO crawler |

### Keyword Research

| Tool | Notes |
|------|-------|
| Google Keyword Planner | Free with Google Ads account |
| Google Trends | Trending topics and seasonal patterns |
| Google Search (autocomplete) | Real user search queries |
| AnswerThePublic | Question-based keyword ideas |
| Also Asked | Related questions people ask |

### Paid Tools (If Budget Allows)

| Tool | Best For |
|------|----------|
| Ahrefs | Backlink analysis, keyword research, competitor analysis |
| SEMrush | All-in-one SEO suite |
| Moz | Domain authority, link building |
| Screaming Frog (paid) | Large site technical audits |
| Surfer SEO | Content optimization |
