---
name: audit-landing-page
description: Run a comprehensive CRO audit on the landing page in this codebase
allowed-tools:
  - Read
  - Glob
  - Grep
  - Bash
  - Write
  - Edit
model: sonnet
---

# Audit Landing Page

Run a comprehensive Conversion Rate Optimization audit on the landing page in this codebase.

## Instructions

1. **Find the landing page files** by searching the codebase:
   - Glob for `**/*landing*.*`, `**/*home*.*`, `**/pages/index.*`, `**/app/page.*`
   - Grep for "hero", "cta", "call-to-action", "landing" in component directories
   - Look in common locations: `src/pages/`, `src/app/`, `pages/`, `app/`, `src/components/`
   - Check for layout files that wrap the landing page

2. **Read all relevant files** including:
   - The main landing page file
   - Any components imported by the landing page (hero, features, pricing, testimonials, footer)
   - CSS/Tailwind configuration for design tokens
   - Any A/B testing configuration if present

3. **Dispatch the CRO Auditor agent** with the page content and structure.

4. **Deliver the full audit** including:
   - Overall conversion score (1-100)
   - Section-by-section analysis
   - MECLABS formula breakdown
   - Prioritized recommendations (quick wins, medium-term, strategic)
   - Specific code changes for the top 5 recommendations
   - A/B testing roadmap

If no landing page is found, inform the user and ask them to specify which files to audit,
or offer to help them create a high-converting landing page from scratch using the
conversion-copywriter agent and conversion-patterns reference.
