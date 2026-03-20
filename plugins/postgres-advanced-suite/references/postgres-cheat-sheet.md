# PostgreSQL Cheat Sheet

## Data Types

### Commonly Used Types

| Type | Description | Example |
|------|-------------|---------|
| `INTEGER` / `INT` | 32-bit integer (-2B to 2B) | `42` |
| `BIGINT` | 64-bit integer | `9223372036854775807` |
| `SERIAL` | Auto-increment integer | Generates sequence |
| `BIGSERIAL` | Auto-increment bigint | Generates sequence |
| `TEXT` | Variable-length string (unlimited) | `'hello world'` |
| `VARCHAR(n)` | Variable-length string (max n) | `'hello'` |
| `BOOLEAN` | true/false | `true`, `false` |
| `TIMESTAMPTZ` | Timestamp with timezone | `'2025-01-01 12:00:00+00'` |
| `TIMESTAMP` | Timestamp without timezone | `'2025-01-01 12:00:00'` |
| `DATE` | Date only | `'2025-01-01'` |
| `NUMERIC(p,s)` | Exact decimal (precision, scale) | `123.45` |
| `REAL` / `FLOAT4` | 32-bit floating point | `3.14` |
| `DOUBLE PRECISION` | 64-bit floating point | `3.14159265358979` |
| `JSONB` | Binary JSON (indexable) | `'{"key": "value"}'` |
| `JSON` | Text JSON (not indexable) | `'{"key": "value"}'` |
| `UUID` | Universally unique identifier | `gen_random_uuid()` |
| `INET` | IP address | `'192.168.1.1'` |
| `CIDR` | IP network | `'192.168.1.0/24'` |
| `BYTEA` | Binary data | `'\xDEADBEEF'` |
| `TEXT[]` | Array of text | `ARRAY['a', 'b', 'c']` |
| `INT[]` | Array of integers | `ARRAY[1, 2, 3]` |
| `INTERVAL` | Time duration | `'2 hours 30 minutes'` |
| `MONEY` | Currency (avoid — use NUMERIC) | `'$12.34'` |

### Type Recommendations

- **Use `TEXT`** over `VARCHAR(n)` — there's no performance difference in PostgreSQL, and `TEXT` avoids arbitrary length constraints.
- **Use `TIMESTAMPTZ`** over `TIMESTAMP` — always store with timezone. Convert on display.
- **Use `BIGINT`** for IDs if you expect > 2 billion rows.
- **Use `JSONB`** over `JSON` — `JSONB` is faster to query and supports indexing.
- **Use `NUMERIC`** for money — never `REAL`, `DOUBLE PRECISION`, or `MONEY`.
- **Use `UUID`** (`gen_random_uuid()`) for public-facing IDs — don't expose sequential IDs.

## JSONB Operations

```sql
-- Access key (returns JSON)
SELECT data->'name' FROM users;

-- Access key (returns text)
SELECT data->>'name' FROM users;

-- Nested access
SELECT data->'address'->>'city' FROM users;

-- Contains
SELECT * FROM users WHERE data @> '{"role": "admin"}';

-- Key exists
SELECT * FROM users WHERE data ? 'email';

-- Any key exists
SELECT * FROM users WHERE data ?| ARRAY['email', 'phone'];

-- All keys exist
SELECT * FROM users WHERE data ?& ARRAY['email', 'phone'];

-- Set/update key
UPDATE users SET data = data || '{"role": "admin"}';

-- Remove key
UPDATE users SET data = data - 'role';

-- Remove nested key
UPDATE users SET data = data #- '{address,zip}';
```

## Array Operations

```sql
-- Contains
SELECT * FROM posts WHERE tags @> ARRAY['javascript'];

-- Is contained by
SELECT * FROM posts WHERE tags <@ ARRAY['javascript', 'react', 'node'];

-- Overlap (any in common)
SELECT * FROM posts WHERE tags && ARRAY['react', 'vue'];

-- Append
UPDATE posts SET tags = tags || ARRAY['new-tag'];

-- Remove element
UPDATE posts SET tags = array_remove(tags, 'old-tag');

-- Unnest (expand to rows)
SELECT unnest(tags) AS tag FROM posts;

-- Aggregate into array
SELECT author_id, array_agg(title) FROM posts GROUP BY author_id;
```

## Window Functions

```sql
-- Row number (ranking)
SELECT name, salary,
  ROW_NUMBER() OVER (ORDER BY salary DESC) AS rank
FROM employees;

-- Rank with gaps
SELECT name, department, salary,
  RANK() OVER (PARTITION BY department ORDER BY salary DESC) AS dept_rank
FROM employees;

-- Dense rank (no gaps)
SELECT name, salary,
  DENSE_RANK() OVER (ORDER BY salary DESC) AS rank
FROM employees;

-- Running total
SELECT date, amount,
  SUM(amount) OVER (ORDER BY date) AS running_total
FROM transactions;

-- Moving average (last 7 rows)
SELECT date, amount,
  AVG(amount) OVER (ORDER BY date ROWS BETWEEN 6 PRECEDING AND CURRENT ROW) AS moving_avg
FROM daily_revenue;

-- Previous/next row values
SELECT date, amount,
  LAG(amount) OVER (ORDER BY date) AS prev_amount,
  LEAD(amount) OVER (ORDER BY date) AS next_amount,
  amount - LAG(amount) OVER (ORDER BY date) AS change
FROM daily_revenue;
```

## Common Table Expressions (CTEs)

```sql
-- Basic CTE
WITH active_users AS (
  SELECT * FROM users WHERE status = 'active'
)
SELECT au.name, COUNT(p.id) AS post_count
FROM active_users au
LEFT JOIN posts p ON p.author_id = au.id
GROUP BY au.id, au.name;

-- Recursive CTE (hierarchical data)
WITH RECURSIVE category_tree AS (
  -- Base case: top-level categories
  SELECT id, name, parent_id, 0 AS depth, name::text AS path
  FROM categories
  WHERE parent_id IS NULL

  UNION ALL

  -- Recursive case: child categories
  SELECT c.id, c.name, c.parent_id, ct.depth + 1,
         (ct.path || ' > ' || c.name)::text
  FROM categories c
  JOIN category_tree ct ON c.parent_id = ct.id
)
SELECT * FROM category_tree ORDER BY path;
```

## Useful Functions

```sql
-- String
length('hello')              -- 5
upper('hello')               -- 'HELLO'
lower('HELLO')               -- 'hello'
trim('  hello  ')            -- 'hello'
substring('hello' FROM 2 FOR 3)  -- 'ell'
replace('hello', 'l', 'r')  -- 'herro'
concat('hello', ' ', 'world')  -- 'hello world'
split_part('a.b.c', '.', 2) -- 'b'
regexp_replace('abc123', '[0-9]', '', 'g')  -- 'abc'

-- Date/Time
now()                        -- Current timestamp with timezone
CURRENT_DATE                 -- Current date
CURRENT_TIMESTAMP            -- Current timestamp
date_trunc('month', now())   -- First of current month
EXTRACT(YEAR FROM now())     -- Current year
age(timestamp1, timestamp2)  -- Interval between
now() - INTERVAL '7 days'    -- 7 days ago
to_char(now(), 'YYYY-MM-DD HH24:MI:SS')  -- Format timestamp

-- Conditional
COALESCE(a, b, c)            -- First non-null value
NULLIF(a, b)                 -- NULL if a = b, else a
GREATEST(a, b, c)            -- Maximum value
LEAST(a, b, c)               -- Minimum value

-- Aggregation
COUNT(*)                     -- Count all rows
COUNT(DISTINCT column)       -- Count unique values
SUM(amount)                  -- Total
AVG(amount)                  -- Average
MAX(value)                   -- Maximum
MIN(value)                   -- Minimum
string_agg(name, ', ')       -- Concatenate strings
array_agg(value)             -- Collect into array
json_agg(row)                -- Collect into JSON array
jsonb_build_object('k', v)   -- Build JSON object

-- UUID
gen_random_uuid()            -- Generate UUID v4
```

## Administrative Commands

```sql
-- Database
CREATE DATABASE mydb;
DROP DATABASE mydb;
\l                           -- List databases (psql)

-- Schema
CREATE SCHEMA myschema;
SET search_path TO myschema, public;

-- Table info
\d tablename                 -- Describe table (psql)
\dt                          -- List tables (psql)
\di                          -- List indexes (psql)

-- Size
SELECT pg_size_pretty(pg_database_size('mydb'));        -- Database size
SELECT pg_size_pretty(pg_total_relation_size('users')); -- Table + indexes
SELECT pg_size_pretty(pg_relation_size('users'));       -- Table only

-- Active queries
SELECT pid, state, query, now() - query_start AS duration
FROM pg_stat_activity
WHERE datname = current_database()
  AND state != 'idle'
ORDER BY duration DESC;

-- Kill a query
SELECT pg_cancel_backend(pid);    -- Graceful (cancels query)
SELECT pg_terminate_backend(pid); -- Force (kills connection)

-- Locks
SELECT * FROM pg_locks WHERE NOT granted;

-- Settings
SHOW ALL;
SHOW shared_buffers;
SHOW work_mem;
ALTER SYSTEM SET work_mem = '64MB';
SELECT pg_reload_conf();  -- Reload without restart
```

## psql Commands

```
\c dbname          -- Connect to database
\l                 -- List databases
\dt                -- List tables
\d tablename       -- Describe table
\di                -- List indexes
\df                -- List functions
\du                -- List roles
\dn                -- List schemas
\x                 -- Toggle expanded display
\timing            -- Toggle query timing
\i filename.sql    -- Execute SQL file
\copy              -- Copy data to/from CSV
\q                 -- Quit

-- Import CSV
\copy users(name, email) FROM 'users.csv' WITH (FORMAT csv, HEADER true);

-- Export CSV
\copy (SELECT * FROM users) TO 'users.csv' WITH (FORMAT csv, HEADER true);
```

## Backup and Restore

```bash
# Dump single database
pg_dump -Fc mydb > mydb.dump           # Custom format (compressed)
pg_dump -Fp mydb > mydb.sql            # Plain SQL

# Dump specific tables
pg_dump -Fc -t users -t orders mydb > partial.dump

# Dump schema only (no data)
pg_dump -Fc --schema-only mydb > schema.dump

# Dump data only (no schema)
pg_dump -Fc --data-only mydb > data.dump

# Restore
pg_restore -d mydb mydb.dump           # Custom format
psql mydb < mydb.sql                   # Plain SQL

# Restore to a different database
createdb newdb
pg_restore -d newdb mydb.dump

# Dump all databases
pg_dumpall > all_databases.sql

# Continuous archiving (WAL-based PITR)
# In postgresql.conf:
# archive_mode = on
# archive_command = 'cp %p /archive/%f'
```
