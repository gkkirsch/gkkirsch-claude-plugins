---
name: cli-setup
description: >
  Scaffold a new CLI tool project with Commander.js, TypeScript, and proper
  project structure. Use when starting a new command-line tool from scratch.
  Triggers: "create CLI tool", "new CLI project", "scaffold CLI", "build a command-line tool".
  NOT for: existing CLI projects, web applications, or library packages.
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# CLI Tool Setup

Scaffold a production-ready CLI tool with Commander.js and TypeScript.

## Project Structure

```
my-cli/
  src/
    index.ts            # Entry point + top-level command registration
    commands/
      init.ts           # Each subcommand in its own file
      deploy.ts
      list.ts
    lib/
      config.ts         # Configuration loading (cosmiconfig)
      logger.ts         # Structured output (chalk + ora)
      errors.ts         # Custom error classes
    types.ts            # Shared type definitions
  bin/
    my-cli              # Executable entry (#!/usr/bin/env node)
  tests/
    commands/
      init.test.ts
  package.json
  tsconfig.json
  .npmignore
```

## Step 1: Initialize the Project

```bash
mkdir my-cli && cd my-cli
npm init -y
```

## Step 2: Install Dependencies

```bash
# Core
npm install commander chalk ora cosmiconfig

# Dev
npm install -D typescript @types/node tsx vitest
```

## Step 3: TypeScript Configuration

```json
// tsconfig.json
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "NodeNext",
    "moduleResolution": "NodeNext",
    "outDir": "dist",
    "rootDir": "src",
    "strict": true,
    "esModuleInterop": true,
    "declaration": true,
    "declarationMap": true,
    "sourceMap": true,
    "skipLibCheck": true
  },
  "include": ["src"],
  "exclude": ["node_modules", "dist", "tests"]
}
```

## Step 4: package.json Configuration

```json
{
  "name": "my-cli",
  "version": "0.1.0",
  "description": "My awesome CLI tool",
  "type": "module",
  "bin": {
    "my-cli": "./dist/index.js"
  },
  "files": [
    "dist"
  ],
  "scripts": {
    "build": "tsc",
    "dev": "tsx src/index.ts",
    "test": "vitest",
    "prepublishOnly": "npm run build"
  },
  "engines": {
    "node": ">=18"
  },
  "keywords": ["cli"],
  "license": "MIT"
}
```

## Step 5: Entry Point

```typescript
// src/index.ts
#!/usr/bin/env node
import { Command } from 'commander';
import { readFileSync } from 'fs';
import { fileURLToPath } from 'url';
import { dirname, join } from 'path';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);
const pkg = JSON.parse(
  readFileSync(join(__dirname, '..', 'package.json'), 'utf-8')
);

const program = new Command();

program
  .name('my-cli')
  .description('My awesome CLI tool')
  .version(pkg.version, '-v, --version');

// Register subcommands
// import { registerInitCommand } from './commands/init.js';
// registerInitCommand(program);

program.parse();
```

**Important**: Add the shebang in the built output. Either:
- Add `#!/usr/bin/env node` as the first line of `src/index.ts` (TypeScript will preserve it), OR
- Use a build script that prepends it to `dist/index.js`

Then make it executable:
```bash
chmod +x dist/index.js
```

## Step 6: Logger Utility

```typescript
// src/lib/logger.ts
import chalk from 'chalk';
import ora, { type Ora } from 'ora';

export const logger = {
  info: (msg: string) => console.error(chalk.blue('i'), msg),
  success: (msg: string) => console.error(chalk.green('✓'), msg),
  warn: (msg: string) => console.error(chalk.yellow('!'), msg),
  error: (msg: string) => console.error(chalk.red('✗'), msg),
  dim: (msg: string) => console.error(chalk.dim(msg)),

  // Data output goes to stdout (pipeable)
  data: (msg: string) => console.log(msg),
  json: (obj: unknown) => console.log(JSON.stringify(obj, null, 2)),

  spinner: (text: string): Ora => ora({ text, stream: process.stderr }),
};
```

**Key rule**: Status/progress to `stderr`, data to `stdout`. This lets users pipe output: `my-cli list | grep active`.

## Step 7: Config Loader

```typescript
// src/lib/config.ts
import { cosmiconfig } from 'cosmiconfig';

interface Config {
  apiUrl?: string;
  token?: string;
  defaultProject?: string;
}

const explorer = cosmiconfig('mycli');

export async function loadConfig(): Promise<Config> {
  // Priority: env vars > config file > defaults
  const result = await explorer.search();
  const fileConfig = result?.config ?? {};

  return {
    apiUrl: process.env.MYCLI_API_URL ?? fileConfig.apiUrl ?? 'https://api.example.com',
    token: process.env.MYCLI_TOKEN ?? fileConfig.token,
    defaultProject: fileConfig.defaultProject,
  };
}
```

This auto-discovers: `.myclirc`, `.myclirc.json`, `.myclirc.yaml`, `mycli.config.js`, `mycli.config.ts`, `"mycli"` in `package.json`.

## Step 8: Custom Error Classes

```typescript
// src/lib/errors.ts
import chalk from 'chalk';

export class CLIError extends Error {
  constructor(
    message: string,
    public readonly suggestions?: string[],
    public readonly context?: Record<string, string>
  ) {
    super(message);
    this.name = 'CLIError';
  }

  format(): string {
    const lines = [chalk.red(`ERROR: ${this.message}`)];

    if (this.context) {
      for (const [key, value] of Object.entries(this.context)) {
        lines.push(chalk.dim(`  → ${key}: ${value}`));
      }
    }

    if (this.suggestions?.length) {
      lines.push('');
      lines.push('Try:');
      this.suggestions.forEach((s, i) => {
        lines.push(`  ${i + 1}. ${s}`);
      });
    }

    return lines.join('\n');
  }
}

// Global error handler — put this in index.ts
export function setupErrorHandling(): void {
  process.on('uncaughtException', (err) => {
    if (err instanceof CLIError) {
      console.error(err.format());
    } else {
      console.error(chalk.red('Unexpected error:'), err.message);
      if (process.env.DEBUG) console.error(err.stack);
    }
    process.exit(1);
  });

  process.on('unhandledRejection', (reason) => {
    console.error(chalk.red('Unhandled rejection:'), reason);
    process.exit(1);
  });
}
```

## Step 9: Example Subcommand

```typescript
// src/commands/init.ts
import { Command } from 'commander';
import { logger } from '../lib/logger.js';
import { CLIError } from '../lib/errors.js';

export function registerInitCommand(program: Command): void {
  program
    .command('init')
    .description('Initialize a new project')
    .argument('[name]', 'project name', 'my-project')
    .option('-t, --template <template>', 'project template', 'default')
    .option('--no-git', 'skip git init')
    .option('--dry-run', 'show what would be created')
    .action(async (name: string, options) => {
      const spinner = logger.spinner(`Creating project ${name}...`);

      if (options.dryRun) {
        logger.info(`Would create project: ${name}`);
        logger.info(`Template: ${options.template}`);
        logger.info(`Git: ${options.git}`);
        return;
      }

      spinner.start();

      try {
        // Your init logic here
        await new Promise(resolve => setTimeout(resolve, 1000)); // placeholder

        spinner.succeed(`Project ${name} created`);
        logger.dim(`  cd ${name} && npm install`);
      } catch (err) {
        spinner.fail('Failed to create project');
        throw new CLIError(
          `Could not create project "${name}"`,
          [`Check if directory "${name}" already exists`, 'Try with a different name'],
          { template: options.template }
        );
      }
    });
}
```

## Step 10: Development and Testing

```bash
# Run during development (no build step)
npm run dev -- init my-app --template react

# Or link globally for testing
npm link
my-cli init my-app

# Run tests
npm test
```

## Step 11: Publish to npm

```bash
# Login to npm
npm login

# Publish (runs prepublishOnly → builds first)
npm publish

# Users install with:
npm install -g my-cli

# Or run without installing:
npx my-cli init my-app
```

## Gotchas

- **ESM vs CJS**: Use `"type": "module"` in package.json. Commander.js supports both. Import with `.js` extensions in TypeScript (`./commands/init.js`).
- **Shebang**: The `#!/usr/bin/env node` line MUST be the first line of the compiled entry point. Without it, `npx` and global installs won't work.
- **`files` in package.json**: Only include `dist/`. Don't publish `src/`, `tests/`, or config files.
- **`engines`**: Specify minimum Node.js version. Commander.js 12+ requires Node 18+.
- **Global vs local**: Design for both. `npx my-cli` should work as well as `npm i -g my-cli`.
- **`console.error` for status**: All non-data output goes to stderr. This is the #1 mistake new CLI authors make.
