---
name: analyze-data
description: >
  Quick data analysis command — detects datasets in your project, runs exploratory analysis,
  generates visualizations, optimizes queries, and builds data pipelines. Works with CSV, JSON,
  Parquet, SQL databases, and API data sources.
  Triggers: "/analyze-data", "analyze my data", "explore this dataset", "visualize my data".
user-invocable: true
argument-hint: "<explore|chart|sql|pipeline> [target] [--format html|pdf|notebook]"
allowed-tools: Read, Write, Edit, Bash, Glob, Grep
model: sonnet
---

# /analyze-data Command

One-command data analysis workflow. Detects datasets, runs statistical analysis, generates visualizations, and produces actionable insights.

## Usage

```
/analyze-data                           # Auto-detect datasets and analyze
/analyze-data explore                   # Run full EDA on detected datasets
/analyze-data explore data/sales.csv    # Analyze specific file
/analyze-data chart revenue             # Generate revenue visualizations
/analyze-data chart data/*.csv          # Visualize all CSVs
/analyze-data sql                       # Find and optimize SQL in project
/analyze-data sql queries/report.sql    # Optimize specific query file
/analyze-data pipeline                  # Design pipeline for detected data flows
/analyze-data pipeline --streaming      # Design streaming pipeline
```

## Options

| Flag | Description |
|------|-------------|
| `--format html` | Output as interactive HTML report |
| `--format notebook` | Output as Jupyter notebook |
| `--format pdf` | Output as static PDF report |
| `--no-viz` | Skip visualization generation |
| `--deep` | Run deep analysis (more statistical tests, larger sample) |
| `--streaming` | Design streaming (not batch) pipeline |

## Procedure

### Step 1: Detect Data Assets

Scan the project for data files and database connections:

1. **Glob** for data files: `**/*.csv`, `**/*.json`, `**/*.parquet`, `**/*.xlsx`, `**/*.tsv`, `**/*.jsonl`
2. **Glob** for SQL files: `**/*.sql`, `**/migrations/**`, `**/queries/**`
3. **Glob** for database configs: `**/schema.prisma`, `**/drizzle.config.*`, `**/knexfile.*`, `**/alembic.ini`, `**/ormconfig.*`
4. **Grep** for connection strings: `DATABASE_URL`, `POSTGRES`, `MYSQL`, `MONGO`, `REDIS` in `.env*` files
5. **Grep** for data loading: `pd.read_csv`, `pd.read_json`, `fs.readFile`, `csv-parse`, `papaparse` in source files
6. **Glob** for notebook files: `**/*.ipynb`
7. **Read** `package.json` or `requirements.txt` for data libraries (pandas, numpy, plotly, d3, chart.js)

Report findings:

```
Data Assets Detected:
- Files: 3 CSVs (data/sales.csv, data/customers.csv, data/products.csv)
- Database: PostgreSQL via Drizzle ORM (src/db/schema.ts)
- SQL queries: 12 files in queries/
- Libraries: pandas, plotly, numpy (requirements.txt)
- Notebooks: 2 Jupyter notebooks in notebooks/
```

### Step 2: Route to Specialist

Based on the subcommand or auto-detection:

| Subcommand | Agent | When |
|------------|-------|------|
| `explore` | data-explorer | Default when data files found |
| `chart` | chart-generator | When visualization requested |
| `sql` | sql-analyst | When SQL files or DB connections found |
| `pipeline` | data-pipeline | When multiple data sources or ETL needed |
| (auto) | data-explorer | Start with EDA, then recommend next steps |

### Step 3: Dispatch Analysis (explore mode)

Dispatch the `data-explorer` agent:

```
Task tool:
  subagent_type: "data-explorer"
  mode: "bypassPermissions"
  prompt: |
    Perform exploratory data analysis on the following dataset(s):
    Files: [detected files]
    Database: [connection info if applicable]

    Generate:
    1. Statistical summary (shape, types, descriptive stats)
    2. Missing data analysis
    3. Distribution analysis for all numeric columns
    4. Correlation matrix
    5. Outlier detection
    6. Key insights and recommendations

    Write analysis script to: analysis/eda.py (or .ts)
    Write report to: analysis/report.md
```

### Step 4: Dispatch Visualization (chart mode)

Dispatch the `chart-generator` agent:

```
Task tool:
  subagent_type: "chart-generator"
  mode: "bypassPermissions"
  prompt: |
    Create visualizations for the analyzed data:
    Dataset: [detected or specified files]
    Analysis results: [from Step 3 if available]
    Target: [specified chart type or auto-select]

    Generate:
    1. Overview dashboard with key metrics
    2. Distribution plots for numeric columns
    3. Correlation heatmap
    4. Time series trends (if temporal data exists)
    5. Category comparisons (if categorical data exists)

    Write to: analysis/charts/ directory
    Format: [--format flag or default to HTML]
```

### Step 5: Dispatch SQL Optimization (sql mode)

Dispatch the `sql-analyst` agent:

```
Task tool:
  subagent_type: "sql-analyst"
  mode: "bypassPermissions"
  prompt: |
    Analyze and optimize SQL in this project:
    SQL files: [detected .sql files]
    ORM schema: [detected schema files]
    Database: [detected DB type]

    Perform:
    1. Query analysis — identify slow patterns
    2. Index recommendations
    3. Schema review
    4. Rewrite queries for performance

    Write optimized queries to: analysis/optimized-queries/
    Write report to: analysis/sql-report.md
```

### Step 6: Dispatch Pipeline Design (pipeline mode)

Dispatch the `data-pipeline` agent:

```
Task tool:
  subagent_type: "data-pipeline"
  mode: "bypassPermissions"
  prompt: |
    Design a data pipeline for this project:
    Data sources: [detected sources]
    Data targets: [detected destinations]
    Stack: [detected language/framework]
    Mode: [batch or streaming]

    Generate:
    1. Pipeline architecture diagram (mermaid)
    2. Implementation code
    3. Data validation schemas
    4. Error handling and retry logic
    5. Monitoring and alerting setup

    Write to: pipeline/ directory
```

### Step 7: Generate Summary Report

After specialist agents complete, generate a summary:

```markdown
# Data Analysis Report

## Dataset Overview
- [Shape, size, column details]

## Key Findings
1. [Most significant insight]
2. [Second insight]
3. [Third insight]

## Visualizations Generated
- [List of charts with descriptions]

## Recommendations
- [Actionable next steps]

## Files Created
- analysis/eda.py — EDA script (re-runnable)
- analysis/report.md — Detailed findings
- analysis/charts/ — Generated visualizations
- analysis/optimized-queries/ — Optimized SQL (if applicable)
```

### Step 8: Suggest Next Steps

Based on findings, recommend follow-up actions:

- **If patterns found**: "Consider building a predictive model for [target]"
- **If data quality issues**: "Run data-pipeline agent to set up validation"
- **If slow queries**: "Run sql mode to optimize the top 5 slowest queries"
- **If time series data**: "Consider forecasting with ARIMA or Prophet"
- **If high-dimensional data**: "Consider dimensionality reduction (PCA/t-SNE)"

## Supported Data Formats

| Format | Extension | Library Used |
|--------|-----------|--------------|
| CSV | .csv, .tsv | pandas / csv-parse |
| JSON | .json, .jsonl | pandas / native |
| Parquet | .parquet | pyarrow / parquet-wasm |
| Excel | .xlsx, .xls | openpyxl / SheetJS |
| SQL | .sql | sqlparse / knex |
| SQLite | .db, .sqlite | sqlite3 / better-sqlite3 |
| HDF5 | .h5, .hdf5 | h5py |
| Feather | .feather | pyarrow |

## Error Recovery

| Error | Cause | Fix |
|-------|-------|-----|
| File too large | Dataset exceeds memory | Use chunked reading or sample |
| Encoding error | Non-UTF-8 file | Detect encoding with chardet |
| Parse error | Malformed CSV/JSON | Try different delimiters or repair |
| Missing library | pandas/plotly not installed | Auto-install with pip/npm |
| DB connection fail | Invalid credentials | Check .env and connection string |
| Permission denied | File access restricted | Check file permissions |

## Notes

- Always work with copies of data — never modify source files
- For large datasets (>1GB), automatically use sampling and chunked processing
- Generated scripts are re-runnable — they can be executed independently
- All visualizations include titles, labels, and legends by default
- SQL optimization always starts with EXPLAIN ANALYZE before recommending changes
