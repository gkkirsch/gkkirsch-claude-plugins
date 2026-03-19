---
name: customer-success
description: >
  Premium customer success and sales suite for SaaS founders, sales teams, and CS managers.
  Four specialized agents: (1) Sales Email Writer — creates cold outreach, follow-ups, demo
  sequences, proposal emails, objection handling, win-back campaigns, and upsell emails using
  AIDA, PAS, problem-solution, and social proof frameworks with subject line variants for every
  pipeline stage. (2) Onboarding Designer — builds complete customer onboarding experiences
  including welcome sequences, product walkthroughs, quick-start guides, milestone celebrations,
  training agendas, and knowledge base templates with time-to-value optimization. (3) Churn
  Analyzer — performs customer health scoring, red flag identification, retention playbooks by
  risk tier, win-back campaigns, QBR templates, cohort analysis, and revenue impact assessment
  using MEDDIC and customer health methodologies. (4) Case Study Writer — transforms customer
  wins into compelling case studies, testimonials, video scripts, social proof, and sales deck
  content using Challenge-Solution-Results structure with data-driven storytelling.
  Triggers: "sales email", "cold outreach", "follow-up email", "demo email", "objection handling",
  "customer onboarding", "welcome sequence", "activation", "time to value", "churn", "retention",
  "health score", "NPS", "QBR", "customer health", "case study", "testimonial", "success story",
  "customer story", "win-back", "upsell email", "renewal", "pipeline", "deal", "prospect".
  NOT for: ad copy (use ad-copy-generator), blog content (use content-strategy-suite), proposals
  (use freelancer-toolkit), or code generation.
version: 1.0.0
argument-hint: "<customer/prospect context> [--type email|onboarding|churn|casestudy]"
allowed-tools: Read, Grep, Glob, Bash, Write, Edit
model: sonnet

metadata:
  superbot:
    emoji: "🤝"
---

# Customer Success & Sales Suite

Premium AI-powered customer success and sales toolkit. Like having a revenue team — SDR, customer success manager, retention specialist, and marketing writer — powered by proven SaaS frameworks that close deals, activate customers, prevent churn, and turn wins into growth engines.

## What This Skill Does

Takes your prospect, customer, or pipeline context — and generates:
- **Sales email sequences** for every pipeline stage from cold outreach to post-sale upsell, with subject line variants and personalization
- **Customer onboarding programs** with welcome sequences, walkthrough scripts, quick-start guides, and milestone celebrations
- **Churn risk analysis** with customer health scoring, retention playbooks, QBR templates, and win-back campaigns
- **Case studies** in multiple formats — long-form, testimonial, video script, social proof, and sales deck versions

## Agents

### Sales Email Writer (`sales-email-writer`)

Creates sales emails for every stage of the customer lifecycle using proven copywriting frameworks.

**What it produces**:
- Cold outreach sequences (AIDA, PAS, problem-solution, curiosity gap, social proof)
- Follow-up cadences (day 1, 3, 7, 14, 30 patterns)
- Demo request and confirmation emails
- Proposal delivery and negotiation emails
- Post-demo follow-up sequences
- Objection handling responses (price, timing, competition, status quo, authority)
- Referral request emails
- Win-back campaigns for lost deals
- Upsell and cross-sell emails for existing customers
- Each template includes 3-5 subject line variants

**Dispatch**:
```
Task tool:
  subagent_type: "sales-email-writer"
  description: "Write sales emails for [context]"
  prompt: |
    Pipeline Stage: [cold outreach/follow-up/demo/proposal/negotiation/post-sale/win-back]
    Prospect: [company, role, industry, size]
    Value Proposition: [what your product solves]
    Context: [previous interactions, pain points]
    Tone: [casual/professional/bold/consultative]
  mode: "bypassPermissions"
```

**Example prompts**:
- "Write a 5-email cold outreach sequence targeting VP of Engineering at mid-market SaaS companies"
- "Create objection handling responses for 'we're already using Competitor X'"
- "Build a win-back campaign for enterprise customers who churned in Q3"
- "Draft upsell emails for customers hitting usage limits on their current plan"

### Onboarding Designer (`onboarding-designer`)

Creates complete customer onboarding experiences that accelerate time-to-value.

**What it produces**:
- Welcome email sequences (day 0, 1, 3, 7, 14, 30)
- Product walkthrough scripts for CS calls
- Quick-start guides by customer segment
- Milestone celebration messages
- Check-in email templates
- Training session agendas
- Knowledge base article templates
- FAQ generation frameworks
- Success metric definition (activation, adoption, engagement)
- Onboarding health dashboards

**Dispatch**:
```
Task tool:
  subagent_type: "onboarding-designer"
  description: "Design onboarding for [customer type]"
  prompt: |
    Customer Type: [self-serve/SMB/mid-market/enterprise]
    Product: [what they bought]
    Key Activation Metrics: [what defines success]
    Timeline: [first 7/14/30/90 days]
    Touchpoint Preference: [high-touch/low-touch/hybrid]
  mode: "bypassPermissions"
```

**Example prompts**:
- "Design a self-serve onboarding program for a project management tool targeting small teams"
- "Create a high-touch enterprise onboarding for a $50K ARR data analytics platform"
- "Build a 30-day welcome email sequence for a marketing automation tool"
- "Design an onboarding health dashboard with activation and adoption metrics"

### Churn Analyzer (`churn-analyzer`)

Analyzes customer data for churn risk and builds retention strategies.

**What it produces**:
- Customer health scoring methodology (usage, engagement, support, NPS)
- Red flag identification with severity levels
- Retention playbooks by risk tier (green/yellow/red)
- Win-back campaign templates
- Exit interview question frameworks
- Cohort analysis prompts
- Revenue impact analysis
- QBR (Quarterly Business Review) templates
- Customer advocacy program design
- Expansion revenue strategies

**Dispatch**:
```
Task tool:
  subagent_type: "churn-analyzer"
  description: "Analyze churn risk for [context]"
  prompt: |
    Customer Data: [file path or description]
    Health Signals: [usage, support tickets, NPS, engagement]
    Contract Details: [renewal date, ARR, plan tier]
    Goal: [prevent churn/build model/design QBR/exit analysis]
  mode: "bypassPermissions"
```

**Example prompts**:
- "Build a customer health scoring model for our SaaS platform with 500 mid-market customers"
- "Analyze this cohort's usage data and identify churn risk patterns"
- "Create a retention playbook for customers showing declining login frequency"
- "Design a QBR template for enterprise accounts worth $100K+ ARR"

### Case Study Writer (`case-study-writer`)

Transforms customer success stories into compelling marketing and sales assets.

**What it produces**:
- Long-form case studies (1,500-2,500 words) with Challenge → Solution → Results
- Short testimonials (100-200 words) for website and email
- Video script versions with B-roll suggestions
- Interview question templates for customer interviews
- Data-driven storytelling with metrics and before/after
- Quote extraction and formatting
- Industry-specific templates
- Blog post, social media, email, and sales deck versions

**Dispatch**:
```
Task tool:
  subagent_type: "case-study-writer"
  description: "Write case study for [customer]"
  prompt: |
    Customer: [company, industry, size]
    Challenge: [problem before your product]
    Solution: [how they use your product]
    Results: [metrics, outcomes, quotes]
    Target Audience: [who this should persuade]
    Format: [long-form/testimonial/video/social/sales-deck]
  mode: "bypassPermissions"
```

**Example prompts**:
- "Write a case study about how TechCorp reduced support tickets by 60% using our AI chatbot"
- "Create a video testimonial script for a customer who grew revenue 3x after implementing our platform"
- "Turn this customer interview transcript into a blog-style case study and social media snippets"
- "Build a sales deck case study slide for our biggest enterprise win"

## Slash Command

Use `/sales-pipeline` for quick access:

```
/sales-pipeline Write a cold outreach sequence for fintech CFOs
```

```
/sales-pipeline --type onboarding Design enterprise onboarding for our analytics platform
```

```
/sales-pipeline --type churn Build a health scoring model for our 200 SMB accounts
```

```
/sales-pipeline --type casestudy Turn Acme's 40% cost reduction into a case study
```

```
/sales-pipeline Full lifecycle package for a B2B SaaS selling to HR teams
```

## Frameworks Included

| Framework | Used In | Best For |
|-----------|---------|----------|
| **AIDA** | Sales Email Writer | Cold outreach structure |
| **PAS (Problem-Agitate-Solve)** | Sales Email Writer | Pain-focused emails |
| **SPIN Selling** | Sales Email Writer, Churn Analyzer | Discovery and needs analysis |
| **BANT** | Sales Email Writer | Lead qualification |
| **MEDDIC** | Sales Email Writer, Churn Analyzer | Enterprise deal qualification |
| **Jobs-to-Be-Done** | Onboarding Designer | Activation design |
| **Customer Health Scoring** | Churn Analyzer | Risk identification |
| **NPS/CSAT Programs** | Churn Analyzer | Voice of customer |
| **Challenge-Solution-Results** | Case Study Writer | Story structure |
| **StoryBrand** | Case Study Writer | Narrative frameworks |

## Reference Library

Deep domain knowledge is available in the `references/` directory:
- **sales-email-templates.md** — 50+ actual email templates organized by pipeline stage with subject lines, bodies, and personalization placeholders
- **onboarding-playbooks.md** — Complete SaaS onboarding frameworks for self-serve through enterprise, with metrics, failure patterns, and milestone-based approaches
- **retention-strategies.md** — Customer health score models, early warning indicators, intervention playbooks, NPS/CSAT program design, QBR frameworks, and expansion revenue strategies

## Tips for Best Results

1. **Be specific about your ICP**: "VP of Engineering at 200-person B2B SaaS" beats "business leaders"
2. **Include real context**: Company names, pain points, previous interactions — specificity drives quality
3. **State the pipeline stage**: Cold vs warm vs existing customer changes everything
4. **Provide metrics**: Revenue numbers, usage data, NPS scores — data makes output actionable
5. **Chain agents for lifecycle coverage**: Email → Onboarding → Health Check → Case Study
