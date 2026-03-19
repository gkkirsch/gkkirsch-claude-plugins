# MongoDB Operations Reference

> Complete MongoDB operational reference covering aggregation stages, index management,
> replica sets, sharding, backup/restore, and production administration.
> Covers MongoDB 7.x with mongosh syntax.

---

## Aggregation Pipeline Stages

### Data Flow Stages

| Stage | Purpose | Example |
|-------|---------|---------|
| `$match` | Filter documents (like find) | `{ $match: { status: "active" } }` |
| `$project` | Reshape documents (include/exclude/compute) | `{ $project: { name: 1, total: { $multiply: ["$price", "$qty"] } } }` |
| `$addFields` | Add new fields (keeps existing) | `{ $addFields: { fullName: { $concat: ["$first", " ", "$last"] } } }` |
| `$set` | Alias for $addFields | Same as $addFields |
| `$unset` | Remove fields | `{ $unset: ["internal_id", "debug_info"] }` |
| `$replaceRoot` | Replace document with subdocument | `{ $replaceRoot: { newRoot: "$address" } }` |
| `$replaceWith` | Alias for $replaceRoot | Same as $replaceRoot |

### Grouping & Reshaping

| Stage | Purpose | Example |
|-------|---------|---------|
| `$group` | Group by key, compute aggregates | `{ $group: { _id: "$category", total: { $sum: "$amount" } } }` |
| `$unwind` | Deconstruct array into documents | `{ $unwind: { path: "$items", preserveNullAndEmptyArrays: true } }` |
| `$bucket` | Categorize into fixed ranges | `{ $bucket: { groupBy: "$price", boundaries: [0, 50, 100, 500], default: "other" } }` |
| `$bucketAuto` | Auto-compute bucket boundaries | `{ $bucketAuto: { groupBy: "$price", buckets: 5 } }` |
| `$facet` | Parallel sub-pipelines | See faceted search example below |

### Sorting, Limiting, Sampling

| Stage | Purpose | Example |
|-------|---------|---------|
| `$sort` | Sort documents | `{ $sort: { created_at: -1, _id: 1 } }` |
| `$limit` | Limit output count | `{ $limit: 20 }` |
| `$skip` | Skip N documents | `{ $skip: 40 }` |
| `$sample` | Random sample | `{ $sample: { size: 10 } }` |
| `$sortByCount` | Group, count, and sort | `{ $sortByCount: "$category" }` |

### Joins & Lookups

| Stage | Purpose | Example |
|-------|---------|---------|
| `$lookup` | Left outer join | See below |
| `$graphLookup` | Recursive lookup (tree/graph) | See below |
| `$unionWith` | Union with another collection | `{ $unionWith: { coll: "archive_orders" } }` |

### Window Functions (5.0+)

| Stage | Purpose | Example |
|-------|---------|---------|
| `$setWindowFields` | Window functions | See below |
| `$densify` | Fill gaps in time series | `{ $densify: { field: "date", range: { step: 1, unit: "day" } } }` |
| `$fill` | Fill null values | `{ $fill: { sortBy: { date: 1 }, output: { price: { method: "locf" } } } }` |

### Output Stages

| Stage | Purpose | Example |
|-------|---------|---------|
| `$out` | Write results to collection (replaces) | `{ $out: "summary_collection" }` |
| `$merge` | Upsert results into collection | `{ $merge: { into: "summary", on: "_id", whenMatched: "merge" } }` |
| `$count` | Count documents | `{ $count: "total" }` |

### Aggregation Expressions Cheat Sheet

```javascript
// Arithmetic
{ $add: [a, b] }        { $subtract: [a, b] }
{ $multiply: [a, b] }   { $divide: [a, b] }
{ $mod: [a, b] }        { $abs: a }
{ $ceil: a }             { $floor: a }
{ $round: [a, places] }

// String
{ $concat: [a, b] }     { $substr: [str, start, len] }
{ $toLower: a }          { $toUpper: a }
{ $trim: { input: a } } { $split: [str, delimiter] }
{ $regexMatch: { input: a, regex: /pattern/ } }
{ $replaceAll: { input: a, find: "old", replacement: "new" } }

// Date
{ $dateToString: { format: "%Y-%m-%d", date: a } }
{ $dateFromString: { dateString: a } }
{ $dateDiff: { startDate: a, endDate: b, unit: "day" } }
{ $dateAdd: { startDate: a, unit: "day", amount: 7 } }
{ $year: a } { $month: a } { $dayOfMonth: a }
{ $hour: a } { $minute: a } { $second: a }
{ $dayOfWeek: a } { $dayOfYear: a } { $week: a }

// Array
{ $size: a }             { $arrayElemAt: [a, idx] }
{ $first: a }            { $last: a }
{ $slice: [a, n] }       { $reverseArray: a }
{ $concatArrays: [a, b] }
{ $filter: { input: a, cond: { $gte: ["$$this", 5] } } }
{ $map: { input: a, in: { $multiply: ["$$this", 2] } } }
{ $reduce: { input: a, initialValue: 0, in: { $add: ["$$value", "$$this"] } } }
{ $in: [val, array] }    { $setIntersection: [a, b] }
{ $setUnion: [a, b] }    { $setDifference: [a, b] }

// Conditional
{ $cond: { if: cond, then: a, else: b } }
{ $ifNull: [a, default] }
{ $switch: { branches: [{ case: cond, then: val }], default: val } }

// Type conversion
{ $toString: a }   { $toInt: a }    { $toDouble: a }
{ $toObjectId: a } { $toBool: a }   { $toDate: a }
{ $convert: { input: a, to: "int", onError: 0 } }

// Accumulator operators (in $group)
{ $sum: a }      { $avg: a }      { $min: a }      { $max: a }
{ $first: a }    { $last: a }     { $push: a }     { $addToSet: a }
{ $count: {} }   { $mergeObjects: a }
{ $stdDevPop: a } { $stdDevSamp: a }
{ $top: { sortBy: {s: -1}, output: a } }    // MongoDB 5.2+
{ $bottom: { sortBy: {s: 1}, output: a } }
{ $topN: { n: 5, sortBy: {s: -1}, output: a } }
{ $bottomN: { n: 5, sortBy: {s: 1}, output: a } }
{ $firstN: { n: 5, input: a } }
{ $lastN: { n: 5, input: a } }
{ $maxN: { n: 5, input: a } }
{ $minN: { n: 5, input: a } }
```

### Lookup Patterns

```javascript
// Basic lookup (1:N)
{
  $lookup: {
    from: "orders",
    localField: "_id",
    foreignField: "user_id",
    as: "user_orders"
  }
}

// Pipeline lookup (filtered, projected)
{
  $lookup: {
    from: "orders",
    let: { userId: "$_id" },
    pipeline: [
      { $match: { $expr: { $eq: ["$user_id", "$$userId"] } } },
      { $match: { status: { $ne: "cancelled" } } },
      { $sort: { created_at: -1 } },
      { $limit: 5 },
      { $project: { order_number: 1, total: 1, status: 1 } }
    ],
    as: "recent_orders"
  }
}

// Graph lookup (recursive — org chart, categories)
{
  $graphLookup: {
    from: "employees",
    startWith: "$manager_id",
    connectFromField: "manager_id",
    connectToField: "_id",
    as: "management_chain",
    maxDepth: 5,
    depthField: "level"
  }
}
```

### Faceted Search

```javascript
db.products.aggregate([
  { $match: { $text: { $search: "wireless mouse" } } },
  {
    $facet: {
      // Search results (paginated)
      results: [
        { $sort: { score: { $meta: "textScore" } } },
        { $skip: 0 },
        { $limit: 20 },
        { $project: { name: 1, price: 1, category: 1, image: 1 } }
      ],
      // Category counts
      categories: [
        { $group: { _id: "$category.primary", count: { $sum: 1 } } },
        { $sort: { count: -1 } }
      ],
      // Price ranges
      priceRanges: [
        {
          $bucket: {
            groupBy: "$pricing.base_price",
            boundaries: [0, 25, 50, 100, 250, 1000],
            default: "1000+",
            output: { count: { $sum: 1 } }
          }
        }
      ],
      // Total count
      totalCount: [{ $count: "total" }]
    }
  }
]);
```

### Window Functions

```javascript
// Running total + moving average
db.daily_sales.aggregate([
  { $sort: { date: 1 } },
  {
    $setWindowFields: {
      partitionBy: "$region",
      sortBy: { date: 1 },
      output: {
        cumulative_revenue: {
          $sum: "$revenue",
          window: { documents: ["unbounded", "current"] }
        },
        moving_avg_7d: {
          $avg: "$revenue",
          window: { range: [-6, "current"], unit: "day" }
        },
        rank_in_region: {
          $rank: {}
        },
        dense_rank: {
          $denseRank: {}
        },
        percentile: {
          $percentile: { p: [0.5, 0.95], input: "$revenue", method: "approximate" }
        }
      }
    }
  }
]);
```

---

## Index Management

### Index Types

```javascript
// Single field
db.users.createIndex({ email: 1 });                   // Ascending
db.users.createIndex({ created_at: -1 });              // Descending

// Compound (follow ESR rule: Equality, Sort, Range)
db.orders.createIndex({ status: 1, created_at: -1, total: 1 });

// Multikey (array fields — auto-detected)
db.products.createIndex({ tags: 1 });

// Text
db.articles.createIndex(
  { title: "text", body: "text" },
  { weights: { title: 10, body: 1 }, default_language: "english" }
);

// Geospatial
db.stores.createIndex({ location: "2dsphere" });       // GeoJSON
db.legacy.createIndex({ coordinates: "2d" });          // Legacy flat

// Hashed (for sharding)
db.users.createIndex({ user_id: "hashed" });

// Wildcard (dynamic field names)
db.products.createIndex({ "attributes.$**": 1 });      // All under attributes
db.events.createIndex({ "$**": 1 });                   // All fields (careful!)
```

### Index Options

```javascript
db.collection.createIndex(
  { email: 1 },
  {
    unique: true,                   // Unique constraint
    sparse: true,                   // Only index docs where field exists
    background: true,               // Build in background (deprecated in 4.2+, always bg now)
    name: "idx_email_unique",       // Custom name
    expireAfterSeconds: 86400,      // TTL index (auto-delete after 24h)
    partialFilterExpression: {      // Partial index (index subset of docs)
      status: "active"
    },
    collation: {                    // Case-insensitive
      locale: "en",
      strength: 2
    },
    hidden: true                    // Create hidden (doesn't affect queries until unhidden)
  }
);
```

### Index Analysis

```javascript
// Explain a query
db.orders.find({ status: "shipped" }).sort({ created_at: -1 }).explain("executionStats");

// Key fields in explain output:
// - queryPlanner.winningPlan.stage:
//   "IXSCAN" = using index (good)
//   "COLLSCAN" = full collection scan (bad)
//   "FETCH" = fetching documents after index scan
//   "SORT" = in-memory sort (bad if large)
//
// - executionStats:
//   totalDocsExamined vs nReturned — ratio should be close to 1:1
//   executionTimeMillis — wall clock time
//   totalKeysExamined — index keys scanned

// List all indexes
db.collection.getIndexes();

// Index size
db.collection.stats().indexSizes;
db.collection.totalIndexSize();

// Index usage stats
db.collection.aggregate([{ $indexStats: {} }]);

// Drop index
db.collection.dropIndex("index_name");
db.collection.dropIndex({ field: 1 });

// Hide/unhide index (test impact without dropping)
db.collection.hideIndex("index_name");
db.collection.unhideIndex("index_name");
```

### ESR Rule (Equality, Sort, Range)

```
Compound index field order:
1. EQUALITY — fields tested with exact match (=, $eq, $in)
2. SORT — fields in the sort specification
3. RANGE — fields with range conditions ($gt, $lt, $gte, $lte)

Example query:
  db.orders.find({ status: "active", total: { $gte: 100 } }).sort({ created_at: -1 })

Optimal index:
  { status: 1, created_at: -1, total: 1 }
  (Equality: status, Sort: created_at, Range: total)

Bad index:
  { total: 1, status: 1, created_at: -1 }
  (Range first = can't use rest of index efficiently)
```

---

## Replica Set Operations

### Setup & Configuration

```javascript
// Initiate replica set
rs.initiate({
  _id: "myReplicaSet",
  members: [
    { _id: 0, host: "mongo1:27017", priority: 2 },    // Preferred primary
    { _id: 1, host: "mongo2:27017", priority: 1 },
    { _id: 2, host: "mongo3:27017", priority: 1 },
  ]
});

// Add members
rs.add("mongo4:27017");
rs.add({ host: "mongo5:27017", priority: 0, hidden: true });  // Hidden member (backups)
rs.addArb("arbiter:27017");  // Arbiter (voting only, no data)

// Remove member
rs.remove("mongo4:27017");

// Status
rs.status();          // Detailed status of all members
rs.conf();            // Current configuration
rs.isMaster();        // Am I the primary?
rs.printReplicationInfo();    // Oplog status
rs.printSecondaryReplicationInfo();  // Replication lag

// Step down (force election)
rs.stepDown(60);      // Step down for 60 seconds

// Reconfigure
cfg = rs.conf();
cfg.members[1].priority = 2;
rs.reconfig(cfg);
```

### Read Preferences

```javascript
// Read preference options:
// primary           — Only primary (default). Strongest consistency.
// primaryPreferred   — Primary if available, otherwise secondary.
// secondary          — Only secondaries. Eventual consistency.
// secondaryPreferred — Secondary if available, otherwise primary.
// nearest            — Lowest latency member.

// mongosh
db.collection.find().readPref("secondaryPreferred");

// Connection string
"mongodb://host1,host2,host3/?replicaSet=myRS&readPreference=secondaryPreferred"

// Node.js driver
const collection = db.collection("users");
const cursor = collection.find({}, {
  readPreference: ReadPreference.SECONDARY_PREFERRED
});
```

### Write Concerns

```javascript
// Write concern options:
// w: 1            — Acknowledged by primary (default)
// w: "majority"   — Acknowledged by majority of members
// w: 0            — Unacknowledged (fire and forget)
// w: <number>     — Acknowledged by N members
// j: true         — Written to journal (durable)
// wtimeout: ms    — Timeout for write concern

db.collection.insertOne(doc, {
  writeConcern: { w: "majority", j: true, wtimeout: 5000 }
});

// Set default write concern
db.adminCommand({
  setDefaultRWConcern: 1,
  defaultWriteConcern: { w: "majority" }
});
```

---

## Sharding

### Setup

```javascript
// Enable sharding on database
sh.enableSharding("mydb");

// Shard a collection
sh.shardCollection("mydb.orders", { customer_id: "hashed" });   // Hashed
sh.shardCollection("mydb.logs", { timestamp: 1 });                // Ranged
sh.shardCollection("mydb.events", { tenant_id: 1, _id: 1 });   // Compound

// Shard key selection rules:
// 1. High cardinality (many unique values)
// 2. Even distribution (no hot spots)
// 3. Query isolation (most queries include shard key)
// 4. Monotonically increasing values are BAD for ranged sharding (hot shard)
//    Use hashed sharding for such fields
```

### Monitoring

```javascript
sh.status();                         // Cluster overview
db.collection.getShardDistribution(); // Data distribution across shards
db.adminCommand({ balancerStatus: 1 }); // Balancer status
sh.isBalancerRunning();              // Is balancer active?

// Chunk management
sh.splitAt("mydb.orders", { customer_id: "C-50000" }); // Manual split
sh.moveChunk("mydb.orders", { customer_id: "C-50000" }, "shard0002"); // Move chunk

// Zone sharding (geographic data placement)
sh.addShardTag("shard0001", "US");
sh.addTagRange("mydb.orders", { region: "US" }, { region: "US\xff" }, "US");
```

---

## Backup & Restore

### mongodump / mongorestore

```bash
# Full backup
mongodump --uri="mongodb://user:pass@host:27017/mydb" --out=/backup/$(date +%Y%m%d)

# Specific collection
mongodump --db=mydb --collection=orders --out=/backup/orders

# With compression
mongodump --uri="mongodb://host:27017" --gzip --archive=/backup/full.gz

# Restore
mongorestore --uri="mongodb://host:27017" /backup/20240319/
mongorestore --uri="mongodb://host:27017" --gzip --archive=/backup/full.gz
mongorestore --db=mydb --collection=orders /backup/orders/mydb/orders.bson

# Restore to different database
mongorestore --nsFrom="mydb.*" --nsTo="mydb_copy.*" /backup/mydb/

# Options
--oplog                  # Include oplog for point-in-time restore
--drop                   # Drop collections before restoring
--numParallelCollections # Parallel restore (default: 4)
--numInsertionWorkersPerCollection  # Writers per collection
```

### mongoexport / mongoimport

```bash
# Export to JSON
mongoexport --db=mydb --collection=users --out=users.json

# Export to CSV
mongoexport --db=mydb --collection=users --type=csv --fields=name,email --out=users.csv

# Export with query
mongoexport --db=mydb --collection=orders --query='{"status":"shipped"}' --out=shipped.json

# Import JSON
mongoimport --db=mydb --collection=users --file=users.json

# Import CSV
mongoimport --db=mydb --collection=users --type=csv --headerline --file=users.csv

# Upsert import
mongoimport --db=mydb --collection=users --mode=upsert --upsertFields=email --file=users.json
```

---

## User & Role Management

```javascript
// Create admin user
db.createUser({
  user: "admin",
  pwd: passwordPrompt(),
  roles: [
    { role: "userAdminAnyDatabase", db: "admin" },
    { role: "readWriteAnyDatabase", db: "admin" },
    { role: "dbAdminAnyDatabase", db: "admin" },
    { role: "clusterAdmin", db: "admin" }
  ]
});

// Create application user (least privilege)
db.createUser({
  user: "app_user",
  pwd: "secure_password",
  roles: [
    { role: "readWrite", db: "myapp" }
  ]
});

// Create read-only user
db.createUser({
  user: "analyst",
  pwd: "secure_password",
  roles: [
    { role: "read", db: "myapp" },
    { role: "read", db: "analytics" }
  ]
});

// Built-in roles:
// read              — Read all non-system collections
// readWrite         — Read + write
// dbAdmin           — Schema management, indexing, stats
// dbOwner           — readWrite + dbAdmin + userAdmin
// userAdmin         — Create/modify users and roles
// clusterAdmin      — Cluster operations
// backup / restore  — Backup/restore operations
// root              — Everything (superuser)

// Custom role
db.createRole({
  role: "orderManager",
  privileges: [
    {
      resource: { db: "myapp", collection: "orders" },
      actions: ["find", "update", "insert"]
    },
    {
      resource: { db: "myapp", collection: "products" },
      actions: ["find"]  // Read-only on products
    }
  ],
  roles: []  // No inherited roles
});

// Manage users
db.getUsers();
db.changeUserPassword("app_user", "new_password");
db.grantRolesToUser("app_user", [{ role: "dbAdmin", db: "myapp" }]);
db.revokeRolesFromUser("app_user", [{ role: "dbAdmin", db: "myapp" }]);
db.dropUser("old_user");
```

---

## Monitoring & Diagnostics

```javascript
// Server status
db.serverStatus();                   // Comprehensive server stats
db.serverStatus().connections;       // Connection counts
db.serverStatus().opcounters;        // Operation counts
db.serverStatus().wiredTiger;        // Storage engine stats

// Collection stats
db.collection.stats();               // Collection statistics
db.collection.storageSize();         // Data size on disk
db.collection.totalIndexSize();      // Total index size
db.collection.dataSize();            // Uncompressed data size

// Current operations
db.currentOp();                      // All current operations
db.currentOp({ "active": true, "secs_running": { "$gt": 5 } });  // Slow ops
db.killOp(opId);                     // Kill operation

// Profiler (query logging)
db.setProfilingLevel(1, { slowms: 100 });  // Log queries > 100ms
db.setProfilingLevel(2);             // Log ALL queries (debug only!)
db.setProfilingLevel(0);             // Disable
db.system.profile.find().sort({ ts: -1 }).limit(10);  // View logged queries

// Connection string diagnostics
db.adminCommand({ getLog: "global" });     // Server log
db.adminCommand({ connPoolStats: 1 });     // Connection pool stats
db.adminCommand({ hostInfo: 1 });          // Host information
db.adminCommand({ buildInfo: 1 });         // Build/version info

// Top operations per collection
db.adminCommand({ top: 1 });
```

### Key Metrics to Monitor

```
Connection metrics:
  db.serverStatus().connections.current     -- Active connections
  db.serverStatus().connections.available   -- Available connections

Operation metrics:
  db.serverStatus().opcounters.query       -- Total queries
  db.serverStatus().opcounters.insert      -- Total inserts
  db.serverStatus().opcounters.update      -- Total updates
  db.serverStatus().opcounters.delete      -- Total deletes

Replication metrics:
  rs.printSecondaryReplicationInfo()       -- Replication lag
  db.serverStatus().repl                   -- Replication state

Cache metrics (WiredTiger):
  db.serverStatus().wiredTiger.cache["bytes currently in the cache"]
  db.serverStatus().wiredTiger.cache["pages read into cache"]
  db.serverStatus().wiredTiger.cache["pages written from cache"]

Lock metrics:
  db.serverStatus().globalLock.activeClients
  db.serverStatus().globalLock.currentQueue

Document metrics:
  db.serverStatus().metrics.document.inserted
  db.serverStatus().metrics.document.updated
  db.serverStatus().metrics.document.deleted
  db.serverStatus().metrics.document.returned
```

---

## mongosh Quick Reference

```bash
# Connect
mongosh                                      # localhost:27017
mongosh "mongodb://user:pass@host:27017/db"
mongosh --host host --port 27017 -u user -p pass --authenticationDatabase admin
mongosh "mongodb+srv://user:pass@cluster.mongodb.net/db"  # Atlas

# Database operations
show dbs                             # List databases
use mydb                             # Switch database
show collections                     # List collections
db.dropDatabase()                    # Drop current database

# Collection operations
db.createCollection("logs", {
  timeseries: {                      # Time series collection
    timeField: "timestamp",
    metaField: "source",
    granularity: "minutes"
  },
  expireAfterSeconds: 86400 * 30     # Auto-delete after 30 days
});

db.createCollection("orders", {
  changeStreamPreAndPostImages: { enabled: true }  # For change stream full doc
});

db.collection.renameCollection("new_name");
db.collection.drop();

# Useful one-liners
db.collection.estimatedDocumentCount();        # Fast approximate count
db.collection.countDocuments({ status: "active" });  # Exact filtered count
db.collection.distinct("category");            # Unique values
db.collection.find().sort({ _id: -1 }).limit(1);  # Latest document
```
