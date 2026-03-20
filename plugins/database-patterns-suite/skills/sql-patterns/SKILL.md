---
name: sql-patterns
description: >
  PostgreSQL SQL patterns and recipes — common queries, CTEs, window functions,
  JSON operations, full-text search, upserts, and advanced SQL techniques.
  Triggers: "sql patterns", "sql recipe", "postgres query", "CTE", "window function",
  "recursive query", "full text search postgres", "upsert", "lateral join",
  "json query postgres", "array operations sql".
  NOT for: schema design (use database-architect), query optimization (use query-optimizer).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# SQL Patterns & Recipes for PostgreSQL

## Common Table Expressions (CTEs)

```sql
-- Basic CTE
WITH active_users AS (
  SELECT id, name, email
  FROM users
  WHERE status = 'active'
    AND last_login > NOW() - INTERVAL '30 days'
)
SELECT u.name, COUNT(o.id) AS order_count
FROM active_users u
JOIN orders o ON o.user_id = u.id
GROUP BY u.name
ORDER BY order_count DESC;

-- Multiple CTEs
WITH
monthly_revenue AS (
  SELECT
    DATE_TRUNC('month', created_at) AS month,
    SUM(amount) AS revenue
  FROM payments
  WHERE status = 'completed'
  GROUP BY 1
),
monthly_growth AS (
  SELECT
    month,
    revenue,
    LAG(revenue) OVER (ORDER BY month) AS prev_revenue,
    ROUND(
      (revenue - LAG(revenue) OVER (ORDER BY month))
      / LAG(revenue) OVER (ORDER BY month) * 100, 2
    ) AS growth_pct
  FROM monthly_revenue
)
SELECT * FROM monthly_growth ORDER BY month DESC;

-- Recursive CTE (tree traversal)
WITH RECURSIVE category_tree AS (
  -- Base case: root categories
  SELECT id, name, parent_id, 0 AS depth, ARRAY[name] AS path
  FROM categories
  WHERE parent_id IS NULL

  UNION ALL

  -- Recursive case: children
  SELECT c.id, c.name, c.parent_id, ct.depth + 1, ct.path || c.name
  FROM categories c
  JOIN category_tree ct ON c.parent_id = ct.id
)
SELECT id, name, depth, array_to_string(path, ' > ') AS breadcrumb
FROM category_tree
ORDER BY path;

-- Recursive CTE with cycle detection
WITH RECURSIVE org_chart AS (
  SELECT id, name, manager_id, ARRAY[id] AS visited
  FROM employees WHERE manager_id IS NULL

  UNION ALL

  SELECT e.id, e.name, e.manager_id, oc.visited || e.id
  FROM employees e
  JOIN org_chart oc ON e.manager_id = oc.id
  WHERE NOT e.id = ANY(oc.visited)  -- prevent cycles
)
SELECT * FROM org_chart;
```

## Window Functions

```sql
-- ROW_NUMBER: unique row numbering
SELECT
  name,
  department,
  salary,
  ROW_NUMBER() OVER (PARTITION BY department ORDER BY salary DESC) AS rank
FROM employees;

-- RANK vs DENSE_RANK
SELECT
  name,
  score,
  RANK() OVER (ORDER BY score DESC) AS rank,        -- gaps after ties (1,2,2,4)
  DENSE_RANK() OVER (ORDER BY score DESC) AS dense   -- no gaps (1,2,2,3)
FROM leaderboard;

-- Running totals and moving averages
SELECT
  date,
  amount,
  SUM(amount) OVER (ORDER BY date) AS running_total,
  AVG(amount) OVER (
    ORDER BY date
    ROWS BETWEEN 6 PRECEDING AND CURRENT ROW
  ) AS moving_avg_7day
FROM daily_sales;

-- LAG / LEAD (previous/next row)
SELECT
  date,
  revenue,
  LAG(revenue, 1) OVER (ORDER BY date) AS prev_day,
  LEAD(revenue, 1) OVER (ORDER BY date) AS next_day,
  revenue - LAG(revenue, 1) OVER (ORDER BY date) AS daily_change
FROM daily_revenue;

-- FIRST_VALUE / LAST_VALUE
SELECT
  employee_id,
  department,
  salary,
  FIRST_VALUE(name) OVER (
    PARTITION BY department ORDER BY salary DESC
  ) AS highest_paid_in_dept
FROM employees;

-- NTILE (distribute rows into N buckets)
SELECT
  name,
  salary,
  NTILE(4) OVER (ORDER BY salary) AS quartile
FROM employees;

-- Percent rank and cumulative distribution
SELECT
  name,
  salary,
  ROUND(PERCENT_RANK() OVER (ORDER BY salary)::numeric, 2) AS pct_rank,
  ROUND(CUME_DIST() OVER (ORDER BY salary)::numeric, 2) AS cum_dist
FROM employees;
```

## UPSERT Patterns

```sql
-- Basic upsert (INSERT ... ON CONFLICT)
INSERT INTO users (email, name, updated_at)
VALUES ('alice@example.com', 'Alice', NOW())
ON CONFLICT (email)
DO UPDATE SET
  name = EXCLUDED.name,
  updated_at = NOW();

-- Upsert with conditional update (only update if newer)
INSERT INTO products (sku, price, updated_at)
VALUES ('ABC-123', 29.99, '2026-03-01')
ON CONFLICT (sku)
DO UPDATE SET
  price = EXCLUDED.price,
  updated_at = EXCLUDED.updated_at
WHERE products.updated_at < EXCLUDED.updated_at;

-- Upsert returning the result
INSERT INTO page_views (url, count, last_seen)
VALUES ('/home', 1, NOW())
ON CONFLICT (url)
DO UPDATE SET
  count = page_views.count + 1,
  last_seen = NOW()
RETURNING id, url, count;

-- Bulk upsert
INSERT INTO inventory (product_id, warehouse_id, quantity)
VALUES
  (1, 1, 100),
  (2, 1, 200),
  (3, 1, 50)
ON CONFLICT (product_id, warehouse_id)
DO UPDATE SET quantity = EXCLUDED.quantity;

-- DO NOTHING (skip duplicates)
INSERT INTO tags (name)
SELECT UNNEST(ARRAY['javascript', 'typescript', 'react'])
ON CONFLICT (name) DO NOTHING;
```

## JSON Operations

```sql
-- JSON column queries
SELECT
  id,
  data->>'name' AS name,               -- text extraction
  data->'address'->>'city' AS city,     -- nested extraction
  (data->>'age')::int AS age            -- cast to int
FROM profiles
WHERE data->>'status' = 'active';

-- JSONB containment (@>)
SELECT * FROM events
WHERE metadata @> '{"type": "signup"}';

-- JSONB existence (?)
SELECT * FROM events
WHERE metadata ? 'error_code';

-- Build JSON from rows
SELECT json_agg(json_build_object(
  'id', id,
  'name', name,
  'email', email
)) AS users
FROM users WHERE status = 'active';

-- Expand JSON array to rows
SELECT
  id,
  tag.value AS tag
FROM posts,
LATERAL jsonb_array_elements_text(tags) AS tag(value);

-- Update nested JSON
UPDATE profiles
SET data = jsonb_set(
  data,
  '{address,city}',
  '"New York"'::jsonb
)
WHERE id = 1;

-- Remove JSON key
UPDATE profiles
SET data = data - 'temporary_field';

-- JSON path queries (PostgreSQL 12+)
SELECT * FROM events
WHERE jsonb_path_exists(
  metadata,
  '$.items[*] ? (@.price > 100)'
);
```

## Full-Text Search

```sql
-- Setup: add tsvector column with index
ALTER TABLE articles ADD COLUMN search_vector tsvector;

UPDATE articles SET search_vector =
  setweight(to_tsvector('english', COALESCE(title, '')), 'A') ||
  setweight(to_tsvector('english', COALESCE(body, '')), 'B');

CREATE INDEX idx_articles_search ON articles USING gin(search_vector);

-- Auto-update with trigger
CREATE FUNCTION articles_search_trigger() RETURNS trigger AS $$
BEGIN
  NEW.search_vector :=
    setweight(to_tsvector('english', COALESCE(NEW.title, '')), 'A') ||
    setweight(to_tsvector('english', COALESCE(NEW.body, '')), 'B');
  RETURN NEW;
END
$$ LANGUAGE plpgsql;

CREATE TRIGGER articles_search_update
  BEFORE INSERT OR UPDATE ON articles
  FOR EACH ROW EXECUTE FUNCTION articles_search_trigger();

-- Search with ranking
SELECT
  id,
  title,
  ts_rank(search_vector, query) AS rank,
  ts_headline('english', body, query,
    'StartSel=<mark>, StopSel=</mark>, MaxFragments=3'
  ) AS snippet
FROM articles,
  to_tsquery('english', 'postgres & performance') AS query
WHERE search_vector @@ query
ORDER BY rank DESC
LIMIT 20;

-- Phrase search
SELECT * FROM articles
WHERE search_vector @@ phraseto_tsquery('english', 'database optimization');

-- Prefix search (autocomplete)
SELECT * FROM articles
WHERE search_vector @@ to_tsquery('english', 'optim:*');
```

## LATERAL Joins

```sql
-- Top N per group (get 3 most recent orders per customer)
SELECT c.id, c.name, recent_orders.*
FROM customers c
CROSS JOIN LATERAL (
  SELECT o.id AS order_id, o.total, o.created_at
  FROM orders o
  WHERE o.customer_id = c.id
  ORDER BY o.created_at DESC
  LIMIT 3
) AS recent_orders;

-- Correlated subquery as a join
SELECT
  d.name AS department,
  stats.avg_salary,
  stats.employee_count
FROM departments d
CROSS JOIN LATERAL (
  SELECT
    AVG(salary) AS avg_salary,
    COUNT(*) AS employee_count
  FROM employees
  WHERE department_id = d.id
) AS stats
WHERE stats.employee_count > 0;
```

## Array Operations

```sql
-- Array contains
SELECT * FROM posts WHERE 'javascript' = ANY(tags);

-- Array overlap (any common element)
SELECT * FROM posts WHERE tags && ARRAY['react', 'vue', 'angular'];

-- Array contains all
SELECT * FROM posts WHERE tags @> ARRAY['javascript', 'typescript'];

-- Unnest array to rows
SELECT id, UNNEST(tags) AS tag FROM posts;

-- Aggregate into array
SELECT
  department,
  ARRAY_AGG(name ORDER BY name) AS members
FROM employees
GROUP BY department;

-- Array append / remove
UPDATE posts SET tags = array_append(tags, 'featured') WHERE id = 1;
UPDATE posts SET tags = array_remove(tags, 'draft') WHERE id = 1;
```

## Bulk Operations

```sql
-- Bulk insert from VALUES
INSERT INTO users (email, name, role)
VALUES
  ('a@example.com', 'Alice', 'user'),
  ('b@example.com', 'Bob', 'admin'),
  ('c@example.com', 'Charlie', 'user');

-- Bulk update with FROM
UPDATE products p
SET price = np.price, updated_at = NOW()
FROM (VALUES
  ('SKU-001', 29.99),
  ('SKU-002', 49.99),
  ('SKU-003', 99.99)
) AS np(sku, price)
WHERE p.sku = np.sku;

-- Bulk delete with subquery
DELETE FROM sessions
WHERE user_id IN (
  SELECT id FROM users WHERE status = 'deleted'
);

-- COPY for fastest bulk loading
COPY users (email, name, created_at)
FROM '/tmp/users.csv'
WITH (FORMAT csv, HEADER true);
```

## Date/Time Patterns

```sql
-- Date truncation for grouping
SELECT
  DATE_TRUNC('month', created_at) AS month,
  COUNT(*) AS signups
FROM users
GROUP BY 1
ORDER BY 1;

-- Generate date series (fill gaps)
SELECT
  d.date,
  COALESCE(s.count, 0) AS signups
FROM generate_series(
  '2026-01-01'::date,
  '2026-12-31'::date,
  '1 day'::interval
) AS d(date)
LEFT JOIN (
  SELECT DATE(created_at) AS date, COUNT(*) AS count
  FROM users GROUP BY 1
) s ON s.date = d.date;

-- Date ranges with OVERLAPS
SELECT * FROM bookings
WHERE (check_in, check_out) OVERLAPS ('2026-06-01', '2026-06-15');

-- Time zone conversions
SELECT
  created_at AT TIME ZONE 'UTC' AT TIME ZONE 'America/New_York' AS eastern_time
FROM events;
```

## Gotchas

1. **CTEs are optimization fences in older PostgreSQL.** Before PostgreSQL 12, CTEs were always materialized (results stored in temp table). In 12+, the planner can inline them. If you're on an older version and a CTE-based query is slow, try rewriting as a subquery.

2. **`DISTINCT ON` is PostgreSQL-specific but incredibly useful.** `SELECT DISTINCT ON (user_id) * FROM events ORDER BY user_id, created_at DESC` gives you the latest event per user. Not standard SQL, but far cleaner than window function alternatives.

3. **`NULLS FIRST` / `NULLS LAST` matters for indexes.** If your query uses `ORDER BY col DESC NULLS LAST`, your index must match: `CREATE INDEX ON t(col DESC NULLS LAST)`. Mismatched null ordering = no index scan.

4. **`IN` vs `= ANY()` for parameterized queries.** `WHERE id IN ($1, $2, $3)` requires a fixed parameter count. `WHERE id = ANY($1::int[])` accepts an array of any length. Use `= ANY()` with ORMs and prepared statements.

5. **`UPDATE ... FROM` is PostgreSQL-specific.** Standard SQL uses `UPDATE ... SET ... WHERE EXISTS (SELECT ...)`. The FROM syntax is cleaner but not portable. If you need cross-database compatibility, use the correlated subquery form.

6. **JSON operators: `->` returns JSON, `->>` returns text.** Forgetting this causes type comparison failures. `data->>'count' > '5'` compares text (lexicographic). Cast: `(data->>'count')::int > 5`.
