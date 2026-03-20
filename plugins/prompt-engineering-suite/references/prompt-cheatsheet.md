# Prompt Engineering Cheatsheet

## System Prompt Structure

```
1. Role      — Who the AI is (1-2 sentences)
2. Constraints — What it must NOT do (boundaries first)
3. Context   — What it has access to, today's date, etc.
4. Instructions — Step-by-step task instructions
5. Output Format — Expected response structure
```

## Prompting Techniques

| Technique | When to Use | Example Trigger |
|-----------|-------------|----------------|
| Zero-shot | Simple tasks | "Classify this as positive/negative" |
| Few-shot | Consistent formatting | Show 2-3 examples before the task |
| Chain of Thought | Reasoning/math | "Think step by step" |
| Self-consistency | High-stakes | Run 5x at temp=0.7, majority vote |
| ReAct | Tool use | "Think → Act → Observe → repeat" |
| Structured output | APIs | Force JSON via tool_use or schema |
| Prompt chain | Multi-step | Step 1 output → Step 2 input |
| Router | Multi-domain | Classify intent → specialist prompt |

## Temperature Guide

```
0.0  → Factual extraction, classification, code generation
0.3  → Balanced writing, analysis, summarization
0.5  → Creative writing with some structure
0.7  → Brainstorming, diverse ideas
1.0+ → Never in production (too random)
```

## Model Selection per Task

```
Haiku  → Routing, classification, moderation, extraction ($)
Sonnet → Generation, analysis, coding, general tasks ($$)
Opus   → Complex reasoning, architecture, research ($$$)
```

## Structured Output (Claude)

```typescript
// Force structured output via tool_use
const response = await client.messages.create({
  model: "claude-sonnet-4-6-20250514",
  tools: [{
    name: "extract",
    description: "Extract data",
    input_schema: { type: "object", properties: { ... }, required: [...] },
  }],
  tool_choice: { type: "tool", name: "extract" },
  messages: [{ role: "user", content: text }],
});
const data = response.content.find(c => c.type === "tool_use")?.input;
```

## Prompt Caching (Anthropic)

```typescript
system: [{
  type: "text",
  text: longSystemPrompt,
  cache_control: { type: "ephemeral" },  // 90% savings on repeated calls
}]
```

## Few-Shot Best Practices

```
1. Put most relevant example LAST (recency bias)
2. Include edge cases in the middle
3. Start with the simplest example
4. Use 3-5 examples (more isn't always better)
5. Match the format you want in the output
```

## Guardrails Checklist

```
Input:
  [ ] Length limit (< 10K chars)
  [ ] Pattern-based injection detection
  [ ] Content moderation (LLM or rule-based)
  [ ] User input never in system prompt directly

Output:
  [ ] JSON schema validation
  [ ] URL verification (no hallucinated links)
  [ ] Topic relevance check
  [ ] PII detection and redaction
  [ ] max_tokens set to prevent runaway costs
```

## Tool Use Loop

```
User Message → LLM → Tool Call?
  Yes → Execute Tool → Tool Result → LLM → Tool Call? (loop)
  No  → Final Response
```

Always check `stop_reason === "end_turn"` to know when the loop ends.

## Prompt Chain Pattern

```
Step 1 (Extract):   Document → Key Points
Step 2 (Analyze):   Key Points → Sentiment + Impact
Step 3 (Summarize): Analysis → Executive Summary
```

Each step: separate API call, focused prompt, validated output.

## Router Pattern

```
User Query → Classifier (Haiku, temp=0, max_tokens=20)
  → "technical" → Tech Support Prompt (Sonnet)
  → "billing"   → Billing Prompt (Sonnet)
  → "escalate"  → Summary for Human (Haiku)
```

## Common Anti-Patterns

```
1. Mega-prompt       → Split into chains
2. No max_tokens     → Set explicit limit
3. Trust raw output  → Always validate/parse
4. "Always/Never"    → Use "strongly prefer" / "avoid unless"
5. Secrets in prompt → Treat system prompt as public
6. No error handling → Retry on 429/503 with backoff
7. Hardcoded model   → Use config variable
8. temp=0 for all    → Match temp to task type
```

## Evaluation Quick Reference

```
Assertions:
  contains("expected text")
  json_schema({ ... })
  regex(/pattern/)
  llm_judge("Is this accurate?", threshold=0.8)
  similarity(reference, threshold=0.85)

Metrics:
  Pass rate        → % of test cases passing all assertions
  Score            → Average assertion pass rate per case
  Latency (P50/P99) → Response time percentiles
  Cost per request → Input + output tokens × model rate

Regression:
  Score drop > 5%      → Alert
  Pass rate drop > 10% → Block deploy
  New failures         → Investigate per-case
```

## Cost Optimization

```
1. Prompt caching (Anthropic)     → 90% savings on system prompt
2. Haiku for routing/classify     → 10-20x cheaper
3. Batch API (Anthropic)          → 50% discount
4. Shorter system prompts         → Linear savings
5. Structured output              → Fewer output tokens
6. max_tokens limit               → Prevents runaway
7. Cache embeddings               → Don't re-embed static docs
```
