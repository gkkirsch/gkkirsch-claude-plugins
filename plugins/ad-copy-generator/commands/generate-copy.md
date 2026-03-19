---
name: generate-copy
description: >
  Generate high-converting ad copy or landing page content. Provide a product description,
  target audience, and platform to get multiple ad variations with headlines, body copy, CTAs,
  and A/B test suggestions. Or request a full landing page and get a complete sales page structure.
  Triggers: "generate ad copy", "write ad", "create sales page", "landing page copy", "marketing copy".
  NOT for: blog posts, SEO content, social media scheduling, or email newsletters.
version: 1.0.0
argument-hint: "<product-description> [--platform facebook|google|instagram|x|email] [--type ad|landing-page]"
allowed-tools: Read, Grep, Glob, Bash, Write, Edit
model: sonnet
user-invocable: true

metadata:
  superbot:
    emoji: "✍️"
---

# Generate Copy

Generate premium ad copy or landing page content powered by proven direct response frameworks.

## What You Need

To generate the best copy, provide:

1. **Product/Service**: What you're selling (name, description, price)
2. **Target Audience**: Who you're selling to (demographics, pain points, desires)
3. **Platform** (optional): Where the ad will run (facebook, google, instagram, x, email, youtube, tiktok)
4. **Type** (optional): `ad` for ad copy variations, `landing-page` for full sales page structure

## How to Use

### Quick Ad Copy
```
/generate-copy My SaaS product helps freelancers track invoices. $29/month. Target: freelance designers and developers who hate admin work.
```

### Platform-Specific
```
/generate-copy --platform google --type ad Project management tool for remote teams. $15/user/month. Competing with Monday.com and Asana.
```

### Full Landing Page
```
/generate-copy --type landing-page Online course teaching parents how to help kids with ADHD focus. $197. Target: parents of kids ages 8-14 with ADHD.
```

## What You Get

### For Ad Copy (`--type ad`, default)
- **Audience analysis**: Core desire, pain, sophistication, awareness, objections
- **5+ headline variations**: Using proven formulas (Direct Promise, Curiosity Gap, PAS, etc.)
- **3 body copy versions**: Short (50-100 words), Medium (150-300 words), Long (400-800 words)
- **CTA options**: Direct, benefit-driven, low-friction, urgency, curiosity
- **Platform formatting**: Copy pre-formatted for your target platform
- **A/B test suggestions**: 3 specific split tests with hypotheses

### For Landing Pages (`--type landing-page`)
- **Full page structure**: Hero → Problem → Agitate → Solution → Features → Social Proof → FAQ → Guarantee → CTA
- **Section-by-section copy**: Ready to paste into any page builder
- **Visual direction**: Image and design recommendations per section
- **Conversion optimization tips**: Specific suggestions for testing and improvement

## Agents

This skill dispatches one of two specialist agents based on your request:

### Copywriter Agent
Handles ad copy generation. Expert in AIDA, PAS, BAB, 4Ps, Triple Hook, and Star-Story-Solution frameworks. Produces platform-specific output optimized for conversion.

**Dispatch**: `subagent_type: "copywriter"`

### Landing Page Builder Agent
Handles full sales page creation. Engineers complete conversion experiences with psychological section-by-section progression. Outputs structured markdown ready for implementation.

**Dispatch**: `subagent_type: "landing-page-builder"`

## Frameworks Used

| Framework | Best For |
|-----------|----------|
| **AIDA** (Attention → Interest → Desire → Action) | Long-form ads, email sequences |
| **PAS** (Problem → Agitate → Solution) | Short social ads, Google ads |
| **BAB** (Before → After → Bridge) | Email, testimonial ads |
| **4Ps** (Promise → Picture → Proof → Push) | Product launches, retargeting |
| **Triple Hook** (Premise → Stakes → Twist) | Video hooks, scroll-stoppers |
| **Star-Story-Solution** | Case studies, video scripts |

## Examples

### Example 1: Facebook Ad for a Course
```
/generate-copy --platform facebook
Digital photography course for beginners. $97.
Target: hobbyist photographers who own a DSLR but shoot in auto mode.
They want to take photos that look professional but feel overwhelmed by camera settings.
```

### Example 2: Google Ads for SaaS
```
/generate-copy --platform google --type ad
AI writing assistant for marketers. $49/month.
Competing with Jasper and Copy.ai. Our differentiator: trained specifically on direct response copy.
Target: marketing managers and small business owners running their own ads.
```

### Example 3: Sales Page for Digital Product
```
/generate-copy --type landing-page
"The Meal Prep Blueprint" — $27 digital guide teaching busy professionals how to prep a week of healthy meals in 2 hours on Sunday.
Includes 52 recipes, shopping list generator, and meal prep video library.
Target: 30-45 year old professionals who want to eat healthy but order takeout 4x/week because they're too tired to cook.
```

## Tips for Best Results

1. **Be specific about your audience**: "freelance web developers making $50-100K who want to scale" beats "small business owners"
2. **Include pain points**: The more you tell the agent about what your audience struggles with, the better the copy
3. **Mention competitors**: Helps the agent position your offer differently
4. **Share existing testimonials**: If you have real results or quotes, include them
5. **Specify the goal**: Direct sale? Lead gen? Webinar signup? Different goals = different copy strategies
