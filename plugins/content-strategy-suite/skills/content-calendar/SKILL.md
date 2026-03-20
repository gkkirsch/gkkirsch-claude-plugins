---
name: content-calendar
description: >
  Content calendar planning and editorial workflow patterns.
  Use when building content calendars, planning editorial schedules,
  managing content pipelines, or organizing multi-channel publishing.
  Triggers: "content calendar", "editorial calendar", "publishing schedule",
  "content pipeline", "content plan", "editorial workflow", "content cadence".
  NOT for: SEO keyword research (see seo-keyword-research), writing individual posts, social media scheduling tools.
version: 1.0.0
allowed-tools: Read, Grep, Glob, Bash
---

# Content Calendar Planning

## Monthly Content Calendar Template

```markdown
# Content Calendar — [Month Year]

## Monthly Theme: [Overarching topic that ties content together]

## Content Mix (follow the 70/20/10 rule)
- 70% proven formats that consistently perform (how-tos, tutorials, listicles)
- 20% experimental formats (new series, collaborations, data studies)
- 10% high-risk/high-reward (hot takes, predictions, controversial opinions)

## Publishing Schedule

### Week 1: [Sub-theme]
| Day | Channel | Type | Title/Topic | Primary KW | Status | Owner |
|-----|---------|------|-------------|------------|--------|-------|
| Mon | Blog | How-to | [title] | [keyword] | Draft | [name] |
| Tue | Newsletter | Deep dive | [title] | — | Planned | [name] |
| Wed | X/Twitter | Thread | [topic] | — | Planned | [name] |
| Thu | Blog | Listicle | [title] | [keyword] | Planned | [name] |
| Fri | LinkedIn | Insight | [topic] | — | Planned | [name] |

### Week 2: [Sub-theme]
| Day | Channel | Type | Title/Topic | Primary KW | Status | Owner |
|-----|---------|------|-------------|------------|--------|-------|
| Mon | Blog | Tutorial | [title] | [keyword] | Planned | [name] |
| Tue | Newsletter | Curated | [title] | — | Planned | [name] |
| Wed | X/Twitter | Poll | [question] | — | Planned | [name] |
| Thu | Blog | Case study | [title] | [keyword] | Planned | [name] |
| Fri | LinkedIn | Story | [topic] | — | Planned | [name] |

### Week 3-4: [Continue pattern...]

## Key Dates & Hooks
- [Industry event/conference — create related content]
- [Product launch/update — announcement + tutorial]
- [Seasonal trend — capitalize on search spikes]
- [Competitor event — counter-programming content]

## Repurposing Plan
Each blog post generates:
1. X/Twitter thread (publish day of blog)
2. LinkedIn post (publish day after blog)
3. Newsletter section (include in weekly send)
4. 3 social media quotes/graphics (schedule over following week)
```

## Content Pipeline Stages

```markdown
# Editorial Workflow

## Pipeline Stages

1. **Ideation** (Backlog)
   - Source: keyword research, audience questions, competitor gaps, team ideas
   - Output: one-line topic + target keyword + estimated intent
   - Gate: relevance score > 3/5, not duplicate of existing content

2. **Brief** (Assigned)
   - Owner writes content brief (see seo-keyword-research skill)
   - SERP analysis complete
   - Outline approved by editor
   - Gate: brief reviewed, target keyword confirmed

3. **Draft** (In Progress)
   - Writer creates first draft following brief
   - Includes all required sections, images, and CTAs
   - Self-edit for clarity and flow
   - Gate: draft complete, word count within 20% of target

4. **Review** (In Review)
   - Editor reviews for: accuracy, tone, SEO, brand voice
   - Feedback given as comments, not rewrites
   - Maximum 2 review rounds (if more needed, brief was unclear)
   - Gate: editor approves

5. **Optimize** (Pre-Publish)
   - Add internal links (pillar page + 2-3 cluster pages)
   - Optimize title tag and meta description
   - Add alt text to all images
   - Add structured data (FAQ, HowTo, Article schema)
   - Set up tracking (UTM params for newsletter, social)
   - Gate: SEO checklist complete

6. **Publish** (Live)
   - Schedule publication time (Tue-Thu 8-10am typically best)
   - Create social media posts for distribution
   - Queue newsletter mention
   - Gate: live URL verified, no broken links/images

7. **Distribute** (Promoting)
   - Share on all social channels with channel-native formatting
   - Send to email list (if feature piece)
   - Share in relevant communities (Reddit, Slack, Discord)
   - Gate: all distribution channels covered within 48 hours

8. **Monitor** (Tracking)
   - Track rankings weekly for 90 days
   - Monitor traffic in analytics
   - Check GSC for impression/click data
   - Gate: 90-day performance review, refresh decision

## Status Tracking
Use consistent labels:
- 🔵 Backlog — idea captured, not yet assigned
- 🟡 In Progress — writer actively working
- 🟠 In Review — with editor
- 🟢 Ready — approved, scheduled for publish
- ✅ Published — live on site
- 🔄 Refresh — flagged for update
- ❌ Killed — idea abandoned (with reason noted)
```

## Content Cadence by Business Stage

```markdown
# Publishing Frequency Guide

## Early Stage (0-1K monthly visitors)
- Blog: 2 posts/week (focus on long-tail, low-competition keywords)
- Newsletter: 1/week (even with a small list — consistency > size)
- Social: 3-5 posts/week on primary platform
- Goal: build topical authority in one cluster before expanding

## Growth Stage (1K-10K monthly visitors)
- Blog: 3-4 posts/week (mix of SEO-driven + thought leadership)
- Newsletter: 1-2/week (segment by interest if possible)
- Social: daily on primary, 3x/week on secondary
- Goal: dominate 2-3 topic clusters, begin monetization

## Scale Stage (10K+ monthly visitors)
- Blog: 4-5 posts/week (multiple writers, editorial team)
- Newsletter: 2-3/week (segmented sends)
- Social: daily on all platforms with native content
- Goal: expand clusters, refresh underperformers, launch gated content

## Cadence Rules (All Stages)
1. **Consistency > volume** — 1 great post every Tuesday beats 5 random posts whenever
2. **Batch create** — write 4 posts in one session rather than 1 per day
3. **Buffer of 2 weeks** — always have 2 weeks of content scheduled ahead
4. **Monday brief, Friday publish** — write briefs Monday, draft Tue-Wed, review Thu, publish Fri
5. **Never skip the newsletter** — it's your owned audience. Social reach declines; email doesn't.
```

## Content Repurposing Matrix

```markdown
# One Piece → Many Formats

## From a Blog Post (source content)

| Format | Channel | When | Effort |
|--------|---------|------|--------|
| Thread (5-8 tweets) | X/Twitter | Day of publish | Low |
| LinkedIn article | LinkedIn | Day after | Low |
| Carousel (5-7 slides) | Instagram/LinkedIn | Day 2 | Medium |
| Short video (60-90s) | TikTok/Reels/Shorts | Week after | Medium |
| Newsletter feature | Email | Next send | Low |
| Infographic | Pinterest/Blog | Week 2 | Medium |
| Podcast talking points | Podcast | When scheduled | Low |
| Webinar/workshop | Live event | Monthly | High |
| Slide deck | SlideShare/Loom | Week 2 | Medium |
| Quora/Reddit answers | Communities | Ongoing | Low |

## Repurposing Rules
1. **Lead with the hook** — every format needs its own compelling opening
2. **Platform-native** — don't paste a blog paragraph as a tweet. Rewrite for the format.
3. **Space it out** — don't publish all formats on the same day. Spread over 2 weeks.
4. **Track what works** — some formats will outperform the original blog post. Double down.
5. **Link back** — every repurposed piece should drive traffic to the canonical blog post.
```

## Quarterly Content Review

```markdown
# Quarterly Content Audit Template

## Performance Tiers (by pageviews or conversions)

### Tier 1: Winners (top 20% of content by traffic)
Action: Refresh, expand, build more content on the same topic cluster
- Update with new data, screenshots, examples
- Add sections for newly relevant subtopics
- Build internal links from new content to these pages
- Create repurposed formats (video, infographic)

### Tier 2: Potential (positions 5-20 in search, or growing traffic)
Action: Optimize to push onto page 1
- Improve title tags (test more compelling CTAs)
- Add FAQ section targeting People Also Ask
- Improve internal linking (add links from high-authority pages)
- Expand thin sections where competitors have more depth

### Tier 3: Underperformers (low traffic, declining, or not ranking)
Action: Evaluate — refresh, merge, or redirect
- If topic is still relevant → major refresh (new angle, more depth)
- If content overlaps with another page → merge into the stronger page + 301 redirect
- If topic is no longer relevant → redirect to related page or 410 remove

### Tier 4: No-index candidates
Action: Remove from search index
- Thin tag/category pages with no unique content
- Old event pages, outdated announcements
- Duplicate content created by CMS (pagination, filters)

## Quarterly Metrics to Track
| Metric | Source | Goal |
|--------|--------|------|
| Organic traffic | GSC / Analytics | +15% QoQ |
| Keyword rankings (page 1) | Rank tracker | +20 new keywords |
| Content published | Calendar | Meet cadence target |
| Newsletter subscribers | Email platform | +10% QoQ |
| Content conversion rate | Analytics | >2% blog → signup |
| Average position | GSC | Improve by 3+ positions |
```

## Gotchas

1. **Publishing without distribution** -- Writing great content is half the work. If you don't distribute it (social, email, communities), no one sees it. Allocate equal time to creation and distribution. A mediocre post with great distribution outperforms a great post with none.

2. **No content buffer** -- Publishing on a deadline-driven schedule leads to rushed, low-quality content. Maintain a 2-week buffer of scheduled content. If you fall behind, reduce cadence temporarily rather than lowering quality.

3. **Ignoring content decay** -- Blog posts lose relevance over time. Content published 12+ months ago needs a freshness audit. Check GSC for declining impressions, update stats and examples, and re-optimize for current SERP features.

4. **Too many channels, not enough depth** -- Posting on 5 platforms poorly is worse than dominating 2 platforms. Start with one social platform and email. Add channels only when your current ones are performing well and you have capacity.

5. **Not segmenting newsletter content** -- Sending the same email to your entire list ignores that your audience has different interests. Segment by topic interest, engagement level, or funnel stage. Even two segments (new vs engaged subscribers) improves open rates significantly.

6. **Measuring vanity metrics** -- Pageviews and social impressions feel good but don't pay bills. Track metrics that connect to revenue: email signups, trial starts, demo requests, and content-attributed conversions. If a post gets 10K views but zero conversions, it's not performing.
