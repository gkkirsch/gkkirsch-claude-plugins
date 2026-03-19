---
name: security-audit
description: >
  Quick security audit command — analyzes your codebase for vulnerabilities, dependency risks,
  authentication flaws, and compliance gaps. Routes to specialist agents based on the audit type.
  Triggers: "/security-audit", "scan for vulnerabilities", "security check", "audit security".
user-invocable: true
argument-hint: "<scan|deps|review|compliance> [path] [--framework pci-dss|gdpr|hipaa|soc2|nist]"
allowed-tools: Read, Write, Edit, Bash, Glob, Grep
model: sonnet
---

# /security-audit Command

One-command security analysis for any codebase. Detects the tech stack, identifies security-relevant code paths, and dispatches the appropriate specialist agent for deep analysis.

## Usage

```
/security-audit                                    # Full security scan
/security-audit scan                               # Vulnerability scan only
/security-audit scan src/api/                      # Scan specific directory
/security-audit deps                               # Dependency audit
/security-audit deps --fix                         # Audit and auto-fix where safe
/security-audit review                             # Security-focused code review
/security-audit review src/auth/                   # Review specific module
/security-audit compliance                         # All compliance frameworks
/security-audit compliance --framework pci-dss     # Specific framework
/security-audit compliance --framework gdpr        # GDPR check
/security-audit compliance --framework hipaa       # HIPAA check
```

## Subcommands

| Subcommand | Agent | Description |
|------------|-------|-------------|
| `scan` | vulnerability-scanner | Scan for OWASP Top 10, injections, auth flaws, SSRF, etc. |
| `deps` | dependency-auditor | Audit dependencies for CVEs, typosquatting, license issues |
| `review` | code-reviewer-security | Deep security code review — auth, crypto, secrets, IaC |
| `compliance` | compliance-checker | Map against PCI DSS, GDPR, HIPAA, SOC 2, NIST, ASVS |

## Procedure

### Step 1: Detect Technology Stack

Read the project root to understand the codebase:

1. **Read** `package.json`, `requirements.txt`, `go.mod`, `Cargo.toml`, `Gemfile`, `pom.xml`, `build.gradle` — determine language/framework
2. **Glob** for security-relevant files:
   - Auth: `**/auth/**`, `**/login.*`, `**/session.*`, `**/jwt.*`, `**/middleware/auth*`
   - Database: `**/models/**`, `**/migrations/**`, `**/queries/**`, `**/*repository*`
   - API: `**/routes/**`, `**/controllers/**`, `**/handlers/**`, `**/api/**`
   - Config: `**/config/**`, `**/.env*`, `**/settings.*`, `**/docker-compose*`
   - IaC: `**/terraform/**`, `**/cloudformation/**`, `**/k8s/**`, `**/helm/**`
3. **Grep** for security-sensitive patterns:
   - SQL queries: `query\(`, `execute\(`, `raw\(`, `\$\{.*\}.*SELECT|INSERT|UPDATE|DELETE`
   - Crypto: `createHash`, `createCipher`, `encrypt`, `decrypt`, `bcrypt`, `argon2`
   - Auth: `jwt.sign`, `jwt.verify`, `passport`, `session`, `cookie`, `token`
   - Exec: `exec\(`, `spawn\(`, `system\(`, `eval\(`, `subprocess`
   - File ops: `readFile`, `writeFile`, `unlink`, `path.join`, `open\(`
4. **Check for security tooling**: `.snyk`, `.npmrc`, `audit-ci.json`, `trivy.yaml`, `.github/workflows/*security*`
5. **Check for existing reports**: `**/security-audit*`, `**/vulnerability-report*`

Report findings:

```
Security Scan Context:
- Language: TypeScript (Node.js)
- Framework: Express + Prisma
- Auth: JWT with passport.js
- Database: PostgreSQL via Prisma ORM
- Security-sensitive files: 47
- Existing security tooling: npm audit (CI), no SAST
- Previous audit reports: None found
```

### Step 2: Route to Agent

Based on the subcommand (or `scan` by default), dispatch the appropriate agent:

#### `scan` — Vulnerability Scanner

```
Task tool:
  subagent_type: "vulnerability-scanner"
  mode: "bypassPermissions"
  prompt: |
    Scan this codebase for security vulnerabilities.
    Stack: [detected stack]
    Security-relevant files: [discovered files]
    Auth pattern: [detected auth]
    Target: [specific path if provided, or full codebase]
    Check for all OWASP Top 10 categories plus:
    - Injection (SQL, NoSQL, command, LDAP, XPath)
    - XSS (stored, reflected, DOM-based)
    - Authentication and session management flaws
    - SSRF, path traversal, open redirects
    - Insecure deserialization, XXE
    - Race conditions and TOCTOU bugs
    - File upload vulnerabilities
    Produce a severity-ranked report with evidence and remediation.
```

#### `deps` — Dependency Auditor

```
Task tool:
  subagent_type: "dependency-auditor"
  mode: "bypassPermissions"
  prompt: |
    Audit all dependencies in this project for security risks.
    Stack: [detected stack]
    Package manager: [npm/pip/maven/cargo]
    Lockfile: [present/missing]
    Perform:
    - CVE lookup for all direct and transitive dependencies
    - Typosquatting detection
    - License compliance check
    - Outdated dependency identification
    - SBOM generation
    - Upgrade path planning for vulnerable packages
    Generate a prioritized report with fix commands.
```

#### `review` — Security Code Reviewer

```
Task tool:
  subagent_type: "code-reviewer-security"
  mode: "bypassPermissions"
  prompt: |
    Perform a security-focused code review.
    Stack: [detected stack]
    Target: [specific module if provided, or full codebase]
    Auth pattern: [detected auth]
    Review domains:
    - Authentication (password hashing, session management, MFA)
    - Authorization (RBAC/ABAC, privilege escalation)
    - Cryptography (algorithms, key management, TLS)
    - Input validation and output encoding
    - Error handling and information leakage
    - Logging (sensitive data in logs)
    - Secrets in code
    - API security
    - Infrastructure-as-code security
    Produce categorized findings with before/after code fixes.
```

#### `compliance` — Compliance Checker

```
Task tool:
  subagent_type: "compliance-checker"
  mode: "bypassPermissions"
  prompt: |
    Check this codebase against compliance requirements.
    Stack: [detected stack]
    Framework: [specified framework or all]
    Target frameworks: [pci-dss, gdpr, hipaa, soc2, nist, asvs, iso27001]
    Perform:
    - Control mapping (code controls to framework requirements)
    - Gap analysis (missing controls)
    - Evidence generation (code references proving compliance)
    - Remediation guidance for gaps
    Generate a compliance report with pass/fail/partial status per control.
```

### Step 3: Results Summary

After the agent completes, present results:

**For vulnerability scan:**
```
Vulnerability Scan Results:
- Critical: 2 (SQL injection in search, command injection in export)
- High: 3 (Broken auth on admin API, SSRF in webhook handler, stored XSS in comments)
- Medium: 5 (Missing CSRF tokens, verbose errors, weak session config, ...)
- Low: 8 (Missing security headers, no rate limiting on public endpoints, ...)
- Total: 18 findings
- Report: docs/security/vulnerability-scan.md
```

**For dependency audit:**
```
Dependency Audit Results:
- Critical CVEs: 1 (lodash prototype pollution — CVE-2021-23337)
- High CVEs: 3 (jsonwebtoken, express-jwt, node-forge)
- Outdated packages: 12 (4 with known vulnerabilities)
- License issues: 2 (GPL-3.0 in MIT project)
- Typosquatting risk: 0
- Fix: npm audit fix --force (handles 3/4 critical+high)
- Report: docs/security/dependency-audit.md
```

**For security code review:**
```
Security Code Review Results:
- Authentication: 2 issues (weak password policy, no account lockout)
- Authorization: 1 issue (missing role check on /admin/export)
- Cryptography: 1 issue (using SHA-256 for passwords instead of bcrypt)
- Input validation: 3 issues (missing validation on 3 POST endpoints)
- Secrets: 1 issue (API key hardcoded in config.ts)
- Total: 8 findings (2 critical, 3 high, 3 medium)
- Report: docs/security/code-review.md
```

**For compliance check:**
```
PCI DSS Compliance Results:
- Requirement 2 (Default passwords): PASS
- Requirement 3 (Stored cardholder data): FAIL — card numbers stored without encryption
- Requirement 4 (Encryption in transit): PASS — TLS 1.2+ enforced
- Requirement 6 (Secure development): PARTIAL — missing input validation on 3 endpoints
- Requirement 8 (Authentication): PARTIAL — no MFA for admin access
- Overall: 6/12 requirements met, 3 partial, 3 failing
- Report: docs/security/pci-dss-compliance.md
```

## Error Recovery

| Error | Cause | Fix |
|-------|-------|-----|
| No source code found | Wrong directory | Provide correct project path |
| No dependencies found | Non-standard package manager | Specify package manager manually |
| Compliance framework unknown | Typo in --framework | Use: pci-dss, gdpr, hipaa, soc2, nist, asvs, iso27001 |
| Too many findings | Large codebase | Scope to specific directory |
| False positives | Framework handles it | Review and dismiss with justification |

## Notes

- Security audits analyze code statically — they do not execute exploits or make network requests
- Findings should be triaged by a security professional before remediation
- Compliance checks provide guidance, not legal certification
- Dependency audits reflect known CVEs at scan time — re-run periodically
- All reports are written to `docs/security/` by default
