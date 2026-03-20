# CLI Tool Patterns Reference

Quick reference for common CLI patterns, conventions, and recipes.

---

## Command Design Cheat Sheet

### Naming Conventions

| Pattern | Example | When to Use |
|---------|---------|-------------|
| verb-noun | `deploy app`, `create project` | Standard CRUD operations |
| noun-verb | `app deploy`, `project create` | Resource-centric tools (like `kubectl`) |
| verb only | `init`, `login`, `build` | Single-purpose or top-level actions |
| noun only | `config`, `logs`, `status` | When the action is implied |

### Standard Verbs

| Verb | Meaning | Aliases |
|------|---------|---------|
| `create` | Make new resource | `new`, `add`, `init` |
| `list` | Show all resources | `ls`, `get` (plural) |
| `get` | Show single resource | `show`, `describe`, `info` |
| `update` | Modify resource | `edit`, `set`, `modify` |
| `delete` | Remove resource | `rm`, `remove`, `destroy` |
| `deploy` | Push to environment | `push`, `release`, `ship` |
| `login` | Authenticate | `auth`, `signin` |
| `logout` | De-authenticate | `signout` |
| `sync` | Synchronize state | `pull`, `fetch` |
| `watch` | Monitor changes | `follow`, `tail` |

### Flag Conventions (Universal)

```
-h, --help          Always present (Commander adds automatically)
-v, --version       Always present (Commander adds automatically)
-V, --verbose       Extra output detail
-q, --quiet         Suppress non-essential output
-f, --force         Skip confirmations
-n, --dry-run       Show what would happen
-o, --output <fmt>  Output format (json, table, csv, yaml)
-c, --config <path> Config file path
    --json          Shorthand for -o json
    --no-color      Disable colors (respects NO_COLOR env)
```

---

## Error Handling Patterns

### Error Display Template

```
ERROR: <what went wrong — human readable>
  → <relevant context key>: <value>
  → <another context key>: <value>

Try:
  1. <concrete action — ideally a command>
  2. <alternative action>
  3. <last resort — docs link>
```

### Exit Code Reference

| Code | Constant | Meaning |
|------|----------|---------|
| 0 | `EXIT_SUCCESS` | Everything worked |
| 1 | `EXIT_ERROR` | General runtime error |
| 2 | `EXIT_USAGE` | Invalid arguments/options |
| 126 | `EXIT_NOPERM` | Permission denied |
| 127 | `EXIT_NOTFOUND` | Command/dependency not found |
| 130 | `EXIT_SIGINT` | User pressed Ctrl+C |

```typescript
// Define as constants
const EXIT = { SUCCESS: 0, ERROR: 1, USAGE: 2, SIGINT: 130 } as const;
```

### Error Classification

```typescript
class ConfigError extends CLIError {
  constructor(key: string, file: string) {
    super(`Missing required config: ${key}`, [
      `Add "${key}" to ${file}`,
      `Set via environment: export MYCLI_${key.toUpperCase()}=value`,
      `Or pass via flag: --${key} <value>`,
    ], { configFile: file });
  }
}

class NetworkError extends CLIError {
  constructor(url: string, statusCode: number) {
    super(`API request failed (${statusCode})`, [
      statusCode === 401 ? 'Run: my-cli login' : undefined,
      statusCode === 404 ? `Check resource exists: my-cli list` : undefined,
      statusCode >= 500 ? 'Wait and retry, or check https://status.example.com' : undefined,
    ].filter(Boolean) as string[], { url, statusCode: String(statusCode) });
  }
}
```

---

## Configuration Patterns

### cosmiconfig Setup

```typescript
import { cosmiconfig, cosmiconfigSync } from 'cosmiconfig';

// Searches for: .mytoolrc, .mytoolrc.json, .mytoolrc.yaml,
// .mytoolrc.yml, .mytoolrc.js, .mytoolrc.ts, .mytoolrc.cjs,
// .mytoolrc.mjs, mytool.config.js, mytool.config.ts,
// mytool.config.cjs, mytool.config.mjs, "mytool" in package.json
const explorer = cosmiconfig('mytool');

// Async (preferred)
const result = await explorer.search();
// result.config — the parsed config object
// result.filepath — where it was found
// result.isEmpty — if the file was empty

// Sync alternative
const syncResult = cosmiconfigSync('mytool').search();
```

### Config Priority Resolution

```typescript
interface ResolvedConfig {
  apiUrl: string;
  token?: string;
  timeout: number;
  verbose: boolean;
}

function resolveConfig(
  cliFlags: Partial<ResolvedConfig>,
  fileConfig: Partial<ResolvedConfig>,
): ResolvedConfig {
  return {
    // CLI flags > env vars > config file > defaults
    apiUrl: cliFlags.apiUrl
      ?? process.env.MYCLI_API_URL
      ?? fileConfig.apiUrl
      ?? 'https://api.example.com',

    token: cliFlags.token
      ?? process.env.MYCLI_TOKEN
      ?? fileConfig.token,

    timeout: cliFlags.timeout
      ?? (process.env.MYCLI_TIMEOUT ? parseInt(process.env.MYCLI_TIMEOUT, 10) : undefined)
      ?? fileConfig.timeout
      ?? 30000,

    verbose: cliFlags.verbose ?? fileConfig.verbose ?? false,
  };
}
```

### XDG Config Paths

```typescript
import { homedir } from 'os';
import { join } from 'path';

function getConfigDir(appName: string): string {
  // XDG_CONFIG_HOME on Linux, ~/Library/Preferences on macOS
  const xdg = process.env.XDG_CONFIG_HOME;
  if (xdg) return join(xdg, appName);

  if (process.platform === 'darwin') {
    return join(homedir(), 'Library', 'Preferences', appName);
  }

  return join(homedir(), '.config', appName);
}

function getDataDir(appName: string): string {
  const xdg = process.env.XDG_DATA_HOME;
  if (xdg) return join(xdg, appName);
  return join(homedir(), '.local', 'share', appName);
}

function getCacheDir(appName: string): string {
  const xdg = process.env.XDG_CACHE_HOME;
  if (xdg) return join(xdg, appName);
  return join(homedir(), '.cache', appName);
}
```

---

## Distribution Recipes

### npm Publishing Checklist

```json
// package.json essentials
{
  "name": "my-cli",
  "version": "1.0.0",
  "bin": { "my-cli": "./dist/index.js" },
  "files": ["dist"],
  "engines": { "node": ">=18" },
  "scripts": {
    "build": "tsc",
    "prepublishOnly": "npm run build && npm test"
  }
}
```

```
# .npmignore (or use "files" allowlist in package.json)
src/
tests/
*.test.ts
.github/
tsconfig.json
```

### Homebrew Formula Template

```ruby
# Formula/my-cli.rb
class MyCli < Formula
  desc "My awesome CLI tool"
  homepage "https://github.com/user/my-cli"
  url "https://registry.npmjs.org/my-cli/-/my-cli-1.0.0.tgz"
  sha256 "abc123..."

  depends_on "node"

  def install
    system "npm", "install", *std_npm_args
    bin.install_symlink Dir["#{libexec}/bin/*"]
  end

  test do
    assert_match "1.0.0", shell_output("#{bin}/my-cli --version")
  end
end
```

### Single Binary with pkg

```bash
# Install pkg
npm install -D @yao-pkg/pkg

# Add to package.json
# "scripts": { "package": "pkg . --targets node18-macos-x64,node18-linux-x64,node18-win-x64" }

# Build
npm run package
# Output: my-cli-macos, my-cli-linux, my-cli-win.exe
```

### Docker Distribution

```dockerfile
FROM node:20-alpine
WORKDIR /app
COPY package*.json ./
RUN npm ci --production
COPY dist/ ./dist/
ENTRYPOINT ["node", "dist/index.js"]
```

```bash
# Build and publish
docker build -t my-cli .
docker tag my-cli ghcr.io/user/my-cli:latest
docker push ghcr.io/user/my-cli:latest

# Users run with:
docker run --rm -v $(pwd):/workspace ghcr.io/user/my-cli init
```

---

## Testing Patterns

### Testing Commands (Integration)

```typescript
import { describe, it, expect } from 'vitest';
import { execSync } from 'child_process';

describe('my-cli', () => {
  const cli = (args: string) =>
    execSync(`tsx src/index.ts ${args}`, { encoding: 'utf-8' });

  it('shows version', () => {
    const output = cli('--version');
    expect(output.trim()).toMatch(/^\d+\.\d+\.\d+$/);
  });

  it('init creates project', () => {
    const output = cli('init test-project --template blank --force');
    expect(output).toContain('Created');
  });

  it('exits with code 2 on invalid args', () => {
    expect(() => cli('--invalid-flag')).toThrow();
  });
});
```

### Testing Command Logic (Unit)

```typescript
import { describe, it, expect, vi } from 'vitest';

// Extract logic from command action into testable functions
describe('deploy', () => {
  it('validates environment', () => {
    expect(() => parseEnvironment('staging')).not.toThrow();
    expect(() => parseEnvironment('invalid')).toThrow();
  });

  it('applies defaults correctly', () => {
    const config = resolveConfig({}, {});
    expect(config.apiUrl).toBe('https://api.example.com');
    expect(config.timeout).toBe(30000);
  });
});
```

### Snapshot Testing Output

```typescript
import { describe, it, expect } from 'vitest';

describe('output formatting', () => {
  it('formats table correctly', () => {
    const output = formatTable([
      { name: 'prod', status: 'active' },
      { name: 'staging', status: 'paused' },
    ]);
    expect(output).toMatchSnapshot();
  });
});
```

---

## Signal Handling

```typescript
// Handle Ctrl+C gracefully
process.on('SIGINT', () => {
  console.error('\nInterrupted. Cleaning up...');
  // Cleanup: remove temp files, close connections, etc.
  process.exit(130);
});

// Handle termination (Docker stop, kill)
process.on('SIGTERM', () => {
  console.error('Terminated. Cleaning up...');
  process.exit(143); // 128 + SIGTERM (15)
});

// Prevent broken pipe errors (stdout closed before write completes)
process.stdout.on('error', (err) => {
  if (err.code === 'EPIPE') process.exit(0);
  throw err;
});
```

---

## Update Notifications

```typescript
import boxen from 'boxen';
import chalk from 'chalk';

async function checkForUpdates(currentVersion: string, packageName: string) {
  try {
    const response = await fetch(`https://registry.npmjs.org/${packageName}/latest`);
    const { version: latest } = await response.json();

    if (latest !== currentVersion) {
      console.error(boxen(
        `Update available: ${chalk.dim(currentVersion)} → ${chalk.green(latest)}\n` +
        `Run ${chalk.cyan(`npm i -g ${packageName}`)} to update`,
        { padding: 1, borderStyle: 'round', borderColor: 'yellow' }
      ));
    }
  } catch {
    // Silently fail — don't break the tool for update checks
  }
}
```

---

## Stdin Piping Support

```typescript
import { createInterface } from 'readline';

async function readStdin(): Promise<string> {
  if (process.stdin.isTTY) return ''; // No piped input

  const lines: string[] = [];
  const rl = createInterface({ input: process.stdin });

  for await (const line of rl) {
    lines.push(line);
  }

  return lines.join('\n');
}

// Usage in command:
program
  .command('process')
  .argument('[file]', 'input file (or pipe via stdin)')
  .action(async (file) => {
    let input: string;

    if (file) {
      input = readFileSync(file, 'utf-8');
    } else {
      input = await readStdin();
      if (!input) {
        console.error('Error: Provide a file argument or pipe input via stdin');
        process.exit(2);
      }
    }

    // Process input...
  });

// Works with:
// my-cli process data.json
// cat data.json | my-cli process
// echo '{"key":"value"}' | my-cli process
```

---

## Relative Timestamps

```typescript
function timeAgo(date: Date | string): string {
  const now = Date.now();
  const then = new Date(date).getTime();
  const seconds = Math.floor((now - then) / 1000);

  if (seconds < 60) return 'just now';
  if (seconds < 3600) return `${Math.floor(seconds / 60)} min ago`;
  if (seconds < 86400) return `${Math.floor(seconds / 3600)} hours ago`;
  if (seconds < 604800) return `${Math.floor(seconds / 86400)} days ago`;
  if (seconds < 2592000) return `${Math.floor(seconds / 604800)} weeks ago`;

  return new Date(date).toLocaleDateString();
}

// In tables: "2 hours ago" instead of "2026-03-19T14:23:00Z"
```

---

## Package Size Optimization

| Technique | Impact | How |
|-----------|--------|-----|
| Tree-shake imports | -30-50% | `import chalk from 'chalk'` not `import * as chalk` |
| Skip heavy deps | Major | Use `figures` (5KB) not `cli-spinners` (50KB) for just symbols |
| Bundle with esbuild | -60-80% | `esbuild src/index.ts --bundle --platform=node --outfile=dist/index.js` |
| Lazy-load commands | Startup time | Import command modules only when that command is invoked |

### Lazy Command Loading

```typescript
program
  .command('deploy')
  .description('Deploy to environment')
  .action(async (...args) => {
    // Only import the deploy module when the command is used
    const { deployAction } = await import('./commands/deploy.js');
    return deployAction(...args);
  });
```
