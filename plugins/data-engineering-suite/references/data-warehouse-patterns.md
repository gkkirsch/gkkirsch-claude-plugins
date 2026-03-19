# Data Warehouse Patterns Reference

## Warehouse Architecture

### Modern Data Stack

```
Sources                    Ingestion           Warehouse              Serving
┌──────────┐    ┌─────────────────────┐    ┌──────────────┐    ┌─────────────┐
│ PostgreSQL│──→ │ Fivetran / Airbyte  │──→ │              │──→ │  Looker     │
│ MySQL     │    │ (EL - raw extract)  │    │  Snowflake   │    │  Metabase   │
│ MongoDB   │    └─────────────────────┘    │  BigQuery    │    │  Superset   │
│ APIs      │                               │  Redshift    │    │  dbt Metrics│
│ S3/GCS    │    ┌─────────────────────┐    │  Databricks  │    └─────────────┘
│ Kafka     │──→ │ dbt (Transform)     │──→ │              │
│ Salesforce│    │ staging → marts     │    │  raw →       │    ┌─────────────┐
│ Stripe    │    └─────────────────────┘    │  staging →   │──→ │  ML / AI    │
└──────────┘                                │  marts       │    │  Feature    │
                                            └──────────────┘    │  Store      │
                                                                └─────────────┘
```

### Layered Architecture

| Layer | Purpose | Materialization | Naming | Example |
|-------|---------|----------------|--------|---------|
| **Raw** | Exact copy of source data | Table (append) | `raw_<source>.<table>` | `raw_stripe.payments` |
| **Staging** | Clean, rename, type-cast | View | `stg_<source>__<entity>` | `stg_stripe__payments` |
| **Intermediate** | Business logic, joins | Ephemeral/View | `int_<entity>__<verb>` | `int_orders__enriched` |
| **Marts** | Business-facing tables | Table/Incremental | `dim_<entity>`, `fct_<verb>` | `dim_customer`, `fct_orders` |
| **Metrics** | Pre-computed KPIs | Table | `metric_<name>` | `metric_monthly_revenue` |

## Partitioning Strategies

### By Warehouse Platform

#### Snowflake

```sql
-- Snowflake uses micro-partitions (automatic, 50-500MB each)
-- Clustering keys help Snowflake organize micro-partitions

-- Cluster by common filter/join columns
CREATE TABLE marts.fct_orders (
    order_id VARCHAR,
    order_date DATE,
    customer_id VARCHAR,
    amount NUMBER(12,2)
)
CLUSTER BY (order_date, customer_id);

-- Re-cluster after heavy writes
ALTER TABLE marts.fct_orders RECLUSTER;

-- Check clustering depth (lower = better)
SELECT SYSTEM$CLUSTERING_DEPTH('marts.fct_orders');
SELECT SYSTEM$CLUSTERING_INFORMATION('marts.fct_orders');
```

#### BigQuery

```sql
-- BigQuery: partition + cluster for optimal performance
CREATE TABLE `project.marts.fct_orders`
PARTITION BY DATE(order_date)          -- Time-based partitioning
CLUSTER BY customer_id, product_id     -- Up to 4 cluster columns
OPTIONS (
    partition_expiration_days = 365,    -- Auto-delete old partitions
    require_partition_filter = TRUE     -- Force partition pruning in queries
)
AS SELECT * FROM `staging.stg_orders`;

-- Integer range partitioning
CREATE TABLE `project.marts.fct_events`
PARTITION BY RANGE_BUCKET(user_id, GENERATE_ARRAY(0, 1000000, 1000))
AS SELECT * FROM staging.events;
```

#### Redshift

```sql
-- Redshift: distribution style + sort keys
CREATE TABLE marts.fct_orders (
    order_id VARCHAR(50),
    order_date DATE,
    customer_id VARCHAR(50),
    amount DECIMAL(12,2)
)
DISTSTYLE KEY           -- Distribute by a key column
DISTKEY(customer_id)    -- Co-locate customer data on same node
COMPOUND SORTKEY(order_date, customer_id);  -- Sort for range queries

-- Distribution styles:
-- EVEN: round-robin (when no good key, or large dimension)
-- KEY: hash by column (co-locate joins)
-- ALL: full copy on each node (small dimensions only, < 5M rows)
-- AUTO: let Redshift decide
```

### Partitioning Decision Guide

```
Choose partition column:
1. Most common WHERE clause filter? → Use that column
2. Time-based data? → Partition by date/month
3. Queries always filter by region/tenant? → Partition by that

Partition size guidelines:
- Each partition: 100MB - 1GB ideal
- Too many partitions (>10,000): metadata overhead
- Too few: no pruning benefit
- No partition: fine for tables < 1GB

Combine with clustering/sort keys:
- Partition = coarse filter (month, year)
- Cluster = fine filter within partition (customer_id, product_id)
```

## Slowly Changing Dimensions (SCDs)

### SCD Type Comparison

| Type | Description | Storage | Complexity | Use When |
|------|-------------|---------|------------|----------|
| **Type 0** | Never changes | Lowest | None | Truly static (birth date, SSN) |
| **Type 1** | Overwrite | Low | Low | Don't need history (email, phone) |
| **Type 2** | New row + versioning | High | Medium | Need full history (address, segment) |
| **Type 3** | Previous + current columns | Medium | Low | Only need one prior value |
| **Type 4** | Separate history table | Medium | Medium | Keep current dim small |
| **Type 6** | Hybrid 1+2+3 | Highest | Highest | Need both history and current in one row |

### SCD Type 2 Implementation (SQL)

```sql
-- Merge pattern for SCD Type 2
-- Step 1: Identify changed records
WITH source AS (
    SELECT
        customer_id,
        customer_name,
        email,
        segment,
        region,
        md5(concat_ws('|',
            coalesce(customer_name, ''),
            coalesce(email, ''),
            coalesce(segment, ''),
            coalesce(region, '')
        )) AS row_hash
    FROM staging.stg_customers
),

current_dim AS (
    SELECT *
    FROM marts.dim_customer
    WHERE is_current = TRUE
),

-- Records that changed
changes AS (
    SELECT s.*
    FROM source s
    LEFT JOIN current_dim c ON s.customer_id = c.customer_id
    WHERE c.customer_id IS NULL           -- New customer
       OR c._hash != s.row_hash           -- Changed attributes
)

-- Step 2: Expire old records
UPDATE marts.dim_customer
SET
    expiration_date = CURRENT_DATE - 1,
    is_current = FALSE
WHERE customer_id IN (SELECT customer_id FROM changes WHERE customer_id IS NOT NULL)
  AND is_current = TRUE;

-- Step 3: Insert new versions
INSERT INTO marts.dim_customer (
    customer_id, customer_name, email, segment, region,
    effective_date, expiration_date, is_current, _version, _hash
)
SELECT
    customer_id, customer_name, email, segment, region,
    CURRENT_DATE,
    '9999-12-31'::date,
    TRUE,
    coalesce((
        SELECT max(_version) + 1
        FROM marts.dim_customer d
        WHERE d.customer_id = changes.customer_id
    ), 1),
    row_hash
FROM changes;
```

### SCD Type 2 Query Patterns

```sql
-- Get current customer record
SELECT * FROM dim_customer
WHERE customer_id = 'C001' AND is_current = TRUE;

-- Get customer as of a specific date
SELECT * FROM dim_customer
WHERE customer_id = 'C001'
  AND effective_date <= '2024-06-15'
  AND expiration_date > '2024-06-15';

-- Join fact table to "as-of" dimension (point-in-time correct)
SELECT
    f.order_date,
    d.customer_name,
    d.segment,
    f.order_amount
FROM fct_orders f
JOIN dim_customer d
    ON f.customer_sk = d.customer_sk;
-- customer_sk in fact table points to the SPECIFIC version
-- that was current at the time of the order
```

## Data Quality Patterns

### dbt Data Quality Framework

```yaml
# Schema tests (generic)
models:
  - name: fct_orders
    data_tests:
      # Freshness
      - dbt_utils.recency:
          datepart: hour
          field: _loaded_at
          interval: 6

      # Volume anomaly detection
      - dbt_expectations.expect_table_row_count_to_be_between:
          min_value: "{{ var('min_daily_orders', 100) }}"
          # Alert if we get suspiciously few orders

    columns:
      - name: order_amount
        data_tests:
          # Range check
          - dbt_expectations.expect_column_values_to_be_between:
              min_value: 0
              max_value: 1000000
              row_condition: "order_status != 'cancelled'"

          # Statistical anomaly
          - dbt_expectations.expect_column_mean_to_be_between:
              min_value: 10
              max_value: 500

      - name: customer_id
        data_tests:
          # Referential integrity
          - relationships:
              to: ref('dim_customer')
              field: customer_id
              where: "is_current = true"

          # Cardinality check
          - dbt_expectations.expect_column_proportion_of_unique_values_to_be_between:
              min_value: 0.05  # At least 5% unique
```

### Data Contract Pattern

```yaml
# data_contracts/orders_contract.yml
contract:
  name: orders
  version: 2
  owner: data-engineering@company.com
  description: "Order events from the commerce platform"

  schema:
    - name: order_id
      type: string
      required: true
      unique: true
      pattern: "^ORD-[A-Z0-9]{8}$"

    - name: customer_id
      type: string
      required: true
      references: customers.customer_id

    - name: order_amount
      type: decimal(12,2)
      required: true
      min: 0
      max: 1000000

    - name: order_date
      type: date
      required: true
      max_staleness: 24h

    - name: status
      type: string
      required: true
      allowed_values: [created, paid, shipped, delivered, cancelled, refunded]

  sla:
    freshness: 1h
    completeness: 99.5%
    volume:
      min_daily: 100
      max_daily: 1000000

  breaking_changes:
    notification: slack:#data-contracts
    approval_required: true
```

## Performance Optimization

### Query Optimization Checklist

```sql
-- 1. Filter early (predicate pushdown)
-- BAD
SELECT * FROM (
    SELECT *, ROW_NUMBER() OVER (ORDER BY created_at) as rn
    FROM huge_table
) WHERE rn <= 100 AND date = '2024-01-01';

-- GOOD (filter pushed down before window function)
SELECT * FROM (
    SELECT *, ROW_NUMBER() OVER (ORDER BY created_at) as rn
    FROM huge_table
    WHERE date = '2024-01-01'  -- Filter first
) WHERE rn <= 100;


-- 2. Avoid SELECT * in production queries
-- BAD
SELECT * FROM fct_orders;

-- GOOD (columnar storage only reads needed columns)
SELECT order_id, customer_id, amount FROM fct_orders;


-- 3. Use approximate functions for large aggregations
-- BAD (exact count, full scan)
SELECT COUNT(DISTINCT customer_id) FROM fct_orders;

-- GOOD (approximate, 97%+ accurate, much faster)
SELECT APPROX_COUNT_DISTINCT(customer_id) FROM fct_orders;
-- Snowflake: APPROX_COUNT_DISTINCT
-- BigQuery: APPROX_COUNT_DISTINCT
-- Redshift: APPROXIMATE COUNT(DISTINCT ...)


-- 4. Materialize common CTEs that are used multiple times
-- BAD (CTE evaluated twice)
WITH monthly AS (
    SELECT date_trunc('month', order_date) AS month, SUM(amount) AS revenue
    FROM fct_orders GROUP BY 1
)
SELECT
    a.month,
    a.revenue,
    a.revenue - b.revenue AS mom_change
FROM monthly a
LEFT JOIN monthly b ON a.month = b.month + INTERVAL '1 month';

-- GOOD (use window function instead)
SELECT
    month,
    revenue,
    revenue - LAG(revenue) OVER (ORDER BY month) AS mom_change
FROM (
    SELECT date_trunc('month', order_date) AS month, SUM(amount) AS revenue
    FROM fct_orders GROUP BY 1
);
```

### Warehouse-Specific Optimization

#### Snowflake

```sql
-- Use result cache (automatic for identical queries within 24h)
-- Use query profile to identify bottlenecks
-- Right-size warehouses: XS for dev, M for dashboards, XL for heavy transforms

-- Transient tables for staging (no Time Travel overhead)
CREATE TRANSIENT TABLE staging.stg_temp_orders (...);

-- Zero-copy cloning for dev/test
CREATE DATABASE analytics_dev CLONE analytics_prod;
```

#### BigQuery

```sql
-- Use partitioning + clustering (reduces bytes scanned = lower cost)
-- Avoid SELECT * (columnar billing)
-- Use BI Engine for dashboard acceleration

-- Preview query cost before running
SELECT * FROM `project.dataset.table`
-- Check "This query will process X bytes" in the UI

-- Slots: on-demand (per-query pricing) vs flat-rate (reserved capacity)
```

#### Redshift

```sql
-- Run ANALYZE after bulk loads
ANALYZE marts.fct_orders;

-- Run VACUUM to reclaim space after deletes/updates
VACUUM FULL marts.fct_orders TO 95 PERCENT;

-- Check query plan
EXPLAIN SELECT ... FROM fct_orders WHERE ...;

-- Identify table design issues
SELECT * FROM svv_table_info WHERE "table" = 'fct_orders';
```

## Common Anti-Patterns

### 1. The "God Table"

```sql
-- BAD: one massive table with everything
CREATE TABLE analytics.everything AS
SELECT
    o.*, c.*, p.*, s.*, pay.*,
    -- 200 columns from 5 tables joined together
FROM orders o
JOIN customers c ON ...
JOIN products p ON ...
JOIN stores s ON ...
JOIN payments pay ON ...;

-- GOOD: normalized star schema with focused fact + dim tables
-- Join at query time or create specific OBTs for specific use cases
```

### 2. Snapshot Abuse

```sql
-- BAD: daily snapshot of entire table (stores N * days rows)
INSERT INTO dim_product_daily
SELECT *, CURRENT_DATE AS snapshot_date
FROM source_products;
-- 1M products * 365 days = 365M rows/year

-- GOOD: SCD Type 2 (only stores changes)
-- 1M products * ~5% change rate * 365 = ~18M rows/year
```

### 3. Missing Grain Documentation

```sql
-- BAD: no grain statement, ambiguous joins
CREATE TABLE fct_orders AS
SELECT ...;  -- Is this per order? Per line item? Per shipment?

-- GOOD: grain explicitly stated
-- fct_order_lines
-- Grain: one row per order line item (order_id + product_id)
-- A single order with 3 products = 3 rows
```

### 4. Hard-Coded Date Filters

```sql
-- BAD: hard-coded dates in models
SELECT * FROM source WHERE date >= '2024-01-01';

-- GOOD: parameterized
SELECT * FROM source
WHERE date >= {{ var('start_date', '2020-01-01') }};

-- Or incremental
{% if is_incremental() %}
WHERE updated_at > (SELECT MAX(updated_at) FROM {{ this }})
{% endif %}
```

## Data Governance

### Column-Level Lineage

```yaml
# Document column transformations
models:
  - name: fct_orders
    columns:
      - name: net_amount
        description: |
          Net order amount after discounts.
          Calculation: gross_amount - discount_amount
          Source: stg_shopify__orders.total_price - stg_shopify__orders.total_discounts
        meta:
          pii: false
          source_columns:
            - stg_shopify__orders.total_price
            - stg_shopify__orders.total_discounts

      - name: customer_email
        description: "Customer email, lowercased"
        meta:
          pii: true
          classification: email
          masking_policy: hash_email
```

### Access Control Pattern

```sql
-- Role-based access control (Snowflake example)
-- Create role hierarchy
CREATE ROLE analyst;
CREATE ROLE data_engineer;
CREATE ROLE data_admin;

GRANT ROLE analyst TO ROLE data_engineer;
GRANT ROLE data_engineer TO ROLE data_admin;

-- Grant access by layer
GRANT USAGE ON SCHEMA marts TO ROLE analyst;
GRANT SELECT ON ALL TABLES IN SCHEMA marts TO ROLE analyst;

GRANT USAGE ON SCHEMA staging TO ROLE data_engineer;
GRANT ALL ON ALL TABLES IN SCHEMA staging TO ROLE data_engineer;

-- Row-level security (for multi-tenant)
CREATE ROW ACCESS POLICY region_policy AS (region VARCHAR)
RETURNS BOOLEAN ->
    CURRENT_ROLE() IN ('data_admin') OR
    region = CURRENT_SESSION()::variant:user_region;

ALTER TABLE marts.fct_orders ADD ROW ACCESS POLICY region_policy ON (region);
```
