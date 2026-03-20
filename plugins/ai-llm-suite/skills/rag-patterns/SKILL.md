---
name: rag-patterns
description: >
  Retrieval-Augmented Generation patterns for production applications.
  Use when building knowledge bases, document Q&A, semantic search,
  embedding pipelines, or context assembly for LLM applications.
  Triggers: "RAG", "retrieval augmented", "embedding", "vector search",
  "document Q&A", "knowledge base", "semantic search", "chunk".
  NOT for: traditional search engines without LLM, pure vector database operations without AI.
version: 1.0.0
allowed-tools: Read, Grep, Glob, Bash
---

# RAG Patterns

## Document Chunking Strategies

```typescript
interface Chunk {
  id: string;
  text: string;
  metadata: {
    source: string;
    page?: number;
    section?: string;
    chunkIndex: number;
    tokenCount: number;
  };
}

// Strategy 1: Recursive character splitting with overlap
function recursiveChunk(
  text: string,
  maxChunkSize: number = 512,
  overlap: number = 50,
  separators: string[] = ['\n\n', '\n', '. ', ' ']
): string[] {
  if (text.length <= maxChunkSize) return [text];

  const separator = separators.find(sep => text.includes(sep)) ?? '';
  const parts = text.split(separator);

  const chunks: string[] = [];
  let currentChunk = '';

  for (const part of parts) {
    const candidate = currentChunk
      ? currentChunk + separator + part
      : part;

    if (candidate.length > maxChunkSize && currentChunk) {
      chunks.push(currentChunk);
      // Overlap: keep tail of previous chunk
      const overlapText = currentChunk.slice(-overlap);
      currentChunk = overlapText + separator + part;
    } else {
      currentChunk = candidate;
    }
  }
  if (currentChunk) chunks.push(currentChunk);

  return chunks;
}

// Strategy 2: Semantic chunking by headings
function semanticChunk(markdown: string): Chunk[] {
  const sections = markdown.split(/^(#{1,3}\s.+)$/gm);
  const chunks: Chunk[] = [];
  let currentSection = '';
  let currentText = '';

  for (const part of sections) {
    if (part.match(/^#{1,3}\s/)) {
      if (currentText.trim()) {
        chunks.push({
          id: crypto.randomUUID(),
          text: `${currentSection}\n${currentText}`.trim(),
          metadata: {
            source: 'document',
            section: currentSection,
            chunkIndex: chunks.length,
            tokenCount: estimateTokens(currentText),
          },
        });
      }
      currentSection = part.trim();
      currentText = '';
    } else {
      currentText += part;
    }
  }
  if (currentText.trim()) {
    chunks.push({
      id: crypto.randomUUID(),
      text: `${currentSection}\n${currentText}`.trim(),
      metadata: {
        source: 'document',
        section: currentSection,
        chunkIndex: chunks.length,
        tokenCount: estimateTokens(currentText),
      },
    });
  }

  return chunks;
}

function estimateTokens(text: string): number {
  return Math.ceil(text.length / 4);
}
```

## Embedding Pipeline

```typescript
import Anthropic from '@anthropic-ai/sdk';

// Embedding with batching and rate limiting
class EmbeddingPipeline {
  private client: Anthropic;
  private batchSize: number;
  private delayMs: number;

  constructor(client: Anthropic, batchSize = 20, delayMs = 100) {
    this.client = client;
    this.batchSize = batchSize;
    this.delayMs = delayMs;
  }

  async embedBatch(texts: string[]): Promise<number[][]> {
    const embeddings: number[][] = [];

    for (let i = 0; i < texts.length; i += this.batchSize) {
      const batch = texts.slice(i, i + this.batchSize);

      // Use voyage-3 or similar embedding model
      const results = await Promise.all(
        batch.map(text => this.embedSingle(text))
      );

      embeddings.push(...results);

      if (i + this.batchSize < texts.length) {
        await new Promise(r => setTimeout(r, this.delayMs));
      }
    }

    return embeddings;
  }

  private async embedSingle(text: string): Promise<number[]> {
    // Implementation depends on embedding provider
    // Example with a generic embedding API:
    const response = await fetch('https://api.embedding-provider.com/embed', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ input: text, model: 'voyage-3' }),
    });
    const data = await response.json();
    return data.embedding;
  }
}

// Cosine similarity for retrieval
function cosineSimilarity(a: number[], b: number[]): number {
  let dotProduct = 0;
  let normA = 0;
  let normB = 0;

  for (let i = 0; i < a.length; i++) {
    dotProduct += a[i] * b[i];
    normA += a[i] * a[i];
    normB += b[i] * b[i];
  }

  return dotProduct / (Math.sqrt(normA) * Math.sqrt(normB));
}
```

## Vector Store Integration (PostgreSQL + pgvector)

```sql
-- Setup pgvector
CREATE EXTENSION IF NOT EXISTS vector;

-- Document chunks table
CREATE TABLE document_chunks (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  content TEXT NOT NULL,
  embedding vector(1536),  -- dimension matches your model
  metadata JSONB DEFAULT '{}',
  source TEXT NOT NULL,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

-- HNSW index for fast similarity search
CREATE INDEX ON document_chunks
  USING hnsw (embedding vector_cosine_ops)
  WITH (m = 16, ef_construction = 64);

-- Similarity search function
CREATE OR REPLACE FUNCTION search_documents(
  query_embedding vector(1536),
  match_count INT DEFAULT 5,
  similarity_threshold FLOAT DEFAULT 0.7
)
RETURNS TABLE (
  id UUID,
  content TEXT,
  metadata JSONB,
  similarity FLOAT
) AS $$
BEGIN
  RETURN QUERY
  SELECT
    dc.id,
    dc.content,
    dc.metadata,
    1 - (dc.embedding <=> query_embedding) AS similarity
  FROM document_chunks dc
  WHERE 1 - (dc.embedding <=> query_embedding) > similarity_threshold
  ORDER BY dc.embedding <=> query_embedding
  LIMIT match_count;
END;
$$ LANGUAGE plpgsql;
```

```typescript
// TypeScript wrapper
import { Pool } from 'pg';

class VectorStore {
  private pool: Pool;

  constructor(connectionString: string) {
    this.pool = new Pool({ connectionString });
  }

  async upsertChunks(chunks: Array<{
    id: string;
    content: string;
    embedding: number[];
    metadata: Record<string, any>;
    source: string;
  }>): Promise<void> {
    const client = await this.pool.connect();
    try {
      await client.query('BEGIN');
      for (const chunk of chunks) {
        await client.query(
          `INSERT INTO document_chunks (id, content, embedding, metadata, source)
           VALUES ($1, $2, $3::vector, $4, $5)
           ON CONFLICT (id) DO UPDATE SET
             content = EXCLUDED.content,
             embedding = EXCLUDED.embedding,
             metadata = EXCLUDED.metadata`,
          [chunk.id, chunk.content, `[${chunk.embedding.join(',')}]`, chunk.metadata, chunk.source]
        );
      }
      await client.query('COMMIT');
    } catch (e) {
      await client.query('ROLLBACK');
      throw e;
    } finally {
      client.release();
    }
  }

  async search(queryEmbedding: number[], topK = 5, threshold = 0.7) {
    const result = await this.pool.query(
      `SELECT * FROM search_documents($1::vector, $2, $3)`,
      [`[${queryEmbedding.join(',')}]`, topK, threshold]
    );
    return result.rows;
  }
}
```

## Context Assembly

```typescript
// Assemble context from retrieved chunks with token budget
interface RetrievedChunk {
  content: string;
  similarity: number;
  metadata: Record<string, any>;
}

function assembleContext(
  chunks: RetrievedChunk[],
  maxTokens: number = 4000,
  options: {
    includeMetadata?: boolean;
    deduplicateThreshold?: number;
  } = {}
): string {
  const { includeMetadata = true, deduplicateThreshold = 0.95 } = options;

  // Deduplicate near-identical chunks
  const unique = chunks.filter((chunk, i) =>
    !chunks.slice(0, i).some(prev =>
      cosineSimilarityText(prev.content, chunk.content) > deduplicateThreshold
    )
  );

  // Build context within token budget
  let totalTokens = 0;
  const selected: string[] = [];

  for (const chunk of unique) {
    const formatted = includeMetadata
      ? `[Source: ${chunk.metadata.source}, Section: ${chunk.metadata.section ?? 'N/A'}]\n${chunk.content}`
      : chunk.content;

    const tokens = estimateTokens(formatted);
    if (totalTokens + tokens > maxTokens) break;

    selected.push(formatted);
    totalTokens += tokens;
  }

  return selected.join('\n\n---\n\n');
}

// Simple text similarity (Jaccard on words)
function cosineSimilarityText(a: string, b: string): number {
  const wordsA = new Set(a.toLowerCase().split(/\s+/));
  const wordsB = new Set(b.toLowerCase().split(/\s+/));
  const intersection = new Set([...wordsA].filter(w => wordsB.has(w)));
  const union = new Set([...wordsA, ...wordsB]);
  return intersection.size / union.size;
}
```

## Full RAG Pipeline

```typescript
class RAGPipeline {
  private client: Anthropic;
  private vectorStore: VectorStore;
  private embedder: EmbeddingPipeline;

  constructor(client: Anthropic, vectorStore: VectorStore, embedder: EmbeddingPipeline) {
    this.client = client;
    this.vectorStore = vectorStore;
    this.embedder = embedder;
  }

  async query(userQuery: string, options: {
    topK?: number;
    contextTokens?: number;
    systemPrompt?: string;
  } = {}): Promise<{ answer: string; sources: Array<{ content: string; source: string }> }> {
    const { topK = 5, contextTokens = 4000 } = options;

    // 1. Embed query
    const [queryEmbedding] = await this.embedder.embedBatch([userQuery]);

    // 2. Retrieve relevant chunks
    const chunks = await this.vectorStore.search(queryEmbedding, topK);

    // 3. Assemble context
    const context = assembleContext(chunks, contextTokens);

    // 4. Generate answer
    const systemPrompt = options.systemPrompt ?? `You are a helpful assistant that answers questions based on the provided context.
If the context doesn't contain enough information to answer, say so — do not make up information.
Always cite which source document your answer comes from.`;

    const response = await this.client.messages.create({
      model: 'claude-sonnet-4-6',
      max_tokens: 1024,
      system: systemPrompt,
      messages: [{
        role: 'user',
        content: `Context:\n${context}\n\nQuestion: ${userQuery}`
      }],
    });

    const answer = response.content[0].type === 'text' ? response.content[0].text : '';

    return {
      answer,
      sources: chunks.map(c => ({
        content: c.content.substring(0, 200),
        source: c.metadata.source,
      })),
    };
  }

  // Ingest documents
  async ingest(documents: Array<{ content: string; source: string }>): Promise<number> {
    let totalChunks = 0;

    for (const doc of documents) {
      const chunks = semanticChunk(doc.content);
      const texts = chunks.map(c => c.text);
      const embeddings = await this.embedder.embedBatch(texts);

      await this.vectorStore.upsertChunks(
        chunks.map((chunk, i) => ({
          ...chunk,
          embedding: embeddings[i],
          source: doc.source,
        }))
      );

      totalChunks += chunks.length;
    }

    return totalChunks;
  }
}
```

## Gotchas

1. **Chunk size vs retrieval quality** — chunks too small lose context, too large dilute relevance. 256-512 tokens is the sweet spot for most applications. Always include overlap (10-20%) to avoid splitting sentences mid-thought.

2. **Embedding model mismatch** — the query embedding model MUST match the document embedding model. Mixing models (e.g., embedding docs with voyage-3 but querying with text-embedding-3) produces meaningless similarity scores.

3. **Lost in the middle** — LLMs attend less to information in the middle of long contexts. Place the most relevant chunks first AND last in your context assembly. Or use reciprocal rank fusion to re-rank.

4. **Stale embeddings after document updates** — updating a document's text without re-embedding creates retrieval drift. Implement a re-indexing pipeline triggered by document changes, not just initial ingestion.

5. **Hallucination despite context** — the model may generate plausible-sounding answers not grounded in the retrieved context. Add explicit instructions: "Only answer from the provided context. Quote the relevant passage." and implement citation verification.

6. **pgvector index rebuild on large inserts** — HNSW indexes don't auto-update efficiently for bulk inserts. Build the index AFTER bulk loading, not before. Use `CREATE INDEX CONCURRENTLY` in production to avoid table locks.
