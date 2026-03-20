# Serverless & Edge Platform Comparison

## Platform Matrix

| Feature | AWS Lambda | Vercel Functions | Cloudflare Workers | Deno Deploy | Netlify Functions |
|---------|-----------|-----------------|-------------------|-------------|------------------|
| **Runtime** | Node.js, Python, Go, Java, .NET, Ruby, Rust | Node.js, Edge (V8) | V8 isolates | Deno (V8) | Node.js, Deno |
| **Cold start** | 100-500ms (Node.js), ~0ms (SnapStart) | ~250ms (Node.js), ~0ms (Edge) | ~0ms | ~0ms | 200-500ms |
| **Max duration** | 15 min | 60s (Hobby), 300s (Pro) | 30s (paid), 10ms CPU (free) | 50ms CPU (free) | 10s (sync), 15 min (background) |
| **Max memory** | 10,240 MB | 1,024 MB (Node.js), 128 MB (Edge) | 128 MB | 512 MB | 1,024 MB |
| **Max payload** | 6 MB (sync), 256 KB (async) | 4.5 MB (Node.js), 4 MB (Edge) | 100 MB | Unknown | 6 MB |
| **Regions** | Single (configurable) | Single or Edge | 200+ edge locations | 35+ edge locations | Single (configurable) |
| **Free tier** | 1M requests + 400K GB-s/mo | 100 GB-hours/mo | 100K requests/day | 100K requests/day | 125K requests/mo |

## Pricing Comparison

### AWS Lambda
```
Per request:  $0.20 per 1M requests
Per compute:  $0.0000166667 per GB-second
Per GB:       $0.0000166667 * memory_gb * duration_seconds

Example: 128MB function, 200ms avg, 1M requests/month
= $0.20 + (0.128 * 0.2 * 1,000,000 * $0.0000166667)
= $0.20 + $0.43
= $0.63/month
```

### Cloudflare Workers
```
Free:         100,000 requests/day (10ms CPU limit)
Paid ($5/mo): 10M requests included
Overage:      $0.30 per additional 1M requests

Example: 1M requests/month
= $5.00/month (well within included 10M)
```

### Vercel
```
Hobby (Free): 100 GB-hours/mo, 100K function invocations
Pro ($20/mo):  1,000 GB-hours/mo, 1M invocations
Overage:       $0.18 per 100 additional GB-hours

Example: 128MB function, 200ms avg, 1M requests/month
= 0.128 * 0.2/3600 * 1,000,000 = 7.1 GB-hours
= Free tier covers it
```

### Deno Deploy
```
Free:         100K requests/day, 1M KV reads/mo
Pro ($20/mo): 5M requests/mo, 1B KV reads/mo, custom domains
```

## Deployment Checklists

### AWS Lambda Deployment Checklist

```
Pre-deploy:
[ ] Bundle with esbuild (target: node20, external: @aws-sdk/*)
[ ] Set memory to optimal (use AWS Lambda Power Tuning)
[ ] Set timeout appropriately (not max 15 min unless needed)
[ ] Environment variables in Parameter Store / Secrets Manager
[ ] IAM role with least-privilege permissions
[ ] VPC config only if needed (adds cold start)

Deploy:
[ ] Use infrastructure-as-code (SST, SAM, CDK, or Terraform)
[ ] Enable X-Ray tracing
[ ] Set reserved concurrency if needed
[ ] Configure dead-letter queue for async invocations
[ ] Set up CloudWatch alarms (errors, throttles, duration)

Post-deploy:
[ ] Verify cold start time
[ ] Check CloudWatch logs for errors
[ ] Test with realistic payload sizes
[ ] Verify IAM permissions are correct
[ ] Monitor cost in AWS Cost Explorer
```

### Cloudflare Workers Deployment Checklist

```
Pre-deploy:
[ ] Bundle size under 1 MB (free) or 10 MB (paid)
[ ] No Node.js-specific APIs used (fs, path, net, etc.)
[ ] Secrets via `wrangler secret put` (not in wrangler.toml)
[ ] KV namespaces created and bound
[ ] D1 databases created with migrations applied
[ ] R2 buckets created and bound

Deploy:
[ ] Test locally with `wrangler dev`
[ ] Deploy to staging first (`wrangler deploy --env staging`)
[ ] Verify bindings work in deployed environment
[ ] Set up custom domain if needed
[ ] Configure cron triggers if using scheduled workers

Post-deploy:
[ ] Check Workers analytics dashboard
[ ] Verify KV propagation (eventual consistency, ~60s)
[ ] Test from multiple regions
[ ] Monitor CPU time (not wall clock time)
```

### Vercel Deployment Checklist

```
Pre-deploy:
[ ] API routes in correct directory (app/api/ for App Router)
[ ] Runtime specified where needed (export const runtime = 'edge')
[ ] Environment variables set in Vercel dashboard
[ ] Cron jobs configured in vercel.json with CRON_SECRET
[ ] Database connection pooling configured (Prisma/pg)
[ ] Middleware matcher configured (avoid running on static files)

Deploy:
[ ] Push to branch for preview deployment
[ ] Verify preview URL works
[ ] Check function logs in Vercel dashboard
[ ] Merge to main for production deployment

Post-deploy:
[ ] Verify all API routes respond correctly
[ ] Check function execution times
[ ] Monitor GB-hours usage
[ ] Test cron jobs fire on schedule
[ ] Verify middleware runs on correct paths only
```

## Limits Quick Reference

### Request/Response Limits

| Platform | Max Request Body | Max Response Body | Max Headers |
|----------|-----------------|-------------------|-------------|
| Lambda (API GW) | 6 MB | 6 MB | 10 KB total |
| Lambda (Function URL) | 6 MB | 6 MB (stream: 20 MB) | 8 KB total |
| Vercel (Node.js) | 4.5 MB | 4.5 MB (stream: unlimited) | 16 KB |
| Vercel (Edge) | 4 MB | 4 MB (stream: unlimited) | 16 KB |
| CF Workers | 100 MB | Unlimited (stream) | 32 KB |

### Concurrency Limits

| Platform | Default Concurrency | Max Configurable |
|----------|-------------------|-----------------|
| Lambda | 1,000 per region | 10,000+ (request increase) |
| Vercel (Hobby) | 10 concurrent | N/A |
| Vercel (Pro) | 1,000 concurrent | Custom |
| CF Workers (Free) | Unlimited* | N/A |
| CF Workers (Paid) | Unlimited* | N/A |

*Workers handle concurrency via V8 isolates. No per-account limit, but per-script limits apply.

### Storage Limits

| Storage | Max Key Size | Max Value Size | Max Keys | Consistency |
|---------|-------------|---------------|----------|-------------|
| CF KV | 512 bytes | 25 MB | Unlimited | Eventually (~60s) |
| CF D1 | N/A | N/A | 10 GB/database | Strong (single region) |
| CF R2 | 1,024 bytes | 5 TB | Unlimited | Strong |
| Vercel KV (Redis) | 512 MB total | 100 MB/value | N/A | Strong |
| Vercel Blob | N/A | 500 MB/file | N/A | Strong |
| DynamoDB | 2,048 bytes (PK) | 400 KB/item | Unlimited | Eventually or Strong |

## Common Patterns by Use Case

### Authentication & Authorization
```
Best platform: Edge (Cloudflare Workers or Vercel Edge)
Why: Zero cold start, runs before origin, minimal compute needed
Pattern: JWT verification at edge, pass user info via headers to origin
Libraries: jose (edge-compatible JWT), @noble/hashes
```

### REST API
```
Best platform: AWS Lambda (complex) or Vercel (Next.js projects)
Why: Full Node.js runtime, database access, complex business logic
Pattern: API Gateway + Lambda functions, or Next.js API routes
Libraries: middy (Lambda), zod (validation), prisma (ORM)
```

### Real-time / WebSocket
```
Best platform: Cloudflare Durable Objects
Why: Stateful edge compute, WebSocket support, global coordination
Alternative: AWS API Gateway WebSocket API + Lambda
Pattern: Durable Object per room/session, WebSocket messages
```

### File Processing
```
Best platform: AWS Lambda (up to 10GB /tmp, 15 min duration)
Why: Highest memory/storage/duration limits
Pattern: S3 trigger → Lambda → process → write to S3
Alternative: Cloudflare Workers with R2 for smaller files
```

### Scheduled Jobs / Cron
```
Best platform: Cloudflare Workers (simple) or Lambda (complex)
CF Workers: cron triggers in wrangler.toml, simple KV-based state
Lambda: EventBridge rules, Step Functions for complex workflows
Vercel: vercel.json crons (limited to 1/day on Hobby)
```

### CDN / Content Transformation
```
Best platform: Cloudflare Workers
Why: Runs at 200+ edge locations, HTMLRewriter API for streaming transforms
Pattern: Intercept response, modify HTML/headers, serve from edge cache
Alternative: Lambda@Edge (limited to CloudFront edge locations)
```

## Migration Cheat Sheet

### Express to Vercel Serverless
```typescript
// Before (Express):
app.get('/api/users', handler);
app.post('/api/users', handler);

// After (Next.js App Router):
// app/api/users/route.ts
export async function GET(request: NextRequest) { ... }
export async function POST(request: NextRequest) { ... }

// Key changes:
// - req.query → request.nextUrl.searchParams
// - req.body → await request.json()
// - res.json() → NextResponse.json()
// - req.headers → request.headers.get()
// - middleware → middleware.ts with matcher config
```

### Express to Cloudflare Workers (Hono)
```typescript
// Before (Express):
const app = express();
app.use(cors());
app.get('/api/users', (req, res) => { ... });

// After (Hono):
const app = new Hono();
app.use('*', cors());
app.get('/api/users', (c) => { ... });

// Key changes:
// - req.query → c.req.query()
// - req.body → c.req.json()
// - res.json(data) → c.json(data)
// - req.params → c.req.param()
// - process.env → c.env (bindings)
// - npm packages → check edge compatibility
```

### Express to AWS Lambda
```typescript
// Before (Express):
app.get('/api/users', (req, res) => { ... });

// After (Lambda):
export const handler = async (event: APIGatewayProxyEvent) => {
  return { statusCode: 200, body: JSON.stringify(data) };
};

// Key changes:
// - req.query → event.queryStringParameters
// - req.body → JSON.parse(event.body)
// - res.json() → { statusCode, body: JSON.stringify() }
// - req.headers → event.headers
// - middleware → middy() wraps
// - Or use serverless-http to wrap existing Express app (quick migration)
```

## Decision Flowchart

```
Start
  ├── Need global low-latency? (< 50ms worldwide)
  │   ├── Yes → Does it need database access?
  │   │   ├── Yes → Cloudflare Workers + D1/KV
  │   │   └── No → Cloudflare Workers or Vercel Edge
  │   └── No → Continue ↓
  ├── Already using Next.js?
  │   ├── Yes → Vercel Functions (API routes)
  │   └── No → Continue ↓
  ├── Need long-running tasks? (> 30 seconds)
  │   ├── Yes → AWS Lambda (up to 15 min)
  │   └── No → Continue ↓
  ├── Need complex AWS integrations? (S3, SQS, DynamoDB)
  │   ├── Yes → AWS Lambda
  │   └── No → Continue ↓
  ├── Want simplest deployment?
  │   ├── Yes → Cloudflare Workers (wrangler deploy)
  │   └── No → Continue ↓
  └── Default → Vercel (if web app) or Lambda (if backend/API)
```
