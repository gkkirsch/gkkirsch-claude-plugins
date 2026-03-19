# Pandas Cookbook — Advanced Patterns and Performance

Quick reference for advanced Pandas patterns, performance optimization, and common recipes. Use this as a lookup when working on data pipeline or analysis tasks.

---

## Memory Optimization

### Dtype Downcasting

```python
import pandas as pd
import numpy as np

# Integer downcasting
df["col"] = pd.to_numeric(df["col"], downcast="integer")
# float64 -> float32
df["col"] = pd.to_numeric(df["col"], downcast="float")
# Object -> category (when < 50% unique)
df["col"] = df["col"].astype("category")
# Use nullable integer types for columns with NaN
df["col"] = df["col"].astype("Int64")  # Capital I = nullable
# Use string dtype instead of object
df["col"] = df["col"].astype("string")  # pd.StringDtype
# Use boolean dtype
df["col"] = df["col"].astype("boolean")  # pd.BooleanDtype
```

### Reading Large Files

```python
# Read specific columns only
df = pd.read_csv("large.csv", usecols=["id", "name", "value"])

# Read with explicit dtypes
df = pd.read_csv("large.csv", dtype={"id": "int32", "category": "category", "value": "float32"})

# Chunked reading
chunks = pd.read_csv("large.csv", chunksize=100_000)
result = pd.concat(process(chunk) for chunk in chunks)

# Parquet is almost always better than CSV
df = pd.read_parquet("data.parquet", columns=["id", "value"])  # Column pruning
df = pd.read_parquet("data.parquet", filters=[("date", ">=", "2024-01-01")])  # Row filtering

# Feather for fast IPC
df.to_feather("data.feather")
df = pd.read_feather("data.feather")
```

---

## Indexing and Selection

### Performance-aware Selection

```python
# FAST: .loc for label-based, .iloc for position-based
row = df.loc[df["id"] == 42]          # O(n) scan
row = df.set_index("id").loc[42]      # O(1) after index is built

# AVOID: chained indexing (creates copies, unpredictable)
# BAD:  df[df["x"] > 5]["y"] = 10  # May not modify df!
# GOOD: df.loc[df["x"] > 5, "y"] = 10

# Multi-condition filtering
mask = (df["a"] > 5) & (df["b"] < 10) & (df["c"].isin(["x", "y"]))
result = df.loc[mask]

# .query() for readable filters (uses numexpr, can be faster)
result = df.query("a > 5 and b < 10 and c in ['x', 'y']")

# .eval() for computed columns (numexpr, avoids temporaries)
df.eval("d = a * b + c", inplace=True)

# isin with set for large lookups
valid_ids = set(other_df["id"])  # O(1) lookup
df_filtered = df[df["id"].isin(valid_ids)]

# between for range queries
df[df["value"].between(10, 100)]
```

---

## GroupBy Patterns

### Aggregation Recipes

```python
# Named aggregation (Pandas 0.25+)
result = df.groupby("category").agg(
    total=("value", "sum"),
    average=("value", "mean"),
    count=("id", "count"),
    unique_users=("user_id", "nunique"),
    first_date=("date", "min"),
    last_date=("date", "max"),
)

# Multiple aggregations per column
result = df.groupby("cat").agg({
    "value": ["sum", "mean", "std", "min", "max"],
    "count": ["sum"],
})
result.columns = ["_".join(col) for col in result.columns]

# Custom aggregation functions
def iqr(x):
    return x.quantile(0.75) - x.quantile(0.25)

result = df.groupby("cat").agg(
    median=("value", "median"),
    iqr=("value", iqr),
    pct_positive=("value", lambda x: (x > 0).mean()),
)

# transform: return same-shaped output (broadcast back)
df["group_mean"] = df.groupby("cat")["value"].transform("mean")
df["z_score"] = df.groupby("cat")["value"].transform(lambda x: (x - x.mean()) / x.std())
df["pct_of_group"] = df["value"] / df.groupby("cat")["value"].transform("sum")
df["rank_in_group"] = df.groupby("cat")["value"].rank(ascending=False)
df["cumsum_in_group"] = df.groupby("cat")["value"].cumsum()

# filter: return subset of groups matching condition
big_groups = df.groupby("cat").filter(lambda x: len(x) >= 10)
high_value_groups = df.groupby("cat").filter(lambda x: x["value"].mean() > 50)

# apply: arbitrary function per group
def top_n(group, n=3):
    return group.nlargest(n, "value")

result = df.groupby("cat").apply(top_n, n=5).reset_index(drop=True)
```

### GroupBy Performance Tips

```python
# 1. Use categorical dtypes for groupby keys
df["cat"] = df["cat"].astype("category")  # Much faster groupby

# 2. Avoid apply when vectorized alternatives exist
# SLOW: df.groupby("cat")["value"].apply(lambda x: x.sum())
# FAST: df.groupby("cat")["value"].sum()

# 3. Use observed=True with categoricals to skip empty groups
df.groupby("cat", observed=True).sum()

# 4. sort=False if you don't need sorted groups
df.groupby("cat", sort=False).sum()

# 5. Use .pipe() for chaining
(df.groupby("cat")
   .agg(total=("value", "sum"))
   .pipe(lambda x: x[x["total"] > 100])
   .sort_values("total", ascending=False))
```

---

## Merging and Joining

### Join Patterns

```python
# Inner join (only matching rows)
merged = pd.merge(df1, df2, on="id", how="inner")

# Left join (keep all from left)
merged = pd.merge(df1, df2, on="id", how="left")

# Left join with indicator to find unmatched
merged = pd.merge(df1, df2, on="id", how="left", indicator=True)
unmatched = merged[merged["_merge"] == "left_only"]

# Anti-join (rows in df1 NOT in df2)
anti = df1[~df1["id"].isin(df2["id"])]
# Or:
anti = pd.merge(df1, df2, on="id", how="left", indicator=True)
anti = anti[anti["_merge"] == "left_only"].drop(columns="_merge")

# Semi-join (rows in df1 that HAVE a match in df2)
semi = df1[df1["id"].isin(df2["id"])]

# Multiple key join
merged = pd.merge(df1, df2, on=["id", "date"], how="inner")

# Join on different column names
merged = pd.merge(df1, df2, left_on="user_id", right_on="id")

# Merge with suffix for overlapping columns
merged = pd.merge(df1, df2, on="id", suffixes=("_left", "_right"))

# Merge_asof for time-based joining (nearest match)
result = pd.merge_asof(
    df1.sort_values("timestamp"),
    df2.sort_values("timestamp"),
    on="timestamp",
    by="id",
    tolerance=pd.Timedelta("1h"),
    direction="backward",
)

# Multiple DataFrame merge
from functools import reduce
dfs = [df1, df2, df3, df4]
merged = reduce(lambda left, right: pd.merge(left, right, on="id", how="outer"), dfs)
```

### Performance: Merge vs Join

```python
# .merge() is the standard method for SQL-like joins
# .join() is a convenience method for index-based joins

# For large datasets:
# 1. Set index on join key before joining (one-time cost, faster repeated joins)
df1_indexed = df1.set_index("id")
df2_indexed = df2.set_index("id")
result = df1_indexed.join(df2_indexed, how="inner")

# 2. Use map() for simple lookups (faster than merge for 1:1)
lookup = df2.set_index("id")["value"]
df1["value_from_df2"] = df1["id"].map(lookup)
```

---

## String Operations

### Vectorized String Methods

```python
# Access via .str accessor
df["name"].str.lower()
df["name"].str.upper()
df["name"].str.strip()
df["name"].str.replace("old", "new")
df["name"].str.replace(r"\s+", " ", regex=True)
df["name"].str.split(" ")
df["name"].str.split(" ").str[0]  # First word
df["name"].str.contains("pattern", regex=True, na=False)
df["name"].str.startswith("prefix")
df["name"].str.endswith("suffix")
df["name"].str.len()
df["name"].str.count(r"\d")  # Count digits
df["name"].str.extract(r"(\d+)")  # Extract first number
df["name"].str.extractall(r"(\d+)")  # Extract all numbers
df["name"].str.findall(r"\w+")  # Find all words
df["name"].str.cat(df["surname"], sep=" ")  # Concatenate columns
df["name"].str.pad(10, side="left", fillchar="0")  # Pad
df["name"].str.zfill(5)  # Zero-pad
df["name"].str.slice(0, 5)  # Substring
df["name"].str.get_dummies(sep=",")  # One-hot from delimited
```

---

## DateTime Operations

### DateTime Recipes

```python
# Parse dates
df["date"] = pd.to_datetime(df["date_str"])
df["date"] = pd.to_datetime(df["date_str"], format="%Y-%m-%d")
df["date"] = pd.to_datetime(df["date_str"], errors="coerce")  # NaT for unparseable

# Extract components
df["year"] = df["date"].dt.year
df["month"] = df["date"].dt.month
df["day"] = df["date"].dt.day
df["dayofweek"] = df["date"].dt.dayofweek  # 0=Monday
df["day_name"] = df["date"].dt.day_name()
df["hour"] = df["date"].dt.hour
df["quarter"] = df["date"].dt.quarter
df["is_weekend"] = df["date"].dt.dayofweek >= 5
df["week"] = df["date"].dt.isocalendar().week

# Date arithmetic
df["days_since"] = (pd.Timestamp.now() - df["date"]).dt.days
df["next_month"] = df["date"] + pd.DateOffset(months=1)
df["week_start"] = df["date"] - pd.to_timedelta(df["date"].dt.dayofweek, unit="d")

# Resampling time series
daily = df.set_index("date").resample("D")["value"].sum()
weekly = df.set_index("date").resample("W")["value"].sum()
monthly = df.set_index("date").resample("ME")["value"].agg(["sum", "mean", "count"])

# Business day operations
df["next_business_day"] = df["date"] + pd.offsets.BDay(1)
df["is_business_day"] = df["date"].dt.dayofweek < 5

# Timezone handling
df["date_utc"] = df["date"].dt.tz_localize("UTC")
df["date_eastern"] = df["date_utc"].dt.tz_convert("US/Eastern")
```

---

## Reshaping

### Pivot, Melt, Stack

```python
# Pivot: long -> wide
wide = df.pivot(index="date", columns="category", values="value")

# Pivot table (with aggregation)
pt = pd.pivot_table(df, index="date", columns="cat", values="value", aggfunc="sum", fill_value=0)

# Melt: wide -> long
long = pd.melt(df, id_vars=["date"], value_vars=["col_a", "col_b"], var_name="metric", value_name="value")

# Stack/unstack
stacked = df.set_index(["date", "category"])["value"].unstack("category")
unstacked = stacked.stack().reset_index(name="value")

# Explode: list column -> one row per element
df_exploded = df.explode("tags")

# Cross-tab
ct = pd.crosstab(df["category"], df["region"], values=df["value"], aggfunc="sum", margins=True)
```

---

## Window Functions

### Rolling, Expanding, EWM

```python
# Rolling (fixed window)
df["ma_7"] = df["value"].rolling(7).mean()
df["ma_30"] = df["value"].rolling(30).mean()
df["rolling_std"] = df["value"].rolling(7).std()
df["rolling_min"] = df["value"].rolling(7).min()
df["rolling_max"] = df["value"].rolling(7).max()
df["rolling_sum"] = df["value"].rolling(7).sum()
df["rolling_median"] = df["value"].rolling(7).median()
df["rolling_count"] = df["value"].rolling(7).count()

# Rolling with min_periods (handle start of series)
df["ma_7"] = df["value"].rolling(7, min_periods=1).mean()

# Center-aligned rolling
df["ma_centered"] = df["value"].rolling(7, center=True).mean()

# Rolling with groupby
df["group_ma_7"] = df.groupby("category")["value"].transform(
    lambda x: x.rolling(7, min_periods=1).mean()
)

# Expanding (cumulative)
df["cumsum"] = df["value"].expanding().sum()
df["cummean"] = df["value"].expanding().mean()
df["cummax"] = df["value"].expanding().max()
df["cummin"] = df["value"].expanding().min()

# Exponentially weighted moving average
df["ema_7"] = df["value"].ewm(span=7).mean()
df["ema_30"] = df["value"].ewm(span=30).mean()

# Shift (lag/lead)
df["prev_value"] = df["value"].shift(1)   # Lag
df["next_value"] = df["value"].shift(-1)  # Lead
df["diff"] = df["value"].diff()            # Difference from previous
df["pct_change"] = df["value"].pct_change()  # Percent change

# Rank
df["rank"] = df["value"].rank(ascending=False)
df["pct_rank"] = df["value"].rank(pct=True)
df["group_rank"] = df.groupby("category")["value"].rank(ascending=False)
```

---

## Performance Patterns

### Vectorization Over Loops

```python
# BAD: iterrows (extremely slow)
for idx, row in df.iterrows():
    df.loc[idx, "new"] = row["a"] * row["b"]

# GOOD: vectorized operation
df["new"] = df["a"] * df["b"]

# BAD: apply with Python function
df["new"] = df.apply(lambda row: row["a"] * row["b"], axis=1)

# GOOD: np.where for conditional
df["label"] = np.where(df["value"] > 50, "high", "low")

# GOOD: np.select for multiple conditions
conditions = [
    df["value"] > 100,
    df["value"] > 50,
    df["value"] > 0,
]
choices = ["very_high", "high", "low"]
df["tier"] = np.select(conditions, choices, default="zero")

# GOOD: pd.cut for binning
df["bin"] = pd.cut(df["value"], bins=[0, 25, 50, 75, 100], labels=["Q1", "Q2", "Q3", "Q4"])

# GOOD: pd.qcut for quantile binning
df["quartile"] = pd.qcut(df["value"], q=4, labels=["Q1", "Q2", "Q3", "Q4"])

# GOOD: map for lookups
mapping = {"A": 1, "B": 2, "C": 3}
df["encoded"] = df["category"].map(mapping)
```

### Copy vs View

```python
# ALWAYS .copy() when you want to modify a subset
subset = df[df["x"] > 5].copy()  # Safe to modify
subset["new_col"] = 42

# Without .copy(), you get a view (modifications affect original)
# This triggers SettingWithCopyWarning
```

---

## Common Recipes

### Deduplication

```python
# Simple dedupe
df.drop_duplicates()
df.drop_duplicates(subset=["id"])
df.drop_duplicates(subset=["id"], keep="last")
df.drop_duplicates(subset=["id", "date"])

# Keep row with max value per group
df.sort_values("value").drop_duplicates(subset=["id"], keep="last")

# Or using idxmax
idx = df.groupby("id")["value"].idxmax()
df_deduped = df.loc[idx]
```

### Missing Data

```python
# Detect
df.isnull().sum()
df.isnull().mean()  # Percentage
df[df.isnull().any(axis=1)]  # Rows with any null

# Fill
df.fillna(0)
df.fillna({"col_a": 0, "col_b": "unknown"})
df.fillna(method="ffill")  # Forward fill
df.fillna(method="bfill")  # Backward fill
df["col"].interpolate(method="linear")

# Fill with group statistics
df["value"] = df.groupby("cat")["value"].transform(lambda x: x.fillna(x.median()))

# Drop
df.dropna()  # Drop rows with any null
df.dropna(subset=["important_col"])  # Only check specific columns
df.dropna(thresh=5)  # Keep rows with at least 5 non-null values
```

### Type Conversion

```python
# Safe numeric conversion
df["col"] = pd.to_numeric(df["col"], errors="coerce")  # NaN for failures
df["col"] = pd.to_numeric(df["col"], errors="coerce", downcast="integer")

# Safe datetime conversion
df["date"] = pd.to_datetime(df["date"], errors="coerce")  # NaT for failures

# Categorical with specific order
df["size"] = pd.Categorical(df["size"], categories=["S", "M", "L", "XL"], ordered=True)
```

### Output Formatting

```python
# Display settings
pd.set_option("display.max_columns", None)
pd.set_option("display.max_rows", 100)
pd.set_option("display.width", None)
pd.set_option("display.float_format", "{:,.2f}".format)
pd.set_option("display.max_colwidth", 50)

# Style for Jupyter
(df.style
    .format({"price": "${:,.2f}", "pct": "{:.1%}"})
    .bar(subset=["value"], color="#5fba7d")
    .highlight_max(subset=["score"], color="lightgreen")
    .highlight_min(subset=["score"], color="lightcoral"))
```

---

## Pandas vs Polars Quick Comparison

| Operation | Pandas | Polars |
|-----------|--------|--------|
| Read CSV | `pd.read_csv()` | `pl.read_csv()` / `pl.scan_csv()` |
| Filter | `df[df["x"] > 5]` | `df.filter(pl.col("x") > 5)` |
| Select | `df[["a", "b"]]` | `df.select("a", "b")` |
| New column | `df["c"] = df["a"] + df["b"]` | `df.with_columns((pl.col("a") + pl.col("b")).alias("c"))` |
| Group agg | `df.groupby("x").agg(...)` | `df.group_by("x").agg(...)` |
| Sort | `df.sort_values("x")` | `df.sort("x")` |
| Join | `pd.merge(df1, df2, on="id")` | `df1.join(df2, on="id")` |
| Null fill | `df.fillna(0)` | `df.fill_null(0)` |
| Apply | `df["x"].apply(fn)` | `df.with_columns(pl.col("x").map_elements(fn))` |

Rule of thumb: Use Pandas for < 1GB exploratory work, Polars for anything performance-critical or > 1GB.

---

## Multi-Index Operations

### Working with MultiIndex

```python
# Create MultiIndex
df = df.set_index(["region", "product"])

# Access levels
df.loc["North"]                        # All products in North
df.loc[("North", "Widget")]           # Specific product
df.loc["North":"South"]               # Range of first level
df.xs("Widget", level="product")      # Cross-section

# Reset specific levels
df.reset_index(level="product")

# Swap levels
df.swaplevel(0, 1)

# Sort MultiIndex (required for slicing)
df = df.sort_index()

# Aggregate at different levels
df.groupby(level="region").sum()
df.groupby(level=["region"]).agg({"value": ["sum", "mean"]})

# Flatten MultiIndex columns
df.columns = ["_".join(col).strip("_") for col in df.columns.values]
```

---

## Pipe Pattern for Clean Chains

### Method Chaining with pipe()

```python
def remove_outliers(df, column, n_std=3):
    mean = df[column].mean()
    std = df[column].std()
    return df[df[column].between(mean - n_std * std, mean + n_std * std)]

def add_features(df, date_col):
    return df.assign(
        month=lambda x: x[date_col].dt.month,
        dayofweek=lambda x: x[date_col].dt.dayofweek,
        is_weekend=lambda x: x[date_col].dt.dayofweek >= 5,
    )

def top_n_per_group(df, group_col, value_col, n=10):
    return df.groupby(group_col).apply(
        lambda x: x.nlargest(n, value_col)
    ).reset_index(drop=True)

# Clean pipeline with pipe
result = (
    pd.read_csv("data.csv")
    .pipe(lambda df: df.dropna(subset=["id", "value"]))
    .pipe(remove_outliers, column="value", n_std=3)
    .pipe(add_features, date_col="date")
    .pipe(top_n_per_group, group_col="category", value_col="value", n=5)
    .sort_values(["category", "value"], ascending=[True, False])
)
```

---

## IO Patterns

### Reading from Various Sources

```python
# SQL with SQLAlchemy
from sqlalchemy import create_engine
engine = create_engine("postgresql://user:pass@host:5432/db")
df = pd.read_sql("SELECT * FROM table WHERE date > '2024-01-01'", engine)
df = pd.read_sql_table("table_name", engine)

# Excel with specific sheets
df = pd.read_excel("data.xlsx", sheet_name="Sheet1")
all_sheets = pd.read_excel("data.xlsx", sheet_name=None)  # Dict of DataFrames

# JSON (various formats)
df = pd.read_json("data.json")                    # Default orient
df = pd.read_json("data.ndjson", lines=True)      # Newline-delimited

# HTML tables from web pages
tables = pd.read_html("https://example.com/table")
df = tables[0]  # First table

# Clipboard (useful in Jupyter)
# df = pd.read_clipboard()

# From dict
df = pd.DataFrame.from_dict({"a": [1, 2], "b": [3, 4]})
df = pd.DataFrame.from_records([{"a": 1, "b": 2}, {"a": 3, "b": 4}])
```

### Writing Optimized Output

```python
# Parquet (recommended for data pipelines)
df.to_parquet("data.parquet", engine="pyarrow", compression="snappy")
df.to_parquet("data.parquet", engine="pyarrow", compression="zstd")  # Better compression

# CSV with specific options
df.to_csv("data.csv", index=False, float_format="%.4f")

# Excel with formatting
with pd.ExcelWriter("report.xlsx", engine="openpyxl") as writer:
    df1.to_excel(writer, sheet_name="Summary", index=False)
    df2.to_excel(writer, sheet_name="Details", index=False)

# SQL (append or replace)
df.to_sql("table_name", engine, if_exists="append", index=False, chunksize=10_000)

# HDF5 for very large datasets
df.to_hdf("data.h5", key="dataset", mode="w", complevel=5, complib="blosc")
df = pd.read_hdf("data.h5", key="dataset")
```

---

## Advanced Aggregations

### Custom Aggregation Patterns

```python
# Weighted average
def weighted_avg(group, value_col, weight_col):
    return (group[value_col] * group[weight_col]).sum() / group[weight_col].sum()

result = df.groupby("category").apply(weighted_avg, "price", "quantity")

# Percentile aggregation
result = df.groupby("cat")["value"].agg([
    ("p25", lambda x: x.quantile(0.25)),
    ("p50", lambda x: x.quantile(0.50)),
    ("p75", lambda x: x.quantile(0.75)),
    ("p90", lambda x: x.quantile(0.90)),
    ("p95", lambda x: x.quantile(0.95)),
    ("p99", lambda x: x.quantile(0.99)),
])

# Mode per group
mode_per_group = df.groupby("cat")["subcategory"].agg(lambda x: x.mode().iloc[0] if len(x.mode()) > 0 else None)

# Describe per group
description = df.groupby("cat")["value"].describe()

# First/last per group (sorted)
first_per_group = df.sort_values("date").groupby("cat").first()
last_per_group = df.sort_values("date").groupby("cat").last()

# Cumulative operations per group
df["cum_sum"] = df.groupby("cat")["value"].cumsum()
df["cum_max"] = df.groupby("cat")["value"].cummax()
df["cum_count"] = df.groupby("cat").cumcount()

# Running difference within group
df["diff"] = df.groupby("cat")["value"].diff()
df["pct_change"] = df.groupby("cat")["value"].pct_change()
```

---

## Categorical Data

### Working with Categories

```python
# Create categorical
df["size"] = pd.Categorical(
    df["size"],
    categories=["XS", "S", "M", "L", "XL", "XXL"],
    ordered=True,
)

# Advantages of categorical:
# 1. Memory savings (stores integers + mapping)
# 2. Faster groupby/sort
# 3. Logical ordering for comparisons
df[df["size"] > "M"]  # Works with ordered categories

# Add/remove categories
df["size"] = df["size"].cat.add_categories(["XXXL"])
df["size"] = df["size"].cat.remove_unused_categories()

# Rename categories
df["size"] = df["size"].cat.rename_categories({"XS": "Extra Small"})

# Convert to codes (integer encoding)
df["size_code"] = df["size"].cat.codes
```

---

## Sparse Data

### Working with Sparse Arrays

```python
# Create sparse column (efficient for mostly-zero data)
df["sparse_col"] = pd.arrays.SparseArray([0, 0, 1, 0, 0, 2, 0, 0, 0, 3])

# Convert dense to sparse
df["sparse"] = df["dense"].astype(pd.SparseDtype("float64", fill_value=0.0))

# Check memory savings
print(f"Dense: {df['dense'].memory_usage():,} bytes")
print(f"Sparse: {df['sparse'].memory_usage():,} bytes")

# get_dummies with sparse output
dummies = pd.get_dummies(df["category"], sparse=True, dtype=int)
```

---

## Useful Utility Functions

### DataFrame Inspection

```python
# Quick data profiling
def quick_profile(df):
    """Print a quick profile of a DataFrame."""
    print(f"Shape: {df.shape}")
    print(f"Memory: {df.memory_usage(deep=True).sum() / 1024**2:.2f} MB")
    print(f"\nDtypes:\n{df.dtypes.value_counts()}")
    print(f"\nMissing:\n{df.isnull().sum()[df.isnull().sum() > 0]}")
    print(f"\nNumeric stats:\n{df.describe().T[['mean', 'std', 'min', 'max']]}")

# Compare two DataFrames
def compare_dfs(df1, df2, on=None):
    """Compare two DataFrames and show differences."""
    if on:
        merged = pd.merge(df1, df2, on=on, how="outer", indicator=True, suffixes=("_old", "_new"))
        print("Left only:", (merged["_merge"] == "left_only").sum())
        print("Right only:", (merged["_merge"] == "right_only").sum())
        print("Both:", (merged["_merge"] == "both").sum())
        return merged
    else:
        print(f"Shape: {df1.shape} vs {df2.shape}")
        print(f"Columns match: {set(df1.columns) == set(df2.columns)}")
        if df1.shape == df2.shape:
            diff = (df1 != df2).sum().sum()
            print(f"Different cells: {diff}")

# Detect potential ID columns
def find_id_columns(df, threshold=0.95):
    """Find columns that could be identifiers (high cardinality)."""
    return [col for col in df.columns if df[col].nunique() / len(df) > threshold]

# Detect constant columns
def find_constant_columns(df):
    """Find columns with only one unique value."""
    return [col for col in df.columns if df[col].nunique() <= 1]

# Detect highly correlated features
def find_correlated_features(df, threshold=0.95):
    """Find pairs of highly correlated numeric features."""
    corr = df.select_dtypes(include="number").corr().abs()
    upper = corr.where(np.triu(np.ones(corr.shape), k=1).astype(bool))
    pairs = []
    for col in upper.columns:
        for idx in upper.index:
            if upper.loc[idx, col] > threshold:
                pairs.append((idx, col, upper.loc[idx, col]))
    return sorted(pairs, key=lambda x: x[2], reverse=True)
```

---

## Error-prone Patterns to Avoid

| Anti-pattern | Problem | Fix |
|-------------|---------|-----|
| `df[col][row] = val` | Chained indexing, may not work | `df.loc[row, col] = val` |
| `for idx, row in df.iterrows()` | Extremely slow | Vectorize or use `apply` |
| `df.append(row)` | Deprecated, slow in loops | Build list, `pd.concat` once |
| `df = df.drop(...)` without `inplace` | Creates copy, wastes memory | Use `inplace=True` or reassign |
| `df.merge(...)` without `how=` | Defaults to inner join | Always specify `how=` |
| `pd.read_csv(...)` no dtype | Auto-detection is slow and wrong | Specify `dtype=` |
| `df.apply(fn, axis=1)` | Row-wise apply is slow | Use vectorized operations |
| `df.groupby().apply(fn)` | Slow for simple aggs | Use built-in `.sum()`, `.mean()` |
| Comparing with `==` to NaN | `NaN != NaN` is True | Use `pd.isna()` or `.isna()` |
| Not using `.copy()` on subsets | SettingWithCopyWarning | Always `.copy()` before modifying |
