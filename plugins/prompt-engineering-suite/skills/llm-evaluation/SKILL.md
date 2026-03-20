---
name: llm-evaluation
description: >
  LLM evaluation frameworks — automated evals, LLM-as-judge, benchmark
  design, A/B testing prompts, regression detection, and production
  monitoring for AI applications.
  Triggers: "LLM evaluation", "prompt eval", "AI testing", "LLM benchmark",
  "prompt regression", "LLM-as-judge", "AI quality".
  NOT for: prompt writing (use prompt-patterns), RAG evaluation (use rag-patterns).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# LLM Evaluation

## Evaluation Framework

### Test Case Structure

```typescript
interface EvalCase {
  id: string;
  input: string;                    // The prompt/query
  expectedOutput?: string;          // Ground truth (if available)
  context?: string;                 // Additional context given to the LLM
  metadata: {
    category: string;               // "classification", "generation", etc.
    difficulty: "easy" | "medium" | "hard";
    tags: string[];
  };
  assertions: Assertion[];          // What to check
}

type Assertion =
  | { type: "contains"; value: string }
  | { type: "not_contains"; value: string }
  | { type: "json_schema"; schema: object }
  | { type: "regex"; pattern: string }
  | { type: "length"; min?: number; max?: number }
  | { type: "llm_judge"; criteria: string; threshold: number }
  | { type: "similarity"; reference: string; threshold: number }
  | { type: "custom"; fn: (output: string) => boolean };
```

### Eval Runner

```typescript
interface EvalResult {
  caseId: string;
  passed: boolean;
  score: number;          // 0-1
  latencyMs: number;
  inputTokens: number;
  outputTokens: number;
  costUsd: number;
  output: string;
  assertions: { type: string; passed: boolean; detail?: string }[];
}

async function runEvals(
  cases: EvalCase[],
  config: {
    model: string;
    systemPrompt: string;
    temperature: number;
    maxTokens: number;
  }
): Promise<{ results: EvalResult[]; summary: EvalSummary }> {
  const results: EvalResult[] = [];

  for (const evalCase of cases) {
    const start = Date.now();

    const response = await llm.chat({
      model: config.model,
      messages: [
        { role: "system", content: config.systemPrompt },
        { role: "user", content: evalCase.input },
      ],
      temperature: config.temperature,
      max_tokens: config.maxTokens,
    });

    const latencyMs = Date.now() - start;
    const output = response.content;

    // Run assertions
    const assertionResults = await Promise.all(
      evalCase.assertions.map((a) => checkAssertion(a, output, evalCase))
    );

    const passed = assertionResults.every((a) => a.passed);
    const score = assertionResults.filter((a) => a.passed).length / assertionResults.length;

    results.push({
      caseId: evalCase.id,
      passed,
      score,
      latencyMs,
      inputTokens: response.usage?.input_tokens ?? 0,
      outputTokens: response.usage?.output_tokens ?? 0,
      costUsd: calculateCost(response.usage, config.model),
      output,
      assertions: assertionResults,
    });
  }

  return {
    results,
    summary: summarizeResults(results),
  };
}

async function checkAssertion(
  assertion: Assertion,
  output: string,
  evalCase: EvalCase
): Promise<{ type: string; passed: boolean; detail?: string }> {
  switch (assertion.type) {
    case "contains":
      return {
        type: "contains",
        passed: output.toLowerCase().includes(assertion.value.toLowerCase()),
        detail: `Looking for "${assertion.value}"`,
      };

    case "not_contains":
      return {
        type: "not_contains",
        passed: !output.toLowerCase().includes(assertion.value.toLowerCase()),
        detail: `Should not contain "${assertion.value}"`,
      };

    case "json_schema": {
      try {
        const json = JSON.parse(output.match(/\{[\s\S]*\}/)?.[0] ?? "");
        // Use ajv or zod to validate against schema
        return { type: "json_schema", passed: true };
      } catch {
        return { type: "json_schema", passed: false, detail: "Invalid JSON" };
      }
    }

    case "regex":
      return {
        type: "regex",
        passed: new RegExp(assertion.pattern).test(output),
        detail: `Pattern: ${assertion.pattern}`,
      };

    case "length":
      const len = output.length;
      const minOk = !assertion.min || len >= assertion.min;
      const maxOk = !assertion.max || len <= assertion.max;
      return {
        type: "length",
        passed: minOk && maxOk,
        detail: `Length: ${len} (min: ${assertion.min}, max: ${assertion.max})`,
      };

    case "llm_judge":
      return await llmJudge(output, evalCase, assertion.criteria, assertion.threshold);

    case "similarity":
      const sim = await computeSimilarity(output, assertion.reference);
      return {
        type: "similarity",
        passed: sim >= assertion.threshold,
        detail: `Similarity: ${sim.toFixed(3)} (threshold: ${assertion.threshold})`,
      };

    default:
      return { type: "unknown", passed: false };
  }
}
```

## LLM-as-Judge

### Pairwise Comparison

```typescript
async function pairwiseJudge(
  query: string,
  responseA: string,
  responseB: string,
  criteria: string
): Promise<{ winner: "A" | "B" | "tie"; reasoning: string }> {
  // Randomize order to avoid position bias
  const flip = Math.random() > 0.5;
  const first = flip ? responseB : responseA;
  const second = flip ? responseA : responseB;

  const judgment = await llm.chat({
    model: "claude-sonnet-4-6-20250514",
    messages: [{
      role: "user",
      content: `Compare these two responses to the query below.

Evaluation criteria: ${criteria}

Query: ${query}

Response 1:
${first}

Response 2:
${second}

Which response is better? Consider: accuracy, completeness, clarity, and relevance.

Respond in this exact format:
Winner: [1 or 2 or tie]
Reasoning: [2-3 sentences explaining why]`,
    }],
    temperature: 0,
  });

  const text = judgment.content;
  const winnerMatch = text.match(/Winner:\s*(1|2|tie)/i);
  let winner: "A" | "B" | "tie" = "tie";

  if (winnerMatch) {
    const raw = winnerMatch[1];
    if (raw === "1") winner = flip ? "B" : "A";
    else if (raw === "2") winner = flip ? "A" : "B";
  }

  return { winner, reasoning: text };
}
```

### Rubric-Based Scoring

```typescript
async function rubricScore(
  output: string,
  rubric: { criterion: string; weight: number }[]
): Promise<{ total: number; scores: { criterion: string; score: number; feedback: string }[] }> {
  const scores = await Promise.all(
    rubric.map(async ({ criterion, weight }) => {
      const response = await llm.chat({
        model: "claude-haiku-4-5-20251001", // Fast model for each criterion
        messages: [{
          role: "user",
          content: `Rate the following output on this criterion: "${criterion}"

Output:
${output.slice(0, 2000)}

Score (1-5):
1 = Completely fails this criterion
2 = Mostly fails, some elements present
3 = Partially meets criterion
4 = Mostly meets criterion
5 = Fully meets or exceeds criterion

Respond with ONLY a number (1-5) and one sentence of feedback.
Format: [score] — [feedback]`,
        }],
        temperature: 0,
        max_tokens: 100,
      });

      const match = response.content.match(/(\d)\s*[-—]\s*(.*)/);
      const score = match ? parseInt(match[1]) : 3;
      const feedback = match?.[2] ?? response.content;

      return { criterion, score: (score / 5) * weight, feedback };
    })
  );

  const total = scores.reduce((sum, s) => sum + s.score, 0) /
    rubric.reduce((sum, r) => sum + r.weight, 0);

  return { total, scores };
}
```

## A/B Testing Prompts

```typescript
interface PromptVariant {
  id: string;
  systemPrompt: string;
  model?: string;
  temperature?: number;
}

interface ABTestResult {
  variantId: string;
  avgScore: number;
  avgLatency: number;
  avgCost: number;
  passRate: number;
  sampleSize: number;
}

async function abTestPrompts(
  variants: PromptVariant[],
  evalCases: EvalCase[],
  runsPerVariant: number = 1
): Promise<ABTestResult[]> {
  const results: ABTestResult[] = [];

  for (const variant of variants) {
    const variantResults: EvalResult[] = [];

    for (let run = 0; run < runsPerVariant; run++) {
      const { results: evalResults } = await runEvals(evalCases, {
        model: variant.model ?? "claude-sonnet-4-6-20250514",
        systemPrompt: variant.systemPrompt,
        temperature: variant.temperature ?? 0,
        maxTokens: 4096,
      });
      variantResults.push(...evalResults);
    }

    results.push({
      variantId: variant.id,
      avgScore: avg(variantResults.map((r) => r.score)),
      avgLatency: avg(variantResults.map((r) => r.latencyMs)),
      avgCost: sum(variantResults.map((r) => r.costUsd)) / runsPerVariant,
      passRate: variantResults.filter((r) => r.passed).length / variantResults.length,
      sampleSize: variantResults.length,
    });
  }

  return results.sort((a, b) => b.avgScore - a.avgScore);
}
```

## Regression Detection

```typescript
interface Baseline {
  promptVersion: string;
  modelVersion: string;
  date: string;
  results: EvalResult[];
  metrics: {
    avgScore: number;
    passRate: number;
    avgLatency: number;
    avgCost: number;
  };
}

async function detectRegression(
  current: EvalResult[],
  baseline: Baseline,
  thresholds: {
    scoreDropPct: number;     // Max acceptable score decrease (e.g., 5%)
    passRateDropPct: number;  // Max acceptable pass rate decrease
    latencyIncreasePct: number;
  } = { scoreDropPct: 5, passRateDropPct: 10, latencyIncreasePct: 50 }
): Promise<{ regressed: boolean; issues: string[] }> {
  const issues: string[] = [];

  const currentMetrics = {
    avgScore: avg(current.map((r) => r.score)),
    passRate: current.filter((r) => r.passed).length / current.length,
    avgLatency: avg(current.map((r) => r.latencyMs)),
  };

  // Score regression
  const scoreDrop = ((baseline.metrics.avgScore - currentMetrics.avgScore) / baseline.metrics.avgScore) * 100;
  if (scoreDrop > thresholds.scoreDropPct) {
    issues.push(`Score dropped ${scoreDrop.toFixed(1)}% (${baseline.metrics.avgScore.toFixed(3)} → ${currentMetrics.avgScore.toFixed(3)})`);
  }

  // Pass rate regression
  const passRateDrop = ((baseline.metrics.passRate - currentMetrics.passRate) / baseline.metrics.passRate) * 100;
  if (passRateDrop > thresholds.passRateDropPct) {
    issues.push(`Pass rate dropped ${passRateDrop.toFixed(1)}% (${(baseline.metrics.passRate * 100).toFixed(1)}% → ${(currentMetrics.passRate * 100).toFixed(1)}%)`);
  }

  // Latency regression
  const latencyIncrease = ((currentMetrics.avgLatency - baseline.metrics.avgLatency) / baseline.metrics.avgLatency) * 100;
  if (latencyIncrease > thresholds.latencyIncreasePct) {
    issues.push(`Latency increased ${latencyIncrease.toFixed(1)}% (${baseline.metrics.avgLatency.toFixed(0)}ms → ${currentMetrics.avgLatency.toFixed(0)}ms)`);
  }

  // Per-case regression (find specific failures)
  for (const current_result of current) {
    const baselineResult = baseline.results.find((r) => r.caseId === current_result.caseId);
    if (baselineResult?.passed && !current_result.passed) {
      issues.push(`Case "${current_result.caseId}" newly failing (was passing in baseline)`);
    }
  }

  return { regressed: issues.length > 0, issues };
}
```

## Production Monitoring

```typescript
// Log every LLM call for monitoring
interface LLMCallLog {
  timestamp: string;
  requestId: string;
  model: string;
  inputTokens: number;
  outputTokens: number;
  latencyMs: number;
  costUsd: number;
  temperature: number;
  promptVersion: string;
  userSegment?: string;
  passed?: boolean;        // If inline eval ran
  error?: string;
}

// Metrics to dashboard
interface LLMDashboardMetrics {
  requestsPerMinute: number;
  avgLatencyMs: number;
  p99LatencyMs: number;
  errorRate: number;
  costPerHour: number;
  tokenUsagePerRequest: { input: number; output: number };
  modelDistribution: Record<string, number>;
  topErrors: { message: string; count: number }[];
}

// Inline quality sampling (evaluate N% of production requests)
async function sampleAndEval(
  input: string,
  output: string,
  sampleRate: number = 0.05 // 5% of requests
): Promise<void> {
  if (Math.random() > sampleRate) return;

  const score = await rubricScore(output, [
    { criterion: "Accuracy and factual correctness", weight: 3 },
    { criterion: "Relevance to the question", weight: 2 },
    { criterion: "Clarity and readability", weight: 1 },
  ]);

  if (score.total < 0.6) {
    // Alert on low-quality responses
    await alertSlack(`Low quality response detected (score: ${score.total.toFixed(2)})\nInput: ${input.slice(0, 200)}`);
  }

  await logMetric("llm_quality_score", score.total);
}
```

## Gotchas

1. **LLM judges have position bias** — The first response in a pairwise comparison is preferred ~60% of the time. Randomize order and average both orderings.

2. **Small eval sets mislead** — 20 test cases aren't enough. Aim for 100+ cases covering edge cases, adversarial inputs, and all categories.

3. **Temperature 0 isn't deterministic** — LLMs with temperature 0 can still produce different outputs across API calls (batching, hardware, etc.). Run each eval case 3+ times and average.

4. **Cost of evals adds up** — Running 100 eval cases through Opus costs ~$5-10. Use Haiku for judges when possible. Cache embeddings for similarity assertions.

5. **Ground truth goes stale** — Expected outputs need updating as products/features change. Review ground truth quarterly.

6. **Don't eval what you don't measure in production** — If your eval suite tests summarization quality but production users care about response speed, your evals are measuring the wrong thing. Align eval criteria with user satisfaction metrics.
