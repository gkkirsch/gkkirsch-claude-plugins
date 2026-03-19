# /test-suite Command

The `/test-suite` command is your entry point for comprehensive testing automation. It analyzes your project, recommends a testing strategy, and helps you implement tests, configure CI pipelines, and achieve quality goals.

## Usage

```
/test-suite [subcommand] [options]
```

## Subcommands

### `/test-suite analyze`

Analyze the current project's testing state — existing tests, coverage gaps, framework configuration, and test pyramid balance.

**What it does:**
1. Scans for test files and frameworks (Jest, Vitest, Playwright, pytest, Go testing)
2. Evaluates test coverage if coverage data exists
3. Assesses the test pyramid: ratio of unit, integration, and E2E tests
4. Identifies untested critical paths (auth, payments, data integrity)
5. Reports on test infrastructure: factories, fixtures, helpers
6. Checks CI/CD configuration for test automation

**Output:** A testing health report with specific recommendations and priorities.

### `/test-suite strategy`

Design a comprehensive testing strategy for the project based on its tech stack, architecture, and current state.

**What it does:**
1. Analyzes the project's architecture and tech stack
2. Proposes a test pyramid with specific percentages and targets
3. Recommends testing frameworks and tools
4. Defines coverage thresholds per module/component
5. Outlines quality gates for CI/CD
6. Creates a phased implementation plan

**Output:** A testing strategy document with framework choices, coverage targets, and implementation priorities.

### `/test-suite implement`

Generate tests for specific modules, components, or features. Specify what you want tested.

**Examples:**
- `/test-suite implement auth` — Generate unit + integration tests for authentication
- `/test-suite implement api` — Generate API endpoint tests
- `/test-suite implement components` — Generate React component tests
- `/test-suite implement e2e-checkout` — Generate E2E checkout flow tests

**What it does:**
1. Reads the source code for the target module
2. Identifies testable behavior and edge cases
3. Generates test files with proper factories, mocks, and assertions
4. Follows existing test patterns and conventions in the project
5. Includes coverage for happy path, error cases, and boundary conditions

### `/test-suite ci`

Set up or optimize CI/CD testing pipelines.

**What it does:**
1. Analyzes existing CI configuration
2. Generates optimized GitHub Actions or GitLab CI workflows
3. Configures test sharding and parallelization
4. Sets up caching for dependencies and browser binaries
5. Implements quality gates and coverage reporting
6. Adds artifact collection for test reports and screenshots

### `/test-suite fix-flaky`

Diagnose and fix flaky tests in the project.

**What it does:**
1. Identifies known flaky test patterns (timing, shared state, race conditions)
2. Analyzes recent test failures for flakiness signals
3. Suggests specific fixes for each flaky test
4. Implements retry logic and quarantine strategies
5. Adds deterministic alternatives (fake timers, seeded data)

### `/test-suite coverage`

Analyze and improve test coverage.

**What it does:**
1. Runs tests with coverage collection
2. Identifies files with low or no coverage
3. Highlights critical paths that need more coverage
4. Suggests specific tests to write to increase coverage
5. Sets up coverage thresholds and CI enforcement

### `/test-suite tdd [feature]`

Start a TDD workflow for a new feature. Guides you through the Red-Green-Refactor cycle.

**Examples:**
- `/test-suite tdd shopping-cart`
- `/test-suite tdd user-authentication`
- `/test-suite tdd payment-processing`

**What it does:**
1. Discusses the feature requirements
2. Writes the first failing test (RED)
3. Implements minimal code to pass (GREEN)
4. Suggests refactoring opportunities (REFACTOR)
5. Repeats for each behavior/requirement

## Agent Selection

The command automatically routes to the appropriate specialist agent:

| Subcommand | Primary Agent | Supporting Agents |
|---|---|---|
| `analyze` | test-architect | — |
| `strategy` | test-architect | — |
| `implement` (unit) | unit-test-specialist | — |
| `implement` (e2e) | e2e-test-engineer | — |
| `ci` | ci-pipeline-builder | — |
| `fix-flaky` | test-architect | unit-test-specialist |
| `coverage` | test-architect | unit-test-specialist |
| `tdd` | test-architect | unit-test-specialist |

## Quick Start

For a project with no tests:
```
/test-suite analyze    # Understand current state
/test-suite strategy   # Design testing approach
/test-suite implement  # Generate initial tests
/test-suite ci         # Set up CI pipeline
```

For an existing project:
```
/test-suite analyze    # Find gaps
/test-suite coverage   # Identify priority areas
/test-suite fix-flaky  # Stabilize existing tests
```

## Notes

- All generated tests follow the project's existing patterns and conventions
- Tests are written to be maintainable: clear names, AAA pattern, no implementation testing
- Factory and fixture files are created or extended as needed
- CI configurations are optimized for fast feedback with proper caching
