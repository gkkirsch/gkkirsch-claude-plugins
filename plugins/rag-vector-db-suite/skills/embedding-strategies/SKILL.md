---
name: embedding-strategies
description: >
  Text chunking, embedding model selection, and retrieval optimization for RAG.
  Use when choosing embedding models, designing chunking strategies,
  implementing hybrid search, or optimizing retrieval quality.
  Triggers: "embedding strategy", "text chunking", "embedding model",
  "hybrid search", "reranking", "chunk size", "semantic search optimization",
  "BM25 hybrid", "embedding dimensions", "retrieval quality".
  NOT for: RAG pipeline architecture (see rag-patterns), vector database setup (see vector-db-development).
version: 1.0.0
allowed-tools: Read, Grep, Glob, Bash
---

# Embedding Strategies

## Text Chunking Approaches

```typescript
// lib/chunkers.ts — Multiple chunking strategies

interface Chunk {
  id: string;
  text: string;
  metadata: {
    source: string;
    chunkIndex: number;
    startChar: number;
    endChar: number;
    tokenCount: number;
    strategy: string;
  };
}

// Strategy 1: Fixed-size with overlap
function fixedSizeChunk(
  text: string,
  options: { chunkSize: number; overlap: number; source: string },
): Chunk[] {
  const { chunkSize, overlap, source } = options;
  const chunks: Chunk[] = [];
  let start = 0;

  while (start < text.length) {
    const end = Math.min(start + chunkSize, text.length);
    const chunkText = text.slice(start, end);

    chunks.push({
      id: `${source}-${chunks.length}`,
      text: chunkText,
      metadata: {
        source,
        chunkIndex: chunks.length,
        startChar: start,
        endChar: end,
        tokenCount: estimateTokens(chunkText),
        strategy: 'fixed-size',
      },
    });

    start += chunkSize - overlap;
  }

  return chunks;
}

// Strategy 2: Recursive text splitting (best general-purpose)
function recursiveChunk(
  text: string,
  options: {
    maxChunkSize: number;
    minChunkSize: number;
    overlap: number;
    source: string;
    separators?: string[];
  },
): Chunk[] {
  const separators = options.separators ?? ['\n\n', '\n', '. ', '? ', '! ', '; ', ', ', ' '];
  const chunks: Chunk[] = [];

  function split(text: string, separatorIndex: number): string[] {
    if (text.length <= options.maxChunkSize) return [text];
    if (separatorIndex >= separators.length) {
      // No more separators — force split at maxChunkSize
      return fixedSizeChunk(text, {
        chunkSize: options.maxChunkSize,
        overlap: options.overlap,
        source: options.source,
      }).map(c => c.text);
    }

    const separator = separators[separatorIndex];
    const parts = text.split(separator);
    const result: string[] = [];
    let current = '';

    for (const part of parts) {
      const candidate = current ? current + separator + part : part;
      if (candidate.length > options.maxChunkSize && current) {
        result.push(current);
        current = part;
      } else {
        current = candidate;
      }
    }
    if (current) result.push(current);

    // Recursively split any chunks that are still too large
    return result.flatMap(chunk =>
      chunk.length > options.maxChunkSize
        ? split(chunk, separatorIndex + 1)
        : [chunk]
    );
  }

  const splitTexts = split(text, 0);
  let charOffset = 0;

  for (const chunkText of splitTexts) {
    if (chunkText.length >= options.minChunkSize) {
      chunks.push({
        id: `${options.source}-${chunks.length}`,
        text: chunkText.trim(),
        metadata: {
          source: options.source,
          chunkIndex: chunks.length,
          startChar: charOffset,
          endChar: charOffset + chunkText.length,
          tokenCount: estimateTokens(chunkText),
          strategy: 'recursive',
        },
      });
    }
    charOffset += chunkText.length;
  }

  return chunks;
}

// Strategy 3: Semantic chunking (by topic/meaning)
async function semanticChunk(
  text: string,
  options: {
    maxChunkSize: number;
    source: string;
    embedder: (text: string) => Promise<number[]>;
    similarityThreshold: number;
  },
): Promise<Chunk[]> {
  // Split into sentences
  const sentences = text.match(/[^.!?]+[.!?]+/g) ?? [text];

  // Get embeddings for each sentence
  const embeddings = await Promise.all(
    sentences.map(s => options.embedder(s.trim()))
  );

  // Group sentences by semantic similarity
  const groups: string[][] = [[sentences[0]]];
  for (let i = 1; i < sentences.length; i++) {
    const similarity = cosineSimilarity(embeddings[i], embeddings[i - 1]);
    const currentGroup = groups[groups.length - 1];
    const currentText = currentGroup.join(' ');

    if (
      similarity >= options.similarityThreshold &&
      currentText.length + sentences[i].length <= options.maxChunkSize
    ) {
      currentGroup.push(sentences[i]);
    } else {
      groups.push([sentences[i]]);
    }
  }

  return groups.map((group, index) => {
    const chunkText = group.join(' ').trim();
    return {
      id: `${options.source}-${index}`,
      text: chunkText,
      metadata: {
        source: options.source,
        chunkIndex: index,
        startChar: 0,
        endChar: chunkText.length,
        tokenCount: estimateTokens(chunkText),
        strategy: 'semantic',
      },
    };
  });
}

function estimateTokens(text: string): number {
  return Math.ceil(text.length / 4); // Rough estimate: ~4 chars per token
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

## Embedding Model Selection

```typescript
// lib/embedding-models.ts — Model comparison and configuration

interface EmbeddingModel {
  name: string;
  dimensions: number;
  maxTokens: number;
  costPer1MTokens: number; // USD
  latencyMs: number;       // Average per request
  quality: 'high' | 'medium' | 'low';
  bestFor: string[];
}

const MODELS: EmbeddingModel[] = [
  {
    name: 'text-embedding-3-large',
    dimensions: 3072, // Can reduce to 256-1536 with Matryoshka
    maxTokens: 8191,
    costPer1MTokens: 0.13,
    latencyMs: 150,
    quality: 'high',
    bestFor: ['multilingual', 'complex queries', 'high-stakes retrieval'],
  },
  {
    name: 'text-embedding-3-small',
    dimensions: 1536, // Can reduce to 256-512 with Matryoshka
    maxTokens: 8191,
    costPer1MTokens: 0.02,
    latencyMs: 80,
    quality: 'medium',
    bestFor: ['cost-sensitive', 'high-volume', 'simple documents'],
  },
  {
    name: 'voyage-3',
    dimensions: 1024,
    maxTokens: 32000,
    costPer1MTokens: 0.06,
    latencyMs: 120,
    quality: 'high',
    bestFor: ['code retrieval', 'long documents', 'technical content'],
  },
  {
    name: 'all-MiniLM-L6-v2',
    dimensions: 384,
    maxTokens: 512,
    costPer1MTokens: 0, // Local/free
    latencyMs: 20,
    quality: 'medium',
    bestFor: ['on-device', 'privacy-sensitive', 'low-latency'],
  },
];

// Matryoshka embedding: truncate to lower dimensions for speed
function truncateEmbedding(embedding: number[], targetDim: number): number[] {
  const truncated = embedding.slice(0, targetDim);
  // Re-normalize
  const norm = Math.sqrt(truncated.reduce((sum, v) => sum + v * v, 0));
  return truncated.map(v => v / norm);
}
```

## Hybrid Search (Vector + Keyword)

```typescript
// lib/hybrid-search.ts — Combine semantic and keyword search

interface SearchResult {
  id: string;
  text: string;
  score: number;
  source: 'vector' | 'keyword' | 'hybrid';
  metadata: Record<string, unknown>;
}

async function hybridSearch(
  query: string,
  options: {
    vectorSearch: (query: string, topK: number) => Promise<SearchResult[]>;
    keywordSearch: (query: string, topK: number) => Promise<SearchResult[]>;
    topK: number;
    alpha: number; // 0 = keyword only, 1 = vector only, 0.7 = balanced
    reranker?: (query: string, results: SearchResult[]) => Promise<SearchResult[]>;
  },
): Promise<SearchResult[]> {
  const { vectorSearch, keywordSearch, topK, alpha, reranker } = options;

  // Fetch from both sources (2x topK for diversity before merging)
  const [vectorResults, keywordResults] = await Promise.all([
    vectorSearch(query, topK * 2),
    keywordSearch(query, topK * 2),
  ]);

  // Reciprocal Rank Fusion (RRF) — merge and rerank
  const k = 60; // RRF constant
  const scores = new Map<string, { score: number; result: SearchResult }>();

  vectorResults.forEach((result, rank) => {
    const rrf = alpha / (k + rank + 1);
    const existing = scores.get(result.id);
    scores.set(result.id, {
      score: (existing?.score ?? 0) + rrf,
      result: { ...result, source: 'hybrid' },
    });
  });

  keywordResults.forEach((result, rank) => {
    const rrf = (1 - alpha) / (k + rank + 1);
    const existing = scores.get(result.id);
    scores.set(result.id, {
      score: (existing?.score ?? 0) + rrf,
      result: existing?.result ?? { ...result, source: 'hybrid' },
    });
  });

  // Sort by combined RRF score
  let merged = [...scores.values()]
    .sort((a, b) => b.score - a.score)
    .slice(0, topK * 2) // Over-fetch for reranker
    .map(s => ({ ...s.result, score: s.score }));

  // Optional: rerank with a cross-encoder for higher quality
  if (reranker) {
    merged = await reranker(query, merged);
  }

  return merged.slice(0, topK);
}

// BM25 keyword search implementation
class BM25 {
  private k1 = 1.5;
  private b = 0.75;
  private avgDocLength = 0;
  private docCount = 0;
  private df = new Map<string, number>(); // Document frequency
  private docs: { id: string; terms: string[]; text: string; metadata: Record<string, unknown> }[] = [];

  addDocuments(docs: { id: string; text: string; metadata: Record<string, unknown> }[]): void {
    for (const doc of docs) {
      const terms = this.tokenize(doc.text);
      this.docs.push({ ...doc, terms });
      const uniqueTerms = new Set(terms);
      for (const term of uniqueTerms) {
        this.df.set(term, (this.df.get(term) ?? 0) + 1);
      }
    }
    this.docCount = this.docs.length;
    this.avgDocLength = this.docs.reduce((sum, d) => sum + d.terms.length, 0) / this.docCount;
  }

  search(query: string, topK: number): SearchResult[] {
    const queryTerms = this.tokenize(query);
    const scores: { id: string; score: number; text: string; metadata: Record<string, unknown> }[] = [];

    for (const doc of this.docs) {
      let score = 0;
      const termFreq = new Map<string, number>();
      for (const term of doc.terms) {
        termFreq.set(term, (termFreq.get(term) ?? 0) + 1);
      }

      for (const term of queryTerms) {
        const tf = termFreq.get(term) ?? 0;
        const df = this.df.get(term) ?? 0;
        if (tf === 0 || df === 0) continue;

        const idf = Math.log((this.docCount - df + 0.5) / (df + 0.5) + 1);
        const tfNorm = (tf * (this.k1 + 1)) / (tf + this.k1 * (1 - this.b + this.b * doc.terms.length / this.avgDocLength));
        score += idf * tfNorm;
      }

      if (score > 0) scores.push({ id: doc.id, score, text: doc.text, metadata: doc.metadata });
    }

    return scores
      .sort((a, b) => b.score - a.score)
      .slice(0, topK)
      .map(s => ({ ...s, source: 'keyword' as const }));
  }

  private tokenize(text: string): string[] {
    return text.toLowerCase().replace(/[^\w\s]/g, '').split(/\s+/).filter(t => t.length > 1);
  }
}
```

## Gotchas

1. **Chunk size vs retrieval quality tradeoff** -- Small chunks (100-200 tokens) are precise but lose context. Large chunks (1000+ tokens) preserve context but dilute relevance signals. Start with 500-800 tokens with 100-token overlap for most use cases. Test with your actual queries and measure recall.

2. **Embedding the query differently from documents** -- Some models (e.g., E5, BGE) require different prefixes for queries vs documents: "query: how to deploy" vs "passage: deployment guide covers...". Using the wrong prefix degrades retrieval quality by 10-20%. Check the model's documentation for required prefixes.

3. **Matryoshka truncation needs renormalization** -- When truncating embeddings to lower dimensions (e.g., 3072 → 512), the truncated vector is no longer unit-length. Renormalize after truncation: divide each component by the L2 norm. Without renormalization, cosine similarity produces incorrect rankings.

4. **BM25 tokenization mismatch** -- If your BM25 tokenizer strips punctuation differently from your query preprocessing, exact keyword matches fail. "Node.js" tokenized as "node" and "js" won't match a query for "nodejs". Normalize consistently: lowercase, strip punctuation, optionally stem.

5. **Hybrid search alpha needs tuning** -- The alpha parameter (vector vs keyword weight) is not one-size-fits-all. Technical queries with specific terms benefit from higher keyword weight (alpha=0.3-0.5). Conceptual/natural language queries benefit from higher vector weight (alpha=0.7-0.9). Tune alpha per query type.

6. **Re-indexing after model change** -- Switching embedding models invalidates ALL existing vectors. You cannot mix vectors from different models in the same index — the embeddings live in different vector spaces. Plan for full re-indexing time when upgrading models, especially with millions of documents.
