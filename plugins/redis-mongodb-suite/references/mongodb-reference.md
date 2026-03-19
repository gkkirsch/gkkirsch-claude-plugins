# MongoDB Reference

Comprehensive reference for MongoDB operations, aggregation framework, indexing, and administration.

---

## CRUD Operations Quick Reference

### Create

```javascript
// Insert single document
db.users.insertOne({
  name: "Alice",
  email: "alice@example.com",
  created_at: new Date()
});

// Insert multiple documents
db.users.insertMany([
  { name: "Bob", email: "bob@example.com" },
  { name: "Charlie", email: "charlie@example.com" }
], { ordered: false });  // Continue on duplicate key errors

// Upsert (insert if not exists, update if exists)
db.users.updateOne(
  { email: "alice@example.com" },
  { $set: { name: "Alice Smith", updated_at: new Date() } },
  { upsert: true }
);
```

### Read

```javascript
// Find with filter, projection, sort, limit
db.products.find(
  { category: "electronics", price: { $gte: 50, $lte: 200 } },  // Filter
  { name: 1, price: 1, _id: 0 }                                  // Projection
).sort({ price: 1 }).limit(20);

// FindOne
db.users.findOne({ email: "alice@example.com" });

// Count
db.orders.countDocuments({ status: "pending" });           // Exact (uses index)
db.orders.estimatedDocumentCount();                         // Fast estimate (metadata)

// Distinct values
db.products.distinct("category");                           // All unique categories
db.products.distinct("category", { price: { $gt: 100 } }); // With filter
```

### Update

```javascript
// Update one
db.users.updateOne(
  { _id: ObjectId("...") },
  {
    $set: { "profile.bio": "Updated bio", updated_at: new Date() },
    $inc: { login_count: 1 },
    $push: { tags: "verified" }
  }
);

// Update many
db.products.updateMany(
  { category: "electronics" },
  { $mul: { price: 0.9 } }  // 10% discount
);

// FindOneAndUpdate (returns document)
const result = db.orders.findOneAndUpdate(
  { _id: orderId, status: "pending" },
  { $set: { status: "processing" } },
  { returnDocument: "after" }          // Return updated doc
);

// Array update with arrayFilters
db.orders.updateOne(
  { _id: orderId },
  { $set: { "items.$[elem].shipped": true } },
  { arrayFilters: [{ "elem.product_id": productId }] }
);
```

### Delete

```javascript
// Delete one
db.users.deleteOne({ _id: ObjectId("...") });

// Delete many
db.sessions.deleteMany({ expires_at: { $lt: new Date() } });

// FindOneAndDelete (returns deleted doc)
const deleted = db.queue.findOneAndDelete(
  { status: "pending" },
  { sort: { priority: -1, created_at: 1 } }  // Highest priority, oldest first
);
```

---

## Query Operators Reference

### Comparison

| Operator | Meaning | Example |
|----------|---------|---------|
| `$eq` | Equal | `{ age: { $eq: 25 } }` |
| `$ne` | Not equal | `{ status: { $ne: "deleted" } }` |
| `$gt` | Greater than | `{ price: { $gt: 100 } }` |
| `$gte` | Greater than or equal | `{ age: { $gte: 18 } }` |
| `$lt` | Less than | `{ stock: { $lt: 10 } }` |
| `$lte` | Less than or equal | `{ priority: { $lte: 3 } }` |
| `$in` | In array | `{ status: { $in: ["active", "pending"] } }` |
| `$nin` | Not in array | `{ role: { $nin: ["admin", "superadmin"] } }` |

### Logical

| Operator | Meaning | Example |
|----------|---------|---------|
| `$and` | All conditions | `{ $and: [{ price: { $gt: 10 } }, { price: { $lt: 100 } }] }` |
| `$or` | Any condition | `{ $or: [{ status: "active" }, { featured: true }] }` |
| `$not` | Negate | `{ price: { $not: { $gt: 100 } } }` |
| `$nor` | None of conditions | `{ $nor: [{ deleted: true }, { archived: true }] }` |

### Element

| Operator | Meaning | Example |
|----------|---------|---------|
| `$exists` | Field exists | `{ phone: { $exists: true } }` |
| `$type` | Field type | `{ age: { $type: "int" } }` |

### Array

| Operator | Meaning | Example |
|----------|---------|---------|
| `$all` | Contains all values | `{ tags: { $all: ["js", "node"] } }` |
| `$elemMatch` | Element matches all conditions | `{ scores: { $elemMatch: { $gte: 80, $lt: 90 } } }` |
| `$size` | Array length | `{ tags: { $size: 3 } }` |

### Evaluation

| Operator | Meaning | Example |
|----------|---------|---------|
| `$regex` | Regular expression | `{ name: { $regex: /^alice/i } }` |
| `$text` | Text search | `{ $text: { $search: "mongodb tutorial" } }` |
| `$expr` | Aggregation expression in find | `{ $expr: { $gt: ["$qty", "$ordered"] } }` |
| `$mod` | Modulo | `{ qty: { $mod: [4, 0] } }` (divisible by 4) |

---

## Update Operators Reference

### Field Operators

| Operator | Example |
|----------|---------|
| `$set` | `{ $set: { "name": "Alice", "address.city": "NYC" } }` |
| `$unset` | `{ $unset: { "deprecated_field": "" } }` |
| `$rename` | `{ $rename: { "old_name": "new_name" } }` |
| `$setOnInsert` | `{ $setOnInsert: { created_at: new Date() } }` (only on upsert insert) |
| `$currentDate` | `{ $currentDate: { updated_at: true } }` |

### Numeric Operators

| Operator | Example |
|----------|---------|
| `$inc` | `{ $inc: { views: 1, score: -5 } }` |
| `$mul` | `{ $mul: { price: 1.1 } }` (multiply by 1.1) |
| `$min` | `{ $min: { low_score: 50 } }` (set if less than current) |
| `$max` | `{ $max: { high_score: 100 } }` (set if greater than current) |

### Array Operators

| Operator | Example |
|----------|---------|
| `$push` | `{ $push: { tags: "new" } }` |
| `$push` + `$each` | `{ $push: { tags: { $each: ["a", "b"], $sort: 1, $slice: -10 } } }` |
| `$addToSet` | `{ $addToSet: { tags: "unique" } }` (no duplicates) |
| `$pull` | `{ $pull: { tags: "remove" } }` |
| `$pull` + condition | `{ $pull: { items: { qty: { $lte: 0 } } } }` |
| `$pop` | `{ $pop: { tags: 1 } }` (remove last, -1 for first) |
| `$` (positional) | `{ $set: { "items.$.price": 29.99 } }` (matched element) |
| `$[]` (all positional) | `{ $inc: { "items.$[].price": 5 } }` (all elements) |
| `$[<id>]` (filtered) | `{ $set: { "items.$[elem].shipped": true } }` + arrayFilters |

---

## Aggregation Pipeline Stages

### Core Stages

| Stage | Purpose | Example |
|-------|---------|---------|
| `$match` | Filter documents | `{ $match: { status: "active" } }` |
| `$group` | Group and aggregate | `{ $group: { _id: "$category", total: { $sum: "$price" } } }` |
| `$project` | Reshape documents | `{ $project: { name: 1, total: { $multiply: ["$price", "$qty"] } } }` |
| `$sort` | Order results | `{ $sort: { total: -1 } }` |
| `$limit` | Limit results | `{ $limit: 10 }` |
| `$skip` | Skip results | `{ $skip: 20 }` |
| `$unwind` | Flatten arrays | `{ $unwind: "$items" }` |
| `$lookup` | Join collections | See below |
| `$addFields` | Add computed fields | `{ $addFields: { total: { $multiply: ["$price", "$qty"] } } }` |
| `$set` | Alias for $addFields | `{ $set: { fullName: { $concat: ["$first", " ", "$last"] } } }` |
| `$unset` | Remove fields | `{ $unset: ["temp_field", "internal_notes"] }` |
| `$count` | Count documents | `{ $count: "total_orders" }` |
| `$facet` | Multiple pipelines | See below |
| `$bucket` | Range-based grouping | See below |
| `$merge` | Write to collection | `{ $merge: { into: "summary_table" } }` |
| `$out` | Replace collection | `{ $out: "output_collection" }` |
| `$unionWith` | Combine collections | `{ $unionWith: "other_collection" }` |

### $lookup (JOIN)

```javascript
// Basic lookup
{ $lookup: {
    from: "orders",
    localField: "_id",
    foreignField: "user_id",
    as: "user_orders"
}}

// Pipeline lookup (more flexible)
{ $lookup: {
    from: "orders",
    let: { userId: "$_id" },
    pipeline: [
      { $match: { $expr: { $eq: ["$user_id", "$$userId"] } } },
      { $match: { status: { $ne: "cancelled" } } },
      { $sort: { created_at: -1 } },
      { $limit: 5 },
      { $project: { _id: 1, total: 1, status: 1, created_at: 1 } }
    ],
    as: "recent_orders"
}}
```

### $facet (Multiple Aggregations in One Pass)

```javascript
db.products.aggregate([
  { $facet: {
    "price_ranges": [
      { $bucket: {
          groupBy: "$price",
          boundaries: [0, 25, 50, 100, 250, 500, Infinity],
          default: "Other",
          output: { count: { $sum: 1 }, avg_price: { $avg: "$price" } }
      }}
    ],
    "top_rated": [
      { $sort: { rating: -1 } },
      { $limit: 5 },
      { $project: { name: 1, rating: 1 } }
    ],
    "category_counts": [
      { $group: { _id: "$category", count: { $sum: 1 } } },
      { $sort: { count: -1 } }
    ],
    "total": [
      { $count: "count" }
    ]
  }}
]);
```

### $setWindowFields (Window Functions)

```javascript
// Running total and rank
db.sales.aggregate([
  { $setWindowFields: {
      partitionBy: "$region",
      sortBy: { date: 1 },
      output: {
        cumulative_revenue: {
          $sum: "$revenue",
          window: { documents: ["unbounded", "current"] }
        },
        rank: {
          $rank: {}
        },
        moving_avg_7d: {
          $avg: "$revenue",
          window: { range: [-6, "current"], unit: "day" }
        }
      }
  }}
]);
```

---

## Aggregation Expressions

### String

```javascript
{ $concat: ["$firstName", " ", "$lastName"] }
{ $toUpper: "$name" }
{ $toLower: "$email" }
{ $trim: { input: "$name" } }
{ $substr: ["$name", 0, 3] }         // First 3 chars
{ $split: ["$fullName", " "] }
{ $regexMatch: { input: "$email", regex: /@gmail\.com$/ } }
{ $replaceAll: { input: "$text", find: "old", replacement: "new" } }
```

### Date

```javascript
{ $year: "$created_at" }
{ $month: "$created_at" }
{ $dayOfMonth: "$created_at" }
{ $dayOfWeek: "$created_at" }        // 1=Sunday, 7=Saturday
{ $hour: "$created_at" }
{ $dateToString: { format: "%Y-%m-%d", date: "$created_at" } }
{ $dateFromString: { dateString: "2024-01-15" } }
{ $dateDiff: { startDate: "$start", endDate: "$end", unit: "day" } }
{ $dateAdd: { startDate: "$created_at", unit: "day", amount: 30 } }
```

### Conditional

```javascript
// If-then-else
{ $cond: { if: { $gte: ["$score", 90] }, then: "A", else: "B" } }

// Switch/case
{ $switch: {
    branches: [
      { case: { $gte: ["$score", 90] }, then: "A" },
      { case: { $gte: ["$score", 80] }, then: "B" },
      { case: { $gte: ["$score", 70] }, then: "C" },
    ],
    default: "F"
}}

// Null coalescing
{ $ifNull: ["$nickname", "$name", "Unknown"] }
```

### Array

```javascript
{ $size: "$items" }                    // Array length
{ $filter: {                           // Filter array elements
    input: "$items",
    as: "item",
    cond: { $gte: ["$$item.price", 50] }
}}
{ $map: {                              // Transform array elements
    input: "$items",
    as: "item",
    in: { name: "$$item.name", total: { $multiply: ["$$item.price", "$$item.qty"] } }
}}
{ $reduce: {                           // Reduce array to single value
    input: "$items",
    initialValue: 0,
    in: { $add: ["$$value", "$$this.price"] }
}}
{ $arrayElemAt: ["$items", 0] }        // First element
{ $first: "$items" }                   // First element (shorter syntax)
{ $last: "$items" }                    // Last element
{ $slice: ["$items", 3] }             // First 3 elements
{ $reverseArray: "$items" }
{ $concatArrays: ["$items1", "$items2"] }
{ $in: ["value", "$array_field"] }     // Check if value in array
```

### Accumulators (in $group)

```javascript
{ $sum: "$amount" }                    // Sum
{ $avg: "$score" }                     // Average
{ $min: "$price" }                     // Minimum
{ $max: "$price" }                     // Maximum
{ $first: "$name" }                    // First in group
{ $last: "$name" }                     // Last in group
{ $push: "$item" }                     // Collect all into array
{ $addToSet: "$category" }             // Collect unique into array
{ $count: {} }                         // Count documents in group
{ $stdDevPop: "$score" }              // Standard deviation (population)
{ $stdDevSamp: "$score" }             // Standard deviation (sample)
```

---

## Index Types and Usage

### Index Types

| Type | Creation | Best For |
|------|----------|----------|
| Single field | `{ email: 1 }` | Equality + range on one field |
| Compound | `{ status: 1, created_at: -1 }` | Multi-field queries |
| Multikey | `{ tags: 1 }` (auto for arrays) | Array field queries |
| Text | `{ title: "text", body: "text" }` | Full-text search |
| Geospatial | `{ location: "2dsphere" }` | Geo queries |
| Hashed | `{ user_id: "hashed" }` | Equality only, shard key |
| Wildcard | `{ "attributes.$**": 1 }` | Dynamic/polymorphic fields |
| TTL | `{ expires_at: 1 }, { expireAfterSeconds: 0 }` | Auto-delete expired docs |
| Partial | `{ field: 1 }, { partialFilterExpression: {...} }` | Index subset of docs |
| Unique | `{ email: 1 }, { unique: true }` | Enforce uniqueness |

### The ESR Rule (Compound Index Field Order)

**E**quality → **S**ort → **R**ange

```javascript
// Query: status = "active" AND created_at DESC AND price >= 50
// Optimal index:
db.products.createIndex({
  status: 1,       // E: Equality (exact match)
  created_at: -1,  // S: Sort (matches sort direction)
  price: 1         // R: Range (inequality)
});

// This order ensures:
// 1. Equality narrows to exact matches first (most selective)
// 2. Sort uses the index (no in-memory sort needed)
// 3. Range scan is bounded by equality + sort
```

### Index Hints and Analysis

```javascript
// Explain query execution
db.orders.find({ status: "pending" }).explain("executionStats");

// Force a specific index
db.orders.find({ status: "pending" }).hint({ status: 1, created_at: -1 });

// Key explain output fields:
// - winningPlan.stage: "IXSCAN" (good) vs "COLLSCAN" (bad)
// - executionStats.totalKeysExamined
// - executionStats.totalDocsExamined
// - executionStats.executionTimeMillis

// List all indexes
db.collection.getIndexes();

// Drop index
db.collection.dropIndex("index_name");

// Index size
db.collection.totalIndexSize();   // Total bytes
db.collection.stats().indexSizes; // Per-index sizes
```

---

## Schema Validation

```javascript
db.createCollection("orders", {
  validator: {
    $jsonSchema: {
      bsonType: "object",
      required: ["customer_id", "items", "status", "created_at"],
      properties: {
        customer_id: {
          bsonType: "objectId",
          description: "must be an ObjectId"
        },
        status: {
          bsonType: "string",
          enum: ["pending", "confirmed", "shipped", "delivered", "cancelled"],
          description: "must be a valid status"
        },
        items: {
          bsonType: "array",
          minItems: 1,
          items: {
            bsonType: "object",
            required: ["product_id", "quantity", "price"],
            properties: {
              product_id: { bsonType: "objectId" },
              quantity: { bsonType: "int", minimum: 1 },
              price: { bsonType: "double", minimum: 0 }
            }
          }
        },
        total: { bsonType: "double", minimum: 0 },
        notes: { bsonType: "string", maxLength: 1000 },
        created_at: { bsonType: "date" }
      }
    }
  },
  validationLevel: "moderate",    // "strict" | "moderate" | "off"
  validationAction: "error"       // "error" | "warn"
});

// Modify validation on existing collection
db.runCommand({
  collMod: "orders",
  validator: { /* updated schema */ },
  validationLevel: "moderate"
});
```

---

## Change Streams

```javascript
// Watch for changes on a collection
const changeStream = db.orders.watch([
  { $match: { "fullDocument.status": "shipped" } }
], {
  fullDocument: "updateLookup"  // Include full document on updates
});

changeStream.on("change", (change) => {
  console.log("Change type:", change.operationType);
  // insert, update, replace, delete, invalidate
  console.log("Document:", change.fullDocument);
  console.log("Update:", change.updateDescription);
});

// Resume from a specific point
const resumeToken = change._id;
const resumed = db.orders.watch([], { resumeAfter: resumeToken });

// Watch entire database
const dbStream = db.watch();

// Watch entire deployment
const client = new MongoClient(uri);
const deploymentStream = client.watch();
```

---

## Transactions

```javascript
const session = client.startSession();

try {
  session.startTransaction({
    readConcern: { level: "snapshot" },
    writeConcern: { w: "majority" },
    readPreference: "primary"
  });

  // All operations use the session
  await db.collection("accounts").updateOne(
    { _id: fromAccount },
    { $inc: { balance: -amount } },
    { session }
  );

  await db.collection("accounts").updateOne(
    { _id: toAccount },
    { $inc: { balance: amount } },
    { session }
  );

  await db.collection("transfers").insertOne({
    from: fromAccount,
    to: toAccount,
    amount,
    timestamp: new Date()
  }, { session });

  await session.commitTransaction();
} catch (error) {
  await session.abortTransaction();
  throw error;
} finally {
  await session.endSession();
}
```

---

## Administration Commands

```javascript
// Database stats
db.stats()
db.collection.stats()

// Server status
db.serverStatus()
db.currentOp()                        // Currently running operations
db.killOp(opId)                       // Kill a long-running operation

// Profiler (slow query log)
db.setProfilingLevel(1, { slowms: 100 })  // Log queries slower than 100ms
db.system.profile.find().sort({ ts: -1 }).limit(10)

// Compact collection (reclaim space)
db.runCommand({ compact: "collection_name" })

// Repair database
db.repairDatabase()

// User management
db.createUser({
  user: "appuser",
  pwd: "password",
  roles: [
    { role: "readWrite", db: "myapp" },
    { role: "read", db: "analytics" }
  ]
});

// Backup
// mongodump --uri="mongodb://..." --out=/backup/$(date +%Y%m%d)
// mongorestore --uri="mongodb://..." /backup/20240115/
```

---

## Connection String Reference

```
# Standard
mongodb://username:password@host:27017/database?authSource=admin

# Replica set
mongodb://host1:27017,host2:27017,host3:27017/database?replicaSet=rs0

# Atlas (SRV)
mongodb+srv://username:password@cluster.mongodb.net/database?retryWrites=true&w=majority

# Common options
?retryWrites=true          # Auto-retry failed writes
&w=majority                # Write concern: majority acknowledgment
&readPreference=secondary  # Read from replicas
&maxPoolSize=50            # Connection pool size
&connectTimeoutMS=10000    # Connection timeout
&socketTimeoutMS=30000     # Socket timeout
&authSource=admin          # Authentication database
&ssl=true                  # Enable SSL/TLS
```

---

## Performance Anti-Patterns

| Anti-Pattern | Problem | Solution |
|-------------|---------|----------|
| COLLSCAN queries | Full collection scan | Add appropriate index |
| `$lookup` in hot path | Multiple round trips | Denormalize data |
| Unbounded `$push` to arrays | Documents exceed 16MB | Use bucket pattern or separate collection |
| `$where` or `$function` | JavaScript evaluation, no index | Use native operators |
| Large `$in` arrays (>1000) | Poor index utilization | Restructure query or use `$lookup` |
| No projection on `find()` | Transfer unnecessary data | Specify needed fields |
| `find().count()` | Deprecated, inaccurate | Use `countDocuments()` |
| Populate N+1 (Mongoose) | One query per referenced doc | Use `$lookup` in aggregation |
| Missing write concern | Silent data loss | Use `w: "majority"` |
| No read preference config | All reads hit primary | Use `secondaryPreferred` for analytics |
