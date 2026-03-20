# Production Rendering with Remotion

Comprehensive guide to rendering Remotion videos for production — local rendering,
AWS Lambda at scale, encoding settings, multi-format output, and CI/CD pipelines.

---

## Local Rendering

### Basic Render Commands

```bash
# Render a composition to MP4 (H.264 codec, default)
npx remotion render src/index.ts MyVideo out/video.mp4

# Render with explicit codec
npx remotion render src/index.ts MyVideo out/video.mp4 --codec=h264

# Render only a portion (useful for testing)
npx remotion render src/index.ts MyVideo out/video.mp4 --frames=0-90

# Override props at render time
npx remotion render src/index.ts MyVideo out/video.mp4 \
  --props='{"title": "Hello World", "backgroundColor": "#1e40af"}'

# Read props from a JSON file
npx remotion render src/index.ts MyVideo out/video.mp4 \
  --props=./props/demo-config.json
```

### Render a Still Image (Thumbnail)

```bash
# Render a single frame as PNG
npx remotion still src/index.ts MyVideo out/thumbnail.png --frame=45

# Render as JPEG with quality setting
npx remotion still src/index.ts MyVideo out/thumbnail.jpg \
  --frame=45 --image-format=jpeg --jpeg-quality=90

# Render multiple thumbnails for different frames
for frame in 0 30 60 90; do
  npx remotion still src/index.ts MyVideo "out/thumb-${frame}.png" --frame=$frame
done
```

### Resolution Scaling

```bash
# Render at half resolution (faster, good for previews)
npx remotion render src/index.ts MyVideo out/preview.mp4 --scale=0.5

# Render at 4K from a 1080p composition
npx remotion render src/index.ts MyVideo out/4k.mp4 --scale=2
```

---

## Codec and Encoding Settings

### Available Codecs

| Codec | Output | Use Case | Notes |
|-------|--------|----------|-------|
| `h264` | `.mp4` | Default, most compatible | Works everywhere, good compression |
| `h265` | `.mp4` | Smaller files, same quality | Not supported on all browsers |
| `vp8` | `.webm` | Web playback | Good for web, open format |
| `vp9` | `.webm` | Better compression than VP8 | Slower to encode |
| `prores` | `.mov` | Professional editing | Very large files, lossless quality |
| `gif` | `.gif` | Short animations | Very large files, 256 colors max |

### CRF (Constant Rate Factor) — Quality Control

CRF controls the quality-size tradeoff. Lower values = better quality, larger files.

```bash
# High quality (larger file) — good for final delivery
npx remotion render src/index.ts MyVideo out/hq.mp4 --crf=18

# Medium quality (balanced) — default
npx remotion render src/index.ts MyVideo out/med.mp4 --crf=23

# Lower quality (smaller file) — good for social media drafts
npx remotion render src/index.ts MyVideo out/lq.mp4 --crf=28
```

CRF reference for H.264:
- **15-17**: Visually lossless. Use for archival or master copies.
- **18-22**: High quality. Recommended for final delivery.
- **23-28**: Medium quality. Good balance for web distribution.
- **29-35**: Low quality. Noticeable artifacts, but very small files.

### Audio Encoding

```bash
# Mute the video (no audio track)
npx remotion render src/index.ts MyVideo out/video.mp4 --muted

# Explicit audio codec
npx remotion render src/index.ts MyVideo out/video.mp4 --audio-codec=aac

# Audio bitrate (default 128k)
npx remotion render src/index.ts MyVideo out/video.mp4 --audio-bitrate=320k
```

Audio codec options:
- `aac` — Default for MP4. Universal compatibility.
- `opus` — Better compression. Used with WebM.
- `mp3` — Wide compatibility. Use if AAC causes issues.
- `pcm-16` — Uncompressed. Only for ProRes or WAV output.

### Pixel Format

```bash
# Standard (default) — works everywhere
npx remotion render src/index.ts MyVideo out/video.mp4 --pixel-format=yuv420p

# Higher color accuracy — for professional workflows
npx remotion render src/index.ts MyVideo out/video.mp4 --pixel-format=yuv444p
```

---

## Multi-Format Output

### Rendering for Multiple Platforms

Create a script that renders the same video in multiple formats:

```bash
#!/bin/bash
# render-all-formats.sh

COMPOSITION="ProductDemo"
ENTRY="src/index.ts"
OUT_DIR="out"

echo "Rendering MP4 (H.264) for general use..."
npx remotion render $ENTRY $COMPOSITION "$OUT_DIR/video-h264.mp4" \
  --codec=h264 --crf=20

echo "Rendering WebM (VP9) for web..."
npx remotion render $ENTRY $COMPOSITION "$OUT_DIR/video-vp9.webm" \
  --codec=vp9 --crf=28

echo "Rendering GIF for social previews..."
npx remotion render $ENTRY $COMPOSITION "$OUT_DIR/video.gif" \
  --codec=gif --every-nth-frame=2 --scale=0.5

echo "Generating thumbnail..."
npx remotion still $ENTRY $COMPOSITION "$OUT_DIR/thumbnail.png" --frame=45

echo "All formats rendered."
```

### Platform-Specific Recommendations

| Platform | Format | Resolution | Duration | FPS | Notes |
|----------|--------|------------|----------|-----|-------|
| YouTube | MP4 H.264 | 1920x1080 | Any | 30/60 | CRF 18-20 |
| Twitter/X | MP4 H.264 | 1280x720 | < 2:20 | 30 | Max 512MB |
| Instagram Reels | MP4 H.264 | 1080x1920 | < 90s | 30 | Vertical 9:16 |
| TikTok | MP4 H.264 | 1080x1920 | < 10min | 30 | Vertical 9:16 |
| LinkedIn | MP4 H.264 | 1920x1080 | < 10min | 30 | Max 5GB |
| Web embed | WebM VP9 | 1280x720 | Any | 30 | Smaller files |
| Email/Slack | GIF | 640x360 | < 15s | 15 | Use every-nth-frame |

---

## AWS Lambda Rendering

For rendering at scale (many videos in parallel) or offloading from your machine.

### Setup

```bash
# Install Lambda packages
npm install @remotion/lambda @aws-sdk/client-lambda @aws-sdk/client-s3

# Deploy the Lambda function (one-time)
npx remotion lambda sites create src/index.ts --site-name=my-video-site
npx remotion lambda functions deploy
```

### Configuration

```ts
// remotion.config.ts (for Lambda)
import { Config } from '@remotion/cli/config';

Config.setVideoImageFormat('jpeg');
Config.setOverwriteOutput(true);
```

### Render via CLI

```bash
# Render on Lambda
npx remotion lambda render \
  --function-name=remotion-render-2024-mem2048mb-disk2048mb-120sec \
  --serve-url=https://remotionlambda-xxxxx.s3.us-east-1.amazonaws.com/sites/my-video-site/index.html \
  MyVideo

# With custom props
npx remotion lambda render \
  --function-name=remotion-render-2024-mem2048mb-disk2048mb-120sec \
  --serve-url=https://remotionlambda-xxxxx.s3.us-east-1.amazonaws.com/sites/my-video-site/index.html \
  MyVideo \
  --props='{"title":"Custom Title"}'
```

### Programmatic Rendering (Node.js)

```ts
// scripts/render-lambda.ts
import {
  renderMediaOnLambda,
  getRenderProgress,
  getSites,
  getOrCreateBucket,
} from '@remotion/lambda/client';

async function renderVideo(props: Record<string, unknown>) {
  const { bucketName } = await getOrCreateBucket({ region: 'us-east-1' });
  const sites = await getSites({ region: 'us-east-1' });
  const site = sites.sites.find((s) => s.id === 'my-video-site');

  if (!site) throw new Error('Site not found. Deploy first.');

  const { renderId } = await renderMediaOnLambda({
    region: 'us-east-1',
    functionName: 'remotion-render-2024-mem2048mb-disk2048mb-120sec',
    serveUrl: site.serveUrl,
    composition: 'ProductDemo',
    inputProps: props,
    codec: 'h264',
    imageFormat: 'jpeg',
    maxRetries: 1,
    framesPerLambda: 20,
    privacy: 'public',
    outName: `renders/${Date.now()}-product-demo.mp4`,
  });

  console.log(`Render started: ${renderId}`);

  // Poll for completion
  let progress;
  do {
    progress = await getRenderProgress({
      renderId,
      bucketName,
      functionName: 'remotion-render-2024-mem2048mb-disk2048mb-120sec',
      region: 'us-east-1',
    });

    if (progress.fatalErrorEncountered) {
      throw new Error(`Render failed: ${progress.errors[0]?.message}`);
    }

    console.log(`Progress: ${(progress.overallProgress * 100).toFixed(1)}%`);
    await new Promise((r) => setTimeout(r, 2000));
  } while (!progress.done);

  console.log(`Render complete: ${progress.outputFile}`);
  return progress.outputFile;
}

// Usage
renderVideo({
  title: 'Q4 Product Update',
  metrics: { users: 50000, revenue: 120000 },
}).catch(console.error);
```

### Lambda Cost Estimation

- Typical 30-second 1080p video: ~$0.05-0.15
- Rendering is parallelized across multiple Lambda invocations
- Each Lambda processes a chunk of frames, then they are stitched together
- Memory: 2048MB recommended. More memory = faster rendering
- Timeout: 120 seconds per chunk is usually enough

---

## CI/CD Pipeline

### GitHub Actions Workflow

```yaml
# .github/workflows/render-video.yml
name: Render Video

on:
  push:
    branches: [main]
    paths:
      - 'src/**'
      - 'public/**'
  workflow_dispatch:
    inputs:
      composition:
        description: 'Composition to render'
        required: true
        default: 'ProductDemo'
      props:
        description: 'JSON props for the composition'
        required: false
        default: '{}'

jobs:
  render:
    runs-on: ubuntu-latest
    timeout-minutes: 30

    steps:
      - uses: actions/checkout@v4

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: 20
          cache: 'npm'

      - name: Install dependencies
        run: npm ci

      - name: Install Chrome dependencies
        run: npx remotion browser ensure

      - name: Render video
        run: |
          COMPOSITION="${{ github.event.inputs.composition || 'ProductDemo' }}"
          PROPS='${{ github.event.inputs.props || '{}' }}'

          npx remotion render src/index.ts "$COMPOSITION" "out/video.mp4" \
            --codec=h264 \
            --crf=20 \
            --props="$PROPS"

      - name: Generate thumbnail
        run: |
          COMPOSITION="${{ github.event.inputs.composition || 'ProductDemo' }}"

          npx remotion still src/index.ts "$COMPOSITION" "out/thumbnail.png" \
            --frame=45

      - name: Upload artifacts
        uses: actions/upload-artifact@v4
        with:
          name: rendered-video
          path: out/
          retention-days: 30

      - name: Upload to S3 (optional)
        if: github.ref == 'refs/heads/main'
        env:
          AWS_ACCESS_KEY_ID: ${{ secrets.AWS_ACCESS_KEY_ID }}
          AWS_SECRET_ACCESS_KEY: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
          AWS_REGION: us-east-1
        run: |
          aws s3 cp out/video.mp4 s3://${{ secrets.S3_BUCKET }}/videos/latest.mp4 \
            --content-type "video/mp4"
          aws s3 cp out/thumbnail.png s3://${{ secrets.S3_BUCKET }}/videos/latest-thumb.png \
            --content-type "image/png"
```

### Data-Driven Batch Rendering

Render personalized videos from a CSV or API:

```ts
// scripts/batch-render.ts
import { bundle } from '@remotion/bundler';
import { renderMedia, selectComposition } from '@remotion/renderer';
import { readFileSync } from 'fs';
import path from 'path';

interface VideoData {
  customerName: string;
  plan: string;
  usageHours: number;
  outputFile: string;
}

async function batchRender() {
  // Load data
  const data: VideoData[] = JSON.parse(
    readFileSync('data/customers.json', 'utf-8')
  );

  // Bundle the Remotion project once
  console.log('Bundling...');
  const bundleLocation = await bundle({
    entryPoint: path.resolve('./src/index.ts'),
    webpackOverride: (config) => config,
  });

  // Render each video
  for (const item of data) {
    console.log(`Rendering video for ${item.customerName}...`);

    const composition = await selectComposition({
      serveUrl: bundleLocation,
      id: 'PersonalizedReport',
      inputProps: {
        customerName: item.customerName,
        plan: item.plan,
        usageHours: item.usageHours,
      },
    });

    await renderMedia({
      composition,
      serveUrl: bundleLocation,
      codec: 'h264',
      outputLocation: `out/${item.outputFile}`,
      inputProps: {
        customerName: item.customerName,
        plan: item.plan,
        usageHours: item.usageHours,
      },
      onProgress: ({ progress }) => {
        if (Math.round(progress * 100) % 25 === 0) {
          console.log(`  ${item.customerName}: ${Math.round(progress * 100)}%`);
        }
      },
    });
  }

  console.log(`Batch complete. Rendered ${data.length} videos.`);
}

batchRender().catch(console.error);
```

### Programmatic Rendering with Node.js API

```ts
// scripts/render-single.ts
import { bundle } from '@remotion/bundler';
import { renderMedia, selectComposition } from '@remotion/renderer';
import path from 'path';

async function render() {
  // Step 1: Bundle the project
  const bundleLocation = await bundle({
    entryPoint: path.resolve('./src/index.ts'),
    webpackOverride: (config) => config,
  });

  // Step 2: Select the composition and set input props
  const composition = await selectComposition({
    serveUrl: bundleLocation,
    id: 'ProductDemo',
    inputProps: {
      title: 'My Product',
      features: ['Fast', 'Reliable', 'Scalable'],
    },
  });

  // Step 3: Render
  await renderMedia({
    composition,
    serveUrl: bundleLocation,
    codec: 'h264',
    outputLocation: 'out/product-demo.mp4',
    crf: 20,
    inputProps: {
      title: 'My Product',
      features: ['Fast', 'Reliable', 'Scalable'],
    },
    onProgress: ({ progress }) => {
      process.stdout.write(`\rRendering: ${(progress * 100).toFixed(1)}%`);
    },
  });

  console.log('\nDone!');
}

render().catch(console.error);
```

---

## Thumbnail Generation

### CLI Thumbnails

```bash
# Default (first frame)
npx remotion still src/index.ts MyVideo out/thumb.png

# Specific frame
npx remotion still src/index.ts MyVideo out/thumb.png --frame=90

# Custom resolution
npx remotion still src/index.ts MyVideo out/thumb.png --frame=90 --scale=0.5

# JPEG with quality
npx remotion still src/index.ts MyVideo out/thumb.jpg \
  --frame=90 --image-format=jpeg --jpeg-quality=85
```

### Programmatic Thumbnails

```ts
import { bundle } from '@remotion/bundler';
import { renderStill, selectComposition } from '@remotion/renderer';
import path from 'path';

async function generateThumbnail(frame: number) {
  const bundleLocation = await bundle({
    entryPoint: path.resolve('./src/index.ts'),
  });

  const composition = await selectComposition({
    serveUrl: bundleLocation,
    id: 'ProductDemo',
  });

  await renderStill({
    composition,
    serveUrl: bundleLocation,
    output: `out/thumbnail-frame-${frame}.png`,
    frame,
    imageFormat: 'png',
  });

  console.log(`Thumbnail saved for frame ${frame}`);
}

// Generate thumbnails at key moments
[0, 45, 90, 150, 240].forEach(generateThumbnail);
```

---

## Performance Optimization

### Rendering Speed Tips

1. **Use JPEG for intermediate frames** — PNG is lossless but slower. JPEG is fine for most videos.
   ```ts
   Config.setVideoImageFormat('jpeg');
   Config.setJpegQuality(80);
   ```

2. **Increase concurrency** — Render multiple frames in parallel.
   ```bash
   npx remotion render src/index.ts MyVideo out/video.mp4 --concurrency=50%
   ```

3. **Use `<Img>` not `<img>`** — Remotion's `<Img>` component waits for the image to load.
   A regular `<img>` can cause blank frames.

4. **Avoid heavy per-frame computation** — Memoize expensive calculations with `useMemo`.

5. **Minimize DOM nodes** — Fewer elements per frame = faster rendering.

6. **Use `--gl=angle`** — Can speed up rendering on some systems.
   ```bash
   npx remotion render src/index.ts MyVideo out/video.mp4 --gl=angle
   ```

### Monitoring Render Performance

```bash
# Verbose output with timing info
npx remotion render src/index.ts MyVideo out/video.mp4 --log=verbose

# Benchmark a composition
npx remotion benchmark src/index.ts MyVideo
```

---

## Embedding in React Apps with @remotion/player

If you want to embed and play Remotion videos in a React app (not render to file):

```tsx
import { Player } from '@remotion/player';
import { MyVideo } from './compositions/MyVideo';

function VideoPreview() {
  return (
    <Player
      component={MyVideo}
      inputProps={{ title: 'Hello' }}
      durationInFrames={300}
      fps={30}
      compositionWidth={1920}
      compositionHeight={1080}
      style={{ width: '100%', maxWidth: 800 }}
      controls
      autoPlay
      loop
    />
  );
}
```

This renders in the browser using the same React components, so you can preview
without a server-side render. Useful for dashboards, preview UIs, and editors.
