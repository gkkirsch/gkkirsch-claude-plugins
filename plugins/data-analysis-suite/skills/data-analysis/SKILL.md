---
name: data-analysis-suite
description: >
  Data Analysis & Visualization Suite — AI-powered toolkit for exploratory data analysis,
  chart generation, SQL optimization, and data pipeline architecture. Performs statistical
  analysis on any dataset, generates publication-quality visualizations with Plotly/Matplotlib/D3,
  optimizes SQL queries and database schemas, and architects robust ETL/ELT pipelines.
  Triggers: "analyze data", "data analysis", "EDA", "exploratory data analysis", "statistics",
  "statistical analysis", "correlation", "distribution", "outlier", "chart", "graph", "plot",
  "visualization", "visualize", "plotly", "matplotlib", "d3", "dashboard", "sql", "query",
  "optimize query", "query performance", "database schema", "indexing", "migration",
  "etl", "elt", "data pipeline", "data validation", "data quality", "data cleaning",
  "time series", "regression", "hypothesis test", "pivot table", "aggregate", "group by".
  Dispatches the appropriate specialist agent: data-explorer, chart-generator, sql-analyst,
  or data-pipeline.
  NOT for: Machine learning model training/tuning, deep learning, NLP/LLM fine-tuning,
  big data cluster management (Spark/Hadoop administration), real-time streaming infrastructure
  (Kafka/Flink ops), or BI tool administration (Tableau/PowerBI server config).
version: 1.0.0
argument-hint: "<explore|chart|sql|pipeline> [target]"
user-invocable: true
allowed-tools: Read, Grep, Glob, Bash
model: sonnet
---

# Data Analysis & Visualization Suite

Production-grade data analysis agents for Claude Code. Four specialist agents that handle exploratory data analysis, chart generation, SQL optimization, and data pipeline architecture — the analytical work that turns raw data into actionable insights.

## Available Agents

### Data Explorer (`data-explorer`)
Performs comprehensive exploratory data analysis on any dataset. Statistical summaries, distribution analysis, correlation matrices, outlier detection, missing data patterns, and automated insight generation.

**Invoke**: Dispatch via Task tool with `subagent_type: "data-explorer"`.

**Example prompts**:
- "Analyze this CSV and tell me what's interesting — correlations, outliers, patterns"
- "Run a full EDA on my sales data including time series decomposition"
- "Check this dataset for data quality issues — missing values, duplicates, inconsistencies"
- "Perform hypothesis testing to determine if treatment A outperforms treatment B"

### Chart Generator (`chart-generator`)
Creates publication-quality visualizations using Plotly, Matplotlib, or D3.js. Automatic chart type selection based on data characteristics, responsive design, accessibility compliance, and export-ready output.

**Invoke**: Dispatch via Task tool with `subagent_type: "chart-generator"`.

**Example prompts**:
- "Create an interactive dashboard showing revenue trends with drill-down by region"
- "Generate a correlation heatmap for all numeric columns in this dataset"
- "Build a D3.js visualization showing the network graph of user interactions"
- "Create a set of charts for my quarterly business review presentation"

### SQL Analyst (`sql-analyst`)
Optimizes SQL queries, designs efficient schemas, plans migrations, and tunes database performance. Supports PostgreSQL, MySQL, SQLite, and SQL Server with dialect-specific optimizations.

**Invoke**: Dispatch via Task tool with `subagent_type: "sql-analyst"`.

**Example prompts**:
- "This query takes 30 seconds — optimize it and explain what indexes I need"
- "Design a normalized schema for an e-commerce platform with orders, products, and inventory"
- "Plan a zero-downtime migration from MySQL to PostgreSQL"
- "Audit my database for missing indexes and inefficient query patterns"

### Data Pipeline (`data-pipeline`)
Architects robust ETL/ELT pipelines with data validation, error handling, and monitoring. Supports batch and streaming patterns with tools like Airflow, dbt, Prefect, and custom Node.js/Python pipelines.

**Invoke**: Dispatch via Task tool with `subagent_type: "data-pipeline"`.

**Example prompts**:
- "Build an ETL pipeline that syncs data from our API to PostgreSQL every hour"
- "Set up dbt models for our analytics warehouse with tests and documentation"
- "Create a data validation layer that catches schema changes and data quality issues"
- "Design a pipeline architecture for processing 10M events/day with exactly-once semantics"

## Quick Start: /analyze-data

Use the `/analyze-data` command for guided data analysis workflows:

```
/analyze-data                    # Auto-detect dataset and analyze
/analyze-data explore sales.csv  # Run EDA on specific file
/analyze-data chart revenue      # Generate revenue visualizations
/analyze-data sql optimize       # Optimize SQL queries in project
/analyze-data pipeline design    # Design data pipeline architecture
```

The `/analyze-data` command chains agents as needed: detect → explore → visualize → optimize.

## Agent Selection Guide

| Need | Agent | Command |
|------|-------|---------|
| Understand a dataset | data-explorer | "Analyze this data" |
| Find patterns/outliers | data-explorer | "Find outliers in sales data" |
| Statistical testing | data-explorer | "Test if conversion rates differ" |
| Create charts | chart-generator | "Visualize revenue trends" |
| Build dashboards | chart-generator | "Create an interactive dashboard" |
| Chart for presentation | chart-generator | "Make presentation-ready charts" |
| Optimize slow queries | sql-analyst | "This query is slow" |
| Design database schema | sql-analyst | "Design schema for [domain]" |
| Plan DB migration | sql-analyst | "Migrate from X to Y" |
| Index recommendations | sql-analyst | "What indexes do I need?" |
| Build ETL pipeline | data-pipeline | "Sync data from API to DB" |
| Data validation | data-pipeline | "Validate incoming data" |
| Pipeline monitoring | data-pipeline | "Add pipeline observability" |
| dbt project setup | data-pipeline | "Set up dbt models" |

## Reference Materials

This skill includes comprehensive reference documents in `references/`:

- **statistical-methods.md** — Hypothesis testing procedures, regression analysis, time series decomposition, probability distributions, effect size calculations, and sampling methods
- **visualization-patterns.md** — Chart type selection flowcharts, color palette systems, responsive design patterns, accessibility guidelines, and platform-specific templates for Plotly/Matplotlib/D3
- **sql-optimization.md** — Indexing strategies, query plan analysis, join optimization, partitioning schemes, CTE patterns, window functions, and database-specific performance tuning

Agents automatically consult these references when working. You can also read them directly for quick answers.

## How It Works

1. You describe what you need (e.g., "analyze my sales data")
2. The SKILL.md routes to the appropriate agent
3. The agent reads your data/queries, detects patterns, and performs the analysis
4. Results are written as scripts, reports, or optimized queries directly to your project
5. The agent provides interpretation, recommendations, and next steps

All analysis follows best practices:
- Statistical rigor: proper test selection, effect sizes, confidence intervals, multiple comparison corrections
- Visualization: perceptually uniform color scales, accessible design, appropriate chart types
- SQL: explain-plan driven optimization, proper indexing, query plan analysis
- Pipelines: idempotent operations, schema validation, error recovery, monitoring
