---
name: postgres-security
description: >
  PostgreSQL security — roles and permissions, row-level security (RLS),
  SSL/TLS configuration, audit logging, and hardening best practices.
  Triggers: "postgres security", "database permissions", "row level security",
  "RLS", "postgres roles", "database audit", "pg_hba.conf", "SSL postgres".
  NOT for: query performance (use postgres-performance), migrations (use postgres-migrations).
version: 1.0.0
allowed-tools: Read, Grep, Glob, Bash, Edit, Write
---

# PostgreSQL Security

## Roles and Permissions

### Role Hierarchy

```
superuser (postgres)
  └── app_admin         (DDL, manages schema)
       ├── app_writer   (INSERT, UPDATE, DELETE)
       └── app_reader   (SELECT only)
            └── app_api  (inherits reader + specific writes)
```

### Create Roles

```sql
-- Application roles (no login — used for permission grouping)
CREATE ROLE app_admin NOLOGIN;
CREATE ROLE app_writer NOLOGIN;
CREATE ROLE app_reader NOLOGIN;

-- Login users that inherit from roles
CREATE USER app_service WITH PASSWORD 'strong-password-here' LOGIN;
GRANT app_writer TO app_service;

CREATE USER app_readonly WITH PASSWORD 'another-password' LOGIN;
GRANT app_reader TO app_readonly;

CREATE USER app_migrate WITH PASSWORD 'migration-password' LOGIN;
GRANT app_admin TO app_migrate;
```

### Grant Permissions

```sql
-- Reader: SELECT on all tables
GRANT USAGE ON SCHEMA public TO app_reader;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO app_reader;
ALTER DEFAULT PRIVILEGES IN SCHEMA public
  GRANT SELECT ON TABLES TO app_reader;

-- Writer: INSERT, UPDATE, DELETE (inherits reader via GRANT)
GRANT app_reader TO app_writer;
GRANT INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO app_writer;
ALTER DEFAULT PRIVILEGES IN SCHEMA public
  GRANT INSERT, UPDATE, DELETE ON TABLES TO app_writer;

-- Writer also needs sequence usage for INSERTs with SERIAL/BIGSERIAL
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO app_writer;
ALTER DEFAULT PRIVILEGES IN SCHEMA public
  GRANT USAGE, SELECT ON SEQUENCES TO app_writer;

-- Admin: DDL operations
GRANT app_writer TO app_admin;
GRANT CREATE ON SCHEMA public TO app_admin;
ALTER DEFAULT PRIVILEGES IN SCHEMA public
  GRANT ALL ON TABLES TO app_admin;
```

### Revoke Dangerous Defaults

```sql
-- Revoke public's default ability to connect and create objects
REVOKE ALL ON DATABASE mydb FROM PUBLIC;
REVOKE ALL ON SCHEMA public FROM PUBLIC;

-- Grant back connect to specific roles
GRANT CONNECT ON DATABASE mydb TO app_reader;
GRANT CONNECT ON DATABASE mydb TO app_writer;
GRANT CONNECT ON DATABASE mydb TO app_admin;
```

### Check Current Permissions

```sql
-- Table permissions
SELECT grantee, table_name, privilege_type
FROM information_schema.role_table_grants
WHERE table_schema = 'public'
ORDER BY table_name, grantee;

-- Role memberships
SELECT r.rolname AS role,
       m.rolname AS member
FROM pg_auth_members am
JOIN pg_roles r ON r.oid = am.roleid
JOIN pg_roles m ON m.oid = am.member
ORDER BY r.rolname;
```

## Row-Level Security (RLS)

RLS lets you control which rows a user can see or modify, enforced at the database level.

### Enable RLS

```sql
-- Enable RLS on the table
ALTER TABLE documents ENABLE ROW LEVEL SECURITY;

-- Force RLS for table owner too (otherwise owner bypasses RLS)
ALTER TABLE documents FORCE ROW LEVEL SECURITY;
```

### Basic Policies

```sql
-- Users can only see their own documents
CREATE POLICY documents_select ON documents
  FOR SELECT
  USING (user_id = current_setting('app.current_user_id')::uuid);

-- Users can only insert documents they own
CREATE POLICY documents_insert ON documents
  FOR INSERT
  WITH CHECK (user_id = current_setting('app.current_user_id')::uuid);

-- Users can only update their own documents
CREATE POLICY documents_update ON documents
  FOR UPDATE
  USING (user_id = current_setting('app.current_user_id')::uuid)
  WITH CHECK (user_id = current_setting('app.current_user_id')::uuid);

-- Users can only delete their own documents
CREATE POLICY documents_delete ON documents
  FOR DELETE
  USING (user_id = current_setting('app.current_user_id')::uuid);
```

### Multi-Tenant RLS

```sql
-- Set tenant context per request
SET app.current_tenant_id = 'tenant-123';

-- All queries automatically filtered to current tenant
CREATE POLICY tenant_isolation ON orders
  FOR ALL
  USING (tenant_id = current_setting('app.current_tenant_id')::uuid);
```

### Setting Context in Application Code

```typescript
// Middleware: set RLS context per request
app.use(async (req, res, next) => {
  const userId = req.auth?.userId;
  const tenantId = req.auth?.tenantId;

  if (userId) {
    await db.query("SELECT set_config('app.current_user_id', $1, true)", [userId]);
  }
  if (tenantId) {
    await db.query("SELECT set_config('app.current_tenant_id', $1, true)", [tenantId]);
  }

  next();
});
```

### Role-Based Policies

```sql
-- Admins can see all rows
CREATE POLICY admin_all ON documents
  FOR ALL
  TO app_admin
  USING (true)
  WITH CHECK (true);

-- Regular users see only their own
CREATE POLICY user_own ON documents
  FOR ALL
  TO app_writer
  USING (user_id = current_setting('app.current_user_id')::uuid)
  WITH CHECK (user_id = current_setting('app.current_user_id')::uuid);

-- Service accounts bypass RLS (use with caution)
CREATE POLICY service_bypass ON documents
  FOR ALL
  TO app_service_account
  USING (true);
```

### Shared Resources with RLS

```sql
-- Documents shared via a sharing table
CREATE POLICY documents_shared ON documents
  FOR SELECT
  USING (
    user_id = current_setting('app.current_user_id')::uuid
    OR id IN (
      SELECT document_id FROM document_shares
      WHERE shared_with = current_setting('app.current_user_id')::uuid
    )
  );
```

## SSL/TLS Configuration

### Server-Side SSL

```ini
# postgresql.conf
ssl = on
ssl_cert_file = '/etc/ssl/certs/server.crt'
ssl_key_file = '/etc/ssl/private/server.key'
ssl_ca_file = '/etc/ssl/certs/ca.crt'        # For client cert verification
ssl_min_protocol_version = 'TLSv1.2'         # Disable TLS 1.0/1.1
```

### Require SSL in pg_hba.conf

```
# pg_hba.conf — require SSL for remote connections
# TYPE   DATABASE   USER         ADDRESS         METHOD
hostssl  all        all          0.0.0.0/0       scram-sha-256
hostssl  all        all          ::/0            scram-sha-256

# Reject non-SSL remote connections
hostnossl all       all          0.0.0.0/0       reject
```

### Connection String with SSL

```
# Require SSL
postgresql://user:pass@host:5432/db?sslmode=require

# Verify server certificate
postgresql://user:pass@host:5432/db?sslmode=verify-full&sslrootcert=/path/to/ca.crt
```

### SSL Modes

| Mode | Encryption | Server Cert Check | MITM Protection |
|------|-----------|-------------------|-----------------|
| `disable` | No | No | No |
| `allow` | Maybe | No | No |
| `prefer` | Yes (if available) | No | No |
| `require` | Yes | No | No |
| `verify-ca` | Yes | Yes (CA match) | Partial |
| `verify-full` | Yes | Yes (CA + hostname) | Yes |

**Use `verify-full` in production.** `require` encrypts but doesn't prevent MITM attacks.

## pg_hba.conf (Client Authentication)

```
# TYPE   DATABASE   USER         ADDRESS           METHOD

# Local connections (Unix socket)
local    all        postgres                        peer
local    all        all                             scram-sha-256

# Localhost
host     all        all          127.0.0.1/32      scram-sha-256
host     all        all          ::1/128            scram-sha-256

# Application server subnet
hostssl  myapp      app_service  10.0.1.0/24       scram-sha-256

# Read replica connections
hostssl  myapp      replicator   10.0.2.0/24       scram-sha-256

# Block everything else
host     all        all          0.0.0.0/0         reject
```

### Authentication Methods (Ranked by Security)

| Method | Security | Use Case |
|--------|----------|----------|
| `scram-sha-256` | Best | Default for password auth (PG 10+) |
| `cert` | Best | Client certificate auth |
| `gss` / `sspi` | High | Kerberos / Active Directory |
| `peer` | High | Local Unix socket (OS user = PG user) |
| `md5` | Medium | Legacy password auth (upgrade to scram) |
| `password` | Low | Plain text (never use over network) |
| `trust` | None | No auth (development only!) |

## Audit Logging

### Built-in Logging

```ini
# postgresql.conf
log_statement = 'mod'            # Log INSERT, UPDATE, DELETE, DDL
log_connections = on              # Log successful connections
log_disconnections = on           # Log disconnections
log_duration = on                 # Log query durations
log_min_duration_statement = 100  # Log queries > 100ms

# Detailed query logging
log_line_prefix = '%t [%p] %u@%d '  # timestamp, pid, user, database
```

### pgAudit Extension

```sql
-- Install
CREATE EXTENSION pgaudit;

-- Configure in postgresql.conf
-- pgaudit.log = 'write, ddl'
-- pgaudit.log_catalog = off
-- pgaudit.log_level = 'log'
-- pgaudit.log_statement_once = on
```

### Application-Level Audit Table

```sql
CREATE TABLE audit_log (
  id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  table_name TEXT NOT NULL,
  record_id  TEXT NOT NULL,
  action     TEXT NOT NULL CHECK (action IN ('INSERT', 'UPDATE', 'DELETE')),
  old_data   JSONB,
  new_data   JSONB,
  changed_by TEXT NOT NULL,
  changed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  ip_address INET
);

CREATE INDEX idx_audit_table_record ON audit_log (table_name, record_id);
CREATE INDEX idx_audit_changed_at ON audit_log (changed_at DESC);

-- Trigger function
CREATE OR REPLACE FUNCTION audit_trigger() RETURNS TRIGGER AS $$
BEGIN
  IF TG_OP = 'INSERT' THEN
    INSERT INTO audit_log (table_name, record_id, action, new_data, changed_by)
    VALUES (TG_TABLE_NAME, NEW.id::text, 'INSERT', to_jsonb(NEW),
            current_setting('app.current_user_id', true));
    RETURN NEW;
  ELSIF TG_OP = 'UPDATE' THEN
    INSERT INTO audit_log (table_name, record_id, action, old_data, new_data, changed_by)
    VALUES (TG_TABLE_NAME, NEW.id::text, 'UPDATE', to_jsonb(OLD), to_jsonb(NEW),
            current_setting('app.current_user_id', true));
    RETURN NEW;
  ELSIF TG_OP = 'DELETE' THEN
    INSERT INTO audit_log (table_name, record_id, action, old_data, changed_by)
    VALUES (TG_TABLE_NAME, OLD.id::text, 'DELETE', to_jsonb(OLD),
            current_setting('app.current_user_id', true));
    RETURN OLD;
  END IF;
END;
$$ LANGUAGE plpgsql;

-- Attach to tables
CREATE TRIGGER users_audit
  AFTER INSERT OR UPDATE OR DELETE ON users
  FOR EACH ROW EXECUTE FUNCTION audit_trigger();

CREATE TRIGGER orders_audit
  AFTER INSERT OR UPDATE OR DELETE ON orders
  FOR EACH ROW EXECUTE FUNCTION audit_trigger();
```

## SQL Injection Prevention

### Parameterized Queries (Always Use These)

```typescript
// SAFE — parameterized query
const user = await db.query(
  'SELECT * FROM users WHERE email = $1',
  [email]
);

// SAFE — Prisma (auto-parameterized)
const user = await prisma.user.findUnique({
  where: { email },
});

// DANGEROUS — string concatenation
const user = await db.query(
  `SELECT * FROM users WHERE email = '${email}'`  // SQL INJECTION!
);
```

### Dynamic Table/Column Names

```typescript
// When you MUST use dynamic identifiers (table names, column names):
import { escapeIdentifier } from 'pg';  // pg library utility

// Or validate against allowlist
const ALLOWED_SORT_COLUMNS = ['name', 'created_at', 'email'];

function buildQuery(sortColumn: string, sortDir: string) {
  if (!ALLOWED_SORT_COLUMNS.includes(sortColumn)) {
    throw new Error(`Invalid sort column: ${sortColumn}`);
  }
  if (!['ASC', 'DESC'].includes(sortDir.toUpperCase())) {
    throw new Error(`Invalid sort direction: ${sortDir}`);
  }

  return `SELECT * FROM users ORDER BY ${sortColumn} ${sortDir}`;
}
```

## Hardening Checklist

```
Authentication:
[ ] scram-sha-256 (not md5 or trust) for all remote connections
[ ] Strong passwords (min 16 chars, random) for all database users
[ ] Separate roles for app, migration, read-only, and admin
[ ] No shared passwords between environments

Network:
[ ] SSL required for all remote connections (hostssl + reject hostnossl)
[ ] verify-full SSL mode in application connection strings
[ ] pg_hba.conf restricts to known IP ranges
[ ] Database port not exposed to public internet
[ ] Firewall rules limit access to app server IPs only

Permissions:
[ ] PUBLIC schema privileges revoked
[ ] Each application uses minimum necessary permissions
[ ] No application connects as superuser
[ ] ALTER DEFAULT PRIVILEGES set for future tables
[ ] RLS enabled for multi-tenant data

Monitoring:
[ ] log_connections and log_disconnections enabled
[ ] log_statement = 'mod' or pgAudit installed
[ ] Failed login attempts monitored and alerted
[ ] pg_stat_activity reviewed for unexpected connections

Data:
[ ] Sensitive columns encrypted at application level
[ ] Backups encrypted at rest
[ ] Point-in-time recovery (PITR) configured and tested
[ ] Backup restoration tested monthly
```

## Gotchas

1. **RLS is bypassed by table owners** — use `FORCE ROW LEVEL SECURITY` if the connecting role owns the table. Or better: app roles should never own tables.

2. **`current_setting()` returns empty string by default** — use `current_setting('app.user_id', true)` (the `true` makes it return NULL instead of erroring when not set).

3. **Default privileges only affect future objects** — `ALTER DEFAULT PRIVILEGES` won't retroactively grant permissions on existing tables. Run `GRANT` separately for current tables.

4. **`pg_hba.conf` is order-dependent** — first matching rule wins. Put restrictive rules before permissive ones.

5. **SCRAM requires password re-set** — switching from `md5` to `scram-sha-256` in pg_hba.conf requires users to change their passwords (the stored hash format is different).

6. **RLS policies are OR'd per command** — if a user matches multiple SELECT policies, they see the union of all matching rows. This is usually what you want but can surprise you.

7. **Superuser bypasses everything** — RLS, permissions, connection limits. Never give application accounts superuser. Use `CREATEROLE` for admin tasks instead.

8. **SSL `require` doesn't verify the server** — it encrypts the connection but accepts any certificate. A MITM attacker could present their own cert. Use `verify-full` for real security.
