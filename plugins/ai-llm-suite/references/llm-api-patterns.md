# LLM API Patterns Reference

Production patterns for working with LLM APIs — streaming, retry logic, token counting, caching, fallback models, and cost management.

---

## Streaming Patterns

### OpenAI Streaming

```python
from openai import OpenAI

client = OpenAI()

# Basic streaming
stream = client.chat.completions.create(
    model="gpt-4o",
    messages=[{"role": "user", "content": "Explain RAG"}],
    stream=True
)

full_response = ""
for chunk in stream:
    delta = chunk.choices[0].delta
    if delta.content:
        print(delta.content, end="", flush=True)
        full_response += delta.content
    if chunk.choices[0].finish_reason:
        print(f"\nFinish reason: {chunk.choices[0].finish_reason}")
```

### Anthropic Streaming

```python
import anthropic

client = anthropic.Anthropic()

with client.messages.stream(
    model="claude-sonnet-4-20250514",
    max_tokens=1024,
    messages=[{"role": "user", "content": "Explain RAG"}]
) as stream:
    for text in stream.text_stream:
        print(text, end="", flush=True)

# With events for detailed tracking
with client.messages.stream(
    model="claude-sonnet-4-20250514",
    max_tokens=1024,
    messages=[{"role": "user", "content": "Explain RAG"}]
) as stream:
    for event in stream:
        if event.type == "content_block_delta":
            print(event.delta.text, end="")
        elif event.type == "message_stop":
            print("\n[Done]")
```

### Server-Sent Events (SSE) API Pattern

```python
from fastapi import FastAPI
from fastapi.responses import StreamingResponse
import json

app = FastAPI()

async def generate_stream(query: str):
    """Stream responses as SSE events."""
    # Send status
    yield f"data: {json.dumps({'type': 'status', 'message': 'Processing...'})}\n\n"

    stream = client.chat.completions.create(
        model="gpt-4o",
        messages=[{"role": "user", "content": query}],
        stream=True
    )

    for chunk in stream:
        if chunk.choices[0].delta.content:
            yield f"data: {json.dumps({'type': 'content', 'text': chunk.choices[0].delta.content})}\n\n"

    yield f"data: {json.dumps({'type': 'done'})}\n\n"
    yield "data: [DONE]\n\n"

@app.post("/api/chat")
async def chat(request: dict):
    return StreamingResponse(
        generate_stream(request["query"]),
        media_type="text/event-stream",
        headers={"Cache-Control": "no-cache", "Connection": "keep-alive"}
    )
```

### Client-Side SSE Consumption

```typescript
// Frontend SSE consumption
async function streamChat(query: string, onChunk: (text: string) => void): Promise<void> {
  const response = await fetch('/api/chat', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ query })
  });

  const reader = response.body!.getReader();
  const decoder = new TextDecoder();
  let buffer = '';

  while (true) {
    const { done, value } = await reader.read();
    if (done) break;

    buffer += decoder.decode(value, { stream: true });
    const lines = buffer.split('\n');
    buffer = lines.pop() || '';

    for (const line of lines) {
      if (line.startsWith('data: ')) {
        const data = line.slice(6);
        if (data === '[DONE]') return;

        try {
          const event = JSON.parse(data);
          if (event.type === 'content') {
            onChunk(event.text);
          }
        } catch (e) {
          // Skip invalid JSON
        }
      }
    }
  }
}
```

---

## Retry Logic

### Exponential Backoff with Jitter

```python
import time
import random
from functools import wraps
from openai import RateLimitError, APITimeoutError, APIConnectionError, InternalServerError

RETRYABLE_ERRORS = (
    RateLimitError,
    APITimeoutError,
    APIConnectionError,
    InternalServerError,
)

def retry_with_backoff(
    max_retries: int = 3,
    base_delay: float = 1.0,
    max_delay: float = 60.0,
    jitter: bool = True
):
    """Decorator for LLM API calls with exponential backoff."""
    def decorator(func):
        @wraps(func)
        def wrapper(*args, **kwargs):
            for attempt in range(max_retries + 1):
                try:
                    return func(*args, **kwargs)
                except RETRYABLE_ERRORS as e:
                    if attempt == max_retries:
                        raise

                    delay = min(base_delay * (2 ** attempt), max_delay)
                    if jitter:
                        delay *= 0.5 + random.random()

                    # Check for Retry-After header
                    if hasattr(e, 'response') and e.response:
                        retry_after = e.response.headers.get('Retry-After')
                        if retry_after:
                            delay = max(delay, float(retry_after))

                    time.sleep(delay)
        return wrapper
    return decorator

# Usage
@retry_with_backoff(max_retries=3)
def call_llm(messages: list) -> str:
    response = client.chat.completions.create(
        model="gpt-4o",
        messages=messages
    )
    return response.choices[0].message.content
```

### Retry Classification

```
Always retry:
  - 429 Rate Limit Exceeded (with backoff)
  - 500 Internal Server Error
  - 502 Bad Gateway
  - 503 Service Unavailable
  - Connection timeouts
  - Network errors

Never retry:
  - 400 Bad Request (fix the request)
  - 401 Unauthorized (fix credentials)
  - 403 Forbidden (fix permissions)
  - 404 Not Found (fix endpoint/model)
  - 422 Unprocessable Entity (fix input)
  - Content filter rejections (modify content)

Maybe retry (with modification):
  - 413 Payload Too Large → reduce input
  - Context length exceeded → truncate input
  - Token limit exceeded → reduce max_tokens
```

---

## Token Counting

### Pre-Request Token Estimation

```python
import tiktoken

class TokenCounter:
    """Count tokens for different models."""

    def __init__(self):
        self.encodings = {}

    def count(self, text: str, model: str = "gpt-4o") -> int:
        """Count tokens in text for a specific model."""
        if model not in self.encodings:
            try:
                self.encodings[model] = tiktoken.encoding_for_model(model)
            except KeyError:
                self.encodings[model] = tiktoken.get_encoding("cl100k_base")
        return len(self.encodings[model].encode(text))

    def count_messages(self, messages: list[dict], model: str = "gpt-4o") -> int:
        """Count tokens in a messages array (includes message overhead)."""
        # Per-message overhead varies by model
        tokens_per_message = 3  # OpenAI chat models
        tokens_per_name = 1

        total = 0
        for message in messages:
            total += tokens_per_message
            for key, value in message.items():
                total += self.count(str(value), model)
                if key == "name":
                    total += tokens_per_name

        total += 3  # Every reply is primed with <|start|>assistant<|message|>
        return total

    def estimate_cost(self, input_tokens: int, output_tokens: int, model: str) -> float:
        """Estimate cost in USD."""
        pricing = {
            "gpt-4o": (2.50, 10.00),
            "gpt-4o-mini": (0.15, 0.60),
            "claude-3-5-sonnet-20241022": (3.00, 15.00),
            "claude-3-haiku-20240307": (0.25, 1.25),
        }
        input_price, output_price = pricing.get(model, (0, 0))
        return (input_tokens * input_price + output_tokens * output_price) / 1_000_000

# Usage
counter = TokenCounter()
msg_tokens = counter.count_messages(messages, "gpt-4o")
estimated_cost = counter.estimate_cost(msg_tokens, 500, "gpt-4o")
```

### Context Window Budgeting

```
Budget allocation strategy for RAG:

Total context: 128,000 tokens (GPT-4o)

┌─────────────────────────────────────────────┐
│ System prompt:          500-2000 tokens      │
│ Chat history:           1000-4000 tokens     │
│ Retrieved context:      2000-8000 tokens     │
│ Reserved for response:  1000-4000 tokens     │
│ Buffer:                 500 tokens            │
└─────────────────────────────────────────────┘

In practice, use 10-20% of context window:
  128K context → use ~12-25K tokens total
  Why? Long contexts degrade attention and increase cost

Guidelines:
  - System prompt: keep under 2000 tokens
  - Chat history: summarize or truncate beyond 10 messages
  - Retrieved context: 5-10 chunks × 200-500 tokens each
  - Response: set max_tokens to expected response length
```

---

## Caching Strategies

### Exact Match Cache

```
Approach: Hash the full request (model + messages + temperature) → cache response
When to use: Deterministic requests (temperature=0), repeated queries
Hit rate: Low (queries rarely repeat exactly)
Implementation: Redis with TTL

Cache key = SHA256(model + json(messages) + str(temperature))
TTL = 1 hour (or longer for static content)
```

### Semantic Cache

```
Approach: Embed the query → find similar cached queries → return cached response
When to use: Many similar queries (customer support, FAQ)
Hit rate: High if queries cluster around common topics
Implementation: Vector store + similarity threshold

Cache lookup:
  1. Embed user query
  2. Search cache collection (cosine similarity)
  3. If similarity > 0.95 → return cached response
  4. Otherwise → call LLM, cache result

Threshold tuning:
  0.99 = very conservative (almost exact match only)
  0.95 = good balance for production
  0.90 = aggressive (may return wrong cached answer)
```

### Prompt Caching (Provider-Level)

```
OpenAI:
  - Automatic for prompts with shared prefixes > 1024 tokens
  - 50% discount on cached input tokens
  - Cache persists for 5-10 minutes of inactivity
  - Design prompts with static content first, dynamic content last

Anthropic:
  - Explicit cache_control blocks in messages
  - 90% discount on cached input tokens (after initial write cost)
  - TTL: 5 minutes, extended on cache hit
  - Best for: long system prompts, few-shot examples, large documents

Strategy: Put stable content (system prompt, examples, reference docs) first,
          variable content (user query, chat history) last.
```

---

## Fallback Models

### Fallback Chain Pattern

```
Primary: GPT-4o (best quality)
  ↓ on failure
Secondary: Claude 3.5 Sonnet (alternative provider)
  ↓ on failure
Tertiary: GPT-4o-mini (degraded quality, always available)
  ↓ on failure
Emergency: Static responses / cached responses / error message
```

### Fallback Decision Matrix

| Failure Type | Fallback Strategy |
|-------------|-------------------|
| Rate limit (429) | Switch to different provider |
| Server error (500/502/503) | Retry same provider first, then switch |
| Timeout | Switch to faster model |
| Context too long | Truncate input, use same model |
| Content filtered | Rephrase request, use same model |
| Model deprecated | Switch to replacement model |
| Provider outage | Switch to different provider entirely |

### Graceful Degradation

```
Tier 1: Full quality (GPT-4o / Claude Sonnet)
  - Complex reasoning, nuanced responses
  - Used when: normal operation

Tier 2: Good quality (GPT-4o-mini / Claude Haiku)
  - Simpler responses, may miss nuance
  - Used when: primary model unavailable or over budget

Tier 3: Basic quality (local model / cached responses)
  - Pre-computed answers for common queries
  - Used when: all API providers down

Tier 4: Minimal (static response)
  - "I'm currently unable to process this request. Please try again later."
  - Used when: everything is down
```

---

## Cost Management

### Cost Optimization Strategies

```
1. Model selection (biggest impact):
   GPT-4o-mini is 17x cheaper than GPT-4o for input
   Use mini for: classification, extraction, simple Q&A
   Use full for: complex reasoning, code generation, analysis

2. Prompt engineering:
   Shorter prompts = less cost
   Remove redundant instructions
   Use concise examples
   Compress system prompts

3. Caching:
   Exact match cache for repeated queries
   Semantic cache for similar queries
   Prompt caching for shared prefixes
   Expected savings: 20-50% of LLM costs

4. Token budgeting:
   Set max_tokens appropriately (not too high)
   Truncate long contexts
   Summarize chat history instead of sending full history

5. Batching:
   Use batch API for non-real-time work (50% discount on OpenAI)
   Batch embedding requests

6. Model routing:
   Classify query complexity → route to cheapest adequate model
   Expected savings: 40-60% vs always using the best model
```

### Cost Per Feature Tracking

```
Track cost at the feature level, not just total:

Feature          | Requests/day | Avg cost/req | Daily cost
─────────────────|──────────────|──────────────|───────────
Chat             | 10,000       | $0.012       | $120
Search           | 5,000        | $0.003       | $15
Summarization    | 2,000        | $0.025       | $50
Classification   | 50,000       | $0.0002      | $10
Embeddings       | 100,000      | $0.00002     | $2
─────────────────|──────────────|──────────────|───────────
Total            | 167,000      |              | $197/day

This reveals:
  - Chat is the most expensive feature (60% of cost)
  - Classification is the most efficient (50K requests for $10)
  - Summarization has the highest per-request cost
```

### Budget Alerts

```python
class BudgetManager:
    """Monitor and enforce LLM spending budgets."""

    def __init__(self, daily_limit: float = 100.0, monthly_limit: float = 2000.0):
        self.daily_limit = daily_limit
        self.monthly_limit = monthly_limit
        self.daily_spend = 0.0
        self.monthly_spend = 0.0

    def can_spend(self, estimated_cost: float) -> dict:
        if self.daily_spend + estimated_cost > self.daily_limit:
            return {"allowed": False, "reason": "daily_limit", "remaining": self.daily_limit - self.daily_spend}
        if self.monthly_spend + estimated_cost > self.monthly_limit:
            return {"allowed": False, "reason": "monthly_limit", "remaining": self.monthly_limit - self.monthly_spend}
        return {"allowed": True, "remaining_daily": self.daily_limit - self.daily_spend}

    def record_spend(self, cost: float):
        self.daily_spend += cost
        self.monthly_spend += cost

    def get_status(self) -> dict:
        return {
            "daily": {"spent": self.daily_spend, "limit": self.daily_limit, "pct": self.daily_spend / self.daily_limit * 100},
            "monthly": {"spent": self.monthly_spend, "limit": self.monthly_limit, "pct": self.monthly_spend / self.monthly_limit * 100},
        }
```

---

## API Best Practices

### Request Configuration

```python
# Production defaults
PRODUCTION_DEFAULTS = {
    "temperature": 0.1,        # Low for deterministic, consistent results
    "max_tokens": 1024,        # Set explicitly, don't rely on model default
    "top_p": 1.0,              # Don't use both temperature and top_p
    "frequency_penalty": 0,    # Usually leave at 0
    "presence_penalty": 0,     # Usually leave at 0
    "timeout": 30,             # Seconds, adjust for expected response time
}

# When to adjust:
# temperature=0: Classification, extraction, factual Q&A
# temperature=0.3-0.7: Creative writing, brainstorming
# temperature=0.7-1.0: Diverse outputs (self-consistency)
# max_tokens: Set to 2x expected response length
# timeout: Set to 3x typical response time
```

### Error Handling Checklist

```
□ Rate limiting → exponential backoff with jitter
□ Timeouts → configurable timeout, retry once
□ Context too long → truncate oldest messages / compress
□ Content filtered → log, notify, return safe response
□ Invalid API key → fail fast, alert ops
□ Model not found → fall back to known model
□ Network error → retry with backoff
□ JSON parse error in response → retry once, then use raw text
□ Empty response → retry once, then return error
□ Unexpected response format → validate, retry, then fail gracefully
```

### Request/Response Logging

```python
import logging
import time

class LLMLogger:
    """Structured logging for LLM API calls."""

    def __init__(self):
        self.logger = logging.getLogger("llm")

    def log_request(self, request_id: str, model: str, messages: list, **kwargs):
        self.logger.info("llm_request", extra={
            "request_id": request_id,
            "model": model,
            "message_count": len(messages),
            "input_chars": sum(len(m.get("content", "")) for m in messages),
            "temperature": kwargs.get("temperature"),
            "max_tokens": kwargs.get("max_tokens"),
            "stream": kwargs.get("stream", False),
        })

    def log_response(self, request_id: str, response, latency_ms: float, cost: float):
        self.logger.info("llm_response", extra={
            "request_id": request_id,
            "model": response.model,
            "input_tokens": response.usage.prompt_tokens,
            "output_tokens": response.usage.completion_tokens,
            "total_tokens": response.usage.total_tokens,
            "finish_reason": response.choices[0].finish_reason,
            "latency_ms": latency_ms,
            "cost_usd": cost,
        })

    def log_error(self, request_id: str, error: Exception, latency_ms: float):
        self.logger.error("llm_error", extra={
            "request_id": request_id,
            "error_type": type(error).__name__,
            "error_message": str(error),
            "latency_ms": latency_ms,
        })
```

---

## Provider Comparison

### Feature Matrix

| Feature | OpenAI | Anthropic | Google | Mistral | Local (vLLM) |
|---------|--------|-----------|--------|---------|--------------|
| Streaming | SSE | SSE | SSE | SSE | SSE |
| Function calling | Yes | Yes (tools) | Yes | Yes | Yes |
| JSON mode | Yes | Yes | Yes | Yes | Via grammar |
| Vision | Yes | Yes | Yes | Yes | Model-dependent |
| Prompt caching | Auto | Explicit | Context cache | No | Prefix caching |
| Batch API | Yes (50% off) | Yes (50% off) | No | Yes | N/A |
| Embeddings | Yes | Yes (voyage) | Yes | Yes | Self-hosted |
| Rate limits | Per-tier | Per-tier | Per-tier | Per-tier | Self-managed |

### API Compatibility

```
OpenAI SDK is the de facto standard. Compatible with:
  - OpenAI API (native)
  - Azure OpenAI (via azure_endpoint)
  - vLLM (OpenAI-compatible server)
  - Ollama (OpenAI-compatible mode)
  - LiteLLM (universal proxy)
  - Groq (OpenAI-compatible)
  - Together AI (OpenAI-compatible)

Anthropic has its own SDK:
  - Different message format
  - System prompt as separate parameter
  - Tool use syntax differs
  - Streaming events differ

Adapter pattern for multi-provider:
  Use an abstraction layer or LiteLLM to normalize
  across providers with a single interface.
```
