# Remotion Component Patterns

Reusable, production-ready animated components for Remotion videos. Each component is
self-contained with full TypeScript types. Copy them into your project and customize.

---

## 1. FadeIn / FadeOut Wrapper

Wraps any content with configurable fade-in and fade-out transitions.

```tsx
import { useCurrentFrame, useVideoConfig, interpolate, Easing } from 'remotion';
import React from 'react';

interface FadeProps {
  children: React.ReactNode;
  fadeInDuration?: number;  // frames
  fadeOutDuration?: number; // frames
  direction?: 'up' | 'down' | 'left' | 'right' | 'none';
  distance?: number;       // pixels to travel
  delay?: number;          // frames before fade starts
}

export const Fade: React.FC<FadeProps> = ({
  children,
  fadeInDuration = 20,
  fadeOutDuration = 15,
  direction = 'up',
  distance = 30,
  delay = 0,
}) => {
  const frame = useCurrentFrame();
  const { durationInFrames } = useVideoConfig();
  const adjustedFrame = frame - delay;

  // Fade in
  const fadeInOpacity = interpolate(
    adjustedFrame,
    [0, fadeInDuration],
    [0, 1],
    { extrapolateLeft: 'clamp', extrapolateRight: 'clamp' }
  );

  // Fade out (starts fadeOutDuration frames before the end)
  const fadeOutStart = durationInFrames - fadeOutDuration - delay;
  const fadeOutOpacity = interpolate(
    adjustedFrame,
    [fadeOutStart, fadeOutStart + fadeOutDuration],
    [1, 0],
    { extrapolateLeft: 'clamp', extrapolateRight: 'clamp' }
  );

  const opacity = Math.min(fadeInOpacity, fadeOutOpacity);

  // Directional slide
  const translateMap = {
    up: { x: 0, y: distance },
    down: { x: 0, y: -distance },
    left: { x: distance, y: 0 },
    right: { x: -distance, y: 0 },
    none: { x: 0, y: 0 },
  };

  const slideProgress = interpolate(
    adjustedFrame,
    [0, fadeInDuration],
    [1, 0],
    { extrapolateLeft: 'clamp', extrapolateRight: 'clamp', easing: Easing.out(Easing.cubic) }
  );

  const tx = translateMap[direction].x * slideProgress;
  const ty = translateMap[direction].y * slideProgress;

  if (adjustedFrame < 0) return null;

  return (
    <div style={{ opacity, transform: `translate(${tx}px, ${ty}px)` }}>
      {children}
    </div>
  );
};
```

---

## 2. TypewriterText

Reveals text character by character with a blinking cursor.

```tsx
import { useCurrentFrame, useVideoConfig, interpolate } from 'remotion';
import React from 'react';

interface TypewriterTextProps {
  text: string;
  charsPerFrame?: number;
  cursorColor?: string;
  fontSize?: number;
  fontFamily?: string;
  color?: string;
  startDelay?: number;
  showCursor?: boolean;
}

export const TypewriterText: React.FC<TypewriterTextProps> = ({
  text,
  charsPerFrame = 0.5,
  cursorColor = '#3b82f6',
  fontSize = 48,
  fontFamily = "'JetBrains Mono', monospace",
  color = '#ffffff',
  startDelay = 0,
  showCursor = true,
}) => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();

  const adjustedFrame = Math.max(0, frame - startDelay);
  const charsToShow = Math.min(
    Math.floor(adjustedFrame * charsPerFrame),
    text.length
  );
  const displayedText = text.slice(0, charsToShow);
  const isComplete = charsToShow >= text.length;

  // Cursor blinks every 0.5 seconds when typing is complete
  const cursorOpacity = isComplete
    ? Math.round(Math.sin(adjustedFrame / (fps * 0.25) * Math.PI) * 0.5 + 0.5)
    : 1;

  return (
    <div style={{ display: 'flex', alignItems: 'center' }}>
      <span
        style={{
          fontSize,
          fontFamily,
          color,
          whiteSpace: 'pre-wrap',
          letterSpacing: 0.5,
        }}
      >
        {displayedText}
      </span>
      {showCursor && (
        <span
          style={{
            display: 'inline-block',
            width: fontSize * 0.05,
            height: fontSize * 0.85,
            backgroundColor: cursorColor,
            marginLeft: 2,
            opacity: cursorOpacity,
            borderRadius: 1,
          }}
        />
      )}
    </div>
  );
};
```

---

## 3. AnimatedCounter

Counts from one number to another with optional formatting (commas, decimals, prefix, suffix).

```tsx
import { useCurrentFrame, interpolate, Easing } from 'remotion';
import React from 'react';

interface AnimatedCounterProps {
  from?: number;
  to: number;
  duration?: number;       // frames
  delay?: number;          // frames
  prefix?: string;         // e.g. "$"
  suffix?: string;         // e.g. "%"
  decimals?: number;
  useCommas?: boolean;
  fontSize?: number;
  color?: string;
  fontWeight?: number;
}

export const AnimatedCounter: React.FC<AnimatedCounterProps> = ({
  from = 0,
  to,
  duration = 60,
  delay = 0,
  prefix = '',
  suffix = '',
  decimals = 0,
  useCommas = true,
  fontSize = 72,
  color = '#ffffff',
  fontWeight = 700,
}) => {
  const frame = useCurrentFrame();
  const adjustedFrame = Math.max(0, frame - delay);

  const rawValue = interpolate(adjustedFrame, [0, duration], [from, to], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
    easing: Easing.out(Easing.cubic),
  });

  const formattedValue = useCommas
    ? rawValue.toLocaleString('en-US', {
        minimumFractionDigits: decimals,
        maximumFractionDigits: decimals,
      })
    : rawValue.toFixed(decimals);

  // Scale pop when animation completes
  const scale = interpolate(
    adjustedFrame,
    [duration - 5, duration, duration + 5],
    [1, 1.08, 1],
    { extrapolateLeft: 'clamp', extrapolateRight: 'clamp' }
  );

  return (
    <div
      style={{
        fontSize,
        fontWeight,
        color,
        fontFamily: 'Inter, system-ui, sans-serif',
        fontVariantNumeric: 'tabular-nums',
        transform: `scale(${scale})`,
        display: 'inline-block',
      }}
    >
      {prefix}{formattedValue}{suffix}
    </div>
  );
};
```

---

## 4. SlideInFromSide

Slides content in from any edge with spring physics.

```tsx
import { useCurrentFrame, useVideoConfig, spring } from 'remotion';
import React from 'react';

interface SlideInFromSideProps {
  children: React.ReactNode;
  from?: 'left' | 'right' | 'top' | 'bottom';
  delay?: number;
  stiffness?: number;
  damping?: number;
  distance?: number;
}

export const SlideInFromSide: React.FC<SlideInFromSideProps> = ({
  children,
  from = 'left',
  delay = 0,
  stiffness = 180,
  damping = 22,
  distance = 400,
}) => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();

  const progress = spring({
    frame: Math.max(0, frame - delay),
    fps,
    config: { stiffness, damping, mass: 1 },
  });

  const offsets = {
    left: { x: -distance * (1 - progress), y: 0 },
    right: { x: distance * (1 - progress), y: 0 },
    top: { x: 0, y: -distance * (1 - progress) },
    bottom: { x: 0, y: distance * (1 - progress) },
  };

  const { x, y } = offsets[from];
  const opacity = progress;

  return (
    <div
      style={{
        transform: `translate(${x}px, ${y}px)`,
        opacity,
      }}
    >
      {children}
    </div>
  );
};
```

---

## 5. CodeBlock with Syntax Highlighting

Displays code with line-by-line reveal animation and optional line highlighting.

```tsx
import { useCurrentFrame, interpolate, Easing } from 'remotion';
import React from 'react';

interface CodeBlockProps {
  code: string;
  language?: string;
  linesPerFrame?: number;
  highlightLines?: number[];
  fontSize?: number;
  backgroundColor?: string;
  textColor?: string;
  highlightColor?: string;
  lineNumberColor?: string;
  showLineNumbers?: boolean;
  startDelay?: number;
}

export const CodeBlock: React.FC<CodeBlockProps> = ({
  code,
  language = 'typescript',
  linesPerFrame = 0.15,
  highlightLines = [],
  fontSize = 22,
  backgroundColor = '#1e1e2e',
  textColor = '#cdd6f4',
  highlightColor = 'rgba(137, 180, 250, 0.12)',
  lineNumberColor = '#585b70',
  showLineNumbers = true,
  startDelay = 0,
}) => {
  const frame = useCurrentFrame();
  const adjustedFrame = Math.max(0, frame - startDelay);
  const lines = code.split('\n');
  const visibleLineCount = Math.min(
    Math.floor(adjustedFrame * linesPerFrame),
    lines.length
  );

  return (
    <div
      style={{
        backgroundColor,
        borderRadius: 16,
        padding: '28px 32px',
        fontFamily: "'JetBrains Mono', 'Fira Code', 'Cascadia Code', monospace",
        fontSize,
        lineHeight: 1.7,
        overflow: 'hidden',
        boxShadow: '0 25px 50px rgba(0,0,0,0.4)',
        border: '1px solid rgba(255,255,255,0.08)',
      }}
    >
      {/* Language badge */}
      <div
        style={{
          fontSize: fontSize * 0.55,
          color: lineNumberColor,
          marginBottom: 16,
          textTransform: 'uppercase',
          letterSpacing: 1.5,
          fontWeight: 600,
        }}
      >
        {language}
      </div>

      {lines.slice(0, visibleLineCount).map((line, i) => {
        const lineFrame = i / linesPerFrame;
        const lineAge = adjustedFrame - lineFrame;
        const lineOpacity = interpolate(lineAge, [0, 8], [0, 1], {
          extrapolateLeft: 'clamp',
          extrapolateRight: 'clamp',
        });
        const lineTranslateX = interpolate(lineAge, [0, 8], [12, 0], {
          extrapolateLeft: 'clamp',
          extrapolateRight: 'clamp',
          easing: Easing.out(Easing.cubic),
        });

        const isHighlighted = highlightLines.includes(i + 1);

        return (
          <div
            key={i}
            style={{
              opacity: lineOpacity,
              transform: `translateX(${lineTranslateX}px)`,
              display: 'flex',
              backgroundColor: isHighlighted ? highlightColor : 'transparent',
              margin: isHighlighted ? '0 -32px' : 0,
              padding: isHighlighted ? '0 32px' : 0,
              borderLeft: isHighlighted ? '3px solid #89b4fa' : '3px solid transparent',
            }}
          >
            {showLineNumbers && (
              <span
                style={{
                  color: lineNumberColor,
                  minWidth: 40,
                  textAlign: 'right',
                  marginRight: 24,
                  userSelect: 'none',
                  fontSize: fontSize * 0.85,
                }}
              >
                {i + 1}
              </span>
            )}
            <span style={{ color: textColor, whiteSpace: 'pre' }}>{line}</span>
          </div>
        );
      })}
    </div>
  );
};
```

---

## 6. ProductScreenshot with Device Frame

Displays a screenshot inside a device mockup (laptop, phone, or browser window).

```tsx
import {
  useCurrentFrame,
  useVideoConfig,
  spring,
  interpolate,
  Img,
  staticFile,
} from 'remotion';
import React from 'react';

interface ProductScreenshotProps {
  src: string;
  device?: 'browser' | 'laptop' | 'phone';
  delay?: number;
  browserTitle?: string;
  browserUrl?: string;
}

export const ProductScreenshot: React.FC<ProductScreenshotProps> = ({
  src,
  device = 'browser',
  delay = 0,
  browserTitle = 'My App',
  browserUrl = 'app.example.com',
}) => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();

  const scaleProgress = spring({
    frame: Math.max(0, frame - delay),
    fps,
    config: { stiffness: 120, damping: 18, mass: 0.8 },
  });

  const shadowOpacity = interpolate(scaleProgress, [0, 1], [0, 0.3]);

  const scale = interpolate(scaleProgress, [0, 1], [0.85, 1]);
  const opacity = interpolate(scaleProgress, [0, 0.3], [0, 1], {
    extrapolateRight: 'clamp',
  });

  if (device === 'browser') {
    return (
      <div
        style={{
          transform: `scale(${scale})`,
          opacity,
          filter: `drop-shadow(0 30px 60px rgba(0,0,0,${shadowOpacity}))`,
        }}
      >
        {/* Browser chrome */}
        <div
          style={{
            backgroundColor: '#2a2a3e',
            borderRadius: '12px 12px 0 0',
            padding: '12px 16px',
            display: 'flex',
            alignItems: 'center',
            gap: 8,
          }}
        >
          {/* Traffic lights */}
          <div style={{ display: 'flex', gap: 6 }}>
            <div style={{ width: 12, height: 12, borderRadius: '50%', backgroundColor: '#ff5f57' }} />
            <div style={{ width: 12, height: 12, borderRadius: '50%', backgroundColor: '#ffbd2e' }} />
            <div style={{ width: 12, height: 12, borderRadius: '50%', backgroundColor: '#28c840' }} />
          </div>

          {/* URL bar */}
          <div
            style={{
              flex: 1,
              backgroundColor: '#1a1a2e',
              borderRadius: 6,
              padding: '6px 12px',
              fontSize: 14,
              color: '#888',
              fontFamily: 'system-ui',
              marginLeft: 8,
            }}
          >
            {browserUrl}
          </div>
        </div>

        {/* Screenshot */}
        <div style={{ borderRadius: '0 0 12px 12px', overflow: 'hidden' }}>
          <Img src={staticFile(src)} style={{ width: '100%', display: 'block' }} />
        </div>
      </div>
    );
  }

  if (device === 'phone') {
    return (
      <div
        style={{
          transform: `scale(${scale})`,
          opacity,
          filter: `drop-shadow(0 30px 60px rgba(0,0,0,${shadowOpacity}))`,
          backgroundColor: '#1a1a1a',
          borderRadius: 40,
          padding: '14px 10px',
          width: 375,
        }}
      >
        {/* Notch */}
        <div
          style={{
            width: 120,
            height: 28,
            backgroundColor: '#1a1a1a',
            borderRadius: '0 0 16px 16px',
            margin: '0 auto',
            position: 'relative',
            top: -14,
            zIndex: 2,
          }}
        />

        <div style={{ borderRadius: 28, overflow: 'hidden', marginTop: -14 }}>
          <Img src={staticFile(src)} style={{ width: '100%', display: 'block' }} />
        </div>

        {/* Home indicator */}
        <div
          style={{
            width: 120,
            height: 4,
            backgroundColor: '#666',
            borderRadius: 2,
            margin: '10px auto 4px',
          }}
        />
      </div>
    );
  }

  // Laptop fallback
  return (
    <div style={{ transform: `scale(${scale})`, opacity }}>
      <Img src={staticFile(src)} style={{ width: '100%', borderRadius: 8 }} />
    </div>
  );
};
```

---

## 7. SplitScreenCompare

Side-by-side comparison with an animated divider.

```tsx
import { useCurrentFrame, useVideoConfig, spring, interpolate, AbsoluteFill } from 'remotion';
import React from 'react';

interface SplitScreenCompareProps {
  leftContent: React.ReactNode;
  rightContent: React.ReactNode;
  leftLabel?: string;
  rightLabel?: string;
  dividerColor?: string;
  delay?: number;
}

export const SplitScreenCompare: React.FC<SplitScreenCompareProps> = ({
  leftContent,
  rightContent,
  leftLabel = 'Before',
  rightLabel = 'After',
  dividerColor = '#3b82f6',
  delay = 0,
}) => {
  const frame = useCurrentFrame();
  const { fps, width } = useVideoConfig();

  // Divider slides from left to center
  const dividerProgress = spring({
    frame: Math.max(0, frame - delay),
    fps,
    config: { stiffness: 80, damping: 20 },
  });

  const dividerX = interpolate(dividerProgress, [0, 1], [0, width / 2]);

  // Labels fade in after divider reaches center
  const labelOpacity = interpolate(
    frame - delay,
    [30, 45],
    [0, 1],
    { extrapolateLeft: 'clamp', extrapolateRight: 'clamp' }
  );

  return (
    <AbsoluteFill>
      {/* Left side */}
      <div
        style={{
          position: 'absolute',
          left: 0,
          top: 0,
          width: dividerX,
          height: '100%',
          overflow: 'hidden',
        }}
      >
        <div style={{ width, height: '100%' }}>
          {leftContent}
        </div>
      </div>

      {/* Right side */}
      <div
        style={{
          position: 'absolute',
          left: dividerX,
          top: 0,
          width: width - dividerX,
          height: '100%',
          overflow: 'hidden',
        }}
      >
        <div style={{ width, height: '100%', marginLeft: -dividerX }}>
          {rightContent}
        </div>
      </div>

      {/* Divider line */}
      <div
        style={{
          position: 'absolute',
          left: dividerX - 2,
          top: 0,
          width: 4,
          height: '100%',
          backgroundColor: dividerColor,
          zIndex: 10,
          boxShadow: `0 0 20px ${dividerColor}`,
        }}
      />

      {/* Labels */}
      <div
        style={{
          position: 'absolute',
          left: 40,
          top: 40,
          opacity: labelOpacity,
          backgroundColor: 'rgba(0,0,0,0.7)',
          padding: '8px 20px',
          borderRadius: 8,
          color: '#fff',
          fontSize: 20,
          fontWeight: 600,
          fontFamily: 'Inter, system-ui',
        }}
      >
        {leftLabel}
      </div>
      <div
        style={{
          position: 'absolute',
          right: 40,
          top: 40,
          opacity: labelOpacity,
          backgroundColor: 'rgba(0,0,0,0.7)',
          padding: '8px 20px',
          borderRadius: 8,
          color: '#fff',
          fontSize: 20,
          fontWeight: 600,
          fontFamily: 'Inter, system-ui',
        }}
      >
        {rightLabel}
      </div>
    </AbsoluteFill>
  );
};
```

---

## 8. ParticleBackground

Floating particle field with configurable density, speed, and color.

```tsx
import { useCurrentFrame, useVideoConfig, interpolate } from 'remotion';
import React, { useMemo } from 'react';

interface Particle {
  x: number;
  y: number;
  size: number;
  speed: number;
  opacity: number;
  angle: number;
}

interface ParticleBackgroundProps {
  count?: number;
  color?: string;
  maxSize?: number;
  speed?: number;
  fadeIn?: boolean;
}

// Deterministic pseudo-random for consistent renders
const seededRandom = (seed: number) => {
  const x = Math.sin(seed * 9301 + 49297) * 49297;
  return x - Math.floor(x);
};

export const ParticleBackground: React.FC<ParticleBackgroundProps> = ({
  count = 60,
  color = '#3b82f6',
  maxSize = 6,
  speed = 0.5,
  fadeIn = true,
}) => {
  const frame = useCurrentFrame();
  const { width, height, durationInFrames } = useVideoConfig();

  // Generate particles once (deterministic from seed, not Math.random)
  const particles = useMemo<Particle[]>(() => {
    return Array.from({ length: count }, (_, i) => ({
      x: seededRandom(i * 7 + 1) * width,
      y: seededRandom(i * 13 + 3) * height,
      size: seededRandom(i * 17 + 5) * maxSize + 1,
      speed: (seededRandom(i * 23 + 7) * 0.5 + 0.5) * speed,
      opacity: seededRandom(i * 29 + 11) * 0.5 + 0.2,
      angle: seededRandom(i * 31 + 13) * Math.PI * 2,
    }));
  }, [count, width, height, maxSize, speed]);

  const containerOpacity = fadeIn
    ? interpolate(frame, [0, 30], [0, 1], { extrapolateRight: 'clamp' })
    : 1;

  return (
    <div
      style={{
        position: 'absolute',
        inset: 0,
        overflow: 'hidden',
        opacity: containerOpacity,
      }}
    >
      {particles.map((p, i) => {
        const dx = Math.cos(p.angle) * p.speed * frame;
        const dy = Math.sin(p.angle) * p.speed * frame;
        const px = ((p.x + dx) % (width + maxSize * 2)) - maxSize;
        const py = ((p.y + dy) % (height + maxSize * 2)) - maxSize;

        // Gentle pulse
        const pulseOpacity =
          p.opacity +
          Math.sin(frame * 0.05 + i) * 0.15;

        return (
          <div
            key={i}
            style={{
              position: 'absolute',
              left: px,
              top: py,
              width: p.size,
              height: p.size,
              borderRadius: '50%',
              backgroundColor: color,
              opacity: Math.max(0, Math.min(1, pulseOpacity)),
              filter: p.size > maxSize * 0.7 ? `blur(${p.size * 0.3}px)` : 'none',
            }}
          />
        );
      })}
    </div>
  );
};
```

---

## 9. ProgressBar

Animated progress bar with optional percentage label and milestone markers.

```tsx
import { useCurrentFrame, interpolate, Easing } from 'remotion';
import React from 'react';

interface Milestone {
  at: number;   // 0-100
  label: string;
}

interface ProgressBarProps {
  value: number;           // target percentage 0-100
  duration?: number;       // frames to animate
  delay?: number;
  color?: string;
  backgroundColor?: string;
  height?: number;
  width?: number;
  showLabel?: boolean;
  milestones?: Milestone[];
  borderRadius?: number;
}

export const ProgressBar: React.FC<ProgressBarProps> = ({
  value,
  duration = 60,
  delay = 0,
  color = '#3b82f6',
  backgroundColor = 'rgba(255,255,255,0.1)',
  height = 16,
  width = 800,
  showLabel = true,
  milestones = [],
  borderRadius = 999,
}) => {
  const frame = useCurrentFrame();
  const adjustedFrame = Math.max(0, frame - delay);

  const progress = interpolate(adjustedFrame, [0, duration], [0, value], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
    easing: Easing.out(Easing.cubic),
  });

  const labelOpacity = interpolate(adjustedFrame, [0, 15], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });

  return (
    <div style={{ width, position: 'relative' }}>
      {/* Label */}
      {showLabel && (
        <div
          style={{
            opacity: labelOpacity,
            fontSize: 24,
            fontWeight: 700,
            color: '#fff',
            marginBottom: 12,
            fontFamily: 'Inter, system-ui',
            fontVariantNumeric: 'tabular-nums',
          }}
        >
          {Math.round(progress)}%
        </div>
      )}

      {/* Track */}
      <div
        style={{
          width: '100%',
          height,
          backgroundColor,
          borderRadius,
          overflow: 'hidden',
          position: 'relative',
        }}
      >
        {/* Fill */}
        <div
          style={{
            width: `${progress}%`,
            height: '100%',
            backgroundColor: color,
            borderRadius,
            transition: 'none',
            boxShadow: `0 0 20px ${color}40`,
          }}
        />
      </div>

      {/* Milestone markers */}
      {milestones.map((m, i) => {
        const markerOpacity = progress >= m.at ? 1 : 0.3;
        return (
          <div
            key={i}
            style={{
              position: 'absolute',
              left: `${m.at}%`,
              bottom: -28,
              transform: 'translateX(-50%)',
              fontSize: 14,
              color: '#94a3b8',
              opacity: markerOpacity,
              fontFamily: 'Inter, system-ui',
              whiteSpace: 'nowrap',
            }}
          >
            {m.label}
          </div>
        );
      })}
    </div>
  );
};
```

---

## 10. LogoReveal

Reveals a logo with a mask wipe, scale bounce, or particle burst effect.

```tsx
import {
  useCurrentFrame,
  useVideoConfig,
  spring,
  interpolate,
  Img,
  staticFile,
  Easing,
} from 'remotion';
import React from 'react';

interface LogoRevealProps {
  src: string;
  effect?: 'scale-bounce' | 'wipe' | 'blur-in';
  size?: number;
  delay?: number;
  backgroundColor?: string;
}

export const LogoReveal: React.FC<LogoRevealProps> = ({
  src,
  effect = 'scale-bounce',
  size = 200,
  delay = 0,
  backgroundColor = 'transparent',
}) => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();
  const adjustedFrame = Math.max(0, frame - delay);

  let style: React.CSSProperties = {};

  if (effect === 'scale-bounce') {
    const scaleRaw = spring({
      frame: adjustedFrame,
      fps,
      config: { stiffness: 200, damping: 12, mass: 0.6 },
    });
    const opacity = interpolate(adjustedFrame, [0, 8], [0, 1], {
      extrapolateRight: 'clamp',
    });
    style = {
      transform: `scale(${scaleRaw})`,
      opacity,
    };
  }

  if (effect === 'wipe') {
    const wipeProgress = interpolate(adjustedFrame, [0, 30], [0, 100], {
      extrapolateLeft: 'clamp',
      extrapolateRight: 'clamp',
      easing: Easing.out(Easing.cubic),
    });
    style = {
      clipPath: `inset(0 ${100 - wipeProgress}% 0 0)`,
      opacity: 1,
    };
  }

  if (effect === 'blur-in') {
    const blurAmount = interpolate(adjustedFrame, [0, 25], [20, 0], {
      extrapolateLeft: 'clamp',
      extrapolateRight: 'clamp',
    });
    const opacity = interpolate(adjustedFrame, [0, 20], [0, 1], {
      extrapolateLeft: 'clamp',
      extrapolateRight: 'clamp',
    });
    const scale = interpolate(adjustedFrame, [0, 25], [1.15, 1], {
      extrapolateLeft: 'clamp',
      extrapolateRight: 'clamp',
      easing: Easing.out(Easing.cubic),
    });
    style = {
      filter: `blur(${blurAmount}px)`,
      opacity,
      transform: `scale(${scale})`,
    };
  }

  return (
    <div
      style={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        backgroundColor,
        ...style,
      }}
    >
      <Img
        src={staticFile(src)}
        style={{ width: size, height: size, objectFit: 'contain' }}
      />
    </div>
  );
};
```

---

## 11. GradientBackground

Animated gradient that shifts colors over time.

```tsx
import { useCurrentFrame, interpolate } from 'remotion';
import React from 'react';

interface GradientBackgroundProps {
  colors?: [string, string, string];
  angle?: number;
  speed?: number;
}

export const GradientBackground: React.FC<GradientBackgroundProps> = ({
  colors = ['#0f172a', '#1e3a5f', '#0f172a'],
  angle = 135,
  speed = 0.3,
}) => {
  const frame = useCurrentFrame();

  // Shift gradient position over time
  const position = frame * speed;
  const gradientAngle = angle + Math.sin(frame * 0.02) * 10;

  return (
    <div
      style={{
        position: 'absolute',
        inset: 0,
        background: `linear-gradient(
          ${gradientAngle}deg,
          ${colors[0]} ${0 + position}%,
          ${colors[1]} ${50 + Math.sin(frame * 0.03) * 15}%,
          ${colors[2]} ${100 + position}%
        )`,
        backgroundSize: '200% 200%',
        backgroundPosition: `${50 + Math.sin(frame * 0.015) * 30}% ${50 + Math.cos(frame * 0.02) * 30}%`,
      }}
    />
  );
};
```

---

## 12. StaggeredList

Renders a list of items with staggered entrance animations.

```tsx
import { useCurrentFrame, useVideoConfig, spring, interpolate } from 'remotion';
import React from 'react';

interface StaggeredListProps {
  items: Array<{
    icon?: string;
    title: string;
    description?: string;
  }>;
  staggerDelay?: number;    // frames between each item
  startDelay?: number;
  itemHeight?: number;
  accentColor?: string;
}

export const StaggeredList: React.FC<StaggeredListProps> = ({
  items,
  staggerDelay = 8,
  startDelay = 0,
  itemHeight = 80,
  accentColor = '#3b82f6',
}) => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
      {items.map((item, index) => {
        const itemDelay = startDelay + index * staggerDelay;
        const itemFrame = Math.max(0, frame - itemDelay);

        const progress = spring({
          frame: itemFrame,
          fps,
          config: { stiffness: 200, damping: 22 },
        });

        const opacity = progress;
        const translateX = interpolate(progress, [0, 1], [60, 0]);

        return (
          <div
            key={index}
            style={{
              opacity,
              transform: `translateX(${translateX}px)`,
              display: 'flex',
              alignItems: 'center',
              gap: 20,
              height: itemHeight,
              padding: '0 24px',
              backgroundColor: 'rgba(255,255,255,0.04)',
              borderRadius: 12,
              borderLeft: `3px solid ${accentColor}`,
            }}
          >
            {item.icon && (
              <span style={{ fontSize: 32 }}>{item.icon}</span>
            )}
            <div>
              <div
                style={{
                  fontSize: 22,
                  fontWeight: 600,
                  color: '#fff',
                  fontFamily: 'Inter, system-ui',
                }}
              >
                {item.title}
              </div>
              {item.description && (
                <div
                  style={{
                    fontSize: 16,
                    color: '#94a3b8',
                    marginTop: 4,
                    fontFamily: 'Inter, system-ui',
                  }}
                >
                  {item.description}
                </div>
              )}
            </div>
          </div>
        );
      })}
    </div>
  );
};
```

---

## Usage in a Composition

Here is how you combine these components in a full video:

```tsx
import { Composition, Sequence, AbsoluteFill } from 'remotion';
import { Fade } from './components/Fade';
import { TypewriterText } from './components/TypewriterText';
import { AnimatedCounter } from './components/AnimatedCounter';
import { ProductScreenshot } from './components/ProductScreenshot';
import { ParticleBackground } from './components/ParticleBackground';
import { LogoReveal } from './components/LogoReveal';
import { StaggeredList } from './components/StaggeredList';

export const ProductLaunchVideo: React.FC = () => {
  return (
    <AbsoluteFill style={{ backgroundColor: '#0f172a' }}>
      <ParticleBackground count={40} color="#3b82f6" />

      {/* Scene 1: Logo reveal */}
      <Sequence from={0} durationInFrames={90}>
        <AbsoluteFill style={{ display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
          <LogoReveal src="logo.png" effect="scale-bounce" size={180} />
        </AbsoluteFill>
      </Sequence>

      {/* Scene 2: Tagline */}
      <Sequence from={60} durationInFrames={120}>
        <AbsoluteFill style={{ display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
          <Fade direction="up">
            <TypewriterText text="Ship 10x faster with AI" fontSize={64} />
          </Fade>
        </AbsoluteFill>
      </Sequence>

      {/* Scene 3: Metrics */}
      <Sequence from={150} durationInFrames={120}>
        <AbsoluteFill style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 80 }}>
          <div style={{ textAlign: 'center' }}>
            <AnimatedCounter to={50000} prefix="$" suffix="/mo" duration={45} />
            <div style={{ color: '#94a3b8', fontSize: 20, marginTop: 8 }}>Revenue</div>
          </div>
          <div style={{ textAlign: 'center' }}>
            <AnimatedCounter to={99.9} suffix="%" decimals={1} duration={45} delay={10} />
            <div style={{ color: '#94a3b8', fontSize: 20, marginTop: 8 }}>Uptime</div>
          </div>
        </AbsoluteFill>
      </Sequence>

      {/* Scene 4: Product screenshot */}
      <Sequence from={240} durationInFrames={120}>
        <AbsoluteFill style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', padding: 80 }}>
          <ProductScreenshot src="screenshots/dashboard.png" device="browser" />
        </AbsoluteFill>
      </Sequence>
    </AbsoluteFill>
  );
};
```
