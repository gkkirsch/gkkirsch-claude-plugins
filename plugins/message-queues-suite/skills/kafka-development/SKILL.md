---
name: kafka-development
description: >
  Apache Kafka event streaming with KafkaJS — producers, consumers, consumer groups,
  partitioning, exactly-once semantics, schema registry, and stream processing.
  Triggers: "kafka", "kafkajs", "event streaming", "event bus",
  "pub sub", "consumer group", "event driven architecture".
  NOT for: Simple job queues (use bullmq-development with BullMQ).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Apache Kafka with KafkaJS

## Setup

```bash
npm install kafkajs
```

## Client Configuration

```typescript
// src/lib/kafka.ts
import { Kafka, logLevel } from "kafkajs";

export const kafka = new Kafka({
  clientId: process.env.KAFKA_CLIENT_ID || "my-app",
  brokers: (process.env.KAFKA_BROKERS || "localhost:9092").split(","),
  ssl: process.env.KAFKA_SSL === "true",
  sasl: process.env.KAFKA_SASL_USERNAME
    ? {
        mechanism: "plain",
        username: process.env.KAFKA_SASL_USERNAME,
        password: process.env.KAFKA_SASL_PASSWORD!,
      }
    : undefined,
  logLevel: logLevel.WARN,
  retry: {
    initialRetryTime: 300,
    retries: 10,
  },
  connectionTimeout: 10_000,
  requestTimeout: 30_000,
});
```

## Topic Administration

```typescript
// src/scripts/create-topics.ts
import { kafka } from "../lib/kafka";

async function createTopics() {
  const admin = kafka.admin();
  await admin.connect();

  await admin.createTopics({
    waitForLeaders: true,
    topics: [
      {
        topic: "user-events",
        numPartitions: 6,        // Scale with consumer count
        replicationFactor: 3,    // For production durability
        configEntries: [
          { name: "retention.ms", value: String(7 * 24 * 60 * 60 * 1000) }, // 7 days
          { name: "cleanup.policy", value: "delete" },
        ],
      },
      {
        topic: "order-events",
        numPartitions: 12,
        replicationFactor: 3,
        configEntries: [
          { name: "retention.ms", value: String(30 * 24 * 60 * 60 * 1000) }, // 30 days
        ],
      },
      {
        topic: "dead-letter",
        numPartitions: 1,
        replicationFactor: 3,
        configEntries: [
          { name: "retention.ms", value: String(90 * 24 * 60 * 60 * 1000) }, // 90 days
        ],
      },
    ],
  });

  console.log("Topics created");
  await admin.disconnect();
}

createTopics().catch(console.error);
```

## Producer

```typescript
// src/producers/event.producer.ts
import { kafka } from "../lib/kafka";
import { Partitioners, CompressionTypes } from "kafkajs";

const producer = kafka.producer({
  createPartitioner: Partitioners.DefaultPartitioner,
  allowAutoTopicCreation: false,       // Explicit topic creation in prod
  transactionalId: "order-producer",   // For exactly-once semantics
  maxInFlightRequests: 5,
  idempotent: true,                    // Prevent duplicate messages
});

// Connect on startup
export async function connectProducer() {
  await producer.connect();
  console.log("Kafka producer connected");
}

// Typed event publishing
interface UserEvent {
  type: "user.created" | "user.updated" | "user.deleted";
  userId: string;
  data: Record<string, unknown>;
  timestamp: string;
}

export async function publishUserEvent(event: UserEvent) {
  await producer.send({
    topic: "user-events",
    compression: CompressionTypes.GZIP,
    messages: [
      {
        key: event.userId,             // Same key -> same partition (ordering)
        value: JSON.stringify(event),
        headers: {
          "event-type": event.type,
          "correlation-id": crypto.randomUUID(),
          "source": "user-service",
        },
      },
    ],
  });
}

// Batch publishing
export async function publishBatch(events: UserEvent[]) {
  await producer.sendBatch({
    compression: CompressionTypes.GZIP,
    topicMessages: [
      {
        topic: "user-events",
        messages: events.map((event) => ({
          key: event.userId,
          value: JSON.stringify(event),
          headers: { "event-type": event.type },
        })),
      },
    ],
  });
}

// Transactional publishing (exactly-once)
export async function publishWithTransaction(events: UserEvent[]) {
  const transaction = await producer.transaction();
  try {
    for (const event of events) {
      await transaction.send({
        topic: "user-events",
        messages: [{ key: event.userId, value: JSON.stringify(event) }],
      });
    }
    await transaction.commit();
  } catch (err) {
    await transaction.abort();
    throw err;
  }
}

export async function disconnectProducer() {
  await producer.disconnect();
}
```

## Consumer

```typescript
// src/consumers/user-events.consumer.ts
import { kafka } from "../lib/kafka";
import { EachMessagePayload } from "kafkajs";

const consumer = kafka.consumer({
  groupId: "user-service-group",      // Consumer group for load balancing
  sessionTimeout: 30_000,             // How long before rebalance on crash
  heartbeatInterval: 3_000,           // Keep-alive frequency
  maxBytesPerPartition: 1_048_576,    // 1MB per partition fetch
  retry: {
    retries: 5,
  },
});

export async function startUserEventsConsumer() {
  await consumer.connect();

  // Subscribe to topic(s)
  await consumer.subscribe({
    topics: ["user-events"],
    fromBeginning: false,              // Only new messages (true = replay all)
  });

  // Process messages
  await consumer.run({
    autoCommit: true,                  // Auto-commit offsets
    autoCommitInterval: 5000,          // Commit every 5s
    autoCommitThreshold: 100,          // Or every 100 messages
    eachMessage: async ({ topic, partition, message, heartbeat }: EachMessagePayload) => {
      const event = JSON.parse(message.value!.toString());
      const eventType = message.headers?.["event-type"]?.toString();

      console.log(`[${topic}:${partition}] ${eventType}`, {
        key: message.key?.toString(),
        offset: message.offset,
      });

      try {
        switch (eventType) {
          case "user.created":
            await handleUserCreated(event);
            break;
          case "user.updated":
            await handleUserUpdated(event);
            break;
          case "user.deleted":
            await handleUserDeleted(event);
            break;
          default:
            console.warn(`Unknown event type: ${eventType}`);
        }

        // Keep session alive during long processing
        await heartbeat();
      } catch (err) {
        // Send to dead letter queue on failure
        await publishToDeadLetter(topic, message, err as Error);
      }
    },
  });
}

async function handleUserCreated(event: any) {
  // Idempotent: check if already processed
  const existing = await db.users.findUnique({ where: { id: event.userId } });
  if (existing) return; // Already processed — skip

  await db.users.create({ data: event.data });
}

async function handleUserUpdated(event: any) {
  await db.users.update({
    where: { id: event.userId },
    data: event.data,
  });
}

async function handleUserDeleted(event: any) {
  await db.users.delete({ where: { id: event.userId } });
}

// Dead letter queue for failed messages
async function publishToDeadLetter(originalTopic: string, message: any, error: Error) {
  const dlqProducer = kafka.producer();
  await dlqProducer.connect();
  await dlqProducer.send({
    topic: "dead-letter",
    messages: [{
      key: message.key,
      value: JSON.stringify({
        originalTopic,
        originalMessage: message.value?.toString(),
        error: error.message,
        stack: error.stack,
        failedAt: new Date().toISOString(),
      }),
      headers: {
        "original-topic": originalTopic,
        "error-type": error.constructor.name,
      },
    }],
  });
  await dlqProducer.disconnect();
}
```

## Manual Offset Commit (At-Least-Once with Control)

```typescript
await consumer.run({
  autoCommit: false,      // Manual commit
  eachMessage: async ({ topic, partition, message }) => {
    const event = JSON.parse(message.value!.toString());

    // Process the message
    await processEvent(event);

    // Commit AFTER successful processing
    await consumer.commitOffsets([{
      topic,
      partition,
      offset: (BigInt(message.offset) + 1n).toString(),
    }]);
  },
});
```

## Batch Processing

```typescript
await consumer.run({
  eachBatch: async ({ batch, resolveOffset, heartbeat, isRunning, isStale }) => {
    for (const message of batch.messages) {
      // Check if consumer is still running (graceful shutdown)
      if (!isRunning() || isStale()) break;

      const event = JSON.parse(message.value!.toString());
      await processEvent(event);

      // Mark this offset as processed
      resolveOffset(message.offset);

      // Keep session alive
      await heartbeat();
    }
  },
});
```

## Partitioning Strategies

```typescript
import { Partitioners } from "kafkajs";

// Default: murmur2 hash of key
const producer = kafka.producer({
  createPartitioner: Partitioners.DefaultPartitioner,
});

// Custom partitioner
const customProducer = kafka.producer({
  createPartitioner: () => {
    return ({ topic, partitionMetadata, message }) => {
      // Route by region
      const region = message.headers?.["region"]?.toString() || "us";
      const partitions = partitionMetadata.length;

      const regionMap: Record<string, number> = {
        us: 0,
        eu: 1,
        asia: 2,
      };

      return regionMap[region] ?? (partitions - 1);
    };
  },
});
```

## Schema Validation

```typescript
// src/schemas/user-event.schema.ts
import { z } from "zod";

export const UserEventSchema = z.object({
  type: z.enum(["user.created", "user.updated", "user.deleted"]),
  userId: z.string().uuid(),
  data: z.record(z.unknown()),
  timestamp: z.string().datetime(),
});

export type UserEvent = z.infer<typeof UserEventSchema>;

// Validate before producing
export async function publishValidatedEvent(event: unknown) {
  const validated = UserEventSchema.parse(event);  // Throws on invalid
  await producer.send({
    topic: "user-events",
    messages: [{
      key: validated.userId,
      value: JSON.stringify(validated),
    }],
  });
}

// Validate on consume
async function handleMessage(message: any) {
  const raw = JSON.parse(message.value!.toString());
  const result = UserEventSchema.safeParse(raw);

  if (!result.success) {
    console.error("Invalid event schema:", result.error);
    await publishToDeadLetter("user-events", message, new Error("Schema validation failed"));
    return;
  }

  await processEvent(result.data);
}
```

## Graceful Shutdown

```typescript
// src/server.ts
import { startUserEventsConsumer } from "./consumers/user-events.consumer";
import { connectProducer, disconnectProducer } from "./producers/event.producer";

let consumer: ReturnType<typeof kafka.consumer>;

async function start() {
  await connectProducer();
  consumer = await startUserEventsConsumer();
  console.log("Kafka services started");
}

async function shutdown(signal: string) {
  console.log(`Received ${signal}. Shutting down...`);

  // Disconnect consumer first (stop receiving)
  if (consumer) {
    await consumer.disconnect();
  }

  // Then disconnect producer (flush pending)
  await disconnectProducer();

  console.log("Kafka services stopped");
  process.exit(0);
}

process.on("SIGTERM", () => shutdown("SIGTERM"));
process.on("SIGINT", () => shutdown("SIGINT"));

start().catch(console.error);
```

## Docker Compose (Local Development)

```yaml
# docker-compose.yml
services:
  kafka:
    image: bitnami/kafka:3.7
    ports:
      - "9092:9092"
    environment:
      - KAFKA_CFG_NODE_ID=1
      - KAFKA_CFG_PROCESS_ROLES=broker,controller
      - KAFKA_CFG_CONTROLLER_QUORUM_VOTERS=1@kafka:9093
      - KAFKA_CFG_LISTENERS=PLAINTEXT://:9092,CONTROLLER://:9093
      - KAFKA_CFG_ADVERTISED_LISTENERS=PLAINTEXT://localhost:9092
      - KAFKA_CFG_CONTROLLER_LISTENER_NAMES=CONTROLLER
      - KAFKA_CFG_INTER_BROKER_LISTENER_NAME=PLAINTEXT
      # KRaft mode (no ZooKeeper)
    volumes:
      - kafka_data:/bitnami/kafka

  kafka-ui:
    image: provectuslabs/kafka-ui:latest
    ports:
      - "8080:8080"
    environment:
      KAFKA_CLUSTERS_0_NAME: local
      KAFKA_CLUSTERS_0_BOOTSTRAPSERVERS: kafka:9092
    depends_on:
      - kafka

volumes:
  kafka_data:
```

```bash
docker compose up -d
# Kafka at localhost:9092, UI at localhost:8080
```

## Health Check

```typescript
// src/health/kafka.health.ts
import { kafka } from "../lib/kafka";

export async function checkKafkaHealth(): Promise<{
  status: "healthy" | "unhealthy";
  details: Record<string, unknown>;
}> {
  const admin = kafka.admin();
  try {
    await admin.connect();
    const topics = await admin.listTopics();
    const groups = await admin.listGroups();

    // Check consumer group lag
    const groupDescriptions = await admin.describeGroups(
      groups.groups.map((g) => g.groupId)
    );

    await admin.disconnect();

    return {
      status: "healthy",
      details: {
        topicCount: topics.length,
        consumerGroups: groups.groups.length,
        topics,
      },
    };
  } catch (err) {
    await admin.disconnect().catch(() => {});
    return {
      status: "unhealthy",
      details: { error: (err as Error).message },
    };
  }
}
```

## Gotchas

1. **Consumer group rebalancing delays processing** — When a consumer joins/leaves a group, ALL consumers pause for rebalancing (default ~25s). Set `sessionTimeout` and `rebalanceTimeout` appropriately. With many consumers, rebalances get slow.

2. **Message ordering is per-partition only** — Messages with different keys may go to different partitions. If you need global ordering, use 1 partition (kills throughput) or ensure related messages share a key.

3. **`fromBeginning: true` replays ALL messages** — If a topic has millions of messages, this takes hours. Only use for new consumer groups or explicit replay scenarios. Default `false` starts from the latest offset.

4. **KafkaJS is maintenance-mode** — The original KafkaJS library has limited maintenance. For new projects, consider `@confluentinc/kafka-javascript` (Confluent's official Node.js client built on librdkafka) for better performance and active support.

5. **Offset commits are batched** — With `autoCommit`, offsets are committed periodically, not per-message. If the consumer crashes between commits, some messages will be reprocessed. Make handlers idempotent.

6. **Partition count cannot decrease** — You can add partitions to an existing topic but never remove them. Plan your initial partition count carefully. Start with `max(expected_consumers, 6)`.
