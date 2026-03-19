---
name: stream-processing
description: >
  Build real-time and near-real-time stream processing pipelines with Kafka Streams,
  Spark Structured Streaming, and Flink. Covers event-driven architectures, exactly-once
  semantics, windowing, and stateful processing.
  Triggers: "stream processing", "real-time pipeline", "Kafka Streams", "Spark Streaming",
  "event-driven", "streaming ETL", "real-time analytics", "CDC pipeline", "change data capture".
  NOT for: batch ETL (use etl-pipeline-builder), static data modeling (use data-modeling),
  Kafka cluster setup (use kafka-architect agent).
version: 1.0.0
argument-hint: "[streaming use case]"
allowed-tools: Read, Grep, Glob, Bash, Write, Edit
model: sonnet
---

# Stream Processing

Build production-grade stream processing pipelines for real-time analytics, event-driven architectures, and continuous data integration.

## Framework Selection

```
What's the latency requirement?
  │
  ├─ Sub-second (true real-time):
  │   ├─ Simple transforms/aggregations → Kafka Streams
  │   ├─ Complex event processing → Apache Flink
  │   └─ Already in Kafka ecosystem → Kafka Streams
  │
  ├─ Seconds to minutes (near real-time):
  │   ├─ Already using Spark → Spark Structured Streaming
  │   ├─ Need SQL interface → Flink SQL / ksqlDB
  │   └─ Light processing → Kafka Streams
  │
  └─ Minutes (micro-batch):
      └─ Spark Structured Streaming (trigger.processingTime)

Team expertise and existing infrastructure matter more than
benchmark differences. Pick what your team can operate.
```

## Kafka Streams

### Topology Pattern — Stateless Processing

```java
import org.apache.kafka.streams.StreamsBuilder;
import org.apache.kafka.streams.KafkaStreams;
import org.apache.kafka.streams.StreamsConfig;
import org.apache.kafka.streams.kstream.*;
import org.apache.kafka.common.serialization.Serdes;

import java.util.Properties;

public class OrderProcessingTopology {

    public static void main(String[] args) {
        Properties props = new Properties();
        props.put(StreamsConfig.APPLICATION_ID_CONFIG, "order-processing");
        props.put(StreamsConfig.BOOTSTRAP_SERVERS_CONFIG, "kafka:9092");
        props.put(StreamsConfig.DEFAULT_KEY_SERDE_CLASS_CONFIG, Serdes.StringSerde.class);
        props.put(StreamsConfig.DEFAULT_VALUE_SERDE_CLASS_CONFIG, Serdes.StringSerde.class);
        props.put(StreamsConfig.PROCESSING_GUARANTEE_CONFIG, StreamsConfig.EXACTLY_ONCE_V2);
        props.put(StreamsConfig.NUM_STREAM_THREADS_CONFIG, 4);

        StreamsBuilder builder = new StreamsBuilder();

        // Source: raw orders
        KStream<String, Order> orders = builder.stream(
            "raw-orders",
            Consumed.with(Serdes.String(), orderSerde())
        );

        // Branch: valid vs invalid orders
        var branches = orders.split(Named.as("order-"))
            .branch((key, order) -> order.isValid(), Branched.as("valid"))
            .defaultBranch(Branched.as("invalid"));

        KStream<String, Order> validOrders = branches.get("order-valid");
        KStream<String, Order> invalidOrders = branches.get("order-invalid");

        // Enrich valid orders
        KTable<String, Customer> customers = builder.table(
            "customers",
            Consumed.with(Serdes.String(), customerSerde())
        );

        KStream<String, EnrichedOrder> enriched = validOrders
            .selectKey((key, order) -> order.getCustomerId())
            .join(customers,
                (order, customer) -> new EnrichedOrder(order, customer),
                Joined.with(Serdes.String(), orderSerde(), customerSerde())
            );

        // Route enriched orders
        enriched
            .filter((key, order) -> order.getAmount() > 1000)
            .to("high-value-orders", Produced.with(Serdes.String(), enrichedOrderSerde()));

        enriched
            .filter((key, order) -> order.getAmount() <= 1000)
            .to("standard-orders", Produced.with(Serdes.String(), enrichedOrderSerde()));

        // Dead letter queue for invalid orders
        invalidOrders
            .mapValues(order -> new DLQRecord(order, "Validation failed"))
            .to("orders-dlq", Produced.with(Serdes.String(), dlqSerde()));

        // Start the topology
        KafkaStreams streams = new KafkaStreams(builder.build(), props);

        // Graceful shutdown
        Runtime.getRuntime().addShutdownHook(new Thread(streams::close));

        streams.start();
    }
}
```

### Windowed Aggregation

```java
// Real-time revenue per product category, 5-minute tumbling windows
KStream<String, Order> orders = builder.stream("orders");

KTable<Windowed<String>, Double> revenueByCategory = orders
    .groupBy(
        (key, order) -> order.getCategory(),
        Grouped.with(Serdes.String(), orderSerde())
    )
    .windowedBy(TimeWindows.ofSizeWithNoGrace(Duration.ofMinutes(5)))
    .aggregate(
        () -> 0.0,
        (category, order, total) -> total + order.getAmount(),
        Materialized.<String, Double, WindowStore<Bytes, byte[]>>as("revenue-by-category")
            .withValueSerde(Serdes.Double())
            .withRetention(Duration.ofDays(7))
    );

// Emit results downstream
revenueByCategory.toStream()
    .map((windowedKey, revenue) -> KeyValue.pair(
        windowedKey.key(),
        new CategoryRevenue(
            windowedKey.key(),
            revenue,
            windowedKey.window().start(),
            windowedKey.window().end()
        )
    ))
    .to("category-revenue", Produced.with(Serdes.String(), categoryRevenueSerde()));
```

### Session Windows (User Activity)

```java
// Group user events into sessions with 30-minute inactivity gap
KStream<String, UserEvent> events = builder.stream("user-events");

KTable<Windowed<String>, SessionSummary> sessions = events
    .groupByKey()
    .windowedBy(SessionWindows.ofInactivityGapWithNoGrace(Duration.ofMinutes(30)))
    .aggregate(
        SessionSummary::new,
        (userId, event, session) -> session.addEvent(event),
        (userId, session1, session2) -> session1.merge(session2),
        Materialized.with(Serdes.String(), sessionSummarySerde())
    );
```

## Spark Structured Streaming

### Basic Streaming Pipeline

```python
from pyspark.sql import SparkSession
from pyspark.sql.functions import (
    col, from_json, window, sum as spark_sum, count, avg,
    current_timestamp, expr, to_timestamp
)
from pyspark.sql.types import (
    StructType, StructField, StringType, DoubleType, TimestampType, IntegerType
)

spark = SparkSession.builder \
    .appName("orders-streaming") \
    .config("spark.sql.shuffle.partitions", 20) \
    .config("spark.streaming.stopGracefullyOnShutdown", True) \
    .getOrCreate()

# Define schema for incoming events
order_schema = StructType([
    StructField("order_id", StringType(), False),
    StructField("customer_id", StringType(), False),
    StructField("product_id", StringType(), False),
    StructField("category", StringType(), True),
    StructField("amount", DoubleType(), False),
    StructField("quantity", IntegerType(), False),
    StructField("event_time", TimestampType(), False),
])

# Read from Kafka
raw_stream = spark.readStream \
    .format("kafka") \
    .option("kafka.bootstrap.servers", "kafka:9092") \
    .option("subscribe", "orders") \
    .option("startingOffsets", "latest") \
    .option("maxOffsetsPerTrigger", 100000) \
    .option("kafka.group.id", "spark-orders-consumer") \
    .load()

# Parse JSON values
orders = raw_stream \
    .select(from_json(col("value").cast("string"), order_schema).alias("data")) \
    .select("data.*") \
    .withWatermark("event_time", "10 minutes")  # Late event tolerance

# ── Aggregation: 5-minute revenue windows ──
revenue_windows = orders \
    .groupBy(
        window(col("event_time"), "5 minutes"),
        col("category")
    ) \
    .agg(
        spark_sum("amount").alias("total_revenue"),
        count("order_id").alias("order_count"),
        avg("amount").alias("avg_order_value"),
    ) \
    .select(
        col("window.start").alias("window_start"),
        col("window.end").alias("window_end"),
        col("category"),
        col("total_revenue"),
        col("order_count"),
        col("avg_order_value"),
        current_timestamp().alias("processed_at"),
    )

# Write aggregated results to warehouse
revenue_query = revenue_windows.writeStream \
    .format("jdbc") \
    .option("url", "jdbc:postgresql://warehouse:5432/analytics") \
    .option("dbtable", "streaming_revenue") \
    .option("checkpointLocation", "/checkpoints/revenue-windows") \
    .outputMode("update") \
    .trigger(processingTime="1 minute") \
    .start()

# ── Enrichment: Join with static dimension ──
products_dim = spark.read \
    .format("jdbc") \
    .option("url", "jdbc:postgresql://warehouse:5432/analytics") \
    .option("dbtable", "dim_product") \
    .load() \
    .cache()

enriched_orders = orders.join(
    products_dim,
    orders.product_id == products_dim.product_id,
    "left"
)

# Write enriched events back to Kafka
enriched_query = enriched_orders \
    .select(
        col("order_id").alias("key"),
        to_json(struct("*")).alias("value")
    ) \
    .writeStream \
    .format("kafka") \
    .option("kafka.bootstrap.servers", "kafka:9092") \
    .option("topic", "enriched-orders") \
    .option("checkpointLocation", "/checkpoints/enriched-orders") \
    .outputMode("append") \
    .start()

# Wait for all streams
spark.streams.awaitAnyTermination()
```

### Streaming Deduplication

```python
# Deduplicate events within a time window using watermark + dropDuplicates
deduplicated = orders \
    .withWatermark("event_time", "1 hour") \
    .dropDuplicatesWithinWatermark(["order_id"])

# For exact dedup across all time: use a stateful operator
# WARNING: State grows unbounded — add TTL or periodic cleanup
```

### Streaming-to-Streaming Join

```python
# Join two streams: orders + payments (within a 1-hour window)
orders = spark.readStream.format("kafka") \
    .option("subscribe", "orders").load() \
    .select(from_json(col("value").cast("string"), order_schema).alias("o")).select("o.*") \
    .withWatermark("order_time", "2 hours")

payments = spark.readStream.format("kafka") \
    .option("subscribe", "payments").load() \
    .select(from_json(col("value").cast("string"), payment_schema).alias("p")).select("p.*") \
    .withWatermark("payment_time", "3 hours")

# Join on order_id within a time range
joined = orders.join(
    payments,
    expr("""
        o_order_id = p_order_id AND
        payment_time >= order_time AND
        payment_time <= order_time + interval 1 hour
    """),
    "leftOuter"
)
```

## CDC (Change Data Capture) Pipelines

### Debezium → Kafka → Warehouse

```python
"""
CDC pipeline: PostgreSQL → Debezium → Kafka → Spark → Data Warehouse

Debezium captures INSERT/UPDATE/DELETE as events with before/after state.
"""

# Debezium event schema
cdc_schema = StructType([
    StructField("before", StructType([
        StructField("id", IntegerType()),
        StructField("name", StringType()),
        StructField("email", StringType()),
        StructField("updated_at", TimestampType()),
    ])),
    StructField("after", StructType([
        StructField("id", IntegerType()),
        StructField("name", StringType()),
        StructField("email", StringType()),
        StructField("updated_at", TimestampType()),
    ])),
    StructField("op", StringType()),  # c=create, u=update, d=delete
    StructField("ts_ms", LongType()),  # Event timestamp
])

# Read CDC events from Kafka
cdc_stream = spark.readStream \
    .format("kafka") \
    .option("subscribe", "dbserver1.public.customers") \
    .load() \
    .select(from_json(col("value").cast("string"), cdc_schema).alias("cdc")) \
    .select("cdc.*")

# Apply CDC operations to target table
def apply_cdc_batch(batch_df, batch_id):
    """Apply inserts, updates, and deletes to the target table."""
    if batch_df.isEmpty():
        return

    # Separate by operation type
    inserts = batch_df.filter(col("op") == "c").select("after.*")
    updates = batch_df.filter(col("op") == "u").select("after.*")
    deletes = batch_df.filter(col("op") == "d").select("before.id")

    # Get latest state per key (handle multiple events for same row in batch)
    from pyspark.sql.window import Window
    w = Window.partitionBy("after.id").orderBy(col("ts_ms").desc())
    latest = batch_df.withColumn("rn", row_number().over(w)).filter(col("rn") == 1)

    # Apply to Delta Lake / warehouse
    target = DeltaTable.forPath(spark, "/warehouse/customers")

    target.alias("t").merge(
        latest.filter(col("op").isin("c", "u")).select("after.*").alias("s"),
        "t.id = s.id"
    ).whenMatchedUpdateAll() \
     .whenNotMatchedInsertAll() \
     .execute()

    # Handle deletes (soft delete)
    if deletes.count() > 0:
        delete_ids = [row.id for row in deletes.collect()]
        target.update(
            condition=col("id").isin(delete_ids),
            set={"is_deleted": lit(True), "deleted_at": current_timestamp()}
        )

cdc_stream.writeStream \
    .foreachBatch(apply_cdc_batch) \
    .option("checkpointLocation", "/checkpoints/cdc-customers") \
    .trigger(processingTime="30 seconds") \
    .start()
```

## Event-Driven Architecture Patterns

### Event Sourcing with Kafka

```python
"""
Event sourcing: derive current state from a log of immutable events.
Events topic is the source of truth. State is a materialized view.
"""

# Event types
class OrderEvent:
    ORDER_CREATED = "OrderCreated"
    ITEM_ADDED = "ItemAdded"
    ITEM_REMOVED = "ItemRemoved"
    ORDER_SUBMITTED = "OrderSubmitted"
    PAYMENT_RECEIVED = "PaymentReceived"
    ORDER_SHIPPED = "OrderShipped"
    ORDER_DELIVERED = "OrderDelivered"
    ORDER_CANCELLED = "OrderCancelled"


def rebuild_order_state(events: list[dict]) -> dict:
    """Rebuild current order state from event history."""
    state = {"items": [], "status": "unknown", "total": 0}

    for event in sorted(events, key=lambda e: e["timestamp"]):
        match event["type"]:
            case OrderEvent.ORDER_CREATED:
                state = {
                    "order_id": event["data"]["order_id"],
                    "customer_id": event["data"]["customer_id"],
                    "items": [],
                    "status": "created",
                    "total": 0,
                    "created_at": event["timestamp"],
                }
            case OrderEvent.ITEM_ADDED:
                state["items"].append(event["data"]["item"])
                state["total"] += event["data"]["item"]["price"]
            case OrderEvent.ITEM_REMOVED:
                item_id = event["data"]["item_id"]
                state["items"] = [i for i in state["items"] if i["id"] != item_id]
                state["total"] = sum(i["price"] for i in state["items"])
            case OrderEvent.ORDER_SUBMITTED:
                state["status"] = "submitted"
                state["submitted_at"] = event["timestamp"]
            case OrderEvent.PAYMENT_RECEIVED:
                state["status"] = "paid"
                state["paid_at"] = event["timestamp"]
            case OrderEvent.ORDER_SHIPPED:
                state["status"] = "shipped"
                state["shipped_at"] = event["timestamp"]
            case OrderEvent.ORDER_CANCELLED:
                state["status"] = "cancelled"
                state["cancelled_at"] = event["timestamp"]
                state["cancel_reason"] = event["data"].get("reason")

    return state
```

### CQRS Pattern

```
Commands (writes):
  PlaceOrder → validate → append to "order-events" topic → ACK

Queries (reads):
  Kafka Streams app materializes events into queryable state stores
  REST API reads from state stores (KTable / RocksDB)

Benefits:
  - Write path is append-only (fast, simple)
  - Read path is pre-computed (fast queries)
  - Full audit trail via event log
  - Multiple read models from same event stream
```

## Monitoring Streaming Pipelines

### Key Metrics to Track

```python
# Kafka Streams metrics to expose via JMX/Prometheus
CRITICAL_METRICS = {
    # Consumer lag — THE most important metric
    "kafka_consumer_group_lag": "records behind",

    # Processing rate
    "stream_thread_process_rate": "records/sec per thread",
    "stream_thread_process_latency_avg": "avg ms per record",

    # State store health
    "state_store_put_latency_avg": "avg state write ms",
    "state_store_fetch_latency_avg": "avg state read ms",
    "state_store_all_num_entries": "total keys in state",

    # Commit health
    "stream_thread_commit_rate": "commits/sec",
    "stream_thread_commit_latency_avg": "avg commit ms",
}

# Alert thresholds
ALERTS = {
    "consumer_lag > 10000": "WARN: falling behind",
    "consumer_lag > 100000": "CRITICAL: significantly behind",
    "process_latency_avg > 500ms": "WARN: slow processing",
    "commit_latency_avg > 5000ms": "CRITICAL: commit bottleneck",
}
```

### Spark Streaming Monitoring

```python
# StreamingQueryListener for custom monitoring
class MetricsListener(StreamingQueryListener):
    def onQueryProgress(self, event):
        progress = event.progress
        metrics = {
            "input_rows_per_second": progress.inputRowsPerSecond,
            "processed_rows_per_second": progress.processedRowsPerSecond,
            "batch_duration_ms": progress.batchDuration,
            "num_input_rows": progress.numInputRows,
            "state_operators": [
                {
                    "operator": op.operatorName,
                    "num_rows_total": op.numRowsTotal,
                    "memory_used_bytes": op.memoryUsedBytes,
                }
                for op in progress.stateOperators
            ],
        }

        # Push to Prometheus/Datadog
        push_metrics(metrics)

        # Alert on processing delays
        if progress.inputRowsPerSecond > 0:
            ratio = progress.processedRowsPerSecond / progress.inputRowsPerSecond
            if ratio < 0.8:
                alert(f"Processing falling behind: {ratio:.1%} throughput ratio")

spark.streams.addListener(MetricsListener())
```

## Exactly-Once Semantics

### The Three Guarantees

```
At-most-once:  Fire and forget. Fast but may lose data.
At-least-once: Retry on failure. May produce duplicates.
Exactly-once:  Each record processed exactly once. Most complex.

Reality check: "Exactly-once" in distributed systems means
"effectively-once" — it's at-least-once delivery + idempotent processing.
```

### Achieving Exactly-Once

```
Kafka Streams:
  processing.guarantee=exactly_once_v2
  (Uses transactions internally — atomic read-process-write)

Spark Structured Streaming:
  Checkpointing + idempotent sinks
  (Replays from checkpoint on failure, sink must handle duplicates)

Application-level:
  Idempotent writes (upsert, not insert)
  + Deduplication by event ID
  + Transactional outbox pattern
```

### Transactional Outbox Pattern

```sql
-- Instead of: write to DB + publish to Kafka (two-phase commit problem)
-- Do this: write to DB including an outbox table (single transaction)

BEGIN;
  INSERT INTO orders (id, customer_id, amount) VALUES (...);
  INSERT INTO outbox (id, aggregate_type, aggregate_id, event_type, payload)
    VALUES (gen_random_uuid(), 'Order', order_id, 'OrderCreated', '{"..."}');
COMMIT;

-- A CDC connector (Debezium) tails the outbox table and publishes to Kafka
-- Then deletes processed outbox rows
```

## Checklist Before Completing

- [ ] Serialization format chosen (Avro + Schema Registry for production)
- [ ] Watermark configured for late event handling
- [ ] Dead letter queue set up for poison messages
- [ ] Checkpointing enabled for fault tolerance
- [ ] Consumer lag monitoring and alerting configured
- [ ] Exactly-once or idempotent processing verified
- [ ] State store size bounded (TTL or windowed retention)
- [ ] Graceful shutdown handles in-flight records
- [ ] Backpressure strategy defined (throttle source or scale consumers)
- [ ] Schema evolution strategy documented (backward/forward compatible)
