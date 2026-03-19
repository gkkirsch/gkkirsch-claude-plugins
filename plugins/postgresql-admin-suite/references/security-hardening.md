# PostgreSQL Security Hardening Reference

Comprehensive reference for securing PostgreSQL deployments covering authentication,
authorization, encryption, auditing, and network security.

---

## 1. Security Checklist

1. Use SCRAM-SHA-256 authentication (not MD5 or trust).
2. Enforce SSL/TLS for all remote connections.
3. Configure pg_hba.conf with least-privilege access rules.
4. Use role-based access control with no direct superuser application access.
5. Apply principle of least privilege — GRANT only what is needed.
6. Enable Row-Level Security (RLS) for multi-tenant applications.
7. Set password expiration and connection limits for all login roles.
8. Install and configure pgAudit for audit logging.
9. Restrict network access — bind to specific IPs, use firewalls.
10. Use certificate-based authentication for service accounts.
11. Encrypt data at rest (TDE or filesystem encryption).
12. Encrypt sensitive columns with pgcrypto.
13. Disable the `trust` authentication method in production.
14. Revoke `CREATE` privilege on the `public` schema from `PUBLIC`.
15. Set `log_connections = on` and `log_disconnections = on`.
16. Regularly rotate credentials and SSL certificates.
17. Keep PostgreSQL updated with the latest security patches.
18. Use `pg_hba.conf` to restrict replication connections.
19. Disable unnecessary extensions and contrib modules.
20. Monitor `pg_stat_activity` for suspicious connections.

---

## 2. Role Management

### Role Hierarchy and Design

PostgreSQL uses a unified role system — there is no distinction between "users" and
"groups" at the engine level. A role with the `LOGIN` attribute is effectively a user.
A role without `LOGIN` is effectively a group.

```sql
-- Create group roles (no login)
CREATE ROLE app_readonly NOLOGIN;
CREATE ROLE app_readwrite NOLOGIN;
CREATE ROLE app_admin NOLOGIN;

-- Create login roles (users)
CREATE ROLE web_app LOGIN PASSWORD 'strong_password_here';
CREATE ROLE analytics_user LOGIN PASSWORD 'another_strong_password';
CREATE ROLE dba_user LOGIN PASSWORD 'dba_password' SUPERUSER;
```

### Role Attributes

```sql
-- Full list of role attributes
CREATE ROLE example_role
  LOGIN                    -- can connect (omit for group roles)
  PASSWORD 'password'      -- set password
  VALID UNTIL '2026-01-01' -- password expiration
  CONNECTION LIMIT 10      -- max concurrent connections
  NOSUPERUSER              -- cannot bypass all access checks
  NOCREATEDB               -- cannot create databases
  NOCREATEROLE             -- cannot create other roles
  NOINHERIT                -- does NOT inherit granted role privileges automatically
  NOREPLICATION            -- cannot initiate streaming replication
  NOBYPASSRLS;             -- cannot bypass Row-Level Security
```

### GRANT / REVOKE Patterns

```sql
-- Grant privileges on tables
GRANT SELECT ON ALL TABLES IN SCHEMA public TO app_readonly;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO app_readwrite;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO app_admin;

-- Grant privileges on sequences
GRANT USAGE ON ALL SEQUENCES IN SCHEMA public TO app_readwrite;

-- Grant schema usage
GRANT USAGE ON SCHEMA public TO app_readonly;
GRANT USAGE, CREATE ON SCHEMA public TO app_readwrite;

-- Set default privileges for future tables
ALTER DEFAULT PRIVILEGES IN SCHEMA public
  GRANT SELECT ON TABLES TO app_readonly;
ALTER DEFAULT PRIVILEGES IN SCHEMA public
  GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO app_readwrite;
ALTER DEFAULT PRIVILEGES IN SCHEMA public
  GRANT USAGE ON SEQUENCES TO app_readwrite;

-- Grant role membership (role inheritance)
GRANT app_readonly TO analytics_user;
GRANT app_readwrite TO web_app;
GRANT app_admin TO dba_user;

-- Revoke dangerous defaults
REVOKE CREATE ON SCHEMA public FROM PUBLIC;
REVOKE ALL ON DATABASE mydb FROM PUBLIC;
```

### Role Inheritance

```sql
-- With INHERIT (default): member automatically has all privileges of the granted role
CREATE ROLE web_app LOGIN INHERIT PASSWORD 'password';
GRANT app_readwrite TO web_app;
-- web_app immediately has all app_readwrite privileges

-- With NOINHERIT: must explicitly SET ROLE to activate privileges
CREATE ROLE admin_user LOGIN NOINHERIT PASSWORD 'password';
GRANT app_admin TO admin_user;
-- admin_user must run: SET ROLE app_admin; to use those privileges
-- This provides an extra layer of protection — privilege escalation is intentional
```

### Least-Privilege Patterns

```sql
-- Application-specific role with minimal permissions
CREATE ROLE orders_service LOGIN PASSWORD 'service_password';
GRANT USAGE ON SCHEMA orders TO orders_service;
GRANT SELECT, INSERT, UPDATE ON orders.orders TO orders_service;
GRANT SELECT ON orders.products TO orders_service;
GRANT USAGE ON SEQUENCE orders.orders_id_seq TO orders_service;
-- No access to other schemas or tables

-- Read-only replica user
CREATE ROLE readonly_replica LOGIN PASSWORD 'replica_password';
GRANT CONNECT ON DATABASE mydb TO readonly_replica;
GRANT USAGE ON SCHEMA public TO readonly_replica;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO readonly_replica;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON TABLES TO readonly_replica;
```

---

## 3. Row-Level Security (RLS)

### Enabling RLS

```sql
-- Enable RLS on a table
ALTER TABLE orders ENABLE ROW LEVEL SECURITY;

-- Force RLS even for table owners (important for testing)
ALTER TABLE orders FORCE ROW LEVEL SECURITY;
```

When RLS is enabled with no policies, all access is denied by default (except for
table owners and superusers unless FORCE ROW LEVEL SECURITY is set).

### Creating Policies

```sql
-- Basic syntax
CREATE POLICY policy_name ON table_name
  [AS PERMISSIVE | RESTRICTIVE]
  [FOR {ALL | SELECT | INSERT | UPDATE | DELETE}]
  [TO {role_name | PUBLIC}]
  [USING (condition)]           -- filters existing rows (SELECT, UPDATE, DELETE)
  [WITH CHECK (condition)];     -- validates new/modified rows (INSERT, UPDATE)
```

### USING vs WITH CHECK

- **USING**: Applied to existing rows. Determines which rows are visible to SELECT,
  and which rows can be affected by UPDATE/DELETE.
- **WITH CHECK**: Applied to new or modified rows. Determines which rows can be
  INSERTed or what values an UPDATE can set.

```sql
-- Users can see their own orders, and can only create orders for themselves
CREATE POLICY user_orders ON orders
  FOR ALL
  TO app_readwrite
  USING (user_id = current_setting('app.current_user_id')::int)
  WITH CHECK (user_id = current_setting('app.current_user_id')::int);
```

### Permissive vs Restrictive Policies

```sql
-- PERMISSIVE (default): multiple permissive policies are OR'd together
-- Any matching permissive policy grants access.
CREATE POLICY allow_own_orders ON orders AS PERMISSIVE
  FOR SELECT TO app_user
  USING (user_id = current_setting('app.current_user_id')::int);

CREATE POLICY allow_public_orders ON orders AS PERMISSIVE
  FOR SELECT TO app_user
  USING (is_public = true);
-- User can see their own orders OR public orders

-- RESTRICTIVE: combined with AND against permissive policies
-- All restrictive policies must pass IN ADDITION to at least one permissive policy.
CREATE POLICY require_active_tenant ON orders AS RESTRICTIVE
  FOR ALL TO app_user
  USING (tenant_active = true);
-- Even if a permissive policy allows access, this must also be satisfied
```

### Multi-Tenant RLS Pattern

```sql
-- Tenant isolation using session variable
CREATE TABLE tenant_data (
  id serial PRIMARY KEY,
  tenant_id int NOT NULL,
  data jsonb NOT NULL
);

ALTER TABLE tenant_data ENABLE ROW LEVEL SECURITY;
ALTER TABLE tenant_data FORCE ROW LEVEL SECURITY;

-- Tenant isolation policy
CREATE POLICY tenant_isolation ON tenant_data
  FOR ALL
  TO app_readwrite
  USING (tenant_id = current_setting('app.tenant_id')::int)
  WITH CHECK (tenant_id = current_setting('app.tenant_id')::int);

-- Application sets tenant context per connection/transaction
SET app.tenant_id = '42';
-- All subsequent queries on tenant_data are automatically filtered
SELECT * FROM tenant_data;  -- only sees tenant 42's data
INSERT INTO tenant_data (tenant_id, data) VALUES (99, '{}');
-- ERROR: new row violates row-level security policy (tenant_id != 42)
```

### Owner-Based RLS Pattern

```sql
CREATE POLICY owner_access ON documents
  FOR ALL TO app_user
  USING (owner_id = current_setting('app.user_id')::int)
  WITH CHECK (owner_id = current_setting('app.user_id')::int);

-- Admin can see everything
CREATE POLICY admin_access ON documents
  FOR ALL TO app_admin
  USING (true)
  WITH CHECK (true);
```

### Role-Based RLS Pattern

```sql
-- Different access levels based on role
CREATE POLICY manager_access ON employees
  FOR SELECT TO manager_role
  USING (department = current_setting('app.department'));

CREATE POLICY hr_full_access ON employees
  FOR ALL TO hr_role
  USING (true);

CREATE POLICY self_access ON employees
  FOR SELECT TO employee_role
  USING (employee_id = current_setting('app.employee_id')::int);
```

### RLS Bypass

```sql
-- Superusers bypass RLS by default
-- Table owners bypass RLS unless FORCE ROW LEVEL SECURITY is set
-- Roles with BYPASSRLS attribute bypass all RLS policies

CREATE ROLE migration_runner LOGIN PASSWORD 'password' BYPASSRLS;
-- Use sparingly — only for migrations and data maintenance tasks
```

### Performance Considerations

```sql
-- Index the columns used in RLS policies
CREATE INDEX idx_tenant_data_tenant_id ON tenant_data (tenant_id);

-- Avoid expensive function calls in USING/WITH CHECK clauses
-- current_setting() is fast; complex subqueries are not

-- Test RLS performance impact
SET app.tenant_id = '42';
EXPLAIN (ANALYZE, BUFFERS) SELECT * FROM tenant_data WHERE id = 100;
-- Look for Filter nodes added by RLS policies
```

---

## 4. pg_hba.conf Patterns

### Authentication Methods

| Method           | Description                                    | Security Level |
|------------------|------------------------------------------------|----------------|
| `trust`          | Accept all connections without password         | NONE (danger)  |
| `reject`         | Reject all connections                          | N/A            |
| `scram-sha-256`  | Challenge-response with salted hash (PG 10+)   | HIGH           |
| `md5`            | MD5-based challenge-response                   | MEDIUM         |
| `password`       | Cleartext password (never use without SSL)      | LOW            |
| `cert`           | SSL client certificate                          | VERY HIGH      |
| `ldap`           | LDAP directory authentication                  | HIGH           |
| `radius`         | RADIUS server authentication                   | HIGH           |
| `gss`            | GSSAPI/Kerberos authentication                 | HIGH           |
| `peer`           | OS user matches database user (local only)     | HIGH           |
| `ident`          | Remote ident server (rarely used)              | LOW            |

### Common pg_hba.conf Configurations

```ini
# TYPE  DATABASE        USER            ADDRESS                 METHOD

# --- Local connections ---
# Allow postgres superuser via OS peer authentication (Unix socket)
local   all             postgres                                peer

# Allow local application users with SCRAM
local   all             all                                     scram-sha-256

# --- Host connections (TCP/IP) ---
# Reject all non-SSL connections from remote hosts
hostnossl all           all             0.0.0.0/0               reject
hostnossl all           all             ::/0                    reject

# Application server subnet (IPv4) — require SSL + SCRAM
hostssl   mydb          web_app         10.0.1.0/24             scram-sha-256

# Analytics server — read-only user, specific IP
hostssl   mydb          analytics_user  10.0.2.50/32            scram-sha-256

# Admin access — certificate authentication only
hostssl   all           dba_user        10.0.0.0/16             cert

# --- Replication connections ---
# Streaming replication from specific standby servers
hostssl   replication   replicator      10.0.3.10/32            scram-sha-256
hostssl   replication   replicator      10.0.3.11/32            scram-sha-256

# --- Deny everything else ---
host      all           all             0.0.0.0/0               reject
host      all           all             ::/0                    reject
```

### Key Rules

1. **Lines are processed top-to-bottom; first match wins.**
2. Put more specific rules before general rules.
3. Always end with a `reject` catch-all.
4. Use `hostssl` instead of `host` to require SSL.
5. Use `hostnossl ... reject` to explicitly block non-SSL TCP connections.
6. Never use `trust` in production.

### Reloading pg_hba.conf

```sql
-- After editing pg_hba.conf, reload without restart
SELECT pg_reload_conf();
-- Or from the command line:
-- pg_ctl reload -D /var/lib/postgresql/data
```

---

## 5. SSL/TLS Configuration

### Generating Self-Signed Certificates (Development)

```bash
# Generate CA key and certificate
openssl req -new -x509 -days 3650 -nodes \
  -out ca.crt -keyout ca.key \
  -subj "/CN=PostgreSQL CA"

# Generate server key and CSR
openssl req -new -nodes \
  -out server.csr -keyout server.key \
  -subj "/CN=db.example.com"

# Sign server certificate with CA
openssl x509 -req -in server.csr \
  -CA ca.crt -CAkey ca.key -CAcreateserial \
  -out server.crt -days 365

# Set permissions
chmod 600 server.key
chown postgres:postgres server.key server.crt ca.crt
```

### postgresql.conf SSL Settings

```ini
# Enable SSL
ssl = on

# Certificate files (relative to data directory)
ssl_cert_file = 'server.crt'
ssl_key_file = 'server.key'
ssl_ca_file = 'ca.crt'            # for client certificate verification

# Minimum TLS version (1.2 recommended, 1.3 for modern clients)
ssl_min_protocol_version = 'TLSv1.2'

# Cipher suites (strong ciphers only)
ssl_ciphers = 'HIGH:MEDIUM:+3DES:!aNULL'

# Prefer server cipher order
ssl_prefer_server_ciphers = on

# ECDH curve
ssl_ecdh_curve = 'prime256v1'

# Certificate Revocation List (optional)
ssl_crl_file = ''
ssl_crl_dir = ''
```

### Enforcing SSL for All Remote Connections

```ini
# In pg_hba.conf — reject all non-SSL TCP connections
hostnossl  all  all  0.0.0.0/0  reject
hostnossl  all  all  ::/0       reject

# Only allow SSL connections
hostssl    all  all  0.0.0.0/0  scram-sha-256
hostssl    all  all  ::/0       scram-sha-256
```

### Client Certificate Authentication

```bash
# Generate client key and CSR
openssl req -new -nodes \
  -out client.csr -keyout client.key \
  -subj "/CN=web_app"

# Sign with the same CA
openssl x509 -req -in client.csr \
  -CA ca.crt -CAkey ca.key -CAcreateserial \
  -out client.crt -days 365
```

```ini
# pg_hba.conf — require client certificate
hostssl  mydb  web_app  10.0.1.0/24  cert  clientcert=verify-full
```

```bash
# Client connection with certificate
psql "host=db.example.com dbname=mydb user=web_app \
  sslmode=verify-full \
  sslcert=client.crt \
  sslkey=client.key \
  sslrootcert=ca.crt"
```

### Verifying SSL Status

```sql
-- Check if current connection uses SSL
SELECT ssl, version, cipher, bits
FROM pg_stat_ssl
WHERE pid = pg_backend_pid();

-- View all SSL connections
SELECT s.pid, s.usename, s.client_addr,
       ssl.ssl, ssl.version, ssl.cipher, ssl.bits
FROM pg_stat_activity s
JOIN pg_stat_ssl ssl ON s.pid = ssl.pid
WHERE s.state = 'active';
```

---

## 6. Password Policies

### SCRAM-SHA-256 vs MD5

```sql
-- Set the default password encryption (in postgresql.conf)
-- password_encryption = 'scram-sha-256'  -- recommended (PG 10+)
-- password_encryption = 'md5'            -- legacy, weaker

-- Check current setting
SHOW password_encryption;

-- When creating a role, the password is hashed with the current method
CREATE ROLE web_app LOGIN PASSWORD 'strong_password';
-- Verify storage method
SELECT rolname, rolpassword FROM pg_authid WHERE rolname = 'web_app';
-- SCRAM: starts with 'SCRAM-SHA-256$...'
-- MD5:   starts with 'md5...'
```

**Why SCRAM is better**: MD5 hashing uses `md5(password + username)`, which is weak
against rainbow tables and doesn't use salting properly. SCRAM-SHA-256 uses a random
salt, iteration count, and channel binding for mutual authentication.

### Password Expiration

```sql
-- Set password expiration
ALTER ROLE web_app VALID UNTIL '2026-06-01';

-- Check expiration dates
SELECT rolname, rolvaliduntil
FROM pg_roles
WHERE rolvaliduntil IS NOT NULL
ORDER BY rolvaliduntil;

-- Remove expiration
ALTER ROLE web_app VALID UNTIL 'infinity';
```

### Connection Limits

```sql
-- Limit concurrent connections per role
ALTER ROLE web_app CONNECTION LIMIT 20;
ALTER ROLE analytics_user CONNECTION LIMIT 5;

-- Check current connection counts vs limits
SELECT r.rolname, r.rolconnlimit,
       count(s.pid) AS current_connections
FROM pg_roles r
LEFT JOIN pg_stat_activity s ON r.rolname = s.usename
WHERE r.rolcanlogin
GROUP BY r.rolname, r.rolconnlimit
ORDER BY r.rolname;
```

### passwordcheck Extension

```sql
-- The passwordcheck module enforces password strength rules
-- Must be loaded via shared_preload_libraries in postgresql.conf:
-- shared_preload_libraries = 'passwordcheck'

-- After loading, it rejects weak passwords on CREATE ROLE and ALTER ROLE:
-- - Minimum 8 characters
-- - Must contain letters and non-letters
-- - Cannot be the username

-- For more advanced policies, use the credcheck extension (third-party)
```

---

## 7. Audit Logging

### pgAudit Extension

pgAudit provides detailed session and object audit logging beyond what PostgreSQL's
built-in `log_statement` offers.

```sql
-- Install pgAudit
-- In postgresql.conf:
-- shared_preload_libraries = 'pgaudit'

CREATE EXTENSION pgaudit;

-- Configure what to audit (session-level)
-- In postgresql.conf:
-- pgaudit.log = 'ddl, write'
-- Options: read, write, function, role, ddl, misc, misc_set, all, none

-- Per-role audit configuration
ALTER ROLE web_app SET pgaudit.log = 'write, ddl';
ALTER ROLE analytics_user SET pgaudit.log = 'read';
```

### Log Levels and Categories

| Category   | What It Logs                                          |
|------------|-------------------------------------------------------|
| `read`     | SELECT, COPY TO                                       |
| `write`    | INSERT, UPDATE, DELETE, TRUNCATE, COPY FROM           |
| `function` | Function calls and DO blocks                          |
| `role`     | GRANT, REVOKE, CREATE/ALTER/DROP ROLE                 |
| `ddl`      | All DDL not covered by ROLE                           |
| `misc`     | DISCARD, FETCH, CHECKPOINT, VACUUM                    |
| `misc_set` | SET commands                                          |
| `all`      | All of the above                                      |

### Object-Level Auditing

```sql
-- Create an auditor role
CREATE ROLE auditor NOLOGIN;

-- Set the object audit role
-- pgaudit.role = 'auditor'  -- in postgresql.conf

-- Grant specific access to the auditor role (triggers object-level audit)
GRANT SELECT ON customers TO auditor;     -- logs all SELECTs on customers
GRANT ALL ON orders TO auditor;           -- logs all DML on orders
-- Only operations matching the granted permissions are logged
```

### Built-in PostgreSQL Logging

```ini
# postgresql.conf - basic audit-related settings

# Log all DDL statements
log_statement = 'ddl'
# Options: none, ddl, mod (ddl + data-modifying), all

# Log all connections and disconnections
log_connections = on
log_disconnections = on

# Log duration of completed statements
log_duration = on

# Log statements exceeding a threshold (ms)
log_min_duration_statement = 1000  # log queries taking > 1 second

# Log line prefix with useful context
log_line_prefix = '%t [%p]: user=%u,db=%d,app=%a,client=%h '

# Log checkpoints
log_checkpoints = on

# Log lock waits
log_lock_waits = on
deadlock_timeout = 1s
```

### Custom Audit Trigger

```sql
-- Generic audit trigger for tracking row changes
CREATE TABLE audit_log (
  id bigserial PRIMARY KEY,
  table_name text NOT NULL,
  operation text NOT NULL,
  old_data jsonb,
  new_data jsonb,
  changed_by text DEFAULT current_user,
  changed_at timestamptz DEFAULT now(),
  client_ip inet DEFAULT inet_client_addr()
);

CREATE OR REPLACE FUNCTION audit_trigger_func()
RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
  IF TG_OP = 'DELETE' THEN
    INSERT INTO audit_log (table_name, operation, old_data)
    VALUES (TG_TABLE_NAME, TG_OP, to_jsonb(OLD));
    RETURN OLD;
  ELSIF TG_OP = 'UPDATE' THEN
    INSERT INTO audit_log (table_name, operation, old_data, new_data)
    VALUES (TG_TABLE_NAME, TG_OP, to_jsonb(OLD), to_jsonb(NEW));
    RETURN NEW;
  ELSIF TG_OP = 'INSERT' THEN
    INSERT INTO audit_log (table_name, operation, new_data)
    VALUES (TG_TABLE_NAME, TG_OP, to_jsonb(NEW));
    RETURN NEW;
  END IF;
END;
$$;

-- Apply to sensitive tables
CREATE TRIGGER audit_customers
  AFTER INSERT OR UPDATE OR DELETE ON customers
  FOR EACH ROW EXECUTE FUNCTION audit_trigger_func();

CREATE TRIGGER audit_orders
  AFTER INSERT OR UPDATE OR DELETE ON orders
  FOR EACH ROW EXECUTE FUNCTION audit_trigger_func();
```

### What to Audit

- **Always audit**: DDL changes, role/permission changes, login failures.
- **Usually audit**: Writes to sensitive tables (financial, PII, authentication).
- **Selectively audit**: Reads on sensitive data (customer records, medical data).
- **Avoid auditing**: High-volume read-only queries on non-sensitive data (noise).

---

## 8. Network Security

### Binding to Specific Interfaces

```ini
# postgresql.conf
# Only listen on specific interfaces (never use '*' in production without firewall)
listen_addresses = '10.0.1.5'          # single interface
listen_addresses = '10.0.1.5,10.0.1.6' # multiple interfaces
listen_addresses = 'localhost'          # local only (default)
```

### Firewall Rules (iptables/nftables)

```bash
# Allow PostgreSQL only from application subnet
iptables -A INPUT -p tcp --dport 5432 -s 10.0.1.0/24 -j ACCEPT
iptables -A INPUT -p tcp --dport 5432 -j DROP

# nftables equivalent
nft add rule inet filter input tcp dport 5432 ip saddr 10.0.1.0/24 accept
nft add rule inet filter input tcp dport 5432 drop
```

### SSH Tunneling

```bash
# Create SSH tunnel to access remote PostgreSQL securely
ssh -L 5432:localhost:5432 user@db-server.example.com -N

# Connect via tunnel
psql -h localhost -p 5432 -U web_app mydb
```

### Cloud Security Groups (AWS Example)

```
Inbound Rules:
  Type: PostgreSQL  |  Port: 5432  |  Source: sg-app-servers
  Type: PostgreSQL  |  Port: 5432  |  Source: 10.0.2.0/24 (VPN subnet)

Outbound Rules:
  Type: All traffic  |  Destination: 0.0.0.0/0  (for updates/replication)
```

### Best Practices

- Run PostgreSQL on a private subnet with no public IP.
- Use a bastion host or VPN for administrative access.
- Use separate security groups for application and admin access.
- Enable VPC flow logs to monitor connection attempts.
- Use connection poolers (PgBouncer) in front of PostgreSQL — they can add TLS
  termination and connection limiting.

---

## 9. Data Encryption

### Transparent Data Encryption (TDE)

PostgreSQL does not have built-in TDE in the community edition. Options include:

```
Option 1: Filesystem-level encryption
  - LUKS (Linux Unified Key Setup) for full-disk encryption
  - dm-crypt for block-level encryption
  - AWS EBS encryption, Azure Disk Encryption, GCP CMEK

Option 2: PostgreSQL TDE patches/forks
  - CyberTec TDE for PostgreSQL
  - Fujitsu Enterprise Postgres with TDE
  - PostgreSQL 16+ has experimental cluster-level encryption support

Option 3: Tablespace-level encryption
  - Use encrypted filesystem mounted as a tablespace
  CREATE TABLESPACE encrypted_space
    LOCATION '/mnt/encrypted_volume/pg_data';
  ALTER TABLE sensitive_data SET TABLESPACE encrypted_space;
```

### Column-Level Encryption with pgcrypto

```sql
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Symmetric encryption (AES-256)
-- Encrypt
INSERT INTO customers (name, ssn_encrypted)
VALUES (
  'John Doe',
  pgp_sym_encrypt('123-45-6789', 'encryption_key_here')
);

-- Decrypt
SELECT name, pgp_sym_decrypt(ssn_encrypted, 'encryption_key_here') AS ssn
FROM customers
WHERE id = 1;

-- Asymmetric encryption (PGP public/private key)
-- Encrypt with public key (anyone can encrypt)
INSERT INTO messages (content_encrypted)
VALUES (pgp_pub_encrypt('secret message', dearmor(pg_read_file('public.key'))));

-- Decrypt with private key (only key holder can decrypt)
SELECT pgp_pub_decrypt(
  content_encrypted,
  dearmor(pg_read_file('private.key')),
  'key_passphrase'
) AS content
FROM messages WHERE id = 1;

-- Hashing (one-way, for password storage or integrity)
-- Use crypt() for password hashing (bcrypt-based)
INSERT INTO app_passwords (user_id, password_hash)
VALUES (1, crypt('user_password', gen_salt('bf', 12)));
-- 'bf' = blowfish/bcrypt, 12 = cost factor

-- Verify password
SELECT user_id FROM app_passwords
WHERE password_hash = crypt('user_password', password_hash);

-- Generic hashing
SELECT digest('data to hash', 'sha256');
SELECT encode(digest('data to hash', 'sha256'), 'hex');
```

### Key Management Best Practices

1. **Never store encryption keys in the database.** Use a secrets manager
   (HashiCorp Vault, AWS KMS, Azure Key Vault).
2. **Rotate keys periodically.** Re-encrypt data when rotating symmetric keys.
3. **Use envelope encryption**: Encrypt data with a data encryption key (DEK),
   encrypt the DEK with a key encryption key (KEK) stored in the secrets manager.
4. **Limit key access**: Only the application role should have access to decryption
   functions. Revoke `EXECUTE` on pgcrypto functions from general roles.

```sql
-- Restrict pgcrypto function access
REVOKE EXECUTE ON ALL FUNCTIONS IN SCHEMA public FROM PUBLIC;
GRANT EXECUTE ON FUNCTION pgp_sym_decrypt(bytea, text) TO app_readwrite;
-- Only app_readwrite can decrypt
```

### Searching Encrypted Data

Encrypted columns cannot be indexed or searched directly. Common patterns:

```sql
-- Option 1: Store a deterministic hash for equality lookups
ALTER TABLE customers ADD COLUMN ssn_hash bytea;
CREATE INDEX idx_customers_ssn_hash ON customers (ssn_hash);

UPDATE customers SET ssn_hash = digest(ssn_plaintext, 'sha256');
-- Search by hash (cannot retrieve the value without decrypting)
SELECT * FROM customers
WHERE ssn_hash = digest('123-45-6789', 'sha256');

-- Option 2: Store a prefix or tokenized version for partial lookups
ALTER TABLE customers ADD COLUMN ssn_last4 char(4);
-- Store only last 4 digits unencrypted for display/lookup

-- Option 3: Use blind indexing with HMAC
ALTER TABLE customers ADD COLUMN ssn_blind_index bytea;
UPDATE customers
SET ssn_blind_index = hmac(ssn_plaintext, 'blind_index_key', 'sha256');
CREATE INDEX idx_customers_ssn_blind ON customers (ssn_blind_index);
```
