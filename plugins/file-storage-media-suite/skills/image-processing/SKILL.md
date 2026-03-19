---
name: image-processing
description: >
  Process and optimize images with Sharp — resize, crop, convert formats,
  generate thumbnails, add watermarks, and build responsive image sets.
  Covers Sharp pipelines, Cloudinary transforms, Next.js Image optimization,
  and batch processing.
  Triggers: "image processing", "resize image", "image optimization",
  "thumbnail generation", "sharp", "image pipeline", "responsive images",
  "webp conversion", "image compression", "watermark".
  NOT for: file upload UI (use file-uploads) or storage setup (use cloud-storage).
version: 1.0.0
argument-hint: "[resize|thumbnail|optimize|watermark|responsive]"
allowed-tools: Read, Grep, Glob, Write, Edit, Bash
---

# Image Processing

Process and optimize images with Sharp for web delivery.

## Setup

```bash
npm install sharp
npm install -D @types/sharp
```

## Core Operations

### Resize & Convert

```typescript
import sharp from 'sharp';

// Resize to width, maintain aspect ratio
async function resizeImage(input: Buffer, width: number): Promise<Buffer> {
  return sharp(input)
    .resize(width)
    .webp({ quality: 80 })
    .toBuffer();
}

// Resize with specific dimensions and fit
async function resizeExact(input: Buffer, width: number, height: number): Promise<Buffer> {
  return sharp(input)
    .resize(width, height, {
      fit: 'cover',          // cover | contain | fill | inside | outside
      position: 'attention', // Smart crop (focus on interesting areas)
    })
    .webp({ quality: 80 })
    .toBuffer();
}

// Convert format
async function convertToWebP(input: Buffer, quality = 80): Promise<Buffer> {
  return sharp(input).webp({ quality }).toBuffer();
}

async function convertToAVIF(input: Buffer, quality = 50): Promise<Buffer> {
  return sharp(input).avif({ quality }).toBuffer();
}
```

### Thumbnail Generation

```typescript
interface ThumbnailConfig {
  width: number;
  height: number;
  quality: number;
  suffix: string;
}

const THUMBNAIL_SIZES: ThumbnailConfig[] = [
  { width: 150, height: 150, quality: 70, suffix: 'thumb' },
  { width: 400, height: 300, quality: 75, suffix: 'small' },
  { width: 800, height: 600, quality: 80, suffix: 'medium' },
  { width: 1200, height: 900, quality: 85, suffix: 'large' },
];

async function generateThumbnails(
  input: Buffer,
  baseName: string
): Promise<Array<{ key: string; buffer: Buffer; width: number }>> {
  return Promise.all(
    THUMBNAIL_SIZES.map(async (config) => {
      const buffer = await sharp(input)
        .resize(config.width, config.height, {
          fit: 'cover',
          position: 'attention',
        })
        .webp({ quality: config.quality })
        .toBuffer();

      return {
        key: `${baseName}-${config.suffix}.webp`,
        buffer,
        width: config.width,
      };
    })
  );
}
```

### Avatar Processing

```typescript
async function processAvatar(input: Buffer): Promise<Buffer> {
  return sharp(input)
    .resize(256, 256, {
      fit: 'cover',
      position: 'attention',
    })
    // Round corners (circle crop)
    .composite([{
      input: Buffer.from(
        `<svg width="256" height="256">
          <circle cx="128" cy="128" r="128" fill="white"/>
        </svg>`
      ),
      blend: 'dest-in',
    }])
    .webp({ quality: 80 })
    .toBuffer();
}
```

### Watermark

```typescript
async function addWatermark(
  image: Buffer,
  watermarkPath: string,
  position: 'southeast' | 'center' = 'southeast'
): Promise<Buffer> {
  const watermark = await sharp(watermarkPath)
    .resize(200) // Watermark max width
    .ensureAlpha(0.5) // 50% opacity
    .toBuffer();

  return sharp(image)
    .composite([{
      input: watermark,
      gravity: position,
    }])
    .toBuffer();
}

// Text watermark
async function addTextWatermark(image: Buffer, text: string): Promise<Buffer> {
  const metadata = await sharp(image).metadata();
  const width = metadata.width || 800;

  const svgText = Buffer.from(`
    <svg width="${width}" height="40">
      <text x="${width - 10}" y="30" text-anchor="end"
        font-family="Arial" font-size="16" fill="rgba(255,255,255,0.5)">
        ${text}
      </text>
    </svg>
  `);

  return sharp(image)
    .composite([{ input: svgText, gravity: 'southeast' }])
    .toBuffer();
}
```

### Image Metadata

```typescript
async function getImageInfo(input: Buffer) {
  const metadata = await sharp(input).metadata();
  return {
    width: metadata.width,
    height: metadata.height,
    format: metadata.format,
    size: metadata.size,
    hasAlpha: metadata.hasAlpha,
    orientation: metadata.orientation,
    space: metadata.space, // sRGB, etc.
  };
}

// Strip EXIF data (privacy — removes GPS, camera info)
async function stripMetadata(input: Buffer): Promise<Buffer> {
  return sharp(input)
    .rotate() // Auto-rotate based on EXIF before stripping
    .withMetadata({ orientation: undefined })
    .toBuffer();
}
```

## Responsive Image Pipeline

```typescript
interface ResponsiveSet {
  srcSet: string;
  sizes: string;
  src: string;      // Fallback
  width: number;
  height: number;
}

const BREAKPOINTS = [400, 800, 1200, 1600, 2000];

async function generateResponsiveSet(
  input: Buffer,
  basePath: string,
  uploadFn: (key: string, buffer: Buffer, type: string) => Promise<string>
): Promise<ResponsiveSet> {
  const metadata = await sharp(input).metadata();
  const aspectRatio = (metadata.height || 1) / (metadata.width || 1);

  // Generate each breakpoint in WebP
  const variants = await Promise.all(
    BREAKPOINTS
      .filter(w => w <= (metadata.width || 0) * 1.5)  // Don't upscale
      .map(async (width) => {
        const height = Math.round(width * aspectRatio);
        const buffer = await sharp(input)
          .resize(width, height, { fit: 'cover' })
          .webp({ quality: 80 })
          .toBuffer();

        const key = `${basePath}-${width}w.webp`;
        const url = await uploadFn(key, buffer, 'image/webp');
        return { width, url };
      })
  );

  // Fallback JPEG
  const fallbackBuffer = await sharp(input)
    .resize(800)
    .jpeg({ quality: 80 })
    .toBuffer();
  const fallbackUrl = await uploadFn(`${basePath}-800w.jpg`, fallbackBuffer, 'image/jpeg');

  return {
    srcSet: variants.map(v => `${v.url} ${v.width}w`).join(', '),
    sizes: '(max-width: 640px) 100vw, (max-width: 1024px) 50vw, 33vw',
    src: fallbackUrl,
    width: metadata.width || 0,
    height: metadata.height || 0,
  };
}
```

## Processing Pipeline (Background Jobs)

```typescript
// workers/image-processor.ts
import { Queue, Worker } from 'bullmq';
import sharp from 'sharp';

const imageQueue = new Queue('image-processing', {
  connection: { url: process.env.REDIS_URL },
});

// Queue a job
export async function queueImageProcessing(fileId: string, key: string) {
  await imageQueue.add('process', { fileId, key }, {
    attempts: 3,
    backoff: { type: 'exponential', delay: 2000 },
  });
}

// Worker processes jobs
const worker = new Worker('image-processing', async (job) => {
  const { fileId, key } = job.data;

  // 1. Download original from storage
  const original = await downloadFromStorage(key);

  // 2. Generate variants
  const [thumb, medium, large, webp, avif] = await Promise.all([
    sharp(original).resize(150, 150, { fit: 'cover' }).webp({ quality: 70 }).toBuffer(),
    sharp(original).resize(800).webp({ quality: 80 }).toBuffer(),
    sharp(original).resize(1600).webp({ quality: 85 }).toBuffer(),
    sharp(original).resize(1200).webp({ quality: 80 }).toBuffer(),
    sharp(original).resize(1200).avif({ quality: 50 }).toBuffer(),
  ]);

  // 3. Upload all variants
  const baseKey = key.replace(/\.[^.]+$/, '');
  await Promise.all([
    uploadToStorage(`${baseKey}-thumb.webp`, thumb, 'image/webp'),
    uploadToStorage(`${baseKey}-medium.webp`, medium, 'image/webp'),
    uploadToStorage(`${baseKey}-large.webp`, large, 'image/webp'),
    uploadToStorage(`${baseKey}-1200.webp`, webp, 'image/webp'),
    uploadToStorage(`${baseKey}-1200.avif`, avif, 'image/avif'),
  ]);

  // 4. Update database
  await prisma.file.update({
    where: { id: fileId },
    data: {
      status: 'READY',
      metadata: {
        variants: {
          thumb: `${baseKey}-thumb.webp`,
          medium: `${baseKey}-medium.webp`,
          large: `${baseKey}-large.webp`,
        },
      },
    },
  });
}, { connection: { url: process.env.REDIS_URL } });
```

## Cloudinary Transforms (Alternative)

```typescript
// URL-based transforms — no server-side processing needed
import { v2 as cloudinary } from 'cloudinary';

// Auto-optimize
const url = cloudinary.url('photo.jpg', {
  quality: 'auto',
  fetch_format: 'auto',
  width: 800,
  crop: 'scale',
});
// → https://res.cloudinary.com/demo/image/upload/q_auto,f_auto,w_800/photo.jpg

// Face-centered crop for avatars
const avatarUrl = cloudinary.url('photo.jpg', {
  width: 200, height: 200,
  gravity: 'face',
  crop: 'fill',
  radius: 'max',  // Circle crop
});

// Background removal
const noBgUrl = cloudinary.url('photo.jpg', {
  effect: 'background_removal',
});

// Responsive breakpoints (auto-detect optimal sizes)
cloudinary.uploader.upload('photo.jpg', {
  responsive_breakpoints: [{
    create_derived: true,
    bytes_step: 20000,
    min_width: 200,
    max_width: 1600,
  }],
});
```

## Next.js Image Component

```tsx
// Already optimized — Next.js handles format conversion, resizing, and caching
import Image from 'next/image';

// Local image
<Image src="/hero.jpg" width={1200} height={600} alt="Hero" priority />

// Remote image (requires next.config.js domains)
<Image src="https://cdn.example.com/photo.jpg" width={800} height={400} alt="Photo" />

// Fill container
<div className="relative aspect-video">
  <Image src="/photo.jpg" fill className="object-cover" alt="Photo" sizes="100vw" />
</div>

// next.config.js
module.exports = {
  images: {
    remotePatterns: [
      { protocol: 'https', hostname: 'cdn.example.com' },
      { protocol: 'https', hostname: 'res.cloudinary.com' },
    ],
    formats: ['image/avif', 'image/webp'],
  },
};
```

## Blur Placeholder (LQIP)

```typescript
// Generate low-quality image placeholder for progressive loading
async function generateBlurPlaceholder(input: Buffer): Promise<string> {
  const blurred = await sharp(input)
    .resize(20) // Tiny size
    .blur()
    .webp({ quality: 20 })
    .toBuffer();

  return `data:image/webp;base64,${blurred.toString('base64')}`;
}

// Usage with Next.js Image
<Image
  src="/photo.jpg"
  width={800}
  height={400}
  alt="Photo"
  placeholder="blur"
  blurDataURL={blurPlaceholder}
/>

// Usage with regular img (CSS blur technique)
<div className="relative">
  <img
    src={blurPlaceholder}
    className="absolute inset-0 w-full h-full object-cover blur-lg scale-110"
    aria-hidden="true"
  />
  <img
    src="/photo-800.webp"
    className="relative w-full"
    loading="lazy"
    onLoad={(e) => e.currentTarget.previousElementSibling?.remove()}
    alt="Photo"
  />
</div>
```

## Best Practices

1. **Always convert to WebP** — 25-35% smaller than JPEG at similar quality
2. **Generate AVIF too** — even better compression, growing browser support
3. **Process asynchronously** — use BullMQ/Redis for background jobs
4. **Strip EXIF data** — protects user privacy (GPS coordinates, camera info)
5. **Auto-rotate first** — apply EXIF orientation before stripping metadata
6. **Don't upscale** — skip breakpoints larger than the original image
7. **Lazy load below the fold** — `loading="lazy"` on all non-hero images
8. **Use blur placeholders** — LQIP gives instant visual feedback
9. **Set explicit width/height** — prevents layout shift (CLS)
10. **Sharp in production** — runs native code, 10-100x faster than JS alternatives
