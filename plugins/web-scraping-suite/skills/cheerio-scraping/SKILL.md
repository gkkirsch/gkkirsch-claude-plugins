---
name: cheerio-scraping
description: >
  Web scraping with Cheerio and fetch — HTML parsing, CSS selectors, data
  extraction, pagination, rate limiting, caching, and structured output.
  Triggers: "cheerio", "scrape html", "parse html", "web scraping",
  "extract data from website", "crawl pages", "scrape static",
  "fetch and parse", "html parser".
  NOT for: JavaScript-rendered pages (use puppeteer-crawling).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Cheerio Web Scraping

## Setup

```bash
npm install cheerio
# No additional deps needed — use built-in fetch (Node 18+)
# Or: npm install node-fetch (for older Node)
```

## Basic Scraping

```typescript
import * as cheerio from "cheerio";

async function scrape(url: string) {
  const response = await fetch(url, {
    headers: {
      "User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36",
      "Accept": "text/html,application/xhtml+xml",
      "Accept-Language": "en-US,en;q=0.9",
    },
  });

  if (!response.ok) {
    throw new Error(`HTTP ${response.status}: ${response.statusText}`);
  }

  const html = await response.text();
  const $ = cheerio.load(html);

  return $;
}
```

## CSS Selectors

```typescript
const $ = cheerio.load(html);

// Basic selectors
$("h1").text();                          // first h1 text
$("h1").first().text();                  // explicit first
$(".product-card").length;               // count elements
$("a").attr("href");                     // first link's href
$("img").attr("src");                    // first image src

// Attribute selectors
$('a[href^="https"]');                   // starts with
$('a[href$=".pdf"]');                    // ends with
$('a[href*="example"]');                 // contains
$('input[type="email"]');                // exact match
$("[data-testid='submit']");             // data attributes

// Positional
$("li:first-child");
$("li:last-child");
$("li:nth-child(2)");
$("tr:even");
$("tr:odd");

// Hierarchy
$("table > tbody > tr");                 // direct children
$(".card .title");                       // descendants
$("h2 + p");                            // adjacent sibling
$("h2 ~ p");                            // all siblings after

// Pseudo-selectors
$("a:contains('Read more')");           // text content
$("div:has(img)");                      // contains child
$("p:not(.hidden)");                    // negation
$("input:empty");                       // no children/text
```

## Data Extraction Patterns

```typescript
// Extract structured data from a list
interface Product {
  name: string;
  price: number;
  url: string;
  image: string;
  rating: number;
}

function extractProducts($: cheerio.CheerioAPI): Product[] {
  const products: Product[] = [];

  $(".product-card").each((i, el) => {
    const $el = $(el);
    products.push({
      name: $el.find(".product-name").text().trim(),
      price: parseFloat($el.find(".price").text().replace(/[^0-9.]/g, "")),
      url: $el.find("a").attr("href") || "",
      image: $el.find("img").attr("src") || "",
      rating: parseFloat($el.find(".rating").attr("data-score") || "0"),
    });
  });

  return products;
}

// Extract table data
function extractTable($: cheerio.CheerioAPI, selector: string) {
  const headers: string[] = [];
  const rows: Record<string, string>[] = [];

  $(`${selector} thead th`).each((i, el) => {
    headers.push($(el).text().trim());
  });

  $(`${selector} tbody tr`).each((i, tr) => {
    const row: Record<string, string> = {};
    $(tr).find("td").each((j, td) => {
      row[headers[j] || `col_${j}`] = $(td).text().trim();
    });
    rows.push(row);
  });

  return rows;
}

// Extract meta tags / SEO data
function extractMeta($: cheerio.CheerioAPI) {
  return {
    title: $("title").text(),
    description: $('meta[name="description"]').attr("content") || "",
    ogTitle: $('meta[property="og:title"]').attr("content") || "",
    ogDescription: $('meta[property="og:description"]').attr("content") || "",
    ogImage: $('meta[property="og:image"]').attr("content") || "",
    canonical: $('link[rel="canonical"]').attr("href") || "",
  };
}

// Extract JSON-LD structured data
function extractJsonLd($: cheerio.CheerioAPI) {
  const scripts: any[] = [];
  $('script[type="application/ld+json"]').each((i, el) => {
    try {
      scripts.push(JSON.parse($(el).html() || "{}"));
    } catch {}
  });
  return scripts;
}
```

## Pagination

```typescript
async function scrapeAllPages(baseUrl: string) {
  let allItems: any[] = [];
  let page = 1;
  let hasMore = true;

  while (hasMore) {
    const url = `${baseUrl}?page=${page}`;
    console.log(`Scraping page ${page}...`);

    const response = await fetch(url, {
      headers: { "User-Agent": "MyScraper/1.0 (contact@example.com)" },
    });
    const html = await response.text();
    const $ = cheerio.load(html);

    const items = extractItems($);
    allItems = allItems.concat(items);

    // Check for next page
    hasMore = $("a.next-page").length > 0 && items.length > 0;
    page++;

    // Rate limit: wait 1-3 seconds between requests
    await sleep(1000 + Math.random() * 2000);
  }

  return allItems;
}

// Follow "next" links instead of page numbers
async function scrapeLinkedPages(startUrl: string) {
  let allItems: any[] = [];
  let currentUrl: string | null = startUrl;

  while (currentUrl) {
    const response = await fetch(currentUrl);
    const html = await response.text();
    const $ = cheerio.load(html);

    allItems = allItems.concat(extractItems($));

    // Get next page URL
    const nextLink = $('a[rel="next"]').attr("href");
    currentUrl = nextLink ? new URL(nextLink, currentUrl).href : null;

    await sleep(1500 + Math.random() * 1500);
  }

  return allItems;
}

function sleep(ms: number) {
  return new Promise(resolve => setTimeout(resolve, ms));
}
```

## Rate Limiting

```typescript
class RateLimiter {
  private queue: (() => Promise<void>)[] = [];
  private running = 0;

  constructor(
    private maxConcurrent: number = 2,
    private delayMs: number = 1000
  ) {}

  async add<T>(fn: () => Promise<T>): Promise<T> {
    while (this.running >= this.maxConcurrent) {
      await new Promise(resolve => setTimeout(resolve, 100));
    }

    this.running++;
    try {
      const result = await fn();
      await new Promise(resolve =>
        setTimeout(resolve, this.delayMs + Math.random() * this.delayMs)
      );
      return result;
    } finally {
      this.running--;
    }
  }
}

// Usage
const limiter = new RateLimiter(2, 1500);
const urls = [/* ... */];

const results = await Promise.all(
  urls.map(url => limiter.add(() => scrapePage(url)))
);
```

## Caching

```typescript
import { readFileSync, writeFileSync, existsSync, mkdirSync } from "fs";
import { createHash } from "crypto";

class ScrapeCache {
  constructor(private cacheDir: string = ".cache") {
    if (!existsSync(cacheDir)) mkdirSync(cacheDir, { recursive: true });
  }

  private keyFor(url: string): string {
    return createHash("md5").update(url).digest("hex");
  }

  get(url: string): string | null {
    const path = `${this.cacheDir}/${this.keyFor(url)}.html`;
    if (!existsSync(path)) return null;

    const stats = require("fs").statSync(path);
    const ageHours = (Date.now() - stats.mtimeMs) / 3600000;
    if (ageHours > 24) return null; // expire after 24h

    return readFileSync(path, "utf-8");
  }

  set(url: string, html: string): void {
    const path = `${this.cacheDir}/${this.keyFor(url)}.html`;
    writeFileSync(path, html);
  }
}

// Usage
const cache = new ScrapeCache();

async function fetchCached(url: string): Promise<string> {
  const cached = cache.get(url);
  if (cached) {
    console.log(`Cache hit: ${url}`);
    return cached;
  }

  const response = await fetch(url);
  const html = await response.text();
  cache.set(url, html);
  return html;
}
```

## Error Handling & Retries

```typescript
async function fetchWithRetry(
  url: string,
  maxRetries = 3,
  baseDelay = 1000
): Promise<string> {
  for (let attempt = 0; attempt <= maxRetries; attempt++) {
    try {
      const response = await fetch(url, {
        headers: { "User-Agent": getRandomUserAgent() },
        signal: AbortSignal.timeout(10000), // 10s timeout
      });

      if (response.status === 429) {
        // Rate limited — wait longer
        const retryAfter = parseInt(response.headers.get("retry-after") || "60");
        console.log(`Rate limited. Waiting ${retryAfter}s...`);
        await sleep(retryAfter * 1000);
        continue;
      }

      if (response.status === 403 || response.status === 503) {
        // Possibly blocked — exponential backoff
        const delay = baseDelay * Math.pow(2, attempt);
        console.log(`Blocked (${response.status}). Retrying in ${delay}ms...`);
        await sleep(delay);
        continue;
      }

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}`);
      }

      return await response.text();
    } catch (error: any) {
      if (attempt === maxRetries) throw error;

      const delay = baseDelay * Math.pow(2, attempt);
      console.log(`Error: ${error.message}. Retrying in ${delay}ms...`);
      await sleep(delay);
    }
  }

  throw new Error("Max retries exceeded");
}

const USER_AGENTS = [
  "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
  "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
  "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
  "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Safari/605.1.15",
];

function getRandomUserAgent() {
  return USER_AGENTS[Math.floor(Math.random() * USER_AGENTS.length)];
}
```

## Output Formats

```typescript
import { writeFileSync } from "fs";

// JSON output
function saveJSON(data: any[], filename: string) {
  writeFileSync(filename, JSON.stringify(data, null, 2));
}

// CSV output
function saveCSV(data: Record<string, any>[], filename: string) {
  if (data.length === 0) return;

  const headers = Object.keys(data[0]);
  const csvRows = [
    headers.join(","),
    ...data.map(row =>
      headers.map(h => {
        const val = String(row[h] ?? "");
        // Escape commas and quotes
        return val.includes(",") || val.includes('"') || val.includes("\n")
          ? `"${val.replace(/"/g, '""')}"`
          : val;
      }).join(",")
    ),
  ];

  writeFileSync(filename, csvRows.join("\n"));
}
```

## Gotchas

1. **Check for JavaScript rendering** — if `view-source:` shows the content but Cheerio doesn't find it, the content is rendered by JavaScript. Use Puppeteer/Playwright instead.

2. **`robots.txt` compliance** — always check `https://example.com/robots.txt` before scraping. Respect `Crawl-delay` directives.

3. **Relative URLs** — `$("a").attr("href")` might return `/page/2` not `https://example.com/page/2`. Always resolve: `new URL(href, baseUrl).href`.

4. **Text normalization** — `.text()` includes whitespace from HTML formatting. Always `.trim()` and consider `.replace(/\s+/g, " ")` for multi-line text.

5. **Encoding issues** — some pages use non-UTF-8 encoding. Check the `<meta charset>` tag and decode accordingly.

6. **Session cookies** — some sites require cookies from an initial page visit. Use a cookie jar or session management.
