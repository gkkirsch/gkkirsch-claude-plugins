---
name: scraping-resilience
description: >
  Web scraping resilience patterns including rate limiting, proxy rotation,
  anti-bot bypass, and data validation for production scrapers.
  Triggers: "scraping resilience", "rate limiting scraper", "proxy rotation",
  "anti-bot", "scraper reliability", "CAPTCHA handling", "request throttling",
  "scraper retry", "user agent rotation", "scraper monitoring".
  NOT for: Cheerio parsing (see cheerio-scraping), Puppeteer automation (see puppeteer-crawling), ethical/legal guidance.
version: 1.0.0
allowed-tools: Read, Grep, Glob, Bash
---

# Scraping Resilience Patterns

## Rate Limiter with Adaptive Throttling

```typescript
// lib/rate-limiter.ts — Respect rate limits and adapt to server responses

class AdaptiveRateLimiter {
  private requestTimes: number[] = [];
  private currentDelay: number;
  private readonly minDelay: number;
  private readonly maxDelay: number;
  private consecutiveErrors = 0;

  constructor(options: {
    requestsPerSecond: number;
    minDelay?: number;    // Floor delay in ms
    maxDelay?: number;    // Ceiling delay in ms
  }) {
    this.currentDelay = 1000 / options.requestsPerSecond;
    this.minDelay = options.minDelay ?? 200;
    this.maxDelay = options.maxDelay ?? 30_000;
  }

  async acquire(): Promise<void> {
    const now = Date.now();

    // Clean old request times (keep last 60 seconds)
    this.requestTimes = this.requestTimes.filter(t => now - t < 60_000);

    // Wait if we need to
    if (this.requestTimes.length > 0) {
      const lastRequest = this.requestTimes[this.requestTimes.length - 1];
      const elapsed = now - lastRequest;
      if (elapsed < this.currentDelay) {
        await this.sleep(this.currentDelay - elapsed);
      }
    }

    this.requestTimes.push(Date.now());
  }

  // Call on successful response
  onSuccess(): void {
    this.consecutiveErrors = 0;
    // Gradually speed up if we've been slowed
    if (this.currentDelay > this.minDelay) {
      this.currentDelay = Math.max(this.minDelay, this.currentDelay * 0.9);
    }
  }

  // Call on rate limit (429) or server error (5xx)
  onRateLimit(): void {
    this.consecutiveErrors++;
    // Exponential backoff
    this.currentDelay = Math.min(
      this.maxDelay,
      this.currentDelay * Math.pow(2, this.consecutiveErrors)
    );
    console.log(`Rate limited. New delay: ${Math.round(this.currentDelay)}ms`);
  }

  // Call on block/ban detection
  onBlocked(): void {
    this.currentDelay = this.maxDelay;
    console.warn(`Blocked! Backing off to ${this.maxDelay}ms`);
  }

  private sleep(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms));
  }
}
```

## Proxy Rotation

```typescript
// lib/proxy-pool.ts — Rotate proxies with health tracking

interface ProxyConfig {
  url: string;          // http://user:pass@host:port
  protocol: 'http' | 'https' | 'socks5';
  region?: string;
  lastUsed?: number;
  failures: number;
  totalRequests: number;
  banned: boolean;
}

class ProxyPool {
  private proxies: ProxyConfig[] = [];
  private currentIndex = 0;

  constructor(proxyUrls: string[]) {
    this.proxies = proxyUrls.map(url => ({
      url,
      protocol: url.startsWith('socks') ? 'socks5' : 'http',
      failures: 0,
      totalRequests: 0,
      banned: false,
    }));
  }

  // Get next available proxy (round-robin with health check)
  getNext(): ProxyConfig | null {
    const available = this.proxies.filter(p => !p.banned);
    if (available.length === 0) return null;

    // Sort by least recently used
    available.sort((a, b) => (a.lastUsed ?? 0) - (b.lastUsed ?? 0));
    const proxy = available[0];
    proxy.lastUsed = Date.now();
    proxy.totalRequests++;
    return proxy;
  }

  // Report success
  reportSuccess(proxyUrl: string): void {
    const proxy = this.proxies.find(p => p.url === proxyUrl);
    if (proxy) {
      proxy.failures = Math.max(0, proxy.failures - 1); // Heal on success
    }
  }

  // Report failure
  reportFailure(proxyUrl: string): void {
    const proxy = this.proxies.find(p => p.url === proxyUrl);
    if (proxy) {
      proxy.failures++;
      if (proxy.failures >= 5) {
        proxy.banned = true;
        console.warn(`Proxy banned: ${proxyUrl} (${proxy.failures} failures)`);

        // Auto-unban after 30 minutes
        setTimeout(() => {
          proxy.banned = false;
          proxy.failures = 0;
          console.log(`Proxy unbanned: ${proxyUrl}`);
        }, 30 * 60 * 1000);
      }
    }
  }

  // Health stats
  getStats(): { total: number; available: number; banned: number } {
    return {
      total: this.proxies.length,
      available: this.proxies.filter(p => !p.banned).length,
      banned: this.proxies.filter(p => p.banned).length,
    };
  }
}
```

## Request Retry with Circuit Breaker

```typescript
// lib/resilient-fetch.ts — Fetch with retries, timeouts, and circuit breaking

interface FetchOptions extends RequestInit {
  timeout?: number;
  maxRetries?: number;
  retryDelay?: number;
  proxy?: ProxyConfig;
}

class CircuitBreaker {
  private failures = 0;
  private lastFailure = 0;
  private state: 'closed' | 'open' | 'half-open' = 'closed';

  constructor(
    private threshold: number = 5,
    private resetTimeout: number = 60_000,
  ) {}

  canRequest(): boolean {
    if (this.state === 'closed') return true;
    if (this.state === 'open') {
      if (Date.now() - this.lastFailure > this.resetTimeout) {
        this.state = 'half-open';
        return true; // Allow one test request
      }
      return false;
    }
    return true; // half-open allows requests
  }

  onSuccess(): void {
    this.failures = 0;
    this.state = 'closed';
  }

  onFailure(): void {
    this.failures++;
    this.lastFailure = Date.now();
    if (this.failures >= this.threshold) {
      this.state = 'open';
      console.warn(`Circuit breaker OPEN after ${this.failures} failures`);
    }
  }
}

async function resilientFetch(
  url: string,
  options: FetchOptions = {},
): Promise<Response> {
  const { timeout = 10_000, maxRetries = 3, retryDelay = 1000, ...fetchOptions } = options;

  for (let attempt = 0; attempt <= maxRetries; attempt++) {
    try {
      const controller = new AbortController();
      const timer = setTimeout(() => controller.abort(), timeout);

      const response = await fetch(url, {
        ...fetchOptions,
        signal: controller.signal,
        headers: {
          'User-Agent': getRandomUserAgent(),
          'Accept': 'text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8',
          'Accept-Language': 'en-US,en;q=0.5',
          'Accept-Encoding': 'gzip, deflate',
          'Connection': 'keep-alive',
          ...fetchOptions.headers,
        },
      });

      clearTimeout(timer);

      if (response.status === 429) {
        const retryAfter = parseInt(response.headers.get('retry-after') ?? '60');
        console.warn(`429 rate limited. Waiting ${retryAfter}s`);
        await sleep(retryAfter * 1000);
        continue;
      }

      if (response.status >= 500 && attempt < maxRetries) {
        await sleep(retryDelay * Math.pow(2, attempt));
        continue;
      }

      return response;
    } catch (error) {
      if (attempt === maxRetries) throw error;
      const isTimeout = (error as Error).name === 'AbortError';
      console.warn(`Attempt ${attempt + 1} failed: ${isTimeout ? 'timeout' : (error as Error).message}`);
      await sleep(retryDelay * Math.pow(2, attempt));
    }
  }

  throw new Error(`Failed after ${maxRetries + 1} attempts`);
}

// User agent pool
function getRandomUserAgent(): string {
  const agents = [
    'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36',
    'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36',
    'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.3 Safari/605.1.15',
    'Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:123.0) Gecko/20100101 Firefox/123.0',
    'Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36',
  ];
  return agents[Math.floor(Math.random() * agents.length)];
}

function sleep(ms: number): Promise<void> {
  return new Promise(resolve => setTimeout(resolve, ms));
}
```

## Data Validation & Deduplication

```typescript
// lib/data-validator.ts — Validate and deduplicate scraped data

import { z } from 'zod';

const ScrapedProductSchema = z.object({
  url: z.string().url(),
  title: z.string().min(1).max(500),
  price: z.number().positive().max(1_000_000),
  currency: z.string().length(3),
  description: z.string().max(10_000).optional(),
  images: z.array(z.string().url()).max(20),
  sku: z.string().optional(),
  inStock: z.boolean(),
  scrapedAt: z.date(),
});

type ScrapedProduct = z.infer<typeof ScrapedProductSchema>;

class DataPipeline<T extends z.ZodType> {
  private seen = new Set<string>();
  private valid: z.infer<T>[] = [];
  private invalid: { data: unknown; errors: string[] }[] = [];

  constructor(private schema: T, private dedupeKey: (item: z.infer<T>) => string) {}

  process(rawData: unknown[]): { valid: z.infer<T>[]; invalid: typeof this.invalid; duplicates: number } {
    let duplicates = 0;

    for (const item of rawData) {
      const result = this.schema.safeParse(item);

      if (!result.success) {
        this.invalid.push({
          data: item,
          errors: result.error.errors.map(e => `${e.path.join('.')}: ${e.message}`),
        });
        continue;
      }

      const key = this.dedupeKey(result.data);
      if (this.seen.has(key)) {
        duplicates++;
        continue;
      }

      this.seen.add(key);
      this.valid.push(result.data);
    }

    return { valid: this.valid, invalid: this.invalid, duplicates };
  }
}

// Usage:
// const pipeline = new DataPipeline(ScrapedProductSchema, (item) => item.url);
// const results = pipeline.process(scrapedItems);
// console.log(`Valid: ${results.valid.length}, Invalid: ${results.invalid.length}, Dupes: ${results.duplicates}`);
```

## Scraper Monitoring

```typescript
// lib/scraper-monitor.ts — Track scraper health metrics

interface ScraperMetrics {
  startedAt: Date;
  pagesRequested: number;
  pagesSucceeded: number;
  pagesFailed: number;
  itemsExtracted: number;
  itemsValidated: number;
  bytesDownloaded: number;
  errorsBy Type: Map<string, number>;
  avgResponseMs: number;
  responseTimes: number[];
}

class ScraperMonitor {
  private metrics: ScraperMetrics = {
    startedAt: new Date(),
    pagesRequested: 0,
    pagesSucceeded: 0,
    pagesFailed: 0,
    itemsExtracted: 0,
    itemsValidated: 0,
    bytesDownloaded: 0,
    errorsBy Type: new Map(),
    avgResponseMs: 0,
    responseTimes: [],
  };

  recordRequest(): void { this.metrics.pagesRequested++; }

  recordSuccess(responseMs: number, bytes: number): void {
    this.metrics.pagesSucceeded++;
    this.metrics.bytesDownloaded += bytes;
    this.metrics.responseTimes.push(responseMs);
    this.metrics.avgResponseMs = this.metrics.responseTimes.reduce((a, b) => a + b, 0) / this.metrics.responseTimes.length;
  }

  recordFailure(errorType: string): void {
    this.metrics.pagesFailed++;
    this.metrics.errorsBy Type.set(errorType, (this.metrics.errorsBy Type.get(errorType) ?? 0) + 1);
  }

  recordItems(extracted: number, validated: number): void {
    this.metrics.itemsExtracted += extracted;
    this.metrics.itemsValidated += validated;
  }

  getReport(): string {
    const m = this.metrics;
    const elapsed = (Date.now() - m.startedAt.getTime()) / 1000;
    const successRate = m.pagesRequested > 0
      ? Math.round((m.pagesSucceeded / m.pagesRequested) * 100)
      : 0;

    return [
      `=== Scraper Report ===`,
      `Duration: ${Math.round(elapsed)}s`,
      `Pages: ${m.pagesSucceeded}/${m.pagesRequested} (${successRate}% success)`,
      `Items: ${m.itemsExtracted} extracted, ${m.itemsValidated} valid`,
      `Data: ${(m.bytesDownloaded / 1024 / 1024).toFixed(1)} MB`,
      `Speed: ${(m.pagesSucceeded / elapsed * 60).toFixed(1)} pages/min`,
      `Avg response: ${Math.round(m.avgResponseMs)}ms`,
      m.errorsBy Type.size > 0
        ? `Errors: ${[...m.errorsBy Type.entries()].map(([k, v]) => `${k}=${v}`).join(', ')}`
        : 'Errors: none',
    ].join('\n');
  }

  // Alert if scraper is unhealthy
  shouldAlert(): { alert: boolean; reason: string } {
    const m = this.metrics;
    if (m.pagesRequested >= 10 && m.pagesSucceeded / m.pagesRequested < 0.5) {
      return { alert: true, reason: 'Success rate below 50%' };
    }
    if (m.avgResponseMs > 10000) {
      return { alert: true, reason: 'Average response time above 10s' };
    }
    if (m.pagesFailed > 20) {
      return { alert: true, reason: 'More than 20 failures' };
    }
    return { alert: false, reason: '' };
  }
}
```

## Gotchas

1. **Fixed delays look like bots** -- A scraper that requests a page every exactly 2.000 seconds is obviously automated. Real humans have variable timing. Add random jitter: `baseDelay + Math.random() * baseDelay * 0.5`. Vary delays between 1-5 seconds for browsing, 0.5-2 seconds for API endpoints.

2. **Ignoring robots.txt and rate limits** -- Even if you can bypass restrictions, ignoring `robots.txt` and `Retry-After` headers can get your IP range permanently banned, or worse, result in legal action. Parse `robots.txt` for crawl-delay, respect `Retry-After` headers, and check the site's ToS before scraping.

3. **Session/cookie accumulation** -- Long-running scrapers accumulate cookies that mark them as returning visitors, making anti-bot fingerprinting easier. Rotate sessions: create a fresh cookie jar every 50-100 requests. Don't carry cookies from domain A to domain B.

4. **Single point of failure with one IP** -- All requests from one IP are trivially rate-limited or banned. Even rotating User-Agents doesn't help if the IP is flagged. Use proxy rotation with diverse IP ranges. Mix residential and datacenter proxies — residential proxies are harder to detect but slower.

5. **Scraping dynamic content without waiting** -- Single-page apps render content via JavaScript after page load. Fetching the raw HTML returns empty containers. Use Puppeteer/Playwright with proper `waitForSelector` or `waitForNetworkIdle`. Check that the data you need is actually in the HTML, not loaded via XHR.

6. **No data validation pipeline** -- Scraping without validation produces garbage data silently. A selector change on the target site can cause every price to be "undefined" or every title to be a nav link. Validate every scraped item against a schema immediately. Alert when validation failure rate exceeds 10%.
