# PostgreSQL Deep Dive Reference

Comprehensive reference for PostgreSQL-specific features, internals, and advanced patterns.
This document covers what makes PostgreSQL unique and powerful beyond standard SQL.

## MVCC (Multi-Version Concurrency Control)

### How MVCC Works

PostgreSQL never overwrites data in place. Every UPDATE creates a new row version (tuple).
Every DELETE marks a row as invisible to future transactions.

**Tuple header fields:**
- `xmin` — Transaction ID that created this tuple version
- `xmax` — Transaction ID that deleted/updated this tuple (0 if live)
- `cmin/cmax` — Command ID within the transaction
- `ctid` — Physical location (page number, offset) pointing to the latest version

**Visibility rules:**
A tuple is visible to a transaction if:
1. `xmin` is committed AND `xmin` started before current transaction's snapshot
2. `xmax` is either 0, aborted, or started after current transaction's snapshot

```sql
-- See tuple metadata
CREATE EXTENSION IF NOT EXISTS pageinspect;

SELECT
    lp AS line_pointer,
    t_xmin AS xmin,
    t_xmax AS xmax,
    t_ctid AS ctid,
    t_infomask::bit(16) AS infomask
FROM heap_page_items(get_raw_page('users', 0));

-- See transaction ID
SELECT txid_current();

-- See snapshot
SELECT txid_current_snapshot();
-- Returns: xmin:xmax:xip_list
-- xmin = oldest active transaction
-- xmax = next transaction ID to be assigned
-- xip_list = currently active transaction IDs
```

### Transaction Isolation Levels

```sql
-- Read Committed (default) — sees committed data at statement start
SET TRANSACTION ISOLATION LEVEL READ COMMITTED;

-- Repeatable Read — sees committed data at transaction start
SET TRANSACTION ISOLATION LEVEL REPEATABLE READ;
-- Difference: within a transaction, repeated reads return same data
-- Serialization errors possible: "could not serialize access due to concurrent update"

-- Serializable — strongest isolation, as if transactions ran sequentially
SET TRANSACTION ISOLATION LEVEL SERIALIZABLE;
-- Uses Serializable Snapshot Isolation (SSI)
-- May throw: "could not serialize access due to read/write dependencies"
-- Application must retry on serialization failure

-- PostgreSQL does NOT have Read Uncommitted — it's treated as Read Committed
```

### MVCC Implications

```
1. Readers never block writers, writers never block readers
2. Dead rows (old versions) accumulate — VACUUM must clean them up
3. Long-running transactions prevent dead row cleanup (hold back xmin horizon)
4. Transaction ID wraparound: must vacuum before reaching 2 billion transactions
5. HOT updates (Heap-Only Tuples): if update doesn't change indexed columns
   AND new version fits on same page, no index update needed — much faster
```

## WAL (Write-Ahead Log)

### WAL Mechanics

Every data change is first written to WAL before modifying data pages.
This ensures crash recovery: replay WAL from last checkpoint to recover.

```
Write sequence:
1. Change written to WAL buffer (in shared memory)
2. WAL buffer flushed to WAL segment file on disk (at commit)
3. Data page modified in shared buffer pool (in memory)
4. Background writer / checkpointer flushes dirty pages to disk

Crash recovery:
1. Start from last checkpoint
2. Replay all WAL records after checkpoint
3. Database is consistent after replay
```

### WAL Configuration

```ini
# WAL segment files (default 16MB each)
wal_level = replica          # minimal, replica, or logical
max_wal_size = 4GB           # Checkpoint triggered when WAL reaches this size
min_wal_size = 1GB           # Don't remove WAL below this total size
checkpoint_timeout = 15min   # Max time between automatic checkpoints
checkpoint_completion_target = 0.9  # Spread I/O over this fraction of interval

# WAL archiving (for point-in-time recovery)
archive_mode = on
archive_command = 'cp %p /archive/%f'

# WAL compression (saves I/O and network for replication)
wal_compression = zstd       # lz4, pglz, zstd, or off (PG 15+)

# Full page writes (first modification after checkpoint writes full page)
full_page_writes = on        # Required for crash safety, can't disable safely
```

### Monitoring WAL

```sql
-- Current WAL position
SELECT pg_current_wal_lsn();

-- WAL usage per statement
EXPLAIN (ANALYZE, WAL) UPDATE users SET name = 'test' WHERE id = 1;

-- WAL statistics
SELECT * FROM pg_stat_wal;
-- wal_records, wal_fpi (full page images), wal_bytes, wal_write_time, wal_sync_time

-- Checkpoint statistics
SELECT * FROM pg_stat_bgwriter;
-- checkpoints_timed, checkpoints_req, checkpoint_write_time, checkpoint_sync_time
```

## TOAST (The Oversized-Attribute Storage Technique)

### How TOAST Works

PostgreSQL pages are fixed at 8KB. Values larger than ~2KB are compressed and/or stored
in a separate TOAST table.

**TOAST strategies:**
```sql
-- PLAIN: No compression, no out-of-line storage (for fixed-width types)
-- EXTENDED: Compress first, then out-of-line if still too large (default for variable-length)
-- EXTERNAL: Out-of-line without compression (faster for pre-compressed data)
-- MAIN: Compress but avoid out-of-line if possible

-- Change TOAST strategy for a column
ALTER TABLE articles ALTER COLUMN body SET STORAGE EXTERNAL;

-- Check TOAST table info
SELECT
    c.relname AS table_name,
    t.relname AS toast_table,
    pg_size_pretty(pg_relation_size(t.oid)) AS toast_size
FROM pg_class c
JOIN pg_class t ON c.reltoastrelid = t.oid
WHERE c.relnamespace = 'public'::regnamespace
ORDER BY pg_relation_size(t.oid) DESC;
```

### TOAST Performance Implications

```
1. Selecting columns stored in TOAST requires extra I/O
   → Only SELECT columns you need (avoid SELECT *)
2. TOAST values are compressed with pglz (or lz4 in PG 14+)
   → ALTER SYSTEM SET default_toast_compression = 'lz4'; (faster)
3. TOAST tables need vacuuming too
4. Large JSONB/TEXT values are TOAST candidates
5. Index-only scans don't work if TOAST columns are requested
```

## Advanced Index Types

### Partial Indexes

```sql
-- Index only the rows you actually query
CREATE INDEX idx_orders_pending ON orders(created_at)
    WHERE status = 'pending';

-- Much smaller than full index, faster to maintain
-- Only useful for queries that include the WHERE condition

-- Partial unique index
CREATE UNIQUE INDEX idx_users_email_active ON users(email)
    WHERE deleted_at IS NULL;
-- Allows multiple deleted users with same email, but only one active
```

### Expression Indexes

```sql
-- Index on function result
CREATE INDEX idx_users_lower_email ON users(LOWER(email));

-- Index on JSONB path
CREATE INDEX idx_users_country ON users((profile->>'country'));

-- Index on date extraction
CREATE INDEX idx_orders_year_month ON orders(
    EXTRACT(YEAR FROM created_at),
    EXTRACT(MONTH FROM created_at)
);

-- Index on text operation
CREATE INDEX idx_products_first_letter ON products(LEFT(name, 1));
```

### GIN (Generalized Inverted Index)

Best for: values that contain multiple elements (arrays, JSONB, full-text search)

```sql
-- JSONB containment
CREATE INDEX idx_products_attrs ON products USING gin (attributes jsonb_path_ops);
-- jsonb_path_ops: supports @> (containment) only, but 2-3x smaller and faster
-- Default: supports @>, ?, ?|, ?&

-- Array operations
CREATE INDEX idx_posts_tags ON posts USING gin (tags);

-- Full-text search
ALTER TABLE articles ADD COLUMN search_vector tsvector
    GENERATED ALWAYS AS (
        setweight(to_tsvector('english', coalesce(title, '')), 'A') ||
        setweight(to_tsvector('english', coalesce(body, '')), 'B')
    ) STORED;
CREATE INDEX idx_articles_search ON articles USING gin (search_vector);

-- Query
SELECT * FROM articles
WHERE search_vector @@ plainto_tsquery('english', 'database optimization');

-- Ranking
SELECT
    title,
    ts_rank(search_vector, plainto_tsquery('english', 'database optimization')) AS rank
FROM articles
WHERE search_vector @@ plainto_tsquery('english', 'database optimization')
ORDER BY rank DESC;

-- Trigram similarity search (fuzzy matching)
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX idx_products_name_trgm ON products USING gin (name gin_trgm_ops);

SELECT * FROM products WHERE name % 'wireles mouse';  -- Fuzzy match
SELECT * FROM products WHERE name ILIKE '%wireless%';  -- ILIKE uses trigram index
SELECT similarity(name, 'wireles mouse') AS sim, name
FROM products ORDER BY sim DESC LIMIT 10;
```

### GiST (Generalized Search Tree)

Best for: spatial data, ranges, nearest-neighbor queries

```sql
-- Range types
CREATE TABLE reservations (
    id SERIAL PRIMARY KEY,
    room_id INT NOT NULL,
    during TSTZRANGE NOT NULL,
    EXCLUDE USING gist (room_id WITH =, during WITH &&)
    -- Prevents overlapping reservations for same room
);

INSERT INTO reservations (room_id, during) VALUES
    (1, '[2024-01-15 09:00, 2024-01-15 11:00)');

-- This will fail (overlaps):
INSERT INTO reservations (room_id, during) VALUES
    (1, '[2024-01-15 10:00, 2024-01-15 12:00)');

-- Range queries
SELECT * FROM reservations
WHERE during && '[2024-01-15, 2024-01-16)'::tstzrange;  -- Overlaps date range

-- Nearest-neighbor with GiST
CREATE INDEX idx_locations_point ON locations USING gist (point);

SELECT *, point <-> '(40.7128, -74.0060)'::point AS distance
FROM locations
ORDER BY point <-> '(40.7128, -74.0060)'::point
LIMIT 10;
```

### BRIN (Block Range Index)

Best for: naturally ordered data (timestamps, auto-increment IDs) in large tables

```sql
CREATE INDEX idx_logs_created ON logs USING brin (created_at)
    WITH (pages_per_range = 32);

-- BRIN stores min/max per block range (128 pages by default)
-- Size: ~1000x smaller than B-tree
-- Speed: slower than B-tree for point queries, good for range scans
-- Best when: data is physically ordered by the indexed column

-- Check correlation (1.0 = perfectly ordered, 0.0 = random)
SELECT
    attname,
    correlation
FROM pg_stats
WHERE tablename = 'logs'
  AND attname = 'created_at';
-- BRIN is effective when |correlation| > 0.9

-- BRIN supports multiple data types
CREATE INDEX idx_logs_multi ON logs USING brin (created_at, server_id, level);
```

### Hash Indexes

```sql
-- Hash index: exact equality lookups only (no range, no sorting)
CREATE INDEX idx_sessions_token ON sessions USING hash (token);

-- PostgreSQL 10+ hash indexes are WAL-logged and crash-safe
-- Slightly smaller than B-tree for exact lookups
-- BUT: no range scans, no ordering, no multi-column
-- Rarely better than B-tree in practice — use B-tree unless benchmarks show improvement
```

## Full-Text Search

### tsvector and tsquery

```sql
-- Create a search configuration
-- Built-in: 'simple', 'english', 'spanish', 'german', 'french', etc.

-- Generate tsvector (document representation)
SELECT to_tsvector('english', 'The quick brown fox jumps over the lazy dog');
-- 'brown':3 'dog':9 'fox':4 'jump':5 'lazi':8 'quick':2

-- Generate tsquery (search query)
SELECT plainto_tsquery('english', 'quick brown fox');
-- 'quick' & 'brown' & 'fox'

SELECT phraseto_tsquery('english', 'quick brown fox');
-- 'quick' <-> 'brown' <-> 'fox'  (phrase search, words must be adjacent)

SELECT websearch_to_tsquery('english', '"quick brown" OR lazy -cat');
-- 'quick' <-> 'brown' | 'lazi' & !'cat'

-- Weighted search vector
SELECT
    setweight(to_tsvector('english', 'Database Optimization Guide'), 'A') ||
    setweight(to_tsvector('english', 'Learn how to optimize PostgreSQL queries for maximum performance'), 'B');
```

### Full-Text Search Setup

```sql
-- Add generated column for search vector
ALTER TABLE articles ADD COLUMN search_vector tsvector
    GENERATED ALWAYS AS (
        setweight(to_tsvector('english', coalesce(title, '')), 'A') ||
        setweight(to_tsvector('english', coalesce(subtitle, '')), 'B') ||
        setweight(to_tsvector('english', coalesce(body, '')), 'C') ||
        setweight(to_tsvector('english', coalesce(tags_text, '')), 'B')
    ) STORED;

CREATE INDEX idx_articles_fts ON articles USING gin (search_vector);

-- Search with ranking
SELECT
    id,
    title,
    ts_rank_cd(search_vector, query) AS rank,
    ts_headline('english', body, query, 'MaxWords=35, MinWords=15, StartSel=<b>, StopSel=</b>') AS snippet
FROM articles,
    websearch_to_tsquery('english', 'postgresql optimization') AS query
WHERE search_vector @@ query
ORDER BY rank DESC
LIMIT 20;
```

### Custom Text Search Configuration

```sql
-- Create custom configuration with synonym support
CREATE TEXT SEARCH DICTIONARY english_syn (
    TEMPLATE = synonym,
    SYNONYMS = my_synonyms  -- File: $SHAREDIR/tsearch_data/my_synonyms.syn
);

CREATE TEXT SEARCH CONFIGURATION my_english (COPY = english);

ALTER TEXT SEARCH CONFIGURATION my_english
    ALTER MAPPING FOR asciiword, asciihword, hword_asciipart
    WITH english_syn, english_stem;

-- Unaccented search (ignore diacritics)
CREATE EXTENSION IF NOT EXISTS unaccent;

CREATE TEXT SEARCH CONFIGURATION my_unaccent (COPY = english);
ALTER TEXT SEARCH CONFIGURATION my_unaccent
    ALTER MAPPING FOR hword, hword_part, word
    WITH unaccent, english_stem;

-- Now: to_tsvector('my_unaccent', 'café') matches 'cafe'
```

## Advisory Locks

```sql
-- Session-level advisory lock (held until session ends or explicitly released)
SELECT pg_advisory_lock(12345);      -- Blocking: waits if already locked
SELECT pg_try_advisory_lock(12345);  -- Non-blocking: returns false if can't lock
SELECT pg_advisory_unlock(12345);    -- Release

-- Transaction-level advisory lock (released at end of transaction)
SELECT pg_advisory_xact_lock(12345);
SELECT pg_try_advisory_xact_lock(12345);
-- No explicit unlock — released automatically at COMMIT/ROLLBACK

-- Two-key advisory locks (64-bit from two 32-bit integers)
SELECT pg_advisory_lock(1, 42);  -- Lock type 1, entity 42
SELECT pg_try_advisory_lock(1, 42);

-- Common patterns:

-- 1. Singleton job execution (only one worker runs this job)
SELECT pg_try_advisory_lock(hashtext('daily_report_job'));
-- If true: run the job, then unlock
-- If false: another worker is already running it

-- 2. Entity-level locking (process order 42)
SELECT pg_advisory_xact_lock(hashtext('process_order'), 42);
-- Other transactions wanting to process order 42 will wait

-- 3. Rate limiting (one operation per entity at a time)
DO $$
BEGIN
    IF NOT pg_try_advisory_xact_lock(hashtext('send_email'), user_id) THEN
        RAISE EXCEPTION 'Already sending email for this user';
    END IF;
    -- Send email...
END $$;

-- Monitor advisory locks
SELECT * FROM pg_locks WHERE locktype = 'advisory';
```

## LISTEN/NOTIFY

```sql
-- Real-time notifications between database connections

-- Listener (in one connection):
LISTEN order_events;

-- Notifier (in another connection):
NOTIFY order_events, '{"order_id": 42, "status": "shipped"}';
-- Or:
SELECT pg_notify('order_events', json_build_object('order_id', 42, 'status', 'shipped')::text);

-- Trigger-based notifications
CREATE OR REPLACE FUNCTION notify_order_change()
RETURNS TRIGGER AS $$
BEGIN
    PERFORM pg_notify(
        'order_events',
        json_build_object(
            'action', TG_OP,
            'order_id', NEW.id,
            'old_status', CASE WHEN TG_OP = 'UPDATE' THEN OLD.status ELSE NULL END,
            'new_status', NEW.status
        )::text
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_order_notify
AFTER INSERT OR UPDATE ON orders
FOR EACH ROW EXECUTE FUNCTION notify_order_change();
```

```javascript
// Node.js LISTEN example
const { Client } = require('pg');
const client = new Client();
await client.connect();

client.on('notification', (msg) => {
  const payload = JSON.parse(msg.payload);
  console.log('Order event:', payload);
});

await client.query('LISTEN order_events');
```

## Foreign Data Wrappers (FDW)

```sql
-- Query external databases as if they were local tables

-- postgres_fdw: Connect to another PostgreSQL database
CREATE EXTENSION postgres_fdw;

CREATE SERVER remote_server
    FOREIGN DATA WRAPPER postgres_fdw
    OPTIONS (host 'remote-host', port '5432', dbname 'other_db');

CREATE USER MAPPING FOR current_user
    SERVER remote_server
    OPTIONS (user 'remote_user', password 'remote_pass');

-- Import all tables from a remote schema
IMPORT FOREIGN SCHEMA public
    FROM SERVER remote_server
    INTO remote_tables;

-- Or create individual foreign tables
CREATE FOREIGN TABLE remote_users (
    id INT,
    email TEXT,
    name TEXT
) SERVER remote_server
  OPTIONS (schema_name 'public', table_name 'users');

-- Query as normal SQL
SELECT * FROM remote_users WHERE email = 'user@example.com';

-- JOIN local and remote tables
SELECT l.order_id, r.email
FROM local_orders l
JOIN remote_users r ON r.id = l.user_id;

-- file_fdw: Query flat files as tables
CREATE EXTENSION file_fdw;
CREATE SERVER file_server FOREIGN DATA WRAPPER file_fdw;

CREATE FOREIGN TABLE csv_import (
    name TEXT,
    email TEXT,
    created_at TEXT
) SERVER file_server
  OPTIONS (filename '/tmp/users.csv', format 'csv', header 'true');

SELECT * FROM csv_import;
```

## Row-Level Security (RLS)

```sql
-- Enable RLS on a table
ALTER TABLE documents ENABLE ROW LEVEL SECURITY;

-- Force RLS for table owners too (by default, owners bypass RLS)
ALTER TABLE documents FORCE ROW LEVEL SECURITY;

-- Policy: users can only see their own documents
CREATE POLICY user_documents ON documents
    FOR ALL
    USING (user_id = current_setting('app.user_id')::int);

-- Separate policies for different operations
CREATE POLICY select_docs ON documents
    FOR SELECT
    USING (
        user_id = current_setting('app.user_id')::int
        OR is_public = true
    );

CREATE POLICY insert_docs ON documents
    FOR INSERT
    WITH CHECK (user_id = current_setting('app.user_id')::int);

CREATE POLICY update_docs ON documents
    FOR UPDATE
    USING (user_id = current_setting('app.user_id')::int)
    WITH CHECK (user_id = current_setting('app.user_id')::int);

CREATE POLICY delete_docs ON documents
    FOR DELETE
    USING (user_id = current_setting('app.user_id')::int AND is_archived = false);

-- Multi-tenant RLS
CREATE POLICY tenant_isolation ON orders
    USING (tenant_id = current_setting('app.tenant_id')::int);

-- Set context per request (in application):
-- BEGIN;
-- SET LOCAL app.tenant_id = '42';
-- SET LOCAL app.user_id = '17';
-- ... queries automatically filtered ...
-- COMMIT;

-- Admin bypass policy
CREATE POLICY admin_all ON documents
    FOR ALL
    USING (current_setting('app.user_role') = 'admin');
```

## Generated Columns

```sql
-- Stored generated column (computed on write, stored on disk)
CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    price DECIMAL(10, 2) NOT NULL,
    tax_rate DECIMAL(5, 4) NOT NULL DEFAULT 0.0875,
    total_price DECIMAL(10, 2) GENERATED ALWAYS AS (price * (1 + tax_rate)) STORED,
    name TEXT NOT NULL,
    name_lower TEXT GENERATED ALWAYS AS (LOWER(name)) STORED
);

-- Can be indexed
CREATE INDEX idx_products_total ON products(total_price);
CREATE INDEX idx_products_name_lower ON products(name_lower);

-- Virtual generated columns (computed on read, not stored) — PostgreSQL 12+ STORED only
-- Virtual columns are planned for future PostgreSQL versions
```

## Identity Columns (Modern Alternative to SERIAL)

```sql
-- GENERATED ALWAYS: PostgreSQL controls the value, overrides not allowed (strict)
CREATE TABLE users (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    email TEXT NOT NULL
);

-- GENERATED BY DEFAULT: allows explicit values but auto-generates if omitted
CREATE TABLE users (
    id INT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    email TEXT NOT NULL
);

-- Custom sequence options
CREATE TABLE events (
    id BIGINT GENERATED ALWAYS AS IDENTITY (START WITH 1000 INCREMENT BY 1) PRIMARY KEY,
    event_type TEXT NOT NULL
);

-- Override GENERATED ALWAYS (for data imports):
INSERT INTO users (id, email) OVERRIDING SYSTEM VALUE VALUES (999, 'admin@example.com');
```

## pg_cron (Scheduled Jobs)

```sql
-- Install extension (requires shared_preload_libraries)
CREATE EXTENSION pg_cron;

-- Schedule vacuum every night at 3 AM
SELECT cron.schedule('nightly-vacuum', '0 3 * * *', 'VACUUM ANALYZE');

-- Refresh materialized view every hour
SELECT cron.schedule('refresh-mv', '0 * * * *',
    'REFRESH MATERIALIZED VIEW CONCURRENTLY mv_monthly_revenue');

-- Clean up old sessions every 15 minutes
SELECT cron.schedule('clean-sessions', '*/15 * * * *',
    'DELETE FROM sessions WHERE expires_at < NOW()');

-- Archive old data daily
SELECT cron.schedule('archive-events', '0 2 * * *', $$
    WITH archived AS (
        DELETE FROM events
        WHERE created_at < NOW() - INTERVAL '90 days'
        RETURNING *
    )
    INSERT INTO events_archive SELECT * FROM archived
$$);

-- List scheduled jobs
SELECT * FROM cron.job;

-- Unschedule a job
SELECT cron.unschedule('nightly-vacuum');

-- View job run history
SELECT * FROM cron.job_run_details ORDER BY start_time DESC LIMIT 20;
```

## pgvector (Vector Embeddings)

```sql
-- Install extension
CREATE EXTENSION vector;

-- Create table with vector column
CREATE TABLE documents (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    embedding vector(1536)  -- OpenAI ada-002 dimensions
);

-- Insert with embedding
INSERT INTO documents (title, content, embedding)
VALUES ('My Document', 'Content here...', '[0.1, 0.2, 0.3, ...]'::vector);

-- Nearest neighbor search (cosine distance)
SELECT
    id, title,
    1 - (embedding <=> '[0.1, 0.2, ...]'::vector) AS similarity
FROM documents
ORDER BY embedding <=> '[0.1, 0.2, ...]'::vector
LIMIT 10;

-- Distance operators:
-- <-> : L2 distance (Euclidean)
-- <=> : cosine distance
-- <#> : inner product (negative, so ORDER BY ascending)

-- Create index for approximate nearest neighbor (ANN)
CREATE INDEX idx_documents_embedding ON documents
    USING ivfflat (embedding vector_cosine_ops)
    WITH (lists = 100);

-- Or HNSW index (better recall, more memory)
CREATE INDEX idx_documents_embedding_hnsw ON documents
    USING hnsw (embedding vector_cosine_ops)
    WITH (m = 16, ef_construction = 64);

-- Hybrid search: combine full-text and vector similarity
SELECT
    d.id, d.title,
    ts_rank(d.search_vector, query) AS text_rank,
    1 - (d.embedding <=> query_embedding) AS vector_similarity,
    -- Weighted hybrid score
    0.3 * ts_rank(d.search_vector, query) +
    0.7 * (1 - (d.embedding <=> query_embedding)) AS hybrid_score
FROM documents d,
    plainto_tsquery('english', 'database optimization') AS query,
    '[0.1, 0.2, ...]'::vector AS query_embedding
WHERE d.search_vector @@ query
ORDER BY hybrid_score DESC
LIMIT 10;
```

## PostGIS Basics

```sql
-- Install extension
CREATE EXTENSION postgis;

-- Create table with geometry column
CREATE TABLE places (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    location GEOGRAPHY(POINT, 4326)  -- WGS 84 coordinate system
);

-- Insert a point (longitude, latitude)
INSERT INTO places (name, location)
VALUES ('Empire State Building', ST_MakePoint(-73.9857, 40.7484)::geography);

-- Find places within 1km of a point
SELECT name, ST_Distance(location, ST_MakePoint(-73.9857, 40.7484)::geography) AS distance_meters
FROM places
WHERE ST_DWithin(location, ST_MakePoint(-73.9857, 40.7484)::geography, 1000)
ORDER BY distance_meters;

-- Spatial index
CREATE INDEX idx_places_location ON places USING gist (location);

-- Bounding box query
SELECT * FROM places
WHERE location && ST_MakeEnvelope(-74.05, 40.70, -73.90, 40.80, 4326)::geography;
```

## Logical Replication

```sql
-- Publisher (source)
ALTER SYSTEM SET wal_level = logical;
-- Restart required

CREATE PUBLICATION my_pub FOR TABLE users, orders;
-- Or: CREATE PUBLICATION my_pub FOR ALL TABLES;

-- Subscriber (destination)
CREATE SUBSCRIPTION my_sub
    CONNECTION 'host=publisher-host dbname=myapp user=replicator'
    PUBLICATION my_pub;

-- Monitor replication
SELECT * FROM pg_stat_subscription;
SELECT * FROM pg_stat_replication;  -- On publisher
SELECT * FROM pg_replication_slots;

-- Selective replication (column filter, PG 15+)
CREATE PUBLICATION pub_partial FOR TABLE users (id, email, name);  -- Only these columns

-- Row filter (PG 15+)
CREATE PUBLICATION pub_active FOR TABLE users WHERE (status = 'active');

-- Manage publications
ALTER PUBLICATION my_pub ADD TABLE products;
ALTER PUBLICATION my_pub DROP TABLE old_table;

-- Refresh subscription after adding tables
ALTER SUBSCRIPTION my_sub REFRESH PUBLICATION;
```

## Performance Tips Summary

### Quick Wins

```sql
-- 1. Update statistics (fix bad query plans)
ANALYZE;

-- 2. Find missing indexes
SELECT
    relname AS table,
    seq_scan,
    seq_tup_read,
    idx_scan,
    CASE WHEN seq_scan > 0
        THEN seq_tup_read / seq_scan
        ELSE 0
    END AS avg_rows_per_seq_scan
FROM pg_stat_user_tables
WHERE seq_scan > 100
ORDER BY seq_tup_read DESC;

-- 3. Find unused indexes (remove to speed up writes)
SELECT indexrelname, idx_scan, pg_size_pretty(pg_relation_size(indexrelid))
FROM pg_stat_user_indexes
WHERE idx_scan = 0
  AND indexrelname NOT LIKE '%pkey%'
ORDER BY pg_relation_size(indexrelid) DESC;

-- 4. Cache hit ratio (should be 99%+)
SELECT
    sum(blks_hit) * 100.0 / sum(blks_hit + blks_read) AS cache_hit_ratio
FROM pg_stat_database;

-- 5. Kill long-running queries
SELECT pg_terminate_backend(pid)
FROM pg_stat_activity
WHERE state = 'active'
  AND NOW() - query_start > INTERVAL '5 minutes'
  AND pid != pg_backend_pid();
```

### Common Configuration Mistakes

```
1. max_connections too high (>200 without pgbouncer)
   → Use connection pooling, keep max_connections = 100-200

2. shared_buffers too low or too high
   → 25% of RAM, max ~8-16GB

3. work_mem too high globally
   → Set low globally, increase per-session for analytics queries

4. random_page_cost = 4.0 on SSD
   → Set to 1.1 for SSD

5. Autovacuum too conservative
   → Lower scale_factor for large tables

6. No statement_timeout
   → Set to 30-60 seconds to prevent runaway queries

7. synchronous_commit = on for non-critical writes
   → Consider off for logging/analytics tables

8. Not using CONCURRENTLY for index creation
   → Always use CREATE INDEX CONCURRENTLY in production
```
