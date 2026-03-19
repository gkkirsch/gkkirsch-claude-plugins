# Queue Technology Comparison

## Full Feature Matrix

| Feature | BullMQ | Celery | Sidekiq | AWS SQS | RabbitMQ | Kafka | Temporal | Inngest |
|---------|--------|--------|---------|---------|----------|-------|----------|---------|
| **Language** | Node.js | Python | Ruby | Any | Any | Any | Any | TypeScript |
| **Broker** | Redis | Redis/RabbitMQ | Redis | Managed | Self/Managed | Self/Managed | Self/Cloud | Managed |
| **Priority Queues** | Yes | Yes | Yes | No (workaround) | Yes | No | Yes | Yes |
| **Delayed Jobs** | Yes | Yes | Yes | Yes (15min max) | Yes (plugin) | No | Yes | Yes |
| **Repeatable/Cron** | Yes | Yes (Celery Beat) | Yes (sidekiq-cron) | No (use EventBridge) | No (use plugin) | No | Yes | Yes |
| **Rate Limiting** | Yes | Yes | Yes (Enterprise) | No | No | Consumer lag | Yes | Yes |
| **Job Dependencies** | Yes (Flows) | Yes (Canvas) | No | No | No | Topic ordering | Yes (workflows) | Yes (steps) |
| **Dead Letter Queue** | Manual | Yes | Yes | Yes (native) | Yes | Yes | No (retries) | Yes |
| **Dashboard** | Bull Board | Flower | Sidekiq Web | CloudWatch | Management UI | Kafka UI | Web UI | Dashboard |
| **Persistence** | Redis (AOF) | Broker-dependent | Redis (AOF) | Managed (durable) | Disk | Disk | DB | Managed |
| **Exactly-Once** | No (at-least-once) | No | No | No (at-least-once) | No | Yes (0.11+) | Yes | Yes |
| **Message Size** | ~512MB (Redis) | Varies | ~512MB | 256KB | ~128MB | 1MB default | Unlimited | Varies |
| **Throughput** | ~10K/s | ~5K/s | ~15K/s | ~3K/s/queue | ~20K/s | ~100K+/s | ~5K/s | ~1K/s |
| **Ordering** | FIFO per queue | FIFO per queue | FIFO per queue | FIFO (option) | Per queue | Per partition | Per workflow | Per function |
| **Cost** | Redis hosting | Broker hosting | Redis hosting | $0.40/M requests | Self-host | Self-host | Self-host/$$ | From $0/mo |

## When to Use Each

### BullMQ — The Default Choice for Node.js

**Choose when:**
- Building with Node.js/TypeScript
- Need priority queues, rate limiting, job dependencies
- Already using Redis
- Team size: 1-20 developers
- Throughput: up to 10K jobs/second

**Skip when:**
- Not using Node.js
- Need exactly-once processing
- Need multi-language support
- Throughput > 50K jobs/second

### Celery — The Python Standard

**Choose when:**
- Building with Python/Django/Flask
- Need task chaining (Canvas)
- Need periodic tasks (Celery Beat)

**Skip when:**
- Not using Python
- Simple scheduling needs (use APScheduler instead)

### AWS SQS — Serverless & Managed

**Choose when:**
- Already on AWS
- Want zero maintenance
- Building with Lambda
- Need guaranteed durability
- Don't need priority queues or job dependencies

**Skip when:**
- Need < 1 second latency (SQS has inherent polling delay)
- Need priority queues
- Need delayed jobs > 15 minutes
- Want to avoid vendor lock-in

### RabbitMQ — Complex Routing

**Choose when:**
- Need complex routing (topic exchanges, headers routing)
- Multi-language environment
- Need message acknowledgment patterns
- Team has RabbitMQ expertise

**Skip when:**
- Simple queue needs (BullMQ is simpler)
- Don't want to manage a broker
- Need job scheduling/cron

### Kafka — High-Throughput Event Streaming

**Choose when:**
- Event-driven architecture (not task queues)
- Throughput > 100K events/second
- Need event replay/reprocessing
- Multiple consumers per event
- Need exactly-once semantics

**Skip when:**
- Simple background jobs (massive overkill)
- Team < 5 developers (operational complexity)
- Need priority queues or job scheduling

### Temporal — Complex Workflows

**Choose when:**
- Long-running workflows (hours, days, weeks)
- Need saga pattern with compensation
- Complex state machines
- Need workflow versioning
- Mission-critical business processes

**Skip when:**
- Simple fire-and-forget jobs
- Budget-conscious (Temporal Cloud is expensive)
- Team unfamiliar with workflow concepts

### Inngest — Serverless Functions

**Choose when:**
- Serverless architecture (Vercel, Netlify, etc.)
- Want managed infrastructure
- Need step functions with automatic retry
- Building event-driven systems
- Small to medium scale

**Skip when:**
- Need self-hosted solution
- High throughput (> 1K/s sustained)
- Need fine-grained queue control

## Cost Comparison (Monthly)

### Self-Hosted (10K jobs/day)

| Solution | Infrastructure | Operational Cost | Total |
|----------|---------------|-----------------|-------|
| BullMQ + Redis | Redis: $15-50/mo | Low (familiar) | ~$30-50 |
| Celery + Redis | Redis: $15-50/mo | Low (familiar) | ~$30-50 |
| RabbitMQ | VM: $20-50/mo | Medium (tuning) | ~$40-80 |
| Kafka | 3+ VMs: $100+/mo | High (expertise) | ~$200+ |
| Temporal | VM: $50+/mo | Medium | ~$100+ |

### Managed Services (10K jobs/day)

| Service | Plan | Monthly Cost |
|---------|------|-------------|
| AWS SQS | Pay-per-use | ~$5-10 |
| AWS SQS + Lambda | Pay-per-use | ~$10-20 |
| CloudAMQP (RabbitMQ) | Lemur (free) to Tiger ($99) | $0-99 |
| Confluent Cloud (Kafka) | Basic | ~$200+ |
| Temporal Cloud | Standard | ~$200+ |
| Inngest | Free tier (10K runs) | $0-25 |
| Upstash Redis (BullMQ) | Pay-per-use | ~$5-15 |

### At Scale (1M jobs/day)

| Solution | Monthly Cost | Notes |
|----------|-------------|-------|
| BullMQ + Upstash | ~$100-200 | Simple, scalable |
| BullMQ + Redis cluster | ~$300-500 | Self-managed |
| AWS SQS + Lambda | ~$200-400 | Fully managed |
| Kafka (Confluent) | ~$1,000+ | Overkill unless streaming |
| Temporal Cloud | ~$500-1,000 | Worth it for complex workflows |

## Migration Paths

### node-cron → BullMQ (Most Common Upgrade)

```typescript
// Before: node-cron (in-process)
cron.schedule('0 2 * * *', () => cleanup());

// After: BullMQ repeatable (distributed)
await queue.add('cleanup', {}, {
  repeat: { pattern: '0 2 * * *' },
});
```

### SQS → BullMQ (Moving Off AWS)

```typescript
// Before: SQS consumer
const command = new ReceiveMessageCommand({ QueueUrl, MaxNumberOfMessages: 10 });
const messages = await sqs.send(command);

// After: BullMQ worker
new Worker('queue', async (job) => {
  // Same handler logic, but with built-in retry, priority, etc.
});
```

### BullMQ → Temporal (Scaling Up Complexity)

Temporal when you need: saga compensation, human-in-the-loop steps, workflow versioning, or multi-day processes. Stay with BullMQ for simple fire-and-forget jobs.

## Redis Sizing Guide for BullMQ

| Jobs/Day | Redis Memory | Redis Plan | Notes |
|----------|-------------|-----------|-------|
| 1K | ~50MB | Mini ($5/mo) | Development |
| 10K | ~200MB | Basic ($15/mo) | Small production |
| 100K | ~1GB | Standard ($50/mo) | Medium production |
| 1M | ~5GB | Premium ($200/mo) | Large production |
| 10M | ~20GB+ | Cluster ($500+/mo) | Enterprise |

Memory depends on job data size. Keep job payloads small (< 1KB) and store large data externally (S3, database).
