---
name: chart-accessibility
description: >
  Accessible data visualization patterns for charts and graphs.
  Use when building WCAG-compliant charts, adding screen reader support,
  implementing keyboard navigation, or designing color-blind safe palettes.
  Triggers: "accessible chart", "chart accessibility", "screen reader chart",
  "color blind chart", "ARIA chart", "keyboard navigation chart", "WCAG data viz".
  NOT for: general web accessibility (see WCAG guidelines), decorative graphics, icon design.
version: 1.0.0
allowed-tools: Read, Grep, Glob, Bash
---

# Chart Accessibility

## ARIA Roles & Properties for Charts

```tsx
// components/AccessibleBarChart.tsx
interface ChartData {
  label: string;
  value: number;
  color: string;
}

function AccessibleBarChart({
  data,
  title,
  description,
}: {
  data: ChartData[];
  title: string;
  description: string;
}) {
  const maxValue = Math.max(...data.map(d => d.value));
  const chartId = useId();

  return (
    <figure
      role="img"
      aria-labelledby={`${chartId}-title`}
      aria-describedby={`${chartId}-desc`}
    >
      <figcaption id={`${chartId}-title`}>{title}</figcaption>
      <p id={`${chartId}-desc`} className="sr-only">{description}</p>

      {/* Visual chart */}
      <svg
        viewBox={`0 0 ${data.length * 60} 200`}
        aria-hidden="true" // Hide SVG from screen readers — table is the accessible version
      >
        {data.map((item, i) => (
          <g key={item.label}>
            <rect
              x={i * 60 + 10}
              y={200 - (item.value / maxValue) * 180}
              width={40}
              height={(item.value / maxValue) * 180}
              fill={item.color}
            />
            <text
              x={i * 60 + 30}
              y={195}
              textAnchor="middle"
              fontSize={12}
            >
              {item.label}
            </text>
          </g>
        ))}
      </svg>

      {/* Accessible data table (screen reader alternative) */}
      <table className="sr-only">
        <caption>{title}</caption>
        <thead>
          <tr>
            <th scope="col">Category</th>
            <th scope="col">Value</th>
          </tr>
        </thead>
        <tbody>
          {data.map(item => (
            <tr key={item.label}>
              <td>{item.label}</td>
              <td>{item.value}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </figure>
  );
}
```

## Keyboard-Navigable Interactive Charts

```tsx
// components/KeyboardChart.tsx
function KeyboardChart({ data, title }: { data: ChartData[]; title: string }) {
  const [focusedIndex, setFocusedIndex] = useState(-1);
  const [announcedText, setAnnouncedText] = useState('');
  const chartRef = useRef<HTMLDivElement>(null);

  const handleKeyDown = (e: React.KeyboardEvent) => {
    switch (e.key) {
      case 'ArrowRight':
      case 'ArrowDown':
        e.preventDefault();
        setFocusedIndex(prev => {
          const next = Math.min(prev + 1, data.length - 1);
          announceDataPoint(next);
          return next;
        });
        break;
      case 'ArrowLeft':
      case 'ArrowUp':
        e.preventDefault();
        setFocusedIndex(prev => {
          const next = Math.max(prev - 1, 0);
          announceDataPoint(next);
          return next;
        });
        break;
      case 'Home':
        e.preventDefault();
        setFocusedIndex(0);
        announceDataPoint(0);
        break;
      case 'End':
        e.preventDefault();
        setFocusedIndex(data.length - 1);
        announceDataPoint(data.length - 1);
        break;
      case 'Escape':
        chartRef.current?.blur();
        setFocusedIndex(-1);
        break;
    }
  };

  function announceDataPoint(index: number) {
    const item = data[index];
    const maxVal = Math.max(...data.map(d => d.value));
    const percentage = Math.round((item.value / maxVal) * 100);
    setAnnouncedText(
      `${item.label}: ${item.value}. ${percentage}% of maximum. Item ${index + 1} of ${data.length}.`
    );
  }

  return (
    <div
      ref={chartRef}
      role="application"
      aria-roledescription="interactive chart"
      aria-label={`${title}. Use arrow keys to navigate data points.`}
      tabIndex={0}
      onKeyDown={handleKeyDown}
      onFocus={() => {
        if (focusedIndex === -1) {
          setFocusedIndex(0);
          announceDataPoint(0);
        }
      }}
    >
      {/* Live region for announcements */}
      <div aria-live="assertive" aria-atomic="true" className="sr-only">
        {announcedText}
      </div>

      {/* Chart rendering with focus indicator */}
      <svg viewBox="0 0 400 200" aria-hidden="true">
        {data.map((item, i) => (
          <rect
            key={item.label}
            x={i * 60 + 10}
            y={200 - (item.value / Math.max(...data.map(d => d.value))) * 180}
            width={40}
            height={(item.value / Math.max(...data.map(d => d.value))) * 180}
            fill={item.color}
            stroke={i === focusedIndex ? '#000' : 'none'}
            strokeWidth={i === focusedIndex ? 3 : 0}
          />
        ))}
      </svg>
    </div>
  );
}
```

## Color-Blind Safe Palettes

```typescript
// lib/accessible-colors.ts

// Categorical palettes safe for all color vision deficiencies
// Tested with: protanopia, deuteranopia, tritanopia, achromatopsia
export const colorBlindSafe = {
  // 8-color palette (Okabe & Ito universal design)
  categorical: [
    '#E69F00', // orange
    '#56B4E9', // sky blue
    '#009E73', // bluish green
    '#F0E442', // yellow
    '#0072B2', // blue
    '#D55E00', // vermillion
    '#CC79A7', // reddish purple
    '#000000', // black
  ],

  // Sequential palettes (light → dark)
  sequential: {
    blue: ['#deebf7', '#9ecae1', '#4292c6', '#2171b5', '#084594'],
    orange: ['#feedde', '#fdbe85', '#fd8d3c', '#e6550d', '#a63603'],
  },

  // Diverging palette (negative → neutral → positive)
  diverging: ['#d73027', '#f46d43', '#fdae61', '#fee08b', '#ffffbf',
              '#d9ef8b', '#a6d96a', '#66bd63', '#1a9850'],
};

// NEVER rely on color alone — add patterns, labels, or shapes
export function getPattern(index: number): string {
  const patterns = [
    'solid',        // ████
    'diagonal',     // ////
    'dots',         // ....
    'crosshatch',   // xxxx
    'horizontal',   // ────
    'vertical',     // ||||
    'zigzag',       // ^^^^
    'waves',        // ~~~~
  ];
  return patterns[index % patterns.length];
}

// SVG pattern definitions for chart fills
export function createSVGPatterns(): string {
  return `
    <defs>
      <pattern id="pattern-diagonal" patternUnits="userSpaceOnUse" width="8" height="8">
        <path d="M-2,2 l4,-4 M0,8 l8,-8 M6,10 l4,-4" stroke="currentColor" strokeWidth="1.5"/>
      </pattern>
      <pattern id="pattern-dots" patternUnits="userSpaceOnUse" width="8" height="8">
        <circle cx="4" cy="4" r="2" fill="currentColor"/>
      </pattern>
      <pattern id="pattern-crosshatch" patternUnits="userSpaceOnUse" width="8" height="8">
        <path d="M0,0 l8,8 M8,0 l-8,8" stroke="currentColor" strokeWidth="1"/>
      </pattern>
      <pattern id="pattern-horizontal" patternUnits="userSpaceOnUse" width="8" height="8">
        <path d="M0,4 l8,0" stroke="currentColor" strokeWidth="1.5"/>
      </pattern>
    </defs>
  `;
}

// Contrast checker (WCAG 2.1 AA requires 4.5:1 for text, 3:1 for large text)
export function contrastRatio(hex1: string, hex2: string): number {
  const lum1 = relativeLuminance(hex1);
  const lum2 = relativeLuminance(hex2);
  const lighter = Math.max(lum1, lum2);
  const darker = Math.min(lum1, lum2);
  return (lighter + 0.05) / (darker + 0.05);
}

function relativeLuminance(hex: string): number {
  const rgb = hexToRgb(hex);
  const [r, g, b] = rgb.map(c => {
    c = c / 255;
    return c <= 0.03928 ? c / 12.92 : Math.pow((c + 0.055) / 1.055, 2.4);
  });
  return 0.2126 * r + 0.7152 * g + 0.0722 * b;
}

function hexToRgb(hex: string): number[] {
  const result = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(hex);
  if (!result) throw new Error(`Invalid hex color: ${hex}`);
  return [parseInt(result[1], 16), parseInt(result[2], 16), parseInt(result[3], 16)];
}
```

## Screen Reader Descriptions

```typescript
// lib/chart-descriptions.ts — Generate meaningful text descriptions

interface DataSeries {
  name: string;
  data: { label: string; value: number }[];
  unit: string;
}

// Generate natural language description of chart data
export function describeChart(
  type: 'bar' | 'line' | 'pie',
  series: DataSeries[],
  title: string,
): string {
  const parts: string[] = [`${title}.`];

  for (const s of series) {
    const values = s.data.map(d => d.value);
    const min = Math.min(...values);
    const max = Math.max(...values);
    const avg = Math.round(values.reduce((a, b) => a + b, 0) / values.length);
    const minItem = s.data.find(d => d.value === min)!;
    const maxItem = s.data.find(d => d.value === max)!;

    if (type === 'bar' || type === 'line') {
      parts.push(
        `${s.name}: ranges from ${min} ${s.unit} (${minItem.label}) to ${max} ${s.unit} (${maxItem.label}), averaging ${avg} ${s.unit} across ${s.data.length} categories.`
      );

      // Detect trend for line charts
      if (type === 'line' && s.data.length >= 3) {
        const trend = detectTrend(values);
        parts.push(`The overall trend is ${trend}.`);
      }
    }

    if (type === 'pie') {
      const total = values.reduce((a, b) => a + b, 0);
      const topItems = [...s.data]
        .sort((a, b) => b.value - a.value)
        .slice(0, 3)
        .map(d => `${d.label} (${Math.round(d.value / total * 100)}%)`);
      parts.push(`Top categories: ${topItems.join(', ')}.`);
    }
  }

  return parts.join(' ');
}

function detectTrend(values: number[]): string {
  const first = values.slice(0, Math.ceil(values.length / 3));
  const last = values.slice(-Math.ceil(values.length / 3));
  const firstAvg = first.reduce((a, b) => a + b) / first.length;
  const lastAvg = last.reduce((a, b) => a + b) / last.length;
  const change = ((lastAvg - firstAvg) / firstAvg) * 100;

  if (Math.abs(change) < 5) return 'stable';
  return change > 0 ? `increasing (up ${Math.round(change)}%)` : `decreasing (down ${Math.round(Math.abs(change))}%)`;
}
```

## Gotchas

1. **SVG alone is invisible to screen readers** -- An SVG chart without a text alternative is completely invisible to screen readers. Always provide one of: (a) a hidden data table with the same data, (b) an `aria-label` or `aria-describedby` with a text summary, or (c) `<desc>` elements within the SVG. The hidden data table approach gives the most detail.

2. **Color as the only differentiator** -- WCAG 1.4.1 requires that color is not the only means of conveying information. A legend that says "red = errors, green = success" fails. Add patterns (stripes, dots), shapes (circle, square, triangle), or direct labels on each data series. About 8% of men have some form of color vision deficiency.

3. **Live region spam** -- Using `aria-live="assertive"` on rapidly updating charts (real-time data, animations) floods the screen reader with announcements. Use `aria-live="polite"` for periodic updates, debounce announcements to every 5+ seconds, and give users a way to pause live updates entirely.

4. **Missing keyboard focus indicators** -- Interactive chart elements that receive keyboard focus must have a visible focus indicator (WCAG 2.4.7). A 2-3px solid outline in a contrasting color is the minimum. Don't rely on the browser's default outline — it's often invisible on colored backgrounds.

5. **Tooltip information only on hover** -- If chart tooltips show additional data on mouse hover but not on keyboard focus, keyboard and screen reader users can't access that information. Ensure tooltips appear on both `:hover` and `:focus`, and that the tooltip content is announced via `aria-live` or `aria-describedby`.

6. **Chart animations without reduced-motion support** -- Animated chart transitions (bars growing, lines drawing) cause problems for users with vestibular disorders. Always check `prefers-reduced-motion: reduce` and skip or simplify animations. Use `@media (prefers-reduced-motion: reduce) { * { animation-duration: 0.01ms !important; } }`.
