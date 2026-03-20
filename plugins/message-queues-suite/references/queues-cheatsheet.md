# Message Queues & Background Jobs Cheatsheet

## BullMQ Quick Reference

### Setup
```bash
npm install bullmq ioredis
```

### Queue + Worker
```typescript
import { Queue, Worker } from "bullmq";

const connection = { host: "localhost", port: 6379, maxRetriesPerRequest: null };

const queue = new Queue("tasks", {
  connection,
  defaultJobOptions: {
    attempts: 3,
    backoff: { type: "exponential", delay: 1000 },
    removeOnComplete: { age: 86400, count: 1000 },
    removeOnFail: { age: 604800 },
  },
});

const worker = new Worker("tasks", async (job) => {
  await job.updateProgress(50);
  return { result: "done" };
}, { connection, concurrency: 5 });

worker.on("completed", (job) => console.log(`Done: ${job.id}`));
worker.on("failed", (job, err) => console.error(`Failed: ${job?.id}`, err));
```

### Adding Jobs
```typescript
await queue.add("send-email", { to: "user@test.com" });                     // Basic
await queue.add("reminder", data, { delay: 3600000 });                       // Delayed (1h)
await queue.add("urgent", data, { priority: 1 });                            // Priority
await queue.add("daily", data, { repeat: { pattern: "0 9 * * *" } });       // Cron
await queue.add("unique", data, { jobId: `user-${userId}` });               // Deduplicated
await queue.addBulk([{ name: "a", data: d1 }, { name: "b", data: d2 }]);   // Bulk
```

### Flows (Parent-Child)
```typescript
import { FlowProducer } from "bullmq";
const flow = new FlowProducer({ connection });
await flow.add({
  name: "parent", queueName: "orders", data: {},
  children: [
    { name: "child1", queueName: "payments", data: {} },
    { name: "child2", queueName: "inventory", data: {} },
  ],
});
```

### Dashboard
```bash
npm install @bull-board/express @bull-board/api
```

### Shutdown
```typescript
await worker.close();
await queue.close();
```

---

## KafkaJS Quick Reference

### Setup
```bash
npm install kafkajs
```

### Client
```typescript
import { Kafka } from "kafkajs";
const kafka = new Kafka({
  clientId: "my-app",
  brokers: ["localhost:9092"],
});
```

### Producer
```typescript
const producer = kafka.producer({ idempotent: true });
await producer.connect();
await producer.send({
  topic: "events",
  messages: [{ key: "user-123", value: JSON.stringify(event) }],
});
await producer.disconnect();
```

### Consumer
```typescript
const consumer = kafka.consumer({ groupId: "my-group" });
await consumer.connect();
await consumer.subscribe({ topics: ["events"] });
await consumer.run({
  eachMessage: async ({ topic, partition, message }) => {
    const event = JSON.parse(message.value!.toString());
    await processEvent(event);
  },
});
```

### Docker (KRaft, no ZooKeeper)
```yaml
services:
  kafka:
    image: bitnami/kafka:3.7
    ports: ["9092:9092"]
    environment:
      KAFKA_CFG_NODE_ID: 1
      KAFKA_CFG_PROCESS_ROLES: broker,controller
      KAFKA_CFG_CONTROLLER_QUORUM_VOTERS: 1@kafka:9093
      KAFKA_CFG_LISTENERS: PLAINTEXT://:9092,CONTROLLER://:9093
      KAFKA_CFG_ADVERTISED_LISTENERS: PLAINTEXT://localhost:9092
      KAFKA_CFG_CONTROLLER_LISTENER_NAMES: CONTROLLER
```

---

## When to Use What

| Scenario | Technology |
|----------|-----------|
| Email sending, image processing, PDF generation | **BullMQ** |
| Cron jobs, scheduled tasks | **BullMQ** |
| Rate-limited API calls | **BullMQ** |
| User activity streaming (100K+ events/sec) | **Kafka** |
| Event sourcing / audit log | **Kafka** |
| Microservice communication (pub/sub) | **Kafka** or **RabbitMQ** |
| Complex routing (topic, fanout, headers) | **RabbitMQ** |
| Simple cloud queue, zero ops | **AWS SQS** |
| Webhook delivery with retries | **BullMQ** |
| Real-time analytics pipeline | **Kafka** |
| Job dependencies (parent-child) | **BullMQ** |

---

## Job Design Checklist

- [ ] Job handler is **idempotent** (safe to retry)
- [ ] Job data is **small** (reference IDs, not full payloads)
- [ ] **Retry strategy** configured (attempts + exponential backoff)
- [ ] **Dead letter queue** for exhausted retries
- [ ] **Timeout** set for long-running jobs
- [ ] **Concurrency** limit on workers
- [ ] **Graceful shutdown** handles SIGTERM
- [ ] **Health check** endpoint monitors queue connection
- [ ] **Monitoring** for queue depth, processing time, failure rate
- [ ] **Stale job cleanup** (removeOnComplete, removeOnFail)

---

## Common Retry Strategies

| Strategy | Config | Delays |
|----------|--------|--------|
| Fixed | `{ type: "fixed", delay: 5000 }` | 5s, 5s, 5s |
| Exponential | `{ type: "exponential", delay: 1000 }` | 1s, 2s, 4s, 8s |
| Custom | `backoffStrategies` registration | Any pattern |

---

## Kafka Partition Rules

| Rule | Example |
|------|---------|
| Same key = same partition | `key: userId` ensures all user events ordered |
| Partitions >= consumers | 6 partitions = max 6 consumers in a group |
| Can add, never remove | Start with `max(consumers, 6)` |
| No key = round-robin | Messages distributed evenly |

---

## Quick Debug Commands

```bash
# BullMQ — check Redis queue
redis-cli KEYS "bull:*"
redis-cli LLEN "bull:email:wait"     # Waiting jobs
redis-cli ZCARD "bull:email:delayed"  # Delayed jobs

# Kafka — command line tools
kafka-topics.sh --list --bootstrap-server localhost:9092
kafka-console-consumer.sh --topic events --from-beginning --bootstrap-server localhost:9092
kafka-consumer-groups.sh --describe --group my-group --bootstrap-server localhost:9092
```
