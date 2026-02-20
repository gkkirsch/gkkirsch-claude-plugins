---
name: cloudflare-deployment
description: >
  Use when deploying an application to Cloudflare. Covers the full deployment lifecycle:
  Pages (static sites, SPAs, SSR frameworks), Workers (API backends, serverless functions,
  cron jobs), data services (D1, KV, R2), wrangler CLI setup, configuration, custom domains,
  monitoring, and troubleshooting. Use when: (1) user says "deploy to cloudflare", "cloudflare
  deploy", "push to cloudflare", "cloudflare workers", "cloudflare pages", (2) user is building
  a web app or API and needs edge hosting, (3) user needs to troubleshoot a Cloudflare
  deployment, (4) user needs to set up D1, KV, or R2, (5) user is migrating from Heroku
  to Cloudflare.
version: 1.0.0
---

# Cloudflare Deployment

Deploy web applications and APIs to Cloudflare's edge network. This skill covers the complete lifecycle from project setup through production monitoring.

For deep-dive references:
- Pages deployment (static sites, SSR frameworks) → `references/pages-deployment.md`
- Workers deployment (APIs, serverless, cron) → `references/workers-deployment.md`
- wrangler.toml configuration patterns → `references/wrangler-config.md`
- Data services (D1, KV, R2) → `references/data-services.md`
- Environment variables and secrets → `references/environment-secrets.md`
- Custom domains and DNS → `references/domains-dns.md`
- Local dev, monitoring, and debugging → `references/monitoring-debugging.md`
- Common errors and fixes → `references/troubleshooting.md`
- Heroku → Cloudflare migration → `references/heroku-migration.md`

## Prerequisites

- **Wrangler CLI**: `npm install -g wrangler` (or per-project: `npm install --save-dev wrangler`)
- **Node.js**: 18+
- **Cloudflare account**: `wrangler login` to authenticate (opens browser OAuth)

Verify setup:

```bash
wrangler --version
wrangler whoami
```

For CI/CD (non-interactive auth):

1. Create API token at https://dash.cloudflare.com/profile/api-tokens
2. Set environment variable: `export CLOUDFLARE_API_TOKEN=your-token`

## Decision Tree: What to Deploy

```
Is it a static site or SPA (React, Vue, Astro static)?
├── YES → Cloudflare Pages
│
├── Does it need server-side rendering (Next.js, Astro SSR, Nuxt)?
│   ├── YES → Cloudflare Pages with Functions
│
├── Is it an API backend or serverless function?
│   ├── YES → Cloudflare Workers
│
├── Is it a full-stack app (frontend + API)?
│   ├── YES → Pages (frontend) + Workers (API), or Pages with Functions
│
└── Migrating from Heroku?
    └── See references/heroku-migration.md
```

### Quick Comparison

| Feature | Pages | Workers | Pages + Functions |
|---------|-------|---------|-------------------|
| Static sites | Yes | No | Yes |
| SSR frameworks | Via Functions | Manual | Yes |
| API endpoints | Via Functions | Yes | Yes |
| Git integration | Yes | No (use CI) | Yes |
| Free tier | Unlimited sites, 500 builds/mo | 100K req/day | Combined |
| Cron triggers | No | Yes | No |
| WebSockets | No | Yes | No |
| Durable Objects | No | Yes | No |

## 1. Deploy a Static Site (Pages)

### Quick deploy via Wrangler

```bash
# Build your project
npm run build

# First deploy creates the project
wrangler pages deploy dist/ --project-name=my-app

# Subsequent deploys update it
npm run build && wrangler pages deploy dist/ --project-name=my-app
```

### Deploy via Git integration

1. Go to https://dash.cloudflare.com → Workers & Pages → Create
2. Connect GitHub/GitLab repo
3. Configure build settings:
   - **Build command**: `npm run build`
   - **Output directory**: `dist` (Vite), `build` (CRA), `.next` (Next.js), `out` (Astro static)
4. Deploy — every push to production branch auto-deploys, PRs get preview URLs

### Framework build output directories

| Framework | Build command | Output directory |
|-----------|---------------|-----------------|
| Vite / React | `npm run build` | `dist` |
| Next.js | `npx @cloudflare/next-on-pages` | `.vercel/output/static` |
| Astro (static) | `npm run build` | `dist` |
| SvelteKit | `npm run build` | `.svelte-kit/cloudflare` |

For detailed framework setup (Next.js, Astro SSR, SvelteKit), see `references/pages-deployment.md`.

### Pages with API Functions

Add server-side logic with file-based routing in a `functions/` directory:

```
my-app/
├── functions/
│   └── api/
│       ├── hello.ts          → GET/POST /api/hello
│       └── users/
│           ├── index.ts      → GET/POST /api/users
│           └── [id].ts       → GET/POST /api/users/:id
├── src/
└── dist/
```

```ts
// functions/api/hello.ts
export const onRequestGet: PagesFunction = async (context) => {
  return Response.json({ message: "Hello from the edge!" });
};
```

## 2. Deploy an API (Workers)

### Create and deploy a Worker

```bash
# Create a new worker project
wrangler init my-api
cd my-api

# Edit src/index.ts, then deploy
wrangler deploy
```

### Basic handler

```ts
// src/index.ts
export default {
  async fetch(request: Request, env: Env, ctx: ExecutionContext): Promise<Response> {
    const url = new URL(request.url);

    if (url.pathname === "/api/hello") {
      return Response.json({ message: "Hello!" });
    }

    return new Response("Not found", { status: 404 });
  },
};
```

### With Hono (recommended routing framework)

```bash
npm install hono
```

```ts
// src/index.ts
import { Hono } from "hono";
import { cors } from "hono/cors";

type Bindings = {
  DB: D1Database;
  API_KEY: string;
};

const app = new Hono<{ Bindings: Bindings }>();
app.use("*", cors());

app.get("/api/users", async (c) => {
  const { results } = await c.env.DB.prepare("SELECT * FROM users").all();
  return c.json(results);
});

app.post("/api/users", async (c) => {
  const { name, email } = await c.req.json();
  await c.env.DB.prepare("INSERT INTO users (name, email) VALUES (?, ?)")
    .bind(name, email)
    .run();
  return c.json({ success: true }, 201);
});

export default app;
```

For cron handlers, queue consumers, Durable Objects, and Service Bindings, see `references/workers-deployment.md`.

## 3. Configure (wrangler.toml)

### Minimal Worker config

```toml
name = "my-api"
main = "src/index.ts"
compatibility_date = "2024-01-01"
```

### Minimal Pages config

```toml
name = "my-site"
pages_build_output_dir = "./dist"
```

### Add bindings (database, KV, storage)

```toml
name = "my-api"
main = "src/index.ts"
compatibility_date = "2024-01-01"

[vars]
ENVIRONMENT = "production"

[[d1_databases]]
binding = "DB"
database_name = "my-database"
database_id = "abc123"

[[kv_namespaces]]
binding = "CACHE"
id = "def456"

[[r2_buckets]]
binding = "UPLOADS"
bucket_name = "user-uploads"
```

### Multi-environment

```toml
[vars]
ENVIRONMENT = "production"

[env.staging]
name = "my-api-staging"
vars = { ENVIRONMENT = "staging" }
```

```bash
wrangler deploy --env staging
```

For full configuration reference, see `references/wrangler-config.md`.

## 4. Data Services

### D1 (SQLite database)

```bash
# Create database
wrangler d1 create my-database

# Create and apply a migration
wrangler d1 migrations create my-database create_users
# Edit the migration SQL, then:
wrangler d1 migrations apply my-database --remote
```

```ts
// Query in code
const user = await env.DB.prepare("SELECT * FROM users WHERE id = ?").bind(id).first();
```

### KV (key-value store)

```bash
wrangler kv namespace create MY_KV
```

```ts
await env.MY_KV.put("session:abc", data, { expirationTtl: 3600 });
const value = await env.MY_KV.get("session:abc", { type: "json" });
```

### R2 (object storage)

```bash
wrangler r2 bucket create my-bucket
```

```ts
await env.BUCKET.put("images/photo.jpg", imageData, {
  httpMetadata: { contentType: "image/jpeg" },
});
const object = await env.BUCKET.get("images/photo.jpg");
```

For full API reference, migration patterns, and choosing between services, see `references/data-services.md`.

## 5. Environment Variables and Secrets

```bash
# Secrets (encrypted, not in wrangler.toml)
wrangler secret put API_KEY
wrangler secret put API_KEY --env staging

# Variables (in wrangler.toml)
[vars]
API_URL = "https://api.example.com"
```

For local development, create `.dev.vars` (add to `.gitignore`):
```
API_KEY=sk-local-dev-key
```

For full environment management, see `references/environment-secrets.md`.

## 6. Local Development

```bash
# Workers
wrangler dev

# Pages with Functions
wrangler pages dev dist/

# Pages with framework dev server
wrangler pages dev -- npm run dev
```

Local dev runs on `http://localhost:8787` with hot reload and local binding emulation.

## 7. Custom Domains

**Pages**: Add in dashboard → Project → Custom domains

**Workers**: Configure routes in `wrangler.toml`:
```toml
routes = [
  { pattern = "api.example.com", custom_domain = true }
]
```

SSL is automatic for all Cloudflare-proxied domains. See `references/domains-dns.md`.

## 8. Monitoring

```bash
# Stream real-time logs from production
wrangler tail

# Filter by status
wrangler tail --status error

# Filter by method
wrangler tail --method POST

# Filter by content
wrangler tail --search "database"
```

For structured logging patterns, Chrome DevTools debugging, and health checks, see `references/monitoring-debugging.md`.

## Key Constraints

- **Workers run V8 isolates, NOT Node.js.** Most Node.js APIs unavailable. Use Web Standard APIs (fetch, Request, Response, crypto). Enable `nodejs_compat` flag for partial Node.js support.
- **1MB code size limit (free) / 10MB (paid).** No `node_modules` at runtime — everything is bundled.
- **No persistent filesystem.** Use D1, KV, or R2 for storage.
- **CPU time: 10ms (free) / 30s (paid) per request.** Wall-clock time is unlimited (I/O doesn't count).
- **D1 is SQLite, not Postgres.** Different syntax — `INTEGER PRIMARY KEY AUTOINCREMENT` not `SERIAL`, `TEXT` not `VARCHAR`, no `JSONB`.
- **KV is eventually consistent.** Writes take up to 60 seconds to propagate globally.

## Quick Reference

| Task | Command |
|------|---------|
| Install Wrangler | `npm install -g wrangler` |
| Authenticate | `wrangler login` |
| Deploy Pages | `wrangler pages deploy dist/ --project-name=my-app` |
| Deploy Worker | `wrangler deploy` |
| Local dev (Worker) | `wrangler dev` |
| Local dev (Pages) | `wrangler pages dev dist/` |
| Stream logs | `wrangler tail` |
| Set secret | `wrangler secret put API_KEY` |
| Create D1 DB | `wrangler d1 create my-database` |
| Apply D1 migration | `wrangler d1 migrations apply my-database --remote` |
| Create KV namespace | `wrangler kv namespace create MY_KV` |
| Create R2 bucket | `wrangler r2 bucket create my-bucket` |
| Deploy to staging | `wrangler deploy --env staging` |

## Troubleshooting

For common issues (size limits, module resolution, CORS, D1 errors, build failures), see `references/troubleshooting.md`.

Quick checks when something goes wrong:

```bash
# Check logs first — always
wrangler tail

# Check deployment status
wrangler deployments list

# Check local dev
wrangler dev --log-level debug

# Redeploy
wrangler deploy
```
