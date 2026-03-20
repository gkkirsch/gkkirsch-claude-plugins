---
name: puppeteer-crawling
description: >
  Browser-based scraping with Puppeteer — dynamic pages, SPA scraping,
  form interaction, screenshots, PDF generation, infinite scroll,
  authentication, and anti-detection.
  Triggers: "puppeteer", "browser scraping", "dynamic scraping",
  "scrape spa", "scrape javascript", "headless browser",
  "browser automation scraping", "crawl dynamic", "infinite scroll scrape".
  NOT for: static HTML pages (use cheerio-scraping).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Puppeteer Browser Scraping

## Setup

```bash
npm install puppeteer          # includes Chromium (~400MB)
# Or:
npm install puppeteer-core     # BYO browser (smaller)
```

## Basic Usage

```typescript
import puppeteer from "puppeteer";

async function scrape(url: string) {
  const browser = await puppeteer.launch({
    headless: true,           // "new" headless mode (default in v21+)
    args: [
      "--no-sandbox",
      "--disable-setuid-sandbox",
      "--disable-dev-shm-usage",  // fix shared memory in Docker
    ],
  });

  try {
    const page = await browser.newPage();

    // Set viewport and user agent
    await page.setViewport({ width: 1280, height: 720 });
    await page.setUserAgent(
      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36"
    );

    // Navigate and wait for content
    await page.goto(url, {
      waitUntil: "networkidle2", // wait until 2 or fewer network requests
      timeout: 30000,
    });

    // Extract data
    const data = await page.evaluate(() => {
      const items: any[] = [];
      document.querySelectorAll(".product-card").forEach(el => {
        items.push({
          name: el.querySelector(".name")?.textContent?.trim() || "",
          price: el.querySelector(".price")?.textContent?.trim() || "",
          url: (el.querySelector("a") as HTMLAnchorElement)?.href || "",
        });
      });
      return items;
    });

    return data;
  } finally {
    await browser.close();
  }
}
```

## Wait Strategies

```typescript
// Wait for selector
await page.waitForSelector(".product-list", { timeout: 10000 });

// Wait for text
await page.waitForFunction(
  () => document.body.textContent?.includes("Results loaded"),
  { timeout: 10000 }
);

// Wait for navigation
await Promise.all([
  page.waitForNavigation({ waitUntil: "networkidle2" }),
  page.click("a.next-page"),
]);

// Wait for network request to complete
await page.waitForResponse(
  response => response.url().includes("/api/products") && response.ok()
);

// Wait for element count
await page.waitForFunction(
  (min) => document.querySelectorAll(".item").length >= min,
  { timeout: 15000 },
  10 // at least 10 items
);

// Custom polling wait
await page.waitForFunction(
  () => {
    const spinner = document.querySelector(".loading-spinner");
    return !spinner || spinner.getAttribute("style")?.includes("display: none");
  },
  { polling: 500, timeout: 10000 }
);
```

## Infinite Scroll

```typescript
async function scrapeInfiniteScroll(page: puppeteer.Page, maxItems = 100) {
  let items: any[] = [];
  let previousHeight = 0;
  let noNewContentCount = 0;

  while (items.length < maxItems && noNewContentCount < 3) {
    // Scroll to bottom
    await page.evaluate(() => window.scrollTo(0, document.body.scrollHeight));

    // Wait for new content to load
    await new Promise(resolve => setTimeout(resolve, 2000));

    // Check if new content appeared
    const currentHeight = await page.evaluate(() => document.body.scrollHeight);
    if (currentHeight === previousHeight) {
      noNewContentCount++;
    } else {
      noNewContentCount = 0;
    }
    previousHeight = currentHeight;

    // Extract current items
    items = await page.evaluate(() => {
      return Array.from(document.querySelectorAll(".feed-item")).map(el => ({
        text: el.querySelector(".content")?.textContent?.trim() || "",
        author: el.querySelector(".author")?.textContent?.trim() || "",
        timestamp: el.querySelector("time")?.getAttribute("datetime") || "",
      }));
    });

    console.log(`Found ${items.length} items...`);
  }

  return items.slice(0, maxItems);
}
```

## Form Interaction & Login

```typescript
async function loginAndScrape(url: string, email: string, password: string) {
  const browser = await puppeteer.launch({ headless: true });
  const page = await browser.newPage();

  // Navigate to login page
  await page.goto("https://example.com/login", { waitUntil: "networkidle2" });

  // Fill login form
  await page.type("#email", email, { delay: 50 });
  await page.type("#password", password, { delay: 50 });

  // Submit and wait for redirect
  await Promise.all([
    page.waitForNavigation({ waitUntil: "networkidle2" }),
    page.click('button[type="submit"]'),
  ]);

  // Verify login succeeded
  const isLoggedIn = await page.evaluate(() =>
    !!document.querySelector(".user-avatar")
  );
  if (!isLoggedIn) throw new Error("Login failed");

  // Now scrape authenticated content
  await page.goto(url, { waitUntil: "networkidle2" });

  // Save cookies for future sessions
  const cookies = await page.cookies();
  require("fs").writeFileSync(
    "cookies.json",
    JSON.stringify(cookies, null, 2)
  );

  // ... extract data

  await browser.close();
}

// Restore cookies in future sessions
async function loadCookies(page: puppeteer.Page) {
  const cookies = JSON.parse(
    require("fs").readFileSync("cookies.json", "utf-8")
  );
  await page.setCookie(...cookies);
}
```

## Network Interception

```typescript
// Block images and CSS for faster scraping
await page.setRequestInterception(true);
page.on("request", request => {
  const type = request.resourceType();
  if (["image", "stylesheet", "font", "media"].includes(type)) {
    request.abort();
  } else {
    request.continue();
  }
});

// Capture API responses
const apiData: any[] = [];
page.on("response", async response => {
  if (response.url().includes("/api/products")) {
    try {
      const json = await response.json();
      apiData.push(json);
    } catch {}
  }
});

await page.goto(url, { waitUntil: "networkidle2" });
console.log("Captured API data:", apiData);
```

## Screenshots & PDF

```typescript
// Screenshot
await page.screenshot({
  path: "screenshot.png",
  fullPage: true,           // capture full scrollable page
  type: "png",
});

// Element screenshot
const element = await page.$(".chart");
await element?.screenshot({ path: "chart.png" });

// PDF generation
await page.pdf({
  path: "output.pdf",
  format: "A4",
  printBackground: true,
  margin: { top: "1cm", right: "1cm", bottom: "1cm", left: "1cm" },
});
```

## Anti-Detection

```typescript
import puppeteer from "puppeteer-extra";
import StealthPlugin from "puppeteer-extra-plugin-stealth";

// Stealth plugin hides automation signals
puppeteer.use(StealthPlugin());

const browser = await puppeteer.launch({
  headless: true,
  args: [
    "--no-sandbox",
    "--disable-blink-features=AutomationControlled",
    "--disable-features=IsolateOrigins,site-per-process",
  ],
});

const page = await browser.newPage();

// Randomize viewport
const viewports = [
  { width: 1920, height: 1080 },
  { width: 1366, height: 768 },
  { width: 1536, height: 864 },
  { width: 1280, height: 720 },
];
const vp = viewports[Math.floor(Math.random() * viewports.length)];
await page.setViewport(vp);

// Override webdriver detection
await page.evaluateOnNewDocument(() => {
  Object.defineProperty(navigator, "webdriver", { get: () => false });

  // Fake plugins
  Object.defineProperty(navigator, "plugins", {
    get: () => [1, 2, 3, 4, 5],
  });

  // Fake languages
  Object.defineProperty(navigator, "languages", {
    get: () => ["en-US", "en"],
  });
});

// Human-like mouse movement
async function humanClick(page: puppeteer.Page, selector: string) {
  const element = await page.$(selector);
  if (!element) return;

  const box = await element.boundingBox();
  if (!box) return;

  // Move to element with slight randomness
  await page.mouse.move(
    box.x + box.width * Math.random(),
    box.y + box.height * Math.random(),
    { steps: 10 + Math.floor(Math.random() * 10) }
  );

  // Random delay before click
  await new Promise(r => setTimeout(r, 100 + Math.random() * 200));
  await page.mouse.click(
    box.x + box.width / 2,
    box.y + box.height / 2
  );
}

// Human-like typing
async function humanType(page: puppeteer.Page, selector: string, text: string) {
  await page.click(selector);
  for (const char of text) {
    await page.keyboard.type(char, {
      delay: 50 + Math.random() * 150,
    });
  }
}
```

## Proxy Support

```typescript
// Single proxy
const browser = await puppeteer.launch({
  args: ["--proxy-server=http://proxy.example.com:8080"],
});

// Authenticated proxy
const page = await browser.newPage();
await page.authenticate({
  username: "proxyuser",
  password: "proxypass",
});

// Rotating proxies
const proxies = [
  "http://proxy1.example.com:8080",
  "http://proxy2.example.com:8080",
  "http://proxy3.example.com:8080",
];

async function scrapeWithProxy(url: string) {
  const proxy = proxies[Math.floor(Math.random() * proxies.length)];
  const browser = await puppeteer.launch({
    args: [`--proxy-server=${proxy}`],
  });

  try {
    const page = await browser.newPage();
    await page.goto(url, { waitUntil: "networkidle2", timeout: 15000 });
    // ... scrape
  } finally {
    await browser.close();
  }
}
```

## Crawlee Framework (Large Scale)

```typescript
// npm install crawlee puppeteer
import { PuppeteerCrawler, Dataset, RequestQueue } from "crawlee";

const crawler = new PuppeteerCrawler({
  maxRequestsPerMinute: 30,       // rate limiting
  maxConcurrency: 5,              // parallel browsers
  requestHandlerTimeoutSecs: 60,
  maxRequestRetries: 3,

  async requestHandler({ page, request, enqueueLinks }) {
    console.log(`Processing: ${request.url}`);

    // Wait for content
    await page.waitForSelector(".product-list");

    // Extract data
    const products = await page.evaluate(() => {
      return Array.from(document.querySelectorAll(".product")).map(el => ({
        name: el.querySelector(".name")?.textContent?.trim(),
        price: el.querySelector(".price")?.textContent?.trim(),
        url: (el.querySelector("a") as HTMLAnchorElement)?.href,
      }));
    });

    // Save to dataset
    await Dataset.pushData(products);

    // Follow pagination links
    await enqueueLinks({
      selector: "a.next-page",
      label: "LIST",
    });
  },

  failedRequestHandler({ request }) {
    console.error(`Failed: ${request.url} (${request.errorMessages.join(", ")})`);
  },
});

// Start crawling
await crawler.run(["https://example.com/products"]);

// Export results
const dataset = await Dataset.open();
await dataset.exportToJSON("products.json");
```

## Gotchas

1. **Memory leaks** — always close browsers/pages in `finally` blocks. A leaked browser process consumes 100-500MB RAM.

2. **`page.evaluate()` runs in browser context** — you can't use Node.js modules or closures. Pass data as arguments: `page.evaluate((arg) => { ... }, myArg)`.

3. **`waitUntil: "networkidle2"` is not foolproof** — SPAs may continue making requests. Combine with `waitForSelector` or `waitForFunction` for reliability.

4. **Headless detection is real** — many sites detect headless Chrome. Use `puppeteer-extra-plugin-stealth` and `--disable-blink-features=AutomationControlled`.

5. **Docker needs `--no-sandbox`** — running Puppeteer in Docker without `--no-sandbox` and `--disable-dev-shm-usage` will crash. Also mount `/dev/shm` as tmpfs.

6. **Timeout defaults are generous** — the default navigation timeout is 30s. Set explicit timeouts to fail fast: `page.setDefaultNavigationTimeout(15000)`.
