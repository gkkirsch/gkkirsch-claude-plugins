# RAG & Vector DB Cheatsheet

## Chunking Decision

| Content | Strategy | Size | Overlap |
|---------|----------|------|---------|
| Technical docs | Recursive character | 512-1024 | 64-128 |
| Legal/contracts | Sentence-based | 256-512 | 128 |
| Code files | AST/function-based | Per function | 0 |
| FAQ/Q&A | Per question | Natural | 0 |
| Books/articles | Recursive character | 1024-2048 | 128-256 |
| Chat/email | Fixed-size | 256-512 | 32 |

**Rule of thumb**: Start with 512 tokens, 64 overlap. Adjust based on retrieval quality.

## Embedding Models

| Model | Dims | Context | Speed | Cost |
|-------|------|---------|-------|------|
| text-embedding-3-small | 1536 (or custom) | 8191 | Fast | $0.02/1M tokens |
| text-embedding-3-large | 3072 (or custom) | 8191 | Medium | $0.13/1M tokens |
| voyage-3 | 1024 | 32000 | Medium | $0.06/1M tokens |
| nomic-embed-text | 768 | 8192 | Fast | Free (local) |
| bge-large-en-v1.5 | 1024 | 512 | Medium | Free (local) |

## Vector DB Quick Comparison

```
pgvector     → Already on Postgres? Use this. Zero new infra.
Pinecone     → Want zero ops? Serverless, auto-scales. $$
ChromaDB     → Prototyping? Runs in-process. Zero setup.
Weaviate     → Need hybrid search? Best built-in BM25.
Qdrant       → Need speed? Rust-based, fastest self-hosted.
```

## Distance Metrics

```
Cosine    → Most embeddings. Measures angle, ignores magnitude.
Euclidean → Spatial data. Measures absolute distance.
Dot Product → Pre-normalized vectors. Fastest computation.
```

pgvector operators: `<=>` cosine, `<->` L2, `<#>` inner product (negative)

## Retrieval Pipeline

```
Query → Embed → Vector Search → [Rerank] → Format Context → LLM → Answer
                     ↑
              Optional: Hybrid Search (vector + BM25)
              Optional: Multi-Query (LLM generates query variations)
              Optional: Metadata Filters (pre-filter by source, date, etc.)
```

## Hybrid Search (RRF)

```typescript
// Reciprocal Rank Fusion
const k = 60;
score = alpha * (1 / (k + vectorRank)) + (1 - alpha) * (1 / (k + keywordRank));
// alpha = 0.7 → favor vector, 0.3 → favor keyword
```

## Prompt Template

```
Answer based on the provided context. If the context doesn't contain
enough information, say so.

Context:
[Source 1: path/file.md]
chunk text here...

[Source 2: path/other.md]
chunk text here...

Question: {query}
```

## Evaluation Metrics

| Metric | What It Measures | Target |
|--------|-----------------|--------|
| Precision@K | Relevant / Retrieved | > 0.8 |
| Recall@K | Relevant found / Total relevant | > 0.9 |
| MRR | First relevant result rank | > 0.7 |
| NDCG | Ranking quality (position-aware) | > 0.8 |

## Common Failure Modes

```
Problem: Poor retrieval quality
  → Check: chunk size too large? Overlap missing? Wrong embedding model?
  → Fix: Try 512 chunks with 64 overlap. Test with manual queries first.

Problem: Lost in the Middle
  → LLM ignores context in the middle of long prompts
  → Fix: Put best chunks first AND last. Limit to 5-7 chunks.

Problem: Query-document mismatch
  → User queries are casual, documents are formal
  → Fix: HyDE — generate hypothetical answer, embed THAT instead.

Problem: Stale data
  → Vector DB has outdated docs
  → Fix: Track content hashes. Incremental re-ingestion pipeline.

Problem: Over-retrieval
  → Too many chunks dilute signal
  → Fix: Score threshold (not just top-K). Rerank before prompting.

Problem: Embedding drift
  → Switching models breaks existing embeddings
  → Fix: Version your model. Re-embed everything on switch.
```

## Production Checklist

- [ ] Embeddings stored WITH source text (never discard originals)
- [ ] Metadata includes: source, chunkIndex, totalChunks, ingestedAt
- [ ] Batch embedding calls (not one-at-a-time)
- [ ] Score threshold on retrieval (not just top-K)
- [ ] Incremental ingestion (hash-based change detection)
- [ ] Embedding model version tracked
- [ ] Rate limiting on embedding API calls
- [ ] Monitoring: query latency, recall, index size, queue depth
- [ ] Error handling for embedding API failures (retry with backoff)
- [ ] Content deduplication before ingestion

## Index Tuning

```
HNSW (default, best for most cases):
  m = 16           # connections per node (higher = better recall, more memory)
  ef_construction = 64  # build quality (higher = slower build, better index)
  ef_search = 40   # query quality (higher = slower query, better recall)

IVFFlat (pgvector only):
  lists = sqrt(num_vectors)  # number of clusters
  probes = 10      # clusters to search (DEFAULT IS 1 — terrible recall!)
```

## Pinecone Filter Syntax

```json
{ "field": { "$eq": "value" } }
{ "field": { "$gt": 10, "$lte": 100 } }
{ "field": { "$in": ["a", "b"] } }
{ "$and": [{ "f1": { "$eq": "a" } }, { "f2": { "$gte": 5 } }] }
```

## pgvector Hybrid Search

```sql
-- Requires: tsvector column + GIN index + HNSW index
-- RRF fusion in SQL (alpha weighting vector vs text rank)
WITH vector_results AS (
  SELECT id, ROW_NUMBER() OVER (ORDER BY embedding <=> $1) AS vr
  FROM documents ORDER BY embedding <=> $1 LIMIT 30
),
text_results AS (
  SELECT id, ROW_NUMBER() OVER (ORDER BY ts_rank(tsv, q) DESC) AS tr
  FROM documents, plainto_tsquery('english', $2) q
  WHERE tsv @@ q LIMIT 30
)
SELECT COALESCE(v.id, t.id),
  0.7 / (60 + COALESCE(v.vr, 999)) + 0.3 / (60 + COALESCE(t.tr, 999)) AS score
FROM vector_results v FULL JOIN text_results t ON v.id = t.id
ORDER BY score DESC LIMIT 10;
```
