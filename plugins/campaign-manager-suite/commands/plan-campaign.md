---
name: plan-campaign
description: Quickly plan a multi-channel marketing campaign. Provide your product, budget, timeline, and goals to receive a complete campaign strategy with channel mix, budget allocation, creative briefs, and KPIs.
argument-hint: "<product description> | budget: <amount> | timeline: <weeks> | goal: <objective>"
agent: ./agents/campaign-strategist.md
---

# Plan Campaign

Use this command to generate a comprehensive marketing campaign plan.

## Usage

```
/plan-campaign My SaaS product for project managers | budget: $15K | timeline: 8 weeks | goal: 500 free trial signups
```

## What You'll Get

1. **Campaign Strategy** — Objectives, positioning, and messaging framework
2. **Channel Mix** — Which channels to use and why, with percentage budget allocation
3. **Budget Breakdown** — Exact dollar amounts per channel, per week
4. **Campaign Calendar** — Week-by-week timeline with specific actions and deliverables
5. **Creative Briefs** — What assets to create for each channel
6. **KPIs & Targets** — Specific metrics to track with target numbers
7. **Contingency Plans** — What to do if campaigns underperform at week 2, 4, 6
8. **Launch Checklist** — Step-by-step pre-launch verification

## Required Information

| Parameter | Description | Example |
|-----------|-------------|---------|
| Product | What you're marketing | "AI-powered email tool for sales teams" |
| Budget | Total campaign budget | "$10,000" or "$5K/month" |
| Timeline | Campaign duration | "8 weeks" or "Q2 2026" |
| Goal | Primary objective | "Generate 200 qualified leads" |

## Optional Context

You can also provide:
- **Target audience** — Demographics, firmographics, psychographics
- **Previous results** — What has worked or failed before
- **Brand guidelines** — Tone, voice, visual constraints
- **Competitive landscape** — Key competitors and positioning
- **Existing assets** — Landing pages, email lists, social accounts
