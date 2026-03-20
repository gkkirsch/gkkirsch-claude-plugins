---
name: recharts-development
description: >
  Recharts data visualization — composable React chart components with
  responsive containers, tooltips, animations, custom shapes, and real-time
  updates. The go-to library for standard business charts in React.
  Triggers: "recharts", "react chart", "line chart react", "bar chart react",
  "pie chart react", "area chart", "dashboard charts".
  NOT for: Highly custom/novel visualizations (use d3-development).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Recharts

## Setup

```bash
npm install recharts
```

## Basic Line Chart

```tsx
import {
  LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip,
  Legend, ResponsiveContainer,
} from "recharts";

interface DataPoint {
  month: string;
  revenue: number;
  profit: number;
}

const data: DataPoint[] = [
  { month: "Jan", revenue: 4000, profit: 2400 },
  { month: "Feb", revenue: 3000, profit: 1398 },
  { month: "Mar", revenue: 2000, profit: 9800 },
  { month: "Apr", revenue: 2780, profit: 3908 },
  { month: "May", revenue: 1890, profit: 4800 },
  { month: "Jun", revenue: 2390, profit: 3800 },
];

function RevenueChart() {
  return (
    <ResponsiveContainer width="100%" height={400}>
      <LineChart data={data} margin={{ top: 5, right: 30, left: 20, bottom: 5 }}>
        <CartesianGrid strokeDasharray="3 3" stroke="#e0e0e0" />
        <XAxis dataKey="month" />
        <YAxis tickFormatter={(v) => `$${v / 1000}k`} />
        <Tooltip
          formatter={(value: number) => [`$${value.toLocaleString()}`, undefined]}
          contentStyle={{ borderRadius: 8 }}
        />
        <Legend />
        <Line
          type="monotone"
          dataKey="revenue"
          stroke="#8884d8"
          strokeWidth={2}
          dot={{ r: 4 }}
          activeDot={{ r: 6 }}
        />
        <Line
          type="monotone"
          dataKey="profit"
          stroke="#82ca9d"
          strokeWidth={2}
        />
      </LineChart>
    </ResponsiveContainer>
  );
}
```

## Bar Chart with Stacking

```tsx
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from "recharts";

function SalesChart({ data }: { data: SalesData[] }) {
  return (
    <ResponsiveContainer width="100%" height={400}>
      <BarChart data={data} margin={{ top: 20, right: 30, left: 20, bottom: 5 }}>
        <CartesianGrid strokeDasharray="3 3" />
        <XAxis dataKey="quarter" />
        <YAxis />
        <Tooltip />
        <Legend />
        <Bar dataKey="online" stackId="sales" fill="#8884d8" radius={[0, 0, 0, 0]} />
        <Bar dataKey="retail" stackId="sales" fill="#82ca9d" radius={[4, 4, 0, 0]} />
      </BarChart>
    </ResponsiveContainer>
  );
}
```

## Composed Chart (Multiple Types)

```tsx
import {
  ComposedChart, Line, Bar, Area, XAxis, YAxis,
  CartesianGrid, Tooltip, Legend, ResponsiveContainer,
} from "recharts";

function DashboardChart({ data }: { data: MetricData[] }) {
  return (
    <ResponsiveContainer width="100%" height={400}>
      <ComposedChart data={data}>
        <CartesianGrid strokeDasharray="3 3" />
        <XAxis dataKey="date" />
        <YAxis yAxisId="left" />
        <YAxis yAxisId="right" orientation="right" />
        <Tooltip />
        <Legend />
        <Area
          yAxisId="left"
          type="monotone"
          dataKey="pageViews"
          fill="#8884d8"
          fillOpacity={0.3}
          stroke="#8884d8"
        />
        <Bar yAxisId="left" dataKey="conversions" fill="#82ca9d" barSize={20} />
        <Line
          yAxisId="right"
          type="monotone"
          dataKey="conversionRate"
          stroke="#ff7300"
          strokeWidth={2}
        />
      </ComposedChart>
    </ResponsiveContainer>
  );
}
```

## Pie / Donut Chart

```tsx
import { PieChart, Pie, Cell, Tooltip, ResponsiveContainer, Label } from "recharts";

const COLORS = ["#0088FE", "#00C49F", "#FFBB28", "#FF8042", "#8884d8"];

function BudgetChart({ data }: { data: { name: string; value: number }[] }) {
  const total = data.reduce((sum, d) => sum + d.value, 0);

  return (
    <ResponsiveContainer width="100%" height={400}>
      <PieChart>
        <Pie
          data={data}
          cx="50%"
          cy="50%"
          innerRadius={80}    // > 0 makes it a donut
          outerRadius={120}
          paddingAngle={2}
          dataKey="value"
          label={({ name, percent }) => `${name} ${(percent * 100).toFixed(0)}%`}
        >
          {data.map((_, index) => (
            <Cell key={index} fill={COLORS[index % COLORS.length]} />
          ))}
          <Label
            value={`$${(total / 1000).toFixed(0)}k`}
            position="center"
            style={{ fontSize: 24, fontWeight: "bold" }}
          />
        </Pie>
        <Tooltip formatter={(value: number) => `$${value.toLocaleString()}`} />
      </PieChart>
    </ResponsiveContainer>
  );
}
```

## Custom Tooltip

```tsx
function CustomTooltip({ active, payload, label }: TooltipProps<number, string>) {
  if (!active || !payload?.length) return null;

  return (
    <div className="bg-white shadow-lg rounded-lg p-3 border">
      <p className="font-semibold text-gray-900">{label}</p>
      {payload.map((entry, i) => (
        <p key={i} style={{ color: entry.color }} className="text-sm">
          {entry.name}: {typeof entry.value === "number"
            ? `$${entry.value.toLocaleString()}`
            : entry.value}
        </p>
      ))}
    </div>
  );
}

// Usage: <Tooltip content={<CustomTooltip />} />
```

## Custom Shapes

```tsx
// Custom bar shape
function RoundedBar(props: any) {
  const { x, y, width, height, fill } = props;
  const radius = 6;
  return (
    <rect x={x} y={y} width={width} height={height} fill={fill} rx={radius} ry={radius} />
  );
}

// Custom dot
function PulseDot(props: any) {
  const { cx, cy, fill, value } = props;
  if (value > 5000) {
    return (
      <g>
        <circle cx={cx} cy={cy} r={8} fill={fill} opacity={0.3} />
        <circle cx={cx} cy={cy} r={4} fill={fill} />
      </g>
    );
  }
  return <circle cx={cx} cy={cy} r={3} fill={fill} />;
}

// Usage:
// <Bar shape={<RoundedBar />} />
// <Line dot={<PulseDot />} />
```

## Real-Time Updates

```tsx
import { useEffect, useRef, useState } from "react";

function RealtimeChart() {
  const [data, setData] = useState<DataPoint[]>([]);
  const intervalRef = useRef<number>();

  useEffect(() => {
    intervalRef.current = window.setInterval(() => {
      setData((prev) => {
        const next = [
          ...prev.slice(-29),  // Keep last 30 points
          {
            time: new Date().toLocaleTimeString(),
            value: Math.random() * 100,
          },
        ];
        return next;
      });
    }, 1000);

    return () => clearInterval(intervalRef.current);
  }, []);

  return (
    <ResponsiveContainer width="100%" height={300}>
      <LineChart data={data}>
        <XAxis dataKey="time" />
        <YAxis domain={[0, 100]} />
        <Line
          type="monotone"
          dataKey="value"
          stroke="#8884d8"
          dot={false}
          isAnimationActive={false}  // Disable animation for real-time
        />
      </LineChart>
    </ResponsiveContainer>
  );
}
```

## Brush (Zoom & Pan)

```tsx
import { Brush } from "recharts";

<LineChart data={longTimeSeries}>
  <XAxis dataKey="date" />
  <YAxis />
  <Line type="monotone" dataKey="value" stroke="#8884d8" />
  <Brush
    dataKey="date"
    height={30}
    stroke="#8884d8"
    startIndex={longTimeSeries.length - 30}  // Show last 30 by default
  />
</LineChart>
```

## Reference Lines & Areas

```tsx
import { ReferenceLine, ReferenceArea } from "recharts";

<LineChart data={data}>
  {/* Horizontal threshold line */}
  <ReferenceLine
    y={5000}
    stroke="red"
    strokeDasharray="3 3"
    label={{ value: "Target", position: "right" }}
  />

  {/* Vertical event marker */}
  <ReferenceLine x="Mar" stroke="#666" label="Launch" />

  {/* Highlighted region */}
  <ReferenceArea
    x1="Apr"
    x2="Jun"
    fill="#8884d8"
    fillOpacity={0.1}
    label="Q2"
  />
</LineChart>
```

## Accessibility

```tsx
<ResponsiveContainer>
  <BarChart
    data={data}
    role="img"
    aria-label="Monthly sales comparison for 2024"
  >
    {/* Add descriptive title */}
    <text x="50%" y={15} textAnchor="middle" className="text-lg font-semibold">
      Monthly Sales 2024
    </text>

    {/* Screen reader description */}
    <desc>
      Bar chart showing monthly sales from January to June.
      Highest sales were in March at $9,800.
    </desc>

    <XAxis dataKey="month" />
    <YAxis />
    <Tooltip />
    <Bar dataKey="sales" fill="#8884d8">
      {data.map((entry, i) => (
        <Cell
          key={i}
          fill={COLORS[i % COLORS.length]}
          aria-label={`${entry.month}: $${entry.sales}`}
        />
      ))}
    </Bar>
  </BarChart>
</ResponsiveContainer>
```

## Theming

```tsx
const chartTheme = {
  colors: {
    primary: "#6366f1",
    secondary: "#22c55e",
    tertiary: "#f59e0b",
    grid: "#e5e7eb",
    text: "#374151",
    background: "#ffffff",
  },
  fonts: {
    family: "Inter, system-ui, sans-serif",
    size: 12,
  },
};

function ThemedChart({ data, theme = chartTheme }: Props) {
  return (
    <ResponsiveContainer width="100%" height={400}>
      <LineChart data={data}>
        <CartesianGrid stroke={theme.colors.grid} strokeDasharray="3 3" />
        <XAxis
          dataKey="date"
          tick={{ fill: theme.colors.text, fontSize: theme.fonts.size }}
          tickLine={{ stroke: theme.colors.grid }}
        />
        <YAxis
          tick={{ fill: theme.colors.text, fontSize: theme.fonts.size }}
          tickLine={{ stroke: theme.colors.grid }}
        />
        <Line dataKey="value" stroke={theme.colors.primary} strokeWidth={2} />
      </LineChart>
    </ResponsiveContainer>
  );
}
```

## Gotchas

1. **Always wrap in `ResponsiveContainer`** — Without it, charts render at 0x0 or a fixed size. Set width="100%" and a fixed height. Never set width/height on the chart component itself when using ResponsiveContainer.

2. **`isAnimationActive={false}` for real-time data** — Animation on rapidly updating data causes stuttering and memory leaks. Disable animation for live dashboards and streaming data.

3. **`dataKey` must match object keys exactly** — `dataKey="Revenue"` won't match `{ revenue: 100 }`. Case matters. Use the exact property name from your data objects.

4. **Pie charts need explicit `<Cell>` for colors** — Unlike Bar/Line which accept `fill` directly, Pie charts render all slices the same color unless you map `<Cell>` components with individual fills.

5. **Stacked charts need `stackId`** — To stack bars or areas, they must share the same `stackId` string. Without it, they overlap instead of stacking.

6. **Large datasets (5K+ points) cause jank** — Recharts renders SVG DOM nodes for each point. For large datasets, downsample first (e.g., LTTB algorithm) or use a Canvas-based library like Chart.js.
