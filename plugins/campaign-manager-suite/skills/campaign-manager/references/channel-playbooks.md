# Channel Playbooks

Detailed operational playbooks for 10 major marketing channels. Each playbook includes setup guidance, targeting strategies, budget recommendations, key metrics, benchmarks, and explicit guidance on when to use — and when to avoid — each channel.

---

## 1. Email Marketing

### Overview
Email remains the highest-ROI marketing channel with an average return of $36-42 for every $1 spent. It is the only channel where you own the audience and control the distribution.

### Setup Essentials
- **ESP selection**: Klaviyo (e-commerce), HubSpot (B2B), Mailchimp (SMB), ConvertKit (creators), Brevo (budget)
- **Authentication**: Configure SPF, DKIM, and DMARC records — non-negotiable for deliverability
- **Warm-up**: New domains/IPs must be warmed gradually — start with 50-100 sends/day, increase by 20-30% daily for 2-4 weeks
- **List hygiene**: Remove bounces immediately, remove non-openers after 90 days, re-engagement campaign before removal

### Segmentation Strategy
Segment by behavior, not just demographics:

| Segment | Criteria | Strategy |
|---------|----------|----------|
| Engaged Subscribers | Opened or clicked in last 30 days | Full campaign sends, product updates |
| Warm Subscribers | Opened in last 31-90 days | Weekly digest, top content only |
| Cold Subscribers | No opens in 90+ days | Re-engagement series, then suppress |
| Recent Purchasers | Bought in last 30 days | Cross-sell, review requests, loyalty |
| Cart Abandoners | Added to cart, no purchase in 1-24 hours | 3-email recovery sequence |
| Browse Abandoners | Viewed product 2+ times, no cart add | Product reminder with social proof |
| VIP Customers | Top 10% by LTV | Early access, exclusive offers |

### Automation Sequences
**Must-have sequences (in priority order):**
1. **Welcome series** (3-5 emails over 7-14 days) — Expected open rate: 50-60%, sets the relationship
2. **Cart abandonment** (3 emails: 1hr, 24hr, 72hr) — Expected recovery rate: 5-15% of abandoned carts
3. **Post-purchase** (3-4 emails: confirmation, shipping, review request, cross-sell) — Drives repeat purchases
4. **Browse abandonment** (1-2 emails: 4hr, 24hr after browse) — 1-3% conversion rate
5. **Win-back** (3 emails over 30 days for lapsed customers) — Re-engages 3-8% of lapsed customers
6. **Lead nurture** (5-7 emails over 4-6 weeks for B2B) — Moves MQLs to SQLs

### Key Metrics and Benchmarks

| Metric | Good | Great | Excellent |
|--------|------|-------|-----------|
| Open Rate | 20% | 28% | 35%+ |
| Click Rate | 2% | 3.5% | 5%+ |
| Click-to-Open Rate | 10% | 15% | 20%+ |
| Unsubscribe Rate | <0.3% | <0.15% | <0.05% |
| Bounce Rate | <2% | <1% | <0.5% |
| Revenue per Email | $0.05 | $0.15 | $0.30+ |

### Budget: $100-500/month for tools + time to create content
### When NOT to Use: If your list is <500 subscribers, focus on list building first. If you have no automation, start there before campaigns.

---

## 2. Facebook/Instagram Ads (Meta Ads)

### Overview
Meta's ad platform reaches 3.7B monthly active users across Facebook, Instagram, Messenger, and Audience Network. Strongest for B2C, DTC, and e-commerce. Increasingly effective for B2B with improved targeting.

### Campaign Structure (Recommended)

```
Campaign (Objective: Conversions)
├── Ad Set 1: Broad Targeting (ages 25-54, interest-based)
│   ├── Ad 1: Video (product demo)
│   ├── Ad 2: Carousel (benefits)
│   └── Ad 3: Static image (testimonial)
├── Ad Set 2: Lookalike 1% (based on purchasers)
│   ├── Ad 1: Video (product demo)
│   ├── Ad 2: UGC-style video
│   └── Ad 3: Static image (offer)
├── Ad Set 3: Remarketing (website visitors 7-30 days)
│   ├── Ad 1: Testimonial
│   ├── Ad 2: Limited time offer
│   └── Ad 3: Dynamic product ads
└── Ad Set 4: Remarketing (cart abandoners 1-7 days)
    ├── Ad 1: Cart reminder
    └── Ad 2: Discount incentive
```

### Audience Targeting Hierarchy
1. **Custom Audiences** (remarketing, email lists) — Highest intent, highest ROAS
2. **Lookalike Audiences** (1% of purchasers, 1% of email list) — Best for prospecting
3. **Interest Targeting** (specific interests + behaviors) — Good for testing new segments
4. **Broad Targeting** (Advantage+ audiences) — Let Meta's algorithm find users, works well with strong creative and sufficient conversion data (50+ conversions/week)

### Creative Best Practices
- **Video first**: Video ads get 2-3x the engagement of static images on Meta
- **First 3 seconds**: Hook the viewer immediately — 65% of users who watch 3 seconds will watch 10+
- **Aspect ratios**: 9:16 for Reels/Stories, 1:1 for Feed, 4:5 for Feed (most real estate)
- **Text overlay**: Keep under 20% of the image — Meta no longer penalizes but performance still suffers
- **UGC style**: User-generated content style ads outperform polished creative by 20-50% for DTC brands
- **Refresh frequency**: Swap creative every 2-3 weeks or when frequency exceeds 3

### Key Metrics and Benchmarks

| Metric | Average | Good | Excellent |
|--------|---------|------|-----------|
| CTR (Link) | 0.9% | 1.5% | 2.5%+ |
| CPC (Link) | $1.20 | $0.80 | $0.40 |
| CPM | $12 | $8 | $5 |
| CVR (Landing Page) | 1.5% | 3% | 5%+ |
| ROAS (E-commerce) | 2.5x | 4x | 6x+ |
| ROAS (Lead Gen) | N/A | N/A | N/A |
| Frequency (Prospecting) | <3 | <2 | <1.5 |

### Budget: Minimum $1,500/month. Sweet spot: $5,000-20,000/month.
### When NOT to Use: Niche B2B with very specific ICP (<100K target users), highly regulated industries (pharma, firearms), or if you have no pixel data and <$1,500/month budget.

---

## 3. Google Search Ads

### Overview
Google Search captures high-intent users actively searching for solutions. Highest intent of any paid channel. Average CVR 3-6% vs. 1-3% for display/social.

### Campaign Structure

```
Campaign: [Product Category] - [Match Type]
├── Ad Group: [Specific Theme]
│   ├── Keywords: 5-15 tightly themed keywords
│   ├── Ad 1: Responsive Search Ad (15 headlines, 4 descriptions)
│   └── Ad 2: Responsive Search Ad (different angle)
├── Ad Group: [Another Theme]
│   └── ...
└── Negative Keyword List: Shared across campaigns
```

### Keyword Strategy
- **Brand keywords**: Always bid on your own brand (competitors will). Expected CPC: $0.50-2.00, CVR: 10-30%
- **High-intent non-brand**: "[product category] software", "best [solution] for [use case]". CPC: $2-15, CVR: 3-8%
- **Competitor keywords**: Bid on competitor brand names. CPC: $3-20, CVR: 1-3%. Lower CVR but captures switchers.
- **Long-tail keywords**: 4+ word specific queries. Lower volume, lower CPC, higher CVR.

### Quality Score Optimization
Quality Score (1-10) directly impacts CPC and ad position. Factors:
- **Expected CTR** (most important): Write compelling ads, use keywords in headlines
- **Ad relevance**: Match ad copy to keyword intent — each ad group should be tightly themed
- **Landing page experience**: Fast load time (<3s), relevant content, mobile-friendly, clear CTA

**Impact of Quality Score on CPC:**
| Quality Score | CPC Multiplier |
|--------------|----------------|
| 10 | 0.5x (50% discount) |
| 8 | 0.7x |
| 7 | 1.0x (baseline) |
| 5 | 1.5x |
| 3 | 3.0x |
| 1 | 5.0x+ |

### Bidding Strategies
- **Manual CPC**: Full control, best for small accounts (<$5K/month) or when learning
- **Target CPA**: Set your target cost-per-acquisition — needs 30+ conversions/month minimum
- **Target ROAS**: Set your target return — needs 50+ conversions/month and revenue tracking
- **Maximize Conversions**: Let Google optimize for volume — good when scaling with flexible CPA targets

### Key Metrics and Benchmarks

| Metric | B2B SaaS | E-commerce | Local Services |
|--------|----------|------------|----------------|
| CTR | 3-5% | 2-4% | 4-7% |
| CPC | $3-8 | $1-3 | $2-10 |
| CVR | 3-5% | 2-4% | 5-10% |
| CPA | $50-200 | $15-50 | $20-80 |

### Budget: Minimum $1,000/month. Sweet spot: $3,000-15,000/month.
### When NOT to Use: If nobody is searching for your product category (new market creation), or if CPCs in your industry exceed your unit economics ($50+ CPC with <$200 LTV).

---

## 4. Google Display Network

### Overview
GDN reaches 90% of internet users across 2M+ websites. Primarily used for remarketing and awareness. Lower intent than Search, but much lower CPM.

### When to Use
- **Remarketing**: Show ads to website visitors who did not convert. This is the #1 use case.
- **Brand awareness**: Reach large audiences at low CPM ($2-8)
- **Complementary to search**: Reinforce brand for users who searched but did not click/convert

### Remarketing Setup
- **Pixel all pages**: Google Ads remarketing tag on every page
- **Audience tiers**: Create audiences by recency and depth of engagement

| Audience | Window | Bid Adjustment | Expected CTR |
|----------|--------|---------------|-------------|
| Cart abandoners (1-3 days) | 3 days | +50% | 0.5-1.5% |
| Cart abandoners (4-14 days) | 14 days | +20% | 0.3-0.8% |
| Product viewers (1-7 days) | 7 days | +30% | 0.3-0.7% |
| All site visitors (1-30 days) | 30 days | Baseline | 0.1-0.4% |
| All site visitors (31-90 days) | 90 days | -20% | 0.05-0.2% |

### Creative Specs
- **Responsive Display Ads**: Upload 5+ images, 5+ headlines, 5 descriptions — Google assembles combinations
- **Key sizes** (if making static): 300x250, 728x90, 160x600, 320x50 (mobile)
- **File size**: <150KB for fast loading
- **CTA**: Always include a clear call-to-action button

### Budget: Minimum $2,000/month (mostly remarketing). Prospecting display needs $5,000+/month.
### When NOT to Use: As your only paid channel — display is a support channel, not a primary acquisition channel. Do not use for direct response without remarketing audiences.

---

## 5. LinkedIn Ads

### Overview
LinkedIn is the B2B advertising platform. 900M+ professional members with unmatched professional targeting. High CPCs ($5-15) but highly qualified audiences for B2B.

### Campaign Types (by objective)
1. **Sponsored Content** (native feed ads): Best for awareness and lead gen. CPM: $30-80. CTR: 0.4-0.7%.
2. **Message Ads (InMail)**: Direct inbox messages. Open rate: 30-50%. Cost per send: $0.50-1.00. Use sparingly — aggressive messaging causes fatigue.
3. **Lead Gen Forms**: Pre-filled forms within LinkedIn — 2-5x higher CVR than landing page forms because of auto-fill. CPL: $30-100.
4. **Document Ads**: Share PDFs/slides natively. Good engagement for thought leadership.
5. **Conversation Ads**: Choose-your-own-adventure style messages. Good for event registration and demo booking.

### Targeting Best Practices
- **Job title targeting**: Most precise but smallest audience. Minimum audience: 50,000.
- **Job function + seniority**: Broader reach with good relevance. Best for most campaigns.
- **Company targeting**: Upload account list (ABM) or target by company size/industry.
- **Skills + groups**: Interest-based targeting, broader but less precise.
- **Exclusions**: Always exclude competitors, job seekers (if not relevant), and current customers.

### Budget: Minimum $3,000/month ($100/day). Sweet spot: $5,000-20,000/month.
### When NOT to Use: B2C products, low-ticket offers (<$500 LTV), or if your ICP is not active on LinkedIn (trades, retail workers, etc.).

---

## 6. Twitter/X Ads and Organic

### Overview
Twitter/X is best for real-time engagement, thought leadership, and reaching tech-savvy and news-focused audiences. Ad platform has improved but remains less sophisticated than Meta or Google.

### Organic Strategy
- **Posting frequency**: 3-5 tweets/day for brands, 1-3 for executives
- **Thread strategy**: Long-form threads (8-15 tweets) drive 2-5x the engagement of single tweets
- **Best content types**: Hot takes on industry news, data visualizations, step-by-step tutorials, contrarian opinions
- **Engagement tactics**: Reply to industry conversations within first 30 minutes, quote-tweet with commentary
- **Timing**: B2B best: Tue-Thu 9-11am EST. B2C best: Mon-Fri 12-3pm EST.

### Paid Strategy
- **Follower campaigns**: $2-5 per follower. Use for building an audience base.
- **Website click campaigns**: $0.50-3.00 per click. Lower quality than Google/Meta but cheaper.
- **Engagement campaigns**: $0.10-0.50 per engagement. Good for amplifying high-performing organic content.

### Budget: Minimum $1,000/month for paid. Organic is time-only.
### When NOT to Use: If your audience is not on Twitter/X (check analytics first). If you cannot commit to daily engagement. If your product requires visual storytelling (use Instagram/TikTok instead).

---

## 7. Content Marketing and SEO

### Overview
Content marketing and SEO build compounding organic traffic over time. Unlike paid channels, the asset continues to generate traffic after the initial investment. Average time to rank: 3-6 months for new content.

### Content Strategy Framework

**Topic cluster model:**
```
Pillar Page (2,000-5,000 words, broad topic)
├── Cluster Article 1 (1,000-2,000 words, specific subtopic)
├── Cluster Article 2 (1,000-2,000 words, specific subtopic)
├── Cluster Article 3 (1,000-2,000 words, specific subtopic)
└── Cluster Article 4 (1,000-2,000 words, specific subtopic)
All cluster articles link to pillar page and to each other.
```

### SEO Essentials
- **Keyword research**: Target keywords with 100-10,000 monthly searches and keyword difficulty <40 (for new sites)
- **On-page**: Title tag (<60 chars, keyword first), meta description (<155 chars), H1 with keyword, 2-5 H2s, internal links (3-5 per article)
- **Content quality signals**: Original research/data, expert quotes, comprehensive coverage, better than page 1 competitors
- **Technical**: Core Web Vitals passing, mobile-friendly, SSL, clean URL structure, XML sitemap, schema markup
- **Backlinks**: Guest posts, HARO/Connectively responses, original research that gets cited, resource page link building

### Content Distribution
Creating content is 30% of the effort. Distribution is 70%:
- Share to all social channels (day 1)
- Email to newsletter subscribers (day 1-3)
- Repurpose into social media posts (week 1-2)
- Submit to relevant communities (Reddit, HN, indie hackers) if genuinely useful
- Pitch to newsletter curators in your niche
- Update and re-promote top performers every 6 months

### Key Metrics

| Metric | 6-Month Target | 12-Month Target |
|--------|---------------|-----------------|
| Organic sessions/month | 5,000 | 20,000 |
| Keyword rankings (page 1) | 20 | 100 |
| Backlinks acquired | 50 | 200 |
| Content-attributed leads | 50/month | 200/month |
| Domain Authority | +5 points | +10 points |

### Budget: $1,000-5,000/month (writers + tools). Compound returns: content published today generates traffic for years.
### When NOT to Use: If you need results in <3 months. If you do not have capacity to produce quality content consistently. If your audience does not search for information (commodity products).

---

## 8. Influencer Marketing

### Overview
Partner with individuals who have built trust with your target audience. Effective for brand awareness, credibility, and reaching audiences resistant to traditional advertising.

### Influencer Tiers

| Tier | Followers | Typical Cost/Post | Engagement Rate | Best For |
|------|----------|-------------------|-----------------|----------|
| Nano | 1K-10K | $50-250 | 5-10% | Hyper-niche, authentic content |
| Micro | 10K-100K | $250-2,500 | 3-7% | Targeted reach, high engagement |
| Mid-tier | 100K-500K | $2,500-10,000 | 2-5% | Balanced reach and engagement |
| Macro | 500K-1M | $10,000-50,000 | 1-3% | Broad awareness |
| Mega | 1M+ | $50,000+ | 0.5-2% | Mass awareness, celebrity effect |

### ROI Tracking
- **Unique discount codes**: Assign each influencer a unique code to track conversions
- **UTM links**: Custom UTM link per influencer for GA4 tracking
- **Affiliate commissions**: 10-20% of sales is standard for performance-based deals
- **Expected ROI**: $5-6.50 earned media value per $1 spent (industry average)

### Finding and Vetting Influencers
1. Search hashtags in your niche on target platforms
2. Check competitor follower lists for mutual follows
3. Use tools: AspireIQ, Upfluence, CreatorIQ, or even manual Instagram search
4. Vet for: Engagement rate (should match tier expectations), audience demographics (check with the creator), content quality, brand alignment, previous brand partnerships

### Budget: Minimum $2,000/month. Start with 5-10 micro-influencers rather than 1 macro.
### When NOT to Use: Highly technical B2B products (influencers have limited reach in niche enterprise). If you cannot track ROI (no unique codes/links). If your product requires hands-on demos that cannot be authentically shown on social.

---

## 9. YouTube Ads

### Overview
YouTube is the second-largest search engine and the world's largest video platform. 2.7B monthly active users. Strong for mid-funnel consideration and brand building. Video completion rates are higher than other platforms.

### Ad Formats

| Format | Length | Payment | Best For | Expected VTR |
|--------|--------|---------|----------|-------------|
| Skippable In-Stream | 15s-3min | CPV (after 30s or interaction) | Brand awareness, consideration | 20-35% |
| Non-Skippable | 15s max | CPM | Guaranteed views, brand awareness | 90%+ (forced) |
| Bumper | 6s max | CPM | Reach, frequency, brand recall | 90%+ (forced) |
| In-Feed (Discovery) | Any | CPC (on click/view) | Tutorials, reviews, longer content | N/A |
| Shorts Ads | <60s | CPM/CPV | Young demographics, mobile-first | 15-25% |

### Creative Tips
- **Hook in first 5 seconds**: Ask a provocative question, show a surprising result, or directly address the viewer's problem
- **Branding by second 5**: Do not wait until the end — most viewers skip after 5 seconds
- **Clear CTA**: Tell viewers exactly what to do next — use both verbal and on-screen CTA
- **Optimal length**: 15-30 seconds for awareness, 60-120 seconds for consideration/education

### Targeting
- **Custom Intent** (people who searched on Google for related terms): Highest intent YouTube targeting
- **In-Market** (people actively researching your category): Good mid-funnel targeting
- **Remarketing** (people who visited your site): Show deeper content to warm audiences
- **Similar Audiences**: Based on remarketing lists, expands reach with similar profiles

### Budget: Minimum $2,500/month. CPV typically $0.03-0.15. Video production costs $500-5,000+ per video.
### When NOT to Use: If you have no video production capability or budget. If your product is hard to demonstrate visually. If your audience is primarily on LinkedIn or niche professional platforms.

---

## 10. TikTok Ads and Organic

### Overview
TikTok has 1.5B+ monthly active users and has expanded well beyond Gen Z. Average daily time on app: 95 minutes. Organic reach is still higher than any other major platform.

### Organic Growth Strategy
- **Posting frequency**: 1-3 times per day (algorithm rewards consistency)
- **First 1 second**: Hook immediately — pattern interrupt, text on screen, direct address
- **Trending sounds**: Use trending audio to boost discoverability (check TikTok Creative Center)
- **Content types that work**: Behind-the-scenes, tutorials, before/after, day-in-the-life, myth-busting, relatable humor
- **Hashtag strategy**: 3-5 hashtags — 1 broad, 2 niche, 1 trending
- **Cross-posting**: TikTok content works on Instagram Reels and YouTube Shorts (remove watermark first)

### Paid Advertising
- **Spark Ads**: Boost organic content (yours or creator's) — best performing ad format on TikTok, 30-50% higher CVR than standard ads because they look native
- **In-Feed Ads**: Standard feed placement. Make it look like organic content — overly polished ads get skipped.
- **TopView**: First ad users see when opening the app. Premium placement, premium cost ($50K+ per day).
- **Branded Hashtag Challenge**: $100K+ entry point. Good for major brands only.

### Creator Partnerships
- TikTok's Creator Marketplace connects brands with creators
- Brief creators on the message but let them create in their own style — audience authenticity is critical
- Budget: $200-2,000 per creator for micro/mid-tier
- Always request usage rights for Spark Ads (boost their content as paid)

### Key Metrics

| Metric | Average | Good | Excellent |
|--------|---------|------|-----------|
| CTR (In-Feed) | 0.5% | 1% | 2%+ |
| CVR | 0.5% | 1.5% | 3%+ |
| CPC | $0.80 | $0.50 | $0.25 |
| CPM | $8 | $5 | $3 |
| Video Completion Rate | 15% | 25% | 40%+ |
| ROAS (E-commerce) | 2x | 4x | 6x+ |

### Budget: Minimum $1,000/month for ads. Organic requires no budget but 5-10 hours/week of content creation.
### When NOT to Use: B2B enterprise sales (audience is consumer-focused). If your brand cannot adopt a casual, authentic tone. If you cannot create short-form vertical video content consistently.

---

## Channel Selection Quick Reference

### By Goal

| Goal | Primary Channels | Supporting Channels |
|------|-----------------|-------------------|
| Immediate sales | Google Search, Meta Ads, Email | Remarketing (Display, YouTube) |
| Lead generation | Google Search, LinkedIn, Content/SEO | Email nurture, Meta remarketing |
| Brand awareness | Meta, YouTube, TikTok, PR | Display, Influencer, Content |
| App installs | Meta, Google UAC, TikTok, Apple Search | Influencer, Content |
| E-commerce growth | Google Shopping, Meta, Email | TikTok, Influencer, SEO |
| B2B pipeline | LinkedIn, Google Search, Content/SEO | Email, Webinars, PR |
| Local business | Google Search (local), Meta (geo-targeted) | Google Maps, Yelp, Email |

### By Budget

| Monthly Budget | Recommended Channels |
|---------------|---------------------|
| <$1K | Email + 1 organic channel (SEO or Social) |
| $1K-3K | Email + Google Search OR Meta Ads |
| $3K-10K | Email + Google Search + Meta Ads |
| $10K-25K | Email + Google Search + Meta + 1-2 additional |
| $25K-50K | Full funnel: Search + Social + Display + Email + Content |
| $50K+ | Omnichannel with testing budget for emerging channels |
