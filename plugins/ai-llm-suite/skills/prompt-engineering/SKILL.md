---
name: prompt-engineering
description: >
  Production prompt engineering patterns and optimization techniques.
  Use when designing system prompts, optimizing token usage, implementing
  structured output, or building multi-step LLM workflows.
  Triggers: "write a prompt", "system prompt", "optimize prompt", "prompt template",
  "few-shot", "chain of thought", "structured output".
  NOT for: fine-tuning models, training data preparation, or non-LLM AI/ML tasks.
version: 1.0.0
allowed-tools: Read, Grep, Glob, Bash
---

# Prompt Engineering Patterns

## System Prompt Architecture

```typescript
// Structured system prompt builder
interface SystemPromptConfig {
  role: string;
  context: string;
  constraints: string[];
  outputFormat: string;
  examples?: Array<{ input: string; output: string }>;
  escapeHatch: string;
}

function buildSystemPrompt(config: SystemPromptConfig): string {
  const sections = [
    `# Role\n${config.role}`,
    `\n## Context\n${config.context}`,
    `\n## Constraints\n${config.constraints.map(c => `- ${c}`).join('\n')}`,
  ];

  if (config.examples?.length) {
    sections.push(
      `\n## Examples\n${config.examples.map(ex =>
        `Input: ${ex.input}\nOutput: ${ex.output}`
      ).join('\n\n')}`
    );
  }

  sections.push(`\n## Output Format\n${config.outputFormat}`);
  sections.push(`\n## Edge Cases\n${config.escapeHatch}`);

  return sections.join('\n');
}

// Example: Customer support classifier
const supportClassifier = buildSystemPrompt({
  role: 'You are a customer support ticket classifier for a SaaS product.',
  context: 'You receive raw customer messages and classify them by urgency and department.',
  constraints: [
    'Only use the categories listed below — never invent new ones',
    'If a message contains multiple issues, classify by the PRIMARY concern',
    'Urgent = data loss, security breach, or complete service outage',
    'Never include PII in your response',
  ],
  outputFormat: `Respond with JSON only:
\`\`\`json
{
  "category": "billing" | "technical" | "feature_request" | "account" | "other",
  "urgency": "low" | "medium" | "high" | "urgent",
  "confidence": 0.0-1.0,
  "reasoning": "one sentence explanation"
}
\`\`\``,
  examples: [
    {
      input: "I can't log in and I have a demo in 10 minutes",
      output: '{"category":"technical","urgency":"urgent","confidence":0.95,"reasoning":"Login failure blocking time-sensitive business activity"}'
    },
    {
      input: "Would be nice to have dark mode",
      output: '{"category":"feature_request","urgency":"low","confidence":0.99,"reasoning":"Feature suggestion with no blocking impact"}'
    }
  ],
  escapeHatch: 'If the message is not in English or is unintelligible, return category "other" with urgency "low" and confidence below 0.5.'
});
```

## Few-Shot Prompting

```typescript
// Dynamic few-shot selector — pick most relevant examples
interface Example {
  input: string;
  output: string;
  tags: string[];
  embedding?: number[];
}

class FewShotSelector {
  private examples: Example[];

  constructor(examples: Example[]) {
    this.examples = examples;
  }

  // Tag-based selection
  selectByTags(tags: string[], count: number = 3): Example[] {
    return this.examples
      .map(ex => ({
        example: ex,
        overlap: ex.tags.filter(t => tags.includes(t)).length
      }))
      .sort((a, b) => b.overlap - a.overlap)
      .slice(0, count)
      .map(r => r.example);
  }

  // Format for prompt injection
  formatExamples(examples: Example[]): string {
    return examples
      .map((ex, i) => `Example ${i + 1}:\nInput: ${ex.input}\nOutput: ${ex.output}`)
      .join('\n\n');
  }
}

// Usage
const selector = new FewShotSelector([
  { input: "refund my order", output: '{"intent":"refund","entity":"order"}', tags: ["billing", "action"] },
  { input: "how do I export CSV?", output: '{"intent":"how_to","entity":"export"}', tags: ["technical", "question"] },
  { input: "your app crashes on iOS", output: '{"intent":"bug_report","entity":"mobile"}', tags: ["technical", "bug"] },
  { input: "cancel subscription", output: '{"intent":"cancel","entity":"subscription"}', tags: ["billing", "action"] },
  { input: "add dark mode please", output: '{"intent":"feature_request","entity":"ui"}', tags: ["feature", "request"] },
]);

const relevant = selector.selectByTags(["billing", "action"], 2);
const exampleBlock = selector.formatExamples(relevant);
```

## Chain of Thought (CoT) Patterns

```typescript
// CoT with structured reasoning extraction
const cotPrompt = `
Analyze this code for security vulnerabilities.

Think step by step:
1. Identify all user inputs
2. Trace each input through the code
3. Check if any input reaches a sensitive operation unsanitized
4. Classify the severity of each finding

<code>
${codeToReview}
</code>

Respond in this exact format:

<reasoning>
[Your step-by-step analysis here]
</reasoning>

<findings>
[
  {"vulnerability": "...", "line": N, "severity": "critical|high|medium|low", "fix": "..."}
]
</findings>
`;

// Extract structured findings from CoT response
function parseCoTResponse(response: string) {
  const reasoningMatch = response.match(/<reasoning>([\s\S]*?)<\/reasoning>/);
  const findingsMatch = response.match(/<findings>([\s\S]*?)<\/findings>/);

  return {
    reasoning: reasoningMatch?.[1]?.trim() ?? '',
    findings: findingsMatch ? JSON.parse(findingsMatch[1].trim()) : [],
  };
}
```

## Structured Output with Validation

```typescript
import Anthropic from '@anthropic-ai/sdk';
import { z } from 'zod';

// Define schema
const ProductReviewSchema = z.object({
  sentiment: z.enum(['positive', 'negative', 'neutral', 'mixed']),
  score: z.number().min(0).max(10),
  themes: z.array(z.string()).min(1).max(5),
  summary: z.string().max(200),
  actionItems: z.array(z.object({
    department: z.string(),
    action: z.string(),
    priority: z.enum(['low', 'medium', 'high']),
  })),
});

type ProductReview = z.infer<typeof ProductReviewSchema>;

async function analyzeReview(
  client: Anthropic,
  reviewText: string
): Promise<ProductReview> {
  const response = await client.messages.create({
    model: 'claude-sonnet-4-6',
    max_tokens: 1024,
    messages: [{
      role: 'user',
      content: `Analyze this product review and respond with ONLY valid JSON matching this schema:

${JSON.stringify(zodToJsonSchema(ProductReviewSchema), null, 2)}

Review: "${reviewText}"`
    }],
  });

  const text = response.content[0].type === 'text' ? response.content[0].text : '';

  // Extract JSON from potential markdown code blocks
  const jsonStr = text.replace(/```json\n?|\n?```/g, '').trim();

  // Validate against schema
  const parsed = ProductReviewSchema.parse(JSON.parse(jsonStr));
  return parsed;
}

// Helper: Zod to JSON Schema (simplified)
function zodToJsonSchema(schema: z.ZodType): object {
  // Use a library like zod-to-json-schema in production
  return { type: 'object', description: 'See Zod schema for details' };
}
```

## Prompt Chaining

```typescript
// Multi-step pipeline with intermediate validation
interface PipelineStep<TInput, TOutput> {
  name: string;
  prompt: (input: TInput) => string;
  parse: (response: string) => TOutput;
  validate?: (output: TOutput) => boolean;
  model?: string;
}

class PromptPipeline {
  private client: Anthropic;

  constructor(client: Anthropic) {
    this.client = client;
  }

  async run<T>(steps: PipelineStep<any, any>[], initialInput: T): Promise<any> {
    let currentInput = initialInput;

    for (const step of steps) {
      console.log(`Running step: ${step.name}`);

      const response = await this.client.messages.create({
        model: step.model ?? 'claude-sonnet-4-6',
        max_tokens: 2048,
        messages: [{ role: 'user', content: step.prompt(currentInput) }],
      });

      const text = response.content[0].type === 'text' ? response.content[0].text : '';
      const parsed = step.parse(text);

      if (step.validate && !step.validate(parsed)) {
        throw new Error(`Validation failed at step: ${step.name}`);
      }

      currentInput = parsed;
    }

    return currentInput;
  }
}

// Example: Document → Summary → Key Points → Action Items
const pipeline = new PromptPipeline(client);
const result = await pipeline.run([
  {
    name: 'summarize',
    prompt: (doc: string) => `Summarize this document in 3 paragraphs:\n\n${doc}`,
    parse: (text) => text.trim(),
  },
  {
    name: 'extract-key-points',
    prompt: (summary: string) => `Extract 5 key points from this summary as a JSON array of strings:\n\n${summary}`,
    parse: (text) => JSON.parse(text.replace(/```json\n?|\n?```/g, '')),
    validate: (points) => Array.isArray(points) && points.length === 5,
  },
  {
    name: 'generate-actions',
    prompt: (points: string[]) =>
      `For each key point, suggest one specific action item. Respond as JSON array of {point, action, owner} objects:\n\n${points.map((p, i) => `${i+1}. ${p}`).join('\n')}`,
    parse: (text) => JSON.parse(text.replace(/```json\n?|\n?```/g, '')),
    model: 'claude-haiku-4-5-20251001', // cheaper model for simple formatting
  },
], documentText);
```

## Token Optimization

```typescript
// Prompt caching for repeated prefixes
const cachedResponse = await client.messages.create({
  model: 'claude-sonnet-4-6',
  max_tokens: 1024,
  system: [
    {
      type: 'text',
      text: longSystemPrompt,  // 1000+ tokens
      cache_control: { type: 'ephemeral' }  // Cache this block
    }
  ],
  messages: [{ role: 'user', content: userQuery }],
});
// First call: full price. Subsequent calls with same prefix: 90% cheaper.

// Token estimation (rough heuristic)
function estimateTokens(text: string): number {
  // ~4 chars per token for English, ~2-3 for code
  const isCode = text.includes('function') || text.includes('class') || text.includes('{');
  const charsPerToken = isCode ? 2.5 : 4;
  return Math.ceil(text.length / charsPerToken);
}

// Cost-aware model selection
function selectModel(inputTokens: number, taskComplexity: 'simple' | 'medium' | 'complex') {
  if (taskComplexity === 'simple' && inputTokens < 1000) return 'claude-haiku-4-5-20251001';
  if (taskComplexity === 'medium') return 'claude-sonnet-4-6';
  return 'claude-opus-4-6';
}
```

## Gotchas

1. **Prompt injection via user input** — never concatenate user input directly into system prompts. Use XML tags to delineate: `<user_input>${input}</user_input>` and instruct the model to treat content within tags as data, not instructions.

2. **JSON output with markdown wrapping** — models often wrap JSON in ` ```json ... ``` ` code blocks. Always strip markdown fences before JSON.parse(): `text.replace(/```json\n?|\n?```/g, '').trim()`.

3. **Temperature 0 doesn't mean deterministic** — the API uses sampling even at temp=0 (nucleus sampling, hardware non-determinism). For reproducibility, also set top_p and use seed parameter if available.

4. **Few-shot examples bias the output distribution** — if all examples show short responses, the model will produce short responses even when longer ones are appropriate. Include diverse example lengths.

5. **System prompt changes invalidate prompt cache** — even a one-character change to the cached prefix means full-price processing. Version your system prompts and avoid dynamic content in cached blocks.

6. **Chain-of-thought increases cost 3-5x** — the reasoning tokens count toward output billing. Use CoT only for tasks that actually require reasoning (math, logic, multi-step analysis), not for simple extraction or classification.
