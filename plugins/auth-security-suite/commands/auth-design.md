# /auth-design Command

Design and implement secure authentication and authorization systems. This command activates the Authentication & Authorization Suite agents to help you build production-grade auth.

## Usage

```
/auth-design [subcommand] [options]
```

## Subcommands

### `oauth`
Design OAuth2/OIDC authentication flows.

```
/auth-design oauth
```

Activates the **auth-architect** agent to help you:
- Choose the right OAuth2 flow for your app type (SPA, server, mobile, CLI)
- Implement Authorization Code + PKCE
- Configure social login providers (Google, GitHub, Apple, Microsoft)
- Set up OIDC discovery and token validation
- Implement token refresh with rotation and reuse detection

### `jwt`
Design JWT token systems with signing, verification, and rotation.

```
/auth-design jwt
```

Activates the **auth-architect** agent to help you:
- Choose signing algorithms (RS256, ES256, HS256)
- Implement JWT signing and verification
- Set up JWKS endpoints and key rotation
- Design access + refresh token strategies
- Build token revocation (blocklist, versioning)

### `rbac`
Design role-based or attribute-based access control systems.

```
/auth-design rbac
```

Activates the **rbac-designer** agent to help you:
- Design role hierarchies and permission models
- Implement RBAC with database schemas
- Build ABAC policy engines for dynamic access control
- Integrate policy engines (OPA, Casbin, Cedar)
- Set up multi-tenant authorization with tenant isolation
- Build frontend permission components

### `api`
Harden API security with rate limiting, CORS, CSRF, and validation.

```
/auth-design api
```

Activates the **api-security-engineer** agent to help you:
- Implement rate limiting (token bucket, sliding window)
- Configure CORS securely for your origin setup
- Add CSRF protection for session-based APIs
- Set security headers (CSP, HSTS, Permissions-Policy)
- Build input validation schemas with Zod/Pydantic
- Implement webhook signature verification
- Set up the complete API security middleware stack

### `identity`
Design user management, MFA, and account recovery systems.

```
/auth-design identity
```

Activates the **identity-manager** agent to help you:
- Build user registration with email verification
- Implement MFA (TOTP, backup codes, passkeys)
- Design password reset and account recovery flows
- Set up social login with account linking
- Build invitation and team membership systems
- Implement GDPR data export and deletion
- Design login flows with MFA challenge

### `full`
Design a complete authentication system from scratch.

```
/auth-design full
```

Activates all agents in sequence to design your entire auth system:
1. **Auth Architect** — Authentication strategy and token design
2. **RBAC Designer** — Permission model and access control
3. **API Security Engineer** — API hardening and middleware
4. **Identity Manager** — User lifecycle and MFA

### `audit`
Review existing auth implementation for security issues.

```
/auth-design audit
```

Reviews your codebase for common auth vulnerabilities:
- Insecure token storage (localStorage)
- Missing CSRF protection
- Weak password hashing (MD5, SHA-256, weak bcrypt)
- Missing rate limiting on auth endpoints
- CORS misconfiguration
- Missing security headers
- JWT algorithm confusion risks
- Session fixation vulnerabilities

## Examples

```
# Design OAuth2 login for a React SPA
/auth-design oauth

# Add RBAC to an existing Express API
/auth-design rbac

# Implement MFA for user accounts
/auth-design identity

# Harden API security with rate limiting and CORS
/auth-design api

# Build complete auth system for a new SaaS app
/auth-design full

# Review existing auth code for vulnerabilities
/auth-design audit
```

## Reference Files

The suite includes detailed reference documents:
- **oauth2-flows.md** — All OAuth2 flows with protocol diagrams and code
- **jwt-security.md** — JWT best practices, algorithms, claims, rotation
- **session-management.md** — Session strategies, cookies, storage, security

## Supported Languages

All agents provide production-ready code in:
- **Node.js** (Express, Fastify)
- **Python** (FastAPI, Django)
- **Go** (net/http, Gin, Echo)

## What This Suite Covers

```
Authentication                 Authorization
├── OAuth2/OIDC               ├── RBAC (Role-Based)
├── JWT tokens                ├── ABAC (Attribute-Based)
├── Sessions & cookies        ├── ReBAC (Relationship-Based)
├── Passkeys/WebAuthn         ├── Policy engines (OPA, Casbin)
├── Social login              ├── Multi-tenant isolation
├── Magic links               ├── Row-level security
├── API keys                  └── Frontend permissions
├── MFA (TOTP, backup codes)
├── Password hashing          API Security
├── Account recovery          ├── Rate limiting
├── Email verification        ├── CORS
└── SSO (SAML, OIDC)         ├── CSRF
                              ├── Security headers
Identity Management           ├── Input validation
├── User registration         ├── Webhook verification
├── Profile management        ├── Request signing
├── Account linking           └── Bot protection
├── Invitations
├── Data export (GDPR)
└── Account deletion
```
