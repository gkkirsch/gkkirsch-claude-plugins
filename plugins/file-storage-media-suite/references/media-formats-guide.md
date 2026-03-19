# Media Formats & Optimization Guide

Quick reference for image, video, and audio formats, quality settings, and optimization targets.

---

## Image Formats

### Format Comparison

| Format | Type | Transparency | Animation | Compression | Size vs JPEG |
|--------|------|-------------|-----------|-------------|-------------|
| **JPEG** | Lossy | No | No | Good | 1x (baseline) |
| **PNG** | Lossless | Yes | No | Poor for photos | 3-5x larger |
| **WebP** | Both | Yes | Yes | Very good | 25-35% smaller |
| **AVIF** | Both | Yes | Yes | Excellent | 50% smaller |
| **GIF** | Lossless (256 colors) | Yes | Yes | Poor | Variable |
| **SVG** | Vector | Yes | Yes (CSS/JS) | N/A | Tiny for shapes |
| **JPEG XL** | Both | Yes | Yes | Best | 60% smaller |
| **HEIF** | Lossy | Yes | Yes | Excellent | ~50% smaller |

### Browser Support (as of 2026)

| Format | Chrome | Firefox | Safari | Edge |
|--------|--------|---------|--------|------|
| JPEG | 100% | 100% | 100% | 100% |
| PNG | 100% | 100% | 100% | 100% |
| WebP | 100% | 100% | 100% | 100% |
| AVIF | 100% | 100% | 16.4+ | 100% |
| JPEG XL | No | No | 17+ | No |
| GIF | 100% | 100% | 100% | 100% |
| SVG | 100% | 100% | 100% | 100% |

### Recommended Quality Settings

| Use Case | WebP Quality | AVIF Quality | Sharp Options |
|----------|-------------|-------------|---------------|
| Photography | 80-85 | 60-70 | `{ quality: 82 }` |
| Product images | 85-90 | 70-75 | `{ quality: 87 }` |
| Thumbnails | 70-80 | 55-65 | `{ quality: 75 }` |
| Avatars | 80 | 65 | `{ quality: 80 }` |
| Screenshots | 90+ | 80+ | `{ quality: 90 }` or lossless |
| Icons/logos | Lossless | Lossless | `{ lossless: true }` or use SVG |
| Blur placeholders | 20-30 | 15-25 | `{ quality: 20 }` |
| OG images | 80 | 65 | `{ quality: 80 }` |

### Image Size Targets (After Optimization)

| Type | Dimensions | Max File Size | Notes |
|------|-----------|---------------|-------|
| Favicon | 32x32, 180x180 | 5KB | ICO + Apple touch |
| Avatar | 128-256px | 20-30KB | Square, face-cropped |
| Thumbnail | 200-400px | 30-50KB | Card previews |
| Card image | 600-800px | 80-120KB | Blog, product cards |
| Hero image | 1920px | 150-250KB | Above the fold |
| Full-size | 2560px | 300-500KB | Lightbox/gallery |
| OG image | 1200x630 | 80-100KB | Social sharing |
| Background | 1920px | 100-200KB | CSS background |

### Responsive Image Breakpoints

```
Mobile:   320px, 375px, 414px     → serve 400-640px images
Tablet:   768px, 1024px           → serve 800-1200px images
Desktop:  1280px, 1440px, 1920px  → serve 1200-1920px images
Retina:   2x DPR                  → serve 2x width images
```

Standard `sizes` attribute patterns:

```html
<!-- Full-width hero -->
sizes="100vw"

<!-- Card in responsive grid -->
sizes="(max-width: 640px) 100vw, (max-width: 1024px) 50vw, 33vw"

<!-- Sidebar image -->
sizes="(max-width: 768px) 100vw, 300px"

<!-- Fixed-width thumbnail -->
sizes="200px"
```

---

## Video Formats

### Container Formats

| Format | Codecs | Use Case | Support |
|--------|--------|----------|---------|
| **MP4** | H.264, H.265, AV1 | Universal web video | 100% |
| **WebM** | VP8, VP9, AV1 | Web-optimized | Chrome, Firefox, Edge |
| **MOV** | H.264, ProRes | Source/editing (Apple) | Safari, QuickTime |
| **MKV** | Any | Archival, subtitles | Limited browser |

### Video Codecs

| Codec | Quality | Speed | File Size | Browser Support |
|-------|---------|-------|-----------|----------------|
| **H.264** | Good | Fast | Baseline | 100% |
| **H.265/HEVC** | Better | Medium | 50% of H.264 | Safari, some Chrome |
| **VP9** | Better | Slow | 50% of H.264 | Chrome, Firefox, Edge |
| **AV1** | Best | Very slow | 30% of H.264 | Chrome 70+, Firefox 67+ |

### Video Quality Settings (FFmpeg CRF)

| CRF Value | Quality | Use Case | Approximate Bitrate (1080p) |
|-----------|---------|----------|---------------------------|
| 18 | Visually lossless | Archival | ~8 Mbps |
| 20 | Excellent | High-quality delivery | ~5 Mbps |
| 23 | Good (default) | Standard web video | ~3 Mbps |
| 26 | Acceptable | Mobile, bandwidth-constrained | ~1.5 Mbps |
| 28 | Low | Preview, thumbnail video | ~1 Mbps |
| 30+ | Poor | Background, ambient | <0.5 Mbps |

### HLS/DASH Bitrate Ladder

| Resolution | Video Bitrate | Audio Bitrate | Target |
|-----------|--------------|--------------|--------|
| 360p (640x360) | 800 kbps | 96 kbps | Mobile 3G |
| 480p (854x480) | 1.4 Mbps | 128 kbps | Mobile 4G |
| 720p (1280x720) | 2.8 Mbps | 128 kbps | Desktop/WiFi |
| 1080p (1920x1080) | 5.0 Mbps | 192 kbps | High-speed |
| 1440p (2560x1440) | 8.0 Mbps | 192 kbps | 4K displays |
| 4K (3840x2160) | 14.0 Mbps | 256 kbps | 4K native |

### Video File Size Estimates (H.264, CRF 23)

| Duration | 720p | 1080p | 4K |
|----------|------|-------|-----|
| 30 sec | ~10 MB | ~18 MB | ~50 MB |
| 1 min | ~20 MB | ~35 MB | ~100 MB |
| 5 min | ~100 MB | ~175 MB | ~500 MB |
| 10 min | ~200 MB | ~350 MB | ~1 GB |
| 1 hour | ~1.2 GB | ~2.1 GB | ~6 GB |

---

## Audio Formats

| Format | Type | Quality | File Size | Use Case |
|--------|------|---------|-----------|----------|
| **MP3** | Lossy | Good (128-320 kbps) | Small | Universal playback |
| **AAC** | Lossy | Better than MP3 | Small | Video audio, streaming |
| **OGG/Opus** | Lossy | Excellent | Smallest | Web, VoIP, streaming |
| **WAV** | Uncompressed | Perfect | Very large | Source/editing |
| **FLAC** | Lossless | Perfect | Large | Archival |

### Audio Bitrate Guide

| Quality | MP3 | AAC | Opus |
|---------|-----|-----|------|
| Voice/podcast | 64 kbps | 48 kbps | 32 kbps |
| FM radio quality | 128 kbps | 96 kbps | 64 kbps |
| CD quality | 256-320 kbps | 192 kbps | 128 kbps |
| Transparent | 320 kbps (VBR V0) | 256 kbps | 160 kbps |

---

## Performance Budgets

### Page Weight Targets

| Page Type | Total Size | Images | Video | JS | CSS |
|-----------|-----------|--------|-------|-----|-----|
| Landing page | < 1.5 MB | < 500 KB | < 500 KB | < 300 KB | < 100 KB |
| Blog post | < 1 MB | < 400 KB | 0 | < 200 KB | < 80 KB |
| Product page | < 2 MB | < 800 KB | < 500 KB | < 300 KB | < 100 KB |
| Dashboard | < 1 MB | < 200 KB | 0 | < 500 KB | < 100 KB |

### Core Web Vitals Targets

| Metric | Good | Needs Work | Poor |
|--------|------|-----------|------|
| **LCP** (Largest Contentful Paint) | < 2.5s | 2.5-4.0s | > 4.0s |
| **INP** (Interaction to Next Paint) | < 200ms | 200-500ms | > 500ms |
| **CLS** (Cumulative Layout Shift) | < 0.1 | 0.1-0.25 | > 0.25 |

### Image Loading Best Practices

```html
<!-- Above the fold: preload, no lazy, set dimensions -->
<img src="hero.webp" width="1200" height="600" fetchpriority="high" alt="..." />

<!-- Below the fold: lazy load, set dimensions -->
<img src="card.webp" width="400" height="300" loading="lazy" decoding="async" alt="..." />

<!-- Always set width and height to prevent CLS -->
<!-- Use aspect-ratio CSS if dimensions are dynamic -->
<style>
  .responsive-img { aspect-ratio: 16/9; width: 100%; object-fit: cover; }
</style>
```

---

## File Signature (Magic Bytes)

| Format | Hex Signature | ASCII |
|--------|--------------|-------|
| JPEG | `FF D8 FF` | ... |
| PNG | `89 50 4E 47 0D 0A 1A 0A` | .PNG.... |
| GIF | `47 49 46 38` | GIF8 |
| WebP | `52 49 46 46 ?? ?? ?? ?? 57 45 42 50` | RIFF....WEBP |
| PDF | `25 50 44 46` | %PDF |
| MP4 | `?? ?? ?? ?? 66 74 79 70` | ....ftyp |
| ZIP | `50 4B 03 04` | PK.. |
| GZIP | `1F 8B` | .. |

Use these for server-side validation — never trust file extensions alone.

---

## MIME Types Reference

| Extension | MIME Type |
|-----------|----------|
| .jpg, .jpeg | image/jpeg |
| .png | image/png |
| .webp | image/webp |
| .avif | image/avif |
| .gif | image/gif |
| .svg | image/svg+xml |
| .ico | image/x-icon |
| .mp4 | video/mp4 |
| .webm | video/webm |
| .m3u8 | application/vnd.apple.mpegurl |
| .ts | video/mp2t |
| .mp3 | audio/mpeg |
| .ogg | audio/ogg |
| .wav | audio/wav |
| .pdf | application/pdf |
| .zip | application/zip |
