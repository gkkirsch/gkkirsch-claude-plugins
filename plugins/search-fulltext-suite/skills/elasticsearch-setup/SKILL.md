---
name: elasticsearch-setup
description: >
  Set up Elasticsearch for advanced search with complex queries, aggregations,
  and analytics at scale.
  Triggers: "elasticsearch", "elastic search", "opensearch", "complex search",
  "aggregations", "search analytics", "large scale search".
  NOT for: simple search (use full-text-postgres or meilisearch-setup).
version: 1.0.0
argument-hint: "[index-name]"
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Elasticsearch Setup

Set up Elasticsearch for production search with complex queries, aggregations, and analytics.

## When to Use Elasticsearch

- Dataset > 1M documents or growing fast
- Need complex queries (nested, geo, range, fuzzy, aggregations)
- Need search analytics (what users search for, click-through rates)
- Need multi-language support with custom analyzers
- Need real-time aggregations (dashboards, facets, analytics)
- Already in the Elastic ecosystem (Kibana, Logstash, APM)

## Step 1: Install and Run

```bash
# Docker (recommended for local dev)
docker run -d --name elasticsearch \
  -p 9200:9200 -p 9300:9300 \
  -e "discovery.type=single-node" \
  -e "xpack.security.enabled=false" \
  -e "ES_JAVA_OPTS=-Xms512m -Xmx512m" \
  docker.elastic.co/elasticsearch/elasticsearch:8.12.0

# Verify
curl http://localhost:9200
```

### Install Client

```bash
npm install @elastic/elasticsearch
```

## Step 2: Client Setup

```typescript
// src/lib/elastic.ts
import { Client } from '@elastic/elasticsearch';

const client = new Client({
  node: process.env.ELASTICSEARCH_URL || 'http://localhost:9200',
  auth: process.env.ELASTICSEARCH_API_KEY ? {
    apiKey: process.env.ELASTICSEARCH_API_KEY,
  } : undefined,
  // For Elastic Cloud:
  // cloud: { id: process.env.ELASTIC_CLOUD_ID },
  // auth: { apiKey: process.env.ELASTIC_API_KEY },
});

export { client };
```

## Step 3: Create Index with Mappings

```typescript
// src/search/setup.ts
import { client } from '../lib/elastic';

export async function createProductIndex() {
  const indexExists = await client.indices.exists({ index: 'products' });
  if (indexExists) {
    console.log('products index already exists');
    return;
  }

  await client.indices.create({
    index: 'products',
    body: {
      settings: {
        number_of_shards: 1,        // Start with 1, scale later
        number_of_replicas: 0,       // 0 for dev, 1+ for production
        analysis: {
          analyzer: {
            product_analyzer: {
              type: 'custom',
              tokenizer: 'standard',
              filter: ['lowercase', 'english_stemmer', 'english_stop'],
            },
            autocomplete_analyzer: {
              type: 'custom',
              tokenizer: 'edge_ngram_tokenizer',
              filter: ['lowercase'],
            },
          },
          tokenizer: {
            edge_ngram_tokenizer: {
              type: 'edge_ngram',
              min_gram: 2,
              max_gram: 15,
              token_chars: ['letter', 'digit'],
            },
          },
          filter: {
            english_stemmer: { type: 'stemmer', language: 'english' },
            english_stop: { type: 'stop', stopwords: '_english_' },
          },
        },
      },
      mappings: {
        properties: {
          title: {
            type: 'text',
            analyzer: 'product_analyzer',
            fields: {
              autocomplete: { type: 'text', analyzer: 'autocomplete_analyzer' },
              keyword: { type: 'keyword' },  // For exact match and sorting
            },
          },
          brand: {
            type: 'text',
            fields: { keyword: { type: 'keyword' } },
          },
          description: { type: 'text', analyzer: 'product_analyzer' },
          category: { type: 'keyword' },
          subcategory: { type: 'keyword' },
          tags: { type: 'keyword' },
          price: { type: 'integer' },       // Store in cents
          rating: { type: 'float' },
          reviewCount: { type: 'integer' },
          inStock: { type: 'boolean' },
          imageUrl: { type: 'keyword', index: false },  // Not searchable
          createdAt: { type: 'date' },
          location: { type: 'geo_point' },  // For geo search
        },
      },
    },
  });

  console.log('products index created');
}
```

## Step 4: Index Documents

```typescript
// src/search/indexer.ts
import { client } from '../lib/elastic';

// Bulk index (for initial load or reindex)
export async function bulkIndexProducts(products: any[]) {
  const body = products.flatMap((doc) => [
    { index: { _index: 'products', _id: doc.id } },
    {
      title: doc.title,
      brand: doc.brand,
      description: doc.description,
      category: doc.category,
      tags: doc.tags,
      price: doc.price,
      rating: doc.rating,
      reviewCount: doc.reviewCount,
      inStock: doc.inStock,
      imageUrl: doc.imageUrl,
      createdAt: doc.createdAt,
    },
  ]);

  const result = await client.bulk({ body, refresh: false });

  if (result.errors) {
    const errors = result.items
      .filter((item: any) => item.index?.error)
      .map((item: any) => item.index.error);
    console.error(`Bulk index errors:`, errors.slice(0, 5));
  }

  return { indexed: products.length, errors: result.errors };
}

// Single document index
export async function indexProduct(product: any) {
  await client.index({
    index: 'products',
    id: product.id,
    body: {
      title: product.title,
      brand: product.brand,
      description: product.description,
      category: product.category,
      tags: product.tags,
      price: product.price,
      rating: product.rating,
      reviewCount: product.reviewCount,
      inStock: product.inStock,
      imageUrl: product.imageUrl,
      createdAt: product.createdAt,
    },
    refresh: 'wait_for', // Make immediately searchable
  });
}

// Delete document
export async function removeProduct(id: string) {
  await client.delete({ index: 'products', id, refresh: 'wait_for' });
}
```

## Step 5: Search Queries

```typescript
// src/services/search.service.ts
import { client } from '../lib/elastic';

interface SearchOptions {
  query: string;
  category?: string;
  brand?: string;
  minPrice?: number;
  maxPrice?: number;
  inStock?: boolean;
  sort?: string;
  page?: number;
  limit?: number;
}

export async function searchProducts(options: SearchOptions) {
  const {
    query, category, brand, minPrice, maxPrice,
    inStock, sort, page = 1, limit = 20,
  } = options;

  // Build bool query
  const must: any[] = [];
  const filter: any[] = [];

  if (query) {
    must.push({
      multi_match: {
        query,
        fields: ['title^3', 'brand^2', 'description', 'tags'],
        type: 'best_fields',
        fuzziness: 'AUTO',  // Typo tolerance
      },
    });
  }

  // Filters (don't affect relevance score)
  if (category) filter.push({ term: { category } });
  if (brand) filter.push({ term: { 'brand.keyword': brand } });
  if (inStock !== undefined) filter.push({ term: { inStock } });
  if (minPrice || maxPrice) {
    filter.push({
      range: {
        price: {
          ...(minPrice && { gte: minPrice }),
          ...(maxPrice && { lte: maxPrice }),
        },
      },
    });
  }

  // Sort
  const sortClause: any[] = [];
  if (sort === 'price_asc') sortClause.push({ price: 'asc' });
  else if (sort === 'price_desc') sortClause.push({ price: 'desc' });
  else if (sort === 'rating') sortClause.push({ rating: 'desc' });
  else if (sort === 'newest') sortClause.push({ createdAt: 'desc' });
  else if (query) sortClause.push({ _score: 'desc' }); // Relevance
  sortClause.push({ _id: 'asc' }); // Tiebreaker

  const result = await client.search({
    index: 'products',
    body: {
      query: {
        bool: {
          must: must.length > 0 ? must : [{ match_all: {} }],
          filter,
        },
      },
      sort: sortClause,
      from: (page - 1) * limit,
      size: limit,
      highlight: {
        fields: {
          title: {},
          description: { fragment_size: 150, number_of_fragments: 1 },
        },
        pre_tags: ['<mark>'],
        post_tags: ['</mark>'],
      },
      // Aggregations for facets
      aggs: {
        categories: { terms: { field: 'category', size: 20 } },
        brands: { terms: { field: 'brand.keyword', size: 20 } },
        price_ranges: {
          range: {
            field: 'price',
            ranges: [
              { to: 2500, key: 'Under $25' },
              { from: 2500, to: 5000, key: '$25-$50' },
              { from: 5000, to: 10000, key: '$50-$100' },
              { from: 10000, key: '$100+' },
            ],
          },
        },
        avg_rating: { avg: { field: 'rating' } },
      },
    },
  });

  return {
    hits: result.hits.hits.map((hit: any) => ({
      id: hit._id,
      score: hit._score,
      ...hit._source,
      highlight: hit.highlight,
    })),
    total: typeof result.hits.total === 'number'
      ? result.hits.total
      : result.hits.total?.value ?? 0,
    aggregations: result.aggregations,
    took: result.took,
  };
}
```

## Autocomplete

```typescript
// Fast autocomplete using edge ngrams
export async function autocomplete(prefix: string, limit = 5) {
  const result = await client.search({
    index: 'products',
    body: {
      query: {
        multi_match: {
          query: prefix,
          fields: ['title.autocomplete', 'brand.autocomplete'],
          type: 'bool_prefix',
        },
      },
      _source: ['title', 'brand', 'category'],
      size: limit,
    },
  });

  return result.hits.hits.map((hit: any) => ({
    text: hit._source.title,
    brand: hit._source.brand,
    category: hit._source.category,
  }));
}
```

## Zero-Downtime Reindexing

```typescript
export async function reindexWithAlias() {
  const timestamp = Date.now();
  const newIndex = `products_${timestamp}`;
  const alias = 'products';

  // 1. Create new index with updated settings
  await client.indices.create({
    index: newIndex,
    body: { /* settings + mappings */ },
  });

  // 2. Reindex all documents
  await client.reindex({
    body: {
      source: { index: alias },
      dest: { index: newIndex },
    },
    wait_for_completion: true,
  });

  // 3. Get old indexes behind the alias
  const aliasInfo = await client.indices.getAlias({ name: alias });
  const oldIndexes = Object.keys(aliasInfo);

  // 4. Atomic alias swap
  await client.indices.updateAliases({
    body: {
      actions: [
        ...oldIndexes.map((idx) => ({ remove: { index: idx, alias } })),
        { add: { index: newIndex, alias } },
      ],
    },
  });

  // 5. Delete old indexes
  for (const idx of oldIndexes) {
    if (idx !== newIndex) {
      await client.indices.delete({ index: idx });
    }
  }
}
```

## Gotchas

- **Mapping changes require reindex** — you can't change field types after creation. Use the alias swap pattern
- **`refresh: 'wait_for'`** makes documents immediately searchable but slows writes. Use `refresh: false` for bulk operations
- Default max result window is 10,000 (`from + size`). Use `search_after` for deep pagination
- **Shard count is fixed at creation** — start with 1 shard per 50GB of data. Over-sharding hurts performance
- Elasticsearch uses significant memory. Minimum 2GB heap (`-Xms2g -Xmx2g`). Give it 50% of available RAM, max 32GB
- `match` queries are analyzed (stemming, lowercase). `term` queries are exact match — use for keyword fields only
- Elastic Cloud pricing is based on RAM + storage. A minimal production setup starts at ~$50/month
- OpenSearch is API-compatible with Elasticsearch 7.x. Most code works with both
