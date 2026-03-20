---
name: output-formatting
description: >
  Format CLI output — colored text, tables, progress bars, spinners, boxed
  messages, and JSON output. Use when building the display layer of a CLI tool.
  Triggers: "CLI output", "terminal colors", "progress bar", "spinner", "CLI table",
  "format terminal output".
  NOT for: argument parsing, interactive prompts.
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# CLI Output Formatting

## Install Core Libraries

```bash
npm install chalk ora cli-table3 boxen figures log-symbols cli-progress listr2
```

## Colored Text with chalk

```typescript
import chalk from 'chalk';

// Basic colors
console.log(chalk.red('Error'));
console.log(chalk.green('Success'));
console.log(chalk.yellow('Warning'));
console.log(chalk.blue('Info'));
console.log(chalk.dim('Secondary text'));

// Modifiers
console.log(chalk.bold('Important'));
console.log(chalk.underline('https://example.com'));
console.log(chalk.italic('Note'));
console.log(chalk.strikethrough('Deprecated'));

// Combine styles
console.log(chalk.red.bold('CRITICAL ERROR'));
console.log(chalk.green.underline('Success: view at https://example.com'));
console.log(chalk.bgRed.white.bold(' FAIL '));

// Template literals
console.log(`
  ${chalk.bold('Project:')} my-app
  ${chalk.bold('Status:')}  ${chalk.green('Active')}
  ${chalk.bold('URL:')}     ${chalk.underline('https://my-app.example.com')}
`);

// Hex and RGB colors
console.log(chalk.hex('#FF6B6B')('Custom color'));
console.log(chalk.rgb(255, 136, 0)('Orange'));
```

### Respecting NO_COLOR

chalk automatically respects the `NO_COLOR` environment variable and `--no-color` flag. No extra code needed. But if you're building a custom logger:

```typescript
const useColor = !process.env.NO_COLOR && process.argv.indexOf('--no-color') === -1;

function colorize(text: string, color: typeof chalk.red): string {
  return useColor ? color(text) : text;
}
```

## Spinners with ora

```typescript
import ora from 'ora';

// Basic spinner
const spinner = ora('Loading configuration...').start();

try {
  const config = await loadConfig();
  spinner.succeed('Configuration loaded');
} catch (err) {
  spinner.fail('Failed to load configuration');
  throw err;
}

// Spinner with updates
const deploy = ora('Deploying...').start();
deploy.text = 'Deploying... building assets';
await buildAssets();
deploy.text = 'Deploying... uploading';
await upload();
deploy.text = 'Deploying... verifying';
await verify();
deploy.succeed('Deployed successfully');

// Spinner variants
spinner.info('Skipped (already up to date)');  // ℹ blue
spinner.warn('Deployed with warnings');         // ⚠ yellow
spinner.succeed('Done');                        // ✓ green
spinner.fail('Failed');                         // ✗ red

// Customize spinner style
const fancy = ora({
  text: 'Processing...',
  spinner: 'dots12',    // many built-in styles: dots, line, star, etc.
  color: 'cyan',
  stream: process.stderr,  // ALWAYS use stderr for spinners
});
```

## Progress Bars with cli-progress

```typescript
import { SingleBar, MultiBar, Presets } from 'cli-progress';
import chalk from 'chalk';

// Single progress bar
const bar = new SingleBar({
  format: `Uploading  ${chalk.cyan('{bar}')} {percentage}% | {value}/{total} files | ETA: {eta}s`,
  barCompleteChar: '\u2588',    // █
  barIncompleteChar: '\u2591',  // ░
  hideCursor: true,
  stream: process.stderr,
}, Presets.shades_classic);

bar.start(100, 0);
for (let i = 0; i <= 100; i++) {
  bar.update(i);
  await sleep(50);
}
bar.stop();

// Multiple concurrent progress bars
const multi = new MultiBar({
  clearOnComplete: false,
  hideCursor: true,
  format: '{name} | {bar} | {percentage}% | {value}/{total}',
  stream: process.stderr,
}, Presets.shades_grey);

const styles = multi.create(150, 0, { name: 'Styles  ' });
const scripts = multi.create(200, 0, { name: 'Scripts ' });
const images = multi.create(80, 0, { name: 'Images  ' });

// Update each bar independently
styles.update(50);
scripts.update(100);
images.update(30);

// When all done
multi.stop();
```

## Task Lists with listr2

```typescript
import { Listr } from 'listr2';

const tasks = new Listr([
  {
    title: 'Install dependencies',
    task: async (ctx, task) => {
      await exec('npm install');
    },
  },
  {
    title: 'Build project',
    task: async (ctx, task) => {
      task.output = 'Compiling TypeScript...';
      await exec('tsc');
      task.output = 'Bundling...';
      await exec('npm run bundle');
    },
    rendererOptions: { persistentOutput: true },
  },
  {
    title: 'Run tests',
    task: async (ctx, task) => {
      await exec('npm test');
    },
  },
  {
    title: 'Deploy',
    task: async (ctx, task) => {
      // Conditional skip
      if (!ctx.shouldDeploy) {
        task.skip('Skipping deploy (--no-deploy)');
        return;
      }
      await deploy();
    },
  },
], {
  concurrent: false,      // run sequentially
  exitOnError: true,      // stop on first failure
  rendererOptions: {
    collapseSubtasks: false,
  },
});

await tasks.run({ shouldDeploy: true });
```

### Nested Tasks

```typescript
const tasks = new Listr([
  {
    title: 'Setup',
    task: (ctx, task) => task.newListr([
      { title: 'Check Node version', task: async () => { /* ... */ } },
      { title: 'Check disk space', task: async () => { /* ... */ } },
      { title: 'Check network', task: async () => { /* ... */ } },
    ], { concurrent: true }), // Run subtasks in parallel
  },
  {
    title: 'Build',
    task: (ctx, task) => task.newListr([
      { title: 'Compile TypeScript', task: async () => { /* ... */ } },
      { title: 'Bundle assets', task: async () => { /* ... */ } },
    ]),
  },
]);
```

## Tables with cli-table3

```typescript
import Table from 'cli-table3';
import chalk from 'chalk';

// Simple table
const table = new Table({
  head: [
    chalk.bold('ID'),
    chalk.bold('Name'),
    chalk.bold('Status'),
    chalk.bold('Created'),
  ],
  style: {
    head: [],        // disable default head colors
    border: [],      // disable default border colors
  },
  colWidths: [12, 25, 12, 15],
});

table.push(
  ['abc-123', 'Production API', chalk.green('Active'), '2 days ago'],
  ['def-456', 'Staging API', chalk.yellow('Paused'), '1 week ago'],
  ['ghi-789', 'Dev API', chalk.red('Error'), '3 hours ago'],
);

console.log(table.toString());
```

### Compact Table (No Borders)

```typescript
const compact = new Table({
  chars: {
    top: '', 'top-mid': '', 'top-left': '', 'top-right': '',
    bottom: '', 'bottom-mid': '', 'bottom-left': '', 'bottom-right': '',
    left: '', 'left-mid': '', mid: '', 'mid-mid': '',
    right: '', 'right-mid': '', middle: '  ',
  },
  style: { 'padding-left': 0, 'padding-right': 1 },
});

compact.push(
  [chalk.dim('ID'), chalk.dim('NAME'), chalk.dim('STATUS')],
  ['abc-123', 'Production API', chalk.green('● Active')],
  ['def-456', 'Staging API', chalk.yellow('○ Paused')],
);

console.log(compact.toString());
```

Output:
```
ID       NAME            STATUS
abc-123  Production API  ● Active
def-456  Staging API     ○ Paused
```

## Boxed Messages with boxen

```typescript
import boxen from 'boxen';
import chalk from 'chalk';

// Success message
console.log(boxen(
  `${chalk.green.bold('Deploy successful!')}\n\n` +
  `URL: ${chalk.underline('https://my-app.example.com')}\n` +
  `Version: ${chalk.cyan('1.2.3')}`,
  {
    padding: 1,
    margin: 1,
    borderStyle: 'round',
    borderColor: 'green',
    title: 'Deployment',
    titleAlignment: 'center',
  }
));

// Update notification
console.error(boxen(
  `Update available: ${chalk.dim('1.0.0')} → ${chalk.green('2.0.0')}\n` +
  `Run ${chalk.cyan('npm i -g my-cli')} to update`,
  {
    padding: 1,
    borderStyle: 'round',
    borderColor: 'yellow',
    title: 'Update',
    titleAlignment: 'center',
    dimBorder: true,
  }
));
```

## Log Symbols

```typescript
import logSymbols from 'log-symbols';

console.log(logSymbols.success, 'Compiled successfully');  // ✓
console.log(logSymbols.error, 'Build failed');             // ✗
console.log(logSymbols.warning, 'Deprecated API used');    // ⚠
console.log(logSymbols.info, 'Watching for changes...');   // ℹ
```

## Building a CLI Logger

Combine everything into a structured logger:

```typescript
import chalk from 'chalk';
import ora, { type Ora } from 'ora';
import logSymbols from 'log-symbols';

interface Logger {
  info: (msg: string) => void;
  success: (msg: string) => void;
  warn: (msg: string) => void;
  error: (msg: string) => void;
  dim: (msg: string) => void;
  data: (msg: string) => void;
  json: (obj: unknown) => void;
  table: (rows: string[][]) => void;
  spinner: (text: string) => Ora;
  header: (text: string) => void;
  divider: () => void;
  newline: () => void;
}

export function createLogger(options: { quiet?: boolean; json?: boolean } = {}): Logger {
  const noop = () => {};

  return {
    // Status messages → stderr (not pipeable)
    info: options.quiet ? noop : (msg) => console.error(logSymbols.info, msg),
    success: options.quiet ? noop : (msg) => console.error(logSymbols.success, msg),
    warn: (msg) => console.error(logSymbols.warning, chalk.yellow(msg)),
    error: (msg) => console.error(logSymbols.error, chalk.red(msg)),
    dim: options.quiet ? noop : (msg) => console.error(chalk.dim(msg)),

    // Data output → stdout (pipeable)
    data: (msg) => console.log(msg),
    json: (obj) => console.log(JSON.stringify(obj, null, 2)),

    // Structured output
    table: (rows) => {
      if (options.json) {
        console.log(JSON.stringify(rows));
        return;
      }
      // Simple aligned table
      const widths = rows[0].map((_, i) =>
        Math.max(...rows.map(r => (r[i] || '').length))
      );
      rows.forEach(row => {
        console.log(row.map((cell, i) => cell.padEnd(widths[i])).join('  '));
      });
    },

    // Visual elements
    spinner: (text) => ora({ text, stream: process.stderr }),
    header: (text) => console.error(`\n${chalk.bold(text)}\n${'─'.repeat(text.length)}`),
    divider: () => console.error(chalk.dim('─'.repeat(40))),
    newline: () => console.error(''),
  };
}
```

Usage:

```typescript
const log = createLogger({ quiet: opts.quiet, json: opts.json });

log.header('Deploying Application');

const spinner = log.spinner('Building...').start();
await build();
spinner.succeed('Built in 3.2s');

log.table([
  ['SERVICE', 'STATUS', 'URL'],
  ['api', chalk.green('Running'), 'https://api.example.com'],
  ['web', chalk.green('Running'), 'https://example.com'],
  ['worker', chalk.yellow('Starting'), '—'],
]);

log.newline();
log.success('Deployment complete');
log.dim('View logs: my-cli logs --follow');
```

## stdout vs stderr Rules

This is the most important concept in CLI output. Get it right and your tool composes with pipes, redirects, and scripts.

| Output Type | Stream | Why |
|-------------|--------|-----|
| **Data** (JSON, lists, IDs, URLs) | `stdout` | Must be pipeable: `my-cli list \| grep active` |
| **Status** (spinners, progress) | `stderr` | Don't pollute piped output |
| **Errors** | `stderr` | Standard Unix convention |
| **Warnings** | `stderr` | Don't pollute piped output |
| **Tables** (when data) | `stdout` | The table IS the data |
| **Tables** (when status) | `stderr` | Summary tables are informational |
| **Prompts** | `stderr` | Don't interfere with piped output |

```bash
# This should work cleanly:
my-cli list --json > output.json        # only data in file
my-cli list | jq '.[] | .name'          # parseable
my-cli deploy 2>deploy.log              # errors captured separately
my-cli list 2>/dev/null                 # suppress status, keep data
```

## Dual Output Mode (Human + Machine)

Every command should support `--json` for machine-readable output:

```typescript
function displayResults(results: Result[], options: { json?: boolean }) {
  if (options.json) {
    // Machine: clean JSON to stdout
    console.log(JSON.stringify(results, null, 2));
    return;
  }

  // Human: formatted table with colors to stdout, status to stderr
  console.error(chalk.dim(`Found ${results.length} results\n`));

  const table = new Table({
    head: ['Name', 'Status', 'Updated'],
  });

  results.forEach(r => {
    table.push([
      r.name,
      r.active ? chalk.green('Active') : chalk.dim('Inactive'),
      timeAgo(r.updatedAt),
    ]);
  });

  console.log(table.toString());
}
```

## Terminal Width Handling

```typescript
// Get terminal width (fallback to 80)
const termWidth = process.stdout.columns || 80;

// Truncate long strings
function truncate(str: string, maxLen: number): string {
  if (str.length <= maxLen) return str;
  return str.slice(0, maxLen - 1) + '\u2026'; // ... character
}

// Responsive table columns
function getColumnWidths(termWidth: number) {
  if (termWidth >= 120) return { id: 12, name: 40, status: 12, url: 50 };
  if (termWidth >= 80)  return { id: 10, name: 25, status: 10, url: 30 };
  return { id: 8, name: 20, status: 8, url: 0 }; // hide URL on narrow terminals
}
```

## Gotchas

- **Spinners and progress bars go to stderr.** Pass `stream: process.stderr` to ora and cli-progress. Default is stdout which breaks piping.
- **chalk auto-detects color support.** Don't manually check — chalk disables colors for non-TTY, piped output, and `NO_COLOR` automatically.
- **cli-table3, not cli-table.** The `cli-table` package is unmaintained. Use `cli-table3` (drop-in compatible, actively maintained).
- **Unicode symbols vary by terminal.** Use the `figures` package for cross-platform symbols. Windows Terminal supports most, but older cmd.exe does not.
- **Clear the line before overwriting.** When updating spinner text, ora handles this. If doing it manually, use `process.stderr.write('\r\x1b[K')` to clear the current line.
- **listr2 captures stdout.** Inside a listr2 task, `console.log()` is captured and shown as task output. Use `task.output = 'message'` instead.
- **boxen adds visual margin.** Don't stack multiple boxen calls — they add too much vertical space. Use one box for a summary, not one per item.
