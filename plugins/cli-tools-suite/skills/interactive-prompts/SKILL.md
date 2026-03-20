---
name: interactive-prompts
description: >
  Build interactive CLI prompts with Inquirer.js — text input, selections,
  confirmations, multi-select, password input, and dynamic prompt flows.
  Triggers: "CLI prompts", "interactive input", "CLI wizard", "user input in terminal".
  NOT for: argument parsing (use argument-parsing skill), output formatting.
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Interactive CLI Prompts with Inquirer.js

## Install

```bash
npm install @inquirer/prompts
```

Inquirer v2+ uses individual imports (tree-shakeable). Don't install the old `inquirer` package.

## Core Prompts

### Text Input

```typescript
import { input } from '@inquirer/prompts';

const name = await input({
  message: 'Project name:',
  default: 'my-project',
  validate: (value) => {
    if (!/^[a-z0-9-]+$/.test(value)) {
      return 'Only lowercase letters, numbers, and hyphens allowed';
    }
    if (value.length < 2) return 'Must be at least 2 characters';
    return true;
  },
  transformer: (value) => value.toLowerCase().replace(/\s+/g, '-'),
});
```

### Confirmation

```typescript
import { confirm } from '@inquirer/prompts';

const proceed = await confirm({
  message: 'This will delete 42 files. Continue?',
  default: false,  // ALWAYS default to safe option for destructive actions
});

if (!proceed) {
  console.log('Cancelled.');
  process.exit(0);
}
```

### Select (Single Choice)

```typescript
import { select } from '@inquirer/prompts';

const template = await select({
  message: 'Choose a template:',
  choices: [
    {
      name: 'React + TypeScript',
      value: 'react-ts',
      description: 'Vite-powered React with TypeScript and Tailwind',
    },
    {
      name: 'Next.js',
      value: 'nextjs',
      description: 'Full-stack React framework with App Router',
    },
    {
      name: 'Express API',
      value: 'express',
      description: 'REST API with TypeScript and Zod validation',
    },
    { name: 'Blank', value: 'blank' },
  ],
});
// Returns the value: 'react-ts', 'nextjs', etc.
```

### Checkbox (Multi-Select)

```typescript
import { checkbox } from '@inquirer/prompts';

const features = await checkbox({
  message: 'Select features to include:',
  choices: [
    { name: 'TypeScript', value: 'typescript', checked: true },
    { name: 'ESLint', value: 'eslint', checked: true },
    { name: 'Prettier', value: 'prettier' },
    { name: 'Vitest', value: 'vitest' },
    { name: 'Docker', value: 'docker' },
    { name: 'CI/CD (GitHub Actions)', value: 'ci' },
  ],
  validate: (choices) => {
    if (choices.length === 0) return 'Select at least one feature';
    return true;
  },
});
// Returns array: ['typescript', 'eslint', 'vitest']
```

### Password

```typescript
import { password } from '@inquirer/prompts';

const apiKey = await password({
  message: 'Enter your API key:',
  mask: '*',       // show asterisks (default: hide input completely)
  validate: (value) => {
    if (value.length < 10) return 'API key too short';
    return true;
  },
});
```

### Number

```typescript
import { number } from '@inquirer/prompts';

const port = await number({
  message: 'Port number:',
  default: 3000,
  min: 1,
  max: 65535,
});
```

### Editor (Opens $EDITOR)

```typescript
import { editor } from '@inquirer/prompts';

const description = await editor({
  message: 'Enter project description:',
  default: '# My Project\n\nDescribe your project here.',
  postfix: '.md',  // temp file extension (for editor syntax highlighting)
});
```

### Search/Autocomplete

```typescript
import { search } from '@inquirer/prompts';

const pkg = await search({
  message: 'Search npm packages:',
  source: async (term) => {
    if (!term) return [];
    const response = await fetch(
      `https://registry.npmjs.org/-/v1/search?text=${term}&size=10`
    );
    const data = await response.json();
    return data.objects.map((obj: any) => ({
      name: `${obj.package.name} (${obj.package.version})`,
      value: obj.package.name,
      description: obj.package.description,
    }));
  },
});
```

## Multi-Step Flows (Wizards)

### Sequential Prompts

```typescript
async function initWizard() {
  // Step 1: Basic info
  const name = await input({ message: 'Project name:', default: 'my-app' });
  const template = await select({
    message: 'Template:',
    choices: [
      { name: 'React', value: 'react' },
      { name: 'Express', value: 'express' },
      { name: 'Fullstack', value: 'fullstack' },
    ],
  });

  // Step 2: Features (depends on template)
  const featureChoices = getFeatureChoices(template); // template-specific options
  const features = await checkbox({
    message: 'Features:',
    choices: featureChoices,
  });

  // Step 3: Advanced (optional)
  const advanced = await confirm({
    message: 'Configure advanced options?',
    default: false,
  });

  let advancedOpts = {};
  if (advanced) {
    const port = await number({ message: 'Dev server port:', default: 3000 });
    const envFile = await confirm({ message: 'Generate .env file?', default: true });
    advancedOpts = { port, envFile };
  }

  // Step 4: Confirm
  console.log('\nProject configuration:');
  console.log(`  Name: ${name}`);
  console.log(`  Template: ${template}`);
  console.log(`  Features: ${features.join(', ')}`);
  if (advanced) console.log(`  Port: ${advancedOpts.port}`);

  const confirmed = await confirm({ message: 'Create project?', default: true });
  if (!confirmed) {
    console.log('Cancelled.');
    process.exit(0);
  }

  return { name, template, features, ...advancedOpts };
}
```

### Progressive Disclosure Pattern

Ask essential questions first. Only ask advanced questions if the user opts in:

```typescript
// Essential (everyone answers these)
const name = await input({ message: 'Project name:' });
const template = await select({ message: 'Template:', choices: [...] });

// Optional depth (most users skip)
const customize = await confirm({
  message: 'Customize advanced settings?',
  default: false
});

if (customize) {
  // Database, auth, deployment, CI/CD, etc.
}
```

## Non-Interactive Mode (CI/CD Compatibility)

Every interactive prompt should have a non-interactive fallback:

```typescript
import { confirm } from '@inquirer/prompts';

interface Options {
  force?: boolean;
  name?: string;
}

async function createProject(options: Options) {
  // If --force, skip all prompts
  const name = options.name ?? await input({ message: 'Project name:' });

  // Skip confirmation with --force
  if (!options.force) {
    const proceed = await confirm({ message: `Create "${name}"?` });
    if (!proceed) process.exit(0);
  }

  // Create the project...
}
```

**Rule**: Every `await input()` / `await select()` should have an equivalent CLI flag. Users must be able to run your tool non-interactively:

```bash
# Interactive:
my-cli init

# Non-interactive (CI/CD):
my-cli init --name my-app --template react --force
```

### Detect Non-Interactive Environment

```typescript
const isInteractive = process.stdin.isTTY && !process.env.CI;

if (!isInteractive && !options.force) {
  console.error('Error: Running in non-interactive mode.');
  console.error('Use --force or provide all options via flags.');
  process.exit(2);
}
```

## Cancellation Handling

Users can press Ctrl+C during any prompt. Handle it gracefully:

```typescript
import { input, ExitPromptError } from '@inquirer/prompts';

try {
  const name = await input({ message: 'Project name:' });
  const template = await select({ message: 'Template:', choices: [...] });
} catch (err) {
  if (err instanceof ExitPromptError) {
    // User pressed Ctrl+C or Escape
    console.log('\nCancelled.');
    process.exit(130);  // 128 + SIGINT (2) = 130
  }
  throw err;
}
```

Or wrap all prompts in a helper:

```typescript
async function prompt<T>(fn: () => Promise<T>): Promise<T> {
  try {
    return await fn();
  } catch (err) {
    if (err instanceof ExitPromptError) {
      console.log('\nCancelled.');
      process.exit(130);
    }
    throw err;
  }
}

// Usage:
const name = await prompt(() => input({ message: 'Name:' }));
```

## Themes and Styling

Customize prompt appearance:

```typescript
import { select, Separator } from '@inquirer/prompts';
import chalk from 'chalk';

const choice = await select({
  message: 'Environment:',
  choices: [
    { name: 'Development', value: 'dev' },
    { name: 'Staging', value: 'staging' },
    new Separator('--- Production ---'),
    { name: chalk.red('Production (use caution)'), value: 'production' },
  ],
  theme: {
    prefix: chalk.blue('?'),
    style: {
      highlight: (text: string) => chalk.cyan.bold(text),
      answer: (text: string) => chalk.green(text),
    },
  },
});
```

## Real-World Patterns

### Login Flow

```typescript
async function loginFlow() {
  const method = await select({
    message: 'How would you like to authenticate?',
    choices: [
      { name: 'Browser (OAuth)', value: 'browser' },
      { name: 'API Key', value: 'api-key' },
      { name: 'Username & Password', value: 'credentials' },
    ],
  });

  switch (method) {
    case 'browser':
      console.log('Opening browser...');
      // Start OAuth flow, open browser, wait for callback
      break;

    case 'api-key':
      const key = await password({ message: 'API Key:' });
      // Validate key against API
      break;

    case 'credentials':
      const username = await input({ message: 'Username:' });
      const pass = await password({ message: 'Password:' });
      // Authenticate
      break;
  }
}
```

### Destructive Action Confirmation

```typescript
async function confirmDeletion(resources: string[]) {
  console.log(chalk.yellow('\nThe following resources will be deleted:'));
  resources.forEach(r => console.log(chalk.dim(`  - ${r}`)));
  console.log();

  const typed = await input({
    message: `Type "${resources.length} resources" to confirm:`,
    validate: (value) => {
      if (value !== `${resources.length} resources`) {
        return 'Input does not match. Type exactly as shown.';
      }
      return true;
    },
  });

  return typed === `${resources.length} resources`;
}
```

### Smart Defaults from Context

```typescript
import { basename } from 'path';
import { existsSync, readFileSync } from 'fs';

function inferProjectName(): string {
  // Try package.json first
  if (existsSync('package.json')) {
    const pkg = JSON.parse(readFileSync('package.json', 'utf-8'));
    if (pkg.name) return pkg.name;
  }
  // Fall back to directory name
  return basename(process.cwd());
}

const name = await input({
  message: 'Project name:',
  default: inferProjectName(),
});
```

## Gotchas

- **@inquirer/prompts, not inquirer.** The new package uses individual imports. The old `inquirer` (v9 and below) uses a different API with `inquirer.prompt([...])` array syntax.
- **Always handle Ctrl+C.** Uncaught `ExitPromptError` prints an ugly stack trace. Wrap prompts in try/catch or a helper.
- **TTY required.** Prompts crash in non-TTY environments (piped input, CI). Always check `process.stdin.isTTY` and provide flag-based alternatives.
- **Don't nest prompts in loops without cleanup.** If you're doing `while (true) { await select(...) }`, make sure Ctrl+C breaks the loop cleanly.
- **validate returns true or a string.** Return `true` for valid, a string for the error message. Don't return `false` (shows no message).
- **Default values show in parentheses.** `default: 'my-app'` shows as `Project name: (my-app)`. Don't duplicate this in the message text.
- **Separator items are not selectable.** Use `new Separator()` to add visual dividers in select/checkbox lists.
