---
name: mongodb-queries
description: >
  Build MongoDB queries, aggregation pipelines, and update operations. Covers CRUD, text search,
  geospatial queries, array operations, and bulk writes. Generates optimized queries with
  proper index usage.
  Triggers: "MongoDB query", "MongoDB aggregation", "MongoDB find", "MongoDB update",
  "MongoDB bulk", "Mongoose query", "MongoDB text search", "MongoDB geo".
  NOT for: schema design (use mongodb-expert agent), cluster operations, backup/restore.
version: 1.0.0
argument-hint: "[query-description]"
allowed-tools: Read, Grep, Glob, Bash, Write, Edit
model: sonnet
---

# MongoDB Queries

Build optimized MongoDB queries and aggregation pipelines. Every query pattern includes the recommended index.

## Query Operators Quick Reference

### Comparison

```javascript
{ field: value }                    // Equality
{ field: { $eq: value } }          // Explicit equality
{ field: { $ne: value } }          // Not equal
{ field: { $gt: value } }          // Greater than
{ field: { $gte: value } }         // Greater than or equal
{ field: { $lt: value } }          // Less than
{ field: { $lte: value } }         // Less than or equal
{ field: { $in: [v1, v2, v3] } }   // In array
{ field: { $nin: [v1, v2, v3] } }  // Not in array
```

### Logical

```javascript
{ $and: [{ field1: v1 }, { field2: v2 }] }   // AND (implicit with comma)
{ $or: [{ field1: v1 }, { field2: v2 }] }     // OR
{ $not: { field: { $gt: 5 } } }               // NOT
{ $nor: [{ field1: v1 }, { field2: v2 }] }    // NOR
```

### Array

```javascript
{ tags: "javascript" }                          // Array contains value
{ tags: { $all: ["javascript", "nodejs"] } }   // Contains all values
{ tags: { $size: 3 } }                         // Array has exactly 3 elements
{ tags: { $elemMatch: { $gte: 5, $lt: 10 } } } // Element matches all conditions
{ "items.0.name": "Widget" }                    // First element's name
```

### Element

```javascript
{ field: { $exists: true } }       // Field exists
{ field: { $type: "string" } }     // Field is string type
```

### Update Operators

```javascript
// Set/unset
{ $set: { "name": "Alice", "address.city": "NYC" } }
{ $unset: { "deprecated_field": "" } }

// Numeric
{ $inc: { "stats.views": 1, "stats.score": -5 } }
{ $mul: { "price": 1.1 } }                              // Multiply by 1.1
{ $min: { "low_score": 50 } }                            // Set if less than current
{ $max: { "high_score": 100 } }                          // Set if greater than current

// Array
{ $push: { tags: "new-tag" } }                           // Append to array
{ $push: { tags: { $each: ["a", "b"], $sort: 1 } } }   // Push multiple, sort
{ $addToSet: { tags: "unique-tag" } }                    // Add if not present
{ $pull: { tags: "remove-me" } }                         // Remove matching elements
{ $pop: { tags: 1 } }                                    // Remove last element (-1 for first)

// Positional
{ $set: { "items.$.price": 29.99 } }                    // Update matched array element
{ $set: { "items.$[elem].price": 29.99 } }              // With arrayFilter
// arrayFilters: [{ "elem.sku": "ABC123" }]
```

## Common Query Patterns

### Pagination

```javascript
// Offset-based (simple but slow for large offsets)
db.products.find({ category: "electronics" })
  .sort({ created_at: -1 })
  .skip(page * pageSize)
  .limit(pageSize);

// Index: { category: 1, created_at: -1 }

// Cursor-based (efficient for infinite scroll)
// First page
db.products.find({ category: "electronics" })
  .sort({ created_at: -1, _id: -1 })
  .limit(pageSize);

// Next pages (use last item's values as cursor)
db.products.find({
  category: "electronics",
  $or: [
    { created_at: { $lt: lastCreatedAt } },
    { created_at: lastCreatedAt, _id: { $lt: lastId } },
  ],
})
  .sort({ created_at: -1, _id: -1 })
  .limit(pageSize);

// Index: { category: 1, created_at: -1, _id: -1 }
```

### Full-Text Search

```javascript
// Create text index
db.articles.createIndex({
  title: "text",
  body: "text",
  tags: "text",
}, {
  weights: { title: 10, tags: 5, body: 1 },  // Title matches score 10x
  name: "article_search",
  default_language: "english",
});

// Search
db.articles.find(
  { $text: { $search: "mongodb performance tuning" } },
  { score: { $meta: "textScore" } }
)
  .sort({ score: { $meta: "textScore" } })
  .limit(20);

// Phrase search
db.articles.find({
  $text: { $search: '"exact phrase" other words' },
});

// Exclude terms
db.articles.find({
  $text: { $search: "mongodb -mysql" },
});
```

### Geospatial Queries

```javascript
// Create 2dsphere index
db.stores.createIndex({ location: "2dsphere" });

// Store with GeoJSON point
db.stores.insertOne({
  name: "Downtown Store",
  location: {
    type: "Point",
    coordinates: [-73.9857, 40.7484],  // [longitude, latitude]
  },
});

// Find stores within 5km
db.stores.find({
  location: {
    $near: {
      $geometry: { type: "Point", coordinates: [-73.9857, 40.7484] },
      $maxDistance: 5000,  // meters
    },
  },
});

// Find stores within a polygon
db.stores.find({
  location: {
    $geoWithin: {
      $geometry: {
        type: "Polygon",
        coordinates: [[
          [-74.0, 40.7], [-73.9, 40.7],
          [-73.9, 40.8], [-74.0, 40.8],
          [-74.0, 40.7],  // Close the polygon
        ]],
      },
    },
  },
});
```

### Bulk Operations

```javascript
const { MongoClient } = require("mongodb");

async function bulkUpsert(collection, documents) {
  const operations = documents.map((doc) => ({
    updateOne: {
      filter: { _id: doc._id },
      update: { $set: doc },
      upsert: true,
    },
  }));

  const result = await collection.bulkWrite(operations, {
    ordered: false,  // Continue on error (parallel execution)
    writeConcern: { w: "majority" },
  });

  return {
    inserted: result.upsertedCount,
    modified: result.modifiedCount,
    matched: result.matchedCount,
  };
}

// Ordered bulk write (stops on first error)
async function orderedBulk(collection, operations) {
  return collection.bulkWrite(operations, { ordered: true });
}
```

### Conditional Updates

```javascript
// Update only if condition is met (atomic)
db.inventory.updateOne(
  {
    product_id: "SKU-001",
    available_quantity: { $gte: 5 },  // Only if enough stock
  },
  {
    $inc: { available_quantity: -5 },
    $push: {
      reservations: {
        order_id: "ORD-123",
        quantity: 5,
        reserved_at: new Date(),
      },
    },
  }
);

// findOneAndUpdate — return the document
const result = await collection.findOneAndUpdate(
  { _id: orderId, status: "pending" },
  {
    $set: { status: "processing", started_at: new Date() },
  },
  {
    returnDocument: "after",  // Return updated document
    projection: { _id: 1, status: 1, items: 1 },
  }
);

if (!result) {
  throw new Error("Order not found or not in pending status");
}
```

## Aggregation Recipes

### Top N per Group

```javascript
// Top 3 products per category by revenue
db.orders.aggregate([
  { $unwind: "$items" },
  {
    $group: {
      _id: { category: "$items.category", product: "$items.product_id" },
      total_revenue: { $sum: { $multiply: ["$items.price", "$items.quantity"] } },
      order_count: { $sum: 1 },
    },
  },
  { $sort: { "_id.category": 1, total_revenue: -1 } },
  {
    $group: {
      _id: "$_id.category",
      top_products: {
        $push: {
          product_id: "$_id.product",
          revenue: "$total_revenue",
          orders: "$order_count",
        },
      },
    },
  },
  {
    $project: {
      category: "$_id",
      top_products: { $slice: ["$top_products", 3] },
    },
  },
]);
```

### Cohort Analysis

```javascript
// Monthly cohort retention analysis
db.users.aggregate([
  // Determine signup cohort
  {
    $addFields: {
      signup_cohort: {
        $dateToString: { format: "%Y-%m", date: "$created_at" },
      },
    },
  },

  // Lookup user activity
  {
    $lookup: {
      from: "events",
      let: { userId: "$_id" },
      pipeline: [
        { $match: { $expr: { $eq: ["$user_id", "$$userId"] } } },
        {
          $group: {
            _id: { $dateToString: { format: "%Y-%m", date: "$event_date" } },
          },
        },
      ],
      as: "active_months",
    },
  },

  // Calculate months since signup
  { $unwind: "$active_months" },
  {
    $addFields: {
      months_since_signup: {
        $dateDiff: {
          startDate: { $dateFromString: { dateString: { $concat: ["$signup_cohort", "-01"] } } },
          endDate: { $dateFromString: { dateString: { $concat: ["$active_months._id", "-01"] } } },
          unit: "month",
        },
      },
    },
  },

  // Aggregate cohort retention
  {
    $group: {
      _id: { cohort: "$signup_cohort", month: "$months_since_signup" },
      active_users: { $addToSet: "$_id" },
    },
  },
  {
    $project: {
      cohort: "$_id.cohort",
      months_since_signup: "$_id.month",
      active_count: { $size: "$active_users" },
    },
  },
  { $sort: { cohort: 1, months_since_signup: 1 } },
]);
```

### Running Totals

```javascript
// Cumulative revenue by day
db.orders.aggregate([
  {
    $group: {
      _id: { $dateToString: { format: "%Y-%m-%d", date: "$order_date" } },
      daily_revenue: { $sum: "$total" },
    },
  },
  { $sort: { _id: 1 } },
  {
    $setWindowFields: {
      sortBy: { _id: 1 },
      output: {
        cumulative_revenue: {
          $sum: "$daily_revenue",
          window: { documents: ["unbounded", "current"] },
        },
      },
    },
  },
]);
```

## Mongoose Patterns (Node.js)

### Schema with Validation

```javascript
const mongoose = require("mongoose");

const orderSchema = new mongoose.Schema(
  {
    order_number: {
      type: String,
      required: true,
      unique: true,
      index: true,
    },
    customer: {
      type: mongoose.Schema.Types.ObjectId,
      ref: "Customer",
      required: true,
      index: true,
    },
    items: [
      {
        product: { type: mongoose.Schema.Types.ObjectId, ref: "Product" },
        name: String,
        price: { type: Number, min: 0 },
        quantity: { type: Number, min: 1, default: 1 },
      },
    ],
    status: {
      type: String,
      enum: ["pending", "paid", "shipped", "delivered", "cancelled"],
      default: "pending",
      index: true,
    },
    total: { type: Number, required: true, min: 0 },
    notes: String,
  },
  {
    timestamps: true,
    toJSON: { virtuals: true },
  }
);

// Virtual field
orderSchema.virtual("item_count").get(function () {
  return this.items.length;
});

// Compound index
orderSchema.index({ customer: 1, createdAt: -1 });
orderSchema.index({ status: 1, createdAt: -1 });

// Pre-save hook
orderSchema.pre("save", function (next) {
  this.total = this.items.reduce(
    (sum, item) => sum + item.price * item.quantity,
    0
  );
  next();
});

// Static method
orderSchema.statics.findByCustomer = function (customerId, options = {}) {
  return this.find({ customer: customerId })
    .sort({ createdAt: -1 })
    .limit(options.limit || 20)
    .populate("items.product", "name price image");
};

// Instance method
orderSchema.methods.cancel = async function () {
  if (this.status === "shipped" || this.status === "delivered") {
    throw new Error("Cannot cancel shipped or delivered orders");
  }
  this.status = "cancelled";
  return this.save();
};

const Order = mongoose.model("Order", orderSchema);
```

### Query Building

```javascript
// Chainable query builder
const orders = await Order.find()
  .where("status").in(["pending", "paid"])
  .where("total").gte(100)
  .where("createdAt").gte(new Date("2024-01-01"))
  .sort("-createdAt")
  .limit(20)
  .select("order_number status total createdAt")
  .populate("customer", "name email")
  .lean()  // Return plain objects (faster, no Mongoose overhead)
  .exec();
```

## Gotchas

- Always check `explain()` output for COLLSCAN — means no index is used
- `$in` with large arrays (>1000 values) performs poorly — restructure the query
- Array updates with `$push` can grow documents beyond 16MB — use bucket pattern
- `$lookup` (JOIN) is expensive — denormalize data for frequently joined collections
- Don't use `$where` (JavaScript evaluation) — much slower than native operators
- Mongoose `populate()` executes N+1 queries — use `$lookup` in aggregation instead
- `find().count()` is deprecated — use `countDocuments()` or `estimatedDocumentCount()`
- Always use `{ ordered: false }` for bulk inserts unless order matters (2-3x faster)
