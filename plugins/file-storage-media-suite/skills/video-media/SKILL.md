---
name: video-media
description: >
  Handle video uploads, transcoding, streaming, and media delivery. Covers
  FFmpeg processing, HLS/DASH adaptive streaming, video thumbnails, Mux/
  Cloudinary video, and HTML5 video player integration.
  Triggers: "video upload", "video processing", "video transcoding",
  "video streaming", "hls", "adaptive bitrate", "video player",
  "ffmpeg", "mux video", "video thumbnail".
  NOT for: image processing (use image-processing) or file uploads (use file-uploads).
version: 1.0.0
argument-hint: "[ffmpeg|hls|player|mux]"
allowed-tools: Read, Grep, Glob, Write, Edit, Bash
---

# Video & Media

Handle video uploads, processing, and streaming delivery.

## Video Processing with FFmpeg

### Install FFmpeg

```bash
# macOS
brew install ffmpeg

# Ubuntu/Debian
sudo apt install ffmpeg

# Node.js wrapper
npm install fluent-ffmpeg @types/fluent-ffmpeg
```

### fluent-ffmpeg Setup

```typescript
import ffmpeg from 'fluent-ffmpeg';
import path from 'path';

// Set binary paths (if not in PATH)
// ffmpeg.setFfmpegPath('/usr/local/bin/ffmpeg');
// ffmpeg.setFfprobePath('/usr/local/bin/ffprobe');

// Get video info
export function getVideoInfo(filePath: string): Promise<ffmpeg.FfprobeData> {
  return new Promise((resolve, reject) => {
    ffmpeg.ffprobe(filePath, (err, data) => {
      err ? reject(err) : resolve(data);
    });
  });
}

// Usage
const info = await getVideoInfo('input.mp4');
// info.format.duration — length in seconds
// info.streams[0].width, info.streams[0].height — dimensions
// info.format.size — file size in bytes
```

### Web-Optimized MP4

```typescript
export function convertToWebMP4(input: string, output: string): Promise<void> {
  return new Promise((resolve, reject) => {
    ffmpeg(input)
      .videoCodec('libx264')
      .audioCodec('aac')
      .outputOptions([
        '-preset slow',          // Better compression (slower encode)
        '-crf 23',               // Quality (18=high, 23=medium, 28=low)
        '-movflags +faststart',  // Enable progressive download
        '-pix_fmt yuv420p',      // Maximum compatibility
        '-profile:v main',       // Broad device support
        '-level 3.1',
        '-maxrate 4M',
        '-bufsize 8M',
        '-b:a 128k',
      ])
      .on('progress', (progress) => {
        console.log(`Processing: ${progress.percent?.toFixed(1)}%`);
      })
      .on('error', reject)
      .on('end', () => resolve())
      .save(output);
  });
}
```

### Generate Thumbnails

```typescript
export function generateThumbnails(
  input: string,
  outputDir: string,
  count = 4
): Promise<string[]> {
  return new Promise((resolve, reject) => {
    const filenames: string[] = [];

    ffmpeg(input)
      .screenshots({
        count,
        folder: outputDir,
        filename: 'thumb-%i.jpg',
        size: '640x360',
      })
      .on('filenames', (fns) => { filenames.push(...fns); })
      .on('error', reject)
      .on('end', () => resolve(filenames.map(f => path.join(outputDir, f))));
  });
}

// Single thumbnail at specific time
export function generateThumbnailAt(
  input: string,
  output: string,
  timeSeconds: number
): Promise<void> {
  return new Promise((resolve, reject) => {
    ffmpeg(input)
      .screenshots({
        timestamps: [timeSeconds],
        filename: path.basename(output),
        folder: path.dirname(output),
        size: '1280x720',
      })
      .on('error', reject)
      .on('end', () => resolve());
  });
}
```

### Create GIF Preview

```typescript
export function createGifPreview(
  input: string,
  output: string,
  startTime = 0,
  duration = 3
): Promise<void> {
  return new Promise((resolve, reject) => {
    ffmpeg(input)
      .setStartTime(startTime)
      .setDuration(duration)
      .outputOptions([
        '-vf', 'fps=10,scale=320:-1:flags=lanczos',
        '-loop', '0',
      ])
      .on('error', reject)
      .on('end', () => resolve())
      .save(output);
  });
}
```

## HLS Adaptive Bitrate Streaming

### Generate HLS with Multiple Qualities

```typescript
interface HLSVariant {
  name: string;
  width: number;
  height: number;
  videoBitrate: string;
  audioBitrate: string;
}

const HLS_VARIANTS: HLSVariant[] = [
  { name: '360p', width: 640, height: 360, videoBitrate: '800k', audioBitrate: '96k' },
  { name: '480p', width: 854, height: 480, videoBitrate: '1400k', audioBitrate: '128k' },
  { name: '720p', width: 1280, height: 720, videoBitrate: '2800k', audioBitrate: '128k' },
  { name: '1080p', width: 1920, height: 1080, videoBitrate: '5000k', audioBitrate: '192k' },
];

export async function generateHLS(input: string, outputDir: string): Promise<string> {
  // Generate each quality variant
  for (const variant of HLS_VARIANTS) {
    await new Promise<void>((resolve, reject) => {
      ffmpeg(input)
        .videoCodec('libx264')
        .audioCodec('aac')
        .outputOptions([
          `-vf scale=${variant.width}:${variant.height}`,
          `-b:v ${variant.videoBitrate}`,
          `-maxrate ${variant.videoBitrate}`,
          `-bufsize ${parseInt(variant.videoBitrate) * 2}k`,
          `-b:a ${variant.audioBitrate}`,
          '-preset fast',
          '-g 48',           // Keyframe interval (2 sec at 24fps)
          '-keyint_min 48',
          '-sc_threshold 0',
          '-hls_time 6',     // Segment duration in seconds
          '-hls_list_size 0',
          '-f hls',
        ])
        .on('error', reject)
        .on('end', () => resolve())
        .save(path.join(outputDir, `${variant.name}.m3u8`));
    });
  }

  // Generate master playlist
  const masterPlaylist = [
    '#EXTM3U',
    '#EXT-X-VERSION:3',
    ...HLS_VARIANTS.map(v =>
      `#EXT-X-STREAM-INF:BANDWIDTH=${parseInt(v.videoBitrate) * 1000},RESOLUTION=${v.width}x${v.height}\n${v.name}.m3u8`
    ),
  ].join('\n');

  const masterPath = path.join(outputDir, 'master.m3u8');
  await writeFile(masterPath, masterPlaylist);

  return masterPath;
}
```

### Serve HLS with Express

```typescript
import express from 'express';
import path from 'path';

const app = express();

// Serve HLS files with proper headers
app.use('/videos', express.static(path.join(__dirname, 'hls-output'), {
  setHeaders: (res, filePath) => {
    if (filePath.endsWith('.m3u8')) {
      res.setHeader('Content-Type', 'application/vnd.apple.mpegurl');
      res.setHeader('Cache-Control', 'no-cache'); // Playlists can change
    } else if (filePath.endsWith('.ts')) {
      res.setHeader('Content-Type', 'video/mp2t');
      res.setHeader('Cache-Control', 'public, max-age=31536000'); // Segments are immutable
    }
    // CORS for cross-origin video players
    res.setHeader('Access-Control-Allow-Origin', '*');
  },
}));
```

## HTML5 Video Player

### Basic Player with HLS.js

```bash
npm install hls.js
```

```tsx
// components/VideoPlayer.tsx
'use client';
import { useEffect, useRef, useState } from 'react';
import Hls from 'hls.js';

interface Props {
  src: string;           // HLS manifest URL or direct video URL
  poster?: string;       // Thumbnail image
  autoPlay?: boolean;
  className?: string;
}

export function VideoPlayer({ src, poster, autoPlay = false, className }: Props) {
  const videoRef = useRef<HTMLVideoElement>(null);
  const [quality, setQuality] = useState<string>('auto');
  const hlsRef = useRef<Hls | null>(null);

  useEffect(() => {
    const video = videoRef.current;
    if (!video) return;

    if (src.endsWith('.m3u8') && Hls.isSupported()) {
      // HLS.js for non-Safari browsers
      const hls = new Hls({
        maxLoadingDelay: 4,
        maxBufferLength: 30,
        liveSyncDurationCount: 3,
      });

      hls.loadSource(src);
      hls.attachMedia(video);
      hls.on(Hls.Events.MANIFEST_PARSED, () => {
        if (autoPlay) video.play().catch(() => {});
      });

      hlsRef.current = hls;

      return () => {
        hls.destroy();
        hlsRef.current = null;
      };
    } else if (video.canPlayType('application/vnd.apple.mpegurl')) {
      // Native HLS (Safari)
      video.src = src;
      if (autoPlay) video.play().catch(() => {});
    } else {
      // Direct video file
      video.src = src;
      if (autoPlay) video.play().catch(() => {});
    }
  }, [src, autoPlay]);

  return (
    <div className={`relative group ${className}`}>
      <video
        ref={videoRef}
        poster={poster}
        controls
        playsInline
        preload="metadata"
        className="w-full rounded-lg"
      />
    </div>
  );
}
```

### Video Upload + Processing Pipeline

```typescript
// Complete video upload workflow
import { Queue, Worker } from 'bullmq';

// 1. Upload endpoint accepts video
app.post('/api/videos/upload', upload.single('video'), async (req, res) => {
  const file = req.file!;

  // Create video record
  const video = await prisma.video.create({
    data: {
      userId: req.user!.id,
      title: req.body.title,
      originalKey: `originals/${randomUUID()}-${file.originalname}`,
      status: 'UPLOADING',
      size: file.size,
    },
  });

  // Upload original to S3
  await uploadFile(video.originalKey, file.buffer, file.mimetype);

  // Queue processing job
  await videoQueue.add('process', { videoId: video.id });

  // Update status
  await prisma.video.update({
    where: { id: video.id },
    data: { status: 'PROCESSING' },
  });

  res.json({ videoId: video.id, status: 'PROCESSING' });
});

// 2. Background worker processes video
const videoQueue = new Queue('video-processing', { connection: redisConfig });

const worker = new Worker('video-processing', async (job) => {
  const { videoId } = job.data;
  const video = await prisma.video.findUnique({ where: { id: videoId } });
  if (!video) return;

  const tmpDir = `/tmp/video-${videoId}`;
  await mkdir(tmpDir, { recursive: true });

  try {
    // Download original
    const originalPath = path.join(tmpDir, 'original.mp4');
    await downloadFile(video.originalKey, originalPath);

    // Get video info
    const info = await getVideoInfo(originalPath);
    const duration = info.format.duration || 0;

    // Generate thumbnail
    const thumbPath = path.join(tmpDir, 'thumb.jpg');
    await generateThumbnailAt(originalPath, thumbPath, Math.min(duration * 0.1, 5));
    const thumbKey = `thumbnails/${videoId}.jpg`;
    await uploadFile(thumbKey, await readFile(thumbPath), 'image/jpeg');

    // Generate HLS variants
    const hlsDir = path.join(tmpDir, 'hls');
    await mkdir(hlsDir, { recursive: true });
    await generateHLS(originalPath, hlsDir);

    // Upload HLS files to S3
    const hlsFiles = await glob(`${hlsDir}/**/*`);
    for (const file of hlsFiles) {
      const key = `hls/${videoId}/${path.relative(hlsDir, file)}`;
      const contentType = file.endsWith('.m3u8')
        ? 'application/vnd.apple.mpegurl'
        : 'video/mp2t';
      await uploadFile(key, await readFile(file), contentType);
    }

    // Update video record
    await prisma.video.update({
      where: { id: videoId },
      data: {
        status: 'READY',
        duration: Math.round(duration),
        thumbnailUrl: `${CDN_URL}/${thumbKey}`,
        hlsUrl: `${CDN_URL}/hls/${videoId}/master.m3u8`,
        width: info.streams[0]?.width,
        height: info.streams[0]?.height,
      },
    });
  } catch (error) {
    await prisma.video.update({
      where: { id: videoId },
      data: { status: 'FAILED' },
    });
    throw error;
  } finally {
    await rm(tmpDir, { recursive: true, force: true });
  }
}, { connection: redisConfig, concurrency: 2 });
```

## Managed Video Services

### Mux

```bash
npm install @mux/mux-node @mux/mux-player-react
```

```typescript
import Mux from '@mux/mux-node';

const mux = new Mux({
  tokenId: process.env.MUX_TOKEN_ID!,
  tokenSecret: process.env.MUX_TOKEN_SECRET!,
});

// Create upload URL (client uploads directly to Mux)
const upload = await mux.video.uploads.create({
  cors_origin: 'https://myapp.com',
  new_asset_settings: {
    playback_policy: ['public'],
    encoding_tier: 'baseline', // or 'smart' for better quality
  },
});

// upload.url — give to client for direct upload
// Mux handles transcoding, HLS, thumbnails, everything

// Get playback ID for player
const asset = await mux.video.assets.retrieve(assetId);
const playbackId = asset.playback_ids?.[0]?.id;
// Stream URL: https://stream.mux.com/{playbackId}.m3u8
// Thumbnail: https://image.mux.com/{playbackId}/thumbnail.jpg
```

```tsx
// React player component
import MuxPlayer from '@mux/mux-player-react';

<MuxPlayer
  playbackId={playbackId}
  metadata={{
    video_title: 'My Video',
    viewer_user_id: userId,
  }}
  autoPlay={false}
  muted
/>
```

### Cloudinary Video

```typescript
import { v2 as cloudinary } from 'cloudinary';

// Upload video with auto-transcoding
const result = await cloudinary.uploader.upload(videoPath, {
  resource_type: 'video',
  folder: 'videos',
  eager: [
    { streaming_profile: 'hd', format: 'm3u8' },       // HLS
    { width: 640, crop: 'scale', format: 'mp4' },       // 640px MP4
    { start_offset: '5', format: 'jpg', crop: 'fill' }, // Thumbnail at 5s
  ],
  eager_async: true,
});

// Video URL with on-the-fly transforms:
// https://res.cloudinary.com/CLOUD/video/upload/w_1280,q_auto/videos/my-video.mp4
```

## Database Schema

```prisma
model Video {
  id           String      @id @default(uuid())
  userId       String
  title        String
  description  String?
  originalKey  String      // S3 key for original upload
  hlsUrl       String?     // HLS manifest URL
  thumbnailUrl String?     // Poster image URL
  duration     Int?        // Seconds
  width        Int?
  height       Int?
  size         Int         // Original file size in bytes
  status       VideoStatus @default(UPLOADING)
  visibility   Visibility  @default(PRIVATE)
  createdAt    DateTime    @default(now())
  updatedAt    DateTime    @updatedAt

  user         User        @relation(fields: [userId], references: [id])

  @@index([userId])
  @@index([status])
}

enum VideoStatus {
  UPLOADING
  PROCESSING
  READY
  FAILED
}

enum Visibility {
  PUBLIC
  PRIVATE
  UNLISTED
}
```

## Best Practices

1. **Never process video synchronously** — always use a background job queue
2. **Use presigned URLs for uploads** — video files are large, don't proxy through your API
3. **Generate HLS for adaptive streaming** — users on slow connections get lower quality automatically
4. **`-movflags +faststart`** — always for MP4, enables progressive download
5. **Consider managed services** (Mux, Cloudinary) — self-hosting video transcoding is expensive and complex
6. **Limit upload size** — set reasonable limits (500MB-2GB depending on your use case)
7. **Generate thumbnails at multiple times** — let users pick the best one
8. **Clean up temp files** — always delete local processing artifacts in a `finally` block
9. **Monitor processing costs** — video transcoding is CPU-intensive, right-size your workers
10. **Use `preload="metadata"`** on video elements — loads only dimensions and duration, not the whole file
