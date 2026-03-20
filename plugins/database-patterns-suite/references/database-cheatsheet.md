# Database Patterns Cheatsheet

## PostgreSQL Data Types

| Type | Use For | Example |
|------|---------|---------|
| `UUID` | Primary keys, unique IDs | `gen_random_uuid()` |
| `TEXT` | Variable-length strings | Names, emails, URLs |
| `VARCHAR(n)` | Length-limited strings | Country codes, zip codes |
| `INT` / `BIGINT` | Integers | Counts, IDs (serial) |
| `NUMERIC(p,s)` | Exact decimals | Money: `NUMERIC(10,2)` |
| `BOOLEAN` | True/false | Flags, toggles |
| `TIMESTAMPTZ` | Date + time + timezone | Created/updated timestamps |
| `DATE` | Date only | Birthdays, deadlines |
| `JSONB` | Structured flexible data | Settings, metadata |
| `TEXT[]` | Array of strings | Tags, categories |
| `INET` | IP addresses | Access logs |
| `TSVECTOR` | Full-text search | Search index column |

## Index Types

| Type | Syntax | Best For |
|------|--------|----------|
| B-tree (default) | `CREATE INDEX ON t(col)` | `=`, `<`, `>`, `BETWEEN`, `IN`, `IS NULL` |
| Hash | `CREATE INDEX ON t USING hash(col)` | `=` only (rarely better than B-tree) |
| GIN | `CREATE INDEX ON t USING gin(col)` | Arrays, JSONB, full-text, trigrams |
| GiST | `CREATE INDEX ON t USING gist(col)` | Geometric, range types, nearest-neighbor |
| BRIN | `CREATE INDEX ON t USING brin(col)` | Large tables with natural ordering (timestamps) |
| Partial | `CREATE INDEX ON t(col) WHERE condition` | Filtered subsets (active records) |
| Covering | `CREATE INDEX ON t(a) INCLUDE (b, c)` | Index-only scans with extra columns |
| Expression | `CREATE INDEX ON t(LOWER(col))` | Function results in WHERE |
| Composite | `CREATE INDEX ON t(a, b, c)` | Multi-column queries (leftmost prefix rule) |

## EXPLAIN Cheatsheet

```sql
EXPLAIN (ANALYZE, BUFFERS, FORMAT TEXT) SELECT ...;
```

| Node | Meaning | Action |
|------|---------|--------|
| Seq Scan | Full table scan | Add index |
| Index Scan | B-tree lookup | Good |
| Index Only Scan | Index covers all columns | Best |
| Bitmap Index Scan | Multiple conditions combined | Good for OR |
| Nested Loop | Join: row-by-row | OK for small tables |
| Hash Join | Join: hash table | OK for medium tables |
| Merge Join | Join: pre-sorted | OK for large tables |
| Sort | ORDER BY without index | Add matching index |
| Aggregate | GROUP BY / COUNT / SUM | Normal |
| Limit | LIMIT applied | Good |

**Red flags:**
- `actual rows` >> `estimated rows` → run `ANALYZE tablename`
- `loops` > 100 → N+1 query, rewrite with JOIN
- `Sort Method: external merge Disk` → increase `work_mem`

## Connection String Format

```
postgresql://user:password@host:5432/dbname?sslmode=require

# With connection pool
postgresql://user:password@host:6432/dbname  # PgBouncer port

# Heroku
DATABASE_URL=postgres://user:pass@host:5432/dbname
```

## Common Constraints

```sql
-- Primary key
id UUID PRIMARY KEY DEFAULT gen_random_uuid()

-- Foreign key with cascade
user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE

-- Unique constraint
email TEXT NOT NULL UNIQUE

-- Composite unique
UNIQUE(org_id, email)

-- Check constraint
status TEXT NOT NULL CHECK (status IN ('active', 'inactive', 'banned'))
age INT CHECK (age >= 0 AND age <= 150)

-- Not null with default
created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
is_active BOOLEAN NOT NULL DEFAULT true
```

## Transaction Isolation Levels

| Level | Dirty Read | Non-repeatable Read | Phantom Read | Use Case |
|-------|-----------|-------------------|-------------|----------|
| Read Uncommitted | Yes | Yes | Yes | Never use in Postgres (treated as Read Committed) |
| Read Committed | No | Yes | Yes | Default. Good for most OLTP |
| Repeatable Read | No | No | No* | Financial calculations, reports |
| Serializable | No | No | No | Strictest. Use for critical consistency |

*PostgreSQL's Repeatable Read also prevents phantom reads (unlike the SQL standard).

```sql
BEGIN ISOLATION LEVEL SERIALIZABLE;
-- ... queries ...
COMMIT;
```

## Redis Commands Quick Reference

| Command | Use | Example |
|---------|-----|---------|
| `SET key value EX ttl` | Cache with TTL | `SET user:1 '{"name":"Alice"}' EX 300` |
| `GET key` | Read cache | `GET user:1` |
| `DEL key [key ...]` | Invalidate | `DEL user:1 user:2` |
| `SETEX key ttl value` | Set with TTL (atomic) | `SETEX session:abc 1800 '{...}'` |
| `SETNX key value` | Set if not exists (mutex) | `SETNX lock:order:1 1` |
| `INCR key` | Atomic counter | `INCR page:views:/home` |
| `EXPIRE key seconds` | Set TTL on existing key | `EXPIRE user:1 600` |
| `TTL key` | Check remaining TTL | `TTL user:1` → `245` |
| `KEYS pattern` | Find keys (SLOW!) | `KEYS user:*` — use SCAN instead |
| `SCAN cursor MATCH pattern` | Safe key iteration | `SCAN 0 MATCH user:* COUNT 100` |
| `HSET key field value` | Hash field set | `HSET user:1 name Alice email a@b.com` |
| `HGET key field` | Hash field get | `HGET user:1 name` |
| `HMGET key f1 f2` | Multiple hash fields | `HMGET user:1 name email` |
| `LPUSH key value` | Add to list head | `LPUSH queue:jobs '{"type":"email"}'` |
| `RPOP key` | Remove from list tail | `RPOP queue:jobs` |
| `BRPOP key timeout` | Blocking pop (worker) | `BRPOP queue:jobs 0` |
| `SADD key member` | Add to set | `SADD online:users user:1` |
| `SMEMBERS key` | Get all set members | `SMEMBERS online:users` |
| `ZADD key score member` | Sorted set add | `ZADD leaderboard 100 user:1` |
| `ZREVRANGE key 0 9` | Top 10 by score | `ZREVRANGE leaderboard 0 9 WITHSCORES` |
| `PUBLISH channel msg` | Pub/sub publish | `PUBLISH events '{"type":"order"}'` |
| `SUBSCRIBE channel` | Pub/sub subscribe | `SUBSCRIBE events` |
| `PFADD key element` | HyperLogLog add | `PFADD visitors:today user:1` |
| `PFCOUNT key` | Approximate count | `PFCOUNT visitors:today` |

## Pool Sizing Rules

| Scenario | Recommended Pool Size |
|----------|----------------------|
| Web app (4 CPU cores) | 10-20 connections |
| API server (8 cores) | 20-40 connections |
| Background workers | 5-10 per worker process |
| Connection via PgBouncer | 100-1000 client, 20-40 server |

Formula: `connections = (CPU cores * 2) + effective_spindle_count`

## Migration Safety Checklist

| Operation | Safe? | Notes |
|-----------|-------|-------|
| ADD COLUMN | Yes | No lock in PG 11+ with default |
| ADD COLUMN NOT NULL | Caution | Add nullable first, backfill, then add constraint |
| DROP COLUMN | Yes | But verify no code references it |
| RENAME COLUMN | Caution | Create view for backward compatibility |
| ADD INDEX | Use CONCURRENTLY | `CREATE INDEX CONCURRENTLY` doesn't lock writes |
| DROP INDEX | Use CONCURRENTLY | `DROP INDEX CONCURRENTLY` doesn't lock reads |
| ADD FOREIGN KEY | Caution | `NOT VALID` first, then `VALIDATE` separately |
| ADD CHECK CONSTRAINT | Caution | Same: `NOT VALID` then `VALIDATE` |
| ALTER COLUMN TYPE | Dangerous | Rewrites entire table, exclusive lock |
| DROP TABLE | Dangerous | Irreversible, check dependencies first |

## Prisma Quick Reference

```bash
npx prisma init                    # Initialize Prisma
npx prisma db pull                 # Introspect existing database
npx prisma generate                # Generate client
npx prisma migrate dev --name init # Create and apply migration
npx prisma migrate deploy          # Apply pending migrations (production)
npx prisma studio                  # Visual database browser
npx prisma db seed                 # Run seed script
```

```prisma
model User {
  id        String   @id @default(uuid())
  email     String   @unique
  name      String
  role      Role     @default(MEMBER)
  posts     Post[]
  createdAt DateTime @default(now()) @map("created_at")
  updatedAt DateTime @updatedAt @map("updated_at")

  @@map("users")
  @@index([email])
}

enum Role {
  OWNER
  ADMIN
  MEMBER
}
```
