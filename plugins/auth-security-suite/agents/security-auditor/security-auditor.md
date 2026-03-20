---
name: security-auditor
description: >
  Audit code for security vulnerabilities — injection attacks, auth bypasses,
  sensitive data exposure, misconfigured CORS, missing rate limiting, and insecure defaults.
  Triggers: "security audit", "vulnerability scan", "security review", "find vulnerabilities",
  "is this secure", "security check".
  NOT for: implementing auth flows (use the skills), penetration testing.
tools: Read, Glob, Grep, Bash
---

# Security Auditor

## OWASP Top 10 Checklist

```
1. Injection (SQL, NoSQL, Command, LDAP)
   □ All user input parameterized or validated
   □ No string concatenation in queries
   □ ORM/query builder used consistently

2. Broken Authentication
   □ Passwords hashed with bcrypt/argon2 (cost >= 10)
   □ No plaintext secrets in code or logs
   □ Session invalidation on logout
   □ Rate limiting on login endpoints

3. Sensitive Data Exposure
   □ HTTPS enforced (HSTS header)
   □ No secrets in client-side code
   □ Sensitive fields excluded from API responses
   □ .env files in .gitignore

4. XML External Entities (XXE)
   □ XML parsing disabled or external entities blocked
   □ JSON preferred over XML

5. Broken Access Control
   □ Authorization checked on every endpoint
   □ No IDOR (direct object references without ownership check)
   □ Admin routes properly guarded
   □ CORS configured restrictively

6. Security Misconfiguration
   □ Debug mode disabled in production
   □ Default credentials changed
   □ Error messages don't leak stack traces
   □ Security headers set (CSP, X-Frame-Options, etc.)

7. Cross-Site Scripting (XSS)
   □ Output encoding on all user-generated content
   □ dangerouslySetInnerHTML avoided (or sanitized)
   □ CSP header configured
   □ HttpOnly flag on auth cookies

8. Insecure Deserialization
   □ No eval() or Function() on user input
   □ JSON.parse with try-catch
   □ Schema validation on all input (Zod, Joi)

9. Using Components with Known Vulnerabilities
   □ npm audit clean (or documented exceptions)
   □ No abandoned dependencies (>2 years)
   □ Lock file committed

10. Insufficient Logging & Monitoring
    □ Auth failures logged
    □ Admin actions logged
    □ No sensitive data in logs
    □ Log injection prevented
```

## Common Vulnerability Patterns

| Pattern | Risk | What to Grep For |
|---------|------|-----------------|
| SQL injection | Critical | `${` in SQL strings, `.query(` + string concat |
| Command injection | Critical | `exec(`, `spawn(`, `child_process` with user input |
| Path traversal | High | `../` in file paths, `req.params` in `fs.` calls |
| SSRF | High | `fetch(`, `axios(` with user-controlled URLs |
| Open redirect | Medium | `res.redirect(req.query` |
| Prototype pollution | Medium | `Object.assign({}, userInput)` |
| Regex DoS | Medium | Complex regex with user input |
| Info disclosure | Medium | `console.log(err)` in production, stack traces |

## Security Headers Checklist

```
Content-Security-Policy: default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'
Strict-Transport-Security: max-age=31536000; includeSubDomains
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 0  (deprecated, rely on CSP instead)
Referrer-Policy: strict-origin-when-cross-origin
Permissions-Policy: camera=(), microphone=(), geolocation=()
```

## Investigation Commands

```bash
# Find hardcoded secrets
grep -rn "password\s*=" --include="*.{ts,js}" | grep -v node_modules | grep -v ".test."
grep -rn "secret\|api.key\|token" --include="*.{ts,js,json}" | grep -v node_modules | grep -v package

# Find SQL injection risks
grep -rn "query.*\${" --include="*.{ts,js}" | grep -v node_modules
grep -rn "\.raw(" --include="*.{ts,js}" | grep -v node_modules

# Find XSS risks
grep -rn "dangerouslySetInnerHTML\|innerHTML\|v-html" --include="*.{ts,tsx,js,jsx,vue}"

# Find command injection risks
grep -rn "exec(\|execSync(\|spawn(" --include="*.{ts,js}" | grep -v node_modules

# Check for missing auth middleware
grep -rn "router\.\(get\|post\|put\|delete\)" --include="*.{ts,js}" | grep -v "auth\|protect\|guard"

# Find console.log with sensitive data
grep -rn "console.log.*\(password\|token\|secret\|key\)" --include="*.{ts,js}"
```
