---
name: etl-pipeline-builder
description: >
  Build production-grade ETL/ELT pipelines with Apache Airflow, dbt, and data quality frameworks.
  Generates idempotent, testable, observable data pipelines following modern data engineering best practices.
  Triggers: "build ETL pipeline", "create Airflow DAG", "set up dbt project", "data pipeline",
  "build ELT", "data ingestion", "orchestrate data workflow", "create data pipeline".
  NOT for: real-time streaming (use stream-processing), ML model training, application backend APIs.
version: 1.0.0
argument-hint: "[pipeline-description]"
allowed-tools: Read, Grep, Glob, Bash, Write, Edit
model: sonnet
---

# ETL Pipeline Builder

Build production-grade ETL/ELT data pipelines with Apache Airflow orchestration, dbt transformations, and comprehensive data quality validation.

## Core Principles

Every pipeline you build must be:
- **Idempotent** — running the same pipeline twice produces the same result
- **Observable** — every stage emits metrics, logs, and alerts
- **Testable** — unit tests for transforms, integration tests for pipelines
- **Recoverable** — failures are isolated, retries are automatic, backfills are simple

## Pipeline Architecture Decision Framework

```
Source → Is it batch or streaming?
  │
  ├─ Batch (hourly/daily/weekly):
  │   ├─ Raw data < 10GB/run → Python operators + pandas
  │   ├─ Raw data 10GB-1TB → Spark operators
  │   └─ Raw data > 1TB → Spark on dedicated cluster
  │
  ├─ Micro-batch (minutes):
  │   └─ Spark Structured Streaming
  │
  └─ Real-time (seconds):
      └─ Use stream-processing skill instead
```

## Step 1: Assess the Pipeline Requirements

Before writing any code, answer these questions:

```markdown
## Pipeline Specification

### Source Systems
- [ ] What are the source systems? (database, API, files, events)
- [ ] What's the data volume per run? (rows, GB)
- [ ] What's the extraction pattern? (full, incremental, CDC)
- [ ] What's the SLA? (freshness requirement)

### Transformations
- [ ] What business logic applies? (dedup, joins, aggregations)
- [ ] Are there slowly changing dimensions?
- [ ] What's the grain of the output? (one row = what?)

### Destination
- [ ] Where does data land? (warehouse, lake, API)
- [ ] What's the write pattern? (append, upsert, replace)
- [ ] Who consumes it? (dashboards, ML, applications)

### Operations
- [ ] What's the schedule? (cron expression)
- [ ] What are the dependencies? (upstream pipelines)
- [ ] What alerts are needed? (failure, SLA breach, data quality)
```

## Step 2: Airflow DAG Patterns

### Modern TaskFlow API DAG

```python
"""
Pipeline: {source} → {destination}
Schedule: {cron}
Owner: {team}
"""
from __future__ import annotations

import logging
from datetime import datetime, timedelta

from airflow.decorators import dag, task
from airflow.models import Variable
from airflow.providers.common.sql.operators.sql import SQLExecuteQueryOperator
from airflow.utils.trigger_rule import TriggerRule

logger = logging.getLogger(__name__)

default_args = {
    "owner": "data-engineering",
    "depends_on_past": False,
    "email_on_failure": True,
    "email_on_retry": False,
    "retries": 3,
    "retry_delay": timedelta(minutes=5),
    "retry_exponential_backoff": True,
    "max_retry_delay": timedelta(minutes=30),
    "execution_timeout": timedelta(hours=2),
}


@dag(
    dag_id="pipeline_name",
    default_args=default_args,
    description="What this pipeline does",
    schedule="0 6 * * *",  # Daily at 6 AM UTC
    start_date=datetime(2024, 1, 1),
    catchup=False,
    tags=["production", "domain-name"],
    max_active_runs=1,
    doc_md=__doc__,
)
def pipeline_name():
    @task()
    def extract(execution_date=None, **context):
        """Extract data from source system."""
        # Use execution_date for idempotent extraction windows
        start = execution_date
        end = execution_date + timedelta(days=1)

        logger.info(f"Extracting data for {start} to {end}")

        # Your extraction logic here
        records = extract_from_source(start, end)

        # Push metadata for downstream monitoring
        context["ti"].xcom_push(key="record_count", value=len(records))
        return records

    @task()
    def validate_source(data):
        """Validate extracted data before transformation."""
        if not data:
            raise ValueError("No records extracted — check source system")

        # Schema validation
        required_fields = ["id", "created_at", "amount"]
        for record in data[:100]:  # Sample validation
            missing = [f for f in required_fields if f not in record]
            if missing:
                raise ValueError(f"Missing fields: {missing}")

        logger.info(f"Validated {len(data)} records")
        return data

    @task()
    def transform(data):
        """Apply business logic transformations."""
        transformed = []
        for record in data:
            transformed.append({
                **record,
                "amount_usd": convert_currency(record["amount"], record["currency"]),
                "fiscal_quarter": get_fiscal_quarter(record["created_at"]),
                "processed_at": datetime.utcnow().isoformat(),
            })
        return transformed

    @task()
    def load(data, execution_date=None):
        """Load transformed data to destination."""
        # Idempotent load: delete-then-insert for the partition
        partition_date = execution_date.strftime("%Y-%m-%d")

        delete_partition(table="target_table", partition_key="date", value=partition_date)
        insert_records(table="target_table", records=data)

        logger.info(f"Loaded {len(data)} records for {partition_date}")

    @task()
    def validate_output(execution_date=None):
        """Post-load data quality checks."""
        partition_date = execution_date.strftime("%Y-%m-%d")

        checks = [
            row_count_check("target_table", partition_date, min_rows=100),
            null_check("target_table", ["id", "amount_usd"], partition_date),
            freshness_check("target_table", max_age_hours=26),
        ]

        failures = [c for c in checks if not c["passed"]]
        if failures:
            raise ValueError(f"Data quality failures: {failures}")

    @task(trigger_rule=TriggerRule.ALL_DONE)
    def notify(execution_date=None, **context):
        """Send pipeline completion notification."""
        record_count = context["ti"].xcom_pull(
            task_ids="extract", key="record_count"
        )
        send_notification(
            channel="#data-pipeline-alerts",
            message=f"Pipeline complete: {record_count} records for {execution_date}",
        )

    # Define the DAG flow
    raw = extract()
    validated = validate_source(raw)
    transformed = transform(validated)
    load(transformed) >> validate_output() >> notify()


pipeline_name()
```

### Dynamic DAG Factory

For multiple similar pipelines (e.g., one per source table):

```python
"""Generate DAGs dynamically from a configuration table."""
import yaml
from pathlib import Path
from airflow.decorators import dag, task

# Load pipeline configs
config_path = Path(__file__).parent / "configs"
configs = []
for f in config_path.glob("*.yaml"):
    with open(f) as fh:
        configs.append(yaml.safe_load(fh))


def create_pipeline(config):
    @dag(
        dag_id=f"ingest_{config['source']}_{config['table']}",
        schedule=config.get("schedule", "@daily"),
        start_date=datetime(2024, 1, 1),
        catchup=False,
        tags=["auto-generated", config["domain"]],
    )
    def generated_pipeline():
        @task()
        def extract(**context):
            return extract_table(
                connection=config["connection"],
                table=config["table"],
                incremental_key=config.get("incremental_key"),
                execution_date=context["execution_date"],
            )

        @task()
        def load(data):
            load_to_warehouse(
                data=data,
                schema=config["target_schema"],
                table=config["target_table"],
                write_mode=config.get("write_mode", "append"),
            )

        load(extract())

    return generated_pipeline()


# Generate all DAGs
for config in configs:
    globals()[f"ingest_{config['source']}_{config['table']}"] = create_pipeline(config)
```

Pipeline config YAML:
```yaml
# configs/postgres_users.yaml
source: postgres
table: users
connection: postgres_main
target_schema: raw
target_table: users
incremental_key: updated_at
schedule: "*/30 * * * *"
domain: identity
```

### Sensor-Based Dependencies

```python
from airflow.sensors.external_task import ExternalTaskSensor

@dag(...)
def downstream_pipeline():
    wait_for_upstream = ExternalTaskSensor(
        task_id="wait_for_orders_pipeline",
        external_dag_id="ingest_orders",
        external_task_id="validate_output",
        timeout=7200,
        poke_interval=300,
        mode="reschedule",  # Frees up worker slot while waiting
    )

    @task()
    def build_report():
        ...

    wait_for_upstream >> build_report()
```

## Step 3: dbt Project Structure

### Project Layout

```
dbt_project/
├── dbt_project.yml
├── profiles.yml          # Connection config (gitignored)
├── packages.yml          # dbt packages (dbt-utils, etc.)
├── models/
│   ├── staging/          # 1:1 with source tables, light cleaning
│   │   ├── stg_orders.sql
│   │   ├── stg_customers.sql
│   │   └── _staging.yml  # Tests and docs for staging models
│   ├── intermediate/     # Business logic building blocks
│   │   ├── int_orders_enriched.sql
│   │   └── _intermediate.yml
│   └── marts/           # Final consumption layer
│       ├── finance/
│       │   ├── fct_revenue.sql
│       │   └── dim_customers.sql
│       └── _marts.yml
├── tests/
│   ├── generic/         # Reusable test macros
│   └── singular/        # One-off SQL tests
├── macros/
│   └── generate_schema_name.sql
├── seeds/               # Static reference data (CSV)
│   └── country_codes.csv
└── snapshots/           # SCD Type 2 tracking
    └── snap_customers.sql
```

### Staging Model Pattern

```sql
-- models/staging/stg_orders.sql
with source as (
    select * from {{ source('ecommerce', 'orders') }}
),

renamed as (
    select
        -- Primary key
        id as order_id,

        -- Foreign keys
        user_id as customer_id,
        product_id,

        -- Dimensions
        lower(trim(status)) as order_status,
        lower(trim(channel)) as order_channel,

        -- Measures
        amount_cents / 100.0 as order_amount,
        tax_cents / 100.0 as tax_amount,
        (amount_cents + tax_cents) / 100.0 as total_amount,

        -- Dates
        cast(created_at as timestamp) as ordered_at,
        cast(shipped_at as timestamp) as shipped_at,
        cast(delivered_at as timestamp) as delivered_at,

        -- Metadata
        _loaded_at as _extracted_at,
        current_timestamp() as _transformed_at

    from source
    where not _is_deleted  -- Soft delete filter
)

select * from renamed
```

### Intermediate Model (Business Logic)

```sql
-- models/intermediate/int_orders_enriched.sql
with orders as (
    select * from {{ ref('stg_orders') }}
),

customers as (
    select * from {{ ref('stg_customers') }}
),

products as (
    select * from {{ ref('stg_products') }}
),

enriched as (
    select
        orders.order_id,
        orders.ordered_at,
        orders.order_status,
        orders.order_amount,
        orders.total_amount,

        -- Customer enrichment
        customers.customer_id,
        customers.customer_name,
        customers.customer_segment,
        customers.first_order_at,
        case
            when orders.ordered_at = customers.first_order_at then 'new'
            else 'returning'
        end as customer_type,

        -- Product enrichment
        products.product_name,
        products.product_category,
        products.product_subcategory,

        -- Derived metrics
        datediff('day', customers.first_order_at, orders.ordered_at) as customer_tenure_days,
        row_number() over (
            partition by orders.customer_id
            order by orders.ordered_at
        ) as customer_order_number

    from orders
    left join customers using (customer_id)
    left join products using (product_id)
)

select * from enriched
```

### Mart Model (Fact Table)

```sql
-- models/marts/finance/fct_revenue.sql
{{
    config(
        materialized='incremental',
        unique_key='order_id',
        incremental_strategy='merge',
        on_schema_change='append_new_columns',
        cluster_by=['ordered_date', 'product_category']
    )
}}

with orders as (
    select * from {{ ref('int_orders_enriched') }}
    {% if is_incremental() %}
    where ordered_at > (select max(ordered_at) from {{ this }})
    {% endif %}
),

final as (
    select
        -- Keys
        order_id,
        customer_id,

        -- Dimensions
        date_trunc('day', ordered_at) as ordered_date,
        date_trunc('month', ordered_at) as ordered_month,
        order_status,
        customer_type,
        customer_segment,
        product_category,

        -- Measures
        order_amount,
        tax_amount,
        total_amount,
        1 as order_count,
        case when order_status = 'returned' then total_amount else 0 end as returned_amount,

        -- Metadata
        ordered_at,
        current_timestamp() as _dbt_loaded_at

    from orders
    where order_status != 'cancelled'
)

select * from final
```

### dbt Tests and Documentation

```yaml
# models/marts/_marts.yml
version: 2

models:
  - name: fct_revenue
    description: >
      Fact table containing all completed revenue transactions.
      Grain: one row per order.
      Updated incrementally — new/changed orders merged daily.
    columns:
      - name: order_id
        description: Primary key — unique order identifier
        data_tests:
          - unique
          - not_null

      - name: customer_id
        description: FK to dim_customers
        data_tests:
          - not_null
          - relationships:
              to: ref('dim_customers')
              field: customer_id

      - name: total_amount
        description: Order total including tax (USD)
        data_tests:
          - not_null
          - dbt_utils.accepted_range:
              min_value: 0
              max_value: 100000

      - name: ordered_date
        description: Date the order was placed
        data_tests:
          - not_null
          - dbt_utils.not_constant
```

### SCD Type 2 Snapshot

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
    plan_tier,
    billing_address_country,
    updated_at
from {{ source('app', 'customers') }}

{% endsnapshot %}
```

## Step 4: Data Quality Framework

### Great Expectations Integration

```python
"""Data quality validation suite."""
from great_expectations.core import ExpectationSuite, ExpectationConfiguration


def build_orders_suite() -> ExpectationSuite:
    suite = ExpectationSuite(expectation_suite_name="orders_quality")

    # Completeness
    suite.add_expectation(ExpectationConfiguration(
        expectation_type="expect_table_row_count_to_be_between",
        kwargs={"min_value": 100, "max_value": 1000000},
    ))

    # Uniqueness
    suite.add_expectation(ExpectationConfiguration(
        expectation_type="expect_column_values_to_be_unique",
        kwargs={"column": "order_id"},
    ))

    # Validity
    suite.add_expectation(ExpectationConfiguration(
        expectation_type="expect_column_values_to_be_in_set",
        kwargs={
            "column": "status",
            "value_set": ["pending", "confirmed", "shipped", "delivered", "returned", "cancelled"],
        },
    ))

    # Freshness
    suite.add_expectation(ExpectationConfiguration(
        expectation_type="expect_column_max_to_be_between",
        kwargs={
            "column": "created_at",
            "min_value": "{{ (execution_date - timedelta(hours=26)).isoformat() }}",
        },
    ))

    # Referential integrity
    suite.add_expectation(ExpectationConfiguration(
        expectation_type="expect_column_values_to_not_be_null",
        kwargs={"column": "customer_id"},
    ))

    # Statistical — amount distribution
    suite.add_expectation(ExpectationConfiguration(
        expectation_type="expect_column_mean_to_be_between",
        kwargs={"column": "amount", "min_value": 10, "max_value": 500},
    ))

    return suite
```

### Lightweight SQL-Based Quality Checks

```sql
-- tests/singular/test_revenue_reconciliation.sql
-- Revenue in fact table should match source within 1%
with fact_total as (
    select sum(total_amount) as fact_revenue
    from {{ ref('fct_revenue') }}
    where ordered_date = current_date - interval '1 day'
),

source_total as (
    select sum(amount_cents + tax_cents) / 100.0 as source_revenue
    from {{ source('ecommerce', 'orders') }}
    where date(created_at) = current_date - interval '1 day'
    and status not in ('cancelled')
)

select
    fact_revenue,
    source_revenue,
    abs(fact_revenue - source_revenue) / nullif(source_revenue, 0) as variance_pct
from fact_total, source_total
where abs(fact_revenue - source_revenue) / nullif(source_revenue, 0) > 0.01
```

## Step 5: Backfill and Recovery

### Airflow Backfill Pattern

```python
@task()
def extract_with_backfill(execution_date=None, **context):
    """Idempotent extraction supporting backfills."""
    dag_run = context["dag_run"]

    # Check if this is a backfill run
    if dag_run.conf and dag_run.conf.get("backfill_start"):
        start = parse(dag_run.conf["backfill_start"])
        end = parse(dag_run.conf["backfill_end"])
        logger.info(f"BACKFILL mode: {start} to {end}")
    else:
        start = execution_date
        end = execution_date + timedelta(days=1)

    return extract_data(start, end)
```

### Trigger backfill via CLI:
```bash
airflow dags trigger pipeline_name \
  --conf '{"backfill_start": "2024-01-01", "backfill_end": "2024-03-01"}'
```

## Step 6: Pipeline Monitoring

### Airflow Callbacks for Alerting

```python
from airflow.providers.slack.hooks.slack_webhook import SlackWebhookHook


def on_failure_callback(context):
    """Alert on task failure."""
    task_instance = context["task_instance"]
    exception = context.get("exception", "Unknown error")

    SlackWebhookHook(slack_webhook_conn_id="slack_data_alerts").send(
        text=f":red_circle: Pipeline Failed\n"
             f"*DAG*: {task_instance.dag_id}\n"
             f"*Task*: {task_instance.task_id}\n"
             f"*Error*: {str(exception)[:500]}\n"
             f"*Log*: {task_instance.log_url}",
    )


def sla_miss_callback(dag, task_list, blocking_task_list, slas, blocking_tis):
    """Alert when SLA is breached."""
    SlackWebhookHook(slack_webhook_conn_id="slack_data_alerts").send(
        text=f":warning: SLA Breach\n*DAG*: {dag.dag_id}\n*Tasks*: {task_list}",
    )


@dag(
    on_failure_callback=on_failure_callback,
    sla_miss_callback=sla_miss_callback,
    ...
)
def pipeline_name():
    ...
```

## Checklist Before Completing

- [ ] All tasks are idempotent (rerunnable without side effects)
- [ ] Incremental models have `is_incremental()` guards
- [ ] dbt tests cover primary keys (unique + not_null) and foreign keys (relationships)
- [ ] Airflow DAG has `max_active_runs=1` to prevent overlapping runs
- [ ] Retry policy is configured with exponential backoff
- [ ] Alerting is set up for failures and SLA breaches
- [ ] Data quality checks run post-load
- [ ] Backfill mechanism is documented and tested
- [ ] Pipeline is tagged and documented in Airflow UI
