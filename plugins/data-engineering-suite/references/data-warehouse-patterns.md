# Data Warehouse Patterns & Best Practices Reference

## Warehouse Architecture

### Modern Data Stack

```
Sources              Ingestion         Storage           Transform        Serve
────────            ──────────        ─────────          ──────────       ──────
PostgreSQL    ──┐                    ┌──────────┐
MySQL         ──┤   Fivetran/       │          │       ┌──────┐
MongoDB       ──┤   Airbyte/    ──► │  Cloud   │  ──►  │ dbt  │  ──►  BI Tools
APIs          ──┤   Custom ETL      │Warehouse │       └──────┘       Dashboards
S3/Files      ──┤                   │          │                      ML Models
Event Streams ──┘                   └──────────┘                      APIs
                                    Snowflake / BigQuery /
                                    Redshift / Databricks
```

### Warehouse Layer Architecture

```
┌─────────────────────────────────────────────┐
│  CONSUMPTION LAYER (marts)                   │
│  dim_customers, fct_revenue, rpt_daily_kpi  │
│  Optimized for BI queries, denormalized      │
├─────────────────────────────────────────────┤
│  BUSINESS LOGIC LAYER (intermediate)         │
│  int_orders_enriched, int_sessions_mapped    │
│  Reusable building blocks, complex joins     │
├─────────────────────────────────────────────┤
│  STAGING LAYER (staging)                     │
│  stg_orders, stg_customers, stg_events      │
│  1:1 with source, light cleaning only        │
├─────────────────────────────────────────────┤
│  RAW LAYER (raw)                             │
│  raw_orders, raw_events, raw_api_responses   │
│  Exact copy from source, append-only         │
└─────────────────────────────────────────────┘
```

### Layer Rules

| Layer | Materialization | Naming | Joins? | Business Logic? |
|-------|----------------|--------|--------|-----------------|
| Raw | Table (append) | `raw_{source}_{entity}` | Never | Never |
| Staging | View or Table | `stg_{source}_{entity}` | Never | Renaming, casting, filtering only |
| Intermediate | Ephemeral or View | `int_{entity}_{verb}` | Yes | Yes — building blocks |
| Marts | Table or Incremental | `fct_` / `dim_` / `rpt_` | Yes | Yes — final business definitions |

## Dimensional Modeling Patterns

### Fact Table Types

```
Transaction Fact:
  One row per business event (order, click, payment)
  Most common. Additive measures (amount, quantity).
  Example: fct_orders

Periodic Snapshot Fact:
  One row per entity per time period (daily balance, monthly inventory)
  Semi-additive: sum across dimensions but not time.
  Example: fct_daily_account_balance

Accumulating Snapshot Fact:
  One row per lifecycle instance (order lifecycle, claim lifecycle)
  Multiple date columns tracking milestones.
  Example: fct_order_lifecycle

Factless Fact:
  One row per event with no measures (attendance, coverage)
  Used to track what happened or what's possible.
  Example: fct_student_attendance
```

### Dimension Types

```
Conformed Dimension:
  Shared across multiple fact tables (dim_date, dim_customer)
  Single source of truth. Critical for cross-process analysis.

Role-Playing Dimension:
  Same dimension used multiple times with different roles
  Example: dim_date as order_date, ship_date, delivery_date
  Implement as views or aliases, not copies.

Degenerate Dimension:
  Dimension attribute stored in fact table (no lookup)
  Example: order_number, invoice_number, transaction_id
  No separate dimension table needed.

Junk Dimension:
  Bag of low-cardinality flags combined into one dimension
  Example: dim_order_flags (is_gift, is_rush, payment_method)
  Keeps fact table narrow.

Mini-Dimension:
  Subset of a large dimension that changes frequently
  Example: dim_customer_demographics (age band, income band)
  Separate from main dim_customer to reduce SCD2 row explosion.
```

## Partitioning Strategies

### By Warehouse Platform

```sql
-- BigQuery: Partition + Cluster
CREATE TABLE analytics.fct_events
PARTITION BY DATE(event_time)
CLUSTER BY user_id, event_type
AS SELECT ...;

-- Snowflake: Automatic micro-partitioning + clustering
CREATE TABLE analytics.fct_events
CLUSTER BY (event_date, user_id)
AS SELECT ...;
-- Snowflake auto-partitions. Cluster keys optimize pruning.

-- Redshift: Distribution + Sort Keys
CREATE TABLE analytics.fct_events (
    event_id BIGINT,
    user_id BIGINT,
    event_date DATE,
    event_type VARCHAR(50),
    amount DECIMAL(12,2)
)
DISTSTYLE KEY
DISTKEY(user_id)       -- Co-locate user data for joins
SORTKEY(event_date);   -- Optimize range scans by date

-- PostgreSQL: Declarative Partitioning
CREATE TABLE analytics.fct_events (
    event_id BIGINT,
    event_date DATE,
    user_id BIGINT
) PARTITION BY RANGE (event_date);

CREATE TABLE fct_events_2024_q1
    PARTITION OF analytics.fct_events
    FOR VALUES FROM ('2024-01-01') TO ('2024-04-01');
```

### Partition Key Selection

```
Choose partition key based on:
1. Most common WHERE clause filter (usually date)
2. Data lifecycle management needs (drop old partitions)
3. Even data distribution across partitions

Guidelines:
  ✓ Date/timestamp columns (daily, monthly)
  ✓ Target 100MB-1GB per partition (compressed)
  ✓ < 10,000 partitions total per table
  ✗ Don't partition on high-cardinality columns (user_id → millions of partitions)
  ✗ Don't partition on columns rarely filtered on
```

## Query Optimization

### Common Anti-Patterns

```sql
-- BAD: SELECT * (reads all columns)
SELECT * FROM fct_orders WHERE date = '2024-03-01';

-- GOOD: Select only needed columns
SELECT order_id, customer_id, amount
FROM fct_orders
WHERE date = '2024-03-01';

-- BAD: Cartesian join (missing join condition)
SELECT * FROM orders, customers;  -- N × M rows!

-- GOOD: Explicit join
SELECT * FROM orders JOIN customers USING (customer_id);

-- BAD: DISTINCT to fix duplicates (hiding a join problem)
SELECT DISTINCT o.order_id, c.name
FROM orders o JOIN customers c ON ...;

-- GOOD: Fix the join that's causing duplication
-- (usually a missing condition or wrong grain)

-- BAD: Subquery in WHERE for large datasets
SELECT * FROM orders
WHERE customer_id IN (SELECT customer_id FROM vip_customers);

-- GOOD: Use JOIN instead
SELECT o.* FROM orders o
JOIN vip_customers v ON o.customer_id = v.customer_id;

-- BAD: Functions on indexed/partition columns
SELECT * FROM events WHERE DATE(event_time) = '2024-03-01';

-- GOOD: Range condition (enables partition pruning)
SELECT * FROM events
WHERE event_time >= '2024-03-01' AND event_time < '2024-03-02';
```

### Materialized Views

```sql
-- For expensive queries run repeatedly
CREATE MATERIALIZED VIEW mv_daily_revenue AS
SELECT
    date_trunc('day', ordered_at) as order_date,
    product_category,
    SUM(total_amount) as revenue,
    COUNT(*) as order_count,
    COUNT(DISTINCT customer_id) as unique_customers
FROM fct_orders
JOIN dim_product USING (product_key)
GROUP BY 1, 2;

-- Refresh strategy
REFRESH MATERIALIZED VIEW mv_daily_revenue;           -- Full refresh
REFRESH MATERIALIZED VIEW CONCURRENTLY mv_daily_revenue;  -- No read lock
```

## Data Quality Patterns

### Quality Dimensions

| Dimension | What It Measures | Example Check |
|-----------|-----------------|---------------|
| Completeness | Missing values | NULL rate < 1% for required fields |
| Uniqueness | Duplicate records | Primary key is unique |
| Validity | Data in expected range | Status IN ('active', 'inactive') |
| Accuracy | Data matches reality | Revenue reconciles with source |
| Timeliness | Data is fresh enough | Max timestamp within 24h |
| Consistency | Data matches across systems | Customer count matches CRM |

### Quality Check SQL Templates

```sql
-- Completeness: NULL rate
SELECT
    'orders' as table_name,
    COUNT(*) as total_rows,
    SUM(CASE WHEN customer_id IS NULL THEN 1 ELSE 0 END) as null_customer_id,
    ROUND(100.0 * SUM(CASE WHEN customer_id IS NULL THEN 1 ELSE 0 END) / COUNT(*), 2) as null_pct
FROM fct_orders
WHERE ordered_date = CURRENT_DATE - 1;

-- Freshness: Latest record timestamp
SELECT
    'orders' as table_name,
    MAX(ordered_at) as latest_record,
    EXTRACT(EPOCH FROM CURRENT_TIMESTAMP - MAX(ordered_at)) / 3600 as hours_since_latest
FROM fct_orders
HAVING EXTRACT(EPOCH FROM CURRENT_TIMESTAMP - MAX(ordered_at)) / 3600 > 26;
-- Returns rows only if data is stale (>26 hours)

-- Volume anomaly: Row count vs 7-day average
WITH daily_counts AS (
    SELECT
        ordered_date,
        COUNT(*) as row_count
    FROM fct_orders
    WHERE ordered_date >= CURRENT_DATE - 8
    GROUP BY ordered_date
),
stats AS (
    SELECT
        AVG(row_count) as avg_count,
        STDDEV(row_count) as stddev_count
    FROM daily_counts
    WHERE ordered_date < CURRENT_DATE - 1  -- Exclude today and yesterday
)
SELECT
    dc.ordered_date,
    dc.row_count,
    s.avg_count,
    ABS(dc.row_count - s.avg_count) / NULLIF(s.stddev_count, 0) as z_score
FROM daily_counts dc, stats s
WHERE dc.ordered_date = CURRENT_DATE - 1
  AND ABS(dc.row_count - s.avg_count) / NULLIF(s.stddev_count, 0) > 3;
-- Alert if yesterday's count is >3 standard deviations from 7-day mean

-- Cross-source reconciliation
WITH source_total AS (
    SELECT SUM(amount) as total FROM raw_orders WHERE date = CURRENT_DATE - 1
),
warehouse_total AS (
    SELECT SUM(total_amount) as total FROM fct_orders WHERE ordered_date = CURRENT_DATE - 1
)
SELECT
    s.total as source_total,
    w.total as warehouse_total,
    ABS(s.total - w.total) / NULLIF(s.total, 0) as variance_pct
FROM source_total s, warehouse_total w
WHERE ABS(s.total - w.total) / NULLIF(s.total, 0) > 0.01;
-- Alert if >1% variance between source and warehouse
```

## Data Lifecycle Management

### Retention Policies

```sql
-- Hot tier (recent data, full detail): 0-90 days
-- Keep in primary warehouse, all columns, full granularity

-- Warm tier (recent history, full detail): 90 days - 2 years
-- Move to cheaper storage class, still queryable
-- Snowflake: ALTER TABLE ... SET DATA_RETENTION_TIME_IN_DAYS = 90;
-- BigQuery: Set partition expiration

-- Cold tier (archive): 2+ years
-- Aggregate to daily/weekly grain, archive raw to object storage
-- Keep dimension snapshots, drop raw event detail

-- Tombstone tier: Beyond legal retention requirements
-- Hard delete with audit trail
```

```sql
-- Automated partition cleanup (PostgreSQL)
DO $$
DECLARE
    partition_name TEXT;
BEGIN
    FOR partition_name IN
        SELECT tablename FROM pg_tables
        WHERE schemaname = 'analytics'
        AND tablename LIKE 'fct_events_%'
        AND tablename < 'fct_events_' || to_char(CURRENT_DATE - INTERVAL '2 years', 'YYYY_MM')
    LOOP
        EXECUTE format('DROP TABLE IF EXISTS analytics.%I', partition_name);
        RAISE NOTICE 'Dropped partition: %', partition_name;
    END LOOP;
END $$;
```

## Access Control Patterns

### Role-Based Access

```sql
-- Create roles for different access levels
CREATE ROLE data_analyst;
CREATE ROLE data_engineer;
CREATE ROLE data_scientist;
CREATE ROLE bi_service_account;

-- Analysts: read marts only
GRANT USAGE ON SCHEMA marts TO data_analyst;
GRANT SELECT ON ALL TABLES IN SCHEMA marts TO data_analyst;

-- Engineers: read/write all layers
GRANT ALL ON SCHEMA raw, staging, intermediate, marts TO data_engineer;
GRANT ALL ON ALL TABLES IN SCHEMA raw, staging, intermediate, marts TO data_engineer;

-- Scientists: read staging + marts, write to sandbox
GRANT USAGE ON SCHEMA staging, marts, sandbox TO data_scientist;
GRANT SELECT ON ALL TABLES IN SCHEMA staging, marts TO data_scientist;
GRANT ALL ON SCHEMA sandbox TO data_scientist;

-- BI tools: read marts only via service account
GRANT USAGE ON SCHEMA marts TO bi_service_account;
GRANT SELECT ON ALL TABLES IN SCHEMA marts TO bi_service_account;

-- Row-level security (Snowflake)
CREATE ROW ACCESS POLICY region_policy AS (region_val VARCHAR)
RETURNS BOOLEAN ->
    CURRENT_ROLE() = 'DATA_ADMIN'
    OR region_val = CURRENT_SESSION()::region;

ALTER TABLE fct_revenue ADD ROW ACCESS POLICY region_policy ON (region);
```

## Performance Benchmarking

### Query Performance Checklist

```
Before optimizing, measure:
1. Query execution time (cold cache vs warm cache)
2. Data scanned (bytes — this is your cost on cloud warehouses)
3. Rows processed vs rows returned (filter efficiency)
4. Number of partitions scanned vs total (partition pruning)
5. Spill to disk (memory pressure indicator)

Optimization priority:
1. Reduce data scanned (partition pruning, column pruning)
2. Reduce shuffle/data movement (co-location, broadcast)
3. Reduce compute (pre-aggregation, materialized views)
4. Increase parallelism (more partitions, more nodes)

Target metrics:
  Dashboard queries: < 5 seconds
  Ad-hoc analysis: < 30 seconds
  Heavy ETL transforms: < 30 minutes
  Full pipeline end-to-end: within SLA (usually 2-4 hours)
```

## Naming Conventions

```
Tables:
  raw_{source}_{entity}           raw_stripe_payments
  stg_{source}_{entity}           stg_stripe_payments
  int_{entity}_{verb}             int_payments_pivoted
  fct_{business_process}          fct_revenue
  dim_{entity}                    dim_customer
  rpt_{report_name}               rpt_weekly_kpi
  snap_{entity}                   snap_customer (SCD2 snapshots)
  bridge_{relationship}           bridge_patient_diagnosis
  mv_{description}                mv_daily_revenue (materialized views)

Columns:
  {entity}_id                     customer_id (natural key)
  {entity}_key                    customer_key (surrogate key)
  {entity}_{attribute}            customer_name
  is_{adjective}                  is_active, is_deleted
  has_{noun}                      has_subscription
  {measure}_{unit}                amount_usd, weight_kg
  {measure}_pct                   discount_pct (0.00 to 1.00)
  {event}_at                      created_at, shipped_at (timestamps)
  {event}_date                    ordered_date (date only)
  _loaded_at                      ETL metadata (underscore prefix)
  _source_system                  ETL metadata
```

## Migration Patterns

### Zero-Downtime Table Migration

```sql
-- 1. Create new table with updated schema
CREATE TABLE fct_orders_v2 (
    -- new schema here
);

-- 2. Backfill from old table
INSERT INTO fct_orders_v2
SELECT
    order_id,
    customer_key,
    -- map old columns to new
    COALESCE(new_column, default_value) as new_column
FROM fct_orders;

-- 3. Set up dual-write (ETL writes to both tables)
-- Run in parallel for 1-2 days to verify consistency

-- 4. Verify data matches
SELECT COUNT(*), SUM(amount) FROM fct_orders WHERE date = CURRENT_DATE - 1
UNION ALL
SELECT COUNT(*), SUM(amount) FROM fct_orders_v2 WHERE date = CURRENT_DATE - 1;

-- 5. Swap via rename (instant, atomic on most warehouses)
ALTER TABLE fct_orders RENAME TO fct_orders_deprecated;
ALTER TABLE fct_orders_v2 RENAME TO fct_orders;

-- 6. Update dependent views/models to point to new table
-- 7. After validation period, drop deprecated table
DROP TABLE fct_orders_deprecated;
```
