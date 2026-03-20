---
name: vector-db-development
description: >
  Vector database setup and operations — pgvector, Pinecone, ChromaDB,
  Weaviate, Qdrant, and Milvus with TypeScript/Python examples for
  indexing, querying, filtering, and production deployment.
  Triggers: "vector database", "pgvector", "pinecone", "chromadb",
  "weaviate", "qdrant", "milvus", "vector search", "similarity search".
  NOT for: RAG patterns and chunking (use rag-patterns).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Vector Database Development

## pgvector (PostgreSQL Extension)

### Setup

```sql
-- Install extension
CREATE EXTENSION IF NOT EXISTS vector;

-- Create table with vector column
CREATE TABLE documents (
  id SERIAL PRIMARY KEY,
  content TEXT NOT NULL,
  embedding vector(1536),  -- Match your model's dimensions
  metadata JSONB DEFAULT '{}',
  created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create HNSW index (recommended for most cases)
CREATE INDEX ON documents
  USING hnsw (embedding vector_cosine_ops)
  WITH (m = 16, ef_construction = 64);

-- Alternative: IVFFlat index (faster build, slightly less accurate)
-- CREATE INDEX ON documents
--   USING ivfflat (embedding vector_cosine_ops)
--   WITH (lists = 100);
```

### TypeScript with pgvector

```typescript
import pg from "pg";
import pgvector from "pgvector/pg";

const pool = new pg.Pool({ connectionString: process.env.DATABASE_URL });

// Register pgvector type
await pgvector.registerType(pool);

// Insert with embedding
async function insertDocument(
  content: string,
  embedding: number[],
  metadata: Record<string, any> = {}
) {
  await pool.query(
    `INSERT INTO documents (content, embedding, metadata)
     VALUES ($1, $2, $3)`,
    [content, pgvector.toSql(embedding), JSON.stringify(metadata)]
  );
}

// Cosine similarity search
async function search(
  queryEmbedding: number[],
  topK: number = 5,
  filter?: Record<string, any>
) {
  let query = `
    SELECT id, content, metadata,
           1 - (embedding <=> $1) AS score
    FROM documents
  `;
  const params: any[] = [pgvector.toSql(queryEmbedding)];

  if (filter) {
    // JSONB containment operator
    query += ` WHERE metadata @> $2`;
    params.push(JSON.stringify(filter));
  }

  query += ` ORDER BY embedding <=> $1 LIMIT $${params.length + 1}`;
  params.push(topK);

  const result = await pool.query(query, params);
  return result.rows;
}

// Hybrid search: vector + full-text
async function hybridSearch(
  queryEmbedding: number[],
  queryText: string,
  topK: number = 5,
  alpha: number = 0.7
) {
  // Requires a tsvector column and GIN index:
  // ALTER TABLE documents ADD COLUMN tsv tsvector
  //   GENERATED ALWAYS AS (to_tsvector('english', content)) STORED;
  // CREATE INDEX ON documents USING gin(tsv);

  const result = await pool.query(
    `WITH vector_results AS (
       SELECT id, content, metadata,
              1 - (embedding <=> $1) AS vector_score,
              ROW_NUMBER() OVER (ORDER BY embedding <=> $1) AS vector_rank
       FROM documents
       ORDER BY embedding <=> $1
       LIMIT $3 * 3
     ),
     text_results AS (
       SELECT id, content, metadata,
              ts_rank(tsv, plainto_tsquery('english', $2)) AS text_score,
              ROW_NUMBER() OVER (ORDER BY ts_rank(tsv, plainto_tsquery('english', $2)) DESC) AS text_rank
       FROM documents
       WHERE tsv @@ plainto_tsquery('english', $2)
       LIMIT $3 * 3
     )
     SELECT COALESCE(v.id, t.id) AS id,
            COALESCE(v.content, t.content) AS content,
            COALESCE(v.metadata, t.metadata) AS metadata,
            ($4 * COALESCE(1.0 / (60 + v.vector_rank), 0) +
             (1 - $4) * COALESCE(1.0 / (60 + t.text_rank), 0)) AS combined_score
     FROM vector_results v
     FULL OUTER JOIN text_results t ON v.id = t.id
     ORDER BY combined_score DESC
     LIMIT $3`,
    [pgvector.toSql(queryEmbedding), queryText, topK, alpha]
  );

  return result.rows;
}

// Batch upsert
async function upsertBatch(
  records: { id: string; content: string; embedding: number[]; metadata: Record<string, any> }[]
) {
  const client = await pool.connect();
  try {
    await client.query("BEGIN");
    for (const record of records) {
      await client.query(
        `INSERT INTO documents (id, content, embedding, metadata)
         VALUES ($1, $2, $3, $4)
         ON CONFLICT (id) DO UPDATE SET
           content = EXCLUDED.content,
           embedding = EXCLUDED.embedding,
           metadata = EXCLUDED.metadata`,
        [record.id, record.content, pgvector.toSql(record.embedding), JSON.stringify(record.metadata)]
      );
    }
    await client.query("COMMIT");
  } catch (e) {
    await client.query("ROLLBACK");
    throw e;
  } finally {
    client.release();
  }
}
```

### pgvector Distance Operators

| Operator | Distance | Use Case |
|----------|----------|----------|
| `<=>` | Cosine distance | Most embeddings (normalized) |
| `<->` | L2 (Euclidean) | Spatial data, when magnitude matters |
| `<#>` | Inner product (negative) | Pre-normalized vectors, maximum inner product |

## Pinecone

### Setup

```typescript
import { Pinecone } from "@pinecone-database/pinecone";

const pinecone = new Pinecone({
  apiKey: process.env.PINECONE_API_KEY!,
});

// Create index (serverless)
await pinecone.createIndex({
  name: "documents",
  dimension: 1536,
  metric: "cosine",
  spec: {
    serverless: {
      cloud: "aws",
      region: "us-east-1",
    },
  },
});

const index = pinecone.index("documents");
```

### Operations

```typescript
// Upsert vectors
async function upsert(
  records: { id: string; embedding: number[]; metadata: Record<string, any>; text: string }[]
) {
  // Pinecone max batch: 100 vectors per upsert
  const batchSize = 100;
  for (let i = 0; i < records.length; i += batchSize) {
    const batch = records.slice(i, i + batchSize);
    await index.upsert(
      batch.map((r) => ({
        id: r.id,
        values: r.embedding,
        metadata: { ...r.metadata, text: r.text }, // Store text in metadata
      }))
    );
  }
}

// Query with metadata filter
async function query(
  embedding: number[],
  topK: number = 5,
  filter?: Record<string, any>
) {
  const result = await index.query({
    vector: embedding,
    topK,
    includeMetadata: true,
    filter, // Pinecone filter syntax: { category: { $eq: "tech" } }
  });

  return result.matches?.map((m) => ({
    id: m.id,
    score: m.score,
    text: m.metadata?.text as string,
    metadata: m.metadata,
  })) ?? [];
}

// Delete by ID or metadata filter
await index.deleteOne("doc-123");
await index.deleteMany(["doc-1", "doc-2", "doc-3"]);
await index.deleteMany({ filter: { source: { $eq: "old-docs" } } });

// Namespaces for multi-tenant isolation
const ns = index.namespace("tenant-abc");
await ns.upsert([...]);
const results = await ns.query({ vector: [...], topK: 5 });
```

### Pinecone Filter Syntax

```typescript
// Equality
{ category: { $eq: "tech" } }

// Comparison
{ price: { $gt: 10, $lte: 100 } }

// In list
{ status: { $in: ["active", "pending"] } }

// Logical operators
{ $and: [{ category: { $eq: "tech" } }, { year: { $gte: 2024 } }] }
{ $or: [{ source: { $eq: "blog" } }, { source: { $eq: "docs" } }] }
```

## ChromaDB

### Setup

```typescript
import { ChromaClient } from "chromadb";

const chroma = new ChromaClient({
  path: "http://localhost:8000", // or use in-memory
});

// Create collection
const collection = await chroma.createCollection({
  name: "documents",
  metadata: { "hnsw:space": "cosine" }, // cosine, l2, or ip
});
```

### Operations

```typescript
// Add documents (ChromaDB can auto-embed with built-in models)
await collection.add({
  ids: ["doc-1", "doc-2", "doc-3"],
  documents: ["First document text", "Second document text", "Third text"],
  embeddings: [embedding1, embedding2, embedding3], // Or omit for auto-embed
  metadatas: [
    { source: "blog", category: "tech" },
    { source: "docs", category: "api" },
    { source: "blog", category: "tutorial" },
  ],
});

// Query
const results = await collection.query({
  queryEmbeddings: [queryEmbedding],
  nResults: 5,
  where: { category: { $eq: "tech" } },
  whereDocument: { $contains: "search term" }, // Full-text filter
});

// Results shape
// results.ids[0]       — matched IDs
// results.documents[0] — matched documents
// results.distances[0] — distance scores
// results.metadatas[0] — metadata objects

// Update
await collection.update({
  ids: ["doc-1"],
  documents: ["Updated text"],
  embeddings: [newEmbedding],
  metadatas: [{ source: "blog", category: "updated" }],
});

// Delete
await collection.delete({
  ids: ["doc-1"],
  where: { source: { $eq: "old" } },
});
```

### ChromaDB Filter Syntax

```typescript
// Metadata filters
{ field: { $eq: "value" } }
{ field: { $ne: "value" } }
{ field: { $gt: 10 } }
{ field: { $gte: 10 } }
{ field: { $lt: 100 } }
{ field: { $lte: 100 } }
{ field: { $in: ["a", "b"] } }
{ field: { $nin: ["a", "b"] } }
{ $and: [{ field1: "a" }, { field2: "b" }] }
{ $or: [{ field1: "a" }, { field2: "b" }] }

// Document content filters (whereDocument)
{ $contains: "search term" }
{ $not_contains: "exclude this" }
```

## Weaviate

### Setup

```typescript
import weaviate, { WeaviateClient } from "weaviate-client";

const client: WeaviateClient = await weaviate.connectToLocal();
// or: await weaviate.connectToWeaviateCloud(url, { authCredentials: new weaviate.ApiKey(key) });

// Create collection (class)
await client.collections.create({
  name: "Document",
  vectorizers: [
    weaviate.configure.vectorizer.none({ name: "custom" }), // Bring your own embeddings
  ],
  properties: [
    { name: "content", dataType: "text" },
    { name: "source", dataType: "text" },
    { name: "category", dataType: "text" },
  ],
});
```

### Operations

```typescript
const collection = client.collections.get("Document");

// Insert
await collection.data.insert({
  properties: {
    content: "Document text here",
    source: "blog",
    category: "tech",
  },
  vectors: { custom: embedding },
});

// Batch insert
const objects = documents.map((doc) => ({
  properties: { content: doc.text, source: doc.source },
  vectors: { custom: doc.embedding },
}));
await collection.data.insertMany(objects);

// Vector search with filters
const result = await collection.query.nearVector(queryEmbedding, {
  limit: 5,
  returnProperties: ["content", "source", "category"],
  returnMetadata: ["distance"],
  filters: weaviate.filter
    .byProperty("category")
    .equal("tech"),
});

// Hybrid search (vector + BM25)
const hybridResult = await collection.query.hybrid("search query", {
  vector: queryEmbedding,
  limit: 5,
  alpha: 0.7, // 1.0 = pure vector, 0.0 = pure keyword
  returnProperties: ["content", "source"],
});

// BM25 keyword search
const bm25Result = await collection.query.bm25("search query", {
  limit: 5,
  returnProperties: ["content", "source"],
});
```

## Qdrant

### Setup

```typescript
import { QdrantClient } from "@qdrant/js-client-rest";

const qdrant = new QdrantClient({
  url: "http://localhost:6333",
  // or: apiKey: process.env.QDRANT_API_KEY
});

// Create collection
await qdrant.createCollection("documents", {
  vectors: {
    size: 1536,
    distance: "Cosine", // Cosine, Euclid, Dot, Manhattan
  },
  optimizers_config: {
    default_segment_number: 2,
  },
});

// Create payload index for filtering
await qdrant.createPayloadIndex("documents", {
  field_name: "category",
  field_schema: "keyword",
});
```

### Operations

```typescript
// Upsert points
await qdrant.upsert("documents", {
  points: documents.map((doc, i) => ({
    id: i, // or UUID string
    vector: doc.embedding,
    payload: {
      content: doc.text,
      source: doc.source,
      category: doc.category,
      created_at: new Date().toISOString(),
    },
  })),
});

// Search with filter
const results = await qdrant.search("documents", {
  vector: queryEmbedding,
  limit: 5,
  filter: {
    must: [
      { key: "category", match: { value: "tech" } },
    ],
    must_not: [
      { key: "source", match: { value: "deprecated" } },
    ],
  },
  with_payload: true,
  score_threshold: 0.7,
});

// Scroll (paginated retrieval)
const scrollResult = await qdrant.scroll("documents", {
  filter: { must: [{ key: "source", match: { value: "blog" } }] },
  limit: 100,
  with_payload: true,
  with_vector: false,
});

// Delete by filter
await qdrant.delete("documents", {
  filter: {
    must: [{ key: "source", match: { value: "old-docs" } }],
  },
});
```

## Database Selection Guide

| Factor | pgvector | Pinecone | ChromaDB | Weaviate | Qdrant |
|--------|----------|----------|----------|----------|--------|
| **Best for** | Existing Postgres | Managed serverless | Prototyping | Multi-modal | High performance |
| **Self-hosted** | Yes (Postgres) | No | Yes | Yes | Yes |
| **Managed** | Supabase, Neon | Yes (only) | No | Weaviate Cloud | Qdrant Cloud |
| **Max vectors** | Billions (with partitioning) | Billions | Millions | Billions | Billions |
| **Hybrid search** | Manual (tsvector) | Sparse vectors | whereDocument | Built-in BM25 | Sparse vectors |
| **Filtering** | SQL + JSONB | Proprietary | Proprietary | GraphQL-like | Proprietary |
| **Multi-tenancy** | Schemas/RLS | Namespaces | Collections | Multi-tenancy | Collections |
| **Cost (small)** | Free (existing DB) | Free tier | Free | Free tier | Free tier |
| **Latency (p99)** | 10-50ms | 50-100ms | 5-20ms (local) | 20-50ms | 5-20ms |

### Decision Tree

1. **Already using PostgreSQL?** → pgvector. No new infrastructure.
2. **Want zero ops?** → Pinecone. Fully managed, scales automatically.
3. **Prototyping / local dev?** → ChromaDB. Runs in-process, zero setup.
4. **Need hybrid search built-in?** → Weaviate. Best native BM25 + vector.
5. **Performance-critical, self-hosted?** → Qdrant. Written in Rust, fastest.
6. **Multi-modal (text + images)?** → Weaviate or Qdrant. Named vector support.

## Production Considerations

### Indexing Strategy

```
Small dataset (< 100K vectors):
  → Flat/brute-force. Exact results, fast enough.

Medium dataset (100K - 10M vectors):
  → HNSW index. Best accuracy/speed tradeoff.
  → Tune: m=16 (connections), ef_construction=64 (build quality)
  → Query: ef_search=40 (higher = more accurate, slower)

Large dataset (> 10M vectors):
  → IVF + PQ (Product Quantization) for memory efficiency
  → Or use managed service (Pinecone, Qdrant Cloud)
  → Consider sharding by metadata (e.g., per-tenant collections)
```

### Monitoring

```typescript
// Track these metrics in production
interface VectorDBMetrics {
  queryLatencyP50: number;   // Target: < 50ms
  queryLatencyP99: number;   // Target: < 200ms
  indexSize: number;         // Total vectors
  avgRecallAt10: number;     // Target: > 0.95
  embeddingQueueDepth: number; // Pending embeddings
  staleDocumentCount: number;  // Documents needing re-embed
}
```

## Gotchas

1. **pgvector HNSW vs IVFFlat** — HNSW is almost always better. IVFFlat requires `SET ivfflat.probes = 10` at query time (default 1 is terrible recall). HNSW has no query-time knobs to forget.

2. **Pinecone upsert is NOT atomic** — If you upsert 100 vectors and it fails midway, some vectors are updated and others aren't. Use idempotent IDs and retry the full batch.

3. **ChromaDB persistence** — In-memory by default. Pass `path="/data/chroma"` to persist to disk. Without this, everything is lost on restart.

4. **Weaviate auto-schema** — Creates properties on first insert if you don't define them. This is convenient but leads to inconsistent schemas. Always define your schema explicitly.

5. **Dimension mismatch** — If you insert a 1536-dim vector into a 768-dim collection, most DBs give a cryptic error. Always verify your embedding model's output dimensions match the collection config.

6. **Metadata size limits** — Pinecone: 40KB per vector metadata. ChromaDB: no hard limit but performance degrades. Store large text in a regular database and reference by ID.
