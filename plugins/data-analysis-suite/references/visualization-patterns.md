# Visualization Patterns Reference

> Premium reference for the Data Analysis Suite plugin.
> Covers chart selection, color systems, accessibility, code templates, and anti-patterns.

---

## Chart Type Selection Guide

### Selection Flowchart

Use the following decision tree to choose the right chart type.

```
What do you want to show?
|
|-- Comparison
|   |-- Among items .............. Bar chart (horizontal if many items)
|   |-- Over time ............... Line chart / slope chart
|   |-- Two variables ........... Scatter plot
|   |-- Part-to-whole ........... Stacked bar / treemap
|
|-- Composition
|   |-- Static .................. Stacked bar / treemap / waffle
|   |-- Over time ............... Stacked area / stream graph
|   |-- Hierarchical ............ Treemap / sunburst
|
|-- Distribution
|   |-- Single variable ......... Histogram / KDE / box plot
|   |-- Two variables ........... Scatter / 2-D histogram / heatmap
|   |-- Multiple groups ......... Violin / ridgeline / small multiples
|
|-- Relationship
|   |-- Two variables ........... Scatter plot
|   |-- Three variables ......... Bubble chart
|   |-- Correlation matrix ...... Heatmap
|   |-- Network ................. Network / chord diagram
|
|-- Trend
|   |-- Single series ........... Line chart / sparkline
|   |-- Multiple series ......... Multi-line / small multiples
|   |-- Part-of-whole trend ..... Stacked area
|
|-- Flow
|   |-- Process stages .......... Sankey diagram / funnel
|   |-- Sequential change ....... Waterfall chart
|
|-- Geographic
|   |-- Regional values ......... Choropleth
|   |-- Point data .............. Bubble map
|   |-- Density ................. Hex-bin map / heatmap overlay
```

#### Secondary Considerations

| Factor | Guidance |
|---|---|
| Number of categories | >7 categories: use horizontal bar, treemap, or small multiples |
| Data type: categorical | Bar, dot plot, lollipop |
| Data type: continuous | Histogram, KDE, scatter |
| Data type: temporal | Line, area, calendar heatmap |
| Audience: executive | KPI cards, sparklines, simplified bar/line |
| Audience: technical | Box plot, violin, scatter with regression, faceted plots |
| Data points: <20 | Lollipop, slope, bar |
| Data points: 20-200 | Scatter, line, grouped bar |
| Data points: >1000 | Hex-bin, heatmap, 2-D histogram |

---

### Detailed Chart Catalog

---

#### Bar Charts

##### Vertical Bar Chart

**When to use:** Comparing values across a small number of categories (fewer than 12). The default choice for categorical comparison.

**When NOT to use:** More than 15 categories (use horizontal bar). Continuous data (use histogram). Time series with many periods (use line chart).

**Data requirements:** One categorical column, one numeric column.

**Plotly (Python):**

```python
import plotly.express as px
import pandas as pd

df = pd.DataFrame({
    "region": ["North", "South", "East", "West"],
    "revenue": [420000, 380000, 510000, 290000]
})

fig = px.bar(
    df, x="region", y="revenue",
    title="Revenue by Region",
    color_discrete_sequence=["#4C78A8"],
    text_auto="$.3s"
)
fig.update_layout(
    yaxis_title="Revenue (USD)",
    xaxis_title=None,
    plot_bgcolor="#FAFAFA",
    font_family="Inter, system-ui, sans-serif",
    title_font_size=18
)
fig.show()
```

**Matplotlib:**

```python
import matplotlib.pyplot as plt

regions = ["North", "South", "East", "West"]
revenue = [420000, 380000, 510000, 290000]

fig, ax = plt.subplots(figsize=(8, 5))
bars = ax.bar(regions, revenue, color="#4C78A8", width=0.6)
ax.bar_label(bars, fmt="${:,.0f}", padding=4, fontsize=9)
ax.set_ylabel("Revenue (USD)")
ax.set_title("Revenue by Region", fontsize=16, fontweight="bold", loc="left")
ax.spines[["top", "right"]].set_visible(False)
ax.yaxis.set_major_formatter(plt.FuncFormatter(lambda x, _: f"${x/1e3:.0f}K"))
plt.tight_layout()
plt.show()
```

**Best practices:**
- Always start the y-axis at zero.
- Sort bars by value (descending) unless the categories have inherent order.
- Use a single color unless encoding a second variable.
- Add direct labels when the chart will be viewed statically (slides, PDFs).

##### Horizontal Bar Chart

**When to use:** Many categories (more than 7), long category labels, ranking comparisons.

**When NOT to use:** Fewer than 4 categories (vertical bar is more natural).

**Plotly:**

```python
fig = px.bar(
    df.sort_values("revenue"),
    x="revenue", y="region",
    orientation="h",
    title="Revenue by Region",
    color_discrete_sequence=["#4C78A8"],
    text_auto="$.3s"
)
fig.update_layout(yaxis_title=None, xaxis_title="Revenue (USD)")
fig.show()
```

**Matplotlib:**

```python
fig, ax = plt.subplots(figsize=(8, 5))
sorted_df = df.sort_values("revenue")
ax.barh(sorted_df["region"], sorted_df["revenue"], color="#4C78A8", height=0.6)
ax.set_xlabel("Revenue (USD)")
ax.set_title("Revenue by Region", fontsize=16, fontweight="bold", loc="left")
ax.spines[["top", "right"]].set_visible(False)
plt.tight_layout()
plt.show()
```

##### Grouped Bar Chart

**When to use:** Comparing subcategories within categories. Two categorical variables and one numeric.

**When NOT to use:** More than 4 groups per category (becomes cluttered). Use small multiples instead.

**Plotly:**

```python
df_grouped = pd.DataFrame({
    "region": ["North", "North", "South", "South", "East", "East", "West", "West"],
    "quarter": ["Q1", "Q2", "Q1", "Q2", "Q1", "Q2", "Q1", "Q2"],
    "revenue": [200000, 220000, 180000, 200000, 240000, 270000, 130000, 160000]
})

fig = px.bar(
    df_grouped, x="region", y="revenue", color="quarter",
    barmode="group",
    title="Revenue by Region and Quarter",
    color_discrete_sequence=["#4C78A8", "#F58518"]
)
fig.show()
```

##### Stacked Bar Chart

**When to use:** Part-to-whole comparisons across categories. Use 100% stacked bar to compare proportions.

**When NOT to use:** When exact comparison of individual segments is critical (only the bottom segment shares a baseline).

**Plotly:**

```python
fig = px.bar(
    df_grouped, x="region", y="revenue", color="quarter",
    barmode="stack",
    title="Revenue by Region (Stacked)",
    color_discrete_sequence=["#4C78A8", "#F58518"]
)
fig.show()
```

##### Diverging Bar Chart

**When to use:** Showing positive and negative values from a center point, such as survey responses (agree/disagree) or profit/loss.

**Plotly:**

```python
df_div = pd.DataFrame({
    "metric": ["Customer Sat.", "Response Time", "Resolution Rate", "NPS"],
    "change": [12, -8, 5, -3]
})

colors = ["#E45756" if v < 0 else "#4C78A8" for v in df_div["change"]]

fig = px.bar(
    df_div, x="change", y="metric", orientation="h",
    title="YoY Change (%)",
    color=df_div["change"].apply(lambda x: "Decrease" if x < 0 else "Increase"),
    color_discrete_map={"Increase": "#4C78A8", "Decrease": "#E45756"}
)
fig.update_layout(xaxis_title="Change (%)", yaxis_title=None)
fig.show()
```

---

#### Line Charts

##### Basic Line Chart

**When to use:** Showing trends over time. One or more continuous series along a time axis.

**When NOT to use:** Fewer than 5 time points (use bar chart). Categorical x-axis with no order.

**Data requirements:** One temporal/ordinal column, one or more numeric columns.

**Plotly:**

```python
import plotly.express as px
import pandas as pd
import numpy as np

dates = pd.date_range("2024-01-01", periods=12, freq="MS")
df_line = pd.DataFrame({
    "date": dates,
    "revenue": np.random.randint(300000, 600000, 12).cumsum() / 12
})

fig = px.line(
    df_line, x="date", y="revenue",
    title="Monthly Revenue Trend",
    markers=True
)
fig.update_layout(
    xaxis_title=None,
    yaxis_title="Revenue (USD)",
    yaxis_tickformat="$,.0f"
)
fig.show()
```

**Matplotlib:**

```python
fig, ax = plt.subplots(figsize=(10, 5))
ax.plot(df_line["date"], df_line["revenue"], marker="o", color="#4C78A8", linewidth=2)
ax.set_title("Monthly Revenue Trend", fontsize=16, fontweight="bold", loc="left")
ax.set_ylabel("Revenue (USD)")
ax.yaxis.set_major_formatter(plt.FuncFormatter(lambda x, _: f"${x/1e3:.0f}K"))
ax.spines[["top", "right"]].set_visible(False)
plt.tight_layout()
plt.show()
```

##### Multi-Series Line Chart

**When to use:** Comparing trends across 2-5 categories over time.

**When NOT to use:** More than 7 series (use small multiples or highlight one series).

**Plotly:**

```python
df_multi = pd.DataFrame({
    "date": np.tile(dates, 3),
    "region": np.repeat(["North", "South", "East"], 12),
    "revenue": np.random.randint(100000, 400000, 36)
})

fig = px.line(
    df_multi, x="date", y="revenue", color="region",
    title="Revenue Trend by Region",
    color_discrete_sequence=["#4C78A8", "#F58518", "#E45756"]
)
fig.show()
```

##### Area Chart

**When to use:** Emphasizing the magnitude of a trend, especially cumulative or part-to-whole over time.

**When NOT to use:** When multiple overlapping series obscure each other (use stacked area or line).

**Plotly:**

```python
fig = px.area(
    df_multi, x="date", y="revenue", color="region",
    title="Revenue by Region (Stacked Area)",
    color_discrete_sequence=["#4C78A8", "#F58518", "#E45756"]
)
fig.show()
```

##### Stepped Line Chart

**When to use:** Discrete changes that persist until the next change (pricing tiers, status changes, step functions).

**Plotly:**

```python
fig = px.line(
    df_line, x="date", y="revenue",
    title="Pricing Tier Over Time",
    line_shape="hv"  # horizontal-then-vertical steps
)
fig.show()
```

---

#### Scatter Plots

##### Basic Scatter Plot

**When to use:** Exploring the relationship between two continuous variables.

**When NOT to use:** Categorical data. Fewer than 10 data points (label directly instead).

**Data requirements:** Two numeric columns. Optional: color variable, size variable.

**Plotly:**

```python
np.random.seed(42)
n = 100
df_scatter = pd.DataFrame({
    "ad_spend": np.random.uniform(1000, 50000, n),
    "revenue": np.random.uniform(5000, 200000, n),
    "region": np.random.choice(["North", "South", "East", "West"], n)
})
df_scatter["revenue"] = df_scatter["ad_spend"] * 3.2 + np.random.normal(0, 10000, n)

fig = px.scatter(
    df_scatter, x="ad_spend", y="revenue", color="region",
    title="Ad Spend vs Revenue",
    trendline="ols",
    color_discrete_sequence=["#4C78A8", "#F58518", "#E45756", "#72B7B2"]
)
fig.update_layout(
    xaxis_title="Ad Spend (USD)",
    yaxis_title="Revenue (USD)"
)
fig.show()
```

**Matplotlib:**

```python
fig, ax = plt.subplots(figsize=(8, 6))
ax.scatter(df_scatter["ad_spend"], df_scatter["revenue"],
           alpha=0.6, color="#4C78A8", edgecolors="white", linewidth=0.5)
ax.set_xlabel("Ad Spend (USD)")
ax.set_ylabel("Revenue (USD)")
ax.set_title("Ad Spend vs Revenue", fontsize=16, fontweight="bold", loc="left")
ax.spines[["top", "right"]].set_visible(False)
plt.tight_layout()
plt.show()
```

##### Bubble Chart

**When to use:** Three variables: x-position, y-position, and size. Optionally, color for a fourth variable.

**Plotly:**

```python
df_scatter["employees"] = np.random.randint(10, 500, n)

fig = px.scatter(
    df_scatter, x="ad_spend", y="revenue",
    size="employees", color="region",
    title="Ad Spend vs Revenue (size = employees)",
    size_max=40,
    color_discrete_sequence=["#4C78A8", "#F58518", "#E45756", "#72B7B2"]
)
fig.show()
```

##### Connected Scatter Plot

**When to use:** Showing a trajectory through two variable space over time. Each point is a time step.

**Plotly:**

```python
fig = px.line(
    df_line, x="revenue", y=df_line["revenue"].diff().fillna(0),
    markers=True, title="Revenue vs Change in Revenue"
)
fig.show()
```

---

#### Histograms

##### Basic Histogram

**When to use:** Showing the distribution of a single continuous variable.

**When NOT to use:** Categorical data. Comparison across groups (use overlapping or small multiples).

**Plotly:**

```python
fig = px.histogram(
    df_scatter, x="revenue",
    nbins=30,
    title="Revenue Distribution",
    color_discrete_sequence=["#4C78A8"]
)
fig.update_layout(
    xaxis_title="Revenue (USD)",
    yaxis_title="Count",
    bargap=0.05
)
fig.show()
```

**Matplotlib:**

```python
fig, ax = plt.subplots(figsize=(8, 5))
ax.hist(df_scatter["revenue"], bins=30, color="#4C78A8", edgecolor="white", linewidth=0.5)
ax.set_xlabel("Revenue (USD)")
ax.set_ylabel("Count")
ax.set_title("Revenue Distribution", fontsize=16, fontweight="bold", loc="left")
ax.spines[["top", "right"]].set_visible(False)
plt.tight_layout()
plt.show()
```

##### Overlapping Histograms

**When to use:** Comparing distributions of the same variable across 2-3 groups.

**Plotly:**

```python
fig = px.histogram(
    df_scatter, x="revenue", color="region",
    barmode="overlay", nbins=30, opacity=0.6,
    title="Revenue Distribution by Region",
    color_discrete_sequence=["#4C78A8", "#F58518", "#E45756", "#72B7B2"]
)
fig.show()
```

##### KDE (Kernel Density Estimate)

**When to use:** Smooth continuous approximation of distribution. Overlaying multiple groups without the bin-size problem.

**Matplotlib + Seaborn:**

```python
import seaborn as sns

fig, ax = plt.subplots(figsize=(8, 5))
for region, color in zip(["North", "South", "East"], ["#4C78A8", "#F58518", "#E45756"]):
    subset = df_scatter[df_scatter["region"] == region]
    sns.kdeplot(subset["revenue"], ax=ax, color=color, label=region, fill=True, alpha=0.3)
ax.set_xlabel("Revenue (USD)")
ax.set_title("Revenue Density by Region", fontsize=16, fontweight="bold", loc="left")
ax.legend()
ax.spines[["top", "right"]].set_visible(False)
plt.tight_layout()
plt.show()
```

---

#### Box Plots and Violin Plots

##### Box Plot

**When to use:** Comparing distributions across categories. Showing median, quartiles, and outliers.

**When NOT to use:** Audience unfamiliar with statistical notation. Very small samples (<5 per group).

**Plotly:**

```python
fig = px.box(
    df_scatter, x="region", y="revenue",
    title="Revenue Distribution by Region",
    color="region",
    color_discrete_sequence=["#4C78A8", "#F58518", "#E45756", "#72B7B2"]
)
fig.update_layout(showlegend=False)
fig.show()
```

##### Violin Plot

**When to use:** Same as box plot but also showing the density shape. Better for revealing bimodal distributions.

**Plotly:**

```python
fig = px.violin(
    df_scatter, x="region", y="revenue",
    title="Revenue Distribution by Region (Violin)",
    color="region", box=True, points="outliers",
    color_discrete_sequence=["#4C78A8", "#F58518", "#E45756", "#72B7B2"]
)
fig.update_layout(showlegend=False)
fig.show()
```

---

#### Heatmaps

##### Correlation Heatmap

**When to use:** Showing pairwise correlations in a dataset with multiple numeric variables.

**Plotly:**

```python
import plotly.figure_factory as ff

corr = df_scatter[["ad_spend", "revenue", "employees"]].corr().round(2)
fig = ff.create_annotated_heatmap(
    z=corr.values,
    x=corr.columns.tolist(),
    y=corr.index.tolist(),
    colorscale="RdBu_r",
    showscale=True
)
fig.update_layout(title="Correlation Matrix")
fig.show()
```

**Matplotlib + Seaborn:**

```python
fig, ax = plt.subplots(figsize=(6, 5))
sns.heatmap(corr, annot=True, cmap="RdBu_r", center=0, vmin=-1, vmax=1,
            square=True, linewidths=0.5, ax=ax)
ax.set_title("Correlation Matrix", fontsize=16, fontweight="bold", loc="left")
plt.tight_layout()
plt.show()
```

##### Pivot Heatmap

**When to use:** Two categorical axes, one numeric value. Sales by product and month, activity by day and hour.

**Plotly:**

```python
import plotly.express as px

pivot = pd.DataFrame(
    np.random.randint(10, 100, (7, 24)),
    index=["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"],
    columns=[f"{h}:00" for h in range(24)]
)

fig = px.imshow(
    pivot, title="Activity by Day and Hour",
    color_continuous_scale="Viridis",
    aspect="auto"
)
fig.update_layout(xaxis_title="Hour", yaxis_title="Day")
fig.show()
```

##### Calendar Heatmap

**When to use:** Showing daily values over months or years (GitHub contribution-style).

**Matplotlib (calplot):**

```python
# pip install calplot
import calplot

dates = pd.date_range("2024-01-01", "2024-12-31", freq="D")
values = pd.Series(np.random.randint(0, 10, len(dates)), index=dates)
calplot.calplot(values, cmap="YlGn", figsize=(16, 3))
plt.show()
```

---

#### Pie and Donut Charts

**When to use:** Showing composition of 2-3 categories where exact values are not critical and the chart will be labeled.

**When NOT to use (most of the time):**
- More than 5 categories.
- Comparing across multiple groups (use stacked bar).
- Audience needs to compare precise values (angles are hard to judge).
- Temporal composition (use stacked area).

**If you must use one, prefer a donut chart** -- the center can hold a total or label.

**Plotly:**

```python
df_pie = pd.DataFrame({
    "source": ["Organic", "Paid", "Referral"],
    "traffic": [55, 30, 15]
})

fig = px.pie(
    df_pie, values="traffic", names="source",
    title="Traffic Sources",
    hole=0.45,  # donut
    color_discrete_sequence=["#4C78A8", "#F58518", "#E45756"]
)
fig.update_traces(textinfo="percent+label", textposition="outside")
fig.show()
```

**Better alternative -- horizontal stacked bar:**

```python
fig = px.bar(
    df_pie, x="traffic", y=["Traffic"], color="source",
    orientation="h", barmode="stack",
    title="Traffic Sources",
    color_discrete_sequence=["#4C78A8", "#F58518", "#E45756"],
    text_auto=True
)
fig.update_layout(yaxis_visible=False, showlegend=True, height=200)
fig.show()
```

---

#### Treemaps

**When to use:** Hierarchical part-to-whole. Showing how segments nest within larger groups and their relative sizes.

**When NOT to use:** No hierarchy. Few categories (bar chart is simpler).

**Plotly:**

```python
df_tree = pd.DataFrame({
    "region": ["North", "North", "South", "South", "East", "East"],
    "product": ["Widget A", "Widget B", "Widget A", "Widget B", "Widget A", "Widget B"],
    "revenue": [120, 80, 90, 110, 150, 60]
})

fig = px.treemap(
    df_tree, path=["region", "product"], values="revenue",
    title="Revenue Breakdown",
    color="revenue", color_continuous_scale="Blues"
)
fig.show()
```

---

#### Sunburst Charts

**When to use:** Hierarchical data where you want to show nested composition. Interactive exploration of multi-level categories.

**When NOT to use:** More than 3 levels deep. Non-hierarchical data.

**Plotly:**

```python
fig = px.sunburst(
    df_tree, path=["region", "product"], values="revenue",
    title="Revenue Hierarchy",
    color="revenue", color_continuous_scale="Oranges"
)
fig.show()
```

---

#### Sankey Diagrams

**When to use:** Showing flow between stages (source to destination). Budget allocation, user journey, energy flow.

**When NOT to use:** No clear directional flow. Too many nodes (>20 becomes unreadable).

**Plotly:**

```python
import plotly.graph_objects as go

fig = go.Figure(go.Sankey(
    node=dict(
        pad=20,
        thickness=20,
        label=["Website", "App", "Signup", "Trial", "Purchase", "Churn"],
        color=["#4C78A8", "#F58518", "#54A24B", "#EECA3B", "#E45756", "#B279A2"]
    ),
    link=dict(
        source=[0, 0, 1, 1, 2, 3, 3],
        target=[2, 3, 2, 3, 4, 4, 5],
        value= [40, 30, 20, 25, 35, 30, 25],
        color=["rgba(76,120,168,0.3)"] * 7
    )
))
fig.update_layout(title="User Journey Flow", font_size=12)
fig.show()
```

---

#### Waterfall Charts

**When to use:** Showing how sequential positive and negative values contribute to a final total. Profit bridges, budget walks.

**Plotly:**

```python
fig = go.Figure(go.Waterfall(
    x=["Revenue", "COGS", "Gross Profit", "OpEx", "Tax", "Net Income"],
    y=[500000, -200000, 300000, -150000, -45000, 105000],
    measure=["absolute", "relative", "total", "relative", "relative", "total"],
    connector={"line": {"color": "#888"}},
    increasing={"marker": {"color": "#54A24B"}},
    decreasing={"marker": {"color": "#E45756"}},
    totals={"marker": {"color": "#4C78A8"}}
))
fig.update_layout(title="Profit Waterfall", yaxis_title="USD", yaxis_tickformat="$,.0f")
fig.show()
```

---

#### Funnel Charts

**When to use:** Conversion funnels, pipeline stages, sequential filtering.

**Plotly:**

```python
fig = px.funnel(
    pd.DataFrame({
        "stage": ["Visits", "Signups", "Trials", "Purchases"],
        "count": [10000, 3000, 800, 200]
    }),
    x="count", y="stage",
    title="Conversion Funnel",
    color_discrete_sequence=["#4C78A8"]
)
fig.show()
```

---

#### Radar / Spider Charts

**When to use:** Comparing multivariate profiles (product features, skill assessment, survey dimensions).

**When NOT to use:** More than 8 axes. Axes have different scales. Precise comparison is required.

**Plotly:**

```python
fig = go.Figure()
categories = ["Speed", "Reliability", "Cost", "Support", "Features"]
fig.add_trace(go.Scatterpolar(
    r=[4, 5, 3, 4, 5], theta=categories, fill="toself", name="Product A",
    line_color="#4C78A8"
))
fig.add_trace(go.Scatterpolar(
    r=[3, 3, 5, 2, 4], theta=categories, fill="toself", name="Product B",
    line_color="#F58518"
))
fig.update_layout(
    polar=dict(radialaxis=dict(visible=True, range=[0, 5])),
    title="Product Comparison"
)
fig.show()
```

---

#### Small Multiples / Faceted Plots

**When to use:** Comparing the same chart across categories. Avoiding spaghetti lines. More than 5 series.

**Plotly:**

```python
fig = px.line(
    df_multi, x="date", y="revenue",
    facet_col="region", facet_col_wrap=2,
    title="Revenue Trend by Region",
    color_discrete_sequence=["#4C78A8"]
)
fig.update_yaxes(matches=None, showticklabels=True)
fig.show()
```

**Matplotlib:**

```python
regions = df_multi["region"].unique()
fig, axes = plt.subplots(1, len(regions), figsize=(14, 4), sharey=True)
for ax, region in zip(axes, regions):
    subset = df_multi[df_multi["region"] == region]
    ax.plot(subset["date"], subset["revenue"], color="#4C78A8")
    ax.set_title(region)
    ax.spines[["top", "right"]].set_visible(False)
    ax.tick_params(axis="x", rotation=45)
plt.suptitle("Revenue Trend by Region", fontsize=16, fontweight="bold")
plt.tight_layout()
plt.show()
```

---

#### Sparklines

**When to use:** Inline trend indicators within tables or KPI cards. Showing shape without axis detail.

**Matplotlib:**

```python
def sparkline(data, ax, color="#4C78A8"):
    ax.plot(data, color=color, linewidth=1.5)
    ax.fill_between(range(len(data)), data, alpha=0.1, color=color)
    ax.set_xlim(0, len(data) - 1)
    ax.axis("off")

fig, axes = plt.subplots(4, 1, figsize=(3, 4))
for ax in axes:
    sparkline(np.random.randn(30).cumsum(), ax)
plt.tight_layout()
plt.show()
```

---

#### Bullet Charts

**When to use:** Showing actual vs target with qualitative ranges (poor / satisfactory / good). KPI dashboards.

**Plotly:**

```python
fig = go.Figure(go.Indicator(
    mode="number+gauge+delta",
    value=280,
    delta={"reference": 250},
    gauge={
        "shape": "bullet",
        "axis": {"range": [None, 400]},
        "threshold": {"line": {"color": "#E45756", "width": 3}, "value": 250},
        "steps": [
            {"range": [0, 150], "color": "#E8E8E8"},
            {"range": [150, 300], "color": "#D0D0D0"},
            {"range": [300, 400], "color": "#B8B8B8"}
        ],
        "bar": {"color": "#4C78A8"}
    },
    title={"text": "Revenue (K)"}
))
fig.update_layout(height=150)
fig.show()
```

---

#### Slope Charts

**When to use:** Comparing two time points or conditions. Showing before/after changes and highlighting which items increased or decreased.

**Matplotlib:**

```python
fig, ax = plt.subplots(figsize=(6, 8))
items = ["Product A", "Product B", "Product C", "Product D"]
before = [30, 45, 20, 50]
after = [40, 35, 55, 48]

for i, item in enumerate(items):
    color = "#54A24B" if after[i] > before[i] else "#E45756"
    ax.plot([0, 1], [before[i], after[i]], marker="o", color=color, linewidth=2)
    ax.text(-0.05, before[i], f"{item}: {before[i]}", ha="right", va="center", fontsize=10)
    ax.text(1.05, after[i], f"{after[i]}", ha="left", va="center", fontsize=10)

ax.set_xlim(-0.3, 1.3)
ax.set_xticks([0, 1])
ax.set_xticklabels(["2023", "2024"], fontsize=12)
ax.set_title("Year-over-Year Change", fontsize=16, fontweight="bold", loc="left")
ax.spines[["top", "right", "bottom", "left"]].set_visible(False)
ax.yaxis.set_visible(False)
plt.tight_layout()
plt.show()
```

---

#### Lollipop Charts

**When to use:** Same use case as bar chart but with a lighter visual footprint. Especially good for many categories.

**Matplotlib:**

```python
fig, ax = plt.subplots(figsize=(8, 6))
categories = ["A", "B", "C", "D", "E", "F", "G", "H"]
values = [85, 72, 68, 60, 55, 48, 42, 30]

ax.hlines(categories, 0, values, color="#4C78A8", linewidth=2)
ax.plot(values, categories, "o", color="#4C78A8", markersize=8)
ax.set_xlabel("Score")
ax.set_title("Product Scores", fontsize=16, fontweight="bold", loc="left")
ax.spines[["top", "right"]].set_visible(False)
plt.tight_layout()
plt.show()
```

---

#### Ridgeline Plots (Joy Plots)

**When to use:** Comparing distributions across many categories. More compact than multiple histograms or violins.

**Matplotlib + joypy:**

```python
# pip install joypy
import joypy

fig, axes = joypy.joyplot(
    df_scatter, by="region", column="revenue",
    figsize=(8, 6), alpha=0.6,
    color=["#4C78A8", "#F58518", "#E45756", "#72B7B2"]
)
plt.title("Revenue Distribution by Region", fontsize=16, fontweight="bold")
plt.show()
```

---

#### Waffle Charts

**When to use:** Part-to-whole for small numbers. More readable than pie charts for exact percentages.

**Matplotlib (pywaffle):**

```python
# pip install pywaffle
from pywaffle import Waffle

fig = plt.figure(
    FigureClass=Waffle,
    rows=10, columns=10,
    values=[55, 30, 15],
    labels=["Organic (55%)", "Paid (30%)", "Referral (15%)"],
    colors=["#4C78A8", "#F58518", "#E45756"],
    title={"label": "Traffic Sources", "loc": "left", "fontsize": 16},
    legend={"loc": "lower left", "fontsize": 10}
)
plt.tight_layout()
plt.show()
```

---

#### Geographic Maps

##### Choropleth Map

**When to use:** Regional data on a geographic boundary (country, state, county). Color intensity encodes value.

**Plotly:**

```python
import plotly.express as px

df_geo = pd.DataFrame({
    "state": ["CA", "TX", "NY", "FL", "IL"],
    "value": [500, 400, 350, 300, 200]
})

fig = px.choropleth(
    df_geo, locations="state", locationmode="USA-states",
    color="value", scope="usa",
    title="Sales by State",
    color_continuous_scale="Blues"
)
fig.show()
```

##### Bubble Map

**When to use:** Point data on a map where size encodes magnitude.

**Plotly:**

```python
df_cities = pd.DataFrame({
    "city": ["New York", "Los Angeles", "Chicago", "Houston"],
    "lat": [40.71, 34.05, 41.88, 29.76],
    "lon": [-74.01, -118.24, -87.63, -95.37],
    "sales": [5000, 4200, 3100, 2800]
})

fig = px.scatter_geo(
    df_cities, lat="lat", lon="lon", size="sales",
    scope="usa", title="Sales by City",
    color_discrete_sequence=["#4C78A8"],
    size_max=30
)
fig.show()
```

##### Hex-Bin Map

**When to use:** Dense point data on a map. Aggregating points into hexagonal bins avoids overplotting.

**Plotly (using H3 or manual bins):**

```python
fig = px.density_mapbox(
    df_cities, lat="lat", lon="lon",
    radius=30, zoom=3,
    mapbox_style="carto-positron",
    title="Sales Density"
)
fig.show()
```

---

#### Network Graphs

**When to use:** Showing connections and relationships between entities. Social networks, dependency trees, knowledge graphs.

**When NOT to use:** More than 100 nodes without filtering (becomes a hairball).

**Plotly + NetworkX:**

```python
import networkx as nx
import plotly.graph_objects as go

G = nx.karate_club_graph()
pos = nx.spring_layout(G, seed=42)

edge_x, edge_y = [], []
for edge in G.edges():
    x0, y0 = pos[edge[0]]
    x1, y1 = pos[edge[1]]
    edge_x += [x0, x1, None]
    edge_y += [y0, y1, None]

fig = go.Figure()
fig.add_trace(go.Scatter(x=edge_x, y=edge_y, mode="lines",
                         line=dict(width=0.5, color="#888"), hoverinfo="none"))
fig.add_trace(go.Scatter(
    x=[pos[n][0] for n in G.nodes()],
    y=[pos[n][1] for n in G.nodes()],
    mode="markers",
    marker=dict(size=10, color="#4C78A8", line=dict(width=1, color="white")),
    text=[str(n) for n in G.nodes()],
    hoverinfo="text"
))
fig.update_layout(title="Network Graph", showlegend=False,
                  xaxis=dict(showgrid=False, zeroline=False, visible=False),
                  yaxis=dict(showgrid=False, zeroline=False, visible=False))
fig.show()
```

---

## Color Systems

### Perceptually Uniform Scales

Perceptually uniform color scales ensure that equal steps in data produce equal steps in perceived color difference. Critical for accurate data interpretation.

| Scale | Best for | Character |
|---|---|---|
| **Viridis** | General purpose | Blue-green-yellow, high contrast |
| **Plasma** | Highlighting extremes | Blue-magenta-yellow |
| **Inferno** | Dark backgrounds | Black-red-yellow |
| **Magma** | Heat intensity | Black-purple-yellow |
| **Cividis** | Colorblind-safe | Blue-yellow only |

**Key hex stops (Viridis):**

```
0.0  #440154   (dark purple)
0.25 #31688E   (blue)
0.5  #35B779   (green)
0.75 #90D743   (yellow-green)
1.0  #FDE725   (yellow)
```

**Key hex stops (Plasma):**

```
0.0  #0D0887   (dark blue)
0.25 #7E03A8   (purple)
0.5  #CC4778   (magenta-pink)
0.75 #F89441   (orange)
1.0  #F0F921   (yellow)
```

**Key hex stops (Inferno):**

```
0.0  #000004   (near black)
0.25 #420A68   (dark purple)
0.5  #BC3754   (red)
0.75 #ED7953   (orange)
1.0  #FCFFA4   (pale yellow)
```

### Sequential Palettes

**Single-hue progressions:** Vary lightness within one hue. Good for ordered data where one end is "more."

```
Blues:    #F7FBFF -> #DEEBF7 -> #C6DBEF -> #9ECAE1 -> #6BAED6 -> #3182BD -> #08519C
Greens:  #F7FCF5 -> #E5F5E0 -> #C7E9C0 -> #A1D99B -> #74C476 -> #31A354 -> #006D2C
Oranges: #FFF5EB -> #FEE6CE -> #FDD0A2 -> #FDAE6B -> #FD8D3C -> #E6550D -> #A63603
```

**Multi-hue progressions:** Span two hues for better perceptual range.

```
YlGnBu:  #FFFFD9 -> #EDF8B1 -> #C7E9B4 -> #7FCDBB -> #41B6C4 -> #1D91C0 -> #225EA8 -> #0C2C84
YlOrRd:  #FFFFCC -> #FFEDA0 -> #FED976 -> #FEB24C -> #FD8D3C -> #FC4E2A -> #E31A1C -> #B10026
```

### Diverging Palettes

Use when data has a meaningful center point (zero, average, threshold).

```
RdBu (Red-Blue):
  -1.0  #67001F (dark red)
  -0.5  #D6604D (medium red)
   0.0  #F7F7F7 (near white)
  +0.5  #4393C3 (medium blue)
  +1.0  #053061 (dark blue)

PuOr (Purple-Orange):
  -1.0  #7F3B08 (dark orange)
  -0.5  #E08214 (medium orange)
   0.0  #F7F7F7 (near white)
  +0.5  #8073AC (medium purple)
  +1.0  #2D004B (dark purple)

BrBG (Brown-Teal):
  -1.0  #543005 (dark brown)
  -0.5  #BF812D (medium brown)
   0.0  #F5F5F5 (near white)
  +0.5  #5AB4AC (medium teal)
  +1.0  #003C30 (dark teal)
```

### Qualitative Palettes

For categorical data with no inherent order.

**Category10 (D3/Plotly default):**

```
#1F77B4  (blue)
#FF7F0E  (orange)
#2CA02C  (green)
#D62728  (red)
#9467BD  (purple)
#8C564B  (brown)
#E377C2  (pink)
#7F7F7F  (gray)
#BCBD22  (olive)
#17BECF  (cyan)
```

**Set1 (Brewer):**

```
#E41A1C  (red)
#377EB8  (blue)
#4DAF4A  (green)
#984EA3  (purple)
#FF7F00  (orange)
#FFFF33  (yellow)
#A65628  (brown)
#F781BF  (pink)
#999999  (gray)
```

**Maximum distinguishable categories:** Generally 7-10 for qualitative palettes. Beyond 12, colors become confusable. Use direct labeling, interactive filtering, or grouping instead.

### Colorblind-Safe Palettes

Approximately 8% of men and 0.5% of women have some form of color vision deficiency.

**Okabe-Ito palette (8 colors):**

```
#E69F00  (orange)
#56B4E9  (sky blue)
#009E73  (bluish green)
#F0E442  (yellow)
#0072B2  (blue)
#D55E00  (vermillion)
#CC79A7  (reddish purple)
#000000  (black)
```

**Wong palette:**

```
#000000  (black)
#E69F00  (orange)
#56B4E9  (sky blue)
#009E73  (bluish green)
#F0E442  (yellow)
#0072B2  (blue)
#D55E00  (vermillion)
#CC79A7  (reddish purple)
```

**Tol palette (Paul Tol's qualitative):**

```
#332288  (indigo)
#88CCEE  (cyan)
#44AA99  (teal)
#117733  (green)
#999933  (olive)
#DDCC77  (sand)
#CC6677  (rose)
#882255  (wine)
#AA4499  (purple)
```

**Testing tools:**
- Color Oracle (free, desktop): simulates deuteranopia, protanopia, tritanopia
- Coblis (web): upload an image to simulate color blindness
- `colorspacious` Python library for programmatic simulation
- Chrome DevTools: Rendering > Emulate vision deficiencies

### Brand Color Integration

**Steps to build a data palette from brand colors:**

1. Start with 1-2 brand colors as your primary and secondary data colors.
2. Generate 4-6 additional colors by adjusting hue in 30-degree increments, keeping saturation and lightness similar.
3. Verify WCAG contrast ratios: all data colors must have at least 3:1 contrast against the chart background.
4. Test in grayscale to ensure values are distinguishable without hue.
5. Check with colorblind simulation tools.

```python
# Building a brand palette programmatically
from colorspacious import cspace_convert
import colorsys

brand_primary = "#1A73E8"  # Google blue example

def generate_palette(hex_color, n=6):
    r, g, b = int(hex_color[1:3], 16)/255, int(hex_color[3:5], 16)/255, int(hex_color[5:7], 16)/255
    h, s, v = colorsys.rgb_to_hsv(r, g, b)
    palette = []
    for i in range(n):
        new_h = (h + i * (1.0 / n)) % 1.0
        nr, ng, nb = colorsys.hsv_to_rgb(new_h, s * 0.85, v)
        palette.append(f"#{int(nr*255):02X}{int(ng*255):02X}{int(nb*255):02X}")
    return palette

print(generate_palette(brand_primary))
```

### Dark Mode

**Color adjustments for dark backgrounds:**

- Reduce saturation by 10-20% to avoid vibration against dark backgrounds.
- Use `#1E1E1E` or `#121212` instead of pure `#000000` for backgrounds.
- Use `#E0E0E0` instead of pure `#FFFFFF` for text and grid lines.
- Increase opacity of fills (data areas appear lighter on dark backgrounds).

**Dark mode Plotly template:**

```python
dark_template = dict(
    layout=dict(
        paper_bgcolor="#121212",
        plot_bgcolor="#1E1E1E",
        font=dict(color="#E0E0E0", family="Inter, system-ui, sans-serif"),
        title=dict(font=dict(color="#FFFFFF", size=18)),
        xaxis=dict(gridcolor="#333333", zerolinecolor="#444444"),
        yaxis=dict(gridcolor="#333333", zerolinecolor="#444444"),
        colorway=["#64B5F6", "#FFB74D", "#E57373", "#81C784", "#BA68C8",
                   "#4DD0E1", "#FFD54F", "#A1887F"]
    )
)
```

---

## Typography for Data Viz

### Font Selection

**Recommended system font stack:**

```css
font-family: Inter, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto,
             "Helvetica Neue", Arial, sans-serif;
```

**Recommended monospace (for data tables and axis labels):**

```css
font-family: "JetBrains Mono", "Fira Code", "SF Mono", "Cascadia Code", monospace;
```

**Google Fonts optimized for data visualization:**
- Inter (body text, labels)
- Source Sans Pro / Source Sans 3 (clean, narrow)
- IBM Plex Sans (corporate, clear)
- Roboto Condensed (compact labels)
- Lato (neutral, readable)
- Tabular Oldstyle Figures: use fonts with tabular (fixed-width) numeral support for aligned numbers

### Font Size Hierarchy

| Element | Size (px) | Size (pt) | Weight |
|---|---|---|---|
| Chart title | 18-24 | 14-18 | Bold (700) |
| Subtitle | 14-16 | 10-12 | Regular (400) |
| Axis title | 12-14 | 9-10 | Semi-bold (600) |
| Axis tick labels | 10-12 | 8-9 | Regular (400) |
| Data labels | 10-12 | 8-9 | Medium (500) |
| Annotations | 11-13 | 8-10 | Regular (400) |
| Caption/source | 9-10 | 7-8 | Regular (400), italic |
| Legend | 10-12 | 8-9 | Regular (400) |

### Number Formatting

| Format | Example | Use case |
|---|---|---|
| Thousands separator | 1,234,567 | Exact values |
| Abbreviated | 1.2M | Dashboard labels |
| Currency | $1,234 | Financial data |
| Currency abbreviated | $1.2M | Dashboard KPIs |
| Percentage | 45.2% | Rates, proportions |
| Decimal: 0 places | 1,235 | Counts, integers |
| Decimal: 1 place | 45.2% | Percentages |
| Decimal: 2 places | $12.34 | Currency |
| Scientific | 1.23e6 | Technical audience |

```python
# Python formatting examples
f"{1234567:,.0f}"        # "1,234,567"
f"{1234567/1e6:.1f}M"    # "1.2M"
f"${1234:.2f}"           # "$1234.00"
f"{0.452:.1%}"           # "45.2%"
```

---

## Layout and Composition

### Grid Systems for Dashboards

**12-column grid (CSS):**

```css
.dashboard-grid {
    display: grid;
    grid-template-columns: repeat(12, 1fr);
    gap: 16px;
    padding: 24px;
}

.card-full    { grid-column: span 12; }
.card-half    { grid-column: span 6; }
.card-third   { grid-column: span 4; }
.card-quarter { grid-column: span 3; }
.card-two-thirds { grid-column: span 8; }
```

**Common dashboard layouts:**

```
+-------------------+-------------------+
|   KPI Card (3)    |   KPI Card (3)    |    Row 1: KPI summary
+---+---+---+---+---+---+---+---+---+---+
|   KPI (3)  |   KPI (3)  |   KPI (3)   |
+------------+---------------------------+
|            |                           |    Row 2: Primary viz
|  Filter    |     Main Chart (8)       |
|  Panel (4) |                           |
+------------+---------------------------+
|     Chart (6)     |     Chart (6)     |    Row 3: Supporting viz
+-------------------+-------------------+
```

### Whitespace Principles

- **Macro whitespace:** Margins around the entire chart (minimum 10% of chart width).
- **Micro whitespace:** Padding inside chart elements (bar gaps, label padding).
- **Breathing room:** Space between title and chart area (8-16px).
- **Bar gap ratio:** Bar width to gap should be approximately 2:1 for grouped bars.
- **Never let data touch the axis borders** -- add 5-10% padding to data range.

### Visual Hierarchy

1. **Title** -- the first thing the reader sees
2. **Data marks** -- bars, lines, points (primary visual)
3. **Annotations** -- callouts on specific data points
4. **Axes** -- reference frame (should recede)
5. **Gridlines** -- subtle reference (lightest element)
6. **Source/footnotes** -- smallest and least prominent

### Aspect Ratios

| Chart type | Recommended ratio | Notes |
|---|---|---|
| Line chart | 3:1 to 2:1 (wide) | Wide formats emphasize trend slope |
| Bar chart (vertical) | 1:1 to 4:3 | Near-square for few categories |
| Bar chart (horizontal) | Height varies with categories | 20-30px per bar |
| Scatter plot | 1:1 | Square preserves correlation perception |
| Heatmap | Depends on data dimensions | Match rows:cols ratio |
| Sparkline | 4:1 to 8:1 (very wide) | Inline with text |
| Map | Match geographic bounds | Do not stretch |

### Responsive Breakpoints

| Breakpoint | Width | Adjustments |
|---|---|---|
| Desktop XL | >1440px | Full dashboard grid, all labels |
| Desktop | 1024-1440px | Standard layout |
| Tablet | 768-1023px | Stack 2-column to 1-column, hide secondary axes |
| Mobile | <768px | Single column, simplified charts, larger touch targets |
| Mobile S | <375px | Sparklines only, remove legends, use tooltips |

---

## Annotation and Storytelling

### Title Writing

**Descriptive title** (states what the chart shows):
> "Monthly Revenue by Region, 2024"

**Declarative title** (states the insight -- preferred for presentations):
> "East Region Revenue Grew 35% While West Declined"

**Rules:**
- Use sentence case, not title case.
- Keep under 10 words for the title line.
- Use a subtitle for context, date range, or methodology note.

### Axis Labeling

- Include units in the axis title: "Revenue (USD thousands)" or "Temperature (C)."
- Remove axis title if the data labels make it obvious.
- Rotate x-axis labels to 45 degrees maximum; prefer horizontal bars if labels are long.
- Use abbreviations on tick labels: "Jan," "Feb" not "January," "February."

### Code Examples for Annotations

**Plotly annotations:**

```python
fig.add_annotation(
    x="2024-06-01", y=480000,
    text="Product launch",
    showarrow=True,
    arrowhead=2,
    arrowsize=1,
    arrowwidth=1.5,
    arrowcolor="#666",
    ax=-40, ay=-40,
    font=dict(size=12, color="#333"),
    bgcolor="white",
    bordercolor="#ccc",
    borderwidth=1,
    borderpad=4
)

# Reference line
fig.add_hline(
    y=350000,
    line_dash="dash",
    line_color="#E45756",
    annotation_text="Target: $350K",
    annotation_position="top right"
)

# Reference band
fig.add_vrect(
    x0="2024-03-01", x1="2024-06-01",
    fillcolor="#4C78A8", opacity=0.08,
    line_width=0,
    annotation_text="Q2", annotation_position="top left"
)
```

**Matplotlib annotations:**

```python
ax.annotate(
    "Product launch",
    xy=(pd.Timestamp("2024-06-01"), 480000),
    xytext=(-60, 30),
    textcoords="offset points",
    arrowprops=dict(arrowstyle="->", color="#666", lw=1.5),
    fontsize=11,
    color="#333",
    bbox=dict(boxstyle="round,pad=0.3", facecolor="white", edgecolor="#ccc")
)

# Reference line
ax.axhline(y=350000, color="#E45756", linestyle="--", linewidth=1, label="Target")

# Reference band
ax.axvspan(pd.Timestamp("2024-03-01"), pd.Timestamp("2024-06-01"),
           alpha=0.08, color="#4C78A8", label="Q2")
```

---

## Responsive Design

### SVG viewBox Approach

```html
<svg viewBox="0 0 800 400" preserveAspectRatio="xMidYMid meet"
     style="width: 100%; height: auto; max-width: 800px;">
  <!-- chart content -->
</svg>
```

### Container-Relative Sizing (D3.js)

```javascript
function responsiveChart(containerSelector) {
    const container = d3.select(containerSelector);
    const width = container.node().getBoundingClientRect().width;
    const height = width * 0.5; // 2:1 aspect ratio
    const margin = { top: 40, right: 20, bottom: 40, left: 60 };
    const innerWidth = width - margin.left - margin.right;
    const innerHeight = height - margin.top - margin.bottom;

    const svg = container.append("svg")
        .attr("viewBox", `0 0 ${width} ${height}`)
        .attr("preserveAspectRatio", "xMidYMid meet")
        .style("width", "100%")
        .style("height", "auto");

    const g = svg.append("g")
        .attr("transform", `translate(${margin.left},${margin.top})`);

    return { svg, g, innerWidth, innerHeight, margin };
}
```

### Mobile-First Chart Design

**Principles:**
1. Start with the smallest viewport, add complexity as width increases.
2. On mobile (<768px): hide legends, use tooltips instead. Increase touch targets to 44x44px.
3. At tablet (768-1023px): show legends below chart. Use abbreviated labels.
4. At desktop (>1024px): full labels, side legends, richer annotations.

```css
/* Chart container responsive styles */
.chart-container {
    width: 100%;
    max-width: 1200px;
    margin: 0 auto;
}

.chart-container svg text {
    font-size: 12px;
}

@media (max-width: 768px) {
    .chart-container svg text { font-size: 10px; }
    .chart-legend { display: none; }
    .chart-tooltip { font-size: 14px; padding: 12px; }
}

@media (min-width: 1024px) {
    .chart-container svg text { font-size: 14px; }
}
```

---

## Accessibility (WCAG 2.1 AA)

### Color Contrast Requirements

| Element | Minimum contrast ratio | Standard |
|---|---|---|
| Normal text (<18px) | 4.5:1 | WCAG AA |
| Large text (>=18px bold, >=24px regular) | 3:1 | WCAG AA |
| Graphical objects (chart elements) | 3:1 | WCAG 1.4.11 |
| Adjacent data series | 3:1 between each other | Best practice |

**Testing contrast:**

```python
def contrast_ratio(hex1, hex2):
    """Calculate WCAG contrast ratio between two hex colors."""
    def luminance(hex_color):
        r, g, b = [int(hex_color[i:i+2], 16) / 255.0 for i in (1, 3, 5)]
        channels = []
        for c in (r, g, b):
            channels.append(c / 12.92 if c <= 0.03928 else ((c + 0.055) / 1.055) ** 2.4)
        return 0.2126 * channels[0] + 0.7152 * channels[1] + 0.0722 * channels[2]

    l1 = luminance(hex1)
    l2 = luminance(hex2)
    lighter = max(l1, l2)
    darker = min(l1, l2)
    return (lighter + 0.05) / (darker + 0.05)

# Example
ratio = contrast_ratio("#4C78A8", "#FFFFFF")
print(f"Contrast ratio: {ratio:.2f}:1")  # Should be >= 3:1 for graphics
```

### Alt Text for Charts

**Formula:** `[Chart type] showing [what data] for [what context]. [Key insight].`

**Examples:**

```html
<img src="chart.png"
     alt="Bar chart showing quarterly revenue for 4 regions in 2024.
          East region leads at $510K, followed by North ($420K),
          South ($380K), and West ($290K).">

<img src="trend.png"
     alt="Line chart showing monthly active users from January to December 2024.
          Users grew steadily from 12,000 to 45,000, with a sharp increase
          in June following the product launch.">
```

### ARIA Labels for Interactive Charts

```html
<div role="figure" aria-label="Revenue comparison by region">
    <svg role="img" aria-describedby="chart-desc">
        <desc id="chart-desc">
            Bar chart comparing 2024 revenue across four regions.
            East: $510,000. North: $420,000. South: $380,000. West: $290,000.
        </desc>
        <!-- chart SVG content -->
    </svg>
</div>
```

### Keyboard Navigation for Interactive Charts

```javascript
// D3.js keyboard navigation pattern
svg.selectAll(".bar")
    .attr("tabindex", 0)
    .attr("role", "img")
    .attr("aria-label", d => `${d.category}: ${d.value}`)
    .on("keydown", function(event, d) {
        if (event.key === "Enter" || event.key === " ") {
            showTooltip(d);
            event.preventDefault();
        }
        if (event.key === "ArrowRight") {
            const next = this.nextElementSibling;
            if (next) next.focus();
        }
        if (event.key === "ArrowLeft") {
            const prev = this.previousElementSibling;
            if (prev) prev.focus();
        }
    });
```

### Pattern Fills for Colorblind Users

```javascript
// D3.js pattern definitions
const defs = svg.append("defs");

// Diagonal lines
defs.append("pattern")
    .attr("id", "diag-lines")
    .attr("patternUnits", "userSpaceOnUse")
    .attr("width", 8).attr("height", 8)
    .append("path")
    .attr("d", "M0,8 L8,0")
    .attr("stroke", "#333").attr("stroke-width", 1.5);

// Dots
defs.append("pattern")
    .attr("id", "dots")
    .attr("patternUnits", "userSpaceOnUse")
    .attr("width", 6).attr("height", 6)
    .append("circle")
    .attr("cx", 3).attr("cy", 3).attr("r", 1.5)
    .attr("fill", "#333");

// Crosshatch
defs.append("pattern")
    .attr("id", "crosshatch")
    .attr("patternUnits", "userSpaceOnUse")
    .attr("width", 8).attr("height", 8)
    .selectAll("path")
    .data(["M0,8 L8,0", "M0,0 L8,8"])
    .enter().append("path")
    .attr("d", d => d)
    .attr("stroke", "#333").attr("stroke-width", 1);
```

### Data Tables as Alternatives

Always provide a data table as an alternative or supplement to charts for screen reader users.

```html
<details>
    <summary>View data table</summary>
    <table>
        <caption>Revenue by Region, 2024</caption>
        <thead>
            <tr>
                <th scope="col">Region</th>
                <th scope="col">Revenue (USD)</th>
                <th scope="col">% of Total</th>
            </tr>
        </thead>
        <tbody>
            <tr><td>East</td><td>$510,000</td><td>32%</td></tr>
            <tr><td>North</td><td>$420,000</td><td>26%</td></tr>
            <tr><td>South</td><td>$380,000</td><td>24%</td></tr>
            <tr><td>West</td><td>$290,000</td><td>18%</td></tr>
        </tbody>
    </table>
</details>
```

---

## Platform Templates

### Plotly Custom Theme

```python
import plotly.graph_objects as go
import plotly.io as pio

custom_template = go.layout.Template(
    layout=go.Layout(
        # Colors
        colorway=["#4C78A8", "#F58518", "#E45756", "#72B7B2",
                  "#54A24B", "#EECA3B", "#B279A2", "#FF9DA6",
                  "#9D755D", "#BAB0AC"],

        # Fonts
        font=dict(
            family="Inter, -apple-system, BlinkMacSystemFont, sans-serif",
            size=13,
            color="#333333"
        ),
        title=dict(
            font=dict(size=18, color="#1a1a1a"),
            x=0,
            xanchor="left",
            y=0.98
        ),

        # Background
        paper_bgcolor="#FFFFFF",
        plot_bgcolor="#FAFAFA",

        # Axes
        xaxis=dict(
            showgrid=False,
            showline=True,
            linecolor="#D0D0D0",
            tickfont=dict(size=11),
            title=dict(font=dict(size=13), standoff=12)
        ),
        yaxis=dict(
            showgrid=True,
            gridcolor="#EBEBEB",
            gridwidth=1,
            showline=False,
            tickfont=dict(size=11),
            title=dict(font=dict(size=13), standoff=12)
        ),

        # Legend
        legend=dict(
            orientation="h",
            yanchor="bottom",
            y=1.02,
            xanchor="left",
            x=0,
            font=dict(size=11),
            bgcolor="rgba(255,255,255,0)"
        ),

        # Margins
        margin=dict(l=60, r=20, t=80, b=50),

        # Hover
        hoverlabel=dict(
            bgcolor="white",
            bordercolor="#ccc",
            font=dict(size=12, family="Inter, sans-serif")
        ),

        # Colorscale defaults
        colorscale=dict(
            sequential=[[0, "#F7FBFF"], [0.5, "#6BAED6"], [1, "#08306B"]],
            diverging=[[0, "#D73027"], [0.5, "#FFFFBF"], [1, "#1A9850"]]
        )
    )
)

# Register the template
pio.templates["custom_clean"] = custom_template
pio.templates.default = "custom_clean"
```

### Matplotlib Custom Style

```python
import matplotlib as mpl
import matplotlib.pyplot as plt

# Define the style
custom_style = {
    # Figure
    "figure.figsize": (10, 6),
    "figure.dpi": 150,
    "figure.facecolor": "white",
    "figure.edgecolor": "white",

    # Font
    "font.family": "sans-serif",
    "font.sans-serif": ["Inter", "Helvetica Neue", "Arial", "sans-serif"],
    "font.size": 12,

    # Axes
    "axes.facecolor": "#FAFAFA",
    "axes.edgecolor": "#D0D0D0",
    "axes.linewidth": 0.8,
    "axes.grid": True,
    "axes.grid.axis": "y",
    "axes.titlesize": 16,
    "axes.titleweight": "bold",
    "axes.titlepad": 16,
    "axes.labelsize": 12,
    "axes.labelpad": 8,
    "axes.spines.top": False,
    "axes.spines.right": False,
    "axes.prop_cycle": plt.cycler(color=[
        "#4C78A8", "#F58518", "#E45756", "#72B7B2",
        "#54A24B", "#EECA3B", "#B279A2", "#FF9DA6",
        "#9D755D", "#BAB0AC"
    ]),

    # Grid
    "grid.color": "#EBEBEB",
    "grid.linewidth": 0.8,
    "grid.alpha": 1.0,

    # Ticks
    "xtick.labelsize": 10,
    "ytick.labelsize": 10,
    "xtick.major.size": 0,
    "ytick.major.size": 0,

    # Legend
    "legend.frameon": False,
    "legend.fontsize": 10,

    # Lines
    "lines.linewidth": 2,
    "lines.markersize": 6,

    # Patches (bars, etc)
    "patch.edgecolor": "white",
    "patch.linewidth": 0.5,

    # Saving
    "savefig.dpi": 300,
    "savefig.bbox": "tight",
    "savefig.pad_inches": 0.2,
}

# Apply globally
mpl.rcParams.update(custom_style)

# Or save as a .mplstyle file and use:
# plt.style.use("path/to/custom.mplstyle")
```

### D3.js Reusable Chart Pattern

```javascript
// Reusable bar chart component
function barChart() {
    // Configuration with defaults
    let width = 600;
    let height = 400;
    let margin = { top: 40, right: 20, bottom: 50, left: 60 };
    let xValue = d => d.category;
    let yValue = d => d.value;
    let color = "#4C78A8";
    let title = "";

    function chart(selection) {
        selection.each(function(data) {
            const innerWidth = width - margin.left - margin.right;
            const innerHeight = height - margin.top - margin.bottom;

            // Create or select SVG
            let svg = d3.select(this).selectAll("svg").data([null]);
            svg = svg.enter().append("svg")
                .attr("viewBox", `0 0 ${width} ${height}`)
                .attr("preserveAspectRatio", "xMidYMid meet")
                .style("width", "100%")
                .style("height", "auto")
                .merge(svg);

            let g = svg.selectAll(".chart-g").data([null]);
            g = g.enter().append("g")
                .attr("class", "chart-g")
                .attr("transform", `translate(${margin.left},${margin.top})`)
                .merge(g);

            // Scales
            const xScale = d3.scaleBand()
                .domain(data.map(xValue))
                .range([0, innerWidth])
                .padding(0.3);

            const yScale = d3.scaleLinear()
                .domain([0, d3.max(data, yValue) * 1.1])
                .range([innerHeight, 0]);

            // Axes
            const xAxis = g.selectAll(".x-axis").data([null]);
            xAxis.enter().append("g")
                .attr("class", "x-axis")
                .attr("transform", `translate(0,${innerHeight})`)
                .merge(xAxis)
                .call(d3.axisBottom(xScale).tickSize(0))
                .select(".domain").attr("stroke", "#D0D0D0");

            const yAxis = g.selectAll(".y-axis").data([null]);
            yAxis.enter().append("g")
                .attr("class", "y-axis")
                .merge(yAxis)
                .call(d3.axisLeft(yScale).ticks(5).tickFormat(d3.format("$,.0f")))
                .select(".domain").remove();

            // Gridlines
            const gridlines = g.selectAll(".gridline").data(yScale.ticks(5));
            gridlines.enter().append("line")
                .attr("class", "gridline")
                .merge(gridlines)
                .attr("x1", 0).attr("x2", innerWidth)
                .attr("y1", d => yScale(d)).attr("y2", d => yScale(d))
                .attr("stroke", "#EBEBEB").attr("stroke-width", 1);
            gridlines.exit().remove();

            // Bars
            const bars = g.selectAll(".bar").data(data);
            bars.enter().append("rect")
                .attr("class", "bar")
                .attr("tabindex", 0)
                .attr("role", "img")
                .merge(bars)
                .transition().duration(500)
                .attr("x", d => xScale(xValue(d)))
                .attr("y", d => yScale(yValue(d)))
                .attr("width", xScale.bandwidth())
                .attr("height", d => innerHeight - yScale(yValue(d)))
                .attr("fill", color)
                .attr("rx", 2);
            bars.exit().remove();

            // Title
            const titleEl = svg.selectAll(".chart-title").data([null]);
            titleEl.enter().append("text")
                .attr("class", "chart-title")
                .merge(titleEl)
                .attr("x", margin.left)
                .attr("y", 24)
                .attr("font-size", "16px")
                .attr("font-weight", "bold")
                .attr("fill", "#1a1a1a")
                .text(title);

            // Tooltip
            const tooltip = d3.select("body").selectAll(".chart-tooltip").data([null]);
            const tooltipEl = tooltip.enter().append("div")
                .attr("class", "chart-tooltip")
                .style("position", "absolute")
                .style("display", "none")
                .style("background", "white")
                .style("border", "1px solid #ccc")
                .style("border-radius", "4px")
                .style("padding", "8px 12px")
                .style("font-size", "12px")
                .style("pointer-events", "none")
                .style("box-shadow", "0 2px 4px rgba(0,0,0,0.1)")
                .merge(tooltip);

            g.selectAll(".bar")
                .on("mouseenter", function(event, d) {
                    d3.select(this).attr("opacity", 0.8);
                    tooltipEl.style("display", "block")
                        .html(`<strong>${xValue(d)}</strong><br>$${yValue(d).toLocaleString()}`);
                })
                .on("mousemove", function(event) {
                    tooltipEl.style("left", (event.pageX + 12) + "px")
                        .style("top", (event.pageY - 20) + "px");
                })
                .on("mouseleave", function() {
                    d3.select(this).attr("opacity", 1);
                    tooltipEl.style("display", "none");
                });
        });
    }

    // Getter/setter methods (chainable API)
    chart.width = function(_) { return arguments.length ? (width = _, chart) : width; };
    chart.height = function(_) { return arguments.length ? (height = _, chart) : height; };
    chart.margin = function(_) { return arguments.length ? (margin = _, chart) : margin; };
    chart.xValue = function(_) { return arguments.length ? (xValue = _, chart) : xValue; };
    chart.yValue = function(_) { return arguments.length ? (yValue = _, chart) : yValue; };
    chart.color = function(_) { return arguments.length ? (color = _, chart) : color; };
    chart.title = function(_) { return arguments.length ? (title = _, chart) : title; };

    return chart;
}

// Usage
const myChart = barChart()
    .width(800)
    .height(450)
    .xValue(d => d.region)
    .yValue(d => d.revenue)
    .color("#4C78A8")
    .title("Revenue by Region");

d3.select("#chart-container")
    .datum(data)
    .call(myChart);
```

---

## Dashboard Design

### KPI Card Pattern

**HTML/CSS:**

```html
<div class="kpi-card">
    <div class="kpi-label">Monthly Revenue</div>
    <div class="kpi-value">$1.24M</div>
    <div class="kpi-delta positive">+12.3% vs last month</div>
    <div class="kpi-sparkline" id="spark-revenue"></div>
</div>

<style>
.kpi-card {
    background: #FFFFFF;
    border: 1px solid #E5E7EB;
    border-radius: 8px;
    padding: 20px 24px;
    min-width: 220px;
}

.kpi-label {
    font-size: 13px;
    color: #6B7280;
    font-weight: 500;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    margin-bottom: 4px;
}

.kpi-value {
    font-size: 32px;
    font-weight: 700;
    color: #111827;
    line-height: 1.2;
    font-variant-numeric: tabular-nums;
}

.kpi-delta {
    font-size: 13px;
    margin-top: 4px;
}

.kpi-delta.positive { color: #059669; }
.kpi-delta.negative { color: #DC2626; }

.kpi-sparkline {
    margin-top: 12px;
    height: 32px;
}
</style>
```

**Plotly Dash KPI card:**

```python
import dash
from dash import html, dcc
import plotly.graph_objects as go

def kpi_card(title, value, delta, spark_data):
    spark_fig = go.Figure(go.Scatter(
        y=spark_data, mode="lines", line=dict(color="#4C78A8", width=1.5),
        fill="tozeroy", fillcolor="rgba(76,120,168,0.1)"
    ))
    spark_fig.update_layout(
        margin=dict(l=0, r=0, t=0, b=0), height=40,
        xaxis=dict(visible=False), yaxis=dict(visible=False),
        paper_bgcolor="rgba(0,0,0,0)", plot_bgcolor="rgba(0,0,0,0)"
    )

    delta_color = "#059669" if delta >= 0 else "#DC2626"
    delta_sign = "+" if delta >= 0 else ""

    return html.Div([
        html.Div(title, className="kpi-label"),
        html.Div(value, className="kpi-value"),
        html.Div(f"{delta_sign}{delta}%", style={"color": delta_color, "fontSize": "13px"}),
        dcc.Graph(figure=spark_fig, config={"displayModeBar": False}, style={"height": "40px"})
    ], className="kpi-card")
```

### Plotly Dash Dashboard Template

```python
import dash
from dash import html, dcc, callback, Output, Input
import plotly.express as px
import pandas as pd

app = dash.Dash(__name__)

app.layout = html.Div([
    # Header
    html.Div([
        html.H1("Sales Dashboard", style={"margin": 0, "fontSize": "24px"}),
        html.P("Updated daily", style={"color": "#6B7280", "margin": "4px 0 0"})
    ], style={"padding": "24px", "borderBottom": "1px solid #E5E7EB"}),

    # Filters
    html.Div([
        html.Div([
            html.Label("Region"),
            dcc.Dropdown(
                id="region-filter",
                options=[{"label": r, "value": r} for r in ["All", "North", "South", "East", "West"]],
                value="All",
                clearable=False
            )
        ], style={"width": "200px"}),
        html.Div([
            html.Label("Date Range"),
            dcc.DatePickerRange(id="date-range")
        ])
    ], style={"display": "flex", "gap": "16px", "padding": "16px 24px",
              "backgroundColor": "#F9FAFB", "borderBottom": "1px solid #E5E7EB"}),

    # KPI row
    html.Div(id="kpi-row", style={
        "display": "grid", "gridTemplateColumns": "repeat(4, 1fr)",
        "gap": "16px", "padding": "24px"
    }),

    # Charts row
    html.Div([
        html.Div([
            dcc.Graph(id="main-chart")
        ], style={"gridColumn": "span 8"}),
        html.Div([
            dcc.Graph(id="side-chart")
        ], style={"gridColumn": "span 4"})
    ], style={
        "display": "grid", "gridTemplateColumns": "repeat(12, 1fr)",
        "gap": "16px", "padding": "0 24px 24px"
    })
], style={"fontFamily": "Inter, sans-serif", "backgroundColor": "#FFFFFF"})


@callback(
    Output("main-chart", "figure"),
    Input("region-filter", "value")
)
def update_main_chart(region):
    # Replace with your data source
    df = pd.DataFrame({
        "month": pd.date_range("2024-01-01", periods=12, freq="MS"),
        "revenue": [320, 340, 380, 410, 450, 480, 460, 490, 520, 510, 540, 580]
    })
    fig = px.line(df, x="month", y="revenue", title="Revenue Trend", markers=True)
    fig.update_layout(template="plotly_white")
    return fig


if __name__ == "__main__":
    app.run(debug=True)
```

### HTML/CSS Dashboard Template

```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Dashboard</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }

        body {
            font-family: Inter, -apple-system, BlinkMacSystemFont, sans-serif;
            background: #F3F4F6;
            color: #111827;
        }

        .dashboard {
            max-width: 1400px;
            margin: 0 auto;
            padding: 24px;
        }

        .dashboard-header {
            margin-bottom: 24px;
        }

        .dashboard-header h1 {
            font-size: 24px;
            font-weight: 700;
        }

        .dashboard-header p {
            color: #6B7280;
            font-size: 14px;
            margin-top: 4px;
        }

        .kpi-row {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
            gap: 16px;
            margin-bottom: 24px;
        }

        .kpi-card {
            background: #FFFFFF;
            border: 1px solid #E5E7EB;
            border-radius: 8px;
            padding: 20px 24px;
        }

        .chart-grid {
            display: grid;
            grid-template-columns: 2fr 1fr;
            gap: 16px;
            margin-bottom: 24px;
        }

        .chart-card {
            background: #FFFFFF;
            border: 1px solid #E5E7EB;
            border-radius: 8px;
            padding: 20px;
        }

        .chart-card h3 {
            font-size: 14px;
            font-weight: 600;
            color: #374151;
            margin-bottom: 16px;
        }

        .chart-placeholder {
            background: #F9FAFB;
            border-radius: 4px;
            height: 300px;
            display: flex;
            align-items: center;
            justify-content: center;
            color: #9CA3AF;
        }

        .chart-row {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 16px;
        }

        @media (max-width: 768px) {
            .chart-grid { grid-template-columns: 1fr; }
            .chart-row { grid-template-columns: 1fr; }
            .kpi-row { grid-template-columns: repeat(2, 1fr); }
        }

        @media (max-width: 480px) {
            .kpi-row { grid-template-columns: 1fr; }
            .dashboard { padding: 16px; }
        }
    </style>
</head>
<body>
    <div class="dashboard">
        <div class="dashboard-header">
            <h1>Sales Dashboard</h1>
            <p>Last updated: March 19, 2026</p>
        </div>

        <div class="kpi-row">
            <div class="kpi-card">
                <div style="font-size:13px;color:#6B7280;text-transform:uppercase;letter-spacing:0.05em">Revenue</div>
                <div style="font-size:32px;font-weight:700;margin-top:4px">$1.24M</div>
                <div style="font-size:13px;color:#059669;margin-top:4px">+12.3% vs last month</div>
            </div>
            <div class="kpi-card">
                <div style="font-size:13px;color:#6B7280;text-transform:uppercase;letter-spacing:0.05em">Customers</div>
                <div style="font-size:32px;font-weight:700;margin-top:4px">3,847</div>
                <div style="font-size:13px;color:#059669;margin-top:4px">+5.1% vs last month</div>
            </div>
            <div class="kpi-card">
                <div style="font-size:13px;color:#6B7280;text-transform:uppercase;letter-spacing:0.05em">Avg Order</div>
                <div style="font-size:32px;font-weight:700;margin-top:4px">$322</div>
                <div style="font-size:13px;color:#DC2626;margin-top:4px">-2.4% vs last month</div>
            </div>
            <div class="kpi-card">
                <div style="font-size:13px;color:#6B7280;text-transform:uppercase;letter-spacing:0.05em">Conversion</div>
                <div style="font-size:32px;font-weight:700;margin-top:4px">4.2%</div>
                <div style="font-size:13px;color:#059669;margin-top:4px">+0.8pp vs last month</div>
            </div>
        </div>

        <div class="chart-grid">
            <div class="chart-card">
                <h3>Revenue Trend</h3>
                <div class="chart-placeholder" id="main-chart">Chart renders here</div>
            </div>
            <div class="chart-card">
                <h3>Revenue by Region</h3>
                <div class="chart-placeholder" id="side-chart">Chart renders here</div>
            </div>
        </div>

        <div class="chart-row">
            <div class="chart-card">
                <h3>Top Products</h3>
                <div class="chart-placeholder" id="products-chart">Chart renders here</div>
            </div>
            <div class="chart-card">
                <h3>Customer Segments</h3>
                <div class="chart-placeholder" id="segments-chart">Chart renders here</div>
            </div>
        </div>
    </div>
</body>
</html>
```

---

## Export and Delivery

### PNG Export Settings

**Plotly:**

```python
fig.write_image(
    "chart.png",
    width=1200,       # pixels
    height=600,
    scale=2,          # 2x for retina / print (effectively 2400x1200)
    engine="kaleido"  # requires pip install kaleido
)
```

**Matplotlib:**

```python
fig.savefig(
    "chart.png",
    dpi=300,           # 300 DPI for print
    bbox_inches="tight",
    pad_inches=0.2,
    facecolor="white",
    transparent=False
)
```

**Recommended settings by use case:**

| Use case | Width (px) | Height (px) | DPI / Scale |
|---|---|---|---|
| Slide (16:9) | 1920 | 1080 | scale=2 |
| Report (A4 half-page) | 1600 | 900 | 300 DPI |
| Social media | 1200 | 628 | scale=2 |
| Dashboard tile | 600 | 400 | scale=1 |
| Email inline | 800 | 500 | scale=1 |
| Print poster | 3000 | 2000 | 300 DPI |

### SVG Optimization

```python
# Plotly SVG export
fig.write_image("chart.svg")

# Matplotlib SVG export
fig.savefig("chart.svg", format="svg", bbox_inches="tight")
```

**Post-processing with svgo (Node.js tool):**

```bash
# Install: npm install -g svgo
svgo chart.svg -o chart.min.svg --multipass
```

### PDF Generation

```python
# Plotly
fig.write_image("chart.pdf", width=800, height=500, engine="kaleido")

# Matplotlib
from matplotlib.backends.backend_pdf import PdfPages

with PdfPages("report.pdf") as pdf:
    fig1, ax1 = plt.subplots()
    ax1.plot([1, 2, 3], [1, 4, 9])
    ax1.set_title("Chart 1")
    pdf.savefig(fig1, bbox_inches="tight")

    fig2, ax2 = plt.subplots()
    ax2.bar(["A", "B", "C"], [3, 7, 5])
    ax2.set_title("Chart 2")
    pdf.savefig(fig2, bbox_inches="tight")
```

### Interactive HTML Embedding

```python
# Plotly -- standalone HTML file
fig.write_html(
    "chart.html",
    include_plotlyjs="cdn",   # use CDN instead of bundling 3MB library
    full_html=True,
    config={
        "displayModeBar": True,
        "modeBarButtonsToRemove": ["lasso2d", "select2d"],
        "displaylogo": False,
        "responsive": True
    }
)

# Plotly -- HTML div for embedding in a page
div_html = fig.to_html(
    include_plotlyjs=False,  # assume plotly.js is loaded separately
    full_html=False,
    div_id="my-chart"
)
```

### Responsive iframe Embedding

```html
<!-- Container with aspect ratio -->
<div style="position: relative; padding-bottom: 56.25%; height: 0; overflow: hidden;">
    <iframe
        src="chart.html"
        style="position: absolute; top: 0; left: 0; width: 100%; height: 100%; border: none;"
        title="Revenue chart"
        loading="lazy">
    </iframe>
</div>
```

### Print Stylesheet

```css
@media print {
    body { background: white; }

    .dashboard { max-width: 100%; padding: 0; }

    .kpi-row {
        grid-template-columns: repeat(4, 1fr);
        gap: 8px;
    }

    .chart-card {
        break-inside: avoid;
        page-break-inside: avoid;
        border: 1px solid #ccc;
    }

    .chart-placeholder {
        /* Ensure chart images print at full quality */
        -webkit-print-color-adjust: exact;
        print-color-adjust: exact;
    }

    /* Hide interactive-only elements */
    .filter-panel, .tooltip, button { display: none !important; }

    /* Force backgrounds to print */
    .kpi-card {
        -webkit-print-color-adjust: exact;
        print-color-adjust: exact;
    }

    @page {
        margin: 1cm;
        size: A4 landscape;
    }
}
```

---

## Anti-Patterns

### Truncated Y-Axes

**Problem:** Starting the y-axis at a value other than zero exaggerates differences in bar charts.

**When it is acceptable:** Line charts showing a narrow range of change (stock prices, temperature). Always label the axis clearly and consider adding a break indicator.

**When it is never acceptable:** Bar charts. The visual area of the bar encodes the value. A truncated baseline makes a 2% difference look like a 50% difference.

```python
# BAD: truncated bar chart
fig = px.bar(df, x="region", y="revenue")
fig.update_yaxes(range=[250000, 520000])  # DO NOT DO THIS

# GOOD: start at zero
fig = px.bar(df, x="region", y="revenue")
fig.update_yaxes(range=[0, None])  # Plotly default, but be explicit
```

### 3D Charts

**Problem:** 3D perspective distorts areas, angles, and lengths. Bars in the back appear smaller. Pie slices facing the viewer appear larger. There is almost no scenario where 3D improves comprehension.

**Rule:** Never use `px.bar_3d`, `ax.bar3d`, 3D pie, or 3D ribbon charts for standard business data. The only legitimate use of 3D is actual 3D data (molecular structures, topography, point clouds).

### Dual Y-Axes

**Problem:** Two y-axes with different scales make it trivial to imply false correlation. The relationship between the two series depends entirely on how the axes are scaled.

**Better alternatives:**
1. Two separate charts stacked vertically with aligned x-axes.
2. Normalize both series to a common scale (index to 100 at start).
3. Use small multiples.

```python
# INSTEAD OF dual y-axes, use subplots:
from plotly.subplots import make_subplots

fig = make_subplots(rows=2, cols=1, shared_xaxes=True, vertical_spacing=0.08)
fig.add_trace(go.Scatter(x=dates, y=revenue, name="Revenue", line_color="#4C78A8"), row=1, col=1)
fig.add_trace(go.Scatter(x=dates, y=users, name="Users", line_color="#F58518"), row=2, col=1)
fig.update_layout(height=500, title="Revenue and Users (separate scales)")
fig.show()
```

### Rainbow Color Scales

**Problem:** Rainbow (jet, HSV) scales are not perceptually uniform. Yellow and cyan bands appear brighter than blue and red, creating false emphasis. They are also inaccessible to colorblind users.

**Fix:** Use viridis, plasma, inferno, magma, or cividis. For diverging data, use RdBu or PuOr.

```python
# BAD
fig = px.imshow(data, color_continuous_scale="rainbow")   # DO NOT USE
fig = px.imshow(data, color_continuous_scale="jet")        # DO NOT USE

# GOOD
fig = px.imshow(data, color_continuous_scale="Viridis")
fig = px.imshow(data, color_continuous_scale="Cividis")    # colorblind-safe
```

### Too Many Categories

**Problem:** More than 7-10 colors in a single chart overwhelm the viewer and make the legend unreadable.

**Fixes:**
1. Group small categories into "Other."
2. Use small multiples instead of color.
3. Use interactive filtering.
4. Highlight one category, gray out the rest.

```python
# Highlight one, gray out the rest
colors = ["#4C78A8" if region == "East" else "#D0D0D0" for region in df["region"]]
fig = px.bar(df, x="region", y="revenue", color="region",
             color_discrete_map={r: c for r, c in zip(df["region"], colors)})
fig.update_layout(showlegend=False)
fig.show()
```

### Chart Junk

**Problem:** Unnecessary decoration that does not encode data -- gradients, shadows, background images, decorative icons inside charts, excessive gridlines.

**Edward Tufte's data-ink ratio:** Maximize the proportion of ink that represents data. Remove anything that does not contribute to understanding.

**Checklist for removing chart junk:**
- Remove background color or gradient (use white or very light gray).
- Remove or lighten gridlines (use `#EBEBEB`, not black).
- Remove top and right axis spines.
- Remove tick marks (use labels only).
- Remove 3D effects, shadows, bevels.
- Remove decorative borders around the chart.
- Verify every visual element encodes information.

### Missing Baselines

**Problem:** Not showing zero on bar charts, or using area charts that do not start at zero, misrepresents the magnitude of values.

**Rule:** Bar charts and area charts must include zero in their range. Line charts and scatter plots may exclude zero when the focus is on variation, not magnitude.

### Inconsistent Scales Across Facets

**Problem:** When using small multiples, each panel has its own auto-scaled y-axis. This makes visual comparison between panels impossible.

**Fix:** Lock axes across facets using shared scales.

```python
# Plotly: ensure consistent scales
fig = px.line(df, x="date", y="revenue", facet_col="region", facet_col_wrap=2)
fig.update_yaxes(matches="y")  # all panels share the same y-axis range
fig.show()

# Matplotlib: use sharey=True
fig, axes = plt.subplots(1, 3, figsize=(14, 4), sharey=True)
```

### Area Chart Misuse

**Problem:** Stacked area charts with a non-zero baseline. Or using filled area for a single series when a line would suffice (the filled area implies volume when you may only want to show a trend).

**Rules:**
1. Stacked area charts must start at zero.
2. Use filled area only when cumulative magnitude matters.
3. For single series trends, prefer a line chart.
4. Order the stacks so the most variable series is on top (least variable at bottom for a stable baseline).

---

## Quick Reference: Hex Color Cheat Sheet

### Grays for Chart Elements

```
Background (light):  #FAFAFA or #F9FAFB
Grid lines:          #EBEBEB or #E5E7EB
Axis lines:          #D0D0D0 or #D1D5DB
Secondary text:      #6B7280
Primary text:        #333333 or #111827
Dark background:     #121212 or #1E1E1E
Dark grid:           #333333
Dark text:           #E0E0E0
```

### Data Colors (10-color safe default)

```
1. #4C78A8  (steel blue)
2. #F58518  (orange)
3. #E45756  (coral red)
4. #72B7B2  (teal)
5. #54A24B  (green)
6. #EECA3B  (gold)
7. #B279A2  (mauve)
8. #FF9DA6  (salmon pink)
9. #9D755D  (brown)
10.#BAB0AC  (warm gray)
```

### Semantic Colors

```
Positive / success:  #059669 (green-600) or #54A24B
Negative / error:    #DC2626 (red-600)   or #E45756
Warning:             #D97706 (amber-600) or #F58518
Info:                #2563EB (blue-600)   or #4C78A8
Neutral:             #6B7280 (gray-500)   or #BAB0AC
```

---

## Appendix: Decision Matrix

Use this table to quickly select a chart type based on your analytical goal and data shape.

| Goal | 1 Variable | 2 Variables | 3+ Variables | Time Axis |
|---|---|---|---|---|
| **Compare** | Bar, lollipop | Grouped bar, dot plot | Small multiples | Line, slope |
| **Distribute** | Histogram, KDE, box | Scatter, 2-D hist | Ridgeline, violin | Calendar heatmap |
| **Compose** | Waffle, pie (if < 4 cat) | Stacked bar, treemap | Sunburst | Stacked area |
| **Trend** | Sparkline | Multi-line | Small multiples | Line, area |
| **Relate** | -- | Scatter | Bubble, heatmap | Connected scatter |
| **Flow** | -- | Sankey | Sankey (multi-stage) | Waterfall |
| **Rank** | Horizontal bar | Slope chart | Bump chart | Slope over time |
| **Spatial** | Choropleth | Bubble map | Hex-bin | Animated map |

---

*Data Analysis Suite -- Visualization Patterns Reference v1.0*
*Generated for Claude Code plugin ecosystem*
