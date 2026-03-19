# Retention Strategies --- Churn Prevention Deep Dive

This reference document provides data-driven frameworks, specific thresholds, and
actionable playbooks for customer retention. Every model, score, and threshold below
is calibrated against B2B SaaS benchmarks and intended for direct implementation.

---

## 1. Customer Health Score Model

### Building a Health Score from Scratch

A customer health score is a composite numeric indicator (0--100) that predicts the
likelihood a customer will renew. The goal is to move from gut-feel to a repeatable,
data-backed signal that the entire CS org can act on.

**Step-by-step build process:**

1. **Inventory your data sources.** List every system that holds customer signals:
   product analytics (Mixpanel, Amplitude, Pendo), CRM (Salesforce, HubSpot),
   support (Zendesk, Intercom), billing (Stripe, Chargebee), communication
   (Gong, email logs).
2. **Select 10--20 candidate metrics.** Pull metrics you believe correlate with
   retention. Examples: weekly active users, feature breadth, support ticket
   volume, executive sponsor engagement, contract value trend.
3. **Run a correlation analysis.** For each metric, calculate the Pearson
   correlation with a binary outcome: renewed (1) vs. churned (0) over the last
   12--24 months. Keep metrics with |r| > 0.15.
4. **Normalize each metric to 0--100.** Use min-max normalization against your
   customer base so scores are comparable across dimensions.
5. **Assign component weights.** Start with the weights below, then iterate.
6. **Validate against historical churn.** Back-test the composite score against
   known outcomes. The score should separate churned accounts from renewed
   accounts with an AUC > 0.75.
7. **Ship a v1 and iterate quarterly.** Perfection is the enemy of deployment.

### Component Weights

| Component           | Weight | What It Captures                                         |
|---------------------|--------|----------------------------------------------------------|
| Product Usage       | 30%    | Depth and breadth of feature adoption, login frequency   |
| Engagement          | 20%    | Webinar attendance, email opens, community participation |
| Support Health      | 15%    | Ticket volume trend, CSAT on tickets, escalation count   |
| Relationship        | 15%    | Exec sponsor access, champion strength, multi-threading  |
| Business Fit        | 10%    | ICP alignment, use-case match, industry vertical fit     |
| Contract Signals    | 10%    | Payment history, contract length, expansion history      |

### Scoring Methodology with Specific Thresholds

Each component is scored 0--100 independently, then the weighted sum produces the
composite health score.

**Product Usage (30% weight) sub-scoring:**

| Sub-metric              | 0--33 (Poor)           | 34--66 (Fair)              | 67--100 (Strong)            |
|-------------------------|------------------------|----------------------------|-----------------------------|
| WAU / Licensed seats    | < 20%                  | 20--60%                    | > 60%                       |
| Core feature adoption   | < 2 of 5 core features | 2--4 of 5                  | 5 of 5                      |
| Workflow completion rate | < 40%                  | 40--75%                    | > 75%                       |
| Data volume trend (MoM) | Declining > 10%        | Flat (+/- 10%)             | Growing > 10%               |

**Engagement (20% weight) sub-scoring:**

| Sub-metric                | 0--33              | 34--66              | 67--100              |
|---------------------------|--------------------|---------------------|----------------------|
| Email open rate (CS)      | < 15%              | 15--40%             | > 40%                |
| Event/webinar attendance  | 0 in last 90 days  | 1 in last 90 days   | 2+ in last 90 days   |
| Community posts/reactions | 0 in last 60 days  | 1--3 in last 60 days| 4+ in last 60 days   |
| Training completion       | < 25% of modules   | 25--75%             | > 75%                |

**Support Health (15% weight) sub-scoring:**

| Sub-metric                  | 0--33                   | 34--66                | 67--100             |
|-----------------------------|-------------------------|-----------------------|---------------------|
| Open P1/P2 tickets          | 2+ open                 | 1 open                | 0 open              |
| Ticket volume trend (MoM)   | Increasing > 25%        | Flat (+/- 25%)        | Decreasing > 25%    |
| Avg CSAT on closed tickets  | < 3.0 / 5               | 3.0--4.2              | > 4.2               |
| Escalation in last 90 days  | 2+                      | 1                     | 0                   |

**Relationship (15% weight) sub-scoring:**

| Sub-metric                    | 0--33                    | 34--66                  | 67--100                 |
|-------------------------------|--------------------------|-------------------------|-------------------------|
| Exec sponsor identified       | No                       | Yes but no contact 60d+ | Yes and active           |
| Contacts engaged (breadth)    | 1 contact                | 2--3 contacts           | 4+ contacts              |
| Champion strength             | No champion / departed   | Passive champion        | Active internal advocate |
| Last meaningful meeting       | > 60 days ago            | 30--60 days ago         | < 30 days ago            |

**Business Fit (10% weight) sub-scoring:**

| Sub-metric           | 0--33                          | 34--66                    | 67--100                 |
|----------------------|--------------------------------|---------------------------|-------------------------|
| ICP match            | Poor (< 2 of 5 criteria)      | Partial (3 of 5)          | Strong (4--5 of 5)     |
| Use-case alignment   | Non-standard / workaround      | Partially supported       | Core use-case           |
| Competitive exposure | Active competitive eval        | Passive awareness         | No competitive signals  |

**Contract Signals (10% weight) sub-scoring:**

| Sub-metric               | 0--33                       | 34--66                    | 67--100                    |
|--------------------------|-----------------------------|---------------------------|----------------------------|
| Payment history          | Late payments (2+ in 12mo)  | 1 late in 12mo            | Always on time             |
| Contract length          | Month-to-month              | Annual                    | Multi-year                 |
| Historical expansion     | Contracted / downgraded     | Flat                      | Expanded in last cycle     |
| Renewal timeline         | < 60 days out, no signal    | 60--120 days out          | > 120 days or already renewed |

### Green / Yellow / Red Definitions

| Tier   | Score Range | Meaning                                                        | Action Posture         |
|--------|-------------|----------------------------------------------------------------|------------------------|
| Green  | 80--100     | Healthy. On track for renewal and potential expansion.         | Proactive expansion    |
| Yellow | 40--79      | At risk. One or more components trending down.                 | Re-engagement required |
| Orange | 20--39      | High risk. Multiple signals negative. Churn likely without intervention. | Save campaign   |
| Red    | 0--19       | Critical. Customer is likely already deciding to leave.        | Executive escalation   |

### Example: Scoring a Real SaaS Account

**Account: Acme Corp** --- 200-seat enterprise license, 14 months into contract.

| Component         | Raw Score | Weight | Weighted |
|-------------------|-----------|--------|----------|
| Product Usage     | 72        | 0.30   | 21.6     |
| Engagement        | 45        | 0.20   | 9.0      |
| Support Health    | 38        | 0.15   | 5.7      |
| Relationship      | 60        | 0.15   | 9.0      |
| Business Fit      | 85        | 0.10   | 8.5      |
| Contract Signals  | 70        | 0.10   | 7.0      |
| **Composite**     |           |        | **60.8** |

**Interpretation:** Yellow tier. Usage is decent but engagement is mediocre and
support health is poor (recent escalation + rising ticket volume). Primary action:
investigate support issues, re-engage stakeholders, schedule a value-realization
workshop within 14 days.

### How to Calibrate and Iterate

1. **Quarterly back-test.** Re-score all accounts that renewed or churned in the
   prior quarter. Calculate precision and recall at each tier boundary.
2. **Adjust weights.** If Product Usage has a stronger correlation with churn than
   initially assumed, increase its weight. Rebalance so weights still sum to 100%.
3. **Add or remove sub-metrics.** If a sub-metric has near-zero variance across
   your customer base, drop it. If a new signal emerges (e.g., a new product
   module), add it.
4. **Benchmark tier distribution.** A healthy portfolio should be roughly 60% Green,
   25% Yellow, 10% Orange, 5% Red. If your distribution skews heavily to one tier,
   recalibrate thresholds.
5. **Survey CSMs.** Ask CSMs to flag the 10 accounts they are most worried about.
   Compare against the model's Red/Orange list. Discrepancies reveal missing signals.

---

## 2. Early Warning Indicators with Thresholds

The following table lists 18 early warning signals organized by category. Each
signal includes measurement guidance, tier thresholds, the system of record, and
the maximum response time once the threshold is crossed.

### Usage Signals

| # | Signal                        | Measure                           | Green            | Yellow             | Red                | Data Source         | Response Time |
|---|-------------------------------|-----------------------------------|------------------|--------------------|--------------------|---------------------|---------------|
| 1 | DAU decline                   | DAU % change over 14 days         | Stable or up     | Drop > 30%         | Drop > 50%         | Product analytics   | 48 hours      |
| 2 | WAU / Licensed seats ratio    | Weekly active users / total seats | > 60%            | 30--60%            | < 30%              | Product analytics   | 1 week        |
| 3 | Core feature abandonment      | # core features used in 30 days   | 4--5 of 5        | 2--3 of 5          | 0--1 of 5          | Product analytics   | 1 week        |
| 4 | Data import / export spike    | Export volume vs. 30-day avg      | < 1.5x avg       | 1.5--3x avg        | > 3x avg           | Product analytics   | 24 hours      |
| 5 | API call volume drop          | API calls % change over 7 days    | Stable or up     | Drop > 40%         | Drop > 70%         | API gateway logs    | 48 hours      |
| 6 | Login frequency decline       | Avg logins/user/week vs. baseline | > 80% of baseline| 50--80% of baseline| < 50% of baseline  | Auth logs           | 1 week        |

### Engagement Signals

| # | Signal                        | Measure                           | Green            | Yellow             | Red                | Data Source         | Response Time |
|---|-------------------------------|-----------------------------------|------------------|--------------------|--------------------|---------------------|---------------|
| 7 | CS email non-response         | Days since last reply to CSM      | < 7 days         | 7--21 days         | > 21 days          | Email / CRM         | 3 days        |
| 8 | Meeting cancellations         | Cancelled meetings in 60 days     | 0                | 1--2               | 3+                 | Calendar / CRM      | 1 week        |
| 9 | Training non-completion       | % assigned training completed     | > 75%            | 40--75%            | < 40%              | LMS                 | 2 weeks       |
| 10| Community silence             | Days since last community activity| < 30 days        | 30--60 days        | > 60 days          | Community platform  | 2 weeks       |

### Support Signals

| # | Signal                        | Measure                           | Green            | Yellow             | Red                | Data Source         | Response Time |
|---|-------------------------------|-----------------------------------|------------------|--------------------|--------------------|---------------------|---------------|
| 11| Ticket volume spike           | Tickets this month vs. 3-mo avg   | < 1.5x avg       | 1.5--2.5x avg      | > 2.5x avg         | Support platform    | 48 hours      |
| 12| Escalation frequency          | Escalations in last 90 days       | 0                | 1                  | 2+                 | Support platform    | 24 hours      |
| 13| Negative CSAT on tickets      | % of tickets rated < 3/5          | < 10%            | 10--25%            | > 25%              | Support platform    | 48 hours      |

### Relationship Signals

| # | Signal                        | Measure                           | Green            | Yellow             | Red                | Data Source         | Response Time |
|---|-------------------------------|-----------------------------------|------------------|--------------------|--------------------|---------------------|---------------|
| 14| Champion departure            | Key contact left the account      | No change        | Role change        | Departed company   | LinkedIn / CRM      | 24 hours      |
| 15| Exec sponsor disengagement    | Days since exec sponsor contact   | < 45 days        | 45--90 days        | > 90 days          | CRM                 | 1 week        |
| 16| Stakeholder contraction       | # of engaged contacts trend       | Stable or growing| Dropped by 1--2    | Dropped by 3+      | CRM                 | 1 week        |

### Business Signals

| # | Signal                        | Measure                           | Green            | Yellow             | Red                | Data Source         | Response Time |
|---|-------------------------------|-----------------------------------|------------------|--------------------|--------------------|---------------------|---------------|
| 17| Competitor mentions           | Competitor named in calls/tickets | 0 in 90 days     | 1 in 90 days       | 2+ in 90 days      | Gong / Support      | 24 hours      |
| 18| Budget / reorg announcement   | Customer announces cuts or reorg  | No signals       | Reorg announced    | Layoffs or budget freeze | News / CRM   | 24 hours      |

### Composite Alert Logic

When 3+ signals cross into Yellow simultaneously, auto-escalate the account to
Orange regardless of composite health score. When any single signal crosses into
Red, trigger an immediate CSM alert and require a documented action plan within
the response time listed above.

---

## 3. Intervention Playbooks by Risk Tier

### Green Tier (Score 80--100): Expansion-Focused Actions

**Objective:** Maximize lifetime value through expansion and advocacy.

| Timeframe | Action                                                      | Owner     |
|-----------|-------------------------------------------------------------|-----------|
| Weekly    | Monitor for expansion triggers (new team, new use case)     | CSM       |
| Bi-weekly | Share relevant product updates, case studies, best practices| CSM       |
| Monthly   | Review usage data for upsell signals (>80% seat utilization)| CSM + AE  |
| Quarterly | Conduct strategic QBR focused on roadmap alignment          | CSM + Exec|
| Ongoing   | Nominate for advisory board, reference program, case study  | CSM       |

**Escalation to Yellow:** If any two health components drop below 60 in a single
month, reclassify and switch to the Yellow playbook.

**Success metrics:** NRR > 120%, expansion rate > 15% annually, NPS > 50.

### Yellow Tier (Score 40--79): Re-Engagement Playbook

**Objective:** Stabilize the account and return it to Green within 60 days.

**Week 1 --- Diagnose:**
1. Pull the full health score breakdown. Identify the 1--2 components dragging
   the score down.
2. Review recent support tickets, call recordings, and email threads for context.
3. Schedule a 30-minute internal sync with the AE, support lead, and CSM manager
   to align on the situation.

**Week 2 --- Reconnect:**
4. Send a personalized outreach to the primary contact acknowledging the gap.
   Do not lead with "we noticed your usage is down." Instead, lead with value:
   "I wanted to share some updates relevant to [their stated goal]."
5. Propose a 45-minute working session to re-align on goals and review a
   success plan.

**Weeks 3--4 --- Deliver Quick Wins:**
6. Execute on 1--2 quick wins identified in the working session (e.g., custom
   training, configuration optimization, integration setup).
7. Introduce the customer to a peer in a similar industry who has solved the
   same problem (peer networking).

**Weeks 5--8 --- Sustain:**
8. Increase touchpoint cadence to weekly for 4 weeks.
9. Share a written success plan with measurable milestones and check in against
   it weekly.
10. Re-score the account at week 8. If score is 70+, return to Green cadence. If
    still below 70, escalate to Orange.

**Escalation criteria:** Score drops below 40, or champion departs, or P1
escalation during this period.

**Success metrics:** Score returns to 70+ within 60 days in > 60% of cases.

### Orange Tier (Score 20--39): Save Campaign / War Room

**Objective:** Prevent churn through concentrated executive and cross-functional
intervention. Target: save 40--50% of Orange accounts.

**Day 1 --- War Room Kickoff:**
1. CSM manager opens a dedicated Slack channel: #save-[account-name].
2. Assemble the save team: CSM, CSM manager, AE, AE manager, support lead,
   product liaison.
3. Document the timeline of decline: when did each signal go Yellow/Red? What
   interventions were attempted?

**Days 2--5 --- Executive Outreach:**
4. CSM manager or VP of CS sends a personal email to the customer's executive
   sponsor: "Your success is my top priority. I'd like to schedule 30 minutes
   this week to understand how we can better support [company]."
5. Prepare a custom value-realization report showing ROI delivered to date,
   benchmarked against peers.

**Days 6--14 --- Concession and Commitment:**
6. If appropriate, offer a structured concession (not a blanket discount).
   Examples: 60 days of premium support at no charge, dedicated SE for
   integration work, early access to a requested feature.
7. Any concession must be tied to a mutual commitment: "We will provide X if
   you commit to Y" (e.g., completing onboarding, attending training, providing
   executive alignment).

**Days 15--30 --- Execution Sprint:**
8. Execute on the mutual commitment plan with daily stand-ups in the save
   channel.
9. Track leading indicators daily: login count, feature usage, ticket sentiment.

**Day 30 --- Decision Point:**
10. Re-score the account. If score is 40+, move to Yellow playbook. If still
    below 40, escalate to Red with a documented recommendation to the CRO.

**Escalation criteria:** Customer explicitly states intent to cancel, or score
drops below 20.

**Success metrics:** 40--50% of Orange accounts return to Yellow or Green within
60 days.

### Red Tier (Score 0--19): Last Resort / Graceful Exit

**Objective:** Make a final save attempt. If unsuccessful, ensure a professional
off-boarding that preserves the relationship for a future win-back.

**Final Save Attempt (Days 1--14):**
1. CRO or VP-level executive reaches out directly. This is the highest-level
   outreach the company makes.
2. Offer a "restart" package: re-onboarding at no cost, dedicated resources for
   90 days, contract restructuring if needed (shorter term, reduced scope).
3. If the customer agrees, treat as a new Orange-tier save campaign with the
   same war room rigor.

**Graceful Exit (if save fails):**
4. Do not burn the bridge. Express genuine gratitude for the partnership.
5. Offer a data export package and reasonable transition timeline (minimum 30
   days).
6. Conduct a structured exit interview (see Section 8).
7. Add to the win-back pipeline with a 90-day and 180-day follow-up scheduled.
8. Internally, complete the churn post-mortem within 14 days.

**Success metrics:** 10--20% of Red accounts saved. 100% of exiting customers
complete an exit interview. 15% win-back rate within 12 months.

---

## 4. NPS / CSAT Program Design

### Setting Up an NPS Program

**Timing options:**

| Type             | When to Send                           | Best For                        |
|------------------|----------------------------------------|---------------------------------|
| Relationship NPS | Every 90 days, staggered across base   | Overall health tracking         |
| Transactional NPS| After key milestones (onboarding, QBR) | Milestone-specific feedback     |
| Post-interaction | After support ticket closure           | Support quality (better as CSAT)|

**Recommended approach:** Relationship NPS every 90 days, sent on a Tuesday or
Wednesday at 10:00 AM in the recipient's local time zone. Stagger sends so that
roughly 1/13 of your base is surveyed each week, creating a continuous feedback
stream rather than a quarterly spike.

**Channel priority:** In-app modal > email > SMS. In-app yields 30--45% response
rates vs. 10--20% for email.

### Response Strategies by Score Range

**Promoters (9--10):**
- Auto-trigger a thank-you message within 1 hour.
- Within 7 days, CSM sends a personal follow-up asking if they would be open to
  a case study, reference call, or G2 review.
- Add to the advocacy pipeline (see Section 6).
- Flag for expansion conversation within 30 days.

**Passives (7--8):**
- CSM reviews the account health score and any open issues within 48 hours.
- Send a follow-up asking: "What's one thing we could do to earn a 9 or 10?"
- Schedule a check-in within 14 days to discuss any feedback.
- Do not ignore passives --- they are the swing segment.

**Detractors (0--6):**
- Alert CSM and CSM manager immediately (< 4 hours).
- CSM calls (not emails) the respondent within 24 hours.
- Acknowledge the feedback, do not defend. Ask: "Can you help me understand
  what's driving this score?"
- Create a documented action plan within 48 hours. Share it with the customer.
- Follow up within 14 days to report on progress.
- Re-survey in 45 days (not 90) to measure recovery.

### Closing the Feedback Loop

The number one driver of NPS response rate over time is demonstrating that
feedback leads to action. Implement a "You Said, We Did" communication:

1. **Individual level:** CSM sends a personal note to every respondent who left
   a comment, summarizing the action taken.
2. **Aggregate level:** Quarterly, publish a summary to all customers: "Based on
   your feedback, we shipped X, changed Y, and are working on Z."
3. **Internal level:** Product team receives a monthly digest of NPS verbatims
   tagged by theme. Top 3 themes are reviewed in the product planning meeting.

### NPS Benchmarks by Industry

| Industry          | Median NPS | Top Quartile |
|-------------------|-----------|--------------|
| B2B SaaS          | 36        | 55+          |
| Fintech           | 32        | 50+          |
| E-commerce        | 45        | 62+          |
| Healthcare IT     | 28        | 45+          |
| Dev Tools         | 42        | 60+          |
| HR Tech           | 30        | 48+          |
| Cybersecurity     | 34        | 52+          |

### CSAT vs. NPS vs. CES --- When to Use Each

| Metric | Question                                        | Scale     | Best For                          | Frequency         |
|--------|-------------------------------------------------|-----------|-----------------------------------|--------------------|
| NPS    | How likely are you to recommend us?             | 0--10     | Overall relationship health       | Every 90 days      |
| CSAT   | How satisfied were you with [interaction]?      | 1--5      | Transactional quality measurement | After each interaction |
| CES    | How easy was it to [accomplish task]?           | 1--7      | Product/process friction detection| After key workflows|

**Rule of thumb:** Use NPS for strategic health, CSAT for support/service quality,
and CES for product usability. Do not survey the same person with more than one
metric in a 14-day window.

### Survey Design Best Practices

- Keep the survey to 1 primary question + 1 open-text follow-up. Every additional
  question reduces completion rate by approximately 10--15%.
- For NPS, always ask: "What is the primary reason for your score?" as the
  follow-up.
- Avoid leading questions. Bad: "How much did you love our new feature?"
  Good: "How would you rate your experience with [feature]?"
- Randomize response options in multi-choice questions to avoid order bias.

### Response Rate Optimization

| Tactic                          | Expected Lift |
|---------------------------------|---------------|
| In-app survey instead of email  | +15--25%      |
| Personalize subject line (name) | +5--8%        |
| Send Tuesday/Wednesday 10am     | +3--5%        |
| Keep to 2 questions max         | +10--15%      |
| Show "takes 30 seconds"         | +5--7%        |
| Follow-up reminder at 72 hours  | +8--12%       |
| "You Said, We Did" in prior comm| +5--10% (long-term) |

---

## 5. QBR (Quarterly Business Review) Frameworks

### Detailed QBR Agenda Template (60 minutes)

| Time       | Section                        | Owner    | Notes                                   |
|------------|--------------------------------|----------|-----------------------------------------|
| 0--5 min   | Welcome and agenda review      | CSM      | Confirm attendees and objectives         |
| 5--15 min  | Business outcomes review       | CSM      | Show progress against stated goals       |
| 15--25 min | Usage and adoption deep dive   | CSM      | Benchmarked against peers and plan       |
| 25--35 min | ROI and value delivered        | CSM      | Quantified in dollars or time saved      |
| 35--45 min | Roadmap preview and feedback   | Product  | 2--3 upcoming features relevant to them  |
| 45--55 min | Strategic planning             | Customer | Customer shares upcoming priorities      |
| 55--60 min | Action items and next steps    | CSM      | Document and assign owners               |

### Pre-QBR Preparation Checklist

Complete at least 5 business days before the QBR:

- [ ] Pull usage data for the quarter (DAU, WAU, feature adoption, data volume).
- [ ] Calculate ROI metrics (time saved, revenue impact, cost reduction).
- [ ] Review all support tickets from the quarter; summarize themes.
- [ ] Check health score trend and identify any declining components.
- [ ] Review open feature requests and their status.
- [ ] Confirm attendees. Ensure at least one executive from the customer side.
- [ ] Prepare a 1-page executive summary for the customer's exec sponsor.
- [ ] Identify 1--2 expansion opportunities to discuss if the conversation allows.
- [ ] Pre-align with your AE on any commercial topics.
- [ ] Send the agenda to the customer 3 business days in advance.

### How to Present ROI Data

**Formula template:**

```
Value Delivered = (Hours Saved x Hourly Rate) + (Revenue Influenced) + (Cost Avoided)
```

**Example:**
- 12 team members saved an average of 4 hours/week using the platform.
- Hourly fully-loaded cost: $75.
- Quarterly value: 12 x 4 x 13 weeks x $75 = $46,800 in labor savings.
- Plus: 3 deals worth $180,000 influenced by platform insights.
- Total quarterly ROI: $226,800 on a $15,000/quarter contract = 15.1x ROI.

Present ROI as a multiple (15x) and in absolute dollars. Executives respond to
multiples; operators respond to absolute numbers. Use both.

### Handling Difficult QBRs

**Scenario: Declining usage.**
Do not hide it. Lead with the data: "Usage has declined 22% this quarter. Let's
understand why and build a plan." Then pivot to diagnosis: "Is this a seasonal
pattern, a team change, or a product gap?" Customers respect transparency.

**Scenario: Missed goals.**
Reframe around what was achieved and what was learned. "We set a target of 80%
adoption. We reached 58%. Here's what we learned about the blockers, and here's
our revised plan for next quarter."

**Scenario: Customer is frustrated.**
Let them vent. Do not interrupt or get defensive. Summarize what you heard. Ask:
"What does a successful next 90 days look like to you?" Then build the QBR around
that answer.

### Executive QBR vs. Operational QBR

| Dimension     | Executive QBR              | Operational QBR              |
|---------------|----------------------------|------------------------------|
| Audience      | VP+ level                  | Managers, power users        |
| Duration      | 30--45 minutes             | 60 minutes                   |
| Focus         | Business outcomes, ROI     | Usage, adoption, best practices |
| Frequency     | Quarterly or semi-annually | Quarterly                    |
| Depth         | High-level trends          | Granular data and workflows  |
| Expansion     | Strategic roadmap          | Tactical upsell (seats, features) |

For enterprise accounts ($100K+ ACV), run both formats. For mid-market ($25K--$100K),
combine into a single QBR that starts executive and goes operational in the second
half. For SMB (< $25K), replace with a 30-minute digital check-in or automated
success report.

### QBR Follow-Up Process

1. **Within 24 hours:** Send a written summary with action items, owners, and
   deadlines. Use a shared document, not a buried email.
2. **Within 1 week:** Complete any quick-win action items. Communicate progress.
3. **At 30 days:** Send a mid-quarter check-in referencing the QBR action items.
4. **At 60 days:** Begin preparing for the next QBR. Flag any at-risk items.

### When to Skip or Modify the QBR

- **Skip if:** The customer explicitly requests no QBRs and the account is Green.
  Replace with a monthly async health report they can review on their own time.
- **Modify if:** The customer is in Orange or Red tier. Replace the standard QBR
  with a focused "recovery review" --- shorter, more frequent (monthly), and
  centered on the action plan rather than the standard agenda.
- **Escalate format if:** The customer is up for renewal in < 90 days. Add a
  renewal-readiness section to the agenda.

---

## 6. Customer Advocacy Programs

### Program Types

| Program         | Description                                         | Effort to Run | Retention Impact |
|-----------------|-----------------------------------------------------|---------------|------------------|
| Advisory Board  | 8--12 customers meet quarterly to advise on roadmap | High          | Very High        |
| Reference Program| Customers available for prospect calls/case studies | Medium        | High             |
| Beta Testers    | Early access to features in exchange for feedback   | Medium        | High             |
| User Group      | Regional or virtual meetups for peer learning       | Medium        | Medium--High     |
| Ambassador      | Power users who evangelize in their networks        | Low--Medium   | High             |

### How to Identify and Recruit Advocates

**Identification criteria (must meet 3+ of 5):**
1. NPS score of 9 or 10 in the last 6 months.
2. Health score in Green tier for 2+ consecutive quarters.
3. Has expanded their contract at least once.
4. Actively uses 4+ of 5 core features.
5. Has voluntarily referred or recommended the product (verbal or written).

**Recruitment approach:**
- CSM sends a personal invitation (not a mass email): "You're one of our most
  successful customers, and I think your perspective would be incredibly valuable
  to our product team. Would you be open to joining our Advisory Board?"
- Frame it as exclusive access and influence, not as a favor.
- Provide a clear commitment ask: "4 meetings per year, 60 minutes each."

### Incentive Structures (Non-Monetary Preferred)

| Incentive                          | Cost   | Perceived Value | Notes                          |
|------------------------------------|--------|-----------------|--------------------------------|
| Early access to new features       | Zero   | Very High       | Most effective single incentive|
| Direct line to product leadership  | Zero   | High            | Advisory board members love this|
| Conference speaking opportunity    | Low    | High            | Builds their personal brand    |
| Co-branded case study              | Low    | Medium--High    | Marketing value for both parties|
| Annual appreciation dinner/event   | Medium | High            | In-person connection matters   |
| Branded swag (premium, not cheap)  | Low    | Medium          | Only if high quality           |
| Charitable donation in their name  | Low    | Medium          | Aligns with CSR-minded customers|

Avoid cash incentives or discounts for advocacy. They attract mercenary behavior
and devalue the relationship.

### Measuring Advocacy Impact

| Metric                             | Target               | Measurement Frequency |
|------------------------------------|----------------------|-----------------------|
| Advocate retention rate            | > 95%                | Quarterly             |
| NRR of advocate accounts           | > 130%               | Quarterly             |
| References completed per quarter   | 3--5 per advocate    | Quarterly             |
| Case studies published per quarter | 2--4                 | Quarterly             |
| Influenced pipeline from referrals | 10--15% of new pipeline | Monthly            |
| G2/Gartner Peer review submissions | 2+ per advocate/year | Semi-annually        |

### Community Building Strategies

1. **Start small.** Launch with a private Slack or Discord for your top 20
   advocates before building a full community platform.
2. **Seed content.** Post 3--5 discussion threads per week for the first 3 months.
   Ask specific questions, not open-ended ones.
3. **Recognize contributors.** Public shout-outs, "Member of the Month," and
   featured posts drive engagement.
4. **Connect members to each other.** The primary value of a community is
   peer-to-peer learning, not vendor-to-customer broadcasting.
5. **Measure monthly active participants,** not total sign-ups. A community of
   50 active members is more valuable than 500 lurkers.

---

## 7. Expansion Revenue Strategies

### Identifying Expansion Signals

| Signal                                  | Strength  | Typical Trigger                              |
|-----------------------------------------|-----------|----------------------------------------------|
| Seat utilization > 90%                  | Strong    | Team is at capacity; needs more licenses     |
| New department requesting access        | Strong    | Organic internal spread                      |
| API usage hitting plan limits           | Strong    | Technical integration deepening              |
| Customer asks about a premium feature   | Strong    | Self-identified need                         |
| Usage of a feature that's a gateway     | Moderate  | E.g., using basic reporting before analytics |
| Customer hires for a role your tool serves | Moderate | Growing investment in the function         |
| Positive QBR with exec sponsor buy-in   | Moderate  | Strategic alignment confirmed                |
| High NPS + completed advocacy activity  | Moderate  | Relationship is strong enough for the ask    |

### Upsell vs. Cross-Sell Triggers

**Upsell** (more of the same product):
- Seat utilization > 85%.
- Usage volume approaching plan tier limits.
- Customer requests higher SLA or priority support.
- New team or office wants access.

**Cross-sell** (different product or module):
- Customer manually doing something your other product automates.
- Feature request that maps to an existing but un-purchased module.
- Adjacent team (e.g., marketing team when sales team is the current user) has a
  relevant pain point.
- Customer's industry peers commonly use the adjacent module.

### Timing Expansion Conversations

| Timing                                        | Appropriateness |
|-----------------------------------------------|-----------------|
| Immediately after a major success milestone   | Excellent       |
| During a QBR where ROI is clearly demonstrated| Excellent       |
| After a Promoter NPS response                 | Good            |
| 90--120 days before renewal                   | Good            |
| During onboarding of a new team               | Acceptable      |
| Right after a support escalation              | Poor            |
| When the account is in Yellow or below        | Poor            |
| During budget freeze or layoffs               | Do not attempt   |

### Land-and-Expand Playbook

1. **Land:** Close a focused initial deal (single team, single use case,
   12-month term). Prioritize adoption over ACV.
2. **Adopt (months 1--3):** Drive the initial team to full adoption. Measure:
   > 80% WAU/seats, > 4 of 5 core features used, CSAT > 4.2.
3. **Prove (months 4--6):** Document and quantify the ROI for the initial team.
   Build a one-page internal case study the champion can share.
4. **Expand (months 6--9):** Champion introduces CSM to the adjacent team or
   exec sponsor. CSM runs a "discovery workshop" for the new team.
5. **Repeat (months 9--12):** Close the expansion deal. Set the next expansion
   target. Roll the playbook forward.

### Usage-Based Expansion Triggers

For products with usage-based pricing or tiered plans:

| Metric                          | Trigger Point         | Action                                     |
|---------------------------------|-----------------------|--------------------------------------------|
| API calls at 80% of plan limit  | 80% threshold         | Proactive outreach: "You're scaling fast"  |
| Storage at 90% of plan limit    | 90% threshold         | Alert + upgrade path with ROI justification|
| Users at 95% of seat limit      | 95% threshold         | Propose a block add-on before overage hits |
| Overage incurred                | Any overage event     | Within 48 hours, present a plan upgrade    |

### Multi-Product Adoption Strategies

1. **Bundle pricing incentive.** Offer a 10--15% discount when adopting a second
   module, structured as a commitment discount (not a concession).
2. **Integrated onboarding.** When a customer buys a second product, assign a
   single CSM across both to ensure a unified experience.
3. **Cross-product value stories.** Show how customers using both products achieve
   30--50% better outcomes than single-product users.
4. **Product-led cross-sell.** Surface the second product's value inside the
   first product's UI (e.g., "This report would be richer with [Module B] data").

### Calculating Net Revenue Retention (NRR)

```
NRR = (Beginning ARR + Expansion - Contraction - Churn) / Beginning ARR x 100
```

**Example:**
- Beginning ARR (Jan 1): $10,000,000
- Expansion (new seats, upsells, cross-sells): $1,800,000
- Contraction (downgrades): $400,000
- Churn (lost customers): $600,000
- NRR = ($10M + $1.8M - $0.4M - $0.6M) / $10M x 100 = **118%**

### NRR Benchmarks and Targets

| Company Stage         | Median NRR | Top Quartile | Elite       |
|-----------------------|-----------|--------------|-------------|
| Early stage (< $10M)  | 95--100%  | 105--110%    | 115%+       |
| Growth ($10M--$100M)  | 105--110% | 115--120%    | 125%+       |
| Scale ($100M+)        | 110--115% | 120--130%    | 140%+       |

An NRR below 100% means the customer base is shrinking even before accounting
for new logo acquisition. The goal is sustained NRR above 110%.

---

## 8. Churn Post-Mortem Framework

### Exit Interview Process

Conduct the exit interview within 14 days of confirmed churn. The interviewer
should be someone other than the day-to-day CSM (e.g., CSM manager or CX
research lead) to reduce bias and encourage candor.

**Interview structure (30 minutes):**

1. **Context (5 min):** "Thank you for your time. We value your honesty and want
   to learn from this experience."
2. **Decision drivers (10 min):** "Walk me through the decision to move away.
   What was the primary driver? Were there secondary factors?"
3. **Product gaps (5 min):** "Were there specific capabilities that were missing
   or didn't meet your needs?"
4. **Service experience (5 min):** "How would you rate the support and
   partnership you received?"
5. **Alternative chosen (3 min):** "Can you share what solution you're moving to
   and what tipped the decision?"
6. **Win-back signal (2 min):** "Under what circumstances, if any, would you
   consider us again in the future?"

Record and transcribe the interview (with permission). Tag the transcript with
the churn taxonomy below.

### Churn Categorization Taxonomy

| Category              | Sub-category               | Definition                                           | Preventable? |
|-----------------------|----------------------------|------------------------------------------------------|--------------|
| **Voluntary**         | Product gap                | Customer needed functionality we don't offer         | Partially    |
| Voluntary             | Competitive loss           | Customer chose a competitor for price or features    | Yes          |
| Voluntary             | Value not realized         | Customer couldn't achieve their goals with us        | Yes          |
| Voluntary             | Champion departure         | Key internal advocate left; no replacement found     | Partially    |
| Voluntary             | Strategic shift            | Customer changed business direction entirely         | No           |
| Voluntary             | Budget cut                 | Customer eliminated the budget line item             | No           |
| **Involuntary**       | Non-payment                | Customer failed to pay and was off-boarded           | Partially    |
| Involuntary           | Acquisition / merger       | Customer was acquired and consolidated tools         | No           |
| Involuntary           | Business closure           | Customer went out of business                        | No           |

**Target:** Preventable churn should represent < 40% of total churn. If it is
higher, there are systemic issues in product, onboarding, or CS to address.

### Pattern Recognition Across Churned Accounts

Run this analysis quarterly on all accounts that churned in the prior 90 days:

1. **Cohort by churn reason.** What % fell into each taxonomy category? Is one
   category growing?
2. **Cohort by acquisition channel.** Do customers from a specific channel churn
   at higher rates? If so, revisit ICP and qualification criteria with sales.
3. **Cohort by ACV.** Is churn concentrated in a specific deal-size band?
   Sub-$10K accounts churning at 2x the rate of $50K+ accounts may indicate an
   SMB product-market fit issue.
4. **Cohort by time-to-first-value.** Accounts that took > 45 days to reach
   first value churn at 2--3x the rate of those who reached it in < 14 days.
5. **Cohort by CSM.** Normalize for portfolio size and account quality. If one
   CSM has materially higher churn, investigate and coach.

### Feeding Learnings Back into Product, Sales, and CS

| Learning                        | Destination | Action                                               |
|---------------------------------|-------------|------------------------------------------------------|
| Repeated product-gap citations  | Product     | Add to feature request tracker with churn $ attached |
| High churn from a specific channel | Sales    | Tighten qualification criteria for that channel      |
| Value not realized in segment X | CS / Onboarding | Redesign onboarding for that segment           |
| Champion departure pattern      | CS          | Build multi-threading into the account plan standard |
| Competitive loss to vendor Y    | Product + Marketing | Competitive battlecard update + feature parity review |

### Win-Back Timing and Approach

| Time Since Churn | Action                                                     | Expected Win-Back Rate |
|------------------|------------------------------------------------------------|------------------------|
| 0--30 days       | None. Respect the decision. Do not pester.                 | N/A                    |
| 30--90 days      | Light touch: share a relevant product update or case study.| 3--5%                  |
| 90--180 days     | Direct outreach: "A lot has changed since [reason]. Can I show you?" | 8--12%       |
| 180--365 days    | Re-engage with a compelling event: new product launch, major feature, industry report. | 10--15% |
| 365+ days        | Annual check-in. If they left due to a product gap, demonstrate that it's been closed. | 5--10% |

**Win-back offer structure:** Do not lead with a discount. Lead with the
resolution of their original churn reason. If the product gap is closed, show
it. If value realization was the issue, offer a structured re-onboarding with
executive sponsorship from your side.

### Churn Cohort Analysis Methodology

1. **Define the cohort.** All accounts with a churn effective date in the
   analysis period (e.g., Q4 2025).
2. **Enrich the data.** For each churned account, capture: ACV, tenure, health
   score at 90/60/30 days pre-churn, churn taxonomy category, CSM, acquisition
   channel, industry, company size.
3. **Calculate churn rates.** Gross dollar churn rate, logo churn rate, and
   segment-specific rates (by ACV band, industry, tenure).
4. **Identify leading indicators.** Which health score components were most
   predictive? At what threshold and how far in advance?
5. **Benchmark internally.** Compare this quarter's churn to the prior 4
   quarters. Is it trending up, down, or flat?
6. **Publish findings.** Share a written post-mortem report with CS leadership,
   product leadership, and the CRO within 21 days of quarter close.
7. **Track action items.** Every finding should produce at least one action item
   with an owner and a deadline. Review progress in the next quarterly analysis.

---

## Quick Reference: Key Metrics and Targets

| Metric                      | Target (Growth Stage) | Target (Scale Stage) |
|-----------------------------|----------------------|----------------------|
| Gross Revenue Retention     | > 90%                | > 93%                |
| Net Revenue Retention       | > 110%               | > 120%               |
| Logo Retention Rate         | > 85%                | > 90%                |
| NPS                         | > 40                 | > 50                 |
| Time to First Value         | < 14 days            | < 10 days            |
| QBR Attendance Rate         | > 80%                | > 85%                |
| Health Score % Green        | > 55%                | > 65%                |
| Preventable Churn %         | < 45%                | < 35%                |
| Expansion Rate              | > 15% of base/year   | > 20% of base/year   |
| Win-Back Rate (12 months)   | > 10%                | > 12%                |
