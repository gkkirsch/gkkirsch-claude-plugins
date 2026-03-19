# Auth Design Skill

## Metadata
- Name: auth-design
- Description: Design and implement secure authentication and authorization systems — OAuth2, JWT, RBAC, API security, and identity management
- Version: 1.0.0

## Trigger
Activate when the user asks about:
- Authentication (login, signup, OAuth, SSO, social login, passkeys)
- Authorization (permissions, roles, RBAC, ABAC, access control)
- JWT (tokens, signing, verification, refresh, rotation)
- Session management (cookies, session stores, session security)
- API security (rate limiting, CORS, CSRF, security headers, input validation)
- Identity management (MFA, password reset, account recovery, user management)
- Webhooks (signature verification, HMAC)
- Security hardening for web applications

## Agents

### auth-architect
**When to use:** OAuth2/OIDC implementation, JWT design, session vs token decisions, passkey/WebAuthn setup, SSO integration, authentication flow design.

**Capabilities:**
- All OAuth2 flows with PKCE (Authorization Code, Client Credentials, Device Grant)
- JWT signing with RS256/ES256, key rotation, JWKS endpoints
- Session-based auth with Redis/PostgreSQL stores
- Passkey/WebAuthn registration and authentication
- SAML 2.0 and OIDC federation for enterprise SSO
- Magic link and passwordless authentication
- API key generation, hashing, and validation
- Refresh token rotation with reuse detection
- Password hashing with Argon2id

### rbac-designer
**When to use:** Permission systems, role hierarchies, policy engines, multi-tenant authorization, row-level security.

**Capabilities:**
- RBAC with hierarchical roles and separation of duties
- ABAC policy engines with dynamic attribute evaluation
- Integration with OPA (Rego policies), Casbin, Cedar
- Multi-tenant authorization with row-level security
- Database schemas for roles, permissions, and assignments
- Permission caching with Redis
- React permission components (Can, HasRole, usePermissions)
- Authorization audit logging

### api-security-engineer
**When to use:** Rate limiting, CORS, CSRF, security headers, input validation, webhook verification, API hardening.

**Capabilities:**
- Rate limiting algorithms (sliding window, token bucket) with Redis
- CORS configuration for SPAs and multi-tenant subdomains
- CSRF protection (double submit cookie, synchronizer token)
- Security headers (CSP, HSTS, Permissions-Policy, X-Content-Type-Options)
- Input validation with Zod (Node.js) and Pydantic (Python)
- Webhook signature verification (HMAC-SHA256, Stripe, GitHub)
- HMAC request signing for API-to-API authentication
- Bot detection and abuse prevention
- Complete API security middleware stack

### identity-manager
**When to use:** User registration, MFA, account recovery, social login, invitations, GDPR compliance.

**Capabilities:**
- User registration with email verification
- Breached password checking (Have I Been Pwned k-anonymity API)
- TOTP MFA with encrypted secret storage and backup codes
- Password reset with one-time tokens and rate limiting
- Social login with account linking (Google, GitHub, Apple, Microsoft)
- Login flow with MFA challenge
- Team invitation system
- GDPR data export and account deletion
- Account lockout after failed attempts

## References

### oauth2-flows
All OAuth2 authorization flows with ASCII protocol diagrams, code examples in Node.js/Python/Go, security checklists, and provider-specific configuration (Google, GitHub, Microsoft, Apple).

### jwt-security
JWT structure, signing algorithm comparison, claim validation checklist, key rotation strategy, token revocation approaches, common vulnerabilities (algorithm confusion, none attack, kid injection), and implementation across Node.js/Python/Go.

### session-management
Cookie security attributes, session store comparison (Redis vs PostgreSQL vs memory), session lifecycle management, security hardening (fixation, hijacking, timeouts), concurrent session management, and cleanup strategies.

## Workflow

1. **Assess** — Understand the application type (SPA, server, mobile, API)
2. **Recommend** — Suggest the appropriate auth strategy
3. **Design** — Provide database schemas, architecture diagrams, flow diagrams
4. **Implement** — Deliver production-ready code with error handling
5. **Harden** — Add security middleware, rate limiting, headers, validation
6. **Review** — Check against OWASP guidelines and security best practices

## Languages

All agents provide code in:
- Node.js (Express) with TypeScript where appropriate
- Python (FastAPI) with Pydantic validation
- Go (net/http) with standard library patterns

## Quality Standards

- Every code example includes error handling
- All sensitive operations include audit logging
- Security best practices are enforced by default
- Database schemas include proper indexes
- Rate limiting is applied to all auth endpoints
- Constant-time comparison for all secret comparisons
- HTTPS enforced everywhere
