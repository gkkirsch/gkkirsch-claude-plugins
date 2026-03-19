# Apache Kafka Architecture & Operations Reference

## Cluster Architecture

### KRaft Mode (Kafka 3.3+, recommended)

```
┌─────────────────────────────────────────────────────┐
│                   Kafka Cluster                      │
│                                                      │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐          │
│  │Controller │  │Controller │  │Controller │  Quorum  │
│  │ (voter)  │  │ (voter)  │  │ (voter)  │          │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘          │
│       │              │              │                │
│  ┌────┴─────┐  ┌────┴─────┐  ┌────┴─────┐          │
│  │ Broker 1 │  │ Broker 2 │  │ Broker 3 │          │
│  │ P0-lead  │  │ P1-lead  │  │ P2-lead  │          │
│  │ P1-repl  │  │ P2-repl  │  │ P0-repl  │          │
│  └──────────┘  └──────────┘  └──────────┘          │
│                                                      │
│  No ZooKeeper dependency. Metadata in internal log.  │
└─────────────────────────────────────────────────────┘
```

### Combined vs Dedicated Mode

```
Combined mode (small clusters, ≤5 brokers):
  Each node runs both controller + broker processes
  Simpler ops, fewer machines
  process.roles=broker,controller

Dedicated mode (production, >5 brokers):
  3 dedicated controller nodes (no broker traffic)
  N broker nodes (data only)
  Better isolation, predictable performance
  controller: process.roles=controller
  broker: process.roles=broker
```

## Broker Configuration Reference

### Essential Settings

```properties
# server.properties

# Identity
node.id=1
process.roles=broker,controller  # or just broker / controller
controller.quorum.voters=1@controller1:9093,2@controller2:9093,3@controller3:9093

# Listeners
listeners=PLAINTEXT://:9092,CONTROLLER://:9093
advertised.listeners=PLAINTEXT://broker1.example.com:9092
inter.broker.listener.name=PLAINTEXT
controller.listener.names=CONTROLLER

# Log storage
log.dirs=/data/kafka/logs
num.partitions=6                    # Default partitions for auto-created topics
default.replication.factor=3        # Default replication
min.insync.replicas=2               # Require 2 replicas ACK for acks=all

# Retention
log.retention.hours=168             # 7 days default
log.retention.bytes=-1              # No size limit (use hours instead)
log.segment.bytes=1073741824        # 1GB segments
log.cleanup.policy=delete           # delete or compact

# Performance
num.network.threads=8               # Network I/O threads
num.io.threads=16                   # Disk I/O threads
socket.send.buffer.bytes=102400
socket.receive.buffer.bytes=102400
socket.request.max.bytes=104857600  # 100MB max request

# Replication
replica.lag.time.max.ms=30000       # Follower max lag before removal from ISR
num.replica.fetchers=4

# Auto topic creation (disable in production)
auto.create.topics.enable=false
```

### Sizing Guidelines

| Metric | Small | Medium | Large |
|--------|-------|--------|-------|
| Messages/sec | <10K | 10K-100K | >100K |
| Brokers | 3 | 5-10 | 10-30+ |
| Partitions/broker | <1000 | 1000-4000 | 4000+ |
| Disk/broker | 500GB | 2TB | 4TB+ |
| Memory/broker | 16GB | 32GB | 64GB |
| CPU/broker | 8 cores | 16 cores | 32 cores |
| Network | 1Gbps | 10Gbps | 25Gbps |

**JVM Heap for brokers**: 6GB is usually optimal. Don't go above 8GB. Kafka relies on OS page cache, not JVM heap.

```bash
export KAFKA_HEAP_OPTS="-Xms6g -Xmx6g"
export KAFKA_JVM_PERFORMANCE_OPTS="-XX:+UseG1GC -XX:MaxGCPauseMillis=20 -XX:InitiatingHeapOccupancyPercent=35"
```

## Topic Design

### Partition Count Formula

```
partitions = max(
    target_throughput_MB_s / partition_throughput_MB_s,
    target_consumer_parallelism
)

# Single partition throughput:
#   Producer: ~10 MB/s (single partition, compressed)
#   Consumer: ~30 MB/s (single partition)

# Example: 200 MB/s target write throughput
#   200 / 10 = 20 partitions minimum
#   Round up to 24 (multiple of replication factor 3)

# Consumer parallelism:
#   max consumers in a group = number of partitions
#   If you need 12 parallel consumers → at least 12 partitions
```

### Key Design Patterns

```
1. Entity-per-topic:
   orders, customers, payments — one topic per entity type
   Key: entity ID (order_id, customer_id)
   Guarantees ordering per entity

2. Event-type-per-topic:
   order-created, order-shipped, order-cancelled
   Pro: consumers subscribe to specific events
   Con: more topics to manage

3. Domain-per-topic (recommended):
   order-events (all order lifecycle events)
   Key: order_id
   Value includes event_type field for filtering
   Best balance of ordering guarantees + simplicity
```

### Compacted Topics

```properties
# For changelog / latest-state topics
cleanup.policy=compact
min.cleanable.dirty.ratio=0.5    # Trigger compaction at 50% dirty
segment.ms=86400000              # Force roll after 24h
delete.retention.ms=86400000     # Keep tombstones for 24h

# Use case: KTable state stores, configuration, user profiles
# Key = entity ID, latest value wins, null value = tombstone (delete)
```

## Producer Configuration

### Reliability Settings

```properties
# Exactly-once / idempotent producer
enable.idempotence=true          # Prevents duplicates from retries
acks=all                         # Wait for all ISR replicas
max.in.flight.requests.per.connection=5  # Max with idempotence

# Retries
retries=2147483647               # Effectively infinite (rely on delivery.timeout)
delivery.timeout.ms=120000       # 2 min total delivery timeout
retry.backoff.ms=100             # Initial retry delay

# Batching (throughput vs latency trade-off)
batch.size=65536                 # 64KB batch
linger.ms=5                     # Wait up to 5ms to fill batch
buffer.memory=67108864           # 64MB buffer

# Compression
compression.type=lz4             # Best throughput. zstd for best ratio.
```

### Partitioner Strategies

```
Default (hash):
  partition = hash(key) % num_partitions
  Guarantees same key → same partition → ordering per key
  Problem: hot keys → hot partitions

Round-robin (null key):
  Even distribution, no ordering guarantee

Sticky (Kafka 2.4+, default for null keys):
  Fills one batch before moving to next partition
  Better batching efficiency than round-robin

Custom:
  Implement Partitioner interface for business logic
  Example: route VIP customers to dedicated partitions
```

## Consumer Configuration

### Essential Settings

```properties
# Consumer group
group.id=my-consumer-group
group.instance.id=consumer-1      # Static membership (faster rebalance)

# Offset management
auto.offset.reset=earliest         # or latest
enable.auto.commit=false           # Manual commit for exactly-once

# Session management
session.timeout.ms=45000           # Max time without heartbeat
heartbeat.interval.ms=3000         # Heartbeat frequency (session.timeout / 15)
max.poll.interval.ms=300000        # Max time between poll() calls
max.poll.records=500               # Records per poll()

# Fetch tuning
fetch.min.bytes=1                  # Min data per fetch (increase for throughput)
fetch.max.bytes=52428800           # 50MB max per fetch
max.partition.fetch.bytes=1048576  # 1MB max per partition per fetch
fetch.max.wait.ms=500              # Max wait for fetch.min.bytes
```

### Consumer Group Rebalancing

```
Eager rebalance (default before 3.1):
  All consumers stop → reassign all partitions → resume
  Downtime during rebalance = seconds to minutes

Cooperative rebalance (recommended):
  partition.assignment.strategy=
    org.apache.kafka.clients.consumer.CooperativeStickyAssignor

  Only revoked partitions stop → reassign those → resume
  Minimal disruption, no full-stop

Static membership (production recommendation):
  group.instance.id=consumer-{hostname}
  session.timeout.ms=300000  # 5 min — tolerates brief restarts

  Consumer keeps partition assignment across restarts (if within timeout)
  Eliminates rebalance storms during rolling deploys
```

## Schema Registry

### Avro Schema Example

```json
{
  "type": "record",
  "name": "Order",
  "namespace": "com.example.events",
  "fields": [
    {"name": "order_id", "type": "string"},
    {"name": "customer_id", "type": "string"},
    {"name": "amount", "type": "double"},
    {"name": "currency", "type": {"type": "string", "default": "USD"}},
    {"name": "status", "type": {
      "type": "enum",
      "name": "OrderStatus",
      "symbols": ["CREATED", "CONFIRMED", "SHIPPED", "DELIVERED", "CANCELLED"]
    }},
    {"name": "items", "type": {"type": "array", "items": {
      "type": "record",
      "name": "OrderItem",
      "fields": [
        {"name": "product_id", "type": "string"},
        {"name": "quantity", "type": "int"},
        {"name": "unit_price", "type": "double"}
      ]
    }}},
    {"name": "event_time", "type": {"type": "long", "logicalType": "timestamp-millis"}},
    {"name": "metadata", "type": ["null", "string"], "default": null}
  ]
}
```

### Schema Evolution Rules

```
BACKWARD compatible (default, recommended):
  New schema can read OLD data
  ✓ Add fields with defaults
  ✓ Remove fields
  ✗ Change field types
  ✗ Rename fields
  Deploy: consumers first, then producers

FORWARD compatible:
  OLD schema can read new data
  ✓ Add fields
  ✓ Remove fields with defaults
  Deploy: producers first, then consumers

FULL compatible (safest):
  Both directions work
  ✓ Add fields with defaults
  ✓ Remove fields with defaults

NONE:
  No compatibility checking — don't use in production
```

## Monitoring

### Critical Metrics

```
BROKER HEALTH:
  UnderReplicatedPartitions        > 0 = replication problem
  OfflinePartitionsCount           > 0 = partitions unavailable
  ActiveControllerCount            != 1 = split brain or no controller
  UncleanLeaderElections/sec       > 0 = potential data loss
  IsrShrinksPerSec / IsrExpandsPerSec  Watch for flapping

THROUGHPUT:
  BytesInPerSec / BytesOutPerSec   Per broker, per topic
  MessagesInPerSec                 Production rate
  RequestsPerSec                   API request rate

LATENCY:
  RequestQueueTimeMs               Time waiting in queue
  ResponseQueueTimeMs              Time waiting to send response
  TotalTimeMs                      Total request processing time
  ProduceRequestPurgatorySize      Requests waiting for acks

RESOURCE:
  NetworkProcessorAvgIdlePercent   < 0.3 = network bottleneck
  RequestHandlerAvgIdlePercent     < 0.3 = I/O bottleneck
  LogFlushRateAndTimeMs            Disk write performance
  linux.disk.read/write_bytes      OS disk I/O
```

### Consumer Lag Monitoring

```bash
# Check consumer lag via CLI
kafka-consumer-groups.sh --bootstrap-server kafka:9092 \
  --describe --group my-consumer-group

# Output:
# GROUP    TOPIC     PARTITION  CURRENT-OFFSET  LOG-END-OFFSET  LAG
# my-group orders    0          1000000         1000050         50
# my-group orders    1          950000          950000          0
# my-group orders    2          980000          990000          10000  ← Problem!
```

### Alert Thresholds

| Metric | Warning | Critical |
|--------|---------|----------|
| Consumer lag (records) | >10,000 | >100,000 |
| UnderReplicatedPartitions | >0 for >5min | >0 for >15min |
| Produce latency p99 | >100ms | >1000ms |
| RequestHandlerIdlePct | <0.3 | <0.1 |
| Disk usage | >70% | >85% |

## Operational Runbook

### Topic Operations

```bash
# Create topic
kafka-topics.sh --bootstrap-server kafka:9092 \
  --create --topic orders \
  --partitions 12 --replication-factor 3 \
  --config retention.ms=604800000 \
  --config min.insync.replicas=2

# Increase partitions (can't decrease!)
kafka-topics.sh --bootstrap-server kafka:9092 \
  --alter --topic orders --partitions 24

# Describe topic
kafka-topics.sh --bootstrap-server kafka:9092 \
  --describe --topic orders

# Delete topic
kafka-topics.sh --bootstrap-server kafka:9092 \
  --delete --topic orders

# List all topics
kafka-topics.sh --bootstrap-server kafka:9092 --list
```

### Consumer Group Operations

```bash
# List consumer groups
kafka-consumer-groups.sh --bootstrap-server kafka:9092 --list

# Reset offsets to earliest
kafka-consumer-groups.sh --bootstrap-server kafka:9092 \
  --group my-group --reset-offsets --to-earliest \
  --topic orders --execute

# Reset offsets to specific timestamp
kafka-consumer-groups.sh --bootstrap-server kafka:9092 \
  --group my-group --reset-offsets \
  --to-datetime 2024-03-01T00:00:00.000 \
  --topic orders --execute

# Reset offsets to specific offset
kafka-consumer-groups.sh --bootstrap-server kafka:9092 \
  --group my-group --reset-offsets \
  --to-offset 1000000 \
  --topic orders:0 --execute  # partition 0 only
```

### Reassign Partitions

```bash
# Generate reassignment plan
kafka-reassign-partitions.sh --bootstrap-server kafka:9092 \
  --generate \
  --topics-to-move-json-file topics.json \
  --broker-list "1,2,3,4,5"

# Execute reassignment
kafka-reassign-partitions.sh --bootstrap-server kafka:9092 \
  --execute \
  --reassignment-json-file plan.json \
  --throttle 50000000  # 50MB/s throttle to avoid impacting production

# Verify completion
kafka-reassign-partitions.sh --bootstrap-server kafka:9092 \
  --verify \
  --reassignment-json-file plan.json
```

### Recovery Procedures

```
Broker failure:
  1. Check if ISR > min.insync.replicas for affected partitions
  2. If yes: automatic failover, no data loss
  3. If no: either wait for broker recovery OR accept unclean election
     kafka-leader-election.sh --election-type UNCLEAN --topic X --partition Y

Consumer stuck:
  1. kafka-consumer-groups.sh --describe → check lag and assignment
  2. Check max.poll.interval.ms (consumer might be ejected)
  3. If rebalance loop: check for long-running processing, increase poll interval
  4. Nuclear option: delete consumer group, consumers re-join fresh

Disk full:
  1. kafka-log-dirs.sh --describe → find largest topics
  2. Reduce retention: kafka-configs.sh --alter --entity-name X --add-config retention.ms=3600000
  3. Force log cleanup: kafka-log-dirs.sh --describe to verify segments deleted
  4. Move partitions to brokers with more space
```
