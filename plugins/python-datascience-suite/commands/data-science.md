# /data-science

Build data pipelines, train ML models, create visualizations, and deploy ML systems.

## Usage

```
/data-science [subcommand] [options]
```

## Subcommands

### `pipeline` — Build a Data Pipeline

Design and implement a data processing pipeline.

```
/data-science pipeline
```

**Process:**
1. Analyze data sources and requirements
2. Select tools (Pandas, Polars, DuckDB) based on data size
3. Build extraction, validation, cleaning, and transformation steps
4. Implement data quality checks
5. Write output to destination format
6. Add logging, error handling, and monitoring

**Options:**
- `--tool pandas|polars|duckdb` — Processing engine (auto-selects based on data size)
- `--source csv|parquet|database|api` — Data source type
- `--output parquet|csv|database` — Output destination

### `train` — Train a Machine Learning Model

Build and train an ML model with proper evaluation.

```
/data-science train
```

**Process:**
1. Define problem type (classification, regression, clustering)
2. Analyze data and select features
3. Build preprocessing pipeline with sklearn
4. Train baseline model
5. Train advanced models (XGBoost, LightGBM)
6. Evaluate with cross-validation and proper metrics
7. Tune hyperparameters with Optuna
8. Generate model card and save artifacts

**Options:**
- `--task classification|regression|clustering|anomaly` — ML task type
- `--model logistic|rf|xgboost|lightgbm|auto` — Model algorithm
- `--tune` — Enable hyperparameter tuning with Optuna
- `--explain` — Generate SHAP explanations

### `visualize` — Create Data Visualizations

Create insightful visualizations and dashboards.

```
/data-science visualize
```

**Capabilities:**
- Distribution plots (histogram, KDE, box, violin)
- Correlation heatmaps and scatter matrices
- Time series plots with trends and anomalies
- Interactive dashboards with Plotly
- EDA reports with comprehensive profiling
- Publication-quality figures

**Options:**
- `--tool matplotlib|seaborn|plotly` — Visualization library
- `--interactive` — Create interactive HTML charts
- `--dashboard` — Build a multi-panel dashboard

### `eda` — Exploratory Data Analysis

Run a comprehensive EDA on a dataset.

```
/data-science eda
```

**Process:**
1. Load and profile the dataset
2. Analyze missing values and data types
3. Plot distributions of all features
4. Compute and visualize correlations
5. Detect outliers and anomalies
6. Analyze target variable (if supervised)
7. Generate summary report with key findings

### `features` — Feature Engineering

Engineer features for machine learning.

```
/data-science features
```

**Capabilities:**
- Numeric: interactions, ratios, log transforms, binning
- Categorical: one-hot, target encoding, frequency encoding
- Temporal: lag features, rolling windows, date extraction
- Text: word count, character stats, pattern extraction
- Scaling: standard, min-max, robust scaling
- Selection: mutual information, permutation importance

### `deploy` — Deploy an ML Model

Package and deploy a trained model for serving.

```
/data-science deploy
```

**Capabilities:**
- FastAPI model server with health checks and batch prediction
- Docker containerization with multi-stage builds
- MLflow model registry integration
- Model monitoring with drift detection
- CI/CD pipeline for model validation and deployment
- A/B testing framework for model comparison

**Options:**
- `--framework fastapi|bentoml` — Serving framework
- `--registry mlflow` — Model registry
- `--monitor` — Add drift detection monitoring

### `notebook` — Jupyter Notebook Operations

Manage and execute Jupyter notebooks.

```
/data-science notebook
```

**Capabilities:**
- Parameterized execution with Papermill
- Batch execution across dates/configurations
- Convert to HTML/PDF reports
- Set up nbdev for notebook-driven development
- Configure version control with Jupytext

### `optimize` — Optimize Data Processing

Optimize data processing for performance and memory.

```
/data-science optimize
```

**Capabilities:**
- Memory profiling and dtype optimization
- Chunked processing for large files
- Polars conversion for performance
- DuckDB integration for SQL analytics
- Caching strategies for repeated computations

---

## Examples

```
# Build a data cleaning pipeline
/data-science pipeline --source csv --output parquet

# Train a classifier with tuning
/data-science train --task classification --model xgboost --tune

# Create an interactive dashboard
/data-science visualize --tool plotly --dashboard

# Run exploratory data analysis
/data-science eda

# Engineer features for ML
/data-science features

# Deploy model with FastAPI
/data-science deploy --framework fastapi --registry mlflow

# Execute notebook for multiple dates
/data-science notebook
```

---

## Agents

This command uses the following specialist agents:

| Agent | Expertise |
|-------|-----------|
| `data-pipeline-engineer` | Pandas, Polars, DuckDB, ETL, data cleaning, feature engineering |
| `ml-model-builder` | Scikit-learn, XGBoost, LightGBM, model selection, hyperparameter tuning |
| `visualization-expert` | Matplotlib, Seaborn, Plotly, dashboards, EDA visualization |
| `mlops-engineer` | MLflow, model serving, experiment tracking, deployment, monitoring |

---

## References

The following reference materials are available:

- `pandas-cookbook` — Advanced Pandas patterns, performance optimization, and recipes
- `ml-algorithms-guide` — Algorithm selection guide, when to use what, tuning tips
- `jupyter-workflows` — Jupyter best practices, nbdev, Papermill, notebook management

---

## Tips

1. **Start with EDA.** Always explore your data before building models.
2. **Use Polars for large data.** Switch from Pandas when datasets exceed 1 GB.
3. **Build pipelines, not scripts.** Use sklearn Pipeline to prevent data leakage.
4. **Track experiments.** Use MLflow to log params, metrics, and models.
5. **Start simple.** Baseline first, then increase complexity only if needed.
6. **Validate properly.** Use stratified cross-validation, never just a single split.
7. **Check for leakage.** If results seem too good, investigate feature leakage.
8. **Document with notebooks.** Combine code, visualizations, and explanations.
9. **Test your data.** Add data quality checks at every pipeline stage.
10. **Monitor in production.** Track data drift and model performance over time.
