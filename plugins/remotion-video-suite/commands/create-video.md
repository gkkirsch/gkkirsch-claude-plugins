# /create-video Command

Create programmatic marketing videos with Remotion. This command activates the Video Producer agent to plan, design, and code a complete Remotion video from your description.

## Usage

```
/create-video <description of the video you want>
```

## How It Works

1. You describe the video you want in plain language
2. The **video-producer** agent creates a scene-by-scene breakdown
3. It writes all Remotion components, scenes, and the main composition
4. The **motion-designer** agent polishes animations and effects
5. You get a complete, renderable Remotion project

## Examples

### Product Demo

```
/create-video 30-second product demo for our SaaS dashboard "MetricsHub"
that shows real-time analytics, custom alerts, and team collaboration.
Brand colors: #1e40af (primary), #f59e0b (accent). Target: YouTube.
```

### Changelog Video

```
/create-video 15-second changelog video for v3.2.0 release.
New features: dark mode, API webhooks, CSV export, faster search.
Minimal style with our logo.
```

### Data Visualization

```
/create-video 20-second animated data visualization showing our Q4 growth:
- Revenue: $0 to $120K
- Users: 0 to 50,000
- Uptime: 99.9%
Style: dark background, blue accent, professional.
```

### Social Media Ad

```
/create-video 15-second Instagram Reel ad for our AI writing tool.
Hook: "Write 10x faster". Show the tool generating a blog post.
End with "Try free at write.ai". Vertical 1080x1920.
```

### Branded Intro

```
/create-video 5-second YouTube channel intro with our logo (logo.png),
company name "Acme Corp", and tagline "Build the future".
Spring animation with particle background.
```

## Subcommands

### `demo`

Create a product demo video.

```
/create-video demo
```

The agent will ask you for:
- Product name and tagline
- Key features to showcase (2-5)
- Screenshots/assets available
- Target platform and duration
- Brand colors and fonts

### `changelog`

Create a release/changelog video.

```
/create-video changelog
```

The agent will ask you for:
- Version number
- List of new features
- Most important change to highlight
- Target platform

### `dataviz`

Create a data visualization video.

```
/create-video dataviz
```

The agent will ask you for:
- Metrics and their values
- Time period
- Comparison data (optional)
- Chart type preference

### `social`

Create a short social media ad/clip.

```
/create-video social
```

The agent will ask you for:
- Platform (Instagram, TikTok, Twitter, LinkedIn)
- Core message/hook
- CTA text and URL
- Duration (default: 15 seconds)

## What You Get

A complete Remotion project with:

```
src/
├── Root.tsx                    # Composition registered and ready
├── compositions/
│   └── YourVideo.tsx           # Main composition
├── scenes/                     # Individual scene components
│   ├── IntroScene.tsx
│   ├── MainScene.tsx
│   └── CTAScene.tsx
├── components/                 # Reusable animation components
│   ├── Fade.tsx
│   ├── TypewriterText.tsx
│   └── ...
└── lib/
    ├── brand.ts               # Your brand colors/fonts
    └── spring-presets.ts      # Animation configs
```

## After Running

```bash
# Preview your video
npx remotion studio

# Render to MP4
npx remotion render src/index.ts YourVideo out/video.mp4

# Render a thumbnail
npx remotion still src/index.ts YourVideo out/thumbnail.png --frame=45
```

## Tips

- **Be specific about your brand** — include hex colors, font names, and logo file paths
- **Mention the platform** — the agent adjusts resolution and pacing for each platform
- **Provide screenshots** — place them in `public/screenshots/` before running
- **Describe the feeling** — "professional and clean" vs "fun and energetic" changes the animation style
- **Iterate** — run the command, preview with `npx remotion studio`, then ask for adjustments

## Reference Files

This suite includes detailed reference docs used by the agents:
- **remotion-quickstart.md** — Setup, first composition, preview/render commands
- **component-patterns.md** — 12+ ready-to-use animated components with full code
- **production-rendering.md** — Rendering to all formats, Lambda, CI/CD
