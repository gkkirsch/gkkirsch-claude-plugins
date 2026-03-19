# Onboarding Playbooks — Complete SaaS Customer Onboarding Frameworks

Comprehensive onboarding frameworks for SaaS products across all customer tiers. Includes step-by-step playbooks, metrics, failure patterns, and recovery strategies. This is not theory — these are executable playbooks based on patterns from 200+ SaaS companies.

---

## 1. SaaS Onboarding: The Foundation

### The Onboarding Spectrum

```
Low-Touch (Self-Serve)          ←→          High-Touch (White-Glove)
─────────────────────────────────────────────────────────────────────
$0-$5K ACV                                           $100K+ ACV
Individual users                                     Enterprise teams
Simple product                                       Complex platform
Product-led                                          Sales-led
Automated emails                                     Dedicated CSM + PM
In-app guides                                        Custom training
Community support                                    24/7 premium support
Days to activate                                     Weeks to implement
```

### The 5 Phases of Onboarding

Every SaaS onboarding, regardless of tier, moves through these phases:

**Phase 1: Orientation (Day 0-1)**
- Goal: Customer knows what they bought and what happens next
- Key moment: First login / first email
- Risk: Confusion, buyer's remorse, overwhelm

**Phase 2: Setup (Day 1-7)**
- Goal: Product is configured and ready to use
- Key moment: First integration / first data import
- Risk: Technical blockers, abandonment, "I'll do this later"

**Phase 3: Activation (Day 7-30)**
- Goal: Customer has experienced core value
- Key moment: "Aha!" — the moment they see the product work for them
- Risk: Using product but not the RIGHT features, shallow adoption

**Phase 4: Habit Formation (Day 14-60)**
- Goal: Customer uses product regularly as part of their workflow
- Key moment: Third session without prompting
- Risk: Novelty wears off, competing priorities, single-user dependency

**Phase 5: Expansion (Day 30-90)**
- Goal: Customer adopts additional features, invites more users
- Key moment: First team member invited, first advanced feature used
- Risk: Stagnation, "good enough" syndrome, champion leaves

### The Activation Ladder

Define a clear ladder for what "activated" means. Generic "activation" is meaningless — make it measurable:

```
Example: Project Management SaaS

Level 0: Account created
Level 1: Profile configured (photo, role, timezone)
Level 2: First project created
Level 3: First task assigned to a team member
Level 4: Team member completes first task
Level 5: 3+ tasks completed by 2+ users in one week    ← ACTIVATED
Level 6: Integration connected (Slack, GitHub, etc.)
Level 7: Weekly report viewed
Level 8: Second project created                         ← POWER USER
```

```
Example: Analytics SaaS

Level 0: Account created
Level 1: Tracking code installed
Level 2: First data flowing (>100 events)
Level 3: First dashboard created
Level 4: First insight shared with team                 ← ACTIVATED
Level 5: Alert/notification configured
Level 6: Custom report built
Level 7: API integration active                         ← POWER USER
```

```
Example: CRM SaaS

Level 0: Account created
Level 1: Contacts imported (>50)
Level 2: First deal created
Level 3: First email sent from CRM
Level 4: Pipeline stages customized
Level 5: 5+ deals in pipeline, 2+ team members active  ← ACTIVATED
Level 6: Reporting dashboard configured
Level 7: Automation rules created                       ← POWER USER
```

---

## 2. Self-Serve Onboarding Playbook

### The Product-Led Onboarding Framework

**Principle**: The product IS the onboarding. Every in-app experience should guide users toward activation without requiring human intervention.

### In-App Onboarding Patterns

**Pattern 1: The Checklist**
```
Welcome to {{PRODUCT}}! Complete these steps to get started:

□ Set up your profile                    [2 min]
□ Connect your {{INTEGRATION}}           [3 min]
□ Create your first {{OBJECT}}           [5 min]
□ Invite a team member                   [1 min]
□ Complete your first {{VALUE_ACTION}}   [5 min]

Progress: ████░░░░░░ 40% complete
```

**When to use**: Products with clear setup steps. Works best with 5-7 items (not 15).

**Pattern 2: The Interactive Tour**
```
Step 1/5: "This is your dashboard"
[Highlight dashboard area with tooltip]
[Button: "Show me around" / "Skip tour"]

Step 2/5: "Create your first project"
[Arrow pointing to "New Project" button]
[Button: "Create now" / "I'll do this later"]

Step 3/5: "Invite your team"
[Highlight team settings with animation]
[Button: "Invite teammates" / "Skip for now"]
```

**When to use**: Visual products where seeing beats reading. Keep under 5 steps.

**Pattern 3: The Template Start**
```
"Start with a template"

[Card: Sales Pipeline]        [Card: Customer Support]
"Track deals from lead        "Manage tickets and
to close"                     response times"

[Card: Project Management]    [Card: Start from scratch]
"Plan sprints and             "Build your own from
track progress"               the ground up"
```

**When to use**: Products with multiple use cases. Templates reduce blank-canvas paralysis.

**Pattern 4: The Progressive Disclosure**
```
Week 1: Show only core features (3-4 items in nav)
Week 2: Reveal intermediate features (notification: "Unlock reports →")
Week 3: Surface advanced features (tooltip: "Power users love this →")
Week 4: Full product access
```

**When to use**: Complex products that overwhelm new users. Show more as they demonstrate readiness.

### Behavioral Email Triggers

**Critical**: Self-serve onboarding emails should be behavior-triggered, not time-based.

| Trigger | Email | Timing |
|---------|-------|--------|
| Signed up, didn't start setup | "Let's finish setting up" — direct link to next step | 2 hours |
| Started setup, didn't finish | "Pick up where you left off" — progress bar + direct link | 24 hours |
| Setup complete, didn't use core feature | "Here's the feature that matters most" — spotlight with example | 24 hours |
| Used core feature once | "Nice! Here's what to try next" — next step in activation ladder | 48 hours |
| Hit first milestone | "You just {{ACHIEVEMENT}}!" — celebration + social proof | Immediately |
| 3 days inactive | "We miss you" — value reminder + quick action link | 72 hours |
| 7 days inactive | "Is everything ok?" — offer help + alternative resource | 7 days |
| 14 days inactive | "Before you go" — last chance + feedback request | 14 days |
| Invited team member | "Your team is growing" — team tips + admin features | Immediately |
| Team member activated | "Your team is using {{PRODUCT}}" — team metrics | 24 hours |

### Self-Serve Metrics Dashboard

| Metric | Formula | Target | Alert Threshold |
|--------|---------|--------|-----------------|
| Signup-to-Setup | % who complete setup within 24h | >60% | <40% |
| Setup-to-Activation | % who activate within 14 days | >30% | <15% |
| Day 1 Return | % who return within 24 hours | >40% | <25% |
| Day 7 Return | % who return within 7 days | >25% | <15% |
| Day 30 Retention | % active at 30 days | >20% | <10% |
| Time to First Value | Median time from signup to first value action | <5 min | >15 min |
| Activation Rate | % reaching "activated" level in 30 days | >25% | <10% |
| Checklist Completion | % completing all onboarding steps | >40% | <20% |

---

## 3. SMB Onboarding Playbook (Guided Self-Serve)

### The Hybrid Model

SMB customers need more guidance than self-serve but can't justify dedicated CSM time. The solution: automated sequences enhanced with human touchpoints at critical moments.

### 30-Day Playbook

```
DAY 0: WELCOME
─────────────────
Automated:
- Welcome email with login credentials and quick-start guide
- In-app welcome checklist activated
- Product tour triggered on first login

Human (optional):
- 15-minute welcome call or personalized video (Loom/Vidyard)
- Agenda: confirm goals, explain what to expect, identify quick win

Success criteria: Customer has logged in and viewed the dashboard
──────────────────────────────────────────────────────────────────

DAY 1-2: SETUP
─────────────────
Automated:
- Email: "Your 3-step setup guide" with direct links
- In-app: Setup wizard with progress tracking
- Tooltip: First key feature highlighted

Human (if struggling):
- Chat support available for setup questions
- CSM checks setup completion status

Success criteria: Core setup complete (integrations, data import, team invite)
──────────────────────────────────────────────────────────────────

DAY 3-5: FIRST VALUE
─────────────────
Automated:
- Email: "The feature that changes everything" — core value feature spotlight
- In-app: Guided workflow for first value action
- Triggered: If first value action completed → celebration email

Human (if stuck):
- CSM sends personalized Loom video showing their specific use case
- Chat outreach: "I noticed you haven't tried {{FEATURE}} yet — need help?"

Success criteria: First value action completed
──────────────────────────────────────────────────────────────────

DAY 7: CHECK-IN
─────────────────
Automated:
- Email: "How's your first week?" — progress summary + next steps
- In-app: "You've completed X of Y steps" with encouragement
- Survey: 1-question pulse ("How easy has setup been? 1-5")

Human (if score < 3):
- CSM sends chat message offering 15-minute troubleshooting call
- Check for incomplete setup steps and send targeted help

Success criteria: Customer has used core feature at least twice
──────────────────────────────────────────────────────────────────

DAY 10-14: DEEPENING
─────────────────
Automated:
- Email: "Power user tip: {{ADVANCED_FEATURE}}" — introduce secondary feature
- In-app: Badge or milestone for reaching activation threshold
- Triggered: If team member invited → "Team collaboration tips" email

Human (if at-risk):
- CSM reviews usage data — flags accounts with <3 sessions
- Outreach email: "Quick question — are you getting what you expected?"
- Offer webinar invite for hands-on group training session

Success criteria: Customer has reached activation level on the ladder
──────────────────────────────────────────────────────────────────

DAY 14: MIDPOINT REVIEW
─────────────────
Automated:
- Email: "Your 2-week progress report" — usage stats, achievements unlocked
- In-app: Milestone celebration if activated; nudge if not
- Trigger: If not activated → escalate to CSM queue for manual review

Success criteria: Clear understanding of customer health; rescue plan if needed
──────────────────────────────────────────────────────────────────

DAY 15-21: HABIT BUILDING
─────────────────
Automated:
- Email: "Weekly digest: Here's what happened in {{PRODUCT}} this week"
- In-app: "Have you tried?" prompts for unused features relevant to their use case
- Triggered: If 3+ sessions this week → "You're a power user" recognition email

Human (if stalling):
- CSM shares case study from similar company in same industry/size
- Offer to set up a custom workflow or template tailored to their process
- Slack/Teams connect: Invite to customer community channel

Success criteria: Customer logs in 3+ times this week without prompting
──────────────────────────────────────────────────────────────────

DAY 21-25: EXPANSION SEEDS
─────────────────
Automated:
- Email: "Did you know? {{PRODUCT}} also does..." — introduce adjacent feature
- In-app: "Invite more teammates" prompt with easy sharing link
- Triggered: If 2+ users active → "Team plan" upgrade nudge (soft, not pushy)

Human:
- CSM identifies expansion opportunities from usage patterns
- If champion identified: brief "admin tips" email with advanced settings
- Plant seeds: "Some of our customers in {{INDUSTRY}} also use us for {{USE_CASE_2}}"

Success criteria: At least one expansion signal (new user, new feature, upgrade inquiry)
──────────────────────────────────────────────────────────────────

DAY 30: GRADUATION
─────────────────
Automated:
- Email: "Your first month in review" — full metrics summary with visuals
- In-app: Onboarding checklist marked complete; transition to "Tips & Tricks" mode
- Survey: Onboarding NPS — "How would you rate your onboarding experience? (0-10)"
- Trigger: Auto-transition from onboarding segment to active-customer segment with a named CSM, regular cadence, and defined milestones. These customers have real budgets, real stakeholders, and real expectations — but not the unlimited patience of enterprise.

### Pre-Kickoff Preparation (Before Day 0)

**CSM receives the deal from Sales and prepares before the customer ever logs in.**

```
Pre-Kickoff Checklist:
□ Review CRM notes: deal summary, champion, decision-maker, buying criteria
□ Review recorded sales demos/calls for promised features and timelines
□ Identify primary use case and success metrics agreed during sales
□ Confirm license count, tier, and any custom terms
□ Set up customer in onboarding tracking system (Gainsight, Vitally, Totango, etc.)
□ Prepare personalized onboarding deck with their logo, goals, and timeline
□ Pre-configure account: workspace name, SSO settings if applicable, pilot data
□ Schedule kickoff meeting within 3 business days of close
□ Send pre-kickoff email with agenda, attendee request, and prep instructions
```

### Kickoff Meeting Structure (60 minutes)

```
AGENDA:
─────────────────────────────────────────────
[5 min]  Introductions & roles
         - CSM introduces themselves, their role, and availability
         - Customer introduces team members and their roles

[10 min] Goal Alignment
         - "Here's what we heard during the sales process: {{GOALS}}"
         - Confirm or adjust: primary goal, secondary goals, timeline
         - Define what success looks like in their words

[15 min] Product Overview & Quick Win
         - Walkthrough of their pre-configured workspace
         - Live demo of the ONE feature most tied to their primary goal
         - Customer completes first action hands-on (screen share)

[10 min] Onboarding Plan Review
         - Walk through the 90-day timeline with milestones
         - Identify internal champion and executive sponsor
         - Agree on meeting cadence (weekly for month 1, biweekly after)

[10 min] Technical Setup Discussion
         - Integration requirements and timeline
         - Data migration plan (who owns what, format, timeline)
         - SSO/security requirements if applicable

[5 min]  Immediate Next Steps
         - Action items with owners and due dates
         - CSM sends recap email within 2 hours
         - Next meeting scheduled before anyone leaves the call

[5 min]  Open Q&A
```

### Week-by-Week Cadence

**Week 1: Foundation**
- CSM ensures all technical setup is complete (integrations, data import, user provisioning)
- Daily check-in via email or Slack: "Any blockers? Need anything?"
- Deliver: Admin training for the champion (30 min, recorded)
- Milestone: All users can log in; primary integration is live

**Week 2: Core Adoption**
- 30-minute call: Review setup, train on primary workflow
- Share: Workflow template customized to their process
- Assign: Champion creates first real project/campaign/pipeline using the product
- Milestone: Champion completes first value action using real data

**Week 3: Team Rollout**
- 30-minute call: Review champion's experience, plan team training
- Deliver: Team training session (45 min, recorded for absent members)
- Share: Quick-reference guide or cheat sheet for end users
- Milestone: 50% of licensed users have logged in and completed setup

**Week 4: Activation Push**
- 30-minute call: Review adoption metrics, address friction points
- CSM reviews: Which users are active? Which are stuck? Why?
- Action: Personalized outreach to inactive users (email from CSM, not automated)
- Milestone: 3+ users active in the last 7 days; primary workflow used 5+ times

**Month 2: Deepening (Biweekly Calls)**
- Biweekly 30-minute calls focused on:
  - Week 5-6: Secondary feature adoption — introduce reporting, automation, or advanced capabilities
  - Week 7-8: Workflow optimization — review how they're using the product vs. best practices; suggest improvements
- CSM shares: Monthly usage report with benchmarks ("You're in the top 30% of similar companies")
- CSM plants seeds: "Companies like yours often expand into {{USE_CASE_2}} around this time"
- Milestone: At least one secondary feature adopted; usage is consistent week-over-week

**Month 3: Go-Live & Hypercare**
- Biweekly 30-minute calls focused on:
  - Week 9-10: Full go-live — confirm all planned users are onboarded and active
  - Week 11-12: Optimization and expansion planning
- Go-Live Support:
  - CSM available for ad-hoc questions during go-live week (extended Slack/email SLA)
  - Daily usage monitoring with alerts for drop-offs
  - Quick-response troubleshooting for any workflow issues
- Milestone: All licensed users active; primary KPI improvement demonstrated

### 90-Day Success Review (Executive-Level)

```
AGENDA: 90-Day Business Review (45 minutes)
─────────────────────────────────────────────
Attendees: Executive sponsor, champion, CSM, CSM manager (optional)

[10 min] Results Review
         - Usage metrics vs. benchmarks
         - Progress against goals set at kickoff
         - ROI indicators (time saved, revenue influenced, efficiency gained)

[10 min] Voice of the User
         - Feedback from end users (quotes, survey results)
         - Feature requests and product feedback
         - What's working well vs. what needs improvement

[10 min] Next Phase Planning
         - Goals for the next quarter
         - Expansion opportunities (more users, more features, higher tier)
         - Upcoming product releases relevant to their use case

[10 min] Relationship Check
         - Onboarding NPS score
         - CSM satisfaction
         - Escalation path review

[5 min]  Action Items & Close
```

### Mid-Market Success Metrics

| Metric | Target | Measured At |
|--------|--------|-------------|
| Kickoff scheduled within | 3 business days of close | Day 0 |
| Technical setup complete | 100% within 7 days | Week 1 |
| Champion activation | First value action within 14 days | Week 2 |
| Team adoption | 50% of users active within 30 days | Month 1 |
| Full adoption | 80% of users active within 60 days | Month 2 |
| Feature breadth | 2+ features in regular use within 60 days | Month 2 |
| Go-live complete | All users onboarded within 75 days | Month 3 |
| Onboarding NPS | 8+ average | Day 90 |
| TTV (Time to Value) | <14 days to first value action | Week 2 |
| Renewal confidence | CSM rates 4+ out of 5 | Day 90 |

---

## 5. Enterprise Onboarding Playbook

Enterprise accounts ($100K+ ACV) demand a multi-phase, multi-stakeholder onboarding program. The stakes are higher, the timelines are longer, and failure is expensive. A single enterprise churn can wipe out an entire quarter of SMB new business.

### Sales-to-CS Handoff Framework

**The handoff is where enterprise onboarding succeeds or fails.** A bad handoff means the CSM starts blind, the customer repeats themselves, and trust erodes before onboarding even begins.

```
REQUIRED HANDOFF DOCUMENT (Sales → CS):
─────────────────────────────────────────────
Section 1: Deal Summary
- Company name, industry, size (employees, revenue)
- ACV, contract term, payment terms
- Products/modules purchased, license count
- Custom terms, SLA commitments, or contractual obligations
- Competitive alternatives evaluated (and why they chose us)

Section 2: Stakeholder Map
- Executive sponsor: Name, title, email, communication preference
- Champion: Name, title, email, relationship strength (1-5)
- Technical lead: Name, title, email, technical sophistication level
- End users: Team(s), count, roles, current tools they're replacing
- Detractors: Anyone who opposed the purchase (name, concern, mitigation)

Section 3: Goals & Success Criteria
- Primary business goal (in the customer's exact words)
- Secondary goals
- Specific metrics they expect to improve (and by how much)
- Timeline expectations communicated during sales
- What "failure" looks like to the executive sponsor

Section 4: Technical Landscape
- Current tech stack (integrations needed on day 1 vs. later)
- Data migration requirements (volume, format, source systems)
- Security/compliance requirements (SOC 2, HIPAA, GDPR, SSO provider)
- IT involvement level (self-serve vs. IT-gated)
- Known technical risks or blockers

Section 5: Relationship Context
- Sales cycle length and key moments
- Objections raised and how they were addressed
- Promises made (features, timelines, support levels)
- Competitor POCs conducted (results, feedback)
- Internal politics or sensitivities to be aware of
```

**Handoff Meeting (30 min, Sales AE + CSM + CS Manager)**
- AE walks through the handoff document live
- CSM asks clarifying questions
- Agree on introduction timing and format (warm intro email from AE)
- AE remains available for first 30 days for relationship continuity

### Multi-Phase Implementation Approach

```
PHASE 0: PLANNING (Week -2 to 0)
─────────────────────────────────
- Internal: CSM reviews handoff, prepares onboarding plan, assembles team
- Internal: Assign implementation resources (Solutions Engineer, if applicable)
- External: AE sends warm intro email connecting customer to CSM
- External: CSM sends welcome packet (onboarding plan, timeline, team contacts)
- Gate: Kickoff meeting scheduled with all key stakeholders

PHASE 1: DISCOVERY & DESIGN (Week 1-2)
─────────────────────────────────────────
- Executive kickoff meeting (90 min) with full stakeholder group
- Technical discovery session: integrations, data, security requirements
- Workflow mapping: Document current processes → design future state in product
- Success plan: Formalize goals, metrics, milestones, and RACI
- Gate: Signed-off success plan and technical architecture approved

PHASE 2: BUILD & CONFIGURE (Week 3-6)
─────────────────────────────────────────
- Environment setup: SSO, permissions, workspace configuration
- Integration development: Connect source systems, validate data flow
- Data migration: Extract, transform, load; validate with customer
- Custom configuration: Workflows, templates, automations, dashboards
- Gate: Staging environment ready for UAT (User Acceptance Testing)

PHASE 3: PILOT (Week 6-10)
─────────────────────────────────────────
- Select pilot group: 1 department, 10-25 users, representative use case
- Pilot training: 2-hour hands-on session + recorded walkthrough
- Pilot support: Dedicated Slack channel, 4-hour response SLA
- Weekly pilot reviews: Usage data, feedback, bug reports, change requests
- Gate: Pilot success criteria met (see below); go/no-go decision for rollout

PHASE 4: ROLLOUT (Week 10-16)
─────────────────────────────────────────
- Department-by-department rollout (2-3 departments per wave)
- Per-wave: Training session (role-specific), go-live support, hypercare period
- Change management: Internal comms from executive sponsor, FAQs, help resources
- Migration: Decommission old tools, redirect workflows
- Gate: All departments live; 80% user adoption; no critical open issues

PHASE 5: OPTIMIZE & EXPAND (Week 16-24)
─────────────────────────────────────────
- Usage optimization: Review workflows, suggest improvements, benchmark vs. peers
- Executive Business Review at 90 days (see below)
- Expansion planning: Additional departments, use cases, or product modules
- Transition: From onboarding cadence to ongoing success cadence
- Gate: Customer confirms they have achieved initial business goals
```

### Stakeholder Management (RACI for Onboarding)

| Activity | Executive Sponsor | Champion | Technical Lead | End Users | CSM | Solutions Engineer |
|----------|:-:|:-:|:-:|:-:|:-:|:-:|
| Goal setting | A | R | C | I | R | I |
| Success plan approval | A | R | C | I | R | I |
| Technical architecture | I | C | A | I | C | R |
| Integration development | I | I | A | I | C | R |
| Data migration | I | C | R | I | C | A |
| Pilot user selection | A | R | C | I | C | I |
| Training delivery | I | C | I | R | A | C |
| Go-live decision | A | R | C | I | R | C |
| Change management comms | A | R | I | I | C | I |
| Executive Business Review | R | R | C | I | A | C |

*R = Responsible, A = Accountable, C = Consulted, I = Informed*

### Pilot Design & Success Criteria

**Selecting the Pilot Group:**
- Choose a department that is enthusiastic (not skeptical)
- Pick a use case that is representative but not the most complex
- Size: 10-25 users (large enough for valid data, small enough to support closely)
- Duration: 3-4 weeks (long enough for habit formation, short enough to maintain momentum)

**Pilot Success Criteria (define before pilot starts):**

| Criteria | Target | Measurement |
|----------|--------|-------------|
| User adoption | 80% of pilot users active weekly | Product analytics |
| Feature usage | Core workflow completed 3+ times per user | Product analytics |
| Data quality | <5% error rate on migrated/imported data | Audit report |
| User satisfaction | Average rating 4+ out of 5 | End-of-pilot survey |
| Performance | No critical bugs; <2 sec page load | QA + monitoring |
| Business impact | Measurable improvement in pilot KPI | Before/after comparison |
| Champion confidence | Champion recommends proceeding | Direct conversation |

### Rollout Planning (Department by Department)

```
WAVE PLANNING TEMPLATE:
─────────────────────────────────────────────
Wave 1: {{DEPARTMENT_A}} (20 users) — Week 10-11
  - Training: Monday 10am (role: Managers), Tuesday 2pm (role: ICs)
  - Go-live: Wednesday
  - Hypercare: Wed-Fri (CSM + SE on standby, 2-hour SLA)
  - Week 2: Monitor, address issues, declare stable

Wave 2: {{DEPARTMENT_B}} (35 users) — Week 12-13
  - Incorporate lessons learned from Wave 1
  - Training: Monday (adjusted based on Wave 1 feedback)
  - Go-live: Wednesday
  - Hypercare: Wed-Fri

Wave 3: {{DEPARTMENT_C}} (50 users) — Week 14-16
  - Largest wave — schedule extra support resources
  - Stagger go-live across 2 days if >40 users
  - Hypercare: Full week
```

### Executive Business Review at 90 Days

```
ENTERPRISE EBR AGENDA (60 minutes)
─────────────────────────────────────────────
Attendees: Executive sponsor, VP/C-level if possible, champion,
           CSM, CS Manager, AE (optional)

[15 min] Business Impact Report
         - Goals set at kickoff vs. actual results
         - Quantified ROI: hours saved, revenue influenced, cost reduced
         - Comparison: before vs. after (with specific data points)
         - User adoption: total active, engagement trends, power users

[10 min] Implementation Summary
         - Timeline: planned vs. actual (explain any variances)
         - Integrations live, data migrated, customizations delivered
         - Open items and resolution timeline

[10 min] User Feedback & Product Health
         - NPS/CSAT scores from end users
         - Top 3 things users love
         - Top 3 friction points or feature requests
         - Product roadmap items relevant to their needs

[15 min] Strategic Planning
         - Next quarter goals and success metrics
         - Expansion opportunities: departments, use cases, modules
         - Upcoming product releases and beta program invitations
         - Contract renewal timeline and process overview

[10 min] Relationship & Support
         - Ongoing cadence: monthly strategic calls, quarterly EBRs
         - Support escalation paths and SLAs
         - Executive alignment: anything the sponsor needs from leadership
         - Action items with owners and deadlines
```

---

## 6. Onboarding Metrics: What to Measure

Measuring onboarding is not optional — it is the difference between guessing and knowing. Every onboarding program should track these metrics, segmented by customer tier.

### TTV (Time to Value) Calculation Methods

**Method 1: First Value Action**
- Definition: Time from account creation to the first action that delivers measurable value
- Example: For a CRM, TTV = time from signup to first deal closed using the platform
- Pros: Simple, clear, directly tied to value
- Cons: May be too narrow; ignores incremental value along the way

**Method 2: Activation Milestone**
- Definition: Time from account creation to reaching the "activated" level on the activation ladder
- Example: For project management SaaS, TTV = time until 3+ tasks completed by 2+ users in one week
- Pros: Captures team adoption, not just individual usage
- Cons: Requires well-defined activation ladder (most companies skip this)

**Method 3: Customer-Declared Value**
- Definition: Time from account creation to the customer saying "this is working for us"
- Measured via: Check-in calls, survey responses, NPS comments
- Pros: Captures perceived value, which drives retention
- Cons: Subjective, harder to automate, depends on CSM diligence

**Benchmark TTV by Tier:**

| Tier | Target TTV | Alert Threshold |
|------|-----------|-----------------|
| Self-serve | <1 day | >3 days |
| SMB | <7 days | >14 days |
| Mid-market | <14 days | >30 days |
| Enterprise | <30 days | >60 days |

### Activation Rate Benchmarks by Tier

| Tier | 7-Day Activation | 14-Day Activation | 30-Day Activation | 90-Day Activation |
|------|-----------------|-------------------|-------------------|-------------------|
| Self-serve | 15-20% | 20-30% | 25-35% | 30-40% |
| SMB | 25-35% | 35-50% | 50-65% | 60-75% |
| Mid-market | 30-40% | 45-60% | 60-75% | 75-85% |
| Enterprise | 20-30% | 35-50% | 55-70% | 80-90% |

*Note: Enterprise starts slower because setup is more complex, but should reach higher final activation due to dedicated support.*

### Onboarding NPS

Run onboarding-specific NPS at the end of the onboarding period (not product NPS — this is about the onboarding experience itself).

**Question**: "On a scale of 0-10, how would you rate your onboarding experience with {{PRODUCT}}?"

**Benchmarks:**
- Excellent: NPS 50+
- Good: NPS 30-50
- Needs improvement: NPS 10-30
- Critical: NPS below 10

**Follow-up segmentation:**
- Promoters (9-10): Ask for referral, case study, or G2 review
- Passives (7-8): Ask what would have made it a 9 or 10
- Detractors (0-6): Immediate CSM escalation; root cause analysis within 48 hours

### Feature Adoption Scoring

Assign points to each feature based on its correlation with long-term retention. Track cumulative score per account.

```
Feature Adoption Score Example:
─────────────────────────────────────────────
Core workflow completed          +10 pts
Integration connected            +8 pts
Team member invited              +8 pts
Second team member active        +5 pts
Report/dashboard created         +6 pts
Automation configured            +7 pts
API key generated                +4 pts
Mobile app installed             +3 pts
Custom template created          +5 pts
Admin settings configured        +3 pts
─────────────────────────────────────────────
Max score: 59 pts

Thresholds:
0-15:  Red    — at risk of churn, immediate intervention
16-30: Yellow — progressing but incomplete, needs nudging
31-45: Green  — healthy adoption, monitor and expand
46-59: Blue   — power user, expansion and advocacy candidate
```

### Setup Completion Funnel

Track drop-off at every step of the onboarding funnel. This is the single most actionable metric — it tells you exactly where customers get stuck.

```
Example Setup Funnel:
─────────────────────────────────────────────
Account created         1,000   100%
Email verified            920    92%   ← 8% drop: fix email deliverability
First login               780    78%   ← 14% drop: improve welcome email urgency
Profile completed         650    65%   ← 13% drop: simplify profile, make optional
Integration connected     420    42%   ← 23% drop: THIS IS YOUR BIGGEST PROBLEM
First value action        310    31%   ← 11% drop: improve guided workflow
Activation milestone      180    18%   ← 13% drop: improve feature discovery
Second week active        120    12%   ← 6% drop: improve habit formation emails
```

### Time-to-Go-Live Benchmarks

| Customer Tier | Target Go-Live | P50 Actual | P90 Actual | Red Flag |
|---------------|---------------|------------|------------|----------|
| Self-serve | Same day | 2 hours | 3 days | >7 days |
| SMB | 7 days | 5 days | 14 days | >21 days |
| Mid-market | 30 days | 25 days | 45 days | >60 days |
| Enterprise | 90 days | 75 days | 120 days | >150 days |

---

## 7. Common Onboarding Failure Patterns

These are the eight most common ways onboarding fails. Each pattern has been observed across hundreds of SaaS companies. Recognizing the pattern early is the key to preventing churn.

### Pattern 1: The Abandoned Setup

**What it looks like**: Customer signs up, starts setup, hits a friction point, and never comes back. Their account sits at 30-60% setup completion forever.

**Symptoms:**
- Setup wizard started but not finished
- Last login was 3+ days ago, during setup
- No value action completed
- Support ticket filed during setup, possibly unresolved

**Root cause**: Setup requires too many steps, too much technical knowledge, or too much time. The customer hit a wall (integration failure, data format issue, confusing UI) and chose "I'll do this later" — which means never.

**The fix:**
- Reduce required setup steps to the absolute minimum for first value (setup the rest later)
- Add a "skip for now" option on every non-essential step
- Implement real-time setup monitoring: if a user stalls for >10 minutes on a step, trigger live chat
- Send a "pick up where you left off" email within 2 hours with a direct deep link to the exact step
- Offer a 15-minute "setup together" call for accounts that stall >24 hours

### Pattern 2: The Single-Threaded Champion

**What it looks like**: One person (the champion) is the only user. They love the product, but nobody else on their team is using it. When the champion leaves the company or changes roles, the account churns immediately.

**Symptoms:**
- 1 active user on a multi-seat license
- High individual engagement, zero team engagement
- Champion attends all calls alone
- No other users have logged in since initial invite

**Root cause**: The champion didn't (or couldn't) get buy-in from their team. Common reasons: no executive mandate, no team training, product perceived as "champion's tool" not "team tool", or the champion is a solo evaluator who never expanded.

**The fix:**
- During onboarding, explicitly ask: "Who else on your team should be using this?"
- Make team invitation part of the activation criteria, not an optional step
- Offer team training sessions (not just champion training)
- Send usage reports that show single-user dependency as a risk: "Only 1 of 8 licensed users is active"
- Identify and cultivate a secondary champion: someone who could own the tool if the primary champion leaves
- Add "team activity" features that only work with multiple users (shared dashboards, @mentions, approvals)

### Pattern 3: The Feature Tourist

**What it looks like**: Customer tries many features superficially but never goes deep on any of them. They click around, explore menus, maybe watch a demo video — but never build a real workflow.

**Symptoms:**
- Multiple features accessed, none used more than once or twice
- No real data imported (using sample/demo data)
- Session times are short and scattered
- No integration connected (they're evaluating, not implementing)

**Root cause**: The customer doesn't know which feature matters most for their specific use case. They're overwhelmed by options and haven't been guided to the ONE thing that will deliver value first.

**The fix:**
- During kickoff/onboarding, identify the single most important use case and focus on that exclusively for the first 2 weeks
- Hide or de-emphasize non-essential features during onboarding (progressive disclosure)
- Create use-case-specific onboarding paths: "You told us you want to {{GOAL}} — here's exactly how"
- Stop sending "did you know?" emails about new features until the customer has mastered the core workflow
- CSM should prescribe, not describe: "Do this next" is better than "Here are 10 things you could try"

### Pattern 4: The Ghost Customer

**What it looks like**: Customer signed the contract, got their credentials, and vanished. No logins, no responses to emails, no attendance at scheduled meetings. The account is technically active but practically dead.

**Symptoms:**
- Zero logins after Day 0 (or only 1 login on Day 0)
- No response to welcome email, check-in emails, or call invitations
- Champion unreachable (email bounces, phone goes to voicemail, Slack messages unread)
- Scheduled kickoff meeting was a no-show

**Root cause**: Usually one of three things: (1) the buyer left the company between signing and onboarding, (2) an internal priority shift made the project deprioritized, or (3) the purchase was made by someone other than the user (procurement-driven), and the actual users were never informed.

**The fix:**
- Reduce time between contract close and first engagement to <48 hours (ghosts happen in the gap)
- Have the AE make a warm introduction to CSM before the AE disengages
- If no response after 3 attempts (email, phone, LinkedIn), escalate to the AE to re-engage through their contact
- Implement a "ghost protocol": Day 1 email → Day 3 email → Day 5 phone call → Day 7 AE escalation → Day 14 executive outreach
- For procurement-driven deals, require the AE to confirm the actual end-user contact before marking the deal closed

### Pattern 5: The Over-Trained Under-User

**What it looks like**: Customer attended every training session, watched every webinar, read every help article — but isn't actually using the product in their daily work. They know the product; they just don't use it.

**Symptoms:**
- High attendance at training and enablement sessions
- Help center article views are high
- Actual product usage (logins, actions, workflows) is low
- Customer says "we love the product" but usage data tells a different story

**Root cause**: Training was about the product, not about the customer's workflow. They understand features in isolation but don't see how to integrate the product into their existing daily processes. The gap between "I know how this works" and "I use this every day" was never bridged.

**The fix:**
- Shift training from feature-based ("here's how reports work") to workflow-based ("here's how to run your Monday standup using {{PRODUCT}}")
- Require customers to complete a real task during training, not just watch a demo
- After training, assign "homework": one specific task to complete before the next session using their real data
- Create custom workflow templates that mirror their existing processes
- Identify the daily trigger: what specific moment in their workday should prompt them to open {{PRODUCT}}? Make that moment as frictionless as possible (bookmark, Slack shortcut, calendar reminder)

### Pattern 6: The Misaligned Expectations

**What it looks like**: Customer expected the product to do something it doesn't do (or doesn't do well). They're frustrated early, vocal about gaps, and start comparing to the competitor they almost chose.

**Symptoms:**
- Feature requests filed in Week 1 (not feature requests — these are "but I thought it could...")
- Negative tone in check-in calls: "When will {{FEATURE}} be available?"
- References to competitor capabilities
- Champion's enthusiasm drops visibly between Week 1 and Week 3
- Support tickets about capabilities, not bugs

**Root cause**: Sales oversold, or the customer assumed capabilities that were never explicitly discussed. The gap between what was promised (or implied) and what was delivered creates immediate buyer's remorse.

**The fix:**
- Pre-onboarding: CSM reviews sales notes and recorded demos for any promises made; flags discrepancies before kickoff
- At kickoff: Explicitly re-confirm what the product does and doesn't do; manage expectations on roadmap items
- If a gap exists: Acknowledge it honestly, explain the workaround, provide a realistic timeline for the feature, and escalate to product if it's a pattern
- Never surprise the customer with a limitation they should have known about — proactive transparency builds trust faster than reactive damage control
- Create a "Sales Promise Tracker" in the handoff document: every commitment made during sales should be logged and verified during onboarding

### Pattern 7: The Technical Blocker

**What it looks like**: Onboarding stalls because of a technical issue that neither the customer nor the CSM can resolve quickly. The customer's IT team is slow to respond, the integration fails, or the data migration hits an unexpected format issue.

**Symptoms:**
- Onboarding timeline slipping by weeks
- Repeated rescheduling of training sessions ("we're not ready yet")
- Open support tickets with no resolution
- Customer blames the product; product team blames customer's environment
- Champion is apologetic but powerless ("I'm waiting on IT")

**Root cause**: Technical requirements were not fully scoped during sales. The customer's environment has complexities (legacy systems, security policies, non-standard data formats) that weren't anticipated. Or, the customer's IT resources are overcommitted and can't prioritize the integration work.

**The fix:**
- Pre-kickoff: Conduct a technical discovery call with the customer's IT lead, not just the business champion
- Create a technical readiness checklist that the customer must complete before kickoff (SSO provider, API access, data export format, firewall rules)
- Offer a "technical fast-track" option: the vendor's Solutions Engineer does the integration work instead of waiting on customer IT
- Set clear SLAs for technical blockers: if unresolved within 5 business days, escalate to CS Manager → VP CS → executive sponsor
- Have a "Plan B" ready: a simplified setup that delivers value without the blocked integration, so onboarding can continue in parallel

### Pattern 8: The Premature Handoff

**What it looks like**: CSM declares onboarding "complete" based on a timeline (30 or 90 days passed) rather than milestones (customer is actually activated and self-sufficient). Customer is handed off to a lower-touch support model before they're ready.

**Symptoms:**
- Onboarding marked complete but activation criteria not met
- Customer health score drops within 30 days of "completing" onboarding
- Increase in support tickets after onboarding ends
- Customer feels abandoned: "My CSM disappeared"
- Renewal risk flagged within 6 months

**Root cause**: CSMs are measured on onboarding throughput (how many customers they can onboard per quarter) rather than onboarding outcomes (how many customers are truly activated). The incentive is to close the onboarding ticket and move on.

**The fix:**
- Define onboarding completion by milestones, not calendar dates (see Section 8 below)
- Require activation criteria to be met before onboarding can be marked complete
- Implement a "warm handoff" period: 2-week overlap where the onboarding CSM and the ongoing CSM are both engaged
- Post-onboarding check-in at Day 30 and Day 60 after handoff to catch early regression
- Measure CSMs on 90-day post-onboarding retention, not just onboarding completion count

---

## 8. Milestone-Based vs Time-Based Onboarding

The most common mistake in onboarding program design is using time as the primary driver. "30-day onboarding" sounds clean, but it ignores reality: some customers activate in 5 days, others need 60.

### Time-Based Onboarding

**How it works**: Every customer goes through the same sequence on the same timeline. Day 1 is welcome, Day 7 is check-in, Day 14 is training, Day 30 is graduation.

**Pros:**
- Simple to design, implement, and automate
- Easy to staff and forecast CSM capacity
- Clear expectations for customer and internal team
- Works well for homogeneous customer base with similar complexity

**Cons:**
- Fast customers are bored and under-stimulated — you're slowing them down
- Slow customers are rushed and overwhelmed — you're leaving them behind
- "Day 14 training" happens regardless of whether the customer finished setup
- Completion is based on time elapsed, not actual readiness
- Creates perverse incentive: CSMs mark onboarding "done" at Day 30 whether the customer is activated or not

### Milestone-Based Onboarding

**How it works**: Customers progress through onboarding by achieving specific milestones. The next phase unlocks when the current phase's criteria are met, regardless of calendar time.

```
Milestone-Based Progression:
─────────────────────────────────────────────
Milestone 1: Setup Complete
  Criteria: Integration live, data imported, 3+ users provisioned
  → Unlocks: Core training module

Milestone 2: First Value
  Criteria: Core workflow completed with real data, champion reports value
  → Unlocks: Team rollout phase

Milestone 3: Team Adoption
  Criteria: 50% of users active, 3+ sessions per user in last 7 days
  → Unlocks: Advanced features and optimization

Milestone 4: Self-Sufficient
  Criteria: 80% adoption, no open support tickets, positive NPS
  → Unlocks: Graduation to ongoing success cadence
```

**Pros:**
- Customers move at their own pace — no one is rushed or bored
- Completion means actual activation, not just elapsed time
- Natural quality gate: you never move to the next phase until the current one is solid
- CSM effort is allocated where it's needed (struggling customers get more time, fast ones graduate quickly)

**Cons:**
- Harder to automate — requires real-time milestone tracking
- Harder to forecast CSM capacity (you don't know when customers will graduate)
- Some customers can get "stuck" at a milestone indefinitely without a time pressure
- Requires clear, measurable milestone definitions (most companies skip this work)

### The Hybrid Approach (Recommended)

**Use milestones as the primary driver, with time-based guardrails:**

```
Hybrid Model:
─────────────────────────────────────────────
Milestone 1: Setup Complete
  Target: 7 days | Max: 14 days
  If not met by Day 14 → escalate to CSM manager, intervention call

Milestone 2: First Value
  Target: 14 days | Max: 30 days
  If not met by Day 30 → executive sponsor outreach, rescue plan

Milestone 3: Team Adoption
  Target: 30 days | Max: 60 days
  If not met by Day 60 → at-risk designation, formal recovery plan

Milestone 4: Self-Sufficient
  Target: 60 days | Max: 90 days
  If not met by Day 90 → onboarding failure review, churn risk assessment
```

This gives you the quality of milestone-based onboarding with the accountability of time-based deadlines. Customers who move fast graduate early. Customers who are slow get automatic escalation before they silently churn.

---

## 9. Segmented Onboarding by Customer Tier

Not all customers are equal, and treating them equally is a waste of resources for high-value accounts and insufficient for complex ones. Segment your onboarding by these five dimensions.

### Segmentation by ACV Tier

| ACV Range | Onboarding Model | CSM Ratio | Touchpoints | Duration |
|-----------|-----------------|-----------|-------------|----------|
| $0-$5K | Self-serve (automated) | 1:500+ | 0 live calls; all in-app + email | 14 days |
| $5K-$25K | Guided self-serve (SMB) | 1:80-120 | 2-3 live calls; hybrid automation | 30 days |
| $25K-$100K | High-touch (mid-market) | 1:20-40 | Weekly calls for month 1; biweekly after | 90 days |
| $100K-$500K | White-glove (enterprise) | 1:8-15 | Weekly calls; dedicated SE; custom implementation | 120-180 days |
| $500K+ | Strategic (named accounts) | 1:3-5 | Multiple weekly touchpoints; exec sponsor alignment; on-site visits | 180+ days |

**Key insight**: The cost of onboarding should be proportional to the ACV. Spending $10K of CSM time to onboard a $3K ACV customer is unsustainable. Spending $2K of CSM time on a $200K ACV customer is negligent.

### Segmentation by Use Case Complexity

**Simple use case** (single workflow, single team):
- Onboarding focus: Speed to first value
- Approach: Template-based setup, in-app guidance, 1-2 training sessions
- Example: Marketing team using email automation tool for newsletter sends

**Moderate use case** (multiple workflows, single team):
- Onboarding focus: Workflow design before product setup
- Approach: Workflow mapping session, phased feature rollout, weekly check-ins
- Example: Sales team using CRM for pipeline management + email sequences + reporting

**Complex use case** (multiple workflows, multiple teams):
- Onboarding focus: Cross-functional alignment before anything technical
- Approach: Stakeholder workshops, department-by-department rollout, change management
- Example: Company-wide project management platform replacing 3 existing tools

### Segmentation by Team Size

| Team Size | Onboarding Approach | Training Strategy | Adoption Target |
|-----------|--------------------|--------------------|-----------------|
| 1-5 users | Individual setup; peer learning | 1 training session for all users | 100% in 14 days |
| 6-20 users | Champion-led rollout | Champion training + team session | 80% in 30 days |
| 21-50 users | Phased rollout (2 waves) | Role-specific training sessions | 75% in 45 days |
| 51-200 users | Phased rollout (3-4 waves) | Train-the-trainer + self-serve materials | 70% in 60 days |
| 200+ users | Program-managed rollout | Learning management system + live sessions + office hours | 60% in 90 days |

**Key insight for large teams**: You cannot train 200 people in live sessions. You need a train-the-trainer model: train 5-10 internal champions who then train their own teams. Your CSM enables the champions, not the end users.

### Segmentation by Technical Sophistication

**Low technical sophistication** (non-technical users, no IT team):
- Provide: Done-for-you setup, screen-share walkthroughs, visual step-by-step guides
- Avoid: API documentation, technical jargon, self-serve integration setup
- Support: Phone/video preferred over email/docs; expect more hand-holding
- Risk: Setup abandonment if any step requires technical knowledge

**Medium technical sophistication** (tech-comfortable users, small IT team):
- Provide: Guided setup with clear documentation, integration marketplace, templates
- Avoid: Assuming they can troubleshoot API errors or configure SSO independently
- Support: Email/chat preferred; documentation should have screenshots
- Risk: Overestimating their ability; they know enough to start but not enough to finish

**High technical sophistication** (developers, dedicated IT team):
- Provide: API documentation, developer sandbox, CLI tools, webhooks guide
- Avoid: Over-explaining basics, patronizing walkthroughs, restricting access to advanced features
- Support: Self-serve preferred; documentation should have code samples and API references
- Risk: They'll build custom solutions that are hard to support; encourage using built-in features first

### Segmentation by Industry Vertical

Different industries have different onboarding considerations. Build industry-specific onboarding tracks for your top 3-5 verticals.

**Healthcare / Life Sciences:**
- Compliance: HIPAA training required before go-live; BAA must be signed
- Data: PHI handling procedures; data residency requirements
- Stakeholders: Often includes compliance officer in onboarding RACI
- Timeline: Add 2-4 weeks for security review and compliance sign-off

**Financial Services:**
- Compliance: SOC 2 documentation review; data encryption verification
- Data: Sensitive financial data requires secure import procedures; audit trails
- Stakeholders: Often includes CISO or security team in technical discovery
- Timeline: Add 2-4 weeks for vendor security assessment (VSA)

**Education / EdTech:**
- Timing: Must align with academic calendar (September start, January start)
- Stakeholders: IT department, faculty, administration — each with different goals
- Data: FERPA compliance; student data handling
- Timeline: Plan 3-6 months ahead of the academic term

**Retail / E-Commerce:**
- Timing: Avoid onboarding during peak seasons (Black Friday through January)
- Data: Product catalog import, inventory sync, POS integration
- Stakeholders: Store managers, marketing team, e-commerce team — fragmented ownership
- Timeline: Target go-live 4-6 weeks before next peak season

**Technology / SaaS:**
- Expectation: High technical sophistication; expect fast onboarding
- Data: API-first integration approach; developer documentation critical
- Stakeholders: Engineering team involved; product team may have opinions about UX
- Timeline: Fastest of all verticals — they'll be impatient if onboarding takes >30 days
