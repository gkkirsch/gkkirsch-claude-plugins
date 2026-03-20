---
name: rag-patterns
description: >
  RAG implementation patterns — document chunking, embedding generation,
  retrieval strategies, hybrid search, reranking, evaluation, and
  production pipeline architecture with TypeScript and Python examples.
  Triggers: "RAG", "retrieval augmented generation", "chunking strategy",
  "embedding", "semantic search", "hybrid search", "reranking".
  NOT for: vector database setup (use vector-db-development), fine-tuning.
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# RAG Implementation Patterns

## Document Chunking

### Recursive Character Splitting (Most Common)

```typescript
interface ChunkOptions {
  chunkSize: number;
  chunkOverlap: number;
  separators: string[];
}

function recursiveChunk(text: string, opts: ChunkOptions): string[] {
  const { chunkSize, chunkOverlap, separators } = opts;
  const chunks: string[] = [];

  function split(text: string, sepIdx: number): string[] {
    if (text.length <= chunkSize) return [text];
    if (sepIdx >= separators.length) {
      // Hard split at chunkSize
      const result: string[] = [];
      for (let i = 0; i < text.length; i += chunkSize - chunkOverlap) {
        result.push(text.slice(i, i + chunkSize));
      }
      return result;
    }

    const sep = separators[sepIdx];
    const parts = text.split(sep).filter(Boolean);
    const merged: string[] = [];
    let current = "";

    for (const part of parts) {
      const candidate = current ? current + sep + part : part;
      if (candidate.length <= chunkSize) {
        current = candidate;
      } else {
        if (current) merged.push(current);
        if (part.length > chunkSize) {
          merged.push(...split(part, sepIdx + 1));
          current = "";
        } else {
          current = part;
        }
      }
    }
    if (current) merged.push(current);
    return merged;
  }

  return split(text, 0);
}

// Usage — tune separators to your content
const chunks = recursiveChunk(document, {
  chunkSize: 512,
  chunkOverlap: 64,
  separators: ["\n\n", "\n", ". ", " "],
});
```

### Semantic Chunking (Higher Quality)

```typescript
import { embedBatch } from "./embeddings";

async function semanticChunk(
  sentences: string[],
  threshold: number = 0.3
): Promise<string[]> {
  const embeddings = await embedBatch(sentences);
  const chunks: string[] = [];
  let current: string[] = [sentences[0]];

  for (let i = 1; i < sentences.length; i++) {
    const similarity = cosineSimilarity(embeddings[i - 1], embeddings[i]);

    if (similarity < threshold) {
      // Low similarity = topic shift = new chunk
      chunks.push(current.join(" "));
      current = [sentences[i]];
    } else {
      current.push(sentences[i]);
    }
  }
  if (current.length) chunks.push(current.join(" "));

  return chunks;
}

function cosineSimilarity(a: number[], b: number[]): number {
  let dot = 0, magA = 0, magB = 0;
  for (let i = 0; i < a.length; i++) {
    dot += a[i] * b[i];
    magA += a[i] * a[i];
    magB += b[i] * b[i];
  }
  return dot / (Math.sqrt(magA) * Math.sqrt(magB));
}
```

### Chunk Size Guidelines

| Content Type | Chunk Size | Overlap | Why |
|-------------|-----------|---------|-----|
| Technical docs | 512-1024 | 64-128 | Preserves code blocks and explanations |
| Legal/contracts | 256-512 | 128 | Dense, every sentence matters |
| Chat/email | 256-512 | 32 | Short messages, less context needed |
| Books/articles | 1024-2048 | 128-256 | Longer narrative flow |
| Code files | Per function | 0 | Use AST-based splitting, not character |
| FAQ/Q&A | Per question | 0 | Natural boundaries already exist |

## Embedding Generation

### OpenAI Embeddings

```typescript
import OpenAI from "openai";

const openai = new OpenAI();

async function embed(texts: string[]): Promise<number[][]> {
  // Batch in groups of 2048 (API limit)
  const batchSize = 2048;
  const allEmbeddings: number[][] = [];

  for (let i = 0; i < texts.length; i += batchSize) {
    const batch = texts.slice(i, i + batchSize);
    const response = await openai.embeddings.create({
      model: "text-embedding-3-small", // 1536 dims, $0.02/1M tokens
      input: batch,
      dimensions: 512, // Optional dimensionality reduction
    });

    allEmbeddings.push(
      ...response.data.map((d) => d.embedding)
    );

    // Rate limit: 3000 RPM for embedding endpoints
    if (i + batchSize < texts.length) {
      await new Promise((r) => setTimeout(r, 100));
    }
  }

  return allEmbeddings;
}
```

### Local Embeddings (Ollama)

```typescript
async function embedLocal(texts: string[]): Promise<number[][]> {
  const results: number[][] = [];

  for (const text of texts) {
    const res = await fetch("http://localhost:11434/api/embeddings", {
      method: "POST",
      body: JSON.stringify({
        model: "nomic-embed-text", // 768 dims, fast
        prompt: text,
      }),
    });
    const data = await res.json();
    results.push(data.embedding);
  }

  return results;
}
```

### Embedding Best Practices

```typescript
// 1. Always store source text alongside embeddings
interface ChunkRecord {
  id: string;
  text: string;           // Original text — NEVER discard
  embedding: number[];
  metadata: {
    source: string;       // File path, URL, or document ID
    chunkIndex: number;
    totalChunks: number;
    createdAt: string;
    documentTitle?: string;
    section?: string;
  };
}

// 2. Prefix queries vs documents (some models need this)
// text-embedding-3-small does NOT need prefixes
// nomic-embed-text DOES: "search_query: " and "search_document: "
function prefixForModel(text: string, type: "query" | "document", model: string): string {
  if (model === "nomic-embed-text") {
    return type === "query" ? `search_query: ${text}` : `search_document: ${text}`;
  }
  return text; // Most models don't need prefixes
}

// 3. Normalize embeddings for cosine similarity
function normalize(vec: number[]): number[] {
  const magnitude = Math.sqrt(vec.reduce((sum, v) => sum + v * v, 0));
  return vec.map((v) => v / magnitude);
}
```

## Retrieval Strategies

### Basic Vector Search

```typescript
async function retrieve(
  query: string,
  topK: number = 5,
  scoreThreshold: number = 0.7
): Promise<ChunkRecord[]> {
  const queryEmbedding = await embed([query]);

  const results = await vectorDb.search({
    vector: queryEmbedding[0],
    topK: topK * 2, // Fetch extra, filter by threshold
    includeMetadata: true,
  });

  return results
    .filter((r) => r.score >= scoreThreshold)
    .slice(0, topK);
}
```

### Hybrid Search (Vector + Keyword)

```typescript
async function hybridSearch(
  query: string,
  topK: number = 5,
  alpha: number = 0.7 // 1.0 = pure vector, 0.0 = pure keyword
): Promise<ChunkRecord[]> {
  const [vectorResults, keywordResults] = await Promise.all([
    vectorSearch(query, topK * 2),
    keywordSearch(query, topK * 2), // BM25 or full-text search
  ]);

  // Reciprocal Rank Fusion (RRF)
  const k = 60; // RRF constant
  const scores = new Map<string, number>();

  vectorResults.forEach((r, i) => {
    const rrf = alpha * (1 / (k + i + 1));
    scores.set(r.id, (scores.get(r.id) ?? 0) + rrf);
  });

  keywordResults.forEach((r, i) => {
    const rrf = (1 - alpha) * (1 / (k + i + 1));
    scores.set(r.id, (scores.get(r.id) ?? 0) + rrf);
  });

  // Merge and sort by combined score
  const allIds = new Set([
    ...vectorResults.map((r) => r.id),
    ...keywordResults.map((r) => r.id),
  ]);
  const allResults = [...allIds]
    .map((id) => ({
      id,
      score: scores.get(id) ?? 0,
      ...(vectorResults.find((r) => r.id === id) ??
        keywordResults.find((r) => r.id === id))!,
    }))
    .sort((a, b) => b.score - a.score);

  return allResults.slice(0, topK);
}
```

### Multi-Query Retrieval

```typescript
async function multiQueryRetrieve(
  query: string,
  topK: number = 5
): Promise<ChunkRecord[]> {
  // Generate query variations with an LLM
  const variations = await generateQueryVariations(query, 3);
  // e.g., "How do I cache API responses?" =>
  //   ["caching strategies for REST APIs",
  //    "HTTP response caching implementation",
  //    "API cache layer patterns"]

  const allQueries = [query, ...variations];
  const resultSets = await Promise.all(
    allQueries.map((q) => vectorSearch(q, topK))
  );

  // Deduplicate by chunk ID, keep highest score
  const best = new Map<string, ChunkRecord>();
  for (const results of resultSets) {
    for (const r of results) {
      const existing = best.get(r.id);
      if (!existing || r.score > existing.score) {
        best.set(r.id, r);
      }
    }
  }

  return [...best.values()]
    .sort((a, b) => b.score - a.score)
    .slice(0, topK);
}
```

## Reranking

### Cross-Encoder Reranking (Cohere)

```typescript
import { CohereClient } from "cohere-ai";

const cohere = new CohereClient({ token: process.env.COHERE_API_KEY });

async function rerankResults(
  query: string,
  documents: ChunkRecord[],
  topN: number = 5
): Promise<ChunkRecord[]> {
  const response = await cohere.rerank({
    query,
    documents: documents.map((d) => d.text),
    topN,
    model: "rerank-english-v3.0",
  });

  return response.results.map((r) => ({
    ...documents[r.index],
    score: r.relevanceScore,
  }));
}
```

### When to Rerank

| Scenario | Rerank? | Why |
|----------|---------|-----|
| < 10 results from vector search | No | Not enough to benefit |
| Hybrid search with many results | Yes | Reconciles two ranking signals |
| High-stakes (legal, medical) | Yes | Precision matters more than latency |
| Chatbot with fast response needs | Maybe | Adds 200-500ms latency |
| Multi-query retrieval | Yes | Multiple result sets need unified ranking |

## RAG Prompt Construction

### Context Window Management

```typescript
function buildRAGPrompt(
  query: string,
  chunks: ChunkRecord[],
  maxContextTokens: number = 4000
): string {
  // Estimate tokens (rough: 1 token ~= 4 chars)
  let totalTokens = 0;
  const includedChunks: ChunkRecord[] = [];

  for (const chunk of chunks) {
    const chunkTokens = Math.ceil(chunk.text.length / 4);
    if (totalTokens + chunkTokens > maxContextTokens) break;
    includedChunks.push(chunk);
    totalTokens += chunkTokens;
  }

  const context = includedChunks
    .map((c, i) => `[Source ${i + 1}: ${c.metadata.source}]\n${c.text}`)
    .join("\n\n---\n\n");

  return `Answer the question based on the provided context. If the context doesn't contain enough information, say so clearly.

Context:
${context}

Question: ${query}

Answer:`;
}
```

### Citation Pattern

```typescript
const citationPrompt = `Answer the question using ONLY the provided sources. Cite sources inline using [1], [2], etc.

If the sources don't contain the answer, say "I don't have enough information to answer this."

Sources:
${chunks.map((c, i) => `[${i + 1}] ${c.text}\n(Source: ${c.metadata.source})`).join("\n\n")}

Question: ${query}

Provide your answer with inline citations:`;
```

## Evaluation

### Retrieval Metrics

```typescript
interface EvalResult {
  precision: number;   // Relevant retrieved / Total retrieved
  recall: number;      // Relevant retrieved / Total relevant
  mrr: number;         // Mean Reciprocal Rank
  ndcg: number;        // Normalized Discounted Cumulative Gain
}

function evaluateRetrieval(
  retrieved: string[],     // Retrieved chunk IDs
  relevant: Set<string>,   // Ground truth relevant IDs
): EvalResult {
  let relevantFound = 0;
  let reciprocalRank = 0;
  let dcg = 0;

  retrieved.forEach((id, i) => {
    const isRelevant = relevant.has(id) ? 1 : 0;
    relevantFound += isRelevant;

    if (isRelevant && reciprocalRank === 0) {
      reciprocalRank = 1 / (i + 1);
    }

    dcg += isRelevant / Math.log2(i + 2);
  });

  // Ideal DCG
  const idealDcg = [...Array(relevant.size)]
    .reduce((sum, _, i) => sum + 1 / Math.log2(i + 2), 0);

  return {
    precision: relevantFound / retrieved.length,
    recall: relevantFound / relevant.size,
    mrr: reciprocalRank,
    ndcg: idealDcg > 0 ? dcg / idealDcg : 0,
  };
}
```

### End-to-End RAG Evaluation

```typescript
// Use an LLM as judge
async function evaluateAnswer(
  question: string,
  generatedAnswer: string,
  groundTruth: string
): Promise<{ score: number; reasoning: string }> {
  const response = await llm.chat({
    messages: [{
      role: "user",
      content: `Rate the generated answer vs the ground truth on a scale of 1-5.

Question: ${question}
Ground Truth: ${groundTruth}
Generated Answer: ${generatedAnswer}

Score (1=wrong, 3=partially correct, 5=fully correct):
Reasoning:`,
    }],
  });

  // Parse score and reasoning from response
  const text = response.content;
  const scoreMatch = text.match(/Score[:\s]*(\d)/);
  return {
    score: scoreMatch ? parseInt(scoreMatch[1]) : 0,
    reasoning: text,
  };
}
```

## Production Pipeline

### Ingestion Pipeline

```typescript
import { Queue } from "bullmq";

const ingestionQueue = new Queue("document-ingestion");

async function ingestDocument(doc: {
  content: string;
  source: string;
  metadata: Record<string, any>;
}) {
  // 1. Extract text (PDF, HTML, markdown, etc.)
  const text = await extractText(doc.content, doc.source);

  // 2. Chunk
  const chunks = recursiveChunk(text, {
    chunkSize: 512,
    chunkOverlap: 64,
    separators: ["\n\n", "\n", ". ", " "],
  });

  // 3. Generate embeddings in batches
  const embeddings = await embed(chunks);

  // 4. Upsert to vector DB with metadata
  const records = chunks.map((text, i) => ({
    id: `${doc.source}-chunk-${i}`,
    text,
    embedding: embeddings[i],
    metadata: {
      ...doc.metadata,
      source: doc.source,
      chunkIndex: i,
      totalChunks: chunks.length,
      ingestedAt: new Date().toISOString(),
    },
  }));

  await vectorDb.upsert(records);

  return { chunksCreated: records.length };
}

// Incremental updates — only re-embed changed documents
async function incrementalIngest(doc: Document) {
  const hash = crypto.createHash("sha256").update(doc.content).digest("hex");
  const existing = await db.query(
    "SELECT content_hash FROM documents WHERE source = $1",
    [doc.source]
  );

  if (existing.rows[0]?.content_hash === hash) {
    return { status: "unchanged" };
  }

  // Delete old chunks
  await vectorDb.deleteByMetadata({ source: doc.source });

  // Re-ingest
  await ingestDocument(doc);
  await db.query(
    "INSERT INTO documents (source, content_hash) VALUES ($1, $2) ON CONFLICT (source) DO UPDATE SET content_hash = $2",
    [doc.source, hash]
  );

  return { status: "updated" };
}
```

### Query Pipeline

```typescript
async function ragQuery(
  query: string,
  options: {
    topK?: number;
    rerank?: boolean;
    hybrid?: boolean;
    filters?: Record<string, any>;
  } = {}
): Promise<{ answer: string; sources: ChunkRecord[] }> {
  const { topK = 5, rerank = true, hybrid = true, filters } = options;

  // 1. Retrieve
  let chunks: ChunkRecord[];
  if (hybrid) {
    chunks = await hybridSearch(query, topK * 3);
  } else {
    chunks = await vectorSearch(query, topK * 3);
  }

  // 2. Filter by metadata if needed
  if (filters) {
    chunks = chunks.filter((c) =>
      Object.entries(filters).every(([k, v]) => c.metadata[k] === v)
    );
  }

  // 3. Rerank
  if (rerank && chunks.length > topK) {
    chunks = await rerankResults(query, chunks, topK);
  } else {
    chunks = chunks.slice(0, topK);
  }

  // 4. Generate answer
  const prompt = buildRAGPrompt(query, chunks);
  const answer = await llm.chat({
    messages: [{ role: "user", content: prompt }],
    temperature: 0.1, // Low temp for factual answers
  });

  return { answer: answer.content, sources: chunks };
}
```

## Common Failure Modes

1. **"Lost in the Middle"** — LLMs pay less attention to context in the middle of long prompts. Put most relevant chunks first and last, less important in the middle.

2. **Chunk boundary splits** — Important information split across two chunks. Fix: increase overlap, or use parent-child chunking (retrieve child, include parent for context).

3. **Embedding drift** — Model updates change embedding space. Fix: version your embedding model, re-embed everything when switching models.

4. **Query-document mismatch** — User queries are short and informal, documents are long and formal. Fix: use HyDE (Hypothetical Document Embeddings) to generate a hypothetical answer, embed that instead of the raw query.

5. **Stale data** — Vector DB has outdated information. Fix: track document hashes, set up incremental re-ingestion pipeline, include `ingestedAt` metadata and filter old results.

6. **Over-retrieval** — Too many chunks dilute the signal. Fix: use score thresholds (not just top-K), rerank before prompting, dynamically adjust K based on score distribution.

## Gotchas

1. **Never discard source text** — Store the original text alongside embeddings. You can't reverse an embedding back to text. If you lose the source, the embedding is useless.

2. **Normalize embeddings** — Cosine similarity on non-normalized vectors gives wrong results. Some DBs auto-normalize (Pinecone does), others don't (pgvector with `<->` uses L2 distance, use `<=>` for cosine).

3. **Batch embedding calls** — Embedding one text at a time is 100x slower than batching. OpenAI supports up to 2048 inputs per call. Always batch.

4. **Different models = incompatible embeddings** — You cannot mix embeddings from different models in the same collection. Pick a model and stick with it, or maintain separate collections.

5. **Chunk overlap prevents information loss** — Without overlap, a sentence split at the boundary loses context on both sides. 10-20% overlap is a good default.

6. **Test with real queries first** — Before building the full pipeline, embed 50 documents and run 20 real queries manually. This reveals chunking and retrieval issues before you invest in infrastructure.
