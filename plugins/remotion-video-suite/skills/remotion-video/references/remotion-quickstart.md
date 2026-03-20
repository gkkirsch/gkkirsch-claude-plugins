# Remotion Quickstart

Complete guide to creating your first programmatic video with Remotion.

## Prerequisites

- Node.js 18+ and npm/pnpm/yarn
- Basic React and TypeScript knowledge
- A code editor (VS Code recommended)

## Step 1: Create a New Project

```bash
# Interactive setup — choose a template
npx create-video@latest my-video

# When prompted:
#   Template: "Hello World" (simplest starting point)
#   Package manager: npm (or your preference)

cd my-video
npm start
```

This starts the Remotion Studio at `http://localhost:3000` with hot reload.

## Step 2: Understand the File Structure

```
my-video/
├── src/
│   ├── Root.tsx              # Entry point — register all Compositions here
│   ├── HelloWorld/
│   │   ├── index.tsx         # The actual video component
│   │   └── ...               # Supporting components
│   └── index.ts              # Remotion entry (registers Root)
├── public/                   # Static assets (images, audio, video, fonts)
├── remotion.config.ts        # Remotion build/render configuration
├── tsconfig.json
└── package.json
```

### Key files explained

**`src/index.ts`** — The entry point Remotion uses to discover your compositions:

```tsx
import { registerRoot } from 'remotion';
import { RemotionRoot } from './Root';

registerRoot(RemotionRoot);
```

**`src/Root.tsx`** — Registers all video compositions (think of each as a "project"):

```tsx
import { Composition } from 'remotion';
import { MyVideo } from './MyVideo';

export const RemotionRoot: React.FC = () => {
  return (
    <>
      <Composition
        id="MyVideo"
        component={MyVideo}
        durationInFrames={150}   // 5 seconds at 30fps
        fps={30}
        width={1920}
        height={1080}
      />
    </>
  );
};
```

**`remotion.config.ts`** — Build and render settings:

```ts
import { Config } from '@remotion/cli/config';

Config.setVideoImageFormat('jpeg');
Config.setOverwriteOutput(true);
```

## Step 3: Write Your First Composition

Create a simple title card with animated text:

```tsx
// src/compositions/TitleCard.tsx
import { useCurrentFrame, useVideoConfig, interpolate, Easing } from 'remotion';

interface TitleCardProps {
  title: string;
  subtitle: string;
  backgroundColor: string;
}

export const TitleCard: React.FC<TitleCardProps> = ({
  title,
  subtitle,
  backgroundColor = '#0f172a',
}) => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();

  // Title fades in and slides up over 1 second
  const titleOpacity = interpolate(frame, [0, fps], [0, 1], {
    extrapolateRight: 'clamp',
  });
  const titleY = interpolate(frame, [0, fps], [40, 0], {
    extrapolateRight: 'clamp',
    easing: Easing.out(Easing.cubic),
  });

  // Subtitle appears after the title, over 0.5 seconds
  const subtitleOpacity = interpolate(frame, [fps * 0.8, fps * 1.3], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });
  const subtitleY = interpolate(frame, [fps * 0.8, fps * 1.3], [20, 0], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
    easing: Easing.out(Easing.cubic),
  });

  // Decorative line grows from center
  const lineWidth = interpolate(frame, [fps * 0.4, fps * 1.2], [0, 200], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
    easing: Easing.out(Easing.cubic),
  });

  return (
    <div
      style={{
        flex: 1,
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        backgroundColor,
        fontFamily: 'Inter, system-ui, sans-serif',
      }}
    >
      <h1
        style={{
          opacity: titleOpacity,
          transform: `translateY(${titleY}px)`,
          fontSize: 80,
          fontWeight: 800,
          color: '#ffffff',
          margin: 0,
          letterSpacing: -2,
        }}
      >
        {title}
      </h1>

      <div
        style={{
          width: lineWidth,
          height: 3,
          backgroundColor: '#3b82f6',
          margin: '24px 0',
          borderRadius: 2,
        }}
      />

      <p
        style={{
          opacity: subtitleOpacity,
          transform: `translateY(${subtitleY}px)`,
          fontSize: 32,
          color: '#94a3b8',
          margin: 0,
          fontWeight: 400,
        }}
      >
        {subtitle}
      </p>
    </div>
  );
};
```

Register it in Root.tsx:

```tsx
import { Composition } from 'remotion';
import { TitleCard } from './compositions/TitleCard';

export const RemotionRoot: React.FC = () => {
  return (
    <>
      <Composition
        id="TitleCard"
        component={TitleCard}
        durationInFrames={150}
        fps={30}
        width={1920}
        height={1080}
        defaultProps={{
          title: 'My Product',
          subtitle: 'Built for developers',
          backgroundColor: '#0f172a',
        }}
      />
    </>
  );
};
```

## Step 4: Preview Your Video

```bash
# Start Remotion Studio (hot-reloading preview)
npx remotion studio

# Or use the legacy command
npx remotion preview
```

The Studio opens in your browser. You can:
- Scrub through the timeline frame by frame
- Edit props in the right panel
- See real-time updates as you save code
- Play the video at actual speed

## Step 5: Add Multiple Scenes with Sequence

```tsx
// src/compositions/ProductDemo.tsx
import { Sequence, useVideoConfig } from 'remotion';
import { TitleCard } from './scenes/TitleCard';
import { FeatureShowcase } from './scenes/FeatureShowcase';
import { CallToAction } from './scenes/CallToAction';

export const ProductDemo: React.FC = () => {
  const { fps } = useVideoConfig();

  return (
    <div style={{ flex: 1, backgroundColor: '#0f172a' }}>
      {/* Scene 1: Title — 3 seconds */}
      <Sequence from={0} durationInFrames={3 * fps} name="Title Card">
        <TitleCard title="Acme Dashboard" subtitle="Analytics made simple" />
      </Sequence>

      {/* Scene 2: Feature showcase — 5 seconds */}
      <Sequence from={3 * fps} durationInFrames={5 * fps} name="Features">
        <FeatureShowcase
          features={[
            { icon: '📊', label: 'Real-time charts' },
            { icon: '🔔', label: 'Smart alerts' },
            { icon: '🚀', label: 'One-click deploy' },
          ]}
        />
      </Sequence>

      {/* Scene 3: CTA — 2 seconds */}
      <Sequence from={8 * fps} durationInFrames={2 * fps} name="CTA">
        <CallToAction text="Start free trial" url="acme.com/start" />
      </Sequence>
    </div>
  );
};
```

## Step 6: Render to Video File

```bash
# Render to MP4 (H.264)
npx remotion render src/index.ts ProductDemo out/product-demo.mp4

# Render to WebM (VP8, smaller file, web-friendly)
npx remotion render src/index.ts ProductDemo out/product-demo.webm --codec=vp8

# Render a GIF (short clips only — large files)
npx remotion render src/index.ts ProductDemo out/product-demo.gif --codec=gif

# Render with custom quality
npx remotion render src/index.ts ProductDemo out/product-demo.mp4 --crf=18

# Render at a different resolution
npx remotion render src/index.ts ProductDemo out/product-demo.mp4 --scale=0.5

# Render a single frame as a still image
npx remotion still src/index.ts ProductDemo out/thumbnail.png --frame=45
```

## Key Imports Reference

| Import | From | Purpose |
|--------|------|---------|
| `Composition` | `remotion` | Register a video with dimensions, fps, and duration |
| `Sequence` | `remotion` | Offset time for child components (scene ordering) |
| `useCurrentFrame` | `remotion` | Get the current frame number (0-based) |
| `useVideoConfig` | `remotion` | Get fps, width, height, durationInFrames |
| `interpolate` | `remotion` | Map frame number to output value with easing |
| `spring` | `remotion` | Physics-based animation value |
| `Easing` | `remotion` | Easing functions for interpolate |
| `Img` | `remotion` | Image component (waits for load before rendering) |
| `Audio` | `remotion` | Audio component with volume control |
| `Video` | `remotion` | Embed video within a composition |
| `AbsoluteFill` | `remotion` | Full-size absolute-positioned container |
| `staticFile` | `remotion` | Reference files in the `public/` directory |
| `registerRoot` | `remotion` | Register the root component (entry point) |
| `delayRender` / `continueRender` | `remotion` | Wait for async data before rendering a frame |

## Common Patterns

### Loading External Data

```tsx
import { delayRender, continueRender, useCurrentFrame } from 'remotion';
import { useEffect, useState } from 'react';

export const DataDrivenScene: React.FC<{ apiUrl: string }> = ({ apiUrl }) => {
  const [data, setData] = useState<any>(null);
  const [handle] = useState(() => delayRender('Loading data...'));

  useEffect(() => {
    fetch(apiUrl)
      .then((res) => res.json())
      .then((json) => {
        setData(json);
        continueRender(handle);
      })
      .catch((err) => {
        console.error(err);
        continueRender(handle);
      });
  }, [apiUrl, handle]);

  if (!data) return null;

  return <div>{/* Render based on data */}</div>;
};
```

### Loading Custom Fonts

```tsx
import { staticFile, delayRender, continueRender } from 'remotion';
import { useEffect, useState } from 'react';

export const loadFont = () => {
  const [handle] = useState(() => delayRender('Loading font...'));

  useEffect(() => {
    const font = new FontFace('CustomFont', `url(${staticFile('fonts/CustomFont.woff2')})`);
    font
      .load()
      .then(() => {
        document.fonts.add(font);
        continueRender(handle);
      })
      .catch((err) => {
        console.error('Font loading failed:', err);
        continueRender(handle);
      });
  }, [handle]);
};
```

### Using Images

```tsx
import { Img, staticFile } from 'remotion';

// Img component ensures the image is loaded before the frame renders
// This prevents blank frames during rendering
export const Screenshot: React.FC = () => {
  return (
    <Img
      src={staticFile('screenshots/dashboard.png')}
      style={{ width: '100%', borderRadius: 12 }}
    />
  );
};
```

### Absolute Positioning with AbsoluteFill

```tsx
import { AbsoluteFill } from 'remotion';

export const LayeredScene: React.FC = () => {
  return (
    <AbsoluteFill style={{ backgroundColor: '#0f172a' }}>
      {/* Background layer */}
      <AbsoluteFill style={{ opacity: 0.3 }}>
        <GradientBackground />
      </AbsoluteFill>

      {/* Content layer */}
      <AbsoluteFill
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
        }}
      >
        <MainContent />
      </AbsoluteFill>

      {/* Overlay layer */}
      <AbsoluteFill style={{ pointerEvents: 'none' }}>
        <WatermarkOverlay />
      </AbsoluteFill>
    </AbsoluteFill>
  );
};
```

## Troubleshooting

### "Could not find a composition"

Make sure your `src/index.ts` calls `registerRoot` and your Root component returns `<Composition>` elements.

### Blank frames during render

Use `<Img>` instead of `<img>` for images. Use `delayRender`/`continueRender` for any async
data loading (fonts, API calls, dynamic imports).

### Video is choppy in preview

The Remotion Studio renders frames on-demand. This is normal during preview. The final rendered
MP4 will be smooth. To check actual frame rate, render a short segment:

```bash
npx remotion render src/index.ts MyVideo out/test.mp4 --frames=0-60
```

### "Cannot find module" errors

Check that all dependencies are installed. Remotion requires specific packages for different
features:

```bash
npm install remotion @remotion/cli              # Core
npm install @remotion/player                     # Embed in React apps
npm install @remotion/lambda                     # AWS Lambda rendering
npm install @remotion/gif                        # GIF support
npm install @remotion/media-utils               # Audio visualization
npm install @remotion/three                      # Three.js integration
```
