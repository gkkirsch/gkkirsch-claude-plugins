---
name: mongodb-schema-design
description: >
  Design MongoDB document schemas, aggregation pipelines, indexes, and data access patterns.
  Follows best practices for embedding vs referencing, schema validation, and migration.
  Triggers: "MongoDB schema", "document schema", "Mongoose model", "aggregation pipeline",
  "MongoDB index", "embed vs reference", "MongoDB migration".
  NOT for: Redis caching (use redis-implementation), SQL database design (use data-modeling).
version: 1.0.0
argument-hint: "[domain or collection to design]"
allowed-tools: Read, Grep, Glob, Bash, Write, Edit
model: sonnet
---

# MongoDB Schema Design

Design document schemas optimized for your application's access patterns, with proper validation, indexing, and migration strategies.

## Design Process

### Step 1: Map Access Patterns

Before designing any schema, list every query the application will run:

```markdown
## Access Pattern Analysis

| # | Operation | Frequency | Latency SLA | Fields Needed |
|---|-----------|-----------|-------------|---------------|
| 1 | Get user by ID | Very high | <10ms | All user fields |
| 2 | Get user by email | High | <10ms | All user fields |
| 3 | List user's orders | High | <50ms | Order summary, last 20 |
| 4 | Search products by category | High | <100ms | Name, price, image, rating |
| 5 | Get order with items + shipping | Medium | <50ms | Full order detail |
| 6 | Analytics: revenue by category/month | Low | <5s | Aggregation |
```

### Step 2: Choose Embedding vs Referencing

Apply the decision matrix from the access pattern analysis:

```
For each relationship, ask:

1. "Do I always need this data when I read the parent?"
   → Yes: EMBED

2. "Can this subdocument array grow without bound?"
   → Yes: REFERENCE (separate collection)

3. "Does this data change independently of the parent?"
   → Yes: REFERENCE

4. "Is this a many-to-many relationship?"
   → Yes: REFERENCE (with array of ObjectIds or junction collection)

5. "Is the subdocument > 16KB or the parent approaching 16MB?"
   → Yes: REFERENCE
```

### Step 3: Define the Schema

#### E-Commerce Example

```javascript
// users collection — embed profile, reference orders
{
  _id: ObjectId("..."),
  email: "alice@example.com",      // Unique index
  password_hash: "...",
  name: "Alice Smith",
  profile: {                        // EMBEDDED — always read with user
    bio: "...",
    avatar_url: "...",
    preferences: { theme: "dark", language: "en" }
  },
  addresses: [                      // EMBEDDED — bounded array (max 5)
    {
      label: "Home",
      street: "123 Main St",
      city: "Portland",
      state: "OR",
      zip: "97201",
      is_default: true
    }
  ],
  created_at: ISODate("2024-01-15"),
  updated_at: ISODate("2024-03-01")
}

// orders collection — embed items, reference user + products
{
  _id: ObjectId("..."),
  order_number: "ORD-2024-001234",   // Unique index
  user_id: ObjectId("..."),           // REFERENCE — user exists independently
  status: "shipped",

  // EXTENDED REFERENCE — frequently-read user fields denormalized
  customer: {
    _id: ObjectId("..."),
    name: "Alice Smith",
    email: "alice@example.com"
  },

  items: [                            // EMBEDDED — part of the order, bounded
    {
      product_id: ObjectId("..."),
      product_name: "Wireless Mouse",  // Denormalized at time of purchase
      product_image: "/images/mouse.jpg",
      quantity: 2,
      unit_price: 29.99,
      subtotal: 59.98
    }
  ],

  shipping: {                         // EMBEDDED — 1:1 with order
    address: { street: "...", city: "...", state: "...", zip: "..." },
    method: "standard",
    tracking_number: "1Z999AA10123456784",
    estimated_delivery: ISODate("2024-03-10"),
    shipped_at: ISODate("2024-03-05")
  },

  payment: {                          // EMBEDDED — 1:1 with order
    method: "credit_card",
    last_four: "4242",
    charged_at: ISODate("2024-03-01"),
    amount: 65.97
  },

  totals: {                           // PRE-COMPUTED — avoid recalculating
    subtotal: 59.98,
    shipping: 5.99,
    tax: 4.80,
    total: 70.77
  },

  timeline: [                         // EMBEDDED — bounded lifecycle events
    { event: "created", at: ISODate("2024-03-01T10:00:00Z") },
    { event: "paid", at: ISODate("2024-03-01T10:01:00Z") },
    { event: "shipped", at: ISODate("2024-03-05T14:00:00Z") }
  ],

  created_at: ISODate("2024-03-01"),
  updated_at: ISODate("2024-03-05")
}

// products collection — standalone, referenced by orders
{
  _id: ObjectId("..."),
  sku: "WM-001",
  name: "Wireless Mouse",
  slug: "wireless-mouse",
  description: "Ergonomic wireless mouse...",

  category: {                        // EMBEDDED — changes rarely
    primary: "Electronics",
    secondary: "Peripherals",
    tertiary: "Mice"
  },

  pricing: {
    base_price: 29.99,
    sale_price: null,
    currency: "USD"
  },

  inventory: {
    in_stock: true,
    quantity: 150,
    warehouse: "PDX-01"
  },

  images: [
    { url: "/images/mouse-1.jpg", alt: "Front view", is_primary: true },
    { url: "/images/mouse-2.jpg", alt: "Side view", is_primary: false }
  ],

  // COMPUTED PATTERN — updated on review write
  reviews_summary: {
    count: 47,
    average_rating: 4.3,
    rating_distribution: { "5": 20, "4": 15, "3": 8, "2": 3, "1": 1 }
  },

  // ATTRIBUTE PATTERN — flexible specs
  attributes: [
    { key: "color", value: "Black" },
    { key: "connectivity", value: "Bluetooth 5.0" },
    { key: "battery_life", value: "12 months" },
    { key: "weight", value: "80g" }
  ],

  tags: ["wireless", "ergonomic", "bluetooth"],
  is_active: true,
  created_at: ISODate("2024-01-01"),
  updated_at: ISODate("2024-03-01")
}

// reviews collection — separate (unbounded per product)
{
  _id: ObjectId("..."),
  product_id: ObjectId("..."),      // Index
  user_id: ObjectId("..."),
  user_name: "Alice S.",            // Denormalized for display
  rating: 5,
  title: "Great mouse!",
  body: "Very comfortable...",
  helpful_votes: 12,
  verified_purchase: true,
  created_at: ISODate("2024-02-15")
}
```

### Step 4: Add Schema Validation

```javascript
db.createCollection("orders", {
  validator: {
    $jsonSchema: {
      bsonType: "object",
      required: ["order_number", "user_id", "status", "items", "totals", "created_at"],
      properties: {
        order_number: {
          bsonType: "string",
          pattern: "^ORD-\\d{4}-\\d{6}$"
        },
        user_id: { bsonType: "objectId" },
        status: {
          bsonType: "string",
          enum: ["pending", "confirmed", "processing", "shipped", "delivered", "cancelled", "returned"]
        },
        items: {
          bsonType: "array",
          minItems: 1,
          items: {
            bsonType: "object",
            required: ["product_id", "quantity", "unit_price"],
            properties: {
              product_id: { bsonType: "objectId" },
              quantity: { bsonType: "int", minimum: 1 },
              unit_price: { bsonType: "double", minimum: 0 }
            }
          }
        },
        totals: {
          bsonType: "object",
          required: ["total"],
          properties: {
            total: { bsonType: "double", minimum: 0 }
          }
        }
      }
    }
  },
  validationLevel: "moderate",   // Apply to inserts + updates (not existing docs)
  validationAction: "error"      // Reject invalid documents
});
```

### Step 5: Create Indexes

```javascript
// Users
db.users.createIndex({ email: 1 }, { unique: true });
db.users.createIndex({ created_at: -1 });

// Orders — compound index following ESR rule
db.orders.createIndex({ user_id: 1, created_at: -1 });  // User's orders, newest first
db.orders.createIndex({ status: 1, created_at: -1 });   // Orders by status
db.orders.createIndex({ order_number: 1 }, { unique: true });
db.orders.createIndex(
  { "shipping.estimated_delivery": 1 },
  { partialFilterExpression: { status: "shipped" } }  // Only index shipped orders
);

// Products
db.products.createIndex({ slug: 1 }, { unique: true });
db.products.createIndex({ "category.primary": 1, "pricing.base_price": 1 });
db.products.createIndex({ tags: 1 });  // Multikey for array
db.products.createIndex({ "attributes.key": 1, "attributes.value": 1 });
db.products.createIndex(
  { name: "text", description: "text", tags: "text" },
  { weights: { name: 10, tags: 5, description: 1 } }
);

// Reviews
db.reviews.createIndex({ product_id: 1, created_at: -1 });
db.reviews.createIndex({ user_id: 1, product_id: 1 }, { unique: true });  // One review per user per product
```

## Mongoose Schema Patterns

### Base Schema with Plugins

```typescript
import mongoose, { Schema, Document } from 'mongoose';

// Timestamp plugin (alternative to timestamps: true for custom behavior)
function timestampPlugin(schema: Schema) {
  schema.add({
    created_at: { type: Date, default: Date.now, immutable: true },
    updated_at: { type: Date, default: Date.now },
  });
  schema.pre('save', function(next) {
    this.updated_at = new Date();
    next();
  });
  schema.pre('findOneAndUpdate', function(next) {
    this.set({ updated_at: new Date() });
    next();
  });
}

// Soft delete plugin
function softDeletePlugin(schema: Schema) {
  schema.add({
    is_deleted: { type: Boolean, default: false, index: true },
    deleted_at: { type: Date },
  });

  // Auto-filter deleted documents
  schema.pre(/^find/, function(next) {
    if (!this.getQuery().is_deleted) {
      this.where({ is_deleted: { $ne: true } });
    }
    next();
  });

  schema.methods.softDelete = function() {
    this.is_deleted = true;
    this.deleted_at = new Date();
    return this.save();
  };

  schema.methods.restore = function() {
    this.is_deleted = false;
    this.deleted_at = undefined;
    return this.save();
  };
}

// Apply globally
mongoose.plugin(timestampPlugin);
```

### Pagination Pattern

```typescript
interface PaginationResult<T> {
  docs: T[];
  totalDocs: number;
  page: number;
  totalPages: number;
  hasNextPage: boolean;
  hasPrevPage: boolean;
}

async function paginate<T>(
  model: mongoose.Model<T>,
  filter: any = {},
  options: { page?: number; limit?: number; sort?: any; select?: string } = {}
): Promise<PaginationResult<T>> {
  const page = Math.max(1, options.page || 1);
  const limit = Math.min(100, Math.max(1, options.limit || 20));
  const skip = (page - 1) * limit;
  const sort = options.sort || { created_at: -1 };

  const [docs, totalDocs] = await Promise.all([
    model.find(filter)
      .sort(sort)
      .skip(skip)
      .limit(limit)
      .select(options.select || '')
      .lean(),
    model.countDocuments(filter),
  ]);

  const totalPages = Math.ceil(totalDocs / limit);

  return {
    docs: docs as T[],
    totalDocs,
    page,
    totalPages,
    hasNextPage: page < totalPages,
    hasPrevPage: page > 1,
  };
}

// Cursor-based pagination (more efficient for large datasets)
async function cursorPaginate<T>(
  model: mongoose.Model<T>,
  filter: any = {},
  options: { cursor?: string; limit?: number; sort?: string }
): Promise<{ docs: T[]; nextCursor: string | null }> {
  const limit = Math.min(100, options.limit || 20);
  const sortField = options.sort || '_id';

  if (options.cursor) {
    filter[sortField] = { $gt: new mongoose.Types.ObjectId(options.cursor) };
  }

  const docs = await model.find(filter)
    .sort({ [sortField]: 1 })
    .limit(limit + 1)  // Fetch one extra to check for next page
    .lean();

  const hasMore = docs.length > limit;
  const results = hasMore ? docs.slice(0, limit) : docs;
  const nextCursor = hasMore ? String((results[results.length - 1] as any)[sortField]) : null;

  return { docs: results as T[], nextCursor };
}
```

## Migration Patterns

### Schema Migration Script

```typescript
// migrations/001_add_customer_segment.ts
import mongoose from 'mongoose';

export async function up(db: mongoose.Connection) {
  // Add new field with default value
  await db.collection('users').updateMany(
    { customer_segment: { $exists: false } },
    { $set: { customer_segment: 'standard' } }
  );

  // Create index for new field
  await db.collection('users').createIndex(
    { customer_segment: 1 },
    { background: true }
  );

  console.log('Migration 001: Added customer_segment to all users');
}

export async function down(db: mongoose.Connection) {
  await db.collection('users').updateMany(
    {},
    { $unset: { customer_segment: '' } }
  );
  await db.collection('users').dropIndex('customer_segment_1');
  console.log('Migration 001: Rolled back customer_segment');
}
```

### Data Reshaping Migration

```javascript
// Migrate from flat structure to nested
// Before: { street, city, state, zip }
// After:  { address: { street, city, state, zip } }

db.users.find({ street: { $exists: true } }).forEach(function(doc) {
  db.users.updateOne(
    { _id: doc._id },
    {
      $set: {
        address: {
          street: doc.street,
          city: doc.city,
          state: doc.state,
          zip: doc.zip
        }
      },
      $unset: {
        street: "",
        city: "",
        state: "",
        zip: ""
      }
    }
  );
});
```

## Checklist Before Completing

- [ ] Every query maps to an index (no COLLSCAN in explain output)
- [ ] No unbounded arrays (all arrays have a practical maximum)
- [ ] Schema validation added for required fields and enums
- [ ] Denormalized fields documented with update strategy
- [ ] Compound indexes follow ESR rule (Equality → Sort → Range)
- [ ] Unique constraints on business keys (email, slug, order_number)
- [ ] TTL index set for temporary data (sessions, tokens, logs)
- [ ] Read preference configured (primary for writes, secondary for analytics)
