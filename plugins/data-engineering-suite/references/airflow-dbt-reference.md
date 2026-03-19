# Apache Airflow & dbt Reference

## Airflow Architecture

### Core Components

```
┌──────────────────────────────────────────────────┐
│                   Airflow Cluster                 │
│                                                   │
│  ┌──────────┐  ┌──────────────┐  ┌─────────────┐│
│  │Scheduler │  │  Metadata DB │  │  Web Server  ││
│  │          │──│ (PostgreSQL) │──│   (UI/API)   ││
│  │ Parses   │  │              │  │              ││
│  │ DAGs,    │  │ DAG state,   │  │ DAG viz,     ││
│  │ triggers │  │ task history, │  │ logs, admin  ││
│  │ tasks    │  │ connections  │  │              ││
│  └────┬─────┘  └──────────────┘  └─────────────┘│
│       │                                          │
│  ┌────┴──────────────────────────────────────┐   │
│  │              Executor                      │   │
│  │  ┌────────┐ ┌────────┐ ┌────────┐        │   │
│  │  │Worker 1│ │Worker 2│ │Worker N│        │   │
│  │  └────────┘ └────────┘ └────────┘        │   │
│  │  (Celery / Kubernetes / Local)            │   │
│  └───────────────────────────────────────────┘   │
└──────────────────────────────────────────────────┘
```

### Executor Selection

| Executor | When to Use | Scaling |
|----------|-------------|---------|
| LocalExecutor | Dev, small workloads (<20 parallel tasks) | Vertical only |
| CeleryExecutor | Medium-large, stable workloads | Horizontal (add workers) |
| KubernetesExecutor | Cloud-native, bursty workloads | Auto-scale pods |
| CeleryKubernetesExecutor | Mixed: steady + bursty | Best of both |

## DAG Authoring Best Practices

### DAG Configuration

```python
from datetime import datetime, timedelta

default_args = {
    "owner": "data-engineering",
    "depends_on_past": False,       # Don't wait for previous run success
    "email_on_failure": True,
    "email_on_retry": False,
    "retries": 3,
    "retry_delay": timedelta(minutes=5),
    "retry_exponential_backoff": True,
    "max_retry_delay": timedelta(minutes=60),
    "execution_timeout": timedelta(hours=2),
    "sla": timedelta(hours=4),      # Alert if task takes >4h
}

@dag(
    dag_id="ingest_orders",
    default_args=default_args,
    schedule="0 6 * * *",           # Cron expression
    start_date=datetime(2024, 1, 1),
    catchup=False,                  # Don't backfill missed runs
    max_active_runs=1,              # Prevent overlapping runs
    max_active_tasks=16,            # Limit concurrent tasks
    tags=["production", "orders"],
    doc_md="Ingests order data from source DB to warehouse.",
    render_template_as_native_obj=True,  # Pass Python objects, not strings
)
def my_dag():
    ...
```

### TaskFlow API Patterns

```python
from airflow.decorators import dag, task, task_group

@dag(...)
def order_pipeline():

    @task()
    def extract(source: str, **context) -> list[dict]:
        """Extract data from source system."""
        execution_date = context["data_interval_start"]
        return query_source(source, execution_date)

    @task()
    def transform(data: list[dict]) -> list[dict]:
        """Apply business logic."""
        return [enrich(record) for record in data]

    @task()
    def load(data: list[dict], target: str):
        """Load to warehouse."""
        upsert_to_warehouse(data, target)

    @task_group()
    def process_source(source: str, target: str):
        """Reusable source processing group."""
        raw = extract(source)
        clean = transform(raw)
        load(clean, target)

    # Parallel processing of multiple sources
    process_source.override(group_id="orders")("orders_db", "raw.orders")
    process_source.override(group_id="payments")("payments_api", "raw.payments")
    process_source.override(group_id="customers")("crm_db", "raw.customers")

order_pipeline()
```

### Dynamic Task Generation

```python
@dag(...)
def dynamic_pipeline():

    @task()
    def get_tables_to_sync() -> list[str]:
        """Dynamically discover tables to process."""
        return ["users", "orders", "products", "inventory"]

    @task()
    def sync_table(table_name: str):
        """Sync a single table."""
        source_data = extract_table(table_name)
        load_table(table_name, source_data)

    @task()
    def validate_all():
        """Run cross-table validation."""
        run_data_quality_checks()

    tables = get_tables_to_sync()
    synced = sync_table.expand(table_name=tables)  # Dynamic fan-out
    synced >> validate_all()

dynamic_pipeline()
```

### XCom Best Practices

```python
# XCom is for SMALL metadata only (<48KB default, configurable)
# Do NOT pass large datasets through XCom

# GOOD: Pass metadata
@task()
def extract(**context):
    path = "/data/staging/orders_20240301.parquet"
    write_to_file(data, path)
    return {"path": path, "row_count": len(data)}  # Metadata only

@task()
def load(metadata: dict):
    data = read_from_file(metadata["path"])
    logger.info(f"Loading {metadata['row_count']} rows")

# BAD: Passing actual data
@task()
def extract():
    return huge_dataframe.to_dict()  # DON'T — can overflow XCom backend
```

### Connections and Variables

```python
from airflow.hooks.base import BaseHook
from airflow.models import Variable

# Connection (use for external system credentials)
conn = BaseHook.get_connection("warehouse_postgres")
# conn.host, conn.login, conn.password, conn.schema, conn.port

# Variable (use for configuration values)
env = Variable.get("environment", default_var="dev")
config = Variable.get("pipeline_config", deserialize_json=True)

# Secret backends (production):
# - AWS Secrets Manager
# - GCP Secret Manager
# - HashiCorp Vault
# Configure in airflow.cfg: secrets_backend = ...
```

### Sensor Patterns

```python
from airflow.sensors.external_task import ExternalTaskSensor
from airflow.sensors.filesystem import FileSensor
from airflow.providers.http.sensors.http import HttpSensor

# Wait for upstream DAG
wait_upstream = ExternalTaskSensor(
    task_id="wait_for_raw_data",
    external_dag_id="ingest_raw_orders",
    external_task_id="load_complete",
    timeout=7200,              # 2 hours max wait
    poke_interval=300,         # Check every 5 min
    mode="reschedule",         # Free up worker slot while waiting
    allowed_states=["success"],
    failed_states=["failed", "skipped"],
)

# Wait for file
wait_file = FileSensor(
    task_id="wait_for_file",
    filepath="/data/incoming/daily_export_{{ ds_nodash }}.csv",
    poke_interval=60,
    timeout=3600,
    mode="reschedule",
)

# Wait for API health
wait_api = HttpSensor(
    task_id="wait_for_api",
    http_conn_id="source_api",
    endpoint="/health",
    response_check=lambda response: response.json()["status"] == "healthy",
    poke_interval=30,
    timeout=600,
)
```

## dbt Reference

### Project Configuration

```yaml
# dbt_project.yml
name: analytics
version: "1.0.0"
config-version: 2
profile: warehouse

model-paths: ["models"]
analysis-paths: ["analyses"]
test-paths: ["tests"]
seed-paths: ["seeds"]
macro-paths: ["macros"]
snapshot-paths: ["snapshots"]
asset-paths: ["assets"]

clean-targets:
  - target
  - dbt_packages

models:
  analytics:
    staging:
      +materialized: view
      +schema: staging
    intermediate:
      +materialized: ephemeral  # CTEs, not actual tables
    marts:
      +materialized: table
      +schema: analytics

vars:
  start_date: "2023-01-01"
```

### Source Definition

```yaml
# models/staging/_sources.yml
version: 2

sources:
  - name: ecommerce
    description: Main ecommerce application database
    database: raw_db
    schema: public
    loader: fivetran
    loaded_at_field: _fivetran_synced

    freshness:
      warn_after: {count: 12, period: hour}
      error_after: {count: 24, period: hour}

    tables:
      - name: orders
        description: Customer orders table
        identifier: app_orders  # Actual table name if different
        columns:
          - name: id
            description: Primary key
            data_tests:
              - unique
              - not_null

      - name: customers
        description: Customer profiles
        columns:
          - name: id
            data_tests:
              - unique
              - not_null
          - name: email
            data_tests:
              - unique
              - not_null
```

### Materialization Strategies

| Type | SQL Object | When to Use |
|------|-----------|-------------|
| `view` | CREATE VIEW | Staging models, light transforms, always-fresh data |
| `table` | CREATE TABLE | Marts consumed by BI, large datasets, complex transforms |
| `incremental` | INSERT/MERGE | Large fact tables, append-mostly data |
| `ephemeral` | CTE (no object) | Intermediate logic reused in multiple models |
| `snapshot` | SCD Type 2 table | Track historical changes to dimensions |

### Incremental Model Strategies

```sql
-- Append (fastest, no dedup)
{{ config(materialized='incremental', incremental_strategy='append') }}

-- Merge/Upsert (handles updates)
{{ config(
    materialized='incremental',
    unique_key='order_id',
    incremental_strategy='merge',
    on_schema_change='append_new_columns'
) }}

SELECT *
FROM {{ source('app', 'orders') }}
{% if is_incremental() %}
WHERE updated_at > (SELECT MAX(updated_at) FROM {{ this }})
{% endif %}

-- Delete+Insert (best for partitioned tables)
{{ config(
    materialized='incremental',
    unique_key='order_id',
    incremental_strategy='delete+insert',
    incremental_predicates=[
        "DBT_INTERNAL_DEST.order_date >= dateadd(day, -3, current_date)"
    ]
) }}
```

### Macros

```sql
-- macros/generate_schema_name.sql
-- Control where models land (override default schema behavior)
{% macro generate_schema_name(custom_schema_name, node) %}
    {% if custom_schema_name %}
        {{ custom_schema_name }}
    {% else %}
        {{ target.schema }}
    {% endif %}
{% endmacro %}

-- macros/cents_to_dollars.sql
{% macro cents_to_dollars(column_name) %}
    ({{ column_name }} / 100.0)::numeric(12,2)
{% endmacro %}

-- Usage in model:
SELECT
    order_id,
    {{ cents_to_dollars('amount_cents') }} as amount_dollars
FROM {{ ref('stg_orders') }}
```

### Testing

```yaml
# Schema tests (in YAML)
models:
  - name: fct_revenue
    columns:
      - name: order_id
        data_tests:
          - unique
          - not_null
      - name: total_amount
        data_tests:
          - not_null
          - dbt_utils.accepted_range:
              min_value: 0
              max_value: 100000
      - name: customer_id
        data_tests:
          - relationships:
              to: ref('dim_customers')
              field: customer_id
```

```sql
-- tests/singular/test_revenue_no_negative.sql
-- Custom SQL test: fails if any rows returned
SELECT order_id, total_amount
FROM {{ ref('fct_revenue') }}
WHERE total_amount < 0
```

### dbt CLI Commands

```bash
# Run all models
dbt run

# Run specific model + all upstream dependencies
dbt run --select +fct_revenue

# Run all models in a directory
dbt run --select staging.*

# Run modified models + downstream
dbt run --select state:modified+

# Test everything
dbt test

# Test specific model
dbt test --select fct_revenue

# Generate and serve docs
dbt docs generate && dbt docs serve

# Source freshness check
dbt source freshness

# Full refresh (rebuild incremental from scratch)
dbt run --full-refresh --select fct_revenue

# Compile SQL without running
dbt compile --select fct_revenue
```

### dbt + Airflow Integration

```python
# Option 1: BashOperator (simple)
from airflow.operators.bash import BashOperator

dbt_run = BashOperator(
    task_id="dbt_run",
    bash_command="cd /opt/dbt/project && dbt run --profiles-dir /opt/dbt/profiles",
    env={"DBT_TARGET": "prod"},
)

# Option 2: cosmos (recommended for production)
# pip install astronomer-cosmos
from cosmos import DbtDag, ProjectConfig, ProfileConfig, ExecutionConfig

dbt_dag = DbtDag(
    project_config=ProjectConfig("/opt/dbt/project"),
    profile_config=ProfileConfig(
        profile_name="warehouse",
        target_name="prod",
    ),
    execution_config=ExecutionConfig(
        dbt_executable_path="/usr/local/bin/dbt",
    ),
    schedule="0 8 * * *",
    start_date=datetime(2024, 1, 1),
    dag_id="dbt_analytics",
    # Each dbt model becomes an Airflow task with correct dependencies
)
```

## Airflow Operational Reference

### CLI Commands

```bash
# DAG management
airflow dags list
airflow dags trigger my_dag
airflow dags trigger my_dag --conf '{"key": "value"}'
airflow dags pause my_dag
airflow dags unpause my_dag
airflow dags backfill my_dag -s 2024-01-01 -e 2024-03-01

# Task management
airflow tasks list my_dag
airflow tasks test my_dag my_task 2024-03-01  # Test single task
airflow tasks run my_dag my_task 2024-03-01   # Run single task
airflow tasks clear my_dag -s 2024-03-01 -e 2024-03-01  # Clear for re-run

# Connection management
airflow connections list
airflow connections add postgres_warehouse \
  --conn-type postgres \
  --conn-host warehouse.example.com \
  --conn-port 5432 \
  --conn-schema analytics \
  --conn-login airflow \
  --conn-password secret

# Variable management
airflow variables set environment production
airflow variables get environment
airflow variables import variables.json
```

### Debugging Failed Tasks

```
1. Check task log in Airflow UI (Task Instance → Log)
2. If log is empty: check scheduler logs for parse errors
3. Common failures:
   - ImportError → missing Python package in worker environment
   - Connection refused → check connection config + network
   - XCom size exceeded → pass file paths instead of data
   - Zombie task → increase execution_timeout
   - Worker killed → OOM, increase worker memory
   - Sensor timeout → increase timeout or check upstream
```

### Performance Tuning

```ini
# airflow.cfg

# Scheduler performance
[scheduler]
min_file_process_interval = 30     # Seconds between DAG file re-scans
dag_dir_list_interval = 300        # Seconds between scanning for new DAG files
max_dagruns_to_create_per_loop = 10
max_tis_per_query = 512

# Prevent scheduler overload
[core]
max_active_tasks_per_dag = 16      # Limit per-DAG parallelism
max_active_runs_per_dag = 3        # Limit concurrent DAG runs
parallelism = 32                   # Total cluster-wide parallel tasks

# Database pool (critical for large deployments)
[database]
sql_alchemy_pool_size = 10
sql_alchemy_max_overflow = 20
sql_alchemy_pool_recycle = 1800
```
