---
name: ts-mastery
description: >
  Quick TypeScript development command — analyzes your codebase and helps with advanced types, type safety,
  JS-to-TS migration, and compilation performance. Routes to the appropriate specialist agent based on your request.
  Triggers: "/ts-mastery", "typescript architect", "advanced types", "generics", "conditional types",
  "mapped types", "template literal types", "branded types", "type safety", "strict mode", "type guards",
  "discriminated unions", "zod validation", "runtime validation", "ts migration", "js to ts",
  "typescript migration", "allowJs", "checkJs", "declaration files", "d.ts",
  "typescript performance", "tsconfig", "barrel files", "project references", "compilation speed".
user-invocable: true
argument-hint: "<types|safety|migrate|perf> [target] [--audit]"
allowed-tools: Read, Write, Edit, Bash, Glob, Grep
model: sonnet
---

# /ts-mastery Command

One-command TypeScript expertise. Analyzes your project, identifies TypeScript patterns and issues, and routes to the appropriate specialist agent for type system architecture, type safety, migration, or performance optimization.

## Usage

```
/ts-mastery                            # Auto-detect and suggest improvements
/ts-mastery types                      # Advanced type system design
/ts-mastery types --generics           # Generic type patterns
/ts-mastery types --conditional        # Conditional and mapped types
/ts-mastery types --branded            # Branded/nominal types
/ts-mastery safety                     # Type safety audit
/ts-mastery safety --strict            # Strict mode setup
/ts-mastery safety --validation        # Runtime validation (Zod, io-ts)
/ts-mastery safety --errors            # Error handling patterns
/ts-mastery migrate                    # JS-to-TS migration
/ts-mastery migrate --plan             # Migration plan and strategy
/ts-mastery migrate --declarations     # Write declaration files
/ts-mastery perf                       # Build performance
/ts-mastery perf --tsconfig            # tsconfig optimization
/ts-mastery perf --barrels             # Barrel file analysis
/ts-mastery --audit                    # Full TypeScript audit
```

## Agent Routing

| Subcommand | Agent | Focus |
|------------|-------|-------|
| `types` | typescript-architect | Generics, conditional types, mapped types, template literals |
| `safety` | type-safety-engineer | Strict mode, type guards, validation, error handling |
| `migrate` | ts-migration-expert | JS→TS conversion, declaration files, incremental adoption |
| `perf` | ts-performance-optimizer | Build speed, tsconfig, barrel files, IDE performance |
