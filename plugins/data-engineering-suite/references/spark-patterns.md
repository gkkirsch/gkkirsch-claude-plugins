# Apache Spark Patterns & Best Practices Reference

## Cluster Sizing Quick Reference

### Executor Configuration

| Cluster Size | Executors | Cores/Executor | Memory/Executor | Driver Memory |
|-------------|-----------|----------------|-----------------|---------------|
| Small (<100GB) | 4-8 | 4 | 8g | 4g |
| Medium (100GB-1TB) | 10-30 | 5 | 16g | 8g |
| Large (1TB-10TB) | 30-100 | 5 | 24g | 16g |
| XL (>10TB) | 100+ | 5 | 32g | 32g |

**Rules of thumb:**
- Leave 1 core per node for OS/YARN overhead
- Leave 1GB per node for OS overhead
- 5 cores per executor is optimal (avoids HDFS throughput bottleneck)
- Memory overhead = max(384MB, 0.1 * executor-memory)

### Dynamic Allocation

```python
spark.conf.set("spark.dynamicAllocation.enabled", True)
spark.conf.set("spark.dynamicAllocation.minExecutors", 2)
spark.conf.set("spark.dynamicAllocation.maxExecutors", 50)
spark.conf.set("spark.dynamicAllocation.initialExecutors", 5)
spark.conf.set("spark.dynamicAllocation.executorIdleTimeout", "60s")
spark.conf.set("spark.dynamicAllocation.schedulerBacklogTimeout", "1s")
```

## Partition Optimization

### Target Partition Size

```
Optimal partition size: 128MB - 256MB (compressed)
Optimal partition count: 2x-4x total cores

# Calculate target partitions
data_size_gb = 500
target_partition_mb = 200
target_partitions = (data_size_gb * 1024) / target_partition_mb  # = 2560
```

### Repartition vs Coalesce

```python
# Repartition: full shuffle, creates equal-sized partitions
# Use when: increasing partitions, need balanced distribution
df = df.repartition(200)                    # by count
df = df.repartition(200, "date", "region")  # by columns (hash partitioning)

# Coalesce: no shuffle, merges adjacent partitions
# Use when: decreasing partitions (e.g., before writing output)
df = df.coalesce(10)

# NEVER coalesce then repartition — defeats the purpose
# NEVER repartition to 1 — driver OOM risk, zero parallelism
```

### Partition Skew Detection

```python
from pyspark.sql.functions import spark_partition_id, count

# Check partition sizes
df.withColumn("partition_id", spark_partition_id()) \
    .groupBy("partition_id") \
    .agg(count("*").alias("row_count")) \
    .orderBy("row_count", ascending=False) \
    .show(20)

# If max/min ratio > 10x, you have skew
```

## Join Optimization

### Join Strategy Decision Tree

```
Both datasets fit in memory?
  └─ Yes → Broadcast join (fastest)

One dataset < 10MB (default)?
  └─ Yes → Auto broadcast join
  └─ Tune: spark.sql.autoBroadcastJoinThreshold

Both large, join key has skew?
  └─ Yes → Salted join (add random prefix to break hot keys)

Both large, join key is sortable?
  └─ Yes → Sort-merge join (default for large-large)
  └─ Pre-sort with bucketing for repeated joins

Neither sortable nor broadcastable?
  └─ Shuffle hash join (spark.sql.join.preferSortMergeJoin=false)
```

### Broadcast Join

```python
from pyspark.sql.functions import broadcast

# Explicit broadcast (override threshold)
result = large_df.join(broadcast(small_df), "join_key")

# Configure threshold (bytes)
spark.conf.set("spark.sql.autoBroadcastJoinThreshold", 50 * 1024 * 1024)  # 50MB
```

### Salted Join for Skew

```python
import pyspark.sql.functions as F

SALT_BUCKETS = 10

# Salt the skewed side
skewed_df = skewed_df.withColumn(
    "salt", (F.rand() * SALT_BUCKETS).cast("int")
)
skewed_df = skewed_df.withColumn(
    "salted_key", F.concat(F.col("join_key"), F.lit("_"), F.col("salt"))
)

# Explode the other side to match all salt values
other_df = other_df.crossJoin(
    spark.range(SALT_BUCKETS).withColumnRenamed("id", "salt")
)
other_df = other_df.withColumn(
    "salted_key", F.concat(F.col("join_key"), F.lit("_"), F.col("salt"))
)

# Join on salted key — evenly distributed
result = skewed_df.join(other_df, "salted_key")
result = result.drop("salt", "salted_key")
```

### Bucketing for Repeated Joins

```python
# Write bucketed table (one-time cost)
df.write \
    .bucketBy(256, "customer_id") \
    .sortBy("customer_id") \
    .saveAsTable("orders_bucketed")

# Subsequent joins on customer_id skip shuffle entirely
orders = spark.table("orders_bucketed")
customers = spark.table("customers_bucketed")  # Same bucket count + key
result = orders.join(customers, "customer_id")  # No shuffle!
```

## Caching Strategy

```python
from pyspark import StorageLevel

# Cache: memory-only, deserialized (fastest reads, most memory)
df.cache()  # Same as df.persist(StorageLevel.MEMORY_ONLY)

# Persist with overflow to disk
df.persist(StorageLevel.MEMORY_AND_DISK)

# Serialized: less memory, more CPU (Kryo recommended)
df.persist(StorageLevel.MEMORY_ONLY_SER)

# When to cache:
# ✓ Dataset used in 2+ actions
# ✓ Dataset is expensive to compute
# ✓ Dataset fits in cluster memory
# ✗ Don't cache if used only once
# ✗ Don't cache if it doesn't fit (causes spill → slower than recompute)

# ALWAYS unpersist when done
df.unpersist()
```

## AQE (Adaptive Query Execution)

```python
# Enable AQE (default in Spark 3.2+)
spark.conf.set("spark.sql.adaptive.enabled", True)

# Auto-coalesce small partitions after shuffle
spark.conf.set("spark.sql.adaptive.coalescePartitions.enabled", True)
spark.conf.set("spark.sql.adaptive.coalescePartitions.minPartitionSize", "64MB")
spark.conf.set("spark.sql.adaptive.advisoryPartitionSizeInBytes", "256MB")

# Auto-convert sort-merge join to broadcast when runtime stats show small table
spark.conf.set("spark.sql.adaptive.autoBroadcastJoinThreshold", "50MB")

# Auto-handle skewed joins
spark.conf.set("spark.sql.adaptive.skewJoin.enabled", True)
spark.conf.set("spark.sql.adaptive.skewJoin.skewedPartitionFactor", 5)
spark.conf.set("spark.sql.adaptive.skewJoin.skewedPartitionThresholdInBytes", "256MB")
```

## I/O Optimization

### File Format Comparison

| Format | Compression | Splittable | Schema Evolution | Best For |
|--------|------------|------------|-----------------|----------|
| Parquet | Excellent (snappy/zstd) | Yes | Column add/remove | Analytics, warehousing |
| ORC | Excellent | Yes | Limited | Hive-ecosystem workloads |
| Delta Lake | Parquet + ACID | Yes | Full (merge schema) | Lakehouse, CDC, upserts |
| Avro | Good | Yes | Full (schema registry) | Streaming, row-oriented |
| CSV | None/gzip | No (gzip) | None | Data exchange, imports |
| JSON | None/gzip | No (gzip) | None | APIs, logs (avoid for analytics) |

### Write Optimization

```python
# Partition by date (most common query filter)
df.write \
    .partitionBy("year", "month") \
    .format("parquet") \
    .option("compression", "zstd") \
    .mode("overwrite") \
    .save("/data/events")

# Avoid over-partitioning: target 128MB-1GB per partition file
# 1M files × 10KB each = disaster (small files problem)

# Compact small files
spark.read.parquet("/data/events/year=2024/month=01") \
    .coalesce(10) \
    .write.mode("overwrite") \
    .parquet("/data/events/year=2024/month=01")
```

### Predicate Pushdown

```python
# Parquet column pruning + predicate pushdown (automatic)
df = spark.read.parquet("/data/events") \
    .select("user_id", "event_type", "amount") \  # Column pruning
    .filter(col("year") == 2024)                   # Partition pruning
    .filter(col("amount") > 100)                   # Predicate pushdown

# Verify pushdown in explain plan
df.explain(True)
# Look for: PushedFilters: [IsNotNull(amount), GreaterThan(amount, 100)]
```

## UDF Performance

```python
# AVOID regular Python UDFs (serialize/deserialize per row)
@udf(returnType=StringType())
def slow_udf(value):  # 10-100x slower than native
    return value.upper()

# USE Pandas UDFs (vectorized, Arrow-based)
from pyspark.sql.functions import pandas_udf
import pandas as pd

@pandas_udf(StringType())
def fast_udf(series: pd.Series) -> pd.Series:  # Near-native speed
    return series.str.upper()

# BEST: Use built-in functions (no serialization overhead)
from pyspark.sql.functions import upper
df = df.withColumn("name_upper", upper(col("name")))
```

## Memory Tuning

```
Executor Memory Layout (16g example):
┌─────────────────────────────────────────┐
│  Reserved (300MB fixed)                 │
├─────────────────────────────────────────┤
│  User Memory: 5.7GB (40% of remaining) │  ← UDFs, data structures
├─────────────────────────────────────────┤
│  Unified Memory: 9.4GB (60%)           │
│  ┌────────────────┬────────────────┐    │
│  │ Storage: 4.7GB │ Execution: 4.7GB│   │  ← Dynamic boundary
│  │ (cached RDDs)  │ (shuffles/joins)│   │
│  └────────────────┴────────────────┘    │
├─────────────────────────────────────────┤
│  Overhead: 1.6GB (10%, min 384MB)      │  ← Off-heap, internal
└─────────────────────────────────────────┘

spark.memory.fraction = 0.6       # Unified memory share
spark.memory.storageFraction = 0.5  # Initial storage/execution split
```

## Common Anti-Patterns

| Anti-Pattern | Why It's Bad | Fix |
|-------------|-------------|-----|
| `collect()` on large data | Pulls all data to driver → OOM | Use `take()`, `show()`, write to storage |
| `count()` for existence check | Scans entire dataset | Use `head(1)` or `isEmpty()` |
| Wide transformations in loops | N shuffles instead of 1 | Batch operations with `union` + single `groupBy` |
| Python UDFs | 10-100x slower than native | Use built-in functions or Pandas UDFs |
| `repartition(1)` before write | Single-threaded write, driver bottleneck | Use `coalesce(N)` with N > 1 |
| Caching everything | Memory pressure, eviction thrashing | Only cache datasets used 2+ times |
| Ignoring explain plan | Unknown shuffle count, missing pushdown | Always check `df.explain()` for new queries |
| String-typed dates | No partition pruning, slow comparisons | Cast to DateType/TimestampType |

## Diagnostic Checklist

```
Performance problem? Check in this order:

1. Spark UI → Stages tab
   □ Any stage taking >10x longer than others? → Skew
   □ Shuffle read/write extremely large? → Reduce shuffle
   □ GC time > 10% of task time? → Memory tuning

2. Spark UI → Storage tab
   □ Cached RDDs showing partial? → Not enough memory
   □ Multiple copies of same data? → Remove duplicate caches

3. Spark UI → SQL tab → query plan
   □ BroadcastHashJoin where expected? → Check threshold
   □ Filter pushed down? → Check predicate pushdown
   □ Exchange (shuffle) nodes → Minimize these

4. Driver logs
   □ "Lost executor" → OOM on executor (increase memory)
   □ "FetchFailedException" → Shuffle service issue
   □ "Container killed by YARN" → Memory overhead too low
```
