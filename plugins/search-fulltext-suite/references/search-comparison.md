# Search Technology Comparison

Detailed comparison of search technologies to help choose the right one for your project.

---

## Feature Comparison Matrix

| Feature | PostgreSQL FTS | Meilisearch | Elasticsearch | Typesense | Algolia |
|---------|---------------|-------------|---------------|-----------|---------|
| **Typo tolerance** | No (needs pg_trgm) | Yes (built-in) | Yes (fuzzy) | Yes (built-in) | Yes (built-in) |
| **Autocomplete** | Basic (prefix) | Excellent | Good (edge ngrams) | Excellent | Excellent |
| **Faceted search** | Manual (COUNT/GROUP) | Built-in | Built-in (aggs) | Built-in | Built-in |
| **Geo search** | PostGIS extension | Built-in | Built-in | Built-in | Built-in |
| **Custom ranking** | SQL ORDER BY | Ranking rules | Function score | Ranking override | Custom ranking |
| **Synonyms** | Custom dictionary | Built-in API | Built-in | Built-in | Built-in |
| **Multi-language** | Many configs | 39 languages | ICU + custom | 30+ languages | 70+ languages |
| **Highlighting** | ts_headline | Built-in | Built-in | Built-in | Built-in |
| **Nested objects** | JSONB + FTS | Flat only | Nested type | Flat only | Flat only |
| **Aggregations** | SQL GROUP BY | Facets only | Full aggs DSL | Facets only | Facets only |
| **Real-time indexing** | Immediate | ~200ms | ~1s (refresh) | ~200ms | ~1s |
| **SDK languages** | Any (SQL) | 10+ | 10+ | 10+ | 10+ |

## Performance Benchmarks (Approximate)

| Metric | PostgreSQL FTS | Meilisearch | Elasticsearch |
|--------|---------------|-------------|---------------|
| **100K docs, simple query** | 5-15ms | 1-5ms | 5-10ms |
| **1M docs, simple query** | 20-50ms | 5-15ms | 10-20ms |
| **10M docs, simple query** | 100-500ms | 15-30ms | 15-30ms |
| **Autocomplete latency** | 10-30ms | 1-5ms | 5-15ms |
| **Index speed (batch)** | 5K/sec | 10K/sec | 5-10K/sec |
| **Memory per 1M docs** | ~500MB (shared w/ DB) | 2-4GB | 2-4GB |

*Benchmarks vary significantly with hardware, document size, and query complexity.*

## Cost Comparison (Monthly)

### Self-Hosted

| Scale | PostgreSQL FTS | Meilisearch | Elasticsearch |
|-------|---------------|-------------|---------------|
| **Dev/Hobby** | $0 (existing DB) | $0 (Docker) | $0 (Docker) |
| **Small (100K docs)** | $0 (existing DB) | $5-10 (VPS) | $15-30 (VPS, needs RAM) |
| **Medium (1M docs)** | $0 (existing DB) | $20-40 (2GB RAM VPS) | $50-100 (4GB RAM) |
| **Large (10M docs)** | $0-50 (may need read replica) | $40-80 (4GB RAM) | $200-500 (cluster) |

### Managed/Cloud

| Scale | Meilisearch Cloud | Elastic Cloud | Algolia | Typesense Cloud |
|-------|-------------------|---------------|---------|----------------|
| **Free tier** | 100K docs, 10K searches/mo | 14-day trial | 10K records | 1M docs |
| **Starter** | $30/mo | $95/mo | $50/mo (100K records) | $30/mo |
| **Production** | $60-300/mo | $200-1000/mo | $150-500/mo | $60-200/mo |
| **Enterprise** | Custom | Custom | Custom | Custom |

## Query Syntax Comparison

### Simple Search

```
PostgreSQL:  SELECT * FROM products WHERE search_vector @@ plainto_tsquery('wireless headphones')
Meilisearch: client.index('products').search('wireless headphones')
Elasticsearch: { "query": { "match": { "title": "wireless headphones" } } }
Typesense:   client.collections('products').documents().search({ q: 'wireless headphones', query_by: 'title' })
```

### Filtered Search

```
PostgreSQL:  ... WHERE search_vector @@ query AND category = 'Electronics' AND price < 10000
Meilisearch: .search('query', { filter: 'category = Electronics AND price < 10000' })
Elasticsearch: { "query": { "bool": { "must": [...], "filter": [{ "term": { "category": "Electronics" } }] } } }
Typesense:   .search({ q: 'query', filter_by: 'category:Electronics && price:<10000' })
```

### Autocomplete

```
PostgreSQL:  ... WHERE search_vector @@ to_tsquery('wire:*')
Meilisearch: .search('wire')  // Built-in prefix matching
Elasticsearch: { "query": { "multi_match": { "query": "wire", "type": "bool_prefix" } } }
Typesense:   .search({ q: 'wire', prefix: true })
```

## Operational Complexity

| Aspect | PostgreSQL FTS | Meilisearch | Elasticsearch |
|--------|---------------|-------------|---------------|
| **Setup time** | 0 (already have PG) | 5 min | 30 min |
| **Config required** | Trigger + index | Near-zero | Significant |
| **Monitoring** | pg_stat, standard PG tools | Built-in stats API | Kibana, cluster health |
| **Backup** | pg_dump (standard) | Snapshots or file copy | Snapshot API |
| **Scaling** | Read replicas | Single node (Cloud: managed) | Shards + replicas |
| **Upgrading** | Standard PG upgrade | Binary replace | Rolling upgrade |
| **Learning curve** | SQL knowledge | Very low | Steep (query DSL) |
| **Failure mode** | DB down = search down | Search-only outage | Complex cluster recovery |

## Migration Path

```
Start here:
  PostgreSQL FTS (free, no new infrastructure)
    ↓ When you need typo tolerance or < 50ms latency
  Meilisearch (easy migration, minimal ops)
    ↓ When you need complex aggregations, analytics, or > 10M docs
  Elasticsearch (full power, full complexity)
```

### Postgres → Meilisearch Migration

1. Export data from Postgres as JSON
2. Create Meilisearch index with settings
3. Bulk import documents
4. Update API to query Meilisearch instead of Postgres
5. Keep Postgres FTS as fallback (feature flag)
6. Remove Postgres FTS after validation

### Meilisearch → Elasticsearch Migration

1. Map Meilisearch settings to Elasticsearch mappings
2. Meilisearch filters → Elasticsearch bool queries
3. Meilisearch ranking rules → Elasticsearch function_score
4. Bulk reindex from source database
5. Update API layer
6. Add Kibana for monitoring

## Hosting Recommendations

| Environment | Recommended |
|-------------|-------------|
| **Side project** | PostgreSQL FTS (zero cost, zero ops) |
| **Startup MVP** | Meilisearch Cloud free tier or self-hosted Docker |
| **Growing SaaS** | Meilisearch Cloud or Typesense Cloud |
| **E-commerce** | Algolia (best search UX) or Elasticsearch |
| **Enterprise** | Elasticsearch (most flexible) or Algolia (least ops) |
| **Data analytics** | Elasticsearch + Kibana (aggregations are unmatched) |

## Common Anti-Patterns

1. **Using search as a database** — Search indexes should mirror your DB, not replace it. Don't rely on search for data integrity.

2. **Indexing everything** — Only index fields that are searched, filtered, sorted, or displayed. Large blobs waste memory and slow indexing.

3. **No fallback** — If your search engine goes down, the app shouldn't crash. Degrade gracefully to basic SQL search.

4. **Premature optimization** — Don't start with Elasticsearch for 10K documents. Postgres FTS handles this trivially.

5. **Ignoring search analytics** — Track what users search for. Zero-result queries tell you what content you're missing. Low click-through tells you your ranking is off.

6. **Sync indexing without error handling** — If search indexing fails on a write, should the write also fail? Usually no. Use async indexing with retry.

7. **Not setting index retention** — Old/deleted records stay in the search index forever unless you clean them up. Schedule periodic reconciliation.
