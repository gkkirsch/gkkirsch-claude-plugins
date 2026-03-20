---
name: astro-content-collections
description: >
  Astro Content Collections, MDX integration, and content layer patterns.
  Use when building content-heavy sites, blogs, documentation portals,
  configuring content schemas, or optimizing images in Astro.
  Triggers: "astro content", "content collections", "astro mdx", "astro blog",
  "astro images", "astro schema", "defineCollection", "astro markdown".
  NOT for: SvelteKit content (see sveltekit-development), generic Markdown parsers, CMS integrations.
version: 1.0.0
allowed-tools: Read, Grep, Glob, Bash
---

# Astro Content Collections

## Collection Schema Definition

```typescript
// src/content/config.ts
import { defineCollection, z, reference } from 'astro:content';

const blog = defineCollection({
  type: 'content', // Markdown/MDX files
  schema: ({ image }) => z.object({
    title: z.string().max(100),
    description: z.string().max(200),
    pubDate: z.coerce.date(),
    updatedDate: z.coerce.date().optional(),
    heroImage: image().refine(img => img.width >= 1080, {
      message: 'Hero image must be at least 1080px wide',
    }),
    author: reference('authors'), // Reference another collection
    tags: z.array(z.string()).default([]),
    draft: z.boolean().default(false),
    category: z.enum(['tutorial', 'guide', 'announcement', 'case-study']),
  }),
});

const authors = defineCollection({
  type: 'data', // JSON/YAML files
  schema: ({ image }) => z.object({
    name: z.string(),
    bio: z.string().max(300),
    avatar: image(),
    socials: z.object({
      twitter: z.string().url().optional(),
      github: z.string().url().optional(),
    }).optional(),
  }),
});

const docs = defineCollection({
  type: 'content',
  schema: z.object({
    title: z.string(),
    section: z.enum(['getting-started', 'api', 'guides', 'reference']),
    order: z.number().int().positive(),
    sidebar: z.object({
      label: z.string().optional(),
      badge: z.enum(['new', 'updated', 'deprecated']).optional(),
      hidden: z.boolean().default(false),
    }).default({}),
  }),
});

export const collections = { blog, authors, docs };
```

## Content Directory Structure

```
src/content/
├── config.ts           # Collection schemas
├── blog/
│   ├── first-post.mdx  # Collection entries
│   ├── second-post.md
│   └── _drafts/        # Underscore prefix = excluded
├── authors/
│   ├── jane.json       # Data collection entries
│   └── john.yaml
└── docs/
    ├── getting-started/
    │   ├── 01-install.md
    │   └── 02-quick-start.md
    └── api/
        └── 01-endpoints.md
```

## Querying Collections

```astro
---
// src/pages/blog/index.astro
import { getCollection, getEntry } from 'astro:content';

// Get all published posts, sorted by date
const posts = (await getCollection('blog', ({ data }) => {
  return data.draft !== true; // Filter out drafts
})).sort((a, b) => b.data.pubDate.valueOf() - a.data.pubDate.valueOf());

// Get a single entry with full type safety
const featuredAuthor = await getEntry('authors', 'jane');

// Get entries by reference
const post = await getEntry('blog', 'first-post');
const author = await getEntry(post.data.author); // Resolves reference
---

<ul>
  {posts.map(post => (
    <li>
      <a href={`/blog/${post.slug}`}>
        <h2>{post.data.title}</h2>
        <time datetime={post.data.pubDate.toISOString()}>
          {post.data.pubDate.toLocaleDateString()}
        </time>
        <span>{post.data.category}</span>
      </a>
    </li>
  ))}
</ul>
```

## Dynamic Post Pages with MDX

```astro
---
// src/pages/blog/[...slug].astro
import { getCollection } from 'astro:content';
import BlogLayout from '../../layouts/BlogLayout.astro';
import { Image } from 'astro:assets';

export async function getStaticPaths() {
  const posts = await getCollection('blog', ({ data }) => !data.draft);
  return posts.map(post => ({
    params: { slug: post.slug },
    props: { post },
  }));
}

const { post } = Astro.props;
const { Content, headings, remarkPluginFrontmatter } = await post.render();
const author = await getEntry(post.data.author);
---

<BlogLayout title={post.data.title} description={post.data.description}>
  <article>
    <Image
      src={post.data.heroImage}
      alt={post.data.title}
      width={1200}
      height={630}
      format="avif"
      quality={80}
    />

    <h1>{post.data.title}</h1>
    <p>By {author.data.name}</p>
    <p>Reading time: {remarkPluginFrontmatter.readingTime} min</p>

    <!-- Table of Contents from headings -->
    <nav>
      <ul>
        {headings
          .filter(h => h.depth <= 3)
          .map(h => (
            <li style={`margin-left: ${(h.depth - 1) * 1}rem`}>
              <a href={`#${h.slug}`}>{h.text}</a>
            </li>
          ))}
      </ul>
    </nav>

    <!-- Rendered MDX content with custom components -->
    <Content components={{ img: Image }} />

    <footer>
      Tags: {post.data.tags.map(tag => (
        <a href={`/tags/${tag}`}>{tag}</a>
      ))}
    </footer>
  </article>
</BlogLayout>
```

## Image Optimization

```astro
---
// Astro's built-in image optimization
import { Image, Picture } from 'astro:assets';
import heroImage from '../assets/hero.png'; // Import = optimized at build

// Remote images need dimensions
const remoteImage = 'https://example.com/photo.jpg';
---

<!-- Local image: auto-optimized, type-safe dimensions -->
<Image
  src={heroImage}
  alt="Hero banner"
  width={1200}
  height={630}
  format="avif"
  quality={80}
  loading="eager"
/>

<!-- Picture element: multiple formats + responsive -->
<Picture
  src={heroImage}
  formats={['avif', 'webp']}
  widths={[400, 800, 1200]}
  sizes="(max-width: 800px) 100vw, 1200px"
  alt="Responsive hero"
/>

<!-- Remote image: must specify dimensions -->
<Image
  src={remoteImage}
  alt="Remote photo"
  width={800}
  height={600}
  inferSize={false}
/>

<!-- In content collections: images validated by schema -->
<!-- heroImage: image().refine(img => img.width >= 1080, ...) -->
```

## MDX Custom Components

```typescript
// src/components/mdx/Callout.astro
---
interface Props {
  type: 'info' | 'warning' | 'danger' | 'tip';
  title?: string;
}
const { type = 'info', title } = Astro.props;
const icons = { info: 'i', warning: '!', danger: 'x', tip: '*' };
---

<aside class={`callout callout-${type}`} role="note">
  <span class="callout-icon">{icons[type]}</span>
  {title && <strong class="callout-title">{title}</strong>}
  <div class="callout-content">
    <slot />
  </div>
</aside>
```

```mdx
{/* src/content/blog/my-post.mdx */}
---
title: "Getting Started"
---
import Callout from '../../components/mdx/Callout.astro';
import CodeBlock from '../../components/mdx/CodeBlock.astro';

# Getting Started

<Callout type="tip" title="Prerequisites">
  Make sure you have Node.js 18+ installed.
</Callout>

<CodeBlock lang="bash" title="Install">
npm create astro@latest
</CodeBlock>
```

## RSS Feed Generation

```typescript
// src/pages/rss.xml.ts
import rss from '@astrojs/rss';
import { getCollection } from 'astro:content';
import sanitizeHtml from 'sanitize-html';
import MarkdownIt from 'markdown-it';

const parser = new MarkdownIt();

export async function GET(context: { site: URL }) {
  const posts = await getCollection('blog', ({ data }) => !data.draft);

  return rss({
    title: 'My Blog',
    description: 'Tutorials and guides',
    site: context.site,
    items: posts
      .sort((a, b) => b.data.pubDate.valueOf() - a.data.pubDate.valueOf())
      .map(post => ({
        title: post.data.title,
        pubDate: post.data.pubDate,
        description: post.data.description,
        link: `/blog/${post.slug}/`,
        content: sanitizeHtml(parser.render(post.body), {
          allowedTags: sanitizeHtml.defaults.allowedTags.concat(['img']),
        }),
        categories: post.data.tags,
      })),
    customData: '<language>en-us</language>',
  });
}
```

## Gotchas

1. **Collection type mismatch** -- `type: 'content'` expects Markdown/MDX files, `type: 'data'` expects JSON/YAML. Putting a `.md` file in a data collection or a `.json` file in a content collection fails silently or with confusing errors. Check the collection type when entries aren't loading.

2. **Image imports in content collections** -- Images referenced in frontmatter must use relative paths from the markdown file (`./hero.png`), not from `src/assets/`. The `image()` schema helper resolves paths relative to the content file's location. Absolute paths or aliases (`~/assets/`) don't work in frontmatter.

3. **Slug generation from filenames** -- The `slug` property is auto-generated from the filename, not the frontmatter title. A file named `01-my-post.md` gets slug `01-my-post`. To control slugs, rename the file or use `getStaticPaths` to remap. You cannot override slugs in frontmatter.

4. **Draft filtering in production** -- Content collections don't auto-filter drafts. You must explicitly filter with `getCollection('blog', ({ data }) => !data.draft)`. Without this filter, draft posts are built and deployed. There's no build-time draft exclusion by default.

5. **MDX component imports are per-file** -- Each MDX file must import its own components. There's no global MDX component registry in Astro. If you use `<Callout>` in 50 posts, each post needs `import Callout from '...'`. Use remark/rehype plugins for truly global transformations instead.

6. **Reference resolution is async** -- `reference('authors')` in a schema only stores the ID, not the full entry. You must call `getEntry(post.data.author)` separately to resolve it. Forgetting this returns just the string ID instead of the author data object.
