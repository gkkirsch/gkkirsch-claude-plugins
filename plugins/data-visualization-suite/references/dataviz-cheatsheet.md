# Data Visualization Cheatsheet

## Recharts Quick Reference

```tsx
// Minimal setup
import { LineChart, Line, XAxis, YAxis, ResponsiveContainer } from "recharts";

<ResponsiveContainer width="100%" height={400}>
  <LineChart data={data}>
    <XAxis dataKey="name" />
    <YAxis />
    <Line dataKey="value" stroke="#6366f1" />
  </LineChart>
</ResponsiveContainer>
```

### Chart Types

| Component | Use For |
|-----------|---------|
| `<LineChart>` | Trends over time |
| `<BarChart>` | Category comparison |
| `<AreaChart>` | Volume over time |
| `<PieChart>` | Part of whole (≤7) |
| `<ComposedChart>` | Mixed types |
| `<ScatterChart>` | Correlation |
| `<RadarChart>` | Multi-metric |
| `<RadialBarChart>` | Circular progress |

### Essential Props

```tsx
// Stacking
<Bar dataKey="a" stackId="group1" />
<Bar dataKey="b" stackId="group1" />

// Custom tooltip
<Tooltip content={<CustomTooltip />} />

// Reference line
<ReferenceLine y={target} stroke="red" strokeDasharray="3 3" />

// Brush (zoom)
<Brush dataKey="date" height={30} />

// Format axis
<YAxis tickFormatter={(v) => `$${v / 1000}k`} />

// Responsive (ALWAYS use)
<ResponsiveContainer width="100%" height={400}>
```

## D3 Quick Reference

### Scales

```typescript
// Numeric
d3.scaleLinear().domain([0, max]).range([height, 0])

// Time
d3.scaleTime().domain([startDate, endDate]).range([0, width])

// Categorical (bars)
d3.scaleBand().domain(categories).range([0, width]).padding(0.2)

// Colors
d3.scaleOrdinal(d3.schemeTableau10)
d3.scaleSequential(d3.interpolateBlues).domain([0, max])
```

### Path Generators

```typescript
// Line
const line = d3.line<D>().x(d => xScale(d.x)).y(d => yScale(d.y)).curve(d3.curveMonotoneX);
<path d={line(data)!} fill="none" stroke="#6366f1" />

// Area
const area = d3.area<D>().x(d => xScale(d.x)).y0(height).y1(d => yScale(d.y));
<path d={area(data)!} fill="#6366f1" opacity={0.3} />

// Arc (pie/donut)
const arc = d3.arc().innerRadius(60).outerRadius(120);
const pie = d3.pie<D>().value(d => d.value);
pie(data).map(d => <path d={arc(d)!} />)
```

### Layouts

```typescript
// Treemap
d3.treemap().size([w, h]).padding(2)(root)

// Force
d3.forceSimulation(nodes)
  .force("link", d3.forceLink(links).id(d => d.id))
  .force("charge", d3.forceManyBody().strength(-100))
  .force("center", d3.forceCenter(w/2, h/2))

// Hierarchy
d3.hierarchy(data).sum(d => d.value)

// Pack (circle packing)
d3.pack().size([w, h]).padding(2)(root)
```

### Color Palettes

| Palette | Type | Count | Colorblind? |
|---------|------|-------|-------------|
| `schemeTableau10` | Categorical | 10 | Yes |
| `interpolateViridis` | Sequential | Continuous | Yes |
| `interpolateBlues` | Sequential | Continuous | Yes |
| `interpolateYlOrRd` | Sequential | Continuous | Partial |
| `interpolateRdBu` | Diverging | Continuous | Yes |

## Chart Selection Guide

| Question | Answer |
|----------|--------|
| How does X change over time? | Line chart |
| How do A, B, C compare? | Bar chart (horizontal if labels are long) |
| What's the breakdown of X? | Pie (≤5) or stacked bar (>5) |
| Is there a correlation between X and Y? | Scatter plot |
| What's the distribution of X? | Histogram |
| How does data flow from A to B? | Sankey diagram |
| What's the hierarchy of X? | Treemap or sunburst |
| How are things connected? | Force-directed graph |
| Where is X happening? | Choropleth map |

## Performance Tips

| Dataset Size | Strategy |
|-------------|----------|
| < 1K points | SVG, full animation |
| 1K - 5K points | SVG, reduce animation |
| 5K - 50K points | Canvas (Chart.js) or downsample |
| 50K+ points | WebGL or aggressive aggregation |

## Common Patterns

```tsx
// Tooltip position tracking
const [tooltip, setTooltip] = useState<{x: number; y: number; data: any} | null>(null);
<circle
  onMouseEnter={(e) => setTooltip({ x: e.clientX, y: e.clientY, data: d })}
  onMouseLeave={() => setTooltip(null)}
/>

// Number formatting
d3.format(",.0f")(1234567)     // "1,234,567"
d3.format("$.2s")(1500000)     // "$1.5M"
d3.format(".1%")(0.127)        // "12.7%"

// Date formatting
d3.timeFormat("%b %d")(date)   // "Mar 15"
d3.timeFormat("%Y-%m")(date)   // "2024-03"

// Data aggregation
d3.rollup(data, v => d3.sum(v, d => d.value), d => d.category)
d3.bin().domain([0, 100]).thresholds(10)(values)
```
