---
name: seo-keyword-research
description: >
  SEO keyword research and content optimization patterns.
  Use when identifying target keywords, analyzing search intent,
  building topic clusters, optimizing content for search, or
  creating keyword-driven content strategies.
  Triggers: "keyword research", "SEO keywords", "search intent",
  "topic cluster", "keyword difficulty", "content gap", "SERP analysis",
  "long tail keywords", "keyword mapping".
  NOT for: technical SEO (sitemaps, robots.txt), link building, paid search/SEM.
version: 1.0.0
allowed-tools: Read, Grep, Glob, Bash
---

# SEO Keyword Research

## Keyword Research Framework

```markdown
# Keyword Research Process

## Step 1: Seed Keywords
Start with 5-10 seed terms from:
- Your product/service categories
- Problems your audience searches for
- Competitor page titles and H1s
- Customer support tickets (actual language used)
- Google Search Console → Performance → Queries (existing rankings)

## Step 2: Expand with Modifiers
For each seed keyword, generate variations:

| Modifier Type | Pattern | Example (seed: "project management") |
|--------------|---------|---------------------------------------|
| How-to | "how to [seed]" | how to project management for remote teams |
| Best | "best [seed] [qualifier]" | best project management software for startups |
| Comparison | "[seed] vs [competitor]" | project management vs task management |
| Year | "[seed] [year]" | project management trends 2026 |
| For audience | "[seed] for [persona]" | project management for freelancers |
| Alternative | "[seed] alternative" | monday alternative |
| Template | "[seed] template" | project management template excel |
| Tool | "[seed] tool" | free project management tool |

## Step 3: Classify Search Intent
Every keyword has one primary intent:

| Intent | Signal Words | Content Type | Example |
|--------|-------------|--------------|---------|
| Informational | how, what, why, guide, tutorial | Blog post, guide | "what is agile project management" |
| Commercial | best, top, review, compare, vs | Comparison, listicle | "best project management tools 2026" |
| Transactional | buy, pricing, discount, sign up | Landing page, pricing | "project management software pricing" |
| Navigational | [brand name], login, docs | Homepage, docs | "asana login" |

## Step 4: Prioritize
Score each keyword on a 1-5 scale:

| Factor | Weight | How to Assess |
|--------|--------|---------------|
| Relevance | 30% | Does this directly relate to what you sell? |
| Volume | 25% | Monthly search volume (use Ahrefs, SEMrush, or Google Keyword Planner) |
| Difficulty | 25% | Can you realistically rank? (DR < 40 = easier) |
| Intent match | 20% | Does the intent align with your content/product? |

Priority Score = (Relevance * 0.3) + (Volume * 0.25) + ((6 - Difficulty) * 0.25) + (Intent * 0.2)
```

## Topic Cluster Architecture

```markdown
# Topic Cluster Model

A topic cluster = 1 pillar page + 5-15 cluster pages + internal links.

## Structure

Pillar Page (broad, high-volume keyword):
  "The Complete Guide to Remote Work"
  └── URL: /remote-work-guide
  └── 3,000-5,000 words
  └── Links TO every cluster page

Cluster Pages (specific, long-tail keywords):
  ├── "Best Video Conferencing Tools for Remote Teams" → /remote-work/video-conferencing
  ├── "How to Set Up a Home Office on a Budget" → /remote-work/home-office-setup
  ├── "Remote Work Time Management Strategies" → /remote-work/time-management
  ├── "Managing Remote Teams Across Time Zones" → /remote-work/time-zones
  ├── "Remote Work Tax Implications by State" → /remote-work/tax-implications
  └── Each links BACK to pillar page

## URL Structure Rules
- Pillar: /[topic]/
- Cluster: /[topic]/[subtopic]/
- Keep URLs to 3-4 path segments max
- Use hyphens, lowercase, no special characters
- Don't change URLs after publishing (or 301 redirect)

## Internal Linking Rules
- Every cluster page links to pillar (exact or partial match anchor text)
- Pillar links to every cluster page (descriptive anchor text)
- Cluster pages link to 2-3 related cluster pages
- Don't link to irrelevant pages just for link juice
- Use descriptive anchor text, not "click here" or "read more"
```

## Content Brief Template

```markdown
# Content Brief: [Target Keyword]

## Keyword Data
- **Primary keyword**: [keyword] (volume: X/mo, KD: Y)
- **Secondary keywords**: [list 3-5]
- **Search intent**: Informational / Commercial / Transactional
- **Target URL**: /path/to/page

## SERP Analysis
| Rank | Title | Word Count | Content Type | Unique Angle |
|------|-------|------------|-------------|--------------|
| 1 | [title] | [count] | [type] | [what they cover uniquely] |
| 2 | [title] | [count] | [type] | [what they cover uniquely] |
| 3 | [title] | [count] | [type] | [what they cover uniquely] |

**Content gap**: What do top 3 results NOT cover that searchers need?
**Featured snippet opportunity**: Is there a featured snippet? What format? (list, table, paragraph)
**People Also Ask**: [list 4-5 PAA questions]

## Content Requirements
- **Title tag**: [60 chars max, primary keyword near front]
- **Meta description**: [155 chars max, include primary keyword + CTA]
- **H1**: [matches search intent, includes primary keyword]
- **Word count target**: [based on SERP average + 20%]
- **Content format**: [guide / listicle / how-to / comparison]

## Outline
1. Introduction (hook + what reader will learn)
2. [Section matching PAA question 1]
3. [Section matching PAA question 2]
4. [Core content sections based on SERP gap analysis]
5. [Section with original data/insight competitors lack]
6. Conclusion + CTA

## On-Page SEO Checklist
- [ ] Primary keyword in title tag, H1, first 100 words, and URL
- [ ] Secondary keywords in H2/H3 subheadings
- [ ] Images with descriptive alt text (include keyword naturally)
- [ ] Internal links to pillar page and 2-3 related cluster pages
- [ ] External links to 2-3 authoritative sources
- [ ] FAQ section targeting PAA questions (use <details> for accordion)
- [ ] Table of contents for posts > 1,500 words
```

## Keyword Mapping Spreadsheet

```markdown
# Keyword Map Structure

| Keyword | Volume | KD | Intent | Assigned URL | Content Type | Status | Cluster |
|---------|--------|-----|--------|-------------|-------------|--------|---------|
| project management software | 12,000 | 72 | Commercial | /tools | Comparison | Published | PM Pillar |
| free project management tools | 8,100 | 45 | Commercial | /tools/free | Listicle | Draft | PM Pillar |
| how to manage a project | 3,600 | 28 | Informational | /guide | Guide | Planned | PM Pillar |
| gantt chart template | 6,500 | 35 | Transactional | /templates/gantt | Template | Published | Templates |

## Rules for Keyword Mapping
1. **One primary keyword per URL** — never target the same keyword on two pages (cannibalization)
2. **Group by intent** — informational keywords go to blog posts, commercial to comparison pages, transactional to product/landing pages
3. **Map to existing content first** — before creating new pages, check if an existing page can be optimized
4. **Track status** — Planned → Draft → Published → Optimized → Monitoring
5. **Cluster assignment** — every page belongs to exactly one topic cluster
```

## Content Refresh Strategy

```markdown
# When and How to Refresh Content

## Identify Refresh Candidates
- Pages ranking positions 5-20 (close to page 1, need a push)
- Pages with declining traffic (compare 3-month periods in GSC)
- Pages older than 12 months with dated information
- Pages ranking for keywords they don't explicitly target

## Refresh Checklist
1. Update the publication date (only after making real changes)
2. Refresh statistics, screenshots, and examples with current data
3. Add new sections addressing PAA questions you didn't cover originally
4. Remove outdated advice or references to deprecated tools/features
5. Improve internal linking (link to new cluster pages published since original)
6. Add/update structured data (FAQ schema, HowTo schema)
7. Optimize title tag and meta description (test new CTAs)
8. Compress and update images (use WebP, add descriptive alt text)

## Don't Do
- Don't change the URL (or 301 redirect if you must)
- Don't remove content that currently ranks for related keywords
- Don't update the date without making substantive content changes
- Don't keyword-stuff sections that read naturally
```

## Gotchas

1. **Keyword cannibalization** -- Two pages targeting the same keyword split ranking signals and neither ranks well. Audit your site for duplicate intent: search `site:yourdomain.com "keyword"` and check if multiple pages appear. Fix by merging pages or differentiating their target keywords.

2. **Volume is not value** -- A keyword with 100 monthly searches and clear purchase intent ("buy project management annual plan") converts better than one with 10,000 searches and informational intent ("what is project management"). Prioritize bottom-of-funnel keywords for revenue, top-of-funnel for brand.

3. **Ignoring search intent mismatch** -- If the SERP shows all listicles and you publish a 5,000-word guide, Google won't rank it. Match the dominant content format on page 1. Check: are top results guides, listicles, tools, videos, or product pages?

4. **Chasing zero-volume keywords blindly** -- Zero-volume in tools doesn't mean zero searches. New and niche queries often have no data. If the topic is relevant and has clear intent, write it. But verify with "allintitle:" search to check competition.

5. **Not tracking keyword rankings after publishing** -- Publishing is step 1. Track rankings weekly for the first 3 months. If not ranking after 3 months, the content needs improvement (more depth, better internal links, or a different angle). Don't just publish and forget.

6. **Over-optimizing anchor text** -- Using exact-match anchor text for every internal link looks spammy. Vary your anchors: exact match (20%), partial match (30%), branded (20%), natural/descriptive (30%). Google penalizes over-optimized internal linking patterns.
