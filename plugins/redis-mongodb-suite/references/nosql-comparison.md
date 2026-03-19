# NoSQL Comparison Reference

Decision guide for choosing between Redis, MongoDB, and other data stores based on workload characteristics.

---

## Redis vs MongoDB — When to Use Each

### Redis

```
USE REDIS WHEN:
  - Sub-millisecond latency required
  - Data fits in memory
  - Caching layer for another database
  - Real-time features (leaderboards, sessions, rate limiting)
  - Pub/Sub messaging or event streaming
  - Temporary data with TTL
  - Atomic counters and distributed locks
  - Simple data structures (strings, hashes, lists, sets, sorted sets)

DO NOT USE REDIS AS:
  - Primary database for large datasets (memory-bound)
  - Complex querying engine (no ad-hoc queries)
  - Document store with nested objects (flat hash only)
  - Relational data with JOINs
  - Full-text search engine (basic text only)
```

### MongoDB

```
USE MONGODB WHEN:
  - Flexible/evolving schema
  - Document-oriented data (JSON-like)
  - Complex queries and aggregation pipelines
  - Full-text search
  - Geospatial queries
  - Data larger than available memory
  - Need for secondary indexes on any field
  - Horizontal scaling (sharding)
  - Change streams for real-time updates

DO NOT USE MONGODB AS:
  - Cache layer (use Redis)
  - Message queue (use Redis Streams, RabbitMQ, or Kafka)
  - Time-series primary store (use TimescaleDB or InfluxDB)
  - Graph database (use Neo4j)
  - Highly relational data with many JOINs (use PostgreSQL)
```

### Side-by-Side Comparison

| Feature | Redis | MongoDB |
|---------|-------|---------|
| **Data model** | Key-value + data structures | Document (BSON/JSON) |
| **Storage** | In-memory (+ disk persistence) | Disk (+ memory-mapped) |
| **Max data size** | Limited by RAM | Limited by disk |
| **Query language** | Commands (GET, SET, ZADD...) | MQL (find, aggregate) |
| **Secondary indexes** | No (key-based only) | Yes (any field) |
| **Transactions** | Lua scripts, MULTI/EXEC | Multi-document ACID |
| **Schema** | Schema-free | Schema-free + optional validation |
| **Replication** | Master-replica | Replica sets (automatic failover) |
| **Sharding** | Redis Cluster (hash slots) | Range/hash/zone sharding |
| **Latency** | Sub-millisecond | 1-10ms typical |
| **Throughput** | 100K+ ops/sec per node | 10K-50K ops/sec per node |
| **Best for** | Cache, sessions, real-time | Application database, analytics |

---

## Broader NoSQL Landscape

### Document Stores

| Database | Best For | Trade-offs |
|----------|----------|------------|
| **MongoDB** | General-purpose documents, flexible schema | Memory-hungry for working set, $lookup is slow |
| **CouchDB** | Offline-first, multi-master replication | Slower queries, eventual consistency |
| **Amazon DocumentDB** | MongoDB-compatible on AWS | Not fully MongoDB-compatible, vendor lock-in |
| **Azure Cosmos DB** | Multi-model, global distribution | Expensive, complex pricing (RU-based) |

### Key-Value Stores

| Database | Best For | Trade-offs |
|----------|----------|------------|
| **Redis** | Caching, real-time, data structures | Memory-bound, limited query |
| **Memcached** | Simple string caching | No persistence, no data structures |
| **Amazon DynamoDB** | Serverless key-value at scale | Complex pricing, limited query patterns |
| **etcd** | Distributed config, service discovery | Small data only (< 8GB recommended) |

### Wide-Column Stores

| Database | Best For | Trade-offs |
|----------|----------|------------|
| **Apache Cassandra** | Write-heavy, time-series, IoT | Complex data modeling, eventual consistency |
| **ScyllaDB** | Cassandra-compatible, higher performance | Newer, smaller community |
| **HBase** | Hadoop ecosystem, large scans | Complex ops, needs Hadoop/HDFS |
| **Google Bigtable** | Massive scale, time-series | GCP only, expensive |

### Graph Databases

| Database | Best For | Trade-offs |
|----------|----------|------------|
| **Neo4j** | Social networks, recommendations, fraud | Scaling limitations, Cypher learning curve |
| **Amazon Neptune** | AWS graph workloads | Vendor lock-in |
| **ArangoDB** | Multi-model (doc + graph + key-value) | Jack of all trades, master of none |

### Search Engines

| Database | Best For | Trade-offs |
|----------|----------|------------|
| **Elasticsearch** | Full-text search, log analytics | Resource-heavy, complex cluster management |
| **OpenSearch** | AWS-managed Elasticsearch fork | Slightly behind Elasticsearch features |
| **Meilisearch** | Simple search, typo-tolerant | Smaller scale, less flexible |
| **Typesense** | Developer-friendly search | Newer, smaller ecosystem |

### Time-Series Databases

| Database | Best For | Trade-offs |
|----------|----------|------------|
| **InfluxDB** | Metrics, IoT, monitoring | Flux query language learning curve |
| **TimescaleDB** | Time-series on PostgreSQL | Requires PostgreSQL expertise |
| **Prometheus** | Metrics collection, alerting | Pull-based, not for long retention |
| **QuestDB** | High-ingest time-series | Newer, smaller community |

### Message Queues (Often Used with NoSQL)

| System | Best For | Trade-offs |
|--------|----------|------------|
| **Apache Kafka** | Event streaming, high throughput | Complex operations, Zookeeper/KRaft |
| **RabbitMQ** | Task queues, routing patterns | Lower throughput than Kafka |
| **Redis Streams** | Lightweight event log | Memory-bound, simpler than Kafka |
| **Amazon SQS** | Serverless message queue | AWS only, higher latency |

---

## Decision Matrix: Choosing the Right Database

### By Access Pattern

| Access Pattern | Best Choice | Why |
|---------------|-------------|-----|
| Key lookup (GET by ID) | Redis or DynamoDB | O(1) lookups |
| Document queries (flexible filters) | MongoDB | Rich query language, indexes |
| Full-text search | Elasticsearch or MongoDB (Atlas Search) | Inverted indexes, relevance scoring |
| Graph traversal | Neo4j | Native graph storage, Cypher |
| Time-series ingestion | InfluxDB or TimescaleDB | Optimized for time-ordered data |
| Wide scans (analytics) | Cassandra or BigQuery | Distributed scan, columnar |
| Real-time streaming | Kafka + Redis | Event sourcing + caching |
| Session storage | Redis | Fast, TTL, atomic operations |
| Leaderboard / ranking | Redis (Sorted Sets) | O(log N) insert and rank |

### By Non-Functional Requirement

| Requirement | Best Choice | Notes |
|-------------|-------------|-------|
| Sub-ms latency | Redis | In-memory, single-threaded |
| Massive write throughput | Cassandra, Kafka | Distributed writes, append-only |
| Global distribution | Cosmos DB, CockroachDB | Built-in geo-replication |
| Strong consistency | PostgreSQL, MongoDB (w:majority) | ACID transactions |
| Eventual consistency OK | Cassandra, DynamoDB | AP in CAP theorem |
| Serverless / zero-ops | DynamoDB, Cosmos DB, PlanetScale | Managed, auto-scaling |
| Open source / self-host | Redis, MongoDB, PostgreSQL | Full control, no vendor lock |
| Cost-sensitive | Redis + PostgreSQL | Free tiers, efficient |

### By Data Characteristics

| Data Type | Recommended | Notes |
|-----------|-------------|-------|
| Structured, relational | PostgreSQL | JOINs, constraints, ACID |
| Semi-structured (JSON) | MongoDB | Flexible schema, rich queries |
| Key-value pairs | Redis or DynamoDB | Simple, fast, scalable |
| Time-series metrics | TimescaleDB or InfluxDB | Compression, retention policies |
| Graph / network data | Neo4j | Relationship-first queries |
| Blobs / files | S3 + metadata in any DB | Object storage for files |
| Search corpus | Elasticsearch | Full-text, facets, relevance |
| Logs / events | Elasticsearch or Kafka | Append-only, searchable |
| Cache / ephemeral | Redis | In-memory, TTL, eviction |
| Configuration | etcd or Consul | Distributed, watchable |

---

## Common Multi-Database Architectures

### Pattern 1: PostgreSQL + Redis

```
PostgreSQL: Primary data store (users, orders, products)
Redis: Cache layer + sessions + rate limiting + queues

Flow:
  Write → PostgreSQL → invalidate Redis cache
  Read → Redis cache → miss → PostgreSQL → populate cache

Most common pattern. Covers 90% of web applications.
```

### Pattern 2: MongoDB + Redis

```
MongoDB: Application database (flexible documents)
Redis: Cache layer + real-time features

Flow:
  Write → MongoDB → invalidate Redis
  Read → Redis → miss → MongoDB → cache in Redis
  Real-time: Redis pub/sub for live updates

Good for: Content management, e-commerce catalogs, user profiles.
```

### Pattern 3: PostgreSQL + MongoDB + Redis

```
PostgreSQL: Transactional data (payments, accounts)
MongoDB: Content/catalog data (products, articles, user-generated)
Redis: Cache, sessions, real-time

Good for: Large platforms with mixed data types.
Anti-pattern warning: Don't use this unless you genuinely need all three.
Operational complexity scales with number of database systems.
```

### Pattern 4: PostgreSQL + Elasticsearch + Redis

```
PostgreSQL: Source of truth
Elasticsearch: Search index (synced via CDC or scheduled)
Redis: Cache, sessions

Flow:
  Write → PostgreSQL → sync to Elasticsearch (async)
  Search → Elasticsearch
  CRUD → PostgreSQL (via Redis cache)

Good for: E-commerce, content platforms needing full-text search.
```

### Pattern 5: Kafka + MongoDB/PostgreSQL + Redis

```
Kafka: Event bus (all writes are events)
MongoDB/PostgreSQL: Materialized views of events
Redis: Cache, real-time features

Flow:
  Write → Kafka event → consumers update database(s) + cache
  Read → Redis cache → miss → database

Good for: Event-driven architectures, microservices, CQRS.
```

---

## Migration Considerations

### Moving from SQL to MongoDB

```
Things that map well:
  - Tables → Collections
  - Rows → Documents
  - Columns → Fields
  - JOINs → Embedded documents or $lookup

Things that don't map well:
  - Complex multi-table JOINs → denormalize into documents
  - Foreign key constraints → application-level enforcement
  - Stored procedures → aggregation pipelines or application code
  - ACID across tables → multi-document transactions (slower)

Rule: If you're doing lots of $lookup, you probably should use SQL.
```

### Moving from Memcached to Redis

```
Direct replacements:
  memcached.set(key, val, ttl)  →  redis.setex(key, ttl, val)
  memcached.get(key)            →  redis.get(key)
  memcached.delete(key)         →  redis.delete(key)
  memcached.incr(key, 1)        →  redis.incr(key)

New capabilities with Redis:
  - Data structures (hashes, lists, sets, sorted sets)
  - Pub/Sub messaging
  - Lua scripting for atomic operations
  - Persistence (RDB + AOF)
  - Streams (event log)
  - Cluster mode (horizontal scaling)

Migration strategy:
  1. Dual-write to both during migration period
  2. Read from Redis, fall back to Memcached
  3. Monitor hit rates on both
  4. Cut over when Redis hit rate matches
  5. Remove Memcached
```

### Adding Redis Cache to Existing Database

```
Phase 1: Read-Through Cache
  - Add Redis GET before database reads
  - On miss, query database, cache result
  - No changes to write path
  - Low risk, immediate benefit

Phase 2: Write-Through Invalidation
  - On database writes, delete corresponding cache keys
  - Prevents stale data
  - Requires identifying which cache keys to invalidate per write

Phase 3: Advanced Patterns
  - Tag-based invalidation for related entities
  - Cache warming for popular content
  - Stampede prevention (mutex/SWR)
  - Monitoring and alerting

Phase 4: Optimization
  - Multi-level cache (L1 in-process + L2 Redis)
  - Pipeline/batch cache operations
  - Compression for large values
  - Key prefix conventions for organized keyspace
```

---

## Operational Complexity Comparison

| Aspect | Redis | MongoDB | PostgreSQL |
|--------|-------|---------|------------|
| **Setup difficulty** | Easy | Moderate | Moderate |
| **Backup** | RDB snapshots, AOF | mongodump, oplog | pg_dump, WAL |
| **Scaling out** | Cluster (manual) | Sharding (built-in) | Read replicas, Citus |
| **Monitoring** | INFO command, slow log | Atlas monitoring, profiler | pg_stat, pgBouncer |
| **Upgrades** | Simple (in-place) | Rolling upgrades | pg_upgrade |
| **Cloud managed** | ElastiCache, Upstash, Redis Cloud | Atlas, DocumentDB | RDS, Supabase, Neon |
| **Community size** | Very large | Large | Very large |
| **Driver support** | All languages | All languages | All languages |
| **Learning curve** | Low (commands) | Medium (MQL + aggregation) | Medium (SQL + extensions) |
| **Operational cost** | Low (simple) | Medium (replica sets) | Medium (vacuuming, tuning) |

---

## Quick Decision Flowchart

```
START: What problem are you solving?

├─ "I need a primary database for my application"
│   ├─ Structured, relational data? → PostgreSQL
│   ├─ Flexible/evolving schema? → MongoDB
│   ├─ Massive write volume (IoT/logs)? → Cassandra or TimescaleDB
│   └─ Serverless, auto-scaling? → DynamoDB or PlanetScale
│
├─ "I need to make my app faster"
│   ├─ Cache database queries? → Redis (cache-aside)
│   ├─ Cache HTTP responses? → CDN + Cache-Control headers
│   ├─ Cache at application level? → Redis + in-process LRU
│   └─ Cache static assets? → CDN (Cloudflare, CloudFront)
│
├─ "I need real-time features"
│   ├─ Live notifications? → Redis Pub/Sub or WebSockets
│   ├─ Leaderboard / ranking? → Redis Sorted Sets
│   ├─ Rate limiting? → Redis (sliding window)
│   ├─ Session management? → Redis (with TTL)
│   └─ Live data sync? → MongoDB Change Streams or Firebase
│
├─ "I need search"
│   ├─ Full-text search? → Elasticsearch or MongoDB Atlas Search
│   ├─ Autocomplete? → Elasticsearch or Meilisearch
│   └─ Faceted search? → Elasticsearch
│
├─ "I need messaging / events"
│   ├─ High-throughput event streaming? → Apache Kafka
│   ├─ Task queue? → Redis Streams or RabbitMQ
│   ├─ Simple pub/sub? → Redis Pub/Sub
│   └─ Serverless queue? → Amazon SQS
│
└─ "I need analytics"
    ├─ Business intelligence? → PostgreSQL or BigQuery
    ├─ Real-time dashboards? → Redis + time-series DB
    ├─ Log analytics? → Elasticsearch + Kibana
    └─ Time-series metrics? → InfluxDB or TimescaleDB
```
