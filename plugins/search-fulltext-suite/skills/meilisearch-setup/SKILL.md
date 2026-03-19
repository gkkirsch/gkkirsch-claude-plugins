---
name: meilisearch-setup
description: >
  Set up Meilisearch for typo-tolerant, instant search with minimal configuration.
  Triggers: "meilisearch", "instant search", "typo tolerant search",
  "search engine setup", "fast search", "search as you type".
  NOT for: PostgreSQL search (use full-text-postgres), Elasticsearch (use elasticsearch-setup).
version: 1.0.0
argument-hint: "[index-name]"
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Meilisearch Setup

Set up typo-tolerant, instant search with Meilisearch. Works out of the box with minimal configuration.

## Why Meilisearch

- **Typo tolerance** built-in — "headpones" matches "headphones"
- **< 50ms** search responses, even with millions of documents
- **Zero configuration** to start — just index and search
- **Faceted search** and filtering built-in
- **RESTful API** — no query DSL to learn
- Free and open source (self-hosted) or Meilisearch Cloud ($30+/mo)

## Step 1: Install and Run

### Local Development

```bash
# macOS
brew install meilisearch
meilisearch --master-key="your-master-key-here"

# Docker
docker run -d -p 7700:7700 \
  -e MEILI_MASTER_KEY="your-master-key-here" \
  -v meili_data:/meili_data \
  getmeili/meilisearch:latest
```

### Production (Heroku)

```bash
# Use the Meilisearch Cloud add-on or self-host on a VPS
# Meilisearch Cloud: https://cloud.meilisearch.com
# Railway: one-click deploy available
```

### Install SDK

```bash
npm install meilisearch
```

## Step 2: Client Setup

```typescript
// src/lib/search.ts
import { MeiliSearch } from 'meilisearch';

const client = new MeiliSearch({
  host: process.env.MEILISEARCH_URL || 'http://localhost:7700',
  apiKey: process.env.MEILISEARCH_API_KEY || 'your-master-key',
});

export { client };
```

## Step 3: Index Configuration

```typescript
// src/search/setup.ts
import { client } from '../lib/search';

export async function setupSearchIndexes() {
  // Create or get the products index
  const index = client.index('products');

  // Configure searchable attributes (order = priority)
  await index.updateSettings({
    searchableAttributes: [
      'title',        // Highest priority
      'brand',
      'description',
      'tags',         // Lowest priority
    ],

    // Attributes returned in search results
    displayedAttributes: [
      'id', 'title', 'brand', 'price', 'imageUrl',
      'category', 'rating', 'description',
    ],

    // Attributes available for filtering
    filterableAttributes: [
      'category', 'brand', 'price', 'rating', 'inStock', 'tags',
    ],

    // Attributes available for sorting
    sortableAttributes: [
      'price', 'rating', 'createdAt',
    ],

    // Ranking rules (default is usually fine)
    rankingRules: [
      'words',        // Number of matching words
      'typo',         // Number of typos
      'proximity',    // Distance between matching words
      'attribute',    // Attribute priority (searchableAttributes order)
      'sort',         // User-requested sort
      'exactness',    // Exact vs prefix match
    ],

    // Typo tolerance settings
    typoTolerance: {
      enabled: true,
      minWordSizeForTypos: {
        oneTypo: 4,   // Words with 4+ chars allow 1 typo
        twoTypos: 8,  // Words with 8+ chars allow 2 typos
      },
    },

    // Pagination
    pagination: {
      maxTotalHits: 1000,
    },
  });

  console.log('Search indexes configured');
}
```

## Step 4: Index Documents

```typescript
// src/search/indexer.ts
import { client } from '../lib/search';
import { db } from '../lib/db';

const BATCH_SIZE = 1000;

// Index all products (initial or full reindex)
export async function indexAllProducts() {
  const index = client.index('products');
  let offset = 0;
  let totalIndexed = 0;

  while (true) {
    const products = await db.product.findMany({
      where: { active: true },
      skip: offset,
      take: BATCH_SIZE,
      select: {
        id: true,
        title: true,
        brand: true,
        description: true,
        category: true,
        price: true,
        rating: true,
        reviewCount: true,
        inStock: true,
        tags: true,
        imageUrl: true,
        createdAt: true,
      },
    });

    if (products.length === 0) break;

    // Transform for search index
    const documents = products.map((p) => ({
      id: p.id,
      title: p.title,
      brand: p.brand,
      description: p.description?.substring(0, 500), // Limit description length
      category: p.category,
      price: p.price,
      rating: p.rating,
      reviewCount: p.reviewCount,
      inStock: p.inStock,
      tags: p.tags,
      imageUrl: p.imageUrl,
      createdAt: p.createdAt.toISOString(),
    }));

    const task = await index.addDocuments(documents);
    console.log(`Indexed batch: ${documents.length} docs (task: ${task.taskUid})`);

    totalIndexed += documents.length;
    offset += BATCH_SIZE;
  }

  console.log(`Total indexed: ${totalIndexed} products`);
}

// Index single document (on create/update)
export async function indexProduct(product: any) {
  const index = client.index('products');
  await index.addDocuments([{
    id: product.id,
    title: product.title,
    brand: product.brand,
    description: product.description?.substring(0, 500),
    category: product.category,
    price: product.price,
    rating: product.rating,
    inStock: product.inStock,
    tags: product.tags,
    imageUrl: product.imageUrl,
    createdAt: product.createdAt.toISOString(),
  }]);
}

// Remove document from index
export async function removeProduct(productId: string) {
  const index = client.index('products');
  await index.deleteDocument(productId);
}
```

## Step 5: Search API

```typescript
// src/routes/search.ts
import { Router } from 'express';
import { client } from '../lib/search';

const router = Router();

router.get('/api/search', async (req, res) => {
  const {
    q = '',
    category,
    brand,
    minPrice,
    maxPrice,
    sort,
    page = '1',
    limit = '20',
  } = req.query;

  const index = client.index('products');

  // Build filter array
  const filters: string[] = [];
  if (category) filters.push(`category = "${category}"`);
  if (brand) filters.push(`brand = "${brand}"`);
  if (minPrice) filters.push(`price >= ${minPrice}`);
  if (maxPrice) filters.push(`price <= ${maxPrice}`);
  filters.push('inStock = true');

  // Build sort
  const sortRules: string[] = [];
  if (sort === 'price_asc') sortRules.push('price:asc');
  if (sort === 'price_desc') sortRules.push('price:desc');
  if (sort === 'rating') sortRules.push('rating:desc');
  if (sort === 'newest') sortRules.push('createdAt:desc');

  const results = await index.search(q as string, {
    filter: filters.length > 0 ? filters : undefined,
    sort: sortRules.length > 0 ? sortRules : undefined,
    limit: parseInt(limit as string),
    offset: (parseInt(page as string) - 1) * parseInt(limit as string),
    attributesToHighlight: ['title', 'description'],
    highlightPreTag: '<mark>',
    highlightPostTag: '</mark>',
    facets: ['category', 'brand'],
  });

  res.json({
    hits: results.hits,
    total: results.estimatedTotalHits,
    page: parseInt(page as string),
    processingTimeMs: results.processingTimeMs,
    facets: results.facetDistribution,
  });
});

export default router;
```

## Step 6: Keep Index in Sync

```typescript
// In your product CRUD routes/services
import { indexProduct, removeProduct } from '../search/indexer';

// After creating a product
const product = await db.product.create({ data: productData });
await indexProduct(product); // Fire and forget, or queue with BullMQ

// After updating a product
const updated = await db.product.update({ where: { id }, data: updates });
await indexProduct(updated);

// After deleting a product
await db.product.delete({ where: { id } });
await removeProduct(id);
```

### Async Indexing with BullMQ (Recommended for Production)

```typescript
import { Queue } from 'bullmq';
const indexQueue = new Queue('search-index', { connection });

// In your CRUD operations
await indexQueue.add('index-product', { productId: product.id });
await indexQueue.add('remove-product', { productId: id });

// Worker
new Worker('search-index', async (job) => {
  if (job.name === 'index-product') {
    const product = await db.product.findUnique({ where: { id: job.data.productId } });
    if (product) await indexProduct(product);
  } else if (job.name === 'remove-product') {
    await removeProduct(job.data.productId);
  }
}, { connection });
```

## Multi-Index Search (Federated)

```typescript
// Search across multiple indexes simultaneously
const results = await client.multiSearch({
  queries: [
    { indexUid: 'products', q: 'wireless headphones', limit: 5 },
    { indexUid: 'articles', q: 'wireless headphones', limit: 3 },
    { indexUid: 'brands', q: 'wireless headphones', limit: 2 },
  ],
});

// results.results[0] = product hits
// results.results[1] = article hits
// results.results[2] = brand hits
```

## API Key Management

```typescript
// Master key: full access (never expose to frontend)
// Search key: read-only search (safe for frontend)
// Admin key: index management (backend only)

// Generate API keys programmatically
const searchKey = await client.createKey({
  description: 'Search-only key for frontend',
  actions: ['search'],
  indexes: ['products'],
  expiresAt: null, // Never expires
});

// Use search key in frontend
const frontendClient = new MeiliSearch({
  host: 'https://your-meilisearch.com',
  apiKey: searchKey.key, // Safe to expose
});
```

## Gotchas

- Meilisearch is **eventually consistent** — documents aren't searchable instantly after indexing. Use `await index.waitForTask(taskUid)` if you need to confirm
- **Primary key** is auto-detected from `id` field. If your documents use a different field name, specify it: `index.addDocuments(docs, { primaryKey: 'productId' })`
- Default result limit is 20 hits, max `pagination.maxTotalHits` is 1000. For pagination beyond 1000, use `offset` + `limit`
- Filter syntax uses SQL-like operators: `=`, `!=`, `>`, `<`, `>=`, `<=`, `IN`, `TO`, `AND`, `OR`, `NOT`
- Meilisearch stores ALL data in RAM + disk. With 1M documents, expect 2-4GB RAM usage
- Index updates are **atomic** per batch. If you add 1000 docs and Meilisearch crashes mid-batch, none of the 1000 are indexed (no partial state)
- Meilisearch Cloud free tier: 100K documents, 10K searches/month. Sufficient for development and small apps
