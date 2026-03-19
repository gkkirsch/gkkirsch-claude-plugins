---
name: security-audit-suite
description: >
  Security & Code Audit Suite — AI-powered toolkit for vulnerability scanning, dependency auditing,
  security-focused code review, and regulatory compliance checking. Detects OWASP Top 10 vulnerabilities,
  audits npm/pip/maven/cargo dependencies for CVEs, reviews authentication and cryptography implementations,
  and maps code against PCI DSS, GDPR, HIPAA, SOC 2, and NIST frameworks.
  Triggers: "security audit", "vulnerability scan", "find vulnerabilities", "security review",
  "dependency audit", "check dependencies", "npm audit", "supply chain", "CVE check",
  "code review security", "auth review", "crypto review", "secrets detection",
  "compliance check", "PCI DSS", "GDPR", "HIPAA", "SOC 2", "NIST", "OWASP",
  "ASVS", "CIS benchmark", "ISO 27001", "security scan", "pentest", "penetration test".
  Dispatches the appropriate specialist agent: vulnerability-scanner, dependency-auditor,
  code-reviewer-security, or compliance-checker.
  NOT for: Network infrastructure security, firewall configuration, cloud IAM policies,
  or runtime application monitoring/WAF rules.
version: 1.0.0
argument-hint: "<scan|deps|review|compliance> [path-or-scope]"
user-invocable: true
allowed-tools: Read, Grep, Glob, Bash
model: sonnet
---

# Security & Code Audit Suite

Production-grade security analysis and compliance checking agents for Claude Code. Four specialist agents that handle vulnerability detection, dependency auditing, security code review, and regulatory compliance — the security work that every production codebase needs.

## Available Agents

### Vulnerability Scanner (`vulnerability-scanner`)
Detects security vulnerabilities in application code. Covers SQL injection, XSS (stored/reflected/DOM), CSRF, command injection, path traversal, SSRF, insecure deserialization, XXE, broken access control, authentication bypass, session flaws, IDOR, race conditions, file upload vulnerabilities, header injection, CRLF injection, open redirects, and clickjacking. Produces severity-ranked findings with exploit scenarios and remediation code.

**Invoke**: Dispatch via Task tool with `subagent_type: "vulnerability-scanner"`.

**Example prompts**:
- "Scan my codebase for security vulnerabilities"
- "Check my Express app for injection attacks"
- "Find XSS vulnerabilities in my React application"
- "Audit my authentication flow for bypass vectors"

### Dependency Auditor (`dependency-auditor`)
Supply chain security analysis. Audits npm/pip/maven/cargo dependencies for known CVEs, detects typosquatting, identifies malicious package indicators, analyzes transitive dependency risk, verifies lockfiles, generates SBOMs, scores vulnerabilities by CVSS, and plans upgrade paths.

**Invoke**: Dispatch via Task tool with `subagent_type: "dependency-auditor"`.

**Example prompts**:
- "Audit my npm dependencies for vulnerabilities"
- "Check my Python requirements for known CVEs"
- "Generate an SBOM for this project"
- "Plan upgrade paths for my vulnerable dependencies"

### Security Code Reviewer (`code-reviewer-security`)
Security-focused code review across authentication, authorization, cryptography, input validation, output encoding, error handling, logging, API security, secrets detection, and infrastructure-as-code. Reviews password hashing, session management, MFA implementation, RBAC/ABAC, key management, TLS configuration, and more.

**Invoke**: Dispatch via Task tool with `subagent_type: "code-reviewer-security"`.

**Example prompts**:
- "Review my authentication implementation for security issues"
- "Check my code for hardcoded secrets and credentials"
- "Review my cryptography usage — am I using the right algorithms?"
- "Audit my authorization logic for privilege escalation"

### Compliance Checker (`compliance-checker`)
Maps codebases against regulatory and industry frameworks. Covers OWASP ASVS, PCI DSS, GDPR technical controls, HIPAA security requirements, SOC 2, CIS benchmarks, NIST Cybersecurity Framework, and ISO 27001. Generates compliance reports, gap analysis, and evidence documentation.

**Invoke**: Dispatch via Task tool with `subagent_type: "compliance-checker"`.

**Example prompts**:
- "Check my app against OWASP ASVS requirements"
- "Run a PCI DSS compliance check on my payment processing code"
- "What GDPR technical controls am I missing?"
- "Map my security controls to the NIST Cybersecurity Framework"

## Quick Start: /security-audit

Use the `/security-audit` command for guided security analysis:

```
/security-audit                          # Full security scan of the codebase
/security-audit scan                     # Vulnerability scan only
/security-audit scan src/api/            # Scan specific directory
/security-audit deps                     # Dependency audit
/security-audit review src/auth/         # Security code review of auth module
/security-audit compliance --framework pci-dss   # PCI DSS compliance check
/security-audit compliance --framework gdpr      # GDPR compliance check
```

## Agent Selection Guide

| Need | Agent | Command |
|------|-------|---------|
| Find vulnerabilities | vulnerability-scanner | "Scan for vulnerabilities" |
| OWASP Top 10 check | vulnerability-scanner | "Check for OWASP Top 10" |
| Injection detection | vulnerability-scanner | "Find injection vulnerabilities" |
| XSS detection | vulnerability-scanner | "Find XSS vulnerabilities" |
| Dependency CVEs | dependency-auditor | "Audit dependencies" |
| Supply chain risk | dependency-auditor | "Check supply chain security" |
| SBOM generation | dependency-auditor | "Generate SBOM" |
| Auth review | code-reviewer-security | "Review authentication" |
| Secrets detection | code-reviewer-security | "Find hardcoded secrets" |
| Crypto review | code-reviewer-security | "Review cryptography usage" |
| PCI DSS | compliance-checker | "PCI DSS compliance check" |
| GDPR | compliance-checker | "GDPR compliance check" |
| HIPAA | compliance-checker | "HIPAA compliance check" |
| SOC 2 | compliance-checker | "SOC 2 compliance check" |
| NIST CSF | compliance-checker | "NIST framework mapping" |

## Reference Materials

This skill includes comprehensive reference documents in `references/`:

- **owasp-top-10.md** — Complete OWASP Top 10 (2021) reference with detection techniques, prevention controls, code examples in multiple languages, and testing methodology
- **secure-coding-patterns.md** — Language-specific secure coding patterns for input validation, output encoding, parameterized queries, session management, CORS, CSP, cookie security, cryptography, and secret management
- **cve-database-guide.md** — Working with CVE/CWE/CVSS, NVD API, GitHub Advisory Database, OSV.dev, Snyk DB, reading advisories, prioritizing remediation, and automated scanning setup

Agents automatically consult these references when working. You can also read them directly for quick answers.

## How It Works

1. You describe what you need (e.g., "scan for vulnerabilities")
2. The SKILL.md routes to the appropriate agent
3. The agent reads your code, identifies the tech stack, and performs deep analysis
4. Findings are generated with severity rankings, evidence, and remediation
5. A report is produced with prioritized action items

All generated reports follow industry standards:
- Vulnerability findings: CVSS-scored, CWE-mapped, with exploit proof-of-concept and fix code
- Dependency audits: CVE-linked, upgrade paths, transitive risk analysis
- Code reviews: Categorized by security domain, severity-ranked, with before/after code
- Compliance reports: Framework-mapped, gap analysis, evidence documentation
