---
name: pro-toolkit
description: >
  Claude Code Pro Toolkit — 5 production-ready agents for professional development workflows.
  Triggers: "security audit", "generate tests", "document code", "performance analysis", "migrate framework".
  Dispatches the appropriate specialist agent based on your request.
  NOT for: general coding help, chat, or tasks unrelated to security/testing/docs/performance/migration.
version: 1.0.0
argument-hint: "<audit|test|document|optimize|migrate> [target]"
allowed-tools: Read, Grep, Glob, Bash
model: sonnet
---

# Claude Code Pro Toolkit

Professional-grade agents for Claude Code that handle the work most developers skip: security audits, test generation, documentation, performance analysis, and migration assistance.

## Available Agents

### Security Auditor (`security-auditor`)
Scans your codebase for OWASP Top 10 vulnerabilities, hardcoded secrets, insecure dependencies, and common attack vectors. Produces a structured report with severity ratings and remediation steps.

**Invoke**: Dispatch via Task tool with `subagent_type: "security-auditor"`.

**Example prompt**: "Audit the src/ directory for security vulnerabilities. Focus on the authentication and API layers."

### Test Generator (`test-generator`)
Analyzes source files, learns your project's test conventions, and generates comprehensive test suites with unit tests, integration tests, and edge cases. Supports Jest, Vitest, pytest, Go testing, and more.

**Invoke**: Dispatch via Task tool with `subagent_type: "test-generator"`.

**Example prompt**: "Generate tests for src/auth/login.ts. Include edge cases for invalid credentials, expired tokens, and rate limiting."

### Code Documenter (`code-documenter`)
Generates JSDoc/TSDoc, Python docstrings, README files, and API documentation. Reads your existing documentation patterns first and matches your style. Won't over-document obvious code.

**Invoke**: Dispatch via Task tool with `subagent_type: "code-documenter"`.

**Example prompt**: "Document all exported functions in src/utils/ and generate a README for the project."

### Performance Optimizer (`performance-optimizer`)
Read-only analysis that finds N+1 queries, unnecessary re-renders, memory leaks, slow algorithms, missed concurrency opportunities, and hot-path bloat. Produces prioritized recommendations.

**Invoke**: Dispatch via Task tool with `subagent_type: "performance-optimizer"`.

**Example prompt**: "Analyze the API layer for performance issues. The /api/users endpoint is slow under load."

### Migration Assistant (`migration-assistant`)
Creates a migration plan and executes it step by step with verification at each stage. Supports framework upgrades, library swaps, and language ports.

**Invoke**: Dispatch via Task tool with `subagent_type: "migration-assistant"`.

**Example prompt**: "Migrate from React 18 to React 19. Check the migration guide and update all affected files."

## Usage Pattern

The recommended pattern is to dispatch agents via the Task tool:

```
Task tool:
  subagent_type: "security-auditor"
  description: "Security audit of auth module"
  prompt: "Perform a full security audit of the authentication system in src/auth/. Focus on JWT handling, password storage, and session management."
  mode: "bypassPermissions"
```

For automated workflows, chain agents:
1. Write code
2. Dispatch `test-generator` to create tests
3. Dispatch `security-auditor` to scan for vulnerabilities
4. Dispatch `code-documenter` to generate docs
5. Dispatch `performance-optimizer` to check for bottlenecks

## Quality Hooks

This plugin includes hooks that enforce quality gates:

- **PostToolUse (Edit|Write)**: Auto-runs the project linter after file modifications.
- **TaskCompleted**: Verifies that tests pass before allowing a task to be marked complete.

See `hooks/hooks.json` for configuration.
