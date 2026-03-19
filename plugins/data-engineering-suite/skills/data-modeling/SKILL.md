---
name: data-modeling
description: >
  Design dimensional data models, star schemas, snowflake schemas, and slowly changing dimensions
  for data warehouses. Covers Kimball methodology, data vault, and modern analytics patterns.
  Triggers: "design data model", "star schema", "snowflake schema", "dimensional model",
  "data warehouse schema", "fact table", "dimension table", "SCD", "slowly changing dimension".
  NOT for: OLTP application database design, NoSQL schema design, API data models.
version: 1.0.0
argument-hint: "[domain or business process]"
allowed-tools: Read, Grep, Glob, Bash, Write, Edit
model: sonnet
---

# Data Modeling

Design production-grade dimensional models for analytics and data warehousing using Kimball methodology, data vault patterns, and modern warehouse optimization techniques.

## Modeling Methodology Selection

```
What's the primary use case?
  │
  ├─ Business reporting & dashboards
  │   └─ Kimball Dimensional Model (star/snowflake schema)
  │       Best for: BI tools, ad-hoc queries, business users
  │
  ├─ Highly regulated / audit-heavy
  │   └─ Data Vault 2.0
  │       Best for: Banking, healthcare, audit trails, multiple source integration
  │
  ├─ Flexible analytics with complex relationships
  │   └─ One Big Table (OBT) / Wide Denormalized
  │       Best for: Small teams, simple queries, columnar warehouses
  │
  └─ Real-time + historical analytics
      └─ Activity Schema / Event-Based
          Best for: Product analytics, user behavior, event streams
```

## Kimball Dimensional Modeling

### The Four-Step Design Process

1. **Select the business process** — What activity generates the data? (orders, page views, claims)
2. **Declare the grain** — What does one row represent? (one order line, one click, one daily snapshot)
3. **Identify the dimensions** — Who, what, where, when, why, how? (customer, product, store, date)
4. **Identify the facts** — What are we measuring? (quantity, amount, duration, count)

### Star Schema Template

```sql
-- ============================================
-- FACT TABLE: One row per [grain description]
-- ============================================
CREATE TABLE fct_orders (
    -- Surrogate key
    order_key           BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

    -- Degenerate dimensions (no lookup table needed)
    order_number        VARCHAR(50) NOT NULL,

    -- Foreign keys to dimension tables
    customer_key        BIGINT NOT NULL REFERENCES dim_customer(customer_key),
    product_key         BIGINT NOT NULL REFERENCES dim_product(product_key),
    date_key            INT NOT NULL REFERENCES dim_date(date_key),
    store_key           BIGINT NOT NULL REFERENCES dim_store(store_key),
    promotion_key       BIGINT NOT NULL REFERENCES dim_promotion(promotion_key),

    -- Additive facts (can be summed across ALL dimensions)
    quantity            INT NOT NULL,
    unit_price          DECIMAL(12,2) NOT NULL,
    discount_amount     DECIMAL(12,2) NOT NULL DEFAULT 0,
    net_amount          DECIMAL(12,2) NOT NULL,
    tax_amount          DECIMAL(12,2) NOT NULL,
    total_amount        DECIMAL(12,2) NOT NULL,

    -- Semi-additive facts (can be summed across some dimensions, not time)
    -- Example: account_balance — sum across accounts, not across dates

    -- Non-additive facts (ratios, percentages — never sum directly)
    gross_margin_pct    DECIMAL(5,4),
    discount_pct        DECIMAL(5,4),

    -- Metadata
    _source_system      VARCHAR(50),
    _loaded_at          TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Partition by date for query performance
-- ALTER TABLE fct_orders PARTITION BY RANGE (date_key);


-- ============================================
-- DIMENSION: Customer (SCD Type 2)
-- ============================================
CREATE TABLE dim_customer (
    customer_key        BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    customer_id         VARCHAR(50) NOT NULL,  -- Natural/business key

    -- Attributes
    customer_name       VARCHAR(200),
    email               VARCHAR(200),
    phone               VARCHAR(50),

    -- Classification
    customer_segment    VARCHAR(50),   -- Enterprise, SMB, Consumer
    industry            VARCHAR(100),
    acquisition_channel VARCHAR(50),   -- Organic, Paid, Referral

    -- Geography
    city                VARCHAR(100),
    state               VARCHAR(100),
    country             VARCHAR(100),
    postal_code         VARCHAR(20),

    -- SCD Type 2 tracking
    effective_from      TIMESTAMP NOT NULL,
    effective_to        TIMESTAMP DEFAULT '9999-12-31',
    is_current          BOOLEAN DEFAULT TRUE,

    -- Metadata
    _source_system      VARCHAR(50),
    _loaded_at          TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_dim_customer_natural ON dim_customer(customer_id, is_current);


-- ============================================
-- DIMENSION: Date (Calendar)
-- ============================================
CREATE TABLE dim_date (
    date_key            INT PRIMARY KEY,          -- YYYYMMDD format
    full_date           DATE NOT NULL UNIQUE,

    -- Calendar attributes
    day_of_week         INT,                      -- 1=Monday, 7=Sunday
    day_of_week_name    VARCHAR(10),              -- Monday, Tuesday...
    day_of_month        INT,
    day_of_year         INT,
    week_of_year        INT,
    month_number        INT,
    month_name          VARCHAR(10),
    quarter_number      INT,
    quarter_name        VARCHAR(5),               -- Q1, Q2, Q3, Q4
    year_number         INT,
    year_month          VARCHAR(7),               -- 2024-01
    year_quarter        VARCHAR(7),               -- 2024-Q1

    -- Fiscal calendar (customize per business)
    fiscal_year         INT,
    fiscal_quarter      INT,
    fiscal_month        INT,

    -- Flags
    is_weekend          BOOLEAN,
    is_holiday          BOOLEAN,
    holiday_name        VARCHAR(100),
    is_business_day     BOOLEAN,

    -- Relative flags (update daily via scheduled job)
    is_current_day      BOOLEAN DEFAULT FALSE,
    is_current_week     BOOLEAN DEFAULT FALSE,
    is_current_month    BOOLEAN DEFAULT FALSE,
    is_current_quarter  BOOLEAN DEFAULT FALSE,
    is_current_year     BOOLEAN DEFAULT FALSE,
    is_prior_year_same_day BOOLEAN DEFAULT FALSE
);


-- ============================================
-- DIMENSION: Product (SCD Type 1 — overwrite)
-- ============================================
CREATE TABLE dim_product (
    product_key         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    product_id          VARCHAR(50) NOT NULL UNIQUE,  -- Natural key

    product_name        VARCHAR(200) NOT NULL,
    product_description TEXT,

    -- Hierarchy
    category            VARCHAR(100),
    subcategory         VARCHAR(100),
    brand               VARCHAR(100),

    -- Attributes
    unit_cost           DECIMAL(12,2),
    unit_list_price     DECIMAL(12,2),
    weight_kg           DECIMAL(8,3),
    is_active           BOOLEAN DEFAULT TRUE,

    -- Metadata
    _source_system      VARCHAR(50),
    _loaded_at          TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### Date Dimension Generator

```python
"""Generate dim_date rows for a given range."""
import pandas as pd
from datetime import date


def generate_date_dimension(start: date, end: date, holidays: dict = None) -> list[dict]:
    """Generate date dimension records.

    Args:
        start: First date to generate
        end: Last date to generate
        holidays: Dict of {date: holiday_name}
    """
    holidays = holidays or {}
    dates = pd.date_range(start, end)
    records = []

    for d in dates:
        dt = d.to_pydatetime().date()
        records.append({
            "date_key": int(dt.strftime("%Y%m%d")),
            "full_date": dt,
            "day_of_week": dt.isoweekday(),
            "day_of_week_name": dt.strftime("%A"),
            "day_of_month": dt.day,
            "day_of_year": dt.timetuple().tm_yday,
            "week_of_year": dt.isocalendar()[1],
            "month_number": dt.month,
            "month_name": dt.strftime("%B"),
            "quarter_number": (dt.month - 1) // 3 + 1,
            "quarter_name": f"Q{(dt.month - 1) // 3 + 1}",
            "year_number": dt.year,
            "year_month": dt.strftime("%Y-%m"),
            "year_quarter": f"{dt.year}-Q{(dt.month - 1) // 3 + 1}",
            "is_weekend": dt.isoweekday() >= 6,
            "is_holiday": dt in holidays,
            "holiday_name": holidays.get(dt),
            "is_business_day": dt.isoweekday() < 6 and dt not in holidays,
        })

    return records
```

## Slowly Changing Dimensions (SCDs)

### SCD Type Decision Matrix

| Type | Behavior | When to Use | Trade-off |
|------|----------|-------------|-----------|
| Type 0 | Never changes | Reference data (country codes, currencies) | No history |
| Type 1 | Overwrite | Corrections, typo fixes | Loses old value |
| Type 2 | New row + versioning | Track full history (address changes, price changes) | Table grows |
| Type 3 | Add column | Track current + previous only | Limited history |
| Type 6 | Hybrid (1+2+3) | Need current, previous, AND full history | Complex |

### SCD Type 2 — dbt Implementation

```sql
-- snapshots/snap_customers.sql
{% snapshot snap_customers %}
{{
    config(
        target_schema='snapshots',
        unique_key='customer_id',
        strategy='timestamp',
        updated_at='updated_at',
        invalidate_hard_deletes=True,
    )
}}

select
    customer_id,
    customer_name,
    email,
    customer_segment,
    city,
    state,
    country,
    updated_at
from {{ source('app', 'customers') }}

{% endsnapshot %}
```

```sql
-- models/marts/dim_customer.sql
-- Build dimension from snapshot with proper SCD2 columns
with snapshot as (
    select * from {{ ref('snap_customers') }}
),

final as (
    select
        {{ dbt_utils.generate_surrogate_key(['customer_id', 'dbt_valid_from']) }} as customer_key,
        customer_id,
        customer_name,
        email,
        customer_segment,
        city,
        state,
        country,

        -- SCD Type 2 columns
        dbt_valid_from as effective_from,
        coalesce(dbt_valid_to, '9999-12-31'::timestamp) as effective_to,
        dbt_valid_to is null as is_current

    from snapshot
)

select * from final
```

### SCD Type 2 — Manual SQL Merge

```sql
-- For warehouses without dbt, manual MERGE pattern:
MERGE INTO dim_customer AS target
USING (
    SELECT * FROM staging_customers
) AS source
ON target.customer_id = source.customer_id
   AND target.is_current = TRUE

-- Update existing current record (close it out)
WHEN MATCHED AND (
    target.customer_name != source.customer_name
    OR target.customer_segment != source.customer_segment
    OR target.city != source.city
) THEN UPDATE SET
    effective_to = CURRENT_TIMESTAMP,
    is_current = FALSE

-- No changes — do nothing (handled by NOT MATCHED only inserting new)
;

-- Insert new version for changed records + brand new records
INSERT INTO dim_customer (customer_id, customer_name, customer_segment, city,
                          effective_from, effective_to, is_current)
SELECT
    s.customer_id, s.customer_name, s.customer_segment, s.city,
    CURRENT_TIMESTAMP, '9999-12-31', TRUE
FROM staging_customers s
WHERE NOT EXISTS (
    SELECT 1 FROM dim_customer d
    WHERE d.customer_id = s.customer_id
    AND d.is_current = TRUE
    AND d.customer_name = s.customer_name
    AND d.customer_segment = s.customer_segment
    AND d.city = s.city
);
```

## Snowflake Schema vs Star Schema

### When to Snowflake

```
Star Schema (default choice):
  fct_orders → dim_product (with category, subcategory inline)

Snowflake Schema (use sparingly):
  fct_orders → dim_product → dim_category → dim_subcategory

Use snowflake ONLY when:
  ✓ Dimension has 50+ attributes across 3+ hierarchies
  ✓ Hierarchy changes independently (product stays same, category restructured)
  ✓ Storage is genuinely constrained (rare in modern warehouses)

Prefer star schema because:
  ✓ Simpler queries (fewer joins)
  ✓ Better BI tool compatibility
  ✓ Easier for business users to understand
  ✓ Modern columnar warehouses handle redundancy efficiently
```

## Advanced Patterns

### Accumulating Snapshot Fact Table

Track lifecycle of a process with milestones:

```sql
CREATE TABLE fct_order_lifecycle (
    order_key           BIGINT PRIMARY KEY,
    customer_key        BIGINT REFERENCES dim_customer,
    product_key         BIGINT REFERENCES dim_product,

    -- Milestone date keys
    order_date_key      INT REFERENCES dim_date,
    payment_date_key    INT REFERENCES dim_date,
    ship_date_key       INT REFERENCES dim_date,
    delivery_date_key   INT REFERENCES dim_date,
    return_date_key     INT REFERENCES dim_date,

    -- Lag measures (days between milestones)
    days_to_payment     INT,
    days_to_ship        INT,
    days_to_deliver     INT,
    days_order_to_deliver INT,

    -- Current status
    current_status      VARCHAR(20),

    -- Measures
    order_amount        DECIMAL(12,2),

    _last_updated       TIMESTAMP
);
```

### Factless Fact Table

Track events or coverage without measures:

```sql
-- Student attendance: which students attended which classes on which days
CREATE TABLE fct_attendance (
    date_key        INT REFERENCES dim_date,
    student_key     BIGINT REFERENCES dim_student,
    class_key       BIGINT REFERENCES dim_class,
    attendance_flag BOOLEAN NOT NULL DEFAULT TRUE,  -- present/absent

    PRIMARY KEY (date_key, student_key, class_key)
);

-- Query: Which students did NOT attend class on a given day?
SELECT s.student_name, c.class_name
FROM dim_student s
CROSS JOIN dim_class c
CROSS JOIN dim_date d
LEFT JOIN fct_attendance a
    ON a.student_key = s.student_key
    AND a.class_key = c.class_key
    AND a.date_key = d.date_key
WHERE d.full_date = '2024-03-15'
  AND a.attendance_flag IS NULL;
```

### Junk Dimension

Consolidate low-cardinality flags into a single dimension:

```sql
CREATE TABLE dim_order_flags (
    order_flags_key     INT PRIMARY KEY,
    is_gift_wrapped     BOOLEAN,
    is_expedited        BOOLEAN,
    is_business_order   BOOLEAN,
    payment_method      VARCHAR(20),  -- credit_card, debit, paypal, crypto
    delivery_type       VARCHAR(20)   -- standard, express, same_day, pickup
);

-- Pre-populate all valid combinations
-- Keeps fact table narrow — one FK instead of 5 columns
```

### Bridge Table for Many-to-Many

```sql
-- One patient can have multiple diagnoses per visit
CREATE TABLE bridge_patient_diagnosis (
    patient_visit_group_key  BIGINT,  -- FK from fct_patient_visits
    diagnosis_key            BIGINT REFERENCES dim_diagnosis,
    diagnosis_rank           INT,     -- Primary=1, Secondary=2, etc.
    weight_factor            DECIMAL(5,4) DEFAULT 1.0,  -- For weighted allocation

    PRIMARY KEY (patient_visit_group_key, diagnosis_key)
);
```

## Data Vault 2.0 (When Required)

### Core Concepts

```
Hub:    Business key + metadata (load date, source)
Link:   Relationship between hubs (M:N joins)
Satellite: Descriptive attributes with history (SCD Type 2 built-in)
```

```sql
-- Hub: immutable business keys
CREATE TABLE h_customer (
    h_customer_hashkey  CHAR(32) PRIMARY KEY,  -- MD5(customer_id)
    customer_id         VARCHAR(50) NOT NULL,
    load_date           TIMESTAMP NOT NULL,
    record_source       VARCHAR(50) NOT NULL
);

-- Satellite: descriptive context with history
CREATE TABLE s_customer_details (
    h_customer_hashkey  CHAR(32) NOT NULL REFERENCES h_customer,
    load_date           TIMESTAMP NOT NULL,
    customer_name       VARCHAR(200),
    email               VARCHAR(200),
    segment             VARCHAR(50),
    hash_diff           CHAR(32) NOT NULL,  -- MD5 of all attributes
    record_source       VARCHAR(50) NOT NULL,

    PRIMARY KEY (h_customer_hashkey, load_date)
);

-- Link: relationship between business concepts
CREATE TABLE l_order (
    l_order_hashkey     CHAR(32) PRIMARY KEY,  -- MD5(order_id + customer_id + product_id)
    h_customer_hashkey  CHAR(32) REFERENCES h_customer,
    h_product_hashkey   CHAR(32) REFERENCES h_product,
    order_id            VARCHAR(50) NOT NULL,
    load_date           TIMESTAMP NOT NULL,
    record_source       VARCHAR(50) NOT NULL
);
```

## Warehouse Optimization

### Partitioning Strategy

```sql
-- Partition fact tables by date (most common access pattern)
CREATE TABLE fct_events (
    event_id        BIGINT,
    event_date      DATE NOT NULL,
    user_id         BIGINT,
    event_type      VARCHAR(50),
    properties      JSONB
) PARTITION BY RANGE (event_date);

-- Create monthly partitions
CREATE TABLE fct_events_2024_01 PARTITION OF fct_events
    FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');
CREATE TABLE fct_events_2024_02 PARTITION OF fct_events
    FOR VALUES FROM ('2024-02-01') TO ('2024-03-01');

-- Auto-create future partitions via pg_partman or cron
```

### Clustering/Sort Keys

```sql
-- BigQuery: Cluster by common filter columns
CREATE TABLE fct_revenue
PARTITION BY ordered_date
CLUSTER BY customer_segment, product_category
AS SELECT ...;

-- Redshift: Sort key on most filtered column
CREATE TABLE fct_revenue (
    ...
) DISTKEY(customer_id) SORTKEY(ordered_date);

-- Snowflake: Cluster key
ALTER TABLE fct_revenue CLUSTER BY (ordered_date, product_category);
```

## Checklist Before Completing

- [ ] Grain is clearly defined and documented for every fact table
- [ ] Every dimension has a surrogate key (not the natural/business key as PK)
- [ ] Date dimension covers the full range needed (past + future)
- [ ] SCD strategy chosen and implemented for each dimension
- [ ] Fact table columns classified as additive/semi-additive/non-additive
- [ ] Foreign keys from fact → dimension are all NOT NULL
- [ ] Junk dimensions consolidate low-cardinality flags
- [ ] Partitioning and clustering align with query patterns
- [ ] NULL handling strategy defined (unknown member row in dimensions)
