# Web Scraping Cheatsheet

## Tool Selection

| Content Type | Tool | Install |
|-------------|------|---------|
| Static HTML | Cheerio + fetch | `npm i cheerio` |
| JS-rendered | Puppeteer | `npm i puppeteer` |
| Large scale | Crawlee | `npm i crawlee puppeteer` |
| Has API | Direct fetch | built-in |

## Cheerio Basics

```typescript
import * as cheerio from "cheerio"
const $ = cheerio.load(html)

// Select
$("h1").text()                    // text content
$("a").attr("href")               // attribute
$(".card").length                  // count
$(".card").each((i, el) => {      // iterate
  $(el).find(".title").text()
})

// Selectors
$("a[href^='https']")             // starts with
$("a:contains('Read more')")      // text content
$("div:has(img)")                 // has child
$("li:nth-child(2)")              // positional
$("tr > td:first-child")         // direct child
```

## Puppeteer Basics

```typescript
import puppeteer from "puppeteer"

const browser = await puppeteer.launch({ headless: true })
const page = await browser.newPage()
await page.goto(url, { waitUntil: "networkidle2" })

// Extract
const data = await page.evaluate(() => {
  return document.querySelector("h1")?.textContent
})

// Wait
await page.waitForSelector(".loaded")
await page.waitForFunction(() => document.querySelectorAll(".item").length > 5)

// Interact
await page.type("#search", "query", { delay: 50 })
await page.click("button.submit")

await browser.close()
```

## Infinite Scroll

```typescript
let prevHeight = 0
while (true) {
  await page.evaluate(() => window.scrollTo(0, document.body.scrollHeight))
  await new Promise(r => setTimeout(r, 2000))
  const height = await page.evaluate(() => document.body.scrollHeight)
  if (height === prevHeight) break
  prevHeight = height
}
```

## Pagination

```typescript
let page = 1, hasMore = true
while (hasMore) {
  const $ = cheerio.load(await fetch(`${url}?page=${page}`).then(r => r.text()))
  items.push(...extract($))
  hasMore = $("a.next").length > 0
  page++
  await sleep(1500 + Math.random() * 1500) // rate limit
}
```

## Rate Limiting

```typescript
function sleep(ms: number) {
  return new Promise(r => setTimeout(r, ms))
}

// Between requests: 1-3s random delay
await sleep(1000 + Math.random() * 2000)
```

## Retry with Backoff

```typescript
for (let attempt = 0; attempt <= 3; attempt++) {
  try {
    const resp = await fetch(url)
    if (resp.status === 429) {
      await sleep(60000) // rate limited
      continue
    }
    return await resp.text()
  } catch {
    await sleep(1000 * Math.pow(2, attempt)) // exponential backoff
  }
}
```

## Block Resources (Faster Scraping)

```typescript
await page.setRequestInterception(true)
page.on("request", req => {
  if (["image", "stylesheet", "font"].includes(req.resourceType())) {
    req.abort()
  } else {
    req.continue()
  }
})
```

## Capture API Responses

```typescript
page.on("response", async resp => {
  if (resp.url().includes("/api/data")) {
    const json = await resp.json()
    console.log(json)
  }
})
```

## Anti-Detection

```bash
npm i puppeteer-extra puppeteer-extra-plugin-stealth
```

```typescript
import puppeteer from "puppeteer-extra"
import Stealth from "puppeteer-extra-plugin-stealth"
puppeteer.use(Stealth())

// Also:
await page.evaluateOnNewDocument(() => {
  Object.defineProperty(navigator, "webdriver", { get: () => false })
})
```

## Proxy

```typescript
const browser = await puppeteer.launch({
  args: ["--proxy-server=http://proxy:8080"],
})
await page.authenticate({ username: "user", password: "pass" })
```

## Save Output

```typescript
// JSON
fs.writeFileSync("data.json", JSON.stringify(data, null, 2))

// CSV
const csv = [
  Object.keys(data[0]).join(","),
  ...data.map(r => Object.values(r).map(v => `"${v}"`).join(","))
].join("\n")
fs.writeFileSync("data.csv", csv)
```

## Login & Cookies

```typescript
// Login
await page.type("#email", "user@test.com")
await page.type("#password", "pass")
await page.click("button[type=submit]")
await page.waitForNavigation()

// Save cookies
const cookies = await page.cookies()
fs.writeFileSync("cookies.json", JSON.stringify(cookies))

// Restore cookies
const cookies = JSON.parse(fs.readFileSync("cookies.json", "utf-8"))
await page.setCookie(...cookies)
```

## Ethical Rules

1. Check robots.txt first
2. Rate limit: 1-2 req/sec, random delays
3. Identify yourself in User-Agent
4. Cache to avoid repeat requests
5. Prefer APIs over scraping
6. Respect Terms of Service
