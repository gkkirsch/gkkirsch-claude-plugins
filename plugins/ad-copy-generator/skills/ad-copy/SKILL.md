---
name: ad-copy
description: >
  Premium ad copy and landing page generator powered by proven direct response frameworks.
  Generates high-converting ad copy for Facebook, Google, Instagram, X, email, YouTube, and TikTok.
  Builds complete sales page structures with hero, problem, solution, features, social proof, FAQ, and CTA sections.
  Uses AIDA, PAS, BAB, 4Ps, Triple Hook, and Star-Story-Solution frameworks.
  Triggers: "generate ad copy", "write ad", "ad copy", "sales page", "landing page copy", "marketing copy",
  "write a Facebook ad", "Google ad copy", "email copy", "headline variations", "copywriting".
  NOT for: blog posts, SEO articles, social media scheduling, PR writing, or technical documentation.
version: 1.0.0
argument-hint: "<product-description> [--platform facebook|google|instagram|x|email] [--type ad|landing-page]"
allowed-tools: Read, Grep, Glob, Bash, Write, Edit
model: sonnet

metadata:
  superbot:
    emoji: "✍️"
---

# Ad Copy Generator

Premium ad copy and landing page generator. Like hiring a $500/hr direct response copywriter — powered by the same frameworks that have generated millions in revenue across every major ad platform.

## What This Skill Does

Takes your product description, target audience, and platform — and generates:
- **Multiple headline variations** using proven formulas (curiosity gap, direct promise, contrarian, social proof)
- **Body copy in 3 lengths** (short for social, medium for Facebook/email, long for sales pages)
- **Platform-specific formatting** with character limits respected
- **CTA options** matched to funnel position
- **A/B split test suggestions** with hypotheses
- **Complete landing page structures** with section-by-section copy

## Agents

### Copywriter (`copywriter`)

Expert direct response copywriter. Specializes in ad copy across all platforms.

**What it produces**:
- 5+ headline variations using different proven formulas
- Short body copy (50-100 words) for Instagram, X, Google Display
- Medium body copy (150-300 words) for Facebook, LinkedIn, email
- Long body copy (400-800 words) for long-form Facebook, email sequences
- 3-5 CTA variations matched to funnel stage
- Platform-formatted output ready to paste
- 3 A/B test recommendations with hypotheses

**Dispatch**:
```
Task tool:
  subagent_type: "copywriter"
  description: "Generate Facebook ad copy for [product]"
  prompt: |
    Product: [description]
    Audience: [who they are and what they struggle with]
    Platform: Facebook
    Goal: Direct sale / Lead gen / Webinar registration
  mode: "bypassPermissions"
```

**Example prompts**:
- "Generate Facebook ad copy for a $97 online course teaching busy parents how to meal prep in 2 hours/week"
- "Write Google Search ads for a SaaS invoicing tool targeting freelance developers. $29/month. Competing with FreshBooks."
- "Create Instagram ad variations for a $27 ebook on ADHD productivity systems for college students"
- "Write an email sequence (3 emails) promoting a $497 coaching program for first-time managers"

### Landing Page Builder (`landing-page-builder`)

Creates complete sales page copy structures ready for implementation.

**What it produces**:
- Hero section (headline, subheadline, CTA, social proof)
- Problem section (pain points, symptoms, consequences)
- Agitate section (failed solutions, root cause, emotional escalation)
- Solution section (product intro, origin story, unique mechanism, 3-step process)
- Features & benefits (value stack with pricing)
- Social proof (testimonial placement, metrics, case studies)
- "Who this is for" qualifier section
- FAQ (8-10 objection-handling Q&As)
- Guarantee section
- Final CTA with urgency
- P.S. section
- Implementation notes (visual direction, A/B tests, optimization tips)

**Dispatch**:
```
Task tool:
  subagent_type: "landing-page-builder"
  description: "Build sales page for [product]"
  prompt: |
    Product: [name and description]
    Price: [amount]
    Audience: [who and what they want]
    Existing assets: [testimonials, case studies, features]
  mode: "bypassPermissions"
```

**Example prompts**:
- "Build a full sales page for 'The Extraction Blueprint' — a $297 course teaching people to create and sell digital products in 90 days"
- "Create an opt-in page for a free PDF guide '7 Facebook Ad Mistakes Costing You $1000/Month'"
- "Write a webinar registration page for a 60-minute masterclass on email marketing for e-commerce"

## Slash Command

Use `/generate-copy` for quick access:

```
/generate-copy My SaaS helps restaurants manage online orders. $99/month. Target: independent restaurant owners frustrated with UberEats fees.
```

```
/generate-copy --type landing-page --platform email Online fitness program for men over 40. $197. 12-week program. Joint-friendly strength training.
```

## Frameworks Included

| Framework | Structure | Best For |
|-----------|-----------|----------|
| **AIDA** | Attention → Interest → Desire → Action | Long-form ads, email sequences, Google ads |
| **PAS** | Problem → Agitate → Solution | Short social ads, any copy under 200 words |
| **BAB** | Before → After → Bridge | Email, testimonial ads, transformation content |
| **4Ps** | Promise → Picture → Proof → Push | Product launches, retargeting, high-ticket offers |
| **Triple Hook** | Premise → Stakes → Twist | Video openings, scroll-stopping first lines |
| **Star-Story-Solution** | Character → Journey → Product | Case studies, video scripts, email stories |

## Reference Library

Deep framework knowledge is available in the `references/` directory:
- **ad-frameworks.md** — Complete breakdown of all 6 frameworks with examples, psychology, and platform-specific guidance
- **sales-page-templates.md** — Full sales page blueprints for low-ticket, mid-ticket, and high-ticket products
- **hook-formulas.md** — 100+ hook templates, the Content Grid method, power words, and platform-specific rules

## Tips for Best Results

1. **Be specific about your audience**: Demographics, psychographics, pain points, and desires. More detail = better copy.
2. **Include real numbers**: Price, customer count, results achieved. Specificity is credibility.
3. **Share existing assets**: Testimonials, case studies, competitor info. Real data beats invented examples.
4. **State your goal**: Different goals need different copy strategies. Direct sale ≠ lead gen ≠ webinar signup.
5. **Mention the platform**: Each platform has different constraints and conventions. Platform-specific copy always outperforms generic copy.
