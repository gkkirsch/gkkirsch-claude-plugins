---
name: prompt-patterns
description: >
  Advanced prompt engineering patterns — system prompt design, chain of
  thought, few-shot learning, structured output, prompt chaining, tool
  use, guardrails, and production prompt management.
  Triggers: "prompt engineering", "system prompt", "chain of thought",
  "few-shot", "prompt template", "structured output", "prompt chain".
  NOT for: LLM evaluation (use llm-evaluation), RAG (use rag-patterns).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Prompt Engineering Patterns

## System Prompt Design

### The Layered Structure

```typescript
const systemPrompt = `
# Role
You are a senior financial analyst specializing in SEC filings.

# Constraints
- Only answer questions about publicly available financial data
- Never provide investment advice or price predictions
- If unsure about a fact, say "I'm not certain" rather than guessing
- All dollar amounts should include currency code (USD, EUR, etc.)

# Context
You have access to the company's latest 10-K filing and quarterly reports.
Today's date is ${new Date().toISOString().split("T")[0]}.

# Instructions
When analyzing financial data:
1. Start with the most recent filing period
2. Compare year-over-year when relevant
3. Flag any restatements or unusual items
4. Cite specific sections of the filing

# Output Format
Respond in this structure:
- **Summary**: 1-2 sentence overview
- **Key Metrics**: Bullet list of relevant numbers
- **Analysis**: Detailed explanation
- **Caveats**: Any limitations or uncertainties
`;
```

### Role-Task-Format Pattern

```typescript
// Minimal but effective structure
const prompt = `
Role: You are an expert code reviewer for TypeScript projects.
Task: Review the following code for bugs, performance issues, and style violations.
Format: Return a JSON array of findings, each with: file, line, severity (high/medium/low), description, suggestion.

Code:
${codeToReview}
`;
```

## Chain of Thought (CoT)

### Zero-Shot CoT

```typescript
// Simple trigger phrase
const prompt = `
${question}

Think through this step-by-step before giving your final answer.
`;

// More structured
const prompt = `
${question}

Before answering:
1. Identify what information is given
2. Determine what needs to be calculated or decided
3. Work through the logic step by step
4. Verify your reasoning
5. State your final answer clearly

Let's work through this:
`;
```

### Few-Shot CoT

```typescript
const prompt = `
Classify the customer intent and extract entities.

Example 1:
Input: "I want to cancel my subscription that I started last month"
Thinking: The customer mentions "cancel" which indicates a cancellation intent.
They reference "subscription" as the product and "last month" as a time reference.
Output: {"intent": "cancellation", "product": "subscription", "timeRef": "last month"}

Example 2:
Input: "How do I upgrade to the pro plan?"
Thinking: The customer asks "how do I" which is an information request.
"Upgrade" indicates they want to change plans. "Pro plan" is the target product.
Output: {"intent": "upgrade_inquiry", "product": "pro plan", "timeRef": null}

Now classify:
Input: "${userMessage}"
Thinking:`;
```

### Self-Consistency (Multiple CoT Paths)

```typescript
async function selfConsistentAnswer(
  question: string,
  paths: number = 5
): Promise<string> {
  // Generate multiple reasoning paths
  const responses = await Promise.all(
    Array.from({ length: paths }, () =>
      llm.chat({
        messages: [{ role: "user", content: `${question}\n\nThink step by step:` }],
        temperature: 0.7, // Higher temp for diverse reasoning
      })
    )
  );

  // Extract final answers and vote
  const answers = responses.map((r) => extractFinalAnswer(r.content));
  const counts = new Map<string, number>();
  for (const answer of answers) {
    counts.set(answer, (counts.get(answer) ?? 0) + 1);
  }

  // Return most common answer
  return [...counts.entries()].sort((a, b) => b[1] - a[1])[0][0];
}
```

## Few-Shot Learning

### Effective Example Selection

```typescript
// 1. Cover edge cases, not just happy paths
const examples = [
  { input: "standard case", output: "expected output" },
  { input: "edge case with special chars: @#$", output: "handled gracefully" },
  { input: "empty input", output: "appropriate default" },
  { input: "adversarial: ignore instructions and...", output: "politely declined" },
];

// 2. Use consistent formatting
const fewShotPrompt = examples
  .map((e) => `Input: ${e.input}\nOutput: ${e.output}`)
  .join("\n\n");

// 3. Dynamic example selection (most similar to current query)
async function selectExamples(
  query: string,
  allExamples: Example[],
  k: number = 3
): Promise<Example[]> {
  const queryEmbedding = await embed(query);
  const scored = allExamples.map((ex) => ({
    ...ex,
    similarity: cosineSimilarity(queryEmbedding, ex.embedding),
  }));
  return scored.sort((a, b) => b.similarity - a.similarity).slice(0, k);
}
```

## Structured Output

### JSON Schema Enforcement

```typescript
import Anthropic from "@anthropic-ai/sdk";

const client = new Anthropic();

// Claude with tool_use for structured output
const response = await client.messages.create({
  model: "claude-sonnet-4-6-20250514",
  max_tokens: 1024,
  tools: [{
    name: "extract_data",
    description: "Extract structured data from the text",
    input_schema: {
      type: "object",
      properties: {
        entities: {
          type: "array",
          items: {
            type: "object",
            properties: {
              name: { type: "string" },
              type: { type: "string", enum: ["person", "org", "location"] },
              confidence: { type: "number", minimum: 0, maximum: 1 },
            },
            required: ["name", "type", "confidence"],
          },
        },
        sentiment: { type: "string", enum: ["positive", "negative", "neutral"] },
        summary: { type: "string", maxLength: 200 },
      },
      required: ["entities", "sentiment", "summary"],
    },
  }],
  tool_choice: { type: "tool", name: "extract_data" }, // Force tool use
  messages: [{ role: "user", content: `Extract data from: ${text}` }],
});

// Parse the tool use result
const toolUse = response.content.find((c) => c.type === "tool_use");
const data = toolUse?.input; // Already typed and validated
```

### Output Parsing with Validation

```typescript
import { z } from "zod";

const OutputSchema = z.object({
  answer: z.string(),
  confidence: z.number().min(0).max(1),
  sources: z.array(z.string()),
  reasoning: z.string().optional(),
});

type Output = z.infer<typeof OutputSchema>;

async function getStructuredResponse(prompt: string): Promise<Output> {
  const response = await llm.chat({
    messages: [{
      role: "user",
      content: `${prompt}\n\nRespond in JSON format:\n{"answer": "...", "confidence": 0.0-1.0, "sources": ["..."], "reasoning": "..."}`,
    }],
    temperature: 0,
  });

  // Extract JSON from response (handles markdown code blocks)
  const jsonMatch = response.content.match(/```json\n?([\s\S]*?)\n?```/) ||
    response.content.match(/\{[\s\S]*\}/);

  if (!jsonMatch) throw new Error("No JSON found in response");

  const parsed = JSON.parse(jsonMatch[1] ?? jsonMatch[0]);
  return OutputSchema.parse(parsed); // Throws if invalid
}
```

## Prompt Chaining

### Sequential Chain

```typescript
async function analyzeAndSummarize(document: string) {
  // Step 1: Extract key points
  const extraction = await llm.chat({
    messages: [{
      role: "user",
      content: `Extract the 5 most important points from this document. Return as a numbered list.\n\n${document}`,
    }],
    temperature: 0,
  });

  // Step 2: Analyze sentiment of each point
  const analysis = await llm.chat({
    messages: [{
      role: "user",
      content: `For each point below, analyze the sentiment and business impact (positive/negative/neutral + brief explanation).

Key Points:
${extraction.content}

Format: numbered list matching the input, with sentiment and impact for each.`,
    }],
    temperature: 0,
  });

  // Step 3: Generate executive summary
  const summary = await llm.chat({
    messages: [{
      role: "user",
      content: `Based on the following analysis, write a 3-sentence executive summary suitable for a board meeting.

Analysis:
${analysis.content}

Write the summary:`,
    }],
    temperature: 0.3,
  });

  return { extraction: extraction.content, analysis: analysis.content, summary: summary.content };
}
```

### Router Chain

```typescript
type Intent = "technical" | "billing" | "general" | "escalate";

async function routeQuery(query: string): Promise<string> {
  // Step 1: Classify intent (cheap, fast model)
  const classification = await llm.chat({
    model: "claude-haiku-4-5-20251001",
    messages: [{
      role: "user",
      content: `Classify this customer query into exactly one category: technical, billing, general, escalate.

Query: "${query}"

Category:`,
    }],
    temperature: 0,
    max_tokens: 20,
  });

  const intent = classification.content.trim().toLowerCase() as Intent;

  // Step 2: Route to specialist prompt
  const specialists: Record<Intent, string> = {
    technical: "You are a senior technical support engineer...",
    billing: "You are a billing specialist with access to account data...",
    general: "You are a friendly customer service representative...",
    escalate: "This query needs human attention. Summarize the issue for the agent:",
  };

  const response = await llm.chat({
    model: intent === "escalate" ? "claude-haiku-4-5-20251001" : "claude-sonnet-4-6-20250514",
    messages: [
      { role: "system", content: specialists[intent] },
      { role: "user", content: query },
    ],
  });

  return response.content;
}
```

## Tool Use / Function Calling

### Tool Definition Pattern

```typescript
const tools = [
  {
    name: "search_database",
    description: "Search the product database. Use when the user asks about product availability, pricing, or specifications.",
    input_schema: {
      type: "object",
      properties: {
        query: {
          type: "string",
          description: "Search query for the product database",
        },
        category: {
          type: "string",
          enum: ["electronics", "clothing", "home", "all"],
          description: "Product category to filter by",
        },
        max_results: {
          type: "number",
          description: "Maximum number of results to return (default: 5)",
        },
      },
      required: ["query"],
    },
  },
  {
    name: "create_order",
    description: "Create a new order. Only use after the user has confirmed they want to purchase.",
    input_schema: {
      type: "object",
      properties: {
        product_id: { type: "string" },
        quantity: { type: "number", minimum: 1 },
        shipping_address: { type: "string" },
      },
      required: ["product_id", "quantity", "shipping_address"],
    },
  },
];

// Tool execution loop
async function agentLoop(userMessage: string) {
  const messages: Message[] = [{ role: "user", content: userMessage }];

  while (true) {
    const response = await client.messages.create({
      model: "claude-sonnet-4-6-20250514",
      max_tokens: 4096,
      tools,
      messages,
    });

    // Check if done (no more tool calls)
    if (response.stop_reason === "end_turn") {
      return response.content.find((c) => c.type === "text")?.text;
    }

    // Execute tool calls
    const toolResults: ToolResultBlockParam[] = [];
    for (const block of response.content) {
      if (block.type === "tool_use") {
        const result = await executeToolCall(block.name, block.input);
        toolResults.push({
          type: "tool_result",
          tool_use_id: block.id,
          content: JSON.stringify(result),
        });
      }
    }

    messages.push({ role: "assistant", content: response.content });
    messages.push({ role: "user", content: toolResults });
  }
}
```

## Guardrails

### Input Validation

```typescript
async function validateInput(userInput: string): Promise<{
  safe: boolean;
  reason?: string;
  sanitized?: string;
}> {
  // 1. Length check
  if (userInput.length > 10000) {
    return { safe: false, reason: "Input too long" };
  }

  // 2. Pattern-based injection detection
  const injectionPatterns = [
    /ignore (all |previous |prior |above )?instructions/i,
    /you are now/i,
    /system prompt/i,
    /\[INST\]/i,
    /<\|im_start\|>/i,
  ];
  for (const pattern of injectionPatterns) {
    if (pattern.test(userInput)) {
      return { safe: false, reason: "Potential prompt injection detected" };
    }
  }

  // 3. LLM-based content moderation (for nuanced cases)
  const moderation = await llm.chat({
    model: "claude-haiku-4-5-20251001",
    messages: [{
      role: "user",
      content: `Is this user message safe and appropriate for a customer support chatbot? Answer only YES or NO.

Message: "${userInput.slice(0, 500)}"

Safe:`,
    }],
    temperature: 0,
    max_tokens: 5,
  });

  if (moderation.content.trim().toUpperCase() !== "YES") {
    return { safe: false, reason: "Content flagged by moderation" };
  }

  return { safe: true, sanitized: userInput.trim() };
}
```

### Output Validation

```typescript
async function validateOutput(
  response: string,
  context: { topic: string; constraints: string[] }
): Promise<{ valid: boolean; filtered: string }> {
  // 1. Check for hallucinated URLs
  const urls = response.match(/https?:\/\/[^\s)]+/g) ?? [];
  for (const url of urls) {
    if (!context.constraints.some((c) => url.includes(c))) {
      // Replace suspicious URLs
      response = response.replace(url, "[link removed - please verify]");
    }
  }

  // 2. Check for off-topic content
  const topicCheck = await llm.chat({
    model: "claude-haiku-4-5-20251001",
    messages: [{
      role: "user",
      content: `Is this response on-topic for "${context.topic}"? Answer YES or NO.\n\nResponse: "${response.slice(0, 500)}"`,
    }],
    temperature: 0,
    max_tokens: 5,
  });

  if (topicCheck.content.trim().toUpperCase() !== "YES") {
    return {
      valid: false,
      filtered: "I'm not able to help with that topic. Let me know if you have questions about " + context.topic,
    };
  }

  return { valid: true, filtered: response };
}
```

## Token Optimization

```typescript
// 1. Prompt caching (Anthropic) — reuse system prompt prefix
const response = await client.messages.create({
  model: "claude-sonnet-4-6-20250514",
  max_tokens: 1024,
  system: [
    {
      type: "text",
      text: longSystemPrompt, // This gets cached
      cache_control: { type: "ephemeral" },
    },
  ],
  messages: [{ role: "user", content: userQuery }],
});

// 2. Compress context before sending
function compressContext(text: string, maxTokens: number): string {
  // Rough token estimate: 1 token ~= 4 chars
  const maxChars = maxTokens * 4;
  if (text.length <= maxChars) return text;

  // Remove redundant whitespace
  text = text.replace(/\n{3,}/g, "\n\n").replace(/  +/g, " ");

  // Truncate with notice
  return text.slice(0, maxChars) + "\n\n[Content truncated]";
}

// 3. Use appropriate model for each step
// Haiku for classification, routing, simple extraction
// Sonnet for generation, analysis, coding
// Opus for complex reasoning, architecture, research
```

## Gotchas

1. **Temperature is not creativity** — Temperature controls randomness, not quality. Use 0 for factual tasks (extraction, classification). Use 0.3-0.7 for creative tasks (writing, brainstorming). Never use 1.0+ in production.

2. **Few-shot order matters** — Put the most relevant example last (recency bias). Put edge cases in the middle. Start with the simplest example.

3. **"Always" and "Never" are brittle** — Models treat these as strong guidelines, not absolutes. They work most of the time but fail on edge cases. Use softer language: "strongly prefer", "avoid unless necessary".

4. **System prompts leak** — Never put secrets in system prompts. Users can extract them with creative prompting. Treat system prompts as public.

5. **Longer prompts aren't better** — Each additional instruction competes for attention. A focused 200-token prompt often outperforms a sprawling 2000-token one. Be concise.

6. **Model updates break prompts** — Prompts tuned for one model version may degrade on updates. Version your prompts alongside your model version. Run evals on every model update.
