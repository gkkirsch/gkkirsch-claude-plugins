# Storage Cost Comparison

Quick reference for cloud storage pricing across providers.

---

## Storage Costs (per GB/month)

| Provider | Standard | Infrequent Access | Archive | Min Duration |
|----------|----------|-------------------|---------|-------------|
| **AWS S3** | $0.023 | $0.0125 | $0.004 (Glacier) | 90 days (IA), 90-180 days (Glacier) |
| **Cloudflare R2** | $0.015 | N/A | N/A | None |
| **Google Cloud** | $0.020 | $0.010 | $0.004 (Archive) | 30 days (Nearline), 365 days (Archive) |
| **Azure Blob** | $0.018 | $0.010 | $0.002 (Archive) | 30 days (Cool), 180 days (Archive) |
| **Backblaze B2** | $0.006 | N/A | N/A | None |
| **Cloudinary** | Included | N/A | N/A | N/A |
| **Vercel Blob** | Included | N/A | N/A | N/A |

## Egress (Data Transfer Out) Costs

| Provider | Per GB | Free Tier | Notes |
|----------|--------|-----------|-------|
| **AWS S3** | $0.09 | 100GB/mo | Through CloudFront: $0.085/GB |
| **Cloudflare R2** | **$0.00** | Unlimited | Zero egress. This is the killer feature. |
| **Google Cloud** | $0.12 | 1GB/mo | |
| **Azure Blob** | $0.087 | 100GB/mo (12mo) | |
| **Backblaze B2** | $0.01 | 1GB/day | Free via Cloudflare (bandwidth alliance) |
| **Cloudinary** | Included | 25GB bandwidth/mo | |
| **Vercel Blob** | Included | 1GB storage | |

## Operations Costs (per 10,000 requests)

| Provider | PUT/POST | GET | DELETE |
|----------|----------|-----|--------|
| **AWS S3** | $0.005 | $0.0004 | Free |
| **Cloudflare R2** | $0.0045 | $0.0036 | Free |
| **Google Cloud** | $0.005 | $0.0004 | Free |
| **Backblaze B2** | Free (2,500/day) | Free (2,500/day) | Free |

## Monthly Cost Scenarios

### Scenario 1: Small App (100GB storage, 500GB egress)

| Provider | Storage | Egress | Operations | Total |
|----------|---------|--------|------------|-------|
| S3 | $2.30 | $45.00 | ~$0.50 | **$47.80** |
| R2 | $1.50 | $0.00 | ~$0.50 | **$2.00** |
| B2 + CF | $0.60 | $0.00 | ~$0.00 | **$0.60** |

### Scenario 2: Medium App (1TB storage, 5TB egress)

| Provider | Storage | Egress | Operations | Total |
|----------|---------|--------|------------|-------|
| S3 | $23.00 | $450.00 | ~$5.00 | **$478.00** |
| R2 | $15.00 | $0.00 | ~$5.00 | **$20.00** |
| B2 + CF | $6.00 | $0.00 | ~$0.00 | **$6.00** |

### Scenario 3: Large App (10TB storage, 50TB egress)

| Provider | Storage | Egress | Operations | Total |
|----------|---------|--------|------------|-------|
| S3 | $230.00 | $4,500.00 | ~$50.00 | **$4,780.00** |
| R2 | $150.00 | $0.00 | ~$50.00 | **$200.00** |

**Key insight**: At scale, egress costs dominate. R2's zero egress makes it 10-20x cheaper for read-heavy workloads.

## Free Tiers Summary

| Provider | Storage | Bandwidth | Duration |
|----------|---------|-----------|----------|
| **AWS S3** | 5GB | 100GB/mo | 12 months |
| **Cloudflare R2** | 10GB | Unlimited | Forever |
| **Google Cloud** | 5GB | 1GB/mo | 12 months (storage), always (egress) |
| **Cloudinary** | 25 credits | 25GB | Forever |
| **Vercel Blob** | 1GB | Included | Forever |
| **Supabase** | 1GB | 2GB | Forever |
| **Backblaze B2** | 10GB | 1GB/day | Forever |

## CDN Pricing

| CDN | Per GB | Free Tier | Notes |
|-----|--------|-----------|-------|
| **Cloudflare** | $0.00 | Unlimited | Free plan includes unlimited bandwidth |
| **CloudFront** | $0.085 | 1TB/mo (12mo) | Decreases with volume |
| **Fastly** | $0.12 | None | |
| **Bunny CDN** | $0.01 | 14-day trial | Very affordable |

## Image Transform Pricing

| Service | Free Tier | Paid | Best For |
|---------|-----------|------|----------|
| **Cloudinary** | 25K transforms/mo | $89/mo (25 credits) | Rich transforms, face detection |
| **imgix** | None | $10/mo + $3/1K origins | Performance-focused |
| **Imgproxy** | Open source | Self-hosted | Maximum control, lowest cost |
| **Vercel Image** | Included | Per request ($5/1K) | Next.js native |

## Recommendation Matrix

| Situation | Recommendation |
|-----------|---------------|
| Starting out, small app | Cloudflare R2 (cheap, simple) |
| AWS ecosystem already | S3 + CloudFront |
| Read-heavy (CDN content, media) | R2 (zero egress) |
| Image-heavy (transforms needed) | Cloudinary or R2 + Sharp |
| Video streaming | Mux (managed) or S3 + CloudFront |
| Budget-sensitive backup/archive | Backblaze B2 + Cloudflare |
| Enterprise with compliance needs | S3 (most certifications) |
| Vercel/Next.js project | Vercel Blob (simplest) |
| Supabase project | Supabase Storage (integrated) |
