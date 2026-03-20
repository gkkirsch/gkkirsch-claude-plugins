---
name: d3-development
description: >
  D3.js data visualization — scales, axes, selections, transitions, force
  layouts, geographic projections, and the D3 + React integration pattern.
  For highly custom, novel, or performance-critical visualizations.
  Triggers: "d3", "d3.js", "d3 chart", "force graph", "geographic map",
  "custom visualization", "svg chart", "treemap", "sunburst".
  NOT for: Standard business charts (use recharts-development).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# D3.js

## Setup

```bash
npm install d3 @types/d3
```

## The D3 + React Pattern

**Key insight**: Use D3 for math (scales, layouts, projections). Use React for rendering (JSX, DOM). Never let D3 touch the DOM in React apps.

```tsx
import * as d3 from "d3";

interface DataPoint {
  date: Date;
  value: number;
}

interface ChartProps {
  data: DataPoint[];
  width: number;
  height: number;
}

const margin = { top: 20, right: 30, bottom: 40, left: 50 };

function LineChart({ data, width, height }: ChartProps) {
  const innerWidth = width - margin.left - margin.right;
  const innerHeight = height - margin.top - margin.bottom;

  // D3 for math
  const xScale = d3.scaleTime()
    .domain(d3.extent(data, (d) => d.date) as [Date, Date])
    .range([0, innerWidth]);

  const yScale = d3.scaleLinear()
    .domain([0, d3.max(data, (d) => d.value)!])
    .nice()
    .range([innerHeight, 0]);

  const line = d3.line<DataPoint>()
    .x((d) => xScale(d.date))
    .y((d) => yScale(d.value))
    .curve(d3.curveMonotoneX);

  // React for rendering
  return (
    <svg width={width} height={height} role="img" aria-label="Line chart">
      <g transform={`translate(${margin.left},${margin.top})`}>
        {/* Grid lines */}
        {yScale.ticks(5).map((tick) => (
          <line
            key={tick}
            x1={0}
            x2={innerWidth}
            y1={yScale(tick)}
            y2={yScale(tick)}
            stroke="#e0e0e0"
            strokeDasharray="2,2"
          />
        ))}

        {/* X axis */}
        <g transform={`translate(0,${innerHeight})`}>
          {xScale.ticks(6).map((tick) => (
            <g key={tick.toISOString()} transform={`translate(${xScale(tick)},0)`}>
              <line y2={6} stroke="#666" />
              <text y={20} textAnchor="middle" fill="#666" fontSize={12}>
                {d3.timeFormat("%b")(tick)}
              </text>
            </g>
          ))}
        </g>

        {/* Y axis */}
        {yScale.ticks(5).map((tick) => (
          <text
            key={tick}
            x={-10}
            y={yScale(tick)}
            textAnchor="end"
            alignmentBaseline="middle"
            fill="#666"
            fontSize={12}
          >
            {tick}
          </text>
        ))}

        {/* Data line */}
        <path d={line(data)!} fill="none" stroke="#6366f1" strokeWidth={2} />

        {/* Data points */}
        {data.map((d, i) => (
          <circle
            key={i}
            cx={xScale(d.date)}
            cy={yScale(d.value)}
            r={3}
            fill="#6366f1"
          />
        ))}
      </g>
    </svg>
  );
}
```

## Scales

```typescript
import * as d3 from "d3";

// Linear (continuous numeric)
const yScale = d3.scaleLinear()
  .domain([0, 100])          // data range
  .range([height, 0])        // pixel range (inverted for y-axis)
  .nice()                    // round domain to nice values
  .clamp(true);              // restrict output to range

// Time (continuous dates)
const xScale = d3.scaleTime()
  .domain([new Date("2024-01"), new Date("2024-12")])
  .range([0, width]);

// Band (categorical, for bar charts)
const xBand = d3.scaleBand<string>()
  .domain(categories)        // ["A", "B", "C"]
  .range([0, width])
  .padding(0.2)              // gap between bars (0-1)
  .paddingOuter(0.1);
// xBand("A") -> x position, xBand.bandwidth() -> bar width

// Ordinal (categorical -> colors)
const colorScale = d3.scaleOrdinal<string>()
  .domain(categories)
  .range(d3.schemeTableau10);  // 10 distinct accessible colors

// Sequential (continuous -> color gradient)
const heatScale = d3.scaleSequential(d3.interpolateYlOrRd)
  .domain([0, 100]);
// heatScale(50) -> "#fb8d3d" (orange)

// Log (for exponential data)
const logScale = d3.scaleLog()
  .domain([1, 1000000])
  .range([0, width]);

// Sqrt (for area encoding — perceived area scales with sqrt)
const radiusScale = d3.scaleSqrt()
  .domain([0, maxPopulation])
  .range([2, 40]);
```

## Bar Chart (D3 + React)

```tsx
function BarChart({ data, width, height }: ChartProps) {
  const innerWidth = width - margin.left - margin.right;
  const innerHeight = height - margin.top - margin.bottom;

  const xScale = d3.scaleBand<string>()
    .domain(data.map((d) => d.category))
    .range([0, innerWidth])
    .padding(0.3);

  const yScale = d3.scaleLinear()
    .domain([0, d3.max(data, (d) => d.value)!])
    .nice()
    .range([innerHeight, 0]);

  return (
    <svg width={width} height={height}>
      <g transform={`translate(${margin.left},${margin.top})`}>
        {data.map((d) => (
          <rect
            key={d.category}
            x={xScale(d.category)}
            y={yScale(d.value)}
            width={xScale.bandwidth()}
            height={innerHeight - yScale(d.value)}
            fill="#6366f1"
            rx={4}
          />
        ))}

        {/* X axis labels */}
        {data.map((d) => (
          <text
            key={d.category}
            x={xScale(d.category)! + xScale.bandwidth() / 2}
            y={innerHeight + 20}
            textAnchor="middle"
            fontSize={12}
            fill="#666"
          >
            {d.category}
          </text>
        ))}
      </g>
    </svg>
  );
}
```

## Force-Directed Graph

```tsx
import { useEffect, useRef, useState } from "react";
import * as d3 from "d3";

interface Node extends d3.SimulationNodeDatum {
  id: string;
  group: number;
}

interface Link extends d3.SimulationLinkDatum<Node> {
  value: number;
}

function ForceGraph({ nodes: initialNodes, links: initialLinks, width, height }: Props) {
  const [nodes, setNodes] = useState<Node[]>([]);
  const [links, setLinks] = useState<Link[]>([]);
  const simRef = useRef<d3.Simulation<Node, Link>>();

  useEffect(() => {
    const nodeCopies = initialNodes.map((d) => ({ ...d }));
    const linkCopies = initialLinks.map((d) => ({ ...d }));

    const simulation = d3.forceSimulation(nodeCopies)
      .force("link", d3.forceLink<Node, Link>(linkCopies).id((d) => d.id).distance(50))
      .force("charge", d3.forceManyBody().strength(-100))
      .force("center", d3.forceCenter(width / 2, height / 2))
      .force("collision", d3.forceCollide().radius(15))
      .on("tick", () => {
        setNodes([...nodeCopies]);
        setLinks([...linkCopies]);
      });

    simRef.current = simulation;
    return () => { simulation.stop(); };
  }, [initialNodes, initialLinks, width, height]);

  const colorScale = d3.scaleOrdinal(d3.schemeCategory10);

  return (
    <svg width={width} height={height}>
      {links.map((link, i) => (
        <line
          key={i}
          x1={(link.source as Node).x}
          y1={(link.source as Node).y}
          x2={(link.target as Node).x}
          y2={(link.target as Node).y}
          stroke="#999"
          strokeOpacity={0.6}
          strokeWidth={Math.sqrt(link.value)}
        />
      ))}
      {nodes.map((node) => (
        <circle
          key={node.id}
          cx={node.x}
          cy={node.y}
          r={8}
          fill={colorScale(String(node.group))}
          stroke="#fff"
          strokeWidth={1.5}
        >
          <title>{node.id}</title>
        </circle>
      ))}
    </svg>
  );
}
```

## Geographic Map (Choropleth)

```tsx
import * as d3 from "d3";
import { feature } from "topojson-client";
import type { Topology } from "topojson-specification";

function ChoroplethMap({ geoData, values, width, height }: Props) {
  const projection = d3.geoAlbersUsa()
    .fitSize([width, height], feature(geoData, geoData.objects.states));

  const pathGenerator = d3.geoPath().projection(projection);

  const colorScale = d3.scaleSequential(d3.interpolateBlues)
    .domain(d3.extent(Object.values(values)) as [number, number]);

  const states = feature(geoData, geoData.objects.states);

  return (
    <svg width={width} height={height}>
      {(states as any).features.map((state: any) => (
        <path
          key={state.id}
          d={pathGenerator(state)!}
          fill={colorScale(values[state.id] ?? 0)}
          stroke="#fff"
          strokeWidth={0.5}
        >
          <title>{`${state.properties.name}: ${values[state.id]}`}</title>
        </path>
      ))}
    </svg>
  );
}
```

## Treemap

```tsx
import * as d3 from "d3";

interface TreeNode {
  name: string;
  value?: number;
  children?: TreeNode[];
}

function Treemap({ data, width, height }: { data: TreeNode; width: number; height: number }) {
  const root = d3.hierarchy(data)
    .sum((d) => d.value ?? 0)
    .sort((a, b) => (b.value ?? 0) - (a.value ?? 0));

  d3.treemap<TreeNode>()
    .size([width, height])
    .paddingInner(2)
    .paddingOuter(4)
    .round(true)(root);

  const colorScale = d3.scaleOrdinal(d3.schemeTableau10);

  return (
    <svg width={width} height={height}>
      {root.leaves().map((leaf, i) => {
        const x0 = (leaf as any).x0;
        const y0 = (leaf as any).y0;
        const x1 = (leaf as any).x1;
        const y1 = (leaf as any).y1;
        const w = x1 - x0;
        const h = y1 - y0;

        return (
          <g key={i} transform={`translate(${x0},${y0})`}>
            <rect width={w} height={h} fill={colorScale(leaf.parent?.data.name ?? "")} rx={2} />
            {w > 40 && h > 20 && (
              <text x={4} y={14} fontSize={11} fill="#fff" fontWeight={500}>
                {leaf.data.name}
              </text>
            )}
          </g>
        );
      })}
    </svg>
  );
}
```

## Responsive Hook

```tsx
import { useRef, useState, useEffect } from "react";

function useChartDimensions(aspectRatio = 16 / 9) {
  const containerRef = useRef<HTMLDivElement>(null);
  const [dimensions, setDimensions] = useState({ width: 600, height: 400 });

  useEffect(() => {
    const observer = new ResizeObserver(([entry]) => {
      const width = entry.contentRect.width;
      setDimensions({ width, height: width / aspectRatio });
    });

    if (containerRef.current) observer.observe(containerRef.current);
    return () => observer.disconnect();
  }, [aspectRatio]);

  return { containerRef, ...dimensions };
}

// Usage:
function Chart({ data }: Props) {
  const { containerRef, width, height } = useChartDimensions();

  return (
    <div ref={containerRef}>
      <LineChart data={data} width={width} height={height} />
    </div>
  );
}
```

## Color Palettes (Accessible)

```typescript
// Categorical (distinct groups)
d3.schemeTableau10     // 10 colors, colorblind-safe
d3.schemeCategory10    // 10 colors, classic D3
d3.schemePaired        // 12 colors, paired light/dark

// Sequential (low -> high)
d3.interpolateBlues    // white -> blue
d3.interpolateYlOrRd   // yellow -> red (heat map)
d3.interpolateViridis   // perceptually uniform, colorblind-safe

// Diverging (negative <-> positive)
d3.interpolateRdBu     // red -> blue
d3.interpolatePiYG     // pink -> green

// Usage with scale
const color = d3.scaleSequential(d3.interpolateViridis).domain([0, 100]);
color(50); // -> "#21918c"
```

## Transitions with useRef (When Needed)

```tsx
import { useRef, useEffect } from "react";
import * as d3 from "d3";

// Only use D3 transitions for complex coordinated animations
// For simple animations, prefer CSS transitions
function AnimatedBar({ x, y, width, height, fill }: BarProps) {
  const ref = useRef<SVGRectElement>(null);

  useEffect(() => {
    d3.select(ref.current)
      .transition()
      .duration(500)
      .ease(d3.easeCubicOut)
      .attr("y", y)
      .attr("height", height);
  }, [y, height]);

  return (
    <rect
      ref={ref}
      x={x}
      y={y + height}  // Start from bottom
      width={width}
      height={0}       // Start at 0
      fill={fill}
      rx={4}
    />
  );
}
```

## Gotchas

1. **Never use D3 selections in React** — `d3.select(ref).append("rect")` fights React's virtual DOM. Use D3 for calculations (scales, layouts, paths) and React for rendering (JSX). The only exception is transitions via `useRef`.

2. **`d3.extent()` returns `[undefined, undefined]` on empty arrays** — Always handle the empty case: `d3.extent(data, d => d.value) as [number, number]` or check `data.length > 0` first.

3. **SVG y-axis is inverted** — y=0 is the TOP of the SVG. For charts, set `range([height, 0])` on yScale so higher values render higher on screen.

4. **`scaleSqrt` for area encoding, not `scaleLinear`** — Human perception of area is proportional to the square root of the value. Using linear scale for bubble radius makes large values look disproportionately large.

5. **Force simulation runs asynchronously** — `d3.forceSimulation().on("tick", ...)` fires many times. In React, call `setState` on each tick to update positions. Stop the simulation in the cleanup function to prevent memory leaks.

6. **GeoJSON vs TopoJSON** — TopoJSON is 80%+ smaller but needs `topojson-client` to convert to GeoJSON for rendering. Always convert with `feature()` before passing to `d3.geoPath()`.
