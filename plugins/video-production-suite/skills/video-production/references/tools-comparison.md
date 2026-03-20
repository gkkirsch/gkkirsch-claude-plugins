# Video Production Tools Comparison

Comprehensive comparison of every major tool in the modern video production stack, organized by category. Includes pricing, platform support, key features, and "best for" recommendations.

Last updated: March 2026

---

## Screen Recording Tools

### Screen Studio (macOS)
- **Price**: $89 one-time (personal), $139 (teams)
- **Platform**: macOS only
- **Key Features**:
  - Auto-zoom: Automatically follows your cursor and zooms into click interactions
  - Beautiful device frames (MacBook, iMac, iPhone, etc.)
  - Gradient backgrounds with customizable colors
  - Cursor smoothing and click effects (ripple, highlight, magnify)
  - Keyboard shortcut visualization
  - Built-in timeline for editing zoom keyframes
  - Export at up to 4K 60fps
  - Wallpaper hiding and desktop cleanup
- **Best For**: Polished product demos and marketing screen recordings where you want that "Apple keynote" quality. The auto-zoom alone saves hours of editing.
- **Limitations**: macOS only. No webcam overlay (though you can composite in a separate tool). No real-time streaming.

### OBS Studio (Cross-Platform)
- **Price**: Free and open source
- **Platform**: macOS, Windows, Linux
- **Key Features**:
  - Unlimited scenes with custom layouts
  - Multiple sources: screen, window, camera, images, text, browser
  - Real-time streaming to Twitch, YouTube, etc.
  - Advanced audio mixer with filters (noise gate, compressor, noise suppression)
  - Plugin ecosystem (hundreds of community plugins)
  - Studio mode for professional multi-source switching
  - Virtual camera output for Zoom/Meet
  - Replay buffer for capturing "highlight moments"
- **Best For**: Free, highly configurable recording and streaming. Ideal when you need complex scene setups, multiple camera angles, or real-time streaming. Tutorial creators who record long sessions.
- **Limitations**: Steeper learning curve. No auto-zoom or cursor beautification built in. Output requires more post-production polish.

### Matte (macOS)
- **Price**: $29/year
- **Platform**: macOS only
- **Key Features**:
  - Minimal, clean recording interface
  - Auto device frames
  - Background customization (gradients, solid colors, images)
  - Rounded corners on recordings
  - Simple trim and crop editing
  - Quick export with size presets
- **Best For**: Quick, clean recordings without the complexity of Screen Studio. Good for developers who want polished screenshots and short clips for documentation or Twitter.
- **Limitations**: Less feature-rich than Screen Studio. No auto-zoom. Limited editing capabilities.

### FocuSee (macOS/Windows)
- **Price**: $69 one-time (personal), $99 (pro)
- **Platform**: macOS, Windows
- **Key Features**:
  - Auto-zoom and pan following cursor movement
  - Click effects and cursor highlighting
  - Background customization
  - Device frame mockups
  - Built-in basic editing (trim, speed adjustment)
  - Annotation tools
  - Export in multiple aspect ratios
- **Best For**: Screen Studio alternative for Windows users. Cross-platform teams that need consistent recording quality across Mac and Windows.
- **Limitations**: Auto-zoom is not quite as polished as Screen Studio. Smaller community and fewer tutorials available.

### CleanShot X (macOS)
- **Price**: $29 one-time (basic), $8/month (cloud)
- **Platform**: macOS only
- **Key Features**:
  - Screenshot and screen recording in one tool
  - Annotation and markup tools (arrows, shapes, text, blur, highlight)
  - Scrolling capture (full webpage screenshots)
  - GIF recording
  - Quick access overlay
  - Cloud upload and sharing
  - Video trimming
  - OCR text recognition from screenshots
- **Best For**: Quick captures for bug reports, documentation, and social media posts. More of a screenshot tool with recording capabilities than a dedicated recorder.
- **Limitations**: Recording features are basic compared to Screen Studio. No auto-zoom. No device frames in recordings.

### Screen Recording Comparison Matrix

| Feature | Screen Studio | OBS | Matte | FocuSee | CleanShot X |
|---------|:---:|:---:|:---:|:---:|:---:|
| Auto-zoom | Yes | No | No | Yes | No |
| Device frames | Yes | No | Yes | Yes | No |
| Custom backgrounds | Yes | Yes | Yes | Yes | No |
| Cursor effects | Yes | No | No | Yes | No |
| Built-in editing | Basic | No | Basic | Basic | Trim only |
| Live streaming | No | Yes | No | No | No |
| Webcam overlay | No | Yes | No | Yes | No |
| Windows support | No | Yes | No | Yes | No |
| Free tier | No | Yes | No | No | No |
| 4K recording | Yes | Yes | Yes | Yes | Yes |
| Price | $89 | Free | $29/yr | $69 | $29 |

**Verdict**: Screen Studio for polished product demos. OBS for streaming and complex setups. CleanShot X for quick captures and bug reports. FocuSee for Windows users who need auto-zoom.

---

## Video Editing Software

### Descript
- **Price**: Free (1 video/month), $24/month (Hobbyist), $33/month (Business)
- **Platform**: macOS, Windows, Web
- **Key Features**:
  - Edit video by editing text (transcript-based editing)
  - Auto-remove filler words ("um", "uh", "like", "you know")
  - AI Voice (Overdub) — clone your voice and generate narration by typing
  - Studio Sound — one-click audio cleanup
  - Eye Contact — AI eye contact correction for webcam footage
  - AI Green Screen — remove backgrounds without a green screen
  - Auto-captions with styling
  - Templates and brand kits
  - Team collaboration with comments
  - Direct publish to YouTube, social media
- **Best For**: Narration-heavy content, podcasters who also make video, solo creators who want speed over granular control, teams that need non-editors to review content.
- **Limitations**: Limited motion graphics capabilities. Color grading is basic. Complex multi-track editing is possible but not its strength.

### CapCut (Desktop)
- **Price**: Free (with watermark on some features), $7.99/month (Pro)
- **Platform**: macOS, Windows, Web, iOS, Android
- **Key Features**:
  - Auto-captions with dozens of animated styles (including viral word-by-word highlight)
  - Text-to-speech with multiple AI voices
  - AI background removal (Chroma Key)
  - Auto reframe (16:9 to 9:16 conversion)
  - Beat sync (auto-align cuts to music beats)
  - Massive library of effects, transitions, stickers, and filters
  - Speed ramping with curves
  - Keyframe animations
  - Cloud project sync across devices
  - Direct TikTok integration
- **Best For**: Social media content creators. Fast editing with trendy effects. Anyone making Reels, TikToks, or YouTube Shorts. Teams that need mobile-to-desktop workflows.
- **Limitations**: Professional features are limited compared to DaVinci Resolve or Premiere. Color grading is preset-based, not node-based. Some advanced features require Pro subscription. Data is processed on ByteDance servers (privacy consideration).

### DaVinci Resolve
- **Price**: Free (full-featured), $295 one-time (Studio — adds GPU acceleration, HDR, collaboration)
- **Platform**: macOS, Windows, Linux
- **Key Features**:
  - Industry-leading color grading (Hollywood standard — used on major films)
  - Node-based color correction with unlimited nodes
  - Fairlight audio post-production suite
  - Fusion visual effects and motion graphics (node-based compositing)
  - Multi-user collaboration (Studio version)
  - Professional editing tools comparable to Premiere Pro and Final Cut
  - Hardware panel support (DaVinci Resolve panels for color grading)
  - AI-powered features: Magic Mask, Speed Warp, voice isolation
  - ACES color management
  - Dual timeline editing
- **Best For**: Professional-quality production. Anyone who needs serious color grading. Long-form content with complex audio. Editors migrating from Premiere Pro.
- **Limitations**: Steeper learning curve. GPU-intensive for real-time playback. Free version lacks GPU acceleration on some systems, neural engine AI features, HDR grading, and multi-user collaboration.

### Final Cut Pro
- **Price**: $299.99 one-time (or $4.99/month subscription)
- **Platform**: macOS only (iPad version available)
- **Key Features**:
  - Magnetic Timeline for fast, intuitive editing
  - Object Tracker — attach graphics to moving objects in footage
  - Cinematic mode support (rack focus adjustments)
  - ProRes RAW support
  - Multicam editing (up to 64 angles)
  - Motion graphics integration (Apple Motion — $49.99)
  - Hardware-optimized for Apple Silicon (extremely fast on M-series chips)
  - HDR support and wide color gamut
  - Roles-based audio organization
  - Compressor integration for advanced export
- **Best For**: Mac users who want the fastest possible editing experience. Creators invested in the Apple ecosystem. Anyone with Apple Silicon hardware who wants hardware-optimized performance.
- **Limitations**: macOS only. No Windows or Linux. Magnetic Timeline takes getting used to for traditional NLE editors. Plugin ecosystem is smaller than Premiere Pro.

### Video Editing Comparison Matrix

| Feature | Descript | CapCut | DaVinci Resolve | Final Cut Pro |
|---------|:---:|:---:|:---:|:---:|
| Text-based editing | Yes | No | No | No |
| Auto-captions | Yes | Yes (best) | Plugin | Plugin |
| Color grading | Basic | Presets | Professional | Good |
| Audio post-production | Good | Basic | Professional | Good |
| Motion graphics | Basic | Templates | Fusion (advanced) | Motion integration |
| AI voice generation | Yes | Yes (TTS) | No | No |
| Team collaboration | Yes | Cloud sync | Yes (Studio) | No |
| Free version | Limited | Most features | Full-featured | No |
| Windows support | Yes | Yes | Yes | No |
| Learning curve | Low | Low | High | Medium |
| Best for social media | Good | Excellent | Overkill | Good |
| Best for long-form | Good | Limited | Excellent | Excellent |
| Price | $24+/mo | Free-$8/mo | Free-$295 | $300 or $5/mo |

**Verdict**: Descript for speed and narration-heavy content. CapCut for social media and quick edits. DaVinci Resolve for professional quality (free version is remarkable). Final Cut Pro for Apple ecosystem speed.

---

## AI Video Generation

### OpenAI Sora 2
- **Price**: Included with ChatGPT Plus ($20/month), Pro ($200/month for more generations)
- **Key Features**:
  - Text-to-video generation up to 20 seconds (1080p)
  - Image-to-video animation
  - Video-to-video style transfer and extension
  - Storyboard mode for multi-scene generation
  - Remix and blend existing videos
  - Multiple aspect ratios (16:9, 9:16, 1:1)
- **Best For**: B-roll generation, concept visualization, social media content experiments, creative shorts.
- **Limitations**: 20 second maximum per generation. Inconsistent with text rendering in video. Can struggle with specific product UI recreation. Watermarked on free/Plus tier.

### Kling AI
- **Price**: Free tier (limited), Pro plans available
- **Key Features**:
  - Text-to-video up to 2 minutes
  - Motion brush for directing movement in specific areas
  - Image-to-video with consistent character/object generation
  - Camera control (pan, zoom, orbit, tracking shots)
  - Lip sync for AI avatar generation
  - High-quality 1080p output
- **Best For**: Longer AI video clips. Controlled camera movements. Creative B-roll with specific motion direction.
- **Limitations**: Processing times can be long. Quality varies between generations. Limited fine control over complex scenes.

### Runway Gen-3 Alpha
- **Price**: Free (limited), $12/month (Standard), $28/month (Pro), $76/month (Unlimited)
- **Key Features**:
  - Text-to-video and image-to-video
  - Motion brush for directing movement
  - Camera controls (pan, tilt, zoom, roll)
  - Multi Motion Brush for complex scene control
  - Extend video clips to create longer sequences
  - Style transfer and video-to-video transformation
  - Custom model training
  - API access for automated workflows
- **Best For**: Creative professionals who need fine control over AI video. Advertising and marketing teams. Integration into existing production pipelines via API.
- **Limitations**: Most useful features require paid plans. Generation credits are consumed quickly on complex prompts.

### Google Veo 3
- **Price**: Included with Gemini Advanced and Google AI Studio
- **Key Features**:
  - Text-to-video generation
  - Native audio generation with video (dialogue, sound effects, ambient sound)
  - High-resolution output (up to 4K in some modes)
  - Integration with Google ecosystem
  - Flow (Veo-based filmmaking tool for longer narratives)
- **Best For**: Generating video with synchronized audio. Projects where ambient sound and dialogue need to be part of the generated content. Google ecosystem users.
- **Limitations**: Access can be limited. Quality can vary. Less community tooling compared to Runway.

### AI Video Comparison Matrix

| Feature | Sora 2 | Kling | Runway Gen-3 | Veo 3 |
|---------|:---:|:---:|:---:|:---:|
| Max duration | 20s | 2 min | 16s (extendable) | Varies |
| Resolution | 1080p | 1080p | Up to 4K | Up to 4K |
| Camera control | Limited | Yes | Yes (advanced) | Limited |
| Motion brush | No | Yes | Yes | No |
| Audio generation | No | Limited | No | Yes |
| Image-to-video | Yes | Yes | Yes | Yes |
| API access | Yes | Limited | Yes | Yes |
| Free tier | With ChatGPT+ | Yes | Limited | With Gemini |
| Best quality | High | High | High | High |

**Verdict**: Sora 2 for quick creative generation. Runway for professional control and API workflows. Kling for longer clips with camera control. Veo 3 for video with native audio.

---

## AI Voiceover Tools

### ElevenLabs
- **Price**: Free (10K chars/month), $5/month (Starter, 30K chars), $22/month (Creator, 100K chars), $99/month (Pro, 500K chars)
- **Key Features**:
  - Industry-leading voice quality (most natural-sounding AI voices)
  - 30+ pre-made voices across languages, accents, and styles
  - Voice cloning from audio samples (Professional and Enterprise plans)
  - Speech-to-speech voice conversion
  - Projects mode for long-form content with multiple speakers
  - SSML support for fine-grained control over pronunciation and pacing
  - API access for automated pipelines
  - Dubbing (automatic translation and voice-matching to 29 languages)
  - Sound effects generation
- **Best For**: Professional voiceover for product videos. Voice cloning for brand consistency. Multi-language content. Highest-quality AI narration available.
- **Limitations**: Best voices and features are on paid plans. Voice cloning requires verified consent. Processing time for long scripts.

### Descript AI Voice (Overdub)
- **Price**: Included with Descript Hobbyist ($24/month) and above
- **Key Features**:
  - Voice cloning from your recordings within Descript
  - Type text to generate speech in your cloned voice
  - Integrated into Descript's editing workflow
  - Correct mistakes in narration by retyping the word
  - Stock AI voices available without cloning
- **Best For**: Descript users who want seamless integration. Fixing narration mistakes without re-recording. Quick regeneration of specific sections.
- **Limitations**: Voice quality is good but not quite ElevenLabs level. Requires Descript subscription. Cloned voice requires sufficient training data from your recordings.

### CapCut Text-to-Speech
- **Price**: Free (with CapCut), some voices require Pro ($7.99/month)
- **Key Features**:
  - Built into CapCut's editing workflow
  - Multiple voice options across languages
  - Speed and pitch control
  - Direct placement on timeline
  - Some character-specific voices (trending TikTok voices)
- **Best For**: Quick social media content where you need a voice but do not want to record. TikTok-style narration. Fast, no-frills voiceover.
- **Limitations**: Voice quality is lower than ElevenLabs. Limited customization. Some popular voices rotate in and out of availability.

### AI Voiceover Comparison Matrix

| Feature | ElevenLabs | Descript AI | CapCut TTS |
|---------|:---:|:---:|:---:|
| Voice quality | Excellent | Good | Decent |
| Voice cloning | Yes (paid) | Yes (included) | No |
| # of voices | 30+ | 15+ | 20+ |
| SSML support | Yes | No | No |
| API access | Yes | No | No |
| Multi-language | 29 languages | Limited | Multiple |
| Dubbing | Yes | No | No |
| Free tier | 10K chars | No (requires sub) | Yes |
| Integration | API, web, SDK | Descript editor | CapCut editor |

**Verdict**: ElevenLabs for highest quality and flexibility. Descript AI for integrated editing workflow. CapCut TTS for quick social content.

---

## AI Avatar and Presenter Tools

### HeyGen
- **Price**: Free (limited), $24/month (Creator), $72/month (Business)
- **Key Features**:
  - AI avatars from real video recordings (photo and video cloning)
  - 200+ diverse stock avatars
  - Script-to-video with lip-synced AI presenter
  - Multi-language translation with lip sync
  - Custom branded backgrounds
  - Template library for common video types
  - API for automated video generation
  - Interactive avatar for live conversations
- **Best For**: Training videos, sales presentations, multi-language content at scale. When you need a "presenter" without recording footage.
- **Limitations**: Uncanny valley effect on close inspection. Custom avatar cloning requires Enterprise plan or additional fees. Monthly generation limits.

### Synthesia
- **Price**: $22/month (Starter), $67/month (Creator), custom (Enterprise)
- **Key Features**:
  - 240+ AI avatars
  - 140+ languages with native-quality lip sync
  - Screen recording integration (avatar alongside screen demo)
  - AI script assistant
  - Templates for training, marketing, and sales
  - Brand kit (logos, colors, fonts)
  - Collaboration features with team workspaces
  - SOC 2 Type II compliance (enterprise-ready)
- **Best For**: Enterprise training and internal communications. HR onboarding videos. Scalable multi-language content production. Compliance-sensitive industries.
- **Limitations**: Avatar quality is good but clearly AI on close inspection. Most natural avatars require higher-tier plans. Limited creative control compared to actual video production.

### CapCut AI Avatar
- **Price**: Included with CapCut Pro ($7.99/month)
- **Key Features**:
  - AI avatar generation from photos
  - Text-to-speech combined with avatar lip sync
  - Integrated into CapCut's editing workflow
  - Multiple styles and expressions
  - Quick generation for social content
- **Best For**: Quick social media content with an AI presenter. Budget-conscious creators who already use CapCut.
- **Limitations**: Fewer avatars and less customization than HeyGen or Synthesia. Quality is lower tier. Limited to CapCut's editing environment.

---

## Music and Sound Effect Libraries

### Epidemic Sound
- **Price**: $15/month (Personal), $49/month (Commercial)
- **Library Size**: 50,000+ tracks, 200,000+ SFX
- **License**: Unlimited use during subscription. Content remains licensed even after cancellation for existing published content.
- **Best For**: Professional creators who need a massive, high-quality library. YouTube, social media, and commercial projects.
- **Unique Feature**: "Find Similar" — find tracks that sound like a reference track you like.

### Artlist
- **Price**: $9.99/month (Music), $16.60/month (Music + SFX), $29.99/month (Max)
- **Library Size**: 30,000+ tracks, 100,000+ SFX
- **License**: Lifetime license for downloaded tracks. Continue using them even after cancellation.
- **Best For**: Filmmakers and video producers who want cinematic, high-end music. The lifetime license model is excellent for peace of mind.
- **Unique Feature**: Lifetime license for downloaded tracks during subscription.

### Mixkit
- **Price**: Free (completely free, no attribution required)
- **Library Size**: 5,000+ tracks, 5,000+ SFX
- **License**: Free for commercial and personal use. No attribution required.
- **Best For**: Budget projects, quick prototypes, and anyone who needs royalty-free music without any cost. Quality is surprisingly good for a free library.
- **Unique Feature**: Completely free with no strings attached.

### Pixabay Music/Audio
- **Price**: Free
- **Library Size**: 10,000+ tracks and SFX
- **License**: Free for commercial use. No attribution required (but appreciated).
- **Best For**: Similar to Mixkit — budget projects and quick needs. Larger library than Mixkit.
- **Unique Feature**: Community-contributed content. Wide variety of genres.

### Beatoven.ai
- **Price**: Free (limited), $6/month (Starter, 15 minutes), $20/month (Pro, 60 minutes)
- **Key Features**:
  - AI-generated custom music based on mood, genre, and instruments
  - Adjust energy, tempo, and mood for different video sections
  - Generate music to match your video's timing
  - Royalty-free for all generated content
- **Best For**: Custom background music that exactly matches your video's energy and timing. When you cannot find the right track in a library.
- **Unique Feature**: Music is generated to your specifications — completely unique every time.

### Music Library Comparison Matrix

| Feature | Epidemic Sound | Artlist | Mixkit | Pixabay | Beatoven.ai |
|---------|:---:|:---:|:---:|:---:|:---:|
| Price | $15/mo | $10/mo | Free | Free | $6-20/mo |
| Library size | 50K+ | 30K+ | 5K+ | 10K+ | AI generated |
| SFX included | Yes | Separate plan | Yes | Yes | No |
| License type | Subscription | Lifetime | Free | Free | Royalty-free |
| Custom music | No | No | No | No | Yes |
| Quality | Professional | Cinematic | Good | Varies | Good |
| Best for | YouTubers | Filmmakers | Budget | Budget | Custom needs |

---

## 3D Mockup and Device Frame Tools

### Rotato (macOS)
- **Price**: $49/year (personal), $99/year (team)
- **Platform**: macOS only
- **Key Features**:
  - Import screenshots or video into 3D device mockups
  - Animate rotation, zoom, and transition between devices
  - Supports: iPhone, iPad, MacBook, iMac, Apple Watch, Android phones, browser windows
  - Export as MP4 video, GIF, or PNG sequence
  - Drag-and-drop simplicity
  - Custom background colors and environments
  - Batch export multiple devices from the same screenshot
- **Best For**: App store preview videos, landing page hero shots, social media product announcements, pitch deck visuals.

### Slant it (Web)
- **Price**: Free
- **Platform**: Web browser (any OS)
- **Key Features**:
  - Browser-based 3D mockup generator
  - Drag and drop screenshots into device templates
  - Multiple device types and angles
  - Quick export as PNG or video
  - No account required
- **Best For**: Quick mockups when you do not need animation. Free alternative to Rotato for static device frames.

### MockRocket
- **Price**: Free tier, $12/month (Pro), $36/month (Business)
- **Platform**: Web browser
- **Key Features**:
  - 3D device mockup generator
  - Video mockups (insert video into device frame)
  - Multiple angles and customizable scenes
  - Animated presentations
  - Custom branding and backgrounds
- **Best For**: Web-based alternative to Rotato with video support. Teams that need cross-platform access.

---

## Interactive Demo Tools

These tools create clickable, interactive product demos — not traditional video, but increasingly used alongside or instead of video demos.

### Arcade
- **Price**: Free (limited), $32/month (Pro), $42/month (Growth)
- **Key Features**:
  - Chrome extension captures clicks into an interactive demo
  - No-code editing of captured flows
  - Embed anywhere (website, docs, email)
  - Analytics on viewer engagement
  - Callouts, tooltips, and step-by-step annotations
  - Branching paths for different user personas
- **Best For**: Self-serve product tours on websites. Sales enablement. Documentation with interactive examples.

### Storylane
- **Price**: Free (limited), $40/month (Starter), $100/month (Growth)
- **Key Features**:
  - Screenshot-based and HTML capture interactive demos
  - Edit mode for modifying captured UI without re-recording
  - Lead capture forms within demos
  - CRM integrations (HubSpot, Salesforce)
  - Analytics dashboard
  - Custom branding
- **Best For**: B2B SaaS sales teams. Product-led growth strategies. Demos that need lead capture.

### Supademo
- **Price**: Free (5 demos), $27/month (Pro), $38/month (Scale)
- **Key Features**:
  - AI-powered step detection from screen recordings
  - Auto-generated descriptions and annotations
  - Multi-language support
  - Embed and share with custom domains
  - Viewer analytics
  - Blur sensitive data automatically
- **Best For**: Quick demo creation with minimal effort. AI-assisted annotation saves time. Good balance of features and price.

### Interactive Demo Comparison Matrix

| Feature | Arcade | Storylane | Supademo |
|---------|:---:|:---:|:---:|
| Capture method | Chrome extension | Screenshot/HTML | Screen recording |
| AI annotation | No | Limited | Yes |
| Lead capture | Growth plan | Yes | Pro plan |
| Branching paths | Yes | Yes | Yes |
| Analytics | Yes | Yes | Yes |
| Free tier | Limited | Limited | 5 demos |
| CRM integration | Limited | HubSpot, Salesforce | Limited |
| Price (starter) | $32/mo | $40/mo | $27/mo |

---

## Recommended Production Stacks

### Solo Creator — Social Media Focus
| Category | Tool | Cost |
|----------|------|------|
| Recording | Screen Studio | $89 (one-time) |
| Editing | CapCut Pro | $8/month |
| Voiceover | CapCut TTS or own voice | Included |
| Music | Mixkit | Free |
| Subtitles | CapCut Auto-Captions | Included |
| **Total** | | **~$8/month + $89** |

### Startup — Product Marketing
| Category | Tool | Cost |
|----------|------|------|
| Recording | Screen Studio | $89 (one-time) |
| Editing | Descript Business | $33/month |
| Voiceover | ElevenLabs Creator | $22/month |
| Music | Epidemic Sound | $15/month |
| Mockups | Rotato | $49/year |
| Demos | Arcade Pro | $32/month |
| **Total** | | **~$106/month + $89** |

### Agency — Client Work
| Category | Tool | Cost |
|----------|------|------|
| Recording | Screen Studio + OBS | $89 (one-time) |
| Editing | DaVinci Resolve Studio | $295 (one-time) |
| Motion | After Effects | $23/month |
| Voiceover | ElevenLabs Pro | $99/month |
| Music | Artlist Max | $30/month |
| Mockups | Rotato (team) | $99/year |
| Color | DaVinci Resolve Color | Included |
| Audio | Fairlight (in DaVinci) | Included |
| **Total** | | **~$152/month + $384** |

### Enterprise — Full Production
| Category | Tool | Cost |
|----------|------|------|
| Recording | Screen Studio + Pro camera setup | $89 + $2,000+ |
| Editing | DaVinci Resolve Studio | $295 (one-time) |
| Motion | After Effects + Cinema 4D | $55/month |
| Voiceover | Professional VO + ElevenLabs Pro | $99/month + $200-500/video |
| Music | Musicbed or Artlist Enterprise | Custom pricing |
| Mockups | Rotato + custom 3D renders | $99/year |
| AI Avatars | HeyGen Business or Synthesia | $67-72/month |
| Interactive | Storylane Growth | $100/month |
| **Total** | | **$300+/month + variable** |
