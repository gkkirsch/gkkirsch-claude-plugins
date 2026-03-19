# Event Streaming Reference

Deep reference for event streaming platforms: Apache Kafka, RabbitMQ, NATS JetStream, AWS SNS/SQS,
and Google Cloud Pub/Sub. Covers configuration, operations, monitoring, and best practices.

## Apache Kafka

### Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                        Kafka Cluster                             │
│                                                                  │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐                      │
│  │ Broker 1 │  │ Broker 2 │  │ Broker 3 │                      │
│  │          │  │          │  │          │                      │
│  │ Topic A  │  │ Topic A  │  │ Topic A  │                      │
│  │ Part 0   │  │ Part 1   │  │ Part 2   │  (leader)           │
│  │ Part 1*  │  │ Part 2*  │  │ Part 0*  │  (* = replica)      │
│  └──────────┘  └──────────┘  └──────────┘                      │
│                                                                  │
│  ┌──────────────────────────────────────┐                       │
│  │ KRaft Controller (replaces ZooKeeper)│                       │
│  └──────────────────────────────────────┘                       │
└─────────────────────────────────────────────────────────────────┘
```

### Topic Configuration Reference

| Config | Default | Recommended | Description |
|--------|---------|-------------|-------------|
| `partitions` | 1 | 6-24 | Number of partitions (can only increase) |
| `replication.factor` | 1 | 3 | Number of replicas per partition |
| `min.insync.replicas` | 1 | 2 | Min replicas that must ack a write |
| `retention.ms` | 604800000 (7d) | Use case specific | How long to retain messages |
| `retention.bytes` | -1 (unlimited) | Use case specific | Max size per partition |
| `cleanup.policy` | delete | delete or compact | How to handle old data |
| `max.message.bytes` | 1048576 (1MB) | 1-10MB | Max message size |
| `compression.type` | producer | lz4 or zstd | Broker-level compression |
| `message.timestamp.type` | CreateTime | CreateTime | Message timestamp source |
| `segment.bytes` | 1073741824 (1GB) | 100-500MB | Log segment size |
| `segment.ms` | 604800000 (7d) | 3600000 (1h) | Log segment rotation time |
| `unclean.leader.election.enable` | false | false | Allow out-of-sync leader election |

### Producer Configuration

```typescript
// KafkaJS producer config
const producer = kafka.producer({
  // Exactly-once semantics (EOS)
  idempotent: true,
  transactionalId: 'order-service-producer',

  // Batching
  allowAutoTopicCreation: false,
  maxInFlightRequests: 5,

  // Retry
  retry: {
    initialRetryTime: 300,
    maxRetryTime: 30000,
    retries: 10,
    factor: 2,
    multiplier: 1.5,
  },
});

// Send with acks=all for durability
await producer.send({
  topic: 'order-events',
  compression: CompressionTypes.LZ4,
  acks: -1,  // all replicas must acknowledge
  timeout: 30000,
  messages: [{
    key: orderId,
    value: JSON.stringify(event),
    headers: {
      'event-type': 'order.created',
      'correlation-id': correlationId,
    },
  }],
});
```

**Producer tuning parameters:**

| Parameter | Default | High Throughput | Low Latency | High Durability |
|-----------|---------|-----------------|-------------|-----------------|
| `acks` | 1 | 1 | 1 | -1 (all) |
| `batch.size` | 16384 | 65536 | 1 | 16384 |
| `linger.ms` | 0 | 50-100 | 0 | 5 |
| `compression` | none | lz4/zstd | none | lz4 |
| `buffer.memory` | 32MB | 64MB | 32MB | 32MB |
| `max.in.flight` | 5 | 5 | 1 | 5 (with idempotent) |
| `retries` | MAX_INT | MAX_INT | 3 | MAX_INT |
| `enable.idempotence` | true | true | false | true |

### Consumer Configuration

```typescript
// KafkaJS consumer config
const consumer = kafka.consumer({
  groupId: 'order-processing-group',

  // Session management
  sessionTimeout: 30000,
  heartbeatInterval: 3000,
  rebalanceTimeout: 60000,

  // Fetch tuning
  maxBytesPerPartition: 1048576,  // 1MB per partition
  maxBytes: 10485760,              // 10MB total per fetch
  minBytes: 1,                     // Don't wait for batch
  maxWaitTimeInMs: 500,

  // Offset management
  retry: {
    initialRetryTime: 1000,
    retries: 5,
  },
});

// Subscribe and run
await consumer.subscribe({
  topics: ['order-events', 'payment-events'],
  fromBeginning: false,
});

await consumer.run({
  autoCommit: false,  // Manual commit for at-least-once
  eachMessage: async ({ topic, partition, message }) => {
    try {
      await processMessage(message);
      await consumer.commitOffsets([{
        topic,
        partition,
        offset: (parseInt(message.offset) + 1).toString(),
      }]);
    } catch (error) {
      // Handle error, send to DLQ
    }
  },
});
```

**Consumer tuning parameters:**

| Parameter | Default | High Throughput | Low Latency |
|-----------|---------|-----------------|-------------|
| `fetch.min.bytes` | 1 | 10000 | 1 |
| `fetch.max.wait.ms` | 500 | 500 | 10 |
| `max.partition.fetch.bytes` | 1MB | 10MB | 1MB |
| `auto.offset.reset` | latest | earliest | latest |
| `enable.auto.commit` | true | true | false |
| `auto.commit.interval.ms` | 5000 | 5000 | N/A |
| `max.poll.records` | 500 | 1000 | 1 |
| `session.timeout.ms` | 30000 | 30000 | 10000 |

### Consumer Group Rebalancing

```
Rebalance triggers:
1. Consumer joins group
2. Consumer leaves group (crash, graceful shutdown)
3. Topic partition count changes
4. Subscription pattern matches new topic

Rebalancing strategies:
- Range: Assigns contiguous partitions to consumers
- RoundRobin: Distributes partitions evenly
- Sticky: Minimizes partition movement during rebalance
- Cooperative: Incremental rebalance (no stop-the-world)

Recommended: CooperativeStickyAssignor (Kafka 2.4+)
```

### Kafka Monitoring Metrics

| Metric | Alert Threshold | Description |
|--------|----------------|-------------|
| `kafka.consumer.lag` | >10000 | Messages behind current offset |
| `kafka.broker.under_replicated_partitions` | >0 | Partitions missing replicas |
| `kafka.broker.offline_partitions_count` | >0 | Partitions with no leader |
| `kafka.broker.active_controller_count` | !=1 | Must be exactly 1 controller |
| `kafka.broker.request_handler_idle_ratio` | <0.5 | Request handler saturation |
| `kafka.broker.network_processor_idle_ratio` | <0.5 | Network thread saturation |
| `kafka.producer.request_latency_avg` | >100ms | Producer to broker latency |
| `kafka.producer.record_error_rate` | >0.01 | Failed produce requests |
| `kafka.consumer.fetch_latency_avg` | >500ms | Consumer fetch latency |
| `kafka.consumer.commit_latency_avg` | >200ms | Offset commit latency |

### Kafka Connect

```json
{
  "name": "postgres-sink-connector",
  "config": {
    "connector.class": "io.confluent.connect.jdbc.JdbcSinkConnector",
    "connection.url": "jdbc:postgresql://read-db:5432/orders_read",
    "connection.user": "${file:/secrets/db-user}",
    "connection.password": "${file:/secrets/db-password}",
    "topics": "order-events",
    "insert.mode": "upsert",
    "pk.mode": "record_value",
    "pk.fields": "orderId",
    "auto.create": true,
    "auto.evolve": true,
    "batch.size": 100,
    "max.retries": 5,
    "retry.backoff.ms": 3000,
    "errors.tolerance": "all",
    "errors.deadletterqueue.topic.name": "dlq-postgres-sink",
    "errors.deadletterqueue.topic.replication.factor": 3,
    "errors.deadletterqueue.context.headers.enable": true
  }
}
```

## RabbitMQ

### Architecture Overview

```
┌──────────────────────────────────────────────────────────────────┐
│                     RabbitMQ Cluster                              │
│                                                                   │
│  ┌────────────┐    ┌────────────┐    ┌────────────┐             │
│  │  Node 1    │    │  Node 2    │    │  Node 3    │             │
│  │ (disc)     │    │ (disc)     │    │ (ram)      │             │
│  └────────────┘    └────────────┘    └────────────┘             │
│                                                                   │
│  Exchanges: topic, direct, fanout, headers                       │
│  Queues: classic, quorum, stream                                 │
│  Bindings: routing key matching                                  │
└──────────────────────────────────────────────────────────────────┘
```

### Exchange Types

| Exchange Type | Routing Logic | Use Case |
|-------------|---------------|----------|
| Direct | Exact routing key match | Point-to-point, RPC |
| Topic | Pattern matching (`*.error`, `order.#`) | Event routing by type |
| Fanout | Broadcast to all bound queues | Notifications, broadcast |
| Headers | Header attribute matching | Complex routing logic |

**Routing key patterns (topic exchange):**

```
*  matches exactly one word:  order.* matches order.created, not order.item.added
#  matches zero or more words: order.# matches order, order.created, order.item.added

Examples:
  order.created     → matches "order.created", "order.*", "order.#", "#"
  order.item.added  → matches "order.item.added", "order.item.*", "order.#", "#"
  payment.failed    → matches "payment.failed", "payment.*", "payment.#", "#"
```

### Queue Types

| Queue Type | Durability | Ordering | Replicated | Use Case |
|-----------|------------|----------|------------|----------|
| Classic | Durable/transient | FIFO per queue | Mirrored (deprecated) | General purpose |
| Quorum | Always durable | FIFO per queue | Raft consensus | Production, critical data |
| Stream | Always durable | Append-only log | Replicated | High throughput, replay |

**Quorum queue configuration:**

```typescript
// Recommended for production
await channel.assertQueue('order.processing', {
  durable: true,
  arguments: {
    'x-queue-type': 'quorum',
    'x-quorum-initial-group-size': 3,
    'x-delivery-limit': 5,                  // Max redeliveries before DLQ
    'x-dead-letter-exchange': 'x.dlq',
    'x-dead-letter-routing-key': 'order.processing.dlq',
    'x-max-length': 100000,                 // Max messages in queue
    'x-overflow': 'reject-publish',          // Reject when full
  },
});
```

### RabbitMQ Monitoring Metrics

| Metric | Alert Threshold | Description |
|--------|----------------|-------------|
| `rabbitmq_queue_messages` | >50000 | Queue depth |
| `rabbitmq_queue_consumers` | ==0 | No consumers attached |
| `rabbitmq_queue_message_rate` | varies | Publish/consume rate |
| `rabbitmq_node_mem_used` | >80% limit | Memory usage |
| `rabbitmq_node_disk_free` | <2GB | Disk space |
| `rabbitmq_connections` | >5000 | Connection count |
| `rabbitmq_channels` | >10000 | Channel count |
| `rabbitmq_unacked_messages` | >1000 | Unacknowledged messages |

### Message Reliability Guarantees

```
Publisher Confirms + Consumer Acks = At-Least-Once Delivery

Publisher side:
  channel.confirmSelect()           # Enable confirms
  channel.publish(exchange, key, msg, { persistent: true })
  channel.waitForConfirms()         # Wait for broker ack

Consumer side:
  channel.consume(queue, handler, { noAck: false })  # Manual ack
  channel.ack(message)              # Acknowledge after processing
  channel.nack(message, false, true) # Reject and requeue
  channel.nack(message, false, false) # Reject without requeue (→ DLQ)
```

## NATS JetStream

### Architecture Overview

```
┌──────────────────────────────────────────────────────────────────┐
│                      NATS Cluster                                 │
│                                                                   │
│  Core NATS: Fire-and-forget pub/sub (at-most-once)              │
│  JetStream: Persistent streaming (at-least-once, exactly-once)  │
│                                                                   │
│  ┌──────────┐    ┌──────────┐    ┌──────────┐                   │
│  │ Server 1 │    │ Server 2 │    │ Server 3 │                   │
│  │ (stream  │    │ (stream  │    │ (stream  │                   │
│  │  leader) │    │ follower)│    │ follower)│                   │
│  └──────────┘    └──────────┘    └──────────┘                   │
└──────────────────────────────────────────────────────────────────┘
```

### Stream Configuration

| Config | Default | Description |
|--------|---------|-------------|
| `Retention` | limits | limits, interest, workqueue |
| `MaxConsumers` | -1 | Max consumers per stream |
| `MaxMsgs` | -1 | Max messages in stream |
| `MaxBytes` | -1 | Max bytes in stream |
| `MaxAge` | 0 (unlimited) | Max age of messages |
| `Storage` | file | file or memory |
| `Replicas` | 1 | Number of replicas (1 or 3) |
| `Discard` | old | old or new (what to discard when full) |
| `DuplicateWindow` | 2m | Window for message deduplication |
| `MaxMsgSize` | -1 | Max size per message |

### Consumer Types

| Consumer | Description | Use Case |
|----------|-------------|----------|
| Pull | Consumer explicitly requests messages | Batch processing, backpressure |
| Push | Server pushes messages to consumer | Real-time processing |
| Durable | Persists position across restarts | Production consumers |
| Ephemeral | Position lost on disconnect | Temporary consumers, testing |

### NATS vs Kafka vs RabbitMQ

| Feature | NATS JetStream | Kafka | RabbitMQ |
|---------|---------------|-------|----------|
| Operational complexity | Very low | High | Medium |
| Memory footprint | Small (~50MB) | Large (~1GB+) | Medium (~300MB) |
| Built-in clustering | Yes (RAFT) | Yes (KRaft/ZK) | Yes (Erlang) |
| Message dedup | Built-in | Producer idempotency | Manual |
| Subject-based routing | Native | No (topic-based) | Exchange/binding |
| Request/reply | Native | Manual | RPC pattern |
| Wildcard subscriptions | `>` and `*` | No | `#` and `*` |
| Key-value store | Built-in | No | No |
| Object store | Built-in | No | No |
| WebSocket support | Built-in | No | Plugin |
| Throughput | High (~1M msg/s) | Very high (~2M msg/s) | Medium (~50K msg/s) |
| Latency | Very low (<1ms) | Low (~2-5ms) | Very low (<1ms) |

## AWS SNS/SQS

### SNS + SQS Fan-Out Pattern

```
Producer → SNS Topic → SQS Queue 1 (Order Processing)
                      → SQS Queue 2 (Notification Service)
                      → SQS Queue 3 (Analytics Pipeline)
                      → Lambda Function (Real-time Processing)
```

### SQS Configuration

| Feature | Standard Queue | FIFO Queue |
|---------|---------------|------------|
| Throughput | Unlimited | 3,000 msg/s (batching) |
| Ordering | Best-effort | Strict FIFO |
| Delivery | At-least-once | Exactly-once |
| Dedup | Manual | Built-in (5 min window) |
| Max message size | 256 KB | 256 KB |
| Retention | 1 min - 14 days | 1 min - 14 days |
| Visibility timeout | 0 - 12 hours | 0 - 12 hours |
| Dead letter queue | Yes | Yes (must be FIFO) |
| Long polling | Yes (up to 20s) | Yes (up to 20s) |

### SQS Best Practices

```typescript
// AWS SDK v3 — SQS consumer
import { SQSClient, ReceiveMessageCommand, DeleteMessageCommand } from '@aws-sdk/client-sqs';

const sqs = new SQSClient({ region: 'us-east-1' });

async function pollMessages(): Promise<void> {
  while (true) {
    const response = await sqs.send(new ReceiveMessageCommand({
      QueueUrl: process.env.QUEUE_URL,
      MaxNumberOfMessages: 10,        // Batch for efficiency
      WaitTimeSeconds: 20,             // Long polling (reduces costs)
      VisibilityTimeout: 60,           // Processing time window
      MessageAttributeNames: ['All'],
    }));

    if (!response.Messages || response.Messages.length === 0) continue;

    for (const message of response.Messages) {
      try {
        await processMessage(JSON.parse(message.Body!));

        await sqs.send(new DeleteMessageCommand({
          QueueUrl: process.env.QUEUE_URL,
          ReceiptHandle: message.ReceiptHandle!,
        }));
      } catch (error) {
        // Message becomes visible again after VisibilityTimeout
        // After maxReceiveCount, goes to DLQ
        logger.error({ err: error, messageId: message.MessageId }, 'Failed to process message');
      }
    }
  }
}
```

## Event Schema Best Practices

### Schema Registry

Use a schema registry (Confluent, AWS Glue, Apicurio) to:
- Store and version event schemas
- Enforce compatibility rules
- Generate code from schemas
- Validate messages at producer/consumer

### Compatibility Modes

| Mode | Add Field | Remove Field | Rename | Change Type |
|------|-----------|-------------|--------|-------------|
| Backward | Optional only | Yes | No | No |
| Forward | Yes | Optional only | No | No |
| Full | Optional only | Optional only | No | No |
| None | Yes | Yes | Yes | Yes |

**Recommended: BACKWARD compatibility** — new consumers can read old events.

### Event Schema Checklist

- [ ] Event type follows naming convention (`domain.aggregate.verb_past_tense`)
- [ ] Event ID is UUID, unique per event
- [ ] Timestamp is ISO 8601 UTC
- [ ] Correlation ID present for tracing
- [ ] Schema version tracked in envelope
- [ ] All new fields are optional with sensible defaults
- [ ] No field renames (add new, deprecate old)
- [ ] No field type changes (add new field with new type)
- [ ] Event is self-contained (no need to call other services to interpret)
- [ ] Sensitive data excluded or encrypted
- [ ] Event size within broker limits
- [ ] Schema registered in schema registry

## Operational Best Practices

### Capacity Planning

| Metric | Formula | Example |
|--------|---------|---------|
| Storage per day | `msg_rate * avg_msg_size * 86400` | 1000/s * 1KB * 86400 = 82GB/day |
| Partitions needed | `target_throughput / per_partition_throughput` | 10K/s / 1K/s = 10 partitions |
| Replication storage | `storage * replication_factor` | 82GB * 3 = 246GB/day |
| Consumer instances | `partitions / (msgs_per_consumer * processing_time)` | 10 / (100 * 0.01) = 10 |
| Network bandwidth | `msg_rate * avg_msg_size * (1 + replication_factor)` | 1K/s * 1KB * 4 = 4MB/s |

### Dead Letter Queue Strategy

```
1. Message fails processing
2. Retry N times with backoff (3-5 attempts)
3. Send to DLQ with error metadata
4. Alert on DLQ depth
5. Investigate and fix root cause
6. Replay DLQ messages after fix
7. Monitor for regression
```

### Message Ordering Guarantees

| Guarantee | How | Trade-off |
|-----------|-----|-----------|
| No ordering | Random partition key | Max throughput |
| Per-entity ordering | Entity ID as partition key | Balanced |
| Per-aggregate ordering | Aggregate ID as partition key | Strong consistency per aggregate |
| Global ordering | Single partition | Lowest throughput |
| Causal ordering | Vector clocks + consumer logic | High complexity |
