---
name: typescript-mastery-suite
description: >
  TypeScript Mastery Suite — complete development toolkit for advanced TypeScript engineering.
  Advanced type system architecture with generics, conditional types, mapped types, and template literal types.
  Type safety engineering with strict mode, type guards, runtime validation (Zod, io-ts, Effect Schema),
  and typed error handling. JavaScript-to-TypeScript migration with incremental strategies, declaration files,
  and strict mode progression. TypeScript compilation performance with tsconfig optimization, project references,
  barrel file analysis, and IDE tuning.
  Triggers: "typescript architect", "advanced types", "generics", "conditional types", "mapped types",
  "template literal types", "branded types", "nominal types", "type inference", "satisfies",
  "type safety", "strict mode", "strict typescript", "type guards", "discriminated unions",
  "runtime validation", "zod schema", "io-ts", "effect schema", "typed errors", "result type",
  "trpc", "type-safe api", "prisma types", "drizzle types", "kysely",
  "typescript migration", "js to ts", "js to typescript", "allowjs", "checkjs",
  "declaration files", "d.ts", "ambient modules", "module augmentation", "definitelytyped",
  "typescript performance", "compilation speed", "tsconfig", "project references",
  "barrel files", "index.ts re-export", "isolatedmodules", "skipLibCheck",
  "type instantiation", "recursive types", "tree shaking typescript".
  Dispatches the appropriate specialist agent: typescript-architect, type-safety-engineer,
  ts-migration-expert, or ts-performance-optimizer.
  NOT for: Runtime JavaScript performance (use a profiler), React/framework-specific patterns
  (use framework-specific plugins), general Node.js development, or CSS/HTML.
version: 1.0.0
argument-hint: "<types|safety|migrate|perf> [target]"
user-invocable: true
allowed-tools: Read, Grep, Glob, Bash
model: sonnet
---

# TypeScript Mastery Suite

Production-grade TypeScript engineering agents for Claude Code. Four specialist agents that handle advanced type system design, type safety, JS-to-TS migration, and compilation performance — the complete TypeScript mastery lifecycle.

## Available Agents

### TypeScript Architect (`typescript-architect`)
Designs advanced type systems for large-scale codebases. Generics (constrained, inferred, defaulted, HKT patterns), conditional types (infer, distributive, recursive), mapped types (key remapping, homomorphic), template literal types (URL parsing, event patterns, string manipulation), type inference (satisfies, as const, const type params), branded/nominal/opaque types, and module system mastery (ESM/CJS interop, moduleResolution bundler).

**Invoke**: Dispatch via Task tool with `subagent_type: "typescript-architect"`.

**Example prompts**:
- "Design a type-safe event system with template literals"
- "Create branded types for my domain IDs"
- "Build a type-safe query builder with generics"
- "Implement HKT patterns for a functional library"

### Type Safety Engineer (`type-safety-engineer`)
Ensures maximum type safety across the codebase. Strict mode mastery (every flag explained), type guards (user-defined, assertion functions, narrowing), discriminated unions (exhaustive checks, state machines), runtime validation (Zod, io-ts, Effect Schema, arktype), error handling (Result types, typed errors), API type safety (tRPC, typed fetch), database type safety (Prisma, Drizzle, Kysely).

**Invoke**: Dispatch via Task tool with `subagent_type: "type-safety-engineer"`.

**Example prompts**:
- "Audit my codebase for type safety gaps"
- "Add Zod validation to my API endpoints"
- "Design a Result type error handling pattern"
- "Set up end-to-end type safety with tRPC"

### TS Migration Expert (`ts-migration-expert`)
Manages JavaScript-to-TypeScript migration. Migration strategies (big bang, incremental allowJs, checkJs gradual), declaration files (.d.ts authoring, ambient modules, module augmentation), strict mode progression (phased noImplicitAny → strictNullChecks → strict), common patterns (Express, React class→functional), monorepo migration (project references, composite), testing during migration, and tooling (ts-migrate, TypeStat, codemods).

**Invoke**: Dispatch via Task tool with `subagent_type: "ts-migration-expert"`.

**Example prompts**:
- "Create a migration plan for my JS project"
- "Write declaration files for this untyped library"
- "Enable strictNullChecks incrementally"
- "Set up project references for our monorepo"

### TS Performance Optimizer (`ts-performance-optimizer`)
Optimizes TypeScript build and IDE performance. Compilation speed (incremental, project references, isolatedModules), tsconfig optimization (skipLibCheck, target/lib, paths), barrel file anti-pattern (detection, measurement, fixes), IDE performance (language service tuning, VS Code settings), bundle analysis (tree-shaking, dead code), type computation performance (recursive type limits, instantiation depth).

**Invoke**: Dispatch via Task tool with `subagent_type: "ts-performance-optimizer"`.

**Example prompts**:
- "Profile and optimize my TypeScript build"
- "Find and fix barrel file performance issues"
- "Set up fast builds with esbuild + tsc type checking"
- "Fix slow IDE autocompletion"

## Quick Start: /ts-mastery

Use the `/ts-mastery` command for guided TypeScript development:

```
/ts-mastery                          # Auto-detect and suggest improvements
/ts-mastery types                    # Advanced type system design
/ts-mastery safety                   # Type safety audit
/ts-mastery migrate                  # JS-to-TS migration
/ts-mastery perf                     # Build performance optimization
/ts-mastery --audit                  # Full TypeScript audit
```

## Agent Selection Guide

| Need | Agent | Trigger |
|------|-------|---------|
| Generic type design | typescript-architect | "Design generics for..." |
| Conditional/mapped types | typescript-architect | "Create utility type..." |
| Template literal types | typescript-architect | "Type-safe string patterns" |
| Branded/opaque types | typescript-architect | "Prevent ID confusion" |
| Strict mode setup | type-safety-engineer | "Enable strict TypeScript" |
| Runtime validation | type-safety-engineer | "Add Zod schemas" |
| Error handling | type-safety-engineer | "Design Result types" |
| API type safety | type-safety-engineer | "End-to-end typed API" |
| JS→TS migration | ts-migration-expert | "Convert JS to TS" |
| Declaration files | ts-migration-expert | "Write .d.ts for..." |
| Monorepo types | ts-migration-expert | "Project references" |
| Build speed | ts-performance-optimizer | "Speed up tsc" |
| tsconfig tuning | ts-performance-optimizer | "Optimize tsconfig" |
| Barrel file issues | ts-performance-optimizer | "Fix barrel imports" |
| IDE performance | ts-performance-optimizer | "Slow autocomplete" |

## Reference Materials

This skill includes comprehensive reference documents in `references/`:

- **advanced-types.md** — Utility types, recursive types, variadic tuples, const assertions, satisfies operator, branded types, conditional/mapped type cheat sheets
- **tsconfig-guide.md** — Every tsconfig option explained, recommended configs for Node/React/Next.js/Library/Monorepo
- **declaration-files.md** — .d.ts authoring, ambient modules, wildcard declarations, module augmentation, global types, DefinitelyTyped

Agents automatically consult these references when working. You can also read them directly for quick answers.

## How It Works

1. You describe your TypeScript challenge (e.g., "design type-safe API routes")
2. The SKILL.md routes to the appropriate agent
3. The agent reads your code, discovers your tsconfig and patterns
4. Solutions are designed and implemented following TypeScript best practices
5. The agent provides results and next steps

All generated artifacts follow these principles:
- **TypeScript 5.x**: Latest features (satisfies, const type params, NoInfer, decorators)
- **Strict mode**: Full strict + noUncheckedIndexedAccess
- **Type inference**: Leverage inference, annotate at boundaries
- **Performance**: Avoid deep recursion, minimize instantiations, no barrel files
- **Validation**: Zod/io-ts at boundaries, type system internally
