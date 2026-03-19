# Messaging Systems Reference

Quick reference for messaging systems and patterns. For full architectural context, see the system-design-architect and reliability-engineer agents.

---

## Apache Kafka

Distributed event streaming platform. High throughput, durable, ordered.

### Core Concepts

```
Producer → Topic → Partition → Consumer Group

┌──────────┐     ┌──────────────────────────────────┐
│ Producer │────►│           Topic: orders            │
└──────────┘     │  ┌──────────┐ ┌──────────┐       │
                 │  │Partition 0│ │Partition 1│       │
                 │  │ [0][1][2] │ │ [0][1][2] │       │
                 │  └──────────┘ └──────────┘       │
                 └──────────────────────────────────┘
                          │              │
                   ┌──────▼──────┐ ┌─────▼───────┐
                   │ Consumer A  │ │ Consumer B  │
                   │ (Group: svc)│ │ (Group: svc)│
                   └─────────────┘ └─────────────┘

Topic: Named stream of events (like a database table)
Partition: Ordered, immutable sequence of messages (parallelism unit)
Offset: Position of a message within a partition
Consumer Group: Set of consumers that divide partition ownership
```

### Partitioning

```
Messages are assigned to partitions by key:
  partition = hash(key) % num_partitions

Key choice determines ordering guarantees:
  Key: user_id → All events for a user go to same partition → ordered per user
  Key: order_id → All events for an order go to same partition → ordered per order
  Key: null → Round-robin distribution → no ordering guarantee, max throughput

Number of partitions:
  Rule of thumb: max(expected_throughput / per_partition_throughput, num_consumers)

  Each partition: ~10 MB/s write, ~30 MB/s read (single consumer)
  Need 100 MB/s write: at least 10 partitions
  Need 20 consumers: at least 20 partitions

  Start with: 3-6 partitions per topic (small/medium scale)
  Scale to: 12-50 partitions (high throughput)
  Warning: Too many partitions → more memory, longer leader election, more files
```

### Consumer Groups

```
Partitions are divided among consumers in a group:

Topic with 4 partitions, Consumer Group with 2 consumers:
  Consumer A: Partition 0, Partition 1
  Consumer B: Partition 2, Partition 3

Topic with 4 partitions, Consumer Group with 4 consumers:
  Consumer A: Partition 0
  Consumer B: Partition 1
  Consumer C: Partition 2
  Consumer D: Partition 3

Topic with 4 partitions, Consumer Group with 6 consumers:
  Consumer A: Partition 0
  Consumer B: Partition 1
  Consumer C: Partition 2
  Consumer D: Partition 3
  Consumer E: IDLE (no partition assigned)
  Consumer F: IDLE (no partition assigned)

→ More consumers than partitions = wasted consumers
→ Partitions = maximum parallelism within a consumer group

Multiple consumer groups can read the same topic independently:
  Group "order-service": Processing orders
  Group "analytics": Building dashboards
  Group "audit-log": Writing to audit store
  Each group maintains its own offsets → independent consumption
```

### Key Configuration

```
Producer:
  acks=all           Highest durability (all replicas acknowledge)
  acks=1             Leader only (faster, slight data loss risk)
  acks=0             Fire and forget (fastest, no durability)

  retries=3          Retry transient failures
  enable.idempotence=true   Prevent duplicate messages on retry
  max.in.flight.requests.per.connection=5  (with idempotence)

  Recommended (high reliability):
    acks=all, enable.idempotence=true, retries=MAX_INT

  Recommended (high throughput):
    acks=1, batch.size=65536, linger.ms=5

Consumer:
  auto.offset.reset=earliest    Start from beginning (new consumer group)
  auto.offset.reset=latest      Start from now (skip history)
  enable.auto.commit=false      Manual offset commit (exactly-once)
  max.poll.records=500          Max messages per poll
  session.timeout.ms=30000      Consumer failure detection
  max.poll.interval.ms=300000   Max time between polls before rebalance

Broker:
  replication.factor=3           3 copies of each partition
  min.insync.replicas=2          At least 2 replicas must ACK
  unclean.leader.election.enable=false  Don't promote out-of-sync replica

  These three settings together = no data loss configuration
```

### Exactly-Once Semantics

```
Kafka provides exactly-once within the Kafka ecosystem:

1. Idempotent Producer (within Kafka):
   enable.idempotence=true
   Kafka deduplicates messages based on producer ID + sequence number
   Guarantees: Each message written exactly once to the partition

2. Transactions (across topics/partitions):
   Producer sends to multiple topics/partitions atomically
   All or nothing — no partial writes

3. Consumer + External System (end-to-end exactly-once):
   Kafka can't guarantee exactly-once to external systems
   Consumer must be idempotent:
     Read message → Process → Commit offset (in same transaction)
     If crash before commit → message re-delivered → idempotent processing handles it

   Implementation:
     // Transactional consumer pattern
     while (true) {
       messages = consumer.poll();
       db.beginTransaction();
       for (msg of messages) {
         // Check idempotency
         if (await isProcessed(msg.key)) continue;
         // Process
         await processMessage(msg);
         await markProcessed(msg.key);
       }
       db.commit();
       consumer.commitSync();
     }
```

### Common Kafka Patterns

```
Event Sourcing:
  Topic = event store
  Compact topic retains latest value per key
  log.cleanup.policy=compact

Change Data Capture (CDC):
  Database → Debezium → Kafka → Consumers
  Stream database changes as events

Stream Processing:
  Kafka Streams / Flink / ksqlDB
  Real-time aggregations, joins, windowing

Dead Letter Queue:
  Failed messages → DLQ topic → manual review / retry
  After N retries: send to orders.dlq topic
  Alert if DLQ grows beyond threshold
```

---

## RabbitMQ

Message broker with advanced routing capabilities. Lower throughput than Kafka but richer routing.

### Core Concepts

```
Producer → Exchange → Binding → Queue → Consumer

┌──────────┐     ┌──────────┐     ┌───────┐     ┌──────────┐
│ Producer │────►│ Exchange │────►│ Queue │────►│ Consumer │
└──────────┘     └──────────┘     └───────┘     └──────────┘

Exchange types:
  Direct:  Route by exact routing key match
  Topic:   Route by pattern matching (*.error, logs.#)
  Fanout:  Broadcast to all bound queues
  Headers: Route by message header values
```

### Exchange Types

```
Direct Exchange:
  Routing key = queue name (simple point-to-point)

  Producer → Exchange[routing_key=payment] → Queue[payment] → Consumer

  Use for: Task distribution, RPC, simple routing

Topic Exchange:
  Routing key = dot-separated pattern with wildcards
  * = exactly one word,  # = zero or more words

  Producer sends: routing_key = "order.created.us"
  Queue A binds: "order.created.*"  → matches!
  Queue B binds: "order.#"          → matches!
  Queue C binds: "payment.*"        → no match

  Use for: Log routing, event distribution by category

Fanout Exchange:
  All messages go to all bound queues (ignore routing key)

  Producer → Exchange → Queue A, Queue B, Queue C (all get the message)

  Use for: Broadcasting, pub/sub, notifications

Headers Exchange:
  Route based on message header values (not routing key)
  x-match: all (AND) or any (OR)

  Use for: Complex routing based on message properties
```

### Message Acknowledgment

```
Manual ACK (recommended for reliability):
  1. Consumer receives message
  2. Consumer processes message
  3. Consumer sends ACK to RabbitMQ
  4. RabbitMQ removes message from queue

  If consumer crashes before ACK → message redelivered to another consumer

  channel.consume(queue, (msg) => {
    try {
      processMessage(msg);
      channel.ack(msg);         // Success → acknowledge
    } catch (error) {
      channel.nack(msg, false, true);  // Failure → requeue
      // (msg, allUpTo, requeue)
    }
  }, { noAck: false });

Auto ACK (noAck: true):
  Message removed as soon as delivered (before processing)
  Fast but no delivery guarantee
  Use only for: Non-critical messages (logs, metrics)

NACK with dead lettering:
  channel.nack(msg, false, false);  // Don't requeue → goes to DLQ
  Configure DLQ: x-dead-letter-exchange, x-dead-letter-routing-key
```

### RabbitMQ vs Kafka

```
Feature              RabbitMQ                    Kafka
──────────────────────────────────────────────────────────────
Model                Message broker              Event log
Consumption          Destructive (message removed) Non-destructive (log retained)
Ordering             Per-queue (FIFO)            Per-partition
Routing              Sophisticated (exchanges)    Simple (topic-based)
Throughput           ~50K msg/sec                ~1M msg/sec
Message size         Optimized for small          Handles large batches
Replay               No (message consumed once)   Yes (re-read from offset)
Consumer groups      Competing consumers          Consumer groups
Use case             Task queues, RPC, routing    Event streaming, CDC, analytics
Protocol             AMQP, STOMP, MQTT            Custom (Kafka protocol)
```

### When to Choose

```
Choose RabbitMQ:
  - Complex routing requirements (topic, header-based routing)
  - Task queues / work distribution
  - RPC patterns (request-reply)
  - Message priority queues
  - Small to medium throughput (< 100K msg/sec)
  - Need per-message acknowledgment and redelivery
  - MQTT/STOMP protocol support needed (IoT)

Choose Kafka:
  - High throughput event streaming (> 100K msg/sec)
  - Event sourcing / event log
  - Need to replay messages (reprocess from offset)
  - Change data capture (CDC)
  - Stream processing (real-time analytics)
  - Multiple consumers reading same data independently
  - Long-term message retention (days to forever)
  - Strict ordering within partition
```

---

## Amazon SQS

Fully managed message queue. Simple, reliable, serverless.

### Standard vs FIFO Queues

```
Standard Queue:
  - At-least-once delivery (occasional duplicates)
  - Best-effort ordering (not guaranteed)
  - Nearly unlimited throughput
  - Use for: Background jobs, notifications, decoupling services

FIFO Queue:
  - Exactly-once processing
  - Strict ordering (within message group)
  - 3,000 messages/sec (with batching)
  - Use for: Financial transactions, order processing, sequential workflows

  Message Group ID → guarantees order within group
  Different groups can be processed in parallel

  Group "user-123": msg1 → msg2 → msg3 (ordered)
  Group "user-456": msg4 → msg5 → msg6 (ordered)
  Processing: msg1 and msg4 can be processed in parallel
```

### Key Configuration

```
Visibility Timeout:
  How long a message is invisible after being received.
  Consumer must delete the message before timeout or it reappears.

  Default: 30 seconds
  Set to: > expected processing time + buffer
  If processing takes 10 seconds: set to 30 seconds

  Long processing? Extend visibility timeout before it expires.

Dead Letter Queue:
  After N failed attempts → message moves to DLQ
  maxReceiveCount: 5 (typical)

  Main Queue → 5 attempts → DLQ

  Monitor DLQ:
  - CloudWatch alarm on ApproximateNumberOfMessagesVisible > 0
  - Investigate and replay failed messages

Long Polling:
  WaitTimeSeconds: 20 (receive waits up to 20s for messages)
  Reduces empty receives (and costs)
  Always use long polling (WaitTimeSeconds > 0)

Message retention:
  Default: 4 days
  Max: 14 days
  After retention: messages deleted permanently
```

### SQS + Lambda

```
SQS triggers Lambda function automatically:

┌──────────┐     ┌───────┐     ┌──────────┐
│ Producer │────►│  SQS  │────►│  Lambda  │
└──────────┘     └───────┘     └──────────┘

Configuration:
  batchSize: 10          Process up to 10 messages per invocation
  maxBatchingWindow: 5s  Wait up to 5s to fill batch
  concurrency: 10        Max 10 concurrent Lambda executions

Error handling:
  - Lambda fails → messages return to queue after visibility timeout
  - After maxReceiveCount → messages go to DLQ
  - Partial batch failure: report failed message IDs
```

---

## Pub/Sub Patterns

### Fan-Out

One message → multiple consumers (each gets a copy).

```
             ┌─────────────┐
         ┌──►│ Consumer A  │  (email notification)
         │   └─────────────┘
┌───────┐│   ┌─────────────┐
│ Event ├┼──►│ Consumer B  │  (SMS notification)
└───────┘│   └─────────────┘
         │   ┌─────────────┐
         └──►│ Consumer C  │  (push notification)
             └─────────────┘

Implementation:
  Kafka: Multiple consumer groups on same topic
  RabbitMQ: Fanout exchange
  SQS: SNS fan-out to multiple SQS queues
  Redis: PUB/SUB channels
```

### Fan-In

Multiple producers → single consumer (aggregate).

```
┌───────────┐     ┌───────┐     ┌──────────┐
│ Service A │────►│       │     │          │
└───────────┘     │ Queue │────►│ Consumer │
┌───────────┐     │       │     │          │
│ Service B │────►│       │     └──────────┘
└───────────┘     └───────┘
┌───────────┐        ▲
│ Service C │────────┘
└───────────┘

Use for: Log aggregation, metric collection, event processing
```

### Request-Reply

Synchronous communication over async messaging.

```
┌──────────┐  request   ┌───────────┐  request   ┌──────────┐
│ Client   │───────────►│ Request Q │───────────►│ Server   │
│          │            └───────────┘            │          │
│          │  reply     ┌───────────┐  reply     │          │
│          │◄───────────│ Reply Q   │◄───────────│          │
└──────────┘            └───────────┘            └──────────┘

Implementation:
  Client creates temporary reply queue
  Sends request with reply_to = reply queue name
  Waits for response on reply queue
  Timeout after N seconds

  RabbitMQ: Built-in RPC pattern with correlation_id
  Kafka: Request topic + response topic with correlation header
```

---

## Delivery Guarantees

```
At-Most-Once:
  Send and forget. Message may be lost.
  Producer sends → no ACK → if fails, message is gone

  Use for: Metrics, logs (where occasional loss is acceptable)

At-Least-Once:
  Retry until acknowledged. Message may be duplicated.
  Producer sends → waits for ACK → retries on failure
  Consumer must be idempotent

  Use for: Most business events, task processing

Exactly-Once:
  Message processed exactly once. Hardest to achieve.

  Within Kafka: Idempotent producer + transactional consumer
  End-to-end: At-least-once + idempotent consumer (deduplication)

  True exactly-once is effectively:
  at-least-once delivery + idempotent processing = exactly-once semantics
```

---

## Dead Letter Queues (DLQ)

Messages that can't be processed after multiple attempts go to a DLQ for investigation.

```
Normal flow:
  Queue → Consumer → Process → ACK → Done

Failure flow:
  Queue → Consumer → Fail → Retry (3x) → Fail → DLQ

┌───────────┐     ┌──────────┐     ┌──────────┐
│   Main    │────►│ Consumer │─ fail 3x ─►│   DLQ    │
│   Queue   │     │          │            │          │
└───────────┘     └──────────┘            └──────────┘

DLQ handling:
  1. Alert on DLQ depth > 0
  2. Investigate: Why are messages failing?
     - Parse error? Fix consumer, replay messages
     - Downstream service down? Wait for recovery, replay
     - Poison message? Log and discard
  3. Replay: Move messages from DLQ back to main queue
  4. Monitor: Track DLQ rate as a quality metric

Implementation (SQS):
  Create DLQ: same type as main queue (Standard or FIFO)
  Configure main queue: RedrivePolicy
  {
    "deadLetterTargetArn": "arn:aws:sqs:...:my-queue-dlq",
    "maxReceiveCount": 5
  }

Implementation (Kafka):
  No built-in DLQ — implement in consumer:
  try {
    processMessage(msg);
  } catch (error) {
    if (retryCount >= MAX_RETRIES) {
      producer.send({ topic: 'orders.dlq', value: msg });
    } else {
      // Retry with backoff
    }
  }

Implementation (RabbitMQ):
  Dead letter exchange + routing:
  Queue arguments:
    x-dead-letter-exchange: "dlx"
    x-dead-letter-routing-key: "dlq.orders"
    x-message-ttl: 60000  (optional: TTL before dead-lettering)
```

---

## Message Ordering

```
Kafka:
  Ordering guaranteed within a partition.
  Use message key to route related messages to same partition.
  Key: order_id → all events for order 123 go to partition N → ordered

RabbitMQ:
  Ordering guaranteed within a single queue with a single consumer.
  Multiple consumers on same queue → no ordering guarantee.
  Use single consumer or partition messages into separate queues.

SQS Standard:
  Best-effort ordering only. NOT guaranteed.
  Use FIFO queue with MessageGroupId for strict ordering.

SQS FIFO:
  Strict ordering within MessageGroupId.
  Different groups can be processed in parallel.

General rule:
  If you need ordering, you need:
  1. A partitioning strategy (key/group)
  2. Single consumer per partition/group
  3. Sequential processing within partition
```

---

## Messaging Anti-Patterns

```
1. Mega-messages (> 256 KB in SQS, > 1 MB in Kafka)
   Fix: Store payload in S3/DB, send reference in message
   Pattern: Claim check — message contains pointer to data

2. No idempotency
   Fix: Deduplicate by message ID in consumer
   Every consumer must handle receiving the same message twice

3. Tight coupling via message content
   Fix: Use schema registry (Avro, Protobuf) with versioning
   Consumers should tolerate unknown fields (forward compatibility)

4. No DLQ
   Fix: Always configure a DLQ for every queue
   Failing messages shouldn't block the queue forever

5. Synchronous over async
   Fix: Don't use request-reply pattern everywhere
   If you're waiting for every response, you're just doing HTTP with extra steps

6. No monitoring
   Fix: Monitor queue depth, consumer lag, processing time, error rate
   Alert on: growing queue depth, consumer lag, DLQ messages

7. Unbounded retries
   Fix: Set max retry count, then DLQ
   Infinite retries can amplify load during outages
```

---

## Selection Guide

```
Need                                    System
──────────────────────────────────────────────────────────────
High throughput event streaming         Kafka
Simple task queue                       SQS / RabbitMQ
Complex routing                         RabbitMQ
Managed/serverless                      SQS / SNS / Google Pub/Sub
Event replay / reprocessing             Kafka
Priority queues                         RabbitMQ
Request-reply (RPC)                     RabbitMQ
IoT / MQTT                              RabbitMQ / EMQX
Exactly-once (within system)            Kafka (transactional)
Global pub/sub                          Google Pub/Sub / SNS
Low latency pub/sub                     Redis Pub/Sub / NATS
Stream processing                       Kafka + Kafka Streams / Flink
```
