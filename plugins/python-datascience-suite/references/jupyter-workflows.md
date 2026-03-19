# Jupyter Workflows — Best Practices, nbdev, and Papermill

Quick reference for productive Jupyter notebook workflows, parameterized execution with Papermill, and notebook-driven development with nbdev.

---

## Jupyter Best Practices

### Notebook Structure

Every notebook should follow this structure:

```
1. Title and Description (Markdown)
   - What this notebook does
   - Author, date, version
   - Dependencies and data sources

2. Setup
   - Imports
   - Configuration
   - Data loading

3. Exploration / Analysis (numbered sections)
   - Each section has a clear purpose
   - Markdown cells explain the "why"
   - Code cells do the "what"

4. Results / Summary
   - Key findings
   - Next steps
   - Artifacts produced
```

### Coding Standards in Notebooks

```python
# Cell 1: Always start with imports and config
import pandas as pd
import numpy as np
import matplotlib.pyplot as plt
import seaborn as sns
from pathlib import Path

# Configuration
DATA_DIR = Path("data")
OUTPUT_DIR = Path("output")
OUTPUT_DIR.mkdir(exist_ok=True)
RANDOM_STATE = 42

# Display settings
pd.set_option("display.max_columns", None)
pd.set_option("display.max_rows", 100)
pd.set_option("display.float_format", "{:,.4f}".format)
plt.style.use("seaborn-v0_8-whitegrid")
%matplotlib inline
```

### Cell Best Practices

1. **One idea per cell**: Each cell should do one thing
2. **Name your outputs**: Don't rely on Jupyter's last-expression display
3. **Use Markdown liberally**: Explain decisions, assumptions, findings
4. **Keep cells short**: If a cell is > 30 lines, break it up
5. **Avoid side effects**: Don't modify global state in helper cells
6. **Use magic commands wisely**: `%%time`, `%load_ext`, `%autoreload`

### Useful Magic Commands

```python
# Timing
%time result = expensive_function()    # Wall time for one call
%timeit x = np.sum(arr)                # Average time over multiple runs
%%time                                 # Time entire cell

# Autoreload (for imported modules during development)
%load_ext autoreload
%autoreload 2  # Reload all modules before executing

# Memory profiling
%load_ext memory_profiler
%memit df = pd.read_csv("large.csv")

# Debug
%debug  # Post-mortem debugger after exception
%pdb on  # Auto-enter debugger on exception

# Environment info
%env  # Show all environment variables
%who  # Show defined variables
%whos  # Show defined variables with details

# Shell commands
!pip install package_name
!ls data/

# Store variables between notebooks
%store variable_name     # Save
%store -r variable_name  # Retrieve in another notebook
```

---

## Notebook Organization

### Directory Structure for Data Science Projects

```
project/
├── notebooks/
│   ├── 01_data_exploration.ipynb
│   ├── 02_feature_engineering.ipynb
│   ├── 03_model_training.ipynb
│   ├── 04_model_evaluation.ipynb
│   └── 05_results_and_reporting.ipynb
├── src/
│   ├── __init__.py
│   ├── data.py          # Data loading and cleaning
│   ├── features.py      # Feature engineering
│   ├── models.py        # Model definitions
│   ├── evaluate.py      # Evaluation functions
│   └── visualize.py     # Visualization functions
├── data/
│   ├── raw/             # Original data (never modified)
│   ├── processed/       # Cleaned/transformed data
│   └── features/        # Feature-engineered data
├── models/              # Saved models
├── reports/             # Generated reports and figures
├── tests/               # Unit tests for src/
├── config/              # Configuration files
├── requirements.txt
└── README.md
```

### Rule: Notebooks for Exploration, Modules for Production

```python
# In notebooks: explore, visualize, iterate
df = pd.read_csv("data/raw/sales.csv")
df.describe()
df.plot(kind="hist")

# In src/: production-ready, tested, reusable code
# src/data.py
def load_sales_data(filepath: str) -> pd.DataFrame:
    """Load and validate sales data."""
    df = pd.read_csv(filepath)
    assert "date" in df.columns, "Missing date column"
    df["date"] = pd.to_datetime(df["date"])
    return df

# In notebooks: import from src/
from src.data import load_sales_data
df = load_sales_data("data/raw/sales.csv")
```

---

## Papermill — Parameterized Notebooks

### What is Papermill?

Papermill lets you parameterize and execute notebooks programmatically. Use it to:
- Run the same analysis for different datasets, dates, or configurations
- Schedule notebook execution in pipelines
- Generate reports automatically

### Setup

```bash
pip install papermill
```

### Parameterizing a Notebook

In your notebook, create a cell tagged with `parameters`:

```python
# Cell tagged as "parameters" (tag it via Jupyter UI or metadata)
# These are default values; Papermill will override them
dataset = "train"
date = "2024-01-01"
n_estimators = 200
output_path = "output/results.csv"
```

### Executing with Papermill

```python
import papermill as pm

# Execute a parameterized notebook
pm.execute_notebook(
    "notebooks/03_model_training.ipynb",     # Input notebook
    "output/03_model_training_prod.ipynb",   # Output notebook (with results)
    parameters={
        "dataset": "production",
        "date": "2024-03-01",
        "n_estimators": 500,
        "output_path": "output/prod_results.csv",
    },
    kernel_name="python3",
)
```

### Batch Execution

```python
import papermill as pm
from datetime import datetime, timedelta

# Run the same notebook for multiple dates
dates = pd.date_range("2024-01-01", "2024-03-01", freq="W")

for date in dates:
    date_str = date.strftime("%Y-%m-%d")
    output_path = f"output/weekly_report_{date_str}.ipynb"

    pm.execute_notebook(
        "notebooks/weekly_report.ipynb",
        output_path,
        parameters={
            "report_date": date_str,
            "output_dir": f"output/reports/{date_str}/",
        },
    )
    print(f"Generated report for {date_str}")
```

### Papermill in Airflow

```python
from airflow import DAG
from airflow.operators.python import PythonOperator
from datetime import datetime
import papermill as pm


def run_notebook(**kwargs):
    execution_date = kwargs["ds"]
    pm.execute_notebook(
        "notebooks/daily_pipeline.ipynb",
        f"output/daily_pipeline_{execution_date}.ipynb",
        parameters={
            "date": execution_date,
            "env": "production",
        },
    )


with DAG("daily_notebook_pipeline", start_date=datetime(2024, 1, 1), schedule="@daily") as dag:
    run = PythonOperator(
        task_id="run_notebook",
        python_callable=run_notebook,
    )
```

### Collecting Results from Papermill

```python
import papermill as pm
import scrapbook as sb

# In the executed notebook, record results:
# sb.glue("accuracy", 0.95)
# sb.glue("feature_importance", importance_df)
# sb.glue("confusion_matrix", cm_fig)

# Read results from executed notebook
nb = sb.read_notebook("output/03_model_training_prod.ipynb")
accuracy = nb.scraps["accuracy"].data
importance = nb.scraps["feature_importance"].data
```

---

## nbdev — Notebook-Driven Development

### What is nbdev?

nbdev lets you develop Python libraries entirely in Jupyter notebooks. It extracts:
- **Code**: from notebook cells into Python modules
- **Tests**: from notebook cells into test files
- **Docs**: from notebook markdown into documentation

### Setup

```bash
pip install nbdev
nbdev_new  # Initialize a new nbdev project
```

### nbdev Directives

```python
#| default_exp core  # This notebook exports to src/core.py

#| export
def process_data(df):
    """Process raw data into features."""
    # This function will be exported to core.py
    return df.dropna()

#| hide
# This cell won't appear in docs
temp = "debugging helper"

#| exporti
def _internal_helper():
    """Internal function (exported but not in __all__)."""
    pass

# Cells without #| export are treated as tests/examples
# They run during nbdev_test but don't get exported
result = process_data(sample_df)
assert len(result) == expected_count
```

### nbdev Workflow

```bash
# Create new notebook
# Write code with #| export directives
# Run cells interactively to test

# Export code from notebooks to modules
nbdev_export

# Run all tests (cells without #| export)
nbdev_test

# Build documentation
nbdev_docs

# Prepare for release
nbdev_prepare  # export + test + docs + clean

# Clean notebook metadata (for git)
nbdev_clean
```

### nbdev Project Structure

```
my_library/
├── nbs/
│   ├── 00_core.ipynb       # Exports to my_library/core.py
│   ├── 01_data.ipynb        # Exports to my_library/data.py
│   ├── 02_models.ipynb      # Exports to my_library/models.py
│   ├── index.ipynb          # README and landing page
│   └── _quarto.yml          # Docs configuration
├── my_library/
│   ├── __init__.py          # Auto-generated
│   ├── core.py              # Auto-generated from 00_core.ipynb
│   ├── data.py              # Auto-generated from 01_data.ipynb
│   └── models.py            # Auto-generated from 02_models.ipynb
├── tests/                   # Auto-generated test files
├── setup.py
└── settings.ini
```

### Benefits of nbdev

1. **Exploratory + Production**: Write code interactively, export to modules
2. **Tests are examples**: Test cells serve as documentation and examples
3. **Docs from notebooks**: Rich documentation with outputs and visualizations
4. **Git-friendly**: `nbdev_clean` strips output for clean diffs
5. **CI/CD integration**: `nbdev_test` runs all notebook tests

---

## Notebook Version Control

### Git Best Practices for Notebooks

Notebooks are JSON files with embedded outputs, making git diffs noisy. Solutions:

```bash
# Option 1: nbstripout (strip output before commit)
pip install nbstripout
nbstripout --install  # Adds git filter

# Option 2: nbdev_clean
nbdev_clean  # Clean all notebooks

# Option 3: jupytext (pair notebooks with .py files)
pip install jupytext
# Pair .ipynb with .py:percent format
jupytext --set-formats ipynb,py:percent notebook.ipynb
```

### Jupytext — Notebooks as Scripts

```bash
# Convert notebook to Python script
jupytext --to py:percent notebook.ipynb

# Convert Python script to notebook
jupytext --to notebook script.py

# Sync paired files (edit either, sync both)
jupytext --sync notebook.ipynb
```

The `.py:percent` format looks like:

```python
# %% [markdown]
# # My Analysis
# This notebook analyzes sales data.

# %%
import pandas as pd
df = pd.read_csv("data.csv")

# %% [markdown]
# ## Data Exploration

# %%
df.describe()
```

Benefits:
- Git diffs are clean (it's just Python)
- Can edit in any IDE
- Syncs bidirectionally with .ipynb

---

## Notebook Extensions and Tools

### JupyterLab Extensions

| Extension | Purpose |
|-----------|---------|
| `jupyterlab-git` | Git integration in JupyterLab |
| `jupyterlab-toc` | Table of contents sidebar |
| `jupyterlab-code-formatter` | Auto-format cells (black, isort) |
| `jupyterlab-variableinspector` | Variable inspector panel |
| `jupyterlab-execute-time` | Show cell execution time |

### VS Code Jupyter

VS Code provides excellent Jupyter support:
- Interactive window (`# %%` cells in .py files)
- Variable explorer
- IntelliSense in cells
- Git integration without output noise
- Debugging support

```python
# In VS Code, use # %% to create cells in .py files
# %%
import pandas as pd
df = pd.read_csv("data.csv")

# %%
df.describe()

# %% [markdown]
# ## Analysis Section
```

---

## Testing Notebooks

### pytest with nbval

```bash
pip install nbval

# Test that notebooks execute without errors
pytest --nbval notebooks/

# Test only specific notebooks
pytest --nbval notebooks/01_data_exploration.ipynb

# Sanitize output for comparison
pytest --nbval --nbval-sanitize-with sanitize.cfg notebooks/
```

### testbook — Unit Testing Notebook Cells

```python
"""Test specific notebook cells."""
from testbook import testbook


@testbook("notebooks/03_model_training.ipynb", execute=True)
def test_model_accuracy(tb):
    """Test that model achieves minimum accuracy."""
    accuracy = tb.ref("accuracy")
    assert accuracy > 0.8, f"Accuracy {accuracy} below threshold"


@testbook("notebooks/02_feature_engineering.ipynb", execute=["setup", "features"])
def test_feature_count(tb):
    """Test that feature engineering produces expected number of features."""
    df = tb.ref("feature_df")
    assert len(df.columns) >= 20
```

---

## Notebook as Reports

### Generating PDF/HTML Reports

```bash
# HTML report
jupyter nbconvert --to html --no-input notebook.ipynb  # Hide code cells

# PDF report
jupyter nbconvert --to pdf notebook.ipynb

# Slides
jupyter nbconvert --to slides notebook.ipynb --post serve

# Execute and convert in one step
jupyter nbconvert --to html --execute notebook.ipynb
```

### Quarto for Polished Reports

```bash
# Install Quarto (https://quarto.org)
# Render notebook to various formats
quarto render notebook.ipynb --to html
quarto render notebook.ipynb --to pdf
quarto render notebook.ipynb --to docx

# Quarto YAML header in first cell:
# ---
# title: "Sales Analysis Q1 2024"
# author: "Data Team"
# format:
#   html:
#     toc: true
#     code-fold: true
# ---
```

---

## Performance Tips

### Large Data in Notebooks

```python
# 1. Use Parquet instead of CSV
df = pd.read_parquet("data.parquet")  # 10-100x faster than CSV

# 2. Read only needed columns
df = pd.read_parquet("data.parquet", columns=["id", "value", "date"])

# 3. Use DuckDB for SQL analytics (no loading into memory)
import duckdb
result = duckdb.sql("SELECT * FROM 'data.parquet' WHERE value > 100").df()

# 4. Profile memory usage
df.info(memory_usage="deep")
df.memory_usage(deep=True).sum() / 1024**2  # MB

# 5. Delete unused DataFrames
del large_intermediate_df
import gc; gc.collect()

# 6. Use dtypes to reduce memory
df = pd.read_csv("data.csv", dtype={"id": "int32", "cat": "category"})
```

### Notebook Execution Optimization

```python
# Cache expensive computations
from functools import lru_cache
import joblib

# Option 1: lru_cache for function results
@lru_cache(maxsize=32)
def expensive_computation(param):
    return compute(param)

# Option 2: joblib for disk caching
from joblib import Memory
memory = Memory("cache/", verbose=0)

@memory.cache
def load_and_process(filepath):
    df = pd.read_csv(filepath)
    return process(df)

# Option 3: pickle intermediate results
import pickle

# Save checkpoint
with open("checkpoint.pkl", "wb") as f:
    pickle.dump({"df": df, "model": model, "metrics": metrics}, f)

# Load checkpoint (skip expensive cells)
with open("checkpoint.pkl", "rb") as f:
    checkpoint = pickle.load(f)
    df = checkpoint["df"]
```

---

## Common Patterns

### Notebook Template

```python
# Cell 1: Header (Markdown)
"""
# [Analysis Title]
**Author:** [Name]
**Date:** [Date]
**Purpose:** [What this notebook does]
**Data:** [Where data comes from]
"""

# Cell 2: Setup
import pandas as pd
import numpy as np
import matplotlib.pyplot as plt
import seaborn as sns
from pathlib import Path
import warnings
warnings.filterwarnings("ignore")

%matplotlib inline
%load_ext autoreload
%autoreload 2

DATA_DIR = Path("data")
OUTPUT_DIR = Path("output")
RANDOM_STATE = 42

# Cell 3: Load Data
df = pd.read_parquet(DATA_DIR / "dataset.parquet")
print(f"Loaded {len(df):,} rows, {len(df.columns)} columns")
df.head()

# Cell 4+: Analysis sections...

# Final Cell: Summary and export
print("Key findings:")
print(f"  1. ...")
print(f"  2. ...")
# Save results
results_df.to_parquet(OUTPUT_DIR / "results.parquet")
```

### Interactive Widgets

```python
import ipywidgets as widgets
from IPython.display import display

# Dropdown for column selection
column_selector = widgets.Dropdown(
    options=df.columns.tolist(),
    description="Column:",
)

# Slider for parameters
threshold = widgets.FloatSlider(
    value=0.5, min=0, max=1, step=0.05,
    description="Threshold:",
)

# Interactive plot
@widgets.interact(column=df.columns.tolist(), bins=(10, 100, 10))
def plot_distribution(column, bins=50):
    plt.figure(figsize=(10, 5))
    df[column].hist(bins=bins)
    plt.title(f"Distribution of {column}")
    plt.show()
```
