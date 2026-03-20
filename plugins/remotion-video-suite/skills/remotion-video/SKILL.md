---
name: remotion-video
description: >
  Remotion programmatic video creation — React-based video framework for product demos,
  changelogs, data visualizations, branded content, and marketing videos.
  Scene composition, animation timing, audio sync, and production rendering.
  Triggers: "remotion", "programmatic video", "create video", "video component",
  "useCurrentFrame", "interpolate", "Composition", "Sequence", "video rendering",
  "product demo video", "changelog video", "data visualization video",
  "marketing video", "motion graphics", "video animation".
  NOT for: browser animations (use framer-motion skill), CSS animations (use css-animations skill).
version: 1.0.0
argument-hint: "<video type or description>"
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
model: sonnet
---

# Remotion — Programmatic Video with React

Remotion is a React framework for creating videos programmatically. Instead of timeline-based editing,
you write React components that render each frame. This means you can use all of React's composition,
state, and logic to build dynamic, data-driven videos.

## Why Programmatic Video

- **Automate** — Generate hundreds of personalized videos from data (customer names, metrics, product screenshots)
- **Version control** — Video source is code. Review, branch, merge, revert like any other software
- **Reuse** — Build a component library of scenes, transitions, and effects once, use everywhere
- **Dynamic data** — Pull from APIs, databases, or props to create real-time content
- **Consistency** — Brand guidelines enforced in code, not by manual editing
- **CI/CD** — Render videos in pipelines triggered by events (new release, data update, customer signup)

## Project Setup

```bash
# Create a new Remotion project
npx create-video@latest my-video

# Or add to an existing React project
npm install remotion @remotion/cli @remotion/bundler
```

### Project Structure

```
my-video/
├── src/
│   ├── Root.tsx              # Register all compositions here
│   ├── compositions/
│   │   ├── MyVideo.tsx       # A full video composition
│   │   ├── scenes/           # Individual scene components
│   │   │   ├── TitleCard.tsx
│   │   │   ├── DemoScene.tsx
│   │   │   └── OutroScene.tsx
│   │   └── components/       # Reusable animation components
│   │       ├── FadeIn.tsx
│   │       ├── TypeWriter.tsx
│   │       └── SlideIn.tsx
│   └── lib/
│       ├── colors.ts         # Brand colors
│       ├── fonts.ts          # Font loading
│       └── spring-presets.ts # Animation presets
├── public/                   # Static assets (images, audio, fonts)
├── remotion.config.ts        # Remotion configuration
└── package.json
```

## Core Concepts

### Composition — The Video Container

A Composition defines a video's dimensions, frame rate, and duration. Register them in Root.tsx.

```tsx
// src/Root.tsx
import { Composition } from 'remotion';
import { ProductDemo } from './compositions/ProductDemo';

export const RemotionRoot: React.FC = () => {
  return (
    <>
      <Composition
        id="ProductDemo"
        component={ProductDemo}
        durationInFrames={300}    // 10 seconds at 30fps
        fps={30}
        width={1920}
        height={1080}
        defaultProps={{
          productName: 'My App',
          tagline: 'Ship faster with AI',
        }}
      />
    </>
  );
};
```

### useCurrentFrame — The Animation Clock

Every component can read the current frame number. This is the foundation of all animation in Remotion.

```tsx
import { useCurrentFrame, useVideoConfig } from 'remotion';

export const MyScene: React.FC = () => {
  const frame = useCurrentFrame();
  const { fps, durationInFrames, width, height } = useVideoConfig();

  // frame goes from 0 to durationInFrames - 1
  const opacity = Math.min(1, frame / 30); // Fade in over 1 second

  return (
    <div style={{ opacity, fontSize: 80, color: 'white' }}>
      Frame {frame} of {durationInFrames}
    </div>
  );
};
```

### interpolate — Map Frame to Values

The `interpolate` function maps a frame number to an output range, with optional clamping and easing.

```tsx
import { interpolate, useCurrentFrame, Easing } from 'remotion';

export const AnimatedTitle: React.FC = () => {
  const frame = useCurrentFrame();

  const opacity = interpolate(frame, [0, 30], [0, 1], {
    extrapolateRight: 'clamp',
  });

  const translateY = interpolate(frame, [0, 30], [50, 0], {
    extrapolateRight: 'clamp',
    easing: Easing.out(Easing.cubic),
  });

  const scale = interpolate(frame, [0, 20, 30], [0.8, 1.05, 1], {
    extrapolateRight: 'clamp',
  });

  return (
    <div
      style={{
        opacity,
        transform: `translateY(${translateY}px) scale(${scale})`,
        fontSize: 72,
        fontWeight: 'bold',
        color: '#fff',
      }}
    >
      Welcome to Remotion
    </div>
  );
};
```

### Sequence — Scene Ordering

`Sequence` offsets the frame counter for its children. Inside a Sequence, `useCurrentFrame()` starts at 0.

```tsx
import { Sequence, useVideoConfig } from 'remotion';

export const FullVideo: React.FC = () => {
  const { fps } = useVideoConfig();

  return (
    <div style={{ flex: 1, background: '#0a0a0a' }}>
      {/* Scene 1: Title card — frames 0-89 (3 seconds) */}
      <Sequence from={0} durationInFrames={3 * fps}>
        <TitleCard title="Product Launch" />
      </Sequence>

      {/* Scene 2: Demo — frames 90-239 (5 seconds) */}
      <Sequence from={3 * fps} durationInFrames={5 * fps}>
        <DemoScene />
      </Sequence>

      {/* Scene 3: Outro — frames 240-299 (2 seconds) */}
      <Sequence from={8 * fps} durationInFrames={2 * fps}>
        <OutroScene cta="Try it free" />
      </Sequence>
    </div>
  );
};
```

### spring() — Physics-Based Animation

Springs create natural motion. They settle automatically — no need to specify exact duration.

```tsx
import { spring, useCurrentFrame, useVideoConfig } from 'remotion';

export const BouncyLogo: React.FC = () => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();

  const scale = spring({
    frame,
    fps,
    config: {
      stiffness: 200,
      damping: 12,
      mass: 0.5,
    },
  });

  return (
    <div style={{ transform: `scale(${scale})` }}>
      <img src="/logo.png" style={{ width: 200 }} />
    </div>
  );
};
```

## Audio Integration

```tsx
import { Audio, Sequence, staticFile, interpolate, useCurrentFrame } from 'remotion';

export const VideoWithAudio: React.FC = () => {
  const frame = useCurrentFrame();

  // Fade audio volume
  const volume = interpolate(frame, [0, 30, 270, 300], [0, 0.8, 0.8, 0], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });

  return (
    <>
      <Audio src={staticFile('background-music.mp3')} volume={volume} />
      <Sequence from={30}>
        <Audio src={staticFile('voiceover.mp3')} volume={1} />
      </Sequence>
      <MainContent />
    </>
  );
};
```

## Integration with claude-code-video-toolkit

The [claude-code-video-toolkit](https://github.com/digitalsamba/claude-code-video-toolkit) provides
pre-built components and workflows for common video patterns with Claude Code:

```bash
npx degit digitalsamba/claude-code-video-toolkit my-video-project
```

It includes templates for product demos, changelog videos, and data visualizations that
Claude Code can customize from a brief. Use it as a starting point and customize the
components to match your brand.

## Preview and Render

```bash
# Start the preview server (live reload in browser)
npx remotion studio

# Render to MP4
npx remotion render src/index.ts ProductDemo out/demo.mp4

# Render a specific frame range
npx remotion render src/index.ts ProductDemo out/demo.mp4 --frames=0-90

# Render a still image (thumbnail)
npx remotion still src/index.ts ProductDemo out/thumbnail.png --frame=45
```

## Tips for Using Claude Code with Remotion

1. **Describe the video, not the code.** Say "Create a 15-second product demo video showing
   our app's dashboard with animated metrics" rather than "Write a React component."
2. **Provide brand guidelines.** Share hex colors, font names, and logo files so the output
   matches your brand from the start.
3. **Break complex videos into scenes.** Each scene should be its own component with clear
   props for the content it displays.
4. **Use the reference files.** The component-patterns.md reference contains 10+ ready-to-use
   animated components. The production-rendering.md reference covers all rendering options.
5. **Iterate in the Remotion Studio.** Use `npx remotion studio` to preview while Claude Code
   writes and edits your components.

## Reference Files

- **remotion-quickstart.md** — Step-by-step setup, first composition, preview and render commands
- **component-patterns.md** — 10+ reusable animated components with full TypeScript code
- **production-rendering.md** — Local rendering, Lambda at scale, encoding, CI/CD pipelines
