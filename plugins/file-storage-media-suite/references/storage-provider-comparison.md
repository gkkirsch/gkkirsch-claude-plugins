# Storage Provider Comparison

## Full Feature Matrix

| Feature | AWS S3 | Cloudflare R2 | Cloudinary | Vercel Blob | Supabase | GCS | Azure Blob |
|---------|--------|---------------|------------|-------------|----------|-----|-----------|
| **Pricing** | | | | | | | |
| Storage/GB/mo | $0.023 | $0.015 | Included | Included | $0.021 | $0.020 | $0.018 |
| Egress/GB | $0.09 | **$0.00** | Included | Included | $0.09 | $0.12 | $0.087 |
| PUT/1K reqs | $0.005 | $0.0045 | Included | Included | N/A | $0.005 | $0.005 |
| GET/1K reqs | $0.0004 | $0.00036 | Included | Included | N/A | $0.0004 | $0.004 |
| **Free Tier** | 5GB/12mo | 10GB | 25K transforms | 1GB | 1GB | 5GB | 5GB |
| **Features** | | | | | | | |
| S3-compatible API | Yes | Yes | No | No | Yes | Yes (interop) | No |
| CDN built-in | CloudFront | Yes | Yes | Edge | CDN | Cloud CDN | Azure CDN |
| Image transforms | No | Basic | **Full** | Basic | Basic | No | No |
| Video processing | No | Stream | **Full** | No | No | Transcoder | Media Services |
| Lifecycle rules | Yes | Yes | No | No | No | Yes | Yes |
| Versioning | Yes | No | Yes | No | No | Yes | Yes |
| Encryption at rest | Yes (SSE) | Yes | Yes | Yes | Yes | Yes | Yes |
| Multipart upload | Yes | Yes | Yes | No | Yes | Yes | Yes |
| Presigned URLs | Yes | Yes | Signed | Token | Yes | Yes | SAS tokens |
| Max object size | 5TB | 5TB | Varies | 500MB | 50GB | 5TB | ~5TB |
| Regions | 30+ | Global edge | N/A | Edge | 12+ | 30+ | 60+ |

## Cost Comparison at Scale

### Monthly cost for serving 1TB of images to 100K users

| Provider | Storage (100GB) | Egress (1TB) | Requests (10M) | Total |
|----------|----------------|-------------|----------------|-------|
| **S3 + CloudFront** | $2.30 | $85.00 | $4.00 | ~$91 |
| **Cloudflare R2** | $1.50 | **$0.00** | $3.60 | ~$5 |
| **Cloudinary** | Included | Included | Included | ~$89 (Advanced plan) |
| **Vercel Blob** | Included | Included | Included | ~$20 (Pro plan) |
| **Supabase** | $2.10 | $90.00 | N/A | ~$92 |

**Verdict**: R2 wins on cost for egress-heavy workloads. Cloudinary wins on features if you need transforms. Vercel Blob is simplest for Next.js.

## Provider Deep Dives

### AWS S3 — When to Use

**Pros**: Industry standard, most tools/libraries support it, full AWS ecosystem, advanced features (versioning, replication, lifecycle, analytics, Lambda triggers).

**Cons**: Egress costs add up fast, complex IAM setup, no built-in image transforms.

**Best for**: Enterprise apps, apps already on AWS, complex storage needs (lifecycle, cross-region replication, compliance).

```
S3 Standard:     $0.023/GB/mo   — Frequently accessed
S3 Standard-IA:  $0.0125/GB/mo  — Infrequent access (30-day minimum)
S3 Glacier:      $0.004/GB/mo   — Archive (minutes to hours retrieval)
S3 Deep Archive: $0.00099/GB/mo — Long-term archive (12-48 hour retrieval)
```

### Cloudflare R2 — When to Use

**Pros**: Zero egress fees (huge savings), S3-compatible API (easy migration), global edge caching, Workers integration.

**Cons**: No lifecycle rules (yet), no versioning, smaller ecosystem than S3, less granular IAM.

**Best for**: Read-heavy apps (media, downloads, user content), migrating from S3 to cut costs, apps already on Cloudflare.

### Cloudinary — When to Use

**Pros**: Automatic format selection (WebP/AVIF), on-the-fly transforms via URL, AI-powered cropping, video transcoding, face detection, background removal.

**Cons**: Can get expensive at scale, vendor lock-in on transform URLs, less control over storage details.

**Best for**: Image-heavy apps (e-commerce, social, CMS), apps needing responsive images without server-side processing, video platforms.

### Vercel Blob — When to Use

**Pros**: Simplest DX for Next.js, edge delivery, client-side upload support, zero config.

**Cons**: 500MB max file size, limited features vs S3, Vercel-specific.

**Best for**: Next.js projects, simple file storage needs, prototypes.

## Migration Guide: S3 → R2

```typescript
// 1. Same SDK, different config
const s3 = new S3Client({
  region: 'auto',  // Changed from 'us-east-1'
  endpoint: `https://${CF_ACCOUNT_ID}.r2.cloudflarestorage.com`,  // Added
  credentials: { /* R2 API tokens */ },
});

// 2. Copy objects with rclone
// rclone copy s3:my-bucket r2:my-bucket --progress

// 3. Dual-write during migration
async function upload(key: string, body: Buffer, contentType: string) {
  await Promise.all([
    s3Client.send(new PutObjectCommand({ Bucket: S3_BUCKET, Key: key, Body: body, ContentType: contentType })),
    r2Client.send(new PutObjectCommand({ Bucket: R2_BUCKET, Key: key, Body: body, ContentType: contentType })),
  ]);
}

// 4. Switch reads to R2 after sync is confirmed
// 5. Remove S3 writes after monitoring period
```

## Security Checklist

- [ ] Bucket is private by default (no public access unless intentional)
- [ ] CORS restricted to your domain(s)
- [ ] Presigned URLs have short expiry (15 min for uploads, 1 hour for downloads)
- [ ] File type validation on both client and server (magic bytes, not just extension)
- [ ] File size limits enforced
- [ ] Storage keys use UUIDs (not user-provided filenames)
- [ ] Incomplete multipart uploads auto-cleaned (lifecycle rule)
- [ ] CDN configured for public assets
- [ ] Access logs enabled for audit trail
- [ ] Encryption at rest enabled
- [ ] IAM/API tokens have minimal permissions (principle of least privilege)
