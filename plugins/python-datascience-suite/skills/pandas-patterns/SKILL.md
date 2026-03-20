---
name: pandas-patterns
description: >
  Production pandas patterns for data manipulation, cleaning, and analysis.
  Use when working with DataFrames, data cleaning, merging datasets,
  aggregations, time series, or optimizing pandas performance.
  Triggers: "pandas", "dataframe", "data cleaning", "merge", "groupby",
  "pivot table", "time series pandas", "csv processing".
  NOT for: visualization (use matplotlib/plotly), ML model training, or SQL queries.
version: 1.0.0
allowed-tools: Read, Grep, Glob, Bash
---

# Pandas Patterns

## Data Loading with Type Safety

```python
import pandas as pd
import numpy as np
from pathlib import Path

# Type-aware CSV loading
def load_csv_typed(filepath: str | Path, date_columns: list[str] = None) -> pd.DataFrame:
    """Load CSV with explicit dtypes to prevent silent type coercion."""
    # Read first to inspect, then reload with proper types
    sample = pd.read_csv(filepath, nrows=5)

    dtypes = {}
    for col in sample.columns:
        if sample[col].dtype == 'object':
            # Check if it looks like a category (few unique values)
            dtypes[col] = 'category' if sample[col].nunique() < 10 else 'string'
        elif sample[col].dtype == 'int64':
            # Use nullable integer to handle missing values
            dtypes[col] = 'Int64'
        elif sample[col].dtype == 'float64':
            dtypes[col] = 'Float64'

    df = pd.read_csv(
        filepath,
        dtype=dtypes,
        parse_dates=date_columns or [],
        na_values=['', 'N/A', 'null', 'None', 'nan'],
    )

    return df


# Chunked reading for large files
def process_large_csv(filepath: str, chunk_size: int = 100_000):
    """Process CSV in chunks to avoid memory issues."""
    results = []
    for chunk in pd.read_csv(filepath, chunksize=chunk_size):
        # Process each chunk
        processed = chunk.groupby('category').agg({'amount': 'sum'})
        results.append(processed)

    return pd.concat(results).groupby(level=0).sum()
```

## Data Cleaning Pipeline

```python
def clean_dataframe(df: pd.DataFrame) -> pd.DataFrame:
    """Standard cleaning pipeline."""
    df = df.copy()

    # 1. Standardize column names
    df.columns = (
        df.columns
        .str.strip()
        .str.lower()
        .str.replace(r'[^a-z0-9]', '_', regex=True)
        .str.replace(r'_+', '_', regex=True)
        .str.strip('_')
    )

    # 2. Remove fully duplicate rows
    initial_len = len(df)
    df = df.drop_duplicates()
    dupes_removed = initial_len - len(df)
    if dupes_removed > 0:
        print(f"Removed {dupes_removed} duplicate rows")

    # 3. Handle missing values per column type
    for col in df.columns:
        null_pct = df[col].isna().mean()
        if null_pct > 0.5:
            print(f"Warning: {col} is {null_pct:.0%} null")

        if df[col].dtype in ['string', 'object', 'category']:
            df[col] = df[col].fillna('unknown')
        # Numeric columns: leave as NaN for explicit handling

    # 4. Strip whitespace from string columns
    str_cols = df.select_dtypes(include=['string', 'object']).columns
    for col in str_cols:
        df[col] = df[col].str.strip()

    return df


# Validation
def validate_dataframe(df: pd.DataFrame, rules: dict) -> list[str]:
    """Validate DataFrame against business rules."""
    errors = []

    for col, checks in rules.items():
        if col not in df.columns:
            errors.append(f"Missing required column: {col}")
            continue

        if 'not_null' in checks and df[col].isna().any():
            null_count = df[col].isna().sum()
            errors.append(f"{col}: {null_count} null values found")

        if 'unique' in checks and df[col].duplicated().any():
            dupe_count = df[col].duplicated().sum()
            errors.append(f"{col}: {dupe_count} duplicate values")

        if 'min' in checks and (df[col] < checks['min']).any():
            errors.append(f"{col}: values below minimum {checks['min']}")

        if 'max' in checks and (df[col] > checks['max']).any():
            errors.append(f"{col}: values above maximum {checks['max']}")

        if 'values' in checks:
            invalid = ~df[col].isin(checks['values'])
            if invalid.any():
                bad_vals = df.loc[invalid, col].unique()[:5]
                errors.append(f"{col}: invalid values {list(bad_vals)}")

    return errors

# Usage
errors = validate_dataframe(df, {
    'email': {'not_null': True, 'unique': True},
    'age': {'not_null': True, 'min': 0, 'max': 150},
    'status': {'values': ['active', 'inactive', 'pending']},
})
```

## Efficient GroupBy and Aggregation

```python
# Named aggregation (pandas 0.25+)
summary = df.groupby('department').agg(
    total_salary=('salary', 'sum'),
    avg_salary=('salary', 'mean'),
    headcount=('employee_id', 'count'),
    max_tenure=('hire_date', lambda x: (pd.Timestamp.now() - x.min()).days),
).round(2)

# Multiple aggregations on same column
metrics = df.groupby(['year', 'quarter']).agg(
    revenue_sum=('revenue', 'sum'),
    revenue_mean=('revenue', 'mean'),
    revenue_median=('revenue', 'median'),
    order_count=('order_id', 'nunique'),
)

# Window functions
df['rolling_avg_7d'] = (
    df.sort_values('date')
    .groupby('product_id')['revenue']
    .transform(lambda x: x.rolling(7, min_periods=1).mean())
)

df['rank_in_category'] = (
    df.groupby('category')['sales']
    .rank(method='dense', ascending=False)
    .astype(int)
)

df['pct_of_total'] = df['revenue'] / df.groupby('region')['revenue'].transform('sum')
```

## Merging and Joining

```python
# Validate merge results
def safe_merge(
    left: pd.DataFrame,
    right: pd.DataFrame,
    on: str | list[str],
    how: str = 'left',
    validate: str = 'many_to_one',
) -> pd.DataFrame:
    """Merge with validation and diagnostics."""
    result = left.merge(right, on=on, how=how, validate=validate, indicator=True)

    merge_stats = result['_merge'].value_counts()
    print(f"Merge results: {dict(merge_stats)}")

    if how == 'left':
        unmatched = (result['_merge'] == 'left_only').sum()
        if unmatched > 0:
            print(f"Warning: {unmatched} left rows had no match")

    result = result.drop(columns='_merge')
    return result


# Anti-join: rows in left NOT in right
def anti_join(left: pd.DataFrame, right: pd.DataFrame, on: str) -> pd.DataFrame:
    """Return rows from left that have no match in right."""
    merged = left.merge(right[[on]], on=on, how='left', indicator=True)
    return merged[merged['_merge'] == 'left_only'].drop(columns='_merge')
```

## Performance Optimization

```python
# 1. Use categories for low-cardinality strings
df['country'] = df['country'].astype('category')  # 80% memory reduction

# 2. Vectorized operations (avoid .apply() with Python functions)
# Bad (slow):
df['full_name'] = df.apply(lambda row: f"{row['first']} {row['last']}", axis=1)
# Good (fast):
df['full_name'] = df['first'] + ' ' + df['last']

# 3. Use .query() for complex filters (faster than chained boolean indexing)
result = df.query('age > 25 and status == "active" and revenue > 1000')

# 4. Avoid iterrows — use vectorized operations or numpy
# Bad:
for idx, row in df.iterrows():
    df.at[idx, 'tax'] = row['price'] * 0.08
# Good:
df['tax'] = df['price'] * 0.08

# 5. Use pyarrow backend (pandas 2.0+)
df = pd.read_csv('data.csv', engine='pyarrow', dtype_backend='pyarrow')
# 2-10x faster reads, lower memory
```

## Gotchas

1. **SettingWithCopyWarning** — `df[df['col'] > 5]['other'] = 1` modifies a copy, not the original. Use `.loc`: `df.loc[df['col'] > 5, 'other'] = 1`. In pandas 2.0+, use `df.copy()` explicitly.

2. **Integer columns with NaN become float** — `pd.Series([1, 2, None])` is `float64` because numpy int can't represent NaN. Use `Int64` (capital I) nullable integer: `pd.array([1, 2, None], dtype='Int64')`.

3. **groupby drops NaN groups by default** — `df.groupby('col')` excludes rows where `col` is NaN. Use `dropna=False`: `df.groupby('col', dropna=False)`.

4. **merge creates duplicate column names silently** — merging DataFrames with overlapping non-key columns adds `_x` and `_y` suffixes silently. Explicitly select columns before merge or use `suffixes=('_left', '_right')`.

5. **read_csv infers types wrong for mixed columns** — a column with `['1', '2', 'N/A', '3']` becomes object, not numeric. Use `na_values=['N/A']` and explicit `dtype` to control parsing.

6. **chained indexing is unpredictable** — `df['col1']['col2']` may return a view or a copy depending on the operation. Always use single `.loc[row, col]` or `.iloc[row, col]` indexing.
