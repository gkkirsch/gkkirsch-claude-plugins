---
name: full-text-postgres
description: >
  Implement full-text search using PostgreSQL built-in capabilities.
  Triggers: "postgres search", "full text search", "tsvector", "postgresql fts",
  "search without elasticsearch", "search in postgres", "text search".
  NOT for: dedicated search engines (use meilisearch-setup or elasticsearch-setup).
version: 1.0.0
argument-hint: "[table-name]"
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# PostgreSQL Full-Text Search

Implement powerful search using PostgreSQL's built-in full-text search. No external dependencies — works with your existing database.

## When to Use Postgres FTS

- Already using PostgreSQL (no new infrastructure)
- Dataset < 1-5 million documents
- Query latency < 200ms is acceptable
- Don't need typo tolerance (Postgres FTS is exact-match + stemming)
- Simple search requirements (no complex facets or ML ranking)

## Step 1: Add Search Column

```sql
-- Add a tsvector column for search
ALTER TABLE products ADD COLUMN search_vector tsvector;

-- Create GIN index for fast full-text queries
CREATE INDEX idx_products_search ON products USING GIN(search_vector);

-- Populate the search vector (weighted fields)
UPDATE products SET search_vector =
  setweight(to_tsvector('english', coalesce(title, '')), 'A') ||
  setweight(to_tsvector('english', coalesce(brand, '')), 'B') ||
  setweight(to_tsvector('english', coalesce(description, '')), 'C') ||
  setweight(to_tsvector('english', coalesce(tags::text, '')), 'D');
```

Weight priority: A (highest) > B > C > D (lowest). Title matches rank higher than description.

## Step 2: Auto-Update with Trigger

```sql
-- Function to update search vector on insert/update
CREATE OR REPLACE FUNCTION products_search_trigger() RETURNS trigger AS $$
BEGIN
  NEW.search_vector :=
    setweight(to_tsvector('english', coalesce(NEW.title, '')), 'A') ||
    setweight(to_tsvector('english', coalesce(NEW.brand, '')), 'B') ||
    setweight(to_tsvector('english', coalesce(NEW.description, '')), 'C') ||
    setweight(to_tsvector('english', coalesce(NEW.tags::text, '')), 'D');
  RETURN NEW;
END
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_products_search
  BEFORE INSERT OR UPDATE OF title, brand, description, tags
  ON products
  FOR EACH ROW
  EXECUTE FUNCTION products_search_trigger();
```

## Step 3: Query with Ranking

```sql
-- Basic search with ranking
SELECT
  id,
  title,
  brand,
  ts_rank(search_vector, query) AS rank,
  ts_headline('english', description, query,
    'MaxWords=30, MinWords=15, StartSel=<mark>, StopSel=</mark>'
  ) AS highlighted_description
FROM products, plainto_tsquery('english', 'wireless headphones') AS query
WHERE search_vector @@ query
ORDER BY rank DESC
LIMIT 20;
```

## Step 4: Node.js/Express Integration

```typescript
// src/services/search.service.ts
import { Pool } from 'pg';

const pool = new Pool();

interface SearchResult {
  id: string;
  title: string;
  brand: string;
  price: number;
  rank: number;
  highlight: string;
}

interface SearchOptions {
  query: string;
  category?: string;
  minPrice?: number;
  maxPrice?: number;
  limit?: number;
  offset?: number;
}

export async function searchProducts(options: SearchOptions): Promise<{
  results: SearchResult[];
  total: number;
}> {
  const { query, category, minPrice, maxPrice, limit = 20, offset = 0 } = options;

  // Build parameterized query
  const params: any[] = [query];
  let paramIndex = 2;

  let whereClause = `search_vector @@ plainto_tsquery('english', $1)`;

  if (category) {
    whereClause += ` AND category = $${paramIndex}`;
    params.push(category);
    paramIndex++;
  }

  if (minPrice !== undefined) {
    whereClause += ` AND price >= $${paramIndex}`;
    params.push(minPrice);
    paramIndex++;
  }

  if (maxPrice !== undefined) {
    whereClause += ` AND price <= $${paramIndex}`;
    params.push(maxPrice);
    paramIndex++;
  }

  // Get results with count
  const countQuery = `SELECT COUNT(*) FROM products WHERE ${whereClause}`;
  const searchQuery = `
    SELECT
      id, title, brand, price, image_url,
      ts_rank(search_vector, plainto_tsquery('english', $1)) AS rank,
      ts_headline('english', description, plainto_tsquery('english', $1),
        'MaxWords=30, MinWords=15, StartSel=<mark>, StopSel=</mark>'
      ) AS highlight
    FROM products
    WHERE ${whereClause}
    ORDER BY rank DESC
    LIMIT $${paramIndex} OFFSET $${paramIndex + 1}
  `;
  params.push(limit, offset);

  const [countResult, searchResult] = await Promise.all([
    pool.query(countQuery, params.slice(0, -2)),
    pool.query(searchQuery, params),
  ]);

  return {
    results: searchResult.rows,
    total: parseInt(countResult.rows[0].count),
  };
}
```

### API Endpoint

```typescript
// src/routes/search.ts
import { Router } from 'express';
import { searchProducts } from '../services/search.service';

const router = Router();

router.get('/api/search', async (req, res) => {
  const {
    q,
    category,
    minPrice,
    maxPrice,
    page = '1',
    limit = '20',
  } = req.query;

  if (!q || typeof q !== 'string') {
    return res.status(400).json({ error: 'Query parameter "q" is required' });
  }

  const offset = (parseInt(page as string) - 1) * parseInt(limit as string);

  const { results, total } = await searchProducts({
    query: q,
    category: category as string,
    minPrice: minPrice ? parseInt(minPrice as string) : undefined,
    maxPrice: maxPrice ? parseInt(maxPrice as string) : undefined,
    limit: parseInt(limit as string),
    offset,
  });

  res.json({
    results,
    total,
    page: parseInt(page as string),
    totalPages: Math.ceil(total / parseInt(limit as string)),
  });
});

export default router;
```

## Advanced Query Types

### Phrase Search

```sql
-- Exact phrase matching
SELECT * FROM products
WHERE search_vector @@ phraseto_tsquery('english', 'noise cancelling headphones');
```

### Prefix Search (Autocomplete)

```sql
-- Match words starting with prefix
SELECT * FROM products
WHERE search_vector @@ to_tsquery('english', 'wire:*');
-- Matches: wireless, wired, wiring
```

### Boolean Queries

```sql
-- AND: both terms required
SELECT * FROM products
WHERE search_vector @@ to_tsquery('english', 'wireless & bluetooth');

-- OR: either term
SELECT * FROM products
WHERE search_vector @@ to_tsquery('english', 'wireless | bluetooth');

-- NOT: exclude term
SELECT * FROM products
WHERE search_vector @@ to_tsquery('english', 'wireless & !earbuds');

-- Proximity: words near each other
SELECT * FROM products
WHERE search_vector @@ to_tsquery('english', 'wireless <2> headphones');
-- "wireless" within 2 words of "headphones"
```

## Prisma Integration

```typescript
// With Prisma — use raw queries for FTS
const results = await prisma.$queryRaw<SearchResult[]>`
  SELECT
    id, title, brand, price,
    ts_rank(search_vector, plainto_tsquery('english', ${query})) AS rank
  FROM products
  WHERE search_vector @@ plainto_tsquery('english', ${query})
  ORDER BY rank DESC
  LIMIT ${limit} OFFSET ${offset}
`;
```

## Trigram Similarity (Fuzzy Matching)

For typo tolerance, add `pg_trgm` extension:

```sql
-- Enable trigram extension
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Create trigram index
CREATE INDEX idx_products_title_trgm ON products USING GIN(title gin_trgm_ops);

-- Fuzzy search with similarity threshold
SELECT title, similarity(title, 'headpones') AS sim
FROM products
WHERE title % 'headpones'  -- % operator uses default threshold (0.3)
ORDER BY sim DESC
LIMIT 10;

-- Adjust threshold
SET pg_trgm.similarity_threshold = 0.2;  -- More fuzzy

-- Combined: full-text + fuzzy fallback
SELECT id, title,
  COALESCE(ts_rank(search_vector, plainto_tsquery('english', $1)), 0) AS fts_rank,
  similarity(title, $1) AS fuzzy_rank
FROM products
WHERE search_vector @@ plainto_tsquery('english', $1)
   OR title % $1
ORDER BY fts_rank DESC, fuzzy_rank DESC
LIMIT 20;
```

## Performance Tips

| Tip | Impact |
|-----|--------|
| Use GIN index (not GiST) for search_vector | 10x faster reads, slower writes |
| Pre-compute tsvector column with trigger | Avoid computing on every query |
| Use `plainto_tsquery` for user input | Handles spaces/punctuation safely |
| Add `LIMIT` to all search queries | Prevents scanning entire index |
| Partial index for active records | `WHERE active = true` reduces index size |
| Separate search from filtering | Use B-tree indexes for filters (category, price) |

## Gotchas

- `to_tsquery` requires valid query syntax (AND, OR, NOT). Use `plainto_tsquery` for raw user input — it's safer
- Stemming means "running" matches "run", "runs", "runner" — this is usually desirable but can surprise
- Postgres FTS does NOT support typo tolerance natively. Add `pg_trgm` for fuzzy matching
- `ts_headline` is expensive — don't call it on large result sets. Limit to displayed results only
- The GIN index must be rebuilt after major data changes. `REINDEX INDEX idx_products_search;`
- Search vector updates via trigger add ~1ms per write. For high-write tables, consider async indexing
- Default `english` config removes common stop words ("the", "a", "is"). Use `simple` config to keep them
