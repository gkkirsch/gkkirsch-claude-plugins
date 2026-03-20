---
name: statistical-analysis
description: >
  Statistical analysis patterns for product analytics and A/B testing.
  Use when implementing metrics calculations, A/B test analysis, cohort
  analysis, funnel analysis, or statistical significance testing.
  Triggers: "A/B test", "statistical significance", "p-value", "cohort",
  "funnel analysis", "retention", "conversion rate", "metrics".
  NOT for: ML model training, deep learning, or data visualization (see visualization-patterns).
version: 1.0.0
allowed-tools: Read, Grep, Glob, Bash
---

# Statistical Analysis Patterns

## A/B Test Analysis

```typescript
// Z-test for conversion rate comparison
interface ABTestResult {
  control: { visitors: number; conversions: number };
  variant: { visitors: number; conversions: number };
  pValue: number;
  significant: boolean;
  lift: number;
  confidenceInterval: [number, number];
  requiredSampleSize: number;
}

function analyzeABTest(
  control: { visitors: number; conversions: number },
  variant: { visitors: number; conversions: number },
  alpha: number = 0.05
): ABTestResult {
  const p1 = control.conversions / control.visitors;
  const p2 = variant.conversions / variant.visitors;

  // Pooled proportion
  const pPool = (control.conversions + variant.conversions) /
                (control.visitors + variant.visitors);

  // Standard error
  const se = Math.sqrt(
    pPool * (1 - pPool) * (1 / control.visitors + 1 / variant.visitors)
  );

  // Z-score
  const z = (p2 - p1) / se;

  // Two-tailed p-value
  const pValue = 2 * (1 - normalCDF(Math.abs(z)));

  // Confidence interval for the difference
  const seDiff = Math.sqrt(
    (p1 * (1 - p1)) / control.visitors +
    (p2 * (1 - p2)) / variant.visitors
  );
  const zAlpha = normalQuantile(1 - alpha / 2);
  const ci: [number, number] = [
    (p2 - p1) - zAlpha * seDiff,
    (p2 - p1) + zAlpha * seDiff,
  ];

  // Required sample size per group (80% power)
  const mde = 0.02; // minimum detectable effect
  const requiredSampleSize = Math.ceil(
    2 * Math.pow((zAlpha + normalQuantile(0.8)) / mde, 2) * pPool * (1 - pPool)
  );

  return {
    control,
    variant,
    pValue,
    significant: pValue < alpha,
    lift: ((p2 - p1) / p1) * 100,
    confidenceInterval: ci,
    requiredSampleSize,
  };
}

// Normal CDF approximation (Abramowitz and Stegun)
function normalCDF(x: number): number {
  const t = 1 / (1 + 0.2316419 * Math.abs(x));
  const d = 0.3989422804014327;
  const p = d * Math.exp(-x * x / 2) * t *
    (0.319381530 + t * (-0.356563782 + t * (1.781477937 + t * (-1.821255978 + t * 1.330274429))));
  return x >= 0 ? 1 - p : p;
}

// Inverse normal (approximation)
function normalQuantile(p: number): number {
  if (p <= 0 || p >= 1) throw new Error('p must be between 0 and 1');
  if (p === 0.5) return 0;
  const t = Math.sqrt(-2 * Math.log(p < 0.5 ? p : 1 - p));
  const c = [2.515517, 0.802853, 0.010328];
  const d = [1.432788, 0.189269, 0.001308];
  let q = t - (c[0] + t * (c[1] + t * c[2])) / (1 + t * (d[0] + t * (d[1] + t * d[2])));
  return p < 0.5 ? -q : q;
}
```

## Funnel Analysis

```typescript
interface FunnelStep {
  name: string;
  count: number;
  conversionRate: number; // from previous step
  overallRate: number;    // from first step
  dropoff: number;        // absolute drop from previous
}

function analyzeFunnel(steps: Array<{ name: string; count: number }>): FunnelStep[] {
  const firstCount = steps[0]?.count ?? 0;

  return steps.map((step, i) => {
    const prevCount = i === 0 ? step.count : steps[i - 1].count;
    return {
      name: step.name,
      count: step.count,
      conversionRate: i === 0 ? 1 : step.count / prevCount,
      overallRate: step.count / firstCount,
      dropoff: prevCount - step.count,
    };
  });
}

// Example usage
const funnel = analyzeFunnel([
  { name: 'Page Visit',     count: 10000 },
  { name: 'Sign Up Click',  count: 3200 },
  { name: 'Form Submitted', count: 2100 },
  { name: 'Email Verified', count: 1500 },
  { name: 'First Purchase', count: 420 },
]);
// funnel[4].overallRate = 0.042 (4.2% overall conversion)
```

## Cohort Retention Analysis

```typescript
interface CohortData {
  cohort: string;       // e.g., "2026-01"
  cohortSize: number;
  retention: number[];  // retention rate per period
}

function buildCohortTable(
  events: Array<{ userId: string; date: string; event: string }>,
  signupEvent: string,
  retentionEvent: string,
  periodType: 'day' | 'week' | 'month' = 'month'
): CohortData[] {
  // Group users by signup cohort
  const signups = events.filter(e => e.event === signupEvent);
  const cohorts = new Map<string, Set<string>>();

  for (const event of signups) {
    const period = getPeriodKey(event.date, periodType);
    if (!cohorts.has(period)) cohorts.set(period, new Set());
    cohorts.get(period)!.add(event.userId);
  }

  // Calculate retention per period
  const retentionEvents = events.filter(e => e.event === retentionEvent);
  const userActivity = new Map<string, Set<string>>();

  for (const event of retentionEvents) {
    const period = getPeriodKey(event.date, periodType);
    if (!userActivity.has(event.userId)) userActivity.set(event.userId, new Set());
    userActivity.get(event.userId)!.add(period);
  }

  const result: CohortData[] = [];
  const sortedCohorts = [...cohorts.keys()].sort();

  for (const cohortKey of sortedCohorts) {
    const cohortUsers = cohorts.get(cohortKey)!;
    const periods = sortedCohorts.filter(p => p >= cohortKey);

    const retention = periods.map(period => {
      const activeCount = [...cohortUsers].filter(userId =>
        userActivity.get(userId)?.has(period)
      ).length;
      return activeCount / cohortUsers.size;
    });

    result.push({
      cohort: cohortKey,
      cohortSize: cohortUsers.size,
      retention,
    });
  }

  return result;
}

function getPeriodKey(dateStr: string, type: 'day' | 'week' | 'month'): string {
  const date = new Date(dateStr);
  if (type === 'month') return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}`;
  if (type === 'week') {
    const jan1 = new Date(date.getFullYear(), 0, 1);
    const week = Math.ceil(((date.getTime() - jan1.getTime()) / 86400000 + jan1.getDay() + 1) / 7);
    return `${date.getFullYear()}-W${String(week).padStart(2, '0')}`;
  }
  return dateStr.slice(0, 10);
}
```

## Moving Averages and Trend Detection

```typescript
// Simple Moving Average
function sma(data: number[], window: number): (number | null)[] {
  return data.map((_, i) => {
    if (i < window - 1) return null;
    const slice = data.slice(i - window + 1, i + 1);
    return slice.reduce((a, b) => a + b, 0) / window;
  });
}

// Exponential Moving Average
function ema(data: number[], alpha: number = 0.3): number[] {
  const result: number[] = [data[0]];
  for (let i = 1; i < data.length; i++) {
    result.push(alpha * data[i] + (1 - alpha) * result[i - 1]);
  }
  return result;
}

// Linear regression for trend line
function linearRegression(data: number[]): { slope: number; intercept: number; r2: number } {
  const n = data.length;
  const xs = data.map((_, i) => i);

  const sumX = xs.reduce((a, b) => a + b, 0);
  const sumY = data.reduce((a, b) => a + b, 0);
  const sumXY = xs.reduce((sum, x, i) => sum + x * data[i], 0);
  const sumX2 = xs.reduce((sum, x) => sum + x * x, 0);

  const slope = (n * sumXY - sumX * sumY) / (n * sumX2 - sumX * sumX);
  const intercept = (sumY - slope * sumX) / n;

  // R-squared
  const meanY = sumY / n;
  const ssRes = data.reduce((sum, y, i) => sum + Math.pow(y - (slope * i + intercept), 2), 0);
  const ssTot = data.reduce((sum, y) => sum + Math.pow(y - meanY, 2), 0);
  const r2 = 1 - ssRes / ssTot;

  return { slope, intercept, r2 };
}

// Anomaly detection (simple z-score)
function detectAnomalies(data: number[], threshold: number = 2): number[] {
  const mean = data.reduce((a, b) => a + b, 0) / data.length;
  const stdDev = Math.sqrt(
    data.reduce((sum, val) => sum + Math.pow(val - mean, 2), 0) / data.length
  );

  return data
    .map((val, i) => ({ index: i, zscore: Math.abs((val - mean) / stdDev) }))
    .filter(d => d.zscore > threshold)
    .map(d => d.index);
}
```

## Gotchas

1. **Peeking at A/B test results** — checking significance multiple times before the test reaches sample size inflates false positive rate. Use sequential testing (SPRT) or pre-commit to a sample size and only check once at the end.

2. **Simpson's paradox in segmented data** — an overall trend can reverse when data is split by segments (e.g., variant wins overall but loses in every demographic). Always segment by key dimensions before drawing conclusions.

3. **Survivorship bias in cohort analysis** — if you only analyze active users, you miss the ones who churned. Always start cohorts from signup/acquisition date and include zeros for churned users.

4. **Small sample significance** — with < 1000 samples per group, A/B tests have low statistical power. A non-significant result doesn't mean no effect — it means you can't detect one. Calculate required sample size before running the test.

5. **Floating point in financial calculations** — `0.1 + 0.2 !== 0.3` in JavaScript. For financial metrics, use integer cents or a decimal library. Never use raw float multiplication for currency.

6. **Timezone mismatch in daily metrics** — if your database stores UTC and your dashboard shows "daily active users" in local time, the same user can appear in two different days. Always aggregate in a consistent timezone and document which one.
