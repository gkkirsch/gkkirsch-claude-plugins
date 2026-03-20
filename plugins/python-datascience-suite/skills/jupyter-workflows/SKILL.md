---
name: jupyter-workflows
description: >
  Jupyter notebook patterns for reproducible data science workflows.
  Use when setting up notebooks, creating reproducible analyses,
  building data pipelines in notebooks, or structuring notebook projects.
  Triggers: "jupyter", "notebook", "ipynb", "colab", "notebook structure",
  "reproducible analysis", "notebook template".
  NOT for: pandas operations (see pandas-patterns), visualization, or deployment.
version: 1.0.0
allowed-tools: Read, Grep, Glob, Bash
---

# Jupyter Workflow Patterns

## Notebook Structure Template

```python
# Cell 1: Setup and Configuration
"""
# Analysis: [Title]
**Author**: [Name]
**Date**: [Date]
**Objective**: [One sentence describing what this analysis answers]

## Data Sources
- source_a.csv: Description (as of YYYY-MM-DD)
- database: schema.table
"""

# Cell 2: Imports and Config
import pandas as pd
import numpy as np
import matplotlib.pyplot as plt
import seaborn as sns
from pathlib import Path
from datetime import datetime
import warnings
warnings.filterwarnings('ignore', category=FutureWarning)

# Configuration
pd.set_option('display.max_columns', 50)
pd.set_option('display.max_rows', 100)
pd.set_option('display.float_format', lambda x: f'{x:,.2f}')
plt.style.use('seaborn-v0_8-darkgrid')
sns.set_palette('husl')

DATA_DIR = Path('../data')
OUTPUT_DIR = Path('../output')
OUTPUT_DIR.mkdir(exist_ok=True)

RANDOM_SEED = 42
np.random.seed(RANDOM_SEED)

# Cell 3: Data Loading
print(f"Loading data at {datetime.now().strftime('%Y-%m-%d %H:%M')}")

df = pd.read_csv(DATA_DIR / 'source.csv', parse_dates=['date'])
print(f"Loaded {len(df):,} rows, {len(df.columns)} columns")
print(f"Date range: {df['date'].min()} to {df['date'].max()}")
df.info()

# Cell 4: Data Quality Check
print("=== Data Quality Report ===")
print(f"\nShape: {df.shape}")
print(f"\nNull counts:\n{df.isnull().sum()[df.isnull().sum() > 0]}")
print(f"\nDuplicate rows: {df.duplicated().sum()}")

for col in df.select_dtypes(include='number').columns:
    print(f"\n{col}: min={df[col].min():.2f}, max={df[col].max():.2f}, "
          f"mean={df[col].mean():.2f}, nulls={df[col].isna().sum()}")
```

## Parameterized Notebooks with Papermill

```python
# parameters cell (tagged with "parameters" in Jupyter)
# These values are overridden when run via papermill
start_date = '2026-01-01'
end_date = '2026-03-01'
region = 'US'
min_revenue = 100

# Run from CLI:
# papermill template.ipynb output_us.ipynb -p region US -p min_revenue 100
# papermill template.ipynb output_eu.ipynb -p region EU -p min_revenue 50
```

```python
# Batch execution script
import papermill as pm
from pathlib import Path

regions = ['US', 'EU', 'APAC']
output_dir = Path('reports')
output_dir.mkdir(exist_ok=True)

for region in regions:
    print(f"Running report for {region}...")
    pm.execute_notebook(
        'template.ipynb',
        str(output_dir / f'report_{region.lower()}.ipynb'),
        parameters={
            'region': region,
            'start_date': '2026-01-01',
            'end_date': '2026-03-31',
        },
        kernel_name='python3',
    )
    print(f"  Done: {output_dir / f'report_{region.lower()}.ipynb'}")
```

## Magic Commands Reference

```python
# Timing
%time result = expensive_function()           # Wall time for single execution
%timeit result = expensive_function()         # Average over many runs
%%time                                         # Time entire cell

# Profiling
%load_ext line_profiler
%lprun -f my_function my_function(data)       # Line-by-line profiling

# Memory
%load_ext memory_profiler
%memit df = pd.read_csv('large.csv')          # Peak memory usage

# Shell commands
!pip install package-name
!ls -la ../data/

# Environment
%env API_KEY=abc123                            # Set env var
%who DataFrame                                 # List variables of type

# Autoreload (for module development)
%load_ext autoreload
%autoreload 2                                  # Reload all modules before execution

# Matplotlib
%matplotlib inline                             # Static plots in notebook
%matplotlib widget                             # Interactive plots (ipympl)
```

## Notebook Testing

```python
# nbval: test notebooks in CI
# pip install nbval
# pytest --nbval my_notebook.ipynb

# Cell-level assertions (run as part of notebook)
def assert_data_quality(df: pd.DataFrame, name: str = "data"):
    """Inline data assertions for notebook cells."""
    issues = []

    if len(df) == 0:
        issues.append(f"{name}: DataFrame is empty")
    if df.duplicated().sum() > 0:
        issues.append(f"{name}: {df.duplicated().sum()} duplicate rows")

    null_cols = df.columns[df.isnull().any()].tolist()
    if null_cols:
        issues.append(f"{name}: nulls in {null_cols}")

    if issues:
        for issue in issues:
            print(f"WARNING: {issue}")
    else:
        print(f"OK: {name} ({len(df):,} rows, {len(df.columns)} cols)")

    return len(issues) == 0

# Usage in notebook cell:
assert assert_data_quality(df_cleaned, "cleaned_orders"), "Data quality check failed!"
```

## Export and Sharing

```python
# Export to various formats
# jupyter nbconvert --to html notebook.ipynb
# jupyter nbconvert --to pdf notebook.ipynb
# jupyter nbconvert --to script notebook.ipynb  (creates .py file)

# Programmatic HTML export with no code cells
import subprocess
subprocess.run([
    'jupyter', 'nbconvert',
    '--to', 'html',
    '--no-input',                    # Hide code cells
    '--no-prompt',                   # Hide In[]/Out[] prompts
    '--output', 'report.html',
    'analysis.ipynb'
])

# Save figures for reports
fig, ax = plt.subplots(figsize=(10, 6))
ax.plot(df['date'], df['revenue'])
ax.set_title('Monthly Revenue')
fig.savefig(OUTPUT_DIR / 'revenue_chart.png', dpi=150, bbox_inches='tight')
plt.show()
```

## Gotchas

1. **Out-of-order cell execution** — the #1 source of notebook bugs. Cells executed out of order create invisible state dependencies. Always "Restart & Run All" before sharing. Use `nbstripout` to clear outputs from git commits.

2. **Global state pollution** — variables from deleted cells persist in the kernel. `del df_temp` after intermediate DataFrames, or restart the kernel periodically. Variable names from earlier experiments silently shadow current work.

3. **Notebook too large for git** — notebooks with large outputs (images, DataFrames) bloat the repository. Use `nbstripout` as a git filter to auto-clear outputs on commit: `nbstripout --install`.

4. **Relative paths break when notebook moves** — `pd.read_csv('data/file.csv')` depends on the notebook's working directory, which varies by launch method. Use `Path(__file__).parent / 'data'` or configure a project root constant.

5. **pip install in notebook installs to wrong environment** — `!pip install package` may install to a different Python than the kernel is using. Use `import sys; !{sys.executable} -m pip install package` to ensure correct environment.

6. **Matplotlib figures accumulate memory** — each `plt.figure()` stays in memory until explicitly closed. In loops generating many figures, call `plt.close('all')` periodically, or use `plt.close(fig)` after saving.
