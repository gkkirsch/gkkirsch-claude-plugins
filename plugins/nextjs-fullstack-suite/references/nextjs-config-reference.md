# Next.js Configuration Reference

Complete reference for next.config.js, environment variables, and TypeScript configuration.

---

## next.config.js Complete Options

```javascript
/** @type {import('next').NextConfig} */
const nextConfig = {
  // === OUTPUT ===
  output: 'standalone' | 'export',  // standalone for Docker, export for static
  distDir: '.next',                  // Build output directory (default: .next)

  // === ROUTING ===
  basePath: '/app',                  // URL prefix (e.g., example.com/app/page)
  trailingSlash: false,              // /about vs /about/
  skipTrailingSlashRedirect: false,   // Don't redirect /about/ to /about

  async redirects() { return [ /* { source, destination, permanent } */ ] },
  async rewrites() { return [ /* { source, destination } */ ] },
  async headers() { return [ /* { source, headers: [{key, value}] } */ ] },

  // === IMAGES ===
  images: {
    remotePatterns: [
      { protocol: 'https', hostname: '**.example.com', pathname: '/images/**' }
    ],
    formats: ['image/avif', 'image/webp'],     // Optimization formats
    deviceSizes: [640, 750, 828, 1080, 1200, 1920, 2048, 3840],
    imageSizes: [16, 32, 48, 64, 96, 128, 256, 384],
    minimumCacheTTL: 60,                         // Seconds
    unoptimized: false,                          // true for static export
    loader: 'default',                           // 'default', 'custom', 'akamai', 'cloudinary', 'imgix'
    loaderFile: './lib/image-loader.ts',          // Custom loader
  },

  // === PERFORMANCE ===
  compress: true,                     // gzip compression (disable if reverse proxy handles it)
  poweredByHeader: false,             // Remove X-Powered-By header
  reactStrictMode: true,              // React strict mode
  productionBrowserSourceMaps: false, // Source maps in production (increases build size)

  // === COMPILATION ===
  typescript: {
    ignoreBuildErrors: false,         // Skip type checking during build
    tsconfigPath: './tsconfig.json',
  },
  eslint: {
    ignoreDuringBuilds: false,        // Skip ESLint during build
    dirs: ['app', 'lib', 'components'],
  },

  // === WEBPACK ===
  webpack: (config, { buildId, dev, isServer, defaultLoaders, nextRuntime, webpack }) => {
    // Custom webpack configuration
    return config;
  },

  // === ENVIRONMENT ===
  env: {
    CUSTOM_VAR: 'value',              // Available as process.env.CUSTOM_VAR (build-time only)
  },

  // === EXPERIMENTAL ===
  experimental: {
    ppr: true,                        // Partial Prerendering
    serverActions: {
      bodySizeLimit: '2mb',           // Max body size for Server Actions
      allowedOrigins: ['my-proxy.example.com'],
    },
    optimizePackageImports: ['lucide-react', 'date-fns'],  // Auto tree-shake
  },

  // === INTERNATIONALIZATION ===
  // (i18n is limited in App Router — use middleware + route groups instead)

  // === TRANSPILATION ===
  transpilePackages: ['@my-org/ui', 'some-package'],  // Transpile specific npm packages

  // === TURBOPACK ===
  // turbopack: {}, // Config for next dev --turbopack
};
```

## TypeScript Configuration

```json
// tsconfig.json
{
  "compilerOptions": {
    "target": "ES2017",
    "lib": ["dom", "dom.iterable", "esnext"],
    "allowJs": true,
    "skipLibCheck": true,
    "strict": true,
    "noEmit": true,
    "esModuleInterop": true,
    "module": "esnext",
    "moduleResolution": "bundler",
    "resolveJsonModule": true,
    "isolatedModules": true,
    "jsx": "preserve",
    "incremental": true,
    "plugins": [{ "name": "next" }],
    "paths": {
      "@/*": ["./*"]
    }
  },
  "include": ["next-env.d.ts", "**/*.ts", "**/*.tsx", ".next/types/**/*.ts"],
  "exclude": ["node_modules"]
}
```

## File Conventions Reference

| File | Purpose | Runs On |
|------|---------|---------|
| `layout.tsx` | Shared UI, persists across navigations | Server |
| `template.tsx` | Shared UI, re-mounts on navigation | Server |
| `page.tsx` | Unique page UI | Server |
| `loading.tsx` | Loading UI (Suspense boundary) | Server |
| `error.tsx` | Error UI (error boundary) | Client |
| `not-found.tsx` | 404 UI | Server |
| `global-error.tsx` | Root error boundary | Client |
| `route.ts` | API endpoint | Server |
| `default.tsx` | Parallel route fallback | Server |
| `middleware.ts` | Request middleware | Edge |
| `instrumentation.ts` | OpenTelemetry setup | Server |
| `opengraph-image.tsx` | Dynamic OG images | Server |
| `icon.tsx` | Dynamic app icons | Server |
| `sitemap.ts` | Dynamic sitemap | Server |
| `robots.ts` | Dynamic robots.txt | Server |
| `manifest.ts` | PWA manifest | Server |

## Dynamic Functions (Opt into Dynamic Rendering)

Using any of these makes the page dynamic (not statically generated):

| Function | Purpose |
|----------|---------|
| `cookies()` | Read/set cookies |
| `headers()` | Read request headers |
| `searchParams` (page prop) | URL search parameters |
| `fetch(url, { cache: 'no-store' })` | Uncached fetch |
| `unstable_noStore()` | Explicit dynamic opt-in |

## Route Segment Config Exports

```typescript
// Any page.tsx or layout.tsx can export these:

export const dynamic = 'auto' | 'force-dynamic' | 'error' | 'force-static';
export const dynamicParams = true | false;
export const revalidate = false | 0 | number;  // seconds
export const fetchCache = 'auto' | 'default-cache' | 'only-cache' | 'force-cache' | 'force-no-store' | 'default-no-store' | 'only-no-store';
export const runtime = 'nodejs' | 'edge';
export const preferredRegion = 'auto' | 'global' | 'home' | string[];
export const maxDuration = 5;  // seconds (for serverless timeout)
```

## Metadata API

```typescript
// Static metadata
export const metadata: Metadata = {
  title: 'My App',
  description: 'Description',
  keywords: ['Next.js', 'React'],
  authors: [{ name: 'Author' }],
  creator: 'Author',
  metadataBase: new URL('https://myapp.com'),
  openGraph: {
    title: 'My App',
    description: 'Description',
    url: 'https://myapp.com',
    siteName: 'My App',
    images: [{ url: '/og.png', width: 1200, height: 630 }],
    locale: 'en_US',
    type: 'website',
  },
  twitter: {
    card: 'summary_large_image',
    title: 'My App',
    description: 'Description',
    images: ['/og.png'],
  },
  robots: {
    index: true,
    follow: true,
    googleBot: {
      index: true,
      follow: true,
    },
  },
  icons: {
    icon: '/favicon.ico',
    apple: '/apple-touch-icon.png',
  },
  manifest: '/manifest.json',
  verification: {
    google: 'verification-code',
  },
};

// Dynamic metadata
export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { slug } = await params;
  const post = await getPost(slug);
  return {
    title: post.title,
    description: post.excerpt,
    openGraph: { images: [post.coverImage] },
  };
}
```

## Common next.config.js Patterns

### Security Headers

```javascript
async headers() {
  return [{
    source: '/(.*)',
    headers: [
      { key: 'X-DNS-Prefetch-Control', value: 'on' },
      { key: 'Strict-Transport-Security', value: 'max-age=63072000; includeSubDomains; preload' },
      { key: 'X-Frame-Options', value: 'SAMEORIGIN' },
      { key: 'X-Content-Type-Options', value: 'nosniff' },
      { key: 'Referrer-Policy', value: 'strict-origin-when-cross-origin' },
      { key: 'Permissions-Policy', value: 'camera=(), microphone=(), geolocation=()' },
    ],
  }];
},
```

### Remove Trailing Slashes

```javascript
module.exports = {
  trailingSlash: false,
  skipTrailingSlashRedirect: false,
};
```

### Custom 404 / 500 Pages

```
app/not-found.tsx      → 404 page
app/global-error.tsx   → 500 page (root error boundary)
app/error.tsx          → Segment-level error page
```

### Import Aliases

```json
// tsconfig.json
{
  "compilerOptions": {
    "paths": {
      "@/*": ["./*"],
      "@/components/*": ["./components/*"],
      "@/lib/*": ["./lib/*"],
      "@/actions/*": ["./actions/*"]
    }
  }
}
```
