---
name: prompt-chaining
description: >
  Multi-step prompt chains, tool use, structured output, and guardrails.
  Use when building multi-step LLM workflows, implementing tool calling,
  enforcing output schemas, or adding safety guardrails to AI pipelines.
  Triggers: "prompt chaining", "multi-step prompt", "tool use LLM",
  "structured output", "guardrails", "LLM pipeline", "chain of thought",
  "function calling", "prompt workflow", "AI pipeline".
  NOT for: single-turn prompts (see prompt-patterns), evaluation metrics (see llm-evaluation).
version: 1.0.0
allowed-tools: Read, Grep, Glob, Bash
---

# Prompt Chaining Patterns

## Sequential Chain

```typescript
// lib/chains.ts — Multi-step LLM pipeline

interface ChainStep<TInput, TOutput> {
  name: string;
  prompt: (input: TInput) => string;
  parse: (response: string) => TOutput;
  validate?: (output: TOutput) => boolean;
  maxRetries?: number;
}

class PromptChain {
  private steps: ChainStep<unknown, unknown>[] = [];
  private logger: (step: string, data: unknown) => void = () => {};

  addStep<TIn, TOut>(step: ChainStep<TIn, TOut>): this {
    this.steps.push(step as ChainStep<unknown, unknown>);
    return this;
  }

  onLog(fn: (step: string, data: unknown) => void): this {
    this.logger = fn;
    return this;
  }

  async execute(initialInput: unknown, llm: LLMClient): Promise<unknown> {
    let currentInput = initialInput;

    for (const step of this.steps) {
      this.logger(step.name, { input: currentInput });

      let output: unknown;
      let attempts = 0;
      const maxRetries = step.maxRetries ?? 2;

      while (attempts <= maxRetries) {
        const prompt = step.prompt(currentInput);
        const response = await llm.complete(prompt);
        output = step.parse(response);

        if (!step.validate || step.validate(output)) break;

        attempts++;
        if (attempts > maxRetries) {
          throw new Error(`Step "${step.name}" failed validation after ${maxRetries} retries`);
        }
        this.logger(step.name, { retry: attempts, reason: 'validation failed' });
      }

      this.logger(step.name, { output });
      currentInput = output;
    }

    return currentInput;
  }
}

// Example: Research → Outline → Draft → Edit chain
const articleChain = new PromptChain()
  .addStep({
    name: 'research',
    prompt: (topic: string) => `Research the topic "${topic}". List 5-7 key points with sources.`,
    parse: (response) => response.split('\n').filter(line => line.trim().startsWith('-')),
    validate: (points) => (points as string[]).length >= 3,
  })
  .addStep({
    name: 'outline',
    prompt: (points: string[]) =>
      `Create a blog post outline from these research points:\n${points.join('\n')}\n\nFormat as JSON: {"title": "...", "sections": [{"heading": "...", "keyPoints": ["..."]}]}`,
    parse: (response) => JSON.parse(response),
    validate: (outline: any) => outline.sections?.length >= 3,
  })
  .addStep({
    name: 'draft',
    prompt: (outline: any) =>
      `Write a 1000-word blog post following this outline:\n${JSON.stringify(outline, null, 2)}\n\nWrite in a conversational, authoritative tone.`,
    parse: (response) => response,
    validate: (draft: string) => draft.split(/\s+/).length > 500,
  })
  .addStep({
    name: 'edit',
    prompt: (draft: string) =>
      `Edit this blog post for clarity, grammar, and engagement. Fix issues but keep the author's voice:\n\n${draft}`,
    parse: (response) => response,
  })
  .onLog((step, data) => console.log(`[${step}]`, JSON.stringify(data).slice(0, 200)));
```

## Tool Use Pattern

```typescript
// lib/tool-use.ts — LLM with callable tools

interface Tool {
  name: string;
  description: string;
  parameters: Record<string, { type: string; description: string; required?: boolean }>;
  execute: (args: Record<string, unknown>) => Promise<string>;
}

const tools: Tool[] = [
  {
    name: 'search_web',
    description: 'Search the web for current information',
    parameters: {
      query: { type: 'string', description: 'Search query', required: true },
      limit: { type: 'number', description: 'Max results (default 5)' },
    },
    execute: async (args) => {
      const results = await searchAPI(args.query as string, (args.limit as number) ?? 5);
      return JSON.stringify(results);
    },
  },
  {
    name: 'calculate',
    description: 'Perform mathematical calculations',
    parameters: {
      expression: { type: 'string', description: 'Math expression to evaluate', required: true },
    },
    execute: async (args) => {
      // Sandboxed math evaluation
      const result = Function(`"use strict"; return (${args.expression})`)();
      return String(result);
    },
  },
  {
    name: 'query_database',
    description: 'Run a read-only SQL query against the analytics database',
    parameters: {
      sql: { type: 'string', description: 'SQL query (SELECT only)', required: true },
    },
    execute: async (args) => {
      const sql = args.sql as string;
      if (!/^\s*SELECT/i.test(sql)) throw new Error('Only SELECT queries allowed');
      const rows = await db.query(sql);
      return JSON.stringify(rows.slice(0, 50)); // Limit response size
    },
  },
];

// Agentic loop: LLM decides which tools to call
async function agentLoop(
  userMessage: string,
  llm: LLMClient,
  availableTools: Tool[],
  maxIterations = 10,
): Promise<string> {
  const messages: Message[] = [
    { role: 'system', content: buildSystemPrompt(availableTools) },
    { role: 'user', content: userMessage },
  ];

  for (let i = 0; i < maxIterations; i++) {
    const response = await llm.chat(messages, {
      tools: availableTools.map(t => ({
        name: t.name,
        description: t.description,
        parameters: t.parameters,
      })),
    });

    // If model returns a final text response, we're done
    if (response.type === 'text') return response.content;

    // If model wants to use a tool
    if (response.type === 'tool_use') {
      const tool = availableTools.find(t => t.name === response.toolName);
      if (!tool) throw new Error(`Unknown tool: ${response.toolName}`);

      try {
        const result = await tool.execute(response.arguments);
        messages.push(
          { role: 'assistant', content: response.raw },
          { role: 'tool', content: result, toolCallId: response.toolCallId },
        );
      } catch (error) {
        messages.push(
          { role: 'assistant', content: response.raw },
          { role: 'tool', content: `Error: ${(error as Error).message}`, toolCallId: response.toolCallId },
        );
      }
    }
  }

  throw new Error('Agent exceeded maximum iterations');
}
```

## Structured Output with Validation

```typescript
// lib/structured-output.ts
import { z } from 'zod';

async function getStructuredOutput<T extends z.ZodType>(
  prompt: string,
  schema: T,
  llm: LLMClient,
  options: { maxRetries?: number; temperature?: number } = {},
): Promise<z.infer<T>> {
  const schemaDescription = JSON.stringify(zodToJsonSchema(schema), null, 2);
  const maxRetries = options.maxRetries ?? 3;

  const systemPrompt = `You must respond with valid JSON matching this exact schema:\n${schemaDescription}\n\nRespond with ONLY the JSON object, no markdown formatting, no explanation.`;

  for (let attempt = 0; attempt < maxRetries; attempt++) {
    const response = await llm.complete(`${systemPrompt}\n\n${prompt}`, {
      temperature: options.temperature ?? 0,
    });

    try {
      // Clean response: strip markdown code fences if present
      const cleaned = response
        .replace(/^```json?\n?/m, '')
        .replace(/\n?```$/m, '')
        .trim();

      const parsed = JSON.parse(cleaned);
      const validated = schema.parse(parsed);
      return validated;
    } catch (error) {
      if (attempt === maxRetries - 1) {
        throw new Error(`Failed to get valid structured output after ${maxRetries} attempts: ${(error as Error).message}`);
      }
      // Retry with error feedback
      prompt += `\n\nYour previous response had this error: ${(error as Error).message}. Please fix and try again.`;
    }
  }

  throw new Error('Unreachable');
}

// Usage:
const ProductReview = z.object({
  sentiment: z.enum(['positive', 'negative', 'neutral']),
  score: z.number().min(0).max(1),
  keyPhrases: z.array(z.string()).min(1).max(5),
  summary: z.string().max(200),
  recommendation: z.boolean(),
});

// const review = await getStructuredOutput(
//   `Analyze this product review: "${reviewText}"`,
//   ProductReview,
//   llm,
// );
```

## Guardrails Pattern

```typescript
// lib/guardrails.ts — Input/output safety checks

interface GuardrailResult {
  passed: boolean;
  violations: string[];
}

type GuardrailCheck = (text: string) => GuardrailResult | Promise<GuardrailResult>;

class GuardrailPipeline {
  private inputChecks: GuardrailCheck[] = [];
  private outputChecks: GuardrailCheck[] = [];

  addInputCheck(check: GuardrailCheck): this {
    this.inputChecks.push(check);
    return this;
  }

  addOutputCheck(check: GuardrailCheck): this {
    this.outputChecks.push(check);
    return this;
  }

  async checkInput(text: string): Promise<GuardrailResult> {
    return this.runChecks(this.inputChecks, text);
  }

  async checkOutput(text: string): Promise<GuardrailResult> {
    return this.runChecks(this.outputChecks, text);
  }

  private async runChecks(checks: GuardrailCheck[], text: string): Promise<GuardrailResult> {
    const allViolations: string[] = [];
    for (const check of checks) {
      const result = await check(text);
      if (!result.passed) allViolations.push(...result.violations);
    }
    return { passed: allViolations.length === 0, violations: allViolations };
  }
}

// Built-in checks
const piiDetector: GuardrailCheck = (text) => {
  const patterns = [
    { regex: /\b\d{3}-\d{2}-\d{4}\b/, type: 'SSN' },
    { regex: /\b\d{16}\b/, type: 'credit card' },
    { regex: /\b[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}\b/i, type: 'email' },
    { regex: /\b\d{3}[-.]?\d{3}[-.]?\d{4}\b/, type: 'phone number' },
  ];
  const violations = patterns
    .filter(p => p.regex.test(text))
    .map(p => `Contains ${p.type}`);
  return { passed: violations.length === 0, violations };
};

const topicBlocker = (blockedTopics: string[]): GuardrailCheck => (text) => {
  const lower = text.toLowerCase();
  const violations = blockedTopics
    .filter(topic => lower.includes(topic.toLowerCase()))
    .map(topic => `Contains blocked topic: ${topic}`);
  return { passed: violations.length === 0, violations };
};

const maxLengthCheck = (maxChars: number): GuardrailCheck => (text) => ({
  passed: text.length <= maxChars,
  violations: text.length > maxChars ? [`Output exceeds ${maxChars} characters (${text.length})`] : [],
});

// Usage:
const guardrails = new GuardrailPipeline()
  .addInputCheck(piiDetector)
  .addInputCheck(topicBlocker(['violence', 'illegal activities']))
  .addOutputCheck(piiDetector)
  .addOutputCheck(maxLengthCheck(5000));

// const inputResult = await guardrails.checkInput(userMessage);
// if (!inputResult.passed) throw new Error(`Input blocked: ${inputResult.violations.join(', ')}`);
// const response = await llm.complete(userMessage);
// const outputResult = await guardrails.checkOutput(response);
// if (!outputResult.passed) return 'I cannot provide that information.';
```

## Gotchas

1. **Chain error propagation** -- One step failing cascades through the entire chain. Each step should have independent error handling and retry logic. Don't let a parsing failure in step 2 crash the entire 5-step pipeline. Catch, log, and either retry or return a meaningful error for that step.

2. **Token budget across chains** -- A 4-step chain with a 4K-token response per step uses 16K output tokens. Long chains can exceed budget limits without any single step being too large. Track cumulative token usage across the chain and fail early if approaching the budget.

3. **Tool use infinite loops** -- An agent that keeps calling the same tool with slightly different parameters can loop forever. Set `maxIterations` and detect repeated tool calls with identical or near-identical arguments. Break the loop and ask the model to synthesize from what it has.

4. **Structured output markdown wrapping** -- Models frequently wrap JSON output in markdown code fences even when told not to. Always strip leading/trailing code fences before JSON.parse(). The pattern `` /^```json?\n?/m `` covers the common cases.

5. **Guardrail false positives** -- Simple regex PII detection flags valid content (e.g., "call 555-0123" in fiction, or "$12.99" matching number patterns). Layer guardrails: fast regex for obvious cases, LLM-based classification for ambiguous ones. Always let humans override guardrail blocks.

6. **Temperature mismatch in chains** -- Creative steps (drafting) need temperature 0.7-1.0. Analytical steps (classification, extraction) need temperature 0-0.2. Using the same temperature for all steps produces either boring creative output or unreliable analytical output. Set temperature per step.
