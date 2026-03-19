---
name: inventory-management
description: >
  Build inventory tracking systems for e-commerce — stock management, low-stock alerts,
  backorder handling, variant tracking, and warehouse management patterns.
  Works with PostgreSQL, Prisma, and Express/Next.js.
  Triggers: "inventory management", "stock tracking", "inventory system",
  "product catalog", "variant management", "low stock".
  NOT for: physical warehouse robotics or ERP systems.
version: 1.0.0
argument-hint: "[setup|variants|alerts]"
allowed-tools: Read, Grep, Glob, Write, Edit, Bash
---

# Inventory Management

Build inventory tracking for e-commerce applications.

## Database Schema

### PostgreSQL + Prisma

```prisma
model Product {
  id          String   @id @default(cuid())
  name        String
  slug        String   @unique
  description String?
  price       Int      // cents
  compareAt   Int?     // original price for sales (cents)
  images      String[] // URLs
  category    String?
  tags        String[]
  status      ProductStatus @default(ACTIVE)
  createdAt   DateTime @default(now())
  updatedAt   DateTime @updatedAt
  variants    Variant[]
  inventory   Inventory?
}

model Variant {
  id        String    @id @default(cuid())
  productId String
  product   Product   @relation(fields: [productId], references: [id])
  name      String    // e.g., "Large / Blue"
  sku       String    @unique
  price     Int?      // override product price if set
  options   Json      // { size: "Large", color: "Blue" }
  inventory Inventory?
}

model Inventory {
  id          String   @id @default(cuid())
  productId   String?  @unique
  product     Product? @relation(fields: [productId], references: [id])
  variantId   String?  @unique
  variant     Variant? @relation(fields: [variantId], references: [id])
  quantity    Int      @default(0)
  reserved    Int      @default(0)  // reserved by in-progress orders
  lowStockAt  Int      @default(5)  // alert threshold
  trackStock  Boolean  @default(true)
  allowBackorder Boolean @default(false)
  updatedAt   DateTime @updatedAt

  @@index([quantity])
}

model InventoryMovement {
  id          String        @id @default(cuid())
  inventoryId String
  type        MovementType
  quantity    Int           // positive = add, negative = remove
  reason      String?       // "order_placed", "order_cancelled", "restock", "adjustment"
  referenceId String?       // order ID, restock ID, etc.
  createdAt   DateTime      @default(now())
  createdBy   String?       // user or system

  @@index([inventoryId, createdAt])
}

enum ProductStatus {
  ACTIVE
  DRAFT
  ARCHIVED
}

enum MovementType {
  SALE
  RESTOCK
  RETURN
  ADJUSTMENT
  RESERVATION
  RELEASE
}
```

## Stock Operations

### Reserve Stock (On Order Placement)

```typescript
async function reserveStock(items: { variantId: string; quantity: number }[]) {
  return prisma.$transaction(async (tx) => {
    for (const item of items) {
      const inventory = await tx.inventory.findUnique({
        where: { variantId: item.variantId },
      });

      if (!inventory) throw new Error(`No inventory record for variant ${item.variantId}`);

      const available = inventory.quantity - inventory.reserved;
      if (!inventory.allowBackorder && available < item.quantity) {
        throw new Error(`Insufficient stock for variant ${item.variantId}. Available: ${available}`);
      }

      await tx.inventory.update({
        where: { variantId: item.variantId },
        data: { reserved: { increment: item.quantity } },
      });

      await tx.inventoryMovement.create({
        data: {
          inventoryId: inventory.id,
          type: 'RESERVATION',
          quantity: -item.quantity,
          reason: 'order_placed',
          referenceId: null, // set after order is created
        },
      });
    }
  });
}
```

### Commit Stock (On Payment Confirmed)

```typescript
async function commitStock(orderId: string, items: { variantId: string; quantity: number }[]) {
  return prisma.$transaction(async (tx) => {
    for (const item of items) {
      await tx.inventory.update({
        where: { variantId: item.variantId },
        data: {
          quantity: { decrement: item.quantity },
          reserved: { decrement: item.quantity },
        },
      });

      const inventory = await tx.inventory.findUnique({ where: { variantId: item.variantId } });

      await tx.inventoryMovement.create({
        data: {
          inventoryId: inventory!.id,
          type: 'SALE',
          quantity: -item.quantity,
          reason: 'payment_confirmed',
          referenceId: orderId,
        },
      });

      // Check low stock alert
      if (inventory && inventory.quantity <= inventory.lowStockAt) {
        await sendLowStockAlert(inventory);
      }
    }
  });
}
```

### Release Stock (On Order Cancelled)

```typescript
async function releaseStock(orderId: string, items: { variantId: string; quantity: number }[]) {
  return prisma.$transaction(async (tx) => {
    for (const item of items) {
      await tx.inventory.update({
        where: { variantId: item.variantId },
        data: { reserved: { decrement: item.quantity } },
      });

      const inventory = await tx.inventory.findUnique({ where: { variantId: item.variantId } });

      await tx.inventoryMovement.create({
        data: {
          inventoryId: inventory!.id,
          type: 'RELEASE',
          quantity: item.quantity,
          reason: 'order_cancelled',
          referenceId: orderId,
        },
      });
    }
  });
}
```

### Restock

```typescript
async function restock(variantId: string, quantity: number, reason?: string) {
  return prisma.$transaction(async (tx) => {
    const inventory = await tx.inventory.update({
      where: { variantId },
      data: { quantity: { increment: quantity } },
    });

    await tx.inventoryMovement.create({
      data: {
        inventoryId: inventory.id,
        type: 'RESTOCK',
        quantity,
        reason: reason || 'manual_restock',
      },
    });

    return inventory;
  });
}
```

## API Endpoints

```typescript
// Check availability
app.get('/api/products/:id/availability', async (req, res) => {
  const product = await prisma.product.findUnique({
    where: { id: req.params.id },
    include: {
      variants: { include: { inventory: true } },
      inventory: true,
    },
  });

  if (!product) return res.status(404).json({ error: 'Not found' });

  if (product.variants.length > 0) {
    const availability = product.variants.map(v => ({
      variantId: v.id,
      name: v.name,
      sku: v.sku,
      available: v.inventory
        ? v.inventory.quantity - v.inventory.reserved
        : 0,
      inStock: v.inventory
        ? (v.inventory.quantity - v.inventory.reserved) > 0 || v.inventory.allowBackorder
        : false,
    }));
    return res.json({ availability });
  }

  const inv = product.inventory;
  res.json({
    available: inv ? inv.quantity - inv.reserved : 0,
    inStock: inv ? (inv.quantity - inv.reserved) > 0 || inv.allowBackorder : false,
  });
});

// Admin: Get low stock products
app.get('/api/admin/low-stock', async (req, res) => {
  const lowStock = await prisma.inventory.findMany({
    where: {
      trackStock: true,
      quantity: { lte: prisma.inventory.fields.lowStockAt },
    },
    include: {
      product: { select: { name: true } },
      variant: { select: { name: true, sku: true } },
    },
    orderBy: { quantity: 'asc' },
  });

  res.json(lowStock);
});

// Admin: Get inventory movements
app.get('/api/admin/inventory/:id/movements', async (req, res) => {
  const movements = await prisma.inventoryMovement.findMany({
    where: { inventoryId: req.params.id },
    orderBy: { createdAt: 'desc' },
    take: 50,
  });

  res.json(movements);
});
```

## Product Variant Patterns

### Size + Color Matrix

```typescript
// Generate all combinations
function generateVariants(options: Record<string, string[]>) {
  const keys = Object.keys(options);
  const combinations: Record<string, string>[] = [{}];

  for (const key of keys) {
    const values = options[key];
    const newCombinations: Record<string, string>[] = [];
    for (const combo of combinations) {
      for (const value of values) {
        newCombinations.push({ ...combo, [key]: value });
      }
    }
    combinations.length = 0;
    combinations.push(...newCombinations);
  }

  return combinations.map(combo => ({
    name: Object.values(combo).join(' / '),
    sku: `${Object.values(combo).map(v => v.toUpperCase().replace(/\s/g, '-')).join('-')}`,
    options: combo,
  }));
}

// Example:
generateVariants({
  size: ['S', 'M', 'L', 'XL'],
  color: ['Red', 'Blue', 'Black'],
});
// → 12 variants: S/Red, S/Blue, S/Black, M/Red, ...
```

## Best Practices

1. **Always use transactions** for stock operations — never update quantity and movements separately
2. **Reserve → Commit pattern** — reserve on order creation, commit on payment confirmation
3. **Log every movement** — full audit trail of why stock changed
4. **Set low stock thresholds** per product/variant, not globally
5. **Allow backorders** as a per-product setting, not all-or-nothing
6. **SKU format**: `CATEGORY-PRODUCT-VARIANT` (e.g., `TEE-BASIC-L-BLU`)
7. **Periodic reconciliation** — schedule inventory counts and compare to system records