---
name: campaign-manager
description: "Multi-channel marketing campaign management — plan, execute, and optimize campaigns across email, social, paid ads, content, and SEO. Includes A/B testing frameworks, budget allocation, performance tracking, attribution modeling, and automated reporting."
triggers:
  - "plan campaign"
  - "marketing campaign"
  - "A/B test"
  - "campaign strategy"
  - "multi-channel marketing"
  - "campaign performance"
  - "ad campaign"
  - "growth strategy"
version: "1.0.0"
argument-hint: "<campaign goal, product, budget, timeline, or analysis request>"
allowed-tools:
  - Read
  - Write
  - Edit
  - Bash
  - Glob
  - Grep
  - WebSearch
  - WebFetch
model: sonnet
---

# Campaign Manager Skill

A comprehensive marketing campaign management skill covering planning, execution, optimization, and analysis across all major marketing channels. This skill provides data-driven frameworks, real benchmarks, and actionable templates — not generic marketing advice.

---

## 1. Campaign Planning Framework

Every campaign follows this seven-stage framework. Do not skip stages.

```
OBJECTIVE → AUDIENCE → CHANNELS → CONTENT → BUDGET → TIMELINE → MEASURE
```

### Stage 1: Objective Setting (SMART Goals)

Every campaign must start with a SMART objective. Translate vague goals into measurable targets:

| Vague Goal | SMART Objective |
|------------|----------------|
| "Increase awareness" | "Reach 500,000 unique users in target demographic within 8 weeks, measured by unique impressions in GA4" |
| "Get more leads" | "Generate 200 marketing-qualified leads at <$75 CPL within 6 weeks via gated content + paid promotion" |
| "Drive sales" | "Generate $50,000 in attributed revenue at 4x ROAS within Q2 via multi-channel paid campaign" |
| "Grow social following" | "Add 5,000 qualified followers on LinkedIn within 12 weeks with >3% engagement rate" |

**Objective hierarchy:**
1. **Primary objective** — The single most important outcome (only one)
2. **Secondary objectives** — 1-2 supporting outcomes
3. **Guardrail metrics** — Constraints (e.g., "maintain CAC below $100", "keep brand sentiment above 80%")

### Stage 2: Audience Definition

Build a complete audience profile using this template:

**Demographics**: Age range, gender, location, income level, education
**Firmographics** (B2B): Company size, industry, revenue, tech stack, growth stage
**Psychographics**: Values, interests, pain points, aspirations, media consumption
**Behavioral**: Purchase history, brand interactions, content consumption, device usage
**Awareness Stage**: Unaware → Problem-aware → Solution-aware → Product-aware → Most aware

Map your audience to channels:
- **Decision-makers (B2B)**: LinkedIn, Google Search, industry publications, email
- **Young consumers (18-24)**: TikTok, Instagram, YouTube, Snapchat
- **General consumers (25-54)**: Facebook, Instagram, Google, YouTube, email
- **Tech-savvy professionals**: Twitter/X, Reddit, Hacker News, podcasts, newsletters
- **High-income consumers**: Instagram, YouTube, podcasts, premium publications

### Stage 3: Channel Selection

Use the Channel Selection Matrix to choose the right channels:

| Channel | Best For | Min Budget/mo | Time to Results | Complexity |
|---------|----------|--------------|-----------------|------------|
| Google Search | High-intent capture | $1,000 | 2-4 weeks | Medium |
| Google Display | Remarketing, awareness | $2,000 | 4-8 weeks | Medium |
| Meta (FB/IG) | B2C acquisition, remarketing | $1,500 | 2-6 weeks | High |
| LinkedIn Ads | B2B lead gen | $3,000 | 4-8 weeks | Medium |
| TikTok Ads | Young demographic reach | $1,000 | 2-4 weeks | Medium |
| YouTube Ads | Brand + consideration | $2,500 | 4-8 weeks | High |
| Email Marketing | Nurture, retention, upsell | $200 | 1-2 weeks | Low |
| Content/SEO | Long-term organic traffic | $1,000 | 3-6 months | Medium |
| Influencer | Trust + reach in niche | $2,000 | 2-4 weeks | High |
| PR | Credibility, awareness | $3,000 | 4-12 weeks | High |

**Channel selection rules:**
- Budget under $3K/month: Pick 1-2 channels maximum
- Budget $3K-10K/month: Pick 2-4 channels
- Budget $10K-50K/month: Pick 4-6 channels
- Budget $50K+/month: Full omnichannel possible

### Stage 4: Content Planning

Map content to funnel stages:

| Funnel Stage | Content Type | Channels | CTA |
|-------------|-------------|----------|-----|
| Awareness | Blog posts, videos, infographics, social posts | SEO, social, display, YouTube | Learn more |
| Consideration | Case studies, webinars, comparison guides, demos | Search, email, remarketing, LinkedIn | Download / Register |
| Decision | Free trials, consultations, ROI calculators, testimonials | Search, email, remarketing | Start free trial / Book demo |
| Retention | Onboarding emails, tutorials, community, newsletters | Email, in-app, community | Upgrade / Refer |

### Stage 5: Budget Allocation

**The 70/20/10 Framework:**
- **70% Proven** — Channels with demonstrated positive ROI in past campaigns
- **20% Promising** — Channels showing potential that need further investment to scale
- **10% Experimental** — New or untested channels for learning and future pipeline

**Budget allocation by campaign type:**

| Campaign Type | Paid Media | Content | Creative | Tools/Tech | Reserve |
|--------------|-----------|---------|----------|-----------|---------|
| Product Launch | 50% | 20% | 15% | 5% | 10% |
| Lead Generation | 60% | 15% | 10% | 5% | 10% |
| Brand Awareness | 55% | 25% | 10% | 5% | 5% |
| E-commerce Sales | 65% | 10% | 10% | 5% | 10% |
| Retention/Upsell | 20% | 30% | 10% | 10% | 30% |

**Budget pacing:**
- Week 1-2: Spend at 60-70% of daily budget (learning phase)
- Week 3-4: Ramp to 100% on winning segments
- Week 5+: Increase spend on top performers, cut underperformers
- Always hold 10% in reserve for optimization and contingencies

### Stage 6: Timeline and Campaign Calendar

**Minimum timelines by channel:**

| Channel | Setup Time | Learning Phase | Optimization Phase | Minimum Campaign |
|---------|-----------|---------------|-------------------|-----------------|
| Google Search | 3-5 days | 2 weeks | Ongoing | 4 weeks |
| Meta Ads | 2-3 days | 1-2 weeks | Ongoing | 4 weeks |
| LinkedIn Ads | 3-5 days | 2-3 weeks | Ongoing | 6 weeks |
| Email Sequence | 1-2 weeks | 1 week | Ongoing | 3 weeks |
| Content/SEO | 2-4 weeks | N/A | 3-6 months | 6 months |
| Influencer | 2-4 weeks | N/A | N/A | 4 weeks |

**Pre-launch checklist (2 weeks before launch):**
1. All tracking pixels and conversion events installed and tested
2. UTM parameter conventions documented and applied
3. Landing pages live, tested on mobile, load time <3 seconds
4. Ad creative approved and uploaded to platforms
5. Email sequences built, tested, and scheduled
6. Analytics dashboards configured with baseline metrics
7. Team roles and responsibilities assigned
8. Reporting cadence and templates confirmed
9. Budget loaded into ad platforms
10. Contingency plans documented

### Stage 7: Measurement Framework

**Key metrics by category:**

**Acquisition Metrics:**
- CAC (Customer Acquisition Cost) = Total marketing spend / New customers acquired
- CPL (Cost Per Lead) = Campaign spend / Leads generated
- CPA (Cost Per Action) = Campaign spend / Desired actions
- ROAS (Return on Ad Spend) = Revenue from ads / Ad spend

**Engagement Metrics:**
- CTR (Click-Through Rate) = Clicks / Impressions
- CVR (Conversion Rate) = Conversions / Clicks (or visitors)
- Bounce Rate = Single-page sessions / Total sessions
- Time on Site / Pages per Session

**Revenue Metrics:**
- LTV (Lifetime Value) = Average Revenue Per User x Average Lifespan
- LTV:CAC Ratio (target: 3:1 or higher)
- CAC Payback Period = CAC / Monthly Revenue Per Customer
- MER (Marketing Efficiency Ratio) = Total Revenue / Total Marketing Spend

**Channel-specific benchmarks:**

| Metric | Google Search | Meta Ads | LinkedIn | Email | Content/SEO |
|--------|-------------|----------|----------|-------|-------------|
| CTR | 3-7% | 0.9-1.6% | 0.4-0.7% | 2-5% | N/A |
| CVR | 3-6% | 1.5-3.5% | 2-5% | 1-3% | 2-5% |
| CPC | $1-5 | $0.50-2.00 | $5-12 | N/A | N/A |
| CPM | $20-50 | $8-15 | $30-80 | $0.50-2 | N/A |

---

## 2. A/B Testing Framework

Refer to `references/ab-testing-guide.md` for the complete A/B testing methodology.

**Quick reference — Hypothesis format:**
```
If we [change X], then [metric Y] will [increase/decrease] by [amount] because [reason based on data/insight].
```

**Minimum sample sizes for common scenarios (95% confidence, 80% power):**

| Baseline CVR | MDE (Relative) | Sample Per Variation |
|-------------|----------------|---------------------|
| 2% | 10% | 38,415 |
| 2% | 20% | 9,823 |
| 5% | 10% | 14,751 |
| 5% | 20% | 3,783 |
| 10% | 10% | 7,009 |
| 10% | 20% | 1,810 |
| 20% | 10% | 3,242 |
| 20% | 20% | 843 |

**Test duration rules:**
- Minimum: 2 full weeks (captures weekday/weekend patterns)
- Ideal: 1-2 full business cycles
- Maximum: 4-6 weeks (external factors contaminate longer tests)
- Never call a test early based on initial results

**Test prioritization (ICE Framework):**
- **I**mpact (1-10): How much will this move the needle?
- **C**onfidence (1-10): How confident are we in the expected outcome?
- **E**ase (1-10): How easy is this to implement?
- Score = (I + C + E) / 3
- Run highest-scoring tests first

---

## 3. Attribution Models

**Model comparison:**

| Model | How It Works | Best For | Weakness |
|-------|-------------|----------|----------|
| Last-Click | 100% credit to final touchpoint | Short sales cycles, direct response | Ignores awareness + consideration |
| First-Click | 100% credit to first touchpoint | Understanding discovery channels | Ignores nurture + conversion |
| Linear | Equal credit to all touchpoints | General understanding | Oversimplifies reality |
| Time-Decay | More credit to recent touchpoints | Longer sales cycles (B2B) | Undervalues awareness |
| Position-Based | 40% first, 40% last, 20% middle | Balanced view of funnel | Arbitrary weighting |
| Data-Driven | ML-based on conversion patterns | High volume (300+ conversions/month) | Black box, needs volume |

**Recommended approach by business type:**
- **E-commerce (short cycle)**: Last-click for daily optimization, position-based for strategy
- **B2B SaaS (long cycle)**: Time-decay for daily, linear for quarterly review
- **DTC brands**: Position-based for acquisition, last-click for remarketing
- **Mobile apps**: Last-click for install campaigns, linear for LTV analysis

---

## 4. UTM Parameter Conventions

Standard UTM structure:
```
?utm_source=[platform]&utm_medium=[channel-type]&utm_campaign=[campaign-name]&utm_content=[creative-variant]&utm_term=[keyword-or-targeting]
```

**Naming conventions (use lowercase, hyphens, no spaces):**

| Parameter | Convention | Examples |
|-----------|-----------|----------|
| utm_source | Platform name | google, facebook, linkedin, newsletter, twitter |
| utm_medium | Channel type | cpc, cpm, social, email, organic, referral, display |
| utm_campaign | Campaign identifier | spring-sale-2026, product-launch-q2, lead-gen-webinar |
| utm_content | Creative variant | hero-image-a, cta-green, headline-v2 |
| utm_term | Keyword/targeting | brand-keywords, lookalike-1pct, retargeting-30d |

---

## 5. Reporting Templates

### Weekly Performance Dashboard

Report every Monday covering the previous 7 days:
1. **Top-line metrics**: Spend, conversions, revenue, ROAS, CAC — with week-over-week change
2. **Channel performance table**: All channels with key metrics and trends
3. **Top performers**: Best 3 ad sets/campaigns and why
4. **Underperformers**: Bottom 3 ad sets/campaigns and recommended action
5. **Budget pacing**: Actual spend vs. planned spend, projected month-end
6. **Action items**: 3-5 specific optimizations to implement this week

### Monthly Review

Report on the 1st of each month:
1. **Executive summary**: 3-sentence overview of the month
2. **Goal tracking**: Progress toward campaign objectives (% complete)
3. **Full metrics breakdown**: All KPIs with month-over-month and year-over-year comparison
4. **Channel deep dive**: Detailed analysis of each channel
5. **A/B test results**: Summary of all tests run, winners, and learnings
6. **Budget analysis**: Planned vs. actual, cost efficiency trends
7. **Competitive observations**: Any notable competitor activity
8. **Next month priorities**: Top 5 focus areas with rationale

### Quarterly Strategy Review

Report every quarter:
1. **Quarter performance summary**: Did we hit our goals?
2. **Channel portfolio analysis**: Which channels earned their budget?
3. **Customer acquisition analysis**: CAC trends, LTV:CAC ratio evolution
4. **Attribution analysis**: Cross-channel influence and customer journey insights
5. **Competitive landscape**: Market position changes
6. **Budget recommendation**: Next quarter budget allocation with rationale
7. **Strategic pivots**: Any major changes in approach
8. **Testing roadmap**: Priority tests for next quarter

---

## 6. Campaign Lifecycle Management

### Phase 1: Planning (Weeks -4 to -2)
- Define objectives and KPIs
- Research audience and competitors
- Select channels and allocate budget
- Brief creative and content teams
- Set up tracking and analytics

### Phase 2: Build (Weeks -2 to 0)
- Create all campaign assets
- Build landing pages and forms
- Set up ad accounts and campaigns
- Configure email sequences
- QA everything: tracking, links, mobile, load speed

### Phase 3: Launch (Week 0)
- Execute launch sequence (staggered or simultaneous)
- Monitor first 24-48 hours closely
- Verify all tracking is firing correctly
- Confirm budget pacing is correct
- Share launch confirmation with stakeholders

### Phase 4: Optimize (Weeks 1-N)
- Daily: Check spend pacing, pause failing ads
- 3x/week: Review performance by segment, adjust bids
- Weekly: Full performance review, optimization actions
- Bi-weekly: Creative refresh assessment, audience review
- Monthly: Strategic review, budget reallocation

### Phase 5: Close and Analyze (Final Week + 1)
- Pause all campaign elements
- Allow 7-14 days for conversion lag (especially B2B)
- Pull final performance data
- Conduct post-mortem analysis
- Document learnings for future campaigns
- Archive creative and copy for reference

---

## 7. Post-Mortem Analysis Framework

After every campaign, answer these questions:

1. **Results vs. Goals**: What was the target? What did we achieve? Express as percentage of goal.
2. **Budget Efficiency**: How much did we spend? What was the effective CPA/CAC? How does this compare to our target and industry benchmarks?
3. **Channel Analysis**: Rank channels by ROAS. Which exceeded expectations? Which disappointed? Why?
4. **Creative Analysis**: Which creative themes/formats performed best? What can we learn about our audience's preferences?
5. **Audience Insights**: Which audience segments converted best? Any surprising segments?
6. **Timing Insights**: Were there day-of-week or time-of-day patterns? Seasonal effects?
7. **Technical Issues**: Any tracking gaps, platform issues, or technical problems that impacted results?
8. **What Worked**: Top 3 things to repeat or scale
9. **What Failed**: Top 3 things to stop or change
10. **Recommendations**: Specific actions for the next campaign

---

## Reference Materials

This skill includes detailed reference guides:

- **`references/ab-testing-guide.md`** — Complete A/B testing methodology with formulas, sample size tables, 20 test ideas, and the ICE prioritization framework
- **`references/channel-playbooks.md`** — Detailed playbook for 10 marketing channels with setup guides, benchmarks, budget recommendations, and dos/don'ts
- **`references/campaign-templates.md`** — 5 ready-to-use campaign templates (Product Launch, Lead Gen, Brand Awareness, E-commerce Sales, SaaS Free Trial) with week-by-week timelines and budget breakdowns

When building a campaign plan or analyzing performance, reference these guides for channel-specific benchmarks, proven templates, and testing frameworks.
