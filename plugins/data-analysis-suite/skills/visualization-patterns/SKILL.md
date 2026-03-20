---
name: visualization-patterns
description: >
  Data visualization patterns for web dashboards and reports.
  Use when building charts, dashboards, data tables, or interactive
  visualizations with D3, Chart.js, Recharts, or plain SVG.
  Triggers: "chart", "graph", "dashboard", "visualization", "plot",
  "bar chart", "line chart", "data table", "sparkline".
  NOT for: backend data processing, ETL, or statistical modeling.
version: 1.0.0
allowed-tools: Read, Grep, Glob, Bash
---

# Data Visualization Patterns

## Recharts (React) — Most Common

```tsx
import {
  LineChart, Line, BarChart, Bar, AreaChart, Area,
  XAxis, YAxis, CartesianGrid, Tooltip, Legend,
  ResponsiveContainer, PieChart, Pie, Cell,
  ComposedChart, Scatter
} from 'recharts';

// Time series with multiple metrics
interface TimeSeriesData {
  date: string;
  revenue: number;
  users: number;
  conversionRate: number;
}

function RevenueChart({ data }: { data: TimeSeriesData[] }) {
  return (
    <ResponsiveContainer width="100%" height={400}>
      <ComposedChart data={data}>
        <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
        <XAxis
          dataKey="date"
          tick={{ fontSize: 12 }}
          tickFormatter={(date) => new Date(date).toLocaleDateString('en-US', { month: 'short', day: 'numeric' })}
        />
        <YAxis
          yAxisId="left"
          tickFormatter={(val) => `$${(val / 1000).toFixed(0)}k`}
        />
        <YAxis
          yAxisId="right"
          orientation="right"
          tickFormatter={(val) => `${val}%`}
        />
        <Tooltip
          contentStyle={{ backgroundColor: '#1f2937', border: '1px solid #374151' }}
          labelFormatter={(date) => new Date(date).toLocaleDateString()}
          formatter={(value: number, name: string) => {
            if (name === 'revenue') return [`$${value.toLocaleString()}`, 'Revenue'];
            if (name === 'conversionRate') return [`${value}%`, 'Conversion'];
            return [value.toLocaleString(), name];
          }}
        />
        <Legend />
        <Bar yAxisId="left" dataKey="revenue" fill="#3b82f6" radius={[4, 4, 0, 0]} />
        <Line yAxisId="right" dataKey="conversionRate" stroke="#10b981" strokeWidth={2} dot={false} />
      </ComposedChart>
    </ResponsiveContainer>
  );
}

// KPI card with sparkline
function KPICard({ title, value, change, sparkData }: {
  title: string;
  value: string;
  change: number;
  sparkData: number[];
}) {
  const isPositive = change >= 0;

  return (
    <div className="p-4 bg-gray-900 rounded-lg border border-gray-800">
      <p className="text-sm text-gray-400">{title}</p>
      <p className="text-2xl font-bold mt-1">{value}</p>
      <div className="flex items-center gap-2 mt-2">
        <span className={`text-sm ${isPositive ? 'text-green-400' : 'text-red-400'}`}>
          {isPositive ? '↑' : '↓'} {Math.abs(change)}%
        </span>
        <ResponsiveContainer width={60} height={24}>
          <LineChart data={sparkData.map((v, i) => ({ v, i }))}>
            <Line
              dataKey="v"
              stroke={isPositive ? '#10b981' : '#ef4444'}
              strokeWidth={1.5}
              dot={false}
            />
          </LineChart>
        </ResponsiveContainer>
      </div>
    </div>
  );
}
```

## Data Table with Sorting and Filtering

```tsx
import { useState, useMemo } from 'react';

interface Column<T> {
  key: keyof T;
  label: string;
  sortable?: boolean;
  render?: (value: T[keyof T], row: T) => React.ReactNode;
  align?: 'left' | 'right' | 'center';
}

function DataTable<T extends Record<string, any>>({
  data, columns, pageSize = 20
}: {
  data: T[];
  columns: Column<T>[];
  pageSize?: number;
}) {
  const [sortKey, setSortKey] = useState<keyof T | null>(null);
  const [sortDir, setSortDir] = useState<'asc' | 'desc'>('asc');
  const [page, setPage] = useState(0);
  const [filter, setFilter] = useState('');

  const filtered = useMemo(() => {
    if (!filter) return data;
    const lower = filter.toLowerCase();
    return data.filter(row =>
      columns.some(col => String(row[col.key]).toLowerCase().includes(lower))
    );
  }, [data, filter, columns]);

  const sorted = useMemo(() => {
    if (!sortKey) return filtered;
    return [...filtered].sort((a, b) => {
      const aVal = a[sortKey], bVal = b[sortKey];
      const cmp = aVal < bVal ? -1 : aVal > bVal ? 1 : 0;
      return sortDir === 'asc' ? cmp : -cmp;
    });
  }, [filtered, sortKey, sortDir]);

  const paged = sorted.slice(page * pageSize, (page + 1) * pageSize);
  const totalPages = Math.ceil(sorted.length / pageSize);

  const handleSort = (key: keyof T) => {
    if (sortKey === key) {
      setSortDir(d => d === 'asc' ? 'desc' : 'asc');
    } else {
      setSortKey(key);
      setSortDir('asc');
    }
  };

  return (
    <div>
      <input
        type="text"
        placeholder="Filter..."
        value={filter}
        onChange={e => { setFilter(e.target.value); setPage(0); }}
        className="mb-4 px-3 py-2 bg-gray-800 border border-gray-700 rounded"
      />
      <table className="w-full text-sm">
        <thead>
          <tr className="border-b border-gray-700">
            {columns.map(col => (
              <th
                key={String(col.key)}
                onClick={() => col.sortable && handleSort(col.key)}
                className={`px-4 py-3 text-${col.align ?? 'left'} ${col.sortable ? 'cursor-pointer hover:text-blue-400' : ''}`}
              >
                {col.label}
                {sortKey === col.key && (sortDir === 'asc' ? ' ↑' : ' ↓')}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {paged.map((row, i) => (
            <tr key={i} className="border-b border-gray-800 hover:bg-gray-800/50">
              {columns.map(col => (
                <td key={String(col.key)} className={`px-4 py-3 text-${col.align ?? 'left'}`}>
                  {col.render ? col.render(row[col.key], row) : String(row[col.key])}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
      <div className="flex justify-between items-center mt-4 text-sm text-gray-400">
        <span>{sorted.length} results</span>
        <div className="flex gap-2">
          <button onClick={() => setPage(p => Math.max(0, p - 1))} disabled={page === 0}>Prev</button>
          <span>{page + 1} / {totalPages}</span>
          <button onClick={() => setPage(p => Math.min(totalPages - 1, p + 1))} disabled={page >= totalPages - 1}>Next</button>
        </div>
      </div>
    </div>
  );
}
```

## Number Formatting Utilities

```typescript
// Consistent number formatting for dashboards
const formatters = {
  currency: (value: number, currency = 'USD') =>
    new Intl.NumberFormat('en-US', { style: 'currency', currency, maximumFractionDigits: 0 }).format(value),

  compact: (value: number) =>
    new Intl.NumberFormat('en-US', { notation: 'compact', maximumFractionDigits: 1 }).format(value),

  percent: (value: number, decimals = 1) =>
    `${(value * 100).toFixed(decimals)}%`,

  duration: (seconds: number) => {
    if (seconds < 60) return `${seconds}s`;
    if (seconds < 3600) return `${Math.floor(seconds / 60)}m ${seconds % 60}s`;
    return `${Math.floor(seconds / 3600)}h ${Math.floor((seconds % 3600) / 60)}m`;
  },

  bytes: (bytes: number) => {
    const units = ['B', 'KB', 'MB', 'GB', 'TB'];
    let i = 0;
    let val = bytes;
    while (val >= 1024 && i < units.length - 1) { val /= 1024; i++; }
    return `${val.toFixed(1)} ${units[i]}`;
  },
};
```

## Gotchas

1. **ResponsiveContainer needs a parent with defined height** — if the parent has `height: auto` or no height, the chart renders at 0px. Always wrap in a container with explicit height or min-height.

2. **Recharts re-renders on every parent render** — wrap chart components in `React.memo()` and memoize data arrays with `useMemo()`. Passing a new array reference on every render causes expensive chart re-animation.

3. **Large datasets crash the browser** — rendering 10,000+ data points as SVG elements is a performance disaster. Downsample data to ~500 points for line charts (use LTTB algorithm), or use canvas-based renderers for large datasets.

4. **Timezone-naive date axes** — if your data timestamps are UTC but you format them in local time, the x-axis labels shift depending on the viewer's timezone. Always be explicit about timezone: display UTC or convert consistently.

5. **Missing empty state** — charts with zero data points render as blank boxes with axes. Always check `data.length === 0` and show an empty state message instead of an empty chart.

6. **Color palette accessibility** — default chart color palettes often fail for color-blind users. Use patterns (dashed lines, different shapes) in addition to colors. Test with a color blindness simulator.
