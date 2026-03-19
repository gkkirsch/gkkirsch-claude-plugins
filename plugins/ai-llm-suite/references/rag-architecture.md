# RAG Architecture Reference

Comprehensive reference for Retrieval-Augmented Generation systems — chunking strategies, embedding models, vector stores, hybrid search, reranking, and evaluation.

---

## RAG Pipeline Overview

```
┌──────────────┐    ┌──────────────┐    ┌──────────────┐    ┌──────────────┐    ┌──────────────┐
│   Ingest     │ →  │   Index      │ →  │   Retrieve   │ →  │   Augment    │ →  │   Generate   │
│              │    │              │    │              │    │              │    │              │
│ • Load docs  │    │ • Chunk      │    │ • Query      │    │ • Rerank     │    │ • LLM call   │
│ • Parse      │    │ • Embed      │    │ • Vector     │    │ • Compress   │    │ • Stream     │
│ • Clean      │    │ • Store      │    │ • Hybrid     │    │ • Format     │    │ • Cite       │
└──────────────┘    └──────────────┘    └──────────────┘    └──────────────┘    └──────────────┘
```

---

## Chunking Strategies

### Strategy Selection Guide

| Strategy | Best For | Chunk Size | Pros | Cons |
|----------|----------|-----------|------|------|
| Fixed-size | Uniform content | 256-1000 chars | Simple, predictable | Splits mid-thought |
| Recursive | General text | 500-1500 chars | Respects boundaries | May create uneven chunks |
| Semantic | Mixed-topic docs | Varies | Best topic coherence | Expensive (requires embeddings) |
| Document-aware | Structured docs | Varies | Preserves hierarchy | Format-dependent |
| Sentence-window | Q&A systems | 1-3 sentences | Precise retrieval | Small context per chunk |
| Parent-child | Need both precision + context | Small index, large return | Best of both worlds | More complex indexing |

### Chunk Size Impact

```
Small chunks (128-256 tokens):
  ✓ More precise retrieval
  ✓ Less noise in context
  ✗ May lack sufficient context
  ✗ More chunks to search
  → Best for: Factual Q&A, entity lookup

Medium chunks (256-512 tokens):
  ✓ Good balance of precision and context
  ✓ Works well for most use cases
  → Best for: General RAG, customer support

Large chunks (512-1500 tokens):
  ✓ More context per chunk
  ✓ Better for synthesis tasks
  ✗ More noise, less precise
  ✗ Fewer chunks fit in context window
  → Best for: Summarization, analysis, code
```

### Overlap Guidelines

```
Overlap = 10-20% of chunk size (standard)
Overlap = 0% (when chunks have natural boundaries like sections)
Overlap = 25-30% (when content density is high and splitting is risky)

Example: 512-token chunks → 50-100 token overlap
```

### Parent-Child Chunking

```
Index: Small chunks (sentence or paragraph level) → precise retrieval
Return: Parent chunk (section level) → sufficient context

Document
├── Section 1 (parent chunk — returned to LLM)
│   ├── Paragraph 1 (child chunk — indexed for search)
│   ├── Paragraph 2 (child chunk — indexed for search)
│   └── Paragraph 3 (child chunk — indexed for search)
├── Section 2 (parent chunk)
│   ├── Paragraph 4 (child chunk)
│   └── Paragraph 5 (child chunk)
```

---

## Embedding Models

### Model Comparison

| Model | Dimensions | Max Tokens | MTEB Score | Speed | Cost | Best For |
|-------|-----------|-----------|-----------|-------|------|----------|
| text-embedding-3-small (OpenAI) | 1536 | 8191 | 62.3 | Fast | $0.02/1M | Cost-sensitive production |
| text-embedding-3-large (OpenAI) | 3072 | 8191 | 64.6 | Medium | $0.13/1M | High-quality retrieval |
| voyage-3 (Voyage AI) | 1024 | 32000 | 67.5 | Medium | $0.06/1M | Long documents, code |
| voyage-code-3 (Voyage AI) | 1024 | 32000 | — | Medium | $0.06/1M | Code retrieval |
| E5-mistral-7b-instruct | 4096 | 32768 | 66.6 | Slow | Self-hosted | On-prem, privacy |
| BGE-M3 | 1024 | 8192 | 64.2 | Medium | Self-hosted | Multilingual, hybrid search |
| Cohere embed-v3 | 1024 | 512 | 64.5 | Fast | $0.10/1M | Multilingual production |
| nomic-embed-text-v1.5 | 768 | 8192 | 62.3 | Fast | Self-hosted | Open source, lightweight |

### Embedding Best Practices

```
1. Match embedding model to your content type:
   - General text → text-embedding-3-small/large
   - Code → voyage-code-3 or specialized code embeddings
   - Multilingual → Cohere embed-v3 or BGE-M3
   - Long docs → voyage-3 (32K context)

2. Normalize embeddings for cosine similarity search

3. Batch embedding requests (100-500 texts per batch) for throughput

4. Cache embeddings — recomputing is wasteful:
   - Hash the input text
   - Store: {text_hash: embedding_vector}
   - Invalidate when content changes

5. Dimensionality reduction:
   - OpenAI models support custom dimensions (e.g., 256 instead of 1536)
   - Reduces storage and speeds search with minimal quality loss
   - Test quality impact before deploying reduced dimensions

6. Query-document asymmetry:
   - Some models use different prefixes for queries vs documents
   - E5: "query: " prefix for queries, "passage: " for documents
   - Forgetting prefixes silently degrades retrieval quality
```

---

## Vector Store Patterns

### Index Selection

| Index Type | Build Time | Query Time | Memory | Recall | Best For |
|-----------|-----------|-----------|--------|--------|----------|
| Flat (brute-force) | O(1) | O(n) | Low | 100% | <100K vectors |
| IVF (Inverted File) | Medium | O(n/k) | Low | 95-99% | 100K-10M vectors |
| HNSW | Slow | O(log n) | High | 99%+ | Most production use |
| ScaNN | Medium | O(log n) | Medium | 99%+ | Large-scale (Google) |
| DiskANN | Slow | O(log n) | Low (SSD) | 99%+ | Cost-sensitive large-scale |

### HNSW Tuning

```
Parameters:
- M (max connections per node): 16-64
  Higher M = better recall, more memory, slower build
  Default: 16 for most cases, 32-48 for high-recall needs

- ef_construction (search width during build): 100-500
  Higher = better recall, slower index build
  Default: 200 is good for most cases

- ef_search (search width during query): 50-500
  Higher = better recall, slower queries
  Default: Start at 100, tune based on recall/latency tradeoff

Typical configurations:
  High recall, latency tolerant:   M=32, ef_construction=400, ef_search=200
  Balanced:                         M=16, ef_construction=200, ef_search=100
  Low latency, accept lower recall: M=12, ef_construction=100, ef_search=50
```

### Distance Metrics

```
Cosine Similarity (most common):
  - Range: -1 to 1 (1 = identical)
  - Invariant to vector magnitude
  - Best for: normalized embeddings, semantic similarity

Euclidean Distance (L2):
  - Range: 0 to ∞ (0 = identical)
  - Sensitive to vector magnitude
  - Best for: when magnitude matters

Dot Product (Inner Product):
  - Range: -∞ to ∞
  - Equivalent to cosine for normalized vectors
  - Best for: maximum inner product search (MIPS)

Recommendation: Use cosine similarity. It's the default for OpenAI and most embedding models.
```

---

## Hybrid Search

### BM25 + Vector Search

```
Architecture:
  Query → [BM25 Search] → Sparse results (keyword matching)
        → [Vector Search] → Dense results (semantic matching)
        → [Fusion] → Merged results
        → [Reranking] → Final ranked results
```

### Reciprocal Rank Fusion (RRF)

```
For each document d appearing in any result list:
  RRF_score(d) = Σ (weight_i / (k + rank_i(d)))

Where:
  k = 60 (standard constant)
  weight_i = weight for result list i
  rank_i(d) = rank of document d in list i

Typical weights:
  Dense (semantic): 0.7
  Sparse (keyword):  0.3

Adjust based on query type:
  Factual queries → increase sparse weight (0.5/0.5)
  Conceptual queries → increase dense weight (0.8/0.2)
```

### When to Use Hybrid Search

```
Use hybrid search when:
  - Users search with specific terms (product names, error codes, IDs)
  - Documents contain technical jargon or abbreviations
  - Exact keyword matches matter (legal, medical, compliance)
  - You need both semantic understanding AND term matching

Use dense-only when:
  - Queries are conversational/natural language
  - Synonym matching is important
  - Cross-language retrieval is needed

Use sparse-only when:
  - Exact string matching is sufficient
  - Documents are very short (titles, tags)
  - Low latency is critical and embeddings add too much overhead
```

---

## Reranking

### Cross-Encoder Reranking

```
Stage 1: Retrieve top-20 to top-100 candidates (fast, approximate)
Stage 2: Rerank with cross-encoder (slow, precise) → return top-5

Why two stages:
  - Vector search: O(log n) on millions of documents → fast but approximate
  - Cross-encoder: O(n) on candidate set → slow but precise
  - Combined: best of both worlds
```

### Reranker Models

| Model | Latency (20 docs) | Quality | Cost | Best For |
|-------|-------------------|---------|------|----------|
| cross-encoder/ms-marco-MiniLM-L-12-v2 | ~50ms | Good | Self-hosted | General purpose, fast |
| cross-encoder/ms-marco-MiniLM-L-6-v2 | ~25ms | Decent | Self-hosted | Latency-critical |
| BAAI/bge-reranker-v2-m3 | ~80ms | Very good | Self-hosted | Multilingual |
| Cohere rerank-v3 | ~100ms | Excellent | API ($1/1K) | Production, high quality |
| Jina Reranker v2 | ~60ms | Very good | API | Good balance |
| voyage-rerank-2 | ~80ms | Excellent | API | Code + text |

### ColBERT Late Interaction

```
Unlike cross-encoders (which concatenate query+doc):
  ColBERT: Embed query and document SEPARATELY, then compute fine-grained similarity

Advantages:
  - Document embeddings can be precomputed and cached
  - Much faster than cross-encoders at query time
  - Better than bi-encoders for precision

Architecture:
  Query tokens: [q1, q2, q3, ...]   → query token embeddings
  Doc tokens:   [d1, d2, d3, ...]    → document token embeddings (pre-computed)
  Score = Σ max_j(sim(qi, dj))       → MaxSim aggregation

When to use:
  - Large candidate sets (100+ documents to rerank)
  - Need better precision than bi-encoders but faster than cross-encoders
  - Can precompute document embeddings
```

---

## RAG Evaluation Metrics

### Core Metrics

| Metric | Measures | Range | Target |
|--------|----------|-------|--------|
| **Faithfulness** | Are claims grounded in context? | 0-1 | >0.85 |
| **Answer Relevance** | Does answer address the question? | 0-1 | >0.80 |
| **Context Precision** | How much retrieved context is relevant? | 0-1 | >0.70 |
| **Context Recall** | Is all necessary info retrieved? | 0-1 | >0.80 |
| **Answer Correctness** | Is the answer factually correct? | 0-1 | >0.85 |

### Retrieval Metrics

```
Hit Rate (Recall@k):
  = (queries where at least one relevant doc in top-k) / (total queries)
  Target: >0.90 for k=10

Mean Reciprocal Rank (MRR):
  = average of (1 / rank of first relevant result)
  Target: >0.70

Normalized Discounted Cumulative Gain (NDCG@k):
  = measures ranking quality with graded relevance
  Target: >0.75

Precision@k:
  = (relevant documents in top-k) / k
  Target: varies by use case
```

### Evaluation Dataset Construction

```
Minimum viable evaluation set:
  - 50 query-answer pairs for initial development
  - 200+ pairs for production evaluation
  - Cover: easy (40%), medium (40%), hard (20%) queries

Each test case should include:
  - query: The user's question
  - ground_truth_answer: The correct answer
  - relevant_doc_ids: Documents that contain the answer
  - category: Query type (factual, analytical, comparison, etc.)
  - difficulty: easy / medium / hard

Sources for test cases:
  1. Real user queries from logs (best source)
  2. Domain expert authored questions
  3. LLM-generated questions from documents (with human review)
  4. FAQs and documentation

Common mistake: Only testing easy queries. Include:
  - Multi-hop questions requiring information from multiple chunks
  - Questions with no answer in the corpus
  - Ambiguous questions
  - Questions requiring temporal reasoning
  - Paraphrased versions of the same question
```

---

## RAG Failure Modes

### Diagnosis Guide

| Symptom | Likely Cause | Fix |
|---------|-------------|-----|
| Answers are wrong despite relevant docs in DB | Bad chunking — relevant info split across chunks | Increase chunk size or use overlap |
| Returns irrelevant documents | Bad embeddings or chunk quality | Try better embedding model, add metadata filtering |
| Answers correct but incomplete | Not enough context retrieved | Increase top-k, add query expansion |
| Hallucinated facts not in context | Model ignoring context, poor prompt | Strengthen "only use context" instruction, lower temperature |
| Slow responses | Too many chunks, large context | Reduce top-k, add reranking, compress context |
| Good for simple but bad for complex queries | Single retrieval pass insufficient | Add query decomposition, iterative retrieval |
| Works for English, fails for other languages | Embedding model not multilingual | Switch to multilingual embeddings (BGE-M3, Cohere v3) |
| Duplicated information in context | Overlapping chunks from same source | Deduplicate by document ID, reduce overlap |
| Answers questions not in the corpus | No "I don't know" behavior | Add explicit out-of-scope handling in prompt |

### Debugging Workflow

```
1. Check retrieval first:
   - Log the retrieved chunks for each query
   - Are the relevant chunks being retrieved?
   - If no → embedding/chunking problem
   - If yes → generation problem

2. Check chunk quality:
   - Read the top-5 chunks manually
   - Do they contain the answer?
   - Are they coherent (not split mid-sentence)?
   - Is there noise (irrelevant content)?

3. Check the prompt:
   - Is the context clearly formatted?
   - Is the model instructed to use only the context?
   - Does it have a fallback for unanswerable questions?

4. Check the model:
   - Try a stronger model (e.g., GPT-4o instead of GPT-4o-mini)
   - Lower the temperature
   - Check if the context window is being exceeded
```

---

## Production RAG Checklist

```markdown
### Ingestion Pipeline
- [ ] Documents are parsed and cleaned (remove boilerplate, headers/footers)
- [ ] Chunking strategy is tuned for your content type
- [ ] Metadata is extracted and stored (source, date, section, page)
- [ ] Duplicate documents are detected and handled
- [ ] Incremental updates work (add/update/delete documents)
- [ ] Pipeline is idempotent (re-running doesn't create duplicates)

### Retrieval
- [ ] Embedding model matches content type and language
- [ ] Vector index is tuned (HNSW parameters, distance metric)
- [ ] Hybrid search if content contains specific terms/codes
- [ ] Metadata filtering is available (date, source, category)
- [ ] Query preprocessing (rewriting, expansion) is configured
- [ ] Top-k and relevance threshold are tuned

### Generation
- [ ] System prompt enforces grounded answers
- [ ] Out-of-scope questions handled ("I don't know")
- [ ] Source citations are included in responses
- [ ] Context window budget is managed
- [ ] Temperature is set appropriately (usually 0-0.2 for RAG)
- [ ] Streaming is implemented for user experience

### Evaluation
- [ ] Evaluation dataset exists (50+ query-answer pairs minimum)
- [ ] Automated evaluation runs on every change
- [ ] Metrics tracked: faithfulness, relevance, precision, recall
- [ ] Regression detection for quality drops
- [ ] Human evaluation done periodically

### Operations
- [ ] Latency tracked (p50, p95, p99)
- [ ] Cost tracked per query
- [ ] Error rates monitored
- [ ] Vector store health checked
- [ ] Embedding model availability monitored
- [ ] Alerting configured for quality drops
```
