---
name: cloud-storage
description: >
  Set up cloud storage with S3, Cloudflare R2, Cloudinary, or Vercel Blob.
  Covers bucket creation, SDK setup, CORS, lifecycle rules, CDN delivery,
  and access control. Provider-agnostic patterns with easy switching.
  Triggers: "cloud storage", "s3 setup", "r2 setup", "cloudinary setup",
  "file storage", "bucket", "cdn", "object storage", "blob storage".
  NOT for: file upload UI (use file-uploads) or image transforms (use image-processing).
version: 1.0.0
argument-hint: "[s3|r2|cloudinary|vercel-blob]"
allowed-tools: Read, Grep, Glob, Write, Edit, Bash
---

# Cloud Storage

Set up cloud storage for your application with the right provider.

## Provider Selection

| Provider | Best For | Egress | Free Tier |
|----------|----------|--------|-----------|
| **AWS S3** | Industry standard, full ecosystem | $0.09/GB | 5GB/12mo |
| **Cloudflare R2** | Zero egress costs | **Free** | 10GB |
| **Cloudinary** | Image/video transforms included | Included | 25K transforms |
| **Vercel Blob** | Next.js projects, edge | Included | 1GB |
| **Supabase Storage** | Supabase projects | $0.09/GB | 1GB |

**Decision shortcut**:
- Serving lots of files to users? → **R2** (zero egress)
- Need image transforms? → **Cloudinary**
- Building with Next.js? → **Vercel Blob**
- Enterprise / need full AWS? → **S3**

## AWS S3

### Setup

```bash
npm install @aws-sdk/client-s3 @aws-sdk/s3-request-presigner
```

```typescript
// lib/storage.ts
import {
  S3Client,
  PutObjectCommand,
  GetObjectCommand,
  DeleteObjectCommand,
  ListObjectsV2Command,
} from '@aws-sdk/client-s3';
import { getSignedUrl } from '@aws-sdk/s3-request-presigner';

const s3 = new S3Client({
  region: process.env.AWS_REGION!,
  credentials: {
    accessKeyId: process.env.AWS_ACCESS_KEY_ID!,
    secretAccessKey: process.env.AWS_SECRET_ACCESS_KEY!,
  },
});

const BUCKET = process.env.S3_BUCKET!;

// Upload a file
export async function uploadFile(key: string, body: Buffer, contentType: string) {
  await s3.send(new PutObjectCommand({
    Bucket: BUCKET,
    Key: key,
    Body: body,
    ContentType: contentType,
    CacheControl: 'public, max-age=31536000, immutable',
  }));
  return `https://${BUCKET}.s3.${process.env.AWS_REGION}.amazonaws.com/${key}`;
}

// Generate presigned upload URL (client uploads directly to S3)
export async function getUploadUrl(key: string, contentType: string, expiresIn = 900) {
  const command = new PutObjectCommand({
    Bucket: BUCKET,
    Key: key,
    ContentType: contentType,
  });
  return getSignedUrl(s3, command, { expiresIn });
}

// Generate presigned download URL (time-limited access to private files)
export async function getDownloadUrl(key: string, expiresIn = 3600) {
  const command = new GetObjectCommand({ Bucket: BUCKET, Key: key });
  return getSignedUrl(s3, command, { expiresIn });
}

// Delete a file
export async function deleteFile(key: string) {
  await s3.send(new DeleteObjectCommand({ Bucket: BUCKET, Key: key }));
}

// List files with prefix
export async function listFiles(prefix: string, maxKeys = 100) {
  const result = await s3.send(new ListObjectsV2Command({
    Bucket: BUCKET,
    Prefix: prefix,
    MaxKeys: maxKeys,
  }));
  return result.Contents?.map(obj => ({
    key: obj.Key!,
    size: obj.Size!,
    lastModified: obj.LastModified!,
  })) || [];
}
```

### S3 Bucket CORS Configuration

```json
[
  {
    "AllowedHeaders": ["*"],
    "AllowedMethods": ["GET", "PUT", "POST"],
    "AllowedOrigins": ["https://yourdomain.com"],
    "ExposeHeaders": ["ETag"],
    "MaxAgeSeconds": 3600
  }
]
```

### S3 Lifecycle Rules

```json
{
  "Rules": [
    {
      "ID": "cleanup-incomplete-uploads",
      "Status": "Enabled",
      "Filter": {},
      "AbortIncompleteMultipartUpload": {
        "DaysAfterInitiation": 7
      }
    },
    {
      "ID": "archive-old-files",
      "Status": "Enabled",
      "Filter": { "Prefix": "uploads/" },
      "Transitions": [
        { "Days": 90, "StorageClass": "STANDARD_IA" },
        { "Days": 365, "StorageClass": "GLACIER" }
      ]
    }
  ]
}
```

## Cloudflare R2

### Setup

```bash
npm install @aws-sdk/client-s3 @aws-sdk/s3-request-presigner
# R2 uses the S3-compatible API — same SDK!
```

```typescript
// lib/r2-storage.ts
import { S3Client, PutObjectCommand } from '@aws-sdk/client-s3';

const r2 = new S3Client({
  region: 'auto',
  endpoint: `https://${process.env.CF_ACCOUNT_ID}.r2.cloudflarestorage.com`,
  credentials: {
    accessKeyId: process.env.R2_ACCESS_KEY_ID!,
    secretAccessKey: process.env.R2_SECRET_ACCESS_KEY!,
  },
});

const BUCKET = process.env.R2_BUCKET!;

export async function uploadToR2(key: string, body: Buffer, contentType: string) {
  await r2.send(new PutObjectCommand({
    Bucket: BUCKET,
    Key: key,
    Body: body,
    ContentType: contentType,
  }));

  // R2 public URL (requires public bucket or custom domain)
  return `https://${process.env.R2_PUBLIC_DOMAIN}/${key}`;
}
```

### R2 Public Access

```
Option A: R2 Custom Domain (recommended)
  → Cloudflare Dashboard → R2 → Settings → Public Access → Custom Domain

Option B: R2.dev subdomain (quick testing)
  → Enable "Allow public access" in bucket settings
  → URL: https://pub-{hash}.r2.dev/{key}
```

## Cloudinary

### Setup

```bash
npm install cloudinary
```

```typescript
// lib/cloudinary.ts
import { v2 as cloudinary } from 'cloudinary';

cloudinary.config({
  cloud_name: process.env.CLOUDINARY_CLOUD_NAME!,
  api_key: process.env.CLOUDINARY_API_KEY!,
  api_secret: process.env.CLOUDINARY_API_SECRET!,
});

// Upload from buffer
export async function uploadImage(buffer: Buffer, folder: string, options?: any) {
  return new Promise<any>((resolve, reject) => {
    const stream = cloudinary.uploader.upload_stream(
      {
        folder,
        resource_type: 'auto',
        transformation: [
          { quality: 'auto', fetch_format: 'auto' },
        ],
        ...options,
      },
      (error, result) => error ? reject(error) : resolve(result!)
    );
    stream.end(buffer);
  });
}

// Upload from URL
export async function uploadFromUrl(url: string, folder: string) {
  return cloudinary.uploader.upload(url, {
    folder,
    resource_type: 'auto',
    quality: 'auto',
    fetch_format: 'auto',
  });
}

// Delete
export async function deleteImage(publicId: string) {
  return cloudinary.uploader.destroy(publicId);
}

// Generate optimized URL
export function getOptimizedUrl(publicId: string, width: number, height?: number) {
  return cloudinary.url(publicId, {
    width,
    height,
    crop: 'fill',
    gravity: 'auto',
    quality: 'auto',
    fetch_format: 'auto',
  });
}

// Generate responsive srcSet
export function getResponsiveSrcSet(publicId: string, widths = [400, 800, 1200, 1600]) {
  return widths.map(w => ({
    width: w,
    url: cloudinary.url(publicId, {
      width: w,
      crop: 'scale',
      quality: 'auto',
      fetch_format: 'auto',
    }),
  }));
}
```

## Vercel Blob

### Setup

```bash
npm install @vercel/blob
```

```typescript
// app/api/upload/route.ts (Next.js App Router)
import { put, del, list } from '@vercel/blob';
import { NextResponse } from 'next/server';

export async function POST(request: Request) {
  const formData = await request.formData();
  const file = formData.get('file') as File;

  if (!file) {
    return NextResponse.json({ error: 'No file' }, { status: 400 });
  }

  const blob = await put(file.name, file, {
    access: 'public',
    addRandomSuffix: true,
  });

  return NextResponse.json(blob);
}

// Client-side upload (bypasses server with token)
import { upload } from '@vercel/blob/client';

async function clientUpload(file: File) {
  const blob = await upload(file.name, file, {
    access: 'public',
    handleUploadUrl: '/api/upload/token', // Your token endpoint
  });
  return blob.url;
}
```

## Provider-Agnostic Storage Interface

```typescript
// lib/storage-interface.ts
interface StorageProvider {
  upload(key: string, body: Buffer, contentType: string): Promise<string>;
  delete(key: string): Promise<void>;
  getUrl(key: string): string;
  getSignedUrl(key: string, expiresIn?: number): Promise<string>;
}

// Factory
function createStorage(provider: 's3' | 'r2' | 'cloudinary'): StorageProvider {
  switch (provider) {
    case 's3': return new S3Storage();
    case 'r2': return new R2Storage();
    case 'cloudinary': return new CloudinaryStorage();
  }
}

// Usage — switch providers by changing one env var
const storage = createStorage(process.env.STORAGE_PROVIDER as any);
const url = await storage.upload('photos/abc.jpg', buffer, 'image/jpeg');
```

## CDN Configuration

### CloudFront + S3

```typescript
// Environment variables
// CDN_URL=https://d123456.cloudfront.net
// S3_BUCKET=my-bucket

// Upload to S3, serve from CloudFront
const key = `images/${uuid}.webp`;
await uploadFile(key, buffer, 'image/webp');
const publicUrl = `${process.env.CDN_URL}/${key}`;
```

### Cache Headers

```typescript
// Set cache headers on upload
await s3.send(new PutObjectCommand({
  Bucket: BUCKET,
  Key: key,
  Body: body,
  ContentType: contentType,

  // Immutable assets (hashed filenames)
  CacheControl: 'public, max-age=31536000, immutable',

  // OR mutable assets (user avatars)
  // CacheControl: 'public, max-age=3600, stale-while-revalidate=86400',
}));
```

## Database Schema for File Tracking

```prisma
model File {
  id          String   @id @default(uuid())
  userId      String
  key         String   @unique   // Storage key (e.g., "uploads/abc/123.jpg")
  filename    String              // Original filename
  contentType String
  size        Int                 // Bytes
  provider    String   @default("s3")  // s3, r2, cloudinary
  status      FileStatus @default(PENDING)
  url         String?             // Public URL after confirmation
  metadata    Json?               // Provider-specific metadata
  createdAt   DateTime @default(now())
  updatedAt   DateTime @updatedAt

  user        User     @relation(fields: [userId], references: [id])

  @@index([userId])
  @@index([status])
}

enum FileStatus {
  PENDING      // Presigned URL generated, upload not confirmed
  CONFIRMED    // Upload verified in storage
  PROCESSING   // Being transformed/optimized
  READY        // Fully processed, available
  FAILED       // Processing failed
  DELETED      // Soft-deleted
}
```

## Best Practices

1. **Use presigned URLs for uploads > 5MB** — keep large files off your server
2. **Set Cache-Control headers** — `immutable` for hashed filenames, short TTL for mutable
3. **Track files in your database** — don't rely on bucket listings for file metadata
4. **Clean up orphaned files** — cron job to delete PENDING files older than 24h
5. **Use CDN for public files** — CloudFront, R2 custom domain, or Cloudinary
6. **Randomize storage keys** — UUIDs prevent enumeration and collisions
7. **Consider R2 for egress-heavy workloads** — zero egress fees saves real money
8. **Set lifecycle rules** — auto-delete incomplete multipart uploads, archive old files
