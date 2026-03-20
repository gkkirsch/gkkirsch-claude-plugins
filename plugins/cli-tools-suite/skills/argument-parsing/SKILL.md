---
name: argument-parsing
description: >
  Design and implement CLI argument parsing with Commander.js — commands, options,
  arguments, validation, and help text. Use for adding commands or restructuring
  CLI argument handling.
  Triggers: "add CLI command", "parse arguments", "add subcommand", "CLI options".
  NOT for: interactive prompts (use interactive-prompts skill), output formatting.
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# CLI Argument Parsing with Commander.js

## Command Anatomy

```
my-cli deploy staging --force --timeout 30 -- extra args
│       │      │        │       │            │
│       │      │        │       │            └─ Passthrough args (after --)
│       │      │        │       └─ Option with value
│       │      │        └─ Boolean option (flag)
│       │      └─ Argument
│       └─ Command (subcommand)
└─ Program name
```

## Arguments

Arguments are positional values. Use angle brackets for required, square brackets for optional.

```typescript
program
  .command('deploy')
  .argument('<environment>', 'target environment (staging, production)')
  .argument('[tag]', 'version tag to deploy', 'latest')  // optional with default
  .action((environment: string, tag: string) => {
    // environment is required — Commander errors if missing
    // tag defaults to 'latest' if not provided
  });
```

### Variadic Arguments

Use `...` for variable-length argument lists. Must be last.

```typescript
program
  .command('install')
  .argument('<packages...>', 'packages to install')
  .action((packages: string[]) => {
    // packages is an array: ['express', 'chalk', 'ora']
  });
```

### Argument Processing

Transform or validate arguments inline:

```typescript
function parsePort(value: string): number {
  const port = parseInt(value, 10);
  if (isNaN(port) || port < 1 || port > 65535) {
    throw new Error(`Invalid port: ${value}. Must be 1-65535.`);
  }
  return port;
}

program
  .command('serve')
  .argument('<port>', 'port number', parsePort)
  .action((port: number) => {
    // port is already validated and parsed to number
  });
```

## Options

### Boolean Flags

```typescript
program
  .option('-f, --force', 'skip confirmation prompts')
  .option('-v, --verbose', 'enable verbose output')
  .option('--no-color', 'disable colored output');  // negation flag

// Usage: my-cli --force --verbose --no-color
// opts.force === true, opts.verbose === true, opts.color === false
```

### Options with Values

```typescript
program
  .option('-o, --output <path>', 'output file path')                    // required value
  .option('-t, --timeout <ms>', 'timeout in milliseconds', '5000')      // with default
  .option('-e, --env <vars...>', 'environment variables')               // variadic
  .requiredOption('-k, --api-key <key>', 'API key (required)');         // mandatory option

// Usage: my-cli -o ./dist -t 10000 -e FOO=bar BAZ=qux -k abc123
// opts.output === './dist'
// opts.timeout === '10000' (string! use parseFloat to convert)
// opts.env === ['FOO=bar', 'BAZ=qux']
// opts.apiKey === 'abc123'
```

### Option Processing (Validation + Transformation)

```typescript
import { InvalidArgumentError } from 'commander';

function parseTimeout(value: string): number {
  const ms = parseInt(value, 10);
  if (isNaN(ms) || ms < 0) {
    throw new InvalidArgumentError('Must be a positive number.');
  }
  return ms;
}

function collectTags(value: string, previous: string[]): string[] {
  return [...previous, value];
}

program
  .option('-t, --timeout <ms>', 'timeout', parseTimeout, 5000)
  .option('--tag <tag>', 'add tag (repeatable)', collectTags, []);

// Usage: my-cli --timeout 3000 --tag v1 --tag stable
// opts.timeout === 3000 (number, validated)
// opts.tag === ['v1', 'stable'] (accumulated)
```

### Option Choices

```typescript
program
  .addOption(
    new Option('-l, --log-level <level>', 'logging level')
      .choices(['debug', 'info', 'warn', 'error'])
      .default('info')
  );

// Usage: my-cli --log-level debug    ← works
// Usage: my-cli --log-level trace    ← Commander shows error + valid choices
```

### Environment Variable Fallback

```typescript
program
  .addOption(
    new Option('-k, --api-key <key>', 'API key')
      .env('MY_CLI_API_KEY')  // reads from env if not on command line
  );
```

### Hidden and Preset Options

```typescript
program
  .addOption(
    new Option('--internal-debug')
      .hideHelp()                    // hidden from --help
  )
  .addOption(
    new Option('-p, --profile [name]', 'AWS profile')
      .preset('default')            // value when flag used without value
  );

// my-cli -p         → opts.profile === 'default'
// my-cli -p staging → opts.profile === 'staging'
// my-cli            → opts.profile === undefined
```

## Command Structure

### Nested Subcommands

```typescript
// my-cli config get <key>
// my-cli config set <key> <value>
// my-cli config list

const configCmd = program
  .command('config')
  .description('Manage configuration');

configCmd
  .command('get')
  .argument('<key>', 'config key')
  .action((key: string) => { /* ... */ });

configCmd
  .command('set')
  .argument('<key>', 'config key')
  .argument('<value>', 'config value')
  .action((key: string, value: string) => { /* ... */ });

configCmd
  .command('list')
  .option('--json', 'output as JSON')
  .action((options) => { /* ... */ });
```

### Command Aliases

```typescript
program
  .command('install')
  .alias('i')          // single alias
  .aliases(['add'])    // multiple aliases
  .description('Install packages')
  .action(() => { /* ... */ });

// All of these work: my-cli install, my-cli i, my-cli add
```

### Default Command

```typescript
program
  .command('serve', { isDefault: true })
  .description('Start the server (default command)')
  .action(() => { /* ... */ });

// my-cli         → runs serve
// my-cli serve   → also runs serve
```

### Command from Separate Files

```typescript
// src/commands/deploy.ts
import { Command } from 'commander';

export function makeDeployCommand(): Command {
  const cmd = new Command('deploy');

  cmd
    .description('Deploy to environment')
    .argument('<env>', 'target environment')
    .option('-f, --force', 'skip confirmations')
    .action(async (env, options) => {
      // deploy logic
    });

  return cmd;
}

// src/index.ts
import { makeDeployCommand } from './commands/deploy.js';
program.addCommand(makeDeployCommand());
```

## Help Text Customization

### Section Headers and Examples

```typescript
program
  .name('my-cli')
  .description('Deploy and manage your applications')
  .addHelpText('after', `

Examples:
  $ my-cli deploy staging
  $ my-cli deploy production --force
  $ my-cli config set api-url https://api.example.com
  $ my-cli logs --follow --tail 100

Environment Variables:
  MY_CLI_API_KEY     API authentication key
  MY_CLI_CONFIG      Path to config file
  NO_COLOR           Disable colored output
`);
```

### Per-Command Help

```typescript
program
  .command('deploy')
  .argument('<env>')
  .addHelpText('after', `

Examples:
  $ my-cli deploy staging              Deploy latest to staging
  $ my-cli deploy production --tag v2  Deploy specific version
  $ my-cli deploy staging --dry-run    Preview without deploying
`);
```

### Custom Help Sorting

```typescript
program.configureHelp({
  sortSubcommands: true,    // alphabetical subcommands
  sortOptions: true,        // alphabetical options
  showGlobalOptions: true,  // show global options on subcommands
});
```

## Validation Patterns

### Cross-Option Validation

Commander validates individual options, but cross-option logic goes in the action:

```typescript
program
  .command('deploy')
  .option('--canary', 'canary deployment')
  .option('--rollback', 'rollback to previous version')
  .action((options) => {
    if (options.canary && options.rollback) {
      console.error('Error: --canary and --rollback are mutually exclusive');
      process.exit(2);
    }
  });
```

### Argument Validation with Custom Errors

```typescript
import { InvalidArgumentError } from 'commander';

function parseEnvironment(value: string): string {
  const valid = ['staging', 'production', 'development'];
  if (!valid.includes(value)) {
    throw new InvalidArgumentError(
      `Must be one of: ${valid.join(', ')}`
    );
  }
  return value;
}
```

### Pre-Action Hooks

Run validation or setup before any command:

```typescript
program.hook('preAction', async (thisCommand, actionCommand) => {
  // Load config before any command runs
  const config = await loadConfig();
  actionCommand.setOptionValue('_config', config);

  // Require auth for certain commands
  if (['deploy', 'delete'].includes(actionCommand.name())) {
    if (!config.token) {
      console.error('Error: Authentication required. Run: my-cli login');
      process.exit(1);
    }
  }
});
```

## Global Options

Options defined on the root program apply to all subcommands:

```typescript
program
  .option('--json', 'output as JSON')
  .option('-q, --quiet', 'suppress non-essential output')
  .option('-d, --debug', 'enable debug output');

// Access in any subcommand:
program
  .command('list')
  .action((localOpts, cmd) => {
    const globalOpts = cmd.optsWithGlobals();
    if (globalOpts.json) {
      logger.json(results);
    } else {
      logger.data(formatTable(results));
    }
  });
```

## Common Patterns

### Version with Build Info

```typescript
import { readFileSync } from 'fs';

const pkg = JSON.parse(readFileSync('./package.json', 'utf-8'));
const buildInfo = process.env.BUILD_SHA?.slice(0, 7) ?? 'dev';

program.version(`${pkg.version} (${buildInfo})`, '-v, --version');
// Output: 1.2.3 (abc1234)
```

### Passthrough Arguments

```typescript
program
  .command('run')
  .argument('<script>')
  .allowUnknownOption()          // don't error on unknown flags
  .passThroughOptions()          // stop parsing after first unknown
  .action((script, options, cmd) => {
    const extraArgs = cmd.args.slice(1); // everything after script name
    // Forward extraArgs to the child process
  });

// my-cli run test -- --coverage --watch
```

### Conflicting Options

```typescript
import { Option } from 'commander';

program
  .addOption(new Option('--json', 'JSON output').conflicts('table'))
  .addOption(new Option('--table', 'table output').conflicts('json'));
```

## Gotchas

- **Option values are strings by default.** `--timeout 5000` gives you `"5000"`. Always use a processing function to convert to number.
- **`--no-` prefix creates negation.** `--no-color` sets `opts.color = false`. This is automatic. Don't also define `--color` or you'll get conflicts.
- **Variadic must be last.** `<files...> <output>` won't work. Put variadic arguments at the end.
- **Exit codes matter.** Use `process.exit(2)` for usage errors (wrong args), `process.exit(1)` for runtime errors. Commander uses exit code 1 for its own errors.
- **`.parse()` must be last.** Any commands or options added after `.parse()` are ignored.
- **`opts` vs `optsWithGlobals()`**: `opts()` only returns the command's own options. `optsWithGlobals()` includes parent/global options too.
- **Commander v12+ requires Node 18+.** If you need Node 16 support, use Commander v11.
